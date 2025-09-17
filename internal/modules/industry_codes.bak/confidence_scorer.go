package industry_codes

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ConfidenceFactors represents the various factors that contribute to confidence scoring
type ConfidenceFactors struct {
	TextMatchScore      float64            `json:"text_match_score"`      // 0.0-1.0
	KeywordMatchScore   float64            `json:"keyword_match_score"`   // 0.0-1.0
	NameMatchScore      float64            `json:"name_match_score"`      // 0.0-1.0
	CategoryMatchScore  float64            `json:"category_match_score"`  // 0.0-1.0
	CodeQualityScore    float64            `json:"code_quality_score"`    // 0.0-1.0
	UsageFrequencyScore float64            `json:"usage_frequency_score"` // 0.0-1.0
	ContextualScore     float64            `json:"contextual_score"`      // 0.0-1.0
	ValidationScore     float64            `json:"validation_score"`      // 0.0-1.0
	CustomFactors       map[string]float64 `json:"custom_factors"`        // Additional custom factors
}

// ConfidenceScore represents a comprehensive confidence score with detailed breakdown
type ConfidenceScore struct {
	OverallScore       float64            `json:"overall_score"`       // 0.0-1.0
	Factors            *ConfidenceFactors `json:"factors"`             // Detailed factor breakdown
	ConfidenceLevel    string             `json:"confidence_level"`    // "low", "medium", "high", "very_high"
	ValidationStatus   string             `json:"validation_status"`   // "valid", "warning", "invalid"
	ValidationMessages []string           `json:"validation_messages"` // Validation feedback
	Recommendations    []string           `json:"recommendations"`     // Improvement suggestions
	LastUpdated        time.Time          `json:"last_updated"`
	ScoreVersion       string             `json:"score_version"` // Algorithm version
	// Enhanced validation fields
	CalibrationData    *CalibrationData    `json:"calibration_data,omitempty"`    // Calibration information
	StatisticalMetrics *StatisticalMetrics `json:"statistical_metrics,omitempty"` // Statistical validation
	UncertaintyMetrics *UncertaintyMetrics `json:"uncertainty_metrics,omitempty"` // Uncertainty quantification
	CrossValidation    *CrossValidation    `json:"cross_validation,omitempty"`    // Cross-validation results
	BenchmarkData      *BenchmarkData      `json:"benchmark_data,omitempty"`      // Benchmarking information
}

// CalibrationData represents confidence score calibration information
type CalibrationData struct {
	CalibratedScore    float64   `json:"calibrated_score"`    // Calibrated confidence score
	CalibrationFactor  float64   `json:"calibration_factor"`  // Calibration factor applied
	CalibrationMethod  string    `json:"calibration_method"`  // Method used for calibration
	CalibrationQuality float64   `json:"calibration_quality"` // Quality of calibration (0.0-1.0)
	LastCalibrated     time.Time `json:"last_calibrated"`     // When calibration was last performed
	CalibrationSample  int       `json:"calibration_sample"`  // Sample size used for calibration
}

// StatisticalMetrics represents statistical validation metrics
type StatisticalMetrics struct {
	ZScore               float64    `json:"z_score"`                // Z-score for statistical validation
	PValue               float64    `json:"p_value"`                // P-value for significance testing
	ConfidenceInterval   [2]float64 `json:"confidence_interval"`    // 95% confidence interval
	StandardError        float64    `json:"standard_error"`         // Standard error of the score
	ReliabilityIndex     float64    `json:"reliability_index"`      // Reliability index (0.0-1.0)
	SignificanceLevel    float64    `json:"significance_level"`     // Statistical significance level
	IsStatisticallyValid bool       `json:"is_statistically_valid"` // Whether score is statistically valid
}

// UncertaintyMetrics represents uncertainty quantification
type UncertaintyMetrics struct {
	UncertaintyScore    float64            `json:"uncertainty_score"`    // Overall uncertainty (0.0-1.0)
	FactorUncertainties map[string]float64 `json:"factor_uncertainties"` // Uncertainty per factor
	TotalUncertainty    float64            `json:"total_uncertainty"`    // Total uncertainty
	ConfidenceRange     [2]float64         `json:"confidence_range"`     // Confidence range
	ReliabilityScore    float64            `json:"reliability_score"`    // Reliability score (0.0-1.0)
	StabilityIndex      float64            `json:"stability_index"`      // Stability index (0.0-1.0)
}

// CrossValidation represents cross-validation results
type CrossValidation struct {
	CrossValidationScore float64   `json:"cross_validation_score"` // Cross-validation score
	FoldScores           []float64 `json:"fold_scores"`            // Scores from each fold
	MeanScore            float64   `json:"mean_score"`             // Mean cross-validation score
	StandardDeviation    float64   `json:"standard_deviation"`     // Standard deviation of scores
	IsStable             bool      `json:"is_stable"`              // Whether score is stable across folds
	StabilityIndex       float64   `json:"stability_index"`        // Stability index (0.0-1.0)
}

// BenchmarkData represents confidence score benchmarking information
type BenchmarkData struct {
	BenchmarkScore      float64              `json:"benchmark_score"`      // Benchmark confidence score
	BenchmarkMethod     string               `json:"benchmark_method"`     // Method used for benchmarking
	BenchmarkQuality    float64              `json:"benchmark_quality"`    // Quality of benchmark (0.0-1.0)
	BenchmarkSample     int                  `json:"benchmark_sample"`     // Sample size used for benchmarking
	BenchmarkMetrics    *BenchmarkMetrics    `json:"benchmark_metrics"`    // Detailed benchmark metrics
	BenchmarkComparison *BenchmarkComparison `json:"benchmark_comparison"` // Comparison with benchmarks
	LastBenchmarked     time.Time            `json:"last_benchmarked"`     // When benchmarking was last performed
	BenchmarkVersion    string               `json:"benchmark_version"`    // Version of benchmark data
}

// BenchmarkMetrics represents detailed benchmark metrics
type BenchmarkMetrics struct {
	IndustryBenchmark   float64 `json:"industry_benchmark"`   // Industry-specific benchmark
	CodeTypeBenchmark   float64 `json:"code_type_benchmark"`  // Code type-specific benchmark
	HistoricalBenchmark float64 `json:"historical_benchmark"` // Historical performance benchmark
	PeerBenchmark       float64 `json:"peer_benchmark"`       // Peer comparison benchmark
	OverallBenchmark    float64 `json:"overall_benchmark"`    // Overall benchmark score
	BenchmarkConfidence float64 `json:"benchmark_confidence"` // Confidence in benchmark (0.0-1.0)
	BenchmarkTrend      string  `json:"benchmark_trend"`      // Trend direction (improving, declining, stable)
	BenchmarkPercentile float64 `json:"benchmark_percentile"` // Percentile rank (0.0-100.0)
}

// BenchmarkComparison represents comparison with benchmarks
type BenchmarkComparison struct {
	ScoreVsIndustry      float64  `json:"score_vs_industry"`     // Score vs industry benchmark
	ScoreVsCodeType      float64  `json:"score_vs_code_type"`    // Score vs code type benchmark
	ScoreVsHistorical    float64  `json:"score_vs_historical"`   // Score vs historical benchmark
	ScoreVsPeer          float64  `json:"score_vs_peer"`         // Score vs peer benchmark
	OverallPerformance   string   `json:"overall_performance"`   // Overall performance rating
	PerformanceGap       float64  `json:"performance_gap"`       // Gap from optimal performance
	ImprovementPotential float64  `json:"improvement_potential"` // Potential for improvement (0.0-1.0)
	Recommendations      []string `json:"recommendations"`       // Benchmark-based recommendations
}

// BenchmarkConfig represents configuration for benchmarking
type BenchmarkConfig struct {
	EnableBenchmarking        bool    `json:"enable_benchmarking"`
	BenchmarkUpdateInterval   int     `json:"benchmark_update_interval"`   // Days between updates
	BenchmarkQualityThreshold float64 `json:"benchmark_quality_threshold"` // Minimum quality for benchmarks
	BenchmarkSampleSize       int     `json:"benchmark_sample_size"`       // Minimum sample size
	BenchmarkTrendWindow      int     `json:"benchmark_trend_window"`      // Days for trend calculation
	BenchmarkPercentileMethod string  `json:"benchmark_percentile_method"` // Method for percentile calculation
	BenchmarkComparisonWeight float64 `json:"benchmark_comparison_weight"` // Weight for benchmark comparison
}

// ValidationRule represents a validation rule for confidence scoring
type ValidationRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"` // "threshold", "pattern", "logic", "statistical"
	Parameters  map[string]interface{} `json:"parameters"`
	Weight      float64                `json:"weight"`
	Enabled     bool                   `json:"enabled"`
}

// ConfidenceScorer provides comprehensive confidence scoring and validation
type ConfidenceScorer struct {
	db              *IndustryCodeDatabase
	metadataMgr     *MetadataManager
	logger          *zap.Logger
	validationRules []*ValidationRule
	weights         map[string]float64
	// Enhanced validation components
	calibrationData                 map[string]*CalibrationData
	historicalScores                []float64
	enableAdvancedValidation        bool
	enableCalibration               bool
	enableStatisticalValidation     bool
	enableUncertaintyQuantification bool
	enableCrossValidation           bool
	enableBenchmarking              bool
	benchmarkConfig                 *BenchmarkConfig
	benchmarkData                   map[string]*BenchmarkData
	industryBenchmarks              map[string]float64
	codeTypeBenchmarks              map[string]float64
}

// NewConfidenceScorer creates a new confidence scorer
func NewConfidenceScorer(db *IndustryCodeDatabase, metadataMgr *MetadataManager, logger *zap.Logger) *ConfidenceScorer {
	scorer := &ConfidenceScorer{
		db:          db,
		metadataMgr: metadataMgr,
		logger:      logger,
		weights: map[string]float64{
			"text_match":      0.25,
			"keyword_match":   0.20,
			"name_match":      0.15,
			"category_match":  0.10,
			"code_quality":    0.15,
			"usage_frequency": 0.10,
			"contextual":      0.05,
		},
		calibrationData:                 make(map[string]*CalibrationData),
		historicalScores:                make([]float64, 0),
		enableAdvancedValidation:        true,
		enableCalibration:               true,
		enableStatisticalValidation:     true,
		enableUncertaintyQuantification: true,
		enableCrossValidation:           true,
		enableBenchmarking:              true,
		benchmarkConfig: &BenchmarkConfig{
			EnableBenchmarking:        true,
			BenchmarkUpdateInterval:   7, // Update weekly
			BenchmarkQualityThreshold: 0.7,
			BenchmarkSampleSize:       100,
			BenchmarkTrendWindow:      30, // 30 days for trend calculation
			BenchmarkPercentileMethod: "linear",
			BenchmarkComparisonWeight: 0.3,
		},
		benchmarkData:      make(map[string]*BenchmarkData),
		industryBenchmarks: make(map[string]float64),
		codeTypeBenchmarks: make(map[string]float64),
	}

	// Initialize default validation rules
	scorer.initializeValidationRules()

	return scorer
}

// CalculateConfidence calculates comprehensive confidence score for a classification result
func (cs *ConfidenceScorer) CalculateConfidence(ctx context.Context, result *ClassificationResult, request *ClassificationRequest) (*ConfidenceScore, error) {
	if result == nil || result.Code == nil {
		return nil, fmt.Errorf("invalid classification result")
	}

	cs.logger.Info("Calculating confidence score",
		zap.String("code", result.Code.Code),
		zap.String("type", string(result.Code.Type)))

	// Calculate individual factor scores
	factors := cs.calculateConfidenceFactors(ctx, result, request)

	// Calculate overall weighted score
	overallScore := cs.calculateOverallScore(factors)

	// Determine confidence level
	confidenceLevel := cs.determineConfidenceLevel(overallScore)

	// Validate the result
	validationStatus, validationMessages := cs.validateResult(result, request, factors)

	// Generate recommendations
	recommendations := cs.generateRecommendations(factors, validationStatus)

	score := &ConfidenceScore{
		OverallScore:       overallScore,
		Factors:            factors,
		ConfidenceLevel:    confidenceLevel,
		ValidationStatus:   validationStatus,
		ValidationMessages: validationMessages,
		Recommendations:    recommendations,
		LastUpdated:        time.Now(),
		ScoreVersion:       "2.0.0", // Updated version for enhanced validation
	}

	// Apply advanced validation if enabled
	if cs.enableAdvancedValidation {
		cs.applyAdvancedValidation(ctx, score, result, request)
	}

	// Apply calibration if enabled
	if cs.enableCalibration {
		cs.applyCalibration(ctx, score, result, request)
	}

	// Apply statistical validation if enabled
	if cs.enableStatisticalValidation {
		cs.applyStatisticalValidation(ctx, score, result, request)
	}

	// Apply uncertainty quantification if enabled
	if cs.enableUncertaintyQuantification {
		cs.applyUncertaintyQuantification(ctx, score, result, request)
	}

	// Apply cross-validation if enabled
	if cs.enableCrossValidation {
		cs.applyCrossValidation(ctx, score, result, request)
	}

	// Apply benchmarking if enabled
	if cs.enableBenchmarking {
		cs.applyBenchmarking(ctx, score, result, request)
	}

	// Update historical scores
	cs.updateHistoricalScores(overallScore)

	cs.logger.Info("Enhanced confidence score calculated",
		zap.String("code", result.Code.Code),
		zap.Float64("overall_score", overallScore),
		zap.String("confidence_level", confidenceLevel),
		zap.String("validation_status", validationStatus),
		zap.Bool("has_calibration", score.CalibrationData != nil),
		zap.Bool("has_statistical", score.StatisticalMetrics != nil),
		zap.Bool("has_uncertainty", score.UncertaintyMetrics != nil),
		zap.Bool("has_cross_validation", score.CrossValidation != nil))

	return score, nil
}

// calculateConfidenceFactors calculates individual confidence factors
func (cs *ConfidenceScorer) calculateConfidenceFactors(ctx context.Context, result *ClassificationResult, request *ClassificationRequest) *ConfidenceFactors {
	factors := &ConfidenceFactors{
		CustomFactors: make(map[string]float64),
	}

	// Text match score based on description similarity
	factors.TextMatchScore = cs.calculateTextMatchScore(result, request)

	// Keyword match score
	factors.KeywordMatchScore = cs.calculateKeywordMatchScore(result, request)

	// Name match score
	factors.NameMatchScore = cs.calculateNameMatchScore(result, request)

	// Category match score
	factors.CategoryMatchScore = cs.calculateCategoryMatchScore(result, request)

	// Code quality score based on metadata
	factors.CodeQualityScore = cs.calculateCodeQualityScore(ctx, result.Code)

	// Usage frequency score
	factors.UsageFrequencyScore = cs.calculateUsageFrequencyScore(ctx, result.Code)

	// Contextual score based on business context
	factors.ContextualScore = cs.calculateContextualScore(result, request)

	// Validation score
	factors.ValidationScore = cs.calculateValidationScore(result, request)

	return factors
}

// calculateTextMatchScore calculates score based on text similarity
func (cs *ConfidenceScorer) calculateTextMatchScore(result *ClassificationResult, request *ClassificationRequest) float64 {
	score := 0.0

	// Combine business name and description for analysis
	analysisText := strings.ToLower(request.BusinessName)
	if request.BusinessDescription != "" {
		analysisText += " " + strings.ToLower(request.BusinessDescription)
	}

	// Calculate similarity with code description
	codeDescription := strings.ToLower(result.Code.Description)
	similarity := cs.calculateTextSimilarity(analysisText, codeDescription)
	score += similarity * 0.6

	// Check for exact phrase matches
	exactMatches := cs.findExactPhraseMatches(analysisText, codeDescription)
	score += float64(exactMatches) * 0.2

	// Check for word overlap
	wordOverlap := cs.calculateWordOverlap(analysisText, codeDescription)
	score += wordOverlap * 0.2

	return math.Min(score, 1.0)
}

// calculateKeywordMatchScore calculates score based on keyword matches
func (cs *ConfidenceScorer) calculateKeywordMatchScore(result *ClassificationResult, request *ClassificationRequest) float64 {
	score := 0.0

	// Combine all text for keyword analysis
	analysisText := strings.ToLower(request.BusinessName)
	if request.BusinessDescription != "" {
		analysisText += " " + strings.ToLower(request.BusinessDescription)
	}
	if len(request.Keywords) > 0 {
		analysisText += " " + strings.ToLower(strings.Join(request.Keywords, " "))
	}

	// Check code keywords against analysis text
	keywordMatches := 0
	totalKeywords := len(result.Code.Keywords)

	for _, keyword := range result.Code.Keywords {
		keywordLower := strings.ToLower(keyword)
		if strings.Contains(analysisText, keywordLower) {
			keywordMatches++
		}
	}

	if totalKeywords > 0 {
		score = float64(keywordMatches) / float64(totalKeywords)
	}

	// Boost score for high-frequency keywords
	if keywordMatches > 0 {
		score += math.Min(float64(keywordMatches)*0.1, 0.3)
	}

	return math.Min(score, 1.0)
}

// calculateNameMatchScore calculates score based on business name matches
func (cs *ConfidenceScorer) calculateNameMatchScore(result *ClassificationResult, request *ClassificationRequest) float64 {
	score := 0.0
	businessName := strings.ToLower(request.BusinessName)
	codeDescription := strings.ToLower(result.Code.Description)

	// Check for business name words in code description
	nameWords := strings.Fields(businessName)
	matches := 0

	for _, word := range nameWords {
		if len(word) > 2 && strings.Contains(codeDescription, word) {
			matches++
		}
	}

	if len(nameWords) > 0 {
		score = float64(matches) / float64(len(nameWords))
	}

	// Check for industry indicators in business name
	industryIndicators := cs.extractIndustryIndicators(businessName)
	indicatorMatches := 0

	for _, indicator := range industryIndicators {
		if strings.Contains(codeDescription, indicator) {
			indicatorMatches++
		}
	}

	if len(industryIndicators) > 0 {
		indicatorScore := float64(indicatorMatches) / float64(len(industryIndicators))
		score = (score + indicatorScore) / 2.0
	}

	return math.Min(score, 1.0)
}

// calculateCategoryMatchScore calculates score based on category matches
func (cs *ConfidenceScorer) calculateCategoryMatchScore(result *ClassificationResult, request *ClassificationRequest) float64 {
	score := 0.0

	// Combine all text for category analysis
	analysisText := strings.ToLower(request.BusinessName)
	if request.BusinessDescription != "" {
		analysisText += " " + strings.ToLower(request.BusinessDescription)
	}

	codeCategory := strings.ToLower(result.Code.Category)

	// Check for category words in analysis text
	categoryWords := strings.Fields(codeCategory)
	matches := 0

	for _, word := range categoryWords {
		if len(word) > 2 && strings.Contains(analysisText, word) {
			matches++
		}
	}

	if len(categoryWords) > 0 {
		score = float64(matches) / float64(len(categoryWords))
	}

	// Check for category synonyms
	synonymMatches := cs.checkCategorySynonyms(codeCategory, analysisText)
	score = math.Max(score, synonymMatches)

	return math.Min(score, 1.0)
}

// calculateCodeQualityScore calculates score based on code metadata quality
func (cs *ConfidenceScorer) calculateCodeQualityScore(ctx context.Context, code *IndustryCode) float64 {
	score := 0.5 // Base score

	// Get metadata for the code
	metadata, err := cs.metadataMgr.GetCodeMetadata(ctx, code.Code, "latest")
	if err == nil && metadata != nil {
		// Factor in data quality
		if metadata.DataQuality != "" {
			switch metadata.DataQuality {
			case "high":
				score += 0.3
			case "medium":
				score += 0.2
			case "low":
				score += 0.1
			}
		}

		// Factor in last update recency
		daysSinceUpdate := time.Since(metadata.LastUpdated).Hours() / 24
		if daysSinceUpdate < 365 { // Less than a year old
			score += 0.2
		} else if daysSinceUpdate < 730 { // Less than 2 years old
			score += 0.1
		}

		// Factor in source reliability
		if metadata.Source != "" {
			// Simple source reliability scoring
			reliableSources := map[string]bool{
				"official":   true,
				"government": true,
				"standard":   true,
			}
			if reliableSources[strings.ToLower(metadata.Source)] {
				score += 0.2
			}
		}
	}

	// Factor in code confidence from database
	score += code.Confidence * 0.2

	return math.Min(score, 1.0)
}

// calculateUsageFrequencyScore calculates score based on code usage frequency
func (cs *ConfidenceScorer) calculateUsageFrequencyScore(ctx context.Context, code *IndustryCode) float64 {
	score := 0.5 // Base score

	// Get metadata for usage statistics
	metadata, err := cs.metadataMgr.GetCodeMetadata(ctx, code.Code, "latest")
	if err == nil && metadata != nil {
		// Factor in usage count
		if metadata.UsageCount > 0 {
			// Normalize usage count (log scale)
			logUsage := math.Log10(float64(metadata.UsageCount))
			usageScore := math.Min(logUsage/3.0, 1.0) // Max score at 1000+ uses
			score += usageScore * 0.4
		}

		// Factor in last update as proxy for recent usage
		daysSinceUpdate := time.Since(metadata.LastUpdated).Hours() / 24
		if daysSinceUpdate < 30 { // Updated in last month
			score += 0.3
		} else if daysSinceUpdate < 90 { // Updated in last 3 months
			score += 0.2
		} else if daysSinceUpdate < 365 { // Updated in last year
			score += 0.1
		}
	}

	return math.Min(score, 1.0)
}

// calculateContextualScore calculates score based on business context
func (cs *ConfidenceScorer) calculateContextualScore(result *ClassificationResult, request *ClassificationRequest) float64 {
	score := 0.5 // Base score

	// Check if website is provided and relevant
	if request.Website != "" {
		// Extract domain keywords
		domainKeywords := cs.extractDomainKeywords(request.Website)

		// Check domain keywords against code description
		codeDescription := strings.ToLower(result.Code.Description)
		matches := 0

		for _, keyword := range domainKeywords {
			if strings.Contains(codeDescription, keyword) {
				matches++
			}
		}

		if len(domainKeywords) > 0 {
			domainScore := float64(matches) / float64(len(domainKeywords))
			score += domainScore * 0.3
		}
	}

	// Check for preferred code types
	if len(request.PreferredCodeTypes) > 0 {
		for _, preferredType := range request.PreferredCodeTypes {
			if result.Code.Type == preferredType {
				score += 0.2
				break
			}
		}
	}

	return math.Min(score, 1.0)
}

// calculateValidationScore calculates score based on validation rules
func (cs *ConfidenceScorer) calculateValidationScore(result *ClassificationResult, request *ClassificationRequest) float64 {
	score := 1.0 // Start with perfect score

	// Apply validation rules
	for _, rule := range cs.validationRules {
		if !rule.Enabled {
			continue
		}

		ruleScore := cs.applyValidationRule(rule, result, request)
		score *= (ruleScore*rule.Weight + (1.0 - rule.Weight))
	}

	return score
}

// calculateOverallScore calculates the overall weighted confidence score
func (cs *ConfidenceScorer) calculateOverallScore(factors *ConfidenceFactors) float64 {
	score := 0.0

	score += factors.TextMatchScore * cs.weights["text_match"]
	score += factors.KeywordMatchScore * cs.weights["keyword_match"]
	score += factors.NameMatchScore * cs.weights["name_match"]
	score += factors.CategoryMatchScore * cs.weights["category_match"]
	score += factors.CodeQualityScore * cs.weights["code_quality"]
	score += factors.UsageFrequencyScore * cs.weights["usage_frequency"]
	score += factors.ContextualScore * cs.weights["contextual"]

	// Apply validation score as a multiplier
	score *= factors.ValidationScore

	return math.Min(score, 1.0)
}

// determineConfidenceLevel determines the confidence level based on overall score
func (cs *ConfidenceScorer) determineConfidenceLevel(score float64) string {
	switch {
	case score >= 0.9:
		return "very_high"
	case score >= 0.7:
		return "high"
	case score >= 0.5:
		return "medium"
	case score >= 0.3:
		return "low"
	default:
		return "very_low"
	}
}

// validateResult validates the classification result
func (cs *ConfidenceScorer) validateResult(result *ClassificationResult, request *ClassificationRequest, factors *ConfidenceFactors) (string, []string) {
	messages := make([]string, 0)
	status := "valid"

	// Check minimum confidence threshold
	if result.Confidence < 0.3 {
		status = "invalid"
		messages = append(messages, "Confidence score below minimum threshold")
	} else if result.Confidence < 0.5 {
		status = "warning"
		messages = append(messages, "Low confidence score - consider manual review")
	}

	// Check for missing critical information
	if request.BusinessName == "" {
		status = "warning"
		messages = append(messages, "Business name is required for accurate classification")
	}

	if request.BusinessDescription == "" {
		status = "warning"
		messages = append(messages, "Business description would improve classification accuracy")
	}

	// Check for conflicting indicators
	if factors.TextMatchScore < 0.2 && factors.KeywordMatchScore < 0.2 {
		status = "warning"
		messages = append(messages, "Low text and keyword match scores suggest potential misclassification")
	}

	// Check code quality
	if factors.CodeQualityScore < 0.3 {
		status = "warning"
		messages = append(messages, "Low code quality score - data may be outdated or unreliable")
	}

	return status, messages
}

// generateRecommendations generates improvement recommendations
func (cs *ConfidenceScorer) generateRecommendations(factors *ConfidenceFactors, validationStatus string) []string {
	var recommendations []string

	if factors.TextMatchScore < 0.5 {
		recommendations = append(recommendations, "Provide more detailed business description to improve text matching")
	}

	if factors.KeywordMatchScore < 0.3 {
		recommendations = append(recommendations, "Include relevant industry keywords in business description")
	}

	if factors.NameMatchScore < 0.4 {
		recommendations = append(recommendations, "Ensure business name clearly indicates industry or business type")
	}

	if factors.CodeQualityScore < 0.5 {
		recommendations = append(recommendations, "Consider using more recent or reliable industry codes")
	}

	if validationStatus == "warning" {
		recommendations = append(recommendations, "Review classification manually for accuracy")
	}

	return recommendations
}

// initializeValidationRules initializes default validation rules
func (cs *ConfidenceScorer) initializeValidationRules() {
	cs.validationRules = []*ValidationRule{
		{
			ID:          "min_confidence",
			Name:        "Minimum Confidence Threshold",
			Description: "Ensures minimum confidence score is met",
			Type:        "threshold",
			Parameters: map[string]interface{}{
				"min_score": 0.3,
			},
			Weight:  0.3,
			Enabled: true,
		},
		{
			ID:          "text_match_consistency",
			Name:        "Text Match Consistency",
			Description: "Validates consistency between text match and keyword match scores",
			Type:        "logic",
			Parameters: map[string]interface{}{
				"max_difference": 0.5,
			},
			Weight:  0.2,
			Enabled: true,
		},
		{
			ID:          "business_name_required",
			Name:        "Business Name Required",
			Description: "Ensures business name is provided for classification",
			Type:        "pattern",
			Parameters: map[string]interface{}{
				"required": true,
			},
			Weight:  0.2,
			Enabled: true,
		},
	}
}

// applyValidationRule applies a specific validation rule
func (cs *ConfidenceScorer) applyValidationRule(rule *ValidationRule, result *ClassificationResult, request *ClassificationRequest) float64 {
	switch rule.Type {
	case "threshold":
		return cs.applyThresholdRule(rule, result)
	case "logic":
		return cs.applyLogicRule(rule, result)
	case "pattern":
		return cs.applyPatternRule(rule, request)
	default:
		return 1.0
	}
}

// applyThresholdRule applies a threshold-based validation rule
func (cs *ConfidenceScorer) applyThresholdRule(rule *ValidationRule, result *ClassificationResult) float64 {
	if minScore, ok := rule.Parameters["min_score"].(float64); ok {
		if result.Confidence < minScore {
			return 0.0
		}
	}
	return 1.0
}

// applyLogicRule applies a logic-based validation rule
func (cs *ConfidenceScorer) applyLogicRule(rule *ValidationRule, result *ClassificationResult) float64 {
	if _, ok := rule.Parameters["max_difference"].(float64); ok {
		// This would need to be implemented based on specific logic requirements
		// For now, return a default score
		return 1.0
	}
	return 1.0
}

// applyPatternRule applies a pattern-based validation rule
func (cs *ConfidenceScorer) applyPatternRule(rule *ValidationRule, request *ClassificationRequest) float64 {
	if required, ok := rule.Parameters["required"].(bool); ok && required {
		if request.BusinessName == "" {
			return 0.0
		}
	}
	return 1.0
}

// Helper methods for text analysis

func (cs *ConfidenceScorer) calculateTextSimilarity(text1, text2 string) float64 {
	words1 := cs.extractKeywords(text1)
	words2 := cs.extractKeywords(text2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	// Jaccard similarity
	intersection := 0
	wordSet1 := make(map[string]bool)
	wordSet2 := make(map[string]bool)

	for _, word := range words1 {
		wordSet1[word] = true
	}

	for _, word := range words2 {
		wordSet2[word] = true
		if wordSet1[word] {
			intersection++
		}
	}

	union := len(wordSet1) + len(wordSet2) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

func (cs *ConfidenceScorer) findExactPhraseMatches(text1, text2 string) int {
	matches := 0
	words1 := strings.Fields(text1)

	for i := 0; i <= len(words1)-2; i++ {
		phrase := strings.Join(words1[i:i+2], " ")
		if strings.Contains(text2, phrase) {
			matches++
		}
	}

	return matches
}

func (cs *ConfidenceScorer) calculateWordOverlap(text1, text2 string) float64 {
	words1 := cs.extractKeywords(text1)
	words2 := cs.extractKeywords(text2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	overlap := 0
	for _, word1 := range words1 {
		for _, word2 := range words2 {
			if word1 == word2 {
				overlap++
				break
			}
		}
	}

	return float64(overlap) / float64(len(words1))
}

func (cs *ConfidenceScorer) extractKeywords(text string) []string {
	// Clean text
	text = strings.ToLower(text)
	text = regexp.MustCompile(`[^\w\s]`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	// Split into words and filter out common stop words
	words := strings.Fields(text)
	var keywords []string

	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "being": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true, "could": true,
		"should": true, "may": true, "might": true, "can": true, "this": true, "that": true,
		"these": true, "those": true, "i": true, "you": true, "he": true, "she": true,
		"it": true, "we": true, "they": true, "me": true, "him": true, "her": true,
		"us": true, "them": true, "my": true, "your": true, "his": true,
		"its": true, "our": true, "their": true, "mine": true, "yours": true,
		"hers": true, "ours": true, "theirs": true,
	}

	for _, word := range words {
		if len(word) > 2 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

func (cs *ConfidenceScorer) extractIndustryIndicators(businessName string) []string {
	// Common industry indicators
	indicators := map[string][]string{
		"restaurant":   {"restaurant", "cafe", "diner", "bistro", "grill", "pizza", "burger"},
		"retail":       {"store", "shop", "market", "retail", "outlet", "boutique"},
		"technology":   {"tech", "software", "digital", "computer", "it", "systems"},
		"healthcare":   {"medical", "health", "clinic", "hospital", "pharmacy", "dental"},
		"automotive":   {"auto", "car", "motor", "tire", "repair", "service"},
		"construction": {"construction", "building", "contractor", "renovation"},
		"financial":    {"bank", "credit", "loan", "insurance", "financial", "investment"},
		"education":    {"school", "university", "college", "academy", "institute"},
		"legal":        {"law", "legal", "attorney", "lawyer", "firm"},
		"real_estate":  {"realty", "property", "estate", "realtor", "broker"},
	}

	var foundIndicators []string
	nameLower := strings.ToLower(businessName)

	for category, categoryIndicators := range indicators {
		for _, indicator := range categoryIndicators {
			if strings.Contains(nameLower, indicator) {
				foundIndicators = append(foundIndicators, category)
				foundIndicators = append(foundIndicators, indicator)
			}
		}
	}

	return cs.deduplicateStringSlice(foundIndicators)
}

func (cs *ConfidenceScorer) checkCategorySynonyms(category, text string) float64 {
	// Category synonyms mapping
	synonyms := map[string][]string{
		"retail":       {"retail", "store", "shop", "market", "outlet"},
		"technology":   {"technology", "tech", "software", "digital", "computer"},
		"healthcare":   {"healthcare", "medical", "health", "clinic", "hospital"},
		"automotive":   {"automotive", "auto", "car", "motor", "vehicle"},
		"construction": {"construction", "building", "contractor", "renovation"},
		"financial":    {"financial", "bank", "credit", "loan", "insurance"},
		"education":    {"education", "school", "university", "college", "academy"},
		"legal":        {"legal", "law", "attorney", "lawyer", "firm"},
		"real_estate":  {"real estate", "realty", "property", "estate", "realtor"},
	}

	categoryLower := strings.ToLower(category)
	textLower := strings.ToLower(text)

	// Check if category has synonyms
	if categorySynonyms, exists := synonyms[categoryLower]; exists {
		matches := 0
		for _, synonym := range categorySynonyms {
			if strings.Contains(textLower, synonym) {
				matches++
			}
		}
		if len(categorySynonyms) > 0 {
			return float64(matches) / float64(len(categorySynonyms))
		}
	}

	return 0.0
}

func (cs *ConfidenceScorer) extractDomainKeywords(website string) []string {
	// Extract domain from website
	domain := website
	if strings.HasPrefix(domain, "http://") {
		domain = strings.TrimPrefix(domain, "http://")
	} else if strings.HasPrefix(domain, "https://") {
		domain = strings.TrimPrefix(domain, "https://")
	}

	// Remove www prefix
	domain = strings.TrimPrefix(domain, "www.")

	// Split domain into parts
	parts := strings.Split(domain, ".")
	if len(parts) > 0 {
		mainPart := parts[0]
		// Split by common separators
		keywords := strings.FieldsFunc(mainPart, func(r rune) bool {
			return r == '-' || r == '_' || r == '.'
		})

		// Filter out common domain words
		filtered := []string{}
		commonWords := map[string]bool{
			"www": true, "com": true, "org": true, "net": true, "co": true, "inc": true,
			"llc": true, "corp": true, "ltd": true, "company": true, "business": true,
		}

		for _, keyword := range keywords {
			if len(keyword) > 2 && !commonWords[strings.ToLower(keyword)] {
				filtered = append(filtered, strings.ToLower(keyword))
			}
		}

		return filtered
	}

	return []string{}
}

func (cs *ConfidenceScorer) deduplicateStringSlice(slice []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// applyAdvancedValidation applies advanced validation techniques
func (cs *ConfidenceScorer) applyAdvancedValidation(ctx context.Context, score *ConfidenceScore, result *ClassificationResult, request *ClassificationRequest) {
	// Enhanced validation logic
	enhancedMessages := cs.performEnhancedValidation(result, request, score.Factors)
	score.ValidationMessages = append(score.ValidationMessages, enhancedMessages...)

	// Update validation status based on enhanced validation
	if len(enhancedMessages) > 0 {
		// Check if any enhanced validation messages indicate critical issues
		hasCriticalIssues := false
		for _, msg := range enhancedMessages {
			if strings.Contains(strings.ToLower(msg), "critical") || strings.Contains(strings.ToLower(msg), "invalid") {
				hasCriticalIssues = true
				break
			}
		}

		if hasCriticalIssues && score.ValidationStatus != "invalid" {
			score.ValidationStatus = "warning"
		}
	}
}

// applyCalibration applies confidence score calibration
func (cs *ConfidenceScorer) applyCalibration(ctx context.Context, score *ConfidenceScore, result *ClassificationResult, request *ClassificationRequest) {
	calibrationData := cs.calculateCalibrationData(score.OverallScore, result, request)
	score.CalibrationData = calibrationData

	// Apply calibrated score if calibration quality is good
	if calibrationData.CalibrationQuality >= 0.7 {
		score.OverallScore = calibrationData.CalibratedScore
		cs.logger.Debug("Applied calibrated confidence score",
			zap.Float64("original_score", score.OverallScore),
			zap.Float64("calibrated_score", calibrationData.CalibratedScore),
			zap.Float64("calibration_quality", calibrationData.CalibrationQuality))
	}
}

// applyStatisticalValidation applies statistical validation
func (cs *ConfidenceScorer) applyStatisticalValidation(ctx context.Context, score *ConfidenceScore, result *ClassificationResult, request *ClassificationRequest) {
	statisticalMetrics := cs.calculateStatisticalMetrics(score.OverallScore, result, request)
	score.StatisticalMetrics = statisticalMetrics

	// Update validation status based on statistical validity
	if !statisticalMetrics.IsStatisticallyValid {
		score.ValidationMessages = append(score.ValidationMessages,
			"Confidence score failed statistical validation - consider manual review")
		if score.ValidationStatus == "valid" {
			score.ValidationStatus = "warning"
		}
	}
}

// applyUncertaintyQuantification applies uncertainty quantification
func (cs *ConfidenceScorer) applyUncertaintyQuantification(ctx context.Context, score *ConfidenceScore, result *ClassificationResult, request *ClassificationRequest) {
	uncertaintyMetrics := cs.calculateUncertaintyMetrics(score.Factors, result, request)
	score.UncertaintyMetrics = uncertaintyMetrics

	// Add uncertainty-based recommendations
	if uncertaintyMetrics.UncertaintyScore > 0.3 {
		score.Recommendations = append(score.Recommendations,
			"High uncertainty detected - consider providing additional business information")
	}
}

// applyCrossValidation applies cross-validation techniques
func (cs *ConfidenceScorer) applyCrossValidation(ctx context.Context, score *ConfidenceScore, result *ClassificationResult, request *ClassificationRequest) {
	crossValidation := cs.performCrossValidation(score.Factors, result, request)
	score.CrossValidation = crossValidation

	// Update validation status based on cross-validation stability
	if !crossValidation.IsStable {
		score.ValidationMessages = append(score.ValidationMessages,
			"Confidence score shows instability across validation folds - consider manual review")
		if score.ValidationStatus == "valid" {
			score.ValidationStatus = "warning"
		}
	}
}

// applyBenchmarking applies confidence score benchmarking
func (cs *ConfidenceScorer) applyBenchmarking(ctx context.Context, score *ConfidenceScore, result *ClassificationResult, request *ClassificationRequest) {
	benchmarkData := cs.calculateBenchmarkData(score.OverallScore, result, request)
	score.BenchmarkData = benchmarkData

	// Apply benchmark-based adjustments if benchmark quality is good
	if benchmarkData.BenchmarkQuality >= cs.benchmarkConfig.BenchmarkQualityThreshold {
		// Adjust score based on benchmark comparison
		adjustedScore := cs.adjustScoreBasedOnBenchmark(score.OverallScore, benchmarkData)
		score.OverallScore = adjustedScore

		cs.logger.Debug("Applied benchmark-adjusted confidence score",
			zap.Float64("original_score", score.OverallScore),
			zap.Float64("benchmark_score", benchmarkData.BenchmarkScore),
			zap.Float64("adjusted_score", adjustedScore),
			zap.Float64("benchmark_quality", benchmarkData.BenchmarkQuality))
	}

	// Add benchmark-based recommendations
	if benchmarkData.BenchmarkComparison != nil && len(benchmarkData.BenchmarkComparison.Recommendations) > 0 {
		score.Recommendations = append(score.Recommendations, benchmarkData.BenchmarkComparison.Recommendations...)
	}
}

// calculateBenchmarkData calculates benchmark data for confidence score
func (cs *ConfidenceScorer) calculateBenchmarkData(score float64, result *ClassificationResult, request *ClassificationRequest) *BenchmarkData {
	// Calculate various benchmark metrics
	benchmarkMetrics := cs.calculateBenchmarkMetrics(score, result, request)

	// Calculate benchmark comparison
	benchmarkComparison := cs.calculateBenchmarkComparison(score, benchmarkMetrics, result, request)

	// Calculate overall benchmark score
	benchmarkScore := cs.calculateOverallBenchmarkScore(benchmarkMetrics)

	return &BenchmarkData{
		BenchmarkScore:      benchmarkScore,
		BenchmarkMethod:     "comprehensive_benchmarking",
		BenchmarkQuality:    cs.calculateBenchmarkQuality(benchmarkMetrics, result, request),
		BenchmarkSample:     len(cs.historicalScores),
		BenchmarkMetrics:    benchmarkMetrics,
		BenchmarkComparison: benchmarkComparison,
		LastBenchmarked:     time.Now(),
		BenchmarkVersion:    "1.0.0",
	}
}

// calculateBenchmarkMetrics calculates detailed benchmark metrics
func (cs *ConfidenceScorer) calculateBenchmarkMetrics(score float64, result *ClassificationResult, request *ClassificationRequest) *BenchmarkMetrics {
	// Calculate industry benchmark
	industryBenchmark := cs.calculateIndustryBenchmark(result, request)

	// Calculate code type benchmark
	codeTypeBenchmark := cs.calculateCodeTypeBenchmark(result)

	// Calculate historical benchmark
	historicalBenchmark := cs.calculateHistoricalBenchmark(score)

	// Calculate peer benchmark
	peerBenchmark := cs.calculatePeerBenchmark(result, request)

	// Calculate overall benchmark
	overallBenchmark := (industryBenchmark + codeTypeBenchmark + historicalBenchmark + peerBenchmark) / 4.0

	// Calculate benchmark confidence
	benchmarks := []float64{industryBenchmark, codeTypeBenchmark, historicalBenchmark, peerBenchmark}
	benchmarkConfidence := cs.calculateBenchmarkConfidence(benchmarks)

	// Calculate benchmark trend
	benchmarkTrend := cs.calculateBenchmarkTrend(score)

	// Calculate benchmark percentile
	benchmarkPercentile := cs.calculateBenchmarkPercentile(score, overallBenchmark)

	return &BenchmarkMetrics{
		IndustryBenchmark:   industryBenchmark,
		CodeTypeBenchmark:   codeTypeBenchmark,
		HistoricalBenchmark: historicalBenchmark,
		PeerBenchmark:       peerBenchmark,
		OverallBenchmark:    overallBenchmark,
		BenchmarkConfidence: benchmarkConfidence,
		BenchmarkTrend:      benchmarkTrend,
		BenchmarkPercentile: benchmarkPercentile,
	}
}

// calculateBenchmarkComparison calculates comparison with benchmarks
func (cs *ConfidenceScorer) calculateBenchmarkComparison(score float64, metrics *BenchmarkMetrics, result *ClassificationResult, request *ClassificationRequest) *BenchmarkComparison {
	// Calculate score vs various benchmarks
	scoreVsIndustry := score - metrics.IndustryBenchmark
	scoreVsCodeType := score - metrics.CodeTypeBenchmark
	scoreVsHistorical := score - metrics.HistoricalBenchmark
	scoreVsPeer := score - metrics.PeerBenchmark

	// Calculate overall performance
	overallPerformance := cs.determineOverallPerformance(score, metrics)

	// Calculate performance gap
	performanceGap := cs.calculatePerformanceGap(score, metrics)

	// Calculate improvement potential
	improvementPotential := cs.calculateImprovementPotential(score, metrics)

	// Generate benchmark-based recommendations
	recommendations := cs.generateBenchmarkRecommendations(score, metrics, result, request)

	return &BenchmarkComparison{
		ScoreVsIndustry:      scoreVsIndustry,
		ScoreVsCodeType:      scoreVsCodeType,
		ScoreVsHistorical:    scoreVsHistorical,
		ScoreVsPeer:          scoreVsPeer,
		OverallPerformance:   overallPerformance,
		PerformanceGap:       performanceGap,
		ImprovementPotential: improvementPotential,
		Recommendations:      recommendations,
	}
}

// calculateIndustryBenchmark calculates industry-specific benchmark
func (cs *ConfidenceScorer) calculateIndustryBenchmark(result *ClassificationResult, request *ClassificationRequest) float64 {
	// Extract industry from result
	industry := ""
	if result.Code != nil && result.Code.Category != "" {
		industry = result.Code.Category
	}

	// Get industry benchmark from cache or calculate
	if benchmark, exists := cs.industryBenchmarks[industry]; exists {
		return benchmark
	}

	// Calculate industry benchmark based on historical data
	// For now, use a simple heuristic based on code type
	baseBenchmark := 0.75 // Default benchmark

	switch result.Code.Type {
	case CodeTypeNAICS:
		baseBenchmark = 0.80 // NAICS codes typically have higher confidence
	case CodeTypeSIC:
		baseBenchmark = 0.70 // SIC codes may have lower confidence
	case CodeTypeMCC:
		baseBenchmark = 0.75 // MCC codes have moderate confidence
	}

	// Cache the benchmark
	cs.industryBenchmarks[industry] = baseBenchmark

	return baseBenchmark
}

// calculateCodeTypeBenchmark calculates code type-specific benchmark
func (cs *ConfidenceScorer) calculateCodeTypeBenchmark(result *ClassificationResult) float64 {
	// Get code type benchmark from cache or calculate
	codeType := string(result.Code.Type)
	if benchmark, exists := cs.codeTypeBenchmarks[codeType]; exists {
		return benchmark
	}

	// Calculate code type benchmark based on historical performance
	baseBenchmark := 0.75 // Default benchmark

	switch result.Code.Type {
	case CodeTypeNAICS:
		baseBenchmark = 0.82 // NAICS codes have high accuracy
	case CodeTypeSIC:
		baseBenchmark = 0.68 // SIC codes have moderate accuracy
	case CodeTypeMCC:
		baseBenchmark = 0.78 // MCC codes have good accuracy
	}

	// Cache the benchmark
	cs.codeTypeBenchmarks[codeType] = baseBenchmark

	return baseBenchmark
}

// calculateHistoricalBenchmark calculates historical performance benchmark
func (cs *ConfidenceScorer) calculateHistoricalBenchmark(score float64) float64 {
	if len(cs.historicalScores) == 0 {
		return 0.75 // Default benchmark if no historical data
	}

	// Calculate average of recent historical scores
	recentScores := cs.getRecentHistoricalScores(30) // Last 30 scores
	if len(recentScores) == 0 {
		return cs.calculateMean(cs.historicalScores)
	}

	return cs.calculateMean(recentScores)
}

// calculatePeerBenchmark calculates peer comparison benchmark
func (cs *ConfidenceScorer) calculatePeerBenchmark(result *ClassificationResult, request *ClassificationRequest) float64 {
	// Simulate peer benchmark based on similar businesses
	// In a real implementation, this would compare with similar businesses

	baseBenchmark := 0.75 // Default peer benchmark

	// Adjust based on business characteristics
	if request != nil && request.BusinessName != "" {
		// Simple heuristic: longer business names might indicate more established businesses
		if len(request.BusinessName) > 20 {
			baseBenchmark += 0.05
		}
	}

	// Adjust based on code type
	switch result.Code.Type {
	case CodeTypeNAICS:
		baseBenchmark += 0.03
	case CodeTypeSIC:
		baseBenchmark -= 0.02
	case CodeTypeMCC:
		baseBenchmark += 0.01
	}

	return math.Min(baseBenchmark, 1.0)
}

// calculateOverallBenchmarkScore calculates overall benchmark score
func (cs *ConfidenceScorer) calculateOverallBenchmarkScore(metrics *BenchmarkMetrics) float64 {
	// Weighted average of all benchmark metrics
	weights := map[string]float64{
		"industry":   0.3,
		"code_type":  0.25,
		"historical": 0.25,
		"peer":       0.2,
	}

	overallScore := metrics.IndustryBenchmark*weights["industry"] +
		metrics.CodeTypeBenchmark*weights["code_type"] +
		metrics.HistoricalBenchmark*weights["historical"] +
		metrics.PeerBenchmark*weights["peer"]

	return overallScore
}

// calculateBenchmarkConfidence calculates confidence in benchmark
func (cs *ConfidenceScorer) calculateBenchmarkConfidence(benchmarks []float64) float64 {
	// Calculate confidence based on sample size and consistency
	sampleSize := len(cs.historicalScores)

	// Base confidence on sample size
	baseConfidence := math.Min(float64(sampleSize)/float64(cs.benchmarkConfig.BenchmarkSampleSize), 1.0)

	// Adjust based on benchmark consistency
	variance := cs.calculateVariance(benchmarks)
	consistencyFactor := 1.0 - math.Min(variance, 0.5)

	// Final confidence is combination of sample size and consistency
	confidence := (baseConfidence*0.7 + consistencyFactor*0.3)

	return math.Min(confidence, 1.0)
}

// calculateBenchmarkQuality calculates quality of benchmark data
func (cs *ConfidenceScorer) calculateBenchmarkQuality(metrics *BenchmarkMetrics, result *ClassificationResult, request *ClassificationRequest) float64 {
	// Base quality on sample size
	sampleSize := len(cs.historicalScores)
	baseQuality := math.Min(float64(sampleSize)/float64(cs.benchmarkConfig.BenchmarkSampleSize), 1.0)

	// Adjust based on benchmark confidence
	confidenceFactor := metrics.BenchmarkConfidence

	// Adjust based on data completeness
	completenessFactor := 1.0
	if result == nil || result.Code == nil {
		completenessFactor = 0.8
	}
	if request == nil || request.BusinessName == "" {
		completenessFactor *= 0.9
	}

	// Final quality is combination of factors
	quality := (baseQuality*0.4 + confidenceFactor*0.4 + completenessFactor*0.2)

	return math.Min(quality, 1.0)
}

// calculateBenchmarkTrend calculates benchmark trend
func (cs *ConfidenceScorer) calculateBenchmarkTrend(score float64) string {
	if len(cs.historicalScores) < 10 {
		return "stable" // Not enough data for trend
	}

	// Get recent and older scores for trend calculation
	recentScores := cs.getRecentHistoricalScores(10)
	olderScores := cs.getRecentHistoricalScores(20)[10:] // Scores 11-20

	if len(recentScores) < 5 || len(olderScores) < 5 {
		return "stable"
	}

	recentAvg := cs.calculateMean(recentScores)
	olderAvg := cs.calculateMean(olderScores)

	trendThreshold := 0.05 // 5% change threshold

	if recentAvg > olderAvg+trendThreshold {
		return "improving"
	} else if recentAvg < olderAvg-trendThreshold {
		return "declining"
	}

	return "stable"
}

// calculateBenchmarkPercentile calculates percentile rank
func (cs *ConfidenceScorer) calculateBenchmarkPercentile(score float64, benchmark float64) float64 {
	if len(cs.historicalScores) == 0 {
		return 50.0 // Default to median if no data
	}

	// Calculate percentile based on historical scores
	sortedScores := make([]float64, len(cs.historicalScores))
	copy(sortedScores, cs.historicalScores)
	sort.Float64s(sortedScores)

	// Find position of current score
	position := 0
	for i, s := range sortedScores {
		if score >= s {
			position = i + 1
		}
	}

	percentile := float64(position) / float64(len(sortedScores)) * 100.0
	return percentile
}

// determineOverallPerformance determines overall performance rating
func (cs *ConfidenceScorer) determineOverallPerformance(score float64, metrics *BenchmarkMetrics) string {
	// Compare score against overall benchmark
	benchmarkDiff := score - metrics.OverallBenchmark

	if benchmarkDiff > 0.1 {
		return "excellent"
	} else if benchmarkDiff > 0.05 {
		return "good"
	} else if benchmarkDiff > -0.05 {
		return "average"
	} else if benchmarkDiff > -0.1 {
		return "below_average"
	} else {
		return "poor"
	}
}

// calculatePerformanceGap calculates gap from optimal performance
func (cs *ConfidenceScorer) calculatePerformanceGap(score float64, metrics *BenchmarkMetrics) float64 {
	// Calculate gap from perfect score (1.0)
	optimalGap := 1.0 - score

	// Calculate gap from benchmark
	benchmarkGap := math.Max(0, metrics.OverallBenchmark-score)

	// Weighted combination
	totalGap := optimalGap*0.7 + benchmarkGap*0.3

	return totalGap
}

// calculateImprovementPotential calculates potential for improvement
func (cs *ConfidenceScorer) calculateImprovementPotential(score float64, metrics *BenchmarkMetrics) float64 {
	// Calculate potential based on gap from optimal performance
	performanceGap := cs.calculatePerformanceGap(score, metrics)

	// Normalize to 0-1 range
	improvementPotential := performanceGap

	// Adjust based on benchmark percentile
	if metrics.BenchmarkPercentile < 25 {
		improvementPotential *= 1.2 // High potential for low percentile
	} else if metrics.BenchmarkPercentile > 75 {
		improvementPotential *= 0.8 // Lower potential for high percentile
	}

	return math.Min(improvementPotential, 1.0)
}

// generateBenchmarkRecommendations generates benchmark-based recommendations
func (cs *ConfidenceScorer) generateBenchmarkRecommendations(score float64, metrics *BenchmarkMetrics, result *ClassificationResult, request *ClassificationRequest) []string {
	var recommendations []string

	// Performance-based recommendations
	if score < metrics.OverallBenchmark {
		recommendations = append(recommendations, "Consider improving data quality to meet industry benchmarks")
	}

	if metrics.BenchmarkPercentile < 25 {
		recommendations = append(recommendations, "Score is in bottom quartile - review classification approach")
	}

	// Trend-based recommendations
	if metrics.BenchmarkTrend == "declining" {
		recommendations = append(recommendations, "Performance trend is declining - investigate recent changes")
	}

	// Code type specific recommendations
	if score < metrics.CodeTypeBenchmark {
		recommendations = append(recommendations, "Score below code type benchmark - verify classification accuracy")
	}

	// Industry specific recommendations
	if score < metrics.IndustryBenchmark {
		recommendations = append(recommendations, "Score below industry benchmark - consider industry-specific factors")
	}

	return recommendations
}

// adjustScoreBasedOnBenchmark adjusts score based on benchmark comparison
func (cs *ConfidenceScorer) adjustScoreBasedOnBenchmark(score float64, benchmarkData *BenchmarkData) float64 {
	if benchmarkData.BenchmarkComparison == nil {
		return score
	}

	// Calculate adjustment factor based on benchmark comparison
	adjustmentFactor := 1.0

	// Adjust based on overall performance
	switch benchmarkData.BenchmarkComparison.OverallPerformance {
	case "excellent":
		adjustmentFactor = 1.02 // Slight boost
	case "good":
		adjustmentFactor = 1.01 // Minor boost
	case "average":
		adjustmentFactor = 1.0 // No change
	case "below_average":
		adjustmentFactor = 0.99 // Minor reduction
	case "poor":
		adjustmentFactor = 0.98 // Reduction
	}

	// Apply adjustment
	adjustedScore := score * adjustmentFactor

	// Ensure score stays within bounds
	if adjustedScore > 1.0 {
		adjustedScore = 1.0
	} else if adjustedScore < 0.0 {
		adjustedScore = 0.0
	}

	return adjustedScore
}

// getRecentHistoricalScores gets recent historical scores
func (cs *ConfidenceScorer) getRecentHistoricalScores(count int) []float64 {
	if len(cs.historicalScores) == 0 {
		return []float64{}
	}

	start := len(cs.historicalScores) - count
	if start < 0 {
		start = 0
	}

	return cs.historicalScores[start:]
}

// performEnhancedValidation performs enhanced validation checks
func (cs *ConfidenceScorer) performEnhancedValidation(result *ClassificationResult, request *ClassificationRequest, factors *ConfidenceFactors) []string {
	var messages []string

	// Enhanced factor consistency checks
	if factors.TextMatchScore > 0.8 && factors.KeywordMatchScore < 0.3 {
		messages = append(messages, "High text match but low keyword match - potential data quality issue")
	}

	if factors.NameMatchScore > 0.7 && factors.CategoryMatchScore < 0.2 {
		messages = append(messages, "Strong name match but weak category match - verify industry classification")
	}

	// Enhanced code quality validation
	if factors.CodeQualityScore < 0.4 {
		messages = append(messages, "Low code quality score - data may be outdated or unreliable")
	}

	// Enhanced contextual validation
	if factors.ContextualScore < 0.2 && factors.TextMatchScore > 0.6 {
		messages = append(messages, "High text match but low contextual relevance - verify business context")
	}

	// Enhanced usage frequency validation
	if factors.UsageFrequencyScore < 0.3 {
		messages = append(messages, "Low usage frequency - code may be rarely used or outdated")
	}

	// Enhanced validation score checks
	if factors.ValidationScore < 0.5 {
		messages = append(messages, "Low validation score - multiple validation rules failed")
	}

	return messages
}

// calculateCalibrationData calculates calibration data for confidence score
func (cs *ConfidenceScorer) calculateCalibrationData(score float64, result *ClassificationResult, request *ClassificationRequest) *CalibrationData {
	// Simple calibration based on historical performance
	calibrationFactor := 1.0
	calibrationQuality := 0.8

	// Adjust calibration based on code type
	switch result.Code.Type {
	case CodeTypeNAICS:
		calibrationFactor = 1.05 // Slight boost for NAICS codes
	case CodeTypeSIC:
		calibrationFactor = 0.95 // Slight reduction for SIC codes
	case CodeTypeMCC:
		calibrationFactor = 1.02 // Minor boost for MCC codes
	}

	// Adjust based on confidence level
	if score < 0.5 {
		calibrationFactor *= 0.98 // Slight reduction for low confidence
	} else if score > 0.8 {
		calibrationFactor *= 1.02 // Slight boost for high confidence
	}

	calibratedScore := score * calibrationFactor
	if calibratedScore > 1.0 {
		calibratedScore = 1.0
	}

	return &CalibrationData{
		CalibratedScore:    calibratedScore,
		CalibrationFactor:  calibrationFactor,
		CalibrationMethod:  "historical_performance",
		CalibrationQuality: calibrationQuality,
		LastCalibrated:     time.Now(),
		CalibrationSample:  len(cs.historicalScores),
	}
}

// calculateStatisticalMetrics calculates statistical validation metrics
func (cs *ConfidenceScorer) calculateStatisticalMetrics(score float64, result *ClassificationResult, request *ClassificationRequest) *StatisticalMetrics {
	// Calculate basic statistical metrics
	mean := cs.calculateMean(cs.historicalScores)
	stdDev := cs.calculateStandardDeviation(cs.historicalScores)

	zScore := 0.0
	if stdDev > 0 {
		zScore = (score - mean) / stdDev
	}

	// Calculate confidence interval (simplified)
	confidenceInterval := [2]float64{
		math.Max(0, score-1.96*stdDev),
		math.Min(1, score+1.96*stdDev),
	}

	// Calculate p-value (simplified)
	pValue := 0.05 // Default p-value
	if math.Abs(zScore) > 1.96 {
		pValue = 0.01
	}

	// Calculate reliability index
	reliabilityIndex := 1.0 - math.Abs(zScore)/3.0 // Normalize to 0-1
	if reliabilityIndex < 0 {
		reliabilityIndex = 0
	}

	// Determine statistical validity
	isStatisticallyValid := math.Abs(zScore) < 2.0 && pValue > 0.01

	return &StatisticalMetrics{
		ZScore:               zScore,
		PValue:               pValue,
		ConfidenceInterval:   confidenceInterval,
		StandardError:        stdDev / math.Sqrt(float64(len(cs.historicalScores))),
		ReliabilityIndex:     reliabilityIndex,
		SignificanceLevel:    0.05,
		IsStatisticallyValid: isStatisticallyValid,
	}
}

// calculateUncertaintyMetrics calculates uncertainty quantification
func (cs *ConfidenceScorer) calculateUncertaintyMetrics(factors *ConfidenceFactors, result *ClassificationResult, request *ClassificationRequest) *UncertaintyMetrics {
	// Calculate factor uncertainties
	factorUncertainties := make(map[string]float64)

	// Uncertainty is inversely proportional to score (higher score = lower uncertainty)
	factorUncertainties["text_match"] = 1.0 - factors.TextMatchScore
	factorUncertainties["keyword_match"] = 1.0 - factors.KeywordMatchScore
	factorUncertainties["name_match"] = 1.0 - factors.NameMatchScore
	factorUncertainties["category_match"] = 1.0 - factors.CategoryMatchScore
	factorUncertainties["code_quality"] = 1.0 - factors.CodeQualityScore
	factorUncertainties["usage_frequency"] = 1.0 - factors.UsageFrequencyScore
	factorUncertainties["contextual"] = 1.0 - factors.ContextualScore
	factorUncertainties["validation"] = 1.0 - factors.ValidationScore

	// Calculate total uncertainty (weighted average)
	totalUncertainty := 0.0
	totalWeight := 0.0

	for factor, uncertainty := range factorUncertainties {
		weight := cs.weights[factor]
		totalUncertainty += uncertainty * weight
		totalWeight += weight
	}

	if totalWeight > 0 {
		totalUncertainty /= totalWeight
	}

	// Calculate overall uncertainty score
	uncertaintyScore := totalUncertainty

	// Calculate confidence range
	confidenceRange := [2]float64{
		math.Max(0, result.Confidence-uncertaintyScore),
		math.Min(1, result.Confidence+uncertaintyScore),
	}

	// Calculate reliability score (inverse of uncertainty)
	reliabilityScore := 1.0 - uncertaintyScore

	// Calculate stability index (based on factor consistency)
	stabilityIndex := cs.calculateStabilityIndex(factors)

	return &UncertaintyMetrics{
		UncertaintyScore:    uncertaintyScore,
		FactorUncertainties: factorUncertainties,
		TotalUncertainty:    totalUncertainty,
		ConfidenceRange:     confidenceRange,
		ReliabilityScore:    reliabilityScore,
		StabilityIndex:      stabilityIndex,
	}
}

// performCrossValidation performs cross-validation techniques
func (cs *ConfidenceScorer) performCrossValidation(factors *ConfidenceFactors, result *ClassificationResult, request *ClassificationRequest) *CrossValidation {
	// Simulate cross-validation by creating different factor combinations
	folds := 5
	foldScores := make([]float64, folds)

	// Create different factor weight combinations for each fold
	for i := 0; i < folds; i++ {
		// Vary weights slightly for each fold
		weightVariation := 0.1 * float64(i) / float64(folds-1)

		// Calculate score with varied weights
		score := factors.TextMatchScore*(cs.weights["text_match"]+weightVariation) +
			factors.KeywordMatchScore*(cs.weights["keyword_match"]-weightVariation*0.5) +
			factors.NameMatchScore*cs.weights["name_match"] +
			factors.CategoryMatchScore*cs.weights["category_match"] +
			factors.CodeQualityScore*cs.weights["code_quality"] +
			factors.UsageFrequencyScore*cs.weights["usage_frequency"] +
			factors.ContextualScore*cs.weights["contextual"]

		score *= factors.ValidationScore
		if score > 1.0 {
			score = 1.0
		}

		foldScores[i] = score
	}

	// Calculate cross-validation metrics
	meanScore := cs.calculateMean(foldScores)
	stdDev := cs.calculateStandardDeviation(foldScores)

	// Determine stability (low standard deviation = stable)
	stabilityIndex := 1.0 - stdDev
	if stabilityIndex < 0 {
		stabilityIndex = 0
	}

	isStable := stdDev < 0.1 // Stable if standard deviation < 0.1

	return &CrossValidation{
		CrossValidationScore: meanScore,
		FoldScores:           foldScores,
		MeanScore:            meanScore,
		StandardDeviation:    stdDev,
		IsStable:             isStable,
		StabilityIndex:       stabilityIndex,
	}
}

// calculateStabilityIndex calculates stability index based on factor consistency
func (cs *ConfidenceScorer) calculateStabilityIndex(factors *ConfidenceFactors) float64 {
	// Calculate variance of factor scores
	scores := []float64{
		factors.TextMatchScore,
		factors.KeywordMatchScore,
		factors.NameMatchScore,
		factors.CategoryMatchScore,
		factors.CodeQualityScore,
		factors.UsageFrequencyScore,
		factors.ContextualScore,
		factors.ValidationScore,
	}

	variance := cs.calculateVariance(scores)

	// Stability index is inverse of normalized variance
	stabilityIndex := 1.0 - math.Min(variance, 1.0)

	return stabilityIndex
}

// updateHistoricalScores updates historical scores for calibration
func (cs *ConfidenceScorer) updateHistoricalScores(score float64) {
	cs.historicalScores = append(cs.historicalScores, score)

	// Keep only last 1000 scores for memory efficiency
	if len(cs.historicalScores) > 1000 {
		cs.historicalScores = cs.historicalScores[len(cs.historicalScores)-1000:]
	}
}

// calculateMean calculates mean of float64 slice
func (cs *ConfidenceScorer) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}

// calculateVariance calculates variance of float64 slice
func (cs *ConfidenceScorer) calculateVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	mean := cs.calculateMean(values)
	variance := 0.0

	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}

	return variance / float64(len(values))
}

// calculateStandardDeviation calculates standard deviation of float64 slice
func (cs *ConfidenceScorer) calculateStandardDeviation(values []float64) float64 {
	variance := cs.calculateVariance(values)
	return math.Sqrt(variance)
}
