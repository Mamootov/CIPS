package probe

import (
	"fmt"
	"regexp"
	"strings"

	http "github.com/bogdanfinn/fhttp"

	"cips/internal/domain"
)

var (
	scriptTagRegex = regexp.MustCompile(`(?is)<script.*?>.*?</script>`)
	titleTagRegex  = regexp.MustCompile(`(?i)<title>(.*?)</title>`)
)

// InspectResponse performs deep multi-layer analysis of response status, headers, and payload.
func InspectResponse(resp *http.Response, body []byte, reqErr error, targetURL string) domain.DiagnosticInfo {
	// 1. Inspect Transport / Socket Errors
	if reqErr != nil {
		errStr := reqErr.Error()

		if strings.Contains(errStr, "GOOGLE_REGION_BLOCKED_INTERCEPT:") {
			urlPart := strings.TrimPrefix(errStr, "GOOGLE_REGION_BLOCKED_INTERCEPT:")
			return domain.DiagnosticInfo{
				Layer:            domain.LayerTargetService,
				InterceptionType: domain.InterceptGeoRestricted,
				FinalURL:         urlPart,
				Summary:          "Redirected to available-regions documentation (Region not supported)",
			}
		}

		if strings.Contains(errStr, "GOOGLE_UNUSUAL_TRAFFIC_INTERCEPT:") {
			urlPart := strings.TrimPrefix(errStr, "GOOGLE_UNUSUAL_TRAFFIC_INTERCEPT:")
			return domain.DiagnosticInfo{
				Layer:            domain.LayerTargetService,
				InterceptionType: domain.InterceptGoogleUnusualTraffic,
				FinalURL:         urlPart,
				Summary:          "Google unusual traffic challenge (/sorry/index captcha)",
			}
		}

		if strings.Contains(errStr, "connectex") || strings.Contains(errStr, "actively refused it") || strings.Contains(errStr, "connection refused") {
			return domain.DiagnosticInfo{
				Layer:            domain.LayerLocalEngine,
				InterceptionType: domain.InterceptLocalEngineError,
				FinalURL:         targetURL,
				Summary:          "Local SOCKS5 port refused connection (Xray engine not ready or port conflict)",
			}
		}

		if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") || strings.Contains(errStr, "i/o timeout") {
			return domain.DiagnosticInfo{
				Layer:            domain.LayerUpstreamProxy,
				InterceptionType: domain.InterceptUpstreamTimeout,
				FinalURL:         targetURL,
				Summary:          "Remote proxy node timed out during request transmission",
			}
		}

		return domain.DiagnosticInfo{
			Layer:            domain.LayerUpstreamProxy,
			InterceptionType: domain.InterceptUnknown,
			FinalURL:         targetURL,
			Summary:          errStr,
		}
	}

	if resp == nil {
		return domain.DiagnosticInfo{
			Layer:            domain.LayerUpstreamProxy,
			InterceptionType: domain.InterceptUnknown,
			Summary:          "Empty response received",
		}
	}

	serverHeader := resp.Header.Get("Server")
	rayID := resp.Header.Get("Cf-Ray")
	location := resp.Header.Get("Location")
	pageTitle := ExtractPageTitle(body)

	// 2. Check Google Unusual Traffic Redirect (/sorry/index) or Region Block (/docs/available-regions)
	if strings.Contains(location, "available-regions") || strings.Contains(location, "/docs/available-regions") {
		return domain.DiagnosticInfo{
			Layer:            domain.LayerTargetService,
			InterceptionType: domain.InterceptGeoRestricted,
			FinalURL:         location,
			HTTPStatus:       resp.StatusCode,
			ServerHeader:     serverHeader,
			PageTitle:        pageTitle,
			Summary:          "Redirected to available-regions documentation (Region not supported)",
		}
	}

	if strings.Contains(location, "google.com/sorry") || strings.Contains(location, "/sorry/index") {
		return domain.DiagnosticInfo{
			Layer:            domain.LayerTargetService,
			InterceptionType: domain.InterceptGoogleUnusualTraffic,
			FinalURL:         location,
			HTTPStatus:       resp.StatusCode,
			ServerHeader:     serverHeader,
			PageTitle:        pageTitle,
			Summary:          "Google unusual traffic challenge (/sorry/index captcha)",
		}
	}

	// 3. Check Cloudflare Challenge (Under-Attack / Turnstile / 403 / 503)
	isCloudflareServer := strings.EqualFold(serverHeader, "cloudflare")
	hasRayID := rayID != ""
	hasCFTitle := strings.Contains(strings.ToLower(pageTitle), "just a moment") || strings.Contains(strings.ToLower(pageTitle), "attention required")

	if (isCloudflareServer && (resp.StatusCode == 403 || resp.StatusCode == 503 || hasRayID)) || hasCFTitle {
		summary := "Cloudflare security challenge / Turnstile"
		if rayID != "" {
			summary = fmt.Sprintf("Cloudflare challenge (Ray-ID: %s)", rayID)
		}
		return domain.DiagnosticInfo{
			Layer:            domain.LayerTargetService,
			InterceptionType: domain.InterceptCloudflareChallenge,
			FinalURL:         targetURL,
			HTTPStatus:       resp.StatusCode,
			ServerHeader:     serverHeader,
			RayID:            rayID,
			PageTitle:        pageTitle,
			Summary:          summary,
		}
	}

	// 4. Check Regional Restriction / Geoblock (Filter scripts to prevent OAuth false positives)
	bodyNoScripts := StripScriptTags(string(body))
	bodyLower := strings.ToLower(bodyNoScripts)

	geoblockTokens := []string{
		"isn't supported in your country",
		"isn't currently supported in your country",
		"location_error",
		"unsupported country",
		"unsupported_country",
		"user location not supported",
		"not available in your country",
		"not available in your region",
		"your region may not be supported",
		"region may not be supported",
		"available-regions",
		"openai's services are not available",
		"claude is not available in your region",
	}

	for _, token := range geoblockTokens {
		if strings.Contains(bodyLower, token) {
			return domain.DiagnosticInfo{
				Layer:            domain.LayerTargetService,
				InterceptionType: domain.InterceptGeoRestricted,
				FinalURL:         targetURL,
				HTTPStatus:       resp.StatusCode,
				ServerHeader:     serverHeader,
				PageTitle:        pageTitle,
				Summary:          "Service restricted in current proxy region/country",
			}
		}
	}

	// 5. Check Unexpected Status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return domain.DiagnosticInfo{
			Layer:            domain.LayerTargetService,
			InterceptionType: domain.InterceptUnknown,
			FinalURL:         targetURL,
			HTTPStatus:       resp.StatusCode,
			ServerHeader:     serverHeader,
			PageTitle:        pageTitle,
			Summary:          fmt.Sprintf("Unexpected HTTP %d from destination", resp.StatusCode),
		}
	}

	// 6. Clean Accessible Destination (Pass)
	return domain.DiagnosticInfo{
		Layer:            domain.LayerNone,
		InterceptionType: domain.InterceptNone,
		FinalURL:         targetURL,
		HTTPStatus:       resp.StatusCode,
		ServerHeader:     serverHeader,
		PageTitle:        pageTitle,
		Summary:          "Service accessible (Pass)",
	}
}

// StripScriptTags removes <script>...</script> blocks from HTML.
func StripScriptTags(html string) string {
	return scriptTagRegex.ReplaceAllString(html, "")
}

// ExtractPageTitle parses <title>...</title> content from HTML.
func ExtractPageTitle(body []byte) string {
	matches := titleTagRegex.FindSubmatch(body)
	if len(matches) > 1 {
		return strings.TrimSpace(string(matches[1]))
	}
	return ""
}
