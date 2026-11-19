package probe

import (
	"testing"
)

func TestChatGPTProbeIdentifiers(t *testing.T) {
	p := NewChatGPTProbe(5000)
	if p.ID() != "chatgpt" {
		t.Errorf("expected id chatgpt, got %s", p.ID())
	}
	if p.Name() != "ChatGPT" {
		t.Errorf("expected name ChatGPT, got %s", p.Name())
	}
}

func TestClaudeProbeIdentifiers(t *testing.T) {
	p := NewClaudeProbe(5000)
	if p.ID() != "claude" {
		t.Errorf("expected id claude, got %s", p.ID())
	}
	if p.Name() != "Claude" {
		t.Errorf("expected name Claude, got %s", p.Name())
	}
}

func TestBuildProbeSuiteSelection(t *testing.T) {
	// Select only ChatGPT and Claude
	probes := BuildProbeSuite([]string{"chatgpt", "claude"}, nil, 5000)

	foundChatGPT := false
	foundClaude := false
	foundGemini := false
	foundIPLeak := false

	for _, p := range probes {
		if p.ID() == "chatgpt" {
			foundChatGPT = true
		}
		if p.ID() == "claude" {
			foundClaude = true
		}
		if p.ID() == "gemini" {
			foundGemini = true
		}
		if p.ID() == "ipleak" {
			foundIPLeak = true
		}
	}

	if !foundChatGPT {
		t.Errorf("expected ChatGPT in probe suite")
	}
	if !foundClaude {
		t.Errorf("expected Claude in probe suite")
	}
	if foundGemini {
		t.Errorf("did not expect Gemini when only chatgpt and claude selected")
	}
	if !foundIPLeak {
		t.Errorf("expected IPLeak auxiliary probe in suite")
	}
}
