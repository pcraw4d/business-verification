package infrastructure

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// RuleEngineCache provides high-performance caching for rule engine results
type RuleEngineCache struct {
	// Cache storage
	classificationCache map[string]*CachedClassificationResult
	riskCache           map[string]*CachedRiskResult

	// Cache configuration
	config CacheConfig

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger

	// Control
	ctx    context.Context
	cancel context.CancelFunc
}

// CachedClassificationResult represents a cached classification result
type CachedClassificationResult struct {
	Result      *RuleEngineClassificationResponse `json:"result"`
	CachedAt    time.Time                         `json:"cached_at"`
	ExpiresAt   time.Time                         `json:"expires_at"`
	AccessCount int                               `json:"access_count"`
}

// CachedRiskResult represents a cached risk detection result
type CachedRiskResult struct {
	Result      *RuleEngineRiskResponse `json:"result"`
	CachedAt    time.Time               `json:"cached_at"`
	ExpiresAt   time.Time               `json:"expires_at"`
	AccessCount int                     `json:"access_count"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Enabled         bool          `json:"enabled"`
	MaxSize         int           `json:"max_size"`
	DefaultTTL      time.Duration `json:"default_ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
	MaxAccessCount  int           `json:"max_access_count"`
}

// NewRuleEngineCache creates a new rule engine cache
func NewRuleEngineCache(logger *log.Logger) *RuleEngineCache {
	if logger == nil {
		logger = log.Default()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &RuleEngineCache{
		classificationCache: make(map[string]*CachedClassificationResult),
		riskCache:           make(map[string]*CachedRiskResult),
		config: CacheConfig{
			Enabled:         true,
			MaxSize:         1000,
			DefaultTTL:      1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			MaxAccessCount:  100,
		},
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Initialize initializes the rule engine cache
func (rec *RuleEngineCache) Initialize(ctx context.Context) error {
	rec.mu.Lock()
	defer rec.mu.Unlock()

	rec.logger.Printf("💾 Initializing Rule Engine Cache")

	// Initialize cache storage
	rec.classificationCache = make(map[string]*CachedClassificationResult)
	rec.riskCache = make(map[string]*CachedRiskResult)

	rec.logger.Printf("✅ Rule Engine Cache initialized with max size %d", rec.config.MaxSize)
	return nil
}

// StartCleanup starts the cache cleanup process
func (rec *RuleEngineCache) StartCleanup(ctx context.Context) {
	rec.logger.Printf("🧹 Starting cache cleanup process")

	ticker := time.NewTicker(rec.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rec.cleanupExpiredEntries()
		}
	}
}

// GetClassification retrieves a cached classification result
func (rec *RuleEngineCache) GetClassification(key string) (*CachedClassificationResult, bool) {
	rec.mu.RLock()
	defer rec.mu.RUnlock()

	cached, exists := rec.classificationCache[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(cached.ExpiresAt) {
		return nil, false
	}

	// Update access count
	cached.AccessCount++

	return cached, true
}

// SetClassification stores a classification result in cache
func (rec *RuleEngineCache) SetClassification(key string, result *RuleEngineClassificationResponse, ttl time.Duration) {
	rec.mu.Lock()
	defer rec.mu.Unlock()

	// Check cache size limit
	if len(rec.classificationCache) >= rec.config.MaxSize {
		rec.evictLeastUsedClassification()
	}

	// Set TTL
	if ttl == 0 {
		ttl = rec.config.DefaultTTL
	}

	cached := &CachedClassificationResult{
		Result:      result,
		CachedAt:    time.Now(),
		ExpiresAt:   time.Now().Add(ttl),
		AccessCount: 0,
	}

	rec.classificationCache[key] = cached
}

// GetRisk retrieves a cached risk detection result
func (rec *RuleEngineCache) GetRisk(key string) (*CachedRiskResult, bool) {
	rec.mu.RLock()
	defer rec.mu.RUnlock()

	cached, exists := rec.riskCache[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(cached.ExpiresAt) {
		return nil, false
	}

	// Update access count
	cached.AccessCount++

	return cached, true
}

// SetRisk stores a risk detection result in cache
func (rec *RuleEngineCache) SetRisk(key string, result *RuleEngineRiskResponse, ttl time.Duration) {
	rec.mu.Lock()
	defer rec.mu.Unlock()

	// Check cache size limit
	if len(rec.riskCache) >= rec.config.MaxSize {
		rec.evictLeastUsedRisk()
	}

	// Set TTL
	if ttl == 0 {
		ttl = rec.config.DefaultTTL
	}

	cached := &CachedRiskResult{
		Result:      result,
		CachedAt:    time.Now(),
		ExpiresAt:   time.Now().Add(ttl),
		AccessCount: 0,
	}

	rec.riskCache[key] = cached
}

// Clear clears all cache entries
func (rec *RuleEngineCache) Clear() {
	rec.mu.Lock()
	defer rec.mu.Unlock()

	rec.classificationCache = make(map[string]*CachedClassificationResult)
	rec.riskCache = make(map[string]*CachedRiskResult)

	rec.logger.Printf("🗑️ Cache cleared")
}

// GetStats returns cache statistics
func (rec *RuleEngineCache) GetStats() CacheStats {
	rec.mu.RLock()
	defer rec.mu.RUnlock()

	stats := CacheStats{
		ClassificationEntries: len(rec.classificationCache),
		RiskEntries:           len(rec.riskCache),
		TotalEntries:          len(rec.classificationCache) + len(rec.riskCache),
		MaxSize:               rec.config.MaxSize,
		HitRate:               0.0, // Would need to track hits/misses
	}

	return stats
}

// HealthCheck performs a health check on the cache
func (rec *RuleEngineCache) HealthCheck(ctx context.Context) error {
	rec.mu.RLock()
	defer rec.mu.RUnlock()

	// Check if cache is enabled
	if !rec.config.Enabled {
		return fmt.Errorf("cache is disabled")
	}

	// Check if cache is not too full
	totalEntries := len(rec.classificationCache) + len(rec.riskCache)
	if totalEntries > rec.config.MaxSize*9/10 { // 90% full
		return fmt.Errorf("cache is nearly full (%d/%d entries)", totalEntries, rec.config.MaxSize)
	}

	return nil
}

// cleanupExpiredEntries removes expired entries from cache
func (rec *RuleEngineCache) cleanupExpiredEntries() {
	rec.mu.Lock()
	defer rec.mu.Unlock()

	now := time.Now()
	cleanedCount := 0

	// Clean classification cache
	for key, cached := range rec.classificationCache {
		if now.After(cached.ExpiresAt) {
			delete(rec.classificationCache, key)
			cleanedCount++
		}
	}

	// Clean risk cache
	for key, cached := range rec.riskCache {
		if now.After(cached.ExpiresAt) {
			delete(rec.riskCache, key)
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		rec.logger.Printf("🧹 Cleaned %d expired cache entries", cleanedCount)
	}
}

// evictLeastUsedClassification evicts the least used classification entry
func (rec *RuleEngineCache) evictLeastUsedClassification() {
	if len(rec.classificationCache) == 0 {
		return
	}

	var leastUsedKey string
	var leastUsedCount int = -1

	for key, cached := range rec.classificationCache {
		if leastUsedCount == -1 || cached.AccessCount < leastUsedCount {
			leastUsedKey = key
			leastUsedCount = cached.AccessCount
		}
	}

	if leastUsedKey != "" {
		delete(rec.classificationCache, leastUsedKey)
		rec.logger.Printf("🗑️ Evicted least used classification entry: %s", leastUsedKey)
	}
}

// evictLeastUsedRisk evicts the least used risk entry
func (rec *RuleEngineCache) evictLeastUsedRisk() {
	if len(rec.riskCache) == 0 {
		return
	}

	var leastUsedKey string
	var leastUsedCount int = -1

	for key, cached := range rec.riskCache {
		if leastUsedCount == -1 || cached.AccessCount < leastUsedCount {
			leastUsedKey = key
			leastUsedCount = cached.AccessCount
		}
	}

	if leastUsedKey != "" {
		delete(rec.riskCache, leastUsedKey)
		rec.logger.Printf("🗑️ Evicted least used risk entry: %s", leastUsedKey)
	}
}

// CacheStats represents cache statistics
type CacheStats struct {
	ClassificationEntries int     `json:"classification_entries"`
	RiskEntries           int     `json:"risk_entries"`
	TotalEntries          int     `json:"total_entries"`
	MaxSize               int     `json:"max_size"`
	HitRate               float64 `json:"hit_rate"`
}
