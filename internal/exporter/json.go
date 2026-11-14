package exporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"cips/internal/domain"
)

// SessionLog represents the complete persisted telemetry artifact.
type SessionLog struct {
	SessionID       string                            `json:"session_id"`
	GeneratedAt     string                            `json:"generated_at"`
	DurationSeconds float64                           `json:"duration_seconds"`
	Summary         domain.ScanSessionStats           `json:"summary"`
	Configurations  map[string]*domain.ProxyEvaluationReport `json:"configurations"`
}

// WriteSessionLog persists structured telemetry to a timestamped JSON file.
func WriteSessionLog(stats domain.ScanSessionStats, reports map[string]*domain.ProxyEvaluationReport, outputDir string) (string, error) {
	if outputDir == "" {
		outputDir = "."
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create log directory: %w", err)
	}

	filename := fmt.Sprintf("session-%s.json", time.Now().Format("20060102-150405"))
	filePath := filepath.Join(outputDir, filename)

	duration := float64(time.Now().Unix() - stats.StartTime)

	logObj := SessionLog{
		SessionID:       stats.SessionID,
		GeneratedAt:     time.Now().UTC().Format(time.RFC3339),
		DurationSeconds: duration,
		Summary:         stats,
		Configurations:  reports,
	}

	data, err := json.MarshalIndent(logObj, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal session telemetry: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write session log: %w", err)
	}

	return filePath, nil
}
