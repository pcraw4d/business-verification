package enrichment

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// DataQualityMonitor provides comprehensive data quality monitoring and reporting
type DataQualityMonitor struct {
	logger *zap.Logger
	tracer trace.Tracer
	config *DataQualityMonitorConfig

	// Component integrations
	qualityScorer       *DataQualityScorer
	freshnessTracker    *DataFreshnessTracker
	reliabilityAssessor *DataSourceReliabilityAssessor

	// Monitoring data
	mu                 sync.RWMutex
	monitoringSessions map[string]*MonitoringSession
	qualityMetrics     map[string]*QualityMetric
	alertHistory       map[string]*AlertRecord
	reportHistory      map[string]*ReportRecord
	lastCleanup        time.Time
}

// DataQualityMonitorConfig contains configuration for monitoring and reporting
type DataQualityMonitorConfig struct {
	// Monitoring settings
	EnableRealTimeMonitoring bool `json:"enable_real_time_monitoring"`
	EnableAlerting           bool `json:"enable_alerting"`
	EnableReporting          bool `json:"enable_reporting"`
	EnableTrendAnalysis      bool `json:"enable_trend_analysis"`

	// Thresholds
	QualityAlertThreshold     float64       `json:"quality_alert_threshold"`
	FreshnessAlertThreshold   time.Duration `json:"freshness_alert_threshold"`
	ReliabilityAlertThreshold float64       `json:"reliability_alert_threshold"`
	CriticalThreshold         float64       `json:"critical_threshold"`

	// Reporting settings
	ReportGenerationInterval time.Duration `json:"report_generation_interval"`
	ReportRetentionPeriod    time.Duration `json:"report_retention_period"`
	MaxReportsPerSession     int           `json:"max_reports_per_session"`

	// Alert settings
	AlertCooldownPeriod      time.Duration `json:"alert_cooldown_period"`
	MaxAlertsPerSession      int           `json:"max_alerts_per_session"`
	AlertEscalationThreshold int           `json:"alert_escalation_threshold"`

	// Performance settings
	MonitoringInterval     time.Duration `json:"monitoring_interval"`
	MetricsRetentionPeriod time.Duration `json:"metrics_retention_period"`
	CleanupInterval        time.Duration `json:"cleanup_interval"`
}

// MonitoringSession represents an active monitoring session
type MonitoringSession struct {
	SessionID      string    `json:"session_id"`
	DataSourceID   string    `json:"data_source_id"`
	DataSourceType string    `json:"data_source_type"`
	DataSourceName string    `json:"data_source_name"`
	StartTime      time.Time `json:"start_time"`
	LastActivity   time.Time `json:"last_activity"`
	Status         string    `json:"status"` // "active", "paused", "stopped"

	// Quality metrics
	OverallQualityScore float64 `json:"overall_quality_score"`
	QualityTrend        string  `json:"quality_trend"` // "improving", "stable", "declining"
	QualityLevel        string  `json:"quality_level"` // "excellent", "good", "fair", "poor", "critical"

	// Component scores
	CompletenessScore float64 `json:"completeness_score"`
	AccuracyScore     float64 `json:"accuracy_score"`
	ConsistencyScore  float64 `json:"consistency_score"`
	FreshnessScore    float64 `json:"freshness_score"`
	ReliabilityScore  float64 `json:"reliability_score"`
	ValidityScore     float64 `json:"validity_score"`

	// Monitoring data
	AssessmentCount int       `json:"assessment_count"`
	AlertCount      int       `json:"alert_count"`
	ReportCount     int       `json:"report_count"`
	LastAssessment  time.Time `json:"last_assessment"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata"`
}

// QualityMetric represents a single quality measurement
type QualityMetric struct {
	MetricID   string    `json:"metric_id"`
	SessionID  string    `json:"session_id"`
	Timestamp  time.Time `json:"timestamp"`
	MetricType string    `json:"metric_type"` // "quality", "freshness", "reliability"

	// Quality assessment
	QualityScore   float64            `json:"quality_score"`
	QualityLevel   string             `json:"quality_level"`
	QualityDetails map[string]float64 `json:"quality_details"`

	// Freshness assessment
	FreshnessScore  float64       `json:"freshness_score"`
	FreshnessLevel  string        `json:"freshness_level"`
	Age             time.Duration `json:"age"`
	UpdateFrequency time.Duration `json:"update_frequency"`

	// Reliability assessment
	ReliabilityScore float64 `json:"reliability_score"`
	ReliabilityLevel string  `json:"reliability_level"`
	UptimePercentage float64 `json:"uptime_percentage"`
	ErrorRate        float64 `json:"error_rate"`

	// Risk assessment
	RiskLevel   string   `json:"risk_level"`
	RiskFactors []string `json:"risk_factors"`
	RiskScore   float64  `json:"risk_score"`

	// Metadata
	ProcessingTime time.Duration `json:"processing_time"`
	DataPoints     int           `json:"data_points"`
}

// AlertRecord represents a quality alert
type AlertRecord struct {
	AlertID         string     `json:"alert_id"`
	SessionID       string     `json:"session_id"`
	AlertType       string     `json:"alert_type"` // "quality", "freshness", "reliability", "critical"
	Severity        string     `json:"severity"`   // "low", "medium", "high", "critical"
	Message         string     `json:"message"`
	CreatedAt       time.Time  `json:"created_at"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	IsActive        bool       `json:"is_active"`
	EscalationLevel int        `json:"escalation_level"`

	// Alert details
	Threshold    float64 `json:"threshold"`
	CurrentValue float64 `json:"current_value"`
	TriggeredBy  string  `json:"triggered_by"`

	// Resolution
	ResolutionAction string `json:"resolution_action,omitempty"`
	ResolutionNotes  string `json:"resolution_notes,omitempty"`
	ResolvedBy       string `json:"resolved_by,omitempty"`
}

// ReportRecord represents a generated quality report
type ReportRecord struct {
	ReportID    string    `json:"report_id"`
	SessionID   string    `json:"session_id"`
	ReportType  string    `json:"report_type"` // "summary", "detailed", "trend", "alert"
	GeneratedAt time.Time `json:"generated_at"`
	TimeRange   TimeRange `json:"time_range"`

	// Report content
	Summary         *QualitySummary `json:"summary"`
	Trends          []*QualityTrend `json:"trends"`
	Alerts          []*AlertSummary `json:"alerts"`
	Recommendations []string        `json:"recommendations"`
	PriorityActions []string        `json:"priority_actions"`

	// Report metadata
	DataPoints     int           `json:"data_points"`
	ProcessingTime time.Duration `json:"processing_time"`
	ReportSize     int64         `json:"report_size"`
}

// QualitySummary contains summary statistics
type QualitySummary struct {
	OverallQualityScore float64 `json:"overall_quality_score"`
	QualityLevel        string  `json:"quality_level"`
	AssessmentCount     int     `json:"assessment_count"`
	AlertCount          int     `json:"alert_count"`
	CriticalIssues      int     `json:"critical_issues"`

	// Component summaries
	CompletenessSummary *ComponentSummary `json:"completeness_summary"`
	AccuracySummary     *ComponentSummary `json:"accuracy_summary"`
	ConsistencySummary  *ComponentSummary `json:"consistency_summary"`
	FreshnessSummary    *ComponentSummary `json:"freshness_summary"`
	ReliabilitySummary  *ComponentSummary `json:"reliability_summary"`
	ValiditySummary     *ComponentSummary `json:"validity_summary"`

	// Risk summary
	RiskLevel   string   `json:"risk_level"`
	RiskScore   float64  `json:"risk_score"`
	RiskFactors []string `json:"risk_factors"`
}

// ComponentSummary contains component-specific statistics
type ComponentSummary struct {
	AverageScore    float64  `json:"average_score"`
	MinScore        float64  `json:"min_score"`
	MaxScore        float64  `json:"max_score"`
	Trend           string   `json:"trend"`
	Issues          int      `json:"issues"`
	Recommendations []string `json:"recommendations"`
}

// QualityTrend represents a quality trend
type QualityTrend struct {
	Component  string           `json:"component"`
	Direction  string           `json:"direction"` // "improving", "stable", "declining"
	Slope      float64          `json:"slope"`
	Confidence float64          `json:"confidence"`
	Periods    int              `json:"periods"`
	LastChange time.Time        `json:"last_change"`
	Prediction *TrendPrediction `json:"prediction,omitempty"`
}

// TrendPrediction contains trend predictions
type TrendPrediction struct {
	PredictedValue float64    `json:"predicted_value"`
	PredictionTime time.Time  `json:"prediction_time"`
	Confidence     float64    `json:"confidence"`
	Range          [2]float64 `json:"range"`
}

// AlertSummary contains alert summary information
type AlertSummary struct {
	AlertType             string        `json:"alert_type"`
	Severity              string        `json:"severity"`
	Count                 int           `json:"count"`
	ActiveCount           int           `json:"active_count"`
	ResolvedCount         int           `json:"resolved_count"`
	AverageResolutionTime time.Duration `json:"average_resolution_time"`
}

// TimeRange represents a time range
type TimeRange struct {
	Start    time.Time     `json:"start"`
	End      time.Time     `json:"end"`
	Duration time.Duration `json:"duration"`
}

// MonitoringResult contains comprehensive monitoring results
type MonitoringResult struct {
	SessionID       string         `json:"session_id"`
	Timestamp       time.Time      `json:"timestamp"`
	QualityMetric   *QualityMetric `json:"quality_metric"`
	Alerts          []*AlertRecord `json:"alerts"`
	Recommendations []string       `json:"recommendations"`
	PriorityActions []string       `json:"priority_actions"`
	ProcessingTime  time.Duration  `json:"processing_time"`
}

// NewDataQualityMonitor creates a new data quality monitor
func NewDataQualityMonitor(logger *zap.Logger, config *DataQualityMonitorConfig) *DataQualityMonitor {
	if config == nil {
		config = getDefaultDataQualityMonitorConfig()
	}

	return &DataQualityMonitor{
		logger:             logger,
		tracer:             trace.NewNoopTracerProvider().Tracer("data_quality_monitor"),
		config:             config,
		monitoringSessions: make(map[string]*MonitoringSession),
		qualityMetrics:     make(map[string]*QualityMetric),
		alertHistory:       make(map[string]*AlertRecord),
		reportHistory:      make(map[string]*ReportRecord),
		lastCleanup:        time.Now(),
	}
}

// SetComponents sets the integrated components
func (dqm *DataQualityMonitor) SetComponents(qualityScorer *DataQualityScorer, freshnessTracker *DataFreshnessTracker, reliabilityAssessor *DataSourceReliabilityAssessor) {
	dqm.qualityScorer = qualityScorer
	dqm.freshnessTracker = freshnessTracker
	dqm.reliabilityAssessor = reliabilityAssessor
}

// StartMonitoring starts monitoring a data source
func (dqm *DataQualityMonitor) StartMonitoring(ctx context.Context, dataSourceID, dataSourceType, dataSourceName string) (*MonitoringSession, error) {
	ctx, span := dqm.tracer.Start(ctx, "data_quality_monitor.start_monitoring",
		trace.WithAttributes(
			attribute.String("data_source_id", dataSourceID),
			attribute.String("data_source_type", dataSourceType),
			attribute.String("data_source_name", dataSourceName),
		))
	defer span.End()

	dqm.mu.Lock()
	defer dqm.mu.Unlock()

	sessionID := fmt.Sprintf("%s-%s-%d", dataSourceID, dataSourceType, time.Now().Unix())

	session := &MonitoringSession{
		SessionID:      sessionID,
		DataSourceID:   dataSourceID,
		DataSourceType: dataSourceType,
		DataSourceName: dataSourceName,
		StartTime:      time.Now(),
		LastActivity:   time.Now(),
		Status:         "active",
		Metadata:       make(map[string]interface{}),
	}

	dqm.monitoringSessions[sessionID] = session

	dqm.logger.Info("Started monitoring session",
		zap.String("session_id", sessionID),
		zap.String("data_source_id", dataSourceID),
		zap.String("data_source_type", dataSourceType),
		zap.String("data_source_name", dataSourceName))

	return session, nil
}

// MonitorQuality performs comprehensive quality monitoring
func (dqm *DataQualityMonitor) MonitorQuality(ctx context.Context, sessionID string, data interface{}) (*MonitoringResult, error) {
	ctx, span := dqm.tracer.Start(ctx, "data_quality_monitor.monitor_quality",
		trace.WithAttributes(
			attribute.String("session_id", sessionID),
		))
	defer span.End()

	startTime := time.Now()

	dqm.mu.RLock()
	session, exists := dqm.monitoringSessions[sessionID]
	dqm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("monitoring session not found: %s", sessionID)
	}

	// Update session activity
	dqm.mu.Lock()
	session.LastActivity = time.Now()
	session.AssessmentCount++
	dqm.mu.Unlock()

	// Perform comprehensive quality assessment
	qualityMetric := dqm.performQualityAssessment(ctx, session, data)

	// Check for alerts
	alerts := dqm.checkAlerts(ctx, session, qualityMetric)

	// Generate recommendations
	recommendations := dqm.generateRecommendations(qualityMetric, alerts)
	priorityActions := dqm.generatePriorityActions(qualityMetric, alerts)

	// Update session metrics
	dqm.updateSessionMetrics(session, qualityMetric, alerts)

	result := &MonitoringResult{
		SessionID:       sessionID,
		Timestamp:       time.Now(),
		QualityMetric:   qualityMetric,
		Alerts:          alerts,
		Recommendations: recommendations,
		PriorityActions: priorityActions,
		ProcessingTime:  time.Since(startTime),
	}

	dqm.logger.Info("Quality monitoring completed",
		zap.String("session_id", sessionID),
		zap.Float64("quality_score", qualityMetric.QualityScore),
		zap.String("quality_level", qualityMetric.QualityLevel),
		zap.Int("alert_count", len(alerts)),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// GenerateReport generates a comprehensive quality report
func (dqm *DataQualityMonitor) GenerateReport(ctx context.Context, sessionID string, reportType string, timeRange TimeRange) (*ReportRecord, error) {
	ctx, span := dqm.tracer.Start(ctx, "data_quality_monitor.generate_report",
		trace.WithAttributes(
			attribute.String("session_id", sessionID),
			attribute.String("report_type", reportType),
		))
	defer span.End()

	startTime := time.Now()

	dqm.mu.RLock()
	session, exists := dqm.monitoringSessions[sessionID]
	dqm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("monitoring session not found: %s", sessionID)
	}

	// Collect metrics for the time range
	metrics := dqm.collectMetricsForTimeRange(sessionID, timeRange)

	// Generate summary
	summary := dqm.generateQualitySummary(metrics, session)

	// Analyze trends
	trends := dqm.analyzeQualityTrends(metrics)

	// Collect alerts
	alerts := dqm.collectAlertsForTimeRange(sessionID, timeRange)
	alertSummaries := dqm.generateAlertSummaries(alerts)

	// Generate recommendations
	recommendations := dqm.generateReportRecommendations(summary, trends, alerts)
	priorityActions := dqm.generateReportPriorityActions(summary, trends, alerts)

	reportID := fmt.Sprintf("report-%s-%s-%d", sessionID, reportType, time.Now().Unix())

	report := &ReportRecord{
		ReportID:        reportID,
		SessionID:       sessionID,
		ReportType:      reportType,
		GeneratedAt:     time.Now(),
		TimeRange:       timeRange,
		Summary:         summary,
		Trends:          trends,
		Alerts:          alertSummaries,
		Recommendations: recommendations,
		PriorityActions: priorityActions,
		DataPoints:      len(metrics),
		ProcessingTime:  time.Since(startTime),
		ReportSize:      int64(len(reportID) + len(reportType)), // Simplified size calculation
	}

	// Store report
	dqm.mu.Lock()
	dqm.reportHistory[reportID] = report
	session.ReportCount++
	dqm.mu.Unlock()

	dqm.logger.Info("Quality report generated",
		zap.String("report_id", reportID),
		zap.String("session_id", sessionID),
		zap.String("report_type", reportType),
		zap.Int("data_points", report.DataPoints),
		zap.Duration("processing_time", report.ProcessingTime))

	return report, nil
}

// GetMonitoringSession retrieves a monitoring session
func (dqm *DataQualityMonitor) GetMonitoringSession(ctx context.Context, sessionID string) (*MonitoringSession, error) {
	dqm.mu.RLock()
	defer dqm.mu.RUnlock()

	if session, exists := dqm.monitoringSessions[sessionID]; exists {
		return session, nil
	}

	return nil, fmt.Errorf("monitoring session not found: %s", sessionID)
}

// GetQualityMetrics retrieves quality metrics for a session
func (dqm *DataQualityMonitor) GetQualityMetrics(ctx context.Context, sessionID string, limit int) ([]*QualityMetric, error) {
	dqm.mu.RLock()
	defer dqm.mu.RUnlock()

	var metrics []*QualityMetric
	count := 0

	for _, metric := range dqm.qualityMetrics {
		if metric.SessionID == sessionID {
			metrics = append(metrics, metric)
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}

	return metrics, nil
}

// GetActiveAlerts retrieves active alerts for a session
func (dqm *DataQualityMonitor) GetActiveAlerts(ctx context.Context, sessionID string) ([]*AlertRecord, error) {
	dqm.mu.RLock()
	defer dqm.mu.RUnlock()

	var alerts []*AlertRecord
	for _, alert := range dqm.alertHistory {
		if alert.SessionID == sessionID && alert.IsActive {
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// ResolveAlert resolves an alert
func (dqm *DataQualityMonitor) ResolveAlert(ctx context.Context, alertID string, resolutionAction, resolutionNotes, resolvedBy string) error {
	dqm.mu.Lock()
	defer dqm.mu.Unlock()

	if alert, exists := dqm.alertHistory[alertID]; exists {
		now := time.Now()
		alert.ResolvedAt = &now
		alert.IsActive = false
		alert.ResolutionAction = resolutionAction
		alert.ResolutionNotes = resolutionNotes
		alert.ResolvedBy = resolvedBy

		dqm.logger.Info("Alert resolved",
			zap.String("alert_id", alertID),
			zap.String("resolution_action", resolutionAction),
			zap.String("resolved_by", resolvedBy))
	}

	return nil
}

// StopMonitoring stops monitoring a session
func (dqm *DataQualityMonitor) StopMonitoring(ctx context.Context, sessionID string) error {
	dqm.mu.Lock()
	defer dqm.mu.Unlock()

	if session, exists := dqm.monitoringSessions[sessionID]; exists {
		session.Status = "stopped"
		session.LastActivity = time.Now()

		dqm.logger.Info("Stopped monitoring session",
			zap.String("session_id", sessionID),
			zap.Int("assessment_count", session.AssessmentCount),
			zap.Int("alert_count", session.AlertCount))
	}

	return nil
}

// Helper methods

func (dqm *DataQualityMonitor) performQualityAssessment(ctx context.Context, session *MonitoringSession, data interface{}) *QualityMetric {
	metricID := fmt.Sprintf("metric-%s-%d", session.SessionID, time.Now().Unix())

	metric := &QualityMetric{
		MetricID:       metricID,
		SessionID:      session.SessionID,
		Timestamp:      time.Now(),
		MetricType:     "quality",
		QualityDetails: make(map[string]float64),
	}

	// Perform quality assessment if scorer is available
	if dqm.qualityScorer != nil {
		if qualityResult, err := dqm.qualityScorer.AssessDataQuality(ctx, data, session.DataSourceType); err == nil {
			metric.QualityScore = qualityResult.OverallScore
			metric.QualityLevel = qualityResult.QualityLevel
			metric.QualityDetails["completeness"] = qualityResult.CompletenessScore
			metric.QualityDetails["accuracy"] = qualityResult.AccuracyScore
			metric.QualityDetails["consistency"] = qualityResult.ConsistencyScore
			metric.QualityDetails["freshness"] = qualityResult.FreshnessScore
			metric.QualityDetails["reliability"] = qualityResult.ReliabilityScore
			metric.QualityDetails["validity"] = qualityResult.ValidityScore
		}
	}

	// Perform freshness assessment if tracker is available
	if dqm.freshnessTracker != nil {
		if freshnessResult, err := dqm.freshnessTracker.AnalyzeFreshness(ctx, session.DataSourceID, session.DataSourceType, session.DataSourceName); err == nil {
			metric.FreshnessScore = freshnessResult.OverallScore
			metric.FreshnessLevel = freshnessResult.FreshnessLevel
			if freshnessResult.CurrentFreshness != nil {
				metric.Age = freshnessResult.CurrentFreshness.Age
				metric.UpdateFrequency = freshnessResult.CurrentFreshness.UpdateFrequency
			}
		}
	}

	// Perform reliability assessment if assessor is available
	if dqm.reliabilityAssessor != nil {
		if reliabilityResult, err := dqm.reliabilityAssessor.AssessSourceReliability(ctx, session.DataSourceID, session.DataSourceType, session.DataSourceName, data); err == nil {
			metric.ReliabilityScore = reliabilityResult.OverallScore
			metric.ReliabilityLevel = reliabilityResult.ReliabilityLevel
			if reliabilityResult.PerformanceMetrics != nil {
				metric.UptimePercentage = reliabilityResult.PerformanceMetrics.UptimePercentage
				metric.ErrorRate = reliabilityResult.PerformanceMetrics.ErrorRate
			}
			if reliabilityResult.RiskAssessment != nil {
				metric.RiskLevel = reliabilityResult.RiskAssessment.RiskLevel
				metric.RiskFactors = reliabilityResult.RiskAssessment.RiskFactors
				metric.RiskScore = reliabilityResult.RiskAssessment.RiskScore
			}
		}
	}

	// Store metric
	dqm.mu.Lock()
	dqm.qualityMetrics[metricID] = metric
	dqm.mu.Unlock()

	return metric
}

func (dqm *DataQualityMonitor) checkAlerts(ctx context.Context, session *MonitoringSession, metric *QualityMetric) []*AlertRecord {
	var alerts []*AlertRecord

	// Check quality alerts
	if metric.QualityScore < dqm.config.QualityAlertThreshold {
		alert := dqm.createAlert(session.SessionID, "quality", "medium",
			fmt.Sprintf("Quality score %.2f is below threshold %.2f", metric.QualityScore, dqm.config.QualityAlertThreshold),
			dqm.config.QualityAlertThreshold, metric.QualityScore, "quality_threshold")
		alerts = append(alerts, alert)
	}

	// Check freshness alerts
	if metric.Age > dqm.config.FreshnessAlertThreshold {
		alert := dqm.createAlert(session.SessionID, "freshness", "medium",
			fmt.Sprintf("Data age %v exceeds threshold %v", metric.Age, dqm.config.FreshnessAlertThreshold),
			float64(dqm.config.FreshnessAlertThreshold), float64(metric.Age), "freshness_threshold")
		alerts = append(alerts, alert)
	}

	// Check reliability alerts
	if metric.ReliabilityScore < dqm.config.ReliabilityAlertThreshold {
		alert := dqm.createAlert(session.SessionID, "reliability", "medium",
			fmt.Sprintf("Reliability score %.2f is below threshold %.2f", metric.ReliabilityScore, dqm.config.ReliabilityAlertThreshold),
			dqm.config.ReliabilityAlertThreshold, metric.ReliabilityScore, "reliability_threshold")
		alerts = append(alerts, alert)
	}

	// Check critical alerts
	if metric.QualityScore < dqm.config.CriticalThreshold {
		alert := dqm.createAlert(session.SessionID, "critical", "critical",
			fmt.Sprintf("Critical quality issue: score %.2f is below critical threshold %.2f", metric.QualityScore, dqm.config.CriticalThreshold),
			dqm.config.CriticalThreshold, metric.QualityScore, "critical_threshold")
		alerts = append(alerts, alert)
	}

	// Store alerts
	dqm.mu.Lock()
	for _, alert := range alerts {
		dqm.alertHistory[alert.AlertID] = alert
		session.AlertCount++
	}
	dqm.mu.Unlock()

	return alerts
}

func (dqm *DataQualityMonitor) createAlert(sessionID, alertType, severity, message string, threshold, currentValue float64, triggeredBy string) *AlertRecord {
	alertID := fmt.Sprintf("alert-%s-%s-%d", sessionID, alertType, time.Now().Unix())

	return &AlertRecord{
		AlertID:         alertID,
		SessionID:       sessionID,
		AlertType:       alertType,
		Severity:        severity,
		Message:         message,
		CreatedAt:       time.Now(),
		IsActive:        true,
		EscalationLevel: 1,
		Threshold:       threshold,
		CurrentValue:    currentValue,
		TriggeredBy:     triggeredBy,
	}
}

func (dqm *DataQualityMonitor) generateRecommendations(metric *QualityMetric, alerts []*AlertRecord) []string {
	var recommendations []string

	// Quality-based recommendations
	if metric.QualityScore < 0.8 {
		recommendations = append(recommendations, "Improve overall data quality")
	}
	if metric.QualityDetails["completeness"] < 0.8 {
		recommendations = append(recommendations, "Address data completeness issues")
	}
	if metric.QualityDetails["accuracy"] < 0.8 {
		recommendations = append(recommendations, "Improve data accuracy")
	}

	// Freshness-based recommendations
	if metric.FreshnessScore < 0.7 {
		recommendations = append(recommendations, "Update data more frequently")
	}

	// Reliability-based recommendations
	if metric.ReliabilityScore < 0.8 {
		recommendations = append(recommendations, "Improve data source reliability")
	}

	// Risk-based recommendations
	if metric.RiskLevel == "high" || metric.RiskLevel == "critical" {
		recommendations = append(recommendations, "Address high-risk data quality issues")
	}

	return recommendations
}

func (dqm *DataQualityMonitor) generatePriorityActions(metric *QualityMetric, alerts []*AlertRecord) []string {
	var actions []string

	// Critical actions
	if metric.QualityScore < dqm.config.CriticalThreshold {
		actions = append(actions, "URGENT: Address critical quality issues immediately")
	}

	// Alert-based actions
	for _, alert := range alerts {
		if alert.Severity == "critical" {
			actions = append(actions, fmt.Sprintf("URGENT: Resolve critical %s alert", alert.AlertType))
		}
	}

	// Quality-based actions
	if metric.QualityDetails["completeness"] < 0.6 {
		actions = append(actions, "PRIORITY: Fix data completeness issues")
	}

	return actions
}

func (dqm *DataQualityMonitor) updateSessionMetrics(session *MonitoringSession, metric *QualityMetric, alerts []*AlertRecord) {
	dqm.mu.Lock()
	defer dqm.mu.Unlock()

	// Update overall quality score
	session.OverallQualityScore = metric.QualityScore
	session.QualityLevel = metric.QualityLevel

	// Update component scores
	session.CompletenessScore = metric.QualityDetails["completeness"]
	session.AccuracyScore = metric.QualityDetails["accuracy"]
	session.ConsistencyScore = metric.QualityDetails["consistency"]
	session.FreshnessScore = metric.FreshnessScore
	session.ReliabilityScore = metric.ReliabilityScore
	session.ValidityScore = metric.QualityDetails["validity"]

	// Update last assessment time
	session.LastAssessment = time.Now()
}

func (dqm *DataQualityMonitor) collectMetricsForTimeRange(sessionID string, timeRange TimeRange) []*QualityMetric {
	dqm.mu.RLock()
	defer dqm.mu.RUnlock()

	var metrics []*QualityMetric
	for _, metric := range dqm.qualityMetrics {
		if metric.SessionID == sessionID &&
			metric.Timestamp.After(timeRange.Start) &&
			metric.Timestamp.Before(timeRange.End) {
			metrics = append(metrics, metric)
		}
	}

	return metrics
}

func (dqm *DataQualityMonitor) generateQualitySummary(metrics []*QualityMetric, session *MonitoringSession) *QualitySummary {
	if len(metrics) == 0 {
		return &QualitySummary{
			QualityLevel:        "unknown",
			CompletenessSummary: &ComponentSummary{},
			AccuracySummary:     &ComponentSummary{},
			ConsistencySummary:  &ComponentSummary{},
			FreshnessSummary:    &ComponentSummary{},
			ReliabilitySummary:  &ComponentSummary{},
			ValiditySummary:     &ComponentSummary{},
		}
	}

	// Calculate averages
	var totalQuality, totalCompleteness, totalAccuracy, totalConsistency, totalFreshness, totalReliability, totalValidity float64
	var criticalIssues int

	for _, metric := range metrics {
		totalQuality += metric.QualityScore
		totalCompleteness += metric.QualityDetails["completeness"]
		totalAccuracy += metric.QualityDetails["accuracy"]
		totalConsistency += metric.QualityDetails["consistency"]
		totalFreshness += metric.FreshnessScore
		totalReliability += metric.ReliabilityScore
		totalValidity += metric.QualityDetails["validity"]

		if metric.QualityScore < dqm.config.CriticalThreshold {
			criticalIssues++
		}
	}

	count := float64(len(metrics))

	return &QualitySummary{
		OverallQualityScore: totalQuality / count,
		QualityLevel:        dqm.determineQualityLevel(totalQuality / count),
		AssessmentCount:     len(metrics),
		CriticalIssues:      criticalIssues,
		CompletenessSummary: &ComponentSummary{AverageScore: totalCompleteness / count},
		AccuracySummary:     &ComponentSummary{AverageScore: totalAccuracy / count},
		ConsistencySummary:  &ComponentSummary{AverageScore: totalConsistency / count},
		FreshnessSummary:    &ComponentSummary{AverageScore: totalFreshness / count},
		ReliabilitySummary:  &ComponentSummary{AverageScore: totalReliability / count},
		ValiditySummary:     &ComponentSummary{AverageScore: totalValidity / count},
	}
}

func (dqm *DataQualityMonitor) analyzeQualityTrends(metrics []*QualityMetric) []*QualityTrend {
	if len(metrics) < 2 {
		return []*QualityTrend{}
	}

	// Sort metrics by timestamp
	// This is a simplified trend analysis
	firstMetric := metrics[0]
	lastMetric := metrics[len(metrics)-1]

	trends := []*QualityTrend{
		{
			Component:  "overall",
			Direction:  dqm.determineTrendDirection(firstMetric.QualityScore, lastMetric.QualityScore),
			Slope:      lastMetric.QualityScore - firstMetric.QualityScore,
			Confidence: 0.7,
			Periods:    len(metrics),
			LastChange: lastMetric.Timestamp,
		},
	}

	return trends
}

func (dqm *DataQualityMonitor) collectAlertsForTimeRange(sessionID string, timeRange TimeRange) []*AlertRecord {
	dqm.mu.RLock()
	defer dqm.mu.RUnlock()

	var alerts []*AlertRecord
	for _, alert := range dqm.alertHistory {
		if alert.SessionID == sessionID &&
			alert.CreatedAt.After(timeRange.Start) &&
			alert.CreatedAt.Before(timeRange.End) {
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

func (dqm *DataQualityMonitor) generateAlertSummaries(alerts []*AlertRecord) []*AlertSummary {
	alertMap := make(map[string]*AlertSummary)

	for _, alert := range alerts {
		key := fmt.Sprintf("%s-%s", alert.AlertType, alert.Severity)
		if summary, exists := alertMap[key]; exists {
			summary.Count++
			if alert.IsActive {
				summary.ActiveCount++
			} else {
				summary.ResolvedCount++
			}
		} else {
			alertMap[key] = &AlertSummary{
				AlertType:     alert.AlertType,
				Severity:      alert.Severity,
				Count:         1,
				ActiveCount:   0,
				ResolvedCount: 0,
			}
			if alert.IsActive {
				alertMap[key].ActiveCount = 1
			} else {
				alertMap[key].ResolvedCount = 1
			}
		}
	}

	var summaries []*AlertSummary
	for _, summary := range alertMap {
		summaries = append(summaries, summary)
	}

	return summaries
}

func (dqm *DataQualityMonitor) generateReportRecommendations(summary *QualitySummary, trends []*QualityTrend, alerts []*AlertRecord) []string {
	var recommendations []string

	// Quality-based recommendations
	if summary.OverallQualityScore < 0.8 {
		recommendations = append(recommendations, "Implement data quality improvement initiatives")
	}

	// Trend-based recommendations
	for _, trend := range trends {
		if trend.Direction == "declining" {
			recommendations = append(recommendations, fmt.Sprintf("Address declining %s quality trend", trend.Component))
		}
	}

	// Alert-based recommendations
	if len(alerts) > 0 {
		recommendations = append(recommendations, "Review and resolve active quality alerts")
	}

	return recommendations
}

func (dqm *DataQualityMonitor) generateReportPriorityActions(summary *QualitySummary, trends []*QualityTrend, alerts []*AlertRecord) []string {
	var actions []string

	// Critical actions
	if summary.CriticalIssues > 0 {
		actions = append(actions, "URGENT: Address critical quality issues")
	}

	// Alert-based actions
	criticalAlerts := 0
	for _, alert := range alerts {
		if alert.Severity == "critical" && alert.IsActive {
			criticalAlerts++
		}
	}
	if criticalAlerts > 0 {
		actions = append(actions, fmt.Sprintf("URGENT: Resolve %d critical alerts", criticalAlerts))
	}

	return actions
}

func (dqm *DataQualityMonitor) determineQualityLevel(score float64) string {
	if score >= 0.9 {
		return "excellent"
	} else if score >= 0.8 {
		return "good"
	} else if score >= 0.7 {
		return "fair"
	} else if score >= 0.5 {
		return "poor"
	} else {
		return "critical"
	}
}

func (dqm *DataQualityMonitor) determineTrendDirection(first, last float64) string {
	diff := last - first
	if diff > 0.05 {
		return "improving"
	} else if diff < -0.05 {
		return "declining"
	} else {
		return "stable"
	}
}

func getDefaultDataQualityMonitorConfig() *DataQualityMonitorConfig {
	return &DataQualityMonitorConfig{
		// Monitoring settings
		EnableRealTimeMonitoring: true,
		EnableAlerting:           true,
		EnableReporting:          true,
		EnableTrendAnalysis:      true,

		// Thresholds
		QualityAlertThreshold:     0.7,
		FreshnessAlertThreshold:   24 * time.Hour,
		ReliabilityAlertThreshold: 0.8,
		CriticalThreshold:         0.5,

		// Reporting settings
		ReportGenerationInterval: 1 * time.Hour,
		ReportRetentionPeriod:    30 * 24 * time.Hour, // 30 days
		MaxReportsPerSession:     100,

		// Alert settings
		AlertCooldownPeriod:      1 * time.Hour,
		MaxAlertsPerSession:      50,
		AlertEscalationThreshold: 5,

		// Performance settings
		MonitoringInterval:     5 * time.Minute,
		MetricsRetentionPeriod: 7 * 24 * time.Hour, // 7 days
		CleanupInterval:        1 * time.Hour,
	}
}
