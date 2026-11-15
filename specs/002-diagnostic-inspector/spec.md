# Feature Specification: Rich Diagnostic Logging & Interception Inspector

**Feature Branch**: `002-diagnostic-inspector`

**Created**: 2026-11-14

**Status**: Draft

**Input**: User description: "سیستم تشخیص و لاگینگ عمیق برای مشخص کردن دقیق دلیل فیل شدن تست‌ها: تفکیک خطای سوکت لوکال و Xray (مانند connectex)، ریدایرکت‌های کلادفلر (Under-Attack / Turnstile)، ریدایرکت‌های کپچای گوگل (/sorry/index برای ترافیک مشکوک)، ارورهای ریجن و ژئوبلاک واقعی، رفع باگ‌های فالزپازیتیو در متن پاسخ، و ذخیره‌سازی داده‌های تشخیصی برای UI."

---

## Clarifications

### Session 2026-11-14
- Q: How should the HTTP probe client handle redirects when evaluating target platforms (such as Google Gemini and AI Studio)? → A: Option A (Follow Redirects with Hop Snooping): The probe client follows up to 5 redirects, actively snooping on the `Location` header and HTTP status code of each intermediate hop. If any intermediate redirect points to an interception target (e.g., `google.com/sorry/index` or Cloudflare challenge domain), the inspector immediately flags the interception type without blindly falling into redirect loops.
- Q: How should the matrix table in the web dashboard present the granular diagnostic details when a user inspects a failed checkpoint? → A: Option A (Interactive Tooltip + Clickable Modal): In the dashboard matrix table, hovering over a failed service checkpoint displays an instant summary badge; clicking the badge opens a structured inspection modal displaying the Error Layer, Interception Type, Final URL, Status Code, Ray-ID, and page title snippet.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Granular Interception & Response Classification (Priority: P1)

As a user testing proxy pools against AI services, I want the system to analyze HTTP status codes, headers, redirect URLs, and HTML payloads to accurately identify the exact mechanism blocking access (Cloudflare security challenge, Google unusual traffic captcha, true regional geoblock, or upstream timeout), so that I understand why a proxy is flagged as non-clean rather than seeing a vague generic failure.

**Why this priority**: Essential diagnostic capability. Without granular response classification, users cannot determine whether an IP is challenged by Cloudflare, rate-limited by Google, or genuinely geoblocked.

**Independent Test**: Can be tested by feeding mock endpoints simulating: (a) a Cloudflare challenge page (`Server: cloudflare`, `cf-ray`, title "Just a moment..."), (b) a Google captcha page redirecting to `/sorry/index`, (c) an authentic Gemini login page containing OAuth scripts, and (d) a true geoblock page. The inspector must classify each into its exact category without false positives.

**Acceptance Scenarios**:
1. **Given** a proxy evaluation encounters an HTTP 403/503 response with Cloudflare headers and a Ray-ID, **When** inspection completes, **Then** the result flags `InterceptionType = CLOUDFLARE_CHALLENGE`, records the Ray-ID, and extracts the page title.
2. **Given** a proxy evaluation receives an HTTP 302 redirecting to `https://www.google.com/sorry/index?continue=...`, **When** inspection completes, **Then** the result flags `InterceptionType = GOOGLE_UNUSUAL_TRAFFIC` and captures the target captcha URL.
3. **Given** a valid, clean Google login page containing standard OAuth scripts with strings like `"access_denied"`, **When** inspection runs, **Then** the inspector recognizes the page as an authentic accessible sign-in and marks the probe as `PASS` (zero false positives).
4. **Given** a page returning a genuine geoblocking banner, **When** inspection runs, **Then** the result flags `InterceptionType = GEO_RESTRICTED`.

---

### User Story 2 - Local Engine Health & Socket Polling (Priority: P2)

As an operator running concurrent batch scans, I want the system to actively poll local tunnel sockets and capture engine standard error outputs before launching probes, so that startup delays and local port collisions (`connectex: actively refused it`) are eliminated and cleanly differentiated from remote network failures.

**Why this priority**: Prevents internal engine race conditions and port conflicts from contaminating proxy benchmark results.

**Independent Test**: Can be tested by intentionally delaying local port listener binding; the engine must poll until the socket accepts connections or times out, reporting `LOCAL_PORT_TIMEOUT` instead of crashing probes.

**Acceptance Scenarios**:
1. **Given** an Xray process is launched for a batch of proxies, **When** local SOCKS5 inbounds are binding, **Then** the system actively probes socket readiness on `127.0.0.1:port` at 20ms intervals until all ports are listening.
2. **Given** a local port fails to bind within the 1.5-second readiness window, **When** polling expires, **Then** the system flags `ErrorLayer = LOCAL_ENGINE` with an explicit diagnostic message (`SOCKS5 port failed to bind`) and aborts probe dispatch.
3. **Given** an engine subprocess writes warning or error lines to its error stream, **When** execution runs, **Then** the system buffers these diagnostic logs in memory for inspection.

---

### User Story 3 - Interactive Diagnostic Telemetry & Tooltip Inspection (Priority: P3)

As a dashboard user, I want to hover over or click any failed checkpoint badge in the matrix grid to inspect a detailed diagnostic summary (showing Error Layer, Interception Mechanism, Final URL, Server Header, Ray-ID, and Page Title), so that I can immediately debug issues directly from the browser.

**Why this priority**: Provides complete transparency to the user and powers future UI redesigns with rich metadata.

**Independent Test**: Can be tested by running a scan with mixed proxies and hovering over failed matrix badges; the UI displays a structured popover containing all inspection fields.

**Acceptance Scenarios**:
1. **Given** a failed service evaluation on the live dashboard, **When** the user hovers over or clicks the `✗ FAIL` badge, **Then** a diagnostic tooltip opens displaying: Error Layer, Category, Final URL, Status Code, and Ray-ID/Snippet.
2. **Given** session telemetry is exported as JSON, **When** the user inspects the file, **Then** every probe result includes a nested `diagnostic_info` object containing lossless inspection details.

---

### Edge Cases

- **Mixed Redirect Chains**: When a target URL redirects through multiple intermediate domains (e.g., Target -> OAuth -> Cloudflare Challenge), the inspector must record the final landing URL and the highest-severity interception indicator.
- **Empty or Truncated Error Responses**: When an upstream proxy connection resets immediately after sending HTTP headers without a body, the inspector must record `UPSTREAM_RESET` with the HTTP status rather than failing with an unhandled exception.
- **Malformed Cloudflare Headers**: If Cloudflare headers are present without a `cf-ray` header, the system must still categorize the response as `CLOUDFLARE_CHALLENGE` based on body signatures.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-017**: System MUST inspect target HTTP responses across status codes, response headers, landing URLs, and body content to classify outcomes into standardized interception categories:
  - `NONE` (Pass / Authentic response)
  - `CLOUDFLARE_CHALLENGE` (Turnstile, Under-Attack mode, Ray-ID present)
  - `GOOGLE_UNUSUAL_TRAFFIC` (Google /sorry/index captcha redirect)
  - `GEO_RESTRICTED` (Location not supported banner)
  - `UPSTREAM_TIMEOUT` (Remote proxy failed to complete handshake)
  - `LOCAL_ENGINE_ERROR` (Local SOCKS5 socket refused connection)
- **FR-018**: System MUST isolate and report errors across three distinct architectural layers (`ErrorLayer`):
  - `LOCAL_ENGINE`: Host OS socket binding, process crashes, port conflicts.
  - `UPSTREAM_PROXY`: Proxy node unreachable, handshake drop, TCP reset.
  - `TARGET_SERVICE`: Google/Cloudflare interception, rate limits, geoblocks.
- **FR-019**: System MUST verify local SOCKS5 inbound readiness via active socket polling (probing at 20ms intervals up to 1500ms) before dispatching any service probe requests.
- **FR-020**: System MUST eliminate false-positive geoblock errors by replacing naive substring searching with semantic validation that ignores OAuth script constants (e.g. `access_denied` inside script tags).
- **FR-021**: System MUST record diagnostic metadata for every probe evaluation, including:
  - Error Layer
  - Interception Category
  - Final Landing URL
  - HTTP Status Code
  - Server Header (e.g., `cloudflare`, `gws`)
  - Cloudflare Ray-ID (if present)
  - Page Title or Error Summary (up to 120 chars)
- **FR-022**: System MUST capture and buffer standard error logs from proxy core subprocesses to diagnose configuration syntax or port binding errors.
- **FR-023**: System MUST expose diagnostic metadata through the REST/SSE stream and render interactive diagnostic tooltips (summary on hover, full inspection modal on click) for failed checkpoints in the dashboard matrix table.
- **FR-024**: System MUST guarantee graceful termination of all subprocesses upon batch completion to eliminate zombie processes and port contention.

---

### Key Entities

- **DiagnosticInfo**: Value object containing `ErrorLayer`, `InterceptionType`, `FinalURL`, `HTTPStatus`, `ServerHeader`, `RayID`, `PageTitle`, and `Message`.
- **ErrorLayer**: Enumeration (`LOCAL_ENGINE`, `UPSTREAM_PROXY`, `TARGET_SERVICE`).
- **InterceptionType**: Enumeration (`NONE`, `CLOUDFLARE_CHALLENGE`, `GOOGLE_UNUSUAL_TRAFFIC`, `GEO_RESTRICTED`, `LOCAL_ENGINE_ERROR`, `UPSTREAM_TIMEOUT`, `UNKNOWN`).

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-007**: 100% of Cloudflare challenges, Google /sorry/index captchas, and regional geoblocks are classified into distinct interception types rather than generic failures.
- **SC-008**: Zero false-negative failures on valid Google responses caused by OAuth JavaScript token matches.
- **SC-009**: 100% of local socket binding failures are detected and reported within 500ms before probe dispatch.
- **SC-010**: All proxy failure reports include actionable diagnostic metadata (error layer, interception type, and final response details).

---

## Assumptions

- Target servers under test expose standard HTTP/2 and HTTP/1.1 response headers.
- Cloudflare security challenges include recognizable markers (`Server: cloudflare`, `cf-ray`, or standard challenge titles).
- Google unusual traffic blocks redirect to URLs matching `google.com/sorry/index`.
- The host OS allows local TCP loopback probing on `127.0.0.1` without administrator elevation.
