package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/config"
	"kyb-platform/services/risk-assessment-service/internal/engine"
	"kyb-platform/services/risk-assessment-service/internal/external"
	"kyb-platform/services/risk-assessment-service/internal/handlers"
	"kyb-platform/services/risk-assessment-service/internal/middleware"
	"kyb-platform/services/risk-assessment-service/internal/ml/service"
	"kyb-platform/services/risk-assessment-service/internal/monitoring"
	"kyb-platform/services/risk-assessment-service/internal/performance"
	"kyb-platform/services/risk-assessment-service/internal/supabase"
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

	logger.Info("🚀 Starting Risk Assessment Service v1.0.0")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	logger.Info("✅ Configuration loaded successfully",
		zap.String("port", cfg.Server.Port),
		zap.String("supabase_url", cfg.Supabase.URL),
		zap.String("log_level", cfg.Logging.Level))

	// Initialize Supabase client
	supabaseConfig := &supabase.Config{
		URL:            cfg.Supabase.URL,
		APIKey:         cfg.Supabase.APIKey,
		ServiceRoleKey: cfg.Supabase.ServiceRoleKey,
		JWTSecret:      cfg.Supabase.JWTSecret,
	}
	supabaseClient, err := supabase.NewClient(supabaseConfig, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Supabase client", zap.Error(err))
	}
	logger.Info("✅ Supabase client initialized")

	// Initialize ML service
	mlService := service.NewMLService(logger)
	if err := mlService.InitializeModels(context.Background()); err != nil {
		logger.Fatal("Failed to initialize ML models", zap.Error(err))
	}
	logger.Info("✅ ML service initialized")

	// Initialize risk engine
	riskEngineConfig := &engine.Config{
		MaxConcurrentRequests: 100,
		RequestTimeout:        500 * time.Millisecond, // Sub-1-second target
		CacheTTL:              5 * time.Minute,
		CircuitBreakerConfig: engine.CircuitBreakerConfig{
			FailureThreshold: 5,
			RecoveryTimeout:  30 * time.Second,
			HalfOpenMaxCalls: 3,
		},
		EnableMetrics: true,
		EnableCaching: true,
	}
	riskEngine := engine.NewRiskEngine(mlService, logger, riskEngineConfig)
	logger.Info("✅ Risk engine initialized")

	// Initialize external data service
	externalDataConfig := &external.ExternalDataConfig{
		NewsAPIKey:           cfg.External.NewsAPI.APIKey,
		OpenCorporatesKey:    cfg.External.OpenCorporates.APIKey,
		GovernmentAPIKey:     cfg.External.OFAC.APIKey,
		Timeout:              15 * time.Second,
		EnableNewsAPI:        cfg.External.NewsAPI.APIKey != "",
		EnableOpenCorporates: cfg.External.OpenCorporates.APIKey != "",
		EnableGovernment:     cfg.External.OFAC.APIKey != "",
	}
	externalDataService := external.NewExternalDataService(externalDataConfig, logger)
	logger.Info("✅ External data service initialized")

	// Initialize performance optimization system
	performanceConfig := performance.DefaultOptimizerConfig()
	performanceConfig.TargetP95 = 1 * time.Second
	performanceConfig.TargetP99 = 2 * time.Second
	performanceConfig.TargetThroughput = 1000

	// Note: In a real implementation, you'd pass the actual database connection
	// For now, we'll initialize without database optimization
	performanceConfig.EnableDBOptimization = false
	performanceOptimizer := performance.NewOptimizer(logger, nil, performanceConfig)
	logger.Info("✅ Performance optimizer initialized")

	// Initialize legacy performance monitor for compatibility
	performanceMonitor := monitoring.NewPerformanceMonitor(logger)
	performanceMonitor.SetTargets(16.67, 1*time.Second, 0.01, 1000) // 1000 req/min target
	logger.Info("✅ Performance monitor initialized")

	// Initialize handlers
	riskAssessmentHandler := handlers.NewRiskAssessmentHandler(supabaseClient, mlService, riskEngine, externalDataService, logger, cfg)
	advancedPredictionHandler := handlers.NewAdvancedPredictionHandler(mlService, logger)
	metricsHandler := handlers.NewMetricsHandler(mlService.GetMetricsCollector(), logger)
	performanceHandler := handlers.NewPerformanceHandler(performanceMonitor, logger)

	// Initialize middleware
	middlewareInstance := middleware.NewMiddleware(logger)

	// Initialize new performance middleware with optimizer
	performanceMiddlewareConfig := performance.DefaultMiddlewareConfig()
	performanceMiddleware := performance.NewPerformanceMiddleware(
		logger,
		performanceOptimizer.GetProfiler(),
		performanceOptimizer.GetResponseMonitor(),
		performanceOptimizer.GetCacheOptimizer(),
		performanceMiddlewareConfig,
	)

	// Legacy performance middleware for compatibility
	legacyPerformanceMiddleware := middleware.NewPerformanceMiddleware(performanceMonitor, logger)

	// Setup router
	router := mux.NewRouter()

	// Add comprehensive middleware
	router.Use(middlewareInstance.RecoveryMiddleware())
	router.Use(middlewareInstance.LoggingMiddleware)
	router.Use(performanceMiddleware.Middleware)         // New performance monitoring
	router.Use(legacyPerformanceMiddleware.Middleware()) // Legacy performance monitoring
	router.Use(middlewareInstance.SecurityMiddleware())
	router.Use(middlewareInstance.RequestSizeMiddleware(10 * 1024 * 1024)) // 10MB limit
	router.Use(middlewareInstance.TimeoutMiddleware(30 * time.Second))
	router.Use(middlewareInstance.CORSMiddleware(
		[]string{"*"}, // Allowed origins
		[]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},       // Allowed methods
		[]string{"Content-Type", "Authorization", "X-Request-ID"}, // Allowed headers
	))
	router.Use(middlewareInstance.RateLimitMiddleware(100)) // 100 requests per minute
	router.Use(middlewareInstance.MetricsMiddleware())
	router.Use(middlewareInstance.HealthCheckMiddleware())

	// Health check endpoint
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Risk assessment endpoints
	api.HandleFunc("/assess", riskAssessmentHandler.HandleRiskAssessment).Methods("POST")
	api.HandleFunc("/assess/batch", riskAssessmentHandler.HandleBatchRiskAssessment).Methods("POST")
	api.HandleFunc("/assess/{id}", riskAssessmentHandler.HandleGetRiskAssessment).Methods("GET")
	api.HandleFunc("/assess/{id}/predict", riskAssessmentHandler.HandleRiskPrediction).Methods("POST")
	api.HandleFunc("/assess/{id}/history", riskAssessmentHandler.HandleRiskHistory).Methods("GET")

	// Advanced prediction endpoints
	api.HandleFunc("/risk/predict-advanced", advancedPredictionHandler.HandleAdvancedPrediction).Methods("POST")
	api.HandleFunc("/models/info", advancedPredictionHandler.HandleGetModelInfo).Methods("GET")
	api.HandleFunc("/models/performance", advancedPredictionHandler.HandleGetModelPerformance).Methods("GET")

	// Compliance endpoints
	api.HandleFunc("/compliance/check", riskAssessmentHandler.HandleComplianceCheck).Methods("POST")
	api.HandleFunc("/sanctions/screen", riskAssessmentHandler.HandleSanctionsScreening).Methods("POST")
	api.HandleFunc("/media/monitor", riskAssessmentHandler.HandleAdverseMediaMonitoring).Methods("POST")

	// Analytics endpoints
	api.HandleFunc("/analytics/trends", riskAssessmentHandler.HandleRiskTrends).Methods("GET")
	api.HandleFunc("/analytics/insights", riskAssessmentHandler.HandleRiskInsights).Methods("GET")

	// Metrics and monitoring endpoints
	api.HandleFunc("/metrics", metricsHandler.HandleGetMetrics).Methods("GET")
	api.HandleFunc("/health", metricsHandler.HandleGetHealth).Methods("GET")
	api.HandleFunc("/performance", metricsHandler.HandleGetPerformanceSnapshot).Methods("GET")

	// Performance monitoring endpoints (legacy)
	api.HandleFunc("/performance/stats", performanceHandler.HandlePerformanceStats).Methods("GET")
	api.HandleFunc("/performance/alerts", performanceHandler.HandlePerformanceAlerts).Methods("GET")
	api.HandleFunc("/performance/health", performanceHandler.HandlePerformanceHealth).Methods("GET")
	api.HandleFunc("/performance/reset", performanceHandler.HandlePerformanceReset).Methods("POST")
	api.HandleFunc("/performance/targets", performanceHandler.HandlePerformanceTargets).Methods("POST")
	api.HandleFunc("/performance/alerts/clear", performanceHandler.HandlePerformanceClearAlerts).Methods("POST")

	// New performance optimization endpoints
	api.HandleFunc("/optimization/report", func(w http.ResponseWriter, r *http.Request) {
		report, err := performanceOptimizer.Optimize()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(report)
	}).Methods("GET")

	api.HandleFunc("/optimization/stats", func(w http.ResponseWriter, r *http.Request) {
		stats := performanceOptimizer.GetPerformanceStats()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}).Methods("GET")

	api.HandleFunc("/optimization/health", func(w http.ResponseWriter, r *http.Request) {
		health := performanceOptimizer.GetHealthStatus()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	}).Methods("GET")

	api.HandleFunc("/optimization/report/full", func(w http.ResponseWriter, r *http.Request) {
		report := performanceOptimizer.GetPerformanceReport()
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(report))
	}).Methods("GET")

	// External data source endpoints
	api.HandleFunc("/external/adverse-media", riskAssessmentHandler.HandleExternalAdverseMediaMonitoring).Methods("POST")
	api.HandleFunc("/external/company-data", riskAssessmentHandler.HandleCompanyDataLookup).Methods("POST")
	api.HandleFunc("/external/compliance", riskAssessmentHandler.HandleExternalComplianceCheck).Methods("POST")
	api.HandleFunc("/external/sources", riskAssessmentHandler.HandleExternalDataSources).Methods("GET")

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start performance monitoring
	go performanceMonitor.StartMonitoring(context.Background())

	// Start server in a goroutine
	go func() {
		logger.Info("🌐 Starting HTTP server", zap.String("addr", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown external data service
	externalDataService.Close()

	// Shutdown risk engine
	if err := riskEngine.Shutdown(ctx); err != nil {
		logger.Error("Risk engine shutdown failed", zap.Error(err))
	}

	// Attempt graceful shutdown
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("✅ Server exited")
}

// Health check handler
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"risk-assessment-service","version":"1.0.0"}`))
}
