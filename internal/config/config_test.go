package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("JWT_SECRET", "test-secret-for-testing")
	defer os.Unsetenv("JWT_SECRET")

	// Test with default values
	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify default values
	if config.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", config.Server.Port)
	}

	if config.Database.Driver != "postgres" {
		t.Errorf("Expected default database driver 'postgres', got %s", config.Database.Driver)
	}

	if config.Environment != Development {
		t.Errorf("Expected default environment 'development', got %s", config.Environment)
	}
}

func TestLoadWithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("PORT", "9090")
	os.Setenv("DB_HOST", "test-host")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("ENV", "production")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("ENV")
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify environment variables are loaded
	if config.Server.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", config.Server.Port)
	}

	if config.Database.Host != "test-host" {
		t.Errorf("Expected database host 'test-host', got %s", config.Database.Host)
	}

	if config.Auth.JWTSecret != "test-secret" {
		t.Errorf("Expected JWT secret 'test-secret', got %s", config.Auth.JWTSecret)
	}

	if config.Environment != Production {
		t.Errorf("Expected environment 'production', got %s", config.Environment)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Driver: "postgres"},
				Auth:     AuthConfig{JWTSecret: "secret", MinPasswordLength: 8},
			},
			wantErr: false,
		},
		{
			name: "invalid port",
			config: &Config{
				Server:   ServerConfig{Port: 0},
				Database: DatabaseConfig{Driver: "postgres"},
				Auth:     AuthConfig{JWTSecret: "secret", MinPasswordLength: 8},
			},
			wantErr: true,
		},
		{
			name: "missing database driver",
			config: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Driver: ""},
				Auth:     AuthConfig{JWTSecret: "secret", MinPasswordLength: 8},
			},
			wantErr: true,
		},
		{
			name: "missing JWT secret",
			config: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Driver: "postgres"},
				Auth:     AuthConfig{JWTSecret: "", MinPasswordLength: 8},
			},
			wantErr: true,
		},
		{
			name: "password too short",
			config: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Driver: "postgres"},
				Auth:     AuthConfig{JWTSecret: "secret", MinPasswordLength: 4},
			},
			wantErr: true,
		},
		{
			name: "enabled external service without base URL",
			config: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Driver: "postgres"},
				Auth:     AuthConfig{JWTSecret: "secret", MinPasswordLength: 8},
				ExternalServices: ExternalServicesConfig{
					BusinessDataAPI: BusinessDataAPIConfig{
						Enabled: true,
						BaseURL: "",
						APIKey:  "key",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "enabled external service without API key",
			config: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Driver: "postgres"},
				Auth:     AuthConfig{JWTSecret: "secret", MinPasswordLength: 8},
				ExternalServices: ExternalServicesConfig{
					BusinessDataAPI: BusinessDataAPIConfig{
						Enabled: true,
						BaseURL: "http://api.example.com",
						APIKey:  "",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     Environment
	}{
		{"development", "development", Development},
		{"dev", "dev", Development},
		{"production", "production", Production},
		{"prod", "prod", Production},
		{"staging", "staging", Staging},
		{"stage", "stage", Staging},
		{"testing", "testing", Testing},
		{"test", "test", Testing},
		{"empty", "", Development},
		{"unknown", "unknown", Development},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("ENV", tt.envValue)
				defer os.Unsetenv("ENV")
			}

			got := getEnvironment()
			if got != tt.want {
				t.Errorf("getEnvironment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvironmentVariableParsing(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue interface{}
		want         interface{}
	}{
		{"string", "TEST_STRING", "test-value", "default", "test-value"},
		{"string default", "TEST_STRING_DEFAULT", "", "default", "default"},
		{"int", "TEST_INT", "123", 0, 123},
		{"int default", "TEST_INT_DEFAULT", "", 0, 0},
		{"bool true", "TEST_BOOL", "true", false, true},
		{"bool false", "TEST_BOOL_FALSE", "false", true, false},
		{"bool default", "TEST_BOOL_DEFAULT", "", false, false},
		{"duration", "TEST_DURATION", "30s", time.Duration(0), 30 * time.Second},
		{"duration default", "TEST_DURATION_DEFAULT", "", time.Duration(0), time.Duration(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			var got interface{}
			switch tt.defaultValue.(type) {
			case string:
				got = getEnvAsString(tt.envKey, tt.defaultValue.(string))
			case int:
				got = getEnvAsInt(tt.envKey, tt.defaultValue.(int))
			case bool:
				got = getEnvAsBool(tt.envKey, tt.defaultValue.(bool))
			case time.Duration:
				got = getEnvAsDuration(tt.envKey, tt.defaultValue.(time.Duration))
			}

			if got != tt.want {
				t.Errorf("Environment variable parsing failed for %s: got %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestStringSliceParsing(t *testing.T) {
	os.Setenv("TEST_SLICE", "value1,value2,value3")
	defer os.Unsetenv("TEST_SLICE")

	got := getEnvAsStringSlice("TEST_SLICE", []string{})
	expected := []string{"value1", "value2", "value3"}

	if len(got) != len(expected) {
		t.Errorf("Expected slice length %d, got %d", len(expected), len(got))
	}

	for i, v := range got {
		if v != expected[i] {
			t.Errorf("Expected value %s at index %d, got %s", expected[i], i, v)
		}
	}
}

func TestFeatureFlags(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("JWT_SECRET", "test-secret-for-testing")
	defer os.Unsetenv("JWT_SECRET")

	// Test default feature flags
	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify default feature flags
	if !config.Features.BusinessClassification {
		t.Error("Expected BusinessClassification to be enabled by default")
	}

	if !config.Features.RiskAssessment {
		t.Error("Expected RiskAssessment to be enabled by default")
	}

	if !config.Features.ComplianceFramework {
		t.Error("Expected ComplianceFramework to be enabled by default")
	}

	if config.Features.AdvancedAnalytics {
		t.Error("Expected AdvancedAnalytics to be disabled by default")
	}

	if config.Features.RealTimeMonitoring {
		t.Error("Expected RealTimeMonitoring to be disabled by default")
	}
}

func TestServerConfig(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("JWT_SECRET", "test-secret-for-testing")
	defer os.Unsetenv("JWT_SECRET")

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test server config defaults
	if config.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", config.Server.Port)
	}

	if config.Server.Host != "0.0.0.0" {
		t.Errorf("Expected default host '0.0.0.0', got %s", config.Server.Host)
	}

	if config.Server.ReadTimeout != 30*time.Second {
		t.Errorf("Expected default read timeout 30s, got %v", config.Server.ReadTimeout)
	}

	// Test CORS defaults
	if len(config.Server.CORS.AllowedOrigins) != 1 || config.Server.CORS.AllowedOrigins[0] != "*" {
		t.Errorf("Expected default CORS allowed origins ['*'], got %v", config.Server.CORS.AllowedOrigins)
	}

	// Test rate limiting defaults
	if !config.Server.RateLimit.Enabled {
		t.Error("Expected rate limiting to be enabled by default")
	}

	if config.Server.RateLimit.RequestsPer != 100 {
		t.Errorf("Expected default requests per window 100, got %d", config.Server.RateLimit.RequestsPer)
	}
}

func TestDatabaseConfig(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("JWT_SECRET", "test-secret-for-testing")
	defer os.Unsetenv("JWT_SECRET")

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test database config defaults
	if config.Database.Driver != "postgres" {
		t.Errorf("Expected default database driver 'postgres', got %s", config.Database.Driver)
	}

	if config.Database.Host != "localhost" {
		t.Errorf("Expected default database host 'localhost', got %s", config.Database.Host)
	}

	if config.Database.Port != 5432 {
		t.Errorf("Expected default database port 5432, got %d", config.Database.Port)
	}

	if config.Database.Database != "business_verification" {
		t.Errorf("Expected default database name 'business_verification', got %s", config.Database.Database)
	}

	if config.Database.MaxOpenConns != 25 {
		t.Errorf("Expected default max open connections 25, got %d", config.Database.MaxOpenConns)
	}
}

func TestAuthConfig(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("JWT_SECRET", "test-secret-for-testing")
	defer os.Unsetenv("JWT_SECRET")

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test auth config defaults
	if config.Auth.MinPasswordLength != 8 {
		t.Errorf("Expected default min password length 8, got %d", config.Auth.MinPasswordLength)
	}

	if !config.Auth.RequireUppercase {
		t.Error("Expected uppercase requirement to be enabled by default")
	}

	if !config.Auth.RequireLowercase {
		t.Error("Expected lowercase requirement to be enabled by default")
	}

	if !config.Auth.RequireNumbers {
		t.Error("Expected numbers requirement to be enabled by default")
	}

	if !config.Auth.RequireSpecial {
		t.Error("Expected special characters requirement to be enabled by default")
	}

	if config.Auth.MaxLoginAttempts != 5 {
		t.Errorf("Expected default max login attempts 5, got %d", config.Auth.MaxLoginAttempts)
	}

	if config.Auth.LockoutDuration != 15*time.Minute {
		t.Errorf("Expected default lockout duration 15m, got %v", config.Auth.LockoutDuration)
	}
}

func TestObservabilityConfig(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("JWT_SECRET", "test-secret-for-testing")
	defer os.Unsetenv("JWT_SECRET")

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test observability config defaults
	if config.Observability.LogLevel != "info" {
		t.Errorf("Expected default log level 'info', got %s", config.Observability.LogLevel)
	}

	if config.Observability.LogFormat != "json" {
		t.Errorf("Expected default log format 'json', got %s", config.Observability.LogFormat)
	}

	if !config.Observability.MetricsEnabled {
		t.Error("Expected metrics to be enabled by default")
	}

	if config.Observability.MetricsPort != 9090 {
		t.Errorf("Expected default metrics port 9090, got %d", config.Observability.MetricsPort)
	}

	if !config.Observability.TracingEnabled {
		t.Error("Expected tracing to be enabled by default")
	}

	if config.Observability.HealthCheckPath != "/health" {
		t.Errorf("Expected default health check path '/health', got %s", config.Observability.HealthCheckPath)
	}
}
