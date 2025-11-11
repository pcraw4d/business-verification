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

	logger.Info("🚀 Starting KYB API Gateway Service v1.0.19 - Fixed risk benchmarks/predictions routing")

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
	router.Use(middleware.CORS(cfg.CORS)) // Enable CORS middleware (FIRST)
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
			"version": "1.0.8",
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
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/classify", gatewayHandler.ProxyToClassification).Methods("POST")
	// OPTIONS handled by CORS middleware
	// Merchant routes - CORS handled by middleware
	api.HandleFunc("/merchants", gatewayHandler.ProxyToMerchants).Methods("GET", "POST", "OPTIONS")

	api.HandleFunc("/merchants/{id}", gatewayHandler.ProxyToMerchants).Methods("GET", "PUT", "DELETE", "OPTIONS")

	api.HandleFunc("/merchants/search", gatewayHandler.ProxyToMerchants).Methods("POST", "OPTIONS")

	api.HandleFunc("/merchants/analytics", gatewayHandler.ProxyToMerchants).Methods("GET", "OPTIONS")

	// Health check routes for backend services
	api.HandleFunc("/classification/health", gatewayHandler.ProxyToClassificationHealth).Methods("GET")
	api.HandleFunc("/merchant/health", gatewayHandler.ProxyToMerchantHealth).Methods("GET")
	api.HandleFunc("/risk/health", gatewayHandler.ProxyToRiskAssessmentHealth).Methods("GET")

	// Risk Assessment routes - CORS handled by middleware
	api.HandleFunc("/risk/assess", gatewayHandler.ProxyToRiskAssessment).Methods("POST", "OPTIONS")
	api.HandleFunc("/risk/benchmarks", gatewayHandler.ProxyToRiskAssessment).Methods("GET", "OPTIONS")
	api.HandleFunc("/risk/predictions/{merchant_id}", gatewayHandler.ProxyToRiskAssessment).Methods("GET", "OPTIONS")
	api.PathPrefix("/risk").HandlerFunc(gatewayHandler.ProxyToRiskAssessment)

	// Business Intelligence routes - CORS handled by middleware
	api.HandleFunc("/bi/analyze", gatewayHandler.ProxyToBI).Methods("POST", "OPTIONS")
	api.PathPrefix("/bi").HandlerFunc(gatewayHandler.ProxyToBI)

	// Authentication routes - CORS handled by middleware
	api.HandleFunc("/auth/register", gatewayHandler.HandleAuthRegister).Methods("POST", "OPTIONS")

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
