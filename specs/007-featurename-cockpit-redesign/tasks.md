# Implementation Tasks: 007 - Cockpit UI/UX Redesign

**Feature Branch**: `007-featurename-cockpit-redesign`  
**Status**: Implemented & Verified  

---

## Task Breakdown

### Phase 1: Structure & Semantic Redesign (`web/index.html`)
- [x] **T001**: Restructure `web/index.html` with clean top navigation header (Branding, 4 Tabs, Status indicator).
- [x] **T002**: Construct View 1 (`#view-dashboard`): Sentry KPI summary cards, Terralogix filter pills dock, and Modal-style matrix table.
- [x] **T003**: Construct View 2 (`#view-map`): Three.js WebGL canvas container, Left Inspector Card, Right Cluster Card, and Bottom-Right 2D Global Traffic Map SVG widget.
- [x] **T004**: Construct View 3 (`#view-probes`): Probe Studio evaluator cards grid and Custom Probe Builder form.
- [x] **T005**: Construct View 4 (`#view-subscriptions`): Live feed URL box, preset filter rows, and Mobile QR code card.
- [x] **T006**: Add Floating Bottom Command Dock (`#floating-dock`), `#ingest-modal`, and `#settings-modal`.

### Phase 2: Cyberpunk / Cockpit Styling (`web/style.css`)
- [x] **T007**: Define modern CSS variables: dark palette (`#02050a` base, `#080e1a` cards, `#10b981` emerald, `#38bdf8` cyan, `#d946ef` neon purple, `#ec4899` neon pink).
- [x] **T008**: Style the 4 top navigation tabs and active pill indicators.
- [x] **T009**: Style the 4 Sentry KPI metric cards and stacked health distribution bar.
- [x] **T010**: Style the Floating Bottom Command Dock (frosted glass capsule with glow buttons) and settings sliders.
- [x] **T011**: Style the 3D Globe viewport overlays (Left Inspector Card, Right Cluster Card, 2D Traffic Map card).

### Phase 3: JavaScript Controllers & 3D WebGL Engine (`web/app.js`)
- [x] **T012**: Implement 4-view tab switching router (`switchView(tabId)`).
- [x] **T013**: Embed Three.js 3D Earth Globe with `web/assets/earth-tactical-dark.jpg` base64 texture, OrbitControls, and starfield.
- [x] **T014**: Implement city coordinate mapping with constellation jittering for multi-node hubs.
- [x] **T015**: Implement Geodesic Great-Circle Orbit Trajectory (`GeodesicOrbitCurve`) preventing curve penetration into Earth interior.
- [x] **T016**: Implement continuous high-speed pure-white photon particle flow and proper Three.js mesh disposal on node switch.
- [x] **T017**: Render 2D Vector Global Traffic Map SVG with country borders and green vs red traffic corridors.
- [x] **T018**: Implement real-time KPI metrics calculations (`p50`, `p95`, average speed, health bar widths) and live SSE named event hydration (`item_update`, `stats`).
- [x] **T019**: Implement automatic host egress origin discovery (`GET /api/origin`) and dynamic origin anchoring.
- [x] **T020**: Implement universal 249-country emoji flag generator and dynamic Dashboard/Globe country filter pills.
- [x] **T021**: Implement automatic download speed sorting for Available proxies.
- [x] **T022**: Implement dynamic Probe Studio timeouts synchronized with engine settings sliders.

### Phase 4: Backend Pipeline & Concurrency Enhancements
- [x] **T023**: Implement `DetectHostOrigin` in `internal/probe/ip_intel.go` and `GET /api/origin` endpoint in `internal/server/handlers.go`.
- [x] **T024**: Implement `SetConcurrency(c int)` and `SetTimeouts(p1, p2 int)` in `internal/pipeline/orchestrator.go`.
- [x] **T025**: Parallelize Stage 2 fallback execution via goroutines and `sync.WaitGroup` to eliminate sequential bottleneck.
- [x] **T026**: Scope Xray engine mutex in `internal/engine/xray.go` and optimize socket readiness budget to prevent worker deadlocks.
- [x] **T027**: Implement thread-safe `Clone()` snapshotting in `internal/domain/models.go` to resolve concurrent map iteration panics.

### Phase 5: Verification & Delivery
- [x] **T028**: Run Go test suite `go test -v ./...` to ensure backend integration integrity (16/16 tests passing).
- [x] **T029**: Verify live multi-view cockpit on port 19090 with Computer Vision screenshot validation across all views.
- [x] **T030**: Validate that no `git push` is executed and local commits adhere to Conventional Commits.
