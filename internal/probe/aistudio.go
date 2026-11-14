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

// AIStudioProbe validates accessibility to Google AI Studio.
type AIStudioProbe struct {
	timeout time.Duration
}

// NewAIStudioProbe constructs a Google AI Studio probe with configured timeout.
func NewAIStudioProbe(timeoutMs int) *AIStudioProbe {
	if timeoutMs <= 0 {
		timeoutMs = 8000
	}
	return &AIStudioProbe{
		timeout: time.Duration(timeoutMs) * time.Millisecond,
	}
}

func (p *AIStudioProbe) ID() string {
	return "aistudio"
}

func (p *AIStudioProbe) Name() string {
	return "Google AI Studio"
}

func (p *AIStudioProbe) Evaluate(ctx context.Context, proxyEndpoint string) domain.ProbeResult {
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

	req, err := http.NewRequestWithContext(ctx, "GET", "https://aistudio.google.com/", nil)
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

	blockedTokens := []string{
		"location_error",
		"unsupported-country",
		"not available in your country",
		"access denied",
		"user location not supported",
	}

	for _, token := range blockedTokens {
		if strings.Contains(bodyStr, token) {
			return domain.ProbeResult{
				ProbeID:      p.ID(),
				ProbeName:    p.Name(),
				Status:       domain.StatusFail,
				HTTPCode:     resp.StatusCode,
				LatencyMs:    latency,
				ErrorMessage: "Region blocked by Google AI Studio",
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
