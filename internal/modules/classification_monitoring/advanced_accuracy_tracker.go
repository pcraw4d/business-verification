package classification_monitoring

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AdvancedAccuracyTracker provides enhanced real-time accuracy tracking with 95%+ target monitoring
type AdvancedAccuracyTracker struct {
	config    *AdvancedAccuracyConfig
	logger    *zap.Logger
	mu        sync.RWMutex
	startTime time.Time

	// Core tracking components
	overallTracker  *OverallAccuracyTracker
	industryTracker *IndustryAccuracyTracker
	ensembleTracker *EnsembleMethodTracker
	mlModelTracker  *MLModelAccuracyTracker
	securityTracker *SecurityAccuracyTracker

	// Real-time monitoring
	realTimeMetrics    *RealTimeMetrics
	alertManager       *AdvancedAlertManager
	performanceMonitor *PerformanceMonitor

	// Historical data
	historicalData []*HistoricalAccuracySnapshot
	trendAnalyzer  *TrendAnalyzer

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	active bool
}

// AdvancedAccuracyConfig holds configuration for advanced accuracy tracking
type AdvancedAccuracyConfig struct {
	// Core settings
	EnableRealTimeTracking    bool    `json:"enable_real_time_tracking"`
	TargetAccuracy            float64 `json:"target_accuracy"`             // 0.95 (95%)
	CriticalAccuracyThreshold float64 `json:"critical_accuracy_threshold"` // 0.90 (90%)
	WarningAccuracyThreshold  float64 `json:"warning_accuracy_threshold"`  // 0.92 (92%)

	// Monitoring intervals
	CollectionInterval    time.Duration `json:"collection_interval"`
	AlertCheckInterval    time.Duration `json:"alert_check_interval"`
	TrendAnalysisInterval time.Duration `json:"trend_analysis_interval"`

	// Data retention
	MetricsRetentionPeriod  time.Duration `json:"metrics_retention_period"`
	HistoricalDataRetention time.Duration `json:"historical_data_retention"`
	MaxHistoricalSnapshots  int           `json:"max_historical_snapshots"`

	// Analysis settings
	SampleWindowSize      int `json:"sample_window_size"`
	TrendWindowSize       int `json:"trend_window_size"`
	MinSamplesForAnalysis int `json:"min_samples_for_analysis"`

	// Security monitoring
	EnableSecurityTracking bool    `json:"enable_security_tracking"`
	SecurityTrustTarget    float64 `json:"security_trust_target"` // 1.0 (100%)

	// Performance monitoring
	EnablePerformanceTracking bool          `json:"enable_performance_tracking"`
	MaxProcessingTime         time.Duration `json:"max_processing_time"`

	// Alerting
	EnableAlerting      bool          `json:"enable_alerting"`
	AlertCooldownPeriod time.Duration `json:"alert_cooldown_period"`
	MaxAlertsPerHour    int           `json:"max_alerts_per_hour"`
}

// OverallAccuracyTracker tracks overall system accuracy
type OverallAccuracyTracker struct {
	mu                     sync.RWMutex
	totalClassifications   int64
	correctClassifications int64
	accuracyScore          float64
	windowedAccuracy       []float64
	confidenceDistribution map[string]int64
	lastUpdated            time.Time
	trendIndicator         string
}

// IndustryAccuracyTracker tracks accuracy by industry
type IndustryAccuracyTracker struct {
	mu         sync.RWMutex
	industries map[string]*IndustryAccuracyMetrics
}

// IndustryAccuracyMetrics represents accuracy metrics for a specific industry
type IndustryAccuracyMetrics struct {
	IndustryName           string
	TotalClassifications   int64
	CorrectClassifications int64
	AccuracyScore          float64
	WindowedAccuracy       []float64
	ConfidenceDistribution map[string]int64
	LastUpdated            time.Time
	TrendIndicator         string
	TopMisclassifications  []*MisclassificationRecord
}

// EnsembleMethodTracker tracks performance of ensemble methods
type EnsembleMethodTracker struct {
	mu      sync.RWMutex
	methods map[string]*MethodAccuracyMetrics
}

// MethodAccuracyMetrics represents accuracy metrics for a specific method
type MethodAccuracyMetrics struct {
	MethodName             string
	TotalClassifications   int64
	CorrectClassifications int64
	AccuracyScore          float64
	AverageConfidence      float64
	AverageLatency         time.Duration
	WindowedAccuracy       []float64
	LastUpdated            time.Time
	TrendIndicator         string
	Weight                 float64
	PerformanceScore       float64
}

// MLModelAccuracyTracker tracks ML model accuracy trends
type MLModelAccuracyTracker struct {
	mu     sync.RWMutex
	models map[string]*MLModelMetrics
}

// MLModelMetrics represents metrics for a specific ML model
type MLModelMetrics struct {
	ModelName          string
	ModelVersion       string
	TotalPredictions   int64
	CorrectPredictions int64
	AccuracyScore      float64
	AverageConfidence  float64
	ModelDriftScore    float64
	LastRetrained      time.Time
	WindowedAccuracy   []float64
	LastUpdated        time.Time
	TrendIndicator     string
	UncertaintyScore   float64
}

// SecurityAccuracyTracker tracks security-related accuracy metrics
type SecurityAccuracyTracker struct {
	mu                      sync.RWMutex
	trustedDataSourceRate   float64
	websiteVerificationRate float64
	securityViolationRate   float64
	confidenceIntegrity     float64
	lastUpdated             time.Time
}

// RealTimeMetrics holds real-time monitoring data
type RealTimeMetrics struct {
	mu                sync.RWMutex
	CurrentAccuracy   float64
	CurrentThroughput float64
	CurrentLatency    time.Duration
	ActiveAlerts      int
	LastUpdate        time.Time
	HealthStatus      string
}

// HistoricalAccuracySnapshot represents a point-in-time accuracy snapshot
type HistoricalAccuracySnapshot struct {
	Timestamp          time.Time
	OverallAccuracy    float64
	IndustryAccuracies map[string]float64
	MethodAccuracies   map[string]float64
	MLModelAccuracies  map[string]float64
	SecurityMetrics    *SecurityAccuracySnapshot
	PerformanceMetrics *PerformanceSnapshot
}

// SecurityAccuracySnapshot represents security metrics at a point in time
type SecurityAccuracySnapshot struct {
	TrustedDataSourceRate   float64
	WebsiteVerificationRate float64
	SecurityViolationRate   float64
	ConfidenceIntegrity     float64
}

// PerformanceSnapshot represents performance metrics at a point in time
type PerformanceSnapshot struct {
	AverageLatency time.Duration
	Throughput     float64
	ErrorRate      float64
	CPUUsage       float64
	MemoryUsage    float64
}

// TrendAnalyzer analyzes accuracy trends
type TrendAnalyzer struct {
	mu     sync.RWMutex
	trends map[string]*TrendData
}

// TrendData represents trend analysis data
type TrendData struct {
	DimensionName     string
	DimensionValue    string
	Trend             string // "improving", "declining", "stable"
	TrendStrength     float64
	LastAnalysis      time.Time
	PredictedAccuracy float64
}

// NewAdvancedAccuracyTracker creates a new advanced accuracy tracker
func NewAdvancedAccuracyTracker(config *AdvancedAccuracyConfig, logger *zap.Logger) *AdvancedAccuracyTracker {
	if config == nil {
		config = DefaultAdvancedAccuracyConfig()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	ctx, cancel := context.WithCancel(context.Background())

	tracker := &AdvancedAccuracyTracker{
		config:    config,
		logger:    logger,
		startTime: time.Now(),
		ctx:       ctx,
		cancel:    cancel,

		// Initialize tracking components
		overallTracker:  NewOverallAccuracyTracker(),
		industryTracker: NewIndustryAccuracyTracker(),
		ensembleTracker: NewEnsembleMethodTracker(),
		mlModelTracker:  NewMLModelAccuracyTracker(),
		securityTracker: NewSecurityAccuracyTracker(),

		// Initialize monitoring components
		realTimeMetrics:    NewRealTimeMetrics(),
		alertManager:       NewAdvancedAlertManager(config, logger),
		performanceMonitor: NewPerformanceMonitor(config, logger),

		// Initialize analysis components
		historicalData: make([]*HistoricalAccuracySnapshot, 0),
		trendAnalyzer:  NewTrendAnalyzer(),
	}

	return tracker
}

// DefaultAdvancedAccuracyConfig returns default configuration for advanced accuracy tracking
func DefaultAdvancedAccuracyConfig() *AdvancedAccuracyConfig {
	return &AdvancedAccuracyConfig{
		EnableRealTimeTracking:    true,
		TargetAccuracy:            0.95, // 95% target
		CriticalAccuracyThreshold: 0.90, // 90% critical
		WarningAccuracyThreshold:  0.92, // 92% warning

		CollectionInterval:    30 * time.Second,
		AlertCheckInterval:    1 * time.Minute,
		TrendAnalysisInterval: 5 * time.Minute,

		MetricsRetentionPeriod:  7 * 24 * time.Hour,  // 7 days
		HistoricalDataRetention: 30 * 24 * time.Hour, // 30 days
		MaxHistoricalSnapshots:  1000,

		SampleWindowSize:      100,
		TrendWindowSize:       20,
		MinSamplesForAnalysis: 10,

		EnableSecurityTracking: true,
		SecurityTrustTarget:    1.0, // 100% target

		EnablePerformanceTracking: true,
		MaxProcessingTime:         500 * time.Millisecond,

		EnableAlerting:      true,
		AlertCooldownPeriod: 15 * time.Minute,
		MaxAlertsPerHour:    10,
	}
}

// Start starts the advanced accuracy tracking
func (aat *AdvancedAccuracyTracker) Start() error {
	aat.mu.Lock()
	defer aat.mu.Unlock()

	if aat.active {
		return fmt.Errorf("advanced accuracy tracker is already active")
	}

	aat.active = true

	// Start background monitoring loops
	go aat.monitoringLoop()
	go aat.alertingLoop()
	go aat.trendAnalysisLoop()
	go aat.dataCleanupLoop()

	aat.logger.Info("Advanced accuracy tracking started",
		zap.Float64("target_accuracy", aat.config.TargetAccuracy),
		zap.Float64("critical_threshold", aat.config.CriticalAccuracyThreshold),
		zap.Duration("collection_interval", aat.config.CollectionInterval),
		zap.Bool("security_tracking", aat.config.EnableSecurityTracking),
		zap.Bool("performance_tracking", aat.config.EnablePerformanceTracking))

	return nil
}

// Stop stops the advanced accuracy tracking
func (aat *AdvancedAccuracyTracker) Stop() {
	aat.mu.Lock()
	defer aat.mu.Unlock()

	if !aat.active {
		return
	}

	aat.active = false
	aat.cancel()

	aat.logger.Info("Advanced accuracy tracking stopped")
}

// TrackClassification tracks a new classification result with enhanced metrics
func (aat *AdvancedAccuracyTracker) TrackClassification(ctx context.Context, result *ClassificationResult) error {
	if !aat.active {
		return fmt.Errorf("advanced accuracy tracker is not active")
	}

	startTime := time.Now()

	// Track in all components
	if err := aat.overallTracker.TrackResult(result); err != nil {
		aat.logger.Error("Failed to track overall result", zap.Error(err))
	}

	if err := aat.industryTracker.TrackResult(result); err != nil {
		aat.logger.Error("Failed to track industry result", zap.Error(err))
	}

	if err := aat.ensembleTracker.TrackResult(result); err != nil {
		aat.logger.Error("Failed to track ensemble result", zap.Error(err))
	}

	if err := aat.mlModelTracker.TrackResult(result); err != nil {
		aat.logger.Error("Failed to track ML model result", zap.Error(err))
	}

	if aat.config.EnableSecurityTracking {
		if err := aat.securityTracker.TrackResult(result); err != nil {
			aat.logger.Error("Failed to track security result", zap.Error(err))
		}
	}

	// Update real-time metrics
	aat.updateRealTimeMetrics()

	// Track performance
	if aat.config.EnablePerformanceTracking {
		processingTime := time.Since(startTime)
		aat.performanceMonitor.RecordProcessingTime(processingTime)
	}

	aat.logger.Debug("Classification tracked",
		zap.String("business_name", result.BusinessName),
		zap.String("method", result.ClassificationMethod),
		zap.Float64("confidence", result.ConfidenceScore),
		zap.Bool("is_correct", result.IsCorrect != nil && *result.IsCorrect))

	return nil
}

// GetOverallAccuracy returns the current overall accuracy
func (aat *AdvancedAccuracyTracker) GetOverallAccuracy() float64 {
	return aat.overallTracker.GetAccuracy()
}

// GetIndustryAccuracy returns accuracy for a specific industry
func (aat *AdvancedAccuracyTracker) GetIndustryAccuracy(industry string) float64 {
	return aat.industryTracker.GetIndustryAccuracy(industry)
}

// GetMethodAccuracy returns accuracy for a specific method
func (aat *AdvancedAccuracyTracker) GetMethodAccuracy(method string) float64 {
	return aat.ensembleTracker.GetMethodAccuracy(method)
}

// GetMLModelAccuracy returns accuracy for a specific ML model
func (aat *AdvancedAccuracyTracker) GetMLModelAccuracy(model string) float64 {
	return aat.mlModelTracker.GetModelAccuracy(model)
}

// GetSecurityMetrics returns current security metrics
func (aat *AdvancedAccuracyTracker) GetSecurityMetrics() *SecurityAccuracySnapshot {
	return aat.securityTracker.GetMetrics()
}

// GetRealTimeMetrics returns current real-time metrics
func (aat *AdvancedAccuracyTracker) GetRealTimeMetrics() *RealTimeMetrics {
	aat.mu.RLock()
	defer aat.mu.RUnlock()

	return &RealTimeMetrics{
		CurrentAccuracy:   aat.realTimeMetrics.CurrentAccuracy,
		CurrentThroughput: aat.realTimeMetrics.CurrentThroughput,
		CurrentLatency:    aat.realTimeMetrics.CurrentLatency,
		ActiveAlerts:      aat.realTimeMetrics.ActiveAlerts,
		LastUpdate:        aat.realTimeMetrics.LastUpdate,
		HealthStatus:      aat.realTimeMetrics.HealthStatus,
	}
}

// GetTrendAnalysis returns trend analysis for all dimensions
func (aat *AdvancedAccuracyTracker) GetTrendAnalysis() map[string]*TrendData {
	return aat.trendAnalyzer.GetAllTrends()
}

// IsTargetAccuracyMet checks if the 95%+ target accuracy is being met
func (aat *AdvancedAccuracyTracker) IsTargetAccuracyMet() bool {
	overallAccuracy := aat.GetOverallAccuracy()
	return overallAccuracy >= aat.config.TargetAccuracy
}

// IsCriticalThresholdBreached checks if accuracy has fallen below critical threshold
func (aat *AdvancedAccuracyTracker) IsCriticalThresholdBreached() bool {
	overallAccuracy := aat.GetOverallAccuracy()
	return overallAccuracy < aat.config.CriticalAccuracyThreshold
}

// GetAccuracyStatus returns the current accuracy status
func (aat *AdvancedAccuracyTracker) GetAccuracyStatus() string {
	overallAccuracy := aat.GetOverallAccuracy()

	switch {
	case overallAccuracy >= aat.config.TargetAccuracy:
		return "excellent"
	case overallAccuracy >= aat.config.WarningAccuracyThreshold:
		return "good"
	case overallAccuracy >= aat.config.CriticalAccuracyThreshold:
		return "warning"
	default:
		return "critical"
	}
}

// monitoringLoop runs the main monitoring loop
func (aat *AdvancedAccuracyTracker) monitoringLoop() {
	ticker := time.NewTicker(aat.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-aat.ctx.Done():
			aat.logger.Info("Monitoring loop stopped")
			return
		case <-ticker.C:
			aat.performMonitoringCycle()
		}
	}
}

// alertingLoop runs the alerting loop
func (aat *AdvancedAccuracyTracker) alertingLoop() {
	ticker := time.NewTicker(aat.config.AlertCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-aat.ctx.Done():
			aat.logger.Info("Alerting loop stopped")
			return
		case <-ticker.C:
			aat.checkAlertConditions()
		}
	}
}

// trendAnalysisLoop runs the trend analysis loop
func (aat *AdvancedAccuracyTracker) trendAnalysisLoop() {
	ticker := time.NewTicker(aat.config.TrendAnalysisInterval)
	defer ticker.Stop()

	for {
		select {
		case <-aat.ctx.Done():
			aat.logger.Info("Trend analysis loop stopped")
			return
		case <-ticker.C:
			aat.performTrendAnalysis()
		}
	}
}

// dataCleanupLoop runs the data cleanup loop
func (aat *AdvancedAccuracyTracker) dataCleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-aat.ctx.Done():
			aat.logger.Info("Data cleanup loop stopped")
			return
		case <-ticker.C:
			aat.performDataCleanup()
		}
	}
}

// performMonitoringCycle performs a complete monitoring cycle
func (aat *AdvancedAccuracyTracker) performMonitoringCycle() {
	startTime := time.Now()

	// Update all metrics
	aat.overallTracker.UpdateMetrics()
	aat.industryTracker.UpdateMetrics()
	aat.ensembleTracker.UpdateMetrics()
	aat.mlModelTracker.UpdateMetrics()

	if aat.config.EnableSecurityTracking {
		aat.securityTracker.UpdateMetrics()
	}

	// Update real-time metrics
	aat.updateRealTimeMetrics()

	// Create historical snapshot
	aat.createHistoricalSnapshot()

	// Log monitoring cycle completion
	cycleTime := time.Since(startTime)
	aat.logger.Debug("Monitoring cycle completed",
		zap.Duration("cycle_time", cycleTime),
		zap.Float64("overall_accuracy", aat.GetOverallAccuracy()),
		zap.String("accuracy_status", aat.GetAccuracyStatus()))
}

// updateRealTimeMetrics updates real-time metrics
func (aat *AdvancedAccuracyTracker) updateRealTimeMetrics() {
	aat.realTimeMetrics.mu.Lock()
	defer aat.realTimeMetrics.mu.Unlock()

	aat.realTimeMetrics.CurrentAccuracy = aat.GetOverallAccuracy()
	aat.realTimeMetrics.CurrentThroughput = aat.calculateCurrentThroughput()
	aat.realTimeMetrics.CurrentLatency = aat.performanceMonitor.GetAverageLatency()
	aat.realTimeMetrics.ActiveAlerts = aat.alertManager.GetActiveAlertCount()
	aat.realTimeMetrics.LastUpdate = time.Now()
	aat.realTimeMetrics.HealthStatus = aat.calculateHealthStatus()
}

// createHistoricalSnapshot creates a historical accuracy snapshot
func (aat *AdvancedAccuracyTracker) createHistoricalSnapshot() {
	snapshot := &HistoricalAccuracySnapshot{
		Timestamp:          time.Now(),
		OverallAccuracy:    aat.GetOverallAccuracy(),
		IndustryAccuracies: aat.industryTracker.GetAllIndustryAccuracies(),
		MethodAccuracies:   aat.ensembleTracker.GetAllMethodAccuracies(),
		MLModelAccuracies:  aat.mlModelTracker.GetAllModelAccuracies(),
		SecurityMetrics:    aat.GetSecurityMetrics(),
		PerformanceMetrics: aat.performanceMonitor.GetPerformanceSnapshot(),
	}

	aat.mu.Lock()
	aat.historicalData = append(aat.historicalData, snapshot)

	// Limit historical data size
	if len(aat.historicalData) > aat.config.MaxHistoricalSnapshots {
		aat.historicalData = aat.historicalData[1:]
	}
	aat.mu.Unlock()
}

// checkAlertConditions checks for alert conditions
func (aat *AdvancedAccuracyTracker) checkAlertConditions() {
	if !aat.config.EnableAlerting {
		return
	}

	// Check overall accuracy
	overallAccuracy := aat.GetOverallAccuracy()

	if overallAccuracy < aat.config.CriticalAccuracyThreshold {
		aat.alertManager.CreateAlert("accuracy_critical", "critical",
			fmt.Sprintf("Overall accuracy %.2f%% is below critical threshold %.2f%%",
				overallAccuracy*100, aat.config.CriticalAccuracyThreshold*100),
			"overall", overallAccuracy, aat.config.CriticalAccuracyThreshold)
	} else if overallAccuracy < aat.config.WarningAccuracyThreshold {
		aat.alertManager.CreateAlert("accuracy_warning", "warning",
			fmt.Sprintf("Overall accuracy %.2f%% is below warning threshold %.2f%%",
				overallAccuracy*100, aat.config.WarningAccuracyThreshold*100),
			"overall", overallAccuracy, aat.config.WarningAccuracyThreshold)
	}

	// Check target accuracy achievement
	if overallAccuracy >= aat.config.TargetAccuracy {
		aat.alertManager.CreateAlert("target_achieved", "info",
			fmt.Sprintf("Target accuracy of %.2f%% has been achieved (current: %.2f%%)",
				aat.config.TargetAccuracy*100, overallAccuracy*100),
			"overall", overallAccuracy, aat.config.TargetAccuracy)
	}

	// Check security metrics if enabled
	if aat.config.EnableSecurityTracking {
		aat.checkSecurityAlerts()
	}

	// Check performance metrics if enabled
	if aat.config.EnablePerformanceTracking {
		aat.checkPerformanceAlerts()
	}
}

// checkSecurityAlerts checks for security-related alerts
func (aat *AdvancedAccuracyTracker) checkSecurityAlerts() {
	securityMetrics := aat.GetSecurityMetrics()

	if securityMetrics.TrustedDataSourceRate < aat.config.SecurityTrustTarget {
		aat.alertManager.CreateAlert("security_trust_low", "high",
			fmt.Sprintf("Trusted data source rate %.2f%% is below target %.2f%%",
				securityMetrics.TrustedDataSourceRate*100, aat.config.SecurityTrustTarget*100),
			"security", securityMetrics.TrustedDataSourceRate, aat.config.SecurityTrustTarget)
	}
}

// checkPerformanceAlerts checks for performance-related alerts
func (aat *AdvancedAccuracyTracker) checkPerformanceAlerts() {
	avgLatency := aat.performanceMonitor.GetAverageLatency()

	if avgLatency > aat.config.MaxProcessingTime {
		aat.alertManager.CreateAlert("performance_latency_high", "medium",
			fmt.Sprintf("Average latency %v exceeds maximum %v",
				avgLatency, aat.config.MaxProcessingTime),
			"performance", float64(avgLatency.Nanoseconds()), float64(aat.config.MaxProcessingTime.Nanoseconds()))
	}
}

// performTrendAnalysis performs trend analysis on all dimensions
func (aat *AdvancedAccuracyTracker) performTrendAnalysis() {
	// Analyze overall trend
	aat.analyzeOverallTrend()

	// Analyze industry trends
	aat.analyzeIndustryTrends()

	// Analyze method trends
	aat.analyzeMethodTrends()

	// Analyze ML model trends
	aat.analyzeMLModelTrends()
}

// analyzeOverallTrend analyzes the overall accuracy trend
func (aat *AdvancedAccuracyTracker) analyzeOverallTrend() {
	windowedAccuracy := aat.overallTracker.GetWindowedAccuracy()
	if len(windowedAccuracy) < aat.config.TrendWindowSize {
		return
	}

	trend := aat.calculateTrend(windowedAccuracy)
	trendStrength := aat.calculateTrendStrength(windowedAccuracy)

	aat.trendAnalyzer.UpdateTrend("overall", "all", trend, trendStrength)
}

// analyzeIndustryTrends analyzes trends for all industries
func (aat *AdvancedAccuracyTracker) analyzeIndustryTrends() {
	industries := aat.industryTracker.GetAllIndustries()

	for _, industry := range industries {
		windowedAccuracy := aat.industryTracker.GetIndustryWindowedAccuracy(industry)
		if len(windowedAccuracy) < aat.config.TrendWindowSize {
			continue
		}

		trend := aat.calculateTrend(windowedAccuracy)
		trendStrength := aat.calculateTrendStrength(windowedAccuracy)

		aat.trendAnalyzer.UpdateTrend("industry", industry, trend, trendStrength)
	}
}

// analyzeMethodTrends analyzes trends for all methods
func (aat *AdvancedAccuracyTracker) analyzeMethodTrends() {
	methods := aat.ensembleTracker.GetAllMethods()

	for _, method := range methods {
		windowedAccuracy := aat.ensembleTracker.GetMethodWindowedAccuracy(method)
		if len(windowedAccuracy) < aat.config.TrendWindowSize {
			continue
		}

		trend := aat.calculateTrend(windowedAccuracy)
		trendStrength := aat.calculateTrendStrength(windowedAccuracy)

		aat.trendAnalyzer.UpdateTrend("method", method, trend, trendStrength)
	}
}

// analyzeMLModelTrends analyzes trends for all ML models
func (aat *AdvancedAccuracyTracker) analyzeMLModelTrends() {
	models := aat.mlModelTracker.GetAllModels()

	for _, model := range models {
		windowedAccuracy := aat.mlModelTracker.GetModelWindowedAccuracy(model)
		if len(windowedAccuracy) < aat.config.TrendWindowSize {
			continue
		}

		trend := aat.calculateTrend(windowedAccuracy)
		trendStrength := aat.calculateTrendStrength(windowedAccuracy)

		aat.trendAnalyzer.UpdateTrend("ml_model", model, trend, trendStrength)
	}
}

// calculateTrend calculates the trend direction from windowed accuracy data
func (aat *AdvancedAccuracyTracker) calculateTrend(windowedAccuracy []float64) string {
	if len(windowedAccuracy) < 2 {
		return "insufficient_data"
	}

	// Calculate simple linear trend
	firstHalf := windowedAccuracy[:len(windowedAccuracy)/2]
	secondHalf := windowedAccuracy[len(windowedAccuracy)/2:]

	firstAvg := aat.calculateAverage(firstHalf)
	secondAvg := aat.calculateAverage(secondHalf)

	diff := secondAvg - firstAvg

	if diff > 0.02 { // 2% improvement
		return "improving"
	} else if diff < -0.02 { // 2% decline
		return "declining"
	} else {
		return "stable"
	}
}

// calculateTrendStrength calculates the strength of the trend
func (aat *AdvancedAccuracyTracker) calculateTrendStrength(windowedAccuracy []float64) float64 {
	if len(windowedAccuracy) < 2 {
		return 0.0
	}

	// Calculate standard deviation as a measure of trend strength
	return aat.calculateStandardDeviation(windowedAccuracy)
}

// calculateAverage calculates the average of a slice of floats
func (aat *AdvancedAccuracyTracker) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateStandardDeviation calculates the standard deviation of a slice of floats
func (aat *AdvancedAccuracyTracker) calculateStandardDeviation(values []float64) float64 {
	if len(values) < 2 {
		return 0.0
	}

	mean := aat.calculateAverage(values)

	sumSquaredDiffs := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquaredDiffs += diff * diff
	}

	variance := sumSquaredDiffs / float64(len(values)-1)
	return math.Sqrt(variance)
}

// calculateCurrentThroughput calculates current throughput
func (aat *AdvancedAccuracyTracker) calculateCurrentThroughput() float64 {
	// Calculate requests per second based on recent activity
	now := time.Now()
	recentClassifications := 0

	aat.mu.RLock()
	for _, snapshot := range aat.historicalData {
		if now.Sub(snapshot.Timestamp) <= time.Minute {
			recentClassifications++
		}
	}
	aat.mu.RUnlock()

	return float64(recentClassifications) / 60.0 // per second
}

// calculateHealthStatus calculates the overall health status
func (aat *AdvancedAccuracyTracker) calculateHealthStatus() string {
	overallAccuracy := aat.GetOverallAccuracy()
	activeAlerts := aat.alertManager.GetActiveAlertCount()

	if overallAccuracy >= aat.config.TargetAccuracy && activeAlerts == 0 {
		return "healthy"
	} else if overallAccuracy >= aat.config.WarningAccuracyThreshold && activeAlerts <= 2 {
		return "warning"
	} else if overallAccuracy >= aat.config.CriticalAccuracyThreshold {
		return "degraded"
	} else {
		return "critical"
	}
}

// performDataCleanup performs cleanup of old data
func (aat *AdvancedAccuracyTracker) performDataCleanup() {
	cutoffTime := time.Now().Add(-aat.config.HistoricalDataRetention)

	aat.mu.Lock()
	defer aat.mu.Unlock()

	// Clean up historical data
	var cleanedData []*HistoricalAccuracySnapshot
	for _, snapshot := range aat.historicalData {
		if snapshot.Timestamp.After(cutoffTime) {
			cleanedData = append(cleanedData, snapshot)
		}
	}
	aat.historicalData = cleanedData

	aat.logger.Debug("Data cleanup completed",
		zap.Int("remaining_snapshots", len(aat.historicalData)),
		zap.Time("cutoff_time", cutoffTime))
}
