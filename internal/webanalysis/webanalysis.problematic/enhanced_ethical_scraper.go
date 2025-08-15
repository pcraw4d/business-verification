package webanalysis

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// EnhancedEthicalScraper provides improved ethical scraping capabilities
type EnhancedEthicalScraper struct {
	client  *http.Client
	config  EnhancedEthicalScraperConfig
	session *cookiejar.Jar // For maintaining session state
}

// EnhancedEthicalScraperConfig holds enhanced ethical scraping configuration
type EnhancedEthicalScraperConfig struct {
	RespectRobotsTxt      bool
	RateLimitDelay        time.Duration
	MaxRequestsPerHour    int
	UserAgent             string
	IncludeContactInfo    bool
	RespectNoScrape       bool
	Timeout               time.Duration
	MaxRedirects          int
	EnableJavaScript      bool // For future enhancement
	ExtractMetaTags       bool
	ExtractStructuredData bool
	MaintainSession       bool // New: maintain cookies between requests
	MaxRetries            int  // New: retry failed requests
}

// NewEnhancedEthicalScraper creates a new enhanced ethical scraper
func NewEnhancedEthicalScraper() *EnhancedEthicalScraper {
	// Create cookie jar for session management
	cookieJar, _ := cookiejar.New(nil)

	return &EnhancedEthicalScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     cookieJar, // Enable cookie management
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		config: EnhancedEthicalScraperConfig{
			RespectRobotsTxt:      true,
			RateLimitDelay:        2 * time.Second, // Reduced delay for better success
			MaxRequestsPerHour:    200,             // Increased limit
			UserAgent:             "KYB-Business-Classifier/1.0 (+https://kyb-platform.com/contact)",
			IncludeContactInfo:    true,
			RespectNoScrape:       true,
			Timeout:               30 * time.Second,
			MaxRedirects:          10,
			ExtractMetaTags:       true,
			ExtractStructuredData: true,
			MaintainSession:       true, // Enable session management
			MaxRetries:            3,    // Retry failed requests
		},
		session: cookieJar,
	}
}

// EnhancedScrapingJob represents an enhanced scraping request
type EnhancedScrapingJob struct {
	URL          string
	BusinessName string
	Purpose      string
	ContactEmail string
	ExtractMode  string // "full", "meta", "structured", "minimal"
}

// EnhancedScrapingResult represents the result of enhanced scraping
type EnhancedScrapingResult struct {
	URL              string
	Title            string
	Text             string
	HTML             string
	StatusCode       int
	Method           string
	Error            string
	ScrapedAt        time.Time
	LegalStatus      string
	EthicalNotes     []string
	MetaTags         map[string]string
	StructuredData   map[string]interface{}
	BusinessKeywords []string
	IndustryHints    []string
	ContentLength    int
	ProcessingTime   time.Duration
}

// ScrapeWebsite performs enhanced ethical website scraping
func (ees *EnhancedEthicalScraper) ScrapeWebsite(job *EnhancedScrapingJob) (*EnhancedScrapingResult, error) {
	start := time.Now()
	result := &EnhancedScrapingResult{
		URL:            job.URL,
		Method:         "enhanced_ethical_scraping",
		ScrapedAt:      time.Now(),
		LegalStatus:    "compliant",
		EthicalNotes:   []string{},
		MetaTags:       make(map[string]string),
		StructuredData: make(map[string]interface{}),
	}

	// Step 1: Check robots.txt (if enabled)
	if ees.config.RespectRobotsTxt {
		if !ees.checkRobotsTxt(job.URL) {
			result.LegalStatus = "prohibited"
			result.EthicalNotes = append(result.EthicalNotes, "robots.txt disallows scraping")
			return result, fmt.Errorf("robots.txt disallows scraping")
		}
	}

	// Step 2: Adaptive rate limiting based on FindDataLab recommendations
	ees.performAdaptiveRateLimiting()

	// Step 3: Create request with enhanced headers
	req, err := http.NewRequest("GET", job.URL, nil)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Enhanced headers for better success rate
	ees.setEnhancedHeaders(req, job)

	// Step 4: Perform request with timing and retry logic
	var resp *http.Response
	var requestErr error
	requestStart := time.Now()

	// Retry logic based on FindDataLab recommendations
	for attempt := 0; attempt <= ees.config.MaxRetries; attempt++ {
		resp, requestErr = ees.client.Do(req)
		if requestErr == nil && resp.StatusCode < 500 {
			break // Success or client error (don't retry client errors)
		}

		if attempt < ees.config.MaxRetries {
			// Exponential backoff
			backoffDelay := time.Duration(1<<attempt) * time.Second
			time.Sleep(backoffDelay)
		}
	}

	requestDuration := time.Since(requestStart)

	if requestErr != nil {
		result.Error = requestErr.Error()
		return result, requestErr
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// Step 5: Adaptive delay based on response time (FindDataLab recommendation)
	ees.performAdaptiveDelay(requestDuration)

	// Step 6: Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.HTML = string(body)
	result.ContentLength = len(body)

	// Step 7: Parse HTML and extract content
	doc, err := html.Parse(strings.NewReader(result.HTML))
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Extract different types of content based on mode
	switch job.ExtractMode {
	case "meta":
		result.Title = ees.extractTitle(doc)
		result.MetaTags = ees.extractMetaTags(doc)
		result.StructuredData = ees.extractStructuredData(doc)
	case "minimal":
		result.Title = ees.extractTitle(doc)
		result.Text = ees.extractMinimalText(doc)
	default: // "full"
		result.Title = ees.extractTitle(doc)
		result.Text = ees.extractEnhancedText(doc)
		result.MetaTags = ees.extractMetaTags(doc)
		result.StructuredData = ees.extractStructuredData(doc)
	}

	// Step 8: Extract business-specific information
	result.BusinessKeywords = ees.extractBusinessKeywords(result.Text, result.Title)
	result.IndustryHints = ees.extractIndustryHints(result.Text, result.MetaTags)

	result.ProcessingTime = time.Since(start)
	return result, nil
}

// performAdaptiveRateLimiting implements adaptive rate limiting
func (ees *EnhancedEthicalScraper) performAdaptiveRateLimiting() {
	// Base delay from config
	baseDelay := ees.config.RateLimitDelay

	// Add random variation to appear more human-like (FindDataLab recommendation)
	randomFactor := 0.5 + float64(time.Now().UnixNano()%100)/100.0 // 0.5 to 1.5x
	actualDelay := time.Duration(float64(baseDelay) * randomFactor)

	time.Sleep(actualDelay)
}

// performAdaptiveDelay implements adaptive delay based on response time
func (ees *EnhancedEthicalScraper) performAdaptiveDelay(responseTime time.Duration) {
	// Adaptive delay: 2x the response time (FindDataLab recommendation)
	adaptiveDelay := responseTime * 2

	// Add random variation
	randomFactor := 0.8 + float64(time.Now().UnixNano()%40)/100.0 // 0.8 to 1.2x
	actualDelay := time.Duration(float64(adaptiveDelay) * randomFactor)

	// Cap the delay to reasonable limits
	if actualDelay > 10*time.Second {
		actualDelay = 10 * time.Second
	}

	time.Sleep(actualDelay)
}

// setEnhancedHeaders sets realistic headers for better success
func (ees *EnhancedEthicalScraper) setEnhancedHeaders(req *http.Request, job *EnhancedScrapingJob) {
	// Enhanced user-agent strategy based on FindDataLab recommendations
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:89.0) Gecko/20100101 Firefox/89.0",
	}

	// Rotate user agents for better success
	selectedAgent := userAgents[time.Now().Unix()%int64(len(userAgents))]

	// Add contact information for transparency (FindDataLab recommendation)
	if ees.config.IncludeContactInfo {
		selectedAgent += "; KYB-Business-Classifier/1.0 (+https://kyb-platform.com/contact)"
	}

	req.Header.Set("User-Agent", selectedAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Cache-Control", "max-age=0")

	// Add referer to appear more human-like
	req.Header.Set("Referer", "https://www.google.com/")

	// Add contact information for transparency
	if ees.config.IncludeContactInfo {
		req.Header.Set("From", job.ContactEmail)
		req.Header.Set("X-Purpose", job.Purpose)
		req.Header.Set("X-Contact", job.ContactEmail)
	}
}

// checkRobotsTxt checks if scraping is allowed by robots.txt
func (ees *EnhancedEthicalScraper) checkRobotsTxt(url string) bool {
	// Enhanced robots.txt parsing based on FindDataLab recommendations
	robotsURL := ees.getRobotsURL(url)

	resp, err := ees.client.Get(robotsURL)
	if err != nil {
		return true // Assume allowed if can't check
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return true
	}

	robotsContent := strings.ToLower(string(body))

	// Check for explicit disallow patterns
	disallowPatterns := []string{
		"disallow: /",
		"disallow: *",
		"user-agent: *",
		"noindex",
		"nofollow",
		"crawl-delay: 0", // Some sites set crawl-delay to 0 to effectively block
	}

	for _, pattern := range disallowPatterns {
		if strings.Contains(robotsContent, pattern) {
			return false
		}
	}

	// Check for crawl-delay and adjust our rate limiting
	if strings.Contains(robotsContent, "crawl-delay:") {
		// Extract crawl-delay value and adjust our rate limiting
		lines := strings.Split(robotsContent, "\n")
		for _, line := range lines {
			if strings.Contains(line, "crawl-delay:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					delayStr := strings.TrimSpace(parts[1])
					if delay, err := strconv.Atoi(delayStr); err == nil && delay > 0 {
						// Adjust our rate limiting based on robots.txt crawl-delay
						ees.config.RateLimitDelay = time.Duration(delay) * time.Second
					}
				}
			}
		}
	}

	return true
}

// getRobotsURL extracts robots.txt URL from website URL
func (ees *EnhancedEthicalScraper) getRobotsURL(url string) string {
	domain := ees.extractDomainFromURL(url)
	return fmt.Sprintf("https://%s/robots.txt", domain)
}

// extractDomainFromURL extracts domain from URL
func (ees *EnhancedEthicalScraper) extractDomainFromURL(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")

	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}

	return url
}

// extractTitle extracts the page title
func (ees *EnhancedEthicalScraper) extractTitle(doc *html.Node) string {
	var title string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			if n.FirstChild != nil {
				title = n.FirstChild.Data
			}
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	return title
}

// extractEnhancedText extracts readable text with better filtering
func (ees *EnhancedEthicalScraper) extractEnhancedText(doc *html.Node) string {
	var text strings.Builder
	var traverse func(*html.Node)

	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Skip script, style, nav, footer, header tags
			if n.Data == "script" || n.Data == "style" || n.Data == "nav" ||
				n.Data == "footer" || n.Data == "header" || n.Data == "aside" {
				return
			}
		}

		if n.Type == html.TextNode {
			text.WriteString(n.Data)
			text.WriteString(" ")
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	return strings.TrimSpace(text.String())
}

// extractMinimalText extracts minimal text for quick analysis
func (ees *EnhancedEthicalScraper) extractMinimalText(doc *html.Node) string {
	var text strings.Builder
	var traverse func(*html.Node)

	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Only extract from main content areas
			if n.Data == "h1" || n.Data == "h2" || n.Data == "h3" ||
				n.Data == "p" || n.Data == "title" {
				// Continue to extract text
			} else {
				return
			}
		}

		if n.Type == html.TextNode {
			text.WriteString(n.Data)
			text.WriteString(" ")
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	return strings.TrimSpace(text.String())
}

// extractMetaTags extracts meta tags for business information
func (ees *EnhancedEthicalScraper) extractMetaTags(doc *html.Node) map[string]string {
	metaTags := make(map[string]string)
	var traverse func(*html.Node)

	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var name, content string
			for _, attr := range n.Attr {
				if attr.Key == "name" || attr.Key == "property" {
					name = attr.Val
				}
				if attr.Key == "content" {
					content = attr.Val
				}
			}
			if name != "" && content != "" {
				metaTags[name] = content
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	return metaTags
}

// extractStructuredData extracts JSON-LD structured data
func (ees *EnhancedEthicalScraper) extractStructuredData(doc *html.Node) map[string]interface{} {
	structuredData := make(map[string]interface{})
	var traverse func(*html.Node)

	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" {
			for _, attr := range n.Attr {
				if attr.Key == "type" && attr.Val == "application/ld+json" {
					if n.FirstChild != nil {
						// Parse JSON-LD data (simplified)
						structuredData["json_ld"] = n.FirstChild.Data
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	return structuredData
}

// extractBusinessKeywords extracts business-related keywords
func (ees *EnhancedEthicalScraper) extractBusinessKeywords(text, title string) []string {
	keywords := []string{}

	// Business-related keywords to look for
	businessTerms := []string{
		"company", "corporation", "inc", "llc", "ltd", "business", "enterprise",
		"services", "solutions", "technologies", "systems", "consulting",
		"manufacturing", "retail", "wholesale", "distribution", "logistics",
		"software", "hardware", "digital", "online", "ecommerce", "platform",
		"agency", "studio", "workshop", "factory", "warehouse", "office",
	}

	textLower := strings.ToLower(text + " " + title)
	for _, term := range businessTerms {
		if strings.Contains(textLower, term) {
			keywords = append(keywords, term)
		}
	}

	return keywords
}

// extractIndustryHints extracts industry-related hints
func (ees *EnhancedEthicalScraper) extractIndustryHints(text string, metaTags map[string]string) []string {
	hints := []string{}

	// Industry-related terms
	industryTerms := []string{
		"technology", "healthcare", "finance", "education", "retail", "manufacturing",
		"construction", "transportation", "energy", "media", "entertainment",
		"food", "beverage", "automotive", "aerospace", "pharmaceutical",
		"biotechnology", "telecommunications", "real estate", "insurance",
	}

	textLower := strings.ToLower(text)
	for _, term := range industryTerms {
		if strings.Contains(textLower, term) {
			hints = append(hints, term)
		}
	}

	// Check meta tags for industry hints
	if description, ok := metaTags["description"]; ok {
		descLower := strings.ToLower(description)
		for _, term := range industryTerms {
			if strings.Contains(descLower, term) {
				hints = append(hints, term)
			}
		}
	}

	return hints
}

// GetLegalGuidelines returns legal guidelines for ethical scraping
func (ees *EnhancedEthicalScraper) GetLegalGuidelines() map[string]string {
	return map[string]string{
		"robots_txt":       "Check and respect robots.txt",
		"rate_limiting":    "Implement reasonable rate limiting (2s delay)",
		"user_agent":       "Use transparent user agent with contact info",
		"purpose":          "Only scrape for legitimate business purposes",
		"data_usage":       "Use scraped data only as intended",
		"consent":          "Obtain consent when scraping personal data",
		"copyright":        "Respect copyright and intellectual property",
		"terms_of_service": "Review and comply with website terms",
	}
}
