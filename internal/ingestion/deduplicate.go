package ingestion

import (
	"bufio"
	"regexp"
	"strings"

	"cips/internal/domain"
)

var uriPattern = regexp.MustCompile(`(?i)(vless|vmess|trojan|ss|socks5|socks|http|https)://[^\s"'<>]+`)

// ExtractAndDeduplicate parses all proxy URIs in the input string and removes duplicate nodes.
func ExtractAndDeduplicate(content string) ([]domain.ProxyConfig, int) {
	seen := make(map[string]bool)
	var configs []domain.ProxyConfig
	totalFound := 0

	scanner := bufio.NewScanner(strings.NewReader(content))
	// Support long lines (e.g. large base64 vmess/sub strings up to 1MB per line)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}

		matches := uriPattern.FindAllString(line, -1)
		for _, rawURI := range matches {
			totalFound++
			cfg, err := ParseURI(rawURI)
			if err != nil {
				continue
			}

			if !seen[cfg.ID] {
				seen[cfg.ID] = true
				configs = append(configs, cfg)
			}
		}
	}

	return configs, totalFound
}
