# Implementation Tasks: 006 - Deep Xray Core Diagnostic Logging

**Feature Branch**: `006-deep-xray-core`  
**Status**: Pending Implementation  

---

## Task Breakdown

### Phase 1: Xray Core Logging & Engine Telemetry
- [x] **T001**: Elevate Xray loglevel from `"none"` to `"warning"` in `internal/engine/xray.go`.
- [x] **T002**: Enhance `RollingLogBuffer` in `internal/engine/logger.go` with `GetLogsForPort(port int) string`.
- [x] **T003**: Extend `DiagnosticInfo` with `EngineLog` and add new `InterceptionType` enums (`BAD_REALITY_KEY`, `BAD_CIPHER`, `BAD_UUID`, `UPSTREAM_REFUSED`, `HANDSHAKE_FAILED`) in `internal/domain/models.go`.

### Phase 2: Universal Inspector Granular Error Parsing
- [x] **T004**: Update `InspectResponse` in `internal/probe/inspector.go` to parse engine stderr lines for specific error signatures.
- [x] **T005**: Add unit tests in `internal/probe/inspector_test.go` verifying classification of Reality, cipher, and handshake errors.

### Phase 3: UI & Diagnostic Modal Integration
- [x] **T006**: Add `#modal-engine-log` container to `#diag-modal` in `web/index.html`.
- [x] **T007**: Add monospace code box styling `.diag-log-box` in `web/style.css`.
- [x] **T008**: Update `renderCheckpoint` in `web/app.js` with distinct labels: `⚠️ Bad Reality`, `⚠️ Bad Cipher`, `⚠️ Bad UUID`, `⚠️ Remote Refused`, `⚠️ Handshake Failed`.
- [x] **T009**: Update `openDiagModal` in `web/app.js` to render the raw engine log snippet.

### Phase 4: Verification & Delivery
- [x] **T010**: Run complete test suite with `go test -v ./...`.
- [x] **T011**: Recompile `cips.exe` and verify live web dashboard on port 19090.
