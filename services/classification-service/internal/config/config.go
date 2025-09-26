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
	MLEnabled             bool
	KeywordMethodEnabled  bool
	EnsembleEnabled       bool
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
			ReadTimeout:  getEnvAsDuration("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", 30*time.Second),
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
			RequestTimeout:        getEnvAsDuration("REQUEST_TIMEOUT", 10*time.Second),
			CacheEnabled:          getEnvAsBool("CACHE_ENABLED", true),
			CacheTTL:              getEnvAsDuration("CACHE_TTL", 5*time.Minute),
			MLEnabled:             getEnvAsBool("ML_ENABLED", true),
			KeywordMethodEnabled:  getEnvAsBool("KEYWORD_METHOD_ENABLED", true),
			EnsembleEnabled:       getEnvAsBool("ENSEMBLE_ENABLED", true),
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
