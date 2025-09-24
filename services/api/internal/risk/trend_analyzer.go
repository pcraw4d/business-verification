package risk

import (
	"fmt"
	"math"
	"sort"
	"time"

	"go.uber.org/zap"
)

// TrendAnalyzer analyzes trends in risk data
type TrendAnalyzer struct {
	logger *zap.Logger
}

// AnalyzeTrend analyzes trends in historical data points
func (ta *TrendAnalyzer) AnalyzeTrend(historicalData []HistoricalDataPoint) (*TrendAnalysis, error) {
	if len(historicalData) < 2 {
		return nil, fmt.Errorf("insufficient data points for trend analysis: %d", len(historicalData))
	}

	// Sort data by timestamp
	sortedData := make([]HistoricalDataPoint, len(historicalData))
	copy(sortedData, historicalData)
	sort.Slice(sortedData, func(i, j int) bool {
		return sortedData[i].Timestamp.Before(sortedData[j].Timestamp)
	})

	// Extract values and timestamps
	values := make([]float64, len(sortedData))
	timestamps := make([]time.Time, len(sortedData))
	for i, point := range sortedData {
		values[i] = point.Value
		timestamps[i] = point.Timestamp
	}

	// Calculate trend slope using linear regression
	slope, r2Score, err := ta.calculateLinearRegression(values, timestamps)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate linear regression: %w", err)
	}

	// Determine trend direction and strength
	direction, strength := ta.determineTrendDirection(slope, r2Score, values)

	// Calculate trend confidence
	confidence := ta.calculateTrendConfidence(r2Score, len(values))

	// Project future value if trend is significant
	var projectedValue, projectionConfidence float64
	if confidence > 0.7 && math.Abs(slope) > 0.1 {
		projectedValue, projectionConfidence = ta.projectFutureValue(values, timestamps, slope)
	}

	return &TrendAnalysis{
		TrendDirection:       direction,
		TrendStrength:        strength,
		TrendSlope:           slope,
		R2Score:              r2Score,
		DataPoints:           len(values),
		TrendConfidence:      confidence,
		ProjectedValue:       projectedValue,
		ProjectionConfidence: projectionConfidence,
	}, nil
}

// calculateLinearRegression calculates linear regression for trend analysis
func (ta *TrendAnalyzer) calculateLinearRegression(values []float64, timestamps []time.Time) (slope, r2Score float64, err error) {
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
	xMean := ta.calculateMean(xValues)
	yMean := ta.calculateMean(values)

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
	r2Score = ta.calculateRSquared(values, xValues, slope, intercept)

	return slope, r2Score, nil
}

// determineTrendDirection determines trend direction and strength
func (ta *TrendAnalyzer) determineTrendDirection(slope, r2Score float64, values []float64) (direction string, strength float64) {
	// Calculate volatility to determine if trend is stable or volatile
	volatility := ta.calculateVolatility(values)

	// Determine direction based on slope
	if math.Abs(slope) < 0.01 {
		if volatility > 0.1 {
			direction = "volatile"
		} else {
			direction = "stable"
		}
	} else if slope > 0 {
		direction = "declining" // For risk scores, increasing values mean declining risk
	} else {
		direction = "improving" // For risk scores, decreasing values mean improving risk
	}

	// Calculate strength based on R² score and slope magnitude
	strength = math.Abs(slope) * r2Score

	// Adjust for volatility
	if volatility > 0.2 {
		strength *= 0.5 // Reduce strength for highly volatile data
	}

	// Normalize strength to 0-1 range
	strength = math.Min(1.0, strength)

	return direction, strength
}

// calculateTrendConfidence calculates confidence in the trend analysis
func (ta *TrendAnalyzer) calculateTrendConfidence(r2Score float64, dataPoints int) float64 {
	// Base confidence on R² score
	confidence := r2Score

	// Adjust for number of data points
	if dataPoints >= 10 {
		confidence *= 1.0
	} else if dataPoints >= 5 {
		confidence *= 0.9
	} else if dataPoints >= 3 {
		confidence *= 0.8
	} else {
		confidence *= 0.6
	}

	return math.Max(0, math.Min(1, confidence))
}

// projectFutureValue projects future value based on trend
func (ta *TrendAnalyzer) projectFutureValue(values []float64, timestamps []time.Time, slope float64) (projectedValue, confidence float64) {
	if len(values) == 0 {
		return 0, 0
	}

	// Project 30 days into the future
	lastTimestamp := timestamps[len(timestamps)-1]
	futureTime := lastTimestamp.Add(30 * 24 * time.Hour)
	startTime := timestamps[0]

	daysFromStart := futureTime.Sub(startTime).Hours() / 24.0
	projectedValue = values[0] + slope*daysFromStart

	// Calculate projection confidence based on trend stability
	volatility := ta.calculateVolatility(values)
	confidence = math.Max(0, 1.0-volatility)

	// Reduce confidence for longer projections
	confidence *= 0.8

	return projectedValue, confidence
}

// calculateRSquared calculates R² score for regression
func (ta *TrendAnalyzer) calculateRSquared(yValues, xValues []float64, slope, intercept float64) float64 {
	if len(yValues) != len(xValues) || len(yValues) < 2 {
		return 0
	}

	yMean := ta.calculateMean(yValues)

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

// calculateVolatility calculates volatility of the data
func (ta *TrendAnalyzer) calculateVolatility(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	mean := ta.calculateMean(values)
	variance := 0.0

	for _, value := range values {
		diff := value - mean
		variance += diff * diff
	}

	variance /= float64(len(values) - 1)
	stdDev := math.Sqrt(variance)

	// Return coefficient of variation
	if mean == 0 {
		return 0
	}

	return stdDev / math.Abs(mean)
}

// calculateMean calculates mean of values
func (ta *TrendAnalyzer) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// DetectSeasonality detects seasonal patterns in data
func (ta *TrendAnalyzer) DetectSeasonality(historicalData []HistoricalDataPoint) (*SeasonalityAnalysis, error) {
	if len(historicalData) < 12 { // Need at least 12 data points for seasonality
		return nil, fmt.Errorf("insufficient data for seasonality analysis: %d", len(historicalData))
	}

	// Group data by month
	monthlyData := make(map[int][]float64)
	for _, point := range historicalData {
		month := int(point.Timestamp.Month())
		monthlyData[month] = append(monthlyData[month], point.Value)
	}

	// Calculate monthly averages
	monthlyAverages := make(map[int]float64)
	for month, values := range monthlyData {
		monthlyAverages[month] = ta.calculateMean(values)
	}

	// Calculate overall mean
	allValues := make([]float64, 0)
	for _, values := range monthlyData {
		allValues = append(allValues, values...)
	}
	overallMean := ta.calculateMean(allValues)

	// Calculate seasonal indices
	seasonalIndices := make(map[int]float64)
	for month, avg := range monthlyAverages {
		seasonalIndices[month] = avg / overallMean
	}

	// Detect if there's significant seasonality
	seasonalityStrength := ta.calculateSeasonalityStrength(seasonalIndices)

	return &SeasonalityAnalysis{
		HasSeasonality:      seasonalityStrength > 0.1,
		SeasonalityStrength: seasonalityStrength,
		SeasonalIndices:     seasonalIndices,
		MonthlyAverages:     monthlyAverages,
	}, nil
}

// SeasonalityAnalysis contains seasonality analysis results
type SeasonalityAnalysis struct {
	HasSeasonality      bool            `json:"has_seasonality"`
	SeasonalityStrength float64         `json:"seasonality_strength"`
	SeasonalIndices     map[int]float64 `json:"seasonal_indices"`
	MonthlyAverages     map[int]float64 `json:"monthly_averages"`
}

// calculateSeasonalityStrength calculates the strength of seasonality
func (ta *TrendAnalyzer) calculateSeasonalityStrength(seasonalIndices map[int]float64) float64 {
	if len(seasonalIndices) == 0 {
		return 0
	}

	// Calculate variance of seasonal indices
	indices := make([]float64, 0, len(seasonalIndices))
	for _, index := range seasonalIndices {
		indices = append(indices, index)
	}

	mean := ta.calculateMean(indices)
	variance := 0.0

	for _, index := range indices {
		diff := index - mean
		variance += diff * diff
	}

	variance /= float64(len(indices))

	// Return coefficient of variation as seasonality strength
	if mean == 0 {
		return 0
	}

	return math.Sqrt(variance) / math.Abs(mean)
}
