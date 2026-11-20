# Requirements Quality Checklist: 007 - Cockpit UI/UX Redesign

- [x] **Clutter-Free Ergonomics**: Action buttons removed from the top header; relocated to an unobtrusive Floating Bottom Command Dock with `20x` concurrency modal.
- [x] **Progressive Disclosure**: Application decomposed into 4 specialized views (`Dashboard`, `Fleet Globe 3D`, `Probe Studio`, `Subscriptions`).
- [x] **Observability Parity**: Sentry-style KPI cards (Available proxies yield, speed, p50/p95 latency, multi-segment health bar) integrated into View 1.
- [x] **Non-Penetrating 3D Geodesic Corridors**: Spherical great-circle orbital curve (`GeodesicOrbitCurve`) strictly vaults above Earth terrain across all 360 degrees.
- [x] **Sleek Pure-White Photon Animation**: Moving data pulses rendered as crisp, tiny white beads (`#ffffff`) without oversaturated halos.
- [x] **Dynamic Host Egress Origin**: Automatic discovery of host public IP, country, and city replacing hardcoded Tehran.
- [x] **Dynamic Filters & Universal Flags**: Universal Unicode emoji flag generation for all 249 countries and data-driven filter pills without hardcoded countries.
- [x] **Speed Sorting**: Available nodes automatically sorted descending by download speed.
- [x] **Dynamic Probe Studio**: Timeouts synchronized with engine settings and gating explicitly set to `Available Nodes Only`.
- [x] **High-Concurrency Pipeline**: Scalable up to 100 parallel workers with thread-safe `Clone()` snapshots and non-blocking Xray engine startups.
- [x] **Non-Negotiable**: No commit without user approval; no git push under any circumstances.
