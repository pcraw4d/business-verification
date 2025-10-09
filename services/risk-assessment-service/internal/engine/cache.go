package engine

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// Cache defines the interface for caching risk assessment results
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Delete(key string)
	Clear()
	GetStats() CacheStats
}

// CacheStats holds cache statistics
type CacheStats struct {
	Hits        int64   `json:"hits"`
	Misses      int64   `json:"misses"`
	Size        int     `json:"size"`
	MaxSize     int     `json:"max_size"`
	HitRate     float64 `json:"hit_rate"`
	MemoryUsage int64   `json:"memory_usage"`
}

// CacheEntry represents a cache entry with expiration
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// InMemoryCache implements an in-memory cache with TTL
type InMemoryCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration
	maxSize int
	stats   CacheStats
	logger  *zap.Logger
	cleanup *time.Ticker
	stopCh  chan struct{}
}

// NewInMemoryCache creates a new in-memory cache
func NewInMemoryCache(ttl time.Duration, logger *zap.Logger) *InMemoryCache {
	cache := &InMemoryCache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
		maxSize: 10000, // Default max size
		logger:  logger,
		stopCh:  make(chan struct{}),
	}

	// Start cleanup goroutine
	cache.cleanup = time.NewTicker(1 * time.Minute)
	go cache.cleanupExpired()

	return cache
}

// Get retrieves a value from the cache
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		c.stats.Misses++
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		c.stats.Misses++
		// Delete expired entry
		delete(c.entries, key)
		return nil, false
	}

	c.stats.Hits++
	return entry.Value, true
}

// Set stores a value in the cache
func (c *InMemoryCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict entries
	if len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	c.entries[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}

	c.stats.Size = len(c.entries)
}

// Delete removes a value from the cache
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
	c.stats.Size = len(c.entries)
}

// Clear removes all entries from the cache
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
	c.stats.Size = 0
}

// GetStats returns cache statistics
func (c *InMemoryCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Size = len(c.entries)
	stats.MaxSize = c.maxSize

	// Calculate hit rate
	total := stats.Hits + stats.Misses
	if total > 0 {
		stats.HitRate = float64(stats.Hits) / float64(total)
	}

	return stats
}

// evictOldest removes the oldest entry from the cache
func (c *InMemoryCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.ExpiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.ExpiresAt
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

// cleanupExpired removes expired entries from the cache
func (c *InMemoryCache) cleanupExpired() {
	for {
		select {
		case <-c.cleanup.C:
			c.mu.Lock()
			now := time.Now()
			for key, entry := range c.entries {
				if now.After(entry.ExpiresAt) {
					delete(c.entries, key)
				}
			}
			c.stats.Size = len(c.entries)
			c.mu.Unlock()
		case <-c.stopCh:
			c.cleanup.Stop()
			return
		}
	}
}

// Shutdown stops the cache cleanup goroutine
func (c *InMemoryCache) Shutdown() {
	close(c.stopCh)
}

// NoOpCache implements a no-operation cache for testing
type NoOpCache struct{}

// NewNoOpCache creates a new no-operation cache
func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

// Get always returns false (cache miss)
func (c *NoOpCache) Get(key string) (interface{}, bool) {
	return nil, false
}

// Set does nothing
func (c *NoOpCache) Set(key string, value interface{}) {
	// No-op
}

// Delete does nothing
func (c *NoOpCache) Delete(key string) {
	// No-op
}

// Clear does nothing
func (c *NoOpCache) Clear() {
	// No-op
}

// GetStats returns empty stats
func (c *NoOpCache) GetStats() CacheStats {
	return CacheStats{}
}
