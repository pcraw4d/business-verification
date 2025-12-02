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
func NewWebsiteScraper(config *ScrapingConfig, logger *zap.Logger) *WebsiteScraper {
	return NewWebsiteScraperWithStrategies(config, logger, "")
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
	strategies := []ScraperStrategy{
		&SimpleHTTPScraper{client: client, logger: logger},
		&BrowserHeadersScraper{client: client, logger: logger},
	}

	// Add Playwright strategy if URL is provided
	if playwrightServiceURL != "" {
		playwrightClient := &http.Client{Timeout: 30 * time.Second}
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
	if len(s.strategies) > 0 {
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

func (s *SimpleHTTPScraper) Scrape(ctx context.Context, targetURL string) (*ScrapedContent, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	// Basic headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; KYB-Bot/1.0)")

	resp, err := s.client.Do(req)
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

func (s *BrowserHeadersScraper) Scrape(ctx context.Context, targetURL string) (*ScrapedContent, error) {
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

	resp, err := s.client.Do(req)
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

func (s *PlaywrightScraper) Scrape(ctx context.Context, targetURL string) (*ScrapedContent, error) {
	// Call Playwright service
	reqBody, _ := json.Marshal(map[string]string{"url": targetURL})

	req, err := http.NewRequestWithContext(ctx, "POST", s.serviceURL+"/scrape", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
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
	aboutSection := findNodeWithIdentifier(doc, []string{"about", "about-us", "company", "who-we-are"})
	if aboutSection != nil {
		aboutText = extractText(aboutSection)
		if len(aboutText) > 100 { // Substantial content
			return aboutText
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

	s.logger.Info("Starting scrape with structured content extraction",
		zap.String("url", targetURL),
		zap.Time("timestamp", startTime))

	// Try strategies in order
	var lastErr error
	var content *ScrapedContent

	for i, strategy := range s.strategies {
		strategyStartTime := time.Now()
		s.logger.Info("Attempting scrape strategy",
			zap.String("strategy", strategy.Name()),
			zap.String("url", targetURL),
			zap.Int("attempt", i+1))

		content, err = strategy.Scrape(ctx, targetURL)
		strategyDuration := time.Since(strategyStartTime)

		if err == nil && content != nil && isContentValid(content) {
			s.logger.Info("Strategy succeeded",
				zap.String("strategy", strategy.Name()),
				zap.Float64("quality_score", content.QualityScore),
				zap.Int("word_count", content.WordCount),
				zap.Duration("strategy_duration_ms", strategyDuration),
				zap.Duration("total_duration_ms", time.Since(startTime)))
			return content, nil
		}

		lastErr = err
		qualityScore := 0.0
		if content != nil {
			qualityScore = content.QualityScore
		}
		s.logger.Warn("Strategy failed, trying next",
			zap.String("strategy", strategy.Name()),
			zap.Error(err),
			zap.Float64("quality_score", qualityScore),
			zap.Duration("strategy_duration_ms", strategyDuration))
	}

	s.logger.Error("All scraping strategies failed",
		zap.String("url", targetURL),
		zap.Error(lastErr),
		zap.Duration("total_duration_ms", time.Since(startTime)))

	return nil, fmt.Errorf("all scraping strategies failed: %w", lastErr)
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

// containsErrorIndicators checks if text contains error page indicators
func containsErrorIndicators(text string) bool {
	lowerText := strings.ToLower(text)
	errorIndicators := []string{
		"404", "not found", "page not found",
		"403", "access denied", "forbidden",
		"500", "internal server error",
		"503", "service unavailable",
		"error", "oops",
	}

	for _, indicator := range errorIndicators {
		if strings.Contains(lowerText, indicator) {
			return true
		}
	}

	return false
}
