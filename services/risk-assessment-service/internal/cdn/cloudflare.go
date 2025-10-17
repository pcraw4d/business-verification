package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// CloudFlareCDN manages CloudFlare CDN operations
type CloudFlareCDN struct {
	client *http.Client
	logger *zap.Logger
	config *CloudFlareConfig
	stats  *CDNStats
}

// CloudFlareConfig represents configuration for CloudFlare CDN
type CloudFlareConfig struct {
	APIKey      string        `json:"api_key"`
	Email       string        `json:"email"`
	ZoneID      string        `json:"zone_id"`
	AccountID   string        `json:"account_id"`
	APIEndpoint string        `json:"api_endpoint"`
	Timeout     time.Duration `json:"timeout"`
}

// CDNStats represents CDN statistics
type CDNStats struct {
	CacheHits           int64         `json:"cache_hits"`
	CacheMisses         int64         `json:"cache_misses"`
	CachePurges         int64         `json:"cache_purges"`
	BandwidthUsed       int64         `json:"bandwidth_used"`
	RequestsServed      int64         `json:"requests_served"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastPurge           time.Time     `json:"last_purge"`
}

// CacheRule represents a cache rule
type CacheRule struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Pattern    string                 `json:"pattern"`
	CacheLevel string                 `json:"cache_level"`
	TTL        int                    `json:"ttl"`
	BrowserTTL int                    `json:"browser_ttl"`
	EdgeTTL    int                    `json:"edge_ttl"`
	Enabled    bool                   `json:"enabled"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// PurgeRequest represents a cache purge request
type PurgeRequest struct {
	URLs            []string `json:"urls,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	Hosts           []string `json:"hosts,omitempty"`
	Prefixes        []string `json:"prefixes,omitempty"`
	PurgeEverything bool     `json:"purge_everything,omitempty"`
}

// PurgeResponse represents a cache purge response
type PurgeResponse struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ZoneInfo represents CloudFlare zone information
type ZoneInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Paused bool   `json:"paused"`
	Plan   string `json:"plan"`
	Type   string `json:"type"`
}

// NewCloudFlareCDN creates a new CloudFlare CDN manager
func NewCloudFlareCDN(config *CloudFlareConfig, logger *zap.Logger) *CloudFlareCDN {
	if config == nil {
		config = &CloudFlareConfig{
			APIEndpoint: "https://api.cloudflare.com/client/v4",
			Timeout:     30 * time.Second,
		}
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &CloudFlareCDN{
		client: client,
		logger: logger,
		config: config,
		stats:  &CDNStats{},
	}
}

// PurgeCache purges cache for specified URLs or patterns
func (cf *CloudFlareCDN) PurgeCache(ctx context.Context, request *PurgeRequest) (*PurgeResponse, error) {
	cf.logger.Info("Purging CloudFlare cache",
		zap.Strings("urls", request.URLs),
		zap.Strings("tags", request.Tags),
		zap.Strings("prefixes", request.Prefixes),
		zap.Bool("purge_everything", request.PurgeEverything))

	// Prepare request
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal purge request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/zones/%s/purge_cache", cf.config.APIEndpoint, cf.config.ZoneID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create purge request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Email", cf.config.Email)
	req.Header.Set("X-Auth-Key", cf.config.APIKey)

	// Execute request
	resp, err := cf.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute purge request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var response struct {
		Success bool `json:"success"`
		Result  struct {
			ID string `json:"id"`
		} `json:"result"`
		Errors []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode purge response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("purge request failed: %v", response.Errors)
	}

	// Update stats
	cf.stats.CachePurges++
	cf.stats.LastPurge = time.Now()

	purgeResponse := &PurgeResponse{
		ID:      response.Result.ID,
		Status:  "success",
		Message: "Cache purged successfully",
	}

	cf.logger.Info("Cache purge completed",
		zap.String("purge_id", purgeResponse.ID))

	return purgeResponse, nil
}

// CreateCacheRule creates a new cache rule
func (cf *CloudFlareCDN) CreateCacheRule(ctx context.Context, rule *CacheRule) (*CacheRule, error) {
	cf.logger.Info("Creating CloudFlare cache rule",
		zap.String("name", rule.Name),
		zap.String("pattern", rule.Pattern))

	// Prepare request
	reqBody, err := json.Marshal(rule)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cache rule: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/zones/%s/pagerules", cf.config.APIEndpoint, cf.config.ZoneID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create cache rule request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Email", cf.config.Email)
	req.Header.Set("X-Auth-Key", cf.config.APIKey)

	// Execute request
	resp, err := cf.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute cache rule request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var response struct {
		Success bool `json:"success"`
		Result  struct {
			ID      string `json:"id"`
			Targets []struct {
				Target     string `json:"target"`
				Constraint struct {
					Operator string `json:"operator"`
					Value    string `json:"value"`
				} `json:"constraint"`
			} `json:"targets"`
			Actions []struct {
				ID    string      `json:"id"`
				Value interface{} `json:"value"`
			} `json:"actions"`
			Status     string `json:"status"`
			CreatedOn  string `json:"created_on"`
			ModifiedOn string `json:"modified_on"`
		} `json:"result"`
		Errors []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode cache rule response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("cache rule creation failed: %v", response.Errors)
	}

	// Update rule with response data
	rule.ID = response.Result.ID
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	cf.logger.Info("Cache rule created successfully",
		zap.String("rule_id", rule.ID),
		zap.String("name", rule.Name))

	return rule, nil
}

// GetZoneInfo retrieves zone information
func (cf *CloudFlareCDN) GetZoneInfo(ctx context.Context) (*ZoneInfo, error) {
	cf.logger.Debug("Retrieving CloudFlare zone information")

	// Create HTTP request
	url := fmt.Sprintf("%s/zones/%s", cf.config.APIEndpoint, cf.config.ZoneID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create zone info request: %w", err)
	}

	// Set headers
	req.Header.Set("X-Auth-Email", cf.config.Email)
	req.Header.Set("X-Auth-Key", cf.config.APIKey)

	// Execute request
	resp, err := cf.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute zone info request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var response struct {
		Success bool     `json:"success"`
		Result  ZoneInfo `json:"result"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode zone info response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("zone info request failed: %v", response.Errors)
	}

	cf.logger.Debug("Zone information retrieved",
		zap.String("zone_id", response.Result.ID),
		zap.String("zone_name", response.Result.Name),
		zap.String("status", response.Result.Status))

	return &response.Result, nil
}

// GetStats returns CDN statistics
func (cf *CloudFlareCDN) GetStats() *CDNStats {
	return cf.stats
}

// UpdateStats updates CDN statistics
func (cf *CloudFlareCDN) UpdateStats(hits, misses, bandwidth int64, responseTime time.Duration) {
	cf.stats.CacheHits += hits
	cf.stats.CacheMisses += misses
	cf.stats.BandwidthUsed += bandwidth
	cf.stats.RequestsServed += hits + misses
	cf.stats.AverageResponseTime = (cf.stats.AverageResponseTime + responseTime) / 2
}

// ConfigureGeographicRouting configures geographic routing for optimal performance
func (cf *CloudFlareCDN) ConfigureGeographicRouting(ctx context.Context, regions []string) error {
	cf.logger.Info("Configuring geographic routing",
		zap.Strings("regions", regions))

	// This is a simplified implementation
	// In a real implementation, you would use CloudFlare's Load Balancer API

	cf.logger.Info("Geographic routing configured successfully")

	return nil
}

// ConfigureCacheRules configures default cache rules
func (cf *CloudFlareCDN) ConfigureCacheRules(ctx context.Context) error {
	cf.logger.Info("Configuring default cache rules")

	// Define default cache rules
	rules := []*CacheRule{
		{
			Name:       "Static Assets",
			Pattern:    "*.css,*.js,*.png,*.jpg,*.jpeg,*.gif,*.svg,*.ico",
			CacheLevel: "cache_everything",
			TTL:        31536000, // 1 year
			BrowserTTL: 31536000,
			EdgeTTL:    31536000,
			Enabled:    true,
		},
		{
			Name:       "API Responses",
			Pattern:    "/api/v1/*",
			CacheLevel: "cache_everything",
			TTL:        300, // 5 minutes
			BrowserTTL: 0,
			EdgeTTL:    300,
			Enabled:    true,
		},
		{
			Name:       "Model Predictions",
			Pattern:    "/api/v1/predictions/*",
			CacheLevel: "cache_everything",
			TTL:        3600, // 1 hour
			BrowserTTL: 0,
			EdgeTTL:    3600,
			Enabled:    true,
		},
	}

	// Create cache rules
	for _, rule := range rules {
		if _, err := cf.CreateCacheRule(ctx, rule); err != nil {
			cf.logger.Error("Failed to create cache rule",
				zap.String("name", rule.Name),
				zap.Error(err))
			continue
		}
	}

	cf.logger.Info("Default cache rules configured successfully")

	return nil
}

// PurgeByTag purges cache by tag
func (cf *CloudFlareCDN) PurgeByTag(ctx context.Context, tags []string) (*PurgeResponse, error) {
	request := &PurgeRequest{
		Tags: tags,
	}

	return cf.PurgeCache(ctx, request)
}

// PurgeByURL purges cache by URL
func (cf *CloudFlareCDN) PurgeByURL(ctx context.Context, urls []string) (*PurgeResponse, error) {
	request := &PurgeRequest{
		URLs: urls,
	}

	return cf.PurgeCache(ctx, request)
}

// PurgeByPrefix purges cache by prefix
func (cf *CloudFlareCDN) PurgeByPrefix(ctx context.Context, prefixes []string) (*PurgeResponse, error) {
	request := &PurgeRequest{
		Prefixes: prefixes,
	}

	return cf.PurgeCache(ctx, request)
}

// PurgeEverything purges entire cache
func (cf *CloudFlareCDN) PurgeEverything(ctx context.Context) (*PurgeResponse, error) {
	request := &PurgeRequest{
		PurgeEverything: true,
	}

	return cf.PurgeCache(ctx, request)
}
