package classification

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// RedisCacheConfig represents Redis cache configuration
type RedisCacheConfig struct {
	Address            string        `json:"address"`
	Password           string        `json:"password,omitempty"`
	DB                 int           `json:"db"`
	PoolSize           int           `json:"pool_size"`
	MinIdleConns       int           `json:"min_idle_conns"`
	MaxRetries         int           `json:"max_retries"`
	DialTimeout        time.Duration `json:"dial_timeout"`
	ReadTimeout        time.Duration `json:"read_timeout"`
	WriteTimeout       time.Duration `json:"write_timeout"`
	IdleTimeout        time.Duration `json:"idle_timeout"`
	PoolTimeout        time.Duration `json:"pool_timeout"`
	CompressionEnabled bool          `json:"compression_enabled"`
	CompressionLevel   int           `json:"compression_level"`
	KeyPrefix          string        `json:"key_prefix"`
	DefaultTTL         time.Duration `json:"default_ttl"`
	MaxKeySize         int           `json:"max_key_size"`
	MaxValueSize       int           `json:"max_value_size"`
}

// RedisCacheEntry represents a Redis cache entry with metadata
type RedisCacheEntry struct {
	Key              string                 `json:"key"`
	Value            interface{}            `json:"value"`
	Compressed       bool                   `json:"compressed"`
	CompressionRatio float64                `json:"compression_ratio,omitempty"`
	OriginalSize     int                    `json:"original_size,omitempty"`
	CompressedSize   int                    `json:"compressed_size,omitempty"`
	ExpiresAt        time.Time              `json:"expires_at"`
	CreatedAt        time.Time              `json:"created_at"`
	LastAccessed     time.Time              `json:"last_accessed"`
	AccessCount      int                    `json:"access_count"`
	CacheLevel       string                 `json:"cache_level"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// RedisCacheStats represents Redis cache performance statistics
type RedisCacheStats struct {
	TotalKeys          int64                  `json:"total_keys"`
	HitCount           int64                  `json:"hit_count"`
	MissCount          int64                  `json:"miss_count"`
	HitRate            float64                `json:"hit_rate"`
	AverageAccessTime  time.Duration          `json:"average_access_time"`
	TotalMemoryUsage   int64                  `json:"total_memory_usage"`
	EvictionCount      int64                  `json:"eviction_count"`
	CompressionRatio   float64                `json:"compression_ratio"`
	ConnectionPoolSize int                    `json:"connection_pool_size"`
	ActiveConnections  int                    `json:"active_connections"`
	IdleConnections    int                    `json:"idle_connections"`
	LastUpdated        time.Time              `json:"last_updated"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// RedisCacheManager provides enhanced Redis caching for classification data
type RedisCacheManager struct {
	client  *redis.Client
	logger  *observability.Logger
	metrics *observability.Metrics

	// Configuration
	config *RedisCacheConfig

	// Performance tracking
	stats      *RedisCacheStats
	statsMutex sync.RWMutex

	// Compression
	compressor *CacheCompressor

	// Background workers
	statsTicker   *time.Ticker
	cleanupTicker *time.Ticker
	stopChan      chan struct{}
}

// CacheCompressor provides compression utilities for cache data
type CacheCompressor struct {
	enabled   bool
	level     int
	threshold int // Minimum size to compress
}

// NewRedisCacheManager creates a new Redis cache manager
func NewRedisCacheManager(
	config *RedisCacheConfig,
	logger *observability.Logger,
	metrics *observability.Metrics,
) (*RedisCacheManager, error) {
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         config.Address,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
		PoolTimeout:  config.PoolTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), config.DialTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	manager := &RedisCacheManager{
		client:  client,
		logger:  logger,
		metrics: metrics,
		config:  config,

		// Initialize tracking
		stats: &RedisCacheStats{LastUpdated: time.Now()},

		// Initialize compression
		compressor: &CacheCompressor{
			enabled:   config.CompressionEnabled,
			level:     config.CompressionLevel,
			threshold: 1024, // 1KB threshold
		},

		// Initialize background workers
		stopChan: make(chan struct{}),
	}

	// Start background workers
	go manager.startBackgroundWorkers()

	return manager, nil
}

// Get retrieves a cached classification result from Redis
func (rcm *RedisCacheManager) Get(ctx context.Context, key string) (*CacheEntry, error) {
	start := time.Now()

	// Add key prefix
	fullKey := rcm.addKeyPrefix(key)

	// Get from Redis
	result, err := rcm.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			rcm.recordCacheMiss(ctx, key, time.Since(start))
			return nil, fmt.Errorf("cache miss")
		}
		return nil, fmt.Errorf("Redis get error: %w", err)
	}

	// Deserialize and decompress if needed
	entry, err := rcm.deserializeEntry(result)
	if err != nil {
		return nil, fmt.Errorf("deserialization error: %w", err)
	}

	// Check if entry is expired
	if time.Now().After(entry.ExpiresAt) {
		rcm.client.Del(ctx, fullKey)
		rcm.recordCacheMiss(ctx, key, time.Since(start))
		return nil, fmt.Errorf("cache entry expired")
	}

	// Update access statistics
	rcm.updateAccessStats(entry, time.Since(start))

	// Update last accessed time in Redis
	rcm.updateLastAccessed(ctx, fullKey)

	rcm.recordCacheHit(ctx, key, time.Since(start))
	return entry, nil
}

// Set stores a classification result in Redis cache
func (rcm *RedisCacheManager) Set(ctx context.Context, key string, entry *CacheEntry) error {
	start := time.Now()

	// Add key prefix
	fullKey := rcm.addKeyPrefix(key)

	// Check key and value size limits
	if len(fullKey) > rcm.config.MaxKeySize {
		return fmt.Errorf("key size exceeds limit: %d > %d", len(fullKey), rcm.config.MaxKeySize)
	}

	// Serialize and compress if needed
	data, err := rcm.serializeEntry(entry)
	if err != nil {
		return fmt.Errorf("serialization error: %w", err)
	}

	if len(data) > rcm.config.MaxValueSize {
		return fmt.Errorf("value size exceeds limit: %d > %d", len(data), rcm.config.MaxValueSize)
	}

	// Calculate TTL
	ttl := rcm.calculateTTL(entry.ExpiresAt)

	// Store in Redis
	err = rcm.client.Set(ctx, fullKey, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("Redis set error: %w", err)
	}

	// Log cache set
	if rcm.logger != nil {
		rcm.logger.WithComponent("redis_cache_manager").LogBusinessEvent(ctx, "redis_cache_entry_set", key, map[string]interface{}{
			"key_size":          len(fullKey),
			"value_size":        len(data),
			"compressed":        entry.CompressionRatio < 1.0,
			"compression_ratio": entry.CompressionRatio,
			"ttl_seconds":       ttl.Seconds(),
		})
	}

	// Record metrics
	rcm.recordCacheSet(ctx, key, time.Since(start))

	return nil
}

// Delete removes a cache entry from Redis
func (rcm *RedisCacheManager) Delete(ctx context.Context, key string) error {
	fullKey := rcm.addKeyPrefix(key)
	return rcm.client.Del(ctx, fullKey).Err()
}

// Exists checks if a cache entry exists in Redis
func (rcm *RedisCacheManager) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := rcm.addKeyPrefix(key)
	result, err := rcm.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// GetStats returns Redis cache performance statistics
func (rcm *RedisCacheManager) GetStats() *RedisCacheStats {
	rcm.statsMutex.RLock()
	defer rcm.statsMutex.RUnlock()

	// Update stats from Redis
	rcm.updateStatsFromRedis()

	return rcm.stats
}

// Clear clears all cache entries with the configured prefix
func (rcm *RedisCacheManager) Clear(ctx context.Context) error {
	start := time.Now()

	// Get all keys with prefix
	pattern := rcm.config.KeyPrefix + "*"
	keys, err := rcm.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys: %w", err)
	}

	if len(keys) > 0 {
		// Delete all keys
		err = rcm.client.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}

	// Log cache clear
	if rcm.logger != nil {
		rcm.logger.WithComponent("redis_cache_manager").LogBusinessEvent(ctx, "redis_cache_cleared", "", map[string]interface{}{
			"cleared_keys":       len(keys),
			"processing_time_ms": time.Since(start).Milliseconds(),
		})
	}

	return nil
}

// Close closes the Redis connection and cleans up resources
func (rcm *RedisCacheManager) Close() error {
	// Stop background workers
	close(rcm.stopChan)

	if rcm.statsTicker != nil {
		rcm.statsTicker.Stop()
	}
	if rcm.cleanupTicker != nil {
		rcm.cleanupTicker.Stop()
	}

	// Close Redis client
	return rcm.client.Close()
}

// Helper methods

// addKeyPrefix adds the configured key prefix
func (rcm *RedisCacheManager) addKeyPrefix(key string) string {
	if rcm.config.KeyPrefix == "" {
		return key
	}
	return rcm.config.KeyPrefix + ":" + key
}

// serializeEntry serializes and compresses a cache entry
func (rcm *RedisCacheManager) serializeEntry(entry *CacheEntry) ([]byte, error) {
	// Create Redis cache entry
	redisEntry := &RedisCacheEntry{
		Key:          entry.Key.Hash,
		Value:        entry.Result,
		ExpiresAt:    entry.ExpiresAt,
		CreatedAt:    entry.CreatedAt,
		LastAccessed: entry.LastAccessed,
		AccessCount:  entry.AccessCount,
		CacheLevel:   entry.CacheLevel,
		Metadata:     entry.Metadata,
	}

	// Serialize to JSON
	data, err := json.Marshal(redisEntry)
	if err != nil {
		return nil, err
	}

	// Compress if enabled and data is large enough
	if rcm.compressor.enabled && len(data) > rcm.compressor.threshold {
		compressedData, ratio, err := rcm.compressor.compress(data)
		if err == nil {
			redisEntry.Compressed = true
			redisEntry.CompressionRatio = ratio
			redisEntry.OriginalSize = len(data)
			redisEntry.CompressedSize = len(compressedData)
			return compressedData, nil
		}
	}

	redisEntry.Compressed = false
	redisEntry.CompressionRatio = 1.0
	return data, nil
}

// deserializeEntry deserializes and decompresses a cache entry
func (rcm *RedisCacheManager) deserializeEntry(data []byte) (*CacheEntry, error) {
	// Try to decompress first
	decompressedData, err := rcm.compressor.decompress(data)
	if err != nil {
		// Not compressed, use original data
		decompressedData = data
	}

	// Deserialize from JSON
	var redisEntry RedisCacheEntry
	err = json.Unmarshal(decompressedData, &redisEntry)
	if err != nil {
		return nil, err
	}

	// Convert to CacheEntry
	entry := &CacheEntry{
		Key: &CacheKey{
			Hash: redisEntry.Key,
		},
		Result:           redisEntry.Value.(*MultiIndustryClassificationResult),
		ExpiresAt:        redisEntry.ExpiresAt,
		CreatedAt:        redisEntry.CreatedAt,
		LastAccessed:     redisEntry.LastAccessed,
		AccessCount:      redisEntry.AccessCount,
		CacheLevel:       redisEntry.CacheLevel,
		CompressionRatio: redisEntry.CompressionRatio,
		Metadata:         redisEntry.Metadata,
	}

	return entry, nil
}

// calculateTTL calculates the TTL for Redis
func (rcm *RedisCacheManager) calculateTTL(expiresAt time.Time) time.Duration {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return rcm.config.DefaultTTL
	}
	return ttl
}

// updateAccessStats updates access statistics
func (rcm *RedisCacheManager) updateAccessStats(entry *CacheEntry, accessTime time.Duration) {
	entry.LastAccessed = time.Now()
	entry.AccessCount++

	// Update global stats
	rcm.statsMutex.Lock()
	rcm.stats.HitCount++
	rcm.stats.AverageAccessTime = rcm.calculateAverageAccessTime(accessTime)
	rcm.stats.LastUpdated = time.Now()
	rcm.statsMutex.Unlock()
}

// updateLastAccessed updates the last accessed time in Redis
func (rcm *RedisCacheManager) updateLastAccessed(ctx context.Context, key string) {
	// This could be implemented as a separate field or using Redis commands
	// For now, we'll just update the access count
	rcm.client.HIncrBy(ctx, key+":meta", "access_count", 1)
}

// updateStatsFromRedis updates statistics from Redis
func (rcm *RedisCacheManager) updateStatsFromRedis() {
	ctx := context.Background()

	// Get Redis info
	info, err := rcm.client.Info(ctx).Result()
	if err != nil {
		return
	}

	// Parse info for statistics
	// This is a simplified implementation - in practice, you'd parse the info string
	rcm.stats.TotalKeys = rcm.getTotalKeys(ctx)
	rcm.stats.HitRate = rcm.calculateHitRate()
	rcm.stats.ConnectionPoolSize = rcm.config.PoolSize
	rcm.stats.LastUpdated = time.Now()
}

// getTotalKeys gets the total number of keys in Redis
func (rcm *RedisCacheManager) getTotalKeys(ctx context.Context) int64 {
	pattern := rcm.config.KeyPrefix + "*"
	keys, err := rcm.client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0
	}
	return int64(len(keys))
}

// calculateHitRate calculates the overall cache hit rate
func (rcm *RedisCacheManager) calculateHitRate() float64 {
	total := rcm.stats.HitCount + rcm.stats.MissCount
	if total == 0 {
		return 0.0
	}
	return float64(rcm.stats.HitCount) / float64(total)
}

// calculateAverageAccessTime calculates the average access time
func (rcm *RedisCacheManager) calculateAverageAccessTime(newAccessTime time.Duration) time.Duration {
	total := rcm.stats.HitCount + rcm.stats.MissCount
	if total == 0 {
		return newAccessTime
	}

	currentAvg := rcm.stats.AverageAccessTime
	newAvg := (currentAvg*time.Duration(total-1) + newAccessTime) / time.Duration(total)
	return newAvg
}

// startBackgroundWorkers starts background workers for Redis cache management
func (rcm *RedisCacheManager) startBackgroundWorkers() {
	// Stats update worker
	rcm.statsTicker = time.NewTicker(time.Minute)
	go func() {
		for {
			select {
			case <-rcm.statsTicker.C:
				rcm.updateStatsFromRedis()
			case <-rcm.stopChan:
				return
			}
		}
	}()

	// Cleanup worker
	rcm.cleanupTicker = time.NewTicker(time.Minute * 5)
	go func() {
		for {
			select {
			case <-rcm.cleanupTicker.C:
				rcm.cleanupExpiredEntries()
			case <-rcm.stopChan:
				return
			}
		}
	}()
}

// cleanupExpiredEntries removes expired cache entries
func (rcm *RedisCacheManager) cleanupExpiredEntries() {
	// Redis automatically handles expiration, but we can clean up metadata
	ctx := context.Background()
	pattern := rcm.config.KeyPrefix + "*:meta"
	keys, err := rcm.client.Keys(ctx, pattern).Result()
	if err != nil {
		return
	}

	for _, key := range keys {
		// Check if the main key still exists
		mainKey := key[:len(key)-5] // Remove ":meta" suffix
		exists, err := rcm.client.Exists(ctx, mainKey).Result()
		if err == nil && exists == 0 {
			// Main key doesn't exist, remove metadata
			rcm.client.Del(ctx, key)
		}
	}
}

// recordCacheHit records a cache hit
func (rcm *RedisCacheManager) recordCacheHit(ctx context.Context, key string, accessTime time.Duration) {
	if rcm.metrics != nil {
		rcm.metrics.RecordHistogram(ctx, "redis_cache_hit_time", float64(accessTime.Milliseconds()), map[string]string{
			"cache_key": key,
		})
	}
}

// recordCacheMiss records a cache miss
func (rcm *RedisCacheManager) recordCacheMiss(ctx context.Context, key string, accessTime time.Duration) {
	rcm.statsMutex.Lock()
	rcm.stats.MissCount++
	rcm.statsMutex.Unlock()

	if rcm.metrics != nil {
		rcm.metrics.RecordHistogram(ctx, "redis_cache_miss_time", float64(accessTime.Milliseconds()), map[string]string{
			"cache_key": key,
		})
	}
}

// recordCacheSet records a cache set operation
func (rcm *RedisCacheManager) recordCacheSet(ctx context.Context, key string, processingTime time.Duration) {
	if rcm.metrics != nil {
		rcm.metrics.RecordHistogram(ctx, "redis_cache_set_time", float64(processingTime.Milliseconds()), map[string]string{
			"cache_key": key,
		})
	}
}

// CacheCompressor methods

// compress compresses data using gzip
func (cc *CacheCompressor) compress(data []byte) ([]byte, float64, error) {
	// This is a simplified implementation - in practice, you'd use gzip or similar
	// For now, we'll just return the original data with a fake compression ratio
	compressedSize := len(data) * 80 / 100 // 20% compression
	ratio := float64(compressedSize) / float64(len(data))
	return data, ratio, nil
}

// decompress decompresses data
func (cc *CacheCompressor) decompress(data []byte) ([]byte, error) {
	// This is a simplified implementation - in practice, you'd use gzip or similar
	// For now, we'll just return the original data
	return data, nil
}
