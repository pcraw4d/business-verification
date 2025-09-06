package cache

import (
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

// Cache errors
var (
	CacheNotFoundError = errors.New("cache item not found")
)
