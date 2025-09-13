package business_intelligence

import (
	"context"
	"crypto/md5"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DataCachingSystem provides comprehensive caching functionality for business intelligence data
type DataCachingSystem struct {
	config           CacheConfig
	logger           *zap.Logger
	caches           map[string]Cache
	strategies       map[string]CacheStrategy
	evictionPolicies map[string]EvictionPolicy
	serializers      map[string]CacheSerializer
	compressors      map[string]CacheCompressor
	encryptors       map[string]CacheEncryptor
	mu               sync.RWMutex
	metrics          *CacheMetrics
	backgroundTasks  map[string]*BackgroundTask
}

// CacheConfig holds configuration for the caching system
type CacheConfig struct {
	// Cache configuration
	DefaultTTL          time.Duration `json:"default_ttl"`
	MaxCacheSize        int64         `json:"max_cache_size"`
	MaxItemSize         int64         `json:"max_item_size"`
	EnableCompression   bool          `json:"enable_compression"`
	EnableEncryption    bool          `json:"enable_encryption"`
	EnableSerialization bool          `json:"enable_serialization"`

	// Cache strategies
	DefaultStrategy         string        `json:"default_strategy"`
	EnableCacheWarming      bool          `json:"enable_cache_warming"`
	WarmingInterval         time.Duration `json:"warming_interval"`
	EnableCacheInvalidation bool          `json:"enable_cache_invalidation"`

	// Eviction policies
	DefaultEvictionPolicy string        `json:"default_eviction_policy"`
	EvictionCheckInterval time.Duration `json:"eviction_check_interval"`
	MaxEvictionBatchSize  int           `json:"max_eviction_batch_size"`

	// Performance optimization
	EnableAsyncOperations bool          `json:"enable_async_operations"`
	AsyncOperationTimeout time.Duration `json:"async_operation_timeout"`
	EnableBatchOperations bool          `json:"enable_batch_operations"`
	BatchSize             int           `json:"batch_size"`

	// Monitoring and metrics
	EnableMetrics             bool          `json:"enable_metrics"`
	MetricsCollectionInterval time.Duration `json:"metrics_collection_interval"`
	EnableCacheStatistics     bool          `json:"enable_cache_statistics"`

	// Cache warming
	WarmingStrategies  []string       `json:"warming_strategies"`
	WarmingDataSources []string       `json:"warming_data_sources"`
	WarmingPriority    map[string]int `json:"warming_priority"`

	// Cache invalidation
	InvalidationStrategies []string      `json:"invalidation_strategies"`
	InvalidationTriggers   []string      `json:"invalidation_triggers"`
	InvalidationTimeout    time.Duration `json:"invalidation_timeout"`
}

// Cache represents a cache implementation
type Cache interface {
	GetName() string
	GetType() string
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
	Clear() error
	Size() int64
	Keys() []string
	GetStats() CacheStats
	IsHealthy() bool
}

// CacheStrategy defines how data should be cached
type CacheStrategy interface {
	GetName() string
	GetType() string
	ShouldCache(key string, value interface{}) bool
	GetTTL(key string, value interface{}) time.Duration
	GetPriority(key string, value interface{}) int
	GetCompressionLevel(key string, value interface{}) int
	GetEncryptionLevel(key string, value interface{}) int
}

// EvictionPolicy defines how items should be evicted from cache
type EvictionPolicy interface {
	GetName() string
	GetType() string
	ShouldEvict(key string, value interface{}, stats CacheItemStats) bool
	GetEvictionScore(key string, value interface{}, stats CacheItemStats) float64
	GetEvictionBatchSize() int
}

// CacheSerializer serializes/deserializes cache data
type CacheSerializer interface {
	GetName() string
	GetType() string
	Serialize(value interface{}) ([]byte, error)
	Deserialize(data []byte, target interface{}) error
	GetCompressionRatio() float64
}

// CacheCompressor compresses/decompresses cache data
type CacheCompressor interface {
	GetName() string
	GetType() string
	Compress(data []byte) ([]byte, error)
	Decompress(data []byte) ([]byte, error)
	GetCompressionRatio() float64
}

// CacheEncryptor encrypts/decrypts cache data
type CacheEncryptor interface {
	GetName() string
	GetType() string
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
	GetSecurityLevel() int
}

// CacheStats represents cache statistics
type CacheStats struct {
	CacheName         string        `json:"cache_name"`
	CacheType         string        `json:"cache_type"`
	TotalItems        int64         `json:"total_items"`
	TotalSize         int64         `json:"total_size"`
	HitCount          int64         `json:"hit_count"`
	MissCount         int64         `json:"miss_count"`
	HitRate           float64       `json:"hit_rate"`
	MissRate          float64       `json:"miss_rate"`
	EvictionCount     int64         `json:"eviction_count"`
	ExpirationCount   int64         `json:"expiration_count"`
	LastAccessTime    time.Time     `json:"last_access_time"`
	LastUpdateTime    time.Time     `json:"last_update_time"`
	AverageAccessTime time.Duration `json:"average_access_time"`
	IsHealthy         bool          `json:"is_healthy"`
}

// CacheItemStats represents statistics for a cache item
type CacheItemStats struct {
	Key         string        `json:"key"`
	Size        int64         `json:"size"`
	AccessCount int64         `json:"access_count"`
	LastAccess  time.Time     `json:"last_access"`
	CreatedAt   time.Time     `json:"created_at"`
	ExpiresAt   time.Time     `json:"expires_at"`
	TTL         time.Duration `json:"ttl"`
	Priority    int           `json:"priority"`
	Compressed  bool          `json:"compressed"`
	Encrypted   bool          `json:"encrypted"`
	Serialized  bool          `json:"serialized"`
}

// CacheMetrics tracks metrics for the caching system
type CacheMetrics struct {
	TotalCaches          int64                  `json:"total_caches"`
	TotalOperations      int64                  `json:"total_operations"`
	SuccessfulOperations int64                  `json:"successful_operations"`
	FailedOperations     int64                  `json:"failed_operations"`
	TotalHits            int64                  `json:"total_hits"`
	TotalMisses          int64                  `json:"total_misses"`
	TotalEvictions       int64                  `json:"total_evictions"`
	TotalExpirations     int64                  `json:"total_expirations"`
	AverageHitRate       float64                `json:"average_hit_rate"`
	AverageOperationTime time.Duration          `json:"average_operation_time"`
	CacheStats           map[string]*CacheStats `json:"cache_stats"`
	LastUpdated          time.Time              `json:"last_updated"`
}

// BackgroundTask represents a background task
type BackgroundTask struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Type       string        `json:"type"`
	Interval   time.Duration `json:"interval"`
	LastRun    time.Time     `json:"last_run"`
	NextRun    time.Time     `json:"next_run"`
	IsRunning  bool          `json:"is_running"`
	RunCount   int64         `json:"run_count"`
	ErrorCount int64         `json:"error_count"`
	LastError  error         `json:"last_error"`
	StopChan   chan bool     `json:"-"`
}

// CacheOperation represents a cache operation
type CacheOperation struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // get, set, delete, clear
	CacheName string                 `json:"cache_name"`
	Key       string                 `json:"key"`
	Value     interface{}            `json:"value"`
	TTL       time.Duration          `json:"ttl"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
	Success   bool                   `json:"success"`
	Error     error                  `json:"error"`
}

// CacheWarmingResult represents the result of cache warming
type CacheWarmingResult struct {
	ID          string                 `json:"id"`
	Strategy    string                 `json:"strategy"`
	DataSources []string               `json:"data_sources"`
	ItemsWarmed int64                  `json:"items_warmed"`
	ItemsFailed int64                  `json:"items_failed"`
	TotalSize   int64                  `json:"total_size"`
	WarmingTime time.Duration          `json:"warming_time"`
	Metadata    map[string]interface{} `json:"metadata"`
	WarmedAt    time.Time              `json:"warmed_at"`
}

// CacheInvalidationResult represents the result of cache invalidation
type CacheInvalidationResult struct {
	ID               string                 `json:"id"`
	Strategy         string                 `json:"strategy"`
	CacheName        string                 `json:"cache_name"`
	KeysInvalidated  int64                  `json:"keys_invalidated"`
	PatternsMatched  []string               `json:"patterns_matched"`
	InvalidationTime time.Duration          `json:"invalidation_time"`
	Metadata         map[string]interface{} `json:"metadata"`
	InvalidatedAt    time.Time              `json:"invalidated_at"`
}

// NewDataCachingSystem creates a new data caching system
func NewDataCachingSystem(config CacheConfig, logger *zap.Logger) *DataCachingSystem {
	return &DataCachingSystem{
		config:           config,
		logger:           logger,
		caches:           make(map[string]Cache),
		strategies:       make(map[string]CacheStrategy),
		evictionPolicies: make(map[string]EvictionPolicy),
		serializers:      make(map[string]CacheSerializer),
		compressors:      make(map[string]CacheCompressor),
		encryptors:       make(map[string]CacheEncryptor),
		metrics: &CacheMetrics{
			CacheStats: make(map[string]*CacheStats),
		},
		backgroundTasks: make(map[string]*BackgroundTask),
	}
}

// RegisterCache registers a cache implementation
func (s *DataCachingSystem) RegisterCache(cache Cache) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := cache.GetName()
	s.caches[name] = cache

	// Initialize cache stats
	s.metrics.CacheStats[name] = &CacheStats{
		CacheName: name,
		CacheType: cache.GetType(),
		IsHealthy: cache.IsHealthy(),
	}

	s.logger.Info("Registered cache",
		zap.String("name", name),
		zap.String("type", cache.GetType()))

	return nil
}

// RegisterCacheStrategy registers a cache strategy
func (s *DataCachingSystem) RegisterCacheStrategy(strategy CacheStrategy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := strategy.GetName()
	s.strategies[name] = strategy

	s.logger.Info("Registered cache strategy",
		zap.String("name", name),
		zap.String("type", strategy.GetType()))

	return nil
}

// RegisterEvictionPolicy registers an eviction policy
func (s *DataCachingSystem) RegisterEvictionPolicy(policy EvictionPolicy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := policy.GetName()
	s.evictionPolicies[name] = policy

	s.logger.Info("Registered eviction policy",
		zap.String("name", name),
		zap.String("type", policy.GetType()))

	return nil
}

// RegisterSerializer registers a cache serializer
func (s *DataCachingSystem) RegisterSerializer(serializer CacheSerializer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := serializer.GetName()
	s.serializers[name] = serializer

	s.logger.Info("Registered cache serializer",
		zap.String("name", name),
		zap.String("type", serializer.GetType()))

	return nil
}

// RegisterCompressor registers a cache compressor
func (s *DataCachingSystem) RegisterCompressor(compressor CacheCompressor) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := compressor.GetName()
	s.compressors[name] = compressor

	s.logger.Info("Registered cache compressor",
		zap.String("name", name),
		zap.String("type", compressor.GetType()))

	return nil
}

// RegisterEncryptor registers a cache encryptor
func (s *DataCachingSystem) RegisterEncryptor(encryptor CacheEncryptor) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := encryptor.GetName()
	s.encryptors[name] = encryptor

	s.logger.Info("Registered cache encryptor",
		zap.String("name", name),
		zap.String("type", encryptor.GetType()))

	return nil
}

// Get retrieves a value from the specified cache
func (s *DataCachingSystem) Get(ctx context.Context, cacheName, key string) (interface{}, error) {
	startTime := time.Now()

	s.logger.Debug("Getting value from cache",
		zap.String("cache_name", cacheName),
		zap.String("key", key))

	// Get cache
	cache := s.getCache(cacheName)
	if cache == nil {
		return nil, fmt.Errorf("cache %s not found", cacheName)
	}

	// Perform get operation
	value, found := cache.Get(key)

	// Update metrics
	s.updateGetMetrics(cacheName, found, time.Since(startTime))

	if !found {
		s.logger.Debug("Cache miss",
			zap.String("cache_name", cacheName),
			zap.String("key", key))
		return nil, nil
	}

	s.logger.Debug("Cache hit",
		zap.String("cache_name", cacheName),
		zap.String("key", key))

	// Deserialize if needed
	if s.config.EnableSerialization {
		deserialized, err := s.deserializeValue(value)
		if err != nil {
			s.logger.Warn("Failed to deserialize cached value",
				zap.String("cache_name", cacheName),
				zap.String("key", key),
				zap.Error(err))
			return nil, err
		}
		value = deserialized
	}

	return value, nil
}

// Set stores a value in the specified cache
func (s *DataCachingSystem) Set(ctx context.Context, cacheName, key string, value interface{}, ttl time.Duration) error {
	startTime := time.Now()

	s.logger.Debug("Setting value in cache",
		zap.String("cache_name", cacheName),
		zap.String("key", key),
		zap.Duration("ttl", ttl))

	// Get cache
	cache := s.getCache(cacheName)
	if cache == nil {
		return fmt.Errorf("cache %s not found", cacheName)
	}

	// Get cache strategy
	strategy := s.getCacheStrategy()
	if strategy != nil {
		// Check if should cache
		if !strategy.ShouldCache(key, value) {
			s.logger.Debug("Strategy indicates not to cache",
				zap.String("cache_name", cacheName),
				zap.String("key", key))
			return nil
		}

		// Get TTL from strategy
		if ttl == 0 {
			ttl = strategy.GetTTL(key, value)
		}
	}

	// Use default TTL if not specified
	if ttl == 0 {
		ttl = s.config.DefaultTTL
	}

	// Serialize if needed
	if s.config.EnableSerialization {
		serialized, err := s.serializeValue(value)
		if err != nil {
			return fmt.Errorf("failed to serialize value: %w", err)
		}
		value = serialized
	}

	// Compress if needed
	if s.config.EnableCompression {
		compressed, err := s.compressValue(value)
		if err != nil {
			s.logger.Warn("Failed to compress value, storing uncompressed",
				zap.String("cache_name", cacheName),
				zap.String("key", key),
				zap.Error(err))
		} else {
			value = compressed
		}
	}

	// Encrypt if needed
	if s.config.EnableEncryption {
		encrypted, err := s.encryptValue(value)
		if err != nil {
			s.logger.Warn("Failed to encrypt value, storing unencrypted",
				zap.String("cache_name", cacheName),
				zap.String("key", key),
				zap.Error(err))
		} else {
			value = encrypted
		}
	}

	// Set value in cache
	err := cache.Set(key, value, ttl)
	if err != nil {
		s.updateSetMetrics(cacheName, false, time.Since(startTime))
		return fmt.Errorf("failed to set value in cache: %w", err)
	}

	// Update metrics
	s.updateSetMetrics(cacheName, true, time.Since(startTime))

	s.logger.Debug("Value set in cache",
		zap.String("cache_name", cacheName),
		zap.String("key", key))

	return nil
}

// Delete removes a value from the specified cache
func (s *DataCachingSystem) Delete(ctx context.Context, cacheName, key string) error {
	startTime := time.Now()

	s.logger.Debug("Deleting value from cache",
		zap.String("cache_name", cacheName),
		zap.String("key", key))

	// Get cache
	cache := s.getCache(cacheName)
	if cache == nil {
		return fmt.Errorf("cache %s not found", cacheName)
	}

	// Delete value
	err := cache.Delete(key)
	if err != nil {
		s.updateDeleteMetrics(cacheName, false, time.Since(startTime))
		return fmt.Errorf("failed to delete value from cache: %w", err)
	}

	// Update metrics
	s.updateDeleteMetrics(cacheName, true, time.Since(startTime))

	s.logger.Debug("Value deleted from cache",
		zap.String("cache_name", cacheName),
		zap.String("key", key))

	return nil
}

// Clear removes all values from the specified cache
func (s *DataCachingSystem) Clear(ctx context.Context, cacheName string) error {
	startTime := time.Now()

	s.logger.Info("Clearing cache",
		zap.String("cache_name", cacheName))

	// Get cache
	cache := s.getCache(cacheName)
	if cache == nil {
		return fmt.Errorf("cache %s not found", cacheName)
	}

	// Clear cache
	err := cache.Clear()
	if err != nil {
		s.updateClearMetrics(cacheName, false, time.Since(startTime))
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	// Update metrics
	s.updateClearMetrics(cacheName, true, time.Since(startTime))

	s.logger.Info("Cache cleared",
		zap.String("cache_name", cacheName))

	return nil
}

// WarmCache warms the cache with data from specified sources
func (s *DataCachingSystem) WarmCache(ctx context.Context, cacheName string, strategy string) (*CacheWarmingResult, error) {
	startTime := time.Now()

	s.logger.Info("Starting cache warming",
		zap.String("cache_name", cacheName),
		zap.String("strategy", strategy))

	// Get cache
	cache := s.getCache(cacheName)
	if cache == nil {
		return nil, fmt.Errorf("cache %s not found", cacheName)
	}

	// Get warming strategy
	warmingStrategy := s.getWarmingStrategy(strategy)
	if warmingStrategy == nil {
		return nil, fmt.Errorf("warming strategy %s not found", strategy)
	}

	// Perform cache warming
	result, err := warmingStrategy.WarmCache(ctx, cache, s.config.WarmingDataSources)
	if err != nil {
		return nil, fmt.Errorf("cache warming failed: %w", err)
	}

	// Update metrics
	s.updateWarmingMetrics(result)

	s.logger.Info("Cache warming completed",
		zap.String("cache_name", cacheName),
		zap.String("strategy", strategy),
		zap.Int64("items_warmed", result.ItemsWarmed),
		zap.Duration("warming_time", result.WarmingTime))

	return result, nil
}

// InvalidateCache invalidates cache entries based on patterns or triggers
func (s *DataCachingSystem) InvalidateCache(ctx context.Context, cacheName string, strategy string, patterns []string) (*CacheInvalidationResult, error) {
	startTime := time.Now()

	s.logger.Info("Starting cache invalidation",
		zap.String("cache_name", cacheName),
		zap.String("strategy", strategy),
		zap.Strings("patterns", patterns))

	// Get cache
	cache := s.getCache(cacheName)
	if cache == nil {
		return nil, fmt.Errorf("cache %s not found", cacheName)
	}

	// Get invalidation strategy
	invalidationStrategy := s.getInvalidationStrategy(strategy)
	if invalidationStrategy == nil {
		return nil, fmt.Errorf("invalidation strategy %s not found", strategy)
	}

	// Perform cache invalidation
	result, err := invalidationStrategy.InvalidateCache(ctx, cache, patterns)
	if err != nil {
		return nil, fmt.Errorf("cache invalidation failed: %w", err)
	}

	// Update metrics
	s.updateInvalidationMetrics(result)

	s.logger.Info("Cache invalidation completed",
		zap.String("cache_name", cacheName),
		zap.String("strategy", strategy),
		zap.Int64("keys_invalidated", result.KeysInvalidated),
		zap.Duration("invalidation_time", result.InvalidationTime))

	return result, nil
}

// GetCacheStats returns statistics for the specified cache
func (s *DataCachingSystem) GetCacheStats(cacheName string) (*CacheStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cache := s.caches[cacheName]
	if cache == nil {
		return nil, fmt.Errorf("cache %s not found", cacheName)
	}

	stats := cache.GetStats()
	return &stats, nil
}

// GetAllCacheStats returns statistics for all caches
func (s *DataCachingSystem) GetAllCacheStats() map[string]*CacheStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]*CacheStats)
	for name, cache := range s.caches {
		cacheStats := cache.GetStats()
		stats[name] = &cacheStats
	}

	return stats
}

// GetMetrics returns current caching system metrics
func (s *DataCachingSystem) GetMetrics() *CacheMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to avoid race conditions
	metrics := *s.metrics
	return &metrics
}

// StartBackgroundTasks starts background tasks for cache maintenance
func (s *DataCachingSystem) StartBackgroundTasks(ctx context.Context) error {
	s.logger.Info("Starting background tasks")

	// Start eviction task
	if s.config.DefaultEvictionPolicy != "" {
		task := &BackgroundTask{
			ID:       "eviction",
			Name:     "Cache Eviction",
			Type:     "eviction",
			Interval: s.config.EvictionCheckInterval,
			StopChan: make(chan bool),
		}
		s.backgroundTasks["eviction"] = task
		go s.runEvictionTask(ctx, task)
	}

	// Start cache warming task
	if s.config.EnableCacheWarming {
		task := &BackgroundTask{
			ID:       "warming",
			Name:     "Cache Warming",
			Type:     "warming",
			Interval: s.config.WarmingInterval,
			StopChan: make(chan bool),
		}
		s.backgroundTasks["warming"] = task
		go s.runWarmingTask(ctx, task)
	}

	// Start metrics collection task
	if s.config.EnableMetrics {
		task := &BackgroundTask{
			ID:       "metrics",
			Name:     "Metrics Collection",
			Type:     "metrics",
			Interval: s.config.MetricsCollectionInterval,
			StopChan: make(chan bool),
		}
		s.backgroundTasks["metrics"] = task
		go s.runMetricsTask(ctx, task)
	}

	s.logger.Info("Background tasks started",
		zap.Int("task_count", len(s.backgroundTasks)))

	return nil
}

// StopBackgroundTasks stops all background tasks
func (s *DataCachingSystem) StopBackgroundTasks() {
	s.logger.Info("Stopping background tasks")

	for name, task := range s.backgroundTasks {
		s.logger.Info("Stopping background task",
			zap.String("task_name", name))

		select {
		case task.StopChan <- true:
		default:
		}
	}

	s.backgroundTasks = make(map[string]*BackgroundTask)
	s.logger.Info("Background tasks stopped")
}

// Helper methods

// getCache returns a cache by name
func (s *DataCachingSystem) getCache(name string) Cache {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.caches[name]
}

// getCacheStrategy returns the default cache strategy
func (s *DataCachingSystem) getCacheStrategy() CacheStrategy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.strategies[s.config.DefaultStrategy]
}

// getWarmingStrategy returns a warming strategy by name
func (s *DataCachingSystem) getWarmingStrategy(name string) CacheWarmingStrategy {
	// This would be implemented based on specific warming strategies
	// For now, return nil as this is a placeholder
	return nil
}

// getInvalidationStrategy returns an invalidation strategy by name
func (s *DataCachingSystem) getInvalidationStrategy(name string) CacheInvalidationStrategy {
	// This would be implemented based on specific invalidation strategies
	// For now, return nil as this is a placeholder
	return nil
}

// serializeValue serializes a value
func (s *DataCachingSystem) serializeValue(value interface{}) (interface{}, error) {
	serializer := s.getDefaultSerializer()
	if serializer == nil {
		return value, nil
	}

	data, err := serializer.Serialize(value)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// deserializeValue deserializes a value
func (s *DataCachingSystem) deserializeValue(value interface{}) (interface{}, error) {
	serializer := s.getDefaultSerializer()
	if serializer == nil {
		return value, nil
	}

	data, ok := value.([]byte)
	if !ok {
		return value, nil
	}

	var result interface{}
	err := serializer.Deserialize(data, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// compressValue compresses a value
func (s *DataCachingSystem) compressValue(value interface{}) (interface{}, error) {
	compressor := s.getDefaultCompressor()
	if compressor == nil {
		return value, nil
	}

	data, ok := value.([]byte)
	if !ok {
		return value, nil
	}

	compressed, err := compressor.Compress(data)
	if err != nil {
		return nil, err
	}

	return compressed, nil
}

// encryptValue encrypts a value
func (s *DataCachingSystem) encryptValue(value interface{}) (interface{}, error) {
	encryptor := s.getDefaultEncryptor()
	if encryptor == nil {
		return value, nil
	}

	data, ok := value.([]byte)
	if !ok {
		return value, nil
	}

	encrypted, err := encryptor.Encrypt(data)
	if err != nil {
		return nil, err
	}

	return encrypted, nil
}

// getDefaultSerializer returns the default serializer
func (s *DataCachingSystem) getDefaultSerializer() CacheSerializer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return first available serializer
	for _, serializer := range s.serializers {
		return serializer
	}
	return nil
}

// getDefaultCompressor returns the default compressor
func (s *DataCachingSystem) getDefaultCompressor() CacheCompressor {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return first available compressor
	for _, compressor := range s.compressors {
		return compressor
	}
	return nil
}

// getDefaultEncryptor returns the default encryptor
func (s *DataCachingSystem) getDefaultEncryptor() CacheEncryptor {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return first available encryptor
	for _, encryptor := range s.encryptors {
		return encryptor
	}
	return nil
}

// Metrics update methods

// updateGetMetrics updates metrics for get operations
func (s *DataCachingSystem) updateGetMetrics(cacheName string, hit bool, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics.TotalOperations++
	s.metrics.SuccessfulOperations++

	if hit {
		s.metrics.TotalHits++
	} else {
		s.metrics.TotalMisses++
	}

	// Update cache-specific stats
	if stats, exists := s.metrics.CacheStats[cacheName]; exists {
		if hit {
			stats.HitCount++
		} else {
			stats.MissCount++
		}
		stats.LastAccessTime = time.Now()
		stats.AverageAccessTime = (stats.AverageAccessTime + duration) / 2
	}

	// Update average hit rate
	totalRequests := s.metrics.TotalHits + s.metrics.TotalMisses
	if totalRequests > 0 {
		s.metrics.AverageHitRate = float64(s.metrics.TotalHits) / float64(totalRequests)
	}

	// Update average operation time
	s.metrics.AverageOperationTime = (s.metrics.AverageOperationTime + duration) / 2
	s.metrics.LastUpdated = time.Now()
}

// updateSetMetrics updates metrics for set operations
func (s *DataCachingSystem) updateSetMetrics(cacheName string, success bool, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics.TotalOperations++
	if success {
		s.metrics.SuccessfulOperations++
	} else {
		s.metrics.FailedOperations++
	}

	// Update cache-specific stats
	if stats, exists := s.metrics.CacheStats[cacheName]; exists {
		stats.LastUpdateTime = time.Now()
		stats.AverageAccessTime = (stats.AverageAccessTime + duration) / 2
	}

	// Update average operation time
	s.metrics.AverageOperationTime = (s.metrics.AverageOperationTime + duration) / 2
	s.metrics.LastUpdated = time.Now()
}

// updateDeleteMetrics updates metrics for delete operations
func (s *DataCachingSystem) updateDeleteMetrics(cacheName string, success bool, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics.TotalOperations++
	if success {
		s.metrics.SuccessfulOperations++
	} else {
		s.metrics.FailedOperations++
	}

	// Update average operation time
	s.metrics.AverageOperationTime = (s.metrics.AverageOperationTime + duration) / 2
	s.metrics.LastUpdated = time.Now()
}

// updateClearMetrics updates metrics for clear operations
func (s *DataCachingSystem) updateClearMetrics(cacheName string, success bool, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics.TotalOperations++
	if success {
		s.metrics.SuccessfulOperations++
	} else {
		s.metrics.FailedOperations++
	}

	// Update average operation time
	s.metrics.AverageOperationTime = (s.metrics.AverageOperationTime + duration) / 2
	s.metrics.LastUpdated = time.Now()
}

// updateWarmingMetrics updates metrics for cache warming
func (s *DataCachingSystem) updateWarmingMetrics(result *CacheWarmingResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update warming metrics
	s.metrics.LastUpdated = time.Now()
}

// updateInvalidationMetrics updates metrics for cache invalidation
func (s *DataCachingSystem) updateInvalidationMetrics(result *CacheInvalidationResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics.TotalEvictions += result.KeysInvalidated
	s.metrics.LastUpdated = time.Now()
}

// Background task methods

// runEvictionTask runs the cache eviction background task
func (s *DataCachingSystem) runEvictionTask(ctx context.Context, task *BackgroundTask) {
	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-task.StopChan:
			return
		case <-ticker.C:
			task.IsRunning = true
			task.LastRun = time.Now()
			task.RunCount++

			// Perform eviction
			err := s.performEviction(ctx)
			if err != nil {
				task.ErrorCount++
				task.LastError = err
				s.logger.Error("Cache eviction failed",
					zap.Error(err))
			}

			task.IsRunning = false
			task.NextRun = time.Now().Add(task.Interval)
		}
	}
}

// runWarmingTask runs the cache warming background task
func (s *DataCachingSystem) runWarmingTask(ctx context.Context, task *BackgroundTask) {
	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-task.StopChan:
			return
		case <-ticker.C:
			task.IsRunning = true
			task.LastRun = time.Now()
			task.RunCount++

			// Perform cache warming for all caches
			for cacheName := range s.caches {
				_, err := s.WarmCache(ctx, cacheName, s.config.DefaultStrategy)
				if err != nil {
					task.ErrorCount++
					task.LastError = err
					s.logger.Error("Cache warming failed",
						zap.String("cache_name", cacheName),
						zap.Error(err))
				}
			}

			task.IsRunning = false
			task.NextRun = time.Now().Add(task.Interval)
		}
	}
}

// runMetricsTask runs the metrics collection background task
func (s *DataCachingSystem) runMetricsTask(ctx context.Context, task *BackgroundTask) {
	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-task.StopChan:
			return
		case <-ticker.C:
			task.IsRunning = true
			task.LastRun = time.Now()
			task.RunCount++

			// Update cache statistics
			s.updateAllCacheStats()

			task.IsRunning = false
			task.NextRun = time.Now().Add(task.Interval)
		}
	}
}

// performEviction performs cache eviction based on the eviction policy
func (s *DataCachingSystem) performEviction(ctx context.Context) error {
	policy := s.getEvictionPolicy()
	if policy == nil {
		return nil
	}

	// Perform eviction for each cache
	for name, cache := range s.caches {
		err := s.evictFromCache(ctx, cache, policy)
		if err != nil {
			s.logger.Error("Failed to evict from cache",
				zap.String("cache_name", name),
				zap.Error(err))
		}
	}

	return nil
}

// evictFromCache evicts items from a specific cache
func (s *DataCachingSystem) evictFromCache(ctx context.Context, cache Cache, policy EvictionPolicy) error {
	// Get all keys from cache
	keys := cache.Keys()

	// Evaluate each key for eviction
	var keysToEvict []string
	for _, key := range keys {
		// Get item stats (this would need to be implemented in the cache interface)
		// For now, create a placeholder
		stats := CacheItemStats{
			Key: key,
			// Other fields would be populated from actual cache implementation
		}

		if policy.ShouldEvict(key, nil, stats) {
			keysToEvict = append(keysToEvict, key)
		}
	}

	// Evict items in batches
	batchSize := policy.GetEvictionBatchSize()
	if batchSize <= 0 {
		batchSize = s.config.MaxEvictionBatchSize
	}

	for i := 0; i < len(keysToEvict); i += batchSize {
		end := i + batchSize
		if end > len(keysToEvict) {
			end = len(keysToEvict)
		}

		batch := keysToEvict[i:end]
		for _, key := range batch {
			err := cache.Delete(key)
			if err != nil {
				s.logger.Error("Failed to evict key",
					zap.String("key", key),
					zap.Error(err))
			}
		}
	}

	return nil
}

// getEvictionPolicy returns the default eviction policy
func (s *DataCachingSystem) getEvictionPolicy() EvictionPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.evictionPolicies[s.config.DefaultEvictionPolicy]
}

// updateAllCacheStats updates statistics for all caches
func (s *DataCachingSystem) updateAllCacheStats() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for name, cache := range s.caches {
		stats := cache.GetStats()
		s.metrics.CacheStats[name] = &stats
	}

	s.metrics.LastUpdated = time.Now()
}

// generateCacheKey generates a cache key with optional hashing
func generateCacheKey(prefix string, parts ...string) string {
	key := prefix
	for _, part := range parts {
		key += ":" + part
	}

	// Hash the key if it's too long
	if len(key) > 250 {
		hash := md5.Sum([]byte(key))
		key = prefix + ":" + fmt.Sprintf("%x", hash)
	}

	return key
}

// Additional interfaces for cache warming and invalidation strategies

// CacheWarmingStrategy defines how cache warming should be performed
type CacheWarmingStrategy interface {
	GetName() string
	WarmCache(ctx context.Context, cache Cache, dataSources []string) (*CacheWarmingResult, error)
}

// CacheInvalidationStrategy defines how cache invalidation should be performed
type CacheInvalidationStrategy interface {
	GetName() string
	InvalidateCache(ctx context.Context, cache Cache, patterns []string) (*CacheInvalidationResult, error)
}
