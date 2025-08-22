package classification_optimization

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// PerformanceTracker tracks performance metrics for classification algorithms
type PerformanceTracker struct {
	logger    *zap.Logger
	mu        sync.RWMutex
	metrics   map[string]*AlgorithmMetrics
	history   map[string][]*AlgorithmMetrics
	startTime time.Time
}

// NewPerformanceTracker creates a new performance tracker
func NewPerformanceTracker(logger *zap.Logger) *PerformanceTracker {
	return &PerformanceTracker{
		logger:    logger,
		metrics:   make(map[string]*AlgorithmMetrics),
		history:   make(map[string][]*AlgorithmMetrics),
		startTime: time.Now(),
	}
}

// RecordClassificationResult records a classification result for performance tracking
func (pt *PerformanceTracker) RecordClassificationResult(category string, result *ClassificationResult) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	// Get or create metrics for the category
	metrics, exists := pt.metrics[category]
	if !exists {
		metrics = &AlgorithmMetrics{
			Accuracy:              0.0,
			Precision:             0.0,
			Recall:                0.0,
			F1Score:               0.0,
			MisclassificationRate: 0.0,
			ConfidenceScore:       0.0,
			ProcessingTime:        0.0,
			Throughput:            0.0,
			ErrorRate:             0.0,
		}
		pt.metrics[category] = metrics
	}

	// Update metrics based on the result
	pt.updateMetrics(metrics, result)

	// Add to history
	pt.history[category] = append(pt.history[category], pt.cloneMetrics(metrics))
}

// ClassificationResult represents a single classification result
type ClassificationResult struct {
	ExpectedCategory  string        `json:"expected_category"`
	PredictedCategory string        `json:"predicted_category"`
	Confidence        float64       `json:"confidence"`
	ProcessingTime    time.Duration `json:"processing_time"`
	Timestamp         time.Time     `json:"timestamp"`
	IsCorrect         bool          `json:"is_correct"`
	ErrorType         string        `json:"error_type,omitempty"`
}

// updateMetrics updates metrics based on a classification result
func (pt *PerformanceTracker) updateMetrics(metrics *AlgorithmMetrics, result *ClassificationResult) {
	// This is a simplified implementation
	// In a real system, you would maintain running averages and more sophisticated calculations

	// For now, we'll use simple averaging
	// In practice, you'd want to use exponential moving averages or other techniques

	// Update confidence score
	metrics.ConfidenceScore = (metrics.ConfidenceScore + result.Confidence) / 2

	// Update processing time
	metrics.ProcessingTime = (metrics.ProcessingTime + float64(result.ProcessingTime.Milliseconds())) / 2

	// Update accuracy (simplified)
	if result.IsCorrect {
		metrics.Accuracy = (metrics.Accuracy + 1.0) / 2
	} else {
		metrics.Accuracy = (metrics.Accuracy + 0.0) / 2
		metrics.MisclassificationRate = (metrics.MisclassificationRate + 1.0) / 2
	}

	// Update throughput (simplified)
	metrics.Throughput = 1.0 / (metrics.ProcessingTime / 1000.0) // requests per second
}

// cloneMetrics creates a copy of metrics for history
func (pt *PerformanceTracker) cloneMetrics(metrics *AlgorithmMetrics) *AlgorithmMetrics {
	return &AlgorithmMetrics{
		Accuracy:              metrics.Accuracy,
		Precision:             metrics.Precision,
		Recall:                metrics.Recall,
		F1Score:               metrics.F1Score,
		MisclassificationRate: metrics.MisclassificationRate,
		ConfidenceScore:       metrics.ConfidenceScore,
		ProcessingTime:        metrics.ProcessingTime,
		Throughput:            metrics.Throughput,
		ErrorRate:             metrics.ErrorRate,
	}
}

// GetCategoryMetrics returns current metrics for a category
func (pt *PerformanceTracker) GetCategoryMetrics(category string) *AlgorithmMetrics {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	metrics, exists := pt.metrics[category]
	if !exists {
		return nil
	}

	return pt.cloneMetrics(metrics)
}

// GetAllMetrics returns metrics for all categories
func (pt *PerformanceTracker) GetAllMetrics() map[string]*AlgorithmMetrics {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	result := make(map[string]*AlgorithmMetrics)
	for category, metrics := range pt.metrics {
		result[category] = pt.cloneMetrics(metrics)
	}

	return result
}

// GetMetricsHistory returns historical metrics for a category
func (pt *PerformanceTracker) GetMetricsHistory(category string, limit int) []*AlgorithmMetrics {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	history, exists := pt.history[category]
	if !exists {
		return nil
	}

	if limit <= 0 || limit > len(history) {
		limit = len(history)
	}

	result := make([]*AlgorithmMetrics, limit)
	copy(result, history[len(history)-limit:])

	return result
}

// ResetMetrics resets metrics for a category
func (pt *PerformanceTracker) ResetMetrics(category string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	delete(pt.metrics, category)
	delete(pt.history, category)
}

// GetPerformanceSummary returns a summary of overall performance
func (pt *PerformanceTracker) GetPerformanceSummary() *PerformanceSummary {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	summary := &PerformanceSummary{
		TotalCategories:         len(pt.metrics),
		AverageAccuracy:         0.0,
		AverageConfidence:       0.0,
		AverageProcessingTime:   0.0,
		CategoriesByPerformance: make(map[string]string),
	}

	if len(pt.metrics) == 0 {
		return summary
	}

	var totalAccuracy float64
	var totalConfidence float64
	var totalProcessingTime float64

	for category, metrics := range pt.metrics {
		totalAccuracy += metrics.Accuracy
		totalConfidence += metrics.ConfidenceScore
		totalProcessingTime += metrics.ProcessingTime

		// Categorize performance
		performance := "poor"
		if metrics.Accuracy > 0.8 {
			performance = "excellent"
		} else if metrics.Accuracy > 0.6 {
			performance = "good"
		} else if metrics.Accuracy > 0.4 {
			performance = "fair"
		}

		summary.CategoriesByPerformance[category] = performance
	}

	summary.AverageAccuracy = totalAccuracy / float64(len(pt.metrics))
	summary.AverageConfidence = totalConfidence / float64(len(pt.metrics))
	summary.AverageProcessingTime = totalProcessingTime / float64(len(pt.metrics))

	return summary
}

// PerformanceSummary represents a summary of overall performance
type PerformanceSummary struct {
	TotalCategories         int               `json:"total_categories"`
	AverageAccuracy         float64           `json:"average_accuracy"`
	AverageConfidence       float64           `json:"average_confidence"`
	AverageProcessingTime   float64           `json:"average_processing_time"`
	CategoriesByPerformance map[string]string `json:"categories_by_performance"`
}
