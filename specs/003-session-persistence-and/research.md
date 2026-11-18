# Research & Empirical Findings: 003 - Session Persistence & Accuracy

---

## 1. Google Gemini & AI Studio Regional Enforcement Mechanics

### 1.1 The SPA Initial Load Paradox
- When an HTTP client executes `GET https://gemini.google.com/`, Google's edge reverse proxy returns `HTTP 200 OK` with a 900 KB Single-Page Application (SPA) shell containing the Wiz / Angular JavaScript bundle, HTML skeleton, and CSS.
- **Critical Discovery**: The text *"Gemini isn't supported in your country yet"* is **not** present in the initial server-rendered HTML payload, regardless of whether the request originates from Iran, Russia, or the United States.
- The region check is executed downstream by client-side JavaScript or Google Account session evaluation.

### 1.2 Google Accounts Redirection Illusion
- When an unauthenticated client requests `GET https://gemini.google.com/app` or `GET https://aistudio.google.com/`, Google responds with `HTTP 302 Found` directing the browser to:
  `https://accounts.google.com/ServiceLogin?continue=...`
- The Google Accounts login page (`accounts.google.com`) returns `HTTP 200 OK` globally for virtually all networks worldwide.
- Probes that follow redirects and check solely for `HTTP 200` without evaluating destination URLs falsely flag every accessible proxy as a verified "PASS".

### 1.3 First-Party Google Edge Geolocation (`/preferences?hl=en`)
- Rather than relying on third-party GeoIP databases (which often mislabel Anycast or BGP-rerouted nodes), Google exposes its own localized country code in preferences metadata:
  ```html
  <html xmlns="http://www.w3.org/1999/xhtml" lang="en-DE">
  ```
- The ISO 3166-1 alpha-2 code in `lang="en-([A-Z]{2})"` directly reflects the geographic jurisdiction Google's edge infrastructure assigns to the connecting TCP socket.
- Querying this endpoint through the proxy tunnel provides zero-assumption ground truth: nodes recognized by Google as `IR`, `RU`, etc., are guaranteed to be blocked on Gemini, while nodes routed through Europe (e.g. `FI`, `NL`) are permitted.

---

## 2. Windows Loopback Readiness & Subprocess Latency

### 2.1 Process Startup Overhead
- Spawning `xray.exe run -c <config>` on Windows requires creating the process, loading dynamic libraries, parsing a multi-inbound JSON configuration, and binding 20 distinct ephemeral TCP ports on `127.0.0.1`.
- Under system load or antivirus file scanning, this process takes between **1.2s and 2.5s**.
- A readiness poller timeout of 1500ms caused frequent false timeouts, killing the core process and leaving the 20 candidate proxies unexamined.
- Increasing the budget to **5000ms** with 20ms polling intervals ensures 100% readiness without artificial delays.

### 2.2 Batch Failure Isolation
- In an Xray configuration with 20 outbounds, if a single outbound contains an unparseable UUID or rejected protocol parameter, the entire Xray process terminates immediately.
- Trapping pool startup errors and falling back to sequential single-node instantiation (`StartPool(ctx, []ProxyConfig{r.Config})`) isolates the faulty configuration and allows the remaining 19 healthy nodes to proceed.
