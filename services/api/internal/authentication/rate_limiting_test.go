package authentication

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewEnhancedRateLimiter(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy:   StrategyRetry,
		GlobalRateLimit:   1000,
		ProviderRateLimit: 100,
		RetryStrategy: RetryStrategy{
			MaxRetries:        3,
			BaseDelay:         time.Second,
			MaxDelay:          30 * time.Second,
			BackoffMultiplier: 2.0,
			JitterFactor:      0.1,
		},
		CircuitBreakerConfig: CircuitBreakerConfig{
			FailureThreshold: 5,
			RecoveryTimeout:  30 * time.Second,
			HalfOpenMaxCalls: 3,
			SuccessThreshold: 0.8,
		},
		FallbackProviders: []FallbackProvider{
			{
				Name:        "fallback1",
				Priority:    1,
				SuccessRate: 0.9,
				IsAvailable: true,
			},
		},
	}

	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	assert.NotNil(t, limiter)
	assert.Equal(t, config, limiter.config)
	assert.Equal(t, logger, limiter.logger)
	assert.Equal(t, StrategyRetry, config.DefaultStrategy)
	assert.Len(t, limiter.fallbackProviders, 1)
	assert.Equal(t, "fallback1", limiter.fallbackProviders["fallback1"].Name)
}

func TestRegisterProvider(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy:   StrategyRetry,
		ProviderRateLimit: 100,
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider
	limiter.RegisterProvider("test-provider", 50, StrategyExponential)

	// Verify provider was registered
	assert.Contains(t, limiter.providers, "test-provider")
	assert.Equal(t, 50, limiter.providers["test-provider"].RequestsPerMinute)
	assert.Equal(t, StrategyExponential, limiter.strategies["test-provider"])

	// Verify circuit breaker was created
	assert.Contains(t, limiter.circuitBreakers, "test-provider")
	assert.Equal(t, StateClosed, limiter.circuitBreakers["test-provider"].State)
}

func TestCheckRateLimit_Allowed(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy:   StrategyRetry,
		ProviderRateLimit: 100,
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider
	limiter.RegisterProvider("test-provider", 10, StrategyRetry)

	// Check rate limit - should be allowed
	result, err := limiter.CheckRateLimit(context.Background(), "test-provider")
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, "test-provider", result.ProviderName)
	assert.Equal(t, 9, result.RemainingRequests)
	assert.Equal(t, StrategyRetry, result.Strategy)
}

func TestCheckRateLimit_Exceeded(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy:   StrategyRetry,
		ProviderRateLimit: 100,
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider with low limit
	limiter.RegisterProvider("test-provider", 2, StrategyRetry)

	// First request - should be allowed
	result1, err := limiter.CheckRateLimit(context.Background(), "test-provider")
	require.NoError(t, err)
	assert.True(t, result1.Allowed)

	// Second request - should be allowed
	result2, err := limiter.CheckRateLimit(context.Background(), "test-provider")
	require.NoError(t, err)
	assert.True(t, result2.Allowed)

	// Third request - should be blocked
	result3, err := limiter.CheckRateLimit(context.Background(), "test-provider")
	require.NoError(t, err)
	assert.False(t, result3.Allowed)
	assert.Equal(t, 0, result3.RemainingRequests)
	assert.True(t, result3.RetryAfter.After(time.Now()))
}

func TestCheckRateLimit_WithFallback(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy:   StrategyFallback,
		ProviderRateLimit: 100,
		FallbackProviders: []FallbackProvider{
			{
				Name:        "fallback1",
				Priority:    1,
				SuccessRate: 0.9,
				IsAvailable: true,
			},
			{
				Name:        "fallback2",
				Priority:    2,
				SuccessRate: 0.8,
				IsAvailable: true,
			},
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider with low limit
	limiter.RegisterProvider("test-provider", 1, StrategyFallback)

	// First request - should be allowed
	result1, err := limiter.CheckRateLimit(context.Background(), "test-provider")
	require.NoError(t, err)
	assert.True(t, result1.Allowed)

	// Second request - should be blocked but fallback available
	result2, err := limiter.CheckRateLimit(context.Background(), "test-provider")
	require.NoError(t, err)
	assert.False(t, result2.Allowed)
	assert.True(t, result2.FallbackAvailable)
	assert.NotNil(t, result2.FallbackProvider)
	assert.Equal(t, "fallback2", result2.FallbackProvider.Name) // Higher priority
}

func TestWaitForRateLimit(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy:   StrategyRetry,
		ProviderRateLimit: 100,
		RetryStrategy: RetryStrategy{
			MaxRetries:        2,
			BaseDelay:         10 * time.Millisecond,
			MaxDelay:          100 * time.Millisecond,
			BackoffMultiplier: 2.0,
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider with low limit
	limiter.RegisterProvider("test-provider", 1, StrategyRetry)

	// First request - should succeed immediately
	err := limiter.WaitForRateLimit(context.Background(), "test-provider")
	assert.NoError(t, err)

	// Second request - should wait and retry
	start := time.Now()
	err = limiter.WaitForRateLimit(context.Background(), "test-provider")
	assert.NoError(t, err)

	// Should have waited at least the base delay
	assert.True(t, time.Since(start) >= 10*time.Millisecond)
}

func TestExecuteWithFallback(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy:   StrategyFallback,
		ProviderRateLimit: 100,
		FallbackProviders: []FallbackProvider{
			{
				Name:        "fallback1",
				Priority:    1,
				SuccessRate: 0.9,
				IsAvailable: true,
			},
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider
	limiter.RegisterProvider("test-provider", 10, StrategyFallback)

	// Define test functions
	primaryFunc := func() (interface{}, error) {
		return "primary-result", nil
	}

	fallbackFuncs := map[string]func() (interface{}, error){
		"fallback1": func() (interface{}, error) {
			return "fallback-result", nil
		},
	}

	// Execute with fallback - should use primary
	result, err := limiter.ExecuteWithFallback(context.Background(), "test-provider", primaryFunc, fallbackFuncs)
	assert.NoError(t, err)
	assert.Equal(t, "primary-result", result)
}

func TestExecuteWithFallback_PrimaryFails(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy:   StrategyFallback,
		ProviderRateLimit: 100,
		FallbackProviders: []FallbackProvider{
			{
				Name:        "fallback1",
				Priority:    1,
				SuccessRate: 0.9,
				IsAvailable: true,
			},
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider
	limiter.RegisterProvider("test-provider", 10, StrategyFallback)

	// Define test functions
	primaryFunc := func() (interface{}, error) {
		return nil, fmt.Errorf("primary failed")
	}

	fallbackFuncs := map[string]func() (interface{}, error){
		"fallback1": func() (interface{}, error) {
			return "fallback-result", nil
		},
	}

	// Execute with fallback - should use fallback
	result, err := limiter.ExecuteWithFallback(context.Background(), "test-provider", primaryFunc, fallbackFuncs)
	assert.NoError(t, err)
	assert.Equal(t, "fallback-result", result)
}

func TestCircuitBreaker(t *testing.T) {
	// Create circuit breaker directly
	config := CircuitBreakerConfig{
		FailureThreshold: 3,
		RecoveryTimeout:  100 * time.Millisecond,
		HalfOpenMaxCalls: 2,
		SuccessThreshold: 0.8,
	}

	cb := &CircuitBreaker{
		ProviderName: "test-provider",
		State:        StateClosed,
		Config:       config,
	}

	// Test circuit breaker in closed state
	assert.Equal(t, StateClosed, cb.State)
	assert.True(t, cb.IsAllowed())

	// Record failures until circuit breaker opens
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
		t.Logf("After failure %d: State=%s, FailureCount=%d", i+1, cb.State, cb.FailureCount)
	}

	// Circuit breaker should be open
	assert.Equal(t, StateOpen, cb.State)
	assert.False(t, cb.IsAllowed())

	// Wait for recovery timeout
	time.Sleep(150 * time.Millisecond)

	// Circuit breaker should be half-open
	assert.True(t, cb.IsAllowed())

	// Record success to close circuit breaker
	cb.RecordSuccess()
	t.Logf("After success: State=%s, SuccessCount=%d, FailureCount=%d", cb.State, cb.SuccessCount, cb.FailureCount)
	assert.Equal(t, StateClosed, cb.State)
	assert.True(t, cb.IsAllowed())
}

func TestRetryWithBackoff(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy:   StrategyExponential,
		ProviderRateLimit: 100,
		RetryStrategy: RetryStrategy{
			MaxRetries:        2,
			BaseDelay:         10 * time.Millisecond,
			MaxDelay:          100 * time.Millisecond,
			BackoffMultiplier: 2.0,
			JitterFactor:      0.1,
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider
	limiter.RegisterProvider("test-provider", 10, StrategyExponential)

	// Define test functions that fail initially then succeed
	attemptCount := 0
	primaryFunc := func() (interface{}, error) {
		attemptCount++
		if attemptCount < 3 {
			return nil, fmt.Errorf("attempt %d failed", attemptCount)
		}
		return "success", nil
	}

	fallbackFuncs := map[string]func() (interface{}, error){
		"fallback1": func() (interface{}, error) {
			return nil, fmt.Errorf("fallback failed")
		},
	}

	// Execute with retry - should succeed after retries
	result, err := limiter.ExecuteWithFallback(context.Background(), "test-provider", primaryFunc, fallbackFuncs)
	assert.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 3, attemptCount)
}

func TestConcurrentRateLimiting(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy:   StrategyRetry,
		ProviderRateLimit: 100,
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider
	limiter.RegisterProvider("test-provider", 5, StrategyRetry)

	// Test concurrent requests
	var wg sync.WaitGroup
	results := make([]bool, 10)
	errors := make([]error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			result, err := limiter.CheckRateLimit(context.Background(), "test-provider")
			results[index] = result.Allowed
			errors[index] = err
		}(i)
	}

	wg.Wait()

	// Count allowed requests
	allowedCount := 0
	for _, allowed := range results {
		if allowed {
			allowedCount++
		}
	}

	// Should have exactly 5 allowed requests (the rate limit)
	assert.Equal(t, 5, allowedCount)

	// All errors should be nil
	for _, err := range errors {
		assert.NoError(t, err)
	}
}

func TestGetRateLimitStats(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy:   StrategyRetry,
		ProviderRateLimit: 100,
		FallbackProviders: []FallbackProvider{
			{
				Name:        "fallback1",
				Priority:    1,
				SuccessRate: 0.9,
				IsAvailable: true,
			},
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register providers
	limiter.RegisterProvider("provider1", 50, StrategyRetry)
	limiter.RegisterProvider("provider2", 75, StrategyExponential)

	// Make some requests
	limiter.CheckRateLimit(context.Background(), "provider1")
	limiter.CheckRateLimit(context.Background(), "provider2")

	// Get stats
	stats := limiter.GetRateLimitStats()

	// Verify stats structure
	assert.Equal(t, 2, stats["total_providers"])
	assert.Equal(t, 1, stats["fallback_providers"])
	assert.Equal(t, 2, stats["circuit_breakers"])
	assert.Equal(t, "retry", stats["default_strategy"])

	// Verify provider stats
	providers := stats["providers"].(map[string]interface{})
	assert.Contains(t, providers, "provider1")
	assert.Contains(t, providers, "provider2")

	// Verify fallback provider stats
	fallbackProviders := stats["fallback_provider_details"].(map[string]interface{})
	assert.Contains(t, fallbackProviders, "fallback1")

	// Verify circuit breaker stats
	circuitBreakers := stats["circuit_breaker_details"].(map[string]interface{})
	assert.Contains(t, circuitBreakers, "provider1")
	assert.Contains(t, circuitBreakers, "provider2")
}

func TestJitterCalculation(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy:   StrategyJitter,
		ProviderRateLimit: 100,
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	baseDelay := 100 * time.Millisecond
	jitterFactor := 0.1

	// Test jitter calculation multiple times
	for i := 0; i < 10; i++ {
		jitteredDelay := limiter.addJitter(baseDelay, jitterFactor)

		// Jittered delay should be within expected range
		minDelay := baseDelay
		maxDelay := baseDelay + time.Duration(float64(baseDelay)*jitterFactor)

		assert.True(t, jitteredDelay >= minDelay)
		assert.True(t, jitteredDelay <= maxDelay)
	}
}

func TestExponentialDelayCalculation(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy:   StrategyExponential,
		ProviderRateLimit: 100,
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	baseDelay := 100 * time.Millisecond
	multiplier := 2.0
	maxDelay := 1 * time.Second

	// Test exponential delay calculation
	delays := []time.Duration{
		limiter.calculateExponentialDelay(baseDelay, multiplier, 0, maxDelay),
		limiter.calculateExponentialDelay(baseDelay, multiplier, 1, maxDelay),
		limiter.calculateExponentialDelay(baseDelay, multiplier, 2, maxDelay),
		limiter.calculateExponentialDelay(baseDelay, multiplier, 3, maxDelay),
	}

	// Verify exponential growth
	assert.Equal(t, 100*time.Millisecond, delays[0])
	assert.Equal(t, 200*time.Millisecond, delays[1])
	assert.Equal(t, 400*time.Millisecond, delays[2])
	assert.Equal(t, 800*time.Millisecond, delays[3])

	// Test max delay cap
	delay := limiter.calculateExponentialDelay(baseDelay, multiplier, 10, maxDelay)
	assert.Equal(t, maxDelay, delay)
}

func TestContextCancellation(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy:   StrategyRetry,
		ProviderRateLimit: 100,
		RetryStrategy: RetryStrategy{
			MaxRetries:        5,
			BaseDelay:         100 * time.Millisecond,
			MaxDelay:          1 * time.Second,
			BackoffMultiplier: 2.0,
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider
	limiter.RegisterProvider("test-provider", 1, StrategyRetry)

	// Use first request
	limiter.CheckRateLimit(context.Background(), "test-provider")

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Try to wait for rate limit with cancelled context
	err := limiter.WaitForRateLimit(ctx, "test-provider")
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestFallbackProviderPriority(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy:   StrategyFallback,
		ProviderRateLimit: 100,
		FallbackProviders: []FallbackProvider{
			{
				Name:        "fallback1",
				Priority:    1,
				SuccessRate: 0.9,
				IsAvailable: true,
			},
			{
				Name:        "fallback2",
				Priority:    3,
				SuccessRate: 0.8,
				IsAvailable: true,
			},
			{
				Name:        "fallback3",
				Priority:    2,
				SuccessRate: 0.95,
				IsAvailable: true,
			},
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider
	limiter.RegisterProvider("test-provider", 1, StrategyFallback)

	// Use first request
	limiter.CheckRateLimit(context.Background(), "test-provider")

	// Check rate limit - should return highest priority fallback
	result, err := limiter.CheckRateLimit(context.Background(), "test-provider")
	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.True(t, result.FallbackAvailable)
	assert.Equal(t, "fallback2", result.FallbackProvider.Name) // Priority 3 (highest)
}

func TestFallbackProviderAvailability(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy: StrategyFallback,
		FallbackProviders: []FallbackProvider{
			{Name: "fallback1", Priority: 1, IsAvailable: true},
			{Name: "fallback2", Priority: 2, IsAvailable: false},
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register providers
	limiter.RegisterProvider("provider1", 10, StrategyFallback)
	limiter.RegisterProvider("provider2", 10, StrategyFallback)

	// Test fallback availability
	available := limiter.hasFallbackProvider("provider1")
	assert.True(t, available)

	// Test best fallback provider selection
	bestProvider := limiter.getBestFallbackProvider("provider1")
	assert.NotNil(t, bestProvider)
	assert.Equal(t, "fallback1", bestProvider.Name)
}

// Optimization and Caching Tests

func TestCheckRateLimitOptimized_WithCaching(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy: StrategyRetry,
		Optimization: OptimizationConfig{
			EnableCaching: true,
			CacheTTL:      5 * time.Second,
			CacheMaxSize:  100,
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider
	limiter.RegisterProvider("test-provider", 10, StrategyRetry)

	// First call should cache miss
	result1, err := limiter.CheckRateLimitOptimized(context.Background(), "test-provider")
	require.NoError(t, err)
	assert.True(t, result1.Allowed)

	// Second call should cache hit
	result2, err := limiter.CheckRateLimitOptimized(context.Background(), "test-provider")
	require.NoError(t, err)
	assert.True(t, result2.Allowed)
	assert.Equal(t, result1.RemainingRequests, result2.RemainingRequests)

	// Verify cache stats
	cacheStats := limiter.GetCacheStats()
	assert.Equal(t, 1, cacheStats["total_entries"])
	assert.Equal(t, 100, cacheStats["max_size"])
}

func TestCheckRateLimitOptimized_WithoutCaching(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy: StrategyRetry,
		Optimization: OptimizationConfig{
			EnableCaching: false,
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider
	limiter.RegisterProvider("test-provider", 10, StrategyRetry)

	// Both calls should be cache misses
	result1, err := limiter.CheckRateLimitOptimized(context.Background(), "test-provider")
	require.NoError(t, err)
	assert.True(t, result1.Allowed)

	result2, err := limiter.CheckRateLimitOptimized(context.Background(), "test-provider")
	require.NoError(t, err)
	assert.True(t, result2.Allowed)

	// Verify no cache entries
	cacheStats := limiter.GetCacheStats()
	assert.Equal(t, 0, cacheStats["total_entries"])
}

func TestPredictiveLimiting(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy: StrategyRetry,
		Optimization: OptimizationConfig{
			EnablePredictiveLimiting: true,
			PredictiveWindow:         1 * time.Minute,
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider with low limit
	limiter.RegisterProvider("test-provider", 5, StrategyRetry)

	// Manually set current requests to trigger predictive limiting (80% threshold)
	provider := limiter.providers["test-provider"]
	provider.CurrentRequests = 4 // 4/5 = 80% usage, should trigger predictive limiting

	// Test the predictive limiting logic directly
	predictiveResult := limiter.predictiveLimitCheck("test-provider")
	assert.NotNil(t, predictiveResult)
	assert.False(t, predictiveResult.Allowed) // Should predict rejection
	assert.Equal(t, StrategyFailFast, predictiveResult.Strategy)

	// Verify optimization stats
	optStats := limiter.GetOptimizationStats()
	providerStats := optStats["test-provider"].(map[string]interface{})
	assert.Greater(t, providerStats["predictive_hits"], int64(0))
}

func TestPredictiveLimiting_Rejection(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy: StrategyRetry,
		Optimization: OptimizationConfig{
			EnablePredictiveLimiting: true,
			PredictiveWindow:         1 * time.Minute,
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider with very low limit
	limiter.RegisterProvider("test-provider", 5, StrategyRetry)

	// Manually set current requests to exceed 80% threshold
	provider := limiter.providers["test-provider"]
	provider.CurrentRequests = 5 // 5/5 = 100% usage, should trigger predictive rejection

	// This request should be predictively rejected
	result := limiter.predictiveLimitCheck("test-provider")
	assert.NotNil(t, result)
	assert.False(t, result.Allowed)
	assert.Equal(t, StrategyFailFast, result.Strategy)

	// Verify optimization stats
	optStats := limiter.GetOptimizationStats()
	providerStats := optStats["test-provider"].(map[string]interface{})
	assert.Greater(t, providerStats["predictive_hits"], int64(0))
}

func TestAdaptiveLimiting(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy: StrategyRetry,
		Optimization: OptimizationConfig{
			EnableAdaptiveLimiting: true,
			AdaptiveThreshold:      0.8,
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider
	limiter.RegisterProvider("test-provider", 100, StrategyRetry)

	// Get initial rate limit
	initialResult, err := limiter.CheckRateLimitOptimized(context.Background(), "test-provider")
	require.NoError(t, err)
	initialLimit := initialResult.RemainingRequests

	// Simulate high success rate by recording successes
	provider := limiter.providers["test-provider"]
	provider.SuccessCount = 90
	provider.FailureCount = 10

	// Trigger adaptive adjustment
	limiter.adaptiveLimitAdjustment("test-provider")

	// Check that rate limit was adjusted
	result, err := limiter.CheckRateLimitOptimized(context.Background(), "test-provider")
	require.NoError(t, err)
	assert.Greater(t, result.RemainingRequests, initialLimit)

	// Verify optimization stats
	optStats := limiter.GetOptimizationStats()
	providerStats := optStats["test-provider"].(map[string]interface{})
	assert.Greater(t, providerStats["adaptive_adjustments"], int64(0))
}

func TestLoadBalancing_RoundRobin(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy: StrategyRetry,
		Optimization: OptimizationConfig{
			EnableLoadBalancing:   true,
			LoadBalancingStrategy: "round_robin",
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register multiple providers
	limiter.RegisterProvider("provider1", 10, StrategyRetry)
	limiter.RegisterProvider("provider2", 10, StrategyRetry)
	limiter.RegisterProvider("provider3", 10, StrategyRetry)

	// Test round-robin load balancing
	originalProvider := "provider1"
	balancedProvider1 := limiter.loadBalanceProvider(originalProvider)
	balancedProvider2 := limiter.loadBalanceProvider(balancedProvider1)
	balancedProvider3 := limiter.loadBalanceProvider(balancedProvider2)

	// Should cycle through providers
	assert.NotEqual(t, originalProvider, balancedProvider1)
	assert.NotEqual(t, balancedProvider1, balancedProvider2)
	assert.NotEqual(t, balancedProvider2, balancedProvider3)

	// Verify optimization stats
	optStats := limiter.GetOptimizationStats()
	providerStats := optStats["provider1"].(map[string]interface{})
	assert.Greater(t, providerStats["load_balanced_requests"], int64(0))
}

func TestLoadBalancing_LeastLoaded(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy: StrategyRetry,
		Optimization: OptimizationConfig{
			EnableLoadBalancing:   true,
			LoadBalancingStrategy: "least_loaded",
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register multiple providers
	limiter.RegisterProvider("provider1", 10, StrategyRetry)
	limiter.RegisterProvider("provider2", 10, StrategyRetry)

	// Make some requests to provider1 to increase its load
	provider1 := limiter.providers["provider1"]
	provider1.CurrentRequests = 8

	// Test least-loaded load balancing
	originalProvider := "provider1"
	balancedProvider := limiter.loadBalanceProvider(originalProvider)

	// Should select provider2 as it has lower load
	assert.Equal(t, "provider2", balancedProvider)
}

func TestRateShaping(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy:   StrategyRetry,
		ProviderRateLimit: 100,
		Optimization: OptimizationConfig{
			EnableRateShaping: true,
			RateShapingWindow: 1 * time.Second,
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider
	limiter.RegisterProvider("test-provider", 100, StrategyRetry)

	// Test rate shaping
	start := time.Now()
	limiter.rateShapeRequest("test-provider")
	duration := time.Since(start)

	// Should have some delay (though minimal for this test)
	assert.Greater(t, duration, time.Duration(0))

	// Verify optimization stats
	optStats := limiter.GetOptimizationStats()
	providerStats := optStats["test-provider"].(map[string]interface{})
	assert.Greater(t, providerStats["rate_shaped_requests"], int64(0))
}

func TestCacheEviction(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy: StrategyRetry,
		Optimization: OptimizationConfig{
			EnableCaching: true,
			CacheTTL:      5 * time.Second,
			CacheMaxSize:  2, // Small cache to trigger eviction
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register providers
	limiter.RegisterProvider("provider1", 10, StrategyRetry)
	limiter.RegisterProvider("provider2", 10, StrategyRetry)
	limiter.RegisterProvider("provider3", 10, StrategyRetry)

	// Fill cache
	limiter.CheckRateLimitOptimized(context.Background(), "provider1")
	limiter.CheckRateLimitOptimized(context.Background(), "provider2")

	// This should trigger eviction
	limiter.CheckRateLimitOptimized(context.Background(), "provider3")

	// Verify cache size is maintained
	cacheStats := limiter.GetCacheStats()
	assert.Equal(t, 2, cacheStats["total_entries"])
}

func TestCacheExpiration(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy: StrategyRetry,
		Optimization: OptimizationConfig{
			EnableCaching: true,
			CacheTTL:      10 * time.Millisecond, // Very short TTL
			CacheMaxSize:  100,
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider
	limiter.RegisterProvider("test-provider", 10, StrategyRetry)

	// First call
	result1, err := limiter.CheckRateLimitOptimized(context.Background(), "test-provider")
	require.NoError(t, err)
	assert.True(t, result1.Allowed)

	// Wait for cache to expire
	time.Sleep(20 * time.Millisecond)

	// Second call should be cache miss due to expiration
	result2, err := limiter.CheckRateLimitOptimized(context.Background(), "test-provider")
	require.NoError(t, err)
	assert.True(t, result2.Allowed)

	// Verify cache miss was recorded
	optStats := limiter.GetOptimizationStats()
	providerStats := optStats["test-provider"].(map[string]interface{})
	assert.Greater(t, providerStats["cache_misses"], int64(0))
}

func TestGetOptimizationStats(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy: StrategyRetry,
		Optimization: OptimizationConfig{
			EnableCaching:            true,
			EnablePredictiveLimiting: true,
			EnableAdaptiveLimiting:   true,
			EnableLoadBalancing:      true,
			EnableRateShaping:        true,
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider
	limiter.RegisterProvider("test-provider", 10, StrategyRetry)

	// Generate some activity
	limiter.CheckRateLimitOptimized(context.Background(), "test-provider")
	limiter.rateShapeRequest("test-provider")
	limiter.loadBalanceProvider("test-provider")

	// Get optimization stats
	stats := limiter.GetOptimizationStats()
	assert.Contains(t, stats, "test-provider")

	providerStats := stats["test-provider"].(map[string]interface{})
	assert.Contains(t, providerStats, "cache_hits")
	assert.Contains(t, providerStats, "cache_misses")
	assert.Contains(t, providerStats, "cache_hit_rate")
	assert.Contains(t, providerStats, "predictive_hits")
	assert.Contains(t, providerStats, "adaptive_adjustments")
	assert.Contains(t, providerStats, "load_balanced_requests")
	assert.Contains(t, providerStats, "rate_shaped_requests")
	assert.Contains(t, providerStats, "last_updated")
}

func TestClearCache(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy: StrategyRetry,
		Optimization: OptimizationConfig{
			EnableCaching: true,
			CacheTTL:      5 * time.Second,
			CacheMaxSize:  100,
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register provider
	limiter.RegisterProvider("test-provider", 10, StrategyRetry)

	// Add some cache entries
	limiter.CheckRateLimitOptimized(context.Background(), "test-provider")

	// Verify cache has entries
	cacheStats := limiter.GetCacheStats()
	assert.Equal(t, 1, cacheStats["total_entries"])

	// Clear cache
	limiter.ClearCache()

	// Verify cache is empty
	cacheStats = limiter.GetCacheStats()
	assert.Equal(t, 0, cacheStats["total_entries"])
}

func TestConcurrentOptimization(t *testing.T) {
	config := &EnhancedRateLimitConfig{
		DefaultStrategy: StrategyRetry,
		Optimization: OptimizationConfig{
			EnableCaching:            true,
			EnablePredictiveLimiting: true,
			EnableAdaptiveLimiting:   true,
			EnableLoadBalancing:      true,
			EnableRateShaping:        true,
		},
	}
	logger := zap.NewNop()
	limiter := NewEnhancedRateLimiter(config, logger)

	// Register multiple providers
	limiter.RegisterProvider("provider1", 100, StrategyRetry)
	limiter.RegisterProvider("provider2", 100, StrategyRetry)
	limiter.RegisterProvider("provider3", 100, StrategyRetry)

	// Run concurrent requests
	var wg sync.WaitGroup
	numGoroutines := 10
	numRequests := 5

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numRequests; j++ {
				providerName := fmt.Sprintf("provider%d", (id+j)%3+1)
				result, err := limiter.CheckRateLimitOptimized(context.Background(), providerName)
				assert.NoError(t, err)
				assert.True(t, result.Allowed)
			}
		}(i)
	}

	wg.Wait()

	// Verify all providers have optimization stats
	optStats := limiter.GetOptimizationStats()
	assert.Contains(t, optStats, "provider1")
	assert.Contains(t, optStats, "provider2")
	assert.Contains(t, optStats, "provider3")
}
