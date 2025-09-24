package classification_monitoring

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AccuracyAlertingSystem manages accuracy-related alerts and reporting
type AccuracyAlertingSystem struct {
	config              *AlertingConfig
	logger              *zap.Logger
	mu                  sync.RWMutex
	alertRules          map[string]*AlertRule
	activeAlerts        map[string]*AccuracyAlert
	alertHistory        []*AccuracyAlert
	notificationService NotificationService
	reportGenerator     *ReportGenerator
	escalationManager   *EscalationManager
	startTime           time.Time
}

// AlertingConfig holds configuration for the alerting system
type AlertingConfig struct {
	EnableRealTimeAlerting     bool          `json:"enable_real_time_alerting"`
	EnableEscalation           bool          `json:"enable_escalation"`
	AlertCooldownPeriod        time.Duration `json:"alert_cooldown_period"`
	AlertRetentionPeriod       time.Duration `json:"alert_retention_period"`
	MaxActiveAlerts            int           `json:"max_active_alerts"`
	NotificationChannels       []string      `json:"notification_channels"`
	ReportingInterval          time.Duration `json:"reporting_interval"`
	EnableAutomaticReports     bool          `json:"enable_automatic_reports"`
	EnableDashboardIntegration bool          `json:"enable_dashboard_integration"`
	ThresholdCheckInterval     time.Duration `json:"threshold_check_interval"`
}

// AlertRule defines a rule for triggering accuracy alerts
type AlertRule struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	Description          string            `json:"description"`
	Condition            AlertCondition    `json:"condition"`
	Severity             string            `json:"severity"`
	Threshold            float64           `json:"threshold"`
	ComparisonOperator   string            `json:"comparison_operator"` // <, >, <=, >=, ==
	DimensionFilter      map[string]string `json:"dimension_filter"`
	TimeWindow           time.Duration     `json:"time_window"`
	MinSampleSize        int               `json:"min_sample_size"`
	Enabled              bool              `json:"enabled"`
	NotificationChannels []string          `json:"notification_channels"`
	EscalationPolicy     *EscalationPolicy `json:"escalation_policy"`
	Actions              []AlertAction     `json:"actions"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
}

// AlertCondition defines what triggers an alert
type AlertCondition struct {
	MetricType      string            `json:"metric_type"` // accuracy, error_rate, confidence, trend
	DimensionName   string            `json:"dimension_name"`
	DimensionValue  string            `json:"dimension_value"`
	AggregationType string            `json:"aggregation_type"` // avg, min, max, count, rate
	Filters         map[string]string `json:"filters"`
}

// AlertAction defines actions to take when an alert is triggered
type AlertAction struct {
	Type       string                 `json:"type"` // notify, escalate, auto_remediate, create_ticket
	Parameters map[string]interface{} `json:"parameters"`
	Enabled    bool                   `json:"enabled"`
}

// EscalationPolicy defines escalation rules for alerts
type EscalationPolicy struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Levels    []EscalationLevel `json:"levels"`
	MaxLevels int               `json:"max_levels"`
	Enabled   bool              `json:"enabled"`
}

// EscalationLevel represents a level in the escalation chain
type EscalationLevel struct {
	Level         int           `json:"level"`
	Delay         time.Duration `json:"delay"`
	Recipients    []string      `json:"recipients"`
	Channels      []string      `json:"channels"`
	StopCondition string        `json:"stop_condition"`
}

// NotificationService interface for sending notifications
type NotificationService interface {
	SendNotification(ctx context.Context, alert *AccuracyAlert, channels []string) error
	GetSupportedChannels() []string
}

// ReportGenerator generates accuracy reports
type ReportGenerator struct {
	config *ReportConfig
	logger *zap.Logger
}

// ReportConfig holds configuration for report generation
type ReportConfig struct {
	EnableScheduledReports bool              `json:"enable_scheduled_reports"`
	ReportFormats          []string          `json:"report_formats"` // json, html, pdf
	ReportDistribution     []string          `json:"report_distribution"`
	IncludeVisualization   bool              `json:"include_visualization"`
	DetailLevel            string            `json:"detail_level"` // summary, detailed, comprehensive
	CustomTemplates        map[string]string `json:"custom_templates"`
}

// AccuracyReport represents a comprehensive accuracy report
type AccuracyReport struct {
	ID                  string                      `json:"id"`
	ReportType          string                      `json:"report_type"`
	GeneratedAt         time.Time                   `json:"generated_at"`
	Period              ReportPeriod                `json:"period"`
	OverallSummary      *AccuracySummary            `json:"overall_summary"`
	DimensionalAnalysis map[string]*AccuracySummary `json:"dimensional_analysis"`
	TrendAnalysis       *TrendAnalysis              `json:"trend_analysis"`
	AlertSummary        *AlertSummary               `json:"alert_summary"`
	Recommendations     []Recommendation            `json:"recommendations"`
	Metadata            map[string]interface{}      `json:"metadata"`
}

// ReportPeriod defines the time period for a report
type ReportPeriod struct {
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
	Label     string        `json:"label"`
}

// AccuracySummary provides a summary of accuracy metrics
type AccuracySummary struct {
	AccuracyRate           float64            `json:"accuracy_rate"`
	ErrorRate              float64            `json:"error_rate"`
	TotalClassifications   int                `json:"total_classifications"`
	CorrectClassifications int                `json:"correct_classifications"`
	AverageConfidence      float64            `json:"average_confidence"`
	ConfidenceDistribution map[string]int     `json:"confidence_distribution"`
	MethodDistribution     map[string]int     `json:"method_distribution"`
	ErrorDistribution      map[string]int     `json:"error_distribution"`
	QualityScore           float64            `json:"quality_score"`
	ComparisonToPrevious   *ComparisonMetrics `json:"comparison_to_previous"`
}

// AlertSummary provides a summary of alerts during the report period
type AlertSummary struct {
	TotalAlerts           int            `json:"total_alerts"`
	AlertsBySeverity      map[string]int `json:"alerts_by_severity"`
	AlertsByType          map[string]int `json:"alerts_by_type"`
	ResolvedAlerts        int            `json:"resolved_alerts"`
	ActiveAlerts          int            `json:"active_alerts"`
	AverageResolutionTime time.Duration  `json:"average_resolution_time"`
	TopAlertRules         []string       `json:"top_alert_rules"`
}

// Recommendation represents a recommendation for improvement
type Recommendation struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Priority    string                 `json:"priority"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Actions     []string               `json:"actions"`
	Impact      string                 `json:"impact"`
	Effort      string                 `json:"effort"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// EscalationManager manages alert escalations
type EscalationManager struct {
	config            *AlertingConfig
	logger            *zap.Logger
	activeEscalations map[string]*EscalationInstance
	mu                sync.RWMutex
}

// EscalationInstance represents an active escalation
type EscalationInstance struct {
	AlertID       string           `json:"alert_id"`
	PolicyID      string           `json:"policy_id"`
	CurrentLevel  int              `json:"current_level"`
	StartedAt     time.Time        `json:"started_at"`
	LastEscalated time.Time        `json:"last_escalated"`
	Completed     bool             `json:"completed"`
	History       []EscalationStep `json:"history"`
}

// EscalationStep represents a step in the escalation process
type EscalationStep struct {
	Level        int       `json:"level"`
	Timestamp    time.Time `json:"timestamp"`
	Recipients   []string  `json:"recipients"`
	Channels     []string  `json:"channels"`
	Success      bool      `json:"success"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// NewAccuracyAlertingSystem creates a new alerting system
func NewAccuracyAlertingSystem(config *AlertingConfig, logger *zap.Logger, notificationService NotificationService) *AccuracyAlertingSystem {
	if config == nil {
		config = DefaultAlertingConfig()
	}

	system := &AccuracyAlertingSystem{
		config:              config,
		logger:              logger,
		alertRules:          make(map[string]*AlertRule),
		activeAlerts:        make(map[string]*AccuracyAlert),
		alertHistory:        make([]*AccuracyAlert, 0),
		notificationService: notificationService,
		startTime:           time.Now(),
	}

	// Initialize report generator
	system.reportGenerator = NewReportGenerator(&ReportConfig{
		EnableScheduledReports: config.EnableAutomaticReports,
		ReportFormats:          []string{"json", "html"},
		DetailLevel:            "detailed",
	}, logger)

	// Initialize escalation manager
	if config.EnableEscalation {
		system.escalationManager = NewEscalationManager(config, logger)
	}

	// Initialize default alert rules
	system.initializeDefaultAlertRules()

	return system
}

// DefaultAlertingConfig returns default alerting configuration
func DefaultAlertingConfig() *AlertingConfig {
	return &AlertingConfig{
		EnableRealTimeAlerting:     true,
		EnableEscalation:           true,
		AlertCooldownPeriod:        15 * time.Minute,
		AlertRetentionPeriod:       30 * 24 * time.Hour, // 30 days
		MaxActiveAlerts:            100,
		NotificationChannels:       []string{"email", "slack"},
		ReportingInterval:          24 * time.Hour,
		EnableAutomaticReports:     true,
		EnableDashboardIntegration: true,
		ThresholdCheckInterval:     5 * time.Minute,
	}
}

// initializeDefaultAlertRules sets up default alert rules
func (aas *AccuracyAlertingSystem) initializeDefaultAlertRules() {
	defaultRules := []*AlertRule{
		{
			ID:          "low_overall_accuracy",
			Name:        "Low Overall Accuracy",
			Description: "Overall accuracy has fallen below acceptable threshold",
			Condition: AlertCondition{
				MetricType:      "accuracy",
				DimensionName:   "overall",
				DimensionValue:  "all",
				AggregationType: "avg",
			},
			Severity:             "high",
			Threshold:            0.85,
			ComparisonOperator:   "<",
			TimeWindow:           1 * time.Hour,
			MinSampleSize:        50,
			Enabled:              true,
			NotificationChannels: []string{"email", "slack"},
			Actions: []AlertAction{
				{
					Type:    "notify",
					Enabled: true,
					Parameters: map[string]interface{}{
						"urgent": true,
					},
				},
			},
			CreatedAt: time.Now(),
		},
		{
			ID:          "critical_accuracy_drop",
			Name:        "Critical Accuracy Drop",
			Description: "Accuracy has dropped critically low",
			Condition: AlertCondition{
				MetricType:      "accuracy",
				DimensionName:   "overall",
				DimensionValue:  "all",
				AggregationType: "avg",
			},
			Severity:             "critical",
			Threshold:            0.75,
			ComparisonOperator:   "<",
			TimeWindow:           30 * time.Minute,
			MinSampleSize:        20,
			Enabled:              true,
			NotificationChannels: []string{"email", "slack", "pagerduty"},
			EscalationPolicy: &EscalationPolicy{
				ID:   "critical_escalation",
				Name: "Critical Issue Escalation",
				Levels: []EscalationLevel{
					{
						Level:      1,
						Delay:      5 * time.Minute,
						Recipients: []string{"team_lead"},
						Channels:   []string{"slack"},
					},
					{
						Level:      2,
						Delay:      15 * time.Minute,
						Recipients: []string{"engineering_manager"},
						Channels:   []string{"email", "phone"},
					},
				},
				MaxLevels: 2,
				Enabled:   true,
			},
			Actions: []AlertAction{
				{
					Type:    "escalate",
					Enabled: true,
				},
			},
			CreatedAt: time.Now(),
		},
		{
			ID:          "high_confidence_errors",
			Name:        "High Confidence Misclassifications",
			Description: "High number of misclassifications with high confidence",
			Condition: AlertCondition{
				MetricType:      "error_rate",
				DimensionName:   "confidence_range",
				DimensionValue:  "very_high",
				AggregationType: "rate",
			},
			Severity:             "medium",
			Threshold:            0.05, // 5% error rate for high confidence
			ComparisonOperator:   ">",
			TimeWindow:           2 * time.Hour,
			MinSampleSize:        30,
			Enabled:              true,
			NotificationChannels: []string{"slack"},
			Actions: []AlertAction{
				{
					Type:    "notify",
					Enabled: true,
				},
			},
			CreatedAt: time.Now(),
		},
		{
			ID:          "method_performance_degradation",
			Name:        "Classification Method Performance Degradation",
			Description: "Specific classification method showing poor performance",
			Condition: AlertCondition{
				MetricType:      "accuracy",
				DimensionName:   "method",
				AggregationType: "avg",
			},
			Severity:             "medium",
			Threshold:            0.80,
			ComparisonOperator:   "<",
			TimeWindow:           4 * time.Hour,
			MinSampleSize:        25,
			Enabled:              true,
			NotificationChannels: []string{"email"},
			Actions: []AlertAction{
				{
					Type:    "notify",
					Enabled: true,
				},
			},
			CreatedAt: time.Now(),
		},
	}

	for _, rule := range defaultRules {
		aas.alertRules[rule.ID] = rule
	}
}

// EvaluateAlerts evaluates all alert rules against current metrics
func (aas *AccuracyAlertingSystem) EvaluateAlerts(ctx context.Context, metrics *MetricsAggregationResult) error {
	aas.mu.Lock()
	defer aas.mu.Unlock()

	aas.logger.Info("Evaluating alert rules", zap.Int("rule_count", len(aas.alertRules)))

	for ruleID, rule := range aas.alertRules {
		if !rule.Enabled {
			continue
		}

		shouldAlert, value, err := aas.evaluateRule(rule, metrics)
		if err != nil {
			aas.logger.Error("Failed to evaluate alert rule",
				zap.String("rule_id", ruleID),
				zap.Error(err))
			continue
		}

		if shouldAlert {
			alert := aas.createAlert(rule, value, metrics.EndTime)
			if err := aas.triggerAlert(ctx, alert); err != nil {
				aas.logger.Error("Failed to trigger alert",
					zap.String("alert_id", alert.ID),
					zap.Error(err))
			}
		}
	}

	// Check for alert resolutions
	aas.checkAlertResolutions(ctx, metrics)

	return nil
}

// evaluateRule evaluates a single alert rule
func (aas *AccuracyAlertingSystem) evaluateRule(rule *AlertRule, metrics *MetricsAggregationResult) (bool, float64, error) {
	var value float64
	var sampleSize int

	// Get the relevant metrics based on the rule condition
	if rule.Condition.DimensionName == "overall" {
		if metrics.OverallMetrics == nil {
			return false, 0, fmt.Errorf("overall metrics not available")
		}

		switch rule.Condition.MetricType {
		case "accuracy":
			value = metrics.OverallMetrics.AccuracyRate
		case "error_rate":
			value = metrics.OverallMetrics.ErrorRate
		case "confidence":
			value = metrics.OverallMetrics.AverageConfidence
		default:
			return false, 0, fmt.Errorf("unsupported metric type: %s", rule.Condition.MetricType)
		}

		sampleSize = metrics.OverallMetrics.TotalClassifications
	} else {
		// Look for dimensional metrics
		dimensionKey := fmt.Sprintf("%s:%s", rule.Condition.DimensionName, rule.Condition.DimensionValue)
		if rule.Condition.DimensionValue == "" {
			// Find the worst performing dimension value
			worstValue := 1.0
			worstKey := ""

			for key, dimMetrics := range metrics.DimensionalMetrics {
				if !strings.HasPrefix(key, rule.Condition.DimensionName+":") {
					continue
				}

				var dimValue float64
				switch rule.Condition.MetricType {
				case "accuracy":
					dimValue = dimMetrics.AccuracyRate
				case "error_rate":
					dimValue = dimMetrics.ErrorRate
				case "confidence":
					dimValue = dimMetrics.AverageConfidence
				default:
					continue
				}

				if (rule.ComparisonOperator == "<" || rule.ComparisonOperator == "<=") && dimValue < worstValue {
					worstValue = dimValue
					worstKey = key
				} else if (rule.ComparisonOperator == ">" || rule.ComparisonOperator == ">=") && dimValue > worstValue {
					worstValue = dimValue
					worstKey = key
				}
			}

			if worstKey != "" {
				value = worstValue
				sampleSize = metrics.DimensionalMetrics[worstKey].TotalClassifications
				// Update rule to specify the dimension value that triggered
				rule.Condition.DimensionValue = strings.Split(worstKey, ":")[1]
			}
		} else {
			dimMetrics, exists := metrics.DimensionalMetrics[dimensionKey]
			if !exists {
				return false, 0, nil // No data for this dimension
			}

			switch rule.Condition.MetricType {
			case "accuracy":
				value = dimMetrics.AccuracyRate
			case "error_rate":
				value = dimMetrics.ErrorRate
			case "confidence":
				value = dimMetrics.AverageConfidence
			default:
				return false, 0, fmt.Errorf("unsupported metric type: %s", rule.Condition.MetricType)
			}

			sampleSize = dimMetrics.TotalClassifications
		}
	}

	// Check sample size requirement
	if sampleSize < rule.MinSampleSize {
		return false, 0, nil
	}

	// Evaluate the condition
	switch rule.ComparisonOperator {
	case "<":
		return value < rule.Threshold, value, nil
	case "<=":
		return value <= rule.Threshold, value, nil
	case ">":
		return value > rule.Threshold, value, nil
	case ">=":
		return value >= rule.Threshold, value, nil
	case "==":
		return value == rule.Threshold, value, nil
	default:
		return false, 0, fmt.Errorf("unsupported comparison operator: %s", rule.ComparisonOperator)
	}
}

// createAlert creates a new alert from a rule and current value
func (aas *AccuracyAlertingSystem) createAlert(rule *AlertRule, currentValue float64, timestamp time.Time) *AccuracyAlert {
	alert := &AccuracyAlert{
		ID:             fmt.Sprintf("alert_%s_%d", rule.ID, timestamp.UnixNano()),
		Type:           "threshold",
		Severity:       rule.Severity,
		DimensionName:  rule.Condition.DimensionName,
		DimensionValue: rule.Condition.DimensionValue,
		CurrentValue:   currentValue,
		ThresholdValue: rule.Threshold,
		Message:        aas.generateAlertMessage(rule, currentValue),
		Timestamp:      timestamp,
		Actions:        aas.generateAlertActions(rule),
		Resolved:       false,
	}

	return alert
}

// generateAlertMessage generates a human-readable alert message
func (aas *AccuracyAlertingSystem) generateAlertMessage(rule *AlertRule, currentValue float64) string {
	dimensionStr := ""
	if rule.Condition.DimensionName != "overall" {
		dimensionStr = fmt.Sprintf(" for %s:%s", rule.Condition.DimensionName, rule.Condition.DimensionValue)
	}

	metricName := rule.Condition.MetricType
	if metricName == "accuracy" || metricName == "error_rate" {
		currentValue *= 100 // Convert to percentage
		threshold := rule.Threshold * 100
		return fmt.Sprintf("%s%s is %.2f%% (threshold: %.2f%%)", metricName, dimensionStr, currentValue, threshold)
	}

	return fmt.Sprintf("%s%s is %.3f (threshold: %.3f)", metricName, dimensionStr, currentValue, rule.Threshold)
}

// generateAlertActions generates alert actions based on rule configuration
func (aas *AccuracyAlertingSystem) generateAlertActions(rule *AlertRule) []string {
	actions := make([]string, 0)

	for _, action := range rule.Actions {
		if !action.Enabled {
			continue
		}

		switch action.Type {
		case "notify":
			actions = append(actions, "Send notifications to configured channels")
		case "escalate":
			actions = append(actions, "Escalate according to escalation policy")
		case "auto_remediate":
			actions = append(actions, "Attempt automatic remediation")
		case "create_ticket":
			actions = append(actions, "Create support ticket")
		}
	}

	// Add generic recommendations
	switch rule.Condition.MetricType {
	case "accuracy":
		actions = append(actions, "Review recent classification changes", "Analyze misclassification patterns", "Check training data quality")
	case "error_rate":
		actions = append(actions, "Investigate error causes", "Review input data quality", "Check system health")
	case "confidence":
		actions = append(actions, "Review model confidence calibration", "Analyze feature importance", "Check for data drift")
	}

	return actions
}

// triggerAlert triggers an alert and performs associated actions
func (aas *AccuracyAlertingSystem) triggerAlert(ctx context.Context, alert *AccuracyAlert) error {
	// Check for cooldown period
	if aas.isInCooldownPeriod(alert) {
		return nil
	}

	// Add to active alerts
	aas.activeAlerts[alert.ID] = alert
	aas.alertHistory = append(aas.alertHistory, alert)

	// Send notifications
	if aas.notificationService != nil {
		rule := aas.alertRules[strings.Split(alert.ID, "_")[1]] // Extract rule ID
		if rule != nil {
			if err := aas.notificationService.SendNotification(ctx, alert, rule.NotificationChannels); err != nil {
				aas.logger.Error("Failed to send notification", zap.String("alert_id", alert.ID), zap.Error(err))
			}
		}
	}

	// Start escalation if configured
	if aas.config.EnableEscalation && aas.escalationManager != nil {
		rule := aas.alertRules[strings.Split(alert.ID, "_")[1]]
		if rule != nil && rule.EscalationPolicy != nil {
			if err := aas.escalationManager.StartEscalation(ctx, alert, rule.EscalationPolicy); err != nil {
				aas.logger.Error("Failed to start escalation", zap.String("alert_id", alert.ID), zap.Error(err))
			}
		}
	}

	// Limit active alerts
	if len(aas.activeAlerts) > aas.config.MaxActiveAlerts {
		aas.cleanupOldestAlerts()
	}

	aas.logger.Warn("Alert triggered",
		zap.String("alert_id", alert.ID),
		zap.String("severity", alert.Severity),
		zap.String("dimension", fmt.Sprintf("%s:%s", alert.DimensionName, alert.DimensionValue)),
		zap.Float64("current_value", alert.CurrentValue),
		zap.Float64("threshold", alert.ThresholdValue),
		zap.String("message", alert.Message))

	return nil
}

// isInCooldownPeriod checks if an alert is in cooldown period
func (aas *AccuracyAlertingSystem) isInCooldownPeriod(alert *AccuracyAlert) bool {
	cutoffTime := time.Now().Add(-aas.config.AlertCooldownPeriod)

	for _, existingAlert := range aas.activeAlerts {
		if existingAlert.DimensionName == alert.DimensionName &&
			existingAlert.DimensionValue == alert.DimensionValue &&
			existingAlert.Timestamp.After(cutoffTime) &&
			!existingAlert.Resolved {
			return true
		}
	}

	return false
}

// checkAlertResolutions checks if any active alerts should be resolved
func (aas *AccuracyAlertingSystem) checkAlertResolutions(ctx context.Context, metrics *MetricsAggregationResult) {
	for alertID, alert := range aas.activeAlerts {
		if alert.Resolved {
			continue
		}

		// Check if the alert condition is no longer met
		rule := aas.findRuleForAlert(alert)
		if rule == nil {
			continue
		}

		shouldAlert, _, err := aas.evaluateRule(rule, metrics)
		if err != nil {
			continue
		}

		if !shouldAlert {
			aas.resolveAlert(ctx, alertID)
		}
	}
}

// findRuleForAlert finds the rule that generated an alert
func (aas *AccuracyAlertingSystem) findRuleForAlert(alert *AccuracyAlert) *AlertRule {
	for _, rule := range aas.alertRules {
		if rule.Condition.DimensionName == alert.DimensionName &&
			rule.Condition.DimensionValue == alert.DimensionValue {
			return rule
		}
	}
	return nil
}

// resolveAlert resolves an active alert
func (aas *AccuracyAlertingSystem) resolveAlert(ctx context.Context, alertID string) {
	alert, exists := aas.activeAlerts[alertID]
	if !exists {
		return
	}

	alert.Resolved = true
	now := time.Now()
	alert.ResolvedAt = &now

	// Stop escalation if active
	if aas.escalationManager != nil {
		aas.escalationManager.StopEscalation(alertID)
	}

	// Send resolution notification
	if aas.notificationService != nil {
		// Create a resolution alert for notification
		resolutionAlert := &AccuracyAlert{
			ID:             fmt.Sprintf("resolved_%s", alertID),
			Type:           "resolution",
			Severity:       "info",
			DimensionName:  alert.DimensionName,
			DimensionValue: alert.DimensionValue,
			Message:        fmt.Sprintf("Alert resolved: %s", alert.Message),
			Timestamp:      now,
		}

		aas.notificationService.SendNotification(ctx, resolutionAlert, []string{"slack"})
	}

	aas.logger.Info("Alert resolved",
		zap.String("alert_id", alertID),
		zap.Duration("duration", now.Sub(alert.Timestamp)))
}

// cleanupOldestAlerts removes the oldest alerts to maintain the limit
func (aas *AccuracyAlertingSystem) cleanupOldestAlerts() {
	if len(aas.activeAlerts) <= aas.config.MaxActiveAlerts {
		return
	}

	// Convert to slice for sorting
	alerts := make([]*AccuracyAlert, 0, len(aas.activeAlerts))
	for _, alert := range aas.activeAlerts {
		alerts = append(alerts, alert)
	}

	// Sort by timestamp (oldest first)
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].Timestamp.Before(alerts[j].Timestamp)
	})

	// Remove oldest alerts
	removeCount := len(alerts) - aas.config.MaxActiveAlerts + 10 // Remove extra to avoid frequent cleanup
	for i := 0; i < removeCount && i < len(alerts); i++ {
		if alerts[i].Resolved {
			delete(aas.activeAlerts, alerts[i].ID)
		}
	}
}

// GenerateReport generates a comprehensive accuracy report
func (aas *AccuracyAlertingSystem) GenerateReport(ctx context.Context, period ReportPeriod, metrics *MetricsAggregationResult) (*AccuracyReport, error) {
	if aas.reportGenerator == nil {
		return nil, fmt.Errorf("report generator not initialized")
	}

	report := &AccuracyReport{
		ID:          fmt.Sprintf("report_%d", time.Now().UnixNano()),
		ReportType:  "accuracy_analysis",
		GeneratedAt: time.Now(),
		Period:      period,
		Metadata:    make(map[string]interface{}),
	}

	// Generate overall summary
	if metrics.OverallMetrics != nil {
		report.OverallSummary = &AccuracySummary{
			AccuracyRate:           metrics.OverallMetrics.AccuracyRate,
			ErrorRate:              metrics.OverallMetrics.ErrorRate,
			TotalClassifications:   metrics.OverallMetrics.TotalClassifications,
			CorrectClassifications: metrics.OverallMetrics.CorrectClassifications,
			AverageConfidence:      metrics.OverallMetrics.AverageConfidence,
			QualityScore:           metrics.OverallMetrics.DataQualityScore,
			ComparisonToPrevious:   metrics.ComparisonMetrics,
		}
	}

	// Generate dimensional analysis
	report.DimensionalAnalysis = make(map[string]*AccuracySummary)
	for dimensionKey, dimMetrics := range metrics.DimensionalMetrics {
		summary := &AccuracySummary{
			AccuracyRate:           dimMetrics.AccuracyRate,
			ErrorRate:              dimMetrics.ErrorRate,
			TotalClassifications:   dimMetrics.TotalClassifications,
			CorrectClassifications: dimMetrics.CorrectClassifications,
			AverageConfidence:      dimMetrics.AverageConfidence,
			QualityScore:           dimMetrics.DataQualityScore,
		}
		report.DimensionalAnalysis[dimensionKey] = summary
	}

	// Add trend analysis
	report.TrendAnalysis = metrics.TrendAnalysis

	// Generate alert summary
	report.AlertSummary = aas.generateAlertSummary(period)

	// Generate recommendations
	report.Recommendations = aas.generateRecommendations(metrics)

	return report, nil
}

// generateAlertSummary generates a summary of alerts for the report period
func (aas *AccuracyAlertingSystem) generateAlertSummary(period ReportPeriod) *AlertSummary {
	summary := &AlertSummary{
		AlertsBySeverity: make(map[string]int),
		AlertsByType:     make(map[string]int),
		TopAlertRules:    make([]string, 0),
	}

	var totalResolutionTime time.Duration
	resolvedCount := 0
	ruleFrequency := make(map[string]int)

	for _, alert := range aas.alertHistory {
		if alert.Timestamp.Before(period.StartTime) || alert.Timestamp.After(period.EndTime) {
			continue
		}

		summary.TotalAlerts++
		summary.AlertsBySeverity[alert.Severity]++
		summary.AlertsByType[alert.Type]++

		if alert.Resolved {
			summary.ResolvedAlerts++
			resolvedCount++
			if alert.ResolvedAt != nil {
				totalResolutionTime += alert.ResolvedAt.Sub(alert.Timestamp)
			}
		} else {
			summary.ActiveAlerts++
		}

		// Track rule frequency
		ruleID := strings.Split(alert.ID, "_")[1]
		ruleFrequency[ruleID]++
	}

	if resolvedCount > 0 {
		summary.AverageResolutionTime = totalResolutionTime / time.Duration(resolvedCount)
	}

	// Find top alert rules
	type ruleFreq struct {
		ruleID string
		count  int
	}
	var ruleFreqs []ruleFreq
	for ruleID, count := range ruleFrequency {
		ruleFreqs = append(ruleFreqs, ruleFreq{ruleID, count})
	}
	sort.Slice(ruleFreqs, func(i, j int) bool {
		return ruleFreqs[i].count > ruleFreqs[j].count
	})

	for i, rf := range ruleFreqs {
		if i >= 5 { // Top 5
			break
		}
		summary.TopAlertRules = append(summary.TopAlertRules, rf.ruleID)
	}

	return summary
}

// generateRecommendations generates recommendations based on metrics and alerts
func (aas *AccuracyAlertingSystem) generateRecommendations(metrics *MetricsAggregationResult) []Recommendation {
	recommendations := make([]Recommendation, 0)

	// Check overall accuracy
	if metrics.OverallMetrics != nil && metrics.OverallMetrics.AccuracyRate < 0.90 {
		recommendations = append(recommendations, Recommendation{
			ID:          "improve_overall_accuracy",
			Type:        "performance",
			Priority:    "high",
			Title:       "Improve Overall Accuracy",
			Description: fmt.Sprintf("Overall accuracy is %.2f%%, below the target of 90%%", metrics.OverallMetrics.AccuracyRate*100),
			Actions:     []string{"Review training data quality", "Analyze misclassification patterns", "Consider model retraining"},
			Impact:      "high",
			Effort:      "medium",
		})
	}

	// Check for high error rates in specific dimensions
	for dimensionKey, dimMetrics := range metrics.DimensionalMetrics {
		if dimMetrics.ErrorRate > 0.15 { // 15% error rate
			parts := strings.Split(dimensionKey, ":")
			if len(parts) == 2 {
				recommendations = append(recommendations, Recommendation{
					ID:          fmt.Sprintf("fix_%s_errors", strings.Replace(dimensionKey, ":", "_", -1)),
					Type:        "dimension_specific",
					Priority:    "medium",
					Title:       fmt.Sprintf("Address High Error Rate in %s", parts[0]),
					Description: fmt.Sprintf("Error rate for %s is %.2f%%", dimensionKey, dimMetrics.ErrorRate*100),
					Actions:     []string{fmt.Sprintf("Analyze %s-specific patterns", parts[0]), "Review feature engineering", "Consider specialized training"},
					Impact:      "medium",
					Effort:      "low",
				})
			}
		}
	}

	// Check for data quality issues
	if metrics.QualityAssessment != nil && metrics.QualityAssessment.OverallQualityScore < 0.8 {
		recommendations = append(recommendations, Recommendation{
			ID:          "improve_data_quality",
			Type:        "data_quality",
			Priority:    "medium",
			Title:       "Improve Data Quality",
			Description: fmt.Sprintf("Overall data quality score is %.2f", metrics.QualityAssessment.OverallQualityScore),
			Actions:     metrics.QualityAssessment.Recommendations,
			Impact:      "high",
			Effort:      "medium",
		})
	}

	return recommendations
}

// Public methods for managing the alerting system

// AddAlertRule adds a new alert rule
func (aas *AccuracyAlertingSystem) AddAlertRule(rule *AlertRule) error {
	aas.mu.Lock()
	defer aas.mu.Unlock()

	if rule.ID == "" {
		return fmt.Errorf("alert rule ID cannot be empty")
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	aas.alertRules[rule.ID] = rule

	aas.logger.Info("Alert rule added", zap.String("rule_id", rule.ID), zap.String("name", rule.Name))
	return nil
}

// UpdateAlertRule updates an existing alert rule
func (aas *AccuracyAlertingSystem) UpdateAlertRule(ruleID string, rule *AlertRule) error {
	aas.mu.Lock()
	defer aas.mu.Unlock()

	if _, exists := aas.alertRules[ruleID]; !exists {
		return fmt.Errorf("alert rule not found: %s", ruleID)
	}

	rule.ID = ruleID
	rule.UpdatedAt = time.Now()
	aas.alertRules[ruleID] = rule

	aas.logger.Info("Alert rule updated", zap.String("rule_id", ruleID))
	return nil
}

// DeleteAlertRule deletes an alert rule
func (aas *AccuracyAlertingSystem) DeleteAlertRule(ruleID string) error {
	aas.mu.Lock()
	defer aas.mu.Unlock()

	if _, exists := aas.alertRules[ruleID]; !exists {
		return fmt.Errorf("alert rule not found: %s", ruleID)
	}

	delete(aas.alertRules, ruleID)
	aas.logger.Info("Alert rule deleted", zap.String("rule_id", ruleID))
	return nil
}

// GetActiveAlerts returns currently active alerts
func (aas *AccuracyAlertingSystem) GetActiveAlerts() []*AccuracyAlert {
	aas.mu.RLock()
	defer aas.mu.RUnlock()

	alerts := make([]*AccuracyAlert, 0, len(aas.activeAlerts))
	for _, alert := range aas.activeAlerts {
		if !alert.Resolved {
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// GetAlertHistory returns alert history
func (aas *AccuracyAlertingSystem) GetAlertHistory(limit int) []*AccuracyAlert {
	aas.mu.RLock()
	defer aas.mu.RUnlock()

	if limit <= 0 || limit > len(aas.alertHistory) {
		limit = len(aas.alertHistory)
	}

	// Return most recent alerts
	start := len(aas.alertHistory) - limit
	result := make([]*AccuracyAlert, limit)
	copy(result, aas.alertHistory[start:])

	return result
}

// GetAlertRules returns all alert rules
func (aas *AccuracyAlertingSystem) GetAlertRules() map[string]*AlertRule {
	aas.mu.RLock()
	defer aas.mu.RUnlock()

	result := make(map[string]*AlertRule)
	for k, v := range aas.alertRules {
		result[k] = v
	}
	return result
}

// NewReportGenerator creates a new report generator
func NewReportGenerator(config *ReportConfig, logger *zap.Logger) *ReportGenerator {
	return &ReportGenerator{
		config: config,
		logger: logger,
	}
}

// NewEscalationManager creates a new escalation manager
func NewEscalationManager(config *AlertingConfig, logger *zap.Logger) *EscalationManager {
	return &EscalationManager{
		config:            config,
		logger:            logger,
		activeEscalations: make(map[string]*EscalationInstance),
	}
}

// StartEscalation starts an escalation process for an alert
func (em *EscalationManager) StartEscalation(ctx context.Context, alert *AccuracyAlert, policy *EscalationPolicy) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if !policy.Enabled {
		return nil
	}

	instance := &EscalationInstance{
		AlertID:      alert.ID,
		PolicyID:     policy.ID,
		CurrentLevel: 0,
		StartedAt:    time.Now(),
		History:      make([]EscalationStep, 0),
	}

	em.activeEscalations[alert.ID] = instance

	// Start escalation in background
	go em.runEscalation(ctx, instance, policy)

	return nil
}

// StopEscalation stops an active escalation
func (em *EscalationManager) StopEscalation(alertID string) {
	em.mu.Lock()
	defer em.mu.Unlock()

	if instance, exists := em.activeEscalations[alertID]; exists {
		instance.Completed = true
		delete(em.activeEscalations, alertID)
		em.logger.Info("Escalation stopped", zap.String("alert_id", alertID))
	}
}

// runEscalation runs the escalation process
func (em *EscalationManager) runEscalation(ctx context.Context, instance *EscalationInstance, policy *EscalationPolicy) {
	for level := 0; level < len(policy.Levels) && level < policy.MaxLevels; level++ {
		// Check if escalation should continue
		em.mu.RLock()
		if instance.Completed {
			em.mu.RUnlock()
			return
		}
		em.mu.RUnlock()

		escalationLevel := policy.Levels[level]

		// Wait for delay if not the first level
		if level > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(escalationLevel.Delay):
			}
		}

		// Execute escalation step
		step := EscalationStep{
			Level:      escalationLevel.Level,
			Timestamp:  time.Now(),
			Recipients: escalationLevel.Recipients,
			Channels:   escalationLevel.Channels,
			Success:    true, // Simplified - assume success
		}

		em.mu.Lock()
		instance.CurrentLevel = level + 1
		instance.LastEscalated = step.Timestamp
		instance.History = append(instance.History, step)
		em.mu.Unlock()

		em.logger.Info("Escalation step executed",
			zap.String("alert_id", instance.AlertID),
			zap.Int("level", escalationLevel.Level),
			zap.Strings("recipients", escalationLevel.Recipients))
	}

	// Mark escalation as completed
	em.mu.Lock()
	instance.Completed = true
	em.mu.Unlock()
}
