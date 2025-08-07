package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// Server represents the main API server
type Server struct {
	config *config.Config
	logger *observability.Logger
	server *http.Server
}

// NewServer creates a new API server instance
func NewServer(cfg *config.Config, logger *observability.Logger) *Server {
	return &Server{
		config: cfg,
		logger: logger,
	}
}

// setupRoutes configures all API routes using Go 1.22's new ServeMux features
func (s *Server) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", s.healthHandler)

	// API versioning with v1 prefix
	mux.HandleFunc("GET /v1/status", s.statusHandler)
	mux.HandleFunc("GET /v1/metrics", s.metricsHandler)

	// API documentation
	mux.HandleFunc("GET /docs", s.docsHandler)
	mux.HandleFunc("GET /docs/", s.docsHandler)

	// API endpoints (to be implemented)
	mux.HandleFunc("POST /v1/auth/register", s.notImplementedHandler)
	mux.HandleFunc("POST /v1/auth/login", s.notImplementedHandler)
	mux.HandleFunc("POST /v1/classify", s.notImplementedHandler)
	mux.HandleFunc("POST /v1/classify/batch", s.notImplementedHandler)

	// Catch-all for undefined routes
	mux.HandleFunc("GET /", s.notFoundHandler)
	mux.HandleFunc("POST /", s.notFoundHandler)
	mux.HandleFunc("PUT /", s.notFoundHandler)
	mux.HandleFunc("DELETE /", s.notFoundHandler)

	return mux
}

// setupMiddleware configures the middleware stack
func (s *Server) setupMiddleware(handler http.Handler) http.Handler {
	// Apply middleware in order (last middleware is applied first)
	handler = s.securityHeadersMiddleware(handler)
	handler = s.corsMiddleware(handler)
	handler = s.requestLoggingMiddleware(handler)
	handler = s.requestIDMiddleware(handler)
	handler = s.recoveryMiddleware(handler)

	return handler
}

// healthHandler handles health check requests
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.WithComponent("api").LogHealthCheck("api", "healthy", map[string]interface{}{
		"endpoint": "/health",
		"method":   r.Method,
		"user_agent": r.UserAgent(),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
}

// statusHandler handles API status requests
func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	s.logger.WithComponent("api").LogAPIRequest(r.Context(), "GET", r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
		"status":"operational",
		"version":"1.0.0",
		"timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"
	}`))
}

// metricsHandler handles metrics requests
func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	s.logger.WithComponent("api").LogAPIRequest(r.Context(), "GET", r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))

	// TODO: Implement actual metrics collection
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("# KYB Tool Metrics\n# This endpoint will be implemented with Prometheus metrics\n"))
}

// docsHandler handles API documentation requests
func (s *Server) docsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	s.logger.WithComponent("api").LogAPIRequest(r.Context(), "GET", r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>KYB Tool API Documentation</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .endpoint { margin: 20px 0; padding: 10px; border-left: 4px solid #007cba; }
        .method { font-weight: bold; color: #007cba; }
    </style>
</head>
<body>
    <h1>KYB Tool API Documentation</h1>
    <p>Welcome to the KYB Tool API. This documentation will be enhanced with OpenAPI/Swagger specification.</p>
    
    <h2>Available Endpoints</h2>
    
    <div class="endpoint">
        <span class="method">GET</span> /health - Health check endpoint
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> /v1/status - API status information
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> /v1/metrics - Prometheus metrics endpoint
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> /v1/auth/register - User registration (coming soon)
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> /v1/auth/login - User login (coming soon)
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> /v1/classify - Business classification (coming soon)
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> /v1/classify/batch - Batch business classification (coming soon)
    </div>
</body>
</html>`))
}

// notFoundHandler handles undefined routes
func (s *Server) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	s.logger.WithComponent("api").LogAPIRequest(r.Context(), r.Method, r.URL.Path, r.UserAgent(), http.StatusNotFound, time.Since(start))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"error":"not_found","message":"The requested endpoint does not exist","path":"` + r.URL.Path + `"}`))
}

// notImplementedHandler handles endpoints that are not yet implemented
func (s *Server) notImplementedHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	s.logger.WithComponent("api").LogAPIRequest(r.Context(), r.Method, r.URL.Path, r.UserAgent(), http.StatusNotImplemented, time.Since(start))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not_implemented","message":"This endpoint is not yet implemented","path":"` + r.URL.Path + `"}`))
}

// securityHeadersMiddleware adds security headers to responses
func (s *Server) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		
		// Remove server information
		w.Header().Set("Server", "KYB-Tool")
		
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware handles Cross-Origin Resource Sharing
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Set CORS headers for actual requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")

		next.ServeHTTP(w, r)
	})
}

// requestLoggingMiddleware logs all incoming requests
func (s *Server) requestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a custom response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(rw, r)
		
		duration := time.Since(start)
		
		s.logger.WithComponent("api").LogAPIRequest(r.Context(), r.Method, r.URL.Path, r.UserAgent(), rw.statusCode, duration)
	})
}

// requestIDMiddleware adds request ID to context and headers
func (s *Server) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract request ID from header or generate new one
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = observability.GenerateRequestID()
		}
		
		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)
		
		// Add request ID to context
		ctx := context.WithValue(r.Context(), observability.RequestIDKey, requestID)
		r = r.WithContext(ctx)
		
		next.ServeHTTP(w, r)
	})
}

// recoveryMiddleware recovers from panics
func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.logger.WithComponent("api").WithError(fmt.Errorf("panic: %v", err)).Error("panic recovered", "method", r.Method, "path", r.URL.Path)
				
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"internal_server_error","message":"An unexpected error occurred"}`))
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Setup routes
	mux := s.setupRoutes()
	
	// Setup middleware
	handler := s.setupMiddleware(mux)
	
	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
		IdleTimeout:  s.config.Server.IdleTimeout,
	}
	
	s.logger.WithComponent("api").LogStartup("1.0.0", "dev", time.Now().Format(time.RFC3339))
	
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.WithComponent("api").LogShutdown("graceful_shutdown")
	
	return s.server.Shutdown(ctx)
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := observability.NewLogger(&cfg.Observability)
	
	// Create server
	server := NewServer(cfg, logger)
	
	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			logger.WithComponent("api").WithError(err).LogShutdown("server_start_failed")
			log.Fatalf("Server failed to start: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		logger.WithComponent("api").WithError(err).LogShutdown("server_shutdown_failed")
		log.Fatalf("Server shutdown failed: %v", err)
	}
	
	logger.WithComponent("api").LogShutdown("server_shutdown_complete")
}
