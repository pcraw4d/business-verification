package classification

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// PerformanceIntegration provides integration between classification services and performance monitoring
type PerformanceIntegration struct {
	performanceMonitor *ComprehensivePerformanceMonitor
	logger             *zap.Logger
	config             *PerformanceIntegrationConfig
}

// PerformanceIntegrationConfig holds configuration for performance monitoring integration
type PerformanceIntegrationConfig struct {
	Enabled                     bool          `json:"enabled"`
	TrackClassificationCalls    bool          `json:"track_classification_calls"`
	TrackSecurityValidation     bool          `json:"track_security_validation"`
	TrackDatabaseQueries        bool          `json:"track_database_queries"`
	TrackMemoryUsage            bool          `json:"track_memory_usage"`
	SampleRate                  float64       `json:"sample_rate"`
	SlowRequestThreshold        time.Duration `json:"slow_request_threshold"`
	MemoryThresholdMB           float64       `json:"memory_threshold_mb"`
	DatabaseQueryThreshold      time.Duration `json:"database_query_threshold"`
	SecurityValidationThreshold time.Duration `json:"security_validation_threshold"`
}

// ClassificationPerformanceContext provides context for tracking classification performance
type ClassificationPerformanceContext struct {
	RequestID              string
	UserID                 string
	SessionID              string
	ServiceName            string
	StartTime              time.Time
	EndTime                time.Time
	ResponseTime           time.Duration
	ProcessingTime         time.Duration
	DatabaseQueryTime      time.Duration
	SecurityValidationTime time.Duration
	MemoryUsageMB          float64
	ConfidenceScore        float64
	KeywordsProcessed      int
	ResultsCount           int
	ErrorOccurred          bool
	ErrorMessage           string
	Metadata               map[string]interface{}
}

// NewPerformanceIntegration creates a new performance monitoring integration
func NewPerformanceIntegration(
	db *sql.DB,
	logger *zap.Logger,
	config *PerformanceIntegrationConfig,
) *PerformanceIntegration {
	if config == nil {
		config = DefaultPerformanceIntegrationConfig()
	}

	// Create performance monitor
	perfConfig := &PerformanceMonitorConfig{
		Enabled:                     config.Enabled,
		CollectionInterval:          30 * time.Second,
		ResponseTimeThreshold:       config.SlowRequestThreshold,
		MemoryUsageThreshold:        config.MemoryThresholdMB,
		DatabaseQueryThreshold:      config.DatabaseQueryThreshold,
		SecurityValidationThreshold: config.SecurityValidationThreshold,
		BufferSize:                  1000,
		AsyncProcessing:             true,
		AlertingEnabled:             true,
		RetentionPeriod:             24 * time.Hour,
	}

	performanceMonitor := NewComprehensivePerformanceMonitor(db, logger, perfConfig)

	return &PerformanceIntegration{
		performanceMonitor: performanceMonitor,
		logger:             logger,
		config:             config,
	}
}

// DefaultPerformanceIntegrationConfig returns default configuration
func DefaultPerformanceIntegrationConfig() *PerformanceIntegrationConfig {
	return &PerformanceIntegrationConfig{
		Enabled:                     true,
		TrackClassificationCalls:    true,
		TrackSecurityValidation:     true,
		TrackDatabaseQueries:        true,
		TrackMemoryUsage:            true,
		SampleRate:                  1.0,
		SlowRequestThreshold:        500 * time.Millisecond,
		MemoryThresholdMB:           512.0,
		DatabaseQueryThreshold:      100 * time.Millisecond,
		SecurityValidationThreshold: 50 * time.Millisecond,
	}
}

// StartClassificationTracking starts tracking a classification request
func (pi *PerformanceIntegration) StartClassificationTracking(
	ctx context.Context,
	requestID, userID, sessionID, serviceName string,
) *ClassificationPerformanceContext {
	if !pi.config.Enabled || !pi.config.TrackClassificationCalls {
		return nil
	}

	startTime := time.Now()

	context := &ClassificationPerformanceContext{
		RequestID:   requestID,
		UserID:      userID,
		SessionID:   sessionID,
		ServiceName: serviceName,
		StartTime:   startTime,
		Metadata:    make(map[string]interface{}),
	}

	pi.logger.Debug("Started classification performance tracking",
		zap.String("request_id", requestID),
		zap.String("service_name", serviceName),
		zap.Time("start_time", startTime))

	return context
}

// EndClassificationTracking ends tracking and records performance metrics
func (pi *PerformanceIntegration) EndClassificationTracking(
	ctx context.Context,
	perfContext *ClassificationPerformanceContext,
	errorOccurred bool,
	errorMessage string,
) {
	if perfContext == nil || !pi.config.Enabled {
		return
	}

	perfContext.EndTime = time.Now()
	perfContext.ResponseTime = perfContext.EndTime.Sub(perfContext.StartTime)
	perfContext.ErrorOccurred = errorOccurred
	perfContext.ErrorMessage = errorMessage

	// Record performance metrics
	pi.recordClassificationMetrics(ctx, perfContext)

	pi.logger.Debug("Ended classification performance tracking",
		zap.String("request_id", perfContext.RequestID),
		zap.Duration("response_time", perfContext.ResponseTime),
		zap.Duration("processing_time", perfContext.ProcessingTime),
		zap.Duration("database_query_time", perfContext.DatabaseQueryTime),
		zap.Duration("security_validation_time", perfContext.SecurityValidationTime),
		zap.Float64("memory_usage_mb", perfContext.MemoryUsageMB),
		zap.Float64("confidence_score", perfContext.ConfidenceScore),
		zap.Int("keywords_processed", perfContext.KeywordsProcessed),
		zap.Int("results_count", perfContext.ResultsCount),
		zap.Bool("error_occurred", errorOccurred))
}

// recordClassificationMetrics records comprehensive classification performance metrics
func (pi *PerformanceIntegration) recordClassificationMetrics(
	ctx context.Context,
	perfContext *ClassificationPerformanceContext,
) {
	// Record response time metric
	responseTimeMetric := &ComprehensivePerformanceMetric{
		ID:               fmt.Sprintf("response_time_%s_%d", perfContext.RequestID, perfContext.StartTime.UnixNano()),
		Timestamp:        perfContext.EndTime,
		MetricType:       "response_time",
		ServiceName:      perfContext.ServiceName,
		ResponseTimeMs:   float64(perfContext.ResponseTime.Milliseconds()),
		ProcessingTimeMs: float64(perfContext.ProcessingTime.Milliseconds()),
		RequestID:        perfContext.RequestID,
		UserID:           perfContext.UserID,
		SessionID:        perfContext.SessionID,
		ErrorOccurred:    perfContext.ErrorOccurred,
		ErrorMessage:     perfContext.ErrorMessage,
		Metadata:         perfContext.Metadata,
	}

	if err := pi.performanceMonitor.RecordPerformanceMetric(ctx, responseTimeMetric); err != nil {
		pi.logger.Error("Failed to record response time metric",
			zap.String("request_id", perfContext.RequestID),
			zap.Error(err))
	}

	// Record classification-specific metrics
	classificationMetric := &ComprehensivePerformanceMetric{
		ID:                     fmt.Sprintf("classification_%s_%d", perfContext.RequestID, perfContext.StartTime.UnixNano()),
		Timestamp:              perfContext.EndTime,
		MetricType:             "classification",
		ServiceName:            perfContext.ServiceName,
		ClassificationAccuracy: perfContext.ConfidenceScore,
		ConfidenceScore:        perfContext.ConfidenceScore,
		KeywordsProcessed:      perfContext.KeywordsProcessed,
		RequestID:              perfContext.RequestID,
		UserID:                 perfContext.UserID,
		SessionID:              perfContext.SessionID,
		ErrorOccurred:          perfContext.ErrorOccurred,
		ErrorMessage:           perfContext.ErrorMessage,
		Metadata:               perfContext.Metadata,
	}

	if err := pi.performanceMonitor.RecordPerformanceMetric(ctx, classificationMetric); err != nil {
		pi.logger.Error("Failed to record classification metric",
			zap.String("request_id", perfContext.RequestID),
			zap.Error(err))
	}
}

// UpdateClassificationResults updates the classification results in the performance context
func (pi *PerformanceIntegration) UpdateClassificationResults(
	perfContext *ClassificationPerformanceContext,
	confidenceScore float64,
	keywordsProcessed int,
	resultsCount int,
) {
	if perfContext == nil {
		return
	}

	perfContext.ConfidenceScore = confidenceScore
	perfContext.KeywordsProcessed = keywordsProcessed
	perfContext.ResultsCount = resultsCount

	// Update metadata
	perfContext.Metadata["confidence_score"] = confidenceScore
	perfContext.Metadata["keywords_processed"] = keywordsProcessed
	perfContext.Metadata["results_count"] = resultsCount
}

// GetPerformanceMetrics returns performance metrics for the integration
func (pi *PerformanceIntegration) GetPerformanceMetrics(
	ctx context.Context,
	startTime, endTime time.Time,
	metricType string,
) ([]*ComprehensivePerformanceMetric, error) {
	if !pi.config.Enabled {
		return nil, fmt.Errorf("performance monitoring is disabled")
	}

	return pi.performanceMonitor.GetPerformanceMetrics(ctx, startTime, endTime, metricType)
}

// GetPerformanceAlerts returns performance alerts
func (pi *PerformanceIntegration) GetPerformanceAlerts(
	ctx context.Context,
	resolved bool,
) ([]*ComprehensivePerformanceAlert, error) {
	if !pi.config.Enabled {
		return nil, fmt.Errorf("performance monitoring is disabled")
	}

	return pi.performanceMonitor.GetPerformanceAlerts(ctx, resolved)
}

// GetPerformanceSummary returns a performance summary
func (pi *PerformanceIntegration) GetPerformanceSummary(ctx context.Context) (map[string]interface{}, error) {
	if !pi.config.Enabled {
		return nil, fmt.Errorf("performance monitoring is disabled")
	}

	return pi.performanceMonitor.GetPerformanceSummary(ctx)
}

// Stop stops the performance monitoring integration
func (pi *PerformanceIntegration) Stop() {
	if pi.performanceMonitor != nil {
		pi.performanceMonitor.Stop()
	}
}
