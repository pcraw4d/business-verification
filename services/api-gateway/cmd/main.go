package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"kyb-platform/services/api-gateway/internal/config"
	"kyb-platform/services/api-gateway/internal/handlers"
	"kyb-platform/services/api-gateway/internal/middleware"
	"kyb-platform/services/api-gateway/internal/supabase"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("🚀 Starting KYB API Gateway Service v1.0.20 - Added dashboard, compliance, and sessions routes")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Log configuration for debugging
	logger.Info("🔧 Configuration loaded",
		zap.String("port", cfg.Server.Port),
		zap.String("environment", cfg.Environment),
		zap.String("classification_url", cfg.Services.ClassificationURL),
		zap.String("merchant_url", cfg.Services.MerchantURL),
		zap.String("frontend_url", cfg.Services.FrontendURL),
		zap.String("bi_service_url", cfg.Services.BIServiceURL),
		zap.String("risk_assessment_url", cfg.Services.RiskAssessmentURL))

	// Initialize Supabase client
	logger.Info("🔧 Initializing Supabase client",
		zap.String("url", cfg.Supabase.URL),
		zap.String("api_key_length", fmt.Sprintf("%d", len(cfg.Supabase.APIKey))))

	supabaseClient, err := supabase.NewClient(&cfg.Supabase, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Supabase client", zap.Error(err))
	}
	logger.Info("✅ Supabase client initialized successfully")

	// Initialize handlers
	gatewayHandler := handlers.NewGatewayHandler(supabaseClient, logger, cfg)

	// Setup router
	router := mux.NewRouter()

	// Apply middleware - CORS must be first to handle preflight requests
	router.Use(middleware.CORS(cfg.CORS))  // Enable CORS middleware (FIRST)
	router.Use(middleware.SecurityHeaders) // Add security headers
	router.Use(middleware.Logging(logger))
	router.Use(middleware.RateLimit(cfg.RateLimit))
	router.Use(middleware.Authentication(supabaseClient, logger))

	// Health check endpoint
	router.HandleFunc("/health", gatewayHandler.HealthCheck).Methods("GET")

	// Prometheus metrics endpoint
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Root endpoint for debugging
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service": "api-gateway",
			"version": "1.0.20",
			"status":  "running",
			"message": "KYB API Gateway is running",
			"endpoints": map[string]string{
				"health":                "/health",
				"classify":              "/api/v1/classify",
				"merchants":             "/api/v1/merchants",
				"risk_assessment":       "/api/v1/risk",
				"classification_health": "/api/v1/classification/health",
				"merchant_health":       "/api/v1/merchant/health",
				"risk_health":           "/api/v1/risk/health",
			},
		})
	}).Methods("GET")

	// API Gateway routes
	// CRITICAL: Route registration order matters!
	// Specific routes MUST be registered before PathPrefix catch-all routes.
	// PathPrefix routes will shadow specific routes if registered first.
	api := router.PathPrefix("/api/v1").Subrouter()
	// NOTE: CORS middleware is already applied to parent router (line 68)
	// Do NOT apply CORS again here to avoid duplicate headers

	// API v3 routes for enhanced endpoints
	apiV3 := router.PathPrefix("/api/v3").Subrouter()
	// NOTE: CORS middleware is already applied to parent router (line 68)
	// Do NOT apply CORS again here to avoid duplicate headers
	apiV3.Use(middleware.SecurityHeaders)
	apiV3.Use(middleware.Logging(logger))
	apiV3.Use(middleware.RateLimit(cfg.RateLimit))
	apiV3.Use(middleware.Authentication(supabaseClient, logger))
	apiV3.HandleFunc("/dashboard/metrics", gatewayHandler.ProxyToDashboardMetricsV3).Methods("GET", "OPTIONS")

	// Classification routes
	api.HandleFunc("/classify", gatewayHandler.ProxyToClassification).Methods("POST")

	// Merchant routes - CORS handled by middleware
	// ORDER MATTERS: Sub-routes with {id} must be registered before base routes
	// 1. Most specific sub-routes first (merchants/{id}/analytics, etc.)
	api.HandleFunc("/merchants/{id}/analytics", gatewayHandler.ProxyToMerchants).Methods("GET", "OPTIONS")
	api.HandleFunc("/merchants/{id}/website-analysis", gatewayHandler.ProxyToMerchants).Methods("GET", "OPTIONS")
	api.HandleFunc("/merchants/{id}/risk-score", gatewayHandler.ProxyToMerchants).Methods("GET", "OPTIONS")

	// 2. General merchant endpoints (search, analytics, statistics) before /merchants/{id}
	api.HandleFunc("/merchants/search", gatewayHandler.ProxyToMerchants).Methods("POST", "OPTIONS")
	api.HandleFunc("/merchants/analytics", gatewayHandler.ProxyToMerchants).Methods("GET", "OPTIONS")
	api.HandleFunc("/merchants/statistics", gatewayHandler.ProxyToMerchants).Methods("GET", "OPTIONS")

	// 3. Base merchant routes
	api.HandleFunc("/merchants/{id}", gatewayHandler.ProxyToMerchants).Methods("GET", "PUT", "DELETE", "OPTIONS")
	api.HandleFunc("/merchants", gatewayHandler.ProxyToMerchants).Methods("GET", "POST", "OPTIONS")

	// 4. PathPrefix catch-all LAST (will match any remaining /merchants/* routes)
	// This must be registered after all specific routes to avoid shadowing them
	api.PathPrefix("/merchants").HandlerFunc(gatewayHandler.ProxyToMerchants)

	// Health check routes for backend services
	api.HandleFunc("/classification/health", gatewayHandler.ProxyToClassificationHealth).Methods("GET")
	api.HandleFunc("/merchant/health", gatewayHandler.ProxyToMerchantHealth).Methods("GET")
	api.HandleFunc("/risk/health", gatewayHandler.ProxyToRiskAssessmentHealth).Methods("GET")

	// Dashboard routes - v1 deprecated, use v3 instead
	// api.HandleFunc("/dashboard/metrics", gatewayHandler.ProxyToDashboardMetricsV1).Methods("GET", "OPTIONS")

	// Compliance routes - CORS handled by middleware (register before PathPrefix routes)
	api.HandleFunc("/compliance/status", gatewayHandler.ProxyToComplianceStatus).Methods("GET", "OPTIONS")

	// Session routes - CORS handled by middleware
	// ORDER MATTERS: Specific routes must be registered before PathPrefix
	// 1. Specific session sub-routes first
	api.HandleFunc("/sessions/current", gatewayHandler.ProxyToSessions).Methods("GET", "OPTIONS")
	api.HandleFunc("/sessions/metrics", gatewayHandler.ProxyToSessions).Methods("GET", "OPTIONS")
	api.HandleFunc("/sessions/activity", gatewayHandler.ProxyToSessions).Methods("GET", "OPTIONS")
	api.HandleFunc("/sessions/status", gatewayHandler.ProxyToSessions).Methods("GET", "OPTIONS")
	// 2. Base sessions route
	api.HandleFunc("/sessions", gatewayHandler.ProxyToSessions).Methods("GET", "POST", "DELETE", "OPTIONS")
	// 3. PathPrefix catch-all LAST
	api.PathPrefix("/sessions").HandlerFunc(gatewayHandler.ProxyToSessions)

	// Analytics routes - CORS handled by middleware
	// ORDER MATTERS: Analytics routes must be registered before /risk PathPrefix
	// Analytics routes are handled by Risk Assessment service
	api.HandleFunc("/analytics/trends", gatewayHandler.ProxyToRiskAssessment).Methods("GET", "OPTIONS")
	api.HandleFunc("/analytics/insights", gatewayHandler.ProxyToRiskAssessment).Methods("GET", "OPTIONS")

	// Risk Assessment routes - CORS handled by middleware
	// ORDER MATTERS: Specific routes with path transformation must be registered before PathPrefix
	// 1. Routes requiring path transformation (registered first)
	api.HandleFunc("/risk/assess", gatewayHandler.ProxyToRiskAssessment).Methods("POST", "OPTIONS")
	api.HandleFunc("/risk/benchmarks", gatewayHandler.ProxyToRiskAssessment).Methods("GET", "OPTIONS")
	api.HandleFunc("/risk/predictions/{merchant_id}", gatewayHandler.ProxyToRiskAssessment).Methods("GET", "OPTIONS")
	api.HandleFunc("/risk/indicators/{id}", gatewayHandler.ProxyToRiskAssessment).Methods("GET", "OPTIONS")
	// 2. PathPrefix catch-all LAST (for remaining /risk/* routes)
	api.PathPrefix("/risk").HandlerFunc(gatewayHandler.ProxyToRiskAssessment)

	// Business Intelligence routes - CORS handled by middleware
	// ORDER MATTERS: Specific route before PathPrefix
	// 1. Specific BI route
	api.HandleFunc("/bi/analyze", gatewayHandler.ProxyToBI).Methods("POST", "OPTIONS")
	// 2. PathPrefix catch-all LAST
	api.PathPrefix("/bi").HandlerFunc(gatewayHandler.ProxyToBI)

	// Authentication routes - CORS handled by middleware
	// NOTE: These routes are registered after PathPrefix routes, but specific routes
	// should still match correctly. If 404 occurs, verify code is deployed to Railway.
	api.HandleFunc("/auth/register", gatewayHandler.HandleAuthRegister).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/login", gatewayHandler.HandleAuthLogin).Methods("POST", "OPTIONS")

	// 404 handler for unmatched routes - must be registered last
	// This provides better error messages and logging for routes that don't exist
	// NOTE: gorilla/mux NotFoundHandler may not work with subrouters. If 404s return
	// plain text instead of JSON, the handler may not be called. Consider using
	// middleware or a different approach if this doesn't work.
	router.NotFoundHandler = http.HandlerFunc(gatewayHandler.HandleNotFound)

	// Frontend proxy (for development)
	if cfg.Environment == "development" {
		router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Info("🌐 API Gateway server starting",
			zap.String("port", cfg.Server.Port),
			zap.String("environment", cfg.Environment))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 Shutting down API Gateway server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("✅ API Gateway server stopped")
}
