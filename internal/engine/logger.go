package engine

import (
	"bufio"
	"bytes"
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

// GlobalEngineLog captures core subprocess logs across runs.
var GlobalEngineLog = NewRollingLogBuffer(500)
