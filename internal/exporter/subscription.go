package exporter

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"cips/internal/domain"
)

// GenerateCleanSubscription creates a subscription payload with smart tags.
func GenerateCleanSubscription(reports []*domain.ProxyEvaluationReport, requiredProbeID string) string {
	var sb strings.Builder

	for _, rep := range reports {
		if !rep.IsClean {
			continue
		}

		if requiredProbeID != "" {
			res, ok := rep.ProbeResults[requiredProbeID]
			if !ok || res.Status != domain.StatusPass {
				continue
			}
		}

		enrichedURI := EnrichProxyRemark(rep)
		rep.EnrichedURI = enrichedURI
		sb.WriteString(enrichedURI)
		sb.WriteString("\n")
	}

	return sb.String()
}

// EnrichProxyRemark rewrites remark with smart tags: [CC] [ISP] [PASS_TAGS] [PINGms] - Original.
func EnrichProxyRemark(r *domain.ProxyEvaluationReport) string {
	cfg := r.Config

	var tagParts []string

	// Country code
	if r.IPInfo.CountryCode != "" {
		tagParts = append(tagParts, fmt.Sprintf("[%s]", r.IPInfo.CountryCode))
	}

	// ISP
	if r.IPInfo.ISP != "" {
		tagParts = append(tagParts, fmt.Sprintf("[%s]", r.IPInfo.ISP))
	}

	// Connection Type (Residential vs Datacenter)
	if r.IPInfo.ConnectionType == "Residential" {
		tagParts = append(tagParts, "[Res]")
	} else if r.IPInfo.ConnectionType == "Data Center" {
		tagParts = append(tagParts, "[DC]")
	}

	// Download Speed
	if r.DownloadSpeedMbps > 0 {
		tagParts = append(tagParts, fmt.Sprintf("[%.1fMbps]", r.DownloadSpeedMbps))
	}

	// Pass tags
	var passedServices []string
	for id, res := range r.ProbeResults {
		if res.Status == domain.StatusPass {
			switch id {
			case "gemini":
				passedServices = append(passedServices, "Gemini")
			case "aistudio":
				passedServices = append(passedServices, "Studio")
			case "chatgpt":
				passedServices = append(passedServices, "ChatGPT")
			case "claude":
				passedServices = append(passedServices, "Claude")
			case "ipleak", "speedtest":
				// Skip auxiliary probes
			default:
				passedServices = append(passedServices, id)
			}
		}
	}
	if len(passedServices) > 0 {
		tagParts = append(tagParts, fmt.Sprintf("[%s]", strings.Join(passedServices, "+")))
	}

	// Ping latency
	if r.Stage1LatencyMs > 0 {
		tagParts = append(tagParts, fmt.Sprintf("[%dms]", r.Stage1LatencyMs))
	}

	tagPrefix := strings.Join(tagParts, " ")
	newRemark := tagPrefix
	if cfg.Remark != "" {
		newRemark = fmt.Sprintf("%s - %s", tagPrefix, cfg.Remark)
	}

	// Protocol-specific URI rewrite
	switch cfg.Protocol {
	case domain.ProtocolVMess:
		return rewriteVMessRemark(cfg.RawURI, newRemark)
	default:
		return rewriteURLFragment(cfg.RawURI, newRemark)
	}
}

func rewriteURLFragment(rawURI, newRemark string) string {
	encoded := strings.ReplaceAll(url.QueryEscape(newRemark), "+", "%20")
	if idx := strings.Index(rawURI, "#"); idx != -1 {
		return rawURI[:idx] + "#" + encoded
	}
	return rawURI + "#" + encoded
}

func rewriteVMessRemark(rawURI, newRemark string) string {
	b64Part := strings.TrimPrefix(rawURI, "vmess://")
	b64Part = strings.TrimPrefix(b64Part, "VMESS://")

	decoded, err := base64.StdEncoding.DecodeString(b64Part)
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(b64Part)
	}
	if err != nil {
		return rawURI
	}

	var data map[string]interface{}
	if err := json.Unmarshal(decoded, &data); err != nil {
		return rawURI
	}

	data["ps"] = newRemark

	newJSON, err := json.Marshal(data)
	if err != nil {
		return rawURI
	}

	return "vmess://" + base64.StdEncoding.EncodeToString(newJSON)
}
