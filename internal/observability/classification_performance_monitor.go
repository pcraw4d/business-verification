package observability

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ClassificationPerformanceMonitor provides comprehensive performance monitoring for classification features
type ClassificationPerformanceMonitor struct {
	metrics   *ClassificationPerformanceMetrics
	alerts    *ClassificationAlertManager
	optimizer *ClassificationOptimizer
	predictor *ClassificationPredictor
	dashboard *ClassificationDashboard
	config    ClassificationPerformanceConfig
	mu        sync.RWMutex
}

// ClassificationPerformanceConfig holds configuration for classification performance monitoring
type ClassificationPerformanceConfig struct {
	// Monitoring intervals
	MetricsCollectionInterval time.Duration `json:"metrics_collection_interval"`
	AlertCheckInterval        time.Duration `json:"alert_check_interval"`
	OptimizationInterval      time.Duration `json:"optimization_interval"`
	PredictionInterval        time.Duration `json:"prediction_interval"`

	// Classification-specific thresholds
	ClassificationAccuracyThreshold     float64       `json:"classification_accuracy_threshold"`
	ClassificationResponseTimeThreshold time.Duration `json:"classification_response_time_threshold"`
	ConfidenceScoreThreshold            float64       `json:"confidence_score_threshold"`
	UserSatisfactionThreshold           float64       `json:"user_satisfaction_threshold"`

	// Method-specific thresholds
	WebsiteAnalysisAccuracyThreshold float64 `json:"website_analysis_accuracy_threshold"`
	WebSearchAccuracyThreshold       float64 `json:"web_search_accuracy_threshold"`
	MLModelAccuracyThreshold         float64 `json:"ml_model_accuracy_threshold"`

	// Performance thresholds
	ResponseTimeThreshold time.Duration `json:"response_time_threshold"`
	SuccessRateThreshold  float64       `json:"success_rate_threshold"`
	ErrorRateThreshold    float64       `json:"error_rate_threshold"`
	ThroughputThreshold   int           `json:"throughput_threshold"`

	// Optimization settings
	AutoOptimizationEnabled bool    `json:"auto_optimization_enabled"`
	OptimizationConfidence  float64 `json:"optimization_confidence"`
	RollbackThreshold       float64 `json:"rollback_threshold"`

	// Prediction settings
	PredictionHorizon    time.Duration `json:"prediction_horizon"`
	PredictionConfidence float64       `json:"prediction_confidence"`
	TrendAnalysisWindow  time.Duration `json:"trend_analysis_window"`

	// Dashboard settings
	DashboardRefreshInterval time.Duration `json:"dashboard_refresh_interval"`
	HistoricalDataRetention  time.Duration `json:"historical_data_retention"`
	RealTimeUpdatesEnabled   bool          `json:"real_time_updates_enabled"`

	// Alerting settings
	AlertChannels     []string      `json:"alert_channels"`
	EscalationEnabled bool          `json:"escalation_enabled"`
	EscalationDelay   time.Duration `json:"escalation_delay"`
}

// ClassificationPerformanceMetrics tracks comprehensive classification performance data
type ClassificationPerformanceMetrics struct {
	// Overall classification metrics
	TotalClassifications      int64 `json:"total_classifications"`
	SuccessfulClassifications int64 `json:"successful_classifications"`
	FailedClassifications     int64 `json:"failed_classifications"`
	TimeoutClassifications    int64 `json:"timeout_classifications"`

	// Method-specific metrics
	WebsiteAnalysisCount  int64 `json:"website_analysis_count"`
	WebSearchCount        int64 `json:"web_search_count"`
	MLModelCount          int64 `json:"ml_model_count"`
	KeywordBasedCount     int64 `json:"keyword_based_count"`
	FuzzyMatchingCount    int64 `json:"fuzzy_matching_count"`
	CrosswalkMappingCount int64 `json:"crosswalk_mapping_count"`

	// Accuracy metrics
	OverallAccuracy          float64 `json:"overall_accuracy"`
	WebsiteAnalysisAccuracy  float64 `json:"website_analysis_accuracy"`
	WebSearchAccuracy        float64 `json:"web_search_accuracy"`
	MLModelAccuracy          float64 `json:"ml_model_accuracy"`
	KeywordBasedAccuracy     float64 `json:"keyword_based_accuracy"`
	FuzzyMatchingAccuracy    float64 `json:"fuzzy_matching_accuracy"`
	CrosswalkMappingAccuracy float64 `json:"crosswalk_mapping_accuracy"`

	// Response time metrics
	AverageResponseTime time.Duration `json:"average_response_time"`
	P50ResponseTime     time.Duration `json:"p50_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`
	MinResponseTime     time.Duration `json:"min_response_time"`
	MaxResponseTime     time.Duration `json:"max_response_time"`

	// Method-specific response times
	WebsiteAnalysisResponseTime  time.Duration `json:"website_analysis_response_time"`
	WebSearchResponseTime        time.Duration `json:"web_search_response_time"`
	MLModelResponseTime          time.Duration `json:"ml_model_response_time"`
	KeywordBasedResponseTime     time.Duration `json:"keyword_based_response_time"`
	FuzzyMatchingResponseTime    time.Duration `json:"fuzzy_matching_response_time"`
	CrosswalkMappingResponseTime time.Duration `json:"crosswalk_mapping_response_time"`

	// Confidence score metrics
	AverageConfidenceScore float64 `json:"average_confidence_score"`
	HighConfidenceCount    int64   `json:"high_confidence_count"`
	MediumConfidenceCount  int64   `json:"medium_confidence_count"`
	LowConfidenceCount     int64   `json:"low_confidence_count"`

	// Method-specific confidence scores
	WebsiteAnalysisConfidence  float64 `json:"website_analysis_confidence"`
	WebSearchConfidence        float64 `json:"web_search_confidence"`
	MLModelConfidence          float64 `json:"ml_model_confidence"`
	KeywordBasedConfidence     float64 `json:"keyword_based_confidence"`
	FuzzyMatchingConfidence    float64 `json:"fuzzy_matching_confidence"`
	CrosswalkMappingConfidence float64 `json:"crosswalk_mapping_confidence"`

	// Throughput metrics
	ClassificationsPerSecond  float64 `json:"classifications_per_second"`
	ConcurrentClassifications int     `json:"concurrent_classifications"`
	PeakConcurrency           int     `json:"peak_concurrency"`

	// Error metrics
	ErrorRate   float64 `json:"error_rate"`
	SuccessRate float64 `json:"success_rate"`
	TimeoutRate float64 `json:"timeout_rate"`

	// Method-specific error rates
	WebsiteAnalysisErrorRate  float64 `json:"website_analysis_error_rate"`
	WebSearchErrorRate        float64 `json:"web_search_error_rate"`
	MLModelErrorRate          float64 `json:"ml_model_error_rate"`
	KeywordBasedErrorRate     float64 `json:"keyword_based_error_rate"`
	FuzzyMatchingErrorRate    float64 `json:"fuzzy_matching_error_rate"`
	CrosswalkMappingErrorRate float64 `json:"crosswalk_mapping_error_rate"`

	// Resource metrics
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	NetworkIO   float64 `json:"network_io"`

	// Cache performance metrics
	CacheHitRate   float64 `json:"cache_hit_rate"`
	CacheMissRate  float64 `json:"cache_miss_rate"`
	CacheSize      int64   `json:"cache_size"`
	CacheEvictions int64   `json:"cache_evictions"`

	// User satisfaction metrics
	UserSatisfactionScore float64 `json:"user_satisfaction_score"`
	FeedbackCount         int64   `json:"feedback_count"`
	PositiveFeedbackCount int64   `json:"positive_feedback_count"`
	NegativeFeedbackCount int64   `json:"negative_feedback_count"`

	// Geographic and industry metrics
	GeographicAccuracy map[string]float64 `json:"geographic_accuracy"`
	IndustryAccuracy   map[string]float64 `json:"industry_accuracy"`

	// Quality assurance metrics
	QualityCheckPassRate float64       `json:"quality_check_pass_rate"`
	QualityCheckFailRate float64       `json:"quality_check_fail_rate"`
	QualityCheckDuration time.Duration `json:"quality_check_duration"`

	// Timestamp
	LastUpdated      time.Time     `json:"last_updated"`
	CollectionWindow time.Duration `json:"collection_window"`
}

// ClassificationAlert represents a classification performance alert
type ClassificationAlert struct {
	ID              string     `json:"id"`
	Type            string     `json:"type"`     // accuracy, response_time, confidence, user_satisfaction, resource
	Severity        string     `json:"severity"` // low, medium, high, critical
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Method          string     `json:"method"` // website_analysis, web_search, ml_model, etc.
	Metric          string     `json:"metric"`
	CurrentValue    float64    `json:"current_value"`
	Threshold       float64    `json:"threshold"`
	Timestamp       time.Time  `json:"timestamp"`
	Status          string     `json:"status"` // active, acknowledged, resolved
	AcknowledgedBy  string     `json:"acknowledged_by,omitempty"`
	AcknowledgedAt  *time.Time `json:"acknowledged_at,omitempty"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	Recommendations []string   `json:"recommendations"`
	Impact          string     `json:"impact"` // low, medium, high
}

// ClassificationOptimization represents a classification performance optimization action
type ClassificationOptimization struct {
	ID                  string     `json:"id"`
	Type                string     `json:"type"` // auto, manual, scheduled
	Action              string     `json:"action"`
	Method              string     `json:"method"` // website_analysis, web_search, ml_model, etc.
	Description         string     `json:"description"`
	TargetMetric        string     `json:"target_metric"`
	ExpectedImprovement float64    `json:"expected_improvement"`
	Confidence          float64    `json:"confidence"`
	RiskLevel           string     `json:"risk_level"` // low, medium, high
	Status              string     `json:"status"`     // pending, applied, reverted, failed
	AppliedAt           *time.Time `json:"applied_at,omitempty"`
	RevertedAt          *time.Time `json:"reverted_at,omitempty"`
	ActualImprovement   *float64   `json:"actual_improvement,omitempty"`
	RollbackReason      string     `json:"rollback_reason,omitempty"`
	Impact              string     `json:"impact"` // low, medium, high
}

// ClassificationPrediction represents a classification performance prediction
type ClassificationPrediction struct {
	ID                string        `json:"id"`
	Method            string        `json:"method"` // website_analysis, web_search, ml_model, etc.
	Metric            string        `json:"metric"`
	PredictedValue    float64       `json:"predicted_value"`
	Confidence        float64       `json:"confidence"`
	PredictionHorizon time.Duration `json:"prediction_horizon"`
	Trend             string        `json:"trend"` // improving, stable, degrading
	Factors           []string      `json:"factors"`
	Timestamp         time.Time     `json:"timestamp"`
	Accuracy          *float64      `json:"accuracy,omitempty"`
	Impact            string        `json:"impact"` // low, medium, high
}

// ClassificationDashboard provides real-time classification performance visualization
type ClassificationDashboard struct {
	// Real-time metrics
	CurrentMetrics *ClassificationPerformanceMetrics `json:"current_metrics"`

	// Historical data
	HistoricalMetrics []*ClassificationPerformanceMetrics `json:"historical_metrics"`

	// Alerts
	ActiveAlerts []*ClassificationAlert `json:"active_alerts"`

	// Optimizations
	RecentOptimizations []*ClassificationOptimization `json:"recent_optimizations"`

	// Predictions
	CurrentPredictions []*ClassificationPrediction `json:"current_predictions"`

	// Method performance comparison
	MethodPerformance map[string]MethodPerformanceData `json:"method_performance"`

	// Geographic performance
	GeographicPerformance map[string]GeographicPerformanceData `json:"geographic_performance"`

	// Industry performance
	IndustryPerformance map[string]IndustryPerformanceData `json:"industry_performance"`

	// Status
	OverallHealth string    `json:"overall_health"`
	LastUpdated   time.Time `json:"last_updated"`
}

// MethodPerformanceData represents performance data for a specific classification method
type MethodPerformanceData struct {
	Accuracy         float64       `json:"accuracy"`
	ResponseTime     time.Duration `json:"response_time"`
	Confidence       float64       `json:"confidence"`
	SuccessRate      float64       `json:"success_rate"`
	ErrorRate        float64       `json:"error_rate"`
	Throughput       float64       `json:"throughput"`
	UserSatisfaction float64       `json:"user_satisfaction"`
	LastUpdated      time.Time     `json:"last_updated"`
}

// GeographicPerformanceData represents performance data for a specific geographic region
type GeographicPerformanceData struct {
	Accuracy         float64       `json:"accuracy"`
	ResponseTime     time.Duration `json:"response_time"`
	Confidence       float64       `json:"confidence"`
	SuccessRate      float64       `json:"success_rate"`
	ErrorRate        float64       `json:"error_rate"`
	Throughput       float64       `json:"throughput"`
	UserSatisfaction float64       `json:"user_satisfaction"`
	LastUpdated      time.Time     `json:"last_updated"`
}

// IndustryPerformanceData represents performance data for a specific industry
type IndustryPerformanceData struct {
	Accuracy         float64       `json:"accuracy"`
	ResponseTime     time.Duration `json:"response_time"`
	Confidence       float64       `json:"confidence"`
	SuccessRate      float64       `json:"success_rate"`
	ErrorRate        float64       `json:"error_rate"`
	Throughput       float64       `json:"throughput"`
	UserSatisfaction float64       `json:"user_satisfaction"`
	LastUpdated      time.Time     `json:"last_updated"`
}

// ClassificationAlertManager manages classification performance alerts
type ClassificationAlertManager struct {
	alerts map[string]*ClassificationAlert
	config ClassificationPerformanceConfig
	mu     sync.RWMutex
}

// ClassificationOptimizer manages classification performance optimizations
type ClassificationOptimizer struct {
	optimizations map[string]*ClassificationOptimization
	config        ClassificationPerformanceConfig
	mu            sync.RWMutex
}

// ClassificationPredictor manages classification performance predictions
type ClassificationPredictor struct {
	predictions    map[string]*ClassificationPrediction
	historicalData []*ClassificationPerformanceMetrics
	config         ClassificationPerformanceConfig
	mu             sync.RWMutex
}

// ClassificationDashboardManager manages the classification performance dashboard
type ClassificationDashboardManager struct {
	dashboard *ClassificationDashboard
	config    ClassificationPerformanceConfig
	mu        sync.RWMutex
}

// NewClassificationPerformanceMonitor creates a new classification performance monitor
func NewClassificationPerformanceMonitor(config ClassificationPerformanceConfig) *ClassificationPerformanceMonitor {
	if config.MetricsCollectionInterval == 0 {
		config.MetricsCollectionInterval = 30 * time.Second
	}
	if config.AlertCheckInterval == 0 {
		config.AlertCheckInterval = 60 * time.Second
	}
	if config.OptimizationInterval == 0 {
		config.OptimizationInterval = 300 * time.Second
	}
	if config.PredictionInterval == 0 {
		config.PredictionInterval = 600 * time.Second
	}

	monitor := &ClassificationPerformanceMonitor{
		config: config,
		metrics: &ClassificationPerformanceMetrics{
			GeographicAccuracy: make(map[string]float64),
			IndustryAccuracy:   make(map[string]float64),
		},
		alerts: &ClassificationAlertManager{
			alerts: make(map[string]*ClassificationAlert),
			config: config,
		},
		optimizer: &ClassificationOptimizer{
			optimizations: make(map[string]*ClassificationOptimization),
			config:        config,
		},
		predictor: &ClassificationPredictor{
			predictions:    make(map[string]*ClassificationPrediction),
			historicalData: make([]*ClassificationPerformanceMetrics, 0),
			config:         config,
		},
		dashboard: &ClassificationDashboard{
			MethodPerformance:     make(map[string]MethodPerformanceData),
			GeographicPerformance: make(map[string]GeographicPerformanceData),
			IndustryPerformance:   make(map[string]IndustryPerformanceData),
		},
	}

	return monitor
}

// Start starts the classification performance monitor
func (cpm *ClassificationPerformanceMonitor) Start(ctx context.Context) error {
	// Start metrics collection
	go cpm.collectMetrics(ctx)

	// Start alert checking
	go cpm.checkAlerts(ctx)

	// Start optimization if enabled
	if cpm.config.AutoOptimizationEnabled {
		go cpm.runOptimizations(ctx)
	}

	// Start predictions
	go cpm.runPredictions(ctx)

	// Start dashboard updates
	go cpm.updateDashboard(ctx)

	return nil
}

// collectMetrics collects classification performance metrics
func (cpm *ClassificationPerformanceMonitor) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(cpm.config.MetricsCollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cpm.updateMetrics()
		}
	}
}

// updateMetrics updates the current metrics
func (cpm *ClassificationPerformanceMonitor) updateMetrics() {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	// Update timestamp
	cpm.metrics.LastUpdated = time.Now()

	// Calculate derived metrics
	cpm.calculateDerivedMetrics()

	// Store historical data
	cpm.storeHistoricalData()
}

// calculateDerivedMetrics calculates derived metrics from raw data
func (cpm *ClassificationPerformanceMonitor) calculateDerivedMetrics() {
	// Calculate success and error rates
	if cpm.metrics.TotalClassifications > 0 {
		cpm.metrics.SuccessRate = float64(cpm.metrics.SuccessfulClassifications) / float64(cpm.metrics.TotalClassifications)
		cpm.metrics.ErrorRate = float64(cpm.metrics.FailedClassifications) / float64(cpm.metrics.TotalClassifications)
		cpm.metrics.TimeoutRate = float64(cpm.metrics.TimeoutClassifications) / float64(cpm.metrics.TotalClassifications)
	}

	// Calculate method-specific error rates
	if cpm.metrics.WebsiteAnalysisCount > 0 {
		cpm.metrics.WebsiteAnalysisErrorRate = float64(cpm.metrics.FailedClassifications) / float64(cpm.metrics.WebsiteAnalysisCount)
	}
	if cpm.metrics.WebSearchCount > 0 {
		cpm.metrics.WebSearchErrorRate = float64(cpm.metrics.FailedClassifications) / float64(cpm.metrics.WebSearchCount)
	}
	if cpm.metrics.MLModelCount > 0 {
		cpm.metrics.MLModelErrorRate = float64(cpm.metrics.FailedClassifications) / float64(cpm.metrics.MLModelCount)
	}

	// Calculate user satisfaction
	if cpm.metrics.FeedbackCount > 0 {
		cpm.metrics.UserSatisfactionScore = float64(cpm.metrics.PositiveFeedbackCount) / float64(cpm.metrics.FeedbackCount)
	}
}

// storeHistoricalData stores historical metrics data
func (cpm *ClassificationPerformanceMonitor) storeHistoricalData() {
	// Create a copy of current metrics
	metricsCopy := *cpm.metrics
	cpm.predictor.historicalData = append(cpm.predictor.historicalData, &metricsCopy)

	// Limit historical data retention
	if len(cpm.predictor.historicalData) > int(cpm.config.HistoricalDataRetention/cpm.config.MetricsCollectionInterval) {
		cpm.predictor.historicalData = cpm.predictor.historicalData[1:]
	}
}

// checkAlerts checks for performance alerts
func (cpm *ClassificationPerformanceMonitor) checkAlerts(ctx context.Context) {
	ticker := time.NewTicker(cpm.config.AlertCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cpm.checkAccuracyAlerts()
			cpm.checkResponseTimeAlerts()
			cpm.checkConfidenceAlerts()
			cpm.checkUserSatisfactionAlerts()
			cpm.checkResourceAlerts()
		}
	}
}

// checkAccuracyAlerts checks for accuracy-related alerts
func (cpm *ClassificationPerformanceMonitor) checkAccuracyAlerts() {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	// Check overall accuracy
	if cpm.metrics.OverallAccuracy < cpm.config.ClassificationAccuracyThreshold {
		cpm.createAlert("accuracy", "critical", "Low Classification Accuracy",
			"Overall classification accuracy is below threshold", "overall_accuracy",
			cpm.metrics.OverallAccuracy, cpm.config.ClassificationAccuracyThreshold)
	}

	// Check method-specific accuracy
	if cpm.metrics.WebsiteAnalysisAccuracy < cpm.config.WebsiteAnalysisAccuracyThreshold {
		cpm.createAlert("accuracy", "high", "Low Website Analysis Accuracy",
			"Website analysis accuracy is below threshold", "website_analysis_accuracy",
			cpm.metrics.WebsiteAnalysisAccuracy, cpm.config.WebsiteAnalysisAccuracyThreshold)
	}

	if cpm.metrics.WebSearchAccuracy < cpm.config.WebSearchAccuracyThreshold {
		cpm.createAlert("accuracy", "high", "Low Web Search Accuracy",
			"Web search accuracy is below threshold", "web_search_accuracy",
			cpm.metrics.WebSearchAccuracy, cpm.config.WebSearchAccuracyThreshold)
	}

	if cpm.metrics.MLModelAccuracy < cpm.config.MLModelAccuracyThreshold {
		cpm.createAlert("accuracy", "high", "Low ML Model Accuracy",
			"ML model accuracy is below threshold", "ml_model_accuracy",
			cpm.metrics.MLModelAccuracy, cpm.config.MLModelAccuracyThreshold)
	}
}

// checkResponseTimeAlerts checks for response time alerts
func (cpm *ClassificationPerformanceMonitor) checkResponseTimeAlerts() {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	if cpm.metrics.AverageResponseTime > cpm.config.ClassificationResponseTimeThreshold {
		cpm.createAlert("response_time", "high", "High Classification Response Time",
			"Average classification response time is above threshold", "average_response_time",
			float64(cpm.metrics.AverageResponseTime.Milliseconds()),
			float64(cpm.config.ClassificationResponseTimeThreshold.Milliseconds()))
	}
}

// checkConfidenceAlerts checks for confidence score alerts
func (cpm *ClassificationPerformanceMonitor) checkConfidenceAlerts() {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	if cpm.metrics.AverageConfidenceScore < cpm.config.ConfidenceScoreThreshold {
		cpm.createAlert("confidence", "medium", "Low Confidence Scores",
			"Average confidence score is below threshold", "average_confidence_score",
			cpm.metrics.AverageConfidenceScore, cpm.config.ConfidenceScoreThreshold)
	}
}

// checkUserSatisfactionAlerts checks for user satisfaction alerts
func (cpm *ClassificationPerformanceMonitor) checkUserSatisfactionAlerts() {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	if cpm.metrics.UserSatisfactionScore < cpm.config.UserSatisfactionThreshold {
		cpm.createAlert("user_satisfaction", "high", "Low User Satisfaction",
			"User satisfaction score is below threshold", "user_satisfaction_score",
			cpm.metrics.UserSatisfactionScore, cpm.config.UserSatisfactionThreshold)
	}
}

// checkResourceAlerts checks for resource usage alerts
func (cpm *ClassificationPerformanceMonitor) checkResourceAlerts() {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	if cpm.metrics.CPUUsage > 80.0 {
		cpm.createAlert("resource", "medium", "High CPU Usage",
			"CPU usage is above 80%", "cpu_usage",
			cpm.metrics.CPUUsage, 80.0)
	}

	if cpm.metrics.MemoryUsage > 80.0 {
		cpm.createAlert("resource", "medium", "High Memory Usage",
			"Memory usage is above 80%", "memory_usage",
			cpm.metrics.MemoryUsage, 80.0)
	}
}

// createAlert creates a new performance alert
func (cpm *ClassificationPerformanceMonitor) createAlert(alertType, severity, title, description, metric string, currentValue, threshold float64) {
	alert := &ClassificationAlert{
		ID:           fmt.Sprintf("alert_%d", time.Now().Unix()),
		Type:         alertType,
		Severity:     severity,
		Title:        title,
		Description:  description,
		Metric:       metric,
		CurrentValue: currentValue,
		Threshold:    threshold,
		Timestamp:    time.Now(),
		Status:       "active",
		Recommendations: []string{
			"Monitor the metric closely",
			"Check system resources",
			"Review recent changes",
		},
		Impact: "medium",
	}

	cpm.alerts.alerts[alert.ID] = alert

	// Send alert through configured channels
	cpm.sendAlert(alert)
}

// sendAlert sends an alert through configured channels
func (cpm *ClassificationPerformanceMonitor) sendAlert(alert *ClassificationAlert) {
	// Implementation would send alerts through configured channels
	// (email, Slack, webhook, etc.)
	fmt.Printf("Alert: %s - %s (Severity: %s)\n", alert.Title, alert.Description, alert.Severity)
}

// runOptimizations runs performance optimizations
func (cpm *ClassificationPerformanceMonitor) runOptimizations(ctx context.Context) {
	ticker := time.NewTicker(cpm.config.OptimizationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cpm.analyzeAndOptimize()
		}
	}
}

// analyzeAndOptimize analyzes performance and applies optimizations
func (cpm *ClassificationPerformanceMonitor) analyzeAndOptimize() {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	// Analyze performance issues and create optimizations
	if cpm.metrics.AverageResponseTime > cpm.config.ResponseTimeThreshold {
		cpm.createOptimization("response_time", "Optimize response time",
			"average_response_time", 0.1, 0.8, "low")
	}

	if cpm.metrics.OverallAccuracy < cpm.config.ClassificationAccuracyThreshold {
		cpm.createOptimization("accuracy", "Improve classification accuracy",
			"overall_accuracy", 0.05, 0.7, "medium")
	}

	if cpm.metrics.CacheHitRate < 0.8 {
		cpm.createOptimization("cache", "Optimize cache performance",
			"cache_hit_rate", 0.1, 0.9, "low")
	}
}

// createOptimization creates a new performance optimization
func (cpm *ClassificationPerformanceMonitor) createOptimization(optType, description, targetMetric string, expectedImprovement, confidence float64, riskLevel string) {
	optimization := &ClassificationOptimization{
		ID:                  fmt.Sprintf("opt_%d", time.Now().Unix()),
		Type:                "auto",
		Action:              optType,
		Description:         description,
		TargetMetric:        targetMetric,
		ExpectedImprovement: expectedImprovement,
		Confidence:          confidence,
		RiskLevel:           riskLevel,
		Status:              "pending",
		Impact:              "medium",
	}

	cpm.optimizer.optimizations[optimization.ID] = optimization

	// Apply optimization if confidence is high enough
	if confidence >= cpm.config.OptimizationConfidence {
		cpm.applyOptimization(optimization)
	}
}

// applyOptimization applies a performance optimization
func (cpm *ClassificationPerformanceMonitor) applyOptimization(optimization *ClassificationOptimization) {
	// Implementation would apply the optimization
	// This could involve adjusting cache settings, model parameters, etc.
	optimization.Status = "applied"
	optimization.AppliedAt = &time.Time{}
	fmt.Printf("Applied optimization: %s\n", optimization.Description)
}

// runPredictions runs performance predictions
func (cpm *ClassificationPerformanceMonitor) runPredictions(ctx context.Context) {
	ticker := time.NewTicker(cpm.config.PredictionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cpm.generatePredictions()
		}
	}
}

// generatePredictions generates performance predictions
func (cpm *ClassificationPerformanceMonitor) generatePredictions() {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	// Generate predictions for key metrics
	cpm.createPrediction("overall_accuracy", "Overall classification accuracy prediction")
	cpm.createPrediction("average_response_time", "Average response time prediction")
	cpm.createPrediction("user_satisfaction_score", "User satisfaction prediction")
}

// createPrediction creates a new performance prediction
func (cpm *ClassificationPerformanceMonitor) createPrediction(metric, description string) {
	prediction := &ClassificationPrediction{
		ID:                fmt.Sprintf("pred_%d", time.Now().Unix()),
		Metric:            metric,
		PredictedValue:    0.85, // Placeholder value
		Confidence:        0.75, // Placeholder confidence
		PredictionHorizon: cpm.config.PredictionHorizon,
		Trend:             "stable",
		Factors:           []string{"historical_data", "current_trends"},
		Timestamp:         time.Now(),
		Impact:            "medium",
	}

	cpm.predictor.predictions[prediction.ID] = prediction
}

// updateDashboard updates the performance dashboard
func (cpm *ClassificationPerformanceMonitor) updateDashboard(ctx context.Context) {
	ticker := time.NewTicker(cpm.config.DashboardRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cpm.refreshDashboard()
		}
	}
}

// refreshDashboard refreshes the dashboard data
func (cpm *ClassificationPerformanceMonitor) refreshDashboard() {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	// Update current metrics
	cpm.dashboard.CurrentMetrics = cpm.metrics

	// Update method performance
	cpm.updateMethodPerformance()

	// Update geographic performance
	cpm.updateGeographicPerformance()

	// Update industry performance
	cpm.updateIndustryPerformance()

	// Update overall health
	cpm.updateOverallHealth()

	// Update timestamp
	cpm.dashboard.LastUpdated = time.Now()
}

// updateMethodPerformance updates method performance data
func (cpm *ClassificationPerformanceMonitor) updateMethodPerformance() {
	cpm.dashboard.MethodPerformance["website_analysis"] = MethodPerformanceData{
		Accuracy:         cpm.metrics.WebsiteAnalysisAccuracy,
		ResponseTime:     cpm.metrics.WebsiteAnalysisResponseTime,
		Confidence:       cpm.metrics.WebsiteAnalysisConfidence,
		SuccessRate:      1.0 - cpm.metrics.WebsiteAnalysisErrorRate,
		ErrorRate:        cpm.metrics.WebsiteAnalysisErrorRate,
		Throughput:       float64(cpm.metrics.WebsiteAnalysisCount),
		UserSatisfaction: cpm.metrics.UserSatisfactionScore,
		LastUpdated:      time.Now(),
	}

	cpm.dashboard.MethodPerformance["web_search"] = MethodPerformanceData{
		Accuracy:         cpm.metrics.WebSearchAccuracy,
		ResponseTime:     cpm.metrics.WebSearchResponseTime,
		Confidence:       cpm.metrics.WebSearchConfidence,
		SuccessRate:      1.0 - cpm.metrics.WebSearchErrorRate,
		ErrorRate:        cpm.metrics.WebSearchErrorRate,
		Throughput:       float64(cpm.metrics.WebSearchCount),
		UserSatisfaction: cpm.metrics.UserSatisfactionScore,
		LastUpdated:      time.Now(),
	}

	cpm.dashboard.MethodPerformance["ml_model"] = MethodPerformanceData{
		Accuracy:         cpm.metrics.MLModelAccuracy,
		ResponseTime:     cpm.metrics.MLModelResponseTime,
		Confidence:       cpm.metrics.MLModelConfidence,
		SuccessRate:      1.0 - cpm.metrics.MLModelErrorRate,
		ErrorRate:        cpm.metrics.MLModelErrorRate,
		Throughput:       float64(cpm.metrics.MLModelCount),
		UserSatisfaction: cpm.metrics.UserSatisfactionScore,
		LastUpdated:      time.Now(),
	}
}

// updateGeographicPerformance updates geographic performance data
func (cpm *ClassificationPerformanceMonitor) updateGeographicPerformance() {
	for region, accuracy := range cpm.metrics.GeographicAccuracy {
		cpm.dashboard.GeographicPerformance[region] = GeographicPerformanceData{
			Accuracy:         accuracy,
			ResponseTime:     cpm.metrics.AverageResponseTime,
			Confidence:       cpm.metrics.AverageConfidenceScore,
			SuccessRate:      cpm.metrics.SuccessRate,
			ErrorRate:        cpm.metrics.ErrorRate,
			Throughput:       cpm.metrics.ClassificationsPerSecond,
			UserSatisfaction: cpm.metrics.UserSatisfactionScore,
			LastUpdated:      time.Now(),
		}
	}
}

// updateIndustryPerformance updates industry performance data
func (cpm *ClassificationPerformanceMonitor) updateIndustryPerformance() {
	for industry, accuracy := range cpm.metrics.IndustryAccuracy {
		cpm.dashboard.IndustryPerformance[industry] = IndustryPerformanceData{
			Accuracy:         accuracy,
			ResponseTime:     cpm.metrics.AverageResponseTime,
			Confidence:       cpm.metrics.AverageConfidenceScore,
			SuccessRate:      cpm.metrics.SuccessRate,
			ErrorRate:        cpm.metrics.ErrorRate,
			Throughput:       cpm.metrics.ClassificationsPerSecond,
			UserSatisfaction: cpm.metrics.UserSatisfactionScore,
			LastUpdated:      time.Now(),
		}
	}
}

// updateOverallHealth updates the overall health status
func (cpm *ClassificationPerformanceMonitor) updateOverallHealth() {
	// Determine overall health based on key metrics
	if cpm.metrics.OverallAccuracy >= 0.9 && cpm.metrics.AverageResponseTime <= 2*time.Second && cpm.metrics.UserSatisfactionScore >= 0.8 {
		cpm.dashboard.OverallHealth = "healthy"
	} else if cpm.metrics.OverallAccuracy >= 0.8 && cpm.metrics.AverageResponseTime <= 5*time.Second && cpm.metrics.UserSatisfactionScore >= 0.7 {
		cpm.dashboard.OverallHealth = "warning"
	} else {
		cpm.dashboard.OverallHealth = "critical"
	}
}

// GetMetrics returns the current performance metrics
func (cpm *ClassificationPerformanceMonitor) GetMetrics() *ClassificationPerformanceMetrics {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()
	return cpm.metrics
}

// GetDashboard returns the current dashboard data
func (cpm *ClassificationPerformanceMonitor) GetDashboard() *ClassificationDashboard {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()
	return cpm.dashboard
}

// GetAlerts returns the current alerts
func (cpm *ClassificationPerformanceMonitor) GetAlerts() []*ClassificationAlert {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	alerts := make([]*ClassificationAlert, 0, len(cpm.alerts.alerts))
	for _, alert := range cpm.alerts.alerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

// GetOptimizations returns the current optimizations
func (cpm *ClassificationPerformanceMonitor) GetOptimizations() []*ClassificationOptimization {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	optimizations := make([]*ClassificationOptimization, 0, len(cpm.optimizer.optimizations))
	for _, optimization := range cpm.optimizer.optimizations {
		optimizations = append(optimizations, optimization)
	}
	return optimizations
}

// GetPredictions returns the current predictions
func (cpm *ClassificationPerformanceMonitor) GetPredictions() []*ClassificationPrediction {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	predictions := make([]*ClassificationPrediction, 0, len(cpm.predictor.predictions))
	for _, prediction := range cpm.predictor.predictions {
		predictions = append(predictions, prediction)
	}
	return predictions
}

// RecordClassification records a classification operation
func (cpm *ClassificationPerformanceMonitor) RecordClassification(method string, success bool, responseTime time.Duration, accuracy, confidence float64) {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	cpm.metrics.TotalClassifications++
	if success {
		cpm.metrics.SuccessfulClassifications++
	} else {
		cpm.metrics.FailedClassifications++
	}

	// Update method-specific counts
	switch method {
	case "website_analysis":
		cpm.metrics.WebsiteAnalysisCount++
		cpm.metrics.WebsiteAnalysisAccuracy = accuracy
		cpm.metrics.WebsiteAnalysisResponseTime = responseTime
		cpm.metrics.WebsiteAnalysisConfidence = confidence
	case "web_search":
		cpm.metrics.WebSearchCount++
		cpm.metrics.WebSearchAccuracy = accuracy
		cpm.metrics.WebSearchResponseTime = responseTime
		cpm.metrics.WebSearchConfidence = confidence
	case "ml_model":
		cpm.metrics.MLModelCount++
		cpm.metrics.MLModelAccuracy = accuracy
		cpm.metrics.MLModelResponseTime = responseTime
		cpm.metrics.MLModelConfidence = confidence
	case "keyword_based":
		cpm.metrics.KeywordBasedCount++
		cpm.metrics.KeywordBasedAccuracy = accuracy
		cpm.metrics.KeywordBasedResponseTime = responseTime
		cpm.metrics.KeywordBasedConfidence = confidence
	case "fuzzy_matching":
		cpm.metrics.FuzzyMatchingCount++
		cpm.metrics.FuzzyMatchingAccuracy = accuracy
		cpm.metrics.FuzzyMatchingResponseTime = responseTime
		cpm.metrics.FuzzyMatchingConfidence = confidence
	case "crosswalk_mapping":
		cpm.metrics.CrosswalkMappingCount++
		cpm.metrics.CrosswalkMappingAccuracy = accuracy
		cpm.metrics.CrosswalkMappingResponseTime = responseTime
		cpm.metrics.CrosswalkMappingConfidence = confidence
	}

	// Update response time metrics
	cpm.updateResponseTimeMetrics(responseTime)

	// Update confidence metrics
	cpm.updateConfidenceMetrics(confidence)
}

// updateResponseTimeMetrics updates response time metrics
func (cpm *ClassificationPerformanceMonitor) updateResponseTimeMetrics(responseTime time.Duration) {
	// Update average response time
	totalTime := cpm.metrics.AverageResponseTime*time.Duration(cpm.metrics.TotalClassifications-1) + responseTime
	cpm.metrics.AverageResponseTime = totalTime / time.Duration(cpm.metrics.TotalClassifications)

	// Update min/max response times
	if responseTime < cpm.metrics.MinResponseTime || cpm.metrics.MinResponseTime == 0 {
		cpm.metrics.MinResponseTime = responseTime
	}
	if responseTime > cpm.metrics.MaxResponseTime {
		cpm.metrics.MaxResponseTime = responseTime
	}
}

// updateConfidenceMetrics updates confidence metrics
func (cpm *ClassificationPerformanceMonitor) updateConfidenceMetrics(confidence float64) {
	// Update average confidence score
	totalConfidence := cpm.metrics.AverageConfidenceScore*float64(cpm.metrics.TotalClassifications-1) + confidence
	cpm.metrics.AverageConfidenceScore = totalConfidence / float64(cpm.metrics.TotalClassifications)

	// Update confidence count buckets
	if confidence >= 0.8 {
		cpm.metrics.HighConfidenceCount++
	} else if confidence >= 0.6 {
		cpm.metrics.MediumConfidenceCount++
	} else {
		cpm.metrics.LowConfidenceCount++
	}
}

// RecordUserFeedback records user feedback
func (cpm *ClassificationPerformanceMonitor) RecordUserFeedback(positive bool) {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	cpm.metrics.FeedbackCount++
	if positive {
		cpm.metrics.PositiveFeedbackCount++
	} else {
		cpm.metrics.NegativeFeedbackCount++
	}
}

// RecordGeographicAccuracy records accuracy for a geographic region
func (cpm *ClassificationPerformanceMonitor) RecordGeographicAccuracy(region string, accuracy float64) {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	cpm.metrics.GeographicAccuracy[region] = accuracy
}

// RecordIndustryAccuracy records accuracy for an industry
func (cpm *ClassificationPerformanceMonitor) RecordIndustryAccuracy(industry string, accuracy float64) {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	cpm.metrics.IndustryAccuracy[industry] = accuracy
}

// RecordResourceUsage records resource usage
func (cpm *ClassificationPerformanceMonitor) RecordResourceUsage(cpuUsage, memoryUsage, diskUsage, networkIO float64) {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	cpm.metrics.CPUUsage = cpuUsage
	cpm.metrics.MemoryUsage = memoryUsage
	cpm.metrics.DiskUsage = diskUsage
	cpm.metrics.NetworkIO = networkIO
}

// RecordCachePerformance records cache performance metrics
func (cpm *ClassificationPerformanceMonitor) RecordCachePerformance(hitRate, missRate float64, cacheSize, evictions int64) {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	cpm.metrics.CacheHitRate = hitRate
	cpm.metrics.CacheMissRate = missRate
	cpm.metrics.CacheSize = cacheSize
	cpm.metrics.CacheEvictions = evictions
}
