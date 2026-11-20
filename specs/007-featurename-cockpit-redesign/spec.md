# Feature Specification: 007 - Cockpit UI/UX Redesign & 4-View Observability Architecture

**Feature Branch**: `007-featurename-cockpit-redesign`  
**Created**: November 20, 2026  
**Status**: Specified & Ready for Implementation  

---

## 1. Executive Summary & Design Vision

This specification defines the complete UI/UX overhaul of CIPS (Clean IP Scanner), transitioning from a basic sidebar-and-table form into an enterprise-grade observability cockpit inspired by **Sentry** (statistical health & latency distribution), **Modal** (technical data density & developer ergonomics), **Terralogix** (interactive 3D globe & transit corridor vectors), and **Cybersecurity Command Centers** (live 2D global traffic maps with country borders).

The redesigned frontend decomposes the application into **4 focused views** with a persistent top navigation bar and an unobtrusive **Floating Bottom Command Dock**, completely removing header clutter while delivering progressive disclosure.

---

## 2. Page & View Specifications

```
┌────────────────────────────────────────────────────────────────────────────────────────┐
│ [⚡ CIPS]   [ 📊 Dashboard ]   [ 🗺️ Fleet Globe 3D ]   [ 🧪 Probe Studio ]   [ 📦 Subscriptions ] │
└────────────────────────────────────────────────────────────────────────────────────────┘
```

### 2.1 View 1: 📊 Dashboard Cockpit (`#view-dashboard`)
1. **Sentry-Style KPI Summary Cards (4 Cards)**:
   - **Card 1: AI Clean Yield**: Large display count `Passed / Total` (e.g. `42 / 128`), with percentage `▲ 32.8% Unrestricted AI Access`.
   - **Card 2: Average Anycast Speed**: Live average speed in Mbps (e.g. `46.5 Mbps`), peak speed tracker.
   - **Card 3: Latency Distribution**: Percentile metrics (`p50: 148ms`, `p95: 390ms`, Jitter).
   - **Card 4: Health Distribution Bar**: Multi-segment stacked horizontal bar visualizing proportions of Clean (Green), Geoblocked (Cyan), Rate Limited (Orange), and Failed (Red) nodes.
2. **Terralogix-Style Filter Dock**:
   - Quick-toggle pills: `All Nodes`, `🌟 AI Clean`, `🌍 Geoblocked`, `⚠️ Core Error`.
   - Country pills: `🇩🇪 Germany`, `🇳🇱 Netherlands`, `🇺🇸 USA`, `🇫🇮 Finland`, etc.
   - Classification pill: `🏠 Residential Only`.
   - Search input with keyboard shortcut hint (`⌘K` / `Ctrl+K`).
3. **Modal-Style Technical Matrix Grid**:
   - Columns: `#`, `Node & Transit Corridor`, `Protocol`, `Exit IP & ISP`, `[Dynamic AI Columns]`, `Speed`, `Status`.
   - Visual route corridor vector: `🇮🇷 Tehrān ──────▶ 🇩🇪 Nuremberg`.
   - Protocol pills: `VLESS`, `VMess`, `Trojan`, `Shadowsocks`.
   - Granular status badges: `200 OK (latency)`, `🌍 Geoblock`, `⚠️ Bad Reality`, `⚠️ Bad Cipher`.
4. **Floating Bottom Command Dock**:
   - Centered frosted glass floating dock:
     - `[ 📥 Ingest Configs ]` (opens modal)
     - `[ 🚀 Start Audit · N Nodes ]` (initiates live scan)
     - `[ 📋 Copy Clean (N) ]` (1-click clipboard export)

---

### 2.2 View 2: 🗺️ Fleet Globe 3D & Global Traffic Map (`#view-map`)
1. **Interactive 3D WebGL Globe**:
   - Realistic satellite Earth sphere using dark-mode terrain imagery with dimmed, non-glaring deserts.
   - Full 360° mouse drag rotation and scroll zoom via Three.js OrbitControls.
   - Deep space backdrop with subtle starfield particles and cyan atmospheric Fresnel limb glow.
2. **City Coordinate Mapping & Constellation Jittering**:
   - Node coordinates placed at exact geographic city locations (Tehran, Frankfurt, Nuremberg, Amsterdam, Helsinki, New York, etc.).
   - Multi-node city hubs (e.g. 4 servers in Frankfurt) rendered with orbital jitter so each server forms a distinct point in a city cluster without overlapping.
   - Origin hub (Tehran, Iran) highlighted with a prominent emerald beacon and concentric pulsing ring.
3. **Single-Corridor Glowing Spectrum Arc**:
   - In resting state, nodes appear as delicate, tiny pins.
   - Upon clicking any node: **ONLY ONE dynamic glowing 3D Bezier arc** ignites between Tehran and that specific target city.
   - The arc transitions across a continuous spectrum palette: Emerald (`#10b981`) ➔ Cyan (`#38bdf8`) ➔ Cyber Purple (`#a855f7`) ➔ Neon Pink (`#ec4899`).
   - Luminous data packet particles animate and travel along the arc, visualizing live throughput.
4. **Floating Left Inspector Card**:
   - Displays complete telemetry for the active node: remark, transit corridor, city & coordinates, cluster jitter rank, protocol, exit IP, ISP/ASN, residential classification, live AI latencies, speed, and 1-click copy button.
5. **Floating Right-Top Cluster Card**:
   - Displays all server nodes belonging to the selected city hub with individual IPs and speeds.
6. **Floating Right-Bottom 2D Global Traffic Map**:
   - Dedicated HUD card containing an authentic 2D vector world map with exact borders for 177 countries.
   - Iran highlighted with emerald glow.
   - Transcontinental traffic lines color-coded:
     - **Green Arcs (`#10b981`)**: Active clean proxy corridors (e.g. to Germany, Netherlands, USA).
     - **Red Arcs (`#f43f5e`)**: Geoblocked or failed connection paths (e.g. to Russia, blocked regions).
   - Live throughput telemetry label: `⚡ 94.2 Mbps Throughput · 128 Tunnels Active`.

---

### 2.3 View 3: 🧪 AI Probe Studio (`#view-probes`)
1. **Interactive Evaluator Cards**:
   - Grid cards for Google Gemini, Google AI Studio, OpenAI ChatGPT, Anthropic Claude, and Low-Bandwidth Speedtest.
   - Each card features:
     - On/Off toggle switch.
     - Target domain and inspection profile.
     - Real-time pass count and average latency.
2. **Custom Probe Builder Card**:
   - Form allowing users to input probe name, target URL, and expected status code/body keyword.
   - Adds dynamically to the active probe catalog and matrix grid without restarting the server.

---

### 2.4 View 4: 📦 Subscriptions & Mobile Hub (`#view-subscriptions`)
1. **Live Subscription Endpoint**:
   - Displays local auto-updating URL: `http://127.0.0.1:19090/api/export?format=sub&filter=clean`.
   - Copy button and interval notification for desktop/mobile clients.
2. **Preset Export Filters**:
   - Quick export buttons: `All AI Clean (N)`, `Top 10 Fastest (>40Mbps)`, `Residential Only`, `Europe Only`.
3. **Smart Tagged Remark Preview**:
   - Live visual box demonstrating the injected tags:
     `[DE] [Hetzner] [RES] [44Mbps] [145ms] [ChatGPT+Gemini] - Reality-Tunnel (3)`.
4. **Mobile QR Code Card**:
   - Generates scannable QR code for instant import into v2rayNG, Streisand, Shadowrocket, and Clash Verge.

---

## 3. Backend-to-Frontend Module Architecture Mapping

| Backend Module / API | Output Data Model | Connected Frontend View & Component |
| :--- | :--- | :--- |
| **`GET /api/session`** | `EvaluationStateResponse` (reports, scan_status, config_count) | **Dashboard**: Matrix table rows, KPI metric calculations (p50/p95, yield rate, speed), Health stacked bar. |
| **`GET /events` (SSE)** | `PipelineEvent` (`ITEM_UPDATE`, `SCAN_COMPLETE`, etc.) | **Live Telemetry**: Real-time row status updates, progress bar in dock, dynamic KPI counter increments. |
| **`domain.IPEnrichment`** | `ExitIP`, `CountryCode`, `City`, `ISP`, `ConnectionType` | **Fleet Globe 3D**: City coordinate mapping, constellation jittering, Left Inspector Card, and 2D Traffic Map status lines. |
| **`internal/probe/factory.go`** | `ProbeSuite`, `AvailableProbes` | **Probe Studio**: Service card active switches, timeout settings, and custom probe catalog. |
| **`GET /api/export`** | Clean base64 subscription / raw text / JSON | **Subscriptions Hub**: 1-click copy feed, format buttons, and preset filtered exports. |
| **`POST /api/scan/start`** | Execution params (`concurrency`, `services`, `speedtest`) | **Floating Command Dock**: Triggered by `[ 🚀 Start Audit ]`. |
| **`POST /api/ingest`** | Parse result, config count | **Ingest Modal**: Triggered by `[ 📥 Ingest Configs ]`. |

---

## 4. User Stories & Acceptance Criteria

### User Story 1 - Seamless Cockpit Navigation (Priority: P1)
**Given** the user opens `http://127.0.0.1:19090`,  
**When** the page loads,  
**Then** the Dashboard view is displayed by default with KPI cards, filter pills, the matrix table, and the floating command dock at the bottom, without any bulky action buttons in the header.

### User Story 2 - Interactive 3D Fleet Globe & Real 2D Traffic Map (Priority: P1)
**Given** configurations have been ingested and scanned,  
**When** the user clicks the `Fleet Globe 3D` tab,  
**Then** a rotatable Three.js satellite Earth globe is rendered with dimmed deserts, small sharp node pins, and an interactive 2D vector Global Traffic Map with country borders and green/red traffic lines in the bottom-right. Clicking a node on the globe illuminates a multi-color spectrum arc from Tehran to that city and updates the Left Inspector Card with full telemetry.

### User Story 3 - Probe Studio Configuration (Priority: P2)
**Given** the user wants to test only specific AI services,  
**When** the user navigates to `Probe Studio` and toggles off Gemini,  
**Then** subsequent scans exclude Gemini, and the matrix grid dynamically adapts its columns.

### User Story 4 - Subscriptions & QR Mobile Sync (Priority: P2)
**Given** a completed or in-progress scan,  
**When** the user opens the `Subscriptions` tab,  
**Then** the live subscription feed URL, preset export options, and mobile QR code are immediately available.
