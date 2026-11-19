package engine

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"sync"
)

// RollingLogBuffer captures and stores the last N lines of subprocess stderr/stdout.
type RollingLogBuffer struct {
	mu       sync.RWMutex
	maxLines int
	lines    []string
}

// NewRollingLogBuffer creates a thread-safe ring buffer.
func NewRollingLogBuffer(maxLines int) *RollingLogBuffer {
	if maxLines <= 0 {
		maxLines = 500
	}
	return &RollingLogBuffer{
		maxLines: maxLines,
		lines:    make([]string, 0, maxLines),
	}
}

// Write implements io.Writer.
func (b *RollingLogBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	scanner := bufio.NewScanner(bytes.NewReader(p))
	for scanner.Scan() {
		line := scanner.Text()
		if len(b.lines) >= b.maxLines {
			b.lines = b.lines[1:]
		}
		b.lines = append(b.lines, line)
	}

	return len(p), nil
}

// Lines returns a snapshot of stored log lines.
func (b *RollingLogBuffer) Lines() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	out := make([]string, len(b.lines))
	copy(out, b.lines)
	return out
}

// GetLogsForPort searches stored lines for mentions of port (or general warnings/errors) and returns the last N lines.
func (b *RollingLogBuffer) GetLogsForPort(port int, maxLines int) []string {
	if maxLines <= 0 {
		maxLines = 3
	}
	b.mu.RLock()
	defer b.mu.RUnlock()

	portPattern := fmt.Sprintf(":%d", port)
	inboundPattern := fmt.Sprintf("inbound-%d", port)

	var matched []string
	for _, line := range b.lines {
		if strings.Contains(line, portPattern) || strings.Contains(line, inboundPattern) {
			matched = append(matched, line)
		}
	}

	// If no port-specific line found, capture the last general warning/error lines
	if len(matched) == 0 {
		for i := len(b.lines) - 1; i >= 0 && len(matched) < maxLines; i-- {
			line := b.lines[i]
			if strings.Contains(line, "failed") || strings.Contains(line, "error") || strings.Contains(line, "refused") || strings.Contains(line, "Warning") {
				matched = append([]string{line}, matched...)
			}
		}
	}

	if len(matched) > maxLines {
		matched = matched[len(matched)-maxLines:]
	}
	return matched
}

// GetLogsSnippetForPort returns a single formatted string of recent logs for port.
func (b *RollingLogBuffer) GetLogsSnippetForPort(port int) string {
	lines := b.GetLogsForPort(port, 3)
	return strings.Join(lines, "\n")
}

// GlobalEngineLog captures core subprocess logs across runs.
var GlobalEngineLog = NewRollingLogBuffer(500)
