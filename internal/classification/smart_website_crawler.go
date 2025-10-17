package classification

import (
	"context"
	"fmt"
	"io"
	"log"
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

// CrawlResult represents the result of a smart crawl operation
type CrawlResult struct {
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

// PageAnalysis represents analysis of a single page
type PageAnalysis struct {
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

// BusinessInfo represents extracted business information
type BusinessInfo struct {
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

// ContactInfo represents contact information
type ContactInfo struct {
	Phone   string            `json:"phone"`
	Email   string            `json:"email"`
	Address string            `json:"address"`
	Website string            `json:"website"`
	Social  map[string]string `json:"social"`
}

// SiteStructure represents the discovered site structure
type SiteStructure struct {
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
	return &SmartWebsiteCrawler{
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
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

	// Homepage gets highest priority
	if pageURL == baseURL || pageURL == baseURL+"/" {
		return 100
	}

	// High priority pages
	highPriorityPatterns := []string{
		"/about", "/about-us", "/aboutus", "/company", "/mission", "/vision",
		"/services", "/products", "/shop", "/store",
	}

	for _, pattern := range highPriorityPatterns {
		if strings.Contains(strings.ToLower(pageURL), pattern) {
			priority += 80
			break
		}
	}

	// Medium priority pages
	mediumPriorityPatterns := []string{
		"/contact", "/contact-us", "/contactus", "/team", "/careers", "/jobs",
	}

	for _, pattern := range mediumPriorityPatterns {
		if strings.Contains(strings.ToLower(pageURL), pattern) {
			priority += 60
			break
		}
	}

	// Lower priority pages
	lowPriorityPatterns := []string{
		"/blog", "/news", "/support", "/help", "/faq", "/privacy", "/terms", "/legal",
	}

	for _, pattern := range lowPriorityPatterns {
		if strings.Contains(strings.ToLower(pageURL), pattern) {
			priority += 40
			break
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

	resp, err := c.client.Do(req)
	if err != nil {
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
	// Implementation for extracting meta tags
	return metaTags
}

func (c *SmartWebsiteCrawler) extractStructuredData(content string) map[string]interface{} {
	structuredData := make(map[string]interface{})
	// Implementation for extracting structured data
	return structuredData
}

func (c *SmartWebsiteCrawler) extractBusinessInfo(content string, pageType string) BusinessInfo {
	businessInfo := BusinessInfo{}
	// Implementation for extracting business information
	return businessInfo
}

func (c *SmartWebsiteCrawler) extractPageKeywords(content string, pageType string) []string {
	var keywords []string
	// Implementation for extracting page-specific keywords
	return keywords
}

func (c *SmartWebsiteCrawler) extractIndustryIndicators(content string) []string {
	var indicators []string
	// Implementation for extracting industry indicators
	return indicators
}

func (c *SmartWebsiteCrawler) calculateRelevanceScore(analysis PageAnalysis) float64 {
	score := 0.0

	// Base score by page type
	switch analysis.PageType {
	case "about", "services", "products":
		score = 0.9
	case "contact", "homepage":
		score = 0.8
	case "blog", "news":
		score = 0.6
	default:
		score = 0.5
	}

	// Adjust based on content quality
	score *= analysis.ContentQuality

	// Adjust based on keyword density
	if len(analysis.Keywords) > 0 {
		score += 0.1
	}

	// Adjust based on industry indicators
	if len(analysis.IndustryIndicators) > 0 {
		score += 0.1
	}

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
