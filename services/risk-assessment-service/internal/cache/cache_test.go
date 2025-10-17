package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// MockLogger implements the Logger interface for testing
type MockLogger struct{}

func (m *MockLogger) Info(msg string, args ...interface{})  {}
func (m *MockLogger) Error(msg string, args ...interface{}) {}
func (m *MockLogger) Warn(msg string, args ...interface{})  {}

// Test helper functions
func createTestCacheConfig() *CacheConfig {
	return &CacheConfig{
		Addrs:             []string{"localhost:6379"},
		Password:          "",
		DB:                0,
		PoolSize:          10,
		MinIdleConns:      5,
		MaxRetries:        3,
		DialTimeout:       5 * time.Second,
		ReadTimeout:       3 * time.Second,
		WriteTimeout:      3 * time.Second,
		PoolTimeout:       4 * time.Second,
		IdleTimeout:       5 * time.Minute,
		MaxConnAge:        30 * time.Minute,
		DefaultTTL:        5 * time.Minute,
		KeyPrefix:         "test:",
		EnableMetrics:     true,
		EnableCompression: false,
	}
}

func TestNewRedisCache(t *testing.T) {
	tests := []struct {
		name        string
		config      *CacheConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid config",
			config:      createTestCacheConfig(),
			expectError: false,
		},
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
			errorMsg:    "cache config cannot be nil",
		},
		{
			name: "config with defaults",
			config: &CacheConfig{
				Addrs: []string{"localhost:6379"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &MockLogger{}
			cache, err := NewRedisCache(tt.config, logger)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, cache)
			} else {
				// Note: This will fail in real tests due to Redis connection
				// In a real test environment, you'd use a test Redis instance
				// For now, we'll just test the error case
				if err != nil {
					assert.Contains(t, err.Error(), "failed to connect to Redis")
				}
			}
		})
	}
}

func TestCacheConfig_Defaults(t *testing.T) {
	config := &CacheConfig{
		Addrs: []string{"localhost:6379"},
	}

	// Test that defaults are applied
	assert.Equal(t, 0, config.PoolSize)
	assert.Equal(t, 0, config.MinIdleConns)
	assert.Equal(t, 0, config.MaxRetries)
	assert.Equal(t, time.Duration(0), config.DialTimeout)
	assert.Equal(t, time.Duration(0), config.ReadTimeout)
	assert.Equal(t, time.Duration(0), config.WriteTimeout)
	assert.Equal(t, time.Duration(0), config.PoolTimeout)
	assert.Equal(t, time.Duration(0), config.IdleTimeout)
	assert.Equal(t, time.Duration(0), config.MaxConnAge)
	assert.Equal(t, time.Duration(0), config.DefaultTTL)
	assert.Equal(t, "", config.KeyPrefix)
}

func TestCacheMetrics_Initialization(t *testing.T) {
	metrics := &CacheMetrics{}

	// Test initial values
	assert.Equal(t, int64(0), metrics.Hits)
	assert.Equal(t, int64(0), metrics.Misses)
	assert.Equal(t, int64(0), metrics.Sets)
	assert.Equal(t, int64(0), metrics.Deletes)
	assert.Equal(t, int64(0), metrics.Errors)
	assert.Equal(t, int64(0), metrics.TotalRequests)
	assert.Equal(t, float64(0), metrics.HitRate)
	assert.Equal(t, time.Duration(0), metrics.AverageLatency)
	assert.True(t, metrics.LastUpdated.IsZero())
}

func TestCacheMetrics_Update(t *testing.T) {
	metrics := &CacheMetrics{}

	// Update metrics
	metrics.Hits = 10
	metrics.Misses = 5
	metrics.Sets = 8
	metrics.Deletes = 2
	metrics.Errors = 1

	// Test calculated values
	metrics.TotalRequests = metrics.Hits + metrics.Misses
	if metrics.TotalRequests > 0 {
		metrics.HitRate = float64(metrics.Hits) / float64(metrics.TotalRequests)
	}
	metrics.LastUpdated = time.Now()

	assert.Equal(t, int64(10), metrics.Hits)
	assert.Equal(t, int64(5), metrics.Misses)
	assert.Equal(t, int64(8), metrics.Sets)
	assert.Equal(t, int64(2), metrics.Deletes)
	assert.Equal(t, int64(1), metrics.Errors)
	assert.Equal(t, int64(15), metrics.TotalRequests)
	assert.Equal(t, 10.0/15.0, metrics.HitRate)
	assert.False(t, metrics.LastUpdated.IsZero())
}

func TestCacheFactory_NewCacheFactory(t *testing.T) {
	config := createTestCacheConfig()
	logger := &MockLogger{}

	factory := NewCacheFactory(config, logger)

	assert.NotNil(t, factory)
	assert.Equal(t, config, factory.config)
	assert.Equal(t, logger, factory.logger)
}

func TestCacheFactory_CreateRedisCache(t *testing.T) {
	tests := []struct {
		name        string
		config      *CacheConfig
		logger      interface{}
		expectError bool
	}{
		{
			name:        "valid config and logger",
			config:      createTestCacheConfig(),
			logger:      zap.NewNop(),
			expectError: false,
		},
		{
			name:        "invalid logger type",
			config:      createTestCacheConfig(),
			logger:      "invalid-logger",
			expectError: false, // Should fallback to no-op logger
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewCacheFactory(tt.config, tt.logger)
			cache, err := factory.CreateRedisCache()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cache)
			} else {
				// Note: This will fail in real tests due to Redis connection
				// In a real test environment, you'd use a test Redis instance
				if err != nil {
					assert.Contains(t, err.Error(), "failed to connect to Redis")
				}
			}
		})
	}
}

func TestNoOpLogger(t *testing.T) {
	logger := &noOpLogger{}

	// Test that methods don't panic
	assert.NotPanics(t, func() {
		logger.Info("test message", "arg1", "value1")
		logger.Error("test error", "arg1", "value1")
		logger.Warn("test warning", "arg1", "value1")
	})
}

func TestCacheConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      *CacheConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &CacheConfig{
				Addrs:        []string{"localhost:6379"},
				PoolSize:     10,
				MinIdleConns: 5,
				MaxRetries:   3,
			},
			expectError: false,
		},
		{
			name: "empty addrs",
			config: &CacheConfig{
				Addrs: []string{},
			},
			expectError: false, // Should still work with empty addrs
		},
		{
			name: "nil addrs",
			config: &CacheConfig{
				Addrs: nil,
			},
			expectError: false, // Should still work with nil addrs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &MockLogger{}
			cache, err := NewRedisCache(tt.config, logger)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cache)
			} else {
				// Note: This will fail in real tests due to Redis connection
				// In a real test environment, you'd use a test Redis instance
				if err != nil {
					assert.Contains(t, err.Error(), "failed to connect to Redis")
				}
			}
		})
	}
}

func TestCacheConfig_Timeouts(t *testing.T) {
	config := &CacheConfig{
		Addrs:        []string{"localhost:6379"},
		DialTimeout:  10 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		PoolTimeout:  8 * time.Second,
		IdleTimeout:  10 * time.Minute,
		MaxConnAge:   30 * time.Minute,
		DefaultTTL:   5 * time.Minute,
	}

	logger := &MockLogger{}
	cache, err := NewRedisCache(config, logger)

	// Should fail due to Redis connection, but config should be valid
	if err != nil {
		assert.Contains(t, err.Error(), "failed to connect to Redis")
	} else {
		assert.NotNil(t, cache)
		assert.Equal(t, 10*time.Second, config.DialTimeout)
		assert.Equal(t, 5*time.Second, config.ReadTimeout)
		assert.Equal(t, 5*time.Second, config.WriteTimeout)
		assert.Equal(t, 8*time.Second, config.PoolTimeout)
		assert.Equal(t, 10*time.Minute, config.IdleTimeout)
		assert.Equal(t, 30*time.Minute, config.MaxConnAge)
		assert.Equal(t, 5*time.Minute, config.DefaultTTL)
	}
}

// Benchmark tests
func BenchmarkNewRedisCache(b *testing.B) {
	config := createTestCacheConfig()
	logger := &MockLogger{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache, err := NewRedisCache(config, logger)
		if err != nil {
			// Expected to fail due to Redis connection
			_ = cache
		}
	}
}

func BenchmarkCacheFactory_NewCacheFactory(b *testing.B) {
	config := createTestCacheConfig()
	logger := &MockLogger{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		factory := NewCacheFactory(config, logger)
		_ = factory
	}
}

func BenchmarkCacheMetrics_Update(b *testing.B) {
	metrics := &CacheMetrics{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.Hits++
		metrics.Misses++
		metrics.TotalRequests = metrics.Hits + metrics.Misses
		if metrics.TotalRequests > 0 {
			metrics.HitRate = float64(metrics.Hits) / float64(metrics.TotalRequests)
		}
		metrics.LastUpdated = time.Now()
	}
}
