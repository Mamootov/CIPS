# Contract: Telemetry & Subscription Export Format

## 1. Clean Subscription Export Format (`.txt` / Base64 / Standard URL Scheme)

The clean subscription exporter produces a standard newline-delimited list of functional proxy URLs, with remarks dynamically rewritten to include verified metadata:

### Syntax Template
```text
<protocol>://<credentials>@<server>:<port>?<params>#[<COUNTRY>] [<ISP>] [<ACTIVE_TAGS>] [<PING>ms] - <ORIGINAL_NAME>
```

### Examples
- **VLESS Reality**:
  `vless://550e8400-e29b-41d4-a716-446655440000@194.87.21.45:443?security=reality&sni=yahoo.com&fp=chrome&pbk=xxx&sid=yyy&type=tcp#[RU] [Timeweb] [Gemini-OK,Studio-OK] [128ms] - Node-01`
- **VMess**:
  `vmess://<base64-encoded-json-with-enriched-ps-remark>`
- **Shadowsocks**:
  `ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@185.120.30.15:8388#[DE] [Hetzner] [Studio-OK] [85ms] - Frankfurt-04`

---

## 2. Structured Diagnostic Session Log (`session-<timestamp>.json`)

Lossless telemetry record saved on disk after every execution session:

```json
{
  "session_id": "cips-20261113-210530",
  "generated_at": "2026-11-13T21:05:30Z",
  "duration_seconds": 18.4,
  "summary": {
    "total_ingested": 150,
    "deduplicated": 120,
    "stage1_reachable": 82,
    "stage1_unreachable": 38,
    "ai_clean_passed": 45,
    "ai_clean_failed": 37
  },
  "configurations": [
    {
      "id": "a1b2c3d4...",
      "protocol": "vless",
      "server": "194.87.21.45",
      "port": 443,
      "stage1_ping_ms": 128,
      "probes": {
        "gemini": {
          "status": "PASS",
          "http_code": 200,
          "latency_ms": 420
        },
        "ai-studio": {
          "status": "PASS",
          "http_code": 200,
          "latency_ms": 380
        }
      },
      "ip_enrichment": {
        "exit_ip": "194.87.21.45",
        "isp": "Timeweb, LLP",
        "asn": "AS9123",
        "country_code": "RU",
        "country_name": "Russian Federation"
      },
      "is_clean": true,
      "export_uri": "vless://..."
    }
  ]
}
```
