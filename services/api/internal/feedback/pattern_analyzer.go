package feedback

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// FeedbackPatternAnalyzer analyzes patterns in feedback data
type FeedbackPatternAnalyzer struct {
	config *MLAnalysisConfig
	logger *zap.Logger
}

// NewFeedbackPatternAnalyzer creates a new feedback pattern analyzer
func NewFeedbackPatternAnalyzer(config *MLAnalysisConfig, logger *zap.Logger) *FeedbackPatternAnalyzer {
	return &FeedbackPatternAnalyzer{
		config: config,
		logger: logger,
	}
}

// AnalyzeMethodPatterns analyzes patterns for a specific classification method
func (fpa *FeedbackPatternAnalyzer) AnalyzeMethodPatterns(ctx context.Context, method ClassificationMethod, feedback []*UserFeedback) ([]*FeedbackPattern, error) {
	if len(feedback) < fpa.config.MinFeedbackThreshold {
		return []*FeedbackPattern{}, nil
	}

	var patterns []*FeedbackPattern

	// Analyze temporal patterns
	temporalPatterns := fpa.analyzeTemporalPatterns(method, feedback)
	patterns = append(patterns, temporalPatterns...)

	// Analyze confidence patterns
	confidencePatterns := fpa.analyzeConfidencePatterns(method, feedback)
	patterns = append(patterns, confidencePatterns...)

	// Analyze accuracy patterns
	accuracyPatterns := fpa.analyzeAccuracyPatterns(method, feedback)
	patterns = append(patterns, accuracyPatterns...)

	// Analyze error patterns
	errorPatterns := fpa.analyzeErrorPatterns(method, feedback)
	patterns = append(patterns, errorPatterns...)

	return patterns, nil
}

// AnalyzeCrossMethodPatterns analyzes patterns across multiple methods
func (fpa *FeedbackPatternAnalyzer) AnalyzeCrossMethodPatterns(ctx context.Context, methodFeedback map[ClassificationMethod][]*UserFeedback) ([]*FeedbackPattern, error) {
	var patterns []*FeedbackPattern

	// Analyze method comparison patterns
	comparisonPatterns := fpa.analyzeMethodComparisonPatterns(methodFeedback)
	patterns = append(patterns, comparisonPatterns...)

	// Analyze ensemble disagreement patterns
	disagreementPatterns := fpa.analyzeEnsembleDisagreementPatterns(methodFeedback)
	patterns = append(patterns, disagreementPatterns...)

	// Analyze performance correlation patterns
	correlationPatterns := fpa.analyzePerformanceCorrelationPatterns(methodFeedback)
	patterns = append(patterns, correlationPatterns...)

	return patterns, nil
}

// analyzeTemporalPatterns analyzes temporal patterns in feedback
func (fpa *FeedbackPatternAnalyzer) analyzeTemporalPatterns(method ClassificationMethod, feedback []*UserFeedback) []*FeedbackPattern {
	var patterns []*FeedbackPattern

	// Group feedback by time periods
	hourlyFeedback := make(map[int]int)
	dailyFeedback := make(map[string]int)

	for _, fb := range feedback {
		hour := fb.CreatedAt.Hour()
		day := fb.CreatedAt.Format("2006-01-02")

		hourlyFeedback[hour]++
		dailyFeedback[day]++
	}

	// Detect peak hours
	var peakHours []int
	maxHourlyCount := 0
	for hour, count := range hourlyFeedback {
		if count > maxHourlyCount {
			maxHourlyCount = count
			peakHours = []int{hour}
		} else if count == maxHourlyCount {
			peakHours = append(peakHours, hour)
		}
	}

	if len(peakHours) > 0 && maxHourlyCount > len(feedback)/4 {
		pattern := &FeedbackPattern{
			PatternID:          fmt.Sprintf("temporal_peak_%s", method),
			PatternType:        "temporal_peak",
			PatternDescription: fmt.Sprintf("Peak feedback activity during hours: %v", peakHours),
			Confidence:         0.8,
			OccurrenceCount:    maxHourlyCount,
			AffectedMethods:    []ClassificationMethod{method},
			TimeWindow:         fpa.config.AnalysisWindowSize,
			Severity:           "medium",
			Trend:              "stable",
			Metadata: map[string]interface{}{
				"peak_hours": peakHours,
				"max_count":  maxHourlyCount,
			},
		}
		patterns = append(patterns, pattern)
	}

	// Detect daily trends
	if len(dailyFeedback) >= 3 {
		var dailyCounts []int
		for _, count := range dailyFeedback {
			dailyCounts = append(dailyCounts, count)
		}

		trend := fpa.calculateTrend(dailyCounts)
		if trend != "stable" {
			pattern := &FeedbackPattern{
				PatternID:          fmt.Sprintf("daily_trend_%s", method),
				PatternType:        "daily_trend",
				PatternDescription: fmt.Sprintf("Daily feedback trend: %s", trend),
				Confidence:         0.7,
				OccurrenceCount:    len(feedback),
				AffectedMethods:    []ClassificationMethod{method},
				TimeWindow:         fpa.config.AnalysisWindowSize,
				Severity:           "low",
				Trend:              trend,
				Metadata: map[string]interface{}{
					"daily_counts": dailyCounts,
					"trend":        trend,
				},
			}
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

// analyzeConfidencePatterns analyzes confidence-related patterns
func (fpa *FeedbackPatternAnalyzer) analyzeConfidencePatterns(method ClassificationMethod, feedback []*UserFeedback) []*FeedbackPattern {
	var patterns []*FeedbackPattern

	// Analyze confidence distribution
	var confidenceScores []float64
	var lowConfidenceCount, highConfidenceCount int

	for _, fb := range feedback {
		if fb.ConfidenceScore > 0 {
			confidenceScores = append(confidenceScores, fb.ConfidenceScore)
			if fb.ConfidenceScore < 0.5 {
				lowConfidenceCount++
			} else if fb.ConfidenceScore > 0.8 {
				highConfidenceCount++
			}
		}
	}

	if len(confidenceScores) == 0 {
		return patterns
	}

	// Calculate average confidence
	var totalConfidence float64
	for _, score := range confidenceScores {
		totalConfidence += score
	}
	avgConfidence := totalConfidence / float64(len(confidenceScores))

	// Detect low confidence pattern
	if lowConfidenceCount > len(confidenceScores)/3 {
		pattern := &FeedbackPattern{
			PatternID:          fmt.Sprintf("low_confidence_%s", method),
			PatternType:        "confidence_low",
			PatternDescription: fmt.Sprintf("High frequency of low confidence scores (%.2f%% below 0.5)", float64(lowConfidenceCount)/float64(len(confidenceScores))*100),
			Confidence:         0.8,
			OccurrenceCount:    lowConfidenceCount,
			AffectedMethods:    []ClassificationMethod{method},
			TimeWindow:         fpa.config.AnalysisWindowSize,
			Severity:           "high",
			Trend:              "stable",
			Metadata: map[string]interface{}{
				"low_confidence_count":   lowConfidenceCount,
				"total_confidence_count": len(confidenceScores),
				"average_confidence":     avgConfidence,
			},
		}
		patterns = append(patterns, pattern)
	}

	// Detect confidence calibration issues
	if avgConfidence < 0.4 || avgConfidence > 0.9 {
		severity := "medium"
		if avgConfidence < 0.3 || avgConfidence > 0.95 {
			severity = "high"
		}

		pattern := &FeedbackPattern{
			PatternID:          fmt.Sprintf("confidence_calibration_%s", method),
			PatternType:        "confidence_calibration",
			PatternDescription: fmt.Sprintf("Confidence calibration issue: average confidence %.2f", avgConfidence),
			Confidence:         0.9,
			OccurrenceCount:    len(confidenceScores),
			AffectedMethods:    []ClassificationMethod{method},
			TimeWindow:         fpa.config.AnalysisWindowSize,
			Severity:           severity,
			Trend:              "stable",
			Metadata: map[string]interface{}{
				"average_confidence": avgConfidence,
				"confidence_scores":  confidenceScores,
			},
		}
		patterns = append(patterns, pattern)
	}

	return patterns
}

// analyzeAccuracyPatterns analyzes accuracy-related patterns
func (fpa *FeedbackPatternAnalyzer) analyzeAccuracyPatterns(method ClassificationMethod, feedback []*UserFeedback) []*FeedbackPattern {
	var patterns []*FeedbackPattern

	// Group feedback by accuracy type
	accuracyFeedback := make(map[string][]*UserFeedback)
	correctionFeedback := make([]*UserFeedback, 0)

	for _, fb := range feedback {
		if fb.FeedbackType == FeedbackTypeAccuracy {
			if accuracy, ok := fb.FeedbackValue["accuracy"].(string); ok {
				accuracyFeedback[accuracy] = append(accuracyFeedback[accuracy], fb)
			}
		} else if fb.FeedbackType == FeedbackTypeCorrection {
			correctionFeedback = append(correctionFeedback, fb)
		}
	}

	// Detect high correction rate
	if len(correctionFeedback) > len(feedback)/4 {
		pattern := &FeedbackPattern{
			PatternID:          fmt.Sprintf("high_corrections_%s", method),
			PatternType:        "high_corrections",
			PatternDescription: fmt.Sprintf("High correction rate: %.2f%% of feedback are corrections", float64(len(correctionFeedback))/float64(len(feedback))*100),
			Confidence:         0.9,
			OccurrenceCount:    len(correctionFeedback),
			AffectedMethods:    []ClassificationMethod{method},
			TimeWindow:         fpa.config.AnalysisWindowSize,
			Severity:           "high",
			Trend:              "stable",
			Metadata: map[string]interface{}{
				"correction_count": len(correctionFeedback),
				"total_feedback":   len(feedback),
				"correction_rate":  float64(len(correctionFeedback)) / float64(len(feedback)),
			},
		}
		patterns = append(patterns, pattern)
	}

	// Detect accuracy trends
	if len(accuracyFeedback) > 0 {
		var accuracyRates []float64
		for accuracy, fbs := range accuracyFeedback {
			if accuracy == "correct" {
				accuracyRates = append(accuracyRates, float64(len(fbs))/float64(len(feedback)))
			}
		}

		if len(accuracyRates) > 0 {
			var totalAccuracy float64
			for _, rate := range accuracyRates {
				totalAccuracy += rate
			}
			avgAccuracy := totalAccuracy / float64(len(accuracyRates))

			if avgAccuracy < 0.7 {
				pattern := &FeedbackPattern{
					PatternID:          fmt.Sprintf("low_accuracy_%s", method),
					PatternType:        "low_accuracy",
					PatternDescription: fmt.Sprintf("Low accuracy rate: %.2f%%", avgAccuracy*100),
					Confidence:         0.9,
					OccurrenceCount:    len(feedback),
					AffectedMethods:    []ClassificationMethod{method},
					TimeWindow:         fpa.config.AnalysisWindowSize,
					Severity:           "high",
					Trend:              "stable",
					Metadata: map[string]interface{}{
						"accuracy_rate":           avgAccuracy,
						"accuracy_feedback_count": len(accuracyFeedback),
					},
				}
				patterns = append(patterns, pattern)
			}
		}
	}

	return patterns
}

// analyzeErrorPatterns analyzes error-related patterns
func (fpa *FeedbackPatternAnalyzer) analyzeErrorPatterns(method ClassificationMethod, feedback []*UserFeedback) []*FeedbackPattern {
	var patterns []*FeedbackPattern

	// Analyze error types
	errorTypes := make(map[string]int)
	var errorFeedback []*UserFeedback

	for _, fb := range feedback {
		if fb.FeedbackType == FeedbackTypeCorrection ||
			strings.Contains(strings.ToLower(fb.FeedbackText), "error") ||
			strings.Contains(strings.ToLower(fb.FeedbackText), "wrong") {
			errorFeedback = append(errorFeedback, fb)

			// Categorize error types based on feedback text
			errorType := fpa.categorizeErrorType(fb.FeedbackText)
			errorTypes[errorType]++
		}
	}

	// Detect common error patterns
	for errorType, count := range errorTypes {
		if count > len(errorFeedback)/3 {
			pattern := &FeedbackPattern{
				PatternID:          fmt.Sprintf("common_error_%s_%s", method, errorType),
				PatternType:        "common_error",
				PatternDescription: fmt.Sprintf("Common error type '%s': %d occurrences", errorType, count),
				Confidence:         0.8,
				OccurrenceCount:    count,
				AffectedMethods:    []ClassificationMethod{method},
				TimeWindow:         fpa.config.AnalysisWindowSize,
				Severity:           "medium",
				Trend:              "stable",
				Metadata: map[string]interface{}{
					"error_type":   errorType,
					"error_count":  count,
					"total_errors": len(errorFeedback),
				},
			}
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

// analyzeMethodComparisonPatterns analyzes patterns when comparing methods
func (fpa *FeedbackPatternAnalyzer) analyzeMethodComparisonPatterns(methodFeedback map[ClassificationMethod][]*UserFeedback) []*FeedbackPattern {
	var patterns []*FeedbackPattern

	methods := make([]ClassificationMethod, 0, len(methodFeedback))
	for method := range methodFeedback {
		methods = append(methods, method)
	}

	if len(methods) < 2 {
		return patterns
	}

	// Compare method performance
	for i := 0; i < len(methods); i++ {
		for j := i + 1; j < len(methods); j++ {
			method1, method2 := methods[i], methods[j]
			fb1, fb2 := methodFeedback[method1], methodFeedback[method2]

			if len(fb1) < fpa.config.MinFeedbackThreshold || len(fb2) < fpa.config.MinFeedbackThreshold {
				continue
			}

			// Calculate performance metrics for each method
			perf1 := fpa.calculateMethodPerformance(fb1)
			perf2 := fpa.calculateMethodPerformance(fb2)

			// Detect significant performance differences
			if perf1.Accuracy-perf2.Accuracy > 0.2 {
				pattern := &FeedbackPattern{
					PatternID:          fmt.Sprintf("performance_gap_%s_vs_%s", method1, method2),
					PatternType:        "performance_gap",
					PatternDescription: fmt.Sprintf("Significant performance gap: %s (%.2f%%) vs %s (%.2f%%)", method1, perf1.Accuracy*100, method2, perf2.Accuracy*100),
					Confidence:         0.8,
					OccurrenceCount:    len(fb1) + len(fb2),
					AffectedMethods:    []ClassificationMethod{method1, method2},
					TimeWindow:         fpa.config.AnalysisWindowSize,
					Severity:           "medium",
					Trend:              "stable",
					Metadata: map[string]interface{}{
						"method1":          method1,
						"method2":          method2,
						"method1_accuracy": perf1.Accuracy,
						"method2_accuracy": perf2.Accuracy,
						"accuracy_gap":     perf1.Accuracy - perf2.Accuracy,
					},
				}
				patterns = append(patterns, pattern)
			}
		}
	}

	return patterns
}

// analyzeEnsembleDisagreementPatterns analyzes patterns of ensemble disagreements
func (fpa *FeedbackPatternAnalyzer) analyzeEnsembleDisagreementPatterns(methodFeedback map[ClassificationMethod][]*UserFeedback) []*FeedbackPattern {
	var patterns []*FeedbackPattern

	// This would require access to ensemble results to detect disagreements
	// For now, we'll analyze based on feedback patterns that suggest disagreements

	// Look for feedback that mentions multiple methods or conflicting results
	var disagreementFeedback []*UserFeedback
	for _, feedbacks := range methodFeedback {
		for _, fb := range feedbacks {
			if strings.Contains(strings.ToLower(fb.FeedbackText), "disagree") ||
				strings.Contains(strings.ToLower(fb.FeedbackText), "conflict") ||
				strings.Contains(strings.ToLower(fb.FeedbackText), "different") {
				disagreementFeedback = append(disagreementFeedback, fb)
			}
		}
	}

	if len(disagreementFeedback) > len(methodFeedback)*2 {
		pattern := &FeedbackPattern{
			PatternID:          "ensemble_disagreement",
			PatternType:        "ensemble_disagreement",
			PatternDescription: fmt.Sprintf("High ensemble disagreement rate: %d feedback items mention disagreements", len(disagreementFeedback)),
			Confidence:         0.7,
			OccurrenceCount:    len(disagreementFeedback),
			AffectedMethods:    []ClassificationMethod{MethodEnsemble},
			TimeWindow:         fpa.config.AnalysisWindowSize,
			Severity:           "medium",
			Trend:              "stable",
			Metadata: map[string]interface{}{
				"disagreement_count": len(disagreementFeedback),
				"total_methods":      len(methodFeedback),
			},
		}
		patterns = append(patterns, pattern)
	}

	return patterns
}

// analyzePerformanceCorrelationPatterns analyzes correlation patterns between methods
func (fpa *FeedbackPatternAnalyzer) analyzePerformanceCorrelationPatterns(methodFeedback map[ClassificationMethod][]*UserFeedback) []*FeedbackPattern {
	var patterns []*FeedbackPattern

	// This would require more sophisticated correlation analysis
	// For now, we'll implement basic correlation detection

	methods := make([]ClassificationMethod, 0, len(methodFeedback))
	for method := range methodFeedback {
		methods = append(methods, method)
	}

	if len(methods) < 2 {
		return patterns
	}

	// Calculate performance correlation between methods
	for i := 0; i < len(methods); i++ {
		for j := i + 1; j < len(methods); j++ {
			method1, method2 := methods[i], methods[j]
			fb1, fb2 := methodFeedback[method1], methodFeedback[method2]

			if len(fb1) < fpa.config.MinFeedbackThreshold || len(fb2) < fpa.config.MinFeedbackThreshold {
				continue
			}

			correlation := fpa.calculatePerformanceCorrelation(fb1, fb2)

			// Detect strong positive or negative correlation
			if correlation > 0.8 || correlation < -0.8 {
				patternType := "positive_correlation"
				if correlation < -0.8 {
					patternType = "negative_correlation"
				}

				pattern := &FeedbackPattern{
					PatternID:          fmt.Sprintf("correlation_%s_vs_%s", method1, method2),
					PatternType:        patternType,
					PatternDescription: fmt.Sprintf("Strong %s between %s and %s: %.2f", patternType, method1, method2, correlation),
					Confidence:         0.8,
					OccurrenceCount:    len(fb1) + len(fb2),
					AffectedMethods:    []ClassificationMethod{method1, method2},
					TimeWindow:         fpa.config.AnalysisWindowSize,
					Severity:           "low",
					Trend:              "stable",
					Metadata: map[string]interface{}{
						"method1":          method1,
						"method2":          method2,
						"correlation":      correlation,
						"correlation_type": patternType,
					},
				}
				patterns = append(patterns, pattern)
			}
		}
	}

	return patterns
}

// Helper methods

// calculateTrend calculates the trend direction from a series of values
func (fpa *FeedbackPatternAnalyzer) calculateTrend(values []int) string {
	if len(values) < 2 {
		return "stable"
	}

	// Simple trend calculation
	increasing := 0
	decreasing := 0

	for i := 1; i < len(values); i++ {
		if values[i] > values[i-1] {
			increasing++
		} else if values[i] < values[i-1] {
			decreasing++
		}
	}

	if increasing > decreasing*2 {
		return "increasing"
	} else if decreasing > increasing*2 {
		return "decreasing"
	}
	return "stable"
}

// categorizeErrorType categorizes error types based on feedback text
func (fpa *FeedbackPatternAnalyzer) categorizeErrorType(feedbackText string) string {
	text := strings.ToLower(feedbackText)

	if strings.Contains(text, "industry") || strings.Contains(text, "category") {
		return "industry_classification"
	} else if strings.Contains(text, "confidence") || strings.Contains(text, "score") {
		return "confidence_scoring"
	} else if strings.Contains(text, "keyword") || strings.Contains(text, "match") {
		return "keyword_matching"
	} else if strings.Contains(text, "ml") || strings.Contains(text, "model") {
		return "ml_model"
	} else if strings.Contains(text, "security") || strings.Contains(text, "trust") {
		return "security_validation"
	}
	return "general_error"
}

// calculateMethodPerformance calculates performance metrics for a method
func (fpa *FeedbackPatternAnalyzer) calculateMethodPerformance(feedback []*UserFeedback) *MethodPerformance {
	perf := &MethodPerformance{
		Accuracy:   0.0,
		Confidence: 0.0,
		ErrorRate:  0.0,
		TotalCount: len(feedback),
	}

	if len(feedback) == 0 {
		return perf
	}

	var totalAccuracy, totalConfidence float64
	var accuracyCount, confidenceCount, errorCount int

	for _, fb := range feedback {
		if fb.FeedbackType == FeedbackTypeAccuracy {
			if accuracy, ok := fb.FeedbackValue["accuracy"].(string); ok && accuracy == "correct" {
				totalAccuracy += 1.0
			}
			accuracyCount++
		}

		if fb.ConfidenceScore > 0 {
			totalConfidence += fb.ConfidenceScore
			confidenceCount++
		}

		if fb.FeedbackType == FeedbackTypeCorrection {
			errorCount++
		}
	}

	if accuracyCount > 0 {
		perf.Accuracy = totalAccuracy / float64(accuracyCount)
	}
	if confidenceCount > 0 {
		perf.Confidence = totalConfidence / float64(confidenceCount)
	}
	perf.ErrorRate = float64(errorCount) / float64(len(feedback))

	return perf
}

// calculatePerformanceCorrelation calculates correlation between two methods' performance
func (fpa *FeedbackPatternAnalyzer) calculatePerformanceCorrelation(fb1, fb2 []*UserFeedback) float64 {
	// This is a simplified correlation calculation
	// In a real implementation, you'd want to match feedback by business or time window

	perf1 := fpa.calculateMethodPerformance(fb1)
	perf2 := fpa.calculateMethodPerformance(fb2)

	// Simple correlation based on accuracy
	// In practice, you'd want more sophisticated correlation analysis
	if perf1.Accuracy > 0.8 && perf2.Accuracy > 0.8 {
		return 0.9
	} else if perf1.Accuracy < 0.5 && perf2.Accuracy < 0.5 {
		return 0.8
	} else if (perf1.Accuracy > 0.8 && perf2.Accuracy < 0.5) || (perf1.Accuracy < 0.5 && perf2.Accuracy > 0.8) {
		return -0.7
	}

	return 0.0
}

// MethodPerformance represents performance metrics for a classification method
type MethodPerformance struct {
	Accuracy   float64 `json:"accuracy"`
	Confidence float64 `json:"confidence"`
	ErrorRate  float64 `json:"error_rate"`
	TotalCount int     `json:"total_count"`
}
