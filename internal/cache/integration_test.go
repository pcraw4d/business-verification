package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestCacheIntegration tests the integration of all cache components
func TestCacheIntegration(t *testing.T) {
	logger := zap.NewNop()

	// Test 1: Memory Cache Factory
	t.Run("MemoryCacheFactory", func(t *testing.T) {
		factory := NewCacheFactory(logger)

		config := &CacheConfig{
			Type:            MemoryCache,
			DefaultTTL:      1 * time.Hour,
			MaxSize:         100,
			KeyPrefix:       "test",
			CleanupInterval: 5 * time.Minute,
		}

		cache, err := factory.CreateCache(config)
		require.NoError(t, err)
		defer cache.Close()

		// Test basic operations
		ctx := context.Background()
		key := "test:key"
		value := []byte("test value")

		err = cache.Set(ctx, key, value, 5*time.Minute)
		require.NoError(t, err)

		retrieved, err := cache.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, retrieved)
	})

	// Test 2: Merchant Cache Service with Memory Cache
	t.Run("MerchantCacheService", func(t *testing.T) {
		// Create memory cache
		config := &CacheConfig{
			Type:            MemoryCache,
			DefaultTTL:      1 * time.Hour,
			MaxSize:         100,
			KeyPrefix:       "test",
			CleanupInterval: 5 * time.Minute,
		}

		cache := NewMemoryCache(config)

		// Create merchant cache service
		serviceConfig := &MerchantCacheConfig{
			DefaultTTL:       15 * time.Minute,
			SearchTTL:        5 * time.Minute,
			DetailTTL:        30 * time.Minute,
			ListTTL:          10 * time.Minute,
			StatsTTL:         1 * time.Hour,
			PortfolioTTL:     20 * time.Minute,
			KeyPrefix:        "test:merchants",
			EnableMonitoring: false, // Disable for this test
		}

		service := NewMerchantCacheService(cache, serviceConfig, logger)
		defer service.Close() // This will also close the cache

		ctx := context.Background()

		// Test merchant detail caching
		merchantID := "merchant-123"
		merchant := map[string]interface{}{
			"id":   merchantID,
			"name": "Test Merchant",
			"type": "retail",
		}

		// Cache merchant detail
		err := service.CacheMerchantDetail(ctx, merchantID, merchant, 10*time.Minute)
		require.NoError(t, err)

		// Retrieve merchant detail
		var retrievedMerchant map[string]interface{}
		found, err := service.GetMerchantDetail(ctx, merchantID, &retrievedMerchant)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, merchantID, retrievedMerchant["id"])
		assert.Equal(t, "Test Merchant", retrievedMerchant["name"])

		// Test cache invalidation
		err = service.InvalidateMerchant(ctx, merchantID)
		require.NoError(t, err)

		// Verify invalidation
		found, err = service.GetMerchantDetail(ctx, merchantID, &retrievedMerchant)
		require.NoError(t, err)
		assert.False(t, found)
	})

	// Test 3: Cache Invalidation Manager
	t.Run("CacheInvalidationManager", func(t *testing.T) {
		// Create memory cache
		config := &CacheConfig{
			Type:            MemoryCache,
			DefaultTTL:      1 * time.Hour,
			MaxSize:         100,
			KeyPrefix:       "test",
			CleanupInterval: 5 * time.Minute,
		}

		cache := NewMemoryCache(config)
		defer cache.Close()

		// Create invalidation manager
		invalidationManager := NewCacheInvalidationManager(*config, logger)
		invalidationManager.SetCaches(cache)

		ctx := context.Background()

		// Add some test data
		testKeys := []string{"test:pattern:1", "test:pattern:2", "test:other:1"}
		for _, key := range testKeys {
			err := cache.Set(ctx, key, []byte("value"), 5*time.Minute)
			require.NoError(t, err)
		}

		// Test pattern invalidation - use a simpler pattern
		err := invalidationManager.InvalidateByPattern(ctx, "test:pattern:1")
		require.NoError(t, err)

		// Verify specific key invalidation
		exists, err := cache.Exists(ctx, "test:pattern:1")
		require.NoError(t, err)
		assert.False(t, exists)

		// Verify other keys still exist
		exists, err = cache.Exists(ctx, "test:pattern:2")
		require.NoError(t, err)
		assert.True(t, exists) // Should still exist

		exists, err = cache.Exists(ctx, "test:other:1")
		require.NoError(t, err)
		assert.True(t, exists) // Should still exist
	})

	// Test 4: Cache Monitoring Service
	t.Run("CacheMonitoringService", func(t *testing.T) {
		// Create memory cache
		config := &CacheConfig{
			Type:            MemoryCache,
			DefaultTTL:      1 * time.Hour,
			MaxSize:         100,
			KeyPrefix:       "test",
			CleanupInterval: 5 * time.Minute,
		}

		cache := NewMemoryCache(config)
		defer cache.Close()

		// Create monitoring service
		monitorConfig := &MonitoringConfig{
			CollectionInterval: 100 * time.Millisecond, // Fast collection for test
			EnableAlerts:       true,
		}

		monitor := NewCacheMonitoringService([]Cache{cache}, monitorConfig, logger)

		// Start monitoring
		err := monitor.Start()
		require.NoError(t, err)
		defer monitor.Stop()

		ctx := context.Background()

		// Add some data to generate metrics
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("test:monitor:%d", i)
			value := []byte(fmt.Sprintf("value-%d", i))
			err := cache.Set(ctx, key, value, 5*time.Minute)
			require.NoError(t, err)
		}

		// Wait for metrics collection
		time.Sleep(200 * time.Millisecond)

		// Get metrics
		metrics := monitor.GetMetrics()
		assert.NotNil(t, metrics)
		assert.True(t, metrics.TotalSize >= 10)

		// Test health status
		health := monitor.GetHealthStatus(ctx)
		assert.NotEmpty(t, health)

		// Test alerts
		alerts := monitor.CheckAlerts()
		// Alerts may or may not be triggered depending on thresholds
		assert.NotNil(t, alerts)
	})
}

// TestCacheFactory tests the cache factory functionality
func TestCacheFactory(t *testing.T) {
	logger := zap.NewNop()
	factory := NewCacheFactory(logger)

	// Test default cache creation
	t.Run("DefaultCache", func(t *testing.T) {
		cache, err := factory.CreateDefaultCache()
		require.NoError(t, err)
		defer cache.Close()

		// Test basic operation
		ctx := context.Background()
		err = cache.Set(ctx, "test", []byte("value"), 5*time.Minute)
		require.NoError(t, err)

		value, err := cache.Get(ctx, "test")
		require.NoError(t, err)
		assert.Equal(t, []byte("value"), value)
	})

	// Test memory cache creation
	t.Run("MemoryCache", func(t *testing.T) {
		config := &CacheConfig{
			Type:       MemoryCache,
			DefaultTTL: 1 * time.Hour,
			MaxSize:    100,
		}

		cache, err := factory.CreateMemoryCacheWithConfig(config)
		require.NoError(t, err)
		defer cache.Close()

		// Test basic operation
		ctx := context.Background()
		err = cache.Set(ctx, "test", []byte("value"), 5*time.Minute)
		require.NoError(t, err)

		value, err := cache.Get(ctx, "test")
		require.NoError(t, err)
		assert.Equal(t, []byte("value"), value)
	})
}
