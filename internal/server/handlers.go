package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"cips/internal/domain"
	"cips/internal/exporter"
	"cips/internal/ingestion"
	"cips/internal/pipeline"
)

// ServerContext encapsulates shared state for HTTP handlers.
type ServerContext struct {
	mu           sync.RWMutex
	candidates   []domain.ProxyConfig
	reports      map[string]*domain.ProxyEvaluationReport
	orchestrator *pipeline.Orchestrator
	broadcaster  *SSEBroadcaster
	lastSession  *domain.ScanSessionStats
}

// NewServerContext initializes state.
func NewServerContext(orch *pipeline.Orchestrator, bc *SSEBroadcaster) *ServerContext {
	return &ServerContext{
		reports:      make(map[string]*domain.ProxyEvaluationReport),
		orchestrator: orch,
		broadcaster:  bc,
	}
}

// HandleIngest processes file uploads, pasted text, or subscription URLs.
func (sc *ServerContext) HandleIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newConfigs []domain.ProxyConfig
	var totalParsed int

	contentType := r.Header.Get("Content-Type")

	// 1. File Upload (multipart)
	if stringsHasPrefix(contentType, "multipart/form-data") {
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid file upload: %v", err), http.StatusBadRequest)
			return
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "failed to read uploaded file", http.StatusInternalServerError)
			return
		}
		newConfigs, totalParsed = ingestion.ExtractAndDeduplicate(string(content))

	} else {
		// 2. JSON Body (raw text or subscription URL)
		var payload struct {
			Text            string `json:"text"`
			SubscriptionURL string `json:"subscription_url"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}

		if payload.SubscriptionURL != "" {
			ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
			defer cancel()

			cfgs, total, err := ingestion.FetchSubscription(ctx, payload.SubscriptionURL)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
				return
			}
			newConfigs = cfgs
			totalParsed = total
		} else if payload.Text != "" {
			newConfigs, totalParsed = ingestion.ExtractAndDeduplicate(payload.Text)
		}
	}

	sc.mu.Lock()
	seen := make(map[string]bool)
	for _, c := range sc.candidates {
		seen[c.ID] = true
	}

	for _, c := range newConfigs {
		if !seen[c.ID] {
			seen[c.ID] = true
			sc.candidates = append(sc.candidates, c)
			sc.reports[c.ID] = domain.NewReport(c)
		}
	}
	allCount := len(sc.candidates)
	sc.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"total_parsed":      totalParsed,
		"unique_candidates": allCount,
		"configs":           newConfigs,
	})
}

// HandleScanStart initiates the background two-stage scan.
func (sc *ServerContext) HandleScanStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Concurrency     int `json:"concurrency"`
		Stage1TimeoutMs int `json:"stage1_timeout_ms"`
		Stage2TimeoutMs int `json:"stage2_timeout_ms"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	sc.mu.RLock()
	candidates := make([]domain.ProxyConfig, len(sc.candidates))
	copy(candidates, sc.candidates)
	sc.mu.RUnlock()

	if len(candidates) == 0 {
		http.Error(w, "no candidates to scan", http.StatusBadRequest)
		return
	}

	go func() {
		reports, err := sc.orchestrator.Execute(context.Background(), candidates)
		if err != nil {
			return
		}

		sc.mu.Lock()
		for _, rep := range reports {
			sc.reports[rep.Config.ID] = rep
		}
		sc.mu.Unlock()
	}()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "started",
		"targets": len(candidates),
	})
}

// HandleScanCancel aborts active scanning.
func (sc *ServerContext) HandleScanCancel(w http.ResponseWriter, r *http.Request) {
	sc.orchestrator.Cancel()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "cancelled"})
}

// HandleExport returns clean subscription text or JSON telemetry.
func (sc *ServerContext) HandleExport(w http.ResponseWriter, r *http.Request) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	format := r.URL.Query().Get("format")
	requiredProbe := r.URL.Query().Get("required_probe")

	if format == "json" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=cips-session-%d.json", time.Now().Unix()))
		_ = json.NewEncoder(w).Encode(sc.reports)
		return
	}

	reportList := make([]*domain.ProxyEvaluationReport, 0, len(sc.reports))
	for _, rep := range sc.reports {
		reportList = append(reportList, rep)
	}

	subContent := exporter.GenerateCleanSubscription(reportList, requiredProbe)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=clean_sub_%d.txt", time.Now().Unix()))
	_, _ = w.Write([]byte(subContent))
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}
