package security

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRateLimiter_Allow(t *testing.T) {
	logger := zap.NewNop()
	config := RateLimitConfig{
		RequestsPerMinute: 10,
		BurstAllowance:    5,
		WindowSize:        time.Minute,
		BlockDuration:     5 * time.Minute,
		EnableRedis:       false,
		EnableLocalCache:  true,
	}

	limiter := NewRateLimiter(nil, logger, config)

	tests := []struct {
		name     string
		requests int
		expected bool
	}{
		{
			name:     "first request allowed",
			requests: 1,
			expected: true,
		},
		{
			name:     "requests within limit allowed",
			requests: 5,
			expected: true,
		},
		{
			name:     "requests at limit allowed",
			requests: 10,
			expected: true,
		},
		{
			name:     "requests over limit denied",
			requests: 11,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientIP := "192.168.1.1"
			userAgent := "test-agent"

			// Make the specified number of requests
			for i := 0; i < tt.requests; i++ {
				result := limiter.Allow(context.Background(), clientIP, userAgent)

				if i < tt.requests-1 {
					// All requests except the last should be allowed
					assert.True(t, result.Allowed, "Request %d should be allowed", i+1)
				} else {
					// Check the last request
					assert.Equal(t, tt.expected, result.Allowed, "Last request should match expected result")
				}
			}
		})
	}
}

func TestRateLimiter_BlockClient(t *testing.T) {
	logger := zap.NewNop()
	config := RateLimitConfig{
		RequestsPerMinute: 2,
		BurstAllowance:    1,
		WindowSize:        time.Minute,
		BlockDuration:     5 * time.Minute,
		EnableRedis:       false,
		EnableLocalCache:  true,
	}

	limiter := NewRateLimiter(nil, logger, config)
	clientIP := "192.168.1.1"
	userAgent := "test-agent"

	// Make requests to trigger rate limiting
	for i := 0; i < 3; i++ {
		result := limiter.Allow(context.Background(), clientIP, userAgent)
		if i < 2 {
			assert.True(t, result.Allowed, "Request %d should be allowed", i+1)
		} else {
			assert.False(t, result.Allowed, "Request %d should be denied", i+1)
		}
	}

	// Make more requests to trigger blocking
	for i := 0; i < 3; i++ {
		result := limiter.Allow(context.Background(), clientIP, userAgent)
		assert.False(t, result.Allowed, "Request after rate limit should be denied")
	}

	// Check if client is blocked
	blocked := limiter.isClientBlocked(context.Background(), clientIP)
	assert.True(t, blocked, "Client should be blocked")
}

func TestRateLimiter_UnblockClient(t *testing.T) {
	logger := zap.NewNop()
	config := RateLimitConfig{
		RequestsPerMinute: 1,
		BurstAllowance:    1,
		WindowSize:        time.Minute,
		BlockDuration:     5 * time.Minute,
		EnableRedis:       false,
		EnableLocalCache:  true,
	}

	limiter := NewRateLimiter(nil, logger, config)
	clientIP := "192.168.1.1"
	userAgent := "test-agent"

	// Trigger blocking
	for i := 0; i < 5; i++ {
		limiter.Allow(context.Background(), clientIP, userAgent)
	}

	// Verify client is blocked
	blocked := limiter.isClientBlocked(context.Background(), clientIP)
	assert.True(t, blocked, "Client should be blocked")

	// Unblock client
	err := limiter.UnblockClient(context.Background(), clientIP)
	assert.NoError(t, err, "Unblocking client should not return error")

	// Verify client is unblocked
	blocked = limiter.isClientBlocked(context.Background(), clientIP)
	assert.False(t, blocked, "Client should be unblocked")
}

func TestRateLimiter_GetClientStats(t *testing.T) {
	logger := zap.NewNop()
	config := RateLimitConfig{
		RequestsPerMinute: 10,
		BurstAllowance:    5,
		WindowSize:        time.Minute,
		BlockDuration:     5 * time.Minute,
		EnableRedis:       false,
		EnableLocalCache:  true,
	}

	limiter := NewRateLimiter(nil, logger, config)
	clientIP := "192.168.1.1"
	userAgent := "test-agent"

	// Make some requests
	for i := 0; i < 3; i++ {
		limiter.Allow(context.Background(), clientIP, userAgent)
	}

	// Get client stats
	stats := limiter.GetClientStats(context.Background(), clientIP)
	assert.NotNil(t, stats, "Client stats should not be nil")
	assert.Equal(t, clientIP, stats.IP, "Client IP should match")
	assert.Equal(t, userAgent, stats.UserAgent, "User agent should match")
	assert.Equal(t, 3, stats.RequestCount, "Request count should be 3")
}

func TestRateLimiter_UpdateRateLimitConfig(t *testing.T) {
	logger := zap.NewNop()
	config := RateLimitConfig{
		RequestsPerMinute: 10,
		BurstAllowance:    5,
		WindowSize:        time.Minute,
		BlockDuration:     5 * time.Minute,
		EnableRedis:       false,
		EnableLocalCache:  true,
	}

	limiter := NewRateLimiter(nil, logger, config)

	// Update configuration
	newConfig := RateLimitConfig{
		RequestsPerMinute: 20,
		BurstAllowance:    10,
		WindowSize:        2 * time.Minute,
		BlockDuration:     10 * time.Minute,
		EnableRedis:       false,
		EnableLocalCache:  true,
	}

	limiter.UpdateRateLimitConfig(newConfig)

	// Verify configuration was updated
	updatedConfig := limiter.GetRateLimitConfig()
	assert.Equal(t, newConfig.RequestsPerMinute, updatedConfig.RequestsPerMinute)
	assert.Equal(t, newConfig.BurstAllowance, updatedConfig.BurstAllowance)
	assert.Equal(t, newConfig.WindowSize, updatedConfig.WindowSize)
	assert.Equal(t, newConfig.BlockDuration, updatedConfig.BlockDuration)
}

func TestRateLimiter_RedisIntegration(t *testing.T) {
	// Skip if Redis is not available
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available, skipping Redis integration test")
	}

	logger := zap.NewNop()
	config := RateLimitConfig{
		RequestsPerMinute: 5,
		BurstAllowance:    2,
		WindowSize:        time.Minute,
		BlockDuration:     5 * time.Minute,
		EnableRedis:       true,
		EnableLocalCache:  false,
	}

	limiter := NewRateLimiter(redisClient, logger, config)
	clientIP := "192.168.1.1"
	userAgent := "test-agent"

	// Make requests
	for i := 0; i < 6; i++ {
		result := limiter.Allow(ctx, clientIP, userAgent)
		if i < 5 {
			assert.True(t, result.Allowed, "Request %d should be allowed", i+1)
		} else {
			assert.False(t, result.Allowed, "Request %d should be denied", i+1)
		}
	}

	// Clean up
	redisClient.Del(ctx, "rate_limit:"+clientIP)
	redisClient.Del(ctx, "client:"+clientIP)
	redisClient.Del(ctx, "block:"+clientIP)
}

func TestRateLimiter_ConcurrentRequests(t *testing.T) {
	logger := zap.NewNop()
	config := RateLimitConfig{
		RequestsPerMinute: 100,
		BurstAllowance:    50,
		WindowSize:        time.Minute,
		BlockDuration:     5 * time.Minute,
		EnableRedis:       false,
		EnableLocalCache:  true,
	}

	limiter := NewRateLimiter(nil, logger, config)
	clientIP := "192.168.1.1"
	userAgent := "test-agent"

	// Make concurrent requests
	results := make(chan *RateLimitResult, 50)

	for i := 0; i < 50; i++ {
		go func() {
			result := limiter.Allow(context.Background(), clientIP, userAgent)
			results <- result
		}()
	}

	// Collect results
	allowedCount := 0
	for i := 0; i < 50; i++ {
		result := <-results
		if result.Allowed {
			allowedCount++
		}
	}

	// Should allow all requests since limit is 100 and we made 50
	assert.Equal(t, 50, allowedCount, "All concurrent requests should be allowed")
}

func TestRateLimiter_DifferentClients(t *testing.T) {
	logger := zap.NewNop()
	config := RateLimitConfig{
		RequestsPerMinute: 5,
		BurstAllowance:    2,
		WindowSize:        time.Minute,
		BlockDuration:     5 * time.Minute,
		EnableRedis:       false,
		EnableLocalCache:  true,
	}

	limiter := NewRateLimiter(nil, logger, config)
	userAgent := "test-agent"

	// Test with different client IPs
	clientIPs := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}

	for _, clientIP := range clientIPs {
		// Each client should be able to make 5 requests
		for i := 0; i < 6; i++ {
			result := limiter.Allow(context.Background(), clientIP, userAgent)
			if i < 5 {
				assert.True(t, result.Allowed, "Client %s request %d should be allowed", clientIP, i+1)
			} else {
				assert.False(t, result.Allowed, "Client %s request %d should be denied", clientIP, i+1)
			}
		}
	}
}

func TestRateLimiter_BlockExpiry(t *testing.T) {
	logger := zap.NewNop()
	config := RateLimitConfig{
		RequestsPerMinute: 1,
		BurstAllowance:    1,
		WindowSize:        time.Minute,
		BlockDuration:     100 * time.Millisecond, // Short block duration for testing
		EnableRedis:       false,
		EnableLocalCache:  true,
	}

	limiter := NewRateLimiter(nil, logger, config)
	clientIP := "192.168.1.1"
	userAgent := "test-agent"

	// Trigger blocking
	for i := 0; i < 5; i++ {
		limiter.Allow(context.Background(), clientIP, userAgent)
	}

	// Verify client is blocked
	blocked := limiter.isClientBlocked(context.Background(), clientIP)
	assert.True(t, blocked, "Client should be blocked")

	// Wait for block to expire
	time.Sleep(150 * time.Millisecond)

	// Verify client is no longer blocked
	blocked = limiter.isClientBlocked(context.Background(), clientIP)
	assert.False(t, blocked, "Client should no longer be blocked after expiry")
}

func TestRateLimiter_GetBlockRemainingTime(t *testing.T) {
	logger := zap.NewNop()
	config := RateLimitConfig{
		RequestsPerMinute: 1,
		BurstAllowance:    1,
		WindowSize:        time.Minute,
		BlockDuration:     5 * time.Minute,
		EnableRedis:       false,
		EnableLocalCache:  true,
	}

	limiter := NewRateLimiter(nil, logger, config)
	clientIP := "192.168.1.1"
	userAgent := "test-agent"

	// Trigger blocking
	for i := 0; i < 5; i++ {
		limiter.Allow(context.Background(), clientIP, userAgent)
	}

	// Get remaining block time
	remainingTime := limiter.getBlockRemainingTime(context.Background(), clientIP)
	assert.True(t, remainingTime > 0, "Remaining block time should be positive")
	assert.True(t, remainingTime <= config.BlockDuration, "Remaining time should not exceed block duration")
}
