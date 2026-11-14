package exporter

import (
	"strings"
	"testing"

	"cips/internal/domain"
)

func TestGenerateCleanSubscription(t *testing.T) {
	reports := []*domain.ProxyEvaluationReport{
		{
			Config: domain.ProxyConfig{
				RawURI:   "vless://uuid@1.2.3.4:443?type=tcp#Node1",
				Protocol: domain.ProtocolVLESS,
				Server:   "1.2.3.4",
				Port:     443,
				Remark:   "Node1",
			},
			IsClean:         true,
			Stage1Passed:    true,
			Stage1LatencyMs: 120,
			IPInfo: domain.IPEnrichment{
				CountryCode: "RU",
				ISP:         "Timeweb",
			},
			ProbeResults: map[string]domain.ProbeResult{
				"gemini": {
					Status:    domain.StatusPass,
					HTTPCode:  200,
					LatencyMs: 350,
				},
			},
		},
		{
			Config: domain.ProxyConfig{
				RawURI:   "vless://bad@5.6.7.8:443#DeadNode",
				Protocol: domain.ProtocolVLESS,
				Server:   "5.6.7.8",
				Port:     443,
			},
			IsClean:      false,
			Stage1Passed: false,
		},
	}

	sub := GenerateCleanSubscription(reports, "")
	if !strings.Contains(sub, "1.2.3.4:443") {
		t.Errorf("clean node missing from subscription")
	}
	if strings.Contains(sub, "5.6.7.8:443") {
		t.Errorf("dead node should have been filtered out")
	}
	if !strings.Contains(sub, "Timeweb") {
		t.Errorf("expected ISP smart tag in remark")
	}
	if !strings.Contains(sub, "Gemini") {
		t.Errorf("expected Gemini pass tag in remark")
	}
}
