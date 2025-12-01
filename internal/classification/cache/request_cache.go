package cache

import (
	"context"
	"sync"
)

type contextKey string

const contentCacheKey contextKey = "content_cache"

// ContentCache stores scraped content per URL within a request
type ContentCache struct {
	mu       sync.RWMutex
	contents map[string]*CachedContent
}

// CachedContent represents cached scraped content
type CachedContent struct {
	TextContent string
	Title       string
	Keywords    []string
	Success     bool
	Error       string
}

// GetOrCreateCache gets or creates a content cache in the context
func GetOrCreateCache(ctx context.Context) *ContentCache {
	if cache, ok := ctx.Value(contentCacheKey).(*ContentCache); ok {
		return cache
	}
	
	// Create new cache
	cache := &ContentCache{
		contents: make(map[string]*CachedContent),
	}
	
	// Store in context (note: context is immutable, so this won't work as expected)
	// We need to return the cache and let the caller store it properly
	return cache
}

// WithContentCache adds a content cache to the context
func WithContentCache(ctx context.Context) (context.Context, *ContentCache) {
	cache := &ContentCache{
		contents: make(map[string]*CachedContent),
	}
	return context.WithValue(ctx, contentCacheKey, cache), cache
}

// Get retrieves cached content for a URL
func (cc *ContentCache) Get(url string) (*CachedContent, bool) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	
	content, exists := cc.contents[url]
	return content, exists
}

// Set stores content for a URL
func (cc *ContentCache) Set(url string, content *CachedContent) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	
	cc.contents[url] = content
}

// GetFromContext retrieves cached content from context
func GetFromContext(ctx context.Context, url string) (*CachedContent, bool) {
	if cache, ok := ctx.Value(contentCacheKey).(*ContentCache); ok {
		return cache.Get(url)
	}
	return nil, false
}

// SetInContext stores content in context cache
func SetInContext(ctx context.Context, url string, content *CachedContent) {
	if cache, ok := ctx.Value(contentCacheKey).(*ContentCache); ok {
		cache.Set(url, content)
	}
}

