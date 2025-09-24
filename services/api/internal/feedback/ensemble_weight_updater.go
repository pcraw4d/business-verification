package feedback

import (
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
)

// NewEnsembleWeightUpdater creates a new ensemble weight updater
func NewEnsembleWeightUpdater(config *AdvancedLearningConfig, logger *zap.Logger) *EnsembleWeightUpdater {
	// Initialize with default weights
	defaultWeights := map[ClassificationMethod]float64{
		MethodKeyword:    0.50, // 50% weight for keyword matching
		MethodML:         0.40, // 40% weight for ML classification
		MethodSimilarity: 0.10, // 10% weight for similarity analysis
	}

	return &EnsembleWeightUpdater{
		config:            config,
		logger:            logger,
		currentWeights:    defaultWeights,
		methodPerformance: make(map[ClassificationMethod]*MethodPerformanceMetrics),
		weightHistory:     make([]*WeightUpdateRecord, 0),
	}
}

// UpdateWeights updates ensemble weights based on feedback analysis
func (ewu *EnsembleWeightUpdater) UpdateWeights(feedback []*UserFeedback) error {
	ewu.mu.Lock()
	defer ewu.mu.Unlock()

	ewu.logger.Info("Updating ensemble weights based on feedback",
		zap.Int("feedback_count", len(feedback)))

	// Analyze method performance from feedback
	methodPerformance, err := ewu.analyzeMethodPerformance(feedback)
	if err != nil {
		return fmt.Errorf("failed to analyze method performance: %w", err)
	}

	// Calculate new weights based on performance
	newWeights, err := ewu.calculateNewWeights(methodPerformance)
	if err != nil {
		return fmt.Errorf("failed to calculate new weights: %w", err)
	}

	// Check if weight changes are significant enough
	if !ewu.shouldUpdateWeights(newWeights) {
		ewu.logger.Info("Weight changes not significant enough for update")
		return nil
	}

	// Record the weight update
	updateRecord := &WeightUpdateRecord{
		ID:              generateID(),
		Timestamp:       time.Now(),
		PreviousWeights: make(map[ClassificationMethod]float64),
		NewWeights:      make(map[ClassificationMethod]float64),
		UpdateReason:    "performance_based_optimization",
		FeedbackCount:   len(feedback),
	}

	// Copy current weights to previous weights
	for method, weight := range ewu.currentWeights {
		updateRecord.PreviousWeights[method] = weight
	}

	// Copy new weights
	for method, weight := range newWeights {
		updateRecord.NewWeights[method] = weight
	}

	// Calculate performance impact
	updateRecord.PerformanceImpact = ewu.calculatePerformanceImpact(methodPerformance)

	// Update current weights
	ewu.currentWeights = newWeights
	ewu.methodPerformance = methodPerformance

	// Add to weight history
	ewu.weightHistory = append(ewu.weightHistory, updateRecord)

	// Maintain history size
	if len(ewu.weightHistory) > 100 {
		ewu.weightHistory = ewu.weightHistory[1:]
	}

	ewu.logger.Info("Ensemble weights updated successfully",
		zap.Any("previous_weights", updateRecord.PreviousWeights),
		zap.Any("new_weights", updateRecord.NewWeights),
		zap.Float64("performance_impact", updateRecord.PerformanceImpact))

	return nil
}

// analyzeMethodPerformance analyzes performance of each classification method
func (ewu *EnsembleWeightUpdater) analyzeMethodPerformance(feedback []*UserFeedback) (map[ClassificationMethod]*MethodPerformanceMetrics, error) {
	// Group feedback by method
	methodFeedback := make(map[ClassificationMethod][]*UserFeedback)
	for _, fb := range feedback {
		methodFeedback[fb.ClassificationMethod] = append(methodFeedback[fb.ClassificationMethod], fb)
	}

	methodPerformance := make(map[ClassificationMethod]*MethodPerformanceMetrics)

	// Analyze each method
	for method, methodFb := range methodFeedback {
		if len(methodFb) < 10 { // Need minimum samples for reliable analysis
			continue
		}

		performance := ewu.calculateMethodPerformance(methodFb)
		methodPerformance[method] = performance

		ewu.logger.Debug("Method performance analyzed",
			zap.String("method", string(method)),
			zap.Int("feedback_count", len(methodFb)),
			zap.Float64("accuracy", performance.Accuracy),
			zap.Float64("average_confidence", performance.Confidence))
	}

	return methodPerformance, nil
}

// calculateMethodPerformance calculates performance metrics for a method
func (ewu *EnsembleWeightUpdater) calculateMethodPerformance(feedback []*UserFeedback) *MethodPerformanceMetrics {
	metrics := &MethodPerformanceMetrics{
		TotalFeedback: len(feedback),
	}

	// Calculate accuracy based on feedback
	correctPredictions := 0
	totalConfidence := 0.0
	confidenceCount := 0

	for _, fb := range feedback {
		// Calculate accuracy (positive feedback indicates correct prediction)
		if fb.FeedbackType == FeedbackTypeAccuracy || fb.FeedbackType == FeedbackTypeClassification {
			correctPredictions++
		}

		// Calculate average confidence
		if fb.ConfidenceScore > 0 {
			totalConfidence += fb.ConfidenceScore
			confidenceCount++
		}
	}

	// Calculate metrics
	metrics.Accuracy = float64(correctPredictions) / float64(len(feedback))
	if confidenceCount > 0 {
		metrics.Confidence = totalConfidence / float64(confidenceCount)
	}

	// Calculate confidence calibration
	metrics.Consistency = ewu.calculateConfidenceCalibration(feedback)

	// Calculate response time (if available)
	metrics.ProcessingTime = float64(ewu.calculateAverageResponseTime(feedback).Milliseconds())

	// Calculate reliability score
	metrics.Reliability = ewu.calculateReliabilityScore(metrics)

	return metrics
}

// calculateNewWeights calculates new ensemble weights based on method performance
func (ewu *EnsembleWeightUpdater) calculateNewWeights(methodPerformance map[ClassificationMethod]*MethodPerformanceMetrics) (map[ClassificationMethod]float64, error) {
	newWeights := make(map[ClassificationMethod]float64)

	// Calculate performance scores for each method
	performanceScores := make(map[ClassificationMethod]float64)
	totalScore := 0.0

	for method, performance := range methodPerformance {
		// Calculate composite performance score
		score := ewu.calculateCompositePerformanceScore(performance)
		performanceScores[method] = score
		totalScore += score
	}

	// Normalize scores to get weights
	if totalScore > 0 {
		for method, score := range performanceScores {
			newWeights[method] = score / totalScore
		}
	} else {
		// Fallback to equal weights if no performance data
		for method := range methodPerformance {
			newWeights[method] = 1.0 / float64(len(methodPerformance))
		}
	}

	// Apply constraints and smoothing
	newWeights = ewu.applyWeightConstraints(newWeights)

	return newWeights, nil
}

// calculateCompositePerformanceScore calculates a composite performance score
func (ewu *EnsembleWeightUpdater) calculateCompositePerformanceScore(performance *MethodPerformanceMetrics) float64 {
	// Weight different performance aspects
	accuracyWeight := 0.4
	confidenceWeight := 0.3
	calibrationWeight := 0.2
	reliabilityWeight := 0.1

	// Calculate weighted score
	score := performance.Accuracy*accuracyWeight +
		performance.Confidence*confidenceWeight +
		performance.Consistency*calibrationWeight +
		performance.Reliability*reliabilityWeight

	// Ensure score is between 0 and 1
	return math.Max(0.0, math.Min(1.0, score))
}

// applyWeightConstraints applies constraints to weight updates
func (ewu *EnsembleWeightUpdater) applyWeightConstraints(newWeights map[ClassificationMethod]float64) map[ClassificationMethod]float64 {
	constrainedWeights := make(map[ClassificationMethod]float64)

	for method, newWeight := range newWeights {
		currentWeight := ewu.currentWeights[method]

		// Calculate maximum allowed change
		maxChange := ewu.config.MaxWeightChange
		maxNewWeight := currentWeight + maxChange
		minNewWeight := currentWeight - maxChange

		// Apply constraints
		constrainedWeight := math.Max(minNewWeight, math.Min(maxNewWeight, newWeight))

		// Ensure minimum weight
		constrainedWeight = math.Max(0.05, constrainedWeight) // Minimum 5% weight

		constrainedWeights[method] = constrainedWeight
	}

	// Renormalize to ensure weights sum to 1.0
	return ewu.normalizeWeights(constrainedWeights)
}

// normalizeWeights normalizes weights to sum to 1.0
func (ewu *EnsembleWeightUpdater) normalizeWeights(weights map[ClassificationMethod]float64) map[ClassificationMethod]float64 {
	total := 0.0
	for _, weight := range weights {
		total += weight
	}

	if total == 0 {
		// Fallback to equal weights
		equalWeight := 1.0 / float64(len(weights))
		for method := range weights {
			weights[method] = equalWeight
		}
		return weights
	}

	// Normalize
	normalizedWeights := make(map[ClassificationMethod]float64)
	for method, weight := range weights {
		normalizedWeights[method] = weight / total
	}

	return normalizedWeights
}

// shouldUpdateWeights determines if weight changes are significant enough
func (ewu *EnsembleWeightUpdater) shouldUpdateWeights(newWeights map[ClassificationMethod]float64) bool {
	maxChange := 0.0

	for method, newWeight := range newWeights {
		currentWeight := ewu.currentWeights[method]
		change := math.Abs(newWeight - currentWeight)
		if change > maxChange {
			maxChange = change
		}
	}

	return maxChange >= ewu.config.WeightUpdateThreshold
}

// calculatePerformanceImpact calculates the expected performance impact of weight changes
func (ewu *EnsembleWeightUpdater) calculatePerformanceImpact(methodPerformance map[ClassificationMethod]*MethodPerformanceMetrics) float64 {
	// Calculate weighted average accuracy improvement
	totalImprovement := 0.0
	totalWeight := 0.0

	for method, performance := range methodPerformance {
		currentWeight := ewu.currentWeights[method]
		newWeight := ewu.currentWeights[method] // This would be the new weight in practice

		// Calculate improvement contribution
		improvement := performance.Accuracy * (newWeight - currentWeight)
		totalImprovement += improvement
		totalWeight += math.Abs(newWeight - currentWeight)
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalImprovement / totalWeight
}

// calculateConfidenceCalibration calculates confidence calibration score
func (ewu *EnsembleWeightUpdater) calculateConfidenceCalibration(feedback []*UserFeedback) float64 {
	if len(feedback) == 0 {
		return 0.0
	}

	// Group by confidence bins
	confidenceBins := make(map[int][]*UserFeedback)
	for _, fb := range feedback {
		if fb.ConfidenceScore > 0 {
			bin := int(fb.ConfidenceScore * 10) // 0.1 bins
			confidenceBins[bin] = append(confidenceBins[bin], fb)
		}
	}

	// Calculate calibration error for each bin
	totalError := 0.0
	binCount := 0

	for bin, binFeedback := range confidenceBins {
		if len(binFeedback) < 5 { // Need minimum samples per bin
			continue
		}

		// Calculate accuracy in this confidence bin
		correct := 0
		for _, fb := range binFeedback {
			if fb.FeedbackType == FeedbackTypeAccuracy || fb.FeedbackType == FeedbackTypeClassification {
				correct++
			}
		}

		accuracy := float64(correct) / float64(len(binFeedback))
		expectedConfidence := float64(bin) / 10.0
		calibrationError := math.Abs(accuracy - expectedConfidence)

		totalError += calibrationError
		binCount++
	}

	if binCount == 0 {
		return 0.0
	}

	// Return calibration score (1.0 - average error)
	return math.Max(0.0, 1.0-totalError/float64(binCount))
}

// calculateAverageResponseTime calculates average response time
func (ewu *EnsembleWeightUpdater) calculateAverageResponseTime(feedback []*UserFeedback) time.Duration {
	if len(feedback) == 0 {
		return 0
	}

	totalTime := time.Duration(0)
	count := 0

	for _, fb := range feedback {
		if fb.ProcessingTimeMs > 0 {
			totalTime += time.Duration(fb.ProcessingTimeMs) * time.Millisecond
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return totalTime / time.Duration(count)
}

// calculateReliabilityScore calculates reliability score based on consistency
func (ewu *EnsembleWeightUpdater) calculateReliabilityScore(performance *MethodPerformanceMetrics) float64 {
	// Calculate consistency based on feedback distribution
	if performance.TotalFeedback < 10 {
		return 0.5 // Default reliability for small samples
	}

	// Simple reliability calculation based on accuracy and consistency
	reliability := (performance.Accuracy + performance.Consistency) / 2.0
	return math.Max(0.0, math.Min(1.0, reliability))
}

// GetCurrentWeights returns current ensemble weights
func (ewu *EnsembleWeightUpdater) GetCurrentWeights() map[ClassificationMethod]float64 {
	ewu.mu.RLock()
	defer ewu.mu.RUnlock()

	// Return a copy to prevent external modification
	weights := make(map[ClassificationMethod]float64)
	for method, weight := range ewu.currentWeights {
		weights[method] = weight
	}

	return weights
}

// GetWeightUpdateMetrics returns weight update metrics
func (ewu *EnsembleWeightUpdater) GetWeightUpdateMetrics() *WeightUpdateMetrics {
	ewu.mu.RLock()
	defer ewu.mu.RUnlock()

	metrics := &WeightUpdateMetrics{
		TotalUpdates:   len(ewu.weightHistory),
		CurrentWeights: make(map[ClassificationMethod]float64),
	}

	// Copy current weights
	for method, weight := range ewu.currentWeights {
		metrics.CurrentWeights[method] = weight
	}

	// Calculate average improvement
	if len(ewu.weightHistory) > 0 {
		totalImprovement := 0.0
		for _, record := range ewu.weightHistory {
			totalImprovement += record.PerformanceImpact
		}
		metrics.AverageImprovement = totalImprovement / float64(len(ewu.weightHistory))
		metrics.LastUpdate = ewu.weightHistory[len(ewu.weightHistory)-1].Timestamp
	}

	return metrics
}

// GetWeightHistory returns weight update history
func (ewu *EnsembleWeightUpdater) GetWeightHistory(limit int) []*WeightUpdateRecord {
	ewu.mu.RLock()
	defer ewu.mu.RUnlock()

	if limit <= 0 || limit > len(ewu.weightHistory) {
		limit = len(ewu.weightHistory)
	}

	// Return recent records
	start := len(ewu.weightHistory) - limit
	if start < 0 {
		start = 0
	}

	history := make([]*WeightUpdateRecord, limit)
	copy(history, ewu.weightHistory[start:])

	return history
}

// ResetWeights resets weights to default values
func (ewu *EnsembleWeightUpdater) ResetWeights() error {
	ewu.mu.Lock()
	defer ewu.mu.Unlock()

	ewu.logger.Info("Resetting ensemble weights to default values")

	// Record the reset
	updateRecord := &WeightUpdateRecord{
		ID:              generateID(),
		Timestamp:       time.Now(),
		PreviousWeights: make(map[ClassificationMethod]float64),
		NewWeights:      make(map[ClassificationMethod]float64),
		UpdateReason:    "manual_reset",
		FeedbackCount:   0,
	}

	// Copy current weights to previous weights
	for method, weight := range ewu.currentWeights {
		updateRecord.PreviousWeights[method] = weight
	}

	// Set default weights
	defaultWeights := map[ClassificationMethod]float64{
		MethodKeyword:    0.50,
		MethodML:         0.40,
		MethodSimilarity: 0.10,
	}

	for method, weight := range defaultWeights {
		updateRecord.NewWeights[method] = weight
	}

	// Update current weights
	ewu.currentWeights = defaultWeights

	// Add to history
	ewu.weightHistory = append(ewu.weightHistory, updateRecord)

	ewu.logger.Info("Ensemble weights reset successfully")

	return nil
}
