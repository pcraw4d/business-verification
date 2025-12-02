package cache

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestWebsiteContentCache_GetSet(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cache := NewWebsiteContentCache(nil, logger, 24*time.Hour)

	// When Redis client is nil, cache is disabled
	if cache.IsEnabled() {
		t.Error("Expected cache to be disabled when Redis client is nil")
	}

	ctx := context.Background()
	url := "https://example.com"
	content := &CachedWebsiteContent{
		TextContent:    "Sample website content",
		Title:          "Example Site",
		Keywords:       []string{"example", "test", "content"},
		StructuredData: map[string]interface{}{"type": "Organization"},
		ScrapedAt:      time.Now(),
		Success:        true,
		StatusCode:     200,
		ContentType:    "text/html",
	}

	// Test Set (should succeed but not actually cache when disabled)
	err := cache.Set(ctx, url, content)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Test Get (should return false when cache is disabled)
	retrieved, found := cache.Get(ctx, url)
	if found {
		t.Error("Expected content not to be found when cache is disabled")
	}
	if retrieved != nil {
		t.Error("Expected nil when cache is disabled")
	}
}

func TestWebsiteContentCache_GetNotFound(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cache := NewWebsiteContentCache(nil, logger, 24*time.Hour)

	ctx := context.Background()
	_, found := cache.Get(ctx, "https://nonexistent.com")
	if found {
		t.Error("Expected content not to be found")
	}
}

func TestWebsiteContentCache_Delete(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cache := NewWebsiteContentCache(nil, logger, 24*time.Hour)

	// When Redis client is nil, cache is disabled
	ctx := context.Background()
	url := "https://example.com"

	// Delete should not error even when cache is disabled
	cache.Delete(ctx, url)

	// Verify content is not found (cache is disabled)
	_, found := cache.Get(ctx, url)
	if found {
		t.Error("Expected content not to be found when cache is disabled")
	}
}

func TestWebsiteContentCache_IsEnabled(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))

	// Test with nil Redis client (should be disabled)
	cache1 := NewWebsiteContentCache(nil, logger, 24*time.Hour)
	if cache1.IsEnabled() {
		t.Error("Expected cache to be disabled when Redis client is nil")
	}

	// Note: Testing with actual Redis client would require a running Redis instance
	// This is better suited for integration tests
}

func TestWebsiteContentCache_Expiration(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cache := NewWebsiteContentCache(nil, logger, 24*time.Hour)

	// When Redis client is nil, cache is disabled
	ctx := context.Background()
	url := "https://example.com"
	content := &CachedWebsiteContent{
		TextContent: "Test content",
		ScrapedAt:   time.Now(),
		Success:     true,
	}

	// Set should succeed but not actually cache when disabled
	err := cache.Set(ctx, url, content)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get should return false when cache is disabled
	_, found := cache.Get(ctx, url)
	if found {
		t.Error("Expected content not to be found when cache is disabled")
	}
}

