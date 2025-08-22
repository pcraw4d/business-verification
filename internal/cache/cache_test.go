package cache

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestCacheTypes(t *testing.T) {
	// Test that basic types can be created
	config := CacheConfig{
		EnableMemoryCache: true,
		EnableDiskCache:   false,
		EnableRedisCache:  false,
		MemoryCacheSize:   1000,
		MemoryCacheTTL:    1 * time.Hour,
		MemoryCachePolicy: "LRU",
	}

	if config.EnableMemoryCache != true {
		t.Errorf("Expected EnableMemoryCache to be true")
	}

	if config.MemoryCacheSize != 1000 {
		t.Errorf("Expected MemoryCacheSize to be 1000, got %d", config.MemoryCacheSize)
	}
}

func TestCacheKeyManager(t *testing.T) {
	logger := zap.NewNop()
	config := CacheConfig{
		KeyPrefix:        "test",
		KeySeparator:     ":",
		KeyHashAlgorithm: "MD5",
	}

	keyManager := NewCacheKeyManager(config, logger)
	if keyManager == nil {
		t.Fatal("Expected key manager to be created")
	}

	key := keyManager.GenerateKey("test-key")
	if key == "" {
		t.Error("Expected generated key to be non-empty")
	}
}

func TestCacheInvalidationManager(t *testing.T) {
	logger := zap.NewNop()
	config := CacheConfig{
		EnableAutoInvalidation: true,
		InvalidationInterval:   5 * time.Minute,
	}

	manager := NewCacheInvalidationManager(config, logger)
	if manager == nil {
		t.Fatal("Expected invalidation manager to be created")
	}
}

func TestCachePerformanceMonitor(t *testing.T) {
	logger := zap.NewNop()
	config := CacheConfig{
		EnableMetrics:   true,
		MetricsInterval: 1 * time.Minute,
	}

	monitor := NewCachePerformanceMonitor(config, logger)
	if monitor == nil {
		t.Fatal("Expected performance monitor to be created")
	}
}
