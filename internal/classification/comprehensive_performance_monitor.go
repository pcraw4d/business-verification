package classification

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ComprehensivePerformanceMonitor provides unified performance monitoring for the classification system
type ComprehensivePerformanceMonitor struct {
	logger              *zap.Logger
	db                  *sql.DB
	responseTimeTracker *ResponseTimeTracker
	databaseMonitor     *DatabaseMonitor
	memoryMonitor       *MemoryMonitor
	queryMonitor        *QueryPerformanceMonitor
	securityMonitor     *SecurityValidationMonitor

	// Configuration
	config *PerformanceMonitorConfig

	// Metrics storage
	metrics map[string]*ComprehensivePerformanceMetric
	alerts  map[string]*ComprehensivePerformanceAlert

	// Thread safety
	mu sync.RWMutex

	// Processing
	stopCh    chan struct{}
	metricsCh chan *ComprehensivePerformanceMetric
	workerWg  sync.WaitGroup
}

// PerformanceMonitorConfig holds configuration for comprehensive performance monitoring
type PerformanceMonitorConfig struct {
	Enabled                     bool          `json:"enabled"`
	CollectionInterval          time.Duration `json:"collection_interval"`
	ResponseTimeThreshold       time.Duration `json:"response_time_threshold"`
	MemoryUsageThreshold        float64       `json:"memory_usage_threshold_mb"`
	DatabaseQueryThreshold      time.Duration `json:"database_query_threshold"`
	SecurityValidationThreshold time.Duration `json:"security_validation_threshold"`
	BufferSize                  int           `json:"buffer_size"`
	AsyncProcessing             bool          `json:"async_processing"`
	AlertingEnabled             bool          `json:"alerting_enabled"`
	RetentionPeriod             time.Duration `json:"retention_period"`
}

// ComprehensivePerformanceMetric represents a comprehensive performance metric
type ComprehensivePerformanceMetric struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	MetricType  string    `json:"metric_type"` // "response_time", "memory", "database", "security"
	ServiceName string    `json:"service_name"`
	Endpoint    string    `json:"endpoint,omitempty"`
	Method      string    `json:"method,omitempty"`

	// Response Time Metrics
	ResponseTimeMs   float64 `json:"response_time_ms,omitempty"`
	ProcessingTimeMs float64 `json:"processing_time_ms,omitempty"`

	// Memory Metrics
	MemoryUsageMB     float64 `json:"memory_usage_mb,omitempty"`
	MemoryAllocatedMB float64 `json:"memory_allocated_mb,omitempty"`
	GCPauseTimeMs     float64 `json:"gc_pause_time_ms,omitempty"`
	GoroutineCount    int     `json:"goroutine_count,omitempty"`

	// Database Metrics
	DatabaseQueryTimeMs     float64 `json:"database_query_time_ms,omitempty"`
	DatabaseQueryCount      int     `json:"database_query_count,omitempty"`
	DatabaseConnectionCount int     `json:"database_connection_count,omitempty"`
	DatabaseCacheHitRatio   float64 `json:"database_cache_hit_ratio,omitempty"`

	// Security Metrics
	SecurityValidationTimeMs  float64 `json:"security_validation_time_ms,omitempty"`
	TrustedDataSourceCount    int     `json:"trusted_data_source_count,omitempty"`
	WebsiteVerificationTimeMs float64 `json:"website_verification_time_ms,omitempty"`

	// Classification Metrics
	ClassificationAccuracy float64 `json:"classification_accuracy,omitempty"`
	ConfidenceScore        float64 `json:"confidence_score,omitempty"`
	KeywordsProcessed      int     `json:"keywords_processed,omitempty"`

	// System Metrics
	CPUUsagePercent float64 `json:"cpu_usage_percent,omitempty"`
	LoadAverage     float64 `json:"load_average,omitempty"`

	// Metadata
	RequestID     string                 `json:"request_id,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
	SessionID     string                 `json:"session_id,omitempty"`
	ErrorOccurred bool                   `json:"error_occurred"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ComprehensivePerformanceAlert represents a performance alert
type ComprehensivePerformanceAlert struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	AlertType   string                 `json:"alert_type"`
	Severity    string                 `json:"severity"` // "low", "medium", "high", "critical"
	MetricType  string                 `json:"metric_type"`
	Threshold   float64                `json:"threshold"`
	ActualValue float64                `json:"actual_value"`
	Message     string                 `json:"message"`
	ServiceName string                 `json:"service_name"`
	Endpoint    string                 `json:"endpoint,omitempty"`
	Resolved    bool                   `json:"resolved"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ResponseTimeTracker tracks response times for API endpoints
type ResponseTimeTracker struct {
	config *ResponseTimeConfig
	logger *zap.Logger
	mu     sync.RWMutex
	stats  map[string]*ResponseTimeStats
}

// ResponseTimeConfig holds configuration for response time tracking
type ResponseTimeConfig struct {
	Enabled              bool          `json:"enabled"`
	SampleRate           float64       `json:"sample_rate"`
	SlowRequestThreshold time.Duration `json:"slow_request_threshold"`
	BufferSize           int           `json:"buffer_size"`
	AsyncProcessing      bool          `json:"async_processing"`
}

// ResponseTimeStats represents response time statistics
type ResponseTimeStats struct {
	Endpoint         string        `json:"endpoint"`
	Method           string        `json:"method"`
	RequestCount     int64         `json:"request_count"`
	TotalTime        time.Duration `json:"total_time"`
	AverageTime      time.Duration `json:"average_time"`
	MinTime          time.Duration `json:"min_time"`
	MaxTime          time.Duration `json:"max_time"`
	P50Time          time.Duration `json:"p50_time"`
	P95Time          time.Duration `json:"p95_time"`
	P99Time          time.Duration `json:"p99_time"`
	SlowRequestCount int64         `json:"slow_request_count"`
	ErrorCount       int64         `json:"error_count"`
	LastUpdated      time.Time     `json:"last_updated"`
}

// MemoryMonitor tracks memory usage and garbage collection
type MemoryMonitor struct {
	logger *zap.Logger
	mu     sync.RWMutex
	stats  *MemoryStats
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	Timestamp        time.Time `json:"timestamp"`
	AllocatedMB      float64   `json:"allocated_mb"`
	TotalAllocatedMB float64   `json:"total_allocated_mb"`
	SystemMB         float64   `json:"system_mb"`
	NumGC            uint32    `json:"num_gc"`
	GCPauseTimeMs    float64   `json:"gc_pause_time_ms"`
	HeapObjects      uint64    `json:"heap_objects"`
	StackInUseMB     float64   `json:"stack_in_use_mb"`
	GoroutineCount   int       `json:"goroutine_count"`
	LastGC           time.Time `json:"last_gc"`
}

// DatabaseMonitor tracks database performance (reusing existing implementation)
type DatabaseMonitor struct {
	db                 *sql.DB
	config             *DatabaseConfig
	metrics            []*DatabaseMetrics
	maxMetrics         int
	slowQueryThreshold time.Duration
	mu                 sync.RWMutex
	queryStats         map[string]*QueryStats
}

// DatabaseConfig holds configuration for database monitoring
type DatabaseConfig struct {
	Enabled            bool          `json:"enabled"`
	SlowQueryThreshold time.Duration `json:"slow_query_threshold"`
	MaxMetricsStored   int           `json:"max_metrics_stored"`
	CollectionInterval time.Duration `json:"collection_interval"`
}

// DatabaseMetrics represents database performance metrics (reusing existing)
type DatabaseMetrics struct {
	Timestamp         time.Time              `json:"timestamp"`
	ConnectionCount   int                    `json:"connection_count"`
	ActiveConnections int                    `json:"active_connections"`
	IdleConnections   int                    `json:"idle_connections"`
	MaxConnections    int                    `json:"max_connections"`
	QueryCount        int64                  `json:"query_count"`
	SlowQueryCount    int64                  `json:"slow_query_count"`
	ErrorCount        int64                  `json:"error_count"`
	AvgQueryTime      float64                `json:"avg_query_time_ms"`
	MaxQueryTime      float64                `json:"max_query_time_ms"`
	DatabaseSize      int64                  `json:"database_size_bytes"`
	TableSizes        map[string]int64       `json:"table_sizes"`
	IndexSizes        map[string]int64       `json:"index_sizes"`
	LockCount         int                    `json:"lock_count"`
	DeadlockCount     int                    `json:"deadlock_count"`
	CacheHitRatio     float64                `json:"cache_hit_ratio"`
	Uptime            time.Duration          `json:"uptime"`
	LastBackup        *time.Time             `json:"last_backup,omitempty"`
	BackupSize        int64                  `json:"backup_size_bytes"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// QueryStats tracks statistics for individual queries (reusing existing)
type QueryStats struct {
	Count      int64         `json:"count"`
	TotalTime  time.Duration `json:"total_time"`
	AvgTime    time.Duration `json:"avg_time"`
	MinTime    time.Duration `json:"min_time"`
	MaxTime    time.Duration `json:"max_time"`
	ErrorCount int64         `json:"error_count"`
	LastSeen   time.Time     `json:"last_seen"`
}

// QueryPerformanceMonitor tracks query performance (reusing existing)
type QueryPerformanceMonitor struct {
	db     *sql.DB
	logger *zap.Logger
	mu     sync.RWMutex
	stats  map[string]*QueryPerformanceStats
}

// ComprehensiveQueryPerformanceStats represents query performance statistics
type ComprehensiveQueryPerformanceStats struct {
	QueryID              int64     `json:"query_id"`
	QueryText            string    `json:"query_text"`
	ExecutionCount       int64     `json:"execution_count"`
	TotalExecutionTime   float64   `json:"total_execution_time_ms"`
	AverageExecutionTime float64   `json:"average_execution_time_ms"`
	MinExecutionTime     float64   `json:"min_execution_time_ms"`
	MaxExecutionTime     float64   `json:"max_execution_time_ms"`
	RowsReturned         int64     `json:"rows_returned"`
	RowsExamined         int64     `json:"rows_examined"`
	IndexUsageScore      float64   `json:"index_usage_score"`
	CacheHitRatio        float64   `json:"cache_hit_ratio"`
	PerformanceCategory  string    `json:"performance_category"`
	OptimizationPriority string    `json:"optimization_priority"`
	LastExecuted         time.Time `json:"last_executed"`
}

// SecurityValidationMonitor tracks security validation performance
type SecurityValidationMonitor struct {
	logger *zap.Logger
	mu     sync.RWMutex
	stats  *SecurityValidationStats
}

// SecurityValidationStats represents security validation statistics
type SecurityValidationStats struct {
	Timestamp                    time.Time `json:"timestamp"`
	TotalValidations             int64     `json:"total_validations"`
	TrustedDataSourceValidations int64     `json:"trusted_data_source_validations"`
	WebsiteVerificationCount     int64     `json:"website_verification_count"`
	AverageValidationTimeMs      float64   `json:"average_validation_time_ms"`
	AverageWebsiteVerificationMs float64   `json:"average_website_verification_ms"`
	SecurityViolations           int64     `json:"security_violations"`
	TrustedDataSourceRate        float64   `json:"trusted_data_source_rate"`
	WebsiteVerificationRate      float64   `json:"website_verification_rate"`
}

// NewComprehensivePerformanceMonitor creates a new comprehensive performance monitor
func NewComprehensivePerformanceMonitor(
	db *sql.DB,
	logger *zap.Logger,
	config *PerformanceMonitorConfig,
) *ComprehensivePerformanceMonitor {
	if config == nil {
		config = DefaultPerformanceMonitorConfig()
	}

	monitor := &ComprehensivePerformanceMonitor{
		logger:    logger,
		db:        db,
		config:    config,
		metrics:   make(map[string]*ComprehensivePerformanceMetric),
		alerts:    make(map[string]*ComprehensivePerformanceAlert),
		stopCh:    make(chan struct{}),
		metricsCh: make(chan *ComprehensivePerformanceMetric, config.BufferSize),
	}

	// Initialize sub-monitors
	monitor.initializeSubMonitors()

	// Start background processing if enabled
	if config.AsyncProcessing {
		monitor.startWorker()
	}

	// Start periodic collection
	go monitor.startPeriodicCollection()

	return monitor
}

// DefaultPerformanceMonitorConfig returns default configuration
func DefaultPerformanceMonitorConfig() *PerformanceMonitorConfig {
	return &PerformanceMonitorConfig{
		Enabled:                     true,
		CollectionInterval:          30 * time.Second,
		ResponseTimeThreshold:       500 * time.Millisecond,
		MemoryUsageThreshold:        512.0, // 512MB
		DatabaseQueryThreshold:      100 * time.Millisecond,
		SecurityValidationThreshold: 50 * time.Millisecond,
		BufferSize:                  1000,
		AsyncProcessing:             true,
		AlertingEnabled:             true,
		RetentionPeriod:             24 * time.Hour,
	}
}

// initializeSubMonitors initializes all sub-monitors
func (cpm *ComprehensivePerformanceMonitor) initializeSubMonitors() {
	// Initialize response time tracker
	responseTimeConfig := &ResponseTimeConfig{
		Enabled:              true,
		SampleRate:           1.0,
		SlowRequestThreshold: cpm.config.ResponseTimeThreshold,
		BufferSize:           cpm.config.BufferSize,
		AsyncProcessing:      cpm.config.AsyncProcessing,
	}
	cpm.responseTimeTracker = NewResponseTimeTracker(responseTimeConfig, cpm.logger)

	// Initialize database monitor
	databaseConfig := &DatabaseConfig{
		Enabled:            true,
		SlowQueryThreshold: cpm.config.DatabaseQueryThreshold,
		MaxMetricsStored:   1000,
		CollectionInterval: cpm.config.CollectionInterval,
	}
	cpm.databaseMonitor = NewDatabaseMonitor(cpm.db, databaseConfig)

	// Initialize memory monitor
	cpm.memoryMonitor = NewMemoryMonitor(cpm.logger)

	// Initialize query performance monitor
	cpm.queryMonitor = NewQueryPerformanceMonitor(cpm.db, cpm.logger)

	// Initialize security validation monitor
	cpm.securityMonitor = NewSecurityValidationMonitor(cpm.logger)
}

// RecordPerformanceMetric records a performance metric
func (cpm *ComprehensivePerformanceMonitor) RecordPerformanceMetric(ctx context.Context, metric *ComprehensivePerformanceMetric) error {
	if !cpm.config.Enabled {
		return nil
	}

	if metric == nil {
		return fmt.Errorf("metric cannot be nil")
	}

	// Set timestamp if not provided
	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}

	// Generate ID if not provided
	if metric.ID == "" {
		metric.ID = fmt.Sprintf("%s_%d", metric.MetricType, metric.Timestamp.UnixNano())
	}

	if cpm.config.AsyncProcessing {
		select {
		case cpm.metricsCh <- metric:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Channel full, process synchronously
			cpm.logger.Warn("performance metrics channel full, processing synchronously")
			return cpm.processMetric(metric)
		}
	}

	return cpm.processMetric(metric)
}

// processMetric processes a performance metric
func (cpm *ComprehensivePerformanceMonitor) processMetric(metric *ComprehensivePerformanceMetric) error {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	// Store metric
	cpm.metrics[metric.ID] = metric

	// Check for alerts
	if cpm.config.AlertingEnabled {
		cpm.checkAlerts(metric)
	}

	// Store in database
	if err := cpm.storeMetricInDatabase(metric); err != nil {
		cpm.logger.Error("failed to store metric in database",
			zap.String("metric_id", metric.ID),
			zap.Error(err))
	}

	return nil
}

// checkAlerts checks if a metric triggers any alerts
func (cpm *ComprehensivePerformanceMonitor) checkAlerts(metric *ComprehensivePerformanceMetric) {
	var alerts []*ComprehensivePerformanceAlert

	switch metric.MetricType {
	case "response_time":
		if metric.ResponseTimeMs > float64(cpm.config.ResponseTimeThreshold.Milliseconds()) {
			alerts = append(alerts, &ComprehensivePerformanceAlert{
				ID:          fmt.Sprintf("response_time_%s_%d", metric.Endpoint, metric.Timestamp.Unix()),
				Timestamp:   time.Now(),
				AlertType:   "slow_response",
				Severity:    cpm.determineSeverity(metric.ResponseTimeMs, float64(cpm.config.ResponseTimeThreshold.Milliseconds())),
				MetricType:  "response_time",
				Threshold:   float64(cpm.config.ResponseTimeThreshold.Milliseconds()),
				ActualValue: metric.ResponseTimeMs,
				Message: fmt.Sprintf("Response time %fms exceeds threshold %fms for endpoint %s",
					metric.ResponseTimeMs, float64(cpm.config.ResponseTimeThreshold.Milliseconds()), metric.Endpoint),
				ServiceName: metric.ServiceName,
				Endpoint:    metric.Endpoint,
				Resolved:    false,
			})
		}

	case "memory":
		if metric.MemoryUsageMB > cpm.config.MemoryUsageThreshold {
			alerts = append(alerts, &ComprehensivePerformanceAlert{
				ID:          fmt.Sprintf("memory_%d", metric.Timestamp.Unix()),
				Timestamp:   time.Now(),
				AlertType:   "high_memory",
				Severity:    cpm.determineSeverity(metric.MemoryUsageMB, cpm.config.MemoryUsageThreshold),
				MetricType:  "memory",
				Threshold:   cpm.config.MemoryUsageThreshold,
				ActualValue: metric.MemoryUsageMB,
				Message: fmt.Sprintf("Memory usage %fMB exceeds threshold %fMB",
					metric.MemoryUsageMB, cpm.config.MemoryUsageThreshold),
				ServiceName: metric.ServiceName,
				Resolved:    false,
			})
		}

	case "database":
		if metric.DatabaseQueryTimeMs > float64(cpm.config.DatabaseQueryThreshold.Milliseconds()) {
			alerts = append(alerts, &ComprehensivePerformanceAlert{
				ID:          fmt.Sprintf("database_%d", metric.Timestamp.Unix()),
				Timestamp:   time.Now(),
				AlertType:   "slow_database_query",
				Severity:    cpm.determineSeverity(metric.DatabaseQueryTimeMs, float64(cpm.config.DatabaseQueryThreshold.Milliseconds())),
				MetricType:  "database",
				Threshold:   float64(cpm.config.DatabaseQueryThreshold.Milliseconds()),
				ActualValue: metric.DatabaseQueryTimeMs,
				Message: fmt.Sprintf("Database query time %fms exceeds threshold %fms",
					metric.DatabaseQueryTimeMs, float64(cpm.config.DatabaseQueryThreshold.Milliseconds())),
				ServiceName: metric.ServiceName,
				Resolved:    false,
			})
		}

	case "security":
		if metric.SecurityValidationTimeMs > float64(cpm.config.SecurityValidationThreshold.Milliseconds()) {
			alerts = append(alerts, &ComprehensivePerformanceAlert{
				ID:          fmt.Sprintf("security_%d", metric.Timestamp.Unix()),
				Timestamp:   time.Now(),
				AlertType:   "slow_security_validation",
				Severity:    cpm.determineSeverity(metric.SecurityValidationTimeMs, float64(cpm.config.SecurityValidationThreshold.Milliseconds())),
				MetricType:  "security",
				Threshold:   float64(cpm.config.SecurityValidationThreshold.Milliseconds()),
				ActualValue: metric.SecurityValidationTimeMs,
				Message: fmt.Sprintf("Security validation time %fms exceeds threshold %fms",
					metric.SecurityValidationTimeMs, float64(cpm.config.SecurityValidationThreshold.Milliseconds())),
				ServiceName: metric.ServiceName,
				Resolved:    false,
			})
		}
	}

	// Store alerts
	for _, alert := range alerts {
		cpm.alerts[alert.ID] = alert
		cpm.logger.Warn("Performance alert triggered",
			zap.String("alert_id", alert.ID),
			zap.String("alert_type", alert.AlertType),
			zap.String("severity", alert.Severity),
			zap.String("message", alert.Message))
	}
}

// determineSeverity determines alert severity based on threshold ratio
func (cpm *ComprehensivePerformanceMonitor) determineSeverity(actual, threshold float64) string {
	ratio := actual / threshold
	if ratio >= 3.0 {
		return "critical"
	} else if ratio >= 2.0 {
		return "high"
	} else if ratio >= 1.5 {
		return "medium"
	}
	return "low"
}

// storeMetricInDatabase stores a metric in the database
func (cpm *ComprehensivePerformanceMonitor) storeMetricInDatabase(metric *ComprehensivePerformanceMetric) error {
	query := `
		INSERT INTO performance_metrics (
			metric_name, metric_value, metric_unit, metric_category,
			threshold_warning, threshold_critical, status, recorded_at, details
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	// Convert metric to database format
	metricName := fmt.Sprintf("%s_%s", metric.MetricType, metric.ServiceName)
	metricValue := cpm.getMetricValue(metric)
	metricUnit := cpm.getMetricUnit(metric)
	metricCategory := metric.MetricType
	status := "OK"
	if metric.ErrorOccurred {
		status = "CRITICAL"
	}

	details, _ := json.Marshal(metric.Metadata)

	_, err := cpm.db.ExecContext(context.Background(), query,
		metricName, metricValue, metricUnit, metricCategory,
		cpm.getWarningThreshold(metric), cpm.getCriticalThreshold(metric),
		status, metric.Timestamp, details)

	return err
}

// getMetricValue extracts the primary metric value
func (cpm *ComprehensivePerformanceMonitor) getMetricValue(metric *ComprehensivePerformanceMetric) float64 {
	switch metric.MetricType {
	case "response_time":
		return metric.ResponseTimeMs
	case "memory":
		return metric.MemoryUsageMB
	case "database":
		return metric.DatabaseQueryTimeMs
	case "security":
		return metric.SecurityValidationTimeMs
	default:
		return 0.0
	}
}

// getMetricUnit returns the unit for the metric
func (cpm *ComprehensivePerformanceMonitor) getMetricUnit(metric *ComprehensivePerformanceMetric) string {
	switch metric.MetricType {
	case "response_time", "database", "security":
		return "ms"
	case "memory":
		return "MB"
	default:
		return "count"
	}
}

// getWarningThreshold returns the warning threshold for the metric
func (cpm *ComprehensivePerformanceMonitor) getWarningThreshold(metric *ComprehensivePerformanceMetric) float64 {
	switch metric.MetricType {
	case "response_time":
		return float64(cpm.config.ResponseTimeThreshold.Milliseconds())
	case "memory":
		return cpm.config.MemoryUsageThreshold
	case "database":
		return float64(cpm.config.DatabaseQueryThreshold.Milliseconds())
	case "security":
		return float64(cpm.config.SecurityValidationThreshold.Milliseconds())
	default:
		return 0.0
	}
}

// getCriticalThreshold returns the critical threshold for the metric
func (cpm *ComprehensivePerformanceMonitor) getCriticalThreshold(metric *ComprehensivePerformanceMetric) float64 {
	return cpm.getWarningThreshold(metric) * 2.0
}

// startWorker starts the background worker for async processing
func (cpm *ComprehensivePerformanceMonitor) startWorker() {
	cpm.workerWg.Add(1)
	go func() {
		defer cpm.workerWg.Done()
		for {
			select {
			case metric := <-cpm.metricsCh:
				if err := cpm.processMetric(metric); err != nil {
					cpm.logger.Error("failed to process metric", zap.Error(err))
				}
			case <-cpm.stopCh:
				return
			}
		}
	}()
}

// startPeriodicCollection starts periodic collection of system metrics
func (cpm *ComprehensivePerformanceMonitor) startPeriodicCollection() {
	ticker := time.NewTicker(cpm.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cpm.collectSystemMetrics()
		case <-cpm.stopCh:
			return
		}
	}
}

// collectSystemMetrics collects system-wide performance metrics
func (cpm *ComprehensivePerformanceMonitor) collectSystemMetrics() {
	// Collect memory metrics
	memoryStats := cpm.memoryMonitor.GetMemoryStats()
	if memoryStats != nil {
		metric := &ComprehensivePerformanceMetric{
			ID:                fmt.Sprintf("memory_%d", time.Now().Unix()),
			Timestamp:         time.Now(),
			MetricType:        "memory",
			ServiceName:       "system",
			MemoryUsageMB:     memoryStats.AllocatedMB,
			MemoryAllocatedMB: memoryStats.TotalAllocatedMB,
			GCPauseTimeMs:     memoryStats.GCPauseTimeMs,
			GoroutineCount:    memoryStats.GoroutineCount,
		}
		cpm.RecordPerformanceMetric(context.Background(), metric)
	}

	// Collect database metrics
	dbMetrics, err := cpm.databaseMonitor.CollectMetrics(context.Background())
	if err == nil && dbMetrics != nil {
		metric := &ComprehensivePerformanceMetric{
			ID:                      fmt.Sprintf("database_%d", time.Now().Unix()),
			Timestamp:               time.Now(),
			MetricType:              "database",
			ServiceName:             "database",
			DatabaseQueryTimeMs:     dbMetrics.AvgQueryTime,
			DatabaseQueryCount:      int(dbMetrics.QueryCount),
			DatabaseConnectionCount: dbMetrics.ConnectionCount,
			DatabaseCacheHitRatio:   dbMetrics.CacheHitRatio,
		}
		cpm.RecordPerformanceMetric(context.Background(), metric)
	}

	// Collect security validation metrics
	securityStats := cpm.securityMonitor.GetSecurityValidationStats()
	if securityStats != nil {
		metric := &ComprehensivePerformanceMetric{
			ID:                        fmt.Sprintf("security_%d", time.Now().Unix()),
			Timestamp:                 time.Now(),
			MetricType:                "security",
			ServiceName:               "security_validation",
			SecurityValidationTimeMs:  securityStats.AverageValidationTimeMs,
			TrustedDataSourceCount:    int(securityStats.TrustedDataSourceValidations),
			WebsiteVerificationTimeMs: securityStats.AverageWebsiteVerificationMs,
		}
		cpm.RecordPerformanceMetric(context.Background(), metric)
	}
}

// GetPerformanceMetrics returns performance metrics for a given time range
func (cpm *ComprehensivePerformanceMonitor) GetPerformanceMetrics(
	ctx context.Context,
	startTime, endTime time.Time,
	metricType string,
) ([]*ComprehensivePerformanceMetric, error) {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	var metrics []*ComprehensivePerformanceMetric
	for _, metric := range cpm.metrics {
		if metric.Timestamp.After(startTime) && metric.Timestamp.Before(endTime) {
			if metricType == "" || metric.MetricType == metricType {
				metrics = append(metrics, metric)
			}
		}
	}

	return metrics, nil
}

// GetPerformanceAlerts returns performance alerts
func (cpm *ComprehensivePerformanceMonitor) GetPerformanceAlerts(
	ctx context.Context,
	resolved bool,
) ([]*ComprehensivePerformanceAlert, error) {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	var alerts []*ComprehensivePerformanceAlert
	for _, alert := range cpm.alerts {
		if alert.Resolved == resolved {
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// GetPerformanceSummary returns a summary of performance metrics
func (cpm *ComprehensivePerformanceMonitor) GetPerformanceSummary(ctx context.Context) (map[string]interface{}, error) {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	// Get latest metrics
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)

	metrics, err := cpm.GetPerformanceMetrics(ctx, oneHourAgo, now, "")
	if err != nil {
		return nil, err
	}

	// Calculate summary statistics
	summary := map[string]interface{}{
		"timestamp": now,
		"period": map[string]interface{}{
			"start": oneHourAgo,
			"end":   now,
		},
		"metrics": map[string]interface{}{
			"total_metrics": len(metrics),
		},
		"alerts": map[string]interface{}{
			"total_alerts":    len(cpm.alerts),
			"active_alerts":   cpm.getActiveAlertCount(),
			"critical_alerts": cpm.getCriticalAlertCount(),
		},
	}

	// Calculate metrics by type
	metricsByType := make(map[string][]*ComprehensivePerformanceMetric)
	for _, metric := range metrics {
		metricsByType[metric.MetricType] = append(metricsByType[metric.MetricType], metric)
	}

	// Calculate averages for each metric type
	for metricType, typeMetrics := range metricsByType {
		if len(typeMetrics) > 0 {
			summary["metrics"].(map[string]interface{})[metricType] = cpm.calculateTypeSummary(metricType, typeMetrics)
		}
	}

	return summary, nil
}

// calculateTypeSummary calculates summary statistics for a metric type
func (cpm *ComprehensivePerformanceMonitor) calculateTypeSummary(metricType string, metrics []*ComprehensivePerformanceMetric) map[string]interface{} {
	summary := map[string]interface{}{
		"count": len(metrics),
	}

	if len(metrics) == 0 {
		return summary
	}

	switch metricType {
	case "response_time":
		var total, min, max float64
		min = metrics[0].ResponseTimeMs
		for _, metric := range metrics {
			total += metric.ResponseTimeMs
			if metric.ResponseTimeMs < min {
				min = metric.ResponseTimeMs
			}
			if metric.ResponseTimeMs > max {
				max = metric.ResponseTimeMs
			}
		}
		summary["average_ms"] = total / float64(len(metrics))
		summary["min_ms"] = min
		summary["max_ms"] = max

	case "memory":
		var total, min, max float64
		min = metrics[0].MemoryUsageMB
		for _, metric := range metrics {
			total += metric.MemoryUsageMB
			if metric.MemoryUsageMB < min {
				min = metric.MemoryUsageMB
			}
			if metric.MemoryUsageMB > max {
				max = metric.MemoryUsageMB
			}
		}
		summary["average_mb"] = total / float64(len(metrics))
		summary["min_mb"] = min
		summary["max_mb"] = max

	case "database":
		var total, min, max float64
		min = metrics[0].DatabaseQueryTimeMs
		for _, metric := range metrics {
			total += metric.DatabaseQueryTimeMs
			if metric.DatabaseQueryTimeMs < min {
				min = metric.DatabaseQueryTimeMs
			}
			if metric.DatabaseQueryTimeMs > max {
				max = metric.DatabaseQueryTimeMs
			}
		}
		summary["average_ms"] = total / float64(len(metrics))
		summary["min_ms"] = min
		summary["max_ms"] = max

	case "security":
		var total, min, max float64
		min = metrics[0].SecurityValidationTimeMs
		for _, metric := range metrics {
			total += metric.SecurityValidationTimeMs
			if metric.SecurityValidationTimeMs < min {
				min = metric.SecurityValidationTimeMs
			}
			if metric.SecurityValidationTimeMs > max {
				max = metric.SecurityValidationTimeMs
			}
		}
		summary["average_ms"] = total / float64(len(metrics))
		summary["min_ms"] = min
		summary["max_ms"] = max
	}

	return summary
}

// getActiveAlertCount returns the count of active alerts
func (cpm *ComprehensivePerformanceMonitor) getActiveAlertCount() int {
	count := 0
	for _, alert := range cpm.alerts {
		if !alert.Resolved {
			count++
		}
	}
	return count
}

// getCriticalAlertCount returns the count of critical alerts
func (cpm *ComprehensivePerformanceMonitor) getCriticalAlertCount() int {
	count := 0
	for _, alert := range cpm.alerts {
		if !alert.Resolved && alert.Severity == "critical" {
			count++
		}
	}
	return count
}

// Stop stops the performance monitor
func (cpm *ComprehensivePerformanceMonitor) Stop() {
	close(cpm.stopCh)
	cpm.workerWg.Wait()
}

// NewResponseTimeTracker creates a new response time tracker
func NewResponseTimeTracker(config *ResponseTimeConfig, logger *zap.Logger) *ResponseTimeTracker {
	return &ResponseTimeTracker{
		config: config,
		logger: logger,
		stats:  make(map[string]*ResponseTimeStats),
	}
}

// NewMemoryMonitor creates a new memory monitor
func NewMemoryMonitor(logger *zap.Logger) *MemoryMonitor {
	return &MemoryMonitor{
		logger: logger,
		stats:  &MemoryStats{},
	}
}

// GetMemoryStats returns current memory statistics
func (mm *MemoryMonitor) GetMemoryStats() *MemoryStats {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	mm.stats = &MemoryStats{
		Timestamp:        time.Now(),
		AllocatedMB:      float64(m.Alloc) / (1024 * 1024),
		TotalAllocatedMB: float64(m.TotalAlloc) / (1024 * 1024),
		SystemMB:         float64(m.Sys) / (1024 * 1024),
		NumGC:            m.NumGC,
		GCPauseTimeMs:    float64(m.PauseNs[(m.NumGC+255)%256]) / 1000000, // Convert to milliseconds
		HeapObjects:      m.HeapObjects,
		StackInUseMB:     float64(m.StackInuse) / (1024 * 1024),
		GoroutineCount:   runtime.NumGoroutine(),
		LastGC:           time.Unix(0, int64(m.LastGC)),
	}

	return mm.stats
}

// NewDatabaseMonitor creates a new database monitor (reusing existing implementation)
func NewDatabaseMonitor(db *sql.DB, config *DatabaseConfig) *DatabaseMonitor {
	return &DatabaseMonitor{
		db:                 db,
		config:             config,
		metrics:            make([]*DatabaseMetrics, 0),
		maxMetrics:         config.MaxMetricsStored,
		slowQueryThreshold: config.SlowQueryThreshold,
		queryStats:         make(map[string]*QueryStats),
	}
}

// CollectMetrics collects database metrics (reusing existing implementation)
func (dm *DatabaseMonitor) CollectMetrics(ctx context.Context) (*DatabaseMetrics, error) {
	metrics := &DatabaseMetrics{
		Timestamp:  time.Now(),
		TableSizes: make(map[string]int64),
		IndexSizes: make(map[string]int64),
		Metadata:   make(map[string]interface{}),
	}

	// Get connection pool stats
	stats := dm.db.Stats()
	metrics.ConnectionCount = stats.OpenConnections
	metrics.MaxConnections = stats.MaxOpenConnections
	metrics.ActiveConnections = stats.InUse
	metrics.IdleConnections = stats.Idle

	// Get database size
	if size, err := dm.getDatabaseSize(ctx); err == nil {
		metrics.DatabaseSize = size
	}

	// Get cache hit ratio
	if hitRatio, err := dm.getCacheHitRatio(ctx); err == nil {
		metrics.CacheHitRatio = hitRatio
	}

	// Store metrics
	dm.storeMetrics(metrics)

	return metrics, nil
}

// getDatabaseSize gets the database size
func (dm *DatabaseMonitor) getDatabaseSize(ctx context.Context) (int64, error) {
	var size int64
	query := "SELECT pg_database_size(current_database())"
	err := dm.db.QueryRowContext(ctx, query).Scan(&size)
	return size, err
}

// getCacheHitRatio gets the cache hit ratio
func (dm *DatabaseMonitor) getCacheHitRatio(ctx context.Context) (float64, error) {
	var hitRatio float64
	query := `
		SELECT 
			round(100.0 * sum(blks_hit) / (sum(blks_hit) + sum(blks_read)), 2) as cache_hit_ratio
		FROM pg_stat_database 
		WHERE datname = current_database()
	`
	err := dm.db.QueryRowContext(ctx, query).Scan(&hitRatio)
	return hitRatio, err
}

// storeMetrics stores metrics in memory
func (dm *DatabaseMonitor) storeMetrics(metrics *DatabaseMetrics) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.metrics = append(dm.metrics, metrics)

	// Keep only the most recent metrics
	if len(dm.metrics) > dm.maxMetrics {
		dm.metrics = dm.metrics[len(dm.metrics)-dm.maxMetrics:]
	}
}

// NewQueryPerformanceMonitor creates a new query performance monitor
func NewQueryPerformanceMonitor(db *sql.DB, logger *zap.Logger) *QueryPerformanceMonitor {
	return &QueryPerformanceMonitor{
		db:     db,
		logger: logger,
		stats:  make(map[string]*QueryPerformanceStats),
	}
}

// NewSecurityValidationMonitor creates a new security validation monitor
func NewSecurityValidationMonitor(logger *zap.Logger) *SecurityValidationMonitor {
	return &SecurityValidationMonitor{
		logger: logger,
		stats:  &SecurityValidationStats{},
	}
}

// GetSecurityValidationStats returns current security validation statistics
func (svm *SecurityValidationMonitor) GetSecurityValidationStats() *SecurityValidationStats {
	svm.mu.Lock()
	defer svm.mu.Unlock()

	// This would be populated by actual security validation calls
	// For now, return current stats
	return svm.stats
}
