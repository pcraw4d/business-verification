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

// TestRedisCacheBasicOperations tests basic Redis cache operations
func TestRedisCacheBasicOperations(t *testing.T) {
	// Skip if Redis is not available
	if !isRedisAvailable() {
		t.Skip("Redis not available, skipping Redis cache tests")
	}

	logger := zap.NewNop()
	config := &RedisCacheConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1, // Use test database
		TTL:      1 * time.Hour,
		PoolSize: 10,
	}

	cache, err := NewRedisCache(config, logger)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// Test Set and Get
	key := "test:key:1"
	value := []byte("test value")
	ttl := 5 * time.Minute

	err = cache.Set(ctx, key, value, ttl)
	require.NoError(t, err)

	retrieved, err := cache.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, value, retrieved)

	// Test Exists
	exists, err := cache.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test GetTTL
	retrievedTTL, err := cache.GetTTL(ctx, key)
	require.NoError(t, err)
	assert.True(t, retrievedTTL > 0)
	assert.True(t, retrievedTTL <= ttl)

	// Test Delete
	err = cache.Delete(ctx, key)
	require.NoError(t, err)

	// Verify deletion
	_, err = cache.Get(ctx, key)
	assert.Equal(t, CacheNotFoundError, err)

	exists, err = cache.Exists(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)
}

// TestRedisCacheTTL tests TTL functionality
func TestRedisCacheTTL(t *testing.T) {
	if !isRedisAvailable() {
		t.Skip("Redis not available, skipping Redis cache tests")
	}

	logger := zap.NewNop()
	config := &RedisCacheConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
		TTL:      1 * time.Hour,
		PoolSize: 10,
	}

	cache, err := NewRedisCache(config, logger)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// Test with TTL
	key := "test:ttl:1"
	value := []byte("test value")
	ttl := 100 * time.Millisecond

	err = cache.Set(ctx, key, value, ttl)
	require.NoError(t, err)

	// Should exist immediately
	exists, err := cache.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should not exist after expiration
	exists, err = cache.Exists(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)

	_, err = cache.Get(ctx, key)
	assert.Equal(t, CacheNotFoundError, err)
}

// TestRedisCacheBulkOperations tests bulk operations
func TestRedisCacheBulkOperations(t *testing.T) {
	if !isRedisAvailable() {
		t.Skip("Redis not available, skipping Redis cache tests")
	}

	logger := zap.NewNop()
	config := &RedisCacheConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
		TTL:      1 * time.Hour,
		PoolSize: 10,
	}

	cache, err := NewRedisCache(config, logger)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// Test SetEntries
	entries := map[string]*CacheEntry{
		"test:bulk:1": {
			Key:       "test:bulk:1",
			Value:     "value1",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(5 * time.Minute),
			Size:      6,
		},
		"test:bulk:2": {
			Key:       "test:bulk:2",
			Value:     "value2",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(5 * time.Minute),
			Size:      6,
		},
	}

	err = cache.SetEntries(ctx, entries)
	require.NoError(t, err)

	// Test GetEntries
	keys := []string{"test:bulk:1", "test:bulk:2"}
	retrieved, err := cache.GetEntries(ctx, keys)
	require.NoError(t, err)
	assert.Len(t, retrieved, 2)

	// Test DeleteEntries
	err = cache.DeleteEntries(ctx, keys)
	require.NoError(t, err)

	// Verify deletion
	retrieved, err = cache.GetEntries(ctx, keys)
	require.NoError(t, err)
	assert.Len(t, retrieved, 0)
}

// TestRedisCacheKeys tests key operations
func TestRedisCacheKeys(t *testing.T) {
	if !isRedisAvailable() {
		t.Skip("Redis not available, skipping Redis cache tests")
	}

	logger := zap.NewNop()
	config := &RedisCacheConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
		TTL:      1 * time.Hour,
		PoolSize: 10,
	}

	cache, err := NewRedisCache(config, logger)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// Set some test keys
	testKeys := []string{
		"test:keys:1",
		"test:keys:2",
		"test:other:1",
	}

	for _, key := range testKeys {
		err = cache.Set(ctx, key, []byte("value"), 5*time.Minute)
		require.NoError(t, err)
	}

	// Test GetKeys with pattern
	keys, err := cache.GetKeys(ctx, "test:keys:*")
	require.NoError(t, err)
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "test:keys:1")
	assert.Contains(t, keys, "test:keys:2")

	// Test GetKeys with all pattern
	allKeys, err := cache.GetKeys(ctx, "test:*")
	require.NoError(t, err)
	assert.Len(t, allKeys, 3)

	// Clean up
	for _, key := range testKeys {
		cache.Delete(ctx, key)
	}
}

// TestRedisCacheStats tests cache statistics
func TestRedisCacheStats(t *testing.T) {
	if !isRedisAvailable() {
		t.Skip("Redis not available, skipping Redis cache tests")
	}

	logger := zap.NewNop()
	config := &RedisCacheConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
		TTL:      1 * time.Hour,
		PoolSize: 10,
	}

	cache, err := NewRedisCache(config, logger)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// Get initial stats
	stats, err := cache.GetStats(ctx)
	require.NoError(t, err)
	initialSize := stats.Size

	// Add some data
	key := "test:stats:1"
	value := []byte("test value")
	err = cache.Set(ctx, key, value, 5*time.Minute)
	require.NoError(t, err)

	// Get updated stats
	stats, err = cache.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, initialSize+1, stats.Size)

	// Test hit
	_, err = cache.Get(ctx, key)
	require.NoError(t, err)

	// Test miss
	_, err = cache.Get(ctx, "nonexistent:key")
	assert.Equal(t, CacheNotFoundError, err)

	// Get final stats
	stats, err = cache.GetStats(ctx)
	require.NoError(t, err)
	assert.True(t, stats.HitCount > 0)
	assert.True(t, stats.MissCount > 0)

	// Clean up
	cache.Delete(ctx, key)
}

// TestRedisCacheClear tests cache clearing
func TestRedisCacheClear(t *testing.T) {
	if !isRedisAvailable() {
		t.Skip("Redis not available, skipping Redis cache tests")
	}

	logger := zap.NewNop()
	config := &RedisCacheConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
		TTL:      1 * time.Hour,
		PoolSize: 10,
	}

	cache, err := NewRedisCache(config, logger)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// Add some data
	keys := []string{"test:clear:1", "test:clear:2", "test:clear:3"}
	for _, key := range keys {
		err = cache.Set(ctx, key, []byte("value"), 5*time.Minute)
		require.NoError(t, err)
	}

	// Verify data exists
	for _, key := range keys {
		exists, err := cache.Exists(ctx, key)
		require.NoError(t, err)
		assert.True(t, exists)
	}

	// Clear cache
	err = cache.Clear(ctx)
	require.NoError(t, err)

	// Verify data is gone
	for _, key := range keys {
		exists, err := cache.Exists(ctx, key)
		require.NoError(t, err)
		assert.False(t, exists)
	}
}

// TestRedisCacheHealthCheck tests health check functionality
func TestRedisCacheHealthCheck(t *testing.T) {
	if !isRedisAvailable() {
		t.Skip("Redis not available, skipping Redis cache tests")
	}

	logger := zap.NewNop()
	config := &RedisCacheConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
		TTL:      1 * time.Hour,
		PoolSize: 10,
	}

	cache, err := NewRedisCache(config, logger)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// Test health check
	err = cache.HealthCheck(ctx)
	require.NoError(t, err)
}

// TestRedisCacheSetTTL tests SetTTL functionality
func TestRedisCacheSetTTL(t *testing.T) {
	if !isRedisAvailable() {
		t.Skip("Redis not available, skipping Redis cache tests")
	}

	logger := zap.NewNop()
	config := &RedisCacheConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
		TTL:      1 * time.Hour,
		PoolSize: 10,
	}

	cache, err := NewRedisCache(config, logger)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// Set a key with long TTL
	key := "test:setttl:1"
	value := []byte("test value")
	err = cache.Set(ctx, key, value, 1*time.Hour)
	require.NoError(t, err)

	// Set new TTL
	newTTL := 100 * time.Millisecond
	err = cache.SetTTL(ctx, key, newTTL)
	require.NoError(t, err)

	// Verify new TTL
	retrievedTTL, err := cache.GetTTL(ctx, key)
	require.NoError(t, err)
	assert.True(t, retrievedTTL > 0)
	assert.True(t, retrievedTTL <= newTTL)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Verify expiration
	exists, err := cache.Exists(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)
}

// TestRedisCacheErrorHandling tests error handling
func TestRedisCacheErrorHandling(t *testing.T) {
	if !isRedisAvailable() {
		t.Skip("Redis not available, skipping Redis cache tests")
	}

	logger := zap.NewNop()
	config := &RedisCacheConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
		TTL:      1 * time.Hour,
		PoolSize: 10,
	}

	cache, err := NewRedisCache(config, logger)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// Test GetTTL on non-existent key
	_, err = cache.GetTTL(ctx, "nonexistent:key")
	assert.Equal(t, CacheNotFoundError, err)

	// Test SetTTL on non-existent key
	err = cache.SetTTL(ctx, "nonexistent:key", 5*time.Minute)
	assert.Equal(t, CacheNotFoundError, err)

	// Test Delete on non-existent key
	err = cache.Delete(ctx, "nonexistent:key")
	assert.Equal(t, CacheNotFoundError, err)
}

// TestRedisCacheConcurrency tests concurrent operations
func TestRedisCacheConcurrency(t *testing.T) {
	if !isRedisAvailable() {
		t.Skip("Redis not available, skipping Redis cache tests")
	}

	logger := zap.NewNop()
	config := &RedisCacheConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
		TTL:      1 * time.Hour,
		PoolSize: 10,
	}

	cache, err := NewRedisCache(config, logger)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// Test concurrent writes
	numGoroutines := 10
	numOperations := 100

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("test:concurrent:%d:%d", goroutineID, j)
				value := []byte(fmt.Sprintf("value:%d:%d", goroutineID, j))

				err := cache.Set(ctx, key, value, 5*time.Minute)
				if err != nil {
					t.Errorf("Set failed: %v", err)
				}

				retrieved, err := cache.Get(ctx, key)
				if err != nil {
					t.Errorf("Get failed: %v", err)
				}

				if string(retrieved) != string(value) {
					t.Errorf("Value mismatch: expected %s, got %s", string(value), string(retrieved))
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// Helper function to check if Redis is available
func isRedisAvailable() bool {
	config := &RedisCacheConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
		TTL:      1 * time.Hour,
		PoolSize: 10,
	}

	cache, err := NewRedisCache(config, zap.NewNop())
	if err != nil {
		return false
	}
	defer cache.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = cache.HealthCheck(ctx)
	return err == nil
}
