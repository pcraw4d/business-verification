package business_intelligence_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/your-org/your-repo/internal/modules/business_intelligence"
)

// MockCache is a mock implementation of Cache interface
type MockCache struct {
	mock.Mock
	data map[string]interface{}
	ttl  map[string]time.Time
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]interface{}),
		ttl:  make(map[string]time.Time),
	}
}

func (m *MockCache) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCache) GetType() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCache) Get(key string) (interface{}, bool) {
	args := m.Called(key)

	// Check TTL
	if expireTime, exists := m.ttl[key]; exists && time.Now().After(expireTime) {
		delete(m.data, key)
		delete(m.ttl, key)
		return nil, false
	}

	if value, exists := m.data[key]; exists {
		return value, true
	}
	return args.Get(0), args.Bool(1)
}

func (m *MockCache) Set(key string, value interface{}, ttl time.Duration) error {
	args := m.Called(key, value, ttl)
	m.data[key] = value
	if ttl > 0 {
		m.ttl[key] = time.Now().Add(ttl)
	}
	return args.Error(0)
}

func (m *MockCache) Delete(key string) error {
	args := m.Called(key)
	delete(m.data, key)
	delete(m.ttl, key)
	return args.Error(0)
}

func (m *MockCache) Clear() error {
	args := m.Called()
	m.data = make(map[string]interface{})
	m.ttl = make(map[string]time.Time)
	return args.Error(0)
}

func (m *MockCache) Size() int64 {
	args := m.Called()
	return int64(len(m.data))
}

func (m *MockCache) Keys() []string {
	args := m.Called()
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

func (m *MockCache) GetStats() business_intelligence.CacheStats {
	args := m.Called()
	return args.Get(0).(business_intelligence.CacheStats)
}

func (m *MockCache) IsHealthy() bool {
	args := m.Called()
	return args.Bool(0)
}

// MockCacheStrategy is a mock implementation of CacheStrategy
type MockCacheStrategy struct {
	mock.Mock
}

func (m *MockCacheStrategy) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCacheStrategy) GetType() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCacheStrategy) ShouldCache(key string, value interface{}) bool {
	args := m.Called(key, value)
	return args.Bool(0)
}

func (m *MockCacheStrategy) GetTTL(key string, value interface{}) time.Duration {
	args := m.Called(key, value)
	return args.Get(0).(time.Duration)
}

func (m *MockCacheStrategy) GetPriority(key string, value interface{}) int {
	args := m.Called(key, value)
	return args.Int(0)
}

func (m *MockCacheStrategy) GetCompressionLevel(key string, value interface{}) int {
	args := m.Called(key, value)
	return args.Int(0)
}

func (m *MockCacheStrategy) GetEncryptionLevel(key string, value interface{}) int {
	args := m.Called(key, value)
	return args.Int(0)
}

// TestDataCachingSystem_Performance tests the performance of the caching system
func TestDataCachingSystem_Performance(t *testing.T) {
	logger := zap.NewNop()

	// Configure caching system
	config := business_intelligence.CacheConfig{
		DefaultTTL:                5 * time.Minute,
		MaxCacheSize:              1000,
		MaxItemSize:               1024 * 1024, // 1MB
		EnableCompression:         true,
		EnableEncryption:          false,
		EnableSerialization:       true,
		DefaultStrategy:           "performance_strategy",
		EnableCacheWarming:        true,
		WarmingInterval:           1 * time.Minute,
		EnableCacheInvalidation:   true,
		DefaultEvictionPolicy:     "lru",
		EvictionCheckInterval:     30 * time.Second,
		MaxEvictionBatchSize:      10,
		EnableAsyncOperations:     true,
		AsyncOperationTimeout:     5 * time.Second,
		EnableBatchOperations:     true,
		BatchSize:                 50,
		EnableMetrics:             true,
		MetricsCollectionInterval: 1 * time.Minute,
		EnableCacheStatistics:     true,
		WarmingStrategies:         []string{"preload", "predictive"},
		WarmingDataSources:        []string{"business_data_api", "internal_db"},
		WarmingPriority: map[string]int{
			"business_data_api": 1,
			"internal_db":       2,
		},
		InvalidationStrategies: []string{"pattern", "time_based"},
		InvalidationTriggers:   []string{"data_update", "schema_change"},
		InvalidationTimeout:    10 * time.Second,
	}

	// Create caching system
	cachingSystem := business_intelligence.NewDataCachingSystem(config, logger)

	// Create mock cache
	mockCache := NewMockCache()
	mockCache.On("GetName").Return("test_cache")
	mockCache.On("GetType").Return("memory")
	mockCache.On("IsHealthy").Return(true)
	mockCache.On("GetStats").Return(business_intelligence.CacheStats{
		CacheName:         "test_cache",
		CacheType:         "memory",
		TotalItems:        0,
		TotalSize:         0,
		HitCount:          0,
		MissCount:         0,
		HitRate:           0.0,
		MissRate:          0.0,
		EvictionCount:     0,
		ExpirationCount:   0,
		LastAccessTime:    time.Now(),
		LastUpdateTime:    time.Now(),
		AverageAccessTime: 0,
		IsHealthy:         true,
	})

	// Create mock strategy
	mockStrategy := &MockCacheStrategy{}
	mockStrategy.On("GetName").Return("performance_strategy")
	mockStrategy.On("GetType").Return("ttl_based")
	mockStrategy.On("ShouldCache", mock.Anything, mock.Anything).Return(true)
	mockStrategy.On("GetTTL", mock.Anything, mock.Anything).Return(5 * time.Minute)
	mockStrategy.On("GetPriority", mock.Anything, mock.Anything).Return(1)
	mockStrategy.On("GetCompressionLevel", mock.Anything, mock.Anything).Return(6)
	mockStrategy.On("GetEncryptionLevel", mock.Anything, mock.Anything).Return(0)

	// Register cache and strategy
	err := cachingSystem.RegisterCache(mockCache)
	require.NoError(t, err)

	err = cachingSystem.RegisterCacheStrategy(mockStrategy)
	require.NoError(t, err)

	t.Run("Cache Hit Performance", func(t *testing.T) {
		// Setup mock for cache hit
		mockCache.On("Get", "test_key_1").Return("cached_value_1", true).Once()

		// Test cache hit performance
		ctx := context.Background()
		startTime := time.Now()

		value, err := cachingSystem.Get(ctx, "test_cache", "test_key_1")

		duration := time.Since(startTime)

		// Assertions
		require.NoError(t, err)
		assert.Equal(t, "cached_value_1", value)
		assert.Less(t, duration, 10*time.Millisecond, "Cache hit should be very fast")

		mockCache.AssertExpectations(t)
	})

	t.Run("Cache Miss Performance", func(t *testing.T) {
		// Setup mock for cache miss
		mockCache.On("Get", "test_key_2").Return(nil, false).Once()

		// Test cache miss performance
		ctx := context.Background()
		startTime := time.Now()

		value, err := cachingSystem.Get(ctx, "test_cache", "test_key_2")

		duration := time.Since(startTime)

		// Assertions
		require.NoError(t, err)
		assert.Nil(t, value)
		assert.Less(t, duration, 5*time.Millisecond, "Cache miss should be fast")

		mockCache.AssertExpectations(t)
	})

	t.Run("Cache Set Performance", func(t *testing.T) {
		// Setup mock for cache set
		mockCache.On("Set", "test_key_3", "test_value_3", 5*time.Minute).Return(nil).Once()

		// Test cache set performance
		ctx := context.Background()
		startTime := time.Now()

		err := cachingSystem.Set(ctx, "test_cache", "test_key_3", "test_value_3", 5*time.Minute)

		duration := time.Since(startTime)

		// Assertions
		require.NoError(t, err)
		assert.Less(t, duration, 10*time.Millisecond, "Cache set should be fast")

		mockCache.AssertExpectations(t)
	})

	t.Run("Batch Operations Performance", func(t *testing.T) {
		// Setup mocks for batch operations
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("batch_key_%d", i)
			value := fmt.Sprintf("batch_value_%d", i)
			mockCache.On("Set", key, value, 5*time.Minute).Return(nil).Once()
		}

		// Test batch set performance
		ctx := context.Background()
		startTime := time.Now()

		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("batch_key_%d", i)
			value := fmt.Sprintf("batch_value_%d", i)
			err := cachingSystem.Set(ctx, "test_cache", key, value, 5*time.Minute)
			require.NoError(t, err)
		}

		duration := time.Since(startTime)

		// Assertions
		assert.Less(t, duration, 100*time.Millisecond, "Batch operations should be efficient")
		assert.Less(t, duration/100, 2*time.Millisecond, "Average operation should be fast")

		mockCache.AssertExpectations(t)
	})

	t.Run("Concurrent Access Performance", func(t *testing.T) {
		// Setup mocks for concurrent access
		for i := 0; i < 50; i++ {
			key := fmt.Sprintf("concurrent_key_%d", i)
			mockCache.On("Get", key).Return(fmt.Sprintf("concurrent_value_%d", i), true).Once()
		}

		// Test concurrent access performance
		ctx := context.Background()
		startTime := time.Now()

		// Simulate concurrent access
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(workerID int) {
				for j := 0; j < 5; j++ {
					key := fmt.Sprintf("concurrent_key_%d", workerID*5+j)
					_, err := cachingSystem.Get(ctx, "test_cache", key)
					require.NoError(t, err)
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		duration := time.Since(startTime)

		// Assertions
		assert.Less(t, duration, 200*time.Millisecond, "Concurrent access should be efficient")

		mockCache.AssertExpectations(t)
	})

	t.Run("Cache Eviction Performance", func(t *testing.T) {
		// Setup mocks for eviction
		mockCache.On("Keys").Return([]string{"key1", "key2", "key3", "key4", "key5"}).Once()
		mockCache.On("Delete", mock.Anything).Return(nil).Times(5)

		// Test eviction performance
		ctx := context.Background()
		startTime := time.Now()

		// Simulate eviction by clearing cache
		err := cachingSystem.Clear(ctx, "test_cache")

		duration := time.Since(startTime)

		// Assertions
		require.NoError(t, err)
		assert.Less(t, duration, 50*time.Millisecond, "Cache eviction should be fast")

		mockCache.AssertExpectations(t)
	})
}

// TestDataCachingSystem_CompressionPerformance tests compression performance
func TestDataCachingSystem_CompressionPerformance(t *testing.T) {
	logger := zap.NewNop()

	// Configure caching system with compression enabled
	config := business_intelligence.CacheConfig{
		DefaultTTL:          5 * time.Minute,
		EnableCompression:   true,
		EnableEncryption:    false,
		EnableSerialization: true,
		EnableMetrics:       true,
	}

	// Create caching system
	cachingSystem := business_intelligence.NewDataCachingSystem(config, logger)

	// Create mock cache
	mockCache := NewMockCache()
	mockCache.On("GetName").Return("compression_test_cache")
	mockCache.On("GetType").Return("memory")
	mockCache.On("IsHealthy").Return(true)
	mockCache.On("GetStats").Return(business_intelligence.CacheStats{
		CacheName:         "compression_test_cache",
		CacheType:         "memory",
		TotalItems:        0,
		TotalSize:         0,
		HitCount:          0,
		MissCount:         0,
		HitRate:           0.0,
		MissRate:          0.0,
		EvictionCount:     0,
		ExpirationCount:   0,
		LastAccessTime:    time.Now(),
		LastUpdateTime:    time.Now(),
		AverageAccessTime: 0,
		IsHealthy:         true,
	})

	// Register cache
	err := cachingSystem.RegisterCache(mockCache)
	require.NoError(t, err)

	t.Run("Large Data Compression", func(t *testing.T) {
		// Create large test data
		largeData := make(map[string]interface{})
		for i := 0; i < 1000; i++ {
			largeData[fmt.Sprintf("field_%d", i)] = fmt.Sprintf("This is a long string value for field %d that should be compressed efficiently", i)
		}

		// Setup mock for cache set with compression
		mockCache.On("Set", "large_data_key", mock.Anything, 5*time.Minute).Return(nil).Once()

		// Test compression performance
		ctx := context.Background()
		startTime := time.Now()

		err := cachingSystem.Set(ctx, "compression_test_cache", "large_data_key", largeData, 5*time.Minute)

		duration := time.Since(startTime)

		// Assertions
		require.NoError(t, err)
		assert.Less(t, duration, 100*time.Millisecond, "Compression should be reasonably fast")

		mockCache.AssertExpectations(t)
	})

	t.Run("Compression Ratio", func(t *testing.T) {
		// Create repetitive data that should compress well
		repetitiveData := map[string]interface{}{
			"repeated_string": "This is a repeated string that should compress very well. " +
				"This is a repeated string that should compress very well. " +
				"This is a repeated string that should compress very well. " +
				"This is a repeated string that should compress very well. " +
				"This is a repeated string that should compress very well. ",
			"numbers": []int{1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5},
		}

		// Setup mock for cache set
		mockCache.On("Set", "repetitive_data_key", mock.Anything, 5*time.Minute).Return(nil).Once()

		// Test compression with repetitive data
		ctx := context.Background()
		startTime := time.Now()

		err := cachingSystem.Set(ctx, "compression_test_cache", "repetitive_data_key", repetitiveData, 5*time.Minute)

		duration := time.Since(startTime)

		// Assertions
		require.NoError(t, err)
		assert.Less(t, duration, 50*time.Millisecond, "Compression of repetitive data should be fast")

		mockCache.AssertExpectations(t)
	})
}

// TestDataCachingSystem_SerializationPerformance tests serialization performance
func TestDataCachingSystem_SerializationPerformance(t *testing.T) {
	logger := zap.NewNop()

	// Configure caching system with serialization enabled
	config := business_intelligence.CacheConfig{
		DefaultTTL:          5 * time.Minute,
		EnableCompression:   false,
		EnableEncryption:    false,
		EnableSerialization: true,
		EnableMetrics:       true,
	}

	// Create caching system
	cachingSystem := business_intelligence.NewDataCachingSystem(config, logger)

	// Create mock cache
	mockCache := NewMockCache()
	mockCache.On("GetName").Return("serialization_test_cache")
	mockCache.On("GetType").Return("memory")
	mockCache.On("IsHealthy").Return(true)
	mockCache.On("GetStats").Return(business_intelligence.CacheStats{
		CacheName:         "serialization_test_cache",
		CacheType:         "memory",
		TotalItems:        0,
		TotalSize:         0,
		HitCount:          0,
		MissCount:         0,
		HitRate:           0.0,
		MissRate:          0.0,
		EvictionCount:     0,
		ExpirationCount:   0,
		LastAccessTime:    time.Now(),
		LastUpdateTime:    time.Now(),
		AverageAccessTime: 0,
		IsHealthy:         true,
	})

	// Register cache
	err := cachingSystem.RegisterCache(mockCache)
	require.NoError(t, err)

	t.Run("Complex Object Serialization", func(t *testing.T) {
		// Create complex test data
		complexData := map[string]interface{}{
			"business_info": map[string]interface{}{
				"name":      "Complex Business Inc",
				"industry":  "Technology",
				"revenue":   1000000.50,
				"employees": 150,
				"locations": []string{"San Francisco", "New York", "London"},
				"founded":   time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC),
				"is_public": true,
				"metadata": map[string]interface{}{
					"tags":        []string{"startup", "tech", "ai"},
					"categories":  []string{"software", "services"},
					"competitors": []string{"Company A", "Company B", "Company C"},
				},
			},
			"financial_data": map[string]interface{}{
				"revenue_history": []float64{500000, 750000, 1000000, 1250000},
				"expenses": map[string]float64{
					"salaries":  600000,
					"rent":      120000,
					"marketing": 80000,
					"equipment": 50000,
				},
				"profit_margin": 0.15,
			},
		}

		// Setup mock for cache set
		mockCache.On("Set", "complex_data_key", mock.Anything, 5*time.Minute).Return(nil).Once()

		// Test serialization performance
		ctx := context.Background()
		startTime := time.Now()

		err := cachingSystem.Set(ctx, "serialization_test_cache", "complex_data_key", complexData, 5*time.Minute)

		duration := time.Since(startTime)

		// Assertions
		require.NoError(t, err)
		assert.Less(t, duration, 50*time.Millisecond, "Complex object serialization should be fast")

		mockCache.AssertExpectations(t)
	})

	t.Run("Serialization Round Trip", func(t *testing.T) {
		// Create test data
		testData := map[string]interface{}{
			"string_field": "test string",
			"number_field": 42,
			"float_field":  3.14159,
			"bool_field":   true,
			"array_field":  []interface{}{1, 2, 3, "four", 5.0},
			"object_field": map[string]interface{}{
				"nested_string": "nested value",
				"nested_number": 100,
			},
		}

		// Setup mocks for round trip
		mockCache.On("Set", "round_trip_key", mock.Anything, 5*time.Minute).Return(nil).Once()
		mockCache.On("Get", "round_trip_key").Return(mock.Anything, true).Once()

		// Test serialization round trip performance
		ctx := context.Background()
		startTime := time.Now()

		// Set data
		err := cachingSystem.Set(ctx, "serialization_test_cache", "round_trip_key", testData, 5*time.Minute)
		require.NoError(t, err)

		// Get data
		retrievedData, err := cachingSystem.Get(ctx, "serialization_test_cache", "round_trip_key")

		duration := time.Since(startTime)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, retrievedData)
		assert.Less(t, duration, 30*time.Millisecond, "Serialization round trip should be fast")

		mockCache.AssertExpectations(t)
	})
}

// BenchmarkDataCachingSystem benchmarks the caching system performance
func BenchmarkDataCachingSystem(b *testing.B) {
	logger := zap.NewNop()

	// Configure caching system
	config := business_intelligence.CacheConfig{
		DefaultTTL:          5 * time.Minute,
		EnableCompression:   true,
		EnableEncryption:    false,
		EnableSerialization: true,
		EnableMetrics:       false, // Disable metrics for benchmarking
	}

	// Create caching system
	cachingSystem := business_intelligence.NewDataCachingSystem(config, logger)

	// Create mock cache
	mockCache := NewMockCache()
	mockCache.On("GetName").Return("benchmark_cache")
	mockCache.On("GetType").Return("memory")
	mockCache.On("IsHealthy").Return(true)
	mockCache.On("GetStats").Return(business_intelligence.CacheStats{
		CacheName:         "benchmark_cache",
		CacheType:         "memory",
		TotalItems:        0,
		TotalSize:         0,
		HitCount:          0,
		MissCount:         0,
		HitRate:           0.0,
		MissRate:          0.0,
		EvictionCount:     0,
		ExpirationCount:   0,
		LastAccessTime:    time.Now(),
		LastUpdateTime:    time.Now(),
		AverageAccessTime: 0,
		IsHealthy:         true,
	})

	// Register cache
	err := cachingSystem.RegisterCache(mockCache)
	if err != nil {
		b.Fatalf("Failed to register cache: %v", err)
	}

	// Setup mocks for benchmarking
	mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockCache.On("Get", mock.Anything).Return("benchmark_value", true)

	b.ResetTimer()

	// Benchmark cache operations
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		key := fmt.Sprintf("benchmark_key_%d", i)
		value := map[string]interface{}{
			"business_name": fmt.Sprintf("Benchmark Company %d", i),
			"industry":      "Technology",
			"revenue":       1000000 + float64(i*100000),
			"employees":     50 + i*10,
		}

		// Set operation
		err := cachingSystem.Set(ctx, "benchmark_cache", key, value, 5*time.Minute)
		if err != nil {
			b.Fatalf("Cache set failed: %v", err)
		}

		// Get operation
		_, err = cachingSystem.Get(ctx, "benchmark_cache", key)
		if err != nil {
			b.Fatalf("Cache get failed: %v", err)
		}
	}
}

// BenchmarkCacheHitRate benchmarks cache hit rate performance
func BenchmarkCacheHitRate(b *testing.B) {
	logger := zap.NewNop()

	// Configure caching system
	config := business_intelligence.CacheConfig{
		DefaultTTL:          5 * time.Minute,
		EnableCompression:   false,
		EnableEncryption:    false,
		EnableSerialization: false,
		EnableMetrics:       false,
	}

	// Create caching system
	cachingSystem := business_intelligence.NewDataCachingSystem(config, logger)

	// Create mock cache
	mockCache := NewMockCache()
	mockCache.On("GetName").Return("hit_rate_cache")
	mockCache.On("GetType").Return("memory")
	mockCache.On("IsHealthy").Return(true)
	mockCache.On("GetStats").Return(business_intelligence.CacheStats{
		CacheName:         "hit_rate_cache",
		CacheType:         "memory",
		TotalItems:        0,
		TotalSize:         0,
		HitCount:          0,
		MissCount:         0,
		HitRate:           0.0,
		MissRate:          0.0,
		EvictionCount:     0,
		ExpirationCount:   0,
		LastAccessTime:    time.Now(),
		LastUpdateTime:    time.Now(),
		AverageAccessTime: 0,
		IsHealthy:         true,
	})

	// Register cache
	err := cachingSystem.RegisterCache(mockCache)
	if err != nil {
		b.Fatalf("Failed to register cache: %v", err)
	}

	// Setup mocks for hit rate benchmarking
	mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Simulate 80% hit rate
	mockCache.On("Get", mock.Anything).Return(func(key string) (interface{}, bool) {
		// Simulate hit rate based on key pattern
		if len(key) > 10 && key[len(key)-1]%5 != 0 {
			return "cached_value", true
		}
		return nil, false
	})

	b.ResetTimer()

	// Benchmark cache hit rate
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		key := fmt.Sprintf("hit_rate_key_%d", i)

		// Get operation (simulating hit rate)
		_, err := cachingSystem.Get(ctx, "hit_rate_cache", key)
		if err != nil {
			b.Fatalf("Cache get failed: %v", err)
		}

		// Set operation for misses
		if i%5 == 0 {
			value := fmt.Sprintf("value_%d", i)
			err := cachingSystem.Set(ctx, "hit_rate_cache", key, value, 5*time.Minute)
			if err != nil {
				b.Fatalf("Cache set failed: %v", err)
			}
		}
	}
}
