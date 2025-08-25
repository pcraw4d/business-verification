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

	"github.com/pcraw4d/business-verification/internal/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// IntelligentCache provides multi-level caching with intelligent optimization
type IntelligentCache struct {
	// Configuration
	config *CacheConfig

	// Observability
	logger *observability.Logger
	tracer trace.Tracer

	// Cache levels
	memoryCache      *MemoryCache
	diskCache        *DiskCache
	distributedCache *DistributedCache

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

// CacheConfig configuration for intelligent caching
type CacheConfig struct {
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

// MemoryCache provides in-memory caching
type MemoryCache struct {
	Data           map[string]*CacheEntry
	Size           int
	TTL            time.Duration
	EvictionPolicy string
	AccessOrder    []string
	Mux            sync.RWMutex
}

// DiskCache provides disk-based caching
type DiskCache struct {
	Path        string
	Size        int64
	TTL         time.Duration
	Compression bool
	Index       map[string]*DiskEntry
	Mux         sync.RWMutex
}

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
func NewIntelligentCache(config *CacheConfig, logger *observability.Logger, tracer trace.Tracer) *IntelligentCache {
	if config == nil {
		config = &CacheConfig{
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
		config: config,
		logger: logger,
		tracer: tracer,
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize cache levels
	ic.memoryCache = &MemoryCache{
		Data:           make(map[string]*CacheEntry),
		Size:           config.MemoryCacheSize,
		TTL:            config.MemoryCacheTTL,
		EvictionPolicy: config.MemoryEvictionPolicy,
		AccessOrder:    make([]string, 0),
	}

	if config.DiskCacheEnabled {
		ic.diskCache = &DiskCache{
			Path:        config.DiskCachePath,
			Size:        config.DiskCacheSize,
			TTL:         config.DiskCacheTTL,
			Compression: config.DiskCompression,
			Index:       make(map[string]*DiskEntry),
		}
		ic.initializeDiskCache()
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
	ic.memoryCache.Mux.RLock()
	defer ic.memoryCache.Mux.RUnlock()

	entry, exists := ic.memoryCache.Data[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	// Update access information
	entry.LastAccessed = time.Now()
	entry.AccessCount++

	// Update access order for LRU
	ic.updateAccessOrder(key)

	return entry.Value, true
}

func (ic *IntelligentCache) setInMemory(key string, value interface{}, ttl time.Duration) error {
	ic.memoryCache.Mux.Lock()
	defer ic.memoryCache.Mux.Unlock()

	// Check if we need to evict entries
	if len(ic.memoryCache.Data) >= ic.memoryCache.Size {
		ic.evictFromMemory()
	}

	entry := &CacheEntry{
		Key:          key,
		Value:        value,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(ttl),
		LastAccessed: time.Now(),
		AccessCount:  1,
		Metadata:     make(map[string]interface{}),
	}

	ic.memoryCache.Data[key] = entry
	ic.updateAccessOrder(key)

	return nil
}

func (ic *IntelligentCache) deleteFromMemory(key string) {
	ic.memoryCache.Mux.Lock()
	defer ic.memoryCache.Mux.Unlock()

	delete(ic.memoryCache.Data, key)
	ic.removeFromAccessOrder(key)
}

func (ic *IntelligentCache) evictFromMemory() {
	switch ic.memoryCache.EvictionPolicy {
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
	if len(ic.memoryCache.AccessOrder) == 0 {
		return
	}

	// Remove least recently used
	key := ic.memoryCache.AccessOrder[0]
	delete(ic.memoryCache.Data, key)
	ic.memoryCache.AccessOrder = ic.memoryCache.AccessOrder[1:]
}

func (ic *IntelligentCache) evictLFU() {
	var leastFrequentKey string
	var minAccessCount int64 = 1<<63 - 1

	for key, entry := range ic.memoryCache.Data {
		if entry.AccessCount < minAccessCount {
			minAccessCount = entry.AccessCount
			leastFrequentKey = key
		}
	}

	if leastFrequentKey != "" {
		delete(ic.memoryCache.Data, leastFrequentKey)
		ic.removeFromAccessOrder(leastFrequentKey)
	}
}

func (ic *IntelligentCache) evictFIFO() {
	if len(ic.memoryCache.AccessOrder) == 0 {
		return
	}

	// Remove first in (oldest)
	key := ic.memoryCache.AccessOrder[0]
	delete(ic.memoryCache.Data, key)
	ic.memoryCache.AccessOrder = ic.memoryCache.AccessOrder[1:]
}

func (ic *IntelligentCache) updateAccessOrder(key string) {
	ic.removeFromAccessOrder(key)
	ic.memoryCache.AccessOrder = append(ic.memoryCache.AccessOrder, key)
}

func (ic *IntelligentCache) removeFromAccessOrder(key string) {
	for i, k := range ic.memoryCache.AccessOrder {
		if k == key {
			ic.memoryCache.AccessOrder = append(ic.memoryCache.AccessOrder[:i], ic.memoryCache.AccessOrder[i+1:]...)
			break
		}
	}
}

func (ic *IntelligentCache) invalidateFromMemory(pattern string) []string {
	ic.memoryCache.Mux.Lock()
	defer ic.memoryCache.Mux.Unlock()

	var affectedKeys []string
	for key := range ic.memoryCache.Data {
		if ic.matchesPattern(key, pattern) {
			delete(ic.memoryCache.Data, key)
			ic.removeFromAccessOrder(key)
			affectedKeys = append(affectedKeys, key)
		}
	}

	return affectedKeys
}

// Disk cache operations

func (ic *IntelligentCache) initializeDiskCache() error {
	// Create cache directory
	if err := os.MkdirAll(ic.diskCache.Path, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Load existing index
	return ic.loadDiskIndex()
}

func (ic *IntelligentCache) loadDiskIndex() error {
	indexFile := filepath.Join(ic.diskCache.Path, "index.json")

	data, err := os.ReadFile(indexFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Index doesn't exist yet
		}
		return fmt.Errorf("failed to read index file: %w", err)
	}

	return json.Unmarshal(data, &ic.diskCache.Index)
}

func (ic *IntelligentCache) saveDiskIndex() error {
	ic.diskCache.Mux.RLock()
	defer ic.diskCache.Mux.RUnlock()

	data, err := json.Marshal(ic.diskCache.Index)
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	indexFile := filepath.Join(ic.diskCache.Path, "index.json")
	return os.WriteFile(indexFile, data, 0644)
}

func (ic *IntelligentCache) getFromDisk(key string) (interface{}, bool) {
	ic.diskCache.Mux.RLock()
	entry, exists := ic.diskCache.Index[key]
	ic.diskCache.Mux.RUnlock()

	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		ic.deleteFromDisk(key)
		return nil, false
	}

	// Read file
	data, err := os.ReadFile(entry.FilePath)
	if err != nil {
		ic.logger.Warn("failed to read disk cache file", map[string]interface{}{
			"key":   key,
			"file":  entry.FilePath,
			"error": err.Error(),
		})
		return nil, false
	}

	// Decompress if needed
	if entry.Compressed {
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
	ic.diskCache.Mux.Lock()
	entry.LastAccessed = time.Now()
	entry.AccessCount++
	ic.diskCache.Mux.Unlock()

	return value, true
}

func (ic *IntelligentCache) setInDisk(key string, value interface{}, ttl time.Duration) error {
	// Serialize value
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}

	// Compress if enabled
	if ic.diskCache.Compression {
		data = ic.compress(data)
	}

	// Generate file path
	hash := fmt.Sprintf("%x", md5.Sum([]byte(key)))
	fileName := fmt.Sprintf("%s.cache", hash)
	filePath := filepath.Join(ic.diskCache.Path, fileName)

	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	// Update index
	ic.diskCache.Mux.Lock()
	defer ic.diskCache.Mux.Unlock()

	ic.diskCache.Index[key] = &DiskEntry{
		Key:          key,
		FilePath:     filePath,
		Size:         int64(len(data)),
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(ttl),
		LastAccessed: time.Now(),
		AccessCount:  1,
		Compressed:   ic.diskCache.Compression,
		Checksum:     fmt.Sprintf("%x", md5.Sum(data)),
	}

	return ic.saveDiskIndex()
}

func (ic *IntelligentCache) deleteFromDisk(key string) {
	ic.diskCache.Mux.Lock()
	defer ic.diskCache.Mux.Unlock()

	entry, exists := ic.diskCache.Index[key]
	if !exists {
		return
	}

	// Remove file
	os.Remove(entry.FilePath)

	// Remove from index
	delete(ic.diskCache.Index, key)
	ic.saveDiskIndex()
}

func (ic *IntelligentCache) invalidateFromDisk(pattern string) []string {
	ic.diskCache.Mux.Lock()
	defer ic.diskCache.Mux.Unlock()

	var affectedKeys []string
	for key, entry := range ic.diskCache.Index {
		if ic.matchesPattern(key, pattern) {
			os.Remove(entry.FilePath)
			delete(ic.diskCache.Index, key)
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
	ic.memoryCache.Mux.Lock()
	defer ic.memoryCache.Mux.Unlock()

	now := time.Now()
	for key, entry := range ic.memoryCache.Data {
		if now.After(entry.ExpiresAt) {
			delete(ic.memoryCache.Data, key)
			ic.removeFromAccessOrder(key)
		}
	}
}

func (ic *IntelligentCache) optimizeDiskCache() {
	// Remove expired entries
	ic.diskCache.Mux.Lock()
	defer ic.diskCache.Mux.Unlock()

	now := time.Now()
	for key, entry := range ic.diskCache.Index {
		if now.After(entry.ExpiresAt) {
			os.Remove(entry.FilePath)
			delete(ic.diskCache.Index, key)
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
	ic.memoryCache.Mux.Lock()
	now := time.Now()
	for key, entry := range ic.memoryCache.Data {
		if now.After(entry.ExpiresAt) {
			delete(ic.memoryCache.Data, key)
			ic.removeFromAccessOrder(key)
		}
	}
	ic.memoryCache.Mux.Unlock()

	// Disk cache cleanup
	if ic.diskCache != nil {
		ic.diskCache.Mux.Lock()
		for key, entry := range ic.diskCache.Index {
			if now.After(entry.ExpiresAt) {
				os.Remove(entry.FilePath)
				delete(ic.diskCache.Index, key)
			}
		}
		ic.diskCache.Mux.Unlock()
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

	ic.logger.Info("intelligent cache shutting down")
}
