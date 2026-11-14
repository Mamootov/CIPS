---
description: "Task list for Clean IP Scanner (CIPS) implementation"
---

# Tasks: Clean IP Scanner (CIPS)

**Input**: Design documents from `specs/001-clean-ip-scanner/`  
**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md`  
**Tests**: Unit tests included for core protocol parsing and contract verification.  
**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

---

## Format: `- [ ] [TaskID] [P?] [Story?] Description with file path`

- **[P]**: Can run in parallel (independent files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., `US1`, `US2`, `US3`, `US4`)
- Exact file paths are specified in every task description.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization, Go module configuration, and directory layout.

- [x] T001 Initialize Go module `cips` and install core dependencies (`github.com/bogdanfinn/tls-client`, `github.com/spf13/cobra`) in `go.mod`
- [x] T002 [P] Setup directory tree (`cmd/cips`, `internal/domain`, `internal/ingestion`, `internal/engine`, `internal/probe`, `internal/pipeline`, `internal/server`, `internal/exporter`, `web`) per `plan.md`
- [x] T003 [P] Implement core binary auto-detector and release downloader (`xray` / `xray.exe`) with mirror fallback in `internal/engine/acquire.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core domain entities, interfaces, and port allocation infrastructure that MUST be complete before ANY user story can begin.

⚠️ **CRITICAL**: No user story work can begin until this foundational phase is complete.

- [x] T004 Define domain entities and value objects (`ProxyConfig`, `Protocol`, `ProbeStatus`, `ProbeDefinition`, `ProbeResult`, `IPEnrichment`, `ProxyEvaluationReport`, `ScanSessionStats`) with strict field validations in `internal/domain/models.go`
- [x] T005 [P] Define `IProxyEngine` and `EngineInstance` interfaces with port-lifecycle contracts in `internal/domain/engine.go`
- [x] T006 [P] Define `IProbe` interface and service evaluation contracts in `internal/domain/probe.go`
- [x] T007 [P] Implement dynamic ephemeral port allocator and resource lifecycle registry in `internal/pipeline/port.go`
- [x] T008 Implement Xray multi-inbound adapter and subprocess driver with tag-based routing in `internal/engine/xray.go`

**Checkpoint**: Core foundation ready — User Story 1 implementation can now proceed.

---

## Phase 3: User Story 1 - Multi-Format Ingestion & Two-Stage Parallel Scanning (Priority: P1) 🎯 MVP

**Goal**: Ingest proxy configs (VLESS, VMess, Trojan, Shadowsocks, SOCKS5, HTTP) from files, clipboard, and remote subscriptions; deduplicate by server/port/credentials; execute rapid Stage 1 reachability ping triage.

**Independent Test**: Provide a text file containing 20 mixed valid and dead proxy links via CLI; verify that all protocols are accurately parsed, duplicates are removed, unresponsive nodes are dropped in Stage 1, and alive nodes report accurate ping latency.

### Tests for User Story 1
- [x] T009 [P] [US1] Unit test for URI parser validating VLESS, VMess, Trojan, SS, and SOCKS5 extraction in `internal/ingestion/parser_test.go`

### Implementation for User Story 1
- [x] T010 [P] [US1] Implement protocol URI parsers (`vless://`, `vmess://`, `trojan://`, `ss://`, `socks5://`, `http://`) in `internal/ingestion/parser.go`
- [x] T011 [P] [US1] Implement SHA-256 deduplication and multi-format text extractor in `internal/ingestion/deduplicate.go`
- [x] T012 [US1] Implement remote HTTP/HTTPS subscription fetcher with Base64 payload decoding in `internal/ingestion/subscription.go`
- [x] T013 [US1] Implement Stage 1 rapid TCP/TLS handshake ping engine in `internal/pipeline/ping.go`
- [x] T014 [US1] Implement bounded concurrency worker pool with semaphore and cancellation context in `internal/pipeline/pool.go`
- [x] T015 [US1] Implement pipeline orchestrator coordinating Stage 1 ping and Stage 2 dispatch in `internal/pipeline/orchestrator.go`

**Checkpoint**: At this point, User Story 1 is fully functional and delivers a viable CLI-based reachability scanner MVP.

---

## Phase 4: User Story 2 - Real-Time Pipeline Dashboard & Live Telemetry (Priority: P2)

**Goal**: Serve embedded local web dashboard on port `19090` with real-time SSE streaming and a matrix grid table featuring vertical service headers and checkpoint indicators.

**Independent Test**: Launch `./cips.exe`, verify dashboard loads on `http://127.0.0.1:19090`, drag and drop a proxy file, click Start, and observe real-time matrix row transitions without browser refresh.

- [x] T016 [P] [US2] Create responsive dark-themed dashboard HTML with drag-and-drop zone, control sliders, and matrix grid in `web/index.html`
- [x] T017 [P] [US2] Implement CSS styles for matrix columns, badges, and checkpoint status markers in `web/style.css`
- [x] T018 [US2] Implement frontend SSE consumer, matrix table renderer, and drag-and-drop file upload in `web/app.js`
- [x] T019 [US2] Implement embedded filesystem asset server with automatic port-conflict fallback to adjacent ports in `internal/server/server.go`
- [x] T020 [US2] Implement Server-Sent Events (SSE) broadcaster for streaming scan progress in `internal/server/sse.go`
- [x] T021 [US2] Implement REST endpoints (`/api/ingest`, `/api/scan/start`, `/api/scan/cancel`, `/api/stream`) in `internal/server/handlers.go`

**Checkpoint**: User Stories 1 and 2 operate cohesively; users can interactively run scans and observe real-time telemetry in the browser.

---

## Phase 5: User Story 3 - Service Probe Customization & IP Enrichment via ipleak (Priority: P3)

**Goal**: Pluggable service evaluation with authentic TLS impersonation (`gemini`, `aistudio`, custom probes) and `ipleak.net` metadata enrichment.

**Independent Test**: Run a batch scan with Gemini, AI Studio, and ipleak enabled; verify that surviving proxies bypass Google bot mitigation, display green checkpoints, and show ISP and exit IP details.

- [x] T022 [P] [US3] Implement authentic Chrome 124 TLS client wrapper with HTTP/2 transport in `internal/probe/client.go`
- [x] T023 [P] [US3] Implement Google Gemini evaluation probe (`https://gemini.google.com/app`) in `internal/probe/gemini.go`
- [x] T024 [P] [US3] Implement Google AI Studio evaluation probe (`https://aistudio.google.com/`) in `internal/probe/aistudio.go`
- [x] T025 [P] [US3] Implement ipleak JSON metadata enricher (`https://ipleak.net/json/`) with jitter and fallback in `internal/probe/ipleak.go`
- [x] T026 [US3] Implement custom URL probe provider with user-defined status and body assertions in `internal/probe/custom.go`
- [x] T027 [US3] Integrate Stage 2 probes into pipeline orchestrator and connect with dashboard column matrix in `internal/pipeline/orchestrator.go`

**Checkpoint**: Deep multi-service AI verification and ISP enrichment are fully operational.

---

## Phase 6: User Story 4 - Resilient Fault Isolation, Structured Logs & Clean Export (Priority: P4)

**Goal**: Isolate faults/hangs, write structured session JSON logs, and generate clean subscription exports with smart tags.

**Independent Test**: Ingest hanging/broken proxies, run scan to completion, verify zero system crashes, check `session-<timestamp>.json` on disk, and copy clean subscription output into a proxy client.

- [x] T028 [P] [US4] Implement clean subscription exporter with smart tags (`[COUNTRY] [ISP] [Tags] [Ping]`) in `internal/exporter/subscription.go`
- [x] T029 [P] [US4] Implement structured session telemetry JSON writer in `internal/exporter/json.go`
- [x] T030 [US4] Implement signal handling (SIGINT/SIGTERM) and panic recovery in `internal/pipeline/recovery.go`
- [x] T031 [US4] Implement `/api/export` endpoint with filtering by passed checkpoints in `internal/server/handlers.go`
- [x] T032 [US4] Add "Copy Clean Subscription" and "Download Clean List" UI controls in `web/app.js`

**Checkpoint**: End-to-end scanning, resilience, and export workflows are fully integrated.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: CLI usability, automated browser launch, documentation, and final end-to-end validation.

- [x] T033 [P] Implement CLI root command with Cobra (`--headless`, `--port`, `--concurrency`, `--timeout`) in `cmd/cips/main.go`
- [x] T034 [P] Add browser auto-launcher utility for Windows/Linux in `internal/server/browser.go`
- [x] T035 Execute end-to-end quickstart scenarios in `specs/001-clean-ip-scanner/quickstart.md` and verify single binary build

---

## Dependencies & Execution Order

### Phase Dependencies
- **Phase 1 (Setup)**: Can start immediately.
- **Phase 2 (Foundational)**: Depends on Phase 1; BLOCKS all User Stories.
- **Phase 3 (User Story 1 - MVP)**: Depends on Phase 2; establishes base ingestion and Stage 1 pipeline.
- **Phase 4 (User Story 2 - UI)**: Depends on Phase 2 & Phase 3 data flow.
- **Phase 5 (User Story 3 - Probes)**: Depends on Phase 3 pipeline orchestrator; enriches Stage 2.
- **Phase 6 (User Story 4 - Resilience & Export)**: Depends on Phase 4 and Phase 5 results.
- **Phase 7 (Polish)**: Integrates all components into the final binary.

### Parallel Opportunities
- In Phase 1: `T002` and `T003` can execute in parallel.
- In Phase 2: `T005`, `T006`, and `T007` can execute in parallel.
- In Phase 3: `T009`, `T010`, and `T011` can execute in parallel.
- In Phase 4: `T016` and `T017` can execute in parallel.
- In Phase 5: `T022`, `T023`, `T024`, and `T025` can execute in parallel.
- In Phase 6: `T028` and `T029` can execute in parallel.
- In Phase 7: `T033` and `T034` can execute in parallel.

---

## Implementation Strategy: MVP First

1. **Step 1**: Complete Phase 1 (Setup) and Phase 2 (Foundational).
2. **Step 2**: Implement Phase 3 (User Story 1) → Test reachability triage via CLI (delivers operational MVP).
3. **Step 3**: Implement Phase 4 (User Story 2) → Add web dashboard and real-time matrix grid.
4. **Step 4**: Implement Phase 5 (User Story 3) → Add Gemini, AI Studio, and ipleak probes.
5. **Step 5**: Implement Phase 6 (User Story 4) → Add resilience and export features.
6. **Step 6**: Execute Phase 7 (Polish & Validation) → Compile single release binary.
