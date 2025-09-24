package monitoring

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// MonitoringAdapter provides backward compatibility for existing monitoring code
// while gradually migrating to the unified monitoring system
type MonitoringAdapter struct {
	unifiedService *UnifiedMonitoringService
	logger         *log.Logger
}

// NewMonitoringAdapter creates a new monitoring adapter
func NewMonitoringAdapter(db *sql.DB, logger *log.Logger) *MonitoringAdapter {
	return &MonitoringAdapter{
		unifiedService: NewUnifiedMonitoringService(db, logger),
		logger:         logger,
	}
}

// DatabaseMetrics represents the legacy database metrics structure
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

// RecordDatabaseMetrics records database metrics using the unified monitoring system
func (ma *MonitoringAdapter) RecordDatabaseMetrics(ctx context.Context, metrics *DatabaseMetrics) error {
	// Record connection metrics
	connectionMetric := &UnifiedMetric{
		ID:                uuid.New(),
		Timestamp:         metrics.Timestamp,
		Component:         "database",
		ComponentInstance: "main",
		ServiceName:       "database_monitor",
		MetricType:        MetricTypeResource,
		MetricCategory:    MetricCategoryConnection,
		MetricName:        "connection_count",
		MetricValue:       float64(metrics.ConnectionCount),
		MetricUnit:        "count",
		Tags: map[string]interface{}{
			"active_connections": metrics.ActiveConnections,
			"idle_connections":   metrics.IdleConnections,
			"max_connections":    metrics.MaxConnections,
		},
		Metadata: map[string]interface{}{
			"original_type": "DatabaseMetrics",
			"uptime":        metrics.Uptime.String(),
		},
		ConfidenceScore: 0.95,
		DataSource:      "database_monitor_adapter",
		CreatedAt:       time.Now(),
	}

	if err := ma.unifiedService.RecordMetric(ctx, connectionMetric); err != nil {
		return fmt.Errorf("failed to record connection metric: %w", err)
	}

	// Record query performance metrics
	queryMetric := &UnifiedMetric{
		ID:                uuid.New(),
		Timestamp:         metrics.Timestamp,
		Component:         "database",
		ComponentInstance: "main",
		ServiceName:       "database_monitor",
		MetricType:        MetricTypePerformance,
		MetricCategory:    MetricCategoryLatency,
		MetricName:        "avg_query_time",
		MetricValue:       metrics.AvgQueryTime,
		MetricUnit:        "ms",
		Tags: map[string]interface{}{
			"total_queries":  metrics.QueryCount,
			"slow_queries":   metrics.SlowQueryCount,
			"error_count":    metrics.ErrorCount,
			"max_query_time": metrics.MaxQueryTime,
		},
		Metadata: map[string]interface{}{
			"original_type": "DatabaseMetrics",
		},
		ConfidenceScore: 0.95,
		DataSource:      "database_monitor_adapter",
		CreatedAt:       time.Now(),
	}

	if err := ma.unifiedService.RecordMetric(ctx, queryMetric); err != nil {
		return fmt.Errorf("failed to record query metric: %w", err)
	}

	// Record cache performance metrics
	cacheMetric := &UnifiedMetric{
		ID:                uuid.New(),
		Timestamp:         metrics.Timestamp,
		Component:         "database",
		ComponentInstance: "main",
		ServiceName:       "database_monitor",
		MetricType:        MetricTypePerformance,
		MetricCategory:    MetricCategoryCache,
		MetricName:        "cache_hit_ratio",
		MetricValue:       metrics.CacheHitRatio,
		MetricUnit:        "percent",
		Tags: map[string]interface{}{
			"lock_count":     metrics.LockCount,
			"deadlock_count": metrics.DeadlockCount,
		},
		Metadata: map[string]interface{}{
			"original_type": "DatabaseMetrics",
		},
		ConfidenceScore: 0.95,
		DataSource:      "database_monitor_adapter",
		CreatedAt:       time.Now(),
	}

	if err := ma.unifiedService.RecordMetric(ctx, cacheMetric); err != nil {
		return fmt.Errorf("failed to record cache metric: %w", err)
	}

	// Record database size metrics
	sizeMetric := &UnifiedMetric{
		ID:                uuid.New(),
		Timestamp:         metrics.Timestamp,
		Component:         "database",
		ComponentInstance: "main",
		ServiceName:       "database_monitor",
		MetricType:        MetricTypeResource,
		MetricCategory:    MetricCategoryGeneral,
		MetricName:        "database_size",
		MetricValue:       float64(metrics.DatabaseSize),
		MetricUnit:        "bytes",
		Tags: map[string]interface{}{
			"table_count": len(metrics.TableSizes),
			"index_count": len(metrics.IndexSizes),
		},
		Metadata: map[string]interface{}{
			"original_type": "DatabaseMetrics",
			"table_sizes":   metrics.TableSizes,
			"index_sizes":   metrics.IndexSizes,
		},
		ConfidenceScore: 0.95,
		DataSource:      "database_monitor_adapter",
		CreatedAt:       time.Now(),
	}

	if err := ma.unifiedService.RecordMetric(ctx, sizeMetric); err != nil {
		return fmt.Errorf("failed to record size metric: %w", err)
	}

	return nil
}

// PerformanceMetric represents a legacy performance metric
type PerformanceMetric struct {
	Name      string                 `json:"name"`
	Value     float64                `json:"value"`
	Unit      string                 `json:"unit"`
	Timestamp time.Time              `json:"timestamp"`
	Tags      map[string]interface{} `json:"tags"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// RecordPerformanceMetric records a performance metric using the unified monitoring system
func (ma *MonitoringAdapter) RecordPerformanceMetric(ctx context.Context, component, serviceName string, metric *PerformanceMetric) error {
	// Determine metric type and category based on name and value
	metricType := MetricTypePerformance
	metricCategory := MetricCategoryGeneral

	switch {
	case contains(metric.Name, "response_time", "latency", "duration"):
		metricCategory = MetricCategoryLatency
	case contains(metric.Name, "throughput", "requests_per_second", "tps"):
		metricCategory = MetricCategoryThroughput
	case contains(metric.Name, "error_rate", "error_count", "errors"):
		metricCategory = MetricCategoryErrorRate
	case contains(metric.Name, "memory", "ram"):
		metricType = MetricTypeResource
		metricCategory = MetricCategoryMemory
	case contains(metric.Name, "cpu", "processor"):
		metricType = MetricTypeResource
		metricCategory = MetricCategoryCPU
	case contains(metric.Name, "connection", "pool"):
		metricType = MetricTypeResource
		metricCategory = MetricCategoryConnection
	case contains(metric.Name, "cache", "hit_ratio"):
		metricCategory = MetricCategoryCache
	}

	unifiedMetric := &UnifiedMetric{
		ID:                uuid.New(),
		Timestamp:         metric.Timestamp,
		Component:         component,
		ComponentInstance: "main",
		ServiceName:       serviceName,
		MetricType:        metricType,
		MetricCategory:    metricCategory,
		MetricName:        metric.Name,
		MetricValue:       metric.Value,
		MetricUnit:        metric.Unit,
		Tags:              metric.Tags,
		Metadata: map[string]interface{}{
			"original_type":   "PerformanceMetric",
			"legacy_metadata": metric.Metadata,
		},
		ConfidenceScore: 0.90,
		DataSource:      "performance_monitor_adapter",
		CreatedAt:       time.Now(),
	}

	return ma.unifiedService.RecordMetric(ctx, unifiedMetric)
}

// RecordAlert records an alert using the unified monitoring system
func (ma *MonitoringAdapter) RecordAlert(ctx context.Context, component, serviceName, alertName, description string, severity AlertSeverity, condition map[string]interface{}) error {
	alert := &UnifiedAlert{
		ID:                uuid.New(),
		CreatedAt:         time.Now(),
		AlertType:         AlertTypeThreshold,
		AlertCategory:     AlertCategoryPerformance,
		Severity:          severity,
		Component:         component,
		ComponentInstance: "main",
		ServiceName:       serviceName,
		AlertName:         alertName,
		Description:       description,
		Condition:         condition,
		Status:            AlertStatusActive,
		Tags:              make(map[string]interface{}),
		Metadata: map[string]interface{}{
			"original_type": "LegacyAlert",
		},
	}

	return ma.unifiedService.RecordAlert(ctx, alert)
}

// GetDatabaseMetricsSummary retrieves database metrics summary using the unified system
func (ma *MonitoringAdapter) GetDatabaseMetricsSummary(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error) {
	return ma.unifiedService.GetMetricsSummary(ctx, "database", "database_monitor", startTime, endTime)
}

// GetPerformanceMetrics retrieves performance metrics using the unified system
func (ma *MonitoringAdapter) GetPerformanceMetrics(ctx context.Context, component, serviceName string, startTime, endTime time.Time, limit int) ([]*UnifiedMetric, error) {
	filters := &MetricFilters{
		Component:   component,
		ServiceName: serviceName,
		StartTime:   &startTime,
		EndTime:     &endTime,
		Limit:       limit,
	}

	return ma.unifiedService.GetMetrics(ctx, filters)
}

// GetActiveAlerts retrieves active alerts using the unified system
func (ma *MonitoringAdapter) GetActiveAlerts(ctx context.Context, component string) ([]*UnifiedAlert, error) {
	filters := &AlertFilters{
		Component: component,
		Status:    AlertStatusActive,
		Limit:     100,
	}

	return ma.unifiedService.GetAlerts(ctx, filters)
}

// UpdateAlertStatus updates alert status using the unified system
func (ma *MonitoringAdapter) UpdateAlertStatus(ctx context.Context, alertID uuid.UUID, status AlertStatus, userID *uuid.UUID) error {
	return ma.unifiedService.UpdateAlertStatus(ctx, alertID, status, userID)
}

// GetUnifiedService returns the underlying unified monitoring service
func (ma *MonitoringAdapter) GetUnifiedService() *UnifiedMonitoringService {
	return ma.unifiedService
}

// Helper function to check if a string contains any of the given substrings
func contains(str string, substrings ...string) bool {
	for _, substring := range substrings {
		if len(str) >= len(substring) {
			for i := 0; i <= len(str)-len(substring); i++ {
				if str[i:i+len(substring)] == substring {
					return true
				}
			}
		}
	}
	return false
}
