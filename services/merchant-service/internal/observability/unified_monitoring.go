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
	config     *UnifiedMonitoringConfig
	db         *sql.DB
	exporters  []MetricExporter
	alerters   []AlertHandler
	collectors map[string]MetricCollector
	logger     *zap.Logger
	metrics    *prometheus.Registry
	mu         sync.RWMutex
}

// UnifiedMonitoringConfig contains configuration for the unified monitoring system.
type UnifiedMonitoringConfig struct {
	DatabaseURL            string
	CollectionInterval     time.Duration
	RetentionPeriod        time.Duration
	UnifiedAlertThresholds map[string]UnifiedAlertThreshold
	Exporters              []ExporterConfig
	Components             []ComponentConfig
	BatchSize              int
	FlushInterval          time.Duration
}

// UnifiedAlertThreshold defines alerting thresholds for specific metrics.
type UnifiedAlertThreshold struct {
	UnifiedMetricName string
	Component         string
	Condition         string // 'gt', 'lt', 'eq', 'ne'
	Value             float64
	Severity          string
	Duration          time.Duration
	Cooldown          time.Duration
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
	Collect(ctx context.Context) ([]UnifiedMetric, error)
	GetType() string
	GetComponent() string
	GetInterval() time.Duration
	Start(ctx context.Context) error
	Stop() error
}

// UnifiedMetric represents a single performance metric.
type UnifiedMetric struct {
	ID                    string                 `json:"id"`
	Timestamp             time.Time              `json:"timestamp"`
	Component             string                 `json:"component"`
	ComponentInstance     string                 `json:"component_instance"`
	ServiceName           string                 `json:"service_name"`
	UnifiedMetricType     string                 `json:"metric_type"`
	UnifiedMetricCategory string                 `json:"metric_category"`
	UnifiedMetricName     string                 `json:"metric_name"`
	UnifiedMetricValue    float64                `json:"metric_value"`
	UnifiedMetricUnit     string                 `json:"metric_unit"`
	Tags                  map[string]string      `json:"tags"`
	Metadata              map[string]interface{} `json:"metadata"`
	RequestID             *string                `json:"request_id"`
	OperationID           *string                `json:"operation_id"`
	UserID                *string                `json:"user_id"`
	ConfidenceScore       *float64               `json:"confidence_score"`
	DataSource            string                 `json:"data_source"`
}

// UnifiedAlert represents a performance alert.
type UnifiedAlert struct {
	ID                    string                 `json:"id"`
	CreatedAt             time.Time              `json:"created_at"`
	UnifiedAlertType      string                 `json:"alert_type"`
	UnifiedAlertCategory  string                 `json:"alert_category"`
	Severity              string                 `json:"severity"`
	Component             string                 `json:"component"`
	ComponentInstance     string                 `json:"component_instance"`
	ServiceName           string                 `json:"service_name"`
	UnifiedAlertName      string                 `json:"alert_name"`
	Description           string                 `json:"description"`
	Condition             map[string]interface{} `json:"condition"`
	CurrentValue          *float64               `json:"current_value"`
	ThresholdValue        *float64               `json:"threshold_value"`
	Status                string                 `json:"status"`
	AcknowledgedBy        *string                `json:"acknowledged_by"`
	AcknowledgedAt        *time.Time             `json:"acknowledged_at"`
	ResolvedAt            *time.Time             `json:"resolved_at"`
	RelatedUnifiedMetrics []string               `json:"related_metrics"`
	RelatedRequests       []string               `json:"related_requests"`
	Tags                  map[string]string      `json:"tags"`
	Metadata              map[string]interface{} `json:"metadata"`
}

// HealthScore represents component health scores.
type HealthScore struct {
	ID                    string                 `json:"id"`
	Timestamp             time.Time              `json:"timestamp"`
	Component             string                 `json:"component"`
	ComponentInstance     string                 `json:"component_instance"`
	ServiceName           string                 `json:"service_name"`
	OverallHealth         float64                `json:"overall_health"`
	PerformanceHealth     float64                `json:"performance_health"`
	ResourceHealth        float64                `json:"resource_health"`
	AvailabilityHealth    float64                `json:"availability_health"`
	SecurityHealth        float64                `json:"security_health"`
	ActiveUnifiedAlerts   int                    `json:"active_alerts"`
	CriticalUnifiedAlerts int                    `json:"critical_alerts"`
	WarningUnifiedAlerts  int                    `json:"warning_alerts"`
	AvgResponseTime       *float64               `json:"avg_response_time"`
	ErrorRate             *float64               `json:"error_rate"`
	Throughput            *float64               `json:"throughput"`
	CPUUsage              *float64               `json:"cpu_usage"`
	MemoryUsage           *float64               `json:"memory_usage"`
	DiskUsage             *float64               `json:"disk_usage"`
	Tags                  map[string]string      `json:"tags"`
	Metadata              map[string]interface{} `json:"metadata"`
}

// MetricExporter interface for exporting metrics to external systems.
type MetricExporter interface {
	Export(ctx context.Context, metrics []UnifiedMetric) error
	GetType() string
	IsEnabled() bool
}

// AlertHandler interface for handling alerts.
type AlertHandler interface {
	HandleAlert(ctx context.Context, alert UnifiedAlert) error
	GetType() string
	IsEnabled() bool
}

// NewUnifiedPerformanceMonitor creates a new unified performance monitor.
func NewUnifiedPerformanceMonitor(config *UnifiedMonitoringConfig, db *sql.DB, logger *zap.Logger) *UnifiedPerformanceMonitor {
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

// RegisterUnifiedAlerter registers an alert handler.
func (m *UnifiedPerformanceMonitor) RegisterUnifiedAlerter(alerter AlertHandler) {
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

// RecordUnifiedMetric records a single performance metric.
func (m *UnifiedPerformanceMonitor) RecordUnifiedMetric(ctx context.Context, metric UnifiedMetric) error {
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
	if err := m.insertUnifiedMetric(ctx, metric); err != nil {
		m.logger.Error("Failed to insert metric", zap.Error(err))
		return fmt.Errorf("failed to insert metric: %w", err)
	}

	// Export to external systems
	if err := m.exportUnifiedMetrics(ctx, []UnifiedMetric{metric}); err != nil {
		m.logger.Error("Failed to export metrics", zap.Error(err))
		// Don't return error for export failures
	}

	// Check for alerts
	if err := m.checkUnifiedAlerts(ctx, metric); err != nil {
		m.logger.Error("Failed to check alerts", zap.Error(err))
		// Don't return error for alert check failures
	}

	return nil
}

// RecordUnifiedMetrics records multiple performance metrics in batch.
func (m *UnifiedPerformanceMonitor) RecordUnifiedMetrics(ctx context.Context, metrics []UnifiedMetric) error {
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
	if err := m.insertUnifiedMetricsBatch(ctx, metrics); err != nil {
		m.logger.Error("Failed to insert metrics batch", zap.Error(err))
		return fmt.Errorf("failed to insert metrics batch: %w", err)
	}

	// Export to external systems
	if err := m.exportUnifiedMetrics(ctx, metrics); err != nil {
		m.logger.Error("Failed to export metrics", zap.Error(err))
		// Don't return error for export failures
	}

	// Check for alerts
	for _, metric := range metrics {
		if err := m.checkUnifiedAlerts(ctx, metric); err != nil {
			m.logger.Error("Failed to check alerts", zap.Error(err))
			// Don't return error for alert check failures
		}
	}

	return nil
}

// GetUnifiedMetrics retrieves metrics based on filters.
func (m *UnifiedPerformanceMonitor) GetUnifiedMetrics(ctx context.Context, filters UnifiedMetricFilters) ([]UnifiedMetric, error) {
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

	if filters.UnifiedMetricType != "" {
		query += fmt.Sprintf(" AND metric_type = $%d", argIndex)
		args = append(args, filters.UnifiedMetricType)
		argIndex++
	}

	if filters.UnifiedMetricCategory != "" {
		query += fmt.Sprintf(" AND metric_category = $%d", argIndex)
		args = append(args, filters.UnifiedMetricCategory)
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

	var metrics []UnifiedMetric
	for rows.Next() {
		var metric UnifiedMetric
		var tagsJSON, metadataJSON sql.NullString
		var requestID, operationID, userID sql.NullString
		var confidenceScore sql.NullFloat64

		err := rows.Scan(
			&metric.ID, &metric.Timestamp, &metric.Component, &metric.ComponentInstance,
			&metric.ServiceName, &metric.UnifiedMetricType, &metric.UnifiedMetricCategory,
			&metric.UnifiedMetricName, &metric.UnifiedMetricValue, &metric.UnifiedMetricUnit,
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
			&healthScore.SecurityHealth, &healthScore.ActiveUnifiedAlerts,
			&healthScore.CriticalUnifiedAlerts, &healthScore.WarningUnifiedAlerts,
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

// UnifiedMetricFilters defines filters for querying metrics.
type UnifiedMetricFilters struct {
	Component             string
	ServiceName           string
	UnifiedMetricType     string
	UnifiedMetricCategory string
	StartTime             time.Time
	EndTime               time.Time
	Limit                 int
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
			m.collectUnifiedMetrics(ctx)
		}
	}
}

// collectUnifiedMetrics collects metrics from all registered collectors.
func (m *UnifiedPerformanceMonitor) collectUnifiedMetrics(ctx context.Context) {
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
			if err := m.RecordUnifiedMetrics(ctx, metrics); err != nil {
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
			m.processUnifiedAlerts(ctx)
		}
	}
}

// processUnifiedAlerts processes active alerts.
func (m *UnifiedPerformanceMonitor) processUnifiedAlerts(ctx context.Context) {
	// Get active alerts
	alerts, err := m.getActiveUnifiedAlerts(ctx)
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

// insertUnifiedMetric inserts a single metric into the database.
func (m *UnifiedPerformanceMonitor) insertUnifiedMetric(ctx context.Context, metric UnifiedMetric) error {
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
		metric.ServiceName, metric.UnifiedMetricType, metric.UnifiedMetricCategory,
		metric.UnifiedMetricName, metric.UnifiedMetricValue, metric.UnifiedMetricUnit,
		tagsJSON, metadataJSON, metric.RequestID, metric.OperationID,
		metric.UserID, metric.ConfidenceScore, metric.DataSource,
	)

	return err
}

// insertUnifiedMetricsBatch inserts multiple metrics into the database in batch.
func (m *UnifiedPerformanceMonitor) insertUnifiedMetricsBatch(ctx context.Context, metrics []UnifiedMetric) error {
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
			metric.ServiceName, metric.UnifiedMetricType, metric.UnifiedMetricCategory,
			metric.UnifiedMetricName, metric.UnifiedMetricValue, metric.UnifiedMetricUnit,
			tagsJSON, metadataJSON, metric.RequestID, metric.OperationID,
			metric.UserID, metric.ConfidenceScore, metric.DataSource,
		)
		if err != nil {
			return fmt.Errorf("failed to insert metric %s: %w", metric.ID, err)
		}
	}

	return nil
}

// exportUnifiedMetrics exports metrics to all registered exporters.
func (m *UnifiedPerformanceMonitor) exportUnifiedMetrics(ctx context.Context, metrics []UnifiedMetric) error {
	for _, exporter := range m.exporters {
		if exporter.IsEnabled() {
			if err := exporter.Export(ctx, metrics); err != nil {
				m.logger.Error("Failed to export metrics", zap.String("exporter", exporter.GetType()), zap.Error(err))
			}
		}
	}
	return nil
}

// checkUnifiedAlerts checks if a metric triggers any alerts.
func (m *UnifiedPerformanceMonitor) checkUnifiedAlerts(ctx context.Context, metric UnifiedMetric) error {
	// Check against configured thresholds
	for _, threshold := range m.config.UnifiedAlertThresholds {
		if threshold.Component == metric.Component && threshold.UnifiedMetricName == metric.UnifiedMetricName {
			if m.evaluateThreshold(metric, threshold) {
				alert := UnifiedAlert{
					ID:                    uuid.New().String(),
					CreatedAt:             time.Now(),
					UnifiedAlertType:      "threshold",
					UnifiedAlertCategory:  "performance",
					Severity:              threshold.Severity,
					Component:             metric.Component,
					ComponentInstance:     metric.ComponentInstance,
					ServiceName:           metric.ServiceName,
					UnifiedAlertName:      fmt.Sprintf("%s threshold exceeded", metric.UnifiedMetricName),
					Description:           fmt.Sprintf("UnifiedMetric %s exceeded threshold %f", metric.UnifiedMetricName, threshold.Value),
					Condition:             map[string]interface{}{"threshold": threshold.Value, "condition": threshold.Condition},
					CurrentValue:          &metric.UnifiedMetricValue,
					ThresholdValue:        &threshold.Value,
					Status:                "active",
					RelatedUnifiedMetrics: []string{metric.ID},
					Tags:                  metric.Tags,
					Metadata:              metric.Metadata,
				}

				if err := m.createUnifiedAlert(ctx, alert); err != nil {
					m.logger.Error("Failed to create alert", zap.Error(err))
				}
			}
		}
	}
	return nil
}

// evaluateThreshold evaluates if a metric meets alert threshold conditions.
func (m *UnifiedPerformanceMonitor) evaluateThreshold(metric UnifiedMetric, threshold UnifiedAlertThreshold) bool {
	switch threshold.Condition {
	case "gt":
		return metric.UnifiedMetricValue > threshold.Value
	case "lt":
		return metric.UnifiedMetricValue < threshold.Value
	case "eq":
		return metric.UnifiedMetricValue == threshold.Value
	case "ne":
		return metric.UnifiedMetricValue != threshold.Value
	default:
		return false
	}
}

// createUnifiedAlert creates a new alert in the database.
func (m *UnifiedPerformanceMonitor) createUnifiedAlert(ctx context.Context, alert UnifiedAlert) error {
	conditionJSON, _ := json.Marshal(alert.Condition)
	tagsJSON, _ := json.Marshal(alert.Tags)
	metadataJSON, _ := json.Marshal(alert.Metadata)
	relatedUnifiedMetricsJSON, _ := json.Marshal(alert.RelatedUnifiedMetrics)
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
		alert.ID, alert.CreatedAt, alert.UnifiedAlertType, alert.UnifiedAlertCategory, alert.Severity,
		alert.Component, alert.ComponentInstance, alert.ServiceName, alert.UnifiedAlertName,
		alert.Description, conditionJSON, alert.CurrentValue, alert.ThresholdValue,
		alert.Status, relatedUnifiedMetricsJSON, relatedRequestsJSON, tagsJSON, metadataJSON,
	)

	return err
}

// getActiveUnifiedAlerts retrieves active alerts from the database.
func (m *UnifiedPerformanceMonitor) getActiveUnifiedAlerts(ctx context.Context) ([]UnifiedAlert, error) {
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

	var alerts []UnifiedAlert
	for rows.Next() {
		var alert UnifiedAlert
		var conditionJSON, tagsJSON, metadataJSON, relatedUnifiedMetricsJSON, relatedRequestsJSON sql.NullString
		var acknowledgedBy sql.NullString
		var acknowledgedAt, resolvedAt sql.NullTime

		err := rows.Scan(
			&alert.ID, &alert.CreatedAt, &alert.UnifiedAlertType, &alert.UnifiedAlertCategory,
			&alert.Severity, &alert.Component, &alert.ComponentInstance,
			&alert.ServiceName, &alert.UnifiedAlertName, &alert.Description,
			&conditionJSON, &alert.CurrentValue, &alert.ThresholdValue,
			&alert.Status, &acknowledgedBy, &acknowledgedAt, &resolvedAt,
			&relatedUnifiedMetricsJSON, &relatedRequestsJSON, &tagsJSON, &metadataJSON,
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
		if relatedUnifiedMetricsJSON.Valid {
			json.Unmarshal([]byte(relatedUnifiedMetricsJSON.String), &alert.RelatedUnifiedMetrics)
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
	metrics, err := m.GetUnifiedMetrics(ctx, UnifiedMetricFilters{
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
	activeUnifiedAlerts, criticalUnifiedAlerts, warningUnifiedAlerts, err := m.getUnifiedAlertCounts(ctx, component)
	if err != nil {
		m.logger.Error("Failed to get alert counts", zap.String("component", component), zap.Error(err))
	} else {
		healthScore.ActiveUnifiedAlerts = activeUnifiedAlerts
		healthScore.CriticalUnifiedAlerts = criticalUnifiedAlerts
		healthScore.WarningUnifiedAlerts = warningUnifiedAlerts
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
func (m *UnifiedPerformanceMonitor) calculatePerformanceHealth(metrics []UnifiedMetric) float64 {
	var responseTimeSum float64
	var responseTimeCount int
	var errorRateSum float64
	var errorRateCount int

	for _, metric := range metrics {
		if metric.UnifiedMetricCategory == "latency" || metric.UnifiedMetricName == "response_time" {
			responseTimeSum += metric.UnifiedMetricValue
			responseTimeCount++
		}
		if metric.UnifiedMetricCategory == "error_rate" {
			errorRateSum += metric.UnifiedMetricValue
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
func (m *UnifiedPerformanceMonitor) calculateResourceHealth(metrics []UnifiedMetric) float64 {
	var cpuSum, memorySum, diskSum float64
	var cpuCount, memoryCount, diskCount int

	for _, metric := range metrics {
		if metric.UnifiedMetricName == "cpu_usage" {
			cpuSum += metric.UnifiedMetricValue
			cpuCount++
		}
		if metric.UnifiedMetricName == "memory_usage" {
			memorySum += metric.UnifiedMetricValue
			memoryCount++
		}
		if metric.UnifiedMetricName == "disk_usage" {
			diskSum += metric.UnifiedMetricValue
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
func (m *UnifiedPerformanceMonitor) calculateAvailabilityHealth(metrics []UnifiedMetric) float64 {
	// Simplified availability calculation
	// In production, you'd want more sophisticated logic based on uptime, error rates, etc.
	return 1.0
}

// calculateSecurityHealth calculates security health score.
func (m *UnifiedPerformanceMonitor) calculateSecurityHealth(metrics []UnifiedMetric) float64 {
	// Simplified security calculation
	// In production, you'd want more sophisticated logic based on security metrics
	return 1.0
}

// getUnifiedAlertCounts retrieves alert counts for a component.
func (m *UnifiedPerformanceMonitor) getUnifiedAlertCounts(ctx context.Context, component string) (int, int, int, error) {
	query := `
		SELECT 
			COUNT(*) as total_alerts,
			COUNT(CASE WHEN severity = 'critical' THEN 1 END) as critical_alerts,
			COUNT(CASE WHEN severity = 'warning' THEN 1 END) as warning_alerts
		FROM unified_performance_alerts
		WHERE component = $1 AND status = 'active'
	`

	var totalUnifiedAlerts, criticalUnifiedAlerts, warningUnifiedAlerts int
	err := m.db.QueryRowContext(ctx, query, component).Scan(&totalUnifiedAlerts, &criticalUnifiedAlerts, &warningUnifiedAlerts)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get alert counts: %w", err)
	}

	return totalUnifiedAlerts, criticalUnifiedAlerts, warningUnifiedAlerts, nil
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
		healthScore.SecurityHealth, healthScore.ActiveUnifiedAlerts,
		healthScore.CriticalUnifiedAlerts, healthScore.WarningUnifiedAlerts,
		healthScore.AvgResponseTime, healthScore.ErrorRate,
		healthScore.Throughput, healthScore.CPUUsage,
		healthScore.MemoryUsage, healthScore.DiskUsage,
		tagsJSON, metadataJSON,
	)

	return err
}

// Helper functions for extracting specific metrics
func (m *UnifiedPerformanceMonitor) extractAvgResponseTime(metrics []UnifiedMetric) *float64 {
	var sum float64
	var count int
	for _, metric := range metrics {
		if metric.UnifiedMetricName == "response_time" {
			sum += metric.UnifiedMetricValue
			count++
		}
	}
	if count > 0 {
		avg := sum / float64(count)
		return &avg
	}
	return nil
}

func (m *UnifiedPerformanceMonitor) extractErrorRate(metrics []UnifiedMetric) *float64 {
	var sum float64
	var count int
	for _, metric := range metrics {
		if metric.UnifiedMetricName == "error_rate" {
			sum += metric.UnifiedMetricValue
			count++
		}
	}
	if count > 0 {
		avg := sum / float64(count)
		return &avg
	}
	return nil
}

func (m *UnifiedPerformanceMonitor) extractThroughput(metrics []UnifiedMetric) *float64 {
	var sum float64
	var count int
	for _, metric := range metrics {
		if metric.UnifiedMetricName == "throughput" {
			sum += metric.UnifiedMetricValue
			count++
		}
	}
	if count > 0 {
		avg := sum / float64(count)
		return &avg
	}
	return nil
}

func (m *UnifiedPerformanceMonitor) extractCPUUsage(metrics []UnifiedMetric) *float64 {
	var sum float64
	var count int
	for _, metric := range metrics {
		if metric.UnifiedMetricName == "cpu_usage" {
			sum += metric.UnifiedMetricValue
			count++
		}
	}
	if count > 0 {
		avg := sum / float64(count)
		return &avg
	}
	return nil
}

func (m *UnifiedPerformanceMonitor) extractMemoryUsage(metrics []UnifiedMetric) *float64 {
	var sum float64
	var count int
	for _, metric := range metrics {
		if metric.UnifiedMetricName == "memory_usage" {
			sum += metric.UnifiedMetricValue
			count++
		}
	}
	if count > 0 {
		avg := sum / float64(count)
		return &avg
	}
	return nil
}

func (m *UnifiedPerformanceMonitor) extractDiskUsage(metrics []UnifiedMetric) *float64 {
	var sum float64
	var count int
	for _, metric := range metrics {
		if metric.UnifiedMetricName == "disk_usage" {
			sum += metric.UnifiedMetricValue
			count++
		}
	}
	if count > 0 {
		avg := sum / float64(count)
		return &avg
	}
	return nil
}
