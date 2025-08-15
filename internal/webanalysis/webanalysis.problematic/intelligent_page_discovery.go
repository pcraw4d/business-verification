package webanalysis

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// PageDiscoveryResult represents the result of intelligent page discovery
type PageDiscoveryResult struct {
	URL               string            `json:"url"`
	RelevanceScore    float64           `json:"relevance_score"`
	PageType          PageType          `json:"page_type"`
	ContentIndicators []string          `json:"content_indicators"`
	BusinessKeywords  []string          `json:"business_keywords"`
	Priority          int               `json:"priority"`
	Depth             int               `json:"depth"`
	DiscoveredAt      time.Time         `json:"discovered_at"`
	Metadata          map[string]string `json:"metadata"`
	PriorityScore     float64           `json:"priority_score"`
	ContentQuality    float64           `json:"content_quality"`
}

// PageType represents the type of page discovered
type PageType string

const (
	PageTypeHome      PageType = "home"
	PageTypeAbout     PageType = "about"
	PageTypeMission   PageType = "mission"
	PageTypeProducts  PageType = "products"
	PageTypeServices  PageType = "services"
	PageTypeContact   PageType = "contact"
	PageTypeTeam      PageType = "team"
	PageTypeCompany   PageType = "company"
	PageTypeBusiness  PageType = "business"
	PageTypeProfile   PageType = "profile"
	PageTypeNews      PageType = "news"
	PageTypeBlog      PageType = "blog"
	PageTypeCareers   PageType = "careers"
	PageTypeInvestors PageType = "investors"
	PageTypeLegal     PageType = "legal"
	PageTypePrivacy   PageType = "privacy"
	PageTypeTerms     PageType = "terms"
	PageTypeUnknown   PageType = "unknown"
)

// IntelligentPageDiscovery manages intelligent page discovery and prioritization
type IntelligentPageDiscovery struct {
	scraper           *WebScraper
	pagePatterns      map[PageType][]*regexp.Regexp
	keywordPatterns   map[PageType][]string
	relevanceWeights  map[string]float64
	discoveryQueue    chan *PageDiscoveryJob
	results           map[string]*PageDiscoveryResult
	mu                sync.RWMutex
	config            DiscoveryConfig
	structureAnalyzer *WebsiteStructureAnalyzer
}

// DiscoveryConfig holds configuration for intelligent page discovery
type DiscoveryConfig struct {
	MaxDiscoveryDepth   int           `json:"max_discovery_depth"`
	MaxPagesPerDomain   int           `json:"max_pages_per_domain"`
	DiscoveryTimeout    time.Duration `json:"discovery_timeout"`
	ConcurrentDiscovery int           `json:"concurrent_discovery"`
	MinRelevanceScore   float64       `json:"min_relevance_score"`
	MaxDiscoveryRetries int           `json:"max_discovery_retries"`
	FollowInternalLinks bool          `json:"follow_internal_links"`
	RespectRobotsTxt    bool          `json:"respect_robots_txt"`
	ExcludePatterns     []string      `json:"exclude_patterns"`
	IncludePatterns     []string      `json:"include_patterns"`
}

// PageDiscoveryJob represents a page discovery task
type PageDiscoveryJob struct {
	URL       string            `json:"url"`
	Business  string            `json:"business"`
	Depth     int               `json:"depth"`
	ParentURL string            `json:"parent_url"`
	Context   context.Context   `json:"-"`
	Priority  int               `json:"priority"`
	Metadata  map[string]string `json:"metadata"`
}

// NewIntelligentPageDiscovery creates a new intelligent page discovery instance
func NewIntelligentPageDiscovery(scraper *WebScraper) *IntelligentPageDiscovery {
	ipd := &IntelligentPageDiscovery{
		scraper:           scraper,
		pagePatterns:      make(map[PageType][]*regexp.Regexp),
		keywordPatterns:   make(map[PageType][]string),
		relevanceWeights:  make(map[string]float64),
		discoveryQueue:    make(chan *PageDiscoveryJob, 1000),
		results:           make(map[string]*PageDiscoveryResult),
		structureAnalyzer: NewWebsiteStructureAnalyzer(),
		config: DiscoveryConfig{
			MaxDiscoveryDepth:   3,
			MaxPagesPerDomain:   50,
			DiscoveryTimeout:    time.Second * 30,
			ConcurrentDiscovery: 5,
			MinRelevanceScore:   0.3,
			MaxDiscoveryRetries: 3,
			FollowInternalLinks: true,
			RespectRobotsTxt:    true,
			ExcludePatterns: []string{
				`\.(pdf|doc|docx|xls|xlsx|ppt|pptx|zip|rar|tar|gz)$`,
				`\.(jpg|jpeg|png|gif|svg|ico|webp)$`,
				`\.(css|js|xml|json|txt)$`,
				`/admin/`,
				`/login/`,
				`/register/`,
				`/cart/`,
				`/checkout/`,
				`/api/`,
				`/ajax/`,
				`/search`,
				`/tag/`,
				`/category/`,
				`/page/\d+`,
				`/comment`,
				`/reply`,
				`/share`,
			},
			IncludePatterns: []string{
				`/about`,
				`/mission`,
				`/vision`,
				`/values`,
				`/team`,
				`/company`,
				`/business`,
				`/profile`,
				`/products`,
				`/services`,
				`/solutions`,
				`/contact`,
				`/careers`,
				`/investors`,
				`/news`,
				`/blog`,
				`/press`,
				`/media`,
			},
		},
	}

	// Initialize page patterns for different page types
	ipd.initializePagePatterns()
	ipd.initializeKeywordPatterns()
	ipd.initializeRelevanceWeights()

	return ipd
}

// initializePagePatterns sets up regex patterns for identifying page types
func (ipd *IntelligentPageDiscovery) initializePagePatterns() {
	ipd.pagePatterns[PageTypeAbout] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)/about`),
		regexp.MustCompile(`(?i)/about-us`),
		regexp.MustCompile(`(?i)/aboutus`),
		regexp.MustCompile(`(?i)/company/about`),
		regexp.MustCompile(`(?i)/who-we-are`),
	}

	ipd.pagePatterns[PageTypeMission] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)/mission`),
		regexp.MustCompile(`(?i)/vision`),
		regexp.MustCompile(`(?i)/values`),
		regexp.MustCompile(`(?i)/purpose`),
		regexp.MustCompile(`(?i)/philosophy`),
		regexp.MustCompile(`(?i)/mission-vision`),
	}

	ipd.pagePatterns[PageTypeProducts] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)/products`),
		regexp.MustCompile(`(?i)/solutions`),
		regexp.MustCompile(`(?i)/offerings`),
		regexp.MustCompile(`(?i)/portfolio`),
		regexp.MustCompile(`(?i)/catalog`),
		regexp.MustCompile(`(?i)/product/`),
	}

	ipd.pagePatterns[PageTypeServices] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)/services`),
		regexp.MustCompile(`(?i)/solutions`),
		regexp.MustCompile(`(?i)/offerings`),
		regexp.MustCompile(`(?i)/capabilities`),
		regexp.MustCompile(`(?i)/expertise`),
		regexp.MustCompile(`(?i)/service/`),
	}

	ipd.pagePatterns[PageTypeContact] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)/contact`),
		regexp.MustCompile(`(?i)/contact-us`),
		regexp.MustCompile(`(?i)/get-in-touch`),
		regexp.MustCompile(`(?i)/reach-us`),
		regexp.MustCompile(`(?i)/locations`),
		regexp.MustCompile(`(?i)/offices`),
	}

	ipd.pagePatterns[PageTypeTeam] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)/team`),
		regexp.MustCompile(`(?i)/leadership`),
		regexp.MustCompile(`(?i)/management`),
		regexp.MustCompile(`(?i)/executives`),
		regexp.MustCompile(`(?i)/people`),
		regexp.MustCompile(`(?i)/staff`),
	}

	ipd.pagePatterns[PageTypeCompany] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)/company`),
		regexp.MustCompile(`(?i)/corporate`),
		regexp.MustCompile(`(?i)/organization`),
		regexp.MustCompile(`(?i)/enterprise`),
		regexp.MustCompile(`(?i)/business`),
	}

	ipd.pagePatterns[PageTypeNews] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)/news`),
		regexp.MustCompile(`(?i)/press`),
		regexp.MustCompile(`(?i)/media`),
		regexp.MustCompile(`(?i)/announcements`),
		regexp.MustCompile(`(?i)/releases`),
	}

	ipd.pagePatterns[PageTypeCareers] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)/careers`),
		regexp.MustCompile(`(?i)/jobs`),
		regexp.MustCompile(`(?i)/employment`),
		regexp.MustCompile(`(?i)/work-with-us`),
		regexp.MustCompile(`(?i)/join-us`),
	}

	ipd.pagePatterns[PageTypeInvestors] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)/investors`),
		regexp.MustCompile(`(?i)/investor-relations`),
		regexp.MustCompile(`(?i)/shareholders`),
		regexp.MustCompile(`(?i)/financial`),
		regexp.MustCompile(`(?i)/ir`),
	}
}

// initializeKeywordPatterns sets up keyword patterns for page type detection
func (ipd *IntelligentPageDiscovery) initializeKeywordPatterns() {
	ipd.keywordPatterns[PageTypeAbout] = []string{
		"about us", "about", "company", "story", "history", "founded", "established",
		"who we are", "our story", "company history", "background",
	}

	ipd.keywordPatterns[PageTypeMission] = []string{
		"mission", "vision", "values", "purpose", "philosophy", "goals",
		"mission statement", "vision statement", "core values", "principles",
	}

	ipd.keywordPatterns[PageTypeProducts] = []string{
		"products", "solutions", "offerings", "portfolio", "catalog",
		"our products", "product line", "solutions", "services",
	}

	ipd.keywordPatterns[PageTypeServices] = []string{
		"services", "solutions", "offerings", "capabilities", "expertise",
		"our services", "service offerings", "what we do", "capabilities",
	}

	ipd.keywordPatterns[PageTypeContact] = []string{
		"contact", "contact us", "get in touch", "reach us", "locations",
		"address", "phone", "email", "office", "headquarters",
	}

	ipd.keywordPatterns[PageTypeTeam] = []string{
		"team", "leadership", "management", "executives", "people",
		"our team", "leadership team", "management team", "staff",
	}

	ipd.keywordPatterns[PageTypeCompany] = []string{
		"company", "corporate", "organization", "enterprise", "business",
		"our company", "corporate information", "organization",
	}

	ipd.keywordPatterns[PageTypeNews] = []string{
		"news", "press", "media", "announcements", "releases",
		"latest news", "press releases", "media coverage", "updates",
	}

	ipd.keywordPatterns[PageTypeCareers] = []string{
		"careers", "jobs", "employment", "work with us", "join us",
		"career opportunities", "job openings", "employment opportunities",
	}

	ipd.keywordPatterns[PageTypeInvestors] = []string{
		"investors", "investor relations", "shareholders", "financial",
		"investor information", "financial reports", "shareholder information",
	}
}

// initializeRelevanceWeights sets up weights for different relevance factors
func (ipd *IntelligentPageDiscovery) initializeRelevanceWeights() {
	ipd.relevanceWeights = map[string]float64{
		"page_type_match":     0.4,
		"keyword_density":     0.25,
		"business_name_match": 0.2,
		"content_quality":     0.1,
		"page_depth":          0.05,
	}
}

// DiscoverPages performs intelligent page discovery for a given website
func (ipd *IntelligentPageDiscovery) DiscoverPages(ctx context.Context, baseURL, business string) ([]*PageDiscoveryResult, error) {
	log.Printf("Starting intelligent page discovery for %s (business: %s)", baseURL, business)

	// Validate and normalize the base URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// Start discovery workers
	workerCtx, cancel := context.WithTimeout(ctx, ipd.config.DiscoveryTimeout)
	defer cancel()

	// Start discovery workers
	var wg sync.WaitGroup
	for i := 0; i < ipd.config.ConcurrentDiscovery; i++ {
		wg.Add(1)
		go ipd.discoveryWorker(workerCtx, &wg)
	}

	// Add initial discovery job
	initialJob := &PageDiscoveryJob{
		URL:       baseURL,
		Business:  business,
		Depth:     0,
		ParentURL: "",
		Context:   workerCtx,
		Priority:  100,
		Metadata: map[string]string{
			"domain": parsedURL.Host,
		},
	}

	ipd.discoveryQueue <- initialJob

	// Wait for discovery to complete
	go func() {
		wg.Wait()
		close(ipd.discoveryQueue)
	}()

	// Collect results
	var results []*PageDiscoveryResult
	ipd.mu.RLock()
	for _, result := range ipd.results {
		results = append(results, result)
	}
	ipd.mu.RUnlock()

	// Sort results by priority and relevance score
	ipd.sortResultsByPriority(results)

	log.Printf("Page discovery completed. Found %d relevant pages", len(results))
	return results, nil
}

// discoveryWorker processes page discovery jobs
func (ipd *IntelligentPageDiscovery) discoveryWorker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case job, ok := <-ipd.discoveryQueue:
			if !ok {
				return
			}
			ipd.processDiscoveryJob(ctx, job)
		case <-ctx.Done():
			return
		}
	}
}

// processDiscoveryJob processes a single page discovery job
func (ipd *IntelligentPageDiscovery) processDiscoveryJob(ctx context.Context, job *PageDiscoveryJob) {
	// Check if we've already processed this URL
	ipd.mu.Lock()
	if _, exists := ipd.results[job.URL]; exists {
		ipd.mu.Unlock()
		return
	}
	ipd.mu.Unlock()

	// Check depth limits
	if job.Depth > ipd.config.MaxDiscoveryDepth {
		return
	}

	// Check if URL should be excluded
	if ipd.shouldExcludeURL(job.URL) {
		return
	}

	// Scrape the page
	scrapingJob := &ScrapingJob{
		URL:        job.URL,
		Business:   job.Business,
		Priority:   job.Priority,
		Timeout:    ipd.config.DiscoveryTimeout,
		MaxRetries: ipd.config.MaxDiscoveryRetries,
	}
	content, err := ipd.scraper.ScrapeWebsite(scrapingJob)
	if err != nil {
		log.Printf("Failed to scrape %s: %v", job.URL, err)
		return
	}

	// Analyze the page
	result := ipd.analyzePage(job, content)

	// Check minimum relevance score
	if result.RelevanceScore < ipd.config.MinRelevanceScore {
		return
	}

	// Store the result
	ipd.mu.Lock()
	ipd.results[job.URL] = result
	ipd.mu.Unlock()

	// Discover additional pages if following internal links
	if ipd.config.FollowInternalLinks && job.Depth < ipd.config.MaxDiscoveryDepth {
		ipd.discoverInternalLinks(ctx, job, content)
	}
}

// analyzePage analyzes a scraped page to determine its type and relevance
func (ipd *IntelligentPageDiscovery) analyzePage(job *PageDiscoveryJob, content *ScrapedContent) *PageDiscoveryResult {
	result := &PageDiscoveryResult{
		URL:               job.URL,
		RelevanceScore:    0.0,
		PageType:          PageTypeUnknown,
		ContentIndicators: []string{},
		BusinessKeywords:  []string{},
		Priority:          job.Priority,
		Depth:             job.Depth,
		DiscoveredAt:      time.Now(),
		Metadata:          job.Metadata,
		PriorityScore:     0.0,
		ContentQuality:    0.0,
	}

	// Determine page type based on URL patterns
	result.PageType = ipd.determinePageType(job.URL, content)

	// Calculate relevance score
	result.RelevanceScore = ipd.calculateRelevanceScore(job, content, result.PageType)

	// Extract content indicators
	result.ContentIndicators = ipd.extractContentIndicators(content)

	// Extract business keywords
	result.BusinessKeywords = ipd.extractBusinessKeywords(content, job.Business)

	// Calculate priority score and content quality
	result.PriorityScore = ipd.calculatePriorityScore(result.PageType)
	result.ContentQuality = ipd.calculateContentQuality(content)

	// Adjust priority based on page type and relevance
	result.Priority = ipd.calculatePriority(result)

	return result
}

// determinePageType determines the type of page based on URL and content
func (ipd *IntelligentPageDiscovery) determinePageType(url string, content *ScrapedContent) PageType {
	// Check URL patterns first
	for pageType, patterns := range ipd.pagePatterns {
		for _, pattern := range patterns {
			if pattern.MatchString(url) {
				return pageType
			}
		}
	}

	// Check content keywords if URL patterns don't match
	text := strings.ToLower(content.Text + " " + content.Title)
	for pageType, keywords := range ipd.keywordPatterns {
		for _, keyword := range keywords {
			if strings.Contains(text, strings.ToLower(keyword)) {
				return pageType
			}
		}
	}

	return PageTypeUnknown
}

// calculateRelevanceScore calculates the relevance score for a page
func (ipd *IntelligentPageDiscovery) calculateRelevanceScore(job *PageDiscoveryJob, content *ScrapedContent, pageType PageType) float64 {
	score := 0.0

	// Page type match score (enhanced for business information pages)
	if pageType != PageTypeUnknown {
		baseScore := ipd.relevanceWeights["page_type_match"]

		// Boost score for high-value page types
		switch pageType {
		case PageTypeAbout, PageTypeMission:
			baseScore *= 1.5
		case PageTypeProducts, PageTypeServices:
			baseScore *= 1.3
		case PageTypeContact, PageTypeTeam:
			baseScore *= 1.2
		case PageTypeCompany, PageTypeBusiness:
			baseScore *= 1.4
		}

		score += baseScore
	}

	// Enhanced keyword density score
	keywordDensity := ipd.calculateKeywordDensity(content, job.Business)
	score += keywordDensity * ipd.relevanceWeights["keyword_density"]

	// Enhanced business name match score
	businessMatch := ipd.calculateBusinessNameMatch(content, job.Business)
	score += businessMatch * ipd.relevanceWeights["business_name_match"]

	// Enhanced content quality score
	contentQuality := ipd.calculateContentQuality(content)
	score += contentQuality * ipd.relevanceWeights["content_quality"]

	// Page depth penalty (reduced for high-priority pages)
	depthPenalty := float64(job.Depth) * ipd.relevanceWeights["page_depth"]
	if pageType == PageTypeAbout || pageType == PageTypeMission || pageType == PageTypeServices || pageType == PageTypeProducts {
		depthPenalty *= 0.5 // Less penalty for high-priority pages
	}
	score -= depthPenalty

	// Business name validation bonus
	if ipd.validateBusinessNameMatch(content, job.Business) {
		score += 0.1 // Bonus for validated business name match
	}

	return score
}

// calculateKeywordDensity calculates the density of business-related keywords
func (ipd *IntelligentPageDiscovery) calculateKeywordDensity(content *ScrapedContent, business string) float64 {
	text := strings.ToLower(content.Text + " " + content.Title)
	words := strings.Fields(text)
	if len(words) == 0 {
		return 0.0
	}

	// Business name keywords
	businessWords := strings.Fields(strings.ToLower(business))
	keywordCount := 0

	for _, word := range businessWords {
		if len(word) > 2 { // Skip very short words
			keywordCount += strings.Count(text, word)
		}
	}

	// Industry-related keywords
	industryKeywords := []string{
		"company", "business", "enterprise", "organization", "corporation",
		"services", "products", "solutions", "offerings", "capabilities",
		"industry", "sector", "market", "clients", "customers",
	}

	for _, keyword := range industryKeywords {
		keywordCount += strings.Count(text, keyword)
	}

	return float64(keywordCount) / float64(len(words))
}

// calculateBusinessNameMatch calculates how well the business name matches the content
func (ipd *IntelligentPageDiscovery) calculateBusinessNameMatch(content *ScrapedContent, business string) float64 {
	text := strings.ToLower(content.Text + " " + content.Title)
	businessLower := strings.ToLower(business)

	// Exact match
	if strings.Contains(text, businessLower) {
		return 1.0
	}

	// Partial match (check for key words from business name)
	businessWords := strings.Fields(businessLower)
	matchedWords := 0

	for _, word := range businessWords {
		if len(word) > 2 && strings.Contains(text, word) {
			matchedWords++
		}
	}

	if len(businessWords) == 0 {
		return 0.0
	}

	return float64(matchedWords) / float64(len(businessWords))
}

// validateBusinessNameMatch validates if the page content matches the business name
func (ipd *IntelligentPageDiscovery) validateBusinessNameMatch(content *ScrapedContent, business string) bool {
	text := strings.ToLower(content.Text + " " + content.Title)
	businessLower := strings.ToLower(business)
	businessWords := strings.Fields(businessLower)

	// Check for exact business name match
	if strings.Contains(text, businessLower) {
		return true
	}

	// Check for key business words (at least 50% match)
	matchedWords := 0
	for _, word := range businessWords {
		if len(word) > 2 && strings.Contains(text, word) {
			matchedWords++
		}
	}

	if len(businessWords) == 0 {
		return false
	}

	matchRatio := float64(matchedWords) / float64(len(businessWords))
	return matchRatio >= 0.5
}

// calculateContentQuality calculates the quality of the page content
func (ipd *IntelligentPageDiscovery) calculateContentQuality(content *ScrapedContent) float64 {
	text := content.Text
	if len(text) == 0 {
		return 0.0
	}

	score := 0.0

	// Length score (prefer pages with substantial content)
	if len(text) > 1000 {
		score += 0.3
	} else if len(text) > 500 {
		score += 0.2
	} else if len(text) > 100 {
		score += 0.1
	}

	// Title quality
	if len(content.Title) > 10 {
		score += 0.2
	}

	// HTML structure quality (presence of meaningful tags)
	if strings.Contains(content.HTML, "<h1>") || strings.Contains(content.HTML, "<h2>") {
		score += 0.2
	}

	if strings.Contains(content.HTML, "<p>") {
		score += 0.2
	}

	// Avoid pages that are mostly navigation or ads
	if strings.Count(text, "menu") > 5 || strings.Count(text, "advertisement") > 3 {
		score -= 0.3
	}

	return score
}

// calculatePriorityScore calculates the priority score for a page type
func (ipd *IntelligentPageDiscovery) calculatePriorityScore(pageType PageType) float64 {
	priorities := map[PageType]float64{
		PageTypeAbout:    0.9,  // High priority - essential business information
		PageTypeMission:  0.8,  // High priority - business purpose and values
		PageTypeServices: 0.85, // High priority - what the business does
		PageTypeProducts: 0.8,  // High priority - what the business sells
		PageTypeContact:  0.7,  // Medium-high priority - contact information
		PageTypeTeam:     0.6,  // Medium priority - team information
		PageTypeCompany:  0.75, // High priority - company information
		PageTypeBusiness: 0.75, // High priority - business information
		PageTypeNews:     0.3,  // Lower priority - news and updates
		PageTypeBlog:     0.3,  // Lower priority - blog posts
		PageTypeCareers:  0.4,  // Lower priority - job opportunities
		PageTypeUnknown:  0.1,  // Lowest priority - unknown page type
	}

	if priority, exists := priorities[pageType]; exists {
		return priority
	}
	return 0.1
}

// calculatePriority calculates the priority for a discovered page
func (ipd *IntelligentPageDiscovery) calculatePriority(result *PageDiscoveryResult) int {
	priority := result.Priority

	// Boost priority for high-value page types
	switch result.PageType {
	case PageTypeAbout, PageTypeMission:
		priority += 50
	case PageTypeProducts, PageTypeServices:
		priority += 40
	case PageTypeContact, PageTypeTeam:
		priority += 30
	case PageTypeCompany, PageTypeBusiness:
		priority += 35
	case PageTypeNews, PageTypeBlog:
		priority += 20
	}

	// Boost priority for high relevance scores
	if result.RelevanceScore > 0.8 {
		priority += 30
	} else if result.RelevanceScore > 0.6 {
		priority += 20
	} else if result.RelevanceScore > 0.4 {
		priority += 10
	}

	// Reduce priority for deeper pages
	priority -= result.Depth * 5

	return priority
}

// extractContentIndicators extracts indicators about the page content
func (ipd *IntelligentPageDiscovery) extractContentIndicators(content *ScrapedContent) []string {
	indicators := []string{}
	text := strings.ToLower(content.Text + " " + content.Title)

	// Check for various content indicators
	if strings.Contains(text, "about") || strings.Contains(text, "company") {
		indicators = append(indicators, "company_information")
	}

	if strings.Contains(text, "mission") || strings.Contains(text, "vision") {
		indicators = append(indicators, "mission_statement")
	}

	if strings.Contains(text, "products") || strings.Contains(text, "services") {
		indicators = append(indicators, "products_services")
	}

	if strings.Contains(text, "contact") || strings.Contains(text, "address") {
		indicators = append(indicators, "contact_information")
	}

	if strings.Contains(text, "team") || strings.Contains(text, "leadership") {
		indicators = append(indicators, "team_information")
	}

	if strings.Contains(text, "news") || strings.Contains(text, "press") {
		indicators = append(indicators, "news_press")
	}

	return indicators
}

// extractBusinessKeywords extracts business-related keywords from content
func (ipd *IntelligentPageDiscovery) extractBusinessKeywords(content *ScrapedContent, business string) []string {
	keywords := []string{}
	text := strings.ToLower(content.Text + " " + content.Title)

	// Add business name words
	businessWords := strings.Fields(strings.ToLower(business))
	keywords = append(keywords, businessWords...)

	// Add industry-related keywords found in content
	industryKeywords := []string{
		"company", "business", "enterprise", "organization", "corporation",
		"services", "products", "solutions", "offerings", "capabilities",
		"industry", "sector", "market", "clients", "customers",
	}

	for _, keyword := range industryKeywords {
		if strings.Contains(text, keyword) {
			keywords = append(keywords, keyword)
		}
	}

	return keywords
}

// discoverInternalLinks discovers internal links from a page
func (ipd *IntelligentPageDiscovery) discoverInternalLinks(ctx context.Context, job *PageDiscoveryJob, content *ScrapedContent) {
	links := ipd.extractInternalLinks(job.URL, content.HTML)

	for _, link := range links {
		// Check if we should include this link
		if !ipd.shouldIncludeURL(link) {
			continue
		}

		// Create new discovery job
		newJob := &PageDiscoveryJob{
			URL:       link,
			Business:  job.Business,
			Depth:     job.Depth + 1,
			ParentURL: job.URL,
			Context:   ctx,
			Priority:  job.Priority - 10, // Lower priority for deeper pages
			Metadata:  job.Metadata,
		}

		// Add to discovery queue
		select {
		case ipd.discoveryQueue <- newJob:
		default:
			// Queue is full, skip this link
		}
	}
}

// extractInternalLinks extracts internal links from HTML content
func (ipd *IntelligentPageDiscovery) extractInternalLinks(baseURL string, htmlContent string) []string {
	var links []string
	baseParsed, err := url.Parse(baseURL)
	if err != nil {
		return links
	}

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return links
	}

	var extractLinks func(*html.Node)
	extractLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href := attr.Val

					// Parse the href
					parsed, err := url.Parse(href)
					if err != nil {
						continue
					}

					// Resolve relative URLs
					if !parsed.IsAbs() {
						parsed = baseParsed.ResolveReference(parsed)
					}

					// Check if it's an internal link
					if parsed.Host == baseParsed.Host {
						links = append(links, parsed.String())
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractLinks(c)
		}
	}

	extractLinks(doc)
	return links
}

// shouldExcludeURL checks if a URL should be excluded from discovery
func (ipd *IntelligentPageDiscovery) shouldExcludeURL(url string) bool {
	for _, pattern := range ipd.config.ExcludePatterns {
		matched, _ := regexp.MatchString(pattern, url)
		if matched {
			return true
		}
	}
	return false
}

// shouldIncludeURL checks if a URL should be included in discovery
func (ipd *IntelligentPageDiscovery) shouldIncludeURL(url string) bool {
	// If no include patterns, include all non-excluded URLs
	if len(ipd.config.IncludePatterns) == 0 {
		return !ipd.shouldExcludeURL(url)
	}

	// Check include patterns
	for _, pattern := range ipd.config.IncludePatterns {
		matched, _ := regexp.MatchString(pattern, url)
		if matched {
			return true
		}
	}
	return false
}

// sortResultsByPriority sorts discovery results by priority and relevance
func (ipd *IntelligentPageDiscovery) sortResultsByPriority(results []*PageDiscoveryResult) {
	// Implementation would use sort.Slice with custom comparison
	// For now, we'll use a simple bubble sort for demonstration
	for i := 0; i < len(results)-1; i++ {
		for j := 0; j < len(results)-i-1; j++ {
			if results[j].Priority < results[j+1].Priority {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}

// GetDiscoveryStats returns statistics about the discovery process
func (ipd *IntelligentPageDiscovery) GetDiscoveryStats() map[string]interface{} {
	ipd.mu.RLock()
	defer ipd.mu.RUnlock()

	stats := map[string]interface{}{
		"total_pages_discovered": len(ipd.results),
		"page_types":             make(map[PageType]int),
		"average_relevance":      0.0,
		"max_relevance":          0.0,
		"min_relevance":          1.0,
	}

	totalRelevance := 0.0
	for _, result := range ipd.results {
		stats["page_types"].(map[PageType]int)[result.PageType]++
		totalRelevance += result.RelevanceScore

		if result.RelevanceScore > stats["max_relevance"].(float64) {
			stats["max_relevance"] = result.RelevanceScore
		}
		if result.RelevanceScore < stats["min_relevance"].(float64) {
			stats["min_relevance"] = result.RelevanceScore
		}
	}

	if len(ipd.results) > 0 {
		stats["average_relevance"] = totalRelevance / float64(len(ipd.results))
	}

	return stats
}

// DiscoverPagesWithStructureAnalysis performs page discovery with structure analysis integration
func (ipd *IntelligentPageDiscovery) DiscoverPagesWithStructureAnalysis(ctx context.Context, website string, business string, maxPages int) ([]*PageDiscoveryResult, *StructureAnalysisResult, error) {
	// Step 1: Perform regular page discovery
	discoveryResults, err := ipd.DiscoverPages(ctx, website, business)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to discover pages: %w", err)
	}

	// Step 2: Convert discovery results to scraped content for structure analysis
	var scrapedPages []*ScrapedContent
	for _, result := range discoveryResults {
		// Create a placeholder scraped content (in production, this would be actual scraped content)
		scrapedContent := &ScrapedContent{
			URL:   result.URL,
			Title: string(result.PageType),
			Text:  strings.Join(result.ContentIndicators, " "),
			HTML:  "<html><body>" + strings.Join(result.ContentIndicators, " ") + "</body></html>",
		}
		scrapedPages = append(scrapedPages, scrapedContent)
	}

	// Step 3: Perform structure analysis
	structureAnalysis, err := ipd.structureAnalyzer.AnalyzeWebsiteStructure(ctx, website, scrapedPages)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to analyze structure: %w", err)
	}

	// Step 4: Enhance discovery results with structure-based scoring
	enhancedResults := ipd.enhanceResultsWithStructureAnalysis(discoveryResults, structureAnalysis)

	return enhancedResults, structureAnalysis, nil
}

// enhanceResultsWithStructureAnalysis enhances page discovery results with structure analysis
func (ipd *IntelligentPageDiscovery) enhanceResultsWithStructureAnalysis(results []*PageDiscoveryResult, structureAnalysis *StructureAnalysisResult) []*PageDiscoveryResult {
	var enhancedResults []*PageDiscoveryResult

	for _, result := range results {
		enhancedResult := *result // Create a copy

		// Add structure-based priority scoring
		enhancedResult.PriorityScore = ipd.calculateStructureBasedPriority(result, structureAnalysis)

		// Add structure-based relevance scoring
		enhancedResult.RelevanceScore = ipd.calculateStructureBasedRelevance(result, structureAnalysis)

		// Add structure-based content quality
		enhancedResult.ContentQuality = ipd.calculateStructureBasedContentQuality(result, structureAnalysis)

		// Add structure metadata
		if enhancedResult.Metadata == nil {
			enhancedResult.Metadata = make(map[string]string)
		}
		enhancedResult.Metadata["structure_quality"] = fmt.Sprintf("%.2f", structureAnalysis.StructureRelevance.StructureQuality)
		enhancedResult.Metadata["navigation_relevance"] = fmt.Sprintf("%.2f", structureAnalysis.StructureRelevance.NavigationRelevance)
		enhancedResult.Metadata["content_relevance"] = fmt.Sprintf("%.2f", structureAnalysis.StructureRelevance.ContentRelevance)
		enhancedResult.Metadata["business_relevance"] = fmt.Sprintf("%.2f", structureAnalysis.StructureRelevance.BusinessRelevance)

		enhancedResults = append(enhancedResults, &enhancedResult)
	}

	return enhancedResults
}

// calculateStructureBasedPriority calculates structure-based priority for a page
func (ipd *IntelligentPageDiscovery) calculateStructureBasedPriority(result *PageDiscoveryResult, structureAnalysis *StructureAnalysisResult) float64 {
	priority := result.PriorityScore

	// Boost priority based on navigation structure
	if structureAnalysis.NavigationStructure != nil {
		// Higher priority for pages found in main navigation
		for _, navItem := range structureAnalysis.NavigationStructure.MainNavigation {
			if strings.Contains(strings.ToLower(navItem.URL), strings.ToLower(result.URL)) {
				priority += 0.2
				break
			}
		}

		// Higher priority for pages with high navigation relevance
		for _, navItem := range structureAnalysis.NavigationStructure.MainNavigation {
			if strings.Contains(strings.ToLower(navItem.URL), strings.ToLower(result.URL)) {
				priority += navItem.Relevance * 0.1
			}
		}
	}

	// Boost priority based on structure quality
	priority += structureAnalysis.StructureRelevance.StructureQuality * 0.1

	// Boost priority based on business information extraction
	if structureAnalysis.BusinessInformation != nil {
		priority += structureAnalysis.BusinessInformation.ExtractionScore * 0.1
	}

	return math.Min(priority, 1.0)
}

// calculateStructureBasedRelevance calculates structure-based relevance for a page
func (ipd *IntelligentPageDiscovery) calculateStructureBasedRelevance(result *PageDiscoveryResult, structureAnalysis *StructureAnalysisResult) float64 {
	relevance := result.RelevanceScore

	// Boost relevance based on content aggregation
	if structureAnalysis.ContentAggregation != nil {
		// Higher relevance for pages that contribute to content aggregation
		pageType := strings.ToLower(string(result.PageType))
		if strings.Contains(strings.ToLower(structureAnalysis.ContentAggregation.AboutContent), pageType) {
			relevance += 0.15
		}
		if strings.Contains(strings.ToLower(structureAnalysis.ContentAggregation.ServicesContent), pageType) {
			relevance += 0.15
		}
		if strings.Contains(strings.ToLower(structureAnalysis.ContentAggregation.ContactContent), pageType) {
			relevance += 0.15
		}
	}

	// Boost relevance based on overall structure relevance
	relevance += structureAnalysis.StructureRelevance.OverallRelevance * 0.2

	return math.Min(relevance, 1.0)
}

// calculateStructureBasedContentQuality calculates structure-based content quality for a page
func (ipd *IntelligentPageDiscovery) calculateStructureBasedContentQuality(result *PageDiscoveryResult, structureAnalysis *StructureAnalysisResult) float64 {
	quality := result.ContentQuality

	// Boost quality based on content aggregation quality
	if structureAnalysis.ContentAggregation != nil {
		quality += structureAnalysis.ContentAggregation.ContentQuality * 0.2
	}

	// Boost quality based on business information completeness
	if structureAnalysis.BusinessInformation != nil {
		quality += structureAnalysis.BusinessInformation.ExtractionScore * 0.1
	}

	// Boost quality based on structure quality
	quality += structureAnalysis.StructureRelevance.StructureQuality * 0.1

	return math.Min(quality, 1.0)
}

// AggregateContentFromMultiplePages aggregates content from multiple discovered pages
func (ipd *IntelligentPageDiscovery) AggregateContentFromMultiplePages(results []*PageDiscoveryResult) *ContentAggregation {
	// Convert page discovery results to scraped content
	var scrapedPages []*ScrapedContent
	for _, result := range results {
		scrapedContent := &ScrapedContent{
			URL:   result.URL,
			Title: string(result.PageType),
			Text:  strings.Join(result.ContentIndicators, " "),
			HTML:  "<html><body>" + strings.Join(result.ContentIndicators, " ") + "</body></html>",
		}
		scrapedPages = append(scrapedPages, scrapedContent)
	}

	// Use the content aggregator to aggregate content
	aggregator := NewContentAggregator()
	aggregation, err := aggregator.AggregateContent(scrapedPages)
	if err != nil {
		// Return minimal aggregation on error
		return &ContentAggregation{
			PageCount:      len(scrapedPages),
			ContentQuality: 0.1,
		}
	}

	return aggregation
}

// CreateStructureBasedClassificationConfidence creates structure-based classification confidence
func (ipd *IntelligentPageDiscovery) CreateStructureBasedClassificationConfidence(results []*PageDiscoveryResult, structureAnalysis *StructureAnalysisResult) float64 {
	confidence := 0.0

	// Base confidence from structure analysis
	if structureAnalysis.StructureRelevance != nil {
		confidence += structureAnalysis.StructureRelevance.ConfidenceScore * 0.4
	}

	// Confidence from page discovery results
	if len(results) > 0 {
		totalRelevance := 0.0
		totalQuality := 0.0
		for _, result := range results {
			totalRelevance += result.RelevanceScore
			totalQuality += result.ContentQuality
		}
		avgRelevance := totalRelevance / float64(len(results))
		avgQuality := totalQuality / float64(len(results))
		confidence += avgRelevance * 0.3
		confidence += avgQuality * 0.3
	}

	return math.Min(confidence, 1.0)
}
