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

// GeminiProbe validates accessibility to Google Gemini.
type GeminiProbe struct {
	timeout time.Duration
}

// NewGeminiProbe constructs a Gemini probe with configured timeout.
func NewGeminiProbe(timeoutMs int) *GeminiProbe {
	if timeoutMs <= 0 {
		timeoutMs = 8000
	}
	return &GeminiProbe{
		timeout: time.Duration(timeoutMs) * time.Millisecond,
	}
}

func (p *GeminiProbe) ID() string {
	return "gemini"
}

func (p *GeminiProbe) Name() string {
	return "Google Gemini"
}

func (p *GeminiProbe) Evaluate(ctx context.Context, proxyEndpoint string) domain.ProbeResult {
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

	req, err := http.NewRequestWithContext(ctx, "GET", "https://gemini.google.com/app", nil)
	if err != nil {
		return domain.ProbeResult{
			ProbeID:      p.ID(),
			ProbeName:    p.Name(),
			Status:       domain.StatusFail,
			ErrorMessage: err.Error(),
		}
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")

	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return domain.ProbeResult{
			ProbeID:      p.ID(),
			ProbeName:    p.Name(),
			Status:       domain.StatusFail,
			LatencyMs:    latency,
			ErrorMessage: err.Error(),
		}
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	bodyStr := strings.ToLower(string(bodyBytes))

	// Status 200 (main app) or 302 (redirect to accounts.google.com)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return domain.ProbeResult{
			ProbeID:      p.ID(),
			ProbeName:    p.Name(),
			Status:       domain.StatusFail,
			HTTPCode:     resp.StatusCode,
			LatencyMs:    latency,
			ErrorMessage: fmt.Sprintf("HTTP %d unexpected status", resp.StatusCode),
		}
	}

	// Assert absence of location restriction tokens
	blockedTokens := []string{
		"location_error",
		"isn't supported in your country",
		"unsupported country",
		"access_denied",
	}

	for _, token := range blockedTokens {
		if strings.Contains(bodyStr, token) {
			return domain.ProbeResult{
				ProbeID:      p.ID(),
				ProbeName:    p.Name(),
				Status:       domain.StatusFail,
				HTTPCode:     resp.StatusCode,
				LatencyMs:    latency,
				ErrorMessage: "Region blocked by Google Gemini",
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
