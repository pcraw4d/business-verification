package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// CacheConfig represents Redis cache configuration
type CacheConfig struct {
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
	IdleCheckFreq     time.Duration `json:"idle_check_freq"`
	MaxConnAge        time.Duration `json:"max_conn_age"`
	DefaultTTL        time.Duration `json:"default_ttl"`
	KeyPrefix         string        `json:"key_prefix"`
	EnableMetrics     bool          `json:"enable_metrics"`
	EnableCompression bool          `json:"enable_compression"`
}

// CacheMetrics represents cache performance metrics
type CacheMetrics struct {
	Hits           int64         `json:"hits"`
	Misses         int64         `json:"misses"`
	Sets           int64         `json:"sets"`
	Deletes        int64         `json:"deletes"`
	Errors         int64         `json:"errors"`
	TotalRequests  int64         `json:"total_requests"`
	HitRate        float64       `json:"hit_rate"`
	AverageLatency time.Duration `json:"average_latency"`
	LastUpdated    time.Time     `json:"last_updated"`
}

// RedisCache implements distributed caching with Redis
type RedisCache struct {
	client  redis.UniversalClient
	config  *CacheConfig
	logger  *zap.Logger
	metrics *CacheMetrics
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(config *CacheConfig, logger *zap.Logger) (*RedisCache, error) {
	if config == nil {
		return nil, fmt.Errorf("cache config cannot be nil")
	}

	// Set default values
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
	if config.IdleCheckFreq == 0 {
		config.IdleCheckFreq = 1 * time.Minute
	}
	if config.MaxConnAge == 0 {
		config.MaxConnAge = 30 * time.Minute
	}
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 5 * time.Minute
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = "risk_assessment:"
	}

	// Create Redis client
	var client redis.UniversalClient
	if len(config.Addrs) == 1 {
		// Single Redis instance
		client = redis.NewClient(&redis.Options{
			Addr:          config.Addrs[0],
			Password:      config.Password,
			DB:            config.DB,
			PoolSize:      config.PoolSize,
			MinIdleConns:  config.MinIdleConns,
			MaxRetries:    config.MaxRetries,
			DialTimeout:   config.DialTimeout,
			ReadTimeout:   config.ReadTimeout,
			WriteTimeout:  config.WriteTimeout,
			PoolTimeout:   config.PoolTimeout,
			IdleTimeout:   config.IdleTimeout,
			IdleCheckFreq: config.IdleCheckFreq,
			MaxConnAge:    config.MaxConnAge,
		})
	} else {
		// Redis Cluster
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:         config.Addrs,
			Password:      config.Password,
			PoolSize:      config.PoolSize,
			MinIdleConns:  config.MinIdleConns,
			MaxRetries:    config.MaxRetries,
			DialTimeout:   config.DialTimeout,
			ReadTimeout:   config.ReadTimeout,
			WriteTimeout:  config.WriteTimeout,
			PoolTimeout:   config.PoolTimeout,
			IdleTimeout:   config.IdleTimeout,
			IdleCheckFreq: config.IdleCheckFreq,
			MaxConnAge:    config.MaxConnAge,
		})
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	cache := &RedisCache{
		client:  client,
		config:  config,
		logger:  logger,
		metrics: &CacheMetrics{},
	}

	// Start metrics collection if enabled
	if config.EnableMetrics {
		go cache.collectMetrics()
	}

	logger.Info("Redis cache initialized successfully",
		zap.Strings("addrs", config.Addrs),
		zap.Int("pool_size", config.PoolSize),
		zap.Duration("default_ttl", config.DefaultTTL))

	return cache, nil
}

// Get retrieves a value from cache
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	start := time.Now()
	defer func() {
		c.updateMetrics("get", time.Since(start), nil)
	}()

	fullKey := c.config.KeyPrefix + key

	val, err := c.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			c.metrics.Misses++
			return ErrCacheMiss
		}
		c.metrics.Errors++
		c.logger.Error("Cache get error",
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("cache get error: %w", err)
	}

	c.metrics.Hits++

	// Unmarshal JSON
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		c.metrics.Errors++
		return fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	return nil
}

// Set stores a value in cache
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		c.updateMetrics("set", time.Since(start), nil)
	}()

	fullKey := c.config.KeyPrefix + key

	// Marshal to JSON
	data, err := json.Marshal(value)
	if err != nil {
		c.metrics.Errors++
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Use default TTL if not specified
	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}

	if err := c.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		c.metrics.Errors++
		c.logger.Error("Cache set error",
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("cache set error: %w", err)
	}

	c.metrics.Sets++
	return nil
}

// Delete removes a value from cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	start := time.Now()
	defer func() {
		c.updateMetrics("delete", time.Since(start), nil)
	}()

	fullKey := c.config.KeyPrefix + key

	if err := c.client.Del(ctx, fullKey).Err(); err != nil {
		c.metrics.Errors++
		c.logger.Error("Cache delete error",
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("cache delete error: %w", err)
	}

	c.metrics.Deletes++
	return nil
}

// Exists checks if a key exists in cache
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := c.config.KeyPrefix + key

	count, err := c.client.Exists(ctx, fullKey).Result()
	if err != nil {
		c.metrics.Errors++
		return false, fmt.Errorf("cache exists error: %w", err)
	}

	return count > 0, nil
}

// GetOrSet retrieves a value from cache or sets it using the provided function
func (c *RedisCache) GetOrSet(ctx context.Context, key string, dest interface{}, ttl time.Duration, setter func() (interface{}, error)) error {
	// Try to get from cache first
	err := c.Get(ctx, key, dest)
	if err == nil {
		return nil
	}
	if err != ErrCacheMiss {
		return err
	}

	// Cache miss, call setter function
	value, err := setter()
	if err != nil {
		return fmt.Errorf("setter function failed: %w", err)
	}

	// Set in cache
	if err := c.Set(ctx, key, value, ttl); err != nil {
		c.logger.Warn("Failed to set cache value",
			zap.String("key", key),
			zap.Error(err))
	}

	// Set the value in dest
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal setter result: %w", err)
	}

	return json.Unmarshal(data, dest)
}

// MGet retrieves multiple values from cache
func (c *RedisCache) MGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = c.config.KeyPrefix + key
	}

	vals, err := c.client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		c.metrics.Errors++
		return nil, fmt.Errorf("cache mget error: %w", err)
	}

	result := make(map[string]interface{})
	for i, val := range vals {
		if val != nil {
			var value interface{}
			if err := json.Unmarshal([]byte(val.(string)), &value); err != nil {
				c.logger.Warn("Failed to unmarshal cached value",
					zap.String("key", keys[i]),
					zap.Error(err))
				continue
			}
			result[keys[i]] = value
			c.metrics.Hits++
		} else {
			c.metrics.Misses++
		}
	}

	return result, nil
}

// MSet stores multiple values in cache
func (c *RedisCache) MSet(ctx context.Context, values map[string]interface{}, ttl time.Duration) error {
	if len(values) == 0 {
		return nil
	}

	// Use pipeline for better performance
	pipe := c.client.Pipeline()

	for key, value := range values {
		fullKey := c.config.KeyPrefix + key

		data, err := json.Marshal(value)
		if err != nil {
			c.metrics.Errors++
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}

		if ttl == 0 {
			ttl = c.config.DefaultTTL
		}

		pipe.Set(ctx, fullKey, data, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		c.metrics.Errors++
		return fmt.Errorf("cache mset error: %w", err)
	}

	c.metrics.Sets += int64(len(values))
	return nil
}

// Clear removes all keys with the configured prefix
func (c *RedisCache) Clear(ctx context.Context) error {
	pattern := c.config.KeyPrefix + "*"

	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		c.metrics.Errors++
		return fmt.Errorf("failed to get keys for clearing: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		c.metrics.Errors++
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	c.metrics.Deletes += int64(len(keys))
	return nil
}

// GetMetrics returns current cache metrics
func (c *RedisCache) GetMetrics() *CacheMetrics {
	c.metrics.TotalRequests = c.metrics.Hits + c.metrics.Misses
	if c.metrics.TotalRequests > 0 {
		c.metrics.HitRate = float64(c.metrics.Hits) / float64(c.metrics.TotalRequests)
	}
	c.metrics.LastUpdated = time.Now()
	return c.metrics
}

// ResetMetrics resets cache metrics
func (c *RedisCache) ResetMetrics() {
	c.metrics = &CacheMetrics{}
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// Health checks cache health
func (c *RedisCache) Health(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// updateMetrics updates cache metrics
func (c *RedisCache) updateMetrics(operation string, latency time.Duration, err error) {
	if !c.config.EnableMetrics {
		return
	}

	if err != nil {
		c.metrics.Errors++
	}

	// Update average latency (simple moving average)
	if c.metrics.AverageLatency == 0 {
		c.metrics.AverageLatency = latency
	} else {
		c.metrics.AverageLatency = (c.metrics.AverageLatency + latency) / 2
	}
}

// collectMetrics periodically collects and logs cache metrics
func (c *RedisCache) collectMetrics() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		metrics := c.GetMetrics()
		c.logger.Info("Cache metrics",
			zap.Int64("hits", metrics.Hits),
			zap.Int64("misses", metrics.Misses),
			zap.Float64("hit_rate", metrics.HitRate),
			zap.Duration("avg_latency", metrics.AverageLatency),
			zap.Int64("errors", metrics.Errors))
	}
}

// Cache errors
var (
	ErrCacheMiss = fmt.Errorf("cache miss")
)
