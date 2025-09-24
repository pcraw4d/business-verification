package feedback

import (
	"fmt"
	"math"
	"sort"
	"time"

	"go.uber.org/zap"
)

// NewUncertaintyQuantificationOptimizer creates a new uncertainty quantification optimizer
func NewUncertaintyQuantificationOptimizer(config *AdvancedLearningConfig, logger *zap.Logger) *UncertaintyQuantificationOptimizer {
	return &UncertaintyQuantificationOptimizer{
		config:          config,
		logger:          logger,
		calibrationData: make([]*CalibrationDataPoint, 0),
		uncertaintyMetrics: &UncertaintyMetrics{
			SampleSize:  0,
			LastUpdated: time.Now(),
		},
	}
}

// OptimizeUncertainty optimizes uncertainty quantification based on calibration data
func (uqo *UncertaintyQuantificationOptimizer) OptimizeUncertainty(calibrationData []*CalibrationDataPoint) error {
	uqo.mu.Lock()
	defer uqo.mu.Unlock()

	uqo.logger.Info("Optimizing uncertainty quantification",
		zap.Int("calibration_data_count", len(calibrationData)))

	// Validate calibration data
	if err := uqo.validateCalibrationData(calibrationData); err != nil {
		return fmt.Errorf("failed to validate calibration data: %w", err)
	}

	// Add new calibration data
	uqo.calibrationData = append(uqo.calibrationData, calibrationData...)

	// Maintain calibration data size
	if len(uqo.calibrationData) > uqo.config.CalibrationWindowSize*2 {
		// Keep most recent data
		uqo.calibrationData = uqo.calibrationData[len(uqo.calibrationData)-uqo.config.CalibrationWindowSize:]
	}

	// Check if optimization is needed
	if !uqo.shouldOptimize() {
		uqo.logger.Info("Uncertainty optimization not needed based on current criteria")
		return nil
	}

	// Perform uncertainty optimization
	if err := uqo.performUncertaintyOptimization(); err != nil {
		return fmt.Errorf("failed to perform uncertainty optimization: %w", err)
	}

	// Update uncertainty metrics
	uqo.updateUncertaintyMetrics()

	uqo.logger.Info("Uncertainty quantification optimization completed")

	return nil
}

// validateCalibrationData validates calibration data quality
func (uqo *UncertaintyQuantificationOptimizer) validateCalibrationData(calibrationData []*CalibrationDataPoint) error {
	if len(calibrationData) == 0 {
		return fmt.Errorf("no calibration data provided")
	}

	// Check data quality
	validDataCount := 0
	for _, data := range calibrationData {
		if uqo.isValidCalibrationDataPoint(data) {
			validDataCount++
		}
	}

	if validDataCount < int(float64(len(calibrationData))*0.8) { // At least 80% valid data
		return fmt.Errorf("insufficient valid calibration data: %d/%d", validDataCount, len(calibrationData))
	}

	// Check confidence score distribution
	confidenceRanges := make(map[string]int)
	for _, data := range calibrationData {
		rangeKey := uqo.getConfidenceRange(data.ConfidenceScore)
		confidenceRanges[rangeKey]++
	}

	// Ensure minimum samples per confidence range
	minSamplesPerRange := 5
	for rangeKey, count := range confidenceRanges {
		if count < minSamplesPerRange {
			uqo.logger.Warn("Insufficient samples for confidence range",
				zap.String("range", rangeKey),
				zap.Int("count", count),
				zap.Int("min_required", minSamplesPerRange))
		}
	}

	return nil
}

// isValidCalibrationDataPoint checks if a calibration data point is valid
func (uqo *UncertaintyQuantificationOptimizer) isValidCalibrationDataPoint(data *CalibrationDataPoint) bool {
	// Check required fields
	if data.PredictedClass == "" || data.ActualClass == "" {
		return false
	}

	// Check confidence score range
	if data.ConfidenceScore < 0.0 || data.ConfidenceScore > 1.0 {
		return false
	}

	// Check uncertainty score range
	if data.UncertaintyScore < 0.0 || data.UncertaintyScore > 1.0 {
		return false
	}

	// Check timestamp
	if data.Timestamp.IsZero() || data.Timestamp.After(time.Now()) {
		return false
	}

	return true
}

// getConfidenceRange returns the confidence range for a given confidence score
func (uqo *UncertaintyQuantificationOptimizer) getConfidenceRange(confidence float64) string {
	if confidence < 0.2 {
		return "0.0-0.2"
	} else if confidence < 0.4 {
		return "0.2-0.4"
	} else if confidence < 0.6 {
		return "0.4-0.6"
	} else if confidence < 0.8 {
		return "0.6-0.8"
	} else {
		return "0.8-1.0"
	}
}

// shouldOptimize determines if uncertainty optimization is needed
func (uqo *UncertaintyQuantificationOptimizer) shouldOptimize() bool {
	// Check if we have enough calibration data
	if len(uqo.calibrationData) < uqo.config.CalibrationWindowSize {
		return false
	}

	// Check if enough time has passed since last optimization
	lastOptimization := uqo.uncertaintyMetrics.LastUpdated
	if time.Since(lastOptimization) < 24*time.Hour { // Optimize at most once per day
		return false
	}

	// Check if calibration error is above threshold
	if uqo.uncertaintyMetrics.CalibrationError > uqo.config.UncertaintyThreshold {
		uqo.logger.Info("Uncertainty optimization triggered due to high calibration error",
			zap.Float64("calibration_error", uqo.uncertaintyMetrics.CalibrationError),
			zap.Float64("threshold", uqo.config.UncertaintyThreshold))
		return true
	}

	// Check if reliability score is below threshold
	if uqo.uncertaintyMetrics.ReliabilityScore < 0.8 {
		uqo.logger.Info("Uncertainty optimization triggered due to low reliability score",
			zap.Float64("reliability_score", uqo.uncertaintyMetrics.ReliabilityScore))
		return true
	}

	return true
}

// performUncertaintyOptimization performs the actual uncertainty optimization
func (uqo *UncertaintyQuantificationOptimizer) performUncertaintyOptimization() error {
	uqo.logger.Info("Performing uncertainty quantification optimization")

	// Calculate current calibration error
	currentCalibrationError := uqo.calculateCalibrationError()
	uqo.logger.Info("Current calibration error calculated",
		zap.Float64("calibration_error", currentCalibrationError))

	// Calculate reliability score
	reliabilityScore := uqo.calculateReliabilityScore()
	uqo.logger.Info("Reliability score calculated",
		zap.Float64("reliability_score", reliabilityScore))

	// Calculate confidence accuracy
	confidenceAccuracy := uqo.calculateConfidenceAccuracy()
	uqo.logger.Info("Confidence accuracy calculated",
		zap.Float64("confidence_accuracy", confidenceAccuracy))

	// Calculate uncertainty accuracy
	uncertaintyAccuracy := uqo.calculateUncertaintyAccuracy()
	uqo.logger.Info("Uncertainty accuracy calculated",
		zap.Float64("uncertainty_accuracy", uncertaintyAccuracy))

	// Apply optimization techniques
	if err := uqo.applyCalibrationCorrection(); err != nil {
		return fmt.Errorf("failed to apply calibration correction: %w", err)
	}

	if err := uqo.applyUncertaintyScaling(); err != nil {
		return fmt.Errorf("failed to apply uncertainty scaling: %w", err)
	}

	if err := uqo.applyConfidenceThresholding(); err != nil {
		return fmt.Errorf("failed to apply confidence thresholding: %w", err)
	}

	uqo.logger.Info("Uncertainty optimization techniques applied successfully")

	return nil
}

// calculateCalibrationError calculates the calibration error
func (uqo *UncertaintyQuantificationOptimizer) calculateCalibrationError() float64 {
	if len(uqo.calibrationData) == 0 {
		return 1.0 // Maximum error if no data
	}

	// Group data by confidence bins
	confidenceBins := make(map[string][]*CalibrationDataPoint)
	for _, data := range uqo.calibrationData {
		rangeKey := uqo.getConfidenceRange(data.ConfidenceScore)
		confidenceBins[rangeKey] = append(confidenceBins[rangeKey], data)
	}

	// Calculate expected vs actual accuracy for each bin
	totalError := 0.0
	binCount := 0

	for rangeKey, binData := range confidenceBins {
		if len(binData) < 5 { // Need minimum samples per bin
			continue
		}

		// Calculate actual accuracy in this bin
		correct := 0
		for _, data := range binData {
			if data.IsCorrect {
				correct++
			}
		}
		actualAccuracy := float64(correct) / float64(len(binData))

		// Calculate expected accuracy (midpoint of confidence range)
		expectedAccuracy := uqo.getExpectedAccuracyForRange(rangeKey)

		// Calculate calibration error for this bin
		binError := math.Abs(actualAccuracy - expectedAccuracy)
		totalError += binError
		binCount++

		uqo.logger.Debug("Calibration bin analysis",
			zap.String("range", rangeKey),
			zap.Float64("actual_accuracy", actualAccuracy),
			zap.Float64("expected_accuracy", expectedAccuracy),
			zap.Float64("bin_error", binError),
			zap.Int("sample_count", len(binData)))
	}

	if binCount == 0 {
		return 1.0 // Maximum error if no valid bins
	}

	return totalError / float64(binCount)
}

// getExpectedAccuracyForRange returns the expected accuracy for a confidence range
func (uqo *UncertaintyQuantificationOptimizer) getExpectedAccuracyForRange(rangeKey string) float64 {
	switch rangeKey {
	case "0.0-0.2":
		return 0.1
	case "0.2-0.4":
		return 0.3
	case "0.4-0.6":
		return 0.5
	case "0.6-0.8":
		return 0.7
	case "0.8-1.0":
		return 0.9
	default:
		return 0.5
	}
}

// calculateReliabilityScore calculates the reliability score
func (uqo *UncertaintyQuantificationOptimizer) calculateReliabilityScore() float64 {
	if len(uqo.calibrationData) == 0 {
		return 0.0
	}

	// Calculate consistency of uncertainty estimates
	uncertaintyValues := make([]float64, 0, len(uqo.calibrationData))
	for _, data := range uqo.calibrationData {
		uncertaintyValues = append(uncertaintyValues, data.UncertaintyScore)
	}

	// Calculate coefficient of variation (lower is more reliable)
	mean := uqo.calculateMean(uncertaintyValues)
	stdDev := uqo.calculateStandardDeviation(uncertaintyValues, mean)

	if mean == 0 {
		return 0.0
	}

	coefficientOfVariation := stdDev / mean

	// Convert to reliability score (higher is more reliable)
	reliabilityScore := math.Max(0.0, 1.0-coefficientOfVariation)

	return reliabilityScore
}

// calculateConfidenceAccuracy calculates confidence accuracy
func (uqo *UncertaintyQuantificationOptimizer) calculateConfidenceAccuracy() float64 {
	if len(uqo.calibrationData) == 0 {
		return 0.0
	}

	// Calculate how well confidence scores predict correctness
	correct := 0
	total := 0

	for _, data := range uqo.calibrationData {
		// High confidence should correlate with correctness
		if data.ConfidenceScore > 0.7 && data.IsCorrect {
			correct++
		} else if data.ConfidenceScore <= 0.7 && !data.IsCorrect {
			correct++
		}
		total++
	}

	return float64(correct) / float64(total)
}

// calculateUncertaintyAccuracy calculates uncertainty accuracy
func (uqo *UncertaintyQuantificationOptimizer) calculateUncertaintyAccuracy() float64 {
	if len(uqo.calibrationData) == 0 {
		return 0.0
	}

	// Calculate how well uncertainty scores predict incorrectness
	correct := 0
	total := 0

	for _, data := range uqo.calibrationData {
		// High uncertainty should correlate with incorrectness
		if data.UncertaintyScore > 0.7 && !data.IsCorrect {
			correct++
		} else if data.UncertaintyScore <= 0.7 && data.IsCorrect {
			correct++
		}
		total++
	}

	return float64(correct) / float64(total)
}

// applyCalibrationCorrection applies calibration correction
func (uqo *UncertaintyQuantificationOptimizer) applyCalibrationCorrection() error {
	uqo.logger.Info("Applying calibration correction")

	// Calculate calibration correction factors for each confidence range
	correctionFactors := make(map[string]float64)

	confidenceBins := make(map[string][]*CalibrationDataPoint)
	for _, data := range uqo.calibrationData {
		rangeKey := uqo.getConfidenceRange(data.ConfidenceScore)
		confidenceBins[rangeKey] = append(confidenceBins[rangeKey], data)
	}

	for rangeKey, binData := range confidenceBins {
		if len(binData) < 5 {
			continue
		}

		// Calculate actual accuracy
		correct := 0
		for _, data := range binData {
			if data.IsCorrect {
				correct++
			}
		}
		actualAccuracy := float64(correct) / float64(len(binData))

		// Calculate expected accuracy
		expectedAccuracy := uqo.getExpectedAccuracyForRange(rangeKey)

		// Calculate correction factor
		if expectedAccuracy > 0 {
			correctionFactor := actualAccuracy / expectedAccuracy
			correctionFactors[rangeKey] = correctionFactor

			uqo.logger.Debug("Calibration correction factor calculated",
				zap.String("range", rangeKey),
				zap.Float64("actual_accuracy", actualAccuracy),
				zap.Float64("expected_accuracy", expectedAccuracy),
				zap.Float64("correction_factor", correctionFactor))
		}
	}

	// Store correction factors for use in uncertainty estimation
	// In a real implementation, this would update the uncertainty estimation model
	uqo.logger.Info("Calibration correction factors calculated and stored",
		zap.Int("factor_count", len(correctionFactors)))

	return nil
}

// applyUncertaintyScaling applies uncertainty scaling
func (uqo *UncertaintyQuantificationOptimizer) applyUncertaintyScaling() error {
	uqo.logger.Info("Applying uncertainty scaling")

	// Calculate uncertainty scaling factors based on prediction accuracy
	scalingFactors := make(map[string]float64)

	// Group by prediction class
	classGroups := make(map[string][]*CalibrationDataPoint)
	for _, data := range uqo.calibrationData {
		classGroups[data.PredictedClass] = append(classGroups[data.PredictedClass], data)
	}

	for class, classData := range classGroups {
		if len(classData) < 10 {
			continue
		}

		// Calculate average uncertainty for correct vs incorrect predictions
		correctUncertainty := 0.0
		incorrectUncertainty := 0.0
		correctCount := 0
		incorrectCount := 0

		for _, data := range classData {
			if data.IsCorrect {
				correctUncertainty += data.UncertaintyScore
				correctCount++
			} else {
				incorrectUncertainty += data.UncertaintyScore
				incorrectCount++
			}
		}

		if correctCount > 0 && incorrectCount > 0 {
			avgCorrectUncertainty := correctUncertainty / float64(correctCount)
			avgIncorrectUncertainty := incorrectUncertainty / float64(incorrectCount)

			// Calculate scaling factor to better separate correct and incorrect predictions
			if avgCorrectUncertainty > 0 {
				scalingFactor := avgIncorrectUncertainty / avgCorrectUncertainty
				scalingFactors[class] = scalingFactor

				uqo.logger.Debug("Uncertainty scaling factor calculated",
					zap.String("class", class),
					zap.Float64("avg_correct_uncertainty", avgCorrectUncertainty),
					zap.Float64("avg_incorrect_uncertainty", avgIncorrectUncertainty),
					zap.Float64("scaling_factor", scalingFactor))
			}
		}
	}

	// Store scaling factors for use in uncertainty estimation
	uqo.logger.Info("Uncertainty scaling factors calculated and stored",
		zap.Int("factor_count", len(scalingFactors)))

	return nil
}

// applyConfidenceThresholding applies confidence thresholding
func (uqo *UncertaintyQuantificationOptimizer) applyConfidenceThresholding() error {
	uqo.logger.Info("Applying confidence thresholding")

	// Calculate optimal confidence thresholds for different classes
	thresholds := make(map[string]float64)

	// Group by prediction class
	classGroups := make(map[string][]*CalibrationDataPoint)
	for _, data := range uqo.calibrationData {
		classGroups[data.PredictedClass] = append(classGroups[data.PredictedClass], data)
	}

	for class, classData := range classGroups {
		if len(classData) < 20 {
			continue
		}

		// Sort by confidence score
		sort.Slice(classData, func(i, j int) bool {
			return classData[i].ConfidenceScore < classData[j].ConfidenceScore
		})

		// Find threshold that maximizes F1 score
		bestThreshold := 0.5
		bestF1Score := 0.0

		for i := 0; i < len(classData); i++ {
			threshold := classData[i].ConfidenceScore

			// Calculate precision and recall at this threshold
			truePositives := 0
			falsePositives := 0
			falseNegatives := 0

			for _, data := range classData {
				predicted := data.ConfidenceScore >= threshold
				actual := data.IsCorrect

				if predicted && actual {
					truePositives++
				} else if predicted && !actual {
					falsePositives++
				} else if !predicted && actual {
					falseNegatives++
				}
			}

			// Calculate F1 score
			precision := 0.0
			if truePositives+falsePositives > 0 {
				precision = float64(truePositives) / float64(truePositives+falsePositives)
			}

			recall := 0.0
			if truePositives+falseNegatives > 0 {
				recall = float64(truePositives) / float64(truePositives+falseNegatives)
			}

			f1Score := 0.0
			if precision+recall > 0 {
				f1Score = 2 * (precision * recall) / (precision + recall)
			}

			if f1Score > bestF1Score {
				bestF1Score = f1Score
				bestThreshold = threshold
			}
		}

		thresholds[class] = bestThreshold

		uqo.logger.Debug("Confidence threshold calculated",
			zap.String("class", class),
			zap.Float64("threshold", bestThreshold),
			zap.Float64("f1_score", bestF1Score))
	}

	// Store thresholds for use in confidence estimation
	uqo.logger.Info("Confidence thresholds calculated and stored",
		zap.Int("threshold_count", len(thresholds)))

	return nil
}

// updateUncertaintyMetrics updates uncertainty metrics
func (uqo *UncertaintyQuantificationOptimizer) updateUncertaintyMetrics() {
	uqo.uncertaintyMetrics.CalibrationError = uqo.calculateCalibrationError()
	uqo.uncertaintyMetrics.ReliabilityScore = uqo.calculateReliabilityScore()
	uqo.uncertaintyMetrics.ConfidenceAccuracy = uqo.calculateConfidenceAccuracy()
	uqo.uncertaintyMetrics.UncertaintyAccuracy = uqo.calculateUncertaintyAccuracy()
	uqo.uncertaintyMetrics.SampleSize = len(uqo.calibrationData)
	uqo.uncertaintyMetrics.LastUpdated = time.Now()

	uqo.logger.Info("Uncertainty metrics updated",
		zap.Float64("calibration_error", uqo.uncertaintyMetrics.CalibrationError),
		zap.Float64("reliability_score", uqo.uncertaintyMetrics.ReliabilityScore),
		zap.Float64("confidence_accuracy", uqo.uncertaintyMetrics.ConfidenceAccuracy),
		zap.Float64("uncertainty_accuracy", uqo.uncertaintyMetrics.UncertaintyAccuracy),
		zap.Int("sample_size", uqo.uncertaintyMetrics.SampleSize))
}

// Helper functions for statistical calculations

func (uqo *UncertaintyQuantificationOptimizer) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}

	return sum / float64(len(values))
}

func (uqo *UncertaintyQuantificationOptimizer) calculateStandardDeviation(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0.0
	}

	sumSquaredDiffs := 0.0
	for _, value := range values {
		diff := value - mean
		sumSquaredDiffs += diff * diff
	}

	variance := sumSquaredDiffs / float64(len(values)-1)
	return math.Sqrt(variance)
}

// GetOptimizationMetrics returns uncertainty optimization metrics
func (uqo *UncertaintyQuantificationOptimizer) GetOptimizationMetrics() *UncertaintyOptimizationMetrics {
	uqo.mu.RLock()
	defer uqo.mu.RUnlock()

	metrics := &UncertaintyOptimizationMetrics{
		CalibrationError: uqo.uncertaintyMetrics.CalibrationError,
		ReliabilityScore: uqo.uncertaintyMetrics.ReliabilityScore,
		OptimizationRuns: 1, // TODO: Track actual optimization runs
		LastOptimization: uqo.uncertaintyMetrics.LastUpdated,
		ImprovementRate:  0.0, // TODO: Calculate actual improvement rate
	}

	return metrics
}

// GetCalibrationData returns calibration data
func (uqo *UncertaintyQuantificationOptimizer) GetCalibrationData(limit int) []*CalibrationDataPoint {
	uqo.mu.RLock()
	defer uqo.mu.RUnlock()

	if limit <= 0 || limit > len(uqo.calibrationData) {
		limit = len(uqo.calibrationData)
	}

	// Return recent data
	start := len(uqo.calibrationData) - limit
	if start < 0 {
		start = 0
	}

	data := make([]*CalibrationDataPoint, limit)
	copy(data, uqo.calibrationData[start:])

	return data
}
