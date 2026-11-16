package probe

import (
	"context"
	"fmt"
	"io"
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

	// 1. Check Google's first-party localized region for this proxy IP
	if country, err := DetectGoogleRegion(ctx, client, proxyEndpoint); err == nil {
		if bannedName, isBanned := GoogleBannedCountries[country]; isBanned {
			diag := domain.DiagnosticInfo{
				Layer:            domain.LayerTargetService,
				InterceptionType: domain.InterceptGeoRestricted,
				FinalURL:         "https://gemini.google.com/",
				HTTPStatus:       403,
				Summary:          fmt.Sprintf("Google localized IP to %s (%s) — Region blocked by Gemini policy", country, bannedName),
			}
			return domain.ProbeResult{
				ProbeID:      p.ID(),
				ProbeName:    p.Name(),
				Status:       domain.StatusFail,
				HTTPCode:     403,
				LatencyMs:    time.Since(start).Milliseconds(),
				ErrorMessage: diag.Summary,
				Diagnostic:   &diag,
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://gemini.google.com/", nil)
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

	var bodyBytes []byte
	if resp != nil && resp.Body != nil {
		bodyBytes, _ = io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		_ = resp.Body.Close()
	}

	diag := InspectResponse(resp, bodyBytes, err, "https://gemini.google.com/")

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
