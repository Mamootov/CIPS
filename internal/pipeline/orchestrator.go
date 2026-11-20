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
	engine          domain.IProxyEngine
	probes          []domain.IProbe
	speedProbe      domain.IProbe
	enableSpeedtest bool
	callbacks       OrchestratorCallbacks
	concurrency     int
	pingTimeout     int
	probeTimeout    int
	cancelFunc      context.CancelFunc
	isRunning       atomic.Bool
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

// SetSpeedOptions configures conditional download speed measurement.
func (o *Orchestrator) SetSpeedOptions(enable bool, probe domain.IProbe) {
	o.enableSpeedtest = enable
	o.speedProbe = probe
}

// SetProbes dynamically updates active probes.
func (o *Orchestrator) SetProbes(probes []domain.IProbe) {
	o.probes = probes
}

// SetConcurrency dynamically updates the worker concurrency.
func (o *Orchestrator) SetConcurrency(c int) {
	if c > 0 {
		o.concurrency = c
	}
}

// SetTimeouts updates Stage 1 ping and Stage 2 probe timeouts.
func (o *Orchestrator) SetTimeouts(p1Timeout, p2Timeout int) {
	if p1Timeout > 0 {
		o.pingTimeout = p1Timeout
	}
	if p2Timeout > 0 {
		o.probeTimeout = p2Timeout
	}
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
				r.State = domain.StateComplete
				r.IsClean = false
				for _, p := range o.probes {
					r.ProbeResults[p.ID()] = domain.ProbeResult{
						ProbeID:      p.ID(),
						ProbeName:    p.Name(),
						Status:       domain.StatusFail,
						ErrorMessage: "Ping handshake timed out or node unreachable",
						Diagnostic: &domain.DiagnosticInfo{
							Layer:            domain.LayerUpstreamProxy,
							InterceptionType: domain.InterceptUpstreamRefused,
							Summary:          "TCP/TLS handshake failed during Stage 1 triage",
						},
					}
				}
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
				o.callbacks.OnItemUpdate(r.Clone())
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
		batchSize := o.concurrency
		if batchSize <= 0 {
			batchSize = 20
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
				// Fallback: Try individual instances in parallel if pool creation failed (e.g. 1 bad config poisoned the batch)
				var fbWg sync.WaitGroup
				for _, r := range chunkReports {
					fbWg.Add(1)
					go func(rep *domain.ProxyEvaluationReport) {
						defer fbWg.Done()
						singleInstance, sErr := o.engine.StartPool(ctx, []domain.ProxyConfig{rep.Config})
						if sErr != nil || len(singleInstance) == 0 {
							rep.State = domain.StateComplete
							rep.IsClean = false
							for _, p := range o.probes {
								rep.ProbeResults[p.ID()] = domain.ProbeResult{
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
								o.callbacks.OnItemUpdate(rep)
							}
							return
						}

						inst := singleInstance[0]
						defer inst.Close()
						o.evaluateProxyProbes(ctx, rep, inst.SOCKS5Endpoint(), &stats)
					}(r)
				}
				fbWg.Wait()
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

	hasAIPass := false

	timeoutMs := o.probeTimeout
	if timeoutMs <= 0 {
		timeoutMs = 8000
	}
	probeCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()
	probeCtx = context.WithValue(probeCtx, "config_id", r.Config.ID)

	for _, p := range o.probes {
		probe := p
		probeWg.Add(1)

		go func() {
			defer probeWg.Done()
			res := probe.Evaluate(probeCtx, endpoint)

			var extractedIP *domain.IPEnrichment
			// If probe provides IP information extraction
			if enricher, ok := probe.(interface {
				ExtractIPInfo(ctx context.Context, endpoint string) (domain.IPEnrichment, error)
			}); ok && res.Status == domain.StatusPass {
				if ipInfo, err := enricher.ExtractIPInfo(probeCtx, endpoint); err == nil {
					extractedIP = &ipInfo
				}
			}

			resultsMu.Lock()
			r.ProbeResults[probe.ID()] = res
			// Only actual AI service probes qualify a node for AI-Clean status (metadata/speed probes are auxiliary)
			if probe.ID() != "ipleak" && probe.ID() != "speedtest" && res.Status == domain.StatusPass {
				hasAIPass = true
			}
			if extractedIP != nil {
				r.IPInfo = *extractedIP
			}
			snapshot := r.Clone()
			resultsMu.Unlock()

			if o.callbacks.OnItemUpdate != nil {
				o.callbacks.OnItemUpdate(snapshot)
			}
		}()
	}

	probeWg.Wait()

	// Stage 2B: Conditional Speedtest (ONLY if at least one AI service passed!)
	if o.enableSpeedtest && hasAIPass && o.speedProbe != nil {
		speedRes := o.speedProbe.Evaluate(probeCtx, endpoint)
		resultsMu.Lock()
		r.ProbeResults[o.speedProbe.ID()] = speedRes
		if speedRes.Status == domain.StatusPass {
			var mbps float64
			if _, sErr := fmt.Sscanf(speedRes.ErrorMessage, "%f Mbps", &mbps); sErr == nil {
				r.DownloadSpeedMbps = mbps
			}
		}
		snapshot := r.Clone()
		resultsMu.Unlock()

		if o.callbacks.OnItemUpdate != nil {
			o.callbacks.OnItemUpdate(snapshot)
		}
	}

	resultsMu.Lock()
	r.State = domain.StateComplete
	r.IsClean = hasAIPass

	if hasAIPass {
		atomic.AddInt64(&stats.Stage2Clean, 1)
	} else {
		atomic.AddInt64(&stats.Stage2Failed, 1)
	}
	finalSnapshot := r.Clone()
	resultsMu.Unlock()

	if o.callbacks.OnItemUpdate != nil {
		o.callbacks.OnItemUpdate(finalSnapshot)
	}
	if o.callbacks.OnStatsUpdate != nil {
		o.callbacks.OnStatsUpdate(*stats)
	}
}
