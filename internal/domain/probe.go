package domain

import (
	"context"
)

// IProbe represents a pluggable service evaluation checker.
type IProbe interface {
	// ID returns unique identifier (e.g. "gemini", "aistudio", "ipleak")
	ID() string

	// Name returns human-readable label for dashboard display
	Name() string

	// Evaluate executes test request through the proxy endpoint (e.g. "127.0.0.1:20005")
	Evaluate(ctx context.Context, proxyEndpoint string) ProbeResult
}
