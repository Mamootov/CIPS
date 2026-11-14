# Contract: Probe & ProxyEngine Interfaces

## Overview
This contract defines the core Go interfaces that ensure CIPS adheres to Hexagonal Architecture (Port & Adapter separation) and Open-Closed Extensibility.

---

## 1. IProxyEngine (Port: Proxy Infrastructure)

Responsible for spinning up local SOCKS5/HTTP inbound tunnels mapped to outbound proxy configurations.

```go
package engine

import (
    "context"
    "cips/internal/domain"
)

type EngineInstance interface {
    // SOCKS5Endpoint returns the local address, e.g. "127.0.0.1:20005"
    SOCKS5Endpoint() string
    // Close releases ports and kills associated processes
    Close() error
}

type IProxyEngine interface {
    // StartPool initializes a multi-inbound proxy pool for a batch of configurations
    StartPool(ctx context.Context, configs []domain.ProxyConfig) ([]EngineInstance, error)
    
    // Ping conducts a fast TCP/TLS latency handshake for Stage 1 triage
    Ping(ctx context.Context, config domain.ProxyConfig, timeoutMs int) (latencyMs int64, err error)
    
    // Shutdown forcefully terminates all running core instances
    Shutdown() error
}
```

---

## 2. IProbe (Port: Service Evaluation)

Responsible for evaluating a target service through a specified local proxy endpoint.

```go
package probe

import (
    "context"
    "cips/internal/domain"
)

type IProbe interface {
    // ID returns the unique probe identifier (e.g. "gemini", "ai-studio", "ipleak")
    ID() string
    
    // Name returns human-readable label for dashboard display
    Name() string
    
    // Evaluate executes the test request through the proxy endpoint
    Evaluate(ctx context.Context, proxyEndpoint string) domain.ProbeResult
}
```

---

## 3. Standard Probes Specification

### 3.1 GeminiProbe
- **Target URL**: `https://gemini.google.com/app`
- **Method**: `GET`
- **TLS Client Profile**: `tlsclient.Chrome_124`
- **HTTP/2**: Enabled
- **Validation**:
  - Response Code == 200
  - Body does NOT contain: `"location_error"`, `"isn't supported in your country"`, or `"Unsupported Country"`

### 3.2 AIStudioProbe
- **Target URL**: `https://aistudio.google.com/`
- **Method**: `GET`
- **TLS Client Profile**: `tlsclient.Chrome_124`
- **Validation**:
  - Response Code == 200 or 302 (redirect to accounts.google.com sign-in is ACCEPTABLE)
  - Must NOT redirect to geo-blocked error pages (e.g., `/unsupported-country` or 451/403)

### 3.3 IPLeakProbe
- **Target URL**: `https://ipleak.net/json/`
- **Method**: `GET`
- **Response**: JSON schema containing `ip`, `isp_name`, `country_name`, `country_code`, `as_number`.
