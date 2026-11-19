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

// ChatGPTProbe evaluates reachability and geoblock status of OpenAI ChatGPT.
type ChatGPTProbe struct {
	timeout time.Duration
}

// NewChatGPTProbe constructs an OpenAI ChatGPT reachability probe.
func NewChatGPTProbe(timeoutMs int) *ChatGPTProbe {
	if timeoutMs <= 0 {
		timeoutMs = 8000
	}
	return &ChatGPTProbe{
		timeout: time.Duration(timeoutMs) * time.Millisecond,
	}
}

func (p *ChatGPTProbe) ID() string {
	return "chatgpt"
}

func (p *ChatGPTProbe) Name() string {
	return "ChatGPT"
}

func (p *ChatGPTProbe) Evaluate(ctx context.Context, proxyEndpoint string) domain.ProbeResult {
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

	targetURL := "https://chatgpt.com/"
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

	// Additional OpenAI specific block signatures
	if resp != nil {
		bodyLower := strings.ToLower(string(bodyBytes))
		if strings.Contains(bodyLower, "unsupported_country") || strings.Contains(bodyLower, "not available in your country") {
			diag.Layer = domain.LayerTargetService
			diag.InterceptionType = domain.InterceptGeoRestricted
			diag.Summary = "OpenAI regional restriction (Country unsupported)"
		} else if strings.Contains(bodyLower, "you have been blocked") || strings.Contains(bodyLower, "access denied") {
			diag.Layer = domain.LayerTargetService
			diag.InterceptionType = domain.InterceptCloudflareChallenge
			diag.Summary = "OpenAI security block (IP blocked by Cloudflare/WAF)"
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
