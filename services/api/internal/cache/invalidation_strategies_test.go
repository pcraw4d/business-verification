package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestTimeBasedInvalidationStrategy tests the time-based invalidation strategy
func TestTimeBasedInvalidationStrategy(t *testing.T) {
	logger := zap.NewNop()
	ttl := 5 * time.Minute

	strategy := NewTimeBasedInvalidationStrategy(ttl, logger)

	assert.Equal(t, "time_based", strategy.GetName())
	assert.Contains(t, strategy.GetDescription(), "5m0s")

	// Create a memory cache for testing
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         100,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	cache := NewMemoryCache(config)
	defer cache.Close()

	ctx := context.Background()

	// Set a key first
	key := "test:key"
	value := []byte("test value")
	err := cache.Set(ctx, key, value, 1*time.Hour)
	require.NoError(t, err)

	// Apply time-based invalidation
	err = strategy.Invalidate(ctx, cache, key)
	require.NoError(t, err)

	// Check that TTL was set
	retrievedTTL, err := cache.GetTTL(ctx, key)
	require.NoError(t, err)
	assert.True(t, retrievedTTL > 0)
	assert.True(t, retrievedTTL <= ttl)
}

// TestEventBasedInvalidationStrategy tests the event-based invalidation strategy
func TestEventBasedInvalidationStrategy(t *testing.T) {
	logger := zap.NewNop()
	strategy := NewEventBasedInvalidationStrategy(logger)

	assert.Equal(t, "event_based", strategy.GetName())
	assert.Equal(t, "Invalidates cache entries based on events", strategy.GetDescription())

	// Create a memory cache for testing
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         100,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	cache := NewMemoryCache(config)
	defer cache.Close()

	ctx := context.Background()

	// Set up event handler
	handlerCalled := false
	handler := func(ctx context.Context, cache Cache, key string) error {
		handlerCalled = true
		return cache.Delete(ctx, key)
	}

	strategy.RegisterEventHandler("invalidate", handler)

	// Set a key
	key := "test:key"
	value := []byte("test value")
	err := cache.Set(ctx, key, value, 1*time.Hour)
	require.NoError(t, err)

	// Trigger invalidation event
	err = strategy.Invalidate(ctx, cache, key)
	require.NoError(t, err)

	// Verify handler was called and key was deleted
	assert.True(t, handlerCalled)

	exists, err := cache.Exists(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)
}

// TestPatternBasedInvalidationStrategy tests the pattern-based invalidation strategy
func TestPatternBasedInvalidationStrategy(t *testing.T) {
	logger := zap.NewNop()
	strategy := NewPatternBasedInvalidationStrategy(logger)

	assert.Equal(t, "pattern_based", strategy.GetName())
	assert.Equal(t, "Invalidates cache entries based on key patterns", strategy.GetDescription())

	// Add patterns
	strategy.AddPattern("test:*", 5*time.Minute)
	strategy.AddPattern("user:*", 10*time.Minute)

	// Create a memory cache for testing
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         100,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	cache := NewMemoryCache(config)
	defer cache.Close()

	ctx := context.Background()

	// Test pattern matching
	testCases := []struct {
		key         string
		pattern     string
		shouldMatch bool
	}{
		{"test:key1", "test:*", true},
		{"test:key2", "test:*", true},
		{"user:123", "user:*", true},
		{"other:key", "test:*", false},
		{"other:key", "user:*", false},
		{"exact:match", "exact:match", true},
		{"exact:match", "exact:nomatch", false},
	}

	for _, tc := range testCases {
		// Set a key
		err := cache.Set(ctx, tc.key, []byte("value"), 1*time.Hour)
		require.NoError(t, err)

		// Apply pattern-based invalidation
		err = strategy.Invalidate(ctx, cache, tc.key)
		require.NoError(t, err)

		// Check TTL if pattern should match
		if tc.shouldMatch {
			ttl, err := cache.GetTTL(ctx, tc.key)
			require.NoError(t, err)
			assert.True(t, ttl > 0)
		}

		// Clean up
		cache.Delete(ctx, tc.key)
	}

	// Test removing pattern
	strategy.RemovePattern("test:*")

	// Set a key and test that pattern no longer matches
	key := "test:key"
	err := cache.Set(ctx, key, []byte("value"), 1*time.Hour)
	require.NoError(t, err)

	err = strategy.Invalidate(ctx, cache, key)
	require.NoError(t, err)

	// TTL should not have changed since pattern was removed
	ttl, err := cache.GetTTL(ctx, key)
	require.NoError(t, err)
	// Allow for small timing differences due to precision
	assert.True(t, ttl >= 59*time.Minute && ttl <= 1*time.Hour, "TTL should be approximately 1 hour, got %v", ttl)
}

// TestLRUBasedInvalidationStrategy tests the LRU-based invalidation strategy
func TestLRUBasedInvalidationStrategy(t *testing.T) {
	logger := zap.NewNop()
	maxSize := 5
	strategy := NewLRUBasedInvalidationStrategy(maxSize, logger)

	assert.Equal(t, "lru_based", strategy.GetName())
	assert.Contains(t, strategy.GetDescription(), "5")

	// Create a memory cache for testing
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         maxSize,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	cache := NewMemoryCache(config)
	defer cache.Close()

	ctx := context.Background()

	// Test invalidation (should not error even when cache is at capacity)
	key := "test:key"
	err := cache.Set(ctx, key, []byte("value"), 1*time.Hour)
	require.NoError(t, err)

	err = strategy.Invalidate(ctx, cache, key)
	require.NoError(t, err)
}

// TestCompositeInvalidationStrategy tests the composite invalidation strategy
func TestCompositeInvalidationStrategy(t *testing.T) {
	logger := zap.NewNop()

	// Create individual strategies
	timeStrategy := NewTimeBasedInvalidationStrategy(5*time.Minute, logger)
	eventStrategy := NewEventBasedInvalidationStrategy(logger)

	// Create composite strategy
	strategies := []InvalidationStrategy{timeStrategy, eventStrategy}
	composite := NewCompositeInvalidationStrategy(strategies, logger)

	assert.Equal(t, "composite", composite.GetName())
	assert.Contains(t, composite.GetDescription(), "2")

	// Test adding strategy
	lruStrategy := NewLRUBasedInvalidationStrategy(10, logger)
	composite.AddStrategy(lruStrategy)

	assert.Len(t, composite.GetStrategies(), 3)

	// Test removing strategy
	composite.RemoveStrategy("lru_based")
	assert.Len(t, composite.GetStrategies(), 2)

	// Create a memory cache for testing
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         100,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	cache := NewMemoryCache(config)
	defer cache.Close()

	ctx := context.Background()

	// Set up event handler for event strategy
	handlerCalled := false
	handler := func(ctx context.Context, cache Cache, key string) error {
		handlerCalled = true
		return nil // Don't delete, just mark as called
	}
	eventStrategy.RegisterEventHandler("invalidate", handler)

	// Set a key
	key := "test:key"
	value := []byte("test value")
	err := cache.Set(ctx, key, value, 1*time.Hour)
	require.NoError(t, err)

	// Apply composite invalidation
	err = composite.Invalidate(ctx, cache, key)
	require.NoError(t, err)

	// Verify both strategies were applied
	assert.True(t, handlerCalled) // Event strategy was called

	// Time strategy should have set TTL
	ttl, err := cache.GetTTL(ctx, key)
	require.NoError(t, err)
	assert.True(t, ttl > 0)
	assert.True(t, ttl <= 5*time.Minute)
}

// TestInvalidationStrategyIntegration tests integration of invalidation strategies
func TestInvalidationStrategyIntegration(t *testing.T) {
	logger := zap.NewNop()

	// Create a memory cache
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         100,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	cache := NewMemoryCache(config)
	defer cache.Close()

	ctx := context.Background()

	// Create and configure strategies
	timeStrategy := NewTimeBasedInvalidationStrategy(2*time.Minute, logger)
	patternStrategy := NewPatternBasedInvalidationStrategy(logger)
	eventStrategy := NewEventBasedInvalidationStrategy(logger)

	// Add patterns
	patternStrategy.AddPattern("user:*", 1*time.Minute)
	patternStrategy.AddPattern("session:*", 30*time.Second)

	// Set up event handler
	eventHandler := func(ctx context.Context, cache Cache, key string) error {
		// Log the invalidation event
		logger.Debug("Event-based invalidation triggered", zap.String("key", key))
		return nil
	}
	eventStrategy.RegisterEventHandler("invalidate", eventHandler)

	// Create composite strategy
	composite := NewCompositeInvalidationStrategy([]InvalidationStrategy{
		timeStrategy,
		patternStrategy,
		eventStrategy,
	}, logger)

	// Test different key types
	testKeys := []string{
		"user:123",
		"session:abc",
		"other:key",
	}

	for _, key := range testKeys {
		// Set key
		err := cache.Set(ctx, key, []byte("value"), 1*time.Hour)
		require.NoError(t, err)

		// Apply composite invalidation
		err = composite.Invalidate(ctx, cache, key)
		require.NoError(t, err)

		// Check TTL was set appropriately
		ttl, err := cache.GetTTL(ctx, key)
		require.NoError(t, err)
		assert.True(t, ttl > 0)

		// Verify TTL is appropriate for the key type
		if key[:4] == "user" {
			assert.True(t, ttl <= 1*time.Minute) // Pattern strategy
		} else if key[:7] == "session" {
			assert.True(t, ttl <= 30*time.Second) // Pattern strategy
		} else {
			assert.True(t, ttl <= 2*time.Minute) // Time strategy
		}
	}
}
