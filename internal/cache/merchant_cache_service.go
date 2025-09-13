package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// MerchantCacheService provides caching for merchant-related data
type MerchantCacheService struct {
	cache     Cache
	monitor   *CacheMonitoringService
	logger    *zap.Logger
	keyPrefix string
}

// MerchantCacheKey represents different types of merchant cache keys
type MerchantCacheKey string

const (
	MerchantListKey      MerchantCacheKey = "merchants:list"
	MerchantDetailKey    MerchantCacheKey = "merchants:detail"
	MerchantSearchKey    MerchantCacheKey = "merchants:search"
	MerchantStatsKey     MerchantCacheKey = "merchants:stats"
	MerchantPortfolioKey MerchantCacheKey = "merchants:portfolio"
)

// MerchantCacheConfig holds configuration for merchant caching
type MerchantCacheConfig struct {
	DefaultTTL       time.Duration `json:"default_ttl"`
	SearchTTL        time.Duration `json:"search_ttl"`
	DetailTTL        time.Duration `json:"detail_ttl"`
	ListTTL          time.Duration `json:"list_ttl"`
	StatsTTL         time.Duration `json:"stats_ttl"`
	PortfolioTTL     time.Duration `json:"portfolio_ttl"`
	KeyPrefix        string        `json:"key_prefix"`
	EnableMonitoring bool          `json:"enable_monitoring"`
}

// NewMerchantCacheService creates a new merchant cache service
func NewMerchantCacheService(cache Cache, config *MerchantCacheConfig, logger *zap.Logger) *MerchantCacheService {
	if config == nil {
		config = &MerchantCacheConfig{
			DefaultTTL:       15 * time.Minute,
			SearchTTL:        5 * time.Minute,
			DetailTTL:        30 * time.Minute,
			ListTTL:          10 * time.Minute,
			StatsTTL:         1 * time.Hour,
			PortfolioTTL:     20 * time.Minute,
			KeyPrefix:        "kyb:merchants",
			EnableMonitoring: true,
		}
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	service := &MerchantCacheService{
		cache:     cache,
		logger:    logger,
		keyPrefix: config.KeyPrefix,
	}

	// Initialize monitoring if enabled
	if config.EnableMonitoring {
		monitorConfig := &MonitoringConfig{
			CollectionInterval: 30 * time.Second,
			EnableAlerts:       true,
		}
		service.monitor = NewCacheMonitoringService([]Cache{cache}, monitorConfig, logger)

		if err := service.monitor.Start(); err != nil {
			logger.Error("Failed to start cache monitor", zap.Error(err))
		}
	}

	logger.Info("Merchant cache service initialized",
		zap.String("key_prefix", config.KeyPrefix),
		zap.Duration("default_ttl", config.DefaultTTL),
		zap.Bool("monitoring_enabled", config.EnableMonitoring))

	return service
}

// CacheMerchantList caches a list of merchants
func (mcs *MerchantCacheService) CacheMerchantList(ctx context.Context, filters map[string]interface{}, merchants interface{}, ttl time.Duration) error {
	key := mcs.generateListKey(filters)

	data, err := json.Marshal(merchants)
	if err != nil {
		return fmt.Errorf("failed to marshal merchant list: %w", err)
	}

	if ttl == 0 {
		ttl = 10 * time.Minute // Default list TTL
	}

	err = mcs.cache.Set(ctx, key, data, ttl)
	if err != nil {
		mcs.logger.Error("Failed to cache merchant list",
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("failed to cache merchant list: %w", err)
	}

	mcs.logger.Debug("Cached merchant list",
		zap.String("key", key),
		zap.Duration("ttl", ttl),
		zap.Int("data_size", len(data)))

	return nil
}

// GetMerchantList retrieves a cached list of merchants
func (mcs *MerchantCacheService) GetMerchantList(ctx context.Context, filters map[string]interface{}, result interface{}) (bool, error) {
	key := mcs.generateListKey(filters)

	data, err := mcs.cache.Get(ctx, key)
	if err != nil {
		if err == CacheNotFoundError {
			return false, nil // Cache miss
		}
		return false, fmt.Errorf("failed to get merchant list from cache: %w", err)
	}

	err = json.Unmarshal(data, result)
	if err != nil {
		mcs.logger.Error("Failed to unmarshal cached merchant list",
			zap.String("key", key),
			zap.Error(err))
		return false, fmt.Errorf("failed to unmarshal cached merchant list: %w", err)
	}

	mcs.logger.Debug("Retrieved merchant list from cache",
		zap.String("key", key),
		zap.Int("data_size", len(data)))

	return true, nil // Cache hit
}

// CacheMerchantDetail caches detailed merchant information
func (mcs *MerchantCacheService) CacheMerchantDetail(ctx context.Context, merchantID string, merchant interface{}, ttl time.Duration) error {
	key := mcs.generateDetailKey(merchantID)

	data, err := json.Marshal(merchant)
	if err != nil {
		return fmt.Errorf("failed to marshal merchant detail: %w", err)
	}

	if ttl == 0 {
		ttl = 30 * time.Minute // Default detail TTL
	}

	err = mcs.cache.Set(ctx, key, data, ttl)
	if err != nil {
		mcs.logger.Error("Failed to cache merchant detail",
			zap.String("key", key),
			zap.String("merchant_id", merchantID),
			zap.Error(err))
		return fmt.Errorf("failed to cache merchant detail: %w", err)
	}

	mcs.logger.Debug("Cached merchant detail",
		zap.String("key", key),
		zap.String("merchant_id", merchantID),
		zap.Duration("ttl", ttl))

	return nil
}

// GetMerchantDetail retrieves cached merchant detail
func (mcs *MerchantCacheService) GetMerchantDetail(ctx context.Context, merchantID string, result interface{}) (bool, error) {
	key := mcs.generateDetailKey(merchantID)

	data, err := mcs.cache.Get(ctx, key)
	if err != nil {
		if err == CacheNotFoundError {
			return false, nil // Cache miss
		}
		return false, fmt.Errorf("failed to get merchant detail from cache: %w", err)
	}

	err = json.Unmarshal(data, result)
	if err != nil {
		mcs.logger.Error("Failed to unmarshal cached merchant detail",
			zap.String("key", key),
			zap.String("merchant_id", merchantID),
			zap.Error(err))
		return false, fmt.Errorf("failed to unmarshal cached merchant detail: %w", err)
	}

	mcs.logger.Debug("Retrieved merchant detail from cache",
		zap.String("key", key),
		zap.String("merchant_id", merchantID))

	return true, nil // Cache hit
}

// CacheMerchantSearch caches search results
func (mcs *MerchantCacheService) CacheMerchantSearch(ctx context.Context, query string, filters map[string]interface{}, results interface{}, ttl time.Duration) error {
	key := mcs.generateSearchKey(query, filters)

	data, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("failed to marshal search results: %w", err)
	}

	if ttl == 0 {
		ttl = 5 * time.Minute // Default search TTL
	}

	err = mcs.cache.Set(ctx, key, data, ttl)
	if err != nil {
		mcs.logger.Error("Failed to cache search results",
			zap.String("key", key),
			zap.String("query", query),
			zap.Error(err))
		return fmt.Errorf("failed to cache search results: %w", err)
	}

	mcs.logger.Debug("Cached search results",
		zap.String("key", key),
		zap.String("query", query),
		zap.Duration("ttl", ttl))

	return nil
}

// GetMerchantSearch retrieves cached search results
func (mcs *MerchantCacheService) GetMerchantSearch(ctx context.Context, query string, filters map[string]interface{}, result interface{}) (bool, error) {
	key := mcs.generateSearchKey(query, filters)

	data, err := mcs.cache.Get(ctx, key)
	if err != nil {
		if err == CacheNotFoundError {
			return false, nil // Cache miss
		}
		return false, fmt.Errorf("failed to get search results from cache: %w", err)
	}

	err = json.Unmarshal(data, result)
	if err != nil {
		mcs.logger.Error("Failed to unmarshal cached search results",
			zap.String("key", key),
			zap.String("query", query),
			zap.Error(err))
		return false, fmt.Errorf("failed to unmarshal cached search results: %w", err)
	}

	mcs.logger.Debug("Retrieved search results from cache",
		zap.String("key", key),
		zap.String("query", query))

	return true, nil // Cache hit
}

// CacheMerchantStats caches merchant statistics
func (mcs *MerchantCacheService) CacheMerchantStats(ctx context.Context, stats interface{}, ttl time.Duration) error {
	key := mcs.generateStatsKey()

	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal merchant stats: %w", err)
	}

	if ttl == 0 {
		ttl = 1 * time.Hour // Default stats TTL
	}

	err = mcs.cache.Set(ctx, key, data, ttl)
	if err != nil {
		mcs.logger.Error("Failed to cache merchant stats",
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("failed to cache merchant stats: %w", err)
	}

	mcs.logger.Debug("Cached merchant stats",
		zap.String("key", key),
		zap.Duration("ttl", ttl))

	return nil
}

// GetMerchantStats retrieves cached merchant statistics
func (mcs *MerchantCacheService) GetMerchantStats(ctx context.Context, result interface{}) (bool, error) {
	key := mcs.generateStatsKey()

	data, err := mcs.cache.Get(ctx, key)
	if err != nil {
		if err == CacheNotFoundError {
			return false, nil // Cache miss
		}
		return false, fmt.Errorf("failed to get merchant stats from cache: %w", err)
	}

	err = json.Unmarshal(data, result)
	if err != nil {
		mcs.logger.Error("Failed to unmarshal cached merchant stats",
			zap.String("key", key),
			zap.Error(err))
		return false, fmt.Errorf("failed to unmarshal cached merchant stats: %w", err)
	}

	mcs.logger.Debug("Retrieved merchant stats from cache",
		zap.String("key", key))

	return true, nil // Cache hit
}

// InvalidateMerchant invalidates all cache entries for a specific merchant
func (mcs *MerchantCacheService) InvalidateMerchant(ctx context.Context, merchantID string) error {
	// Get all keys with our prefix to invalidate all merchant-related data
	pattern := mcs.keyPrefix + ":*"
	keys, err := mcs.cache.GetKeys(ctx, pattern)
	if err != nil {
		mcs.logger.Error("Failed to get cache keys for invalidation",
			zap.String("pattern", pattern),
			zap.String("merchant_id", merchantID),
			zap.Error(err))
		return err
	}

	// Delete all keys
	for _, key := range keys {
		err := mcs.cache.Delete(ctx, key)
		if err != nil && err != CacheNotFoundError {
			mcs.logger.Error("Failed to invalidate cache key",
				zap.String("key", key),
				zap.String("merchant_id", merchantID),
				zap.Error(err))
		}
	}

	mcs.logger.Info("Invalidated cache for merchant",
		zap.String("merchant_id", merchantID),
		zap.Int("keys_invalidated", len(keys)))

	return nil
}

// InvalidateAll invalidates all merchant cache entries
func (mcs *MerchantCacheService) InvalidateAll(ctx context.Context) error {
	// Get all keys with our prefix
	keys, err := mcs.cache.GetKeys(ctx, mcs.keyPrefix+":*")
	if err != nil {
		return fmt.Errorf("failed to get cache keys: %w", err)
	}

	// Delete all keys
	err = mcs.cache.DeleteEntries(ctx, keys)
	if err != nil {
		return fmt.Errorf("failed to delete cache entries: %w", err)
	}

	mcs.logger.Info("Invalidated all merchant cache entries",
		zap.Int("keys_invalidated", len(keys)))

	return nil
}

// GetCacheStats returns cache statistics
func (mcs *MerchantCacheService) GetCacheStats(ctx context.Context) (*CacheStats, error) {
	return mcs.cache.GetStats(ctx)
}

// GetMonitor returns the cache monitor
func (mcs *MerchantCacheService) GetMonitor() *CacheMonitoringService {
	return mcs.monitor
}

// Close closes the cache service and monitor
func (mcs *MerchantCacheService) Close() error {
	if mcs.monitor != nil {
		if err := mcs.monitor.Stop(); err != nil {
			mcs.logger.Error("Failed to stop cache monitor", zap.Error(err))
		}
	}

	return mcs.cache.Close()
}

// Helper methods for key generation

func (mcs *MerchantCacheService) generateListKey(filters map[string]interface{}) string {
	if filters == nil || len(filters) == 0 {
		return fmt.Sprintf("%s:list:default", mcs.keyPrefix)
	}

	// Create a hash of the filters for the key
	filterHash := fmt.Sprintf("%v", filters)
	return fmt.Sprintf("%s:list:%s", mcs.keyPrefix, filterHash)
}

func (mcs *MerchantCacheService) generateDetailKey(merchantID string) string {
	return fmt.Sprintf("%s:detail:%s", mcs.keyPrefix, merchantID)
}

func (mcs *MerchantCacheService) generateSearchKey(query string, filters map[string]interface{}) string {
	if query == "" && (filters == nil || len(filters) == 0) {
		return fmt.Sprintf("%s:search:default", mcs.keyPrefix)
	}

	searchHash := fmt.Sprintf("%s:%v", query, filters)
	return fmt.Sprintf("%s:search:%s", mcs.keyPrefix, searchHash)
}

func (mcs *MerchantCacheService) generateStatsKey() string {
	return fmt.Sprintf("%s:stats", mcs.keyPrefix)
}
