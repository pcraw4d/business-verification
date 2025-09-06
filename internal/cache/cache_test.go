package cache

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestCacheTypes(t *testing.T) {
	// Test that basic types can be created
	config := CacheConfig{
		Type:             MemoryCache,
		DefaultTTL:       1 * time.Hour,
		MaxSize:          1000,
		KeyPrefix:        "test",
		KeySeparator:     ":",
		KeyHashAlgorithm: "md5",
		CleanupInterval:  5 * time.Minute,
		MetricsInterval:  1 * time.Minute,
	}

	if config.Type != MemoryCache {
		t.Errorf("Expected Type to be MemoryCache")
	}

	if config.MaxSize != 1000 {
		t.Errorf("Expected MaxSize to be 1000, got %d", config.MaxSize)
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
		Type:             MemoryCache,
		DefaultTTL:       5 * time.Minute,
		MaxSize:          1000,
		KeyPrefix:        "test",
		KeySeparator:     ":",
		KeyHashAlgorithm: "md5",
		CleanupInterval:  5 * time.Minute,
		MetricsInterval:  1 * time.Minute,
	}

	manager := NewCacheInvalidationManager(config, logger)
	if manager == nil {
		t.Fatal("Expected invalidation manager to be created")
	}
}

func TestCachePerformanceMonitor(t *testing.T) {
	logger := zap.NewNop()
	config := CacheConfig{
		Type:             MemoryCache,
		DefaultTTL:       1 * time.Hour,
		MaxSize:          1000,
		KeyPrefix:        "test",
		KeySeparator:     ":",
		KeyHashAlgorithm: "md5",
		CleanupInterval:  5 * time.Minute,
		MetricsInterval:  1 * time.Minute,
	}

	monitor := NewCachePerformanceMonitor(config, logger)
	if monitor == nil {
		t.Fatal("Expected performance monitor to be created")
	}
}
