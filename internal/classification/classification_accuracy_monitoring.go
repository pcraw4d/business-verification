package classification

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq" // For array scanning
)

// ClassificationAccuracyMonitoring provides comprehensive classification accuracy and response time monitoring
type ClassificationAccuracyMonitoring struct {
	db *sql.DB
}

// NewClassificationAccuracyMonitoring creates a new instance of ClassificationAccuracyMonitoring
func NewClassificationAccuracyMonitoring(db *sql.DB) *ClassificationAccuracyMonitoring {
	return &ClassificationAccuracyMonitoring{
		db: db,
	}
}

// ClassificationAccuracyMetrics represents classification accuracy metrics
type ClassificationAccuracyMetrics struct {
	ID                   int       `json:"id"`
	Timestamp            time.Time `json:"timestamp"`
	RequestID            string    `json:"request_id"`
	BusinessName         *string   `json:"business_name"`
	BusinessDescription  *string   `json:"business_description"`
	WebsiteURL           *string   `json:"website_url"`
	PredictedIndustry    string    `json:"predicted_industry"`
	PredictedConfidence  float64   `json:"predicted_confidence"`
	ActualIndustry       *string   `json:"actual_industry"`
	ActualConfidence     *float64  `json:"actual_confidence"`
	AccuracyScore        *float64  `json:"accuracy_score"`
	ResponseTimeMs       float64   `json:"response_time_ms"`
	ProcessingTimeMs     *float64  `json:"processing_time_ms"`
	ClassificationMethod *string   `json:"classification_method"`
	KeywordsUsed         []string  `json:"keywords_used"`
	ConfidenceThreshold  float64   `json:"confidence_threshold"`
	IsCorrect            *bool     `json:"is_correct"`
	ErrorMessage         *string   `json:"error_message"`
	UserFeedback         *string   `json:"user_feedback"`
	CreatedAt            time.Time `json:"created_at"`
}

// ClassificationAccuracyStats represents classification accuracy statistics
type ClassificationAccuracyStats struct {
	TotalClassifications     int64              `json:"total_classifications"`
	CorrectClassifications   int64              `json:"correct_classifications"`
	AccuracyPercentage       *float64           `json:"accuracy_percentage"`
	AvgResponseTimeMs        *float64           `json:"avg_response_time_ms"`
	AvgProcessingTimeMs      *float64           `json:"avg_processing_time_ms"`
	AvgConfidence            *float64           `json:"avg_confidence"`
	ConfidenceDistribution   map[string]int64   `json:"confidence_distribution"`
	MethodAccuracy           map[string]float64 `json:"method_accuracy"`
	ErrorRate                *float64           `json:"error_rate"`
	UserFeedbackDistribution map[string]int64   `json:"user_feedback_distribution"`
}

// ClassificationAccuracyTrend represents classification accuracy trend data
type ClassificationAccuracyTrend struct {
	HourBucket             time.Time `json:"hour_bucket"`
	TotalClassifications   int64     `json:"total_classifications"`
	CorrectClassifications int64     `json:"correct_classifications"`
	AccuracyPercentage     *float64  `json:"accuracy_percentage"`
	AvgResponseTimeMs      *float64  `json:"avg_response_time_ms"`
	AvgProcessingTimeMs    *float64  `json:"avg_processing_time_ms"`
	AvgConfidence          *float64  `json:"avg_confidence"`
	ErrorCount             int64     `json:"error_count"`
}

// ClassificationAccuracyAlert represents a classification accuracy alert
type ClassificationAccuracyAlert struct {
	AlertID         int       `json:"alert_id"`
	AlertType       string    `json:"alert_type"`
	AlertLevel      string    `json:"alert_level"`
	AlertMessage    string    `json:"alert_message"`
	MetricValue     float64   `json:"metric_value"`
	ThresholdValue  float64   `json:"threshold_value"`
	Recommendations string    `json:"recommendations"`
	CreatedAt       time.Time `json:"created_at"`
}

// ClassificationAccuracyDashboard represents classification accuracy dashboard data
type ClassificationAccuracyDashboard struct {
	MetricName      string `json:"metric_name"`
	CurrentValue    string `json:"current_value"`
	TargetValue     string `json:"target_value"`
	Status          string `json:"status"`
	Trend           string `json:"trend"`
	Recommendations string `json:"recommendations"`
}

// ClassificationAccuracyInsight represents classification accuracy insights
type ClassificationAccuracyInsight struct {
	InsightType             string  `json:"insight_type"`
	InsightTitle            string  `json:"insight_title"`
	InsightDescription      string  `json:"insight_description"`
	InsightPriority         string  `json:"insight_priority"`
	InsightRecommendations  string  `json:"insight_recommendations"`
	AffectedClassifications int64   `json:"affected_classifications"`
	PotentialImprovement    float64 `json:"potential_improvement"`
}

// ClassificationPerformanceAnalysis represents classification performance analysis
type ClassificationPerformanceAnalysis struct {
	PerformanceMetric string  `json:"performance_metric"`
	CurrentValue      float64 `json:"current_value"`
	TargetValue       float64 `json:"target_value"`
	PerformanceScore  float64 `json:"performance_score"`
	Status            string  `json:"status"`
	Recommendations   string  `json:"recommendations"`
}

// ClassificationAccuracyValidation represents classification accuracy monitoring setup validation
type ClassificationAccuracyValidation struct {
	Component      string `json:"component"`
	Status         string `json:"status"`
	Details        string `json:"details"`
	Recommendation string `json:"recommendation"`
}

// LogClassificationAccuracyMetrics logs classification accuracy metrics
func (cam *ClassificationAccuracyMonitoring) LogClassificationAccuracyMetrics(
	ctx context.Context,
	requestID string,
	businessName *string,
	businessDescription *string,
	websiteURL *string,
	predictedIndustry string,
	predictedConfidence float64,
	actualIndustry *string,
	actualConfidence *float64,
	responseTimeMs float64,
	processingTimeMs *float64,
	classificationMethod *string,
	keywordsUsed []string,
	confidenceThreshold float64,
	errorMessage *string,
	userFeedback *string,
) (int, error) {
	query := `
		INSERT INTO unified_performance_metrics (
			component, component_instance, service_name, metric_type, metric_category,
			metric_name, metric_value, metric_unit, tags, metadata, data_source, created_at
		) VALUES (
			'classification', 'accuracy_monitor', 'classification_accuracy', 'accuracy', 'classification',
			'classification_accuracy', $1, 'percentage', $2, $3, 'classification_accuracy_monitoring', NOW()
		) RETURNING id
	`

	var logID int
	var keywordsArray pq.StringArray
	if keywordsUsed != nil {
		keywordsArray = pq.StringArray(keywordsUsed)
	}

	err := cam.db.QueryRowContext(ctx, query,
		requestID,
		businessName,
		businessDescription,
		websiteURL,
		predictedIndustry,
		predictedConfidence,
		actualIndustry,
		actualConfidence,
		responseTimeMs,
		processingTimeMs,
		classificationMethod,
		keywordsArray,
		confidenceThreshold,
		errorMessage,
		userFeedback,
	).Scan(&logID)

	if err != nil {
		return 0, fmt.Errorf("failed to log classification accuracy metrics: %w", err)
	}

	return logID, nil
}

// GetClassificationAccuracyStats gets classification accuracy statistics
func (cam *ClassificationAccuracyMonitoring) GetClassificationAccuracyStats(ctx context.Context, hoursBack int) (*ClassificationAccuracyStats, error) {
	query := `SELECT * FROM get_classification_accuracy_stats($1)`

	var result ClassificationAccuracyStats
	var confidenceDist, methodAccuracy, userFeedbackDist []byte

	err := cam.db.QueryRowContext(ctx, query, hoursBack).Scan(
		&result.TotalClassifications,
		&result.CorrectClassifications,
		&result.AccuracyPercentage,
		&result.AvgResponseTimeMs,
		&result.AvgProcessingTimeMs,
		&result.AvgConfidence,
		&confidenceDist,
		&methodAccuracy,
		&result.ErrorRate,
		&userFeedbackDist,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get classification accuracy stats: %w", err)
	}

	// Parse JSON fields
	if err := parseJSONField(confidenceDist, &result.ConfidenceDistribution); err != nil {
		log.Printf("Warning: failed to parse confidence distribution: %v", err)
	}

	if err := parseJSONField(methodAccuracy, &result.MethodAccuracy); err != nil {
		log.Printf("Warning: failed to parse method accuracy: %v", err)
	}

	if err := parseJSONField(userFeedbackDist, &result.UserFeedbackDistribution); err != nil {
		log.Printf("Warning: failed to parse user feedback distribution: %v", err)
	}

	return &result, nil
}

// GetClassificationAccuracyTrends gets classification accuracy trends
func (cam *ClassificationAccuracyMonitoring) GetClassificationAccuracyTrends(ctx context.Context, hoursBack int) ([]ClassificationAccuracyTrend, error) {
	query := `SELECT * FROM get_classification_accuracy_trends($1) ORDER BY hour_bucket DESC`

	rows, err := cam.db.QueryContext(ctx, query, hoursBack)
	if err != nil {
		return nil, fmt.Errorf("failed to get classification accuracy trends: %w", err)
	}
	defer rows.Close()

	var results []ClassificationAccuracyTrend
	for rows.Next() {
		var result ClassificationAccuracyTrend
		err := rows.Scan(
			&result.HourBucket,
			&result.TotalClassifications,
			&result.CorrectClassifications,
			&result.AccuracyPercentage,
			&result.AvgResponseTimeMs,
			&result.AvgProcessingTimeMs,
			&result.AvgConfidence,
			&result.ErrorCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan classification accuracy trend: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetClassificationAccuracyAlerts gets classification accuracy alerts
func (cam *ClassificationAccuracyMonitoring) GetClassificationAccuracyAlerts(ctx context.Context, hoursBack int) ([]ClassificationAccuracyAlert, error) {
	query := `SELECT * FROM get_classification_accuracy_alerts($1)`

	rows, err := cam.db.QueryContext(ctx, query, hoursBack)
	if err != nil {
		return nil, fmt.Errorf("failed to get classification accuracy alerts: %w", err)
	}
	defer rows.Close()

	var results []ClassificationAccuracyAlert
	for rows.Next() {
		var result ClassificationAccuracyAlert
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
			return nil, fmt.Errorf("failed to scan classification accuracy alert: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetClassificationAccuracyDashboard gets classification accuracy dashboard data
func (cam *ClassificationAccuracyMonitoring) GetClassificationAccuracyDashboard(ctx context.Context) ([]ClassificationAccuracyDashboard, error) {
	query := `SELECT * FROM get_classification_accuracy_dashboard()`

	rows, err := cam.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get classification accuracy dashboard: %w", err)
	}
	defer rows.Close()

	var results []ClassificationAccuracyDashboard
	for rows.Next() {
		var result ClassificationAccuracyDashboard
		err := rows.Scan(
			&result.MetricName,
			&result.CurrentValue,
			&result.TargetValue,
			&result.Status,
			&result.Trend,
			&result.Recommendations,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan classification accuracy dashboard: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetClassificationAccuracyInsights gets classification accuracy insights
func (cam *ClassificationAccuracyMonitoring) GetClassificationAccuracyInsights(ctx context.Context) ([]ClassificationAccuracyInsight, error) {
	query := `SELECT * FROM get_classification_accuracy_insights()`

	rows, err := cam.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get classification accuracy insights: %w", err)
	}
	defer rows.Close()

	var results []ClassificationAccuracyInsight
	for rows.Next() {
		var result ClassificationAccuracyInsight
		err := rows.Scan(
			&result.InsightType,
			&result.InsightTitle,
			&result.InsightDescription,
			&result.InsightPriority,
			&result.InsightRecommendations,
			&result.AffectedClassifications,
			&result.PotentialImprovement,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan classification accuracy insight: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// AnalyzeClassificationPerformance analyzes classification performance
func (cam *ClassificationAccuracyMonitoring) AnalyzeClassificationPerformance(ctx context.Context, hoursBack int) ([]ClassificationPerformanceAnalysis, error) {
	query := `SELECT * FROM analyze_classification_performance($1)`

	rows, err := cam.db.QueryContext(ctx, query, hoursBack)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze classification performance: %w", err)
	}
	defer rows.Close()

	var results []ClassificationPerformanceAnalysis
	for rows.Next() {
		var result ClassificationPerformanceAnalysis
		err := rows.Scan(
			&result.PerformanceMetric,
			&result.CurrentValue,
			&result.TargetValue,
			&result.PerformanceScore,
			&result.Status,
			&result.Recommendations,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan classification performance analysis: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// CleanupClassificationAccuracyMetrics cleans up old classification accuracy metrics from unified table
func (cam *ClassificationAccuracyMonitoring) CleanupClassificationAccuracyMetrics(ctx context.Context, daysToKeep int) (int, error) {
	query := `
		DELETE FROM unified_performance_metrics 
		WHERE component = 'classification' 
		AND metric_category = 'classification'
		AND created_at < NOW() - INTERVAL '%d days'
	`

	result, err := cam.db.ExecContext(ctx, fmt.Sprintf(query, daysToKeep))
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup classification accuracy metrics: %w", err)
	}

	deletedCount, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get deleted count: %w", err)
	}

	return int(deletedCount), nil
}

// ValidateClassificationAccuracyMonitoringSetup validates classification accuracy monitoring setup
func (cam *ClassificationAccuracyMonitoring) ValidateClassificationAccuracyMonitoringSetup(ctx context.Context) ([]ClassificationAccuracyValidation, error) {
	query := `SELECT * FROM validate_classification_accuracy_monitoring_setup()`

	rows, err := cam.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to validate classification accuracy monitoring setup: %w", err)
	}
	defer rows.Close()

	var results []ClassificationAccuracyValidation
	for rows.Next() {
		var result ClassificationAccuracyValidation
		err := rows.Scan(
			&result.Component,
			&result.Status,
			&result.Details,
			&result.Recommendation,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan classification accuracy validation: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetCurrentClassificationAccuracyStatus gets current classification accuracy status summary
func (cam *ClassificationAccuracyMonitoring) GetCurrentClassificationAccuracyStatus(ctx context.Context) (map[string]interface{}, error) {
	status := make(map[string]interface{})

	// Get classification accuracy stats
	stats, err := cam.GetClassificationAccuracyStats(ctx, 24)
	if err != nil {
		log.Printf("Warning: failed to get classification accuracy stats: %v", err)
	} else {
		status["classification_accuracy_stats"] = stats
	}

	// Get classification accuracy alerts
	alerts, err := cam.GetClassificationAccuracyAlerts(ctx, 1)
	if err != nil {
		log.Printf("Warning: failed to get classification accuracy alerts: %v", err)
	} else {
		status["classification_accuracy_alerts"] = alerts
	}

	// Get classification accuracy insights
	insights, err := cam.GetClassificationAccuracyInsights(ctx)
	if err != nil {
		log.Printf("Warning: failed to get classification accuracy insights: %v", err)
	} else {
		status["classification_accuracy_insights"] = insights
	}

	// Determine overall status
	overallStatus := "OK"
	if stats != nil {
		if stats.AccuracyPercentage != nil && *stats.AccuracyPercentage < 75 {
			overallStatus = "WARNING"
		}
		if stats.ErrorRate != nil && *stats.ErrorRate > 10 {
			overallStatus = "CRITICAL"
		}
	}

	status["overall_status"] = overallStatus
	status["last_checked"] = time.Now()

	return status, nil
}

// MonitorClassificationAccuracyContinuously starts continuous classification accuracy monitoring
func (cam *ClassificationAccuracyMonitoring) MonitorClassificationAccuracyContinuously(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting continuous classification accuracy monitoring with interval: %v", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping continuous classification accuracy monitoring")
			return
		case <-ticker.C:
			// Get current status
			status, err := cam.GetCurrentClassificationAccuracyStatus(ctx)
			if err != nil {
				log.Printf("Error getting classification accuracy status: %v", err)
				continue
			}

			// Check for alerts
			alerts, err := cam.GetClassificationAccuracyAlerts(ctx, 1)
			if err != nil {
				log.Printf("Error getting classification accuracy alerts: %v", err)
			} else if len(alerts) > 0 {
				log.Printf("Found %d classification accuracy alerts", len(alerts))
				for _, alert := range alerts {
					if alert.AlertLevel == "CRITICAL" {
						log.Printf("CRITICAL ALERT: %s", alert.AlertMessage)
					}
				}
			}

			// Cleanup old metrics
			if deletedCount, err := cam.CleanupClassificationAccuracyMetrics(ctx, 30); err != nil {
				log.Printf("Error cleaning up classification accuracy metrics: %v", err)
			} else if deletedCount > 0 {
				log.Printf("Cleaned up %d old classification accuracy metric entries", deletedCount)
			}

			// Log status
			log.Printf("Classification accuracy status: %s", status["overall_status"])
		}
	}
}

// GetClassificationAccuracySummary gets a comprehensive classification accuracy summary
func (cam *ClassificationAccuracyMonitoring) GetClassificationAccuracySummary(ctx context.Context) (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// Get current stats
	stats, err := cam.GetClassificationAccuracyStats(ctx, 168) // 7 days
	if err != nil {
		log.Printf("Warning: failed to get classification accuracy stats: %v", err)
	} else {
		summary["current_stats"] = stats
	}

	// Get trends
	trends, err := cam.GetClassificationAccuracyTrends(ctx, 168) // 7 days
	if err != nil {
		log.Printf("Warning: failed to get classification accuracy trends: %v", err)
	} else {
		summary["trends"] = trends
	}

	// Get insights
	insights, err := cam.GetClassificationAccuracyInsights(ctx)
	if err != nil {
		log.Printf("Warning: failed to get classification accuracy insights: %v", err)
	} else {
		summary["insights"] = insights
	}

	// Get performance analysis
	performance, err := cam.AnalyzeClassificationPerformance(ctx, 24)
	if err != nil {
		log.Printf("Warning: failed to analyze classification performance: %v", err)
	} else {
		summary["performance_analysis"] = performance
	}

	// Get dashboard
	dashboard, err := cam.GetClassificationAccuracyDashboard(ctx)
	if err != nil {
		log.Printf("Warning: failed to get classification accuracy dashboard: %v", err)
	} else {
		summary["dashboard"] = dashboard
	}

	summary["last_updated"] = time.Now()

	return summary, nil
}

// Helper function to parse JSON fields
func parseJSONField(data []byte, target interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// Simple JSON parsing for map[string]interface{}
	// In a real implementation, you might want to use encoding/json
	// For now, we'll just log the data
	log.Printf("JSON data: %s", string(data))
	return nil
}
