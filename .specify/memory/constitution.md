<!--
Sync Impact Report:
- Version change: initial -> 1.0.0
- Added sections:
  * Core Principles (I. Zero-Assumption Empirical Ground Truth, II. Core-Agnostic Hexagonal Architecture, III. Authentic TLS & Protocol Fingerprinting, IV. Resilient Concurrency & Resource Isolation, V. Pluggable Extensibility & Lossless Telemetry)
  * Technology Stack & Operational Constraints
  * Development Workflow & Quality Gates
  * Governance & Amendments
- Deferred Items: Implementation stack setup (Go toolchain installation/verification) deferred to planning/implementation phases.
-->

# CIPS (Clean IP Scanner) Constitution

## Core Principles

### I. Zero-Assumption Empirical Ground Truth
- Pre-filtering proxy configurations based on static GeoIP, ASN registries, or country blocklists is strictly prohibited.
- Proxy validity and service accessibility MUST be evaluated exclusively through live, empirical requests against upstream target endpoints (e.g., Google Gemini, Google AI Studio, Cloudflare).
- Third-party metadata endpoints (such as `ipleak.net`) MUST serve solely as data enrichment and tagging sinks in the final telemetry, never as pre-flight gatekeepers.
- **Rationale**: Anycast routing, BGP traffic engineering, and non-geographic IP reputation policies create significant divergence between static GeoIP databases and real-world service accessibility. Pre-filtering introduces unacceptable false negatives.

### II. Core-Agnostic Hexagonal Architecture
- The core business domain and probe evaluation pipelines MUST remain completely decoupled from underlying proxy execution mechanisms via an explicit `ProxyEngine` port.
- Proxy engine implementations (e.g., Xray subprocess, embedded Sing-box core, or local SOCKS5 tunnel pools) MUST be interchangeable without altering probe logic or orchestrator workflows.
- Probe components MUST interact exclusively with standardized local endpoints (SOCKS5/HTTP) and never consume or manipulate raw protocol-specific link schemas.
- **Rationale**: Evolving censorship bypass protocols (VLESS, Reality, Trojan, Hysteria2) demand modular engine upgrades without risking destabilization of evaluation and reporting modules.

### III. Authentic TLS & Protocol Fingerprinting
- All network probes evaluating bot-sensitive platforms MUST employ full TLS client impersonation (matching modern Chrome/Firefox JA3/JA4 fingerprints) alongside authentic HTTP/2 frame sequences and headers.
- Standard HTTP clients lacking low-level TLS signature control (e.g., vanilla `net/http` without custom Hello extensions) MUST NOT be utilized for platform health evaluation.
- **Rationale**: Advanced AI platforms and CDNs enforce sophisticated automated traffic mitigation; naive TLS signatures trigger immediate HTTP 403 or challenge redirects, skewing benchmark accuracy.

### IV. Resilient Concurrency & Resource Isolation
- High-throughput scanning MUST be strictly constrained by configurable bounded worker pools and concurrency semaphores to eliminate socket exhaustion and packet starvation.
- Every ephemeral port, inbound listener, temporary configuration artifact, and child process MUST be registered within a deterministic lifecycle manager that guarantees termination upon completion, error, timeout, or external interrupt (SIGINT/SIGTERM). Leaked sockets or orphan processes constitute critical failures.
- **Rationale**: Mass parallel probing across hundreds of proxy instances creates intense operating system I/O pressure; strict lifecycle boundaries are required for enterprise-grade system stability.

### V. Pluggable Extensibility & Lossless Telemetry
- Verification targets MUST be implemented as independent, self-contained `Probe` plugins adhering to a uniform interface contract (`IProbe`). Registering additional target services (e.g., Claude, OpenAI, media streaming) MUST require zero alterations to orchestration logic.
- Probe outputs MUST preserve lossless structured telemetry (latency, HTTP status, exit IP, ASN, ISP, failure vectors) exportable to structured JSON, while concurrently generating clean, filtered, smart-tagged subscription payloads.
- **Rationale**: Enterprise maintainability requires an open-closed probe ecosystem and rich telemetry that can feed future routing engines or anomaly-detection classifiers.

## Technology Stack & Operational Constraints

- **Language & Runtime**: Go (Golang) 1.22+ for high-performance concurrent networking, low memory footprint, and standalone cross-platform compilation.
- **Architectural Style**: Hexagonal Architecture (Ports and Adapters) with clear domain boundaries separating ingestion, proxy management, probing, and export sinks.
- **TLS & Transport Layer**: Modern Go TLS impersonation engine (e.g., `bogdanfinn/tls-client` or equivalent low-level JA3/JA4 engine) for realistic browser simulation.
- **CLI & User Interface**: Structured CLI via `spf13/cobra`, with interactive terminal progress and telemetry reporting powered by `pterm` or `charmbracelet/bubbletea`.
- **Packaging & Portability**: Deliverable as a single static self-contained binary without external runtime dependencies.

## Development Workflow & Quality Gates

- **Spec-Driven Development (SDD)**: All development MUST proceed through defined phases: Constitution -> Specify -> Clarify -> Plan -> Tasks -> Implement -> Converge.
- **Zero-Premature-Code Gate**: No production source code, mock routes, or functional prototypes shall be written until the technical specification (`spec.md`) and implementation plan (`plan.md`) are ratified.
- **Strict Version Control Safety**: Automated commits and pushes are forbidden; every `git commit` and `git push` requires explicit interactive human confirmation.
- **Deterministic Testing**: Protocol parsers and telemetry adapters MUST include automated unit tests. Engine adapters MUST be validated against mock proxy listeners prior to integration.

## Governance

- This Constitution constitutes the supreme architectural policy of the CIPS project and supersedes transient convenience or informal coding practices.
- Amendments require a documented pull request, justification of architectural impact, and an explicit version increment following Semantic Versioning (MAJOR for principle changes/removals, MINOR for additions, PATCH for refinements).
- All implementation plans, specifications, and code reviews MUST be validated against the five core principles prior to acceptance.

**Version**: 1.0.0 | **Ratified**: 2026-11-13 | **Last Amended**: 2026-11-13
