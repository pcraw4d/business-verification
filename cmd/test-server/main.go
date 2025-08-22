package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pcraw4d/business-verification/internal/middleware"
)

func main() {
	// Create a simple test router with mock handlers
	router := http.NewServeMux()

	// Add v3 API routes with simple mock handlers
	router.HandleFunc("/api/v3/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "data": {"total_requests": 5000, "success_rate": 99.5}, "meta": {"response_time": "50ms", "timestamp": "2024-01-15T10:30:00Z"}}`))
	})

	router.HandleFunc("/api/v3/alerts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "data": [{"id": "alert_1", "name": "Test Alert", "severity": "warning"}], "meta": {"response_time": "30ms", "timestamp": "2024-01-15T10:30:00Z"}}`))
	})

	router.HandleFunc("/api/v3/performance/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "data": {"response_time": 120, "throughput": 200}, "meta": {"response_time": "40ms", "timestamp": "2024-01-15T10:30:00Z"}}`))
	})

	router.HandleFunc("/api/v3/errors", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "data": [{"id": "error_1", "error_type": "test", "severity": "info"}], "meta": {"response_time": "25ms", "timestamp": "2024-01-15T10:30:00Z"}}`))
	})

	router.HandleFunc("/api/v3/analytics/business/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "data": {"active_users": 150, "total_verifications": 2500}, "meta": {"response_time": "35ms", "timestamp": "2024-01-15T10:30:00Z"}}`))
	})

	router.HandleFunc("/api/v3/integrations/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "data": {"status": "healthy", "integrations": 3}, "meta": {"response_time": "20ms", "timestamp": "2024-01-15T10:30:00Z"}}`))
	})

	// Add health check endpoint (no auth required)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "timestamp": "2024-01-15T10:30:00Z"}`))
	})

	// Add rate limit stats endpoint
	router.HandleFunc("/api/v3/admin/rate-limit-stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "data": {"total_clients": 5, "config": {"requests_per_minute": 1000}}, "meta": {"response_time": "15ms", "timestamp": "2024-01-15T10:30:00Z"}}`))
	})

	// Configure authentication
	authConfig := middleware.AuthConfig{
		JWTSecret:    "test-jwt-secret-key",
		APIKeyHeader: "X-API-Key",
		APIKeys: map[string]string{
			"test-api-key-123": "test-user-1",
			"test-api-key-456": "test-user-2",
		},
		RequireAuth: true,
		ExemptPaths: []string{
			"/health",
			"/api/v3/admin/rate-limit-stats",
		},
	}

	// Configure rate limiting
	rateLimitConfig := middleware.RateLimitConfig{
		RequestsPerMinute: 1000,
		RequestsPerHour:   10000,
		BurstSize:         100,
		EnableRateLimit:   true,
		ExemptPaths: []string{
			"/health",
			"/api/v3/admin/rate-limit-stats",
		},
	}

	// Create rate limit store
	rateLimitStore := middleware.NewRateLimitStore(rateLimitConfig)
	rateLimitStore.StartCleanup()

	// Apply middleware
	handler := middleware.AuthMiddleware(authConfig)(
		middleware.RateLimitMiddleware(rateLimitStore)(router),
	)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		println("Starting enhanced test server on :8080")
		println("Available endpoints:")
		println("  GET  /api/v3/dashboard")
		println("  GET  /api/v3/alerts")
		println("  GET  /api/v3/performance/metrics")
		println("  GET  /api/v3/errors")
		println("  GET  /api/v3/analytics/business/metrics")
		println("  GET  /api/v3/integrations/status")
		println("  GET  /health (no auth required)")
		println("  GET  /api/v3/admin/rate-limit-stats (no auth required)")
		println("")
		println("Authentication:")
		println("  - JWT Bearer token: Authorization: Bearer <jwt-token>")
		println("  - API Key: Authorization: ApiKey test-api-key-123")
		println("")
		println("Rate Limiting:")
		println("  - 1000 requests per minute per client")
		println("  - 10000 requests per hour per client")
		println("")
		println("Use Ctrl+C to stop the server")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			println("Server error:", err.Error())
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		println("Server forced to shutdown:", err.Error())
	}

	println("Server exited")
}
