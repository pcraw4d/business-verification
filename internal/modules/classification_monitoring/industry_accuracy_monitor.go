package classification_monitoring

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// IndustryAccuracyMonitor provides detailed industry-specific accuracy monitoring
type IndustryAccuracyMonitor struct {
	config *IndustryAccuracyConfig
	logger *zap.Logger
	mu     sync.RWMutex

	// Core tracking
	industryMetrics  map[string]*DetailedIndustryMetrics
	industryRankings []*IndustryRanking
	industryTrends   map[string]*IndustryTrendAnalysis

	// Performance tracking
	lastUpdate     time.Time
	updateCount    int64
	processingTime time.Duration

	// Alerting
	alertThresholds map[string]*IndustryAlertThresholds
	activeAlerts    map[string]*IndustryAlert
}

// IndustryAccuracyConfig holds configuration for industry accuracy monitoring
type IndustryAccuracyConfig struct {
	EnableDetailedTracking    bool `json:"enable_detailed_tracking"`
	EnableIndustryRankings    bool `json:"enable_industry_rankings"`
	EnableTrendAnalysis       bool `json:"enable_trend_analysis"`
	EnablePerformanceTracking bool `json:"enable_performance_tracking"`

	// Thresholds
	MinSamplesForAnalysis int `json:"min_samples_for_analysis"`
	MinSamplesForRanking  int `json:"min_samples_for_ranking"`
	MinSamplesForTrends   int `json:"min_samples_for_trends"`

	// Update intervals
	UpdateInterval        time.Duration `json:"update_interval"`
	RankingUpdateInterval time.Duration `json:"ranking_update_interval"`
	TrendAnalysisInterval time.Duration `json:"trend_analysis_interval"`

	// Data retention
	MetricsRetentionPeriod  time.Duration `json:"metrics_retention_period"`
	MaxHistoricalDataPoints int           `json:"max_historical_data_points"`

	// Alerting
	EnableAlerting           bool          `json:"enable_alerting"`
	AlertCooldownPeriod      time.Duration `json:"alert_cooldown_period"`
	DefaultAccuracyThreshold float64       `json:"default_accuracy_threshold"`
}

// DetailedIndustryMetrics represents detailed metrics for an industry
type DetailedIndustryMetrics struct {
	IndustryName             string
	TotalClassifications     int64
	CorrectClassifications   int64
	IncorrectClassifications int64
	AccuracyScore            float64
	ConfidenceScore          float64
	AverageProcessingTime    time.Duration

	// Detailed breakdowns
	ConfidenceDistribution   map[string]int64
	MethodPerformance        map[string]*MethodIndustryMetrics
	ErrorTypeDistribution    map[string]int64
	BusinessTypeDistribution map[string]int64

	// Time-based metrics
	HourlyAccuracy map[int]float64
	DailyAccuracy  map[string]float64
	WeeklyAccuracy map[string]float64

	// Historical data
	HistoricalAccuracy       []*AccuracyDataPoint
	HistoricalConfidence     []*ConfidenceDataPoint
	HistoricalProcessingTime []*ProcessingTimeDataPoint

	// Performance indicators
	TrendIndicator   string
	PerformanceGrade string
	ReliabilityScore float64
	ConsistencyScore float64

	// Metadata
	LastUpdated      time.Time
	FirstSeen        time.Time
	DataQualityScore float64
}

// MethodIndustryMetrics represents method performance within an industry
type MethodIndustryMetrics struct {
	MethodName             string
	TotalClassifications   int64
	CorrectClassifications int64
	AccuracyScore          float64
	AverageConfidence      float64
	AverageProcessingTime  time.Duration
	ErrorRate              float64
	LastUpdated            time.Time
}

// IndustryRanking represents industry performance ranking
type IndustryRanking struct {
	Rank                 int
	IndustryName         string
	AccuracyScore        float64
	TotalClassifications int64
	ConfidenceScore      float64
	ReliabilityScore     float64
	PerformanceGrade     string
	TrendIndicator       string
	LastUpdated          time.Time
}

// IndustryTrendAnalysis represents trend analysis for an industry
type IndustryTrendAnalysis struct {
	IndustryName      string
	OverallTrend      string
	TrendStrength     float64
	TrendDirection    string
	PredictedAccuracy float64
	ConfidenceTrend   string
	PerformanceTrend  string

	// Detailed trend data
	AccuracyTrend       []*TrendDataPoint
	ConfidenceTrendData []*TrendDataPoint
	VolumeTrend         []*TrendDataPoint

	// Analysis metadata
	LastAnalysis        time.Time
	AnalysisConfidence  float64
	SeasonalityDetected bool
	AnomaliesDetected   []*AnomalyDetection
}

// IndustryAlertThresholds defines alert thresholds for an industry
type IndustryAlertThresholds struct {
	IndustryName            string
	AccuracyThreshold       float64
	ConfidenceThreshold     float64
	ProcessingTimeThreshold time.Duration
	ErrorRateThreshold      float64
	VolumeThreshold         int64
	CustomThresholds        map[string]float64
}

// IndustryAlert represents an industry-specific alert
type IndustryAlert struct {
	ID             string
	IndustryName   string
	AlertType      string
	Severity       string
	Message        string
	CurrentValue   float64
	ThresholdValue float64
	Timestamp      time.Time
	Status         string
	Actions        []string
	Metadata       map[string]interface{}
}

// Data point structures
type AccuracyDataPoint struct {
	Timestamp  time.Time
	Accuracy   float64
	SampleSize int64
}

type ConfidenceDataPoint struct {
	Timestamp  time.Time
	Confidence float64
	SampleSize int64
}

type ProcessingTimeDataPoint struct {
	Timestamp      time.Time
	ProcessingTime time.Duration
	SampleSize     int64
}

type TrendDataPoint struct {
	Timestamp  time.Time
	Value      float64
	SampleSize int64
}

type AnomalyDetection struct {
	Timestamp     time.Time
	AnomalyType   string
	Severity      string
	Description   string
	Value         float64
	ExpectedValue float64
}

// NewIndustryAccuracyMonitor creates a new industry accuracy monitor
func NewIndustryAccuracyMonitor(config *IndustryAccuracyConfig, logger *zap.Logger) *IndustryAccuracyMonitor {
	if config == nil {
		config = DefaultIndustryAccuracyConfig()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &IndustryAccuracyMonitor{
		config:           config,
		logger:           logger,
		industryMetrics:  make(map[string]*DetailedIndustryMetrics),
		industryRankings: make([]*IndustryRanking, 0),
		industryTrends:   make(map[string]*IndustryTrendAnalysis),
		alertThresholds:  make(map[string]*IndustryAlertThresholds),
		activeAlerts:     make(map[string]*IndustryAlert),
	}
}

// DefaultIndustryAccuracyConfig returns default configuration
func DefaultIndustryAccuracyConfig() *IndustryAccuracyConfig {
	return &IndustryAccuracyConfig{
		EnableDetailedTracking:    true,
		EnableIndustryRankings:    true,
		EnableTrendAnalysis:       true,
		EnablePerformanceTracking: true,

		MinSamplesForAnalysis: 10,
		MinSamplesForRanking:  50,
		MinSamplesForTrends:   20,

		UpdateInterval:        30 * time.Second,
		RankingUpdateInterval: 5 * time.Minute,
		TrendAnalysisInterval: 10 * time.Minute,

		MetricsRetentionPeriod:  7 * 24 * time.Hour,
		MaxHistoricalDataPoints: 1000,

		EnableAlerting:           true,
		AlertCooldownPeriod:      15 * time.Minute,
		DefaultAccuracyThreshold: 0.90,
	}
}

// TrackClassification tracks a classification result for industry analysis
func (iam *IndustryAccuracyMonitor) TrackClassification(ctx context.Context, result *ClassificationResult) error {
	startTime := time.Now()

	iam.mu.Lock()
	defer iam.mu.Unlock()

	// Extract industry information
	industry := iam.extractIndustry(result)
	if industry == "" {
		industry = "unknown"
	}

	// Get or create industry metrics
	metrics, exists := iam.industryMetrics[industry]
	if !exists {
		metrics = iam.createIndustryMetrics(industry)
		iam.industryMetrics[industry] = metrics
	}

	// Update metrics
	iam.updateIndustryMetrics(metrics, result)

	// Update method performance within industry
	iam.updateMethodIndustryMetrics(metrics, result)

	// Update time-based metrics
	iam.updateTimeBasedMetrics(metrics, result)

	// Add to historical data
	iam.addHistoricalData(metrics, result)

	// Update performance indicators
	iam.updatePerformanceIndicators(metrics)

	// Check for alerts
	if iam.config.EnableAlerting {
		iam.checkIndustryAlerts(industry, metrics)
	}

	// Track processing time
	iam.processingTime = time.Since(startTime)
	iam.updateCount++
	iam.lastUpdate = time.Now()

	return nil
}

// GetIndustryMetrics returns detailed metrics for an industry
func (iam *IndustryAccuracyMonitor) GetIndustryMetrics(industry string) *DetailedIndustryMetrics {
	iam.mu.RLock()
	defer iam.mu.RUnlock()

	if metrics, exists := iam.industryMetrics[industry]; exists {
		return iam.copyIndustryMetrics(metrics)
	}
	return nil
}

// GetAllIndustryMetrics returns metrics for all industries
func (iam *IndustryAccuracyMonitor) GetAllIndustryMetrics() map[string]*DetailedIndustryMetrics {
	iam.mu.RLock()
	defer iam.mu.RUnlock()

	result := make(map[string]*DetailedIndustryMetrics)
	for industry, metrics := range iam.industryMetrics {
		result[industry] = iam.copyIndustryMetrics(metrics)
	}
	return result
}

// GetIndustryRankings returns current industry rankings
func (iam *IndustryAccuracyMonitor) GetIndustryRankings() []*IndustryRanking {
	iam.mu.RLock()
	defer iam.mu.RUnlock()

	// Return a copy to prevent race conditions
	result := make([]*IndustryRanking, len(iam.industryRankings))
	for i, ranking := range iam.industryRankings {
		result[i] = &IndustryRanking{
			Rank:                 ranking.Rank,
			IndustryName:         ranking.IndustryName,
			AccuracyScore:        ranking.AccuracyScore,
			TotalClassifications: ranking.TotalClassifications,
			ConfidenceScore:      ranking.ConfidenceScore,
			ReliabilityScore:     ranking.ReliabilityScore,
			PerformanceGrade:     ranking.PerformanceGrade,
			TrendIndicator:       ranking.TrendIndicator,
			LastUpdated:          ranking.LastUpdated,
		}
	}
	return result
}

// GetIndustryTrendAnalysis returns trend analysis for an industry
func (iam *IndustryAccuracyMonitor) GetIndustryTrendAnalysis(industry string) *IndustryTrendAnalysis {
	iam.mu.RLock()
	defer iam.mu.RUnlock()

	if analysis, exists := iam.industryTrends[industry]; exists {
		return iam.copyTrendAnalysis(analysis)
	}
	return nil
}

// GetAllIndustryTrends returns trend analysis for all industries
func (iam *IndustryAccuracyMonitor) GetAllIndustryTrends() map[string]*IndustryTrendAnalysis {
	iam.mu.RLock()
	defer iam.mu.RUnlock()

	result := make(map[string]*IndustryTrendAnalysis)
	for industry, analysis := range iam.industryTrends {
		result[industry] = iam.copyTrendAnalysis(analysis)
	}
	return result
}

// GetTopPerformingIndustries returns the top N performing industries
func (iam *IndustryAccuracyMonitor) GetTopPerformingIndustries(limit int) []*IndustryRanking {
	rankings := iam.GetIndustryRankings()

	if limit <= 0 || limit > len(rankings) {
		limit = len(rankings)
	}

	return rankings[:limit]
}

// GetBottomPerformingIndustries returns the bottom N performing industries
func (iam *IndustryAccuracyMonitor) GetBottomPerformingIndustries(limit int) []*IndustryRanking {
	rankings := iam.GetIndustryRankings()

	if limit <= 0 || limit > len(rankings) {
		limit = len(rankings)
	}

	// Return the last N items (bottom performers)
	start := len(rankings) - limit
	if start < 0 {
		start = 0
	}

	return rankings[start:]
}

// GetIndustryPerformanceSummary returns a performance summary for an industry
func (iam *IndustryAccuracyMonitor) GetIndustryPerformanceSummary(industry string) *IndustryPerformanceSummary {
	metrics := iam.GetIndustryMetrics(industry)
	if metrics == nil {
		return nil
	}

	trends := iam.GetIndustryTrendAnalysis(industry)

	return &IndustryPerformanceSummary{
		IndustryName:          industry,
		AccuracyScore:         metrics.AccuracyScore,
		ConfidenceScore:       metrics.ConfidenceScore,
		ReliabilityScore:      metrics.ReliabilityScore,
		ConsistencyScore:      metrics.ConsistencyScore,
		PerformanceGrade:      metrics.PerformanceGrade,
		TrendIndicator:        metrics.TrendIndicator,
		TotalClassifications:  metrics.TotalClassifications,
		AverageProcessingTime: metrics.AverageProcessingTime,
		DataQualityScore:      metrics.DataQualityScore,
		TrendAnalysis:         trends,
		LastUpdated:           metrics.LastUpdated,
	}
}

// IndustryPerformanceSummary represents a performance summary for an industry
type IndustryPerformanceSummary struct {
	IndustryName          string
	AccuracyScore         float64
	ConfidenceScore       float64
	ReliabilityScore      float64
	ConsistencyScore      float64
	PerformanceGrade      string
	TrendIndicator        string
	TotalClassifications  int64
	AverageProcessingTime time.Duration
	DataQualityScore      float64
	TrendAnalysis         *IndustryTrendAnalysis
	LastUpdated           time.Time
}

// Helper methods

// extractIndustry extracts industry information from classification result
func (iam *IndustryAccuracyMonitor) extractIndustry(result *ClassificationResult) string {
	if result.Metadata != nil {
		if industryValue, exists := result.Metadata["industry"]; exists {
			if industry, ok := industryValue.(string); ok {
				return industry
			}
		}
	}

	// Fallback to actual classification
	return result.ActualClassification
}

// createIndustryMetrics creates new industry metrics
func (iam *IndustryAccuracyMonitor) createIndustryMetrics(industry string) *DetailedIndustryMetrics {
	now := time.Now()

	return &DetailedIndustryMetrics{
		IndustryName:             industry,
		ConfidenceDistribution:   make(map[string]int64),
		MethodPerformance:        make(map[string]*MethodIndustryMetrics),
		ErrorTypeDistribution:    make(map[string]int64),
		BusinessTypeDistribution: make(map[string]int64),
		HourlyAccuracy:           make(map[int]float64),
		DailyAccuracy:            make(map[string]float64),
		WeeklyAccuracy:           make(map[string]float64),
		HistoricalAccuracy:       make([]*AccuracyDataPoint, 0),
		HistoricalConfidence:     make([]*ConfidenceDataPoint, 0),
		HistoricalProcessingTime: make([]*ProcessingTimeDataPoint, 0),
		FirstSeen:                now,
		LastUpdated:              now,
	}
}

// updateIndustryMetrics updates industry metrics with new classification result
func (iam *IndustryAccuracyMonitor) updateIndustryMetrics(metrics *DetailedIndustryMetrics, result *ClassificationResult) {
	metrics.TotalClassifications++
	metrics.LastUpdated = time.Now()

	if result.IsCorrect != nil {
		if *result.IsCorrect {
			metrics.CorrectClassifications++
		} else {
			metrics.IncorrectClassifications++
		}

		// Update accuracy score
		metrics.AccuracyScore = float64(metrics.CorrectClassifications) / float64(metrics.TotalClassifications)
	}

	// Update confidence score
	metrics.ConfidenceScore = (metrics.ConfidenceScore*float64(metrics.TotalClassifications-1) + result.ConfidenceScore) / float64(metrics.TotalClassifications)

	// Update confidence distribution
	confidenceRange := getConfidenceRange(result.ConfidenceScore)
	metrics.ConfidenceDistribution[confidenceRange]++

	// Update error type distribution
	if result.IsCorrect != nil && !*result.IsCorrect {
		errorType := classifyErrorType(result)
		metrics.ErrorTypeDistribution[errorType]++
	}
}

// updateMethodIndustryMetrics updates method performance within an industry
func (iam *IndustryAccuracyMonitor) updateMethodIndustryMetrics(metrics *DetailedIndustryMetrics, result *ClassificationResult) {
	method := result.ClassificationMethod

	methodMetrics, exists := metrics.MethodPerformance[method]
	if !exists {
		methodMetrics = &MethodIndustryMetrics{
			MethodName: method,
		}
		metrics.MethodPerformance[method] = methodMetrics
	}

	methodMetrics.TotalClassifications++
	methodMetrics.LastUpdated = time.Now()

	if result.IsCorrect != nil {
		if *result.IsCorrect {
			methodMetrics.CorrectClassifications++
		}

		// Update accuracy score
		methodMetrics.AccuracyScore = float64(methodMetrics.CorrectClassifications) / float64(methodMetrics.TotalClassifications)

		// Update error rate
		methodMetrics.ErrorRate = float64(methodMetrics.TotalClassifications-methodMetrics.CorrectClassifications) / float64(methodMetrics.TotalClassifications)
	}

	// Update average confidence
	methodMetrics.AverageConfidence = (methodMetrics.AverageConfidence*float64(methodMetrics.TotalClassifications-1) + result.ConfidenceScore) / float64(methodMetrics.TotalClassifications)
}

// updateTimeBasedMetrics updates time-based metrics
func (iam *IndustryAccuracyMonitor) updateTimeBasedMetrics(metrics *DetailedIndustryMetrics, result *ClassificationResult) {
	now := time.Now()
	hour := now.Hour()
	date := now.Format("2006-01-02")
	week := now.Format("2006-W01")

	// Update hourly accuracy
	if result.IsCorrect != nil {
		accuracy := 0.0
		if *result.IsCorrect {
			accuracy = 1.0
		}

		// Simple moving average for hourly accuracy
		currentHourly := metrics.HourlyAccuracy[hour]
		metrics.HourlyAccuracy[hour] = (currentHourly + accuracy) / 2.0

		// Update daily accuracy
		currentDaily := metrics.DailyAccuracy[date]
		metrics.DailyAccuracy[date] = (currentDaily + accuracy) / 2.0

		// Update weekly accuracy
		currentWeekly := metrics.WeeklyAccuracy[week]
		metrics.WeeklyAccuracy[week] = (currentWeekly + accuracy) / 2.0
	}
}

// addHistoricalData adds data to historical tracking
func (iam *IndustryAccuracyMonitor) addHistoricalData(metrics *DetailedIndustryMetrics, result *ClassificationResult) {
	now := time.Now()

	// Add accuracy data point
	if result.IsCorrect != nil {
		accuracy := 0.0
		if *result.IsCorrect {
			accuracy = 1.0
		}

		metrics.HistoricalAccuracy = append(metrics.HistoricalAccuracy, &AccuracyDataPoint{
			Timestamp:  now,
			Accuracy:   accuracy,
			SampleSize: 1,
		})
	}

	// Add confidence data point
	metrics.HistoricalConfidence = append(metrics.HistoricalConfidence, &ConfidenceDataPoint{
		Timestamp:  now,
		Confidence: result.ConfidenceScore,
		SampleSize: 1,
	})

	// Limit historical data size
	if len(metrics.HistoricalAccuracy) > iam.config.MaxHistoricalDataPoints {
		metrics.HistoricalAccuracy = metrics.HistoricalAccuracy[1:]
	}
	if len(metrics.HistoricalConfidence) > iam.config.MaxHistoricalDataPoints {
		metrics.HistoricalConfidence = metrics.HistoricalConfidence[1:]
	}
}

// updatePerformanceIndicators updates performance indicators
func (iam *IndustryAccuracyMonitor) updatePerformanceIndicators(metrics *DetailedIndustryMetrics) {
	// Calculate trend indicator
	if len(metrics.HistoricalAccuracy) >= 10 {
		// Convert to float64 slice for trend calculation
		accuracyValues := make([]float64, len(metrics.HistoricalAccuracy))
		for i, point := range metrics.HistoricalAccuracy {
			accuracyValues[i] = point.Accuracy
		}
		metrics.TrendIndicator = calculateTrendIndicator(accuracyValues)
	}

	// Calculate performance grade
	metrics.PerformanceGrade = iam.calculatePerformanceGrade(metrics)

	// Calculate reliability score
	metrics.ReliabilityScore = iam.calculateReliabilityScore(metrics)

	// Calculate consistency score
	metrics.ConsistencyScore = iam.calculateConsistencyScore(metrics)

	// Calculate data quality score
	metrics.DataQualityScore = iam.calculateDataQualityScore(metrics)
}

// calculatePerformanceGrade calculates a performance grade for an industry
func (iam *IndustryAccuracyMonitor) calculatePerformanceGrade(metrics *DetailedIndustryMetrics) string {
	score := metrics.AccuracyScore

	switch {
	case score >= 0.95:
		return "A+"
	case score >= 0.90:
		return "A"
	case score >= 0.85:
		return "B+"
	case score >= 0.80:
		return "B"
	case score >= 0.75:
		return "C+"
	case score >= 0.70:
		return "C"
	case score >= 0.65:
		return "D+"
	case score >= 0.60:
		return "D"
	default:
		return "F"
	}
}

// calculateReliabilityScore calculates reliability score
func (iam *IndustryAccuracyMonitor) calculateReliabilityScore(metrics *DetailedIndustryMetrics) float64 {
	// Combine accuracy and consistency
	accuracyWeight := 0.7
	consistencyWeight := 0.3

	consistency := iam.calculateConsistencyScore(metrics)

	return (metrics.AccuracyScore * accuracyWeight) + (consistency * consistencyWeight)
}

// calculateConsistencyScore calculates consistency score
func (iam *IndustryAccuracyMonitor) calculateConsistencyScore(metrics *DetailedIndustryMetrics) float64 {
	if len(metrics.HistoricalAccuracy) < 10 {
		return 0.0
	}

	// Calculate standard deviation of recent accuracy
	recent := metrics.HistoricalAccuracy[len(metrics.HistoricalAccuracy)-10:]
	values := make([]float64, len(recent))
	for i, point := range recent {
		values[i] = point.Accuracy
	}

	stdDev := calculateStandardDeviation(values)

	// Lower standard deviation = higher consistency
	return 1.0 - stdDev
}

// calculateDataQualityScore calculates data quality score
func (iam *IndustryAccuracyMonitor) calculateDataQualityScore(metrics *DetailedIndustryMetrics) float64 {
	// Combine various quality indicators
	confidenceWeight := 0.4
	volumeWeight := 0.3
	consistencyWeight := 0.3

	confidenceScore := metrics.ConfidenceScore
	volumeScore := float64(metrics.TotalClassifications) / 1000.0 // Normalize to 0-1
	if volumeScore > 1.0 {
		volumeScore = 1.0
	}
	consistencyScore := iam.calculateConsistencyScore(metrics)

	return (confidenceScore * confidenceWeight) + (volumeScore * volumeWeight) + (consistencyScore * consistencyWeight)
}

// checkIndustryAlerts checks for industry-specific alerts
func (iam *IndustryAccuracyMonitor) checkIndustryAlerts(industry string, metrics *DetailedIndustryMetrics) {
	thresholds, exists := iam.alertThresholds[industry]
	if !exists {
		// Use default thresholds
		thresholds = &IndustryAlertThresholds{
			IndustryName:            industry,
			AccuracyThreshold:       iam.config.DefaultAccuracyThreshold,
			ConfidenceThreshold:     0.80,
			ProcessingTimeThreshold: 2 * time.Second,
			ErrorRateThreshold:      0.10,
			VolumeThreshold:         10,
		}
		iam.alertThresholds[industry] = thresholds
	}

	// Check accuracy threshold
	if metrics.AccuracyScore < thresholds.AccuracyThreshold {
		iam.createIndustryAlert(industry, "accuracy_low", "medium",
			fmt.Sprintf("Industry %s accuracy %.2f%% is below threshold %.2f%%",
				industry, metrics.AccuracyScore*100, thresholds.AccuracyThreshold*100),
			metrics.AccuracyScore, thresholds.AccuracyThreshold)
	}

	// Check confidence threshold
	if metrics.ConfidenceScore < thresholds.ConfidenceThreshold {
		iam.createIndustryAlert(industry, "confidence_low", "low",
			fmt.Sprintf("Industry %s confidence %.2f%% is below threshold %.2f%%",
				industry, metrics.ConfidenceScore*100, thresholds.ConfidenceThreshold*100),
			metrics.ConfidenceScore, thresholds.ConfidenceThreshold)
	}

	// Check error rate threshold
	errorRate := float64(metrics.IncorrectClassifications) / float64(metrics.TotalClassifications)
	if errorRate > thresholds.ErrorRateThreshold {
		iam.createIndustryAlert(industry, "error_rate_high", "high",
			fmt.Sprintf("Industry %s error rate %.2f%% is above threshold %.2f%%",
				industry, errorRate*100, thresholds.ErrorRateThreshold*100),
			errorRate, thresholds.ErrorRateThreshold)
	}
}

// createIndustryAlert creates an industry-specific alert
func (iam *IndustryAccuracyMonitor) createIndustryAlert(industry, alertType, severity, message string, currentValue, thresholdValue float64) {
	alertID := fmt.Sprintf("industry_%s_%s_%d", industry, alertType, time.Now().UnixNano())

	alert := &IndustryAlert{
		ID:             alertID,
		IndustryName:   industry,
		AlertType:      alertType,
		Severity:       severity,
		Message:        message,
		CurrentValue:   currentValue,
		ThresholdValue: thresholdValue,
		Timestamp:      time.Now(),
		Status:         "active",
		Actions:        iam.generateIndustryAlertActions(alertType, severity),
		Metadata:       make(map[string]interface{}),
	}

	iam.activeAlerts[alertID] = alert

	iam.logger.Warn("Industry alert created",
		zap.String("alert_id", alertID),
		zap.String("industry", industry),
		zap.String("type", alertType),
		zap.String("severity", severity),
		zap.String("message", message))
}

// generateIndustryAlertActions generates actions for industry alerts
func (iam *IndustryAccuracyMonitor) generateIndustryAlertActions(alertType, severity string) []string {
	actions := make([]string, 0)

	switch alertType {
	case "accuracy_low":
		actions = append(actions, "investigate_industry_specific_issues", "review_training_data", "check_feature_engineering")
	case "confidence_low":
		actions = append(actions, "analyze_confidence_calibration", "review_model_performance")
	case "error_rate_high":
		actions = append(actions, "investigate_error_patterns", "review_classification_logic", "escalate_to_team")
	}

	// Add severity-specific actions
	switch severity {
	case "high":
		actions = append(actions, "immediate_attention_required", "escalate_to_management")
	case "medium":
		actions = append(actions, "priority_investigation")
	case "low":
		actions = append(actions, "monitor_trend")
	}

	return actions
}

// copyIndustryMetrics creates a copy of industry metrics
func (iam *IndustryAccuracyMonitor) copyIndustryMetrics(metrics *DetailedIndustryMetrics) *DetailedIndustryMetrics {
	// Create a deep copy to prevent race conditions
	copy := &DetailedIndustryMetrics{
		IndustryName:             metrics.IndustryName,
		TotalClassifications:     metrics.TotalClassifications,
		CorrectClassifications:   metrics.CorrectClassifications,
		IncorrectClassifications: metrics.IncorrectClassifications,
		AccuracyScore:            metrics.AccuracyScore,
		ConfidenceScore:          metrics.ConfidenceScore,
		AverageProcessingTime:    metrics.AverageProcessingTime,
		TrendIndicator:           metrics.TrendIndicator,
		PerformanceGrade:         metrics.PerformanceGrade,
		ReliabilityScore:         metrics.ReliabilityScore,
		ConsistencyScore:         metrics.ConsistencyScore,
		DataQualityScore:         metrics.DataQualityScore,
		LastUpdated:              metrics.LastUpdated,
		FirstSeen:                metrics.FirstSeen,

		// Copy maps
		ConfidenceDistribution:   make(map[string]int64),
		MethodPerformance:        make(map[string]*MethodIndustryMetrics),
		ErrorTypeDistribution:    make(map[string]int64),
		BusinessTypeDistribution: make(map[string]int64),
		HourlyAccuracy:           make(map[int]float64),
		DailyAccuracy:            make(map[string]float64),
		WeeklyAccuracy:           make(map[string]float64),

		// Copy historical data
		HistoricalAccuracy:       make([]*AccuracyDataPoint, len(metrics.HistoricalAccuracy)),
		HistoricalConfidence:     make([]*ConfidenceDataPoint, len(metrics.HistoricalConfidence)),
		HistoricalProcessingTime: make([]*ProcessingTimeDataPoint, len(metrics.HistoricalProcessingTime)),
	}

	// Copy map contents
	for k, v := range metrics.ConfidenceDistribution {
		copy.ConfidenceDistribution[k] = v
	}

	for k, v := range metrics.MethodPerformance {
		copy.MethodPerformance[k] = &MethodIndustryMetrics{
			MethodName:             v.MethodName,
			TotalClassifications:   v.TotalClassifications,
			CorrectClassifications: v.CorrectClassifications,
			AccuracyScore:          v.AccuracyScore,
			AverageConfidence:      v.AverageConfidence,
			AverageProcessingTime:  v.AverageProcessingTime,
			ErrorRate:              v.ErrorRate,
			LastUpdated:            v.LastUpdated,
		}
	}

	for k, v := range metrics.ErrorTypeDistribution {
		copy.ErrorTypeDistribution[k] = v
	}

	for k, v := range metrics.BusinessTypeDistribution {
		copy.BusinessTypeDistribution[k] = v
	}

	for k, v := range metrics.HourlyAccuracy {
		copy.HourlyAccuracy[k] = v
	}

	for k, v := range metrics.DailyAccuracy {
		copy.DailyAccuracy[k] = v
	}

	for k, v := range metrics.WeeklyAccuracy {
		copy.WeeklyAccuracy[k] = v
	}

	// Copy historical data
	for i, point := range metrics.HistoricalAccuracy {
		copy.HistoricalAccuracy[i] = &AccuracyDataPoint{
			Timestamp:  point.Timestamp,
			Accuracy:   point.Accuracy,
			SampleSize: point.SampleSize,
		}
	}

	for i, point := range metrics.HistoricalConfidence {
		copy.HistoricalConfidence[i] = &ConfidenceDataPoint{
			Timestamp:  point.Timestamp,
			Confidence: point.Confidence,
			SampleSize: point.SampleSize,
		}
	}

	for i, point := range metrics.HistoricalProcessingTime {
		copy.HistoricalProcessingTime[i] = &ProcessingTimeDataPoint{
			Timestamp:      point.Timestamp,
			ProcessingTime: point.ProcessingTime,
			SampleSize:     point.SampleSize,
		}
	}

	return copy
}

// calculateStandardDeviation calculates the standard deviation of a slice of floats
func calculateStandardDeviation(values []float64) float64 {
	if len(values) < 2 {
		return 0.0
	}

	mean := calculateAverage(values)

	sumSquaredDiffs := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquaredDiffs += diff * diff
	}

	variance := sumSquaredDiffs / float64(len(values)-1)
	return math.Sqrt(variance)
}

// copyTrendAnalysis creates a copy of trend analysis
func (iam *IndustryAccuracyMonitor) copyTrendAnalysis(analysis *IndustryTrendAnalysis) *IndustryTrendAnalysis {
	if analysis == nil {
		return nil
	}

	return &IndustryTrendAnalysis{
		IndustryName:        analysis.IndustryName,
		OverallTrend:        analysis.OverallTrend,
		TrendStrength:       analysis.TrendStrength,
		TrendDirection:      analysis.TrendDirection,
		PredictedAccuracy:   analysis.PredictedAccuracy,
		ConfidenceTrend:     analysis.ConfidenceTrend,
		PerformanceTrend:    analysis.PerformanceTrend,
		LastAnalysis:        analysis.LastAnalysis,
		AnalysisConfidence:  analysis.AnalysisConfidence,
		SeasonalityDetected: analysis.SeasonalityDetected,
		AnomaliesDetected:   analysis.AnomaliesDetected,
	}
}
