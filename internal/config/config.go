package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment type
type Environment string

// Environment constants
const (
	Development Environment = "development"
	Production  Environment = "production"
	Testing     Environment = "testing"
	Staging     Environment = "staging"
)

// FeaturesConfig holds feature flag configuration
type FeaturesConfig struct {
	BusinessClassification bool `json:"business_classification" yaml:"business_classification"`
	RiskAssessment         bool `json:"risk_assessment" yaml:"risk_assessment"`
	ComplianceFramework    bool `json:"compliance_framework" yaml:"compliance_framework"`
	AdvancedAnalytics      bool `json:"advanced_analytics" yaml:"advanced_analytics"`
	RealTimeMonitoring     bool `json:"real_time_monitoring" yaml:"real_time_monitoring"`
}

// Config holds all configuration for the application
type Config struct {
	Environment string `json:"environment" yaml:"environment"`

	// Provider selection
	Provider ProviderConfig `json:"provider" yaml:"provider"`

	// Provider-specific configurations
	Supabase SupabaseConfig `json:"supabase" yaml:"supabase"`

	// Common configurations
	Server           ServerConfig           `json:"server" yaml:"server"`
	CORS             CORSConfig             `json:"cors" yaml:"cors"`
	RateLimit        RateLimitConfig        `json:"rate_limit" yaml:"rate_limit"`
	Auth             AuthConfig             `json:"auth" yaml:"auth"`
	Database         DatabaseConfig         `json:"database" yaml:"database"`
	Observability    ObservabilityConfig    `json:"observability" yaml:"observability"`
	ExternalServices ExternalServicesConfig `json:"external_services" yaml:"external_services"`
	Security         SecurityConfig         `json:"security" yaml:"security"`
	Features         FeaturesConfig         `json:"features" yaml:"features"`
}

// ProviderConfig holds provider selection configuration
type ProviderConfig struct {
	Database string `json:"database" yaml:"database"` // "supabase", "aws", "gcp"
	Auth     string `json:"auth" yaml:"auth"`         // "supabase", "aws", "gcp"
	Cache    string `json:"cache" yaml:"cache"`       // "supabase", "aws", "gcp"
	Storage  string `json:"storage" yaml:"storage"`   // "supabase", "aws", "gcp"
}

// SupabaseConfig holds Supabase-specific configuration
type SupabaseConfig struct {
	URL            string `json:"url" yaml:"url"`
	APIKey         string `json:"api_key" yaml:"api_key"`
	ServiceRoleKey string `json:"service_role_key" yaml:"service_role_key"`
	JWTSecret      string `json:"jwt_secret" yaml:"jwt_secret"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         int           `json:"port" yaml:"port"`
	Host         string        `json:"host" yaml:"host"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
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

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	JWTSecret         string        `json:"jwt_secret" yaml:"jwt_secret"`
	JWTExpiration     time.Duration `json:"jwt_expiration" yaml:"jwt_expiration"`
	RefreshExpiration time.Duration `json:"refresh_expiration" yaml:"refresh_expiration"`
	MinPasswordLength int           `json:"min_password_length" yaml:"min_password_length"`
	RequireUppercase  bool          `json:"require_uppercase" yaml:"require_uppercase"`
	RequireLowercase  bool          `json:"require_lowercase" yaml:"require_lowercase"`
	RequireNumbers    bool          `json:"require_numbers" yaml:"require_numbers"`
	RequireSpecial    bool          `json:"require_special" yaml:"require_special"`
	MaxLoginAttempts  int           `json:"max_login_attempts" yaml:"max_login_attempts"`
	LockoutDuration   time.Duration `json:"lockout_duration" yaml:"lockout_duration"`
	RefreshCookieName string        `json:"refresh_cookie_name" yaml:"refresh_cookie_name"`
	CSRFCookieName    string        `json:"csrf_cookie_name" yaml:"csrf_cookie_name"`
	CookieDomain      string        `json:"cookie_domain" yaml:"cookie_domain"`
	CookiePath        string        `json:"cookie_path" yaml:"cookie_path"`
	CookieSecure      bool          `json:"cookie_secure" yaml:"cookie_secure"`
	CookieSameSite    string        `json:"cookie_same_site" yaml:"cookie_same_site"`
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

// ObservabilityConfig holds observability-related configuration
type ObservabilityConfig struct {
	LogLevel  string `json:"log_level" yaml:"log_level"`
	LogFormat string `json:"log_format" yaml:"log_format"`

	MetricsEnabled bool   `json:"metrics_enabled" yaml:"metrics_enabled"`
	MetricsPort    int    `json:"metrics_port" yaml:"metrics_port"`
	MetricsPath    string `json:"metrics_path" yaml:"metrics_path"`

	TracingEnabled bool   `json:"tracing_enabled" yaml:"tracing_enabled"`
	TracingURL     string `json:"tracing_url" yaml:"tracing_url"`

	HealthCheckPath string `json:"health_check_path" yaml:"health_check_path"`
}

// ExternalServicesConfig holds external service configurations
type ExternalServicesConfig struct {
	BusinessDataAPI     BusinessDataAPIConfig     `json:"business_data_api" yaml:"business_data_api"`
	RiskAssessmentAPI   RiskAssessmentAPIConfig   `json:"risk_assessment_api" yaml:"risk_assessment_api"`
	ComplianceAPI       ComplianceAPIConfig       `json:"compliance_api" yaml:"compliance_api"`
	ClassificationCache ClassificationCacheConfig `json:"classification_cache" yaml:"classification_cache"`
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

// ClassificationCacheConfig holds classification cache configuration
type ClassificationCacheConfig struct {
	Enabled         bool          `json:"enabled" yaml:"enabled"`
	TTL             time.Duration `json:"ttl" yaml:"ttl"`
	MaxEntries      int           `json:"max_entries" yaml:"max_entries"`
	JanitorInterval time.Duration `json:"janitor_interval" yaml:"janitor_interval"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	IPBlock    IPBlockConfig    `json:"ip_block" yaml:"ip_block"`
	Validation ValidationConfig `json:"validation" yaml:"validation"`
}

// IPBlockConfig holds IP blocking configuration
type IPBlockConfig struct {
	Enabled       bool          `json:"enabled" yaml:"enabled"`
	Threshold     int           `json:"threshold" yaml:"threshold"`
	Window        time.Duration `json:"window" yaml:"window"`
	BlockDuration time.Duration `json:"block_duration" yaml:"block_duration"`
	Whitelist     []string      `json:"whitelist" yaml:"whitelist"`
	Blacklist     []string      `json:"blacklist" yaml:"blacklist"`
}

// ValidationConfig holds request validation configuration
type ValidationConfig struct {
	Enabled       bool     `json:"enabled" yaml:"enabled"`
	MaxBodySize   int64    `json:"max_body_size" yaml:"max_body_size"`
	RequiredPaths []string `json:"required_paths" yaml:"required_paths"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		Environment:      getEnvAsString("ENV", "development"),
		Provider:         getProviderConfig(),
		Supabase:         getSupabaseConfig(),
		Server:           getServerConfig(),
		CORS:             getCORSConfig(),
		RateLimit:        getRateLimitConfig(),
		Auth:             getAuthConfig(),
		Database:         getDatabaseConfig(),
		Observability:    getObservabilityConfig(),
		ExternalServices: getExternalServicesConfig(),
		Security:         getSecurityConfig(),
		Features:         getFeaturesConfig(),
	}

	return config, nil
}

// getProviderConfig returns provider configuration from environment variables
func getProviderConfig() ProviderConfig {
	return ProviderConfig{
		Database: getEnvAsString("PROVIDER_DATABASE", "supabase"),
		Auth:     getEnvAsString("PROVIDER_AUTH", "supabase"),
		Cache:    getEnvAsString("PROVIDER_CACHE", "supabase"),
		Storage:  getEnvAsString("PROVIDER_STORAGE", "supabase"),
	}
}

// getSupabaseConfig returns Supabase configuration from environment variables
func getSupabaseConfig() SupabaseConfig {
	return SupabaseConfig{
		URL:            getEnvAsString("SUPABASE_URL", ""),
		APIKey:         getEnvAsString("SUPABASE_API_KEY", ""),
		ServiceRoleKey: getEnvAsString("SUPABASE_SERVICE_ROLE_KEY", ""),
		JWTSecret:      getEnvAsString("SUPABASE_JWT_SECRET", ""),
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
	}
}

// getCORSConfig returns CORS configuration from environment variables
func getCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   getEnvAsStringSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
		AllowedMethods:   getEnvAsStringSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		AllowedHeaders:   getEnvAsStringSlice("CORS_ALLOWED_HEADERS", []string{"*"}),
		AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
		MaxAge:           getEnvAsInt("CORS_MAX_AGE", 86400),
	}
}

// getRateLimitConfig returns rate limiting configuration from environment variables
func getRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:     getEnvAsBool("RATE_LIMIT_ENABLED", true),
		RequestsPer: getEnvAsInt("RATE_LIMIT_REQUESTS_PER", 100),
		WindowSize:  getEnvAsInt("RATE_LIMIT_WINDOW_SIZE", 60),
		BurstSize:   getEnvAsInt("RATE_LIMIT_BURST_SIZE", 200),
	}
}

// getAuthConfig returns authentication configuration from environment variables
func getAuthConfig() AuthConfig {
	return AuthConfig{
		JWTSecret:         getEnvAsString("JWT_SECRET", ""),
		JWTExpiration:     getEnvAsDuration("JWT_EXPIRATION", 24*time.Hour),
		RefreshExpiration: getEnvAsDuration("REFRESH_EXPIRATION", 168*time.Hour),
		MinPasswordLength: getEnvAsInt("MIN_PASSWORD_LENGTH", 8),
		RequireUppercase:  getEnvAsBool("REQUIRE_UPPERCASE", true),
		RequireLowercase:  getEnvAsBool("REQUIRE_LOWERCASE", true),
		RequireNumbers:    getEnvAsBool("REQUIRE_NUMBERS", true),
		RequireSpecial:    getEnvAsBool("REQUIRE_SPECIAL", true),
		MaxLoginAttempts:  getEnvAsInt("MAX_LOGIN_ATTEMPTS", 5),
		LockoutDuration:   getEnvAsDuration("LOCKOUT_DURATION", 15*time.Minute),
		RefreshCookieName: getEnvAsString("REFRESH_COOKIE_NAME", "refresh_token"),
		CSRFCookieName:    getEnvAsString("CSRF_COOKIE_NAME", "XSRF-TOKEN"),
		CookieDomain:      getEnvAsString("COOKIE_DOMAIN", ""),
		CookiePath:        getEnvAsString("COOKIE_PATH", "/"),
		CookieSecure:      getEnvAsBool("COOKIE_SECURE", true),
		CookieSameSite:    getEnvAsString("COOKIE_SAMESITE", "Lax"),
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

// getObservabilityConfig returns observability configuration from environment variables
func getObservabilityConfig() ObservabilityConfig {
	return ObservabilityConfig{
		LogLevel:  getEnvAsString("LOG_LEVEL", "info"),
		LogFormat: getEnvAsString("LOG_FORMAT", "json"),

		MetricsEnabled: getEnvAsBool("METRICS_ENABLED", true),
		MetricsPort:    getEnvAsInt("METRICS_PORT", 9090),
		MetricsPath:    getEnvAsString("METRICS_PATH", "/metrics"),

		TracingEnabled: getEnvAsBool("TRACING_ENABLED", true),
		TracingURL:     getEnvAsString("TRACING_URL", "http://localhost:14268/api/traces"),

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
		ClassificationCache: ClassificationCacheConfig{
			Enabled:         getEnvAsBool("CLASSIFICATION_CACHE_ENABLED", true),
			TTL:             getEnvAsDuration("CLASSIFICATION_CACHE_TTL", 10*time.Minute),
			MaxEntries:      getEnvAsInt("CLASSIFICATION_CACHE_MAX_ENTRIES", 10000),
			JanitorInterval: getEnvAsDuration("CLASSIFICATION_CACHE_JANITOR_INTERVAL", 1*time.Minute),
		},
	}
}

// getSecurityConfig returns security configuration from environment variables
func getSecurityConfig() SecurityConfig {
	return SecurityConfig{
		IPBlock: IPBlockConfig{
			Enabled:       getEnvAsBool("IP_BLOCK_ENABLED", true),
			Threshold:     getEnvAsInt("IP_BLOCK_THRESHOLD", 20),
			Window:        getEnvAsDuration("IP_BLOCK_WINDOW", 5*time.Minute),
			BlockDuration: getEnvAsDuration("IP_BLOCK_DURATION", 30*time.Minute),
			Whitelist:     getEnvAsStringSlice("IP_BLOCK_WHITELIST", []string{}),
			Blacklist:     getEnvAsStringSlice("IP_BLOCK_BLACKLIST", []string{}),
		},
		Validation: ValidationConfig{
			Enabled:       getEnvAsBool("VALIDATION_ENABLED", true),
			MaxBodySize:   getEnvAsInt64("VALIDATION_MAX_BODY_SIZE", 10*1024*1024), // 10MB
			RequiredPaths: getEnvAsStringSlice("VALIDATION_REQUIRED_PATHS", []string{"/v1/"}),
		},
	}
}

// getFeaturesConfig returns feature flag configuration from environment variables
func getFeaturesConfig() FeaturesConfig {
	return FeaturesConfig{
		BusinessClassification: getEnvAsBool("FEATURE_BUSINESS_CLASSIFICATION", true),
		RiskAssessment:         getEnvAsBool("FEATURE_RISK_ASSESSMENT", true),
		ComplianceFramework:    getEnvAsBool("FEATURE_COMPLIANCE_FRAMEWORK", true),
		AdvancedAnalytics:      getEnvAsBool("FEATURE_ADVANCED_ANALYTICS", false),
		RealTimeMonitoring:     getEnvAsBool("FEATURE_REAL_TIME_MONITORING", false),
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

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
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

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.Driver == "" {
		return fmt.Errorf("database driver is required")
	}

	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if c.Auth.MinPasswordLength < 6 {
		return fmt.Errorf("minimum password length must be at least 6")
	}

	// Validate external services configuration
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
	switch env {
	case "production", "prod":
		return Production
	case "staging", "stage":
		return Staging
	case "testing", "test":
		return Testing
	case "development", "dev":
		return Development
	default:
		return Development
	}
}
