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

	"github.com/pcraw4d/business-verification/internal/classification/repository"
	"github.com/pcraw4d/business-verification/internal/machine_learning"
	"github.com/pcraw4d/business-verification/internal/shared"
)

// MultiMethodClassifier provides enhanced classification using multiple methods
type MultiMethodClassifier struct {
	keywordRepo              repository.KeywordRepository
	mlClassifier             *machine_learning.ContentClassifier
	weightedConfidenceScorer *WeightedConfidenceScorer
	reasoningEngine          *ReasoningEngine
	qualityMetricsService    *QualityMetricsService
	logger                   *log.Logger
	monitor                  *ClassificationAccuracyMonitoring
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
		weightedConfidenceScorer: NewWeightedConfidenceScorer(logger),
		reasoningEngine:          NewReasoningEngine(logger),
		qualityMetricsService:    NewQualityMetricsService(logger),
		logger:                   logger,
		monitor:                  nil, // Will be set separately if monitoring is needed
	}
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
	keywords := mmc.extractKeywords(businessName, description, websiteURL)

	// Classify using keyword repository
	classificationResult, err := mmc.keywordRepo.ClassifyBusinessByKeywords(ctx, keywords)
	if err != nil {
		return nil, fmt.Errorf("keyword classification failed: %w", err)
	}

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
			"classification_codes": classificationResult.Codes,
		},
	}

	return result, nil
}

// performMLClassification performs ML-based classification
func (mmc *MultiMethodClassifier) performMLClassification(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*shared.IndustryClassification, error) {
	if mmc.mlClassifier == nil {
		return nil, fmt.Errorf("ML classifier not available")
	}

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
			"model_id":        mlResult.ModelID,
			"model_version":   mlResult.ModelVersion,
			"all_predictions": mlResult.Classifications,
			"quality_score":   mlResult.QualityScore,
		},
	}

	return result, nil
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

	// Create ensemble result
	result := &shared.IndustryClassification{
		IndustryCode:         bestIndustry,
		IndustryName:         bestIndustry,
		ConfidenceScore:      ensembleConfidence,
		ClassificationMethod: "ensemble",
		Description:          fmt.Sprintf("Ensemble classification from %d methods", len(successfulResults)),
		Evidence:             fmt.Sprintf("Combined results from %d classification methods", len(successfulResults)),
		Metadata: map[string]interface{}{
			"method_count":    len(successfulResults),
			"method_results":  successfulResults,
			"weighted_scores": weightedIndustryScores,
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

func (mmc *MultiMethodClassifier) extractKeywords(businessName, description, websiteURL string) []string {
	var keywords []string

	// Extract from business name
	if businessName != "" {
		keywords = append(keywords, strings.ToLower(businessName))
	}

	// Extract from description
	if description != "" {
		words := strings.Fields(strings.ToLower(description))
		keywords = append(keywords, words...)
	}

	// Extract from website URL - now with actual content scraping
	if websiteURL != "" {
		// First, try to scrape actual website content
		scrapedKeywords := mmc.extractKeywordsFromWebsite(context.Background(), websiteURL)
		if len(scrapedKeywords) > 0 {
			keywords = append(keywords, scrapedKeywords...)
			mmc.logger.Printf("✅ Extracted %d keywords from website content: %v", len(scrapedKeywords), scrapedKeywords)
		} else {
			// Fallback to domain name extraction if scraping fails
			if strings.Contains(websiteURL, "://") {
				parts := strings.Split(websiteURL, "://")
				if len(parts) > 1 {
					domain := strings.Split(parts[1], "/")[0]
					domainParts := strings.Split(domain, ".")
					keywords = append(keywords, domainParts[0])
					mmc.logger.Printf("⚠️ Website scraping failed, using domain name: %s", domainParts[0])
				}
			}
		}
	}

	return keywords
}

// extractKeywordsFromWebsite scrapes website content and extracts business-relevant keywords
func (mmc *MultiMethodClassifier) extractKeywordsFromWebsite(ctx context.Context, websiteURL string) []string {
	startTime := time.Now()
	mmc.logger.Printf("🌐 Starting website scraping for: %s", websiteURL)
	
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

	// Create HTTP client with enhanced configuration
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  false,
		},
	}

	// Create request with enhanced headers
	req, err := http.NewRequestWithContext(ctx, "GET", websiteURL, nil)
	if err != nil {
		mmc.logger.Printf("❌ Failed to create request for %s: %v", websiteURL, err)
		return []string{}
	}

	// Set comprehensive headers to mimic a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Cache-Control", "max-age=0")

	mmc.logger.Printf("📡 Making HTTP request to: %s", websiteURL)

	// Make request with timeout context
	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

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

	// Read response body with size limit
	maxSize := int64(5 * 1024 * 1024) // 5MB limit
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSize))
	if err != nil {
		mmc.logger.Printf("❌ Failed to read response body from %s: %v", websiteURL, err)
		return []string{}
	}

	mmc.logger.Printf("📄 Read %d bytes from %s", len(body), websiteURL)

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
	
	duration := time.Since(startTime)
	mmc.logger.Printf("✅ Website scraping completed for %s in %v - extracted %d keywords: %v", 
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
