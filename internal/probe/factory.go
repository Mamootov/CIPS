package probe

import (
	"cips/internal/domain"
)

// BuildProbeSuite constructs the active list of service probes based on IDs and custom definitions.
func BuildProbeSuite(selectedIDs []string, customDefs []domain.ProbeDefinition, timeoutMs int) []domain.IProbe {
	if timeoutMs <= 0 {
		timeoutMs = 8000
	}

	var probes []domain.IProbe
	selectedMap := make(map[string]bool)
	for _, id := range selectedIDs {
		selectedMap[id] = true
	}

	// If no specific service IDs provided, default to Gemini + AI Studio
	if len(selectedIDs) == 0 && len(customDefs) == 0 {
		selectedMap["gemini"] = true
		selectedMap["aistudio"] = true
	}

	// 1. Built-in AI Presets
	if selectedMap["gemini"] {
		probes = append(probes, NewGeminiProbe(timeoutMs))
	}
	if selectedMap["aistudio"] {
		probes = append(probes, NewAIStudioProbe(timeoutMs))
	}
	if selectedMap["chatgpt"] {
		probes = append(probes, NewChatGPTProbe(timeoutMs))
	}
	if selectedMap["claude"] {
		probes = append(probes, NewClaudeProbe(timeoutMs))
	}

	// 2. User-Defined Custom Probes
	for _, def := range customDefs {
		if def.TargetURL != "" {
			probes = append(probes, NewCustomProbe(def))
		}
	}

	// 3. Always append IP Intelligence as auxiliary metadata enrichment
	probes = append(probes, NewIPIntelProbe(timeoutMs))

	return probes
}
