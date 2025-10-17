package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Save original environment variables
	originalEnv := make(map[string]string)
	envVars := []string{
		"PORT", "HOST", "READ_TIMEOUT", "WRITE_TIMEOUT", "IDLE_TIMEOUT",
		"SUPABASE_URL", "SUPABASE_ANON_KEY", "SUPABASE_SERVICE_ROLE_KEY", "SUPABASE_JWT_SECRET",
		"REDIS_URL", "REDIS_PASSWORD", "REDIS_DB", "REDIS_POOL_SIZE",
		"ML_MODEL_PATH", "ML_TRAINING_DATA", "ML_BATCH_SIZE", "ML_LEARNING_RATE",
		"THOMSON_REUTERS_API_KEY", "OFAC_API_KEY", "WORLDCHECK_API_KEY",
		"NEWS_API_KEY", "OPENCORPORATES_API_KEY",
		"LOG_LEVEL", "LOG_FORMAT",
	}

	for _, envVar := range envVars {
		originalEnv[envVar] = os.Getenv(envVar)
	}

	// Clean up after test
	defer func() {
		for _, envVar := range envVars {
			if originalEnv[envVar] == "" {
				os.Unsetenv(envVar)
			} else {
				os.Setenv(envVar, originalEnv[envVar])
			}
		}
	}()

	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(*testing.T, *Config)
	}{
		{
			name:    "default configuration",
			envVars: map[string]string{},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "8080", cfg.Server.Port)
				assert.Equal(t, "0.0.0.0", cfg.Server.Host)
				assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout)
				assert.Equal(t, 30*time.Second, cfg.Server.WriteTimeout)
				assert.Equal(t, 60*time.Second, cfg.Server.IdleTimeout)

				assert.Equal(t, "redis://localhost:6379", cfg.Redis.URL)
				assert.Equal(t, 0, cfg.Redis.DB)
				assert.Equal(t, 50, cfg.Redis.PoolSize)
				assert.Equal(t, 10, cfg.Redis.MinIdleConns)
				assert.Equal(t, 20, cfg.Redis.MaxIdleConns)
				assert.Equal(t, 5*time.Second, cfg.Redis.DialTimeout)
				assert.Equal(t, 3*time.Second, cfg.Redis.ReadTimeout)
				assert.Equal(t, 3*time.Second, cfg.Redis.WriteTimeout)
				assert.Equal(t, 4*time.Second, cfg.Redis.PoolTimeout)
				assert.Equal(t, 5*time.Minute, cfg.Redis.IdleTimeout)
				assert.Equal(t, 3, cfg.Redis.MaxRetries)
				assert.Equal(t, 8*time.Millisecond, cfg.Redis.MinRetryBackoff)
				assert.Equal(t, 512*time.Millisecond, cfg.Redis.MaxRetryBackoff)
				assert.True(t, cfg.Redis.EnableFallback)
				assert.True(t, cfg.Redis.FallbackToMemory)
				assert.Equal(t, "ra:", cfg.Redis.KeyPrefix)

				assert.Equal(t, "./models", cfg.ML.ModelPath)
				assert.Equal(t, "./data", cfg.ML.TrainingData)
				assert.Equal(t, 32, cfg.ML.BatchSize)
				assert.Equal(t, 0.01, cfg.ML.LearningRate)
				assert.Equal(t, 1000, cfg.ML.MaxIterations)

				assert.Equal(t, "https://api.thomsonreuters.com", cfg.External.ThomsonReuters.BaseURL)
				assert.Equal(t, 30*time.Second, cfg.External.ThomsonReuters.Timeout)

				assert.Equal(t, "https://api.treasury.gov", cfg.External.OFAC.BaseURL)
				assert.Equal(t, 30*time.Second, cfg.External.OFAC.Timeout)

				assert.Equal(t, "https://api.worldcheck.com", cfg.External.WorldCheck.BaseURL)
				assert.Equal(t, 30*time.Second, cfg.External.WorldCheck.Timeout)

				assert.Equal(t, "https://newsapi.org/v2", cfg.External.NewsAPI.BaseURL)
				assert.Equal(t, 30*time.Second, cfg.External.NewsAPI.Timeout)

				assert.Equal(t, "https://api.opencorporates.com", cfg.External.OpenCorporates.BaseURL)
				assert.Equal(t, 30*time.Second, cfg.External.OpenCorporates.Timeout)

				assert.Equal(t, "info", cfg.Logging.Level)
				assert.Equal(t, "json", cfg.Logging.Format)
			},
		},
		{
			name: "custom configuration",
			envVars: map[string]string{
				"PORT":                      "9090",
				"HOST":                      "127.0.0.1",
				"READ_TIMEOUT":              "60s",
				"WRITE_TIMEOUT":             "60s",
				"IDLE_TIMEOUT":              "120s",
				"SUPABASE_URL":              "https://test.supabase.co",
				"SUPABASE_ANON_KEY":         "test-anon-key",
				"SUPABASE_SERVICE_ROLE_KEY": "test-service-role-key",
				"SUPABASE_JWT_SECRET":       "test-jwt-secret",
				"REDIS_URL":                 "redis://test:6379",
				"REDIS_PASSWORD":            "test-password",
				"REDIS_DB":                  "1",
				"REDIS_POOL_SIZE":           "100",
				"REDIS_MIN_IDLE_CONNS":      "20",
				"REDIS_MAX_IDLE_CONNS":      "40",
				"REDIS_DIAL_TIMEOUT":        "10s",
				"REDIS_READ_TIMEOUT":        "5s",
				"REDIS_WRITE_TIMEOUT":       "5s",
				"REDIS_POOL_TIMEOUT":        "8s",
				"REDIS_IDLE_TIMEOUT":        "10m",
				"REDIS_MAX_RETRIES":         "5",
				"REDIS_MIN_RETRY_BACKOFF":   "16ms",
				"REDIS_MAX_RETRY_BACKOFF":   "1s",
				"REDIS_ENABLE_FALLBACK":     "false",
				"REDIS_FALLBACK_TO_MEMORY":  "false",
				"REDIS_KEY_PREFIX":          "test:",
				"ML_MODEL_PATH":             "/custom/models",
				"ML_TRAINING_DATA":          "/custom/data",
				"ML_BATCH_SIZE":             "64",
				"ML_LEARNING_RATE":          "0.001",
				"ML_MAX_ITERATIONS":         "2000",
				"THOMSON_REUTERS_API_KEY":   "test-tr-key",
				"THOMSON_REUTERS_BASE_URL":  "https://test.tr.com",
				"THOMSON_REUTERS_TIMEOUT":   "60s",
				"OFAC_API_KEY":              "test-ofac-key",
				"OFAC_BASE_URL":             "https://test.ofac.com",
				"OFAC_TIMEOUT":              "60s",
				"WORLDCHECK_API_KEY":        "test-wc-key",
				"WORLDCHECK_BASE_URL":       "https://test.wc.com",
				"WORLDCHECK_TIMEOUT":        "60s",
				"NEWS_API_KEY":              "test-news-key",
				"NEWS_API_BASE_URL":         "https://test.news.com",
				"NEWS_API_TIMEOUT":          "60s",
				"OPENCORPORATES_API_KEY":    "test-oc-key",
				"OPENCORPORATES_BASE_URL":   "https://test.oc.com",
				"OPENCORPORATES_TIMEOUT":    "60s",
				"LOG_LEVEL":                 "debug",
				"LOG_FORMAT":                "text",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "9090", cfg.Server.Port)
				assert.Equal(t, "127.0.0.1", cfg.Server.Host)
				assert.Equal(t, 60*time.Second, cfg.Server.ReadTimeout)
				assert.Equal(t, 60*time.Second, cfg.Server.WriteTimeout)
				assert.Equal(t, 120*time.Second, cfg.Server.IdleTimeout)

				assert.Equal(t, "https://test.supabase.co", cfg.Supabase.URL)
				assert.Equal(t, "test-anon-key", cfg.Supabase.APIKey)
				assert.Equal(t, "test-service-role-key", cfg.Supabase.ServiceRoleKey)
				assert.Equal(t, "test-jwt-secret", cfg.Supabase.JWTSecret)

				assert.Equal(t, "redis://test:6379", cfg.Redis.URL)
				assert.Equal(t, "test-password", cfg.Redis.Password)
				assert.Equal(t, 1, cfg.Redis.DB)
				assert.Equal(t, 100, cfg.Redis.PoolSize)
				assert.Equal(t, 20, cfg.Redis.MinIdleConns)
				assert.Equal(t, 40, cfg.Redis.MaxIdleConns)
				assert.Equal(t, 10*time.Second, cfg.Redis.DialTimeout)
				assert.Equal(t, 5*time.Second, cfg.Redis.ReadTimeout)
				assert.Equal(t, 5*time.Second, cfg.Redis.WriteTimeout)
				assert.Equal(t, 8*time.Second, cfg.Redis.PoolTimeout)
				assert.Equal(t, 10*time.Minute, cfg.Redis.IdleTimeout)
				assert.Equal(t, 5, cfg.Redis.MaxRetries)
				assert.Equal(t, 16*time.Millisecond, cfg.Redis.MinRetryBackoff)
				assert.Equal(t, 1*time.Second, cfg.Redis.MaxRetryBackoff)
				assert.False(t, cfg.Redis.EnableFallback)
				assert.False(t, cfg.Redis.FallbackToMemory)
				assert.Equal(t, "test:", cfg.Redis.KeyPrefix)

				assert.Equal(t, "/custom/models", cfg.ML.ModelPath)
				assert.Equal(t, "/custom/data", cfg.ML.TrainingData)
				assert.Equal(t, 64, cfg.ML.BatchSize)
				assert.Equal(t, 0.001, cfg.ML.LearningRate)
				assert.Equal(t, 2000, cfg.ML.MaxIterations)

				assert.Equal(t, "test-tr-key", cfg.External.ThomsonReuters.APIKey)
				assert.Equal(t, "https://test.tr.com", cfg.External.ThomsonReuters.BaseURL)
				assert.Equal(t, 60*time.Second, cfg.External.ThomsonReuters.Timeout)

				assert.Equal(t, "test-ofac-key", cfg.External.OFAC.APIKey)
				assert.Equal(t, "https://test.ofac.com", cfg.External.OFAC.BaseURL)
				assert.Equal(t, 60*time.Second, cfg.External.OFAC.Timeout)

				assert.Equal(t, "test-wc-key", cfg.External.WorldCheck.APIKey)
				assert.Equal(t, "https://test.wc.com", cfg.External.WorldCheck.BaseURL)
				assert.Equal(t, 60*time.Second, cfg.External.WorldCheck.Timeout)

				assert.Equal(t, "test-news-key", cfg.External.NewsAPI.APIKey)
				assert.Equal(t, "https://test.news.com", cfg.External.NewsAPI.BaseURL)
				assert.Equal(t, 60*time.Second, cfg.External.NewsAPI.Timeout)

				assert.Equal(t, "test-oc-key", cfg.External.OpenCorporates.APIKey)
				assert.Equal(t, "https://test.oc.com", cfg.External.OpenCorporates.BaseURL)
				assert.Equal(t, 60*time.Second, cfg.External.OpenCorporates.Timeout)

				assert.Equal(t, "debug", cfg.Logging.Level)
				assert.Equal(t, "text", cfg.Logging.Format)
			},
		},
		{
			name: "invalid duration values",
			envVars: map[string]string{
				"READ_TIMEOUT":  "invalid-duration",
				"WRITE_TIMEOUT": "invalid-duration",
				"IDLE_TIMEOUT":  "invalid-duration",
			},
			validate: func(t *testing.T, cfg *Config) {
				// Should fall back to defaults
				assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout)
				assert.Equal(t, 30*time.Second, cfg.Server.WriteTimeout)
				assert.Equal(t, 60*time.Second, cfg.Server.IdleTimeout)
			},
		},
		{
			name: "invalid integer values",
			envVars: map[string]string{
				"REDIS_DB":        "invalid-int",
				"REDIS_POOL_SIZE": "invalid-int",
				"ML_BATCH_SIZE":   "invalid-int",
			},
			validate: func(t *testing.T, cfg *Config) {
				// Should fall back to defaults
				assert.Equal(t, 0, cfg.Redis.DB)
				assert.Equal(t, 50, cfg.Redis.PoolSize)
				assert.Equal(t, 32, cfg.ML.BatchSize)
			},
		},
		{
			name: "invalid float values",
			envVars: map[string]string{
				"ML_LEARNING_RATE": "invalid-float",
			},
			validate: func(t *testing.T, cfg *Config) {
				// Should fall back to defaults
				assert.Equal(t, 0.01, cfg.ML.LearningRate)
			},
		},
		{
			name: "invalid boolean values",
			envVars: map[string]string{
				"REDIS_ENABLE_FALLBACK":    "invalid-bool",
				"REDIS_FALLBACK_TO_MEMORY": "invalid-bool",
			},
			validate: func(t *testing.T, cfg *Config) {
				// Should fall back to defaults
				assert.True(t, cfg.Redis.EnableFallback)
				assert.True(t, cfg.Redis.FallbackToMemory)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables
			for _, envVar := range envVars {
				os.Unsetenv(envVar)
			}

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Load configuration
			cfg, err := Load()
			require.NoError(t, err)
			require.NotNil(t, cfg)

			// Validate configuration
			tt.validate(t, cfg)
		})
	}
}

func TestGetEnvAsString(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "environment variable set",
			key:          "TEST_STRING",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "environment variable not set",
			key:          "TEST_STRING_NOT_SET",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "environment variable empty",
			key:          "TEST_STRING_EMPTY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if provided
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnvAsString(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		expected     int
	}{
		{
			name:         "valid integer",
			key:          "TEST_INT",
			defaultValue: 10,
			envValue:     "42",
			expected:     42,
		},
		{
			name:         "invalid integer",
			key:          "TEST_INT_INVALID",
			defaultValue: 10,
			envValue:     "not-a-number",
			expected:     10,
		},
		{
			name:         "environment variable not set",
			key:          "TEST_INT_NOT_SET",
			defaultValue: 10,
			envValue:     "",
			expected:     10,
		},
		{
			name:         "zero value",
			key:          "TEST_INT_ZERO",
			defaultValue: 10,
			envValue:     "0",
			expected:     0,
		},
		{
			name:         "negative value",
			key:          "TEST_INT_NEGATIVE",
			defaultValue: 10,
			envValue:     "-5",
			expected:     -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if provided
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnvAsInt(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvAsFloat64(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue float64
		envValue     string
		expected     float64
	}{
		{
			name:         "valid float",
			key:          "TEST_FLOAT",
			defaultValue: 1.0,
			envValue:     "3.14",
			expected:     3.14,
		},
		{
			name:         "invalid float",
			key:          "TEST_FLOAT_INVALID",
			defaultValue: 1.0,
			envValue:     "not-a-float",
			expected:     1.0,
		},
		{
			name:         "environment variable not set",
			key:          "TEST_FLOAT_NOT_SET",
			defaultValue: 1.0,
			envValue:     "",
			expected:     1.0,
		},
		{
			name:         "zero value",
			key:          "TEST_FLOAT_ZERO",
			defaultValue: 1.0,
			envValue:     "0.0",
			expected:     0.0,
		},
		{
			name:         "negative value",
			key:          "TEST_FLOAT_NEGATIVE",
			defaultValue: 1.0,
			envValue:     "-2.5",
			expected:     -2.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if provided
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnvAsFloat64(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvAsDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue time.Duration
		envValue     string
		expected     time.Duration
	}{
		{
			name:         "valid duration",
			key:          "TEST_DURATION",
			defaultValue: 1 * time.Second,
			envValue:     "5s",
			expected:     5 * time.Second,
		},
		{
			name:         "invalid duration",
			key:          "TEST_DURATION_INVALID",
			defaultValue: 1 * time.Second,
			envValue:     "not-a-duration",
			expected:     1 * time.Second,
		},
		{
			name:         "environment variable not set",
			key:          "TEST_DURATION_NOT_SET",
			defaultValue: 1 * time.Second,
			envValue:     "",
			expected:     1 * time.Second,
		},
		{
			name:         "complex duration",
			key:          "TEST_DURATION_COMPLEX",
			defaultValue: 1 * time.Second,
			envValue:     "1h30m45s",
			expected:     1*time.Hour + 30*time.Minute + 45*time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if provided
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnvAsDuration(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvAsStringSlice(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue []string
		envValue     string
		expected     []string
	}{
		{
			name:         "valid string slice",
			key:          "TEST_STRING_SLICE",
			defaultValue: []string{"default"},
			envValue:     "a,b,c",
			expected:     []string{"a", "b", "c"},
		},
		{
			name:         "single value",
			key:          "TEST_STRING_SLICE_SINGLE",
			defaultValue: []string{"default"},
			envValue:     "single",
			expected:     []string{"single"},
		},
		{
			name:         "environment variable not set",
			key:          "TEST_STRING_SLICE_NOT_SET",
			defaultValue: []string{"default"},
			envValue:     "",
			expected:     []string{"default"},
		},
		{
			name:         "empty value",
			key:          "TEST_STRING_SLICE_EMPTY",
			defaultValue: []string{"default"},
			envValue:     "",
			expected:     []string{"default"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if provided
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnvAsStringSlice(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvAsBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		expected     bool
	}{
		{
			name:         "true value",
			key:          "TEST_BOOL_TRUE",
			defaultValue: false,
			envValue:     "true",
			expected:     true,
		},
		{
			name:         "false value",
			key:          "TEST_BOOL_FALSE",
			defaultValue: true,
			envValue:     "false",
			expected:     false,
		},
		{
			name:         "1 value",
			key:          "TEST_BOOL_1",
			defaultValue: false,
			envValue:     "1",
			expected:     true,
		},
		{
			name:         "0 value",
			key:          "TEST_BOOL_0",
			defaultValue: true,
			envValue:     "0",
			expected:     false,
		},
		{
			name:         "invalid value",
			key:          "TEST_BOOL_INVALID",
			defaultValue: true,
			envValue:     "invalid",
			expected:     true,
		},
		{
			name:         "environment variable not set",
			key:          "TEST_BOOL_NOT_SET",
			defaultValue: true,
			envValue:     "",
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if provided
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnvAsBool(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkLoad(b *testing.B) {
	// Clear environment variables
	envVars := []string{
		"PORT", "HOST", "SUPABASE_URL", "REDIS_URL", "ML_MODEL_PATH",
		"THOMSON_REUTERS_API_KEY", "LOG_LEVEL",
	}
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg, err := Load()
		if err != nil {
			b.Fatal(err)
		}
		_ = cfg
	}
}

func BenchmarkGetEnvAsString(b *testing.B) {
	os.Setenv("BENCHMARK_STRING", "test-value")
	defer os.Unsetenv("BENCHMARK_STRING")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getEnvAsString("BENCHMARK_STRING", "default")
	}
}

func BenchmarkGetEnvAsInt(b *testing.B) {
	os.Setenv("BENCHMARK_INT", "42")
	defer os.Unsetenv("BENCHMARK_INT")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getEnvAsInt("BENCHMARK_INT", 10)
	}
}

func BenchmarkGetEnvAsFloat64(b *testing.B) {
	os.Setenv("BENCHMARK_FLOAT", "3.14")
	defer os.Unsetenv("BENCHMARK_FLOAT")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getEnvAsFloat64("BENCHMARK_FLOAT", 1.0)
	}
}

func BenchmarkGetEnvAsDuration(b *testing.B) {
	os.Setenv("BENCHMARK_DURATION", "5s")
	defer os.Unsetenv("BENCHMARK_DURATION")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getEnvAsDuration("BENCHMARK_DURATION", 1*time.Second)
	}
}

func BenchmarkGetEnvAsBool(b *testing.B) {
	os.Setenv("BENCHMARK_BOOL", "true")
	defer os.Unsetenv("BENCHMARK_BOOL")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getEnvAsBool("BENCHMARK_BOOL", false)
	}
}
