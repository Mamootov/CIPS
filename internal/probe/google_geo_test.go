package probe

import (
	"testing"
)

func TestParseGoogleRegionFromHTML(t *testing.T) {
	htmlSample := `<!DOCTYPE html><html xmlns="http://www.w3.org/1999/xhtml" lang="en-DE"><head><title>Search settings</title></head><body>Preferences</body></html>`
	if m := langRegex.FindStringSubmatch(htmlSample); len(m) > 1 {
		if m[1] != "DE" {
			t.Errorf("expected country DE, got %s", m[1])
		}
	} else {
		t.Errorf("failed to match country from lang attribute")
	}

	htmlSampleIR := `<!DOCTYPE html><html xmlns="http://www.w3.org/1999/xhtml" lang="en-IR"><head><title>Search settings</title></head><body>Preferences</body></html>`
	if m := langRegex.FindStringSubmatch(htmlSampleIR); len(m) > 1 {
		country := m[1]
		if country != "IR" {
			t.Errorf("expected country IR, got %s", country)
		}
		if _, isBanned := GoogleBannedCountries[country]; !isBanned {
			t.Errorf("expected IR to be in GoogleBannedCountries")
		}
	} else {
		t.Errorf("failed to match country from lang attribute")
	}
}
