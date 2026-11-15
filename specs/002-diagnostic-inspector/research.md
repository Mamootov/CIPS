# Phase 0 Research: Diagnostic Logging & Interception Inspector

## Objective
Establish concrete technical patterns to diagnose proxy test failures accurately, distinguishing between host socket delays, upstream proxy drops, Cloudflare challenges, Google unusual traffic captchas, and true regional geoblocks.

---

### Research Item 1: Cloudflare Security Interception Detection

- **Decision**: Multi-vector inspection examining `Server` header, `cf-ray` header, HTTP status (403/503), and HTML title/script markers.
- **Rationale**:
  1. Cloudflare protected sites return HTTP 403 or 503 accompanied by headers: `Server: cloudflare`, `cf-ray: <id>`, and `cf-cache-status`.
  2. The page title typically contains `"Just a moment..."` or `"Attention Required! | Cloudflare"`.
  3. Detecting these markers immediately classifies the block as `CLOUDFLARE_CHALLENGE` and captures the Ray-ID for operator inspection.
- **Alternatives Considered**:
  - *Status code check only*: Fails because a 403 could also be an application-level permission error or regional geoblock.

---

### Research Item 2: Google Unusual Traffic Captcha Detection

- **Decision**: Hop-by-hop redirect snooping and URL inspection for `google.com/sorry/index`.
- **Rationale**:
  1. When Google flags an IP for automated queries, it issues an HTTP 302 redirecting to `https://www.google.com/sorry/index?continue=...`.
  2. By inspecting the `Location` header on each redirect hop, CIPS detects the interception immediately without following the redirect into a captcha submission loop.
  3. Classified unambiguously as `GOOGLE_UNUSUAL_TRAFFIC`.
- **Alternatives Considered**:
  - *Blindly following all redirects*: Loads the heavy reCAPTCHA script page, wasting bandwidth and triggering confusing body token checks.

---

### Research Item 3: Elimination of False-Positive Geoblocks (OAuth Constant Bug)

- **Decision**: Replace naive body substring search with structural HTML tag filtering and targeted error banners.
- **Rationale**:
  1. Standard Google OAuth login pages include `access_denied` as a standard error code constant inside JavaScript bundles (`signin.js`). Naive substring checking flags legitimate clean login pages as blocked.
  2. True geoblocks display explicit user-facing warning banners: `"isn't supported in your country"`, `"location_error"`, or `"unsupported country"` inside visible body headings/paragraphs, not inside script tags.
  3. CIPS strips `<script>...</script>` blocks before scanning for geoblock tokens.
- **Alternatives Considered**:
  - *Relying only on HTTP 200 vs 403*: Google returns HTTP 200 for its `"not supported in your country"` notice page, so body inspection remains mandatory.

---

### Research Item 4: Active Socket Readiness Polling vs. Blind Sleep

- **Decision**: Active loopback connection polling (`net.DialTimeout("tcp", "127.0.0.1:port", 25ms)`) with 1500ms ceiling.
- **Rationale**:
  1. Replacing `time.Sleep(100ms)` with an active polling loop ensures probes fire immediately when all ports in a pool are listening, avoiding both unnecessary latency and premature connection attempts (`connectex: actively refused it`).
  2. If any port in the pool fails to bind within 1500ms, the engine fails fast with an explicit `LOCAL_ENGINE` error.
- **Alternatives Considered**:
  - *Increasing sleep to 1000ms*: Adds 1 full second of dead latency to every batch, reducing throughput by 80%.

---

### Research Item 5: Xray Process Lifecycle & Stderr Diagnostics

- **Decision**: Attach an in-memory rolling buffer to `cmd.Stderr` and enforce explicit `cmd.Process.Kill()` upon pool termination.
- **Rationale**:
  1. When an Xray process is started for a batch, closing the pool must forcefully terminate the subprocess and unbind ports before those ports are recycled.
  2. Capturing `Stderr` into a circular buffer allows CIPS to log exact Xray runtime warnings (e.g., port binding conflicts or TLS parsing errors) directly to the diagnostic stream.
- **Alternatives Considered**:
  - *Letting subprocesses linger until parent exit*: Leaks tens of processes, exhausts Windows socket handles, and triggers port bind errors on subsequent runs.
