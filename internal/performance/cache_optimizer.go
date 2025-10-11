package performance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CacheOptimizer provides intelligent caching for performance optimization
type CacheOptimizer struct {
	logger   *zap.Logger
	profiler *Profiler
	caches   map[string]Cache
	config   *CacheConfig
	stats    *CacheStats
	mu       sync.RWMutex
}

// Cache interface for different cache implementations
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	GetStats() CacheStats
}

// InMemoryCache implements an in-memory cache with TTL
type InMemoryCache struct {
	data     map[string]*CacheItem
	mu       sync.RWMutex
	stats    CacheStats
	profiler *Profiler
	logger   *zap.Logger
}

// CacheItem represents a cached item
type CacheItem struct {
	Value        interface{} `json:"value"`
	ExpiresAt    time.Time   `json:"expires_at"`
	CreatedAt    time.Time   `json:"created_at"`
	AccessCount  int64       `json:"access_count"`
	LastAccessed time.Time   `json:"last_accessed"`
}

// CacheConfig contains cache configuration
type CacheConfig struct {
	DefaultTTL         time.Duration `json:"default_ttl"`
	MaxSize            int           `json:"max_size"`
	CleanupInterval    time.Duration `json:"cleanup_interval"`
	EnableStats        bool          `json:"enable_stats"`
	EnableProfiling    bool          `json:"enable_profiling"`
	LRUEnabled         bool          `json:"lru_enabled"`
	CompressionEnabled bool          `json:"compression_enabled"`
}

// CacheStats contains cache performance statistics
type CacheStats struct {
	Hits              int64         `json:"hits"`
	Misses            int64         `json:"misses"`
	Sets              int64         `json:"sets"`
	Deletes           int64         `json:"deletes"`
	Evictions         int64         `json:"evictions"`
	Size              int           `json:"size"`
	HitRate           float64       `json:"hit_rate"`
	AverageAccessTime time.Duration `json:"average_access_time"`
	TotalAccessTime   time.Duration `json:"total_access_time"`
	LastUpdated       time.Time     `json:"last_updated"`
}

// CacheKey represents a cache key with metadata
type CacheKey struct {
	Key       string            `json:"key"`
	Namespace string            `json:"namespace"`
	Tags      []string          `json:"tags"`
	Metadata  map[string]string `json:"metadata"`
}

// NewCacheOptimizer creates a new cache optimizer
func NewCacheOptimizer(logger *zap.Logger, profiler *Profiler, config *CacheConfig) *CacheOptimizer {
	optimizer := &CacheOptimizer{
		logger:   logger,
		profiler: profiler,
		caches:   make(map[string]Cache),
		config:   config,
		stats:    &CacheStats{},
	}

	// Create default cache
	defaultCache := NewInMemoryCache(logger, profiler, config)
	optimizer.caches["default"] = defaultCache

	// Start cleanup routine
	go optimizer.cleanup()

	return optimizer
}

// NewInMemoryCache creates a new in-memory cache
func NewInMemoryCache(logger *zap.Logger, profiler *Profiler, config *CacheConfig) *InMemoryCache {
	cache := &InMemoryCache{
		data:     make(map[string]*CacheItem),
		profiler: profiler,
		logger:   logger,
	}

	// Start cleanup routine
	go cache.cleanup(config.CleanupInterval)

	return cache
}

// Get retrieves a value from cache
func (c *InMemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if c.profiler != nil {
			c.profiler.RecordMetric("cache_get", duration)
		}
	}()

	c.mu.RLock()
	item, exists := c.data[key]
	c.mu.RUnlock()

	if !exists {
		c.stats.Misses++
		c.updateHitRate()
		return nil, fmt.Errorf("key not found: %s", key)
	}

	// Check if expired
	if time.Now().After(item.ExpiresAt) {
		c.mu.Lock()
		delete(c.data, key)
		c.mu.Unlock()
		c.stats.Misses++
		c.updateHitRate()
		return nil, fmt.Errorf("key expired: %s", key)
	}

	// Update access statistics
	c.mu.Lock()
	item.AccessCount++
	item.LastAccessed = time.Now()
	c.mu.Unlock()

	c.stats.Hits++
	c.stats.TotalAccessTime += time.Since(start)
	c.stats.AverageAccessTime = c.stats.TotalAccessTime / time.Duration(c.stats.Hits+c.stats.Misses)
	c.updateHitRate()

	return item.Value, nil
}

// Set stores a value in cache
func (c *InMemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if c.profiler != nil {
			c.profiler.RecordMetric("cache_set", duration)
		}
	}()

	now := time.Now()
	item := &CacheItem{
		Value:        value,
		ExpiresAt:    now.Add(ttl),
		CreatedAt:    now,
		AccessCount:  0,
		LastAccessed: now,
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict items
	if len(c.data) >= c.getMaxSize() {
		c.evictLRU()
	}

	c.data[key] = item
	c.stats.Sets++
	c.stats.Size = len(c.data)
	c.stats.LastUpdated = now

	return nil
}

// Delete removes a value from cache
func (c *InMemoryCache) Delete(ctx context.Context, key string) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if c.profiler != nil {
			c.profiler.RecordMetric("cache_delete", duration)
		}
	}()

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.data[key]; exists {
		delete(c.data, key)
		c.stats.Deletes++
		c.stats.Size = len(c.data)
		c.stats.LastUpdated = time.Now()
	}

	return nil
}

// Clear removes all values from cache
func (c *InMemoryCache) Clear(ctx context.Context) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if c.profiler != nil {
			c.profiler.RecordMetric("cache_clear", duration)
		}
	}()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*CacheItem)
	c.stats.Size = 0
	c.stats.LastUpdated = time.Now()

	return nil
}

// GetStats returns cache statistics
func (c *InMemoryCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Size = len(c.data)
	stats.LastUpdated = time.Now()

	return stats
}

// evictLRU evicts the least recently used item
func (c *InMemoryCache) evictLRU() {
	if len(c.data) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time

	for key, item := range c.data {
		if oldestKey == "" || item.LastAccessed.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.LastAccessed
		}
	}

	if oldestKey != "" {
		delete(c.data, oldestKey)
		c.stats.Evictions++
	}
}

// getMaxSize returns the maximum cache size
func (c *InMemoryCache) getMaxSize() int {
	// Default max size if not configured
	return 10000
}

// updateHitRate updates the cache hit rate
func (c *InMemoryCache) updateHitRate() {
	total := c.stats.Hits + c.stats.Misses
	if total > 0 {
		c.stats.HitRate = float64(c.stats.Hits) / float64(total)
	}
}

// cleanup removes expired items from cache
func (c *InMemoryCache) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		expiredKeys := make([]string, 0)

		for key, item := range c.data {
			if now.After(item.ExpiresAt) {
				expiredKeys = append(expiredKeys, key)
			}
		}

		for _, key := range expiredKeys {
			delete(c.data, key)
		}

		c.stats.Size = len(c.data)
		c.mu.Unlock()

		if len(expiredKeys) > 0 {
			c.logger.Debug("Cleaned up expired cache items",
				zap.Int("expired_count", len(expiredKeys)))
		}
	}
}

// Get retrieves a value from the specified cache
func (co *CacheOptimizer) Get(ctx context.Context, cacheName, key string) (interface{}, error) {
	co.mu.RLock()
	cache, exists := co.caches[cacheName]
	co.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("cache not found: %s", cacheName)
	}

	return cache.Get(ctx, key)
}

// Set stores a value in the specified cache
func (co *CacheOptimizer) Set(ctx context.Context, cacheName, key string, value interface{}, ttl time.Duration) error {
	co.mu.RLock()
	cache, exists := co.caches[cacheName]
	co.mu.RUnlock()

	if !exists {
		return fmt.Errorf("cache not found: %s", cacheName)
	}

	if ttl == 0 {
		ttl = co.config.DefaultTTL
	}

	return cache.Set(ctx, key, value, ttl)
}

// Delete removes a value from the specified cache
func (co *CacheOptimizer) Delete(ctx context.Context, cacheName, key string) error {
	co.mu.RLock()
	cache, exists := co.caches[cacheName]
	co.mu.RUnlock()

	if !exists {
		return fmt.Errorf("cache not found: %s", cacheName)
	}

	return cache.Delete(ctx, key)
}

// Clear removes all values from the specified cache
func (co *CacheOptimizer) Clear(ctx context.Context, cacheName string) error {
	co.mu.RLock()
	cache, exists := co.caches[cacheName]
	co.mu.RUnlock()

	if !exists {
		return fmt.Errorf("cache not found: %s", cacheName)
	}

	return cache.Clear(ctx)
}

// CreateCache creates a new cache
func (co *CacheOptimizer) CreateCache(name string) error {
	co.mu.Lock()
	defer co.mu.Unlock()

	if _, exists := co.caches[name]; exists {
		return fmt.Errorf("cache already exists: %s", name)
	}

	cache := NewInMemoryCache(co.logger, co.profiler, co.config)
	co.caches[name] = cache

	co.logger.Info("Cache created", zap.String("name", name))
	return nil
}

// GetCacheStats returns statistics for all caches
func (co *CacheOptimizer) GetCacheStats() map[string]CacheStats {
	co.mu.RLock()
	defer co.mu.RUnlock()

	stats := make(map[string]CacheStats)
	for name, cache := range co.caches {
		stats[name] = cache.GetStats()
	}

	return stats
}

// GetOverallStats returns overall cache statistics
func (co *CacheOptimizer) GetOverallStats() *CacheStats {
	co.mu.RLock()
	defer co.mu.RUnlock()

	overall := &CacheStats{
		LastUpdated: time.Now(),
	}

	for _, cache := range co.caches {
		stats := cache.GetStats()
		overall.Hits += stats.Hits
		overall.Misses += stats.Misses
		overall.Sets += stats.Sets
		overall.Deletes += stats.Deletes
		overall.Evictions += stats.Evictions
		overall.Size += stats.Size
		overall.TotalAccessTime += stats.TotalAccessTime
	}

	// Calculate overall hit rate
	total := overall.Hits + overall.Misses
	if total > 0 {
		overall.HitRate = float64(overall.Hits) / float64(total)
	}

	// Calculate overall average access time
	if total > 0 {
		overall.AverageAccessTime = overall.TotalAccessTime / time.Duration(total)
	}

	return overall
}

// cleanup performs periodic cache maintenance
func (co *CacheOptimizer) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		stats := co.GetOverallStats()

		// Log cache performance
		if co.config.EnableStats {
			co.logger.Info("Cache performance summary",
				zap.Float64("hit_rate", stats.HitRate),
				zap.Int64("hits", stats.Hits),
				zap.Int64("misses", stats.Misses),
				zap.Int("total_size", stats.Size),
				zap.Duration("avg_access_time", stats.AverageAccessTime))
		}

		// Warn about low hit rates
		if stats.HitRate < 0.7 && (stats.Hits+stats.Misses) > 100 {
			co.logger.Warn("Low cache hit rate detected",
				zap.Float64("hit_rate", stats.HitRate),
				zap.Int64("total_requests", stats.Hits+stats.Misses))
		}
	}
}

// GetPerformanceReport generates a cache performance report
func (co *CacheOptimizer) GetPerformanceReport() string {
	stats := co.GetOverallStats()
	cacheStats := co.GetCacheStats()

	report := fmt.Sprintf("=== CACHE PERFORMANCE REPORT ===\n")
	report += fmt.Sprintf("Generated: %s\n", stats.LastUpdated.Format(time.RFC3339))
	report += fmt.Sprintf("Overall Hit Rate: %.2f%%\n", stats.HitRate*100)
	report += fmt.Sprintf("Total Hits: %d\n", stats.Hits)
	report += fmt.Sprintf("Total Misses: %d\n", stats.Misses)
	report += fmt.Sprintf("Total Sets: %d\n", stats.Sets)
	report += fmt.Sprintf("Total Deletes: %d\n", stats.Deletes)
	report += fmt.Sprintf("Total Evictions: %d\n", stats.Evictions)
	report += fmt.Sprintf("Total Size: %d\n", stats.Size)
	report += fmt.Sprintf("Average Access Time: %v\n", stats.AverageAccessTime)

	report += fmt.Sprintf("\n=== CACHE BREAKDOWN ===\n")
	for name, cacheStat := range cacheStats {
		report += fmt.Sprintf("%s:\n", name)
		report += fmt.Sprintf("  Hit Rate: %.2f%%\n", cacheStat.HitRate*100)
		report += fmt.Sprintf("  Hits: %d\n", cacheStat.Hits)
		report += fmt.Sprintf("  Misses: %d\n", cacheStat.Misses)
		report += fmt.Sprintf("  Size: %d\n", cacheStat.Size)
		report += fmt.Sprintf("  Average Access Time: %v\n", cacheStat.AverageAccessTime)
		report += fmt.Sprintf("\n")
	}

	report += fmt.Sprintf("\n=== CONFIGURATION ===\n")
	report += fmt.Sprintf("Default TTL: %v\n", co.config.DefaultTTL)
	report += fmt.Sprintf("Max Size: %d\n", co.config.MaxSize)
	report += fmt.Sprintf("Cleanup Interval: %v\n", co.config.CleanupInterval)
	report += fmt.Sprintf("LRU Enabled: %t\n", co.config.LRUEnabled)
	report += fmt.Sprintf("Compression Enabled: %t\n", co.config.CompressionEnabled)

	return report
}

// DefaultCacheConfig returns a default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		DefaultTTL:         5 * time.Minute,
		MaxSize:            10000,
		CleanupInterval:    1 * time.Minute,
		EnableStats:        true,
		EnableProfiling:    true,
		LRUEnabled:         true,
		CompressionEnabled: false,
	}
}
