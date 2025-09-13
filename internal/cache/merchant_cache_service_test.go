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

// TestMerchantCacheServiceBasicOperations tests basic merchant cache operations
func TestMerchantCacheServiceBasicOperations(t *testing.T) {
	// Create a memory cache for testing
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         1000,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	cache := NewMemoryCache(config)

	logger := zap.NewNop()
	service := NewMerchantCacheService(cache, nil, logger)
	defer service.Close() // This will also close the underlying cache

	ctx := context.Background()

	// Test merchant list caching
	merchants := []map[string]interface{}{
		{"id": "1", "name": "Merchant 1"},
		{"id": "2", "name": "Merchant 2"},
	}

	filters := map[string]interface{}{
		"status": "active",
		"type":   "retail",
	}

	// Cache merchant list
	err := service.CacheMerchantList(ctx, filters, merchants, 5*time.Minute)
	require.NoError(t, err)

	// Retrieve merchant list
	var retrievedMerchants []map[string]interface{}
	found, err := service.GetMerchantList(ctx, filters, &retrievedMerchants)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Len(t, retrievedMerchants, 2)
	assert.Equal(t, "Merchant 1", retrievedMerchants[0]["name"])

	// Test cache miss with different filters
	var emptyMerchants []map[string]interface{}
	differentFilters := map[string]interface{}{
		"status": "inactive",
	}
	found, err = service.GetMerchantList(ctx, differentFilters, &emptyMerchants)
	require.NoError(t, err)
	assert.False(t, found)
}

// TestMerchantCacheServiceDetailOperations tests merchant detail caching
func TestMerchantCacheServiceDetailOperations(t *testing.T) {
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         1000,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	cache := NewMemoryCache(config)

	logger := zap.NewNop()
	service := NewMerchantCacheService(cache, nil, logger)
	defer service.Close() // This will also close the underlying cache

	ctx := context.Background()

	// Test merchant detail caching
	merchantID := "merchant-123"
	merchant := map[string]interface{}{
		"id":         merchantID,
		"name":       "Test Merchant",
		"email":      "test@merchant.com",
		"status":     "active",
		"created_at": time.Now(),
		"risk_level": "medium",
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

	// Test cache miss with different merchant ID
	var emptyMerchant map[string]interface{}
	found, err = service.GetMerchantDetail(ctx, "nonexistent-merchant", &emptyMerchant)
	require.NoError(t, err)
	assert.False(t, found)
}

// TestMerchantCacheServiceSearchOperations tests search caching
func TestMerchantCacheServiceSearchOperations(t *testing.T) {
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         1000,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	cache := NewMemoryCache(config)

	logger := zap.NewNop()
	service := NewMerchantCacheService(cache, nil, logger)
	defer service.Close() // This will also close the underlying cache

	ctx := context.Background()

	// Test search caching
	query := "test merchant"
	filters := map[string]interface{}{
		"category": "retail",
		"status":   "active",
	}

	searchResults := []map[string]interface{}{
		{"id": "1", "name": "Test Merchant 1", "score": 0.95},
		{"id": "2", "name": "Test Merchant 2", "score": 0.87},
	}

	// Cache search results
	err := service.CacheMerchantSearch(ctx, query, filters, searchResults, 3*time.Minute)
	require.NoError(t, err)

	// Retrieve search results
	var retrievedResults []map[string]interface{}
	found, err := service.GetMerchantSearch(ctx, query, filters, &retrievedResults)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Len(t, retrievedResults, 2)
	assert.Equal(t, "Test Merchant 1", retrievedResults[0]["name"])

	// Test cache miss with different query
	var emptyResults []map[string]interface{}
	found, err = service.GetMerchantSearch(ctx, "different query", filters, &emptyResults)
	require.NoError(t, err)
	assert.False(t, found)
}

// TestMerchantCacheServiceStatsOperations tests stats caching
func TestMerchantCacheServiceStatsOperations(t *testing.T) {
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         1000,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	cache := NewMemoryCache(config)

	logger := zap.NewNop()
	service := NewMerchantCacheService(cache, nil, logger)
	defer service.Close() // This will also close the underlying cache

	ctx := context.Background()

	// Test stats caching
	stats := map[string]interface{}{
		"total_merchants":    1500,
		"active_merchants":   1200,
		"pending_merchants":  200,
		"inactive_merchants": 100,
		"risk_distribution": map[string]int{
			"low":    800,
			"medium": 500,
			"high":   200,
		},
	}

	// Cache stats
	err := service.CacheMerchantStats(ctx, stats, 30*time.Minute)
	require.NoError(t, err)

	// Retrieve stats
	var retrievedStats map[string]interface{}
	found, err := service.GetMerchantStats(ctx, &retrievedStats)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, float64(1500), retrievedStats["total_merchants"])
	assert.Equal(t, float64(1200), retrievedStats["active_merchants"])

	// Test cache miss (stats should be cached)
	found, err = service.GetMerchantStats(ctx, &retrievedStats)
	require.NoError(t, err)
	assert.True(t, found) // Should still be cached
}

// TestMerchantCacheServiceInvalidation tests cache invalidation
func TestMerchantCacheServiceInvalidation(t *testing.T) {
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         1000,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	cache := NewMemoryCache(config)

	logger := zap.NewNop()
	service := NewMerchantCacheService(cache, nil, logger)
	defer service.Close() // This will also close the underlying cache

	ctx := context.Background()

	merchantID := "merchant-456"
	merchant := map[string]interface{}{
		"id":   merchantID,
		"name": "Test Merchant",
	}

	// Cache merchant detail
	err := service.CacheMerchantDetail(ctx, merchantID, merchant, 10*time.Minute)
	require.NoError(t, err)

	// Cache some lists and searches
	merchants := []map[string]interface{}{merchant}
	err = service.CacheMerchantList(ctx, nil, merchants, 10*time.Minute)
	require.NoError(t, err)

	searchResults := []map[string]interface{}{merchant}
	err = service.CacheMerchantSearch(ctx, "test", nil, searchResults, 10*time.Minute)
	require.NoError(t, err)

	// Verify data is cached
	var retrievedMerchant map[string]interface{}
	found, err := service.GetMerchantDetail(ctx, merchantID, &retrievedMerchant)
	require.NoError(t, err)
	assert.True(t, found)

	// Invalidate merchant
	err = service.InvalidateMerchant(ctx, merchantID)
	require.NoError(t, err)

	// Verify merchant detail is invalidated
	found, err = service.GetMerchantDetail(ctx, merchantID, &retrievedMerchant)
	require.NoError(t, err)
	assert.False(t, found)

	// Verify lists and searches are also invalidated
	var retrievedList []map[string]interface{}
	found, err = service.GetMerchantList(ctx, nil, &retrievedList)
	require.NoError(t, err)
	assert.False(t, found)

	var retrievedSearch []map[string]interface{}
	found, err = service.GetMerchantSearch(ctx, "test", nil, &retrievedSearch)
	require.NoError(t, err)
	assert.False(t, found)
}

// TestMerchantCacheServiceInvalidateAll tests complete cache invalidation
func TestMerchantCacheServiceInvalidateAll(t *testing.T) {
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         1000,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	cache := NewMemoryCache(config)

	logger := zap.NewNop()
	service := NewMerchantCacheService(cache, nil, logger)
	defer service.Close() // This will also close the underlying cache

	ctx := context.Background()

	// Cache various data
	merchant := map[string]interface{}{"id": "1", "name": "Test"}
	err := service.CacheMerchantDetail(ctx, "1", merchant, 10*time.Minute)
	require.NoError(t, err)

	merchants := []map[string]interface{}{merchant}
	err = service.CacheMerchantList(ctx, nil, merchants, 10*time.Minute)
	require.NoError(t, err)

	stats := map[string]interface{}{"total": 100}
	err = service.CacheMerchantStats(ctx, stats, 10*time.Minute)
	require.NoError(t, err)

	// Verify data is cached
	var retrievedMerchant map[string]interface{}
	found, err := service.GetMerchantDetail(ctx, "1", &retrievedMerchant)
	require.NoError(t, err)
	assert.True(t, found)

	// Invalidate all
	err = service.InvalidateAll(ctx)
	require.NoError(t, err)

	// Verify all data is invalidated
	found, err = service.GetMerchantDetail(ctx, "1", &retrievedMerchant)
	require.NoError(t, err)
	assert.False(t, found)

	var retrievedList []map[string]interface{}
	found, err = service.GetMerchantList(ctx, nil, &retrievedList)
	require.NoError(t, err)
	assert.False(t, found)

	var retrievedStats map[string]interface{}
	found, err = service.GetMerchantStats(ctx, &retrievedStats)
	require.NoError(t, err)
	assert.False(t, found)
}

// TestMerchantCacheServiceKeyGeneration tests key generation
func TestMerchantCacheServiceKeyGeneration(t *testing.T) {
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         1000,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	cache := NewMemoryCache(config)

	logger := zap.NewNop()
	service := NewMerchantCacheService(cache, nil, logger)
	defer service.Close() // This will also close the underlying cache

	ctx := context.Background()

	// Test different filter combinations generate different keys
	filters1 := map[string]interface{}{"status": "active"}
	filters2 := map[string]interface{}{"status": "inactive"}
	filters3 := map[string]interface{}{"status": "active", "type": "retail"}

	merchants := []map[string]interface{}{{"id": "1"}}

	// Cache with different filters
	err := service.CacheMerchantList(ctx, filters1, merchants, 5*time.Minute)
	require.NoError(t, err)

	err = service.CacheMerchantList(ctx, filters2, merchants, 5*time.Minute)
	require.NoError(t, err)

	err = service.CacheMerchantList(ctx, filters3, merchants, 5*time.Minute)
	require.NoError(t, err)

	// Verify each filter combination has its own cache entry
	var retrieved1, retrieved2, retrieved3 []map[string]interface{}
	found1, err := service.GetMerchantList(ctx, filters1, &retrieved1)
	require.NoError(t, err)
	assert.True(t, found1)

	found2, err := service.GetMerchantList(ctx, filters2, &retrieved2)
	require.NoError(t, err)
	assert.True(t, found2)

	found3, err := service.GetMerchantList(ctx, filters3, &retrieved3)
	require.NoError(t, err)
	assert.True(t, found3)
}

// TestMerchantCacheServiceMonitoring tests cache monitoring
func TestMerchantCacheServiceMonitoring(t *testing.T) {
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         1000,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	cache := NewMemoryCache(config)

	logger := zap.NewNop()
	serviceConfig := &MerchantCacheConfig{
		EnableMonitoring: true,
	}
	service := NewMerchantCacheService(cache, serviceConfig, logger)
	defer service.Close() // This will also close the underlying cache

	// Test that monitor is created
	monitor := service.GetMonitor()
	assert.NotNil(t, monitor)

	// Test cache stats
	ctx := context.Background()
	stats, err := service.GetCacheStats(ctx)
	require.NoError(t, err)
	assert.NotNil(t, stats)
}

// TestMerchantCacheServiceErrorHandling tests error handling
func TestMerchantCacheServiceErrorHandling(t *testing.T) {
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         1000,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	cache := NewMemoryCache(config)

	logger := zap.NewNop()
	service := NewMerchantCacheService(cache, nil, logger)
	defer service.Close() // This will also close the underlying cache

	ctx := context.Background()

	// Test with invalid data (should not cause panic)
	invalidData := make(chan int) // Channels cannot be marshaled to JSON

	err := service.CacheMerchantList(ctx, nil, invalidData, 5*time.Minute)
	assert.Error(t, err) // Should return marshaling error

	// Test with nil data
	err = service.CacheMerchantList(ctx, nil, nil, 5*time.Minute)
	assert.NoError(t, err) // Should handle nil gracefully
}

// TestMerchantCacheServiceConcurrency tests concurrent operations
func TestMerchantCacheServiceConcurrency(t *testing.T) {
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         1000,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	cache := NewMemoryCache(config)

	logger := zap.NewNop()
	service := NewMerchantCacheService(cache, nil, logger)
	defer service.Close() // This will also close the underlying cache

	ctx := context.Background()

	// Test concurrent operations
	numGoroutines := 5
	numOperations := 20

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < numOperations; j++ {
				merchantID := fmt.Sprintf("merchant-%d-%d", goroutineID, j)
				merchant := map[string]interface{}{
					"id":   merchantID,
					"name": fmt.Sprintf("Merchant %d-%d", goroutineID, j),
				}

				// Cache merchant detail
				err := service.CacheMerchantDetail(ctx, merchantID, merchant, 5*time.Minute)
				if err != nil {
					t.Errorf("CacheMerchantDetail failed: %v", err)
				}

				// Retrieve merchant detail
				var retrieved map[string]interface{}
				found, err := service.GetMerchantDetail(ctx, merchantID, &retrieved)
				if err != nil {
					t.Errorf("GetMerchantDetail failed: %v", err)
				}
				if !found {
					t.Errorf("Expected to find merchant %s", merchantID)
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
