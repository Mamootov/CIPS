# Data Model: 006 - Deep Xray Core Diagnostic Logging

---

## 1. Domain Entities Extension

### 1.1 `DiagnosticInfo` Additions
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
    EngineLog        string           `json:"engine_log,omitempty"` // Raw Xray stderr excerpt
}
```

### 1.2 New `InterceptionType` Values
```go
const (
    // ... existing ...
    InterceptBadRealityKey   InterceptionType = "BAD_REALITY_KEY"   // Reality pbk or short ID invalid
    InterceptBadCipher       InterceptionType = "BAD_CIPHER"       // Deprecated or unsupported Shadowsocks cipher
    InterceptBadUUID         InterceptionType = "BAD_UUID"         // Malformed user UUID
    InterceptUpstreamRefused InterceptionType = "UPSTREAM_REFUSED" // Target server rejected connection
    InterceptHandshakeFailed InterceptionType = "HANDSHAKE_FAILED" // TLS / WebSocket / gRPC handshake dropped
)
```
