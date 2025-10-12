package training

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// ModelCalibrator provides model calibration capabilities to improve prediction accuracy
type ModelCalibrator struct {
	logger *zap.Logger
}

// CalibrationPoint represents a calibration data point
type CalibrationPoint struct {
	PredictedProbability float64
	ActualOutcome        float64
	Weight               float64
}

// CalibrationConfig holds configuration for model calibration
type CalibrationConfig struct {
	NumBins              int     `json:"num_bins"`
	MinSamplesPerBin     int     `json:"min_samples_per_bin"`
	RegularizationFactor float64 `json:"regularization_factor"`
	MaxIterations        int     `json:"max_iterations"`
	ConvergenceThreshold float64 `json:"convergence_threshold"`
}

// CalibrationResult contains the results of model calibration
type CalibrationResult struct {
	CalibratedModel    map[string]interface{} `json:"calibrated_model"`
	CalibrationCurve   []CalibrationPoint     `json:"calibration_curve"`
	ReliabilityDiagram []ReliabilityPoint     `json:"reliability_diagram"`
	CalibrationMetrics CalibrationMetrics     `json:"calibration_metrics"`
	ImprovementMetrics CalibrationImprovement `json:"improvement_metrics"`
	CalibrationTime    time.Duration          `json:"calibration_time"`
	NumSamples         int                    `json:"num_samples"`
	ValidationScore    float64                `json:"validation_score"`
}

// ReliabilityPoint represents a point in the reliability diagram
type ReliabilityPoint struct {
	BinCenter          float64    `json:"bin_center"`
	BinCount           int        `json:"bin_count"`
	AveragePrediction  float64    `json:"average_prediction"`
	AverageActual      float64    `json:"average_actual"`
	CalibrationError   float64    `json:"calibration_error"`
	ConfidenceInterval [2]float64 `json:"confidence_interval"`
}

// CalibrationMetrics contains various calibration metrics
type CalibrationMetrics struct {
	ExpectedCalibrationError float64 `json:"expected_calibration_error"`
	MaximumCalibrationError  float64 `json:"maximum_calibration_error"`
	ReliabilityScore         float64 `json:"reliability_score"`
	SharpnessScore           float64 `json:"sharpness_score"`
	BrierScore               float64 `json:"brier_score"`
	LogLoss                  float64 `json:"log_loss"`
	CalibrationSlope         float64 `json:"calibration_slope"`
	CalibrationIntercept     float64 `json:"calibration_intercept"`
}

// CalibrationImprovement shows improvement after calibration
type CalibrationImprovement struct {
	ECEReduction       float64 `json:"ece_reduction"`
	BrierImprovement   float64 `json:"brier_improvement"`
	LogLossImprovement float64 `json:"log_loss_improvement"`
	ReliabilityGain    float64 `json:"reliability_gain"`
	SharpnessGain      float64 `json:"sharpness_gain"`
}

// NewModelCalibrator creates a new model calibrator
func NewModelCalibrator(logger *zap.Logger) *ModelCalibrator {
	return &ModelCalibrator{
		logger: logger,
	}
}

// CalibrateModel calibrates a model using isotonic regression
func (mc *ModelCalibrator) CalibrateModel(ctx context.Context, config CalibrationConfig, predictions []models.RiskAssessment, actuals []float64) (*CalibrationResult, error) {
	startTime := time.Now()

	if len(predictions) != len(actuals) {
		return nil, fmt.Errorf("predictions and actuals must have the same length")
	}

	if len(predictions) == 0 {
		return nil, fmt.Errorf("no data provided for calibration")
	}

	mc.logger.Info("Starting model calibration",
		zap.Int("num_samples", len(predictions)),
		zap.Int("num_bins", config.NumBins))

	// Extract predicted probabilities
	predictedProbs := make([]float64, len(predictions))
	for i, pred := range predictions {
		predictedProbs[i] = pred.RiskScore
	}

	// Calculate baseline metrics
	baselineMetrics := mc.calculateCalibrationMetrics(predictedProbs, actuals)

	// Perform isotonic regression calibration
	calibratedProbs, calibrationCurve := mc.performIsotonicRegression(predictedProbs, actuals, config)

	// Calculate calibrated metrics
	calibratedMetrics := mc.calculateCalibrationMetrics(calibratedProbs, actuals)

	// Generate reliability diagram
	reliabilityDiagram := mc.generateReliabilityDiagram(calibratedProbs, actuals, config.NumBins)

	// Calculate improvement metrics
	improvement := CalibrationImprovement{
		ECEReduction:       baselineMetrics.ExpectedCalibrationError - calibratedMetrics.ExpectedCalibrationError,
		BrierImprovement:   baselineMetrics.BrierScore - calibratedMetrics.BrierScore,
		LogLossImprovement: baselineMetrics.LogLoss - calibratedMetrics.LogLoss,
		ReliabilityGain:    calibratedMetrics.ReliabilityScore - baselineMetrics.ReliabilityScore,
		SharpnessGain:      calibratedMetrics.SharpnessScore - baselineMetrics.SharpnessScore,
	}

	// Create calibration model (simplified representation)
	calibratedModel := map[string]interface{}{
		"calibration_type":   "isotonic_regression",
		"num_bins":           config.NumBins,
		"calibration_curve":  calibrationCurve,
		"baseline_metrics":   baselineMetrics,
		"calibrated_metrics": calibratedMetrics,
		"improvement":        improvement,
	}

	result := &CalibrationResult{
		CalibratedModel:    calibratedModel,
		CalibrationCurve:   calibrationCurve,
		ReliabilityDiagram: reliabilityDiagram,
		CalibrationMetrics: calibratedMetrics,
		ImprovementMetrics: improvement,
		CalibrationTime:    time.Since(startTime),
		NumSamples:         len(predictions),
		ValidationScore:    calibratedMetrics.ReliabilityScore,
	}

	mc.logger.Info("Model calibration completed",
		zap.Duration("calibration_time", result.CalibrationTime),
		zap.Float64("baseline_ece", baselineMetrics.ExpectedCalibrationError),
		zap.Float64("calibrated_ece", calibratedMetrics.ExpectedCalibrationError),
		zap.Float64("ece_improvement", improvement.ECEReduction),
		zap.Float64("validation_score", result.ValidationScore))

	return result, nil
}

// performIsotonicRegression performs isotonic regression for calibration
func (mc *ModelCalibrator) performIsotonicRegression(predictions, actuals []float64, config CalibrationConfig) ([]float64, []CalibrationPoint) {
	// Create calibration points
	points := make([]CalibrationPoint, len(predictions))
	for i := range predictions {
		points[i] = CalibrationPoint{
			PredictedProbability: predictions[i],
			ActualOutcome:        actuals[i],
			Weight:               1.0,
		}
	}

	// Sort by predicted probability
	sort.Slice(points, func(i, j int) bool {
		return points[i].PredictedProbability < points[j].PredictedProbability
	})

	// Apply isotonic regression (simplified implementation)
	calibratedProbs := make([]float64, len(points))
	calibrationCurve := make([]CalibrationPoint, 0)

	// Group into bins for calibration curve
	binSize := len(points) / config.NumBins
	if binSize < 1 {
		binSize = 1
	}

	for i := 0; i < len(points); i += binSize {
		end := i + binSize
		if end > len(points) {
			end = len(points)
		}

		bin := points[i:end]

		// Calculate average actual outcome for this bin
		var sumActual float64
		for _, point := range bin {
			sumActual += point.ActualOutcome
		}
		avgActual := sumActual / float64(len(bin))

		// Calculate average predicted probability for this bin
		var sumPred float64
		for _, point := range bin {
			sumPred += point.PredictedProbability
		}
		avgPred := sumPred / float64(len(bin))

		// Apply calibration (simplified isotonic regression)
		calibratedProb := mc.applyIsotonicCalibration(avgPred, avgActual)

		// Update calibrated probabilities for this bin
		for j := i; j < end; j++ {
			calibratedProbs[j] = calibratedProb
		}

		// Add to calibration curve
		calibrationCurve = append(calibrationCurve, CalibrationPoint{
			PredictedProbability: avgPred,
			ActualOutcome:        avgActual,
			Weight:               float64(len(bin)),
		})
	}

	return calibratedProbs, calibrationCurve
}

// applyIsotonicCalibration applies isotonic calibration to a probability
func (mc *ModelCalibrator) applyIsotonicCalibration(predicted, actual float64) float64 {
	// Simplified isotonic regression implementation
	// In practice, this would use a more sophisticated algorithm

	// Apply smoothing to reduce overfitting
	alpha := 0.1 // Smoothing factor
	calibrated := alpha*actual + (1-alpha)*predicted

	// Ensure probability is in valid range
	if calibrated < 0 {
		calibrated = 0
	}
	if calibrated > 1 {
		calibrated = 1
	}

	return calibrated
}

// calculateCalibrationMetrics calculates various calibration metrics
func (mc *ModelCalibrator) calculateCalibrationMetrics(predictions, actuals []float64) CalibrationMetrics {
	if len(predictions) != len(actuals) || len(predictions) == 0 {
		return CalibrationMetrics{}
	}

	// Calculate Expected Calibration Error (ECE)
	ece := mc.calculateECE(predictions, actuals, 10)

	// Calculate Maximum Calibration Error (MCE)
	mce := mc.calculateMCE(predictions, actuals, 10)

	// Calculate Brier Score
	brierScore := mc.calculateBrierScore(predictions, actuals)

	// Calculate Log Loss
	logLoss := mc.calculateLogLoss(predictions, actuals)

	// Calculate Reliability Score
	reliabilityScore := mc.calculateReliabilityScore(predictions, actuals)

	// Calculate Sharpness Score
	sharpnessScore := mc.calculateSharpnessScore(predictions)

	// Calculate calibration slope and intercept (simplified)
	slope, intercept := mc.calculateCalibrationSlopeIntercept(predictions, actuals)

	return CalibrationMetrics{
		ExpectedCalibrationError: ece,
		MaximumCalibrationError:  mce,
		ReliabilityScore:         reliabilityScore,
		SharpnessScore:           sharpnessScore,
		BrierScore:               brierScore,
		LogLoss:                  logLoss,
		CalibrationSlope:         slope,
		CalibrationIntercept:     intercept,
	}
}

// calculateECE calculates Expected Calibration Error
func (mc *ModelCalibrator) calculateECE(predictions, actuals []float64, numBins int) float64 {
	if len(predictions) == 0 {
		return 0
	}

	// Create bins
	bins := make([][]int, numBins)
	for i := range bins {
		bins[i] = make([]int, 0)
	}

	// Assign predictions to bins
	for i, pred := range predictions {
		binIndex := int(pred * float64(numBins))
		if binIndex >= numBins {
			binIndex = numBins - 1
		}
		bins[binIndex] = append(bins[binIndex], i)
	}

	// Calculate ECE
	var ece float64
	totalSamples := float64(len(predictions))

	for _, bin := range bins {
		if len(bin) == 0 {
			continue
		}

		// Calculate average prediction and actual for this bin
		var sumPred, sumActual float64
		for _, idx := range bin {
			sumPred += predictions[idx]
			sumActual += actuals[idx]
		}

		avgPred := sumPred / float64(len(bin))
		avgActual := sumActual / float64(len(bin))

		// Calculate bin weight and error
		binWeight := float64(len(bin)) / totalSamples
		binError := math.Abs(avgPred - avgActual)

		ece += binWeight * binError
	}

	return ece
}

// calculateMCE calculates Maximum Calibration Error
func (mc *ModelCalibrator) calculateMCE(predictions, actuals []float64, numBins int) float64 {
	if len(predictions) == 0 {
		return 0
	}

	// Create bins
	bins := make([][]int, numBins)
	for i := range bins {
		bins[i] = make([]int, 0)
	}

	// Assign predictions to bins
	for i, pred := range predictions {
		binIndex := int(pred * float64(numBins))
		if binIndex >= numBins {
			binIndex = numBins - 1
		}
		bins[binIndex] = append(bins[binIndex], i)
	}

	// Calculate MCE
	var mce float64

	for _, bin := range bins {
		if len(bin) == 0 {
			continue
		}

		// Calculate average prediction and actual for this bin
		var sumPred, sumActual float64
		for _, idx := range bin {
			sumPred += predictions[idx]
			sumActual += actuals[idx]
		}

		avgPred := sumPred / float64(len(bin))
		avgActual := sumActual / float64(len(bin))

		// Calculate bin error
		binError := math.Abs(avgPred - avgActual)

		if binError > mce {
			mce = binError
		}
	}

	return mce
}

// calculateBrierScore calculates Brier Score
func (mc *ModelCalibrator) calculateBrierScore(predictions, actuals []float64) float64 {
	if len(predictions) == 0 {
		return 0
	}

	var brierScore float64
	for i := range predictions {
		error := predictions[i] - actuals[i]
		brierScore += error * error
	}

	return brierScore / float64(len(predictions))
}

// calculateLogLoss calculates Log Loss
func (mc *ModelCalibrator) calculateLogLoss(predictions, actuals []float64) float64 {
	if len(predictions) == 0 {
		return 0
	}

	var logLoss float64
	for i := range predictions {
		pred := predictions[i]
		actual := actuals[i]

		// Avoid log(0) by adding small epsilon
		epsilon := 1e-15
		if pred < epsilon {
			pred = epsilon
		}
		if pred > 1-epsilon {
			pred = 1 - epsilon
		}

		logLoss += actual*math.Log(pred) + (1-actual)*math.Log(1-pred)
	}

	return -logLoss / float64(len(predictions))
}

// calculateReliabilityScore calculates Reliability Score
func (mc *ModelCalibrator) calculateReliabilityScore(predictions, actuals []float64) float64 {
	// Reliability score is 1 - ECE
	ece := mc.calculateECE(predictions, actuals, 10)
	return 1 - ece
}

// calculateSharpnessScore calculates Sharpness Score
func (mc *ModelCalibrator) calculateSharpnessScore(predictions []float64) float64 {
	if len(predictions) == 0 {
		return 0
	}

	// Calculate variance of predictions
	var mean, variance float64

	// Calculate mean
	for _, pred := range predictions {
		mean += pred
	}
	mean /= float64(len(predictions))

	// Calculate variance
	for _, pred := range predictions {
		diff := pred - mean
		variance += diff * diff
	}
	variance /= float64(len(predictions))

	// Sharpness score is the variance
	return variance
}

// calculateCalibrationSlopeIntercept calculates calibration slope and intercept
func (mc *ModelCalibrator) calculateCalibrationSlopeIntercept(predictions, actuals []float64) (float64, float64) {
	if len(predictions) < 2 {
		return 1.0, 0.0
	}

	// Simple linear regression to find slope and intercept
	var sumX, sumY, sumXY, sumXX float64
	n := float64(len(predictions))

	for i := range predictions {
		x := predictions[i]
		y := actuals[i]

		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	// Calculate slope and intercept
	slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	return slope, intercept
}

// generateReliabilityDiagram generates reliability diagram data
func (mc *ModelCalibrator) generateReliabilityDiagram(predictions, actuals []float64, numBins int) []ReliabilityPoint {
	if len(predictions) == 0 {
		return nil
	}

	// Create bins
	bins := make([][]int, numBins)
	for i := range bins {
		bins[i] = make([]int, 0)
	}

	// Assign predictions to bins
	for i, pred := range predictions {
		binIndex := int(pred * float64(numBins))
		if binIndex >= numBins {
			binIndex = numBins - 1
		}
		bins[binIndex] = append(bins[binIndex], i)
	}

	// Generate reliability points
	reliabilityPoints := make([]ReliabilityPoint, 0, numBins)

	for binIndex, bin := range bins {
		if len(bin) == 0 {
			continue
		}

		// Calculate bin center
		binCenter := (float64(binIndex) + 0.5) / float64(numBins)

		// Calculate average prediction and actual for this bin
		var sumPred, sumActual float64
		for _, idx := range bin {
			sumPred += predictions[idx]
			sumActual += actuals[idx]
		}

		avgPred := sumPred / float64(len(bin))
		avgActual := sumActual / float64(len(bin))

		// Calculate calibration error
		calibrationError := math.Abs(avgPred - avgActual)

		// Calculate confidence interval (simplified)
		confidenceInterval := [2]float64{
			avgActual - 1.96*math.Sqrt(avgActual*(1-avgActual)/float64(len(bin))),
			avgActual + 1.96*math.Sqrt(avgActual*(1-avgActual)/float64(len(bin))),
		}

		// Ensure confidence interval is in valid range
		if confidenceInterval[0] < 0 {
			confidenceInterval[0] = 0
		}
		if confidenceInterval[1] > 1 {
			confidenceInterval[1] = 1
		}

		reliabilityPoints = append(reliabilityPoints, ReliabilityPoint{
			BinCenter:          binCenter,
			BinCount:           len(bin),
			AveragePrediction:  avgPred,
			AverageActual:      avgActual,
			CalibrationError:   calibrationError,
			ConfidenceInterval: confidenceInterval,
		})
	}

	return reliabilityPoints
}

// ValidateCalibration validates the calibration quality
func (mc *ModelCalibrator) ValidateCalibration(result *CalibrationResult) bool {
	// Check if calibration improved the model
	improvement := result.ImprovementMetrics

	// Consider calibration successful if:
	// 1. ECE was reduced
	// 2. Brier score improved
	// 3. Reliability score improved
	success := improvement.ECEReduction > 0 &&
		improvement.BrierImprovement > 0 &&
		improvement.ReliabilityGain > 0

	mc.logger.Info("Calibration validation",
		zap.Bool("success", success),
		zap.Float64("ece_reduction", improvement.ECEReduction),
		zap.Float64("brier_improvement", improvement.BrierImprovement),
		zap.Float64("reliability_gain", improvement.ReliabilityGain))

	return success
}
