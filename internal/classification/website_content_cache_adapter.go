package classification

import (
	"context"
	"fmt"
)

// WebsiteContentCacheAdapter adapts external cache implementations to WebsiteContentCacher interface
// This allows different cache packages to be used without circular dependencies
type WebsiteContentCacheAdapter struct {
	getFunc    func(ctx context.Context, url string) (*CachedWebsiteContent, bool)
	setFunc    func(ctx context.Context, url string, content *CachedWebsiteContent) error
	enabledFunc func() bool
}

// NewWebsiteContentCacheAdapter creates a new cache adapter from function closures
func NewWebsiteContentCacheAdapter(
	getFunc func(ctx context.Context, url string) (*CachedWebsiteContent, bool),
	setFunc func(ctx context.Context, url string, content *CachedWebsiteContent) error,
	enabledFunc func() bool,
) WebsiteContentCacher {
	return &WebsiteContentCacheAdapter{
		getFunc:     getFunc,
		setFunc:     setFunc,
		enabledFunc: enabledFunc,
	}
}

// Get retrieves cached website content
func (a *WebsiteContentCacheAdapter) Get(ctx context.Context, url string) (*CachedWebsiteContent, bool) {
	if a.getFunc == nil {
		return nil, false
	}
	return a.getFunc(ctx, url)
}

// Set stores website content in cache
func (a *WebsiteContentCacheAdapter) Set(ctx context.Context, url string, content *CachedWebsiteContent) error {
	if a.setFunc == nil {
		return fmt.Errorf("set function not provided")
	}
	return a.setFunc(ctx, url, content)
}

// IsEnabled returns whether the cache is enabled
func (a *WebsiteContentCacheAdapter) IsEnabled() bool {
	if a.enabledFunc == nil {
		return false
	}
	return a.enabledFunc()
}

