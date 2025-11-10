package classification

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"kyb-platform/internal/classification/repository"
)

// UnifiedClassifier integrates website analysis with existing classification methods
type UnifiedClassifier struct {
	logger           *log.Logger
	websiteAnalyzer  *EnhancedWebsiteAnalyzer
	keywordRepo      repository.KeywordRepository
	scoringAlgorithm *EnhancedScoringAlgorithm
}

// ClassificationInput represents all available input data for classification
type ClassificationInput struct {
	BusinessName    string
	WebsiteURL      string
	Description     string
	WebsiteAnalysis *EnhancedAnalysisResult
}

// UnifiedClassificationResult represents the final classification result
type UnifiedClassificationResult struct {
	BusinessName            string               `json:"business_name"`
	PrimaryIndustry         string               `json:"primary_industry"`
	IndustryConfidence      float64              `json:"industry_confidence"`
	BusinessType            string               `json:"business_type"`
	BusinessTypeConfidence  float64              `json:"business_type_confidence"`
	MCCCodes                []IndustryCode       `json:"mcc_codes"`
	SICCodes                []IndustryCode       `json:"sic_codes"`
	NAICSCodes              []IndustryCode       `json:"naics_codes"`
	Keywords                []string             `json:"keywords"`
	ConfidenceScore         float64              `json:"confidence_score"`
	ClassificationReasoning string               `json:"classification_reasoning"`
	MethodWeights           map[string]float64   `json:"method_weights"`
	DataSources             []DataSource         `json:"data_sources"`
	WebsiteAnalysis         *WebsiteAnalysisData `json:"website_analysis,omitempty"`
	Timestamp               time.Time            `json:"timestamp"`
	ProcessingTime          time.Duration        `json:"processing_time"`
}

// DataSource represents a data source used in classification
type DataSource struct {
	Source      string   `json:"source"`
	Weight      float64  `json:"weight"`
	Confidence  float64  `json:"confidence"`
	Keywords    []string `json:"keywords"`
	Description string   `json:"description"`
}

// WebsiteAnalysisData represents aggregated website analysis data
type WebsiteAnalysisData struct {
	Success           bool                   `json:"success"`
	PagesAnalyzed     int                    `json:"pages_analyzed"`
	RelevantPages     int                    `json:"relevant_pages"`
	KeywordsExtracted []string               `json:"keywords_extracted"`
	IndustrySignals   []string               `json:"industry_signals"`
	AnalysisMethod    string                 `json:"analysis_method"`
	ProcessingTime    time.Duration          `json:"processing_time"`
	OverallRelevance  float64                `json:"overall_relevance"`
	ContentQuality    float64                `json:"content_quality"`
	StructuredData    map[string]interface{} `json:"structured_data,omitempty"`
}

// NewUnifiedClassifier creates a new unified classifier
func NewUnifiedClassifier(
	logger *log.Logger,
	websiteAnalyzer *EnhancedWebsiteAnalyzer,
	keywordRepo repository.KeywordRepository,
	scoringAlgorithm *EnhancedScoringAlgorithm,
) *UnifiedClassifier {
	return &UnifiedClassifier{
		logger:           logger,
		websiteAnalyzer:  websiteAnalyzer,
		keywordRepo:      keywordRepo,
		scoringAlgorithm: scoringAlgorithm,
	}
}

// ClassifyBusiness performs unified classification using all available data sources
func (uc *UnifiedClassifier) ClassifyBusiness(ctx context.Context, input *ClassificationInput) (*UnifiedClassificationResult, error) {
	startTime := time.Now()
	uc.logger.Printf("🚀 [UnifiedClassifier] Starting unified classification for: %s", input.BusinessName)

	result := &UnifiedClassificationResult{
		BusinessName:  input.BusinessName,
		Timestamp:     time.Now(),
		MethodWeights: make(map[string]float64),
		DataSources:   []DataSource{},
	}

	// Step 1: Extract data from all sources
	uc.logger.Printf("📊 [UnifiedClassifier] Step 1: Extracting data from all sources")
	dataSources := uc.extractDataFromAllSources(ctx, input)
	result.DataSources = dataSources

	// Step 2: Calculate dynamic weights based on data quality
	uc.logger.Printf("📊 [UnifiedClassifier] Step 2: Calculating dynamic weights")
	weights := uc.calculateDynamicWeights(dataSources)
	result.MethodWeights = weights

	// Step 3: Combine keywords from all sources with weights
	uc.logger.Printf("📊 [UnifiedClassifier] Step 3: Combining keywords with weights")
	combinedKeywords := uc.combineKeywordsWithWeights(dataSources, weights)

	// Step 4: Perform classification using combined data
	uc.logger.Printf("📊 [UnifiedClassifier] Step 4: Performing classification")
	classificationResult, err := uc.performWeightedClassification(ctx, combinedKeywords, weights)
	if err != nil {
		return nil, fmt.Errorf("classification failed: %w", err)
	}

	// Step 5: Generate industry codes
	uc.logger.Printf("📊 [UnifiedClassifier] Step 5: Generating industry codes")
	industryCodes := uc.generateIndustryCodes(ctx, classificationResult, combinedKeywords)

	// Step 6: Build final result
	result.PrimaryIndustry = classificationResult.Industry.Name
	result.IndustryConfidence = classificationResult.Confidence
	result.BusinessType = uc.determineBusinessType(combinedKeywords, input.WebsiteAnalysis)
	result.BusinessTypeConfidence = uc.calculateBusinessTypeConfidence(combinedKeywords, input.WebsiteAnalysis)
	result.MCCCodes = industryCodes.MCC
	result.SICCodes = industryCodes.SIC
	result.NAICSCodes = industryCodes.NAICS
	result.Keywords = uc.extractTopKeywords(combinedKeywords)
	result.ConfidenceScore = uc.calculateOverallConfidence(classificationResult, weights, dataSources)
	result.ClassificationReasoning = uc.generateClassificationReasoning(result, dataSources, weights)
	result.WebsiteAnalysis = uc.convertWebsiteAnalysis(input.WebsiteAnalysis)
	result.ProcessingTime = time.Since(startTime)

	uc.logger.Printf("✅ [UnifiedClassifier] Unified classification completed in %v - Confidence: %.2f",
		result.ProcessingTime, result.ConfidenceScore)

	return result, nil
}

// extractDataFromAllSources extracts data from all available sources
func (uc *UnifiedClassifier) extractDataFromAllSources(ctx context.Context, input *ClassificationInput) []DataSource {
	sources := []DataSource{}

	// Source 1: Business Name Analysis
	nameKeywords := uc.extractKeywordsFromBusinessName(input.BusinessName)
	if len(nameKeywords) > 0 {
		sources = append(sources, DataSource{
			Source:      "business_name",
			Weight:      0.0, // Will be calculated dynamically
			Confidence:  uc.calculateNameConfidence(input.BusinessName),
			Keywords:    nameKeywords,
			Description: fmt.Sprintf("Extracted %d keywords from business name", len(nameKeywords)),
		})
	}

	// Source 2: Website URL Analysis
	if input.WebsiteURL != "" {
		urlKeywords := uc.extractKeywordsFromURL(input.WebsiteURL)
		if len(urlKeywords) > 0 {
			sources = append(sources, DataSource{
				Source:      "website_url",
				Weight:      0.0, // Will be calculated dynamically
				Confidence:  uc.calculateURLConfidence(input.WebsiteURL),
				Keywords:    urlKeywords,
				Description: fmt.Sprintf("Extracted %d keywords from website URL", len(urlKeywords)),
			})
		}
	}

	// Source 3: Website Content Analysis
	if input.WebsiteAnalysis != nil && input.WebsiteAnalysis.Success {
		websiteKeywords := uc.extractKeywordsFromWebsiteAnalysis(input.WebsiteAnalysis)
		industrySignals := uc.extractIndustrySignalsFromWebsiteAnalysis(input.WebsiteAnalysis)

		allWebsiteKeywords := append(websiteKeywords, industrySignals...)
		if len(allWebsiteKeywords) > 0 {
			sources = append(sources, DataSource{
				Source:      "website_content",
				Weight:      0.0, // Will be calculated dynamically
				Confidence:  input.WebsiteAnalysis.OverallConfidence,
				Keywords:    allWebsiteKeywords,
				Description: fmt.Sprintf("Extracted %d keywords from %d analyzed pages", len(allWebsiteKeywords), input.WebsiteAnalysis.CrawlResult.PagesAnalyzed),
			})
		}
	}

	// Source 4: Structured Data Analysis
	if input.WebsiteAnalysis != nil && input.WebsiteAnalysis.StructuredData != nil {
		structuredKeywords := uc.extractKeywordsFromStructuredData(input.WebsiteAnalysis.StructuredData)
		if len(structuredKeywords) > 0 {
			sources = append(sources, DataSource{
				Source:      "structured_data",
				Weight:      0.0, // Will be calculated dynamically
				Confidence:  uc.calculateStructuredDataConfidence(input.WebsiteAnalysis.StructuredData),
				Keywords:    structuredKeywords,
				Description: fmt.Sprintf("Extracted %d keywords from structured data", len(structuredKeywords)),
			})
		}
	}

	uc.logger.Printf("📊 [UnifiedClassifier] Extracted data from %d sources", len(sources))
	return sources
}

// calculateDynamicWeights calculates weights based on data quality and confidence
func (uc *UnifiedClassifier) calculateDynamicWeights(sources []DataSource) map[string]float64 {
	weights := make(map[string]float64)

	if len(sources) == 0 {
		return weights
	}

	// Calculate base weights based on confidence and data quality
	totalWeight := 0.0
	for _, source := range sources {
		baseWeight := uc.calculateBaseWeight(source)
		weights[source.Source] = baseWeight
		totalWeight += baseWeight
	}

	// Normalize weights to sum to 100%
	if totalWeight > 0 {
		for source := range weights {
			weights[source] = (weights[source] / totalWeight) * 100
		}
	}

	uc.logger.Printf("📊 [UnifiedClassifier] Calculated weights: %+v", weights)
	return weights
}

// calculateBaseWeight calculates the base weight for a data source
func (uc *UnifiedClassifier) calculateBaseWeight(source DataSource) float64 {
	baseWeight := source.Confidence

	// Adjust weight based on source type and data quality
	switch source.Source {
	case "website_content":
		// Website content gets higher weight when available and high quality
		baseWeight *= 1.5
		if len(source.Keywords) > 10 {
			baseWeight *= 1.2 // Bonus for rich content
		}
	case "structured_data":
		// Structured data is highly reliable when available
		baseWeight *= 1.3
	case "business_name":
		// Business name is always available but may be ambiguous
		baseWeight *= 0.8
		if len(source.Keywords) > 3 {
			baseWeight *= 1.1 // Bonus for descriptive names
		}
	case "website_url":
		// URL provides some context but limited
		baseWeight *= 0.6
	}

	// Ensure weight is between 0 and 1
	if baseWeight > 1.0 {
		baseWeight = 1.0
	}
	if baseWeight < 0.1 {
		baseWeight = 0.1
	}

	return baseWeight
}

// combineKeywordsWithWeights combines keywords from all sources with their weights
func (uc *UnifiedClassifier) combineKeywordsWithWeights(sources []DataSource, weights map[string]float64) []ContextualKeyword {
	keywordMap := make(map[string]float64)
	keywordSources := make(map[string][]string)

	// Combine keywords with weights
	for _, source := range sources {
		weight := weights[source.Source] / 100.0 // Convert percentage to decimal
		for _, keyword := range source.Keywords {
			normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
			if normalizedKeyword != "" {
				keywordMap[normalizedKeyword] += weight * source.Confidence
				keywordSources[normalizedKeyword] = append(keywordSources[normalizedKeyword], source.Source)
			}
		}
	}

	// Convert to ContextualKeyword slice
	var contextualKeywords []ContextualKeyword
	for keyword, score := range keywordMap {
		contextualKeywords = append(contextualKeywords, ContextualKeyword{
			Keyword: keyword,
			Weight:  score, // Use Weight instead of Score
			Context: strings.Join(keywordSources[keyword], ","), // Use Context instead of Source
		})
	}

	// Sort by weight (highest first)
	sort.Slice(contextualKeywords, func(i, j int) bool {
		return contextualKeywords[i].Weight > contextualKeywords[j].Weight
	})

	uc.logger.Printf("📊 [UnifiedClassifier] Combined %d unique keywords from %d sources", len(contextualKeywords), len(sources))
	return contextualKeywords
}

// performWeightedClassification performs classification using weighted keywords
func (uc *UnifiedClassifier) performWeightedClassification(ctx context.Context, keywords []ContextualKeyword, weights map[string]float64) (*repository.ClassificationResult, error) {
	if len(keywords) == 0 {
		return &repository.ClassificationResult{
			Industry:   &repository.Industry{Name: "General Business", ID: 26},
			Confidence: 0.50,
			Keywords:   []string{},
			Patterns:   []string{},
			Codes:      []repository.ClassificationCode{},
			Reasoning:  "No keywords available for classification",
		}, nil
	}

	// Convert ContextualKeyword to string slice for ClassifyBusinessByKeywords
	keywordStrings := make([]string, len(keywords))
	for i, kw := range keywords {
		keywordStrings[i] = kw.Keyword
	}
	// Use the existing classification system with keywords
	return uc.keywordRepo.ClassifyBusinessByKeywords(ctx, keywordStrings)
}

// generateIndustryCodes generates top 3 codes for each industry classification type
func (uc *UnifiedClassifier) generateIndustryCodes(ctx context.Context, classificationResult *repository.ClassificationResult, keywords []ContextualKeyword) struct {
	MCC   []IndustryCode
	SIC   []IndustryCode
	NAICS []IndustryCode
} {
	// Get industry ID from classification result
	industryID := classificationResult.Industry.ID

	// Generate codes for each type
	mccCodes := uc.getTopIndustryCodes(ctx, industryID, "MCC", 3)
	sicCodes := uc.getTopIndustryCodes(ctx, industryID, "SIC", 3)
	naicsCodes := uc.getTopIndustryCodes(ctx, industryID, "NAICS", 3)

	return struct {
		MCC   []IndustryCode
		SIC   []IndustryCode
		NAICS []IndustryCode
	}{
		MCC:   mccCodes,
		SIC:   sicCodes,
		NAICS: naicsCodes,
	}
}

// getTopIndustryCodes retrieves top N industry codes for a given type
func (uc *UnifiedClassifier) getTopIndustryCodes(ctx context.Context, industryID int, codeType string, limit int) []IndustryCode {
	codes, err := uc.keywordRepo.GetClassificationCodesByIndustry(ctx, industryID)
	if err != nil {
		uc.logger.Printf("⚠️ [UnifiedClassifier] Failed to get %s codes for industry %d: %v", codeType, industryID, err)
		return []IndustryCode{}
	}

	// Filter by code type
	var filteredCodes []IndustryCode
	for _, code := range codes {
		if code.CodeType == codeType {
			// Convert repository.ClassificationCode to IndustryCode
			filteredCodes = append(filteredCodes, IndustryCode{
				Code:        code.Code,
				Description: code.Description,
				Confidence:  0.8, // Default confidence
			})
		}
	}

	// Sort by confidence (highest first)
	sort.Slice(filteredCodes, func(i, j int) bool {
		return filteredCodes[i].Confidence > filteredCodes[j].Confidence
	})

	// Take top N
	if len(filteredCodes) > limit {
		filteredCodes = filteredCodes[:limit]
	}

	// Convert to IndustryCode format
	var industryCodes []IndustryCode
	for _, code := range filteredCodes {
		industryCodes = append(industryCodes, IndustryCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
		})
	}

	return industryCodes
}

// Helper methods for keyword extraction and analysis

func (uc *UnifiedClassifier) extractKeywordsFromBusinessName(businessName string) []string {
	// Simple keyword extraction from business name
	words := strings.Fields(strings.ToLower(businessName))
	var keywords []string
	for _, word := range words {
		// Remove common words and keep meaningful terms
		if !uc.isCommonWord(word) && len(word) > 2 {
			keywords = append(keywords, word)
		}
	}
	return keywords
}

func (uc *UnifiedClassifier) extractKeywordsFromURL(websiteURL string) []string {
	// Extract domain name and path segments
	domain := strings.ToLower(websiteURL)
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "www.")

	parts := strings.Split(domain, "/")
	domainParts := strings.Split(parts[0], ".")

	var keywords []string
	for _, part := range domainParts {
		if !uc.isCommonWord(part) && len(part) > 2 {
			keywords = append(keywords, part)
		}
	}
	return keywords
}

func (uc *UnifiedClassifier) extractKeywordsFromWebsiteAnalysis(analysis *EnhancedAnalysisResult) []string {
	var keywords []string

	if analysis.RelevanceAnalysis != nil {
		// Extract keywords from IndustrySignals
		for _, signal := range analysis.RelevanceAnalysis.IndustrySignals {
			keywords = append(keywords, signal.Signal)
		}
	}

	return keywords
}

func (uc *UnifiedClassifier) extractIndustrySignalsFromWebsiteAnalysis(analysis *EnhancedAnalysisResult) []string {
	var signals []string

	if analysis.RelevanceAnalysis != nil {
		for _, signal := range analysis.RelevanceAnalysis.IndustrySignals {
			signals = append(signals, signal.Industry)
		}
	}

	return signals
}

func (uc *UnifiedClassifier) extractKeywordsFromStructuredData(structuredData *ExtractorStructuredDataResult) []string {
	var keywords []string

	// Extract from Schema.org data
	for _, item := range structuredData.SchemaOrgData {
		if item.Type == "Organization" || item.Type == "LocalBusiness" {
			if name, ok := item.Properties["name"].(string); ok {
				keywords = append(keywords, uc.extractKeywordsFromBusinessName(name)...)
			}
		}
	}

	return keywords
}

func (uc *UnifiedClassifier) isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "and": true, "or": true, "but": true, "in": true, "on": true,
		"at": true, "to": true, "for": true, "of": true, "with": true, "by": true,
		"com": true, "org": true, "net": true, "www": true, "http": true, "https": true,
	}
	return commonWords[word]
}

func (uc *UnifiedClassifier) calculateNameConfidence(businessName string) float64 {
	// Higher confidence for longer, more descriptive names
	words := strings.Fields(businessName)
	if len(words) >= 3 {
		return 0.9
	} else if len(words) == 2 {
		return 0.7
	}
	return 0.5
}

func (uc *UnifiedClassifier) calculateURLConfidence(websiteURL string) float64 {
	// Higher confidence for descriptive domain names
	domain := strings.ToLower(websiteURL)
	if strings.Contains(domain, ".") && !strings.Contains(domain, "localhost") {
		return 0.6
	}
	return 0.3
}

func (uc *UnifiedClassifier) calculateStructuredDataConfidence(structuredData *ExtractorStructuredDataResult) float64 {
	// Higher confidence when structured data is available
	if len(structuredData.SchemaOrgData) > 0 {
		return 0.9
	}
	return 0.5
}

func (uc *UnifiedClassifier) determineBusinessType(keywords []ContextualKeyword, websiteAnalysis *EnhancedAnalysisResult) string {
	// Simple business type determination based on keywords
	keywordMap := make(map[string]float64)
	for _, kw := range keywords {
		keywordMap[kw.Keyword] = kw.Weight
	}

	// Check for business type indicators
	if keywordMap["store"] > 0.5 || keywordMap["shop"] > 0.5 || keywordMap["retail"] > 0.5 {
		return "Retail Store"
	}
	if keywordMap["restaurant"] > 0.5 || keywordMap["food"] > 0.5 || keywordMap["cafe"] > 0.5 {
		return "Restaurant"
	}
	if keywordMap["service"] > 0.5 || keywordMap["consulting"] > 0.5 {
		return "Service Business"
	}

	return "General Business"
}

func (uc *UnifiedClassifier) calculateBusinessTypeConfidence(keywords []ContextualKeyword, websiteAnalysis *EnhancedAnalysisResult) float64 {
	// Calculate confidence based on keyword strength
	if len(keywords) == 0 {
		return 0.5
	}

	totalScore := 0.0
	for _, kw := range keywords {
		totalScore += kw.Weight
	}

	avgScore := totalScore / float64(len(keywords))
	if avgScore > 0.8 {
		return 0.9
	} else if avgScore > 0.6 {
		return 0.7
	}
	return 0.5
}

func (uc *UnifiedClassifier) extractTopKeywords(keywords []ContextualKeyword) []string {
	var topKeywords []string
	limit := 10
	if len(keywords) < limit {
		limit = len(keywords)
	}

	for i := 0; i < limit; i++ {
		topKeywords = append(topKeywords, keywords[i].Keyword)
	}

	return topKeywords
}

func (uc *UnifiedClassifier) calculateOverallConfidence(classificationResult *repository.ClassificationResult, weights map[string]float64, sources []DataSource) float64 {
	// Base confidence from classification
	baseConfidence := classificationResult.Confidence

	// Adjust based on data source quality
	sourceQuality := 0.0
	totalWeight := 0.0
	for _, source := range sources {
		weight := weights[source.Source] / 100.0
		sourceQuality += source.Confidence * weight
		totalWeight += weight
	}

	if totalWeight > 0 {
		avgSourceQuality := sourceQuality / totalWeight
		// Combine base confidence with source quality
		return (baseConfidence*0.7 + avgSourceQuality*0.3)
	}

	return baseConfidence
}

func (uc *UnifiedClassifier) generateClassificationReasoning(result *UnifiedClassificationResult, sources []DataSource, weights map[string]float64) string {
	reasoning := fmt.Sprintf("Primary industry identified as '%s' with %.0f%% confidence. ",
		result.PrimaryIndustry, result.IndustryConfidence*100)

	// Add data source information
	reasoning += "Classification based on "
	var sourceDescriptions []string
	for _, source := range sources {
		if weights[source.Source] > 5 { // Only mention sources with significant weight
			sourceDescriptions = append(sourceDescriptions,
				fmt.Sprintf("%s (%.0f%%)", source.Source, weights[source.Source]))
		}
	}
	reasoning += strings.Join(sourceDescriptions, ", ") + ". "

	// Add website analysis details if available
	if result.WebsiteAnalysis != nil && result.WebsiteAnalysis.Success {
		reasoning += fmt.Sprintf("Website analysis analyzed %d pages with %d relevant pages. ",
			result.WebsiteAnalysis.PagesAnalyzed, result.WebsiteAnalysis.RelevantPages)
		reasoning += fmt.Sprintf("Extracted %d keywords from website content. ",
			len(result.WebsiteAnalysis.KeywordsExtracted))
	}

	// Add keyword information
	reasoning += fmt.Sprintf("Total of %d keywords analyzed across all data sources. ", len(result.Keywords))
	reasoning += "High confidence classification based on multiple data sources and weighted analysis."

	return reasoning
}

func (uc *UnifiedClassifier) convertWebsiteAnalysis(analysis *EnhancedAnalysisResult) *WebsiteAnalysisData {
	if analysis == nil {
		return nil
	}

	return &WebsiteAnalysisData{
		Success:           analysis.Success,
		PagesAnalyzed:     len(analysis.CrawlResult.PagesAnalyzed),
		RelevantPages:     analysis.CrawlResult.RelevantPages,
		KeywordsExtracted: uc.extractKeywordsFromWebsiteAnalysis(analysis),
		IndustrySignals:   uc.extractIndustrySignalsFromWebsiteAnalysis(analysis),
		AnalysisMethod:    "smart_crawling",
		ProcessingTime:    analysis.ProcessingTime,
		OverallRelevance:  analysis.OverallConfidence,
		ContentQuality:    analysis.OverallConfidence,
		StructuredData:    map[string]interface{}{},
	}
}
