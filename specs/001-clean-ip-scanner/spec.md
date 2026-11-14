# Feature Specification: Clean IP Scanner (CIPS)

**Feature Branch**: `001-clean-ip-scanner`

**Created**: 2026-11-13

**Status**: Draft

**Input**: User description: "از ipleak استفاده کنه و شرکت و ایپی کانفیگ رو داشته باشه. UI/UX تمیز و زیبا داشته باشه(ظاهر حرفه ای داشته باشه). حتما ایرور هندلینگ قوی داشته باشه تا در صورت کرش یا از بین رفتن یه پروسس بتونه همچنان کار کنه و برنامه کرش نکنه یا باعث کرش سیستم نشه. برای گرفتن ورودی ها هم بشه فایل بهش داد(درگ اند دراپ) فرمت جیسون یا md یا نوت. خروجی نهایی که کاربر ببینه یه صفحه باشه که بصورت لایو پایپ لاین و کارکرد و پاسخ هر کانفیگ رو ببینه، در نهایت هم یه فایل لاگ درست کنه و اینارو توش بریزه و خروجی تمیز و کانفیگ های قابل استفاده رو در همون بخش به همراه ایپی و شرکت و پینگشون بنویسه و بگه روی چه سرویس هایی پاسخ دادن. کانفیگ ها هر نوعی میتونن باشن vless, vmess, shadowsocks, socks5, https, ... اول که قرار شد یه پینگ از کانفیگ ها بگیره، بعدش کانفیگ هایی که پینگ دادن رو روی سایت هایی که کاربر به عنوان پروب بتونه اضافه یا کم کنه تست کنه."

---

## Clarifications

### Session 2026-11-13
- Q: How should the system determine whether an evaluated proxy qualifies as "Clean / Usable" for inclusion in the final exported subscription? → A: Option B with Checkpoint Matrix Table: The system enforces a configurable qualification policy (user can require any AI pass or specific targets). The live dashboard MUST render a structured matrix table where service probe names (Gemini, AI Studio, ipleak, custom probes) form dedicated vertical/column headers, with distinct visual checkpoint status indicators (e.g., checkmarks, cross marks, latency badges) per proxy row, rather than relying solely on text tags.
- Q: How should the application serve and launch the live web dashboard during execution? → A: Option A with Non-Standard Port: The application serves the embedded dashboard on an uncommon high-range default port (19090, avoiding clashes with 8080/3000) with dynamic fallback to an adjacent available port if occupied. It automatically opens the user's default browser on launch and supports a `--headless` / `--no-browser` CLI flag for headless server execution.
- Q: How should the system handle the acquisition and location of the underlying proxy core binary (e.g., xray.exe / sing-box)? → A: Option A (Smart Auto-Detect with Auto-Download): The application inspects the executable directory and system PATH first. If the required core binary is missing, it automatically downloads and unpacks the matching verified release binary (with mirror fallback) while displaying download progress in the terminal and web dashboard.
- Q: Should the ingestion engine support fetching configurations directly from remote subscription URLs (HTTP/HTTPS) in addition to local file drag-and-drop and clipboard pasting? → A: Option A (Unified Ingestion with Remote Subscriptions): The ingestion engine supports direct HTTP/HTTPS subscription URLs (automatically retrieving and decoding Base64 or plain-text proxy feeds like standard V2Ray clients), alongside local drag-and-drop files (JSON, Markdown, TXT) and clipboard paste.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Multi-Format Ingestion & Two-Stage Parallel Scanning (Priority: P1)

As a user with a bulk list of diverse proxy configurations (VLESS, VMess, Trojan, Shadowsocks, SOCKS5, HTTP/HTTPS), I want to drag and drop my configuration files (JSON, Markdown, plain text notes) into the tool, so that the system immediately initiates a two-stage evaluation pipeline: first eliminating unreachable nodes through an initial latency ping, and then validating active nodes against target services.

**Why this priority**: Core value proposition. Without resilient batch ingestion and the two-stage pipeline, the system cannot process real-world proxy pools or protect downstream probes from dead-link timeouts.

**Independent Test**: Can be tested independently by submitting a mixed text file containing valid, invalid, and unreachable proxy links. The system must parse all recognized protocol links, discard malformed lines with informative warnings, filter out unresponsive nodes during stage 1 (ping), and output the ping status of responsive nodes.

**Acceptance Scenarios**:
1. **Given** a user has a file containing mixed protocols (`vless://`, `vmess://`, `ss://`, `trojan://`, plain text notes), **When** the user drags and drops the file onto the ingestion area, **Then** the system extracts all valid proxy configurations, deduplicates them, and displays the total count of parsed candidates.
2. **Given** 100 parsed proxy configurations where 40 are offline, **When** the scan starts, **Then** Stage 1 executes concurrent connectivity handshakes, drops the 40 unresponsive nodes within the configured timeout window, and advances only the 60 responsive nodes to Stage 2.
3. **Given** an invalid or truncated configuration string, **When** ingestion runs, **Then** the parser safely flags the error with line context without crashing the intake process.

---

### User Story 2 - Real-Time Pipeline Dashboard & Live Telemetry (Priority: P2)

As a network operator, I want to watch a clean, professional, live web dashboard displaying the pipeline progression of every configuration in real time, so that I can monitor active tests, latency benchmarks, and granular service responses as they occur.

**Why this priority**: Users require real-time visibility into long-running batch operations rather than waiting blindly for a terminal script to finish.

**Independent Test**: Can be tested by running a batch scan with simulated network delays and observing that the dashboard renders state transitions (Queued → Handshake → Testing Probes → Enriched → Complete) live without UI stuttering or requiring manual browser refreshes.

**Acceptance Scenarios**:
1. **Given** an active batch scan in progress, **When** individual configurations pass or fail Stage 1 and Stage 2, **Then** the dashboard updates their state cards and latency metrics immediately via live data streaming.
2. **Given** a configuration completing Stage 2, **When** probe evaluation finishes, **Then** the dashboard displays each service as a dedicated vertical column checkpoint in the main matrix table, showing an unambiguous visual status indicator (e.g., green checkmark for pass, red cross for fail, pending spinner) and response latency for each service.
3. **Given** an active scan running, **When** the user toggles view filters (e.g., "Show Only Clean AI-Ready", "Show Failed"), **Then** the dashboard updates the list dynamically without interrupting background workers.

---

### User Story 3 - Service Probe Customization & IP Enrichment via ipleak (Priority: P3)

As a power user, I want to dynamically configure which service probes are active (adding custom endpoint checks, enabling/disabling Gemini or Google AI Studio) and enrich surviving proxies with exit IP, hosting company, and geographical data from ipleak, so that I understand the exact operational profile of each proxy.

**Why this priority**: Different workflows require validating different platforms; IP metadata provides transparency regarding the infrastructure backing each viable proxy.

**Independent Test**: Can be tested by enabling a custom probe and the ipleak enricher, running a verified live proxy, and checking that the resulting telemetry card contains ISP name, ASN, country, and probe HTTP response details.

**Acceptance Scenarios**:
1. **Given** a user opens probe settings, **When** the user adds a new target URL with expected response conditions, **Then** subsequent scans include that target in Stage 2 evaluation alongside default AI probes.
2. **Given** a proxy that successfully passes Stage 2 probes, **When** the enrichment phase executes, **Then** the system queries the IP leak intelligence service through that specific proxy tunnel and attaches exit IP, ISP name, and hosting provider to its record.
3. **Given** the ipleak query times out or returns malformed data, **When** probe evaluation was successful, **Then** the proxy is still marked as functional, with IP metadata flagged as "Enrichment Unavailable" rather than failing the proxy.

---

### User Story 4 - Resilient Fault Isolation, Structured Logs & Clean Export (Priority: P4)

As a user, I want the system to isolate proxy crashes, network disconnects, or hanging sockets so that scanning continues uninterrupted, while automatically writing structured session logs and producing a curated, ready-to-use subscription payload of clean configurations.

**Why this priority**: High-concurrency network scanners frequently crash or hang when dealing with adversarial or broken proxy endpoints; robust isolation and actionable export deliver enterprise reliability.

**Independent Test**: Can be tested by injecting a proxy designed to hang indefinitely and one designed to violently reset connections mid-handshake; the scanner must enforce strict timeout ceilings, log the failure vectors to disk, and export clean proxies without system instability.

**Acceptance Scenarios**:
1. **Given** a proxy configuration that accepts connections but never returns data, **When** Stage 1 or Stage 2 reaches the per-node timeout limit, **Then** the worker forcefully closes the connection, records a timeout diagnostic, and proceeds to the next item.
2. **Given** a scan session concludes, **When** the user inspects the output directory, **Then** a structured session log file (JSON format) and a clean subscription export (annotated with ISP, IP, and pass tags) are persisted.
3. **Given** the user views the "Usable Configs" section on the dashboard, **When** the user clicks "Copy Clean Subscription", **Then** the annotated usable configuration URLs are copied to the clipboard in standard importable format.

---

### Edge Cases

- **Corrupted Input Files**: Uploading non-text binaries or deeply nested invalid JSON files must trigger a graceful user error banner without process termination.
- **Extreme Latency / Stalled Handshakes**: Upstream proxies that accept TCP SYN but freeze on TLS ClientHello must be killed after a strictly enforced timeout (default: 5 seconds for Stage 1, 10 seconds for Stage 2).
- **Rate-Limiting on Metadata Endpoints**: If `ipleak.net` applies rate-limiting across concurrent queries, the enricher must utilize jittered dispatch and fall back gracefully without invalidating successful service probes.
- **DNS Poisoning & Hijacked Upstreams**: If an upstream proxy redirects probe requests to a local captive portal or fake HTTP 200 page, response content assertion (checking for expected application tokens, not just HTTP status) must flag the anomaly as a probe failure.
- **Sudden Process Interruption**: When a user halts the scanner mid-run, all active network sockets, temporary tunnel listeners, and worker routines must be cleanly released within 2 seconds.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support drag-and-drop file ingestion, manual clipboard pasting, and remote HTTP/HTTPS subscription URL intake (with automatic Base64/plain payload decoding) for JSON, Markdown (`.md`), and plain text notes.
- **FR-002**: Ingestion parser MUST extract proxy protocol URIs matching `vless://`, `vmess://`, `trojan://`, `ss://` (Shadowsocks), `socks5://`, and `http://`/`https://`.
- **FR-003**: System MUST automatically deduplicate incoming proxy configurations based on target server address, port, and security credentials before pipeline queuing.
- **FR-004**: System MUST execute a two-stage testing pipeline:
  - **Stage 1 (Ping & Reachability)**: Rapid TCP/TLS handshake latency measurement; unresponsive nodes are discarded.
  - **Stage 2 (Service Probing & Enrichment)**: Evaluation of responsive nodes against active target probes and IP intelligence endpoints.
- **FR-005**: System MUST provide pre-configured, bot-resilient probes for:
  - Google Gemini (`gemini.google.com`)
  - Google AI Studio (`aistudio.google.com`)
- **FR-006**: System MUST allow users to add, modify, or disable target service probes with custom target URLs and expected response criteria (status code, body patterns).
- **FR-007**: System MUST query an IP intelligence service (`ipleak.net` JSON endpoint) through the verified proxy tunnel to obtain:
  - Exit IP Address
  - Internet Service Provider (ISP) / Hosting Organization
  - Country / Region Code
- **FR-008**: System MUST serve a local web dashboard presenting a modern, professional user interface featuring:
  - Default binding to an uncommon high-range localhost port (`127.0.0.1:19090`) with automatic increment fallback if occupied, auto-opening the user's default browser on launch, with a `--headless` / `--no-browser` override
  - Live progress counters (Total Ingested, Reachable, AI-Ready, Failed)
  - Interactive pipeline matrix table streaming real-time status updates per configuration, where each active service probe (e.g., Gemini, Google AI Studio, ipleak, and user-defined probes) has a dedicated vertical column header with visual checkpoint status indicators (pass/fail/latency) per row
  - Configurable concurrency, timeout, and qualification export filters
  - Dedicated "Clean & Usable" export drawer allowing selective export based on service checkpoint criteria
- **FR-009**: System MUST enforce an isolated lifecycle for every proxy evaluation: socket timeouts, process isolation, and automated garbage collection of tunnel resources.
- **FR-010**: System MUST NOT crash or hang when individual proxies disconnect abruptly, produce malformed HTTP responses, or stall indefinitely.
- **FR-011**: System MUST persist a timestamped session log file containing comprehensive diagnostic telemetry for all evaluated configurations.
- **FR-012**: System MUST generate an annotated output subscription list containing only clean, verified configurations matching user-selected checkpoint criteria, with remarks enriched with ISP, country, and verified service indicators (e.g., `[US-AI-Clean] [Timeweb] MyProxy`).
- **FR-013**: System MUST permit one-click copying and direct file download of the clean subscription output.
- **FR-014**: System MUST execute all upstream probe requests using authentic modern browser TLS client hellos (JA3/JA4 fingerprinting) to prevent false-negative bot mitigations.
- **FR-015**: System MUST provide a responsive pause/cancel control that immediately halts pipeline execution and releases all system network resources.
- **FR-016**: System MUST automatically detect the presence of the required core binary (`xray` / `xray.exe`) in the application directory or system PATH, and if absent, automatically download and unpack the verified release binary for the host platform with progress feedback.

---

### Key Entities

- **ProxyConfiguration**: Represents a parsed proxy endpoint containing raw URI, protocol type (VLESS, VMess, etc.), destination server, port, credentials, and original remark.
- **ProbeDefinition**: Represents a target evaluation endpoint containing name, URL, timeout, expected HTTP status, required response body regex/tokens, and active status.
- **ProbeResult**: Represents the outcome of an individual probe evaluation against a proxy, including probe name, target status (PASS/FAIL/SKIP), HTTP response code, response latency, and failure reason.
- **IPEnrichmentData**: Represents metadata fetched from the exit IP intelligence service, including public IP, ISP/Organization, ASN, and country code.
- **EvaluationReport**: Aggregates all evaluation results for a single proxy: initial ping latency, collection of probe results, IP enrichment data, overall usability verdict, and enriched export URI.
- **ScanSession**: Encapsulates a batch execution run, tracking start time, total ingested items, progress state, concurrency settings, active worker pool, and final output telemetry file paths.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can ingest a batch of 500+ mixed proxy configurations via drag-and-drop within 3 seconds.
- **SC-002**: Stage 1 reachability triage processes at least 50 proxy configurations per second under standard broadband conditions.
- **SC-003**: 100% of unresponsive or hanging proxies are terminated within the user-configured timeout window without causing worker deadlocks or memory leaks.
- **SC-004**: Zero false-negative bot blocks on Google Gemini and Google AI Studio caused by synthetic/scripted TLS fingerprints on verified clean IPs.
- **SC-005**: The live web dashboard reflects state changes and latency metrics within 200 milliseconds of test completion.
- **SC-006**: Generated clean subscription output imports seamlessly into industry-standard proxy client applications (V2RayN, Sing-box, Clash, v2rayNG) without syntax errors.

---

## Assumptions

- Target operating environment is Windows 10/11 or modern Linux/macOS with standard local networking privileges (ability to bind ephemeral localhost ports).
- The underlying proxy core (e.g., Xray or Sing-box) is packaged or accessible locally to handle low-level protocol multiplexing and cryptographic handshakes.
- The user operates on a network capable of establishing outbound connections to public Internet proxy destinations (subject to local ISP firewall conditions).
- External intelligence services (`ipleak.net`) remain accessible over HTTP/HTTPS when accessed through functional proxy egress tunnels.
- The local dashboard binds to `127.0.0.1` on an available port (e.g., 8080 or auto-assigned) for secure local single-user access.
