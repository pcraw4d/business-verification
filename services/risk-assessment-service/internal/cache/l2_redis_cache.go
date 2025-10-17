package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// L2RedisCache implements a Redis-based distributed cache
type L2RedisCache struct {
	client     *redis.Client
	logger     *zap.Logger
	keyPrefix  string
	defaultTTL time.Duration
	stats      *RedisCacheStats
}

// RedisCacheStats represents statistics for the Redis cache
type RedisCacheStats struct {
	Hits       int64     `json:"hits"`
	Misses     int64     `json:"misses"`
	Sets       int64     `json:"sets"`
	Deletes    int64     `json:"deletes"`
	Errors     int64     `json:"errors"`
	HitRate    float64   `json:"hit_rate"`
	LastAccess time.Time `json:"last_access"`
}

// RedisCacheConfig represents configuration for the Redis cache
type RedisCacheConfig struct {
	Addrs             []string      `json:"addrs"`
	Password          string        `json:"password"`
	DB                int           `json:"db"`
	PoolSize          int           `json:"pool_size"`
	MinIdleConns      int           `json:"min_idle_conns"`
	MaxRetries        int           `json:"max_retries"`
	DialTimeout       time.Duration `json:"dial_timeout"`
	ReadTimeout       time.Duration `json:"read_timeout"`
	WriteTimeout      time.Duration `json:"write_timeout"`
	PoolTimeout       time.Duration `json:"pool_timeout"`
	IdleTimeout       time.Duration `json:"idle_timeout"`
	MaxConnAge        time.Duration `json:"max_conn_age"`
	DefaultTTL        time.Duration `json:"default_ttl"`
	KeyPrefix         string        `json:"key_prefix"`
	EnableMetrics     bool          `json:"enable_metrics"`
	EnableCompression bool          `json:"enable_compression"`
}

// NewL2RedisCache creates a new L2 Redis cache
func NewL2RedisCache(config *RedisCacheConfig, logger *zap.Logger) (*L2RedisCache, error) {
	if config == nil {
		return nil, fmt.Errorf("redis cache config cannot be nil")
	}

	// Set defaults
	if config.PoolSize == 0 {
		config.PoolSize = 10
	}
	if config.MinIdleConns == 0 {
		config.MinIdleConns = 5
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.DialTimeout == 0 {
		config.DialTimeout = 5 * time.Second
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 3 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 3 * time.Second
	}
	if config.PoolTimeout == 0 {
		config.PoolTimeout = 4 * time.Second
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = 5 * time.Minute
	}
	if config.MaxConnAge == 0 {
		config.MaxConnAge = 30 * time.Minute
	}
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 5 * time.Minute
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = "ra:"
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         config.Addrs[0], // Use first address for single Redis instance
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		PoolTimeout:  config.PoolTimeout,
		IdleTimeout:  config.IdleTimeout,
		MaxConnAge:   config.MaxConnAge,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &L2RedisCache{
		client:     client,
		logger:     logger,
		keyPrefix:  config.KeyPrefix,
		defaultTTL: config.DefaultTTL,
		stats:      &RedisCacheStats{},
	}, nil
}

// Get retrieves a value from the Redis cache
func (c *L2RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	fullKey := c.keyPrefix + key

	val, err := c.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			c.stats.Misses++
			return nil, nil // Cache miss
		}
		c.stats.Errors++
		return nil, fmt.Errorf("failed to get from Redis: %w", err)
	}

	// Deserialize value
	var value interface{}
	if err := json.Unmarshal([]byte(val), &value); err != nil {
		c.stats.Errors++
		return nil, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	c.stats.Hits++
	c.stats.LastAccess = time.Now()

	return value, nil
}

// Set stores a value in the Redis cache
func (c *L2RedisCache) Set(ctx context.Context, key string, value interface{}) error {
	return c.SetWithTTL(ctx, key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the Redis cache with a specific TTL
func (c *L2RedisCache) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := c.keyPrefix + key

	// Serialize value
	data, err := json.Marshal(value)
	if err != nil {
		c.stats.Errors++
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Set in Redis
	if err := c.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		c.stats.Errors++
		return fmt.Errorf("failed to set in Redis: %w", err)
	}

	c.stats.Sets++
	return nil
}

// Delete removes a value from the Redis cache
func (c *L2RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := c.keyPrefix + key

	if err := c.client.Del(ctx, fullKey).Err(); err != nil {
		c.stats.Errors++
		return fmt.Errorf("failed to delete from Redis: %w", err)
	}

	c.stats.Deletes++
	return nil
}

// Exists checks if a key exists in the Redis cache
func (c *L2RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := c.keyPrefix + key

	count, err := c.client.Exists(ctx, fullKey).Result()
	if err != nil {
		c.stats.Errors++
		return false, fmt.Errorf("failed to check existence in Redis: %w", err)
	}

	return count > 0, nil
}

// Expire sets an expiration time for a key
func (c *L2RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := c.keyPrefix + key

	if err := c.client.Expire(ctx, fullKey, ttl).Err(); err != nil {
		c.stats.Errors++
		return fmt.Errorf("failed to set expiration in Redis: %w", err)
	}

	return nil
}

// TTL returns the time to live for a key
func (c *L2RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := c.keyPrefix + key

	ttl, err := c.client.TTL(ctx, fullKey).Result()
	if err != nil {
		c.stats.Errors++
		return 0, fmt.Errorf("failed to get TTL from Redis: %w", err)
	}

	return ttl, nil
}

// GetMultiple retrieves multiple values from the Redis cache
func (c *L2RedisCache) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	// Add prefix to all keys
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = c.keyPrefix + key
	}

	// Get multiple values
	vals, err := c.client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		c.stats.Errors++
		return nil, fmt.Errorf("failed to get multiple from Redis: %w", err)
	}

	// Process results
	result := make(map[string]interface{})
	for i, val := range vals {
		if val != nil {
			// Deserialize value
			var value interface{}
			if err := json.Unmarshal([]byte(val.(string)), &value); err != nil {
				c.stats.Errors++
				continue
			}
			result[keys[i]] = value
			c.stats.Hits++
		} else {
			c.stats.Misses++
		}
	}

	c.stats.LastAccess = time.Now()
	return result, nil
}

// SetMultiple stores multiple values in the Redis cache
func (c *L2RedisCache) SetMultiple(ctx context.Context, items map[string]interface{}) error {
	return c.SetMultipleWithTTL(ctx, items, c.defaultTTL)
}

// SetMultipleWithTTL stores multiple values in the Redis cache with a specific TTL
func (c *L2RedisCache) SetMultipleWithTTL(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	// Prepare pipeline
	pipe := c.client.Pipeline()

	for key, value := range items {
		fullKey := c.keyPrefix + key

		// Serialize value
		data, err := json.Marshal(value)
		if err != nil {
			c.stats.Errors++
			continue
		}

		// Add to pipeline
		pipe.Set(ctx, fullKey, data, ttl)
	}

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		c.stats.Errors++
		return fmt.Errorf("failed to set multiple in Redis: %w", err)
	}

	c.stats.Sets += int64(len(items))
	return nil
}

// Clear removes all keys with the prefix from the Redis cache
func (c *L2RedisCache) Clear(ctx context.Context) error {
	pattern := c.keyPrefix + "*"

	// Get all keys matching pattern
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		c.stats.Errors++
		return fmt.Errorf("failed to get keys from Redis: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	// Delete all keys
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		c.stats.Errors++
		return fmt.Errorf("failed to clear Redis cache: %w", err)
	}

	c.stats.Deletes += int64(len(keys))
	return nil
}

// Stats returns cache statistics
func (c *L2RedisCache) Stats() *RedisCacheStats {
	stats := *c.stats

	// Calculate hit rate
	total := stats.Hits + stats.Misses
	if total > 0 {
		stats.HitRate = float64(stats.Hits) / float64(total)
	}

	return &stats
}

// GetRedisInfo returns Redis server information
func (c *L2RedisCache) GetRedisInfo(ctx context.Context) (map[string]string, error) {
	info, err := c.client.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// Parse info string into map
	result := make(map[string]string)
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				result[parts[0]] = parts[1]
			}
		}
	}

	return result, nil
}

// GetMemoryUsage returns Redis memory usage information
func (c *L2RedisCache) GetMemoryUsage(ctx context.Context) (map[string]interface{}, error) {
	info, err := c.GetRedisInfo(ctx)
	if err != nil {
		return nil, err
	}

	memoryInfo := make(map[string]interface{})

	// Extract memory-related information
	if used, ok := info["used_memory"]; ok {
		memoryInfo["used_memory"] = used
	}
	if peak, ok := info["used_memory_peak"]; ok {
		memoryInfo["used_memory_peak"] = peak
	}
	if rss, ok := info["used_memory_rss"]; ok {
		memoryInfo["used_memory_rss"] = rss
	}
	if fragmentation, ok := info["mem_fragmentation_ratio"]; ok {
		memoryInfo["mem_fragmentation_ratio"] = fragmentation
	}

	return memoryInfo, nil
}

// Health checks the Redis connection health
func (c *L2RedisCache) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}

// Close closes the Redis connection
func (c *L2RedisCache) Close() error {
	return c.client.Close()
}

// GetClient returns the underlying Redis client
func (c *L2RedisCache) GetClient() *redis.Client {
	return c.client
}

// Helper methods

func (c *L2RedisCache) incrementStats(stat string) {
	switch stat {
	case "hits":
		c.stats.Hits++
	case "misses":
		c.stats.Misses++
	case "sets":
		c.stats.Sets++
	case "deletes":
		c.stats.Deletes++
	case "errors":
		c.stats.Errors++
	}
	c.stats.LastAccess = time.Now()
}
