package domain

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

type Protocol string

const (
	ProtocolVLESS       Protocol = "vless"
	ProtocolVMess       Protocol = "vmess"
	ProtocolTrojan      Protocol = "trojan"
	ProtocolShadowsocks Protocol = "shadowsocks"
	ProtocolSocks5      Protocol = "socks5"
	ProtocolHTTP        Protocol = "http"
)

// ProxyConfig represents a parsed, normalized proxy endpoint.
type ProxyConfig struct {
	ID          string            `json:"id"`
	RawURI      string            `json:"raw_uri"`
	Protocol    Protocol          `json:"protocol"`
	Server      string            `json:"server"`
	Port        int               `json:"port"`
	UUID        string            `json:"uuid"`
	Security    string            `json:"security,omitempty"` // tls, reality, none
	SNI         string            `json:"sni,omitempty"`
	Host        string            `json:"host,omitempty"`
	Path        string            `json:"path,omitempty"`
	Type        string            `json:"network,omitempty"` // tcp, ws, grpc
	Remark      string            `json:"remark,omitempty"`
	ExtraParams map[string]string `json:"extra_params,omitempty"`
}

// ComputeID calculates a deterministic deduplication hash for the proxy.
func (c *ProxyConfig) ComputeID() string {
	raw := fmt.Sprintf("%s://%s:%s@%s:%d?sec=%s&sni=%s&type=%s&path=%s",
		strings.ToLower(string(c.Protocol)),
		c.UUID,
		c.Server,
		c.Server,
		c.Port,
		c.Security,
		c.SNI,
		c.Type,
		c.Path,
	)
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h)
}

// Validate ensures all required attributes are valid.
func (c *ProxyConfig) Validate() error {
	if c.Server == "" {
		return fmt.Errorf("proxy server address cannot be empty")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port %d: must be in range 1-65535", c.Port)
	}
	if c.Protocol == "" {
		return fmt.Errorf("proxy protocol cannot be empty")
	}
	return nil
}

type ProbeStatus string

const (
	StatusPending ProbeStatus = "PENDING"
	StatusRunning ProbeStatus = "RUNNING"
	StatusPass    ProbeStatus = "PASS"
	StatusFail    ProbeStatus = "FAIL"
	StatusSkipped ProbeStatus = "SKIPPED"
)

// ProbeDefinition represents a configurable target service verification check.
type ProbeDefinition struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	TargetURL      string            `json:"target_url"`
	Method         string            `json:"method"`
	ExpectedStatus int               `json:"expected_status"`
	RequiredBody   string            `json:"required_body,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	TimeoutMs      int               `json:"timeout_ms"`
	Enabled        bool              `json:"enabled"`
}

type ErrorLayer string

const (
	LayerNone          ErrorLayer = "NONE"
	LayerLocalEngine   ErrorLayer = "LOCAL_ENGINE"   // Host OS socket, process crash, port bind refused
	LayerUpstreamProxy ErrorLayer = "UPSTREAM_PROXY" // Remote proxy node unreachable, TLS reset, dropped SYN
	LayerTargetService ErrorLayer = "TARGET_SERVICE" // Target platform interception (Cloudflare, Google, Geoblock)
)

type InterceptionType string

const (
	InterceptNone                 InterceptionType = "NONE"
	InterceptCloudflareChallenge  InterceptionType = "CLOUDFLARE_CHALLENGE"  // Turnstile, Under-Attack mode, 403/503
	InterceptGoogleUnusualTraffic InterceptionType = "GOOGLE_UNUSUAL_TRAFFIC" // Google /sorry/index captcha redirect
	InterceptGeoRestricted        InterceptionType = "GEO_RESTRICTED"         // Region not supported banner
	InterceptLocalEngineError     InterceptionType = "LOCAL_ENGINE_ERROR"     // connectex / socket refused
	InterceptBadRealityKey        InterceptionType = "BAD_REALITY_KEY"        // Reality pbk or short ID invalid
	InterceptBadCipher            InterceptionType = "BAD_CIPHER"             // Deprecated or unsupported Shadowsocks cipher
	InterceptBadUUID              InterceptionType = "BAD_UUID"               // Malformed user UUID
	InterceptUpstreamRefused      InterceptionType = "UPSTREAM_REFUSED"       // Target server rejected connection
	InterceptHandshakeFailed      InterceptionType = "HANDSHAKE_FAILED"       // TLS / WebSocket / gRPC handshake dropped
	InterceptUpstreamTimeout      InterceptionType = "UPSTREAM_TIMEOUT"       // Remote node timed out
	InterceptUnknown              InterceptionType = "UNKNOWN"
)

// DiagnosticInfo enriches ProbeResult with structured inspection telemetry.
type DiagnosticInfo struct {
	Layer            ErrorLayer       `json:"layer"`
	InterceptionType InterceptionType `json:"interception_type"`
	FinalURL         string           `json:"final_url,omitempty"`
	HTTPStatus       int              `json:"http_status,omitempty"`
	ServerHeader     string           `json:"server_header,omitempty"`
	RayID            string           `json:"ray_id,omitempty"`
	PageTitle        string           `json:"page_title,omitempty"`
	Summary          string           `json:"summary"`
	EngineLog        string           `json:"engine_log,omitempty"`
}

// ProbeResult represents the outcome of a single probe evaluation.
type ProbeResult struct {
	ProbeID      string          `json:"probe_id"`
	ProbeName    string          `json:"probe_name"`
	Status       ProbeStatus     `json:"status"`
	HTTPCode     int             `json:"http_code"`
	LatencyMs    int64           `json:"latency_ms"`
	ErrorMessage string          `json:"error_message,omitempty"`
	Diagnostic   *DiagnosticInfo `json:"diagnostic,omitempty"`
}

// IPEnrichment holds intelligence data returned from ipwho.is, ipleak.net, or Cloudflare.
type IPEnrichment struct {
	ExitIP         string  `json:"exit_ip"`
	CountryCode    string  `json:"country_code"`
	CountryName    string  `json:"country_name"`
	City           string  `json:"city,omitempty"`
	ISP            string  `json:"isp"`
	ASN            string  `json:"asn"`
	Org            string  `json:"org,omitempty"`
	ConnectionType string  `json:"connection_type,omitempty"` // "Residential", "Data Center", "Business", "Wireless", "Unknown"
	Latitude       float64 `json:"latitude,omitempty"`
	Longitude      float64 `json:"longitude,omitempty"`
	QueryTimeMs    int64   `json:"query_time_ms"`
	Success        bool    `json:"success"`
}

type EvaluationState string

const (
	StateQueued        EvaluationState = "QUEUED"
	StateStage1Testing EvaluationState = "STAGE1_HANDSHAKE"
	StateStage1Dead    EvaluationState = "STAGE1_DEAD"
	StateStage2Testing EvaluationState = "STAGE2_PROBING"
	StateComplete      EvaluationState = "COMPLETE"
	StateCancelled     EvaluationState = "CANCELLED"
)

// ProxyEvaluationReport contains aggregate telemetry for a tested proxy.
type ProxyEvaluationReport struct {
	Config            ProxyConfig            `json:"config"`
	State             EvaluationState        `json:"state"`
	Stage1LatencyMs   int64                  `json:"stage1_latency_ms"`
	Stage1Passed      bool                   `json:"stage1_passed"`
	ProbeResults      map[string]ProbeResult `json:"probe_results"`
	IPInfo            IPEnrichment           `json:"ip_info"`
	DownloadSpeedMbps float64                `json:"download_speed_mbps"`
	IsClean           bool                   `json:"is_clean"`
	EnrichedURI       string                 `json:"enriched_uri"`
	Timestamp         int64                  `json:"timestamp"`
}

// NewReport initializes a report in QUEUED state.
func NewReport(cfg ProxyConfig) *ProxyEvaluationReport {
	return &ProxyEvaluationReport{
		Config:       cfg,
		State:        StateQueued,
		ProbeResults: make(map[string]ProbeResult),
		Timestamp:    time.Now().Unix(),
	}
}

// Clone returns a thread-safe deep copy of the report, isolating maps and pointers.
func (r *ProxyEvaluationReport) Clone() *ProxyEvaluationReport {
	if r == nil {
		return nil
	}
	cp := *r
	if r.ProbeResults != nil {
		cp.ProbeResults = make(map[string]ProbeResult, len(r.ProbeResults))
		for k, v := range r.ProbeResults {
			if v.Diagnostic != nil {
				dCopy := *v.Diagnostic
				v.Diagnostic = &dCopy
			}
			cp.ProbeResults[k] = v
		}
	}
	return &cp
}

// ScanSessionStats tracks global batch progress counters.
type ScanSessionStats struct {
	SessionID     string `json:"session_id"`
	StartTime     int64  `json:"start_time"`
	TotalIngested int    `json:"total_ingested"`
	Deduplicated  int    `json:"deduplicated"`
	Stage1Alive   int64  `json:"stage1_alive"`
	Stage1Dead    int64  `json:"stage1_dead"`
	Stage2Clean   int64  `json:"stage2_clean"`
	Stage2Failed  int64  `json:"stage2_failed"`
	IsRunning     bool   `json:"is_running"`
}
