package cache

import (
	"context"
	"testing"
)

func TestContentCache_GetSet(t *testing.T) {
	cache := &ContentCache{
		contents: make(map[string]*CachedContent),
	}

	url := "https://example.com"
	content := &CachedContent{
		TextContent: "Test content",
		Title:       "Test Title",
		Keywords:    []string{"test", "example"},
		Success:     true,
	}

	// Test Set
	cache.Set(url, content)

	// Test Get
	retrieved, found := cache.Get(url)
	if !found {
		t.Error("Expected content to be found")
	}
	if retrieved == nil {
		t.Fatal("Expected retrieved content to be non-nil")
	}
	if retrieved.TextContent != content.TextContent {
		t.Errorf("Expected TextContent %q, got %q", content.TextContent, retrieved.TextContent)
	}
	if retrieved.Title != content.Title {
		t.Errorf("Expected Title %q, got %q", content.Title, retrieved.Title)
	}
}

func TestContentCache_GetNotFound(t *testing.T) {
	cache := &ContentCache{
		contents: make(map[string]*CachedContent),
	}

	_, found := cache.Get("https://nonexistent.com")
	if found {
		t.Error("Expected content not to be found")
	}
}

func TestWithContentCache(t *testing.T) {
	ctx := context.Background()
	
	ctx, cache := WithContentCache(ctx)
	
	if cache == nil {
		t.Fatal("Expected cache to be non-nil")
	}

	// Verify cache is in context
	retrievedCache := GetOrCreateCache(ctx)
	if retrievedCache != cache {
		t.Error("Expected retrieved cache to be the same instance")
	}
}

func TestGetFromContext_SetInContext(t *testing.T) {
	ctx, _ := WithContentCache(context.Background())

	url := "https://example.com"
	content := &CachedContent{
		TextContent: "Test content",
		Success:     true,
	}

	// Test SetInContext
	SetInContext(ctx, url, content)

	// Test GetFromContext
	retrieved, found := GetFromContext(ctx, url)
	if !found {
		t.Error("Expected content to be found in context")
	}
	if retrieved == nil {
		t.Fatal("Expected retrieved content to be non-nil")
	}
	if retrieved.TextContent != content.TextContent {
		t.Errorf("Expected TextContent %q, got %q", content.TextContent, retrieved.TextContent)
	}
}

func TestGetFromContext_NotFound(t *testing.T) {
	ctx, _ := WithContentCache(context.Background())

	_, found := GetFromContext(ctx, "https://nonexistent.com")
	if found {
		t.Error("Expected content not to be found")
	}
}

func TestGetFromContext_NoCacheInContext(t *testing.T) {
	ctx := context.Background() // No cache in context

	_, found := GetFromContext(ctx, "https://example.com")
	if found {
		t.Error("Expected content not to be found when cache not in context")
	}
}

