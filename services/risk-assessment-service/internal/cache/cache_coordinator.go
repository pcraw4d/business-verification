package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CacheCoordinator manages multi-level caching with automatic failover
type CacheCoordinator struct {
	l1Cache  *L1MemoryCache
	l2Cache  *L2RedisCache
	logger   *zap.Logger
	stats    *CoordinatorStats
	mu       sync.RWMutex
	enabled  bool
	fallback bool
}

// CoordinatorStats represents statistics for the cache coordinator
type CoordinatorStats struct {
	L1Hits      int64   `json:"l1_hits"`
	L2Hits      int64   `json:"l2_hits"`
	L1Misses    int64   `json:"l1_misses"`
	L2Misses    int64   `json:"l2_misses"`
	L1Sets      int64   `json:"l1_sets"`
	L2Sets      int64   `json:"l2_sets"`
	L1Deletes   int64   `json:"l1_deletes"`
	L2Deletes   int64   `json:"l2_deletes"`
	L1Evictions int64   `json:"l1_evictions"`
	L2Errors    int64   `json:"l2_errors"`
	TotalHits   int64   `json:"total_hits"`
	TotalMisses int64   `json:"total_misses"`
	HitRate     float64 `json:"hit_rate"`
}

// CacheCoordinatorConfig represents configuration for the cache coordinator
type CacheCoordinatorConfig struct {
	L1Config       *MemoryCacheConfig `json:"l1_config"`
	L2Config       *RedisCacheConfig  `json:"l2_config"`
	EnableL1       bool               `json:"enable_l1"`
	EnableL2       bool               `json:"enable_l2"`
	EnableFallback bool               `json:"enable_fallback"`
	SyncInterval   time.Duration      `json:"sync_interval"`
}

// NewCacheCoordinator creates a new cache coordinator
func NewCacheCoordinator(config *CacheCoordinatorConfig, logger *zap.Logger) (*CacheCoordinator, error) {
	if config == nil {
		return nil, fmt.Errorf("cache coordinator config cannot be nil")
	}

	coordinator := &CacheCoordinator{
		logger:   logger,
		stats:    &CoordinatorStats{},
		enabled:  true,
		fallback: config.EnableFallback,
	}

	// Initialize L1 cache
	if config.EnableL1 && config.L1Config != nil {
		coordinator.l1Cache = NewL1MemoryCacheWithConfig(config.L1Config)

		// Set eviction callback to sync with L2
		coordinator.l1Cache.SetOnEvict(func(key string, value interface{}) {
			// Optionally sync evicted items to L2
			if coordinator.l2Cache != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if err := coordinator.l2Cache.Set(ctx, key, value); err != nil {
					coordinator.logger.Warn("Failed to sync evicted item to L2",
						zap.String("key", key),
						zap.Error(err))
				}
			}
		})
	}

	// Initialize L2 cache
	if config.EnableL2 && config.L2Config != nil {
		l2Cache, err := NewL2RedisCache(config.L2Config, logger)
		if err != nil {
			if config.EnableFallback {
				coordinator.logger.Warn("Failed to initialize L2 cache, continuing with L1 only",
					zap.Error(err))
			} else {
				return nil, fmt.Errorf("failed to initialize L2 cache: %w", err)
			}
		} else {
			coordinator.l2Cache = l2Cache
		}
	}

	// Start sync routine if both caches are enabled
	if coordinator.l1Cache != nil && coordinator.l2Cache != nil && config.SyncInterval > 0 {
		go coordinator.startSyncRoutine(config.SyncInterval)
	}

	coordinator.logger.Info("Cache coordinator initialized",
		zap.Bool("l1_enabled", coordinator.l1Cache != nil),
		zap.Bool("l2_enabled", coordinator.l2Cache != nil),
		zap.Bool("fallback_enabled", coordinator.fallback))

	return coordinator, nil
}

// Get retrieves a value from the cache hierarchy
func (cc *CacheCoordinator) Get(ctx context.Context, key string) (interface{}, error) {
	if !cc.enabled {
		return nil, fmt.Errorf("cache coordinator is disabled")
	}

	// Try L1 cache first
	if cc.l1Cache != nil {
		if value, found := cc.l1Cache.Get(key); found {
			cc.mu.Lock()
			cc.stats.L1Hits++
			cc.stats.TotalHits++
			cc.mu.Unlock()
			return value, nil
		}

		cc.mu.Lock()
		cc.stats.L1Misses++
		cc.mu.Unlock()
	}

	// Try L2 cache
	if cc.l2Cache != nil {
		value, err := cc.l2Cache.Get(ctx, key)
		if err != nil {
			cc.mu.Lock()
			cc.stats.L2Errors++
			cc.mu.Unlock()

			if !cc.fallback {
				return nil, fmt.Errorf("L2 cache error: %w", err)
			}

			cc.logger.Warn("L2 cache error, continuing without cache",
				zap.String("key", key),
				zap.Error(err))
			return nil, nil
		}

		if value != nil {
			// Cache hit in L2, promote to L1
			if cc.l1Cache != nil {
				cc.l1Cache.Set(key, value)
			}

			cc.mu.Lock()
			cc.stats.L2Hits++
			cc.stats.TotalHits++
			cc.mu.Unlock()

			return value, nil
		}

		cc.mu.Lock()
		cc.stats.L2Misses++
		cc.mu.Unlock()
	}

	// Cache miss
	cc.mu.Lock()
	cc.stats.TotalMisses++
	cc.mu.Unlock()

	return nil, nil
}

// Set stores a value in the cache hierarchy
func (cc *CacheCoordinator) Set(ctx context.Context, key string, value interface{}) error {
	return cc.SetWithTTL(ctx, key, value, 0)
}

// SetWithTTL stores a value in the cache hierarchy with a specific TTL
func (cc *CacheCoordinator) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !cc.enabled {
		return fmt.Errorf("cache coordinator is disabled")
	}

	var l2Err error

	// Set in L1 cache
	if cc.l1Cache != nil {
		if ttl > 0 {
			cc.l1Cache.SetWithTTL(key, value, ttl)
		} else {
			cc.l1Cache.Set(key, value)
		}

		cc.mu.Lock()
		cc.stats.L1Sets++
		cc.mu.Unlock()
	}

	// Set in L2 cache
	if cc.l2Cache != nil {
		if ttl > 0 {
			l2Err = cc.l2Cache.SetWithTTL(ctx, key, value, ttl)
		} else {
			l2Err = cc.l2Cache.Set(ctx, key, value)
		}

		if l2Err != nil {
			cc.mu.Lock()
			cc.stats.L2Errors++
			cc.mu.Unlock()

			if !cc.fallback {
				return fmt.Errorf("L2 cache error: %w", l2Err)
			}

			cc.logger.Warn("L2 cache error during set, continuing with L1 only",
				zap.String("key", key),
				zap.Error(l2Err))
		} else {
			cc.mu.Lock()
			cc.stats.L2Sets++
			cc.mu.Unlock()
		}
	}

	return nil
}

// Delete removes a value from the cache hierarchy
func (cc *CacheCoordinator) Delete(ctx context.Context, key string) error {
	if !cc.enabled {
		return fmt.Errorf("cache coordinator is disabled")
	}

	var l2Err error

	// Delete from L1 cache
	if cc.l1Cache != nil {
		cc.l1Cache.Delete(key)
		cc.mu.Lock()
		cc.stats.L1Deletes++
		cc.mu.Unlock()
	}

	// Delete from L2 cache
	if cc.l2Cache != nil {
		l2Err = cc.l2Cache.Delete(ctx, key)
		if l2Err != nil {
			cc.mu.Lock()
			cc.stats.L2Errors++
			cc.mu.Unlock()

			if !cc.fallback {
				return fmt.Errorf("L2 cache error: %w", l2Err)
			}

			cc.logger.Warn("L2 cache error during delete, continuing with L1 only",
				zap.String("key", key),
				zap.Error(l2Err))
		} else {
			cc.mu.Lock()
			cc.stats.L2Deletes++
			cc.mu.Unlock()
		}
	}

	return nil
}

// Clear removes all values from the cache hierarchy
func (cc *CacheCoordinator) Clear(ctx context.Context) error {
	if !cc.enabled {
		return fmt.Errorf("cache coordinator is disabled")
	}

	// Clear L1 cache
	if cc.l1Cache != nil {
		cc.l1Cache.Clear()
	}

	// Clear L2 cache
	if cc.l2Cache != nil {
		if err := cc.l2Cache.Clear(ctx); err != nil {
			if !cc.fallback {
				return fmt.Errorf("L2 cache error: %w", err)
			}

			cc.logger.Warn("L2 cache error during clear, continuing with L1 only",
				zap.Error(err))
		}
	}

	return nil
}

// Exists checks if a key exists in the cache hierarchy
func (cc *CacheCoordinator) Exists(ctx context.Context, key string) (bool, error) {
	if !cc.enabled {
		return false, fmt.Errorf("cache coordinator is disabled")
	}

	// Check L1 cache first
	if cc.l1Cache != nil {
		if _, found := cc.l1Cache.Get(key); found {
			return true, nil
		}
	}

	// Check L2 cache
	if cc.l2Cache != nil {
		exists, err := cc.l2Cache.Exists(ctx, key)
		if err != nil {
			if !cc.fallback {
				return false, fmt.Errorf("L2 cache error: %w", err)
			}

			cc.logger.Warn("L2 cache error during exists check",
				zap.String("key", key),
				zap.Error(err))
			return false, nil
		}

		return exists, nil
	}

	return false, nil
}

// GetMultiple retrieves multiple values from the cache hierarchy
func (cc *CacheCoordinator) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if !cc.enabled {
		return nil, fmt.Errorf("cache coordinator is disabled")
	}

	result := make(map[string]interface{})
	missingKeys := make([]string, 0)

	// Try L1 cache first
	if cc.l1Cache != nil {
		for _, key := range keys {
			if value, found := cc.l1Cache.Get(key); found {
				result[key] = value
				cc.mu.Lock()
				cc.stats.L1Hits++
				cc.stats.TotalHits++
				cc.mu.Unlock()
			} else {
				missingKeys = append(missingKeys, key)
				cc.mu.Lock()
				cc.stats.L1Misses++
				cc.mu.Unlock()
			}
		}
	} else {
		missingKeys = keys
	}

	// Try L2 cache for missing keys
	if cc.l2Cache != nil && len(missingKeys) > 0 {
		l2Results, err := cc.l2Cache.GetMultiple(ctx, missingKeys)
		if err != nil {
			if !cc.fallback {
				return nil, fmt.Errorf("L2 cache error: %w", err)
			}

			cc.logger.Warn("L2 cache error during multiple get",
				zap.Error(err))
		} else {
			// Merge L2 results
			for key, value := range l2Results {
				result[key] = value

				// Promote to L1
				if cc.l1Cache != nil {
					cc.l1Cache.Set(key, value)
				}

				cc.mu.Lock()
				cc.stats.L2Hits++
				cc.stats.TotalHits++
				cc.mu.Unlock()
			}
		}
	}

	// Count misses for keys not found in either cache
	cc.mu.Lock()
	cc.stats.TotalMisses += int64(len(keys) - len(result))
	cc.mu.Unlock()

	return result, nil
}

// SetMultiple stores multiple values in the cache hierarchy
func (cc *CacheCoordinator) SetMultiple(ctx context.Context, items map[string]interface{}) error {
	return cc.SetMultipleWithTTL(ctx, items, 0)
}

// SetMultipleWithTTL stores multiple values in the cache hierarchy with a specific TTL
func (cc *CacheCoordinator) SetMultipleWithTTL(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if !cc.enabled {
		return fmt.Errorf("cache coordinator is disabled")
	}

	// Set in L1 cache
	if cc.l1Cache != nil {
		for key, value := range items {
			if ttl > 0 {
				cc.l1Cache.SetWithTTL(key, value, ttl)
			} else {
				cc.l1Cache.Set(key, value)
			}
		}

		cc.mu.Lock()
		cc.stats.L1Sets += int64(len(items))
		cc.mu.Unlock()
	}

	// Set in L2 cache
	if cc.l2Cache != nil {
		var err error
		if ttl > 0 {
			err = cc.l2Cache.SetMultipleWithTTL(ctx, items, ttl)
		} else {
			err = cc.l2Cache.SetMultiple(ctx, items)
		}

		if err != nil {
			cc.mu.Lock()
			cc.stats.L2Errors++
			cc.mu.Unlock()

			if !cc.fallback {
				return fmt.Errorf("L2 cache error: %w", err)
			}

			cc.logger.Warn("L2 cache error during multiple set",
				zap.Error(err))
		} else {
			cc.mu.Lock()
			cc.stats.L2Sets += int64(len(items))
			cc.mu.Unlock()
		}
	}

	return nil
}

// Stats returns cache coordinator statistics
func (cc *CacheCoordinator) Stats() *CoordinatorStats {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	stats := *cc.stats

	// Calculate hit rate
	total := stats.TotalHits + stats.TotalMisses
	if total > 0 {
		stats.HitRate = float64(stats.TotalHits) / float64(total)
	}

	return &stats
}

// GetL1Stats returns L1 cache statistics
func (cc *CacheCoordinator) GetL1Stats() *MemoryCacheStats {
	if cc.l1Cache != nil {
		return cc.l1Cache.Stats()
	}
	return nil
}

// GetL2Stats returns L2 cache statistics
func (cc *CacheCoordinator) GetL2Stats() *RedisCacheStats {
	if cc.l2Cache != nil {
		return cc.l2Cache.Stats()
	}
	return nil
}

// Health checks the health of both cache levels
func (cc *CacheCoordinator) Health(ctx context.Context) error {
	if !cc.enabled {
		return fmt.Errorf("cache coordinator is disabled")
	}

	var l2Err error

	// Check L1 cache health (always healthy if initialized)
	if cc.l1Cache == nil {
		// L1 cache not initialized - this is not necessarily an error
		// as the coordinator can work with L2 only
	}

	// Check L2 cache health
	if cc.l2Cache != nil {
		l2Err = cc.l2Cache.Health(ctx)
	}

	// Return error if L2 is unhealthy and no L1 fallback
	if l2Err != nil && cc.l1Cache == nil {
		return fmt.Errorf("L2 cache unhealthy and no L1 fallback: %w", l2Err)
	}

	// Return error if L2 is unhealthy and fallback is disabled
	if l2Err != nil && !cc.fallback {
		return fmt.Errorf("L2 cache unhealthy: %w", l2Err)
	}

	return nil
}

// Enable enables the cache coordinator
func (cc *CacheCoordinator) Enable() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.enabled = true
}

// Disable disables the cache coordinator
func (cc *CacheCoordinator) Disable() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.enabled = false
}

// Close closes the cache coordinator and its underlying caches
func (cc *CacheCoordinator) Close() error {
	if cc.l2Cache != nil {
		return cc.l2Cache.Close()
	}
	return nil
}

// Helper methods

func (cc *CacheCoordinator) startSyncRoutine(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		cc.syncCaches()
	}
}

func (cc *CacheCoordinator) syncCaches() {
	// This is a placeholder for cache synchronization logic
	// In a real implementation, you might want to sync certain keys
	// or perform cache warming operations
}
