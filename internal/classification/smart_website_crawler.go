package classification

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// SmartWebsiteCrawler implements intelligent website crawling with page prioritization
type SmartWebsiteCrawler struct {
	logger        *log.Logger
	client        *http.Client
	maxPages      int
	maxDepth      int
	respectRobots bool
	pageTimeout   time.Duration
}

// CrawlerCrawlResult represents the result of a smart crawl operation
// (renamed to avoid conflict with content_relevance_analyzer.go)
type CrawlerCrawlResult struct {
	BaseURL       string             `json:"base_url"`
	PagesAnalyzed []PageAnalysis     `json:"pages_analyzed"`
	TotalPages    int                `json:"total_pages"`
	RelevantPages int                `json:"relevant_pages"`
	Keywords      []string           `json:"keywords"`
	IndustryScore map[string]float64 `json:"industry_score"`
	BusinessInfo  BusinessInfo       `json:"business_info"`
	SiteStructure SiteStructure      `json:"site_structure"`
	CrawlDuration time.Duration      `json:"crawl_duration"`
	Success       bool               `json:"success"`
	Error         string             `json:"error,omitempty"`
}

// CrawlerPageAnalysis represents analysis of a single page
// (renamed to avoid conflict with content_relevance_analyzer.go)
type CrawlerPageAnalysis struct {
	URL                string                 `json:"url"`
	Title              string                 `json:"title"`
	PageType           string                 `json:"page_type"`
	RelevanceScore     float64                `json:"relevance_score"`
	ContentQuality     float64                `json:"content_quality"`
	Keywords           []string               `json:"keywords"`
	IndustryIndicators []string               `json:"industry_indicators"`
	BusinessInfo       BusinessInfo           `json:"business_info"`
	MetaTags           map[string]string      `json:"meta_tags"`
	StructuredData     map[string]interface{} `json:"structured_data"`
	ResponseTime       time.Duration          `json:"response_time"`
	StatusCode         int                    `json:"status_code"`
	ContentLength      int                    `json:"content_length"`
	LastModified       time.Time              `json:"last_modified"`
	Priority           int                    `json:"priority"`
}

// CrawlerBusinessInfo represents extracted business information
// (renamed to avoid conflict with content_relevance_analyzer.go)
type CrawlerBusinessInfo struct {
	BusinessName  string      `json:"business_name"`
	Description   string      `json:"description"`
	Services      []string    `json:"services"`
	Products      []string    `json:"products"`
	ContactInfo   ContactInfo `json:"contact_info"`
	BusinessHours string      `json:"business_hours"`
	Location      string      `json:"location"`
	Industry      string      `json:"industry"`
	BusinessType  string      `json:"business_type"`
}

// CrawlerContactInfo represents contact information
// (renamed to avoid conflict with content_relevance_analyzer.go)
type CrawlerContactInfo struct {
	Phone   string            `json:"phone"`
	Email   string            `json:"email"`
	Address string            `json:"address"`
	Website string            `json:"website"`
	Social  map[string]string `json:"social"`
}

// CrawlerSiteStructure represents the discovered site structure
// (renamed to avoid conflict with content_relevance_analyzer.go)
type CrawlerSiteStructure struct {
	Homepage        string   `json:"homepage"`
	AboutPages      []string `json:"about_pages"`
	ServicePages    []string `json:"service_pages"`
	ProductPages    []string `json:"product_pages"`
	ContactPages    []string `json:"contact_pages"`
	BlogPages       []string `json:"blog_pages"`
	EcommercePages  []string `json:"ecommerce_pages"`
	OtherPages      []string `json:"other_pages"`
	TotalDiscovered int      `json:"total_discovered"`
}

// PageType represents different types of pages
type PageType string

const (
	PageTypeHomepage  PageType = "homepage"
	PageTypeAbout     PageType = "about"
	PageTypeServices  PageType = "services"
	PageTypeProducts  PageType = "products"
	PageTypeContact   PageType = "contact"
	PageTypeBlog      PageType = "blog"
	PageTypeEcommerce PageType = "ecommerce"
	PageTypeSupport   PageType = "support"
	PageTypeCareers   PageType = "careers"
	PageTypeNews      PageType = "news"
	PageTypeOther     PageType = "other"
)

// NewSmartWebsiteCrawler creates a new smart website crawler
func NewSmartWebsiteCrawler(logger *log.Logger) *SmartWebsiteCrawler {
	// Create custom dialer that forces IPv4 DNS resolution using Google DNS
	// This addresses DNS resolution failures in containerized environments like Railway
	baseDialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	// Create custom DNS resolver with multiple fallback servers
	// DNS servers in order of preference: Google DNS, Cloudflare, Google DNS secondary
	dnsServers := []string{"8.8.8.8:53", "1.1.1.1:53", "8.8.4.4:53"}
	dnsResolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			// Try each DNS server with retry logic
			var lastErr error
			for _, server := range dnsServers {
				d := net.Dialer{
					Timeout: 5 * time.Second,
				}
				conn, err := d.DialContext(ctx, "udp4", server)
				if err == nil {
					return conn, nil
				}
				lastErr = err
				// Log DNS server failure (if logger is available)
				if logger != nil {
					logger.Printf("⚠️ [DNS] Failed to connect to DNS server %s: %v", server, err)
				}
			}
			return nil, fmt.Errorf("all DNS servers failed, last error: %w", lastErr)
		},
	}

	// Custom DialContext that forces IPv4 resolution
	customDialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		// Force IPv4 by using "tcp4" instead of "tcp"
		if network == "tcp" {
			network = "tcp4"
		}

		// Parse address to get host and port
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("failed to split host:port: %w", err)
		}

		// Resolve host using custom DNS resolver with retry logic
		var ips []net.IPAddr
		var dnsErr error
		maxRetries := 3
		for attempt := 1; attempt <= maxRetries; attempt++ {
			ips, dnsErr = dnsResolver.LookupIPAddr(ctx, host)
			if dnsErr == nil {
				break
			}
			// Exponential backoff: 1s, 2s, 4s
			if attempt < maxRetries {
				backoff := time.Duration(attempt) * time.Second
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(backoff):
					// Retry after backoff
				}
			}
		}
		if dnsErr != nil {
			return nil, fmt.Errorf("DNS lookup failed for %s after %d attempts: %w", host, maxRetries, dnsErr)
		}

		// Use first IPv4 address
		var ip net.IP
		for _, ipAddr := range ips {
			if ipAddr.IP.To4() != nil {
				ip = ipAddr.IP
				break
			}
		}

		if ip == nil {
			return nil, fmt.Errorf("no IPv4 address found for %s", host)
		}

		// Dial using resolved IPv4 address
		return baseDialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
	}

	return &SmartWebsiteCrawler{
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				DialContext:          customDialContext,
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				MaxIdleConnsPerHost: 2,
			},
		},
		maxPages:      20, // Maximum pages to crawl
		maxDepth:      3,  // Maximum crawl depth
		respectRobots: true,
		pageTimeout:   15 * time.Second,
	}
}

// CrawlWebsite performs intelligent website crawling with page prioritization
func (c *SmartWebsiteCrawler) CrawlWebsite(ctx context.Context, websiteURL string) (*CrawlResult, error) {
	startTime := time.Now()
	c.logger.Printf("🕷️ [SmartCrawler] Starting smart crawl for: %s", websiteURL)

	result := &CrawlResult{
		BaseURL:       websiteURL,
		PagesAnalyzed: []PageAnalysis{},
		Keywords:      []string{},
		IndustryScore: make(map[string]float64),
		BusinessInfo:  BusinessInfo{},
		SiteStructure: SiteStructure{},
		Success:       false,
	}

	// Validate and normalize URL
	baseURL, err := c.normalizeURL(websiteURL)
	if err != nil {
		result.Error = fmt.Sprintf("URL validation failed: %v", err)
		return result, err
	}

	// Check robots.txt if enabled
	if c.respectRobots {
		if blocked, err := c.checkRobotsTxt(ctx, baseURL); err == nil && blocked {
			result.Error = "Website blocked by robots.txt"
			return result, fmt.Errorf("website blocked by robots.txt")
		}
	}

	// Discover site structure and prioritize pages
	discoveredPages, err := c.discoverSiteStructure(ctx, baseURL)
	if err != nil {
		c.logger.Printf("⚠️ [SmartCrawler] Site structure discovery failed: %v", err)
		// Fallback to homepage only
		discoveredPages = []string{baseURL}
	}

	// Prioritize pages based on relevance
	prioritizedPages := c.prioritizePages(discoveredPages, baseURL)
	c.logger.Printf("📊 [SmartCrawler] Discovered %d pages, prioritizing %d", len(discoveredPages), len(prioritizedPages))

	// Analyze prioritized pages
	pageAnalyses := c.analyzePages(ctx, prioritizedPages)
	result.PagesAnalyzed = pageAnalyses
	result.TotalPages = len(pageAnalyses)

	// Aggregate results
	c.aggregateResults(result, pageAnalyses)

	result.CrawlDuration = time.Since(startTime)
	result.Success = true

	c.logger.Printf("✅ [SmartCrawler] Smart crawl completed in %v - %d pages analyzed", result.CrawlDuration, result.TotalPages)
	return result, nil
}

// discoverSiteStructure discovers the site structure using multiple methods
func (c *SmartWebsiteCrawler) discoverSiteStructure(ctx context.Context, baseURL string) ([]string, error) {
	var discoveredPages []string
	seen := make(map[string]bool)

	// Method 1: Parse sitemap.xml
	sitemapPages, err := c.parseSitemap(ctx, baseURL)
	if err == nil {
		for _, page := range sitemapPages {
			if !seen[page] {
				discoveredPages = append(discoveredPages, page)
				seen[page] = true
			}
		}
		c.logger.Printf("🗺️ [SmartCrawler] Found %d pages in sitemap", len(sitemapPages))
	}

	// Method 2: Crawl homepage for internal links
	homepageLinks, err := c.extractInternalLinks(ctx, baseURL)
	if err == nil {
		for _, link := range homepageLinks {
			if !seen[link] {
				discoveredPages = append(discoveredPages, link)
				seen[link] = true
			}
		}
		c.logger.Printf("🔗 [SmartCrawler] Found %d internal links from homepage", len(homepageLinks))
	}

	// Method 3: Common page patterns
	commonPages := c.generateCommonPagePatterns(baseURL)
	for _, page := range commonPages {
		if !seen[page] {
			discoveredPages = append(discoveredPages, page)
			seen[page] = true
		}
	}

	return discoveredPages, nil
}

// parseSitemap parses sitemap.xml to discover pages
func (c *SmartWebsiteCrawler) parseSitemap(ctx context.Context, baseURL string) ([]string, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	sitemapURL := fmt.Sprintf("%s://%s/sitemap.xml", parsedURL.Scheme, parsedURL.Host)

	req, err := http.NewRequestWithContext(ctx, "GET", sitemapURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("sitemap not found or inaccessible")
	}

	// Parse XML sitemap (simplified implementation)
	// In a real implementation, you'd use an XML parser
	var pages []string

	// For now, return common page patterns
	// TODO: Implement proper XML parsing
	return pages, nil
}

// extractInternalLinks extracts internal links from a page
func (c *SmartWebsiteCrawler) extractInternalLinks(ctx context.Context, pageURL string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("page not accessible")
	}

	// Parse HTML to extract links
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var links []string
	baseURL, _ := url.Parse(pageURL)

	var extractLinks func(*html.Node)
	extractLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					linkURL, err := url.Parse(attr.Val)
					if err != nil {
						continue
					}

					// Convert relative URLs to absolute
					absoluteURL := baseURL.ResolveReference(linkURL)

					// Only include internal links
					if absoluteURL.Host == baseURL.Host {
						links = append(links, absoluteURL.String())
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractLinks(c)
		}
	}

	extractLinks(doc)
	return links, nil
}

// generateCommonPagePatterns generates common page URL patterns
func (c *SmartWebsiteCrawler) generateCommonPagePatterns(baseURL string) []string {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return []string{}
	}

	base := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	commonPatterns := []string{
		"/about",
		"/about-us",
		"/aboutus",
		"/company",
		"/services",
		"/products",
		"/shop",
		"/store",
		"/contact",
		"/contact-us",
		"/contactus",
		"/mission",
		"/vision",
		"/team",
		"/careers",
		"/jobs",
		"/blog",
		"/news",
		"/support",
		"/help",
		"/faq",
		"/privacy",
		"/terms",
		"/legal",
	}

	var pages []string
	for _, pattern := range commonPatterns {
		pages = append(pages, base+pattern)
	}

	return pages
}

// prioritizePages prioritizes pages based on relevance for business analysis
func (c *SmartWebsiteCrawler) prioritizePages(pages []string, baseURL string) []string {
	type pagePriority struct {
		URL      string
		Priority int
		PageType PageType
	}

	var prioritized []pagePriority

	for _, page := range pages {
		priority := c.calculatePagePriority(page, baseURL)
		pageType := c.detectPageType(page)

		prioritized = append(prioritized, pagePriority{
			URL:      page,
			Priority: priority,
			PageType: pageType,
		})
	}

	// Sort by priority (highest first)
	sort.Slice(prioritized, func(i, j int) bool {
		return prioritized[i].Priority > prioritized[j].Priority
	})

	// Limit to maxPages
	if len(prioritized) > c.maxPages {
		prioritized = prioritized[:c.maxPages]
	}

	var result []string
	for _, p := range prioritized {
		result = append(result, p.URL)
	}

	return result
}

// calculatePagePriority calculates priority score for a page
func (c *SmartWebsiteCrawler) calculatePagePriority(pageURL, baseURL string) int {
	priority := 0
	urlLower := strings.ToLower(pageURL)

	// Homepage gets highest priority
	if pageURL == baseURL || pageURL == baseURL+"/" {
		return 100
	}

	// Highest priority pages (90-100): about, products, services, sale, sales
	highestPriorityPatterns := []string{
		"/about", "/about-us", "/aboutus", "/company", "/mission", "/vision",
		"/products", "/product", "/services", "/service",
		"/sale", "/sales", "/shop", "/store",
	}

	for _, pattern := range highestPriorityPatterns {
		if strings.Contains(urlLower, pattern) {
			priority += 95 // Highest priority weight
			break
		}
	}

	// High priority pages (70-80): contact, team, careers, locations
	if priority == 0 {
		highPriorityPatterns := []string{
			"/contact", "/contact-us", "/contactus",
			"/team", "/careers", "/jobs", "/locations", "/location",
		}

		for _, pattern := range highPriorityPatterns {
			if strings.Contains(urlLower, pattern) {
				priority += 75
				break
			}
		}
	}

	// Medium priority pages (50-60): blog, news, case-studies, portfolio
	if priority == 0 {
		mediumPriorityPatterns := []string{
			"/blog", "/news", "/case-studies", "/case_studies", "/portfolio",
		}

		for _, pattern := range mediumPriorityPatterns {
			if strings.Contains(urlLower, pattern) {
				priority += 55
				break
			}
		}
	}

	// Low priority pages (30-40): support, help, faq, privacy, terms
	if priority == 0 {
		lowPriorityPatterns := []string{
			"/support", "/help", "/faq", "/privacy", "/terms", "/legal",
		}

		for _, pattern := range lowPriorityPatterns {
			if strings.Contains(urlLower, pattern) {
				priority += 35
				break
			}
		}
	}

	// Default priority for other pages
	if priority == 0 {
		priority = 20
	}

	return priority
}

// detectPageType detects the type of page based on URL
func (c *SmartWebsiteCrawler) detectPageType(pageURL string) PageType {
	urlLower := strings.ToLower(pageURL)

	if strings.Contains(urlLower, "/about") || strings.Contains(urlLower, "/company") ||
		strings.Contains(urlLower, "/mission") || strings.Contains(urlLower, "/vision") {
		return PageTypeAbout
	}

	if strings.Contains(urlLower, "/services") {
		return PageTypeServices
	}

	if strings.Contains(urlLower, "/products") || strings.Contains(urlLower, "/shop") ||
		strings.Contains(urlLower, "/store") {
		return PageTypeProducts
	}

	if strings.Contains(urlLower, "/contact") {
		return PageTypeContact
	}

	if strings.Contains(urlLower, "/blog") || strings.Contains(urlLower, "/news") {
		return PageTypeBlog
	}

	if strings.Contains(urlLower, "/shop") || strings.Contains(urlLower, "/store") ||
		strings.Contains(urlLower, "/cart") || strings.Contains(urlLower, "/checkout") {
		return PageTypeEcommerce
	}

	return PageTypeOther
}

// analyzePages analyzes multiple pages concurrently
func (c *SmartWebsiteCrawler) analyzePages(ctx context.Context, pages []string) []PageAnalysis {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var analyses []PageAnalysis

	// Limit concurrent requests
	semaphore := make(chan struct{}, 5) // Max 5 concurrent requests

	for _, page := range pages {
		wg.Add(1)
		go func(pageURL string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			analysis := c.analyzePage(ctx, pageURL)

			mu.Lock()
			analyses = append(analyses, analysis)
			mu.Unlock()
		}(page)
	}

	wg.Wait()
	return analyses
}

// analyzePage analyzes a single page
func (c *SmartWebsiteCrawler) analyzePage(ctx context.Context, pageURL string) PageAnalysis {
	startTime := time.Now()

	analysis := PageAnalysis{
		URL:                pageURL,
		PageType:           string(c.detectPageType(pageURL)),
		RelevanceScore:     0.0,
		ContentQuality:     0.0,
		Keywords:           []string{},
		IndustryIndicators: []string{},
		BusinessInfo:       BusinessInfo{},
		MetaTags:           make(map[string]string),
		StructuredData:     make(map[string]interface{}),
		StatusCode:         0,
		ContentLength:      0,
		Priority:           c.calculatePagePriority(pageURL, ""),
	}

	// Create request with timeout
	reqCtx, cancel := context.WithTimeout(ctx, c.pageTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", pageURL, nil)
	if err != nil {
		analysis.RelevanceScore = 0.0
		return analysis
	}

	// Set headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	// Retry logic for HTTP requests (up to 3 attempts with exponential backoff)
	var resp *http.Response
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err = c.client.Do(req)
		if err == nil {
			break
		}
		
		// Distinguish between DNS errors, network errors, and HTTP errors
		if dnsErr, ok := err.(*net.DNSError); ok {
			c.logger.Printf("⚠️ [PageAnalysis] DNS error for %s (attempt %d/%d): %v", pageURL, attempt, maxRetries, dnsErr)
		} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			c.logger.Printf("⚠️ [PageAnalysis] Timeout error for %s (attempt %d/%d): %v", pageURL, attempt, maxRetries, netErr)
		} else {
			c.logger.Printf("⚠️ [PageAnalysis] Network error for %s (attempt %d/%d): %v", pageURL, attempt, maxRetries, err)
		}
		
		// Exponential backoff: 1s, 2s, 4s
		if attempt < maxRetries {
			backoff := time.Duration(attempt) * time.Second
			select {
			case <-reqCtx.Done():
				analysis.RelevanceScore = 0.0
				return analysis
			case <-time.After(backoff):
				// Retry after backoff
			}
		}
	}
	
	if err != nil {
		c.logger.Printf("❌ [PageAnalysis] Failed to fetch %s after %d attempts: %v", pageURL, maxRetries, err)
		analysis.RelevanceScore = 0.0
		return analysis
	}
	defer resp.Body.Close()

	analysis.StatusCode = resp.StatusCode
	analysis.ResponseTime = time.Since(startTime)

	if resp.StatusCode != 200 {
		analysis.RelevanceScore = 0.0
		return analysis
	}

	// Read content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		analysis.RelevanceScore = 0.0
		return analysis
	}

	analysis.ContentLength = len(body)
	content := string(body)

	// Extract title
	analysis.Title = c.extractTitle(content)

	// Extract meta tags
	analysis.MetaTags = c.extractMetaTags(content)

	// Extract structured data
	analysis.StructuredData = c.extractStructuredData(content)

	// Extract business information
	analysis.BusinessInfo = c.extractBusinessInfo(content, analysis.PageType)

	// Extract keywords
	analysis.Keywords = c.extractPageKeywords(content, analysis.PageType)

	// Extract industry indicators
	analysis.IndustryIndicators = c.extractIndustryIndicators(content)

	// Calculate relevance score
	analysis.RelevanceScore = c.calculateRelevanceScore(analysis)

	// Calculate content quality
	analysis.ContentQuality = c.calculateContentQuality(analysis)

	return analysis
}

// Additional helper methods would be implemented here...
// (extractTitle, extractMetaTags, extractStructuredData, etc.)

// normalizeURL normalizes and validates a URL
func (c *SmartWebsiteCrawler) normalizeURL(websiteURL string) (string, error) {
	if !strings.HasPrefix(websiteURL, "http://") && !strings.HasPrefix(websiteURL, "https://") {
		websiteURL = "https://" + websiteURL
	}

	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return "", err
	}

	if parsedURL.Host == "" {
		return "", fmt.Errorf("invalid URL: missing host")
	}

	return parsedURL.String(), nil
}

// checkRobotsTxt checks if crawling is allowed by robots.txt
func (c *SmartWebsiteCrawler) checkRobotsTxt(ctx context.Context, baseURL string) (bool, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return false, err
	}

	robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)

	req, err := http.NewRequestWithContext(ctx, "GET", robotsURL, nil)
	if err != nil {
		return false, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, nil // No robots.txt, allow crawling
	}

	// Parse robots.txt (simplified)
	// In a real implementation, you'd use a proper robots.txt parser
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	content := strings.ToLower(string(body))

	// Check for disallow rules
	if strings.Contains(content, "disallow: /") {
		return true, nil // Blocked
	}

	return false, nil // Allowed
}

// Placeholder methods for content extraction
func (c *SmartWebsiteCrawler) extractTitle(content string) string {
	// Extract title from HTML
	titleRegex := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
	matches := titleRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func (c *SmartWebsiteCrawler) extractMetaTags(content string) map[string]string {
	metaTags := make(map[string]string)
	
	// Extract meta description
	metaDescRegex := regexp.MustCompile(`(?i)<meta[^>]*name=["']description["'][^>]*content=["']([^"']+)["']`)
	metaDescMatches := metaDescRegex.FindStringSubmatch(content)
	if len(metaDescMatches) > 1 {
		metaTags["description"] = strings.TrimSpace(metaDescMatches[1])
	}
	
	// Extract meta keywords
	metaKeywordsRegex := regexp.MustCompile(`(?i)<meta[^>]*name=["']keywords["'][^>]*content=["']([^"']+)["']`)
	metaKeywordsMatches := metaKeywordsRegex.FindStringSubmatch(content)
	if len(metaKeywordsMatches) > 1 {
		metaTags["keywords"] = strings.TrimSpace(metaKeywordsMatches[1])
	}
	
	// Extract Open Graph title
	ogTitleRegex := regexp.MustCompile(`(?i)<meta[^>]*property=["']og:title["'][^>]*content=["']([^"']+)["']`)
	ogTitleMatches := ogTitleRegex.FindStringSubmatch(content)
	if len(ogTitleMatches) > 1 {
		metaTags["og:title"] = strings.TrimSpace(ogTitleMatches[1])
	}
	
	// Extract Open Graph description
	ogDescRegex := regexp.MustCompile(`(?i)<meta[^>]*property=["']og:description["'][^>]*content=["']([^"']+)["']`)
	ogDescMatches := ogDescRegex.FindStringSubmatch(content)
	if len(ogDescMatches) > 1 {
		metaTags["og:description"] = strings.TrimSpace(ogDescMatches[1])
	}
	
	return metaTags
}

func (c *SmartWebsiteCrawler) extractStructuredData(content string) map[string]interface{} {
	structuredData := make(map[string]interface{})
	
	// Extract JSON-LD structured data
	jsonLdRegex := regexp.MustCompile(`(?i)<script[^>]*type=["']application/ld\+json["'][^>]*>([^<]+)</script>`)
	jsonLdMatches := jsonLdRegex.FindAllStringSubmatch(content, -1)
	
	for i, match := range jsonLdMatches {
		if len(match) > 1 {
			jsonContent := strings.TrimSpace(match[1])
			// Store JSON-LD content (parsing would require JSON library)
			structuredData[fmt.Sprintf("json-ld-%d", i)] = jsonContent
		}
	}
	
	// Extract microdata (basic extraction)
	// Look for itemscope attributes
	itemScopeRegex := regexp.MustCompile(`(?i)itemscope[^>]*itemtype=["']([^"']+)["']`)
	itemScopeMatches := itemScopeRegex.FindAllStringSubmatch(content, -1)
	for i, match := range itemScopeMatches {
		if len(match) > 1 {
			structuredData[fmt.Sprintf("microdata-type-%d", i)] = strings.TrimSpace(match[1])
		}
	}
	
	// Extract itemprop values
	itemPropRegex := regexp.MustCompile(`(?i)itemprop=["']([^"']+)["'][^>]*content=["']([^"']+)["']`)
	itemPropMatches := itemPropRegex.FindAllStringSubmatch(content, -1)
	for _, match := range itemPropMatches {
		if len(match) >= 3 {
			prop := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])
			structuredData[prop] = value
		}
	}
	
	return structuredData
}

func (c *SmartWebsiteCrawler) extractBusinessInfo(content string, pageType string) BusinessInfo {
	businessInfo := BusinessInfo{}
	// Implementation for extracting business information
	return businessInfo
}

// extractTextFromHTML extracts clean text content from HTML
func (c *SmartWebsiteCrawler) extractTextFromHTML(htmlContent string) string {
	// Remove script and style tags completely
	htmlContent = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`).ReplaceAllString(htmlContent, "")
	htmlContent = regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`).ReplaceAllString(htmlContent, "")
	
	// Remove HTML tags
	htmlContent = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(htmlContent, " ")
	
	// Decode HTML entities (basic)
	htmlContent = strings.ReplaceAll(htmlContent, "&nbsp;", " ")
	htmlContent = strings.ReplaceAll(htmlContent, "&amp;", "&")
	htmlContent = strings.ReplaceAll(htmlContent, "&lt;", "<")
	htmlContent = strings.ReplaceAll(htmlContent, "&gt;", ">")
	htmlContent = strings.ReplaceAll(htmlContent, "&quot;", "\"")
	htmlContent = strings.ReplaceAll(htmlContent, "&#39;", "'")
	
	// Clean up whitespace
	htmlContent = regexp.MustCompile(`\s+`).ReplaceAllString(htmlContent, " ")
	
	return strings.TrimSpace(htmlContent)
}

// extractStructuredKeywords extracts keywords from structured HTML elements
// (title, meta description, headings h1-h6)
func (c *SmartWebsiteCrawler) extractStructuredKeywords(content string) []string {
	var keywords []string
	seen := make(map[string]bool)
	
	// Extract from title (highest weight)
	titleRegex := regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)
	titleMatches := titleRegex.FindStringSubmatch(content)
	if len(titleMatches) > 1 {
		titleText := strings.TrimSpace(titleMatches[1])
		titleWords := c.extractWordsFromText(titleText)
		for _, word := range titleWords {
			wordLower := strings.ToLower(word)
			if !seen[wordLower] && len(wordLower) > 2 {
				seen[wordLower] = true
				keywords = append(keywords, wordLower)
			}
		}
	}
	
	// Extract from meta description
	metaDescRegex := regexp.MustCompile(`(?i)<meta[^>]*name=["']description["'][^>]*content=["']([^"']+)["']`)
	metaDescMatches := metaDescRegex.FindStringSubmatch(content)
	if len(metaDescMatches) > 1 {
		metaText := strings.TrimSpace(metaDescMatches[1])
		metaWords := c.extractWordsFromText(metaText)
		for _, word := range metaWords {
			wordLower := strings.ToLower(word)
			if !seen[wordLower] && len(wordLower) > 2 {
				seen[wordLower] = true
				keywords = append(keywords, wordLower)
			}
		}
	}
	
	// Extract from headings (h1-h6) - weighted by importance
	headingRegex := regexp.MustCompile(`(?i)<h([1-6])[^>]*>([^<]+)</h[1-6]>`)
	headingMatches := headingRegex.FindAllStringSubmatch(content, -1)
	for _, match := range headingMatches {
		if len(match) >= 3 {
			headingText := strings.TrimSpace(match[2])
			headingWords := c.extractWordsFromText(headingText)
			for _, word := range headingWords {
				wordLower := strings.ToLower(word)
				if !seen[wordLower] && len(wordLower) > 2 {
					seen[wordLower] = true
					keywords = append(keywords, wordLower)
				}
			}
		}
	}
	
	return keywords
}

// extractWordsFromText extracts meaningful words from text (filters stop words)
func (c *SmartWebsiteCrawler) extractWordsFromText(text string) []string {
	// Common stop words to filter
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "as": true, "is": true, "are": true,
		"was": true, "were": true, "be": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "do": true, "does": true, "did": true,
		"will": true, "would": true, "should": true, "could": true, "may": true, "might": true,
		"this": true, "that": true, "these": true, "those": true, "it": true, "its": true,
		"we": true, "you": true, "they": true, "he": true, "she": true, "him": true, "her": true,
		"our": true, "your": true, "their": true, "my": true, "me": true, "us": true,
		"about": true, "into": true, "through": true, "during": true, "before": true,
		"after": true, "above": true, "below": true, "up": true, "down": true, "out": true,
		"off": true, "over": true, "under": true, "again": true, "further": true,
		"then": true, "once": true, "here": true, "there": true, "when": true, "where": true,
		"why": true, "how": true, "all": true, "each": true, "few": true, "more": true,
		"most": true, "other": true, "some": true, "such": true, "no": true, "nor": true,
		"not": true, "only": true, "own": true, "same": true, "so": true, "than": true,
		"too": true, "very": true, "can": true, "just": true, "don": true,
	}
	
	// Split text into words
	words := regexp.MustCompile(`\b[a-zA-Z]{3,}\b`).FindAllString(text, -1)
	var filteredWords []string
	
	for _, word := range words {
		wordLower := strings.ToLower(word)
		if !stopWords[wordLower] && len(wordLower) >= 3 {
			filteredWords = append(filteredWords, wordLower)
		}
	}
	
	return filteredWords
}

// extractBusinessKeywordsFromText extracts business-relevant keywords from text using patterns
func (c *SmartWebsiteCrawler) extractBusinessKeywordsFromText(textContent string) []string {
	var keywords []string
	seen := make(map[string]bool)
	
	// Convert to lowercase for processing
	text := strings.ToLower(textContent)
	
	// Business-relevant keyword patterns (expanded from plan)
	businessPatterns := []string{
		// Food & Beverage (expanded) - single words first, then phrases
		`\b(wine|wines|winery|vineyard|vintner|sommelier|tasting|cellar|bottle|vintage|grape|grapes|grapevine|oenology|alcohol|spirits|liquor|beer|brewery|distillery|beverage|beverages|restaurant|cafe|coffee|food|dining|kitchen|catering|bakery|bar|pub)\b`,
		`\b(wine shop|wine store|wine bar|wine merchant|wine retailer)\b`,
		
		// Retail (expanded) - single words first, then phrases
		`\b(retail|retailer|storefront|merchandise|inventory|POS|checkout|showroom|boutique|outlet|marketplace|vendor|seller|selling|commerce|store|shop)\b`,
		`\b(retail store|retail shop|brick and mortar|brick-and-mortar|physical store|point of sale|cash register|sales floor)\b`,
		
		// E-commerce (new) - single words first, then phrases
		`\b(ecommerce|e-commerce)\b`,
		`\b(online store|online shop|digital storefront|web store|internet retailer|online marketplace|digital commerce|online sales|web sales|internet sales|online retail)\b`,
		
		// Technology
		`\b(technology|software|tech|app|digital|web|mobile|cloud|ai|ml|data|cyber|security|programming|development|IT|computer|internet|online|platform|api|database|saas)\b`,
		
		// Healthcare
		`\b(healthcare|medical|clinic|hospital|doctor|dentist|therapy|wellness|pharmacy|medicine|patient|treatment|health|care|nurse|physician)\b`,
		
		// Legal
		`\b(legal|law|attorney|lawyer|court|litigation|patent|trademark|copyright|legal services|advocacy|justice|legal advice|law firm)\b`,
		
		// Finance
		`\b(finance|banking|investment|insurance|accounting|tax|financial|credit|loan|money|capital|funding|payment|transaction|wealth)\b`,
		
		// Real Estate - single words first, then phrases
		`\b(property|construction|building|architecture|design|interior|home|house|apartment|rental|mortgage)\b`,
		`\b(real estate|property management)\b`,
		
		// Education
		`\b(education|school|university|training|learning|course|academy|institute|student|teacher|teaching|academic|degree|certification)\b`,
		
		// Consulting
		`\b(consulting|advisory|strategy|management|business|corporate|professional|services|expert|specialist|consultant)\b`,
		
		// Manufacturing
		`\b(manufacturing|production|factory|industrial|automotive|machinery|equipment|assembly)\b`,
		
		// Transportation
		`\b(transportation|logistics|shipping|delivery|freight|warehouse|supply chain|trucking)\b`,
		
		// Entertainment
		`\b(entertainment|media|marketing|advertising|design|creative|art|music|film)\b`,
		
		// Energy
		`\b(energy|utilities|renewable|solar|wind|oil|gas|power|electricity)\b`,
		
		// Agriculture
		`\b(agriculture|farming|food production|crop|livestock|organic|sustainable)\b`,
		
		// Travel
		`\b(travel|tourism|hospitality|hotel|accommodation|vacation|booking|trip)\b`,
	}
	
	// Extract keywords using patterns
	for _, pattern := range businessPatterns {
		matches := regexp.MustCompile(pattern).FindAllString(text, -1)
		for _, match := range matches {
			// Normalize match (remove extra spaces)
			match = strings.TrimSpace(strings.ToLower(match))
			if !seen[match] && len(match) >= 3 {
				seen[match] = true
				keywords = append(keywords, match)
			}
		}
	}
	
	return keywords
}

// extractPhrases extracts multi-word phrases from text
func (c *SmartWebsiteCrawler) extractPhrases(textContent string, minWords, maxWords int) []string {
	var phrases []string
	seen := make(map[string]bool)
	
	// Extract words from text
	words := c.extractWordsFromText(textContent)
	
	// Generate phrases of different lengths
	for i := 0; i < len(words); i++ {
		for length := minWords; length <= maxWords && i+length <= len(words); length++ {
			phrase := strings.Join(words[i:i+length], " ")
			phraseLower := strings.ToLower(phrase)
			
			// Filter out phrases that are too short or too long
			if len(phraseLower) >= 4 && len(phraseLower) <= 50 && !seen[phraseLower] {
				seen[phraseLower] = true
				phrases = append(phrases, phraseLower)
			}
		}
	}
	
	return phrases
}

// keywordScore represents a keyword with its relevance score
type keywordScore struct {
	keyword string
	score   float64
}

// combineAndRankKeywords combines keywords from different sources and ranks them by relevance
func (c *SmartWebsiteCrawler) combineAndRankKeywords(structuredKeywords, bodyKeywords, phrases []string, pageType string) []keywordScore {
	keywordScores := make(map[string]float64)
	
	// Weight keywords by source and position
	// Structured keywords (title, meta, headings) get highest weight
	for _, kw := range structuredKeywords {
		// Title keywords get 1.0, meta gets 0.9, headings get 0.8
		// For simplicity, we'll use 0.9 for all structured keywords
		keywordScores[kw] += 0.9
	}
	
	// Body keywords get medium weight
	for _, kw := range bodyKeywords {
		keywordScores[kw] += 0.6
	}
	
	// Phrases get weight based on length (longer phrases = more specific = higher weight)
	for _, phrase := range phrases {
		wordCount := len(strings.Fields(phrase))
		weight := 0.5 + float64(wordCount-2)*0.1 // 2-word: 0.5, 3-word: 0.6
		if weight > 0.8 {
			weight = 0.8 // Cap at 0.8
		}
		keywordScores[phrase] += weight
	}
	
	// Boost keywords based on page type relevance
	pageTypeBoost := c.getPageTypeBoost(pageType)
	for kw := range keywordScores {
		keywordScores[kw] *= pageTypeBoost
	}
	
	// Convert to slice and sort
	var scoredKeywords []keywordScore
	for kw, score := range keywordScores {
		scoredKeywords = append(scoredKeywords, keywordScore{keyword: kw, score: score})
	}
	
	// Sort by score descending
	sort.Slice(scoredKeywords, func(i, j int) bool {
		return scoredKeywords[i].score > scoredKeywords[j].score
	})
	
	return scoredKeywords
}

// getPageTypeBoost returns a boost multiplier based on page type
func (c *SmartWebsiteCrawler) getPageTypeBoost(pageType string) float64 {
	switch pageType {
	case "about", "services", "products":
		return 1.0 // Highest relevance
	case "homepage":
		return 0.95
	case "contact":
		return 0.85
	case "blog", "news":
		return 0.7
	default:
		return 0.8
	}
}

// limitToTopKeywords returns the top N keywords from scored keywords
func (c *SmartWebsiteCrawler) limitToTopKeywords(scoredKeywords []keywordScore, limit int) []string {
	if len(scoredKeywords) > limit {
		scoredKeywords = scoredKeywords[:limit]
	}
	
	keywords := make([]string, len(scoredKeywords))
	for i, kw := range scoredKeywords {
		keywords[i] = kw.keyword
	}
	
	return keywords
}

// extractPageKeywords extracts keywords from HTML page content
// Returns top 30 keywords sorted by relevance
func (c *SmartWebsiteCrawler) extractPageKeywords(content string, pageType string) []string {
	// 1. Extract clean text from HTML
	textContent := c.extractTextFromHTML(content)
	
	// 2. Extract from structured elements (title, meta, headings)
	structuredKeywords := c.extractStructuredKeywords(content)
	
	// 3. Extract from body text using business patterns
	bodyKeywords := c.extractBusinessKeywordsFromText(textContent)
	
	// 4. Extract phrases (2-word, 3-word)
	phrases := c.extractPhrases(textContent, 2, 3)
	
	// 5. Combine, deduplicate, and rank
	allKeywords := c.combineAndRankKeywords(structuredKeywords, bodyKeywords, phrases, pageType)
	
	// 6. Return top 30
	return c.limitToTopKeywords(allKeywords, 30)
}

// extractIndustryIndicators extracts industry-specific indicators from page content
func (c *SmartWebsiteCrawler) extractIndustryIndicators(content string) []string {
	var indicators []string
	seen := make(map[string]bool)
	
	// Convert to lowercase for processing
	text := strings.ToLower(content)
	
	// Industry-specific patterns with high confidence signals
	industryPatterns := map[string][]string{
		"food_beverage": {
			"wine", "wines", "winery", "vineyard", "vintner", "sommelier", "tasting", "cellar",
			"bottle", "vintage", "grape", "grapes", "grapevine", "oenology", "wine shop", "wine store",
			"wine bar", "wine merchant", "wine retailer", "alcohol", "spirits", "liquor", "beer",
			"brewery", "distillery", "beverage", "beverages", "restaurant", "cafe", "coffee", "food",
			"dining", "kitchen", "catering", "bakery", "bar", "pub",
		},
		"technology": {
			"technology", "software", "tech", "app", "digital", "web", "mobile", "cloud", "ai",
			"machine learning", "ml", "data", "cyber", "security", "programming", "development",
			"IT", "computer", "internet", "online", "platform", "api", "database", "saas",
		},
		"healthcare": {
			"healthcare", "medical", "clinic", "hospital", "doctor", "dentist", "therapy", "wellness",
			"pharmacy", "medicine", "patient", "treatment", "health", "care", "nurse", "physician",
		},
		"legal": {
			"legal", "law", "attorney", "lawyer", "court", "litigation", "patent", "trademark",
			"copyright", "legal services", "advocacy", "justice", "legal advice", "law firm",
		},
		"retail": {
			"retail", "retailer", "retail store", "retail shop", "brick and mortar", "brick-and-mortar",
			"physical store", "storefront", "merchandise", "inventory", "point of sale", "POS",
			"checkout", "cash register", "sales floor", "showroom", "boutique", "outlet",
			"marketplace", "vendor", "seller", "selling", "commerce", "store", "shop",
			"ecommerce", "e-commerce", "online store", "online shop", "digital storefront",
		},
		"finance": {
			"finance", "banking", "investment", "insurance", "accounting", "tax", "financial",
			"credit", "loan", "money", "capital", "funding", "payment", "transaction", "wealth",
		},
		"real_estate": {
			"real estate", "property", "construction", "building", "architecture", "design",
			"interior", "home", "house", "apartment", "rental", "mortgage", "property management",
		},
		"education": {
			"education", "school", "university", "training", "learning", "course", "academy",
			"institute", "student", "teacher", "teaching", "academic", "degree", "certification",
		},
		"consulting": {
			"consulting", "advisory", "strategy", "management", "business", "corporate", "professional",
			"services", "expert", "specialist", "consultant",
		},
		"manufacturing": {
			"manufacturing", "production", "factory", "industrial", "automotive", "machinery",
			"equipment", "assembly",
		},
		"transportation": {
			"transportation", "logistics", "shipping", "delivery", "freight", "warehouse",
			"supply chain", "trucking",
		},
		"entertainment": {
			"entertainment", "media", "marketing", "advertising", "design", "creative", "art",
			"music", "film",
		},
	}
	
	// Extract industry indicators using patterns
	for industry, patterns := range industryPatterns {
		for _, pattern := range patterns {
			// Use word boundary matching for better accuracy
			patternRegex := regexp.MustCompile(`\b` + regexp.QuoteMeta(pattern) + `\b`)
			if patternRegex.MatchString(text) {
				indicator := industry + ":" + pattern
				indicatorLower := strings.ToLower(indicator)
				if !seen[indicatorLower] {
					seen[indicatorLower] = true
					indicators = append(indicators, indicator)
				}
			}
		}
	}
	
	return indicators
}

func (c *SmartWebsiteCrawler) calculateRelevanceScore(analysis PageAnalysis) float64 {
	score := 0.0

	// Base score by page type - increased for industry-revealing pages
	switch analysis.PageType {
	case "about", "services", "products":
		score = 0.95 // Increased from 0.9 for industry-revealing pages
	case "contact", "homepage":
		score = 0.8
	case "blog", "news":
		score = 0.6
	default:
		score = 0.5
	}

	// Add 10% boost if structured data is present
	if analysis.StructuredData != nil && len(analysis.StructuredData) > 0 {
		score += 0.10
	}

	// Add 5% boost for high content quality (>0.7)
	if analysis.ContentQuality > 0.7 {
		score += 0.05
	}

	// Reduce score by 20% for low content length (<500 chars)
	if analysis.ContentLength > 0 && analysis.ContentLength < 500 {
		score *= 0.8
	}

	// Adjust based on content quality (multiply after boosts)
	score *= analysis.ContentQuality

	// Adjust based on keyword density
	if len(analysis.Keywords) > 0 {
		score += 0.1
	}

	// Adjust based on industry indicators
	if len(analysis.IndustryIndicators) > 0 {
		score += 0.1
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

func (c *SmartWebsiteCrawler) calculateContentQuality(analysis PageAnalysis) float64 {
	quality := 0.5

	// Length factor
	if analysis.ContentLength > 1000 {
		quality += 0.2
	} else if analysis.ContentLength > 500 {
		quality += 0.1
	}

	// Title factor
	if analysis.Title != "" {
		quality += 0.1
	}

	// Meta tags factor
	if len(analysis.MetaTags) > 0 {
		quality += 0.1
	}

	// Structured data factor
	if len(analysis.StructuredData) > 0 {
		quality += 0.1
	}

	if quality > 1.0 {
		quality = 1.0
	}

	return quality
}

func (c *SmartWebsiteCrawler) aggregateResults(result *CrawlResult, analyses []PageAnalysis) {
	// Aggregate keywords from all pages
	keywordCounts := make(map[string]int)
	industryScores := make(map[string]float64)

	for _, analysis := range analyses {
		// Aggregate keywords
		for _, keyword := range analysis.Keywords {
			keywordCounts[keyword]++
		}

		// Aggregate industry indicators
		for _, indicator := range analysis.IndustryIndicators {
			industryScores[indicator] += analysis.RelevanceScore
		}

		// Update business info with most relevant data
		if analysis.RelevanceScore > 0.8 && analysis.BusinessInfo.BusinessName != "" {
			result.BusinessInfo = analysis.BusinessInfo
		}
	}

	// Sort keywords by frequency
	type keywordFreq struct {
		keyword string
		count   int
	}

	var sortedKeywords []keywordFreq
	for keyword, count := range keywordCounts {
		sortedKeywords = append(sortedKeywords, keywordFreq{keyword, count})
	}

	sort.Slice(sortedKeywords, func(i, j int) bool {
		return sortedKeywords[i].count > sortedKeywords[j].count
	})

	// Take top keywords
	for i, kf := range sortedKeywords {
		if i >= 20 { // Limit to top 20 keywords
			break
		}
		result.Keywords = append(result.Keywords, kf.keyword)
	}

	result.IndustryScore = industryScores
	result.RelevantPages = len(analyses)
}
