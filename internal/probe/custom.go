package probe

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"

	"cips/internal/domain"
)

// CustomProbe evaluates a user-defined HTTP endpoint and assertion rule.
type CustomProbe struct {
	def     domain.ProbeDefinition
	timeout time.Duration
}

// NewCustomProbe creates a probe based on user definition.
func NewCustomProbe(def domain.ProbeDefinition) *CustomProbe {
	timeout := time.Duration(def.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 8000 * time.Millisecond
	}
	return &CustomProbe{
		def:     def,
		timeout: timeout,
	}
}

func (p *CustomProbe) ID() string {
	return p.def.ID
}

func (p *CustomProbe) Name() string {
	return p.def.Name
}

func (p *CustomProbe) Evaluate(ctx context.Context, proxyEndpoint string) domain.ProbeResult {
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

	method := strings.ToUpper(p.def.Method)
	if method == "" {
		method = "GET"
	}

	req, err := http.NewRequestWithContext(ctx, method, p.def.TargetURL, nil)
	if err != nil {
		return domain.ProbeResult{
			ProbeID:      p.ID(),
			ProbeName:    p.Name(),
			Status:       domain.StatusFail,
			ErrorMessage: err.Error(),
		}
	}

	// Set custom headers or default Chrome headers
	if len(p.def.Headers) > 0 {
		for k, v := range p.def.Headers {
			req.Header.Set(k, v)
		}
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/124.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "*/*")
	}

	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()

	var bodyBytes []byte
	if resp != nil && resp.Body != nil {
		bodyBytes, _ = io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		_ = resp.Body.Close()
	}

	diag := InspectResponse(resp, bodyBytes, err, p.def.TargetURL)
	if diag.InterceptionType == domain.InterceptLocalEngineError || diag.InterceptionType == domain.InterceptUpstreamTimeout {
		return domain.ProbeResult{
			ProbeID:      p.ID(),
			ProbeName:    p.Name(),
			Status:       domain.StatusFail,
			LatencyMs:    latency,
			ErrorMessage: diag.Summary,
			Diagnostic:   &diag,
		}
	}

	if err != nil {
		return domain.ProbeResult{
			ProbeID:      p.ID(),
			ProbeName:    p.Name(),
			Status:       domain.StatusFail,
			LatencyMs:    latency,
			ErrorMessage: err.Error(),
			Diagnostic:   &diag,
		}
	}

	expectedStatus := p.def.ExpectedStatus
	if expectedStatus <= 0 {
		expectedStatus = http.StatusOK
	}

	if resp.StatusCode != expectedStatus {
		return domain.ProbeResult{
			ProbeID:      p.ID(),
			ProbeName:    p.Name(),
			Status:       domain.StatusFail,
			HTTPCode:     resp.StatusCode,
			LatencyMs:    latency,
			ErrorMessage: fmt.Sprintf("HTTP status %d does not match expected %d", resp.StatusCode, expectedStatus),
		}
	}

	if p.def.RequiredBody != "" {
		bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		if err != nil {
			return domain.ProbeResult{
				ProbeID:      p.ID(),
				ProbeName:    p.Name(),
				Status:       domain.StatusFail,
				HTTPCode:     resp.StatusCode,
				LatencyMs:    latency,
				ErrorMessage: "failed to read response body",
			}
		}

		matched, err := regexp.Match(p.def.RequiredBody, bodyBytes)
		if err != nil || !matched {
			return domain.ProbeResult{
				ProbeID:      p.ID(),
				ProbeName:    p.Name(),
				Status:       domain.StatusFail,
				HTTPCode:     resp.StatusCode,
				LatencyMs:    latency,
				ErrorMessage: "required body token assertion failed",
			}
		}
	}

	return domain.ProbeResult{
		ProbeID:   p.ID(),
		ProbeName: p.Name(),
		Status:    domain.StatusPass,
		HTTPCode:  resp.StatusCode,
		LatencyMs: latency,
	}
}
