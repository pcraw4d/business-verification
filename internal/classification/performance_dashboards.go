package classification

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// PerformanceDashboards provides comprehensive database performance monitoring and dashboard functionality
type PerformanceDashboards struct {
	db *sql.DB
}

// NewPerformanceDashboards creates a new instance of PerformanceDashboards
func NewPerformanceDashboards(db *sql.DB) *PerformanceDashboards {
	return &PerformanceDashboards{
		db: db,
	}
}

// PerformanceMetric represents a performance metric
type PerformanceMetric struct {
	MetricName      string          `json:"metric_name"`
	MetricValue     float64         `json:"metric_value"`
	MetricUnit      string          `json:"metric_unit"`
	MetricCategory  string          `json:"metric_category"`
	Status          string          `json:"status"`
	Details         json.RawMessage `json:"details"`
	Recommendations string          `json:"recommendations"`
}

// QueryPerformanceAnalysis represents query performance analysis
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

// IndexPerformanceAnalysis represents index performance analysis
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

// TablePerformanceAnalysis represents table performance analysis
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

// ConnectionPerformanceAnalysis represents connection performance analysis
type ConnectionPerformanceAnalysis struct {
	ConnectionCount       int     `json:"connection_count"`
	ActiveConnections     int     `json:"active_connections"`
	IdleConnections       int     `json:"idle_connections"`
	MaxConnections        int     `json:"max_connections"`
	UtilizationPercentage float64 `json:"utilization_percentage"`
	Status                string  `json:"status"`
	Recommendations       string  `json:"recommendations"`
}

// PerformanceDashboard represents performance dashboard data
type PerformanceDashboard struct {
	DashboardSection string `json:"dashboard_section"`
	MetricName       string `json:"metric_name"`
	CurrentValue     string `json:"current_value"`
	TargetValue      string `json:"target_value"`
	Status           string `json:"status"`
	Trend            string `json:"trend"`
	Recommendations  string `json:"recommendations"`
}

// PerformanceTrend represents performance trend data
type PerformanceTrend struct {
	MetricName          string  `json:"metric_name"`
	DateRecorded        string  `json:"date_recorded"`
	AvgValue            float64 `json:"avg_value"`
	MaxValue            float64 `json:"max_value"`
	MinValue            float64 `json:"min_value"`
	TrendDirection      string  `json:"trend_direction"`
	PerformanceCategory string  `json:"performance_category"`
}

// PerformanceSummary represents performance summary data
type PerformanceSummary struct {
	SummaryCategory string    `json:"summary_category"`
	TotalMetrics    int       `json:"total_metrics"`
	OkCount         int       `json:"ok_count"`
	WarningCount    int       `json:"warning_count"`
	CriticalCount   int       `json:"critical_count"`
	OverallStatus   string    `json:"overall_status"`
	LastUpdated     time.Time `json:"last_updated"`
}

// PerformanceDataExport represents exported performance data
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

// CollectPerformanceMetrics collects comprehensive performance metrics
func (pd *PerformanceDashboards) CollectPerformanceMetrics(ctx context.Context) ([]PerformanceMetric, error) {
	query := `SELECT * FROM collect_performance_metrics()`

	rows, err := pd.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to collect performance metrics: %w", err)
	}
	defer rows.Close()

	var results []PerformanceMetric
	for rows.Next() {
		var result PerformanceMetric
		err := rows.Scan(
			&result.MetricName,
			&result.MetricValue,
			&result.MetricUnit,
			&result.MetricCategory,
			&result.Status,
			&result.Details,
			&result.Recommendations,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan performance metric: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetQueryPerformanceAnalysis gets detailed query performance analysis
func (pd *PerformanceDashboards) GetQueryPerformanceAnalysis(ctx context.Context) ([]QueryPerformanceAnalysis, error) {
	query := `SELECT * FROM get_query_performance_analysis()`

	rows, err := pd.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get query performance analysis: %w", err)
	}
	defer rows.Close()

	var results []QueryPerformanceAnalysis
	for rows.Next() {
		var result QueryPerformanceAnalysis
		err := rows.Scan(
			&result.QueryID,
			&result.QueryText,
			&result.Calls,
			&result.TotalTime,
			&result.MeanTime,
			&result.MinTime,
			&result.MaxTime,
			&result.StddevTime,
			&result.Rows,
			&result.PerformanceCategory,
			&result.OptimizationPriority,
			&result.Recommendations,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan query performance analysis: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetIndexPerformanceAnalysis gets index performance analysis
func (pd *PerformanceDashboards) GetIndexPerformanceAnalysis(ctx context.Context) ([]IndexPerformanceAnalysis, error) {
	query := `SELECT * FROM get_index_performance_analysis()`

	rows, err := pd.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get index performance analysis: %w", err)
	}
	defer rows.Close()

	var results []IndexPerformanceAnalysis
	for rows.Next() {
		var result IndexPerformanceAnalysis
		err := rows.Scan(
			&result.TableName,
			&result.IndexName,
			&result.IndexSizeMB,
			&result.IndexScans,
			&result.IndexTuplesRead,
			&result.IndexTuplesFetched,
			&result.EfficiencyScore,
			&result.UsageCategory,
			&result.OptimizationPriority,
			&result.Recommendations,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan index performance analysis: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetTablePerformanceAnalysis gets table performance analysis
func (pd *PerformanceDashboards) GetTablePerformanceAnalysis(ctx context.Context) ([]TablePerformanceAnalysis, error) {
	query := `SELECT * FROM get_table_performance_analysis()`

	rows, err := pd.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get table performance analysis: %w", err)
	}
	defer rows.Close()

	var results []TablePerformanceAnalysis
	for rows.Next() {
		var result TablePerformanceAnalysis
		err := rows.Scan(
			&result.TableName,
			&result.TableSizeMB,
			&result.RowCount,
			&result.DeadTuples,
			&result.LiveTuples,
			&result.BloatPercentage,
			&result.SeqScans,
			&result.SeqTuplesRead,
			&result.IdxScans,
			&result.IdxTuplesFetched,
			&result.PerformanceScore,
			&result.OptimizationPriority,
			&result.Recommendations,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table performance analysis: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetConnectionPerformanceAnalysis gets connection performance analysis
func (pd *PerformanceDashboards) GetConnectionPerformanceAnalysis(ctx context.Context) (*ConnectionPerformanceAnalysis, error) {
	query := `SELECT * FROM get_connection_performance_analysis()`

	var result ConnectionPerformanceAnalysis
	err := pd.db.QueryRowContext(ctx, query).Scan(
		&result.ConnectionCount,
		&result.ActiveConnections,
		&result.IdleConnections,
		&result.MaxConnections,
		&result.UtilizationPercentage,
		&result.Status,
		&result.Recommendations,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get connection performance analysis: %w", err)
	}

	return &result, nil
}

// GeneratePerformanceDashboard generates performance dashboard data
func (pd *PerformanceDashboards) GeneratePerformanceDashboard(ctx context.Context) ([]PerformanceDashboard, error) {
	query := `SELECT * FROM generate_performance_dashboard()`

	rows, err := pd.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate performance dashboard: %w", err)
	}
	defer rows.Close()

	var results []PerformanceDashboard
	for rows.Next() {
		var result PerformanceDashboard
		err := rows.Scan(
			&result.DashboardSection,
			&result.MetricName,
			&result.CurrentValue,
			&result.TargetValue,
			&result.Status,
			&result.Trend,
			&result.Recommendations,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan performance dashboard: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// LogPerformanceMetrics logs current performance metrics
func (pd *PerformanceDashboards) LogPerformanceMetrics(ctx context.Context) error {
	query := `SELECT log_performance_metrics()`

	_, err := pd.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to log performance metrics: %w", err)
	}

	return nil
}

// GetPerformanceTrends gets performance trends over time
func (pd *PerformanceDashboards) GetPerformanceTrends(ctx context.Context, daysBack int) ([]PerformanceTrend, error) {
	query := `SELECT * FROM get_performance_trends($1) ORDER BY date_recorded DESC`

	rows, err := pd.db.QueryContext(ctx, query, daysBack)
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

// GetPerformanceAlerts gets current performance alerts
func (pd *PerformanceDashboards) GetPerformanceAlerts(ctx context.Context) ([]PerformanceAlert, error) {
	query := `SELECT * FROM get_performance_alerts()`

	rows, err := pd.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance alerts: %w", err)
	}
	defer rows.Close()

	var results []PerformanceAlert
	for rows.Next() {
		var result PerformanceAlert
		err := rows.Scan(
			&result.AlertID,
			&result.MetricName,
			&result.AlertLevel,
			&result.MetricValue,
			&result.ThresholdValue,
			&result.AlertMessage,
			&result.Recommendations,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan performance alert: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetPerformanceSummary gets performance summary
func (pd *PerformanceDashboards) GetPerformanceSummary(ctx context.Context) ([]PerformanceSummary, error) {
	query := `SELECT * FROM get_performance_summary()`

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

// SetupAutomatedPerformanceMonitoring sets up automated performance monitoring
func (pd *PerformanceDashboards) SetupAutomatedPerformanceMonitoring(ctx context.Context) (string, error) {
	query := `SELECT setup_automated_performance_monitoring()`

	var result string
	err := pd.db.QueryRowContext(ctx, query).Scan(&result)
	if err != nil {
		return "", fmt.Errorf("failed to setup automated performance monitoring: %w", err)
	}

	return result, nil
}

// GetPerformanceDashboardData gets performance dashboard data
func (pd *PerformanceDashboards) GetPerformanceDashboardData(ctx context.Context) ([]map[string]interface{}, error) {
	query := `SELECT * FROM get_performance_dashboard_data()`

	rows, err := pd.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance dashboard data: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var section, metric, value, status string
		var lastUpdated time.Time

		err := rows.Scan(
			&section,
			&metric,
			&value,
			&status,
			&lastUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan performance dashboard data: %w", err)
		}

		result := map[string]interface{}{
			"section":      section,
			"metric":       metric,
			"value":        value,
			"status":       status,
			"last_updated": lastUpdated,
		}

		results = append(results, result)
	}

	return results, nil
}

// ExportPerformanceData exports performance data for analysis
func (pd *PerformanceDashboards) ExportPerformanceData(ctx context.Context, daysBack int) ([]PerformanceDataExport, error) {
	query := `SELECT * FROM export_performance_data($1) ORDER BY export_date DESC`

	rows, err := pd.db.QueryContext(ctx, query, daysBack)
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

// ValidatePerformanceMonitoringSetup validates performance monitoring setup
func (pd *PerformanceDashboards) ValidatePerformanceMonitoringSetup(ctx context.Context) ([]PerformanceValidation, error) {
	query := `SELECT * FROM validate_performance_monitoring_setup()`

	rows, err := pd.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to validate performance monitoring setup: %w", err)
	}
	defer rows.Close()

	var results []PerformanceValidation
	for rows.Next() {
		var result PerformanceValidation
		err := rows.Scan(
			&result.Component,
			&result.Status,
			&result.Details,
			&result.Recommendation,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan performance validation: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// RunAutomatedPerformanceMonitoring runs automated performance monitoring
func (pd *PerformanceDashboards) RunAutomatedPerformanceMonitoring(ctx context.Context) error {
	query := `SELECT automated_performance_monitoring()`

	_, err := pd.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to run automated performance monitoring: %w", err)
	}

	return nil
}

// GetCurrentPerformanceStatus gets current performance status summary
func (pd *PerformanceDashboards) GetCurrentPerformanceStatus(ctx context.Context) (map[string]interface{}, error) {
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
		if metric.Status == "CRITICAL" {
			overallStatus = "CRITICAL"
			break
		} else if metric.Status == "WARNING" && overallStatus != "CRITICAL" {
			overallStatus = "WARNING"
		}
	}

	status["overall_status"] = overallStatus
	status["last_checked"] = time.Now()

	return status, nil
}

// MonitorPerformanceContinuously starts continuous performance monitoring
func (pd *PerformanceDashboards) MonitorPerformanceContinuously(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting continuous performance monitoring with interval: %v", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping continuous performance monitoring")
			return
		case <-ticker.C:
			if err := pd.RunAutomatedPerformanceMonitoring(ctx); err != nil {
				log.Printf("Error running automated performance monitoring: %v", err)
			} else {
				log.Println("Automated performance monitoring completed successfully")
			}
		}
	}
}

// GetPerformanceInsights gets performance insights and recommendations
func (pd *PerformanceDashboards) GetPerformanceInsights(ctx context.Context) (map[string]interface{}, error) {
	insights := make(map[string]interface{})

	// Get query performance analysis
	queryAnalysis, err := pd.GetQueryPerformanceAnalysis(ctx)
	if err != nil {
		log.Printf("Warning: failed to get query performance analysis: %v", err)
	} else {
		insights["query_analysis"] = queryAnalysis
	}

	// Get index performance analysis
	indexAnalysis, err := pd.GetIndexPerformanceAnalysis(ctx)
	if err != nil {
		log.Printf("Warning: failed to get index performance analysis: %v", err)
	} else {
		insights["index_analysis"] = indexAnalysis
	}

	// Get table performance analysis
	tableAnalysis, err := pd.GetTablePerformanceAnalysis(ctx)
	if err != nil {
		log.Printf("Warning: failed to get table performance analysis: %v", err)
	} else {
		insights["table_analysis"] = tableAnalysis
	}

	// Get connection performance analysis
	connectionAnalysis, err := pd.GetConnectionPerformanceAnalysis(ctx)
	if err != nil {
		log.Printf("Warning: failed to get connection performance analysis: %v", err)
	} else {
		insights["connection_analysis"] = connectionAnalysis
	}

	// Get performance trends
	trends, err := pd.GetPerformanceTrends(ctx, 7)
	if err != nil {
		log.Printf("Warning: failed to get performance trends: %v", err)
	} else {
		insights["performance_trends"] = trends
	}

	insights["last_updated"] = time.Now()

	return insights, nil
}
