# Requirements Quality Checklist: 004 - IP Intelligence & Speedtest

- [x] **Zero-Assumption Empirical Ground Truth**: IP metadata is enriched after empirical probing; connection type (`Residential` vs `Data Center`) is reported as diagnostic data, never as a drop gate.
- [x] **No Premature Code Gates**: Requirements follow direct research and user requests for IP intelligence and bandwidth testing.
- [x] **Bandwidth Safety**: Strict 3000ms hard ceiling and optional toggle prevent unexpected mobile/metered data consumption.
- [x] **Multi-Tier Fallback**: IP intelligence is protected by 3-tier fallback (`ipwho.is` -> `ipleak.net` -> `cloudflare trace`).
- [x] **Visual Matrix Completeness**: Speed column has distinct color tiers; connection type has visual badges (`🏠 Res` / `🏢 DC`).
- [x] **Export Formatting Consistency**: Smart tags in subscription export include speed and connection type.
