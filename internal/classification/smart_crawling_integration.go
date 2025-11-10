package classification

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// SmartCrawlingIntegration integrates smart crawling with existing classification pipeline
type SmartCrawlingIntegration struct {
	logger             *log.Logger
	enhancedAnalyzer   *IntegrationEnhancedWebsiteAnalyzer
	existingClassifier ClassificationService
}

// IntegrationEnhancedWebsiteAnalyzer represents the enhanced website analyzer (placeholder)
// (renamed to avoid conflict with enhanced_website_analyzer.go)
type IntegrationEnhancedWebsiteAnalyzer struct {
	logger *log.Logger
}

// NewIntegrationEnhancedWebsiteAnalyzer creates a new enhanced website analyzer
func NewIntegrationEnhancedWebsiteAnalyzer(logger *log.Logger) *IntegrationEnhancedWebsiteAnalyzer {
	return &IntegrationEnhancedWebsiteAnalyzer{logger: logger}
}

// AnalyzeWebsite performs website analysis (placeholder implementation)
func (ewa *IntegrationEnhancedWebsiteAnalyzer) AnalyzeWebsite(ctx context.Context, websiteURL string) (*IntegrationEnhancedAnalysisResult, error) {
	// Placeholder implementation - would integrate with actual smart crawling
		result := &IntegrationEnhancedAnalysisResult{
		WebsiteURL:        websiteURL,
		AnalysisTimestamp: time.Now(),
		Success:           true,
		ProcessingTime:    100 * time.Millisecond,
		OverallConfidence: 0.8,
		BusinessClassification: &IntegrationBusinessClassificationResult{
			PrimaryIndustry:        "Technology",
			IndustryConfidence:     0.8,
			BusinessType:           "Software Development",
			BusinessTypeConfidence: 0.7,
			Keywords:               []string{"software", "development", "technology"},
			ConfidenceScore:        0.8,
			AnalysisMethod:         "smart_crawling",
		},
	}
	return result, nil
}

// IntegrationEnhancedAnalysisResult represents the result of enhanced analysis
// (renamed to avoid conflict with enhanced_website_analyzer.go)
type IntegrationEnhancedAnalysisResult struct {
	WebsiteURL             string                        `json:"website_url"`
	CrawlResult            *IntegrationCrawlResult                  `json:"crawl_result"`
	RelevanceAnalysis      *IntegrationRelevanceAnalysisResult      `json:"relevance_analysis"`
	StructuredData         *StructuredDataResult         `json:"structured_data"`
	BusinessClassification *IntegrationBusinessClassificationResult `json:"business_classification"`
	AnalysisTimestamp      time.Time                     `json:"analysis_timestamp"`
	ProcessingTime         time.Duration                 `json:"processing_time"`
	OverallConfidence      float64                       `json:"overall_confidence"`
	Success                bool                          `json:"success"`
	Error                  string                        `json:"error,omitempty"`
}

// IntegrationBusinessClassificationResult represents the final business classification
// (renamed to avoid conflict with enhanced_website_analyzer.go)
type IntegrationBusinessClassificationResult struct {
	PrimaryIndustry        string               `json:"primary_industry"`
	IndustryConfidence     float64              `json:"industry_confidence"`
	BusinessType           string               `json:"business_type"`
	BusinessTypeConfidence float64              `json:"business_type_confidence"`
			MCCCodes               []WebsiteClassificationCode `json:"mcc_codes"`
			SICCodes               []WebsiteClassificationCode `json:"sic_codes"`
			NAICSCodes             []WebsiteClassificationCode `json:"naics_codes"`
	Keywords               []string             `json:"keywords"`
	ConfidenceScore        float64              `json:"confidence_score"`
	AnalysisMethod         string               `json:"analysis_method"`
}

// IntegrationCrawlResult represents the result of a smart crawl operation
// (renamed to avoid conflict with content_relevance_analyzer.go)
type IntegrationCrawlResult struct {
	BaseURL       string             `json:"base_url"`
	PagesAnalyzed []IntegrationPageAnalysis     `json:"pages_analyzed"`
	TotalPages    int                `json:"total_pages"`
	RelevantPages int                `json:"relevant_pages"`
	Keywords      []string           `json:"keywords"`
	IndustryScore map[string]float64 `json:"industry_score"`
	BusinessInfo  IntegrationBusinessInfo       `json:"business_info"`
	SiteStructure IntegrationSiteStructure      `json:"site_structure"`
	CrawlDuration time.Duration      `json:"crawl_duration"`
	Success       bool               `json:"success"`
	Error         string             `json:"error,omitempty"`
}

// IntegrationPageAnalysis represents analysis of a single page
// (renamed to avoid conflict with content_relevance_analyzer.go)
type IntegrationPageAnalysis struct {
	URL                string                 `json:"url"`
	Title              string                 `json:"title"`
	PageType           string                 `json:"page_type"`
	RelevanceScore     float64                `json:"relevance_score"`
	ContentQuality     float64                `json:"content_quality"`
	Keywords           []string               `json:"keywords"`
	IndustryIndicators []string               `json:"industry_indicators"`
	BusinessInfo       IntegrationBusinessInfo           `json:"business_info"`
	MetaTags           map[string]string      `json:"meta_tags"`
	StructuredData     map[string]interface{} `json:"structured_data"`
	ResponseTime       time.Duration          `json:"response_time"`
	StatusCode         int                    `json:"status_code"`
	ContentLength      int                    `json:"content_length"`
	LastModified       time.Time              `json:"last_modified"`
	Priority           int                    `json:"priority"`
}

// IntegrationBusinessInfo represents extracted business information
// (renamed to avoid conflict with content_relevance_analyzer.go)
type IntegrationBusinessInfo struct {
	BusinessName  string      `json:"business_name"`
	Description   string      `json:"description"`
	Services      []string    `json:"services"`
	Products      []string    `json:"products"`
	ContactInfo   IntegrationContactInfo `json:"contact_info"`
	BusinessHours string      `json:"business_hours"`
	Location      string      `json:"location"`
	Industry      string      `json:"industry"`
	BusinessType  string      `json:"business_type"`
}

// IntegrationContactInfo represents contact information
// (renamed to avoid conflict with content_relevance_analyzer.go)
type IntegrationContactInfo struct {
	Phone   string            `json:"phone"`
	Email   string            `json:"email"`
	Address string            `json:"address"`
	Website string            `json:"website"`
	Social  map[string]string `json:"social"`
}

// IntegrationSiteStructure represents the discovered site structure
// (renamed to avoid conflict with content_relevance_analyzer.go)
type IntegrationSiteStructure struct {
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

// IntegrationRelevanceAnalysisResult represents the result of content relevance analysis
// (renamed to avoid conflict with content_relevance_analyzer.go)
type IntegrationRelevanceAnalysisResult struct {
	OverallRelevance    float64            `json:"overall_relevance"`
	PageRelevance       map[string]float64 `json:"page_relevance"`
	TopKeywords         []KeywordSignal    `json:"top_keywords"`
	DetectedIndustries  []IntegrationIndustrySignal   `json:"detected_industries"`
	ContentQualityScore float64            `json:"content_quality_score"`
	ConfidenceScore     float64            `json:"confidence_score"`
	AnalysisDuration    time.Duration      `json:"analysis_duration"`
	Success             bool               `json:"success"`
	Error               string             `json:"error,omitempty"`
}

// KeywordSignal represents a keyword with its relevance and confidence
type KeywordSignal struct {
	Keyword    string  `json:"keyword"`
	Relevance  float64 `json:"relevance"`
	Confidence float64 `json:"confidence"`
	Context    string  `json:"context"`
	Source     string  `json:"source"`
}

// IntegrationIndustrySignal represents an industry-specific signal
// (renamed to avoid conflict with content_relevance_analyzer.go)
type IntegrationIndustrySignal struct {
	Industry   string  `json:"industry"`
	Signal     string  `json:"signal"`
	Strength   float64 `json:"strength"`
	Confidence float64 `json:"confidence"`
	Context    string  `json:"context"`
	Source     string  `json:"source"`
}

// StructuredDataResult represents structured data extraction results
type StructuredDataResult struct {
	BusinessInfo    IntegrationBusinessInfo             `json:"business_info"`
	ExtractionScore float64                  `json:"extraction_score"`
	SchemaOrgData   map[string]interface{}   `json:"schema_org_data"`
	OpenGraphData   map[string]interface{}   `json:"open_graph_data"`
	TwitterCardData map[string]interface{}   `json:"twitter_card_data"`
	MicrodataItems  []map[string]interface{} `json:"microdata_items"`
	Success         bool                     `json:"success"`
	Error           string                   `json:"error,omitempty"`
}

// ClassificationService interface for existing classification methods
type ClassificationService interface {
	ClassifyBusiness(ctx context.Context, businessName, websiteURL string) (*ClassificationResult, error)
}

// ClassificationResult represents the existing classification result structure
type ClassificationResult struct {
	ID                      string                 `json:"id"`
	BusinessName            string                 `json:"business_name"`
	PrimaryIndustry         string                 `json:"primary_industry"`
	IndustryConfidence      float64                `json:"industry_confidence"`
	MCCCodes                []IndustryCode         `json:"mcc_codes"`
	SICCodes                []IndustryCode         `json:"sic_codes"`
	NAICSCodes              []IndustryCode         `json:"naics_codes"`
	Keywords                []string               `json:"keywords"`
	ConfidenceScore         float64                `json:"confidence_score"`
	ClassificationReasoning string                 `json:"classification_reasoning"`
	WebsiteAnalysis         *WebsiteAnalysisResult `json:"website_analysis"`
	Metadata                map[string]interface{} `json:"metadata"`
	ProcessingTime          time.Duration          `json:"processing_time"`
	Timestamp               time.Time              `json:"timestamp"`
}

// IndustryCode represents an industry classification code
type IndustryCode struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
	RiskLevel   string  `json:"risk_level,omitempty"`
}

// WebsiteAnalysisResult represents website analysis results
type WebsiteAnalysisResult struct {
	WebsiteURL        string                   `json:"website_url"`
	PagesAnalyzed     int                      `json:"pages_analyzed"`
	RelevantPages     int                      `json:"relevant_pages"`
	KeywordsExtracted []string                 `json:"keywords_extracted"`
	BusinessInfo      *BusinessInfo            `json:"business_info"`
	StructuredData    *StructuredDataResult    `json:"structured_data"`
	ContentRelevance  *RelevanceAnalysisResult `json:"content_relevance"`
	AnalysisMethod    string                   `json:"analysis_method"`
	ProcessingTime    time.Duration            `json:"processing_time"`
	Success           bool                     `json:"success"`
	Error             string                   `json:"error,omitempty"`
}

// NewSmartCrawlingIntegration creates a new smart crawling integration
func NewSmartCrawlingIntegration(logger *log.Logger, existingClassifier ClassificationService) *SmartCrawlingIntegration {
	return &SmartCrawlingIntegration{
		logger:             logger,
		enhancedAnalyzer:   NewIntegrationEnhancedWebsiteAnalyzer(logger),
		existingClassifier: existingClassifier,
	}
}

// ClassifyBusinessWithSmartCrawling performs classification using smart crawling + existing methods
func (sci *SmartCrawlingIntegration) ClassifyBusinessWithSmartCrawling(ctx context.Context, businessName, websiteURL string) (*ClassificationResult, error) {
	startTime := time.Now()
	sci.logger.Printf("🚀 [SmartCrawlingIntegration] Starting enhanced classification for: %s", businessName)

	// Step 1: Perform existing classification
	sci.logger.Printf("📊 [SmartCrawlingIntegration] Step 1: Existing classification")
	existingResult, err := sci.existingClassifier.ClassifyBusiness(ctx, businessName, websiteURL)
	if err != nil {
		sci.logger.Printf("⚠️ [SmartCrawlingIntegration] Existing classification failed: %v", err)
		// Continue with smart crawling even if existing classification fails
		existingResult = &ClassificationResult{
			BusinessName: businessName,
			Keywords:     []string{},
			MCCCodes:     []IndustryCode{},
			SICCodes:     []IndustryCode{},
			NAICSCodes:   []IndustryCode{},
		}
	}

	// Step 2: Perform smart crawling analysis (if website URL provided)
	var websiteAnalysis *WebsiteAnalysisResult
	if websiteURL != "" {
		sci.logger.Printf("📊 [SmartCrawlingIntegration] Step 2: Smart crawling analysis")
		enhancedResult, err := sci.enhancedAnalyzer.AnalyzeWebsite(ctx, websiteURL)
		if err != nil {
			sci.logger.Printf("⚠️ [SmartCrawlingIntegration] Smart crawling failed: %v", err)
			// Continue with existing classification
		} else {
			websiteAnalysis = sci.convertEnhancedResultToWebsiteAnalysis(enhancedResult)
		}
	}

	// Step 3: Merge and enhance results
	sci.logger.Printf("📊 [SmartCrawlingIntegration] Step 3: Merging results")
	enhancedResult := sci.mergeClassificationResults(existingResult, websiteAnalysis, businessName, websiteURL)

	// Step 4: Generate enhanced classification reasoning
	sci.logger.Printf("📊 [SmartCrawlingIntegration] Step 4: Generating reasoning")
	enhancedResult.ClassificationReasoning = sci.generateClassificationReasoning(enhancedResult, websiteAnalysis)

	enhancedResult.ProcessingTime = time.Since(startTime)
	enhancedResult.Timestamp = time.Now()

	sci.logger.Printf("✅ [SmartCrawlingIntegration] Enhanced classification completed in %v", enhancedResult.ProcessingTime)
	return enhancedResult, nil
}

// convertEnhancedResultToWebsiteAnalysis converts enhanced analysis result to website analysis format
func (sci *SmartCrawlingIntegration) convertEnhancedResultToWebsiteAnalysis(enhancedResult *IntegrationEnhancedAnalysisResult) *WebsiteAnalysisResult {
	if enhancedResult == nil {
		return nil
	}

	websiteAnalysis := &WebsiteAnalysisResult{
		WebsiteURL:     enhancedResult.WebsiteURL,
		AnalysisMethod: "smart_crawling",
		ProcessingTime: enhancedResult.ProcessingTime,
		Success:        enhancedResult.Success,
		Error:          enhancedResult.Error,
	}

	if enhancedResult.CrawlResult != nil {
		websiteAnalysis.PagesAnalyzed = enhancedResult.CrawlResult.TotalPages
		websiteAnalysis.RelevantPages = enhancedResult.CrawlResult.RelevantPages
		websiteAnalysis.KeywordsExtracted = enhancedResult.CrawlResult.Keywords
		// Convert IntegrationBusinessInfo to BusinessInfo (from content_relevance_analyzer.go)
		businessInfo := BusinessInfo{
			BusinessName:  enhancedResult.CrawlResult.BusinessInfo.BusinessName,
			Description:   enhancedResult.CrawlResult.BusinessInfo.Description,
			Services:      enhancedResult.CrawlResult.BusinessInfo.Services,
			Products:      enhancedResult.CrawlResult.BusinessInfo.Products,
			ContactInfo: ContactInfo{
				Phone:   enhancedResult.CrawlResult.BusinessInfo.ContactInfo.Phone,
				Email:   enhancedResult.CrawlResult.BusinessInfo.ContactInfo.Email,
				Address: enhancedResult.CrawlResult.BusinessInfo.ContactInfo.Address,
				Website: enhancedResult.CrawlResult.BusinessInfo.ContactInfo.Website,
				Social:  enhancedResult.CrawlResult.BusinessInfo.ContactInfo.Social,
			},
			BusinessHours: enhancedResult.CrawlResult.BusinessInfo.BusinessHours,
			Location:      enhancedResult.CrawlResult.BusinessInfo.Location,
			Industry:      enhancedResult.CrawlResult.BusinessInfo.Industry,
			BusinessType:  enhancedResult.CrawlResult.BusinessInfo.BusinessType,
		}
		websiteAnalysis.BusinessInfo = &businessInfo
	}

	if enhancedResult.StructuredData != nil {
		websiteAnalysis.StructuredData = enhancedResult.StructuredData
	}

	if enhancedResult.RelevanceAnalysis != nil {
		// Convert IntegrationRelevanceAnalysisResult to RelevanceAnalysisResult
		// For now, create a minimal conversion - full conversion would require mapping all fields
		// This is a placeholder - the actual conversion would need to map all fields properly
		websiteAnalysis.ContentRelevance = nil // TODO: Convert IntegrationRelevanceAnalysisResult to RelevanceAnalysisResult
	}

	return websiteAnalysis
}

// mergeClassificationResults merges existing classification with smart crawling results
func (sci *SmartCrawlingIntegration) mergeClassificationResults(existing *ClassificationResult, websiteAnalysis *WebsiteAnalysisResult, businessName, websiteURL string) *ClassificationResult {
	enhanced := &ClassificationResult{
		ID:                 existing.ID,
		BusinessName:       businessName,
		PrimaryIndustry:    existing.PrimaryIndustry,
		IndustryConfidence: existing.IndustryConfidence,
		MCCCodes:           existing.MCCCodes,
		SICCodes:           existing.SICCodes,
		NAICSCodes:         existing.NAICSCodes,
		Keywords:           existing.Keywords,
		ConfidenceScore:    existing.ConfidenceScore,
		WebsiteAnalysis:    websiteAnalysis,
		Metadata:           existing.Metadata,
		Timestamp:          time.Now(),
	}

	// Enhance with smart crawling data if available
	if websiteAnalysis != nil && websiteAnalysis.Success {
		// Merge keywords
		enhanced.Keywords = sci.mergeKeywords(existing.Keywords, websiteAnalysis.KeywordsExtracted)

		// Enhance classification codes with smart crawling insights
		if websiteAnalysis.StructuredData != nil && websiteAnalysis.StructuredData.BusinessInfo.BusinessName != "" {
			enhanced.PrimaryIndustry = sci.determinePrimaryIndustry(existing.PrimaryIndustry, websiteAnalysis)
			enhanced.IndustryConfidence = sci.calculateEnhancedConfidence(existing.IndustryConfidence, websiteAnalysis)
		}

		// Add website analysis metadata
		if enhanced.Metadata == nil {
			enhanced.Metadata = make(map[string]interface{})
		}
		enhanced.Metadata["smart_crawling"] = map[string]interface{}{
			"pages_analyzed":     websiteAnalysis.PagesAnalyzed,
			"relevant_pages":     websiteAnalysis.RelevantPages,
			"keywords_extracted": len(websiteAnalysis.KeywordsExtracted),
			"analysis_method":    websiteAnalysis.AnalysisMethod,
			"processing_time":    websiteAnalysis.ProcessingTime.String(),
		}
	}

	return enhanced
}

// mergeKeywords merges existing keywords with smart crawling keywords
func (sci *SmartCrawlingIntegration) mergeKeywords(existingKeywords, crawledKeywords []string) []string {
	keywordMap := make(map[string]bool)
	var mergedKeywords []string

	// Add existing keywords
	for _, keyword := range existingKeywords {
		if !keywordMap[keyword] {
			keywordMap[keyword] = true
			mergedKeywords = append(mergedKeywords, keyword)
		}
	}

	// Add crawled keywords
	for _, keyword := range crawledKeywords {
		if !keywordMap[keyword] {
			keywordMap[keyword] = true
			mergedKeywords = append(mergedKeywords, keyword)
		}
	}

	return mergedKeywords
}

// determinePrimaryIndustry determines the best primary industry from existing and smart crawling data
func (sci *SmartCrawlingIntegration) determinePrimaryIndustry(existingIndustry string, websiteAnalysis *WebsiteAnalysisResult) string {
	// If existing industry is empty or generic, use smart crawling insights
	if existingIndustry == "" || existingIndustry == "Unknown" || existingIndustry == "General" {
		if websiteAnalysis.StructuredData != nil && websiteAnalysis.StructuredData.BusinessInfo.Industry != "" {
			return websiteAnalysis.StructuredData.BusinessInfo.Industry
		}
		if websiteAnalysis.ContentRelevance != nil && len(websiteAnalysis.ContentRelevance.IndustrySignals) > 0 {
			// Get the strongest industry signal
			strongestSignal := websiteAnalysis.ContentRelevance.IndustrySignals[0]
			return sci.mapIndustrySignalToIndustry(strongestSignal.Industry)
		}
	}

	// Use existing industry if it's specific and confident
	return existingIndustry
}

// mapIndustrySignalToIndustry maps industry signals to readable industry names
func (sci *SmartCrawlingIntegration) mapIndustrySignalToIndustry(signal string) string {
	industryMap := map[string]string{
		"food_beverage": "Food & Beverage",
		"technology":    "Technology",
		"healthcare":    "Healthcare",
		"legal":         "Legal Services",
		"retail":        "Retail",
		"finance":       "Financial Services",
		"real_estate":   "Real Estate",
		"education":     "Education",
		"consulting":    "Consulting",
		"manufacturing": "Manufacturing",
	}

	if industry, exists := industryMap[signal]; exists {
		return industry
	}
	return strings.Title(strings.ReplaceAll(signal, "_", " "))
}

// calculateEnhancedConfidence calculates enhanced confidence score
func (sci *SmartCrawlingIntegration) calculateEnhancedConfidence(existingConfidence float64, websiteAnalysis *WebsiteAnalysisResult) float64 {
	// Start with existing confidence
	enhancedConfidence := existingConfidence

	// Boost confidence if smart crawling provides additional evidence
	if websiteAnalysis != nil && websiteAnalysis.Success {
		// Boost for structured data
		if websiteAnalysis.StructuredData != nil && websiteAnalysis.StructuredData.ExtractionScore > 0.5 {
			enhancedConfidence += 0.1
		}

		// Boost for content relevance
		if websiteAnalysis.ContentRelevance != nil && websiteAnalysis.ContentRelevance.ConfidenceScore > 0.7 {
			enhancedConfidence += 0.1
		}

		// Boost for multiple pages analyzed
		if websiteAnalysis.PagesAnalyzed > 3 {
			enhancedConfidence += 0.05
		}

		// Boost for relevant pages
		if websiteAnalysis.RelevantPages > 0 {
			enhancedConfidence += 0.05
		}
	}

	// Cap at 1.0
	if enhancedConfidence > 1.0 {
		enhancedConfidence = 1.0
	}

	return enhancedConfidence
}

// generateClassificationReasoning generates comprehensive classification reasoning
func (sci *SmartCrawlingIntegration) generateClassificationReasoning(result *ClassificationResult, websiteAnalysis *WebsiteAnalysisResult) string {
	var reasoningParts []string

	// Base reasoning from existing classification
	if result.PrimaryIndustry != "" {
		reasoningParts = append(reasoningParts, fmt.Sprintf("Primary industry identified as '%s' with %.0f%% confidence",
			result.PrimaryIndustry, result.IndustryConfidence*100))
	}

	// Add smart crawling insights
	if websiteAnalysis != nil && websiteAnalysis.Success {
		reasoningParts = append(reasoningParts, fmt.Sprintf("Website analysis of %s analyzed %d pages with %d relevant pages",
			websiteAnalysis.WebsiteURL, websiteAnalysis.PagesAnalyzed, websiteAnalysis.RelevantPages))

		// Add structured data insights
		if websiteAnalysis.StructuredData != nil && websiteAnalysis.StructuredData.BusinessInfo.BusinessName != "" {
			reasoningParts = append(reasoningParts, fmt.Sprintf("Structured data extraction found business name '%s' and industry '%s'",
				websiteAnalysis.StructuredData.BusinessInfo.BusinessName, websiteAnalysis.StructuredData.BusinessInfo.Industry))
		}

		// Add keyword insights
		if len(websiteAnalysis.KeywordsExtracted) > 0 {
			keywordList := strings.Join(websiteAnalysis.KeywordsExtracted[:min(5, len(websiteAnalysis.KeywordsExtracted))], ", ")
			reasoningParts = append(reasoningParts, fmt.Sprintf("Website keywords extracted: %s", keywordList))
		}

		// Add industry signal insights
		if websiteAnalysis.ContentRelevance != nil && len(websiteAnalysis.ContentRelevance.IndustrySignals) > 0 {
			strongestSignal := websiteAnalysis.ContentRelevance.IndustrySignals[0]
			reasoningParts = append(reasoningParts, fmt.Sprintf("Industry signal detection identified '%s' with %.0f%% strength",
				strongestSignal.Industry, strongestSignal.Strength*100))
		}
	}

	// Add classification method insights
	if len(result.Keywords) > 0 {
		reasoningParts = append(reasoningParts, fmt.Sprintf("Classification based on %d keywords and industry pattern matching", len(result.Keywords)))
	}

	// Add confidence insights
	if result.ConfidenceScore > 0.8 {
		reasoningParts = append(reasoningParts, "High confidence classification based on multiple data sources")
	} else if result.ConfidenceScore > 0.6 {
		reasoningParts = append(reasoningParts, "Moderate confidence classification with some uncertainty")
	} else {
		reasoningParts = append(reasoningParts, "Lower confidence classification requiring manual review")
	}

	// Combine all reasoning parts
	if len(reasoningParts) == 0 {
		return "Classification based on business name analysis and industry pattern recognition."
	}

	return strings.Join(reasoningParts, ". ") + "."
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
