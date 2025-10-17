package cache

import (
	"container/list"
	"sync"
	"time"
)

// L1MemoryCache implements an in-memory LRU cache with TTL support
type L1MemoryCache struct {
	capacity  int
	ttl       time.Duration
	items     map[string]*list.Element
	evictList *list.List
	mu        sync.RWMutex
	stats     *MemoryCacheStats
	onEvict   func(string, interface{})
}

// MemoryCacheItem represents an item in the memory cache
type MemoryCacheItem struct {
	key        string
	value      interface{}
	expiresAt  time.Time
	accessTime time.Time
}

// MemoryCacheStats represents statistics for the memory cache
type MemoryCacheStats struct {
	Hits       int64     `json:"hits"`
	Misses     int64     `json:"misses"`
	Sets       int64     `json:"sets"`
	Deletes    int64     `json:"deletes"`
	Evictions  int64     `json:"evictions"`
	Size       int       `json:"size"`
	Capacity   int       `json:"capacity"`
	HitRate    float64   `json:"hit_rate"`
	LastAccess time.Time `json:"last_access"`
}

// NewL1MemoryCache creates a new L1 memory cache
func NewL1MemoryCache(capacity int, ttl time.Duration) *L1MemoryCache {
	return &L1MemoryCache{
		capacity:  capacity,
		ttl:       ttl,
		items:     make(map[string]*list.Element),
		evictList: list.New(),
		stats:     &MemoryCacheStats{Capacity: capacity},
	}
}

// SetOnEvict sets the eviction callback function
func (c *L1MemoryCache) SetOnEvict(fn func(string, interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onEvict = fn
}

// Get retrieves a value from the cache
func (c *L1MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if item exists
	element, exists := c.items[key]
	if !exists {
		c.stats.Misses++
		return nil, false
	}

	item := element.Value.(*MemoryCacheItem)

	// Check if item has expired
	if time.Now().After(item.expiresAt) {
		c.removeElement(element)
		c.stats.Misses++
		return nil, false
	}

	// Update access time and move to front
	item.accessTime = time.Now()
	c.evictList.MoveToFront(element)
	c.stats.Hits++
	c.stats.LastAccess = time.Now()

	return item.value, true
}

// Set stores a value in the cache
func (c *L1MemoryCache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.ttl)
}

// SetWithTTL stores a value in the cache with a specific TTL
func (c *L1MemoryCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	expiresAt := now.Add(ttl)

	// Check if item already exists
	if element, exists := c.items[key]; exists {
		// Update existing item
		item := element.Value.(*MemoryCacheItem)
		item.value = value
		item.expiresAt = expiresAt
		item.accessTime = now
		c.evictList.MoveToFront(element)
		c.stats.Sets++
		return
	}

	// Create new item
	item := &MemoryCacheItem{
		key:        key,
		value:      value,
		expiresAt:  expiresAt,
		accessTime: now,
	}

	// Add to front of eviction list
	element := c.evictList.PushFront(item)
	c.items[key] = element
	c.stats.Sets++
	c.stats.Size = len(c.items)

	// Evict if over capacity
	if c.evictList.Len() > c.capacity {
		c.evict()
	}
}

// Delete removes a value from the cache
func (c *L1MemoryCache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, exists := c.items[key]
	if !exists {
		return false
	}

	c.removeElement(element)
	c.stats.Deletes++
	c.stats.Size = len(c.items)

	return true
}

// Clear removes all items from the cache
func (c *L1MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Call eviction callback for all items
	if c.onEvict != nil {
		for _, element := range c.items {
			item := element.Value.(*MemoryCacheItem)
			c.onEvict(item.key, item.value)
		}
	}

	c.items = make(map[string]*list.Element)
	c.evictList = list.New()
	c.stats.Size = 0
}

// Size returns the current number of items in the cache
func (c *L1MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Capacity returns the maximum capacity of the cache
func (c *L1MemoryCache) Capacity() int {
	return c.capacity
}

// Stats returns cache statistics
func (c *L1MemoryCache) Stats() *MemoryCacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := *c.stats
	stats.Size = len(c.items)

	// Calculate hit rate
	total := stats.Hits + stats.Misses
	if total > 0 {
		stats.HitRate = float64(stats.Hits) / float64(total)
	}

	return &stats
}

// CleanupExpired removes expired items from the cache
func (c *L1MemoryCache) CleanupExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	removed := 0

	// Iterate through items and remove expired ones
	for _, element := range c.items {
		item := element.Value.(*MemoryCacheItem)
		if now.After(item.expiresAt) {
			c.removeElement(element)
			removed++
		}
	}

	c.stats.Size = len(c.items)
	return removed
}

// GetKeys returns all keys in the cache
func (c *L1MemoryCache) GetKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}

	return keys
}

// GetOldest returns the oldest item in the cache
func (c *L1MemoryCache) GetOldest() (string, interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.evictList.Len() == 0 {
		return "", nil, false
	}

	element := c.evictList.Back()
	item := element.Value.(*MemoryCacheItem)

	return item.key, item.value, true
}

// GetNewest returns the newest item in the cache
func (c *L1MemoryCache) GetNewest() (string, interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.evictList.Len() == 0 {
		return "", nil, false
	}

	element := c.evictList.Front()
	item := element.Value.(*MemoryCacheItem)

	return item.key, item.value, true
}

// Helper methods

func (c *L1MemoryCache) removeElement(element *list.Element) {
	item := element.Value.(*MemoryCacheItem)

	// Call eviction callback
	if c.onEvict != nil {
		c.onEvict(item.key, item.value)
	}

	// Remove from map and list
	delete(c.items, item.key)
	c.evictList.Remove(element)
	c.stats.Evictions++
}

func (c *L1MemoryCache) evict() {
	// Remove the least recently used item
	element := c.evictList.Back()
	if element != nil {
		c.removeElement(element)
	}
}

// StartCleanupRoutine starts a background routine to clean up expired items
func (c *L1MemoryCache) StartCleanupRoutine(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			removed := c.CleanupExpired()
			if removed > 0 {
				// Log cleanup activity if needed
			}
		}
	}()
}

// MemoryCacheConfig represents configuration for the memory cache
type MemoryCacheConfig struct {
	Capacity        int           `json:"capacity"`
	DefaultTTL      time.Duration `json:"default_ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
	EnableStats     bool          `json:"enable_stats"`
}

// NewL1MemoryCacheWithConfig creates a new L1 memory cache with configuration
func NewL1MemoryCacheWithConfig(config *MemoryCacheConfig) *L1MemoryCache {
	cache := NewL1MemoryCache(config.Capacity, config.DefaultTTL)

	if config.CleanupInterval > 0 {
		cache.StartCleanupRoutine(config.CleanupInterval)
	}

	return cache
}
