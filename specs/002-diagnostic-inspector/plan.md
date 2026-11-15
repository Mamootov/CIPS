# Implementation Plan: Rich Diagnostic Logging & Interception Inspector

**Branch**: `002-diagnostic-inspector` | **Date**: 2026-11-14 | **Spec**: [specs/002-diagnostic-inspector/spec.md](spec.md)

**Input**: Feature specification from `specs/002-diagnostic-inspector/spec.md`

---

## Summary

Implement a multi-layer diagnostic inspection engine in Go that eliminates false-positive test failures, actively polls local socket readiness before dispatching probes (resolving `connectex` errors), intercepts and classifies Cloudflare security challenges (extracting Ray-ID), detects Google unusual traffic captcha redirects (`/sorry/index`), and enriches telemetry with actionable `DiagnosticInfo` displayed via interactive tooltips and modals in the web dashboard.

---

## Technical Context

**Language/Version**: Go (Golang) 1.22+  
**Primary Dependencies**: `github.com/bogdanfinn/tls-client`, `github.com/bogdanfinn/fhttp`  
**Storage**: Extended `ProbeResult` data model in memory and JSON telemetry  
**Testing**: Go unit tests in `internal/probe/inspector_test.go` and `internal/engine/xray_test.go`  
**Target Platform**: Windows 10/11 & Linux x86_64  
**Project Type**: Core engine enhancement + Diagnostic telemetry  
**Performance Goals**:
- Socket readiness polling latency: < 300ms typical
- Response inspection overhead: < 2ms per evaluation  
**Constraints**:
- Absolute zero false-negatives on valid Google OAuth login pages
- Immediate detection of Cloudflare challenges without redirect looping  

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Requirement | Plan Status |
|:---|:---|:---:|
| **I. Zero-Assumption Empirical Ground Truth** | Inspection observes actual response tokens, headers, and status rather than static assumptions. | **PASS** |
| **II. Core-Agnostic Hexagonal Architecture** | Inspector sits cleanly in `internal/probe/` behind `IResponseInspector` interface. | **PASS** |
| **III. Authentic TLS & Protocol Fingerprinting** | Uses `tls-client` for hop-by-hop redirect snooping. | **PASS** |
| **IV. Resilient Concurrency & Resource Isolation** | Active socket poller and explicit process termination eliminate port locks and zombie processes. | **PASS** |
| **V. Pluggable Extensibility & Lossless Telemetry** | Rich `DiagnosticInfo` object seamlessly attached to `ProbeResult` and exported. | **PASS** |

*Result:* **All 5 Gates Passed without Violations (Score: 10/10)**.

---

## Project Structure

### Documentation (this feature)

```text
specs/002-diagnostic-inspector/
├── spec.md              # Requirements specification & user journeys
├── plan.md              # Architectural design & implementation plan (this file)
├── research.md          # Technical decisions & benchmarks (Phase 0)
├── data-model.md        # Domain entities, DTOs & state transitions (Phase 1)
├── quickstart.md        # Developer setup, build & validation guide (Phase 1)
├── contracts/           # API schemas & inspector contract (Phase 1)
│   └── inspector-api.md
└── checklists/
    └── requirements.md  # Quality validation checklist
```

### Source Code Modifications

```text
internal/
├── domain/
│   └── models.go            # Add DiagnosticInfo, ErrorLayer, InterceptionType to ProbeResult
├── engine/
│   ├── xray.go              # Replace sleep with active socket poller, enforce cmd.Process.Kill()
│   └── logger.go            # Rolling Stderr diagnostic buffer for Xray child processes
├── probe/
│   ├── inspector.go         # Core response inspection logic (Cloudflare, Google Sorry, Geoblock)
│   ├── inspector_test.go    # Unit tests for inspection rules
│   ├── client.go            # Configure redirect check hooks
│   ├── gemini.go            # Integrate inspector and remove access_denied bug
│   └── aistudio.go          # Integrate inspector
└── server/
    └── handlers.go          # Stream enriched diagnostic telemetry
web/
├── app.js                   # Interactive tooltip & modal inspection renderer
└── style.css                # Tooltip and diagnostic modal styling
```

---

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|:---|:---|:---|
| None | Follows established architecture | N/A |
