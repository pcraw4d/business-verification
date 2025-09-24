package cache

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CacheInvalidationManager handles cache invalidation strategies
type CacheInvalidationManager struct {
	// Configuration
	config CacheConfig

	// Cache layers for invalidation
	caches []Cache

	// Invalidation patterns and rules
	invalidationRules []InvalidationRule

	// Thread safety
	mu sync.RWMutex

	// Invalidation statistics
	stats     *InvalidationStats
	statsLock sync.RWMutex

	// Logging
	logger *zap.Logger
}

// InvalidationRule defines a cache invalidation rule
type InvalidationRule struct {
	Pattern     string        `json:"pattern"`     // Regex pattern to match keys
	TTL         time.Duration `json:"ttl"`         // Time to live for matched keys
	Priority    int           `json:"priority"`    // Rule priority (higher = more important)
	Description string        `json:"description"` // Rule description
}

// InvalidationStats holds invalidation statistics
type InvalidationStats struct {
	InvalidationsByPattern  int64         `json:"invalidations_by_pattern"`
	InvalidationsByTags     int64         `json:"invalidations_by_tags"`
	InvalidationsByTTL      int64         `json:"invalidations_by_ttl"`
	TotalInvalidations      int64         `json:"total_invalidations"`
	LastInvalidation        time.Time     `json:"last_invalidation"`
	AverageInvalidationTime time.Duration `json:"average_invalidation_time"`
}

// NewCacheInvalidationManager creates a new cache invalidation manager
func NewCacheInvalidationManager(config CacheConfig, logger *zap.Logger) *CacheInvalidationManager {
	return &CacheInvalidationManager{
		config:            config,
		invalidationRules: make([]InvalidationRule, 0),
		stats:             &InvalidationStats{},
		logger:            logger,
	}
}

// SetCaches sets the cache layers for invalidation
func (cim *CacheInvalidationManager) SetCaches(caches ...Cache) {
	cim.mu.Lock()
	defer cim.mu.Unlock()

	cim.caches = caches
}

// AddInvalidationRule adds a new invalidation rule
func (cim *CacheInvalidationManager) AddInvalidationRule(rule InvalidationRule) error {
	cim.mu.Lock()
	defer cim.mu.Unlock()

	// Validate pattern
	if rule.Pattern != "" {
		if _, err := regexp.Compile(rule.Pattern); err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}
	}

	// Add rule
	cim.invalidationRules = append(cim.invalidationRules, rule)

	// Sort by priority (higher priority first)
	cim.sortInvalidationRules()

	cim.logger.Info("Added invalidation rule",
		zap.String("pattern", rule.Pattern),
		zap.Int("priority", rule.Priority),
		zap.String("description", rule.Description))

	return nil
}

// RemoveInvalidationRule removes an invalidation rule by pattern
func (cim *CacheInvalidationManager) RemoveInvalidationRule(pattern string) {
	cim.mu.Lock()
	defer cim.mu.Unlock()

	for i, rule := range cim.invalidationRules {
		if rule.Pattern == pattern {
			cim.invalidationRules = append(cim.invalidationRules[:i], cim.invalidationRules[i+1:]...)
			cim.logger.Info("Removed invalidation rule", zap.String("pattern", pattern))
			break
		}
	}
}

// InvalidateByPattern invalidates cache entries matching a pattern
func (cim *CacheInvalidationManager) InvalidateByPattern(ctx context.Context, pattern string) error {
	start := time.Now()

	cim.mu.RLock()
	defer cim.mu.RUnlock()

	// Compile pattern
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	var invalidatedCount int64

	// Invalidate from all cache layers
	for _, cache := range cim.caches {
		keys, err := cache.GetKeys(ctx, pattern)
		if err == nil {
			for _, key := range keys {
				if regex.MatchString(key) {
					if err := cache.Delete(ctx, key); err == nil {
						invalidatedCount++
					}
				}
			}
		}
	}

	// Update statistics
	cim.updateInvalidationStats("pattern", invalidatedCount, time.Since(start))

	cim.logger.Info("Invalidated cache entries by pattern",
		zap.String("pattern", pattern),
		zap.Int64("count", invalidatedCount))

	return nil
}

// InvalidateByTags invalidates cache entries with specific tags
func (cim *CacheInvalidationManager) InvalidateByTags(ctx context.Context, tags []string) error {
	start := time.Now()

	cim.mu.RLock()
	defer cim.mu.RUnlock()

	var invalidatedCount int64

	// Create tag pattern
	tagPattern := strings.Join(tags, "|")
	regex, err := regexp.Compile(tagPattern)
	if err != nil {
		return fmt.Errorf("invalid tag pattern: %w", err)
	}

	// Invalidate from all cache layers
	for _, cache := range cim.caches {
		keys, err := cache.GetKeys(ctx, "*")
		if err == nil {
			for _, key := range keys {
				// Check if key contains any of the tags
				if regex.MatchString(key) {
					if err := cache.Delete(ctx, key); err == nil {
						invalidatedCount++
					}
				}
			}
		}
	}

	// Update statistics
	cim.updateInvalidationStats("tags", invalidatedCount, time.Since(start))

	cim.logger.Info("Invalidated cache entries by tags",
		zap.Strings("tags", tags),
		zap.Int64("count", invalidatedCount))

	return nil
}

// InvalidateByTTL invalidates cache entries that have expired
func (cim *CacheInvalidationManager) InvalidateByTTL(ctx context.Context) error {
	start := time.Now()

	cim.mu.RLock()
	defer cim.mu.RUnlock()

	var invalidatedCount int64

	// TTL invalidation is handled automatically by cache implementations
	// This method is mainly for statistics and monitoring
	_ = cim.caches // Acknowledge that we're checking caches but not using them
	cim.logger.Debug("TTL invalidation handled automatically by cache implementation")

	// Update statistics
	cim.updateInvalidationStats("ttl", invalidatedCount, time.Since(start))

	cim.logger.Debug("TTL-based invalidation completed",
		zap.Int64("count", invalidatedCount))

	return nil
}

// InvalidateAll invalidates all cache entries
func (cim *CacheInvalidationManager) InvalidateAll(ctx context.Context) error {
	start := time.Now()

	cim.mu.RLock()
	defer cim.mu.RUnlock()

	var invalidatedCount int64

	// Invalidate from all cache layers
	for _, cache := range cim.caches {
		if err := cache.Clear(ctx); err == nil {
			invalidatedCount++
		}
	}

	// Update statistics
	cim.updateInvalidationStats("all", invalidatedCount, time.Since(start))

	cim.logger.Info("Invalidated all cache entries",
		zap.Int64("layers_cleared", invalidatedCount))

	return nil
}

// GetInvalidationRules returns all invalidation rules
func (cim *CacheInvalidationManager) GetInvalidationRules() []InvalidationRule {
	cim.mu.RLock()
	defer cim.mu.RUnlock()

	rules := make([]InvalidationRule, len(cim.invalidationRules))
	copy(rules, cim.invalidationRules)
	return rules
}

// GetInvalidationStats returns invalidation statistics
func (cim *CacheInvalidationManager) GetInvalidationStats() *InvalidationStats {
	cim.statsLock.RLock()
	defer cim.statsLock.RUnlock()

	stats := *cim.stats
	return &stats
}

// ResetInvalidationStats resets invalidation statistics
func (cim *CacheInvalidationManager) ResetInvalidationStats() {
	cim.statsLock.Lock()
	defer cim.statsLock.Unlock()

	cim.stats = &InvalidationStats{}
}

// Helper methods

func (cim *CacheInvalidationManager) sortInvalidationRules() {
	// Sort rules by priority (higher priority first)
	// This is a simple bubble sort - in production you might want something more efficient
	for i := 0; i < len(cim.invalidationRules)-1; i++ {
		for j := i + 1; j < len(cim.invalidationRules); j++ {
			if cim.invalidationRules[i].Priority < cim.invalidationRules[j].Priority {
				cim.invalidationRules[i], cim.invalidationRules[j] = cim.invalidationRules[j], cim.invalidationRules[i]
			}
		}
	}
}

func (cim *CacheInvalidationManager) updateInvalidationStats(method string, count int64, duration time.Duration) {
	cim.statsLock.Lock()
	defer cim.statsLock.Unlock()

	cim.stats.TotalInvalidations += count
	cim.stats.LastInvalidation = time.Now()

	switch method {
	case "pattern":
		cim.stats.InvalidationsByPattern += count
	case "tags":
		cim.stats.InvalidationsByTags += count
	case "ttl":
		cim.stats.InvalidationsByTTL += count
	}

	// Update average invalidation time
	if cim.stats.TotalInvalidations > 0 {
		totalTime := cim.stats.AverageInvalidationTime * time.Duration(cim.stats.TotalInvalidations-count)
		newAverage := (totalTime + duration) / time.Duration(cim.stats.TotalInvalidations)
		cim.stats.AverageInvalidationTime = newAverage
	}
}
