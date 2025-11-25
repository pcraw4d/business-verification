package classification

import (
	"context"
	"encoding/json"
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
	"unicode"

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
			// Force IPv4 UDP connection to our custom DNS server
			// Ignore the network and address parameters to prevent system DNS fallback
			// Try each DNS server with retry logic
			var lastErr error
			for _, server := range dnsServers {
				d := net.Dialer{
					Timeout: 5 * time.Second,
				}
				// Always use udp4 to force IPv4, ignore the network parameter
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
	
	// Extract Open Graph type (Phase 6.3)
	ogTypeRegex := regexp.MustCompile(`(?i)<meta[^>]*property=["']og:type["'][^>]*content=["']([^"']+)["']`)
	ogTypeMatches := ogTypeRegex.FindStringSubmatch(content)
	if len(ogTypeMatches) > 1 {
		metaTags["og:type"] = strings.TrimSpace(ogTypeMatches[1])
	}
	
	// Extract Open Graph site name
	ogSiteNameRegex := regexp.MustCompile(`(?i)<meta[^>]*property=["']og:site_name["'][^>]*content=["']([^"']+)["']`)
	ogSiteNameMatches := ogSiteNameRegex.FindStringSubmatch(content)
	if len(ogSiteNameMatches) > 1 {
		metaTags["og:site_name"] = strings.TrimSpace(ogSiteNameMatches[1])
	}
	
	// Extract Twitter Card title (Phase 6.3)
	twitterTitleRegex := regexp.MustCompile(`(?i)<meta[^>]*name=["']twitter:title["'][^>]*content=["']([^"']+)["']`)
	twitterTitleMatches := twitterTitleRegex.FindStringSubmatch(content)
	if len(twitterTitleMatches) > 1 {
		metaTags["twitter:title"] = strings.TrimSpace(twitterTitleMatches[1])
	}
	
	// Extract Twitter Card description
	twitterDescRegex := regexp.MustCompile(`(?i)<meta[^>]*name=["']twitter:description["'][^>]*content=["']([^"']+)["']`)
	twitterDescMatches := twitterDescRegex.FindStringSubmatch(content)
	if len(twitterDescMatches) > 1 {
		metaTags["twitter:description"] = strings.TrimSpace(twitterDescMatches[1])
	}
	
	// Extract Twitter Card type
	twitterCardRegex := regexp.MustCompile(`(?i)<meta[^>]*name=["']twitter:card["'][^>]*content=["']([^"']+)["']`)
	twitterCardMatches := twitterCardRegex.FindStringSubmatch(content)
	if len(twitterCardMatches) > 1 {
		metaTags["twitter:card"] = strings.TrimSpace(twitterCardMatches[1])
	}
	
	return metaTags
}

func (c *SmartWebsiteCrawler) extractStructuredData(content string) map[string]interface{} {
	structuredData := make(map[string]interface{})
	
	// Extract JSON-LD structured data with proper parsing
	jsonLdRegex := regexp.MustCompile(`(?i)<script[^>]*type=["']application/ld\+json["'][^>]*>([\s\S]*?)</script>`)
	jsonLdMatches := jsonLdRegex.FindAllStringSubmatch(content, -1)
	
	for i, match := range jsonLdMatches {
		if len(match) > 1 {
			jsonContent := strings.TrimSpace(match[1])
			
			// Skip empty JSON-LD content
			if jsonContent == "" {
				continue
			}
			
			// Parse JSON-LD content
			var jsonData interface{}
			if err := json.Unmarshal([]byte(jsonContent), &jsonData); err == nil {
				// Successfully parsed - extract business information
				c.extractBusinessInfoFromJSONLD(jsonData, structuredData)
				structuredData[fmt.Sprintf("json-ld-%d", i)] = jsonData
			} else {
				// Store raw content if parsing fails
				structuredData[fmt.Sprintf("json-ld-raw-%d", i)] = jsonContent
				if c.logger != nil {
					c.logger.Printf("⚠️ [StructuredData] Failed to parse JSON-LD block %d: %v", i, err)
				}
			}
		}
	}
	
	// Extract microdata (enhanced extraction)
	c.extractMicrodata(content, structuredData)
	
	return structuredData
}

// extractBusinessInfoFromJSONLD extracts business information from parsed JSON-LD data
func (c *SmartWebsiteCrawler) extractBusinessInfoFromJSONLD(data interface{}, result map[string]interface{}) {
	switch v := data.(type) {
	case map[string]interface{}:
		c.processJSONLDObject(v, result)
	case []interface{}:
		// Handle arrays of objects
		for _, item := range v {
			c.extractBusinessInfoFromJSONLD(item, result)
		}
	}
}

// processJSONLDObject processes a single JSON-LD object
func (c *SmartWebsiteCrawler) processJSONLDObject(obj map[string]interface{}, result map[string]interface{}) {
	// Get the @type to identify Schema.org type
	typeValue, hasType := obj["@type"]
	if !hasType {
		// Try without @ prefix (some implementations use "type")
		typeValue, hasType = obj["type"]
	}
	
	if hasType {
		schemaType := fmt.Sprintf("%v", typeValue)
		// Only set schema_type if not already set, or if this is a business type (prefer business types)
		if existingType, exists := result["schema_type"]; !exists || !c.isBusinessType(fmt.Sprintf("%v", existingType)) {
			result["schema_type"] = schemaType
		}
		
		// Extract business-relevant information based on type
		if c.isBusinessType(schemaType) {
			// Extract business name
			if name, ok := c.extractStringValue(obj, "name"); ok {
				result["business_name"] = name
			}
			
			// Extract description
			if desc, ok := c.extractStringValue(obj, "description"); ok {
				result["description"] = desc
			}
			
			// Extract industry/industry code
			if industry, ok := c.extractStringValue(obj, "industry"); ok {
				result["industry"] = industry
			}
			
			// Extract services
			if services := c.extractArrayValue(obj, "service", "services"); len(services) > 0 {
				result["services"] = services
			}
			
			// Extract products
			if products := c.extractArrayValue(obj, "product", "products"); len(products) > 0 {
				result["products"] = products
			}
			
			// Extract address information
			if address, ok := obj["address"].(map[string]interface{}); ok {
				if street, ok := c.extractStringValue(address, "streetAddress"); ok {
					result["address_street"] = street
				}
				if city, ok := c.extractStringValue(address, "addressLocality"); ok {
					result["address_city"] = city
				}
				if state, ok := c.extractStringValue(address, "addressRegion"); ok {
					result["address_state"] = state
				}
			}
			
			// Extract contact information
			if phone, ok := c.extractStringValue(obj, "telephone"); ok {
				result["phone"] = phone
			}
			if email, ok := c.extractStringValue(obj, "email"); ok {
				result["email"] = email
			}
		}
	}
	
	// Recursively process nested objects
	for key, value := range obj {
		if key == "@type" || key == "type" || key == "@context" {
			continue
		}
		if nestedObj, ok := value.(map[string]interface{}); ok {
			c.processJSONLDObject(nestedObj, result)
		} else if nestedArray, ok := value.([]interface{}); ok {
			for _, item := range nestedArray {
				if nestedObj, ok := item.(map[string]interface{}); ok {
					c.processJSONLDObject(nestedObj, result)
				}
			}
		}
	}
}

// isBusinessType checks if a Schema.org type is business-related
func (c *SmartWebsiteCrawler) isBusinessType(schemaType string) bool {
	businessTypes := []string{
		"LocalBusiness", "Store", "Restaurant", "FoodEstablishment",
		"WineShop", "LiquorStore", "RetailStore", "ClothingStore",
		"ElectronicsStore", "BookStore", "ToyStore", "GroceryStore",
		"AutoDealer", "BicycleStore", "HardwareStore", "JewelryStore",
		"PetStore", "SportingGoodsStore", "TireShop", "WholesaleStore",
		"ProfessionalService", "LegalService", "AccountingService",
		"FinancialService", "RealEstateAgent", "InsuranceAgency",
		"TravelAgency", "AutomatedTeller", "BankOrCreditUnion",
		"Organization", "Corporation", "NGO", "GovernmentOrganization",
	}
	
	schemaTypeLower := strings.ToLower(schemaType)
	for _, bt := range businessTypes {
		if strings.Contains(schemaTypeLower, strings.ToLower(bt)) {
			return true
		}
	}
	return false
}

// extractStringValue safely extracts a string value from a map
func (c *SmartWebsiteCrawler) extractStringValue(obj map[string]interface{}, keys ...string) (string, bool) {
	for _, key := range keys {
		if val, ok := obj[key]; ok {
			if str, ok := val.(string); ok {
				return str, true
			}
			// Try to convert to string
			return fmt.Sprintf("%v", val), true
		}
	}
	return "", false
}

// extractArrayValue extracts array values, handling both single objects and arrays
func (c *SmartWebsiteCrawler) extractArrayValue(obj map[string]interface{}, keys ...string) []string {
	var results []string
	for _, key := range keys {
		if val, ok := obj[key]; ok {
			switch v := val.(type) {
			case []interface{}:
				for _, item := range v {
					if itemMap, ok := item.(map[string]interface{}); ok {
						if name, ok := c.extractStringValue(itemMap, "name", "title"); ok {
							results = append(results, name)
						}
					} else if str, ok := item.(string); ok {
						results = append(results, str)
					}
				}
			case map[string]interface{}:
				if name, ok := c.extractStringValue(v, "name", "title"); ok {
					results = append(results, name)
				}
			case string:
				results = append(results, v)
			}
		}
	}
	return results
}

// extractMicrodata extracts microdata with enhanced parsing
func (c *SmartWebsiteCrawler) extractMicrodata(content string, result map[string]interface{}) {
	// Look for itemscope attributes with itemtype
	itemScopeRegex := regexp.MustCompile(`(?i)<[^>]*itemscope[^>]*itemtype=["']([^"']+)["'][^>]*>`)
	itemScopeMatches := itemScopeRegex.FindAllStringSubmatch(content, -1)
	for i, match := range itemScopeMatches {
		if len(match) > 1 {
			itemType := strings.TrimSpace(match[1])
			result[fmt.Sprintf("microdata-type-%d", i)] = itemType
			
			// Check if it's a business type
			if c.isBusinessType(itemType) {
				result["has_business_microdata"] = true
			}
		}
	}
	
	// Extract itemprop values with better pattern matching
	// Pattern 1: itemprop="name" content="value"
	itemPropContentRegex := regexp.MustCompile(`(?i)itemprop=["']([^"']+)["'][^>]*content=["']([^"']+)["']`)
	itemPropMatches := itemPropContentRegex.FindAllStringSubmatch(content, -1)
	for _, match := range itemPropMatches {
		if len(match) >= 3 {
			prop := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])
			result[fmt.Sprintf("microdata-%s", prop)] = value
		}
	}
	
	// Pattern 2: <span itemprop="name">value</span>
	itemPropTagRegex := regexp.MustCompile(`(?i)<[^>]*itemprop=["']([^"']+)["'][^>]*>([^<]+)</[^>]*>`)
	itemPropTagMatches := itemPropTagRegex.FindAllStringSubmatch(content, -1)
	for _, match := range itemPropTagMatches {
		if len(match) >= 3 {
			prop := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])
			// Only add if not already set (content attribute takes precedence)
			key := fmt.Sprintf("microdata-%s", prop)
			if _, exists := result[key]; !exists {
				result[key] = value
			}
		}
	}
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

// structuredKeyword represents a keyword with its source and position weight
type structuredKeyword struct {
	keyword string
	weight  float64 // Position-based weight (title=1.0, meta=0.9, h1=0.9, h2=0.8, etc.)
}

// extractStructuredKeywords extracts keywords from structured HTML elements with position weighting
// Returns keywords with their position-based weights
func (c *SmartWebsiteCrawler) extractStructuredKeywords(content string) []structuredKeyword {
	var keywords []structuredKeyword
	seen := make(map[string]bool)
	
	// Extract from title (highest weight = 1.0)
	titleRegex := regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)
	titleMatches := titleRegex.FindStringSubmatch(content)
	if len(titleMatches) > 1 {
		titleText := strings.TrimSpace(titleMatches[1])
		titleWords := c.extractWordsFromText(titleText)
		for _, word := range titleWords {
			wordLower := strings.ToLower(word)
			if !seen[wordLower] && len(wordLower) > 2 {
				seen[wordLower] = true
				keywords = append(keywords, structuredKeyword{keyword: wordLower, weight: 1.0})
			}
		}
	}
	
	// Extract from meta description (weight = 0.9)
	metaDescRegex := regexp.MustCompile(`(?i)<meta[^>]*name=["']description["'][^>]*content=["']([^"']+)["']`)
	metaDescMatches := metaDescRegex.FindStringSubmatch(content)
	if len(metaDescMatches) > 1 {
		metaText := strings.TrimSpace(metaDescMatches[1])
		metaWords := c.extractWordsFromText(metaText)
		for _, word := range metaWords {
			wordLower := strings.ToLower(word)
			if !seen[wordLower] && len(wordLower) > 2 {
				seen[wordLower] = true
				keywords = append(keywords, structuredKeyword{keyword: wordLower, weight: 0.9})
			}
		}
	}
	
	// Extract from Open Graph title (weight = 0.92) - Phase 6.3
	ogTitleRegex := regexp.MustCompile(`(?i)<meta[^>]*property=["']og:title["'][^>]*content=["']([^"']+)["']`)
	ogTitleMatches := ogTitleRegex.FindStringSubmatch(content)
	if len(ogTitleMatches) > 1 {
		ogTitleText := strings.TrimSpace(ogTitleMatches[1])
		ogTitleWords := c.extractWordsFromText(ogTitleText)
		for _, word := range ogTitleWords {
			wordLower := strings.ToLower(word)
			if !seen[wordLower] && len(wordLower) > 2 {
				seen[wordLower] = true
				keywords = append(keywords, structuredKeyword{keyword: wordLower, weight: 0.92})
			}
		}
	}
	
	// Extract from Open Graph description (weight = 0.88) - Phase 6.3
	ogDescRegex := regexp.MustCompile(`(?i)<meta[^>]*property=["']og:description["'][^>]*content=["']([^"']+)["']`)
	ogDescMatches := ogDescRegex.FindStringSubmatch(content)
	if len(ogDescMatches) > 1 {
		ogDescText := strings.TrimSpace(ogDescMatches[1])
		ogDescWords := c.extractWordsFromText(ogDescText)
		for _, word := range ogDescWords {
			wordLower := strings.ToLower(word)
			if !seen[wordLower] && len(wordLower) > 2 {
				seen[wordLower] = true
				keywords = append(keywords, structuredKeyword{keyword: wordLower, weight: 0.88})
			}
		}
	}
	
	// Extract from Twitter Card title (weight = 0.90) - Phase 6.3
	twitterTitleRegex := regexp.MustCompile(`(?i)<meta[^>]*name=["']twitter:title["'][^>]*content=["']([^"']+)["']`)
	twitterTitleMatches := twitterTitleRegex.FindStringSubmatch(content)
	if len(twitterTitleMatches) > 1 {
		twitterTitleText := strings.TrimSpace(twitterTitleMatches[1])
		twitterTitleWords := c.extractWordsFromText(twitterTitleText)
		for _, word := range twitterTitleWords {
			wordLower := strings.ToLower(word)
			if !seen[wordLower] && len(wordLower) > 2 {
				seen[wordLower] = true
				keywords = append(keywords, structuredKeyword{keyword: wordLower, weight: 0.90})
			}
		}
	}
	
	// Extract from Twitter Card description (weight = 0.86) - Phase 6.3
	twitterDescRegex := regexp.MustCompile(`(?i)<meta[^>]*name=["']twitter:description["'][^>]*content=["']([^"']+)["']`)
	twitterDescMatches := twitterDescRegex.FindStringSubmatch(content)
	if len(twitterDescMatches) > 1 {
		twitterDescText := strings.TrimSpace(twitterDescMatches[1])
		twitterDescWords := c.extractWordsFromText(twitterDescText)
		for _, word := range twitterDescWords {
			wordLower := strings.ToLower(word)
			if !seen[wordLower] && len(wordLower) > 2 {
				seen[wordLower] = true
				keywords = append(keywords, structuredKeyword{keyword: wordLower, weight: 0.86})
			}
		}
	}
	
	// Extract from headings (h1-h6) - weighted by importance
	// h1=0.9, h2=0.8, h3=0.7, h4=0.65, h5=0.6, h6=0.55
	headingRegex := regexp.MustCompile(`(?i)<h([1-6])[^>]*>([^<]+)</h[1-6]>`)
	headingMatches := headingRegex.FindAllStringSubmatch(content, -1)
	for _, match := range headingMatches {
		if len(match) >= 3 {
			headingLevel := match[1]
			headingText := strings.TrimSpace(match[2])
			headingWords := c.extractWordsFromText(headingText)
			
			// Calculate weight based on heading level
			var weight float64
			switch headingLevel {
			case "1":
				weight = 0.9
			case "2":
				weight = 0.8
			case "3":
				weight = 0.7
			case "4":
				weight = 0.65
			case "5":
				weight = 0.6
			case "6":
				weight = 0.55
			default:
				weight = 0.6
			}
			
			for _, word := range headingWords {
				wordLower := strings.ToLower(word)
				if !seen[wordLower] && len(wordLower) > 2 {
					seen[wordLower] = true
					keywords = append(keywords, structuredKeyword{keyword: wordLower, weight: weight})
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

// combineAndRankKeywordsEnhanced combines keywords with enhanced relevance scoring
// Implements Phase 4.2 and 4.3: Context-aware extraction and relevance scoring
func (c *SmartWebsiteCrawler) combineAndRankKeywordsEnhanced(structuredKeywords []structuredKeyword, bodyKeywords, phrases []string, pageType string, textContent string) []keywordScore {
	keywordScores := make(map[string]float64)
	
	// Weight structured keywords by their position (title=1.0, meta=0.9, h1=0.9, h2=0.8, etc.)
	for _, skw := range structuredKeywords {
		keywordScores[skw.keyword] += skw.weight
	}
	
	// Body keywords get medium weight (0.6) with frequency boost
	textLower := strings.ToLower(textContent)
	for _, kw := range bodyKeywords {
		// Base weight for body keywords
		baseWeight := 0.6
		
		// Count frequency in content using word boundaries to avoid substring matches
		freq := c.countKeywordFrequency(textLower, kw)
		frequencyBoost := float64(freq) * 0.05 // 5% boost per occurrence, capped at 0.2
		if frequencyBoost > 0.2 {
			frequencyBoost = 0.2
		}
		
		keywordScores[kw] += baseWeight + frequencyBoost
	}
	
	// Phrases get weight based on length and frequency
	for _, phrase := range phrases {
		wordCount := len(strings.Fields(phrase))
		baseWeight := 0.5 + float64(wordCount-2)*0.1 // 2-word: 0.5, 3-word: 0.6
		if baseWeight > 0.8 {
			baseWeight = 0.8 // Cap at 0.8
		}
		
		// Frequency boost for phrases (use word boundaries for multi-word phrases)
		freq := c.countPhraseFrequency(textLower, phrase)
		frequencyBoost := float64(freq) * 0.03 // 3% boost per occurrence, capped at 0.15
		if frequencyBoost > 0.15 {
			frequencyBoost = 0.15
		}
		
		keywordScores[phrase] += baseWeight + frequencyBoost
	}
	
	// Boost keywords based on page type relevance (Phase 4.2)
	pageTypeBoost := c.getPageTypeBoost(pageType)
	for kw := range keywordScores {
		keywordScores[kw] *= pageTypeBoost
	}
	
	// Co-occurrence boost: keywords that appear together get a small boost
	// (This is simplified - full co-occurrence analysis would be more complex)
	coOccurrenceBoost := c.calculateCoOccurrenceBoost(keywordScores, textLower)
	for kw, boost := range coOccurrenceBoost {
		keywordScores[kw] += boost
	}
	
	// Normalize scores to 0-1 range
	maxScore := 0.0
	for _, score := range keywordScores {
		if score > maxScore {
			maxScore = score
		}
	}
	if maxScore > 0 {
		for kw := range keywordScores {
			keywordScores[kw] /= maxScore
		}
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

// countKeywordFrequency counts keyword occurrences using word boundaries to avoid substring matches
func (c *SmartWebsiteCrawler) countKeywordFrequency(text, keyword string) int {
	// Use word boundaries to avoid counting substrings (e.g., "wine" in "winery")
	pattern := `\b` + regexp.QuoteMeta(keyword) + `\b`
	matches := regexp.MustCompile(pattern).FindAllString(text, -1)
	return len(matches)
}

// countPhraseFrequency counts phrase occurrences in text
func (c *SmartWebsiteCrawler) countPhraseFrequency(text, phrase string) int {
	// For phrases, use simple count but escape special regex characters
	escapedPhrase := regexp.QuoteMeta(phrase)
	matches := regexp.MustCompile(escapedPhrase).FindAllString(text, -1)
	return len(matches)
}

// calculateCoOccurrenceBoost calculates a small boost for keywords that appear near each other
// Optimized to limit computation for large keyword sets
func (c *SmartWebsiteCrawler) calculateCoOccurrenceBoost(keywordScores map[string]float64, text string) map[string]float64 {
	boosts := make(map[string]float64)
	
	// Limit co-occurrence analysis to top 20 keywords to avoid O(n²) performance issues
	keywords := make([]string, 0, len(keywordScores))
	for kw := range keywordScores {
		keywords = append(keywords, kw)
	}
	
	// If too many keywords, only analyze top ones
	maxKeywords := 20
	if len(keywords) > maxKeywords {
		// Sort by score and take top N
		type kwScore struct {
			keyword string
			score   float64
		}
		scored := make([]kwScore, 0, len(keywords))
		for _, kw := range keywords {
			scored = append(scored, kwScore{keyword: kw, score: keywordScores[kw]})
		}
		sort.Slice(scored, func(i, j int) bool {
			return scored[i].score > scored[j].score
		})
		keywords = make([]string, 0, maxKeywords)
		for i := 0; i < maxKeywords && i < len(scored); i++ {
			keywords = append(keywords, scored[i].keyword)
		}
	}
	
	// Check pairs of keywords (limited set)
	for i, kw1 := range keywords {
		for j, kw2 := range keywords {
			if i >= j || len(kw1) < 3 || len(kw2) < 3 {
				continue
			}
			
			// Find all positions of both keywords using word boundaries
			pattern1 := `\b` + regexp.QuoteMeta(kw1) + `\b`
			pattern2 := `\b` + regexp.QuoteMeta(kw2) + `\b`
			
			indices1 := regexp.MustCompile(pattern1).FindAllStringIndex(text, -1)
			indices2 := regexp.MustCompile(pattern2).FindAllStringIndex(text, -1)
			
			// Check if any occurrences are within 50 characters
			foundCoOccurrence := false
			for _, idx1 := range indices1 {
				for _, idx2 := range indices2 {
					pos1 := idx1[0]
					pos2 := idx2[0]
					distance := pos1 - pos2
					if distance < 0 {
						distance = -distance
					}
					
					// If keywords appear within 50 characters, give small boost
					if distance < 50 {
						boosts[kw1] += 0.02
						boosts[kw2] += 0.02
						foundCoOccurrence = true
						break
					}
				}
				if foundCoOccurrence {
					break
				}
			}
		}
	}
	
	return boosts
}

// limitToTopKeywordsWithThreshold filters keywords by relevance threshold and returns top N
func (c *SmartWebsiteCrawler) limitToTopKeywordsWithThreshold(scoredKeywords []keywordScore, limit int, threshold float64) []string {
	// Filter by threshold first
	var filtered []keywordScore
	for _, kw := range scoredKeywords {
		if kw.score >= threshold {
			filtered = append(filtered, kw)
		}
	}
	
	// Limit to top N
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}
	
	keywords := make([]string, len(filtered))
	for i, kw := range filtered {
		keywords[i] = kw.keyword
	}
	
	return keywords
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
// Phase 6: Enhanced with structured data keyword extraction
func (c *SmartWebsiteCrawler) extractPageKeywords(content string, pageType string) []string {
	// 1. Extract clean text from HTML
	textContent := c.extractTextFromHTML(content)
	
	// 2. Extract from structured elements (title, meta, headings) with position weights
	structuredKeywords := c.extractStructuredKeywords(content)
	
	// 3. Extract from structured data (JSON-LD, microdata) - Phase 6
	structuredData := c.extractStructuredData(content)
	structuredDataKeywords := c.extractKeywordsFromStructuredData(structuredData)
	
	// 4. Extract from body text using business patterns
	bodyKeywords := c.extractBusinessKeywordsFromText(textContent)
	
	// 5. Extract phrases (2-word, 3-word)
	phrases := c.extractPhrases(textContent, 2, 3)
	
	// 6. Combine structured data keywords with other structured keywords (high weight)
	allStructuredKeywords := append(structuredKeywords, structuredDataKeywords...)
	
	// 7. Combine, deduplicate, and rank with enhanced relevance scoring
	allKeywords := c.combineAndRankKeywordsEnhanced(allStructuredKeywords, bodyKeywords, phrases, pageType, textContent)
	
	// 8. Filter by relevance threshold (0.3) and return top 30
	return c.limitToTopKeywordsWithThreshold(allKeywords, 30, 0.3)
}

// extractKeywordsFromStructuredData extracts keywords from structured data (JSON-LD, microdata)
// Phase 6.1 and 6.2: Extract keywords from parsed structured data
func (c *SmartWebsiteCrawler) extractKeywordsFromStructuredData(structuredData map[string]interface{}) []structuredKeyword {
	var keywords []structuredKeyword
	seen := make(map[string]bool)
	
	// Extract from business name (high weight)
	if name, ok := structuredData["business_name"].(string); ok && name != "" {
		words := c.extractWordsFromText(name)
		for _, word := range words {
			wordLower := strings.ToLower(word)
			if len(wordLower) >= 3 && !seen[wordLower] {
				seen[wordLower] = true
				keywords = append(keywords, structuredKeyword{
					keyword: wordLower,
					weight:  0.95, // Very high weight for structured data business name
				})
			}
		}
	}
	
	// Extract from description (high weight)
	if desc, ok := structuredData["description"].(string); ok && desc != "" {
		words := c.extractWordsFromText(desc)
		for _, word := range words {
			wordLower := strings.ToLower(word)
			if len(wordLower) >= 3 && !seen[wordLower] {
				seen[wordLower] = true
				keywords = append(keywords, structuredKeyword{
					keyword: wordLower,
					weight:  0.90, // High weight for structured data description
				})
			}
		}
	}
	
	// Extract from industry (very high weight)
	if industry, ok := structuredData["industry"].(string); ok && industry != "" {
		words := c.extractWordsFromText(industry)
		for _, word := range words {
			wordLower := strings.ToLower(word)
			if len(wordLower) >= 3 && !seen[wordLower] {
				seen[wordLower] = true
				keywords = append(keywords, structuredKeyword{
					keyword: wordLower,
					weight:  1.0, // Maximum weight for industry from structured data
				})
			}
		}
	}
	
	// Extract from services (medium-high weight)
	if services, ok := structuredData["services"].([]string); ok && len(services) > 0 {
		for _, service := range services {
			if service == "" {
				continue
			}
			words := c.extractWordsFromText(service)
			for _, word := range words {
				wordLower := strings.ToLower(word)
				if len(wordLower) >= 3 && !seen[wordLower] {
					seen[wordLower] = true
					keywords = append(keywords, structuredKeyword{
						keyword: wordLower,
						weight:  0.85, // High weight for services
					})
				}
			}
		}
	}
	
	// Extract from products (medium-high weight)
	if products, ok := structuredData["products"].([]string); ok && len(products) > 0 {
		for _, product := range products {
			if product == "" {
				continue
			}
			words := c.extractWordsFromText(product)
			for _, word := range words {
				wordLower := strings.ToLower(word)
				if len(wordLower) >= 3 && !seen[wordLower] {
					seen[wordLower] = true
					keywords = append(keywords, structuredKeyword{
						keyword: wordLower,
						weight:  0.85, // High weight for products
					})
				}
			}
		}
	}
	
	// Extract from Schema.org type (high weight for business type keywords)
	if schemaType, ok := structuredData["schema_type"].(string); ok && schemaType != "" {
		// Extract meaningful parts from Schema.org type (e.g., "WineShop" -> "wine", "shop")
		typeWords := c.splitCamelCase(schemaType)
		for _, word := range typeWords {
			wordLower := strings.ToLower(word)
			if len(wordLower) >= 3 && !seen[wordLower] {
				seen[wordLower] = true
				keywords = append(keywords, structuredKeyword{
					keyword: wordLower,
					weight:  0.88, // High weight for Schema.org type
				})
			}
		}
	}
	
	// Extract from microdata properties
	for key, value := range structuredData {
		if strings.HasPrefix(key, "microdata-") {
			if strValue, ok := value.(string); ok && strValue != "" {
				words := c.extractWordsFromText(strValue)
				for _, word := range words {
					wordLower := strings.ToLower(word)
					if len(wordLower) >= 3 && !seen[wordLower] {
						seen[wordLower] = true
						keywords = append(keywords, structuredKeyword{
							keyword: wordLower,
							weight:  0.80, // Medium-high weight for microdata
						})
					}
				}
			}
		}
	}
	
	return keywords
}

// splitCamelCase splits camelCase or PascalCase strings into words
// Handles edge cases: empty strings, single characters, all caps, etc.
func (c *SmartWebsiteCrawler) splitCamelCase(s string) []string {
	if s == "" {
		return []string{}
	}
	
	var words []string
	var currentWord strings.Builder
	
	for i, r := range s {
		// If we encounter an uppercase letter and we have a current word, start a new word
		if i > 0 && unicode.IsUpper(r) && currentWord.Len() > 0 {
			word := currentWord.String()
			if len(word) >= 2 { // Only add words with 2+ characters
				words = append(words, word)
			}
			currentWord.Reset()
		}
		currentWord.WriteRune(unicode.ToLower(r))
	}
	
	// Add the last word
	if currentWord.Len() >= 2 {
		words = append(words, currentWord.String())
	}
	
	// If no words were split (e.g., all lowercase or single word), return the lowercase version
	if len(words) == 0 && len(s) >= 2 {
		words = append(words, strings.ToLower(s))
	}
	
	return words
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
