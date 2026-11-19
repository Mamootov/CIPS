# Implementation Plan: 004 - IP Intelligence and Low-Bandwidth Chunked Speedtest

**Feature Branch**: `004-ip-intelligence-and`  
**Status**: Ready for Implementation  

---

## 1. Technical Architecture & Data Flow

```
┌────────────────────────────────────────────────────────┐
│                        web/app.js                      │
│  - Toggle "Enable Speed Test" & Size Selector          │
│  - Speed Column with Color Tiers (🚀, ⚡, 🐌)          │
│  - IP/ISP Column with 🏠 Res / 🏢 DC Badges            │
└───────────────────────────┬────────────────────────────┘
                            │ REST / SSE
┌───────────────────────────▼────────────────────────────┐
│                  internal/server/handlers.go           │
│  - HandleScanStart parses enable_speedtest & bytes     │
│  - Broadcasts live speed & enriched IP data via SSE    │
└───────────────────────────┬────────────────────────────┘
                            │
┌────────────────────────────────────────────────────────┐
│               internal/pipeline/orchestrator.go        │
│  - Stage 2A: Service Probes (Gemini, AI Studio, IPIntel│
│  - Gate: Did at least one AI service pass?             │
│  - Stage 2B: Run Speedtest ONLY if hasAIPass == true   │
└──────────────┬───────────────────────────┬─────────────┘
               │                           │
┌──────────────▼────────────┐ ┌────────────▼─────────────┐
│ internal/probe/ip_intel   │ │ internal/probe/speedtest │
│ - Tier 1: ipwho.is (HTTPS)│ │ - speed.cloudflare.com   │
│ - Tier 2: ipleak.net JSON │ │ - 3s hard timeout        │
│ - Tier 3: CF cdn trace    │ │ - Accurate Mbps calc     │
└───────────────────────────┘ └──────────────────────────┘
```

---

## 2. Component Design & Changes

### 2.1 Domain Models (`internal/domain/models.go`)
- Update `IPEnrichment`:
  - `ConnectionType string` (`"Residential"`, `"Data Center"`, `"Business"`, `"Unknown"`)
  - `ASN string`
  - `Org string`
- Update `ProxyEvaluationReport`:
  - `DownloadSpeedMbps float64`

### 2.2 IP Intelligence Probe (`internal/probe/ip_intel.go`)
- Query `https://ipwho.is/` through proxy tunnel.
- Parse JSON fields:
  - `ip`, `country`, `country_code`, `city`, `connection.isp`, `connection.asn`, `connection.org`, `connection.type`.
- If `ipwho.is` fails, falls back to `https://ipleak.net/json/`.
- If `ipleak.net` fails, falls back to `https://speed.cloudflare.com/cdn-cgi/trace`.

### 2.3 Low-Bandwidth Speedtest Probe (`internal/probe/speedtest.go`)
- Probe ID: `"speedtest"`.
- Requests `https://speed.cloudflare.com/__down?bytes={size}` (default 3,000,000 bytes = 3MB).
- Context with hard 3000ms timeout:
  - Reads chunks with `io.Copy(io.Discard, io.LimitReader(resp.Body, maxBytes))`.
  - Measures microsecond duration:
    $$\text{Mbps} = \frac{\text{BytesRead} \times 8}{\text{ElapsedSeconds} \times 1,000,000}$$
  - Returns `ProbeResult` with `LatencyMs` and formatted `Summary = fmt.Sprintf("%.2f Mbps", mbps)`.

### 2.4 Subscription Exporter (`internal/exporter/subscription.go`)
- Format: `[CC] [ISP] [Type] [Speed] [Ping] [Probes] - Remark`
- Example: `[US] [Hetzner] [DC] [18.4Mbps] [145ms] [Gemini+Studio] - Node01`

### 2.5 Web Interface (`web/index.html`, `web/style.css`, `web/app.js`)
- Sidebar Setting:
  - Checkbox toggle: `Enable Speedtest` (default ON)
  - Slider/Select: `Speed Payload Size` (1MB, 3MB, 5MB - default 3MB)
- Matrix Grid:
  - Insert `Speed` column between `Ping` and `Gemini`.
  - Badges:
    - `>= 10 Mbps`: `badge badge-pass` (`🚀 24.5 Mbps`)
    - `2 - 10 Mbps`: `badge badge-running` (`⚡ 5.2 Mbps`)
    - `< 2 Mbps`: `badge badge-warning` (`🐌 0.8 Mbps`)
    - `0 / Dead`: `badge badge-fail` (`✗ 0 Mbps`)
    - Pending: `badge badge-pending` (`○`)
  - `IP / ISP` column:
    - Flag icon / Country code + ISP + Type pill (`🏠 Res` or `🏢 DC`).
