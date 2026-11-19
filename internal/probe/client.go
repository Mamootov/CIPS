package probe

import (
	"fmt"
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

// BuildTLSClient creates an authentic Chrome 124 client routed through proxyEndpoint (e.g. 127.0.0.1:20005).
func BuildTLSClient(proxyEndpoint string, timeout time.Duration) (tls_client.HttpClient, error) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	redirectCheck := func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("stopped after 5 redirects")
		}
		target := req.URL.String()
		if strings.Contains(target, "google.com/sorry") || strings.Contains(target, "/sorry/index") {
			return fmt.Errorf("GOOGLE_UNUSUAL_TRAFFIC_INTERCEPT:%s", target)
		}
		if strings.Contains(target, "available-regions") || strings.Contains(target, "/docs/available-regions") {
			return fmt.Errorf("GOOGLE_REGION_BLOCKED_INTERCEPT:%s", target)
		}
		return nil
	}

	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithTimeoutSeconds(int(timeout.Seconds())),
		tls_client.WithInsecureSkipVerify(),
		tls_client.WithCustomRedirectFunc(redirectCheck),
	}

	if proxyEndpoint != "" {
		if !strings.HasPrefix(proxyEndpoint, "socks5://") && !strings.HasPrefix(proxyEndpoint, "http://") {
			proxyEndpoint = "socks5://" + proxyEndpoint
		}
		options = append(options, tls_client.WithProxyUrl(proxyEndpoint))
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create tls client: %w", err)
	}

	return client, nil
}

// ApplyChromeHeaders injects realistic Chrome browser navigation headers without anomalous Sec-Fetch triggers.
func ApplyChromeHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
}
