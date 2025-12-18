package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
)

// HrequestsClient handles communication with the Python hrequests-scraper service
type HrequestsClient struct {
	serviceURL string
	client     *http.Client
	logger     *zap.Logger
}

// HrequestsScrapeRequest represents the request payload for hrequests service
type HrequestsScrapeRequest struct {
	URL string `json:"url"`
}

// HrequestsScrapeResponse represents the response from hrequests service
type HrequestsScrapeResponse struct {
	Success       bool           `json:"success"`
	Content       *ScrapedContent `json:"content,omitempty"`
	Error         string         `json:"error,omitempty"`
	Method        string         `json:"method,omitempty"`
	LatencyMS     int64          `json:"latency_ms,omitempty"`
}

// NewHrequestsClient creates a new hrequests client
// Reads HREQUESTS_SERVICE_URL from environment (default: http://hrequests-scraper:8080)
func NewHrequestsClient(logger *zap.Logger) *HrequestsClient {
	serviceURL := os.Getenv("HREQUESTS_SERVICE_URL")
	if serviceURL == "" {
		serviceURL = "http://hrequests-scraper:8080"
	}

	client := &http.Client{
		Timeout: 5 * time.Second, // Default timeout for hrequests (fast scraping)
	}

	return &HrequestsClient{
		serviceURL: serviceURL,
		client:     client,
		logger:     logger,
	}
}

// Scrape calls the hrequests service to scrape a website
func (c *HrequestsClient) Scrape(ctx context.Context, url string) (*ScrapedContent, error) {
	if c.serviceURL == "" {
		return nil, fmt.Errorf("hrequests service URL is not configured")
	}

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		c.logger.Warn("⚠️ [Hrequests] Context cancelled before scrape",
			zap.String("url", url),
			zap.Error(ctx.Err()))
		return nil, ctx.Err()
	default:
		// Context is valid, proceed
	}

	startTime := time.Now()

	// Prepare request payload
	reqBody := HrequestsScrapeRequest{
		URL: url,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		c.logger.Error("❌ [Hrequests] Failed to marshal request",
			zap.String("url", url),
			zap.Error(err))
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", c.serviceURL+"/scrape", bytes.NewReader(jsonBody))
	if err != nil {
		c.logger.Error("❌ [Hrequests] Failed to create request",
			zap.String("url", url),
			zap.String("service_url", c.serviceURL),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	c.logger.Debug("🔍 [Hrequests] Sending scrape request",
		zap.String("url", url),
		zap.String("service_url", c.serviceURL))

	// Execute request
	resp, err := c.client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		c.logger.Warn("⚠️ [Hrequests] HTTP request failed",
			zap.String("url", url),
			zap.Error(err),
			zap.Duration("duration", duration))
		
		// Check if error is due to context cancellation
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		
		return nil, fmt.Errorf("hrequests service request failed: %w", err)
	}
	defer resp.Body.Close()

	c.logger.Debug("✅ [Hrequests] HTTP request succeeded",
		zap.String("url", url),
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", duration))

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("❌ [Hrequests] Failed to read response body",
			zap.String("url", url),
			zap.Error(err))
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("⚠️ [Hrequests] Service returned error status",
			zap.String("url", url),
			zap.Int("status_code", resp.StatusCode),
			zap.String("response_body", string(body)))
		return nil, fmt.Errorf("hrequests service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response
	var hrequestsResp HrequestsScrapeResponse
	if err := json.Unmarshal(body, &hrequestsResp); err != nil {
		c.logger.Error("❌ [Hrequests] Failed to unmarshal response",
			zap.String("url", url),
			zap.Error(err),
			zap.String("response_body", string(body)))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check if scraping was successful
	if !hrequestsResp.Success {
		c.logger.Warn("⚠️ [Hrequests] Scraping failed",
			zap.String("url", url),
			zap.String("error", hrequestsResp.Error))
		return nil, fmt.Errorf("hrequests scraping failed: %s", hrequestsResp.Error)
	}

	// Validate content is present
	if hrequestsResp.Content == nil {
		c.logger.Warn("⚠️ [Hrequests] Response missing content",
			zap.String("url", url))
		return nil, fmt.Errorf("hrequests response missing content")
	}

	c.logger.Info("✅ [Hrequests] Scraping succeeded",
		zap.String("url", url),
		zap.Float64("quality_score", hrequestsResp.Content.QualityScore),
		zap.Int("word_count", hrequestsResp.Content.WordCount),
		zap.Duration("duration", duration))

	return hrequestsResp.Content, nil
}

// HealthCheck checks if the hrequests service is available
func (c *HrequestsClient) HealthCheck(ctx context.Context) error {
	if c.serviceURL == "" {
		return fmt.Errorf("hrequests service URL is not configured")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.serviceURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}



