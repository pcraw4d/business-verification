package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MemoryCacheImpl implements an in-memory cache with TTL and eviction
type MemoryCacheImpl struct {
	mu            sync.RWMutex
	data          map[string]*cacheItem
	config        *CacheConfig
	stats         *CacheStats
	cleanupTicker *time.Ticker
	stopChan      chan struct{}
}

// cacheItem represents an item stored in the memory cache
type cacheItem struct {
	Value       []byte
	CreatedAt   time.Time
	ExpiresAt   time.Time
	TTL         time.Duration
	Size        int64
	AccessCount int64
	LastAccess  time.Time
}

// NewMemoryCache creates a new memory cache instance
func NewMemoryCache(config *CacheConfig) *MemoryCacheImpl {
	if config == nil {
		config = &CacheConfig{
			Type:            MemoryCache,
			DefaultTTL:      1 * time.Hour,
			MaxSize:         1000,
			CleanupInterval: 5 * time.Minute,
		}
	}

	cache := &MemoryCacheImpl{
		data:     make(map[string]*cacheItem),
		config:   config,
		stats:    &CacheStats{},
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine
	cache.startCleanup()

	return cache
}

// Get retrieves a value from the cache
func (mc *MemoryCacheImpl) Get(ctx context.Context, key string) ([]byte, error) {
	mc.mu.RLock()
	item, exists := mc.data[key]
	mc.mu.RUnlock()

	if !exists {
		mc.recordMiss()
		return nil, &CacheNotFoundError{Key: key}
	}

	// Check if item has expired
	if !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt) {
		mc.mu.Lock()
		delete(mc.data, key)
		mc.stats.ExpiredCount++
		mc.mu.Unlock()
		mc.recordMiss()
		return nil, &CacheNotFoundError{Key: key}
	}

	// Update access statistics
	mc.mu.Lock()
	item.AccessCount++
	item.LastAccess = time.Now()
	mc.stats.HitCount++
	mc.mu.Unlock()

	return item.Value, nil
}

// Set stores a value in the cache
func (mc *MemoryCacheImpl) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl == 0 {
		ttl = mc.config.DefaultTTL
	}

	// Check if cache is full and evict if necessary
	mc.mu.Lock()
	if int64(len(mc.data)) >= mc.config.MaxSize {
		mc.evictLRU()
	}
	mc.mu.Unlock()

	item := &cacheItem{
		Value:       value,
		CreatedAt:   time.Now(),
		TTL:         ttl,
		Size:        int64(len(value)),
		AccessCount: 0,
		LastAccess:  time.Now(),
	}

	if ttl > 0 {
		item.ExpiresAt = time.Now().Add(ttl)
	}

	mc.mu.Lock()
	mc.data[key] = item
	mc.stats.Size = int64(len(mc.data))
	mc.mu.Unlock()

	return nil
}

// Delete removes a value from the cache
func (mc *MemoryCacheImpl) Delete(ctx context.Context, key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.data[key]; exists {
		delete(mc.data, key)
		mc.stats.Size = int64(len(mc.data))
		return nil
	}

	return &CacheNotFoundError{Key: key}
}

// Exists checks if a key exists in the cache
func (mc *MemoryCacheImpl) Exists(ctx context.Context, key string) (bool, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	item, exists := mc.data[key]
	if !exists {
		return false, nil
	}

	// Check if item has expired
	if !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt) {
		return false, nil
	}

	return true, nil
}

// GetTTL returns the remaining TTL for a key
func (mc *MemoryCacheImpl) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	item, exists := mc.data[key]
	if !exists {
		return 0, &CacheNotFoundError{Key: key}
	}

	if item.ExpiresAt.IsZero() {
		return 0, nil // No expiration
	}

	remaining := time.Until(item.ExpiresAt)
	if remaining <= 0 {
		return 0, &CacheNotFoundError{Key: key}
	}

	return remaining, nil
}

// SetTTL updates the TTL for an existing key
func (mc *MemoryCacheImpl) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	item, exists := mc.data[key]
	if !exists {
		return &CacheNotFoundError{Key: key}
	}

	item.TTL = ttl
	if ttl > 0 {
		item.ExpiresAt = time.Now().Add(ttl)
	} else {
		item.ExpiresAt = time.Time{} // No expiration
	}

	return nil
}

// Clear removes all keys from the cache
func (mc *MemoryCacheImpl) Clear(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.data = make(map[string]*cacheItem)
	mc.stats.Size = 0
	return nil
}

// GetStats returns cache statistics
func (mc *MemoryCacheImpl) GetStats(ctx context.Context) (*CacheStats, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Calculate hit rate
	totalRequests := mc.stats.HitCount + mc.stats.MissCount
	hitRate := 0.0
	if totalRequests > 0 {
		hitRate = float64(mc.stats.HitCount) / float64(totalRequests)
	}

	stats := &CacheStats{
		HitCount:      mc.stats.HitCount,
		MissCount:     mc.stats.MissCount,
		HitRate:       hitRate,
		Size:          mc.stats.Size,
		MaxSize:       mc.config.MaxSize,
		EvictionCount: mc.stats.EvictionCount,
		ExpiredCount:  mc.stats.ExpiredCount,
	}

	return stats, nil
}

// Close closes the cache and stops cleanup
func (mc *MemoryCacheImpl) Close() error {
	close(mc.stopChan)
	if mc.cleanupTicker != nil {
		mc.cleanupTicker.Stop()
	}
	return nil
}

// startCleanup starts the cleanup goroutine
func (mc *MemoryCacheImpl) startCleanup() {
	if mc.config.CleanupInterval <= 0 {
		return
	}

	mc.cleanupTicker = time.NewTicker(mc.config.CleanupInterval)
	go func() {
		for {
			select {
			case <-mc.cleanupTicker.C:
				mc.cleanup()
			case <-mc.stopChan:
				return
			}
		}
	}()
}

// cleanup removes expired items from the cache
func (mc *MemoryCacheImpl) cleanup() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for key, item := range mc.data {
		if !item.ExpiresAt.IsZero() && now.After(item.ExpiresAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(mc.data, key)
		mc.stats.ExpiredCount++
	}

	mc.stats.Size = int64(len(mc.data))
}

// evictLRU evicts the least recently used item
func (mc *MemoryCacheImpl) evictLRU() {
	if len(mc.data) == 0 {
		return
	}

	var oldestKey string
	var oldestAccess time.Time
	var lowestAccessCount int64

	first := true
	for key, item := range mc.data {
		if first {
			oldestKey = key
			oldestAccess = item.LastAccess
			lowestAccessCount = item.AccessCount
			first = false
			continue
		}

		// Prefer items with lower access count
		if item.AccessCount < lowestAccessCount {
			oldestKey = key
			oldestAccess = item.LastAccess
			lowestAccessCount = item.AccessCount
		} else if item.AccessCount == lowestAccessCount && item.LastAccess.Before(oldestAccess) {
			// If access counts are equal, prefer older access time
			oldestKey = key
			oldestAccess = item.LastAccess
		}
	}

	if oldestKey != "" {
		delete(mc.data, oldestKey)
		mc.stats.EvictionCount++
	}
}

// recordMiss records a cache miss
func (mc *MemoryCacheImpl) recordMiss() {
	mc.mu.Lock()
	mc.stats.MissCount++
	mc.mu.Unlock()
}

// GetKeys returns all keys in the cache
func (mc *MemoryCacheImpl) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	keys := make([]string, 0, len(mc.data))
	for key := range mc.data {
		// Simple pattern matching - can be enhanced with regex
		if pattern == "" || key == pattern {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// GetEntries returns multiple entries by keys
func (mc *MemoryCacheImpl) GetEntries(ctx context.Context, keys []string) (map[string]*CacheEntry, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	entries := make(map[string]*CacheEntry)
	for _, key := range keys {
		item, exists := mc.data[key]
		if !exists {
			continue
		}

		// Check if item has expired
		if !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt) {
			continue
		}

		entries[key] = &CacheEntry{
			Key:       key,
			Value:     item.Value,
			CreatedAt: item.CreatedAt,
			ExpiresAt: item.ExpiresAt,
			TTL:       item.TTL,
			Size:      item.Size,
		}
	}

	return entries, nil
}

// SetEntries sets multiple entries atomically
func (mc *MemoryCacheImpl) SetEntries(ctx context.Context, entries map[string]*CacheEntry) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Check if cache will be full and evict if necessary
	totalEntries := len(mc.data) + len(entries)
	for totalEntries > int(mc.config.MaxSize) {
		mc.evictLRU()
		totalEntries--
	}

	for key, entry := range entries {
		item := &cacheItem{
			Value:       entry.Value,
			CreatedAt:   entry.CreatedAt,
			ExpiresAt:   entry.ExpiresAt,
			TTL:         entry.TTL,
			Size:        entry.Size,
			AccessCount: 0,
			LastAccess:  time.Now(),
		}

		mc.data[key] = item
	}

	mc.stats.Size = int64(len(mc.data))
	return nil
}

// DeleteEntries removes multiple entries by keys
func (mc *MemoryCacheImpl) DeleteEntries(ctx context.Context, keys []string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	for _, key := range keys {
		delete(mc.data, key)
	}

	mc.stats.Size = int64(len(mc.data))
	return nil
}

// GetSize returns the current size of the cache
func (mc *MemoryCacheImpl) GetSize() int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return int64(len(mc.data))
}

// GetMemoryUsage returns the approximate memory usage in bytes
func (mc *MemoryCacheImpl) GetMemoryUsage() int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var totalSize int64
	for _, item := range mc.data {
		totalSize += item.Size
		// Add overhead for key and metadata
		totalSize += 64 // Approximate overhead per entry
	}

	return totalSize
}

// GetExpiredCount returns the number of expired items
func (mc *MemoryCacheImpl) GetExpiredCount() int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.stats.ExpiredCount
}

// GetEvictionCount returns the number of evicted items
func (mc *MemoryCacheImpl) GetEvictionCount() int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.stats.EvictionCount
}

// ResetStats resets the cache statistics
func (mc *MemoryCacheImpl) ResetStats() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.stats = &CacheStats{
		Size: int64(len(mc.data)),
	}
}

// GetConfig returns the cache configuration
func (mc *MemoryCacheImpl) GetConfig() *CacheConfig {
	return mc.config
}

// String returns a string representation of the cache
func (mc *MemoryCacheImpl) String() string {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return fmt.Sprintf("MemoryCache{size=%d, maxSize=%d, hitRate=%.2f%%}",
		len(mc.data), mc.config.MaxSize, mc.stats.HitRate*100)
}
