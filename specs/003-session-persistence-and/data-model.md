# Data Model & REST Contracts: 003 - Session Persistence & Accuracy

---

## 1. REST Endpoints

### 1.1 `GET /api/session`
Returns current scan state and all reports held in server memory. Used to hydrate the frontend on page load/refresh.

**Response** (`application/json`):
```json
{
  "reports": [
    {
      "config": {
        "id": "vless-8e31a2bc",
        "protocol": "vless",
        "server": "194.87.112.5",
        "port": 443,
        "uuid": "a1b2c3d4-...",
        "remark": "Node-01"
      },
      "stage1_passed": true,
      "stage1_latency_ms": 142,
      "state": "COMPLETE",
      "is_clean": true,
      "probe_results": {
        "gemini": {
          "probe_id": "gemini",
          "probe_name": "Google Gemini",
          "status": "PASS",
          "http_code": 200,
          "latency_ms": 312
        },
        "aistudio": {
          "probe_id": "aistudio",
          "probe_name": "Google AI Studio",
          "status": "PASS",
          "http_code": 200,
          "latency_ms": 280
        }
      },
      "ip_info": {
        "query": "194.87.112.5",
        "country": "Germany",
        "country_code": "DE",
        "isp": "Hetzner Online GmbH"
      }
    }
  ],
  "total": 1,
  "is_running": false
}
```

---

### 1.2 `POST /api/reset`
Terminates any active pipeline execution, unbinds ports, and purges all in-memory candidates and reports.

**Response** (`application/json`):
```json
{
  "status": "cleared"
}
```

---

## 2. In-Memory Cache Model

### 2.1 `geoCache` (Thread-Safe Memory Map)
```go
var (
    geoCache   = make(map[string]string) // key: ProxyConfig.ID, value: 2-letter ISO country code (e.g. "DE", "IR")
    geoCacheMu sync.RWMutex
)
```
- **Key**: `ProxyConfig.ID` (immutable unique hash of protocol, server, port, uuid).
- **Value**: Two-letter uppercase ISO country code extracted from Google Anycast.
- **Scope**: Reused by concurrent Gemini and AI Studio probes evaluating the same proxy node; distinct across different proxy candidates.

---

## 3. Policy Reference Model

### 3.1 `GoogleBannedCountries`
```go
var GoogleBannedCountries = map[string]string{
    "IR": "Iran",
    "RU": "Russia",
    "BY": "Belarus",
    "CN": "China",
    "KP": "North Korea",
    "SY": "Syria",
    "CU": "Cuba",
    "AF": "Afghanistan",
}
```
Direct mapping of Google's officially published unserviced regions for consumer Gemini and Google AI Studio APIs.
