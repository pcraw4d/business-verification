package classification

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"kyb-platform/internal/external"
)

// WebsiteContentService provides unified website content extraction with request-scoped deduplication
type WebsiteContentService struct {
	scraper              *external.WebsiteScraper
	smartCrawler         *SmartWebsiteCrawler
	contentCache         WebsiteContentCacher
	logger               interface {
		Printf(format string, v ...interface{})
	}
	extractionMutex      sync.Mutex
	inFlightExtractions  map[string]*extractionInProgress
}

// extractionInProgress tracks an in-flight website content extraction
type extractionInProgress struct {
	result *WebsiteContentResult
	err    error
	done   chan struct{}
	mu     sync.Mutex
}

// WebsiteContentResult represents the result of website content extraction
type WebsiteContentResult struct {
	TextContent    string
	Title          string
	Keywords       []string
	StructuredData map[string]interface{}
	Success        bool
	Error          string
	Duration       time.Duration
	StatusCode     int
	ContentType    string
}

// NewWebsiteContentService creates a new website content service
func NewWebsiteContentService(
	scraper *external.WebsiteScraper,
	smartCrawler *SmartWebsiteCrawler,
	contentCache WebsiteContentCacher,
	logger interface {
		Printf(format string, v ...interface{})
	},
) *WebsiteContentService {
	return &WebsiteContentService{
		scraper:             scraper,
		smartCrawler:        smartCrawler,
		contentCache:        contentCache,
		logger:              logger,
		inFlightExtractions: make(map[string]*extractionInProgress),
	}
}

// ExtractWebsiteContent extracts website content once per request, with deduplication
// This ensures the website is only scraped once even if multiple components request it
func (wcs *WebsiteContentService) ExtractWebsiteContent(
	ctx context.Context,
	websiteURL string,
	useFastPath bool,
) (*WebsiteContentResult, error) {
	if websiteURL == "" {
		return &WebsiteContentResult{Success: false}, nil
	}

	// Check ClassificationContext first (request-scoped cache)
	if classificationCtx, ok := GetClassificationContext(ctx); ok {
		if content := classificationCtx.GetWebsiteContent(); content != "" {
			wcs.logger.Printf("📦 [WebsiteContentService] Using content from ClassificationContext for %s", websiteURL)
			return &WebsiteContentResult{
				TextContent: content,
				Success:     true,
				Duration:    0, // Already cached
			}, nil
		}
	}

	// Check if extraction already in progress (deduplication)
	wcs.extractionMutex.Lock()
	if inFlight, exists := wcs.inFlightExtractions[websiteURL]; exists {
		wcs.extractionMutex.Unlock()
		wcs.logger.Printf("⏳ [WebsiteContentService] Waiting for in-flight extraction for %s", websiteURL)
		<-inFlight.done
		inFlight.mu.Lock()
		defer inFlight.mu.Unlock()
		return inFlight.result, inFlight.err
	}

	// Create new extraction
	inFlight := &extractionInProgress{
		done: make(chan struct{}),
	}
	wcs.inFlightExtractions[websiteURL] = inFlight
	wcs.extractionMutex.Unlock()

	// Perform extraction
	defer func() {
		close(inFlight.done)
		wcs.extractionMutex.Lock()
		delete(wcs.inFlightExtractions, websiteURL)
		wcs.extractionMutex.Unlock()
	}()

	startTime := time.Now()
	result := &WebsiteContentResult{}

	// Check Redis cache first if enabled
	if wcs.contentCache != nil && wcs.contentCache.IsEnabled() {
		if cached, found := wcs.contentCache.Get(ctx, websiteURL); found {
			wcs.logger.Printf("📦 [WebsiteContentService] Cache hit for %s", websiteURL)
			result.TextContent = cached.TextContent
			result.Title = cached.Title
			result.Keywords = cached.Keywords
			result.StructuredData = cached.StructuredData
			result.Success = cached.Success
			result.StatusCode = cached.StatusCode
			result.ContentType = cached.ContentType
			result.Duration = time.Since(startTime)

			// Store in ClassificationContext for reuse in this request
			if classificationCtx, ok := GetClassificationContext(ctx); ok {
				classificationCtx.SetWebsiteContent(result.TextContent)
				classificationCtx.SetStructuredData(result.StructuredData)
			}

			inFlight.result = result
			return result, nil
		}
	}

	// Perform scraping
	if useFastPath {
		// Use fast single-page scraping
		wcs.logger.Printf("🚀 [WebsiteContentService] Fast-path scraping for %s", websiteURL)
		if wcs.scraper != nil {
			scrapingResult, err := wcs.scraper.ScrapeWebsite(ctx, websiteURL)
			if err != nil {
				result.Error = err.Error()
				result.Success = false
				inFlight.result = result
				inFlight.err = err
				return result, err
			}

			// Extract text content from HTML if needed
			result.TextContent = wcs.extractTextFromScrapingResult(scrapingResult)
			result.Title = wcs.extractTitleFromScrapingResult(scrapingResult)
			result.Keywords = []string{} // Will be extracted later if needed
			result.Success = scrapingResult.StatusCode >= 200 && scrapingResult.StatusCode < 300
			result.StatusCode = scrapingResult.StatusCode
			result.ContentType = scrapingResult.ContentType
			
			// Early termination: Check if content is sufficient (Task 1.5)
			// Use fast-path mode flag for threshold selection
			if wcs.isContentSufficient(result.TextContent, result.Keywords, useFastPath) {
				wcs.logger.Printf("✅ [WebsiteContentService] Content sufficient, skipping full crawl (early termination)")
				result.Duration = time.Since(startTime)
				// Cache and store in context (same as below)
				if result.Success && wcs.contentCache != nil && wcs.contentCache.IsEnabled() {
					cached := &CachedWebsiteContent{
						TextContent:    result.TextContent,
						Title:          result.Title,
						Keywords:       result.Keywords,
						StructuredData: make(map[string]interface{}),
						ScrapedAt:      time.Now(),
						Success:        true,
						StatusCode:     result.StatusCode,
						ContentType:    result.ContentType,
					}
					wcs.contentCache.Set(ctx, websiteURL, cached)
				}
				if classificationCtx, ok := GetClassificationContext(ctx); ok {
					classificationCtx.SetWebsiteContent(result.TextContent)
					classificationCtx.SetStructuredData(result.StructuredData)
				}
				inFlight.result = result
				return result, nil
			}
		} else {
			result.Error = "scraper not available"
			result.Success = false
			inFlight.result = result
			inFlight.err = fmt.Errorf("scraper not available")
			return result, inFlight.err
		}
	} else {
		// Use smart crawling for comprehensive analysis
		wcs.logger.Printf("🔍 [WebsiteContentService] Full crawl for %s", websiteURL)
		if wcs.smartCrawler != nil {
			crawlResult, err := wcs.smartCrawler.CrawlWebsite(ctx, websiteURL)
			if err != nil {
				result.Error = err.Error()
				result.Success = false
				inFlight.result = result
				inFlight.err = err
				return result, err
			}

			// Aggregate content from all pages
			result.TextContent = wcs.aggregateCrawlContent(crawlResult)
			result.Keywords = crawlResult.Keywords
			result.StructuredData = wcs.extractStructuredDataFromCrawl(crawlResult)
			result.Success = crawlResult.Success
		} else {
			result.Error = "smart crawler not available"
			result.Success = false
			inFlight.result = result
			inFlight.err = fmt.Errorf("smart crawler not available")
			return result, inFlight.err
		}
	}

	result.Duration = time.Since(startTime)

	// Cache the result if successful
	if result.Success && wcs.contentCache != nil && wcs.contentCache.IsEnabled() {
		cached := &CachedWebsiteContent{
			TextContent:    result.TextContent,
			Title:          result.Title,
			Keywords:       result.Keywords,
			StructuredData: result.StructuredData,
			ScrapedAt:      time.Now(),
			Success:        true,
			StatusCode:     result.StatusCode,
			ContentType:    result.ContentType,
		}
		if err := wcs.contentCache.Set(ctx, websiteURL, cached); err != nil {
			wcs.logger.Printf("⚠️ [WebsiteContentService] Failed to cache content: %v", err)
		}
	}

	// Store in ClassificationContext for reuse in this request
	if classificationCtx, ok := GetClassificationContext(ctx); ok {
		classificationCtx.SetWebsiteContent(result.TextContent)
		classificationCtx.SetStructuredData(result.StructuredData)
	}

	inFlight.result = result
	return result, nil
}

// extractTextFromScrapingResult extracts text content from external.ScrapingResult
func (wcs *WebsiteContentService) extractTextFromScrapingResult(result *external.ScrapingResult) string {
	// The external.ScrapingResult has Content field which is HTML
	// We need to extract text from it, but for now just return the content
	// In a full implementation, we'd parse HTML and extract text
	return result.Content
}

// extractTitleFromScrapingResult extracts title from external.ScrapingResult
func (wcs *WebsiteContentService) extractTitleFromScrapingResult(result *external.ScrapingResult) string {
	// Extract title from HTML content
	// This is a simple regex-based extraction
	titleRegex := regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)
	matches := titleRegex.FindStringSubmatch(result.Content)
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
	return ""
}

// aggregateCrawlContent aggregates content from all crawled pages
func (wcs *WebsiteContentService) aggregateCrawlContent(crawlResult *CrawlResult) string {
	var content strings.Builder
	for _, page := range crawlResult.PagesAnalyzed {
		// PageAnalysis has Title and Keywords, but not TextContent directly
		// We'll combine title and keywords, or use structured data if available
		if page.Title != "" {
			content.WriteString(page.Title)
			content.WriteString(" ")
		}
		if len(page.Keywords) > 0 {
			content.WriteString(strings.Join(page.Keywords, " "))
			content.WriteString(" ")
		}
		// If structured data has text content, use it
		if textContent, ok := page.StructuredData["text_content"].(string); ok && textContent != "" {
			content.WriteString(textContent)
			content.WriteString("\n\n")
		}
	}
	return content.String()
}

// extractStructuredDataFromCrawl extracts structured data from crawl result
func (wcs *WebsiteContentService) extractStructuredDataFromCrawl(crawlResult *CrawlResult) map[string]interface{} {
	structuredData := make(map[string]interface{})
	if crawlResult.BusinessInfo.BusinessName != "" {
		structuredData["business_name"] = crawlResult.BusinessInfo.BusinessName
	}
	if crawlResult.BusinessInfo.Description != "" {
		structuredData["description"] = crawlResult.BusinessInfo.Description
	}
	if len(crawlResult.BusinessInfo.Services) > 0 {
		structuredData["services"] = crawlResult.BusinessInfo.Services
	}
	if len(crawlResult.BusinessInfo.Products) > 0 {
		structuredData["products"] = crawlResult.BusinessInfo.Products
	}
	return structuredData
}

// isContentSufficient checks if content is sufficient to skip full crawl (Task 1.5: Early Termination)
// For fast-path mode, uses more lenient thresholds (300 chars, 5 keywords)
// For regular mode, uses standard thresholds (500 chars, 10 keywords)
func (wcs *WebsiteContentService) isContentSufficient(textContent string, keywords []string, useFastPath bool) bool {
	minContentLength := 500
	minKeywordCount := 10
	if useFastPath {
		minContentLength = 300  // Lower threshold for fast-path
		minKeywordCount = 5     // Lower threshold for fast-path
	}
	
	contentLength := len(strings.TrimSpace(textContent))
	keywordCount := len(keywords)
	
	sufficient := contentLength >= minContentLength && keywordCount >= minKeywordCount
	
	modeLabel := "REGULAR"
	if useFastPath {
		modeLabel = "FAST-PATH"
	}
	
	if sufficient {
		wcs.logger.Printf("📊 [WebsiteContentService] [%s] Content sufficient: %d chars, %d keywords", modeLabel, contentLength, keywordCount)
	} else {
		wcs.logger.Printf("📊 [WebsiteContentService] [%s] Content insufficient: %d chars (need %d), %d keywords (need %d)", 
			modeLabel, contentLength, minContentLength, keywordCount, minKeywordCount)
	}
	
	return sufficient
}

// ContentQuality represents the quality assessment of website content (Task 3.2)
type ContentQuality struct {
	Quality      string  // "insufficient", "minimal", "good", "optimal"
	Length       int     // Content length in characters
	KeywordCount int     // Number of keywords extracted
	Score        float64 // Quality score (0.0-1.0)
}

// assessContentQuality assesses the quality of extracted content (Task 3.2: Smart Crawling)
func (wcs *WebsiteContentService) assessContentQuality(textContent string, keywords []string) ContentQuality {
	length := len(strings.TrimSpace(textContent))
	keywordCount := len(keywords)
	
	// Quality thresholds
	const insufficientThreshold = 200  // < 200 chars: insufficient
	const minimalThreshold = 500       // 200-500 chars: minimal
	const goodThreshold = 1500         // 500-1500 chars: good
	// > 1500 chars: optimal
	
	const minKeywordsInsufficient = 5
	const minKeywordsMinimal = 10
	const minKeywordsGood = 20
	
	var quality string
	var score float64
	
	if length < insufficientThreshold || keywordCount < minKeywordsInsufficient {
		quality = "insufficient"
		score = 0.2
	} else if length < minimalThreshold || keywordCount < minKeywordsMinimal {
		quality = "minimal"
		score = 0.5
	} else if length < goodThreshold || keywordCount < minKeywordsGood {
		quality = "good"
		score = 0.8
	} else {
		quality = "optimal"
		score = 1.0
	}
	
	return ContentQuality{
		Quality:      quality,
		Length:       length,
		KeywordCount: keywordCount,
		Score:        score,
	}
}

