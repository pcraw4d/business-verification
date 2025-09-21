// Package observability provides unified monitoring capabilities for the KYB platform.
// This package consolidates all monitoring functionality into a single, efficient system.
package observability

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// UnifiedPerformanceMonitor provides centralized monitoring for all system components.
// It consolidates the functionality of multiple monitoring systems into a single,
// efficient, and maintainable solution.
type UnifiedPerformanceMonitor struct {
	config     *MonitoringConfig
	db         *sql.DB
	exporters  []MetricExporter
	alerters   []AlertHandler
	collectors map[string]MetricCollector
	logger     *zap.Logger
	metrics    *prometheus.Registry
	mu         sync.RWMutex
}

// MonitoringConfig contains configuration for the unified monitoring system.
type MonitoringConfig struct {
	DatabaseURL        string
	CollectionInterval time.Duration
	RetentionPeriod    time.Duration
	AlertThresholds    map[string]AlertThreshold
	Exporters          []ExporterConfig
	Components         []ComponentConfig
	BatchSize          int
	FlushInterval      time.Duration
}

// AlertThreshold defines alerting thresholds for specific metrics.
type AlertThreshold struct {
	MetricName string
	Component  string
	Condition  string // 'gt', 'lt', 'eq', 'ne'
	Value      float64
	Severity   string
	Duration   time.Duration
	Cooldown   time.Duration
}

// ExporterConfig contains configuration for metric exporters.
type ExporterConfig struct {
	Type    string                 // 'prometheus', 'log', 'webhook'
	Config  map[string]interface{} // Exporter-specific configuration
	Enabled bool
}

// ComponentConfig contains configuration for monitored components.
type ComponentConfig struct {
	Name        string
	ServiceName string
	Collectors  []string
	Enabled     bool
}

// MetricCollector interface for collecting metrics from different components.
type MetricCollector interface {
	Collect(ctx context.Context) ([]Metric, error)
	GetType() string
	GetComponent() string
	GetInterval() time.Duration
	Start(ctx context.Context) error
	Stop() error
}

// Metric represents a single performance metric.
type Metric struct {
	ID                string                 `json:"id"`
	Timestamp         time.Time              `json:"timestamp"`
	Component         string                 `json:"component"`
	ComponentInstance string                 `json:"component_instance"`
	ServiceName       string                 `json:"service_name"`
	MetricType        string                 `json:"metric_type"`
	MetricCategory    string                 `json:"metric_category"`
	MetricName        string                 `json:"metric_name"`
	MetricValue       float64                `json:"metric_value"`
	MetricUnit        string                 `json:"metric_unit"`
	Tags              map[string]string      `json:"tags"`
	Metadata          map[string]interface{} `json:"metadata"`
	RequestID         *string                `json:"request_id"`
	OperationID       *string                `json:"operation_id"`
	UserID            *string                `json:"user_id"`
	ConfidenceScore   *float64               `json:"confidence_score"`
	DataSource        string                 `json:"data_source"`
}

// Alert represents a performance alert.
type Alert struct {
	ID                string                 `json:"id"`
	CreatedAt         time.Time              `json:"created_at"`
	AlertType         string                 `json:"alert_type"`
	AlertCategory     string                 `json:"alert_category"`
	Severity          string                 `json:"severity"`
	Component         string                 `json:"component"`
	ComponentInstance string                 `json:"component_instance"`
	ServiceName       string                 `json:"service_name"`
	AlertName         string                 `json:"alert_name"`
	Description       string                 `json:"description"`
	Condition         map[string]interface{} `json:"condition"`
	CurrentValue      *float64               `json:"current_value"`
	ThresholdValue    *float64               `json:"threshold_value"`
	Status            string                 `json:"status"`
	AcknowledgedBy    *string                `json:"acknowledged_by"`
	AcknowledgedAt    *time.Time             `json:"acknowledged_at"`
	ResolvedAt        *time.Time             `json:"resolved_at"`
	RelatedMetrics    []string               `json:"related_metrics"`
	RelatedRequests   []string               `json:"related_requests"`
	Tags              map[string]string      `json:"tags"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// HealthScore represents component health scores.
type HealthScore struct {
	ID                 string                 `json:"id"`
	Timestamp          time.Time              `json:"timestamp"`
	Component          string                 `json:"component"`
	ComponentInstance  string                 `json:"component_instance"`
	ServiceName        string                 `json:"service_name"`
	OverallHealth      float64                `json:"overall_health"`
	PerformanceHealth  float64                `json:"performance_health"`
	ResourceHealth     float64                `json:"resource_health"`
	AvailabilityHealth float64                `json:"availability_health"`
	SecurityHealth     float64                `json:"security_health"`
	ActiveAlerts       int                    `json:"active_alerts"`
	CriticalAlerts     int                    `json:"critical_alerts"`
	WarningAlerts      int                    `json:"warning_alerts"`
	AvgResponseTime    *float64               `json:"avg_response_time"`
	ErrorRate          *float64               `json:"error_rate"`
	Throughput         *float64               `json:"throughput"`
	CPUUsage           *float64               `json:"cpu_usage"`
	MemoryUsage        *float64               `json:"memory_usage"`
	DiskUsage          *float64               `json:"disk_usage"`
	Tags               map[string]string      `json:"tags"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// MetricExporter interface for exporting metrics to external systems.
type MetricExporter interface {
	Export(ctx context.Context, metrics []Metric) error
	GetType() string
	IsEnabled() bool
}

// AlertHandler interface for handling alerts.
type AlertHandler interface {
	HandleAlert(ctx context.Context, alert Alert) error
	GetType() string
	IsEnabled() bool
}

// NewUnifiedPerformanceMonitor creates a new unified performance monitor.
func NewUnifiedPerformanceMonitor(config *MonitoringConfig, db *sql.DB, logger *zap.Logger) *UnifiedPerformanceMonitor {
	return &UnifiedPerformanceMonitor{
		config:     config,
		db:         db,
		exporters:  make([]MetricExporter, 0),
		alerters:   make([]AlertHandler, 0),
		collectors: make(map[string]MetricCollector),
		logger:     logger,
		metrics:    prometheus.NewRegistry(),
	}
}

// RegisterCollector registers a metric collector.
func (m *UnifiedPerformanceMonitor) RegisterCollector(name string, collector MetricCollector) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.collectors[name] = collector
}

// RegisterExporter registers a metric exporter.
func (m *UnifiedPerformanceMonitor) RegisterExporter(exporter MetricExporter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.exporters = append(m.exporters, exporter)
}

// RegisterAlerter registers an alert handler.
func (m *UnifiedPerformanceMonitor) RegisterAlerter(alerter AlertHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alerters = append(m.alerters, alerter)
}

// Start starts the unified performance monitor.
func (m *UnifiedPerformanceMonitor) Start(ctx context.Context) error {
	m.logger.Info("Starting unified performance monitor")

	// Start all collectors
	for name, collector := range m.collectors {
		if err := collector.Start(ctx); err != nil {
			m.logger.Error("Failed to start collector", zap.String("collector", name), zap.Error(err))
			return fmt.Errorf("failed to start collector %s: %w", name, err)
		}
		m.logger.Info("Started collector", zap.String("collector", name))
	}

	// Start metric collection loop
	go m.collectionLoop(ctx)

	// Start alert processing loop
	go m.alertProcessingLoop(ctx)

	// Start health score calculation loop
	go m.healthScoreLoop(ctx)

	m.logger.Info("Unified performance monitor started successfully")
	return nil
}

// Stop stops the unified performance monitor.
func (m *UnifiedPerformanceMonitor) Stop() error {
	m.logger.Info("Stopping unified performance monitor")

	// Stop all collectors
	for name, collector := range m.collectors {
		if err := collector.Stop(); err != nil {
			m.logger.Error("Failed to stop collector", zap.String("collector", name), zap.Error(err))
		}
		m.logger.Info("Stopped collector", zap.String("collector", name))
	}

	m.logger.Info("Unified performance monitor stopped")
	return nil
}

// RecordMetric records a single performance metric.
func (m *UnifiedPerformanceMonitor) RecordMetric(ctx context.Context, metric Metric) error {
	// Generate ID if not provided
	if metric.ID == "" {
		metric.ID = uuid.New().String()
	}

	// Set timestamp if not provided
	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}

	// Set data source if not provided
	if metric.DataSource == "" {
		metric.DataSource = "application"
	}

	// Insert into database
	if err := m.insertMetric(ctx, metric); err != nil {
		m.logger.Error("Failed to insert metric", zap.Error(err))
		return fmt.Errorf("failed to insert metric: %w", err)
	}

	// Export to external systems
	if err := m.exportMetrics(ctx, []Metric{metric}); err != nil {
		m.logger.Error("Failed to export metrics", zap.Error(err))
		// Don't return error for export failures
	}

	// Check for alerts
	if err := m.checkAlerts(ctx, metric); err != nil {
		m.logger.Error("Failed to check alerts", zap.Error(err))
		// Don't return error for alert check failures
	}

	return nil
}

// RecordMetrics records multiple performance metrics in batch.
func (m *UnifiedPerformanceMonitor) RecordMetrics(ctx context.Context, metrics []Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	// Generate IDs and timestamps for metrics that don't have them
	for i := range metrics {
		if metrics[i].ID == "" {
			metrics[i].ID = uuid.New().String()
		}
		if metrics[i].Timestamp.IsZero() {
			metrics[i].Timestamp = time.Now()
		}
		if metrics[i].DataSource == "" {
			metrics[i].DataSource = "application"
		}
	}

	// Insert into database in batch
	if err := m.insertMetricsBatch(ctx, metrics); err != nil {
		m.logger.Error("Failed to insert metrics batch", zap.Error(err))
		return fmt.Errorf("failed to insert metrics batch: %w", err)
	}

	// Export to external systems
	if err := m.exportMetrics(ctx, metrics); err != nil {
		m.logger.Error("Failed to export metrics", zap.Error(err))
		// Don't return error for export failures
	}

	// Check for alerts
	for _, metric := range metrics {
		if err := m.checkAlerts(ctx, metric); err != nil {
			m.logger.Error("Failed to check alerts", zap.Error(err))
			// Don't return error for alert check failures
		}
	}

	return nil
}

// GetMetrics retrieves metrics based on filters.
func (m *UnifiedPerformanceMonitor) GetMetrics(ctx context.Context, filters MetricFilters) ([]Metric, error) {
	query := `
		SELECT id, timestamp, component, component_instance, service_name,
		       metric_type, metric_category, metric_name, metric_value, metric_unit,
		       tags, metadata, request_id, operation_id, user_id,
		       confidence_score, data_source, created_at
		FROM unified_performance_metrics
		WHERE 1=1
	`

	args := make([]interface{}, 0)
	argIndex := 1

	if filters.Component != "" {
		query += fmt.Sprintf(" AND component = $%d", argIndex)
		args = append(args, filters.Component)
		argIndex++
	}

	if filters.ServiceName != "" {
		query += fmt.Sprintf(" AND service_name = $%d", argIndex)
		args = append(args, filters.ServiceName)
		argIndex++
	}

	if filters.MetricType != "" {
		query += fmt.Sprintf(" AND metric_type = $%d", argIndex)
		args = append(args, filters.MetricType)
		argIndex++
	}

	if filters.MetricCategory != "" {
		query += fmt.Sprintf(" AND metric_category = $%d", argIndex)
		args = append(args, filters.MetricCategory)
		argIndex++
	}

	if !filters.StartTime.IsZero() {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, filters.StartTime)
		argIndex++
	}

	if !filters.EndTime.IsZero() {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, filters.EndTime)
		argIndex++
	}

	if filters.Limit > 0 {
		query += fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
	} else {
		query += " ORDER BY timestamp DESC"
	}

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	var metrics []Metric
	for rows.Next() {
		var metric Metric
		var tagsJSON, metadataJSON sql.NullString
		var requestID, operationID, userID sql.NullString
		var confidenceScore sql.NullFloat64

		err := rows.Scan(
			&metric.ID, &metric.Timestamp, &metric.Component, &metric.ComponentInstance,
			&metric.ServiceName, &metric.MetricType, &metric.MetricCategory,
			&metric.MetricName, &metric.MetricValue, &metric.MetricUnit,
			&tagsJSON, &metadataJSON, &requestID, &operationID, &userID,
			&confidenceScore, &metric.DataSource, &metric.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}

		// Parse JSON fields
		if tagsJSON.Valid {
			json.Unmarshal([]byte(tagsJSON.String), &metric.Tags)
		}
		if metadataJSON.Valid {
			json.Unmarshal([]byte(metadataJSON.String), &metric.Metadata)
		}
		if requestID.Valid {
			metric.RequestID = &requestID.String
		}
		if operationID.Valid {
			metric.OperationID = &operationID.String
		}
		if userID.Valid {
			metric.UserID = &userID.String
		}
		if confidenceScore.Valid {
			metric.ConfidenceScore = &confidenceScore.Float64
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetHealthScores retrieves health scores for components.
func (m *UnifiedPerformanceMonitor) GetHealthScores(ctx context.Context, component string) ([]HealthScore, error) {
	query := `
		SELECT id, timestamp, component, component_instance, service_name,
		       overall_health, performance_health, resource_health,
		       availability_health, security_health, active_alerts,
		       critical_alerts, warning_alerts, avg_response_time,
		       error_rate, throughput, cpu_usage, memory_usage,
		       disk_usage, tags, metadata
		FROM performance_health_scores
		WHERE component = $1
		ORDER BY timestamp DESC
		LIMIT 100
	`

	rows, err := m.db.QueryContext(ctx, query, component)
	if err != nil {
		return nil, fmt.Errorf("failed to query health scores: %w", err)
	}
	defer rows.Close()

	var healthScores []HealthScore
	for rows.Next() {
		var healthScore HealthScore
		var tagsJSON, metadataJSON sql.NullString
		var avgResponseTime, errorRate, throughput, cpuUsage, memoryUsage, diskUsage sql.NullFloat64

		err := rows.Scan(
			&healthScore.ID, &healthScore.Timestamp, &healthScore.Component,
			&healthScore.ComponentInstance, &healthScore.ServiceName,
			&healthScore.OverallHealth, &healthScore.PerformanceHealth,
			&healthScore.ResourceHealth, &healthScore.AvailabilityHealth,
			&healthScore.SecurityHealth, &healthScore.ActiveAlerts,
			&healthScore.CriticalAlerts, &healthScore.WarningAlerts,
			&avgResponseTime, &errorRate, &throughput, &cpuUsage,
			&memoryUsage, &diskUsage, &tagsJSON, &metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan health score: %w", err)
		}

		// Parse JSON fields
		if tagsJSON.Valid {
			json.Unmarshal([]byte(tagsJSON.String), &healthScore.Tags)
		}
		if metadataJSON.Valid {
			json.Unmarshal([]byte(metadataJSON.String), &healthScore.Metadata)
		}
		if avgResponseTime.Valid {
			healthScore.AvgResponseTime = &avgResponseTime.Float64
		}
		if errorRate.Valid {
			healthScore.ErrorRate = &errorRate.Float64
		}
		if throughput.Valid {
			healthScore.Throughput = &throughput.Float64
		}
		if cpuUsage.Valid {
			healthScore.CPUUsage = &cpuUsage.Float64
		}
		if memoryUsage.Valid {
			healthScore.MemoryUsage = &memoryUsage.Float64
		}
		if diskUsage.Valid {
			healthScore.DiskUsage = &diskUsage.Float64
		}

		healthScores = append(healthScores, healthScore)
	}

	return healthScores, nil
}

// MetricFilters defines filters for querying metrics.
type MetricFilters struct {
	Component      string
	ServiceName    string
	MetricType     string
	MetricCategory string
	StartTime      time.Time
	EndTime        time.Time
	Limit          int
}

// collectionLoop runs the metric collection loop.
func (m *UnifiedPerformanceMonitor) collectionLoop(ctx context.Context) {
	ticker := time.NewTicker(m.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.collectMetrics(ctx)
		}
	}
}

// collectMetrics collects metrics from all registered collectors.
func (m *UnifiedPerformanceMonitor) collectMetrics(ctx context.Context) {
	m.mu.RLock()
	collectors := make(map[string]MetricCollector)
	for name, collector := range m.collectors {
		collectors[name] = collector
	}
	m.mu.RUnlock()

	for name, collector := range collectors {
		metrics, err := collector.Collect(ctx)
		if err != nil {
			m.logger.Error("Failed to collect metrics", zap.String("collector", name), zap.Error(err))
			continue
		}

		if len(metrics) > 0 {
			if err := m.RecordMetrics(ctx, metrics); err != nil {
				m.logger.Error("Failed to record metrics", zap.String("collector", name), zap.Error(err))
			}
		}
	}
}

// alertProcessingLoop processes alerts.
func (m *UnifiedPerformanceMonitor) alertProcessingLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Check alerts every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.processAlerts(ctx)
		}
	}
}

// processAlerts processes active alerts.
func (m *UnifiedPerformanceMonitor) processAlerts(ctx context.Context) {
	// Get active alerts
	alerts, err := m.getActiveAlerts(ctx)
	if err != nil {
		m.logger.Error("Failed to get active alerts", zap.Error(err))
		return
	}

	// Process each alert
	for _, alert := range alerts {
		for _, alerter := range m.alerters {
			if alerter.IsEnabled() {
				if err := alerter.HandleAlert(ctx, alert); err != nil {
					m.logger.Error("Failed to handle alert", zap.String("alerter", alerter.GetType()), zap.Error(err))
				}
			}
		}
	}
}

// healthScoreLoop calculates and updates health scores.
func (m *UnifiedPerformanceMonitor) healthScoreLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute) // Update health scores every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.calculateHealthScores(ctx)
		}
	}
}

// calculateHealthScores calculates health scores for all components.
func (m *UnifiedPerformanceMonitor) calculateHealthScores(ctx context.Context) {
	// Get unique components
	components, err := m.getComponents(ctx)
	if err != nil {
		m.logger.Error("Failed to get components", zap.Error(err))
		return
	}

	for _, component := range components {
		healthScore, err := m.calculateComponentHealthScore(ctx, component)
		if err != nil {
			m.logger.Error("Failed to calculate health score", zap.String("component", component), zap.Error(err))
			continue
		}

		if err := m.updateHealthScore(ctx, healthScore); err != nil {
			m.logger.Error("Failed to update health score", zap.String("component", component), zap.Error(err))
		}
	}
}

// insertMetric inserts a single metric into the database.
func (m *UnifiedPerformanceMonitor) insertMetric(ctx context.Context, metric Metric) error {
	tagsJSON, _ := json.Marshal(metric.Tags)
	metadataJSON, _ := json.Marshal(metric.Metadata)

	query := `
		INSERT INTO unified_performance_metrics (
			id, timestamp, component, component_instance, service_name,
			metric_type, metric_category, metric_name, metric_value, metric_unit,
			tags, metadata, request_id, operation_id, user_id,
			confidence_score, data_source
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err := m.db.ExecContext(ctx, query,
		metric.ID, metric.Timestamp, metric.Component, metric.ComponentInstance,
		metric.ServiceName, metric.MetricType, metric.MetricCategory,
		metric.MetricName, metric.MetricValue, metric.MetricUnit,
		tagsJSON, metadataJSON, metric.RequestID, metric.OperationID,
		metric.UserID, metric.ConfidenceScore, metric.DataSource,
	)

	return err
}

// insertMetricsBatch inserts multiple metrics into the database in batch.
func (m *UnifiedPerformanceMonitor) insertMetricsBatch(ctx context.Context, metrics []Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	// Use batch insert for better performance
	query := `
		INSERT INTO unified_performance_metrics (
			id, timestamp, component, component_instance, service_name,
			metric_type, metric_category, metric_name, metric_value, metric_unit,
			tags, metadata, request_id, operation_id, user_id,
			confidence_score, data_source
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	stmt, err := m.db.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, metric := range metrics {
		tagsJSON, _ := json.Marshal(metric.Tags)
		metadataJSON, _ := json.Marshal(metric.Metadata)

		_, err := stmt.ExecContext(ctx,
			metric.ID, metric.Timestamp, metric.Component, metric.ComponentInstance,
			metric.ServiceName, metric.MetricType, metric.MetricCategory,
			metric.MetricName, metric.MetricValue, metric.MetricUnit,
			tagsJSON, metadataJSON, metric.RequestID, metric.OperationID,
			metric.UserID, metric.ConfidenceScore, metric.DataSource,
		)
		if err != nil {
			return fmt.Errorf("failed to insert metric %s: %w", metric.ID, err)
		}
	}

	return nil
}

// exportMetrics exports metrics to all registered exporters.
func (m *UnifiedPerformanceMonitor) exportMetrics(ctx context.Context, metrics []Metric) error {
	for _, exporter := range m.exporters {
		if exporter.IsEnabled() {
			if err := exporter.Export(ctx, metrics); err != nil {
				m.logger.Error("Failed to export metrics", zap.String("exporter", exporter.GetType()), zap.Error(err))
			}
		}
	}
	return nil
}

// checkAlerts checks if a metric triggers any alerts.
func (m *UnifiedPerformanceMonitor) checkAlerts(ctx context.Context, metric Metric) error {
	// Check against configured thresholds
	for _, threshold := range m.config.AlertThresholds {
		if threshold.Component == metric.Component && threshold.MetricName == metric.MetricName {
			if m.evaluateThreshold(metric, threshold) {
				alert := Alert{
					ID:                uuid.New().String(),
					CreatedAt:         time.Now(),
					AlertType:         "threshold",
					AlertCategory:     "performance",
					Severity:          threshold.Severity,
					Component:         metric.Component,
					ComponentInstance: metric.ComponentInstance,
					ServiceName:       metric.ServiceName,
					AlertName:         fmt.Sprintf("%s threshold exceeded", metric.MetricName),
					Description:       fmt.Sprintf("Metric %s exceeded threshold %f", metric.MetricName, threshold.Value),
					Condition:         map[string]interface{}{"threshold": threshold.Value, "condition": threshold.Condition},
					CurrentValue:      &metric.MetricValue,
					ThresholdValue:    &threshold.Value,
					Status:            "active",
					RelatedMetrics:    []string{metric.ID},
					Tags:              metric.Tags,
					Metadata:          metric.Metadata,
				}

				if err := m.createAlert(ctx, alert); err != nil {
					m.logger.Error("Failed to create alert", zap.Error(err))
				}
			}
		}
	}
	return nil
}

// evaluateThreshold evaluates if a metric meets alert threshold conditions.
func (m *UnifiedPerformanceMonitor) evaluateThreshold(metric Metric, threshold AlertThreshold) bool {
	switch threshold.Condition {
	case "gt":
		return metric.MetricValue > threshold.Value
	case "lt":
		return metric.MetricValue < threshold.Value
	case "eq":
		return metric.MetricValue == threshold.Value
	case "ne":
		return metric.MetricValue != threshold.Value
	default:
		return false
	}
}

// createAlert creates a new alert in the database.
func (m *UnifiedPerformanceMonitor) createAlert(ctx context.Context, alert Alert) error {
	conditionJSON, _ := json.Marshal(alert.Condition)
	tagsJSON, _ := json.Marshal(alert.Tags)
	metadataJSON, _ := json.Marshal(alert.Metadata)
	relatedMetricsJSON, _ := json.Marshal(alert.RelatedMetrics)
	relatedRequestsJSON, _ := json.Marshal(alert.RelatedRequests)

	query := `
		INSERT INTO unified_performance_alerts (
			id, created_at, alert_type, alert_category, severity,
			component, component_instance, service_name, alert_name,
			description, condition, current_value, threshold_value,
			status, related_metrics, related_requests, tags, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`

	_, err := m.db.ExecContext(ctx, query,
		alert.ID, alert.CreatedAt, alert.AlertType, alert.AlertCategory, alert.Severity,
		alert.Component, alert.ComponentInstance, alert.ServiceName, alert.AlertName,
		alert.Description, conditionJSON, alert.CurrentValue, alert.ThresholdValue,
		alert.Status, relatedMetricsJSON, relatedRequestsJSON, tagsJSON, metadataJSON,
	)

	return err
}

// getActiveAlerts retrieves active alerts from the database.
func (m *UnifiedPerformanceMonitor) getActiveAlerts(ctx context.Context) ([]Alert, error) {
	query := `
		SELECT id, created_at, alert_type, alert_category, severity,
		       component, component_instance, service_name, alert_name,
		       description, condition, current_value, threshold_value,
		       status, acknowledged_by, acknowledged_at, resolved_at,
		       related_metrics, related_requests, tags, metadata
		FROM unified_performance_alerts
		WHERE status = 'active'
		ORDER BY created_at DESC
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active alerts: %w", err)
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var alert Alert
		var conditionJSON, tagsJSON, metadataJSON, relatedMetricsJSON, relatedRequestsJSON sql.NullString
		var acknowledgedBy sql.NullString
		var acknowledgedAt, resolvedAt sql.NullTime

		err := rows.Scan(
			&alert.ID, &alert.CreatedAt, &alert.AlertType, &alert.AlertCategory,
			&alert.Severity, &alert.Component, &alert.ComponentInstance,
			&alert.ServiceName, &alert.AlertName, &alert.Description,
			&conditionJSON, &alert.CurrentValue, &alert.ThresholdValue,
			&alert.Status, &acknowledgedBy, &acknowledgedAt, &resolvedAt,
			&relatedMetricsJSON, &relatedRequestsJSON, &tagsJSON, &metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}

		// Parse JSON fields
		if conditionJSON.Valid {
			json.Unmarshal([]byte(conditionJSON.String), &alert.Condition)
		}
		if tagsJSON.Valid {
			json.Unmarshal([]byte(tagsJSON.String), &alert.Tags)
		}
		if metadataJSON.Valid {
			json.Unmarshal([]byte(metadataJSON.String), &alert.Metadata)
		}
		if relatedMetricsJSON.Valid {
			json.Unmarshal([]byte(relatedMetricsJSON.String), &alert.RelatedMetrics)
		}
		if relatedRequestsJSON.Valid {
			json.Unmarshal([]byte(relatedRequestsJSON.String), &alert.RelatedRequests)
		}
		if acknowledgedBy.Valid {
			alert.AcknowledgedBy = &acknowledgedBy.String
		}
		if acknowledgedAt.Valid {
			alert.AcknowledgedAt = &acknowledgedAt.Time
		}
		if resolvedAt.Valid {
			alert.ResolvedAt = &resolvedAt.Time
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// getComponents retrieves unique components from the database.
func (m *UnifiedPerformanceMonitor) getComponents(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT component FROM unified_performance_metrics ORDER BY component`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query components: %w", err)
	}
	defer rows.Close()

	var components []string
	for rows.Next() {
		var component string
		if err := rows.Scan(&component); err != nil {
			return nil, fmt.Errorf("failed to scan component: %w", err)
		}
		components = append(components, component)
	}

	return components, nil
}

// calculateComponentHealthScore calculates health score for a component.
func (m *UnifiedPerformanceMonitor) calculateComponentHealthScore(ctx context.Context, component string) (*HealthScore, error) {
	// Get recent metrics for the component
	metrics, err := m.GetMetrics(ctx, MetricFilters{
		Component: component,
		StartTime: time.Now().Add(-5 * time.Minute),
		EndTime:   time.Now(),
		Limit:     1000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics for component %s: %w", component, err)
	}

	// Calculate health scores based on metrics
	healthScore := &HealthScore{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Component: component,
	}

	// Calculate performance health based on response times and error rates
	performanceHealth := m.calculatePerformanceHealth(metrics)
	healthScore.PerformanceHealth = performanceHealth

	// Calculate resource health based on CPU, memory, and disk usage
	resourceHealth := m.calculateResourceHealth(metrics)
	healthScore.ResourceHealth = resourceHealth

	// Calculate availability health based on uptime and error rates
	availabilityHealth := m.calculateAvailabilityHealth(metrics)
	healthScore.AvailabilityHealth = availabilityHealth

	// Calculate security health based on security-related metrics
	securityHealth := m.calculateSecurityHealth(metrics)
	healthScore.SecurityHealth = securityHealth

	// Calculate overall health as weighted average
	healthScore.OverallHealth = (performanceHealth + resourceHealth + availabilityHealth + securityHealth) / 4.0

	// Get alert counts
	activeAlerts, criticalAlerts, warningAlerts, err := m.getAlertCounts(ctx, component)
	if err != nil {
		m.logger.Error("Failed to get alert counts", zap.String("component", component), zap.Error(err))
	} else {
		healthScore.ActiveAlerts = activeAlerts
		healthScore.CriticalAlerts = criticalAlerts
		healthScore.WarningAlerts = warningAlerts
	}

	// Extract performance indicators
	healthScore.AvgResponseTime = m.extractAvgResponseTime(metrics)
	healthScore.ErrorRate = m.extractErrorRate(metrics)
	healthScore.Throughput = m.extractThroughput(metrics)
	healthScore.CPUUsage = m.extractCPUUsage(metrics)
	healthScore.MemoryUsage = m.extractMemoryUsage(metrics)
	healthScore.DiskUsage = m.extractDiskUsage(metrics)

	return healthScore, nil
}

// calculatePerformanceHealth calculates performance health score.
func (m *UnifiedPerformanceMonitor) calculatePerformanceHealth(metrics []Metric) float64 {
	var responseTimeSum float64
	var responseTimeCount int
	var errorRateSum float64
	var errorRateCount int

	for _, metric := range metrics {
		if metric.MetricCategory == "latency" || metric.MetricName == "response_time" {
			responseTimeSum += metric.MetricValue
			responseTimeCount++
		}
		if metric.MetricCategory == "error_rate" {
			errorRateSum += metric.MetricValue
			errorRateCount++
		}
	}

	// Calculate health based on response times and error rates
	// This is a simplified calculation - in production, you'd want more sophisticated logic
	if responseTimeCount > 0 {
		avgResponseTime := responseTimeSum / float64(responseTimeCount)
		if avgResponseTime > 1000 { // 1 second
			return 0.5
		} else if avgResponseTime > 500 { // 500ms
			return 0.7
		}
	}

	if errorRateCount > 0 {
		avgErrorRate := errorRateSum / float64(errorRateCount)
		if avgErrorRate > 0.05 { // 5%
			return 0.3
		} else if avgErrorRate > 0.01 { // 1%
			return 0.7
		}
	}

	return 1.0 // Good performance
}

// calculateResourceHealth calculates resource health score.
func (m *UnifiedPerformanceMonitor) calculateResourceHealth(metrics []Metric) float64 {
	var cpuSum, memorySum, diskSum float64
	var cpuCount, memoryCount, diskCount int

	for _, metric := range metrics {
		if metric.MetricName == "cpu_usage" {
			cpuSum += metric.MetricValue
			cpuCount++
		}
		if metric.MetricName == "memory_usage" {
			memorySum += metric.MetricValue
			memoryCount++
		}
		if metric.MetricName == "disk_usage" {
			diskSum += metric.MetricValue
			diskCount++
		}
	}

	// Calculate health based on resource usage
	// This is a simplified calculation
	if cpuCount > 0 {
		avgCPU := cpuSum / float64(cpuCount)
		if avgCPU > 90 {
			return 0.2
		} else if avgCPU > 80 {
			return 0.5
		}
	}

	if memoryCount > 0 {
		avgMemory := memorySum / float64(memoryCount)
		if avgMemory > 90 {
			return 0.2
		} else if avgMemory > 80 {
			return 0.5
		}
	}

	return 1.0 // Good resource health
}

// calculateAvailabilityHealth calculates availability health score.
func (m *UnifiedPerformanceMonitor) calculateAvailabilityHealth(metrics []Metric) float64 {
	// Simplified availability calculation
	// In production, you'd want more sophisticated logic based on uptime, error rates, etc.
	return 1.0
}

// calculateSecurityHealth calculates security health score.
func (m *UnifiedPerformanceMonitor) calculateSecurityHealth(metrics []Metric) float64 {
	// Simplified security calculation
	// In production, you'd want more sophisticated logic based on security metrics
	return 1.0
}

// getAlertCounts retrieves alert counts for a component.
func (m *UnifiedPerformanceMonitor) getAlertCounts(ctx context.Context, component string) (int, int, int, error) {
	query := `
		SELECT 
			COUNT(*) as total_alerts,
			COUNT(CASE WHEN severity = 'critical' THEN 1 END) as critical_alerts,
			COUNT(CASE WHEN severity = 'warning' THEN 1 END) as warning_alerts
		FROM unified_performance_alerts
		WHERE component = $1 AND status = 'active'
	`

	var totalAlerts, criticalAlerts, warningAlerts int
	err := m.db.QueryRowContext(ctx, query, component).Scan(&totalAlerts, &criticalAlerts, &warningAlerts)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get alert counts: %w", err)
	}

	return totalAlerts, criticalAlerts, warningAlerts, nil
}

// updateHealthScore updates health score in the database.
func (m *UnifiedPerformanceMonitor) updateHealthScore(ctx context.Context, healthScore *HealthScore) error {
	tagsJSON, _ := json.Marshal(healthScore.Tags)
	metadataJSON, _ := json.Marshal(healthScore.Metadata)

	query := `
		INSERT INTO performance_health_scores (
			id, timestamp, component, component_instance, service_name,
			overall_health, performance_health, resource_health,
			availability_health, security_health, active_alerts,
			critical_alerts, warning_alerts, avg_response_time,
			error_rate, throughput, cpu_usage, memory_usage,
			disk_usage, tags, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
	`

	_, err := m.db.ExecContext(ctx, query,
		healthScore.ID, healthScore.Timestamp, healthScore.Component,
		healthScore.ComponentInstance, healthScore.ServiceName,
		healthScore.OverallHealth, healthScore.PerformanceHealth,
		healthScore.ResourceHealth, healthScore.AvailabilityHealth,
		healthScore.SecurityHealth, healthScore.ActiveAlerts,
		healthScore.CriticalAlerts, healthScore.WarningAlerts,
		healthScore.AvgResponseTime, healthScore.ErrorRate,
		healthScore.Throughput, healthScore.CPUUsage,
		healthScore.MemoryUsage, healthScore.DiskUsage,
		tagsJSON, metadataJSON,
	)

	return err
}

// Helper functions for extracting specific metrics
func (m *UnifiedPerformanceMonitor) extractAvgResponseTime(metrics []Metric) *float64 {
	var sum float64
	var count int
	for _, metric := range metrics {
		if metric.MetricName == "response_time" {
			sum += metric.MetricValue
			count++
		}
	}
	if count > 0 {
		avg := sum / float64(count)
		return &avg
	}
	return nil
}

func (m *UnifiedPerformanceMonitor) extractErrorRate(metrics []Metric) *float64 {
	var sum float64
	var count int
	for _, metric := range metrics {
		if metric.MetricName == "error_rate" {
			sum += metric.MetricValue
			count++
		}
	}
	if count > 0 {
		avg := sum / float64(count)
		return &avg
	}
	return nil
}

func (m *UnifiedPerformanceMonitor) extractThroughput(metrics []Metric) *float64 {
	var sum float64
	var count int
	for _, metric := range metrics {
		if metric.MetricName == "throughput" {
			sum += metric.MetricValue
			count++
		}
	}
	if count > 0 {
		avg := sum / float64(count)
		return &avg
	}
	return nil
}

func (m *UnifiedPerformanceMonitor) extractCPUUsage(metrics []Metric) *float64 {
	var sum float64
	var count int
	for _, metric := range metrics {
		if metric.MetricName == "cpu_usage" {
			sum += metric.MetricValue
			count++
		}
	}
	if count > 0 {
		avg := sum / float64(count)
		return &avg
	}
	return nil
}

func (m *UnifiedPerformanceMonitor) extractMemoryUsage(metrics []Metric) *float64 {
	var sum float64
	var count int
	for _, metric := range metrics {
		if metric.MetricName == "memory_usage" {
			sum += metric.MetricValue
			count++
		}
	}
	if count > 0 {
		avg := sum / float64(count)
		return &avg
	}
	return nil
}

func (m *UnifiedPerformanceMonitor) extractDiskUsage(metrics []Metric) *float64 {
	var sum float64
	var count int
	for _, metric := range metrics {
		if metric.MetricName == "disk_usage" {
			sum += metric.MetricValue
			count++
		}
	}
	if count > 0 {
		avg := sum / float64(count)
		return &avg
	}
	return nil
}
