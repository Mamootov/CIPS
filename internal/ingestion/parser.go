package ingestion

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"cips/internal/domain"
)

// ParseURI parses a single proxy URI into a standardized domain.ProxyConfig.
func ParseURI(raw string) (domain.ProxyConfig, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return domain.ProxyConfig{}, fmt.Errorf("empty proxy URI")
	}

	lower := strings.ToLower(raw)
	switch {
	case strings.HasPrefix(lower, "vless://"):
		return parseVLESS(raw)
	case strings.HasPrefix(lower, "vmess://"):
		return parseVMess(raw)
	case strings.HasPrefix(lower, "trojan://"):
		return parseTrojan(raw)
	case strings.HasPrefix(lower, "ss://"):
		return parseShadowsocks(raw)
	case strings.HasPrefix(lower, "socks5://") || strings.HasPrefix(lower, "socks://"):
		return parseSocks5(raw)
	case strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://"):
		return parseHTTP(raw)
	default:
		return domain.ProxyConfig{}, fmt.Errorf("unsupported protocol scheme in: %s", raw)
	}
}

func parseVLESS(raw string) (domain.ProxyConfig, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return domain.ProxyConfig{}, fmt.Errorf("failed to parse vless url: %w", err)
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return domain.ProxyConfig{}, fmt.Errorf("invalid port in vless url: %w", err)
	}

	q := u.Query()
	extra := make(map[string]string)
	for k, v := range q {
		if len(v) > 0 {
			extra[k] = v[0]
		}
	}

	cfg := domain.ProxyConfig{
		RawURI:      raw,
		Protocol:    domain.ProtocolVLESS,
		Server:      u.Hostname(),
		Port:        port,
		UUID:        u.User.Username(),
		Security:    q.Get("security"),
		SNI:         q.Get("sni"),
		Host:        q.Get("host"),
		Path:        q.Get("path"),
		Type:        q.Get("type"),
		Remark:      u.Fragment,
		ExtraParams: extra,
	}

	cfg.ID = cfg.ComputeID()
	return cfg, cfg.Validate()
}

func parseVMess(raw string) (domain.ProxyConfig, error) {
	b64Part := strings.TrimPrefix(raw, "vmess://")
	b64Part = strings.TrimPrefix(b64Part, "VMESS://")

	decoded, err := decodeBase64Flexible(b64Part)
	if err != nil {
		return domain.ProxyConfig{}, fmt.Errorf("failed to decode vmess base64: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(decoded, &data); err != nil {
		return domain.ProxyConfig{}, fmt.Errorf("failed to parse vmess json: %w", err)
	}

	server, _ := data["add"].(string)
	portStr := fmt.Sprintf("%v", data["port"])
	port, _ := strconv.Atoi(portStr)
	id, _ := data["id"].(string)
	ps, _ := data["ps"].(string)
	netType, _ := data["net"].(string)
	tlsStr, _ := data["tls"].(string)
	sni, _ := data["sni"].(string)
	host, _ := data["host"].(string)
	path, _ := data["path"].(string)

	cfg := domain.ProxyConfig{
		RawURI:      raw,
		Protocol:    domain.ProtocolVMess,
		Server:      server,
		Port:        port,
		UUID:        id,
		Security:    tlsStr,
		SNI:         sni,
		Host:        host,
		Path:        path,
		Type:        netType,
		Remark:      ps,
		ExtraParams: make(map[string]string),
	}

	cfg.ID = cfg.ComputeID()
	return cfg, cfg.Validate()
}

func parseTrojan(raw string) (domain.ProxyConfig, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return domain.ProxyConfig{}, fmt.Errorf("failed to parse trojan url: %w", err)
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return domain.ProxyConfig{}, fmt.Errorf("invalid port in trojan url: %w", err)
	}

	q := u.Query()
	extra := make(map[string]string)
	for k, v := range q {
		if len(v) > 0 {
			extra[k] = v[0]
		}
	}

	cfg := domain.ProxyConfig{
		RawURI:      raw,
		Protocol:    domain.ProtocolTrojan,
		Server:      u.Hostname(),
		Port:        port,
		UUID:        u.User.Username(),
		Security:    q.Get("security"),
		SNI:         q.Get("sni"),
		Host:        q.Get("host"),
		Path:        q.Get("path"),
		Type:        q.Get("type"),
		Remark:      u.Fragment,
		ExtraParams: extra,
	}

	cfg.ID = cfg.ComputeID()
	return cfg, cfg.Validate()
}

func parseShadowsocks(raw string) (domain.ProxyConfig, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return domain.ProxyConfig{}, fmt.Errorf("failed to parse ss url: %w", err)
	}

	var method, password, server string
	var port int

	if u.User != nil {
		// Format: ss://method:password@host:port or ss://base64@host:port
		userPart := u.User.Username()
		if pass, hasPass := u.User.Password(); hasPass {
			method = userPart
			password = pass
		} else {
			// userPart might be base64-encoded method:password
			if dec, err := decodeBase64Flexible(userPart); err == nil {
				parts := strings.SplitN(string(dec), ":", 2)
				if len(parts) == 2 {
					method = parts[0]
					password = parts[1]
				}
			}
		}
		server = u.Hostname()
		port, _ = strconv.Atoi(u.Port())
	} else {
		// Format: ss://base64(method:password@host:port)
		b64Part := strings.TrimPrefix(raw, "ss://")
		if idx := strings.Index(b64Part, "#"); idx != -1 {
			b64Part = b64Part[:idx]
		}
		if dec, err := decodeBase64Flexible(b64Part); err == nil {
			decStr := string(dec)
			subURL, err := url.Parse("ss://" + decStr)
			if err == nil {
				return parseShadowsocks(subURL.String() + "#" + u.Fragment)
			}
		}
	}

	extra := make(map[string]string)
	if method != "" {
		extra["method"] = method
	}

	cfg := domain.ProxyConfig{
		RawURI:      raw,
		Protocol:    domain.ProtocolShadowsocks,
		Server:      server,
		Port:        port,
		UUID:        password,
		Remark:      u.Fragment,
		ExtraParams: extra,
	}

	cfg.ID = cfg.ComputeID()
	return cfg, cfg.Validate()
}

func parseSocks5(raw string) (domain.ProxyConfig, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return domain.ProxyConfig{}, fmt.Errorf("failed to parse socks url: %w", err)
	}

	port, _ := strconv.Atoi(u.Port())
	cfg := domain.ProxyConfig{
		RawURI:   raw,
		Protocol: domain.ProtocolSocks5,
		Server:   u.Hostname(),
		Port:     port,
		Remark:   u.Fragment,
	}
	cfg.ID = cfg.ComputeID()
	return cfg, cfg.Validate()
}

func parseHTTP(raw string) (domain.ProxyConfig, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return domain.ProxyConfig{}, fmt.Errorf("failed to parse http url: %w", err)
	}

	port, _ := strconv.Atoi(u.Port())
	if port == 0 {
		if strings.HasPrefix(raw, "https://") {
			port = 443
		} else {
			port = 80
		}
	}

	cfg := domain.ProxyConfig{
		RawURI:   raw,
		Protocol: domain.ProtocolHTTP,
		Server:   u.Hostname(),
		Port:     port,
		Remark:   u.Fragment,
	}
	cfg.ID = cfg.ComputeID()
	return cfg, cfg.Validate()
}

func decodeBase64Flexible(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	// Add padding if missing
	if pad := len(s) % 4; pad > 0 {
		s += strings.Repeat("=", 4-pad)
	}

	// Try standard base64 first
	if b, err := base64.StdEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	// Try URL-safe base64
	return base64.URLEncoding.DecodeString(s)
}
