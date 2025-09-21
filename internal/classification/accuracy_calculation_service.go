package classification

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"
)

// AccuracyCalculationService provides comprehensive accuracy calculation and analysis
type AccuracyCalculationService struct {
	db     *sql.DB
	logger *log.Logger
}

// NewAccuracyCalculationService creates a new accuracy calculation service
func NewAccuracyCalculationService(db *sql.DB, logger *log.Logger) *AccuracyCalculationService {
	if logger == nil {
		logger = log.Default()
	}

	return &AccuracyCalculationService{
		db:     db,
		logger: logger,
	}
}

// AccuracyCalculationResult represents the result of accuracy calculation
type AccuracyCalculationResult struct {
	OverallAccuracy          float64                    `json:"overall_accuracy"`
	IndustrySpecificAccuracy map[string]float64         `json:"industry_specific_accuracy"`
	ConfidenceDistribution   ConfidenceDistribution     `json:"confidence_distribution"`
	SecurityMetrics          SecurityAccuracyMetrics    `json:"security_metrics"`
	PerformanceMetrics       PerformanceAccuracyMetrics `json:"performance_metrics"`
	CalculationTimestamp     time.Time                  `json:"calculation_timestamp"`
	DataPointsAnalyzed       int64                      `json:"data_points_analyzed"`
	TimeRangeAnalyzed        string                     `json:"time_range_analyzed"`
}

// ConfidenceDistribution represents the distribution of confidence scores
type ConfidenceDistribution struct {
	HighConfidence    int64                     `json:"high_confidence"`   // >0.8
	MediumConfidence  int64                     `json:"medium_confidence"` // 0.5-0.8
	LowConfidence     int64                     `json:"low_confidence"`    // <0.5
	AverageConfidence float64                   `json:"average_confidence"`
	ConfidenceRanges  []AccuracyConfidenceRange `json:"confidence_ranges"`
}

// AccuracyConfidenceRange represents a specific confidence range for accuracy calculation
type AccuracyConfidenceRange struct {
	RangeStart float64 `json:"range_start"`
	RangeEnd   float64 `json:"range_end"`
	Count      int64   `json:"count"`
	Accuracy   float64 `json:"accuracy"`
	Percentage float64 `json:"percentage"`
}

// SecurityAccuracyMetrics represents security-related accuracy metrics
type SecurityAccuracyMetrics struct {
	TrustedDataSourceAccuracy   float64 `json:"trusted_data_source_accuracy"`
	WebsiteVerificationAccuracy float64 `json:"website_verification_accuracy"`
	SecurityViolationRate       float64 `json:"security_violation_rate"`
	DataSourceTrustRate         float64 `json:"data_source_trust_rate"`
	WebsiteVerificationRate     float64 `json:"website_verification_rate"`
	TrustedDataPoints           int64   `json:"trusted_data_points"`
	VerifiedWebsitePoints       int64   `json:"verified_website_points"`
	TotalSecurityValidations    int64   `json:"total_security_validations"`
}

// PerformanceAccuracyMetrics represents performance-related accuracy metrics
type PerformanceAccuracyMetrics struct {
	AverageResponseTimeMs     float64            `json:"average_response_time_ms"`
	AverageProcessingTimeMs   float64            `json:"average_processing_time_ms"`
	HighPerformanceAccuracy   float64            `json:"high_performance_accuracy"`   // <200ms
	MediumPerformanceAccuracy float64            `json:"medium_performance_accuracy"` // 200-500ms
	LowPerformanceAccuracy    float64            `json:"low_performance_accuracy"`    // >500ms
	PerformanceRanges         []PerformanceRange `json:"performance_ranges"`
}

// PerformanceRange represents a specific performance range
type PerformanceRange struct {
	RangeStart float64 `json:"range_start"`
	RangeEnd   float64 `json:"range_end"`
	Count      int64   `json:"count"`
	Accuracy   float64 `json:"accuracy"`
	Percentage float64 `json:"percentage"`
}

// IndustryAccuracyBreakdown represents detailed accuracy breakdown by industry
type IndustryAccuracyBreakdown struct {
	IndustryName           string   `json:"industry_name"`
	TotalClassifications   int64    `json:"total_classifications"`
	CorrectClassifications int64    `json:"correct_classifications"`
	AccuracyPercentage     float64  `json:"accuracy_percentage"`
	AverageConfidence      float64  `json:"average_confidence"`
	AverageResponseTime    float64  `json:"average_response_time"`
	CommonErrors           []string `json:"common_errors"`
	ImprovementSuggestions []string `json:"improvement_suggestions"`
}

// CalculateOverallAccuracy calculates the overall accuracy rate
func (acs *AccuracyCalculationService) CalculateOverallAccuracy(ctx context.Context, hoursBack int) (float64, error) {
	query := `
		SELECT 
			COUNT(*) as total_classifications,
			COUNT(CASE WHEN metadata->>'is_correct' = 'true' THEN 1 END) as correct_classifications
		FROM unified_performance_metrics 
		WHERE component = 'classification' 
		AND metric_category = 'classification'
		AND created_at >= NOW() - INTERVAL '%d hours'
		AND metadata->>'actual_industry' IS NOT NULL
	`

	var totalClassifications, correctClassifications int64
	err := acs.db.QueryRowContext(ctx, fmt.Sprintf(query, hoursBack)).Scan(
		&totalClassifications,
		&correctClassifications,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to calculate overall accuracy: %w", err)
	}

	if totalClassifications == 0 {
		return 0, nil
	}

	accuracy := float64(correctClassifications) / float64(totalClassifications)
	acs.logger.Printf("📊 Overall accuracy calculated: %.2f%% (%d/%d)", accuracy*100, correctClassifications, totalClassifications)

	return accuracy, nil
}

// CalculateIndustrySpecificAccuracy calculates accuracy for each industry
func (acs *AccuracyCalculationService) CalculateIndustrySpecificAccuracy(ctx context.Context, hoursBack int) (map[string]float64, error) {
	query := `
		SELECT 
			predicted_industry,
			COUNT(*) as total_classifications,
			COUNT(CASE WHEN is_correct = true THEN 1 END) as correct_classifications
		FROM classification_accuracy_metrics 
		WHERE created_at >= NOW() - INTERVAL '%d hours'
		AND actual_industry IS NOT NULL
		GROUP BY predicted_industry
		ORDER BY total_classifications DESC
	`

	rows, err := acs.db.QueryContext(ctx, fmt.Sprintf(query, hoursBack))
	if err != nil {
		return nil, fmt.Errorf("failed to calculate industry-specific accuracy: %w", err)
	}
	defer rows.Close()

	industryAccuracy := make(map[string]float64)

	for rows.Next() {
		var industry string
		var totalClassifications, correctClassifications int64

		err := rows.Scan(&industry, &totalClassifications, &correctClassifications)
		if err != nil {
			return nil, fmt.Errorf("failed to scan industry accuracy: %w", err)
		}

		if totalClassifications > 0 {
			accuracy := float64(correctClassifications) / float64(totalClassifications)
			industryAccuracy[industry] = accuracy
			acs.logger.Printf("📊 Industry accuracy [%s]: %.2f%% (%d/%d)",
				industry, accuracy*100, correctClassifications, totalClassifications)
		}
	}

	return industryAccuracy, nil
}

// CalculateConfidenceDistribution calculates the distribution of confidence scores
func (acs *AccuracyCalculationService) CalculateConfidenceDistribution(ctx context.Context, hoursBack int) (*ConfidenceDistribution, error) {
	query := `
		SELECT 
			predicted_confidence,
			COUNT(*) as count,
			COUNT(CASE WHEN is_correct = true THEN 1 END) as correct_count
		FROM classification_accuracy_metrics 
		WHERE created_at >= NOW() - INTERVAL '%d hours'
		AND predicted_confidence IS NOT NULL
		GROUP BY predicted_confidence
		ORDER BY predicted_confidence
	`

	rows, err := acs.db.QueryContext(ctx, fmt.Sprintf(query, hoursBack))
	if err != nil {
		return nil, fmt.Errorf("failed to calculate confidence distribution: %w", err)
	}
	defer rows.Close()

	distribution := &ConfidenceDistribution{
		ConfidenceRanges: make([]AccuracyConfidenceRange, 0),
	}

	var totalCount, totalCorrect int64
	var confidenceSum float64

	// Define confidence ranges
	ranges := []struct {
		start, end float64
		name       string
	}{
		{0.0, 0.1, "0.0-0.1"},
		{0.1, 0.2, "0.1-0.2"},
		{0.2, 0.3, "0.2-0.3"},
		{0.3, 0.4, "0.3-0.4"},
		{0.4, 0.5, "0.4-0.5"},
		{0.5, 0.6, "0.5-0.6"},
		{0.6, 0.7, "0.6-0.7"},
		{0.7, 0.8, "0.7-0.8"},
		{0.8, 0.9, "0.8-0.9"},
		{0.9, 1.0, "0.9-1.0"},
	}

	rangeCounts := make(map[string]int64)
	rangeCorrects := make(map[string]int64)

	for rows.Next() {
		var confidence float64
		var count, correctCount int64

		err := rows.Scan(&confidence, &count, &correctCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan confidence data: %w", err)
		}

		totalCount += count
		totalCorrect += correctCount
		confidenceSum += confidence * float64(count)

		// Categorize into ranges
		for _, r := range ranges {
			if confidence >= r.start && confidence < r.end {
				rangeCounts[r.name] += count
				rangeCorrects[r.name] += correctCount
				break
			}
		}

		// Categorize into high/medium/low
		if confidence > 0.8 {
			distribution.HighConfidence += count
		} else if confidence >= 0.5 {
			distribution.MediumConfidence += count
		} else {
			distribution.LowConfidence += count
		}
	}

	// Calculate average confidence
	if totalCount > 0 {
		distribution.AverageConfidence = confidenceSum / float64(totalCount)
	}

	// Build confidence ranges
	for _, r := range ranges {
		count := rangeCounts[r.name]
		correct := rangeCorrects[r.name]

		var accuracy float64
		if count > 0 {
			accuracy = float64(correct) / float64(count)
		}

		var percentage float64
		if totalCount > 0 {
			percentage = float64(count) / float64(totalCount) * 100
		}

		distribution.ConfidenceRanges = append(distribution.ConfidenceRanges, AccuracyConfidenceRange{
			RangeStart: r.start,
			RangeEnd:   r.end,
			Count:      count,
			Accuracy:   accuracy,
			Percentage: percentage,
		})
	}

	acs.logger.Printf("📊 Confidence distribution calculated: High=%.2f%%, Medium=%.2f%%, Low=%.2f%%, Avg=%.3f",
		float64(distribution.HighConfidence)/float64(totalCount)*100,
		float64(distribution.MediumConfidence)/float64(totalCount)*100,
		float64(distribution.LowConfidence)/float64(totalCount)*100,
		distribution.AverageConfidence)

	return distribution, nil
}

// CalculateSecurityMetrics calculates security-related accuracy metrics
func (acs *AccuracyCalculationService) CalculateSecurityMetrics(ctx context.Context, hoursBack int) (*SecurityAccuracyMetrics, error) {
	query := `
		SELECT 
			COUNT(*) as total_validations,
			COUNT(CASE WHEN classification_method = 'trusted_source' THEN 1 END) as trusted_data_points,
			COUNT(CASE WHEN classification_method = 'website_verified' THEN 1 END) as verified_website_points,
			COUNT(CASE WHEN classification_method = 'trusted_source' AND is_correct = true THEN 1 END) as trusted_correct,
			COUNT(CASE WHEN classification_method = 'website_verified' AND is_correct = true THEN 1 END) as verified_correct,
			COUNT(CASE WHEN classification_method NOT IN ('trusted_source', 'website_verified') THEN 1 END) as security_violations
		FROM classification_accuracy_metrics 
		WHERE created_at >= NOW() - INTERVAL '%d hours'
	`

	var totalValidations, trustedDataPoints, verifiedWebsitePoints, trustedCorrect, verifiedCorrect, securityViolations int64

	err := acs.db.QueryRowContext(ctx, fmt.Sprintf(query, hoursBack)).Scan(
		&totalValidations,
		&trustedDataPoints,
		&verifiedWebsitePoints,
		&trustedCorrect,
		&verifiedCorrect,
		&securityViolations,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to calculate security metrics: %w", err)
	}

	metrics := &SecurityAccuracyMetrics{
		TrustedDataPoints:        trustedDataPoints,
		VerifiedWebsitePoints:    verifiedWebsitePoints,
		TotalSecurityValidations: totalValidations,
	}

	// Calculate accuracy rates
	if trustedDataPoints > 0 {
		metrics.TrustedDataSourceAccuracy = float64(trustedCorrect) / float64(trustedDataPoints)
	}

	if verifiedWebsitePoints > 0 {
		metrics.WebsiteVerificationAccuracy = float64(verifiedCorrect) / float64(verifiedWebsitePoints)
	}

	// Calculate trust rates
	if totalValidations > 0 {
		metrics.DataSourceTrustRate = float64(trustedDataPoints) / float64(totalValidations)
		metrics.WebsiteVerificationRate = float64(verifiedWebsitePoints) / float64(totalValidations)
		metrics.SecurityViolationRate = float64(securityViolations) / float64(totalValidations)
	}

	acs.logger.Printf("🔒 Security metrics calculated: Trust Rate=%.2f%%, Verification Rate=%.2f%%, Violation Rate=%.2f%%",
		metrics.DataSourceTrustRate*100,
		metrics.WebsiteVerificationRate*100,
		metrics.SecurityViolationRate*100)

	return metrics, nil
}

// CalculatePerformanceMetrics calculates performance-related accuracy metrics
func (acs *AccuracyCalculationService) CalculatePerformanceMetrics(ctx context.Context, hoursBack int) (*PerformanceAccuracyMetrics, error) {
	query := `
		SELECT 
			AVG(response_time_ms) as avg_response_time,
			AVG(COALESCE(processing_time_ms, response_time_ms)) as avg_processing_time,
			COUNT(*) as total_classifications,
			COUNT(CASE WHEN response_time_ms < 200 AND is_correct = true THEN 1 END) as high_perf_correct,
			COUNT(CASE WHEN response_time_ms < 200 THEN 1 END) as high_perf_total,
			COUNT(CASE WHEN response_time_ms >= 200 AND response_time_ms < 500 AND is_correct = true THEN 1 END) as medium_perf_correct,
			COUNT(CASE WHEN response_time_ms >= 200 AND response_time_ms < 500 THEN 1 END) as medium_perf_total,
			COUNT(CASE WHEN response_time_ms >= 500 AND is_correct = true THEN 1 END) as low_perf_correct,
			COUNT(CASE WHEN response_time_ms >= 500 THEN 1 END) as low_perf_total
		FROM classification_accuracy_metrics 
		WHERE created_at >= NOW() - INTERVAL '%d hours'
		AND response_time_ms IS NOT NULL
	`

	var avgResponseTime, avgProcessingTime sql.NullFloat64
	var totalClassifications, highPerfCorrect, highPerfTotal, mediumPerfCorrect, mediumPerfTotal, lowPerfCorrect, lowPerfTotal int64

	err := acs.db.QueryRowContext(ctx, fmt.Sprintf(query, hoursBack)).Scan(
		&avgResponseTime,
		&avgProcessingTime,
		&totalClassifications,
		&highPerfCorrect,
		&highPerfTotal,
		&mediumPerfCorrect,
		&mediumPerfTotal,
		&lowPerfCorrect,
		&lowPerfTotal,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to calculate performance metrics: %w", err)
	}

	metrics := &PerformanceAccuracyMetrics{
		PerformanceRanges: make([]PerformanceRange, 0),
	}

	if avgResponseTime.Valid {
		metrics.AverageResponseTimeMs = avgResponseTime.Float64
	}
	if avgProcessingTime.Valid {
		metrics.AverageProcessingTimeMs = avgProcessingTime.Float64
	}

	// Calculate performance-based accuracy
	if highPerfTotal > 0 {
		metrics.HighPerformanceAccuracy = float64(highPerfCorrect) / float64(highPerfTotal)
	}
	if mediumPerfTotal > 0 {
		metrics.MediumPerformanceAccuracy = float64(mediumPerfCorrect) / float64(mediumPerfTotal)
	}
	if lowPerfTotal > 0 {
		metrics.LowPerformanceAccuracy = float64(lowPerfCorrect) / float64(lowPerfTotal)
	}

	// Define performance ranges
	performanceRanges := []struct {
		start, end float64
		name       string
	}{
		{0, 100, "0-100ms"},
		{100, 200, "100-200ms"},
		{200, 300, "200-300ms"},
		{300, 500, "300-500ms"},
		{500, 1000, "500-1000ms"},
		{1000, math.Inf(1), "1000ms+"},
	}

	// Get detailed performance distribution
	for _, r := range performanceRanges {
		var count, correct int64
		var accuracy float64

		if r.end == math.Inf(1) {
			// Handle the last range (1000ms+)
			query = `
				SELECT 
					COUNT(*) as count,
					COUNT(CASE WHEN is_correct = true THEN 1 END) as correct
				FROM classification_accuracy_metrics 
				WHERE created_at >= NOW() - INTERVAL '%d hours'
				AND response_time_ms >= $1
			`
			err = acs.db.QueryRowContext(ctx, fmt.Sprintf(query, hoursBack), r.start).Scan(&count, &correct)
		} else {
			query = `
				SELECT 
					COUNT(*) as count,
					COUNT(CASE WHEN is_correct = true THEN 1 END) as correct
				FROM classification_accuracy_metrics 
				WHERE created_at >= NOW() - INTERVAL '%d hours'
				AND response_time_ms >= $1 AND response_time_ms < $2
			`
			err = acs.db.QueryRowContext(ctx, fmt.Sprintf(query, hoursBack), r.start, r.end).Scan(&count, &correct)
		}

		if err != nil {
			acs.logger.Printf("Warning: failed to get performance range data for %s: %v", r.name, err)
			continue
		}

		if count > 0 {
			accuracy = float64(correct) / float64(count)
		}

		var percentage float64
		if totalClassifications > 0 {
			percentage = float64(count) / float64(totalClassifications) * 100
		}

		metrics.PerformanceRanges = append(metrics.PerformanceRanges, PerformanceRange{
			RangeStart: r.start,
			RangeEnd:   r.end,
			Count:      count,
			Accuracy:   accuracy,
			Percentage: percentage,
		})
	}

	acs.logger.Printf("⚡ Performance metrics calculated: Avg Response=%.2fms, High Perf Accuracy=%.2f%%, Medium Perf Accuracy=%.2f%%, Low Perf Accuracy=%.2f%%",
		metrics.AverageResponseTimeMs,
		metrics.HighPerformanceAccuracy*100,
		metrics.MediumPerformanceAccuracy*100,
		metrics.LowPerformanceAccuracy*100)

	return metrics, nil
}

// CalculateComprehensiveAccuracy performs a comprehensive accuracy calculation
func (acs *AccuracyCalculationService) CalculateComprehensiveAccuracy(ctx context.Context, hoursBack int) (*AccuracyCalculationResult, error) {
	acs.logger.Printf("🚀 Starting comprehensive accuracy calculation for last %d hours", hoursBack)

	startTime := time.Now()

	// Calculate overall accuracy
	overallAccuracy, err := acs.CalculateOverallAccuracy(ctx, hoursBack)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate overall accuracy: %w", err)
	}

	// Calculate industry-specific accuracy
	industryAccuracy, err := acs.CalculateIndustrySpecificAccuracy(ctx, hoursBack)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate industry-specific accuracy: %w", err)
	}

	// Calculate confidence distribution
	confidenceDistribution, err := acs.CalculateConfidenceDistribution(ctx, hoursBack)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate confidence distribution: %w", err)
	}

	// Calculate security metrics
	securityMetrics, err := acs.CalculateSecurityMetrics(ctx, hoursBack)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate security metrics: %w", err)
	}

	// Calculate performance metrics
	performanceMetrics, err := acs.CalculatePerformanceMetrics(ctx, hoursBack)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate performance metrics: %w", err)
	}

	// Get total data points analyzed
	var totalDataPoints int64
	query := `SELECT COUNT(*) FROM unified_performance_metrics WHERE component = 'classification' AND metric_category = 'classification' AND created_at >= NOW() - INTERVAL '%d hours'`
	err = acs.db.QueryRowContext(ctx, fmt.Sprintf(query, hoursBack)).Scan(&totalDataPoints)
	if err != nil {
		acs.logger.Printf("Warning: failed to get total data points: %v", err)
		totalDataPoints = 0
	}

	result := &AccuracyCalculationResult{
		OverallAccuracy:          overallAccuracy,
		IndustrySpecificAccuracy: industryAccuracy,
		ConfidenceDistribution:   *confidenceDistribution,
		SecurityMetrics:          *securityMetrics,
		PerformanceMetrics:       *performanceMetrics,
		CalculationTimestamp:     time.Now(),
		DataPointsAnalyzed:       totalDataPoints,
		TimeRangeAnalyzed:        fmt.Sprintf("Last %d hours", hoursBack),
	}

	calculationTime := time.Since(startTime)
	acs.logger.Printf("✅ Comprehensive accuracy calculation completed in %.2fms", float64(calculationTime.Nanoseconds())/1e6)
	acs.logger.Printf("📊 Results: Overall Accuracy=%.2f%%, Industries=%d, Data Points=%d",
		overallAccuracy*100, len(industryAccuracy), totalDataPoints)

	return result, nil
}

// GetIndustryAccuracyBreakdown provides detailed breakdown by industry
func (acs *AccuracyCalculationService) GetIndustryAccuracyBreakdown(ctx context.Context, hoursBack int) ([]IndustryAccuracyBreakdown, error) {
	query := `
		SELECT 
			predicted_industry,
			COUNT(*) as total_classifications,
			COUNT(CASE WHEN is_correct = true THEN 1 END) as correct_classifications,
			AVG(predicted_confidence) as avg_confidence,
			AVG(response_time_ms) as avg_response_time
		FROM classification_accuracy_metrics 
		WHERE created_at >= NOW() - INTERVAL '%d hours'
		AND actual_industry IS NOT NULL
		GROUP BY predicted_industry
		ORDER BY total_classifications DESC
	`

	rows, err := acs.db.QueryContext(ctx, fmt.Sprintf(query, hoursBack))
	if err != nil {
		return nil, fmt.Errorf("failed to get industry accuracy breakdown: %w", err)
	}
	defer rows.Close()

	var breakdowns []IndustryAccuracyBreakdown

	for rows.Next() {
		var breakdown IndustryAccuracyBreakdown
		var avgConfidence, avgResponseTime sql.NullFloat64

		err := rows.Scan(
			&breakdown.IndustryName,
			&breakdown.TotalClassifications,
			&breakdown.CorrectClassifications,
			&avgConfidence,
			&avgResponseTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan industry breakdown: %w", err)
		}

		if breakdown.TotalClassifications > 0 {
			breakdown.AccuracyPercentage = float64(breakdown.CorrectClassifications) / float64(breakdown.TotalClassifications) * 100
		}

		if avgConfidence.Valid {
			breakdown.AverageConfidence = avgConfidence.Float64
		}
		if avgResponseTime.Valid {
			breakdown.AverageResponseTime = avgResponseTime.Float64
		}

		// Generate improvement suggestions based on accuracy
		if breakdown.AccuracyPercentage < 70 {
			breakdown.ImprovementSuggestions = append(breakdown.ImprovementSuggestions,
				"Increase keyword coverage for this industry",
				"Review classification algorithm parameters",
				"Add more training data for this industry")
		} else if breakdown.AccuracyPercentage < 85 {
			breakdown.ImprovementSuggestions = append(breakdown.ImprovementSuggestions,
				"Fine-tune keyword weights",
				"Review edge cases and exceptions")
		} else {
			breakdown.ImprovementSuggestions = append(breakdown.ImprovementSuggestions,
				"Maintain current performance",
				"Monitor for any degradation")
		}

		breakdowns = append(breakdowns, breakdown)
	}

	acs.logger.Printf("📊 Industry accuracy breakdown generated for %d industries", len(breakdowns))

	return breakdowns, nil
}

// ValidateAccuracyCalculation validates the accuracy calculation setup
func (acs *AccuracyCalculationService) ValidateAccuracyCalculation(ctx context.Context) error {
	// Check if required tables exist
	requiredTables := []string{"classification_accuracy_metrics"}

	for _, table := range requiredTables {
		var exists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`

		err := acs.db.QueryRowContext(ctx, query, table).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check table existence for %s: %w", table, err)
		}

		if !exists {
			return fmt.Errorf("required table %s does not exist", table)
		}
	}

	// Check if we have any data to analyze
	var count int64
	query := `SELECT COUNT(*) FROM unified_performance_metrics WHERE component = 'classification' AND metric_category = 'classification'`
	err := acs.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check data availability: %w", err)
	}

	if count == 0 {
		acs.logger.Printf("⚠️ No classification accuracy data available for analysis")
	}

	acs.logger.Printf("✅ Accuracy calculation validation passed - %d data points available", count)
	return nil
}
