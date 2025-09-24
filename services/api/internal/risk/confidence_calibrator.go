package risk

import (
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
)

// ConfidenceCalibrator calibrates confidence scores based on historical accuracy
type ConfidenceCalibrator struct {
	logger *zap.Logger
}

// CalibrateConfidence calibrates confidence score based on historical accuracy
func (cc *ConfidenceCalibrator) CalibrateConfidence(factorID string, baseConfidence float64, historicalData []HistoricalDataPoint) (*ConfidenceCalibration, error) {
	if len(historicalData) == 0 {
		// No historical data available, return base confidence
		return &ConfidenceCalibration{
			CalibratedConfidence: baseConfidence,
			CalibrationFactor:    1.0,
			HistoricalAccuracy:   0.0,
			CalibrationMethod:    "no_historical_data",
		}, nil
	}

	// Calculate historical accuracy
	historicalAccuracy, err := cc.calculateHistoricalAccuracy(historicalData)
	if err != nil {
		cc.logger.Warn("Failed to calculate historical accuracy",
			zap.String("factor_id", factorID),
			zap.Error(err))
		return &ConfidenceCalibration{
			CalibratedConfidence: baseConfidence,
			CalibrationFactor:    1.0,
			HistoricalAccuracy:   0.0,
			CalibrationMethod:    "error_in_calculation",
		}, nil
	}

	// Calculate calibration factor
	calibrationFactor := cc.calculateCalibrationFactor(baseConfidence, historicalAccuracy)

	// Apply calibration
	calibratedConfidence := cc.applyCalibration(baseConfidence, calibrationFactor)

	// Determine calibration method
	calibrationMethod := cc.determineCalibrationMethod(historicalAccuracy, len(historicalData))

	return &ConfidenceCalibration{
		CalibratedConfidence: calibratedConfidence,
		CalibrationFactor:    calibrationFactor,
		HistoricalAccuracy:   historicalAccuracy,
		CalibrationMethod:    calibrationMethod,
	}, nil
}

// calculateHistoricalAccuracy calculates accuracy based on historical data consistency
func (cc *ConfidenceCalibrator) calculateHistoricalAccuracy(historicalData []HistoricalDataPoint) (float64, error) {
	if len(historicalData) < 2 {
		return 0.0, fmt.Errorf("insufficient historical data for accuracy calculation")
	}

	// Extract values and timestamps
	values := make([]float64, len(historicalData))
	timestamps := make([]time.Time, len(historicalData))

	for i, point := range historicalData {
		values[i] = point.Value
		timestamps[i] = point.Timestamp
	}

	// Calculate multiple accuracy metrics
	consistencyAccuracy := cc.calculateConsistencyAccuracy(values)
	trendAccuracy := cc.calculateTrendAccuracy(values, timestamps)
	volatilityAccuracy := cc.calculateVolatilityAccuracy(values)
	dataQualityAccuracy := cc.calculateDataQualityAccuracy(historicalData)

	// Weighted average of different accuracy metrics
	accuracy := (consistencyAccuracy*0.3 + trendAccuracy*0.3 + volatilityAccuracy*0.2 + dataQualityAccuracy*0.2)

	return math.Max(0, math.Min(1, accuracy)), nil
}

// calculateConsistencyAccuracy calculates accuracy based on data consistency
func (cc *ConfidenceCalibrator) calculateConsistencyAccuracy(values []float64) float64 {
	if len(values) < 2 {
		return 0.0
	}

	// Calculate coefficient of variation (lower is better)
	mean := cc.calculateMean(values)
	if mean == 0 {
		return 0.0
	}

	stdDev := cc.calculateStdDev(values, mean)
	coefficientOfVariation := stdDev / math.Abs(mean)

	// Convert to accuracy score (lower CV = higher accuracy)
	// CV of 0.1 = 90% accuracy, CV of 0.5 = 50% accuracy, etc.
	accuracy := math.Max(0, 1.0-coefficientOfVariation)

	return accuracy
}

// calculateTrendAccuracy calculates accuracy based on trend stability
func (cc *ConfidenceCalibrator) calculateTrendAccuracy(values []float64, timestamps []time.Time) float64 {
	if len(values) < 3 {
		return 0.5 // Default moderate accuracy for insufficient data
	}

	// Calculate trend stability using linear regression
	slope, r2Score, err := cc.calculateLinearRegression(values, timestamps)
	if err != nil {
		return 0.5 // Default if calculation fails
	}

	// Higher R² score indicates more stable trend = higher accuracy
	// But also consider slope - very steep trends might indicate instability
	trendStability := r2Score

	// Penalize very steep trends
	slopePenalty := math.Min(1.0, math.Abs(slope)/10.0) // Normalize slope
	trendStability *= (1.0 - slopePenalty)

	return math.Max(0, math.Min(1, trendStability))
}

// calculateVolatilityAccuracy calculates accuracy based on volatility
func (cc *ConfidenceCalibrator) calculateVolatilityAccuracy(values []float64) float64 {
	if len(values) < 2 {
		return 0.5
	}

	// Calculate volatility as standard deviation
	mean := cc.calculateMean(values)
	stdDev := cc.calculateStdDev(values, mean)

	// Normalize volatility (assume values are in 0-100 range)
	normalizedVolatility := stdDev / 100.0

	// Convert to accuracy (lower volatility = higher accuracy)
	accuracy := math.Max(0, 1.0-normalizedVolatility)

	return accuracy
}

// calculateDataQualityAccuracy calculates accuracy based on data quality indicators
func (cc *ConfidenceCalibrator) calculateDataQualityAccuracy(historicalData []HistoricalDataPoint) float64 {
	if len(historicalData) == 0 {
		return 0.0
	}

	// Check data completeness
	completeDataPoints := 0
	for _, point := range historicalData {
		if !point.Timestamp.IsZero() && !math.IsNaN(point.Value) && !math.IsInf(point.Value, 0) {
			completeDataPoints++
		}
	}

	completenessScore := float64(completeDataPoints) / float64(len(historicalData))

	// Check data freshness
	freshDataPoints := 0
	now := time.Now()
	for _, point := range historicalData {
		if point.Timestamp.After(now.Add(-30 * 24 * time.Hour)) { // Within last 30 days
			freshDataPoints++
		}
	}

	freshnessScore := float64(freshDataPoints) / float64(len(historicalData))

	// Check data source diversity
	sourceCount := make(map[string]int)
	for _, point := range historicalData {
		sourceCount[point.Source]++
	}

	diversityScore := math.Min(1.0, float64(len(sourceCount))/3.0) // Normalize to max 3 sources

	// Weighted average of quality indicators
	qualityScore := (completenessScore*0.5 + freshnessScore*0.3 + diversityScore*0.2)

	return qualityScore
}

// calculateCalibrationFactor calculates the factor to apply to base confidence
func (cc *ConfidenceCalibrator) calculateCalibrationFactor(baseConfidence, historicalAccuracy float64) float64 {
	// If historical accuracy is higher than base confidence, increase confidence
	// If historical accuracy is lower than base confidence, decrease confidence

	if historicalAccuracy > baseConfidence {
		// Historical data is more reliable than base confidence suggests
		// Increase confidence, but not too much
		increase := (historicalAccuracy - baseConfidence) * 0.5
		return 1.0 + increase
	} else {
		// Historical data is less reliable than base confidence suggests
		// Decrease confidence
		decrease := (baseConfidence - historicalAccuracy) * 0.7
		return 1.0 - decrease
	}
}

// applyCalibration applies the calibration factor to base confidence
func (cc *ConfidenceCalibrator) applyCalibration(baseConfidence, calibrationFactor float64) float64 {
	calibratedConfidence := baseConfidence * calibrationFactor

	// Ensure confidence stays within bounds
	return math.Max(0.1, math.Min(1.0, calibratedConfidence))
}

// determineCalibrationMethod determines the method used for calibration
func (cc *ConfidenceCalibrator) determineCalibrationMethod(historicalAccuracy float64, dataPoints int) string {
	if dataPoints < 3 {
		return "insufficient_data"
	} else if historicalAccuracy > 0.8 {
		return "high_accuracy_calibration"
	} else if historicalAccuracy > 0.6 {
		return "moderate_accuracy_calibration"
	} else if historicalAccuracy > 0.4 {
		return "low_accuracy_calibration"
	} else {
		return "very_low_accuracy_calibration"
	}
}

// calculateLinearRegression calculates linear regression for trend analysis
func (cc *ConfidenceCalibrator) calculateLinearRegression(values []float64, timestamps []time.Time) (slope, r2Score float64, err error) {
	if len(values) != len(timestamps) || len(values) < 2 {
		return 0, 0, fmt.Errorf("invalid data for linear regression")
	}

	// Convert timestamps to numeric values (days since first timestamp)
	startTime := timestamps[0]
	xValues := make([]float64, len(timestamps))
	for i, t := range timestamps {
		xValues[i] = t.Sub(startTime).Hours() / 24.0 // Convert to days
	}

	// Calculate means
	xMean := cc.calculateMean(xValues)
	yMean := cc.calculateMean(values)

	// Calculate slope and intercept
	var numerator, denominator float64
	for i := 0; i < len(xValues); i++ {
		numerator += (xValues[i] - xMean) * (values[i] - yMean)
		denominator += (xValues[i] - xMean) * (xValues[i] - xMean)
	}

	if denominator == 0 {
		return 0, 0, fmt.Errorf("cannot calculate slope: denominator is zero")
	}

	slope = numerator / denominator
	intercept := yMean - slope*xMean

	// Calculate R² score
	r2Score = cc.calculateRSquared(values, xValues, slope, intercept)

	return slope, r2Score, nil
}

// calculateRSquared calculates R² score for regression
func (cc *ConfidenceCalibrator) calculateRSquared(yValues, xValues []float64, slope, intercept float64) float64 {
	if len(yValues) != len(xValues) || len(yValues) < 2 {
		return 0
	}

	yMean := cc.calculateMean(yValues)

	var ssRes, ssTot float64
	for i := 0; i < len(yValues); i++ {
		// Predicted value
		predicted := slope*xValues[i] + intercept

		// Residual sum of squares
		ssRes += (yValues[i] - predicted) * (yValues[i] - predicted)

		// Total sum of squares
		ssTot += (yValues[i] - yMean) * (yValues[i] - yMean)
	}

	if ssTot == 0 {
		return 0
	}

	return 1 - (ssRes / ssTot)
}

// Helper functions for statistical calculations
func (cc *ConfidenceCalibrator) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (cc *ConfidenceCalibrator) calculateStdDev(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0
	}

	sumSquaredDiffs := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquaredDiffs += diff * diff
	}

	variance := sumSquaredDiffs / float64(len(values)-1)
	return math.Sqrt(variance)
}

// CalibrateConfidenceForFactor calibrates confidence for a specific factor with custom parameters
func (cc *ConfidenceCalibrator) CalibrateConfidenceForFactor(factorID string, baseConfidence float64, historicalData []HistoricalDataPoint, customParams map[string]interface{}) (*ConfidenceCalibration, error) {
	// Apply custom parameters if provided
	calibration, err := cc.CalibrateConfidence(factorID, baseConfidence, historicalData)
	if err != nil {
		return nil, err
	}

	// Apply custom adjustments
	if customParams != nil {
		if factor, exists := customParams["factor_multiplier"]; exists {
			if multiplier, ok := factor.(float64); ok {
				calibration.CalibratedConfidence *= multiplier
				calibration.CalibratedConfidence = math.Max(0.1, math.Min(1.0, calibration.CalibratedConfidence))
			}
		}

		if method, exists := customParams["calibration_method"]; exists {
			if methodStr, ok := method.(string); ok {
				calibration.CalibrationMethod = methodStr
			}
		}
	}

	return calibration, nil
}
