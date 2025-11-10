// Package cache provides intelligent query caching for the KYB Platform
// This module implements a comprehensive caching strategy for frequently accessed database queries
// to significantly improve application performance and reduce database load.

package cache

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// QueryCacheManager manages intelligent query caching for the KYB Platform
type QueryCacheManager struct {
	redisClient *redis.Client
	localCache  *LocalCache
	config      *QueryCacheConfig
	metrics     *QueryCacheMetrics
	mu          sync.RWMutex
}

// CacheConfig defines the configuration for query caching
// QueryCacheConfig holds configuration for query cache
// (renamed to avoid conflict with types.go CacheConfig)
type QueryCacheConfig struct {
	// Redis configuration
	RedisHost     string `json:"redis_host"`
	RedisPort     int    `json:"redis_port"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`

	// Cache TTL settings
	DefaultTTL        time.Duration `json:"default_ttl"`
	ClassificationTTL time.Duration `json:"classification_ttl"`
	RiskAssessmentTTL time.Duration `json:"risk_assessment_ttl"`
	UserDataTTL       time.Duration `json:"user_data_ttl"`
	BusinessDataTTL   time.Duration `json:"business_data_ttl"`

	// Local cache settings
	LocalCacheSize int           `json:"local_cache_size"`
	LocalCacheTTL  time.Duration `json:"local_cache_ttl"`

	// Cache invalidation settings
	EnableInvalidation bool          `json:"enable_invalidation"`
	InvalidationDelay  time.Duration `json:"invalidation_delay"`

	// Performance settings
	EnableCompression bool `json:"enable_compression"`
	EnableMetrics     bool `json:"enable_metrics"`
}

// CacheMetrics tracks cache performance metrics
// QueryCacheMetrics tracks metrics for query cache
// (renamed to avoid conflict with intelligent_cache.go CacheMetrics)
type QueryCacheMetrics struct {
	Hits           int64         `json:"hits"`
	Misses         int64         `json:"misses"`
	Sets           int64         `json:"sets"`
	Deletes        int64         `json:"deletes"`
	Errors         int64         `json:"errors"`
	HitRate        float64       `json:"hit_rate"`
	AverageGetTime time.Duration `json:"average_get_time"`
	AverageSetTime time.Duration `json:"average_set_time"`
	mu             sync.RWMutex
}

// CacheKey represents a cache key with metadata
type CacheKey struct {
	QueryType string                 `json:"query_type"`
	Params    map[string]interface{} `json:"params"`
	Version   string                 `json:"version"`
	TTL       time.Duration          `json:"ttl"`
}

// CacheResult represents a cached query result
type CacheResult struct {
	Data      interface{}   `json:"data"`
	Timestamp time.Time     `json:"timestamp"`
	TTL       time.Duration `json:"ttl"`
	HitCount  int64         `json:"hit_count"`
}

// LocalCache provides in-memory caching for frequently accessed data
type LocalCache struct {
	data    map[string]*CacheResult
	mu      sync.RWMutex
	maxSize int
	ttl     time.Duration
}

// NewQueryCacheManager creates a new query cache manager
func NewQueryCacheManager(config *QueryCacheConfig) (*QueryCacheManager, error) {
	if config == nil {
		config = getDefaultQueryCacheConfig()
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.RedisHost, config.RedisPort),
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed, falling back to local cache only: %v", err)
		redisClient = nil
	}

	// Initialize local cache
	localCache := &LocalCache{
		data:    make(map[string]*CacheResult),
		maxSize: config.LocalCacheSize,
		ttl:     config.LocalCacheTTL,
	}

	// Initialize metrics
	metrics := &QueryCacheMetrics{}

	manager := &QueryCacheManager{
		redisClient: redisClient,
		localCache:  localCache,
		config:      config,
		metrics:     metrics,
	}

	// Start background cleanup for local cache
	go manager.startLocalCacheCleanup()

	// Start metrics collection if enabled
	if config.EnableMetrics {
		go manager.startMetricsCollection()
	}

	return manager, nil
}

// Get retrieves a value from cache
func (qcm *QueryCacheManager) Get(ctx context.Context, key *CacheKey) (interface{}, bool, error) {
	start := time.Now()
	defer func() {
		qcm.updateMetrics("get", time.Since(start), false)
	}()

	cacheKey := qcm.generateCacheKey(key)

	// Try local cache first
	if result, found := qcm.getFromLocalCache(cacheKey); found {
		qcm.updateMetrics("get", time.Since(start), true)
		return result.Data, true, nil
	}

	// Try Redis cache if available
	if qcm.redisClient != nil {
		if result, found, err := qcm.getFromRedisCache(ctx, cacheKey); err == nil && found {
			// Store in local cache for faster subsequent access
			qcm.setInLocalCache(cacheKey, result)
			qcm.updateMetrics("get", time.Since(start), true)
			return result.Data, true, nil
		}
	}

	qcm.updateMetrics("get", time.Since(start), false)
	return nil, false, nil
}

// Set stores a value in cache
func (qcm *QueryCacheManager) Set(ctx context.Context, key *CacheKey, value interface{}) error {
	start := time.Now()
	defer func() {
		qcm.updateMetrics("set", time.Since(start), false)
	}()

	cacheKey := qcm.generateCacheKey(key)
	ttl := qcm.getTTL(key)

	// Create cache result
	result := &CacheResult{
		Data:      value,
		Timestamp: time.Now(),
		TTL:       ttl,
		HitCount:  0,
	}

	// Store in local cache
	qcm.setInLocalCache(cacheKey, result)

	// Store in Redis cache if available
	if qcm.redisClient != nil {
		if err := qcm.setInRedisCache(ctx, cacheKey, result, ttl); err != nil {
			log.Printf("Warning: Failed to set Redis cache: %v", err)
		}
	}

	qcm.updateMetrics("set", time.Since(start), false)
	return nil
}

// Delete removes a value from cache
func (qcm *QueryCacheManager) Delete(ctx context.Context, key *CacheKey) error {
	start := time.Now()
	defer func() {
		qcm.updateMetrics("delete", time.Since(start), false)
	}()

	cacheKey := qcm.generateCacheKey(key)

	// Remove from local cache
	qcm.deleteFromLocalCache(cacheKey)

	// Remove from Redis cache if available
	if qcm.redisClient != nil {
		if err := qcm.redisClient.Del(ctx, cacheKey).Err(); err != nil {
			log.Printf("Warning: Failed to delete from Redis cache: %v", err)
		}
	}

	qcm.updateMetrics("delete", time.Since(start), false)
	return nil
}

// InvalidateByPattern invalidates cache entries matching a pattern
func (qcm *QueryCacheManager) InvalidateByPattern(ctx context.Context, pattern string) error {
	if !qcm.config.EnableInvalidation {
		return nil
	}

	// Invalidate local cache entries
	qcm.invalidateLocalCacheByPattern(pattern)

	// Invalidate Redis cache entries if available
	if qcm.redisClient != nil {
		keys, err := qcm.redisClient.Keys(ctx, pattern).Result()
		if err != nil {
			return fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
		}

		if len(keys) > 0 {
			if err := qcm.redisClient.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("failed to delete keys for pattern %s: %w", pattern, err)
			}
		}
	}

	return nil
}

// GetMetrics returns current cache metrics
func (qcm *QueryCacheManager) GetMetrics() *QueryCacheMetrics {
	qcm.metrics.mu.RLock()
	defer qcm.metrics.mu.RUnlock()

	// Calculate hit rate
	total := qcm.metrics.Hits + qcm.metrics.Misses
	if total > 0 {
		qcm.metrics.HitRate = float64(qcm.metrics.Hits) / float64(total) * 100
	}

	return &QueryCacheMetrics{
		Hits:           qcm.metrics.Hits,
		Misses:         qcm.metrics.Misses,
		Sets:           qcm.metrics.Sets,
		Deletes:        qcm.metrics.Deletes,
		Errors:         qcm.metrics.Errors,
		HitRate:        qcm.metrics.HitRate,
		AverageGetTime: qcm.metrics.AverageGetTime,
		AverageSetTime: qcm.metrics.AverageSetTime,
	}
}

// getFromLocalCache retrieves a value from local cache
func (qcm *QueryCacheManager) getFromLocalCache(key string) (*CacheResult, bool) {
	qcm.localCache.mu.RLock()
	defer qcm.localCache.mu.RUnlock()

	result, exists := qcm.localCache.data[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Since(result.Timestamp) > result.TTL {
		return nil, false
	}

	// Update hit count
	result.HitCount++
	return result, true
}

// setInLocalCache stores a value in local cache
func (qcm *QueryCacheManager) setInLocalCache(key string, result *CacheResult) {
	qcm.localCache.mu.Lock()
	defer qcm.localCache.mu.Unlock()

	// Check if we need to evict entries
	if len(qcm.localCache.data) >= qcm.localCache.maxSize {
		qcm.evictOldestLocalCacheEntry()
	}

	qcm.localCache.data[key] = result
}

// deleteFromLocalCache removes a value from local cache
func (qcm *QueryCacheManager) deleteFromLocalCache(key string) {
	qcm.localCache.mu.Lock()
	defer qcm.localCache.mu.Unlock()

	delete(qcm.localCache.data, key)
}

// getFromRedisCache retrieves a value from Redis cache
func (qcm *QueryCacheManager) getFromRedisCache(ctx context.Context, key string) (*CacheResult, bool, error) {
	data, err := qcm.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, err
	}

	var result CacheResult
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, false, err
	}

	return &result, true, nil
}

// setInRedisCache stores a value in Redis cache
func (qcm *QueryCacheManager) setInRedisCache(ctx context.Context, key string, result *CacheResult, ttl time.Duration) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return qcm.redisClient.Set(ctx, key, data, ttl).Err()
}

// generateCacheKey generates a unique cache key
func (qcm *QueryCacheManager) generateCacheKey(key *CacheKey) string {
	// Create a deterministic key from query type and parameters
	keyData := fmt.Sprintf("%s:%s:%v", key.QueryType, key.Version, key.Params)
	hash := md5.Sum([]byte(keyData))
	return fmt.Sprintf("kyb:cache:%s:%x", key.QueryType, hash)
}

// getTTL returns the appropriate TTL for a cache key
func (qcm *QueryCacheManager) getTTL(key *CacheKey) time.Duration {
	if key.TTL > 0 {
		return key.TTL
	}

	switch key.QueryType {
	case "classification":
		return qcm.config.ClassificationTTL
	case "risk_assessment":
		return qcm.config.RiskAssessmentTTL
	case "user_data":
		return qcm.config.UserDataTTL
	case "business_data":
		return qcm.config.BusinessDataTTL
	default:
		return qcm.config.DefaultTTL
	}
}

// updateMetrics updates cache performance metrics
func (qcm *QueryCacheManager) updateMetrics(operation string, duration time.Duration, hit bool) {
	if !qcm.config.EnableMetrics {
		return
	}

	qcm.metrics.mu.Lock()
	defer qcm.metrics.mu.Unlock()

	switch operation {
	case "get":
		if hit {
			qcm.metrics.Hits++
		} else {
			qcm.metrics.Misses++
		}
		// Update average get time
		totalGets := qcm.metrics.Hits + qcm.metrics.Misses
		qcm.metrics.AverageGetTime = (qcm.metrics.AverageGetTime*time.Duration(totalGets-1) + duration) / time.Duration(totalGets)
	case "set":
		qcm.metrics.Sets++
		// Update average set time
		qcm.metrics.AverageSetTime = (qcm.metrics.AverageSetTime*time.Duration(qcm.metrics.Sets-1) + duration) / time.Duration(qcm.metrics.Sets)
	case "delete":
		qcm.metrics.Deletes++
	}
}

// startLocalCacheCleanup starts background cleanup for local cache
func (qcm *QueryCacheManager) startLocalCacheCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		qcm.cleanupLocalCache()
	}
}

// cleanupLocalCache removes expired entries from local cache
func (qcm *QueryCacheManager) cleanupLocalCache() {
	qcm.localCache.mu.Lock()
	defer qcm.localCache.mu.Unlock()

	now := time.Now()
	for key, result := range qcm.localCache.data {
		if now.Sub(result.Timestamp) > result.TTL {
			delete(qcm.localCache.data, key)
		}
	}
}

// evictOldestLocalCacheEntry evicts the oldest entry from local cache
func (qcm *QueryCacheManager) evictOldestLocalCacheEntry() {
	var oldestKey string
	var oldestTime time.Time

	for key, result := range qcm.localCache.data {
		if oldestKey == "" || result.Timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = result.Timestamp
		}
	}

	if oldestKey != "" {
		delete(qcm.localCache.data, oldestKey)
	}
}

// invalidateLocalCacheByPattern invalidates local cache entries matching a pattern
func (qcm *QueryCacheManager) invalidateLocalCacheByPattern(pattern string) {
	qcm.localCache.mu.Lock()
	defer qcm.localCache.mu.Unlock()

	for key := range qcm.localCache.data {
		if matched, _ := matchPattern(key, pattern); matched {
			delete(qcm.localCache.data, key)
		}
	}
}

// startMetricsCollection starts background metrics collection
func (qcm *QueryCacheManager) startMetricsCollection() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		metrics := qcm.GetMetrics()
		log.Printf("Cache Metrics - Hits: %d, Misses: %d, Hit Rate: %.2f%%, Avg Get Time: %v, Avg Set Time: %v",
			metrics.Hits, metrics.Misses, metrics.HitRate, metrics.AverageGetTime, metrics.AverageSetTime)
	}
}

// getDefaultQueryCacheConfig returns default cache configuration
func getDefaultQueryCacheConfig() *QueryCacheConfig {
	return &QueryCacheConfig{
		RedisHost:          "localhost",
		RedisPort:          6379,
		RedisPassword:      "",
		RedisDB:            0,
		DefaultTTL:         15 * time.Minute,
		ClassificationTTL:  30 * time.Minute,
		RiskAssessmentTTL:  1 * time.Hour,
		UserDataTTL:        2 * time.Hour,
		BusinessDataTTL:    1 * time.Hour,
		LocalCacheSize:     1000,
		LocalCacheTTL:      5 * time.Minute,
		EnableInvalidation: true,
		InvalidationDelay:  1 * time.Second,
		EnableCompression:  false,
		EnableMetrics:      true,
	}
}

// matchPattern checks if a key matches a pattern (simple wildcard matching)
func matchPattern(key, pattern string) (bool, error) {
	// Simple wildcard matching implementation
	// This could be enhanced with more sophisticated pattern matching
	if pattern == "*" {
		return true, nil
	}

	// Check if key contains the pattern
	return len(key) >= len(pattern) && key[:len(pattern)] == pattern, nil
}

// CacheKeyBuilder helps build cache keys for different query types
type CacheKeyBuilder struct {
	queryType string
	params    map[string]interface{}
	version   string
	ttl       time.Duration
}

// NewCacheKeyBuilder creates a new cache key builder
func NewCacheKeyBuilder(queryType string) *CacheKeyBuilder {
	return &CacheKeyBuilder{
		queryType: queryType,
		params:    make(map[string]interface{}),
		version:   "v1",
	}
}

// AddParam adds a parameter to the cache key
func (b *CacheKeyBuilder) AddParam(key string, value interface{}) *CacheKeyBuilder {
	b.params[key] = value
	return b
}

// SetVersion sets the version for the cache key
func (b *CacheKeyBuilder) SetVersion(version string) *CacheKeyBuilder {
	b.version = version
	return b
}

// SetTTL sets the TTL for the cache key
func (b *CacheKeyBuilder) SetTTL(ttl time.Duration) *CacheKeyBuilder {
	b.ttl = ttl
	return b
}

// Build builds the cache key
func (b *CacheKeyBuilder) Build() *CacheKey {
	return &CacheKey{
		QueryType: b.queryType,
		Params:    b.params,
		Version:   b.version,
		TTL:       b.ttl,
	}
}

// Predefined cache key builders for common query types
func NewClassificationCacheKey(businessID string, websiteURL string) *CacheKeyBuilder {
	return NewCacheKeyBuilder("classification").
		AddParam("business_id", businessID).
		AddParam("website_url", websiteURL).
		SetTTL(30 * time.Minute)
}

func NewRiskAssessmentCacheKey(businessID string, riskLevel string) *CacheKeyBuilder {
	return NewCacheKeyBuilder("risk_assessment").
		AddParam("business_id", businessID).
		AddParam("risk_level", riskLevel).
		SetTTL(1 * time.Hour)
}

func NewUserDataCacheKey(userID string) *CacheKeyBuilder {
	return NewCacheKeyBuilder("user_data").
		AddParam("user_id", userID).
		SetTTL(2 * time.Hour)
}

func NewBusinessDataCacheKey(businessID string) *CacheKeyBuilder {
	return NewCacheKeyBuilder("business_data").
		AddParam("business_id", businessID).
		SetTTL(1 * time.Hour)
}
