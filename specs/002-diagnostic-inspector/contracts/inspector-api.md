# Contract: Response Inspector Interface

## Overview
Defines the functional contract for `internal/probe/inspector.go`.

```go
package probe

import (
    http "github.com/bogdanfinn/fhttp"
    "cips/internal/domain"
)

type IResponseInspector interface {
    // Inspect analyzes response headers, status, redirects, and body to produce structured diagnostic telemetry.
    Inspect(resp *http.Response, body []byte, reqErr error, targetURL string) domain.DiagnosticInfo
}
```

## Diagnostic Rules Matrix

| Condition | Detected Layer | Interception Type | Diagnostic Action |
|:---|:---:|:---:|:---|
| `reqErr` contains `connectex: ... actively refused it` | `LOCAL_ENGINE` | `LOCAL_ENGINE_ERROR` | Flag local socket bind failure |
| `reqErr` contains `timeout` or `context deadline` | `UPSTREAM_PROXY` | `UPSTREAM_TIMEOUT` | Flag proxy node latency / unreachable |
| `resp.StatusCode` in [403, 503] && `Server: cloudflare` | `TARGET_SERVICE` | `CLOUDFLARE_CHALLENGE` | Extract `cf-ray`, page title |
| `Location` header or `final_url` contains `google.com/sorry` | `TARGET_SERVICE` | `GOOGLE_UNUSUAL_TRAFFIC` | Extract captcha landing URL |
| Visible body contains `isn't supported in your country` | `TARGET_SERVICE` | `GEO_RESTRICTED` | Extract geoblock notice |
| `resp.StatusCode` == 200 && no blocked tokens | `NONE` | `NONE` | Mark `PASS` |
