package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the API Gateway
type Config struct {
	Server      ServerConfig
	Supabase    SupabaseConfig
	CORS        CORSConfig
	RateLimit   RateLimitConfig
	Services    ServicesConfig
	Environment string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// SupabaseConfig holds Supabase configuration
type SupabaseConfig struct {
	URL            string
	APIKey         string
	ServiceRoleKey string
	JWTSecret      string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled     bool
	RequestsPer int
	WindowSize  int
	BurstSize   int
}

// ServicesConfig holds configuration for backend services
type ServicesConfig struct {
	ClassificationURL string
	MerchantURL       string
	FrontendURL       string
	BIServiceURL      string
	RiskAssessmentURL string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnvAsString("PORT", "8080"),
			Host:         getEnvAsString("HOST", "0.0.0.0"),
			ReadTimeout:  getEnvAsDuration("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getEnvAsDuration("IDLE_TIMEOUT", 60*time.Second),
		},
		Supabase: SupabaseConfig{
			URL:            getEnvAsString("SUPABASE_URL", ""),
			APIKey:         getEnvAsString("SUPABASE_ANON_KEY", getEnvAsString("SUPABASE_API_KEY", "")),
			ServiceRoleKey: getEnvAsString("SUPABASE_SERVICE_ROLE_KEY", ""),
			JWTSecret:      getEnvAsString("SUPABASE_JWT_SECRET", ""),
		},
		CORS: CORSConfig{
			// Default to specific frontend origin instead of wildcard to avoid browser rejection with credentials
			// Wildcard (*) cannot be used with AllowCredentials=true in browsers
			AllowedOrigins:   getEnvAsStringSlice("CORS_ALLOWED_ORIGINS", []string{"https://frontend-service-production-b225.up.railway.app"}),
			AllowedMethods:   getEnvAsStringSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders:   getEnvAsStringSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization"}),
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           getEnvAsInt("CORS_MAX_AGE", 86400),
		},
		RateLimit: RateLimitConfig{
			Enabled:     getEnvAsBool("RATE_LIMIT_ENABLED", true),
			RequestsPer: getEnvAsInt("RATE_LIMIT_REQUESTS_PER", 1000),
			WindowSize:  getEnvAsInt("RATE_LIMIT_WINDOW_SIZE", 3600),
			BurstSize:   getEnvAsInt("RATE_LIMIT_BURST_SIZE", 2000),
		},
		Environment: getEnvAsString("ENVIRONMENT", "production"),
	}

	// Set service URLs based on environment (after Environment is set)
	cfg.Services = ServicesConfig{
		// Use localhost URLs for local development, Railway URLs for production
		// This allows local development without impacting Railway deployment
		ClassificationURL: getServiceURL("CLASSIFICATION_SERVICE_URL", "classification-service", cfg.Environment),
		MerchantURL:       getServiceURL("MERCHANT_SERVICE_URL", "merchant-service", cfg.Environment),
		FrontendURL:       getServiceURL("FRONTEND_URL", "frontend-service", cfg.Environment),
		BIServiceURL:      getServiceURL("BI_SERVICE_URL", "bi-service", cfg.Environment),
		RiskAssessmentURL: getServiceURL("RISK_ASSESSMENT_SERVICE_URL", "risk-assessment-service", cfg.Environment),
	}

	// Validate required configuration
	if cfg.Supabase.URL == "" {
		return nil, fmt.Errorf("SUPABASE_URL is required")
	}
	if cfg.Supabase.APIKey == "" {
		return nil, fmt.Errorf("SUPABASE_ANON_KEY is required")
	}

	return cfg, nil
}

// Helper functions for environment variable parsing
func getEnvAsString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// getServiceURL returns the service URL based on environment.
// For local development (ENVIRONMENT=development), uses localhost URLs.
// For production, uses Railway URLs or environment variable if set.
// This ensures local development works without impacting Railway deployment.
func getServiceURL(envVar, serviceName, environment string) string {
	// If explicitly set via environment variable, use it
	if url := os.Getenv(envVar); url != "" {
		return url
	}

	// For local development, use localhost URLs
	if environment == "development" {
		// Default ports for local services (can be overridden via env vars)
		portMap := map[string]string{
			"classification-service": getEnvAsString("CLASSIFICATION_SERVICE_PORT", "8081"),
			"merchant-service":       getEnvAsString("MERCHANT_SERVICE_PORT", "8082"),
			"frontend-service":       getEnvAsString("FRONTEND_SERVICE_PORT", "3000"),
			"bi-service":             getEnvAsString("BI_SERVICE_PORT", "8083"),
			"risk-assessment-service": getEnvAsString("RISK_ASSESSMENT_SERVICE_PORT", "8084"),
		}
		port := portMap[serviceName]
		if port == "" {
			port = "8080" // Default fallback
		}
		return fmt.Sprintf("http://localhost:%s", port)
	}

	// For production, use Railway URLs
	railwayURLs := map[string]string{
		"classification-service":  "https://classification-service-production.up.railway.app",
		"merchant-service":        "https://merchant-service-production.up.railway.app",
		"frontend-service":        "https://frontend-service-production-b225.up.railway.app",
		"bi-service":              "https://bi-service-production.up.railway.app",
		"risk-assessment-service": "https://risk-assessment-service-production.up.railway.app",
	}
	return railwayURLs[serviceName]
}
