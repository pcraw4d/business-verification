package webanalysis

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ScrapedContent represents the content extracted from a website
type ScrapedContent struct {
	URL           string            `json:"url"`
	Title         string            `json:"title"`
	HTML          string            `json:"html"`
	Text          string            `json:"text"`
	Headers       map[string]string `json:"headers"`
	StatusCode    int               `json:"status_code"`
	ResponseTime  time.Duration     `json:"response_time"`
	ProxyUsed     *Proxy            `json:"proxy_used"`
	ExtractedData map[string]string `json:"extracted_data"`
	Error         string            `json:"error,omitempty"`
}

// ScrapingJob represents a scraping task
type ScrapingJob struct {
	URL        string            `json:"url"`
	Business   string            `json:"business"`
	Priority   int               `json:"priority"`
	Retries    int               `json:"retries"`
	MaxRetries int               `json:"max_retries"`
	Timeout    time.Duration     `json:"timeout"`
	Headers    map[string]string `json:"headers"`
	UserAgent  string            `json:"user_agent"`
}

// WebScraper manages web scraping operations
type WebScraper struct {
	proxyMgr    *ProxyManager
	rateLimiter *RateLimiter
	config      ScraperConfig
	userAgents  []string
	mu          sync.RWMutex
}

// ScraperConfig holds configuration for web scraping
type ScraperConfig struct {
	DefaultTimeout   time.Duration `json:"default_timeout"`
	MaxRetries       int           `json:"max_retries"`
	RetryDelay       time.Duration `json:"retry_delay"`
	MaxConcurrent    int           `json:"max_concurrent"`
	RateLimitPerSec  int           `json:"rate_limit_per_sec"`
	FollowRedirects  bool          `json:"follow_redirects"`
	RespectRobotsTxt bool          `json:"respect_robots_txt"`
	ExtractImages    bool          `json:"extract_images"`
	ExtractLinks     bool          `json:"extract_links"`
}

// RateLimiter manages request rate limiting
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

// NewWebScraper creates a new web scraper instance
func NewWebScraper(proxyMgr *ProxyManager) *WebScraper {
	return &WebScraper{
		proxyMgr: proxyMgr,
		rateLimiter: &RateLimiter{
			requests: make(map[string][]time.Time),
			limit:    2, // 2 requests per second per domain
			window:   time.Second,
		},
		config: ScraperConfig{
			DefaultTimeout:   time.Second * 30,
			MaxRetries:       3,
			RetryDelay:       time.Second * 2,
			MaxConcurrent:    10,
			RateLimitPerSec:  2,
			FollowRedirects:  true,
			RespectRobotsTxt: true,
			ExtractImages:    false,
			ExtractLinks:     true,
		},
		userAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:89.0) Gecko/20100101 Firefox/89.0",
		},
	}
}

// ScrapeWebsite scrapes a website using the provided job configuration
func (ws *WebScraper) ScrapeWebsite(job *ScrapingJob) (*ScrapedContent, error) {
	start := time.Now()

	// Set default values
	if job.Timeout == 0 {
		job.Timeout = ws.config.DefaultTimeout
	}
	if job.MaxRetries == 0 {
		job.MaxRetries = ws.config.MaxRetries
	}
	if job.UserAgent == "" {
		job.UserAgent = ws.getRandomUserAgent()
	}

	// Rate limiting
	ws.rateLimiter.Wait(job.URL)

	// Scrape with retry logic
	content, err := ws.scrapeWithRetry(job)
	if err != nil {
		return &ScrapedContent{
			URL:   job.URL,
			Error: err.Error(),
		}, err
	}

	content.ResponseTime = time.Since(start)

	// Extract additional data
	ws.extractBusinessData(content)

	return content, nil
}

// scrapeWithRetry attempts to scrape with retry logic
func (ws *WebScraper) scrapeWithRetry(job *ScrapingJob) (*ScrapedContent, error) {
	var lastErr error

	for attempt := 0; attempt <= job.MaxRetries; attempt++ {
		// Get proxy for this attempt
		proxy, err := ws.proxyMgr.GetNextProxy()
		if err != nil {
			log.Printf("Failed to get proxy for attempt %d: %v", attempt+1, err)
			// Continue without proxy if none available
		}

		content, err := ws.scrape(job, proxy)
		if err == nil {
			return content, nil
		}

		lastErr = err
		log.Printf("Scraping attempt %d failed for %s: %v", attempt+1, job.URL, err)

		// Mark proxy as unhealthy if it failed
		if proxy != nil {
			ws.proxyMgr.markProxyUnhealthy(proxy)
		}

		// Wait before retry (exponential backoff)
		if attempt < job.MaxRetries {
			delay := ws.config.RetryDelay * time.Duration(attempt+1)
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("failed to scrape after %d attempts: %w", job.MaxRetries+1, lastErr)
}

// scrape performs the actual HTTP request and content extraction
func (ws *WebScraper) scrape(job *ScrapingJob, proxy *Proxy) (*ScrapedContent, error) {
	// Create HTTP client
	client := &http.Client{
		Timeout: job.Timeout,
	}

	// Configure proxy if available
	if proxy != nil {
		proxyURL, err := url.Parse(fmt.Sprintf("http://%s:%d", proxy.IP, proxy.Port))
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}

		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	// Create request
	req, err := http.NewRequest("GET", job.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", job.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// Add custom headers
	for key, value := range job.Headers {
		req.Header.Set(key, value)
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
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

	// Create scraped content
	content := &ScrapedContent{
		URL:           job.URL,
		HTML:          string(body),
		Headers:       headers,
		StatusCode:    resp.StatusCode,
		ProxyUsed:     proxy,
		ExtractedData: make(map[string]string),
	}

	// Extract title and text
	content.Title = ws.extractTitle(content.HTML)
	content.Text = ws.extractText(content.HTML)

	return content, nil
}

// extractTitle extracts the page title from HTML
func (ws *WebScraper) extractTitle(html string) string {
	// Simple title extraction (in production, use proper HTML parser)
	titleStart := strings.Index(html, "<title>")
	if titleStart == -1 {
		return ""
	}

	titleStart += 7 // length of "<title>"
	titleEnd := strings.Index(html[titleStart:], "</title>")
	if titleEnd == -1 {
		return ""
	}

	return strings.TrimSpace(html[titleStart : titleStart+titleEnd])
}

// extractText extracts plain text from HTML
func (ws *WebScraper) extractText(html string) string {
	// Simple text extraction (in production, use proper HTML parser)
	// Remove HTML tags and decode entities
	text := html

	// Remove script and style tags
	text = ws.removeTag(text, "script")
	text = ws.removeTag(text, "style")
	text = ws.removeTag(text, "noscript")

	// Remove HTML tags
	text = ws.removeHTMLTags(text)

	// Clean up whitespace
	text = ws.cleanupWhitespace(text)

	return text
}

// removeTag removes a specific HTML tag and its content
func (ws *WebScraper) removeTag(html, tagName string) string {
	startTag := fmt.Sprintf("<%s", tagName)
	endTag := fmt.Sprintf("</%s>", tagName)

	for {
		start := strings.Index(strings.ToLower(html), startTag)
		if start == -1 {
			break
		}

		// Find the end of the opening tag
		tagEnd := strings.Index(html[start:], ">")
		if tagEnd == -1 {
			break
		}
		tagEnd += start

		// Find the closing tag
		end := strings.Index(strings.ToLower(html[tagEnd:]), endTag)
		if end == -1 {
			break
		}
		end += tagEnd + len(endTag)

		// Remove the tag and its content
		html = html[:start] + html[end:]
	}

	return html
}

// removeHTMLTags removes HTML tags from text
func (ws *WebScraper) removeHTMLTags(html string) string {
	var result strings.Builder
	inTag := false

	for _, char := range html {
		if char == '<' {
			inTag = true
		} else if char == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// cleanupWhitespace cleans up excessive whitespace
func (ws *WebScraper) cleanupWhitespace(text string) string {
	// Replace multiple spaces with single space
	text = strings.Join(strings.Fields(text), " ")

	// Replace multiple newlines with single newline
	text = strings.ReplaceAll(text, "\n\n", "\n")
	text = strings.ReplaceAll(text, "\r\n\r\n", "\n")

	return strings.TrimSpace(text)
}

// extractBusinessData extracts business-related information from scraped content
func (ws *WebScraper) extractBusinessData(content *ScrapedContent) {
	// Extract business name patterns
	businessName := ws.extractBusinessName(content.Text)
	if businessName != "" {
		content.ExtractedData["business_name"] = businessName
	}

	// Extract email addresses
	emails := ws.extractEmails(content.Text)
	if len(emails) > 0 {
		content.ExtractedData["emails"] = strings.Join(emails, ", ")
	}

	// Extract phone numbers
	phones := ws.extractPhoneNumbers(content.Text)
	if len(phones) > 0 {
		content.ExtractedData["phones"] = strings.Join(phones, ", ")
	}

	// Extract addresses
	address := ws.extractAddress(content.Text)
	if address != "" {
		content.ExtractedData["address"] = address
	}
}

// extractBusinessName attempts to extract business name from text
func (ws *WebScraper) extractBusinessName(text string) string {
	// Simple business name extraction (in production, use NLP)
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 3 && len(line) < 100 {
			// Look for common business name patterns
			if strings.Contains(strings.ToLower(line), "inc") ||
				strings.Contains(strings.ToLower(line), "llc") ||
				strings.Contains(strings.ToLower(line), "corp") ||
				strings.Contains(strings.ToLower(line), "ltd") {
				return line
			}
		}
	}

	// Fallback: look for "Corporation" in the text
	if strings.Contains(text, "Corporation") {
		words := strings.Fields(text)
		for i, word := range words {
			if strings.Contains(strings.ToLower(word), "corporation") {
				// Try to get the full business name
				if i > 0 {
					return words[i-1] + " " + word
				}
				return word
			}
		}
	}

	return ""
}

// extractEmails extracts email addresses from text
func (ws *WebScraper) extractEmails(text string) []string {
	// Simple email extraction (in production, use regex)
	var emails []string
	words := strings.Fields(text)

	for _, word := range words {
		if strings.Contains(word, "@") && strings.Contains(word, ".") {
			// Basic email validation
			if len(word) > 5 && len(word) < 100 {
				emails = append(emails, word)
			}
		}
	}

	return emails
}

// extractPhoneNumbers extracts phone numbers from text
func (ws *WebScraper) extractPhoneNumbers(text string) []string {
	// Simple phone extraction (in production, use regex)
	var phones []string

	// Normalize text by replacing newlines with spaces
	normalizedText := strings.ReplaceAll(text, "\n", " ")

	// Look for phone numbers in parentheses format first
	if strings.Contains(normalizedText, "(") && strings.Contains(normalizedText, ")") {
		// Simple regex-like extraction for (555) 123-4567 format
		start := strings.Index(normalizedText, "(")
		if start != -1 {
			end := strings.Index(normalizedText[start:], ")")
			if end != -1 {
				phonePart := normalizedText[start : start+end+1]
				// Look for the rest of the phone number
				restStart := start + end + 1
				if restStart < len(normalizedText) {
					// Find the next space or end
					restEnd := restStart
					for restEnd < len(normalizedText) && normalizedText[restEnd] != ' ' && normalizedText[restEnd] != '.' && normalizedText[restEnd] != ',' {
						restEnd++
					}
					if restEnd > restStart {
						fullPhone := phonePart + normalizedText[restStart:restEnd]
						phones = append(phones, fullPhone)
					}
				}
			}
		}
	}

	// Also check individual words
	words := strings.Fields(normalizedText)
	for _, word := range words {
		// Remove common separators
		clean := strings.ReplaceAll(word, "-", "")
		clean = strings.ReplaceAll(clean, "(", "")
		clean = strings.ReplaceAll(clean, ")", "")
		clean = strings.ReplaceAll(clean, " ", "")

		// Check if it looks like a phone number
		if len(clean) >= 10 && len(clean) <= 15 {
			hasDigits := false
			for _, char := range clean {
				if char >= '0' && char <= '9' {
					hasDigits = true
					break
				}
			}
			if hasDigits {
				phones = append(phones, word)
			}
		}
	}

	return phones
}

// extractAddress attempts to extract address from text
func (ws *WebScraper) extractAddress(text string) string {
	// Simple address extraction (in production, use NLP)
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 20 && len(line) < 200 {
			// Look for address patterns
			if strings.Contains(strings.ToLower(line), "street") ||
				strings.Contains(strings.ToLower(line), "avenue") ||
				strings.Contains(strings.ToLower(line), "road") ||
				strings.Contains(strings.ToLower(line), "drive") {
				return line
			}
		}
	}

	return ""
}

// getRandomUserAgent returns a random user agent
func (ws *WebScraper) getRandomUserAgent() string {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	if len(ws.userAgents) == 0 {
		return "Mozilla/5.0 (compatible; KYB-WebScraper/1.0)"
	}

	return ws.userAgents[rand.Intn(len(ws.userAgents))]
}

// Wait waits for rate limiting
func (rl *RateLimiter) Wait(url string) {
	domain := extractDomain(url)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Remove old requests
	var validRequests []time.Time
	for _, reqTime := range rl.requests[domain] {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}
	rl.requests[domain] = validRequests

	// Check if we need to wait
	if len(validRequests) >= rl.limit {
		oldestRequest := validRequests[0]
		waitTime := rl.window - now.Sub(oldestRequest)
		if waitTime > 0 {
			time.Sleep(waitTime)
		}
	}

	// Add current request
	rl.requests[domain] = append(rl.requests[domain], now)
}

// extractDomain extracts domain from URL
func extractDomain(urlStr string) string {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}
	return parsed.Hostname()
}
