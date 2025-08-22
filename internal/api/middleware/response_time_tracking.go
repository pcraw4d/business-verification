package middleware

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ResponseTimeConfig configures the response time tracking system
type ResponseTimeConfig struct {
	// Tracking settings
	Enabled           bool     `json:"enabled"`
	TrackAllEndpoints bool     `json:"track_all_endpoints"`
	TrackedEndpoints  []string `json:"tracked_endpoints"`
	ExcludedEndpoints []string `json:"excluded_endpoints"`
	SampleRate        float64  `json:"sample_rate"` // 0.0-1.0, percentage of requests to track

	// Thresholds and alerting
	WarningThreshold          time.Duration `json:"warning_threshold"`
	CriticalThreshold         time.Duration `json:"critical_threshold"`
	AlertOnThresholdExceeded  bool          `json:"alert_on_threshold_exceeded"`
	AlertOnPercentileExceeded bool          `json:"alert_on_percentile_exceeded"`

	// Percentile tracking
	TrackPercentiles []float64     `json:"track_percentiles"` // e.g., [50, 90, 95, 99]
	PercentileWindow time.Duration `json:"percentile_window"`

	// Aggregation settings
	AggregationWindow       time.Duration `json:"aggregation_window"`
	MinSamplesForPercentile int           `json:"min_samples_for_percentile"`
	MaxSamplesPerWindow     int           `json:"max_samples_per_window"`

	// Storage and retention
	RetentionPeriod time.Duration `json:"retention_period"`
	CleanupInterval time.Duration `json:"cleanup_interval"`

	// Performance settings
	AsyncProcessing bool `json:"async_processing"`
	BufferSize      int  `json:"buffer_size"`
}

// DefaultResponseTimeConfig returns default configuration
func DefaultResponseTimeConfig() *ResponseTimeConfig {
	return &ResponseTimeConfig{
		Enabled:                   true,
		TrackAllEndpoints:         true,
		SampleRate:                1.0, // Track all requests
		WarningThreshold:          2 * time.Second,
		CriticalThreshold:         5 * time.Second,
		AlertOnThresholdExceeded:  true,
		AlertOnPercentileExceeded: true,
		TrackPercentiles:          []float64{50, 90, 95, 99},
		PercentileWindow:          5 * time.Minute,
		AggregationWindow:         1 * time.Minute,
		MinSamplesForPercentile:   10,
		MaxSamplesPerWindow:       10000,
		RetentionPeriod:           24 * time.Hour,
		CleanupInterval:           10 * time.Minute,
		AsyncProcessing:           true,
		BufferSize:                1000,
	}
}

// ResponseTimeMetric represents a single response time measurement
type ResponseTimeMetric struct {
	Endpoint     string            `json:"endpoint"`
	Method       string            `json:"method"`
	ResponseTime time.Duration     `json:"response_time"`
	StatusCode   int               `json:"status_code"`
	UserID       string            `json:"user_id,omitempty"`
	RequestID    string            `json:"request_id,omitempty"`
	Timestamp    time.Time         `json:"timestamp"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ResponseTimeStats represents aggregated response time statistics
type ResponseTimeStats struct {
	Endpoint           string                    `json:"endpoint"`
	Method             string                    `json:"method"`
	WindowStart        time.Time                 `json:"window_start"`
	WindowEnd          time.Time                 `json:"window_end"`
	SampleCount        int                       `json:"sample_count"`
	MinResponseTime    time.Duration             `json:"min_response_time"`
	MaxResponseTime    time.Duration             `json:"max_response_time"`
	MeanResponseTime   time.Duration             `json:"mean_response_time"`
	MedianResponseTime time.Duration             `json:"median_response_time"`
	Percentiles        map[float64]time.Duration `json:"percentiles"`
	StandardDeviation  time.Duration             `json:"standard_deviation"`
	ErrorCount         int                       `json:"error_count"`
	ErrorRate          float64                   `json:"error_rate"`
	Throughput         float64                   `json:"throughput"` // requests per second
}

// ResponseTimeAlert represents a response time alert
type ResponseTimeAlert struct {
	ID           string            `json:"id"`
	Endpoint     string            `json:"endpoint"`
	Method       string            `json:"method"`
	AlertType    string            `json:"alert_type"` // threshold, percentile, trend
	Severity     string            `json:"severity"`   // warning, critical
	Message      string            `json:"message"`
	CurrentValue time.Duration     `json:"current_value"`
	Threshold    time.Duration     `json:"threshold"`
	Percentile   float64           `json:"percentile,omitempty"`
	TriggeredAt  time.Time         `json:"triggered_at"`
	ResolvedAt   *time.Time        `json:"resolved_at,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ResponseTimeTrend represents trend analysis
type ResponseTimeTrend struct {
	Endpoint       string           `json:"endpoint"`
	Method         string           `json:"method"`
	TrendDirection string           `json:"trend_direction"` // improving, degrading, stable
	TrendStrength  float64          `json:"trend_strength"`  // 0-1.0
	ChangePercent  float64          `json:"change_percent"`
	Period         time.Duration    `json:"period"`
	Confidence     float64          `json:"confidence"`
	DataPoints     []TrendDataPoint `json:"data_points,omitempty"`
	Seasonality    *SeasonalityInfo `json:"seasonality,omitempty"`
	Anomalies      []AnomalyPoint   `json:"anomalies,omitempty"`
}

// TrendDataPoint represents a single data point in trend analysis
type TrendDataPoint struct {
	Timestamp    time.Time     `json:"timestamp"`
	ResponseTime time.Duration `json:"response_time"`
	RequestCount int           `json:"request_count"`
	ErrorRate    float64       `json:"error_rate"`
	Percentile95 time.Duration `json:"percentile_95"`
	Percentile99 time.Duration `json:"percentile_99"`
}

// SeasonalityInfo represents seasonal patterns in the data
type SeasonalityInfo struct {
	HasSeasonality bool          `json:"has_seasonality"`
	Period         time.Duration `json:"period"`
	Strength       float64       `json:"strength"` // 0-1.0
	PeakTimes      []time.Time   `json:"peak_times,omitempty"`
	ValleyTimes    []time.Time   `json:"valley_times,omitempty"`
}

// AnomalyPoint represents an anomaly in the response time data
type AnomalyPoint struct {
	Timestamp    time.Time     `json:"timestamp"`
	ResponseTime time.Duration `json:"response_time"`
	ExpectedTime time.Duration `json:"expected_time"`
	Deviation    float64       `json:"deviation"` // standard deviations from expected
	Severity     string        `json:"severity"`  // low, medium, high, critical
	Description  string        `json:"description"`
}

// TrendAnalysisReport represents a comprehensive trend analysis report
type TrendAnalysisReport struct {
	ID              string                        `json:"id"`
	GeneratedAt     time.Time                     `json:"generated_at"`
	AnalysisPeriod  time.Duration                 `json:"analysis_period"`
	StartTime       time.Time                     `json:"start_time"`
	EndTime         time.Time                     `json:"end_time"`
	OverallTrend    *ResponseTimeTrend            `json:"overall_trend"`
	EndpointTrends  map[string]*ResponseTimeTrend `json:"endpoint_trends"`
	MethodTrends    map[string]*ResponseTimeTrend `json:"method_trends"`
	KeyInsights     []TrendInsight                `json:"key_insights"`
	Recommendations []TrendRecommendation         `json:"recommendations"`
	Anomalies       []AnomalyPoint                `json:"anomalies"`
	Seasonality     map[string]*SeasonalityInfo   `json:"seasonality"`
	Summary         TrendSummary                  `json:"summary"`
}

// TrendInsight represents a key insight from trend analysis
type TrendInsight struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // performance_degradation, improvement, anomaly, seasonality
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"` // low, medium, high, critical
	Confidence  float64   `json:"confidence"`
	Evidence    string    `json:"evidence"`
	Timestamp   time.Time `json:"timestamp"`
}

// TrendRecommendation represents a recommendation based on trend analysis
type TrendRecommendation struct {
	ID                   string    `json:"id"`
	Title                string    `json:"title"`
	Description          string    `json:"description"`
	Category             string    `json:"category"` // optimization, monitoring, alerting, capacity
	Priority             int       `json:"priority"` // 1-10
	Impact               string    `json:"impact"`
	Effort               string    `json:"effort"` // low, medium, high
	EstimatedImprovement float64   `json:"estimated_improvement"`
	Confidence           float64   `json:"confidence"`
	Actions              []string  `json:"actions"`
	CreatedAt            time.Time `json:"created_at"`
}

// TrendSummary provides a high-level summary of trend analysis
type TrendSummary struct {
	TotalEndpoints     int     `json:"total_endpoints"`
	ImprovingEndpoints int     `json:"improving_endpoints"`
	DegradingEndpoints int     `json:"degrading_endpoints"`
	StableEndpoints    int     `json:"stable_endpoints"`
	AverageImprovement float64 `json:"average_improvement"`
	AverageDegradation float64 `json:"average_degradation"`
	AnomalyCount       int     `json:"anomaly_count"`
	SeasonalPatterns   int     `json:"seasonal_patterns"`
	OverallHealth      string  `json:"overall_health"` // excellent, good, fair, poor, critical
}

// TrendAnalysisConfig configures trend analysis behavior
type TrendAnalysisConfig struct {
	// Analysis settings
	MinDataPoints     int           `json:"min_data_points"`    // minimum points for trend analysis
	TrendWindow       time.Duration `json:"trend_window"`       // window for trend calculation
	SeasonalityWindow time.Duration `json:"seasonality_window"` // window for seasonality detection
	AnomalyThreshold  float64       `json:"anomaly_threshold"`  // standard deviations for anomaly detection

	// Reporting settings
	GenerateInsights        bool `json:"generate_insights"`
	GenerateRecommendations bool `json:"generate_recommendations"`
	IncludeSeasonality      bool `json:"include_seasonality"`
	IncludeAnomalies        bool `json:"include_anomalies"`

	// Algorithm settings
	UseLinearRegression     bool `json:"use_linear_regression"`
	UseMovingAverage        bool `json:"use_moving_average"`
	UseExponentialSmoothing bool `json:"use_exponential_smoothing"`

	// Thresholds
	ImprovementThreshold float64 `json:"improvement_threshold"` // % change to consider improvement
	DegradationThreshold float64 `json:"degradation_threshold"` // % change to consider degradation
	StabilityThreshold   float64 `json:"stability_threshold"`   // % change to consider stable
}

// DefaultTrendAnalysisConfig returns default trend analysis configuration
func DefaultTrendAnalysisConfig() *TrendAnalysisConfig {
	return &TrendAnalysisConfig{
		MinDataPoints:           20,
		TrendWindow:             24 * time.Hour,
		SeasonalityWindow:       7 * 24 * time.Hour,
		AnomalyThreshold:        2.5,
		GenerateInsights:        true,
		GenerateRecommendations: true,
		IncludeSeasonality:      true,
		IncludeAnomalies:        true,
		UseLinearRegression:     true,
		UseMovingAverage:        true,
		UseExponentialSmoothing: true,
		ImprovementThreshold:    5.0, // 5% improvement
		DegradationThreshold:    5.0, // 5% degradation
		StabilityThreshold:      2.0, // 2% change considered stable
	}
}

// ResponseTimeOptimizationStrategy defines an optimization strategy
type ResponseTimeOptimizationStrategy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`   // caching, database, connection, algorithm
	Priority    int                    `json:"priority"`   // 1-10, higher is more important
	Impact      string                 `json:"impact"`     // low, medium, high, critical
	Confidence  float64                `json:"confidence"` // 0-1.0
	Actions     []OptimizationAction   `json:"actions"`
	Conditions  map[string]interface{} `json:"conditions"`
	Enabled     bool                   `json:"enabled"`
}

// OptimizationAction represents a specific optimization action
type OptimizationAction struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Type            string                 `json:"type"` // config_change, code_change, resource_adjustment
	Parameters      map[string]interface{} `json:"parameters"`
	EstimatedImpact float64                `json:"estimated_impact"` // expected improvement percentage
	Risk            string                 `json:"risk"`             // low, medium, high
	Rollback        bool                   `json:"rollback"`         // whether action can be rolled back
}

// ResponseTimeOptimizationResult represents the result of a response time optimization attempt
type ResponseTimeOptimizationResult struct {
	ID             string                 `json:"id"`
	StrategyID     string                 `json:"strategy_id"`
	ActionID       string                 `json:"action_id"`
	Status         string                 `json:"status"` // pending, executing, completed, failed, rolled_back
	StartTime      time.Time              `json:"start_time"`
	EndTime        *time.Time             `json:"end_time,omitempty"`
	BeforeMetrics  map[string]interface{} `json:"before_metrics"`
	AfterMetrics   map[string]interface{} `json:"after_metrics,omitempty"`
	Improvement    float64                `json:"improvement,omitempty"` // actual improvement percentage
	Error          string                 `json:"error,omitempty"`
	RollbackReason string                 `json:"rollback_reason,omitempty"`
}

// PerformanceRecommendation represents a performance improvement recommendation
type PerformanceRecommendation struct {
	ID                   string               `json:"id"`
	Title                string               `json:"title"`
	Description          string               `json:"description"`
	Category             string               `json:"category"`
	Priority             int                  `json:"priority"`
	Impact               string               `json:"impact"`
	Confidence           float64              `json:"confidence"`
	Actions              []OptimizationAction `json:"actions"`
	EstimatedImprovement float64              `json:"estimated_improvement"`
	Effort               string               `json:"effort"` // low, medium, high
	CreatedAt            time.Time            `json:"created_at"`
	Status               string               `json:"status"` // new, in_progress, completed, dismissed
}

// ResponseTimeTracker tracks and analyzes response times
type ResponseTimeTracker struct {
	config *ResponseTimeConfig
	logger *zap.Logger

	// Trend analysis configuration
	trendConfig *TrendAnalysisConfig

	// Data storage
	metrics map[string][]*ResponseTimeMetric // key: endpoint_method
	stats   map[string]*ResponseTimeStats    // key: endpoint_method_window
	alerts  map[string]*ResponseTimeAlert    // key: alert_id
	trends  map[string]*ResponseTimeTrend    // key: endpoint_method
	reports map[string]*TrendAnalysisReport  // key: report_id

	// Thread safety
	mutex sync.RWMutex

	// Processing
	stopCh    chan struct{}
	metricsCh chan *ResponseTimeMetric
	workerWg  sync.WaitGroup
}

// NewResponseTimeTracker creates a new response time tracker
func NewResponseTimeTracker(config *ResponseTimeConfig, logger *zap.Logger) *ResponseTimeTracker {
	if config == nil {
		config = DefaultResponseTimeConfig()
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	tracker := &ResponseTimeTracker{
		config:      config,
		logger:      logger,
		trendConfig: DefaultTrendAnalysisConfig(),
		metrics:     make(map[string][]*ResponseTimeMetric),
		stats:       make(map[string]*ResponseTimeStats),
		alerts:      make(map[string]*ResponseTimeAlert),
		trends:      make(map[string]*ResponseTimeTrend),
		reports:     make(map[string]*TrendAnalysisReport),
		stopCh:      make(chan struct{}),
		metricsCh:   make(chan *ResponseTimeMetric, config.BufferSize),
	}

	if config.AsyncProcessing {
		tracker.startWorker()
	}

	// Start cleanup goroutine
	go tracker.cleanupWorker()

	return tracker
}

// TrackResponseTime records a response time measurement
func (rtt *ResponseTimeTracker) TrackResponseTime(ctx context.Context, metric *ResponseTimeMetric) error {
	if !rtt.config.Enabled {
		return nil
	}

	if metric == nil {
		return errors.New("metric cannot be nil")
	}

	// Check if endpoint should be tracked
	if !rtt.shouldTrackEndpoint(metric.Endpoint) {
		return nil
	}

	// Apply sample rate
	if rtt.config.SampleRate < 1.0 && !rtt.shouldSample() {
		return nil
	}

	// Set timestamp if not provided
	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}

	if rtt.config.AsyncProcessing {
		select {
		case rtt.metricsCh <- metric:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Channel full, log warning and process synchronously
			rtt.logger.Warn("response time metrics channel full, processing synchronously")
			return rtt.processMetric(metric)
		}
	}

	return rtt.processMetric(metric)
}

// GetResponseTimeStats retrieves response time statistics for an endpoint
func (rtt *ResponseTimeTracker) GetResponseTimeStats(ctx context.Context, endpoint, method string, window time.Duration) (*ResponseTimeStats, error) {
	if endpoint == "" {
		return nil, errors.New("endpoint is required")
	}

	rtt.mutex.RLock()
	defer rtt.mutex.RUnlock()

	key := rtt.makeStatsKey(endpoint, method, window)
	stats, exists := rtt.stats[key]
	if !exists {
		return nil, fmt.Errorf("no stats found for endpoint %s method %s", endpoint, method)
	}

	return stats, nil
}

// GetResponseTimePercentile retrieves a specific percentile for an endpoint
func (rtt *ResponseTimeTracker) GetResponseTimePercentile(ctx context.Context, endpoint, method string, percentile float64) (time.Duration, error) {
	if percentile < 0 || percentile > 100 {
		return 0, errors.New("percentile must be between 0 and 100")
	}

	stats, err := rtt.GetResponseTimeStats(ctx, endpoint, method, rtt.config.AggregationWindow)
	if err != nil {
		return 0, err
	}

	value, exists := stats.Percentiles[percentile]
	if !exists {
		return 0, fmt.Errorf("percentile %.1f not available", percentile)
	}

	return value, nil
}

// GetResponseTimeTrend retrieves trend analysis for an endpoint
func (rtt *ResponseTimeTracker) GetResponseTimeTrend(ctx context.Context, endpoint, method string) (*ResponseTimeTrend, error) {
	if endpoint == "" {
		return nil, errors.New("endpoint is required")
	}

	rtt.mutex.RLock()
	defer rtt.mutex.RUnlock()

	key := fmt.Sprintf("%s_%s", endpoint, method)
	trend, exists := rtt.trends[key]
	if !exists {
		return nil, fmt.Errorf("no trend data found for endpoint %s method %s", endpoint, method)
	}

	return trend, nil
}

// GetActiveAlerts retrieves all active response time alerts
func (rtt *ResponseTimeTracker) GetActiveAlerts(ctx context.Context) []*ResponseTimeAlert {
	rtt.mutex.RLock()
	defer rtt.mutex.RUnlock()

	var activeAlerts []*ResponseTimeAlert
	for _, alert := range rtt.alerts {
		if alert.ResolvedAt == nil {
			activeAlerts = append(activeAlerts, alert)
		}
	}

	return activeAlerts
}

// GetEndpointPerformance retrieves comprehensive performance data for an endpoint
func (rtt *ResponseTimeTracker) GetEndpointPerformance(ctx context.Context, endpoint, method string) (*EndpointPerformance, error) {
	if endpoint == "" {
		return nil, errors.New("endpoint is required")
	}

	// Get current stats
	currentStats, err := rtt.GetResponseTimeStats(ctx, endpoint, method, rtt.config.AggregationWindow)
	if err != nil {
		return nil, err
	}

	// Get trend data
	trend, _ := rtt.GetResponseTimeTrend(ctx, endpoint, method)

	// Get active alerts
	alerts := rtt.GetActiveAlerts(ctx)
	var endpointAlerts []*ResponseTimeAlert
	for _, alert := range alerts {
		if alert.Endpoint == endpoint && alert.Method == method {
			endpointAlerts = append(endpointAlerts, alert)
		}
	}

	return &EndpointPerformance{
		Endpoint:     endpoint,
		Method:       method,
		CurrentStats: currentStats,
		Trend:        trend,
		ActiveAlerts: endpointAlerts,
		LastUpdated:  time.Now(),
	}, nil
}

// ListTrackedEndpoints returns all tracked endpoints
func (rtt *ResponseTimeTracker) ListTrackedEndpoints(ctx context.Context) []string {
	rtt.mutex.RLock()
	defer rtt.mutex.RUnlock()

	endpoints := make(map[string]bool)
	for key := range rtt.metrics {
		endpoint, _, _ := rtt.parseKey(key)
		endpoints[endpoint] = true
	}

	result := make([]string, 0, len(endpoints))
	for endpoint := range endpoints {
		result = append(result, endpoint)
	}

	sort.Strings(result)
	return result
}

// GetSlowestEndpoints returns the slowest endpoints based on P95 response time
func (rtt *ResponseTimeTracker) GetSlowestEndpoints(ctx context.Context, limit int) ([]*EndpointPerformance, error) {
	if limit <= 0 {
		limit = 10
	}

	endpoints := rtt.ListTrackedEndpoints(ctx)
	var performances []*EndpointPerformance

	for _, endpoint := range endpoints {
		// Get performance for each method
		methods := rtt.getMethodsForEndpoint(endpoint)
		for _, method := range methods {
			performance, err := rtt.GetEndpointPerformance(ctx, endpoint, method)
			if err != nil {
				continue
			}
			performances = append(performances, performance)
		}
	}

	// Sort by P95 response time (descending)
	sort.Slice(performances, func(i, j int) bool {
		p95i, _ := performances[i].CurrentStats.Percentiles[95]
		p95j, _ := performances[j].CurrentStats.Percentiles[95]
		return p95i > p95j
	})

	if len(performances) > limit {
		performances = performances[:limit]
	}

	return performances, nil
}

// Cleanup removes old data based on retention policy
func (rtt *ResponseTimeTracker) Cleanup() error {
	rtt.mutex.Lock()
	defer rtt.mutex.Unlock()

	cutoff := time.Now().Add(-rtt.config.RetentionPeriod)

	// Cleanup old metrics
	metricsRemoved := 0
	for key, metrics := range rtt.metrics {
		var filtered []*ResponseTimeMetric
		for _, metric := range metrics {
			if metric.Timestamp.After(cutoff) {
				filtered = append(filtered, metric)
			}
		}
		if len(filtered) == 0 {
			delete(rtt.metrics, key)
		} else {
			rtt.metrics[key] = filtered
		}
		metricsRemoved += len(metrics) - len(filtered)
	}

	// Cleanup old stats
	statsRemoved := 0
	for key, stats := range rtt.stats {
		if stats.WindowEnd.Before(cutoff) {
			delete(rtt.stats, key)
			statsRemoved++
		}
	}

	// Cleanup resolved alerts
	alertsRemoved := 0
	for key, alert := range rtt.alerts {
		if alert.ResolvedAt != nil && alert.ResolvedAt.Before(cutoff) {
			delete(rtt.alerts, key)
			alertsRemoved++
		}
	}

	rtt.logger.Info("response time tracker cleanup completed",
		zap.Int("metrics_removed", metricsRemoved),
		zap.Int("stats_removed", statsRemoved),
		zap.Int("alerts_removed", alertsRemoved))

	return nil
}

// Shutdown gracefully shuts down the tracker
func (rtt *ResponseTimeTracker) Shutdown() error {
	close(rtt.stopCh)

	if rtt.config.AsyncProcessing {
		rtt.workerWg.Wait()
	}

	rtt.logger.Info("response time tracker shut down")
	return nil
}

// ResolveAlert marks an alert as resolved
func (rtt *ResponseTimeTracker) ResolveAlert(ctx context.Context, alertID string) error {
	rtt.mutex.Lock()
	defer rtt.mutex.Unlock()

	alert, exists := rtt.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert %s not found", alertID)
	}

	if alert.ResolvedAt != nil {
		return fmt.Errorf("alert %s already resolved", alertID)
	}

	now := time.Now()
	alert.ResolvedAt = &now

	rtt.logger.Info("response time alert resolved",
		zap.String("alert_id", alertID),
		zap.String("endpoint", alert.Endpoint),
		zap.String("method", alert.Method),
		zap.String("severity", alert.Severity),
		zap.Duration("response_time", alert.CurrentValue),
		zap.Duration("threshold", alert.Threshold))

	return nil
}

// UpdateThresholds updates the warning and critical thresholds
func (rtt *ResponseTimeTracker) UpdateThresholds(ctx context.Context, warning, critical time.Duration) error {
	if warning <= 0 || critical <= 0 {
		return errors.New("thresholds must be positive")
	}
	if warning >= critical {
		return errors.New("warning threshold must be less than critical threshold")
	}

	rtt.mutex.Lock()
	defer rtt.mutex.Unlock()

	oldWarning := rtt.config.WarningThreshold
	oldCritical := rtt.config.CriticalThreshold

	rtt.config.WarningThreshold = warning
	rtt.config.CriticalThreshold = critical

	rtt.logger.Info("response time thresholds updated",
		zap.Duration("old_warning", oldWarning),
		zap.Duration("new_warning", warning),
		zap.Duration("old_critical", oldCritical),
		zap.Duration("new_critical", critical))

	return nil
}

// GetThresholds returns the current warning and critical thresholds
func (rtt *ResponseTimeTracker) GetThresholds(ctx context.Context) (warning, critical time.Duration) {
	rtt.mutex.RLock()
	defer rtt.mutex.RUnlock()

	return rtt.config.WarningThreshold, rtt.config.CriticalThreshold
}

// GetAlertHistory returns alert history with optional filtering
func (rtt *ResponseTimeTracker) GetAlertHistory(ctx context.Context, filters map[string]interface{}) []*ResponseTimeAlert {
	rtt.mutex.RLock()
	defer rtt.mutex.RUnlock()

	var filteredAlerts []*ResponseTimeAlert

	for _, alert := range rtt.alerts {
		if rtt.matchesAlertFilters(alert, filters) {
			filteredAlerts = append(filteredAlerts, alert)
		}
	}

	// Sort by triggered time (newest first)
	sort.Slice(filteredAlerts, func(i, j int) bool {
		return filteredAlerts[i].TriggeredAt.After(filteredAlerts[j].TriggeredAt)
	})

	return filteredAlerts
}

// matchesAlertFilters checks if an alert matches the given filters
func (rtt *ResponseTimeTracker) matchesAlertFilters(alert *ResponseTimeAlert, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "endpoint":
			if endpoint, ok := value.(string); ok && alert.Endpoint != endpoint {
				return false
			}
		case "method":
			if method, ok := value.(string); ok && alert.Method != method {
				return false
			}
		case "severity":
			if severity, ok := value.(string); ok && alert.Severity != severity {
				return false
			}
		case "alert_type":
			if alertType, ok := value.(string); ok && alert.AlertType != alertType {
				return false
			}
		case "resolved":
			if resolved, ok := value.(bool); ok && (alert.ResolvedAt != nil) != resolved {
				return false
			}
		case "since":
			if since, ok := value.(time.Time); ok && alert.TriggeredAt.Before(since) {
				return false
			}
		case "until":
			if until, ok := value.(time.Time); ok && alert.TriggeredAt.After(until) {
				return false
			}
		}
	}
	return true
}

// GetAlertStatistics returns statistics about alerts
func (rtt *ResponseTimeTracker) GetAlertStatistics(ctx context.Context) map[string]interface{} {
	rtt.mutex.RLock()
	defer rtt.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_alerts":      len(rtt.alerts),
		"active_alerts":     0,
		"resolved_alerts":   0,
		"critical_alerts":   0,
		"warning_alerts":    0,
		"threshold_alerts":  0,
		"percentile_alerts": 0,
		"endpoints":         make(map[string]int),
		"methods":           make(map[string]int),
	}

	for _, alert := range rtt.alerts {
		if alert.ResolvedAt == nil {
			stats["active_alerts"] = stats["active_alerts"].(int) + 1
		} else {
			stats["resolved_alerts"] = stats["resolved_alerts"].(int) + 1
		}

		switch alert.Severity {
		case "critical":
			stats["critical_alerts"] = stats["critical_alerts"].(int) + 1
		case "warning":
			stats["warning_alerts"] = stats["warning_alerts"].(int) + 1
		}

		switch alert.AlertType {
		case "threshold":
			stats["threshold_alerts"] = stats["threshold_alerts"].(int) + 1
		case "percentile":
			stats["percentile_alerts"] = stats["percentile_alerts"].(int) + 1
		}

		// Count by endpoint
		endpoints := stats["endpoints"].(map[string]int)
		endpoints[alert.Endpoint]++

		// Count by method
		methods := stats["methods"].(map[string]int)
		methods[alert.Method]++
	}

	return stats
}

// CheckThresholdViolations checks for current threshold violations across all endpoints
func (rtt *ResponseTimeTracker) CheckThresholdViolations(ctx context.Context) []*ResponseTimeAlert {
	rtt.mutex.RLock()
	defer rtt.mutex.RUnlock()

	var violations []*ResponseTimeAlert

	// Check all current stats for threshold violations
	for key, stats := range rtt.stats {
		if stats.SampleCount < rtt.config.MinSamplesForPercentile {
			continue
		}

		endpoint, method, _ := rtt.parseKey(key)

		// Check P95 against thresholds
		if p95, exists := stats.Percentiles[95]; exists {
			if p95 >= rtt.config.CriticalThreshold {
				violations = append(violations, &ResponseTimeAlert{
					ID:           fmt.Sprintf("violation_%s_%s_critical_%d", endpoint, method, time.Now().Unix()),
					Endpoint:     endpoint,
					Method:       method,
					AlertType:    "threshold_violation",
					Severity:     "critical",
					Message:      fmt.Sprintf("P95 response time %s exceeds critical threshold %s", p95, rtt.config.CriticalThreshold),
					CurrentValue: p95,
					Threshold:    rtt.config.CriticalThreshold,
					TriggeredAt:  time.Now(),
				})
			} else if p95 >= rtt.config.WarningThreshold {
				violations = append(violations, &ResponseTimeAlert{
					ID:           fmt.Sprintf("violation_%s_%s_warning_%d", endpoint, method, time.Now().Unix()),
					Endpoint:     endpoint,
					Method:       method,
					AlertType:    "threshold_violation",
					Severity:     "warning",
					Message:      fmt.Sprintf("P95 response time %s exceeds warning threshold %s", p95, rtt.config.WarningThreshold),
					CurrentValue: p95,
					Threshold:    rtt.config.WarningThreshold,
					TriggeredAt:  time.Now(),
				})
			}
		}
	}

	return violations
}

// GetThresholdViolationSummary returns a summary of current threshold violations
func (rtt *ResponseTimeTracker) GetThresholdViolationSummary(ctx context.Context) map[string]interface{} {
	violations := rtt.CheckThresholdViolations(ctx)

	summary := map[string]interface{}{
		"total_violations": len(violations),
		"critical_count":   0,
		"warning_count":    0,
		"endpoints":        make(map[string]map[string]interface{}),
	}

	for _, violation := range violations {
		if violation.Severity == "critical" {
			summary["critical_count"] = summary["critical_count"].(int) + 1
		} else {
			summary["warning_count"] = summary["warning_count"].(int) + 1
		}

		// Group by endpoint
		endpoints := summary["endpoints"].(map[string]map[string]interface{})
		if _, exists := endpoints[violation.Endpoint]; !exists {
			endpoints[violation.Endpoint] = map[string]interface{}{
				"critical_violations": 0,
				"warning_violations":  0,
				"methods":             make(map[string]interface{}),
			}
		}

		endpointData := endpoints[violation.Endpoint]
		if violation.Severity == "critical" {
			endpointData["critical_violations"] = endpointData["critical_violations"].(int) + 1
		} else {
			endpointData["warning_violations"] = endpointData["warning_violations"].(int) + 1
		}

		// Group by method
		methods := endpointData["methods"].(map[string]interface{})
		if _, exists := methods[violation.Method]; !exists {
			methods[violation.Method] = map[string]interface{}{
				"current_p95": violation.CurrentValue.String(),
				"threshold":   violation.Threshold.String(),
				"severity":    violation.Severity,
			}
		}
	}

	return summary
}

// Internal methods

func (rtt *ResponseTimeTracker) startWorker() {
	rtt.workerWg.Add(1)
	go func() {
		defer rtt.workerWg.Done()
		for {
			select {
			case metric := <-rtt.metricsCh:
				if err := rtt.processMetric(metric); err != nil {
					rtt.logger.Error("failed to process metric", zap.Error(err))
				}
			case <-rtt.stopCh:
				return
			}
		}
	}()
}

func (rtt *ResponseTimeTracker) cleanupWorker() {
	ticker := time.NewTicker(rtt.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := rtt.Cleanup(); err != nil {
				rtt.logger.Error("cleanup failed", zap.Error(err))
			}
		case <-rtt.stopCh:
			return
		}
	}
}

func (rtt *ResponseTimeTracker) processMetric(metric *ResponseTimeMetric) error {
	rtt.mutex.Lock()
	defer rtt.mutex.Unlock()

	// Store metric
	key := rtt.makeMetricsKey(metric.Endpoint, metric.Method)
	rtt.metrics[key] = append(rtt.metrics[key], metric)

	// Limit samples per window
	if len(rtt.metrics[key]) > rtt.config.MaxSamplesPerWindow {
		rtt.metrics[key] = rtt.metrics[key][len(rtt.metrics[key])-rtt.config.MaxSamplesPerWindow:]
	}

	// Update stats
	rtt.updateStats(metric)

	// Check for alerts
	rtt.checkAlerts(metric)

	// Update trends
	rtt.updateTrends(metric)

	return nil
}

func (rtt *ResponseTimeTracker) updateStats(metric *ResponseTimeMetric) {
	key := rtt.makeStatsKey(metric.Endpoint, metric.Method, rtt.config.AggregationWindow)

	// Get or create stats
	stats, exists := rtt.stats[key]
	if !exists {
		now := time.Now()
		windowStart := now.Truncate(rtt.config.AggregationWindow)
		stats = &ResponseTimeStats{
			Endpoint:    metric.Endpoint,
			Method:      metric.Method,
			WindowStart: windowStart,
			WindowEnd:   windowStart.Add(rtt.config.AggregationWindow),
			Percentiles: make(map[float64]time.Duration),
		}
		rtt.stats[key] = stats
	}

	// Update basic stats
	stats.SampleCount++
	if metric.ResponseTime < stats.MinResponseTime || stats.MinResponseTime == 0 {
		stats.MinResponseTime = metric.ResponseTime
	}
	if metric.ResponseTime > stats.MaxResponseTime {
		stats.MaxResponseTime = metric.ResponseTime
	}

	// Update error count
	if metric.StatusCode >= 400 {
		stats.ErrorCount++
	}

	// Calculate percentiles periodically
	if stats.SampleCount >= rtt.config.MinSamplesForPercentile {
		rtt.calculatePercentiles(stats)
	}
}

func (rtt *ResponseTimeTracker) calculatePercentiles(stats *ResponseTimeStats) {
	// Get all response times for this endpoint/method
	key := rtt.makeMetricsKey(stats.Endpoint, stats.Method)
	metrics, exists := rtt.metrics[key]
	if !exists {
		return
	}

	// Filter metrics within the stats window
	var responseTimes []time.Duration
	for _, metric := range metrics {
		if metric.Timestamp.After(stats.WindowStart) && metric.Timestamp.Before(stats.WindowEnd) {
			responseTimes = append(responseTimes, metric.ResponseTime)
		}
	}

	if len(responseTimes) < rtt.config.MinSamplesForPercentile {
		return
	}

	// Sort response times
	sort.Slice(responseTimes, func(i, j int) bool {
		return responseTimes[i] < responseTimes[j]
	})

	// Calculate mean and median
	var total time.Duration
	for _, rt := range responseTimes {
		total += rt
	}
	stats.MeanResponseTime = total / time.Duration(len(responseTimes))
	stats.MedianResponseTime = responseTimes[len(responseTimes)/2]

	// Calculate percentiles
	for _, percentile := range rtt.config.TrackPercentiles {
		index := int(float64(len(responseTimes)-1) * percentile / 100.0)
		if index >= 0 && index < len(responseTimes) {
			stats.Percentiles[percentile] = responseTimes[index]
		}
	}

	// Calculate standard deviation
	var variance float64
	mean := float64(stats.MeanResponseTime)
	for _, rt := range responseTimes {
		diff := float64(rt) - mean
		variance += diff * diff
	}
	variance /= float64(len(responseTimes))
	stats.StandardDeviation = time.Duration(math.Sqrt(variance))

	// Calculate error rate and throughput
	stats.ErrorRate = float64(stats.ErrorCount) / float64(stats.SampleCount) * 100.0
	windowDuration := stats.WindowEnd.Sub(stats.WindowStart).Seconds()
	stats.Throughput = float64(stats.SampleCount) / windowDuration
}

func (rtt *ResponseTimeTracker) checkAlerts(metric *ResponseTimeMetric) {
	// Check threshold alerts
	if rtt.config.AlertOnThresholdExceeded {
		if metric.ResponseTime >= rtt.config.CriticalThreshold {
			rtt.createAlert(metric, "threshold", "critical", rtt.config.CriticalThreshold, 0)
		} else if metric.ResponseTime >= rtt.config.WarningThreshold {
			rtt.createAlert(metric, "threshold", "warning", rtt.config.WarningThreshold, 0)
		}
	}

	// Check percentile alerts
	if rtt.config.AlertOnPercentileExceeded {
		stats, exists := rtt.stats[rtt.makeStatsKey(metric.Endpoint, metric.Method, rtt.config.PercentileWindow)]
		if exists {
			for _, percentile := range rtt.config.TrackPercentiles {
				if percentileValue, ok := stats.Percentiles[percentile]; ok {
					if metric.ResponseTime >= percentileValue {
						rtt.createAlert(metric, "percentile", "warning", percentileValue, percentile)
					}
				}
			}
		}
	}
}

func (rtt *ResponseTimeTracker) createAlert(metric *ResponseTimeMetric, alertType, severity string, threshold time.Duration, percentile float64) {
	alertID := fmt.Sprintf("rt_alert_%s_%s_%s_%d", metric.Endpoint, metric.Method, alertType, time.Now().UnixNano())

	message := fmt.Sprintf("Response time %s exceeded %s threshold", metric.ResponseTime, threshold)
	if percentile > 0 {
		message = fmt.Sprintf("Response time %s exceeded P%.0f threshold (%s)", metric.ResponseTime, percentile, threshold)
	}

	alert := &ResponseTimeAlert{
		ID:           alertID,
		Endpoint:     metric.Endpoint,
		Method:       metric.Method,
		AlertType:    alertType,
		Severity:     severity,
		Message:      message,
		CurrentValue: metric.ResponseTime,
		Threshold:    threshold,
		Percentile:   percentile,
		TriggeredAt:  time.Now(),
		Metadata:     metric.Metadata,
	}

	rtt.alerts[alertID] = alert

	rtt.logger.Warn("response time alert triggered",
		zap.String("alert_id", alertID),
		zap.String("endpoint", metric.Endpoint),
		zap.String("method", metric.Method),
		zap.String("severity", severity),
		zap.Duration("response_time", metric.ResponseTime),
		zap.Duration("threshold", threshold))
}

func (rtt *ResponseTimeTracker) updateTrends(metric *ResponseTimeMetric) {
	// Simple trend calculation - compare current P95 with previous window
	key := fmt.Sprintf("%s_%s", metric.Endpoint, metric.Method)

	currentStats, exists := rtt.stats[rtt.makeStatsKey(metric.Endpoint, metric.Method, rtt.config.PercentileWindow)]
	if !exists {
		return
	}

	if _, exists := currentStats.Percentiles[95]; !exists {
		return
	}

	trend, exists := rtt.trends[key]
	if !exists {
		trend = &ResponseTimeTrend{
			Endpoint:       metric.Endpoint,
			Method:         metric.Method,
			TrendDirection: "stable",
			TrendStrength:  0.0,
			ChangePercent:  0.0,
			Period:         rtt.config.PercentileWindow,
			Confidence:     0.5,
		}
		rtt.trends[key] = trend
	}

	// Update trend based on P95 changes
	// This is a simplified trend calculation
	// In a real implementation, you'd use more sophisticated trend analysis
	if trend.ChangePercent > 10 {
		trend.TrendDirection = "degrading"
		trend.TrendStrength = math.Min(1.0, trend.ChangePercent/50.0)
	} else if trend.ChangePercent < -10 {
		trend.TrendDirection = "improving"
		trend.TrendStrength = math.Min(1.0, math.Abs(trend.ChangePercent)/50.0)
	} else {
		trend.TrendDirection = "stable"
		trend.TrendStrength = 0.0
	}
}

func (rtt *ResponseTimeTracker) shouldTrackEndpoint(endpoint string) bool {
	if rtt.config.TrackAllEndpoints {
		// Check exclusions
		for _, excluded := range rtt.config.ExcludedEndpoints {
			if endpoint == excluded {
				return false
			}
		}
		return true
	}

	// Check inclusions
	for _, included := range rtt.config.TrackedEndpoints {
		if endpoint == included {
			return true
		}
	}
	return false
}

func (rtt *ResponseTimeTracker) shouldSample() bool {
	return rtt.config.SampleRate >= 1.0 || (rtt.config.SampleRate > 0 && rand.Float64() < rtt.config.SampleRate)
}

func (rtt *ResponseTimeTracker) makeMetricsKey(endpoint, method string) string {
	return fmt.Sprintf("%s_%s", endpoint, method)
}

func (rtt *ResponseTimeTracker) makeStatsKey(endpoint, method string, window time.Duration) string {
	windowStart := time.Now().Truncate(window)
	return fmt.Sprintf("%s_%s_%d", endpoint, method, windowStart.Unix())
}

func (rtt *ResponseTimeTracker) parseKey(key string) (endpoint, method string, window int64) {
	// Parse key in format "endpoint_method" or "endpoint_method_timestamp"
	parts := strings.Split(key, "_")
	if len(parts) < 2 {
		return "", "", 0
	}

	// For metrics key: "endpoint_method"
	if len(parts) == 2 {
		return parts[0], parts[1], 0
	}

	// For stats key: "endpoint_method_timestamp"
	if len(parts) >= 3 {
		endpoint := parts[0]
		method := parts[1]
		// Reconstruct method if it contains underscores
		if len(parts) > 3 {
			method = strings.Join(parts[1:len(parts)-1], "_")
		}
		return endpoint, method, 0
	}

	return "", "", 0
}

func (rtt *ResponseTimeTracker) getMethodsForEndpoint(endpoint string) []string {
	rtt.mutex.RLock()
	defer rtt.mutex.RUnlock()

	methods := make(map[string]bool)
	for key := range rtt.metrics {
		if parsedEndpoint, method, _ := rtt.parseKey(key); parsedEndpoint == endpoint {
			methods[method] = true
		}
	}

	result := make([]string, 0, len(methods))
	for method := range methods {
		result = append(result, method)
	}
	return result
}

// EndpointPerformance represents comprehensive performance data for an endpoint
type EndpointPerformance struct {
	Endpoint     string               `json:"endpoint"`
	Method       string               `json:"method"`
	CurrentStats *ResponseTimeStats   `json:"current_stats"`
	Trend        *ResponseTimeTrend   `json:"trend,omitempty"`
	ActiveAlerts []*ResponseTimeAlert `json:"active_alerts"`
	LastUpdated  time.Time            `json:"last_updated"`
}

// ResponseTimeOptimizer handles response time optimization and tuning
type ResponseTimeOptimizer struct {
	tracker *ResponseTimeTracker
	logger  *zap.Logger
	config  *OptimizationConfig

	// Optimization state
	strategies      map[string]*ResponseTimeOptimizationStrategy
	results         map[string]*ResponseTimeOptimizationResult
	recommendations map[string]*PerformanceRecommendation

	// Thread safety
	mutex sync.RWMutex

	// Processing
	stopCh   chan struct{}
	workerWg sync.WaitGroup
}

// OptimizationConfig configures the optimization behavior
type OptimizationConfig struct {
	Enabled                 bool          `json:"enabled"`
	AutoOptimizationEnabled bool          `json:"auto_optimization_enabled"`
	OptimizationInterval    time.Duration `json:"optimization_interval"`
	MinImprovementThreshold float64       `json:"min_improvement_threshold"`
	MaxOptimizationAttempts int           `json:"max_optimization_attempts"`
	RollbackThreshold       float64       `json:"rollback_threshold"`
	ConfidenceThreshold     float64       `json:"confidence_threshold"`
	EnableRollback          bool          `json:"enable_rollback"`
	EnableLearning          bool          `json:"enable_learning"`
	LearningWindow          time.Duration `json:"learning_window"`
}

// DefaultOptimizationConfig returns default optimization configuration
func DefaultOptimizationConfig() *OptimizationConfig {
	return &OptimizationConfig{
		Enabled:                 true,
		AutoOptimizationEnabled: false, // Disabled by default for safety
		OptimizationInterval:    5 * time.Minute,
		MinImprovementThreshold: 5.0, // 5% minimum improvement
		MaxOptimizationAttempts: 10,
		RollbackThreshold:       -10.0, // Rollback if 10% degradation
		ConfidenceThreshold:     0.7,
		EnableRollback:          true,
		EnableLearning:          true,
		LearningWindow:          24 * time.Hour,
	}
}

// NewResponseTimeOptimizer creates a new response time optimizer
func NewResponseTimeOptimizer(tracker *ResponseTimeTracker, config *OptimizationConfig, logger *zap.Logger) *ResponseTimeOptimizer {
	if config == nil {
		config = DefaultOptimizationConfig()
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	optimizer := &ResponseTimeOptimizer{
		tracker:         tracker,
		logger:          logger,
		config:          config,
		strategies:      make(map[string]*ResponseTimeOptimizationStrategy),
		results:         make(map[string]*ResponseTimeOptimizationResult),
		recommendations: make(map[string]*PerformanceRecommendation),
		stopCh:          make(chan struct{}),
	}

	// Initialize default strategies
	optimizer.initializeDefaultStrategies()

	// Start optimization worker if auto-optimization is enabled
	if config.AutoOptimizationEnabled {
		optimizer.startOptimizationWorker()
	}

	return optimizer
}

// initializeDefaultStrategies initializes default optimization strategies
func (rto *ResponseTimeOptimizer) initializeDefaultStrategies() {
	// Cache optimization strategy
	rto.strategies["cache_optimization"] = &ResponseTimeOptimizationStrategy{
		ID:          "cache_optimization",
		Name:        "Cache Optimization",
		Description: "Optimize caching behavior to improve response times",
		Category:    "caching",
		Priority:    8,
		Impact:      "high",
		Confidence:  0.8,
		Enabled:     true,
		Actions: []OptimizationAction{
			{
				ID:              "increase_cache_size",
				Name:            "Increase Cache Size",
				Description:     "Increase cache size to reduce cache misses",
				Type:            "config_change",
				Parameters:      map[string]interface{}{"cache_size_multiplier": 1.5},
				EstimatedImpact: 15.0,
				Risk:            "low",
				Rollback:        true,
			},
			{
				ID:              "optimize_cache_ttl",
				Name:            "Optimize Cache TTL",
				Description:     "Adjust cache TTL based on access patterns",
				Type:            "config_change",
				Parameters:      map[string]interface{}{"ttl_multiplier": 1.2},
				EstimatedImpact: 10.0,
				Risk:            "low",
				Rollback:        true,
			},
		},
		Conditions: map[string]interface{}{
			"p95_response_time": 1000, // ms
			"cache_hit_rate":    0.8,  // 80%
		},
	}

	// Database optimization strategy
	rto.strategies["database_optimization"] = &ResponseTimeOptimizationStrategy{
		ID:          "database_optimization",
		Name:        "Database Optimization",
		Description: "Optimize database queries and connections",
		Category:    "database",
		Priority:    9,
		Impact:      "high",
		Confidence:  0.7,
		Enabled:     true,
		Actions: []OptimizationAction{
			{
				ID:              "increase_connection_pool",
				Name:            "Increase Connection Pool",
				Description:     "Increase database connection pool size",
				Type:            "config_change",
				Parameters:      map[string]interface{}{"pool_size_multiplier": 1.3},
				EstimatedImpact: 20.0,
				Risk:            "medium",
				Rollback:        true,
			},
			{
				ID:              "optimize_queries",
				Name:            "Optimize Database Queries",
				Description:     "Add database indexes and optimize slow queries",
				Type:            "code_change",
				Parameters:      map[string]interface{}{"query_timeout": 5000},
				EstimatedImpact: 25.0,
				Risk:            "high",
				Rollback:        false,
			},
		},
		Conditions: map[string]interface{}{
			"p95_response_time":    2000, // ms
			"database_connections": 0.8,  // 80% utilization
		},
	}

	// Connection optimization strategy
	rto.strategies["connection_optimization"] = &ResponseTimeOptimizationStrategy{
		ID:          "connection_optimization",
		Name:        "Connection Optimization",
		Description: "Optimize HTTP connection pooling and timeouts",
		Category:    "connection",
		Priority:    7,
		Impact:      "medium",
		Confidence:  0.6,
		Enabled:     true,
		Actions: []OptimizationAction{
			{
				ID:              "increase_http_pool",
				Name:            "Increase HTTP Connection Pool",
				Description:     "Increase HTTP client connection pool size",
				Type:            "config_change",
				Parameters:      map[string]interface{}{"max_idle_conns": 100, "max_conns_per_host": 10},
				EstimatedImpact: 12.0,
				Risk:            "low",
				Rollback:        true,
			},
			{
				ID:              "optimize_timeouts",
				Name:            "Optimize Request Timeouts",
				Description:     "Adjust request timeouts based on response patterns",
				Type:            "config_change",
				Parameters:      map[string]interface{}{"timeout_multiplier": 1.2},
				EstimatedImpact: 8.0,
				Risk:            "low",
				Rollback:        true,
			},
		},
		Conditions: map[string]interface{}{
			"p95_response_time": 1500, // ms
			"connection_errors": 0.05, // 5% error rate
		},
	}

	// Algorithm optimization strategy
	rto.strategies["algorithm_optimization"] = &ResponseTimeOptimizationStrategy{
		ID:          "algorithm_optimization",
		Name:        "Algorithm Optimization",
		Description: "Optimize processing algorithms and data structures",
		Category:    "algorithm",
		Priority:    6,
		Impact:      "medium",
		Confidence:  0.5,
		Enabled:     true,
		Actions: []OptimizationAction{
			{
				ID:              "parallel_processing",
				Name:            "Enable Parallel Processing",
				Description:     "Enable parallel processing for independent operations",
				Type:            "config_change",
				Parameters:      map[string]interface{}{"parallel_workers": 4},
				EstimatedImpact: 18.0,
				Risk:            "medium",
				Rollback:        true,
			},
			{
				ID:              "optimize_data_structures",
				Name:            "Optimize Data Structures",
				Description:     "Use more efficient data structures for processing",
				Type:            "code_change",
				Parameters:      map[string]interface{}{"use_maps": true, "preallocate_slices": true},
				EstimatedImpact: 15.0,
				Risk:            "high",
				Rollback:        false,
			},
		},
		Conditions: map[string]interface{}{
			"p95_response_time": 3000, // ms
			"cpu_usage":         0.7,  // 70% CPU usage
		},
	}
}

// AnalyzePerformance analyzes current performance and generates recommendations
func (rto *ResponseTimeOptimizer) AnalyzePerformance(ctx context.Context) ([]*PerformanceRecommendation, error) {
	rto.mutex.Lock()
	defer rto.mutex.Unlock()

	var recommendations []*PerformanceRecommendation

	// Get current performance metrics
	metrics := rto.getCurrentMetrics()

	// Analyze each strategy
	for _, strategy := range rto.strategies {
		if !strategy.Enabled {
			continue
		}

		// Check if strategy conditions are met
		if rto.shouldApplyStrategy(strategy, metrics) {
			recommendation := rto.createRecommendation(strategy, metrics)
			recommendations = append(recommendations, recommendation)
			rto.recommendations[recommendation.ID] = recommendation
		}
	}

	// Sort recommendations by priority and impact
	sort.Slice(recommendations, func(i, j int) bool {
		if recommendations[i].Priority != recommendations[j].Priority {
			return recommendations[i].Priority > recommendations[j].Priority
		}
		return recommendations[i].EstimatedImprovement > recommendations[j].EstimatedImprovement
	})

	rto.logger.Info("performance analysis completed",
		zap.Int("recommendations_count", len(recommendations)))

	return recommendations, nil
}

// getCurrentMetrics collects current performance metrics
func (rto *ResponseTimeOptimizer) getCurrentMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Get response time statistics for all endpoints
	endpoints := rto.tracker.ListTrackedEndpoints(context.Background())
	var p95ResponseTimes []time.Duration
	var totalRequests int64

	for _, endpoint := range endpoints {
		methods := rto.tracker.getMethodsForEndpoint(endpoint)
		for _, method := range methods {
			stats, err := rto.tracker.GetResponseTimeStats(context.Background(), endpoint, method, rto.tracker.config.AggregationWindow)
			if err == nil && stats != nil {
				if p95, exists := stats.Percentiles[95]; exists {
					p95ResponseTimes = append(p95ResponseTimes, p95)
				}
				totalRequests += int64(stats.SampleCount)
			}
		}
	}

	// Calculate aggregate metrics
	if len(p95ResponseTimes) > 0 {
		sort.Slice(p95ResponseTimes, func(i, j int) bool {
			return p95ResponseTimes[i] < p95ResponseTimes[j]
		})

		// Use median P95 as overall P95
		medianIndex := len(p95ResponseTimes) / 2
		metrics["p95_response_time"] = float64(p95ResponseTimes[medianIndex].Milliseconds())
	}

	metrics["total_requests"] = totalRequests
	metrics["active_endpoints"] = len(endpoints)

	// Add mock metrics for demonstration (in real implementation, these would come from system monitoring)
	metrics["cache_hit_rate"] = 0.85
	metrics["database_connections"] = 0.6
	metrics["connection_errors"] = 0.02
	metrics["cpu_usage"] = 0.45

	return metrics
}

// shouldApplyStrategy checks if a strategy should be applied based on current metrics
func (rto *ResponseTimeOptimizer) shouldApplyStrategy(strategy *ResponseTimeOptimizationStrategy, metrics map[string]interface{}) bool {
	for condition, threshold := range strategy.Conditions {
		if value, exists := metrics[condition]; exists {
			switch v := value.(type) {
			case float64:
				if thresholdFloat, ok := threshold.(float64); ok {
					if v < thresholdFloat {
						return false // Condition not met
					}
				}
			case int64:
				if thresholdInt, ok := threshold.(int); ok {
					if v < int64(thresholdInt) {
						return false // Condition not met
					}
				}
			}
		}
	}
	return true
}

// createRecommendation creates a performance recommendation from a strategy
func (rto *ResponseTimeOptimizer) createRecommendation(strategy *ResponseTimeOptimizationStrategy, metrics map[string]interface{}) *PerformanceRecommendation {
	recommendationID := fmt.Sprintf("rec_%s_%d", strategy.ID, time.Now().Unix())

	// Calculate estimated improvement based on actions
	var totalImprovement float64
	for _, action := range strategy.Actions {
		totalImprovement += action.EstimatedImpact
	}

	// Cap improvement at 50% to be realistic
	if totalImprovement > 50.0 {
		totalImprovement = 50.0
	}

	return &PerformanceRecommendation{
		ID:                   recommendationID,
		Title:                strategy.Name,
		Description:          strategy.Description,
		Category:             strategy.Category,
		Priority:             strategy.Priority,
		Impact:               strategy.Impact,
		Confidence:           strategy.Confidence,
		Actions:              strategy.Actions,
		EstimatedImprovement: totalImprovement,
		Effort:               rto.calculateEffort(strategy.Actions),
		CreatedAt:            time.Now(),
		Status:               "new",
	}
}

// calculateEffort calculates the effort level for a set of actions
func (rto *ResponseTimeOptimizer) calculateEffort(actions []OptimizationAction) string {
	var totalRisk int
	for _, action := range actions {
		switch action.Risk {
		case "low":
			totalRisk += 1
		case "medium":
			totalRisk += 2
		case "high":
			totalRisk += 3
		}
	}

	avgRisk := float64(totalRisk) / float64(len(actions))
	if avgRisk <= 1.5 {
		return "low"
	} else if avgRisk <= 2.5 {
		return "medium"
	}
	return "high"
}

// ExecuteOptimization executes a specific optimization action
func (rto *ResponseTimeOptimizer) ExecuteOptimization(ctx context.Context, strategyID, actionID string) (*ResponseTimeOptimizationResult, error) {
	rto.mutex.Lock()
	defer rto.mutex.Unlock()

	// Find strategy and action
	strategy, exists := rto.strategies[strategyID]
	if !exists {
		return nil, fmt.Errorf("strategy %s not found", strategyID)
	}

	var targetAction *OptimizationAction
	for _, action := range strategy.Actions {
		if action.ID == actionID {
			targetAction = &action
			break
		}
	}

	if targetAction == nil {
		return nil, fmt.Errorf("action %s not found in strategy %s", actionID, strategyID)
	}

	// Create optimization result
	resultID := fmt.Sprintf("opt_%s_%s_%d", strategyID, actionID, time.Now().Unix())
	result := &ResponseTimeOptimizationResult{
		ID:            resultID,
		StrategyID:    strategyID,
		ActionID:      actionID,
		Status:        "pending",
		StartTime:     time.Now(),
		BeforeMetrics: rto.getCurrentMetrics(),
	}

	rto.results[resultID] = result

	// Execute optimization in background
	go rto.executeOptimizationAction(ctx, result, targetAction)

	return result, nil
}

// executeOptimizationAction executes an optimization action
func (rto *ResponseTimeOptimizer) executeOptimizationAction(ctx context.Context, result *ResponseTimeOptimizationResult, action *OptimizationAction) {
	rto.mutex.Lock()
	result.Status = "executing"
	rto.mutex.Unlock()

	rto.logger.Info("executing optimization action",
		zap.String("result_id", result.ID),
		zap.String("action_id", action.ID),
		zap.String("action_name", action.Name))

	// Simulate action execution
	time.Sleep(2 * time.Second) // Simulate processing time

	// Check if context was cancelled
	select {
	case <-ctx.Done():
		rto.mutex.Lock()
		result.Status = "failed"
		result.Error = "context cancelled"
		rto.mutex.Unlock()
		return
	default:
	}

	// Get metrics after optimization
	rto.mutex.Lock()
	result.Status = "completed"
	now := time.Now()
	result.EndTime = &now
	result.AfterMetrics = rto.getCurrentMetrics()
	rto.mutex.Unlock()

	// Calculate improvement
	improvement := rto.calculateImprovement(result)
	result.Improvement = improvement

	// Check if rollback is needed
	if rto.config.EnableRollback && improvement < rto.config.RollbackThreshold {
		rto.logger.Warn("optimization resulted in performance degradation, rolling back",
			zap.String("result_id", result.ID),
			zap.Float64("improvement", improvement),
			zap.Float64("rollback_threshold", rto.config.RollbackThreshold))

		rto.rollbackOptimization(result, action)
	}

	rto.logger.Info("optimization action completed",
		zap.String("result_id", result.ID),
		zap.String("action_id", action.ID),
		zap.Float64("improvement", improvement))
}

// calculateImprovement calculates the performance improvement
func (rto *ResponseTimeOptimizer) calculateImprovement(result *ResponseTimeOptimizationResult) float64 {
	beforeP95, beforeExists := result.BeforeMetrics["p95_response_time"].(float64)
	afterP95, afterExists := result.AfterMetrics["p95_response_time"].(float64)

	if !beforeExists || !afterExists || beforeP95 == 0 {
		return 0.0
	}

	// Calculate improvement percentage (positive = improvement, negative = degradation)
	improvement := ((beforeP95 - afterP95) / beforeP95) * 100.0
	return improvement
}

// rollbackOptimization rolls back an optimization
func (rto *ResponseTimeOptimizer) rollbackOptimization(result *ResponseTimeOptimizationResult, action *OptimizationAction) {
	if !action.Rollback {
		rto.mutex.Lock()
		result.Status = "failed"
		result.RollbackReason = "action does not support rollback"
		rto.mutex.Unlock()
		return
	}

	rto.mutex.Lock()
	result.Status = "rolled_back"
	result.RollbackReason = "performance degradation detected"
	rto.mutex.Unlock()

	rto.logger.Info("optimization rolled back",
		zap.String("result_id", result.ID),
		zap.String("action_id", action.ID))
}

// GetOptimizationResults returns optimization results with optional filtering
func (rto *ResponseTimeOptimizer) GetOptimizationResults(ctx context.Context, filters map[string]interface{}) []*ResponseTimeOptimizationResult {
	rto.mutex.RLock()
	defer rto.mutex.RUnlock()

	var filteredResults []*ResponseTimeOptimizationResult

	for _, result := range rto.results {
		if rto.matchesResultFilters(result, filters) {
			filteredResults = append(filteredResults, result)
		}
	}

	// Sort by start time (newest first)
	sort.Slice(filteredResults, func(i, j int) bool {
		return filteredResults[i].StartTime.After(filteredResults[j].StartTime)
	})

	return filteredResults
}

// matchesResultFilters checks if a result matches the given filters
func (rto *ResponseTimeOptimizer) matchesResultFilters(result *ResponseTimeOptimizationResult, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "strategy_id":
			if strategyID, ok := value.(string); ok && result.StrategyID != strategyID {
				return false
			}
		case "action_id":
			if actionID, ok := value.(string); ok && result.ActionID != actionID {
				return false
			}
		case "status":
			if status, ok := value.(string); ok && result.Status != status {
				return false
			}
		case "since":
			if since, ok := value.(time.Time); ok && result.StartTime.Before(since) {
				return false
			}
		case "until":
			if until, ok := value.(time.Time); ok && result.StartTime.After(until) {
				return false
			}
		}
	}
	return true
}

// GetOptimizationStatistics returns statistics about optimizations
func (rto *ResponseTimeOptimizer) GetOptimizationStatistics(ctx context.Context) map[string]interface{} {
	rto.mutex.RLock()
	defer rto.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_optimizations": len(rto.results),
		"completed":           0,
		"failed":              0,
		"rolled_back":         0,
		"pending":             0,
		"executing":           0,
		"strategies":          make(map[string]int),
		"average_improvement": 0.0,
		"success_rate":        0.0,
	}

	var totalImprovement float64
	var completedCount int

	for _, result := range rto.results {
		switch result.Status {
		case "completed":
			stats["completed"] = stats["completed"].(int) + 1
			completedCount++
			if result.Improvement > 0 {
				totalImprovement += result.Improvement
			}
		case "failed":
			stats["failed"] = stats["failed"].(int) + 1
		case "rolled_back":
			stats["rolled_back"] = stats["rolled_back"].(int) + 1
		case "pending":
			stats["pending"] = stats["pending"].(int) + 1
		case "executing":
			stats["executing"] = stats["executing"].(int) + 1
		}

		// Count by strategy
		strategies := stats["strategies"].(map[string]int)
		strategies[result.StrategyID]++
	}

	// Calculate averages
	if completedCount > 0 {
		stats["average_improvement"] = totalImprovement / float64(completedCount)
		stats["success_rate"] = float64(completedCount) / float64(len(rto.results)) * 100.0
	}

	return stats
}

// startOptimizationWorker starts the automatic optimization worker
func (rto *ResponseTimeOptimizer) startOptimizationWorker() {
	rto.workerWg.Add(1)
	go func() {
		defer rto.workerWg.Done()
		ticker := time.NewTicker(rto.config.OptimizationInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := rto.runAutomaticOptimization(context.Background()); err != nil {
					rto.logger.Error("automatic optimization failed", zap.Error(err))
				}
			case <-rto.stopCh:
				return
			}
		}
	}()
}

// runAutomaticOptimization runs automatic optimization
func (rto *ResponseTimeOptimizer) runAutomaticOptimization(ctx context.Context) error {
	// Get recommendations
	recommendations, err := rto.AnalyzePerformance(ctx)
	if err != nil {
		return fmt.Errorf("failed to analyze performance: %w", err)
	}

	// Execute high-priority recommendations
	for _, recommendation := range recommendations {
		if recommendation.Priority >= 8 && recommendation.Confidence >= rto.config.ConfidenceThreshold {
			rto.logger.Info("executing automatic optimization",
				zap.String("recommendation_id", recommendation.ID),
				zap.String("title", recommendation.Title),
				zap.Int("priority", recommendation.Priority))

			// Execute the first action of the recommendation
			if len(recommendation.Actions) > 0 {
				action := recommendation.Actions[0]
				_, err := rto.ExecuteOptimization(ctx, recommendation.Category, action.ID)
				if err != nil {
					rto.logger.Error("failed to execute automatic optimization",
						zap.String("recommendation_id", recommendation.ID),
						zap.Error(err))
				}
			}
		}
	}

	return nil
}

// Shutdown gracefully shuts down the optimizer
func (rto *ResponseTimeOptimizer) Shutdown() error {
	close(rto.stopCh)
	rto.workerWg.Wait()
	return nil
}

// =============================================================================
// TREND ANALYSIS AND REPORTING METHODS
// =============================================================================

// GenerateTrendAnalysisReport creates a comprehensive trend analysis report
func (rtt *ResponseTimeTracker) GenerateTrendAnalysisReport(ctx context.Context, startTime, endTime time.Time) (*TrendAnalysisReport, error) {
	rtt.mutex.RLock()
	defer rtt.mutex.RUnlock()

	report := &TrendAnalysisReport{
		ID:              generateTrendID(),
		GeneratedAt:     time.Now(),
		AnalysisPeriod:  endTime.Sub(startTime),
		StartTime:       startTime,
		EndTime:         endTime,
		EndpointTrends:  make(map[string]*ResponseTimeTrend),
		MethodTrends:    make(map[string]*ResponseTimeTrend),
		KeyInsights:     make([]TrendInsight, 0),
		Recommendations: make([]TrendRecommendation, 0),
		Anomalies:       make([]AnomalyPoint, 0),
		Seasonality:     make(map[string]*SeasonalityInfo),
	}

	// Generate overall trend
	overallTrend, err := rtt.calculateOverallTrend(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate overall trend: %w", err)
	}
	report.OverallTrend = overallTrend

	// Generate endpoint-specific trends
	for key, metrics := range rtt.metrics {
		if len(metrics) < rtt.trendConfig.MinDataPoints {
			continue
		}

		endpoint, method := rtt.ParseEndpointKey(key)
		trend, err := rtt.calculateEndpointTrend(ctx, endpoint, method, startTime, endTime)
		if err != nil {
			rtt.logger.Warn("failed to calculate trend for endpoint",
				zap.String("endpoint", endpoint),
				zap.String("method", method),
				zap.Error(err))
			continue
		}

		report.EndpointTrends[key] = trend
		report.MethodTrends[method] = trend
	}

	// Generate insights and recommendations
	if rtt.trendConfig.GenerateInsights {
		insights, err := rtt.generateTrendInsights(ctx, report)
		if err != nil {
			rtt.logger.Warn("failed to generate insights", zap.Error(err))
		} else {
			report.KeyInsights = insights
		}
	}

	if rtt.trendConfig.GenerateRecommendations {
		recommendations, err := rtt.generateTrendRecommendations(ctx, report)
		if err != nil {
			rtt.logger.Warn("failed to generate recommendations", zap.Error(err))
		} else {
			report.Recommendations = recommendations
		}
	}

	// Detect anomalies
	if rtt.trendConfig.IncludeAnomalies {
		anomalies, err := rtt.DetectAnomalies(ctx, startTime, endTime)
		if err != nil {
			rtt.logger.Warn("failed to detect anomalies", zap.Error(err))
		} else {
			report.Anomalies = anomalies
		}
	}

	// Analyze seasonality
	if rtt.trendConfig.IncludeSeasonality {
		seasonality, err := rtt.AnalyzeSeasonality(ctx, startTime, endTime)
		if err != nil {
			rtt.logger.Warn("failed to analyze seasonality", zap.Error(err))
		} else {
			report.Seasonality = seasonality
		}
	}

	// Generate summary
	report.Summary = rtt.GenerateTrendSummary(report)

	// Store report
	rtt.reports[report.ID] = report

	return report, nil
}

// calculateOverallTrend calculates the overall trend across all endpoints
func (rtt *ResponseTimeTracker) calculateOverallTrend(ctx context.Context, startTime, endTime time.Time) (*ResponseTimeTrend, error) {
	var allDataPoints []TrendDataPoint

	// Collect data points from all endpoints
	for _, metrics := range rtt.metrics {
		dataPoints, err := rtt.extractDataPoints(metrics, startTime, endTime)
		if err != nil {
			continue
		}
		allDataPoints = append(allDataPoints, dataPoints...)
	}

	if len(allDataPoints) < rtt.trendConfig.MinDataPoints {
		return nil, fmt.Errorf("insufficient data points for trend analysis: %d < %d",
			len(allDataPoints), rtt.trendConfig.MinDataPoints)
	}

	return rtt.calculateTrendFromDataPoints("overall", "all", allDataPoints)
}

// calculateEndpointTrend calculates trend for a specific endpoint
func (rtt *ResponseTimeTracker) calculateEndpointTrend(ctx context.Context, endpoint, method string, startTime, endTime time.Time) (*ResponseTimeTrend, error) {
	key := rtt.BuildEndpointKey(endpoint, method)
	metrics, exists := rtt.metrics[key]
	if !exists {
		return nil, fmt.Errorf("no metrics found for endpoint: %s %s", method, endpoint)
	}

	dataPoints, err := rtt.extractDataPoints(metrics, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to extract data points: %w", err)
	}

	if len(dataPoints) < rtt.trendConfig.MinDataPoints {
		return nil, fmt.Errorf("insufficient data points for trend analysis: %d < %d",
			len(dataPoints), rtt.trendConfig.MinDataPoints)
	}

	return rtt.calculateTrendFromDataPoints(endpoint, method, dataPoints)
}

// extractDataPoints extracts trend data points from metrics
func (rtt *ResponseTimeTracker) extractDataPoints(metrics []*ResponseTimeMetric, startTime, endTime time.Time) ([]TrendDataPoint, error) {
	var dataPoints []TrendDataPoint
	windowSize := rtt.trendConfig.TrendWindow / time.Duration(rtt.trendConfig.MinDataPoints)

	// Group metrics by time windows
	windowMetrics := make(map[time.Time][]*ResponseTimeMetric)
	for _, metric := range metrics {
		if metric.Timestamp.Before(startTime) || metric.Timestamp.After(endTime) {
			continue
		}

		windowStart := metric.Timestamp.Truncate(windowSize)
		windowMetrics[windowStart] = append(windowMetrics[windowStart], metric)
	}

	// Calculate aggregated data points for each window
	for windowStart, windowData := range windowMetrics {
		if len(windowData) == 0 {
			continue
		}

		dataPoint := rtt.calculateDataPointFromMetrics(windowData)
		dataPoint.Timestamp = windowStart
		dataPoints = append(dataPoints, dataPoint)
	}

	// Sort by timestamp
	sort.Slice(dataPoints, func(i, j int) bool {
		return dataPoints[i].Timestamp.Before(dataPoints[j].Timestamp)
	})

	return dataPoints, nil
}

// calculateDataPointFromMetrics calculates a single data point from a group of metrics
func (rtt *ResponseTimeTracker) calculateDataPointFromMetrics(metrics []*ResponseTimeMetric) TrendDataPoint {
	if len(metrics) == 0 {
		return TrendDataPoint{}
	}

	var responseTimes []time.Duration
	var errorCount int
	var totalRequests int

	for _, metric := range metrics {
		responseTimes = append(responseTimes, metric.ResponseTime)
		totalRequests++
		if metric.StatusCode >= 400 {
			errorCount++
		}
	}

	// Calculate percentiles
	sort.Slice(responseTimes, func(i, j int) bool {
		return responseTimes[i] < responseTimes[j]
	})

	percentile95 := rtt.calculatePercentile(responseTimes, 0.95)
	percentile99 := rtt.calculatePercentile(responseTimes, 0.99)

	// Calculate average response time
	var totalDuration time.Duration
	for _, rt := range responseTimes {
		totalDuration += rt
	}
	averageResponseTime := totalDuration / time.Duration(len(responseTimes))

	// Calculate error rate
	errorRate := 0.0
	if totalRequests > 0 {
		errorRate = float64(errorCount) / float64(totalRequests) * 100.0
	}

	return TrendDataPoint{
		ResponseTime: averageResponseTime,
		RequestCount: totalRequests,
		ErrorRate:    errorRate,
		Percentile95: percentile95,
		Percentile99: percentile99,
	}
}

// calculateTrendFromDataPoints calculates trend from data points using multiple algorithms
func (rtt *ResponseTimeTracker) calculateTrendFromDataPoints(endpoint, method string, dataPoints []TrendDataPoint) (*ResponseTimeTrend, error) {
	if len(dataPoints) < 2 {
		return nil, fmt.Errorf("insufficient data points for trend calculation")
	}

	trend := &ResponseTimeTrend{
		Endpoint:   endpoint,
		Method:     method,
		DataPoints: dataPoints,
		Period:     dataPoints[len(dataPoints)-1].Timestamp.Sub(dataPoints[0].Timestamp),
	}

	// Calculate trend using multiple algorithms
	var trends []float64
	var confidences []float64

	if rtt.trendConfig.UseLinearRegression {
		slope, confidence := rtt.calculateLinearRegressionTrend(dataPoints)
		trends = append(trends, slope)
		confidences = append(confidences, confidence)
	}

	if rtt.trendConfig.UseMovingAverage {
		slope, confidence := rtt.calculateMovingAverageTrend(dataPoints)
		trends = append(trends, slope)
		confidences = append(confidences, confidence)
	}

	if rtt.trendConfig.UseExponentialSmoothing {
		slope, confidence := rtt.calculateExponentialSmoothingTrend(dataPoints)
		trends = append(trends, slope)
		confidences = append(confidences, confidence)
	}

	// Combine results
	if len(trends) > 0 {
		trend.TrendStrength = rtt.calculateAverageTrendStrength(trends, confidences)
		trend.ChangePercent = rtt.calculateChangePercent(dataPoints)
		trend.TrendDirection = rtt.determineTrendDirection(trend.ChangePercent)
		trend.Confidence = rtt.calculateAverageConfidence(confidences)
	}

	return trend, nil
}

// calculateLinearRegressionTrend calculates trend using linear regression
func (rtt *ResponseTimeTracker) calculateLinearRegressionTrend(dataPoints []TrendDataPoint) (float64, float64) {
	if len(dataPoints) < 2 {
		return 0.0, 0.0
	}

	var sumX, sumY, sumXY, sumX2 float64
	n := float64(len(dataPoints))

	for i, point := range dataPoints {
		x := float64(i)
		y := float64(point.ResponseTime.Milliseconds())

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope (trend)
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	// Calculate R-squared (confidence)
	meanY := sumY / n
	var ssRes, ssTot float64
	for i, point := range dataPoints {
		x := float64(i)
		y := float64(point.ResponseTime.Milliseconds())
		yPred := slope*x + (sumY/n - slope*sumX/n)
		ssRes += (y - yPred) * (y - yPred)
		ssTot += (y - meanY) * (y - meanY)
	}

	confidence := 0.0
	if ssTot > 0 {
		confidence = 1 - (ssRes / ssTot)
	}

	return slope, confidence
}

// calculateMovingAverageTrend calculates trend using moving average
func (rtt *ResponseTimeTracker) calculateMovingAverageTrend(dataPoints []TrendDataPoint) (float64, float64) {
	if len(dataPoints) < 3 {
		return 0.0, 0.0
	}

	windowSize := 3
	if len(dataPoints) < windowSize*2 {
		windowSize = len(dataPoints) / 2
	}

	// Calculate moving averages
	var movingAverages []float64
	for i := windowSize - 1; i < len(dataPoints); i++ {
		var sum time.Duration
		for j := i - windowSize + 1; j <= i; j++ {
			sum += dataPoints[j].ResponseTime
		}
		movingAverages = append(movingAverages, float64(sum.Milliseconds())/float64(windowSize))
	}

	// Calculate trend from moving averages
	if len(movingAverages) < 2 {
		return 0.0, 0.0
	}

	slope := (movingAverages[len(movingAverages)-1] - movingAverages[0]) / float64(len(movingAverages)-1)

	// Calculate confidence based on consistency
	var variance float64
	mean := (movingAverages[0] + movingAverages[len(movingAverages)-1]) / 2
	for _, ma := range movingAverages {
		variance += (ma - mean) * (ma - mean)
	}
	variance /= float64(len(movingAverages))

	confidence := 1.0 / (1.0 + variance/1000.0) // Normalize to 0-1

	return slope, confidence
}

// calculateExponentialSmoothingTrend calculates trend using exponential smoothing
func (rtt *ResponseTimeTracker) calculateExponentialSmoothingTrend(dataPoints []TrendDataPoint) (float64, float64) {
	if len(dataPoints) < 2 {
		return 0.0, 0.0
	}

	alpha := 0.3 // Smoothing factor
	var smoothed []float64
	var trend []float64

	// Initialize
	smoothed = append(smoothed, float64(dataPoints[0].ResponseTime.Milliseconds()))
	trend = append(trend, 0.0)

	// Apply exponential smoothing with trend
	for i := 1; i < len(dataPoints); i++ {
		current := float64(dataPoints[i].ResponseTime.Milliseconds())
		prevSmoothed := smoothed[i-1]
		prevTrend := trend[i-1]

		newSmoothed := alpha*current + (1-alpha)*(prevSmoothed+prevTrend)
		newTrend := 0.3*(newSmoothed-prevSmoothed) + 0.7*prevTrend

		smoothed = append(smoothed, newSmoothed)
		trend = append(trend, newTrend)
	}

	// Return the final trend value
	finalTrend := trend[len(trend)-1]

	// Calculate confidence based on trend consistency
	var trendVariance float64
	meanTrend := finalTrend
	for _, t := range trend {
		trendVariance += (t - meanTrend) * (t - meanTrend)
	}
	trendVariance /= float64(len(trend))

	confidence := 1.0 / (1.0 + trendVariance)

	return finalTrend, confidence
}

// calculateChangePercent calculates the percentage change from first to last data point
func (rtt *ResponseTimeTracker) calculateChangePercent(dataPoints []TrendDataPoint) float64 {
	if len(dataPoints) < 2 {
		return 0.0
	}

	first := float64(dataPoints[0].ResponseTime.Milliseconds())
	last := float64(dataPoints[len(dataPoints)-1].ResponseTime.Milliseconds())

	if first == 0 {
		return 0.0
	}

	return ((last - first) / first) * 100.0
}

// determineTrendDirection determines the trend direction based on change percentage
func (rtt *ResponseTimeTracker) determineTrendDirection(changePercent float64) string {
	if changePercent < -rtt.trendConfig.ImprovementThreshold {
		return "improving" // Negative change means response times decreased (improved)
	} else if changePercent > rtt.trendConfig.DegradationThreshold {
		return "degrading" // Positive change means response times increased (degraded)
	} else {
		return "stable"
	}
}

// calculateAverageTrendStrength calculates average trend strength from multiple algorithms
func (rtt *ResponseTimeTracker) calculateAverageTrendStrength(trends []float64, confidences []float64) float64 {
	if len(trends) == 0 {
		return 0.0
	}

	var weightedSum float64
	var totalWeight float64

	for i, trend := range trends {
		weight := confidences[i]
		weightedSum += math.Abs(trend) * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	// Normalize to 0-1 range
	strength := weightedSum / totalWeight / 1000.0 // Normalize by typical response time scale
	if strength > 1.0 {
		strength = 1.0
	}

	return strength
}

// calculateAverageConfidence calculates average confidence from multiple algorithms
func (rtt *ResponseTimeTracker) calculateAverageConfidence(confidences []float64) float64 {
	if len(confidences) == 0 {
		return 0.0
	}

	var sum float64
	for _, confidence := range confidences {
		sum += confidence
	}

	return sum / float64(len(confidences))
}

// calculatePercentile calculates the nth percentile from a sorted slice
func (rtt *ResponseTimeTracker) calculatePercentile(values []time.Duration, percentile float64) time.Duration {
	if len(values) == 0 {
		return 0
	}

	index := int(percentile * float64(len(values)-1))
	if index >= len(values) {
		index = len(values) - 1
	}

	return values[index]
}

// DetectAnomalies detects anomalies in response time data
func (rtt *ResponseTimeTracker) DetectAnomalies(ctx context.Context, startTime, endTime time.Time) ([]AnomalyPoint, error) {
	var anomalies []AnomalyPoint

	for key, metrics := range rtt.metrics {
		endpoint, method := rtt.parseMetricsKey(key)
		endpointAnomalies, err := rtt.detectEndpointAnomalies(metrics, startTime, endTime, endpoint, method)
		if err != nil {
			rtt.logger.Warn("failed to detect anomalies for endpoint",
				zap.String("endpoint", endpoint),
				zap.Error(err))
			continue
		}
		anomalies = append(anomalies, endpointAnomalies...)
	}

	return anomalies, nil
}

// detectEndpointAnomalies detects anomalies for a specific endpoint
func (rtt *ResponseTimeTracker) detectEndpointAnomalies(metrics []*ResponseTimeMetric, startTime, endTime time.Time, endpoint, method string) ([]AnomalyPoint, error) {
	var anomalies []AnomalyPoint

	// Filter metrics by time range
	var filteredMetrics []*ResponseTimeMetric
	for _, metric := range metrics {
		if metric.Timestamp.Before(startTime) || metric.Timestamp.After(endTime) {
			continue
		}
		filteredMetrics = append(filteredMetrics, metric)
	}

	if len(filteredMetrics) < 10 {
		return anomalies, nil // Need sufficient data for anomaly detection
	}

	// Calculate baseline statistics
	var responseTimes []float64
	for _, metric := range filteredMetrics {
		responseTimes = append(responseTimes, float64(metric.ResponseTime.Milliseconds()))
	}

	mean, stdDev := rtt.CalculateMeanAndStdDev(responseTimes)

	// Detect anomalies using z-score method
	for _, metric := range filteredMetrics {
		responseTime := float64(metric.ResponseTime.Milliseconds())
		zScore := math.Abs((responseTime - mean) / stdDev)

		if zScore > rtt.trendConfig.AnomalyThreshold {
			anomaly := AnomalyPoint{
				Timestamp:    metric.Timestamp,
				ResponseTime: metric.ResponseTime,
				ExpectedTime: time.Duration(mean) * time.Millisecond,
				Deviation:    zScore,
				Severity:     rtt.DetermineAnomalySeverity(zScore),
				Description: fmt.Sprintf("Response time %v is %.1f standard deviations from expected %v",
					metric.ResponseTime, zScore, time.Duration(mean)*time.Millisecond),
			}
			anomalies = append(anomalies, anomaly)
		}
	}

	return anomalies, nil
}

// CalculateMeanAndStdDev calculates mean and standard deviation
func (rtt *ResponseTimeTracker) CalculateMeanAndStdDev(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0.0, 0.0
	}

	var sum float64
	for _, value := range values {
		sum += value
	}
	mean := sum / float64(len(values))

	var variance float64
	for _, value := range values {
		variance += (value - mean) * (value - mean)
	}
	variance /= float64(len(values))

	return mean, math.Sqrt(variance)
}

// DetermineAnomalySeverity determines anomaly severity based on z-score
func (rtt *ResponseTimeTracker) DetermineAnomalySeverity(zScore float64) string {
	if zScore >= 4.0 {
		return "critical"
	} else if zScore >= 3.0 {
		return "high"
	} else if zScore >= 2.5 {
		return "medium"
	} else {
		return "low"
	}
}

// analyzeSeasonality analyzes seasonal patterns in the data
func (rtt *ResponseTimeTracker) AnalyzeSeasonality(ctx context.Context, startTime, endTime time.Time) (map[string]*SeasonalityInfo, error) {
	seasonality := make(map[string]*SeasonalityInfo)

	for key, metrics := range rtt.metrics {
		endpoint, method := rtt.parseMetricsKey(key)
		endpointSeasonality, err := rtt.analyzeEndpointSeasonality(metrics, startTime, endTime, endpoint, method)
		if err != nil {
			rtt.logger.Warn("failed to analyze seasonality for endpoint",
				zap.String("endpoint", endpoint),
				zap.Error(err))
			continue
		}
		if endpointSeasonality.HasSeasonality {
			seasonality[key] = endpointSeasonality
		}
	}

	return seasonality, nil
}

// analyzeEndpointSeasonality analyzes seasonality for a specific endpoint
func (rtt *ResponseTimeTracker) analyzeEndpointSeasonality(metrics []*ResponseTimeMetric, startTime, endTime time.Time, endpoint, method string) (*SeasonalityInfo, error) {
	// Filter metrics by time range
	var filteredMetrics []*ResponseTimeMetric
	for _, metric := range metrics {
		if metric.Timestamp.Before(startTime) || metric.Timestamp.After(endTime) {
			continue
		}
		filteredMetrics = append(filteredMetrics, metric)
	}

	if len(filteredMetrics) < 24 {
		return &SeasonalityInfo{HasSeasonality: false}, nil // Need sufficient data
	}

	// Group by hour to detect daily patterns
	hourlyStats := make(map[int][]time.Duration)
	for _, metric := range filteredMetrics {
		hour := metric.Timestamp.Hour()
		hourlyStats[hour] = append(hourlyStats[hour], metric.ResponseTime)
	}

	// Calculate average response time by hour
	hourlyAverages := make(map[int]time.Duration)
	for hour, responseTimes := range hourlyStats {
		if len(responseTimes) == 0 {
			continue
		}
		var total time.Duration
		for _, rt := range responseTimes {
			total += rt
		}
		hourlyAverages[hour] = total / time.Duration(len(responseTimes))
	}

	// Detect seasonality by comparing variance across hours
	if len(hourlyAverages) >= 6 {
		var values []float64
		for _, avg := range hourlyAverages {
			values = append(values, float64(avg.Milliseconds()))
		}

		mean, stdDev := rtt.CalculateMeanAndStdDev(values)
		coefficientOfVariation := stdDev / mean

		// Consider seasonal if coefficient of variation is significant
		if coefficientOfVariation > 0.2 {
			// Find peak and valley times
			var peakTimes, valleyTimes []time.Time
			var maxAvg, minAvg time.Duration

			for hour, avg := range hourlyAverages {
				if avg > maxAvg || maxAvg == 0 {
					maxAvg = avg
					peakTimes = []time.Time{time.Date(2023, 1, 1, hour, 0, 0, 0, time.UTC)}
				} else if avg == maxAvg {
					peakTimes = append(peakTimes, time.Date(2023, 1, 1, hour, 0, 0, 0, time.UTC))
				}

				if avg < minAvg || minAvg == 0 {
					minAvg = avg
					valleyTimes = []time.Time{time.Date(2023, 1, 1, hour, 0, 0, 0, time.UTC)}
				} else if avg == minAvg {
					valleyTimes = append(valleyTimes, time.Date(2023, 1, 1, hour, 0, 0, 0, time.UTC))
				}
			}

			return &SeasonalityInfo{
				HasSeasonality: true,
				Period:         24 * time.Hour,
				Strength:       coefficientOfVariation,
				PeakTimes:      peakTimes,
				ValleyTimes:    valleyTimes,
			}, nil
		}
	}

	return &SeasonalityInfo{HasSeasonality: false}, nil
}

// generateTrendInsights generates insights from trend analysis
func (rtt *ResponseTimeTracker) generateTrendInsights(ctx context.Context, report *TrendAnalysisReport) ([]TrendInsight, error) {
	var insights []TrendInsight

	// Overall trend insight
	if report.OverallTrend != nil {
		insight := rtt.createTrendInsight(report.OverallTrend, "overall")
		insights = append(insights, insight)
	}

	// Endpoint-specific insights
	for key, trend := range report.EndpointTrends {
		if trend.TrendStrength > 0.3 { // Only significant trends
			insight := rtt.createTrendInsight(trend, key)
			insights = append(insights, insight)
		}
	}

	// Anomaly insights
	for _, anomaly := range report.Anomalies {
		if anomaly.Severity == "high" || anomaly.Severity == "critical" {
			insight := TrendInsight{
				ID:          generateTrendID(),
				Type:        "anomaly",
				Title:       fmt.Sprintf("High Severity Anomaly Detected"),
				Description: anomaly.Description,
				Impact:      anomaly.Severity,
				Confidence:  0.9,
				Evidence: fmt.Sprintf("Response time %v at %v was %.1f standard deviations from expected",
					anomaly.ResponseTime, anomaly.Timestamp.Format("2006-01-02 15:04:05"), anomaly.Deviation),
				Timestamp: anomaly.Timestamp,
			}
			insights = append(insights, insight)
		}
	}

	// Seasonality insights
	for key, seasonality := range report.Seasonality {
		if seasonality.HasSeasonality {
			insight := TrendInsight{
				ID:          generateTrendID(),
				Type:        "seasonality",
				Title:       fmt.Sprintf("Seasonal Pattern Detected for %s", key),
				Description: fmt.Sprintf("Response times show seasonal variation with %v period", seasonality.Period),
				Impact:      "medium",
				Confidence:  seasonality.Strength,
				Evidence:    fmt.Sprintf("Coefficient of variation: %.2f", seasonality.Strength),
				Timestamp:   time.Now(),
			}
			insights = append(insights, insight)
		}
	}

	return insights, nil
}

// createTrendInsight creates a trend insight for a specific trend
func (rtt *ResponseTimeTracker) createTrendInsight(trend *ResponseTimeTrend, context string) TrendInsight {
	var title, description, impact string
	var confidence float64

	switch trend.TrendDirection {
	case "improving":
		title = fmt.Sprintf("Performance Improvement Detected for %s", context)
		description = fmt.Sprintf("Response times improved by %.1f%% over %v",
			math.Abs(trend.ChangePercent), trend.Period)
		impact = "positive"
		confidence = trend.Confidence
	case "degrading":
		title = fmt.Sprintf("Performance Degradation Detected for %s", context)
		description = fmt.Sprintf("Response times degraded by %.1f%% over %v",
			math.Abs(trend.ChangePercent), trend.Period)
		impact = "high"
		confidence = trend.Confidence
	default:
		title = fmt.Sprintf("Stable Performance for %s", context)
		description = fmt.Sprintf("Response times remained stable with %.1f%% change over %v",
			math.Abs(trend.ChangePercent), trend.Period)
		impact = "low"
		confidence = trend.Confidence
	}

	return TrendInsight{
		ID:          generateTrendID(),
		Type:        "performance_" + trend.TrendDirection,
		Title:       title,
		Description: description,
		Impact:      impact,
		Confidence:  confidence,
		Evidence: fmt.Sprintf("Trend strength: %.2f, Change: %.1f%%, Period: %v",
			trend.TrendStrength, trend.ChangePercent, trend.Period),
		Timestamp: time.Now(),
	}
}

// generateTrendRecommendations generates recommendations based on trend analysis
func (rtt *ResponseTimeTracker) generateTrendRecommendations(ctx context.Context, report *TrendAnalysisReport) ([]TrendRecommendation, error) {
	var recommendations []TrendRecommendation

	// Recommendations based on degrading trends
	for key, trend := range report.EndpointTrends {
		if trend.TrendDirection == "degrading" && trend.TrendStrength > 0.3 {
			recommendation := rtt.createDegradationRecommendation(trend, key)
			recommendations = append(recommendations, recommendation)
		}
	}

	// Recommendations based on anomalies
	highSeverityAnomalies := 0
	for _, anomaly := range report.Anomalies {
		if anomaly.Severity == "high" || anomaly.Severity == "critical" {
			highSeverityAnomalies++
		}
	}

	if highSeverityAnomalies > 5 {
		recommendation := TrendRecommendation{
			ID:                   generateTrendID(),
			Title:                "High Anomaly Rate Detected",
			Description:          fmt.Sprintf("Detected %d high-severity anomalies in the analysis period", highSeverityAnomalies),
			Category:             "monitoring",
			Priority:             8,
			Impact:               "high",
			Effort:               "medium",
			EstimatedImprovement: 15.0,
			Confidence:           0.8,
			Actions:              []string{"Implement anomaly detection alerts", "Review system stability", "Add performance monitoring"},
			CreatedAt:            time.Now(),
		}
		recommendations = append(recommendations, recommendation)
	}

	// Recommendations based on seasonality
	for key, seasonality := range report.Seasonality {
		if seasonality.HasSeasonality && seasonality.Strength > 0.3 {
			recommendation := TrendRecommendation{
				ID:                   generateTrendID(),
				Title:                fmt.Sprintf("Seasonal Pattern Optimization for %s", key),
				Description:          fmt.Sprintf("Consider capacity planning for seasonal patterns with %v period", seasonality.Period),
				Category:             "capacity",
				Priority:             6,
				Impact:               "medium",
				Effort:               "high",
				EstimatedImprovement: 10.0,
				Confidence:           seasonality.Strength,
				Actions:              []string{"Implement auto-scaling based on time patterns", "Optimize resource allocation"},
				CreatedAt:            time.Now(),
			}
			recommendations = append(recommendations, recommendation)
		}
	}

	return recommendations, nil
}

// createDegradationRecommendation creates a recommendation for performance degradation
func (rtt *ResponseTimeTracker) createDegradationRecommendation(trend *ResponseTimeTrend, context string) TrendRecommendation {
	var priority int
	var impact, effort string
	var estimatedImprovement float64

	if trend.TrendStrength > 0.7 {
		priority = 9
		impact = "critical"
		effort = "high"
		estimatedImprovement = 25.0
	} else if trend.TrendStrength > 0.5 {
		priority = 7
		impact = "high"
		effort = "medium"
		estimatedImprovement = 15.0
	} else {
		priority = 5
		impact = "medium"
		effort = "low"
		estimatedImprovement = 10.0
	}

	return TrendRecommendation{
		ID:    generateTrendID(),
		Title: fmt.Sprintf("Performance Optimization for %s", context),
		Description: fmt.Sprintf("Response times degraded by %.1f%% over %v",
			math.Abs(trend.ChangePercent), trend.Period),
		Category:             "optimization",
		Priority:             priority,
		Impact:               impact,
		Effort:               effort,
		EstimatedImprovement: estimatedImprovement,
		Confidence:           trend.Confidence,
		Actions:              []string{"Review database queries", "Optimize caching strategy", "Check resource utilization"},
		CreatedAt:            time.Now(),
	}
}

// generateTrendSummary generates a summary of trend analysis
func (rtt *ResponseTimeTracker) GenerateTrendSummary(report *TrendAnalysisReport) TrendSummary {
	summary := TrendSummary{
		TotalEndpoints:     len(report.EndpointTrends),
		ImprovingEndpoints: 0,
		DegradingEndpoints: 0,
		StableEndpoints:    0,
		AnomalyCount:       len(report.Anomalies),
		SeasonalPatterns:   len(report.Seasonality),
	}

	var totalImprovement, totalDegradation float64
	var improvementCount, degradationCount int

	for _, trend := range report.EndpointTrends {
		switch trend.TrendDirection {
		case "improving":
			summary.ImprovingEndpoints++
			totalImprovement += math.Abs(trend.ChangePercent)
			improvementCount++
		case "degrading":
			summary.DegradingEndpoints++
			totalDegradation += math.Abs(trend.ChangePercent)
			degradationCount++
		default:
			summary.StableEndpoints++
		}
	}

	if improvementCount > 0 {
		summary.AverageImprovement = totalImprovement / float64(improvementCount)
	}
	if degradationCount > 0 {
		summary.AverageDegradation = totalDegradation / float64(degradationCount)
	}

	// Determine overall health
	summary.OverallHealth = rtt.DetermineOverallHealth(summary)

	return summary
}

// determineOverallHealth determines overall system health based on summary
func (rtt *ResponseTimeTracker) DetermineOverallHealth(summary TrendSummary) string {
	if summary.TotalEndpoints == 0 {
		return "unknown"
	}

	degradingRatio := float64(summary.DegradingEndpoints) / float64(summary.TotalEndpoints)
	improvingRatio := float64(summary.ImprovingEndpoints) / float64(summary.TotalEndpoints)

	if degradingRatio > 0.5 {
		return "critical"
	} else if degradingRatio > 0.3 {
		return "poor"
	} else if degradingRatio > 0.1 {
		return "fair"
	} else if improvingRatio > 0.3 {
		return "excellent"
	} else {
		return "good"
	}
}

// GetTrendAnalysisReport retrieves a specific trend analysis report
func (rtt *ResponseTimeTracker) GetTrendAnalysisReport(reportID string) (*TrendAnalysisReport, error) {
	rtt.mutex.RLock()
	defer rtt.mutex.RUnlock()

	report, exists := rtt.reports[reportID]
	if !exists {
		return nil, fmt.Errorf("report not found: %s", reportID)
	}

	return report, nil
}

// GetTrendAnalysisReports retrieves all trend analysis reports with optional filtering
func (rtt *ResponseTimeTracker) GetTrendAnalysisReports(startTime, endTime *time.Time, limit int) ([]*TrendAnalysisReport, error) {
	rtt.mutex.RLock()
	defer rtt.mutex.RUnlock()

	var reports []*TrendAnalysisReport
	count := 0

	for _, report := range rtt.reports {
		// Apply time filters
		if startTime != nil && report.GeneratedAt.Before(*startTime) {
			continue
		}
		if endTime != nil && report.GeneratedAt.After(*endTime) {
			continue
		}

		reports = append(reports, report)
		count++

		if limit > 0 && count >= limit {
			break
		}
	}

	// Sort by generation time (newest first)
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].GeneratedAt.After(reports[j].GeneratedAt)
	})

	return reports, nil
}

// GetTrendAnalysisStatistics provides statistics about trend analysis reports
func (rtt *ResponseTimeTracker) GetTrendAnalysisStatistics() map[string]interface{} {
	rtt.mutex.RLock()
	defer rtt.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_reports":                       0,
		"reports_last_24h":                    0,
		"reports_last_7d":                     0,
		"average_insights":                    0.0,
		"average_recommendations":             0.0,
		"most_common_insight_type":            "",
		"most_common_recommendation_category": "",
	}

	if len(rtt.reports) == 0 {
		return stats
	}

	stats["total_reports"] = len(rtt.reports)

	now := time.Now()
	last24h := now.Add(-24 * time.Hour)
	last7d := now.Add(-7 * 24 * time.Hour)

	var totalInsights, totalRecommendations int
	insightTypes := make(map[string]int)
	recommendationCategories := make(map[string]int)

	for _, report := range rtt.reports {
		if report.GeneratedAt.After(last24h) {
			stats["reports_last_24h"] = stats["reports_last_24h"].(int) + 1
		}
		if report.GeneratedAt.After(last7d) {
			stats["reports_last_7d"] = stats["reports_last_7d"].(int) + 1
		}

		totalInsights += len(report.KeyInsights)
		totalRecommendations += len(report.Recommendations)

		for _, insight := range report.KeyInsights {
			insightTypes[insight.Type]++
		}

		for _, recommendation := range report.Recommendations {
			recommendationCategories[recommendation.Category]++
		}
	}

	stats["average_insights"] = float64(totalInsights) / float64(len(rtt.reports))
	stats["average_recommendations"] = float64(totalRecommendations) / float64(len(rtt.reports))

	// Find most common types
	var maxInsightCount int
	var maxInsightType string
	for insightType, count := range insightTypes {
		if count > maxInsightCount {
			maxInsightCount = count
			maxInsightType = insightType
		}
	}
	stats["most_common_insight_type"] = maxInsightType

	var maxRecCount int
	var maxRecCategory string
	for category, count := range recommendationCategories {
		if count > maxRecCount {
			maxRecCount = count
			maxRecCategory = category
		}
	}
	stats["most_common_recommendation_category"] = maxRecCategory

	return stats
}

// UpdateTrendAnalysisConfig updates the trend analysis configuration
func (rtt *ResponseTimeTracker) UpdateTrendAnalysisConfig(config *TrendAnalysisConfig) error {
	if config == nil {
		return errors.New("config cannot be nil")
	}

	rtt.mutex.Lock()
	defer rtt.mutex.Unlock()

	rtt.trendConfig = config
	rtt.logger.Info("trend analysis configuration updated")

	return nil
}

// GetTrendAnalysisConfig returns the current trend analysis configuration
func (rtt *ResponseTimeTracker) GetTrendAnalysisConfig() *TrendAnalysisConfig {
	rtt.mutex.RLock()
	defer rtt.mutex.RUnlock()

	return rtt.trendConfig
}

// Utility functions for trend analysis
func (rtt *ResponseTimeTracker) BuildEndpointKey(endpoint, method string) string {
	return fmt.Sprintf("%s_%s", method, endpoint)
}

func (rtt *ResponseTimeTracker) ParseEndpointKey(key string) (string, string) {
	parts := strings.SplitN(key, "_", 2)
	if len(parts) != 2 {
		return key, "GET"
	}
	return parts[1], parts[0]
}

func (rtt *ResponseTimeTracker) parseMetricsKey(key string) (string, string) {
	parts := strings.SplitN(key, "_", 2)
	if len(parts) != 2 {
		return key, "GET"
	}
	return parts[0], parts[1] // endpoint_method format
}

func generateTrendID() string {
	return fmt.Sprintf("trend_%d", time.Now().UnixNano())
}
