package classification

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// UsageMonitoring provides comprehensive monitoring for Supabase free tier usage and limits
type UsageMonitoring struct {
	db *sql.DB
}

// NewUsageMonitoring creates a new instance of UsageMonitoring
func NewUsageMonitoring(db *sql.DB) *UsageMonitoring {
	return &UsageMonitoring{
		db: db,
	}
}

// DatabaseStorageUsage represents database storage usage metrics
type DatabaseStorageUsage struct {
	DatabaseName    string  `json:"database_name"`
	TotalSizeMB     float64 `json:"total_size_mb"`
	LimitMB         float64 `json:"limit_mb"`
	UsagePercentage float64 `json:"usage_percentage"`
	Status          string  `json:"status"`
}

// TableSize represents table size information
type TableSize struct {
	TableName   string  `json:"table_name"`
	SizeMB      float64 `json:"size_mb"`
	RowCount    int64   `json:"row_count"`
	IndexSizeMB float64 `json:"index_size_mb"`
}

// ConnectionUsage represents connection usage metrics
type ConnectionUsage struct {
	CurrentConnections int     `json:"current_connections"`
	MaxConnections     int     `json:"max_connections"`
	UsagePercentage    float64 `json:"usage_percentage"`
	Status             string  `json:"status"`
}

// QueryPerformance represents query performance metrics
type QueryPerformance struct {
	QueryType          string  `json:"query_type"`
	AvgExecutionTimeMs float64 `json:"avg_execution_time_ms"`
	TotalExecutions    int64   `json:"total_executions"`
	SlowQueries        int64   `json:"slow_queries"`
	Status             string  `json:"status"`
}

// IndexUsage represents index usage information
type IndexUsage struct {
	TableName       string  `json:"table_name"`
	IndexName       string  `json:"index_name"`
	IndexSizeMB     float64 `json:"index_size_mb"`
	UsageCount      int64   `json:"usage_count"`
	EfficiencyScore float64 `json:"efficiency_score"`
	Status          string  `json:"status"`
}

// FreeTierLimit represents free tier limit information
type FreeTierLimit struct {
	LimitName       string  `json:"limit_name"`
	CurrentUsage    float64 `json:"current_usage"`
	LimitValue      float64 `json:"limit_value"`
	UsagePercentage float64 `json:"usage_percentage"`
	Status          string  `json:"status"`
	Description     string  `json:"description"`
}

// UsageReport represents a comprehensive usage report
type UsageReport struct {
	ReportSection   string  `json:"report_section"`
	MetricName      string  `json:"metric_name"`
	CurrentValue    string  `json:"current_value"`
	LimitValue      string  `json:"limit_value"`
	UsagePercentage float64 `json:"usage_percentage"`
	Status          string  `json:"status"`
	Recommendation  string  `json:"recommendation"`
}

// UsageTrend represents usage trend information
type UsageTrend struct {
	MetricName         string  `json:"metric_name"`
	DateRecorded       string  `json:"date_recorded"`
	AvgUsagePercentage float64 `json:"avg_usage_percentage"`
	MaxUsagePercentage float64 `json:"max_usage_percentage"`
	MinUsagePercentage float64 `json:"min_usage_percentage"`
	TrendDirection     string  `json:"trend_direction"`
}

// OptimizationOpportunity represents optimization opportunities
type OptimizationOpportunity struct {
	OptimizationType string `json:"optimization_type"`
	Description      string `json:"description"`
	PotentialSavings string `json:"potential_savings"`
	Priority         string `json:"priority"`
	ActionRequired   string `json:"action_required"`
}

// MonitoringDashboard represents monitoring dashboard data
type MonitoringDashboard struct {
	Section     string    `json:"section"`
	Metric      string    `json:"metric"`
	Value       string    `json:"value"`
	Status      string    `json:"status"`
	LastUpdated time.Time `json:"last_updated"`
}

// UsageDataExport represents exported usage data
type UsageDataExport struct {
	ExportDate    string  `json:"export_date"`
	MetricName    string  `json:"metric_name"`
	DailyAvgUsage float64 `json:"daily_avg_usage"`
	DailyMaxUsage float64 `json:"daily_max_usage"`
	DailyMinUsage float64 `json:"daily_min_usage"`
	StatusSummary string  `json:"status_summary"`
}

// MonitoringValidation represents monitoring setup validation
type MonitoringValidation struct {
	Component      string `json:"component"`
	Status         string `json:"status"`
	Details        string `json:"details"`
	Recommendation string `json:"recommendation"`
}

// CheckDatabaseStorageUsage checks database storage usage
func (um *UsageMonitoring) CheckDatabaseStorageUsage(ctx context.Context) (*DatabaseStorageUsage, error) {
	query := `SELECT * FROM check_database_storage_usage()`

	var result DatabaseStorageUsage
	err := um.db.QueryRowContext(ctx, query).Scan(
		&result.DatabaseName,
		&result.TotalSizeMB,
		&result.LimitMB,
		&result.UsagePercentage,
		&result.Status,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to check database storage usage: %w", err)
	}

	return &result, nil
}

// CheckTableSizes checks table sizes
func (um *UsageMonitoring) CheckTableSizes(ctx context.Context) ([]TableSize, error) {
	query := `SELECT * FROM check_table_sizes() ORDER BY size_mb DESC`

	rows, err := um.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check table sizes: %w", err)
	}
	defer rows.Close()

	var results []TableSize
	for rows.Next() {
		var result TableSize
		err := rows.Scan(
			&result.TableName,
			&result.SizeMB,
			&result.RowCount,
			&result.IndexSizeMB,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table size result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// CheckConnectionUsage checks connection usage
func (um *UsageMonitoring) CheckConnectionUsage(ctx context.Context) (*ConnectionUsage, error) {
	query := `SELECT * FROM check_connection_usage()`

	var result ConnectionUsage
	err := um.db.QueryRowContext(ctx, query).Scan(
		&result.CurrentConnections,
		&result.MaxConnections,
		&result.UsagePercentage,
		&result.Status,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to check connection usage: %w", err)
	}

	return &result, nil
}

// CheckQueryPerformance checks query performance
func (um *UsageMonitoring) CheckQueryPerformance(ctx context.Context) ([]QueryPerformance, error) {
	query := `SELECT * FROM check_query_performance()`

	rows, err := um.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check query performance: %w", err)
	}
	defer rows.Close()

	var results []QueryPerformance
	for rows.Next() {
		var result QueryPerformance
		err := rows.Scan(
			&result.QueryType,
			&result.AvgExecutionTimeMs,
			&result.TotalExecutions,
			&result.SlowQueries,
			&result.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan query performance result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// CheckIndexUsage checks index usage
func (um *UsageMonitoring) CheckIndexUsage(ctx context.Context) ([]IndexUsage, error) {
	query := `SELECT * FROM check_index_usage() ORDER BY index_size_mb DESC`

	rows, err := um.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check index usage: %w", err)
	}
	defer rows.Close()

	var results []IndexUsage
	for rows.Next() {
		var result IndexUsage
		err := rows.Scan(
			&result.TableName,
			&result.IndexName,
			&result.IndexSizeMB,
			&result.UsageCount,
			&result.EfficiencyScore,
			&result.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan index usage result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// CheckFreeTierLimits checks free tier limits
func (um *UsageMonitoring) CheckFreeTierLimits(ctx context.Context) ([]FreeTierLimit, error) {
	query := `SELECT * FROM check_free_tier_limits()`

	rows, err := um.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check free tier limits: %w", err)
	}
	defer rows.Close()

	var results []FreeTierLimit
	for rows.Next() {
		var result FreeTierLimit
		err := rows.Scan(
			&result.LimitName,
			&result.CurrentUsage,
			&result.LimitValue,
			&result.UsagePercentage,
			&result.Status,
			&result.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan free tier limit result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GenerateUsageReport generates a comprehensive usage report
func (um *UsageMonitoring) GenerateUsageReport(ctx context.Context) ([]UsageReport, error) {
	query := `SELECT * FROM generate_usage_report()`

	rows, err := um.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate usage report: %w", err)
	}
	defer rows.Close()

	var results []UsageReport
	for rows.Next() {
		var result UsageReport
		err := rows.Scan(
			&result.ReportSection,
			&result.MetricName,
			&result.CurrentValue,
			&result.LimitValue,
			&result.UsagePercentage,
			&result.Status,
			&result.Recommendation,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage report result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// LogUsageMetrics logs current usage metrics
func (um *UsageMonitoring) LogUsageMetrics(ctx context.Context) error {
	query := `SELECT log_usage_metrics()`

	_, err := um.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to log usage metrics: %w", err)
	}

	return nil
}

// GetUsageTrends gets usage trends over time
func (um *UsageMonitoring) GetUsageTrends(ctx context.Context, daysBack int) ([]UsageTrend, error) {
	query := `SELECT * FROM get_usage_trends($1) ORDER BY date_recorded DESC`

	rows, err := um.db.QueryContext(ctx, query, daysBack)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage trends: %w", err)
	}
	defer rows.Close()

	var results []UsageTrend
	for rows.Next() {
		var result UsageTrend
		var dateRecorded time.Time
		err := rows.Scan(
			&result.MetricName,
			&dateRecorded,
			&result.AvgUsagePercentage,
			&result.MaxUsagePercentage,
			&result.MinUsagePercentage,
			&result.TrendDirection,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage trend result: %w", err)
		}
		result.DateRecorded = dateRecorded.Format("2006-01-02")
		results = append(results, result)
	}

	return results, nil
}

// CheckOptimizationOpportunities checks for optimization opportunities
func (um *UsageMonitoring) CheckOptimizationOpportunities(ctx context.Context) ([]OptimizationOpportunity, error) {
	query := `SELECT * FROM check_optimization_opportunities()`

	rows, err := um.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check optimization opportunities: %w", err)
	}
	defer rows.Close()

	var results []OptimizationOpportunity
	for rows.Next() {
		var result OptimizationOpportunity
		err := rows.Scan(
			&result.OptimizationType,
			&result.Description,
			&result.PotentialSavings,
			&result.Priority,
			&result.ActionRequired,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan optimization opportunity result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// SetupAutomatedMonitoring sets up automated monitoring
func (um *UsageMonitoring) SetupAutomatedMonitoring(ctx context.Context) (string, error) {
	query := `SELECT setup_automated_monitoring()`

	var result string
	err := um.db.QueryRowContext(ctx, query).Scan(&result)
	if err != nil {
		return "", fmt.Errorf("failed to setup automated monitoring: %w", err)
	}

	return result, nil
}

// GetMonitoringDashboard gets monitoring dashboard data
func (um *UsageMonitoring) GetMonitoringDashboard(ctx context.Context) ([]MonitoringDashboard, error) {
	query := `SELECT * FROM get_monitoring_dashboard()`

	rows, err := um.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get monitoring dashboard: %w", err)
	}
	defer rows.Close()

	var results []MonitoringDashboard
	for rows.Next() {
		var result MonitoringDashboard
		err := rows.Scan(
			&result.Section,
			&result.Metric,
			&result.Value,
			&result.Status,
			&result.LastUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monitoring dashboard result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// ExportUsageData exports usage data for analysis
func (um *UsageMonitoring) ExportUsageData(ctx context.Context, daysBack int) ([]UsageDataExport, error) {
	query := `SELECT * FROM export_usage_data($1) ORDER BY export_date DESC`

	rows, err := um.db.QueryContext(ctx, query, daysBack)
	if err != nil {
		return nil, fmt.Errorf("failed to export usage data: %w", err)
	}
	defer rows.Close()

	var results []UsageDataExport
	for rows.Next() {
		var result UsageDataExport
		var exportDate time.Time
		err := rows.Scan(
			&exportDate,
			&result.MetricName,
			&result.DailyAvgUsage,
			&result.DailyMaxUsage,
			&result.DailyMinUsage,
			&result.StatusSummary,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage data export result: %w", err)
		}
		result.ExportDate = exportDate.Format("2006-01-02")
		results = append(results, result)
	}

	return results, nil
}

// ValidateMonitoringSetup validates monitoring setup
func (um *UsageMonitoring) ValidateMonitoringSetup(ctx context.Context) ([]MonitoringValidation, error) {
	query := `SELECT * FROM validate_monitoring_setup()`

	rows, err := um.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to validate monitoring setup: %w", err)
	}
	defer rows.Close()

	var results []MonitoringValidation
	for rows.Next() {
		var result MonitoringValidation
		err := rows.Scan(
			&result.Component,
			&result.Status,
			&result.Details,
			&result.Recommendation,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monitoring validation result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// RunAutomatedMonitoring runs automated monitoring
func (um *UsageMonitoring) RunAutomatedMonitoring(ctx context.Context) error {
	query := `SELECT automated_usage_monitoring()`

	_, err := um.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to run automated monitoring: %w", err)
	}

	return nil
}

// GetCurrentUsageStatus gets current usage status summary
func (um *UsageMonitoring) GetCurrentUsageStatus(ctx context.Context) (map[string]interface{}, error) {
	status := make(map[string]interface{})

	// Get database storage usage
	storageUsage, err := um.CheckDatabaseStorageUsage(ctx)
	if err != nil {
		log.Printf("Warning: failed to get storage usage: %v", err)
	} else {
		status["database_storage"] = storageUsage
	}

	// Get connection usage
	connectionUsage, err := um.CheckConnectionUsage(ctx)
	if err != nil {
		log.Printf("Warning: failed to get connection usage: %v", err)
	} else {
		status["connection_usage"] = connectionUsage
	}

	// Get free tier limits
	limits, err := um.CheckFreeTierLimits(ctx)
	if err != nil {
		log.Printf("Warning: failed to get free tier limits: %v", err)
	} else {
		status["free_tier_limits"] = limits
	}

	// Determine overall status
	overallStatus := "OK"
	for _, limit := range limits {
		if limit.Status == "CRITICAL" {
			overallStatus = "CRITICAL"
			break
		} else if limit.Status == "WARNING" && overallStatus != "CRITICAL" {
			overallStatus = "WARNING"
		}
	}

	status["overall_status"] = overallStatus
	status["last_checked"] = time.Now()

	return status, nil
}

// MonitorUsageContinuously starts continuous usage monitoring
func (um *UsageMonitoring) MonitorUsageContinuously(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting continuous usage monitoring with interval: %v", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping continuous usage monitoring")
			return
		case <-ticker.C:
			if err := um.RunAutomatedMonitoring(ctx); err != nil {
				log.Printf("Error running automated monitoring: %v", err)
			} else {
				log.Println("Automated monitoring completed successfully")
			}
		}
	}
}

// GetUsageAlerts gets current usage alerts
func (um *UsageMonitoring) GetUsageAlerts(ctx context.Context) ([]map[string]interface{}, error) {
	query := `
		SELECT metric_name, metric_value, metric_unit, usage_percentage, status, notes, recorded_at
		FROM usage_monitoring
		WHERE status IN ('WARNING', 'CRITICAL')
		AND recorded_at >= NOW() - INTERVAL '24 hours'
		ORDER BY recorded_at DESC
	`

	rows, err := um.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage alerts: %w", err)
	}
	defer rows.Close()

	var alerts []map[string]interface{}
	for rows.Next() {
		var metricName, metricUnit, status, notes string
		var metricValue, usagePercentage float64
		var recordedAt time.Time

		err := rows.Scan(
			&metricName,
			&metricValue,
			&metricUnit,
			&usagePercentage,
			&status,
			&notes,
			&recordedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage alert: %w", err)
		}

		alert := map[string]interface{}{
			"metric_name":      metricName,
			"metric_value":     metricValue,
			"metric_unit":      metricUnit,
			"usage_percentage": usagePercentage,
			"status":           status,
			"notes":            notes,
			"recorded_at":      recordedAt,
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}
