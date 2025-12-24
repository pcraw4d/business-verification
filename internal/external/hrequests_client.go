package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

// FlexibleTime handles flexible time unmarshaling for defensive JSON parsing
// Handles null, invalid formats, and missing time fields gracefully
type FlexibleTime struct {
	time.Time
}

// UnmarshalJSON implements custom JSON unmarshaling for time fields
// Handles null, invalid formats, and missing time fields gracefully
func (ft *FlexibleTime) UnmarshalJSON(b []byte) error {
	// Trim whitespace
	trimmed := strings.TrimSpace(string(b))
	
	// Handle null
	if trimmed == "null" || trimmed == "" {
		ft.Time = time.Time{} // Zero time
		return nil
	}
	
	// Try standard ISO 8601 format
	if err := json.Unmarshal(b, &ft.Time); err == nil {
		return nil
	}
	
	// Try common time formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04:05",
	}
	
	// Remove quotes if present
	timeStr := strings.Trim(trimmed, `"`)
	
	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			ft.Time = t
			return nil
		}
	}
	
	// Fallback to current time if all parsing fails
	// This ensures we don't fail on invalid time formats
	ft.Time = time.Now()
	return nil
}

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

// hrequestsScrapedContentRaw is used for defensive unmarshaling with flexible time handling
type hrequestsScrapedContentRaw struct {
	RawHTML     string                 `json:"raw_html"`
	PlainText   string                 `json:"plain_text"`
	Title       string                 `json:"title"`
	MetaDesc    string                 `json:"meta_description"`
	Headings    []string               `json:"headings"`
	NavMenu     []string               `json:"navigation"`
	AboutText   string                 `json:"about_text"`
	ProductList []string               `json:"products"`
	ContactInfo string                 `json:"contact"`
	MainContent string                 `json:"main_content"`
	WordCount   int                    `json:"word_count"`
	Language    string                 `json:"language"`
	HasLogo     bool                   `json:"has_logo"`
	QualityScore float64               `json:"quality_score"`
	Domain      string                 `json:"domain"`
	ScrapedAt   FlexibleTime           `json:"scraped_at"` // Use FlexibleTime for defensive parsing
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// hrequestsScrapeResponseRaw is used for defensive unmarshaling
type hrequestsScrapeResponseRaw struct {
	Success   bool                        `json:"success"`
	Content   *hrequestsScrapedContentRaw `json:"content,omitempty"`
	Error     string                      `json:"error,omitempty"`
	Method    string                      `json:"method,omitempty"`
	LatencyMS int64                       `json:"latency_ms,omitempty"`
}

// toScrapedContent converts the raw struct to ScrapedContent
func (raw *hrequestsScrapedContentRaw) toScrapedContent() *ScrapedContent {
	if raw == nil {
		return nil
	}
	return &ScrapedContent{
		RawHTML:     raw.RawHTML,
		PlainText:   raw.PlainText,
		Title:       raw.Title,
		MetaDesc:    raw.MetaDesc,
		Headings:    raw.Headings,
		NavMenu:     raw.NavMenu,
		AboutText:   raw.AboutText,
		ProductList: raw.ProductList,
		ContactInfo: raw.ContactInfo,
		MainContent: raw.MainContent,
		WordCount:   raw.WordCount,
		Language:    raw.Language,
		HasLogo:     raw.HasLogo,
		QualityScore: raw.QualityScore,
		Domain:      raw.Domain,
		ScrapedAt:   raw.ScrapedAt.Time, // Convert FlexibleTime to time.Time
		Metadata:    raw.Metadata,
	}
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
			zap.String("response_body", truncateStringLocal(string(body), 500)))
		return nil, fmt.Errorf("hrequests service returned status %d: %s", resp.StatusCode, truncateStringLocal(string(body), 200))
	}

	// Validate response content-type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") && contentType != "" {
		c.logger.Warn("⚠️ [Hrequests] Unexpected content-type",
			zap.String("url", url),
			zap.String("content_type", contentType),
			zap.String("response_body_preview", truncateString(string(body), 200)))
		// Continue anyway - some services don't set content-type correctly
	}

	// Validate JSON structure before unmarshaling
	if !isValidJSON(body) {
		c.logger.Error("❌ [Hrequests] Response is not valid JSON",
			zap.String("url", url),
			zap.String("response_body", truncateStringLocal(string(body), 500)),
			zap.String("content_type", contentType))
		return nil, fmt.Errorf("hrequests service returned invalid JSON response")
	}

	// Parse JSON response with defensive handling using raw structs
	var rawResp hrequestsScrapeResponseRaw
	if err := json.Unmarshal(body, &rawResp); err != nil {
		// Enhanced error logging with response structure analysis
		c.logger.Error("❌ [Hrequests] Failed to unmarshal response",
			zap.String("url", url),
			zap.Error(err),
			zap.String("response_body", truncateStringLocal(string(body), 500)),
			zap.String("content_type", contentType),
			zap.Int("body_length", len(body)),
			zap.Bool("starts_with_brace", len(body) > 0 && body[0] == '{'))
		
		// Try to extract partial information for debugging
		var partialResp map[string]interface{}
		if partialErr := json.Unmarshal(body, &partialResp); partialErr == nil {
			c.logger.Info("📊 [Hrequests] Partial response structure",
				zap.String("url", url),
				zap.Any("keys", getMapKeys(partialResp)))
		}
		
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert raw response to standard response
	hrequestsResp := HrequestsScrapeResponse{
		Success:   rawResp.Success,
		Error:     rawResp.Error,
		Method:    rawResp.Method,
		LatencyMS: rawResp.LatencyMS,
	}
	
	// Convert raw content to ScrapedContent
	if rawResp.Content != nil {
		hrequestsResp.Content = rawResp.Content.toScrapedContent()
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
			zap.String("url", url),
			zap.Bool("success", hrequestsResp.Success),
			zap.String("error", hrequestsResp.Error))
		return nil, fmt.Errorf("hrequests response missing content")
	}

	// Validate and fix scraped_at time if needed (defensive handling)
	if hrequestsResp.Content.ScrapedAt.IsZero() {
		c.logger.Debug("⚠️ [Hrequests] ScrapedAt is zero, setting to current time",
			zap.String("url", url))
		hrequestsResp.Content.ScrapedAt = time.Now()
	}

	// Validate content quality
	if hrequestsResp.Content.QualityScore < 0 || hrequestsResp.Content.QualityScore > 1 {
		c.logger.Warn("⚠️ [Hrequests] Invalid quality score, normalizing",
			zap.String("url", url),
			zap.Float64("original_score", hrequestsResp.Content.QualityScore))
		if hrequestsResp.Content.QualityScore < 0 {
			hrequestsResp.Content.QualityScore = 0
		} else if hrequestsResp.Content.QualityScore > 1 {
			hrequestsResp.Content.QualityScore = 1
		}
	}

	c.logger.Info("✅ [Hrequests] Scraping succeeded",
		zap.String("url", url),
		zap.Float64("quality_score", hrequestsResp.Content.QualityScore),
		zap.Int("word_count", hrequestsResp.Content.WordCount),
		zap.Time("scraped_at", hrequestsResp.Content.ScrapedAt),
		zap.Duration("duration", duration))

	return hrequestsResp.Content, nil
}

// isValidJSON checks if a byte slice contains valid JSON
func isValidJSON(data []byte) bool {
	var js interface{}
	return json.Unmarshal(data, &js) == nil
}

// truncateString truncates a string to maxLen characters, adding ellipsis if truncated
// This is a local helper function for this file
func truncateStringLocal(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// getMapKeys extracts keys from a map for logging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
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



