---
description: "Task list for Rich Diagnostic Logging & Interception Inspector implementation"
---

# Tasks: Rich Diagnostic Logging & Interception Inspector

**Input**: Design documents from `specs/002-diagnostic-inspector/`  
**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md`  
**Tests**: Unit tests in `internal/probe/inspector_test.go`.  
**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

---

## Format: `- [ ] [TaskID] [P?] [Story?] Description with file path`

- **[P]**: Can run in parallel (independent files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., `US1`, `US2`, `US3`)
- Exact file paths are specified in every task description.

---

## Phase 1: Setup

**Purpose**: Verification of workspace and baseline build status.

- [x] T001 [P] Verify feature branch environment and dependencies in `specs/002-diagnostic-inspector/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core diagnostic domain models, enums, and logging buffers that MUST be complete before user stories can begin.

⚠️ **CRITICAL**: No user story work can begin until this foundational phase is complete.

- [x] T002 Define `ErrorLayer`, `InterceptionType`, and `DiagnosticInfo` structs and attach to `ProbeResult` in `internal/domain/models.go`
- [x] T003 [P] Implement thread-safe rolling Stderr diagnostic ring buffer in `internal/engine/logger.go`

**Checkpoint**: Foundation ready — User Story 1 implementation can now proceed.

---

## Phase 3: User Story 1 - Granular Interception & Response Classification (Priority: P1) 🎯 MVP

**Goal**: Classify Cloudflare challenges (extract Ray-ID), Google unusual traffic captchas (`/sorry/index`), and true regional geoblocks while eliminating OAuth script false positives.

**Independent Test**: Execute unit test with mock HTTP responses (Cloudflare 403 with `cf-ray`, Google `/sorry/index` redirect, clean sign-in with OAuth scripts, and geoblock notice); verify 100% classification accuracy with zero false positives.

### Tests for User Story 1
- [x] T004 [P] [US1] Unit test for response inspector testing Cloudflare, Google /sorry, OAuth signin, and geoblocks in `internal/probe/inspector_test.go`

### Implementation for User Story 1
- [x] T005 [US1] Implement response inspector matching headers, status, redirects, and body in `internal/probe/inspector.go`
- [x] T006 [US1] Configure redirect check hooks with hop snooping in `internal/probe/client.go`
- [x] T007 [US1] Integrate inspector into `internal/probe/gemini.go` and remove simplistic `access_denied` body check
- [x] T008 [US1] Integrate inspector into `internal/probe/aistudio.go` and `internal/probe/custom.go`

**Checkpoint**: Response classification and OAuth false-positive removal are fully operational.

---

## Phase 4: User Story 2 - Local Engine Health & Socket Polling (Priority: P2)

**Goal**: Replace blind sleep with active loopback socket readiness polling, enforce explicit process termination on pool closure, and isolate `connectex` errors.

**Independent Test**: Start Xray pool with socket poller; verify connections proceed only when port is ready, without `connectex` socket refusal, and child processes terminate cleanly upon pool closure.

- [x] T009 [P] [US2] Implement active socket readiness poller (`WaitForPortReady`) in `internal/engine/xray.go`
- [x] T010 [US2] Attach stderr buffer to Xray subprocess and capture startup error diagnostics in `internal/engine/xray.go`
- [x] T011 [US2] Enforce explicit child process termination and immediate port release on pool closure in `internal/engine/xray.go`

**Checkpoint**: Local engine race conditions and port collision errors are resolved.

---

## Phase 5: User Story 3 - Interactive Diagnostic Telemetry & Tooltip Inspection (Priority: P3)

**Goal**: Render interactive diagnostic tooltips (summary on hover, full inspection modal on click) for failed checkpoints in dashboard matrix table and JSON logs.

**Independent Test**: Load web dashboard on `http://127.0.0.1:19090`, hover over failed checkpoint to view summary badge, and click to open full inspection modal displaying Ray-ID, URL, and error layer.

- [x] T012 [P] [US3] Update diagnostic tooltip CSS styles and inspection modal layout in `web/style.css`
- [x] T013 [US3] Implement inspection modal and hover popover in `web/app.js`
- [x] T014 [US3] Ensure `/api/stream` and `/api/export` serialize full `diagnostic_info` in `internal/server/handlers.go`

**Checkpoint**: Diagnostic telemetry is visible and interactively inspectable in the web interface.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Test suite execution, binary re-compilation, and live validation.

- [x] T015 Execute full test suite (`go test -v ./...`) and verify all inspection rules pass
- [x] T016 Recompile release binary (`go build -o cips.exe ./cmd/cips`) and verify on `http://127.0.0.1:19090`

---

## Dependencies & Execution Order

### Phase Dependencies
- **Phase 1 (Setup)**: Can start immediately.
- **Phase 2 (Foundational)**: Depends on Phase 1; BLOCKS all User Stories.
- **Phase 3 (User Story 1 - MVP)**: Depends on Phase 2; provides core response inspector.
- **Phase 4 (User Story 2)**: Depends on Phase 2; stabilizes Xray socket readiness.
- **Phase 5 (User Story 3)**: Depends on Phase 3 and Phase 4 telemetry.
- **Phase 6 (Polish)**: Integrates all components into release binary.

### Parallel Opportunities
- In Phase 2: `T003` can run in parallel with `T002`.
- In Phase 3: `T004` (tests) can be drafted alongside `T005`.
- In Phase 4: `T009` and `T010` can be worked concurrently.
- In Phase 5: `T012` and `T014` can run in parallel.
