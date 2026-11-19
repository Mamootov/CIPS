package probe

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"

	"cips/internal/domain"
)

// ClaudeProbe evaluates reachability and regional unlock for Anthropic Claude.
type ClaudeProbe struct {
	timeout time.Duration
}

// NewClaudeProbe constructs an Anthropic Claude reachability probe.
func NewClaudeProbe(timeoutMs int) *ClaudeProbe {
	if timeoutMs <= 0 {
		timeoutMs = 8000
	}
	return &ClaudeProbe{
		timeout: time.Duration(timeoutMs) * time.Millisecond,
	}
}

func (p *ClaudeProbe) ID() string {
	return "claude"
}

func (p *ClaudeProbe) Name() string {
	return "Claude"
}

func (p *ClaudeProbe) Evaluate(ctx context.Context, proxyEndpoint string) domain.ProbeResult {
	start := time.Now()

	client, err := BuildTLSClient(proxyEndpoint, p.timeout)
	if err != nil {
		return domain.ProbeResult{
			ProbeID:      p.ID(),
			ProbeName:    p.Name(),
			Status:       domain.StatusFail,
			LatencyMs:    time.Since(start).Milliseconds(),
			ErrorMessage: fmt.Sprintf("tls client error: %v", err),
		}
	}

	targetURL := "https://claude.ai/login"
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return domain.ProbeResult{
			ProbeID:      p.ID(),
			ProbeName:    p.Name(),
			Status:       domain.StatusFail,
			ErrorMessage: err.Error(),
		}
	}

	ApplyChromeHeaders(req)

	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()

	var bodyBytes []byte
	if resp != nil && resp.Body != nil {
		bodyBytes, _ = io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		_ = resp.Body.Close()
	}

	diag := InspectResponse(resp, bodyBytes, err, targetURL)

	// Additional Anthropic specific regional and WAF signatures
	if resp != nil {
		loc := resp.Header.Get("Location")
		if strings.Contains(loc, "supported-countries") || strings.Contains(loc, "not-supported") {
			diag.Layer = domain.LayerTargetService
			diag.InterceptionType = domain.InterceptGeoRestricted
			diag.Summary = "Anthropic Claude regional restriction (Redirected to supported-countries)"
		}

		bodyLower := strings.ToLower(string(bodyBytes))
		if strings.Contains(bodyLower, "claude is not available in your region") || strings.Contains(bodyLower, "supported countries") {
			diag.Layer = domain.LayerTargetService
			diag.InterceptionType = domain.InterceptGeoRestricted
			diag.Summary = "Anthropic Claude regional restriction (Region not supported)"
		}
	}

	if diag.InterceptionType != domain.InterceptNone {
		httpCode := 0
		if resp != nil {
			httpCode = resp.StatusCode
		}
		return domain.ProbeResult{
			ProbeID:      p.ID(),
			ProbeName:    p.Name(),
			Status:       domain.StatusFail,
			HTTPCode:     httpCode,
			LatencyMs:    latency,
			ErrorMessage: diag.Summary,
			Diagnostic:   &diag,
		}
	}

	return domain.ProbeResult{
		ProbeID:    p.ID(),
		ProbeName:  p.Name(),
		Status:     domain.StatusPass,
		HTTPCode:   resp.StatusCode,
		LatencyMs:  latency,
		Diagnostic: &diag,
	}
}
