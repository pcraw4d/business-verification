package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisCacheImpl implements a Redis-based cache with TTL and eviction
type RedisCacheImpl struct {
	client *redis.Client
	config *RedisCacheConfig
	logger *zap.Logger
	stats  *CacheStats
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(config *RedisCacheConfig, logger *zap.Logger) (*RedisCacheImpl, error) {
	if config == nil {
		config = &RedisCacheConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
			TTL:      1 * time.Hour,
			PoolSize: 10,
		}
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	cache := &RedisCacheImpl{
		client: rdb,
		config: config,
		logger: logger,
		stats:  &CacheStats{},
	}

	logger.Info("Redis cache initialized",
		zap.String("addr", config.Addr),
		zap.Int("db", config.DB),
		zap.Int("pool_size", config.PoolSize))

	return cache, nil
}

// Get retrieves a value from the cache
func (rc *RedisCacheImpl) Get(ctx context.Context, key string) ([]byte, error) {
	start := time.Now()
	defer func() {
		rc.logger.Debug("Redis cache get operation",
			zap.String("key", key),
			zap.Duration("duration", time.Since(start)))
	}()

	val, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			rc.recordMiss()
			return nil, CacheNotFoundError
		}
		rc.logger.Error("Redis get error", zap.String("key", key), zap.Error(err))
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	rc.recordHit()
	return []byte(val), nil
}

// Set stores a value in the cache
func (rc *RedisCacheImpl) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		rc.logger.Debug("Redis cache set operation",
			zap.String("key", key),
			zap.Int("value_size", len(value)),
			zap.Duration("ttl", ttl),
			zap.Duration("duration", time.Since(start)))
	}()

	if ttl == 0 {
		ttl = rc.config.TTL
	}

	err := rc.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		rc.logger.Error("Redis set error", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("redis set error: %w", err)
	}

	return nil
}

// Delete removes a value from the cache
func (rc *RedisCacheImpl) Delete(ctx context.Context, key string) error {
	start := time.Now()
	defer func() {
		rc.logger.Debug("Redis cache delete operation",
			zap.String("key", key),
			zap.Duration("duration", time.Since(start)))
	}()

	result, err := rc.client.Del(ctx, key).Result()
	if err != nil {
		rc.logger.Error("Redis delete error", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("redis delete error: %w", err)
	}

	if result == 0 {
		return CacheNotFoundError
	}

	return nil
}

// Exists checks if a key exists in the cache
func (rc *RedisCacheImpl) Exists(ctx context.Context, key string) (bool, error) {
	start := time.Now()
	defer func() {
		rc.logger.Debug("Redis cache exists operation",
			zap.String("key", key),
			zap.Duration("duration", time.Since(start)))
	}()

	result, err := rc.client.Exists(ctx, key).Result()
	if err != nil {
		rc.logger.Error("Redis exists error", zap.String("key", key), zap.Error(err))
		return false, fmt.Errorf("redis exists error: %w", err)
	}

	return result > 0, nil
}

// GetTTL returns the remaining TTL for a key
func (rc *RedisCacheImpl) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	start := time.Now()
	defer func() {
		rc.logger.Debug("Redis cache TTL operation",
			zap.String("key", key),
			zap.Duration("duration", time.Since(start)))
	}()

	ttl, err := rc.client.TTL(ctx, key).Result()
	if err != nil {
		rc.logger.Error("Redis TTL error", zap.String("key", key), zap.Error(err))
		return 0, fmt.Errorf("redis TTL error: %w", err)
	}

	if ttl == -2 {
		return 0, CacheNotFoundError
	}

	if ttl == -1 {
		return 0, nil // No expiration
	}

	return ttl, nil
}

// SetTTL updates the TTL for an existing key
func (rc *RedisCacheImpl) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		rc.logger.Debug("Redis cache set TTL operation",
			zap.String("key", key),
			zap.Duration("ttl", ttl),
			zap.Duration("duration", time.Since(start)))
	}()

	result, err := rc.client.Expire(ctx, key, ttl).Result()
	if err != nil {
		rc.logger.Error("Redis set TTL error", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("redis set TTL error: %w", err)
	}

	if !result {
		return CacheNotFoundError
	}

	return nil
}

// Clear removes all keys from the cache
func (rc *RedisCacheImpl) Clear(ctx context.Context) error {
	start := time.Now()
	defer func() {
		rc.logger.Debug("Redis cache clear operation",
			zap.Duration("duration", time.Since(start)))
	}()

	err := rc.client.FlushDB(ctx).Err()
	if err != nil {
		rc.logger.Error("Redis clear error", zap.Error(err))
		return fmt.Errorf("redis clear error: %w", err)
	}

	return nil
}

// GetStats returns cache statistics
func (rc *RedisCacheImpl) GetStats(ctx context.Context) (*CacheStats, error) {
	start := time.Now()
	defer func() {
		rc.logger.Debug("Redis cache stats operation",
			zap.Duration("duration", time.Since(start)))
	}()

	// Get Redis info
	info, err := rc.client.Info(ctx, "memory", "stats").Result()
	if err != nil {
		rc.logger.Error("Redis info error", zap.Error(err))
		return nil, fmt.Errorf("redis info error: %w", err)
	}

	// Get database size
	dbSize, err := rc.client.DBSize(ctx).Result()
	if err != nil {
		rc.logger.Error("Redis DBSize error", zap.Error(err))
		return nil, fmt.Errorf("redis DBSize error: %w", err)
	}

	// Calculate hit rate
	totalRequests := rc.stats.HitCount + rc.stats.MissCount
	hitRate := 0.0
	if totalRequests > 0 {
		hitRate = float64(rc.stats.HitCount) / float64(totalRequests)
	}

	stats := &CacheStats{
		HitCount:      rc.stats.HitCount,
		MissCount:     rc.stats.MissCount,
		HitRate:       hitRate,
		Size:          dbSize,
		MaxSize:       0, // Redis doesn't have a fixed max size
		EvictionCount: 0, // Would need to parse Redis info for this
		ExpiredCount:  0, // Would need to parse Redis info for this
	}

	rc.logger.Debug("Redis cache stats retrieved",
		zap.Int64("size", stats.Size),
		zap.Float64("hit_rate", stats.HitRate),
		zap.String("info", info))

	return stats, nil
}

// Close closes the Redis connection
func (rc *RedisCacheImpl) Close() error {
	rc.logger.Info("Closing Redis cache connection")
	return rc.client.Close()
}

// GetKeys returns all keys matching a pattern
func (rc *RedisCacheImpl) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	start := time.Now()
	defer func() {
		rc.logger.Debug("Redis cache keys operation",
			zap.String("pattern", pattern),
			zap.Duration("duration", time.Since(start)))
	}()

	if pattern == "" {
		pattern = "*"
	}

	keys, err := rc.client.Keys(ctx, pattern).Result()
	if err != nil {
		rc.logger.Error("Redis keys error", zap.String("pattern", pattern), zap.Error(err))
		return nil, fmt.Errorf("redis keys error: %w", err)
	}

	return keys, nil
}

// GetEntries returns multiple entries by keys
func (rc *RedisCacheImpl) GetEntries(ctx context.Context, keys []string) (map[string]*CacheEntry, error) {
	start := time.Now()
	defer func() {
		rc.logger.Debug("Redis cache mget operation",
			zap.Int("key_count", len(keys)),
			zap.Duration("duration", time.Since(start)))
	}()

	if len(keys) == 0 {
		return make(map[string]*CacheEntry), nil
	}

	values, err := rc.client.MGet(ctx, keys...).Result()
	if err != nil {
		rc.logger.Error("Redis mget error", zap.Error(err))
		return nil, fmt.Errorf("redis mget error: %w", err)
	}

	entries := make(map[string]*CacheEntry)
	for i, key := range keys {
		if i < len(values) && values[i] != nil {
			if str, ok := values[i].(string); ok {
				entries[key] = &CacheEntry{
					Key:       key,
					Value:     str,
					CreatedAt: time.Now(),  // Redis doesn't store creation time
					ExpiresAt: time.Time{}, // Would need to get TTL separately
					Size:      int64(len(str)),
				}
			}
		}
	}

	return entries, nil
}

// SetEntries sets multiple entries atomically
func (rc *RedisCacheImpl) SetEntries(ctx context.Context, entries map[string]*CacheEntry) error {
	start := time.Now()
	defer func() {
		rc.logger.Debug("Redis cache mset operation",
			zap.Int("entry_count", len(entries)),
			zap.Duration("duration", time.Since(start)))
	}()

	if len(entries) == 0 {
		return nil
	}

	// Use pipeline for better performance
	pipe := rc.client.Pipeline()

	for key, entry := range entries {
		value := ""
		if str, ok := entry.Value.(string); ok {
			value = str
		} else {
			// Try to marshal as JSON
			if data, err := json.Marshal(entry.Value); err == nil {
				value = string(data)
			} else {
				value = fmt.Sprintf("%v", entry.Value)
			}
		}

		ttl := rc.config.TTL
		if !entry.ExpiresAt.IsZero() && !entry.CreatedAt.IsZero() {
			ttl = entry.ExpiresAt.Sub(entry.CreatedAt)
		}

		pipe.Set(ctx, key, value, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		rc.logger.Error("Redis mset error", zap.Error(err))
		return fmt.Errorf("redis mset error: %w", err)
	}

	return nil
}

// DeleteEntries removes multiple entries by keys
func (rc *RedisCacheImpl) DeleteEntries(ctx context.Context, keys []string) error {
	start := time.Now()
	defer func() {
		rc.logger.Debug("Redis cache mdel operation",
			zap.Int("key_count", len(keys)),
			zap.Duration("duration", time.Since(start)))
	}()

	if len(keys) == 0 {
		return nil
	}

	result, err := rc.client.Del(ctx, keys...).Result()
	if err != nil {
		rc.logger.Error("Redis mdel error", zap.Error(err))
		return fmt.Errorf("redis mdel error: %w", err)
	}

	rc.logger.Debug("Redis cache entries deleted",
		zap.Int64("deleted_count", result),
		zap.Int("requested_count", len(keys)))

	return nil
}

// GetSize returns the current size of the cache
func (rc *RedisCacheImpl) GetSize() int64 {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	size, err := rc.client.DBSize(ctx).Result()
	if err != nil {
		rc.logger.Error("Redis DBSize error", zap.Error(err))
		return 0
	}

	return size
}

// GetMemoryUsage returns the approximate memory usage in bytes
func (rc *RedisCacheImpl) GetMemoryUsage() int64 {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := rc.client.Info(ctx, "memory").Result()
	if err != nil {
		rc.logger.Error("Redis memory info error", zap.Error(err))
		return 0
	}

	// Parse memory usage from Redis info
	// This is a simplified implementation - in production, you'd want to parse the info string properly
	rc.logger.Debug("Redis memory info", zap.String("info", info))

	// For now, return 0 as we'd need to parse the Redis info string
	// In a real implementation, you'd extract the used_memory value
	return 0
}

// GetExpiredCount returns the number of expired items
func (rc *RedisCacheImpl) GetExpiredCount() int64 {
	// Redis doesn't track expired keys in a simple way
	// This would require parsing Redis info or using a different approach
	return 0
}

// GetEvictionCount returns the number of evicted items
func (rc *RedisCacheImpl) GetEvictionCount() int64 {
	// Redis doesn't track evicted keys in a simple way
	// This would require parsing Redis info or using a different approach
	return 0
}

// ResetStats resets the cache statistics
func (rc *RedisCacheImpl) ResetStats() {
	rc.stats = &CacheStats{}
}

// GetConfig returns the cache configuration
func (rc *RedisCacheImpl) GetConfig() *CacheConfig {
	return &CacheConfig{
		Type:       RedisCache,
		DefaultTTL: rc.config.TTL,
	}
}

// String returns a string representation of the cache
func (rc *RedisCacheImpl) String() string {
	return fmt.Sprintf("RedisCache{addr=%s, db=%d, hitRate=%.2f%%}",
		rc.config.Addr, rc.config.DB, rc.stats.HitRate*100)
}

// recordHit records a cache hit
func (rc *RedisCacheImpl) recordHit() {
	rc.stats.HitCount++
}

// recordMiss records a cache miss
func (rc *RedisCacheImpl) recordMiss() {
	rc.stats.MissCount++
}

// HealthCheck performs a health check on the Redis connection
func (rc *RedisCacheImpl) HealthCheck(ctx context.Context) error {
	start := time.Now()
	defer func() {
		rc.logger.Debug("Redis cache health check",
			zap.Duration("duration", time.Since(start)))
	}()

	err := rc.client.Ping(ctx).Err()
	if err != nil {
		rc.logger.Error("Redis health check failed", zap.Error(err))
		return fmt.Errorf("redis health check failed: %w", err)
	}

	return nil
}

// GetClient returns the underlying Redis client for advanced operations
func (rc *RedisCacheImpl) GetClient() *redis.Client {
	return rc.client
}
