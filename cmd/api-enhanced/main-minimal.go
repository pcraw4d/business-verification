package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// MinimalServer represents a minimal server for testing
type MinimalServer struct {
	server *http.Server
	logger *log.Logger
}

// NewMinimalServer creates a new minimal server
func NewMinimalServer(port string) *MinimalServer {
	logger := log.New(os.Stdout, "[minimal-server] ", log.LstdFlags)

	// Create HTTP server
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		healthStatus := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "3.1.0",
			"features": map[string]interface{}{
				"database_driven_classification": true,
				"supabase_integration":           true,
				"enhanced_keyword_matching":      true,
				"confidence_scoring":             true,
				"industry_detection":             true,
			},
		}
		json.NewEncoder(w).Encode(healthStatus)
	})

	// Serve static files from web directory
	fs := http.FileServer(http.Dir("web/"))
	mux.Handle("/", fs)

	server := &MinimalServer{
		logger: logger,
	}

	server.server = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	return server
}

// Start starts the server
func (s *MinimalServer) Start() error {
	s.logger.Printf("🚀 Starting Minimal Server on port %s", s.server.Addr)
	return s.server.ListenAndServe()
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := NewMinimalServer(port)

	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
