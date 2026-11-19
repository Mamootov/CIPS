package probe

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseIPAPIDataCenter(t *testing.T) {
	sample := `{"status":"success","country":"Germany","countryCode":"DE","city":"Frankfurt","isp":"Hetzner Online GmbH","org":"Hetzner Online GmbH","as":"AS24940 Hetzner Online GmbH","query":"185.120.45.10","hosting":true}`

	var data struct {
		Status      string `json:"status"`
		Country     string `json:"country"`
		CountryCode string `json:"countryCode"`
		City        string `json:"city"`
		ISP         string `json:"isp"`
		Org         string `json:"org"`
		AS          string `json:"as"`
		Query       string `json:"query"`
		Hosting     bool   `json:"hosting"`
	}

	if err := json.Unmarshal([]byte(sample), &data); err != nil {
		t.Fatalf("failed to unmarshal sample: %v", err)
	}

	connType := "Residential"
	if data.Hosting {
		connType = "Data Center"
	}

	if connType != "Data Center" {
		t.Errorf("expected Data Center, got %s", connType)
	}
	if data.CountryCode != "DE" {
		t.Errorf("expected DE, got %s", data.CountryCode)
	}
}

func TestParseIPWhoIsResidential(t *testing.T) {
	sample := `{"success":true,"ip":"95.82.12.4","country":"United States","country_code":"US","city":"Dallas","connection":{"asn":7018,"org":"AT&T Services","isp":"AT&T Internet","type":"Residential"}}`

	var data struct {
		Success     bool   `json:"success"`
		IP          string `json:"ip"`
		Country     string `json:"country"`
		CountryCode string `json:"country_code"`
		City        string `json:"city"`
		Connection  struct {
			ASN  int    `json:"asn"`
			Org  string `json:"org"`
			ISP  string `json:"isp"`
			Type string `json:"type"`
		} `json:"connection"`
	}

	if err := json.Unmarshal([]byte(sample), &data); err != nil {
		t.Fatalf("failed to unmarshal sample: %v", err)
	}

	if data.Connection.Type != "Residential" {
		t.Errorf("expected Residential, got %s", data.Connection.Type)
	}
}

func TestParseCloudflareTrace(t *testing.T) {
	traceSample := "fl=52f10\nh=speed.cloudflare.com\nip=185.209.196.174\nts=1700000000\nloc=DE\n"
	ip := ""
	loc := ""
	for _, line := range strings.Split(traceSample, "\n") {
		if strings.HasPrefix(line, "ip=") {
			ip = strings.TrimPrefix(line, "ip=")
		} else if strings.HasPrefix(line, "loc=") {
			loc = strings.TrimPrefix(line, "loc=")
		}
	}

	if ip != "185.209.196.174" {
		t.Errorf("expected 185.209.196.174, got %s", ip)
	}
	if loc != "DE" {
		t.Errorf("expected DE, got %s", loc)
	}
}
