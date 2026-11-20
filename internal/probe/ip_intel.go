package probe

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"

	"cips/internal/domain"
)

// IPIntelProbe retrieves exit IP, ISP, ASN, and connection type (Residential vs Data Center).
type IPIntelProbe struct {
	timeout time.Duration
}

// NewIPIntelProbe creates an IP intelligence metadata enricher.
func NewIPIntelProbe(timeoutMs int) *IPIntelProbe {
	if timeoutMs <= 0 {
		timeoutMs = 8000
	}
	return &IPIntelProbe{
		timeout: time.Duration(timeoutMs) * time.Millisecond,
	}
}

// NewIPLeakProbe is a compatibility alias for NewIPIntelProbe.
func NewIPLeakProbe(timeoutMs int) *IPIntelProbe {
	return NewIPIntelProbe(timeoutMs)
}

func (p *IPIntelProbe) ID() string {
	return "ipleak"
}

func (p *IPIntelProbe) Name() string {
	return "IP Intelligence"
}

func (p *IPIntelProbe) Evaluate(ctx context.Context, proxyEndpoint string) domain.ProbeResult {
	start := time.Now()

	// Add slight random jitter (50ms - 150ms) to prevent burst collisions
	jitter := time.Duration(50+rand.Intn(100)) * time.Millisecond
	select {
	case <-time.After(jitter):
	case <-ctx.Done():
		return domain.ProbeResult{
			ProbeID:      p.ID(),
			ProbeName:    p.Name(),
			Status:       domain.StatusFail,
			ErrorMessage: "Context canceled",
		}
	}

	ipInfo, err := p.ExtractIPInfo(ctx, proxyEndpoint)
	latency := time.Since(start).Milliseconds()

	if err != nil || !ipInfo.Success {
		errMsg := "Failed to retrieve IP metadata"
		if err != nil {
			errMsg = err.Error()
		}
		return domain.ProbeResult{
			ProbeID:      p.ID(),
			ProbeName:    p.Name(),
			Status:       domain.StatusFail,
			LatencyMs:    latency,
			ErrorMessage: errMsg,
		}
	}

	return domain.ProbeResult{
		ProbeID:   p.ID(),
		ProbeName: p.Name(),
		Status:    domain.StatusPass,
		HTTPCode:  200,
		LatencyMs: latency,
	}
}

// ExtractIPInfo queries multi-tier IP intelligence sources with automatic fallback.
func (p *IPIntelProbe) ExtractIPInfo(ctx context.Context, proxyEndpoint string) (domain.IPEnrichment, error) {
	client, err := BuildTLSClient(proxyEndpoint, p.timeout)
	if err != nil {
		return domain.IPEnrichment{Success: false}, err
	}

	start := time.Now()

	// Tier 1: ip-api.com (Includes hosting: true/false for Residential vs Datacenter)
	if info, err := queryIPAPI(ctx, client); err == nil && info.Success {
		info.QueryTimeMs = time.Since(start).Milliseconds()
		return info, nil
	}

	// Tier 2: ipwho.is (HTTPS fallback with connection.type)
	if info, err := queryIPWhoIs(ctx, client); err == nil && info.Success {
		info.QueryTimeMs = time.Since(start).Milliseconds()
		return info, nil
	}

	// Tier 3: ipleak.net JSON API
	if info, err := queryIPLeakNet(ctx, client); err == nil && info.Success {
		info.QueryTimeMs = time.Since(start).Milliseconds()
		return info, nil
	}

	// Tier 4: Cloudflare CDN trace fallback
	if info, err := queryCloudflareTrace(ctx, client); err == nil && info.Success {
		info.QueryTimeMs = time.Since(start).Milliseconds()
		return info, nil
	}

	return domain.IPEnrichment{Success: false}, fmt.Errorf("all IP intelligence providers failed")
}

func queryIPAPI(ctx context.Context, client tls_client.HttpClient) (domain.IPEnrichment, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", "http://ip-api.com/json/?fields=status,country,countryCode,city,isp,org,as,query,hosting", nil)
	if err != nil {
		return domain.IPEnrichment{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 Chrome/124.0.0.0")

	resp, err := client.Do(req)
	if err != nil {
		return domain.IPEnrichment{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.IPEnrichment{}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
	if err != nil {
		return domain.IPEnrichment{}, err
	}

	var data struct {
		Status      string  `json:"status"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		City        string  `json:"city"`
		ISP         string  `json:"isp"`
		Org         string  `json:"org"`
		AS          string  `json:"as"`
		Query       string  `json:"query"`
		Hosting     bool    `json:"hosting"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
	}

	if err := json.Unmarshal(body, &data); err != nil || data.Status != "success" {
		return domain.IPEnrichment{}, fmt.Errorf("invalid response from ip-api")
	}

	connType := "Residential"
	if data.Hosting {
		connType = "Data Center"
	}

	return domain.IPEnrichment{
		ExitIP:         data.Query,
		CountryCode:    data.CountryCode,
		CountryName:    data.Country,
		City:           data.City,
		ISP:            data.ISP,
		ASN:            data.AS,
		Org:            data.Org,
		ConnectionType: connType,
		Latitude:       data.Lat,
		Longitude:      data.Lon,
		Success:        true,
	}, nil
}

func queryIPWhoIs(ctx context.Context, client tls_client.HttpClient) (domain.IPEnrichment, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", "https://ipwho.is/", nil)
	if err != nil {
		return domain.IPEnrichment{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 Chrome/124.0.0.0")

	resp, err := client.Do(req)
	if err != nil {
		return domain.IPEnrichment{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.IPEnrichment{}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 32*1024))
	if err != nil {
		return domain.IPEnrichment{}, err
	}

	var data struct {
		Success     bool    `json:"success"`
		IP          string  `json:"ip"`
		Country     string  `json:"country"`
		CountryCode string  `json:"country_code"`
		City        string  `json:"city"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		Connection  struct {
			ASN  int    `json:"asn"`
			Org  string `json:"org"`
			ISP  string `json:"isp"`
			Type string `json:"type"`
		} `json:"connection"`
	}

	if err := json.Unmarshal(body, &data); err != nil || !data.Success {
		return domain.IPEnrichment{}, fmt.Errorf("invalid response from ipwho.is")
	}

	connType := data.Connection.Type
	if connType == "" {
		connType = "Unknown"
	}

	asnStr := ""
	if data.Connection.ASN > 0 {
		asnStr = fmt.Sprintf("AS%d", data.Connection.ASN)
	}

	return domain.IPEnrichment{
		ExitIP:         data.IP,
		CountryCode:    data.CountryCode,
		CountryName:    data.Country,
		City:           data.City,
		ISP:            data.Connection.ISP,
		ASN:            asnStr,
		Org:            data.Connection.Org,
		ConnectionType: connType,
		Latitude:       data.Latitude,
		Longitude:      data.Longitude,
		Success:        true,
	}, nil
}

// DetectHostOrigin discovers the real direct egress IP and location of the host running CIPS.
func DetectHostOrigin(ctx context.Context) domain.IPEnrichment {
	reqCtx, cancel := context.WithTimeout(ctx, 3500*time.Millisecond)
	defer cancel()

	// Tier 1: query ipwho.is directly
	req, err := http.NewRequestWithContext(reqCtx, "GET", "https://ipwho.is/", nil)
	if err == nil {
		req.Header.Set("User-Agent", "Mozilla/5.0 Chrome/124.0.0.0")
		client, cErr := BuildTLSClient("", 3500*time.Millisecond)
		if cErr == nil {
			resp, rErr := client.Do(req)
			if rErr == nil && resp.StatusCode == 200 {
				body, _ := io.ReadAll(io.LimitReader(resp.Body, 32*1024))
				_ = resp.Body.Close()
				var data struct {
					Success     bool    `json:"success"`
					IP          string  `json:"ip"`
					Country     string  `json:"country"`
					CountryCode string  `json:"country_code"`
					City        string  `json:"city"`
					Latitude    float64 `json:"latitude"`
					Longitude   float64 `json:"longitude"`
					Connection  struct {
						ISP string `json:"isp"`
						Org string `json:"org"`
					} `json:"connection"`
				}
				if json.Unmarshal(body, &data) == nil && data.Success && data.CountryCode != "" {
					return domain.IPEnrichment{
						ExitIP:      data.IP,
						CountryCode: data.CountryCode,
						CountryName: data.Country,
						City:        data.City,
						ISP:         data.Connection.ISP,
						Org:         data.Connection.Org,
						Latitude:    data.Latitude,
						Longitude:   data.Longitude,
						Success:     true,
					}
				}
			}
		}
	}

	// Safe fallback (Tehran Origin) if host has no raw egress or lookup failed
	return domain.IPEnrichment{
		ExitIP:      "127.0.0.1",
		CountryCode: "IR",
		CountryName: "Iran",
		City:        "Tehran",
		Latitude:    35.6892,
		Longitude:   51.3890,
		Success:     true,
	}
}

func queryIPLeakNet(ctx context.Context, client tls_client.HttpClient) (domain.IPEnrichment, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", "https://ipleak.net/json/", nil)
	if err != nil {
		return domain.IPEnrichment{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 Chrome/124.0.0.0")

	resp, err := client.Do(req)
	if err != nil {
		return domain.IPEnrichment{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.IPEnrichment{}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 32*1024))
	if err != nil {
		return domain.IPEnrichment{}, err
	}

	var data struct {
		IP          string `json:"ip"`
		CountryName string `json:"country_name"`
		CountryCode string `json:"country_code"`
		ISPName     string `json:"isp_name"`
		ASNumber    string `json:"as_number"`
	}

	if err := json.Unmarshal(body, &data); err != nil || data.IP == "" {
		return domain.IPEnrichment{}, fmt.Errorf("invalid response from ipleak.net")
	}

	return domain.IPEnrichment{
		ExitIP:         data.IP,
		CountryCode:    data.CountryCode,
		CountryName:    data.CountryName,
		ISP:            data.ISPName,
		ASN:            data.ASNumber,
		ConnectionType: "Unknown",
		Success:        true,
	}, nil
}

func queryCloudflareTrace(ctx context.Context, client tls_client.HttpClient) (domain.IPEnrichment, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", "https://speed.cloudflare.com/cdn-cgi/trace", nil)
	if err != nil {
		return domain.IPEnrichment{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return domain.IPEnrichment{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.IPEnrichment{}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
	if err != nil {
		return domain.IPEnrichment{}, err
	}

	ip := ""
	loc := ""
	scanner := bufio.NewScanner(bytes.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ip=") {
			ip = strings.TrimPrefix(line, "ip=")
		} else if strings.HasPrefix(line, "loc=") {
			loc = strings.TrimPrefix(line, "loc=")
		}
	}

	if ip == "" {
		return domain.IPEnrichment{}, fmt.Errorf("ip not found in cloudflare trace")
	}

	return domain.IPEnrichment{
		ExitIP:         ip,
		CountryCode:    loc,
		CountryName:    loc,
		ISP:            "Cloudflare Edge",
		ConnectionType: "Unknown",
		Success:        true,
	}, nil
}
