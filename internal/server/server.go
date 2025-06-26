package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"teamcity-mcp/internal/cache"
	"teamcity-mcp/internal/config"
	"teamcity-mcp/internal/health"
	"teamcity-mcp/internal/mcp"
	"teamcity-mcp/internal/metrics"
	"teamcity-mcp/internal/teamcity"
)

// Server represents the MCP server
type Server struct {
	cfg      *config.Config
	logger   *zap.SugaredLogger
	tc       *teamcity.Client
	cache    *cache.Cache
	health   *health.Checker
	mcp      *mcp.Handler
	upgrader websocket.Upgrader
	mu       sync.RWMutex
}

// New creates a new MCP server instance
func New(cfg *config.Config, logger *zap.SugaredLogger) (*Server, error) {
	// Create TeamCity client
	tc, err := teamcity.NewClient(cfg.TeamCity, logger)
	if err != nil {
		return nil, fmt.Errorf("creating TeamCity client: %w", err)
	}

	// Create cache
	cache, err := cache.New(cfg.Cache)
	if err != nil {
		return nil, fmt.Errorf("creating cache: %w", err)
	}

	// Create health checker
	health := health.New(tc, logger)

	// Create MCP handler
	mcpHandler := mcp.NewHandler(tc, cache, logger)

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Configure properly for production
		},
	}

	return &Server{
		cfg:      cfg,
		logger:   logger,
		tc:       tc,
		cache:    cache,
		health:   health,
		mcp:      mcpHandler,
		upgrader: upgrader,
	}, nil
}

// Start starts the server with the specified transport
func (s *Server) Start(ctx context.Context, transport string) error {
	switch transport {
	case "http":
		return s.startHTTP(ctx)
	case "stdio":
		return s.startSTDIO(ctx)
	default:
		return fmt.Errorf("unsupported transport: %s", transport)
	}
}

// startHTTP starts the HTTP server
func (s *Server) startHTTP(ctx context.Context) error {
	mux := http.NewServeMux()

	// MCP endpoint
	mux.HandleFunc("/mcp", s.handleMCP)

	// Health endpoints
	mux.HandleFunc("/healthz", s.health.LivenessHandler)
	mux.HandleFunc("/readyz", s.health.ReadinessHandler)
	mux.HandleFunc("/metrics", s.handleMetrics)

	server := &http.Server{
		Addr:    s.cfg.Server.ListenAddr,
		Handler: s.authMiddleware(mux),
	}

	// Configure TLS if certificates are provided
	if s.cfg.Server.TLSCert != "" && s.cfg.Server.TLSKey != "" {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS13,
		}
		server.TLSConfig = tlsConfig
	}

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		s.logger.Info("Starting HTTP server", "addr", s.cfg.Server.ListenAddr)
		if s.cfg.Server.TLSCert != "" && s.cfg.Server.TLSKey != "" {
			errChan <- server.ListenAndServeTLS(s.cfg.Server.TLSCert, s.cfg.Server.TLSKey)
		} else {
			errChan <- server.ListenAndServe()
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		s.logger.Info("Shutting down HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}

// startSTDIO starts the STDIO transport
func (s *Server) startSTDIO(ctx context.Context) error {
	s.logger.Info("Starting STDIO transport")

	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			var req json.RawMessage
			if err := decoder.Decode(&req); err != nil {
				if err == io.EOF {
					return nil
				}
				s.logger.Error("Failed to decode request", "error", err)
				continue
			}

			resp, err := s.mcp.HandleRequest(ctx, req)
			if err != nil {
				s.logger.Error("Failed to handle request", "error", err)
				continue
			}

			if resp != nil {
				if err := encoder.Encode(resp); err != nil {
					s.logger.Error("Failed to encode response", "error", err)
				}
			}
		}
	}
}

// handleMCP handles MCP requests over HTTP/WebSocket
func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	if websocket.IsWebSocketUpgrade(r) {
		s.handleWebSocket(w, r)
		return
	}

	// Handle regular HTTP MCP request
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	resp, err := s.mcp.HandleRequest(r.Context(), req)
	if err != nil {
		s.logger.Error("Failed to handle MCP request", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("Failed to encode response", "error", err)
	}
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("Failed to upgrade to WebSocket", "error", err)
		return
	}
	defer conn.Close()

	metrics.ServerConnections.WithLabelValues("websocket").Inc()
	defer metrics.ServerConnections.WithLabelValues("websocket").Dec()

	s.logger.Info("WebSocket connection established")

	for {
		var req json.RawMessage
		if err := conn.ReadJSON(&req); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Error("WebSocket error", "error", err)
			}
			break
		}

		resp, err := s.mcp.HandleRequest(r.Context(), req)
		if err != nil {
			s.logger.Error("Failed to handle WebSocket request", "error", err)
			continue
		}

		if resp != nil {
			if err := conn.WriteJSON(resp); err != nil {
				s.logger.Error("Failed to write WebSocket response", "error", err)
				break
			}
		}
	}
}

// handleMetrics handles Prometheus metrics endpoint
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// This will be implemented by importing prometheus handler
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("# Metrics endpoint placeholder\n"))
}

// authMiddleware provides HMAC-based authentication
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health endpoints
		if strings.HasPrefix(r.URL.Path, "/health") || strings.HasPrefix(r.URL.Path, "/ready") || strings.HasPrefix(r.URL.Path, "/metrics") {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Bearer token required", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if !s.validateToken(token) {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// validateToken validates the HMAC token
func (s *Server) validateToken(token string) bool {
	// Simple HMAC validation - in production, implement proper token validation
	mac := hmac.New(sha256.New, []byte(s.cfg.Server.ServerSecret))
	mac.Write([]byte("teamcity-mcp"))
	expectedMAC := mac.Sum(nil)
	expectedToken := hex.EncodeToString(expectedMAC)

	return hmac.Equal([]byte(token), []byte(expectedToken))
}

// UpdateConfig updates the server configuration (for SIGHUP)
func (s *Server) UpdateConfig(cfg *config.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = cfg
	s.logger.Info("Configuration updated")
}
