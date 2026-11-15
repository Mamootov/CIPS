# Data Model: Diagnostic Logging & Interception Inspector

## Domain Entities & Enums

### 1. ErrorLayer
Identifies the exact architectural boundary where an evaluation failure occurred.

```go
type ErrorLayer string

const (
    LayerNone          ErrorLayer = "NONE"
    LayerLocalEngine   ErrorLayer = "LOCAL_ENGINE"   // Host OS socket, process crash, port bind refused
    LayerUpstreamProxy ErrorLayer = "UPSTREAM_PROXY" // Remote proxy node unreachable, TLS reset, dropped SYN
    LayerTargetService ErrorLayer = "TARGET_SERVICE" // Target platform interception (Cloudflare, Google, Geoblock)
)
```

### 2. InterceptionType
Categorizes the specific defense or failure mechanism encountered.

```go
type InterceptionType string

const (
    InterceptNone                 InterceptionType = "NONE"
    InterceptCloudflareChallenge  InterceptionType = "CLOUDFLARE_CHALLENGE"  // Turnstile, Under-Attack mode, 403/503
    InterceptGoogleUnusualTraffic InterceptionType = "GOOGLE_UNUSUAL_TRAFFIC" // Google /sorry/index captcha redirect
    InterceptGeoRestricted        InterceptionType = "GEO_RESTRICTED"         // Region not supported banner
    InterceptLocalEngineError     InterceptionType = "LOCAL_ENGINE_ERROR"     // connectex / socket refused
    InterceptUpstreamTimeout      InterceptionType = "UPSTREAM_TIMEOUT"       // Remote node timed out
    InterceptUnknown              InterceptionType = "UNKNOWN"
)
```

### 3. DiagnosticInfo (Value Object)
Enriches `domain.ProbeResult` with transparent, verifiable inspection telemetry.

```go
type DiagnosticInfo struct {
    Layer            ErrorLayer       `json:"layer"`
    InterceptionType InterceptionType `json:"interception_type"`
    FinalURL         string           `json:"final_url,omitempty"`
    HTTPStatus       int              `json:"http_status,omitempty"`
    ServerHeader     string           `json:"server_header,omitempty"`
    RayID            string           `json:"ray_id,omitempty"`
    PageTitle        string           `json:"page_title,omitempty"`
    Summary          string           `json:"summary"`
}
```

### 4. Integration with ProbeResult

```go
type ProbeResult struct {
    ProbeID      string          `json:"probe_id"`
    ProbeName    string          `json:"probe_name"`
    Status       ProbeStatus     `json:"status"`
    HTTPCode     int             `json:"http_code"`
    LatencyMs    int64           `json:"latency_ms"`
    ErrorMessage string          `json:"error_message,omitempty"`
    Diagnostic   *DiagnosticInfo `json:"diagnostic,omitempty"` // Added for Feature 002
}
```
