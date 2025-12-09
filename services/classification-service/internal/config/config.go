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
	MaxPagesToAnalyze        int
	PageAnalysisTimeout      time.Duration
	OverallTimeout           time.Duration
	ConcurrentPages          int
	BrandMatchEnabled        bool
	BrandMatchMCCRange       string
	// Feature flags
	MultiPageAnalysisEnabled      bool
	StructuredDataExtractionEnabled bool
	// Fast-path scraping configuration
	FastPathScrapingEnabled  bool
	MaxConcurrentPages       int
	CrawlDelayMs             int
	FastPathMaxPages         int
	WebsiteScrapingTimeout   time.Duration
	// Website content caching
	WebsiteContentCacheTTL   time.Duration
	EnableWebsiteContentCache bool
	// Early termination configuration (Task 1.5)
	EnableEarlyTermination   bool
	EarlyTerminationConfidenceThreshold float64
	MinContentLengthForML    int
	SkipFullCrawlIfContentSufficient bool
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
			ReadTimeout:  getEnvAsDuration("READ_TIMEOUT", 100*time.Second), // Increased to 100s to accommodate optimized adaptive timeout (68s max processing + 32s buffer)
			WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", 120*time.Second), // Increased to 120s for long-running classifications
			IdleTimeout:  getEnvAsDuration("IDLE_TIMEOUT", 60*time.Second),
		},
		Supabase: SupabaseConfig{
			URL:            getEnvAsString("SUPABASE_URL", ""),
			APIKey:         getEnvAsString("SUPABASE_ANON_KEY", ""),
			ServiceRoleKey: getEnvAsString("SUPABASE_SERVICE_ROLE_KEY", ""),
			JWTSecret:      getEnvAsString("SUPABASE_JWT_SECRET", ""),
		},
		Classification: ClassificationConfig{
			MaxConcurrentRequests: getEnvAsInt("MAX_CONCURRENT_REQUESTS", 100),
			// FIX #5: Changed default timeout from 10s to 120s to match worker timeout
			RequestTimeout:        getEnvAsDuration("REQUEST_TIMEOUT", 120*time.Second),
			CacheEnabled:          getEnvAsBool("CACHE_ENABLED", true),
			CacheTTL:              getEnvAsDuration("CACHE_TTL", 5*time.Minute),
			RedisURL:              getEnvAsString("REDIS_URL", ""),
			RedisEnabled:          getEnvAsBool("REDIS_ENABLED", false),
			MLEnabled:             getEnvAsBool("ML_ENABLED", true),
			KeywordMethodEnabled:  getEnvAsBool("KEYWORD_METHOD_ENABLED", true),
			EnsembleEnabled:       getEnvAsBool("ENSEMBLE_ENABLED", true),
			// Multi-page analysis configuration
			MaxPagesToAnalyze:        getEnvAsInt("CLASSIFICATION_MAX_PAGES_TO_ANALYZE", 15),
			PageAnalysisTimeout:      getEnvAsDuration("CLASSIFICATION_PAGE_ANALYSIS_TIMEOUT", 15*time.Second),
			OverallTimeout:           getEnvAsDuration("CLASSIFICATION_OVERALL_TIMEOUT", 60*time.Second),
			ConcurrentPages:          getEnvAsInt("CLASSIFICATION_CONCURRENT_PAGES", 5),
			BrandMatchEnabled:        getEnvAsBool("CLASSIFICATION_BRAND_MATCH_ENABLED", true),
			BrandMatchMCCRange:       getEnvAsString("CLASSIFICATION_BRAND_MATCH_MCC_RANGE", "3000-3831"),
			// Feature flags
			MultiPageAnalysisEnabled:      getEnvAsBool("ENABLE_MULTI_PAGE_ANALYSIS", true),
			StructuredDataExtractionEnabled: getEnvAsBool("ENABLE_STRUCTURED_DATA_EXTRACTION", true),
			// Fast-path scraping configuration
			FastPathScrapingEnabled:  getEnvAsBool("ENABLE_FAST_PATH_SCRAPING", true),
			MaxConcurrentPages:       getEnvAsInt("CLASSIFICATION_MAX_CONCURRENT_PAGES", 3),
			CrawlDelayMs:             getEnvAsInt("CLASSIFICATION_CRAWL_DELAY_MS", 500),
			FastPathMaxPages:         getEnvAsInt("CLASSIFICATION_FAST_PATH_MAX_PAGES", 8),
			WebsiteScrapingTimeout:   getEnvAsDuration("CLASSIFICATION_WEBSITE_SCRAPING_TIMEOUT", 15*time.Second), // Increased from 5s to 15s for better success rate
			// Website content caching
			WebsiteContentCacheTTL:   getEnvAsDuration("WEBSITE_CONTENT_CACHE_TTL", 24*time.Hour),
			EnableWebsiteContentCache: getEnvAsBool("ENABLE_WEBSITE_CONTENT_CACHE", true),
			// Early termination configuration (Task 1.5)
			EnableEarlyTermination:   getEnvAsBool("ENABLE_EARLY_TERMINATION", true),
			EarlyTerminationConfidenceThreshold: getEnvAsFloat("EARLY_TERMINATION_CONFIDENCE_THRESHOLD", 0.85),
			MinContentLengthForML:    getEnvAsInt("MIN_CONTENT_LENGTH_FOR_ML", 50),
			SkipFullCrawlIfContentSufficient: getEnvAsBool("SKIP_FULL_CRAWL_IF_CONTENT_SUFFICIENT", true),
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
