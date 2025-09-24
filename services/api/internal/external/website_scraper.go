package external

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
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

// WebsiteScraper handles website scraping with robust error handling and retry logic
type WebsiteScraper struct {
	client *http.Client
	config *ScrapingConfig
	logger *zap.Logger
}

// NewWebsiteScraper creates a new website scraper with the given configuration
func NewWebsiteScraper(config *ScrapingConfig, logger *zap.Logger) *WebsiteScraper {
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

	return &WebsiteScraper{
		client: client,
		config: config,
		logger: logger,
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
}

// ScrapeWebsite scrapes a website with retry logic and comprehensive error handling
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
	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
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
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	// Read response body with size limit
	body, err := s.readResponseBody(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

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
