# Feature Specification: 005 - Dynamic Multi-Service Probe Catalog with ChatGPT, Claude, and Custom Probes

**Feature Branch**: `005-dynamic-service-probes`  
**Created**: November 19, 2026  
**Status**: Specified & Ready for Review  
**Input**: User requirement for customizable probe selection, dynamic matrix grid rendering, native ChatGPT/Claude support, and removal of Persian geoblock hardcoded tokens from the universal inspector.

---

## 1. Executive Summary

Feature 005 transforms CIPS from a hardcoded 2-service checker (Gemini + AI Studio) into a **dynamic, multi-target validation suite**. It enables users to:
1. **Dynamically select active validation targets** (e.g. only test `ChatGPT`, or test `Gemini` + `Claude`) from a catalog of built-in AI presets.
2. **Add custom endpoints** on the fly (e.g. `https://flow.ai` or custom APIs).
3. **Render a fully dynamic matrix table** where only the user's active choices appear as vertical column checkpoints, eliminating dead or irrelevant table columns.
4. **Clean and internationalize the inspection engine**: Per explicit architectural decision, remove Persian language tokens from `inspector.go` and rely strictly on universal English tokens, Chrome 124 `Accept-Language: en-US` headers, and protocol-level WAF/Cloudflare signals.
5. **Dynamically calculate AI-Clean qualification**: A node is certified as `AI Clean` if and only if at least one of the user's *actively selected* target services returns a verified pass.

---

## 2. User Scenarios & Testing

### User Story 1 - Dynamic Service Selection & Adaptive Matrix Table (Priority: P1)
As an operator focusing exclusively on ChatGPT, I want to select only ChatGPT in the configuration panel, so that the matrix grid only tests and displays ChatGPT without cluttering the screen with Gemini or AI Studio columns.

**Why this priority**: Eliminates visual clutter, cuts scan time in half, and allows operators to test proxies for specific single-purpose AI workflows.

**Acceptance Scenarios**:
1. **Given** the service catalog chips in the settings panel, **When** the user checks only `ChatGPT` and unchecks `Gemini` and `AI Studio`, **Then** the matrix table renders headers: `#`, `Protocol`, `Server & Port`, `Ping`, `Speed`, `ChatGPT`, `IP / ISP`, `Status`.
2. **Given** an active scan with only `ChatGPT` selected, **When** evaluated, **Then** the pipeline only executes the `chatgpt` probe, omitting Gemini and AI Studio requests completely.
3. **Given** a proxy where ChatGPT passes, **When** evaluated, **Then** the node is marked `🌟 AI Clean`, and subscription export formats the tag as `[ChatGPT]`.

---

### User Story 2 - Native ChatGPT Probe (`chatgpt.com`) (Priority: P1)
As an AI developer, I want CIPS to accurately detect if a proxy can open ChatGPT (`chatgpt.com`), catching Cloudflare Turnstile blocks and OpenAI geoblock warnings.

**Why this priority**: ChatGPT is the most widely used AI tool, protected by strict Cloudflare WAF, Arkose Labs, and regional restrictions.

**Acceptance Scenarios**:
1. **Given** a dirty or datacenter proxy flagged by OpenAI, **When** evaluated against `https://chatgpt.com/`, **Then** the probe intercepts Cloudflare 403/503 Turnstile challenge or OpenAI geoblock tokens (`not available in your country`) and marks the result as `FAIL (🛡️ Cloudflare or 🌍 Geoblock)`.
2. **Given** a clean residential or unflagged hosting proxy, **When** evaluated, **Then** `https://chatgpt.com/` returns HTTP 200 without challenges, marking the result as `✓ PASS`.

---

### User Story 3 - Native Anthropic Claude Probe (`claude.ai`) (Priority: P1)
As an engineer using Anthropic Claude, I want CIPS to test reachability to `claude.ai`, detecting Claude's regional block redirects.

**Why this priority**: Anthropic enforces strict regional fences, redirecting unsupported regions to `anthropic.com/supported-countries`.

**Acceptance Scenarios**:
1. **Given** a proxy whose IP is unsupported by Claude, **When** requesting `https://claude.ai/`, **Then** the client intercepts redirects targeting `supported-countries` or Cloudflare Turnstile challenges and marks the result as `FAIL`.
2. **Given** a clean proxy in a supported region (US, UK, EU), **When** evaluated, **Then** the probe completes with `✓ PASS`.

---

### User Story 4 - Universal English-Only Inspector Engine (Priority: P1)
As an open-source maintainer, I want the core inspection engine (`inspector.go`) to be completely free of regional or Persian language hardcoded strings, relying on international English tokens and protocol signals.

**Why this priority**: Our client sends `Accept-Language: en-US,en;q=0.9`, ensuring target services return English payloads. Removing local language strings cleans the codebase and eliminates false-positive substring collisions on custom sites.

**Acceptance Scenarios**:
1. **Given** `inspector.go`, **When** inspected, **Then** all Persian strings (`"در کشورتان دردسترس نیست"`, etc.) are removed from `geoblockTokens`.
2. **Given** an error response from Google, OpenAI, or Cloudflare, **When** evaluated, **Then** the probe relies on English tokens (`"isn't supported in your country"`, `"not available in your country"`, `"unsupported country"`) and WAF headers (`cf-ray`).

---

### User Story 5 - Custom Target URL Probe Builder (Priority: P2)
As a power user, I want to type any custom URL (e.g. `https://flow.ai/` or `https://groq.com/`) and test proxies against it.

**Acceptance Scenarios**:
1. **Given** the custom probe input, **When** the user enters `https://flow.ai/` and clicks `+ Add`, **Then** a new service chip is created, and the matrix table dynamically adds a `Flow` column.
2. **Given** an active scan, **When** the custom probe runs, **Then** it evaluates reachability with the Universal Inspector (checking HTTP 200 and WAF challenges).

---

## 3. Requirements

### Functional Requirements
- **FR-001**: System MUST provide a probe catalog containing built-in presets: `gemini`, `aistudio`, `chatgpt`, `claude`, and dynamic `custom`.
- **FR-002**: System MUST implement `ChatGPTProbe` testing `https://chatgpt.com/` with Chrome 124 TLS profile and hop inspection.
- **FR-003**: System MUST implement `ClaudeProbe` testing `https://claude.ai/` with redirect interception for `supported-countries`.
- **FR-004**: System MUST strip all non-English / Persian tokens from `inspector.go`, maintaining a clean international token list.
- **FR-005**: Scan start API (`POST /api/scan/start`) MUST accept `active_probes: []string` and `custom_probes: []ProbeDefinition`.
- **FR-006**: The orchestrator MUST configure its active probe suite dynamically per scan request, executing ONLY selected probes.
- **FR-007**: A proxy report MUST be marked `IsClean = true` IF AND ONLY IF at least one of the *actively selected* service probes returns `StatusPass`. Auxiliary probes (`ipleak`, `speedtest`) MUST NOT qualify a node.
- **FR-008**: Frontend matrix grid MUST dynamically generate `<th>` columns and `<td>` checkpoint cells based on currently active probes.
- **FR-009**: Frontend MUST provide service selector chips in the settings sidebar with toggle states.
- **FR-010**: Frontend MUST provide a custom URL modal/form allowing users to add custom target endpoints.
- **FR-011**: Subscription remarks MUST format passed service tags dynamically matching active surviving targets (e.g. `[ChatGPT]`, `[Gemini+Claude]`).

---

## 4. Edge Cases & Constraints

- **No Services Selected**: If a user unchecks all services, the scan start button MUST be disabled with a toast notice: *"Please select at least one validation target"*.
- **Cloudflare Turnstile JS Execution**: Probes cannot run full browser JavaScript; therefore, WAF challenges MUST be detected at the HTTP/header layer (`Server: cloudflare`, `cf-ray`, `challenges.cloudflare.com`) and classified definitively as `CLOUDFLARE_CHALLENGE`.
- **OpenAI Session Cookies**: Unauthenticated requests to `chatgpt.com` return HTTP 200 for clean IPs and challenge/block for flagged IPs.

