package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"cips/internal/domain"
)

// SSEBroadcaster manages active EventSource client connections.
type SSEBroadcaster struct {
	mu      sync.Mutex
	clients map[chan []byte]bool
}

// NewSSEBroadcaster creates a new broadcaster.
func NewSSEBroadcaster() *SSEBroadcaster {
	return &SSEBroadcaster{
		clients: make(map[chan []byte]bool),
	}
}

// ServeHTTP handles the SSE streaming connection.
func (b *SSEBroadcaster) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clientChan := make(chan []byte, 100)
	b.mu.Lock()
	b.clients[clientChan] = true
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.clients, clientChan)
		close(clientChan)
		b.mu.Unlock()
	}()

	// Send initial comment to establish stream
	_, _ = fmt.Fprintf(w, ": stream opened\n\n")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-clientChan:
			if !ok {
				return
			}
			_, _ = w.Write(msg)
			flusher.Flush()
		}
	}
}

// BroadcastEvent broadcasts an SSE event with eventName and JSON payload.
func (b *SSEBroadcaster) BroadcastEvent(eventName string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	msg := []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, data))

	b.mu.Lock()
	defer b.mu.Unlock()

	for ch := range b.clients {
		select {
		case ch <- msg:
		default:
			// Non-blocking drop if client buffer is full
		}
	}
}

// BroadcastItemUpdate emits an item_update event.
func (b *SSEBroadcaster) BroadcastItemUpdate(report *domain.ProxyEvaluationReport) {
	b.BroadcastEvent("item_update", report)
}

// BroadcastStats emits a stats event.
func (b *SSEBroadcaster) BroadcastStats(stats domain.ScanSessionStats) {
	b.BroadcastEvent("stats", stats)
}
