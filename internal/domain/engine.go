package domain

import (
	"context"
)

// EngineInstance represents an active local inbound listener connected to an outbound proxy.
type EngineInstance interface {
	// SOCKS5Endpoint returns the local host:port string, e.g. "127.0.0.1:20005"
	SOCKS5Endpoint() string

	// ProxyConfig returns the associated proxy configuration
	Config() ProxyConfig

	// Close terminates this tunnel instance and frees ports
	Close() error
}

// IProxyEngine defines the port for launching proxy tunnel backends.
type IProxyEngine interface {
	// StartPool starts a local multi-inbound proxy pool for a batch of configurations
	StartPool(ctx context.Context, configs []ProxyConfig) ([]EngineInstance, error)

	// Ping performs a lightweight TCP/TLS reachability check for Stage 1 triage
	Ping(ctx context.Context, config ProxyConfig, timeoutMs int) (latencyMs int64, err error)

	// Shutdown terminates all active underlying engine processes and frees resources
	Shutdown() error
}
