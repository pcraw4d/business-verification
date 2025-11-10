package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// SimpleRedisCache implements a Redis-based cache
// (renamed to avoid conflict with redis.go RedisCacheImpl)
type SimpleRedisCache struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

// NewSimpleRedisCache creates a new Redis cache instance
// (renamed to avoid conflict with redis.go NewRedisCache)
func NewSimpleRedisCache(addr, password string, db int, prefix string, ttl time.Duration) (*SimpleRedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &SimpleRedisCache{
		client: client,
		prefix: prefix,
		ttl:    ttl,
	}, nil
}

// Get retrieves a value from cache
func (c *SimpleRedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	fullKey := c.getFullKey(key)

	data, err := c.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("cache get failed: %w", err)
	}

	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	return value, nil
}

// Set stores a value in cache
func (c *SimpleRedisCache) Set(ctx context.Context, key string, value interface{}) error {
	return c.SetWithTTL(ctx, key, value, c.ttl)
}

// SetWithTTL stores a value in cache with custom TTL
func (c *SimpleRedisCache) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := c.getFullKey(key)

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	if err := c.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("cache set failed: %w", err)
	}

	return nil
}

// Delete removes a value from cache
func (c *SimpleRedisCache) Delete(ctx context.Context, key string) error {
	fullKey := c.getFullKey(key)
	return c.client.Del(ctx, fullKey).Err()
}

// Clear removes all keys with the cache prefix
func (c *SimpleRedisCache) Clear(ctx context.Context) error {
	pattern := c.prefix + "*"
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed to delete key %s: %w", iter.Val(), err)
		}
	}

	return iter.Err()
}

// Exists checks if a key exists in cache
func (c *SimpleRedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := c.getFullKey(key)
	count, err := c.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// getFullKey returns the full cache key with prefix
func (c *SimpleRedisCache) getFullKey(key string) string {
	if c.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", c.prefix, key)
}

// Close closes the Redis connection
func (c *SimpleRedisCache) Close() error {
	return c.client.Close()
}

// Health checks the health of the Redis connection
func (c *SimpleRedisCache) Health(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
