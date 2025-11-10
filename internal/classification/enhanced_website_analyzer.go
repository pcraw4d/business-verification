package classification

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"
)

// EnhancedWebsiteAnalyzer integrates smart crawling, content analysis, and structured data extraction
type EnhancedWebsiteAnalyzer struct {
	logger                  *log.Logger
	smartCrawler            *SmartWebsiteCrawler
	relevanceAnalyzer       *ContentRelevanceAnalyzer
	structuredDataExtractor *StructuredDataExtractor
}

// WebsiteClassificationCode represents a classification code (MCC, SIC, NAICS)
// (renamed to avoid conflict with crosswalk_types.go)
type WebsiteClassificationCode struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// EnhancedAnalysisResult represents the complete enhanced analysis result
type EnhancedAnalysisResult struct {
	WebsiteURL             string                        `json:"website_url"`
	CrawlResult            *CrawlResult                  `json:"crawl_result"`
	RelevanceAnalysis      *RelevanceAnalysisResult      `json:"relevance_analysis"`
	StructuredData         *ExtractorStructuredDataResult         `json:"structured_data"`
	BusinessClassification *BusinessClassificationResult `json:"business_classification"`
	AnalysisTimestamp      time.Time                     `json:"analysis_timestamp"`
	ProcessingTime         time.Duration                 `json:"processing_time"`
	OverallConfidence      float64                       `json:"overall_confidence"`
	Success                bool                          `json:"success"`
	Error                  string                        `json:"error,omitempty"`
}

// BusinessClassificationResult represents the final business classification
type BusinessClassificationResult struct {
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

// NewEnhancedWebsiteAnalyzer creates a new enhanced website analyzer
func NewEnhancedWebsiteAnalyzer(logger *log.Logger) *EnhancedWebsiteAnalyzer {
	return &EnhancedWebsiteAnalyzer{
		logger:                  logger,
		smartCrawler:            NewSmartWebsiteCrawler(logger),
		relevanceAnalyzer:       NewContentRelevanceAnalyzer(logger),
		structuredDataExtractor: NewStructuredDataExtractor(logger),
	}
}

// AnalyzeWebsite performs comprehensive website analysis
func (ewa *EnhancedWebsiteAnalyzer) AnalyzeWebsite(ctx context.Context, websiteURL string) (*EnhancedAnalysisResult, error) {
	startTime := time.Now()
	ewa.logger.Printf("🚀 [EnhancedAnalyzer] Starting enhanced website analysis for: %s", websiteURL)

	result := &EnhancedAnalysisResult{
		WebsiteURL:        websiteURL,
		AnalysisTimestamp: time.Now(),
		Success:           false,
	}

	// Step 1: Smart crawling
	ewa.logger.Printf("📊 [EnhancedAnalyzer] Step 1: Smart crawling")
	crawlResult, err := ewa.smartCrawler.CrawlWebsite(ctx, websiteURL)
	if err != nil {
		result.Error = fmt.Sprintf("Smart crawling failed: %v", err)
		return result, err
	}
	result.CrawlResult = crawlResult

	// Step 2: Content relevance analysis
	ewa.logger.Printf("📊 [EnhancedAnalyzer] Step 2: Content relevance analysis")
	relevanceAnalysis, err := ewa.relevanceAnalyzer.AnalyzeContentRelevance(ctx, crawlResult)
	if err != nil {
		result.Error = fmt.Sprintf("Relevance analysis failed: %v", err)
		return result, err
	}
	result.RelevanceAnalysis = relevanceAnalysis

	// Step 3: Structured data extraction (from most relevant pages)
	ewa.logger.Printf("📊 [EnhancedAnalyzer] Step 3: Structured data extraction")
	structuredData := ewa.extractStructuredDataFromRelevantPages(crawlResult, relevanceAnalysis)
	result.StructuredData = structuredData

	// Step 4: Business classification
	ewa.logger.Printf("📊 [EnhancedAnalyzer] Step 4: Business classification")
	businessClassification := ewa.performBusinessClassification(crawlResult, relevanceAnalysis, structuredData)
	result.BusinessClassification = businessClassification

	// Calculate overall confidence
	result.OverallConfidence = ewa.calculateOverallConfidence(result)
	result.ProcessingTime = time.Since(startTime)
	result.Success = true

	ewa.logger.Printf("✅ [EnhancedAnalyzer] Enhanced analysis completed in %v - Confidence: %.2f",
		result.ProcessingTime, result.OverallConfidence)

	return result, nil
}

// extractStructuredDataFromRelevantPages extracts structured data from the most relevant pages
func (ewa *EnhancedWebsiteAnalyzer) extractStructuredDataFromRelevantPages(crawlResult *CrawlResult, relevanceAnalysis *RelevanceAnalysisResult) *ExtractorStructuredDataResult {
	// Find the most relevant pages (top 3)
	var relevantPages []PageAnalysis
	for _, page := range crawlResult.PagesAnalyzed {
		if relevance, exists := relevanceAnalysis.PageRelevance[page.URL]; exists && relevance > 0.7 {
			relevantPages = append(relevantPages, page)
		}
	}

	// Sort by relevance score
	sort.Slice(relevantPages, func(i, j int) bool {
		scoreI := relevanceAnalysis.PageRelevance[relevantPages[i].URL]
		scoreJ := relevanceAnalysis.PageRelevance[relevantPages[j].URL]
		return scoreI > scoreJ
	})

	// Limit to top 3 pages
	if len(relevantPages) > 3 {
		relevantPages = relevantPages[:3]
	}

	// Extract structured data from relevant pages
	// For now, we'll simulate this - in a real implementation, you'd re-fetch the pages
	// and extract structured data from their HTML content
	structuredData := &ExtractorStructuredDataResult{
		SchemaOrgData:   []SchemaOrgItem{},
		OpenGraphData:   make(map[string]string),
		TwitterCardData: make(map[string]string),
		Microdata:       []MicrodataItem{},
		BusinessInfo:    ExtractorBusinessInfo{},
		ProductInfo:     []ProductInfo{},
		ServiceInfo:     []ServiceInfo{},
		EventInfo:       []EventInfo{},
		ExtractionScore: 0.0,
	}

	// Aggregate business information from relevant pages
	for _, page := range relevantPages {
		if page.BusinessInfo.BusinessName != "" && structuredData.BusinessInfo.BusinessName == "" {
			structuredData.BusinessInfo = ExtractorBusinessInfo{
				BusinessName:  page.BusinessInfo.BusinessName,
				Description:   page.BusinessInfo.Description,
				Services:      page.BusinessInfo.Services,
				Products:      page.BusinessInfo.Products,
				ContactInfo:   ExtractorContactInfo{
					Phone:   page.BusinessInfo.ContactInfo.Phone,
					Email:   page.BusinessInfo.ContactInfo.Email,
					Address: page.BusinessInfo.ContactInfo.Address,
					Website: page.BusinessInfo.ContactInfo.Website,
					Social:  page.BusinessInfo.ContactInfo.Social,
				},
				BusinessHours: page.BusinessInfo.BusinessHours,
				Location:      page.BusinessInfo.Location,
				Industry:      page.BusinessInfo.Industry,
				BusinessType:  page.BusinessInfo.BusinessType,
			}
		}
	}

	ewa.logger.Printf("📊 [EnhancedAnalyzer] Extracted structured data from %d relevant pages", len(relevantPages))
	return structuredData
}

// performBusinessClassification performs business classification based on all analysis results
func (ewa *EnhancedWebsiteAnalyzer) performBusinessClassification(crawlResult *CrawlResult, relevanceAnalysis *RelevanceAnalysisResult, structuredData *ExtractorStructuredDataResult) *BusinessClassificationResult {
	classification := &BusinessClassificationResult{
		MCCCodes:       []WebsiteClassificationCode{},
		SICCodes:       []WebsiteClassificationCode{},
		NAICSCodes:     []WebsiteClassificationCode{},
		Keywords:       []string{},
		AnalysisMethod: "enhanced_website_analysis",
	}

	// Determine primary industry from industry signals
	if len(relevanceAnalysis.IndustrySignals) > 0 {
		// Sort industry signals by strength
		sort.Slice(relevanceAnalysis.IndustrySignals, func(i, j int) bool {
			return relevanceAnalysis.IndustrySignals[i].Strength > relevanceAnalysis.IndustrySignals[j].Strength
		})

		primarySignal := relevanceAnalysis.IndustrySignals[0]
		classification.PrimaryIndustry = primarySignal.Industry
		classification.IndustryConfidence = primarySignal.Confidence
	}

	// Determine business type from structured data
	if structuredData.BusinessInfo.BusinessType != "" {
		classification.BusinessType = structuredData.BusinessInfo.BusinessType
		classification.BusinessTypeConfidence = 0.8
	}

	// Generate classification codes based on industry
	classification.MCCCodes, classification.SICCodes, classification.NAICSCodes = ewa.generateClassificationCodes(classification.PrimaryIndustry)

	// Aggregate keywords from all sources
	keywordMap := make(map[string]int)

	// From crawl result
	for _, keyword := range crawlResult.Keywords {
		keywordMap[keyword]++
	}

	// From business indicators
	for _, indicator := range relevanceAnalysis.BusinessIndicators {
		keywordMap[indicator.Value]++
	}

	// Sort keywords by frequency
	type keywordFreq struct {
		keyword string
		count   int
	}

	var sortedKeywords []keywordFreq
	for keyword, count := range keywordMap {
		sortedKeywords = append(sortedKeywords, keywordFreq{keyword, count})
	}

	sort.Slice(sortedKeywords, func(i, j int) bool {
		return sortedKeywords[i].count > sortedKeywords[j].count
	})

	// Take top keywords
	for i, kf := range sortedKeywords {
		if i >= 15 { // Limit to top 15 keywords
			break
		}
		classification.Keywords = append(classification.Keywords, kf.keyword)
	}

	// Calculate overall confidence score
	classification.ConfidenceScore = ewa.calculateClassificationConfidence(classification, relevanceAnalysis)

	return classification
}

// generateClassificationCodes generates classification codes based on industry
func (ewa *EnhancedWebsiteAnalyzer) generateClassificationCodes(industry string) ([]WebsiteClassificationCode, []WebsiteClassificationCode, []WebsiteClassificationCode) {
	var mccCodes, sicCodes, naicsCodes []WebsiteClassificationCode

	// Industry-specific code mappings
	industryMappings := map[string]struct {
		mcc   []WebsiteClassificationCode
		sic   []WebsiteClassificationCode
		naics []WebsiteClassificationCode
	}{
		"food_beverage": {
			mcc: []WebsiteClassificationCode{
				{Code: "5813", Description: "Drinking Places (Alcoholic Beverages)", Confidence: 0.95},
				{Code: "5814", Description: "Fast Food Restaurants", Confidence: 0.85},
				{Code: "5411", Description: "Grocery Stores, Supermarkets", Confidence: 0.75},
			},
			sic: []WebsiteClassificationCode{
				{Code: "5813", Description: "Drinking Places (Alcoholic Beverages)", Confidence: 0.95},
				{Code: "5812", Description: "Eating Places", Confidence: 0.85},
				{Code: "5411", Description: "Grocery Stores", Confidence: 0.75},
			},
			naics: []WebsiteClassificationCode{
				{Code: "445310", Description: "Beer, Wine, and Liquor Stores", Confidence: 0.95},
				{Code: "722511", Description: "Full-Service Restaurants", Confidence: 0.85},
				{Code: "445110", Description: "Supermarkets and Other Grocery Stores", Confidence: 0.75},
			},
		},
		"technology": {
			mcc: []WebsiteClassificationCode{
				{Code: "7372", Description: "Computer Programming Services", Confidence: 0.95},
				{Code: "7371", Description: "Computer Programming Services", Confidence: 0.85},
				{Code: "7379", Description: "Computer Related Services", Confidence: 0.75},
			},
			sic: []WebsiteClassificationCode{
				{Code: "7372", Description: "Computer Programming Services", Confidence: 0.95},
				{Code: "7371", Description: "Computer Programming Services", Confidence: 0.85},
				{Code: "7379", Description: "Computer Related Services", Confidence: 0.75},
			},
			naics: []WebsiteClassificationCode{
				{Code: "541511", Description: "Custom Computer Programming Services", Confidence: 0.95},
				{Code: "541512", Description: "Computer Systems Design Services", Confidence: 0.85},
				{Code: "541519", Description: "Other Computer Related Services", Confidence: 0.75},
			},
		},
		"healthcare": {
			mcc: []WebsiteClassificationCode{
				{Code: "8011", Description: "Doctors and Physicians", Confidence: 0.95},
				{Code: "8021", Description: "Dentists and Orthodontists", Confidence: 0.85},
				{Code: "8041", Description: "Chiropractors", Confidence: 0.75},
			},
			sic: []WebsiteClassificationCode{
				{Code: "8011", Description: "Offices and Clinics of Doctors of Medicine", Confidence: 0.95},
				{Code: "8021", Description: "Offices and Clinics of Dentists", Confidence: 0.85},
				{Code: "8041", Description: "Offices and Clinics of Chiropractors", Confidence: 0.75},
			},
			naics: []WebsiteClassificationCode{
				{Code: "621111", Description: "Offices of Physicians (except Mental Health Specialists)", Confidence: 0.95},
				{Code: "621210", Description: "Offices of Dentists", Confidence: 0.85},
				{Code: "621310", Description: "Offices of Chiropractors", Confidence: 0.75},
			},
		},
		"retail": {
			mcc: []WebsiteClassificationCode{
				{Code: "5310", Description: "Department Stores", Confidence: 0.95},
				{Code: "5311", Description: "Department Stores", Confidence: 0.85},
				{Code: "5411", Description: "Grocery Stores, Supermarkets", Confidence: 0.75},
			},
			sic: []WebsiteClassificationCode{
				{Code: "5311", Description: "Department Stores", Confidence: 0.95},
				{Code: "5312", Description: "Variety Stores", Confidence: 0.85},
				{Code: "5411", Description: "Grocery Stores", Confidence: 0.75},
			},
			naics: []WebsiteClassificationCode{
				{Code: "452111", Description: "Department Stores", Confidence: 0.95},
				{Code: "452112", Description: "Discount Department Stores", Confidence: 0.85},
				{Code: "445110", Description: "Supermarkets and Other Grocery Stores", Confidence: 0.75},
			},
		},
	}

	if mapping, exists := industryMappings[industry]; exists {
		mccCodes = mapping.mcc
		sicCodes = mapping.sic
		naicsCodes = mapping.naics
	} else {
		// Default to technology if industry not found
		mapping := industryMappings["technology"]
		mccCodes = mapping.mcc
		sicCodes = mapping.sic
		naicsCodes = mapping.naics
	}

	return mccCodes, sicCodes, naicsCodes
}

// calculateClassificationConfidence calculates confidence score for business classification
func (ewa *EnhancedWebsiteAnalyzer) calculateClassificationConfidence(classification *BusinessClassificationResult, relevanceAnalysis *RelevanceAnalysisResult) float64 {
	confidence := 0.5 // Base confidence

	// Industry confidence factor
	confidence += classification.IndustryConfidence * 0.3

	// Business type confidence factor
	confidence += classification.BusinessTypeConfidence * 0.2

	// Relevance analysis confidence factor
	confidence += relevanceAnalysis.ConfidenceScore * 0.2

	// Keywords factor
	if len(classification.Keywords) > 10 {
		confidence += 0.1
	} else if len(classification.Keywords) > 5 {
		confidence += 0.05
	}

	// Industry signals factor
	if len(relevanceAnalysis.IndustrySignals) > 3 {
		confidence += 0.1
	} else if len(relevanceAnalysis.IndustrySignals) > 1 {
		confidence += 0.05
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// calculateOverallConfidence calculates overall confidence for the entire analysis
func (ewa *EnhancedWebsiteAnalyzer) calculateOverallConfidence(result *EnhancedAnalysisResult) float64 {
	confidence := 0.0

	// Crawl result confidence
	if result.CrawlResult != nil && result.CrawlResult.Success {
		confidence += 0.3
	}

	// Relevance analysis confidence
	if result.RelevanceAnalysis != nil {
		confidence += result.RelevanceAnalysis.ConfidenceScore * 0.3
	}

	// Structured data confidence
	if result.StructuredData != nil {
		confidence += result.StructuredData.ExtractionScore * 0.2
	}

	// Business classification confidence
	if result.BusinessClassification != nil {
		confidence += result.BusinessClassification.ConfidenceScore * 0.2
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}
