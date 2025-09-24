package classification_monitoring

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RealTimeEnsembleMethodTracker provides real-time monitoring of ensemble method performance
type RealTimeEnsembleMethodTracker struct {
	config *EnsembleMethodConfig
	logger *zap.Logger
	mu     sync.RWMutex

	// Core tracking
	methodMetrics  map[string]*RealTimeMethodMetrics
	methodRankings []*MethodRanking
	methodTrends   map[string]*MethodTrendAnalysis

	// Real-time monitoring
	realTimeStats     *EnsembleRealTimeStats
	performanceAlerts map[string]*MethodAlert

	// Performance analysis
	weightOptimizer     *MethodWeightOptimizer
	performanceAnalyzer *MethodPerformanceAnalyzer
}

// EnsembleMethodConfig holds configuration for ensemble method tracking
type EnsembleMethodConfig struct {
	EnableRealTimeTracking   bool `json:"enable_real_time_tracking"`
	EnableMethodRankings     bool `json:"enable_method_rankings"`
	EnableTrendAnalysis      bool `json:"enable_trend_analysis"`
	EnableWeightOptimization bool `json:"enable_weight_optimization"`

	// Thresholds
	MinSamplesForAnalysis int `json:"min_samples_for_analysis"`
	MinSamplesForRanking  int `json:"min_samples_for_ranking"`
	MinSamplesForTrends   int `json:"min_samples_for_trends"`

	// Update intervals
	UpdateInterval             time.Duration `json:"update_interval"`
	RankingUpdateInterval      time.Duration `json:"ranking_update_interval"`
	TrendAnalysisInterval      time.Duration `json:"trend_analysis_interval"`
	WeightOptimizationInterval time.Duration `json:"weight_optimization_interval"`

	// Performance thresholds
	AccuracyThreshold  float64       `json:"accuracy_threshold"`
	LatencyThreshold   time.Duration `json:"latency_threshold"`
	ErrorRateThreshold float64       `json:"error_rate_threshold"`

	// Alerting
	EnableAlerting      bool          `json:"enable_alerting"`
	AlertCooldownPeriod time.Duration `json:"alert_cooldown_period"`
}

// RealTimeMethodMetrics represents real-time metrics for a method
type RealTimeMethodMetrics struct {
	MethodName               string
	TotalClassifications     int64
	CorrectClassifications   int64
	IncorrectClassifications int64
	AccuracyScore            float64
	AverageConfidence        float64
	AverageLatency           time.Duration
	ErrorRate                float64

	// Real-time indicators
	CurrentAccuracy   float64
	CurrentLatency    time.Duration
	CurrentThroughput float64
	CurrentErrorRate  float64

	// Performance indicators
	PerformanceScore float64
	ReliabilityScore float64
	EfficiencyScore  float64
	QualityScore     float64

	// Historical data
	WindowedAccuracy   []float64
	WindowedLatency    []time.Duration
	WindowedConfidence []float64

	// Metadata
	LastUpdated time.Time
	FirstSeen   time.Time
	Status      string // "active", "degraded", "critical"
}

// MethodRanking represents method performance ranking
type MethodRanking struct {
	Rank                 int
	MethodName           string
	AccuracyScore        float64
	PerformanceScore     float64
	ReliabilityScore     float64
	EfficiencyScore      float64
	TotalClassifications int64
	AverageLatency       time.Duration
	Status               string
	TrendIndicator       string
	LastUpdated          time.Time
}

// MethodTrendAnalysis represents trend analysis for a method
type MethodTrendAnalysis struct {
	MethodName           string
	OverallTrend         string
	AccuracyTrend        string
	LatencyTrend         string
	ConfidenceTrend      string
	TrendStrength        float64
	PredictedPerformance float64

	// Detailed trend data
	AccuracyTrendData   []*TrendDataPoint
	LatencyTrendData    []*TrendDataPoint
	ConfidenceTrendData []*TrendDataPoint

	// Analysis metadata
	LastAnalysis       time.Time
	AnalysisConfidence float64
	AnomaliesDetected  []*MethodAnomaly
}

// EnsembleRealTimeStats represents real-time ensemble statistics
type EnsembleRealTimeStats struct {
	TotalMethods      int
	ActiveMethods     int
	DegradedMethods   int
	CriticalMethods   int
	OverallAccuracy   float64
	OverallLatency    time.Duration
	OverallThroughput float64
	OverallErrorRate  float64
	LastUpdated       time.Time
}

// MethodAlert represents an alert for a specific method
type MethodAlert struct {
	ID             string
	MethodName     string
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

// MethodWeightOptimizer optimizes method weights based on performance
type MethodWeightOptimizer struct {
	mu                sync.RWMutex
	currentWeights    map[string]float64
	optimizedWeights  map[string]float64
	lastOptimization  time.Time
	optimizationCount int64
}

// MethodPerformanceAnalyzer analyzes method performance patterns
type MethodPerformanceAnalyzer struct {
	mu                 sync.RWMutex
	performanceHistory map[string][]*PerformanceDataPoint
	analysisResults    map[string]*PerformanceAnalysis
}

// MethodAnomaly represents a detected anomaly in method performance
type MethodAnomaly struct {
	Timestamp     time.Time
	AnomalyType   string
	Severity      string
	Description   string
	Value         float64
	ExpectedValue float64
	Impact        string
}

// PerformanceDataPoint represents a performance data point
type PerformanceDataPoint struct {
	Timestamp  time.Time
	Accuracy   float64
	Latency    time.Duration
	Confidence float64
	Throughput float64
	ErrorRate  float64
}

// PerformanceAnalysis represents performance analysis results
type PerformanceAnalysis struct {
	MethodName        string
	AverageAccuracy   float64
	AverageLatency    time.Duration
	AverageConfidence float64
	AverageThroughput float64
	AverageErrorRate  float64
	PerformanceGrade  string
	ReliabilityGrade  string
	EfficiencyGrade   string
	QualityGrade      string
	LastAnalysis      time.Time
}

// NewRealTimeEnsembleMethodTracker creates a new real-time ensemble method tracker
func NewRealTimeEnsembleMethodTracker(config *EnsembleMethodConfig, logger *zap.Logger) *RealTimeEnsembleMethodTracker {
	if config == nil {
		config = DefaultEnsembleMethodConfig()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &RealTimeEnsembleMethodTracker{
		config:              config,
		logger:              logger,
		methodMetrics:       make(map[string]*RealTimeMethodMetrics),
		methodRankings:      make([]*MethodRanking, 0),
		methodTrends:        make(map[string]*MethodTrendAnalysis),
		realTimeStats:       &EnsembleRealTimeStats{},
		performanceAlerts:   make(map[string]*MethodAlert),
		weightOptimizer:     NewMethodWeightOptimizer(),
		performanceAnalyzer: NewMethodPerformanceAnalyzer(),
	}
}

// DefaultEnsembleMethodConfig returns default configuration
func DefaultEnsembleMethodConfig() *EnsembleMethodConfig {
	return &EnsembleMethodConfig{
		EnableRealTimeTracking:   true,
		EnableMethodRankings:     true,
		EnableTrendAnalysis:      true,
		EnableWeightOptimization: true,

		MinSamplesForAnalysis: 10,
		MinSamplesForRanking:  50,
		MinSamplesForTrends:   20,

		UpdateInterval:             30 * time.Second,
		RankingUpdateInterval:      5 * time.Minute,
		TrendAnalysisInterval:      10 * time.Minute,
		WeightOptimizationInterval: 1 * time.Hour,

		AccuracyThreshold:  0.90,
		LatencyThreshold:   2 * time.Second,
		ErrorRateThreshold: 0.10,

		EnableAlerting:      true,
		AlertCooldownPeriod: 15 * time.Minute,
	}
}

// TrackMethodResult tracks a method classification result
func (emt *RealTimeEnsembleMethodTracker) TrackMethodResult(ctx context.Context, result *ClassificationResult) error {
	emt.mu.Lock()
	defer emt.mu.Unlock()

	method := result.ClassificationMethod

	// Get or create method metrics
	metrics, exists := emt.methodMetrics[method]
	if !exists {
		metrics = emt.createMethodMetrics(method)
		emt.methodMetrics[method] = metrics
	}

	// Update metrics
	emt.updateMethodMetrics(metrics, result)

	// Update real-time indicators
	emt.updateRealTimeIndicators(metrics, result)

	// Update real-time stats
	emt.updateRealTimeStats()

	// Add to performance analyzer
	emt.performanceAnalyzer.AddDataPoint(method, &PerformanceDataPoint{
		Timestamp:  result.Timestamp,
		Accuracy:   emt.calculateAccuracy(result),
		Latency:    time.Since(result.Timestamp), // Approximate
		Confidence: result.ConfidenceScore,
		Throughput: 1.0, // Will be calculated properly
		ErrorRate:  emt.calculateErrorRate(result),
	})

	// Check for alerts
	if emt.config.EnableAlerting {
		emt.checkMethodAlerts(method, metrics)
	}

	return nil
}

// GetMethodMetrics returns metrics for a specific method
func (emt *RealTimeEnsembleMethodTracker) GetMethodMetrics(method string) *RealTimeMethodMetrics {
	emt.mu.RLock()
	defer emt.mu.RUnlock()

	if metrics, exists := emt.methodMetrics[method]; exists {
		return emt.copyMethodMetrics(metrics)
	}
	return nil
}

// GetAllMethodMetrics returns metrics for all methods
func (emt *RealTimeEnsembleMethodTracker) GetAllMethodMetrics() map[string]*RealTimeMethodMetrics {
	emt.mu.RLock()
	defer emt.mu.RUnlock()

	result := make(map[string]*RealTimeMethodMetrics)
	for method, metrics := range emt.methodMetrics {
		result[method] = emt.copyMethodMetrics(metrics)
	}
	return result
}

// GetMethodRankings returns current method rankings
func (emt *RealTimeEnsembleMethodTracker) GetMethodRankings() []*MethodRanking {
	emt.mu.RLock()
	defer emt.mu.RUnlock()

	// Return a copy to prevent race conditions
	result := make([]*MethodRanking, len(emt.methodRankings))
	for i, ranking := range emt.methodRankings {
		result[i] = &MethodRanking{
			Rank:                 ranking.Rank,
			MethodName:           ranking.MethodName,
			AccuracyScore:        ranking.AccuracyScore,
			PerformanceScore:     ranking.PerformanceScore,
			ReliabilityScore:     ranking.ReliabilityScore,
			EfficiencyScore:      ranking.EfficiencyScore,
			TotalClassifications: ranking.TotalClassifications,
			AverageLatency:       ranking.AverageLatency,
			Status:               ranking.Status,
			TrendIndicator:       ranking.TrendIndicator,
			LastUpdated:          ranking.LastUpdated,
		}
	}
	return result
}

// GetRealTimeStats returns current real-time ensemble statistics
func (emt *RealTimeEnsembleMethodTracker) GetRealTimeStats() *EnsembleRealTimeStats {
	emt.mu.RLock()
	defer emt.mu.RUnlock()

	return &EnsembleRealTimeStats{
		TotalMethods:      emt.realTimeStats.TotalMethods,
		ActiveMethods:     emt.realTimeStats.ActiveMethods,
		DegradedMethods:   emt.realTimeStats.DegradedMethods,
		CriticalMethods:   emt.realTimeStats.CriticalMethods,
		OverallAccuracy:   emt.realTimeStats.OverallAccuracy,
		OverallLatency:    emt.realTimeStats.OverallLatency,
		OverallThroughput: emt.realTimeStats.OverallThroughput,
		OverallErrorRate:  emt.realTimeStats.OverallErrorRate,
		LastUpdated:       emt.realTimeStats.LastUpdated,
	}
}

// GetOptimizedWeights returns optimized method weights
func (emt *RealTimeEnsembleMethodTracker) GetOptimizedWeights() map[string]float64 {
	return emt.weightOptimizer.GetOptimizedWeights()
}

// GetPerformanceAnalysis returns performance analysis for a method
func (emt *RealTimeEnsembleMethodTracker) GetPerformanceAnalysis(method string) *PerformanceAnalysis {
	return emt.performanceAnalyzer.GetAnalysis(method)
}

// Helper methods

// createMethodMetrics creates new method metrics
func (emt *RealTimeEnsembleMethodTracker) createMethodMetrics(method string) *RealTimeMethodMetrics {
	now := time.Now()

	return &RealTimeMethodMetrics{
		MethodName:         method,
		WindowedAccuracy:   make([]float64, 0),
		WindowedLatency:    make([]time.Duration, 0),
		WindowedConfidence: make([]float64, 0),
		FirstSeen:          now,
		LastUpdated:        now,
		Status:             "active",
	}
}

// updateMethodMetrics updates method metrics with new result
func (emt *RealTimeEnsembleMethodTracker) updateMethodMetrics(metrics *RealTimeMethodMetrics, result *ClassificationResult) {
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
		metrics.ErrorRate = float64(metrics.IncorrectClassifications) / float64(metrics.TotalClassifications)
	}

	// Update average confidence
	metrics.AverageConfidence = (metrics.AverageConfidence*float64(metrics.TotalClassifications-1) + result.ConfidenceScore) / float64(metrics.TotalClassifications)

	// Update windowed data
	accuracy := emt.calculateAccuracy(result)
	metrics.WindowedAccuracy = append(metrics.WindowedAccuracy, accuracy)
	if len(metrics.WindowedAccuracy) > 100 {
		metrics.WindowedAccuracy = metrics.WindowedAccuracy[1:]
	}

	metrics.WindowedConfidence = append(metrics.WindowedConfidence, result.ConfidenceScore)
	if len(metrics.WindowedConfidence) > 100 {
		metrics.WindowedConfidence = metrics.WindowedConfidence[1:]
	}
}

// updateRealTimeStats updates the overall real-time statistics
func (emt *RealTimeEnsembleMethodTracker) updateRealTimeStats() {
	totalMethods := len(emt.methodMetrics)
	activeMethods := 0
	degradedMethods := 0
	criticalMethods := 0

	var totalAccuracy, totalLatency, totalThroughput, totalErrorRate float64

	for _, metrics := range emt.methodMetrics {
		switch metrics.Status {
		case "active":
			activeMethods++
		case "degraded":
			degradedMethods++
		case "critical":
			criticalMethods++
		}

		totalAccuracy += metrics.AccuracyScore
		totalLatency += float64(metrics.AverageLatency.Milliseconds())
		totalThroughput += metrics.CurrentThroughput
		totalErrorRate += metrics.ErrorRate
	}

	if totalMethods > 0 {
		emt.realTimeStats = &EnsembleRealTimeStats{
			TotalMethods:      totalMethods,
			ActiveMethods:     activeMethods,
			DegradedMethods:   degradedMethods,
			CriticalMethods:   criticalMethods,
			OverallAccuracy:   totalAccuracy / float64(totalMethods),
			OverallLatency:    time.Duration(totalLatency/float64(totalMethods)) * time.Millisecond,
			OverallThroughput: totalThroughput / float64(totalMethods),
			OverallErrorRate:  totalErrorRate / float64(totalMethods),
			LastUpdated:       time.Now(),
		}
	} else {
		emt.realTimeStats = &EnsembleRealTimeStats{
			TotalMethods:      0,
			ActiveMethods:     0,
			DegradedMethods:   0,
			CriticalMethods:   0,
			OverallAccuracy:   0.0,
			OverallLatency:    0,
			OverallThroughput: 0.0,
			OverallErrorRate:  0.0,
			LastUpdated:       time.Now(),
		}
	}
}

// updateRealTimeIndicators updates real-time indicators
func (emt *RealTimeEnsembleMethodTracker) updateRealTimeIndicators(metrics *RealTimeMethodMetrics, result *ClassificationResult) {
	// Update current accuracy (from recent window)
	if len(metrics.WindowedAccuracy) > 0 {
		windowSize := 10
		start := 0
		if len(metrics.WindowedAccuracy) > windowSize {
			start = len(metrics.WindowedAccuracy) - windowSize
		}
		recent := metrics.WindowedAccuracy[start:]
		if len(recent) > 0 {
			sum := 0.0
			for _, acc := range recent {
				sum += acc
			}
			metrics.CurrentAccuracy = sum / float64(len(recent))
		}
	}

	// Update current latency (approximate)
	metrics.CurrentLatency = time.Since(result.Timestamp)
	metrics.AverageLatency = (metrics.AverageLatency*time.Duration(metrics.TotalClassifications-1) + metrics.CurrentLatency) / time.Duration(metrics.TotalClassifications)

	// Update current throughput (requests per second)
	metrics.CurrentThroughput = float64(metrics.TotalClassifications) / time.Since(metrics.FirstSeen).Seconds()

	// Update current error rate
	metrics.CurrentErrorRate = metrics.ErrorRate

	// Update performance indicators
	emt.updatePerformanceIndicators(metrics)

	// Update method status
	emt.updateMethodStatus(metrics)
}

// updatePerformanceIndicators updates performance indicators
func (emt *RealTimeEnsembleMethodTracker) updatePerformanceIndicators(metrics *RealTimeMethodMetrics) {
	// Calculate performance score (combination of accuracy and efficiency)
	accuracyWeight := 0.6
	efficiencyWeight := 0.4

	// Efficiency based on latency (lower is better)
	efficiency := 1.0
	if metrics.AverageLatency > 0 {
		efficiency = 1.0 / (1.0 + float64(metrics.AverageLatency.Nanoseconds())/1e9)
	}

	metrics.PerformanceScore = (metrics.AccuracyScore * accuracyWeight) + (efficiency * efficiencyWeight)

	// Calculate reliability score (consistency of accuracy)
	metrics.ReliabilityScore = emt.calculateReliabilityScore(metrics)

	// Calculate efficiency score (throughput vs latency)
	metrics.EfficiencyScore = emt.calculateEfficiencyScore(metrics)

	// Calculate quality score (confidence and accuracy)
	metrics.QualityScore = (metrics.AccuracyScore * 0.7) + (metrics.AverageConfidence * 0.3)
}

// updateMethodStatus updates method status based on performance
func (emt *RealTimeEnsembleMethodTracker) updateMethodStatus(metrics *RealTimeMethodMetrics) {
	if metrics.AccuracyScore < emt.config.AccuracyThreshold*0.8 ||
		metrics.ErrorRate > emt.config.ErrorRateThreshold*1.5 ||
		metrics.AverageLatency > emt.config.LatencyThreshold*2 {
		metrics.Status = "critical"
	} else if metrics.AccuracyScore < emt.config.AccuracyThreshold ||
		metrics.ErrorRate > emt.config.ErrorRateThreshold ||
		metrics.AverageLatency > emt.config.LatencyThreshold {
		metrics.Status = "degraded"
	} else {
		metrics.Status = "active"
	}
}

// calculateAccuracy calculates accuracy from classification result
func (emt *RealTimeEnsembleMethodTracker) calculateAccuracy(result *ClassificationResult) float64 {
	if result.IsCorrect == nil {
		return 0.0
	}
	if *result.IsCorrect {
		return 1.0
	}
	return 0.0
}

// calculateErrorRate calculates error rate from classification result
func (emt *RealTimeEnsembleMethodTracker) calculateErrorRate(result *ClassificationResult) float64 {
	if result.IsCorrect == nil {
		return 0.0
	}
	if *result.IsCorrect {
		return 0.0
	}
	return 1.0
}

// calculateReliabilityScore calculates reliability score
func (emt *RealTimeEnsembleMethodTracker) calculateReliabilityScore(metrics *RealTimeMethodMetrics) float64 {
	if len(metrics.WindowedAccuracy) < 10 {
		return 0.0
	}

	// Calculate standard deviation of recent accuracy
	recent := metrics.WindowedAccuracy[len(metrics.WindowedAccuracy)-10:]
	stdDev := calculateStandardDeviation(recent)

	// Lower standard deviation = higher reliability
	return 1.0 - stdDev
}

// calculateEfficiencyScore calculates efficiency score
func (emt *RealTimeEnsembleMethodTracker) calculateEfficiencyScore(metrics *RealTimeMethodMetrics) float64 {
	// Combine throughput and latency
	throughputScore := math.Min(metrics.CurrentThroughput/10.0, 1.0) // Normalize to 0-1
	latencyScore := 1.0 / (1.0 + float64(metrics.AverageLatency.Nanoseconds())/1e9)

	return (throughputScore * 0.6) + (latencyScore * 0.4)
}

// checkMethodAlerts checks for method-specific alerts
func (emt *RealTimeEnsembleMethodTracker) checkMethodAlerts(method string, metrics *RealTimeMethodMetrics) {
	// Check accuracy threshold
	if metrics.AccuracyScore < emt.config.AccuracyThreshold {
		emt.createMethodAlert(method, "accuracy_low", "medium",
			fmt.Sprintf("Method %s accuracy %.2f%% is below threshold %.2f%%",
				method, metrics.AccuracyScore*100, emt.config.AccuracyThreshold*100),
			metrics.AccuracyScore, emt.config.AccuracyThreshold)
	}

	// Check latency threshold
	if metrics.AverageLatency > emt.config.LatencyThreshold {
		emt.createMethodAlert(method, "latency_high", "medium",
			fmt.Sprintf("Method %s latency %v exceeds threshold %v",
				method, metrics.AverageLatency, emt.config.LatencyThreshold),
			float64(metrics.AverageLatency.Nanoseconds()), float64(emt.config.LatencyThreshold.Nanoseconds()))
	}

	// Check error rate threshold
	if metrics.ErrorRate > emt.config.ErrorRateThreshold {
		emt.createMethodAlert(method, "error_rate_high", "high",
			fmt.Sprintf("Method %s error rate %.2f%% exceeds threshold %.2f%%",
				method, metrics.ErrorRate*100, emt.config.ErrorRateThreshold*100),
			metrics.ErrorRate, emt.config.ErrorRateThreshold)
	}
}

// createMethodAlert creates a method-specific alert
func (emt *RealTimeEnsembleMethodTracker) createMethodAlert(method, alertType, severity, message string, currentValue, thresholdValue float64) {
	alertID := fmt.Sprintf("method_%s_%s_%d", method, alertType, time.Now().UnixNano())

	alert := &MethodAlert{
		ID:             alertID,
		MethodName:     method,
		AlertType:      alertType,
		Severity:       severity,
		Message:        message,
		CurrentValue:   currentValue,
		ThresholdValue: thresholdValue,
		Timestamp:      time.Now(),
		Status:         "active",
		Actions:        emt.generateMethodAlertActions(alertType, severity),
		Metadata:       make(map[string]interface{}),
	}

	emt.performanceAlerts[alertID] = alert

	emt.logger.Warn("Method alert created",
		zap.String("alert_id", alertID),
		zap.String("method", method),
		zap.String("type", alertType),
		zap.String("severity", severity),
		zap.String("message", message))
}

// generateMethodAlertActions generates actions for method alerts
func (emt *RealTimeEnsembleMethodTracker) generateMethodAlertActions(alertType, severity string) []string {
	actions := make([]string, 0)

	switch alertType {
	case "accuracy_low":
		actions = append(actions, "investigate_method_performance", "review_training_data", "check_feature_engineering")
	case "latency_high":
		actions = append(actions, "analyze_performance_bottlenecks", "optimize_processing", "check_system_resources")
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

// copyMethodMetrics creates a copy of method metrics
func (emt *RealTimeEnsembleMethodTracker) copyMethodMetrics(metrics *RealTimeMethodMetrics) *RealTimeMethodMetrics {
	return &RealTimeMethodMetrics{
		MethodName:               metrics.MethodName,
		TotalClassifications:     metrics.TotalClassifications,
		CorrectClassifications:   metrics.CorrectClassifications,
		IncorrectClassifications: metrics.IncorrectClassifications,
		AccuracyScore:            metrics.AccuracyScore,
		AverageConfidence:        metrics.AverageConfidence,
		AverageLatency:           metrics.AverageLatency,
		ErrorRate:                metrics.ErrorRate,
		CurrentAccuracy:          metrics.CurrentAccuracy,
		CurrentLatency:           metrics.CurrentLatency,
		CurrentThroughput:        metrics.CurrentThroughput,
		CurrentErrorRate:         metrics.CurrentErrorRate,
		PerformanceScore:         metrics.PerformanceScore,
		ReliabilityScore:         metrics.ReliabilityScore,
		EfficiencyScore:          metrics.EfficiencyScore,
		QualityScore:             metrics.QualityScore,
		LastUpdated:              metrics.LastUpdated,
		FirstSeen:                metrics.FirstSeen,
		Status:                   metrics.Status,
	}
}

// NewMethodWeightOptimizer creates a new method weight optimizer
func NewMethodWeightOptimizer() *MethodWeightOptimizer {
	return &MethodWeightOptimizer{
		currentWeights:   make(map[string]float64),
		optimizedWeights: make(map[string]float64),
	}
}

// GetOptimizedWeights returns optimized method weights
func (mwo *MethodWeightOptimizer) GetOptimizedWeights() map[string]float64 {
	mwo.mu.RLock()
	defer mwo.mu.RUnlock()

	result := make(map[string]float64)
	for method, weight := range mwo.optimizedWeights {
		result[method] = weight
	}
	return result
}

// NewMethodPerformanceAnalyzer creates a new method performance analyzer
func NewMethodPerformanceAnalyzer() *MethodPerformanceAnalyzer {
	return &MethodPerformanceAnalyzer{
		performanceHistory: make(map[string][]*PerformanceDataPoint),
		analysisResults:    make(map[string]*PerformanceAnalysis),
	}
}

// AddDataPoint adds a performance data point
func (mpa *MethodPerformanceAnalyzer) AddDataPoint(method string, point *PerformanceDataPoint) {
	mpa.mu.Lock()
	defer mpa.mu.Unlock()

	mpa.performanceHistory[method] = append(mpa.performanceHistory[method], point)

	// Limit history size
	if len(mpa.performanceHistory[method]) > 1000 {
		mpa.performanceHistory[method] = mpa.performanceHistory[method][500:]
	}
}

// GetAnalysis returns performance analysis for a method
func (mpa *MethodPerformanceAnalyzer) GetAnalysis(method string) *PerformanceAnalysis {
	mpa.mu.RLock()
	defer mpa.mu.RUnlock()

	if analysis, exists := mpa.analysisResults[method]; exists {
		return analysis
	}
	return nil
}
