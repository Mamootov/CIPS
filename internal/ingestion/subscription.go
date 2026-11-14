package ingestion

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"cips/internal/domain"
)

// FetchSubscription retrieves a remote subscription URL, decodes base64 if needed, and returns extracted configs.
func FetchSubscription(ctx context.Context, subURL string) ([]domain.ProxyConfig, int, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", subURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid subscription url: %w", err)
	}

	req.Header.Set("User-Agent", "v2rayN/6.23 CIPS-Fetcher/1.0")
	req.Header.Set("Accept", "*/*")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch subscription: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("subscription server returned HTTP %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read subscription response: %w", err)
	}

	rawText := string(bodyBytes)

	// If content is base64 encoded, decode it
	trimmed := strings.TrimSpace(rawText)
	if !strings.Contains(trimmed, "://") {
		if decoded, err := decodeBase64Flexible(trimmed); err == nil {
			rawText = string(decoded)
		}
	}

	configs, total := ExtractAndDeduplicate(rawText)
	return configs, total, nil
}
