package engine

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"cips/internal/domain"
	"cips/internal/pipeline"
)

// XrayInstance implements domain.EngineInstance for a single proxy in a pool.
type XrayInstance struct {
	endpoint string
	config   domain.ProxyConfig
	onClose  func() error
}

func (xi *XrayInstance) SOCKS5Endpoint() string {
	return xi.endpoint
}

func (xi *XrayInstance) Config() domain.ProxyConfig {
	return xi.config
}

func (xi *XrayInstance) Close() error {
	if xi.onClose != nil {
		return xi.onClose()
	}
	return nil
}

// XrayEngine manages Xray core instances with multi-inbound pool routing.
type XrayEngine struct {
	mu          sync.Mutex
	xrayPath    string
	activeCmds  map[int]*exec.Cmd
	tempFiles   []string
	portManager *pipeline.PortManager
}

// NewXrayEngine creates an XrayEngine instance.
func NewXrayEngine(xrayPath string) *XrayEngine {
	return &XrayEngine{
		xrayPath:    xrayPath,
		activeCmds:  make(map[int]*exec.Cmd),
		portManager: pipeline.DefaultPortManager,
	}
}

// Ping performs a direct TCP/TLS handshake against the proxy server for Stage 1.
func (xe *XrayEngine) Ping(ctx context.Context, cfg domain.ProxyConfig, timeoutMs int) (int64, error) {
	if timeoutMs <= 0 {
		timeoutMs = 3000
	}
	timeout := time.Duration(timeoutMs) * time.Millisecond

	addr := net.JoinHostPort(cfg.Server, strconv.Itoa(cfg.Port))
	dialer := &net.Dialer{Timeout: timeout}

	start := time.Now()

	// If TLS/Reality is configured, test TLS handshake as well
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

// StartPool creates a single Xray process routing multiple local inbounds to multiple outbounds.
func (xe *XrayEngine) StartPool(ctx context.Context, configs []domain.ProxyConfig) ([]domain.EngineInstance, error) {
	if len(configs) == 0 {
		return nil, nil
	}

	ports, err := xe.portManager.AllocateBatch(len(configs))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate local ports for pool: %w", err)
	}

	xrayConfig, instances := xe.buildPoolConfig(configs, ports)

	// Save temporary config JSON
	tmpFile, err := os.CreateTemp("", "cips_xray_*.json")
	if err != nil {
		xe.portManager.ReleaseBatch(ports)
		return nil, fmt.Errorf("failed to create temp xray config: %w", err)
	}

	configBytes, err := json.MarshalIndent(xrayConfig, "", "  ")
	if err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		xe.portManager.ReleaseBatch(ports)
		return nil, fmt.Errorf("failed to marshal xray config: %w", err)
	}

	if _, err := tmpFile.Write(configBytes); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		xe.portManager.ReleaseBatch(ports)
		return nil, fmt.Errorf("failed to write xray config: %w", err)
	}
	configPath := tmpFile.Name()
	_ = tmpFile.Close()

	cmd := exec.CommandContext(ctx, xe.xrayPath, "run", "-c", configPath)
	cmd.Stdout = nil
	cmd.Stderr = GlobalEngineLog

	if err := cmd.Start(); err != nil {
		_ = os.Remove(configPath)
		xe.portManager.ReleaseBatch(ports)
		return nil, fmt.Errorf("failed to start xray process: %w", err)
	}

	cmdPid := cmd.Process.Pid

	xe.mu.Lock()
	xe.activeCmds[cmdPid] = cmd
	xe.tempFiles = append(xe.tempFiles, configPath)
	xe.mu.Unlock()

	// Dynamic socket readiness polling (1.5s for single fallback, up to 3.5s for large multi-inbound pools)
	readyBudget := 1500 * time.Millisecond
	if len(ports) > 1 {
		readyBudget = time.Duration(1200+len(ports)*40) * time.Millisecond
		if readyBudget > 4000*time.Millisecond {
			readyBudget = 4000 * time.Millisecond
		}
	}

	if err := WaitForPortsReady(ports, readyBudget); err != nil {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = os.Remove(configPath)
		xe.mu.Lock()
		delete(xe.activeCmds, cmdPid)
		xe.mu.Unlock()
		xe.portManager.ReleaseBatch(ports)
		return nil, fmt.Errorf("xray socket initialization failed: %w", err)
	}

	var killOnce sync.Once
	killProcess := func() {
		killOnce.Do(func() {
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			_ = os.Remove(configPath)
			xe.mu.Lock()
			delete(xe.activeCmds, cmdPid)
			xe.mu.Unlock()
		})
	}

	// Wrap instances with cleanup
	finalInstances := make([]domain.EngineInstance, len(instances))
	for i, inst := range instances {
		port := ports[i]
		finalInstances[i] = &XrayInstance{
			endpoint: inst.endpoint,
			config:   inst.config,
			onClose: func() error {
				xe.portManager.ReleasePort(port)
				killProcess()
				return nil
			},
		}
	}

	// Goroutine to cleanup process when context finishes or command exits
	go func() {
		_ = cmd.Wait()
		killProcess()
		xe.portManager.ReleaseBatch(ports)
	}()

	return finalInstances, nil
}

// Shutdown terminates all active Xray child processes and cleans up temporary config files.
func (xe *XrayEngine) Shutdown() error {
	xe.mu.Lock()
	defer xe.mu.Unlock()

	for _, cmd := range xe.activeCmds {
		if cmd != nil && cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}
	xe.activeCmds = make(map[int]*exec.Cmd)

	for _, f := range xe.tempFiles {
		_ = os.Remove(f)
	}
	xe.tempFiles = nil

	return nil
}

func (xe *XrayEngine) buildPoolConfig(configs []domain.ProxyConfig, ports []int) (map[string]interface{}, []*XrayInstance) {
	inbounds := make([]map[string]interface{}, 0, len(configs))
	outbounds := make([]map[string]interface{}, 0, len(configs)+1)
	rules := make([]map[string]interface{}, 0, len(configs))
	instances := make([]*XrayInstance, 0, len(configs))

	for i, cfg := range configs {
		port := ports[i]
		inboundTag := fmt.Sprintf("inbound_%d", port)
		outboundTag := fmt.Sprintf("outbound_%d", port)

		inbounds = append(inbounds, map[string]interface{}{
			"tag":      inboundTag,
			"port":     port,
			"listen":   "127.0.0.1",
			"protocol": "socks",
			"settings": map[string]interface{}{
				"auth": "noauth",
				"udp":  false,
			},
		})

		outbound := xe.convertConfigToOutbound(cfg, outboundTag)
		outbounds = append(outbounds, outbound)

		rules = append(rules, map[string]interface{}{
			"type":        "field",
			"inboundTag":  []string{inboundTag},
			"outboundTag": outboundTag,
		})

		instances = append(instances, &XrayInstance{
			endpoint: fmt.Sprintf("127.0.0.1:%d", port),
			config:   cfg,
		})
	}

	// Fallback direct outbound
	outbounds = append(outbounds, map[string]interface{}{
		"tag":      "direct",
		"protocol": "freedom",
	})

	config := map[string]interface{}{
		"log": map[string]interface{}{
			"loglevel": "warning",
		},
		"inbounds":  inbounds,
		"outbounds": outbounds,
		"routing": map[string]interface{}{
			"domainStrategy": "AsIs",
			"rules":          rules,
		},
	}

	return config, instances
}

// WaitForPortsReady polls a list of ports until all are accepting TCP connections or timeout occurs.
func WaitForPortsReady(ports []int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for _, p := range ports {
		addr := fmt.Sprintf("127.0.0.1:%d", p)
		ready := false
		for time.Now().Before(deadline) {
			conn, err := net.DialTimeout("tcp", addr, 25*time.Millisecond)
			if err == nil {
				_ = conn.Close()
				ready = true
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
		if !ready {
			return fmt.Errorf("local port %d failed to accept connections within %v", p, timeout)
		}
	}
	return nil
}

func (xe *XrayEngine) convertConfigToOutbound(cfg domain.ProxyConfig, tag string) map[string]interface{} {
	switch cfg.Protocol {
	case domain.ProtocolVLESS:
		return xe.buildVLESSOutbound(cfg, tag)
	case domain.ProtocolVMess:
		return xe.buildVMessOutbound(cfg, tag)
	case domain.ProtocolTrojan:
		return xe.buildTrojanOutbound(cfg, tag)
	case domain.ProtocolShadowsocks:
		return xe.buildShadowsocksOutbound(cfg, tag)
	case domain.ProtocolSocks5:
		return xe.buildSocks5Outbound(cfg, tag)
	default:
		return xe.buildVLESSOutbound(cfg, tag)
	}
}

func (xe *XrayEngine) buildVLESSOutbound(cfg domain.ProxyConfig, tag string) map[string]interface{} {
	streamSettings := xe.buildStreamSettings(cfg)

	userObj := map[string]interface{}{
		"id":         cfg.UUID,
		"encryption": "none",
	}
	if flow, ok := cfg.ExtraParams["flow"]; ok && flow != "" {
		userObj["flow"] = flow
	}

	return map[string]interface{}{
		"tag":      tag,
		"protocol": "vless",
		"settings": map[string]interface{}{
			"vnext": []map[string]interface{}{
				{
					"address": cfg.Server,
					"port":    cfg.Port,
					"users":   []map[string]interface{}{userObj},
				},
			},
		},
		"streamSettings": streamSettings,
	}
}

func (xe *XrayEngine) buildVMessOutbound(cfg domain.ProxyConfig, tag string) map[string]interface{} {
	streamSettings := xe.buildStreamSettings(cfg)
	return map[string]interface{}{
		"tag":      tag,
		"protocol": "vmess",
		"settings": map[string]interface{}{
			"vnext": []map[string]interface{}{
				{
					"address": cfg.Server,
					"port":    cfg.Port,
					"users": []map[string]interface{}{
						{
							"id":       cfg.UUID,
							"alterId":  0,
							"security": "auto",
						},
					},
				},
			},
		},
		"streamSettings": streamSettings,
	}
}

func (xe *XrayEngine) buildTrojanOutbound(cfg domain.ProxyConfig, tag string) map[string]interface{} {
	streamSettings := xe.buildStreamSettings(cfg)
	return map[string]interface{}{
		"tag":      tag,
		"protocol": "trojan",
		"settings": map[string]interface{}{
			"servers": []map[string]interface{}{
				{
					"address":  cfg.Server,
					"port":     cfg.Port,
					"password": cfg.UUID,
				},
			},
		},
		"streamSettings": streamSettings,
	}
}

func (xe *XrayEngine) buildShadowsocksOutbound(cfg domain.ProxyConfig, tag string) map[string]interface{} {
	method := cfg.ExtraParams["method"]
	if method == "" {
		method = "aes-256-gcm"
	}
	return map[string]interface{}{
		"tag":      tag,
		"protocol": "shadowsocks",
		"settings": map[string]interface{}{
			"servers": []map[string]interface{}{
				{
					"address":  cfg.Server,
					"port":     cfg.Port,
					"method":   method,
					"password": cfg.UUID,
				},
			},
		},
	}
}

func (xe *XrayEngine) buildSocks5Outbound(cfg domain.ProxyConfig, tag string) map[string]interface{} {
	return map[string]interface{}{
		"tag":      tag,
		"protocol": "socks",
		"settings": map[string]interface{}{
			"servers": []map[string]interface{}{
				{
					"address": cfg.Server,
					"port":    cfg.Port,
				},
			},
		},
	}
}

func (xe *XrayEngine) buildStreamSettings(cfg domain.ProxyConfig) map[string]interface{} {
	network := cfg.Type
	if network == "" {
		network = "tcp"
	}

	stream := map[string]interface{}{
		"network": network,
	}

	sni := cfg.SNI
	if sni == "" {
		sni = cfg.Host
	}
	if sni == "" {
		sni = cfg.Server
	}

	switch cfg.Security {
	case "reality":
		stream["security"] = "reality"
		reality := map[string]interface{}{
			"serverName":  sni,
			"fingerprint": "chrome",
		}
		if pbk, ok := cfg.ExtraParams["pbk"]; ok {
			reality["publicKey"] = pbk
		}
		if sid, ok := cfg.ExtraParams["sid"]; ok {
			reality["shortId"] = sid
		}
		if spx, ok := cfg.ExtraParams["spx"]; ok {
			reality["spiderX"] = spx
		}
		stream["realitySettings"] = reality

	case "tls":
		stream["security"] = "tls"
		stream["tlsSettings"] = map[string]interface{}{
			"serverName":  sni,
			"fingerprint": "chrome",
		}
	}

	switch strings.ToLower(network) {
	case "ws":
		ws := map[string]interface{}{
			"path": cfg.Path,
		}
		if cfg.Host != "" {
			ws["headers"] = map[string]string{"Host": cfg.Host}
		}
		stream["wsSettings"] = ws
	case "grpc":
		grpc := map[string]interface{}{
			"serviceName": cfg.Path,
		}
		stream["grpcSettings"] = grpc
	}

	return stream
}
