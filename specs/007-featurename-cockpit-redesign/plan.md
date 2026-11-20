# Implementation Plan: 007 - Cockpit UI/UX Redesign & 4-View Observability Architecture

**Feature Branch**: `007-featurename-cockpit-redesign`  
**Status**: Ready for Implementation  

---

## 1. Technical Architecture & File Modifications

```
┌────────────────────────────────────────────────────────────────────────┐
│                        web/index.html                                  │
│  - Persistent Header with 4 Navigation Tabs                            │
│  - View 1: #view-dashboard (KPI Cards + Filter Dock + Matrix Table)    │
│  - View 2: #view-map (3D WebGL Canvas + Overlays + 2D Traffic Map)     │
│  - View 3: #view-probes (Probe Cards Grid + Builder Form)              │
│  - View 4: #view-subscriptions (Feed URL + Presets + QR Container)     │
│  - Floating Bottom Command Dock & Modals (Ingest, Diag Inspector)      │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼────────────────────────────────────┐
│                        web/style.css                                   │
│  - Dark Cockpit Palette (#02050a base, #080e1a cards, #10b981 emerald) │
│  - Sentry-style KPI Card Flexbox & Stacked Health Bars                 │
│  - Terralogix Pill Badges, Glassmorphism Overlays, Modal Code Blocks   │
│  - Floating Command Dock Styles (fixed bottom blur capsule)            │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼────────────────────────────────────┐
│                        web/app.js                                      │
│  - Tab Router & View Switcher                                          │
│  - Three.js 3D Earth Globe Controller (rotation, zoom, starfield)      │
│  - Raycasting Node Click ➔ Spectrum Arc & Inspector Card Sync          │
│  - Constellation City Jitter Algorithm                                 │
│  - 2D Vector Traffic Map SVG Controller (green/red traffic paths)      │
│  - Real-time KPI Metric Calculations (p50, p95, Average Speed, Yield)  │
│  - SSE Live Event Hydration & Floating Dock State Transitions          │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Component Implementation Details

### 2.1 View 1: Dashboard Cockpit
- Relocate sidebar elements into `#ingest-modal`.
- Top Header contains only branding, 4 tabs, and live status badge.
- Implement Sentry KPI metrics calculator:
  - `p50` and `p95` calculated from completed probe latency arrays.
  - Health bar segment widths calculated dynamically: `cleanCount / total`, `geoCount / total`, `rateCount / total`, `failCount / total`.
- Filter dock with live search (`Ctrl+K` handler).

### 2.2 View 2: Fleet Globe 3D & Global Traffic Map
- Three.js renderer attached to `#globe-canvas`.
- Earth sphere loaded with `web/assets/earth-tactical-dark.jpg` base64 data texture.
- Node pins rendered at city lat/lon with constellation jittering.
- Clicking any node:
  - Generates a single quadratic Bezier curve from Tehran to the selected node.
  - Applies 5-stop vertex color gradient (Emerald ➔ Cyan ➔ Purple ➔ Pink ➔ Rose).
  - Animates 4 data packet meshes along the curve.
  - Populates `#node-inspector-card`.
- 2D vector Traffic Map widget in `#traffic-map-card` renders green vs red lines based on backend report pass/fail state.

### 2.3 View 3: Probe Studio
- Renders cards for Gemini, AI Studio, ChatGPT, Claude, and Speedtest.
- Custom probe form directly triggers `POST /api/scan/start` payload customization.

### 2.4 View 4: Subscriptions Hub
- Renders local feed URL with 1-click clipboard copy.
- Generates dynamic QR code via embedded SVG QR pattern.
- Preset export buttons filter reports and copy to clipboard.
