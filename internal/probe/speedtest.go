package probe

import (
	"context"
	"fmt"
	"math"
	"time"

	http "github.com/bogdanfinn/fhttp"

	"cips/internal/domain"
)

// SpeedtestProbe performs low-bandwidth download speed measurement via Cloudflare Anycast edge.
type SpeedtestProbe struct {
	maxBytes int64
	timeout  time.Duration
}

// NewSpeedtestProbe creates a speed test probe with bounded byte payload and strict deadline.
func NewSpeedtestProbe(maxBytes int, timeoutMs int) *SpeedtestProbe {
	if maxBytes <= 0 {
		maxBytes = 3 * 1024 * 1024 // 3 MB default
	}
	if timeoutMs <= 0 {
		timeoutMs = 3000 // 3 seconds default
	}
	return &SpeedtestProbe{
		maxBytes: int64(maxBytes),
		timeout:  time.Duration(timeoutMs) * time.Millisecond,
	}
}

func (p *SpeedtestProbe) ID() string {
	return "speedtest"
}

func (p *SpeedtestProbe) Name() string {
	return "Download Speed"
}

func (p *SpeedtestProbe) Evaluate(ctx context.Context, proxyEndpoint string) domain.ProbeResult {
	start := time.Now()

	mbps, err := p.MeasureSpeed(ctx, proxyEndpoint)
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

	return domain.ProbeResult{
		ProbeID:      p.ID(),
		ProbeName:    p.Name(),
		Status:       domain.StatusPass,
		HTTPCode:     200,
		LatencyMs:    latency,
		ErrorMessage: fmt.Sprintf("%.2f Mbps", mbps),
	}
}

// MeasureSpeed downloads up to maxBytes or until 3-second deadline expires, returning throughput in Mbps.
func (p *SpeedtestProbe) MeasureSpeed(ctx context.Context, proxyEndpoint string) (float64, error) {
	client, err := BuildTLSClient(proxyEndpoint, p.timeout+1*time.Second)
	if err != nil {
		return 0, err
	}

	speedCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	downloadURL := fmt.Sprintf("https://speed.cloudflare.com/__down?bytes=%d", p.maxBytes)
	req, err := http.NewRequestWithContext(speedCtx, "GET", downloadURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 Chrome/124.0.0.0")

	startTime := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("speedtest HTTP %d", resp.StatusCode)
	}

	// Stream chunks into discard up to maxBytes
	buf := make([]byte, 32*1024)
	var totalBytes int64

	for {
		if speedCtx.Err() != nil {
			break
		}
		n, rErr := resp.Body.Read(buf)
		if n > 0 {
			totalBytes += int64(n)
		}
		if rErr != nil {
			break
		}
	}

	elapsed := time.Since(startTime).Seconds()
	if elapsed <= 0 || totalBytes < 10*1024 { // Minimum 10 KB to calculate meaningful speed
		return 0, fmt.Errorf("insufficient data transferred (under 10KB)")
	}

	// Mbps = (Bytes * 8) / (ElapsedSeconds * 1,000,000)
	mbps := (float64(totalBytes) * 8.0) / (elapsed * 1000000.0)
	mbps = math.Round(mbps*100) / 100

	return mbps, nil
}
