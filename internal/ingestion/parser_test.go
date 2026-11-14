package ingestion

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"cips/internal/domain"
)

func TestParseVLESS(t *testing.T) {
	raw := "vless://550e8400-e29b-41d4-a716-446655440000@194.87.21.45:443?security=reality&sni=yahoo.com&fp=chrome&pbk=testpubkey&sid=testsid&type=tcp#MyNode"
	cfg, err := ParseURI(raw)
	if err != nil {
		t.Fatalf("unexpected error parsing vless: %v", err)
	}

	if cfg.Protocol != domain.ProtocolVLESS {
		t.Errorf("expected protocol vless, got %s", cfg.Protocol)
	}
	if cfg.Server != "194.87.21.45" {
		t.Errorf("expected server 194.87.21.45, got %s", cfg.Server)
	}
	if cfg.Port != 443 {
		t.Errorf("expected port 443, got %d", cfg.Port)
	}
	if cfg.UUID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected uuid, got %s", cfg.UUID)
	}
	if cfg.Security != "reality" {
		t.Errorf("expected security reality, got %s", cfg.Security)
	}
	if cfg.SNI != "yahoo.com" {
		t.Errorf("expected sni yahoo.com, got %s", cfg.SNI)
	}
	if cfg.Remark != "MyNode" {
		t.Errorf("expected remark MyNode, got %s", cfg.Remark)
	}
}

func TestParseVMess(t *testing.T) {
	vmessData := map[string]interface{}{
		"v":    "2",
		"ps":   "TestVMess",
		"add":  "185.200.10.5",
		"port": "8080",
		"id":   "b831381d-6324-4d53-ad4f-8cda48b30811",
		"aid":  0,
		"net":  "ws",
		"type": "none",
		"host": "cloudflare.com",
		"path": "/graphql",
		"tls":  "tls",
	}
	b, _ := json.Marshal(vmessData)
	raw := "vmess://" + base64.StdEncoding.EncodeToString(b)

	cfg, err := ParseURI(raw)
	if err != nil {
		t.Fatalf("unexpected error parsing vmess: %v", err)
	}

	if cfg.Protocol != domain.ProtocolVMess {
		t.Errorf("expected protocol vmess, got %s", cfg.Protocol)
	}
	if cfg.Server != "185.200.10.5" {
		t.Errorf("expected server 185.200.10.5, got %s", cfg.Server)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
	if cfg.Type != "ws" {
		t.Errorf("expected ws, got %s", cfg.Type)
	}
	if cfg.Path != "/graphql" {
		t.Errorf("expected path /graphql, got %s", cfg.Path)
	}
	if cfg.Remark != "TestVMess" {
		t.Errorf("expected remark TestVMess, got %s", cfg.Remark)
	}
}

func TestParseTrojan(t *testing.T) {
	raw := "trojan://secretpass@my.trojan.server:443?security=tls&sni=google.com#TrojanRemark"
	cfg, err := ParseURI(raw)
	if err != nil {
		t.Fatalf("unexpected error parsing trojan: %v", err)
	}

	if cfg.Protocol != domain.ProtocolTrojan {
		t.Errorf("expected trojan, got %s", cfg.Protocol)
	}
	if cfg.UUID != "secretpass" {
		t.Errorf("expected secretpass, got %s", cfg.UUID)
	}
	if cfg.Server != "my.trojan.server" {
		t.Errorf("expected my.trojan.server, got %s", cfg.Server)
	}
	if cfg.SNI != "google.com" {
		t.Errorf("expected google.com, got %s", cfg.SNI)
	}
}

func TestParseShadowsocks(t *testing.T) {
	userinfo := base64.URLEncoding.EncodeToString([]byte("aes-256-gcm:mysecretpassword"))
	raw := "ss://" + userinfo + "@1.2.3.4:8388#SSNode"
	cfg, err := ParseURI(raw)
	if err != nil {
		t.Fatalf("unexpected error parsing ss: %v", err)
	}

	if cfg.Protocol != domain.ProtocolShadowsocks {
		t.Errorf("expected shadowsocks, got %s", cfg.Protocol)
	}
	if cfg.Server != "1.2.3.4" {
		t.Errorf("expected server 1.2.3.4, got %s", cfg.Server)
	}
	if cfg.Port != 8388 {
		t.Errorf("expected port 8388, got %d", cfg.Port)
	}
	if cfg.UUID != "mysecretpassword" {
		t.Errorf("expected password mysecretpassword, got %s", cfg.UUID)
	}
}
