package cache

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"kyb-platform/internal/observability"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// IntelligentCache provides multi-level caching with intelligent optimization
type IntelligentCache struct {
	// Configuration
	config *IntelligentCacheConfig

	// Observability
	logger *observability.Logger
	tracer trace.Tracer

	// Cache levels
	memoryCache      *MemoryCacheImpl
	diskCache        *DiskCache
	distributedCache *DistributedCache

	// Disk cache index
	diskIndex map[string]*cacheItem
	diskMux   sync.RWMutex

	// Cache management
	manager    *CacheManager
	managerMux sync.RWMutex

	// Performance monitoring
	monitor    *CacheMonitor
	monitorMux sync.RWMutex

	// Warming and optimization
	warmer    *CacheWarmer
	warmerMux sync.RWMutex

	// Context for shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// IntelligentCacheConfig configuration for intelligent caching
type IntelligentCacheConfig struct {
	// Memory cache settings
	MemoryCacheSize      int
	MemoryCacheTTL       time.Duration
	MemoryEvictionPolicy string

	// Disk cache settings
	DiskCacheEnabled bool
	DiskCachePath    string
	DiskCacheSize    int64
	DiskCacheTTL     time.Duration
	DiskCompression  bool

	// Distributed cache settings
	DistributedCacheEnabled bool
	DistributedCacheURL     string
	DistributedCacheTTL     time.Duration
	DistributedCachePool    int

	// Cache warming settings
	WarmingEnabled   bool
	WarmingInterval  time.Duration
	WarmingBatchSize int
	WarmingStrategy  string

	// Performance settings
	PerformanceMonitoring bool
	PerformanceInterval   time.Duration
	HitRateThreshold      float64
	OptimizationInterval  time.Duration

	// Invalidation settings
	InvalidationStrategy  string
	InvalidationBatchSize int
	InvalidationTimeout   time.Duration
	InvalidationCooldown  time.Duration
}

// Note: MemoryCache and DiskCache types are defined in their respective files
// memory.go and disk_cache.go to avoid conflicts

// DistributedCache provides distributed caching
type DistributedCache struct {
	URL         string
	TTL         time.Duration
	Pool        int
	Connections map[string]*DistributedConnection
	Mux         sync.RWMutex
}

// CacheEntry represents a cache entry
type CacheEntry struct {
	Key          string
	Value        interface{}
	CreatedAt    time.Time
	ExpiresAt    time.Time
	LastAccessed time.Time
	AccessCount  int64
	Size         int64
	Compressed   bool
	Metadata     map[string]interface{}
}

// DiskEntry represents a disk cache entry
type DiskEntry struct {
	Key          string
	FilePath     string
	Size         int64
	CreatedAt    time.Time
	ExpiresAt    time.Time
	LastAccessed time.Time
	AccessCount  int64
	Compressed   bool
	Checksum     string
}

// DistributedConnection represents a distributed cache connection
type DistributedConnection struct {
	ID         string
	URL        string
	Connected  bool
	LastPing   time.Time
	Latency    time.Duration
	ErrorCount int
}

// CacheManager manages cache operations
type CacheManager struct {
	Strategy         string
	Invalidations    map[string]*InvalidationInfo
	Optimizations    map[string]*OptimizationInfo
	LastOptimization time.Time
	Mux              sync.RWMutex
}

// InvalidationInfo represents cache invalidation information
type InvalidationInfo struct {
	Pattern           string
	LastInvalidated   time.Time
	InvalidationCount int64
	AffectedKeys      []string
}

// OptimizationInfo represents cache optimization information
type OptimizationInfo struct {
	Type              string
	LastOptimized     time.Time
	OptimizationCount int64
	Improvement       float64
}

// CacheMonitor monitors cache performance
type CacheMonitor struct {
	Metrics    *CacheMetrics
	Alerts     []*CacheAlert
	Thresholds map[string]float64
	LastAlert  time.Time
	Mux        sync.RWMutex
}

// CacheMetrics represents cache performance metrics
type CacheMetrics struct {
	MemoryHits        int64
	MemoryMisses      int64
	DiskHits          int64
	DiskMisses        int64
	DistributedHits   int64
	DistributedMisses int64
	TotalHits         int64
	TotalMisses       int64
	HitRate           float64
	MemoryUsage       int64
	DiskUsage         int64
	AverageLatency    time.Duration
	Evictions         int64
	Invalidations     int64
	LastUpdate        time.Time
}

// CacheAlert represents a cache alert
type CacheAlert struct {
	ID           string
	Type         string
	Severity     string
	Message      string
	Metric       string
	Value        float64
	Threshold    float64
	Timestamp    time.Time
	Acknowledged bool
}

// CacheWarmer manages cache warming
type CacheWarmer struct {
	Strategy     string
	WarmingQueue []*WarmingTask
	WarmingStats map[string]*WarmingStats
	LastWarming  time.Time
	Mux          sync.RWMutex
}

// WarmingTask represents a cache warming task
type WarmingTask struct {
	ID          string
	Key         string
	Priority    int
	CreatedAt   time.Time
	Attempts    int
	MaxAttempts int
	LastAttempt time.Time
}

// WarmingStats represents warming statistics
type WarmingStats struct {
	TotalWarmed     int64
	SuccessfulWarms int64
	FailedWarms     int64
	AverageWarmTime time.Duration
	LastWarmed      time.Time
}

// NewIntelligentCache creates a new intelligent cache
func NewIntelligentCache(config *IntelligentCacheConfig, logger *observability.Logger, tracer trace.Tracer) *IntelligentCache {
	if config == nil {
		config = &IntelligentCacheConfig{
			MemoryCacheSize:         1000,
			MemoryCacheTTL:          30 * time.Minute,
			MemoryEvictionPolicy:    "lru",
			DiskCacheEnabled:        true,
			DiskCachePath:           "./cache",
			DiskCacheSize:           100 * 1024 * 1024, // 100MB
			DiskCacheTTL:            2 * time.Hour,
			DiskCompression:         true,
			DistributedCacheEnabled: false,
			DistributedCacheURL:     "",
			DistributedCacheTTL:     1 * time.Hour,
			DistributedCachePool:    10,
			WarmingEnabled:          true,
			WarmingInterval:         5 * time.Minute,
			WarmingBatchSize:        100,
			WarmingStrategy:         "frequent",
			PerformanceMonitoring:   true,
			PerformanceInterval:     30 * time.Second,
			HitRateThreshold:        0.8,
			OptimizationInterval:    10 * time.Minute,
			InvalidationStrategy:    "pattern",
			InvalidationBatchSize:   100,
			InvalidationTimeout:     30 * time.Second,
			InvalidationCooldown:    1 * time.Minute,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	ic := &IntelligentCache{
		config:    config,
		logger:    logger,
		tracer:    tracer,
		ctx:       ctx,
		cancel:    cancel,
		diskIndex: make(map[string]*cacheItem),
	}

	// Initialize cache levels
	memoryConfig := &CacheConfig{
		Type:       MemoryCache,
		DefaultTTL: config.MemoryCacheTTL,
		MaxSize:    config.MemoryCacheSize,
	}
	ic.memoryCache = NewMemoryCache(memoryConfig)

	if config.DiskCacheEnabled {
		diskConfig := DiskCacheConfig{
			Path: config.DiskCachePath,
			Size: config.DiskCacheSize,
			TTL:  config.DiskCacheTTL,
		}
		var err error
		ic.diskCache, err = NewDiskCache(diskConfig, ic.logger.GetZapLogger())
		if err != nil {
			// Log error and continue without disk cache
			ic.logger.Error("Failed to initialize disk cache", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	if config.DistributedCacheEnabled {
		ic.distributedCache = &DistributedCache{
			URL:         config.DistributedCacheURL,
			TTL:         config.DistributedCacheTTL,
			Pool:        config.DistributedCachePool,
			Connections: make(map[string]*DistributedConnection),
		}
		ic.initializeDistributedCache()
	}

	// Initialize cache manager
	ic.manager = &CacheManager{
		Strategy:      config.InvalidationStrategy,
		Invalidations: make(map[string]*InvalidationInfo),
		Optimizations: make(map[string]*OptimizationInfo),
	}

	// Initialize cache monitor
	ic.monitor = &CacheMonitor{
		Metrics:    &CacheMetrics{},
		Alerts:     make([]*CacheAlert, 0),
		Thresholds: make(map[string]float64),
	}
	ic.monitor.Thresholds["hit_rate"] = config.HitRateThreshold
	ic.monitor.Thresholds["memory_usage"] = 0.9
	ic.monitor.Thresholds["disk_usage"] = 0.9

	// Initialize cache warmer
	ic.warmer = &CacheWarmer{
		Strategy:     config.WarmingStrategy,
		WarmingQueue: make([]*WarmingTask, 0),
		WarmingStats: make(map[string]*WarmingStats),
	}

	// Start background workers
	ic.startBackgroundWorkers()

	return ic
}

// startBackgroundWorkers starts background cache management workers
func (ic *IntelligentCache) startBackgroundWorkers() {
	// Performance monitoring worker
	go ic.performanceMonitoringWorker()

	// Cache warming worker
	go ic.cacheWarmingWorker()

	// Cache optimization worker
	go ic.cacheOptimizationWorker()

	// Cache cleanup worker
	go ic.cacheCleanupWorker()

	// Cache invalidation worker
	go ic.cacheInvalidationWorker()
}

// performanceMonitoringWorker monitors cache performance
func (ic *IntelligentCache) performanceMonitoringWorker() {
	ticker := time.NewTicker(ic.config.PerformanceInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ic.ctx.Done():
			return
		case <-ticker.C:
			ic.updatePerformanceMetrics()
		}
	}
}

// cacheWarmingWorker manages cache warming
func (ic *IntelligentCache) cacheWarmingWorker() {
	ticker := time.NewTicker(ic.config.WarmingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ic.ctx.Done():
			return
		case <-ticker.C:
			ic.performCacheWarming()
		}
	}
}

// cacheOptimizationWorker manages cache optimization
func (ic *IntelligentCache) cacheOptimizationWorker() {
	ticker := time.NewTicker(ic.config.OptimizationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ic.ctx.Done():
			return
		case <-ticker.C:
			ic.performCacheOptimization()
		}
	}
}

// cacheCleanupWorker manages cache cleanup
func (ic *IntelligentCache) cacheCleanupWorker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ic.ctx.Done():
			return
		case <-ticker.C:
			ic.performCacheCleanup()
		}
	}
}

// cacheInvalidationWorker manages cache invalidation
func (ic *IntelligentCache) cacheInvalidationWorker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ic.ctx.Done():
			return
		case <-ticker.C:
			ic.performCacheInvalidation()
		}
	}
}

// Get retrieves a value from cache
func (ic *IntelligentCache) Get(ctx context.Context, key string) (interface{}, bool) {
	ctx, span := ic.tracer.Start(ctx, "IntelligentCache.Get")
	defer span.End()

	span.SetAttributes(attribute.String("cache_key", key))

	// Try memory cache first
	if value, found := ic.getFromMemory(key); found {
		ic.updateMetrics("memory_hit", 1)
		span.SetAttributes(attribute.String("cache_level", "memory"))
		return value, true
	}

	// Try disk cache
	if ic.diskCache != nil {
		if value, found := ic.getFromDisk(key); found {
			ic.updateMetrics("disk_hit", 1)
			span.SetAttributes(attribute.String("cache_level", "disk"))
			return value, true
		}
	}

	// Try distributed cache
	if ic.distributedCache != nil {
		if value, found := ic.getFromDistributed(ctx, key); found {
			ic.updateMetrics("distributed_hit", 1)
			span.SetAttributes(attribute.String("cache_level", "distributed"))
			return value, true
		}
	}

	// Cache miss
	ic.updateMetrics("miss", 1)
	span.SetAttributes(attribute.String("cache_level", "miss"))
	return nil, false
}

// Set stores a value in cache
func (ic *IntelligentCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	ctx, span := ic.tracer.Start(ctx, "IntelligentCache.Set")
	defer span.End()

	span.SetAttributes(
		attribute.String("cache_key", key),
		attribute.String("ttl", ttl.String()),
	)

	// Store in memory cache
	if err := ic.setInMemory(key, value, ttl); err != nil {
		return fmt.Errorf("failed to set in memory cache: %w", err)
	}

	// Store in disk cache if enabled
	if ic.diskCache != nil {
		if err := ic.setInDisk(key, value, ttl); err != nil {
			ic.logger.Warn("failed to set in disk cache", map[string]interface{}{
				"key":   key,
				"error": err.Error(),
			})
		}
	}

	// Store in distributed cache if enabled
	if ic.distributedCache != nil {
		if err := ic.setInDistributed(ctx, key, value, ttl); err != nil {
			ic.logger.Warn("failed to set in distributed cache", map[string]interface{}{
				"key":   key,
				"error": err.Error(),
			})
		}
	}

	return nil
}

// Delete removes a value from cache
func (ic *IntelligentCache) Delete(ctx context.Context, key string) error {
	ctx, span := ic.tracer.Start(ctx, "IntelligentCache.Delete")
	defer span.End()

	span.SetAttributes(attribute.String("cache_key", key))

	// Delete from memory cache
	ic.deleteFromMemory(key)

	// Delete from disk cache
	if ic.diskCache != nil {
		ic.deleteFromDisk(key)
	}

	// Delete from distributed cache
	if ic.distributedCache != nil {
		ic.deleteFromDistributed(ctx, key)
	}

	return nil
}

// Invalidate invalidates cache entries matching a pattern
func (ic *IntelligentCache) Invalidate(ctx context.Context, pattern string) error {
	ctx, span := ic.tracer.Start(ctx, "IntelligentCache.Invalidate")
	defer span.End()

	span.SetAttributes(attribute.String("pattern", pattern))

	ic.manager.Mux.Lock()
	defer ic.manager.Mux.Unlock()

	// Record invalidation
	ic.manager.Invalidations[pattern] = &InvalidationInfo{
		Pattern:           pattern,
		LastInvalidated:   time.Now(),
		InvalidationCount: 1,
		AffectedKeys:      make([]string, 0),
	}

	// Invalidate from all cache levels
	affectedKeys := ic.invalidateFromMemory(pattern)
	ic.manager.Invalidations[pattern].AffectedKeys = affectedKeys

	if ic.diskCache != nil {
		diskKeys := ic.invalidateFromDisk(pattern)
		ic.manager.Invalidations[pattern].AffectedKeys = append(
			ic.manager.Invalidations[pattern].AffectedKeys, diskKeys...)
	}

	if ic.distributedCache != nil {
		distributedKeys := ic.invalidateFromDistributed(ctx, pattern)
		ic.manager.Invalidations[pattern].AffectedKeys = append(
			ic.manager.Invalidations[pattern].AffectedKeys, distributedKeys...)
	}

	ic.updateMetrics("invalidations", 1)

	return nil
}

// Warm warms the cache with frequently accessed data
func (ic *IntelligentCache) Warm(ctx context.Context, keys []string) error {
	ctx, span := ic.tracer.Start(ctx, "IntelligentCache.Warm")
	defer span.End()

	span.SetAttributes(attribute.Int("key_count", len(keys)))

	ic.warmer.Mux.Lock()
	defer ic.warmer.Mux.Unlock()

	for _, key := range keys {
		task := &WarmingTask{
			ID:          generateTaskID(),
			Key:         key,
			Priority:    1,
			CreatedAt:   time.Now(),
			Attempts:    0,
			MaxAttempts: 3,
		}
		ic.warmer.WarmingQueue = append(ic.warmer.WarmingQueue, task)
	}

	return nil
}

// GetMetrics returns cache performance metrics
func (ic *IntelligentCache) GetMetrics() *CacheMetrics {
	ic.monitor.Mux.RLock()
	defer ic.monitor.Mux.RUnlock()

	return ic.monitor.Metrics
}

// GetAlerts returns cache alerts
func (ic *IntelligentCache) GetAlerts() []*CacheAlert {
	ic.monitor.Mux.RLock()
	defer ic.monitor.Mux.RUnlock()

	return ic.monitor.Alerts
}

// Memory cache operations

func (ic *IntelligentCache) getFromMemory(key string) (interface{}, bool) {
	ic.memoryCache.mu.RLock()
	defer ic.memoryCache.mu.RUnlock()

	entry, exists := ic.memoryCache.data[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	// Update access information
	entry.LastAccess = time.Now()
	entry.AccessCount++

	// Update access order for LRU
	ic.updateAccessOrder(key)

	return entry.Value, true
}

func (ic *IntelligentCache) setInMemory(key string, value interface{}, ttl time.Duration) error {
	ic.memoryCache.mu.Lock()
	defer ic.memoryCache.mu.Unlock()

	// Check if we need to evict entries
	if len(ic.memoryCache.data) >= ic.memoryCache.config.MaxSize {
		ic.evictFromMemory()
	}

	// Create cache item for memory cache
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	entry := &cacheItem{
		Value:       valueBytes,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(ttl),
		TTL:         ttl,
		Size:        int64(len(valueBytes)),
		AccessCount: 1,
		LastAccess:  time.Now(),
	}

	ic.memoryCache.data[key] = entry
	ic.updateAccessOrder(key)

	return nil
}

func (ic *IntelligentCache) deleteFromMemory(key string) {
	ic.memoryCache.mu.Lock()
	defer ic.memoryCache.mu.Unlock()

	delete(ic.memoryCache.data, key)
	ic.removeFromAccessOrder(key)
}

func (ic *IntelligentCache) evictFromMemory() {
	switch ic.config.MemoryEvictionPolicy {
	case "lru":
		ic.evictLRU()
	case "lfu":
		ic.evictLFU()
	case "fifo":
		ic.evictFIFO()
	default:
		ic.evictLRU()
	}
}

func (ic *IntelligentCache) evictLRU() {
	if len(ic.memoryCache.data) == 0 {
		return
	}

	// Find least recently used item
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, item := range ic.memoryCache.data {
		if first || item.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.LastAccess
			first = false
		}
	}

	if oldestKey != "" {
		delete(ic.memoryCache.data, oldestKey)
	}
}

func (ic *IntelligentCache) evictLFU() {
	var leastFrequentKey string
	var minAccessCount int64 = 1<<63 - 1

	for key, entry := range ic.memoryCache.data {
		if entry.AccessCount < minAccessCount {
			minAccessCount = entry.AccessCount
			leastFrequentKey = key
		}
	}

	if leastFrequentKey != "" {
		delete(ic.memoryCache.data, leastFrequentKey)
	}
}

func (ic *IntelligentCache) evictFIFO() {
	if len(ic.memoryCache.data) == 0 {
		return
	}

	// Find oldest item by creation time
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, item := range ic.memoryCache.data {
		if first || item.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.CreatedAt
			first = false
		}
	}

	if oldestKey != "" {
		delete(ic.memoryCache.data, oldestKey)
	}
}

func (ic *IntelligentCache) updateAccessOrder(key string) {
	// Update access time for LRU tracking
	if item, exists := ic.memoryCache.data[key]; exists {
		item.LastAccess = time.Now()
	}
}

func (ic *IntelligentCache) removeFromAccessOrder(key string) {
	// No-op since we're not maintaining a separate access order list
	// Access order is tracked via LastAccess field in cacheItem
}

func (ic *IntelligentCache) invalidateFromMemory(pattern string) []string {
	ic.memoryCache.mu.Lock()
	defer ic.memoryCache.mu.Unlock()

	var affectedKeys []string
	for key := range ic.memoryCache.data {
		if ic.matchesPattern(key, pattern) {
			delete(ic.memoryCache.data, key)
			ic.removeFromAccessOrder(key)
			affectedKeys = append(affectedKeys, key)
		}
	}

	return affectedKeys
}

// Disk cache operations

func (ic *IntelligentCache) initializeDiskCache() error {
	// Create cache directory
	if err := os.MkdirAll(ic.diskCache.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Load existing index
	return ic.loadDiskIndex()
}

func (ic *IntelligentCache) loadDiskIndex() error {
	indexFile := filepath.Join(ic.diskCache.basePath, "index.json")

	data, err := os.ReadFile(indexFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Index doesn't exist yet
		}
		return fmt.Errorf("failed to read index file: %w", err)
	}

	return json.Unmarshal(data, &ic.diskIndex)
}

func (ic *IntelligentCache) saveDiskIndex() error {
	ic.diskMux.RLock()
	defer ic.diskMux.RUnlock()

	data, err := json.Marshal(ic.diskIndex)
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	indexFile := filepath.Join(ic.diskCache.basePath, "index.json")
	return os.WriteFile(indexFile, data, 0644)
}

func (ic *IntelligentCache) getFromDisk(key string) (interface{}, bool) {
	ic.diskMux.RLock()
	entry, exists := ic.diskIndex[key]
	ic.diskMux.RUnlock()

	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		ic.deleteFromDisk(key)
		return nil, false
	}

	// Construct file path from key
	filePath := filepath.Join(ic.diskCache.basePath, key+".cache")

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		ic.logger.Warn("failed to read disk cache file", map[string]interface{}{
			"key":   key,
			"file":  filePath,
			"error": err.Error(),
		})
		return nil, false
	}

	// Decompress if needed (simplified - assume compression based on config)
	if ic.config.DiskCompression {
		data = ic.decompress(data)
	}

	// Deserialize
	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		ic.logger.Warn("failed to deserialize disk cache value", map[string]interface{}{
			"key":   key,
			"error": err.Error(),
		})
		return nil, false
	}

	// Update access information
	ic.diskMux.Lock()
	entry.LastAccess = time.Now()
	entry.AccessCount++
	ic.diskMux.Unlock()

	return value, true
}

func (ic *IntelligentCache) setInDisk(key string, value interface{}, ttl time.Duration) error {
	// Serialize value
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}

	// Compress if enabled
	if ic.config.DiskCompression {
		data = ic.compress(data)
	}

	// Generate file path
	hash := fmt.Sprintf("%x", md5.Sum([]byte(key)))
	fileName := fmt.Sprintf("%s.cache", hash)
	filePath := filepath.Join(ic.diskCache.basePath, fileName)

	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	// Update index
	ic.diskMux.Lock()
	defer ic.diskMux.Unlock()

	ic.diskIndex[key] = &cacheItem{
		Value:       data,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(ttl),
		TTL:         ttl,
		Size:        int64(len(data)),
		AccessCount: 1,
		LastAccess:  time.Now(),
	}

	return ic.saveDiskIndex()
}

func (ic *IntelligentCache) deleteFromDisk(key string) {
	ic.diskMux.Lock()
	defer ic.diskMux.Unlock()

	_, exists := ic.diskIndex[key]
	if !exists {
		return
	}

	// Remove file
	filePath := filepath.Join(ic.diskCache.basePath, key+".cache")
	os.Remove(filePath)

	// Remove from index
	delete(ic.diskIndex, key)
	ic.saveDiskIndex()
}

func (ic *IntelligentCache) invalidateFromDisk(pattern string) []string {
	ic.diskMux.Lock()
	defer ic.diskMux.Unlock()

	var affectedKeys []string
	for key := range ic.diskIndex {
		if ic.matchesPattern(key, pattern) {
			filePath := filepath.Join(ic.diskCache.basePath, key+".cache")
			os.Remove(filePath)
			delete(ic.diskIndex, key)
			affectedKeys = append(affectedKeys, key)
		}
	}

	ic.saveDiskIndex()
	return affectedKeys
}

// Distributed cache operations (simplified implementation)

func (ic *IntelligentCache) initializeDistributedCache() error {
	// Simplified implementation - in production, connect to Redis/Memcached
	ic.logger.Info("distributed cache initialized", map[string]interface{}{
		"url": ic.distributedCache.URL,
	})
	return nil
}

func (ic *IntelligentCache) getFromDistributed(ctx context.Context, key string) (interface{}, bool) {
	// Simplified implementation - in production, use Redis/Memcached client
	return nil, false
}

func (ic *IntelligentCache) setInDistributed(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Simplified implementation - in production, use Redis/Memcached client
	return nil
}

func (ic *IntelligentCache) deleteFromDistributed(ctx context.Context, key string) {
	// Simplified implementation - in production, use Redis/Memcached client
}

func (ic *IntelligentCache) invalidateFromDistributed(ctx context.Context, pattern string) []string {
	// Simplified implementation - in production, use Redis/Memcached client
	return []string{}
}

// Cache warming operations

func (ic *IntelligentCache) performCacheWarming() {
	ic.warmer.Mux.Lock()
	defer ic.warmer.Mux.Unlock()

	if len(ic.warmer.WarmingQueue) == 0 {
		return
	}

	// Sort by priority
	sort.Slice(ic.warmer.WarmingQueue, func(i, j int) bool {
		return ic.warmer.WarmingQueue[i].Priority > ic.warmer.WarmingQueue[j].Priority
	})

	// Process batch
	batchSize := ic.config.WarmingBatchSize
	if batchSize > len(ic.warmer.WarmingQueue) {
		batchSize = len(ic.warmer.WarmingQueue)
	}

	for i := 0; i < batchSize; i++ {
		task := ic.warmer.WarmingQueue[i]
		task.Attempts++
		task.LastAttempt = time.Now()

		// Simulate warming (in production, fetch actual data)
		ic.logger.Info("warming cache", map[string]interface{}{
			"key":     task.Key,
			"attempt": task.Attempts,
		})

		// Remove from queue
		ic.warmer.WarmingQueue = ic.warmer.WarmingQueue[1:]
	}

	ic.warmer.LastWarming = time.Now()
}

// Cache optimization operations

func (ic *IntelligentCache) performCacheOptimization() {
	ic.manager.Mux.Lock()
	defer ic.manager.Mux.Unlock()

	// Optimize memory cache
	ic.optimizeMemoryCache()

	// Optimize disk cache
	if ic.diskCache != nil {
		ic.optimizeDiskCache()
	}

	ic.manager.LastOptimization = time.Now()
}

func (ic *IntelligentCache) optimizeMemoryCache() {
	// Remove expired entries
	ic.memoryCache.mu.Lock()
	defer ic.memoryCache.mu.Unlock()

	now := time.Now()
	for key, entry := range ic.memoryCache.data {
		if now.After(entry.ExpiresAt) {
			delete(ic.memoryCache.data, key)
			ic.removeFromAccessOrder(key)
		}
	}
}

func (ic *IntelligentCache) optimizeDiskCache() {
	// Remove expired entries
	ic.diskMux.Lock()
	defer ic.diskMux.Unlock()

	now := time.Now()
	for key, entry := range ic.diskIndex {
		if now.After(entry.ExpiresAt) {
			filePath := filepath.Join(ic.diskCache.basePath, key+".cache")
			os.Remove(filePath)
			delete(ic.diskIndex, key)
		}
	}

	ic.saveDiskIndex()
}

// Cache cleanup operations

func (ic *IntelligentCache) performCacheCleanup() {
	// Clean up expired entries
	ic.cleanupExpiredEntries()

	// Clean up old alerts
	ic.cleanupOldAlerts()
}

func (ic *IntelligentCache) cleanupExpiredEntries() {
	// Memory cache cleanup
	ic.memoryCache.mu.Lock()
	now := time.Now()
	for key, entry := range ic.memoryCache.data {
		if now.After(entry.ExpiresAt) {
			delete(ic.memoryCache.data, key)
			ic.removeFromAccessOrder(key)
		}
	}
	ic.memoryCache.mu.Unlock()

	// Disk cache cleanup
	if ic.diskCache != nil {
		ic.diskMux.Lock()
		for key, entry := range ic.diskIndex {
			if now.After(entry.ExpiresAt) {
				filePath := filepath.Join(ic.diskCache.basePath, key+".cache")
				os.Remove(filePath)
				delete(ic.diskIndex, key)
			}
		}
		ic.diskMux.Unlock()
		ic.saveDiskIndex()
	}
}

func (ic *IntelligentCache) cleanupOldAlerts() {
	ic.monitor.Mux.Lock()
	defer ic.monitor.Mux.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour) // Keep alerts for 24 hours
	newAlerts := make([]*CacheAlert, 0)

	for _, alert := range ic.monitor.Alerts {
		if alert.Timestamp.After(cutoff) {
			newAlerts = append(newAlerts, alert)
		}
	}

	ic.monitor.Alerts = newAlerts
}

// Cache invalidation operations

func (ic *IntelligentCache) performCacheInvalidation() {
	// Process pending invalidations
	ic.manager.Mux.Lock()
	defer ic.manager.Mux.Unlock()

	for pattern, info := range ic.manager.Invalidations {
		if time.Since(info.LastInvalidated) > ic.config.InvalidationCooldown {
			// Process invalidation
			ic.logger.Info("processing cache invalidation", map[string]interface{}{
				"pattern":       pattern,
				"affected_keys": len(info.AffectedKeys),
			})
		}
	}
}

// Performance monitoring operations

func (ic *IntelligentCache) updatePerformanceMetrics() {
	ic.monitor.Mux.Lock()
	defer ic.monitor.Mux.Unlock()

	// Calculate hit rate
	total := ic.monitor.Metrics.TotalHits + ic.monitor.Metrics.TotalMisses
	if total > 0 {
		ic.monitor.Metrics.HitRate = float64(ic.monitor.Metrics.TotalHits) / float64(total)
	}

	// Check for alerts
	if ic.monitor.Metrics.HitRate < ic.monitor.Thresholds["hit_rate"] {
		ic.createAlert("hit_rate", "low", fmt.Sprintf("Cache hit rate is %.2f%%", ic.monitor.Metrics.HitRate*100), ic.monitor.Metrics.HitRate, ic.monitor.Thresholds["hit_rate"])
	}

	ic.monitor.Metrics.LastUpdate = time.Now()
}

func (ic *IntelligentCache) updateMetrics(metric string, value int64) {
	ic.monitor.Mux.Lock()
	defer ic.monitor.Mux.Unlock()

	switch metric {
	case "memory_hit":
		ic.monitor.Metrics.MemoryHits += value
		ic.monitor.Metrics.TotalHits += value
	case "disk_hit":
		ic.monitor.Metrics.DiskHits += value
		ic.monitor.Metrics.TotalHits += value
	case "distributed_hit":
		ic.monitor.Metrics.DistributedHits += value
		ic.monitor.Metrics.TotalHits += value
	case "miss":
		ic.monitor.Metrics.TotalMisses += value
	case "invalidations":
		ic.monitor.Metrics.Invalidations += value
	}
}

func (ic *IntelligentCache) createAlert(alertType, severity, message string, value, threshold float64) {
	// Check cooldown
	if time.Since(ic.monitor.LastAlert) < 5*time.Minute {
		return
	}

	alert := &CacheAlert{
		ID:        fmt.Sprintf("cache-alert-%d", time.Now().Unix()),
		Type:      alertType,
		Severity:  severity,
		Message:   message,
		Metric:    alertType,
		Value:     value,
		Threshold: threshold,
		Timestamp: time.Now(),
	}

	ic.monitor.Alerts = append(ic.monitor.Alerts, alert)
	ic.monitor.LastAlert = time.Now()

	ic.logger.Warn("cache alert created", map[string]interface{}{
		"alert_id":  alert.ID,
		"type":      alert.Type,
		"severity":  alert.Severity,
		"message":   alert.Message,
		"value":     alert.Value,
		"threshold": alert.Threshold,
	})
}

// Utility functions

func (ic *IntelligentCache) matchesPattern(key, pattern string) bool {
	// Simplified pattern matching - in production, use regex
	return key == pattern || pattern == "*"
}

func (ic *IntelligentCache) compress(data []byte) []byte {
	// Simplified compression - in production, use gzip
	return data
}

func (ic *IntelligentCache) decompress(data []byte) []byte {
	// Simplified decompression - in production, use gzip
	return data
}

func generateTaskID() string {
	return fmt.Sprintf("task-%d", time.Now().UnixNano())
}

// Shutdown shuts down the intelligent cache
func (ic *IntelligentCache) Shutdown() {
	ic.cancel()

	// Save disk index
	if ic.diskCache != nil {
		ic.saveDiskIndex()
	}

	ic.logger.Info("intelligent cache shutting down", map[string]interface{}{})
}
