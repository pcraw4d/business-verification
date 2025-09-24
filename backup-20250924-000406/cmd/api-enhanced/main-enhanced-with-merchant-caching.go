package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/api/routes"
	"github.com/pcraw4d/business-verification/internal/cache"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/internal/services"
	"go.uber.org/zap"
)

// EnhancedMerchantCachingServer represents the enhanced server with merchant portfolio and caching
type EnhancedMerchantCachingServer struct {
	server                 *http.Server
	merchantService        *services.MerchantPortfolioService
	merchantHandler        *handlers.MerchantPortfolioHandler
	cacheFactory           *cache.CacheFactory
	merchantCacheService   *cache.MerchantCacheService
	cacheMonitoringService *cache.CacheMonitoringService
	alertingSystem         *cache.AlertingSystem
	logger                 *log.Logger
	zapLogger              *zap.Logger
}

// NewEnhancedMerchantCachingServer creates a new server with merchant portfolio and caching
func NewEnhancedMerchantCachingServer(port string) *EnhancedMerchantCachingServer {
	logger := log.New(os.Stdout, "🏪 ", log.LstdFlags|log.Lshortfile)
	zapLogger, _ := zap.NewProduction()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Supabase client
	supabaseConfig := &database.SupabaseConfig{
		URL:            cfg.Supabase.URL,
		APIKey:         cfg.Supabase.APIKey,
		ServiceRoleKey: cfg.Supabase.ServiceRoleKey,
		JWTSecret:      cfg.Supabase.JWTSecret,
	}
	supabaseClient, err := database.NewSupabaseClient(supabaseConfig, logger)
	if err != nil {
		logger.Fatalf("Failed to create Supabase client: %v", err)
	}

	// Connect to Supabase
	ctx := context.Background()
	if err := supabaseClient.Connect(ctx); err != nil {
		logger.Fatalf("Failed to connect to Supabase: %v", err)
	}
	logger.Printf("✅ Successfully connected to Supabase")

	// Create underlying merchant portfolio service
	underlyingMerchantService := services.NewMerchantPortfolioService(supabaseClient, logger)
	logger.Printf("✅ Underlying Merchant Portfolio Service initialized")

	// Initialize caching system
	cacheFactory := cache.NewCacheFactory(zapLogger)

	// Create memory cache for merchant data
	cacheConfig := &cache.CacheConfig{
		Type:            cache.MemoryCache,
		DefaultTTL:      30 * time.Minute, // Cache merchant data for 30 minutes
		MaxSize:         10000,            // Cache up to 10,000 merchant records
		KeyPrefix:       "merchant",
		CleanupInterval: 5 * time.Minute,
	}

	merchantCache, err := cacheFactory.CreateCache(cacheConfig)
	if err != nil {
		logger.Fatalf("Failed to create merchant cache: %v", err)
	}
	logger.Printf("✅ Merchant cache initialized")

	// Create merchant cache service
	merchantCacheService := cache.NewMerchantCacheService(merchantCache, nil, zapLogger)
	logger.Printf("✅ Merchant Cache Service initialized")

	// Create cache monitoring service
	monitoringConfig := &cache.MonitoringConfig{
		EnableMonitoring:    true,
		MetricsInterval:     30 * time.Second,
		HealthCheckInterval: 1 * time.Minute,
	}

	cacheMonitoringService := cache.NewCacheMonitoringService(monitoringConfig, []cache.Cache{merchantCache}, zapLogger)
	if err := cacheMonitoringService.Start(ctx); err != nil {
		logger.Printf("⚠️ Warning: Failed to start cache monitoring: %v", err)
	} else {
		logger.Printf("✅ Cache monitoring service started")
	}

	// Create alerting system
	alertingConfig := &cache.AlertingConfig{
		EnableAlerts:      true,
		CooldownPeriod:    5 * time.Minute,
		MaxHistoryEntries: 1000,
		EnableEscalation:  true,
		AlertThresholds: map[string]cache.AlertThreshold{
			"hit_rate_low": {
				Warning:  0.7, // Alert if hit rate drops below 70%
				Critical: 0.5, // Critical if hit rate drops below 50%
				Enabled:  true,
			},
			"memory_usage_high": {
				Warning:  0.8,  // Alert if memory usage exceeds 80%
				Critical: 0.95, // Critical if memory usage exceeds 95%
				Enabled:  true,
			},
			"error_rate_high": {
				Warning:  0.05, // Alert if error rate exceeds 5%
				Critical: 0.1,  // Critical if error rate exceeds 10%
				Enabled:  true,
			},
		},
	}

	alertingSystem := cache.NewAlertingSystem(alertingConfig, cacheMonitoringService.GetMetricsCollector(), zapLogger)

	// Register alert handlers
	loggingHandler := cache.NewLoggingAlertHandler(zapLogger)
	alertingSystem.RegisterHandler("hit_rate_low", loggingHandler)
	alertingSystem.RegisterHandler("memory_usage_high", loggingHandler)
	alertingSystem.RegisterHandler("error_rate_high", loggingHandler)

	logger.Printf("✅ Cache alerting system initialized")

	// Create cached merchant portfolio service
	merchantService := services.NewCachedMerchantPortfolioService(
		underlyingMerchantService,
		merchantCacheService,
		logger,
		zapLogger,
	)
	logger.Printf("✅ Cached Merchant Portfolio Service initialized")

	// Create merchant portfolio handler
	merchantHandler := handlers.NewMerchantPortfolioHandler(merchantService, logger)
	logger.Printf("✅ Merchant Portfolio Handler initialized")

	// Create server
	server := &EnhancedMerchantCachingServer{
		merchantService:        merchantService,
		merchantHandler:        merchantHandler,
		cacheFactory:           cacheFactory,
		merchantCacheService:   merchantCacheService,
		cacheMonitoringService: cacheMonitoringService,
		alertingSystem:         alertingSystem,
		logger:                 logger,
		zapLogger:              zapLogger,
	}

	// Setup routes
	mux := http.NewServeMux()
	server.setupRoutes(mux)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server.server = httpServer

	logger.Printf("✅ Enhanced Merchant Caching Server initialized on port %s", port)
	return server
}

// setupRoutes configures the HTTP routes
func (s *EnhancedMerchantCachingServer) setupRoutes(mux *http.ServeMux) {
	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)

	// Cache status endpoint
	mux.HandleFunc("/v1/cache/status", s.handleCacheStatus)

	// Cache metrics endpoint
	mux.HandleFunc("/v1/cache/metrics", s.handleCacheMetrics)

	// Cache alerts endpoint
	mux.HandleFunc("/v1/cache/alerts", s.handleCacheAlerts)

	// Register merchant portfolio routes
	observabilityLogger := observability.NewLogger(s.zapLogger)
	merchantRouteConfig := &routes.MerchantRouteConfig{
		MerchantPortfolioHandler: s.merchantHandler,
		AuthMiddleware:           nil, // TODO: Add auth middleware
		RateLimiter:              nil, // TODO: Add rate limiter
		Logger:                   observabilityLogger,
		EnableBulkOperations:     true,
		EnableSessionManagement:  true,
		MaxBulkOperationSize:     100,
	}

	routes.RegisterMerchantRoutes(mux, merchantRouteConfig)
}

// handleHealth handles health check requests
func (s *EnhancedMerchantCachingServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get cache health status
	cacheHealth := s.cacheMonitoringService.GetHealthStatus()

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"services": map[string]interface{}{
			"merchant_portfolio": "active",
			"cache":              cacheHealth,
			"monitoring":         "active",
			"alerting":           "active",
		},
	}

	json.NewEncoder(w).Encode(health)
}

// handleCacheStatus handles cache status requests
func (s *EnhancedMerchantCachingServer) handleCacheStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := s.cacheMonitoringService.GetHealthStatus()
	json.NewEncoder(w).Encode(status)
}

// handleCacheMetrics handles cache metrics requests
func (s *EnhancedMerchantCachingServer) handleCacheMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	metrics := s.cacheMonitoringService.GetMetrics()
	json.NewEncoder(w).Encode(metrics)
}

// handleCacheAlerts handles cache alerts requests
func (s *EnhancedMerchantCachingServer) handleCacheAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := context.Background()
	alerts := s.alertingSystem.CheckAlerts(ctx)

	response := map[string]interface{}{
		"active_alerts": alerts,
		"alert_history": s.alertingSystem.GetAlertHistory(),
	}

	json.NewEncoder(w).Encode(response)
}

// Start starts the server
func (s *EnhancedMerchantCachingServer) Start() error {
	s.logger.Printf("🚀 Starting Enhanced Merchant Caching Server on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop gracefully stops the server
func (s *EnhancedMerchantCachingServer) Stop(ctx context.Context) error {
	s.logger.Printf("🛑 Stopping Enhanced Merchant Caching Server...")

	// Stop cache monitoring
	if err := s.cacheMonitoringService.Stop(); err != nil {
		s.logger.Printf("⚠️ Error stopping cache monitoring: %v", err)
	}

	// Close merchant cache service
	if err := s.merchantCacheService.Close(); err != nil {
		s.logger.Printf("⚠️ Error closing merchant cache service: %v", err)
	}

	// Shutdown HTTP server
	return s.server.Shutdown(ctx)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := NewEnhancedMerchantCachingServer(port)

	// Handle graceful shutdown
	go func() {
		<-make(chan os.Signal, 1)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	// Start server
	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
