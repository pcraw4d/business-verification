package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/services/classification-service/internal/config"
	"kyb-platform/services/classification-service/internal/handlers"
	"kyb-platform/services/classification-service/internal/supabase"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Failed to sync logger: %v", err)
		}
	}()

	logger.Info("🚀 Starting Classification Service")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	logger.Info("✅ Configuration loaded successfully",
		zap.String("port", cfg.Server.Port),
		zap.String("supabase_url", cfg.Supabase.URL))

	// Initialize Supabase client
	supabaseClient, err := supabase.NewClient(&cfg.Supabase, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Supabase client", zap.Error(err))
	}

	// Initialize handlers
	classificationHandler := handlers.NewClassificationHandler(supabaseClient, logger, cfg)

	// Setup router
	router := mux.NewRouter()

	// Add middleware
	router.Use(loggingMiddleware(logger))
	router.Use(corsMiddleware())
	router.Use(rateLimitMiddleware())

	// Register routes
	router.HandleFunc("/health", classificationHandler.HandleHealth).Methods("GET")
	router.HandleFunc("/v1/classify", classificationHandler.HandleClassification).Methods("POST")
	router.HandleFunc("/classify", classificationHandler.HandleClassification).Methods("POST") // Alias for backward compatibility

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("🌐 Classification Service starting",
			zap.String("port", cfg.Server.Port),
			zap.String("host", cfg.Server.Host))

		logger.Info("🚀 Classification Service listening",
			zap.String("address", ":"+cfg.Server.Port),
			zap.String("port", cfg.Server.Port),
			zap.String("host", cfg.Server.Host))

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Classification Service server failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("🛑 Classification Service shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Fatal("Classification Service forced to shutdown", zap.Error(err))
	}

	logger.Info("✅ Classification Service exited gracefully")
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			logger.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("url", r.URL.String()),
				zap.Int("status", wrapped.statusCode),
				zap.Duration("duration", time.Since(start)),
				zap.String("user_agent", r.UserAgent()),
				zap.String("remote_addr", r.RemoteAddr))
		})
	}
}

// corsMiddleware adds CORS headers
func corsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// rateLimitMiddleware adds basic rate limiting
func rateLimitMiddleware() func(http.Handler) http.Handler {
	// Simple in-memory rate limiter
	requests := make(map[string][]time.Time)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := r.RemoteAddr
			now := time.Now()

			// Clean old requests (older than 1 minute)
			if clientRequests, exists := requests[clientIP]; exists {
				var validRequests []time.Time
				for _, reqTime := range clientRequests {
					if now.Sub(reqTime) < time.Minute {
						validRequests = append(validRequests, reqTime)
					}
				}
				requests[clientIP] = validRequests
			}

			// Check rate limit (100 requests per minute)
			if len(requests[clientIP]) >= 100 {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			// Add current request
			requests[clientIP] = append(requests[clientIP], now)

			next.ServeHTTP(w, r)
		})
	}
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
