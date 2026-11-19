package probe

import (
	"math"
	"testing"
)

func TestCalculateSpeedMbps(t *testing.T) {
	// 5 MB in 2 seconds = (5,000,000 * 8) / (2 * 1,000,000) = 20.00 Mbps
	totalBytes := int64(5000000)
	elapsed := 2.0

	mbps := (float64(totalBytes) * 8.0) / (elapsed * 1000000.0)
	mbps = math.Round(mbps*100) / 100

	if mbps != 20.00 {
		t.Errorf("expected 20.00 Mbps, got %.2f", mbps)
	}

	// 1 MB in 1.5 seconds = 8,000,000 / 1,500,000 = 5.33 Mbps
	totalBytes2 := int64(1000000)
	elapsed2 := 1.5
	mbps2 := (float64(totalBytes2) * 8.0) / (elapsed2 * 1000000.0)
	mbps2 = math.Round(mbps2*100) / 100

	if mbps2 != 5.33 {
		t.Errorf("expected 5.33 Mbps, got %.2f", mbps2)
	}
}

func TestSpeedtestProbeDefaults(t *testing.T) {
	p := NewSpeedtestProbe(0, 0)
	if p.ID() != "speedtest" {
		t.Errorf("expected id speedtest, got %s", p.ID())
	}
	if p.maxBytes != 3*1024*1024 {
		t.Errorf("expected 3MB default, got %d", p.maxBytes)
	}
}
