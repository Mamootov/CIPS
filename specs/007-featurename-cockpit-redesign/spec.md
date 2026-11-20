# Feature Specification: 007 - Cockpit UI/UX Redesign & 4-View Observability Architecture

**Feature Identifier**: `007-cockpit-redesign`  
**Created**: November 20, 2026  
**Last Updated**: November 20, 2026  
**Status**: Implemented & Verified  

---

## 1. Executive Summary & Design Vision

This specification defines the complete UI/UX overhaul of CIPS (Clean IP Scanner), transitioning from a basic form into an enterprise-grade observability cockpit inspired by **Sentry** (statistical health & latency distribution), **Modal** (technical data density & developer ergonomics), **Terralogix** (interactive 3D globe & transit corridor vectors), and **Cybersecurity Command Centers** (live 2D global traffic maps with country borders).

The frontend organizes the application into **4 focused views** with a persistent top navigation bar and an unobtrusive **Floating Bottom Command Dock**, completely removing header clutter while delivering progressive disclosure.

---

## 2. Page & View Specifications

```
+----------------------------------------------------------------------------------------+
| [CIPS v1.2]   [ Dashboard ]   [ Fleet Globe 3D ]   [ Probe Studio ]   [ Subscriptions ] |
+----------------------------------------------------------------------------------------+
```

### 2.1 View 1: Dashboard Cockpit (`#view-dashboard`)
1. **Sentry-Style KPI Summary Cards (4 Cards)**:
   - **Card 1: Available Proxies**: Large display count `Passed / Total` (e.g. `1 / 2`), with percentage `▲ 50.0% Unrestricted AI Access`.
   - **Card 2: Average Anycast Speed**: Live average speed in Mbps (e.g. `48.4 Mbps`), peak speed tracker.
   - **Card 3: Latency Distribution**: Percentile metrics (`p50`, `p95`, Jitter).
   - **Card 4: Health Distribution Bar**: Multi-segment stacked horizontal bar visualizing proportions of Available (Green), Geoblocked (Cyan), Rate Limited (Orange), and Failed (Red) nodes.
2. **Terralogix-Style Dynamic Filter Dock**:
   - Quick-toggle pills: `All Nodes`, `Available`, `Geoblocked`, `Core Error`, `Residential Only`.
   - **Dynamic Country Pills**: 100% data-driven; generated dynamically based on actual tested exit countries (e.g. `Germany (38)`, `Netherlands (14)`). Zero hardcoded country buttons.
   - **Automatic Speed Sorting**: When the `Available` filter is selected, candidate nodes are automatically sorted in descending order by download speed (`download_speed_mbps`).
   - Quick search input (`#node-search-input`) with global keyboard shortcut (`Cmd+K` / `Ctrl+K`).
3. **Modal-Style Technical Matrix Grid**:
   - Columns: `#`, `Node & Transit Corridor`, `Protocol`, `Exit IP & ISP`, `[Dynamic AI Columns]`, `Speed`, `Status`.
   - Dynamic Transit Corridor: `${origin.flag} ${origin.city} -> ${dest.flag} ${dest.city}`.
   - Protocol tags: `VLESS`, `VMess`, `Trojan`, `Shadowsocks`.
   - Granular status badges: `200 OK (latency)`, `Geoblock`, `Bad Reality`, `Block / Err`, `Dead (Ping Fail)`.
4. **Floating Bottom Command Dock**:
   - Frosted glass floating dock:
     - `[ Ingest Configs ]` (opens modal)
     - `[ 20x ]` (opens Pipeline Concurrency & Engine Settings modal)
     - `[ Start Live Audit · N Nodes ]` (initiates live scan)
     - `[ Copy Available (N) ]` (1-click clipboard export)
     - `[ Trash ]` (clears loaded configs and reports)

---

### 2.2 View 2: Fleet Globe 3D & Global Traffic Map (`#view-map`)
1. **Interactive 3D WebGL Globe**:
   - Realistic satellite Earth sphere using dark-mode terrain imagery (`earth-tactical-dark.jpg`) with dimmed, non-glaring deserts.
   - Full 360 rotation and scroll zoom via Three.js OrbitControls.
   - Deep space backdrop with subtle starfield particles and cyan atmospheric limb glow.
2. **Dynamic Host Egress Origin Discovery**:
   - Origin point on Earth is auto-discovered via `GET /api/origin` (`DetectHostOrigin`).
   - Displays real public IP, Country Code, City, and geographic coordinates of the machine running CIPS (e.g. `Nuremberg` when hosted on VPS, `Tehran` when running in Iran).
3. **Geodesic Orbit Great-Circle Trajectory (`GeodesicOrbitCurve`)**:
   - Computes rotational axis perpendicular to the plane of origin and destination vectors: `N = normalize(v1 x v2)`.
   - Evaluates trajectory along the spherical Great Circle with altitude envelope `R(t) = r + sin(t*pi) * maxAlt >= r`.
   - **Guaranteed non-penetration**: The trajectory strictly glides above the surface of the globe across all 360 degrees and never cuts through the Earth's interior, even for antipodal or opposite-hemisphere destinations (e.g., Australia, New Zealand, Canada).
4. **Continuous High-Speed Laser Light Flow**:
   - 8 tiny, pure-white photon beads (`SphereGeometry(0.045)`, `#ffffff`) streaming along the fiber-optic line.
   - Clean aesthetics: No oversized or oversaturated colored halos.
   - Strict resource lifecycle management: Disposes of Three.js geometries/materials on node switch, preventing orphan or frozen beads from previous servers.
5. **Floating Left Inspector Card**:
   - Displays complete telemetry for the active node: remark, transit corridor, city & coordinates, cluster jitter rank, protocol, exit IP, ISP/ASN, residential classification, live AI latencies, speed, and 1-click copy button.
6. **Floating Right-Top Cluster Card**:
   - Displays all server nodes belonging to the selected city hub with individual IPs and speeds.
7. **Floating Right-Bottom 2D Global Traffic Map**:
   - Dedicated HUD card containing an authentic 2D vector world map with borders for 177 countries.
   - Live corridor matching the active 3D node drawn from detected host origin to destination city:
     - **Green Arc (`#10b981`)**: Available proxy passing target checks.
     - **Red Dashed Arc (`#f43f5e`)**: Blocked, rate-limited, or failed connection path.

---

### 2.3 View 3: AI Probe Studio (`#view-probes`)
1. **Interactive Evaluator Cards**:
   - Grid cards for Google Gemini, Google AI Studio, OpenAI ChatGPT, Anthropic Claude, and Low-Bandwidth Speedtest.
   - Each card features:
     - On/Off toggle switch.
     - Target domain and inspection profile.
     - Real-time pass status and empirical live latency.
     - **Dynamic Timeouts**: Reflects the active Stage 2 timeout configured in the Settings modal (`8,000 ms`), updating in real time when settings change.
2. **Speedtest Gating**:
   - Labeled `Available Nodes Only` (rendered in bright emerald green), enforcing zero bandwidth waste on blocked nodes.
3. **Custom Probe Builder Card**:
   - Form allowing users to input probe name, target URL, and expected status code/body keyword.
   - Dynamically extends active probe suite without server restart.

---

### 2.4 View 4: Subscriptions & Mobile Hub (`#view-subscriptions`)
1. **Live Subscription Endpoint**:
   - Displays local auto-updating URL: `http://127.0.0.1:19090/api/export?format=sub&filter=clean`.
   - Copy button and interval notification for desktop/mobile clients.
2. **Preset Export Filters**:
   - Quick export buttons: `All Available Nodes (N)`, `Top 10 Fastest (>40Mbps)`, `Residential Only`, `Europe Only`.
3. **Smart Tagged Remark Preview**:
   - Live visual box demonstrating injected tags:
     `[DE] [Hetzner] [RES] [48Mbps] [145ms] [ChatGPT+Gemini] - Node-DE-Fast`.
4. **Mobile QR Code Card**:
   - Scannable QR code for instant import into v2rayNG, Streisand, Shadowrocket, and Clash Verge.

---

## 3. High-Concurrency Engine Settings Modal

Triggered via `20x` button in the Floating Command Dock:
- **Worker Concurrency**: Adjustable from 5 to 100 parallel workers.
- **Stage 1 Ping Timeout**: 1,000 to 8,000 ms.
- **Stage 2 Probe Timeout**: 3,000 to 15,000 ms.
- **Speedtest Payload Size**: 1 to 5 MB.
- **Engine Safety**: Xray process startups are scoped to prevent lock contention, and Stage 2 fallback uses parallel goroutines (`sync.WaitGroup`) instead of sequential blocking loops.
- **Thread Safety**: Deep copy snapshots (`Clone()`) prevent Go runtime `concurrent map iteration and map write` panics during live SSE streaming.

---

## 4. Backend-to-Frontend Module Architecture Mapping

| Backend Module / API | Output Data Model | Connected Frontend View & Component |
| :--- | :--- | :--- |
| **`GET /api/session`** | `EvaluationStateResponse` (reports, scan_status, config_count, origin) | **Dashboard**: Matrix table rows, KPI calculations, Health stacked bar, auto-detected origin. |
| **`GET /api/origin`** | `domain.IPEnrichment` (IP, CountryCode, City, Lat, Lon) | **Auto Egress Hub**: Configures origin anchor for 3D Globe, table corridors, and 2D Traffic Map. |
| **`GET /api/stream` (SSE)** | Named events (`item_update`, `stats`) | **Live Telemetry**: Real-time row status updates, progress bar in dock, dynamic KPI counter increments. |
| **`internal/probe/ip_intel.go`** | `domain.IPEnrichment` (ExitIP, CountryCode, City, ISP, ConnectionType, Lat, Lon) | **Fleet Globe 3D**: Dynamic country flag mapping, Great-Circle trajectories, Left Inspector Card, and 2D Traffic Map. |
| **`internal/probe/factory.go`** | `ProbeSuite`, `AvailableProbes` | **Probe Studio**: Service card active switches, dynamic timeout displays, and custom probe catalog. |
| **`GET /api/export`** | Clean base64 subscription / raw text / JSON | **Subscriptions Hub**: 1-click copy feed, format buttons, and preset filtered exports. |
| **`POST /api/scan/start`** | Execution params (`concurrency`, `services`, `speedtest`, timeouts) | **Floating Command Dock**: Triggered by `[ Start Live Audit ]`. |
| **`POST /api/ingest`** | Parse result, config count | **Ingest Modal**: Triggered by `[ Ingest Configs ]`. |

---

## 5. User Stories & Acceptance Criteria

### User Story 1 - Seamless Cockpit Navigation (Priority: P1)
**Given** the user opens `http://127.0.0.1:19090`,  
**When** the page loads,  
**Then** the Dashboard view is displayed by default with KPI cards, dynamic filter pills, the matrix table, and the floating command dock at the bottom, without any bulky action buttons in the header.

### User Story 2 - Interactive 3D Fleet Globe & Non-Penetrating Great-Circle Corridors (Priority: P1)
**Given** configurations have been ingested and scanned,  
**When** the user clicks the `Fleet Globe 3D` tab,  
**Then** a rotatable Three.js satellite Earth globe is rendered with dimmed deserts, small sharp node pins, and an interactive 2D vector Global Traffic Map with country borders. Clicking a node illuminates a single Great-Circle geodesic trajectory from the detected host origin to that city, with tiny pure-white photon beads streaming continuously above the spherical surface without clipping into the Earth's interior.

### User Story 3 - Probe Studio Configuration & Dynamic Timeouts (Priority: P2)
**Given** the user opens `Probe Studio` or adjusts engine timeouts,  
**When** the timeout slider in Settings is modified,  
**Then** all service cards display the synchronized timeout (`8,000 ms`), and the Speedtest gating clearly displays `Available Nodes Only`.

### User Story 4 - Subscriptions & Mobile Sync (Priority: P2)
**Given** a completed or in-progress scan,  
**When** the user opens the `Subscriptions` tab,  
**Then** the live subscription feed URL, preset export options, and mobile QR code are immediately available.
