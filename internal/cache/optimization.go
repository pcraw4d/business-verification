package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// OptimizationCacheManager handles multi-level caching optimization
// (renamed to avoid conflict with intelligent_cache.go CacheManager)
type OptimizationCacheManager struct {
	l1Cache *sync.Map     // In-memory L1 cache
	l2Cache *redis.Client // Redis L2 cache
	config  *OptimizationCacheConfig
	stats   *OptimizationCacheStats
	mu      sync.RWMutex
}

// OptimizationCacheConfig contains cache optimization settings
// (renamed to avoid conflict with types.go CacheConfig)
type OptimizationCacheConfig struct {
	L1TTL           time.Duration `yaml:"l1_ttl"`
	L2TTL           time.Duration `yaml:"l2_ttl"`
	MaxL1Size       int           `yaml:"max_l1_size"`
	Strategy        string        `yaml:"strategy"` // "write-through", "write-behind", "write-around"
	EnableWarming   bool          `yaml:"enable_warming"`
	WarmingInterval time.Duration `yaml:"warming_interval"`
	Compression     bool          `yaml:"compression"`
}

// OptimizationCacheStats tracks cache performance metrics
// (renamed to avoid conflict with memory.go CacheStats)
type OptimizationCacheStats struct {
	L1Hits        int64 `json:"l1_hits"`
	L1Misses      int64 `json:"l1_misses"`
	L2Hits        int64 `json:"l2_hits"`
	L2Misses      int64 `json:"l2_misses"`
	L1Size        int   `json:"l1_size"`
	L2Size        int64 `json:"l2_size"`
	TotalRequests int64 `json:"total_requests"`
	mu            sync.RWMutex
}

// OptimizationCacheEntry represents a cached item
// (renamed to avoid conflict with intelligent_cache.go CacheEntry)
type OptimizationCacheEntry struct {
	Data        interface{} `json:"data"`
	ExpiresAt   time.Time   `json:"expires_at"`
	CreatedAt   time.Time   `json:"created_at"`
	AccessCount int64       `json:"access_count"`
}

// NewOptimizationCacheManager creates a new optimized cache manager
func NewOptimizationCacheManager(redisClient *redis.Client, config *OptimizationCacheConfig) *OptimizationCacheManager {
	cm := &OptimizationCacheManager{
		l1Cache: &sync.Map{},
		l2Cache: redisClient,
		config:  config,
		stats:   &OptimizationCacheStats{},
	}

	// Start cache warming if enabled
	if config.EnableWarming {
		go cm.startCacheWarming()
	}

	// Start cache cleanup routine
	go cm.startCacheCleanup()

	return cm
}

// Get retrieves a value from cache with multi-level fallback
func (cm *OptimizationCacheManager) Get(ctx context.Context, key string) (interface{}, error) {
	cm.stats.mu.Lock()
	cm.stats.TotalRequests++
	cm.stats.mu.Unlock()

	// Try L1 cache first
	if value, found := cm.getFromL1(key); found {
		cm.stats.mu.Lock()
		cm.stats.L1Hits++
		cm.stats.mu.Unlock()
		return value, nil
	}

	cm.stats.mu.Lock()
	cm.stats.L1Misses++
	cm.stats.mu.Unlock()

	// Try L2 cache (Redis)
	value, err := cm.getFromL2(ctx, key)
	if err == nil && value != nil {
		cm.stats.mu.Lock()
		cm.stats.L2Hits++
		cm.stats.mu.Unlock()

		// Store in L1 cache for faster access
		cm.setInL1(key, value, cm.config.L1TTL)
		return value, nil
	}

	cm.stats.mu.Lock()
	cm.stats.L2Misses++
	cm.stats.mu.Unlock()

	return nil, fmt.Errorf("cache miss for key: %s", key)
}

// Set stores a value in cache with write strategy
func (cm *OptimizationCacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	switch cm.config.Strategy {
	case "write-through":
		return cm.writeThrough(ctx, key, value, ttl)
	case "write-behind":
		return cm.writeBehind(ctx, key, value, ttl)
	case "write-around":
		return cm.writeAround(ctx, key, value, ttl)
	default:
		return cm.writeThrough(ctx, key, value, ttl)
	}
}

// writeThrough writes to both L1 and L2 caches immediately
func (cm *OptimizationCacheManager) writeThrough(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Write to L1 cache
	cm.setInL1(key, value, ttl)

	// Write to L2 cache
	return cm.setInL2(ctx, key, value, ttl)
}

// writeBehind writes to L1 immediately and L2 asynchronously
func (cm *OptimizationCacheManager) writeBehind(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Write to L1 cache immediately
	cm.setInL1(key, value, ttl)

	// Write to L2 cache asynchronously
	go func() {
		if err := cm.setInL2(context.Background(), key, value, ttl); err != nil {
			log.Printf("Warning: Failed to write to L2 cache asynchronously: %v", err)
		}
	}()

	return nil
}

// writeAround writes only to L2 cache, bypassing L1
func (cm *OptimizationCacheManager) writeAround(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Remove from L1 cache if exists
	cm.l1Cache.Delete(key)

	// Write to L2 cache
	return cm.setInL2(ctx, key, value, ttl)
}

// Invalidate removes a key from all cache levels
func (cm *OptimizationCacheManager) Invalidate(ctx context.Context, key string) error {
	// Remove from L1 cache
	cm.l1Cache.Delete(key)

	// Remove from L2 cache
	return cm.l2Cache.Del(ctx, key).Err()
}

// InvalidatePattern removes keys matching a pattern from all cache levels
func (cm *OptimizationCacheManager) InvalidatePattern(ctx context.Context, pattern string) error {
	// Get matching keys from L2 cache
	keys, err := cm.l2Cache.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
	}

	// Remove from L2 cache
	if len(keys) > 0 {
		if err := cm.l2Cache.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete keys from L2 cache: %w", err)
		}
	}

	// Remove from L1 cache (iterate through all keys)
	cm.l1Cache.Range(func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			if matched, _ := matchOptimizationPattern(keyStr, pattern); matched {
				cm.l1Cache.Delete(key)
			}
		}
		return true
	})

	return nil
}

// GetStats returns current cache statistics
func (cm *OptimizationCacheManager) GetStats() *OptimizationCacheStats {
	cm.stats.mu.RLock()
	defer cm.stats.mu.RUnlock()

	// Update L1 size
	cm.stats.L1Size = cm.getL1Size()

	// Update L2 size
	if info, err := cm.l2Cache.Info(context.Background(), "memory").Result(); err == nil {
		// Parse memory info to get approximate size
		// This is a simplified approach - in production, you might want more sophisticated size tracking
		cm.stats.L2Size = int64(len(info))
	}

	return &OptimizationCacheStats{
		L1Hits:        cm.stats.L1Hits,
		L1Misses:      cm.stats.L1Misses,
		L2Hits:        cm.stats.L2Hits,
		L2Misses:      cm.stats.L2Misses,
		L1Size:        cm.stats.L1Size,
		L2Size:        cm.stats.L2Size,
		TotalRequests: cm.stats.TotalRequests,
	}
}

// GetHitRate returns the overall cache hit rate
func (cm *OptimizationCacheManager) GetHitRate() float64 {
	cm.stats.mu.RLock()
	defer cm.stats.mu.RUnlock()

	if cm.stats.TotalRequests == 0 {
		return 0.0
	}

	totalHits := cm.stats.L1Hits + cm.stats.L2Hits
	return float64(totalHits) / float64(cm.stats.TotalRequests) * 100.0
}

// L1 cache operations
func (cm *OptimizationCacheManager) getFromL1(key string) (interface{}, bool) {
	if value, ok := cm.l1Cache.Load(key); ok {
		entry, ok := value.(*OptimizationCacheEntry)
		if !ok {
			return nil, false
		}

		// Check if expired
		if time.Now().After(entry.ExpiresAt) {
			cm.l1Cache.Delete(key)
			return nil, false
		}

		// Update access count
		entry.AccessCount++
		return entry.Data, true
	}

	return nil, false
}

func (cm *OptimizationCacheManager) setInL1(key string, value interface{}, ttl time.Duration) {
	entry := &OptimizationCacheEntry{
		Data:        value,
		ExpiresAt:   time.Now().Add(ttl),
		CreatedAt:   time.Now(),
		AccessCount: 1,
	}

	cm.l1Cache.Store(key, entry)

	// Check if we need to evict old entries
	if cm.getL1Size() > cm.config.MaxL1Size {
		cm.evictL1LRU()
	}
}

func (cm *OptimizationCacheManager) getL1Size() int {
	size := 0
	cm.l1Cache.Range(func(key, value interface{}) bool {
		size++
		return true
	})
	return size
}

func (cm *OptimizationCacheManager) evictL1LRU() {
	// Simple LRU eviction - remove oldest entries
	// In production, you might want a more sophisticated eviction strategy
	var oldestKey interface{}
	var oldestTime time.Time

	cm.l1Cache.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*OptimizationCacheEntry); ok {
			if oldestTime.IsZero() || entry.CreatedAt.Before(oldestTime) {
				oldestTime = entry.CreatedAt
				oldestKey = key
			}
		}
		return true
	})

	if oldestKey != nil {
		cm.l1Cache.Delete(oldestKey)
	}
}

// L2 cache operations
func (cm *OptimizationCacheManager) getFromL2(ctx context.Context, key string) (interface{}, error) {
	data, err := cm.l2Cache.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var value interface{}
	if err := json.Unmarshal([]byte(data), &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	return value, nil
}

func (cm *OptimizationCacheManager) setInL2(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal data for caching: %w", err)
	}

	return cm.l2Cache.Set(ctx, key, data, ttl).Err()
}

// Cache warming
func (cm *OptimizationCacheManager) startCacheWarming() {
	ticker := time.NewTicker(cm.config.WarmingInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := cm.warmCache(); err != nil {
			log.Printf("Warning: Cache warming failed: %v", err)
		}
	}
}

func (cm *OptimizationCacheManager) warmCache() error {
	log.Println("Starting cache warming...")

	// Get frequently accessed business IDs
	businessIDs, err := cm.getFrequentlyAccessedBusinesses()
	if err != nil {
		return fmt.Errorf("failed to get frequently accessed businesses: %w", err)
	}

	// Pre-load business data
	for _, businessID := range businessIDs {
		go cm.preloadBusinessData(businessID)
	}

	log.Printf("Cache warming initiated for %d businesses", len(businessIDs))
	return nil
}

func (cm *OptimizationCacheManager) getFrequentlyAccessedBusinesses() ([]string, error) {
	// This would typically query your database for frequently accessed businesses
	// For now, return a sample list
	return []string{
		"business-001",
		"business-002",
		"business-003",
	}, nil
}

func (cm *OptimizationCacheManager) preloadBusinessData(businessID string) {
	ctx := context.Background()

	// Check if already cached
	if _, err := cm.Get(ctx, fmt.Sprintf("business:verification:%s", businessID)); err == nil {
		return // Already cached
	}

	// This would typically fetch from your database
	// For now, we'll simulate the data
	businessData := map[string]interface{}{
		"id":     businessID,
		"name":   fmt.Sprintf("Business %s", businessID),
		"status": "verified",
	}

	// Cache the data
	if err := cm.Set(ctx, fmt.Sprintf("business:verification:%s", businessID), businessData, cm.config.L2TTL); err != nil {
		log.Printf("Warning: Failed to preload business data for %s: %v", businessID, err)
	}
}

// Cache cleanup
func (cm *OptimizationCacheManager) startCacheCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cm.cleanupExpiredL1Entries()
	}
}

func (cm *OptimizationCacheManager) cleanupExpiredL1Entries() {
	now := time.Now()

	cm.l1Cache.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*OptimizationCacheEntry); ok {
			if now.After(entry.ExpiresAt) {
				cm.l1Cache.Delete(key)
			}
		}
		return true
	})
}

// Utility functions
// matchOptimizationPattern checks if a string matches a pattern (for optimization cache)
// (renamed to avoid conflict with query_cache_manager.go matchPattern)
func matchOptimizationPattern(str, pattern string) (bool, error) {
	// Simple pattern matching - in production, you might want more sophisticated matching
	// This is a basic implementation for demonstration
	if pattern == "*" {
		return true, nil
	}

	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(str) >= len(prefix) && str[:len(prefix)] == prefix, nil
	}

	return str == pattern, nil
}

// CacheOptimizer provides cache optimization utilities
type CacheOptimizer struct {
	cacheManager *OptimizationCacheManager
}

// NewCacheOptimizer creates a new cache optimizer
func NewCacheOptimizer(cacheManager *OptimizationCacheManager) *CacheOptimizer {
	return &CacheOptimizer{
		cacheManager: cacheManager,
	}
}

// OptimizeCacheConfiguration optimizes cache settings based on usage patterns
func (co *CacheOptimizer) OptimizeCacheConfiguration() (*OptimizationCacheConfig, error) {
	stats := co.cacheManager.GetStats()

	// Analyze hit rates and adjust configuration
	config := &OptimizationCacheConfig{
		L1TTL:           co.cacheManager.config.L1TTL,
		L2TTL:           co.cacheManager.config.L2TTL,
		MaxL1Size:       co.cacheManager.config.MaxL1Size,
		Strategy:        co.cacheManager.config.Strategy,
		EnableWarming:   co.cacheManager.config.EnableWarming,
		WarmingInterval: co.cacheManager.config.WarmingInterval,
		Compression:     co.cacheManager.config.Compression,
	}

	// Adjust L1 TTL based on hit rate
	hitRate := co.cacheManager.GetHitRate()
	if hitRate < 70 {
		// Low hit rate - increase TTL
		config.L1TTL = config.L1TTL * 2
		config.L2TTL = config.L2TTL * 2
	} else if hitRate > 90 {
		// High hit rate - can reduce TTL to save memory
		config.L1TTL = config.L1TTL / 2
	}

	// Adjust L1 size based on usage
	if stats.L1Size > int(float64(config.MaxL1Size)*0.9) {
		config.MaxL1Size = config.MaxL1Size * 2
	}

	return config, nil
}

// GenerateCacheReport generates a comprehensive cache performance report
func (co *CacheOptimizer) GenerateCacheReport() (*CacheReport, error) {
	stats := co.cacheManager.GetStats()

	report := &CacheReport{
		Timestamp:       time.Now(),
		Stats:           stats,
		HitRate:         co.cacheManager.GetHitRate(),
		Recommendations: co.generateRecommendations(stats),
	}

	return report, nil
}

// CacheReport contains cache performance analysis
type CacheReport struct {
	Timestamp       time.Time               `json:"timestamp"`
	Stats           *OptimizationCacheStats `json:"stats"`
	HitRate         float64                 `json:"hit_rate"`
	Recommendations []string                `json:"recommendations"`
}

func (co *CacheOptimizer) generateRecommendations(stats *OptimizationCacheStats) []string {
	var recommendations []string

	hitRate := co.cacheManager.GetHitRate()

	if hitRate < 70 {
		recommendations = append(recommendations, "Cache hit rate is low - consider increasing TTL or improving cache warming")
	}

	if stats.L1Size > int(float64(co.cacheManager.config.MaxL1Size)*0.8) {
		recommendations = append(recommendations, "L1 cache is nearly full - consider increasing max size or improving eviction strategy")
	}

	if stats.L1Misses > stats.L2Hits {
		recommendations = append(recommendations, "High L1 miss rate - consider adjusting L1 cache strategy")
	}

	return recommendations
}
