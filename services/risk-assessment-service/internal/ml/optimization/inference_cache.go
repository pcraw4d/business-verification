package optimization

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// InferenceCache caches ML model inference results
type InferenceCache struct {
	logger  *zap.Logger
	mu      sync.RWMutex
	cache   map[string]*CacheEntry
	stats   *CacheStats
	config  *CacheConfig
	maxSize int
	ttl     time.Duration
}

// CacheEntry represents a cached inference result
type CacheEntry struct {
	Key       string                 `json:"key"`
	Result    map[string]interface{} `json:"result"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt time.Time              `json:"expires_at"`
	HitCount  int64                  `json:"hit_count"`
	LastHit   time.Time              `json:"last_hit"`
	ModelID   string                 `json:"model_id"`
	InputHash string                 `json:"input_hash"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	TotalHits      int64         `json:"total_hits"`
	TotalMisses    int64         `json:"total_misses"`
	HitRate        float64       `json:"hit_rate"`
	CacheSize      int           `json:"cache_size"`
	Evictions      int64         `json:"evictions"`
	AverageHitTime time.Duration `json:"average_hit_time"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	MaxSize           int           `json:"max_size"`
	DefaultTTL        time.Duration `json:"default_ttl"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`
	EnableMetrics     bool          `json:"enable_metrics"`
	EnableCompression bool          `json:"enable_compression"`
}

// InferenceRequest represents a request for model inference
type InferenceRequest struct {
	ModelID   string                 `json:"model_id"`
	Input     map[string]interface{} `json:"input"`
	Options   map[string]interface{} `json:"options"`
	RequestID string                 `json:"request_id"`
}

// InferenceResult represents the result of model inference
type InferenceResult struct {
	RequestID     string                 `json:"request_id"`
	ModelID       string                 `json:"model_id"`
	Result        map[string]interface{} `json:"result"`
	InferenceTime time.Duration          `json:"inference_time"`
	Cached        bool                   `json:"cached"`
	CacheKey      string                 `json:"cache_key"`
	CreatedAt     time.Time              `json:"created_at"`
}

// NewInferenceCache creates a new inference cache
func NewInferenceCache(config *CacheConfig, logger *zap.Logger) *InferenceCache {
	if config == nil {
		config = &CacheConfig{
			MaxSize:           10000,
			DefaultTTL:        1 * time.Hour,
			CleanupInterval:   5 * time.Minute,
			EnableMetrics:     true,
			EnableCompression: false,
		}
	}

	cache := &InferenceCache{
		logger:  logger,
		cache:   make(map[string]*CacheEntry),
		stats:   &CacheStats{},
		config:  config,
		maxSize: config.MaxSize,
		ttl:     config.DefaultTTL,
	}

	// Start cleanup goroutine
	go cache.cleanupLoop()

	return cache
}

// Get retrieves a cached inference result
func (ic *InferenceCache) Get(ctx context.Context, request *InferenceRequest) (*InferenceResult, bool) {
	start := time.Now()

	// Generate cache key
	cacheKey := ic.generateCacheKey(request)

	ic.mu.RLock()
	entry, exists := ic.cache[cacheKey]
	ic.mu.RUnlock()

	if !exists {
		ic.mu.Lock()
		ic.stats.TotalMisses++
		ic.updateHitRate()
		ic.mu.Unlock()

		ic.logger.Debug("Cache miss",
			zap.String("cache_key", cacheKey),
			zap.String("model_id", request.ModelID))

		return nil, false
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		ic.mu.Lock()
		delete(ic.cache, cacheKey)
		ic.stats.TotalMisses++
		ic.updateHitRate()
		ic.mu.Unlock()

		ic.logger.Debug("Cache entry expired",
			zap.String("cache_key", cacheKey),
			zap.String("model_id", request.ModelID))

		return nil, false
	}

	// Update hit statistics
	ic.mu.Lock()
	entry.HitCount++
	entry.LastHit = time.Now()
	ic.stats.TotalHits++
	ic.updateHitRate()
	ic.mu.Unlock()

	hitTime := time.Since(start)

	// Update average hit time
	ic.mu.Lock()
	ic.stats.AverageHitTime = (ic.stats.AverageHitTime + hitTime) / 2
	ic.mu.Unlock()

	ic.logger.Debug("Cache hit",
		zap.String("cache_key", cacheKey),
		zap.String("model_id", request.ModelID),
		zap.Int64("hit_count", entry.HitCount),
		zap.Duration("hit_time", hitTime))

	return &InferenceResult{
		RequestID:     request.RequestID,
		ModelID:       request.ModelID,
		Result:        entry.Result,
		InferenceTime: hitTime,
		Cached:        true,
		CacheKey:      cacheKey,
		CreatedAt:     entry.CreatedAt,
	}, true
}

// Set stores an inference result in the cache
func (ic *InferenceCache) Set(ctx context.Context, request *InferenceRequest, result map[string]interface{}) error {
	// Generate cache key
	cacheKey := ic.generateCacheKey(request)

	// Create cache entry
	entry := &CacheEntry{
		Key:       cacheKey,
		Result:    result,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ic.ttl),
		HitCount:  0,
		LastHit:   time.Now(),
		ModelID:   request.ModelID,
		InputHash: ic.generateInputHash(request.Input),
	}

	ic.mu.Lock()
	defer ic.mu.Unlock()

	// Check if cache is full
	if len(ic.cache) >= ic.maxSize {
		ic.evictOldest()
	}

	// Store entry
	ic.cache[cacheKey] = entry
	ic.stats.CacheSize = len(ic.cache)

	ic.logger.Debug("Cache entry stored",
		zap.String("cache_key", cacheKey),
		zap.String("model_id", request.ModelID),
		zap.Time("expires_at", entry.ExpiresAt))

	return nil
}

// Invalidate invalidates cache entries for a specific model
func (ic *InferenceCache) Invalidate(ctx context.Context, modelID string) error {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	var invalidated int
	for key, entry := range ic.cache {
		if entry.ModelID == modelID {
			delete(ic.cache, key)
			invalidated++
		}
	}

	ic.stats.CacheSize = len(ic.cache)

	ic.logger.Info("Cache invalidated for model",
		zap.String("model_id", modelID),
		zap.Int("invalidated_entries", invalidated))

	return nil
}

// Clear clears all cache entries
func (ic *InferenceCache) Clear(ctx context.Context) error {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	ic.cache = make(map[string]*CacheEntry)
	ic.stats.CacheSize = 0

	ic.logger.Info("Cache cleared")

	return nil
}

// GetStats returns cache statistics
func (ic *InferenceCache) GetStats() *CacheStats {
	ic.mu.RLock()
	defer ic.mu.RUnlock()

	stats := *ic.stats
	stats.CacheSize = len(ic.cache)
	return &stats
}

// Helper methods

func (ic *InferenceCache) generateCacheKey(request *InferenceRequest) string {
	// Create a hash of the model ID and input data
	data := map[string]interface{}{
		"model_id": request.ModelID,
		"input":    request.Input,
		"options":  request.Options,
	}

	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)

	return fmt.Sprintf("inference_%x", hash)
}

func (ic *InferenceCache) generateInputHash(input map[string]interface{}) string {
	jsonData, _ := json.Marshal(input)
	hash := sha256.Sum256(jsonData)
	return fmt.Sprintf("%x", hash)
}

func (ic *InferenceCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range ic.cache {
		if oldestKey == "" || entry.LastHit.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastHit
		}
	}

	if oldestKey != "" {
		delete(ic.cache, oldestKey)
		ic.stats.Evictions++

		ic.logger.Debug("Cache entry evicted",
			zap.String("cache_key", oldestKey))
	}
}

func (ic *InferenceCache) updateHitRate() {
	total := ic.stats.TotalHits + ic.stats.TotalMisses
	if total > 0 {
		ic.stats.HitRate = float64(ic.stats.TotalHits) / float64(total)
	}
}

func (ic *InferenceCache) cleanupLoop() {
	ticker := time.NewTicker(ic.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		ic.cleanupExpired()
	}
}

func (ic *InferenceCache) cleanupExpired() {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	now := time.Now()
	var expired []string

	for key, entry := range ic.cache {
		if now.After(entry.ExpiresAt) {
			expired = append(expired, key)
		}
	}

	for _, key := range expired {
		delete(ic.cache, key)
	}

	if len(expired) > 0 {
		ic.stats.CacheSize = len(ic.cache)
		ic.logger.Debug("Expired cache entries cleaned up",
			zap.Int("expired_count", len(expired)))
	}
}

// CacheMetrics represents cache performance metrics
type CacheMetrics struct {
	HitRate        float64       `json:"hit_rate"`
	MissRate       float64       `json:"miss_rate"`
	AverageHitTime time.Duration `json:"average_hit_time"`
	CacheSize      int           `json:"cache_size"`
	EvictionRate   float64       `json:"eviction_rate"`
	MemoryUsage    int64         `json:"memory_usage"`
}

// GetMetrics returns detailed cache metrics
func (ic *InferenceCache) GetMetrics() *CacheMetrics {
	stats := ic.GetStats()

	metrics := &CacheMetrics{
		HitRate:        stats.HitRate,
		MissRate:       1.0 - stats.HitRate,
		AverageHitTime: stats.AverageHitTime,
		CacheSize:      stats.CacheSize,
		EvictionRate:   float64(stats.Evictions) / float64(stats.TotalHits+stats.TotalMisses),
		MemoryUsage:    int64(stats.CacheSize * 1024), // Rough estimate
	}

	return metrics
}

// PrefetchResult prefetches common inference results
func (ic *InferenceCache) PrefetchResult(ctx context.Context, modelID string, commonInputs []map[string]interface{}) error {
	ic.logger.Info("Starting cache prefetch",
		zap.String("model_id", modelID),
		zap.Int("input_count", len(commonInputs)))

	for i, input := range commonInputs {
		request := &InferenceRequest{
			ModelID:   modelID,
			Input:     input,
			Options:   make(map[string]interface{}),
			RequestID: fmt.Sprintf("prefetch_%d", i),
		}

		// Check if already cached
		if _, exists := ic.Get(ctx, request); exists {
			continue
		}

		// Simulate inference result
		result := map[string]interface{}{
			"risk_score": 0.75,
			"risk_level": "medium",
			"factors":    []string{"industry_risk", "country_risk"},
		}

		// Cache the result
		if err := ic.Set(ctx, request, result); err != nil {
			ic.logger.Warn("Failed to prefetch result",
				zap.String("model_id", modelID),
				zap.Int("input_index", i),
				zap.Error(err))
		}
	}

	ic.logger.Info("Cache prefetch completed",
		zap.String("model_id", modelID))

	return nil
}
