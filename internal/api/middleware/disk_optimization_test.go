package middleware

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestDiskOptimizationManager(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultDiskOptimizationConfig()

	// Use temporary directory for testing
	tempDir, err := os.MkdirTemp("", "disk_optimization_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config.CacheDirectory = tempDir
	config.MetricsEnabled = false // Disable monitoring for tests

	manager := NewDiskOptimizationManager(config, logger)
	defer manager.Shutdown()

	t.Run("default configuration", func(t *testing.T) {
		if manager.config == nil {
			t.Error("expected config to be set")
		}
		if manager.cache == nil {
			t.Error("expected cache to be initialized")
		}
		if manager.fileManager == nil {
			t.Error("expected file manager to be initialized")
		}
		if manager.monitor == nil {
			t.Error("expected monitor to be initialized")
		}
	})

	t.Run("read and write file", func(t *testing.T) {
		testData := []byte("test file content")
		testFile := filepath.Join(tempDir, "test_file.txt")

		// Write file
		ctx := context.Background()
		err := manager.WriteFile(ctx, testFile, testData)
		if err != nil {
			t.Errorf("expected no error writing file, got %v", err)
		}

		// Read file
		data, err := manager.ReadFile(ctx, testFile)
		if err != nil {
			t.Errorf("expected no error reading file, got %v", err)
		}

		if string(data) != string(testData) {
			t.Errorf("expected %s, got %s", string(testData), string(data))
		}
	})

	t.Run("cache functionality", func(t *testing.T) {
		testData := []byte("cached file content")
		testFile := filepath.Join(tempDir, "cached_file.txt")

		// Create file directly without using manager to ensure cache miss
		ctx := context.Background()
		err := os.WriteFile(testFile, testData, 0644)
		if err != nil {
			t.Errorf("expected no error creating file, got %v", err)
		}

		// First read (cache miss)
		data1, err := manager.ReadFile(ctx, testFile)
		if err != nil {
			t.Errorf("expected no error reading file, got %v", err)
		}

		// Second read (cache hit)
		data2, err := manager.ReadFile(ctx, testFile)
		if err != nil {
			t.Errorf("expected no error reading file, got %v", err)
		}

		if string(data1) != string(data2) {
			t.Error("cached data should match original data")
		}

		// Check cache statistics
		stats := manager.GetStats()
		if stats.CacheHits == 0 {
			t.Error("expected at least one cache hit")
		}
		if stats.CacheMisses == 0 {
			t.Error("expected at least one cache miss")
		}
		if stats.TotalReads < 2 {
			t.Error("expected at least 2 reads for cache test")
		}
	})

	t.Run("statistics collection", func(t *testing.T) {
		stats := manager.GetStats()
		if stats == nil {
			t.Error("expected stats to be returned")
		}

		if stats.TotalReads == 0 && stats.TotalWrites == 0 {
			t.Error("expected some read or write operations to be recorded")
		}
	})

	t.Run("optimization", func(t *testing.T) {
		err := manager.OptimizeDisk()
		if err != nil {
			t.Errorf("expected no error during optimization, got %v", err)
		}
	})

	t.Run("shutdown", func(t *testing.T) {
		err := manager.Shutdown()
		if err != nil {
			t.Errorf("expected no error during shutdown, got %v", err)
		}
	})
}

func TestDiskCache(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultDiskOptimizationConfig()

	// Use temporary directory for testing
	tempDir, err := os.MkdirTemp("", "disk_cache_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config.CacheDirectory = tempDir
	cache := NewDiskCache(config, logger)
	defer cache.Close()

	t.Run("basic cache operations", func(t *testing.T) {
		key := "test_key"
		data := []byte("test data")

		// Put data
		cache.Put(key, data)

		// Get data
		retrieved, found := cache.Get(key)
		if !found {
			t.Error("expected data to be found in cache")
		}
		if string(retrieved) != string(data) {
			t.Errorf("expected %s, got %s", string(data), string(retrieved))
		}

		// Check cache size
		if cache.GetSize() == 0 {
			t.Error("expected cache size to be greater than 0")
		}
		if cache.GetFileCount() == 0 {
			t.Error("expected cache file count to be greater than 0")
		}
	})

	t.Run("cache eviction", func(t *testing.T) {
		// Set small cache limits
		config.MaxCacheSize = 100 // 100 bytes
		config.MaxCacheFiles = 2
		smallCache := NewDiskCache(config, logger)
		defer smallCache.Close()

		// Fill cache beyond limits
		for i := 0; i < 5; i++ {
			key := string(rune('a' + i))
			data := make([]byte, 50) // 50 bytes each
			smallCache.Put(key, data)
		}

		// Cache should have evicted some entries
		if smallCache.GetFileCount() > config.MaxCacheFiles {
			t.Errorf("expected cache file count <= %d, got %d",
				config.MaxCacheFiles, smallCache.GetFileCount())
		}
	})

	t.Run("TTL expiration", func(t *testing.T) {
		config.DefaultTTL = 10 * time.Millisecond
		ttlCache := NewDiskCache(config, logger)
		defer ttlCache.Close()

		key := "ttl_test"
		data := []byte("ttl test data")

		// Put data
		ttlCache.Put(key, data)

		// Immediate retrieval should work
		_, found := ttlCache.Get(key)
		if !found {
			t.Error("expected data to be found immediately")
		}

		// Wait for expiration
		time.Sleep(20 * time.Millisecond)

		// Should be expired
		_, found = ttlCache.Get(key)
		if found {
			t.Error("expected data to be expired")
		}
	})

	t.Run("cleanup expired entries", func(t *testing.T) {
		config.DefaultTTL = 10 * time.Millisecond
		cleanupCache := NewDiskCache(config, logger)
		defer cleanupCache.Close()

		// Add expired entries
		for i := 0; i < 3; i++ {
			key := string(rune('x' + i))
			data := []byte("expired data")
			cleanupCache.Put(key, data)
		}

		// Wait for expiration
		time.Sleep(20 * time.Millisecond)

		// Cleanup expired entries
		cleanupCache.CleanupExpired()

		// Cache should be empty
		if cleanupCache.GetFileCount() != 0 {
			t.Errorf("expected 0 files after cleanup, got %d", cleanupCache.GetFileCount())
		}
	})
}

func TestFileManager(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultDiskOptimizationConfig()
	config.BufferSize = 1024
	config.WriteBufferSize = 1024

	fileManager := NewFileManager(config, logger)

	// Use temporary directory for testing
	tempDir, err := os.MkdirTemp("", "file_manager_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("small file operations", func(t *testing.T) {
		testData := []byte("small file content")
		testFile := filepath.Join(tempDir, "small_file.txt")

		ctx := context.Background()

		// Write small file
		err := fileManager.WriteFile(ctx, testFile, testData)
		if err != nil {
			t.Errorf("expected no error writing small file, got %v", err)
		}

		// Read small file
		data, err := fileManager.ReadFile(ctx, testFile)
		if err != nil {
			t.Errorf("expected no error reading small file, got %v", err)
		}

		if string(data) != string(testData) {
			t.Errorf("expected %s, got %s", string(testData), string(data))
		}
	})

	t.Run("large file operations", func(t *testing.T) {
		// Create large test data (larger than buffer)
		testData := make([]byte, 5*1024) // 5KB
		for i := range testData {
			testData[i] = byte(i % 256)
		}

		testFile := filepath.Join(tempDir, "large_file.bin")
		ctx := context.Background()

		// Write large file
		err := fileManager.WriteFile(ctx, testFile, testData)
		if err != nil {
			t.Errorf("expected no error writing large file, got %v", err)
		}

		// Read large file
		data, err := fileManager.ReadFile(ctx, testFile)
		if err != nil {
			t.Errorf("expected no error reading large file, got %v", err)
		}

		if len(data) != len(testData) {
			t.Errorf("expected length %d, got %d", len(testData), len(data))
		}

		// Verify content
		for i, b := range data {
			if b != testData[i] {
				t.Errorf("data mismatch at position %d: expected %d, got %d", i, testData[i], b)
				break
			}
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		testData := make([]byte, 10*1024) // 10KB
		testFile := filepath.Join(tempDir, "cancelled_file.bin")

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Write should fail with context cancellation
		err := fileManager.WriteFile(ctx, testFile, testData)
		if err == nil {
			t.Error("expected error due to context cancellation")
		}
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})

	t.Run("empty file operations", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "empty_file.txt")
		ctx := context.Background()

		// Write empty file
		err := fileManager.WriteFile(ctx, testFile, []byte{})
		if err != nil {
			t.Errorf("expected no error writing empty file, got %v", err)
		}

		// Read empty file
		data, err := fileManager.ReadFile(ctx, testFile)
		if err != nil {
			t.Errorf("expected no error reading empty file, got %v", err)
		}

		if len(data) != 0 {
			t.Errorf("expected empty data, got %d bytes", len(data))
		}
	})
}

func TestDiskMonitor(t *testing.T) {
	config := DefaultDiskOptimizationConfig()
	stats := &DiskStats{}
	logger := zap.NewNop()

	monitor := NewDiskMonitor(config, stats, logger)

	t.Run("metrics collection", func(t *testing.T) {
		// Collect metrics
		monitor.CollectMetrics()

		if stats.LastUpdated.IsZero() {
			t.Error("expected LastUpdated to be set")
		}
	})
}

func TestDiskCacheEvictionPolicies(t *testing.T) {
	logger := zap.NewNop()

	// Use temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cache_eviction_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testEvictionPolicy := func(policy string) {
		t.Run("eviction_policy_"+policy, func(t *testing.T) {
			config := DefaultDiskOptimizationConfig()
			config.CacheDirectory = filepath.Join(tempDir, policy)
			config.CacheEvictionPolicy = policy
			config.MaxCacheSize = 200 // 200 bytes
			config.MaxCacheFiles = 3

			cache := NewDiskCache(config, logger)
			defer cache.Close()

			// Add more entries than cache can hold
			for i := 0; i < 5; i++ {
				key := string(rune('A' + i))
				data := make([]byte, 100) // 100 bytes each
				cache.Put(key, data)

				// Access some entries more frequently for LFU testing
				if policy == "lfu" && i < 2 {
					cache.Get(key)
					cache.Get(key)
				}

				// Add delay for TTL testing
				if policy == "ttl" {
					time.Sleep(1 * time.Millisecond)
				}
			}

			// Cache should respect limits
			if cache.GetFileCount() > config.MaxCacheFiles {
				t.Errorf("cache exceeded max files: %d > %d",
					cache.GetFileCount(), config.MaxCacheFiles)
			}

			if cache.GetSize() > config.MaxCacheSize {
				t.Errorf("cache exceeded max size: %d > %d",
					cache.GetSize(), config.MaxCacheSize)
			}
		})
	}

	testEvictionPolicy("lru")
	testEvictionPolicy("lfu")
	testEvictionPolicy("ttl")
}

func BenchmarkDiskOptimizationManager_ReadFile(b *testing.B) {
	logger := zap.NewNop()
	config := DefaultDiskOptimizationConfig()

	// Use temporary directory for benchmarking
	tempDir, err := os.MkdirTemp("", "disk_optimization_bench")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config.CacheDirectory = tempDir
	config.MetricsEnabled = false

	manager := NewDiskOptimizationManager(config, logger)
	defer manager.Shutdown()

	// Create test file
	testData := make([]byte, 1024) // 1KB
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	testFile := filepath.Join(tempDir, "benchmark_file.bin")
	ctx := context.Background()

	err = manager.WriteFile(ctx, testFile, testData)
	if err != nil {
		b.Fatalf("failed to write test file: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := manager.ReadFile(ctx, testFile)
			if err != nil {
				b.Errorf("read failed: %v", err)
			}
		}
	})
}

func BenchmarkDiskOptimizationManager_WriteFile(b *testing.B) {
	logger := zap.NewNop()
	config := DefaultDiskOptimizationConfig()

	// Use temporary directory for benchmarking
	tempDir, err := os.MkdirTemp("", "disk_optimization_bench_write")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config.CacheDirectory = tempDir
	config.MetricsEnabled = false

	manager := NewDiskOptimizationManager(config, logger)
	defer manager.Shutdown()

	// Create test data
	testData := make([]byte, 1024) // 1KB
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(tempDir, "benchmark_write_file_"+string(rune(i%26+'A'))+".bin")
		err := manager.WriteFile(ctx, testFile, testData)
		if err != nil {
			b.Errorf("write failed: %v", err)
		}
	}
}

func BenchmarkDiskCache_Get(b *testing.B) {
	logger := zap.NewNop()
	config := DefaultDiskOptimizationConfig()

	// Use temporary directory for benchmarking
	tempDir, err := os.MkdirTemp("", "disk_cache_bench")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config.CacheDirectory = tempDir
	cache := NewDiskCache(config, logger)
	defer cache.Close()

	// Pre-populate cache
	testData := []byte("benchmark test data")
	for i := 0; i < 100; i++ {
		key := string(rune('A' + i%26))
		cache.Put(key, testData)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := string(rune('A' + (b.N % 26)))
			_, _ = cache.Get(key)
		}
	})
}

func BenchmarkDiskCache_Put(b *testing.B) {
	logger := zap.NewNop()
	config := DefaultDiskOptimizationConfig()

	// Use temporary directory for benchmarking
	tempDir, err := os.MkdirTemp("", "disk_cache_bench_put")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config.CacheDirectory = tempDir
	cache := NewDiskCache(config, logger)
	defer cache.Close()

	testData := []byte("benchmark test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := string(rune('A' + i%26))
		cache.Put(key, testData)
	}
}
