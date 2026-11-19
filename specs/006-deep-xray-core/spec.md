# Feature Specification: 006 - Deep Xray Core Diagnostic Logging and Granular Error Diagnostics

**Feature Branch**: `006-deep-xray-core`  
**Created**: November 19, 2026  
**Status**: Specified & Ready for Review  
**Input**: User request to lower log level (more detailed logs) to diagnose the exact root causes behind "Port Refused" and local engine errors.

---

## 1. Executive Summary

In current scans, when a proxy configuration fails local Xray initialization or upstream handshake, the dashboard displays a generic **`⚠️ Port Refused`** badge. This occurs because Xray's loglevel is hardcoded to `"none"`, discarding all internal diagnostic stderr output from the core process.

Feature 006 lowers the Xray log level to **`"warning"`**, captures and parses stderr output in real time, and classifies generic socket failures into actionable diagnostic signatures (e.g. **`⚠️ Bad Reality Key`**, **`⚠️ Bad Cipher`**, **`⚠️ Remote Refused`**, **`⚠️ TLS Timeout`**, **`⚠️ WS Failed`**). Furthermore, it exposes the raw Xray log snippet directly within the click-to-inspect Diagnostic Modal.

---

## 2. User Scenarios & Testing

### User Story 1 - Granular Error Badges Instead of Generic "Port Refused" (Priority: P1)
As a proxy tester, I want the matrix grid badge to explain *why* the local engine or proxy failed (e.g. whether it was an invalid Reality key, bad UUID, or remote server refusal), rather than seeing a vague "Port Refused" message across different failure types.

**Why this priority**: Saves hours of debugging by instantly revealing whether the failure was caused by client-side configuration syntax (fixable in link) vs dead upstream server.

**Acceptance Scenarios**:
1. **Given** a VLESS Reality configuration with an invalid public key or short ID, **When** evaluated, **Then** Xray logs `failed to parse reality / invalid public key`, and the badge renders **`⚠️ Bad Reality Key`**.
2. **Given** a Shadowsocks configuration with an unsupported or deprecated cipher, **When** evaluated, **Then** Xray logs `unknown cipher`, and the badge renders **`⚠️ Bad Cipher`**.
3. **Given** a proxy whose remote server is reachable by TCP ping but actively drops the proxy protocol handshake, **When** evaluated, **Then** the badge renders **`⚠️ Handshake Failed`**.
4. **Given** a proxy with an invalid UUID format, **When** evaluated, **Then** the badge renders **`⚠️ Bad UUID`**.

---

### User Story 2 - Raw Core Log Excerpt in Diagnostic Modal (Priority: P1)
As an operator, when I click on any failed checkpoint badge, I want the Diagnostic Modal to display the exact log line produced by Xray for that specific proxy node.

**Why this priority**: Provides ground-truth technical proof from the core engine binary for advanced troubleshooting.

**Acceptance Scenarios**:
1. **Given** a failed proxy node, **When** the user clicks the failure badge, **Then** the Diagnostic Modal opens and populates a new field **`Engine Log / Stderr`** with the exact timestamped log message from Xray.

---

### User Story 3 - Tagged Inbound/Outbound Log Attribution (Priority: P2)
As a system evaluator running 20 concurrent proxies in a single Xray process, I want log messages to be accurately attributed to the specific proxy config that caused them, rather than bleeding across all 20 nodes.

**Acceptance Scenarios**:
1. **Given** 20 proxies running in a multi-inbound pool, **When** proxy on port 20005 triggers an error, **Then** Xray logs containing `inbound-20005` or the outbound tag are routed specifically to that proxy's `DiagnosticInfo`.

---

## 3. Requirements

### Functional Requirements
- **FR-001**: System MUST configure Xray core process with `"loglevel": "warning"` in `buildPoolConfig`.
- **FR-002**: System MUST capture Xray stderr in a thread-safe rolling buffer and index log lines by inbound port / outbound tag.
- **FR-003**: System MUST parse Xray stderr lines to detect signature error types:
  - Reality key/handshake failure -> `InterceptBadRealityKey` (`⚠️ Bad Reality Key`)
  - Shadowsocks cipher failure -> `InterceptBadCipher` (`⚠️ Bad Cipher`)
  - UUID parsing failure -> `InterceptBadUUID` (`⚠️ Bad UUID`)
  - Remote connection refused / reset -> `InterceptUpstreamRefused` (`⚠️ Remote Refused`)
  - TLS / WebSocket handshake failure -> `InterceptHandshakeFailed` (`⚠️ Handshake Failed`)
- **FR-004**: System MUST attach the exact relevant core stderr snippet into `DiagnosticInfo.EngineLog`.
- **FR-005**: Diagnostic Modal (`#diag-modal`) MUST render the `Engine Log` row with monospace styling when log data is present.
- **FR-006**: Matrix grid badges MUST render specific error labels for each distinct engine error subtype.
- **FR-007**: When `StartPool` fails during batch startup, the fallback loop MUST capture the single-instance Xray stderr and attribute it directly to that configuration.

---

## 4. Edge Cases & Constraints

- **High-Volume Log Flooding**: At `"warning"` level, healthy connections produce minimal logs. To prevent memory bloat during massive scans, each report MUST retain only the last 3 most recent relevant stderr lines.
- **Log Demuxing Across Shared Process**: When 20 inbounds share one process, Xray prefixes routing lines with `[inbound-PORT]`. The log parser MUST use regex `\[inbound-(\d+)\]` to demux logs to their respective reports.
