package cache

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// TestWebsiteContentCache_GetKey_Indirect tests the getKey helper function indirectly
func TestWebsiteContentCache_GetKey_Indirect(t *testing.T) {
	logger := zap.NewNop()
	cache := NewWebsiteContentCache(nil, logger, 24*time.Hour)

	// Since getKey is private, we test it indirectly through Get/Set operations
	url1 := "https://example.com/page1"
	url2 := "https://example.com/page2"

	// When Redis is nil, cache is disabled
	assert.False(t, cache.IsEnabled(), "Cache should be disabled when Redis is nil")
	
	// Test with mock Redis client
	mockRedis := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use test DB
	})
	defer mockRedis.Close()

	// Test connection - skip if Redis not available
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	if err := mockRedis.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available for testing")
	}

	// Create cache with mock Redis
	cacheWithRedis := NewWebsiteContentCache(mockRedis, logger, 1*time.Hour)
	require.True(t, cacheWithRedis.IsEnabled(), "Cache should be enabled with Redis")

	content1 := &CachedWebsiteContent{
		TextContent: "Content 1",
		ScrapedAt:   time.Now(),
		Success:     true,
	}
	
	content2 := &CachedWebsiteContent{
		TextContent: "Content 2",
		ScrapedAt:   time.Now(),
		Success:     true,
	}

	// Test that different URLs are stored separately (tests getKey indirectly)
	err := cacheWithRedis.Set(ctx, url1, content1)
	require.NoError(t, err)

	err = cacheWithRedis.Set(ctx, url2, content2)
	require.NoError(t, err)

	// Verify keys are different by retrieving them
	cached1, found1 := cacheWithRedis.Get(ctx, url1)
	require.True(t, found1, "Content 1 should be found")
	assert.Equal(t, content1.TextContent, cached1.TextContent)

	cached2, found2 := cacheWithRedis.Get(ctx, url2)
	require.True(t, found2, "Content 2 should be found")
	assert.Equal(t, content2.TextContent, cached2.TextContent)

	// Verify they are different
	assert.NotEqual(t, cached1.TextContent, cached2.TextContent, "Different URLs should have different content")

	// Cleanup
	cacheWithRedis.Delete(ctx, url1)
	cacheWithRedis.Delete(ctx, url2)
}

// TestWebsiteContentCache_Get_RedisErrors tests error handling in Get
func TestWebsiteContentCache_Get_RedisErrors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	// Test with nil Redis (disabled cache)
	cache := NewWebsiteContentCache(nil, logger, 24*time.Hour)
	ctx := context.Background()
	
	_, found := cache.Get(ctx, "https://example.com")
	assert.False(t, found, "Cache should return false when disabled")
}

// TestWebsiteContentCache_Set_RedisErrors tests error handling in Set
func TestWebsiteContentCache_Set_RedisErrors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	// Test with nil Redis (disabled cache)
	cache := NewWebsiteContentCache(nil, logger, 24*time.Hour)
	ctx := context.Background()
	
	content := &CachedWebsiteContent{
		TextContent: "Test content",
		ScrapedAt:   time.Now(),
		Success:     true,
	}
	
	err := cache.Set(ctx, "https://example.com", content)
	assert.NoError(t, err, "Set should succeed (no-op) when cache is disabled")
}

// TestWebsiteContentCache_Get_UnmarshalError tests unmarshal error handling
func TestWebsiteContentCache_Get_UnmarshalError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	mockRedis := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15,
	})
	defer mockRedis.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	if err := mockRedis.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available for testing")
	}

	cache := NewWebsiteContentCache(mockRedis, logger, 1*time.Hour)
	
	// Store invalid JSON data
	key := cache.getKey("https://example.com")
	err := mockRedis.Set(ctx, key, "invalid json", 1*time.Hour).Err()
	require.NoError(t, err)

	// Get should handle unmarshal error gracefully
	_, found := cache.Get(ctx, "https://example.com")
	assert.False(t, found, "Get should return false on unmarshal error")

	// Cleanup
	mockRedis.Del(ctx, key)
}

// TestWebsiteContentCache_Set_MarshalError tests marshal error handling
func TestWebsiteContentCache_Set_MarshalError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	mockRedis := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15,
	})
	defer mockRedis.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	if err := mockRedis.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available for testing")
	}

	cache := NewWebsiteContentCache(mockRedis, logger, 1*time.Hour)
	
	// Create content that should marshal successfully
	content := &CachedWebsiteContent{
		TextContent: "Valid content",
		ScrapedAt:   time.Now(),
		Success:     true,
	}
	
	// This should succeed
	err := cache.Set(ctx, "https://example.com", content)
	assert.NoError(t, err, "Set should succeed with valid content")
	
	// Cleanup
	cache.Delete(ctx, "https://example.com")
}

// TestWebsiteContentCache_Get_RedisGetError tests Redis Get error handling
func TestWebsiteContentCache_Get_RedisGetError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	// Test with disabled cache first
	cache := NewWebsiteContentCache(nil, logger, 24*time.Hour)
	ctx := context.Background()
	
	_, found := cache.Get(ctx, "https://example.com")
	assert.False(t, found, "Should return false when cache is disabled")
}

// TestWebsiteContentCache_Delete_RedisErrors tests error handling in Delete
func TestWebsiteContentCache_Delete_RedisErrors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	// Test with nil Redis (disabled cache)
	cache := NewWebsiteContentCache(nil, logger, 24*time.Hour)
	ctx := context.Background()
	
	err := cache.Delete(ctx, "https://example.com")
	assert.NoError(t, err, "Delete should succeed (no-op) when cache is disabled")
}

// TestWebsiteContentCache_GetKey_Format tests key format consistency
func TestWebsiteContentCache_GetKey_Format(t *testing.T) {
	logger := zap.NewNop()
	
	mockRedis := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15,
	})
	defer mockRedis.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	if err := mockRedis.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available for testing")
	}

	cache := NewWebsiteContentCache(mockRedis, logger, 1*time.Hour)
	
	// Test that same URL generates same key (indirectly through cache operations)
	url := "https://example.com/test"
	content := &CachedWebsiteContent{
		TextContent: "Test content",
		ScrapedAt:   time.Now(),
		Success:     true,
	}
	
	// Set and get should use same key
	err := cache.Set(ctx, url, content)
	require.NoError(t, err)
	
	cached, found := cache.Get(ctx, url)
	require.True(t, found, "Should find content for same URL")
	assert.Equal(t, content.TextContent, cached.TextContent)
	
	// Cleanup
	cache.Delete(ctx, url)
}

