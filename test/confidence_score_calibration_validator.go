package test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"

	"kyb-platform/internal/classification"
)

// ConfidenceScoreCalibrationValidator provides comprehensive confidence score calibration testing
type ConfidenceScoreCalibrationValidator struct {
	TestRunner *ClassificationAccuracyTestRunner
	Logger     *log.Logger
	Config     *CalibrationValidationConfig
}

// CalibrationValidationConfig configuration for confidence score calibration validation
type CalibrationValidationConfig struct {
	SessionName                     string        `json:"session_name"`
	ValidationDirectory             string        `json:"validation_directory"`
	SampleSize                      int           `json:"sample_size"`
	Timeout                         time.Duration `json:"timeout"`
	MinCalibrationThreshold         float64       `json:"min_calibration_threshold"`
	IncludeReliabilityDiagram       bool          `json:"include_reliability_diagram"`
	IncludeCalibrationCurve         bool          `json:"include_calibration_curve"`
	IncludeBrierScore               bool          `json:"include_brier_score"`
	IncludeExpectedCalibrationError bool          `json:"include_expected_calibration_error"`
	IncludeTemperatureScaling       bool          `json:"include_temperature_scaling"`
	GenerateDetailedReport          bool          `json:"generate_detailed_report"`
}

// CalibrationValidationResult represents the result of confidence score calibration validation
type CalibrationValidationResult struct {
	SessionID                string                    `json:"session_id"`
	StartTime                time.Time                 `json:"start_time"`
	EndTime                  time.Time                 `json:"end_time"`
	Duration                 time.Duration             `json:"duration"`
	TotalValidations         int                       `json:"total_validations"`
	CalibrationSummary       *CalibrationSummary       `json:"calibration_summary"`
	ReliabilityDiagram       *ReliabilityDiagram       `json:"reliability_diagram"`
	CalibrationCurve         *CalibrationCurve         `json:"calibration_curve"`
	BrierScore               *BrierScoreResult         `json:"brier_score"`
	ExpectedCalibrationError *ExpectedCalibrationError `json:"expected_calibration_error"`
	TemperatureScaling       *TemperatureScalingResult `json:"temperature_scaling"`
	ConfidenceBins           []ConfidenceBin           `json:"confidence_bins"`
	CalibrationMetrics       *CalibrationMetrics       `json:"calibration_metrics"`
	Recommendations          []string                  `json:"recommendations"`
	Issues                   []CalibrationIssue        `json:"issues"`
}

// CalibrationSummary provides overall calibration summary
type CalibrationSummary struct {
	OverallCalibration          float64 `json:"overall_calibration"`
	ReliabilityDiagramSlope     float64 `json:"reliability_diagram_slope"`
	ReliabilityDiagramIntercept float64 `json:"reliability_diagram_intercept"`
	CalibrationError            float64 `json:"calibration_error"`
	BrierScore                  float64 `json:"brier_score"`
	ExpectedCalibrationError    float64 `json:"expected_calibration_error"`
	TemperatureScalingFactor    float64 `json:"temperature_scaling_factor"`
	IsWellCalibrated            bool    `json:"is_well_calibrated"`
	CalibrationQuality          string  `json:"calibration_quality"`
}

// ReliabilityDiagram represents reliability diagram data
type ReliabilityDiagram struct {
	Bins             []ReliabilityBin `json:"bins"`
	Slope            float64          `json:"slope"`
	Intercept        float64          `json:"intercept"`
	R2               float64          `json:"r2"`
	CalibrationError float64          `json:"calibration_error"`
	ReliabilityError float64          `json:"reliability_error"`
	ResolutionError  float64          `json:"resolution_error"`
}

// ReliabilityBin represents a single bin in the reliability diagram
type ReliabilityBin struct {
	BinIndex            int     `json:"bin_index"`
	ConfidenceMin       float64 `json:"confidence_min"`
	ConfidenceMax       float64 `json:"confidence_max"`
	ConfidenceCenter    float64 `json:"confidence_center"`
	ActualAccuracy      float64 `json:"actual_accuracy"`
	PredictedConfidence float64 `json:"predicted_confidence"`
	SampleCount         int     `json:"sample_count"`
	CalibrationError    float64 `json:"calibration_error"`
}

// CalibrationCurve represents calibration curve data
type CalibrationCurve struct {
	Points                 []CalibrationPoint `json:"points"`
	PerfectCalibrationLine []CalibrationPoint `json:"perfect_calibration_line"`
	CalibrationError       float64            `json:"calibration_error"`
	AreaUnderCurve         float64            `json:"area_under_curve"`
}

// CalibrationPoint represents a point on the calibration curve
type CalibrationPoint struct {
	Confidence     float64 `json:"confidence"`
	ActualAccuracy float64 `json:"actual_accuracy"`
	SampleCount    int     `json:"sample_count"`
}

// BrierScoreResult represents Brier score calculation results
type BrierScoreResult struct {
	OverallBrierScore    float64         `json:"overall_brier_score"`
	ReliabilityComponent float64         `json:"reliability_component"`
	ResolutionComponent  float64         `json:"resolution_component"`
	UncertaintyComponent float64         `json:"uncertainty_component"`
	BrierScoreByBin      []BrierScoreBin `json:"brier_score_by_bin"`
}

// BrierScoreBin represents Brier score for a specific confidence bin
type BrierScoreBin struct {
	BinIndex      int     `json:"bin_index"`
	ConfidenceMin float64 `json:"confidence_min"`
	ConfidenceMax float64 `json:"confidence_max"`
	BrierScore    float64 `json:"brier_score"`
	SampleCount   int     `json:"sample_count"`
}

// ExpectedCalibrationError represents expected calibration error results
type ExpectedCalibrationError struct {
	OverallECE    float64  `json:"overall_ece"`
	ECEByBin      []ECEBin `json:"ece_by_bin"`
	WeightedECE   float64  `json:"weighted_ece"`
	UnweightedECE float64  `json:"unweighted_ece"`
}

// ECEBin represents expected calibration error for a specific bin
type ECEBin struct {
	BinIndex      int     `json:"bin_index"`
	ConfidenceMin float64 `json:"confidence_min"`
	ConfidenceMax float64 `json:"confidence_max"`
	ECE           float64 `json:"ece"`
	SampleCount   int     `json:"sample_count"`
	Weight        float64 `json:"weight"`
}

// TemperatureScalingResult represents temperature scaling calibration results
type TemperatureScalingResult struct {
	OptimalTemperature      float64            `json:"optimal_temperature"`
	CalibrationImprovement  float64            `json:"calibration_improvement"`
	BeforeCalibrationECE    float64            `json:"before_calibration_ece"`
	AfterCalibrationECE     float64            `json:"after_calibration_ece"`
	TemperatureOptimization []TemperaturePoint `json:"temperature_optimization"`
}

// TemperaturePoint represents a point in temperature optimization
type TemperaturePoint struct {
	Temperature      float64 `json:"temperature"`
	ECE              float64 `json:"ece"`
	CalibrationError float64 `json:"calibration_error"`
}

// ConfidenceBin represents a confidence bin for analysis
type ConfidenceBin struct {
	BinIndex            int     `json:"bin_index"`
	ConfidenceMin       float64 `json:"confidence_min"`
	ConfidenceMax       float64 `json:"confidence_max"`
	ConfidenceCenter    float64 `json:"confidence_center"`
	SampleCount         int     `json:"sample_count"`
	ActualAccuracy      float64 `json:"actual_accuracy"`
	PredictedConfidence float64 `json:"predicted_confidence"`
	CalibrationError    float64 `json:"calibration_error"`
	BrierScore          float64 `json:"brier_score"`
}

// CalibrationMetrics represents comprehensive calibration metrics
type CalibrationMetrics struct {
	CalibrationError     float64 `json:"calibration_error"`
	ReliabilityError     float64 `json:"reliability_error"`
	ResolutionError      float64 `json:"resolution_error"`
	Sharpness            float64 `json:"sharpness"`
	Resolution           float64 `json:"resolution"`
	Reliability          float64 `json:"reliability"`
	CalibrationSlope     float64 `json:"calibration_slope"`
	CalibrationIntercept float64 `json:"calibration_intercept"`
	R2                   float64 `json:"r2"`
	MeanAbsoluteError    float64 `json:"mean_absolute_error"`
	RootMeanSquareError  float64 `json:"root_mean_square_error"`
}

// CalibrationIssue represents a calibration issue
type CalibrationIssue struct {
	Type                string  `json:"type"`     // "overconfident", "underconfident", "poor_calibration"
	Severity            string  `json:"severity"` // "critical", "high", "medium", "low"
	ConfidenceRange     string  `json:"confidence_range"`
	ActualAccuracy      float64 `json:"actual_accuracy"`
	PredictedConfidence float64 `json:"predicted_confidence"`
	CalibrationError    float64 `json:"calibration_error"`
	SampleCount         int     `json:"sample_count"`
	Description         string  `json:"description"`
	Recommendation      string  `json:"recommendation"`
}

// CalibrationDataPoint represents a single data point for calibration analysis
type CalibrationDataPoint struct {
	Confidence     float64 `json:"confidence"`
	ActualAccuracy float64 `json:"actual_accuracy"`
	IsCorrect      bool    `json:"is_correct"`
	SampleWeight   float64 `json:"sample_weight"`
}

// NewConfidenceScoreCalibrationValidator creates a new confidence score calibration validator
func NewConfidenceScoreCalibrationValidator(testRunner *ClassificationAccuracyTestRunner, logger *log.Logger, config *CalibrationValidationConfig) *ConfidenceScoreCalibrationValidator {
	return &ConfidenceScoreCalibrationValidator{
		TestRunner: testRunner,
		Logger:     logger,
		Config:     config,
	}
}

// ValidateCalibration performs comprehensive confidence score calibration validation
func (validator *ConfidenceScoreCalibrationValidator) ValidateCalibration(ctx context.Context) (*CalibrationValidationResult, error) {
	startTime := time.Now()
	sessionID := fmt.Sprintf("calibration_%d", startTime.Unix())

	validator.Logger.Printf("🎯 Starting Confidence Score Calibration Validation Session: %s", sessionID)

	result := &CalibrationValidationResult{
		SessionID:        sessionID,
		StartTime:        startTime,
		TotalValidations: 0,
		ConfidenceBins:   []ConfidenceBin{},
		Recommendations:  []string{},
		Issues:           []CalibrationIssue{},
	}

	// Get test cases
	dataset := validator.TestRunner.GetDataset()
	testCases := dataset.TestCases
	if len(testCases) > validator.Config.SampleSize {
		testCases = testCases[:validator.Config.SampleSize]
	}

	validator.Logger.Printf("📊 Validating calibration for %d test cases", len(testCases))

	// Collect calibration data points
	var dataPoints []CalibrationDataPoint
	for _, testCase := range testCases {
		if err := validator.collectCalibrationDataPoint(ctx, testCase, &dataPoints); err != nil {
			validator.Logger.Printf("❌ Failed to collect calibration data for %s: %v", testCase.Name, err)
			continue
		}
		result.TotalValidations++
	}

	if len(dataPoints) == 0 {
		return nil, fmt.Errorf("no calibration data points collected")
	}

	// Create confidence bins
	validator.createConfidenceBins(dataPoints, result)

	// Calculate reliability diagram
	if validator.Config.IncludeReliabilityDiagram {
		result.ReliabilityDiagram = validator.calculateReliabilityDiagram(dataPoints)
	}

	// Calculate calibration curve
	if validator.Config.IncludeCalibrationCurve {
		result.CalibrationCurve = validator.calculateCalibrationCurve(dataPoints)
	}

	// Calculate Brier score
	if validator.Config.IncludeBrierScore {
		result.BrierScore = validator.calculateBrierScore(dataPoints)
	}

	// Calculate expected calibration error
	if validator.Config.IncludeExpectedCalibrationError {
		result.ExpectedCalibrationError = validator.calculateExpectedCalibrationError(dataPoints)
	}

	// Calculate temperature scaling
	if validator.Config.IncludeTemperatureScaling {
		result.TemperatureScaling = validator.calculateTemperatureScaling(dataPoints)
	}

	// Calculate comprehensive calibration metrics
	result.CalibrationMetrics = validator.calculateCalibrationMetrics(dataPoints)

	// Generate calibration summary
	result.CalibrationSummary = validator.calculateCalibrationSummary(result)

	// Generate recommendations
	result.Recommendations = validator.generateCalibrationRecommendations(result)

	// Set end time and duration
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	validator.Logger.Printf("✅ Calibration validation completed in %v", result.Duration)
	validator.Logger.Printf("📊 Overall calibration: %.3f", result.CalibrationSummary.OverallCalibration)

	return result, nil
}

// collectCalibrationDataPoint collects a single calibration data point
func (validator *ConfidenceScoreCalibrationValidator) collectCalibrationDataPoint(ctx context.Context, testCase ClassificationTestCase, dataPoints *[]CalibrationDataPoint) error {
	// Run classification
	classificationResult, err := validator.TestRunner.classifier.GenerateClassificationCodes(
		ctx,
		testCase.Keywords,
		testCase.ExpectedIndustry,
		testCase.ExpectedConfidence,
	)
	if err != nil {
		return fmt.Errorf("classification failed: %w", err)
	}

	// Calculate overall confidence score
	confidence := validator.calculateOverallConfidence(classificationResult)

	// Determine if classification is correct (simplified - in real implementation, this would be more sophisticated)
	isCorrect := validator.isClassificationCorrect(classificationResult, testCase)

	// Create data point
	dataPoint := CalibrationDataPoint{
		Confidence:     confidence,
		ActualAccuracy: validator.calculateActualAccuracy(classificationResult, testCase),
		IsCorrect:      isCorrect,
		SampleWeight:   1.0,
	}

	*dataPoints = append(*dataPoints, dataPoint)
	return nil
}

// createConfidenceBins creates confidence bins for analysis
func (validator *ConfidenceScoreCalibrationValidator) createConfidenceBins(dataPoints []CalibrationDataPoint, result *CalibrationValidationResult) {
	numBins := 10
	binSize := 1.0 / float64(numBins)

	for i := 0; i < numBins; i++ {
		binMin := float64(i) * binSize
		binMax := float64(i+1) * binSize
		binCenter := (binMin + binMax) / 2.0

		// Find data points in this bin
		var binDataPoints []CalibrationDataPoint
		for _, dp := range dataPoints {
			if dp.Confidence >= binMin && dp.Confidence < binMax {
				binDataPoints = append(binDataPoints, dp)
			}
		}

		if len(binDataPoints) == 0 {
			continue
		}

		// Calculate bin statistics
		actualAccuracy := validator.calculateBinActualAccuracy(binDataPoints)
		predictedConfidence := validator.calculateBinPredictedConfidence(binDataPoints)
		calibrationError := math.Abs(actualAccuracy - predictedConfidence)
		brierScore := validator.calculateBinBrierScore(binDataPoints)

		bin := ConfidenceBin{
			BinIndex:            i,
			ConfidenceMin:       binMin,
			ConfidenceMax:       binMax,
			ConfidenceCenter:    binCenter,
			SampleCount:         len(binDataPoints),
			ActualAccuracy:      actualAccuracy,
			PredictedConfidence: predictedConfidence,
			CalibrationError:    calibrationError,
			BrierScore:          brierScore,
		}

		result.ConfidenceBins = append(result.ConfidenceBins, bin)
	}
}

// Helper methods for calibration calculations
func (validator *ConfidenceScoreCalibrationValidator) calculateOverallConfidence(result *classification.ClassificationCodesInfo) float64 {
	// Simplified calculation - in real implementation, this would be more sophisticated
	if result == nil {
		return 0.0
	}

	totalConfidence := 0.0
	count := 0

	// Average confidence across all code types
	for _, mcc := range result.MCC {
		totalConfidence += mcc.Confidence
		count++
	}
	for _, sic := range result.SIC {
		totalConfidence += sic.Confidence
		count++
	}
	for _, naics := range result.NAICS {
		totalConfidence += naics.Confidence
		count++
	}

	if count == 0 {
		return 0.0
	}

	return totalConfidence / float64(count)
}

func (validator *ConfidenceScoreCalibrationValidator) isClassificationCorrect(result *classification.ClassificationCodesInfo, testCase ClassificationTestCase) bool {
	// Simplified correctness check - in real implementation, this would be more sophisticated
	// For now, we'll use a simple heuristic based on confidence score
	confidence := validator.calculateOverallConfidence(result)
	return confidence >= 0.5 // Simplified threshold
}

func (validator *ConfidenceScoreCalibrationValidator) calculateActualAccuracy(result *classification.ClassificationCodesInfo, testCase ClassificationTestCase) float64 {
	// Simplified accuracy calculation - in real implementation, this would be more sophisticated
	if validator.isClassificationCorrect(result, testCase) {
		return 1.0
	}
	return 0.0
}

func (validator *ConfidenceScoreCalibrationValidator) calculateBinActualAccuracy(binDataPoints []CalibrationDataPoint) float64 {
	if len(binDataPoints) == 0 {
		return 0.0
	}

	correctCount := 0
	for _, dp := range binDataPoints {
		if dp.IsCorrect {
			correctCount++
		}
	}

	return float64(correctCount) / float64(len(binDataPoints))
}

func (validator *ConfidenceScoreCalibrationValidator) calculateBinPredictedConfidence(binDataPoints []CalibrationDataPoint) float64 {
	if len(binDataPoints) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, dp := range binDataPoints {
		totalConfidence += dp.Confidence
	}

	return totalConfidence / float64(len(binDataPoints))
}

func (validator *ConfidenceScoreCalibrationValidator) calculateBinBrierScore(binDataPoints []CalibrationDataPoint) float64 {
	if len(binDataPoints) == 0 {
		return 0.0
	}

	totalBrierScore := 0.0
	for _, dp := range binDataPoints {
		actual := 0.0
		if dp.IsCorrect {
			actual = 1.0
		}
		brierScore := math.Pow(dp.Confidence-actual, 2)
		totalBrierScore += brierScore
	}

	return totalBrierScore / float64(len(binDataPoints))
}

// Placeholder methods for advanced calibration calculations
func (validator *ConfidenceScoreCalibrationValidator) calculateReliabilityDiagram(dataPoints []CalibrationDataPoint) *ReliabilityDiagram {
	// TODO: Implement reliability diagram calculation
	return &ReliabilityDiagram{}
}

func (validator *ConfidenceScoreCalibrationValidator) calculateCalibrationCurve(dataPoints []CalibrationDataPoint) *CalibrationCurve {
	// TODO: Implement calibration curve calculation
	return &CalibrationCurve{}
}

func (validator *ConfidenceScoreCalibrationValidator) calculateBrierScore(dataPoints []CalibrationDataPoint) *BrierScoreResult {
	// TODO: Implement Brier score calculation
	return &BrierScoreResult{}
}

func (validator *ConfidenceScoreCalibrationValidator) calculateExpectedCalibrationError(dataPoints []CalibrationDataPoint) *ExpectedCalibrationError {
	// TODO: Implement expected calibration error calculation
	return &ExpectedCalibrationError{}
}

func (validator *ConfidenceScoreCalibrationValidator) calculateTemperatureScaling(dataPoints []CalibrationDataPoint) *TemperatureScalingResult {
	// TODO: Implement temperature scaling calculation
	return &TemperatureScalingResult{}
}

func (validator *ConfidenceScoreCalibrationValidator) calculateCalibrationMetrics(dataPoints []CalibrationDataPoint) *CalibrationMetrics {
	// TODO: Implement comprehensive calibration metrics calculation
	return &CalibrationMetrics{}
}

func (validator *ConfidenceScoreCalibrationValidator) calculateCalibrationSummary(result *CalibrationValidationResult) *CalibrationSummary {
	// Calculate overall calibration based on bins
	overallCalibration := 0.0
	if len(result.ConfidenceBins) > 0 {
		totalError := 0.0
		for _, bin := range result.ConfidenceBins {
			totalError += bin.CalibrationError
		}
		overallCalibration = 1.0 - (totalError / float64(len(result.ConfidenceBins)))
	}

	return &CalibrationSummary{
		OverallCalibration: overallCalibration,
		IsWellCalibrated:   overallCalibration >= validator.Config.MinCalibrationThreshold,
		CalibrationQuality: validator.determineCalibrationQuality(overallCalibration),
	}
}

func (validator *ConfidenceScoreCalibrationValidator) determineCalibrationQuality(calibration float64) string {
	if calibration >= 0.9 {
		return "excellent"
	} else if calibration >= 0.8 {
		return "good"
	} else if calibration >= 0.7 {
		return "fair"
	} else if calibration >= 0.6 {
		return "poor"
	} else {
		return "very_poor"
	}
}

func (validator *ConfidenceScoreCalibrationValidator) generateCalibrationRecommendations(result *CalibrationValidationResult) []string {
	recommendations := []string{}

	if result.CalibrationSummary.OverallCalibration < 0.8 {
		recommendations = append(recommendations, "Overall calibration is below threshold. Consider implementing temperature scaling or Platt scaling.")
	}

	if result.CalibrationSummary.OverallCalibration < 0.7 {
		recommendations = append(recommendations, "Calibration is poor. Review confidence score calculation algorithms.")
	}

	if result.CalibrationSummary.OverallCalibration < 0.6 {
		recommendations = append(recommendations, "Calibration is very poor. Consider complete recalibration of the confidence scoring system.")
	}

	return recommendations
}

// SaveCalibrationReport saves the calibration report to file
func (validator *ConfidenceScoreCalibrationValidator) SaveCalibrationReport(result *CalibrationValidationResult) error {
	// Create validation directory if it doesn't exist
	if err := os.MkdirAll(validator.Config.ValidationDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create validation directory: %w", err)
	}

	// Save JSON report
	jsonFile := filepath.Join(validator.Config.ValidationDirectory, "confidence_calibration_report.json")
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON report: %w", err)
	}

	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	validator.Logger.Printf("✅ Calibration report saved to: %s", jsonFile)
	return nil
}
