# Requirements Quality Checklist: 007 - Cockpit UI/UX Redesign

- [x] **Clutter-Free Ergonomics**: Action buttons removed from the top header; relocated to an unobtrusive Floating Bottom Command Dock.
- [x] **Progressive Disclosure**: Application decomposed into 4 specialized views (`Dashboard`, `Fleet Globe 3D`, `Probe Studio`, `Subscriptions`).
- [x] **Observability Parity**: Sentry-style KPI cards (AI Clean yield, speed, p50/p95 latency, multi-segment health bar) integrated into View 1.
- [x] **BASTION 3D Earth Quality**: 3D globe rendered with Three.js using dimmed, non-glaring desert satellite terrain and atmosphere limb.
- [x] **Focused Spectrum Corridor**: Clicking any node illuminates ONLY one multi-color gradient arc with moving packet particles from source to target.
- [x] **Constellation City Jittering**: Nodes in the same geographic hub (e.g. Frankfurt) are visually jittered into distinct cluster positions.
- [x] **Real 2D Global Traffic Map**: Bottom-right mini widget features authentic vector country borders for 177 countries and green vs red traffic lines.
- [x] **Non-Negotiable**: No commit without user approval; no git push under any circumstances.
