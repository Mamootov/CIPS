# Implementation Tasks: 003 - Session Persistence and Ground Truth Validation Accuracy

**Feature Branch**: `003-session-persistence-and`  
**Status**: All Tasks Complete  

---

## Task Breakdown

### Phase 1: Ground-Truth Geolocation & Interception Accuracy
- [x] **T001**: Implement `DetectGoogleRegion` in `internal/probe/google_geo.go` querying `https://www.google.com/preferences?hl=en` and extracting `lang="en-XX"` with 4s timeout.
- [x] **T002**: Define `GoogleBannedCountries` map (`IR`, `RU`, `BY`, `CN`, `KP`, `SY`, `CU`, `AF`) and unit tests in `internal/probe/google_geo_test.go`.
- [x] **T003**: Integrate `DetectGoogleRegion` pre-check and Persian error tokens (`"در کشورتان دردسترس نیست"`) in `internal/probe/gemini.go` and `internal/probe/inspector.go`.
- [x] **T004**: Integrate `DetectGoogleRegion` and `available-regions` redirect interception in `internal/probe/aistudio.go` and `internal/probe/client.go`.

### Phase 2: Orchestration Fault Tolerance & Socket Readiness
- [x] **T005**: Increase `WaitForPortsReady` readiness timeout from 1500ms to 5000ms in `internal/engine/xray.go`.
- [x] **T006**: Implement single-node fallback loop in `internal/pipeline/orchestrator.go` when `StartPool` returns an error on a chunk.
- [x] **T007**: Ensure faulty configurations in failed batches receive explicit `LOCAL_ENGINE_ERROR` reports instead of blank pending cells.
- [x] **T008**: Propagate `config_id` via context in `evaluateProxyProbes` for collision-free cache scoping across ephemeral port recycling.

### Phase 3: Session Persistence & Real-Time Export
- [x] **T009**: Implement thread-safe `UpdateReport` in `ServerContext` (`internal/server/handlers.go`).
- [x] **T010**: Wire `srv.Context().UpdateReport(report)` into `OnItemUpdate` callback in `cmd/cips/main.go` so `GET /api/export` returns live clean nodes mid-scan.
- [x] **T011**: Implement `GET /api/session` (`HandleSessionState`) and `POST /api/reset` (`HandleReset`) in `internal/server/handlers.go` and `internal/server/server.go`.

### Phase 4: Frontend Hydration & UI Controls
- [x] **T012**: Implement `loadInitialSession()` on `DOMContentLoaded` in `web/app.js` to hydrate reports, counters, and scanning state upon page refresh.
- [x] **T013**: Add `🗑️ Clear All Configs` button in `web/index.html` with `.btn-outline-danger` styling in `web/style.css` and confirmation handler in `web/app.js`.
- [x] **T014**: Render `✗ N/A` for completed proxy reports with missing probe results in `web/app.js`.

### Phase 5: Verification & Delivery
- [x] **T015**: Run full unit test suite `go test -v ./...` verifying 100% pass across all packages.
- [x] **T016**: Compile release binary `cips.exe` and test live daemon on port 19090.
