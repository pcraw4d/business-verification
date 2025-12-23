package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"runtime"
	"runtime/debug"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"kyb-platform/internal/classification"
	classificationAdapters "kyb-platform/internal/classification/adapters"
	keywordRepo "kyb-platform/internal/classification/repository"
	"kyb-platform/internal/machine_learning/infrastructure"
	serviceAdapters "kyb-platform/services/classification-service/internal/adapters"
	"kyb-platform/services/classification-service/internal/cache"
	"kyb-platform/services/classification-service/internal/config"
	"kyb-platform/services/classification-service/internal/errors"
	"kyb-platform/services/classification-service/internal/handlers"
	"kyb-platform/services/classification-service/internal/supabase"
)

// websiteScraperAdapter adapts EnhancedWebsiteScraper to WebsiteScraperInterface
type websiteScraperAdapter struct {
	scraper *classification.EnhancedWebsiteScraper
}

func (w *websiteScraperAdapter) ScrapeWebsite(ctx context.Context, websiteURL string) interface{} {
	result := w.scraper.ScrapeWebsite(ctx, websiteURL)
	return result // Return as interface{} to match interface
}

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
		zap.String("supabase_url", cfg.Supabase.URL),
		zap.Duration("read_timeout", cfg.Server.ReadTimeout),
		zap.Duration("write_timeout", cfg.Server.WriteTimeout))
	
	// Log critical feature flags for monitoring
	logger.Info("📋 Feature flags configuration",
		zap.Bool("ml_enabled", cfg.Classification.MLEnabled),
		zap.Bool("ensemble_enabled", cfg.Classification.EnsembleEnabled),
		zap.Bool("keyword_method_enabled", cfg.Classification.KeywordMethodEnabled),
		zap.Bool("multi_page_analysis_enabled", cfg.Classification.MultiPageAnalysisEnabled),
		zap.Bool("early_termination_enabled", cfg.Classification.EnableEarlyTermination),
		zap.Float64("early_termination_threshold", cfg.Classification.EarlyTerminationConfidenceThreshold),
		zap.Bool("cache_enabled", cfg.Classification.CacheEnabled),
		zap.Bool("redis_enabled", cfg.Classification.RedisEnabled))
	
	// Log service URLs for verification (masked for security)
	pythonMLURL := os.Getenv("PYTHON_ML_SERVICE_URL")
	playwrightURL := os.Getenv("PLAYWRIGHT_SERVICE_URL")
	logger.Info("🔗 Service URLs configuration",
		zap.Bool("python_ml_service_configured", pythonMLURL != ""),
		zap.Bool("playwright_service_configured", playwrightURL != ""),
		zap.String("embedding_service_url", cfg.Classification.EmbeddingServiceURL),
		zap.String("llm_service_url", cfg.Classification.LLMServiceURL))

	// Apply Go memory limit if provided (helps avoid OOM kills on Railway)
	applyMemoryLimit(logger)

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

	// Initialize Phase 1 enhanced website scraper for keyword extraction
	enhancedScraper := classification.NewEnhancedWebsiteScraper(stdLogger)
	logger.Info("✅ Phase 1 enhanced website scraper initialized for keyword extraction")

	// Create adapter to bridge EnhancedWebsiteScraper to WebsiteScraperInterface
	scraperAdapter := &websiteScraperAdapter{scraper: enhancedScraper}

	// Initialize classification repository with Phase 1 enhanced scraper
	keywordRepoInstance := keywordRepo.NewSupabaseKeywordRepositoryWithScraper(dbClient, stdLogger, scraperAdapter)
	logger.Info("✅ Classification repository initialized with Phase 1 enhanced scraper")

	// Initialize website content cache if enabled
	var websiteContentCache *cache.WebsiteContentCache
	if cfg.Classification.EnableWebsiteContentCache && cfg.Classification.RedisEnabled && cfg.Classification.RedisURL != "" {
		// Create Redis client for website content cache
		redisOpt, err := redis.ParseURL(cfg.Classification.RedisURL)
		if err == nil {
			redisClient := redis.NewClient(redisOpt)
			// Test connection
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := redisClient.Ping(ctx).Err(); err == nil {
				websiteContentCache = cache.NewWebsiteContentCache(
					redisClient,
					logger,
					cfg.Classification.WebsiteContentCacheTTL,
				)
				logger.Info("✅ Website content cache initialized",
					zap.Duration("ttl", cfg.Classification.WebsiteContentCacheTTL))
			} else {
				logger.Warn("⚠️ Failed to connect to Redis for website content cache, caching disabled",
					zap.Error(err))
			}
			cancel()
		} else {
			logger.Warn("⚠️ Failed to parse Redis URL for website content cache",
				zap.Error(err))
		}
	}

	// Initialize classification services
	industryDetector := classification.NewIndustryDetectionService(keywordRepoInstance, stdLogger)
	codeGenerator := classification.NewClassificationCodeGenerator(keywordRepoInstance, stdLogger)

	// Set website content cache on industry detector's multi-method classifier if available
	if websiteContentCache != nil && industryDetector != nil {
		// Create adapter to bridge cache package and classification package
		cacheAdapter := classification.NewWebsiteContentCacheAdapter(
			func(ctx context.Context, url string) (*classification.CachedWebsiteContent, bool) {
				cached, found := websiteContentCache.Get(ctx, url)
				if !found {
					return nil, false
				}
				// Convert cache package type to classification package type
				return &classification.CachedWebsiteContent{
					TextContent:    cached.TextContent,
					Title:          cached.Title,
					Keywords:       cached.Keywords,
					StructuredData: cached.StructuredData,
					ScrapedAt:      cached.ScrapedAt,
					Success:        cached.Success,
					StatusCode:     cached.StatusCode,
					ContentType:    cached.ContentType,
				}, true
			},
			func(ctx context.Context, url string, content *classification.CachedWebsiteContent) error {
				// Convert classification package type to cache package type
				cached := &cache.CachedWebsiteContent{
					TextContent:    content.TextContent,
					Title:          content.Title,
					Keywords:       content.Keywords,
					StructuredData: content.StructuredData,
					ScrapedAt:      content.ScrapedAt,
					Success:        content.Success,
					StatusCode:     content.StatusCode,
					ContentType:    content.ContentType,
				}
				return websiteContentCache.Set(ctx, url, cached)
			},
			func() bool {
				return websiteContentCache.IsEnabled()
			},
		)

		// Set cache on industry detector's multi-method classifier
		industryDetector.SetContentCache(cacheAdapter)
		logger.Info("✅ Website content cache set on classification services")
	}

	// Initialize Python ML service if URL is configured
	// OPTIMIZATION: Lazy initialization - defer heavy ML service initialization to reduce cold start time
	var pythonMLService *infrastructure.PythonMLService
	pythonMLServiceURL := os.Getenv("PYTHON_ML_SERVICE_URL")
	if pythonMLServiceURL != "" {
		logger.Info("🐍 Creating Python ML Service client (lazy initialization)",
			zap.String("url", pythonMLServiceURL))
		pythonMLService = infrastructure.NewPythonMLService(pythonMLServiceURL, stdLogger)

		// OPTIMIZATION: Initialize in background goroutine to avoid blocking startup
		// This reduces cold start time from ~30-40s to <10s
		go func() {
			logger.Info("🐍 [Background] Initializing Python ML Service",
				zap.String("url", pythonMLServiceURL))
			initCtx, initCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer initCancel()
			if err := pythonMLService.InitializeWithRetry(initCtx, 3); err != nil {
				logger.Warn("⚠️ Failed to initialize Python ML Service after retries, continuing without enhanced classification",
					zap.Error(err))
				// Note: Service will be nil-checked before use, so this is safe
			} else {
				logger.Info("✅ [Background] Python ML Service initialized successfully with retry logic")
			}
		}()
		logger.Info("✅ Python ML Service client created (initialization in background)")
	} else {
		logger.Info("ℹ️ Python ML Service URL not configured, enhanced classification will not be available")
	}

	// Phase 3: Initialize embedding classifier if URL is configured
	// OPTIMIZATION: Lazy initialization - defer to reduce cold start time
	if cfg.Classification.EmbeddingServiceURL != "" {
		logger.Info("🔍 Creating Embedding Classifier client (lazy initialization)",
			zap.String("url", cfg.Classification.EmbeddingServiceURL))
		embeddingClassifier := classification.NewEmbeddingClassifier(
			cfg.Classification.EmbeddingServiceURL,
			keywordRepoInstance,
			stdLogger,
		)
		industryDetector.SetEmbeddingClassifier(embeddingClassifier)
		logger.Info("✅ Embedding Classifier client created (ready for use)")
	} else {
		logger.Info("ℹ️ Embedding Service URL not configured, Layer 2 (embeddings) will not be available")
	}

	// Phase 4: Initialize LLM classifier if URL is configured
	// OPTIMIZATION: Lazy initialization - defer to reduce cold start time
	if cfg.Classification.LLMServiceURL != "" {
		logger.Info("🤖 Creating LLM Classifier client (lazy initialization)",
			zap.String("url", cfg.Classification.LLMServiceURL))
		llmClassifier := classification.NewLLMClassifier(
			cfg.Classification.LLMServiceURL,
			stdLogger,
		)
		industryDetector.SetLLMClassifier(llmClassifier)
		logger.Info("✅ LLM Classifier client created (ready for use)")
	} else {
		logger.Info("ℹ️ LLM Service URL not configured, Layer 3 (LLM) will not be available")
	}

	// Phase 5: Initialize classification cache
	classificationCache := classification.NewClassificationCache(keywordRepoInstance, stdLogger)
	industryDetector.SetClassificationCache(classificationCache)
	logger.Info("✅ [Phase 5] Classification cache initialized")

	logger.Info("✅ Classification services initialized",
		zap.Bool("industry_detector", industryDetector != nil),
		zap.Bool("code_generator", codeGenerator != nil),
		zap.Bool("python_ml_service", pythonMLService != nil),
		zap.Bool("embedding_classifier", cfg.Classification.EmbeddingServiceURL != ""),
		zap.Bool("llm_classifier", cfg.Classification.LLMServiceURL != ""),
		zap.Bool("classification_cache", classificationCache != nil))

	// Initialize handlers
	classificationHandler := handlers.NewClassificationHandler(
		supabaseClient,
		logger,
		cfg,
		industryDetector,
		codeGenerator,
		keywordRepoInstance, // OPTIMIZATION #5.2: Pass repository for accuracy tracking
		pythonMLService,     // Pass Python ML service (can be nil)
	)

	// Phase 5: Initialize dashboard handler
	dashboardHandler := handlers.NewDashboardHandlerWithLogger(keywordRepoInstance, stdLogger)
	logger.Info("✅ [Phase 5] Dashboard handler initialized")

	// Setup router
	router := mux.NewRouter()

	// Add middleware
	router.Use(recoveryMiddleware(logger))   // Recovery first to catch all panics
	router.Use(securityHeadersMiddleware())  // Add security headers
	router.Use(loggingMiddleware(logger))
	router.Use(corsMiddleware())
	router.Use(rateLimitMiddleware())
	// Priority 3 Fix: Increased timeout to 120s to match worker pool timeout and allow website scraping (86s adaptive timeout)
	router.Use(timeoutMiddleware(120 * time.Second)) // Increased from 30s to 120s for website scraping support

	// Register routes
	router.HandleFunc("/health", classificationHandler.HandleHealth).Methods("GET")
	router.HandleFunc("/health/cache", classificationHandler.HandleCacheHealth).Methods("GET")
	router.HandleFunc("/admin/circuit-breaker/reset", classificationHandler.HandleResetCircuitBreaker).Methods("POST")
	router.HandleFunc("/v1/classify", classificationHandler.HandleClassification).Methods("POST")
	router.HandleFunc("/classify", classificationHandler.HandleClassification).Methods("POST") // Alias for backward compatibility
	// OPTIMIZATION #5.2: Validation API endpoint
	router.HandleFunc("/v1/classify/validate", classificationHandler.HandleValidation).Methods("POST")
	router.HandleFunc("/classify/validate", classificationHandler.HandleValidation).Methods("POST") // Alias for backward compatibility
	// Phase 4: Async LLM status endpoints
	router.HandleFunc("/v1/classify/status/{processing_id}", classificationHandler.HandleLLMStatus).Methods("GET")
	router.HandleFunc("/classify/status/{processing_id}", classificationHandler.HandleLLMStatus).Methods("GET") // Alias
	router.HandleFunc("/v1/classify/async-stats", classificationHandler.HandleAsyncLLMStats).Methods("GET")
	router.HandleFunc("/classify/async-stats", classificationHandler.HandleAsyncLLMStats).Methods("GET") // Alias

	// Phase 5: Dashboard endpoints
	router.HandleFunc("/api/dashboard/summary", dashboardHandler.GetSummary).Methods("GET")
	router.HandleFunc("/api/dashboard/timeseries", dashboardHandler.GetTimeSeries).Methods("GET")

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Optional pprof (guarded by env)
	startPprof(logger)
	// Lightweight periodic memory diagnostics
	startMemoryDiagnostics(logger)

	// OPTIMIZATION: Pre-warm service by calling health endpoint after startup
	// This helps reduce cold start latency for first real request
	go func() {
		// Wait a moment for server to be ready
		time.Sleep(2 * time.Second)
		
		// Pre-warm by calling health endpoint
		healthURL := fmt.Sprintf("http://localhost:%s/health", cfg.Server.Port)
		logger.Info("🔥 Pre-warming service", zap.String("url", healthURL))
		
		client := &http.Client{Timeout: 5 * time.Second}
		if resp, err := client.Get(healthURL); err == nil {
			resp.Body.Close()
			logger.Info("✅ Service pre-warmed successfully")
		} else {
			logger.Warn("⚠️ Pre-warm failed (non-critical)", zap.Error(err))
		}
	}()

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

	// FIX #3: Stop worker pool first to prevent new requests from being processed
	if classificationHandler.WorkerPool != nil {
		logger.Info("Stopping worker pool...")
		classificationHandler.WorkerPool.Stop()
	}

	// Stop cleanup goroutines (FIX #7)
	if classificationHandler.ShutdownCancel != nil {
		logger.Info("Stopping cleanup goroutines...")
		classificationHandler.ShutdownCancel()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Classification Service forced to shutdown", zap.Error(err))
	}

	logger.Info("✅ Classification Service exited gracefully")
}

// applyMemoryLimit sets a memory limit based on environment variables.
// Priority:
// 1) GOMEMLIMIT_BYTES (bytes)
// 2) GOMEMLIMIT_MB (megabytes)
// 3) GOMEMLIMIT (already handled by Go runtime if set)
// 4) Default: 768 MiB
func applyMemoryLimit(logger *zap.Logger) {
	if val := os.Getenv("GOMEMLIMIT"); val != "" {
		logger.Info("Using GOMEMLIMIT from environment", zap.String("GOMEMLIMIT", val))
		return
	}

	if bytesStr := os.Getenv("GOMEMLIMIT_BYTES"); bytesStr != "" {
		if bytesVal, err := strconv.ParseInt(bytesStr, 10, 64); err == nil && bytesVal > 0 {
			debug.SetMemoryLimit(bytesVal)
			logger.Info("GOMEMLIMIT applied from GOMEMLIMIT_BYTES",
				zap.Int64("bytes", bytesVal))
			return
		}
	}

	if mbStr := os.Getenv("GOMEMLIMIT_MB"); mbStr != "" {
		if mbVal, err := strconv.ParseInt(mbStr, 10, 64); err == nil && mbVal > 0 {
			bytesVal := mbVal * 1024 * 1024
			debug.SetMemoryLimit(bytesVal)
			logger.Info("GOMEMLIMIT applied from GOMEMLIMIT_MB",
				zap.Int64("mb", mbVal),
				zap.Int64("bytes", bytesVal))
			return
		}
	}

	// Default to 768 MiB to reduce OOM risk on small Railway instances
	defaultLimit := int64(768 * 1024 * 1024)
	debug.SetMemoryLimit(defaultLimit)
	logger.Info("GOMEMLIMIT applied with default",
		zap.Int64("bytes", defaultLimit))
}

// startPprof starts a pprof server if PPROF_ENABLED=true.
// Uses PPROF_ADDR (default :6060). Intended for staging/production only if access-controlled.
func startPprof(logger *zap.Logger) {
	if strings.ToLower(os.Getenv("PPROF_ENABLED")) != "true" {
		return
	}
	addr := os.Getenv("PPROF_ADDR")
	if addr == "" {
		addr = ":6060"
	}
	go func() {
		logger.Info("pprof server starting", zap.String("addr", addr))
		//nolint:gosec // pprof intended for trusted access only
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.Warn("pprof server stopped", zap.Error(err))
		}
	}()
}

// startMemoryDiagnostics logs memory stats periodically to catch leaks/pressure.
// Enhanced with threshold alerts for memory usage monitoring.
func startMemoryDiagnostics(logger *zap.Logger) {
	// Memory threshold constants
	const (
		memoryWarningThreshold  = 70.0 // 70% memory usage - warning level
		memoryCriticalThreshold = 85.0 // 85% memory usage - critical level
	)

	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for range ticker.C {
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)

			// Calculate memory usage percentage
			memUsagePercent := float64(ms.Alloc) / float64(ms.Sys) * 100

			// Log memory stats
			logger.Info("memstats",
				zap.Uint64("alloc_bytes", ms.Alloc),
				zap.Uint64("heap_alloc_bytes", ms.HeapAlloc),
				zap.Uint64("heap_sys_bytes", ms.HeapSys),
				zap.Uint64("heap_inuse_bytes", ms.HeapInuse),
				zap.Uint64("num_gc", uint64(ms.NumGC)),
				zap.Float64("mem_usage_percent", memUsagePercent))

			// Check memory thresholds and alert
			if memUsagePercent > memoryCriticalThreshold {
				logger.Error("CRITICAL: Memory usage exceeds critical threshold",
					zap.Float64("mem_usage_percent", memUsagePercent),
					zap.Uint64("alloc_bytes", ms.Alloc),
					zap.Uint64("sys_bytes", ms.Sys),
					zap.Uint64("num_gc", uint64(ms.NumGC)),
					zap.Float64("threshold", memoryCriticalThreshold))
			} else if memUsagePercent > memoryWarningThreshold {
				logger.Warn("WARNING: Memory usage exceeds warning threshold",
					zap.Float64("mem_usage_percent", memUsagePercent),
					zap.Uint64("alloc_bytes", ms.Alloc),
					zap.Uint64("sys_bytes", ms.Sys),
					zap.Float64("threshold", memoryWarningThreshold))
			}
		}
	}()
}

// recoveryMiddleware catches panics and prevents service crashes
func recoveryMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("Panic recovered in HTTP handler",
						zap.Any("panic", err),
						zap.String("method", r.Method),
						zap.String("url", r.URL.String()),
						zap.String("remote_addr", r.RemoteAddr),
						zap.Stack("stack"))
					
					// Return 500 error instead of crashing
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
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

// rateLimitMiddleware adds enhanced rate limiting using golang.org/x/time/rate (Phase 5)
func rateLimitMiddleware() func(http.Handler) http.Handler {
	// Phase 5: Use token bucket rate limiter from golang.org/x/time/rate
	// Default: 100 requests per minute per IP (allows bursts up to 10)
	// Rate: 100 requests/minute = 100/60 requests/second ≈ 1.67 req/s
	// Burst: Allow up to 10 requests in quick succession
	rateLimitPerIP := rate.Limit(100.0 / 60.0) // 100 requests per minute
	burstSize := 10

	// Map to store rate limiters per IP address
	var (
		limiters = make(map[string]*rate.Limiter)
		mu       sync.RWMutex
	)

	// Cleanup old limiters periodically to prevent memory leak
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			// Keep only recent limiters (simple cleanup - in production, consider LRU cache)
			if len(limiters) > 10000 {
				// Clear half of the limiters (simple strategy)
				cleared := 0
				for ip := range limiters {
					if cleared >= len(limiters)/2 {
						break
					}
					delete(limiters, ip)
					cleared++
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract client IP (consider X-Forwarded-For header for proxies)
			clientIP := r.RemoteAddr
			if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
				// Take the first IP in the chain
				ips := strings.Split(forwardedFor, ",")
				if len(ips) > 0 {
					clientIP = strings.TrimSpace(ips[0])
				}
			} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
				clientIP = realIP
			}

			// Get or create rate limiter for this IP
			mu.RLock()
			limiter, exists := limiters[clientIP]
			mu.RUnlock()

			if !exists {
				mu.Lock()
				// Double-check after acquiring write lock
				limiter, exists = limiters[clientIP]
				if !exists {
					limiter = rate.NewLimiter(rateLimitPerIP, burstSize)
					limiters[clientIP] = limiter
				}
				mu.Unlock()
			}

			// Check if request is allowed
			if !limiter.Allow() {
				errors.WriteError(w, r, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded", fmt.Sprintf("Too many requests from IP %s. Limit: 100 requests per minute", clientIP))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// timeoutMiddleware adds request timeout middleware (Phase 5)
// Priority 3 Fix: Enhanced with timeout monitoring and logging
func timeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			requestPath := r.URL.Path
			
			// Create context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Create a response writer wrapper to detect if response was written
			done := make(chan bool, 1)
			wrapped := &timeoutResponseWriter{
				ResponseWriter: w,
				done:           done,
			}

			// Handle request in goroutine
			go func() {
				next.ServeHTTP(wrapped, r.WithContext(ctx))
				done <- true
			}()

			// Wait for either completion or timeout
			select {
			case <-done:
				// Request completed successfully
				duration := time.Since(startTime)
				// Log slow requests (>30s) for monitoring
				if duration > 30*time.Second {
					log.Printf("⏱️ [TIMEOUT-MIDDLEWARE] Slow request completed: %s %s (duration: %v, timeout: %v)",
						r.Method, requestPath, duration, timeout)
				}
				return
			case <-ctx.Done():
				// Timeout occurred
				duration := time.Since(startTime)
				log.Printf("❌ [TIMEOUT-MIDDLEWARE] Request timeout: %s %s (duration: %v, timeout: %v)",
					r.Method, requestPath, duration, timeout)
				
				// FIX: Acquire lock BEFORE checking wroteHeader to prevent race condition
				// This prevents concurrent map read/write errors
				wrapped.mu.Lock()
				if !wrapped.wroteHeader {
					wrapped.wroteHeader = true
					wrapped.mu.Unlock()
					errors.WriteError(w, r, http.StatusRequestTimeout, "REQUEST_TIMEOUT", "Request timeout", fmt.Sprintf("Request exceeded timeout of %v", timeout))
				} else {
					wrapped.mu.Unlock()
				}
			}
		})
	}
}

// timeoutResponseWriter wraps http.ResponseWriter to track if headers were written
type timeoutResponseWriter struct {
	http.ResponseWriter
	done        chan bool
	wroteHeader bool
	mu          sync.Mutex
}

func (tw *timeoutResponseWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if !tw.wroteHeader {
		tw.wroteHeader = true
		tw.ResponseWriter.WriteHeader(code)
	}
}

func (tw *timeoutResponseWriter) Write(b []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if !tw.wroteHeader {
		tw.wroteHeader = true
	}
	return tw.ResponseWriter.Write(b)
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
