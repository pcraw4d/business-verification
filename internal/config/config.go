package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment represents the application environment
type Environment string

const (
	Development Environment = "development"
	Staging     Environment = "staging"
	Production  Environment = "production"
	Testing     Environment = "testing"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Server ServerConfig `json:"server" yaml:"server"`

	// Database configuration
	Database DatabaseConfig `json:"database" yaml:"database"`

	// Authentication configuration
	Auth AuthConfig `json:"auth" yaml:"auth"`

	// Observability configuration
	Observability ObservabilityConfig `json:"observability" yaml:"observability"`

	// External services configuration
	ExternalServices ExternalServicesConfig `json:"external_services" yaml:"external_services"`

	// Feature flags
	Features FeatureFlags `json:"features" yaml:"features"`

	// Environment
	Environment Environment `json:"environment" yaml:"environment"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         int           `json:"port" yaml:"port"`
	Host         string        `json:"host" yaml:"host"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout"`

	// CORS configuration
	CORS CORSConfig `json:"cors" yaml:"cors"`

	// Rate limiting
	RateLimit RateLimitConfig `json:"rate_limit" yaml:"rate_limit"`
}

// CORSConfig holds CORS-related configuration
type CORSConfig struct {
	AllowedOrigins   []string `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods" yaml:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers" yaml:"allowed_headers"`
	AllowCredentials bool     `json:"allow_credentials" yaml:"allow_credentials"`
	MaxAge           int      `json:"max_age" yaml:"max_age"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled     bool `json:"enabled" yaml:"enabled"`
	RequestsPer int  `json:"requests_per" yaml:"requests_per"`
	WindowSize  int  `json:"window_size" yaml:"window_size"`
	BurstSize   int  `json:"burst_size" yaml:"burst_size"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Driver   string `json:"driver" yaml:"driver"`
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Database string `json:"database" yaml:"database"`
	SSLMode  string `json:"ssl_mode" yaml:"ssl_mode"`

	// Connection pool settings
	MaxOpenConns    int           `json:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`

	// Migration settings
	AutoMigrate bool `json:"auto_migrate" yaml:"auto_migrate"`
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	// JWT configuration
	JWTSecret         string        `json:"jwt_secret" yaml:"jwt_secret"`
	JWTExpiration     time.Duration `json:"jwt_expiration" yaml:"jwt_expiration"`
	RefreshExpiration time.Duration `json:"refresh_expiration" yaml:"refresh_expiration"`

	// Password configuration
	MinPasswordLength int  `json:"min_password_length" yaml:"min_password_length"`
	RequireUppercase  bool `json:"require_uppercase" yaml:"require_uppercase"`
	RequireLowercase  bool `json:"require_lowercase" yaml:"require_lowercase"`
	RequireNumbers    bool `json:"require_numbers" yaml:"require_numbers"`
	RequireSpecial    bool `json:"require_special" yaml:"require_special"`

	// Account lockout
	MaxLoginAttempts int           `json:"max_login_attempts" yaml:"max_login_attempts"`
	LockoutDuration  time.Duration `json:"lockout_duration" yaml:"lockout_duration"`
}

// ObservabilityConfig holds observability-related configuration
type ObservabilityConfig struct {
	// Logging
	LogLevel  string `json:"log_level" yaml:"log_level"`
	LogFormat string `json:"log_format" yaml:"log_format"`

	// Metrics
	MetricsEnabled bool   `json:"metrics_enabled" yaml:"metrics_enabled"`
	MetricsPort    int    `json:"metrics_port" yaml:"metrics_port"`
	MetricsPath    string `json:"metrics_path" yaml:"metrics_path"`

	// Tracing
	TracingEnabled bool   `json:"tracing_enabled" yaml:"tracing_enabled"`
	TracingURL     string `json:"tracing_url" yaml:"tracing_url"`

	// Health checks
	HealthCheckPath string `json:"health_check_path" yaml:"health_check_path"`
}

// ExternalServicesConfig holds external service configurations
type ExternalServicesConfig struct {
	// Business data providers
	BusinessDataAPI BusinessDataAPIConfig `json:"business_data_api" yaml:"business_data_api"`

	// Risk assessment services
	RiskAssessmentAPI RiskAssessmentAPIConfig `json:"risk_assessment_api" yaml:"risk_assessment_api"`

	// Compliance services
	ComplianceAPI ComplianceAPIConfig `json:"compliance_api" yaml:"compliance_api"`
}

// BusinessDataAPIConfig holds business data API configuration
type BusinessDataAPIConfig struct {
	Enabled    bool          `json:"enabled" yaml:"enabled"`
	BaseURL    string        `json:"base_url" yaml:"base_url"`
	APIKey     string        `json:"api_key" yaml:"api_key"`
	Timeout    time.Duration `json:"timeout" yaml:"timeout"`
	MaxRetries int           `json:"max_retries" yaml:"max_retries"`
}

// RiskAssessmentAPIConfig holds risk assessment API configuration
type RiskAssessmentAPIConfig struct {
	Enabled    bool          `json:"enabled" yaml:"enabled"`
	BaseURL    string        `json:"base_url" yaml:"base_url"`
	APIKey     string        `json:"api_key" yaml:"api_key"`
	Timeout    time.Duration `json:"timeout" yaml:"timeout"`
	MaxRetries int           `json:"max_retries" yaml:"max_retries"`
}

// ComplianceAPIConfig holds compliance API configuration
type ComplianceAPIConfig struct {
	Enabled    bool          `json:"enabled" yaml:"enabled"`
	BaseURL    string        `json:"base_url" yaml:"base_url"`
	APIKey     string        `json:"api_key" yaml:"api_key"`
	Timeout    time.Duration `json:"timeout" yaml:"timeout"`
	MaxRetries int           `json:"max_retries" yaml:"max_retries"`
}

// FeatureFlags holds feature flag configurations
type FeatureFlags struct {
	// Core features
	BusinessClassification bool `json:"business_classification" yaml:"business_classification"`
	RiskAssessment         bool `json:"risk_assessment" yaml:"risk_assessment"`
	ComplianceFramework    bool `json:"compliance_framework" yaml:"compliance_framework"`

	// Advanced features
	AdvancedAnalytics  bool `json:"advanced_analytics" yaml:"advanced_analytics"`
	RealTimeMonitoring bool `json:"real_time_monitoring" yaml:"real_time_monitoring"`
	BatchProcessing    bool `json:"batch_processing" yaml:"batch_processing"`

	// Integration features
	APIKeyManagement    bool `json:"api_key_management" yaml:"api_key_management"`
	WebhookSupport      bool `json:"webhook_support" yaml:"webhook_support"`
	ExportFunctionality bool `json:"export_functionality" yaml:"export_functionality"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		Environment:      getEnvironment(),
		Server:           getServerConfig(),
		Database:         getDatabaseConfig(),
		Auth:             getAuthConfig(),
		Observability:    getObservabilityConfig(),
		ExternalServices: getExternalServicesConfig(),
		Features:         getFeatureFlags(),
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server configuration
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	// Validate database configuration
	if c.Database.Driver == "" {
		return fmt.Errorf("database driver is required")
	}

	// Validate auth configuration
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if c.Auth.MinPasswordLength < 8 {
		return fmt.Errorf("minimum password length must be at least 8")
	}

	// Validate external services
	if c.ExternalServices.BusinessDataAPI.Enabled {
		if c.ExternalServices.BusinessDataAPI.BaseURL == "" {
			return fmt.Errorf("business data API base URL is required when enabled")
		}
		if c.ExternalServices.BusinessDataAPI.APIKey == "" {
			return fmt.Errorf("business data API key is required when enabled")
		}
	}

	return nil
}

// getEnvironment returns the current environment
func getEnvironment() Environment {
	env := os.Getenv("ENV")
	switch strings.ToLower(env) {
	case "production", "prod":
		return Production
	case "staging", "stage":
		return Staging
	case "testing", "test":
		return Testing
	default:
		return Development
	}
}

// getServerConfig returns server configuration from environment variables
func getServerConfig() ServerConfig {
	return ServerConfig{
		Port:         getEnvAsInt("PORT", 8080),
		Host:         getEnvAsString("HOST", "0.0.0.0"),
		ReadTimeout:  getEnvAsDuration("READ_TIMEOUT", 30*time.Second),
		WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:  getEnvAsDuration("IDLE_TIMEOUT", 60*time.Second),
		CORS: CORSConfig{
			AllowedOrigins:   getEnvAsStringSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
			AllowedMethods:   getEnvAsStringSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders:   getEnvAsStringSlice("CORS_ALLOWED_HEADERS", []string{"*"}),
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           getEnvAsInt("CORS_MAX_AGE", 86400),
		},
		RateLimit: RateLimitConfig{
			Enabled:     getEnvAsBool("RATE_LIMIT_ENABLED", true),
			RequestsPer: getEnvAsInt("RATE_LIMIT_REQUESTS_PER", 100),
			WindowSize:  getEnvAsInt("RATE_LIMIT_WINDOW_SIZE", 60),
			BurstSize:   getEnvAsInt("RATE_LIMIT_BURST_SIZE", 200),
		},
	}
}

// getDatabaseConfig returns database configuration from environment variables
func getDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Driver:          getEnvAsString("DB_DRIVER", "postgres"),
		Host:            getEnvAsString("DB_HOST", "localhost"),
		Port:            getEnvAsInt("DB_PORT", 5432),
		Username:        getEnvAsString("DB_USERNAME", "postgres"),
		Password:        getEnvAsString("DB_PASSWORD", ""),
		Database:        getEnvAsString("DB_DATABASE", "business_verification"),
		SSLMode:         getEnvAsString("DB_SSL_MODE", "disable"),
		MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		AutoMigrate:     getEnvAsBool("DB_AUTO_MIGRATE", true),
	}
}

// getAuthConfig returns authentication configuration from environment variables
func getAuthConfig() AuthConfig {
	return AuthConfig{
		JWTSecret:         getEnvAsString("JWT_SECRET", ""),
		JWTExpiration:     getEnvAsDuration("JWT_EXPIRATION", 24*time.Hour),
		RefreshExpiration: getEnvAsDuration("REFRESH_EXPIRATION", 7*24*time.Hour),
		MinPasswordLength: getEnvAsInt("MIN_PASSWORD_LENGTH", 8),
		RequireUppercase:  getEnvAsBool("REQUIRE_UPPERCASE", true),
		RequireLowercase:  getEnvAsBool("REQUIRE_LOWERCASE", true),
		RequireNumbers:    getEnvAsBool("REQUIRE_NUMBERS", true),
		RequireSpecial:    getEnvAsBool("REQUIRE_SPECIAL", true),
		MaxLoginAttempts:  getEnvAsInt("MAX_LOGIN_ATTEMPTS", 5),
		LockoutDuration:   getEnvAsDuration("LOCKOUT_DURATION", 15*time.Minute),
	}
}

// getObservabilityConfig returns observability configuration from environment variables
func getObservabilityConfig() ObservabilityConfig {
	return ObservabilityConfig{
		LogLevel:        getEnvAsString("LOG_LEVEL", "info"),
		LogFormat:       getEnvAsString("LOG_FORMAT", "json"),
		MetricsEnabled:  getEnvAsBool("METRICS_ENABLED", true),
		MetricsPort:     getEnvAsInt("METRICS_PORT", 9090),
		MetricsPath:     getEnvAsString("METRICS_PATH", "/metrics"),
		TracingEnabled:  getEnvAsBool("TRACING_ENABLED", true),
		TracingURL:      getEnvAsString("TRACING_URL", "http://localhost:14268/api/traces"),
		HealthCheckPath: getEnvAsString("HEALTH_CHECK_PATH", "/health"),
	}
}

// getExternalServicesConfig returns external services configuration from environment variables
func getExternalServicesConfig() ExternalServicesConfig {
	return ExternalServicesConfig{
		BusinessDataAPI: BusinessDataAPIConfig{
			Enabled:    getEnvAsBool("BUSINESS_DATA_API_ENABLED", false),
			BaseURL:    getEnvAsString("BUSINESS_DATA_API_BASE_URL", ""),
			APIKey:     getEnvAsString("BUSINESS_DATA_API_KEY", ""),
			Timeout:    getEnvAsDuration("BUSINESS_DATA_API_TIMEOUT", 30*time.Second),
			MaxRetries: getEnvAsInt("BUSINESS_DATA_API_MAX_RETRIES", 3),
		},
		RiskAssessmentAPI: RiskAssessmentAPIConfig{
			Enabled:    getEnvAsBool("RISK_ASSESSMENT_API_ENABLED", false),
			BaseURL:    getEnvAsString("RISK_ASSESSMENT_API_BASE_URL", ""),
			APIKey:     getEnvAsString("RISK_ASSESSMENT_API_KEY", ""),
			Timeout:    getEnvAsDuration("RISK_ASSESSMENT_API_TIMEOUT", 30*time.Second),
			MaxRetries: getEnvAsInt("RISK_ASSESSMENT_API_MAX_RETRIES", 3),
		},
		ComplianceAPI: ComplianceAPIConfig{
			Enabled:    getEnvAsBool("COMPLIANCE_API_ENABLED", false),
			BaseURL:    getEnvAsString("COMPLIANCE_API_BASE_URL", ""),
			APIKey:     getEnvAsString("COMPLIANCE_API_KEY", ""),
			Timeout:    getEnvAsDuration("COMPLIANCE_API_TIMEOUT", 30*time.Second),
			MaxRetries: getEnvAsInt("COMPLIANCE_API_MAX_RETRIES", 3),
		},
	}
}

// getFeatureFlags returns feature flags from environment variables
func getFeatureFlags() FeatureFlags {
	return FeatureFlags{
		BusinessClassification: getEnvAsBool("FEATURE_BUSINESS_CLASSIFICATION", true),
		RiskAssessment:         getEnvAsBool("FEATURE_RISK_ASSESSMENT", true),
		ComplianceFramework:    getEnvAsBool("FEATURE_COMPLIANCE_FRAMEWORK", true),
		AdvancedAnalytics:      getEnvAsBool("FEATURE_ADVANCED_ANALYTICS", false),
		RealTimeMonitoring:     getEnvAsBool("FEATURE_REAL_TIME_MONITORING", false),
		BatchProcessing:        getEnvAsBool("FEATURE_BATCH_PROCESSING", false),
		APIKeyManagement:       getEnvAsBool("FEATURE_API_KEY_MANAGEMENT", true),
		WebhookSupport:         getEnvAsBool("FEATURE_WEBHOOK_SUPPORT", false),
		ExportFunctionality:    getEnvAsBool("FEATURE_EXPORT_FUNCTIONALITY", false),
	}
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
