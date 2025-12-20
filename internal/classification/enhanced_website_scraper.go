package classification

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"kyb-platform/internal/external"

	"github.com/andybalholm/brotli"
	"go.uber.org/zap"
)

// EnhancedWebsiteScraper provides advanced website scraping capabilities
// Now uses the enhanced external.WebsiteScraper with Phase 1 features (multi-tier strategies, structured content)
type EnhancedWebsiteScraper struct {
	logger         *log.Logger
	client         *http.Client
	contentCache   WebsiteContentCacher // Optional cache for website content
	externalScraper *external.WebsiteScraper // Enhanced scraper with Phase 1 features
	zapLogger      *zap.Logger // For external scraper
}

// WebsiteContentCacher interface for caching website content
// This allows different cache implementations to be used
type WebsiteContentCacher interface {
	Get(ctx context.Context, url string) (*CachedWebsiteContent, bool)
	Set(ctx context.Context, url string, content *CachedWebsiteContent) error
	IsEnabled() bool
}

// CachedWebsiteContent represents cached website content
type CachedWebsiteContent struct {
	TextContent    string                 `json:"text_content"`
	Title          string                 `json:"title"`
	Keywords       []string               `json:"keywords"`
	StructuredData map[string]interface{} `json:"structured_data"`
	ScrapedAt      time.Time              `json:"scraped_at"`
	Success        bool                   `json:"success"`
	StatusCode     int                    `json:"status_code,omitempty"`
	ContentType    string                 `json:"content_type,omitempty"`
}

// ScrapingResult represents the result of a website scraping operation
type ScrapingResult struct {
	URL           string            `json:"url"`
	StatusCode    int               `json:"status_code"`
	Content       string            `json:"content"`
	TextContent   string            `json:"text_content"`
	Keywords      []string          `json:"keywords"`
	ContentType   string            `json:"content_type"`
	ContentLength int64             `json:"content_length"`
	Headers       map[string]string `json:"headers"`
	FinalURL      string            `json:"final_url"`
	ScrapedAt     time.Time         `json:"scraped_at"`
	Duration      time.Duration     `json:"duration"`
	Error         string            `json:"error,omitempty"`
	Success       bool              `json:"success"`
}

// NewEnhancedWebsiteScraper creates a new enhanced website scraper
// Now uses external.WebsiteScraper with Phase 1 features (multi-tier strategies, structured content extraction)
func NewEnhancedWebsiteScraper(logger *log.Logger) *EnhancedWebsiteScraper {
	// Create zap logger for external scraper (convert std logger to zap)
	// Use production logger that outputs to stdout (Railway will capture this)
	zapLogger, err := zap.NewProduction()
	if err != nil {
		// Fallback to no-op if production logger fails
		zapLogger = zap.NewNop()
		if logger != nil {
			logger.Printf("⚠️ Failed to create zap logger for external scraper, using no-op logger")
		}
	}

	// Log initialization with Phase 1 features
	if logger != nil {
		playwrightURL := os.Getenv("PLAYWRIGHT_SERVICE_URL")
		if playwrightURL != "" {
			logger.Printf("✅ [Phase1] Initializing enhanced scraper with Playwright service: %s", playwrightURL)
		} else {
			logger.Printf("ℹ️ [Phase1] Initializing enhanced scraper without Playwright (PLAYWRIGHT_SERVICE_URL not set)")
		}
	}

	// Create enhanced external scraper with Phase 1 features
	// This automatically reads PLAYWRIGHT_SERVICE_URL from environment
	config := external.DefaultScrapingConfig()
	config.Timeout = 20 * time.Second
	externalScraper := external.NewWebsiteScraper(config, zapLogger)

	return &EnhancedWebsiteScraper{
		logger:          logger,
		client: &http.Client{
			Timeout: 20 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				MaxIdleConnsPerHost: 2,
			},
		},
		externalScraper: externalScraper,
		zapLogger:      zapLogger,
	}
}

// SetContentCache sets the content cache for the scraper
func (ews *EnhancedWebsiteScraper) SetContentCache(cache WebsiteContentCacher) {
	ews.contentCache = cache
}

// ScrapeWebsite performs enhanced website scraping with comprehensive error handling
// Now uses external.WebsiteScraper with Phase 1 features (multi-tier strategies, structured content)
func (ews *EnhancedWebsiteScraper) ScrapeWebsite(ctx context.Context, websiteURL string) *ScrapingResult {
	startTime := time.Now()
	ews.logger.Printf("🌐 [Enhanced] ScrapeWebsite called for: %s", websiteURL)
	ews.logger.Printf("🔍 [Enhanced] External scraper available: %v", ews.externalScraper != nil)
	
	result := &ScrapingResult{
		URL:       websiteURL,
		ScrapedAt: time.Now(),
		Success:   false,
	}

	// Check cache first if enabled
	if ews.contentCache != nil && ews.contentCache.IsEnabled() {
		if cached, found := ews.contentCache.Get(ctx, websiteURL); found {
			ews.logger.Printf("📦 [Enhanced] Using cached content for %s", websiteURL)
			result.TextContent = cached.TextContent
			result.Keywords = cached.Keywords
			result.Success = cached.Success
			result.StatusCode = cached.StatusCode
			result.ContentType = cached.ContentType
			result.Duration = time.Since(startTime)
			return result
		}
	}

	// Use enhanced external scraper with Phase 1 features
	// This will use multi-tier strategies (SimpleHTTP → BrowserHeaders → Playwright)
	// and extract structured content with quality scoring
	if ews.externalScraper != nil {
		ews.logger.Printf("🌐 [Enhanced] Using Phase 1 enhanced scraper for: %s", websiteURL)
		ews.logger.Printf("🔍 [Phase1] Starting scrape with structured content extraction for: %s", websiteURL)
		
		// Verify context before passing it - detailed logging
		ews.logger.Printf("🔍 [Enhanced] [ContextCheck] Checking context state before passing to external scraper")
		ctxErr := ctx.Err()
		ews.logger.Printf("🔍 [Enhanced] [ContextCheck] Context error state: %v", ctxErr)
		
		if deadline, ok := ctx.Deadline(); ok {
			timeUntilDeadline := time.Until(deadline)
			ews.logger.Printf("⏱️ [Enhanced] [ContextCheck] Context deadline: %v from now (deadline: %v, now: %v, valid: %v)", 
				timeUntilDeadline, deadline, time.Now(), ctxErr == nil)
		} else {
			ews.logger.Printf("⏱️ [Enhanced] [ContextCheck] Context has no deadline (valid: %v)", ctxErr == nil)
		}
		
		if ctxErr != nil {
			ews.logger.Printf("❌ [Enhanced] [ContextCheck] Context already cancelled before calling external scraper: %v", ctxErr)
			result.Error = ctxErr.Error()
			result.Success = false
			return result
		}
		
		// Use ScrapeWebsite which now automatically uses ScrapeWithStructuredContent when strategies are available
		ews.logger.Printf("🚀 [Phase1] Calling external.ScrapeWebsite() for: %s (context valid: %v)", websiteURL, ctxErr == nil)
		scrapingResult, err := ews.externalScraper.ScrapeWebsite(ctx, websiteURL)
		ews.logger.Printf("📥 [Phase1] External scraper returned (result nil: %v, error: %v)", scrapingResult == nil, err != nil)
		if err == nil && scrapingResult != nil {
			// Convert external.ScrapingResult to internal.ScrapingResult
			result.URL = scrapingResult.URL
			result.StatusCode = scrapingResult.StatusCode
			result.Content = scrapingResult.Content
			result.ContentType = scrapingResult.ContentType
			result.ContentLength = scrapingResult.ContentLength
			result.Headers = scrapingResult.Headers
			result.FinalURL = scrapingResult.FinalURL
			result.ScrapedAt = scrapingResult.ScrapedAt
			result.Duration = scrapingResult.Duration
			result.Success = scrapingResult.StatusCode >= 200 && scrapingResult.StatusCode < 300

			// Extract text content from HTML if structured content is available
			if scrapingResult.StructuredContent != nil {
				// Use the weighted combined text from structured content
				result.TextContent = scrapingResult.StructuredContent.PlainText
				result.Keywords = ews.extractBusinessKeywords(result.TextContent)
				
				ews.logger.Printf("✅ [Phase1] Strategy succeeded - Quality: %.2f, Words: %d", 
					scrapingResult.StructuredContent.QualityScore,
					scrapingResult.StructuredContent.WordCount)
				ews.logger.Printf("✅ [Enhanced] Phase 1 scraper succeeded - Quality: %.2f, Words: %d", 
					scrapingResult.StructuredContent.QualityScore,
					scrapingResult.StructuredContent.WordCount)
			} else {
				// Fallback to extracting text from HTML
				result.TextContent = ews.extractTextFromHTML(scrapingResult.Content)
				result.Keywords = ews.extractBusinessKeywords(result.TextContent)
				ews.logger.Printf("⚠️ [Phase1] No structured content available, using HTML extraction")
			}

			// Cache the result if successful
			if result.Success && ews.contentCache != nil && ews.contentCache.IsEnabled() {
				cached := &CachedWebsiteContent{
					TextContent:    result.TextContent,
					Title:          ews.extractTitle(scrapingResult.Content),
					Keywords:       result.Keywords,
					StructuredData: make(map[string]interface{}),
					ScrapedAt:      time.Now(),
					Success:        true,
					StatusCode:     result.StatusCode,
					ContentType:    result.ContentType,
				}
				ews.contentCache.Set(ctx, websiteURL, cached)
			}

			return result
		}
		
		// If external scraper failed, log and fall through to legacy method
		if err != nil {
			ews.logger.Printf("❌ [Phase1] All scraping strategies failed: %v", err)
			ews.logger.Printf("⚠️ [Enhanced] Phase 1 scraper failed, falling back to legacy method: %v", err)
		}
	}

	// Legacy scraping method (fallback if external scraper not available or failed)
	ews.logger.Printf("🌐 [Enhanced] Starting legacy website scraping for: %s", websiteURL)

	// Validate and normalize URL
	normalizedURL, err := ews.normalizeURL(websiteURL)
	if err != nil {
		result.Error = fmt.Sprintf("URL validation failed: %v", err)
		ews.logger.Printf("❌ [Enhanced] %s", result.Error)
		return result
	}
	result.URL = normalizedURL

	// Create request with enhanced headers
	req, err := ews.createRequest(ctx, normalizedURL)
	if err != nil {
		result.Error = fmt.Sprintf("Request creation failed: %v", err)
		ews.logger.Printf("❌ [Enhanced] %s", result.Error)
		return result
	}

	// Make HTTP request
	resp, err := ews.client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("HTTP request failed: %v", err)
		ews.logger.Printf("❌ [Enhanced] %s", result.Error)
		return result
	}
	defer resp.Body.Close()

	// Process response
	result.StatusCode = resp.StatusCode
	result.ContentType = resp.Header.Get("Content-Type")
	result.FinalURL = resp.Request.URL.String()

	// Extract headers
	result.Headers = make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			result.Headers[key] = values[0]
		}
	}

	ews.logger.Printf("📊 [Enhanced] Response received - Status: %d, Content-Type: %s, Final URL: %s",
		result.StatusCode, result.ContentType, result.FinalURL)

	// Check for errors
	if resp.StatusCode >= 400 {
		result.Error = fmt.Sprintf("HTTP error: %d %s", resp.StatusCode, resp.Status)
		ews.logger.Printf("❌ [Enhanced] %s", result.Error)
		return result
	}

	// Read response body
	body, err := ews.readResponseBody(resp)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to read response body: %v", err)
		ews.logger.Printf("❌ [Enhanced] %s", result.Error)
		return result
	}

	// Check for CAPTCHA before processing
	captchaResult := DetectCAPTCHA(resp, body)
	if captchaResult.Detected {
		result.Error = fmt.Sprintf("CAPTCHA detected (%s)", captchaResult.Type)
		ews.logger.Printf("🚫 [Enhanced] CAPTCHA detected (%s) for %s - stopping", captchaResult.Type, websiteURL)
		return result // Stop immediately when CAPTCHA is detected
	}

	result.Content = string(body)
	result.ContentLength = int64(len(body))

	// Extract text content
	result.TextContent = ews.extractTextFromHTML(result.Content)
	ews.logger.Printf("🧹 [Enhanced] Extracted %d characters of text content", len(result.TextContent))

	// Extract keywords
	result.Keywords = ews.extractBusinessKeywords(result.TextContent)
	ews.logger.Printf("🔍 [Enhanced] Extracted %d keywords: %v", len(result.Keywords), result.Keywords)

	// Extract title from HTML
	title := ews.extractTitle(result.Content)
	if title != "" {
		ews.logger.Printf("📄 [Enhanced] Extracted title: %s", title)
	}

	result.Duration = time.Since(startTime)
	result.Success = true

	// Cache the result if cache is enabled
	if ews.contentCache != nil && ews.contentCache.IsEnabled() && result.Success {
		cached := &CachedWebsiteContent{
			TextContent:   result.TextContent,
			Title:         title,
			Keywords:      result.Keywords,
			StructuredData: make(map[string]interface{}),
			ScrapedAt:     time.Now(),
			Success:       true,
			StatusCode:    result.StatusCode,
			ContentType:   result.ContentType,
		}
		if err := ews.contentCache.Set(ctx, websiteURL, cached); err != nil {
			ews.logger.Printf("⚠️ [Enhanced] Failed to cache content: %v", err)
		}
	}

	ews.logger.Printf("✅ [Enhanced] Website scraping completed for %s in %v", normalizedURL, result.Duration)
	return result
}

// extractTitle extracts the title from HTML content
func (ews *EnhancedWebsiteScraper) extractTitle(htmlContent string) string {
	// Try to extract from <title> tag
	titleRegex := regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)
	matches := titleRegex.FindStringSubmatch(htmlContent)
	if len(matches) > 1 {
		title := strings.TrimSpace(matches[1])
		// Decode HTML entities
		title = strings.ReplaceAll(title, "&amp;", "&")
		title = strings.ReplaceAll(title, "&lt;", "<")
		title = strings.ReplaceAll(title, "&gt;", ">")
		title = strings.ReplaceAll(title, "&quot;", "\"")
		title = strings.ReplaceAll(title, "&#39;", "'")
		title = strings.ReplaceAll(title, "&nbsp;", " ")
		return title
	}

	// Try Open Graph title
	ogTitleRegex := regexp.MustCompile(`(?i)<meta[^>]*property=["']og:title["'][^>]*content=["']([^"']+)["']`)
	ogMatches := ogTitleRegex.FindStringSubmatch(htmlContent)
	if len(ogMatches) > 1 {
		return strings.TrimSpace(ogMatches[1])
	}

	return ""
}

// normalizeURL validates and normalizes the URL
func (ews *EnhancedWebsiteScraper) normalizeURL(websiteURL string) (string, error) {
	// Add scheme if missing
	if !strings.HasPrefix(websiteURL, "http://") && !strings.HasPrefix(websiteURL, "https://") {
		websiteURL = "https://" + websiteURL
		ews.logger.Printf("🔧 [Enhanced] Added HTTPS scheme: %s", websiteURL)
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %w", err)
	}

	// Ensure we have a valid host
	if parsedURL.Host == "" {
		return "", fmt.Errorf("missing host in URL")
	}

	return parsedURL.String(), nil
}

// createRequest creates an HTTP request with enhanced headers
func (ews *EnhancedWebsiteScraper) createRequest(ctx context.Context, websiteURL string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", websiteURL, nil)
	if err != nil {
		return nil, err
	}

	// Set comprehensive headers with randomization to mimic a real browser
	headers := GetRandomizedHeaders(GetUserAgent())
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

// readResponseBody reads the response body with size limiting and decompression
func (ews *EnhancedWebsiteScraper) readResponseBody(resp *http.Response) ([]byte, error) {
	maxSize := int64(10 * 1024 * 1024) // 10MB limit

	// Check if content is compressed
	contentEncoding := resp.Header.Get("Content-Encoding")
	ews.logger.Printf("📦 [Enhanced] Content-Encoding: %s", contentEncoding)

	var reader io.Reader = resp.Body

	// Handle different compression types
	switch contentEncoding {
	case "gzip":
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			ews.logger.Printf("⚠️ [Enhanced] Failed to create gzip reader: %v", err)
			// Fallback to reading without decompression
		} else {
			defer gzipReader.Close()
			reader = gzipReader
			ews.logger.Printf("📦 [Enhanced] Using gzip decompression")
		}
	case "br":
		// Handle Brotli compression
		ews.logger.Printf("📦 [Enhanced] Brotli compression detected - attempting decompression")
		body, err := ews.decompressBrotli(resp.Body, maxSize)
		if err != nil {
			ews.logger.Printf("⚠️ [Enhanced] Brotli decompression failed: %v", err)
			// Fallback to reading without decompression
			reader = resp.Body
		} else {
			ews.logger.Printf("📦 [Enhanced] Brotli decompression successful")
			return body, nil
		}
	case "deflate":
		ews.logger.Printf("📦 [Enhanced] Deflate compression detected - using raw reader")
		// For deflate, we'll read as-is since Go's deflate support is limited
	default:
		ews.logger.Printf("📦 [Enhanced] No compression or unsupported compression: %s", contentEncoding)
	}

	body, err := io.ReadAll(io.LimitReader(reader, maxSize))
	if err != nil {
		return nil, err
	}

	ews.logger.Printf("📄 [Enhanced] Read %d bytes from response (decompressed)", len(body))
	return body, nil
}

// decompressBrotli attempts to decompress Brotli-compressed content
func (ews *EnhancedWebsiteScraper) decompressBrotli(body io.Reader, maxSize int64) ([]byte, error) {
	ews.logger.Printf("📦 [Enhanced] Attempting Brotli decompression")

	// Create a Brotli reader
	brotliReader := brotli.NewReader(io.LimitReader(body, maxSize))

	// Read the decompressed data
	decompressedData, err := io.ReadAll(brotliReader)
	if err != nil {
		ews.logger.Printf("⚠️ [Enhanced] Brotli decompression failed: %v", err)
		return nil, fmt.Errorf("Brotli decompression failed: %w", err)
	}

	ews.logger.Printf("📦 [Enhanced] Brotli decompression successful - %d bytes decompressed", len(decompressedData))
	return decompressedData, nil
}

// isLikelyHTML checks if the data looks like HTML content
func (ews *EnhancedWebsiteScraper) isLikelyHTML(data []byte) bool {
	// Convert to string and check for HTML indicators
	content := string(data)

	// Check for common HTML tags
	htmlIndicators := []string{"<html", "<head", "<body", "<div", "<span", "<p", "<h1", "<h2", "<h3", "<title", "<meta", "<script", "<style"}

	for _, indicator := range htmlIndicators {
		if strings.Contains(strings.ToLower(content), indicator) {
			return true
		}
	}

	// Check for high ratio of printable characters
	printableCount := 0
	for _, b := range data {
		if b >= 32 && b <= 126 { // Printable ASCII range
			printableCount++
		}
	}

	// If more than 70% of characters are printable, it's likely text/HTML
	return float64(printableCount)/float64(len(data)) > 0.7
}

// extractTextFromHTML extracts clean text content from HTML
func (ews *EnhancedWebsiteScraper) extractTextFromHTML(htmlContent string) string {
	// Remove script and style tags completely
	htmlContent = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`).ReplaceAllString(htmlContent, "")
	htmlContent = regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`).ReplaceAllString(htmlContent, "")
	htmlContent = regexp.MustCompile(`(?i)<noscript[^>]*>.*?</noscript>`).ReplaceAllString(htmlContent, "")

	// Remove HTML comments
	htmlContent = regexp.MustCompile(`<!--.*?-->`).ReplaceAllString(htmlContent, "")

	// Remove HTML tags
	htmlContent = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(htmlContent, " ")

	// Decode HTML entities (basic ones)
	htmlContent = strings.ReplaceAll(htmlContent, "&amp;", "&")
	htmlContent = strings.ReplaceAll(htmlContent, "&lt;", "<")
	htmlContent = strings.ReplaceAll(htmlContent, "&gt;", ">")
	htmlContent = strings.ReplaceAll(htmlContent, "&quot;", "\"")
	htmlContent = strings.ReplaceAll(htmlContent, "&#39;", "'")
	htmlContent = strings.ReplaceAll(htmlContent, "&nbsp;", " ")

	// Clean up whitespace
	htmlContent = regexp.MustCompile(`\s+`).ReplaceAllString(htmlContent, " ")

	return strings.TrimSpace(htmlContent)
}

// extractBusinessKeywords extracts business-relevant keywords from text content
func (ews *EnhancedWebsiteScraper) extractBusinessKeywords(textContent string) []string {
	var keywords []string

	// Convert to lowercase for processing
	text := strings.ToLower(textContent)

	// Enhanced business-relevant keyword patterns
	businessPatterns := []string{
		// Food & Beverage (Priority 5.3: Enhanced keywords including beverage manufacturing)
		`\b(restaurant|restaurants|cafe|cafes|coffee|coffee shop|coffeehouse|food|dining|kitchen|catering|bakery|bakeries|bar|bars|pub|pubs|brewery|breweries|winery|wineries|wine|beer|cocktail|cocktails|menu|chef|chefs|cook|cooking|cuisine|cuisines|delivery|takeout|take-out|dine-in|fast food|fast-food|casual dining|fine dining|bistro|eatery|diner|tavern|gastropub|brewpub|food truck|food trucks|food service|foodservice|beverage|beverages|beverage manufacturing|drink|drinks|alcohol|alcoholic|spirits|liquor|soft drink|soda|juice|bottled beverage|carbonated|cola|wine bar|wine shop|wine store|wine merchant|wine tasting|wine cellar|sommelier|vintner|vineyard|grapes|grapevine|vintage|bottle|cellar|tasting|oenology|distillery|distilleries|cocktail bar|mixology|bartender|server|waiter|waitress|host|hostess|reservation|reservations|table|seating|patio|outdoor|indoor|ambiance|atmosphere|decor|design|cuisine type|cuisine style|specialty|signature|dish|dishes|appetizer|appetizers|entree|entrees|dessert|desserts|brunch|breakfast|lunch|dinner|supper|happy hour|happy-hour|specials|promotion|promotions|gift card|gift cards|loyalty|rewards|membership|franchise|chain|location|locations|branch|branches)\b`,
		// Technology
		`\b(technology|software|tech|app|digital|web|mobile|cloud|ai|ml|data|cyber|security|programming|development|IT|computer|internet|online|platform|api|database)\b`,
		// Healthcare
		`\b(healthcare|medical|clinic|hospital|doctor|dentist|therapy|wellness|pharmacy|medicine|patient|treatment|health|care|nurse|physician)\b`,
		// Legal
		`\b(legal|law|attorney|lawyer|court|litigation|patent|trademark|copyright|legal services|advocacy|justice|legal advice|law firm)\b`,
		// Retail & E-commerce
		`\b(retail|store|shop|ecommerce|online|fashion|clothing|electronics|beauty|products|merchandise|selling|commerce|shopping|marketplace)\b`,
		// Finance
		`\b(finance|banking|investment|insurance|accounting|tax|financial|credit|loan|money|capital|funding|payment|transaction|wealth)\b`,
		// Real Estate
		`\b(real estate|property|construction|building|architecture|design|interior|home|house|apartment|rental|mortgage|property management)\b`,
		// Education
		`\b(education|school|university|training|learning|course|academy|institute|student|teacher|teaching|academic|degree|certification)\b`,
		// Consulting & Professional Services
		`\b(consulting|advisory|strategy|management|business|corporate|professional|services|expert|specialist|consultant|advisory)\b`,
		// Manufacturing
		`\b(manufacturing|production|factory|industrial|automotive|machinery|equipment|assembly|production|industrial|machinery)\b`,
		// Transportation & Logistics
		`\b(transportation|logistics|shipping|delivery|freight|warehouse|supply chain|trucking|logistics|shipping|delivery)\b`,
		// Entertainment & Media (Priority 5.3: Enhanced keywords)
		`\b(entertainment|media|streaming|video|audio|podcast|music|film|movie|cinema|television|tv|broadcasting|publishing|content|creative|art|gaming|game|esports|sports|events|concert|festival|theater|theatre|performance|show|production|studio|record|label|artist|actor|director|producer|cinematography|animation|visual effects|vfx|post-production|editing|sound|recording|distribution|platform|channel|network|broadcast|live|stream|on-demand|subscription|ticket|venue|arena|stadium)\b`,
		// Energy & Utilities
		`\b(energy|utilities|renewable|solar|wind|oil|gas|power|electricity|energy|utilities|renewable|solar|wind)\b`,
		// Agriculture
		`\b(agriculture|farming|food production|crop|livestock|organic|sustainable|agriculture|farming|food production|crop)\b`,
		// Travel & Hospitality
		`\b(travel|tourism|hospitality|hotel|accommodation|vacation|booking|trip|travel|tourism|hospitality|hotel|accommodation)\b`,
	}

	// Extract keywords using patterns
	for _, pattern := range businessPatterns {
		matches := regexp.MustCompile(pattern).FindAllString(text, -1)
		for _, match := range matches {
			// Remove duplicates and add to keywords
			if !ews.containsKeyword(keywords, match) {
				keywords = append(keywords, match)
			}
		}
	}

	// Also extract common business words
	commonBusinessWords := []string{
		"service", "services", "company", "business", "corp", "corporation", "inc", "llc", "ltd",
		"enterprise", "solutions", "systems", "group", "associates", "partners", "consulting",
		"management", "development", "production", "distribution", "marketing", "sales",
		"customer", "clients", "professional", "expert", "specialist", "quality", "premium",
		"innovative", "leading", "trusted", "reliable", "experienced", "established",
		"support", "help", "contact", "about", "team", "staff", "employees", "office",
		"location", "address", "phone", "email", "website", "online", "digital",
	}

	for _, word := range commonBusinessWords {
		if strings.Contains(text, word) && !ews.containsKeyword(keywords, word) {
			keywords = append(keywords, word)
		}
	}

	// Limit to top 25 keywords to avoid noise
	if len(keywords) > 25 {
		keywords = keywords[:25]
	}

	return keywords
}

// containsKeyword checks if a keyword already exists in the slice
func (ews *EnhancedWebsiteScraper) containsKeyword(keywords []string, keyword string) bool {
	for _, k := range keywords {
		if k == keyword {
			return true
		}
	}
	return false
}
