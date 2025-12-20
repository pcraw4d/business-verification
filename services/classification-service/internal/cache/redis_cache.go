package cache

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// maskRedisURL masks sensitive parts of Redis URL for logging
func maskRedisURL(url string) string {
	if url == "" {
		return ""
	}
	// Mask password if present (format: redis://user:password@host:port)
	if strings.Contains(url, "@") {
		parts := strings.Split(url, "@")
		if len(parts) == 2 {
			// Mask the part before @ (contains password)
			authPart := parts[0]
			if strings.Contains(authPart, ":") {
				authParts := strings.Split(authPart, ":")
				if len(authParts) >= 3 {
					// redis://user:password format
					return fmt.Sprintf("redis://%s:***@%s", authParts[1], parts[1])
				}
			}
		}
	}
	// If no password detected, return as-is (or mask host if needed)
	return url
}

// RedisCache provides distributed caching with Redis, with fallback to in-memory cache
type RedisCache struct {
	client        *redis.Client
	prefix        string
	fallbackCache map[string]*cacheEntry
	fallbackMutex sync.RWMutex
	logger        *zap.Logger
	enabled       bool
}

type cacheEntry struct {
	Data      []byte
	ExpiresAt time.Time
}


// NewRedisCache creates a new Redis cache with fallback to in-memory cache
func NewRedisCache(redisURL string, prefix string, logger *zap.Logger) *RedisCache {
	rc := &RedisCache{
		prefix:        prefix,
		fallbackCache: make(map[string]*cacheEntry),
		logger:        logger,
		enabled:       false,
	}

	// Parse Redis URL and create client
	if redisURL != "" {
		// Fix: Handle Railway template variables - strip template syntax if not resolved
		// Railway template variables like "r${{ Redis.REDIS_URL }}" need to be resolved
		// If template variable not resolved, try to extract actual URL or use as-is
		actualRedisURL := redisURL
		if strings.Contains(redisURL, "${{") {
			// Template variable not resolved - try to extract or log warning
			logger.Warn("Redis URL appears to contain unresolved template variable",
				zap.String("redis_url", maskRedisURL(redisURL)),
				zap.String("hint", "Ensure Railway template variables are resolved"))
			// Try to extract Redis URL from template if possible
			// For now, return early to use in-memory cache
			return rc
		}
		
		opt, err := redis.ParseURL(actualRedisURL)
		if err != nil {
			logger.Warn("Failed to parse Redis URL, using in-memory cache only",
				zap.String("redis_url", maskRedisURL(actualRedisURL)),
				zap.Error(err))
			return rc
		}

		rc.client = redis.NewClient(opt)

		// Test connection with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := rc.client.Ping(ctx).Err(); err != nil {
			logger.Warn("Redis connection failed, using in-memory cache only",
				zap.Error(err))
			rc.client = nil
			return rc
		}

		rc.enabled = true
		logger.Info("Redis cache enabled",
			zap.String("prefix", prefix))
	} else {
		logger.Info("Redis URL not provided, using in-memory cache only")
	}

	return rc
}

// Get retrieves a value from cache (Redis first, then fallback)
// Returns the JSON bytes and a boolean indicating if the value was found
func (rc *RedisCache) Get(ctx context.Context, key string) ([]byte, bool) {
	fullKey := rc.getFullKey(key)

	// Try Redis first if enabled
	if rc.enabled && rc.client != nil {
		data, err := rc.client.Get(ctx, fullKey).Bytes()
		if err == nil {
			rc.logger.Debug("Cache hit from Redis",
				zap.String("key", key))
			return data, true
		} else if err != redis.Nil {
			rc.logger.Warn("Redis get error, falling back to in-memory cache",
				zap.String("key", key),
				zap.Error(err))
		}
	}

	// Fallback to in-memory cache
	rc.fallbackMutex.RLock()
	defer rc.fallbackMutex.RUnlock()

	entry, exists := rc.fallbackCache[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		// Expired - remove it
		rc.fallbackMutex.RUnlock()
		rc.fallbackMutex.Lock()
		delete(rc.fallbackCache, key)
		rc.fallbackMutex.Unlock()
		rc.fallbackMutex.RLock()
		return nil, false
	}

	rc.logger.Debug("Cache hit from in-memory fallback",
		zap.String("key", key))
	return entry.Data, true
}

// Set stores a value in cache (both Redis and fallback)
// Accepts JSON bytes to avoid circular dependencies
func (rc *RedisCache) Set(ctx context.Context, key string, data []byte, ttl time.Duration) {
	fullKey := rc.getFullKey(key)

	// Store in Redis if enabled
	if rc.enabled && rc.client != nil {
		if err := rc.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
			rc.logger.Warn("Failed to store in Redis, using in-memory cache only",
				zap.String("key", key),
				zap.Error(err))
		} else {
			rc.logger.Debug("Stored in Redis cache",
				zap.String("key", key),
				zap.Duration("ttl", ttl))
		}
	}

	// Always store in fallback cache
	rc.fallbackMutex.Lock()
	rc.fallbackCache[key] = &cacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
	}
	rc.fallbackMutex.Unlock()

	rc.logger.Debug("Stored in fallback cache",
		zap.String("key", key))
}

// Delete removes a value from cache (both Redis and fallback)
func (rc *RedisCache) Delete(ctx context.Context, key string) {
	fullKey := rc.getFullKey(key)

	// Delete from Redis if enabled
	if rc.enabled && rc.client != nil {
		if err := rc.client.Del(ctx, fullKey).Err(); err != nil {
			rc.logger.Warn("Failed to delete from Redis",
				zap.String("key", key),
				zap.Error(err))
		}
	}

	// Delete from fallback cache
	rc.fallbackMutex.Lock()
	delete(rc.fallbackCache, key)
	rc.fallbackMutex.Unlock()
}

// Clear clears all cache entries (both Redis and fallback)
func (rc *RedisCache) Clear(ctx context.Context) error {
	// Clear Redis if enabled
	if rc.enabled && rc.client != nil {
		pattern := rc.prefix + ":*"
		iter := rc.client.Scan(ctx, 0, pattern, 0).Iterator()
		for iter.Next(ctx) {
			if err := rc.client.Del(ctx, iter.Val()).Err(); err != nil {
				rc.logger.Warn("Failed to delete key from Redis during clear",
					zap.String("key", iter.Val()),
					zap.Error(err))
			}
		}
		if err := iter.Err(); err != nil {
			return fmt.Errorf("failed to clear Redis cache: %w", err)
		}
	}

	// Clear fallback cache
	rc.fallbackMutex.Lock()
	rc.fallbackCache = make(map[string]*cacheEntry)
	rc.fallbackMutex.Unlock()

	return nil
}

// Health checks Redis connectivity
func (rc *RedisCache) Health(ctx context.Context) error {
	if !rc.enabled || rc.client == nil {
		return fmt.Errorf("Redis cache not enabled")
	}
	return rc.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (rc *RedisCache) Close() error {
	if rc.client != nil {
		return rc.client.Close()
	}
	return nil
}

// getFullKey returns the full cache key with prefix
func (rc *RedisCache) getFullKey(key string) string {
	if rc.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", rc.prefix, key)
}


