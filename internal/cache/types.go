package cache

import (
	"context"
	"errors"
	"time"
)

// CacheType represents the type of cache backend
type CacheType string

const (
	MemoryCache CacheType = "memory"
	RedisCache  CacheType = "redis"
	FileCache   CacheType = "file"
)

// MemoryCacheConfig holds configuration for the memory cache
type MemoryCacheConfig struct {
	Size   int           `json:"size"`   // Maximum number of items
	TTL    time.Duration `json:"ttl"`    // Time to live
	Policy string        `json:"policy"` // LRU, LFU, FIFO
}

// RedisCacheConfig holds configuration for the Redis cache
type RedisCacheConfig struct {
	Addr     string        `json:"addr"`      // Redis server address
	Password string        `json:"password"`  // Redis password
	DB       int           `json:"db"`        // Redis database
	TTL      time.Duration `json:"ttl"`       // Time to live
	PoolSize int           `json:"pool_size"` // Connection pool size
}

// CacheConfig holds general cache configuration
type CacheConfig struct {
	Type             CacheType     `json:"type"`               // Cache type
	DefaultTTL       time.Duration `json:"default_ttl"`        // Default time to live
	MaxSize          int           `json:"max_size"`           // Maximum number of items
	KeyPrefix        string        `json:"key_prefix"`         // Key prefix
	KeySeparator     string        `json:"key_separator"`      // Key separator
	KeyHashAlgorithm string        `json:"key_hash_algorithm"` // Key hash algorithm
	CleanupInterval  time.Duration `json:"cleanup_interval"`   // Cleanup interval
	MetricsInterval  time.Duration `json:"metrics_interval"`   // Metrics collection interval
}

// Cache interface defines the contract for cache implementations
type Cache interface {
	// Basic operations
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// TTL operations
	GetTTL(ctx context.Context, key string) (time.Duration, error)
	SetTTL(ctx context.Context, key string, ttl time.Duration) error

	// Bulk operations
	GetEntries(ctx context.Context, keys []string) (map[string]*CacheEntry, error)
	SetEntries(ctx context.Context, entries map[string]*CacheEntry) error
	DeleteEntries(ctx context.Context, keys []string) error

	// Utility operations
	Clear(ctx context.Context) error
	GetKeys(ctx context.Context, pattern string) ([]string, error)
	GetStats(ctx context.Context) (*CacheStats, error)
	Close() error

	// Monitoring
	GetSize() int64
	GetMemoryUsage() int64
	GetExpiredCount() int64
	GetEvictionCount() int64
	ResetStats()
	GetConfig() *CacheConfig
	String() string
}

// Cache errors
var (
	CacheNotFoundError = errors.New("cache item not found")
)
