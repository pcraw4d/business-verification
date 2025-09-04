package classification

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// ConnectionPoolMonitoring provides comprehensive database connection pool monitoring and optimization
type ConnectionPoolMonitoring struct {
	db *sql.DB
}

// NewConnectionPoolMonitoring creates a new instance of ConnectionPoolMonitoring
func NewConnectionPoolMonitoring(db *sql.DB) *ConnectionPoolMonitoring {
	return &ConnectionPoolMonitoring{
		db: db,
	}
}

// ConnectionPoolStats represents connection pool statistics
type ConnectionPoolStats struct {
	ActiveConnections            int     `json:"active_connections"`
	IdleConnections              int     `json:"idle_connections"`
	TotalConnections             int     `json:"total_connections"`
	MaxConnections               int     `json:"max_connections"`
	ConnectionUtilization        float64 `json:"connection_utilization"`
	AvgConnectionDurationSeconds float64 `json:"avg_connection_duration_seconds"`
	ConnectionErrors             int     `json:"connection_errors"`
	ConnectionTimeouts           int     `json:"connection_timeouts"`
	PoolHitRatio                 float64 `json:"pool_hit_ratio"`
	PoolMissRatio                float64 `json:"pool_miss_ratio"`
	AvgWaitTimeMs                float64 `json:"avg_wait_time_ms"`
	MaxWaitTimeMs                float64 `json:"max_wait_time_ms"`
	ConnectionCreationRate       float64 `json:"connection_creation_rate"`
	ConnectionDestructionRate    float64 `json:"connection_destruction_rate"`
	PoolStatus                   string  `json:"pool_status"`
}

// ConnectionPoolTrend represents connection pool trend data
type ConnectionPoolTrend struct {
	HourBucket               time.Time `json:"hour_bucket"`
	AvgActiveConnections     float64   `json:"avg_active_connections"`
	AvgIdleConnections       float64   `json:"avg_idle_connections"`
	AvgTotalConnections      float64   `json:"avg_total_connections"`
	AvgConnectionUtilization float64   `json:"avg_connection_utilization"`
	AvgPoolHitRatio          float64   `json:"avg_pool_hit_ratio"`
	AvgWaitTimeMs            float64   `json:"avg_wait_time_ms"`
	ConnectionErrorsCount    int64     `json:"connection_errors_count"`
	PoolStatusChanges        int64     `json:"pool_status_changes"`
}

// ConnectionPoolAlert represents a connection pool alert
type ConnectionPoolAlert struct {
	AlertID         int       `json:"alert_id"`
	AlertType       string    `json:"alert_type"`
	AlertLevel      string    `json:"alert_level"`
	AlertMessage    string    `json:"alert_message"`
	MetricValue     float64   `json:"metric_value"`
	ThresholdValue  float64   `json:"threshold_value"`
	Recommendations string    `json:"recommendations"`
	CreatedAt       time.Time `json:"created_at"`
}

// ConnectionPoolDashboard represents connection pool dashboard data
type ConnectionPoolDashboard struct {
	MetricName      string `json:"metric_name"`
	CurrentValue    string `json:"current_value"`
	TargetValue     string `json:"target_value"`
	Status          string `json:"status"`
	Trend           string `json:"trend"`
	Recommendations string `json:"recommendations"`
}

// ConnectionPoolInsight represents connection pool insights
type ConnectionPoolInsight struct {
	InsightType            string  `json:"insight_type"`
	InsightTitle           string  `json:"insight_title"`
	InsightDescription     string  `json:"insight_description"`
	InsightPriority        string  `json:"insight_priority"`
	InsightRecommendations string  `json:"insight_recommendations"`
	AffectedConnections    int64   `json:"affected_connections"`
	PotentialImprovement   float64 `json:"potential_improvement"`
}

// ConnectionPoolOptimization represents connection pool optimization recommendations
type ConnectionPoolOptimization struct {
	SettingName      string `json:"setting_name"`
	CurrentValue     string `json:"current_value"`
	RecommendedValue string `json:"recommended_value"`
	Reason           string `json:"reason"`
	ImpactLevel      string `json:"impact_level"`
}

// ConnectionPoolValidation represents connection pool monitoring setup validation
type ConnectionPoolValidation struct {
	Component      string `json:"component"`
	Status         string `json:"status"`
	Details        string `json:"details"`
	Recommendation string `json:"recommendation"`
}

// GetConnectionPoolStats gets current connection pool statistics
func (cpm *ConnectionPoolMonitoring) GetConnectionPoolStats(ctx context.Context) (*ConnectionPoolStats, error) {
	query := `SELECT * FROM get_connection_pool_stats()`

	var result ConnectionPoolStats
	err := cpm.db.QueryRowContext(ctx, query).Scan(
		&result.ActiveConnections,
		&result.IdleConnections,
		&result.TotalConnections,
		&result.MaxConnections,
		&result.ConnectionUtilization,
		&result.AvgConnectionDurationSeconds,
		&result.ConnectionErrors,
		&result.ConnectionTimeouts,
		&result.PoolHitRatio,
		&result.PoolMissRatio,
		&result.AvgWaitTimeMs,
		&result.MaxWaitTimeMs,
		&result.ConnectionCreationRate,
		&result.ConnectionDestructionRate,
		&result.PoolStatus,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get connection pool stats: %w", err)
	}

	return &result, nil
}

// LogConnectionPoolMetrics logs connection pool metrics
func (cpm *ConnectionPoolMonitoring) LogConnectionPoolMetrics(ctx context.Context, stats *ConnectionPoolStats, recommendations *string) (int, error) {
	query := `SELECT log_connection_pool_metrics($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	var logID int
	err := cpm.db.QueryRowContext(ctx, query,
		stats.ActiveConnections,
		stats.IdleConnections,
		stats.TotalConnections,
		stats.MaxConnections,
		stats.ConnectionUtilization,
		stats.AvgConnectionDurationSeconds,
		stats.ConnectionErrors,
		stats.ConnectionTimeouts,
		stats.PoolHitRatio,
		stats.PoolMissRatio,
		stats.AvgWaitTimeMs,
		stats.MaxWaitTimeMs,
		stats.ConnectionCreationRate,
		stats.ConnectionDestructionRate,
		stats.PoolStatus,
		recommendations,
	).Scan(&logID)

	if err != nil {
		return 0, fmt.Errorf("failed to log connection pool metrics: %w", err)
	}

	return logID, nil
}

// GetConnectionPoolTrends gets connection pool trends
func (cpm *ConnectionPoolMonitoring) GetConnectionPoolTrends(ctx context.Context, hoursBack int) ([]ConnectionPoolTrend, error) {
	query := `SELECT * FROM get_connection_pool_trends($1) ORDER BY hour_bucket DESC`

	rows, err := cpm.db.QueryContext(ctx, query, hoursBack)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection pool trends: %w", err)
	}
	defer rows.Close()

	var results []ConnectionPoolTrend
	for rows.Next() {
		var result ConnectionPoolTrend
		err := rows.Scan(
			&result.HourBucket,
			&result.AvgActiveConnections,
			&result.AvgIdleConnections,
			&result.AvgTotalConnections,
			&result.AvgConnectionUtilization,
			&result.AvgPoolHitRatio,
			&result.AvgWaitTimeMs,
			&result.ConnectionErrorsCount,
			&result.PoolStatusChanges,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan connection pool trend: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetConnectionPoolAlerts gets connection pool alerts
func (cpm *ConnectionPoolMonitoring) GetConnectionPoolAlerts(ctx context.Context, hoursBack int) ([]ConnectionPoolAlert, error) {
	query := `SELECT * FROM get_connection_pool_alerts($1)`

	rows, err := cpm.db.QueryContext(ctx, query, hoursBack)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection pool alerts: %w", err)
	}
	defer rows.Close()

	var results []ConnectionPoolAlert
	for rows.Next() {
		var result ConnectionPoolAlert
		err := rows.Scan(
			&result.AlertID,
			&result.AlertType,
			&result.AlertLevel,
			&result.AlertMessage,
			&result.MetricValue,
			&result.ThresholdValue,
			&result.Recommendations,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan connection pool alert: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetConnectionPoolDashboard gets connection pool dashboard data
func (cpm *ConnectionPoolMonitoring) GetConnectionPoolDashboard(ctx context.Context) ([]ConnectionPoolDashboard, error) {
	query := `SELECT * FROM get_connection_pool_dashboard()`

	rows, err := cpm.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection pool dashboard: %w", err)
	}
	defer rows.Close()

	var results []ConnectionPoolDashboard
	for rows.Next() {
		var result ConnectionPoolDashboard
		err := rows.Scan(
			&result.MetricName,
			&result.CurrentValue,
			&result.TargetValue,
			&result.Status,
			&result.Trend,
			&result.Recommendations,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan connection pool dashboard: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetConnectionPoolInsights gets connection pool insights
func (cpm *ConnectionPoolMonitoring) GetConnectionPoolInsights(ctx context.Context) ([]ConnectionPoolInsight, error) {
	query := `SELECT * FROM get_connection_pool_insights()`

	rows, err := cpm.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection pool insights: %w", err)
	}
	defer rows.Close()

	var results []ConnectionPoolInsight
	for rows.Next() {
		var result ConnectionPoolInsight
		err := rows.Scan(
			&result.InsightType,
			&result.InsightTitle,
			&result.InsightDescription,
			&result.InsightPriority,
			&result.InsightRecommendations,
			&result.AffectedConnections,
			&result.PotentialImprovement,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan connection pool insight: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// OptimizeConnectionPoolSettings gets connection pool optimization recommendations
func (cpm *ConnectionPoolMonitoring) OptimizeConnectionPoolSettings(ctx context.Context) ([]ConnectionPoolOptimization, error) {
	query := `SELECT * FROM optimize_connection_pool_settings()`

	rows, err := cpm.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection pool optimization settings: %w", err)
	}
	defer rows.Close()

	var results []ConnectionPoolOptimization
	for rows.Next() {
		var result ConnectionPoolOptimization
		err := rows.Scan(
			&result.SettingName,
			&result.CurrentValue,
			&result.RecommendedValue,
			&result.Reason,
			&result.ImpactLevel,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan connection pool optimization: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// CleanupConnectionPoolMetrics cleans up old connection pool metrics
func (cpm *ConnectionPoolMonitoring) CleanupConnectionPoolMetrics(ctx context.Context, daysToKeep int) (int, error) {
	query := `SELECT cleanup_connection_pool_metrics($1)`

	var deletedCount int
	err := cpm.db.QueryRowContext(ctx, query, daysToKeep).Scan(&deletedCount)

	if err != nil {
		return 0, fmt.Errorf("failed to cleanup connection pool metrics: %w", err)
	}

	return deletedCount, nil
}

// ValidateConnectionPoolMonitoringSetup validates connection pool monitoring setup
func (cpm *ConnectionPoolMonitoring) ValidateConnectionPoolMonitoringSetup(ctx context.Context) ([]ConnectionPoolValidation, error) {
	query := `SELECT * FROM validate_connection_pool_monitoring_setup()`

	rows, err := cpm.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to validate connection pool monitoring setup: %w", err)
	}
	defer rows.Close()

	var results []ConnectionPoolValidation
	for rows.Next() {
		var result ConnectionPoolValidation
		err := rows.Scan(
			&result.Component,
			&result.Status,
			&result.Details,
			&result.Recommendation,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan connection pool validation: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetCurrentConnectionPoolStatus gets current connection pool status summary
func (cpm *ConnectionPoolMonitoring) GetCurrentConnectionPoolStatus(ctx context.Context) (map[string]interface{}, error) {
	status := make(map[string]interface{})

	// Get connection pool stats
	stats, err := cpm.GetConnectionPoolStats(ctx)
	if err != nil {
		log.Printf("Warning: failed to get connection pool stats: %v", err)
	} else {
		status["connection_pool_stats"] = stats
	}

	// Get connection pool alerts
	alerts, err := cpm.GetConnectionPoolAlerts(ctx, 1)
	if err != nil {
		log.Printf("Warning: failed to get connection pool alerts: %v", err)
	} else {
		status["connection_pool_alerts"] = alerts
	}

	// Get connection pool insights
	insights, err := cpm.GetConnectionPoolInsights(ctx)
	if err != nil {
		log.Printf("Warning: failed to get connection pool insights: %v", err)
	} else {
		status["connection_pool_insights"] = insights
	}

	// Determine overall status
	overallStatus := "OK"
	if stats != nil {
		if stats.ConnectionUtilization > 90 {
			overallStatus = "CRITICAL"
		} else if stats.ConnectionUtilization > 75 {
			overallStatus = "WARNING"
		}
	}

	status["overall_status"] = overallStatus
	status["last_checked"] = time.Now()

	return status, nil
}

// MonitorConnectionPoolContinuously starts continuous connection pool monitoring
func (cpm *ConnectionPoolMonitoring) MonitorConnectionPoolContinuously(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting continuous connection pool monitoring with interval: %v", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping continuous connection pool monitoring")
			return
		case <-ticker.C:
			// Get current stats
			stats, err := cpm.GetConnectionPoolStats(ctx)
			if err != nil {
				log.Printf("Error getting connection pool stats: %v", err)
				continue
			}

			// Log metrics
			if _, err := cpm.LogConnectionPoolMetrics(ctx, stats, nil); err != nil {
				log.Printf("Error logging connection pool metrics: %v", err)
			}

			// Check for alerts
			alerts, err := cpm.GetConnectionPoolAlerts(ctx, 1)
			if err != nil {
				log.Printf("Error getting connection pool alerts: %v", err)
			} else if len(alerts) > 0 {
				log.Printf("Found %d connection pool alerts", len(alerts))
				for _, alert := range alerts {
					if alert.AlertLevel == "CRITICAL" {
						log.Printf("CRITICAL ALERT: %s", alert.AlertMessage)
					}
				}
			}

			// Cleanup old metrics
			if deletedCount, err := cpm.CleanupConnectionPoolMetrics(ctx, 30); err != nil {
				log.Printf("Error cleaning up connection pool metrics: %v", err)
			} else if deletedCount > 0 {
				log.Printf("Cleaned up %d old connection pool metric entries", deletedCount)
			}
		}
	}
}

// GetConnectionPoolSummary gets a comprehensive connection pool summary
func (cpm *ConnectionPoolMonitoring) GetConnectionPoolSummary(ctx context.Context) (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// Get current stats
	stats, err := cpm.GetConnectionPoolStats(ctx)
	if err != nil {
		log.Printf("Warning: failed to get connection pool stats: %v", err)
	} else {
		summary["current_stats"] = stats
	}

	// Get trends
	trends, err := cpm.GetConnectionPoolTrends(ctx, 168) // 7 days
	if err != nil {
		log.Printf("Warning: failed to get connection pool trends: %v", err)
	} else {
		summary["trends"] = trends
	}

	// Get insights
	insights, err := cpm.GetConnectionPoolInsights(ctx)
	if err != nil {
		log.Printf("Warning: failed to get connection pool insights: %v", err)
	} else {
		summary["insights"] = insights
	}

	// Get optimization recommendations
	optimizations, err := cpm.OptimizeConnectionPoolSettings(ctx)
	if err != nil {
		log.Printf("Warning: failed to get connection pool optimizations: %v", err)
	} else {
		summary["optimizations"] = optimizations
	}

	// Get dashboard
	dashboard, err := cpm.GetConnectionPoolDashboard(ctx)
	if err != nil {
		log.Printf("Warning: failed to get connection pool dashboard: %v", err)
	} else {
		summary["dashboard"] = dashboard
	}

	summary["last_updated"] = time.Now()

	return summary, nil
}

// AnalyzeConnectionPoolPerformance analyzes connection pool performance and provides recommendations
func (cpm *ConnectionPoolMonitoring) AnalyzeConnectionPoolPerformance(ctx context.Context) (map[string]interface{}, error) {
	analysis := make(map[string]interface{})

	// Get current stats
	stats, err := cpm.GetConnectionPoolStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection pool stats: %w", err)
	}

	// Analyze utilization
	utilizationAnalysis := make(map[string]interface{})
	utilizationAnalysis["current_utilization"] = stats.ConnectionUtilization
	utilizationAnalysis["status"] = stats.PoolStatus
	utilizationAnalysis["recommendation"] = cpm.getUtilizationRecommendation(stats.ConnectionUtilization)
	analysis["utilization"] = utilizationAnalysis

	// Analyze hit ratio
	hitRatioAnalysis := make(map[string]interface{})
	hitRatioAnalysis["current_hit_ratio"] = stats.PoolHitRatio
	hitRatioAnalysis["status"] = cpm.getHitRatioStatus(stats.PoolHitRatio)
	hitRatioAnalysis["recommendation"] = cpm.getHitRatioRecommendation(stats.PoolHitRatio)
	analysis["hit_ratio"] = hitRatioAnalysis

	// Analyze wait times
	waitTimeAnalysis := make(map[string]interface{})
	waitTimeAnalysis["current_wait_time"] = stats.AvgWaitTimeMs
	waitTimeAnalysis["status"] = cpm.getWaitTimeStatus(stats.AvgWaitTimeMs)
	waitTimeAnalysis["recommendation"] = cpm.getWaitTimeRecommendation(stats.AvgWaitTimeMs)
	analysis["wait_time"] = waitTimeAnalysis

	// Analyze errors
	errorAnalysis := make(map[string]interface{})
	errorAnalysis["current_errors"] = stats.ConnectionErrors
	errorAnalysis["status"] = cpm.getErrorStatus(stats.ConnectionErrors)
	errorAnalysis["recommendation"] = cpm.getErrorRecommendation(stats.ConnectionErrors)
	analysis["errors"] = errorAnalysis

	// Overall performance score
	performanceScore := cpm.calculatePerformanceScore(stats)
	analysis["performance_score"] = performanceScore
	analysis["overall_recommendation"] = cpm.getOverallRecommendation(performanceScore)

	return analysis, nil
}

// Helper methods for analysis
func (cpm *ConnectionPoolMonitoring) getUtilizationRecommendation(utilization float64) string {
	if utilization > 90 {
		return "Critical: Increase max_connections immediately"
	} else if utilization > 75 {
		return "Warning: Consider increasing max_connections"
	} else if utilization > 50 {
		return "Fair: Monitor utilization closely"
	}
	return "Good: Utilization is healthy"
}

func (cpm *ConnectionPoolMonitoring) getHitRatioStatus(hitRatio float64) string {
	if hitRatio < 80 {
		return "CRITICAL"
	} else if hitRatio < 90 {
		return "WARNING"
	} else if hitRatio < 95 {
		return "FAIR"
	}
	return "GOOD"
}

func (cpm *ConnectionPoolMonitoring) getHitRatioRecommendation(hitRatio float64) string {
	if hitRatio < 80 {
		return "Critical: Increase connection pool size"
	} else if hitRatio < 90 {
		return "Warning: Optimize connection reuse"
	} else if hitRatio < 95 {
		return "Fair: Monitor hit ratio"
	}
	return "Good: Hit ratio is excellent"
}

func (cpm *ConnectionPoolMonitoring) getWaitTimeStatus(waitTime float64) string {
	if waitTime > 1000 {
		return "CRITICAL"
	} else if waitTime > 500 {
		return "WARNING"
	} else if waitTime > 100 {
		return "FAIR"
	}
	return "GOOD"
}

func (cpm *ConnectionPoolMonitoring) getWaitTimeRecommendation(waitTime float64) string {
	if waitTime > 1000 {
		return "Critical: Increase connection pool size"
	} else if waitTime > 500 {
		return "Warning: Optimize connection management"
	} else if waitTime > 100 {
		return "Fair: Monitor wait times"
	}
	return "Good: Wait times are acceptable"
}

func (cpm *ConnectionPoolMonitoring) getErrorStatus(errors int) string {
	if errors > 10 {
		return "CRITICAL"
	} else if errors > 5 {
		return "WARNING"
	} else if errors > 1 {
		return "FAIR"
	}
	return "GOOD"
}

func (cpm *ConnectionPoolMonitoring) getErrorRecommendation(errors int) string {
	if errors > 10 {
		return "Critical: Investigate connection errors immediately"
	} else if errors > 5 {
		return "Warning: Check database health"
	} else if errors > 1 {
		return "Fair: Monitor connection errors"
	}
	return "Good: No connection errors"
}

func (cpm *ConnectionPoolMonitoring) calculatePerformanceScore(stats *ConnectionPoolStats) float64 {
	score := 100.0

	// Deduct points for high utilization
	if stats.ConnectionUtilization > 90 {
		score -= 30
	} else if stats.ConnectionUtilization > 75 {
		score -= 20
	} else if stats.ConnectionUtilization > 50 {
		score -= 10
	}

	// Deduct points for low hit ratio
	if stats.PoolHitRatio < 80 {
		score -= 25
	} else if stats.PoolHitRatio < 90 {
		score -= 15
	} else if stats.PoolHitRatio < 95 {
		score -= 5
	}

	// Deduct points for high wait times
	if stats.AvgWaitTimeMs > 1000 {
		score -= 20
	} else if stats.AvgWaitTimeMs > 500 {
		score -= 15
	} else if stats.AvgWaitTimeMs > 100 {
		score -= 10
	}

	// Deduct points for errors
	if stats.ConnectionErrors > 10 {
		score -= 25
	} else if stats.ConnectionErrors > 5 {
		score -= 15
	} else if stats.ConnectionErrors > 1 {
		score -= 10
	}

	if score < 0 {
		score = 0
	}

	return score
}

func (cpm *ConnectionPoolMonitoring) getOverallRecommendation(score float64) string {
	if score >= 80 {
		return "Excellent: Connection pool performance is optimal"
	} else if score >= 60 {
		return "Good: Connection pool performance is acceptable"
	} else if score >= 40 {
		return "Fair: Connection pool performance needs improvement"
	} else if score >= 20 {
		return "Poor: Connection pool performance requires optimization"
	}
	return "Critical: Connection pool performance is severely degraded"
}
