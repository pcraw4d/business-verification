package classification

import (
	"context"
	"log"
	"regexp"
	"strings"

	"kyb-platform/internal/classification/methods"
)

// extractTitleFromHTML extracts the title from HTML content
func extractTitleFromHTML(htmlContent string) string {
	if htmlContent == "" {
		return ""
	}
	
	// Try to extract title tag
	titleRegex := regexp.MustCompile(`(?i)<title[^>]*>(.*?)</title>`)
	matches := titleRegex.FindStringSubmatch(htmlContent)
	if len(matches) > 1 {
		title := strings.TrimSpace(matches[1])
		// Decode basic HTML entities
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

// websiteScraperAdapter adapts EnhancedWebsiteScraper to methods.WebsiteScraper interface
// It also integrates SmartWebsiteCrawler for multi-page scraping
type websiteScraperAdapter struct {
	scraper      *EnhancedWebsiteScraper
	smartCrawler *SmartWebsiteCrawler
	logger       *log.Logger
}

// ScrapeWebsite implements methods.WebsiteScraper interface
func (w *websiteScraperAdapter) ScrapeWebsite(ctx context.Context, websiteURL string) *methods.ScrapingResult {
	result := w.scraper.ScrapeWebsite(ctx, websiteURL)
	if result == nil {
		return &methods.ScrapingResult{
			Success: false,
			Error:   "scraping returned nil result",
		}
	}

	// Extract title from HTML content if available
	title := extractTitleFromHTML(result.Content)
	
	// Convert classification.ScrapingResult to methods.ScrapingResult
	return &methods.ScrapingResult{
		URL:           result.URL,
		StatusCode:    result.StatusCode,
		Content:       result.Content,
		TextContent:   result.TextContent,
		Title:         title,
		Keywords:      result.Keywords,
		ContentType:   result.ContentType,
		ContentLength: result.ContentLength,
		Headers:       result.Headers,
		FinalURL:      result.FinalURL,
		Success:       result.Success,
		Error:         result.Error,
	}
}

// codeGeneratorAdapter adapts ClassificationCodeGenerator to methods.CodeGenerator interface
type codeGeneratorAdapter struct {
	generator *ClassificationCodeGenerator
}

// GenerateClassificationCodes implements methods.CodeGenerator interface
func (c *codeGeneratorAdapter) GenerateClassificationCodes(
	ctx context.Context,
	keywords []string,
	detectedIndustry string,
	confidence float64,
	additionalIndustries ...methods.IndustryResult,
) (*methods.ClassificationCodesInfo, error) {
	// Convert methods.IndustryResult to classification.IndustryResult
	industries := make([]IndustryResult, len(additionalIndustries))
	for i, ind := range additionalIndustries {
		industries[i] = IndustryResult{
			IndustryName: ind.IndustryName,
			Confidence:   ind.Confidence,
		}
	}

	// Call the actual generator
	result, err := c.generator.GenerateClassificationCodes(
		ctx,
		keywords,
		detectedIndustry,
		confidence,
		industries...,
	)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return &methods.ClassificationCodesInfo{}, nil
	}

	// Convert classification.ClassificationCodesInfo to methods.ClassificationCodesInfo
	return &methods.ClassificationCodesInfo{
		MCC:   convertMCCCodesToMethods(result.MCC),
		SIC:   convertSICCodesToMethods(result.SIC),
		NAICS: convertNAICSCodesToMethods(result.NAICS),
	}, nil
}

// convertMCCCodesToMethods converts classification.MCCCode to methods.MCCCode
func convertMCCCodesToMethods(codes []MCCCode) []methods.MCCCode {
	result := make([]methods.MCCCode, len(codes))
	for i, code := range codes {
		result[i] = methods.MCCCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
			Keywords:    code.Keywords,
		}
	}
	return result
}

// convertSICCodesToMethods converts classification.SICCode to methods.SICCode
func convertSICCodesToMethods(codes []SICCode) []methods.SICCode {
	result := make([]methods.SICCode, len(codes))
	for i, code := range codes {
		result[i] = methods.SICCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
			Keywords:    code.Keywords,
		}
	}
	return result
}

// convertNAICSCodesToMethods converts classification.NAICSCode to methods.NAICSCode
func convertNAICSCodesToMethods(codes []NAICSCode) []methods.NAICSCode {
	result := make([]methods.NAICSCode, len(codes))
	for i, code := range codes {
		result[i] = methods.NAICSCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
			Keywords:    code.Keywords,
		}
	}
	return result
}

// ScrapeMultiPage implements methods.WebsiteScraper interface for multi-page scraping
func (w *websiteScraperAdapter) ScrapeMultiPage(ctx context.Context, websiteURL string) string {
	if w.smartCrawler == nil {
		// Create SmartWebsiteCrawler if not already created
		if w.logger != nil {
			w.logger.Printf("🕷️ Creating SmartWebsiteCrawler for multi-page scraping")
		} else {
			// Fallback: create a default logger
			w.logger = log.Default()
		}
		w.smartCrawler = NewSmartWebsiteCrawler(w.logger)
	}

	// Use SmartWebsiteCrawler to crawl multiple pages
	crawlResult, err := w.smartCrawler.CrawlWebsite(ctx, websiteURL)
	if err != nil {
		if w.logger != nil {
			w.logger.Printf("⚠️ Multi-page crawling failed: %v", err)
		}
		return ""
	}

	if crawlResult == nil || len(crawlResult.PagesAnalyzed) == 0 {
		if w.logger != nil {
			w.logger.Printf("⚠️ Multi-page crawling returned no pages")
		}
		return ""
	}

	// Extract text content from all pages
	var combinedText []string
	for _, page := range crawlResult.PagesAnalyzed {
		// Combine title and keywords as text content
		pageText := ""
		if page.Title != "" {
			pageText = page.Title
		}
		if len(page.Keywords) > 0 {
			keywordsText := strings.Join(page.Keywords, " ")
			if pageText != "" {
				pageText = pageText + " " + keywordsText
			} else {
				pageText = keywordsText
			}
		}
		// Add business info if available
		if page.BusinessInfo.BusinessName != "" {
			pageText = pageText + " " + page.BusinessInfo.BusinessName
		}
		if page.BusinessInfo.Description != "" {
			pageText = pageText + " " + page.BusinessInfo.Description
		}
		if pageText != "" {
			combinedText = append(combinedText, pageText)
		}
	}

	result := strings.Join(combinedText, " ")
	if w.logger != nil {
		w.logger.Printf("✅ Multi-page scraping extracted %d characters from %d pages", len(result), len(crawlResult.PagesAnalyzed))
	}
	return result
}

// NewWebsiteScraperAdapter creates an adapter for EnhancedWebsiteScraper with multi-page support
// It accepts an optional logger parameter for SmartWebsiteCrawler initialization
func NewWebsiteScraperAdapter(scraper *EnhancedWebsiteScraper) methods.WebsiteScraper {
	// Create a default logger (we can't access scraper's logger directly due to package structure)
	logger := log.Default()
	
	return &websiteScraperAdapter{
		scraper:      scraper,
		smartCrawler: nil, // Lazy initialization
		logger:       logger,
	}
}

// NewCodeGeneratorAdapter creates an adapter for ClassificationCodeGenerator
func NewCodeGeneratorAdapter(generator *ClassificationCodeGenerator) methods.CodeGenerator {
	return &codeGeneratorAdapter{generator: generator}
}

