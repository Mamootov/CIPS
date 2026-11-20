package server

import (
	"context"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"time"

	"cips/internal/pipeline"
)

// Server wraps the local HTTP dashboard server.
type Server struct {
	httpServer  *http.Server
	listener    net.Listener
	port        int
	broadcaster *SSEBroadcaster
	context     *ServerContext
}

// NewServer initializes the HTTP router and services.
func NewServer(orch *pipeline.Orchestrator, webFS fs.FS) *Server {
	bc := NewSSEBroadcaster()
	sc := NewServerContext(orch, bc)

	mux := http.NewServeMux()

	// API Routes
	mux.HandleFunc("/api/origin", sc.HandleOrigin)
	mux.HandleFunc("/api/ingest", sc.HandleIngest)
	mux.HandleFunc("/api/scan/start", sc.HandleScanStart)
	mux.HandleFunc("/api/scan/cancel", sc.HandleScanCancel)
	mux.HandleFunc("/api/session", sc.HandleSessionState)
	mux.HandleFunc("/api/reset", sc.HandleReset)
	mux.HandleFunc("/api/export", sc.HandleExport)
	mux.Handle("/api/stream", bc)

	// Static Web Assets
	if webFS != nil {
		fileServer := http.FileServer(http.FS(webFS))
		mux.Handle("/", fileServer)
	}

	return &Server{
		broadcaster: bc,
		context:     sc,
		httpServer: &http.Server{
			Handler:      mux,
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 0, // Disabled for SSE streaming
		},
	}
}

// Broadcaster returns the SSEBroadcaster instance.
func (s *Server) Broadcaster() *SSEBroadcaster {
	return s.broadcaster
}

// Context returns the ServerContext instance.
func (s *Server) Context() *ServerContext {
	return s.context
}

// ListenWithFallback attempts to bind defaultPort, incrementing if occupied.
func (s *Server) ListenWithFallback(defaultPort int) (int, error) {
	for p := defaultPort; p < defaultPort+20; p++ {
		addr := fmt.Sprintf("127.0.0.1:%d", p)
		l, err := net.Listen("tcp", addr)
		if err == nil {
			s.listener = l
			s.port = p
			return p, nil
		}
	}
	return 0, fmt.Errorf("unable to bind any port between %d and %d", defaultPort, defaultPort+20)
}

// Start begins serving requests on the bound listener.
func (s *Server) Start() error {
	if s.listener == nil {
		return fmt.Errorf("listener is not initialized; call ListenWithFallback first")
	}
	return s.httpServer.Serve(s.listener)
}

// Shutdown gracefully terminates the web server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Port returns the active bound port.
func (s *Server) Port() int {
	return s.port
}
