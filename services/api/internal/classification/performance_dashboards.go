package classification

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// PerformanceDashboardsUnified provides comprehensive database performance monitoring and dashboard functionality
// using the unified monitoring schema instead of the old redundant tables
type PerformanceDashboardsUnified struct {
	db *sql.DB
}

// NewPerformanceDashboardsUnified creates a new instance of PerformanceDashboardsUnified
func NewPerformanceDashboardsUnified(db *sql.DB) *PerformanceDashboardsUnified {
	return &PerformanceDashboardsUnified{
		db: db,
	}
}

// PerformanceMetric represents a performance metric from unified_performance_metrics
type PerformanceMetric struct {
	ID                string          `json:"id"`
	Timestamp         time.Time       `json:"timestamp"`
	Component         string          `json:"component"`
	ComponentInstance string          `json:"component_instance"`
	ServiceName       string          `json:"service_name"`
	MetricType        string          `json:"metric_type"`
	MetricCategory    string          `json:"metric_category"`
	MetricName        string          `json:"metric_name"`
	MetricValue       float64         `json:"metric_value"`
	MetricUnit        string          `json:"metric_unit"`
	Tags              json.RawMessage `json:"tags"`
	Metadata          json.RawMessage `json:"metadata"`
	RequestID         string          `json:"request_id"`
	OperationID       string          `json:"operation_id"`
	UserID            string          `json:"user_id"`
	ConfidenceScore   float64         `json:"confidence_score"`
	DataSource        string          `json:"data_source"`
	CreatedAt         time.Time       `json:"created_at"`
}

// PerformanceAlert represents a performance alert from unified_performance_alerts
type PerformanceAlert struct {
	ID                string          `json:"id"`
	Timestamp         time.Time       `json:"timestamp"`
	Component         string          `json:"component"`
	ComponentInstance string          `json:"component_instance"`
	ServiceName       string          `json:"service_name"`
	AlertType         string          `json:"alert_type"`
	Severity          string          `json:"severity"`
	Status            string          `json:"status"`
	Title             string          `json:"title"`
	Description       string          `json:"description"`
	MetricName        string          `json:"metric_name"`
	MetricValue       float64         `json:"metric_value"`
	ThresholdValue    float64         `json:"threshold_value"`
	Tags              json.RawMessage `json:"tags"`
	Metadata          json.RawMessage `json:"metadata"`
	RequestID         string          `json:"request_id"`
	OperationID       string          `json:"operation_id"`
	UserID            string          `json:"user_id"`
	ResolvedAt        *time.Time      `json:"resolved_at"`
	CreatedAt         time.Time       `json:"created_at"`
}

// PerformanceReport represents a performance report from unified_performance_reports
type PerformanceReport struct {
	ID                string          `json:"id"`
	Timestamp         time.Time       `json:"timestamp"`
	Component         string          `json:"component"`
	ComponentInstance string          `json:"component_instance"`
	ServiceName       string          `json:"service_name"`
	ReportType        string          `json:"report_type"`
	ReportCategory    string          `json:"report_category"`
	ReportName        string          `json:"report_name"`
	ReportData        json.RawMessage `json:"report_data"`
	Tags              json.RawMessage `json:"tags"`
	Metadata          json.RawMessage `json:"metadata"`
	RequestID         string          `json:"request_id"`
	OperationID       string          `json:"operation_id"`
	UserID            string          `json:"user_id"`
	CreatedAt         time.Time       `json:"created_at"`
}

// IntegrationHealth represents integration health from performance_integration_health
type IntegrationHealth struct {
	ID                string          `json:"id"`
	Timestamp         time.Time       `json:"timestamp"`
	Component         string          `json:"component"`
	ComponentInstance string          `json:"component_instance"`
	ServiceName       string          `json:"service_name"`
	IntegrationType   string          `json:"integration_type"`
	HealthStatus      string          `json:"health_status"`
	HealthScore       float64         `json:"health_score"`
	ResponseTime      float64         `json:"response_time"`
	ErrorRate         float64         `json:"error_rate"`
	LastSuccess       *time.Time      `json:"last_success"`
	LastFailure       *time.Time      `json:"last_failure"`
	Tags              json.RawMessage `json:"tags"`
	Metadata          json.RawMessage `json:"metadata"`
	CreatedAt         time.Time       `json:"created_at"`
}

// QueryPerformanceAnalysis represents query performance analysis derived from unified metrics
type QueryPerformanceAnalysis struct {
	QueryID              int64   `json:"query_id"`
	QueryText            string  `json:"query_text"`
	Calls                int64   `json:"calls"`
	TotalTime            float64 `json:"total_time"`
	MeanTime             float64 `json:"mean_time"`
	MinTime              float64 `json:"min_time"`
	MaxTime              float64 `json:"max_time"`
	StddevTime           float64 `json:"stddev_time"`
	Rows                 int64   `json:"rows"`
	PerformanceCategory  string  `json:"performance_category"`
	OptimizationPriority string  `json:"optimization_priority"`
	Recommendations      string  `json:"recommendations"`
}

// IndexPerformanceAnalysis represents index performance analysis derived from unified metrics
type IndexPerformanceAnalysis struct {
	TableName            string  `json:"table_name"`
	IndexName            string  `json:"index_name"`
	IndexSizeMB          float64 `json:"index_size_mb"`
	IndexScans           int64   `json:"index_scans"`
	IndexTuplesRead      int64   `json:"index_tuples_read"`
	IndexTuplesFetched   int64   `json:"index_tuples_fetched"`
	EfficiencyScore      float64 `json:"efficiency_score"`
	UsageCategory        string  `json:"usage_category"`
	OptimizationPriority string  `json:"optimization_priority"`
	Recommendations      string  `json:"recommendations"`
}

// TablePerformanceAnalysis represents table performance analysis derived from unified metrics
type TablePerformanceAnalysis struct {
	TableName            string  `json:"table_name"`
	TableSizeMB          float64 `json:"table_size_mb"`
	RowCount             int64   `json:"row_count"`
	DeadTuples           int64   `json:"dead_tuples"`
	LiveTuples           int64   `json:"live_tuples"`
	BloatPercentage      float64 `json:"bloat_percentage"`
	SeqScans             int64   `json:"seq_scans"`
	SeqTuplesRead        int64   `json:"seq_tuples_read"`
	IdxScans             int64   `json:"idx_scans"`
	IdxTuplesFetched     int64   `json:"idx_tuples_fetched"`
	PerformanceScore     float64 `json:"performance_score"`
	OptimizationPriority string  `json:"optimization_priority"`
	Recommendations      string  `json:"recommendations"`
}

// ConnectionPerformanceAnalysis represents connection performance analysis derived from unified metrics
type ConnectionPerformanceAnalysis struct {
	ConnectionCount       int     `json:"connection_count"`
	ActiveConnections     int     `json:"active_connections"`
	IdleConnections       int     `json:"idle_connections"`
	MaxConnections        int     `json:"max_connections"`
	UtilizationPercentage float64 `json:"utilization_percentage"`
	Status                string  `json:"status"`
	Recommendations       string  `json:"recommendations"`
}

// PerformanceDashboard represents performance dashboard data derived from unified metrics
type PerformanceDashboard struct {
	DashboardSection string `json:"dashboard_section"`
	MetricName       string `json:"metric_name"`
	CurrentValue     string `json:"current_value"`
	TargetValue      string `json:"target_value"`
	Status           string `json:"status"`
	Trend            string `json:"trend"`
	Recommendations  string `json:"recommendations"`
}

// PerformanceTrend represents performance trend data derived from unified metrics
type PerformanceTrend struct {
	MetricName          string  `json:"metric_name"`
	DateRecorded        string  `json:"date_recorded"`
	AvgValue            float64 `json:"avg_value"`
	MaxValue            float64 `json:"max_value"`
	MinValue            float64 `json:"min_value"`
	TrendDirection      string  `json:"trend_direction"`
	PerformanceCategory string  `json:"performance_category"`
}

// PerformanceSummary represents performance summary data derived from unified metrics
type PerformanceSummary struct {
	SummaryCategory string    `json:"summary_category"`
	TotalMetrics    int       `json:"total_metrics"`
	OkCount         int       `json:"ok_count"`
	WarningCount    int       `json:"warning_count"`
	CriticalCount   int       `json:"critical_count"`
	OverallStatus   string    `json:"overall_status"`
	LastUpdated     time.Time `json:"last_updated"`
}

// PerformanceDataExport represents exported performance data derived from unified metrics
type PerformanceDataExport struct {
	ExportDate    string  `json:"export_date"`
	MetricName    string  `json:"metric_name"`
	DailyAvgValue float64 `json:"daily_avg_value"`
	DailyMaxValue float64 `json:"daily_max_value"`
	DailyMinValue float64 `json:"daily_min_value"`
	StatusSummary string  `json:"status_summary"`
}

// PerformanceValidation represents performance monitoring setup validation
type PerformanceValidation struct {
	Component      string `json:"component"`
	Status         string `json:"status"`
	Details        string `json:"details"`
	Recommendation string `json:"recommendation"`
}

// CollectPerformanceMetrics collects comprehensive performance metrics from unified_performance_metrics
func (pd *PerformanceDashboardsUnified) CollectPerformanceMetrics(ctx context.Context) ([]PerformanceMetric, error) {
	query := `
		SELECT 
			id, timestamp, component, component_instance, service_name,
			metric_type, metric_category, metric_name, metric_value, metric_unit,
			tags, metadata, request_id, operation_id, user_id,
			confidence_score, data_source, created_at
		FROM unified_performance_metrics 
		WHERE timestamp >= NOW() - INTERVAL '1 hour'
		ORDER BY timestamp DESC
		LIMIT 1000
	`

	rows, err := pd.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to collect performance metrics: %w", err)
	}
	defer rows.Close()

	var results []PerformanceMetric
	for rows.Next() {
		var result PerformanceMetric
		err := rows.Scan(
			&result.ID,
			&result.Timestamp,
			&result.Component,
			&result.ComponentInstance,
			&result.ServiceName,
			&result.MetricType,
			&result.MetricCategory,
			&result.MetricName,
			&result.MetricValue,
			&result.MetricUnit,
			&result.Tags,
			&result.Metadata,
			&result.RequestID,
			&result.OperationID,
			&result.UserID,
			&result.ConfidenceScore,
			&result.DataSource,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan performance metric: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetPerformanceAlerts gets current performance alerts from unified_performance_alerts
func (pd *PerformanceDashboardsUnified) GetPerformanceAlerts(ctx context.Context) ([]PerformanceAlert, error) {
	query := `
		SELECT 
			id, timestamp, component, component_instance, service_name,
			alert_type, severity, status, title, description,
			metric_name, metric_value, threshold_value,
			tags, metadata, request_id, operation_id, user_id,
			resolved_at, created_at
		FROM unified_performance_alerts 
		WHERE status = 'active' OR status = 'warning'
		ORDER BY timestamp DESC
		LIMIT 100
	`

	rows, err := pd.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance alerts: %w", err)
	}
	defer rows.Close()

	var results []PerformanceAlert
	for rows.Next() {
		var result PerformanceAlert
		err := rows.Scan(
			&result.ID,
			&result.Timestamp,
			&result.Component,
			&result.ComponentInstance,
			&result.ServiceName,
			&result.AlertType,
			&result.Severity,
			&result.Status,
			&result.Title,
			&result.Description,
			&result.MetricName,
			&result.MetricValue,
			&result.ThresholdValue,
			&result.Tags,
			&result.Metadata,
			&result.RequestID,
			&result.OperationID,
			&result.UserID,
			&result.ResolvedAt,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan performance alert: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetPerformanceSummary gets performance summary derived from unified metrics
func (pd *PerformanceDashboardsUnified) GetPerformanceSummary(ctx context.Context) ([]PerformanceSummary, error) {
	query := `
		SELECT 
			component as summary_category,
			COUNT(*) as total_metrics,
			COUNT(CASE WHEN metric_value <= 100 THEN 1 END) as ok_count,
			COUNT(CASE WHEN metric_value > 100 AND metric_value <= 500 THEN 1 END) as warning_count,
			COUNT(CASE WHEN metric_value > 500 THEN 1 END) as critical_count,
			CASE 
				WHEN COUNT(CASE WHEN metric_value > 500 THEN 1 END) > 0 THEN 'CRITICAL'
				WHEN COUNT(CASE WHEN metric_value > 100 AND metric_value <= 500 THEN 1 END) > 0 THEN 'WARNING'
				ELSE 'OK'
			END as overall_status,
			MAX(timestamp) as last_updated
		FROM unified_performance_metrics 
		WHERE timestamp >= NOW() - INTERVAL '1 hour'
		GROUP BY component
		ORDER BY component
	`

	rows, err := pd.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance summary: %w", err)
	}
	defer rows.Close()

	var results []PerformanceSummary
	for rows.Next() {
		var result PerformanceSummary
		err := rows.Scan(
			&result.SummaryCategory,
			&result.TotalMetrics,
			&result.OkCount,
			&result.WarningCount,
			&result.CriticalCount,
			&result.OverallStatus,
			&result.LastUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan performance summary: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetPerformanceTrends gets performance trends over time from unified metrics
func (pd *PerformanceDashboardsUnified) GetPerformanceTrends(ctx context.Context, daysBack int) ([]PerformanceTrend, error) {
	query := `
		SELECT 
			metric_name,
			DATE(timestamp) as date_recorded,
			AVG(metric_value) as avg_value,
			MAX(metric_value) as max_value,
			MIN(metric_value) as min_value,
			CASE 
				WHEN AVG(metric_value) > LAG(AVG(metric_value)) OVER (PARTITION BY metric_name ORDER BY DATE(timestamp)) THEN 'increasing'
				WHEN AVG(metric_value) < LAG(AVG(metric_value)) OVER (PARTITION BY metric_name ORDER BY DATE(timestamp)) THEN 'decreasing'
				ELSE 'stable'
			END as trend_direction,
			metric_category as performance_category
		FROM unified_performance_metrics 
		WHERE timestamp >= NOW() - INTERVAL '%d days'
		GROUP BY metric_name, DATE(timestamp), metric_category
		ORDER BY date_recorded DESC, metric_name
	`

	rows, err := pd.db.QueryContext(ctx, fmt.Sprintf(query, daysBack))
	if err != nil {
		return nil, fmt.Errorf("failed to get performance trends: %w", err)
	}
	defer rows.Close()

	var results []PerformanceTrend
	for rows.Next() {
		var result PerformanceTrend
		var dateRecorded time.Time
		err := rows.Scan(
			&result.MetricName,
			&dateRecorded,
			&result.AvgValue,
			&result.MaxValue,
			&result.MinValue,
			&result.TrendDirection,
			&result.PerformanceCategory,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan performance trend: %w", err)
		}
		result.DateRecorded = dateRecorded.Format("2006-01-02")
		results = append(results, result)
	}

	return results, nil
}

// GetCurrentPerformanceStatus gets current performance status summary from unified metrics
func (pd *PerformanceDashboardsUnified) GetCurrentPerformanceStatus(ctx context.Context) (map[string]interface{}, error) {
	status := make(map[string]interface{})

	// Get performance metrics
	metrics, err := pd.CollectPerformanceMetrics(ctx)
	if err != nil {
		log.Printf("Warning: failed to get performance metrics: %v", err)
	} else {
		status["performance_metrics"] = metrics
	}

	// Get performance summary
	summary, err := pd.GetPerformanceSummary(ctx)
	if err != nil {
		log.Printf("Warning: failed to get performance summary: %v", err)
	} else {
		status["performance_summary"] = summary
	}

	// Get performance alerts
	alerts, err := pd.GetPerformanceAlerts(ctx)
	if err != nil {
		log.Printf("Warning: failed to get performance alerts: %v", err)
	} else {
		status["performance_alerts"] = alerts
	}

	// Determine overall status
	overallStatus := "OK"
	for _, metric := range metrics {
		if metric.MetricValue > 500 {
			overallStatus = "CRITICAL"
			break
		} else if metric.MetricValue > 100 && overallStatus != "CRITICAL" {
			overallStatus = "WARNING"
		}
	}

	status["overall_status"] = overallStatus
	status["last_checked"] = time.Now()

	return status, nil
}

// GetPerformanceInsights gets performance insights and recommendations from unified metrics
func (pd *PerformanceDashboardsUnified) GetPerformanceInsights(ctx context.Context) (map[string]interface{}, error) {
	insights := make(map[string]interface{})

	// Get performance trends
	trends, err := pd.GetPerformanceTrends(ctx, 7)
	if err != nil {
		log.Printf("Warning: failed to get performance trends: %v", err)
	} else {
		insights["performance_trends"] = trends
	}

	// Get component performance breakdown
	componentQuery := `
		SELECT 
			component,
			metric_category,
			AVG(metric_value) as avg_value,
			MAX(metric_value) as max_value,
			MIN(metric_value) as min_value,
			COUNT(*) as metric_count
		FROM unified_performance_metrics 
		WHERE timestamp >= NOW() - INTERVAL '1 hour'
		GROUP BY component, metric_category
		ORDER BY component, metric_category
	`

	rows, err := pd.db.QueryContext(ctx, componentQuery)
	if err != nil {
		log.Printf("Warning: failed to get component performance: %v", err)
	} else {
		defer rows.Close()

		var componentData []map[string]interface{}
		for rows.Next() {
			var component, category string
			var avgValue, maxValue, minValue float64
			var count int64

			err := rows.Scan(&component, &category, &avgValue, &maxValue, &minValue, &count)
			if err != nil {
				log.Printf("Warning: failed to scan component performance: %v", err)
				continue
			}

			componentData = append(componentData, map[string]interface{}{
				"component":    component,
				"category":     category,
				"avg_value":    avgValue,
				"max_value":    maxValue,
				"min_value":    minValue,
				"metric_count": count,
			})
		}
		insights["component_performance"] = componentData
	}

	insights["last_updated"] = time.Now()

	return insights, nil
}

// ExportPerformanceData exports performance data for analysis from unified metrics
func (pd *PerformanceDashboardsUnified) ExportPerformanceData(ctx context.Context, daysBack int) ([]PerformanceDataExport, error) {
	query := `
		SELECT 
			DATE(timestamp) as export_date,
			metric_name,
			AVG(metric_value) as daily_avg_value,
			MAX(metric_value) as daily_max_value,
			MIN(metric_value) as daily_min_value,
			CASE 
				WHEN AVG(metric_value) > 500 THEN 'CRITICAL'
				WHEN AVG(metric_value) > 100 THEN 'WARNING'
				ELSE 'OK'
			END as status_summary
		FROM unified_performance_metrics 
		WHERE timestamp >= NOW() - INTERVAL '%d days'
		GROUP BY DATE(timestamp), metric_name
		ORDER BY export_date DESC, metric_name
	`

	rows, err := pd.db.QueryContext(ctx, fmt.Sprintf(query, daysBack))
	if err != nil {
		return nil, fmt.Errorf("failed to export performance data: %w", err)
	}
	defer rows.Close()

	var results []PerformanceDataExport
	for rows.Next() {
		var result PerformanceDataExport
		var exportDate time.Time
		err := rows.Scan(
			&exportDate,
			&result.MetricName,
			&result.DailyAvgValue,
			&result.DailyMaxValue,
			&result.DailyMinValue,
			&result.StatusSummary,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan performance data export: %w", err)
		}
		result.ExportDate = exportDate.Format("2006-01-02")
		results = append(results, result)
	}

	return results, nil
}

// ValidatePerformanceMonitoringSetup validates performance monitoring setup using unified tables
func (pd *PerformanceDashboardsUnified) ValidatePerformanceMonitoringSetup(ctx context.Context) ([]PerformanceValidation, error) {
	var results []PerformanceValidation

	// Check if unified tables exist and have data
	validationQueries := []struct {
		component string
		query     string
		details   string
	}{
		{
			component: "unified_performance_metrics",
			query:     "SELECT COUNT(*) FROM unified_performance_metrics WHERE timestamp >= NOW() - INTERVAL '1 hour'",
			details:   "Check if unified_performance_metrics table has recent data",
		},
		{
			component: "unified_performance_alerts",
			query:     "SELECT COUNT(*) FROM unified_performance_alerts WHERE timestamp >= NOW() - INTERVAL '1 hour'",
			details:   "Check if unified_performance_alerts table has recent data",
		},
		{
			component: "unified_performance_reports",
			query:     "SELECT COUNT(*) FROM unified_performance_reports WHERE timestamp >= NOW() - INTERVAL '1 hour'",
			details:   "Check if unified_performance_reports table has recent data",
		},
		{
			component: "performance_integration_health",
			query:     "SELECT COUNT(*) FROM performance_integration_health WHERE timestamp >= NOW() - INTERVAL '1 hour'",
			details:   "Check if performance_integration_health table has recent data",
		},
	}

	for _, vq := range validationQueries {
		var count int64
		err := pd.db.QueryRowContext(ctx, vq.query).Scan(&count)

		validation := PerformanceValidation{
			Component: vq.component,
			Details:   vq.details,
		}

		if err != nil {
			validation.Status = "ERROR"
			validation.Recommendation = fmt.Sprintf("Failed to query %s: %v", vq.component, err)
		} else if count == 0 {
			validation.Status = "WARNING"
			validation.Recommendation = fmt.Sprintf("%s table exists but has no recent data", vq.component)
		} else {
			validation.Status = "OK"
			validation.Recommendation = fmt.Sprintf("%s table is healthy with %d recent records", vq.component, count)
		}

		results = append(results, validation)
	}

	return results, nil
}

// MonitorPerformanceContinuously starts continuous performance monitoring using unified tables
func (pd *PerformanceDashboardsUnified) MonitorPerformanceContinuously(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting continuous performance monitoring with interval: %v", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping continuous performance monitoring")
			return
		case <-ticker.C:
			// Perform monitoring checks
			status, err := pd.GetCurrentPerformanceStatus(ctx)
			if err != nil {
				log.Printf("Error getting performance status: %v", err)
			} else {
				log.Printf("Performance monitoring completed successfully. Overall status: %v", status["overall_status"])
			}
		}
	}
}
