package cache

import (
	"context"
	"time"
)

// Cache defines the interface for caching operations
type Cache interface {
	// Basic operations
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	
	// Advanced operations
	GetOrSet(ctx context.Context, key string, dest interface{}, ttl time.Duration, setter func() (interface{}, error)) error
	MGet(ctx context.Context, keys []string) (map[string]interface{}, error)
	MSet(ctx context.Context, values map[string]interface{}, ttl time.Duration) error
	Clear(ctx context.Context) error
	
	// Management operations
	GetMetrics() *CacheMetrics
	ResetMetrics()
	Close() error
	Health(ctx context.Context) error
}

// CacheFactory creates cache instances
type CacheFactory struct {
	config *CacheConfig
	logger interface{} // Will be replaced with proper logger type
}

// NewCacheFactory creates a new cache factory
func NewCacheFactory(config *CacheConfig, logger interface{}) *CacheFactory {
	return &CacheFactory{
		config: config,
		logger: logger,
	}
}

// CreateRedisCache creates a Redis cache instance
func (f *CacheFactory) CreateRedisCache() (Cache, error) {
	// Type assertion for logger - this will be fixed when we have proper logger type
	logger, ok := f.logger.(interface {
		Info(string, ...interface{})
		Error(string, ...interface{})
		Warn(string, ...interface{})
	})
	if !ok {
		// Create a no-op logger if type assertion fails
		logger = &noOpLogger{}
	}
	
	return NewRedisCache(f.config, logger)
}

// noOpLogger is a no-operation logger for fallback
type noOpLogger struct{}

func (l *noOpLogger) Info(msg string, args ...interface{})  {}
func (l *noOpLogger) Error(msg string, args ...interface{}) {}
func (l *noOpLogger) Warn(msg string, args ...interface{})  {}
