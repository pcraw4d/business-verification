package performance

import (
	"sync"
	"testing"
	"time"
)

// MockCache is a simple in-memory cache for testing
type MockCache struct {
	data map[string]interface{}
	ttl  map[string]time.Time
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]interface{}),
		ttl:  make(map[string]time.Time),
	}
}

func (c *MockCache) Get(key string) (interface{}, bool) {
	if expiry, exists := c.ttl[key]; exists {
		if time.Now().After(expiry) {
			delete(c.data, key)
			delete(c.ttl, key)
			return nil, false
		}
	}
	value, exists := c.data[key]
	return value, exists
}

func (c *MockCache) Set(key string, value interface{}, ttl time.Duration) {
	c.data[key] = value
	c.ttl[key] = time.Now().Add(ttl)
}

// TestCacheHitPerformance tests cache hit performance
func TestCacheHitPerformance(t *testing.T) {
	cache := NewMockCache()
	cache.Set("test-key", "test-value", 5*time.Minute)

	iterations := 10000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		_, _ = cache.Get("test-key")
	}

	duration := time.Since(start)
	avgLatency := duration / time.Duration(iterations)

	if avgLatency > 1*time.Microsecond {
		t.Errorf("Cache hit latency too high: %v (expected < 1μs)", avgLatency)
	}

	t.Logf("Cache Performance:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Total Time: %v", duration)
	t.Logf("  Average Latency: %v", avgLatency)
	t.Logf("  Operations/Second: %.2f", float64(iterations)/duration.Seconds())
}

// TestCacheMissPerformance tests cache miss performance
func TestCacheMissPerformance(t *testing.T) {
	cache := NewMockCache()

	iterations := 10000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		_, _ = cache.Get("non-existent-key")
	}

	duration := time.Since(start)
	avgLatency := duration / time.Duration(iterations)

	if avgLatency > 1*time.Microsecond {
		t.Errorf("Cache miss latency too high: %v (expected < 1μs)", avgLatency)
	}

	t.Logf("Cache Miss Performance:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Total Time: %v", duration)
	t.Logf("  Average Latency: %v", avgLatency)
}

// TestCacheTTLPerformance tests cache TTL expiration performance
func TestCacheTTLPerformance(t *testing.T) {
	cache := NewMockCache()
	cache.Set("expired-key", "value", 1*time.Nanosecond)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	start := time.Now()
	_, exists := cache.Get("expired-key")
	duration := time.Since(start)

	if exists {
		t.Error("Expected cache miss for expired key")
	}

	if duration > 1*time.Millisecond {
		t.Errorf("TTL check took too long: %v", duration)
	}

	t.Logf("TTL Check Performance: %v", duration)
}

// TestCacheConcurrentAccess tests cache performance under concurrent access
func TestCacheConcurrentAccess(t *testing.T) {
	cache := NewMockCache()
	cache.Set("concurrent-key", "value", 5*time.Minute)

	goroutines := 100
	iterationsPerGoroutine := 100

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterationsPerGoroutine; j++ {
				_, _ = cache.Get("concurrent-key")
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	totalOperations := goroutines * iterationsPerGoroutine
	avgLatency := duration / time.Duration(totalOperations)

	t.Logf("Concurrent Cache Performance:")
	t.Logf("  Goroutines: %d", goroutines)
	t.Logf("  Operations per Goroutine: %d", iterationsPerGoroutine)
	t.Logf("  Total Operations: %d", totalOperations)
	t.Logf("  Total Time: %v", duration)
	t.Logf("  Average Latency: %v", avgLatency)
	t.Logf("  Operations/Second: %.2f", float64(totalOperations)/duration.Seconds())
}

