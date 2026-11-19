# Implementation Plan: 006 - Deep Xray Core Diagnostic Logging and Granular Error Diagnostics

**Feature Branch**: `006-deep-xray-core`  
**Status**: Ready for Implementation  

---

## 1. Technical Architecture & Component Flow

```
┌────────────────────────────────────────────────────────┐
│                        web/app.js                      │
│  - Badges: ⚠️ Bad Reality, ⚠️ Bad Cipher, etc.          │
│  - Modal: Shows Core Engine Log pre box                │
└───────────────────────────┬────────────────────────────┘
                            │ Click badge
┌───────────────────────────▼────────────────────────────┐
│                  internal/domain/models.go             │
│  - DiagnosticInfo.EngineLog string                     │
│  - New InterceptionTypes (BAD_REALITY, BAD_CIPHER, ...)│
└───────────────────────────┬────────────────────────────┘
                            │
┌───────────────────────────▼────────────────────────────┐
│               internal/probe/inspector.go              │
│  - Reads recent Xray stderr for target loopback port   │
│  - Parses signature errors & categorizes specific type │
│  - Sets diag.EngineLog with timestamped stderr snippet │
└───────────────────────────┬────────────────────────────┘
                            │
┌───────────────────────────▼────────────────────────────┐
│                  internal/engine/xray.go               │
│  - "loglevel": "warning"                               │
│  - Indexed RollingLogBuffer (inbound port -> logs)     │
└────────────────────────────────────────────────────────┘
```

---

## 2. Component Design & Changes

### 2.1 Xray Engine Logging (`internal/engine/xray.go`, `logger.go`)
- Change `"loglevel": "none"` to `"loglevel": "warning"`.
- In `RollingLogBuffer`: add `GetLogsForPort(port int) string`.
  - When lines match `\[inbound-(\d+)\]` or port numbers, associate with that port.

### 2.2 Domain Telemetry (`internal/domain/models.go`)
- Add fields to `DiagnosticInfo`:
  ```go
  EngineLog string `json:"engine_log,omitempty"`
  ```
- Add InterceptionTypes:
  ```go
  InterceptBadRealityKey   InterceptionType = "BAD_REALITY_KEY"
  InterceptBadCipher       InterceptionType = "BAD_CIPHER"
  InterceptBadUUID         InterceptionType = "BAD_UUID"
  InterceptUpstreamRefused InterceptionType = "UPSTREAM_REFUSED"
  InterceptHandshakeFailed InterceptionType = "HANDSHAKE_FAILED"
  ```

### 2.3 Universal Inspector (`internal/probe/inspector.go`)
- When a transport/socket error occurs on `127.0.0.1:port`:
  - Fetch recent stderr lines for that port from `GlobalEngineLog`.
  - Parse error signatures:
    - `"reality"` -> `InterceptBadRealityKey`
    - `"cipher"` -> `InterceptBadCipher`
    - `"uuid"` -> `InterceptBadUUID`
    - `"refused"` -> `InterceptUpstreamRefused`
    - `"handshake"` -> `InterceptHandshakeFailed`
  - Populate `diag.EngineLog`.

### 2.4 Diagnostic Modal & Matrix Grid (`web/index.html`, `web/style.css`, `web/app.js`)
- Render dedicated badges:
  - `BAD_REALITY_KEY`: `⚠️ Bad Reality`
  - `BAD_CIPHER`: `⚠️ Bad Cipher`
  - `BAD_UUID`: `⚠️ Bad UUID`
  - `UPSTREAM_REFUSED`: `⚠️ Remote Refused`
  - `HANDSHAKE_FAILED`: `⚠️ Handshake Failed`
- Add monospace log viewer block `#modal-engine-log` in `#diag-modal`.
