package pipeline

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"time"

	"cips/internal/domain"
)

// FastPing performs a low-overhead TCP/TLS handshake to verify remote node reachability.
func FastPing(ctx context.Context, cfg domain.ProxyConfig, timeoutMs int) (int64, error) {
	if timeoutMs <= 0 {
		timeoutMs = 3000
	}
	timeout := time.Duration(timeoutMs) * time.Millisecond

	addr := net.JoinHostPort(cfg.Server, strconv.Itoa(cfg.Port))
	dialer := &net.Dialer{Timeout: timeout}

	start := time.Now()

	// If TLS or Reality is specified, execute TLS handshake to catch TLS layer blocks
	if cfg.Security == "tls" || cfg.Security == "reality" {
		sni := cfg.SNI
		if sni == "" {
			sni = cfg.Server
		}
		tlsDialer := &tls.Dialer{
			NetDialer: dialer,
			Config: &tls.Config{
				ServerName:         sni,
				InsecureSkipVerify: true,
			},
		}

		conn, err := tlsDialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			return 0, err
		}
		_ = conn.Close()
	} else {
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			return 0, err
		}
		_ = conn.Close()
	}

	latency := time.Since(start).Milliseconds()
	return latency, nil
}
