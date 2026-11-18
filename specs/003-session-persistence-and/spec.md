# Feature Specification: 003 - Session Persistence and Ground Truth Validation Accuracy

**Feature Branch**: `003-session-persistence-and`  
**Created**: November 18, 2026  
**Status**: Implemented & Ratified  
**Input**: User feedback on AI validation false-positives, page refresh data loss, and mid-scan subscription copying.

---

## 1. Executive Summary

Feature 003 resolves two critical operational challenges discovered in real-world deployment:
1. **Validation False-Positive Discrepancy**: Raw HTTP GET requests to `gemini.google.com/` and `aistudio.google.com/` returned `HTTP 200 OK` from Google edge servers worldwide (including sanctioned regions) because the landing page is a static Single-Page Application (SPA) shell or redirects unauthenticated visitors to `accounts.google.com/ServiceLogin`. In real browsers, Google enforces regional policy dynamically.
2. **Dashboard State Hydration & Control Gaps**: Browser page refreshes (F5) cleared all client-side state in `web/app.js` despite the Go backend holding the evaluation data in memory. Furthermore, subscription export was only flushed after complete session termination, preventing users from copying verified nodes while testing was in progress.

---

## 2. User Scenarios & Testing

### User Story 1 - Ground-Truth AI Validation via Google First-Party Geolocation (Priority: P1)
As a proxy operator, I want CIPS to detect whether Google's own edge servers permit Gemini and AI Studio on a given proxy IP, so that nodes marked as PASS do not fail with "Gemini isn't supported in your country yet" when opened in my browser.

**Why this priority**: Eliminates false-positives where dead/restricted proxies are falsely tagged as AI-Clean.

**Acceptance Scenarios**:
1. **Given** a proxy whose egress IP is located in a Google-banned region (`IR`, `RU`, `BY`, `CN`, `KP`, `SY`, `CU`, `AF`), **When** Gemini or AI Studio probes execute, **Then** `DetectGoogleRegion` extracts the Google-assigned country tag (`lang="en-XX"`) and marks the probe as `FAIL (GEO_RESTRICTED)` with diagnostic reason.
2. **Given** a proxy located in Russia that routes through clean European BGP Anycast (e.g., Finland or Netherlands), **When** evaluated, **Then** Google assigns `FI` or `NL`, permitting the node to proceed to application evaluation without being preemptively dropped.
3. **Given** an unauthenticated request to `aistudio.google.com`, **When** the service attempts to redirect to `available-regions`, **Then** the client intercepts the hop immediately without misinterpreting the HTTP 200 documentation page as a pass.
4. **Given** a response in Persian containing `"در کشورتان دردسترس نیست"`, **When** inspected, **Then** the inspector classifies it as `GEO_RESTRICTED`.

---

### User Story 2 - State Persistence Across Page Refreshes (Priority: P1)
As a user monitoring a long proxy scan, I want to refresh the browser (F5) or re-open the tab without losing loaded proxies, scan progress, or probe results.

**Why this priority**: Prevents frustrating data loss during long scans involving hundreds of nodes.

**Acceptance Scenarios**:
1. **Given** 100 proxies loaded into the system, **When** the user refreshes the browser page, **Then** `web/app.js` issues `GET /api/session`, populates the matrix table with all previous reports, restores active badge counts, and re-attaches to the live SSE stream if a scan is running.

---

### User Story 3 - Real-Time Mid-Scan Subscription Export (Priority: P1)
As an operator, I want to click "Copy Clean Subscription" at any time during an active scan, so that I can immediately use verified nodes without waiting for all remaining nodes to finish.

**Why this priority**: In large batches (100+ proxies), waiting 5–10 minutes for slow nodes to complete blocks productive work.

**Acceptance Scenarios**:
1. **Given** a scan of 50 proxies where 5 have completed and passed, **When** the user clicks "Copy Clean Subscription", **Then** the 5 clean proxies are immediately formatted with smart tags (`[CC] [ISP] [PASS] [PINGms]`) and copied to the clipboard.

---

### User Story 4 - Fault-Tolerant Single-Node Pool Fallback (Priority: P2)
As a system operator, I want a single malformed proxy configuration (e.g. invalid Reality public key or dead server) not to crash or invalidate the entire batch of 20 concurrent proxies in Xray.

**Why this priority**: Prevents "batch poisoning" where one bad node causes 19 healthy nodes to have empty/untested cells (`○`).

**Acceptance Scenarios**:
1. **Given** a chunk of 20 proxies containing 1 malformed configuration, **When** `StartPool` fails to initialize the combined Xray configuration, **Then** the orchestrator automatically falls back to testing the remaining 19 configs individually, marking only the malformed configuration as `⚠️ Local Engine Port Bind Failed`.

---

### User Story 5 - Manual Configuration Clear Control (Priority: P2)
As a user, I want an explicit "Clear All Configs" button to purge all ingested data from memory and reset the dashboard.

**Acceptance Scenarios**:
1. **Given** active configurations in memory, **When** the user clicks "🗑️ Clear All Configs" and confirms, **Then** the system sends `POST /api/reset`, cancels any active scan, purges `sc.candidates` and `sc.reports`, and updates all connected dashboards to zero state.

---

## 3. Requirements

### Functional Requirements
- **FR-001**: System MUST query Google's first-party localized preferences endpoint (`https://www.google.com/preferences?hl=en`) through the proxy tunnel to extract Google's detected 2-letter country code (`lang="en-XX"` or `hl=en-XX`).
- **FR-002**: System MUST cross-reference the extracted Google country code against Google's official unserviced policy list (`IR`, `RU`, `BY`, `CN`, `KP`, `SY`, `CU`, `AF`) and flag matching nodes with `InterceptGeoRestricted`.
- **FR-003**: System MUST inspect target response bodies for multilingual geoblock tokens (including Persian phrases: `"در کشورتان دردسترس نیست"`, `"در کشور شما در دسترس نیست"`, `"جمینای در کشورتان دردسترس نیست"`, `"پشتیبانی نمی‌شود"`).
- **FR-004**: System MUST intercept redirect cascades targeting `/docs/available-regions` or `available-regions` in Google AI Studio and classify them as `InterceptGeoRestricted`.
- **FR-005**: System MUST expose `GET /api/session` returning all currently stored reports, total candidate count, and scanning state.
- **FR-006**: Frontend MUST invoke `GET /api/session` upon `DOMContentLoaded` to hydrate table rows and badge statistics.
- **FR-007**: System MUST safely update `ServerContext.reports` in real time on every `OnItemUpdate` callback invocation so that `GET /api/export` returns immediately valid payloads mid-scan.
- **FR-008**: System MUST expose `POST /api/reset` to terminate running scans, purge candidate/report maps, and broadcast reset statistics.
- **FR-009**: The orchestrator MUST implement single-node fallback isolation when multi-inbound pool startup fails, ensuring healthy nodes within the chunk are evaluated.
- **FR-010**: Loopback socket readiness polling (`WaitForPortsReady`) MUST use a timeout budget of at least 5000ms on Windows hosts.

---

## 4. Edge Cases & Constraints

- **Port Recycling Cache Collision**: Because local SOCKS5 ports (e.g. `127.0.0.1:20000`) are allocated, released, and re-allocated across subsequent batches, `geoCache` MUST be keyed by `config_id` (`ProxyConfig.ID`), NOT by `proxyEndpoint`.
- **Google Preferences Latency Spike**: `DetectGoogleRegion` MUST enforce a strict 4-second internal context timeout so an unresponsive Google connection cannot stall subsequent probe evaluations.
- **Incomplete Report Badge Rendering**: In the web UI, if an item reaches state `COMPLETE` but a probe result is missing, the cell MUST render `✗ N/A` instead of an ambiguous pending circle (`○`).
