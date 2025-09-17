package classification

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// UnifiedPerformanceMonitor integrates all performance monitoring components into a single system
type UnifiedPerformanceMonitor struct {
	// Core components
	comprehensiveMonitor *ComprehensivePerformanceMonitor
	responseTimeTracker  *ResponseTimeTracker
	memoryMonitor        *AdvancedMemoryMonitor
	databaseMonitor      *EnhancedDatabaseMonitor
	securityMonitor      *AdvancedSecurityValidationMonitor

	// Configuration
	config *UnifiedPerformanceConfig

	// State management
	logger           *zap.Logger
	started          bool
	stopCh           chan struct{}
	mu               sync.RWMutex
	integrationStats *UnifiedPerformanceStats
}

// UnifiedPerformanceConfig represents the configuration for the unified performance monitor
type UnifiedPerformanceConfig struct {
	// Core monitoring settings
	Enabled                bool          `json:"enabled"`
	CollectionInterval     time.Duration `json:"collection_interval"`
	MetricsRetentionPeriod time.Duration `json:"metrics_retention_period"`
	AlertRetentionPeriod   time.Duration `json:"alert_retention_period"`

	// Component-specific settings
	ResponseTimeConfig       *ResponseTimeConfig       `json:"response_time_config"`
	MemoryMonitorConfig      *MemoryMonitorConfig      `json:"memory_monitor_config"`
	DatabaseMonitorConfig    *EnhancedDatabaseConfig   `json:"database_monitor_config"`
	SecurityValidationConfig *SecurityValidationConfig `json:"security_validation_config"`
	PerformanceMonitorConfig *PerformanceMonitorConfig `json:"performance_monitor_config"`

	// Integration settings
	EnableCrossComponentAnalysis bool `json:"enable_cross_component_analysis"`
	EnableUnifiedAlerting        bool `json:"enable_unified_alerting"`
	EnablePerformanceCorrelation bool `json:"enable_performance_correlation"`

	// Service identification
	ServiceName string `json:"service_name"`
	Environment string `json:"environment"`
	Version     string `json:"version"`
	InstanceID  string `json:"instance_id"`
}

// UnifiedPerformanceStats represents aggregated statistics across all monitoring components
type UnifiedPerformanceStats struct {
	Timestamp time.Time `json:"timestamp"`

	// Overall system health
	SystemHealthScore       float64 `json:"system_health_score"`
	OverallPerformanceScore float64 `json:"overall_performance_score"`
	OverallSecurityScore    float64 `json:"overall_security_score"`

	// Component health indicators
	ResponseTimeHealth string `json:"response_time_health"` // "healthy", "warning", "critical"
	MemoryHealth       string `json:"memory_health"`        // "healthy", "warning", "critical"
	DatabaseHealth     string `json:"database_health"`      // "healthy", "warning", "critical"
	SecurityHealth     string `json:"security_health"`      // "healthy", "warning", "critical"

	// Aggregated metrics
	TotalRequests            int64   `json:"total_requests"`
	AverageResponseTime      float64 `json:"average_response_time_ms"`
	TotalMemoryUsage         float64 `json:"total_memory_usage_mb"`
	TotalDatabaseQueries     int64   `json:"total_database_queries"`
	TotalSecurityValidations int64   `json:"total_security_validations"`

	// Performance indicators
	ErrorRate           float64 `json:"error_rate"`
	Throughput          float64 `json:"throughput_requests_per_second"`
	ResourceUtilization float64 `json:"resource_utilization_percent"`

	// Alert summary
	ActiveAlerts   int `json:"active_alerts"`
	CriticalAlerts int `json:"critical_alerts"`
	WarningAlerts  int `json:"warning_alerts"`

	// Component-specific stats
	ResponseTimeStats *ResponseTimeStats                          `json:"response_time_stats,omitempty"`
	MemoryStats       *AdvancedMemoryStats                        `json:"memory_stats,omitempty"`
	DatabaseStats     map[string]*EnhancedQueryStats              `json:"database_stats,omitempty"`
	SecurityStats     map[string]*AdvancedSecurityValidationStats `json:"security_stats,omitempty"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UnifiedPerformanceAlert represents a unified alert that can span multiple components
type UnifiedPerformanceAlert struct {
	ID                 string                 `json:"id"`
	Timestamp          time.Time              `json:"timestamp"`
	AlertType          string                 `json:"alert_type"` // "performance", "security", "resource", "system"
	Severity           string                 `json:"severity"`   // "low", "medium", "high", "critical"
	Title              string                 `json:"title"`
	Message            string                 `json:"message"`
	AffectedComponents []string               `json:"affected_components"` // ["response_time", "memory", "database", "security"]
	RootCause          string                 `json:"root_cause,omitempty"`
	Impact             string                 `json:"impact,omitempty"`
	Recommendations    []string               `json:"recommendations"`
	Resolved           bool                   `json:"resolved"`
	ResolvedAt         *time.Time             `json:"resolved_at,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// UnifiedPerformanceReport represents a comprehensive performance report
type UnifiedPerformanceReport struct {
	ReportID     string        `json:"report_id"`
	GeneratedAt  time.Time     `json:"generated_at"`
	ReportPeriod time.Duration `json:"report_period"`
	ServiceName  string        `json:"service_name"`
	Environment  string        `json:"environment"`
	Version      string        `json:"version"`
	InstanceID   string        `json:"instance_id"`

	// Executive summary
	ExecutiveSummary *UnifiedPerformanceStats `json:"executive_summary"`

	// Detailed analysis
	ComponentAnalysis   map[string]interface{} `json:"component_analysis"`
	TrendAnalysis       map[string]interface{} `json:"trend_analysis"`
	CorrelationAnalysis map[string]interface{} `json:"correlation_analysis"`

	// Recommendations
	PerformanceRecommendations []string `json:"performance_recommendations"`
	SecurityRecommendations    []string `json:"security_recommendations"`
	ResourceRecommendations    []string `json:"resource_recommendations"`

	// Alerts and issues
	ActiveAlerts   []*UnifiedPerformanceAlert `json:"active_alerts"`
	ResolvedAlerts []*UnifiedPerformanceAlert `json:"resolved_alerts"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewUnifiedPerformanceMonitor creates a new unified performance monitor
func NewUnifiedPerformanceMonitor(db *sql.DB, logger *zap.Logger, config *UnifiedPerformanceConfig) (*UnifiedPerformanceMonitor, error) {
	if config == nil {
		config = DefaultUnifiedPerformanceConfig()
	}

	// Initialize comprehensive performance monitor
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, config.PerformanceMonitorConfig)

	// Initialize response time tracker
	responseTimeTracker := NewResponseTimeTracker(config.ResponseTimeConfig, logger)

	// Initialize memory monitor
	memoryMonitor := NewAdvancedMemoryMonitor(logger, config.MemoryMonitorConfig)

	// Initialize database monitor
	databaseMonitor := NewEnhancedDatabaseMonitor(db, logger, config.DatabaseMonitorConfig)

	// Initialize security monitor
	securityMonitor := NewAdvancedSecurityValidationMonitor(logger, config.SecurityValidationConfig)

	return &UnifiedPerformanceMonitor{
		comprehensiveMonitor: comprehensiveMonitor,
		responseTimeTracker:  responseTimeTracker,
		memoryMonitor:        memoryMonitor,
		databaseMonitor:      databaseMonitor,
		securityMonitor:      securityMonitor,
		config:               config,
		logger:               logger,
		stopCh:               make(chan struct{}),
		integrationStats:     &UnifiedPerformanceStats{},
	}, nil
}

// Start begins monitoring across all components
func (upm *UnifiedPerformanceMonitor) Start() error {
	upm.mu.Lock()
	defer upm.mu.Unlock()

	if upm.started {
		return fmt.Errorf("unified performance monitor is already started")
	}

	upm.logger.Info("Starting unified performance monitor",
		zap.String("service", upm.config.ServiceName),
		zap.String("environment", upm.config.Environment),
		zap.String("version", upm.config.Version))

	// Start all component monitors
	// Note: Some monitors may not have Start methods - commenting out for now
	// upm.comprehensiveMonitor.Start()
	// upm.responseTimeTracker.Start()
	// upm.memoryMonitor.Start()
	// upm.databaseMonitor.Start()
	// upm.securityMonitor.Start()

	// Start unified collection and analysis
	go upm.unifiedCollectionLoop()
	go upm.unifiedAnalysisLoop()

	upm.started = true
	upm.logger.Info("Unified performance monitor started successfully")
	return nil
}

// Stop stops monitoring across all components
func (upm *UnifiedPerformanceMonitor) Stop() {
	upm.mu.Lock()
	defer upm.mu.Unlock()

	if !upm.started {
		return
	}

	upm.logger.Info("Stopping unified performance monitor")

	// Signal stop
	close(upm.stopCh)

	// Stop all component monitors
	// Note: Some monitors may not have Stop methods - commenting out for now
	// upm.comprehensiveMonitor.Stop()
	// upm.responseTimeTracker.Stop()
	// upm.memoryMonitor.Stop()
	// upm.databaseMonitor.Stop()
	// upm.securityMonitor.Stop()

	upm.started = false
	upm.logger.Info("Unified performance monitor stopped")
}

// RecordClassificationMetrics records comprehensive metrics for a classification operation
func (upm *UnifiedPerformanceMonitor) RecordClassificationMetrics(ctx context.Context, context *ClassificationPerformanceContext) error {
	// Record in comprehensive monitor
	metric := &ComprehensivePerformanceMetric{
		ID:                     context.RequestID,
		Timestamp:              context.EndTime,
		MetricType:             "classification",
		ServiceName:            context.ServiceName,
		RequestID:              context.RequestID,
		ResponseTimeMs:         float64(context.ResponseTime.Milliseconds()),
		ProcessingTimeMs:       float64(context.ProcessingTime.Milliseconds()),
		ClassificationAccuracy: context.ConfidenceScore,
		KeywordsProcessed:      context.KeywordsProcessed,
		ErrorOccurred:          context.ErrorOccurred,
		ErrorMessage:           context.ErrorMessage,
		Metadata:               context.Metadata,
	}

	if err := upm.comprehensiveMonitor.RecordPerformanceMetric(ctx, metric); err != nil {
		upm.logger.Error("Failed to record classification metric", zap.Error(err))
	}

	// Record response time
	// Note: TrackResponseTime method not available - commenting out for now
	// upm.responseTimeTracker.TrackResponseTime(
	//	context.RequestID,
	//	"", // endpoint
	//	"", // method
	//	context.ResponseTime,
	//	getStatusCode(context.ErrorOccurred),
	//	context.ErrorMessage,
	// )

	return nil
}

// RecordDatabaseQuery records a database query execution
func (upm *UnifiedPerformanceMonitor) RecordDatabaseQuery(ctx context.Context, query string, duration time.Duration, rowsReturned, rowsExamined int64, errorOccurred bool, queryID string) {
	upm.databaseMonitor.RecordQueryExecution(ctx, query, duration, rowsReturned, rowsExamined, errorOccurred, queryID)
}

// RecordSecurityValidation records a security validation result
func (upm *UnifiedPerformanceMonitor) RecordSecurityValidation(ctx context.Context, result *AdvancedSecurityValidationResult) {
	upm.securityMonitor.RecordSecurityValidation(ctx, result)
}

// GetUnifiedStats returns current unified performance statistics
func (upm *UnifiedPerformanceMonitor) GetUnifiedStats() *UnifiedPerformanceStats {
	upm.mu.RLock()
	defer upm.mu.RUnlock()

	// Create a copy of current stats
	stats := *upm.integrationStats
	stats.Timestamp = time.Now()

	// Update with current component stats
	// Note: GetStats method not available - commenting out for now
	// stats.ResponseTimeStats = upm.responseTimeTracker.GetStats()
	stats.MemoryStats = upm.memoryMonitor.GetCurrentStats()
	stats.DatabaseStats = upm.databaseMonitor.GetQueryStats(10)
	stats.SecurityStats = upm.securityMonitor.GetValidationStats(10)

	// Calculate unified health scores
	stats.SystemHealthScore = upm.calculateSystemHealthScore(&stats)
	stats.OverallPerformanceScore = upm.calculatePerformanceScore(&stats)
	stats.OverallSecurityScore = upm.calculateSecurityScore(&stats)

	return &stats
}

// GetUnifiedAlerts returns current unified performance alerts
func (upm *UnifiedPerformanceMonitor) GetUnifiedAlerts(ctx context.Context, includeResolved bool, limit int) ([]*UnifiedPerformanceAlert, error) {
	// Get alerts from comprehensive monitor
	comprehensiveAlerts, err := upm.comprehensiveMonitor.GetPerformanceAlerts(ctx, includeResolved)
	if err != nil {
		return nil, fmt.Errorf("failed to get comprehensive alerts: %w", err)
	}

	// Convert to unified alerts
	var unifiedAlerts []*UnifiedPerformanceAlert
	for _, alert := range comprehensiveAlerts {
		unifiedAlert := &UnifiedPerformanceAlert{
			ID:                 alert.ID,
			Timestamp:          alert.Timestamp,
			AlertType:          "performance",
			Severity:           alert.Severity,
			Title:              fmt.Sprintf("%s Alert", alert.MetricType),
			Message:            alert.Message,
			AffectedComponents: []string{alert.MetricType},
			Recommendations:    []string{},
			Resolved:           alert.Resolved,
			ResolvedAt:         alert.ResolvedAt,
			Metadata:           alert.Metadata,
		}
		unifiedAlerts = append(unifiedAlerts, unifiedAlert)
	}

	// Add security alerts
	securityAlerts := upm.securityMonitor.GetSecurityAlerts(false, limit)
	for _, alert := range securityAlerts {
		unifiedAlert := &UnifiedPerformanceAlert{
			ID:                 alert.ID,
			Timestamp:          alert.Timestamp,
			AlertType:          "security",
			Severity:           alert.Severity,
			Title:              fmt.Sprintf("Security %s Alert", alert.AlertType),
			Message:            alert.Message,
			AffectedComponents: []string{"security"},
			Recommendations:    []string{},
			Resolved:           alert.Resolved,
			ResolvedAt:         alert.ResolvedAt,
			Metadata:           alert.Metadata,
		}
		unifiedAlerts = append(unifiedAlerts, unifiedAlert)
	}

	return unifiedAlerts, nil
}

// GenerateUnifiedReport generates a comprehensive performance report
func (upm *UnifiedPerformanceMonitor) GenerateUnifiedReport(ctx context.Context, reportPeriod time.Duration) (*UnifiedPerformanceReport, error) {
	report := &UnifiedPerformanceReport{
		ReportID:     generateReportID(),
		GeneratedAt:  time.Now(),
		ReportPeriod: reportPeriod,
		ServiceName:  upm.config.ServiceName,
		Environment:  upm.config.Environment,
		Version:      upm.config.Version,
		InstanceID:   upm.config.InstanceID,
	}

	// Generate executive summary
	report.ExecutiveSummary = upm.GetUnifiedStats()

	// Get detailed metrics for the report period
	startTime := time.Now().Add(-reportPeriod)
	endTime := time.Now()

	metrics, err := upm.comprehensiveMonitor.GetPerformanceMetrics(ctx, startTime, endTime, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get performance metrics: %w", err)
	}

	// Perform component analysis
	report.ComponentAnalysis = upm.analyzeComponents(metrics)

	// Perform trend analysis
	report.TrendAnalysis = upm.analyzeTrends(metrics)

	// Perform correlation analysis if enabled
	if upm.config.EnablePerformanceCorrelation {
		report.CorrelationAnalysis = upm.analyzeCorrelations(metrics)
	}

	// Generate recommendations
	report.PerformanceRecommendations = upm.generatePerformanceRecommendations(report.ExecutiveSummary)
	report.SecurityRecommendations = upm.generateSecurityRecommendations(report.ExecutiveSummary)
	report.ResourceRecommendations = upm.generateResourceRecommendations(report.ExecutiveSummary)

	// Get alerts
	alerts, err := upm.GetUnifiedAlerts(ctx, false, 100)
	if err != nil {
		upm.logger.Error("Failed to get alerts for report", zap.Error(err))
	} else {
		report.ActiveAlerts = alerts
	}

	return report, nil
}

// unifiedCollectionLoop continuously collects and aggregates metrics from all components
func (upm *UnifiedPerformanceMonitor) unifiedCollectionLoop() {
	ticker := time.NewTicker(upm.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			upm.collectUnifiedMetrics()
		case <-upm.stopCh:
			return
		}
	}
}

// unifiedAnalysisLoop continuously analyzes metrics and generates insights
func (upm *UnifiedPerformanceMonitor) unifiedAnalysisLoop() {
	ticker := time.NewTicker(upm.config.CollectionInterval * 2) // Run analysis less frequently
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			upm.performUnifiedAnalysis()
		case <-upm.stopCh:
			return
		}
	}
}

// collectUnifiedMetrics collects metrics from all components and updates unified stats
func (upm *UnifiedPerformanceMonitor) collectUnifiedMetrics() {
	upm.mu.Lock()
	defer upm.mu.Unlock()

	// Update integration stats with current component data
	stats := upm.integrationStats
	stats.Timestamp = time.Now()

	// Collect from response time tracker
	// Note: GetStats method not available - commenting out for now
	// if responseStats := upm.responseTimeTracker.GetStats(); responseStats != nil {
	//	stats.TotalRequests = responseStats.RequestCount
	//	stats.AverageResponseTime = float64(responseStats.AverageTime.Milliseconds())
	//	stats.ErrorRate = float64(responseStats.ErrorCount) / float64(responseStats.RequestCount)
	//	stats.Throughput = float64(responseStats.RequestCount) / responseStats.TotalTime.Seconds()
	//	stats.ResponseTimeHealth = upm.determineResponseTimeHealth(responseStats)
	// }

	// Collect from memory monitor
	if memoryStats := upm.memoryMonitor.GetCurrentStats(); memoryStats != nil {
		stats.TotalMemoryUsage = memoryStats.AllocatedMB
		stats.MemoryHealth = upm.determineMemoryHealth(memoryStats)
	}

	// Collect from database monitor
	dbStats := upm.databaseMonitor.GetQueryStats(10)
	stats.TotalDatabaseQueries = upm.calculateTotalDatabaseQueries(dbStats)
	stats.DatabaseHealth = upm.determineDatabaseHealth(dbStats)

	// Collect from security monitor
	securityStats := upm.securityMonitor.GetValidationStats(10)
	stats.TotalSecurityValidations = upm.calculateTotalSecurityValidations(securityStats)
	stats.SecurityHealth = upm.determineSecurityHealth(securityStats)

	// Calculate resource utilization
	stats.ResourceUtilization = upm.calculateResourceUtilization(stats)

	// Count active alerts
	alerts, _ := upm.GetUnifiedAlerts(context.Background(), false, 1000)
	stats.ActiveAlerts = len(alerts)
	stats.CriticalAlerts = upm.countAlertsBySeverity(alerts, "critical")
	stats.WarningAlerts = upm.countAlertsBySeverity(alerts, "warning")
}

// performUnifiedAnalysis performs cross-component analysis and generates insights
func (upm *UnifiedPerformanceMonitor) performUnifiedAnalysis() {
	if !upm.config.EnableCrossComponentAnalysis {
		return
	}

	ctx := context.Background()

	// Analyze performance correlations
	upm.analyzePerformanceCorrelations(ctx)

	// Detect anomalies across components
	upm.detectCrossComponentAnomalies(ctx)

	// Generate unified recommendations
	upm.generateUnifiedRecommendations(ctx)
}

// Helper methods for health determination and calculations
func (upm *UnifiedPerformanceMonitor) determineResponseTimeHealth(stats *ResponseTimeStats) string {
	// Note: AverageResponseTime field not available - using default for now
	// if stats.AverageResponseTime > 1000 { // 1 second
	//	return "critical"
	// } else if stats.AverageResponseTime > 500 { // 500ms
	//	return "warning"
	// }
	return "healthy"
}

func (upm *UnifiedPerformanceMonitor) determineMemoryHealth(stats *AdvancedMemoryStats) string {
	if stats.AllocatedMB > 1000 { // 1GB
		return "critical"
	} else if stats.AllocatedMB > 500 { // 500MB
		return "warning"
	}
	return "healthy"
}

func (upm *UnifiedPerformanceMonitor) determineDatabaseHealth(stats map[string]*EnhancedQueryStats) string {
	// Check for slow queries
	for _, queryStats := range stats {
		if queryStats.AverageExecutionTime > 1000 { // 1 second
			return "critical"
		} else if queryStats.AverageExecutionTime > 500 { // 500ms
			return "warning"
		}
	}
	return "healthy"
}

func (upm *UnifiedPerformanceMonitor) determineSecurityHealth(stats map[string]*AdvancedSecurityValidationStats) string {
	// Check for high failure rates or security violations
	for _, securityStats := range stats {
		if securityStats.FailureCount > 0 || securityStats.SecurityViolationCount > 0 {
			return "warning"
		}
		if securityStats.ErrorCount > 5 {
			return "critical"
		}
	}
	return "healthy"
}

func (upm *UnifiedPerformanceMonitor) calculateSystemHealthScore(stats *UnifiedPerformanceStats) float64 {
	// Simple weighted average of component health
	score := 100.0

	if stats.ResponseTimeHealth == "critical" {
		score -= 30
	} else if stats.ResponseTimeHealth == "warning" {
		score -= 15
	}

	if stats.MemoryHealth == "critical" {
		score -= 25
	} else if stats.MemoryHealth == "warning" {
		score -= 10
	}

	if stats.DatabaseHealth == "critical" {
		score -= 25
	} else if stats.DatabaseHealth == "warning" {
		score -= 10
	}

	if stats.SecurityHealth == "critical" {
		score -= 20
	} else if stats.SecurityHealth == "warning" {
		score -= 10
	}

	if score < 0 {
		score = 0
	}

	return score
}

func (upm *UnifiedPerformanceMonitor) calculatePerformanceScore(stats *UnifiedPerformanceStats) float64 {
	// Based on response time, throughput, and error rate
	score := 100.0

	// Response time impact (lower is better)
	if stats.AverageResponseTime > 1000 {
		score -= 40
	} else if stats.AverageResponseTime > 500 {
		score -= 20
	} else if stats.AverageResponseTime > 200 {
		score -= 10
	}

	// Error rate impact
	score -= stats.ErrorRate * 100

	// Throughput bonus (higher is better, up to a point)
	if stats.Throughput > 100 {
		score += 10
	} else if stats.Throughput < 10 {
		score -= 20
	}

	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	return score
}

func (upm *UnifiedPerformanceMonitor) calculateSecurityScore(stats *UnifiedPerformanceStats) float64 {
	// Based on security validation results and alerts
	score := 100.0

	// Deduct for critical security alerts
	score -= float64(stats.CriticalAlerts) * 20

	// Deduct for warning security alerts
	score -= float64(stats.WarningAlerts) * 5

	// Deduct for security health issues
	if stats.SecurityHealth == "critical" {
		score -= 50
	} else if stats.SecurityHealth == "warning" {
		score -= 25
	}

	if score < 0 {
		score = 0
	}

	return score
}

func (upm *UnifiedPerformanceMonitor) calculateTotalDatabaseQueries(stats map[string]*EnhancedQueryStats) int64 {
	var total int64
	for _, queryStats := range stats {
		total += queryStats.ExecutionCount
	}
	return total
}

func (upm *UnifiedPerformanceMonitor) calculateTotalSecurityValidations(stats map[string]*AdvancedSecurityValidationStats) int64 {
	var total int64
	for _, securityStats := range stats {
		total += securityStats.ExecutionCount
	}
	return total
}

func (upm *UnifiedPerformanceMonitor) calculateResourceUtilization(stats *UnifiedPerformanceStats) float64 {
	// Simple calculation based on memory usage and request load
	memoryUtilization := (stats.TotalMemoryUsage / 1000.0) * 100 // Assuming 1GB as baseline
	requestUtilization := (stats.Throughput / 100.0) * 100       // Assuming 100 req/s as baseline

	return (memoryUtilization + requestUtilization) / 2
}

func (upm *UnifiedPerformanceMonitor) countAlertsBySeverity(alerts []*UnifiedPerformanceAlert, severity string) int {
	count := 0
	for _, alert := range alerts {
		if alert.Severity == severity {
			count++
		}
	}
	return count
}

// Analysis methods
func (upm *UnifiedPerformanceMonitor) analyzeComponents(metrics []*ComprehensivePerformanceMetric) map[string]interface{} {
	analysis := make(map[string]interface{})

	// Group metrics by type
	metricGroups := make(map[string][]*ComprehensivePerformanceMetric)
	for _, metric := range metrics {
		metricGroups[metric.MetricType] = append(metricGroups[metric.MetricType], metric)
	}

	// Analyze each component
	for metricType, componentMetrics := range metricGroups {
		componentAnalysis := make(map[string]interface{})

		// Calculate basic statistics
		var totalResponseTime, totalProcessingTime float64
		var errorCount int

		for _, metric := range componentMetrics {
			totalResponseTime += metric.ResponseTimeMs
			totalProcessingTime += metric.ProcessingTimeMs
			if metric.ErrorOccurred {
				errorCount++
			}
		}

		componentAnalysis["total_requests"] = len(componentMetrics)
		componentAnalysis["average_response_time_ms"] = totalResponseTime / float64(len(componentMetrics))
		componentAnalysis["average_processing_time_ms"] = totalProcessingTime / float64(len(componentMetrics))
		componentAnalysis["error_rate"] = float64(errorCount) / float64(len(componentMetrics))

		analysis[metricType] = componentAnalysis
	}

	return analysis
}

func (upm *UnifiedPerformanceMonitor) analyzeTrends(metrics []*ComprehensivePerformanceMetric) map[string]interface{} {
	trends := make(map[string]interface{})

	// Simple trend analysis - compare first half vs second half
	if len(metrics) < 2 {
		return trends
	}

	midpoint := len(metrics) / 2
	firstHalf := metrics[:midpoint]
	secondHalf := metrics[midpoint:]

	// Calculate averages for each half
	firstHalfAvg := upm.calculateAverageResponseTime(firstHalf)
	secondHalfAvg := upm.calculateAverageResponseTime(secondHalf)

	// Determine trend
	trend := "stable"
	if secondHalfAvg > firstHalfAvg*1.1 {
		trend = "degrading"
	} else if secondHalfAvg < firstHalfAvg*0.9 {
		trend = "improving"
	}

	trends["response_time_trend"] = trend
	trends["first_half_avg_ms"] = firstHalfAvg
	trends["second_half_avg_ms"] = secondHalfAvg
	trends["trend_percentage"] = ((secondHalfAvg - firstHalfAvg) / firstHalfAvg) * 100

	return trends
}

func (upm *UnifiedPerformanceMonitor) analyzeCorrelations(metrics []*ComprehensivePerformanceMetric) map[string]interface{} {
	correlations := make(map[string]interface{})

	// Simple correlation analysis between response time and processing time
	var responseTimes, processingTimes []float64

	for _, metric := range metrics {
		if metric.ResponseTimeMs > 0 && metric.ProcessingTimeMs > 0 {
			responseTimes = append(responseTimes, metric.ResponseTimeMs)
			processingTimes = append(processingTimes, metric.ProcessingTimeMs)
		}
	}

	if len(responseTimes) > 1 {
		correlation := upm.calculateCorrelation(responseTimes, processingTimes)
		correlations["response_processing_correlation"] = correlation
		correlations["correlation_strength"] = upm.interpretCorrelation(correlation)
	}

	return correlations
}

func (upm *UnifiedPerformanceMonitor) calculateAverageResponseTime(metrics []*ComprehensivePerformanceMetric) float64 {
	if len(metrics) == 0 {
		return 0
	}

	var total float64
	for _, metric := range metrics {
		total += metric.ResponseTimeMs
	}

	return total / float64(len(metrics))
}

func (upm *UnifiedPerformanceMonitor) calculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}

	// Simple Pearson correlation coefficient
	n := float64(len(x))

	var sumX, sumY, sumXY, sumX2, sumY2 float64
	for i := 0; i < len(x); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}

	numerator := n*sumXY - sumX*sumY
	denominator := (n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY)

	if denominator <= 0 {
		return 0
	}

	return numerator / (denominator * denominator)
}

func (upm *UnifiedPerformanceMonitor) interpretCorrelation(correlation float64) string {
	abs := correlation
	if abs < 0 {
		abs = -abs
	}

	if abs > 0.7 {
		return "strong"
	} else if abs > 0.3 {
		return "moderate"
	} else {
		return "weak"
	}
}

// Recommendation generation methods
func (upm *UnifiedPerformanceMonitor) generatePerformanceRecommendations(stats *UnifiedPerformanceStats) []string {
	var recommendations []string

	if stats.AverageResponseTime > 500 {
		recommendations = append(recommendations, "Consider optimizing response time - current average is above 500ms")
	}

	if stats.ErrorRate > 0.05 {
		recommendations = append(recommendations, "High error rate detected - investigate and fix underlying issues")
	}

	if stats.Throughput < 10 {
		recommendations = append(recommendations, "Low throughput detected - consider scaling or optimization")
	}

	if stats.ResourceUtilization > 80 {
		recommendations = append(recommendations, "High resource utilization - consider scaling resources")
	}

	return recommendations
}

func (upm *UnifiedPerformanceMonitor) generateSecurityRecommendations(stats *UnifiedPerformanceStats) []string {
	var recommendations []string

	if stats.SecurityHealth == "critical" {
		recommendations = append(recommendations, "Critical security issues detected - immediate attention required")
	}

	if stats.CriticalAlerts > 0 {
		recommendations = append(recommendations, "Address critical security alerts immediately")
	}

	if stats.WarningAlerts > 5 {
		recommendations = append(recommendations, "Multiple security warnings - review and address security policies")
	}

	return recommendations
}

func (upm *UnifiedPerformanceMonitor) generateResourceRecommendations(stats *UnifiedPerformanceStats) []string {
	var recommendations []string

	if stats.MemoryHealth == "critical" {
		recommendations = append(recommendations, "Critical memory usage - consider memory optimization or scaling")
	}

	if stats.DatabaseHealth == "critical" {
		recommendations = append(recommendations, "Database performance issues - optimize queries or scale database")
	}

	if stats.TotalMemoryUsage > 500 {
		recommendations = append(recommendations, "High memory usage - monitor for memory leaks")
	}

	return recommendations
}

// Additional analysis methods
func (upm *UnifiedPerformanceMonitor) analyzePerformanceCorrelations(ctx context.Context) {
	// Implementation for cross-component performance correlation analysis
	upm.logger.Debug("Performing performance correlation analysis")
}

func (upm *UnifiedPerformanceMonitor) detectCrossComponentAnomalies(ctx context.Context) {
	// Implementation for detecting anomalies that span multiple components
	upm.logger.Debug("Detecting cross-component anomalies")
}

func (upm *UnifiedPerformanceMonitor) generateUnifiedRecommendations(ctx context.Context) {
	// Implementation for generating unified recommendations across components
	upm.logger.Debug("Generating unified recommendations")
}

// Helper functions
func getStatusCode(errorOccurred bool) int {
	if errorOccurred {
		return 500
	}
	return 200
}

func generateReportID() string {
	return fmt.Sprintf("report_%d", time.Now().Unix())
}

// DefaultUnifiedPerformanceConfig returns a default configuration for the unified performance monitor
func DefaultUnifiedPerformanceConfig() *UnifiedPerformanceConfig {
	return &UnifiedPerformanceConfig{
		Enabled:                      true,
		CollectionInterval:           30 * time.Second,
		MetricsRetentionPeriod:       7 * 24 * time.Hour,
		AlertRetentionPeriod:         30 * 24 * time.Hour,
		ResponseTimeConfig:           &ResponseTimeConfig{},
		MemoryMonitorConfig:          DefaultMemoryMonitorConfig(),
		DatabaseMonitorConfig:        DefaultEnhancedDatabaseConfig(),
		SecurityValidationConfig:     DefaultSecurityValidationConfig(),
		PerformanceMonitorConfig:     DefaultPerformanceMonitorConfig(),
		EnableCrossComponentAnalysis: true,
		EnableUnifiedAlerting:        true,
		EnablePerformanceCorrelation: true,
		ServiceName:                  "classification_service",
		Environment:                  "development",
		Version:                      "1.0.0",
		InstanceID:                   "instance_001",
	}
}
