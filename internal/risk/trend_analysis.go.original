package risk

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// TrendAnalysisService provides comprehensive risk trend analysis functionality
type TrendAnalysisService struct {
	logger *observability.Logger
}

// NewTrendAnalysisService creates a new trend analysis service
func NewTrendAnalysisService(logger *observability.Logger) *TrendAnalysisService {
	return &TrendAnalysisService{
		logger: logger,
	}
}

// TrendAnalysisResult represents the result of trend analysis
type TrendAnalysisResult struct {
	BusinessID           string                `json:"business_id"`
	AnalysisPeriod       string                `json:"analysis_period"`
	OverallTrend         TrendDirection        `json:"overall_trend"`
	OverallTrendStrength float64               `json:"overall_trend_strength"`
	CategoryTrends       map[string]TrendData  `json:"category_trends"`
	FactorTrends         map[string]TrendData  `json:"factor_trends"`
	Seasonality          SeasonalityAnalysis   `json:"seasonality"`
	Volatility           VolatilityAnalysis    `json:"volatility"`
	Predictions          []TrendPrediction     `json:"predictions"`
	Anomalies            []TrendAnomaly        `json:"anomalies"`
	Recommendations      []TrendRecommendation `json:"recommendations"`
	GeneratedAt          time.Time             `json:"generated_at"`
}

// TrendDirection represents the direction of a trend
type TrendDirection string

const (
	TrendDirectionIncreasing TrendDirection = "increasing"
	TrendDirectionDecreasing TrendDirection = "decreasing"
	TrendDirectionStable     TrendDirection = "stable"
	TrendDirectionVolatile   TrendDirection = "volatile"
)

// TrendData represents trend data for a specific category or factor
type TrendData struct {
	Direction     TrendDirection `json:"direction"`
	Strength      float64        `json:"strength"`   // 0.0 to 1.0
	Slope         float64        `json:"slope"`      // Rate of change
	R2            float64        `json:"r_squared"`  // Goodness of fit
	Confidence    float64        `json:"confidence"` // 0.0 to 1.0
	DataPoints    int            `json:"data_points"`
	StartValue    float64        `json:"start_value"`
	EndValue      float64        `json:"end_value"`
	ChangePercent float64        `json:"change_percent"`
	Period        string         `json:"period"`
}

// SeasonalityAnalysis represents seasonality patterns
type SeasonalityAnalysis struct {
	HasSeasonality bool               `json:"has_seasonality"`
	Pattern        string             `json:"pattern"`  // "monthly", "quarterly", "yearly"
	Strength       float64            `json:"strength"` // 0.0 to 1.0
	PeakPeriods    []time.Time        `json:"peak_periods"`
	TroughPeriods  []time.Time        `json:"trough_periods"`
	SeasonalData   map[string]float64 `json:"seasonal_data"`
}

// VolatilityAnalysis represents volatility patterns
type VolatilityAnalysis struct {
	OverallVolatility     float64            `json:"overall_volatility"`
	VolatilityTrend       TrendDirection     `json:"volatility_trend"`
	HighVolatilityPeriods []time.Time        `json:"high_volatility_periods"`
	LowVolatilityPeriods  []time.Time        `json:"low_volatility_periods"`
	VolatilityByCategory  map[string]float64 `json:"volatility_by_category"`
}

// TrendPrediction represents a trend prediction
type TrendPrediction struct {
	Horizon        time.Duration  `json:"horizon"`
	PredictedValue float64        `json:"predicted_value"`
	Confidence     float64        `json:"confidence"`
	LowerBound     float64        `json:"lower_bound"`
	UpperBound     float64        `json:"upper_bound"`
	TrendDirection TrendDirection `json:"trend_direction"`
	KeyFactors     []string       `json:"key_factors"`
}

// TrendAnomaly represents a detected anomaly
type TrendAnomaly struct {
	Timestamp     time.Time `json:"timestamp"`
	Value         float64   `json:"value"`
	ExpectedValue float64   `json:"expected_value"`
	Deviation     float64   `json:"deviation"`
	Severity      string    `json:"severity"` // "low", "medium", "high", "critical"
	Description   string    `json:"description"`
	Category      string    `json:"category"`
}

// TrendRecommendation represents a recommendation based on trend analysis
type TrendRecommendation struct {
	Type        string  `json:"type"`     // "mitigation", "monitoring", "investigation"
	Priority    string  `json:"priority"` // "low", "medium", "high", "critical"
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Action      string  `json:"action"`
	Timeline    string  `json:"timeline"`
	Impact      string  `json:"impact"`
	Confidence  float64 `json:"confidence"`
}

// AnalyzeTrends performs comprehensive trend analysis on risk data
func (s *TrendAnalysisService) AnalyzeTrends(ctx context.Context, businessID string, assessments []*RiskAssessment, period string) (*TrendAnalysisResult, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Starting trend analysis",
		"request_id", requestID,
		"business_id", businessID,
		"assessment_count", len(assessments),
		"period", period,
	)

	if len(assessments) < 2 {
		return s.createInsufficientDataResult(businessID, period), nil
	}

	// Sort assessments by time
	sort.Slice(assessments, func(i, j int) bool {
		return assessments[i].AssessedAt.Before(assessments[j].AssessedAt)
	})

	result := &TrendAnalysisResult{
		BusinessID:     businessID,
		AnalysisPeriod: period,
		GeneratedAt:    time.Now(),
	}

	// Analyze overall trends
	overallTrend := s.analyzeOverallTrend(assessments)
	result.OverallTrend = overallTrend.Direction
	result.OverallTrendStrength = overallTrend.Strength

	// Analyze category trends
	result.CategoryTrends = s.analyzeCategoryTrends(assessments)

	// Analyze factor trends
	result.FactorTrends = s.analyzeFactorTrends(assessments)

	// Analyze seasonality
	result.Seasonality = s.analyzeSeasonality(assessments)

	// Analyze volatility
	result.Volatility = s.analyzeVolatility(assessments)

	// Generate predictions
	result.Predictions = s.generatePredictions(assessments)

	// Detect anomalies
	result.Anomalies = s.detectAnomalies(assessments)

	// Generate recommendations
	result.Recommendations = s.generateRecommendations(result)

	s.logger.Info("Trend analysis completed",
		"request_id", requestID,
		"business_id", businessID,
		"overall_trend", result.OverallTrend,
		"trend_strength", result.OverallTrendStrength,
		"anomaly_count", len(result.Anomalies),
		"recommendation_count", len(result.Recommendations),
	)

	return result, nil
}

// analyzeOverallTrend analyzes the overall risk trend
func (s *TrendAnalysisService) analyzeOverallTrend(assessments []*RiskAssessment) TrendData {
	if len(assessments) < 2 {
		return TrendData{
			Direction:  TrendDirectionStable,
			Strength:   0.0,
			Confidence: 0.0,
		}
	}

	// Extract overall scores over time
	var scores []float64
	var times []time.Time
	for _, assessment := range assessments {
		scores = append(scores, assessment.OverallScore)
		times = append(times, assessment.AssessedAt)
	}

	// Calculate linear regression
	slope, r2, confidence := s.calculateLinearRegression(times, scores)

	// Determine trend direction
	direction := s.determineTrendDirection(slope, r2)

	// Calculate trend strength
	strength := math.Abs(slope) * r2

	// Calculate change percentage
	startValue := scores[0]
	endValue := scores[len(scores)-1]
	changePercent := ((endValue - startValue) / startValue) * 100

	return TrendData{
		Direction:     direction,
		Strength:      strength,
		Slope:         slope,
		R2:            r2,
		Confidence:    confidence,
		DataPoints:    len(scores),
		StartValue:    startValue,
		EndValue:      endValue,
		ChangePercent: changePercent,
		Period:        s.calculatePeriod(times[0], times[len(times)-1]),
	}
}

// analyzeCategoryTrends analyzes trends for each risk category
func (s *TrendAnalysisService) analyzeCategoryTrends(assessments []*RiskAssessment) map[string]TrendData {
	trends := make(map[string]TrendData)

	if len(assessments) == 0 {
		return trends
	}

	// Get all categories from the first assessment
	categories := make(map[RiskCategory]bool)
	for category := range assessments[0].CategoryScores {
		categories[category] = true
	}

	// Analyze trends for each category
	for category := range categories {
		var scores []float64
		var times []time.Time

		for _, assessment := range assessments {
			if score, exists := assessment.CategoryScores[category]; exists {
				scores = append(scores, score.Score)
				times = append(times, assessment.AssessedAt)
			}
		}

		if len(scores) >= 2 {
			slope, r2, confidence := s.calculateLinearRegression(times, scores)
			direction := s.determineTrendDirection(slope, r2)
			strength := math.Abs(slope) * r2

			startValue := scores[0]
			endValue := scores[len(scores)-1]
			changePercent := ((endValue - startValue) / startValue) * 100

			trends[string(category)] = TrendData{
				Direction:     direction,
				Strength:      strength,
				Slope:         slope,
				R2:            r2,
				Confidence:    confidence,
				DataPoints:    len(scores),
				StartValue:    startValue,
				EndValue:      endValue,
				ChangePercent: changePercent,
				Period:        s.calculatePeriod(times[0], times[len(times)-1]),
			}
		}
	}

	return trends
}

// analyzeFactorTrends analyzes trends for each risk factor
func (s *TrendAnalysisService) analyzeFactorTrends(assessments []*RiskAssessment) map[string]TrendData {
	trends := make(map[string]TrendData)

	if len(assessments) == 0 {
		return trends
	}

	// Get all factors from all assessments
	factors := make(map[string]bool)
	for _, assessment := range assessments {
		for _, factor := range assessment.FactorScores {
			factors[factor.FactorID] = true
		}
	}

	// Analyze trends for each factor
	for factorID := range factors {
		var scores []float64
		var times []time.Time

		for _, assessment := range assessments {
			for _, factor := range assessment.FactorScores {
				if factor.FactorID == factorID {
					scores = append(scores, factor.Score)
					times = append(times, assessment.AssessedAt)
					break
				}
			}
		}

		if len(scores) >= 2 {
			slope, r2, confidence := s.calculateLinearRegression(times, scores)
			direction := s.determineTrendDirection(slope, r2)
			strength := math.Abs(slope) * r2

			startValue := scores[0]
			endValue := scores[len(scores)-1]
			changePercent := ((endValue - startValue) / startValue) * 100

			trends[factorID] = TrendData{
				Direction:     direction,
				Strength:      strength,
				Slope:         slope,
				R2:            r2,
				Confidence:    confidence,
				DataPoints:    len(scores),
				StartValue:    startValue,
				EndValue:      endValue,
				ChangePercent: changePercent,
				Period:        s.calculatePeriod(times[0], times[len(times)-1]),
			}
		}
	}

	return trends
}

// analyzeSeasonality analyzes seasonality patterns
func (s *TrendAnalysisService) analyzeSeasonality(assessments []*RiskAssessment) SeasonalityAnalysis {
	if len(assessments) < 12 {
		return SeasonalityAnalysis{
			HasSeasonality: false,
			Pattern:        "insufficient_data",
			Strength:       0.0,
		}
	}

	// Group scores by month
	monthlyScores := make(map[int][]float64)
	for _, assessment := range assessments {
		month := int(assessment.AssessedAt.Month())
		monthlyScores[month] = append(monthlyScores[month], assessment.OverallScore)
	}

	// Calculate average scores by month
	monthlyAverages := make(map[int]float64)
	for month, scores := range monthlyScores {
		if len(scores) > 0 {
			sum := 0.0
			for _, score := range scores {
				sum += score
			}
			monthlyAverages[month] = sum / float64(len(scores))
		}
	}

	// Calculate seasonality strength
	overallAverage := 0.0
	count := 0
	for _, avg := range monthlyAverages {
		overallAverage += avg
		count++
	}
	if count > 0 {
		overallAverage /= float64(count)
	}

	variance := 0.0
	for _, avg := range monthlyAverages {
		variance += math.Pow(avg-overallAverage, 2)
	}
	if count > 0 {
		variance /= float64(count)
	}

	seasonalityStrength := math.Sqrt(variance) / overallAverage

	// Determine if seasonality exists
	hasSeasonality := seasonalityStrength > 0.1

	// Find peak and trough periods
	var peakValue, troughValue float64

	for _, avg := range monthlyAverages {
		if avg > peakValue {
			peakValue = avg
		}
		if avg < troughValue || troughValue == 0 {
			troughValue = avg
		}
	}

	// Create seasonal data
	seasonalData := make(map[string]float64)
	for month, avg := range monthlyAverages {
		seasonalData[fmt.Sprintf("%d", month)] = avg
	}

	return SeasonalityAnalysis{
		HasSeasonality: hasSeasonality,
		Pattern:        "monthly",
		Strength:       seasonalityStrength,
		SeasonalData:   seasonalData,
	}
}

// analyzeVolatility analyzes volatility patterns
func (s *TrendAnalysisService) analyzeVolatility(assessments []*RiskAssessment) VolatilityAnalysis {
	if len(assessments) < 2 {
		return VolatilityAnalysis{
			OverallVolatility: 0.0,
			VolatilityTrend:   TrendDirectionStable,
		}
	}

	// Calculate overall volatility
	var scores []float64
	for _, assessment := range assessments {
		scores = append(scores, assessment.OverallScore)
	}

	mean := 0.0
	for _, score := range scores {
		mean += score
	}
	mean /= float64(len(scores))

	variance := 0.0
	for _, score := range scores {
		variance += math.Pow(score-mean, 2)
	}
	variance /= float64(len(scores))

	overallVolatility := math.Sqrt(variance)

	// Calculate volatility trend
	var volatilities []float64
	for i := 1; i < len(scores); i++ {
		volatility := math.Abs(scores[i] - scores[i-1])
		volatilities = append(volatilities, volatility)
	}

	volatilitySlope, _, _ := s.calculateLinearRegression(
		make([]time.Time, len(volatilities)),
		volatilities,
	)

	volatilityTrend := s.determineTrendDirection(volatilitySlope, 0.5)

	// Calculate volatility by category
	volatilityByCategory := make(map[string]float64)
	if len(assessments) > 0 {
		for category := range assessments[0].CategoryScores {
			var categoryScores []float64
			for _, assessment := range assessments {
				if score, exists := assessment.CategoryScores[category]; exists {
					categoryScores = append(categoryScores, score.Score)
				}
			}

			if len(categoryScores) >= 2 {
				categoryMean := 0.0
				for _, score := range categoryScores {
					categoryMean += score
				}
				categoryMean /= float64(len(categoryScores))

				categoryVariance := 0.0
				for _, score := range categoryScores {
					categoryVariance += math.Pow(score-categoryMean, 2)
				}
				categoryVariance /= float64(len(categoryScores))

				volatilityByCategory[string(category)] = math.Sqrt(categoryVariance)
			}
		}
	}

	return VolatilityAnalysis{
		OverallVolatility:    overallVolatility,
		VolatilityTrend:      volatilityTrend,
		VolatilityByCategory: volatilityByCategory,
	}
}

// generatePredictions generates trend predictions
func (s *TrendAnalysisService) generatePredictions(assessments []*RiskAssessment) []TrendPrediction {
	var predictions []TrendPrediction

	if len(assessments) < 3 {
		return predictions
	}

	// Extract recent scores for prediction
	var scores []float64
	var times []time.Time
	for _, assessment := range assessments {
		scores = append(scores, assessment.OverallScore)
		times = append(times, assessment.AssessedAt)
	}

	// Generate predictions for different horizons
	horizons := []time.Duration{
		30 * 24 * time.Hour,  // 1 month
		90 * 24 * time.Hour,  // 3 months
		180 * 24 * time.Hour, // 6 months
	}

	for _, horizon := range horizons {
		prediction := s.predictTrend(scores, times, horizon)
		if prediction != nil {
			predictions = append(predictions, *prediction)
		}
	}

	return predictions
}

// detectAnomalies detects anomalies in the trend data
func (s *TrendAnalysisService) detectAnomalies(assessments []*RiskAssessment) []TrendAnomaly {
	var anomalies []TrendAnomaly

	if len(assessments) < 3 {
		return anomalies
	}

	// Calculate moving average and standard deviation
	var scores []float64
	var times []time.Time
	for _, assessment := range assessments {
		scores = append(scores, assessment.OverallScore)
		times = append(times, assessment.AssessedAt)
	}

	// Use simple moving average for anomaly detection
	windowSize := 3
	if len(scores) < windowSize {
		return anomalies
	}

	for i := windowSize; i < len(scores); i++ {
		// Calculate moving average
		sum := 0.0
		for j := i - windowSize; j < i; j++ {
			sum += scores[j]
		}
		movingAverage := sum / float64(windowSize)

		// Calculate standard deviation
		variance := 0.0
		for j := i - windowSize; j < i; j++ {
			variance += math.Pow(scores[j]-movingAverage, 2)
		}
		variance /= float64(windowSize)
		stdDev := math.Sqrt(variance)

		// Check for anomaly (2 standard deviations from mean)
		currentScore := scores[i]
		deviation := math.Abs(currentScore - movingAverage)

		if deviation > 2*stdDev {
			severity := "low"
			if deviation > 3*stdDev {
				severity = "high"
			} else if deviation > 2.5*stdDev {
				severity = "medium"
			}

			anomaly := TrendAnomaly{
				Timestamp:     times[i],
				Value:         currentScore,
				ExpectedValue: movingAverage,
				Deviation:     deviation,
				Severity:      severity,
				Description:   fmt.Sprintf("Score %.1f deviates %.1f from expected %.1f", currentScore, deviation, movingAverage),
				Category:      "overall_risk",
			}

			anomalies = append(anomalies, anomaly)
		}
	}

	return anomalies
}

// generateRecommendations generates recommendations based on trend analysis
func (s *TrendAnalysisService) generateRecommendations(result *TrendAnalysisResult) []TrendRecommendation {
	var recommendations []TrendRecommendation

	// Overall trend recommendations
	if result.OverallTrend == TrendDirectionIncreasing {
		recommendations = append(recommendations, TrendRecommendation{
			Type:        "mitigation",
			Priority:    "high",
			Title:       "Address Increasing Risk Trend",
			Description: "Overall risk is trending upward. Immediate action required.",
			Action:      "Conduct detailed risk assessment and implement mitigation strategies",
			Timeline:    "1-2 weeks",
			Impact:      "High - Prevents further risk escalation",
			Confidence:  result.OverallTrendStrength,
		})
	}

	// Volatility recommendations
	if result.Volatility.OverallVolatility > 15.0 {
		recommendations = append(recommendations, TrendRecommendation{
			Type:        "monitoring",
			Priority:    "medium",
			Title:       "High Risk Volatility Detected",
			Description: "Risk scores show high volatility, indicating unstable risk profile.",
			Action:      "Increase monitoring frequency and investigate volatility drivers",
			Timeline:    "1 month",
			Impact:      "Medium - Improves risk predictability",
			Confidence:  0.8,
		})
	}

	// Anomaly recommendations
	if len(result.Anomalies) > 0 {
		criticalAnomalies := 0
		for _, anomaly := range result.Anomalies {
			if anomaly.Severity == "high" || anomaly.Severity == "critical" {
				criticalAnomalies++
			}
		}

		if criticalAnomalies > 0 {
			recommendations = append(recommendations, TrendRecommendation{
				Type:        "investigation",
				Priority:    "critical",
				Title:       "Critical Anomalies Detected",
				Description: fmt.Sprintf("%d critical anomalies detected in risk data", criticalAnomalies),
				Action:      "Immediately investigate anomaly causes and implement corrective actions",
				Timeline:    "1 week",
				Impact:      "Critical - Addresses data quality and risk accuracy",
				Confidence:  0.9,
			})
		}
	}

	// Category-specific recommendations
	for category, trend := range result.CategoryTrends {
		if trend.Direction == TrendDirectionIncreasing && trend.Strength > 0.5 {
			recommendations = append(recommendations, TrendRecommendation{
				Type:        "mitigation",
				Priority:    "medium",
				Title:       fmt.Sprintf("Address %s Risk Trend", category),
				Description: fmt.Sprintf("%s risk is trending upward with %.1f%% strength", category, trend.Strength*100),
				Action:      fmt.Sprintf("Focus mitigation efforts on %s risk factors", category),
				Timeline:    "2-4 weeks",
				Impact:      "Medium - Reduces category-specific risk",
				Confidence:  trend.Confidence,
			})
		}
	}

	return recommendations
}

// Helper methods
func (s *TrendAnalysisService) calculateLinearRegression(times []time.Time, values []float64) (slope, r2, confidence float64) {
	if len(values) < 2 {
		return 0.0, 0.0, 0.0
	}

	// Convert times to numeric values for regression
	var xValues []float64
	for i := range times {
		xValues = append(xValues, float64(i))
	}

	// Calculate means
	xMean := 0.0
	yMean := 0.0
	for i := range xValues {
		xMean += xValues[i]
		yMean += values[i]
	}
	xMean /= float64(len(xValues))
	yMean /= float64(len(values))

	// Calculate slope and intercept
	numerator := 0.0
	denominator := 0.0
	for i := range xValues {
		numerator += (xValues[i] - xMean) * (values[i] - yMean)
		denominator += (xValues[i] - xMean) * (xValues[i] - xMean)
	}

	if denominator == 0 {
		return 0.0, 0.0, 0.0
	}

	slope = numerator / denominator
	intercept := yMean - slope*xMean

	// Calculate R-squared
	ssRes := 0.0
	ssTot := 0.0
	for i := range xValues {
		predicted := slope*xValues[i] + intercept
		ssRes += math.Pow(values[i]-predicted, 2)
		ssTot += math.Pow(values[i]-yMean, 2)
	}

	if ssTot == 0 {
		r2 = 0.0
	} else {
		r2 = 1 - (ssRes / ssTot)
	}

	// Calculate confidence based on R-squared and data points
	confidence = r2 * math.Min(1.0, float64(len(values))/10.0)

	return slope, r2, confidence
}

func (s *TrendAnalysisService) determineTrendDirection(slope, r2 float64) TrendDirection {
	if r2 < 0.3 {
		return TrendDirectionVolatile
	}

	if slope > 0.5 {
		return TrendDirectionIncreasing
	} else if slope < -0.5 {
		return TrendDirectionDecreasing
	}

	return TrendDirectionStable
}

func (s *TrendAnalysisService) calculatePeriod(start, end time.Time) string {
	duration := end.Sub(start)
	days := int(duration.Hours() / 24)

	switch {
	case days <= 30:
		return "1month"
	case days <= 90:
		return "3months"
	case days <= 180:
		return "6months"
	case days <= 365:
		return "1year"
	default:
		return fmt.Sprintf("%dmonths", days/30)
	}
}

func (s *TrendAnalysisService) predictTrend(scores []float64, times []time.Time, horizon time.Duration) *TrendPrediction {
	if len(scores) < 3 {
		return nil
	}

	// Simple linear prediction
	slope, r2, _ := s.calculateLinearRegression(times, scores)

	lastScore := scores[len(scores)-1]

	// Predict future value
	timeSteps := horizon.Hours() / 24.0 // Convert to days
	predictedValue := lastScore + (slope * timeSteps)

	// Cap predicted value
	if predictedValue < 0 {
		predictedValue = 0
	} else if predictedValue > 100 {
		predictedValue = 100
	}

	// Calculate confidence bounds
	confidence := r2 * 0.8                  // Reduce confidence for predictions
	boundRange := 10.0 * (1.0 - confidence) // Wider bounds for lower confidence

	lowerBound := predictedValue - boundRange
	upperBound := predictedValue + boundRange

	if lowerBound < 0 {
		lowerBound = 0
	}
	if upperBound > 100 {
		upperBound = 100
	}

	direction := s.determineTrendDirection(slope, r2)

	return &TrendPrediction{
		Horizon:        horizon,
		PredictedValue: predictedValue,
		Confidence:     confidence,
		LowerBound:     lowerBound,
		UpperBound:     upperBound,
		TrendDirection: direction,
		KeyFactors:     []string{"historical_trend", "linear_extrapolation"},
	}
}

func (s *TrendAnalysisService) createInsufficientDataResult(businessID, period string) *TrendAnalysisResult {
	return &TrendAnalysisResult{
		BusinessID:           businessID,
		AnalysisPeriod:       period,
		OverallTrend:         TrendDirectionStable,
		OverallTrendStrength: 0.0,
		CategoryTrends:       make(map[string]TrendData),
		FactorTrends:         make(map[string]TrendData),
		Seasonality: SeasonalityAnalysis{
			HasSeasonality: false,
			Pattern:        "insufficient_data",
			Strength:       0.0,
		},
		Volatility: VolatilityAnalysis{
			OverallVolatility: 0.0,
			VolatilityTrend:   TrendDirectionStable,
		},
		Predictions:     []TrendPrediction{},
		Anomalies:       []TrendAnomaly{},
		Recommendations: []TrendRecommendation{},
		GeneratedAt:     time.Now(),
	}
}
