# Requirements Quality Checklist: 005 - Dynamic Service Probes

- [x] **Zero-Assumption Empirical Ground Truth**: All targets (built-in or custom) are evaluated through real live HTTP/TLS handshakes, never by static IP lookups.
- [x] **No Premature Code Gates**: Specification and requirements ratified before implementation.
- [x] **Universal English Tokens**: Core inspector stripped of localized Persian strings, relying strictly on universal WAF and English tokens matching Chrome 124 headers.
- [x] **Dynamic Table Architecture**: Matrix grid adapts its header columns and cells to match exactly what the user selected.
- [x] **Clean Gating**: A node is marked AI-Clean if and only if at least one selected target service passed.
- [x] **Extensibility**: Users can add arbitrary URLs with customizable assertions.
