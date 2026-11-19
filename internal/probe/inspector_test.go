package probe

import (
	"errors"
	"strings"
	"testing"

	http "github.com/bogdanfinn/fhttp"

	"cips/internal/domain"
	"cips/internal/engine"
)

func TestInspectCloudflareChallenge(t *testing.T) {
	resp := &http.Response{
		StatusCode: 403,
		Header: http.Header{
			"Server": []string{"cloudflare"},
			"Cf-Ray": []string{"8e31a2bc4f1a2389-FRA"},
		},
	}
	body := []byte(`<!DOCTYPE html><html><head><title>Just a moment...</title></head><body>Checking your browser before accessing the website.</body></html>`)

	diag := InspectResponse(resp, body, nil, "https://gemini.google.com/app")

	if diag.Layer != domain.LayerTargetService {
		t.Errorf("expected TARGET_SERVICE layer, got %s", diag.Layer)
	}
	if diag.InterceptionType != domain.InterceptCloudflareChallenge {
		t.Errorf("expected CLOUDFLARE_CHALLENGE, got %s", diag.InterceptionType)
	}
	if diag.RayID != "8e31a2bc4f1a2389-FRA" {
		t.Errorf("expected ray id, got %s", diag.RayID)
	}
	if diag.PageTitle != "Just a moment..." {
		t.Errorf("expected title, got %s", diag.PageTitle)
	}
}

func TestInspectGoogleUnusualTraffic(t *testing.T) {
	resp := &http.Response{
		StatusCode: 302,
		Header: http.Header{
			"Location": []string{"https://www.google.com/sorry/index?continue=https://gemini.google.com/app"},
		},
	}

	diag := InspectResponse(resp, nil, nil, "https://gemini.google.com/app")

	if diag.InterceptionType != domain.InterceptGoogleUnusualTraffic {
		t.Errorf("expected GOOGLE_UNUSUAL_TRAFFIC, got %s", diag.InterceptionType)
	}
	if diag.FinalURL != "https://www.google.com/sorry/index?continue=https://gemini.google.com/app" {
		t.Errorf("expected sorry url, got %s", diag.FinalURL)
	}
}

func TestInspectOAuthSignInPass(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Header: http.Header{
			"Server": []string{"gws"},
		},
	}
	// Google OAuth sign-in contains access_denied in JavaScript bundles, but it is an authentic sign-in page!
	body := []byte(`<!DOCTYPE html><html><head><title>Sign in - Google Accounts</title><script>var error="access_denied";</script></head><body><h1>Sign in with your Google Account</h1></body></html>`)

	diag := InspectResponse(resp, body, nil, "https://gemini.google.com/app")

	if diag.InterceptionType != domain.InterceptNone {
		t.Errorf("expected NONE (Pass), got %s (Summary: %s)", diag.InterceptionType, diag.Summary)
	}
}

func TestInspectGeoblockRestricted(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{},
	}
	body := []byte(`<!DOCTYPE html><html><head><title>Gemini</title></head><body><h2>Gemini isn't supported in your country yet</h2></body></html>`)

	diag := InspectResponse(resp, body, nil, "https://gemini.google.com/app")

	if diag.InterceptionType != domain.InterceptGeoRestricted {
		t.Errorf("expected GEO_RESTRICTED, got %s", diag.InterceptionType)
	}
}

func TestInspectAIStudioAvailableRegionsRedirect(t *testing.T) {
	resp := &http.Response{
		StatusCode: 302,
		Header: http.Header{
			"Location": []string{"https://ai.google.dev/gemini-api/docs/available-regions"},
		},
	}

	diag := InspectResponse(resp, nil, nil, "https://aistudio.google.com/")

	if diag.InterceptionType != domain.InterceptGeoRestricted {
		t.Errorf("expected GEO_RESTRICTED for available-regions, got %s", diag.InterceptionType)
	}
}

func TestInspectLocalEngineRefused(t *testing.T) {
	reqErr := errors.New("dial tcp 127.0.0.1:20005: connectex: No connection could be made because the target machine actively refused it.")

	diag := InspectResponse(nil, nil, reqErr, "https://gemini.google.com/app")

	if diag.Layer != domain.LayerLocalEngine {
		t.Errorf("expected LOCAL_ENGINE layer, got %s", diag.Layer)
	}
	if diag.InterceptionType != domain.InterceptLocalEngineError {
		t.Errorf("expected LOCAL_ENGINE_ERROR, got %s", diag.InterceptionType)
	}
}

func TestInspectRealityKeyFailure(t *testing.T) {
	engine.GlobalEngineLog.Write([]byte("2026/11/19 14:00:00 [Warning] failed to parse reality config: invalid public key length\n"))
	reqErr := errors.New("dial tcp 127.0.0.1:20005: connectex: actively refused it")

	diag := InspectResponse(nil, nil, reqErr, "https://chatgpt.com/")

	if diag.InterceptionType != domain.InterceptBadRealityKey {
		t.Errorf("expected BAD_REALITY_KEY, got %s", diag.InterceptionType)
	}
	if !strings.Contains(diag.EngineLog, "reality") {
		t.Errorf("expected engine log to contain reality, got %s", diag.EngineLog)
	}
}
