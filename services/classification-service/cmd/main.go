package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"runtime"
	"runtime/debug"
	"strconv"

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
	var pythonMLService *infrastructure.PythonMLService
	pythonMLServiceURL := os.Getenv("PYTHON_ML_SERVICE_URL")
	if pythonMLServiceURL != "" {
		logger.Info("🐍 Initializing Python ML Service",
			zap.String("url", pythonMLServiceURL))
		pythonMLService = infrastructure.NewPythonMLService(pythonMLServiceURL, stdLogger)

		// Initialize the service with retry logic for resilience (3 retries with exponential backoff)
		// This handles transient ML service startup issues gracefully
		initCtx, initCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer initCancel()
		if err := pythonMLService.InitializeWithRetry(initCtx, 3); err != nil {
			logger.Warn("⚠️ Failed to initialize Python ML Service after retries, continuing without enhanced classification",
				zap.Error(err))
			pythonMLService = nil // Set to nil so service continues without it
		} else {
			logger.Info("✅ Python ML Service initialized successfully with retry logic")
		}
	} else {
		logger.Info("ℹ️ Python ML Service URL not configured, enhanced classification will not be available")
	}

	// Phase 3: Initialize embedding classifier if URL is configured
	if cfg.Classification.EmbeddingServiceURL != "" {
		logger.Info("🔍 Initializing Embedding Classifier (Phase 3)",
			zap.String("url", cfg.Classification.EmbeddingServiceURL))
		embeddingClassifier := classification.NewEmbeddingClassifier(
			cfg.Classification.EmbeddingServiceURL,
			keywordRepoInstance,
			stdLogger,
		)
		industryDetector.SetEmbeddingClassifier(embeddingClassifier)
		logger.Info("✅ Embedding Classifier initialized and set on Industry Detection Service")
	} else {
		logger.Info("ℹ️ Embedding Service URL not configured, Layer 2 (embeddings) will not be available")
	}

	// Phase 4: Initialize LLM classifier if URL is configured
	if cfg.Classification.LLMServiceURL != "" {
		logger.Info("🤖 Initializing LLM Classifier (Phase 4)",
			zap.String("url", cfg.Classification.LLMServiceURL))
		llmClassifier := classification.NewLLMClassifier(
			cfg.Classification.LLMServiceURL,
			stdLogger,
		)
		industryDetector.SetLLMClassifier(llmClassifier)
		logger.Info("✅ LLM Classifier initialized and set on Industry Detection Service")
	} else {
		logger.Info("ℹ️ LLM Service URL not configured, Layer 3 (LLM) will not be available")
	}

	logger.Info("✅ Classification services initialized",
		zap.Bool("industry_detector", industryDetector != nil),
		zap.Bool("code_generator", codeGenerator != nil),
		zap.Bool("python_ml_service", pythonMLService != nil),
		zap.Bool("embedding_classifier", cfg.Classification.EmbeddingServiceURL != ""),
		zap.Bool("llm_classifier", cfg.Classification.LLMServiceURL != ""))

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

	// Setup router
	router := mux.NewRouter()

	// Add middleware
	router.Use(recoveryMiddleware(logger))   // Recovery first to catch all panics
	router.Use(securityHeadersMiddleware())  // Add security headers
	router.Use(loggingMiddleware(logger))
	router.Use(corsMiddleware())
	router.Use(rateLimitMiddleware())

	// Register routes
	router.HandleFunc("/health", classificationHandler.HandleHealth).Methods("GET")
	router.HandleFunc("/v1/classify", classificationHandler.HandleClassification).Methods("POST")
	router.HandleFunc("/classify", classificationHandler.HandleClassification).Methods("POST") // Alias for backward compatibility
	// OPTIMIZATION #5.2: Validation API endpoint
	router.HandleFunc("/v1/classify/validate", classificationHandler.HandleValidation).Methods("POST")
	router.HandleFunc("/classify/validate", classificationHandler.HandleValidation).Methods("POST") // Alias for backward compatibility

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

// rateLimitMiddleware adds basic rate limiting
// FIX #4: Added mutex protection to prevent race conditions
func rateLimitMiddleware() func(http.Handler) http.Handler {
	var (
		requests = make(map[string][]time.Time)
		mu       sync.RWMutex
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := r.RemoteAddr
			now := time.Now()

			mu.Lock()
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
				mu.Unlock()
				errors.WriteError(w, r, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded", "Too many requests from this IP address")
				return
			}

			// Add current request
			requests[clientIP] = append(requests[clientIP], now)
			mu.Unlock()

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
