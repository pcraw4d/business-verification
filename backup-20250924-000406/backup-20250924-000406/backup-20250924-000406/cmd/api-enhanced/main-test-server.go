package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// TestServer represents a test server that doesn't require real database connections
type TestServer struct {
	server *http.Server
	logger *log.Logger
}

// NewTestServer creates a new test server
func NewTestServer(port string) *TestServer {
	logger := log.New(os.Stdout, "[test-server] ", log.LstdFlags)

	// Create HTTP server
	mux := http.NewServeMux()

	// Add test endpoints
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	mux.HandleFunc("/classify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			BusinessName string `json:"business_name"`
			Description  string `json:"description"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Create a mock classification result
		result := map[string]interface{}{
			"business_name": req.BusinessName,
			"description":   req.Description,
			"classification": map[string]interface{}{
				"industry":   "Technology",
				"confidence": 0.85,
				"codes": []map[string]interface{}{
					{
						"type":        "NAICS",
						"code":        "541511",
						"description": "Custom Computer Programming Services",
					},
				},
			},
			"timestamp": time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &TestServer{
		server: server,
		logger: logger,
	}
}

// Start starts the test server
func (s *TestServer) Start() error {
	s.logger.Printf("🚀 Starting test server on port %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop stops the test server
func (s *TestServer) Stop(ctx context.Context) error {
	s.logger.Printf("🛑 Stopping test server")
	return s.server.Shutdown(ctx)
}

func main() {
	port := "8080"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	server := NewTestServer(port)

	// Handle graceful shutdown
	go func() {
		<-time.After(30 * time.Second) // Auto-shutdown after 30 seconds for testing
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
