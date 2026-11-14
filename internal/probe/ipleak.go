package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"time"

	http "github.com/bogdanfinn/fhttp"

	"cips/internal/domain"
)

// IPLeakProbe retrieves exit IP and ISP intelligence through the proxy tunnel.
type IPLeakProbe struct {
	timeout time.Duration
}

// NewIPLeakProbe creates an IPLeak metadata enricher.
func NewIPLeakProbe(timeoutMs int) *IPLeakProbe {
	if timeoutMs <= 0 {
		timeoutMs = 8000
	}
	return &IPLeakProbe{
		timeout: time.Duration(timeoutMs) * time.Millisecond,
	}
}

func (p *IPLeakProbe) ID() string {
	return "ipleak"
}

func (p *IPLeakProbe) Name() string {
	return "IPLeak Info"
}

type ipLeakJSONResponse struct {
	IP          string `json:"ip"`
	CountryName string `json:"country_name"`
	CountryCode string `json:"country_code"`
	ISPName     string `json:"isp_name"`
	ASNumber    string `json:"as_number"`
}

func (p *IPLeakProbe) Evaluate(ctx context.Context, proxyEndpoint string) domain.ProbeResult {
	start := time.Now()

	// Add slight random jitter (100ms - 250ms) to protect public endpoint from concurrent bursts
	jitter := time.Duration(100+rand.Intn(150)) * time.Millisecond
	select {
	case <-time.After(jitter):
	case <-ctx.Done():
		return domain.ProbeResult{
			ProbeID:      p.ID(),
			ProbeName:    p.Name(),
			Status:       domain.StatusFail,
			ErrorMessage: "Context canceled",
		}
	}

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

	req, err := http.NewRequestWithContext(ctx, "GET", "https://ipleak.net/json/", nil)
	if err != nil {
		return domain.ProbeResult{
			ProbeID:      p.ID(),
			ProbeName:    p.Name(),
			Status:       domain.StatusFail,
			ErrorMessage: err.Error(),
		}
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/124.0.0.0")
	req.Header.Set("Accept", "application/json")

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

	if resp.StatusCode != http.StatusOK {
		return domain.ProbeResult{
			ProbeID:      p.ID(),
			ProbeName:    p.Name(),
			Status:       domain.StatusFail,
			HTTPCode:     resp.StatusCode,
			LatencyMs:    latency,
			ErrorMessage: fmt.Sprintf("ipleak HTTP %d", resp.StatusCode),
		}
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return domain.ProbeResult{
			ProbeID:      p.ID(),
			ProbeName:    p.Name(),
			Status:       domain.StatusFail,
			LatencyMs:    latency,
			ErrorMessage: "Failed to read ipleak body",
		}
	}

	var data ipLeakJSONResponse
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return domain.ProbeResult{
			ProbeID:      p.ID(),
			ProbeName:    p.Name(),
			Status:       domain.StatusFail,
			LatencyMs:    latency,
			ErrorMessage: "Malformed JSON from ipleak",
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

// ExtractIPInfo queries ipleak directly and returns structured domain.IPEnrichment.
func (p *IPLeakProbe) ExtractIPInfo(ctx context.Context, proxyEndpoint string) (domain.IPEnrichment, error) {
	client, err := BuildTLSClient(proxyEndpoint, p.timeout)
	if err != nil {
		return domain.IPEnrichment{Success: false}, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://ipleak.net/json/", nil)
	if err != nil {
		return domain.IPEnrichment{Success: false}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/124.0.0.0")

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return domain.IPEnrichment{Success: false}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.IPEnrichment{Success: false}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 32*1024))
	if err != nil {
		return domain.IPEnrichment{Success: false}, err
	}

	var data ipLeakJSONResponse
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return domain.IPEnrichment{Success: false}, err
	}

	return domain.IPEnrichment{
		ExitIP:      data.IP,
		CountryCode: data.CountryCode,
		CountryName: data.CountryName,
		ISP:         data.ISPName,
		ASN:         data.ASNumber,
		QueryTimeMs: time.Since(start).Milliseconds(),
		Success:     true,
	}, nil
}
