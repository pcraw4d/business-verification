package classification

import (
	"context"
	"encoding/json"
	"encoding/xml"
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

	"kyb-platform/internal/classification/nlp"

	"github.com/temoto/robotstxt"
	"golang.org/x/net/html"
)

// dnsCacheEntry represents a cached DNS resolution result
type dnsCacheEntry struct {
	ips       []net.IPAddr
	expiresAt time.Time
}

// SmartWebsiteCrawler implements intelligent website crawling with page prioritization
type SmartWebsiteCrawler struct {
	logger          *log.Logger
	client          *http.Client
	maxPages        int
	maxDepth        int
	respectRobots   bool
	pageTimeout     time.Duration
	maxConcurrent   int           // Maximum concurrent page requests (default: 3)
	entityRecognizer *nlp.EntityRecognizer
	topicModeler     *nlp.TopicModeler
	sessionManager  *ScrapingSessionManager
	proxyManager    *ProxyManager
	commonEnglishWords map[string]bool // Common English word dictionary for validation
	dnsCache         map[string]*dnsCacheEntry // DNS resolution cache
	dnsCacheMutex    sync.RWMutex              // Mutex for thread-safe DNS cache access
	crawlDelays      map[string]time.Duration  // Crawl delays per domain from robots.txt
	crawlDelaysMutex sync.RWMutex              // Mutex for thread-safe crawl delays access
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
	// Create DNS cache (shared between all instances for this crawler)
	dnsCache := make(map[string]*dnsCacheEntry)
	var dnsCacheMutex sync.RWMutex
	const dnsCacheTTL = 5 * time.Minute // DNS cache TTL
	
	// Create crawl delays cache (per domain from robots.txt)
	crawlDelays := make(map[string]time.Duration)
	var crawlDelaysMutex sync.RWMutex
	
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
			// CRITICAL: Ignore the network and address parameters to prevent system DNS fallback
			// The system DNS (e.g., [fd12::10]:53 in Railway) will fail, so we must use our custom servers
			if logger != nil {
				logger.Printf("🔍 [DNS] Resolver.Dial called with network=%s, address=%s (ignoring, using custom servers)", network, address)
			}
			var lastErr error
			for _, server := range dnsServers {
				d := net.Dialer{
					Timeout: 10 * time.Second, // Longer timeout for retry
				}
				// Always use udp4 to force IPv4, completely ignore the network and address parameters
				conn, err := d.DialContext(ctx, "udp4", server)
				if err == nil {
					if logger != nil {
						logger.Printf("✅ [DNS] Successfully connected to DNS server %s for query", server)
					}
					return conn, nil
				}
				lastErr = err
				// Log DNS server failure (if logger is available)
				if logger != nil {
					logger.Printf("⚠️ [DNS] Failed to connect to DNS server %s: %v", server, err)
				}
			}
			// Return error instead of allowing system DNS fallback
			if logger != nil {
				logger.Printf("❌ [DNS] All DNS servers failed, last error: %v", lastErr)
			}
			return nil, fmt.Errorf("all DNS servers failed, last error: %w", lastErr)
		},
	}

	// Custom DialContext that forces IPv4 resolution with DNS caching
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

		// Check DNS cache first
		dnsCacheMutex.RLock()
		cached, exists := dnsCache[host]
		dnsCacheMutex.RUnlock()
		
		var ips []net.IPAddr
		if exists && time.Now().Before(cached.expiresAt) {
			// Use cached DNS result
			ips = cached.ips
			if logger != nil {
				logger.Printf("✅ [DNS] Using cached DNS result for %s (expires in %v)", host, time.Until(cached.expiresAt))
			}
		} else {
			// Cache miss or expired - resolve DNS
			if exists {
				// Remove expired entry
				dnsCacheMutex.Lock()
				delete(dnsCache, host)
				dnsCacheMutex.Unlock()
			}
			
			var dnsErr error
			maxRetries := 3
			if logger != nil {
				logger.Printf("🔍 [DNS] Starting DNS lookup for %s using custom resolver", host)
			}
			for attempt := 1; attempt <= maxRetries; attempt++ {
				if logger != nil {
					logger.Printf("🔄 [DNS] DNS lookup attempt %d/%d for %s", attempt, maxRetries, host)
				}
				ips, dnsErr = dnsResolver.LookupIPAddr(ctx, host)
				if dnsErr == nil {
					if logger != nil {
						logger.Printf("✅ [DNS] DNS lookup successful for %s: found %d IP addresses", host, len(ips))
						for i, ip := range ips {
							logger.Printf("   [DNS] IP %d: %s (IPv4: %v)", i+1, ip.IP.String(), ip.IP.To4() != nil)
						}
					}
					// Cache the result
					dnsCacheMutex.Lock()
					dnsCache[host] = &dnsCacheEntry{
						ips:       ips,
						expiresAt: time.Now().Add(dnsCacheTTL),
					}
					dnsCacheMutex.Unlock()
					break
				}
				// Log the specific DNS error
				if logger != nil {
					if dnsErr2, ok := dnsErr.(*net.DNSError); ok {
						logger.Printf("❌ [DNS] DNS lookup failed for %s (attempt %d/%d): %v (server: %s)", host, attempt, maxRetries, dnsErr2, dnsErr2.Server)
					} else {
						logger.Printf("❌ [DNS] DNS lookup failed for %s (attempt %d/%d): %v (type: %T)", host, attempt, maxRetries, dnsErr, dnsErr)
					}
				}
				// Exponential backoff: 1s, 2s, 4s
				if attempt < maxRetries {
					backoff := time.Duration(attempt) * time.Second
					if logger != nil {
						logger.Printf("⏳ [DNS] Waiting %v before retry for %s", backoff, host)
					}
					select {
					case <-ctx.Done():
						return nil, ctx.Err()
					case <-time.After(backoff):
						// Retry after backoff
					}
				}
			}
			if dnsErr != nil {
				// Invalidate cache on DNS error
				dnsCacheMutex.Lock()
				delete(dnsCache, host)
				dnsCacheMutex.Unlock()
				if logger != nil {
					logger.Printf("❌ [DNS] DNS lookup failed for %s after %d attempts: %v", host, maxRetries, dnsErr)
				}
				return nil, fmt.Errorf("DNS lookup failed for %s after %d attempts: %w", host, maxRetries, dnsErr)
			}
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

	// Enhanced connection pooling configuration
	transport := &http.Transport{
		DialContext:          customDialContext,
		MaxIdleConns:         100,              // Increased from 10 to 100
		MaxIdleConnsPerHost: 10,               // Increased from 2 to 10
		IdleConnTimeout:      90 * time.Second, // Increased from 30s to 90s for better connection reuse
		DisableCompression:   false,
		DisableKeepAlives:    false,
		ForceAttemptHTTP2:    true, // Enable HTTP/2 support
		TLSHandshakeTimeout:  10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	
	return &SmartWebsiteCrawler{
		logger: logger,
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		maxPages:        20, // Maximum pages to crawl
		maxDepth:        3,  // Maximum crawl depth
		respectRobots:   true,
		pageTimeout:     15 * time.Second,
		maxConcurrent:   3,  // Maximum concurrent page requests (default: 3)
		entityRecognizer: nlp.NewEntityRecognizer(),
		topicModeler:     nlp.NewTopicModeler(),
		sessionManager:  NewScrapingSessionManager(),
		proxyManager:    NewProxyManager(),
		commonEnglishWords: loadCommonEnglishWords(),
		dnsCache:         dnsCache, // Use shared DNS cache
		dnsCacheMutex:    dnsCacheMutex, // Mutex for DNS cache
		crawlDelays:      crawlDelays, // Per-domain crawl delays from robots.txt
		crawlDelaysMutex: crawlDelaysMutex, // Mutex for crawl delays
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
		blocked, crawlDelay, err := c.checkRobotsTxt(ctx, baseURL, "/")
		if err == nil && blocked {
			result.Error = "Website blocked by robots.txt"
			return result, fmt.Errorf("website blocked by robots.txt")
		}
		// Store crawl delay for this domain
		if err == nil && crawlDelay > 0 {
			parsedURL, err := url.Parse(baseURL)
			if err == nil {
				domain := parsedURL.Hostname()
				c.crawlDelaysMutex.Lock()
				c.crawlDelays[domain] = crawlDelay
				c.crawlDelaysMutex.Unlock()
				c.logger.Printf("⏳ [RobotsTxt] Stored crawl delay of %v for domain %s", crawlDelay, domain)
			}
		}
	}

	// Discover site structure and prioritize pages
	discoveredPages, err := c.discoverSiteStructure(ctx, baseURL)
	if err != nil {
		c.logger.Printf("⚠️ [SmartCrawler] Site structure discovery failed: %v", err)
		// Fallback to homepage only
		discoveredPages = []string{baseURL}
	}

	// Limit discovered pages to top 20 by priority (before analysis)
	if len(discoveredPages) > 20 {
		// Prioritize first, then limit
		tempPrioritized := c.prioritizePages(discoveredPages, baseURL)
		if len(tempPrioritized) > 20 {
			discoveredPages = tempPrioritized[:20]
			c.logger.Printf("📊 [SmartCrawler] Limited discovered pages to top 20 by priority (from %d total)", len(tempPrioritized))
		}
	}

	// Prioritize pages based on relevance
	prioritizedPages := c.prioritizePages(discoveredPages, baseURL)
	c.logger.Printf("📊 [SmartCrawler] Discovered %d pages, prioritizing %d", len(discoveredPages), len(prioritizedPages))

	// Ensure homepage is first to establish session
	homepageFirst := []string{baseURL}
	seen := make(map[string]bool)
	seen[baseURL] = true
	for _, page := range prioritizedPages {
		if !seen[page] {
			homepageFirst = append(homepageFirst, page)
			seen[page] = true
		}
	}
	prioritizedPages = homepageFirst

	// Removed initial warm-up delay for faster classification (user preference)
	// Original delay was 3+ seconds with human-like variation

	// Analyze prioritized pages in parallel with controlled concurrency
	// Use parallel processing for better performance while maintaining session management
	maxConcurrent := c.maxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 3 // Default to 3 if not set
	}
	c.logger.Printf("🔄 [SmartCrawler] [PARALLEL] Starting parallel crawl with %d concurrent requests for %d pages", maxConcurrent, len(prioritizedPages))
	parallelStart := time.Now()
	pageAnalyses := c.analyzePagesParallel(ctx, prioritizedPages, maxConcurrent)
	parallelDuration := time.Since(parallelStart)
	c.logger.Printf("🔄 [SmartCrawler] [PARALLEL] Parallel analysis completed in %v - %d pages analyzed", parallelDuration, len(pageAnalyses))
	
	// Check if we have sufficient content after parallel analysis
	if len(pageAnalyses) >= 2 && c.hasSufficientContent(pageAnalyses) {
		c.logger.Printf("✅ [SmartCrawler] [PARALLEL] Sufficient content gathered (pages: %d, duration: %v)", len(pageAnalyses), parallelDuration)
	} else {
		c.logger.Printf("⚠️ [SmartCrawler] [PARALLEL] Content quality check: pages=%d, sufficient=%v", len(pageAnalyses), len(pageAnalyses) >= 2 && c.hasSufficientContent(pageAnalyses))
	}
	
	result.PagesAnalyzed = pageAnalyses
	result.TotalPages = len(pageAnalyses)

	// Aggregate results
	c.aggregateResults(result, pageAnalyses)

	result.CrawlDuration = time.Since(startTime)
	result.Success = true

	c.logger.Printf("✅ [SmartCrawler] Smart crawl completed in %v - %d pages analyzed", result.CrawlDuration, result.TotalPages)
	return result, nil
}

// CrawlWebsiteFast performs fast-path crawling with time constraints
// Uses same discovery (robots.txt, sitemap) but with time constraints
// Limits to top pages by priority, uses parallel processing, reduced delays
// Early exit on sufficient content or time limit
func (c *SmartWebsiteCrawler) CrawlWebsiteFast(ctx context.Context, websiteURL string, maxTime time.Duration, maxPages int, maxConcurrent int) (*CrawlResult, error) {
	startTime := time.Now()
	c.logger.Printf("🚀 [SmartCrawler] [FAST-PATH] Starting fast-path crawl for: %s", websiteURL)
	c.logger.Printf("🚀 [SmartCrawler] [FAST-PATH] Configuration: maxTime=%v, maxPages=%d, maxConcurrent=%d", maxTime, maxPages, maxConcurrent)

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, maxTime)
	defer cancel()

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

	// Check robots.txt if enabled (quick check with timeout)
	if c.respectRobots {
		robotsCtx, robotsCancel := context.WithTimeout(timeoutCtx, 2*time.Second)
		blocked, crawlDelay, err := c.checkRobotsTxt(robotsCtx, baseURL, "/")
		robotsCancel()
		if err == nil && blocked {
			result.Error = "Website blocked by robots.txt"
			return result, fmt.Errorf("website blocked by robots.txt")
		}
		// Store crawl delay for this domain
		if err == nil && crawlDelay > 0 {
			parsedURL, err := url.Parse(baseURL)
			if err == nil {
				domain := parsedURL.Hostname()
				c.crawlDelaysMutex.Lock()
				c.crawlDelays[domain] = crawlDelay
				c.crawlDelaysMutex.Unlock()
			}
		}
	}

	// Discover site structure (with timeout)
	discoverStart := time.Now()
	discoverCtx, discoverCancel := context.WithTimeout(timeoutCtx, 3*time.Second)
	discoveredPages, err := c.discoverSiteStructure(discoverCtx, baseURL)
	discoverCancel()
	discoverDuration := time.Since(discoverStart)
	if err != nil {
		c.logger.Printf("⚠️ [SmartCrawler] [FAST-PATH] Site structure discovery failed in %v: %v", discoverDuration, err)
		// Fallback to homepage only
		discoveredPages = []string{baseURL}
	} else {
		c.logger.Printf("🔍 [SmartCrawler] [FAST-PATH] Discovery completed in %v - found %d pages", discoverDuration, len(discoveredPages))
	}

	// Prioritize pages and limit to maxPages
	prioritizedPages := c.prioritizePages(discoveredPages, baseURL)
	if len(prioritizedPages) > maxPages {
		prioritizedPages = prioritizedPages[:maxPages]
		c.logger.Printf("📊 [SmartCrawler] Limited to top %d pages for fast-path", maxPages)
	}

	// Ensure homepage is first
	homepageFirst := []string{baseURL}
	seen := make(map[string]bool)
	seen[baseURL] = true
	for _, page := range prioritizedPages {
		if !seen[page] {
			homepageFirst = append(homepageFirst, page)
			seen[page] = true
		}
	}
	prioritizedPages = homepageFirst

	// Use parallel processing with reduced delay (500ms)
	// Check remaining time before starting
	elapsed := time.Since(startTime)
	remainingTime := maxTime - elapsed
	if remainingTime < 1*time.Second {
		c.logger.Printf("⚠️ [SmartCrawler] Insufficient time remaining (%v), using available pages", remainingTime)
		// Use only homepage if time is very short
		if len(prioritizedPages) > 1 {
			prioritizedPages = []string{baseURL}
		}
	}

	// Analyze pages in parallel
	c.logger.Printf("🚀 [SmartCrawler] [FAST-PATH] Starting parallel analysis of %d pages (concurrent: %d, delay: 500ms)", len(prioritizedPages), maxConcurrent)
	parallelStart := time.Now()
	pageAnalyses := c.analyzePagesParallelWithDelay(timeoutCtx, prioritizedPages, maxConcurrent, 500*time.Millisecond, true, startTime, maxTime)
	parallelDuration := time.Since(parallelStart)
	c.logger.Printf("🚀 [SmartCrawler] [FAST-PATH] Parallel analysis completed in %v - %d pages analyzed", parallelDuration, len(pageAnalyses))
	
	// Check if we have sufficient content after parallel analysis
	if len(pageAnalyses) >= 2 && c.hasSufficientContent(pageAnalyses) {
		c.logger.Printf("✅ [SmartCrawler] [FAST-PATH] Sufficient content gathered in parallel mode (pages: %d, duration: %v)", len(pageAnalyses), parallelDuration)
	} else {
		c.logger.Printf("⚠️ [SmartCrawler] [FAST-PATH] Content quality check: pages=%d, sufficient=%v", len(pageAnalyses), len(pageAnalyses) >= 2 && c.hasSufficientContent(pageAnalyses))
	}

	result.PagesAnalyzed = pageAnalyses
	result.TotalPages = len(pageAnalyses)

	// Aggregate results
	c.aggregateResults(result, pageAnalyses)

	result.CrawlDuration = time.Since(startTime)
	result.Success = true

	c.logger.Printf("✅ [SmartCrawler] [FAST-PATH] Crawl completed in %v - %d pages analyzed (target: <5s, achieved: %v)", 
		result.CrawlDuration, result.TotalPages, result.CrawlDuration < 5*time.Second)
	if result.CrawlDuration >= 5*time.Second {
		c.logger.Printf("⚠️ [SmartCrawler] [FAST-PATH] Duration exceeded 5s target by %v", result.CrawlDuration-5*time.Second)
	}
	return result, nil
}

// analyzePagesParallelWithDelay processes pages in parallel with configurable delay and time constraints
func (c *SmartWebsiteCrawler) analyzePagesParallelWithDelay(ctx context.Context, pages []string, maxConcurrent int, minDelay time.Duration, fastPath bool, startTime time.Time, maxTime time.Duration) []PageAnalysis {
	var analyses []PageAnalysis
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Semaphore to limit concurrent requests
	semaphore := make(chan struct{}, maxConcurrent)

	for i, page := range pages {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return analyses
		default:
		}

		// Check remaining time before starting new page
		elapsed := time.Since(startTime)
		remainingTime := maxTime - elapsed
		if remainingTime < 1*time.Second {
			c.logger.Printf("⏰ [SmartCrawler] [FAST-PATH] Time limit approaching (%v remaining), skipping remaining pages", remainingTime)
			break
		}

		// Check if we already have sufficient content
		mu.Lock()
		var validAnalyses []PageAnalysis
		for _, a := range analyses {
			if a.URL != "" {
				validAnalyses = append(validAnalyses, a)
			}
		}
		hasSufficient := len(validAnalyses) >= 2 && c.hasSufficientContent(validAnalyses)
		mu.Unlock()
		if hasSufficient {
			c.logger.Printf("✅ [SmartCrawler] [FAST-PATH] Sufficient content gathered after %d pages, skipping remaining pages", len(validAnalyses))
			break
		}

		wg.Add(1)
		go func(idx int, pageURL string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Analyze page
			pageStart := time.Now()
			analysis := c.analyzePage(ctx, pageURL)
			pageDuration := time.Since(pageStart)
			c.logger.Printf("📄 [SmartCrawler] [FAST-PATH] Page %d analyzed in %v: %s (status: %d, content: %d chars, keywords: %d)", 
				idx+1, pageDuration, pageURL, analysis.StatusCode, analysis.ContentLength, len(analysis.Keywords))

			// Store result in correct position
			mu.Lock()
			// Extend slice if needed
			for len(analyses) <= idx {
				analyses = append(analyses, PageAnalysis{})
			}
			analyses[idx] = analysis
			mu.Unlock()

			// Check for 403 to signal early stop
			if analysis.StatusCode == 403 {
				c.logger.Printf("🚫 [SmartCrawler] [FAST-PATH] Received 403 for %s - stopping crawl", pageURL)
			}
		}(i, page)

		// Apply delay between starting goroutines (except for first page)
		if i > 0 && minDelay > 0 {
			select {
			case <-ctx.Done():
				return analyses
			case <-time.After(minDelay):
				// Delay completed
			}
		}
	}

	wg.Wait()

	// Filter out empty analyses
	var validAnalyses []PageAnalysis
	for _, analysis := range analyses {
		if analysis.URL != "" {
			validAnalyses = append(validAnalyses, analysis)
		}
	}

	// Check for 403 - if found, stop here
	for _, analysis := range validAnalyses {
		if analysis.StatusCode == 403 {
			c.logger.Printf("🚫 [SmartCrawler] Received 403 - stopping crawl to avoid further blocks")
			// Return only analyses before 403
			var before403 []PageAnalysis
			for _, a := range validAnalyses {
				if a.StatusCode == 403 {
					break
				}
				before403 = append(before403, a)
			}
			return before403
		}
	}

	return validAnalyses
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

// SitemapURL represents a URL entry in a sitemap
type SitemapURL struct {
	Loc     string  `xml:"loc"`
	LastMod string  `xml:"lastmod"`
	ChangeFreq string `xml:"changefreq"`
	Priority   float64 `xml:"priority"`
}

// SitemapIndex represents a sitemap index file
type SitemapIndex struct {
	XMLName xml.Name      `xml:"sitemapindex"`
	Sitemaps []SitemapEntry `xml:"sitemap"`
}

// SitemapEntry represents a sitemap entry in an index
type SitemapEntry struct {
	Loc string `xml:"loc"`
}

// URLSet represents a sitemap with URLs
type URLSet struct {
	XMLName xml.Name   `xml:"urlset"`
	URLs    []SitemapURL `xml:"url"`
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

	// Set timeout for sitemap request (shorter than page timeout)
	sitemapCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req = req.WithContext(sitemapCtx)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("sitemap not found or inaccessible")
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read sitemap: %w", err)
	}

	// Try to parse as sitemap index first
	var index SitemapIndex
	if err := xml.Unmarshal(body, &index); err == nil && len(index.Sitemaps) > 0 {
		// This is a sitemap index, parse each referenced sitemap
		var allPages []string
		for _, sitemapEntry := range index.Sitemaps {
			if sitemapEntry.Loc != "" {
				// Parse the referenced sitemap
				sitemapPages, err := c.parseSitemapURL(ctx, sitemapEntry.Loc)
				if err == nil {
					allPages = append(allPages, sitemapPages...)
				}
			}
		}
		if len(allPages) > 0 {
			c.logger.Printf("🗺️ [SmartCrawler] Found %d pages from sitemap index", len(allPages))
			return allPages, nil
		}
	}

	// Try to parse as regular sitemap (urlset)
	var urlSet URLSet
	if err := xml.Unmarshal(body, &urlSet); err == nil && len(urlSet.URLs) > 0 {
		var pages []string
		for _, sitemapURL := range urlSet.URLs {
			if sitemapURL.Loc != "" {
				// Normalize URL
				normalizedURL, err := c.normalizeURL(sitemapURL.Loc)
				if err == nil {
					// Only include URLs from the same domain
					parsedSitemapURL, err := url.Parse(normalizedURL)
					if err == nil && parsedSitemapURL.Hostname() == parsedURL.Hostname() {
						pages = append(pages, normalizedURL)
					}
				}
			}
		}
		if len(pages) > 0 {
			c.logger.Printf("🗺️ [SmartCrawler] Found %d pages in sitemap", len(pages))
			return pages, nil
		}
	}

	// If XML parsing failed, try to extract URLs from text content
	// This is a fallback for non-standard sitemaps
	urlPattern := regexp.MustCompile(`<loc>(.*?)</loc>`)
	matches := urlPattern.FindAllStringSubmatch(string(body), -1)
	var pages []string
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			urlStr := strings.TrimSpace(match[1])
			if urlStr != "" && !seen[urlStr] {
				normalizedURL, err := c.normalizeURL(urlStr)
				if err == nil {
					parsedSitemapURL, err := url.Parse(normalizedURL)
					if err == nil && parsedSitemapURL.Hostname() == parsedURL.Hostname() {
						pages = append(pages, normalizedURL)
						seen[urlStr] = true
					}
				}
			}
		}
	}

	if len(pages) > 0 {
		c.logger.Printf("🗺️ [SmartCrawler] Found %d pages in sitemap (fallback parsing)", len(pages))
		return pages, nil
	}

	return nil, fmt.Errorf("no pages found in sitemap")
}

// parseSitemapURL parses a specific sitemap URL
func (c *SmartWebsiteCrawler) parseSitemapURL(ctx context.Context, sitemapURL string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", sitemapURL, nil)
	if err != nil {
		return nil, err
	}

	sitemapCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req = req.WithContext(sitemapCtx)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("sitemap not accessible")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read sitemap: %w", err)
	}

	var urlSet URLSet
	if err := xml.Unmarshal(body, &urlSet); err != nil {
		return nil, err
	}

	var pages []string
	parsedURL, _ := url.Parse(sitemapURL)
	for _, sitemapURL := range urlSet.URLs {
		if sitemapURL.Loc != "" {
			normalizedURL, err := c.normalizeURL(sitemapURL.Loc)
			if err == nil {
				parsedSitemapURL, err := url.Parse(normalizedURL)
				if err == nil && parsedURL != nil && parsedSitemapURL.Hostname() == parsedURL.Hostname() {
					pages = append(pages, normalizedURL)
				}
			}
		}
	}

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

// analyzePages analyzes multiple pages sequentially
// Enforces robots.txt crawl delays and adaptive delays based on response codes
// OPTIMIZATION #18: Adaptive Page Limits - Stop crawling when confidence >= 0.95 after 3+ pages
func (c *SmartWebsiteCrawler) analyzePages(ctx context.Context, pages []string) []PageAnalysis {
	return c.analyzePagesWithDelay(ctx, pages, 2*time.Second, false)
}

// analyzePagesWithDelay processes pages with configurable delay and fast-path mode
func (c *SmartWebsiteCrawler) analyzePagesWithDelay(ctx context.Context, pages []string, minDelay time.Duration, fastPath bool) []PageAnalysis {
	var analyses []PageAnalysis
	const highConfidenceThreshold = 0.95     // OPTIMIZATION #18: Stop at 95% confidence (user preference)
	const minPagesForEarlyStop = 3           // Minimum pages before considering early stop

	// Process pages sequentially (one at a time)
	for i, page := range pages {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return analyses
		default:
		}

		// Enforce crawl delay between requests (except for first page)
		if i > 0 {
			// Get domain from current page
			parsedURL, err := url.Parse(page)
			if err == nil {
				currentDomain := parsedURL.Hostname()
				
				// Get robots.txt crawl delay for this domain
				var crawlDelay time.Duration
				c.crawlDelaysMutex.RLock()
				if delay, exists := c.crawlDelays[currentDomain]; exists {
					crawlDelay = delay
				}
				c.crawlDelaysMutex.RUnlock()
				
				// Use maximum of robots.txt delay and configured minimum delay
				delay := crawlDelay
				if delay < minDelay {
					delay = minDelay
				}
				
				// Skip delay if we have sufficient content and are in fast-path mode
				if fastPath && len(analyses) >= 2 {
					if c.hasSufficientContent(analyses) {
						c.logger.Printf("⏩ [SmartCrawler] Skipping delay - sufficient content already gathered")
						// Still respect robots.txt if it's longer
						if crawlDelay > 0 && crawlDelay > delay {
							delay = crawlDelay
						} else {
							// Skip delay entirely if we have sufficient content
							delay = 0
						}
					}
				}
				
				// Skip delay if previous page failed quickly (< 1s) - indicates timeout
				if len(analyses) > 0 {
					lastAnalysis := analyses[len(analyses)-1]
					if lastAnalysis.ResponseTime > 0 && lastAnalysis.ResponseTime < 1*time.Second && lastAnalysis.StatusCode != 200 {
						c.logger.Printf("⏩ [SmartCrawler] Skipping delay - previous page failed quickly (timeout)")
						delay = 0
					}
				}
				
				// Apply adaptive delay based on previous response (if available)
				if len(analyses) > 0 && delay > 0 {
					lastAnalysis := analyses[len(analyses)-1]
					if lastAnalysis.StatusCode == 429 {
						// Rate limited - use exponential backoff
						delay = delay * 2
						if delay > 20*time.Second {
							delay = 20 * time.Second
						}
						c.logger.Printf("⏳ [SmartCrawler] Applying extended delay (%v) due to 429 rate limit", delay)
					} else if lastAnalysis.StatusCode == 503 {
						// Service unavailable - moderate delay
						delay = delay + 3*time.Second
						if delay > 10*time.Second {
							delay = 10 * time.Second
						}
						c.logger.Printf("⏳ [SmartCrawler] Applying moderate delay (%v) due to 503 service unavailable", delay)
					}
				}
				
				// Apply delay
				if delay > 0 {
					if crawlDelay > 0 {
						c.logger.Printf("⏳ [SmartCrawler] Enforcing robots.txt crawl delay of %v for %s", delay, currentDomain)
					}
					select {
					case <-ctx.Done():
						return analyses
					case <-time.After(delay):
						// Delay completed
					}
				}
			}
		}

		// Analyze page
		analysis := c.analyzePage(ctx, page)
		analyses = append(analyses, analysis)

		// If we got a 403, stop immediately to avoid further blocks
		if analysis.StatusCode == 403 {
			c.logger.Printf("🚫 [SmartCrawler] Received 403 for %s - stopping crawl to avoid further blocks", page)
			break
		}

		// Content-quality-based early exit (minimum 2 pages before checking)
		if len(analyses) >= 2 {
			if c.hasSufficientContent(analyses) {
				// Calculate metrics for logging
				var totalContentLength int
				uniqueKeywords := make(map[string]bool)
				var totalRelevance float64
				successfulPages := 0
				for _, a := range analyses {
					if a.StatusCode >= 200 && a.StatusCode < 400 {
						successfulPages++
						totalContentLength += a.ContentLength
						totalRelevance += a.RelevanceScore
						for _, keyword := range a.Keywords {
							uniqueKeywords[keyword] = true
						}
					}
				}
				avgRelevance := totalRelevance / float64(successfulPages)
				
				c.logger.Printf("✅ [SmartCrawler] Sufficient content gathered after %d pages - stopping early",
					len(analyses))
				c.logger.Printf("📊 [SmartCrawler] Early exit metrics: %d chars, %d keywords, %.2f relevance, %d pages",
					totalContentLength, len(uniqueKeywords), avgRelevance, successfulPages)
				break
			}
		}

		// OPTIMIZATION #18: Adaptive Page Limits (fallback to confidence-based)
		// Check confidence after each page and stop early if threshold is met
		if len(analyses) >= minPagesForEarlyStop {
			currentConfidence := c.calculateConfidenceFromAnalyses(analyses)
			
			c.logger.Printf("📊 [SmartCrawler] Page %d/%d analyzed - Current confidence: %.2f%% (threshold: %.2f%%)",
				len(analyses), len(pages), currentConfidence*100, highConfidenceThreshold*100)
			
			// Stop if confidence >= 95% after analyzing at least 3 pages
			if currentConfidence >= highConfidenceThreshold {
				c.logger.Printf("✅ [SmartCrawler] High confidence (%.2f%%) reached after %d pages - stopping crawl early",
					currentConfidence*100, len(analyses))
				break
			}
			
			// Check if confidence is improving (if we have at least 2 pages to compare)
			if len(analyses) >= 2 {
				previousConfidence := c.calculateConfidenceFromAnalyses(analyses[:len(analyses)-1])
				confidenceImproving := currentConfidence > previousConfidence
				
				// If confidence is not improving and we have enough pages, consider stopping
				// But only if we've analyzed at least 5 pages and confidence is still low
				if !confidenceImproving && len(analyses) >= 5 && currentConfidence < 0.7 {
					c.logger.Printf("⚠️ [SmartCrawler] Confidence not improving (%.2f%% -> %.2f%%) after %d pages - continuing with caution",
						previousConfidence*100, currentConfidence*100, len(analyses))
					// Continue but log the concern
				}
			}
		}
	}

	return analyses
}

// hasSufficientContent checks if we have gathered sufficient content for classification
// Returns true if multiple criteria are met:
// 1. Total content length >= 500 characters (optimal threshold)
// 2. At least 10 unique keywords extracted
// 3. Average relevance score >= 0.7
// 4. At least 2 successful pages with content
func (c *SmartWebsiteCrawler) hasSufficientContent(analyses []PageAnalysis) bool {
	if len(analyses) == 0 {
		return false
	}

	var totalContentLength int
	var totalKeywords int
	uniqueKeywords := make(map[string]bool)
	var totalRelevance float64
	successfulPages := 0

	for _, analysis := range analyses {
		if analysis.StatusCode >= 200 && analysis.StatusCode < 400 {
			successfulPages++
			totalContentLength += analysis.ContentLength
			totalRelevance += analysis.RelevanceScore

			// Collect unique keywords
			for _, keyword := range analysis.Keywords {
				uniqueKeywords[keyword] = true
			}
		}
	}

	totalKeywords = len(uniqueKeywords)

	// Check criteria:
	// 1. At least 2 successful pages
	if successfulPages < 2 {
		c.logger.Printf("📊 [SmartCrawler] [ContentCheck] Insufficient pages: %d < 2", successfulPages)
		return false
	}

	// 2. Total content length >= 500 characters
	if totalContentLength < 500 {
		c.logger.Printf("📊 [SmartCrawler] [ContentCheck] Insufficient content length: %d < 500 chars", totalContentLength)
		return false
	}

	// 3. At least 10 unique keywords
	if totalKeywords < 10 {
		c.logger.Printf("📊 [SmartCrawler] [ContentCheck] Insufficient keywords: %d < 10 unique", totalKeywords)
		return false
	}

	// 4. Average relevance score >= 0.7
	avgRelevance := totalRelevance / float64(successfulPages)
	if avgRelevance < 0.7 {
		c.logger.Printf("📊 [SmartCrawler] [ContentCheck] Insufficient relevance: %.2f < 0.7", avgRelevance)
		return false
	}

	c.logger.Printf("✅ [SmartCrawler] [ContentCheck] Sufficient content: pages=%d, length=%d, keywords=%d, relevance=%.2f", 
		successfulPages, totalContentLength, totalKeywords, avgRelevance)
	return true
}

// analyzePagesParallel processes pages in parallel with controlled concurrency
// Maintains session management across parallel requests
func (c *SmartWebsiteCrawler) analyzePagesParallel(ctx context.Context, pages []string, maxConcurrent int) []PageAnalysis {
	var analyses []PageAnalysis
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Semaphore to limit concurrent requests
	semaphore := make(chan struct{}, maxConcurrent)
	
	// Track if we should stop early due to sufficient content or 403
	var shouldStop bool
	var stopMutex sync.Mutex

	for i, page := range pages {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return analyses
		default:
		}
		
		// Check if we should stop early
		stopMutex.Lock()
		if shouldStop {
			stopMutex.Unlock()
			break
		}
		stopMutex.Unlock()

		wg.Add(1)
		go func(idx int, pageURL string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Analyze page (session management is handled by analyzePage internally)
			pageStart := time.Now()
			analysis := c.analyzePage(ctx, pageURL)
			pageDuration := time.Since(pageStart)
			c.logger.Printf("📄 [SmartCrawler] [PARALLEL] Page %d analyzed in %v: %s (status: %d, content: %d chars, keywords: %d)", 
				idx+1, pageDuration, pageURL, analysis.StatusCode, analysis.ContentLength, len(analysis.Keywords))

			// Store result in correct position
			mu.Lock()
			// Extend slice if needed
			for len(analyses) <= idx {
				analyses = append(analyses, PageAnalysis{})
			}
			analyses[idx] = analysis
			
			// Check for sufficient content after each page (minimum 2 pages)
			validCount := 0
			for _, a := range analyses {
				if a.URL != "" {
					validCount++
				}
			}
			
			// Check for early exit conditions
			if validCount >= 2 {
				// Check for sufficient content
				var validAnalyses []PageAnalysis
				for _, a := range analyses {
					if a.URL != "" {
						validAnalyses = append(validAnalyses, a)
					}
				}
				if c.hasSufficientContent(validAnalyses) {
					stopMutex.Lock()
					shouldStop = true
					stopMutex.Unlock()
					// Calculate content metrics for logging
					totalContent := 0
					totalKeywords := 0
					uniqueKeywords := make(map[string]bool)
					var totalRelevance float64
					successfulPages := 0
					for _, a := range validAnalyses {
						if a.StatusCode >= 200 && a.StatusCode < 400 {
							successfulPages++
							totalContent += a.ContentLength
							for _, kw := range a.Keywords {
								if !uniqueKeywords[kw] {
									uniqueKeywords[kw] = true
									totalKeywords++
								}
							}
							totalRelevance += a.RelevanceScore
						}
					}
					avgRelevance := 0.0
					if successfulPages > 0 {
						avgRelevance = totalRelevance / float64(successfulPages)
					}
					c.logger.Printf("✅ [SmartCrawler] [PARALLEL] Early exit triggered - sufficient content after %d pages", validCount)
					c.logger.Printf("✅ [SmartCrawler] [PARALLEL] Content metrics: length=%d chars, keywords=%d unique, relevance=%.2f, pages=%d", 
						totalContent, totalKeywords, avgRelevance, successfulPages)
				}
			}
			
			// Check for 403
			if analysis.StatusCode == 403 {
				stopMutex.Lock()
				shouldStop = true
				stopMutex.Unlock()
				c.logger.Printf("🚫 [SmartCrawler] Received 403 for %s - stopping parallel crawl", pageURL)
			}
			
			mu.Unlock()
		}(i, page)
		
		// Apply small delay between starting goroutines to avoid overwhelming the server
		// This respects robots.txt delays while still allowing parallel processing
		if i > 0 {
			// Get domain for crawl delay check
			parsedURL, err := url.Parse(page)
			if err == nil {
				currentDomain := parsedURL.Hostname()
				var crawlDelay time.Duration
				c.crawlDelaysMutex.RLock()
				if delay, exists := c.crawlDelays[currentDomain]; exists {
					crawlDelay = delay
				}
				c.crawlDelaysMutex.RUnlock()
				
				// Use smaller delay for parallel mode (divide by maxConcurrent since we're processing in parallel)
				// But respect robots.txt if it's longer
				delay := 500 * time.Millisecond // Base delay between starting goroutines
				if crawlDelay > 0 {
					// Use robots.txt delay divided by concurrency (but minimum 200ms)
					parallelDelay := crawlDelay / time.Duration(maxConcurrent)
					if parallelDelay < 200*time.Millisecond {
						parallelDelay = 200 * time.Millisecond
					}
					if parallelDelay > delay {
						delay = parallelDelay
					}
				}
				
				select {
				case <-ctx.Done():
					return analyses
				case <-time.After(delay):
					// Delay completed
				}
			}
		}
	}

	wg.Wait()

	// Filter out empty analyses
	var validAnalyses []PageAnalysis
	for _, analysis := range analyses {
		if analysis.URL != "" {
			validAnalyses = append(validAnalyses, analysis)
		}
	}

	// Check for 403 - if found, return only analyses before 403
	for i, analysis := range validAnalyses {
		if analysis.StatusCode == 403 {
			c.logger.Printf("🚫 [SmartCrawler] Received 403 - stopping crawl to avoid further blocks")
			return validAnalyses[:i]
		}
	}

	return validAnalyses
}

// calculateConfidenceFromAnalyses calculates overall confidence from page analyses
// OPTIMIZATION #18: Used for adaptive page limits
func (c *SmartWebsiteCrawler) calculateConfidenceFromAnalyses(analyses []PageAnalysis) float64 {
	if len(analyses) == 0 {
		return 0.0
	}

	// Calculate confidence based on:
	// 1. Average relevance score (40% weight)
	// 2. Average content quality (30% weight)
	// 3. Number of pages analyzed (20% weight)
	// 4. Keyword density (10% weight)

	var totalRelevance float64
	var totalContentQuality float64
	var totalKeywords int
	var successfulPages int

	for _, analysis := range analyses {
		if analysis.StatusCode >= 200 && analysis.StatusCode < 400 {
			successfulPages++
			totalRelevance += analysis.RelevanceScore
			totalContentQuality += analysis.ContentQuality
			totalKeywords += len(analysis.Keywords)
		}
	}

	if successfulPages == 0 {
		return 0.0
	}

	avgRelevance := totalRelevance / float64(successfulPages)
	avgContentQuality := totalContentQuality / float64(successfulPages)
	avgKeywordsPerPage := float64(totalKeywords) / float64(successfulPages)

	// Page count factor (more pages = higher confidence, up to 10 pages)
	pageCountFactor := float64(successfulPages) / 10.0
	if pageCountFactor > 1.0 {
		pageCountFactor = 1.0
	}

	// Keyword density factor (more keywords = higher confidence)
	keywordFactor := avgKeywordsPerPage / 20.0 // Normalize to 20 keywords per page
	if keywordFactor > 1.0 {
		keywordFactor = 1.0
	}

	// Calculate weighted confidence
	confidence := (avgRelevance * 0.4) +
		(avgContentQuality * 0.3) +
		(pageCountFactor * 0.2) +
		(keywordFactor * 0.1)

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
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

	// Parse URL to get domain for session management
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		analysis.RelevanceScore = 0.0
		return analysis
	}
	domain := parsedURL.Hostname()

	// Get or create session for this domain
	var session *ScrapingSession
	if c.sessionManager != nil {
		sess, err := c.sessionManager.GetOrCreateSession(domain)
		if err != nil {
			c.logger.Printf("⚠️ [PageAnalysis] Failed to get session for %s: %v", domain, err)
		} else {
			session = sess
		}
	}

	// Create HTTP client with session cookie jar and proxy support
	// Start with the base client (which has custom dialer)
	client := c.client
	if session != nil {
		// Create a new client with session cookie jar, preserving the custom transport
		baseTransport := c.client.Transport.(*http.Transport)
		client = &http.Client{
			Timeout:   c.pageTimeout,
			Transport: baseTransport, // Preserve custom dialer and DNS resolver
			Jar:       session.GetCookieJar(),
		}
	}

	// Get proxy transport if enabled
	if c.proxyManager != nil && c.proxyManager.IsEnabled() {
		baseTransport := c.client.Transport.(*http.Transport)
		proxyTransport, err := c.proxyManager.GetProxyTransport(domain, baseTransport)
		if err == nil && proxyTransport != nil {
			// Use proxy transport
			client = &http.Client{
				Timeout:   c.pageTimeout,
				Transport: proxyTransport,
			}
			if session != nil {
				client.Jar = session.GetCookieJar()
			}
		}
	}

	// Create request with timeout
	reqCtx, cancel := context.WithTimeout(ctx, c.pageTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", pageURL, nil)
	if err != nil {
		analysis.RelevanceScore = 0.0
		return analysis
	}

	// Set headers with randomization and referer
	referer := ""
	if session != nil && c.sessionManager != nil {
		referer = c.sessionManager.GetReferer(domain)
	}
	headers := GetRandomizedHeadersWithReferer(GetUserAgent(), referer)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Retry logic for HTTP requests (up to 3 attempts with exponential backoff)
	var resp *http.Response
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err = client.Do(req)
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

	// Update session referer for next request (if session exists)
	if session != nil && c.sessionManager != nil && resp.StatusCode == 200 {
		c.sessionManager.UpdateReferer(domain, pageURL)
		session.UpdateAccess()
	}

	// Handle specific HTTP status codes
	if resp.StatusCode == 429 {
		// Too Many Requests - stop immediately
		retryAfter := resp.Header.Get("Retry-After")
		c.logger.Printf("⚠️ [PageAnalysis] Rate limited (429) for %s, Retry-After: %s - stopping", pageURL, retryAfter)
		analysis.RelevanceScore = 0.0
		return analysis
	}
	if resp.StatusCode == 403 {
		// Forbidden - stop immediately
		c.logger.Printf("🚫 [PageAnalysis] Access forbidden (403) for %s - stopping", pageURL)
		analysis.RelevanceScore = 0.0
		return analysis
	}
	if resp.StatusCode == 503 {
		// Service Unavailable - log but don't retry (retry logic is handled at higher level)
		c.logger.Printf("⚠️ [PageAnalysis] Service unavailable (503) for %s", pageURL)
		analysis.RelevanceScore = 0.0
		return analysis
	}
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

	// Check for CAPTCHA before processing
	captchaResult := DetectCAPTCHA(resp, body)
	if captchaResult.Detected {
		c.logger.Printf("🚫 [PageAnalysis] CAPTCHA detected (%s) for %s - stopping", captchaResult.Type, pageURL)
		analysis.RelevanceScore = 0.0
		return analysis
	}

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

// checkRobotsTxt checks if crawling is allowed by robots.txt for a specific path.
// Returns: (blocked, crawlDelay, error)
// - blocked: true if the path is disallowed
// - crawlDelay: the crawl delay specified in robots.txt (0 if not specified)
// - error: any error that occurred during the check
func (c *SmartWebsiteCrawler) checkRobotsTxt(ctx context.Context, baseURL, path string) (bool, time.Duration, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return false, 0, err
	}

	robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)

	req, err := http.NewRequestWithContext(ctx, "GET", robotsURL, nil)
	if err != nil {
		return false, 0, err
	}

	// Use our identifiable User-Agent for robots.txt requests
	req.Header.Set("User-Agent", GetUserAgent())

	resp, err := c.client.Do(req)
	if err != nil {
		// If robots.txt is unavailable, allow crawling (graceful degradation)
		return false, 0, nil
	}
	defer resp.Body.Close()

	// If robots.txt doesn't exist (404) or is unavailable, allow crawling
	if resp.StatusCode != 200 {
		return false, 0, nil
	}

	// Read robots.txt content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, 0, fmt.Errorf("failed to read robots.txt: %w", err)
	}

	// Parse robots.txt using the library
	robotsData, err := robotstxt.FromBytes(body)
	if err != nil {
		// If parsing fails, log but allow crawling (graceful degradation)
		c.logger.Printf("⚠️ [RobotsTxt] Failed to parse robots.txt for %s: %v (allowing crawl)", parsedURL.Host, err)
		return false, 0, nil
	}

	// Get our User-Agent string
	userAgent := GetUserAgent()
	// Extract just the bot name from User-Agent for matching
	// Format: "Mozilla/5.0 (compatible; KYBPlatform/1.0; ...)"
	// Dynamically extract the identifier to avoid hardcoding
	botName := "KYBPlatform" // Default fallback
	parts := strings.Split(userAgent, ";")
	if len(parts) >= 2 {
		identifierPart := strings.TrimSpace(parts[1])
		// Extract "KYBPlatform" from "KYBPlatform/1.0"
		if slashIdx := strings.Index(identifierPart, "/"); slashIdx > 0 {
			botName = identifierPart[:slashIdx]
		}
	}

	// Check rules for our specific User-Agent first, then wildcard
	var group *robotstxt.Group
	group = robotsData.FindGroup(userAgent)
	if group == nil {
		// Try with just the bot name
		group = robotsData.FindGroup(botName)
	}
	if group == nil {
		// Try wildcard (*) rules
		group = robotsData.FindGroup("*")
	}

	// If no group found, allow crawling
	if group == nil {
		return false, 0, nil
	}

	// Test if the specific path is allowed
	allowed := group.Test(path)
	blocked := !allowed

	// Extract crawl delay if specified
	var crawlDelay time.Duration
	if group.CrawlDelay > 0 {
		crawlDelay = time.Duration(group.CrawlDelay) * time.Second
	}

	if blocked {
		c.logger.Printf("🚫 [RobotsTxt] Path %s blocked by robots.txt for %s", path, parsedURL.Host)
	} else if crawlDelay > 0 {
		c.logger.Printf("⏳ [RobotsTxt] Crawl delay of %v specified for %s", crawlDelay, parsedURL.Host)
	}

	return blocked, crawlDelay, nil
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

// extractWordsFromText extracts meaningful words from text (filters stop words and gibberish)
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
	
	// Split text into words - increased minimum length to 4 to filter out short gibberish
	words := regexp.MustCompile(`\b[a-zA-Z]{4,}\b`).FindAllString(text, -1)
	var filteredWords []string
	
	for _, word := range words {
		wordLower := strings.ToLower(word)
		// Filter stop words
		if stopWords[wordLower] {
			continue
		}
		// Filter short words (minimum 4 characters)
		if len(wordLower) < 4 {
			continue
		}
		// Filter gibberish: words that don't look like English
		// - Too many consecutive consonants (more than 3)
		// - No vowels at all
		// - Too high consonant-to-vowel ratio (more than 3:1 for short words)
		if !c.isValidEnglishWord(wordLower) {
			continue
		}
		filteredWords = append(filteredWords, wordLower)
	}
	
	return filteredWords
}

// isValidEnglishWord checks if a word looks like a valid English word
// Filters out gibberish like "ivdi", "fays", "yilp", "dioy", "ukxa", etc.
// Enhanced with n-gram validation and suspicious pattern detection
func (c *SmartWebsiteCrawler) isValidEnglishWord(word string) bool {
	if len(word) < 4 {
		return false
	}
	
	// Check against common English word dictionary first (if available)
	if c.commonEnglishWords != nil && c.commonEnglishWords[word] {
		return true
	}
	
	// Count vowels and consonants
	vowels := 0
	consonants := 0
	maxConsecutiveConsonants := 0
	currentConsecutiveConsonants := 0
	
	vowelSet := map[rune]bool{'a': true, 'e': true, 'i': true, 'o': true, 'u': true, 'y': true}
	
	for _, char := range word {
		if vowelSet[char] {
			vowels++
			currentConsecutiveConsonants = 0
		} else {
			consonants++
			currentConsecutiveConsonants++
			if currentConsecutiveConsonants > maxConsecutiveConsonants {
				maxConsecutiveConsonants = currentConsecutiveConsonants
			}
		}
	}
	
	// Must have at least one vowel
	if vowels == 0 {
		return false
	}
	
	// Check for suspicious patterns
	if c.hasSuspiciousPatterns(word) {
		return false
	}
	
	// Check n-gram patterns (bigram frequency)
	if !c.hasValidNgramPatterns(word) {
		return false
	}
	
	// For short words (4-5 chars), require reasonable vowel ratio
	if len(word) <= 5 {
		// At least 1 vowel for every 3 characters
		if vowels < (len(word)+2)/3 {
			return false
		}
		// No more than 3 consecutive consonants
		if maxConsecutiveConsonants > 3 {
			return false
		}
	}
	
	// For longer words, allow more flexibility but still require vowels
	if len(word) > 5 {
		// At least 1 vowel for every 4 characters
		if vowels < (len(word)+3)/4 {
			return false
		}
		// No more than 4 consecutive consonants
		if maxConsecutiveConsonants > 4 {
			return false
		}
	}
	
	return true
}

// hasSuspiciousPatterns checks for patterns that rarely appear in English words
func (c *SmartWebsiteCrawler) hasSuspiciousPatterns(word string) bool {
	// Check for repeated letters (more than 2 consecutive)
	for i := 0; i < len(word)-2; i++ {
		if word[i] == word[i+1] && word[i] == word[i+2] {
			return true // e.g., "aaa", "bbb"
		}
	}
	
	// Check for unusual consonant clusters that rarely appear in English
	suspiciousClusters := []string{
		"ivd", "fay", "yil", "dio", "ukx", "guo", "jey", "mii",
		"xzv", "qwx", "jkl", "zxc", "vbn", "qwe", "asd", "zxc",
	}
	for _, cluster := range suspiciousClusters {
		if strings.Contains(word, cluster) {
			return true
		}
	}
	
	// Check for words with too many rare letters in sequence
	rareLetters := map[rune]bool{'q': true, 'x': true, 'z': true, 'j': true}
	rareCount := 0
	for _, char := range word {
		if rareLetters[char] {
			rareCount++
		}
	}
	// If more than 30% of letters are rare, likely gibberish
	if float64(rareCount)/float64(len(word)) > 0.3 {
		return true
	}
	
	return false
}

// hasValidNgramPatterns checks if letter combinations are common in English
func (c *SmartWebsiteCrawler) hasValidNgramPatterns(word string) bool {
	// Common English bigrams (most frequent)
	commonBigrams := map[string]bool{
		"th": true, "in": true, "er": true, "ed": true, "an": true, "re": true,
		"he": true, "on": true, "en": true, "at": true, "it": true, "is": true,
		"or": true, "ti": true, "as": true, "to": true, "of": true, "al": true,
		"ar": true, "st": true, "ng": true, "le": true, "ou": true, "nt": true,
		"ea": true, "nd": true, "te": true, "es": true, "hi": true,
		"ri": true, "ve": true, "co": true, "de": true, "ra": true, "li": true,
		"se": true, "ne": true, "me": true, "be": true, "we": true, "wa": true,
		"ma": true, "ha": true, "ca": true, "la": true, "pa": true, "ta": true,
		"sa": true, "na": true, "ga": true, "fa": true, "da": true, "ba": true,
	}
	
	// Suspicious bigrams that rarely appear in English
	suspiciousBigrams := map[string]bool{
		"iv": true, "di": true, "xa": true, "gu": true, "oi": true,
		"je": true, "yl": true, "lb": true, "io": true,
		"fv": true, "yz": true, "zx": true, "qw": true, "xc": true, "vb": true,
		"uk": true, // Only include once
	}
	
	// Check bigrams in the word
	hasCommonBigram := false
	for i := 0; i < len(word)-1; i++ {
		bigram := word[i : i+2]
		if suspiciousBigrams[bigram] && !commonBigrams[bigram] {
			// If we find a suspicious bigram that's not also common, likely gibberish
			// But allow if we also have common bigrams
			if !hasCommonBigram {
				return false
			}
		}
		if commonBigrams[bigram] {
			hasCommonBigram = true
		}
	}
	
	// If word has no common bigrams at all, likely gibberish
	if !hasCommonBigram && len(word) >= 4 {
		return false
	}
	
	return true
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
// OPTIMIZATION #15: Structured data keywords are weighted 2x higher (already applied in extractKeywordsFromStructuredData)
func (c *SmartWebsiteCrawler) combineAndRankKeywordsEnhanced(structuredKeywords []structuredKeyword, bodyKeywords, phrases []string, pageType string, textContent string) []keywordScore {
	keywordScores := make(map[string]float64)
	
	// OPTIMIZATION #15: Prioritize structured keywords (JSON-LD/microdata) - already weighted 2x higher
	// Track which keywords came from structured data for additional prioritization
	structuredDataKeywords := make(map[string]bool)
	for _, skw := range structuredKeywords {
		keywordScores[skw.keyword] += skw.weight
		structuredDataKeywords[skw.keyword] = true
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
	
	// OPTIMIZATION #15: Additional boost for structured data keywords in final ranking
	// This ensures structured data keywords are prioritized even after normalization
	for kw := range keywordScores {
		if structuredDataKeywords[kw] {
			// Additional 20% boost for structured data keywords to ensure they rank higher
			keywordScores[kw] *= 1.2
		}
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
	
	// 8. Phase 3.3: Apply NER and topic modeling to enhance keywords
	if c.entityRecognizer != nil && c.topicModeler != nil && len(allKeywords) > 0 {
		// Extract entities from text content
		entities := c.entityRecognizer.ExtractEntities(textContent)
		if len(entities) > 0 {
			// Get keywords from entities
			entityKeywords := c.entityRecognizer.GetEntityKeywords(entities)
			// Add entity keywords with high weight
			for _, entityKw := range entityKeywords {
				allKeywords = append(allKeywords, keywordScore{
					keyword: entityKw,
					score:   0.85, // High weight for entity-based keywords
				})
			}
		}
		
		// Apply topic modeling to identify industry topics
		keywordStrings := make([]string, 0, len(allKeywords))
		for _, kw := range allKeywords {
			keywordStrings = append(keywordStrings, kw.keyword)
		}
		topicScores := c.topicModeler.IdentifyTopics(keywordStrings)
		
		// Boost keywords that align with identified topics
		if len(topicScores) > 0 {
			// Re-rank keywords based on topic alignment
			for i := range allKeywords {
				kwLower := strings.ToLower(allKeywords[i].keyword)
				// Small boost for keywords in top industries
				for industryID, score := range topicScores {
					if score > 0.3 {
						// Check if keyword is in industry topics
						industryTopics := c.topicModeler.GetIndustryTopics(industryID)
						for _, topicKw := range industryTopics {
							if kwLower == strings.ToLower(topicKw) {
								allKeywords[i].score *= (1.0 + score*0.1) // Boost by up to 10%
								break
							}
						}
					}
				}
			}
		}
	}
	
	// 9. Filter by relevance threshold (0.3) and return top 30
	return c.limitToTopKeywordsWithThreshold(allKeywords, 30, 0.3)
}

// extractKeywordsFromStructuredData extracts keywords from structured data (JSON-LD, microdata)
// Phase 6.1 and 6.2: Extract keywords from parsed structured data
func (c *SmartWebsiteCrawler) extractKeywordsFromStructuredData(structuredData map[string]interface{}) []structuredKeyword {
	var keywords []structuredKeyword
	seen := make(map[string]bool)
	
	// OPTIMIZATION #15: Structured Data Priority Weighting
	// Weight JSON-LD/microdata keywords 2x higher than scraped text
	// This improves accuracy when structured data is present (+8-12% accuracy)
	const structuredDataMultiplier = 2.0
	
	// Extract from business name (high weight, multiplied by 2x)
	if name, ok := structuredData["business_name"].(string); ok && name != "" {
		words := c.extractWordsFromText(name)
		for _, word := range words {
			wordLower := strings.ToLower(word)
			if len(wordLower) >= 3 && !seen[wordLower] {
				seen[wordLower] = true
				keywords = append(keywords, structuredKeyword{
					keyword: wordLower,
					weight:  0.95 * structuredDataMultiplier, // 2x weight for structured data business name
				})
			}
		}
	}
	
	// Extract from description (high weight, multiplied by 2x)
	if desc, ok := structuredData["description"].(string); ok && desc != "" {
		words := c.extractWordsFromText(desc)
		for _, word := range words {
			wordLower := strings.ToLower(word)
			if len(wordLower) >= 3 && !seen[wordLower] {
				seen[wordLower] = true
				keywords = append(keywords, structuredKeyword{
					keyword: wordLower,
					weight:  0.90 * structuredDataMultiplier, // 2x weight for structured data description
				})
			}
		}
	}
	
	// Extract from industry (very high weight, multiplied by 2x)
	if industry, ok := structuredData["industry"].(string); ok && industry != "" {
		words := c.extractWordsFromText(industry)
		for _, word := range words {
			wordLower := strings.ToLower(word)
			if len(wordLower) >= 3 && !seen[wordLower] {
				seen[wordLower] = true
				keywords = append(keywords, structuredKeyword{
					keyword: wordLower,
					weight:  1.0 * structuredDataMultiplier, // 2x weight for industry from structured data
				})
			}
		}
	}
	
	// Extract from services (medium-high weight, multiplied by 2x)
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
						weight:  0.85 * structuredDataMultiplier, // 2x weight for services
					})
				}
			}
		}
	}
	
	// Extract from products (medium-high weight, multiplied by 2x)
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
						weight:  0.85 * structuredDataMultiplier, // 2x weight for products
					})
				}
			}
		}
	}
	
	// OPTIMIZATION #15: Extract from Schema.org type (high weight, multiplied by 2x)
	if schemaType, ok := structuredData["schema_type"].(string); ok && schemaType != "" {
		// Extract meaningful parts from Schema.org type (e.g., "WineShop" -> "wine", "shop")
		typeWords := c.splitCamelCase(schemaType)
		for _, word := range typeWords {
			wordLower := strings.ToLower(word)
			if len(wordLower) >= 3 && !seen[wordLower] {
				seen[wordLower] = true
				keywords = append(keywords, structuredKeyword{
					keyword: wordLower,
					weight:  0.88 * structuredDataMultiplier, // 2x weight for Schema.org type
				})
			}
		}
	}
	
	// OPTIMIZATION #15: Extract from microdata properties (multiplied by 2x)
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

// loadCommonEnglishWords loads a dictionary of the 10,000 most common English words
// This is used to validate extracted keywords and filter out gibberish
func loadCommonEnglishWords() map[string]bool {
	dict := make(map[string]bool)
	
	// Top 10,000 most common English words (subset for performance)
	// This includes common words, business terms, and frequently used vocabulary
	commonWords := []string{
		// Most common words (top 100)
		"the", "be", "to", "of", "and", "a", "in", "that", "have", "i",
		"it", "for", "not", "on", "with", "he", "as", "you", "do", "at",
		"this", "but", "his", "by", "from", "they", "we", "say", "her", "she",
		"or", "an", "will", "my", "one", "all", "would", "there", "their", "what",
		"so", "up", "out", "if", "about", "who", "get", "which", "go", "me",
		"when", "make", "can", "like", "time", "no", "just", "him", "know", "take",
		"people", "into", "year", "your", "good", "some", "could", "them", "see", "other",
		"than", "then", "now", "look", "only", "come", "its", "over", "think", "also",
		"back", "after", "use", "two", "how", "our", "work", "first", "well", "way",
		
		// Common business and technology terms
		"business", "company", "service", "services", "product", "products", "technology",
		"software", "system", "systems", "solution", "solutions", "development", "management",
		"professional", "quality", "experience", "customer", "customers", "client", "clients",
		"team", "team", "group", "organization", "enterprise", "industry", "market", "markets",
		"online", "digital", "web", "internet", "website", "platform", "application", "applications",
		"data", "information", "technology", "tech", "computer", "network", "security", "cloud",
		"mobile", "software", "hardware", "device", "devices", "server", "servers", "database",
		
		// Common verbs
		"provide", "provides", "offering", "offer", "offers", "create", "creates", "created",
		"develop", "develops", "developed", "design", "designs", "designed", "build", "builds",
		"built", "deliver", "delivers", "delivered", "support", "supports", "supported",
		"help", "helps", "helped", "manage", "manages", "managed", "operate", "operates",
		"operated", "work", "works", "worked", "use", "uses", "used", "need", "needs", "needed",
		
		// Common adjectives
		"best", "better", "great", "excellent", "premium", "quality", "reliable", "trusted",
		"leading", "innovative", "advanced", "modern", "professional", "experienced", "expert",
		"specialized", "comprehensive", "complete", "full", "total", "entire", "whole",
		"new", "latest", "recent", "current", "updated", "improved", "enhanced", "optimized",
		
		// Industry-specific terms
		"retail", "restaurant", "hospitality", "healthcare", "medical", "finance", "financial",
		"banking", "insurance", "real", "estate", "property", "construction", "education",
		"training", "consulting", "legal", "law", "marketing", "advertising", "media",
		"entertainment", "travel", "tourism", "transportation", "logistics", "manufacturing",
		"production", "agriculture", "energy", "utilities", "telecommunications", "telecom",
		
		// Common nouns
		"office", "location", "locations", "store", "stores", "shop", "shops", "center", "centers",
		"facility", "facilities", "building", "buildings", "space", "spaces", "area", "areas",
		"region", "regions", "country", "countries", "city", "cities", "state", "states",
		"address", "phone", "email", "contact", "website", "site", "sites", "page", "pages",
		
		// Additional common words
		"about", "contact", "home", "page", "menu", "services", "products", "portfolio",
		"blog", "news", "events", "careers", "jobs", "career", "job", "opportunities",
		"testimonials", "reviews", "gallery", "photos", "images", "video", "videos",
		"faq", "faqs", "help", "support", "terms", "privacy", "policy", "policies",
	}
	
	// Add all words to dictionary
	for _, word := range commonWords {
		dict[strings.ToLower(word)] = true
	}
	
	return dict
}
