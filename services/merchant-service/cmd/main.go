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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"kyb-platform/services/merchant-service/internal/config"
	"kyb-platform/services/merchant-service/internal/errors"
	"kyb-platform/services/merchant-service/internal/handlers"
	"kyb-platform/services/merchant-service/internal/supabase"
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

	logger.Info("🚀 Starting Merchant Service")

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
	logger.Info("✅ Supabase client initialized")

	// Initialize handlers
	merchantHandler := handlers.NewMerchantHandler(supabaseClient, logger, cfg)

	// Setup router
	router := mux.NewRouter()

	// Add middleware
	router.Use(securityHeadersMiddleware()) // Add security headers first
	router.Use(loggingMiddleware(logger))
	router.Use(corsMiddleware())
	router.Use(rateLimitMiddleware())

	// Register routes
	router.HandleFunc("/health", merchantHandler.HandleHealth).Methods("GET")

	// Prometheus metrics endpoint
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Merchant management routes
	// IMPORTANT: Register more specific routes BEFORE less specific ones
	// This ensures /api/v1/merchants/{id}/analytics matches before /api/v1/merchants/{id}

	// Merchant-specific sub-routes (must be registered BEFORE /merchants/{id})
	// Merchant routes - CRITICAL: Route registration order matters!
	// Specific routes MUST be registered before less specific routes.
	// Routes are registered in order from most specific to least specific.
	//
	// ORDER MATTERS:
	// 1. Most specific sub-routes first (merchants/{id}/analytics, etc.)
	//    These must come before /merchants/{id} to avoid route conflicts
	router.HandleFunc("/api/v1/merchants/{id}/analytics", merchantHandler.HandleMerchantSpecificAnalytics).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/merchants/{id}/website-analysis", merchantHandler.HandleMerchantWebsiteAnalysis).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/merchants/{id}/risk-score", merchantHandler.HandleMerchantRiskScore).Methods("GET", "OPTIONS")

	// 2. General merchant endpoints (must be registered BEFORE /merchants/{id} to avoid conflicts)
	//    Routes like /merchants/analytics must come before /merchants/{id} because
	//    the router might interpret "analytics" as an {id} parameter
	router.HandleFunc("/api/v1/merchants/analytics", merchantHandler.HandleMerchantAnalytics).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/merchants/statistics", merchantHandler.HandleMerchantStatistics).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/merchants/search", merchantHandler.HandleMerchantSearch).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/merchants/portfolio-types", merchantHandler.HandleMerchantPortfolioTypes).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/merchants/risk-levels", merchantHandler.HandleMerchantRiskLevels).Methods("GET", "OPTIONS")

	// 3. Base merchant routes (registered last among /api/v1/merchants routes)
	router.HandleFunc("/api/v1/merchants", merchantHandler.HandleCreateMerchant).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/merchants", merchantHandler.HandleListMerchants).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/merchants/{id}", merchantHandler.HandleGetMerchant).Methods("GET", "OPTIONS")

	// Alias routes for backward compatibility
	router.HandleFunc("/merchants", merchantHandler.HandleCreateMerchant).Methods("POST")
	router.HandleFunc("/merchants", merchantHandler.HandleListMerchants).Methods("GET")
	router.HandleFunc("/merchants/{id}", merchantHandler.HandleGetMerchant).Methods("GET")

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
		logger.Info("🌐 Merchant Service starting",
			zap.String("port", cfg.Server.Port),
			zap.String("host", cfg.Server.Host))

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Merchant Service server failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("🛑 Merchant Service shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Fatal("Merchant Service forced to shutdown", zap.Error(err))
	}

	logger.Info("✅ Merchant Service exited gracefully")
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
// NOTE: This service is typically behind an API Gateway that handles CORS.
// CORS headers should only be set for direct requests (when Origin header is present).
// For internal requests from API Gateway (no Origin), skip CORS to avoid duplicate headers.
func corsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the origin from the request
			origin := r.Header.Get("Origin")

			// CRITICAL: If there's no Origin header, this is likely an internal request
			// from the API Gateway. Skip CORS headers entirely - the API Gateway will handle CORS.
			// This prevents duplicate CORS headers when proxying through the API Gateway.
			if origin == "" {
				// Internal request - skip CORS, let API Gateway handle it
				next.ServeHTTP(w, r)
				return
			}

			// External request with Origin header - set CORS headers
			// CRITICAL: Remove any existing CORS headers first to avoid duplicates
			w.Header().Del("Access-Control-Allow-Origin")
			w.Header().Del("Access-Control-Allow-Methods")
			w.Header().Del("Access-Control-Allow-Headers")
			w.Header().Del("Access-Control-Allow-Credentials")
			w.Header().Del("Access-Control-Max-Age")

			// Set CORS headers - use the requesting origin
			// Note: When using credentials, we must use a specific origin, not "*"
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight requests
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
