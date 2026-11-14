# Quickstart & Validation Guide: Clean IP Scanner (CIPS)

## Overview
This document guides developers and testers through setting up, building, running, and verifying CIPS on Windows and Linux.

---

## 1. Prerequisites

1. **Go Toolchain**: Go 1.22+ (`go version`).
2. **Core Binary**: `xray.exe` (Windows) or `xray` (Linux).
   *(Note: If missing, CIPS automatically downloads the verified release from official GitHub releases/mirrors).*
3. **Browser**: Any modern browser (Chrome, Edge, Firefox) for the dashboard view.

---

## 2. Build & Launch

### Compile Single Binary
```bash
# In repository root
go mod tidy
go build -o cips.exe ./cmd/cips
```

### Launch Interactive Mode (Web Dashboard)
```bash
./cips.exe
```
*Expected Output:*
- Prints startup banner with version and port `http://127.0.0.1:19090`
- Automatically opens default web browser to the CIPS dashboard

### Launch Headless / CLI Server Mode
```bash
./cips.exe --headless --port 19090
```

---

## 3. End-to-End Validation Scenarios

### Scenario A: Drag-and-Drop Ingestion & Two-Stage Scan
1. Open the dashboard at `http://127.0.0.1:19090`.
2. Drag and drop a sample text file containing 10+ mixed proxy links (`vless://`, `vmess://`, `ss://`).
3. Verify that the candidate count matches the number of valid non-duplicate links.
4. Click **Start Scan**:
   - Observe Stage 1 ping tests executing concurrently.
   - Watch the live matrix table: unresponsive nodes transition to `STAGE1_DEAD`.
   - Responsive nodes proceed to Stage 2: observe live column checkpoints for **Gemini**, **Google AI Studio**, and **IPLeak**.
5. When complete, verify that the **Clean & Usable** tab displays only nodes with passing checkpoints.

### Scenario B: Resilient Teardown on User Cancellation
1. Ingest a large batch (100+ proxies) and click **Start Scan**.
2. Mid-scan (at ~30% progress), click **Cancel Scan**.
3. Verify:
   - All active worker routines halt within 1 second.
   - All ephemeral ports are immediately released.
   - System memory returns to baseline without leaked background processes.

### Scenario C: Export Clean Subscription
1. After scan completion, click **Copy Subscription** or **Download Clean List**.
2. Paste the exported URL into V2RayN, v2rayNG, or Clash.
3. Verify that the remark contains enriched tags (e.g. `[US] [Google-OK] [Hetzner]`).
