package cache

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestRedisCache_GetSet_NoRedis(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cache := NewRedisCache("", "test", logger)

	key := "test-key"
	data := []byte(`{"test": "data"}`)

	// Test Set
	ctx := context.Background()
	cache.Set(ctx, key, data, 5*time.Minute)

	// Test Get
	retrieved, found := cache.Get(ctx, key)
	if !found {
		t.Error("Expected data to be found in fallback cache")
	}
	if string(retrieved) != string(data) {
		t.Errorf("Expected data %q, got %q", string(data), string(retrieved))
	}
}

func TestRedisCache_GetNotFound(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cache := NewRedisCache("", "test", logger)

	ctx := context.Background()
	_, found := cache.Get(ctx, "nonexistent-key")
	if found {
		t.Error("Expected data not to be found")
	}
}

func TestRedisCache_Delete(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cache := NewRedisCache("", "test", logger)

	key := "test-key"
	data := []byte(`{"test": "data"}`)

	ctx := context.Background()
	cache.Set(ctx, key, data, 5*time.Minute)

	// Verify it's there
	_, found := cache.Get(ctx, key)
	if !found {
		t.Fatal("Expected data to be found before delete")
	}

	// Delete it
	cache.Delete(ctx, key)

	// Verify it's gone
	_, found = cache.Get(ctx, key)
	if found {
		t.Error("Expected data not to be found after delete")
	}
}

func TestRedisCache_Expiration(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cache := NewRedisCache("", "test", logger)

	key := "test-key"
	data := []byte(`{"test": "data"}`)

	ctx := context.Background()
	// Set with very short TTL
	cache.Set(ctx, key, data, 10*time.Millisecond)

	// Should be found immediately
	_, found := cache.Get(ctx, key)
	if !found {
		t.Error("Expected data to be found immediately")
	}

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Should not be found after expiration
	_, found = cache.Get(ctx, key)
	if found {
		t.Error("Expected data not to be found after expiration")
	}
}

func TestRedisCache_Health_NoRedis(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cache := NewRedisCache("", "test", logger)

	ctx := context.Background()
	err := cache.Health(ctx)
	if err == nil {
		t.Error("Expected health check to fail when Redis not enabled")
	}
}

