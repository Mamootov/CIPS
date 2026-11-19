# Implementation Plan: 005 - Dynamic Multi-Service Probe Catalog

**Feature Branch**: `005-dynamic-service-probes`  
**Status**: Ready for Implementation  

---

## 1. Technical Architecture & Component Flow

```
┌────────────────────────────────────────────────────────┐
│                        web/app.js                      │
│  - Service Chips: Gemini, AI Studio, ChatGPT, Claude   │
│  - Custom URL Builder: "+ Add Custom"                  │
│  - Dynamic Table Headers & Cells Rendering             │
└───────────────────────────┬────────────────────────────┘
                            │ POST /api/scan/start
                            │ { services: ["chatgpt", ...], custom_probes: [...] }
┌───────────────────────────▼────────────────────────────┐
│                  internal/server/handlers.go           │
│  - Builds requested probe instances via ProbeFactory   │
│  - Sets active probe slice in orchestrator             │
└───────────────────────────┬────────────────────────────┘
                            │
┌───────────────────────────▼────────────────────────────┐
│               internal/pipeline/orchestrator.go        │
│  - Evaluates ONLY selected probes in parallel          │
│  - Sets IsClean = true if any selected service passes  │
│  - Runs Speedtest conditionally if any selected passes │
└──────────────┬───────────────────────────┬─────────────┘
               │                           │
┌──────────────▼────────────┐ ┌────────────▼─────────────┐
│    Built-in AI Probes     │ │ Custom / Universal Engine│
│ - gemini.go               │ │ - custom.go (Arbitrary)  │
│ - aistudio.go             │ │ - inspector.go (English) │
│ - chatgpt.go (New)        │ │ - ip_intel.go (Metadata) │
│ - claude.go (New)         │ │ - speedtest.go (Bandwidth│
└───────────────────────────┘ └──────────────────────────┘
```

---

## 2. Component Implementation Details

### 2.1 ChatGPT Native Probe (`internal/probe/chatgpt.go`)
- Target URL: `https://chatgpt.com/`
- Request Headers: Standard Chrome 124 TLS fingerprint (`Accept-Language: en-US,en;q=0.9`)
- Inspection rules:
  - If status 403 / 503 or `Server: cloudflare` with `cf-ray`: Flag `CLOUDFLARE_CHALLENGE`.
  - If body contains `"not available in your country"` or `"unsupported_country"`: Flag `GEO_RESTRICTED`.
  - If status 200 without challenge tokens: Flag `PASS`.

### 2.2 Claude Native Probe (`internal/probe/claude.go`)
- Target URL: `https://claude.ai/login` (or `https://claude.ai/`)
- Inspection rules:
  - Intercept redirect to `anthropic.com/supported-countries`: Flag `GEO_RESTRICTED`.
  - Detect Cloudflare Turnstile challenge: Flag `CLOUDFLARE_CHALLENGE`.
  - If status 200 without block tokens: Flag `PASS`.

### 2.3 Universal English-Only Inspector (`internal/probe/inspector.go`)
- Remove Persian strings (`"در کشورتان دردسترس نیست"`, etc.) from `geoblockTokens`.
- Maintain canonical English tokens:
  ```go
  geoblockTokens := []string{
      "isn't supported in your country",
      "isn't currently supported in your country",
      "location_error",
      "unsupported country",
      "user location not supported",
      "not available in your country",
      "not available in your region",
      "your region may not be supported",
      "region may not be supported",
      "available-regions",
      "unsupported_country",
  }
  ```

### 2.4 Probe Factory (`internal/probe/factory.go`)
- Constructs slice of `domain.IProbe` based on list of service IDs and custom definitions:
  - `"gemini"` -> `NewGeminiProbe()`
  - `"aistudio"` -> `NewAIStudioProbe()`
  - `"chatgpt"` -> `NewChatGPTProbe()`
  - `"claude"` -> `NewClaudeProbe()`
  - Custom -> `NewCustomProbe(def)`

### 2.5 Dynamic Frontend Grid (`web/app.js`, `web/index.html`)
- Sidebar includes selectable chips: `Gemini`, `AI Studio`, `ChatGPT`, `Claude`, and `+ Custom`.
- `renderTableHeaders()`: Dynamically inserts `<th>` tags for each active service.
- `populateRowContent()`: Dynamically inserts `<td>` tags matching the active headers.
- Export tags format based on surviving selected services.
