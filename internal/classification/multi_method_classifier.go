package classification

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"kyb-platform/internal/classification/repository"
	"kyb-platform/internal/machine_learning"
	"kyb-platform/internal/machine_learning/infrastructure"
	"kyb-platform/internal/shared"
)

// CachedWebsiteContent represents cached website content
type CachedWebsiteContent struct {
	Keywords  []string
	Timestamp time.Time
}

// WebsiteCache provides thread-safe caching for website content
type WebsiteCache struct {
	cache map[string]*CachedWebsiteContent
	mu    sync.RWMutex
	ttl   time.Duration
}

// NewWebsiteCache creates a new website cache with the specified TTL
func NewWebsiteCache(ttl time.Duration) *WebsiteCache {
	return &WebsiteCache{
		cache: make(map[string]*CachedWebsiteContent),
		ttl:   ttl,
	}
}

// Get retrieves cached content for a URL if it exists and is not expired
func (wc *WebsiteCache) Get(url string) []string {
	wc.mu.RLock()
	defer wc.mu.RUnlock()

	cached, exists := wc.cache[url]
	if !exists {
		return nil
	}

	// Check if cache entry is still valid
	if time.Since(cached.Timestamp) >= wc.ttl {
		return nil
	}

	return cached.Keywords
}

// Set stores content in the cache
func (wc *WebsiteCache) Set(url string, keywords []string) {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	wc.cache[url] = &CachedWebsiteContent{
		Keywords:  keywords,
		Timestamp: time.Now(),
	}
}

// MultiMethodClassifier provides enhanced classification using multiple methods
type MultiMethodClassifier struct {
	keywordRepo              repository.KeywordRepository
	mlClassifier             *machine_learning.ContentClassifier
	pythonMLService          interface{} // *infrastructure.PythonMLService - using interface to avoid import cycle
	weightedConfidenceScorer *WeightedConfidenceScorer
	reasoningEngine          *ReasoningEngine
	qualityMetricsService    *QualityMetricsService
	enhancedScraper          *EnhancedWebsiteScraper
	logger                   *log.Logger
	monitor                  *ClassificationAccuracyMonitoring
	websiteCache             *WebsiteCache // Cache for website scraping results
}

// NewMultiMethodClassifier creates a new multi-method classifier
func NewMultiMethodClassifier(
	keywordRepo repository.KeywordRepository,
	mlClassifier *machine_learning.ContentClassifier,
	logger *log.Logger,
) *MultiMethodClassifier {
	if logger == nil {
		logger = log.Default()
	}

	return &MultiMethodClassifier{
		keywordRepo:              keywordRepo,
		mlClassifier:             mlClassifier,
		pythonMLService:          nil, // Can be set via SetPythonMLService
		weightedConfidenceScorer: NewWeightedConfidenceScorer(logger),
		reasoningEngine:          NewReasoningEngine(logger),
		qualityMetricsService:    NewQualityMetricsService(logger),
		enhancedScraper:          NewEnhancedWebsiteScraper(logger),
		logger:                   logger,
		monitor:                  nil, // Will be set separately if monitoring is needed
		websiteCache:             NewWebsiteCache(24 * time.Hour), // Cache for 24 hours
	}
}

// NewMultiMethodClassifierWithPythonML creates a new multi-method classifier with Python ML service support
func NewMultiMethodClassifierWithPythonML(
	keywordRepo repository.KeywordRepository,
	mlClassifier *machine_learning.ContentClassifier,
	pythonMLService interface{}, // *infrastructure.PythonMLService - using interface to avoid import cycle
	logger *log.Logger,
) *MultiMethodClassifier {
	classifier := NewMultiMethodClassifier(keywordRepo, mlClassifier, logger)
	classifier.pythonMLService = pythonMLService
	return classifier
}

// SetPythonMLService sets the Python ML service for enhanced classification
func (mmc *MultiMethodClassifier) SetPythonMLService(pythonMLService interface{}) {
	mmc.pythonMLService = pythonMLService
}

// NewMultiMethodClassifierWithMonitoring creates a new multi-method classifier with monitoring
func NewMultiMethodClassifierWithMonitoring(
	keywordRepo repository.KeywordRepository,
	mlClassifier *machine_learning.ContentClassifier,
	logger *log.Logger,
	monitor *ClassificationAccuracyMonitoring,
) *MultiMethodClassifier {
	if logger == nil {
		logger = log.Default()
	}

	return &MultiMethodClassifier{
		keywordRepo:              keywordRepo,
		mlClassifier:             mlClassifier,
		weightedConfidenceScorer: NewWeightedConfidenceScorer(logger),
		reasoningEngine:          NewReasoningEngine(logger),
		qualityMetricsService:    NewQualityMetricsService(logger),
		logger:                   logger,
		monitor:                  monitor,
	}
}

// MultiMethodClassificationResult represents the result of multi-method classification
type MultiMethodClassificationResult struct {
	BusinessName            string                              `json:"business_name"`
	PrimaryClassification   *shared.IndustryClassification      `json:"primary_classification"`
	MethodResults           []shared.ClassificationMethodResult `json:"method_results"`
	EnsembleConfidence      float64                             `json:"ensemble_confidence"`
	ClassificationReasoning string                              `json:"classification_reasoning"`
	QualityMetrics          *shared.ClassificationQuality       `json:"quality_metrics"`
	ProcessingTime          time.Duration                       `json:"processing_time"`
	CreatedAt               time.Time                           `json:"created_at"`
}

// ClassifyWithMultipleMethods performs classification using multiple methods in parallel
func (mmc *MultiMethodClassifier) ClassifyWithMultipleMethods(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*MultiMethodClassificationResult, error) {
	startTime := time.Now()
	requestID := mmc.generateRequestID()

	mmc.logger.Printf("🚀 Starting multi-method classification for: %s (request: %s)", businessName, requestID)

	// Create channels for results
	resultChan := make(chan shared.ClassificationMethodResult, 3)
	errorChan := make(chan error, 3)

	// Create a WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Method 1: Keyword-based classification
	wg.Add(1)
	go func() {
		defer wg.Done()
		mmc.logger.Printf("🔄 Method 1: Keyword-based classification (request: %s)", requestID)

		methodStart := time.Now()
		result, err := mmc.performKeywordClassification(ctx, businessName, description, websiteURL)
		methodTime := time.Since(methodStart)

		method := shared.ClassificationMethodResult{
			MethodName:     "keyword_classification",
			MethodType:     "keyword",
			ProcessingTime: methodTime,
			Success:        err == nil,
		}

		if err != nil {
			method.Error = err.Error()
			mmc.logger.Printf("⚠️ Keyword classification failed: %v (request: %s)", err, requestID)
		} else {
			method.Result = result
			method.Confidence = result.ConfidenceScore
			method.Evidence = result.Keywords
			method.Keywords = result.Keywords
			mmc.logger.Printf("✅ Keyword classification completed: %s (confidence: %.2f%%) (request: %s)",
				result.IndustryName, result.ConfidenceScore*100, requestID)
		}

		resultChan <- method
	}()

	// Method 2: ML-based classification
	wg.Add(1)
	go func() {
		defer wg.Done()
		mmc.logger.Printf("🔄 Method 2: ML-based classification (request: %s)", requestID)

		methodStart := time.Now()
		result, err := mmc.performMLClassification(ctx, businessName, description, websiteURL)
		methodTime := time.Since(methodStart)

		method := shared.ClassificationMethodResult{
			MethodName:     "ml_classification",
			MethodType:     "ml",
			ProcessingTime: methodTime,
			Success:        err == nil,
		}

		if err != nil {
			method.Error = err.Error()
			mmc.logger.Printf("⚠️ ML classification failed: %v (request: %s)", err, requestID)
		} else {
			method.Result = result
			method.Confidence = result.ConfidenceScore
			method.Evidence = []string{fmt.Sprintf("ML model prediction with confidence %.2f%%", result.ConfidenceScore*100)}
			mmc.logger.Printf("✅ ML classification completed: %s (confidence: %.2f%%) (request: %s)",
				result.IndustryName, result.ConfidenceScore*100, requestID)
		}

		resultChan <- method
	}()

	// Method 3: Description-based classification
	wg.Add(1)
	go func() {
		defer wg.Done()
		mmc.logger.Printf("🔄 Method 3: Description-based classification (request: %s)", requestID)

		methodStart := time.Now()
		result, err := mmc.performDescriptionClassification(ctx, businessName, description, websiteURL)
		methodTime := time.Since(methodStart)

		method := shared.ClassificationMethodResult{
			MethodName:     "description_classification",
			MethodType:     "description",
			ProcessingTime: methodTime,
			Success:        err == nil,
		}

		if err != nil {
			method.Error = err.Error()
			mmc.logger.Printf("⚠️ Description classification failed: %v (request: %s)", err, requestID)
		} else {
			method.Result = result
			method.Confidence = result.ConfidenceScore
			method.Evidence = []string{fmt.Sprintf("Description analysis with confidence %.2f%%", result.ConfidenceScore*100)}
			mmc.logger.Printf("✅ Description classification completed: %s (confidence: %.2f%%) (request: %s)",
				result.IndustryName, result.ConfidenceScore*100, requestID)
		}

		resultChan <- method
	}()

	// Wait for all goroutines to complete
	wg.Wait()
	close(resultChan)
	close(errorChan)

	// Collect results
	var methodResults []shared.ClassificationMethodResult
	for method := range resultChan {
		methodResults = append(methodResults, method)
	}

	// Log any errors
	for err := range errorChan {
		mmc.logger.Printf("⚠️ Error in multi-method classification: %v (request: %s)", err, requestID)
	}

	// Calculate ensemble result
	ensembleResult := mmc.calculateEnsembleResult(methodResults, businessName)

	// Calculate weighted confidence using the sophisticated scorer
	weightedConfidenceResult, err := mmc.weightedConfidenceScorer.CalculateWeightedConfidence(ctx, methodResults)
	if err != nil {
		mmc.logger.Printf("⚠️ Failed to calculate weighted confidence: %v (request: %s)", err, requestID)
		// Fallback to simple ensemble confidence
		ensembleResult.ConfidenceScore = mmc.calculateEnsembleConfidence(methodResults, ensembleResult.IndustryName)
	} else {
		// Use the sophisticated weighted confidence
		ensembleResult.ConfidenceScore = weightedConfidenceResult.FinalConfidence
		// Add weighted confidence details to metadata
		if ensembleResult.Metadata == nil {
			ensembleResult.Metadata = make(map[string]interface{})
		}
		ensembleResult.Metadata["weighted_confidence_details"] = weightedConfidenceResult
	}

	// Calculate comprehensive quality metrics using the sophisticated service
	var qualityMetrics *shared.ClassificationQuality
	comprehensiveQualityMetrics, err := mmc.qualityMetricsService.CalculateComprehensiveQualityMetrics(
		ctx, methodResults, ensembleResult, &shared.BusinessClassificationRequest{
			BusinessName: businessName,
			Description:  description,
			WebsiteURL:   websiteURL,
		})
	if err != nil {
		mmc.logger.Printf("⚠️ Failed to calculate comprehensive quality metrics: %v (request: %s)", err, requestID)
		// Fallback to simple quality metrics
		qualityMetrics = mmc.calculateQualityMetrics(methodResults)
	} else {
		// Use comprehensive quality metrics
		qualityMetrics = &shared.ClassificationQuality{
			OverallQuality:     comprehensiveQualityMetrics.OverallQuality,
			MethodAgreement:    comprehensiveQualityMetrics.MethodAgreement,
			ConfidenceVariance: comprehensiveQualityMetrics.ConfidenceMetrics.ConfidenceVariance,
			EvidenceStrength:   comprehensiveQualityMetrics.EvidenceStrength,
			DataCompleteness:   comprehensiveQualityMetrics.DataCompleteness,
		}

		// Add comprehensive quality metrics to metadata
		if ensembleResult.Metadata == nil {
			ensembleResult.Metadata = make(map[string]interface{})
		}
		ensembleResult.Metadata["comprehensive_quality_metrics"] = comprehensiveQualityMetrics
	}

	// Generate sophisticated classification reasoning
	var reasoning string
	reasoningResult, err := mmc.reasoningEngine.GenerateReasoning(ctx, businessName, methodResults, ensembleResult, qualityMetrics)
	if err != nil {
		mmc.logger.Printf("⚠️ Failed to generate reasoning: %v (request: %s)", err, requestID)
		// Fallback to simple reasoning
		reasoning = mmc.generateClassificationReasoning(methodResults, ensembleResult)
	} else {
		// Use sophisticated reasoning
		reasoning = reasoningResult.Summary
		// Add detailed reasoning to metadata
		if ensembleResult.Metadata == nil {
			ensembleResult.Metadata = make(map[string]interface{})
		}
		ensembleResult.Metadata["detailed_reasoning"] = reasoningResult
	}

	// Create final result
	finalResult := &MultiMethodClassificationResult{
		BusinessName:            businessName,
		PrimaryClassification:   ensembleResult,
		MethodResults:           methodResults,
		EnsembleConfidence:      ensembleResult.ConfidenceScore,
		ClassificationReasoning: reasoning,
		QualityMetrics:          qualityMetrics,
		ProcessingTime:          time.Since(startTime),
		CreatedAt:               time.Now(),
	}

	// Record performance metrics
	mmc.recordMultiMethodMetrics(ctx, requestID, businessName, methodResults, finalResult, time.Since(startTime), nil)

	mmc.logger.Printf("✅ Multi-method classification completed: %s (ensemble confidence: %.2f%%) (request: %s)",
		ensembleResult.IndustryName, ensembleResult.ConfidenceScore*100, requestID)

	return finalResult, nil
}

// performKeywordClassification performs keyword-based classification
func (mmc *MultiMethodClassifier) performKeywordClassification(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*shared.IndustryClassification, error) {
	// Extract keywords from business information
	keywords := mmc.extractKeywords(businessName, websiteURL)

	// Classify using keyword repository
	classificationResult, err := mmc.keywordRepo.ClassifyBusinessByKeywords(ctx, keywords)
	if err != nil {
		return nil, fmt.Errorf("keyword classification failed: %w", err)
	}

	// Convert classification codes to the expected format
	// Convert []ClassificationCode to []*ClassificationCode
	codePointers := make([]*repository.ClassificationCode, len(classificationResult.Codes))
	for i := range classificationResult.Codes {
		codePointers[i] = &classificationResult.Codes[i]
	}
	classificationCodes := mmc.convertClassificationCodes(codePointers)

	// Convert to shared format
	result := &shared.IndustryClassification{
		IndustryCode:         classificationResult.Industry.Name,
		IndustryName:         classificationResult.Industry.Name,
		ConfidenceScore:      classificationResult.Confidence,
		ClassificationMethod: "keyword",
		Keywords:             keywords,
		Description:          fmt.Sprintf("Keyword-based classification using %d keywords", len(keywords)),
		Evidence:             fmt.Sprintf("Matched %d keywords: %v", len(keywords), keywords),
		ProcessingTime:       time.Duration(0), // Will be set by caller
		Metadata: map[string]interface{}{
			"keywords_matched":     len(keywords),
			"classification_codes": classificationCodes,
		},
	}

	return result, nil
}

// performMLClassification performs ML-based classification
// Uses Python ML service (DistilBART) if available, otherwise falls back to Go ML classifier
func (mmc *MultiMethodClassifier) performMLClassification(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*shared.IndustryClassification, error) {
	// Try Python ML service (DistilBART) first if available
	if mmc.pythonMLService != nil {
		mmc.logger.Printf("🤖 Using Python ML service (DistilBART) for classification")
		return mmc.performPythonMLClassification(ctx, businessName, description, websiteURL)
	}

	// Fallback to Go ML classifier
	if mmc.mlClassifier == nil {
		return nil, fmt.Errorf("ML classifier not available")
	}

	mmc.logger.Printf("📝 Using Go ML classifier (fallback)")
	
	// Combine business information for ML analysis (using trusted content only)
	content := mmc.extractTrustedContent(ctx, businessName, description, websiteURL)

	// Perform ML classification
	mlResult, err := mmc.mlClassifier.ClassifyContent(ctx, content, "")
	if err != nil {
		return nil, fmt.Errorf("ML classification failed: %w", err)
	}

	// Find the best classification from ML result
	if len(mlResult.Classifications) == 0 {
		return nil, fmt.Errorf("no classifications returned from ML model")
	}

	// Get the highest confidence classification
	bestClassification := mlResult.Classifications[0]
	for _, classification := range mlResult.Classifications {
		if classification.Confidence > bestClassification.Confidence {
			bestClassification = classification
		}
	}

	// Get classification codes for the ML-detected industry
	var classificationCodes shared.ClassificationCodes
	if bestClassification.Label != "unknown" {
		// Try to get classification codes based on the ML-detected industry
		// For now, we'll use a simple mapping - in production this should be more sophisticated
		classificationCodes = mmc.getClassificationCodesForIndustry(ctx, bestClassification.Label)
	}

	// Convert to shared format
	result := &shared.IndustryClassification{
		IndustryCode:         bestClassification.Label,
		IndustryName:         bestClassification.Label,
		ConfidenceScore:      bestClassification.Confidence,
		ClassificationMethod: "ml",
		Description:          fmt.Sprintf("ML-based classification using %s model", mlResult.ModelID),
		Evidence:             fmt.Sprintf("ML model prediction with confidence %.2f%%", bestClassification.Confidence*100),
		ProcessingTime:       mlResult.ProcessingTime,
		Metadata: map[string]interface{}{
			"model_id":             mlResult.ModelID,
			"model_version":        mlResult.ModelVersion,
			"all_predictions":      mlResult.Classifications,
			"quality_score":        mlResult.QualityScore,
			"classification_codes": classificationCodes,
		},
	}

	return result, nil
}

// performPythonMLClassification performs ML classification using Python ML service (DistilBART)
// Phase 1 Enhancement: Extracts keywords first and enhances ML input with keyword context
func (mmc *MultiMethodClassifier) performPythonMLClassification(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*shared.IndustryClassification, error) {
	// PHASE 1: Extract keywords from database FIRST (before ML classification)
	keywords := mmc.extractKeywords(businessName, websiteURL)
	mmc.logger.Printf("🔑 [Phase 1] Extracted %d keywords before ML classification", len(keywords))

	// Extract website content if URL is provided
	var websiteContent string
	if websiteURL != "" {
		scrapingResult := mmc.enhancedScraper.ScrapeWebsite(ctx, websiteURL)
		if scrapingResult.Success {
			websiteContent = scrapingResult.TextContent
		}
	}

	// Prepare content - use description if website content is empty
	contentToSend := websiteContent
	if contentToSend == "" && description != "" {
		contentToSend = description
	}
	if contentToSend == "" && businessName != "" {
		contentToSend = businessName
	}

	// PHASE 1: Enhance ML input with keywords
	enhancedContent := mmc.enhanceMLInputWithKeywords(contentToSend, keywords)
	mmc.logger.Printf("📝 [Phase 1] Enhanced ML input with %d keywords", len(keywords))

	// Minimum content length check
	const minContentLength = 20
	if len(strings.TrimSpace(enhancedContent)) < minContentLength {
		mmc.logger.Printf("⚠️ Content too short (%d chars < %d), falling back to Go ML classifier", len(enhancedContent), minContentLength)
		// Fallback to Go ML classifier
		return mmc.performGoMLClassification(ctx, businessName, description, websiteURL)
	}

	// Create enhanced classification request with keyword-enhanced content
	req := &infrastructure.EnhancedClassificationRequest{
		BusinessName:     businessName,
		Description:      description,
		WebsiteURL:       websiteURL,
		WebsiteContent:   enhancedContent, // Use enhanced content with keywords
		MaxResults:       5,
		MaxContentLength: 1024,
	}

	// Type assert to get the actual PythonMLService
	pms, ok := mmc.pythonMLService.(interface {
		ClassifyEnhanced(ctx context.Context, req *infrastructure.EnhancedClassificationRequest) (*infrastructure.EnhancedClassificationResponse, error)
	})
	if !ok {
		mmc.logger.Printf("⚠️ Python ML service type assertion failed, falling back to Go ML classifier")
		return mmc.performGoMLClassification(ctx, businessName, description, websiteURL)
	}

	// Call Python ML service
	enhancedResp, err := pms.ClassifyEnhanced(ctx, req)
	if err != nil {
		mmc.logger.Printf("⚠️ Python ML service classification failed: %v, falling back to Go ML classifier", err)
		return mmc.performGoMLClassification(ctx, businessName, description, websiteURL)
	}

	// Build result from enhanced response
	if len(enhancedResp.Classifications) == 0 {
		mmc.logger.Printf("⚠️ No classifications from Python ML service, falling back to Go ML classifier")
		return mmc.performGoMLClassification(ctx, businessName, description, websiteURL)
	}

	// Get primary industry and confidence
	primaryIndustry := enhancedResp.Classifications[0].Label
	confidence := enhancedResp.Classifications[0].Confidence

	// Build all industry scores map
	allScores := make(map[string]float64)
	for _, classification := range enhancedResp.Classifications {
		allScores[classification.Label] = classification.Confidence
	}

	// Get classification codes for the detected industry
	classificationCodes := mmc.getClassificationCodesForIndustry(ctx, primaryIndustry)

	// PHASE 1: Validate ML results against keywords
	validatedResult := mmc.validateMLAgainstKeywords(ctx, businessName, description, websiteURL, primaryIndustry, confidence, keywords, classificationCodes, enhancedResp)

	mmc.logger.Printf("✅ Python ML service classification: %s (confidence: %.2f%%)", validatedResult.IndustryName, validatedResult.ConfidenceScore*100)
	return validatedResult, nil
}

// performGoMLClassification performs ML classification using Go ML classifier (fallback)
// Enhanced with keyword fallback for better accuracy when ML fails or has low confidence
func (mmc *MultiMethodClassifier) performGoMLClassification(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*shared.IndustryClassification, error) {
	if mmc.mlClassifier == nil {
		// Fallback to keyword-based classification
		mmc.logger.Printf("⚠️ ML classifier not available, using keyword fallback")
		return mmc.performKeywordClassification(ctx, businessName, description, websiteURL)
	}

	// Combine business information for ML analysis
	content := mmc.extractTrustedContent(ctx, businessName, description, websiteURL)

	// Validate content quality
	const minContentLength = 10
	if len(content) < minContentLength {
		mmc.logger.Printf("⚠️ Insufficient content for ML classification (length: %d < %d), using keyword fallback", len(content), minContentLength)
		return mmc.performKeywordClassification(ctx, businessName, description, websiteURL)
	}

	// Perform ML classification
	mlResult, err := mmc.mlClassifier.ClassifyContent(ctx, content, "")
	if err != nil {
		mmc.logger.Printf("⚠️ ML classification failed: %v, using keyword fallback", err)
		return mmc.performKeywordClassification(ctx, businessName, description, websiteURL)
	}

	// Validate results
	if len(mlResult.Classifications) == 0 {
		mmc.logger.Printf("⚠️ No classifications from ML, using keyword fallback")
		return mmc.performKeywordClassification(ctx, businessName, description, websiteURL)
	}

	// Get the highest confidence classification
	bestClassification := mlResult.Classifications[0]
	for _, classification := range mlResult.Classifications {
		if classification.Confidence > bestClassification.Confidence {
			bestClassification = classification
		}
	}

	// Validate confidence threshold
	const minConfidence = 0.5
	if bestClassification.Confidence < minConfidence {
		mmc.logger.Printf("⚠️ ML confidence too low (%.2f < %.2f), using keyword fallback",
			bestClassification.Confidence, minConfidence)
		return mmc.performKeywordClassification(ctx, businessName, description, websiteURL)
	}

	// Map ML label to database industry
	industryName := mmc.mapMLLabelToIndustry(ctx, bestClassification.Label)
	if industryName == "" {
		mmc.logger.Printf("⚠️ Could not map ML label '%s' to industry, using keyword fallback", bestClassification.Label)
		return mmc.performKeywordClassification(ctx, businessName, description, websiteURL)
	}

	// Get classification codes
	classificationCodes := mmc.getClassificationCodesForIndustry(ctx, industryName)

	// Convert to shared format
	result := &shared.IndustryClassification{
		IndustryCode:         industryName,
		IndustryName:         industryName,
		ConfidenceScore:      bestClassification.Confidence,
		ClassificationMethod: "ml_fallback",
		Keywords:             []string{},
		Description:          fmt.Sprintf("ML-based classification using %s model (fallback)", mlResult.ModelID),
		Evidence:             fmt.Sprintf("ML model prediction with confidence %.2f%%", bestClassification.Confidence*100),
		ProcessingTime:       mlResult.ProcessingTime,
		Metadata: map[string]interface{}{
			"model_id":             mlResult.ModelID,
			"model_version":        mlResult.ModelVersion,
			"all_predictions":      mlResult.Classifications,
			"quality_score":        mlResult.QualityScore,
			"original_label":       bestClassification.Label,
			"mapped_industry":      industryName,
			"classification_codes": classificationCodes,
		},
	}

	return result, nil
}

// mapMLLabelToIndustry maps ML model labels to database industry names
// This handles cases where ML labels don't exactly match database industry names
func (mmc *MultiMethodClassifier) mapMLLabelToIndustry(ctx context.Context, mlLabel string) string {
	// Normalize the label
	normalizedLabel := strings.ToLower(strings.TrimSpace(mlLabel))

	// Handle common cases
	if normalizedLabel == "unknown" || normalizedLabel == "" {
		return ""
	}

	// Try direct match first
	// Check if the label matches an industry name in the database
	// For now, we'll use a simple mapping approach
	// In production, this should query the database for industry names

	// Common industry mappings
	industryMappings := map[string]string{
		"technology":           "Technology",
		"tech":                 "Technology",
		"software":             "Technology",
		"healthcare":           "Healthcare",
		"health":               "Healthcare",
		"medical":              "Healthcare",
		"retail":               "Retail",
		"retailer":             "Retail",
		"manufacturing":        "Manufacturing",
		"manufacturer":         "Manufacturing",
		"financial":            "Financial Services",
		"finance":              "Financial Services",
		"banking":              "Financial Services",
		"construction":         "Construction",
		"professional services": "Professional Services",
		"professional":         "Professional Services",
		"transportation":       "Transportation",
		"transport":            "Transportation",
		"general business":     "General Business",
		"general":              "General Business",
	}

	// Check direct mapping
	if industry, ok := industryMappings[normalizedLabel]; ok {
		return industry
	}

	// Try partial matching
	for key, industry := range industryMappings {
		if strings.Contains(normalizedLabel, key) || strings.Contains(key, normalizedLabel) {
			return industry
		}
	}

	// If no mapping found, try to use the label as-is (capitalized)
	if normalizedLabel != "" {
		// Capitalize first letter
		return strings.Title(normalizedLabel)
	}

	return ""
}

// performDescriptionClassification performs description-based classification using only verified data sources
func (mmc *MultiMethodClassifier) performDescriptionClassification(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*shared.IndustryClassification, error) {
	// SECURITY: Only use verified data sources for classification
	// Business descriptions are user-provided and cannot be trusted
	// Website URLs must be ownership-verified before use

	// Step 1: Validate and filter trusted data sources
	trustedContent := mmc.extractTrustedContent(ctx, businessName, description, websiteURL)

	// Step 2: Extract industry indicators from trusted content only
	industryIndicators := mmc.extractIndustryIndicators(trustedContent)

	// Step 3: Calculate confidence based on trusted data quality
	confidence := mmc.calculateTrustedDataConfidence(industryIndicators, trustedContent)

	// Step 4: Determine industry based on verified indicators
	industryName := mmc.determineIndustryFromIndicators(industryIndicators)

	// Get classification codes for the detected industry
	classificationCodes := mmc.getClassificationCodesForIndustry(ctx, industryName)

	// Convert to shared format
	result := &shared.IndustryClassification{
		IndustryCode:         industryName,
		IndustryName:         industryName,
		ConfidenceScore:      confidence,
		ClassificationMethod: "verified_content_analysis",
		Description:          "Description-based classification using verified data sources only",
		Evidence:             fmt.Sprintf("Verified content analysis identified %d industry indicators", len(industryIndicators)),
		ProcessingTime:       time.Duration(0), // Will be set by caller
		Metadata: map[string]interface{}{
			"industry_indicators":    industryIndicators,
			"trusted_content_length": len(trustedContent),
			"security_validated":     true,
			"data_sources":           mmc.getDataSourceInfo(businessName, description, websiteURL),
			"classification_codes":   classificationCodes,
		},
	}

	return result, nil
}

// calculateEnsembleResult calculates the ensemble result from multiple method results
func (mmc *MultiMethodClassifier) calculateEnsembleResult(
	methodResults []shared.ClassificationMethodResult,
	businessName string,
) *shared.IndustryClassification {
	// Filter successful results
	var successfulResults []shared.ClassificationMethodResult
	for _, method := range methodResults {
		if method.Success && method.Result != nil {
			successfulResults = append(successfulResults, method)
		}
	}

	if len(successfulResults) == 0 {
		// Fallback to default classification
		return &shared.IndustryClassification{
			IndustryCode:         "General Business",
			IndustryName:         "General Business",
			ConfidenceScore:      0.50,
			ClassificationMethod: "ensemble_fallback",
			Description:          "Fallback classification due to method failures",
			Evidence:             "No successful classification methods",
			Metadata: map[string]interface{}{
				"fallback_reason": "no_successful_methods",
			},
		}
	}

	// Calculate weighted average based on method confidence and type
	var totalWeight float64
	var weightedIndustryScores = make(map[string]float64)
	var industryCounts = make(map[string]int)

	for _, method := range successfulResults {
		// Weight based on method type and confidence
		weight := mmc.calculateMethodWeight(method)
		
		// PHASE 3: Adjust weight based on crosswalk consistency
		if method.Result != nil && method.Result.Metadata != nil {
			if codes, ok := method.Result.Metadata["classification_codes"].(shared.ClassificationCodes); ok {
				crosswalkScore := mmc.validateCodesAgainstCrosswalks(context.Background(), codes)
				if crosswalkScore > 0.8 {
					weight *= 1.15 // Boost weight for consistent codes
					mmc.logger.Printf("✅ [Phase 3] Boosted method weight for %s (crosswalk consistency: %.2f)", method.MethodType, crosswalkScore)
				} else if crosswalkScore < 0.5 {
					weight *= 0.85 // Reduce weight for inconsistent codes
					mmc.logger.Printf("⚠️ [Phase 3] Reduced method weight for %s (crosswalk consistency: %.2f)", method.MethodType, crosswalkScore)
				}
			}
		}
		
		totalWeight += weight

		industryName := method.Result.IndustryName
		weightedIndustryScores[industryName] += method.Result.ConfidenceScore * weight
		industryCounts[industryName]++
	}

	// Find the industry with the highest weighted score
	var bestIndustry string
	var bestScore float64

	for industry, score := range weightedIndustryScores {
		normalizedScore := score / totalWeight
		if normalizedScore > bestScore {
			bestScore = normalizedScore
			bestIndustry = industry
		}
	}

	// Calculate ensemble confidence
	ensembleConfidence := mmc.calculateEnsembleConfidence(successfulResults, bestIndustry)

	// Get classification codes for the best industry
	classificationCodes := mmc.getClassificationCodesForIndustry(context.Background(), bestIndustry)

	// Create ensemble result
	result := &shared.IndustryClassification{
		IndustryCode:         bestIndustry,
		IndustryName:         bestIndustry,
		ConfidenceScore:      ensembleConfidence,
		ClassificationMethod: "ensemble",
		Description:          fmt.Sprintf("Ensemble classification from %d methods", len(successfulResults)),
		Evidence:             fmt.Sprintf("Combined results from %d classification methods", len(successfulResults)),
		Metadata: map[string]interface{}{
			"method_count":         len(successfulResults),
			"method_results":       successfulResults,
			"weighted_scores":      weightedIndustryScores,
			"classification_codes": classificationCodes,
		},
	}

	return result
}

// calculateMethodWeight calculates the weight for a classification method
func (mmc *MultiMethodClassifier) calculateMethodWeight(method shared.ClassificationMethodResult) float64 {
	// Base weights for different method types
	baseWeights := map[string]float64{
		"keyword":     0.5, // Keyword matching is reliable
		"ml":          0.4, // ML is sophisticated
		"description": 0.1, // Description analysis is now conservative and supplementary
	}

	baseWeight := baseWeights[method.MethodType]
	if baseWeight == 0 {
		baseWeight = 0.1 // Default weight for unknown methods
	}

	// Adjust weight based on confidence
	confidenceMultiplier := method.Confidence

	// Adjust weight based on processing time (faster is better, but not too much)
	timeMultiplier := 1.0
	if method.ProcessingTime > 5*time.Second {
		timeMultiplier = 0.8 // Penalize very slow methods
	}

	return baseWeight * confidenceMultiplier * timeMultiplier
}

// calculateEnsembleConfidence calculates the ensemble confidence score
func (mmc *MultiMethodClassifier) calculateEnsembleConfidence(
	successfulResults []shared.ClassificationMethodResult,
	bestIndustry string,
) float64 {
	if len(successfulResults) == 0 {
		return 0.0
	}

	// Count how many methods agree on the best industry
	agreementCount := 0
	var totalConfidence float64

	for _, method := range successfulResults {
		if method.Result.IndustryName == bestIndustry {
			agreementCount++
		}
		totalConfidence += method.Confidence
	}

	// Calculate agreement ratio
	agreementRatio := float64(agreementCount) / float64(len(successfulResults))

	// Calculate average confidence
	averageConfidence := totalConfidence / float64(len(successfulResults))

	// Ensemble confidence combines agreement and average confidence
	ensembleConfidence := (agreementRatio * 0.6) + (averageConfidence * 0.4)

	// Ensure confidence is within bounds
	if ensembleConfidence > 1.0 {
		ensembleConfidence = 1.0
	}
	if ensembleConfidence < 0.0 {
		ensembleConfidence = 0.0
	}

	return ensembleConfidence
}

// calculateQualityMetrics calculates quality metrics for the classification
func (mmc *MultiMethodClassifier) calculateQualityMetrics(
	methodResults []shared.ClassificationMethodResult,
) *shared.ClassificationQuality {
	var successfulResults []shared.ClassificationMethodResult
	for _, method := range methodResults {
		if method.Success && method.Result != nil {
			successfulResults = append(successfulResults, method)
		}
	}

	if len(successfulResults) == 0 {
		return &shared.ClassificationQuality{
			OverallQuality:     0.0,
			MethodAgreement:    0.0,
			ConfidenceVariance: 1.0,
			EvidenceStrength:   0.0,
			DataCompleteness:   0.0,
		}
	}

	// Calculate method agreement
	industryCounts := make(map[string]int)
	for _, method := range successfulResults {
		industryCounts[method.Result.IndustryName]++
	}

	maxAgreement := 0
	for _, count := range industryCounts {
		if count > maxAgreement {
			maxAgreement = count
		}
	}
	methodAgreement := float64(maxAgreement) / float64(len(successfulResults))

	// Calculate confidence variance
	var confidences []float64
	for _, method := range successfulResults {
		confidences = append(confidences, method.Confidence)
	}
	confidenceVariance := mmc.calculateVariance(confidences)

	// Calculate evidence strength
	var totalEvidenceStrength float64
	for _, method := range successfulResults {
		evidenceStrength := float64(len(method.Evidence)) * method.Confidence
		totalEvidenceStrength += evidenceStrength
	}
	evidenceStrength := totalEvidenceStrength / float64(len(successfulResults))

	// Calculate data completeness
	dataCompleteness := float64(len(successfulResults)) / 3.0 // 3 methods total
	if dataCompleteness > 1.0 {
		dataCompleteness = 1.0
	}

	// Calculate overall quality
	overallQuality := (methodAgreement * 0.4) + ((1.0 - confidenceVariance) * 0.3) + (evidenceStrength * 0.2) + (dataCompleteness * 0.1)

	return &shared.ClassificationQuality{
		OverallQuality:     overallQuality,
		MethodAgreement:    methodAgreement,
		ConfidenceVariance: confidenceVariance,
		EvidenceStrength:   evidenceStrength,
		DataCompleteness:   dataCompleteness,
	}
}

// generateClassificationReasoning generates human-readable reasoning for the classification
func (mmc *MultiMethodClassifier) generateClassificationReasoning(
	methodResults []shared.ClassificationMethodResult,
	ensembleResult *shared.IndustryClassification,
) string {
	var reasoning strings.Builder

	reasoning.WriteString(fmt.Sprintf("Classification for '%s' was determined using %d methods: ",
		ensembleResult.IndustryName, len(methodResults)))

	var successfulMethods []string
	for _, method := range methodResults {
		if method.Success {
			successfulMethods = append(successfulMethods, method.MethodName)
		}
	}

	reasoning.WriteString(strings.Join(successfulMethods, ", "))
	reasoning.WriteString(". ")

	// Add method-specific reasoning
	for _, method := range methodResults {
		if method.Success && method.Result != nil {
			reasoning.WriteString(fmt.Sprintf("%s method identified '%s' with %.1f%% confidence. ",
				strings.Title(method.MethodType), method.Result.IndustryName, method.Confidence*100))
		}
	}

	reasoning.WriteString(fmt.Sprintf("The ensemble result combines these findings to determine '%s' as the primary classification with %.1f%% confidence.",
		ensembleResult.IndustryName, ensembleResult.ConfidenceScore*100))

	return reasoning.String()
}

// Helper methods

func (mmc *MultiMethodClassifier) extractKeywords(businessName, websiteURL string) []string {
	var keywords []string

	// Extract from business name
	if businessName != "" {
		keywords = append(keywords, strings.ToLower(businessName))
	}

	// Note: Description processing removed for security reasons
	// Business descriptions provided by merchants can be unreliable, misleading, or fraudulent

	// Extract from website URL - now with enhanced content scraping
	if websiteURL != "" {
		// Use enhanced scraper for better results
		scrapingResult := mmc.enhancedScraper.ScrapeWebsite(context.Background(), websiteURL)
		if scrapingResult.Success && len(scrapingResult.Keywords) > 0 {
			keywords = append(keywords, scrapingResult.Keywords...)
			mmc.logger.Printf("✅ Enhanced scraper extracted %d keywords from website content: %v",
				len(scrapingResult.Keywords), scrapingResult.Keywords)
		} else {
			// Enhanced fallback: try to extract meaningful keywords from domain name
			domainKeywords := mmc.extractDomainKeywords(websiteURL)
			if len(domainKeywords) > 0 {
				keywords = append(keywords, domainKeywords...)
				mmc.logger.Printf("⚠️ Enhanced website scraping failed (%s), extracted domain keywords: %v",
					scrapingResult.Error, domainKeywords)
			} else {
				// Final fallback: basic domain name extraction
				if strings.Contains(websiteURL, "://") {
					parts := strings.Split(websiteURL, "://")
					if len(parts) > 1 {
						domain := strings.Split(parts[1], "/")[0]
						domainParts := strings.Split(domain, ".")
						if len(domainParts) > 0 {
							keywords = append(keywords, domainParts[0])
							mmc.logger.Printf("⚠️ Using basic domain name extraction: %s", domainParts[0])
						}
					}
				}
			}
		}
	}

	return keywords
}

// extractKeywordsFromWebsite scrapes website content and extracts business-relevant keywords
// Enhanced with caching and reduced timeout for better performance
func (mmc *MultiMethodClassifier) extractKeywordsFromWebsite(ctx context.Context, websiteURL string) []string {
	startTime := time.Now()
	mmc.logger.Printf("🌐 Starting website scraping for: %s", websiteURL)

	// Check cache first
	if cached := mmc.websiteCache.Get(websiteURL); cached != nil {
		mmc.logger.Printf("✅ Using cached website content for: %s (saved %.2fs)", websiteURL, time.Since(startTime).Seconds())
		return cached
	}

	// Validate URL
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		mmc.logger.Printf("❌ Invalid URL format for %s: %v", websiteURL, err)
		return []string{}
	}

	if parsedURL.Scheme == "" {
		websiteURL = "https://" + websiteURL
		mmc.logger.Printf("🔧 Added HTTPS scheme: %s", websiteURL)
	}

	// Create context with strict timeout (reduced from 15s to 5s)
	scrapeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Create HTTP client with shorter timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:          10,
			IdleConnTimeout:       30 * time.Second,
			DisableCompression:    false,
			ResponseHeaderTimeout: 3 * time.Second,
		},
	}

	// Create request with enhanced headers
	req, err := http.NewRequestWithContext(scrapeCtx, "GET", websiteURL, nil)
	if err != nil {
		mmc.logger.Printf("❌ Failed to create request for %s: %v", websiteURL, err)
		return []string{}
	}

	// Set comprehensive headers with randomization to mimic a real browser
	headers := GetRandomizedHeaders(GetUserAgent())
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	mmc.logger.Printf("📡 Making HTTP request to: %s", websiteURL)

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		mmc.logger.Printf("❌ HTTP request failed for %s: %v", websiteURL, err)
		return []string{}
	}
	defer resp.Body.Close()

	// Log response details
	mmc.logger.Printf("📊 Response received - Status: %d, Content-Type: %s, Content-Length: %d",
		resp.StatusCode, resp.Header.Get("Content-Type"), resp.ContentLength)

	// Check status code with detailed logging
	if resp.StatusCode >= 400 {
		mmc.logger.Printf("❌ HTTP error for %s: %d %s", websiteURL, resp.StatusCode, resp.Status)
		// Try to read error response body
		if body, readErr := io.ReadAll(resp.Body); readErr == nil && len(body) > 0 {
			mmc.logger.Printf("📄 Error response body (first 500 chars): %s", string(body[:min(500, len(body))]))
		}
		return []string{}
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "application/xhtml") {
		mmc.logger.Printf("⚠️ Unexpected content type for %s: %s", websiteURL, contentType)
	}

	// Read response body with size limit (reduced to 1MB for better performance)
	const maxSize = 1 * 1024 * 1024 // 1MB limit
	bodyReader := io.LimitReader(resp.Body, maxSize)
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		mmc.logger.Printf("❌ Failed to read response body from %s: %v", websiteURL, err)
		return []string{}
	}

	mmc.logger.Printf("📄 Read %d bytes from %s", len(body), websiteURL)

	// Check for CAPTCHA before processing
	captchaResult := DetectCAPTCHA(resp, body)
	if captchaResult.Detected {
		mmc.logger.Printf("🚫 CAPTCHA detected (%s) for %s - stopping", captchaResult.Type, websiteURL)
		return []string{} // Stop immediately when CAPTCHA is detected
	}

	// Extract text content from HTML
	textContent := mmc.extractTextFromHTML(string(body))
	mmc.logger.Printf("🧹 Extracted %d characters of text content from HTML", len(textContent))

	// Log sample of extracted text for debugging
	if len(textContent) > 0 {
		sampleText := textContent[:min(200, len(textContent))]
		mmc.logger.Printf("📝 Sample extracted text: %s...", sampleText)
	}

	// Extract business-relevant keywords
	keywords := mmc.extractBusinessKeywords(textContent)

	// Cache result
	mmc.websiteCache.Set(websiteURL, keywords)

	duration := time.Since(startTime)
	mmc.logger.Printf("✅ Website scraping completed for %s in %v - extracted %d keywords: %v (cached)",
		websiteURL, duration, len(keywords), keywords)

	return keywords
}

// extractTextFromHTML extracts clean text content from HTML
func (mmc *MultiMethodClassifier) extractTextFromHTML(htmlContent string) string {
	// Simple HTML tag removal (for production, consider using a proper HTML parser)
	// Remove script and style tags completely
	htmlContent = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`).ReplaceAllString(htmlContent, "")
	htmlContent = regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`).ReplaceAllString(htmlContent, "")

	// Remove HTML tags
	htmlContent = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(htmlContent, " ")

	// Clean up whitespace
	htmlContent = regexp.MustCompile(`\s+`).ReplaceAllString(htmlContent, " ")

	return strings.TrimSpace(htmlContent)
}

// extractBusinessKeywords extracts business-relevant keywords from text content
func (mmc *MultiMethodClassifier) extractBusinessKeywords(textContent string) []string {
	var keywords []string

	// Convert to lowercase for processing
	text := strings.ToLower(textContent)

	// Business-relevant keyword patterns
	businessPatterns := []string{
		// Industry keywords
		`\b(restaurant|cafe|coffee|food|dining|kitchen|catering|bakery|bar|pub|brewery|winery)\b`,
		`\b(technology|software|tech|app|digital|web|mobile|cloud|ai|ml|data|cyber|security)\b`,
		`\b(healthcare|medical|clinic|hospital|doctor|dentist|therapy|wellness|pharmacy)\b`,
		`\b(legal|law|attorney|lawyer|court|litigation|patent|trademark|copyright)\b`,
		`\b(retail|store|shop|ecommerce|online|fashion|clothing|electronics|beauty)\b`,
		`\b(finance|banking|investment|insurance|accounting|tax|financial|credit|loan)\b`,
		`\b(real estate|property|construction|building|architecture|design|interior)\b`,
		`\b(education|school|university|training|learning|course|academy|institute)\b`,
		`\b(consulting|advisory|strategy|management|business|corporate|professional)\b`,
		`\b(manufacturing|production|factory|industrial|automotive|machinery|equipment)\b`,
		`\b(transportation|logistics|shipping|delivery|freight|warehouse|supply chain)\b`,
		`\b(entertainment|media|marketing|advertising|design|creative|art|music|film)\b`,
		`\b(energy|utilities|renewable|solar|wind|oil|gas|power|electricity)\b`,
		`\b(agriculture|farming|food production|crop|livestock|organic|sustainable)\b`,
		`\b(travel|tourism|hospitality|hotel|accommodation|vacation|booking|trip)\b`,
	}

	// Extract keywords using patterns
	for _, pattern := range businessPatterns {
		matches := regexp.MustCompile(pattern).FindAllString(text, -1)
		for _, match := range matches {
			// Remove duplicates and add to keywords
			if !mmc.containsKeyword(keywords, match) {
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
	}

	for _, word := range commonBusinessWords {
		if strings.Contains(text, word) && !mmc.containsKeyword(keywords, word) {
			keywords = append(keywords, word)
		}
	}

	// Limit to top 20 keywords to avoid noise
	if len(keywords) > 20 {
		keywords = keywords[:20]
	}

	return keywords
}

// containsKeyword checks if a keyword already exists in the slice
func (mmc *MultiMethodClassifier) containsKeyword(keywords []string, keyword string) bool {
	for _, k := range keywords {
		if k == keyword {
			return true
		}
	}
	return false
}

// extractDomainKeywords extracts meaningful keywords from a domain name
func (mmc *MultiMethodClassifier) extractDomainKeywords(websiteURL string) []string {
	var keywords []string

	// Clean the URL
	cleanURL := strings.TrimPrefix(websiteURL, "https://")
	cleanURL = strings.TrimPrefix(cleanURL, "http://")
	cleanURL = strings.TrimPrefix(cleanURL, "www.")

	// Extract domain name
	parts := strings.Split(cleanURL, ".")
	if len(parts) == 0 {
		return keywords
	}

	domain := parts[0]

	// Split domain by common separators and extract meaningful words
	domainWords := strings.Fields(strings.ReplaceAll(domain, "-", " "))
	domainWords = append(domainWords, strings.Fields(strings.ReplaceAll(domain, "_", " "))...)

	// Filter out common non-meaningful words and add meaningful ones
	commonWords := map[string]bool{
		"the": true, "and": true, "or": true, "of": true, "in": true, "on": true,
		"at": true, "to": true, "for": true, "with": true, "by": true, "from": true,
		"com": true, "net": true, "org": true, "co": true, "inc": true, "llc": true,
		"corp": true, "ltd": true, "group": true, "company": true, "business": true,
	}

	for _, word := range domainWords {
		word = strings.ToLower(strings.TrimSpace(word))
		if len(word) > 2 && !commonWords[word] && !mmc.containsKeyword(keywords, word) {
			keywords = append(keywords, word)
		}
	}

	// If no meaningful words found, use the domain as a single keyword
	if len(keywords) == 0 && len(domain) > 2 {
		keywords = append(keywords, strings.ToLower(domain))
	}

	return keywords
}

// extractTrustedContent extracts only verified and trusted content for classification
func (mmc *MultiMethodClassifier) extractTrustedContent(ctx context.Context, businessName, description, websiteURL string) string {
	var content strings.Builder

	// Always include business name (it's the primary identifier)
	if businessName != "" {
		content.WriteString(businessName)
		content.WriteString(" ")
	}

	// SECURITY: Skip user-provided description - it cannot be trusted
	// Business descriptions are often manipulated or inaccurate
	mmc.logger.Printf("🔒 SECURITY: Skipping user-provided description for classification")

	// Only include website URL if ownership has been verified
	if websiteURL != "" {
		if mmc.isWebsiteOwnershipVerified(ctx, websiteURL, businessName) {
			content.WriteString(websiteURL)
			mmc.logger.Printf("✅ SECURITY: Using verified website URL: %s", websiteURL)
		} else {
			mmc.logger.Printf("⚠️ SECURITY: Skipping unverified website URL: %s", websiteURL)
		}
	}

	return strings.TrimSpace(content.String())
}

// isWebsiteOwnershipVerified checks if website ownership has been verified
func (mmc *MultiMethodClassifier) isWebsiteOwnershipVerified(ctx context.Context, websiteURL, businessName string) bool {
	// In a real implementation, this would check against a verification database
	// For now, we'll implement basic validation rules

	// Extract domain from URL
	domain := mmc.extractDomainFromURL(websiteURL)
	if domain == "" {
		return false
	}

	// Check if domain matches business name (basic validation)
	if mmc.doesDomainMatchBusinessName(domain, businessName) {
		return true
	}

	// TODO: Integrate with website verification service
	// This should check against the AdvancedVerifier results
	// For now, we'll be conservative and require explicit verification
	return false
}

// extractDomainFromURL extracts domain from URL
func (mmc *MultiMethodClassifier) extractDomainFromURL(url string) string {
	// Simple domain extraction - in production, use proper URL parsing
	if strings.HasPrefix(url, "http://") {
		url = strings.TrimPrefix(url, "http://")
	} else if strings.HasPrefix(url, "https://") {
		url = strings.TrimPrefix(url, "https://")
	}

	// Remove path and query parameters
	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}

	return url
}

// doesDomainMatchBusinessName checks if domain matches business name
func (mmc *MultiMethodClassifier) doesDomainMatchBusinessName(domain, businessName string) bool {
	// Convert to lowercase for comparison
	domain = strings.ToLower(domain)
	businessName = strings.ToLower(businessName)

	// Remove common TLDs
	domain = strings.TrimSuffix(domain, ".com")
	domain = strings.TrimSuffix(domain, ".org")
	domain = strings.TrimSuffix(domain, ".net")
	domain = strings.TrimSuffix(domain, ".co")
	domain = strings.TrimSuffix(domain, ".io")

	// Remove spaces and special characters from business name
	businessName = strings.ReplaceAll(businessName, " ", "")
	businessName = strings.ReplaceAll(businessName, "&", "and")
	businessName = strings.ReplaceAll(businessName, "-", "")
	businessName = strings.ReplaceAll(businessName, "_", "")

	// Check if domain contains business name or vice versa
	return strings.Contains(domain, businessName) || strings.Contains(businessName, domain)
}

// getDataSourceInfo returns information about data sources used
func (mmc *MultiMethodClassifier) getDataSourceInfo(businessName, description, websiteURL string) map[string]interface{} {
	sources := map[string]interface{}{
		"business_name": map[string]interface{}{
			"used":    true,
			"trusted": true,
			"reason":  "Primary business identifier",
		},
		"description": map[string]interface{}{
			"used":    false,
			"trusted": false,
			"reason":  "User-provided data cannot be trusted for classification",
		},
		"website_url": map[string]interface{}{
			"used":    websiteURL != "" && mmc.isWebsiteOwnershipVerified(context.Background(), websiteURL, businessName),
			"trusted": false, // Will be true only if verified
			"reason":  "Website ownership must be verified before use",
		},
	}

	return sources
}

// calculateTrustedDataConfidence calculates confidence based on trusted data quality
func (mmc *MultiMethodClassifier) calculateTrustedDataConfidence(indicators []string, content string) float64 {
	if len(indicators) == 0 {
		return 0.2 // Very low confidence for no indicators from trusted data
	}

	// Base confidence on number of indicators and content length
	indicatorConfidence := float64(len(indicators)) * 0.15 // Lower multiplier for trusted data
	contentConfidence := float64(len(content)) / 2000.0    // Longer content = higher confidence

	confidence := indicatorConfidence + contentConfidence
	if confidence > 0.8 {
		confidence = 0.8 // Cap confidence for description-based method
	}
	if confidence < 0.1 {
		confidence = 0.1
	}

	return confidence
}

func (mmc *MultiMethodClassifier) extractIndustryIndicators(content string) []string {
	// Simple industry indicator extraction
	// In a real implementation, this would be more sophisticated
	indicators := []string{}

	content = strings.ToLower(content)

	// Technology indicators
	if strings.Contains(content, "software") || strings.Contains(content, "tech") || strings.Contains(content, "app") {
		indicators = append(indicators, "technology")
	}

	// Retail indicators
	if strings.Contains(content, "store") || strings.Contains(content, "retail") || strings.Contains(content, "shop") {
		indicators = append(indicators, "retail")
	}

	// Financial indicators
	if strings.Contains(content, "bank") || strings.Contains(content, "finance") || strings.Contains(content, "investment") {
		indicators = append(indicators, "financial")
	}

	// Healthcare indicators
	if strings.Contains(content, "health") || strings.Contains(content, "medical") || strings.Contains(content, "clinic") {
		indicators = append(indicators, "healthcare")
	}

	return indicators
}

func (mmc *MultiMethodClassifier) calculateDescriptionConfidence(indicators []string, content string) float64 {
	if len(indicators) == 0 {
		return 0.3 // Low confidence for no indicators
	}

	// Base confidence on number of indicators and content length
	indicatorConfidence := float64(len(indicators)) * 0.2
	contentConfidence := float64(len(content)) / 1000.0 // Longer content = higher confidence

	confidence := indicatorConfidence + contentConfidence
	if confidence > 0.9 {
		confidence = 0.9
	}
	if confidence < 0.1 {
		confidence = 0.1
	}

	return confidence
}

func (mmc *MultiMethodClassifier) determineIndustryFromIndicators(indicators []string) string {
	if len(indicators) == 0 {
		return "General Business"
	}

	// Count indicator frequency
	indicatorCounts := make(map[string]int)
	for _, indicator := range indicators {
		indicatorCounts[indicator]++
	}

	// Find most common indicator
	var bestIndicator string
	var maxCount int
	for indicator, count := range indicatorCounts {
		if count > maxCount {
			maxCount = count
			bestIndicator = indicator
		}
	}

	// Map indicators to industry names
	industryMap := map[string]string{
		"technology": "Technology",
		"retail":     "Retail Trade",
		"financial":  "Financial Services",
		"healthcare": "Healthcare",
	}

	if industry, exists := industryMap[bestIndicator]; exists {
		return industry
	}

	return "General Business"
}

func (mmc *MultiMethodClassifier) calculateVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	// Calculate mean
	var sum float64
	for _, value := range values {
		sum += value
	}
	mean := sum / float64(len(values))

	// Calculate variance
	var variance float64
	for _, value := range values {
		diff := value - mean
		variance += diff * diff
	}
	variance /= float64(len(values))

	return variance
}

func (mmc *MultiMethodClassifier) generateRequestID() string {
	return fmt.Sprintf("multi_method_%d", time.Now().UnixNano())
}

// convertClassificationCodes converts database classification codes to the expected format
func (mmc *MultiMethodClassifier) convertClassificationCodes(codes []*repository.ClassificationCode) shared.ClassificationCodes {
	classificationCodes := shared.ClassificationCodes{
		MCC:   []shared.MCCCode{},
		SIC:   []shared.SICCode{},
		NAICS: []shared.NAICSCode{},
	}

	for _, code := range codes {
		if code == nil {
			continue
		}

		// Calculate confidence based on the classification result
		confidence := 0.8 // Default confidence for database codes

		switch strings.ToUpper(code.CodeType) {
		case "MCC":
			classificationCodes.MCC = append(classificationCodes.MCC, shared.MCCCode{
				Code:        code.Code,
				Description: code.Description,
				Confidence:  confidence,
			})
		case "SIC":
			classificationCodes.SIC = append(classificationCodes.SIC, shared.SICCode{
				Code:        code.Code,
				Description: code.Description,
				Confidence:  confidence,
			})
		case "NAICS":
			classificationCodes.NAICS = append(classificationCodes.NAICS, shared.NAICSCode{
				Code:        code.Code,
				Description: code.Description,
				Confidence:  confidence,
			})
		}
	}

	// Limit to top 3 codes per type for better performance
	if len(classificationCodes.MCC) > 3 {
		classificationCodes.MCC = classificationCodes.MCC[:3]
	}
	if len(classificationCodes.SIC) > 3 {
		classificationCodes.SIC = classificationCodes.SIC[:3]
	}
	if len(classificationCodes.NAICS) > 3 {
		classificationCodes.NAICS = classificationCodes.NAICS[:3]
	}

	mmc.logger.Printf("✅ Converted %d classification codes: %d MCC, %d SIC, %d NAICS",
		len(codes), len(classificationCodes.MCC), len(classificationCodes.SIC), len(classificationCodes.NAICS))

	return classificationCodes
}

// getClassificationCodesForIndustry retrieves classification codes for a given industry name
func (mmc *MultiMethodClassifier) getClassificationCodesForIndustry(ctx context.Context, industryName string) shared.ClassificationCodes {
	// Try to find the industry by name
	industry, err := mmc.keywordRepo.GetIndustryByName(ctx, industryName)
	if err != nil {
		mmc.logger.Printf("⚠️ Failed to get industry by name '%s': %v", industryName, err)
		return shared.ClassificationCodes{
			MCC:   []shared.MCCCode{},
			SIC:   []shared.SICCode{},
			NAICS: []shared.NAICSCode{},
		}
	}

	// Get classification codes for the industry
	codes, err := mmc.keywordRepo.GetCachedClassificationCodes(ctx, industry.ID)
	if err != nil {
		mmc.logger.Printf("⚠️ Failed to get classification codes for industry %d: %v", industry.ID, err)
		return shared.ClassificationCodes{
			MCC:   []shared.MCCCode{},
			SIC:   []shared.SICCode{},
			NAICS: []shared.NAICSCode{},
		}
	}

	// Convert to the expected format
	return mmc.convertClassificationCodes(codes)
}

func (mmc *MultiMethodClassifier) recordMultiMethodMetrics(
	ctx context.Context,
	requestID string,
	businessName string,
	methodResults []shared.ClassificationMethodResult,
	finalResult *MultiMethodClassificationResult,
	responseTime time.Duration,
	err error,
) {
	if mmc.monitor == nil {
		return // No monitoring configured
	}

	// Prepare metrics data
	metrics := &ClassificationAccuracyMetrics{
		Timestamp:            time.Now(),
		RequestID:            requestID,
		PredictedIndustry:    finalResult.PrimaryClassification.IndustryName,
		PredictedConfidence:  finalResult.EnsembleConfidence,
		ResponseTimeMs:       float64(responseTime.Nanoseconds()) / 1e6,
		ClassificationMethod: stringPtr("multi_method_ensemble"),
		KeywordsUsed:         []string{}, // Will be populated from method results
		ConfidenceThreshold:  0.5,
		CreatedAt:            time.Now(),
	}

	// Set error message if there was an error
	if err != nil {
		errorMsg := err.Error()
		metrics.ErrorMessage = &errorMsg
	}

	// Record metrics asynchronously
	go func() {
		// Note: This would call the actual monitoring method when implemented
		// if err := mmc.monitor.RecordClassificationMetrics(ctx, metrics); err != nil {
		//     mmc.logger.Printf("⚠️ Failed to record multi-method metrics: %v", err)
		// }
	}()
}

// ============================================================================
// PHASE 1: Keyword-Enhanced ML Input
// ============================================================================

// enhanceMLInputWithKeywords enhances ML input content with keyword context
// This helps ML models understand the business better by providing relevant keywords
func (mmc *MultiMethodClassifier) enhanceMLInputWithKeywords(content string, keywords []string) string {
	if len(keywords) == 0 {
		return content
	}

	// Append keywords as context to help ML understand the business
	keywordContext := fmt.Sprintf("\n\nRelevant Keywords: %s", strings.Join(keywords, ", "))
	
	// Limit keyword context to avoid overwhelming the ML model
	const maxKeywordContextLength = 200
	if len(keywordContext) > maxKeywordContextLength {
		// Take first N keywords that fit
		maxKeywords := 10
		if len(keywords) < maxKeywords {
			maxKeywords = len(keywords)
		}
		keywordList := strings.Join(keywords[:maxKeywords], ", ")
		keywordContext = fmt.Sprintf("\n\nRelevant Keywords: %s", keywordList)
		if len(keywordContext) > maxKeywordContextLength {
			keywordContext = keywordContext[:maxKeywordContextLength] + "..."
		}
	}

	return content + keywordContext
}

// validateMLAgainstKeywords validates ML classification results against keyword-based classification
// This improves accuracy by comparing ML predictions with keyword matches
func (mmc *MultiMethodClassifier) validateMLAgainstKeywords(
	ctx context.Context,
	businessName, description, websiteURL string,
	mlIndustry string,
	mlConfidence float64,
	keywords []string,
	classificationCodes shared.ClassificationCodes,
	enhancedResp *infrastructure.EnhancedClassificationResponse,
) *shared.IndustryClassification {
	// Get keyword-based classification for comparison
	keywordResult, err := mmc.performKeywordClassification(ctx, businessName, description, websiteURL)
	if err != nil {
		mmc.logger.Printf("⚠️ [Phase 1] Keyword classification failed for validation: %v, using ML result as-is", err)
		// Return ML result without validation if keyword classification fails
		return mmc.buildMLResult(mlIndustry, mlConfidence, classificationCodes, enhancedResp)
	}

	// Compare ML and keyword results
	mlIndustryLower := strings.ToLower(mlIndustry)
	keywordIndustryLower := strings.ToLower(keywordResult.IndustryName)

	// Check if industries match (case-insensitive)
	industriesMatch := mlIndustryLower == keywordIndustryLower ||
		strings.Contains(mlIndustryLower, keywordIndustryLower) ||
		strings.Contains(keywordIndustryLower, mlIndustryLower)

	adjustedConfidence := mlConfidence
	adjustedIndustry := mlIndustry
	validationDetails := make(map[string]interface{})

	if industriesMatch {
		// Both methods agree - boost confidence
		confidenceBoost := 0.2 // 20% boost
		newConfidence := mlConfidence * (1.0 + confidenceBoost)
		if newConfidence > 1.0 {
			adjustedConfidence = 1.0
		} else {
			adjustedConfidence = newConfidence
		}
		mmc.logger.Printf("✅ [Phase 1] ML and keywords agree on '%s' - boosting confidence: %.2f%% -> %.2f%%",
			mlIndustry, mlConfidence*100, adjustedConfidence*100)
		validationDetails["validation_status"] = "agreement"
		validationDetails["confidence_boost"] = confidenceBoost
		validationDetails["keyword_industry"] = keywordResult.IndustryName
		validationDetails["keyword_confidence"] = keywordResult.ConfidenceScore
	} else {
		// Methods disagree - adjust based on confidence levels
		if mlConfidence < 0.6 && keywordResult.ConfidenceScore > 0.7 {
			// ML has low confidence, keywords have high confidence - prefer keywords
			adjustedIndustry = keywordResult.IndustryName
			adjustedConfidence = keywordResult.ConfidenceScore * 0.9 // Slight reduction for method disagreement
			mmc.logger.Printf("🔄 [Phase 1] ML confidence low (%.2f%%) but keywords confident (%.2f%%) - using keyword result: %s",
				mlConfidence*100, keywordResult.ConfidenceScore*100, adjustedIndustry)
			validationDetails["validation_status"] = "prefer_keywords"
			validationDetails["reason"] = "ml_low_confidence_keywords_high"
		} else if mlConfidence > 0.8 && keywordResult.ConfidenceScore < 0.5 {
			// ML has high confidence, keywords have low confidence - keep ML result but reduce confidence slightly
			adjustedConfidence = mlConfidence * 0.95 // Slight reduction for disagreement
			mmc.logger.Printf("⚠️ [Phase 1] ML confident (%.2f%%) but keywords disagree (%.2f%%) - keeping ML result with slight reduction",
				mlConfidence*100, keywordResult.ConfidenceScore*100)
			validationDetails["validation_status"] = "prefer_ml"
			validationDetails["reason"] = "ml_high_confidence_keywords_low"
		} else {
			// Both have medium confidence - keep ML result but note disagreement
			adjustedConfidence = mlConfidence * 0.9 // Moderate reduction for disagreement
			mmc.logger.Printf("⚠️ [Phase 1] ML and keywords disagree - ML: %s (%.2f%%), Keywords: %s (%.2f%%) - keeping ML with reduced confidence",
				mlIndustry, mlConfidence*100, keywordResult.IndustryName, keywordResult.ConfidenceScore*100)
			validationDetails["validation_status"] = "disagreement"
			validationDetails["keyword_industry"] = keywordResult.IndustryName
			validationDetails["keyword_confidence"] = keywordResult.ConfidenceScore
		}
	}

	// Build result with validation metadata
	result := mmc.buildMLResult(adjustedIndustry, adjustedConfidence, classificationCodes, enhancedResp)
	
	// Add validation details to metadata
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["keyword_validation"] = validationDetails
	result.Metadata["original_ml_confidence"] = mlConfidence
	result.Metadata["original_ml_industry"] = mlIndustry
	result.Metadata["keywords_used"] = keywords

	return result
}

// buildMLResult builds IndustryClassification from ML response with validation adjustments
func (mmc *MultiMethodClassifier) buildMLResult(
	industry string,
	confidence float64,
	classificationCodes shared.ClassificationCodes,
	enhancedResp *infrastructure.EnhancedClassificationResponse,
) *shared.IndustryClassification {
	// Build all industry scores map
	allScores := make(map[string]float64)
	for _, classification := range enhancedResp.Classifications {
		allScores[classification.Label] = classification.Confidence
	}

	return &shared.IndustryClassification{
		IndustryCode:         industry,
		IndustryName:         industry,
		PrimaryIndustry:     industry,
		ConfidenceScore:      confidence,
		ClassificationMethod: "ml_distilbart",
		ContentSummary:       enhancedResp.Summary,
		Explanation:          enhancedResp.Explanation,
		AllIndustryScores:    allScores,
		ProcessingTime:       time.Duration(0), // Will be set by caller
		Metadata: map[string]interface{}{
			"model_version":        enhancedResp.ModelVersion,
			"quantization_enabled":  enhancedResp.QuantizationEnabled,
			"all_classifications":  enhancedResp.Classifications,
			"classification_codes":  classificationCodes,
		},
	}
}

// ============================================================================
// PHASE 3: Ensemble Enhancement with Crosswalks
// ============================================================================

// validateCodesAgainstCrosswalks validates codes against crosswalk relationships
// Returns a consistency score (0.0-1.0) indicating how well codes match crosswalks
func (mmc *MultiMethodClassifier) validateCodesAgainstCrosswalks(
	ctx context.Context,
	codes shared.ClassificationCodes,
) float64 {
	// Get code metadata repository from keyword repo if available
	codeMetadataRepo := mmc.getCodeMetadataRepository()
	if codeMetadataRepo == nil {
		return 0.0 // No metadata repo available
	}

	totalChecks := 0
	consistentChecks := 0

	// Check MCC codes
	for _, mcc := range codes.MCC {
		crosswalks, err := codeMetadataRepo.GetCrosswalkCodes(ctx, "MCC", mcc.Code)
		if err == nil && len(crosswalks) > 0 {
			for _, cw := range crosswalks {
				totalChecks++
				if mmc.codeExistsInResults(cw.Code, cw.CodeType, codes) {
					consistentChecks++
				}
			}
		}
	}

	// Check NAICS codes
	for _, naics := range codes.NAICS {
		crosswalks, err := codeMetadataRepo.GetCrosswalkCodes(ctx, "NAICS", naics.Code)
		if err == nil && len(crosswalks) > 0 {
			for _, cw := range crosswalks {
				totalChecks++
				if mmc.codeExistsInResults(cw.Code, cw.CodeType, codes) {
					consistentChecks++
				}
			}
		}
	}

	// Check SIC codes
	for _, sic := range codes.SIC {
		crosswalks, err := codeMetadataRepo.GetCrosswalkCodes(ctx, "SIC", sic.Code)
		if err == nil && len(crosswalks) > 0 {
			for _, cw := range crosswalks {
				totalChecks++
				if mmc.codeExistsInResults(cw.Code, cw.CodeType, codes) {
					consistentChecks++
				}
			}
		}
	}

	if totalChecks == 0 {
		return 0.0 // No crosswalks to check
	}

	return float64(consistentChecks) / float64(totalChecks)
}

// codeExistsInResults checks if a code exists in the classification results
func (mmc *MultiMethodClassifier) codeExistsInResults(code, codeType string, codes shared.ClassificationCodes) bool {
	switch codeType {
	case "MCC":
		for _, mcc := range codes.MCC {
			if mcc.Code == code {
				return true
			}
		}
	case "SIC":
		for _, sic := range codes.SIC {
			if sic.Code == code {
				return true
			}
		}
	case "NAICS":
		for _, naics := range codes.NAICS {
			if naics.Code == code {
				return true
			}
		}
	}
	return false
}

// getCodeMetadataRepository gets code metadata repository from keyword repo
func (mmc *MultiMethodClassifier) getCodeMetadataRepository() *repository.CodeMetadataRepository {
	// Try to get from SupabaseKeywordRepository
	if supabaseRepo, ok := mmc.keywordRepo.(*repository.SupabaseKeywordRepository); ok {
		client := supabaseRepo.GetSupabaseClient()
		if client != nil {
			return repository.NewCodeMetadataRepository(client, mmc.logger)
		}
	}
	return nil
}

// selectCodesWithCrosswalkConsistency selects codes that have crosswalk relationships
// This ensures consistency across code types when methods disagree
func (mmc *MultiMethodClassifier) selectCodesWithCrosswalkConsistency(
	ctx context.Context,
	mlCodes, keywordCodes, descriptionCodes shared.ClassificationCodes,
) shared.ClassificationCodes {
	codeMetadataRepo := mmc.getCodeMetadataRepository()
	if codeMetadataRepo == nil {
		// No metadata repo - return ML codes as default
		return mlCodes
	}

	// Score each method's codes by crosswalk consistency
	mlScore := mmc.validateCodesAgainstCrosswalks(ctx, mlCodes)
	keywordScore := mmc.validateCodesAgainstCrosswalks(ctx, keywordCodes)
	descriptionScore := mmc.validateCodesAgainstCrosswalks(ctx, descriptionCodes)

	// Prefer codes with highest crosswalk consistency
	if mlScore >= keywordScore && mlScore >= descriptionScore {
		mmc.logger.Printf("📊 [Phase 3] Selected ML codes (crosswalk consistency: %.2f)", mlScore)
		return mlCodes
	} else if keywordScore >= descriptionScore {
		mmc.logger.Printf("📊 [Phase 3] Selected keyword codes (crosswalk consistency: %.2f)", keywordScore)
		return keywordCodes
	} else {
		mmc.logger.Printf("📊 [Phase 3] Selected description codes (crosswalk consistency: %.2f)", descriptionScore)
		return descriptionCodes
	}
}

