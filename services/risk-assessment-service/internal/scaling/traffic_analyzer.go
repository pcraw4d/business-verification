package scaling

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TrafficAnalyzer analyzes traffic patterns for predictive scaling
type TrafficAnalyzer struct {
	logger   *zap.Logger
	mu       sync.RWMutex
	stats    *TrafficStats
	config   *TrafficAnalyzerConfig
	patterns map[string]*TrafficPattern
	history  []*TrafficDataPoint
}

// TrafficStats represents statistics for traffic analysis
type TrafficStats struct {
	TotalDataPoints  int64         `json:"total_data_points"`
	AverageRequests  float64       `json:"average_requests"`
	PeakRequests     float64       `json:"peak_requests"`
	ValleyRequests   float64       `json:"valley_requests"`
	TrafficVariance  float64       `json:"traffic_variance"`
	PatternAccuracy  float64       `json:"pattern_accuracy"`
	LastAnalysis     time.Time     `json:"last_analysis"`
	AnalysisDuration time.Duration `json:"analysis_duration"`
}

// TrafficAnalyzerConfig represents configuration for traffic analysis
type TrafficAnalyzerConfig struct {
	HistoryWindow     time.Duration `json:"history_window"`
	AnalysisInterval  time.Duration `json:"analysis_interval"`
	MinDataPoints     int           `json:"min_data_points"`
	PatternThreshold  float64       `json:"pattern_threshold"`
	EnablePredictions bool          `json:"enable_predictions"`
	PredictionHorizon time.Duration `json:"prediction_horizon"`
	EnableMetrics     bool          `json:"enable_metrics"`
	EnableLogging     bool          `json:"enable_logging"`
}

// TrafficPattern represents a detected traffic pattern
type TrafficPattern struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Type           string                 `json:"type"`
	Confidence     float64                `json:"confidence"`
	Frequency      time.Duration          `json:"frequency"`
	Amplitude      float64                `json:"amplitude"`
	Phase          float64                `json:"phase"`
	Trend          string                 `json:"trend"`
	Seasonality    map[string]interface{} `json:"seasonality"`
	LastDetected   time.Time              `json:"last_detected"`
	DetectionCount int64                  `json:"detection_count"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// TrafficDataPoint represents a single traffic data point
type TrafficDataPoint struct {
	Timestamp    time.Time              `json:"timestamp"`
	Requests     float64                `json:"requests"`
	ResponseTime float64                `json:"response_time"`
	ErrorRate    float64                `json:"error_rate"`
	CPUUsage     float64                `json:"cpu_usage"`
	MemoryUsage  float64                `json:"memory_usage"`
	ActiveUsers  int64                  `json:"active_users"`
	QueueLength  int64                  `json:"queue_length"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// TrafficPrediction represents a traffic prediction
type TrafficPrediction struct {
	ID                string                 `json:"id"`
	Timestamp         time.Time              `json:"timestamp"`
	PredictedRequests float64                `json:"predicted_requests"`
	Confidence        float64                `json:"confidence"`
	PredictionType    string                 `json:"prediction_type"`
	Horizon           time.Duration          `json:"horizon"`
	Factors           []string               `json:"factors"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// NewTrafficAnalyzer creates a new traffic analyzer
func NewTrafficAnalyzer(config *TrafficAnalyzerConfig, logger *zap.Logger) *TrafficAnalyzer {
	if config == nil {
		config = &TrafficAnalyzerConfig{
			HistoryWindow:     24 * time.Hour,
			AnalysisInterval:  5 * time.Minute,
			MinDataPoints:     100,
			PatternThreshold:  0.7,
			EnablePredictions: true,
			PredictionHorizon: 1 * time.Hour,
			EnableMetrics:     true,
			EnableLogging:     true,
		}
	}

	return &TrafficAnalyzer{
		logger:   logger,
		stats:    &TrafficStats{},
		config:   config,
		patterns: make(map[string]*TrafficPattern),
		history:  make([]*TrafficDataPoint, 0),
	}
}

// AddDataPoint adds a new traffic data point
func (ta *TrafficAnalyzer) AddDataPoint(dataPoint *TrafficDataPoint) error {
	ta.mu.Lock()
	defer ta.mu.Unlock()

	// Add to history
	ta.history = append(ta.history, dataPoint)

	// Remove old data points
	cutoff := time.Now().Add(-ta.config.HistoryWindow)
	for i, point := range ta.history {
		if point.Timestamp.After(cutoff) {
			ta.history = ta.history[i:]
			break
		}
	}

	// Update statistics
	ta.stats.TotalDataPoints++
	ta.updateStats()

	ta.logger.Debug("Traffic data point added",
		zap.Time("timestamp", dataPoint.Timestamp),
		zap.Float64("requests", dataPoint.Requests),
		zap.Float64("response_time", dataPoint.ResponseTime),
		zap.Float64("error_rate", dataPoint.ErrorRate))

	return nil
}

// AnalyzeTraffic analyzes traffic patterns
func (ta *TrafficAnalyzer) AnalyzeTraffic(ctx context.Context) error {
	start := time.Now()

	ta.mu.RLock()
	dataPoints := make([]*TrafficDataPoint, len(ta.history))
	copy(dataPoints, ta.history)
	ta.mu.RUnlock()

	if len(dataPoints) < ta.config.MinDataPoints {
		ta.logger.Debug("Insufficient data points for analysis",
			zap.Int("available", len(dataPoints)),
			zap.Int("required", ta.config.MinDataPoints))
		return nil
	}

	ta.logger.Info("Starting traffic analysis",
		zap.Int("data_points", len(dataPoints)))

	// Detect patterns
	patterns, err := ta.detectPatterns(dataPoints)
	if err != nil {
		return fmt.Errorf("failed to detect patterns: %w", err)
	}

	// Update patterns
	ta.mu.Lock()
	for _, pattern := range patterns {
		ta.patterns[pattern.ID] = pattern
	}
	ta.mu.Unlock()

	// Update statistics
	ta.mu.Lock()
	ta.stats.LastAnalysis = time.Now()
	ta.stats.AnalysisDuration = time.Since(start)
	ta.mu.Unlock()

	ta.logger.Info("Traffic analysis completed",
		zap.Int("patterns_detected", len(patterns)),
		zap.Duration("analysis_duration", time.Since(start)))

	return nil
}

// PredictTraffic predicts future traffic based on patterns
func (ta *TrafficAnalyzer) PredictTraffic(ctx context.Context, horizon time.Duration) ([]*TrafficPrediction, error) {
	if !ta.config.EnablePredictions {
		return nil, fmt.Errorf("predictions are disabled")
	}

	ta.mu.RLock()
	dataPoints := make([]*TrafficDataPoint, len(ta.history))
	copy(dataPoints, ta.history)
	patterns := make(map[string]*TrafficPattern)
	for k, v := range ta.patterns {
		patterns[k] = v
	}
	ta.mu.RUnlock()

	if len(dataPoints) < ta.config.MinDataPoints {
		return nil, fmt.Errorf("insufficient data for prediction")
	}

	ta.logger.Info("Starting traffic prediction",
		zap.Duration("horizon", horizon),
		zap.Int("data_points", len(dataPoints)),
		zap.Int("patterns", len(patterns)))

	// Generate predictions
	predictions, err := ta.generatePredictions(dataPoints, patterns, horizon)
	if err != nil {
		return nil, fmt.Errorf("failed to generate predictions: %w", err)
	}

	ta.logger.Info("Traffic prediction completed",
		zap.Int("predictions_generated", len(predictions)))

	return predictions, nil
}

// GetStats returns traffic analysis statistics
func (ta *TrafficAnalyzer) GetStats() *TrafficStats {
	ta.mu.RLock()
	defer ta.mu.RUnlock()

	stats := *ta.stats
	return &stats
}

// GetPatterns returns detected traffic patterns
func (ta *TrafficAnalyzer) GetPatterns() map[string]*TrafficPattern {
	ta.mu.RLock()
	defer ta.mu.RUnlock()

	patterns := make(map[string]*TrafficPattern)
	for k, v := range ta.patterns {
		patterns[k] = v
	}

	return patterns
}

// GetHistory returns traffic history
func (ta *TrafficAnalyzer) GetHistory() []*TrafficDataPoint {
	ta.mu.RLock()
	defer ta.mu.RUnlock()

	history := make([]*TrafficDataPoint, len(ta.history))
	copy(history, ta.history)

	return history
}

// Helper methods

func (ta *TrafficAnalyzer) updateStats() {
	if len(ta.history) == 0 {
		return
	}

	var totalRequests, peakRequests, valleyRequests float64
	var responseTimes, errorRates []float64

	peakRequests = ta.history[0].Requests
	valleyRequests = ta.history[0].Requests

	for _, point := range ta.history {
		totalRequests += point.Requests
		responseTimes = append(responseTimes, point.ResponseTime)
		errorRates = append(errorRates, point.ErrorRate)

		if point.Requests > peakRequests {
			peakRequests = point.Requests
		}
		if point.Requests < valleyRequests {
			valleyRequests = point.Requests
		}
	}

	ta.stats.AverageRequests = totalRequests / float64(len(ta.history))
	ta.stats.PeakRequests = peakRequests
	ta.stats.ValleyRequests = valleyRequests

	// Calculate variance
	var variance float64
	for _, point := range ta.history {
		diff := point.Requests - ta.stats.AverageRequests
		variance += diff * diff
	}
	ta.stats.TrafficVariance = variance / float64(len(ta.history))
}

func (ta *TrafficAnalyzer) detectPatterns(dataPoints []*TrafficDataPoint) ([]*TrafficPattern, error) {
	var patterns []*TrafficPattern

	// Detect daily patterns
	dailyPattern, err := ta.detectDailyPattern(dataPoints)
	if err == nil && dailyPattern != nil {
		patterns = append(patterns, dailyPattern)
	}

	// Detect weekly patterns
	weeklyPattern, err := ta.detectWeeklyPattern(dataPoints)
	if err == nil && weeklyPattern != nil {
		patterns = append(patterns, weeklyPattern)
	}

	// Detect trend patterns
	trendPattern, err := ta.detectTrendPattern(dataPoints)
	if err == nil && trendPattern != nil {
		patterns = append(patterns, trendPattern)
	}

	// Detect seasonal patterns
	seasonalPattern, err := ta.detectSeasonalPattern(dataPoints)
	if err == nil && seasonalPattern != nil {
		patterns = append(patterns, seasonalPattern)
	}

	return patterns, nil
}

func (ta *TrafficAnalyzer) detectDailyPattern(dataPoints []*TrafficDataPoint) (*TrafficPattern, error) {
	if len(dataPoints) < 24 {
		return nil, fmt.Errorf("insufficient data for daily pattern detection")
	}

	// Group by hour
	hourlyData := make(map[int][]float64)
	for _, point := range dataPoints {
		hour := point.Timestamp.Hour()
		hourlyData[hour] = append(hourlyData[hour], point.Requests)
	}

	// Calculate average for each hour
	hourlyAverages := make(map[int]float64)
	for hour, values := range hourlyData {
		var sum float64
		for _, value := range values {
			sum += value
		}
		hourlyAverages[hour] = sum / float64(len(values))
	}

	// Calculate pattern confidence
	// Convert map[int]float64 to map[interface{}]float64 for pattern confidence calculation
	hourlyAveragesInterface := make(map[interface{}]float64)
	for hour, avg := range hourlyAverages {
		hourlyAveragesInterface[hour] = avg
	}
	confidence := ta.calculatePatternConfidence(hourlyAveragesInterface)

	if confidence < ta.config.PatternThreshold {
		return nil, fmt.Errorf("daily pattern confidence too low: %f", confidence)
	}

	pattern := &TrafficPattern{
		ID:             "daily_pattern",
		Name:           "Daily Traffic Pattern",
		Type:           "daily",
		Confidence:     confidence,
		Frequency:      24 * time.Hour,
		LastDetected:   time.Now(),
		DetectionCount: 1,
		Metadata: map[string]interface{}{
			"hourly_averages": hourlyAverages,
		},
	}

	ta.logger.Info("Daily pattern detected",
		zap.Float64("confidence", confidence))

	return pattern, nil
}

func (ta *TrafficAnalyzer) detectWeeklyPattern(dataPoints []*TrafficDataPoint) (*TrafficPattern, error) {
	if len(dataPoints) < 168 { // 7 days * 24 hours
		return nil, fmt.Errorf("insufficient data for weekly pattern detection")
	}

	// Group by day of week
	dailyData := make(map[time.Weekday][]float64)
	for _, point := range dataPoints {
		day := point.Timestamp.Weekday()
		dailyData[day] = append(dailyData[day], point.Requests)
	}

	// Calculate average for each day
	dailyAverages := make(map[time.Weekday]float64)
	for day, values := range dailyData {
		var sum float64
		for _, value := range values {
			sum += value
		}
		dailyAverages[day] = sum / float64(len(values))
	}

	// Calculate pattern confidence
	// Convert map[time.Weekday]float64 to map[interface{}]float64 for pattern confidence calculation
	dailyAveragesInterface := make(map[interface{}]float64)
	for day, avg := range dailyAverages {
		dailyAveragesInterface[day] = avg
	}
	confidence := ta.calculatePatternConfidence(dailyAveragesInterface)

	if confidence < ta.config.PatternThreshold {
		return nil, fmt.Errorf("weekly pattern confidence too low: %f", confidence)
	}

	pattern := &TrafficPattern{
		ID:             "weekly_pattern",
		Name:           "Weekly Traffic Pattern",
		Type:           "weekly",
		Confidence:     confidence,
		Frequency:      7 * 24 * time.Hour,
		LastDetected:   time.Now(),
		DetectionCount: 1,
		Metadata: map[string]interface{}{
			"daily_averages": dailyAverages,
		},
	}

	ta.logger.Info("Weekly pattern detected",
		zap.Float64("confidence", confidence))

	return pattern, nil
}

func (ta *TrafficAnalyzer) detectTrendPattern(dataPoints []*TrafficDataPoint) (*TrafficPattern, error) {
	if len(dataPoints) < 10 {
		return nil, fmt.Errorf("insufficient data for trend pattern detection")
	}

	// Calculate linear trend
	var sumX, sumY, sumXY, sumXX float64
	n := float64(len(dataPoints))

	for i, point := range dataPoints {
		x := float64(i)
		y := point.Requests
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	// Calculate slope
	slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)

	// Determine trend
	var trend string
	if slope > 0.1 {
		trend = "increasing"
	} else if slope < -0.1 {
		trend = "decreasing"
	} else {
		trend = "stable"
	}

	// Calculate confidence based on R-squared
	yMean := sumY / n
	var ssRes, ssTot float64
	for i, point := range dataPoints {
		x := float64(i)
		y := point.Requests
		yPred := slope*x + (sumY-slope*sumX)/n
		ssRes += (y - yPred) * (y - yPred)
		ssTot += (y - yMean) * (y - yMean)
	}

	confidence := 1 - (ssRes / ssTot)
	if confidence < 0 {
		confidence = 0
	}

	if confidence < ta.config.PatternThreshold {
		return nil, fmt.Errorf("trend pattern confidence too low: %f", confidence)
	}

	pattern := &TrafficPattern{
		ID:             "trend_pattern",
		Name:           "Traffic Trend Pattern",
		Type:           "trend",
		Confidence:     confidence,
		Trend:          trend,
		LastDetected:   time.Now(),
		DetectionCount: 1,
		Metadata: map[string]interface{}{
			"slope":     slope,
			"r_squared": confidence,
		},
	}

	ta.logger.Info("Trend pattern detected",
		zap.String("trend", trend),
		zap.Float64("slope", slope),
		zap.Float64("confidence", confidence))

	return pattern, nil
}

func (ta *TrafficAnalyzer) detectSeasonalPattern(dataPoints []*TrafficDataPoint) (*TrafficPattern, error) {
	if len(dataPoints) < 720 { // 30 days * 24 hours
		return nil, fmt.Errorf("insufficient data for seasonal pattern detection")
	}

	// Group by month
	monthlyData := make(map[int][]float64)
	for _, point := range dataPoints {
		month := int(point.Timestamp.Month())
		monthlyData[month] = append(monthlyData[month], point.Requests)
	}

	// Calculate average for each month
	monthlyAverages := make(map[int]float64)
	for month, values := range monthlyData {
		var sum float64
		for _, value := range values {
			sum += value
		}
		monthlyAverages[month] = sum / float64(len(values))
	}

	// Calculate pattern confidence
	// Convert map[int]float64 to map[interface{}]float64 for pattern confidence calculation
	monthlyAveragesInterface := make(map[interface{}]float64)
	for month, avg := range monthlyAverages {
		monthlyAveragesInterface[month] = avg
	}
	confidence := ta.calculatePatternConfidence(monthlyAveragesInterface)

	if confidence < ta.config.PatternThreshold {
		return nil, fmt.Errorf("seasonal pattern confidence too low: %f", confidence)
	}

	pattern := &TrafficPattern{
		ID:             "seasonal_pattern",
		Name:           "Seasonal Traffic Pattern",
		Type:           "seasonal",
		Confidence:     confidence,
		Frequency:      365 * 24 * time.Hour,
		LastDetected:   time.Now(),
		DetectionCount: 1,
		Seasonality: map[string]interface{}{
			"monthly_averages": monthlyAverages,
		},
		Metadata: map[string]interface{}{
			"monthly_averages": monthlyAverages,
		},
	}

	ta.logger.Info("Seasonal pattern detected",
		zap.Float64("confidence", confidence))

	return pattern, nil
}

func (ta *TrafficAnalyzer) calculatePatternConfidence(data map[interface{}]float64) float64 {
	if len(data) < 2 {
		return 0
	}

	// Calculate coefficient of variation
	var values []float64
	for _, value := range data {
		values = append(values, value)
	}

	var sum, sumSquares float64
	for _, value := range values {
		sum += value
		sumSquares += value * value
	}

	mean := sum / float64(len(values))
	variance := (sumSquares / float64(len(values))) - (mean * mean)
	stdDev := math.Sqrt(variance)

	if mean == 0 {
		return 0
	}

	coefficientOfVariation := stdDev / mean

	// Convert to confidence (lower CV = higher confidence)
	confidence := 1 / (1 + coefficientOfVariation)

	return confidence
}

func (ta *TrafficAnalyzer) generatePredictions(dataPoints []*TrafficDataPoint, patterns map[string]*TrafficPattern, horizon time.Duration) ([]*TrafficPrediction, error) {
	var predictions []*TrafficPrediction

	// Generate predictions based on patterns
	for _, pattern := range patterns {
		patternPredictions, err := ta.generatePatternPredictions(dataPoints, pattern, horizon)
		if err != nil {
			ta.logger.Warn("Failed to generate pattern predictions",
				zap.String("pattern_id", pattern.ID),
				zap.Error(err))
			continue
		}

		predictions = append(predictions, patternPredictions...)
	}

	// If no patterns, use simple trend prediction
	if len(predictions) == 0 {
		trendPredictions := ta.generateTrendPredictions(dataPoints, horizon)
		predictions = append(predictions, trendPredictions...)
	}

	return predictions, nil
}

func (ta *TrafficAnalyzer) generatePatternPredictions(dataPoints []*TrafficDataPoint, pattern *TrafficPattern, horizon time.Duration) ([]*TrafficPrediction, error) {
	var predictions []*TrafficPrediction

	// Generate predictions based on pattern type
	switch pattern.Type {
	case "daily":
		predictions = ta.generateDailyPredictions(dataPoints, pattern, horizon)
	case "weekly":
		predictions = ta.generateWeeklyPredictions(dataPoints, pattern, horizon)
	case "trend":
		predictions = ta.generateTrendPredictions(dataPoints, horizon)
	case "seasonal":
		predictions = ta.generateSeasonalPredictions(dataPoints, pattern, horizon)
	default:
		return nil, fmt.Errorf("unknown pattern type: %s", pattern.Type)
	}

	return predictions, nil
}

func (ta *TrafficAnalyzer) generateDailyPredictions(dataPoints []*TrafficDataPoint, pattern *TrafficPattern, horizon time.Duration) []*TrafficPrediction {
	var predictions []*TrafficPrediction

	// Get hourly averages from pattern metadata
	hourlyAverages, ok := pattern.Metadata["hourly_averages"].(map[int]float64)
	if !ok {
		return predictions
	}

	// Generate predictions for each hour in the horizon
	now := time.Now()
	for i := 0; i < int(horizon.Hours()); i++ {
		futureTime := now.Add(time.Duration(i) * time.Hour)
		hour := futureTime.Hour()

		if avg, exists := hourlyAverages[hour]; exists {
			prediction := &TrafficPrediction{
				ID:                fmt.Sprintf("daily_pred_%d", i),
				Timestamp:         futureTime,
				PredictedRequests: avg,
				Confidence:        pattern.Confidence,
				PredictionType:    "daily_pattern",
				Horizon:           time.Duration(i) * time.Hour,
				Factors:           []string{"daily_pattern", "hour_of_day"},
				Metadata: map[string]interface{}{
					"pattern_id": pattern.ID,
					"hour":       hour,
				},
			}
			predictions = append(predictions, prediction)
		}
	}

	return predictions
}

func (ta *TrafficAnalyzer) generateWeeklyPredictions(dataPoints []*TrafficDataPoint, pattern *TrafficPattern, horizon time.Duration) []*TrafficPrediction {
	var predictions []*TrafficPrediction

	// Get daily averages from pattern metadata
	dailyAverages, ok := pattern.Metadata["daily_averages"].(map[time.Weekday]float64)
	if !ok {
		return predictions
	}

	// Generate predictions for each day in the horizon
	now := time.Now()
	for i := 0; i < int(horizon.Hours()/24); i++ {
		futureTime := now.Add(time.Duration(i) * 24 * time.Hour)
		day := futureTime.Weekday()

		if avg, exists := dailyAverages[day]; exists {
			prediction := &TrafficPrediction{
				ID:                fmt.Sprintf("weekly_pred_%d", i),
				Timestamp:         futureTime,
				PredictedRequests: avg,
				Confidence:        pattern.Confidence,
				PredictionType:    "weekly_pattern",
				Horizon:           time.Duration(i) * 24 * time.Hour,
				Factors:           []string{"weekly_pattern", "day_of_week"},
				Metadata: map[string]interface{}{
					"pattern_id": pattern.ID,
					"day":        day,
				},
			}
			predictions = append(predictions, prediction)
		}
	}

	return predictions
}

func (ta *TrafficAnalyzer) generateTrendPredictions(dataPoints []*TrafficDataPoint, horizon time.Duration) []*TrafficPrediction {
	var predictions []*TrafficPrediction

	if len(dataPoints) < 2 {
		return predictions
	}

	// Calculate simple linear trend
	lastPoint := dataPoints[len(dataPoints)-1]
	firstPoint := dataPoints[0]

	timeDiff := lastPoint.Timestamp.Sub(firstPoint.Timestamp).Hours()
	requestDiff := lastPoint.Requests - firstPoint.Requests

	slope := requestDiff / timeDiff

	// Generate predictions
	now := time.Now()
	for i := 1; i <= int(horizon.Hours()); i++ {
		futureTime := now.Add(time.Duration(i) * time.Hour)
		predictedRequests := lastPoint.Requests + (slope * float64(i))

		// Ensure non-negative predictions
		if predictedRequests < 0 {
			predictedRequests = 0
		}

		prediction := &TrafficPrediction{
			ID:                fmt.Sprintf("trend_pred_%d", i),
			Timestamp:         futureTime,
			PredictedRequests: predictedRequests,
			Confidence:        0.5, // Lower confidence for trend predictions
			PredictionType:    "trend",
			Horizon:           time.Duration(i) * time.Hour,
			Factors:           []string{"linear_trend"},
			Metadata: map[string]interface{}{
				"slope": slope,
			},
		}
		predictions = append(predictions, prediction)
	}

	return predictions
}

func (ta *TrafficAnalyzer) generateSeasonalPredictions(dataPoints []*TrafficDataPoint, pattern *TrafficPattern, horizon time.Duration) []*TrafficPrediction {
	var predictions []*TrafficPrediction

	// Get monthly averages from pattern metadata
	monthlyAverages, ok := pattern.Metadata["monthly_averages"].(map[int]float64)
	if !ok {
		return predictions
	}

	// Generate predictions for each month in the horizon
	now := time.Now()
	for i := 0; i < int(horizon.Hours()/(24*30)); i++ {
		futureTime := now.Add(time.Duration(i) * 24 * 30 * time.Hour)
		month := int(futureTime.Month())

		if avg, exists := monthlyAverages[month]; exists {
			prediction := &TrafficPrediction{
				ID:                fmt.Sprintf("seasonal_pred_%d", i),
				Timestamp:         futureTime,
				PredictedRequests: avg,
				Confidence:        pattern.Confidence,
				PredictionType:    "seasonal_pattern",
				Horizon:           time.Duration(i) * 24 * 30 * time.Hour,
				Factors:           []string{"seasonal_pattern", "month"},
				Metadata: map[string]interface{}{
					"pattern_id": pattern.ID,
					"month":      month,
				},
			}
			predictions = append(predictions, prediction)
		}
	}

	return predictions
}
