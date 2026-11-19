# Data Model: 004 - IP Intelligence & Low-Bandwidth Speedtest

---

## 1. Domain Entities

### 1.1 `IPEnrichment` Model Extension
```go
type IPEnrichment struct {
    Query          string `json:"query"`
    Country        string `json:"country"`
    CountryCode    string `json:"country_code"`
    Region         string `json:"region"`
    City           string `json:"city"`
    ISP            string `json:"isp"`
    ConnectionType string `json:"connection_type"` // "Residential", "Data Center", "Business", "Wireless", "Unknown"
    ASN            string `json:"asn"`             // e.g. "AS24940"
    Org            string `json:"org"`             // e.g. "Hetzner Online GmbH"
}
```

### 1.2 `ProxyEvaluationReport` Extension
```go
type ProxyEvaluationReport struct {
    // ... existing fields ...
    DownloadSpeedMbps float64 `json:"download_speed_mbps"` // Speed in Megabits per second
}
```

---

## 2. API Contract Updates

### 2.1 `POST /api/scan/start`
Request Payload:
```json
{
  "concurrency": 20,
  "stage1_timeout_ms": 3000,
  "stage2_timeout_ms": 8000,
  "enable_speedtest": true,
  "speedtest_bytes": 3000000
}
```

### 2.2 `GET /api/session` & SSE Stream Item Event
Each serialized `ProxyEvaluationReport` will now include:
```json
{
  "download_speed_mbps": 18.42,
  "ip_info": {
    "query": "185.120.45.10",
    "country": "Germany",
    "country_code": "DE",
    "isp": "Hetzner Online GmbH",
    "connection_type": "Data Center",
    "asn": "AS24940",
    "org": "Hetzner Online GmbH"
  }
}
```
