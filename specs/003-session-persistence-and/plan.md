# Implementation Plan: 003 - Session Persistence and Ground Truth Validation Accuracy

**Feature Branch**: `003-session-persistence-and`  
**Status**: Implemented & Ratified  

---

## 1. Technical Context & Root Cause Analysis

### 1.1 The Google Accounts 200 OK False-Positive
Google serves the client-side SPA shell (`https://gemini.google.com/`) with `HTTP 200 OK` globally. When accessed unauthenticated, `/app` and `aistudio.google.com` redirect to `https://accounts.google.com/ServiceLogin` which returns `HTTP 200 OK` from every country. Checking solely for status 200 and English geoblock tokens falsely passes proxies where Google actually denies service once authenticated or localized.

**Solution**:
1. Probe Google's first-party edge localization via `https://www.google.com/preferences?hl=en` using Chrome TLS fingerprinting.
2. Read the assigned country code attribute (`lang="en-XX"` or `hl=en-XX`).
3. If Google assigns an unserviced country (`IR`, `RU`, `BY`, `CN`, `KP`, `SY`, `CU`, `AF`), halt and flag `GEO_RESTRICTED`.
4. If Google assigns a serviced country (e.g. `FI`, `NL`, `DE`, `US`), proceed to live service evaluation.

### 1.2 Multi-Inbound Batch Failure & Missing Results
Xray multi-inbound configuration binds 20 local ports in a single process. On Windows:
- Process spawning and port binding under load took >1.5s, triggering false timeouts in `WaitForPortsReady`.
- One invalid configuration (malformed Reality public key or dead IP) caused Xray to exit with code 23, causing the entire 20-node batch to be skipped without recording probe results.

**Solution**:
1. Increase `WaitForPortsReady` timeout to 5000ms.
2. If `StartPool` returns an error for a chunk, fall back to evaluating each configuration in an isolated single-node pool (`StartPool(ctx, []ProxyConfig{r.Config})`).
3. For genuinely unbindable nodes, populate explicit `ProbeResult` failures with `LOCAL_ENGINE_ERROR` diagnostics and broadcast the update.

---

## 2. Component Architecture & Changes

```
┌────────────────────────────────────────────────────────┐
│                        web/app.js                      │
│  - loadInitialSession() via GET /api/session           │
│  - Clear All via POST /api/reset                       │
│  - Real-time SSE listener + ✗ N/A complete fallback    │
└───────────────────────────┬────────────────────────────┘
                            │ REST / SSE
┌───────────────────────────▼────────────────────────────┐
│                  internal/server/handlers.go           │
│  - HandleSessionState (GET /api/session)               │
│  - HandleReset (POST /api/reset)                       │
│  - UpdateReport (thread-safe real-time memory sync)    │
│  - HandleExport (immediate clean node export)          │
└───────────────────────────┬────────────────────────────┘
                            │
┌───────────────────────────▼────────────────────────────┐
│               internal/pipeline/orchestrator.go        │
│  - 5000ms socket readiness poller                      │
│  - Single-node fallback on batch pool error            │
│  - Config ID propagation into context                  │
└──────────────┬───────────────────────────┬─────────────┘
               │                           │
┌──────────────▼────────────┐ ┌────────────▼─────────────┐
│ internal/probe/google_geo │ │ internal/probe/inspector │
│ - Anycast lang="en-XX"    │ │ - Persian token check    │
│ - Banned policy matching  │ │ - available-regions hop  │
│ - Config-keyed cache      │ │ - OAuth filter           │
└───────────────────────────┘ └──────────────────────────┘
```

---

## 3. Verification & Testing Strategy

- `internal/probe/google_geo_test.go`: Verify regex extraction of country codes from HTML bodies (`lang="en-DE"`, `lang="en-IR"`) and policy enforcement.
- `internal/probe/inspector_test.go`: Verify interception of `available-regions` redirects, OAuth login passes, and Persian geoblock text.
- Full system compilation: `go test -v ./...` and `go build -o cips.exe ./cmd/cips`.
