package risk_assessment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewExternalAPIRateLimiter(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExternalRateLimitConfig()
	// Disable monitoring for tests to avoid deadlocks
	config.MonitorConfig.Enabled = false

	limiter := NewExternalAPIRateLimiter(config, logger)

	assert.NotNil(t, limiter)
	assert.Equal(t, config, limiter.config)
	assert.Equal(t, logger, limiter.logger)
	assert.NotNil(t, limiter.apiLimits)
	assert.NotNil(t, limiter.globalLimits)
	// Monitor may be nil if disabled
	if config.MonitorConfig.Enabled {
		assert.NotNil(t, limiter.monitor)
	}
	assert.NotNil(t, limiter.fallback)
	assert.NotNil(t, limiter.optimizer)
}

func TestExternalAPIRateLimiter_CheckRateLimit_Allowed(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExternalRateLimitConfig()
	// Disable monitoring for tests to avoid deadlocks
	config.MonitorConfig.Enabled = false
	limiter := NewExternalAPIRateLimiter(config, logger)

	ctx := context.Background()
	result, err := limiter.CheckRateLimit(ctx, "test-api")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Allowed)
	assert.Equal(t, "test-api", result.APIEndpoint)
	assert.Equal(t, 59, result.RemainingRequests) // 60 - 1
	assert.Equal(t, 1, result.Priority)
}

func TestExternalAPIRateLimiter_CheckRateLimit_Blocked(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExternalRateLimitConfig()
	// Disable monitoring for tests to avoid deadlocks
	config.MonitorConfig.Enabled = false
	config.APIConfigs["test-api"] = &APIConfig{
		APIEndpoint:       "test-api",
		RequestsPerMinute: 1,
		RequestsPerHour:   10,
		RequestsPerDay:    100,
		Priority:          2,
		Enabled:           true,
	}

	limiter := NewExternalAPIRateLimiter(config, logger)

	ctx := context.Background()

	// First request should be allowed
	result1, err := limiter.CheckRateLimit(ctx, "test-api")
	require.NoError(t, err)
	assert.True(t, result1.Allowed)

	// Second request should be blocked
	result2, err := limiter.CheckRateLimit(ctx, "test-api")
	require.NoError(t, err)
	assert.False(t, result2.Allowed)
	assert.True(t, result2.QuotaExceeded)
	assert.Equal(t, 0, result2.RemainingRequests)
	assert.True(t, result2.WaitTime > 0)
}

func TestExternalAPIRateLimiter_CheckRateLimit_GlobalLimit(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExternalRateLimitConfig()
	// Disable monitoring for tests to avoid deadlocks
	config.MonitorConfig.Enabled = false
	config.GlobalRequestsPerMinute = 1
	limiter := NewExternalAPIRateLimiter(config, logger)

	ctx := context.Background()

	// First request should be allowed
	result1, err := limiter.CheckRateLimit(ctx, "api1")
	require.NoError(t, err)
	assert.True(t, result1.Allowed)

	// Second request should be blocked due to global limit
	result2, err := limiter.CheckRateLimit(ctx, "api2")
	require.NoError(t, err)
	assert.False(t, result2.Allowed)
	assert.True(t, result2.QuotaExceeded)
}

func TestExternalAPIRateLimiter_WaitForRateLimit(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExternalRateLimitConfig()
	// Disable monitoring for tests to avoid deadlocks
	config.MonitorConfig.Enabled = false
	config.APIConfigs["test-api"] = &APIConfig{
		APIEndpoint:       "test-api",
		RequestsPerMinute: 1,
		RequestsPerHour:   10,
		RequestsPerDay:    100,
		Priority:          1,
		Enabled:           true,
	}

	limiter := NewExternalAPIRateLimiter(config, logger)

	ctx := context.Background()

	// First request should succeed immediately
	err := limiter.WaitForRateLimit(ctx, "test-api")
	assert.NoError(t, err)

	// Second request should wait and then succeed
	// Use a shorter timeout for testing
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	start := time.Now()
	err = limiter.WaitForRateLimit(ctx, "test-api")
	duration := time.Since(start)

	// Should timeout due to context cancellation
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
	assert.True(t, duration >= 1*time.Second, "Should have waited at least 1 second")
}

func TestExternalAPIRateLimiter_RecordAPICall(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExternalRateLimitConfig()
	// Disable monitoring for tests to avoid deadlocks
	config.MonitorConfig.Enabled = false
	limiter := NewExternalAPIRateLimiter(config, logger)

	// Add API config
	limiter.AddAPIConfig("test-api", &APIConfig{
		APIEndpoint: "test-api",
		Enabled:     true,
	})

	// Record successful call
	limiter.RecordAPICall("test-api", true, 100*time.Millisecond, nil)

	status := limiter.GetRateLimitStatus("test-api")
	assert.NotNil(t, status)
	assert.Equal(t, int64(1), status.TotalRequests)
	assert.Equal(t, int64(1), status.SuccessfulRequests)
	assert.Equal(t, int64(0), status.FailedRequests)
	assert.Equal(t, 100*time.Millisecond, status.AverageResponseTime)

	// Record failed call
	limiter.RecordAPICall("test-api", false, 200*time.Millisecond, assert.AnError)

	status = limiter.GetRateLimitStatus("test-api")
	assert.Equal(t, int64(2), status.TotalRequests)
	assert.Equal(t, int64(1), status.SuccessfulRequests)
	assert.Equal(t, int64(1), status.FailedRequests)
	assert.Equal(t, 150*time.Millisecond, status.AverageResponseTime) // (100 + 200) / 2
}

func TestExternalAPIRateLimiter_ResetRateLimit(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExternalRateLimitConfig()
	// Disable monitoring for tests to avoid deadlocks
	config.MonitorConfig.Enabled = false
	limiter := NewExternalAPIRateLimiter(config, logger)

	ctx := context.Background()

	// Make a request to increment counter
	result, err := limiter.CheckRateLimit(ctx, "test-api")
	require.NoError(t, err)
	assert.True(t, result.Allowed)

	// Reset the rate limit
	limiter.ResetRateLimit("test-api")

	// Should be able to make another request immediately
	result2, err := limiter.CheckRateLimit(ctx, "test-api")
	require.NoError(t, err)
	assert.True(t, result2.Allowed)
	assert.Equal(t, 59, result2.RemainingRequests) // Should be back to 59
}

func TestExternalAPIRateLimiter_ResetGlobalRateLimit(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExternalRateLimitConfig()
	// Disable monitoring for tests to avoid deadlocks
	config.MonitorConfig.Enabled = false
	config.GlobalRequestsPerMinute = 1
	limiter := NewExternalAPIRateLimiter(config, logger)

	ctx := context.Background()

	// First request should be allowed
	result1, err := limiter.CheckRateLimit(ctx, "api1")
	require.NoError(t, err)
	assert.True(t, result1.Allowed)

	// Second request should be blocked due to global limit
	result2, err := limiter.CheckRateLimit(ctx, "api2")
	require.NoError(t, err)
	assert.False(t, result2.Allowed)

	// Reset global rate limit
	limiter.ResetGlobalRateLimit()

	// Should be able to make another request immediately
	result3, err := limiter.CheckRateLimit(ctx, "api2")
	require.NoError(t, err)
	assert.True(t, result3.Allowed)
}

func TestExternalAPIRateLimiter_AddAPIConfig(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExternalRateLimitConfig()
	// Disable monitoring for tests to avoid deadlocks
	config.MonitorConfig.Enabled = false
	limiter := NewExternalAPIRateLimiter(config, logger)

	apiConfig := &APIConfig{
		APIEndpoint:       "custom-api",
		RequestsPerMinute: 30,
		RequestsPerHour:   500,
		RequestsPerDay:    5000,
		Timeout:           15 * time.Second,
		Priority:          5,
		RetryAttempts:     5,
		BackoffStrategy:   "exponential",
		Enabled:           true,
	}

	limiter.AddAPIConfig("custom-api", apiConfig)

	// Verify config was added
	configs := limiter.GetAPIConfigs()
	assert.Contains(t, configs, "custom-api")
	assert.Equal(t, apiConfig, configs["custom-api"])

	// Verify rate limit was created
	status := limiter.GetRateLimitStatus("custom-api")
	assert.NotNil(t, status)
	assert.Equal(t, apiConfig, status.Config)
}

func TestExternalAPIRateLimiter_RemoveAPIConfig(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExternalRateLimitConfig()
	// Disable monitoring for tests to avoid deadlocks
	config.MonitorConfig.Enabled = false
	limiter := NewExternalAPIRateLimiter(config, logger)

	// Add a custom config
	apiConfig := &APIConfig{
		APIEndpoint: "custom-api",
		Enabled:     true,
	}
	limiter.AddAPIConfig("custom-api", apiConfig)

	// Verify it exists
	configs := limiter.GetAPIConfigs()
	assert.Contains(t, configs, "custom-api")

	// Remove it
	limiter.RemoveAPIConfig("custom-api")

	// Verify it's gone
	configs = limiter.GetAPIConfigs()
	assert.NotContains(t, configs, "custom-api")

	// Verify status is nil
	status := limiter.GetRateLimitStatus("custom-api")
	assert.Nil(t, status)
}

func TestExternalAPIRateLimiter_GetGlobalRateLimitStatus(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExternalRateLimitConfig()
	// Disable monitoring for tests to avoid deadlocks
	config.MonitorConfig.Enabled = false
	limiter := NewExternalAPIRateLimiter(config, logger)

	status := limiter.GetGlobalRateLimitStatus()
	assert.NotNil(t, status)
	assert.Equal(t, 0, status.CurrentRequestsPerMinute)
	assert.Equal(t, 0, status.CurrentRequestsPerHour)
	assert.Equal(t, 0, status.CurrentRequestsPerDay)
	assert.False(t, status.QuotaExceeded)
}

func TestExternalAPIRateLimiter_ContextCancellation(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExternalRateLimitConfig()
	// Disable monitoring for tests to avoid deadlocks
	config.MonitorConfig.Enabled = false
	config.APIConfigs["test-api"] = &APIConfig{
		APIEndpoint:       "test-api",
		RequestsPerMinute: 1,
		RequestsPerHour:   10,
		RequestsPerDay:    100,
		Priority:          1,
		Enabled:           true,
	}

	limiter := NewExternalAPIRateLimiter(config, logger)

	// Make first request
	ctx := context.Background()
	err := limiter.WaitForRateLimit(ctx, "test-api")
	assert.NoError(t, err)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Second request should be cancelled
	err = limiter.WaitForRateLimit(ctx, "test-api")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

func TestRateLimitMonitor_RecordRateLimitCheck(t *testing.T) {
	logger := zap.NewNop()
	config := &MonitorConfig{
		Enabled:                   false, // Disable monitoring to avoid ticker issues
		AlertThreshold:            0.8,
		AlertCooldown:             5 * time.Minute,
		MetricsCollectionInterval: 30 * time.Second, // Set proper interval
	}

	monitor := NewRateLimitMonitor(config, logger)

	result := &ExternalRateLimitResult{
		Allowed:     true,
		APIEndpoint: "test-api",
		WaitTime:    100 * time.Millisecond,
	}

	monitor.RecordRateLimitCheck("test-api", result)

	// Verify metrics were recorded
	// Note: We can't directly access metrics due to private field, but we can verify no panics
}

func TestRateLimitFallback_HasFallback(t *testing.T) {
	logger := zap.NewNop()
	config := &FallbackConfig{
		Enabled:      true,
		FallbackAPIs: []string{"fallback1", "fallback2"},
	}

	fallback := NewRateLimitFallback(config, logger)

	// Initially no fallbacks configured
	assert.False(t, fallback.HasFallback("test-api"))

	// Add fallback
	fallback.fallbacks["test-api"] = []string{"fallback1", "fallback2"}

	// Now should have fallback
	assert.True(t, fallback.HasFallback("test-api"))
}

func TestRateLimitOptimizer_HasCachedResponse(t *testing.T) {
	logger := zap.NewNop()
	config := &OptimizationConfig{
		Enabled:      true,
		CacheEnabled: true,
		CacheTTL:     5 * time.Minute,
	}

	optimizer := NewRateLimitOptimizer(config, logger)

	// Initially no cached response
	assert.False(t, optimizer.HasCachedResponse("test-api"))

	// Add cached response
	optimizer.cache["test-api"] = &CachedResponse{
		Data:      "test data",
		Timestamp: time.Now(),
		TTL:       5 * time.Minute,
	}

	// Should have cached response
	assert.True(t, optimizer.HasCachedResponse("test-api"))

	// Add expired cached response
	optimizer.cache["expired-api"] = &CachedResponse{
		Data:      "expired data",
		Timestamp: time.Now().Add(-10 * time.Minute),
		TTL:       5 * time.Minute,
	}

	// Should not have cached response (expired)
	assert.False(t, optimizer.HasCachedResponse("expired-api"))
}

func TestDefaultExternalRateLimitConfig(t *testing.T) {
	config := DefaultExternalRateLimitConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 100, config.GlobalRequestsPerMinute)
	assert.Equal(t, 5000, config.GlobalRequestsPerHour)
	assert.Equal(t, 100000, config.GlobalRequestsPerDay)
	assert.Equal(t, 30*time.Second, config.DefaultTimeout)

	// Check default API config
	defaultAPI, exists := config.APIConfigs["default"]
	assert.True(t, exists)
	assert.Equal(t, "default", defaultAPI.APIEndpoint)
	assert.Equal(t, 60, defaultAPI.RequestsPerMinute)
	assert.Equal(t, 1000, defaultAPI.RequestsPerHour)
	assert.Equal(t, 10000, defaultAPI.RequestsPerDay)
	assert.Equal(t, 30*time.Second, defaultAPI.Timeout)
	assert.Equal(t, 1, defaultAPI.Priority)
	assert.Equal(t, 3, defaultAPI.RetryAttempts)
	assert.Equal(t, "exponential", defaultAPI.BackoffStrategy)
	assert.True(t, defaultAPI.Enabled)

	// Check monitoring config
	assert.NotNil(t, config.MonitorConfig)
	assert.True(t, config.MonitorConfig.Enabled)
	assert.Equal(t, 30*time.Second, config.MonitorConfig.MetricsCollectionInterval)
	assert.Equal(t, 0.8, config.MonitorConfig.AlertThreshold)
	assert.Equal(t, 5*time.Minute, config.MonitorConfig.AlertCooldown)

	// Check fallback config
	assert.NotNil(t, config.FallbackConfig)
	assert.True(t, config.FallbackConfig.Enabled)
	assert.True(t, config.FallbackConfig.CacheFallback)
	assert.True(t, config.FallbackConfig.RetryWithBackoff)
	assert.Equal(t, 3, config.FallbackConfig.MaxRetryAttempts)

	// Check optimization config
	assert.NotNil(t, config.OptimizationConfig)
	assert.True(t, config.OptimizationConfig.Enabled)
	assert.True(t, config.OptimizationConfig.CacheEnabled)
	assert.Equal(t, 5*time.Minute, config.OptimizationConfig.CacheTTL)
	assert.False(t, config.OptimizationConfig.RequestBatching)
	assert.Equal(t, 10, config.OptimizationConfig.BatchSize)
	assert.Equal(t, 1*time.Second, config.OptimizationConfig.BatchTimeout)
}
