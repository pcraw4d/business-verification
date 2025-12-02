package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// CachedWebsiteContent represents cached website content
type CachedWebsiteContent struct {
	TextContent    string                 `json:"text_content"`
	Title          string                 `json:"title"`
	Keywords       []string               `json:"keywords"`
	StructuredData map[string]interface{} `json:"structured_data"`
	ScrapedAt      time.Time              `json:"scraped_at"`
	Success        bool                   `json:"success"`
	StatusCode     int                    `json:"status_code,omitempty"`
	ContentType    string                 `json:"content_type,omitempty"`
}

// WebsiteContentCache provides caching for website content with Redis backend
type WebsiteContentCache struct {
	redisClient *redis.Client
	logger      *zap.Logger
	enabled     bool
	ttl         time.Duration
	prefix      string
}

// NewWebsiteContentCache creates a new website content cache
func NewWebsiteContentCache(redisClient *redis.Client, logger *zap.Logger, ttl time.Duration) *WebsiteContentCache {
	enabled := redisClient != nil
	if enabled {
		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			logger.Warn("Redis client not available for website content cache, caching disabled",
				zap.Error(err))
			enabled = false
		}
	}

	return &WebsiteContentCache{
		redisClient: redisClient,
		logger:      logger,
		enabled:     enabled,
		ttl:         ttl,
		prefix:      "website:content",
	}
}

// Get retrieves cached website content for a URL
func (wcc *WebsiteContentCache) Get(ctx context.Context, url string) (*CachedWebsiteContent, bool) {
	if !wcc.enabled || wcc.redisClient == nil {
		return nil, false
	}

	key := wcc.getKey(url)
	data, err := wcc.redisClient.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false
	}
	if err != nil {
		wcc.logger.Debug("Failed to get website content from cache",
			zap.String("url", url),
			zap.Error(err))
		return nil, false
	}

	var cached CachedWebsiteContent
	if err := json.Unmarshal(data, &cached); err != nil {
		wcc.logger.Warn("Failed to unmarshal cached website content",
			zap.String("url", url),
			zap.Error(err))
		return nil, false
	}

	wcc.logger.Debug("Website content cache hit",
		zap.String("url", url),
		zap.Time("scraped_at", cached.ScrapedAt))
	return &cached, true
}

// Set stores website content in cache
func (wcc *WebsiteContentCache) Set(ctx context.Context, url string, content *CachedWebsiteContent) error {
	if !wcc.enabled || wcc.redisClient == nil {
		return nil
	}

	key := wcc.getKey(url)
	data, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal website content: %w", err)
	}

	if err := wcc.redisClient.Set(ctx, key, data, wcc.ttl).Err(); err != nil {
		wcc.logger.Warn("Failed to set website content in cache",
			zap.String("url", url),
			zap.Error(err))
		return err
	}

	wcc.logger.Debug("Website content cached",
		zap.String("url", url),
		zap.Duration("ttl", wcc.ttl))
	return nil
}

// Delete removes cached website content
func (wcc *WebsiteContentCache) Delete(ctx context.Context, url string) error {
	if !wcc.enabled || wcc.redisClient == nil {
		return nil
	}

	key := wcc.getKey(url)
	if err := wcc.redisClient.Del(ctx, key).Err(); err != nil {
		wcc.logger.Warn("Failed to delete website content from cache",
			zap.String("url", url),
			zap.Error(err))
		return err
	}

	return nil
}

// getKey returns the full cache key for a URL
func (wcc *WebsiteContentCache) getKey(url string) string {
	return fmt.Sprintf("%s:%s", wcc.prefix, url)
}

// IsEnabled returns whether the cache is enabled
func (wcc *WebsiteContentCache) IsEnabled() bool {
	return wcc.enabled
}

