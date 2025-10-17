//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/cache"
	"kyb-platform/services/risk-assessment-service/internal/config"
	"kyb-platform/services/risk-assessment-service/internal/models"
)

// Note: TestCache is defined in test_helpers.go

// SetupTestCache creates a test cache connection
func SetupTestCache(t *testing.T) *TestCache {
	logger := zap.NewNop()

	// Load test configuration
	cfg, err := config.Load()
	require.NoError(t, err)

	// Override with test values
	cfg.Redis.URL = "redis://localhost:6379"
	cfg.Redis.DB = 1 // Use different DB for tests
	cfg.Redis.KeyPrefix = "test:"

	// Create cache
	cacheInstance, err := cache.NewRedisCache(&cache.CacheConfig{
		Addrs:             []string{"localhost:6379"},
		Password:          "",
		DB:                1,
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
	}, logger)

	if err != nil {
		t.Skipf("Skipping cache integration test: Redis not available: %v", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cacheInstance.Health(ctx); err != nil {
		t.Skipf("Skipping cache integration test: Redis not available: %v", err)
	}

	return &TestCache{
		cache:  cacheInstance,
		logger: logger,
	}
}

// TeardownTestCache cleans up test cache
func (tc *TestCache) TeardownTestCache() {
	if tc.cache != nil {
		tc.cache.Close()
	}
}

// CleanupTestData removes test data from cache
func (tc *TestCache) CleanupTestData(t *testing.T) {
	ctx := context.Background()
	err := tc.cache.Clear(ctx)
	if err != nil {
		t.Logf("Failed to clear test cache: %v", err)
	}
}

func TestCache_BasicOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test cache
	cache := SetupTestCache(t)
	defer cache.TeardownTestCache()
	defer cache.CleanupTestData(t)

	ctx := context.Background()

	tests := []struct {
		name        string
		key         string
		value       interface{}
		ttl         time.Duration
		expectError bool
	}{
		{
			name:        "set string value",
			key:         "test-string",
			value:       "test-value",
			ttl:         5 * time.Minute,
			expectError: false,
		},
		{
			name:        "set integer value",
			key:         "test-int",
			value:       42,
			ttl:         5 * time.Minute,
			expectError: false,
		},
		{
			name:        "set float value",
			key:         "test-float",
			value:       3.14,
			ttl:         5 * time.Minute,
			expectError: false,
		},
		{
			name:        "set boolean value",
			key:         "test-bool",
			value:       true,
			ttl:         5 * time.Minute,
			expectError: false,
		},
		{
			name:        "set map value",
			key:         "test-map",
			value:       map[string]interface{}{"key": "value", "number": 123},
			ttl:         5 * time.Minute,
			expectError: false,
		},
		{
			name:        "set slice value",
			key:         "test-slice",
			value:       []string{"item1", "item2", "item3"},
			ttl:         5 * time.Minute,
			expectError: false,
		},
		{
			name:        "set struct value",
			key:         "test-struct",
			value:       models.RiskAssessment{ID: "test-id", BusinessName: "Test Company"},
			ttl:         5 * time.Minute,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Set
			err := cache.cache.Set(ctx, tt.key, tt.value, tt.ttl)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Test Get
			var retrieved interface{}
			err = cache.cache.Get(ctx, tt.key, &retrieved)
			assert.NoError(t, err)
			assert.Equal(t, tt.value, retrieved)

			// Test Exists
			exists, err := cache.cache.Exists(ctx, tt.key)
			assert.NoError(t, err)
			assert.True(t, exists)

			// Test Delete
			err = cache.cache.Delete(ctx, tt.key)
			assert.NoError(t, err)

			// Verify deletion
			exists, err = cache.cache.Exists(ctx, tt.key)
			assert.NoError(t, err)
			assert.False(t, exists)
		})
	}
}

func TestCache_GetOrSet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test cache
	cache := SetupTestCache(t)
	defer cache.TeardownTestCache()
	defer cache.CleanupTestData(t)

	ctx := context.Background()

	tests := []struct {
		name           string
		key            string
		setter         func() (interface{}, error)
		expectedValue  interface{}
		expectError    bool
		expectCacheHit bool
	}{
		{
			name: "cache miss - setter success",
			key:  "get-or-set-miss",
			setter: func() (interface{}, error) {
				return "setter-value", nil
			},
			expectedValue:  "setter-value",
			expectError:    false,
			expectCacheHit: false,
		},
		{
			name: "cache hit",
			key:  "get-or-set-hit",
			setter: func() (interface{}, error) {
				return "setter-value", nil
			},
			expectedValue:  "cached-value",
			expectError:    false,
			expectCacheHit: true,
		},
		{
			name: "setter error",
			key:  "get-or-set-error",
			setter: func() (interface{}, error) {
				return nil, assert.AnError
			},
			expectedValue:  nil,
			expectError:    true,
			expectCacheHit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup cache for hit test
			if tt.expectCacheHit {
				err := cache.cache.Set(ctx, tt.key, tt.expectedValue, 5*time.Minute)
				require.NoError(t, err)
			}

			// Test GetOrSet
			var result interface{}
			err := cache.cache.GetOrSet(ctx, tt.key, &result, 5*time.Minute, tt.setter)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, result)
			}
		})
	}
}

func TestCache_MultiOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test cache
	cache := SetupTestCache(t)
	defer cache.TeardownTestCache()
	defer cache.CleanupTestData(t)

	ctx := context.Background()

	// Test MSet
	values := map[string]interface{}{
		"multi-key-1": "value1",
		"multi-key-2": "value2",
		"multi-key-3": "value3",
	}

	err := cache.cache.MSet(ctx, values, 5*time.Minute)
	assert.NoError(t, err)

	// Test MGet
	keys := []string{"multi-key-1", "multi-key-2", "multi-key-3"}
	results, err := cache.cache.MGet(ctx, keys)
	assert.NoError(t, err)

	assert.Len(t, results, 3)
	assert.Equal(t, "value1", results["multi-key-1"])
	assert.Equal(t, "value2", results["multi-key-2"])
	assert.Equal(t, "value3", results["multi-key-3"])

	// Test MGet with non-existent keys
	nonExistentKeys := []string{"non-existent-1", "non-existent-2"}
	nonExistentResults, err := cache.cache.MGet(ctx, nonExistentKeys)
	assert.NoError(t, err)

	assert.Len(t, nonExistentResults, 0)

	// Test MGet with mixed keys
	mixedKeys := []string{"multi-key-1", "non-existent-1", "multi-key-2"}
	mixedResults, err := cache.cache.MGet(ctx, mixedKeys)
	assert.NoError(t, err)

	assert.Len(t, mixedResults, 2)
	assert.Equal(t, "value1", mixedResults["multi-key-1"])
	assert.Equal(t, "value2", mixedResults["multi-key-2"])
}

func TestCache_TTL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test cache
	cache := SetupTestCache(t)
	defer cache.TeardownTestCache()
	defer cache.CleanupTestData(t)

	ctx := context.Background()

	// Test with short TTL
	shortTTL := 1 * time.Second
	err := cache.cache.Set(ctx, "short-ttl-key", "short-ttl-value", shortTTL)
	assert.NoError(t, err)

	// Verify value exists
	exists, err := cache.cache.Exists(ctx, "short-ttl-key")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Wait for TTL to expire
	time.Sleep(2 * time.Second)

	// Verify value no longer exists
	exists, err = cache.cache.Exists(ctx, "short-ttl-key")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Test with long TTL
	longTTL := 1 * time.Hour
	err = cache.cache.Set(ctx, "long-ttl-key", "long-ttl-value", longTTL)
	assert.NoError(t, err)

	// Verify value exists
	exists, err = cache.cache.Exists(ctx, "long-ttl-key")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test with default TTL
	err = cache.cache.Set(ctx, "default-ttl-key", "default-ttl-value", 0)
	assert.NoError(t, err)

	// Verify value exists
	exists, err = cache.cache.Exists(ctx, "default-ttl-key")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestCache_Clear(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test cache
	cache := SetupTestCache(t)
	defer cache.TeardownTestCache()
	defer cache.CleanupTestData(t)

	ctx := context.Background()

	// Set multiple keys
	keys := []string{"clear-key-1", "clear-key-2", "clear-key-3"}
	for _, key := range keys {
		err := cache.cache.Set(ctx, key, "value", 5*time.Minute)
		require.NoError(t, err)
	}

	// Verify all keys exist
	for _, key := range keys {
		exists, err := cache.cache.Exists(ctx, key)
		assert.NoError(t, err)
		assert.True(t, exists)
	}

	// Clear cache
	err := cache.cache.Clear(ctx)
	assert.NoError(t, err)

	// Verify all keys are gone
	for _, key := range keys {
		exists, err := cache.cache.Exists(ctx, key)
		assert.NoError(t, err)
		assert.False(t, exists)
	}
}

func TestCache_Metrics(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test cache
	cache := SetupTestCache(t)
	defer cache.TeardownTestCache()
	defer cache.CleanupTestData(t)

	ctx := context.Background()

	// Reset metrics
	cache.cache.ResetMetrics()

	// Perform operations
	err := cache.cache.Set(ctx, "metrics-key-1", "value1", 5*time.Minute)
	assert.NoError(t, err)

	err = cache.cache.Set(ctx, "metrics-key-2", "value2", 5*time.Minute)
	assert.NoError(t, err)

	var result interface{}
	err = cache.cache.Get(ctx, "metrics-key-1", &result)
	assert.NoError(t, err)

	err = cache.cache.Get(ctx, "non-existent-key", &result)
	assert.Error(t, err) // Should be cache miss

	err = cache.cache.Delete(ctx, "metrics-key-1")
	assert.NoError(t, err)

	// Get metrics
	metrics := cache.cache.GetMetrics()

	assert.Equal(t, int64(2), metrics.Sets)
	assert.Equal(t, int64(1), metrics.Hits)
	assert.Equal(t, int64(1), metrics.Misses)
	assert.Equal(t, int64(1), metrics.Deletes)
	assert.Equal(t, int64(2), metrics.TotalRequests)
	assert.Equal(t, 0.5, metrics.HitRate) // 1 hit out of 2 total requests
	assert.NotZero(t, metrics.LastUpdated)
}

func TestCache_Concurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test cache
	cache := SetupTestCache(t)
	defer cache.TeardownTestCache()
	defer cache.CleanupTestData(t)

	ctx := context.Background()

	// Test concurrent sets
	numGoroutines := 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			key := fmt.Sprintf("concurrent-key-%d", id)
			value := fmt.Sprintf("concurrent-value-%d", id)

			err := cache.cache.Set(ctx, key, value, 5*time.Minute)
			results <- err
		}(i)
	}

	// Wait for all sets to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err, "Concurrent set %d failed", i)
	}

	// Test concurrent gets
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			key := fmt.Sprintf("concurrent-key-%d", id)
			var value interface{}

			err := cache.cache.Get(ctx, key, &value)
			results <- err
		}(i)
	}

	// Wait for all gets to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err, "Concurrent get %d failed", i)
	}

	// Test concurrent deletes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			key := fmt.Sprintf("concurrent-key-%d", id)

			err := cache.cache.Delete(ctx, key)
			results <- err
		}(i)
	}

	// Wait for all deletes to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err, "Concurrent delete %d failed", i)
	}
}

func TestCache_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test cache
	cache := SetupTestCache(t)
	defer cache.TeardownTestCache()
	defer cache.CleanupTestData(t)

	ctx := context.Background()

	tests := []struct {
		name         string
		testFunction func() error
		expectError  bool
	}{
		{
			name: "get non-existent key",
			testFunction: func() error {
				var result interface{}
				return cache.cache.Get(ctx, "non-existent-key", &result)
			},
			expectError: true,
		},
		{
			name: "delete non-existent key",
			testFunction: func() error {
				return cache.cache.Delete(ctx, "non-existent-key")
			},
			expectError: false, // Delete should not error for non-existent keys
		},
		{
			name: "set with invalid value",
			testFunction: func() error {
				// Create a value that can't be marshaled
				invalidValue := make(chan int)
				return cache.cache.Set(ctx, "invalid-key", invalidValue, 5*time.Minute)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testFunction()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCache_Health(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test cache
	cache := SetupTestCache(t)
	defer cache.TeardownTestCache()
	defer cache.CleanupTestData(t)

	ctx := context.Background()

	// Test health check
	err := cache.cache.Health(ctx)
	assert.NoError(t, err)

	// Test health check with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	err = cache.cache.Health(timeoutCtx)
	assert.NoError(t, err)
}

// Benchmark tests
func BenchmarkCache_Set(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping integration benchmark")
	}

	// Setup test cache
	cache := SetupTestCache(&testing.T{})
	defer cache.TeardownTestCache()
	defer cache.CleanupTestData(&testing.T{})

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark-set-%d", i)
		value := fmt.Sprintf("benchmark-value-%d", i)

		err := cache.cache.Set(ctx, key, value, 5*time.Minute)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCache_Get(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping integration benchmark")
	}

	// Setup test cache
	cache := SetupTestCache(&testing.T{})
	defer cache.TeardownTestCache()
	defer cache.CleanupTestData(&testing.T{})

	ctx := context.Background()

	// Pre-populate cache
	key := "benchmark-get-key"
	value := "benchmark-get-value"
	err := cache.cache.Set(ctx, key, value, 5*time.Minute)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result interface{}
		err := cache.cache.Get(ctx, key, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCache_Concurrent(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping integration benchmark")
	}

	// Setup test cache
	cache := SetupTestCache(&testing.T{})
	defer cache.TeardownTestCache()
	defer cache.CleanupTestData(&testing.T{})

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("benchmark-concurrent-%d", i)
			value := fmt.Sprintf("benchmark-concurrent-value-%d", i)

			err := cache.cache.Set(ctx, key, value, 5*time.Minute)
			if err != nil {
				b.Fatal(err)
			}

			var result interface{}
			err = cache.cache.Get(ctx, key, &result)
			if err != nil {
				b.Fatal(err)
			}

			i++
		}
	})
}
