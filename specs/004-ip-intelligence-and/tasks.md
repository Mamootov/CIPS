# Implementation Tasks: 004 - IP Intelligence and Low-Bandwidth Chunked Speedtest

**Feature Branch**: `004-ip-intelligence-and`  
**Status**: Pending Implementation  

---

## Task Breakdown

### Phase 1: Domain & Model Extensions
- [x] **T001**: Extend `IPEnrichment` struct in `internal/domain/models.go` with `ConnectionType`, `ASN`, and `Org` fields.
- [x] **T002**: Extend `ProxyEvaluationReport` struct in `internal/domain/models.go` with `DownloadSpeedMbps` field.

### Phase 2: IP Intelligence Probe
- [x] **T003**: Implement `IPIntelProbe` in `internal/probe/ip_intel.go` querying `https://ipwho.is/` through the proxy tunnel with JSON decoding for `connection.type`, `connection.isp`, and `connection.asn`.
- [x] **T004**: Add multi-tier fallback to `https://ipleak.net/json/` and `https://speed.cloudflare.com/cdn-cgi/trace` in `internal/probe/ip_intel.go`.
- [x] **T005**: Add unit tests for JSON parsing and fallback behavior in `internal/probe/ip_intel_test.go`.

### Phase 3: Low-Bandwidth Speedtest Probe
- [x] **T006**: Implement `SpeedtestProbe` in `internal/probe/speedtest.go` streaming from `https://speed.cloudflare.com/__down?bytes={size}` with a 3000ms hard deadline.
- [x] **T007**: Calculate exact throughput in Mbps ($(\text{bytes} \times 8) / (\text{seconds} \times 1,000,000)$) and set `r.DownloadSpeedMbps`.
- [x] **T008**: Add unit tests for speed throughput computation in `internal/probe/speedtest_test.go`.

### Phase 4: Pipeline & API Integration
- [x] **T009**: Update `HandleScanStart` in `internal/server/handlers.go` to parse `enable_speedtest` and `speedtest_bytes` from request body.
- [x] **T010**: Update `Orchestrator` in `internal/pipeline/orchestrator.go` to conditionally execute Speedtest strictly AFTER AI service probes succeed (at least one AI service passed) and record `DownloadSpeedMbps`.
- [x] **T011**: Update `EnrichProxyRemark` in `internal/exporter/subscription.go` to embed `[Res]`/`[DC]` and `[{speed}Mbps]` tags.

### Phase 5: UI & Dashboard Integration
- [x] **T012**: Add `Enable Speed Test` toggle and payload size dropdown in `web/index.html`.
- [x] **T013**: Add `.badge-speed-fast`, `.badge-speed-medium`, `.badge-speed-slow`, `.pill-residential`, and `.pill-datacenter` in `web/style.css`.
- [x] **T014**: Update `web/app.js` to render the `Speed` column, render `🏠 Res` / `🏢 DC` in `IP / ISP` cell, and pass speedtest options in scan start.

### Phase 6: Verification & Compilation
- [x] **T015**: Run complete test suite with `go test -v ./...`.
- [x] **T016**: Recompile `cips.exe` and verify live web dashboard on port 19090.
