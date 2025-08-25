package caching

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CacheType represents the type of cache
type CacheType string

const (
	CacheTypeLRU         CacheType = "lru"         // Least Recently Used
	CacheTypeLFU         CacheType = "lfu"         // Least Frequently Used
	CacheTypeARC         CacheType = "arc"         // Adaptive Replacement Cache
	CacheTypeTTL         CacheType = "ttl"         // Time To Live
	CacheTypeFIFO        CacheType = "fifo"        // First In First Out
	CacheTypeLIRS        CacheType = "lirs"        // Low Inter-reference Recency Set
	CacheType2Q          CacheType = "2q"          // 2Q Cache
	CacheTypeClock       CacheType = "clock"       // Clock (Second Chance) Algorithm
	CacheTypeRandom      CacheType = "random"      // Random Eviction
	CacheTypeIntelligent CacheType = "intelligent" // Intelligent Adaptive Cache
)

// CacheEntry represents a cache entry
type CacheEntry struct {
	Key         string
	Value       interface{}
	Size        int64
	AccessCount int64
	LastAccess  time.Time
	CreatedAt   time.Time
	ExpiresAt   *time.Time
	Priority    int
	Tags        []string
	Metadata    map[string]interface{}
	mu          sync.RWMutex
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits              int64
	Misses            int64
	Evictions         int64
	Expirations       int64
	TotalSize         int64
	EntryCount        int64
	HitRate           float64
	MissRate          float64
	AverageAccessTime time.Duration
	LastReset         time.Time
	mu                sync.RWMutex
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	Type            CacheType
	MaxSize         int64
	MaxEntries      int64
	DefaultTTL      time.Duration
	CleanupInterval time.Duration
	EvictionPolicy  EvictionPolicy
	Compression     bool
	Persistence     bool
	PersistencePath string
	ShardCount      int
	EnableStats     bool
	EnableMetrics   bool
	Logger          *zap.Logger
}

// EvictionPolicy represents cache eviction policy
type EvictionPolicy struct {
	Type              string
	MaxMemoryUsage    int64
	MaxEntryAge       time.Duration
	MaxAccessCount    int64
	PriorityThreshold int
	AdaptiveThreshold bool
}

// IntelligentCache represents an intelligent cache
type IntelligentCache struct {
	config CacheConfig
	shards []*CacheShard
	stats  *CacheStats
	logger *zap.Logger
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

// CacheShard represents a cache shard
type CacheShard struct {
	entries map[string]*CacheEntry
	policy  EvictionPolicy
	stats   *CacheStats
	mu      sync.RWMutex
	index   int
}

// CacheResult represents a cache operation result
type CacheResult struct {
	Value       interface{}
	Found       bool
	Expired     bool
	Size        int64
	AccessCount int64
	LastAccess  time.Time
	Error       error
}

// CacheOperation represents a cache operation
type CacheOperation struct {
	Type      string
	Key       string
	Value     interface{}
	Size      int64
	Timestamp time.Time
	Duration  time.Duration
	Error     error
}

// CacheAnalytics represents cache analytics
type CacheAnalytics struct {
	HitRate           float64
	MissRate          float64
	EvictionRate      float64
	ExpirationRate    float64
	AverageEntrySize  int64
	AverageAccessTime time.Duration
	PopularKeys       []string
	HotKeys           []string
	ColdKeys          []string
	AccessPatterns    map[string]int64
	SizeDistribution  map[string]int64
	LastUpdated       time.Time
}

// NewIntelligentCache creates a new intelligent cache
func NewIntelligentCache(config CacheConfig) (*IntelligentCache, error) {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}

	if config.MaxSize <= 0 {
		config.MaxSize = 100 * 1024 * 1024 // 100MB default
	}

	if config.MaxEntries <= 0 {
		config.MaxEntries = 10000 // 10K entries default
	}

	if config.DefaultTTL <= 0 {
		config.DefaultTTL = 1 * time.Hour // 1 hour default
	}

	if config.CleanupInterval <= 0 {
		config.CleanupInterval = 5 * time.Minute // 5 minutes default
	}

	if config.ShardCount <= 0 {
		config.ShardCount = 16 // 16 shards default
	}

	ctx, cancel := context.WithCancel(context.Background())

	cache := &IntelligentCache{
		config: config,
		stats:  &CacheStats{},
		logger: config.Logger,
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize shards
	cache.shards = make([]*CacheShard, config.ShardCount)
	for i := 0; i < config.ShardCount; i++ {
		cache.shards[i] = &CacheShard{
			entries: make(map[string]*CacheEntry),
			policy:  config.EvictionPolicy,
			stats:   &CacheStats{},
			index:   i,
		}
	}

	// Start background cleanup
	go cache.cleanupWorker()

	// Start analytics collection if enabled
	if config.EnableStats {
		go cache.analyticsWorker()
	}

	return cache, nil
}

// Get retrieves a value from the cache
func (ic *IntelligentCache) Get(key string) *CacheResult {
	start := time.Now()
	shard := ic.getShard(key)

	shard.mu.RLock()
	entry, exists := shard.entries[key]
	shard.mu.RUnlock()

	if !exists {
		ic.recordMiss(shard)
		return &CacheResult{Found: false}
	}

	// Check if entry is expired
	if entry.isExpired() {
		ic.recordExpiration(shard)
		ic.removeEntry(shard, key)
		return &CacheResult{Found: false, Expired: true}
	}

	// Update access statistics
	entry.updateAccess()
	ic.recordHit(shard)

	duration := time.Since(start)
	ic.recordAccessTime(duration)

	return &CacheResult{
		Value:       entry.Value,
		Found:       true,
		Size:        entry.Size,
		AccessCount: entry.AccessCount,
		LastAccess:  entry.LastAccess,
	}
}

// Set stores a value in the cache
func (ic *IntelligentCache) Set(key string, value interface{}, options ...CacheOption) error {
	start := time.Now()
	shard := ic.getShard(key)

	// Apply options
	opts := ic.defaultOptions()
	for _, option := range options {
		option(opts)
	}

	// Calculate entry size
	size := ic.calculateSize(value)
	if size > ic.config.MaxSize {
		return fmt.Errorf("entry size %d exceeds max cache size %d", size, ic.config.MaxSize)
	}

	// Create cache entry
	entry := &CacheEntry{
		Key:         key,
		Value:       value,
		Size:        size,
		AccessCount: 1,
		LastAccess:  time.Now(),
		CreatedAt:   time.Now(),
		Priority:    opts.Priority,
		Tags:        opts.Tags,
		Metadata:    opts.Metadata,
	}

	// Set expiration if specified
	if opts.TTL > 0 {
		expiresAt := time.Now().Add(opts.TTL)
		entry.ExpiresAt = &expiresAt
	} else if ic.config.DefaultTTL > 0 {
		expiresAt := time.Now().Add(ic.config.DefaultTTL)
		entry.ExpiresAt = &expiresAt
	}

	// Check if we need to evict entries
	ic.ensureCapacity(shard, size)

	// Store entry
	shard.mu.Lock()
	shard.entries[key] = entry
	shard.mu.Unlock()

	duration := time.Since(start)
	ic.recordAccessTime(duration)

	return nil
}

// Delete removes a value from the cache
func (ic *IntelligentCache) Delete(key string) bool {
	shard := ic.getShard(key)
	return ic.removeEntry(shard, key)
}

// Clear removes all entries from the cache
func (ic *IntelligentCache) Clear() {
	for _, shard := range ic.shards {
		shard.mu.Lock()
		shard.entries = make(map[string]*CacheEntry)
		shard.mu.Unlock()
	}
	ic.resetStats()
}

// GetStats returns cache statistics
func (ic *IntelligentCache) GetStats() *CacheStats {
	ic.stats.mu.RLock()
	defer ic.stats.mu.RUnlock()

	stats := &CacheStats{}
	*stats = *ic.stats

	// Calculate rates
	total := stats.Hits + stats.Misses
	if total > 0 {
		stats.HitRate = float64(stats.Hits) / float64(total)
		stats.MissRate = float64(stats.Misses) / float64(total)
	}

	return stats
}

// GetAnalytics returns detailed cache analytics
func (ic *IntelligentCache) GetAnalytics() *CacheAnalytics {
	analytics := &CacheAnalytics{
		AccessPatterns:   make(map[string]int64),
		SizeDistribution: make(map[string]int64),
		LastUpdated:      time.Now(),
	}

	var totalHits, totalMisses, totalEvictions, totalExpirations int64
	var totalSize, totalEntries int64
	var totalAccessTime time.Duration
	var accessCount int64

	// Collect data from all shards
	for _, shard := range ic.shards {
		shard.mu.RLock()
		shard.stats.mu.RLock()

		totalHits += shard.stats.Hits
		totalMisses += shard.stats.Misses
		totalEvictions += shard.stats.Evictions
		totalExpirations += shard.stats.Expirations
		totalSize += shard.stats.TotalSize
		totalEntries += shard.stats.EntryCount
		totalAccessTime += shard.stats.AverageAccessTime
		accessCount++

		// Collect entry statistics
		for key, entry := range shard.entries {
			analytics.AccessPatterns[key] = entry.AccessCount
			sizeRange := ic.getSizeRange(entry.Size)
			analytics.SizeDistribution[sizeRange]++
		}

		shard.stats.mu.RUnlock()
		shard.mu.RUnlock()
	}

	// Calculate analytics
	total := totalHits + totalMisses
	if total > 0 {
		analytics.HitRate = float64(totalHits) / float64(total)
		analytics.MissRate = float64(totalMisses) / float64(total)
	}

	if totalEntries > 0 {
		analytics.EvictionRate = float64(totalEvictions) / float64(totalEntries)
		analytics.ExpirationRate = float64(totalExpirations) / float64(totalEntries)
		analytics.AverageEntrySize = totalSize / totalEntries
	}

	if accessCount > 0 {
		analytics.AverageAccessTime = totalAccessTime / time.Duration(accessCount)
	}

	// Identify popular, hot, and cold keys
	analytics.PopularKeys = ic.getPopularKeys(10)
	analytics.HotKeys = ic.getHotKeys(10)
	analytics.ColdKeys = ic.getColdKeys(10)

	return analytics
}

// Close closes the cache and performs cleanup
func (ic *IntelligentCache) Close() error {
	ic.cancel()
	ic.Clear()
	return nil
}

// getShard returns the shard for a given key
func (ic *IntelligentCache) getShard(key string) *CacheShard {
	hash := ic.hashKey(key)
	return ic.shards[hash%len(ic.shards)]
}

// hashKey generates a hash for the key
func (ic *IntelligentCache) hashKey(key string) int {
	hash := 0
	for _, char := range key {
		hash = 31*hash + int(char)
	}
	return hash
}

// calculateSize calculates the size of a value
func (ic *IntelligentCache) calculateSize(value interface{}) int64 {
	// This is a simplified size calculation
	// In a real implementation, you might use reflection or serialization
	switch v := value.(type) {
	case string:
		return int64(len(v))
	case []byte:
		return int64(len(v))
	case int, int32, int64, float32, float64, bool:
		return 8
	default:
		// Default size for complex types
		return 64
	}
}

// ensureCapacity ensures there's enough capacity for a new entry
func (ic *IntelligentCache) ensureCapacity(shard *CacheShard, newSize int64) {
	shard.mu.Lock()
	defer shard.mu.Unlock()

	currentSize := shard.stats.TotalSize
	currentEntries := shard.stats.EntryCount

	// Calculate per-shard limits
	perShardMaxSize := ic.config.MaxSize / int64(len(ic.shards))
	perShardMaxEntries := ic.config.MaxEntries / int64(len(ic.shards))

	// Check if we need to evict entries
	for (currentSize+newSize > perShardMaxSize || currentEntries >= perShardMaxEntries) && len(shard.entries) > 0 {
		keyToEvict := ic.selectEvictionCandidate(shard)
		if keyToEvict == "" {
			break
		}

		entry := shard.entries[keyToEvict]
		currentSize -= entry.Size
		currentEntries--
		delete(shard.entries, keyToEvict)
		ic.recordEviction(shard)
	}

	shard.stats.TotalSize = currentSize
	shard.stats.EntryCount = currentEntries
}

// selectEvictionCandidate selects an entry to evict based on the eviction policy
func (ic *IntelligentCache) selectEvictionCandidate(shard *CacheShard) string {
	switch ic.config.Type {
	case CacheTypeLRU:
		return ic.selectLRUCandidate(shard)
	case CacheTypeLFU:
		return ic.selectLFUCandidate(shard)
	case CacheTypeARC:
		return ic.selectARCCandidate(shard)
	case CacheTypeFIFO:
		return ic.selectFIFOCandidate(shard)
	case CacheTypeLIRS:
		return ic.selectLIRSCandidate(shard)
	case CacheType2Q:
		return ic.select2QCandidate(shard)
	case CacheTypeClock:
		return ic.selectClockCandidate(shard)
	case CacheTypeRandom:
		return ic.selectRandomCandidate(shard)
	case CacheTypeIntelligent:
		return ic.selectIntelligentCandidate(shard)
	default:
		return ic.selectLRUCandidate(shard)
	}
}

// selectLRUCandidate selects the least recently used entry
func (ic *IntelligentCache) selectLRUCandidate(shard *CacheShard) string {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range shard.entries {
		if oldestKey == "" || entry.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastAccess
		}
	}

	return oldestKey
}

// selectLFUCandidate selects the least frequently used entry
func (ic *IntelligentCache) selectLFUCandidate(shard *CacheShard) string {
	var leastFrequentKey string
	var leastFrequentCount int64 = -1

	for key, entry := range shard.entries {
		if leastFrequentKey == "" || entry.AccessCount < leastFrequentCount {
			leastFrequentKey = key
			leastFrequentCount = entry.AccessCount
		}
	}

	return leastFrequentKey
}

// selectARCCandidate selects an entry using Adaptive Replacement Cache algorithm
func (ic *IntelligentCache) selectARCCandidate(shard *CacheShard) string {
	// Simplified ARC implementation
	// In a real implementation, this would maintain T1, T2, B1, B2 lists
	return ic.selectLRUCandidate(shard)
}

// selectFIFOCandidate selects the first in entry
func (ic *IntelligentCache) selectFIFOCandidate(shard *CacheShard) string {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range shard.entries {
		if oldestKey == "" || entry.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.CreatedAt
		}
	}

	return oldestKey
}

// selectLIRSCandidate selects an entry using LIRS algorithm
func (ic *IntelligentCache) selectLIRSCandidate(shard *CacheShard) string {
	// Simplified LIRS implementation
	return ic.selectLRUCandidate(shard)
}

// select2QCandidate selects an entry using 2Q algorithm
func (ic *IntelligentCache) select2QCandidate(shard *CacheShard) string {
	// Simplified 2Q implementation
	return ic.selectLRUCandidate(shard)
}

// selectClockCandidate selects an entry using Clock algorithm
func (ic *IntelligentCache) selectClockCandidate(shard *CacheShard) string {
	// Simplified Clock implementation
	return ic.selectLRUCandidate(shard)
}

// selectRandomCandidate selects a random entry
func (ic *IntelligentCache) selectRandomCandidate(shard *CacheShard) string {
	// This would require maintaining a list of keys for O(1) random selection
	// For now, we'll use a simple approach
	for key := range shard.entries {
		return key
	}
	return ""
}

// selectIntelligentCandidate selects an entry using intelligent adaptive algorithm
func (ic *IntelligentCache) selectIntelligentCandidate(shard *CacheShard) string {
	// Intelligent adaptive algorithm that considers multiple factors:
	// - Access frequency
	// - Recency
	// - Entry size
	// - Priority
	// - Tags and metadata

	var bestKey string
	var bestScore float64 = -1

	for key, entry := range shard.entries {
		score := ic.calculateIntelligentScore(entry)
		if bestKey == "" || score < bestScore {
			bestKey = key
			bestScore = score
		}
	}

	return bestKey
}

// calculateIntelligentScore calculates an intelligent score for eviction
func (ic *IntelligentCache) calculateIntelligentScore(entry *CacheEntry) float64 {
	// Factors to consider:
	// 1. Access frequency (lower is worse)
	// 2. Recency (older is worse)
	// 3. Size (larger is worse)
	// 4. Priority (lower is worse)
	// 5. Age (older is worse)

	now := time.Now()
	age := now.Sub(entry.CreatedAt)
	recency := now.Sub(entry.LastAccess)

	// Normalize factors
	freqScore := 1.0 / float64(entry.AccessCount+1)
	recencyScore := recency.Seconds() / 3600.0 // Hours since last access
	sizeScore := float64(entry.Size) / float64(ic.config.MaxSize)
	priorityScore := float64(10-entry.Priority) / 10.0
	ageScore := age.Seconds() / 86400.0 // Days since creation

	// Weighted combination
	score := 0.3*freqScore + 0.25*recencyScore + 0.2*sizeScore + 0.15*priorityScore + 0.1*ageScore

	return score
}

// removeEntry removes an entry from the cache
func (ic *IntelligentCache) removeEntry(shard *CacheShard, key string) bool {
	shard.mu.Lock()
	defer shard.mu.Unlock()

	entry, exists := shard.entries[key]
	if !exists {
		return false
	}

	shard.stats.TotalSize -= entry.Size
	shard.stats.EntryCount--
	delete(shard.entries, key)

	return true
}

// recordHit records a cache hit
func (ic *IntelligentCache) recordHit(shard *CacheShard) {
	shard.stats.mu.Lock()
	shard.stats.Hits++
	shard.stats.mu.Unlock()

	ic.stats.mu.Lock()
	ic.stats.Hits++
	ic.stats.mu.Unlock()
}

// recordMiss records a cache miss
func (ic *IntelligentCache) recordMiss(shard *CacheShard) {
	shard.stats.mu.Lock()
	shard.stats.Misses++
	shard.stats.mu.Unlock()

	ic.stats.mu.Lock()
	ic.stats.Misses++
	ic.stats.mu.Unlock()
}

// recordEviction records a cache eviction
func (ic *IntelligentCache) recordEviction(shard *CacheShard) {
	shard.stats.mu.Lock()
	shard.stats.Evictions++
	shard.stats.mu.Unlock()

	ic.stats.mu.Lock()
	ic.stats.Evictions++
	ic.stats.mu.Unlock()
}

// recordExpiration records a cache expiration
func (ic *IntelligentCache) recordExpiration(shard *CacheShard) {
	shard.stats.mu.Lock()
	shard.stats.Expirations++
	shard.stats.mu.Unlock()

	ic.stats.mu.Lock()
	ic.stats.Expirations++
	ic.stats.mu.Unlock()
}

// recordAccessTime records access time for analytics
func (ic *IntelligentCache) recordAccessTime(duration time.Duration) {
	ic.stats.mu.Lock()
	ic.stats.AverageAccessTime = (ic.stats.AverageAccessTime + duration) / 2
	ic.stats.mu.Unlock()
}

// resetStats resets cache statistics
func (ic *IntelligentCache) resetStats() {
	ic.stats.mu.Lock()
	ic.stats.Hits = 0
	ic.stats.Misses = 0
	ic.stats.Evictions = 0
	ic.stats.Expirations = 0
	ic.stats.TotalSize = 0
	ic.stats.EntryCount = 0
	ic.stats.HitRate = 0
	ic.stats.MissRate = 0
	ic.stats.AverageAccessTime = 0
	ic.stats.LastReset = time.Now()
	ic.stats.mu.Unlock()
}

// cleanupWorker performs periodic cleanup of expired entries
func (ic *IntelligentCache) cleanupWorker() {
	ticker := time.NewTicker(ic.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ic.ctx.Done():
			return
		case <-ticker.C:
			ic.cleanupExpiredEntries()
		}
	}
}

// cleanupExpiredEntries removes expired entries from all shards
func (ic *IntelligentCache) cleanupExpiredEntries() {
	for _, shard := range ic.shards {
		shard.mu.Lock()
		for key, entry := range shard.entries {
			if entry.isExpired() {
				shard.stats.TotalSize -= entry.Size
				shard.stats.EntryCount--
				delete(shard.entries, key)
				ic.recordExpiration(shard)
			}
		}
		shard.mu.Unlock()
	}
}

// analyticsWorker collects analytics data
func (ic *IntelligentCache) analyticsWorker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ic.ctx.Done():
			return
		case <-ticker.C:
			// Analytics collection logic would go here
			// For now, we'll just log basic stats
			stats := ic.GetStats()
			ic.logger.Debug("Cache analytics",
				zap.Float64("hit_rate", stats.HitRate),
				zap.Int64("total_entries", stats.EntryCount),
				zap.Int64("total_size", stats.TotalSize),
			)
		}
	}
}

// getPopularKeys returns the most popular keys
func (ic *IntelligentCache) getPopularKeys(limit int) []string {
	// Implementation would collect and sort by access count
	return []string{}
}

// getHotKeys returns the hottest keys (recently accessed)
func (ic *IntelligentCache) getHotKeys(limit int) []string {
	// Implementation would collect and sort by last access time
	return []string{}
}

// getColdKeys returns the coldest keys (not recently accessed)
func (ic *IntelligentCache) getColdKeys(limit int) []string {
	// Implementation would collect and sort by last access time (ascending)
	return []string{}
}

// getSizeRange returns the size range for analytics
func (ic *IntelligentCache) getSizeRange(size int64) string {
	switch {
	case size < 1024:
		return "0-1KB"
	case size < 10240:
		return "1-10KB"
	case size < 102400:
		return "10-100KB"
	case size < 1048576:
		return "100KB-1MB"
	default:
		return "1MB+"
	}
}

// defaultOptions returns default cache options
func (ic *IntelligentCache) defaultOptions() *CacheOptions {
	return &CacheOptions{
		TTL:      0,
		Priority: 5,
		Tags:     []string{},
		Metadata: make(map[string]interface{}),
	}
}

// CacheOptions represents cache operation options
type CacheOptions struct {
	TTL      time.Duration
	Priority int
	Tags     []string
	Metadata map[string]interface{}
}

// CacheOption represents a cache option function
type CacheOption func(*CacheOptions)

// WithTTL sets the TTL for a cache entry
func WithTTL(ttl time.Duration) CacheOption {
	return func(opts *CacheOptions) {
		opts.TTL = ttl
	}
}

// WithPriority sets the priority for a cache entry
func WithPriority(priority int) CacheOption {
	return func(opts *CacheOptions) {
		opts.Priority = priority
	}
}

// WithTags sets tags for a cache entry
func WithTags(tags ...string) CacheOption {
	return func(opts *CacheOptions) {
		opts.Tags = append(opts.Tags, tags...)
	}
}

// WithMetadata sets metadata for a cache entry
func WithMetadata(metadata map[string]interface{}) CacheOption {
	return func(opts *CacheOptions) {
		for k, v := range metadata {
			opts.Metadata[k] = v
		}
	}
}

// isExpired checks if a cache entry is expired
func (ce *CacheEntry) isExpired() bool {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	if ce.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*ce.ExpiresAt)
}

// updateAccess updates access statistics for a cache entry
func (ce *CacheEntry) updateAccess() {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	ce.AccessCount++
	ce.LastAccess = time.Now()
}
