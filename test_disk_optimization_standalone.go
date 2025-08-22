package main

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
	"time"
)

// Mock logger for standalone testing
type MockLogger struct{}

func (ml *MockLogger) Info(msg string, fields ...interface{}) {
	fmt.Printf("INFO: %s\n", msg)
}

func (ml *MockLogger) Error(msg string, fields ...interface{}) {
	fmt.Printf("ERROR: %s\n", msg)
}

func (ml *MockLogger) Debug(msg string, fields ...interface{}) {
	fmt.Printf("DEBUG: %s\n", msg)
}

// DiskOptimizationConfig configures disk I/O optimization
type DiskOptimizationConfig struct {
	CacheEnabled        bool
	CacheDirectory      string
	MaxCacheSize        int64
	MaxCacheFiles       int
	CacheEvictionPolicy string
	DefaultTTL          time.Duration
	BufferSize          int
	ReadAheadSize       int
	WriteBufferSize     int
	SyncThreshold       int64
	MaxConcurrentOps    int
	IOTimeout           time.Duration
	RetryAttempts       int
	RetryDelay          time.Duration
	MetricsEnabled      bool
	MetricsInterval     time.Duration
}

// DiskOptimizationManager manages disk I/O optimization
type DiskOptimizationManager struct {
	config      *DiskOptimizationConfig
	logger      *MockLogger
	cache       *DiskCache
	fileManager *FileManager
	stats       *DiskStats
	semaphore   chan struct{}
	stopChan    chan struct{}
	mu          sync.RWMutex
}

// DiskCache manages disk-based caching
type DiskCache struct {
	config      *DiskOptimizationConfig
	logger      *MockLogger
	entries     map[string]*CacheEntry
	accessOrder []*CacheEntry
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
	Checksum     string
}

// FileManager handles optimized file operations
type FileManager struct {
	config  *DiskOptimizationConfig
	logger  *MockLogger
	buffers sync.Pool
}

// DiskStats tracks disk I/O statistics
type DiskStats struct {
	TotalReads        int64
	TotalBytesRead    int64
	AverageReadTime   time.Duration
	ReadErrors        int64
	TotalWrites       int64
	TotalBytesWritten int64
	AverageWriteTime  time.Duration
	WriteErrors       int64
	CacheHits         int64
	CacheMisses       int64
	CacheEvictions    int64
	CacheSize         int64
	CacheFileCount    int
	LastUpdated       time.Time
}

// DefaultDiskOptimizationConfig returns default configuration
func DefaultDiskOptimizationConfig() *DiskOptimizationConfig {
	return &DiskOptimizationConfig{
		CacheEnabled:        true,
		CacheDirectory:      "/tmp/disk_cache_standalone",
		MaxCacheSize:        100 * 1024 * 1024, // 100MB
		MaxCacheFiles:       100,
		CacheEvictionPolicy: "lru",
		DefaultTTL:          1 * time.Hour,
		BufferSize:          32 * 1024,  // 32KB
		ReadAheadSize:       64 * 1024,  // 64KB
		WriteBufferSize:     32 * 1024,  // 32KB
		SyncThreshold:       512 * 1024, // 512KB
		MaxConcurrentOps:    5,
		IOTimeout:           10 * time.Second,
		RetryAttempts:       2,
		RetryDelay:          50 * time.Millisecond,
		MetricsEnabled:      false, // Disabled for tests
		MetricsInterval:     10 * time.Second,
	}
}

// NewDiskOptimizationManager creates a new manager
func NewDiskOptimizationManager(config *DiskOptimizationConfig, logger *MockLogger) *DiskOptimizationManager {
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

	manager.cache = NewDiskCache(config, logger)
	manager.fileManager = NewFileManager(config, logger)

	return manager
}

// ReadFile reads a file with optimization and caching
func (dom *DiskOptimizationManager) ReadFile(ctx context.Context, filePath string) ([]byte, error) {
	start := time.Now()

	// Check cache first
	if dom.config.CacheEnabled {
		if data, found := dom.cache.Get(filePath); found {
			dom.stats.CacheHits++
			dom.updateReadStats(len(data), time.Since(start), nil)
			return data, nil
		}
		dom.stats.CacheMisses++
	}

	// Acquire semaphore
	select {
	case dom.semaphore <- struct{}{}:
		defer func() { <-dom.semaphore }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Read file
	data, err := dom.fileManager.ReadFile(ctx, filePath)
	duration := time.Since(start)

	if err != nil {
		dom.stats.ReadErrors++
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

// WriteFile writes a file with optimization
func (dom *DiskOptimizationManager) WriteFile(ctx context.Context, filePath string, data []byte) error {
	start := time.Now()

	// Acquire semaphore
	select {
	case dom.semaphore <- struct{}{}:
		defer func() { <-dom.semaphore }()
	case <-ctx.Done():
		return ctx.Err()
	}

	// Write file
	err := dom.fileManager.WriteFile(ctx, filePath, data)
	duration := time.Since(start)

	if err != nil {
		dom.stats.WriteErrors++
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

// GetStats returns current statistics
func (dom *DiskOptimizationManager) GetStats() *DiskStats {
	dom.mu.RLock()
	defer dom.mu.RUnlock()

	stats := *dom.stats
	stats.LastUpdated = time.Now()

	if dom.config.CacheEnabled {
		stats.CacheSize = dom.cache.GetSize()
		stats.CacheFileCount = dom.cache.GetFileCount()
	}

	return &stats
}

// Shutdown gracefully shuts down the manager
func (dom *DiskOptimizationManager) Shutdown() error {
	close(dom.stopChan)
	if dom.config.CacheEnabled {
		return dom.cache.Close()
	}
	return nil
}

// updateReadStats updates read statistics
func (dom *DiskOptimizationManager) updateReadStats(bytes int, duration time.Duration, err error) {
	dom.stats.TotalReads++
	if err == nil {
		dom.stats.TotalBytesRead += int64(bytes)
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
	dom.stats.TotalWrites++
	if err == nil {
		dom.stats.TotalBytesWritten += int64(bytes)
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

// NewDiskCache creates a new disk cache
func NewDiskCache(config *DiskOptimizationConfig, logger *MockLogger) *DiskCache {
	cache := &DiskCache{
		config:      config,
		logger:      logger,
		entries:     make(map[string]*CacheEntry),
		accessOrder: make([]*CacheEntry, 0),
	}

	// Create cache directory
	if err := os.MkdirAll(config.CacheDirectory, 0755); err != nil {
		logger.Error(fmt.Sprintf("failed to create cache directory: %v", err))
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
	entry.AccessCount++

	// Move to front for LRU
	if dc.config.CacheEvictionPolicy == "lru" {
		dc.moveToFront(entry)
	}

	// Read from cache file
	data, err := os.ReadFile(entry.FilePath)
	if err != nil {
		dc.logger.Error(fmt.Sprintf("failed to read from cache: %v", err))
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
		dc.logger.Error(fmt.Sprintf("failed to write to cache: %v", err))
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

	// Remove existing entry
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

// Close closes the cache
func (dc *DiskCache) Close() error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	for key := range dc.entries {
		dc.removeEntry(key)
	}

	return nil
}

// shouldEvict determines if eviction is needed
func (dc *DiskCache) shouldEvict() bool {
	return dc.currentSize > dc.config.MaxCacheSize || len(dc.entries) > dc.config.MaxCacheFiles
}

// evictInternal performs eviction
func (dc *DiskCache) evictInternal() {
	if len(dc.entries) == 0 {
		return
	}

	var victimKey string

	switch dc.config.CacheEvictionPolicy {
	case "lru":
		if len(dc.accessOrder) > 0 {
			victim := dc.accessOrder[0]
			victimKey = victim.Key
		}
	case "lfu":
		var minAccess int64 = -1
		for key, entry := range dc.entries {
			if minAccess == -1 || entry.AccessCount < minAccess {
				minAccess = entry.AccessCount
				victimKey = key
			}
		}
	default:
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
		dc.logger.Error(fmt.Sprintf("failed to remove cache file: %v", err))
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

// moveToFront moves an entry to front for LRU
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
func NewFileManager(config *DiskOptimizationConfig, logger *MockLogger) *FileManager {
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
	// Create directory
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

// readBuffered reads using buffered I/O
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

// writeBuffered writes using buffered I/O
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
	}

	return nil
}

// min returns minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestDiskOptimization runs comprehensive disk optimization tests
func TestDiskOptimization() {
	fmt.Println("=== Testing Disk Optimization System ===")

	logger := &MockLogger{}
	config := DefaultDiskOptimizationConfig()

	// Use temporary directory for testing
	tempDir := "/tmp/disk_optimization_test_standalone"
	config.CacheDirectory = tempDir

	// Clean up before and after
	os.RemoveAll(tempDir)
	defer os.RemoveAll(tempDir)

	manager := NewDiskOptimizationManager(config, logger)
	defer manager.Shutdown()

	// Test 1: Basic file operations
	fmt.Println("\nTest 1: Basic file operations")
	testData := []byte("test file content for disk optimization")
	testFile := filepath.Join(tempDir, "test_file.txt")

	ctx := context.Background()
	err := manager.WriteFile(ctx, testFile, testData)
	if err != nil {
		fmt.Printf("❌ Write failed: %v\n", err)
		return
	}

	data, err := manager.ReadFile(ctx, testFile)
	if err != nil {
		fmt.Printf("❌ Read failed: %v\n", err)
		return
	}

	if string(data) != string(testData) {
		fmt.Printf("❌ Data mismatch: expected %s, got %s\n", string(testData), string(data))
		return
	}
	fmt.Println("✅ Basic file operations test passed")

	// Test 2: Cache functionality
	fmt.Println("\nTest 2: Cache functionality")
	cachedFile := filepath.Join(tempDir, "cached_file.txt")
	cachedData := []byte("cached file content")

	// Write and read to populate cache
	err = manager.WriteFile(ctx, cachedFile, cachedData)
	if err != nil {
		fmt.Printf("❌ Cache write failed: %v\n", err)
		return
	}

	// First read (cache miss)
	_, err = manager.ReadFile(ctx, cachedFile)
	if err != nil {
		fmt.Printf("❌ First read failed: %v\n", err)
		return
	}

	// Second read (cache hit)
	_, err = manager.ReadFile(ctx, cachedFile)
	if err != nil {
		fmt.Printf("❌ Second read failed: %v\n", err)
		return
	}

	stats := manager.GetStats()
	if stats.CacheHits == 0 {
		fmt.Println("❌ Expected cache hits")
		return
	}
	fmt.Printf("✅ Cache functionality test passed - Cache hits: %d, misses: %d\n",
		stats.CacheHits, stats.CacheMisses)

	// Test 3: Large file handling
	fmt.Println("\nTest 3: Large file handling")
	largeData := make([]byte, 100*1024) // 100KB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	largeFile := filepath.Join(tempDir, "large_file.bin")
	err = manager.WriteFile(ctx, largeFile, largeData)
	if err != nil {
		fmt.Printf("❌ Large file write failed: %v\n", err)
		return
	}

	readData, err := manager.ReadFile(ctx, largeFile)
	if err != nil {
		fmt.Printf("❌ Large file read failed: %v\n", err)
		return
	}

	if len(readData) != len(largeData) {
		fmt.Printf("❌ Large file size mismatch: expected %d, got %d\n",
			len(largeData), len(readData))
		return
	}
	fmt.Println("✅ Large file handling test passed")

	// Test 4: Concurrent operations
	fmt.Println("\nTest 4: Concurrent operations")
	var wg sync.WaitGroup
	numGoroutines := 10
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			concurrentFile := filepath.Join(tempDir, fmt.Sprintf("concurrent_%d.txt", id))
			concurrentData := []byte(fmt.Sprintf("concurrent data %d", id))

			if err := manager.WriteFile(ctx, concurrentFile, concurrentData); err != nil {
				errors <- err
				return
			}

			if _, err := manager.ReadFile(ctx, concurrentFile); err != nil {
				errors <- err
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	if len(errors) > 0 {
		fmt.Printf("❌ Concurrent operations failed: %v\n", <-errors)
		return
	}
	fmt.Println("✅ Concurrent operations test passed")

	// Test 5: Statistics
	fmt.Println("\nTest 5: Statistics")
	finalStats := manager.GetStats()
	if finalStats.TotalReads == 0 && finalStats.TotalWrites == 0 {
		fmt.Println("❌ No operations recorded in statistics")
		return
	}
	fmt.Printf("✅ Statistics test passed - Reads: %d, Writes: %d, Cache size: %d bytes\n",
		finalStats.TotalReads, finalStats.TotalWrites, finalStats.CacheSize)

	fmt.Println("\n🎉 All disk optimization tests passed!")
}

// BenchmarkDiskOptimization runs performance benchmarks
func BenchmarkDiskOptimization() {
	fmt.Println("\n=== Benchmarking Disk Optimization System ===")

	logger := &MockLogger{}
	config := DefaultDiskOptimizationConfig()

	tempDir := "/tmp/disk_optimization_bench_standalone"
	config.CacheDirectory = tempDir

	os.RemoveAll(tempDir)
	defer os.RemoveAll(tempDir)

	manager := NewDiskOptimizationManager(config, logger)
	defer manager.Shutdown()

	ctx := context.Background()

	// Benchmark 1: File read performance
	fmt.Println("\nBenchmarking file reads...")
	testData := make([]byte, 4*1024) // 4KB
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	benchFile := filepath.Join(tempDir, "benchmark_file.bin")
	manager.WriteFile(ctx, benchFile, testData)

	iterations := 1000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, err := manager.ReadFile(ctx, benchFile)
		if err != nil {
			fmt.Printf("❌ Benchmark read failed: %v\n", err)
			return
		}
	}
	duration := time.Since(start)
	fmt.Printf("✅ File reads: %d operations in %v (%.2f ops/sec)\n",
		iterations, duration, float64(iterations)/duration.Seconds())

	// Benchmark 2: File write performance
	fmt.Println("\nBenchmarking file writes...")
	writeData := make([]byte, 1024) // 1KB
	start = time.Now()
	for i := 0; i < iterations; i++ {
		writeFile := filepath.Join(tempDir, fmt.Sprintf("write_bench_%d.bin", i))
		err := manager.WriteFile(ctx, writeFile, writeData)
		if err != nil {
			fmt.Printf("❌ Benchmark write failed: %v\n", err)
			return
		}
	}
	duration = time.Since(start)
	fmt.Printf("✅ File writes: %d operations in %v (%.2f ops/sec)\n",
		iterations, duration, float64(iterations)/duration.Seconds())

	// Benchmark 3: Cache performance
	fmt.Println("\nBenchmarking cache operations...")
	cacheIterations := 10000
	start = time.Now()
	for i := 0; i < cacheIterations; i++ {
		// Read the same file to test cache hits
		_, err := manager.ReadFile(ctx, benchFile)
		if err != nil {
			fmt.Printf("❌ Cache benchmark failed: %v\n", err)
			return
		}
	}
	duration = time.Since(start)
	fmt.Printf("✅ Cache reads: %d operations in %v (%.2f ops/sec)\n",
		cacheIterations, duration, float64(cacheIterations)/duration.Seconds())

	fmt.Println("\n🎉 All benchmarks completed!")
}

func main() {
	TestDiskOptimization()
	BenchmarkDiskOptimization()
}
