package risk

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
)

// EnhancedRiskCalculator provides advanced risk calculation capabilities
type EnhancedRiskCalculator struct {
	registry             *RiskCategoryRegistry
	logger               *zap.Logger
	config               *EnhancedCalculationConfig
	trendAnalyzer        *TrendAnalyzer
	correlationAnalyzer  *CorrelationAnalyzer
	confidenceCalibrator *ConfidenceCalibrator
}

// EnhancedCalculationConfig contains configuration for enhanced risk calculations
type EnhancedCalculationConfig struct {
	EnableTrendAnalysis         bool    `json:"enable_trend_analysis"`
	EnableCorrelationAnalysis   bool    `json:"enable_correlation_analysis"`
	EnableConfidenceCalibration bool    `json:"enable_confidence_calibration"`
	TrendWindowDays             int     `json:"trend_window_days"`
	CorrelationThreshold        float64 `json:"correlation_threshold"`
	MinDataPointsForTrend       int     `json:"min_data_points_for_trend"`
	OutlierDetectionEnabled     bool    `json:"outlier_detection_enabled"`
	OutlierThreshold            float64 `json:"outlier_threshold"`
	WeightDecayFactor           float64 `json:"weight_decay_factor"`
}

// Note: TrendAnalyzer is already defined in trend_analyzer.go
// type TrendAnalyzer struct {
//	logger *zap.Logger
// }

// Note: CorrelationAnalyzer and ConfidenceCalibrator are defined in their respective files

// EnhancedRiskFactorInput represents enhanced input for risk factor calculation
type EnhancedRiskFactorInput struct {
	FactorID       string                 `json:"factor_id"`
	Data           map[string]interface{} `json:"data"`
	HistoricalData []HistoricalDataPoint  `json:"historical_data,omitempty"`
	RelatedFactors []string               `json:"related_factors,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
	Source         string                 `json:"source"`
	Reliability    float64                `json:"reliability"`
	Context        map[string]interface{} `json:"context,omitempty"`
}

// HistoricalDataPoint represents a historical data point for trend analysis
type HistoricalDataPoint struct {
	Timestamp time.Time              `json:"timestamp"`
	Value     float64                `json:"value"`
	Source    string                 `json:"source"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// EnhancedRiskFactorResult represents enhanced risk factor calculation result
type EnhancedRiskFactorResult struct {
	FactorID     string       `json:"factor_id"`
	FactorName   string       `json:"factor_name"`
	Category     RiskCategory `json:"category"`
	Subcategory  string       `json:"subcategory"`
	Score        float64      `json:"score"`
	Level        RiskLevel    `json:"level"`
	Confidence   float64      `json:"confidence"`
	Explanation  string       `json:"explanation"`
	Evidence     []string     `json:"evidence"`
	CalculatedAt time.Time    `json:"calculated_at"`
	RawValue     interface{}  `json:"raw_value,omitempty"`
	Formula      string       `json:"formula,omitempty"`

	// Enhanced features
	TrendAnalysis         *TrendAnalysis         `json:"trend_analysis,omitempty"`
	CorrelationAnalysis   *CorrelationAnalysis   `json:"correlation_analysis,omitempty"`
	OutlierDetection      *OutlierDetection      `json:"outlier_detection,omitempty"`
	ConfidenceCalibration *ConfidenceCalibration `json:"confidence_calibration,omitempty"`
	WeightedScore         float64                `json:"weighted_score"`
	ImpactScore           float64                `json:"impact_score"`
	VolatilityScore       float64                `json:"volatility_score"`
	PredictiveScore       float64                `json:"predictive_score,omitempty"`
}

// TrendAnalysis contains trend analysis results
type TrendAnalysis struct {
	TrendDirection       string  `json:"trend_direction"` // "improving", "stable", "declining", "volatile"
	TrendStrength        float64 `json:"trend_strength"`  // 0.0 to 1.0
	TrendSlope           float64 `json:"trend_slope"`
	R2Score              float64 `json:"r2_score"`
	DataPoints           int     `json:"data_points"`
	TrendConfidence      float64 `json:"trend_confidence"`
	ProjectedValue       float64 `json:"projected_value,omitempty"`
	ProjectionConfidence float64 `json:"projection_confidence,omitempty"`
}

// CorrelationAnalysis contains correlation analysis results
type CorrelationAnalysis struct {
	CorrelatedFactors   []CorrelatedFactor `json:"correlated_factors"`
	MaxCorrelation      float64            `json:"max_correlation"`
	AvgCorrelation      float64            `json:"avg_correlation"`
	CorrelationStrength string             `json:"correlation_strength"`
}

// CorrelatedFactor represents a correlated risk factor
type CorrelatedFactor struct {
	FactorID     string  `json:"factor_id"`
	FactorName   string  `json:"factor_name"`
	Correlation  float64 `json:"correlation"`
	Significance float64 `json:"significance"`
	Relationship string  `json:"relationship"` // "positive", "negative", "complex"
}

// OutlierDetection contains outlier detection results
type OutlierDetection struct {
	HasOutliers    bool      `json:"has_outliers"`
	OutlierCount   int       `json:"outlier_count"`
	OutlierIndices []int     `json:"outlier_indices,omitempty"`
	OutlierValues  []float64 `json:"outlier_values,omitempty"`
	OutlierScore   float64   `json:"outlier_score"`
	AdjustedScore  float64   `json:"adjusted_score,omitempty"`
}

// ConfidenceCalibration contains confidence calibration results
type ConfidenceCalibration struct {
	CalibratedConfidence float64 `json:"calibrated_confidence"`
	CalibrationFactor    float64 `json:"calibration_factor"`
	HistoricalAccuracy   float64 `json:"historical_accuracy"`
	CalibrationMethod    string  `json:"calibration_method"`
}

// NewEnhancedRiskCalculator creates a new enhanced risk calculator
func NewEnhancedRiskCalculator(registry *RiskCategoryRegistry, logger *zap.Logger, config *EnhancedCalculationConfig) *EnhancedRiskCalculator {
	return &EnhancedRiskCalculator{
		registry:             registry,
		logger:               logger,
		config:               config,
		trendAnalyzer:        &TrendAnalyzer{logger: logger},
		correlationAnalyzer:  &CorrelationAnalyzer{logger: logger},
		confidenceCalibrator: &ConfidenceCalibrator{logger: logger},
	}
}

// CalculateEnhancedFactor calculates enhanced risk score for a specific factor
func (c *EnhancedRiskCalculator) CalculateEnhancedFactor(ctx context.Context, input EnhancedRiskFactorInput) (*EnhancedRiskFactorResult, error) {
	startTime := time.Now()

	c.logger.Info("Starting enhanced risk factor calculation",
		zap.String("factor_id", input.FactorID),
		zap.String("source", input.Source),
		zap.Float64("reliability", input.Reliability))

	// Get the factor definition
	factorDef, exists := c.registry.GetFactor(input.FactorID)
	if !exists {
		return nil, fmt.Errorf("risk factor %s not found", input.FactorID)
	}

	// Validate input data
	if err := c.validateEnhancedInput(input, factorDef); err != nil {
		return nil, fmt.Errorf("invalid input for factor %s: %w", input.FactorID, err)
	}

	// Calculate base score using existing logic
	baseResult, err := c.calculateBaseScore(input, factorDef)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate base score for factor %s: %w", input.FactorID, err)
	}

	// Initialize enhanced result
	result := &EnhancedRiskFactorResult{
		FactorID:     input.FactorID,
		FactorName:   factorDef.Name,
		Category:     factorDef.Category,
		Subcategory:  factorDef.Subcategory,
		Score:        baseResult.Score,
		Level:        baseResult.Level,
		Confidence:   baseResult.Confidence,
		Explanation:  baseResult.Explanation,
		Evidence:     baseResult.Evidence,
		CalculatedAt: time.Now(),
		RawValue:     baseResult.RawValue,
		Formula:      baseResult.Formula,
	}

	// Perform trend analysis if enabled and historical data available
	if c.config.EnableTrendAnalysis && len(input.HistoricalData) >= c.config.MinDataPointsForTrend {
		trendAnalysis, err := c.trendAnalyzer.AnalyzeTrend(input.HistoricalData)
		if err != nil {
			c.logger.Warn("Trend analysis failed", zap.Error(err))
		} else {
			result.TrendAnalysis = trendAnalysis
			// Adjust score based on trend
			result.Score = c.adjustScoreForTrend(result.Score, trendAnalysis)
		}
	}

	// Perform correlation analysis if enabled
	if c.config.EnableCorrelationAnalysis && len(input.RelatedFactors) > 0 {
		correlationAnalysis, err := c.correlationAnalyzer.AnalyzeCorrelations(input.FactorID, input.RelatedFactors, input.Data)
		if err != nil {
			c.logger.Warn("Correlation analysis failed", zap.Error(err))
		} else {
			result.CorrelationAnalysis = correlationAnalysis
			// Adjust confidence based on correlations
			result.Confidence = c.adjustConfidenceForCorrelation(result.Confidence, correlationAnalysis)
		}
	}

	// Perform outlier detection if enabled
	if c.config.OutlierDetectionEnabled && len(input.HistoricalData) > 0 {
		outlierDetection, err := c.detectOutliers(input.HistoricalData, result.Score)
		if err != nil {
			c.logger.Warn("Outlier detection failed", zap.Error(err))
		} else {
			result.OutlierDetection = outlierDetection
			// Adjust score if outliers are detected
			if outlierDetection.HasOutliers && outlierDetection.AdjustedScore > 0 {
				result.Score = outlierDetection.AdjustedScore
			}
		}
	}

	// Perform confidence calibration if enabled
	if c.config.EnableConfidenceCalibration {
		confidenceCalibration, err := c.confidenceCalibrator.CalibrateConfidence(input.FactorID, result.Confidence, input.HistoricalData)
		if err != nil {
			c.logger.Warn("Confidence calibration failed", zap.Error(err))
		} else {
			result.ConfidenceCalibration = confidenceCalibration
			result.Confidence = confidenceCalibration.CalibratedConfidence
		}
	}

	// Calculate additional scores
	result.WeightedScore = c.calculateWeightedScore(result.Score, factorDef.Weight)
	result.ImpactScore = c.calculateImpactScore(result.Score, factorDef.Category)
	result.VolatilityScore = c.calculateVolatilityScore(input.HistoricalData)

	// Calculate predictive score if trend analysis is available
	if result.TrendAnalysis != nil {
		result.PredictiveScore = c.calculatePredictiveScore(result.Score, result.TrendAnalysis)
	}

	duration := time.Since(startTime)
	c.logger.Info("Enhanced risk factor calculation completed",
		zap.String("factor_id", input.FactorID),
		zap.Float64("final_score", result.Score),
		zap.Float64("confidence", result.Confidence),
		zap.Duration("duration", duration))

	return result, nil
}

// calculateBaseScore calculates the base score using existing logic
func (c *EnhancedRiskCalculator) calculateBaseScore(input EnhancedRiskFactorInput, factorDef *RiskFactorDefinition) (*RiskFactorResult, error) {
	// Convert to standard input format
	standardInput := RiskFactorInput{
		FactorID:    input.FactorID,
		Data:        input.Data,
		Timestamp:   input.Timestamp,
		Source:      input.Source,
		Reliability: input.Reliability,
	}

	// Use existing calculator
	calculator := NewRiskFactorCalculator(c.registry)
	return calculator.CalculateFactor(standardInput)
}

// validateEnhancedInput validates enhanced input data
func (c *EnhancedRiskCalculator) validateEnhancedInput(input EnhancedRiskFactorInput, factorDef *RiskFactorDefinition) error {
	// Validate base input
	standardInput := RiskFactorInput{
		FactorID:    input.FactorID,
		Data:        input.Data,
		Timestamp:   input.Timestamp,
		Source:      input.Source,
		Reliability: input.Reliability,
	}

	calculator := NewRiskFactorCalculator(c.registry)
	if err := calculator.validateInput(standardInput, factorDef); err != nil {
		return err
	}

	// Validate historical data
	if len(input.HistoricalData) > 0 {
		for i, point := range input.HistoricalData {
			if point.Timestamp.IsZero() {
				return fmt.Errorf("historical data point %d has zero timestamp", i)
			}
			if math.IsNaN(point.Value) || math.IsInf(point.Value, 0) {
				return fmt.Errorf("historical data point %d has invalid value: %f", i, point.Value)
			}
		}
	}

	// Validate reliability
	if input.Reliability < 0.0 || input.Reliability > 1.0 {
		return fmt.Errorf("reliability must be between 0.0 and 1.0, got %f", input.Reliability)
	}

	return nil
}

// adjustScoreForTrend adjusts the score based on trend analysis
func (c *EnhancedRiskCalculator) adjustScoreForTrend(score float64, trend *TrendAnalysis) float64 {
	if trend == nil {
		return score
	}

	// Adjust score based on trend direction and strength
	adjustment := 0.0

	switch trend.TrendDirection {
	case "improving":
		adjustment = -trend.TrendStrength * 10 // Reduce risk score
	case "declining":
		adjustment = trend.TrendStrength * 10 // Increase risk score
	case "volatile":
		adjustment = trend.TrendStrength * 5 // Moderate increase for volatility
	case "stable":
		// No adjustment for stable trends
	}

	adjustedScore := score + adjustment

	// Ensure score stays within bounds
	return math.Max(0, math.Min(100, adjustedScore))
}

// adjustConfidenceForCorrelation adjusts confidence based on correlation analysis
func (c *EnhancedRiskCalculator) adjustConfidenceForCorrelation(confidence float64, correlation *CorrelationAnalysis) float64 {
	if correlation == nil {
		return confidence
	}

	// Higher correlation with other factors increases confidence
	correlationBoost := correlation.AvgCorrelation * 0.1

	adjustedConfidence := confidence + correlationBoost

	// Ensure confidence stays within bounds
	return math.Max(0, math.Min(1, adjustedConfidence))
}

// detectOutliers detects outliers in historical data
func (c *EnhancedRiskCalculator) detectOutliers(historicalData []HistoricalDataPoint, currentScore float64) (*OutlierDetection, error) {
	if len(historicalData) < 3 {
		return &OutlierDetection{HasOutliers: false}, nil
	}

	// Extract values
	values := make([]float64, len(historicalData))
	for i, point := range historicalData {
		values[i] = point.Value
	}

	// Calculate statistics
	mean := c.calculateMean(values)
	stdDev := c.calculateStdDev(values, mean)

	// Detect outliers using z-score method
	var outliers []int
	var outlierValues []float64

	for i, value := range values {
		zScore := math.Abs((value - mean) / stdDev)
		if zScore > c.config.OutlierThreshold {
			outliers = append(outliers, i)
			outlierValues = append(outlierValues, value)
		}
	}

	// Calculate outlier score
	outlierScore := float64(len(outliers)) / float64(len(values))

	// Calculate adjusted score (remove outliers and recalculate)
	var adjustedScore float64
	if len(outliers) > 0 && len(outliers) < len(values)/2 {
		// Remove outliers and recalculate mean
		cleanValues := make([]float64, 0, len(values)-len(outliers))
		for i, value := range values {
			isOutlier := false
			for _, outlierIndex := range outliers {
				if i == outlierIndex {
					isOutlier = true
					break
				}
			}
			if !isOutlier {
				cleanValues = append(cleanValues, value)
			}
		}
		adjustedScore = c.calculateMean(cleanValues)
	} else {
		adjustedScore = currentScore
	}

	return &OutlierDetection{
		HasOutliers:    len(outliers) > 0,
		OutlierCount:   len(outliers),
		OutlierIndices: outliers,
		OutlierValues:  outlierValues,
		OutlierScore:   outlierScore,
		AdjustedScore:  adjustedScore,
	}, nil
}

// calculateWeightedScore calculates weighted score based on factor weight
func (c *EnhancedRiskCalculator) calculateWeightedScore(score float64, weight float64) float64 {
	return score * weight
}

// calculateImpactScore calculates impact score based on category
func (c *EnhancedRiskCalculator) calculateImpactScore(score float64, category RiskCategory) float64 {
	// Different categories have different impact multipliers
	impactMultipliers := map[RiskCategory]float64{
		RiskCategoryFinancial:     1.2,
		RiskCategoryOperational:   1.0,
		RiskCategoryRegulatory:    1.3,
		RiskCategoryReputational:  1.1,
		RiskCategoryCybersecurity: 1.4,
	}

	multiplier := impactMultipliers[category]
	if multiplier == 0 {
		multiplier = 1.0
	}

	return score * multiplier
}

// calculateVolatilityScore calculates volatility score from historical data
func (c *EnhancedRiskCalculator) calculateVolatilityScore(historicalData []HistoricalDataPoint) float64 {
	if len(historicalData) < 2 {
		return 0.0
	}

	values := make([]float64, len(historicalData))
	for i, point := range historicalData {
		values[i] = point.Value
	}

	mean := c.calculateMean(values)
	stdDev := c.calculateStdDev(values, mean)

	// Normalize volatility to 0-100 scale
	// Higher standard deviation = higher volatility
	volatilityScore := (stdDev / mean) * 100

	return math.Min(100, volatilityScore)
}

// calculatePredictiveScore calculates predictive score based on trend
func (c *EnhancedRiskCalculator) calculatePredictiveScore(currentScore float64, trend *TrendAnalysis) float64 {
	if trend == nil || trend.ProjectedValue == 0 {
		return currentScore
	}

	// Weight current score and projected value
	weight := 0.7 // 70% current, 30% projected
	predictiveScore := (currentScore * weight) + (trend.ProjectedValue * (1 - weight))

	return math.Max(0, math.Min(100, predictiveScore))
}

// Helper functions for statistical calculations
func (c *EnhancedRiskCalculator) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (c *EnhancedRiskCalculator) calculateStdDev(values []float64, mean float64) float64 {
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
