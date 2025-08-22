package middleware

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// DiskOptimizationConfig configures disk I/O optimization
type DiskOptimizationConfig struct {
	// Cache Configuration
	CacheEnabled        bool          `json:"cache_enabled"`
	CacheDirectory      string        `json:"cache_directory"`
	MaxCacheSize        int64         `json:"max_cache_size"`        // in bytes
	MaxCacheFiles       int           `json:"max_cache_files"`       // maximum number of cached files
	CacheEvictionPolicy string        `json:"cache_eviction_policy"` // "lru", "lfu", "ttl"
	DefaultTTL          time.Duration `json:"default_ttl"`           // time to live for cache entries

	// File I/O Configuration
	BufferSize        int   `json:"buffer_size"`        // buffer size for file operations
	ReadAheadSize     int   `json:"read_ahead_size"`    // read-ahead buffer size
	WriteBufferSize   int   `json:"write_buffer_size"`  // write buffer size
	SyncThreshold     int64 `json:"sync_threshold"`     // sync to disk after this many bytes
	UseDirectIO       bool  `json:"use_direct_io"`      // use direct I/O when possible
	EnableCompression bool  `json:"enable_compression"` // compress cached files

	// Performance Configuration
	MaxConcurrentOps int           `json:"max_concurrent_ops"` // maximum concurrent disk operations
	IOTimeout        time.Duration `json:"io_timeout"`         // timeout for I/O operations
	RetryAttempts    int           `json:"retry_attempts"`     // retry attempts for failed operations
	RetryDelay       time.Duration `json:"retry_delay"`        // delay between retries

	// Monitoring Configuration
	MetricsEnabled  bool          `json:"metrics_enabled"`  // enable metrics collection
	MetricsInterval time.Duration `json:"metrics_interval"` // metrics collection interval
	EnableProfiling bool          `json:"enable_profiling"` // enable disk I/O profiling
}

// DiskOptimizationManager manages disk I/O optimization
type DiskOptimizationManager struct {
	config      *DiskOptimizationConfig
	logger      *zap.Logger
	cache       *DiskCache
	fileManager *FileManager
	monitor     *DiskMonitor
	stats       *DiskStats
	semaphore   chan struct{} // limits concurrent operations
	stopChan    chan struct{}
	mu          sync.RWMutex
}

// DiskCache manages disk-based caching
type DiskCache struct {
	config      *DiskOptimizationConfig
	logger      *zap.Logger
	entries     map[string]*CacheEntry
	accessOrder []*CacheEntry // for LRU
	currentSize int64
	mu          sync.RWMutex
}

// CacheEntry represents a cached file entry
type CacheEntry struct {
	Key          string
	FilePath     string
	Size         int64
	CreatedAt    time.Time
	LastAccessed time.Time
	AccessCount  int64
	TTL          time.Duration
	Compressed   bool
	Checksum     string
}

// FileManager handles optimized file operations
type FileManager struct {
	config  *DiskOptimizationConfig
	logger  *zap.Logger
	buffers sync.Pool
	mu      sync.Mutex
}

// DiskMonitor monitors disk I/O performance
type DiskMonitor struct {
	config *DiskOptimizationConfig
	stats  *DiskStats
	logger *zap.Logger
	mu     sync.RWMutex
}

// DiskStats tracks disk I/O statistics
type DiskStats struct {
	// Read Statistics
	TotalReads      int64         `json:"total_reads"`
	TotalBytesRead  int64         `json:"total_bytes_read"`
	AverageReadTime time.Duration `json:"average_read_time"`
	ReadErrors      int64         `json:"read_errors"`

	// Write Statistics
	TotalWrites       int64         `json:"total_writes"`
	TotalBytesWritten int64         `json:"total_bytes_written"`
	AverageWriteTime  time.Duration `json:"average_write_time"`
	WriteErrors       int64         `json:"write_errors"`

	// Cache Statistics
	CacheHits      int64 `json:"cache_hits"`
	CacheMisses    int64 `json:"cache_misses"`
	CacheEvictions int64 `json:"cache_evictions"`
	CacheSize      int64 `json:"cache_size"`
	CacheFileCount int   `json:"cache_file_count"`

	// Performance Statistics
	AverageIOTime    time.Duration `json:"average_io_time"`
	MaxIOTime        time.Duration `json:"max_io_time"`
	MinIOTime        time.Duration `json:"min_io_time"`
	ActiveOperations int32         `json:"active_operations"`
	QueuedOperations int32         `json:"queued_operations"`

	// System Statistics
	DiskUsage      float64   `json:"disk_usage"`      // percentage
	AvailableSpace int64     `json:"available_space"` // bytes
	IOUtilization  float64   `json:"io_utilization"`  // percentage
	LastUpdated    time.Time `json:"last_updated"`
}

// NewDiskOptimizationManager creates a new disk optimization manager
func NewDiskOptimizationManager(config *DiskOptimizationConfig, logger *zap.Logger) *DiskOptimizationManager {
	if config == nil {
		config = DefaultDiskOptimizationConfig()
	}

	manager := &DiskOptimizationManager{
		config:    config,
		logger:    logger,
		stats:     &DiskStats{LastUpdated: time.Now()},
		semaphore: make(chan struct{}, config.MaxConcurrentOps),
		stopChan:  make(chan struct{}),
	}

	// Initialize components
	manager.cache = NewDiskCache(config, logger)
	manager.fileManager = NewFileManager(config, logger)
	manager.monitor = NewDiskMonitor(config, manager.stats, logger)

	// Start monitoring if enabled
	if config.MetricsEnabled {
		go manager.startMonitoring()
	}

	return manager
}

// DefaultDiskOptimizationConfig returns default disk optimization configuration
func DefaultDiskOptimizationConfig() *DiskOptimizationConfig {
	return &DiskOptimizationConfig{
		// Cache Configuration
		CacheEnabled:        true,
		CacheDirectory:      "/tmp/disk_cache",
		MaxCacheSize:        1 * 1024 * 1024 * 1024, // 1GB
		MaxCacheFiles:       1000,
		CacheEvictionPolicy: "lru",
		DefaultTTL:          24 * time.Hour,

		// File I/O Configuration
		BufferSize:        64 * 1024,   // 64KB
		ReadAheadSize:     128 * 1024,  // 128KB
		WriteBufferSize:   64 * 1024,   // 64KB
		SyncThreshold:     1024 * 1024, // 1MB
		UseDirectIO:       false,
		EnableCompression: false,

		// Performance Configuration
		MaxConcurrentOps: 10,
		IOTimeout:        30 * time.Second,
		RetryAttempts:    3,
		RetryDelay:       100 * time.Millisecond,

		// Monitoring Configuration
		MetricsEnabled:  true,
		MetricsInterval: 30 * time.Second,
		EnableProfiling: false,
	}
}

// ReadFile reads a file with optimization and caching
func (dom *DiskOptimizationManager) ReadFile(ctx context.Context, filePath string) ([]byte, error) {
	start := time.Now()

	// Check cache first
	if dom.config.CacheEnabled {
		if data, found := dom.cache.Get(filePath); found {
			atomic.AddInt64(&dom.stats.CacheHits, 1)
			dom.updateReadStats(len(data), time.Since(start), nil)
			return data, nil
		}
		atomic.AddInt64(&dom.stats.CacheMisses, 1)
	}

	// Acquire semaphore for concurrent operation limiting
	select {
	case dom.semaphore <- struct{}{}:
		defer func() { <-dom.semaphore }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	atomic.AddInt32(&dom.stats.ActiveOperations, 1)
	defer atomic.AddInt32(&dom.stats.ActiveOperations, -1)

	// Read file with optimization
	data, err := dom.fileManager.ReadFile(ctx, filePath)
	duration := time.Since(start)

	if err != nil {
		atomic.AddInt64(&dom.stats.ReadErrors, 1)
		dom.updateReadStats(0, duration, err)
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Cache the data
	if dom.config.CacheEnabled && len(data) > 0 {
		dom.cache.Put(filePath, data)
	}

	dom.updateReadStats(len(data), duration, nil)
	return data, nil
}

// WriteFile writes a file with optimization and caching
func (dom *DiskOptimizationManager) WriteFile(ctx context.Context, filePath string, data []byte) error {
	start := time.Now()

	// Acquire semaphore for concurrent operation limiting
	select {
	case dom.semaphore <- struct{}{}:
		defer func() { <-dom.semaphore }()
	case <-ctx.Done():
		return ctx.Err()
	}

	atomic.AddInt32(&dom.stats.ActiveOperations, 1)
	defer atomic.AddInt32(&dom.stats.ActiveOperations, -1)

	// Write file with optimization
	err := dom.fileManager.WriteFile(ctx, filePath, data)
	duration := time.Since(start)

	if err != nil {
		atomic.AddInt64(&dom.stats.WriteErrors, 1)
		dom.updateWriteStats(0, duration, err)
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	// Update cache
	if dom.config.CacheEnabled {
		dom.cache.Put(filePath, data)
	}

	dom.updateWriteStats(len(data), duration, nil)
	return nil
}

// GetStats returns current disk I/O statistics
func (dom *DiskOptimizationManager) GetStats() *DiskStats {
	dom.mu.RLock()
	defer dom.mu.RUnlock()

	stats := *dom.stats
	stats.LastUpdated = time.Now()

	// Update cache stats
	if dom.config.CacheEnabled {
		stats.CacheSize = dom.cache.GetSize()
		stats.CacheFileCount = dom.cache.GetFileCount()
	}

	return &stats
}

// OptimizeDisk performs disk optimization based on current metrics
func (dom *DiskOptimizationManager) OptimizeDisk() error {
	dom.mu.Lock()
	defer dom.mu.Unlock()

	// Access stats directly to avoid deadlock
	stats := dom.stats

	// Optimize cache based on usage patterns
	if dom.config.CacheEnabled {
		// Clean up expired entries
		dom.cache.CleanupExpired()

		// Trigger eviction if cache is too large
		if stats.CacheSize > dom.config.MaxCacheSize {
			dom.cache.Evict()
		}

		// Adjust buffer sizes based on I/O patterns
		if stats.TotalReads > 0 && stats.AverageReadTime > 100*time.Millisecond {
			// Increase buffer size for slow reads
			dom.config.BufferSize = min(dom.config.BufferSize*2, 1024*1024) // Max 1MB
			dom.logger.Info("increased buffer size due to slow reads",
				zap.Int("new_buffer_size", dom.config.BufferSize))
		}
	}

	return nil
}

// Shutdown gracefully shuts down the disk optimization manager
func (dom *DiskOptimizationManager) Shutdown() error {
	select {
	case <-dom.stopChan:
		// Already shut down
		return nil
	default:
		close(dom.stopChan)
	}

	if dom.config.CacheEnabled {
		return dom.cache.Close()
	}

	return nil
}

// updateReadStats updates read statistics
func (dom *DiskOptimizationManager) updateReadStats(bytes int, duration time.Duration, err error) {
	atomic.AddInt64(&dom.stats.TotalReads, 1)
	if err == nil {
		atomic.AddInt64(&dom.stats.TotalBytesRead, int64(bytes))
	}

	// Update average read time
	dom.mu.Lock()
	if dom.stats.TotalReads > 0 {
		alpha := 0.1
		dom.stats.AverageReadTime = time.Duration(
			float64(dom.stats.AverageReadTime)*(1-alpha) + float64(duration)*alpha)
	} else {
		dom.stats.AverageReadTime = duration
	}
	dom.mu.Unlock()
}

// updateWriteStats updates write statistics
func (dom *DiskOptimizationManager) updateWriteStats(bytes int, duration time.Duration, err error) {
	atomic.AddInt64(&dom.stats.TotalWrites, 1)
	if err == nil {
		atomic.AddInt64(&dom.stats.TotalBytesWritten, int64(bytes))
	}

	// Update average write time
	dom.mu.Lock()
	if dom.stats.TotalWrites > 0 {
		alpha := 0.1
		dom.stats.AverageWriteTime = time.Duration(
			float64(dom.stats.AverageWriteTime)*(1-alpha) + float64(duration)*alpha)
	} else {
		dom.stats.AverageWriteTime = duration
	}
	dom.mu.Unlock()
}

// startMonitoring starts the disk monitoring goroutine
func (dom *DiskOptimizationManager) startMonitoring() {
	ticker := time.NewTicker(dom.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			dom.monitor.CollectMetrics()
		case <-dom.stopChan:
			return
		}
	}
}

// NewDiskCache creates a new disk cache
func NewDiskCache(config *DiskOptimizationConfig, logger *zap.Logger) *DiskCache {
	cache := &DiskCache{
		config:      config,
		logger:      logger,
		entries:     make(map[string]*CacheEntry),
		accessOrder: make([]*CacheEntry, 0),
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(config.CacheDirectory, 0755); err != nil {
		logger.Error("failed to create cache directory", zap.Error(err))
	}

	return cache
}

// Get retrieves data from cache
func (dc *DiskCache) Get(key string) ([]byte, bool) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	entry, exists := dc.entries[key]
	if !exists {
		return nil, false
	}

	// Check TTL
	if entry.TTL > 0 && time.Since(entry.CreatedAt) > entry.TTL {
		dc.removeEntry(key)
		return nil, false
	}

	// Update access information
	entry.LastAccessed = time.Now()
	atomic.AddInt64(&entry.AccessCount, 1)

	// Move to front for LRU
	if dc.config.CacheEvictionPolicy == "lru" {
		dc.moveToFront(entry)
	}

	// Read from cache file
	data, err := os.ReadFile(entry.FilePath)
	if err != nil {
		dc.logger.Error("failed to read from cache", zap.Error(err))
		dc.removeEntry(key)
		return nil, false
	}

	return data, true
}

// Put stores data in cache
func (dc *DiskCache) Put(key string, data []byte) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Generate cache file path
	hash := md5.Sum([]byte(key))
	filename := hex.EncodeToString(hash[:])
	filePath := filepath.Join(dc.config.CacheDirectory, filename)

	// Write to cache file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		dc.logger.Error("failed to write to cache", zap.Error(err))
		return
	}

	// Calculate checksum
	checksum := fmt.Sprintf("%x", md5.Sum(data))

	entry := &CacheEntry{
		Key:          key,
		FilePath:     filePath,
		Size:         int64(len(data)),
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		AccessCount:  0,
		TTL:          dc.config.DefaultTTL,
		Checksum:     checksum,
	}

	// Remove existing entry if present
	if existingEntry, exists := dc.entries[key]; exists {
		dc.removeEntry(key)
		dc.currentSize -= existingEntry.Size
	}

	// Add new entry
	dc.entries[key] = entry
	dc.accessOrder = append(dc.accessOrder, entry)
	dc.currentSize += entry.Size

	// Trigger eviction if needed
	if dc.shouldEvict() {
		dc.evictInternal()
	}
}

// GetSize returns current cache size
func (dc *DiskCache) GetSize() int64 {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return dc.currentSize
}

// GetFileCount returns number of cached files
func (dc *DiskCache) GetFileCount() int {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return len(dc.entries)
}

// CleanupExpired removes expired cache entries
func (dc *DiskCache) CleanupExpired() {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	now := time.Now()
	for key, entry := range dc.entries {
		if entry.TTL > 0 && now.Sub(entry.CreatedAt) > entry.TTL {
			dc.removeEntry(key)
		}
	}
}

// Evict performs cache eviction based on configured policy
func (dc *DiskCache) Evict() {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.evictInternal()
}

// Close closes the cache and cleans up resources
func (dc *DiskCache) Close() error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Clean up all cache files
	for key := range dc.entries {
		dc.removeEntry(key)
	}

	return nil
}

// shouldEvict determines if eviction is needed
func (dc *DiskCache) shouldEvict() bool {
	return dc.currentSize > dc.config.MaxCacheSize || len(dc.entries) > dc.config.MaxCacheFiles
}

// evictInternal performs internal eviction logic
func (dc *DiskCache) evictInternal() {
	if len(dc.entries) == 0 {
		return
	}

	var victimKey string

	switch dc.config.CacheEvictionPolicy {
	case "lru":
		// Least Recently Used
		if len(dc.accessOrder) > 0 {
			victim := dc.accessOrder[0]
			victimKey = victim.Key
		}
	case "lfu":
		// Least Frequently Used
		var minAccess int64 = -1
		for key, entry := range dc.entries {
			if minAccess == -1 || entry.AccessCount < minAccess {
				minAccess = entry.AccessCount
				victimKey = key
			}
		}
	case "ttl":
		// Oldest entry
		var oldest time.Time
		for key, entry := range dc.entries {
			if oldest.IsZero() || entry.CreatedAt.Before(oldest) {
				oldest = entry.CreatedAt
				victimKey = key
			}
		}
	default:
		// Default to LRU
		if len(dc.accessOrder) > 0 {
			victim := dc.accessOrder[0]
			victimKey = victim.Key
		}
	}

	if victimKey != "" {
		dc.removeEntry(victimKey)
	}
}

// removeEntry removes an entry from cache
func (dc *DiskCache) removeEntry(key string) {
	entry, exists := dc.entries[key]
	if !exists {
		return
	}

	// Remove cache file
	if err := os.Remove(entry.FilePath); err != nil && !os.IsNotExist(err) {
		dc.logger.Error("failed to remove cache file", zap.Error(err))
	}

	// Remove from data structures
	delete(dc.entries, key)
	dc.currentSize -= entry.Size

	// Remove from access order
	for i, e := range dc.accessOrder {
		if e.Key == key {
			dc.accessOrder = append(dc.accessOrder[:i], dc.accessOrder[i+1:]...)
			break
		}
	}
}

// moveToFront moves an entry to the front of access order (for LRU)
func (dc *DiskCache) moveToFront(entry *CacheEntry) {
	// Remove from current position
	for i, e := range dc.accessOrder {
		if e.Key == entry.Key {
			dc.accessOrder = append(dc.accessOrder[:i], dc.accessOrder[i+1:]...)
			break
		}
	}

	// Add to front
	dc.accessOrder = append([]*CacheEntry{entry}, dc.accessOrder...)
}

// NewFileManager creates a new file manager
func NewFileManager(config *DiskOptimizationConfig, logger *zap.Logger) *FileManager {
	return &FileManager{
		config: config,
		logger: logger,
		buffers: sync.Pool{
			New: func() interface{} {
				return make([]byte, config.BufferSize)
			},
		},
	}
}

// ReadFile reads a file with optimization
func (fm *FileManager) ReadFile(ctx context.Context, filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get file size for buffer allocation
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	size := fileInfo.Size()
	if size == 0 {
		return []byte{}, nil
	}

	// Use buffered reading for large files
	if size > int64(fm.config.BufferSize) {
		return fm.readBuffered(ctx, file, size)
	}

	// Read small files directly
	data := make([]byte, size)
	_, err = io.ReadFull(file, data)
	return data, err
}

// WriteFile writes a file with optimization
func (fm *FileManager) WriteFile(ctx context.Context, filePath string, data []byte) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Use buffered writing for large files
	if len(data) > fm.config.WriteBufferSize {
		return fm.writeBuffered(ctx, file, data)
	}

	// Write small files directly
	_, err = file.Write(data)
	return err
}

// readBuffered reads a file using buffered I/O
func (fm *FileManager) readBuffered(ctx context.Context, file *os.File, size int64) ([]byte, error) {
	buffer := fm.buffers.Get().([]byte)
	defer fm.buffers.Put(buffer)

	reader := bufio.NewReaderSize(file, len(buffer))
	var result bytes.Buffer
	result.Grow(int(size))

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		n, err := reader.Read(buffer)
		if n > 0 {
			result.Write(buffer[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return result.Bytes(), nil
}

// writeBuffered writes data using buffered I/O
func (fm *FileManager) writeBuffered(ctx context.Context, file *os.File, data []byte) error {
	buffer := fm.buffers.Get().([]byte)
	defer fm.buffers.Put(buffer)

	writer := bufio.NewWriterSize(file, len(buffer))
	defer writer.Flush()

	written := 0
	for written < len(data) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		chunkSize := min(len(buffer), len(data)-written)
		n, err := writer.Write(data[written : written+chunkSize])
		if err != nil {
			return err
		}
		written += n

		// Sync to disk periodically
		if int64(written) >= fm.config.SyncThreshold {
			if err := writer.Flush(); err != nil {
				return err
			}
			if err := file.Sync(); err != nil {
				return err
			}
		}
	}

	return nil
}

// NewDiskMonitor creates a new disk monitor
func NewDiskMonitor(config *DiskOptimizationConfig, stats *DiskStats, logger *zap.Logger) *DiskMonitor {
	return &DiskMonitor{
		config: config,
		stats:  stats,
		logger: logger,
	}
}

// CollectMetrics collects disk I/O metrics
func (dm *DiskMonitor) CollectMetrics() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Update last updated time
	dm.stats.LastUpdated = time.Now()

	// In a real implementation, this would collect actual disk metrics
	// from the system using platform-specific APIs
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
