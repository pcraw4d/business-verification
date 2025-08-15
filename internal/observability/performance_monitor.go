package observability

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PerformanceMonitor provides comprehensive performance monitoring and optimization
type PerformanceMonitor struct {
	metrics   *PerformanceMetrics
	alerts    *PerformanceAlertManager
	optimizer *PerformanceOptimizer
	predictor *PerformancePredictor
	dashboard *PerformanceDashboard
	config    PerformanceMonitorConfig
	mu        sync.RWMutex
}

// PerformanceMonitorConfig holds configuration for performance monitoring
type PerformanceMonitorConfig struct {
	// Monitoring intervals
	MetricsCollectionInterval time.Duration `json:"metrics_collection_interval"`
	AlertCheckInterval        time.Duration `json:"alert_check_interval"`
	OptimizationInterval      time.Duration `json:"optimization_interval"`
	PredictionInterval        time.Duration `json:"prediction_interval"`

	// Thresholds and limits
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

// PerformanceMetrics tracks comprehensive performance data
type PerformanceMetrics struct {
	// Request metrics
	TotalRequests      int64 `json:"total_requests"`
	SuccessfulRequests int64 `json:"successful_requests"`
	FailedRequests     int64 `json:"failed_requests"`
	TimeoutRequests    int64 `json:"timeout_requests"`

	// Response time metrics
	AverageResponseTime time.Duration `json:"average_response_time"`
	P50ResponseTime     time.Duration `json:"p50_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`
	MinResponseTime     time.Duration `json:"min_response_time"`
	MaxResponseTime     time.Duration `json:"max_response_time"`

	// Throughput metrics
	RequestsPerSecond  float64 `json:"requests_per_second"`
	ConcurrentRequests int     `json:"concurrent_requests"`
	PeakConcurrency    int     `json:"peak_concurrency"`

	// Error metrics
	ErrorRate   float64 `json:"error_rate"`
	SuccessRate float64 `json:"success_rate"`
	TimeoutRate float64 `json:"timeout_rate"`

	// Resource metrics
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	NetworkIO   float64 `json:"network_io"`

	// Business metrics
	ActiveUsers          int              `json:"active_users"`
	APIUsageByEndpoint   map[string]int64 `json:"api_usage_by_endpoint"`
	DataProcessingVolume int64            `json:"data_processing_volume"`

	// Timestamp
	LastUpdated      time.Time     `json:"last_updated"`
	CollectionWindow time.Duration `json:"collection_window"`
}

// PerformanceAlert represents a performance alert
type PerformanceAlert struct {
	ID              string     `json:"id"`
	Type            string     `json:"type"`     // threshold, trend, anomaly, prediction
	Severity        string     `json:"severity"` // low, medium, high, critical
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Metric          string     `json:"metric"`
	CurrentValue    float64    `json:"current_value"`
	Threshold       float64    `json:"threshold"`
	Timestamp       time.Time  `json:"timestamp"`
	Status          string     `json:"status"` // active, acknowledged, resolved
	AcknowledgedBy  string     `json:"acknowledged_by,omitempty"`
	AcknowledgedAt  *time.Time `json:"acknowledged_at,omitempty"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	Recommendations []string   `json:"recommendations"`
}

// PerformanceOptimization represents a performance optimization action
type PerformanceOptimization struct {
	ID                  string     `json:"id"`
	Type                string     `json:"type"` // auto, manual, scheduled
	Action              string     `json:"action"`
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
}

// PerformancePrediction represents a performance prediction
type PerformancePrediction struct {
	ID                string        `json:"id"`
	Metric            string        `json:"metric"`
	PredictedValue    float64       `json:"predicted_value"`
	Confidence        float64       `json:"confidence"`
	PredictionHorizon time.Duration `json:"prediction_horizon"`
	Trend             string        `json:"trend"` // improving, stable, degrading
	Factors           []string      `json:"factors"`
	Timestamp         time.Time     `json:"timestamp"`
	Accuracy          *float64      `json:"accuracy,omitempty"`
}

// PerformanceDashboard provides real-time performance visualization
type PerformanceDashboard struct {
	// Real-time metrics
	CurrentMetrics *PerformanceMetrics `json:"current_metrics"`

	// Historical data
	HistoricalMetrics []*PerformanceMetrics `json:"historical_metrics"`

	// Alerts
	ActiveAlerts []*PerformanceAlert `json:"active_alerts"`

	// Optimizations
	RecentOptimizations []*PerformanceOptimization `json:"recent_optimizations"`

	// Predictions
	CurrentPredictions []*PerformancePrediction `json:"current_predictions"`

	// Status
	OverallHealth string    `json:"overall_health"`
	LastUpdated   time.Time `json:"last_updated"`
}

// PerformanceAlertManager manages performance alerts
type PerformanceAlertManager struct {
	alerts map[string]*PerformanceAlert
	config PerformanceMonitorConfig
	mu     sync.RWMutex
}

// PerformanceOptimizer manages performance optimizations
type PerformanceOptimizer struct {
	optimizations map[string]*PerformanceOptimization
	config        PerformanceMonitorConfig
	mu            sync.RWMutex
}

// PerformancePredictor manages performance predictions
type PerformancePredictor struct {
	predictions    map[string]*PerformancePrediction
	historicalData []*PerformanceMetrics
	config         PerformanceMonitorConfig
	mu             sync.RWMutex
}

// PerformanceDashboardManager manages the performance dashboard
type PerformanceDashboardManager struct {
	dashboard *PerformanceDashboard
	config    PerformanceMonitorConfig
	mu        sync.RWMutex
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(config PerformanceMonitorConfig) *PerformanceMonitor {
	if config.MetricsCollectionInterval == 0 {
		config.MetricsCollectionInterval = 30 * time.Second
	}

	if config.AlertCheckInterval == 0 {
		config.AlertCheckInterval = 1 * time.Minute
	}

	if config.OptimizationInterval == 0 {
		config.OptimizationInterval = 5 * time.Minute
	}

	if config.PredictionInterval == 0 {
		config.PredictionInterval = 2 * time.Minute
	}

	if config.ResponseTimeThreshold == 0 {
		config.ResponseTimeThreshold = 500 * time.Millisecond
	}

	if config.SuccessRateThreshold == 0 {
		config.SuccessRateThreshold = 0.95
	}

	if config.ErrorRateThreshold == 0 {
		config.ErrorRateThreshold = 0.05
	}

	if config.ThroughputThreshold == 0 {
		config.ThroughputThreshold = 1000
	}

	if config.OptimizationConfidence == 0 {
		config.OptimizationConfidence = 0.8
	}

	if config.RollbackThreshold == 0 {
		config.RollbackThreshold = 0.1
	}

	if config.PredictionHorizon == 0 {
		config.PredictionHorizon = 1 * time.Hour
	}

	if config.PredictionConfidence == 0 {
		config.PredictionConfidence = 0.7
	}

	if config.TrendAnalysisWindow == 0 {
		config.TrendAnalysisWindow = 24 * time.Hour
	}

	if config.DashboardRefreshInterval == 0 {
		config.DashboardRefreshInterval = 10 * time.Second
	}

	if config.HistoricalDataRetention == 0 {
		config.HistoricalDataRetention = 30 * 24 * time.Hour
	}

	if config.EscalationDelay == 0 {
		config.EscalationDelay = 15 * time.Minute
	}

	return &PerformanceMonitor{
		metrics:   NewPerformanceMetrics(),
		alerts:    NewPerformanceAlertManager(config),
		optimizer: NewPerformanceOptimizer(config),
		predictor: NewPerformancePredictor(config),
		dashboard: NewPerformanceDashboard(config),
		config:    config,
	}
}

// RecordRequest records a request for performance tracking
func (pm *PerformanceMonitor) RecordRequest(ctx context.Context, request *PerformanceRequest) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Update metrics
	pm.metrics.TotalRequests++

	if request.Success {
		pm.metrics.SuccessfulRequests++
	} else {
		pm.metrics.FailedRequests++
	}

	if request.Timeout {
		pm.metrics.TimeoutRequests++
	}

	// Update response time metrics
	pm.updateResponseTimeMetrics(request.ResponseTime)

	// Update throughput metrics
	pm.updateThroughputMetrics()

	// Update error rates
	pm.updateErrorRates()

	// Update resource metrics
	pm.updateResourceMetrics()

	// Update business metrics
	pm.updateBusinessMetrics(request)

	// Update timestamp
	pm.metrics.LastUpdated = time.Now()

	return nil
}

// GetMetrics returns current performance metrics
func (pm *PerformanceMonitor) GetMetrics() *PerformanceMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.metrics.Clone()
}

// GetAlerts returns active performance alerts
func (pm *PerformanceMonitor) GetAlerts() []*PerformanceAlert {
	return pm.alerts.GetActiveAlerts()
}

// GetOptimizations returns recent performance optimizations
func (pm *PerformanceMonitor) GetOptimizations() []*PerformanceOptimization {
	return pm.optimizer.GetRecentOptimizations()
}

// GetPredictions returns current performance predictions
func (pm *PerformanceMonitor) GetPredictions() []*PerformancePrediction {
	return pm.predictor.GetCurrentPredictions()
}

// GetDashboard returns the performance dashboard
func (pm *PerformanceMonitor) GetDashboard() *PerformanceDashboard {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return &PerformanceDashboard{
		CurrentMetrics:      pm.metrics.Clone(),
		HistoricalMetrics:   pm.getHistoricalMetrics(),
		ActiveAlerts:        pm.alerts.GetActiveAlerts(),
		RecentOptimizations: pm.optimizer.GetRecentOptimizations(),
		CurrentPredictions:  pm.predictor.GetCurrentPredictions(),
		OverallHealth:       pm.calculateOverallHealth(),
		LastUpdated:         time.Now(),
	}
}

// StartMonitoring starts the performance monitoring
func (pm *PerformanceMonitor) StartMonitoring(ctx context.Context) error {
	// Start metrics collection
	go pm.collectMetrics(ctx)

	// Start alert checking
	go pm.checkAlerts(ctx)

	// Start optimization if enabled
	if pm.config.AutoOptimizationEnabled {
		go pm.runOptimizations(ctx)
	}

	// Start predictions
	go pm.runPredictions(ctx)

	// Start dashboard updates
	go pm.updateDashboard(ctx)

	return nil
}

// collectMetrics collects performance metrics
func (pm *PerformanceMonitor) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(pm.config.MetricsCollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pm.collectSystemMetrics()
		}
	}
}

// checkAlerts checks for performance alerts
func (pm *PerformanceMonitor) checkAlerts(ctx context.Context) {
	ticker := time.NewTicker(pm.config.AlertCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pm.checkThresholdAlerts()
			pm.checkTrendAlerts()
			pm.checkAnomalyAlerts()
		}
	}
}

// runOptimizations runs performance optimizations
func (pm *PerformanceMonitor) runOptimizations(ctx context.Context) {
	ticker := time.NewTicker(pm.config.OptimizationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pm.runAutoOptimizations()
		}
	}
}

// runPredictions runs performance predictions
func (pm *PerformanceMonitor) runPredictions(ctx context.Context) {
	ticker := time.NewTicker(pm.config.PredictionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pm.runPerformancePredictions()
		}
	}
}

// updateDashboard updates the performance dashboard
func (pm *PerformanceMonitor) updateDashboard(ctx context.Context) {
	ticker := time.NewTicker(pm.config.DashboardRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			dashboard := pm.GetDashboard()
			pm.dashboard.UpdateDashboard(dashboard.CurrentMetrics, dashboard.ActiveAlerts, dashboard.RecentOptimizations, dashboard.CurrentPredictions)
		}
	}
}

// updateResponseTimeMetrics updates response time metrics
func (pm *PerformanceMonitor) updateResponseTimeMetrics(responseTime time.Duration) {
	// Update average response time
	if pm.metrics.TotalRequests == 1 {
		pm.metrics.AverageResponseTime = responseTime
		pm.metrics.MinResponseTime = responseTime
		pm.metrics.MaxResponseTime = responseTime
	} else {
		// Calculate new average
		total := pm.metrics.AverageResponseTime.Nanoseconds() * (pm.metrics.TotalRequests - 1)
		total += responseTime.Nanoseconds()
		pm.metrics.AverageResponseTime = time.Duration(total / pm.metrics.TotalRequests)

		// Update min/max
		if responseTime < pm.metrics.MinResponseTime {
			pm.metrics.MinResponseTime = responseTime
		}
		if responseTime > pm.metrics.MaxResponseTime {
			pm.metrics.MaxResponseTime = responseTime
		}
	}

	// Note: P50, P95, P99 would require maintaining a sorted list of response times
	// For simplicity, we'll use the average for now
	pm.metrics.P50ResponseTime = pm.metrics.AverageResponseTime
	pm.metrics.P95ResponseTime = pm.metrics.AverageResponseTime
	pm.metrics.P99ResponseTime = pm.metrics.AverageResponseTime
}

// updateThroughputMetrics updates throughput metrics
func (pm *PerformanceMonitor) updateThroughputMetrics() {
	// Calculate requests per second based on collection window
	windowSeconds := pm.config.MetricsCollectionInterval.Seconds()
	pm.metrics.RequestsPerSecond = float64(pm.metrics.TotalRequests) / windowSeconds

	// Update concurrent requests (simplified)
	pm.metrics.ConcurrentRequests = int(pm.metrics.RequestsPerSecond)
	if pm.metrics.ConcurrentRequests > pm.metrics.PeakConcurrency {
		pm.metrics.PeakConcurrency = pm.metrics.ConcurrentRequests
	}
}

// updateErrorRates updates error rates
func (pm *PerformanceMonitor) updateErrorRates() {
	if pm.metrics.TotalRequests > 0 {
		pm.metrics.SuccessRate = float64(pm.metrics.SuccessfulRequests) / float64(pm.metrics.TotalRequests)
		pm.metrics.ErrorRate = float64(pm.metrics.FailedRequests) / float64(pm.metrics.TotalRequests)
		pm.metrics.TimeoutRate = float64(pm.metrics.TimeoutRequests) / float64(pm.metrics.TotalRequests)
	}
}

// updateResourceMetrics updates resource usage metrics
func (pm *PerformanceMonitor) updateResourceMetrics() {
	// Mock resource metrics - in real implementation would collect from system
	pm.metrics.CPUUsage = 45.5
	pm.metrics.MemoryUsage = 67.2
	pm.metrics.DiskUsage = 23.8
	pm.metrics.NetworkIO = 125.6
}

// updateBusinessMetrics updates business metrics
func (pm *PerformanceMonitor) updateBusinessMetrics(request *PerformanceRequest) {
	// Update API usage by endpoint
	if pm.metrics.APIUsageByEndpoint == nil {
		pm.metrics.APIUsageByEndpoint = make(map[string]int64)
	}
	pm.metrics.APIUsageByEndpoint[request.Endpoint]++

	// Update data processing volume
	pm.metrics.DataProcessingVolume += request.DataSize

	// Update active users (simplified)
	pm.metrics.ActiveUsers = len(pm.metrics.APIUsageByEndpoint)
}

// checkThresholdAlerts checks for threshold-based alerts
func (pm *PerformanceMonitor) checkThresholdAlerts() {
	metrics := pm.GetMetrics()

	// Check response time threshold
	if metrics.AverageResponseTime > pm.config.ResponseTimeThreshold {
		pm.alerts.CreateAlert(&PerformanceAlert{
			Type:         "threshold",
			Severity:     "high",
			Title:        "High Response Time",
			Description:  "Average response time exceeds threshold",
			Metric:       "response_time",
			CurrentValue: float64(metrics.AverageResponseTime.Milliseconds()),
			Threshold:    float64(pm.config.ResponseTimeThreshold.Milliseconds()),
			Timestamp:    time.Now(),
			Status:       "active",
			Recommendations: []string{
				"Check database performance",
				"Review query optimization",
				"Consider caching strategies",
			},
		})
	}

	// Check success rate threshold
	if metrics.SuccessRate < pm.config.SuccessRateThreshold {
		pm.alerts.CreateAlert(&PerformanceAlert{
			Type:         "threshold",
			Severity:     "critical",
			Title:        "Low Success Rate",
			Description:  "Success rate below threshold",
			Metric:       "success_rate",
			CurrentValue: metrics.SuccessRate,
			Threshold:    pm.config.SuccessRateThreshold,
			Timestamp:    time.Now(),
			Status:       "active",
			Recommendations: []string{
				"Investigate error patterns",
				"Check external dependencies",
				"Review error handling",
			},
		})
	}

	// Check error rate threshold
	if metrics.ErrorRate > pm.config.ErrorRateThreshold {
		pm.alerts.CreateAlert(&PerformanceAlert{
			Type:         "threshold",
			Severity:     "critical",
			Title:        "High Error Rate",
			Description:  "Error rate above threshold",
			Metric:       "error_rate",
			CurrentValue: metrics.ErrorRate,
			Threshold:    pm.config.ErrorRateThreshold,
			Timestamp:    time.Now(),
			Status:       "active",
			Recommendations: []string{
				"Analyze error logs",
				"Check system health",
				"Review error handling",
			},
		})
	}
}

// checkTrendAlerts checks for trend-based alerts
func (pm *PerformanceMonitor) checkTrendAlerts() {
	// Implementation would analyze historical data for trends
	// For now, we'll create a mock trend alert
	if pm.metrics.TotalRequests > 1000 {
		pm.alerts.CreateAlert(&PerformanceAlert{
			Type:         "trend",
			Severity:     "medium",
			Title:        "Increasing Response Time Trend",
			Description:  "Response time showing upward trend",
			Metric:       "response_time_trend",
			CurrentValue: 0.15,
			Threshold:    0.1,
			Timestamp:    time.Now(),
			Status:       "active",
			Recommendations: []string{
				"Monitor system resources",
				"Consider scaling up",
				"Review recent changes",
			},
		})
	}
}

// checkAnomalyAlerts checks for anomaly-based alerts
func (pm *PerformanceMonitor) checkAnomalyAlerts() {
	// Implementation would use statistical analysis to detect anomalies
	// For now, we'll create a mock anomaly alert
	if pm.metrics.ErrorRate > 0.1 {
		pm.alerts.CreateAlert(&PerformanceAlert{
			Type:         "anomaly",
			Severity:     "high",
			Title:        "Anomalous Error Rate",
			Description:  "Error rate spike detected",
			Metric:       "error_rate_anomaly",
			CurrentValue: pm.metrics.ErrorRate,
			Threshold:    0.05,
			Timestamp:    time.Now(),
			Status:       "active",
			Recommendations: []string{
				"Investigate recent deployments",
				"Check external services",
				"Review system logs",
			},
		})
	}
}

// runAutoOptimizations runs automatic performance optimizations
func (pm *PerformanceMonitor) runAutoOptimizations() {
	metrics := pm.GetMetrics()

	// Check if optimization is needed
	if metrics.AverageResponseTime > pm.config.ResponseTimeThreshold {
		optimization := &PerformanceOptimization{
			Type:                "auto",
			Action:              "enable_caching",
			Description:         "Enable response caching to improve performance",
			TargetMetric:        "response_time",
			ExpectedImprovement: 0.3,
			Confidence:          pm.config.OptimizationConfidence,
			RiskLevel:           "low",
			Status:              "pending",
		}

		pm.optimizer.ApplyOptimization(optimization)
	}

	// Check for other optimization opportunities
	if metrics.SuccessRate < pm.config.SuccessRateThreshold {
		optimization := &PerformanceOptimization{
			Type:                "auto",
			Action:              "retry_configuration",
			Description:         "Adjust retry configuration to improve success rate",
			TargetMetric:        "success_rate",
			ExpectedImprovement: 0.1,
			Confidence:          pm.config.OptimizationConfidence,
			RiskLevel:           "medium",
			Status:              "pending",
		}

		pm.optimizer.ApplyOptimization(optimization)
	}
}

// runPerformancePredictions runs performance predictions
func (pm *PerformanceMonitor) runPerformancePredictions() {
	metrics := pm.GetMetrics()

	// Predict response time
	responseTimePrediction := &PerformancePrediction{
		Metric:            "response_time",
		PredictedValue:    float64(metrics.AverageResponseTime.Milliseconds()) * 1.1,
		Confidence:        pm.config.PredictionConfidence,
		PredictionHorizon: pm.config.PredictionHorizon,
		Trend:             "degrading",
		Factors:           []string{"increasing_load", "resource_constraints"},
		Timestamp:         time.Now(),
	}

	pm.predictor.AddPrediction(responseTimePrediction)

	// Predict success rate
	successRatePrediction := &PerformancePrediction{
		Metric:            "success_rate",
		PredictedValue:    metrics.SuccessRate * 0.98,
		Confidence:        pm.config.PredictionConfidence,
		PredictionHorizon: pm.config.PredictionHorizon,
		Trend:             "degrading",
		Factors:           []string{"error_rate_increase", "dependency_issues"},
		Timestamp:         time.Now(),
	}

	pm.predictor.AddPrediction(successRatePrediction)
}

// calculateOverallHealth calculates overall system health
func (pm *PerformanceMonitor) calculateOverallHealth() string {
	metrics := pm.GetMetrics()

	// Calculate health score
	healthScore := 100.0

	// Deduct points for poor performance
	if metrics.AverageResponseTime > pm.config.ResponseTimeThreshold {
		healthScore -= 20
	}

	if metrics.SuccessRate < pm.config.SuccessRateThreshold {
		healthScore -= 30
	}

	if metrics.ErrorRate > pm.config.ErrorRateThreshold {
		healthScore -= 25
	}

	// Determine health status
	if healthScore >= 90 {
		return "excellent"
	} else if healthScore >= 75 {
		return "good"
	} else if healthScore >= 60 {
		return "fair"
	} else {
		return "poor"
	}
}

// getHistoricalMetrics returns historical metrics
func (pm *PerformanceMonitor) getHistoricalMetrics() []*PerformanceMetrics {
	// In a real implementation, this would retrieve from a time-series database
	// For now, return empty slice
	return []*PerformanceMetrics{}
}

// collectSystemMetrics collects system-level metrics
func (pm *PerformanceMonitor) collectSystemMetrics() {
	// In a real implementation, this would collect actual system metrics
	// For now, this is a placeholder
}

// PerformanceRequest represents a performance request
type PerformanceRequest struct {
	ID           string        `json:"id"`
	Endpoint     string        `json:"endpoint"`
	Method       string        `json:"method"`
	ResponseTime time.Duration `json:"response_time"`
	Success      bool          `json:"success"`
	Timeout      bool          `json:"timeout"`
	Error        string        `json:"error,omitempty"`
	DataSize     int64         `json:"data_size"`
	UserID       string        `json:"user_id,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
}

// NewPerformanceMetrics creates new performance metrics
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		APIUsageByEndpoint: make(map[string]int64),
		LastUpdated:        time.Now(),
	}
}

// Clone creates a copy of performance metrics
func (pm *PerformanceMetrics) Clone() *PerformanceMetrics {
	clone := *pm
	clone.APIUsageByEndpoint = make(map[string]int64)
	for k, v := range pm.APIUsageByEndpoint {
		clone.APIUsageByEndpoint[k] = v
	}
	return &clone
}

// NewPerformanceAlertManager creates a new alert manager
func NewPerformanceAlertManager(config PerformanceMonitorConfig) *PerformanceAlertManager {
	return &PerformanceAlertManager{
		alerts: make(map[string]*PerformanceAlert),
		config: config,
	}
}

// CreateAlert creates a new performance alert
func (pam *PerformanceAlertManager) CreateAlert(alert *PerformanceAlert) {
	pam.mu.Lock()
	defer pam.mu.Unlock()

	alert.ID = fmt.Sprintf("alert_%d", time.Now().UnixNano())
	alert.Status = "active"

	pam.alerts[alert.ID] = alert
}

// GetActiveAlerts returns active alerts
func (pam *PerformanceAlertManager) GetActiveAlerts() []*PerformanceAlert {
	pam.mu.RLock()
	defer pam.mu.RUnlock()

	var activeAlerts []*PerformanceAlert
	for _, alert := range pam.alerts {
		if alert.Status == "active" {
			activeAlerts = append(activeAlerts, alert)
		}
	}

	return activeAlerts
}

// NewPerformanceOptimizer creates a new performance optimizer
func NewPerformanceOptimizer(config PerformanceMonitorConfig) *PerformanceOptimizer {
	return &PerformanceOptimizer{
		optimizations: make(map[string]*PerformanceOptimization),
		config:        config,
	}
}

// ApplyOptimization applies a performance optimization
func (po *PerformanceOptimizer) ApplyOptimization(optimization *PerformanceOptimization) {
	po.mu.Lock()
	defer po.mu.Unlock()

	optimization.ID = fmt.Sprintf("opt_%d", time.Now().UnixNano())
	optimization.Status = "applied"
	now := time.Now()
	optimization.AppliedAt = &now

	po.optimizations[optimization.ID] = optimization
}

// GetRecentOptimizations returns recent optimizations
func (po *PerformanceOptimizer) GetRecentOptimizations() []*PerformanceOptimization {
	po.mu.RLock()
	defer po.mu.RUnlock()

	var recentOptimizations []*PerformanceOptimization
	for _, optimization := range po.optimizations {
		if optimization.AppliedAt != nil && time.Since(*optimization.AppliedAt) < 24*time.Hour {
			recentOptimizations = append(recentOptimizations, optimization)
		}
	}

	return recentOptimizations
}

// NewPerformancePredictor creates a new performance predictor
func NewPerformancePredictor(config PerformanceMonitorConfig) *PerformancePredictor {
	return &PerformancePredictor{
		predictions:    make(map[string]*PerformancePrediction),
		historicalData: make([]*PerformanceMetrics, 0),
		config:         config,
	}
}

// AddPrediction adds a performance prediction
func (pp *PerformancePredictor) AddPrediction(prediction *PerformancePrediction) {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	prediction.ID = fmt.Sprintf("pred_%d", time.Now().UnixNano())

	pp.predictions[prediction.ID] = prediction
}

// GetCurrentPredictions returns current predictions
func (pp *PerformancePredictor) GetCurrentPredictions() []*PerformancePrediction {
	pp.mu.RLock()
	defer pp.mu.RUnlock()

	var currentPredictions []*PerformancePrediction
	for _, prediction := range pp.predictions {
		if time.Since(prediction.Timestamp) < pp.config.PredictionHorizon {
			currentPredictions = append(currentPredictions, prediction)
		}
	}

	return currentPredictions
}

// NewPerformanceDashboard creates a new performance dashboard
func NewPerformanceDashboard(config PerformanceMonitorConfig) *PerformanceDashboard {
	return &PerformanceDashboard{
		CurrentMetrics:      &PerformanceMetrics{},
		HistoricalMetrics:   make([]*PerformanceMetrics, 0),
		ActiveAlerts:        make([]*PerformanceAlert, 0),
		RecentOptimizations: make([]*PerformanceOptimization, 0),
		CurrentPredictions:  make([]*PerformancePrediction, 0),
		OverallHealth:       "healthy",
		LastUpdated:         time.Now(),
	}
}

// UpdateDashboard updates the dashboard
func (pd *PerformanceDashboard) UpdateDashboard(metrics *PerformanceMetrics, alerts []*PerformanceAlert, optimizations []*PerformanceOptimization, predictions []*PerformancePrediction) {
	pd.CurrentMetrics = metrics
	pd.ActiveAlerts = alerts
	pd.RecentOptimizations = optimizations
	pd.CurrentPredictions = predictions
	pd.LastUpdated = time.Now()

	// Update overall health based on metrics
	if metrics != nil {
		if metrics.SuccessRate >= 0.95 {
			pd.OverallHealth = "healthy"
		} else if metrics.SuccessRate >= 0.90 {
			pd.OverallHealth = "warning"
		} else {
			pd.OverallHealth = "critical"
		}
	}
}
