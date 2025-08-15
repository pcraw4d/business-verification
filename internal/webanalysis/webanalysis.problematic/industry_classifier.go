package webanalysis

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// MLClassificationRequest represents a request for ML-based classification
type MLClassificationRequest struct {
	BusinessName        string                 `json:"business_name"`
	BusinessDescription string                 `json:"business_description"`
	Keywords            []string               `json:"keywords"`
	WebsiteContent      string                 `json:"website_content"`
	IndustryHints       []string               `json:"industry_hints"`
	GeographicRegion    string                 `json:"geographic_region"`
	BusinessType        string                 `json:"business_type"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// MLClassificationResult represents the result of ML-based classification
type MLClassificationResult struct {
	IndustryCode       string                 `json:"industry_code"`
	IndustryName       string                 `json:"industry_name"`
	ConfidenceScore    float64                `json:"confidence_score"`
	ModelType          string                 `json:"model_type"`
	ModelVersion       string                 `json:"model_version"`
	InferenceTime      time.Duration          `json:"inference_time"`
	ModelPredictions   []ModelPrediction      `json:"model_predictions"`
	EnsembleScore      float64                `json:"ensemble_score"`
	FeatureImportance  map[string]float64     `json:"feature_importance"`
	ProcessingMetadata map[string]interface{} `json:"processing_metadata"`
}

// ModelPrediction represents a prediction from a single model
type ModelPrediction struct {
	ModelID         string  `json:"model_id"`
	ModelType       string  `json:"model_type"`
	IndustryCode    string  `json:"industry_code"`
	IndustryName    string  `json:"industry_name"`
	ConfidenceScore float64 `json:"confidence_score"`
	RawScore        float64 `json:"raw_score"`
}

// MLClassifier interface for ML-based classification
type MLClassifier interface {
	Classify(ctx context.Context, request *MLClassificationRequest) (*MLClassificationResult, error)
}

// ModelManager interface for model management
type ModelManager interface {
	GetModelByType(ctx context.Context, modelType string) (*ModelInfo, error)
}

// ModelInfo represents information about a loaded model
type ModelInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Version     string                 `json:"version"`
	Status      string                 `json:"status"`
	LoadedAt    time.Time              `json:"loaded_at"`
	LastUsed    time.Time              `json:"last_used"`
	UsageCount  int64                  `json:"usage_count"`
	Performance *ModelPerformance      `json:"performance,omitempty"`
	Config      *ModelConfig           `json:"config,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ModelPerformance represents performance metrics for a model
type ModelPerformance struct {
	Accuracy        float64   `json:"accuracy"`
	Precision       float64   `json:"precision"`
	Recall          float64   `json:"recall"`
	F1Score         float64   `json:"f1_score"`
	InferenceTime   float64   `json:"inference_time_ms"`
	Throughput      float64   `json:"throughput_requests_per_sec"`
	MemoryUsage     float64   `json:"memory_usage_mb"`
	LastEvaluated   time.Time `json:"last_evaluated"`
	EvaluationCount int       `json:"evaluation_count"`
}

// ModelConfig represents configuration for a model
type ModelConfig struct {
	ModelType       string                 `json:"model_type"`
	ModelPath       string                 `json:"model_path"`
	Version         string                 `json:"version"`
	MaxBatchSize    int                    `json:"max_batch_size"`
	Timeout         time.Duration          `json:"timeout"`
	FallbackEnabled bool                   `json:"fallback_enabled"`
	Parameters      map[string]interface{} `json:"parameters"`
	Preprocessing   map[string]interface{} `json:"preprocessing"`
	Postprocessing  map[string]interface{} `json:"postprocessing"`
}

// IndustryClassifier handles industry classification with multi-industry support
type IndustryClassifier struct {
	classifiers            map[string]IndustryClassifierRule
	keywords               map[string][]string
	confidenceRules        []ConfidenceRule
	semanticAnalyzer       *SemanticAnalyzer
	contentQualityAnalyzer *EnhancedContentAnalyzer
	industryPatterns       map[string][]string
	evidenceExtractors     map[string]EvidenceExtractor

	// ML integration
	mlClassifier          MLClassifier
	modelManager          ModelManager
	mlEnabled             bool
	mlConfidenceThreshold float64
	ensembleEnabled       bool
}

// IndustryClassifierRule represents a classification rule
type IndustryClassifierRule struct {
	Industry   string
	Keywords   []string
	Weight     float64
	Confidence float64
	NAICSCode  string
	SICCode    string
}

// ConfidenceRule represents a confidence calculation rule
type ConfidenceRule struct {
	Condition     string
	Multiplier    float64
	MinConfidence float64
}

// EvidenceExtractor represents an evidence extraction function
type EvidenceExtractor func(content string, rule IndustryClassifierRule) []string

// IndustryClassificationResult represents enhanced industry classification results
type IndustryClassificationResult struct {
	Industry           string                 `json:"industry"`
	NAICSCode          string                 `json:"naics_code"`
	SICCode            string                 `json:"sic_code"`
	Confidence         float64                `json:"confidence"`
	Evidence           []string               `json:"evidence"`
	Keywords           []string               `json:"keywords"`
	SemanticScore      float64                `json:"semantic_score"`
	ContentQuality     float64                `json:"content_quality"`
	IndustryPatterns   []string               `json:"industry_patterns"`
	ClassificationTime time.Time              `json:"classification_time"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// NewIndustryClassifier creates a new industry classifier
func NewIndustryClassifier(semanticAnalyzer *SemanticAnalyzer, contentQualityAnalyzer *EnhancedContentAnalyzer) *IndustryClassifier {
	ic := &IndustryClassifier{
		classifiers:            make(map[string]IndustryClassifierRule),
		keywords:               make(map[string][]string),
		semanticAnalyzer:       semanticAnalyzer,
		contentQualityAnalyzer: contentQualityAnalyzer,
		industryPatterns:       make(map[string][]string),
		evidenceExtractors:     make(map[string]EvidenceExtractor),

		// ML integration defaults
		mlEnabled:             false,
		mlConfidenceThreshold: 0.7,
		ensembleEnabled:       true,
	}

	// Initialize with basic industry classifiers
	ic.initializeClassifiers()
	ic.initializeIndustryPatterns()
	ic.initializeEvidenceExtractors()

	return ic
}

// NewIndustryClassifierWithML creates a new industry classifier with ML integration
func NewIndustryClassifierWithML(semanticAnalyzer *SemanticAnalyzer, contentQualityAnalyzer *EnhancedContentAnalyzer,
	mlClassifier MLClassifier, modelManager ModelManager) *IndustryClassifier {
	ic := &IndustryClassifier{
		classifiers:            make(map[string]IndustryClassifierRule),
		keywords:               make(map[string][]string),
		semanticAnalyzer:       semanticAnalyzer,
		contentQualityAnalyzer: contentQualityAnalyzer,
		industryPatterns:       make(map[string][]string),
		evidenceExtractors:     make(map[string]EvidenceExtractor),

		// ML integration
		mlClassifier:          mlClassifier,
		modelManager:          modelManager,
		mlEnabled:             true,
		mlConfidenceThreshold: 0.7,
		ensembleEnabled:       true,
	}

	// Initialize with basic industry classifiers
	ic.initializeClassifiers()
	ic.initializeIndustryPatterns()
	ic.initializeEvidenceExtractors()

	return ic
}

// ClassifyContent performs industry classification on content
func (ic *IndustryClassifier) ClassifyContent(ctx context.Context, content string, maxResults int) ([]IndustryClassification, error) {
	if maxResults == 0 {
		maxResults = 3 // Default to top 3
	}

	// Try ML-based classification first if enabled
	if ic.mlEnabled && ic.mlClassifier != nil {
		mlResults, err := ic.performMLClassification(ctx, content, maxResults)
		if err == nil && len(mlResults) > 0 {
			// Check if ML results meet confidence threshold
			highConfidenceResults := ic.filterHighConfidenceResults(mlResults)
			if len(highConfidenceResults) > 0 {
				return highConfidenceResults, nil
			}
		}
	}

	// Fall back to traditional classification
	return ic.performTraditionalClassification(ctx, content, maxResults)
}

// performMLClassification performs ML-based classification
func (ic *IndustryClassifier) performMLClassification(ctx context.Context, content string, maxResults int) ([]IndustryClassification, error) {
	// Create ML classification request
	request := &MLClassificationRequest{
		BusinessDescription: content,
		WebsiteContent:      content,
		Metadata:            make(map[string]interface{}),
	}

	// Perform ML classification
	mlResult, err := ic.mlClassifier.Classify(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("ML classification failed: %w", err)
	}

	// Convert ML result to IndustryClassification
	results := make([]IndustryClassification, 0)

	// Add main result
	mainResult := IndustryClassification{
		Industry:   mlResult.IndustryName,
		NAICSCode:  mlResult.IndustryCode,
		SICCode:    "", // Would need mapping from NAICS to SIC
		Confidence: mlResult.ConfidenceScore,
		Evidence:   fmt.Sprintf("ML classification (model: %s, ensemble_score: %.3f)", mlResult.ModelType, mlResult.EnsembleScore),
		Keywords:   []string{}, // Would be extracted from feature importance
		Metadata: map[string]interface{}{
			"ml_model_type":      mlResult.ModelType,
			"ml_model_version":   mlResult.ModelVersion,
			"ensemble_score":     mlResult.EnsembleScore,
			"inference_time_ms":  mlResult.InferenceTime.Milliseconds(),
			"feature_importance": mlResult.FeatureImportance,
		},
	}
	results = append(results, mainResult)

	// Add ensemble predictions if available
	if ic.ensembleEnabled && len(mlResult.ModelPredictions) > 1 {
		for _, pred := range mlResult.ModelPredictions {
			if pred.ConfidenceScore >= ic.mlConfidenceThreshold {
				ensembleResult := IndustryClassification{
					Industry:   pred.IndustryName,
					NAICSCode:  pred.IndustryCode,
					SICCode:    "", // Would need mapping
					Confidence: pred.ConfidenceScore,
					Evidence:   fmt.Sprintf("Ensemble prediction (model: %s)", pred.ModelType),
					Keywords:   []string{},
					Metadata: map[string]interface{}{
						"ensemble_model_id":   pred.ModelID,
						"ensemble_model_type": pred.ModelType,
						"raw_score":           pred.RawScore,
					},
				}
				results = append(results, ensembleResult)
			}
		}
	}

	// Limit results to maxResults
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results, nil
}

// performTraditionalClassification performs traditional keyword-based classification
func (ic *IndustryClassifier) performTraditionalClassification(ctx context.Context, content string, maxResults int) ([]IndustryClassification, error) {
	// Normalize content
	normalizedContent := strings.ToLower(content)

	// Calculate scores for each industry
	scores := make(map[string]float64)
	evidence := make(map[string]string)

	for industry, rule := range ic.classifiers {
		score := ic.calculateIndustryScore(normalizedContent, rule)
		if score > 0 {
			scores[industry] = score
			evidence[industry] = ic.findEvidence(normalizedContent, rule)
		}
	}

	// Sort industries by score and get top results
	topIndustries := ic.getTopIndustries(scores, maxResults)

	// Create classification results
	var results []IndustryClassification
	for _, industry := range topIndustries {
		rule := ic.classifiers[industry]
		confidence := ic.calculateConfidence(scores[industry], rule, normalizedContent)

		classification := IndustryClassification{
			Industry:   industry,
			NAICSCode:  rule.NAICSCode,
			SICCode:    rule.SICCode,
			Confidence: confidence,
			Evidence:   evidence[industry],
			Keywords:   rule.Keywords,
			Metadata: map[string]interface{}{
				"classification_method": "traditional_keyword_based",
			},
		}

		results = append(results, classification)
	}

	return results, nil
}

// filterHighConfidenceResults filters results based on confidence threshold
func (ic *IndustryClassifier) filterHighConfidenceResults(results []IndustryClassification) []IndustryClassification {
	filtered := make([]IndustryClassification, 0)
	for _, result := range results {
		if result.Confidence >= ic.mlConfidenceThreshold {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// calculateIndustryScore calculates the score for an industry
func (ic *IndustryClassifier) calculateIndustryScore(content string, rule IndustryClassifierRule) float64 {
	score := 0.0

	// Check for keyword matches
	for _, keyword := range rule.Keywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			score += rule.Weight
		}
	}

	// Apply confidence multiplier
	score *= rule.Confidence

	return score
}

// findEvidence finds evidence for the classification
func (ic *IndustryClassifier) findEvidence(content string, rule IndustryClassifierRule) string {
	var foundKeywords []string

	for _, keyword := range rule.Keywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			foundKeywords = append(foundKeywords, keyword)
		}
	}

	if len(foundKeywords) > 0 {
		return fmt.Sprintf("Found keywords: %s", strings.Join(foundKeywords, ", "))
	}

	return "No specific keywords found"
}

// getTopIndustries gets the top industries by score
func (ic *IndustryClassifier) getTopIndustries(scores map[string]float64, maxResults int) []string {
	// Create a slice of industries with scores
	type industryScore struct {
		industry string
		score    float64
	}

	var industryScores []industryScore
	for industry, score := range scores {
		industryScores = append(industryScores, industryScore{industry, score})
	}

	// Sort by score (descending)
	for i := 0; i < len(industryScores); i++ {
		for j := i + 1; j < len(industryScores); j++ {
			if industryScores[i].score < industryScores[j].score {
				industryScores[i], industryScores[j] = industryScores[j], industryScores[i]
			}
		}
	}

	// Get top results
	var topIndustries []string
	for i, is := range industryScores {
		if i >= maxResults {
			break
		}
		topIndustries = append(topIndustries, is.industry)
	}

	return topIndustries
}

// calculateConfidence calculates the confidence score
func (ic *IndustryClassifier) calculateConfidence(score float64, rule IndustryClassifierRule, content string) float64 {
	confidence := score * rule.Confidence

	// Apply confidence rules
	for _, confidenceRule := range ic.confidenceRules {
		if ic.matchesCondition(content, confidenceRule.Condition) {
			confidence *= confidenceRule.Multiplier
		}
	}

	// Ensure confidence is within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// matchesCondition checks if content matches a condition
func (ic *IndustryClassifier) matchesCondition(content, condition string) bool {
	// Simple condition matching for now
	return strings.Contains(content, strings.ToLower(condition))
}

// initializeClassifiers initializes the industry classifiers
func (ic *IndustryClassifier) initializeClassifiers() {
	// Technology
	ic.classifiers["Technology"] = IndustryClassifierRule{
		Industry:   "Technology",
		Keywords:   []string{"software", "technology", "digital", "app", "platform", "saas", "cloud", "ai", "machine learning", "data", "analytics"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "511210",
		SICCode:    "7372",
	}

	// Financial Services
	ic.classifiers["Financial Services"] = IndustryClassifierRule{
		Industry:   "Financial Services",
		Keywords:   []string{"bank", "financial", "investment", "insurance", "credit", "loan", "mortgage", "trading", "wealth", "asset"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "522110",
		SICCode:    "6021",
	}

	// Healthcare
	ic.classifiers["Healthcare"] = IndustryClassifierRule{
		Industry:   "Healthcare",
		Keywords:   []string{"health", "medical", "pharmaceutical", "hospital", "clinic", "doctor", "patient", "treatment", "medicine", "therapy"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "621111",
		SICCode:    "8011",
	}

	// Manufacturing
	ic.classifiers["Manufacturing"] = IndustryClassifierRule{
		Industry:   "Manufacturing",
		Keywords:   []string{"manufacturing", "factory", "production", "industrial", "machinery", "equipment", "assembly", "supply chain"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "332996",
		SICCode:    "3499",
	}

	// Retail
	ic.classifiers["Retail"] = IndustryClassifierRule{
		Industry:   "Retail",
		Keywords:   []string{"retail", "store", "shop", "commerce", "ecommerce", "online store", "marketplace", "consumer", "product"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "454110",
		SICCode:    "5961",
	}

	// Consulting
	ic.classifiers["Consulting"] = IndustryClassifierRule{
		Industry:   "Consulting",
		Keywords:   []string{"consulting", "advisory", "strategy", "management", "business", "professional services", "expertise"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "541611",
		SICCode:    "8742",
	}

	// Real Estate
	ic.classifiers["Real Estate"] = IndustryClassifierRule{
		Industry:   "Real Estate",
		Keywords:   []string{"real estate", "property", "realty", "broker", "agent", "development", "construction", "housing"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "531210",
		SICCode:    "6531",
	}

	// Education
	ic.classifiers["Education"] = IndustryClassifierRule{
		Industry:   "Education",
		Keywords:   []string{"education", "school", "university", "college", "training", "learning", "academic", "student"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "611110",
		SICCode:    "8221",
	}

	// Transportation
	ic.classifiers["Transportation"] = IndustryClassifierRule{
		Industry:   "Transportation",
		Keywords:   []string{"transportation", "logistics", "shipping", "delivery", "freight", "trucking", "warehouse", "supply chain"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "484121",
		SICCode:    "4213",
	}

	// Energy
	ic.classifiers["Energy"] = IndustryClassifierRule{
		Industry:   "Energy",
		Keywords:   []string{"energy", "oil", "gas", "renewable", "solar", "wind", "power", "utility", "electricity"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "221111",
		SICCode:    "4911",
	}

	// Initialize confidence rules
	ic.confidenceRules = []ConfidenceRule{
		{
			Condition:     "about us",
			Multiplier:    1.2,
			MinConfidence: 0.1,
		},
		{
			Condition:     "services",
			Multiplier:    1.1,
			MinConfidence: 0.1,
		},
		{
			Condition:     "products",
			Multiplier:    1.1,
			MinConfidence: 0.1,
		},
	}
}

// initializeIndustryPatterns initializes industry-specific content patterns
func (ic *IndustryClassifier) initializeIndustryPatterns() {
	ic.industryPatterns = map[string][]string{
		"Technology": {
			"software development", "technology solutions", "digital transformation", "cloud computing",
			"artificial intelligence", "machine learning", "data analytics", "cybersecurity",
			"mobile applications", "web development", "IT consulting", "technology consulting",
		},
		"Financial Services": {
			"financial services", "banking solutions", "investment management", "wealth management",
			"insurance services", "credit services", "loan services", "financial consulting",
			"asset management", "financial planning", "trading services", "risk management",
		},
		"Healthcare": {
			"healthcare services", "medical services", "patient care", "healthcare solutions",
			"medical technology", "healthcare consulting", "pharmaceutical services", "medical devices",
			"healthcare management", "clinical services", "healthcare technology", "medical consulting",
		},
		"Manufacturing": {
			"manufacturing services", "production services", "industrial manufacturing", "custom manufacturing",
			"manufacturing solutions", "industrial equipment", "manufacturing consulting", "supply chain",
			"quality control", "manufacturing technology", "industrial automation", "manufacturing management",
		},
		"Retail": {
			"retail services", "ecommerce solutions", "online retail", "retail consulting",
			"consumer products", "retail technology", "retail management", "customer service",
			"retail solutions", "merchandising", "retail analytics", "retail operations",
		},
		"Consulting": {
			"consulting services", "business consulting", "management consulting", "strategy consulting",
			"professional services", "advisory services", "business solutions", "consulting expertise",
			"business advisory", "management services", "strategic consulting", "professional advisory",
		},
		"Real Estate": {
			"real estate services", "property management", "real estate consulting", "property development",
			"real estate solutions", "property services", "real estate advisory", "construction services",
			"property investment", "real estate technology", "property consulting", "development services",
		},
		"Education": {
			"educational services", "learning solutions", "education consulting", "training services",
			"educational technology", "academic services", "education management", "learning management",
			"educational consulting", "training solutions", "academic consulting", "education solutions",
		},
	}
}

// initializeEvidenceExtractors initializes evidence extraction functions
func (ic *IndustryClassifier) initializeEvidenceExtractors() {
	ic.evidenceExtractors = map[string]EvidenceExtractor{
		"Technology": func(content string, rule IndustryClassifierRule) []string {
			return ic.extractTechnologyEvidence(content, rule)
		},
		"Financial Services": func(content string, rule IndustryClassifierRule) []string {
			return ic.extractFinancialEvidence(content, rule)
		},
		"Healthcare": func(content string, rule IndustryClassifierRule) []string {
			return ic.extractHealthcareEvidence(content, rule)
		},
		"Manufacturing": func(content string, rule IndustryClassifierRule) []string {
			return ic.extractManufacturingEvidence(content, rule)
		},
		"Retail": func(content string, rule IndustryClassifierRule) []string {
			return ic.extractRetailEvidence(content, rule)
		},
		"Consulting": func(content string, rule IndustryClassifierRule) []string {
			return ic.extractConsultingEvidence(content, rule)
		},
		"Real Estate": func(content string, rule IndustryClassifierRule) []string {
			return ic.extractRealEstateEvidence(content, rule)
		},
		"Education": func(content string, rule IndustryClassifierRule) []string {
			return ic.extractEducationEvidence(content, rule)
		},
	}
}

// Enhanced classification methods

// ClassifyContentEnhanced performs enhanced industry classification with semantic analysis
func (ic *IndustryClassifier) ClassifyContentEnhanced(ctx context.Context, content *ScrapedContent, business string, maxResults int) ([]IndustryClassificationResult, error) {
	if maxResults == 0 {
		maxResults = 3
	}

	// Perform semantic analysis
	var semanticResult *SemanticAnalysisResult
	if ic.semanticAnalyzer != nil {
		semanticResult = ic.semanticAnalyzer.AnalyzeSemanticContent(content, business)
	}

	// Perform content quality analysis
	var contentQualityResult *EnhancedContentAnalysis
	if ic.contentQualityAnalyzer != nil {
		contentQualityResult = ic.contentQualityAnalyzer.AnalyzeContent(content, business)
	}

	// Calculate enhanced scores for each industry
	scores := make(map[string]float64)
	evidence := make(map[string][]string)
	industryPatterns := make(map[string][]string)

	normalizedContent := strings.ToLower(content.Text)

	for industry, rule := range ic.classifiers {
		// Base keyword score
		baseScore := ic.calculateIndustryScore(normalizedContent, rule)

		// Semantic analysis enhancement
		semanticScore := 0.0
		if semanticResult != nil {
			semanticScore = ic.calculateSemanticEnhancement(industry, semanticResult)
		}

		// Content quality enhancement
		qualityScore := 0.0
		if contentQualityResult != nil {
			qualityScore = ic.calculateQualityEnhancement(industry, contentQualityResult)
		}

		// Industry pattern recognition
		patternScore := ic.calculatePatternScore(industry, normalizedContent)

		// Combined score
		totalScore := baseScore + semanticScore + qualityScore + patternScore

		if totalScore > 0 {
			scores[industry] = totalScore
			evidence[industry] = ic.extractEnhancedEvidence(industry, normalizedContent, rule)
			industryPatterns[industry] = ic.extractIndustryPatterns(industry, normalizedContent)
		}
	}

	// Sort industries by score and get top results
	topIndustries := ic.getTopIndustries(scores, maxResults)

	// Create enhanced classification results
	var results []IndustryClassificationResult
	for _, industry := range topIndustries {
		rule := ic.classifiers[industry]
		confidence := ic.calculateEnhancedConfidence(scores[industry], rule, normalizedContent, semanticResult, contentQualityResult)

		result := IndustryClassificationResult{
			Industry:           industry,
			NAICSCode:          rule.NAICSCode,
			SICCode:            rule.SICCode,
			Confidence:         confidence,
			Evidence:           evidence[industry],
			Keywords:           rule.Keywords,
			SemanticScore:      ic.getSemanticScore(industry, semanticResult),
			ContentQuality:     ic.getContentQualityScore(contentQualityResult),
			IndustryPatterns:   industryPatterns[industry],
			ClassificationTime: time.Now(),
			Metadata: map[string]interface{}{
				"base_score":     scores[industry],
				"semantic_score": ic.getSemanticScore(industry, semanticResult),
				"quality_score":  ic.getContentQualityScore(contentQualityResult),
				"pattern_score":  ic.calculatePatternScore(industry, normalizedContent),
			},
		}

		results = append(results, result)
	}

	return results, nil
}

// calculateSemanticEnhancement calculates semantic enhancement for an industry
func (ic *IndustryClassifier) calculateSemanticEnhancement(industry string, semanticResult *SemanticAnalysisResult) float64 {
	if semanticResult == nil {
		return 0.0
	}

	score := 0.0

	// Industry confidence from semantic analysis
	if confidence, exists := semanticResult.IndustryConfidence[strings.ToLower(industry)]; exists {
		score += confidence * 0.4
	}

	// Industry keywords from semantic analysis
	for _, keyword := range semanticResult.IndustryKeywords {
		if ic.isIndustryKeyword(industry, keyword) {
			score += 0.1
		}
	}

	// Business description relevance
	if semanticResult.BusinessDescription != "" {
		score += 0.2
	}

	return score
}

// calculateQualityEnhancement calculates quality enhancement for an industry
func (ic *IndustryClassifier) calculateQualityEnhancement(industry string, contentQualityResult *EnhancedContentAnalysis) float64 {
	if contentQualityResult == nil || contentQualityResult.ContentQuality == nil {
		return 0.0
	}

	score := 0.0

	// Content quality contribution
	score += contentQualityResult.ContentQuality.OverallQuality * 0.3

	// Meta tag quality contribution
	if contentQualityResult.MetaTags != nil {
		score += contentQualityResult.MetaTags.Quality * 0.2
	}

	// Structured data quality contribution
	if contentQualityResult.StructuredData != nil {
		score += contentQualityResult.StructuredData.Quality * 0.2
	}

	return score
}

// calculatePatternScore calculates pattern recognition score
func (ic *IndustryClassifier) calculatePatternScore(industry string, content string) float64 {
	patterns, exists := ic.industryPatterns[industry]
	if !exists {
		return 0.0
	}

	score := 0.0
	for _, pattern := range patterns {
		if strings.Contains(content, strings.ToLower(pattern)) {
			score += 0.2
		}
	}

	return score
}

// extractEnhancedEvidence extracts enhanced evidence for classification
func (ic *IndustryClassifier) extractEnhancedEvidence(industry string, content string, rule IndustryClassifierRule) []string {
	var evidence []string

	// Use industry-specific evidence extractor if available
	if extractor, exists := ic.evidenceExtractors[industry]; exists {
		evidence = extractor(content, rule)
	} else {
		// Default evidence extraction
		evidence = ic.findEvidenceList(content, rule)
	}

	return evidence
}

// extractIndustryPatterns extracts industry-specific patterns
func (ic *IndustryClassifier) extractIndustryPatterns(industry string, content string) []string {
	patterns, exists := ic.industryPatterns[industry]
	if !exists {
		return []string{}
	}

	var foundPatterns []string
	for _, pattern := range patterns {
		if strings.Contains(content, strings.ToLower(pattern)) {
			foundPatterns = append(foundPatterns, pattern)
		}
	}

	return foundPatterns
}

// calculateEnhancedConfidence calculates enhanced confidence score
func (ic *IndustryClassifier) calculateEnhancedConfidence(score float64, rule IndustryClassifierRule, content string, semanticResult *SemanticAnalysisResult, contentQualityResult *EnhancedContentAnalysis) float64 {
	confidence := score * rule.Confidence

	// Apply confidence rules
	for _, confidenceRule := range ic.confidenceRules {
		if ic.matchesCondition(content, confidenceRule.Condition) {
			confidence *= confidenceRule.Multiplier
		}
	}

	// Semantic analysis confidence boost
	if semanticResult != nil {
		confidence += semanticResult.SemanticScore * 0.2
	}

	// Content quality confidence boost
	if contentQualityResult != nil && contentQualityResult.ContentQuality != nil {
		confidence += contentQualityResult.ContentQuality.OverallQuality * 0.1
	}

	// Ensure confidence is within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// Helper methods

func (ic *IndustryClassifier) isIndustryKeyword(industry string, keyword string) bool {
	rule, exists := ic.classifiers[industry]
	if !exists {
		return false
	}

	for _, ruleKeyword := range rule.Keywords {
		if strings.Contains(strings.ToLower(keyword), strings.ToLower(ruleKeyword)) {
			return true
		}
	}

	return false
}

func (ic *IndustryClassifier) getSemanticScore(industry string, semanticResult *SemanticAnalysisResult) float64 {
	if semanticResult == nil {
		return 0.0
	}

	if confidence, exists := semanticResult.IndustryConfidence[strings.ToLower(industry)]; exists {
		return confidence
	}

	return 0.0
}

func (ic *IndustryClassifier) getContentQualityScore(contentQualityResult *EnhancedContentAnalysis) float64 {
	if contentQualityResult == nil || contentQualityResult.ContentQuality == nil {
		return 0.0
	}

	return contentQualityResult.ContentQuality.OverallQuality
}

func (ic *IndustryClassifier) findEvidenceList(content string, rule IndustryClassifierRule) []string {
	var evidence []string

	for _, keyword := range rule.Keywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			evidence = append(evidence, fmt.Sprintf("Found keyword: %s", keyword))
		}
	}

	if len(evidence) == 0 {
		evidence = append(evidence, "No specific keywords found")
	}

	return evidence
}

// Industry-specific evidence extraction methods

func (ic *IndustryClassifier) extractTechnologyEvidence(content string, rule IndustryClassifierRule) []string {
	var evidence []string
	contentLower := strings.ToLower(content)

	techPatterns := []string{"software", "technology", "digital", "platform", "solution", "development", "system"}
	for _, pattern := range techPatterns {
		if strings.Contains(contentLower, pattern) {
			evidence = append(evidence, fmt.Sprintf("Technology pattern: %s", pattern))
		}
	}

	return evidence
}

func (ic *IndustryClassifier) extractFinancialEvidence(content string, rule IndustryClassifierRule) []string {
	var evidence []string
	contentLower := strings.ToLower(content)

	financialPatterns := []string{"financial", "banking", "investment", "insurance", "credit", "loan", "wealth"}
	for _, pattern := range financialPatterns {
		if strings.Contains(contentLower, pattern) {
			evidence = append(evidence, fmt.Sprintf("Financial pattern: %s", pattern))
		}
	}

	return evidence
}

func (ic *IndustryClassifier) extractHealthcareEvidence(content string, rule IndustryClassifierRule) []string {
	var evidence []string
	contentLower := strings.ToLower(content)

	healthcarePatterns := []string{"health", "medical", "patient", "treatment", "care", "clinical", "therapeutic"}
	for _, pattern := range healthcarePatterns {
		if strings.Contains(contentLower, pattern) {
			evidence = append(evidence, fmt.Sprintf("Healthcare pattern: %s", pattern))
		}
	}

	return evidence
}

func (ic *IndustryClassifier) extractManufacturingEvidence(content string, rule IndustryClassifierRule) []string {
	var evidence []string
	contentLower := strings.ToLower(content)

	manufacturingPatterns := []string{"manufacturing", "production", "factory", "industrial", "equipment", "assembly"}
	for _, pattern := range manufacturingPatterns {
		if strings.Contains(contentLower, pattern) {
			evidence = append(evidence, fmt.Sprintf("Manufacturing pattern: %s", pattern))
		}
	}

	return evidence
}

func (ic *IndustryClassifier) extractRetailEvidence(content string, rule IndustryClassifierRule) []string {
	var evidence []string
	contentLower := strings.ToLower(content)

	retailPatterns := []string{"retail", "store", "shop", "commerce", "consumer", "product", "merchandise"}
	for _, pattern := range retailPatterns {
		if strings.Contains(contentLower, pattern) {
			evidence = append(evidence, fmt.Sprintf("Retail pattern: %s", pattern))
		}
	}

	return evidence
}

func (ic *IndustryClassifier) extractConsultingEvidence(content string, rule IndustryClassifierRule) []string {
	var evidence []string
	contentLower := strings.ToLower(content)

	consultingPatterns := []string{"consulting", "advisory", "strategy", "management", "professional", "expertise"}
	for _, pattern := range consultingPatterns {
		if strings.Contains(contentLower, pattern) {
			evidence = append(evidence, fmt.Sprintf("Consulting pattern: %s", pattern))
		}
	}

	return evidence
}

func (ic *IndustryClassifier) extractRealEstateEvidence(content string, rule IndustryClassifierRule) []string {
	var evidence []string
	contentLower := strings.ToLower(content)

	realEstatePatterns := []string{"real estate", "property", "development", "construction", "housing", "broker"}
	for _, pattern := range realEstatePatterns {
		if strings.Contains(contentLower, pattern) {
			evidence = append(evidence, fmt.Sprintf("Real Estate pattern: %s", pattern))
		}
	}

	return evidence
}

func (ic *IndustryClassifier) extractEducationEvidence(content string, rule IndustryClassifierRule) []string {
	var evidence []string
	contentLower := strings.ToLower(content)

	educationPatterns := []string{"education", "learning", "training", "academic", "student", "school", "university"}
	for _, pattern := range educationPatterns {
		if strings.Contains(contentLower, pattern) {
			evidence = append(evidence, fmt.Sprintf("Education pattern: %s", pattern))
		}
	}

	return evidence
}
