package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configurations for the Classification Service
type Config struct {
	Server         ServerConfig
	Supabase       SupabaseConfig
	Classification ClassificationConfig
	Logging        LoggingConfig
}

// ServerConfig holds server-specific configurations
type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// SupabaseConfig holds Supabase integration configurations
type SupabaseConfig struct {
	URL            string
	APIKey         string
	ServiceRoleKey string
	JWTSecret      string
}

// ClassificationConfig holds classification-specific configurations
type ClassificationConfig struct {
	MaxConcurrentRequests int
	RequestTimeout        time.Duration
	CacheEnabled          bool
	CacheTTL              time.Duration
	RedisURL              string // Redis URL for distributed caching (optional)
	RedisEnabled          bool   // Whether to use Redis cache
	MLEnabled             bool
	KeywordMethodEnabled  bool
	EnsembleEnabled       bool
	// Multi-page analysis configuration
	MaxPagesToAnalyze   int
	PageAnalysisTimeout time.Duration
	OverallTimeout      time.Duration
	ConcurrentPages     int
	BrandMatchEnabled   bool
	BrandMatchMCCRange  string
	// Feature flags
	MultiPageAnalysisEnabled        bool
	StructuredDataExtractionEnabled bool
	// Fast-path scraping configuration
	FastPathScrapingEnabled bool
	MaxConcurrentPages      int
	CrawlDelayMs            int
	FastPathMaxPages        int
	WebsiteScrapingTimeout  time.Duration
	// Website content caching
	WebsiteContentCacheTTL    time.Duration
	EnableWebsiteContentCache bool
	// Early termination configuration (Task 1.5)
	EnableEarlyTermination              bool
	EarlyTerminationConfidenceThreshold float64
	MinContentLengthForML               int
	SkipFullCrawlIfContentSufficient    bool
	// Priority 5.3: Ensemble weight configuration
	PythonMLWeight float64 // Weight for Python ML service (default: 0.60)
	GoClassificationWeight float64 // Weight for Go classification (default: 0.40)
	// Phase 3: Embedding service configuration
	EmbeddingServiceURL string // URL of the embedding service (optional)
	// Phase 4: LLM service configuration
	LLMServiceURL string // URL of the LLM service (optional)
	// Rate limiting configuration (Fix 5.2)
	GlobalRateLimit          int  // Global requests per minute
	PerIPRateLimit           int  // Per-IP requests per minute
	RateLimitBurst           int  // Burst size for rate limiting
	EnableOverloadProtection bool // Circuit breaker for overload protection
}

// LoggingConfig holds logging configurations
type LoggingConfig struct {
	Level  string
	Format string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnvAsString("PORT", "8081"),
			Host:         getEnvAsString("HOST", "0.0.0.0"),
		// ALIGNED TIMEOUTS: Read/Write timeouts set to 120s to match processing timeout
		// This ensures HTTP server doesn't close connection before processing completes
		ReadTimeout:  getEnvAsDuration("READ_TIMEOUT", 120*time.Second),  // Aligned with processing timeout (120s)
		WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", 120*time.Second), // Aligned with processing timeout (120s)
		IdleTimeout:  getEnvAsDuration("IDLE_TIMEOUT", 60*time.Second),
		},
		Supabase: SupabaseConfig{
			URL:            getEnvAsString("SUPABASE_URL", ""),
			APIKey:         getEnvAsString("SUPABASE_ANON_KEY", ""),
			ServiceRoleKey: getEnvAsString("SUPABASE_SERVICE_ROLE_KEY", ""),
			JWTSecret:      getEnvAsString("SUPABASE_JWT_SECRET", ""),
		},
		Classification: ClassificationConfig{
			// Lower default concurrency to reduce memory pressure in production.
			// Reduced from 40 to 20 to prevent OOM kills (50% reduction in memory pressure)
			MaxConcurrentRequests: getEnvAsInt("MAX_CONCURRENT_REQUESTS", 20),
			// Rate limiting configuration
			GlobalRateLimit:      getEnvAsInt("GLOBAL_RATE_LIMIT", 200),      // Global requests per minute
			PerIPRateLimit:       getEnvAsInt("PER_IP_RATE_LIMIT", 100),      // Per-IP requests per minute
			RateLimitBurst:       getEnvAsInt("RATE_LIMIT_BURST", 20),        // Burst size for rate limiting
			EnableOverloadProtection: getEnvAsBool("ENABLE_OVERLOAD_PROTECTION", true), // Circuit breaker for overload
			// FIX #5: Changed default timeout from 10s to 120s to match worker timeout
			RequestTimeout:       getEnvAsDuration("REQUEST_TIMEOUT", 120*time.Second),
			CacheEnabled:         getEnvAsBool("CACHE_ENABLED", true),
			CacheTTL:             getEnvAsDuration("CACHE_TTL", 10*time.Minute), // Increased from 5m to 10m to improve cache hit rate from 49.6% to 60-70%
			RedisURL:             getEnvAsString("REDIS_URL", ""),
			RedisEnabled:         getEnvAsBool("REDIS_ENABLED", false),
			MLEnabled:            getEnvAsBool("ML_ENABLED", true),
			KeywordMethodEnabled: getEnvAsBool("KEYWORD_METHOD_ENABLED", true),
			EnsembleEnabled:      getEnvAsBool("ENSEMBLE_ENABLED", true),
			// Multi-page analysis configuration
			MaxPagesToAnalyze:   getEnvAsInt("CLASSIFICATION_MAX_PAGES_TO_ANALYZE", 15),
			PageAnalysisTimeout: getEnvAsDuration("CLASSIFICATION_PAGE_ANALYSIS_TIMEOUT", 15*time.Second),
			OverallTimeout:      getEnvAsDuration("CLASSIFICATION_OVERALL_TIMEOUT", 90*time.Second), // Increased from 60s to 90s to accommodate timeout budget (86s)
			ConcurrentPages:     getEnvAsInt("CLASSIFICATION_CONCURRENT_PAGES", 5),
			BrandMatchEnabled:   getEnvAsBool("CLASSIFICATION_BRAND_MATCH_ENABLED", true),
			BrandMatchMCCRange:  getEnvAsString("CLASSIFICATION_BRAND_MATCH_MCC_RANGE", "3000-3831"),
			// Feature flags
			MultiPageAnalysisEnabled:        getEnvAsBool("ENABLE_MULTI_PAGE_ANALYSIS", true),
			StructuredDataExtractionEnabled: getEnvAsBool("ENABLE_STRUCTURED_DATA_EXTRACTION", true),
			// Fast-path scraping configuration
			FastPathScrapingEnabled: getEnvAsBool("ENABLE_FAST_PATH_SCRAPING", true),
			MaxConcurrentPages:      getEnvAsInt("CLASSIFICATION_MAX_CONCURRENT_PAGES", 3),
			CrawlDelayMs:            getEnvAsInt("CLASSIFICATION_CRAWL_DELAY_MS", 500),
			FastPathMaxPages:        getEnvAsInt("CLASSIFICATION_FAST_PATH_MAX_PAGES", 8),
			WebsiteScrapingTimeout:  getEnvAsDuration("CLASSIFICATION_WEBSITE_SCRAPING_TIMEOUT", 15*time.Second), // Reduced from 20s to 15s for faster failure detection
			// Website content caching
			WebsiteContentCacheTTL:    getEnvAsDuration("WEBSITE_CONTENT_CACHE_TTL", 24*time.Hour),
			EnableWebsiteContentCache: getEnvAsBool("ENABLE_WEBSITE_CONTENT_CACHE", true),
			// Early termination configuration (Task 1.5)
			EnableEarlyTermination:              getEnvAsBool("ENABLE_EARLY_TERMINATION", true),
			EarlyTerminationConfidenceThreshold: getEnvAsFloat("EARLY_TERMINATION_CONFIDENCE_THRESHOLD", 0.70), // Reduced from 0.85 to allow more ML service usage
			MinContentLengthForML:               getEnvAsInt("MIN_CONTENT_LENGTH_FOR_ML", 30), // Reduced from 50 to make ML service more accessible
			SkipFullCrawlIfContentSufficient:    getEnvAsBool("SKIP_FULL_CRAWL_IF_CONTENT_SUFFICIENT", true),
			// Priority 5.3: Ensemble weight configuration (adjustable based on accuracy)
			PythonMLWeight:          getEnvAsFloat("PYTHON_ML_WEIGHT", 0.60),
			GoClassificationWeight:  getEnvAsFloat("GO_CLASSIFICATION_WEIGHT", 0.40),
			// Phase 3: Embedding service configuration
			EmbeddingServiceURL: getEnvAsString("EMBEDDING_SERVICE_URL", ""),
			// Phase 4: LLM service configuration
			LLMServiceURL: getEnvAsString("LLM_SERVICE_URL", ""),
		},
		Logging: LoggingConfig{
			Level:  getEnvAsString("LOG_LEVEL", "info"),
			Format: getEnvAsString("LOG_FORMAT", "json"),
		},
	}

	// Validate required Supabase configuration
	if cfg.Supabase.URL == "" || cfg.Supabase.APIKey == "" || cfg.Supabase.ServiceRoleKey == "" {
		return nil, fmt.Errorf("Supabase environment variables (SUPABASE_URL, SUPABASE_ANON_KEY, SUPABASE_SERVICE_ROLE_KEY) must be set")
	}

	// Validate service URLs (warn if critical services are not configured)
	// Note: These are warnings, not errors, as services may be optional or configured later
	if cfg.Classification.MLEnabled && os.Getenv("PYTHON_ML_SERVICE_URL") == "" {
		// Log warning but don't fail - service will work without ML
		fmt.Printf("⚠️  WARNING: ML_ENABLED is true but PYTHON_ML_SERVICE_URL is not set. ML classification will be unavailable.\n")
	}
	if os.Getenv("PLAYWRIGHT_SERVICE_URL") == "" {
		// Log info - Playwright is optional fallback
		fmt.Printf("ℹ️  INFO: PLAYWRIGHT_SERVICE_URL is not set. Playwright scraping strategy will be disabled.\n")
	}

	return cfg, nil
}

// Helper functions to get environment variables with defaults
func getEnvAsString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if durationValue, err := time.ParseDuration(value); err == nil {
			return durationValue
		}
	}
	return defaultValue
}

func getEnvAsStringSlice(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.Split(value, ",")
	}
	return defaultValue
}
