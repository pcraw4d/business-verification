package cache

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// CacheFactory creates cache instances based on configuration
type CacheFactory struct {
	logger *zap.Logger
}

// NewCacheFactory creates a new cache factory
func NewCacheFactory(logger *zap.Logger) *CacheFactory {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &CacheFactory{
		logger: logger,
	}
}

// CreateCache creates a cache instance based on the configuration
func (cf *CacheFactory) CreateCache(config *CacheConfig) (Cache, error) {
	if config == nil {
		return nil, fmt.Errorf("cache configuration cannot be nil")
	}

	// CacheConfig from types.go has Type field
	// We assume the config passed is from types.go (which has Type field)
	// If optimization.go CacheConfig is passed, it will fail at compile time
	// which is expected - use types.go CacheConfig instead
	cacheType := config.Type

	switch cacheType {
	case MemoryCache:
		return cf.createMemoryCache(config)
	case RedisCache:
		return cf.createRedisCache(config)
	case FileCache:
		return cf.createFileCache(config)
	default:
		return nil, fmt.Errorf("unsupported cache type: %s", cacheType)
	}
}

// createMemoryCache creates a memory cache instance
func (cf *CacheFactory) createMemoryCache(config *CacheConfig) (Cache, error) {
	cf.logger.Info("Creating memory cache",
		zap.Int("max_size", config.MaxSize),
		zap.Duration("default_ttl", config.DefaultTTL))

	return NewMemoryCache(config), nil
}

// createRedisCache creates a Redis cache instance
func (cf *CacheFactory) createRedisCache(config *CacheConfig) (Cache, error) {
	// Convert CacheConfig to RedisCacheConfig
	redisConfig := &RedisCacheConfig{
		Addr:     "localhost:6379", // Default Redis address
		Password: "",               // Default no password
		DB:       0,                // Default database
		TTL:      config.DefaultTTL,
		PoolSize: 10, // Default pool size
	}

	cf.logger.Info("Creating Redis cache",
		zap.String("addr", redisConfig.Addr),
		zap.Int("db", redisConfig.DB),
		zap.Duration("default_ttl", redisConfig.TTL))

	return NewRedisCache(redisConfig, cf.logger)
}

// createFileCache creates a file cache instance
func (cf *CacheFactory) createFileCache(config *CacheConfig) (Cache, error) {
	cf.logger.Info("Creating file cache",
		zap.Duration("default_ttl", config.DefaultTTL))

	// For now, return a memory cache as file cache is not implemented
	// In a real implementation, you would create a file-based cache
	cf.logger.Warn("File cache not implemented, falling back to memory cache")
	return cf.createMemoryCache(config)
}

// CreateDefaultCache creates a cache with default configuration
func (cf *CacheFactory) CreateDefaultCache() (Cache, error) {
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         1000,
		KeyPrefix:       "kyb",
		KeySeparator:    ":",
		CleanupInterval: 5 * time.Minute,
		MetricsInterval: 1 * time.Minute,
	}

	return cf.CreateCache(config)
}

// CreateRedisCacheWithConfig creates a Redis cache with specific configuration
func (cf *CacheFactory) CreateRedisCacheWithConfig(redisConfig *RedisCacheConfig) (Cache, error) {
	if redisConfig == nil {
		return nil, fmt.Errorf("Redis configuration cannot be nil")
	}

	cf.logger.Info("Creating Redis cache with custom configuration",
		zap.String("addr", redisConfig.Addr),
		zap.Int("db", redisConfig.DB),
		zap.Int("pool_size", redisConfig.PoolSize))

	return NewRedisCache(redisConfig, cf.logger)
}

// CreateMemoryCacheWithConfig creates a memory cache with specific configuration
func (cf *CacheFactory) CreateMemoryCacheWithConfig(config *CacheConfig) (Cache, error) {
	if config == nil {
		return nil, fmt.Errorf("cache configuration cannot be nil")
	}

	// Ensure it's a memory cache
	config.Type = MemoryCache

	cf.logger.Info("Creating memory cache with custom configuration",
		zap.Int("max_size", config.MaxSize),
		zap.Duration("default_ttl", config.DefaultTTL))

	return NewMemoryCache(config), nil
}
