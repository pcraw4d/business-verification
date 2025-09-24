package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DiskCache provides a disk-based cache with file persistence
type DiskCache struct {
	// Configuration
	config DiskCacheConfig

	// File system operations
	basePath string

	// Thread safety
	mu sync.RWMutex

	// Statistics
	stats     *DiskCacheStats
	statsLock sync.RWMutex

	// Logging
	logger *zap.Logger

	// Control
	stopChannel chan struct{}
}

// DiskCacheConfig holds configuration for the disk cache
type DiskCacheConfig struct {
	Path string        `json:"path"` // Cache directory
	Size int64         `json:"size"` // Max size in bytes
	TTL  time.Duration `json:"ttl"`  // Time to live
}

// DiskCacheStats holds disk cache statistics
type DiskCacheStats struct {
	Hits      int64   `json:"hits"`
	Misses    int64   `json:"misses"`
	Evictions int64   `json:"evictions"`
	Size      int64   `json:"size"`
	MaxSize   int64   `json:"max_size"`
	HitRate   float64 `json:"hit_rate"`
	FileCount int64   `json:"file_count"`
	Errors    int64   `json:"errors"`
}

// diskCacheItem represents an item stored on disk
type diskCacheItem struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	CreatedAt   time.Time   `json:"created_at"`
	ExpiresAt   time.Time   `json:"expires_at"`
	AccessedAt  time.Time   `json:"accessed_at"`
	AccessCount int64       `json:"access_count"`
	Size        int64       `json:"size"`
	Compressed  bool        `json:"compressed"`
	Encrypted   bool        `json:"encrypted"`
}

// CacheItem represents a cache item returned by the disk cache
type CacheItem struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	CreatedAt   time.Time   `json:"created_at"`
	ExpiresAt   time.Time   `json:"expires_at"`
	AccessedAt  time.Time   `json:"accessed_at"`
	AccessCount int64       `json:"access_count"`
	Size        int64       `json:"size"`
	Compressed  bool        `json:"compressed"`
	Encrypted   bool        `json:"encrypted"`
}

// NewDiskCache creates a new disk cache
func NewDiskCache(config DiskCacheConfig, logger *zap.Logger) (*DiskCache, error) {
	if config.Path == "" {
		config.Path = "./cache/disk"
	}
	if config.Size <= 0 {
		config.Size = 100 * 1024 * 1024 // 100MB default
	}
	if config.TTL <= 0 {
		config.TTL = 24 * time.Hour // 24 hours default
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(config.Path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	dc := &DiskCache{
		config:      config,
		basePath:    config.Path,
		stats:       &DiskCacheStats{MaxSize: config.Size},
		logger:      logger,
		stopChannel: make(chan struct{}),
	}

	// Start cleanup goroutine
	go dc.startCleanup()

	return dc, nil
}

// Get retrieves a value from the disk cache
func (dc *DiskCache) Get(ctx context.Context, key string) (*CacheItem, bool) {
	// Generate file path
	filePath := dc.getFilePath(key)

	dc.mu.RLock()
	defer dc.mu.RUnlock()

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		dc.incrementMiss()
		return nil, false
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		dc.logger.Error("Failed to read cache file",
			zap.String("key", key),
			zap.String("path", filePath),
			zap.Error(err))
		dc.incrementError()
		dc.incrementMiss()
		return nil, false
	}

	// Unmarshal item
	var item diskCacheItem
	if err := json.Unmarshal(data, &item); err != nil {
		dc.logger.Error("Failed to unmarshal cache item",
			zap.String("key", key),
			zap.Error(err))
		dc.incrementError()
		dc.incrementMiss()
		return nil, false
	}

	// Check if item has expired
	if time.Now().After(item.ExpiresAt) {
		// Remove expired file
		go dc.removeFile(filePath)
		dc.incrementMiss()
		return nil, false
	}

	// Update access statistics
	item.AccessedAt = time.Now()
	item.AccessCount++

	// Update file with new access info
	go dc.updateAccessInfo(filePath, &item)

	dc.incrementHit()

	return &CacheItem{
		Key:         item.Key,
		Value:       item.Value,
		CreatedAt:   item.CreatedAt,
		ExpiresAt:   item.ExpiresAt,
		AccessedAt:  item.AccessedAt,
		AccessCount: item.AccessCount,
		Size:        item.Size,
		Compressed:  item.Compressed,
		Encrypted:   item.Encrypted,
	}, true
}

// Set stores a value in the disk cache
func (dc *DiskCache) Set(ctx context.Context, key string, value interface{}, expiresAt time.Time) error {
	// Check cache size and evict if necessary
	if err := dc.ensureSpace(); err != nil {
		return fmt.Errorf("failed to ensure space: %w", err)
	}

	// Create cache item
	item := &diskCacheItem{
		Key:         key,
		Value:       value,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		AccessedAt:  time.Now(),
		AccessCount: 1,
		Size:        dc.calculateSize(value),
		Compressed:  false, // TODO: Implement compression
		Encrypted:   false, // TODO: Implement encryption
	}

	// Marshal to JSON
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal cache item: %w", err)
	}

	// Generate file path
	filePath := dc.getFilePath(key)

	// Write to file
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		dc.incrementError()
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	// Update statistics
	dc.updateStats()

	return nil
}

// Delete removes a value from the disk cache
func (dc *DiskCache) Delete(ctx context.Context, key string) error {
	filePath := dc.getFilePath(key)

	dc.mu.Lock()
	defer dc.mu.Unlock()

	if err := os.Remove(filePath); err != nil {
		if !os.IsNotExist(err) {
			dc.incrementError()
			return fmt.Errorf("failed to delete cache file: %w", err)
		}
	}

	// Update statistics
	dc.updateStats()

	return nil
}

// Clear removes all items from the disk cache
func (dc *DiskCache) Clear(ctx context.Context) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Remove all files in cache directory
	err := filepath.WalkDir(dc.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && path != dc.basePath {
			if err := os.Remove(path); err != nil {
				dc.logger.Warn("Failed to remove cache file",
					zap.String("path", path),
					zap.Error(err))
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to clear cache directory: %w", err)
	}

	// Reset statistics
	dc.resetStats()

	return nil
}

// GetStats returns disk cache statistics
func (dc *DiskCache) GetStats() *DiskCacheStats {
	dc.statsLock.RLock()
	defer dc.statsLock.RUnlock()

	stats := *dc.stats
	return &stats
}

// Close closes the disk cache
func (dc *DiskCache) Close() error {
	close(dc.stopChannel)
	return nil
}

// Helper methods

func (dc *DiskCache) getFilePath(key string) string {
	// Use hash of key to avoid filesystem issues with special characters
	hash := fmt.Sprintf("%x", key)
	return filepath.Join(dc.basePath, hash+".cache")
}

func (dc *DiskCache) calculateSize(value interface{}) int64 {
	data, err := json.Marshal(value)
	if err != nil {
		return 0
	}
	return int64(len(data))
}

func (dc *DiskCache) ensureSpace() error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Check current size
	currentSize := dc.getCurrentSize()
	if currentSize < dc.config.Size {
		return nil // Enough space
	}

	// Need to evict some files
	return dc.evictOldestFiles(currentSize - dc.config.Size + 1024*1024) // Leave 1MB buffer
}

func (dc *DiskCache) getCurrentSize() int64 {
	var totalSize int64

	err := filepath.WalkDir(dc.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			if info, err := d.Info(); err == nil {
				totalSize += info.Size()
			}
		}

		return nil
	})

	if err != nil {
		dc.logger.Error("Failed to calculate cache size", zap.Error(err))
		return 0
	}

	return totalSize
}

func (dc *DiskCache) evictOldestFiles(requiredSpace int64) error {
	// Get all files with their access times
	type fileInfo struct {
		path       string
		accessTime time.Time
		size       int64
	}

	var files []fileInfo

	err := filepath.WalkDir(dc.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			if info, err := d.Info(); err == nil {
				files = append(files, fileInfo{
					path:       path,
					accessTime: info.ModTime(),
					size:       info.Size(),
				})
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan cache files: %w", err)
	}

	// Sort by access time (oldest first)
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].accessTime.After(files[j].accessTime) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Remove oldest files until we have enough space
	var freedSpace int64
	for _, file := range files {
		if freedSpace >= requiredSpace {
			break
		}

		if err := os.Remove(file.path); err != nil {
			dc.logger.Warn("Failed to evict cache file",
				zap.String("path", file.path),
				zap.Error(err))
			continue
		}

		freedSpace += file.size
		dc.incrementEviction()
	}

	return nil
}

func (dc *DiskCache) removeFile(filePath string) {
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		dc.logger.Warn("Failed to remove expired cache file",
			zap.String("path", filePath),
			zap.Error(err))
	}
}

func (dc *DiskCache) updateAccessInfo(filePath string, item *diskCacheItem) {
	data, err := json.Marshal(item)
	if err != nil {
		dc.logger.Error("Failed to marshal updated cache item",
			zap.String("key", item.Key),
			zap.Error(err))
		return
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		dc.logger.Error("Failed to update cache file access info",
			zap.String("path", filePath),
			zap.Error(err))
	}
}

func (dc *DiskCache) startCleanup() {
	ticker := time.NewTicker(1 * time.Hour) // Cleanup every hour
	defer ticker.Stop()

	for {
		select {
		case <-dc.stopChannel:
			return
		case <-ticker.C:
			dc.cleanupExpired()
		}
	}
}

func (dc *DiskCache) cleanupExpired() {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	now := time.Now()

	err := filepath.WalkDir(dc.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			// Read file and check expiration
			data, err := os.ReadFile(path)
			if err != nil {
				return nil // Skip files we can't read
			}

			var item diskCacheItem
			if err := json.Unmarshal(data, &item); err != nil {
				return nil // Skip corrupted files
			}

			if now.After(item.ExpiresAt) {
				if err := os.Remove(path); err != nil {
					dc.logger.Warn("Failed to remove expired cache file",
						zap.String("path", path),
						zap.Error(err))
				}
			}
		}

		return nil
	})

	if err != nil {
		dc.logger.Error("Failed to cleanup expired cache files", zap.Error(err))
	}

	// Update statistics
	dc.updateStats()
}

func (dc *DiskCache) incrementHit() {
	dc.statsLock.Lock()
	defer dc.statsLock.Unlock()
	dc.stats.Hits++
	dc.updateHitRate()
}

func (dc *DiskCache) incrementMiss() {
	dc.statsLock.Lock()
	defer dc.statsLock.Unlock()
	dc.stats.Misses++
	dc.updateHitRate()
}

func (dc *DiskCache) incrementEviction() {
	dc.statsLock.Lock()
	defer dc.statsLock.Unlock()
	dc.stats.Evictions++
}

func (dc *DiskCache) incrementError() {
	dc.statsLock.Lock()
	defer dc.statsLock.Unlock()
	dc.stats.Errors++
}

func (dc *DiskCache) updateHitRate() {
	total := dc.stats.Hits + dc.stats.Misses
	if total > 0 {
		dc.stats.HitRate = float64(dc.stats.Hits) / float64(total)
	}
}

func (dc *DiskCache) updateStats() {
	dc.statsLock.Lock()
	defer dc.statsLock.Unlock()

	dc.stats.Size = dc.getCurrentSize()

	// Count files
	fileCount := int64(0)
	filepath.WalkDir(dc.basePath, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			fileCount++
		}
		return nil
	})
	dc.stats.FileCount = fileCount
}

func (dc *DiskCache) resetStats() {
	dc.statsLock.Lock()
	defer dc.statsLock.Unlock()

	dc.stats = &DiskCacheStats{
		MaxSize: dc.config.Size,
	}
}
