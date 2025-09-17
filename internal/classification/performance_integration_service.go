package classification

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PerformanceIntegrationService provides a high-level interface for integrating all performance monitoring components
type PerformanceIntegrationService struct {
	unifiedMonitor *UnifiedPerformanceMonitor
	logger         *zap.Logger
	config         *PerformanceIntegrationServiceConfig
	mu             sync.RWMutex
}

// PerformanceIntegrationServiceConfig represents the configuration for the performance integration service
type PerformanceIntegrationServiceConfig struct {
	// Service settings
	Enabled                  bool          `json:"enabled"`
	AutoStart                bool          `json:"auto_start"`
	HealthCheckInterval      time.Duration `json:"health_check_interval"`
	ReportGenerationInterval time.Duration `json:"report_generation_interval"`

	// Monitoring settings
	EnableRealTimeMonitoring bool `json:"enable_real_time_monitoring"`
	EnableHistoricalAnalysis bool `json:"enable_historical_analysis"`
	EnablePredictiveAnalysis bool `json:"enable_predictive_analysis"`

	// Alerting settings
	EnableAlerting  bool               `json:"enable_alerting"`
	AlertChannels   []string           `json:"alert_channels"` // "log", "webhook", "email"
	AlertThresholds map[string]float64 `json:"alert_thresholds"`

	// Reporting settings
	EnableAutoReporting   bool     `json:"enable_auto_reporting"`
	ReportFormats         []string `json:"report_formats"` // "json", "html", "pdf"
	ReportStorageLocation string   `json:"report_storage_location"`

	// Service identification
	ServiceName string `json:"service_name"`
	Environment string `json:"environment"`
	Version     string `json:"version"`
	InstanceID  string `json:"instance_id"`
}

// PerformanceIntegrationHealth represents the health status of the performance integration service
type PerformanceIntegrationHealth struct {
	ServiceName     string                 `json:"service_name"`
	Status          string                 `json:"status"` // "healthy", "degraded", "unhealthy"
	LastHealthCheck time.Time              `json:"last_health_check"`
	ComponentHealth map[string]string      `json:"component_health"`
	OverallScore    float64                `json:"overall_score"`
	ActiveIssues    []string               `json:"active_issues"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// NewPerformanceIntegrationService creates a new performance integration service
func NewPerformanceIntegrationService(db *sql.DB, logger *zap.Logger, config *PerformanceIntegrationServiceConfig) (*PerformanceIntegrationService, error) {
	if config == nil {
		config = DefaultPerformanceIntegrationConfig()
	}

	// Create unified performance monitor
	unifiedConfig := &UnifiedPerformanceConfig{
		Enabled:                      config.Enabled,
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
		ServiceName:                  config.ServiceName,
		Environment:                  config.Environment,
		Version:                      config.Version,
		InstanceID:                   config.InstanceID,
	}

	unifiedMonitor, err := NewUnifiedPerformanceMonitor(db, logger, unifiedConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create unified performance monitor: %w", err)
	}

	service := &PerformanceIntegrationService{
		unifiedMonitor: unifiedMonitor,
		logger:         logger,
		config:         config,
	}

	// Auto-start if configured
	if config.AutoStart {
		if err := service.Start(); err != nil {
			return nil, fmt.Errorf("failed to auto-start service: %w", err)
		}
	}

	return service, nil
}

// Start starts the performance integration service
func (pis *PerformanceIntegrationService) Start() error {
	pis.mu.Lock()
	defer pis.mu.Unlock()

	pis.logger.Info("Starting performance integration service",
		zap.String("service", pis.config.ServiceName),
		zap.String("environment", pis.config.Environment))

	// Start the unified monitor
	if err := pis.unifiedMonitor.Start(); err != nil {
		return fmt.Errorf("failed to start unified monitor: %w", err)
	}

	// Start health monitoring if enabled
	if pis.config.HealthCheckInterval > 0 {
		go pis.healthMonitoringLoop()
	}

	// Start auto-reporting if enabled
	if pis.config.EnableAutoReporting && pis.config.ReportGenerationInterval > 0 {
		go pis.autoReportingLoop()
	}

	pis.logger.Info("Performance integration service started successfully")
	return nil
}

// Stop stops the performance integration service
func (pis *PerformanceIntegrationService) Stop() {
	pis.mu.Lock()
	defer pis.mu.Unlock()

	pis.logger.Info("Stopping performance integration service")

	// Stop the unified monitor
	pis.unifiedMonitor.Stop()

	pis.logger.Info("Performance integration service stopped")
}

// RecordClassificationOperation records a complete classification operation with all performance metrics
func (pis *PerformanceIntegrationService) RecordClassificationOperation(ctx context.Context, operation *ClassificationOperation) error {
	// Create performance context
	perfContext := &ClassificationPerformanceContext{
		RequestID:         operation.RequestID,
		ServiceName:       operation.ServiceName,
		StartTime:         operation.StartTime,
		EndTime:           operation.EndTime,
		ResponseTime:      operation.EndTime.Sub(operation.StartTime),
		ProcessingTime:    operation.ProcessingTime,
		ConfidenceScore:   operation.ConfidenceScore,
		KeywordsProcessed: operation.KeywordsCount,
		ResultsCount:      operation.ResultsCount,
		ErrorOccurred:     operation.ErrorOccurred,
		ErrorMessage:      operation.ErrorMessage,
		Metadata:          operation.Metadata,
	}

	// Record in unified monitor
	if err := pis.unifiedMonitor.RecordClassificationMetrics(ctx, perfContext); err != nil {
		pis.logger.Error("Failed to record classification metrics", zap.Error(err))
		return err
	}

	// Record database queries if any
	for _, query := range operation.DatabaseQueries {
		pis.unifiedMonitor.RecordDatabaseQuery(
			ctx,
			query.Query,
			query.Duration,
			query.RowsReturned,
			query.RowsExamined,
			query.ErrorOccurred,
			query.QueryID,
		)
	}

	// Record security validations if any
	for _, validation := range operation.SecurityValidations {
		pis.unifiedMonitor.RecordSecurityValidation(ctx, validation)
	}

	return nil
}

// GetSystemHealth returns the current health status of the performance monitoring system
func (pis *PerformanceIntegrationService) GetSystemHealth() *PerformanceIntegrationHealth {
	health := &PerformanceIntegrationHealth{
		ServiceName:     pis.config.ServiceName,
		LastHealthCheck: time.Now(),
		ComponentHealth: make(map[string]string),
		ActiveIssues:    []string{},
		Recommendations: []string{},
		Metadata:        make(map[string]interface{}),
	}

	// Get unified stats
	stats := pis.unifiedMonitor.GetUnifiedStats()
	health.OverallScore = stats.SystemHealthScore

	// Determine overall status
	if stats.SystemHealthScore >= 90 {
		health.Status = "healthy"
	} else if stats.SystemHealthScore >= 70 {
		health.Status = "degraded"
	} else {
		health.Status = "unhealthy"
	}

	// Check component health
	health.ComponentHealth["response_time"] = stats.ResponseTimeHealth
	health.ComponentHealth["memory"] = stats.MemoryHealth
	health.ComponentHealth["database"] = stats.DatabaseHealth
	health.ComponentHealth["security"] = stats.SecurityHealth

	// Identify active issues
	if stats.ResponseTimeHealth == "critical" {
		health.ActiveIssues = append(health.ActiveIssues, "Critical response time issues")
	}
	if stats.MemoryHealth == "critical" {
		health.ActiveIssues = append(health.ActiveIssues, "Critical memory usage")
	}
	if stats.DatabaseHealth == "critical" {
		health.ActiveIssues = append(health.ActiveIssues, "Critical database performance issues")
	}
	if stats.SecurityHealth == "critical" {
		health.ActiveIssues = append(health.ActiveIssues, "Critical security issues")
	}

	// Generate recommendations
	health.Recommendations = pis.generateHealthRecommendations(stats)

	// Add metadata
	health.Metadata["total_requests"] = stats.TotalRequests
	health.Metadata["average_response_time_ms"] = stats.AverageResponseTime
	health.Metadata["active_alerts"] = stats.ActiveAlerts
	health.Metadata["critical_alerts"] = stats.CriticalAlerts

	return health
}

// GetPerformanceReport generates a comprehensive performance report
func (pis *PerformanceIntegrationService) GetPerformanceReport(ctx context.Context, reportPeriod time.Duration) (*UnifiedPerformanceReport, error) {
	return pis.unifiedMonitor.GenerateUnifiedReport(ctx, reportPeriod)
}

// GetActiveAlerts returns currently active performance alerts
func (pis *PerformanceIntegrationService) GetActiveAlerts(ctx context.Context, limit int) ([]*UnifiedPerformanceAlert, error) {
	return pis.unifiedMonitor.GetUnifiedAlerts(ctx, false, limit)
}

// GetPerformanceMetrics returns performance metrics for a given time range
func (pis *PerformanceIntegrationService) GetPerformanceMetrics(ctx context.Context, startTime, endTime time.Time, metricType string) ([]*ComprehensivePerformanceMetric, error) {
	return pis.unifiedMonitor.comprehensiveMonitor.GetPerformanceMetrics(ctx, startTime, endTime, metricType)
}

// GetUnifiedStats returns current unified performance statistics
func (pis *PerformanceIntegrationService) GetUnifiedStats() *UnifiedPerformanceStats {
	return pis.unifiedMonitor.GetUnifiedStats()
}

// healthMonitoringLoop continuously monitors the health of the performance monitoring system
func (pis *PerformanceIntegrationService) healthMonitoringLoop() {
	ticker := time.NewTicker(pis.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			health := pis.GetSystemHealth()

			// Log health status
			pis.logger.Info("Performance monitoring system health check",
				zap.String("status", health.Status),
				zap.Float64("score", health.OverallScore),
				zap.Int("active_issues", len(health.ActiveIssues)))

			// Send alerts if unhealthy
			if health.Status == "unhealthy" && pis.config.EnableAlerting {
				pis.sendHealthAlert(health)
			}
		}
	}
}

// autoReportingLoop automatically generates and stores performance reports
func (pis *PerformanceIntegrationService) autoReportingLoop() {
	ticker := time.NewTicker(pis.config.ReportGenerationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()

			// Generate report for the last interval period
			report, err := pis.GetPerformanceReport(ctx, pis.config.ReportGenerationInterval)
			if err != nil {
				pis.logger.Error("Failed to generate auto-report", zap.Error(err))
				continue
			}

			// Store report if configured
			if pis.config.ReportStorageLocation != "" {
				if err := pis.storeReport(report); err != nil {
					pis.logger.Error("Failed to store auto-report", zap.Error(err))
				}
			}

			pis.logger.Info("Auto-report generated successfully",
				zap.String("report_id", report.ReportID),
				zap.Duration("period", report.ReportPeriod))
		}
	}
}

// sendHealthAlert sends an alert about system health issues
func (pis *PerformanceIntegrationService) sendHealthAlert(health *PerformanceIntegrationHealth) {
	alert := &UnifiedPerformanceAlert{
		ID:                 fmt.Sprintf("health_alert_%d", time.Now().Unix()),
		Timestamp:          time.Now(),
		AlertType:          "system",
		Severity:           "high",
		Title:              "Performance Monitoring System Health Alert",
		Message:            fmt.Sprintf("System health is %s with score %.1f", health.Status, health.OverallScore),
		AffectedComponents: []string{"system"},
		Recommendations:    health.Recommendations,
		Metadata: map[string]interface{}{
			"active_issues":    health.ActiveIssues,
			"component_health": health.ComponentHealth,
		},
	}

	// Log the alert
	pis.logger.Error("Performance monitoring system health alert",
		zap.String("alert_id", alert.ID),
		zap.String("severity", alert.Severity),
		zap.String("message", alert.Message),
		zap.Strings("active_issues", health.ActiveIssues))

	// TODO: Implement additional alert channels (webhook, email, etc.)
}

// storeReport stores a performance report to the configured location
func (pis *PerformanceIntegrationService) storeReport(report *UnifiedPerformanceReport) error {
	// TODO: Implement report storage (file system, database, cloud storage, etc.)
	pis.logger.Info("Storing performance report",
		zap.String("report_id", report.ReportID),
		zap.String("location", pis.config.ReportStorageLocation))
	return nil
}

// generateHealthRecommendations generates recommendations based on system health
func (pis *PerformanceIntegrationService) generateHealthRecommendations(stats *UnifiedPerformanceStats) []string {
	var recommendations []string

	// Performance recommendations
	if stats.AverageResponseTime > 500 {
		recommendations = append(recommendations, "Optimize response time - current average is above 500ms")
	}

	if stats.ErrorRate > 0.05 {
		recommendations = append(recommendations, "Address high error rate - investigate underlying issues")
	}

	// Resource recommendations
	if stats.MemoryHealth == "critical" {
		recommendations = append(recommendations, "Critical memory usage - consider scaling or optimization")
	}

	if stats.DatabaseHealth == "critical" {
		recommendations = append(recommendations, "Critical database performance - optimize queries or scale")
	}

	// Security recommendations
	if stats.SecurityHealth == "critical" {
		recommendations = append(recommendations, "Critical security issues - immediate attention required")
	}

	if stats.CriticalAlerts > 0 {
		recommendations = append(recommendations, "Address critical alerts immediately")
	}

	return recommendations
}

// ClassificationOperation represents a complete classification operation with all associated metrics
type ClassificationOperation struct {
	RequestID           string                              `json:"request_id"`
	ServiceName         string                              `json:"service_name"`
	Endpoint            string                              `json:"endpoint"`
	Method              string                              `json:"method"`
	StartTime           time.Time                           `json:"start_time"`
	EndTime             time.Time                           `json:"end_time"`
	ProcessingTime      time.Duration                       `json:"processing_time"`
	ConfidenceScore     float64                             `json:"confidence_score"`
	KeywordsCount       int                                 `json:"keywords_count"`
	ResultsCount        int                                 `json:"results_count"`
	CacheHitRatio       float64                             `json:"cache_hit_ratio"`
	ErrorOccurred       bool                                `json:"error_occurred"`
	ErrorMessage        string                              `json:"error_message,omitempty"`
	DatabaseQueries     []DatabaseQueryExecution            `json:"database_queries,omitempty"`
	SecurityValidations []*AdvancedSecurityValidationResult `json:"security_validations,omitempty"`
	Metadata            map[string]interface{}              `json:"metadata,omitempty"`
}

// DatabaseQueryExecution represents a database query execution within a classification operation
type DatabaseQueryExecution struct {
	QueryID       string        `json:"query_id"`
	Query         string        `json:"query"`
	Duration      time.Duration `json:"duration"`
	RowsReturned  int64         `json:"rows_returned"`
	RowsExamined  int64         `json:"rows_examined"`
	ErrorOccurred bool          `json:"error_occurred"`
	ErrorMessage  string        `json:"error_message,omitempty"`
}

// DefaultPerformanceIntegrationConfig returns a default configuration for the performance integration service
func DefaultPerformanceIntegrationConfig() *PerformanceIntegrationServiceConfig {
	return &PerformanceIntegrationServiceConfig{
		Enabled:                  true,
		AutoStart:                true,
		HealthCheckInterval:      5 * time.Minute,
		ReportGenerationInterval: 1 * time.Hour,
		EnableRealTimeMonitoring: true,
		EnableHistoricalAnalysis: true,
		EnablePredictiveAnalysis: false,
		EnableAlerting:           true,
		AlertChannels:            []string{"log"},
		AlertThresholds: map[string]float64{
			"response_time_ms": 1000,
			"error_rate":       0.05,
			"memory_usage_mb":  1000,
		},
		EnableAutoReporting:   true,
		ReportFormats:         []string{"json"},
		ReportStorageLocation: "./reports",
		ServiceName:           "classification_service",
		Environment:           "development",
		Version:               "1.0.0",
		InstanceID:            "instance_001",
	}
}
