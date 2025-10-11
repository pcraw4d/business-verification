package data

import (
	"context"
	"crypto/md5"
	"fmt"
	"math"
	"strings"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// HybridBlender blends synthetic and real time-series data for LSTM input
type HybridBlender struct {
	synthetic *SyntheticDataGenerator
	history   *HistoryCollector
	logger    *zap.Logger
}

// NewHybridBlender creates a new hybrid blender
func NewHybridBlender(synthetic *SyntheticDataGenerator, history *HistoryCollector, logger *zap.Logger) *HybridBlender {
	return &HybridBlender{
		synthetic: synthetic,
		history:   history,
		logger:    logger,
	}
}

// BuildTimeSeries builds a time series for LSTM input by blending real and synthetic data
func (hb *HybridBlender) BuildTimeSeries(ctx context.Context, business *models.RiskAssessmentRequest, sequenceLength int) ([]RiskDataPoint, error) {
	// Generate a business ID from the business name for consistency
	businessID := hb.generateBusinessID(business.BusinessName)

	// Try to get real history
	realHistory, hasRealHistory := hb.history.GetBusinessHistory(ctx, businessID, sequenceLength)

	if hasRealHistory && len(realHistory) >= sequenceLength/2 {
		// Sufficient real data: use mostly real + synthetic padding
		return hb.blendRealWithSynthetic(realHistory, business, sequenceLength, 0.8)
	}

	// Insufficient real data: use mostly synthetic + real enhancement
	syntheticHistory := hb.synthetic.GenerateHistoricalSequence(business, sequenceLength)
	return hb.enhanceWithRealPatterns(syntheticHistory, realHistory, business)
}

// blendRealWithSynthetic blends real data with synthetic data, favoring real data
func (hb *HybridBlender) blendRealWithSynthetic(realHistory []RiskDataPoint, business *models.RiskAssessmentRequest, sequenceLength int, realDataWeight float64) ([]RiskDataPoint, error) {
	// Generate synthetic data for the full sequence
	syntheticHistory := hb.synthetic.GenerateHistoricalSequence(business, sequenceLength)

	// Create blended sequence
	blendedSequence := make([]RiskDataPoint, sequenceLength)
	realDataCount := len(realHistory)

	// Calculate how many real data points to use
	realPointsToUse := int(float64(sequenceLength) * realDataWeight)
	if realPointsToUse > realDataCount {
		realPointsToUse = realDataCount
	}

	// Start with synthetic data
	for i := 0; i < sequenceLength; i++ {
		blendedSequence[i] = syntheticHistory[i]
	}

	// Overlay real data at the end (most recent data)
	startIndex := sequenceLength - realPointsToUse
	for i := 0; i < realPointsToUse; i++ {
		realIndex := realDataCount - realPointsToUse + i
		if realIndex >= 0 && realIndex < realDataCount {
			blendedSequence[startIndex+i] = realHistory[realIndex]
		}
	}

	// Smooth the transition between synthetic and real data
	hb.smoothTransition(blendedSequence, startIndex, realPointsToUse)

	hb.logger.Debug("Blended real with synthetic data",
		zap.String("business_name", business.BusinessName),
		zap.Int("sequence_length", sequenceLength),
		zap.Int("real_data_points", realPointsToUse),
		zap.Float64("real_data_weight", realDataWeight))

	return blendedSequence, nil
}

// enhanceWithRealPatterns enhances synthetic data with patterns from real data
func (hb *HybridBlender) enhanceWithRealPatterns(syntheticHistory []RiskDataPoint, realHistory []RiskDataPoint, business *models.RiskAssessmentRequest) ([]RiskDataPoint, error) {
	if len(realHistory) == 0 {
		// No real data available, return synthetic as-is
		return syntheticHistory, nil
	}

	// Analyze real data patterns
	realPatterns := hb.analyzeRealDataPatterns(realHistory)

	// Apply real patterns to synthetic data
	enhancedSequence := make([]RiskDataPoint, len(syntheticHistory))
	for i, syntheticPoint := range syntheticHistory {
		enhancedPoint := syntheticPoint

		// Adjust risk score based on real patterns
		enhancedPoint.RiskScore = hb.adjustRiskScoreWithPatterns(syntheticPoint.RiskScore, realPatterns)

		// Adjust other features based on real patterns
		enhancedPoint.FinancialHealth = hb.adjustFeatureWithPatterns(syntheticPoint.FinancialHealth, realPatterns.FinancialHealthPattern)
		enhancedPoint.ComplianceScore = hb.adjustFeatureWithPatterns(syntheticPoint.ComplianceScore, realPatterns.CompliancePattern)
		enhancedPoint.MarketConditions = hb.adjustFeatureWithPatterns(syntheticPoint.MarketConditions, realPatterns.MarketPattern)

		enhancedSequence[i] = enhancedPoint
	}

	hb.logger.Debug("Enhanced synthetic data with real patterns",
		zap.String("business_name", business.BusinessName),
		zap.Int("synthetic_points", len(syntheticHistory)),
		zap.Int("real_points", len(realHistory)))

	return enhancedSequence, nil
}

// smoothTransition smooths the transition between synthetic and real data
func (hb *HybridBlender) smoothTransition(sequence []RiskDataPoint, transitionStart, realDataCount int) {
	if transitionStart <= 0 || transitionStart >= len(sequence) {
		return
	}

	// Smooth the transition by interpolating between synthetic and real data
	transitionLength := min(3, realDataCount/2) // Smooth over 3 points or half the real data
	startSmooth := max(0, transitionStart-transitionLength)

	for i := startSmooth; i < transitionStart; i++ {
		// Calculate interpolation factor
		factor := float64(i-startSmooth) / float64(transitionStart-startSmooth)

		// Get the synthetic point at this position
		syntheticPoint := sequence[i]

		// Get the corresponding real point (if available)
		realIndex := i - transitionStart + realDataCount - 1
		if realIndex >= 0 && realIndex < realDataCount {
			realPoint := sequence[transitionStart+realIndex]

			// Interpolate between synthetic and real
			sequence[i].RiskScore = syntheticPoint.RiskScore*(1-factor) + realPoint.RiskScore*factor
			sequence[i].FinancialHealth = syntheticPoint.FinancialHealth*(1-factor) + realPoint.FinancialHealth*factor
			sequence[i].ComplianceScore = syntheticPoint.ComplianceScore*(1-factor) + realPoint.ComplianceScore*factor
			sequence[i].MarketConditions = syntheticPoint.MarketConditions*(1-factor) + realPoint.MarketConditions*factor
		}
	}
}

// analyzeRealDataPatterns analyzes patterns in real data
func (hb *HybridBlender) analyzeRealDataPatterns(realHistory []RiskDataPoint) *RealDataPatterns {
	patterns := &RealDataPatterns{}

	if len(realHistory) == 0 {
		return patterns
	}

	// Calculate averages
	patterns.AverageRiskScore = hb.calculateAverage(realHistory, func(dp RiskDataPoint) float64 { return dp.RiskScore })
	patterns.AverageFinancialHealth = hb.calculateAverage(realHistory, func(dp RiskDataPoint) float64 { return dp.FinancialHealth })
	patterns.AverageComplianceScore = hb.calculateAverage(realHistory, func(dp RiskDataPoint) float64 { return dp.ComplianceScore })
	patterns.AverageMarketConditions = hb.calculateAverage(realHistory, func(dp RiskDataPoint) float64 { return dp.MarketConditions })

	// Calculate trends
	patterns.RiskScoreTrend = hb.calculateTrend(realHistory, func(dp RiskDataPoint) float64 { return dp.RiskScore })
	patterns.FinancialHealthTrend = hb.calculateTrend(realHistory, func(dp RiskDataPoint) float64 { return dp.FinancialHealth })
	patterns.ComplianceScoreTrend = hb.calculateTrend(realHistory, func(dp RiskDataPoint) float64 { return dp.ComplianceScore })

	// Calculate volatility
	patterns.RiskVolatility = hb.calculateVolatility(realHistory, func(dp RiskDataPoint) float64 { return dp.RiskScore })

	// Calculate patterns for feature adjustment
	patterns.FinancialHealthPattern = hb.calculateFeaturePattern(realHistory, func(dp RiskDataPoint) float64 { return dp.FinancialHealth })
	patterns.CompliancePattern = hb.calculateFeaturePattern(realHistory, func(dp RiskDataPoint) float64 { return dp.ComplianceScore })
	patterns.MarketPattern = hb.calculateFeaturePattern(realHistory, func(dp RiskDataPoint) float64 { return dp.MarketConditions })

	return patterns
}

// adjustRiskScoreWithPatterns adjusts risk score based on real data patterns
func (hb *HybridBlender) adjustRiskScoreWithPatterns(syntheticRiskScore float64, patterns *RealDataPatterns) float64 {
	// Adjust based on real data average and trend
	adjustedScore := syntheticRiskScore

	// Apply average adjustment (pull towards real data average)
	averageAdjustment := (patterns.AverageRiskScore - syntheticRiskScore) * 0.3
	adjustedScore += averageAdjustment

	// Apply trend adjustment
	trendAdjustment := patterns.RiskScoreTrend * 0.1
	adjustedScore += trendAdjustment

	// Ensure score is in valid range
	return math.Max(0.0, math.Min(1.0, adjustedScore))
}

// adjustFeatureWithPatterns adjusts a feature based on real data patterns
func (hb *HybridBlender) adjustFeatureWithPatterns(syntheticValue float64, pattern FeaturePattern) float64 {
	// Apply average adjustment
	averageAdjustment := (pattern.Average - syntheticValue) * 0.2
	adjustedValue := syntheticValue + averageAdjustment

	// Apply trend adjustment
	trendAdjustment := pattern.Trend * 0.05
	adjustedValue += trendAdjustment

	// Ensure value is in valid range
	return math.Max(0.0, math.Min(1.0, adjustedValue))
}

// calculateFeaturePattern calculates patterns for a specific feature
func (hb *HybridBlender) calculateFeaturePattern(history []RiskDataPoint, extractor func(RiskDataPoint) float64) FeaturePattern {
	pattern := FeaturePattern{}

	if len(history) == 0 {
		return pattern
	}

	pattern.Average = hb.calculateAverage(history, extractor)
	pattern.Trend = hb.calculateTrend(history, extractor)
	pattern.Volatility = hb.calculateVolatility(history, extractor)

	return pattern
}

// calculateAverage calculates the average of a feature
func (hb *HybridBlender) calculateAverage(history []RiskDataPoint, extractor func(RiskDataPoint) float64) float64 {
	if len(history) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, dp := range history {
		sum += extractor(dp)
	}
	return sum / float64(len(history))
}

// calculateTrend calculates the trend (slope) of a feature over time
func (hb *HybridBlender) calculateTrend(history []RiskDataPoint, extractor func(RiskDataPoint) float64) float64 {
	if len(history) < 2 {
		return 0.0
	}

	// Simple linear regression to calculate trend
	n := len(history)
	sumX, sumY, sumXY, sumXX := 0.0, 0.0, 0.0, 0.0

	for i, dp := range history {
		x := float64(i)
		y := extractor(dp)
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	// Calculate slope
	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumXX - sumX*sumX)
	return slope
}

// calculateVolatility calculates the volatility (standard deviation) of a feature
func (hb *HybridBlender) calculateVolatility(history []RiskDataPoint, extractor func(RiskDataPoint) float64) float64 {
	if len(history) < 2 {
		return 0.0
	}

	// Calculate mean
	sum := 0.0
	for _, dp := range history {
		sum += extractor(dp)
	}
	mean := sum / float64(len(history))

	// Calculate variance
	variance := 0.0
	for _, dp := range history {
		diff := extractor(dp) - mean
		variance += diff * diff
	}
	variance /= float64(len(history) - 1)

	// Return standard deviation
	return math.Sqrt(variance)
}

// GetBlendingStrategy returns the blending strategy used for a business
func (hb *HybridBlender) GetBlendingStrategy(ctx context.Context, businessID string, sequenceLength int) BlendingStrategy {
	realHistory, hasRealHistory := hb.history.GetBusinessHistory(ctx, businessID, sequenceLength)

	if !hasRealHistory || len(realHistory) == 0 {
		return BlendingStrategy{
			Strategy:        "synthetic_only",
			RealDataWeight:  0.0,
			RealDataPoints:  0,
			SyntheticPoints: sequenceLength,
		}
	}

	realDataCount := len(realHistory)
	realDataWeight := float64(realDataCount) / float64(sequenceLength)

	if realDataWeight >= 0.8 {
		return BlendingStrategy{
			Strategy:        "real_dominant",
			RealDataWeight:  realDataWeight,
			RealDataPoints:  realDataCount,
			SyntheticPoints: sequenceLength - realDataCount,
		}
	} else if realDataWeight >= 0.5 {
		return BlendingStrategy{
			Strategy:        "balanced_blend",
			RealDataWeight:  realDataWeight,
			RealDataPoints:  realDataCount,
			SyntheticPoints: sequenceLength - realDataCount,
		}
	} else {
		return BlendingStrategy{
			Strategy:        "synthetic_enhanced",
			RealDataWeight:  realDataWeight,
			RealDataPoints:  realDataCount,
			SyntheticPoints: sequenceLength - realDataCount,
		}
	}
}

// RealDataPatterns contains patterns extracted from real data
type RealDataPatterns struct {
	AverageRiskScore        float64
	AverageFinancialHealth  float64
	AverageComplianceScore  float64
	AverageMarketConditions float64
	RiskScoreTrend          float64
	FinancialHealthTrend    float64
	ComplianceScoreTrend    float64
	RiskVolatility          float64
	FinancialHealthPattern  FeaturePattern
	CompliancePattern       FeaturePattern
	MarketPattern           FeaturePattern
}

// FeaturePattern contains patterns for a specific feature
type FeaturePattern struct {
	Average    float64
	Trend      float64
	Volatility float64
}

// BlendingStrategy describes the blending strategy used
type BlendingStrategy struct {
	Strategy        string  `json:"strategy"`
	RealDataWeight  float64 `json:"real_data_weight"`
	RealDataPoints  int     `json:"real_data_points"`
	SyntheticPoints int     `json:"synthetic_points"`
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// generateBusinessID generates a consistent business ID from business name
func (hb *HybridBlender) generateBusinessID(businessName string) string {
	// Normalize the business name
	normalized := strings.ToLower(strings.TrimSpace(businessName))

	// Create MD5 hash for consistency
	hash := md5.Sum([]byte(normalized))

	// Return first 8 characters of hash as business ID
	return fmt.Sprintf("biz_%x", hash)[:12]
}
