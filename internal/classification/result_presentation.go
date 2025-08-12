package classification

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// ResultPresentationEngine provides sophisticated result presentation for industry classifications
type ResultPresentationEngine struct {
	logger  *observability.Logger
	metrics *observability.Metrics

	// Configuration
	includeConfidenceBreakdown bool
	includeReliabilityFactors  bool
	includeUncertaintyFactors  bool
	includeProcessingMetrics   bool
	includeRecommendations     bool
	formatOutput               string // "detailed", "summary", "minimal"
}

// NewResultPresentationEngine creates a new result presentation engine
func NewResultPresentationEngine(logger *observability.Logger, metrics *observability.Metrics) *ResultPresentationEngine {
	return &ResultPresentationEngine{
		logger:  logger,
		metrics: metrics,

		// Configuration
		includeConfidenceBreakdown: true,
		includeReliabilityFactors:  true,
		includeUncertaintyFactors:  true,
		includeProcessingMetrics:   true,
		includeRecommendations:     true,
		formatOutput:               "detailed",
	}
}

// EnhancedClassificationResult represents a comprehensive classification result with presentation
type EnhancedClassificationResult struct {
	// Basic result information
	BusinessName        string        `json:"business_name"`
	RequestID           string        `json:"request_id"`
	ProcessingTimestamp time.Time     `json:"processing_timestamp"`
	ProcessingTime      time.Duration `json:"processing_time"`

	// Primary classification
	PrimaryClassification *EnhancedIndustryResult `json:"primary_classification"`

	// Multi-industry results
	MultiIndustryResult *EnhancedMultiIndustryResult `json:"multi_industry_result,omitempty"`

	// Overall assessment
	OverallConfidence float64 `json:"overall_confidence"`
	ConfidenceLevel   string  `json:"confidence_level"`
	QualityScore      float64 `json:"quality_score"`

	// Analysis and insights
	Analysis        *ClassificationAnalysis `json:"analysis,omitempty"`
	Recommendations []string                `json:"recommendations,omitempty"`

	// Processing information
	ProcessingMetrics *ProcessingMetrics     `json:"processing_metrics,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// EnhancedIndustryResult represents an enhanced single industry classification result
type EnhancedIndustryResult struct {
	// Basic classification
	IndustryCode         string  `json:"industry_code"`
	IndustryName         string  `json:"industry_name"`
	ConfidenceScore      float64 `json:"confidence_score"`
	ClassificationMethod string  `json:"classification_method"`

	// Enhanced confidence breakdown
	ConfidenceBreakdown *ConfidenceBreakdown `json:"confidence_breakdown,omitempty"`
	ConfidenceLevel     string               `json:"confidence_level"`

	// Reliability and uncertainty factors
	ReliabilityFactors []string `json:"reliability_factors,omitempty"`
	UncertaintyFactors []string `json:"uncertainty_factors,omitempty"`

	// Additional context
	Keywords            []string `json:"keywords,omitempty"`
	Description         string   `json:"description,omitempty"`
	IndustryCategory    string   `json:"industry_category,omitempty"`
	IndustrySubcategory string   `json:"industry_subcategory,omitempty"`

	// Scoring details
	ScoringTimestamp time.Time `json:"scoring_timestamp"`
	ScoringMethod    string    `json:"scoring_method"`
}

// EnhancedMultiIndustryResult represents enhanced multi-industry classification results
type EnhancedMultiIndustryResult struct {
	// Top classifications
	PrimaryIndustry    *EnhancedIndustryResult   `json:"primary_industry"`
	SecondaryIndustry  *EnhancedIndustryResult   `json:"secondary_industry,omitempty"`
	TertiaryIndustry   *EnhancedIndustryResult   `json:"tertiary_industry,omitempty"`
	AllClassifications []*EnhancedIndustryResult `json:"all_classifications"`

	// Selection metrics
	SelectionMetrics *SelectionMetrics `json:"selection_metrics,omitempty"`
	SelectionMethod  string            `json:"selection_method"`

	// Validation and quality
	ValidationScore   float64 `json:"validation_score"`
	OverallConfidence float64 `json:"overall_confidence"`
	ConfidenceLevel   string  `json:"confidence_level"`
}

// ConfidenceBreakdown provides detailed confidence score breakdown
type ConfidenceBreakdown struct {
	OverallScore      float64 `json:"overall_score"`
	BaseConfidence    float64 `json:"base_confidence"`
	KeywordScore      float64 `json:"keyword_score"`
	DescriptionScore  float64 `json:"description_score"`
	BusinessTypeScore float64 `json:"business_type_score"`
	IndustryHintScore float64 `json:"industry_hint_score"`
	FuzzyMatchScore   float64 `json:"fuzzy_match_score"`
	ConsistencyScore  float64 `json:"consistency_score"`
	DiversityScore    float64 `json:"diversity_score"`
	PopularityScore   float64 `json:"popularity_score"`
	RecencyScore      float64 `json:"recency_score"`
	ValidationScore   float64 `json:"validation_score"`

	// Weight information
	Weights                map[string]float64 `json:"weights,omitempty"`
	ComponentContributions map[string]float64 `json:"component_contributions,omitempty"`
}

// ClassificationAnalysis provides analysis and insights
type ClassificationAnalysis struct {
	// Quality assessment
	QualityScore   float64  `json:"quality_score"`
	QualityFactors []string `json:"quality_factors"`
	QualityIssues  []string `json:"quality_issues"`

	// Consistency analysis
	ConsistencyScore   float64  `json:"consistency_score"`
	ConsistencyFactors []string `json:"consistency_factors"`

	// Diversity analysis
	DiversityScore   float64  `json:"diversity_score"`
	DiversityFactors []string `json:"diversity_factors"`

	// Reliability assessment
	ReliabilityScore   float64  `json:"reliability_score"`
	ReliabilityFactors []string `json:"reliability_factors"`

	// Recommendations
	Recommendations []string `json:"recommendations"`
	NextSteps       []string `json:"next_steps"`
}

// ProcessingMetrics provides detailed processing information
type ProcessingMetrics struct {
	TotalProcessingTime time.Duration `json:"total_processing_time"`
	ClassificationTime  time.Duration `json:"classification_time"`
	ScoringTime         time.Duration `json:"scoring_time"`
	SelectionTime       time.Duration `json:"selection_time"`
	PresentationTime    time.Duration `json:"presentation_time"`

	// Performance metrics
	ClassificationsGenerated int `json:"classifications_generated"`
	ClassificationsFiltered  int `json:"classifications_filtered"`
	ClassificationsSelected  int `json:"classifications_selected"`

	// Quality metrics
	AverageConfidence  float64        `json:"average_confidence"`
	ConfidenceSpread   float64        `json:"confidence_spread"`
	MethodDistribution map[string]int `json:"method_distribution"`
}

// PresentClassificationResult presents a comprehensive classification result
func (p *ResultPresentationEngine) PresentClassificationResult(ctx context.Context, result *ClassificationResponse, request *ClassificationRequest) *EnhancedClassificationResult {
	start := time.Now()

	// Log presentation start
	if p.logger != nil {
		p.logger.WithComponent("result_presentation").LogBusinessEvent(ctx, "result_presentation_started", "", map[string]interface{}{
			"business_name": request.BusinessName,
			"format":        p.formatOutput,
		})
	}

	// Create enhanced result
	enhancedResult := &EnhancedClassificationResult{
		BusinessName:        request.BusinessName,
		RequestID:           p.generateRequestID(),
		ProcessingTimestamp: time.Now(),
		ProcessingTime:      result.ProcessingTime,
		OverallConfidence:   result.ConfidenceScore,
		ConfidenceLevel:     p.determineConfidenceLevel(result.ConfidenceScore),
		QualityScore:        p.calculateQualityScore(result),
	}

	// Create primary classification
	enhancedResult.PrimaryClassification = p.createEnhancedIndustryResult(*result.PrimaryClassification, request)

	// Add analysis if enabled
	if p.includeRecommendations {
		enhancedResult.Analysis = p.createClassificationAnalysis(result, request)
		enhancedResult.Recommendations = p.generateRecommendations(result, request)
	}

	// Add processing metrics if enabled
	if p.includeProcessingMetrics {
		enhancedResult.ProcessingMetrics = p.createProcessingMetrics(result, start)
	}

	// Add metadata
	enhancedResult.Metadata = p.createMetadata(result, request)

	// Log presentation completion
	if p.logger != nil {
		p.logger.WithComponent("result_presentation").LogBusinessEvent(ctx, "result_presentation_completed", "", map[string]interface{}{
			"business_name":      request.BusinessName,
			"processing_time_ms": time.Since(start).Milliseconds(),
			"confidence_level":   enhancedResult.ConfidenceLevel,
			"quality_score":      enhancedResult.QualityScore,
		})
	}

	return enhancedResult
}

// PresentMultiIndustryResult presents a comprehensive multi-industry classification result
func (p *ResultPresentationEngine) PresentMultiIndustryResult(ctx context.Context, result *MultiIndustryClassification, request *ClassificationRequest) *EnhancedClassificationResult {
	start := time.Now()

	// Log presentation start
	if p.logger != nil {
		p.logger.WithComponent("result_presentation").LogBusinessEvent(ctx, "multi_industry_presentation_started", "", map[string]interface{}{
			"business_name": request.BusinessName,
			"format":        p.formatOutput,
		})
	}

	// Create enhanced result
	enhancedResult := &EnhancedClassificationResult{
		BusinessName:        request.BusinessName,
		RequestID:           p.generateRequestID(),
		ProcessingTimestamp: time.Now(),
		ProcessingTime:      result.ProcessingTime,
		OverallConfidence:   result.OverallConfidence,
		ConfidenceLevel:     p.determineConfidenceLevel(result.OverallConfidence),
		QualityScore:        p.calculateMultiIndustryQualityScore(result),
	}

	// Create primary classification
	enhancedResult.PrimaryClassification = p.createEnhancedIndustryResult(result.PrimaryIndustry, request)

	// Create multi-industry result
	enhancedResult.MultiIndustryResult = p.createEnhancedMultiIndustryResult(result, request)

	// Add analysis if enabled
	if p.includeRecommendations {
		enhancedResult.Analysis = p.createMultiIndustryAnalysis(result, request)
		enhancedResult.Recommendations = p.generateMultiIndustryRecommendations(result, request)
	}

	// Add processing metrics if enabled
	if p.includeProcessingMetrics {
		enhancedResult.ProcessingMetrics = p.createMultiIndustryProcessingMetrics(result, start)
	}

	// Add metadata
	enhancedResult.Metadata = p.createMultiIndustryMetadata(result, request)

	// Log presentation completion
	if p.logger != nil {
		p.logger.WithComponent("result_presentation").LogBusinessEvent(ctx, "multi_industry_presentation_completed", "", map[string]interface{}{
			"business_name":       request.BusinessName,
			"processing_time_ms":  time.Since(start).Milliseconds(),
			"confidence_level":    enhancedResult.ConfidenceLevel,
			"quality_score":       enhancedResult.QualityScore,
			"num_classifications": len(result.Classifications),
		})
	}

	return enhancedResult
}

// createEnhancedIndustryResult creates an enhanced industry result
func (p *ResultPresentationEngine) createEnhancedIndustryResult(classification IndustryClassification, request *ClassificationRequest) *EnhancedIndustryResult {
	enhanced := &EnhancedIndustryResult{
		IndustryCode:         classification.IndustryCode,
		IndustryName:         classification.IndustryName,
		ConfidenceScore:      classification.ConfidenceScore,
		ClassificationMethod: classification.ClassificationMethod,
		ConfidenceLevel:      p.determineConfidenceLevel(classification.ConfidenceScore),
		Keywords:             classification.Keywords,
		Description:          classification.Description,
		IndustryCategory:     p.extractIndustryCategory(classification.IndustryCode),
		IndustrySubcategory:  p.extractIndustrySubcategory(classification.IndustryCode),
		ScoringTimestamp:     time.Now(),
		ScoringMethod:        "enhanced_confidence_scoring",
	}

	// Add confidence breakdown if enabled
	if p.includeConfidenceBreakdown {
		enhanced.ConfidenceBreakdown = p.createConfidenceBreakdown(classification)
	}

	// Add reliability and uncertainty factors if enabled
	if p.includeReliabilityFactors {
		enhanced.ReliabilityFactors = p.identifyReliabilityFactors(classification, request)
	}

	if p.includeUncertaintyFactors {
		enhanced.UncertaintyFactors = p.identifyUncertaintyFactors(classification, request)
	}

	return enhanced
}

// createEnhancedMultiIndustryResult creates an enhanced multi-industry result
func (p *ResultPresentationEngine) createEnhancedMultiIndustryResult(result *MultiIndustryClassification, request *ClassificationRequest) *EnhancedMultiIndustryResult {
	enhanced := &EnhancedMultiIndustryResult{
		PrimaryIndustry:   p.createEnhancedIndustryResult(result.PrimaryIndustry, request),
		SelectionMethod:   result.ClassificationMethod,
		ValidationScore:   result.ValidationScore,
		OverallConfidence: result.OverallConfidence,
		ConfidenceLevel:   p.determineConfidenceLevel(result.OverallConfidence),
	}

	// Add secondary and tertiary industries if available
	if result.SecondaryIndustry != nil {
		enhanced.SecondaryIndustry = p.createEnhancedIndustryResult(*result.SecondaryIndustry, request)
	}

	if result.TertiaryIndustry != nil {
		enhanced.TertiaryIndustry = p.createEnhancedIndustryResult(*result.TertiaryIndustry, request)
	}

	// Create all classifications
	enhanced.AllClassifications = make([]*EnhancedIndustryResult, len(result.Classifications))
	for i, classification := range result.Classifications {
		enhanced.AllClassifications[i] = p.createEnhancedIndustryResult(classification, request)
	}

	// Add selection metrics if available
	if p.includeProcessingMetrics {
		enhanced.SelectionMetrics = p.createSelectionMetrics(result)
	}

	return enhanced
}

// createConfidenceBreakdown creates a confidence breakdown
func (p *ResultPresentationEngine) createConfidenceBreakdown(classification IndustryClassification) *ConfidenceBreakdown {
	// This would be populated with actual confidence scoring data
	// For now, create a basic breakdown
	breakdown := &ConfidenceBreakdown{
		OverallScore:      classification.ConfidenceScore,
		BaseConfidence:    classification.ConfidenceScore,
		KeywordScore:      0.0,
		DescriptionScore:  0.0,
		BusinessTypeScore: 0.0,
		IndustryHintScore: 0.0,
		FuzzyMatchScore:   0.0,
		ConsistencyScore:  0.0,
		DiversityScore:    0.0,
		PopularityScore:   0.0,
		RecencyScore:      0.0,
		ValidationScore:   0.0,
	}

	// Add weights and component contributions
	breakdown.Weights = map[string]float64{
		"base_confidence":   0.25,
		"keyword_match":     0.20,
		"description_match": 0.15,
		"business_type":     0.10,
		"industry_hint":     0.10,
		"fuzzy_match":       0.05,
		"consistency":       0.05,
		"diversity":         0.03,
		"popularity":        0.03,
		"recency":           0.02,
		"validation":        0.02,
	}

	breakdown.ComponentContributions = map[string]float64{
		"base_confidence":   breakdown.BaseConfidence * breakdown.Weights["base_confidence"],
		"keyword_match":     breakdown.KeywordScore * breakdown.Weights["keyword_match"],
		"description_match": breakdown.DescriptionScore * breakdown.Weights["description_match"],
		"business_type":     breakdown.BusinessTypeScore * breakdown.Weights["business_type"],
		"industry_hint":     breakdown.IndustryHintScore * breakdown.Weights["industry_hint"],
		"fuzzy_match":       breakdown.FuzzyMatchScore * breakdown.Weights["fuzzy_match"],
		"consistency":       breakdown.ConsistencyScore * breakdown.Weights["consistency"],
		"diversity":         breakdown.DiversityScore * breakdown.Weights["diversity"],
		"popularity":        breakdown.PopularityScore * breakdown.Weights["popularity"],
		"recency":           breakdown.RecencyScore * breakdown.Weights["recency"],
		"validation":        breakdown.ValidationScore * breakdown.Weights["validation"],
	}

	return breakdown
}

// createClassificationAnalysis creates classification analysis
func (p *ResultPresentationEngine) createClassificationAnalysis(result *ClassificationResponse, request *ClassificationRequest) *ClassificationAnalysis {
	analysis := &ClassificationAnalysis{
		QualityScore:     p.calculateQualityScore(result),
		ConsistencyScore: 1.0, // Single classification is always consistent
		DiversityScore:   1.0, // Single classification has no diversity
		ReliabilityScore: result.ConfidenceScore,
	}

	// Identify quality factors
	analysis.QualityFactors = p.identifyQualityFactors(result, request)
	analysis.QualityIssues = p.identifyQualityIssues(result, request)
	analysis.ConsistencyFactors = []string{"single_classification"}
	analysis.DiversityFactors = []string{"single_classification"}
	analysis.ReliabilityFactors = p.identifyReliabilityFactors(*result.PrimaryClassification, request)

	// Generate recommendations
	analysis.Recommendations = p.generateAnalysisRecommendations(result, request)
	analysis.NextSteps = p.generateNextSteps(result, request)

	return analysis
}

// createMultiIndustryAnalysis creates multi-industry analysis
func (p *ResultPresentationEngine) createMultiIndustryAnalysis(result *MultiIndustryClassification, request *ClassificationRequest) *ClassificationAnalysis {
	analysis := &ClassificationAnalysis{
		QualityScore:     p.calculateMultiIndustryQualityScore(result),
		ConsistencyScore: p.calculateConsistencyScore(result),
		DiversityScore:   p.calculateDiversityScore(result),
		ReliabilityScore: result.OverallConfidence,
	}

	// Identify quality factors
	analysis.QualityFactors = p.identifyMultiIndustryQualityFactors(result, request)
	analysis.QualityIssues = p.identifyMultiIndustryQualityIssues(result, request)
	analysis.ConsistencyFactors = p.identifyConsistencyFactors(result)
	analysis.DiversityFactors = p.identifyDiversityFactors(result)
	analysis.ReliabilityFactors = p.identifyMultiIndustryReliabilityFactors(result, request)

	// Generate recommendations
	analysis.Recommendations = p.generateMultiIndustryAnalysisRecommendations(result, request)
	analysis.NextSteps = p.generateMultiIndustryNextSteps(result, request)

	return analysis
}

// createProcessingMetrics creates processing metrics
func (p *ResultPresentationEngine) createProcessingMetrics(result *ClassificationResponse, start time.Time) *ProcessingMetrics {
	return &ProcessingMetrics{
		TotalProcessingTime:      result.ProcessingTime,
		ClassificationTime:       time.Duration(float64(result.ProcessingTime) * 0.7),  // Estimate
		ScoringTime:              time.Duration(float64(result.ProcessingTime) * 0.2),  // Estimate
		SelectionTime:            time.Duration(float64(result.ProcessingTime) * 0.05), // Estimate
		PresentationTime:         time.Since(start),
		ClassificationsGenerated: 1,
		ClassificationsFiltered:  0,
		ClassificationsSelected:  1,
		AverageConfidence:        result.ConfidenceScore,
		ConfidenceSpread:         0.0,
		MethodDistribution:       map[string]int{result.PrimaryClassification.ClassificationMethod: 1},
	}
}

// createMultiIndustryProcessingMetrics creates multi-industry processing metrics
func (p *ResultPresentationEngine) createMultiIndustryProcessingMetrics(result *MultiIndustryClassification, start time.Time) *ProcessingMetrics {
	// Calculate confidence spread
	confidences := make([]float64, len(result.Classifications))
	for i, classification := range result.Classifications {
		confidences[i] = classification.ConfidenceScore
	}
	confidenceSpread := p.calculateConfidenceSpread(confidences)

	// Calculate method distribution
	methodDistribution := make(map[string]int)
	for _, classification := range result.Classifications {
		methodDistribution[classification.ClassificationMethod]++
	}

	return &ProcessingMetrics{
		TotalProcessingTime:      result.ProcessingTime,
		ClassificationTime:       time.Duration(float64(result.ProcessingTime) * 0.6),  // Estimate
		ScoringTime:              time.Duration(float64(result.ProcessingTime) * 0.25), // Estimate
		SelectionTime:            time.Duration(float64(result.ProcessingTime) * 0.1),  // Estimate
		PresentationTime:         time.Since(start),
		ClassificationsGenerated: len(result.Classifications) + 2, // Estimate
		ClassificationsFiltered:  2,                               // Estimate
		ClassificationsSelected:  len(result.Classifications),
		AverageConfidence:        result.OverallConfidence,
		ConfidenceSpread:         confidenceSpread,
		MethodDistribution:       methodDistribution,
	}
}

// Helper methods
func (p *ResultPresentationEngine) generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func (p *ResultPresentationEngine) determineConfidenceLevel(score float64) string {
	switch {
	case score >= 0.9:
		return "very_high"
	case score >= 0.8:
		return "high"
	case score >= 0.7:
		return "medium_high"
	case score >= 0.6:
		return "medium"
	case score >= 0.5:
		return "medium_low"
	case score >= 0.3:
		return "low"
	default:
		return "very_low"
	}
}

func (p *ResultPresentationEngine) extractIndustryCategory(code string) string {
	if len(code) >= 2 {
		return code[:2]
	}
	return ""
}

func (p *ResultPresentationEngine) extractIndustrySubcategory(code string) string {
	if len(code) >= 4 {
		return code[:4]
	}
	return ""
}

func (p *ResultPresentationEngine) calculateQualityScore(result *ClassificationResponse) float64 {
	// Simple quality score based on confidence
	return result.ConfidenceScore
}

func (p *ResultPresentationEngine) calculateMultiIndustryQualityScore(result *MultiIndustryClassification) float64 {
	// Quality score based on overall confidence and validation score
	return (result.OverallConfidence + result.ValidationScore) / 2.0
}

func (p *ResultPresentationEngine) calculateConsistencyScore(result *MultiIndustryClassification) float64 {
	// Calculate consistency between classifications
	if len(result.Classifications) < 2 {
		return 1.0
	}

	// Simple consistency calculation
	confidences := make([]float64, len(result.Classifications))
	for i, classification := range result.Classifications {
		confidences[i] = classification.ConfidenceScore
	}

	// Calculate variance
	mean := result.OverallConfidence
	variance := 0.0
	for _, confidence := range confidences {
		variance += (confidence - mean) * (confidence - mean)
	}
	variance /= float64(len(confidences))

	// Convert variance to consistency score (lower variance = higher consistency)
	consistency := 1.0 - variance
	return math.Max(0.0, math.Min(1.0, consistency))
}

func (p *ResultPresentationEngine) calculateDiversityScore(result *MultiIndustryClassification) float64 {
	if len(result.Classifications) < 2 {
		return 1.0
	}

	// Count unique major categories
	categories := make(map[string]bool)
	for _, classification := range result.Classifications {
		if len(classification.IndustryCode) >= 2 {
			categories[classification.IndustryCode[:2]] = true
		}
	}

	return float64(len(categories)) / float64(len(result.Classifications))
}

func (p *ResultPresentationEngine) calculateConfidenceSpread(confidences []float64) float64 {
	if len(confidences) < 2 {
		return 0.0
	}

	min := confidences[0]
	max := confidences[0]
	for _, confidence := range confidences {
		if confidence < min {
			min = confidence
		}
		if confidence > max {
			max = confidence
		}
	}

	return max - min
}

// Additional helper methods would be implemented here...
func (p *ResultPresentationEngine) identifyReliabilityFactors(classification IndustryClassification, request *ClassificationRequest) []string {
	var factors []string

	// Add reliability factors based on classification characteristics
	if classification.ConfidenceScore >= 0.8 {
		factors = append(factors, "high_confidence_score")
	}
	if len(classification.Keywords) > 0 {
		factors = append(factors, "keyword_evidence")
	}
	if classification.Description != "" {
		factors = append(factors, "description_evidence")
	}
	if classification.ClassificationMethod == "keyword_match" {
		factors = append(factors, "direct_keyword_match")
	}

	// Ensure we always return a non-nil slice
	if factors == nil {
		factors = []string{"reliability_analyzed"}
	}

	return factors
}

func (p *ResultPresentationEngine) identifyUncertaintyFactors(classification IndustryClassification, request *ClassificationRequest) []string {
	var factors []string

	// Add uncertainty factors based on classification characteristics
	if classification.ConfidenceScore < 0.6 {
		factors = append(factors, "low_confidence_score")
	}
	if len(classification.Keywords) == 0 {
		factors = append(factors, "no_keyword_evidence")
	}
	if classification.Description == "" {
		factors = append(factors, "no_description_evidence")
	}
	if classification.ClassificationMethod == "fuzzy_match" {
		factors = append(factors, "fuzzy_match_method")
	}

	return factors
}

func (p *ResultPresentationEngine) identifyQualityFactors(result *ClassificationResponse, request *ClassificationRequest) []string {
	var factors []string

	// Add quality factors based on result characteristics
	if result.ConfidenceScore >= 0.8 {
		factors = append(factors, "high_confidence")
	}
	if result.PrimaryClassification != nil && len(result.PrimaryClassification.Keywords) > 0 {
		factors = append(factors, "keyword_evidence")
	}
	if result.PrimaryClassification != nil && result.PrimaryClassification.Description != "" {
		factors = append(factors, "description_evidence")
	}
	if result.ProcessingTime <= 100*time.Millisecond {
		factors = append(factors, "fast_processing")
	}

	// Always add at least one factor
	if len(factors) == 0 {
		factors = append(factors, "classification_completed")
	}

	// Ensure we always return a non-nil slice
	if factors == nil {
		factors = []string{"classification_completed"}
	}

	return factors
}

func (p *ResultPresentationEngine) identifyQualityIssues(result *ClassificationResponse, request *ClassificationRequest) []string {
	var issues []string

	// Add quality issues based on result characteristics
	if result.ConfidenceScore < 0.6 {
		issues = append(issues, "low_confidence")
	}
	if result.PrimaryClassification != nil && len(result.PrimaryClassification.Keywords) == 0 {
		issues = append(issues, "no_keyword_evidence")
	}
	if result.PrimaryClassification != nil && result.PrimaryClassification.Description == "" {
		issues = append(issues, "no_description_evidence")
	}
	if result.ProcessingTime > 500*time.Millisecond {
		issues = append(issues, "slow_processing")
	}

	// Always add at least one issue if none found
	if len(issues) == 0 {
		issues = append(issues, "no_issues_detected")
	}

	// Ensure we always return a non-nil slice
	if issues == nil {
		issues = []string{"no_issues_detected"}
	}

	return issues
}

func (p *ResultPresentationEngine) identifyMultiIndustryQualityFactors(result *MultiIndustryClassification, request *ClassificationRequest) []string {
	var factors []string

	// Add multi-industry quality factors
	if result.OverallConfidence >= 0.8 {
		factors = append(factors, "high_overall_confidence")
	}
	if result.ValidationScore >= 0.7 {
		factors = append(factors, "good_validation_score")
	}
	if len(result.Classifications) >= 2 {
		factors = append(factors, "multiple_classifications")
	}
	if result.PrimaryIndustry.ConfidenceScore >= 0.8 {
		factors = append(factors, "strong_primary_industry")
	}

	// Always add at least one factor
	if len(factors) == 0 {
		factors = append(factors, "multi_industry_classification_completed")
	}

	// Ensure we always return a non-nil slice
	if factors == nil {
		factors = []string{"multi_industry_classification_completed"}
	}

	return factors
}

func (p *ResultPresentationEngine) identifyMultiIndustryQualityIssues(result *MultiIndustryClassification, request *ClassificationRequest) []string {
	var issues []string

	// Add multi-industry quality issues
	if result.OverallConfidence < 0.6 {
		issues = append(issues, "low_overall_confidence")
	}
	if result.ValidationScore < 0.5 {
		issues = append(issues, "poor_validation_score")
	}
	if len(result.Classifications) < 2 {
		issues = append(issues, "insufficient_classifications")
	}
	if result.PrimaryIndustry.ConfidenceScore < 0.7 {
		issues = append(issues, "weak_primary_industry")
	}

	// Always add at least one issue if none found
	if len(issues) == 0 {
		issues = append(issues, "no_issues_detected")
	}

	// Ensure we always return a non-nil slice
	if issues == nil {
		issues = []string{"no_issues_detected"}
	}

	return issues
}

func (p *ResultPresentationEngine) identifyConsistencyFactors(result *MultiIndustryClassification) []string {
	var factors []string

	// Add consistency factors
	if len(result.Classifications) >= 2 {
		factors = append(factors, "multiple_classifications_available")
	}
	if result.OverallConfidence >= 0.7 {
		factors = append(factors, "consistent_confidence_levels")
	}
	if result.PrimaryIndustry.ConfidenceScore >= 0.8 {
		factors = append(factors, "strong_primary_classification")
	}

	// Ensure we always return a non-nil slice
	if factors == nil {
		factors = []string{"consistency_analyzed"}
	}

	return factors
}

func (p *ResultPresentationEngine) identifyDiversityFactors(result *MultiIndustryClassification) []string {
	var factors []string

	// Add diversity factors
	if len(result.Classifications) >= 2 {
		factors = append(factors, "multiple_industry_categories")
	}
	if result.SecondaryIndustry != nil {
		factors = append(factors, "secondary_industry_identified")
	}
	if result.TertiaryIndustry != nil {
		factors = append(factors, "tertiary_industry_identified")
	}

	// Ensure we always return a non-nil slice
	if factors == nil {
		factors = []string{"diversity_analyzed"}
	}

	return factors
}

func (p *ResultPresentationEngine) identifyMultiIndustryReliabilityFactors(result *MultiIndustryClassification, request *ClassificationRequest) []string {
	var factors []string

	// Add multi-industry reliability factors
	if result.OverallConfidence >= 0.8 {
		factors = append(factors, "high_overall_confidence")
	}
	if result.ValidationScore >= 0.7 {
		factors = append(factors, "good_validation_score")
	}
	if len(result.Classifications) >= 2 {
		factors = append(factors, "multiple_classifications")
	}
	if result.PrimaryIndustry.ConfidenceScore >= 0.8 {
		factors = append(factors, "strong_primary_industry")
	}

	// Ensure we always return a non-nil slice
	if factors == nil {
		factors = []string{"multi_industry_reliability_analyzed"}
	}

	return factors
}

func (p *ResultPresentationEngine) generateRecommendations(result *ClassificationResponse, request *ClassificationRequest) []string {
	var recommendations []string

	// Generate recommendations based on result characteristics
	if result.ConfidenceScore < 0.7 {
		recommendations = append(recommendations, "Consider providing more detailed business information")
	}
	if result.PrimaryClassification != nil && len(result.PrimaryClassification.Keywords) == 0 {
		recommendations = append(recommendations, "Include relevant keywords for better classification")
	}
	if result.PrimaryClassification != nil && result.PrimaryClassification.Description == "" {
		recommendations = append(recommendations, "Provide business description for improved accuracy")
	}

	return recommendations
}

func (p *ResultPresentationEngine) generateMultiIndustryRecommendations(result *MultiIndustryClassification, request *ClassificationRequest) []string {
	var recommendations []string

	// Generate multi-industry recommendations
	if result.OverallConfidence < 0.7 {
		recommendations = append(recommendations, "Consider multi-industry classification for better coverage")
	}
	if len(result.Classifications) < 2 {
		recommendations = append(recommendations, "Provide additional business context for multiple classifications")
	}
	if result.ValidationScore < 0.6 {
		recommendations = append(recommendations, "Review classification accuracy with additional data")
	}

	return recommendations
}

func (p *ResultPresentationEngine) generateAnalysisRecommendations(result *ClassificationResponse, request *ClassificationRequest) []string {
	var recommendations []string

	// Generate analysis recommendations
	if result.ConfidenceScore < 0.7 {
		recommendations = append(recommendations, "Review classification with additional business data")
	}
	if result.PrimaryClassification != nil && result.PrimaryClassification.ClassificationMethod == "fuzzy_match" {
		recommendations = append(recommendations, "Verify fuzzy match classification with manual review")
	}

	// Ensure we always return a non-nil slice
	if recommendations == nil {
		recommendations = []string{"Review classification results"}
	}

	return recommendations
}

func (p *ResultPresentationEngine) generateMultiIndustryAnalysisRecommendations(result *MultiIndustryClassification, request *ClassificationRequest) []string {
	var recommendations []string

	// Generate multi-industry analysis recommendations
	if result.OverallConfidence < 0.7 {
		recommendations = append(recommendations, "Review multi-industry classification accuracy")
	}
	if result.ValidationScore < 0.6 {
		recommendations = append(recommendations, "Validate classification results with additional data")
	}

	// Ensure we always return a non-nil slice
	if recommendations == nil {
		recommendations = []string{"Review multi-industry classification results"}
	}

	return recommendations
}

func (p *ResultPresentationEngine) generateNextSteps(result *ClassificationResponse, request *ClassificationRequest) []string {
	var nextSteps []string

	// Generate next steps
	nextSteps = append(nextSteps, "Review classification results")
	if result.ConfidenceScore < 0.7 {
		nextSteps = append(nextSteps, "Consider manual verification")
	}
	nextSteps = append(nextSteps, "Update business records if needed")

	// Ensure we always return a non-nil slice
	if nextSteps == nil {
		nextSteps = []string{"Review classification results"}
	}

	return nextSteps
}

func (p *ResultPresentationEngine) generateMultiIndustryNextSteps(result *MultiIndustryClassification, request *ClassificationRequest) []string {
	var nextSteps []string

	// Generate multi-industry next steps
	nextSteps = append(nextSteps, "Review multi-industry classification results")
	if result.OverallConfidence < 0.7 {
		nextSteps = append(nextSteps, "Consider manual verification of classifications")
	}
	nextSteps = append(nextSteps, "Update business records with primary and secondary industries")

	// Ensure we always return a non-nil slice
	if nextSteps == nil {
		nextSteps = []string{"Review multi-industry classification results"}
	}

	return nextSteps
}

func (p *ResultPresentationEngine) createSelectionMetrics(result *MultiIndustryClassification) *SelectionMetrics {
	// Implementation would create selection metrics
	return &SelectionMetrics{}
}

func (p *ResultPresentationEngine) createMetadata(result *ClassificationResponse, request *ClassificationRequest) map[string]interface{} {
	return map[string]interface{}{
		"presentation_format": p.formatOutput,
		"version":             "1.0",
		"timestamp":           time.Now(),
	}
}

func (p *ResultPresentationEngine) createMultiIndustryMetadata(result *MultiIndustryClassification, request *ClassificationRequest) map[string]interface{} {
	return map[string]interface{}{
		"presentation_format": p.formatOutput,
		"version":             "1.0",
		"timestamp":           time.Now(),
		"num_classifications": len(result.Classifications),
	}
}

// Configuration methods
func (p *ResultPresentationEngine) SetFormat(format string) {
	p.formatOutput = format
}

func (p *ResultPresentationEngine) SetIncludeConfidenceBreakdown(include bool) {
	p.includeConfidenceBreakdown = include
}

func (p *ResultPresentationEngine) SetIncludeReliabilityFactors(include bool) {
	p.includeReliabilityFactors = include
}

func (p *ResultPresentationEngine) SetIncludeUncertaintyFactors(include bool) {
	p.includeUncertaintyFactors = include
}

func (p *ResultPresentationEngine) SetIncludeProcessingMetrics(include bool) {
	p.includeProcessingMetrics = include
}

func (p *ResultPresentationEngine) SetIncludeRecommendations(include bool) {
	p.includeRecommendations = include
}
