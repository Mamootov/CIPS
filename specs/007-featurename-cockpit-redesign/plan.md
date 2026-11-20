# Implementation Plan: 007 - Cockpit UI/UX Redesign & 4-View Observability Architecture

**Feature Branch**: `007-featurename-cockpit-redesign`  
**Status**: Implemented & Verified  

---

## 1. Technical Architecture & File Modifications

```
+------------------------------------------------------------------------+
|                        web/index.html                                  |
|  - Persistent Header with 4 Navigation Tabs                            |
|  - View 1: #view-dashboard (KPI Cards + Filter Dock + Matrix Table)    |
|  - View 2: #view-map (3D WebGL Canvas + Overlays + 2D Traffic Map)     |
|  - View 3: #view-probes (Probe Cards Grid + Builder Form)              |
|  - View 4: #view-subscriptions (Feed URL + Presets + QR Container)     |
|  - Floating Bottom Command Dock & Modals (Ingest, Settings, Inspector) |
+-----------------------------------+------------------------------------+
                                    |
+-----------------------------------v------------------------------------+
|                        web/style.css                                   |
|  - Dark Cockpit Palette (#02050a base, #080e1a cards, #10b981 emerald) |
|  - Sentry-style KPI Card Flexbox & Stacked Health Bars                 |
|  - Terralogix Pill Badges, Glassmorphism Overlays, Modal Code Blocks   |
|  - Floating Command Dock Styles (fixed bottom blur capsule)            |
|  - Concurrency & Timeout Sliders CSS Styling                           |
+-----------------------------------+------------------------------------+
                                    |
+-----------------------------------v------------------------------------+
|                        web/app.js                                      |
|  - Tab Router & View Switcher                                          |
|  - Three.js 3D Tactical Earth Globe Controller (rotation, zoom, stars) |
|  - GeodesicOrbitCurve Non-Penetrating Great-Circle Corridors           |
|  - Continuous High-Speed Pure White Photon Stream Animation (60 FPS)   |
|  - Auto-detected Host Origin Egress Sync (GET /api/origin)             |
|  - Universal 249-Country Flags & Dynamic Filter Pills                  |
|  - Automatic Speed Sorting for Available Proxies                       |
|  - Dynamic Probe Studio Timeouts & Available Nodes Gating              |
|  - 2D Vector Traffic Map SVG Controller (green/red traffic paths)      |
|  - Real-time KPI Metric Calculations (p50, p95, Average Speed, Yield)  |
|  - Named SSE Event Listeners ('item_update', 'stats')                  |
+-----------------------------------+------------------------------------+
                                    |
+-----------------------------------v------------------------------------+
|                        Go Backend Modules                              |
|  - internal/probe/ip_intel.go: DetectHostOrigin()                      |
|  - internal/server/handlers.go: GET /api/origin, concurrency support   |
|  - internal/pipeline/orchestrator.go: SetConcurrency, parallel fallback|
|  - internal/domain/models.go: Thread-safe Clone() deep copy            |
|  - internal/engine/xray.go: Scoped mutex, dynamic port readiness budget|
+------------------------------------------------------------------------+
```

---

## 2. Component Implementation Details

### 2.1 View 1: Dashboard Cockpit
- Persistent top header with 4 tabs and live engine indicator (`XRAY READY`).
- Sentry KPI metrics calculator:
  - `p50` and `p95` calculated from completed probe latency arrays.
  - Health bar segment widths calculated dynamically: `cleanCount / total`, `geoCount / total`, `rateCount / total`, `failCount / total`.
- Dynamic Filter Dock:
  - Dynamic country pills generated from evaluated reports.
  - Automatic download speed sorting when `Available` is active.

### 2.2 View 2: Fleet Globe 3D & Global Traffic Map
- Three.js renderer attached to `#globe-canvas` with `earth-tactical-dark.jpg` texture.
- Auto-detected host origin dynamically anchors source point on the globe and 2D map.
- Geodesic Great-Circle trajectory (`GeodesicOrbitCurve`) guarantees line stays above spherical terrain.
- Tiny pure-white photon beads animate continuously at 60 FPS.
- Three.js mesh/material disposal prevents orphan beads on server switch.
- 2D vector Traffic Map widget mirrors active node route in real time.

### 2.3 View 3: Probe Studio
- Evaluator cards for Gemini, AI Studio, ChatGPT, Claude, and Speedtest.
- Synchronized timeout displays matching active engine timeout (`8,000 ms`).
- Speedtest gating clearly set to `Available Nodes Only`.

### 2.4 View 4: Subscriptions Hub
- Renders local feed URL with 1-click clipboard copy.
- Generates dynamic QR code via embedded SVG QR pattern.
- Preset export buttons filter reports and copy to clipboard.
