# Feature Specification: 004 - IP Intelligence and Low-Bandwidth Chunked Speedtest

**Feature Branch**: `004-ip-intelligence-and`  
**Created**: November 18, 2026  
**Status**: Specified & Ready for Review  
**Input**: Comprehensive research on `ipwho.is` IP enrichment and Cloudflare CDN low-bandwidth streaming speedtest.

---

## 1. Executive Summary

Feature 004 expands CIPS from a pure reachability and geoblock validator into a full-spectrum performance and threat-intelligence suite. It introduces:
1. **Deep IP Intelligence**: Extracts exit IP, ISP, ASN, organization, country, and critically **`connection.type`** (distinguishing **`Residential`** from **`Data Center`**), with a multi-tier fallback architecture (`ipwho.is` -> `ipleak.net` -> Cloudflare trace).
2. **Low-Bandwidth Chunked Speedtest**: Measures real-world download throughput in **Mbps** using Cloudflare Anycast edge streaming (`speed.cloudflare.com/__down?bytes=N`), constrained by a strict 3-second hard timeout to conserve mobile and metered bandwidth.
3. **Interactive Control & Enriched Export**: Provides a dashboard toggle to enable/disable speed testing, a dedicated `Speed` column in the matrix grid with performance tier badges, and enhanced subscription remarks carrying ISP, connection type, and speed ratings.

---

## 2. User Scenarios & Testing

### User Story 1 - Deep IP Intelligence & Connection Classification (Priority: P1)
As an operator screening proxies for AI models, I want to see whether a node's IP is residential or datacenter, its ISP, and its ASN, so that I can choose high-reputation residential or premium hosting IPs that avoid Cloudflare and Google fraud blocks.

**Why this priority**: IP reputation and connection type directly correlate with long-term survival against Google and Cloudflare bot mitigations.

**Acceptance Scenarios**:
1. **Given** an active proxy tunnel, **When** the IP enrichment probe executes, **Then** it queries `https://ipwho.is/` through the proxy, extracting IP, country, city, ISP, ASN, and `connection.type`.
2. **Given** an IP classified with `connection.type == "Data Center"`, **When** rendered in the dashboard, **Then** the UI displays an `🏢 DC` badge alongside the ISP name.
3. **Given** an IP classified with `connection.type == "Residential"`, **When** rendered in the dashboard, **Then** the UI displays an `🏠 Res` badge.
4. **Given** `ipwho.is` fails or times out, **When** evaluated, **Then** the system automatically falls back to `ipleak.net` and `cloudflare trace` to ensure IP and ISP metadata are never lost.

---

### User Story 2 - Real-World Low-Bandwidth Speed Testing for AI-Verified Nodes (Priority: P1)
As a proxy user, I want to measure the download throughput ONLY for configurations that successfully unlocked at least one AI service (Gemini or AI Studio), so that I do not waste bandwidth testing broken or geoblocked proxies.

**Why this priority**: Saves massive amounts of cellular/metered bandwidth by filtering out proxies that are useless for AI work before running download tests.

**Acceptance Scenarios**:
1. **Given** a proxy where both Gemini and AI Studio failed, **When** evaluated, **Then** the speed test is automatically SKIPPED, consuming 0 bytes of download bandwidth.
2. **Given** a proxy where Gemini OR AI Studio passed, **When** evaluated, **Then** the client requests `https://speed.cloudflare.com/__down?bytes=5000000` with a strict 3000ms timeout budget.
3. **Given** a 3-second download duration, **When** the timer expires or download finishes, **Then** the system calculates speed in Mbps ($(\text{Bytes} \times 8) / (\text{Seconds} \times 1,000,000)$) and terminates the TCP connection immediately.
4. **Given** a proxy that transfers fewer than 20 KB within 3 seconds, **Then** the system classifies speed as `0.0 Mbps` or `Slow` without hanging.

---

### User Story 3 - User Control over Speed Testing (Priority: P2)
As a user with a limited cellular data plan, I want a toggle switch in the dashboard to turn speed testing on or off and configure the maximum download ceiling.

**Acceptance Scenarios**:
1. **Given** the pipeline configuration panel, **When** the user toggles "Enable Speed Test" OFF, **Then** Stage 2 skips the speed test probe, completing scans faster and consuming zero download data.
2. **Given** the speed test payload slider, **When** the user selects 1MB, 3MB, or 5MB, **Then** subsequent scans adjust the test payload size accordingly.

---

### User Story 4 - Visual Matrix Grid with Dedicated Speed Metrics (Priority: P2)
As a user monitoring the dashboard, I want a dedicated `Speed` column in the matrix table with color-coded performance badges.

**Acceptance Scenarios**:
1. **Given** speed >= 10.0 Mbps, **When** rendered, **Then** badge displays `🚀 {speed} Mbps` in bright green.
2. **Given** speed between 2.0 and 9.9 Mbps, **When** rendered, **Then** badge displays `⚡ {speed} Mbps` in cyan.
3. **Given** speed < 2.0 Mbps, **When** rendered, **Then** badge displays `🐌 {speed} Mbps` in warning orange.

---

### User Story 5 - Enriched Subscription Export Format (Priority: P3)
As an operator exporting clean nodes, I want the generated subscription remarks to embed ISP, connection type, speed, and latency tags.

**Acceptance Scenarios**:
1. **Given** a verified clean proxy, **When** exported to subscription, **Then** the remark is formatted as:
   `[US] [Hetzner] [DC] [18.4Mbps] [145ms] [Gemini+Studio] - OriginalRemark`

---

## 3. Requirements

### Functional Requirements
- **FR-001**: System MUST implement an IP intelligence probe querying `https://ipwho.is/` through the proxy tunnel with JSON response parsing.
- **FR-002**: System MUST parse `connection.type` (`Data Center`, `Residential`, `Business`, `Wireless`) from `ipwho.is` into `domain.IPEnrichment`.
- **FR-003**: System MUST provide automatic fallback to `https://ipleak.net/json/` and `https://speed.cloudflare.com/cdn-cgi/trace` if `ipwho.is` is unreachable.
- **FR-004**: System MUST implement a dedicated speed test probe querying `https://speed.cloudflare.com/__down?bytes={size}`.
- **FR-005**: Speed test probe MUST enforce a hard context timeout of 3000ms and close the underlying response body immediately upon deadline expiration.
- **FR-006**: System MUST calculate download speed in Megabits per second (Mbps) using exact elapsed microsecond precision.
- **FR-007**: Scan start API (`POST /api/scan/start`) MUST accept `enable_speedtest: bool` and `speedtest_bytes: int` parameters.
- **FR-008**: The orchestrator MUST conditionally trigger the speed test probe strictly AFTER AI service evaluation, and ONLY for configurations where at least one AI service probe (Gemini or AI Studio) has achieved a `PASS` status. If no AI probe passed, the speed test MUST be skipped, conserving 100% of download data.
- **FR-009**: Dashboard UI MUST provide a toggle switch `Enable Speed Test` and payload size selector in the settings sidebar.
- **FR-010**: Dashboard matrix table MUST include a dedicated `Speed` column rendering color-coded badges (`🚀`, `⚡`, `🐌`, `○`).
- **FR-011**: Dashboard matrix table MUST render connection type badges (`🏠 Res`, `🏢 DC`) in the `IP / ISP` column.
- **FR-012**: Subscription exporter (`exporter.EnrichProxyRemark`) MUST embed connection type (`[Res]`/`[DC]`) and speed (`[{speed}Mbps]`) tags when available.

---

## 4. Edge Cases & Constraints

- **Proxy Egress Rate Limiting on IPWho.is**: If hundreds of proxies share the same datacenter subnet, `ipwho.is` might return HTTP 429. The probe MUST detect status 429 and seamlessly fall back to `ipleak.net` without failing the probe.
- **Slow / Throttled Proxies during Speedtest**: If a proxy hangs or downloads at dial-up speeds (<5 KB/s), the 3-second hard timeout MUST terminate the connection immediately so the scanner does not stall.
- **Bandwidth Conservation Ceiling**: The maximum download payload per proxy MUST NOT exceed 5 Megabytes under any circumstance.
