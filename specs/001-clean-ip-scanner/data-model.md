# Data Model & Domain Entities: Clean IP Scanner (CIPS)

## Overview
This document specifies the core domain entities, data transfer objects, lifecycle states, and validation rules for the CIPS application in Go.

---

## 1. Domain Entities & Value Objects

### 1.1 ProxyConfig (Domain Entity)
Represents a parsed, normalized proxy endpoint ready for execution.

```go
type Protocol string

const (
    ProtocolVLESS       Protocol = "vless"
    ProtocolVMess       Protocol = "vmess"
    ProtocolTrojan      Protocol = "trojan"
    ProtocolShadowsocks Protocol = "shadowsocks"
    ProtocolSocks5      Protocol = "socks5"
    ProtocolHTTP        Protocol = "http"
)

type ProxyConfig struct {
    ID          string            `json:"id"`          // Deterministic SHA-256 hash of (Protocol + Server + Port + UUID/Secret)
    RawURI      string            `json:"raw_uri"`     // Original link or JSON line
    Protocol    Protocol          `json:"protocol"`    // Protocol enum
    Server      string            `json:"server"`      // Remote host or IP
    Port        int               `json:"port"`        // Remote port
    UUID        string            `json:"uuid"`        // User ID / Password / Key
    Security    string            `json:"security"`    // tls / reality / none
    SNI         string            `json:"sni"`         // Server Name Indication
    Host        string            `json:"host"`        // HTTP Host header / WS Host
    Path        string            `json:"path"`        // WS / gRPC / HTTP path
    Type        string            `json:"network"`     // tcp / ws / grpc
    Remark      string            `json:"remark"`      // Original remark or name
    ExtraParams map[string]string `json:"extra_params"`// Protocol-specific attributes (pbk, sid, fp, etc.)
}
```

*Validation Rules:*
- `ID`: Non-empty, 64-char hex string (used for deduplication).
- `Server`: Must be a valid IPv4/IPv6 or fully qualified domain name.
- `Port`: Integer in range `1 <= Port <= 65535`.
- `UUID`: Required for VLESS, VMess, Trojan, and Shadowsocks.

---

### 1.2 ProbeDefinition & ProbeResult (Value Objects)
Represents a configurable service verification target and its evaluated outcome.

```go
type ProbeStatus string

const (
    StatusPending ProbeStatus = "PENDING"
    StatusRunning ProbeStatus = "RUNNING"
    StatusPass    ProbeStatus = "PASS"
    StatusFail    ProbeStatus = "FAIL"
    StatusSkipped ProbeStatus = "SKIPPED"
)

type ProbeDefinition struct {
    ID             string            `json:"id"`             // e.g., "gemini", "ai-studio", "ipleak"
    Name           string            `json:"name"`           // Display title
    TargetURL      string            `json:"target_url"`     // Evaluation endpoint
    Method         string            `json:"method"`         // GET, HEAD, POST
    ExpectedStatus int               `json:"expected_status"`// e.g., 200
    RequiredBody   string            `json:"required_body"`  // Substring or regex assertion (optional)
    Headers        map[string]string `json:"headers"`        // Custom HTTP headers
    TimeoutMs      int               `json:"timeout_ms"`     // Max evaluation duration
    Enabled        bool              `json:"enabled"`        // Whether probe runs in Stage 2
}

type ProbeResult struct {
    ProbeID     string        `json:"probe_id"`
    ProbeName   string        `json:"probe_name"`
    Status      ProbeStatus   `json:"status"`
    HTTPCode    int           `json:"http_code"`
    LatencyMs   int64         `json:"latency_ms"`
    ErrorMessage string       `json:"error_message,omitempty"`
}
```

---

### 1.3 IPEnrichment (Value Object)
Telemetry retrieved from `https://ipleak.net/json/` through the tested proxy tunnel.

```go
type IPEnrichment struct {
    ExitIP      string `json:"exit_ip"`
    CountryCode string `json:"country_code"`
    CountryName string `json:"country_name"`
    ISP         string `json:"isp"`
    ASN         string `json:"asn"`
    QueryTimeMs int64  `json:"query_time_ms"`
    Success     bool   `json:"success"`
}
```

---

### 1.4 ProxyEvaluationReport (Aggregate Entity)
Aggregates all evaluation results for a single proxy instance.

```go
type EvaluationState string

const (
    StateQueued        EvaluationState = "QUEUED"
    StateStage1Testing EvaluationState = "STAGE1_HANDSHAKE"
    StateStage1Failed  EvaluationState = "STAGE1_DEAD"
    StateStage2Testing EvaluationState = "STAGE2_PROBING"
    StateComplete      EvaluationState = "COMPLETE"
    StateCancelled     EvaluationState = "CANCELLED"
)

type ProxyEvaluationReport struct {
    Config          ProxyConfig            `json:"config"`
    State           EvaluationState        `json:"state"`
    Stage1LatencyMs int64                  `json:"stage1_latency_ms"` // Ping latency
    Stage1Passed    bool                   `json:"stage1_passed"`
    ProbeResults    map[string]ProbeResult `json:"probe_results"`     // Keyed by ProbeID
    IPInfo          IPEnrichment           `json:"ip_info"`
    IsClean         bool                   `json:"is_clean"`          // Satisfies user qualification filter
    EnrichedURI     string                 `json:"enriched_uri"`      // Subscription URI with modified remark
    Timestamp       int64                  `json:"timestamp"`
}
```

---

## 2. State Machine Transitions

```text
[ Raw Input (File / Text / Sub URL) ]
                 │
                 ▼ (Parser & Deduplication)
             [ QUEUED ]
                 │
                 ▼ (Stage 1: TCP/TLS Handshake)
        [ STAGE1_HANDSHAKE ]
           /            \
          / (Timeout/Refused) \ (Success)
         ▼                      ▼
  [ STAGE1_DEAD ]        [ STAGE2_PROBING ]
  (Pipeline Ends)               │
                         ┌──────┴────────────────┐
                         │ (Parallel Probes)     │
                         ├─► Gemini Probe        │
                         ├─► AI Studio Probe     │
                         ├─► Custom Probes       │
                         └─► ipleak Enrichment   │
                                │
                                ▼ (All Probes Settled)
                         [ COMPLETE ]
                                │
                         (Evaluate Policy)
                           /          \
                          ▼            ▼
                   [ IsClean=true ]   [ IsClean=false ]
```

---

## 3. Session State & Metrics

```go
type ScanSessionStats struct {
    SessionID       string `json:"session_id"`
    StartTime       int64  `json:"start_time"`
    TotalIngested   int    `json:"total_ingested"`
    Deduplicated    int    `json:"deduplicated"`
    Stage1Alive     int    `json:"stage1_alive"`
    Stage1Dead      int    `json:"stage1_dead"`
    Stage2Clean     int    `json:"stage2_clean"`
    Stage2Failed    int    `json:"stage2_failed"`
    IsRunning       bool   `json:"is_running"`
}
```
