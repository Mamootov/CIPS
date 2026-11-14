# Phase 0 Research: Clean IP Scanner (CIPS)

## Objective
Establish concrete, production-grade technical decisions for the CIPS engine in Go, ensuring compliance with the CIPS Constitution (Zero-Assumption Empirical Validation, Core-Agnostic Hexagonal Architecture, Authentic TLS Impersonation, Resilient Resource Lifecycle, and Pluggable Extensibility).

---

### Research Item 1: Proxy Engine Integration Pattern

- **Decision**: Multi-Port Inbound Adapter over Managed Subprocess with Hot JSON Routing (`engine/xray/`).
- **Rationale**:
  1. Running hundreds of separate `xray.exe` instances wastes gigabytes of memory and risks Windows file descriptor exhaustion.
  2. Generating a unified runtime configuration with a pool of local SOCKS5 inbound listeners (`127.0.0.1:20000` through `127.0.0.1:20050`) mapped 1:1 to distinct outbound proxy definitions via tag-based routing allows a single high-performance Xray core instance to multiplex up to 50 concurrent proxy tunnels with negligible CPU/RAM overhead (<30MB RAM).
  3. Batching proxy chunks through this rotating inbound pool ensures deterministic port allocation and instant lifecycle teardown upon scan completion or cancellation.
- **Alternatives Considered**:
  - *Embedded Go Xray/Sing-box libraries*: Extremely difficult dependency hell due to incompatible Go module versions between Xray/Sing-box and custom TLS clients.
  - *One-Process-Per-Proxy Subprocess*: Massive memory footprint and heavy process creation overhead on Windows (100–300ms per spawn).

---

### Research Item 2: Authentic TLS Fingerprint Impersonation in Go

- **Decision**: `github.com/bogdanfinn/tls-client` (Chrome 124/130 profile with HTTP/2 transport).
- **Rationale**:
  1. Powered by `utls` and a modified `fhttp` stack, `bogdanfinn/tls-client` accurately emulates modern Chrome TLS ClientHello extensions, elliptic curves, cipher suites, signature algorithms, and ALPN.
  2. Standard Go `net/http` advertises static Go TLS fingerprints that Google Gemini, Google AI Studio, and Cloudflare flag immediately with HTTP 403 or captcha redirects.
  3. `tls-client` natively supports routing requests through local SOCKS5/HTTP proxies with per-request timeout enforcement.
- **Alternatives Considered**:
  - *Vanilla `net/http` with custom TLS config*: Lacks GREASE, custom extension ordering, and HTTP/2 SETTINGS frame customization; easily fingerprinted and blocked by Google bot mitigation.
  - *`cgo` bindings to `curl-impersonate`*: Cross-compilation across Windows/Linux becomes brittle and requires external C runtimes, violating single-binary portability.

---

### Research Item 3: Embedded Web Dashboard & Live Event Streaming

- **Decision**: Go `embed.FS` serving a Single-Page Application (HTML5 + Tailwind CSS + Vanilla JS) over Server-Sent Events (SSE) & REST endpoints.
- **Rationale**:
  1. SSE provides lightweight, unidirectional real-time streaming of state updates (ideal for streaming live pipeline matrix updates to the browser) with automatic reconnection and zero handshake overhead compared to full duplex WebSockets.
  2. Embedding static assets directly into the Go binary (`go:embed web/*`) guarantees zero runtime dependencies, eliminating separate node server installations or asset missing errors.
  3. Drag-and-drop file ingestion, subscription fetching, probe customization, and subscription copying can be executed directly via simple REST calls (`/api/ingest`, `/api/scan/start`, `/api/probes`, `/api/export`).
- **Alternatives Considered**:
  - *Wails / WebView2*: Requires platform-specific webview runtime installations and large compile binaries; does not support headless server execution.
  - *Terminal UI only (Bubbletea)*: Fails the user requirement for drag-and-drop file ingestion and graphical matrix view.

---

### Research Item 4: IP Intelligence & Metadata Enrichment (`ipleak.net`)

- **Decision**: Dedicated `IPEnricher` probe querying `https://ipleak.net/json/` through the tested proxy tunnel with exponential jitter and fallback parsing.
- **Rationale**:
  1. `https://ipleak.net/json/` provides clean JSON output containing `ip`, `country_name`, `country_code`, `isp_name`, and `as_number`.
  2. To avoid rate-limiting triggers during high-concurrency batch execution, queries to ipleak are dispatched with 150–300ms random jitter only for proxies that successfully pass Stage 1 and at least one Stage 2 service probe.
  3. If ipleak fails or times out, the proxy is still marked as usable; metadata fields default gracefully to `"Unknown ISP" / "Unenriched"` without failing the valid proxy.
- **Alternatives Considered**:
  - *Local MaxMind GeoIP Database*: Fast, but static; cannot provide real-time exit IP verification or current ISP reputation.
  - *`ipinfo.io` / `ip-api.com`*: Restrictive free-tier rate limits (45 req/min on ip-api, strict token needed on ipinfo).

---

### Research Item 5: Resilient Concurrency & Resource Lifecycle Management

- **Decision**: Bounded Worker Pool with Dynamic Ephemeral Port Acquisition and `context.Context` cancellation trees.
- **Rationale**:
  1. A semaphore bounded to `worker_count` (default: 20, configurable up to 100) strictly meters concurrent network I/O.
  2. Local inbound ports are tested for availability (`net.ListenTCP("tcp", &net.TCPAddr{Port: 0})`) before binding to avoid collisions with existing host services.
  3. Each test run executes inside a root `context.WithCancel`; a system signal listener (`SIGINT`, `SIGTERM`, Windows `os.Interrupt`) triggers immediate context cancellation, closing all open sockets, killing child core processes via process tree termination, and releasing all locked ports in <500ms.
- **Alternatives Considered**:
  - *Unbounded Goroutines*: Causes packet drops, DNS resolution timeouts, and ephemeral port exhaustion on Windows when handling lists with 500+ items.
