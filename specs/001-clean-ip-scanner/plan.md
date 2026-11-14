# Implementation Plan: Clean IP Scanner (CIPS)

**Branch**: `001-clean-ip-scanner` | **Date**: 2026-11-13 | **Spec**: [specs/001-clean-ip-scanner/spec.md](spec.md)

**Input**: Feature specification from `specs/001-clean-ip-scanner/spec.md`

---

## Summary

Build an enterprise-grade, high-throughput proxy validator in Go that ingests mixed configurations (VLESS, VMess, Trojan, Shadowsocks, SOCKS5, HTTP), executes a two-stage parallel pipeline (rapid reachability handshake followed by authentic TLS-fingerprinted service probing for Google Gemini and Google AI Studio), enriches surviving nodes via `ipleak.net`, and renders a real-time matrix dashboard on an embedded local web server (`127.0.0.1:19090`) with one-click export of clean subscription feeds.

---

## Technical Context

**Language/Version**: Go (Golang) 1.22+  
**Primary Dependencies**:
- `github.com/bogdanfinn/tls-client` (JA3/JA4 authentic Chrome 124 TLS & HTTP/2 impersonation)
- `github.com/spf13/cobra` (CLI structure, `--headless`, `--port`, `--concurrency` flags)
- Go Standard Library `net/http` & `embed.FS` (embedded zero-dependency web dashboard)  
**Storage**: Transient in-memory pipeline state + structured JSON session logs (`session-<timestamp>.json`) and clean subscription files (`clean_sub.txt`)  
**Testing**: Go native unit & contract tests (`go test -v ./...`) with mock local proxy listeners  
**Target Platform**: Windows 10/11 & Linux x86_64 (Single self-contained binary)  
**Project Type**: Hybrid CLI tool + Embedded Local Web Dashboard  
**Performance Goals**:
- Stage 1 Reachability Triage: >= 50 configurations/sec
- Stage 2 Probing: Bounded concurrency up to 50 concurrent tunnels without packet drop
- Dashboard SSE Latency: < 200ms per state transition  
**Constraints**:
- Absolute resource safety: 0 leaked ports, 0 orphan subprocesses upon cancellation or crash
- Zero false negatives from bot mitigation on Google services  
**Scale/Scope**: Batches up to 5,000 proxy configurations per execution session  

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Requirement | Plan Status |
|:---|:---|:---:|
| **I. Zero-Assumption Empirical Ground Truth** | No pre-filtering by GeoIP, ASN, or country. ipleak metadata used exclusively for enrichment/tagging. | **PASS** |
| **II. Core-Agnostic Hexagonal Architecture** | `IProxyEngine` port isolates business logic from underlying proxy binaries (Xray/Sing-box). | **PASS** |
| **III. Authentic TLS & Protocol Fingerprinting** | `bogdanfinn/tls-client` enforces authentic Chrome JA3/JA4 ClientHello and HTTP/2 frames. | **PASS** |
| **IV. Resilient Concurrency & Resource Isolation** | Bounded worker semaphore, ephemeral port lifecycle registry, graceful context teardown. | **PASS** |
| **V. Pluggable Extensibility & Lossless Telemetry** | Modular `IProbe` interface for adding custom targets; lossless JSON + annotated subscription export. | **PASS** |

*Result:* **All 5 Gates Passed without Violations (Score: 10/10)**.

---

## Project Structure

### Documentation (this feature)

```text
specs/001-clean-ip-scanner/
├── spec.md              # Requirements specification & user journeys
├── plan.md              # Architectural design & implementation plan (this file)
├── research.md          # Technical decisions & benchmarks (Phase 0)
├── data-model.md        # Domain entities, DTOs & state transitions (Phase 1)
├── quickstart.md        # Developer setup, build & validation guide (Phase 1)
├── contracts/           # API schemas, probe interfaces & export format (Phase 1)
│   ├── probe-interface.md
│   ├── api-schema.json
│   └── export-format.md
└── checklists/
    └── requirements.md  # Quality validation checklist
```

### Source Code Layout (repository root)

```text
cips/
├── cmd/
│   └── cips/
│       └── main.go              # CLI entrypoint, flag parsing, server launcher
├── internal/
│   ├── domain/                  # Pure enterprise models & port interfaces
│   │   ├── config.go            # ProxyConfig entity & protocol types
│   │   ├── probe.go             # IProbe interface & ProbeResult DTO
│   │   ├── engine.go            # IProxyEngine interface & tunnel lifecycle
│   │   └── report.go            # EvaluationReport aggregate & stats
│   ├── ingestion/               # Intake parsers & deduplication
│   │   ├── parser.go            # VLESS, VMess, Trojan, SS URI parsers
│   │   ├── subscription.go      # Remote HTTP/Base64 subscription fetcher
│   │   └── deduplicate.go       # Hash-based duplicate stripper
│   ├── engine/                  # Proxy core adapters
│   │   ├── xray.go              # Xray multi-inbound manager & subprocess driver
│   │   └── acquire.go           # Smart auto-detection & release downloader
│   ├── probe/                   # Service verification plugins
│   │   ├── gemini.go            # Google Gemini evaluator
│   │   ├── aistudio.go          # Google AI Studio evaluator
│   │   ├── ipleak.go            # ipleak.net JSON metadata enricher
│   │   └── custom.go            # User-defined URL probe
│   ├── pipeline/                # Two-stage concurrent orchestrator
│   │   ├── pool.go              # Worker semaphore & port allocator
│   │   └── orchestrator.go      # Stage 1 handshake -> Stage 2 probe coordinator
│   ├── server/                  # Embedded Web UI backend
│   │   ├── server.go            # HTTP router, port binder & browser opener
│   │   ├── sse.go               # Server-Sent Events live stream broadcaster
│   │   └── handlers.go          # /api/ingest, /api/scan, /api/export
│   └── exporter/                # Output formatting
│       ├── subscription.go      # Clean subscription generator with smart tags
│       └── json.go              # Lossless JSON telemetry logger
├── web/                         # Embedded frontend single-page application
│   ├── index.html               # Semantic HTML5 layout with Drag-and-Drop area
│   ├── app.js                   # Reactive matrix grid & SSE consumer
│   └── style.css                # Clean dark-themed CSS styling
├── go.mod
├── go.sum
└── .gitignore
```

---

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|:---|:---|:---|
| None | All patterns follow standard Go conventions and Hexagonal Architecture | N/A |
