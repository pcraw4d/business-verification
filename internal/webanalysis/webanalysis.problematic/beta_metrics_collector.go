package webanalysis

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// BetaMetricsCollector collects and analyzes comprehensive beta testing metrics
type BetaMetricsCollector struct {
	logger        *zap.Logger
	mu            sync.RWMutex
	metrics       map[string]*BetaMetrics // testID -> metrics
	globalMetrics *GlobalBetaMetrics
	abTestResults map[string][]*ABTestResult // testID -> results
	userFeedback  map[string][]*BetaFeedback // testID -> feedback
}

// GlobalBetaMetrics represents global metrics across all beta tests
type GlobalBetaMetrics struct {
	TotalTests          int       `json:"total_tests"`
	TotalABTests        int       `json:"total_ab_tests"`
	TotalFeedback       int       `json:"total_feedback"`
	EnhancedSuccessRate float64   `json:"enhanced_success_rate"`
	BasicSuccessRate    float64   `json:"basic_success_rate"`
	AverageSatisfaction float64   `json:"average_satisfaction"`
	AverageAccuracy     float64   `json:"average_accuracy"`
	AverageSpeed        float64   `json:"average_speed"`
	UserFeedbackCount   int       `json:"user_feedback_count"`
	UniqueUsers         int       `json:"unique_users"`
	LastUpdated         time.Time `json:"last_updated"`
}

// NewBetaMetricsCollector creates a new beta metrics collector
func NewBetaMetricsCollector(logger *zap.Logger) *BetaMetricsCollector {
	return &BetaMetricsCollector{
		logger:        logger,
		metrics:       make(map[string]*BetaMetrics),
		globalMetrics: &GlobalBetaMetrics{},
		abTestResults: make(map[string][]*ABTestResult),
		userFeedback:  make(map[string][]*BetaFeedback),
	}
}

// RecordABTestResult records an A/B test result
func (bmc *BetaMetricsCollector) RecordABTestResult(result *ABTestResult) error {
	bmc.mu.Lock()
	defer bmc.mu.Unlock()

	if result == nil {
		return fmt.Errorf("cannot record nil result")
	}

	// Add to test-specific results
	bmc.abTestResults[result.TestID] = append(bmc.abTestResults[result.TestID], result)

	// Update global metrics
	bmc.globalMetrics.TotalABTests++

	// Invalidate cached metrics for this test
	delete(bmc.metrics, result.TestID)

	bmc.logger.Debug("Recorded A/B test result",
		zap.String("test_id", result.TestID),
		zap.String("method", result.Method),
		zap.Bool("success", result.Success),
		zap.Duration("response_time", result.ResponseTime),
		zap.Float64("accuracy", result.Accuracy),
		zap.Float64("data_quality", result.DataQuality),
	)

	return nil
}

// RecordUserFeedback records user feedback
func (bmc *BetaMetricsCollector) RecordUserFeedback(feedback *BetaFeedback) error {
	bmc.mu.Lock()
	defer bmc.mu.Unlock()

	if feedback == nil {
		return fmt.Errorf("cannot record nil feedback")
	}

	// Add to test-specific feedback
	bmc.userFeedback[feedback.TestID] = append(bmc.userFeedback[feedback.TestID], feedback)

	// Update global metrics
	bmc.globalMetrics.TotalFeedback++
	bmc.globalMetrics.UserFeedbackCount++

	// Invalidate cached metrics for this test
	delete(bmc.metrics, feedback.TestID)

	bmc.logger.Debug("Recorded user feedback",
		zap.String("test_id", feedback.TestID),
		zap.String("user_id", feedback.UserID),
		zap.String("method", feedback.Method),
		zap.Int("satisfaction", feedback.Satisfaction),
		zap.Int("accuracy", feedback.Accuracy),
		zap.Int("speed", feedback.Speed),
	)

	return nil
}

// GetMetrics retrieves metrics for a specific test
func (bmc *BetaMetricsCollector) GetMetrics(ctx context.Context, testID string) (*BetaMetrics, error) {
	bmc.mu.RLock()
	if metrics, exists := bmc.metrics[testID]; exists {
		bmc.mu.RUnlock()
		return metrics, nil
	}
	bmc.mu.RUnlock()

	// Compute metrics
	bmc.mu.Lock()
	defer bmc.mu.Unlock()

	// Double-check cache after acquiring write lock
	if metrics, exists := bmc.metrics[testID]; exists {
		return metrics, nil
	}

	metrics, err := bmc.computeTestMetrics(testID)
	if err != nil {
		return nil, fmt.Errorf("failed to compute metrics: %w", err)
	}

	// Cache the metrics
	bmc.metrics[testID] = metrics

	return metrics, nil
}

// GetGlobalMetrics retrieves global metrics across all tests
func (bmc *BetaMetricsCollector) GetGlobalMetrics(ctx context.Context) (*GlobalBetaMetrics, error) {
	bmc.mu.Lock()
	defer bmc.mu.Unlock()

	// Update global metrics
	err := bmc.updateGlobalMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to update global metrics: %w", err)
	}

	return bmc.globalMetrics, nil
}

// GetMetricsByTimeRange retrieves metrics within a time range
func (bmc *BetaMetricsCollector) GetMetricsByTimeRange(ctx context.Context, start, end time.Time) ([]*BetaMetrics, error) {
	bmc.mu.RLock()
	defer bmc.mu.RUnlock()

	var metrics []*BetaMetrics
	for testID := range bmc.abTestResults {
		// Check if any results are within the time range
		hasResultsInRange := false
		for _, result := range bmc.abTestResults[testID] {
			if result.Timestamp.After(start) && result.Timestamp.Before(end) {
				hasResultsInRange = true
				break
			}
		}

		if hasResultsInRange {
			testMetrics, err := bmc.computeTestMetrics(testID)
			if err != nil {
				bmc.logger.Error("Failed to compute metrics for test",
					zap.String("test_id", testID),
					zap.Error(err),
				)
				continue
			}
			metrics = append(metrics, testMetrics)
		}
	}

	return metrics, nil
}

// GetMethodComparison retrieves comparison metrics between scraping methods
func (bmc *BetaMetricsCollector) GetMethodComparison(ctx context.Context) (*MethodComparison, error) {
	bmc.mu.RLock()
	defer bmc.mu.RUnlock()

	comparison := &MethodComparison{
		Generated: time.Now(),
	}

	// Collect all results
	var basicResults, enhancedResults []*ABTestResult
	for _, results := range bmc.abTestResults {
		for _, result := range results {
			if result.Method == "basic" {
				basicResults = append(basicResults, result)
			} else if result.Method == "enhanced" {
				enhancedResults = append(enhancedResults, result)
			}
		}
	}

	// Calculate basic method metrics
	if len(basicResults) > 0 {
		comparison.BasicMethod = bmc.calculateMethodMetrics(basicResults)
	}

	// Calculate enhanced method metrics
	if len(enhancedResults) > 0 {
		comparison.EnhancedMethod = bmc.calculateMethodMetrics(enhancedResults)
	}

	// Calculate improvements
	if comparison.BasicMethod != nil && comparison.EnhancedMethod != nil {
		comparison.Improvements = &MethodImprovements{
			SuccessRateImprovement:  comparison.EnhancedMethod.SuccessRate - comparison.BasicMethod.SuccessRate,
			ResponseTimeImprovement: comparison.BasicMethod.AverageResponseTime - comparison.EnhancedMethod.AverageResponseTime,
			AccuracyImprovement:     comparison.EnhancedMethod.AverageAccuracy - comparison.BasicMethod.AverageAccuracy,
			DataQualityImprovement:  comparison.EnhancedMethod.AverageDataQuality - comparison.BasicMethod.AverageDataQuality,
		}
	}

	return comparison, nil
}

// computeTestMetrics computes metrics for a specific test
func (bmc *BetaMetricsCollector) computeTestMetrics(testID string) (*BetaMetrics, error) {
	metrics := &BetaMetrics{
		Generated: time.Now(),
	}

	// Get A/B test results for this test
	results, exists := bmc.abTestResults[testID]
	if !exists {
		return metrics, nil
	}

	// Calculate metrics from A/B test results
	var basicResults, enhancedResults []*ABTestResult
	for _, result := range results {
		if result.Method == "basic" {
			basicResults = append(basicResults, result)
		} else if result.Method == "enhanced" {
			enhancedResults = append(enhancedResults, result)
		}
	}

	metrics.TotalTests = len(results)

	// Calculate success rates
	if len(basicResults) > 0 {
		basicSuccess := 0
		for _, result := range basicResults {
			if result.Success {
				basicSuccess++
			}
		}
		metrics.BasicSuccessRate = float64(basicSuccess) / float64(len(basicResults))
	}

	if len(enhancedResults) > 0 {
		enhancedSuccess := 0
		for _, result := range enhancedResults {
			if result.Success {
				enhancedSuccess++
			}
		}
		metrics.EnhancedSuccessRate = float64(enhancedSuccess) / float64(len(enhancedResults))
	}

	// Get user feedback for this test
	feedback, exists := bmc.userFeedback[testID]
	if exists && len(feedback) > 0 {
		metrics.UserFeedbackCount = len(feedback)

		// Calculate average satisfaction, accuracy, and speed
		var totalSatisfaction, totalAccuracy, totalSpeed int
		var satisfactionCount, accuracyCount, speedCount int

		for _, f := range feedback {
			if f.Satisfaction > 0 {
				totalSatisfaction += f.Satisfaction
				satisfactionCount++
			}
			if f.Accuracy > 0 {
				totalAccuracy += f.Accuracy
				accuracyCount++
			}
			if f.Speed > 0 {
				totalSpeed += f.Speed
				speedCount++
			}
		}

		if satisfactionCount > 0 {
			metrics.AverageSatisfaction = float64(totalSatisfaction) / float64(satisfactionCount)
		}
		if accuracyCount > 0 {
			metrics.AverageAccuracy = float64(totalAccuracy) / float64(accuracyCount)
		}
		if speedCount > 0 {
			metrics.AverageSpeed = float64(totalSpeed) / float64(speedCount)
		}
	}

	return metrics, nil
}

// updateGlobalMetrics updates global metrics across all tests
func (bmc *BetaMetricsCollector) updateGlobalMetrics() error {
	bmc.globalMetrics.LastUpdated = time.Now()
	bmc.globalMetrics.TotalTests = len(bmc.abTestResults)

	// Calculate global success rates
	var totalBasicSuccess, totalBasicTests int
	var totalEnhancedSuccess, totalEnhancedTests int
	var totalSatisfaction, totalAccuracy, totalSpeed int
	var satisfactionCount, accuracyCount, speedCount int
	uniqueUsers := make(map[string]bool)

	for testID, results := range bmc.abTestResults {
		for _, result := range results {
			if result.Method == "basic" {
				totalBasicTests++
				if result.Success {
					totalBasicSuccess++
				}
			} else if result.Method == "enhanced" {
				totalEnhancedTests++
				if result.Success {
					totalEnhancedSuccess++
				}
			}
		}

		// Collect feedback metrics
		if feedback, exists := bmc.userFeedback[testID]; exists {
			for _, f := range feedback {
				uniqueUsers[f.UserID] = true

				if f.Satisfaction > 0 {
					totalSatisfaction += f.Satisfaction
					satisfactionCount++
				}
				if f.Accuracy > 0 {
					totalAccuracy += f.Accuracy
					accuracyCount++
				}
				if f.Speed > 0 {
					totalSpeed += f.Speed
					speedCount++
				}
			}
		}
	}

	// Calculate averages
	if totalBasicTests > 0 {
		bmc.globalMetrics.BasicSuccessRate = float64(totalBasicSuccess) / float64(totalBasicTests)
	}
	if totalEnhancedTests > 0 {
		bmc.globalMetrics.EnhancedSuccessRate = float64(totalEnhancedSuccess) / float64(totalEnhancedTests)
	}
	if satisfactionCount > 0 {
		bmc.globalMetrics.AverageSatisfaction = float64(totalSatisfaction) / float64(satisfactionCount)
	}
	if accuracyCount > 0 {
		bmc.globalMetrics.AverageAccuracy = float64(totalAccuracy) / float64(accuracyCount)
	}
	if speedCount > 0 {
		bmc.globalMetrics.AverageSpeed = float64(totalSpeed) / float64(speedCount)
	}

	bmc.globalMetrics.UniqueUsers = len(uniqueUsers)

	return nil
}

// calculateMethodMetrics calculates metrics for a specific method
func (bmc *BetaMetricsCollector) calculateMethodMetrics(results []*ABTestResult) *MethodMetrics {
	if len(results) == 0 {
		return nil
	}

	metrics := &MethodMetrics{
		Method: results[0].Method,
	}

	// Calculate success rate
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}
	metrics.SuccessRate = float64(successCount) / float64(len(results))

	// Calculate average response time
	totalResponseTime := time.Duration(0)
	for _, result := range results {
		totalResponseTime += result.ResponseTime
	}
	metrics.AverageResponseTime = totalResponseTime / time.Duration(len(results))

	// Calculate average accuracy
	totalAccuracy := 0.0
	for _, result := range results {
		totalAccuracy += result.Accuracy
	}
	metrics.AverageAccuracy = totalAccuracy / float64(len(results))

	// Calculate average data quality
	totalDataQuality := 0.0
	for _, result := range results {
		totalDataQuality += result.DataQuality
	}
	metrics.AverageDataQuality = totalDataQuality / float64(len(results))

	return metrics
}

// ExportMetrics exports metrics to JSON format
func (bmc *BetaMetricsCollector) ExportMetrics(ctx context.Context, testID string) ([]byte, error) {
	metrics, err := bmc.GetMetrics(ctx, testID)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(metrics, "", "  ")
}

// ExportGlobalMetrics exports global metrics to JSON format
func (bmc *BetaMetricsCollector) ExportGlobalMetrics(ctx context.Context) ([]byte, error) {
	metrics, err := bmc.GetGlobalMetrics(ctx)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(metrics, "", "  ")
}

// ExportMethodComparison exports method comparison to JSON format
func (bmc *BetaMetricsCollector) ExportMethodComparison(ctx context.Context) ([]byte, error) {
	comparison, err := bmc.GetMethodComparison(ctx)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(comparison, "", "  ")
}

// MethodComparison represents comparison between scraping methods
type MethodComparison struct {
	Generated      time.Time           `json:"generated"`
	BasicMethod    *MethodMetrics      `json:"basic_method"`
	EnhancedMethod *MethodMetrics      `json:"enhanced_method"`
	Improvements   *MethodImprovements `json:"improvements"`
}

// MethodMetrics represents metrics for a specific method
type MethodMetrics struct {
	Method              string        `json:"method"`
	SuccessRate         float64       `json:"success_rate"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	AverageAccuracy     float64       `json:"average_accuracy"`
	AverageDataQuality  float64       `json:"average_data_quality"`
}

// MethodImprovements represents improvements between methods
type MethodImprovements struct {
	SuccessRateImprovement  float64       `json:"success_rate_improvement"`
	ResponseTimeImprovement time.Duration `json:"response_time_improvement"`
	AccuracyImprovement     float64       `json:"accuracy_improvement"`
	DataQualityImprovement  float64       `json:"data_quality_improvement"`
}
