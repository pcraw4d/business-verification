package classification

import (
	"context"
	"log"
	"time"

	"kyb-platform/internal/shared"
)

// NewMethodPerformanceTracker creates a new method performance tracker
func NewMethodPerformanceTracker(config PerformanceWeightConfig, logger *log.Logger) *MethodPerformanceTracker {
	if logger == nil {
		logger = log.Default()
	}

	return &MethodPerformanceTracker{
		performanceData: make(map[string]*MethodPerformanceData),
		config:          config,
		logger:          logger,
	}
}

// RecordResult records a classification result for performance tracking
func (mpt *MethodPerformanceTracker) RecordResult(
	methodName string,
	result *shared.ClassificationMethodResult,
) {
	mpt.mutex.Lock()
	defer mpt.mutex.Unlock()

	// Get or create performance data for this method
	data, exists := mpt.performanceData[methodName]
	if !exists {
		data = &MethodPerformanceData{
			MethodName:      methodName,
			AccuracyHistory: make([]AccuracyDataPoint, 0),
			LatencyHistory:  make([]LatencyDataPoint, 0),
			WeightHistory:   make([]WeightDataPoint, 0),
			LastUpdated:     time.Now(),
		}
		mpt.performanceData[methodName] = data
	}

	// Update basic metrics
	data.TotalRequests++
	if result.Success {
		data.SuccessfulRequests++
	} else {
		data.FailedRequests++
	}

	// Update accuracy (if available)
	if result.Success && result.Result != nil {
		// Calculate accuracy based on confidence score
		// This is a simplified approach - in a real system, you'd have ground truth data
		accuracy := result.Result.ConfidenceScore
		data.LastAccuracy = accuracy

		// Update average accuracy
		if data.TotalRequests == 1 {
			data.AverageAccuracy = accuracy
		} else {
			// Exponential moving average
			alpha := 0.1 // Learning rate
			data.AverageAccuracy = (alpha * accuracy) + ((1 - alpha) * data.AverageAccuracy)
		}

		// Add to accuracy history
		accuracyPoint := AccuracyDataPoint{
			Timestamp:  time.Now(),
			Accuracy:   accuracy,
			SampleSize: 1,
		}
		data.AccuracyHistory = append(data.AccuracyHistory, accuracyPoint)

		// Keep only recent history (last 100 points)
		if len(data.AccuracyHistory) > 100 {
			data.AccuracyHistory = data.AccuracyHistory[len(data.AccuracyHistory)-100:]
		}
	}

	// Update latency
	if result.ProcessingTime > 0 {
		data.LastLatency = result.ProcessingTime

		// Update average latency
		if data.TotalRequests == 1 {
			data.AverageLatency = result.ProcessingTime
		} else {
			// Exponential moving average
			alpha := 0.1 // Learning rate
			data.AverageLatency = time.Duration(
				(alpha * float64(result.ProcessingTime)) +
					((1 - alpha) * float64(data.AverageLatency)),
			)
		}

		// Add to latency history
		latencyPoint := LatencyDataPoint{
			Timestamp: time.Now(),
			Latency:   result.ProcessingTime,
		}
		data.LatencyHistory = append(data.LatencyHistory, latencyPoint)

		// Keep only recent history (last 100 points)
		if len(data.LatencyHistory) > 100 {
			data.LatencyHistory = data.LatencyHistory[len(data.LatencyHistory)-100:]
		}
	}

	data.LastUpdated = time.Now()

	// Log performance update
	mpt.logger.Printf("📊 Performance update for method '%s': accuracy=%.3f, latency=%v, total_requests=%d",
		methodName, data.AverageAccuracy, data.AverageLatency, data.TotalRequests)
}

// UpdateMethodWeight updates the weight for a method and records it in history
func (mpt *MethodPerformanceTracker) UpdateMethodWeight(methodName string, newWeight float64) {
	mpt.mutex.Lock()
	defer mpt.mutex.Unlock()

	data, exists := mpt.performanceData[methodName]
	if !exists {
		data = &MethodPerformanceData{
			MethodName:      methodName,
			AccuracyHistory: make([]AccuracyDataPoint, 0),
			LatencyHistory:  make([]LatencyDataPoint, 0),
			WeightHistory:   make([]WeightDataPoint, 0),
			LastUpdated:     time.Now(),
		}
		mpt.performanceData[methodName] = data
	}

	// Update current weight
	data.CurrentWeight = newWeight

	// Add to weight history
	weightPoint := WeightDataPoint{
		Timestamp: time.Now(),
		Weight:    newWeight,
		Reason:    "performance_based_adjustment",
	}
	data.WeightHistory = append(data.WeightHistory, weightPoint)

	// Keep only recent history (last 50 points)
	if len(data.WeightHistory) > 50 {
		data.WeightHistory = data.WeightHistory[len(data.WeightHistory)-50:]
	}

	mpt.logger.Printf("📊 Weight updated for method '%s': %.3f", methodName, newWeight)
}

// GetAllPerformanceData returns all performance data
func (mpt *MethodPerformanceTracker) GetAllPerformanceData() map[string]*MethodPerformanceData {
	mpt.mutex.RLock()
	defer mpt.mutex.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]*MethodPerformanceData)
	for methodName, data := range mpt.performanceData {
		// Create a deep copy
		copy := &MethodPerformanceData{
			MethodName:         data.MethodName,
			TotalRequests:      data.TotalRequests,
			SuccessfulRequests: data.SuccessfulRequests,
			FailedRequests:     data.FailedRequests,
			AverageAccuracy:    data.AverageAccuracy,
			AverageLatency:     data.AverageLatency,
			LastAccuracy:       data.LastAccuracy,
			LastLatency:        data.LastLatency,
			LastUpdated:        data.LastUpdated,
			CurrentWeight:      data.CurrentWeight,
		}

		// Copy slices
		copy.AccuracyHistory = make([]AccuracyDataPoint, len(data.AccuracyHistory))
		for i, point := range data.AccuracyHistory {
			copy.AccuracyHistory[i] = point
		}

		copy.LatencyHistory = make([]LatencyDataPoint, len(data.LatencyHistory))
		for i, point := range data.LatencyHistory {
			copy.LatencyHistory[i] = point
		}

		copy.WeightHistory = make([]WeightDataPoint, len(data.WeightHistory))
		for i, point := range data.WeightHistory {
			copy.WeightHistory[i] = point
		}

		result[methodName] = copy
	}

	return result
}

// GetMethodPerformanceData returns performance data for a specific method
func (mpt *MethodPerformanceTracker) GetMethodPerformanceData(methodName string) (*MethodPerformanceData, bool) {
	mpt.mutex.RLock()
	defer mpt.mutex.RUnlock()

	data, exists := mpt.performanceData[methodName]
	if !exists {
		return nil, false
	}

	// Return a copy
	copy := &MethodPerformanceData{
		MethodName:         data.MethodName,
		TotalRequests:      data.TotalRequests,
		SuccessfulRequests: data.SuccessfulRequests,
		FailedRequests:     data.FailedRequests,
		AverageAccuracy:    data.AverageAccuracy,
		AverageLatency:     data.AverageLatency,
		LastAccuracy:       data.LastAccuracy,
		LastLatency:        data.LastLatency,
		LastUpdated:        data.LastUpdated,
		CurrentWeight:      data.CurrentWeight,
	}

	// Copy slices
	copy.AccuracyHistory = make([]AccuracyDataPoint, len(data.AccuracyHistory))
	for i, point := range data.AccuracyHistory {
		copy.AccuracyHistory[i] = point
	}

	copy.LatencyHistory = make([]LatencyDataPoint, len(data.LatencyHistory))
	for i, point := range data.LatencyHistory {
		copy.LatencyHistory[i] = point
	}

	copy.WeightHistory = make([]WeightDataPoint, len(data.WeightHistory))
	for i, point := range data.WeightHistory {
		copy.WeightHistory[i] = point
	}

	return copy, true
}

// GetPerformanceSummary returns a summary of performance metrics
func (mpt *MethodPerformanceTracker) GetPerformanceSummary() map[string]interface{} {
	mpt.mutex.RLock()
	defer mpt.mutex.RUnlock()

	summary := make(map[string]interface{})

	totalRequests := int64(0)
	totalSuccessful := int64(0)
	totalFailed := int64(0)
	methodCount := len(mpt.performanceData)

	for _, data := range mpt.performanceData {
		totalRequests += data.TotalRequests
		totalSuccessful += data.SuccessfulRequests
		totalFailed += data.FailedRequests
	}

	summary["total_methods"] = methodCount
	summary["total_requests"] = totalRequests
	summary["total_successful"] = totalSuccessful
	summary["total_failed"] = totalFailed

	if totalRequests > 0 {
		summary["overall_success_rate"] = float64(totalSuccessful) / float64(totalRequests)
		summary["overall_error_rate"] = float64(totalFailed) / float64(totalRequests)
	} else {
		summary["overall_success_rate"] = 0.0
		summary["overall_error_rate"] = 0.0
	}

	// Method-specific summaries
	methodSummaries := make(map[string]interface{})
	for methodName, data := range mpt.performanceData {
		methodSummary := map[string]interface{}{
			"total_requests":      data.TotalRequests,
			"successful_requests": data.SuccessfulRequests,
			"failed_requests":     data.FailedRequests,
			"average_accuracy":    data.AverageAccuracy,
			"average_latency_ms":  data.AverageLatency.Milliseconds(),
			"current_weight":      data.CurrentWeight,
			"last_updated":        data.LastUpdated,
		}

		if data.TotalRequests > 0 {
			methodSummary["success_rate"] = float64(data.SuccessfulRequests) / float64(data.TotalRequests)
			methodSummary["error_rate"] = float64(data.FailedRequests) / float64(data.TotalRequests)
		} else {
			methodSummary["success_rate"] = 0.0
			methodSummary["error_rate"] = 0.0
		}

		methodSummaries[methodName] = methodSummary
	}

	summary["methods"] = methodSummaries

	return summary
}

// CleanupOldData removes old performance data based on configuration
func (mpt *MethodPerformanceTracker) CleanupOldData() {
	mpt.mutex.Lock()
	defer mpt.mutex.Unlock()

	cutoffTime := time.Now().Add(-mpt.config.PerformanceWindow)

	for methodName, data := range mpt.performanceData {
		// Clean up old accuracy history
		var newAccuracyHistory []AccuracyDataPoint
		for _, point := range data.AccuracyHistory {
			if point.Timestamp.After(cutoffTime) {
				newAccuracyHistory = append(newAccuracyHistory, point)
			}
		}
		data.AccuracyHistory = newAccuracyHistory

		// Clean up old latency history
		var newLatencyHistory []LatencyDataPoint
		for _, point := range data.LatencyHistory {
			if point.Timestamp.After(cutoffTime) {
				newLatencyHistory = append(newLatencyHistory, point)
			}
		}
		data.LatencyHistory = newLatencyHistory

		// Clean up old weight history
		var newWeightHistory []WeightDataPoint
		for _, point := range data.WeightHistory {
			if point.Timestamp.After(cutoffTime) {
				newWeightHistory = append(newWeightHistory, point)
			}
		}
		data.WeightHistory = newWeightHistory

		mpt.logger.Printf("🧹 Cleaned up old data for method '%s': accuracy_points=%d, latency_points=%d, weight_points=%d",
			methodName, len(newAccuracyHistory), len(newLatencyHistory), len(newWeightHistory))
	}
}

// StartCleanupRoutine starts a routine to periodically clean up old data
func (mpt *MethodPerformanceTracker) StartCleanupRoutine(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // Clean up every hour
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				mpt.logger.Printf("🧹 Performance data cleanup routine stopped")
				return
			case <-ticker.C:
				mpt.CleanupOldData()
			}
		}
	}()
}
