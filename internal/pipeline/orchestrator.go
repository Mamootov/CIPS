package pipeline

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"cips/internal/domain"
)

// OrchestratorCallbacks defines real-time event hooks for dashboard/CLI updates.
type OrchestratorCallbacks struct {
	OnStatsUpdate func(stats domain.ScanSessionStats)
	OnItemUpdate  func(report *domain.ProxyEvaluationReport)
}

// Orchestrator manages the end-to-end two-stage proxy scanning pipeline.
type Orchestrator struct {
	engine      domain.IProxyEngine
	probes      []domain.IProbe
	callbacks   OrchestratorCallbacks
	concurrency int
	pingTimeout int
	probeTimeout int
	cancelFunc  context.CancelFunc
	isRunning   atomic.Bool
}

// NewOrchestrator constructs a configured pipeline orchestrator.
func NewOrchestrator(
	engine domain.IProxyEngine,
	probes []domain.IProbe,
	concurrency int,
	pingTimeout int,
	probeTimeout int,
	callbacks OrchestratorCallbacks,
) *Orchestrator {
	if concurrency <= 0 {
		concurrency = 20
	}
	if pingTimeout <= 0 {
		pingTimeout = 3000
	}
	if probeTimeout <= 0 {
		probeTimeout = 8000
	}

	return &Orchestrator{
		engine:       engine,
		probes:       probes,
		callbacks:    callbacks,
		concurrency:  concurrency,
		pingTimeout:  pingTimeout,
		probeTimeout: probeTimeout,
	}
}

// SetProbes dynamically updates active probes.
func (o *Orchestrator) SetProbes(probes []domain.IProbe) {
	o.probes = probes
}

// Cancel terminates any currently executing scan session immediately.
func (o *Orchestrator) Cancel() {
	if o.cancelFunc != nil {
		o.cancelFunc()
	}
	o.isRunning.Store(false)
	_ = o.engine.Shutdown()
}

// IsRunning indicates if a scan is currently active.
func (o *Orchestrator) IsRunning() bool {
	return o.isRunning.Load()
}

// Execute runs the two-stage evaluation pipeline across the provided proxy candidates.
func (o *Orchestrator) Execute(parentCtx context.Context, candidates []domain.ProxyConfig) ([]*domain.ProxyEvaluationReport, error) {
	if len(candidates) == 0 {
		return nil, nil
	}

	if !o.isRunning.CompareAndSwap(false, true) {
		return nil, fmt.Errorf("a scan is already in progress")
	}
	defer o.isRunning.Store(false)

	ctx, cancel := context.WithCancel(parentCtx)
	o.cancelFunc = cancel
	defer cancel()

	stats := domain.ScanSessionStats{
		SessionID:     fmt.Sprintf("session-%d", time.Now().Unix()),
		StartTime:     time.Now().Unix(),
		TotalIngested: len(candidates),
		Deduplicated:  len(candidates),
		IsRunning:     true,
	}

	if o.callbacks.OnStatsUpdate != nil {
		o.callbacks.OnStatsUpdate(stats)
	}

	reports := make([]*domain.ProxyEvaluationReport, len(candidates))
	for i, cfg := range candidates {
		reports[i] = domain.NewReport(cfg)
		if o.callbacks.OnItemUpdate != nil {
			o.callbacks.OnItemUpdate(reports[i])
		}
	}

	// -------------------------------------------------------------
	// STAGE 1: Fast Reachability Handshake (Ping Triage)
	// -------------------------------------------------------------
	stage1Pool := NewWorkerPool(ctx, o.concurrency*2) // Fast pings can use higher concurrency
	var aliveReports []*domain.ProxyEvaluationReport
	var aliveMu sync.Mutex

	for _, rep := range reports {
		r := rep
		r.State = domain.StateStage1Testing
		if o.callbacks.OnItemUpdate != nil {
			o.callbacks.OnItemUpdate(r)
		}

		stage1Pool.Submit(func() {
			latency, err := FastPing(ctx, r.Config, o.pingTimeout)
			if err != nil || latency <= 0 {
				r.Stage1Passed = false
				r.Stage1LatencyMs = 0
				r.State = domain.StateStage1Dead
				atomic.AddInt64(&stats.Stage1Dead, 1)
			} else {
				r.Stage1Passed = true
				r.Stage1LatencyMs = latency
				r.State = domain.StateStage2Testing
				atomic.AddInt64(&stats.Stage1Alive, 1)

				aliveMu.Lock()
				aliveReports = append(aliveReports, r)
				aliveMu.Unlock()
			}

			if o.callbacks.OnItemUpdate != nil {
				o.callbacks.OnItemUpdate(r)
			}
			if o.callbacks.OnStatsUpdate != nil {
				o.callbacks.OnStatsUpdate(stats)
			}
		})
	}

	stage1Pool.Wait()

	if ctx.Err() != nil {
		stats.IsRunning = false
		if o.callbacks.OnStatsUpdate != nil {
			o.callbacks.OnStatsUpdate(stats)
		}
		return reports, ctx.Err()
	}

	// -------------------------------------------------------------
	// STAGE 2: Service Probing & IP Metadata Enrichment
	// -------------------------------------------------------------
	if len(aliveReports) > 0 && len(o.probes) > 0 {
		batchSize := 20
		if o.concurrency < batchSize {
			batchSize = o.concurrency
		}

		for start := 0; start < len(aliveReports); start += batchSize {
			if ctx.Err() != nil {
				break
			}

			end := start + batchSize
			if end > len(aliveReports) {
				end = len(aliveReports)
			}

			chunkReports := aliveReports[start:end]
			chunkConfigs := make([]domain.ProxyConfig, len(chunkReports))
			for i, r := range chunkReports {
				chunkConfigs[i] = r.Config
			}

			// Launch pool of local proxy tunnels
			instances, err := o.engine.StartPool(ctx, chunkConfigs)
			if err != nil {
				// Fallback: Try individual instances if pool creation failed (e.g. 1 bad config poisoned the batch)
				for _, r := range chunkReports {
					singleInstance, sErr := o.engine.StartPool(ctx, []domain.ProxyConfig{r.Config})
					if sErr != nil || len(singleInstance) == 0 {
						r.State = domain.StateComplete
						r.IsClean = false
						for _, p := range o.probes {
							r.ProbeResults[p.ID()] = domain.ProbeResult{
								ProbeID:      p.ID(),
								ProbeName:    p.Name(),
								Status:       domain.StatusFail,
								ErrorMessage: fmt.Sprintf("Tunnel initialization failed: %v", sErr),
								Diagnostic: &domain.DiagnosticInfo{
									Layer:            domain.LayerLocalEngine,
									InterceptionType: domain.InterceptLocalEngineError,
									Summary:          fmt.Sprintf("Xray local port bind failed: %v", sErr),
								},
							}
						}
						atomic.AddInt64(&stats.Stage2Failed, 1)
						if o.callbacks.OnItemUpdate != nil {
							o.callbacks.OnItemUpdate(r)
						}
						continue
					}

					inst := singleInstance[0]
					o.evaluateProxyProbes(ctx, r, inst.SOCKS5Endpoint(), &stats)
					_ = inst.Close()
				}
				continue
			}

			var chunkWg sync.WaitGroup
			for i, inst := range instances {
				rep := chunkReports[i]
				endpoint := inst.SOCKS5Endpoint()

				chunkWg.Add(1)
				go func(r *domain.ProxyEvaluationReport, ep string) {
					defer chunkWg.Done()
					o.evaluateProxyProbes(ctx, r, ep, &stats)
				}(rep, endpoint)
			}

			chunkWg.Wait()

			// Teardown instances
			for _, inst := range instances {
				_ = inst.Close()
			}
		}
	}

	stats.IsRunning = false
	if o.callbacks.OnStatsUpdate != nil {
		o.callbacks.OnStatsUpdate(stats)
	}

	return reports, nil
}

func (o *Orchestrator) evaluateProxyProbes(
	ctx context.Context,
	r *domain.ProxyEvaluationReport,
	endpoint string,
	stats *domain.ScanSessionStats,
) {
	var probeWg sync.WaitGroup
	var resultsMu sync.Mutex

	hasPass := false
	probeCtx := context.WithValue(ctx, "config_id", r.Config.ID)

	for _, p := range o.probes {
		probe := p
		probeWg.Add(1)

		go func() {
			defer probeWg.Done()
			res := probe.Evaluate(probeCtx, endpoint)

			resultsMu.Lock()
			r.ProbeResults[probe.ID()] = res
			if res.Status == domain.StatusPass {
				hasPass = true
			}

			// If probe provides IP information extraction
			if enricher, ok := probe.(interface {
				ExtractIPInfo(ctx context.Context, endpoint string) (domain.IPEnrichment, error)
			}); ok && res.Status == domain.StatusPass {
				if ipInfo, err := enricher.ExtractIPInfo(ctx, endpoint); err == nil {
					r.IPInfo = ipInfo
				}
			}
			resultsMu.Unlock()

			if o.callbacks.OnItemUpdate != nil {
				o.callbacks.OnItemUpdate(r)
			}
		}()
	}

	probeWg.Wait()

	r.State = domain.StateComplete
	r.IsClean = hasPass

	if hasPass {
		atomic.AddInt64(&stats.Stage2Clean, 1)
	} else {
		atomic.AddInt64(&stats.Stage2Failed, 1)
	}

	if o.callbacks.OnItemUpdate != nil {
		o.callbacks.OnItemUpdate(r)
	}
	if o.callbacks.OnStatsUpdate != nil {
		o.callbacks.OnStatsUpdate(*stats)
	}
}
