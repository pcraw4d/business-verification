package external

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/html"
)

// ScrapingConfig holds configuration for website scraping
type ScrapingConfig struct {
	Timeout           time.Duration `json:"timeout"`
	MaxRetries        int           `json:"max_retries"`
	RetryDelay        time.Duration `json:"retry_delay"`
	MaxRedirects      int           `json:"max_redirects"`
	UserAgent         string        `json:"user_agent"`
	FollowRedirects   bool          `json:"follow_redirects"`
	VerifySSL         bool          `json:"verify_ssl"`
	MaxResponseSize   int64         `json:"max_response_size"`
	RateLimitDelay    time.Duration `json:"rate_limit_delay"`
	EnableCompression bool          `json:"enable_compression"`
}

// DefaultScrapingConfig returns default configuration for website scraping
func DefaultScrapingConfig() *ScrapingConfig {
	return &ScrapingConfig{
		Timeout:           30 * time.Second,
		MaxRetries:        3,
		RetryDelay:        2 * time.Second,
		MaxRedirects:      5,
		UserAgent:         "KYB-Platform-Bot/1.0 (+https://kyb-platform.com/bot)",
		FollowRedirects:   true,
		VerifySSL:         true,
		MaxResponseSize:   10 * 1024 * 1024, // 10MB
		RateLimitDelay:    1 * time.Second,
		EnableCompression: true,
	}
}

// ScraperStrategy interface for different scraping strategies
type ScraperStrategy interface {
	Scrape(ctx context.Context, url string) (*ScrapedContent, error)
	Name() string
}

// WebsiteScraper handles website scraping with robust error handling and retry logic
type WebsiteScraper struct {
	client    *http.Client
	config    *ScrapingConfig
	logger    *zap.Logger
	strategies []ScraperStrategy // Multi-tier scraping strategies
	playwrightServiceURL string  // URL for Playwright service (optional)
}

// NewWebsiteScraper creates a new website scraper with the given configuration
// It automatically reads PLAYWRIGHT_SERVICE_URL from environment if available
func NewWebsiteScraper(config *ScrapingConfig, logger *zap.Logger) *WebsiteScraper {
	// Check for Playwright service URL in environment
	playwrightServiceURL := os.Getenv("PLAYWRIGHT_SERVICE_URL")
	return NewWebsiteScraperWithStrategies(config, logger, playwrightServiceURL)
}

// NewWebsiteScraperWithStrategies creates a new website scraper with multi-tier strategies
func NewWebsiteScraperWithStrategies(config *ScrapingConfig, logger *zap.Logger, playwrightServiceURL string) *WebsiteScraper {
	if config == nil {
		config = DefaultScrapingConfig()
	}

	// Create transport with custom settings
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !config.VerifySSL,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  !config.EnableCompression,
	}

	// Create HTTP client
	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	// Add redirect handling if enabled
	if config.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= config.MaxRedirects {
				return fmt.Errorf("too many redirects (%d)", len(via))
			}
			return nil
		}
	} else {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// Initialize strategies
	// Note: HTTP client timeout should be set to match or be less than context deadline
	// Context deadline is typically 20s, so client timeout of 30s could cause issues
	// The client timeout is a hard limit, but context cancellation should take precedence
	// For now, we keep 30s as a safety net, but requests use context which will cancel at 20s
	strategies := []ScraperStrategy{
		&SimpleHTTPScraper{client: client, logger: logger},
		&BrowserHeadersScraper{client: client, logger: logger},
	}

	// Add Playwright strategy if URL is provided
	// FIX: Increase client timeout to account for queue wait times (25s) + scrape time (15s) + overhead (10s)
	// The getClientWithContextTimeout will still respect context deadline, but base timeout needs to be higher
	// to handle queue wait scenarios
	if playwrightServiceURL != "" {
		playwrightClient := &http.Client{Timeout: 60 * time.Second}
		strategies = append(strategies, &PlaywrightScraper{
			serviceURL: playwrightServiceURL,
			client:     playwrightClient,
			logger:     logger,
		})
	}

	return &WebsiteScraper{
		client:             client,
		config:             config,
		logger:             logger,
		strategies:         strategies,
		playwrightServiceURL: playwrightServiceURL,
	}
}

// ScrapingResult represents the result of a website scraping operation
type ScrapingResult struct {
	URL           string            `json:"url"`
	StatusCode    int               `json:"status_code"`
	Content       string            `json:"content"`
	ContentType   string            `json:"content_type"`
	ContentLength int64             `json:"content_length"`
	Headers       map[string]string `json:"headers"`
	FinalURL      string            `json:"final_url"`
	ScrapedAt     time.Time         `json:"scraped_at"`
	Duration      time.Duration     `json:"duration"`
	RetryCount    int               `json:"retry_count"`
	Error         string            `json:"error,omitempty"`
	// Enhanced structured content (populated by ScrapeWithStructuredContent)
	StructuredContent *ScrapedContent `json:"structured_content,omitempty"`
}

// ScrapedContent represents structured content extracted from a website
type ScrapedContent struct {
	// Existing fields
	RawHTML   string `json:"raw_html"`
	PlainText string `json:"plain_text"`

	// High-signal structured content
	Title       string   `json:"title"`
	MetaDesc    string   `json:"meta_description"`
	Headings    []string `json:"headings"`        // H1, H2, H3
	NavMenu     []string `json:"navigation"`      // Nav items (business areas)
	AboutText   string   `json:"about_text"`      // About/Company section
	ProductList []string `json:"products"`        // Products/services
	ContactInfo string   `json:"contact"`         // Contact page content

	// Quality metrics
	WordCount   int     `json:"word_count"`
	Language    string  `json:"language"`
	HasLogo     bool    `json:"has_logo"`
	QualityScore float64 `json:"quality_score"`

	// Metadata
	Domain    string    `json:"domain"`
	ScrapedAt time.Time `json:"scraped_at"`
}

// ScrapeWebsite scrapes a website with retry logic and comprehensive error handling
// If strategies are available, it uses the enhanced ScrapeWithStructuredContent method
func (s *WebsiteScraper) ScrapeWebsite(ctx context.Context, targetURL string) (*ScrapingResult, error) {
	startTime := time.Now()

	// Validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme == "" {
		targetURL = "https://" + targetURL
		parsedURL, err = url.Parse(targetURL)
		if err != nil {
			return nil, fmt.Errorf("invalid URL after adding scheme: %w", err)
		}
	}

	s.logger.Info("Starting website scraping",
		zap.String("url", targetURL),
		zap.String("user_agent", s.config.UserAgent))

	// If strategies are available, use enhanced method
	s.logger.Info("🔍 [Phase1] Checking strategies availability",
		zap.Int("strategies_count", len(s.strategies)),
		zap.String("url", targetURL))
	
	if len(s.strategies) > 0 {
		s.logger.Info("✅ [Phase1] Strategies available, using ScrapeWithStructuredContent",
			zap.String("url", targetURL),
			zap.Int("strategy_count", len(s.strategies)))
		structuredContent, err := s.ScrapeWithStructuredContent(ctx, targetURL)
		if err == nil && structuredContent != nil {
			// Convert ScrapedContent to ScrapingResult for backward compatibility
			result := &ScrapingResult{
				URL:              targetURL,
				StatusCode:       200, // Assume success if we got content
				Content:          structuredContent.RawHTML,
				ContentType:      "text/html",
				ContentLength:    int64(len(structuredContent.RawHTML)),
				Headers:          make(map[string]string),
				FinalURL:         targetURL,
				ScrapedAt:        structuredContent.ScrapedAt,
				Duration:         time.Since(startTime),
				RetryCount:       0,
				StructuredContent: structuredContent,
			}

			s.logger.Info("Website scraping completed successfully (enhanced method)",
				zap.String("url", targetURL),
				zap.Float64("quality_score", structuredContent.QualityScore),
				zap.Int("word_count", structuredContent.WordCount),
				zap.Duration("duration", result.Duration))

			return result, nil
		}
		// If enhanced method failed, fall through to legacy method
		s.logger.Warn("Enhanced scraping failed, falling back to legacy method",
			zap.String("url", targetURL),
			zap.Error(err))
	}

	// Legacy method with retry logic (for backward compatibility)
	var lastError error
	var result *ScrapingResult

	// Retry loop
	var attempt int
	for attempt = 0; attempt <= s.config.MaxRetries; attempt++ {
		if attempt > 0 {
			s.logger.Info("Retrying website scraping",
				zap.String("url", targetURL),
				zap.Int("attempt", attempt+1),
				zap.Duration("delay", s.config.RetryDelay))

			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(s.config.RetryDelay):
			}
		}

		result, lastError = s.performScrape(ctx, targetURL, attempt)
		if lastError == nil {
			break
		}

		// Log error for this attempt
		s.logger.Warn("Scraping attempt failed",
			zap.String("url", targetURL),
			zap.Int("attempt", attempt+1),
			zap.Error(lastError))

		// Don't retry on certain errors
		if s.shouldNotRetry(lastError) {
			break
		}
	}

	// Set final result details
	if result != nil {
		result.Duration = time.Since(startTime)
		result.RetryCount = attempt
		if lastError != nil {
			result.Error = lastError.Error()
		}
	}

	if lastError != nil {
		return result, fmt.Errorf("all scraping attempts failed: %w", lastError)
	}

	s.logger.Info("Website scraping completed successfully",
		zap.String("url", targetURL),
		zap.Int("status_code", result.StatusCode),
		zap.Int64("content_length", result.ContentLength),
		zap.Duration("duration", result.Duration),
		zap.Int("retry_count", result.RetryCount))

	return result, nil
}

// performScrape performs a single scraping attempt
func (s *WebsiteScraper) performScrape(ctx context.Context, targetURL string, attempt int) (*ScrapingResult, error) {
	httpStartTime := time.Now()

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		s.logger.Error("Failed to create HTTP request",
			zap.String("url", targetURL),
			zap.Error(err),
			zap.String("stage", "request_creation"))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", s.config.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// Add rate limiting delay
	if attempt > 0 {
		time.Sleep(s.config.RateLimitDelay)
	}

	// Execute request
	resp, err := s.client.Do(req)
	httpDuration := time.Since(httpStartTime)
	if err != nil {
		s.logger.Error("HTTP request failed",
			zap.String("url", targetURL),
			zap.Error(err),
			zap.String("stage", "http_request"),
			zap.Duration("duration_ms", httpDuration))
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	s.logger.Info("HTTP request succeeded",
		zap.String("url", targetURL),
		zap.Int("status_code", resp.StatusCode),
		zap.String("content_type", resp.Header.Get("Content-Type")),
		zap.Duration("duration_ms", httpDuration))

	// Check status code
	if resp.StatusCode != 200 {
		s.logger.Warn("Non-200 status code",
			zap.String("url", targetURL),
			zap.Int("status_code", resp.StatusCode))
		// Continue anyway - may still get useful content
	}

	// Read response body with size limit
	bodyStartTime := time.Now()
	body, err := s.readResponseBody(resp)
	bodyDuration := time.Since(bodyStartTime)
	if err != nil {
		s.logger.Error("Failed to read response body",
			zap.String("url", targetURL),
			zap.Error(err),
			zap.String("stage", "body_read"),
			zap.Duration("duration_ms", bodyDuration))
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	s.logger.Info("Response body read",
		zap.String("url", targetURL),
		zap.Int("body_size_bytes", len(body)),
		zap.Duration("duration_ms", bodyDuration))

	// Extract headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Create result
	result := &ScrapingResult{
		URL:           targetURL,
		StatusCode:    resp.StatusCode,
		Content:       string(body),
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: int64(len(body)),
		Headers:       headers,
		FinalURL:      resp.Request.URL.String(),
		ScrapedAt:     time.Now(),
	}

	return result, nil
}

// readResponseBody reads the response body with size limit
func (s *WebsiteScraper) readResponseBody(resp *http.Response) ([]byte, error) {
	// Check content length header
	if resp.ContentLength > s.config.MaxResponseSize {
		return nil, fmt.Errorf("response too large: %d bytes (max: %d)", resp.ContentLength, s.config.MaxResponseSize)
	}

	// Read body with limit
	body, err := io.ReadAll(io.LimitReader(resp.Body, s.config.MaxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check actual size
	if int64(len(body)) > s.config.MaxResponseSize {
		return nil, fmt.Errorf("response body too large: %d bytes (max: %d)", len(body), s.config.MaxResponseSize)
	}

	return body, nil
}

// shouldNotRetry determines if an error should not be retried
func (s *WebsiteScraper) shouldNotRetry(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Don't retry on client errors (4xx)
	if strings.Contains(errStr, "HTTP error: 4") {
		return true
	}

	// Don't retry on context cancellation
	if strings.Contains(errStr, "context canceled") || strings.Contains(errStr, "context deadline exceeded") {
		return true
	}

	// Don't retry on invalid URLs
	if strings.Contains(errStr, "invalid URL") {
		return true
	}

	return false
}

// ScrapeMultipleWebsites scrapes multiple websites concurrently with rate limiting
func (s *WebsiteScraper) ScrapeMultipleWebsites(ctx context.Context, urls []string, maxConcurrency int) (map[string]*ScrapingResult, error) {
	if maxConcurrency <= 0 {
		maxConcurrency = 5 // Default concurrency
	}

	// Create semaphore for concurrency control
	semaphore := make(chan struct{}, maxConcurrency)
	results := make(map[string]*ScrapingResult)
	errors := make(map[string]error)

	// Create channel for results
	resultChan := make(chan struct {
		url    string
		result *ScrapingResult
		err    error
	}, len(urls))

	// Start scraping goroutines
	for _, url := range urls {
		go func(targetURL string) {
			semaphore <- struct{}{} // Acquire semaphore
			defer func() {
				<-semaphore // Release semaphore
			}()

			result, err := s.ScrapeWebsite(ctx, targetURL)
			resultChan <- struct {
				url    string
				result *ScrapingResult
				err    error
			}{targetURL, result, err}
		}(url)
	}

	// Collect results
	for i := 0; i < len(urls); i++ {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		case result := <-resultChan:
			if result.err != nil {
				errors[result.url] = result.err
				s.logger.Error("Failed to scrape website",
					zap.String("url", result.url),
					zap.Error(result.err))
			} else {
				results[result.url] = result.result
			}
		}
	}

	// Log summary
	s.logger.Info("Multiple website scraping completed",
		zap.Int("total_urls", len(urls)),
		zap.Int("successful", len(results)),
		zap.Int("failed", len(errors)))

	return results, nil
}

// ValidateWebsiteAccessibility checks if a website is accessible and scrapeable
func (s *WebsiteScraper) ValidateWebsiteAccessibility(ctx context.Context, targetURL string) (*AccessibilityResult, error) {
	result := &AccessibilityResult{
		URL:       targetURL,
		CheckedAt: time.Now(),
	}

	// Try to scrape the website
	scrapingResult, err := s.ScrapeWebsite(ctx, targetURL)
	if err != nil {
		result.Accessible = false
		result.Error = err.Error()
		return result, nil
	}

	// Check if content is meaningful
	result.Accessible = true
	result.StatusCode = scrapingResult.StatusCode
	result.ContentType = scrapingResult.ContentType
	result.ContentLength = scrapingResult.ContentLength

	// Additional checks
	result.HasRobotsTxt = s.checkRobotsTxt(ctx, targetURL)
	result.HasSitemap = s.checkSitemap(ctx, targetURL)
	result.IsBlocked = s.detectBlocking(scrapingResult)

	return result, nil
}

// AccessibilityResult represents the accessibility check result
type AccessibilityResult struct {
	URL           string    `json:"url"`
	Accessible    bool      `json:"accessible"`
	StatusCode    int       `json:"status_code"`
	ContentType   string    `json:"content_type"`
	ContentLength int64     `json:"content_length"`
	HasRobotsTxt  bool      `json:"has_robots_txt"`
	HasSitemap    bool      `json:"has_sitemap"`
	IsBlocked     bool      `json:"is_blocked"`
	Error         string    `json:"error,omitempty"`
	CheckedAt     time.Time `json:"checked_at"`
}

// checkRobotsTxt checks if the website has a robots.txt file
func (s *WebsiteScraper) checkRobotsTxt(ctx context.Context, baseURL string) bool {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return false
	}

	robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)

	req, err := http.NewRequestWithContext(ctx, "GET", robotsURL, nil)
	if err != nil {
		return false
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

// checkSitemap checks if the website has a sitemap
func (s *WebsiteScraper) checkSitemap(ctx context.Context, baseURL string) bool {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return false
	}

	sitemapURL := fmt.Sprintf("%s://%s/sitemap.xml", parsedURL.Scheme, parsedURL.Host)

	req, err := http.NewRequestWithContext(ctx, "GET", sitemapURL, nil)
	if err != nil {
		return false
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

// ============================================================================
// Scraper Strategy Implementations
// ============================================================================

// SimpleHTTPScraper is the fastest strategy, works ~60% of time
type SimpleHTTPScraper struct {
	client *http.Client
	logger *zap.Logger
}

func (s *SimpleHTTPScraper) Name() string {
	return "simple_http"
}

// getClientWithContextTimeout returns a client with timeout that respects context deadline
// FIX: Root Cause #3 - Ensure client timeout <= context deadline
func (s *SimpleHTTPScraper) getClientWithContextTimeout(ctx context.Context, baseClient *http.Client) *http.Client {
	// Safety check: if baseClient is nil, return nil (shouldn't happen, but defensive)
	if baseClient == nil {
		s.logger.Warn("⚠️ [Phase1] Base client is nil in getClientWithContextTimeout")
		return nil
	}
	
	// Check if context is already expired
	if ctx.Err() != nil {
		s.logger.Warn("⚠️ [Phase1] Context already expired in getClientWithContextTimeout",
			zap.Error(ctx.Err()))
		return baseClient // Return base client, context cancellation will handle it
	}
	
	// If context has no deadline, use base client as-is
	if deadline, ok := ctx.Deadline(); ok {
		timeRemaining := time.Until(deadline)
		// If context is already expired or has very little time, use base client
		if timeRemaining <= 0 {
			return baseClient
		}
		// Add small buffer (500ms) to ensure context cancellation happens before client timeout
		clientTimeout := timeRemaining - 500*time.Millisecond
		if clientTimeout < 0 {
			clientTimeout = 100 * time.Millisecond // Minimum timeout
		}
		// If calculated timeout is less than base client timeout, create new client
		if clientTimeout < baseClient.Timeout {
			return &http.Client{
				Transport: baseClient.Transport,
				Timeout:   clientTimeout,
			}
		}
	}
	// Use base client if context has no deadline or base timeout is already appropriate
	return baseClient
}

func (s *SimpleHTTPScraper) Scrape(ctx context.Context, targetURL string) (*ScrapedContent, error) {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		s.logger.Warn("⚠️ [Phase1] Context already cancelled before SimpleHTTP strategy",
			zap.String("url", targetURL),
			zap.Error(ctx.Err()))
		return nil, ctx.Err()
	default:
		// Context is still valid, proceed
	}
	
	// FIX: Create client with timeout that respects context deadline (Root Cause #3)
	client := s.getClientWithContextTimeout(ctx, s.client)
	if client == nil {
		return nil, fmt.Errorf("failed to create HTTP client: base client is nil")
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	// Basic headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; KYB-Bot/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	content := extractStructuredContent(doc, string(body), targetURL)
	return content, nil
}

// BrowserHeadersScraper uses realistic browser headers, works ~80% of time
type BrowserHeadersScraper struct {
	client *http.Client
	logger *zap.Logger
}

func (s *BrowserHeadersScraper) Name() string {
	return "browser_headers"
}

// getClientWithContextTimeout returns a client with timeout that respects context deadline
// FIX: Root Cause #3 - Ensure client timeout <= context deadline
func (s *BrowserHeadersScraper) getClientWithContextTimeout(ctx context.Context, baseClient *http.Client) *http.Client {
	// Safety check: if baseClient is nil, return nil (shouldn't happen, but defensive)
	if baseClient == nil {
		s.logger.Warn("⚠️ [Phase1] Base client is nil in getClientWithContextTimeout")
		return nil
	}
	
	// Check if context is already expired
	if ctx.Err() != nil {
		s.logger.Warn("⚠️ [Phase1] Context already expired in getClientWithContextTimeout",
			zap.Error(ctx.Err()))
		return baseClient // Return base client, context cancellation will handle it
	}
	
	// If context has no deadline, use base client as-is
	if deadline, ok := ctx.Deadline(); ok {
		timeRemaining := time.Until(deadline)
		// If context is already expired or has very little time, use base client
		if timeRemaining <= 0 {
			return baseClient
		}
		// Add small buffer (500ms) to ensure context cancellation happens before client timeout
		clientTimeout := timeRemaining - 500*time.Millisecond
		if clientTimeout < 0 {
			clientTimeout = 100 * time.Millisecond // Minimum timeout
		}
		// If calculated timeout is less than base client timeout, create new client
		if clientTimeout < baseClient.Timeout {
			return &http.Client{
				Transport: baseClient.Transport,
				Timeout:   clientTimeout,
			}
		}
	}
	// Use base client if context has no deadline or base timeout is already appropriate
	return baseClient
}

func (s *BrowserHeadersScraper) Scrape(ctx context.Context, targetURL string) (*ScrapedContent, error) {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		s.logger.Warn("⚠️ [Phase1] Context already cancelled before BrowserHeaders strategy",
			zap.String("url", targetURL),
			zap.Error(ctx.Err()))
		return nil, ctx.Err()
	default:
		// Context is still valid, proceed
	}
	
	// FIX: Create client with timeout that respects context deadline (Root Cause #3)
	client := s.getClientWithContextTimeout(ctx, s.client)
	if client == nil {
		return nil, fmt.Errorf("failed to create HTTP client: base client is nil")
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	// Realistic browser headers to avoid bot detection
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	// Handle gzip
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	content := extractStructuredContent(doc, string(body), targetURL)
	return content, nil
}

// PlaywrightScraper uses external Playwright service for JS-heavy sites, works ~95% of time
type PlaywrightScraper struct {
	serviceURL string
	client     *http.Client
	logger     *zap.Logger
}

func (s *PlaywrightScraper) Name() string {
	return "playwright"
}

// getClientWithContextTimeout returns a client with timeout that respects context deadline
// FIX: Root Cause #3 - Ensure client timeout <= context deadline
func (s *PlaywrightScraper) getClientWithContextTimeout(ctx context.Context, baseClient *http.Client) *http.Client {
	// Safety check: if baseClient is nil, return nil (shouldn't happen, but defensive)
	if baseClient == nil {
		s.logger.Warn("⚠️ [Phase1] Base client is nil in getClientWithContextTimeout")
		return nil
	}
	
	// Check if context is already expired
	if ctx.Err() != nil {
		s.logger.Warn("⚠️ [Phase1] Context already expired in getClientWithContextTimeout",
			zap.Error(ctx.Err()))
		return baseClient // Return base client, context cancellation will handle it
	}
	
	// If context has no deadline, use base client as-is
	if deadline, ok := ctx.Deadline(); ok {
		timeRemaining := time.Until(deadline)
		// If context is already expired or has very little time, use base client
		if timeRemaining <= 0 {
			return baseClient
		}
		// FIX: For Playwright, account for queue wait time (up to 25s) in timeout calculation
		// Queue wait can be significant, so we need more buffer
		// Add buffer (1s) to ensure context cancellation happens before client timeout
		// But ensure we have at least 30s for queue wait (25s) + scrape start (5s)
		const minPlaywrightTimeout = 30 * time.Second
		clientTimeout := timeRemaining - 1*time.Second
		if clientTimeout < minPlaywrightTimeout {
			// If context doesn't have enough time, use minimum (will likely timeout, but try anyway)
			clientTimeout = minPlaywrightTimeout
		}
		// If calculated timeout is less than base client timeout, create new client
		if clientTimeout < baseClient.Timeout {
			return &http.Client{
				Transport: baseClient.Transport,
				Timeout:   clientTimeout,
			}
		}
	}
	// Use base client if context has no deadline or base timeout is already appropriate
	return baseClient
}

func (s *PlaywrightScraper) Scrape(ctx context.Context, targetURL string) (*ScrapedContent, error) {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		s.logger.Warn("⚠️ [Phase1] Context already cancelled before Playwright strategy",
			zap.String("url", targetURL),
			zap.Error(ctx.Err()))
		return nil, ctx.Err()
	default:
		// Context is still valid, proceed
	}
	
	// FIX: Create client with timeout that respects context deadline (Root Cause #3)
	client := s.getClientWithContextTimeout(ctx, s.client)
	if client == nil {
		return nil, fmt.Errorf("failed to create HTTP client: base client is nil")
	}
	
	// Call Playwright service
	reqBody, _ := json.Marshal(map[string]string{"url": targetURL})

	req, err := http.NewRequestWithContext(ctx, "POST", s.serviceURL+"/scrape", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("playwright service error: %d %s", resp.StatusCode, resp.Status)
	}

	var result struct {
		HTML    string `json:"html"`
		Error   string `json:"error"`
		Success bool   `json:"success"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("playwright error: %s", result.Error)
	}

	if !result.Success {
		return nil, fmt.Errorf("playwright service returned failure")
	}

	// Parse HTML from Playwright
	doc, err := html.Parse(strings.NewReader(result.HTML))
	if err != nil {
		return nil, err
	}

	content := extractStructuredContent(doc, result.HTML, targetURL)
	return content, nil
}

// ============================================================================
// Structured Content Extraction Functions
// ============================================================================

// extractStructuredContent extracts structured content from HTML document
func extractStructuredContent(doc *html.Node, rawHTML string, targetURL string) *ScrapedContent {
	content := &ScrapedContent{
		RawHTML:   rawHTML,
		ScrapedAt: time.Now(),
	}

	// Extract domain
	if parsedURL, err := url.Parse(targetURL); err == nil {
		content.Domain = parsedURL.Host
	} else {
		content.Domain = ""
	}

	// Extract title
	content.Title = extractTitle(doc)

	// Extract meta description
	content.MetaDesc = extractMetaDescription(doc)

	// Extract all headings (H1-H3) - high signal for classification
	content.Headings = extractHeadings(doc)

	// Extract navigation menu - indicates business areas
	content.NavMenu = extractNavigation(doc)

	// Extract "About" section - highest quality content
	content.AboutText = extractAboutSection(doc)

	// Extract product/service listings
	content.ProductList = extractProductsServices(doc)

	// Extract contact page content
	content.ContactInfo = extractContactInfo(doc)

	// Detect language
	content.Language = detectLanguage(doc)

	// Check for logo (indicator of legitimate business)
	content.HasLogo = hasLogo(doc)

	// Combine all text with priority weighting
	content.PlainText = combineTextWithWeights(content)

	// Count words
	content.WordCount = len(strings.Fields(content.PlainText))

	// Calculate quality score
	content.QualityScore = calculateContentQuality(content)

	return content
}

// extractTitle extracts the title from HTML
func extractTitle(doc *html.Node) string {
	if title := findNodeByTag(doc, "title"); title != nil {
		text := extractText(title)
		if text != "" {
			return text
		}
	}

	// Fallback: look for h1
	if h1 := findNodeByTag(doc, "h1"); h1 != nil {
		return extractText(h1)
	}

	return ""
}

// extractMetaDescription extracts meta description
func extractMetaDescription(doc *html.Node) string {
	var description string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var isDescription bool
			var content string

			for _, attr := range n.Attr {
				if attr.Key == "name" && attr.Val == "description" {
					isDescription = true
				}
				if attr.Key == "content" {
					content = attr.Val
				}
			}

			if isDescription && content != "" {
				description = content
				return
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return description
}

// extractHeadings extracts H1, H2, H3 headings
func extractHeadings(doc *html.Node) []string {
	headings := []string{}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "h1" || n.Data == "h2" || n.Data == "h3" {
				text := extractText(n)
				if text != "" && len(text) < 200 { // Reasonable heading length
					headings = append(headings, text)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return headings
}

// extractNavigation extracts navigation menu items
func extractNavigation(doc *html.Node) []string {
	navItems := []string{}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "nav" {
			// Found nav element, extract all links
			extractLinks(n, &navItems)
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return navItems
}

// extractAboutSection extracts "About" section - HIGHEST PRIORITY
func extractAboutSection(doc *html.Node) string {
	aboutText := ""

	// Strategy 1: Look for section/div with id/class containing "about"
	identifiers := []string{"about", "about-us", "aboutus", "company", "who-we-are", "who-we-are", "our-story", "our-mission", "overview", "intro", "introduction"}
	aboutSection := findNodeWithIdentifier(doc, identifiers)
	if aboutSection != nil {
		aboutText = extractText(aboutSection)
		if len(aboutText) > 100 { // Substantial content
			return aboutText
		}
	}

	// Strategy 2: Look for headings containing "about" and extract following paragraphs
	var foundAboutText string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if foundAboutText != "" {
			return // Already found
		}
		if n.Type == html.ElementNode && (n.Data == "h1" || n.Data == "h2" || n.Data == "h3") {
			headingText := strings.ToLower(extractText(n))
			if strings.Contains(headingText, "about") || strings.Contains(headingText, "company") || strings.Contains(headingText, "who we are") {
				// Found about heading, extract following content
				next := n.NextSibling
				content := []string{}
				paragraphCount := 0
				maxParagraphs := 5 // Limit to avoid grabbing too much
				
				for next != nil && paragraphCount < maxParagraphs {
					if next.Type == html.ElementNode && (next.Data == "p" || next.Data == "div" || next.Data == "section") {
						text := extractText(next)
						text = strings.TrimSpace(text)
						if len(text) > 20 { // Meaningful paragraph
							content = append(content, text)
							paragraphCount++
						}
					}
					next = next.NextSibling
				}
				
				if len(content) > 0 {
					combined := strings.Join(content, " ")
					if len(combined) > 100 {
						foundAboutText = combined
						return
					}
				}
			}
		}
		
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	if foundAboutText != "" {
		return foundAboutText
	}

	// Strategy 3: Look for main content area (article, main, or first large div)
	mainContent := findNodeByTag(doc, "main")
	if mainContent == nil {
		mainContent = findNodeByTag(doc, "article")
	}
	if mainContent != nil {
		text := extractText(mainContent)
		// Take first 500 characters as about text if substantial
		if len(text) > 200 {
			// Find first paragraph or section
			var extractFirstParagraph func(*html.Node) string
			extractFirstParagraph = func(n *html.Node) string {
				if n.Type == html.ElementNode && n.Data == "p" {
					text := extractText(n)
					if len(text) > 100 {
						return text
					}
				}
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if result := extractFirstParagraph(c); result != "" {
						return result
					}
				}
				return ""
			}
			if para := extractFirstParagraph(mainContent); para != "" {
				return para
			}
			// Fallback: return first 500 chars
			if len(text) > 500 {
				return text[:500]
			}
			return text
		}
	}

	return aboutText
}

// extractProductsServices extracts products/services listings
func extractProductsServices(doc *html.Node) []string {
	products := []string{}

	// Look for common product/service listing patterns
	productSection := findNodeWithIdentifier(doc, []string{
		"products", "services", "menu", "offerings", "solutions",
	})

	if productSection != nil {
		// Extract list items
		var f func(*html.Node)
		f = func(n *html.Node) {
			if n.Type == html.ElementNode && (n.Data == "li" || n.Data == "h3" || n.Data == "h4") {
				text := extractText(n)
				if text != "" && len(text) < 100 { // Reasonable product name
					products = append(products, text)
				}
			}

			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
		f(productSection)
	}

	return products
}

// extractContactInfo extracts contact page content
func extractContactInfo(doc *html.Node) string {
	contactSection := findNodeWithIdentifier(doc, []string{"contact", "contact-us", "get-in-touch"})
	if contactSection != nil {
		return extractText(contactSection)
	}
	return ""
}

// combineTextWithWeights combines text with priority weighting
func combineTextWithWeights(content *ScrapedContent) string {
	parts := []string{}

	// Title (highest weight - repeat 3x)
	if content.Title != "" {
		parts = append(parts, content.Title, content.Title, content.Title)
	}

	// Meta description (high weight - repeat 2x)
	if content.MetaDesc != "" {
		parts = append(parts, content.MetaDesc, content.MetaDesc)
	}

	// About text (high weight - repeat 2x)
	if content.AboutText != "" {
		parts = append(parts, content.AboutText, content.AboutText)
	}

	// Headings (medium weight - repeat 1x)
	if len(content.Headings) > 0 {
		parts = append(parts, strings.Join(content.Headings, ". "))
	}

	// Navigation (medium weight)
	if len(content.NavMenu) > 0 {
		parts = append(parts, strings.Join(content.NavMenu, ". "))
	}

	// Products (lower weight)
	if len(content.ProductList) > 0 {
		parts = append(parts, strings.Join(content.ProductList, ". "))
	}

	return strings.Join(parts, ". ")
}

// calculateContentQuality calculates quality score (0.0 - 1.0)
func calculateContentQuality(content *ScrapedContent) float64 {
	score := 0.0

	// Has title? +0.15
	if content.Title != "" {
		score += 0.15
	}

	// Has meta description? +0.15
	if content.MetaDesc != "" {
		score += 0.15
	}

	// Has headings? +0.15
	if len(content.Headings) > 0 {
		score += 0.15
	}

	// Has about section? +0.20 (most important)
	if content.AboutText != "" && len(content.AboutText) > 100 {
		score += 0.20
	}

	// Sufficient word count? +0.15
	if content.WordCount >= 200 {
		score += 0.15
	}

	// Has navigation? +0.10
	if len(content.NavMenu) > 0 {
		score += 0.10
	}

	// Has logo? +0.10
	if content.HasLogo {
		score += 0.10
	}

	return math.Min(score, 1.0)
}

// Helper: Find node by tag
func findNodeByTag(n *html.Node, tag string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findNodeByTag(c, tag); result != nil {
			return result
		}
	}

	return nil
}

// Helper: Find node with identifier in id/class
func findNodeWithIdentifier(n *html.Node, identifiers []string) *html.Node {
	if n.Type == html.ElementNode {
		for _, attr := range n.Attr {
			if attr.Key == "id" || attr.Key == "class" {
				lowerVal := strings.ToLower(attr.Val)
				for _, identifier := range identifiers {
					if strings.Contains(lowerVal, identifier) {
						return n
					}
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findNodeWithIdentifier(c, identifiers); result != nil {
			return result
		}
	}

	return nil
}

// Helper: Extract text from node
func extractText(n *html.Node) string {
	var text strings.Builder

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			text.WriteString(n.Data)
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	// Clean whitespace
	result := strings.TrimSpace(text.String())
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")

	return result
}

// Helper: Extract links from node
func extractLinks(n *html.Node, items *[]string) {
	if n.Type == html.ElementNode && n.Data == "a" {
		text := extractText(n)
		if text != "" && len(text) < 100 {
			*items = append(*items, text)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractLinks(c, items)
	}
}

// detectLanguage detects the language of the content (simplified - defaults to "en")
func detectLanguage(doc *html.Node) string {
	// Check for lang attribute
	var f func(*html.Node) string
	f = func(n *html.Node) string {
		if n.Type == html.ElementNode {
			for _, attr := range n.Attr {
				if attr.Key == "lang" {
					return attr.Val
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if lang := f(c); lang != "" {
				return lang
			}
		}
		return ""
	}

	lang := f(doc)
	if lang != "" {
		return lang
	}

	return "en" // Default to English
}

// hasLogo checks if the page has a logo (simplified check for img tags)
func hasLogo(doc *html.Node) bool {
	var hasImg bool
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, attr := range n.Attr {
				if attr.Key == "alt" {
					alt := strings.ToLower(attr.Val)
					if strings.Contains(alt, "logo") {
						hasImg = true
						return
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return hasImg
}

// detectBlocking detects if the website is blocking scraping
func (s *WebsiteScraper) detectBlocking(result *ScrapingResult) bool {
	if result == nil {
		return false
	}

	content := strings.ToLower(result.Content)

	// Check for common blocking indicators
	blockingIndicators := []string{
		"access denied",
		"blocked",
		"captcha",
		"cloudflare",
		"forbidden",
		"rate limit",
		"too many requests",
		"please wait",
		"checking your browser",
	}

	for _, indicator := range blockingIndicators {
		if strings.Contains(content, indicator) {
			return true
		}
	}

	// Check for blocking status codes
	if result.StatusCode == 403 || result.StatusCode == 429 {
		return true
	}

	return false
}

// ScrapeWithStructuredContent scrapes a website and extracts structured content
// This is the enhanced method that uses multi-tier strategies and structured extraction
func (s *WebsiteScraper) ScrapeWithStructuredContent(ctx context.Context, targetURL string) (*ScrapedContent, error) {
	startTime := time.Now()
	
	// CRITICAL: Log immediately to verify function is being called with new code
	s.logger.Info("🔍 [Phase1] ScrapeWithStructuredContent ENTRY",
		zap.String("url", targetURL),
		zap.Time("start_time", startTime),
		zap.String("function", "ScrapeWithStructuredContent"))
	
	// IMMEDIATELY check context state - this must be first
	ctxErr := ctx.Err()
	s.logger.Info("🔍 [Phase1] [ContextCheck] Checking context state at function entry",
		zap.String("url", targetURL),
		zap.Any("context_err", ctxErr),
		zap.Bool("context_cancelled", ctxErr != nil))
	
	if ctxErr != nil {
		s.logger.Error("❌ [Phase1] Context already cancelled before scraping",
			zap.String("url", targetURL),
			zap.Error(ctxErr),
			zap.String("error_type", fmt.Sprintf("%T", ctxErr)))
		return nil, fmt.Errorf("context already cancelled: %w", ctxErr)
	}
	
	// Log context deadline and HTTP client timeout for debugging
	var contextDeadline time.Duration
	var clientTimeout time.Duration
	var hasDeadline bool
	
	if deadline, ok := ctx.Deadline(); ok {
		hasDeadline = true
		contextDeadline = time.Until(deadline)
		s.logger.Info("⏱️ [Phase1] [ContextCheck] Context deadline check - HAS DEADLINE",
			zap.String("url", targetURL),
			zap.Duration("time_until_deadline", contextDeadline),
			zap.Time("deadline", deadline),
			zap.Time("current_time", time.Now()),
			zap.Bool("context_valid", ctx.Err() == nil),
			zap.String("deadline_string", deadline.Format(time.RFC3339Nano)))
	} else {
		hasDeadline = false
		s.logger.Info("⏱️ [Phase1] [ContextCheck] Context deadline check - NO DEADLINE",
			zap.String("url", targetURL),
			zap.Bool("context_valid", ctx.Err() == nil))
	}
	
	// Log HTTP client timeout for strategies
	if s.client != nil {
		clientTimeout = s.client.Timeout
		s.logger.Info("⏱️ [Phase1] [HTTPClient] HTTP client timeout configuration",
			zap.String("url", targetURL),
			zap.Duration("client_timeout", clientTimeout),
			zap.Duration("context_deadline", contextDeadline),
			zap.Bool("has_context_deadline", hasDeadline),
			zap.Bool("client_will_limit", clientTimeout > 0 && hasDeadline && clientTimeout < contextDeadline))
	} else {
		s.logger.Warn("⚠️ [Phase1] [HTTPClient] HTTP client is nil",
			zap.String("url", targetURL))
	}

	// Validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		s.logger.Error("Invalid URL",
			zap.String("url", targetURL),
			zap.Error(err),
			zap.String("stage", "url_validation"))
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme == "" {
		targetURL = "https://" + targetURL
		parsedURL, err = url.Parse(targetURL)
		if err != nil {
			s.logger.Error("Invalid URL after adding scheme",
				zap.String("url", targetURL),
				zap.Error(err),
				zap.String("stage", "url_validation"))
			return nil, fmt.Errorf("invalid URL after adding scheme: %w", err)
		}
	}

	s.logger.Info("🌐 [Phase1] Starting scrape with structured content extraction",
		zap.String("url", targetURL),
		zap.Time("timestamp", startTime))

	// Try strategies in order
	var lastErr error
	var content *ScrapedContent

	for i, strategy := range s.strategies {
		strategyStartTime := time.Now()
		s.logger.Info("🔍 [Phase1] Attempting scrape strategy",
			zap.String("strategy", strategy.Name()),
			zap.String("url", targetURL),
			zap.Int("attempt", i+1))

		// CRITICAL DEBUG: Log right before strategy call
		s.logger.Info("🔍 [Phase1] [DEBUG] About to call strategy.Scrape()",
			zap.String("strategy", strategy.Name()),
			zap.String("url", targetURL))

		content, err = strategy.Scrape(ctx, targetURL)
		strategyDuration := time.Since(strategyStartTime)

		// CRITICAL DEBUG: Log immediately after strategy returns
		var errorMsg string
		if err != nil {
			errorMsg = err.Error()
		} else {
			errorMsg = "none"
		}
		s.logger.Info("🔍 [Phase1] [DEBUG] Strategy.Scrape() returned",
			zap.String("strategy", strategy.Name()),
			zap.Bool("content_is_nil", content == nil),
			zap.Bool("error_is_nil", err == nil),
			zap.String("error", errorMsg),
			zap.Duration("duration", strategyDuration))

		// CRITICAL: Log immediately after strategy returns to verify execution path
		s.logger.Info("🔍 [Phase1] [StrategyResult] Strategy returned",
			zap.String("strategy", strategy.Name()),
			zap.Bool("content_is_nil", content == nil),
			zap.Bool("error_is_nil", err == nil),
			zap.String("error_msg", errorMsg),
			zap.Duration("duration", strategyDuration))

		// Log detailed validation information
		if content != nil {
			s.logger.Info("📊 [Phase1] [Validation] Content received from strategy, logging validation details",
				zap.String("strategy", strategy.Name()),
				zap.Bool("content_is_nil", content == nil),
				zap.Float64("quality_score", content.QualityScore),
				zap.Int("word_count", content.WordCount))
			logContentValidationDetails(s.logger, content, strategy.Name())
		} else {
			s.logger.Warn("⚠️ [Phase1] [Validation] Content is nil from strategy",
				zap.String("strategy", strategy.Name()),
				zap.Error(err))
		}

		// Use logging version for better debugging
		isValid := content != nil && err == nil && isContentValidWithLogging(content, s.logger, strategy.Name())
		if err == nil && content != nil && isValid {
			s.logger.Info("✅ [Phase1] Strategy succeeded",
				zap.String("strategy", strategy.Name()),
				zap.Float64("quality_score", content.QualityScore),
				zap.Int("word_count", content.WordCount),
				zap.Duration("strategy_duration_ms", strategyDuration),
				zap.Duration("total_duration_ms", time.Since(startTime)))
			return content, nil
		}

		lastErr = err
		qualityScore := 0.0
		wordCount := 0
		hasTitle := false
		hasMetaDesc := false
		if content != nil {
			qualityScore = content.QualityScore
			wordCount = content.WordCount
			hasTitle = content.Title != ""
			hasMetaDesc = content.MetaDesc != ""
		}
		s.logger.Warn("⚠️ [Phase1] Strategy failed, trying next",
			zap.String("strategy", strategy.Name()),
			zap.Error(err),
			zap.Float64("quality_score", qualityScore),
			zap.Int("word_count", wordCount),
			zap.Bool("has_title", hasTitle),
			zap.Bool("has_meta_desc", hasMetaDesc),
			zap.Bool("meets_word_count", wordCount >= 50),
			zap.Bool("meets_quality_threshold", qualityScore >= 0.5),
			zap.Duration("strategy_duration_ms", strategyDuration))
	}

	s.logger.Error("❌ [Phase1] All scraping strategies failed",
		zap.String("url", targetURL),
		zap.Error(lastErr),
		zap.Duration("total_duration_ms", time.Since(startTime)))

	return nil, fmt.Errorf("all scraping strategies failed: %w", lastErr)
}

// logContentValidationDetails logs detailed validation information for debugging
func logContentValidationDetails(logger *zap.Logger, content *ScrapedContent, strategyName string) {
	logger.Info("🔍 [Phase1] [Validation] logContentValidationDetails ENTRY",
		zap.String("strategy", strategyName),
		zap.Bool("content_is_nil", content == nil))
	
	if content == nil {
		logger.Warn("⚠️ [Phase1] [Validation] Content is nil",
			zap.String("strategy", strategyName))
		return
	}
	
	// Log all content fields for debugging
	logger.Info("📋 [Phase1] [Validation] Content fields",
		zap.String("strategy", strategyName),
		zap.String("title", content.Title),
		zap.String("meta_desc", content.MetaDesc),
		zap.Int("headings_count", len(content.Headings)),
		zap.Int("nav_count", len(content.NavMenu)),
		zap.String("about_text_preview", truncateString(content.AboutText, 100)),
		zap.Int("products_count", len(content.ProductList)),
		zap.String("contact_info_preview", truncateString(content.ContactInfo, 100)),
		zap.Int("plain_text_length", len(content.PlainText)))
	
	logger.Info("🔍 [Phase1] [Validation] Content validation details",
		zap.String("strategy", strategyName),
		zap.Int("word_count", content.WordCount),
		zap.Bool("has_title", content.Title != ""),
		zap.Bool("has_meta_desc", content.MetaDesc != ""),
		zap.Float64("quality_score", content.QualityScore),
		zap.Bool("meets_word_count", content.WordCount >= 50),
		zap.Bool("meets_metadata", content.Title != "" || content.MetaDesc != ""),
		zap.Bool("meets_quality_threshold", content.QualityScore >= 0.5),
		zap.Bool("is_valid", isContentValid(content)))
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// isContentValid validates content quality before accepting it
func isContentValid(content *ScrapedContent) bool {
	if content == nil {
		return false
	}

	// Minimum word count
	if content.WordCount < 50 {
		return false
	}

	// Must have basic metadata
	if content.Title == "" && content.MetaDesc == "" {
		return false
	}

	// Check for error pages
	if containsErrorIndicators(content.PlainText) {
		return false
	}

	// Quality score threshold
	if content.QualityScore < 0.5 {
		return false
	}

	return true
}

// isContentValidWithLogging validates content quality and logs the reason for failure
func isContentValidWithLogging(content *ScrapedContent, logger *zap.Logger, strategyName string) bool {
	if content == nil {
		logger.Debug("❌ [Phase1] [Validation] Content is nil")
		return false
	}

	// Minimum word count
	if content.WordCount < 50 {
		logger.Debug("❌ [Phase1] [Validation] Word count too low",
			zap.Int("word_count", content.WordCount),
			zap.Int("required", 50))
		return false
	}

	// Must have basic metadata
	if content.Title == "" && content.MetaDesc == "" {
		logger.Debug("❌ [Phase1] [Validation] Missing title and meta description")
		return false
	}

	// Check for error pages
	if containsErrorIndicators(content.PlainText) {
		logger.Debug("❌ [Phase1] [Validation] Contains error indicators",
			zap.String("strategy", strategyName))
		return false
	}

	// Quality score threshold
	if content.QualityScore < 0.5 {
		logger.Debug("❌ [Phase1] [Validation] Quality score too low",
			zap.Float64("quality_score", content.QualityScore),
			zap.Float64("required", 0.5))
		return false
	}

	logger.Debug("✅ [Phase1] [Validation] Content is valid",
		zap.String("strategy", strategyName),
		zap.Int("word_count", content.WordCount),
		zap.Float64("quality_score", content.QualityScore))
	return true
}

// containsErrorIndicators checks if text contains error page indicators
// Uses more specific patterns to avoid false positives (e.g., "error handling" is legitimate)
func containsErrorIndicators(text string) bool {
	lowerText := strings.ToLower(text)
	
	// More specific error patterns to avoid false positives
	errorPatterns := []string{
		"404", "not found", "page not found", "page cannot be found",
		"403", "access denied", "forbidden",
		"500", "internal server error", "500 error",
		"503", "service unavailable", "503 error",
		"oops! something went wrong",
		"this page doesn't exist",
		"the page you're looking for",
		"we couldn't find that page",
	}
	
	// Check for specific error patterns (not just "error" which is too broad)
	for _, pattern := range errorPatterns {
		if strings.Contains(lowerText, pattern) {
			return true
		}
	}
	
	// Check for HTTP status codes at start of text (common in error pages)
	if matched, _ := regexp.MatchString(`^(404|403|500|503|502|504)`, strings.TrimSpace(lowerText)); matched {
		return true
	}

	return false
}
