package cache

import "time"

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
