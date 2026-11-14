package probe

import (
	"fmt"
	"strings"
	"time"

	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

// BuildTLSClient creates an authentic Chrome 124 client routed through proxyEndpoint (e.g. 127.0.0.1:20005).
func BuildTLSClient(proxyEndpoint string, timeout time.Duration) (tls_client.HttpClient, error) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithTimeoutSeconds(int(timeout.Seconds())),
		tls_client.WithInsecureSkipVerify(),
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

// ApplyChromeHeaders injects realistic Chrome browser navigation headers.
func ApplyChromeHeaders(req interface{ HeaderSet(key, value string) }) {
	req.HeaderSet("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.HeaderSet("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.HeaderSet("Accept-Language", "en-US,en;q=0.9")
	req.HeaderSet("Sec-Ch-Ua", `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`)
	req.HeaderSet("Sec-Ch-Ua-Mobile", "?0")
	req.HeaderSet("Sec-Ch-Ua-Platform", `"Windows"`)
	req.HeaderSet("Sec-Fetch-Dest", "document")
	req.HeaderSet("Sec-Fetch-Mode", "navigate")
	req.HeaderSet("Sec-Fetch-Site", "none")
	req.HeaderSet("Sec-Fetch-User", "?1")
	req.HeaderSet("Upgrade-Insecure-Requests", "1")
}
