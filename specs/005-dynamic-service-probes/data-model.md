# Data Model: 005 - Dynamic Service Probes

---

## 1. REST API Contract Updates

### 1.1 `POST /api/scan/start`
Request Payload:
```json
{
  "concurrency": 20,
  "stage1_timeout_ms": 3000,
  "stage2_timeout_ms": 8000,
  "enable_speedtest": true,
  "speedtest_bytes": 3000000,
  "services": ["chatgpt", "claude"],
  "custom_probes": [
    {
      "id": "flow",
      "name": "Flow AI",
      "target_url": "https://flow.ai/login",
      "method": "GET",
      "expected_status": 200,
      "timeout_ms": 8000
    }
  ]
}
```

---

## 2. Dynamic Service Catalog Model

```go
type ServicePreset struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    DefaultURL  string `json:"default_url"`
    Description string `json:"description"`
}

var AvailablePresets = []ServicePreset{
    {ID: "gemini", Name: "Gemini", DefaultURL: "https://gemini.google.com/", Description: "Google Gemini Web App"},
    {ID: "aistudio", Name: "AI Studio", DefaultURL: "https://aistudio.google.com/", Description: "Google AI Studio Developer Console"},
    {ID: "chatgpt", Name: "ChatGPT", DefaultURL: "https://chatgpt.com/", Description: "OpenAI ChatGPT Web Interface"},
    {ID: "claude", Name: "Claude", DefaultURL: "https://claude.ai/", Description: "Anthropic Claude Web Interface"},
}
```
