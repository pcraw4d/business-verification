package external

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// VerificationSuccessMonitor tracks and analyzes verification success rates
type VerificationSuccessMonitor struct {
	config    *SuccessMonitorConfig
	logger    *zap.Logger
	metrics   *SuccessMetrics
	mu        sync.RWMutex
	startTime time.Time
}

// SuccessMonitorConfig holds configuration for success rate monitoring
type SuccessMonitorConfig struct {
	EnableRealTimeMonitoring bool          `json:"enable_real_time_monitoring"`
	EnableFailureAnalysis    bool          `json:"enable_failure_analysis"`
	EnableTrendAnalysis      bool          `json:"enable_trend_analysis"`
	EnableAlerting           bool          `json:"enable_alerting"`
	TargetSuccessRate        float64       `json:"target_success_rate"`      // 0.90 = 90%
	AlertThreshold           float64       `json:"alert_threshold"`          // 0.85 = 85%
	MetricsRetentionPeriod   time.Duration `json:"metrics_retention_period"` // 30 days
	AnalysisWindow           time.Duration `json:"analysis_window"`          // 1 hour
	TrendWindow              time.Duration `json:"trend_window"`             // 24 hours
	MinDataPoints            int           `json:"min_data_points"`          // 100
	MaxDataPoints            int           `json:"max_data_points"`          // 10000
}

// SuccessMetrics holds aggregated success rate metrics
type SuccessMetrics struct {
	TotalAttempts       int64         `json:"total_attempts"`
	SuccessfulAttempts  int64         `json:"successful_attempts"`
	FailedAttempts      int64         `json:"failed_attempts"`
	SuccessRate         float64       `json:"success_rate"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastUpdated         time.Time     `json:"last_updated"`
	DataPoints          []DataPoint   `json:"data_points"`
}

// DataPoint represents a single verification attempt
type DataPoint struct {
	Timestamp     time.Time              `json:"timestamp"`
	URL           string                 `json:"url"`
	Success       bool                   `json:"success"`
	ResponseTime  time.Duration          `json:"response_time"`
	StatusCode    int                    `json:"status_code"`
	ErrorType     string                 `json:"error_type,omitempty"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	StrategyUsed  string                 `json:"strategy_used,omitempty"`
	UserAgentUsed string                 `json:"user_agent_used,omitempty"`
	ProxyUsed     *Proxy                 `json:"proxy_used,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// FailureAnalysis holds analysis of verification failures
type FailureAnalysis struct {
	TotalFailures     int64                   `json:"total_failures"`
	FailureRate       float64                 `json:"failure_rate"`
	CommonErrorTypes  map[string]int64        `json:"common_error_types"`
	ProblematicURLs   map[string]int64        `json:"problematic_urls"`
	StrategyFailures  map[string]int64        `json:"strategy_failures"`
	TimeBasedPatterns map[string]int64        `json:"time_based_patterns"`
	Recommendations   []FailureRecommendation `json:"recommendations"`
	LastAnalyzed      time.Time               `json:"last_analyzed"`
}

// FailureRecommendation provides actionable recommendations for improving success rates
type FailureRecommendation struct {
	Type        string  `json:"type"`     // "strategy", "url", "timing", "configuration"
	Priority    string  `json:"priority"` // "high", "medium", "low"
	Description string  `json:"description"`
	Impact      float64 `json:"impact"` // Estimated improvement in success rate
	Action      string  `json:"action"` // Recommended action
}

// TrendAnalysis holds trend analysis data
type TrendAnalysis struct {
	Period            time.Duration      `json:"period"`
	SuccessRateTrend  float64            `json:"success_rate_trend"`  // Positive = improving, Negative = declining
	VolumeTrend       float64            `json:"volume_trend"`        // Change in verification volume
	ResponseTimeTrend float64            `json:"response_time_trend"` // Change in average response time
	Seasonality       map[string]float64 `json:"seasonality"`         // Hourly/daily patterns
	Predictions       []Prediction       `json:"predictions"`
	LastUpdated       time.Time          `json:"last_updated"`
}

// Prediction holds future success rate predictions
type Prediction struct {
	Timestamp   time.Time `json:"timestamp"`
	SuccessRate float64   `json:"success_rate"`
	Confidence  float64   `json:"confidence"` // 0.0 to 1.0
}

// NewVerificationSuccessMonitor creates a new success rate monitor
func NewVerificationSuccessMonitor(config *SuccessMonitorConfig, logger *zap.Logger) *VerificationSuccessMonitor {
	if config == nil {
		config = DefaultSuccessMonitorConfig()
	}

	monitor := &VerificationSuccessMonitor{
		config:    config,
		logger:    logger,
		metrics:   &SuccessMetrics{},
		startTime: time.Now(),
	}

	// Start background analysis if enabled
	if config.EnableRealTimeMonitoring {
		go monitor.startBackgroundAnalysis()
	}

	return monitor
}

// DefaultSuccessMonitorConfig returns default configuration
func DefaultSuccessMonitorConfig() *SuccessMonitorConfig {
	return &SuccessMonitorConfig{
		EnableRealTimeMonitoring: true,
		EnableFailureAnalysis:    true,
		EnableTrendAnalysis:      true,
		EnableAlerting:           true,
		TargetSuccessRate:        0.90,                // 90%
		AlertThreshold:           0.85,                // 85%
		MetricsRetentionPeriod:   30 * 24 * time.Hour, // 30 days
		AnalysisWindow:           1 * time.Hour,
		TrendWindow:              24 * time.Hour,
		MinDataPoints:            100,
		MaxDataPoints:            10000,
	}
}

// RecordAttempt records a verification attempt
func (m *VerificationSuccessMonitor) RecordAttempt(ctx context.Context, dataPoint DataPoint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Add timestamp if not set
	if dataPoint.Timestamp.IsZero() {
		dataPoint.Timestamp = time.Now()
	}

	// Update metrics
	m.metrics.TotalAttempts++
	if dataPoint.Success {
		m.metrics.SuccessfulAttempts++
	} else {
		m.metrics.FailedAttempts++
	}

	// Calculate success rate
	if m.metrics.TotalAttempts > 0 {
		m.metrics.SuccessRate = float64(m.metrics.SuccessfulAttempts) / float64(m.metrics.TotalAttempts)
	}

	// Update average response time
	if m.metrics.TotalAttempts == 1 {
		m.metrics.AverageResponseTime = dataPoint.ResponseTime
	} else {
		// Weighted average
		total := m.metrics.AverageResponseTime * time.Duration(m.metrics.TotalAttempts-1)
		m.metrics.AverageResponseTime = (total + dataPoint.ResponseTime) / time.Duration(m.metrics.TotalAttempts)
	}

	// Add data point
	m.metrics.DataPoints = append(m.metrics.DataPoints, dataPoint)
	m.metrics.LastUpdated = time.Now()

	// Cleanup old data points
	m.cleanupOldDataPoints()

	// Check for alerts
	if m.config.EnableAlerting {
		m.checkAlerts(ctx)
	}

	m.logger.Debug("Recorded verification attempt",
		zap.String("url", dataPoint.URL),
		zap.Bool("success", dataPoint.Success),
		zap.Duration("response_time", dataPoint.ResponseTime),
		zap.Float64("current_success_rate", m.metrics.SuccessRate))

	return nil
}

// GetMetrics returns current success rate metrics
func (m *VerificationSuccessMonitor) GetMetrics() *SuccessMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid race conditions
	metrics := *m.metrics
	metrics.DataPoints = make([]DataPoint, len(m.metrics.DataPoints))
	copy(metrics.DataPoints, m.metrics.DataPoints)

	return &metrics
}

// AnalyzeFailures performs failure analysis
func (m *VerificationSuccessMonitor) AnalyzeFailures(ctx context.Context) (*FailureAnalysis, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	analysis := &FailureAnalysis{
		CommonErrorTypes:  make(map[string]int64),
		ProblematicURLs:   make(map[string]int64),
		StrategyFailures:  make(map[string]int64),
		TimeBasedPatterns: make(map[string]int64),
		LastAnalyzed:      time.Now(),
	}

	// Filter data points to analysis window
	cutoff := time.Now().Add(-m.config.AnalysisWindow)
	var recentDataPoints []DataPoint
	for _, dp := range m.metrics.DataPoints {
		if dp.Timestamp.After(cutoff) {
			recentDataPoints = append(recentDataPoints, dp)
		}
	}

	if len(recentDataPoints) == 0 {
		return analysis, nil
	}

	// Analyze failures
	for _, dp := range recentDataPoints {
		if !dp.Success {
			analysis.TotalFailures++

			// Count error types
			if dp.ErrorType != "" {
				analysis.CommonErrorTypes[dp.ErrorType]++
			}

			// Count problematic URLs
			analysis.ProblematicURLs[dp.URL]++

			// Count strategy failures
			if dp.StrategyUsed != "" {
				analysis.StrategyFailures[dp.StrategyUsed]++
			}

			// Count time-based patterns (hour of day)
			hour := dp.Timestamp.Format("15")
			analysis.TimeBasedPatterns[hour]++
		}
	}

	// Calculate failure rate
	totalAttempts := int64(len(recentDataPoints))
	if totalAttempts > 0 {
		analysis.FailureRate = float64(analysis.TotalFailures) / float64(totalAttempts)
	}

	// Generate recommendations
	analysis.Recommendations = m.generateRecommendations(analysis)

	return analysis, nil
}

// AnalyzeTrends performs trend analysis
func (m *VerificationSuccessMonitor) AnalyzeTrends(ctx context.Context) (*TrendAnalysis, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	analysis := &TrendAnalysis{
		Period:      m.config.TrendWindow,
		Seasonality: make(map[string]float64),
		LastUpdated: time.Now(),
	}

	// Filter data points to trend window
	cutoff := time.Now().Add(-m.config.TrendWindow)
	var recentDataPoints []DataPoint
	for _, dp := range m.metrics.DataPoints {
		if dp.Timestamp.After(cutoff) {
			recentDataPoints = append(recentDataPoints, dp)
		}
	}

	if len(recentDataPoints) < m.config.MinDataPoints {
		return analysis, fmt.Errorf("insufficient data points for trend analysis: %d < %d", len(recentDataPoints), m.config.MinDataPoints)
	}

	// Calculate trends
	analysis.SuccessRateTrend = m.calculateSuccessRateTrend(recentDataPoints)
	analysis.VolumeTrend = m.calculateVolumeTrend(recentDataPoints)
	analysis.ResponseTimeTrend = m.calculateResponseTimeTrend(recentDataPoints)

	// Calculate seasonality patterns
	analysis.Seasonality = m.calculateSeasonality(recentDataPoints)

	// Generate predictions
	analysis.Predictions = m.generatePredictions(recentDataPoints)

	return analysis, nil
}

// GetSuccessRate returns the current success rate
func (m *VerificationSuccessMonitor) GetSuccessRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.metrics.SuccessRate
}

// IsTargetAchieved checks if the target success rate is achieved
func (m *VerificationSuccessMonitor) IsTargetAchieved() bool {
	return m.GetSuccessRate() >= m.config.TargetSuccessRate
}

// GetConfig returns the current configuration
func (m *VerificationSuccessMonitor) GetConfig() *SuccessMonitorConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// UpdateConfig updates the monitor configuration
func (m *VerificationSuccessMonitor) UpdateConfig(config *SuccessMonitorConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate configuration
	if config.TargetSuccessRate < 0 || config.TargetSuccessRate > 1 {
		return fmt.Errorf("target success rate must be between 0 and 1")
	}

	if config.AlertThreshold < 0 || config.AlertThreshold > 1 {
		return fmt.Errorf("alert threshold must be between 0 and 1")
	}

	if config.AlertThreshold >= config.TargetSuccessRate {
		return fmt.Errorf("alert threshold must be less than target success rate")
	}

	m.config = config

	m.logger.Info("Updated success monitor configuration",
		zap.Float64("target_success_rate", config.TargetSuccessRate),
		zap.Float64("alert_threshold", config.AlertThreshold))

	return nil
}

// ResetMetrics resets all metrics (useful for testing)
func (m *VerificationSuccessMonitor) ResetMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics = &SuccessMetrics{}
	m.startTime = time.Now()

	m.logger.Info("Reset success monitor metrics")
}

// Private helper methods

func (m *VerificationSuccessMonitor) cleanupOldDataPoints() {
	if len(m.metrics.DataPoints) <= m.config.MaxDataPoints {
		return
	}

	// Remove oldest data points
	cutoff := time.Now().Add(-m.config.MetricsRetentionPeriod)
	var validDataPoints []DataPoint
	for _, dp := range m.metrics.DataPoints {
		if dp.Timestamp.After(cutoff) {
			validDataPoints = append(validDataPoints, dp)
		}
	}

	// If still too many, keep only the most recent
	if len(validDataPoints) > m.config.MaxDataPoints {
		start := len(validDataPoints) - m.config.MaxDataPoints
		validDataPoints = validDataPoints[start:]
	}

	m.metrics.DataPoints = validDataPoints
}

func (m *VerificationSuccessMonitor) checkAlerts(ctx context.Context) {
	currentRate := m.metrics.SuccessRate

	if currentRate < m.config.AlertThreshold {
		m.logger.Warn("Success rate below alert threshold",
			zap.Float64("current_rate", currentRate),
			zap.Float64("alert_threshold", m.config.AlertThreshold),
			zap.Float64("target_rate", m.config.TargetSuccessRate))
	}

	if currentRate >= m.config.TargetSuccessRate {
		m.logger.Info("Target success rate achieved",
			zap.Float64("current_rate", currentRate),
			zap.Float64("target_rate", m.config.TargetSuccessRate))
	}
}

func (m *VerificationSuccessMonitor) generateRecommendations(analysis *FailureAnalysis) []FailureRecommendation {
	var recommendations []FailureRecommendation

	// Strategy-based recommendations
	for strategy, failures := range analysis.StrategyFailures {
		if failures > 10 { // Threshold for significant failures
			recommendations = append(recommendations, FailureRecommendation{
				Type:        "strategy",
				Priority:    "high",
				Description: fmt.Sprintf("Strategy '%s' has %d failures", strategy, failures),
				Impact:      0.05, // Estimated 5% improvement
				Action:      fmt.Sprintf("Review and optimize strategy '%s'", strategy),
			})
		}
	}

	// URL-based recommendations
	for url, failures := range analysis.ProblematicURLs {
		if failures > 5 { // Threshold for problematic URLs
			recommendations = append(recommendations, FailureRecommendation{
				Type:        "url",
				Priority:    "medium",
				Description: fmt.Sprintf("URL '%s' has %d failures", url, failures),
				Impact:      0.02, // Estimated 2% improvement
				Action:      fmt.Sprintf("Investigate and optimize handling for URL '%s'", url),
			})
		}
	}

	// Error type recommendations
	for errorType, count := range analysis.CommonErrorTypes {
		if count > 20 { // Threshold for common errors
			recommendations = append(recommendations, FailureRecommendation{
				Type:        "configuration",
				Priority:    "high",
				Description: fmt.Sprintf("Error type '%s' occurs %d times", errorType, count),
				Impact:      0.08, // Estimated 8% improvement
				Action:      fmt.Sprintf("Implement better handling for error type '%s'", errorType),
			})
		}
	}

	return recommendations
}

func (m *VerificationSuccessMonitor) calculateSuccessRateTrend(dataPoints []DataPoint) float64 {
	if len(dataPoints) < 2 {
		return 0
	}

	// Split data points into two halves
	mid := len(dataPoints) / 2
	firstHalf := dataPoints[:mid]
	secondHalf := dataPoints[mid:]

	// Calculate success rates for each half
	firstRate := m.calculateSuccessRate(firstHalf)
	secondRate := m.calculateSuccessRate(secondHalf)

	// Return the difference (positive = improving)
	return secondRate - firstRate
}

func (m *VerificationSuccessMonitor) calculateVolumeTrend(dataPoints []DataPoint) float64 {
	if len(dataPoints) < 2 {
		return 0
	}

	// Split data points into two halves
	mid := len(dataPoints) / 2
	firstHalf := dataPoints[:mid]
	secondHalf := dataPoints[mid:]

	// Calculate volumes for each half
	firstVolume := float64(len(firstHalf))
	secondVolume := float64(len(secondHalf))

	// Return the percentage change
	if firstVolume == 0 {
		return 0
	}
	return (secondVolume - firstVolume) / firstVolume
}

func (m *VerificationSuccessMonitor) calculateResponseTimeTrend(dataPoints []DataPoint) float64 {
	if len(dataPoints) < 2 {
		return 0
	}

	// Split data points into two halves
	mid := len(dataPoints) / 2
	firstHalf := dataPoints[:mid]
	secondHalf := dataPoints[mid:]

	// Calculate average response times for each half
	firstAvg := m.calculateAverageResponseTime(firstHalf)
	secondAvg := m.calculateAverageResponseTime(secondHalf)

	// Return the percentage change (negative = improving)
	if firstAvg == 0 {
		return 0
	}
	return float64(secondAvg-firstAvg) / float64(firstAvg)
}

func (m *VerificationSuccessMonitor) calculateSeasonality(dataPoints []DataPoint) map[string]float64 {
	seasonality := make(map[string]float64)
	hourlyCounts := make(map[string]int)
	hourlySuccesses := make(map[string]int)

	for _, dp := range dataPoints {
		hour := dp.Timestamp.Format("15")
		hourlyCounts[hour]++
		if dp.Success {
			hourlySuccesses[hour]++
		}
	}

	for hour, count := range hourlyCounts {
		if count > 0 {
			successRate := float64(hourlySuccesses[hour]) / float64(count)
			seasonality[hour] = successRate
		}
	}

	return seasonality
}

func (m *VerificationSuccessMonitor) generatePredictions(dataPoints []DataPoint) []Prediction {
	var predictions []Prediction

	// Simple linear prediction based on recent trend
	if len(dataPoints) >= 10 {
		trend := m.calculateSuccessRateTrend(dataPoints)
		currentRate := m.calculateSuccessRate(dataPoints)

		// Predict next 6 hours
		for i := 1; i <= 6; i++ {
			predictedRate := currentRate + (trend * float64(i))
			if predictedRate < 0 {
				predictedRate = 0
			} else if predictedRate > 1 {
				predictedRate = 1
			}

			predictions = append(predictions, Prediction{
				Timestamp:   time.Now().Add(time.Duration(i) * time.Hour),
				SuccessRate: predictedRate,
				Confidence:  0.7, // Moderate confidence for simple linear prediction
			})
		}
	}

	return predictions
}

func (m *VerificationSuccessMonitor) calculateSuccessRate(dataPoints []DataPoint) float64 {
	if len(dataPoints) == 0 {
		return 0
	}

	successes := 0
	for _, dp := range dataPoints {
		if dp.Success {
			successes++
		}
	}

	return float64(successes) / float64(len(dataPoints))
}

func (m *VerificationSuccessMonitor) calculateAverageResponseTime(dataPoints []DataPoint) time.Duration {
	if len(dataPoints) == 0 {
		return 0
	}

	total := time.Duration(0)
	for _, dp := range dataPoints {
		total += dp.ResponseTime
	}

	return total / time.Duration(len(dataPoints))
}

func (m *VerificationSuccessMonitor) startBackgroundAnalysis() {
	ticker := time.NewTicker(m.config.AnalysisWindow)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()

			// Perform failure analysis
			if m.config.EnableFailureAnalysis {
				analysis, err := m.AnalyzeFailures(ctx)
				if err != nil {
					m.logger.Error("Failed to analyze failures", zap.Error(err))
				} else {
					m.logger.Debug("Background failure analysis completed",
						zap.Int64("total_failures", analysis.TotalFailures),
						zap.Float64("failure_rate", analysis.FailureRate))
				}
			}

			// Perform trend analysis
			if m.config.EnableTrendAnalysis {
				trends, err := m.AnalyzeTrends(ctx)
				if err != nil {
					m.logger.Error("Failed to analyze trends", zap.Error(err))
				} else {
					m.logger.Debug("Background trend analysis completed",
						zap.Float64("success_rate_trend", trends.SuccessRateTrend),
						zap.Float64("volume_trend", trends.VolumeTrend))
				}
			}
		}
	}
}
