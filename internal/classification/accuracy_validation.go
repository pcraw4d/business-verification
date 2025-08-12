package classification

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// AccuracyValidationEngine provides comprehensive accuracy validation for industry classifications
type AccuracyValidationEngine struct {
	logger  *observability.Logger
	metrics *observability.Metrics

	// Configuration
	accuracyThresholds map[string]float64
	validationRules    map[string]ValidationRule
	feedbackRules      map[string]FeedbackRule

	// Validation data
	knownClassifications map[string]KnownClassification
	industryBenchmarks   map[string]IndustryBenchmark
	accuracyHistory      []AccuracyRecord

	// Metrics tracking
	totalValidations      int64
	successfulValidations int64
	averageAccuracy       float64
	accuracyByIndustry    map[string]float64
	accuracyByMethod      map[string]float64
}

// ValidationRule defines a rule for validating classification accuracy
type ValidationRule struct {
	Name        string
	Description string
	Weight      float64
	Validator   func(classification IndustryClassification, known KnownClassification) (bool, float64, string)
}

// FeedbackRule defines a rule for generating feedback on classification accuracy
type FeedbackRule struct {
	Name        string
	Description string
	Weight      float64
	Generator   func(classification IndustryClassification, known KnownClassification) (string, float64)
}

// KnownClassification represents a known correct classification for validation
type KnownClassification struct {
	BusinessName       string
	IndustryCode       string
	IndustryName       string
	ConfidenceLevel    string
	Source             string
	ValidationDate     time.Time
	IsVerified         bool
	VerificationMethod string
}

// IndustryBenchmark represents benchmark data for an industry
type IndustryBenchmark struct {
	IndustryCode      string
	IndustryName      string
	AverageAccuracy   float64
	SampleSize        int
	CommonKeywords    []string
	TypicalConfidence float64
	LastUpdated       time.Time
}

// AccuracyRecord represents a single accuracy validation record
type AccuracyRecord struct {
	BusinessName         string
	ClassificationID     string
	PredictedCode        string
	ActualCode           string
	Accuracy             float64
	ValidationMethod     string
	Timestamp            time.Time
	ConfidenceScore      float64
	ClassificationMethod string
	Feedback             []string
}

// ValidationResult represents the result of an accuracy validation
type ValidationResult struct {
	IsAccurate          bool
	AccuracyScore       float64
	ConfidenceLevel     string
	ValidationMethod    string
	Feedback            []string
	Recommendations     []string
	BenchmarkComparison *BenchmarkComparison
	HistoricalTrend     *AccuracyTrend
	Timestamp           time.Time
}

// BenchmarkComparison compares current accuracy against industry benchmarks
type BenchmarkComparison struct {
	IndustryBenchmark float64
	CurrentAccuracy   float64
	Difference        float64
	Percentile        float64
	IsAboveBenchmark  bool
	BenchmarkSource   string
}

// AccuracyTrend shows historical accuracy trends
type AccuracyTrend struct {
	PeriodDays      int
	AverageAccuracy float64
	TrendDirection  string // "improving", "declining", "stable"
	TrendStrength   float64
	DataPoints      int
}

// NewAccuracyValidationEngine creates a new accuracy validation engine
func NewAccuracyValidationEngine(logger *observability.Logger, metrics *observability.Metrics) *AccuracyValidationEngine {
	engine := &AccuracyValidationEngine{
		logger:  logger,
		metrics: metrics,

		// Configuration
		accuracyThresholds: map[string]float64{
			"excellent":  0.95,
			"good":       0.85,
			"acceptable": 0.75,
			"poor":       0.60,
		},
		validationRules: make(map[string]ValidationRule),
		feedbackRules:   make(map[string]FeedbackRule),

		// Data storage
		knownClassifications: make(map[string]KnownClassification),
		industryBenchmarks:   make(map[string]IndustryBenchmark),
		accuracyHistory:      make([]AccuracyRecord, 0),

		// Metrics
		accuracyByIndustry: make(map[string]float64),
		accuracyByMethod:   make(map[string]float64),
	}

	// Initialize validation rules
	engine.initializeValidationRules()

	// Initialize feedback rules
	engine.initializeFeedbackRules()

	return engine
}

// ValidateClassification validates a classification against known data
func (a *AccuracyValidationEngine) ValidateClassification(ctx context.Context, classification IndustryClassification, businessName string) *ValidationResult {
	start := time.Now()

	// Log validation start
	if a.logger != nil {
		a.logger.WithComponent("accuracy_validation").LogBusinessEvent(ctx, "accuracy_validation_started", "", map[string]interface{}{
			"business_name":    businessName,
			"industry_code":    classification.IndustryCode,
			"confidence_score": classification.ConfidenceScore,
		})
	}

	// Find known classification for this business
	known, exists := a.knownClassifications[businessName]

	var result *ValidationResult
	if exists {
		result = a.validateAgainstKnown(ctx, classification, known)
	} else {
		result = a.validateAgainstBenchmarks(ctx, classification)
	}

	// Record accuracy metrics
	a.recordAccuracyMetrics(classification, result)

	// Generate feedback and recommendations
	result.Feedback = a.generateFeedback(classification, known)
	result.Recommendations = a.generateRecommendations(classification, result)

	// Calculate benchmark comparison
	result.BenchmarkComparison = a.calculateBenchmarkComparison(classification.IndustryCode, result.AccuracyScore)

	// Calculate historical trend
	result.HistoricalTrend = a.calculateAccuracyTrend(classification.IndustryCode)

	// Log validation completion
	if a.logger != nil {
		a.logger.WithComponent("accuracy_validation").LogBusinessEvent(ctx, "accuracy_validation_completed", "", map[string]interface{}{
			"business_name":      businessName,
			"accuracy_score":     result.AccuracyScore,
			"is_accurate":        result.IsAccurate,
			"processing_time_ms": time.Since(start).Milliseconds(),
		})
	}

	return result
}

// validateAgainstKnown validates classification against known correct data
func (a *AccuracyValidationEngine) validateAgainstKnown(ctx context.Context, classification IndustryClassification, known KnownClassification) *ValidationResult {
	result := &ValidationResult{
		Timestamp: time.Now(),
	}

	// Calculate accuracy score
	accuracyScore := a.calculateAccuracyScore(classification, known)
	result.AccuracyScore = accuracyScore

	// Determine if accurate based on thresholds
	result.IsAccurate = accuracyScore >= a.accuracyThresholds["acceptable"]
	result.ConfidenceLevel = a.determineAccuracyLevel(accuracyScore)
	result.ValidationMethod = "known_data_comparison"

	// Record accuracy history
	record := AccuracyRecord{
		BusinessName:         known.BusinessName,
		ClassificationID:     fmt.Sprintf("val_%d", time.Now().UnixNano()),
		PredictedCode:        classification.IndustryCode,
		ActualCode:           known.IndustryCode,
		Accuracy:             accuracyScore,
		ValidationMethod:     result.ValidationMethod,
		Timestamp:            time.Now(),
		ConfidenceScore:      classification.ConfidenceScore,
		ClassificationMethod: classification.ClassificationMethod,
	}
	a.accuracyHistory = append(a.accuracyHistory, record)

	return result
}

// validateAgainstBenchmarks validates classification against industry benchmarks
func (a *AccuracyValidationEngine) validateAgainstBenchmarks(ctx context.Context, classification IndustryClassification) *ValidationResult {
	result := &ValidationResult{
		Timestamp: time.Now(),
	}

	// Get industry benchmark
	benchmark, exists := a.industryBenchmarks[classification.IndustryCode]

	if exists {
		// Calculate accuracy based on benchmark
		accuracyScore := a.calculateBenchmarkAccuracy(classification, benchmark)
		result.AccuracyScore = accuracyScore
		result.IsAccurate = accuracyScore >= a.accuracyThresholds["acceptable"]
		result.ConfidenceLevel = a.determineAccuracyLevel(accuracyScore)
		result.ValidationMethod = "benchmark_comparison"
	} else {
		// Use confidence score as proxy for accuracy
		result.AccuracyScore = classification.ConfidenceScore
		result.IsAccurate = classification.ConfidenceScore >= a.accuracyThresholds["acceptable"]
		result.ConfidenceLevel = a.determineAccuracyLevel(classification.ConfidenceScore)
		result.ValidationMethod = "confidence_based"
	}

	return result
}

// calculateAccuracyScore calculates accuracy score between predicted and actual
func (a *AccuracyValidationEngine) calculateAccuracyScore(classification IndustryClassification, known KnownClassification) float64 {
	// Exact match gets highest score
	if classification.IndustryCode == known.IndustryCode {
		return 1.0
	}

	// Check if same major category (first 2 digits)
	if len(classification.IndustryCode) >= 2 && len(known.IndustryCode) >= 2 {
		if classification.IndustryCode[:2] == known.IndustryCode[:2] {
			return 0.8
		}
	}

	// Check keyword similarity
	keywordSimilarity := a.calculateKeywordSimilarity(classification, known)

	// Base score on confidence and keyword similarity
	baseScore := (classification.ConfidenceScore + keywordSimilarity) / 2

	// Apply penalty for different industries
	return math.Max(0.0, baseScore-0.3)
}

// calculateKeywordSimilarity calculates similarity between classification and known keywords
func (a *AccuracyValidationEngine) calculateKeywordSimilarity(classification IndustryClassification, known KnownClassification) float64 {
	// This would compare keywords from the classification with expected keywords for the known industry
	// For now, return a default similarity score
	return 0.5
}

// calculateBenchmarkAccuracy calculates accuracy based on industry benchmarks
func (a *AccuracyValidationEngine) calculateBenchmarkAccuracy(classification IndustryClassification, benchmark IndustryBenchmark) float64 {
	// Base accuracy on confidence score relative to benchmark
	confidenceRatio := classification.ConfidenceScore / benchmark.TypicalConfidence

	// Adjust based on benchmark accuracy
	adjustedAccuracy := benchmark.AverageAccuracy * confidenceRatio

	// Ensure within reasonable bounds
	return math.Max(0.0, math.Min(1.0, adjustedAccuracy))
}

// determineAccuracyLevel determines the accuracy level based on score
func (a *AccuracyValidationEngine) determineAccuracyLevel(accuracyScore float64) string {
	switch {
	case accuracyScore >= a.accuracyThresholds["excellent"]:
		return "excellent"
	case accuracyScore >= a.accuracyThresholds["good"]:
		return "good"
	case accuracyScore >= a.accuracyThresholds["acceptable"]:
		return "acceptable"
	default:
		return "poor"
	}
}

// recordAccuracyMetrics records accuracy metrics for tracking
func (a *AccuracyValidationEngine) recordAccuracyMetrics(classification IndustryClassification, result *ValidationResult) {
	a.totalValidations++

	if result.IsAccurate {
		a.successfulValidations++
	}

	// Update average accuracy
	a.averageAccuracy = float64(a.successfulValidations) / float64(a.totalValidations)

	// Update industry-specific accuracy
	if current, exists := a.accuracyByIndustry[classification.IndustryCode]; exists {
		a.accuracyByIndustry[classification.IndustryCode] = (current + result.AccuracyScore) / 2
	} else {
		a.accuracyByIndustry[classification.IndustryCode] = result.AccuracyScore
	}

	// Update method-specific accuracy
	if current, exists := a.accuracyByMethod[classification.ClassificationMethod]; exists {
		a.accuracyByMethod[classification.ClassificationMethod] = (current + result.AccuracyScore) / 2
	} else {
		a.accuracyByMethod[classification.ClassificationMethod] = result.AccuracyScore
	}

	// Record metrics
	if a.metrics != nil {
		a.metrics.RecordBusinessClassification("accuracy_validation_total", "1")
		if result.IsAccurate {
			a.metrics.RecordBusinessClassification("accuracy_validation_success", "1")
		}
		a.metrics.RecordBusinessClassification("accuracy_score", fmt.Sprintf("%.3f", result.AccuracyScore))
	}
}

// generateFeedback generates feedback on the classification
func (a *AccuracyValidationEngine) generateFeedback(classification IndustryClassification, known KnownClassification) []string {
	var feedback []string

	// Apply feedback rules
	for _, rule := range a.feedbackRules {
		if rule.Generator != nil {
			message, weight := rule.Generator(classification, known)
			if weight > 0.5 { // Only include high-weight feedback
				feedback = append(feedback, message)
			}
		}
	}

	// Add default feedback if none generated
	if len(feedback) == 0 {
		feedback = append(feedback, "Classification completed successfully")
	}

	return feedback
}

// generateRecommendations generates recommendations for improvement
func (a *AccuracyValidationEngine) generateRecommendations(classification IndustryClassification, result *ValidationResult) []string {
	var recommendations []string

	if result.AccuracyScore < a.accuracyThresholds["good"] {
		recommendations = append(recommendations, "Consider manual review for low-confidence classifications")
		recommendations = append(recommendations, "Provide additional business context for improved accuracy")
	}

	if classification.ConfidenceScore < 0.7 {
		recommendations = append(recommendations, "Review classification method and consider alternative approaches")
	}

	if len(classification.Keywords) == 0 {
		recommendations = append(recommendations, "Include relevant keywords for better classification accuracy")
	}

	// Add default recommendation
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Classification accuracy is within acceptable range")
	}

	return recommendations
}

// calculateBenchmarkComparison calculates comparison against industry benchmarks
func (a *AccuracyValidationEngine) calculateBenchmarkComparison(industryCode string, currentAccuracy float64) *BenchmarkComparison {
	benchmark, exists := a.industryBenchmarks[industryCode]
	if !exists {
		// Return default benchmark comparison when no benchmark data exists
		return &BenchmarkComparison{
			IndustryBenchmark: 0.75, // Default benchmark
			CurrentAccuracy:   currentAccuracy,
			Difference:        currentAccuracy - 0.75,
			Percentile:        50.0, // Default to median
			IsAboveBenchmark:  currentAccuracy > 0.75,
			BenchmarkSource:   "default_benchmark",
		}
	}

	difference := currentAccuracy - benchmark.AverageAccuracy
	percentile := a.calculatePercentile(currentAccuracy, industryCode)

	return &BenchmarkComparison{
		IndustryBenchmark: benchmark.AverageAccuracy,
		CurrentAccuracy:   currentAccuracy,
		Difference:        difference,
		Percentile:        percentile,
		IsAboveBenchmark:  difference > 0,
		BenchmarkSource:   "industry_benchmark",
	}
}

// calculateAccuracyTrend calculates historical accuracy trend for an industry
func (a *AccuracyValidationEngine) calculateAccuracyTrend(industryCode string) *AccuracyTrend {
	// Filter accuracy history for this industry
	var industryRecords []AccuracyRecord
	for _, record := range a.accuracyHistory {
		if record.PredictedCode == industryCode {
			industryRecords = append(industryRecords, record)
		}
	}

	if len(industryRecords) < 2 {
		return &AccuracyTrend{
			PeriodDays:      30,
			AverageAccuracy: 0.0,
			TrendDirection:  "insufficient_data",
			TrendStrength:   0.0,
			DataPoints:      len(industryRecords),
		}
	}

	// Calculate trend
	totalAccuracy := 0.0
	for _, record := range industryRecords {
		totalAccuracy += record.Accuracy
	}
	averageAccuracy := totalAccuracy / float64(len(industryRecords))

	// Simple trend calculation (could be enhanced with linear regression)
	trendDirection := "stable"
	trendStrength := 0.0

	if len(industryRecords) >= 10 {
		// Calculate trend over last 10 records
		recent := industryRecords[len(industryRecords)-10:]
		older := industryRecords[:10]

		recentAvg := 0.0
		for _, r := range recent {
			recentAvg += r.Accuracy
		}
		recentAvg /= float64(len(recent))

		olderAvg := 0.0
		for _, r := range older {
			olderAvg += r.Accuracy
		}
		olderAvg /= float64(len(older))

		trendStrength = recentAvg - olderAvg

		if trendStrength > 0.05 {
			trendDirection = "improving"
		} else if trendStrength < -0.05 {
			trendDirection = "declining"
		}
	}

	return &AccuracyTrend{
		PeriodDays:      30,
		AverageAccuracy: averageAccuracy,
		TrendDirection:  trendDirection,
		TrendStrength:   math.Abs(trendStrength),
		DataPoints:      len(industryRecords),
	}
}

// calculatePercentile calculates the percentile of current accuracy within historical data
func (a *AccuracyValidationEngine) calculatePercentile(currentAccuracy float64, industryCode string) float64 {
	var accuracies []float64
	for _, record := range a.accuracyHistory {
		if record.PredictedCode == industryCode {
			accuracies = append(accuracies, record.Accuracy)
		}
	}

	if len(accuracies) == 0 {
		return 50.0 // Default to median if no data
	}

	// Count how many are below current accuracy
	belowCount := 0
	for _, acc := range accuracies {
		if acc < currentAccuracy {
			belowCount++
		}
	}

	return float64(belowCount) / float64(len(accuracies)) * 100.0
}

// AddKnownClassification adds a known classification for validation
func (a *AccuracyValidationEngine) AddKnownClassification(known KnownClassification) {
	a.knownClassifications[known.BusinessName] = known
}

// AddIndustryBenchmark adds an industry benchmark for validation
func (a *AccuracyValidationEngine) AddIndustryBenchmark(benchmark IndustryBenchmark) {
	a.industryBenchmarks[benchmark.IndustryCode] = benchmark
}

// GetAccuracyMetrics returns current accuracy metrics
func (a *AccuracyValidationEngine) GetAccuracyMetrics() map[string]interface{} {
	return map[string]interface{}{
		"total_validations":      a.totalValidations,
		"successful_validations": a.successfulValidations,
		"average_accuracy":       a.averageAccuracy,
		"accuracy_by_industry":   a.accuracyByIndustry,
		"accuracy_by_method":     a.accuracyByMethod,
	}
}

// initializeValidationRules initializes the validation rules
func (a *AccuracyValidationEngine) initializeValidationRules() {
	a.validationRules["exact_match"] = ValidationRule{
		Name:        "Exact Industry Code Match",
		Description: "Validates exact match of industry codes",
		Weight:      1.0,
		Validator: func(classification IndustryClassification, known KnownClassification) (bool, float64, string) {
			return classification.IndustryCode == known.IndustryCode, 1.0, "exact_match"
		},
	}

	a.validationRules["major_category"] = ValidationRule{
		Name:        "Major Category Match",
		Description: "Validates match within major industry category",
		Weight:      0.8,
		Validator: func(classification IndustryClassification, known KnownClassification) (bool, float64, string) {
			if len(classification.IndustryCode) >= 2 && len(known.IndustryCode) >= 2 {
				return classification.IndustryCode[:2] == known.IndustryCode[:2], 0.8, "major_category_match"
			}
			return false, 0.0, "no_match"
		},
	}

	a.validationRules["confidence_threshold"] = ValidationRule{
		Name:        "Confidence Threshold",
		Description: "Validates confidence score meets minimum threshold",
		Weight:      0.6,
		Validator: func(classification IndustryClassification, known KnownClassification) (bool, float64, string) {
			return classification.ConfidenceScore >= 0.7, classification.ConfidenceScore, "confidence_threshold"
		},
	}
}

// initializeFeedbackRules initializes the feedback rules
func (a *AccuracyValidationEngine) initializeFeedbackRules() {
	a.feedbackRules["high_confidence"] = FeedbackRule{
		Name:        "High Confidence Classification",
		Description: "Provides feedback for high-confidence classifications",
		Weight:      0.9,
		Generator: func(classification IndustryClassification, known KnownClassification) (string, float64) {
			if classification.ConfidenceScore >= 0.9 {
				return "High-confidence classification with strong evidence", 0.9
			}
			return "", 0.0
		},
	}

	a.feedbackRules["keyword_evidence"] = FeedbackRule{
		Name:        "Keyword Evidence",
		Description: "Provides feedback based on keyword evidence",
		Weight:      0.7,
		Generator: func(classification IndustryClassification, known KnownClassification) (string, float64) {
			if len(classification.Keywords) > 0 {
				return "Classification supported by keyword evidence", 0.7
			}
			return "Limited keyword evidence available", 0.3
		},
	}

	a.feedbackRules["method_quality"] = FeedbackRule{
		Name:        "Classification Method Quality",
		Description: "Provides feedback based on classification method",
		Weight:      0.8,
		Generator: func(classification IndustryClassification, known KnownClassification) (string, float64) {
			switch classification.ClassificationMethod {
			case "keyword_match":
				return "Direct keyword match method used", 0.8
			case "description_match":
				return "Description-based classification method used", 0.7
			case "fuzzy_match":
				return "Fuzzy matching method used - consider manual review", 0.5
			default:
				return "Standard classification method used", 0.6
			}
		},
	}
}
