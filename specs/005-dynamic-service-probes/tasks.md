# Implementation Tasks: 005 - Dynamic Multi-Service Probe Catalog

**Feature Branch**: `005-dynamic-service-probes`  
**Status**: Pending Implementation  

---

## Task Breakdown

### Phase 1: Universal Inspector Cleanup
- [x] **T001**: Strip non-English / Persian tokens from `internal/probe/inspector.go`, keeping universal English geoblock and Cloudflare signals.
- [x] **T002**: Update `internal/probe/inspector_test.go` to verify English geoblock and Cloudflare assertions.

### Phase 2: Native ChatGPT & Claude Probes
- [x] **T003**: Implement `ChatGPTProbe` in `internal/probe/chatgpt.go` testing `https://chatgpt.com/` with Cloudflare challenge and OpenAI geoblock inspection.
- [x] **T004**: Implement `ClaudeProbe` in `internal/probe/claude.go` testing `https://claude.ai/` with redirect interception for `supported-countries`.
- [x] **T005**: Add unit tests in `internal/probe/ai_probes_test.go`.

### Phase 3: Probe Factory & Dynamic Pipeline Configuration
- [x] **T006**: Implement `BuildProbeSuite` in `internal/probe/factory.go` accepting service IDs and custom probe definitions.
- [x] **T007**: Update `HandleScanStart` in `internal/server/handlers.go` to parse `services` and `custom_probes` and dynamically configure orchestrator probes.
- [x] **T008**: Update `EnrichProxyRemark` in `internal/exporter/subscription.go` to format remark tags matching active passed services.

### Phase 4: Dynamic Frontend Grid & Service Selector
- [x] **T009**: Add service selector chips (`Gemini`, `AI Studio`, `ChatGPT`, `Claude`, `+ Custom`) and custom probe modal to `web/index.html`.
- [x] **T010**: Add CSS styling for service chips, toggle states, and custom modal in `web/style.css`.
- [x] **T011**: Implement `renderTableHeaders()` in `web/app.js` to dynamically generate matrix columns based on active services.
- [x] **T012**: Update `populateRowContent()` in `web/app.js` to render checkpoint cells matching active columns.
- [x] **T013**: Wire custom probe creation modal and forward active services in scan start payload in `web/app.js`.

### Phase 5: Verification & Delivery
- [x] **T014**: Run full unit test suite `go test -v ./...`.
- [x] **T015**: Recompile binary `cips.exe` and verify live dynamic matrix on port 19090.
