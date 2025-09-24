package monitoring

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// UnifiedMonitoringService provides a single interface for all monitoring operations
// using the unified monitoring schema
type UnifiedMonitoringService struct {
	db     *sql.DB
	logger *log.Logger
}

// NewUnifiedMonitoringService creates a new instance of the unified monitoring service
func NewUnifiedMonitoringService(db *sql.DB, logger *log.Logger) *UnifiedMonitoringService {
	return &UnifiedMonitoringService{
		db:     db,
		logger: logger,
	}
}

// MetricType represents the type of metric being recorded
type MetricType string

const (
	MetricTypePerformance MetricType = "performance"
	MetricTypeResource    MetricType = "resource"
	MetricTypeBusiness    MetricType = "business"
	MetricTypeSecurity    MetricType = "security"
)

// MetricCategory represents the category of metric
type MetricCategory string

const (
	MetricCategoryLatency    MetricCategory = "latency"
	MetricCategoryThroughput MetricCategory = "throughput"
	MetricCategoryErrorRate  MetricCategory = "error_rate"
	MetricCategoryMemory     MetricCategory = "memory"
	MetricCategoryCPU        MetricCategory = "cpu"
	MetricCategoryConnection MetricCategory = "connection"
	MetricCategoryCache      MetricCategory = "cache"
	MetricCategoryValidation MetricCategory = "validation"
	MetricCategoryAccuracy   MetricCategory = "accuracy"
	MetricCategoryGeneral    MetricCategory = "general"
)

// UnifiedMetric represents a metric in the unified monitoring system
type UnifiedMetric struct {
	ID                uuid.UUID              `json:"id"`
	Timestamp         time.Time              `json:"timestamp"`
	Component         string                 `json:"component"`
	ComponentInstance string                 `json:"component_instance"`
	ServiceName       string                 `json:"service_name"`
	MetricType        MetricType             `json:"metric_type"`
	MetricCategory    MetricCategory         `json:"metric_category"`
	MetricName        string                 `json:"metric_name"`
	MetricValue       float64                `json:"metric_value"`
	MetricUnit        string                 `json:"metric_unit"`
	Tags              map[string]interface{} `json:"tags"`
	Metadata          map[string]interface{} `json:"metadata"`
	RequestID         *uuid.UUID             `json:"request_id,omitempty"`
	OperationID       *uuid.UUID             `json:"operation_id,omitempty"`
	UserID            *uuid.UUID             `json:"user_id,omitempty"`
	ConfidenceScore   float64                `json:"confidence_score"`
	DataSource        string                 `json:"data_source"`
	CreatedAt         time.Time              `json:"created_at"`
}

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeThreshold    AlertType = "threshold"
	AlertTypeAnomaly      AlertType = "anomaly"
	AlertTypeTrend        AlertType = "trend"
	AlertTypeAvailability AlertType = "availability"
)

// AlertCategory represents the category of alert
type AlertCategory string

const (
	AlertCategoryPerformance AlertCategory = "performance"
	AlertCategoryResource    AlertCategory = "resource"
	AlertCategoryBusiness    AlertCategory = "business"
	AlertCategorySecurity    AlertCategory = "security"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityInfo     AlertSeverity = "info"
)

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "active"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved     AlertStatus = "resolved"
	AlertStatusSuppressed   AlertStatus = "suppressed"
)

// UnifiedAlert represents an alert in the unified monitoring system
type UnifiedAlert struct {
	ID                uuid.UUID              `json:"id"`
	CreatedAt         time.Time              `json:"created_at"`
	AlertType         AlertType              `json:"alert_type"`
	AlertCategory     AlertCategory          `json:"alert_category"`
	Severity          AlertSeverity          `json:"severity"`
	Component         string                 `json:"component"`
	ComponentInstance string                 `json:"component_instance"`
	ServiceName       string                 `json:"service_name"`
	AlertName         string                 `json:"alert_name"`
	Description       string                 `json:"description"`
	Condition         map[string]interface{} `json:"condition"`
	CurrentValue      *float64               `json:"current_value,omitempty"`
	ThresholdValue    *float64               `json:"threshold_value,omitempty"`
	Status            AlertStatus            `json:"status"`
	AcknowledgedBy    *uuid.UUID             `json:"acknowledged_by,omitempty"`
	AcknowledgedAt    *time.Time             `json:"acknowledged_at,omitempty"`
	ResolvedAt        *time.Time             `json:"resolved_at,omitempty"`
	RelatedMetrics    []uuid.UUID            `json:"related_metrics"`
	RelatedRequests   []uuid.UUID            `json:"related_requests"`
	Tags              map[string]interface{} `json:"tags"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// RecordMetric records a metric in the unified monitoring system
func (ums *UnifiedMonitoringService) RecordMetric(ctx context.Context, metric *UnifiedMetric) error {
	query := `
		INSERT INTO unified_performance_metrics (
			id, timestamp, component, component_instance, service_name,
			metric_type, metric_category, metric_name, metric_value, metric_unit,
			tags, metadata, request_id, operation_id, user_id,
			confidence_score, data_source, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
	`

	// Convert maps to JSONB
	tagsJSON, err := json.Marshal(metric.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	metadataJSON, err := json.Marshal(metric.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = ums.db.ExecContext(ctx, query,
		metric.ID,
		metric.Timestamp,
		metric.Component,
		metric.ComponentInstance,
		metric.ServiceName,
		string(metric.MetricType),
		string(metric.MetricCategory),
		metric.MetricName,
		metric.MetricValue,
		metric.MetricUnit,
		tagsJSON,
		metadataJSON,
		metric.RequestID,
		metric.OperationID,
		metric.UserID,
		metric.ConfidenceScore,
		metric.DataSource,
		metric.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to record metric: %w", err)
	}

	return nil
}

// RecordAlert records an alert in the unified monitoring system
func (ums *UnifiedMonitoringService) RecordAlert(ctx context.Context, alert *UnifiedAlert) error {
	query := `
		INSERT INTO unified_performance_alerts (
			id, created_at, alert_type, alert_category, severity,
			component, component_instance, service_name, alert_name, description,
			condition, current_value, threshold_value, status,
			acknowledged_by, acknowledged_at, resolved_at,
			related_metrics, related_requests, tags, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
		)
	`

	// Convert maps to JSONB
	conditionJSON, err := json.Marshal(alert.Condition)
	if err != nil {
		return fmt.Errorf("failed to marshal condition: %w", err)
	}

	tagsJSON, err := json.Marshal(alert.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	metadataJSON, err := json.Marshal(alert.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = ums.db.ExecContext(ctx, query,
		alert.ID,
		alert.CreatedAt,
		string(alert.AlertType),
		string(alert.AlertCategory),
		string(alert.Severity),
		alert.Component,
		alert.ComponentInstance,
		alert.ServiceName,
		alert.AlertName,
		alert.Description,
		conditionJSON,
		alert.CurrentValue,
		alert.ThresholdValue,
		string(alert.Status),
		alert.AcknowledgedBy,
		alert.AcknowledgedAt,
		alert.ResolvedAt,
		pq.Array(alert.RelatedMetrics),
		pq.Array(alert.RelatedRequests),
		tagsJSON,
		metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to record alert: %w", err)
	}

	return nil
}

// GetMetrics retrieves metrics from the unified monitoring system
func (ums *UnifiedMonitoringService) GetMetrics(ctx context.Context, filters *MetricFilters) ([]*UnifiedMetric, error) {
	query := `
		SELECT 
			id, timestamp, component, component_instance, service_name,
			metric_type, metric_category, metric_name, metric_value, metric_unit,
			tags, metadata, request_id, operation_id, user_id,
			confidence_score, data_source, created_at
		FROM unified_performance_metrics
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if filters != nil {
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
			args = append(args, string(filters.MetricType))
			argIndex++
		}

		if filters.MetricCategory != "" {
			query += fmt.Sprintf(" AND metric_category = $%d", argIndex)
			args = append(args, string(filters.MetricCategory))
			argIndex++
		}

		if filters.StartTime != nil {
			query += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
			args = append(args, *filters.StartTime)
			argIndex++
		}

		if filters.EndTime != nil {
			query += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
			args = append(args, *filters.EndTime)
			argIndex++
		}

		if filters.Limit > 0 {
			query += fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d", argIndex)
			args = append(args, filters.Limit)
		} else {
			query += " ORDER BY timestamp DESC"
		}
	}

	rows, err := ums.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*UnifiedMetric
	for rows.Next() {
		metric := &UnifiedMetric{}
		var tagsJSON, metadataJSON []byte
		var metricType, metricCategory string

		err := rows.Scan(
			&metric.ID,
			&metric.Timestamp,
			&metric.Component,
			&metric.ComponentInstance,
			&metric.ServiceName,
			&metricType,
			&metricCategory,
			&metric.MetricName,
			&metric.MetricValue,
			&metric.MetricUnit,
			&tagsJSON,
			&metadataJSON,
			&metric.RequestID,
			&metric.OperationID,
			&metric.UserID,
			&metric.ConfidenceScore,
			&metric.DataSource,
			&metric.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}

		metric.MetricType = MetricType(metricType)
		metric.MetricCategory = MetricCategory(metricCategory)

		// Unmarshal JSON fields
		if err := json.Unmarshal(tagsJSON, &metric.Tags); err != nil {
			ums.logger.Printf("Warning: failed to unmarshal tags for metric %s: %v", metric.ID, err)
			metric.Tags = make(map[string]interface{})
		}

		if err := json.Unmarshal(metadataJSON, &metric.Metadata); err != nil {
			ums.logger.Printf("Warning: failed to unmarshal metadata for metric %s: %v", metric.ID, err)
			metric.Metadata = make(map[string]interface{})
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetAlerts retrieves alerts from the unified monitoring system
func (ums *UnifiedMonitoringService) GetAlerts(ctx context.Context, filters *AlertFilters) ([]*UnifiedAlert, error) {
	query := `
		SELECT 
			id, created_at, alert_type, alert_category, severity,
			component, component_instance, service_name, alert_name, description,
			condition, current_value, threshold_value, status,
			acknowledged_by, acknowledged_at, resolved_at,
			related_metrics, related_requests, tags, metadata
		FROM unified_performance_alerts
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if filters != nil {
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

		if filters.AlertCategory != "" {
			query += fmt.Sprintf(" AND alert_category = $%d", argIndex)
			args = append(args, string(filters.AlertCategory))
			argIndex++
		}

		if filters.Severity != "" {
			query += fmt.Sprintf(" AND severity = $%d", argIndex)
			args = append(args, string(filters.Severity))
			argIndex++
		}

		if filters.Status != "" {
			query += fmt.Sprintf(" AND status = $%d", argIndex)
			args = append(args, string(filters.Status))
			argIndex++
		}

		if filters.StartTime != nil {
			query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
			args = append(args, *filters.StartTime)
			argIndex++
		}

		if filters.EndTime != nil {
			query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
			args = append(args, *filters.EndTime)
			argIndex++
		}

		if filters.Limit > 0 {
			query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argIndex)
			args = append(args, filters.Limit)
		} else {
			query += " ORDER BY created_at DESC"
		}
	}

	rows, err := ums.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*UnifiedAlert
	for rows.Next() {
		alert := &UnifiedAlert{}
		var conditionJSON, tagsJSON, metadataJSON []byte
		var alertType, alertCategory, severity, status string
		var relatedMetrics, relatedRequests pq.StringArray

		err := rows.Scan(
			&alert.ID,
			&alert.CreatedAt,
			&alertType,
			&alertCategory,
			&severity,
			&alert.Component,
			&alert.ComponentInstance,
			&alert.ServiceName,
			&alert.AlertName,
			&alert.Description,
			&conditionJSON,
			&alert.CurrentValue,
			&alert.ThresholdValue,
			&status,
			&alert.AcknowledgedBy,
			&alert.AcknowledgedAt,
			&alert.ResolvedAt,
			&relatedMetrics,
			&relatedRequests,
			&tagsJSON,
			&metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}

		alert.AlertType = AlertType(alertType)
		alert.AlertCategory = AlertCategory(alertCategory)
		alert.Severity = AlertSeverity(severity)
		alert.Status = AlertStatus(status)

		// Convert string arrays to UUID arrays
		for _, metricIDStr := range relatedMetrics {
			if metricID, err := uuid.Parse(metricIDStr); err == nil {
				alert.RelatedMetrics = append(alert.RelatedMetrics, metricID)
			}
		}

		for _, requestIDStr := range relatedRequests {
			if requestID, err := uuid.Parse(requestIDStr); err == nil {
				alert.RelatedRequests = append(alert.RelatedRequests, requestID)
			}
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(conditionJSON, &alert.Condition); err != nil {
			ums.logger.Printf("Warning: failed to unmarshal condition for alert %s: %v", alert.ID, err)
			alert.Condition = make(map[string]interface{})
		}

		if err := json.Unmarshal(tagsJSON, &alert.Tags); err != nil {
			ums.logger.Printf("Warning: failed to unmarshal tags for alert %s: %v", alert.ID, err)
			alert.Tags = make(map[string]interface{})
		}

		if err := json.Unmarshal(metadataJSON, &alert.Metadata); err != nil {
			ums.logger.Printf("Warning: failed to unmarshal metadata for alert %s: %v", alert.ID, err)
			alert.Metadata = make(map[string]interface{})
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// UpdateAlertStatus updates the status of an alert
func (ums *UnifiedMonitoringService) UpdateAlertStatus(ctx context.Context, alertID uuid.UUID, status AlertStatus, userID *uuid.UUID) error {
	query := `
		UPDATE unified_performance_alerts 
		SET status = $1, acknowledged_by = $2, acknowledged_at = $3, resolved_at = $4
		WHERE id = $5
	`

	var acknowledgedAt, resolvedAt *time.Time
	now := time.Now()

	switch status {
	case AlertStatusAcknowledged:
		acknowledgedAt = &now
	case AlertStatusResolved:
		resolvedAt = &now
	}

	_, err := ums.db.ExecContext(ctx, query, string(status), userID, acknowledgedAt, resolvedAt, alertID)
	if err != nil {
		return fmt.Errorf("failed to update alert status: %w", err)
	}

	return nil
}

// MetricFilters represents filters for querying metrics
type MetricFilters struct {
	Component      string
	ServiceName    string
	MetricType     MetricType
	MetricCategory MetricCategory
	StartTime      *time.Time
	EndTime        *time.Time
	Limit          int
}

// AlertFilters represents filters for querying alerts
type AlertFilters struct {
	Component     string
	ServiceName   string
	AlertCategory AlertCategory
	Severity      AlertSeverity
	Status        AlertStatus
	StartTime     *time.Time
	EndTime       *time.Time
	Limit         int
}

// GetMetricsSummary returns a summary of metrics for a given time period
func (ums *UnifiedMonitoringService) GetMetricsSummary(ctx context.Context, component, serviceName string, startTime, endTime time.Time) (map[string]interface{}, error) {
	query := `
		SELECT 
			metric_category,
			COUNT(*) as count,
			AVG(metric_value) as avg_value,
			MIN(metric_value) as min_value,
			MAX(metric_value) as max_value,
			STDDEV(metric_value) as stddev_value
		FROM unified_performance_metrics
		WHERE component = $1 AND service_name = $2 AND timestamp BETWEEN $3 AND $4
		GROUP BY metric_category
		ORDER BY metric_category
	`

	rows, err := ums.db.QueryContext(ctx, query, component, serviceName, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics summary: %w", err)
	}
	defer rows.Close()

	summary := make(map[string]interface{})
	summary["component"] = component
	summary["service_name"] = serviceName
	summary["start_time"] = startTime
	summary["end_time"] = endTime
	summary["categories"] = make(map[string]interface{})

	for rows.Next() {
		var category string
		var count int64
		var avgValue, minValue, maxValue, stddevValue sql.NullFloat64

		err := rows.Scan(&category, &count, &avgValue, &minValue, &maxValue, &stddevValue)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metrics summary: %w", err)
		}

		categoryData := map[string]interface{}{
			"count": count,
		}

		if avgValue.Valid {
			categoryData["avg_value"] = avgValue.Float64
		}
		if minValue.Valid {
			categoryData["min_value"] = minValue.Float64
		}
		if maxValue.Valid {
			categoryData["max_value"] = maxValue.Float64
		}
		if stddevValue.Valid {
			categoryData["stddev_value"] = stddevValue.Float64
		}

		summary["categories"].(map[string]interface{})[category] = categoryData
	}

	return summary, nil
}

// GetActiveAlertsCount returns the count of active alerts by severity
func (ums *UnifiedMonitoringService) GetActiveAlertsCount(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT severity, COUNT(*) as count
		FROM unified_performance_alerts
		WHERE status = 'active'
		GROUP BY severity
	`

	rows, err := ums.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active alerts count: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var severity string
		var count int

		err := rows.Scan(&severity, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan active alerts count: %w", err)
		}

		counts[severity] = count
	}

	return counts, nil
}
