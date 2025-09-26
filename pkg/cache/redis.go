package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the Redis client with KYB-specific functionality
type RedisClient struct {
	client *redis.Client
	prefix string
}

// NewRedisClient creates a new Redis client for caching
func NewRedisClient(addr, password, prefix string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{
		client: client,
		prefix: prefix,
	}, nil
}

// Set stores a value in Redis with expiration
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	fullKey := r.prefix + ":" + key

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.Set(ctx, fullKey, data, expiration).Err()
}

// Get retrieves a value from Redis
func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := r.prefix + ":" + key

	data, err := r.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get value: %w", err)
	}

	return json.Unmarshal(data, dest)
}

// Delete removes a value from Redis
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	fullKey := r.prefix + ":" + key
	return r.client.Del(ctx, fullKey).Err()
}

// Exists checks if a key exists in Redis
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := r.prefix + ":" + key
	result := r.client.Exists(ctx, fullKey)
	return result.Val() > 0, result.Err()
}

// SetNX sets a value only if the key doesn't exist
func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	fullKey := r.prefix + ":" + key

	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.SetNX(ctx, fullKey, data, expiration).Result()
}

// Increment increments a counter in Redis
func (r *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	fullKey := r.prefix + ":" + key
	return r.client.Incr(ctx, fullKey).Result()
}

// Expire sets expiration for a key
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	fullKey := r.prefix + ":" + key
	return r.client.Expire(ctx, fullKey, expiration).Err()
}

// Health checks Redis connectivity
func (r *RedisClient) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Cache keys for different data types
const (
	ClassificationCacheKey = "classification"
	MerchantCacheKey       = "merchant"
	UserCacheKey           = "user"
	SessionCacheKey        = "session"
	RateLimitCacheKey      = "rate_limit"
)

// Cache expiration times
const (
	ClassificationCacheExpiration = 1 * time.Hour
	MerchantCacheExpiration       = 30 * time.Minute
	UserCacheExpiration           = 15 * time.Minute
	SessionCacheExpiration        = 24 * time.Hour
	RateLimitCacheExpiration      = 1 * time.Minute
)
