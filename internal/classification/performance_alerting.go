package classification

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq" // For array scanning
)

// PerformanceAlerting provides comprehensive performance alerting and monitoring
type PerformanceAlerting struct {
	db *sql.DB
}

// NewPerformanceAlerting creates a new instance of PerformanceAlerting
func NewPerformanceAlerting(db *sql.DB) *PerformanceAlerting {
	return &PerformanceAlerting{
		db: db,
	}
}

// PerformanceAlert represents a performance alert
type PerformanceAlert struct {
	ID               int        `json:"id"`
	AlertID          string     `json:"alert_id"`
	AlertType        string     `json:"alert_type"`
	AlertLevel       string     `json:"alert_level"`
	AlertCategory    string     `json:"alert_category"`
	AlertTitle       string     `json:"alert_title"`
	AlertMessage     string     `json:"alert_message"`
	MetricName       string     `json:"metric_name"`
	MetricValue      *float64   `json:"metric_value"`
	ThresholdValue   *float64   `json:"threshold_value"`
	ThresholdType    string     `json:"threshold_type"`
	SeverityScore    int        `json:"severity_score"`
	Status           string     `json:"status"`
	AcknowledgedBy   *string    `json:"acknowledged_by"`
	AcknowledgedAt   *time.Time `json:"acknowledged_at"`
	ResolvedAt       *time.Time `json:"resolved_at"`
	ResolutionNotes  *string    `json:"resolution_notes"`
	AffectedSystems  []string   `json:"affected_systems"`
	Recommendations  []string   `json:"recommendations"`
	EscalationLevel  int        `json:"escalation_level"`
	EscalationSentAt *time.Time `json:"escalation_sent_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// PerformanceCheckResult represents the result of a performance check
type PerformanceCheckResult struct {
	CheckType       string `json:"check_type"`
	AlertsGenerated int    `json:"alerts_generated"`
	CheckStatus     string `json:"check_status"`
	CheckMessage    string `json:"check_message"`
}

// AlertStatistics represents alert statistics
type AlertStatistics struct {
	TotalAlerts              int64            `json:"total_alerts"`
	ActiveAlerts             int64            `json:"active_alerts"`
	AcknowledgedAlerts       int64            `json:"acknowledged_alerts"`
	ResolvedAlerts           int64            `json:"resolved_alerts"`
	CriticalAlerts           int64            `json:"critical_alerts"`
	HighAlerts               int64            `json:"high_alerts"`
	MediumAlerts             int64            `json:"medium_alerts"`
	LowAlerts                int64            `json:"low_alerts"`
	AlertsByCategory         map[string]int64 `json:"alerts_by_category"`
	AlertsByType             map[string]int64 `json:"alerts_by_type"`
	AvgResolutionTimeMinutes *float64         `json:"avg_resolution_time_minutes"`
}

// PerformanceAlertValidation represents alerting setup validation
type PerformanceAlertValidation struct {
	Component      string `json:"component"`
	Status         string `json:"status"`
	Details        string `json:"details"`
	Recommendation string `json:"recommendation"`
}

// GeneratePerformanceAlert generates a new performance alert
func (pa *PerformanceAlerting) GeneratePerformanceAlert(
	ctx context.Context,
	alertType string,
	alertLevel string,
	alertCategory string,
	alertTitle string,
	alertMessage string,
	metricName string,
	metricValue *float64,
	thresholdValue *float64,
	thresholdType string,
	affectedSystems []string,
	recommendations []string,
) (string, error) {
	query := `SELECT generate_performance_alert($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	var alertID string
	var affectedSystemsArray, recommendationsArray pq.StringArray
	if affectedSystems != nil {
		affectedSystemsArray = pq.StringArray(affectedSystems)
	}
	if recommendations != nil {
		recommendationsArray = pq.StringArray(recommendations)
	}

	err := pa.db.QueryRowContext(ctx, query,
		alertType,
		alertLevel,
		alertCategory,
		alertTitle,
		alertMessage,
		metricName,
		metricValue,
		thresholdValue,
		thresholdType,
		affectedSystemsArray,
		recommendationsArray,
	).Scan(&alertID)

	if err != nil {
		return "", fmt.Errorf("failed to generate performance alert: %w", err)
	}

	return alertID, nil
}

// CheckDatabasePerformanceAlerts checks for database performance alerts
func (pa *PerformanceAlerting) CheckDatabasePerformanceAlerts(ctx context.Context) ([]PerformanceAlert, error) {
	query := `SELECT * FROM check_database_performance_alerts()`

	rows, err := pa.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check database performance alerts: %w", err)
	}
	defer rows.Close()

	var results []PerformanceAlert
	for rows.Next() {
		var result PerformanceAlert
		var recommendations pq.StringArray

		err := rows.Scan(
			&result.AlertID,
			&result.AlertType,
			&result.AlertLevel,
			&result.AlertTitle,
			&result.AlertMessage,
			&result.MetricName,
			&result.MetricValue,
			&result.ThresholdValue,
			&recommendations,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan database performance alert: %w", err)
		}

		result.Recommendations = []string(recommendations)
		results = append(results, result)
	}

	return results, nil
}

// CheckClassificationAccuracyAlerts checks for classification accuracy alerts
func (pa *PerformanceAlerting) CheckClassificationAccuracyAlerts(ctx context.Context) ([]PerformanceAlert, error) {
	query := `SELECT * FROM check_classification_accuracy_alerts()`

	rows, err := pa.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check classification accuracy alerts: %w", err)
	}
	defer rows.Close()

	var results []PerformanceAlert
	for rows.Next() {
		var result PerformanceAlert
		var recommendations pq.StringArray

		err := rows.Scan(
			&result.AlertID,
			&result.AlertType,
			&result.AlertLevel,
			&result.AlertTitle,
			&result.AlertMessage,
			&result.MetricName,
			&result.MetricValue,
			&result.ThresholdValue,
			&recommendations,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan classification accuracy alert: %w", err)
		}

		result.Recommendations = []string(recommendations)
		results = append(results, result)
	}

	return results, nil
}

// CheckSystemResourceAlerts checks for system resource alerts
func (pa *PerformanceAlerting) CheckSystemResourceAlerts(ctx context.Context) ([]PerformanceAlert, error) {
	query := `SELECT * FROM check_system_resource_alerts()`

	rows, err := pa.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check system resource alerts: %w", err)
	}
	defer rows.Close()

	var results []PerformanceAlert
	for rows.Next() {
		var result PerformanceAlert
		var recommendations pq.StringArray

		err := rows.Scan(
			&result.AlertID,
			&result.AlertType,
			&result.AlertLevel,
			&result.AlertTitle,
			&result.AlertMessage,
			&result.MetricName,
			&result.MetricValue,
			&result.ThresholdValue,
			&recommendations,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan system resource alert: %w", err)
		}

		result.Recommendations = []string(recommendations)
		results = append(results, result)
	}

	return results, nil
}

// GetActivePerformanceAlerts gets all active performance alerts
func (pa *PerformanceAlerting) GetActivePerformanceAlerts(ctx context.Context) ([]PerformanceAlert, error) {
	query := `SELECT * FROM get_active_performance_alerts()`

	rows, err := pa.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active performance alerts: %w", err)
	}
	defer rows.Close()

	var results []PerformanceAlert
	for rows.Next() {
		var result PerformanceAlert
		var affectedSystems, recommendations pq.StringArray

		err := rows.Scan(
			&result.AlertID,
			&result.AlertType,
			&result.AlertLevel,
			&result.AlertCategory,
			&result.AlertTitle,
			&result.AlertMessage,
			&result.MetricName,
			&result.MetricValue,
			&result.ThresholdValue,
			&result.SeverityScore,
			&result.Status,
			&affectedSystems,
			&recommendations,
			&result.EscalationLevel,
			&result.CreatedAt,
			&result.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan active performance alert: %w", err)
		}

		result.AffectedSystems = []string(affectedSystems)
		result.Recommendations = []string(recommendations)
		results = append(results, result)
	}

	return results, nil
}

// AcknowledgePerformanceAlert acknowledges a performance alert
func (pa *PerformanceAlerting) AcknowledgePerformanceAlert(ctx context.Context, alertID, acknowledgedBy string) (bool, error) {
	query := `SELECT acknowledge_performance_alert($1, $2)`

	var acknowledged bool
	err := pa.db.QueryRowContext(ctx, query, alertID, acknowledgedBy).Scan(&acknowledged)

	if err != nil {
		return false, fmt.Errorf("failed to acknowledge performance alert: %w", err)
	}

	return acknowledged, nil
}

// ResolvePerformanceAlert resolves a performance alert
func (pa *PerformanceAlerting) ResolvePerformanceAlert(ctx context.Context, alertID string, resolutionNotes *string) (bool, error) {
	query := `SELECT resolve_performance_alert($1, $2)`

	var resolved bool
	err := pa.db.QueryRowContext(ctx, query, alertID, resolutionNotes).Scan(&resolved)

	if err != nil {
		return false, fmt.Errorf("failed to resolve performance alert: %w", err)
	}

	return resolved, nil
}

// RunAllPerformanceChecks runs all performance checks
func (pa *PerformanceAlerting) RunAllPerformanceChecks(ctx context.Context) ([]PerformanceCheckResult, error) {
	query := `SELECT * FROM run_all_performance_checks()`

	rows, err := pa.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to run all performance checks: %w", err)
	}
	defer rows.Close()

	var results []PerformanceCheckResult
	for rows.Next() {
		var result PerformanceCheckResult
		err := rows.Scan(
			&result.CheckType,
			&result.AlertsGenerated,
			&result.CheckStatus,
			&result.CheckMessage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan performance check result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetAlertStatistics gets alert statistics
func (pa *PerformanceAlerting) GetAlertStatistics(ctx context.Context, hoursBack int) (*AlertStatistics, error) {
	query := `SELECT * FROM get_alert_statistics($1)`

	var result AlertStatistics
	var alertsByCategory, alertsByType []byte

	err := pa.db.QueryRowContext(ctx, query, hoursBack).Scan(
		&result.TotalAlerts,
		&result.ActiveAlerts,
		&result.AcknowledgedAlerts,
		&result.ResolvedAlerts,
		&result.CriticalAlerts,
		&result.HighAlerts,
		&result.MediumAlerts,
		&result.LowAlerts,
		&alertsByCategory,
		&alertsByType,
		&result.AvgResolutionTimeMinutes,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get alert statistics: %w", err)
	}

	// Parse JSON fields
	if err := parseJSONField(alertsByCategory, &result.AlertsByCategory); err != nil {
		log.Printf("Warning: failed to parse alerts by category: %v", err)
	}

	if err := parseJSONField(alertsByType, &result.AlertsByType); err != nil {
		log.Printf("Warning: failed to parse alerts by type: %v", err)
	}

	return &result, nil
}

// CleanupOldPerformanceAlerts cleans up old performance alerts
func (pa *PerformanceAlerting) CleanupOldPerformanceAlerts(ctx context.Context, daysToKeep int) (int, error) {
	query := `SELECT cleanup_old_performance_alerts($1)`

	var deletedCount int
	err := pa.db.QueryRowContext(ctx, query, daysToKeep).Scan(&deletedCount)

	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old performance alerts: %w", err)
	}

	return deletedCount, nil
}

// ValidateAlertingSetup validates the alerting setup
func (pa *PerformanceAlerting) ValidateAlertingSetup(ctx context.Context) ([]PerformanceAlertValidation, error) {
	query := `SELECT * FROM validate_alerting_setup()`

	rows, err := pa.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to validate alerting setup: %w", err)
	}
	defer rows.Close()

	var results []PerformanceAlertValidation
	for rows.Next() {
		var result PerformanceAlertValidation
		err := rows.Scan(
			&result.Component,
			&result.Status,
			&result.Details,
			&result.Recommendation,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alerting validation: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetCurrentAlertStatus gets current alert status summary
func (pa *PerformanceAlerting) GetCurrentAlertStatus(ctx context.Context) (map[string]interface{}, error) {
	status := make(map[string]interface{})

	// Get active alerts
	activeAlerts, err := pa.GetActivePerformanceAlerts(ctx)
	if err != nil {
		log.Printf("Warning: failed to get active alerts: %v", err)
	} else {
		status["active_alerts"] = activeAlerts
		status["active_alert_count"] = len(activeAlerts)
	}

	// Get alert statistics
	stats, err := pa.GetAlertStatistics(ctx, 24)
	if err != nil {
		log.Printf("Warning: failed to get alert statistics: %v", err)
	} else {
		status["alert_statistics"] = stats
	}

	// Determine overall status
	overallStatus := "OK"
	if len(activeAlerts) > 0 {
		criticalCount := 0
		highCount := 0
		for _, alert := range activeAlerts {
			if alert.AlertLevel == "CRITICAL" {
				criticalCount++
			} else if alert.AlertLevel == "HIGH" {
				highCount++
			}
		}

		if criticalCount > 0 {
			overallStatus = "CRITICAL"
		} else if highCount > 0 {
			overallStatus = "WARNING"
		} else {
			overallStatus = "FAIR"
		}
	}

	status["overall_status"] = overallStatus
	status["last_checked"] = time.Now()

	return status, nil
}

// MonitorPerformanceContinuously starts continuous performance monitoring
func (pa *PerformanceAlerting) MonitorPerformanceContinuously(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting continuous performance monitoring with interval: %v", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping continuous performance monitoring")
			return
		case <-ticker.C:
			// Run all performance checks
			checkResults, err := pa.RunAllPerformanceChecks(ctx)
			if err != nil {
				log.Printf("Error running performance checks: %v", err)
				continue
			}

			// Log check results
			for _, result := range checkResults {
				log.Printf("Performance check %s: %s - %d alerts generated",
					result.CheckType, result.CheckStatus, result.AlertsGenerated)
			}

			// Get current alert status
			status, err := pa.GetCurrentAlertStatus(ctx)
			if err != nil {
				log.Printf("Error getting alert status: %v", err)
			} else {
				log.Printf("Alert status: %s", status["overall_status"])
			}

			// Cleanup old alerts
			if deletedCount, err := pa.CleanupOldPerformanceAlerts(ctx, 30); err != nil {
				log.Printf("Error cleaning up old alerts: %v", err)
			} else if deletedCount > 0 {
				log.Printf("Cleaned up %d old alert entries", deletedCount)
			}
		}
	}
}

// GetPerformanceAlertSummary gets a comprehensive performance alert summary
func (pa *PerformanceAlerting) GetPerformanceAlertSummary(ctx context.Context) (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// Get current status
	status, err := pa.GetCurrentAlertStatus(ctx)
	if err != nil {
		log.Printf("Warning: failed to get current alert status: %v", err)
	} else {
		summary["current_status"] = status
	}

	// Get alert statistics
	stats, err := pa.GetAlertStatistics(ctx, 168) // 7 days
	if err != nil {
		log.Printf("Warning: failed to get alert statistics: %v", err)
	} else {
		summary["alert_statistics"] = stats
	}

	// Get active alerts
	activeAlerts, err := pa.GetActivePerformanceAlerts(ctx)
	if err != nil {
		log.Printf("Warning: failed to get active alerts: %v", err)
	} else {
		summary["active_alerts"] = activeAlerts
	}

	// Run performance checks
	checkResults, err := pa.RunAllPerformanceChecks(ctx)
	if err != nil {
		log.Printf("Warning: failed to run performance checks: %v", err)
	} else {
		summary["performance_checks"] = checkResults
	}

	summary["last_updated"] = time.Now()

	return summary, nil
}
