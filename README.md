<p align="center">
  <img src="web/assets/logo.png" width="180" height="180" alt="CIPS Logo" style="border-radius: 24px;" />
</p>

<h1 align="center">CIPS — Clean IP Scanner</h1>

<p align="center">
  <b>High-throughput proxy auditor, benchmark engine, and observability cockpit for LLM APIs.</b><br>
  Validates proxies against Google Gemini, Google AI Studio, OpenAI ChatGPT, and Anthropic Claude using genuine Chrome 124 TLS/HTTP2 fingerprints, multi-port core routing, and Cloudflare Anycast bandwidth benchmarking.
</p>

<p align="center">
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go" alt="Go Version" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT" /></a>
  <a href="https://github.com/Mamootov/CIPS"><img src="https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey" alt="Platform" /></a>
</p>

---

## The Problem

Testing proxies for artificial intelligence services is unreliable with traditional ping or curl tools. A proxy node that responds to simple ICMP/TCP pings or opens Google Search can still fail silently on generative AI platforms due to:
1. **TLS Fingerprint Mismatch**: Target WAFs (such as Cloudflare Turnstile) inspect JA3/JA4 fingerprints and reject default Go, Python, or curl handshakes with HTTP 403.
2. **Internal Region Redirection**: Google AI Studio and Gemini return HTTP 200 on landing pages while silently redirecting restricted regions to `/available-regions`.
3. **Geolocation Tainting**: Search engines localize datacenter IP locations based on historical user search queries, causing false negatives.

CIPS solves this by treating target service verification as an empirical, browser-identical ground truth.

---

## How It Works & Architecture

CIPS implements a non-blocking, two-stage evaluation pipeline backed by an internal proxy orchestration engine and an embedded web observability cockpit.

```
Candidate Proxies (VLESS, VMess, Trojan, SS, SOCKS5)
                    │
                    ▼
       [ Ingestion & Deduplication ]
                    │
                    ▼
     [ Stage 1: Fast TCP/TLS Triage ] ──────────► (Dead Nodes Dropped Early)
                    │
                    ▼
  [ Stage 2: Dynamic Local Port Router ]
     (Spawns isolated local SOCKS5 tunnels via Xray core)
                    │
                    ▼
   [ Stage 2A: Browser-Mimic Probing ]
     (Chrome 124 TLS + HTTP/2 client sends real requests to:)
       ├── Google Gemini (gemini.google.com)
       ├── Google AI Studio (aistudio.google.com)
       ├── OpenAI ChatGPT (chatgpt.com)
       ├── Anthropic Claude (claude.ai)
       └── IP Intelligence (Residential vs Data Center, ASN, City)
                    │
                    ▼
         Passed at least 1 AI service?
            ├── YES ──► [ Stage 2B: Anycast Speedtest (Cloudflare Edge) ]
            └── NO  ──► [ Skip Speedtest & Flag Failure Layer ]
                    │
                    ▼
       [ Real-Time SSE Broadcaster ]
                    │
         ┌──────────┴──────────┐
         ▼                     ▼
[ Web Cockpit UI ]    [ Clean Subscription Feed ]
(3D Globe & Tables)   (Smart tagged URL for clients)
```

### 1. Ingestion & Normalization
CIPS accepts subscription links, raw text blocks, and file uploads. It decodes base64 envelopes, parses remarks (retaining URI encoding), and deduplicates nodes based on unique connection fingerprints (`protocol://server:port`).

### 2. Stage 1: Fast Reachability Handshake
Before allocating local sockets or spawning sub-processes, candidate nodes undergo a concurrent TCP/TLS ping triage with a strict 1.5s–3s timeout. Dead hosts and invalid ports are weeded out early without wasting local system resources.

### 3. Local Multi-Port Proxy Engine
Surviving candidates are batched into dynamically allocated local SOCKS5 inbounds managed by an internal `Xray-core` instance. Local port readiness is actively monitored via socket polling with per-port error isolation. If an invalid Reality public key or malformed cipher causes a batch failure, the engine automatically falls back to isolated single-instance goroutines.

### 4. Stage 2A: Browser-Mimic Probing & Diagnostic Inspection
Each proxy route is probed using a custom Chrome 124 TLS client over HTTP/2 with exact browser cipher suites, header ordering, and pseudo-header structures. 

The diagnostic inspector parses HTTP response layers, redirection chains, and stderr logs into precise error classifications:
- `GEO_RESTRICTED`: Region limitation banners or hops to `/available-regions`.
- `CLOUDFLARE_CHALLENGE`: Turnstile challenges, Under-Attack interstitial screens, or WAF blocks.
- `BAD_REALITY_KEY`: TLS handshakes rejected due to invalid public keys (`pbk`) or short IDs (`sid`).
- `UPSTREAM_REFUSED`: Target server socket resets or dropped SYNs.

### 5. Stage 2B: Conditional Anycast Speedtest
To avoid wasting proxy quota and system bandwidth, the streaming download speedtest (3 MB chunk against Cloudflare's global Anycast CDN) executes **only** on proxies that have successfully passed at least one active AI service test.

### 6. Real-Time Telemetry & Clean Export
Every state transition is broadcast over Server-Sent Events (`/api/stream`) using thread-safe, cloned memory snapshots. Validated proxies are immediately available through an auto-updating subscription endpoint with enriched metadata tags.

---

## Cockpit Views

- **Dashboard**: Aggregated Sentry-style health indicators, percentile latencies (`p50`, `p95`), jitter calculations, and an observability table with automatic speed sorting for available nodes.
- **Fleet Globe 3D**: An interactive Three.js tactical Earth featuring non-penetrating `GeodesicOrbitCurve` flight arcs that sweep above the planetary terrain across all coordinates (including Australia and the Americas), illuminated by real-time white photon packet animations.
- **2D Global Traffic Map**: A synchronized planar vector projection displaying exact country borders and live traffic paths corresponding to the selected 3D node.
- **Probe Studio**: Fine-grained controls to toggle individual AI validators on or off, define custom HTTP/JSON endpoints, and configure engine timeouts.
- **Subscriptions Hub**: Dynamic, client-ready subscription feeds with preset filters (Fastest, Residential, Regional) and mobile QR codes.

---

## Supported Protocols

| Protocol | Security / Transports |
| :--- | :--- |
| **VLESS** | Reality, XTLS-Vision, WebSocket, gRPC, TCP |
| **VMess** | WebSocket, TCP, TLS |
| **Trojan** | TLS, WebSocket, gRPC |
| **Shadowsocks** | AEAD Ciphers (2022-blake3, ChaCha20-Poly1305, AES-128/256-GCM) |
| **SOCKS5 / HTTP** | Standard authentications |

---

## Quick Start

CIPS is written in pure Go with embedded static assets and natively supports **Windows**, **Linux**, and **macOS** (x86_64 and Apple Silicon ARM64).

### Pre-built Binary
1. Download the archive for your operating system from [Releases](https://github.com/Mamootov/CIPS/releases).
2. Launch the application:
   - **Windows**: Double-click `run.bat` or run:
     ```cmd
     cips.exe
     ```
   - **Linux / macOS**:
     ```bash
     chmod +x run.sh cips
     ./run.sh
     ```
3. Open `http://127.0.0.1:19090` in your web browser.

### Build from Source
Prerequisites: Go 1.22+ and Git.

```bash
git clone https://github.com/Mamootov/CIPS.git
cd CIPS

# On Linux / macOS:
go build -ldflags="-s -w" -o cips ./cmd/cips
./cips

# On Windows:
go build -ldflags="-s -w" -o cips.exe ./cmd/cips
./cips.exe
```

---

## Sanitized Subscription Feed

CIPS functions as an automated proxy filter. Configure your client (v2rayNG, Clash Verge, Shadowrocket, Streisand) to fetch the sanitized feed URL:

```
http://127.0.0.1:19090/api/export?format=sub&filter=clean
```

Verified proxies are tagged with real-time performance indicators:
```
vless://...#[DE] [Hetzner] [RES] [48Mbps] [145ms] [ChatGPT+Gemini] - Node-01
```

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
