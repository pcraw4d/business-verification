package classification

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// QueryPerformanceMonitoring provides comprehensive query performance monitoring and optimization
type QueryPerformanceMonitoring struct {
	db *sql.DB
}

// NewQueryPerformanceMonitoring creates a new instance of QueryPerformanceMonitoring
func NewQueryPerformanceMonitoring(db *sql.DB) *QueryPerformanceMonitoring {
	return &QueryPerformanceMonitoring{
		db: db,
	}
}

// QueryPerformanceAnalysisResult represents query performance analysis results
type QueryPerformanceAnalysisResult struct {
	QueryID                int64    `json:"query_id"`
	PerformanceScore       float64  `json:"performance_score"`
	PerformanceCategory    string   `json:"performance_category"`
	OptimizationPriority   string   `json:"optimization_priority"`
	Recommendations        string   `json:"recommendations"`
	IndexSuggestions       []string `json:"index_suggestions"`
	QueryOptimizationHints []string `json:"query_optimization_hints"`
}

// QueryPerformanceStats represents query performance statistics
type QueryPerformanceStats struct {
	TotalQueries              int64           `json:"total_queries"`
	AvgExecutionTimeMs        float64         `json:"avg_execution_time_ms"`
	MaxExecutionTimeMs        float64         `json:"max_execution_time_ms"`
	MinExecutionTimeMs        float64         `json:"min_execution_time_ms"`
	TotalRowsReturned         int64           `json:"total_rows_returned"`
	TotalRowsExamined         int64           `json:"total_rows_examined"`
	AvgIndexUsageScore        float64         `json:"avg_index_usage_score"`
	AvgCacheHitRatio          float64         `json:"avg_cache_hit_ratio"`
	AvgComplexityScore        float64         `json:"avg_complexity_score"`
	PerformanceDistribution   json.RawMessage `json:"performance_distribution"`
	TopSlowQueries            json.RawMessage `json:"top_slow_queries"`
	OptimizationOpportunities json.RawMessage `json:"optimization_opportunities"`
}

// QueryPerformanceTrend represents query performance trend data
type QueryPerformanceTrend struct {
	HourBucket         time.Time `json:"hour_bucket"`
	TotalQueries       int64     `json:"total_queries"`
	AvgExecutionTimeMs float64   `json:"avg_execution_time_ms"`
	AvgIndexUsageScore float64   `json:"avg_index_usage_score"`
	AvgCacheHitRatio   float64   `json:"avg_cache_hit_ratio"`
	PerformanceScore   float64   `json:"performance_score"`
	SlowQueryCount     int64     `json:"slow_query_count"`
}

// QueryPerformanceAlert represents a query performance alert
type QueryPerformanceAlert struct {
	AlertID             int       `json:"alert_id"`
	AlertType           string    `json:"alert_type"`
	AlertLevel          string    `json:"alert_level"`
	AlertMessage        string    `json:"alert_message"`
	QueryID             int64     `json:"query_id"`
	QueryText           string    `json:"query_text"`
	ExecutionTimeMs     float64   `json:"execution_time_ms"`
	PerformanceCategory string    `json:"performance_category"`
	Recommendations     string    `json:"recommendations"`
	CreatedAt           time.Time `json:"created_at"`
}

// QueryPerformanceDashboard represents query performance dashboard data
type QueryPerformanceDashboard struct {
	MetricName      string `json:"metric_name"`
	CurrentValue    string `json:"current_value"`
	TargetValue     string `json:"target_value"`
	Status          string `json:"status"`
	Trend           string `json:"trend"`
	Recommendations string `json:"recommendations"`
}

// QueryPerformanceInsight represents query performance insights
type QueryPerformanceInsight struct {
	InsightType            string  `json:"insight_type"`
	InsightTitle           string  `json:"insight_title"`
	InsightDescription     string  `json:"insight_description"`
	InsightPriority        string  `json:"insight_priority"`
	InsightRecommendations string  `json:"insight_recommendations"`
	AffectedQueries        int64   `json:"affected_queries"`
	PotentialImprovement   float64 `json:"potential_improvement"`
}

// QueryPerformanceValidation represents query performance monitoring setup validation
type QueryPerformanceValidation struct {
	Component      string `json:"component"`
	Status         string `json:"status"`
	Details        string `json:"details"`
	Recommendation string `json:"recommendation"`
}

// AnalyzeQueryPerformance analyzes query performance and provides optimization recommendations
func (qpm *QueryPerformanceMonitoring) AnalyzeQueryPerformance(ctx context.Context, queryText string, executionTimeMs float64, rowsReturned, rowsExamined int64) (*QueryPerformanceAnalysisResult, error) {
	query := `SELECT * FROM analyze_query_performance($1, $2, $3, $4)`

	var result QueryPerformanceAnalysisResult
	var indexSuggestions, queryOptimizationHints []string

	err := qpm.db.QueryRowContext(ctx, query, queryText, executionTimeMs, rowsReturned, rowsExamined).Scan(
		&result.QueryID,
		&result.PerformanceScore,
		&result.PerformanceCategory,
		&result.OptimizationPriority,
		&result.Recommendations,
		&indexSuggestions,
		&queryOptimizationHints,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to analyze query performance: %w", err)
	}

	result.IndexSuggestions = indexSuggestions
	result.QueryOptimizationHints = queryOptimizationHints

	return &result, nil
}

// LogQueryPerformance logs query performance data
func (qpm *QueryPerformanceMonitoring) LogQueryPerformance(ctx context.Context, queryText string, executionTimeMs float64, rowsReturned, rowsExamined int64, userID, sessionID, requestID *string) (int64, error) {
	query := `SELECT log_query_performance($1, $2, $3, $4, $5, $6, $7)`

	var logID int64
	err := qpm.db.QueryRowContext(ctx, query, queryText, executionTimeMs, rowsReturned, rowsExamined, userID, sessionID, requestID).Scan(&logID)

	if err != nil {
		return 0, fmt.Errorf("failed to log query performance: %w", err)
	}

	return logID, nil
}

// GetQueryPerformanceStats gets query performance statistics
func (qpm *QueryPerformanceMonitoring) GetQueryPerformanceStats(ctx context.Context, hoursBack int) (*QueryPerformanceStats, error) {
	query := `SELECT * FROM get_query_performance_stats($1)`

	var result QueryPerformanceStats
	err := qpm.db.QueryRowContext(ctx, query, hoursBack).Scan(
		&result.TotalQueries,
		&result.AvgExecutionTimeMs,
		&result.MaxExecutionTimeMs,
		&result.MinExecutionTimeMs,
		&result.TotalRowsReturned,
		&result.TotalRowsExamined,
		&result.AvgIndexUsageScore,
		&result.AvgCacheHitRatio,
		&result.AvgComplexityScore,
		&result.PerformanceDistribution,
		&result.TopSlowQueries,
		&result.OptimizationOpportunities,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get query performance stats: %w", err)
	}

	return &result, nil
}

// GetQueryPerformanceTrends gets query performance trends
func (qpm *QueryPerformanceMonitoring) GetQueryPerformanceTrends(ctx context.Context, hoursBack int) ([]QueryPerformanceTrend, error) {
	query := `SELECT * FROM get_query_performance_trends($1) ORDER BY hour_bucket DESC`

	rows, err := qpm.db.QueryContext(ctx, query, hoursBack)
	if err != nil {
		return nil, fmt.Errorf("failed to get query performance trends: %w", err)
	}
	defer rows.Close()

	var results []QueryPerformanceTrend
	for rows.Next() {
		var result QueryPerformanceTrend
		err := rows.Scan(
			&result.HourBucket,
			&result.TotalQueries,
			&result.AvgExecutionTimeMs,
			&result.AvgIndexUsageScore,
			&result.AvgCacheHitRatio,
			&result.PerformanceScore,
			&result.SlowQueryCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan query performance trend: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetQueryPerformanceAlerts gets query performance alerts
func (qpm *QueryPerformanceMonitoring) GetQueryPerformanceAlerts(ctx context.Context, hoursBack int) ([]QueryPerformanceAlert, error) {
	query := `SELECT * FROM get_query_performance_alerts($1)`

	rows, err := qpm.db.QueryContext(ctx, query, hoursBack)
	if err != nil {
		return nil, fmt.Errorf("failed to get query performance alerts: %w", err)
	}
	defer rows.Close()

	var results []QueryPerformanceAlert
	for rows.Next() {
		var result QueryPerformanceAlert
		err := rows.Scan(
			&result.AlertID,
			&result.AlertType,
			&result.AlertLevel,
			&result.AlertMessage,
			&result.QueryID,
			&result.QueryText,
			&result.ExecutionTimeMs,
			&result.PerformanceCategory,
			&result.Recommendations,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan query performance alert: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetQueryPerformanceDashboard gets query performance dashboard data
func (qpm *QueryPerformanceMonitoring) GetQueryPerformanceDashboard(ctx context.Context) ([]QueryPerformanceDashboard, error) {
	query := `SELECT * FROM get_query_performance_dashboard()`

	rows, err := qpm.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get query performance dashboard: %w", err)
	}
	defer rows.Close()

	var results []QueryPerformanceDashboard
	for rows.Next() {
		var result QueryPerformanceDashboard
		err := rows.Scan(
			&result.MetricName,
			&result.CurrentValue,
			&result.TargetValue,
			&result.Status,
			&result.Trend,
			&result.Recommendations,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan query performance dashboard: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// CleanupQueryPerformanceLogs cleans up old query performance logs
func (qpm *QueryPerformanceMonitoring) CleanupQueryPerformanceLogs(ctx context.Context, daysToKeep int) (int, error) {
	query := `SELECT cleanup_query_performance_logs($1)`

	var deletedCount int
	err := qpm.db.QueryRowContext(ctx, query, daysToKeep).Scan(&deletedCount)

	if err != nil {
		return 0, fmt.Errorf("failed to cleanup query performance logs: %w", err)
	}

	return deletedCount, nil
}

// GetQueryPerformanceInsights gets query performance insights
func (qpm *QueryPerformanceMonitoring) GetQueryPerformanceInsights(ctx context.Context) ([]QueryPerformanceInsight, error) {
	query := `SELECT * FROM get_query_performance_insights()`

	rows, err := qpm.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get query performance insights: %w", err)
	}
	defer rows.Close()

	var results []QueryPerformanceInsight
	for rows.Next() {
		var result QueryPerformanceInsight
		err := rows.Scan(
			&result.InsightType,
			&result.InsightTitle,
			&result.InsightDescription,
			&result.InsightPriority,
			&result.InsightRecommendations,
			&result.AffectedQueries,
			&result.PotentialImprovement,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan query performance insight: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// ValidateQueryPerformanceMonitoringSetup validates query performance monitoring setup
func (qpm *QueryPerformanceMonitoring) ValidateQueryPerformanceMonitoringSetup(ctx context.Context) ([]QueryPerformanceValidation, error) {
	query := `SELECT * FROM validate_query_performance_monitoring_setup()`

	rows, err := qpm.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to validate query performance monitoring setup: %w", err)
	}
	defer rows.Close()

	var results []QueryPerformanceValidation
	for rows.Next() {
		var result QueryPerformanceValidation
		err := rows.Scan(
			&result.Component,
			&result.Status,
			&result.Details,
			&result.Recommendation,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan query performance validation: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetCurrentQueryPerformanceStatus gets current query performance status summary
func (qpm *QueryPerformanceMonitoring) GetCurrentQueryPerformanceStatus(ctx context.Context) (map[string]interface{}, error) {
	status := make(map[string]interface{})

	// Get performance stats
	stats, err := qpm.GetQueryPerformanceStats(ctx, 1)
	if err != nil {
		log.Printf("Warning: failed to get query performance stats: %v", err)
	} else {
		status["performance_stats"] = stats
	}

	// Get performance alerts
	alerts, err := qpm.GetQueryPerformanceAlerts(ctx, 1)
	if err != nil {
		log.Printf("Warning: failed to get query performance alerts: %v", err)
	} else {
		status["performance_alerts"] = alerts
	}

	// Get performance insights
	insights, err := qpm.GetQueryPerformanceInsights(ctx)
	if err != nil {
		log.Printf("Warning: failed to get query performance insights: %v", err)
	} else {
		status["performance_insights"] = insights
	}

	// Determine overall status
	overallStatus := "OK"
	if stats != nil {
		if stats.AvgExecutionTimeMs > 1000 {
			overallStatus = "CRITICAL"
		} else if stats.AvgExecutionTimeMs > 500 {
			overallStatus = "WARNING"
		}
	}

	status["overall_status"] = overallStatus
	status["last_checked"] = time.Now()

	return status, nil
}

// MonitorQueryPerformanceContinuously starts continuous query performance monitoring
func (qpm *QueryPerformanceMonitoring) MonitorQueryPerformanceContinuously(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting continuous query performance monitoring with interval: %v", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping continuous query performance monitoring")
			return
		case <-ticker.C:
			// Cleanup old logs
			if deletedCount, err := qpm.CleanupQueryPerformanceLogs(ctx, 30); err != nil {
				log.Printf("Error cleaning up query performance logs: %v", err)
			} else if deletedCount > 0 {
				log.Printf("Cleaned up %d old query performance log entries", deletedCount)
			}

			// Check for alerts
			alerts, err := qpm.GetQueryPerformanceAlerts(ctx, 1)
			if err != nil {
				log.Printf("Error getting query performance alerts: %v", err)
			} else if len(alerts) > 0 {
				log.Printf("Found %d query performance alerts", len(alerts))
				for _, alert := range alerts {
					if alert.AlertLevel == "CRITICAL" {
						log.Printf("CRITICAL ALERT: %s", alert.AlertMessage)
					}
				}
			}
		}
	}
}

// GetQueryPerformanceSummary gets a comprehensive query performance summary
func (qpm *QueryPerformanceMonitoring) GetQueryPerformanceSummary(ctx context.Context) (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// Get current stats
	stats, err := qpm.GetQueryPerformanceStats(ctx, 24)
	if err != nil {
		log.Printf("Warning: failed to get query performance stats: %v", err)
	} else {
		summary["current_stats"] = stats
	}

	// Get trends
	trends, err := qpm.GetQueryPerformanceTrends(ctx, 168) // 7 days
	if err != nil {
		log.Printf("Warning: failed to get query performance trends: %v", err)
	} else {
		summary["trends"] = trends
	}

	// Get insights
	insights, err := qpm.GetQueryPerformanceInsights(ctx)
	if err != nil {
		log.Printf("Warning: failed to get query performance insights: %v", err)
	} else {
		summary["insights"] = insights
	}

	// Get dashboard
	dashboard, err := qpm.GetQueryPerformanceDashboard(ctx)
	if err != nil {
		log.Printf("Warning: failed to get query performance dashboard: %v", err)
	} else {
		summary["dashboard"] = dashboard
	}

	summary["last_updated"] = time.Now()

	return summary, nil
}

// AnalyzeSlowQueries analyzes slow queries and provides optimization recommendations
func (qpm *QueryPerformanceMonitoring) AnalyzeSlowQueries(ctx context.Context, minExecutionTimeMs float64) ([]QueryPerformanceAlert, error) {
	query := `
		SELECT 
			id,
			'SLOW_QUERY' as alert_type,
			CASE 
				WHEN execution_time_ms > 5000 THEN 'CRITICAL'
				WHEN execution_time_ms > 2000 THEN 'HIGH'
				WHEN execution_time_ms > 1000 THEN 'MEDIUM'
				ELSE 'LOW'
			END as alert_level,
			'Slow query detected: ' || ROUND(execution_time_ms, 2) || 'ms execution time' as alert_message,
			query_id,
			LEFT(query_text, 200) as query_text,
			execution_time_ms,
			performance_category,
			recommendations,
			executed_at
		FROM query_performance_log
		WHERE execution_time_ms >= $1
		AND executed_at >= NOW() - INTERVAL '24 hours'
		ORDER BY execution_time_ms DESC
		LIMIT 50
	`

	rows, err := qpm.db.QueryContext(ctx, query, minExecutionTimeMs)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze slow queries: %w", err)
	}
	defer rows.Close()

	var results []QueryPerformanceAlert
	for rows.Next() {
		var result QueryPerformanceAlert
		err := rows.Scan(
			&result.AlertID,
			&result.AlertType,
			&result.AlertLevel,
			&result.AlertMessage,
			&result.QueryID,
			&result.QueryText,
			&result.ExecutionTimeMs,
			&result.PerformanceCategory,
			&result.Recommendations,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan slow query: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetQueryPerformanceMetrics gets key query performance metrics
func (qpm *QueryPerformanceMonitoring) GetQueryPerformanceMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Get basic metrics
	query := `
		SELECT 
			COUNT(*) as total_queries,
			ROUND(AVG(execution_time_ms), 2) as avg_execution_time,
			ROUND(MAX(execution_time_ms), 2) as max_execution_time,
			COUNT(CASE WHEN execution_time_ms > 1000 THEN 1 END) as slow_queries,
			ROUND(AVG(index_usage_score), 2) as avg_index_usage,
			ROUND(AVG(cache_hit_ratio), 2) as avg_cache_hit_ratio
		FROM query_performance_log
		WHERE executed_at >= NOW() - INTERVAL '1 hour'
	`

	var totalQueries int64
	var avgExecutionTime, maxExecutionTime, avgIndexUsage, avgCacheHitRatio float64
	var slowQueries int64

	err := qpm.db.QueryRowContext(ctx, query).Scan(
		&totalQueries,
		&avgExecutionTime,
		&maxExecutionTime,
		&slowQueries,
		&avgIndexUsage,
		&avgCacheHitRatio,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get query performance metrics: %w", err)
	}

	metrics["total_queries"] = totalQueries
	metrics["avg_execution_time_ms"] = avgExecutionTime
	metrics["max_execution_time_ms"] = maxExecutionTime
	metrics["slow_queries"] = slowQueries
	metrics["avg_index_usage_score"] = avgIndexUsage
	metrics["avg_cache_hit_ratio"] = avgCacheHitRatio
	metrics["last_updated"] = time.Now()

	return metrics, nil
}
