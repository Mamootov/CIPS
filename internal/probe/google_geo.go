package probe

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"sync"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

var (
	langRegex = regexp.MustCompile(`lang="en-([A-Z]{2})"`)
	hlRegex   = regexp.MustCompile(`hl=en-([A-Z]{2})`)

	// GoogleBannedCountries lists countries where Google officially denies Gemini & AI Studio access.
	// Source: support.google.com/gemini/answer/13575153 & ai.google.dev/gemini-api/docs/available-regions
	GoogleBannedCountries = map[string]string{
		"IR": "Iran",
		"RU": "Russia",
		"BY": "Belarus",
		"CN": "China",
		"KP": "North Korea",
		"SY": "Syria",
		"CU": "Cuba",
		"AF": "Afghanistan",
	}

	geoCache   = make(map[string]string)
	geoCacheMu sync.RWMutex
)

// DetectGoogleRegion queries Google directly through the proxy tunnel to extract the 2-letter country code Google assigned to this IP.
func DetectGoogleRegion(ctx context.Context, client tls_client.HttpClient, proxyEndpoint string) (string, error) {
	geoCacheMu.RLock()
	if cached, ok := geoCache[proxyEndpoint]; ok {
		geoCacheMu.RUnlock()
		return cached, nil
	}
	geoCacheMu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.google.com/preferences?hl=en", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return "", err
	}

	bodyStr := string(bodyBytes)
	if m := langRegex.FindStringSubmatch(bodyStr); len(m) > 1 {
		country := m[1]
		geoCacheMu.Lock()
		geoCache[proxyEndpoint] = country
		geoCacheMu.Unlock()
		return country, nil
	}

	if m := hlRegex.FindStringSubmatch(bodyStr); len(m) > 1 {
		country := m[1]
		geoCacheMu.Lock()
		geoCache[proxyEndpoint] = country
		geoCacheMu.Unlock()
		return country, nil
	}

	return "", fmt.Errorf("unable to detect google assigned region")
}
