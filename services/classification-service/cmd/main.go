package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/internal/classification"
	classificationAdapters "kyb-platform/internal/classification/adapters"
	"kyb-platform/internal/classification/repository"
	"kyb-platform/services/classification-service/internal/errors"
	serviceAdapters "kyb-platform/services/classification-service/internal/adapters"
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

	// Create database client adapter for classification repository
	stdLogger := log.New(&zapLoggerAdapter{logger: logger}, "", 0)
	dbClient, err := serviceAdapters.CreateDatabaseClient(&cfg.Supabase, stdLogger)
	if err != nil {
		logger.Fatal("Failed to create database client adapter", zap.Error(err))
	}

	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := dbClient.Connect(ctx); err != nil {
		logger.Warn("Failed to connect to database, continuing anyway", zap.Error(err))
	}

	// Initialize adapters to break import cycle
	classificationAdapters.Init()
	logger.Info("✅ Classification adapters initialized")

	// Initialize classification repository
	keywordRepo := repository.NewSupabaseKeywordRepository(dbClient, stdLogger)

	// Initialize classification services
	industryDetector := classification.NewIndustryDetectionService(keywordRepo, stdLogger)
	codeGenerator := classification.NewClassificationCodeGenerator(keywordRepo, stdLogger)

	logger.Info("✅ Classification services initialized",
		zap.Bool("industry_detector", industryDetector != nil),
		zap.Bool("code_generator", codeGenerator != nil))

	// Initialize handlers
	classificationHandler := handlers.NewClassificationHandler(
		supabaseClient,
		logger,
		cfg,
		industryDetector,
		codeGenerator,
	)

	// Setup router
	router := mux.NewRouter()

	// Add middleware
	router.Use(securityHeadersMiddleware()) // Add security headers first
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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
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

// securityHeadersMiddleware adds security headers to HTTP responses
func securityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set security headers
			// HSTS (only for HTTPS)
			if r.TLS != nil {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			// X-Frame-Options
			w.Header().Set("X-Frame-Options", "DENY")

			// X-Content-Type-Options
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// X-XSS-Protection
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Referrer-Policy
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions-Policy
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			// Remove server information
			w.Header().Set("Server", "")

			// Additional security headers
			w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
			w.Header().Set("X-Download-Options", "noopen")
			w.Header().Set("X-DNS-Prefetch-Control", "off")

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
				errors.WriteError(w, r, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded", "Too many requests from this IP address")
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

// zapLoggerAdapter adapts zap.Logger to io.Writer for standard log.Logger
type zapLoggerAdapter struct {
	logger *zap.Logger
}

func (z *zapLoggerAdapter) Write(p []byte) (n int, err error) {
	z.logger.Info(strings.TrimSpace(string(p)))
	return len(p), nil
}
