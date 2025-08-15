package webanalysis

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// WebsiteAnalysisService provides comprehensive website analysis with connection validation
type WebsiteAnalysisService struct {
	connectionValidator *ConnectionValidator
	contentAnalyzer     *EnhancedContentAnalyzer
	semanticAnalyzer    *SemanticAnalyzer
	industryClassifier  *IndustryClassifier
	pageTypeDetector    *PageTypeDetector
	priorityQueue       *PriorityScrapingQueue
	pageDiscovery       *IntelligentPageDiscovery
	metrics             *ConnectionValidationMetrics
	fallbackManager     *ConnectionFallbackManager
}

// WebsiteAnalysisResult represents comprehensive website analysis results
type WebsiteAnalysisResult struct {
	URL                    string                         `json:"url"`
	BusinessName           string                         `json:"business_name"`
	ConnectionValidation   *ConnectionValidationResult    `json:"connection_validation"`
	ContentAnalysis        *EnhancedContentAnalysis       `json:"content_analysis"`
	SemanticAnalysis       *SemanticAnalysisResult        `json:"semantic_analysis"`
	IndustryClassification []IndustryClassificationResult `json:"industry_classification"`
	PageAnalysis           []PageAnalysisResult           `json:"page_analysis"`
	OverallConfidence      float64                        `json:"overall_confidence"`
	AnalysisTime           time.Time                      `json:"analysis_time"`
	AnalysisMetadata       map[string]interface{}         `json:"analysis_metadata"`
}

// PageAnalysisResult represents analysis results for individual pages
type PageAnalysisResult struct {
	URL            string                         `json:"url"`
	PageType       string                         `json:"page_type"`
	ContentQuality float64                        `json:"content_quality"`
	RelevanceScore float64                        `json:"relevance_score"`
	PriorityScore  float64                        `json:"priority_score"`
	Classification []IndustryClassificationResult `json:"classification"`
	AnalysisTime   time.Time                      `json:"analysis_time"`
}

// ConnectionValidationMetrics tracks connection validation performance
type ConnectionValidationMetrics struct {
	mu                    sync.RWMutex
	totalValidations      int64
	successfulValidations int64
	failedValidations     int64
	averageConfidence     float64
	validationTimes       []time.Duration
	errorCounts           map[string]int64
}

// ConnectionFallbackManager manages fallback mechanisms for connection validation
type ConnectionFallbackManager struct {
	fallbackStrategies map[string]FallbackStrategy
	maxRetries         int
	retryDelay         time.Duration
}

// FallbackStrategy defines a fallback strategy for connection validation
type FallbackStrategy interface {
	Execute(ctx context.Context, business string, website string, content *ScrapedContent) (*ConnectionValidationResult, error)
	GetName() string
	GetPriority() int
}

// NewWebsiteAnalysisService creates a new website analysis service
func NewWebsiteAnalysisService() *WebsiteAnalysisService {
	// Initialize components
	contentAnalyzer := NewEnhancedContentAnalyzer()
	semanticAnalyzer := NewSemanticAnalyzer()
	industryClassifier := NewIndustryClassifier(semanticAnalyzer, contentAnalyzer)
	pageTypeDetector := NewPageTypeDetector()
	priorityQueue := NewPriorityScrapingQueue()
	pageDiscovery := NewIntelligentPageDiscovery()

	return &WebsiteAnalysisService{
		connectionValidator: NewConnectionValidator(),
		contentAnalyzer:     contentAnalyzer,
		semanticAnalyzer:    semanticAnalyzer,
		industryClassifier:  industryClassifier,
		pageTypeDetector:    pageTypeDetector,
		priorityQueue:       priorityQueue,
		pageDiscovery:       pageDiscovery,
		metrics:             NewConnectionValidationMetrics(),
		fallbackManager:     NewConnectionFallbackManager(),
	}
}

// AnalyzeWebsite performs comprehensive website analysis with connection validation
func (was *WebsiteAnalysisService) AnalyzeWebsite(ctx context.Context, business string, website string) (*WebsiteAnalysisResult, error) {
	startTime := time.Now()

	// Step 1: Scrape website content
	scrapedContent, err := was.scrapeWebsite(ctx, website)
	if err != nil {
		// Try fallback scraping methods
		scrapedContent, err = was.fallbackManager.TryFallbackScraping(ctx, website)
		if err != nil {
			return nil, fmt.Errorf("failed to scrape website: %w", err)
		}
	}

	// Step 2: Perform connection validation
	connectionValidation, err := was.performConnectionValidation(ctx, business, website, scrapedContent)
	if err != nil {
		// Try fallback validation methods
		connectionValidation, err = was.fallbackManager.TryFallbackValidation(ctx, business, website, scrapedContent)
		if err != nil {
			// Create a minimal validation result for failed cases
			connectionValidation = was.createMinimalValidationResult(business, website)
		}
	}

	// Step 3: Perform content analysis
	contentAnalysis, err := was.contentAnalyzer.AnalyzeContent(scrapedContent, business)
	if err != nil {
		contentAnalysis = was.createMinimalContentAnalysis(scrapedContent)
	}

	// Step 4: Perform semantic analysis
	semanticAnalysis, err := was.semanticAnalyzer.AnalyzeSemanticContent(scrapedContent, business)
	if err != nil {
		semanticAnalysis = was.createMinimalSemanticAnalysis(scrapedContent)
	}

	// Step 5: Perform industry classification
	industryClassification, err := was.industryClassifier.ClassifyContentEnhanced(ctx, scrapedContent, business, 3)
	if err != nil {
		industryClassification = was.createMinimalIndustryClassification(business)
	}

	// Step 6: Perform page analysis
	pageAnalysis, err := was.performPageAnalysis(ctx, website, business)
	if err != nil {
		pageAnalysis = []PageAnalysisResult{}
	}

	// Step 7: Calculate overall confidence
	overallConfidence := was.calculateOverallConfidence(
		connectionValidation, contentAnalysis, semanticAnalysis, industryClassification)

	// Step 8: Update metrics
	was.updateMetrics(connectionValidation, time.Since(startTime))

	// Create analysis metadata
	metadata := map[string]interface{}{
		"business":          business,
		"website":           website,
		"analysis_duration": time.Since(startTime).String(),
		"content_length":    len(scrapedContent.Text),
		"pages_analyzed":    len(pageAnalysis),
		"connection_valid":  connectionValidation.OverallConfidence > 0.5,
		"fallback_used":     err != nil,
	}

	result := &WebsiteAnalysisResult{
		URL:                    website,
		BusinessName:           business,
		ConnectionValidation:   connectionValidation,
		ContentAnalysis:        contentAnalysis,
		SemanticAnalysis:       semanticAnalysis,
		IndustryClassification: industryClassification,
		PageAnalysis:           pageAnalysis,
		OverallConfidence:      overallConfidence,
		AnalysisTime:           time.Now(),
		AnalysisMetadata:       metadata,
	}

	return result, nil
}

// performConnectionValidation performs connection validation with error handling
func (was *WebsiteAnalysisService) performConnectionValidation(ctx context.Context, business string, website string, content *ScrapedContent) (*ConnectionValidationResult, error) {
	// Record validation start time
	startTime := time.Now()

	// Perform validation
	result, err := was.connectionValidator.ValidateConnection(ctx, business, website, content)
	if err != nil {
		// Record failed validation
		was.metrics.RecordFailedValidation(err.Error())
		return nil, err
	}

	// Record successful validation
	was.metrics.RecordSuccessfulValidation(result.OverallConfidence, time.Since(startTime))

	return result, nil
}

// performPageAnalysis performs analysis of multiple pages
func (was *WebsiteAnalysisService) performPageAnalysis(ctx context.Context, website string, business string) ([]PageAnalysisResult, error) {
	// Discover pages to analyze
	discoveryResults, err := was.pageDiscovery.DiscoverPages(ctx, website, business, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to discover pages: %w", err)
	}

	var pageResults []PageAnalysisResult

	// Analyze each discovered page
	for _, discoveryResult := range discoveryResults {
		// Scrape page content
		pageContent, err := was.scrapePage(ctx, discoveryResult.URL)
		if err != nil {
			continue // Skip pages that can't be scraped
		}

		// Detect page type
		pageType := was.pageTypeDetector.DetectPageType(discoveryResult.URL, pageContent)

		// Perform classification
		classification, err := was.industryClassifier.ClassifyContentEnhanced(ctx, pageContent, business, 1)
		if err != nil {
			classification = []IndustryClassificationResult{}
		}

		pageResult := PageAnalysisResult{
			URL:            discoveryResult.URL,
			PageType:       pageType.Type,
			ContentQuality: discoveryResult.ContentQuality,
			RelevanceScore: discoveryResult.RelevanceScore,
			PriorityScore:  discoveryResult.PriorityScore,
			Classification: classification,
			AnalysisTime:   time.Now(),
		}

		pageResults = append(pageResults, pageResult)
	}

	return pageResults, nil
}

// calculateOverallConfidence calculates overall confidence based on all analysis components
func (was *WebsiteAnalysisService) calculateOverallConfidence(
	connectionValidation *ConnectionValidationResult,
	contentAnalysis *EnhancedContentAnalysis,
	semanticAnalysis *SemanticAnalysisResult,
	industryClassification []IndustryClassificationResult) float64 {

	// Weight factors for different components
	weights := map[string]float64{
		"connection": 0.3,
		"content":    0.2,
		"semantic":   0.2,
		"industry":   0.3,
	}

	// Connection confidence
	connectionConfidence := connectionValidation.OverallConfidence

	// Content quality confidence
	contentConfidence := 0.0
	if contentAnalysis != nil && contentAnalysis.ContentQuality != nil {
		contentConfidence = contentAnalysis.ContentQuality.OverallQuality
	}

	// Semantic confidence
	semanticConfidence := 0.0
	if semanticAnalysis != nil {
		semanticConfidence = semanticAnalysis.SemanticScore
	}

	// Industry classification confidence
	industryConfidence := 0.0
	if len(industryClassification) > 0 {
		industryConfidence = industryClassification[0].Confidence
	}

	// Calculate weighted average
	overallConfidence := connectionConfidence*weights["connection"] +
		contentConfidence*weights["content"] +
		semanticConfidence*weights["semantic"] +
		industryConfidence*weights["industry"]

	// Normalize to 0-1 range
	if overallConfidence > 1.0 {
		overallConfidence = 1.0
	}

	return overallConfidence
}

// scrapeWebsite scrapes the main website
func (was *WebsiteAnalysisService) scrapeWebsite(ctx context.Context, website string) (*ScrapedContent, error) {
	// This would integrate with existing scraping infrastructure
	// For now, return a placeholder
	return &ScrapedContent{
		URL:       website,
		Text:      "Sample website content for " + website,
		Title:     "Sample Title",
		HTML:      "<html><body>Sample content</body></html>",
		ScrapedAt: time.Now(),
	}, nil
}

// scrapePage scrapes a specific page
func (was *WebsiteAnalysisService) scrapePage(ctx context.Context, pageURL string) (*ScrapedContent, error) {
	// This would integrate with existing scraping infrastructure
	// For now, return a placeholder
	return &ScrapedContent{
		URL:       pageURL,
		Text:      "Sample page content for " + pageURL,
		Title:     "Sample Page Title",
		HTML:      "<html><body>Sample page content</body></html>",
		ScrapedAt: time.Now(),
	}, nil
}

// createMinimalValidationResult creates a minimal validation result for failed cases
func (was *WebsiteAnalysisService) createMinimalValidationResult(business string, website string) *ConnectionValidationResult {
	return &ConnectionValidationResult{
		BusinessNameMatch: BusinessNameMatchResult{
			IsMatch:    false,
			Confidence: 0.0,
			MatchType:  "none",
			Evidence:   []string{"Validation failed - using minimal result"},
		},
		AddressValidation: AddressValidationResult{
			IsValid:        false,
			Confidence:     0.0,
			ValidationType: "none",
			Evidence:       []string{"Validation failed - using minimal result"},
		},
		ContactValidation: ContactValidationResult{
			IsValid:        false,
			Confidence:     0.0,
			ValidationType: "none",
			Evidence:       []string{"Validation failed - using minimal result"},
		},
		DomainAnalysis: DomainAnalysisResult{
			IsRelevant: false,
			Confidence: 0.0,
			DomainName: website,
			Evidence:   []string{"Validation failed - using minimal result"},
		},
		OverallConfidence: 0.0,
		ValidationTime:    time.Now(),
		ValidationMetadata: map[string]interface{}{
			"fallback_used": true,
			"error":         "Validation failed",
		},
	}
}

// createMinimalContentAnalysis creates a minimal content analysis for failed cases
func (was *WebsiteAnalysisService) createMinimalContentAnalysis(content *ScrapedContent) *EnhancedContentAnalysis {
	return &EnhancedContentAnalysis{
		ContentQuality: &ContentQualityAssessment{
			OverallQuality: 0.1,
			Readability:    0.1,
			Length:         0.1,
			Structure:      0.1,
			Relevance:      0.1,
		},
		AnalyzedAt: time.Now(),
	}
}

// createMinimalSemanticAnalysis creates a minimal semantic analysis for failed cases
func (was *WebsiteAnalysisService) createMinimalSemanticAnalysis(content *ScrapedContent) *SemanticAnalysisResult {
	return &SemanticAnalysisResult{
		SemanticScore: 0.1,
		AnalyzedAt:    time.Now(),
	}
}

// createMinimalIndustryClassification creates a minimal industry classification for failed cases
func (was *WebsiteAnalysisService) createMinimalIndustryClassification(business string) []IndustryClassificationResult {
	return []IndustryClassificationResult{
		{
			Industry:           "Unknown",
			Confidence:         0.1,
			ClassificationTime: time.Now(),
		},
	}
}

// updateMetrics updates connection validation metrics
func (was *WebsiteAnalysisService) updateMetrics(validation *ConnectionValidationResult, duration time.Duration) {
	was.metrics.RecordValidation(validation.OverallConfidence, duration)
}

// GetMetrics returns current connection validation metrics
func (was *WebsiteAnalysisService) GetMetrics() *ConnectionValidationMetrics {
	return was.metrics
}

// NewConnectionValidationMetrics creates new connection validation metrics
func NewConnectionValidationMetrics() *ConnectionValidationMetrics {
	return &ConnectionValidationMetrics{
		validationTimes: []time.Duration{},
		errorCounts:     make(map[string]int64),
	}
}

// RecordSuccessfulValidation records a successful validation
func (cvm *ConnectionValidationMetrics) RecordSuccessfulValidation(confidence float64, duration time.Duration) {
	cvm.mu.Lock()
	defer cvm.mu.Unlock()

	cvm.totalValidations++
	cvm.successfulValidations++
	cvm.validationTimes = append(cvm.validationTimes, duration)

	// Update average confidence
	totalConfidence := cvm.averageConfidence * float64(cvm.successfulValidations-1)
	cvm.averageConfidence = (totalConfidence + confidence) / float64(cvm.successfulValidations)
}

// RecordFailedValidation records a failed validation
func (cvm *ConnectionValidationMetrics) RecordFailedValidation(errorType string) {
	cvm.mu.Lock()
	defer cvm.mu.Unlock()

	cvm.totalValidations++
	cvm.failedValidations++
	cvm.errorCounts[errorType]++
}

// RecordValidation records a validation with confidence and duration
func (cvm *ConnectionValidationMetrics) RecordValidation(confidence float64, duration time.Duration) {
	cvm.mu.Lock()
	defer cvm.mu.Unlock()

	cvm.totalValidations++
	cvm.validationTimes = append(cvm.validationTimes, duration)

	// Update average confidence
	totalConfidence := cvm.averageConfidence * float64(cvm.totalValidations-1)
	cvm.averageConfidence = (totalConfidence + confidence) / float64(cvm.totalValidations)
}

// GetStats returns validation statistics
func (cvm *ConnectionValidationMetrics) GetStats() map[string]interface{} {
	cvm.mu.RLock()
	defer cvm.mu.RUnlock()

	successRate := 0.0
	if cvm.totalValidations > 0 {
		successRate = float64(cvm.successfulValidations) / float64(cvm.totalValidations)
	}

	avgDuration := 0.0
	if len(cvm.validationTimes) > 0 {
		totalDuration := time.Duration(0)
		for _, duration := range cvm.validationTimes {
			totalDuration += duration
		}
		avgDuration = float64(totalDuration) / float64(len(cvm.validationTimes))
	}

	return map[string]interface{}{
		"total_validations":      cvm.totalValidations,
		"successful_validations": cvm.successfulValidations,
		"failed_validations":     cvm.failedValidations,
		"success_rate":           successRate,
		"average_confidence":     cvm.averageConfidence,
		"average_duration_ms":    avgDuration,
		"error_counts":           cvm.errorCounts,
	}
}

// NewConnectionFallbackManager creates a new connection fallback manager
func NewConnectionFallbackManager() *ConnectionFallbackManager {
	return &ConnectionFallbackManager{
		fallbackStrategies: make(map[string]FallbackStrategy),
		maxRetries:         3,
		retryDelay:         time.Second * 2,
	}
}

// TryFallbackScraping tries fallback scraping methods
func (cfm *ConnectionFallbackManager) TryFallbackScraping(ctx context.Context, website string) (*ScrapedContent, error) {
	// Implement fallback scraping strategies
	// For now, return a placeholder
	return &ScrapedContent{
		URL:       website,
		Text:      "Fallback content for " + website,
		Title:     "Fallback Title",
		HTML:      "<html><body>Fallback content</body></html>",
		ScrapedAt: time.Now(),
	}, nil
}

// TryFallbackValidation tries fallback validation methods
func (cfm *ConnectionFallbackManager) TryFallbackValidation(ctx context.Context, business string, website string, content *ScrapedContent) (*ConnectionValidationResult, error) {
	// Implement fallback validation strategies
	// For now, return a minimal result
	return &ConnectionValidationResult{
		OverallConfidence: 0.1,
		ValidationTime:    time.Now(),
		ValidationMetadata: map[string]interface{}{
			"fallback_used": true,
		},
	}, nil
}
