package classification

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"
)

// ConfidenceCalibrator calibrates confidence scores to match actual accuracy
type ConfidenceCalibrator struct {
	logger *log.Logger
	mu     sync.RWMutex

	// Calibration data by confidence bin
	calibrationBins map[int]*CalibrationBin
	binSize         float64 // Size of each confidence bin (e.g., 0.1 for 10 bins)
	numBins         int     // Number of bins (default 10)

	// Target accuracy
	targetAccuracy float64 // Target accuracy (default 0.95)

	// Historical data
	totalClassifications int64
	correctClassifications int64
	lastCalibration      time.Time
	calibrationInterval  time.Duration // How often to recalibrate

	// Calibration adjustments
	confidenceAdjustments map[int]float64 // Bin -> adjustment factor
}

// CalibrationBin represents accuracy data for a confidence bin
type CalibrationBin struct {
	BinIndex         int     // Bin index (0 to numBins-1)
	ConfidenceRange  string  // Range string (e.g., "0.8-0.9")
	PredictedAccuracy float64 // Average confidence in this bin
	ActualAccuracy    float64 // Actual accuracy in this bin
	CalibrationError float64 // |PredictedAccuracy - ActualAccuracy|
	SampleSize       int64   // Number of samples in this bin
	LastUpdated      time.Time
}

// CalibrationResult represents the result of calibration analysis
type CalibrationResult struct {
	IsCalibrated      bool              `json:"is_calibrated"`
	OverallAccuracy   float64           `json:"overall_accuracy"`
	TargetAccuracy    float64           `json:"target_accuracy"`
	CalibrationBins   []CalibrationBin  `json:"calibration_bins"`
	RecommendedThreshold float64        `json:"recommended_threshold"` // Min confidence to achieve target
	Adjustments       map[int]float64   `json:"adjustments"` // Bin -> adjustment factor
	LastCalibration   time.Time         `json:"last_calibration"`
}

// NewConfidenceCalibrator creates a new confidence calibrator
func NewConfidenceCalibrator(logger *log.Logger) *ConfidenceCalibrator {
	if logger == nil {
		logger = log.Default()
	}

	calibrator := &ConfidenceCalibrator{
		logger:                logger,
		calibrationBins:       make(map[int]*CalibrationBin),
		binSize:               0.1, // 10 bins of 0.1 each
		numBins:               10,
		targetAccuracy:         0.95, // 95% target
		calibrationInterval:   24 * time.Hour, // Recalibrate daily
		confidenceAdjustments: make(map[int]float64),
	}

	// Initialize bins
	for i := 0; i < calibrator.numBins; i++ {
		calibrator.calibrationBins[i] = &CalibrationBin{
			BinIndex:    i,
			ConfidenceRange: fmt.Sprintf("%.1f-%.1f", float64(i)*calibrator.binSize, float64(i+1)*calibrator.binSize),
			LastUpdated: time.Now(),
		}
	}

	return calibrator
}

// RecordClassification records a classification result for calibration
func (cc *ConfidenceCalibrator) RecordClassification(
	ctx context.Context,
	confidence float64,
	actualIndustry string,
	predictedIndustry string,
	isCorrect bool,
) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	// Clamp confidence to [0.0, 1.0]
	confidence = math.Max(0.0, math.Min(1.0, confidence))

	// Determine bin index
	binIndex := int(confidence / cc.binSize)
	if binIndex >= cc.numBins {
		binIndex = cc.numBins - 1
	}

	// Get or create bin
	bin := cc.calibrationBins[binIndex]

	// Update bin statistics
	bin.SampleSize++
	bin.PredictedAccuracy = (bin.PredictedAccuracy*float64(bin.SampleSize-1) + confidence) / float64(bin.SampleSize)

	// Update actual accuracy
	accuracy := 0.0
	if isCorrect {
		accuracy = 1.0
		cc.correctClassifications++
	}
	bin.ActualAccuracy = (bin.ActualAccuracy*float64(bin.SampleSize-1) + accuracy) / float64(bin.SampleSize)

	// Calculate calibration error
	bin.CalibrationError = math.Abs(bin.PredictedAccuracy - bin.ActualAccuracy)
	bin.LastUpdated = time.Now()

	// Update overall statistics
	cc.totalClassifications++

	// Log calibration data periodically
	if cc.totalClassifications%100 == 0 {
		cc.logger.Printf("📊 [Calibration] Recorded %d classifications, overall accuracy: %.2f%%",
			cc.totalClassifications, float64(cc.correctClassifications)/float64(cc.totalClassifications)*100)
	}

	return nil
}

// Calibrate performs calibration analysis and adjusts confidence scores
func (cc *ConfidenceCalibrator) Calibrate(ctx context.Context) (*CalibrationResult, error) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.logger.Printf("🔧 [Calibration] Starting confidence calibration analysis")

	// Calculate overall accuracy
	overallAccuracy := 0.0
	if cc.totalClassifications > 0 {
		overallAccuracy = float64(cc.correctClassifications) / float64(cc.totalClassifications)
	}

	// Build calibration bins list
	bins := make([]CalibrationBin, 0, cc.numBins)
	for i := 0; i < cc.numBins; i++ {
		if bin := cc.calibrationBins[i]; bin.SampleSize > 0 {
			bins = append(bins, *bin)
		}
	}

	// Sort bins by bin index
	sort.Slice(bins, func(i, j int) bool {
		return bins[i].BinIndex < bins[j].BinIndex
	})

	// Calculate adjustments for each bin
	adjustments := make(map[int]float64)
	for _, bin := range bins {
		if bin.SampleSize >= 10 { // Minimum samples for reliable calibration
			// Calculate adjustment factor to align predicted with actual
			if bin.PredictedAccuracy > 0 {
				adjustmentFactor := bin.ActualAccuracy / bin.PredictedAccuracy
				adjustments[bin.BinIndex] = adjustmentFactor
				cc.confidenceAdjustments[bin.BinIndex] = adjustmentFactor
			}
		}
	}

	// Calculate recommended threshold to achieve target accuracy
	recommendedThreshold := cc.calculateRecommendedThreshold(bins)

	// Check if calibration is good
	isCalibrated := cc.isWellCalibrated(bins)

	result := &CalibrationResult{
		IsCalibrated:        isCalibrated,
		OverallAccuracy:     overallAccuracy,
		TargetAccuracy:      cc.targetAccuracy,
		CalibrationBins:     bins,
		RecommendedThreshold: recommendedThreshold,
		Adjustments:         adjustments,
		LastCalibration:     time.Now(),
	}

	cc.lastCalibration = time.Now()

	cc.logger.Printf("✅ [Calibration] Calibration complete: overall accuracy=%.2f%%, target=%.2f%%, threshold=%.2f, calibrated=%v",
		overallAccuracy*100, cc.targetAccuracy*100, recommendedThreshold, isCalibrated)

	return result, nil
}

// AdjustConfidence adjusts a confidence score based on calibration data
func (cc *ConfidenceCalibrator) AdjustConfidence(confidence float64) float64 {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	// Clamp confidence to [0.0, 1.0]
	confidence = math.Max(0.0, math.Min(1.0, confidence))

	// Determine bin index
	binIndex := int(confidence / cc.binSize)
	if binIndex >= cc.numBins {
		binIndex = cc.numBins - 1
	}

	// Apply adjustment if available
	if adjustment, exists := cc.confidenceAdjustments[binIndex]; exists {
		adjusted := confidence * adjustment
		return math.Max(0.0, math.Min(1.0, adjusted))
	}

	return confidence
}

// GetCalibrationResult returns current calibration status
func (cc *ConfidenceCalibrator) GetCalibrationResult(ctx context.Context) (*CalibrationResult, error) {
	return cc.Calibrate(ctx)
}

// calculateRecommendedThreshold calculates the minimum confidence threshold to achieve target accuracy
func (cc *ConfidenceCalibrator) calculateRecommendedThreshold(bins []CalibrationBin) float64 {
	if len(bins) == 0 {
		return 0.5 // Default threshold
	}

	// Sort bins by bin index (confidence level)
	sort.Slice(bins, func(i, j int) bool {
		return bins[i].BinIndex > bins[j].BinIndex // Descending order
	})

	// Find the lowest confidence bin that meets target accuracy
	for _, bin := range bins {
		if bin.SampleSize >= 10 && bin.ActualAccuracy >= cc.targetAccuracy {
			// Return the lower bound of this bin
			return float64(bin.BinIndex) * cc.binSize
		}
	}

	// If no bin meets target, return the highest confidence bin's lower bound
	if len(bins) > 0 {
		highestBin := bins[0]
		return float64(highestBin.BinIndex) * cc.binSize
	}

	return 0.5 // Default fallback
}

// isWellCalibrated checks if the calibration is good (low calibration error)
func (cc *ConfidenceCalibrator) isWellCalibrated(bins []CalibrationBin) bool {
	if len(bins) == 0 {
		return false
	}

	// Check if average calibration error is low
	totalError := 0.0
	totalSamples := int64(0)

	for _, bin := range bins {
		if bin.SampleSize >= 10 {
			totalError += bin.CalibrationError * float64(bin.SampleSize)
			totalSamples += bin.SampleSize
		}
	}

	if totalSamples == 0 {
		return false
	}

	avgError := totalError / float64(totalSamples)
	
	// Consider well-calibrated if average error is less than 0.1 (10%)
	return avgError < 0.1
}

// GetStatistics returns calibration statistics
func (cc *ConfidenceCalibrator) GetStatistics() map[string]interface{} {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	overallAccuracy := 0.0
	if cc.totalClassifications > 0 {
		overallAccuracy = float64(cc.correctClassifications) / float64(cc.totalClassifications)
	}

	stats := map[string]interface{}{
		"total_classifications":   cc.totalClassifications,
		"correct_classifications": cc.correctClassifications,
		"overall_accuracy":        overallAccuracy,
		"target_accuracy":         cc.targetAccuracy,
		"num_bins":               cc.numBins,
		"last_calibration":       cc.lastCalibration,
		"bins_with_data":         len(cc.calibrationBins),
	}

	return stats
}

// SetTargetAccuracy sets the target accuracy
func (cc *ConfidenceCalibrator) SetTargetAccuracy(target float64) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.targetAccuracy = math.Max(0.0, math.Min(1.0, target))
	cc.logger.Printf("📊 [Calibration] Target accuracy set to %.2f%%", cc.targetAccuracy*100)
}

// GetTargetAccuracy returns the target accuracy
func (cc *ConfidenceCalibrator) GetTargetAccuracy() float64 {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.targetAccuracy
}

// ShouldRecalibrate checks if recalibration is needed
func (cc *ConfidenceCalibrator) ShouldRecalibrate() bool {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	if cc.lastCalibration.IsZero() {
		return true // Never calibrated
	}

	return time.Since(cc.lastCalibration) >= cc.calibrationInterval
}

// GetRecommendedConfidenceThreshold returns the recommended minimum confidence threshold
func (cc *ConfidenceCalibrator) GetRecommendedConfidenceThreshold(ctx context.Context) (float64, error) {
	result, err := cc.Calibrate(ctx)
	if err != nil {
		return 0.5, err // Default fallback
	}

	return result.RecommendedThreshold, nil
}

