package risk

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// RiskAlertSystem manages risk alerts and notifications
type RiskAlertSystem struct {
	logger        *zap.Logger
	config        *AlertSystemConfig
	thresholds    *AlertThresholdManager
	notifications *NotificationService
	alertStore    AlertStore
}

// AlertSystemConfig contains configuration for the alert system
type AlertSystemConfig struct {
	EnableRealTimeAlerts     bool          `json:"enable_real_time_alerts"`
	EnableBatchAlerts        bool          `json:"enable_batch_alerts"`
	EnableEscalationAlerts   bool          `json:"enable_escalation_alerts"`
	AlertCooldownPeriod      time.Duration `json:"alert_cooldown_period"`
	MaxAlertsPerHour         int           `json:"max_alerts_per_hour"`
	AlertRetentionDays       int           `json:"alert_retention_days"`
	NotificationChannels     []string      `json:"notification_channels"`
	DefaultAlertLevel        RiskLevel     `json:"default_alert_level"`
	EnableAlertAggregation   bool          `json:"enable_alert_aggregation"`
	AggregationWindowMinutes int           `json:"aggregation_window_minutes"`
}

// AlertStore interface for storing and retrieving alerts
type AlertStore interface {
	StoreAlert(ctx context.Context, alert *RiskAlert) error
	GetAlerts(ctx context.Context, businessID string, filters AlertFilters) ([]RiskAlert, error)
	GetActiveAlerts(ctx context.Context, businessID string) ([]RiskAlert, error)
	UpdateAlertStatus(ctx context.Context, alertID string, status AlertStatus) error
	DeleteOldAlerts(ctx context.Context, olderThan time.Time) error
}

// AlertFilters contains filters for querying alerts
type AlertFilters struct {
	BusinessID string       `json:"business_id,omitempty"`
	FactorID   string       `json:"factor_id,omitempty"`
	Category   RiskCategory `json:"category,omitempty"`
	Level      RiskLevel    `json:"level,omitempty"`
	Status     AlertStatus  `json:"status,omitempty"`
	StartDate  time.Time    `json:"start_date,omitempty"`
	EndDate    time.Time    `json:"end_date,omitempty"`
	Limit      int          `json:"limit,omitempty"`
	Offset     int          `json:"offset,omitempty"`
}

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "active"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved     AlertStatus = "resolved"
	AlertStatusSuppressed   AlertStatus = "suppressed"
	AlertStatusExpired      AlertStatus = "expired"
)

// RiskAlert represents a risk alert
// Note: This type is already defined in models.go
// type RiskAlert struct {
//	ID                   string                 `json:"id"`
//	BusinessID           string                 `json:"business_id"`
//	FactorID             string                 `json:"factor_id"`
//	FactorName           string                 `json:"factor_name"`
//	Category             RiskCategory           `json:"category"`
//	AlertType            AlertType              `json:"alert_type"`
//	Level                RiskLevel              `json:"level"`
//	Status               AlertStatus            `json:"status"`
//	Title                string                 `json:"title"`
//	Message              string                 `json:"message"`
//	CurrentValue         float64                `json:"current_value"`
//	ThresholdValue       float64                `json:"threshold_value"`
//	Severity             AlertSeverity          `json:"severity"`
//	Priority             AlertPriority          `json:"priority"`
//	Source               string                 `json:"source"`
//	TriggeredAt          time.Time              `json:"triggered_at"`
//	AcknowledgedAt       *time.Time             `json:"acknowledged_at,omitempty"`
//	ResolvedAt           *time.Time             `json:"resolved_at,omitempty"`
//	ExpiresAt            *time.Time             `json:"expires_at,omitempty"`
//	EscalatedAt          *time.Time             `json:"escalated_at,omitempty"`
//	EscalationLevel      int                    `json:"escalation_level"`
//	NotificationSent     bool                   `json:"notification_sent"`
//	NotificationChannels []string               `json:"notification_channels"`
//	Actions              []AlertAction          `json:"actions,omitempty"`
//	Metadata             map[string]interface{} `json:"metadata,omitempty"`
//	RelatedAlerts        []string               `json:"related_alerts,omitempty"`
//	Tags                 []string               `json:"tags,omitempty"`
// }

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeThreshold   AlertType = "threshold"
	AlertTypeTrend       AlertType = "trend"
	AlertTypeAnomaly     AlertType = "anomaly"
	AlertTypeCompliance  AlertType = "compliance"
	AlertTypeDeadline    AlertType = "deadline"
	AlertTypeCorrelation AlertType = "correlation"
	AlertTypeEscalation  AlertType = "escalation"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "low"
	AlertSeverityMedium   AlertSeverity = "medium"
	AlertSeverityHigh     AlertSeverity = "high"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertPriority represents the priority of an alert
type AlertPriority string

const (
	AlertPriorityLow    AlertPriority = "low"
	AlertPriorityMedium AlertPriority = "medium"
	AlertPriorityHigh   AlertPriority = "high"
	AlertPriorityUrgent AlertPriority = "urgent"
)

// AlertAction represents an action to be taken for an alert
type AlertAction struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Assignee    string                 `json:"assignee,omitempty"`
	DueDate     time.Time              `json:"due_date,omitempty"`
	Status      string                 `json:"status"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// NewRiskAlertSystem creates a new risk alert system
func NewRiskAlertSystem(logger *zap.Logger, config *AlertSystemConfig, alertStore AlertStore) *RiskAlertSystem {
	return &RiskAlertSystem{
		logger:        logger,
		config:        config,
		thresholds:    NewAlertThresholdManager(logger),
		notifications: NewNotificationService(logger),
		alertStore:    alertStore,
	}
}

// EvaluateRiskForAlerts evaluates risk data and generates alerts
// CheckAndTriggerAlerts checks risk factors and triggers alerts if thresholds are exceeded
func (ras *RiskAlertSystem) CheckAndTriggerAlerts(ctx context.Context, factors []RiskFactorDetail) ([]AlertDetail, error) {
	var alerts []AlertDetail

	for _, factor := range factors {
		// Check if factor exceeds threshold
		if factor.Score > 0.7 { // Example threshold - should be configurable
			alert := AlertDetail{
				ID:           fmt.Sprintf("alert_%s_%d", factor.FactorType, time.Now().Unix()),
				BusinessID:   "", // This should be passed from context or factor
				AlertType:    "threshold_exceeded",
				Severity:     AlertSeverityHigh,
				Title:        fmt.Sprintf("High Risk: %s", factor.FactorType),
				Description:  factor.Description,
				RiskFactor:   factor.FactorType,
				Threshold:    0.7,
				CurrentValue: factor.Score,
				Status:       AlertStatusActive,
				CreatedAt:    time.Now(),
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

func (ras *RiskAlertSystem) EvaluateRiskForAlerts(ctx context.Context, riskData *RiskAssessment) error {
	ras.logger.Info("Evaluating risk for alerts",
		zap.String("business_id", riskData.BusinessID),
		zap.Float64("overall_score", riskData.OverallScore))

	// Check for overall risk alerts
	if err := ras.checkOverallRiskAlerts(ctx, riskData); err != nil {
		ras.logger.Warn("Failed to check overall risk alerts", zap.Error(err))
	}

	// Check for factor-specific alerts
	for _, factorScore := range riskData.FactorScores {
		if err := ras.checkFactorAlerts(ctx, riskData.BusinessID, factorScore); err != nil {
			ras.logger.Warn("Failed to check factor alerts",
				zap.String("factor_id", factorScore.FactorID),
				zap.Error(err))
		}
	}

	// Check for category-specific alerts
	if err := ras.checkCategoryAlerts(ctx, riskData); err != nil {
		ras.logger.Warn("Failed to check category alerts", zap.Error(err))
	}

	// Check for trend-based alerts
	if err := ras.checkTrendAlerts(ctx, riskData); err != nil {
		ras.logger.Warn("Failed to check trend alerts", zap.Error(err))
	}

	return nil
}

// checkOverallRiskAlerts checks for overall risk level alerts
func (ras *RiskAlertSystem) checkOverallRiskAlerts(ctx context.Context, riskData *RiskAssessment) error {
	// Get thresholds for overall risk
	thresholds, err := ras.thresholds.GetThresholds("overall_risk", riskData.BusinessID)
	if err != nil {
		return fmt.Errorf("failed to get overall risk thresholds: %w", err)
	}

	// Check if overall risk exceeds thresholds
	for _, threshold := range thresholds {
		if ras.exceedsThreshold(riskData.OverallScore, threshold) {
			alert, err := ras.createAlert(ctx, AlertTypeThreshold, "overall_risk", "Overall Risk",
				riskData.BusinessID, riskData.OverallScore, threshold, riskData.OverallLevel)
			if err != nil {
				return fmt.Errorf("failed to create overall risk alert: %w", err)
			}

			if err := ras.processAlert(ctx, alert); err != nil {
				return fmt.Errorf("failed to process overall risk alert: %w", err)
			}
		}
	}

	return nil
}

// checkFactorAlerts checks for factor-specific alerts
func (ras *RiskAlertSystem) checkFactorAlerts(ctx context.Context, businessID string, factorScore RiskScore) error {
	// Get thresholds for this factor
	thresholds, err := ras.thresholds.GetThresholds(factorScore.FactorID, businessID)
	if err != nil {
		return fmt.Errorf("failed to get factor thresholds: %w", err)
	}

	// Check if factor score exceeds thresholds
	for _, threshold := range thresholds {
		if ras.exceedsThreshold(factorScore.Score, threshold) {
			alert, err := ras.createAlert(ctx, AlertTypeThreshold, factorScore.FactorID, factorScore.FactorName,
				businessID, factorScore.Score, threshold, factorScore.Level)
			if err != nil {
				return fmt.Errorf("failed to create factor alert: %w", err)
			}

			if err := ras.processAlert(ctx, alert); err != nil {
				return fmt.Errorf("failed to process factor alert: %w", err)
			}
		}
	}

	return nil
}

// checkCategoryAlerts checks for category-specific alerts
func (ras *RiskAlertSystem) checkCategoryAlerts(ctx context.Context, riskData *RiskAssessment) error {
	// Check each category for alerts
	for category, categoryScore := range riskData.CategoryScores {
		// Get thresholds for this category
		thresholds, err := ras.thresholds.GetThresholds(string(category), riskData.BusinessID)
		if err != nil {
			continue // Skip if no thresholds defined
		}

		// Check if category score exceeds thresholds
		for _, threshold := range thresholds {
			if ras.exceedsThreshold(categoryScore.Score, threshold) {
				alert, err := ras.createAlert(ctx, AlertTypeThreshold, string(category), string(category),
					riskData.BusinessID, categoryScore.Score, threshold, categoryScore.Level)
				if err != nil {
					ras.logger.Warn("Failed to create category alert",
						zap.String("category", string(category)),
						zap.Error(err))
					continue
				}

				if err := ras.processAlert(ctx, alert); err != nil {
					ras.logger.Warn("Failed to process category alert",
						zap.String("category", string(category)),
						zap.Error(err))
				}
			}
		}
	}

	return nil
}

// checkTrendAlerts checks for trend-based alerts
func (ras *RiskAlertSystem) checkTrendAlerts(ctx context.Context, riskData *RiskAssessment) error {
	// This would require historical data to detect trends
	// For now, we'll implement basic trend detection based on current data

	// Check for rapid changes in risk levels
	for _, factorScore := range riskData.FactorScores {
		// Get recent alerts for this factor to detect rapid changes
		recentAlerts, err := ras.alertStore.GetAlerts(ctx, riskData.BusinessID, AlertFilters{
			FactorID:  factorScore.FactorID,
			StartDate: time.Now().Add(-24 * time.Hour),
			Status:    AlertStatusActive,
		})
		if err != nil {
			continue
		}

		// If there are multiple recent alerts for the same factor, create a trend alert
		if len(recentAlerts) >= 3 {
			alert, err := ras.createTrendAlert(ctx, factorScore, recentAlerts)
			if err != nil {
				ras.logger.Warn("Failed to create trend alert",
					zap.String("factor_id", factorScore.FactorID),
					zap.Error(err))
				continue
			}

			if err := ras.processAlert(ctx, alert); err != nil {
				ras.logger.Warn("Failed to process trend alert",
					zap.String("factor_id", factorScore.FactorID),
					zap.Error(err))
			}
		}
	}

	return nil
}

// exceedsThreshold checks if a value exceeds a threshold
func (ras *RiskAlertSystem) exceedsThreshold(value float64, threshold AlertThreshold) bool {
	switch threshold.Operator {
	case ">":
		return value > threshold.Value
	case ">=":
		return value >= threshold.Value
	case "<":
		return value < threshold.Value
	case "<=":
		return value <= threshold.Value
	case "==":
		return value == threshold.Value
	case "!=":
		return value != threshold.Value
	default:
		return false
	}
}

// createAlert creates a new alert
func (ras *RiskAlertSystem) createAlert(ctx context.Context, alertType AlertType, factorID, factorName, businessID string,
	currentValue float64, threshold AlertThreshold, level RiskLevel) (*RiskAlert, error) {

	alertID := fmt.Sprintf("alert_%s_%s_%d", businessID, factorID, time.Now().Unix())

	// Determine severity and priority
	severity := ras.determineSeverity(level, threshold)
	_ = ras.determinePriority(severity, alertType) // Priority not used in current RiskAlert type

	// Create alert message
	message := ras.createAlertMessage(factorName, currentValue, threshold, level)

	// Note: Expiration time not used in current RiskAlert type

	alert := &RiskAlert{
		ID:             alertID,
		BusinessID:     businessID,
		RiskFactor:     factorID,
		Level:          level,
		Message:        message,
		Score:          currentValue,
		Threshold:      threshold.Value,
		TriggeredAt:    time.Now(),
		Acknowledged:   false,
		AcknowledgedAt: nil,
	}

	return alert, nil
}

// createTrendAlert creates a trend-based alert
func (ras *RiskAlertSystem) createTrendAlert(ctx context.Context, factorScore RiskScore, recentAlerts []RiskAlert) (*RiskAlert, error) {
	alertID := fmt.Sprintf("trend_alert_%s_%d", factorScore.FactorID, time.Now().Unix())

	// Calculate trend severity based on number of recent alerts
	severity := AlertSeverityMedium
	if len(recentAlerts) >= 5 {
		severity = AlertSeverityHigh
	}
	if len(recentAlerts) >= 10 {
		severity = AlertSeverityCritical
	}

	_ = ras.determinePriority(severity, AlertTypeTrend) // Priority not used in current RiskAlert type

	message := fmt.Sprintf("Multiple alerts detected for %s in the last 24 hours (%d alerts). This indicates a deteriorating trend that requires immediate attention.",
		factorScore.FactorName, len(recentAlerts))

	alert := &RiskAlert{
		ID:             alertID,
		BusinessID:     recentAlerts[0].BusinessID,
		RiskFactor:     factorScore.FactorID,
		Level:          factorScore.Level,
		Message:        message,
		Score:          factorScore.Score,
		Threshold:      0.0, // No specific threshold for trend alerts
		TriggeredAt:    time.Now(),
		Acknowledged:   false,
		AcknowledgedAt: nil,
	}

	// Note: RelatedAlerts field is not available in the current RiskAlert type

	return alert, nil
}

// processAlert processes a new alert
func (ras *RiskAlertSystem) processAlert(ctx context.Context, alert *RiskAlert) error {
	// Check for cooldown period
	if ras.isInCooldown(ctx, alert) {
		ras.logger.Info("Alert in cooldown period, skipping",
			zap.String("alert_id", alert.ID),
			zap.String("risk_factor", alert.RiskFactor))
		return nil
	}

	// Check rate limits
	if ras.exceedsRateLimit(ctx, alert.BusinessID) {
		ras.logger.Warn("Rate limit exceeded for business",
			zap.String("business_id", alert.BusinessID))
		return nil
	}

	// Store alert
	if err := ras.alertStore.StoreAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to store alert: %w", err)
	}

	// Send notifications
	if err := ras.sendNotifications(ctx, alert); err != nil {
		ras.logger.Warn("Failed to send notifications",
			zap.String("alert_id", alert.ID),
			zap.Error(err))
	}

	// Check for escalation
	if ras.shouldEscalate(alert) {
		if err := ras.escalateAlert(ctx, alert); err != nil {
			ras.logger.Warn("Failed to escalate alert",
				zap.String("alert_id", alert.ID),
				zap.Error(err))
		}
	}

	ras.logger.Info("Alert processed successfully",
		zap.String("alert_id", alert.ID),
		zap.String("risk_factor", alert.RiskFactor),
		zap.String("level", string(alert.Level)))

	return nil
}

// determineSeverity determines alert severity based on risk level and threshold
func (ras *RiskAlertSystem) determineSeverity(level RiskLevel, threshold AlertThreshold) AlertSeverity {
	// Base severity on risk level
	switch level {
	case RiskLevelCritical:
		return AlertSeverityCritical
	case RiskLevelHigh:
		return AlertSeverityHigh
	case RiskLevelMedium:
		return AlertSeverityMedium
	case RiskLevelLow:
		return AlertSeverityLow
	default:
		return AlertSeverityMedium
	}
}

// determinePriority determines alert priority based on severity and type
func (ras *RiskAlertSystem) determinePriority(severity AlertSeverity, alertType AlertType) AlertPriority {
	// Base priority on severity
	switch severity {
	case AlertSeverityCritical:
		return AlertPriorityUrgent
	case AlertSeverityHigh:
		return AlertPriorityHigh
	case AlertSeverityMedium:
		return AlertPriorityMedium
	case AlertSeverityLow:
		return AlertPriorityLow
	default:
		return AlertPriorityMedium
	}
}

// createAlertTitle creates an alert title
func (ras *RiskAlertSystem) createAlertTitle(factorName string, level RiskLevel) string {
	return fmt.Sprintf("%s Risk Alert: %s", strings.Title(string(level)), factorName)
}

// createAlertMessage creates an alert message
func (ras *RiskAlertSystem) createAlertMessage(factorName string, currentValue float64, threshold AlertThreshold, level RiskLevel) string {
	return fmt.Sprintf("Risk factor '%s' has reached %s level with a score of %.2f, exceeding the threshold of %.2f (%s). Immediate attention is required to address this risk.",
		factorName, string(level), currentValue, threshold.Value, threshold.Operator)
}

// isInCooldown checks if an alert is in cooldown period
func (ras *RiskAlertSystem) isInCooldown(ctx context.Context, alert *RiskAlert) bool {
	if ras.config.AlertCooldownPeriod <= 0 {
		return false
	}

	// Check for recent alerts for the same factor
	recentAlerts, err := ras.alertStore.GetAlerts(ctx, alert.BusinessID, AlertFilters{
		FactorID:  alert.RiskFactor,
		StartDate: time.Now().Add(-ras.config.AlertCooldownPeriod),
		Status:    AlertStatusActive,
	})
	if err != nil {
		return false
	}

	return len(recentAlerts) > 0
}

// exceedsRateLimit checks if rate limit is exceeded
func (ras *RiskAlertSystem) exceedsRateLimit(ctx context.Context, businessID string) bool {
	if ras.config.MaxAlertsPerHour <= 0 {
		return false
	}

	// Check alerts in the last hour
	recentAlerts, err := ras.alertStore.GetAlerts(ctx, businessID, AlertFilters{
		StartDate: time.Now().Add(-1 * time.Hour),
	})
	if err != nil {
		return false
	}

	return len(recentAlerts) >= ras.config.MaxAlertsPerHour
}

// shouldEscalate determines if an alert should be escalated
func (ras *RiskAlertSystem) shouldEscalate(alert *RiskAlert) bool {
	if !ras.config.EnableEscalationAlerts {
		return false
	}

	// Escalate critical alerts immediately
	if alert.Level == RiskLevelCritical {
		return true
	}

	// Escalate high severity alerts that haven't been acknowledged
	if alert.Level == RiskLevelHigh && !alert.Acknowledged {
		// Check if alert is older than escalation threshold
		escalationThreshold := time.Now().Add(-2 * time.Hour) // 2 hours
		return alert.TriggeredAt.Before(escalationThreshold)
	}

	return false
}

// escalateAlert escalates an alert
func (ras *RiskAlertSystem) escalateAlert(ctx context.Context, alert *RiskAlert) error {
	// Note: EscalationLevel and EscalatedAt fields are not available in the current RiskAlert type
	// For now, we'll just log the escalation and update the alert status

	// Update alert in store
	if err := ras.alertStore.UpdateAlertStatus(ctx, alert.ID, AlertStatusActive); err != nil {
		return fmt.Errorf("failed to update escalated alert: %w", err)
	}

	// Send escalation notifications
	if err := ras.sendNotifications(ctx, alert); err != nil {
		return fmt.Errorf("failed to send escalation notifications: %w", err)
	}

	ras.logger.Info("Alert escalated",
		zap.String("alert_id", alert.ID),
		zap.String("risk_factor", alert.RiskFactor))

	return nil
}

// sendNotifications sends notifications for an alert
func (ras *RiskAlertSystem) sendNotifications(ctx context.Context, alert *RiskAlert) error {
	// Note: NotificationSent and NotificationChannels fields are not available in the current RiskAlert type
	// For now, we'll use the default notification channels from config

	// Determine notification channels
	channels := ras.config.NotificationChannels

	// Send notifications through each channel
	for _, channel := range channels {
		if err := ras.notifications.SendNotification(ctx, channel, alert); err != nil {
			ras.logger.Warn("Failed to send notification",
				zap.String("alert_id", alert.ID),
				zap.String("channel", channel),
				zap.Error(err))
		}
	}

	return nil
}

// GetActiveAlerts retrieves active alerts for a business
func (ras *RiskAlertSystem) GetActiveAlerts(ctx context.Context, businessID string) ([]RiskAlert, error) {
	return ras.alertStore.GetActiveAlerts(ctx, businessID)
}

// AcknowledgeAlert acknowledges an alert
func (ras *RiskAlertSystem) AcknowledgeAlert(ctx context.Context, alertID string, userID string) error {
	// Update alert status
	if err := ras.alertStore.UpdateAlertStatus(ctx, alertID, AlertStatusAcknowledged); err != nil {
		return fmt.Errorf("failed to acknowledge alert: %w", err)
	}

	ras.logger.Info("Alert acknowledged",
		zap.String("alert_id", alertID),
		zap.String("user_id", userID))

	return nil
}

// ResolveAlert resolves an alert
func (ras *RiskAlertSystem) ResolveAlert(ctx context.Context, alertID string, userID string, resolution string) error {
	// Update alert status
	if err := ras.alertStore.UpdateAlertStatus(ctx, alertID, AlertStatusResolved); err != nil {
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	ras.logger.Info("Alert resolved",
		zap.String("alert_id", alertID),
		zap.String("user_id", userID),
		zap.String("resolution", resolution))

	return nil
}

// CleanupOldAlerts removes old alerts beyond retention period
func (ras *RiskAlertSystem) CleanupOldAlerts(ctx context.Context) error {
	cutoffDate := time.Now().AddDate(0, 0, -ras.config.AlertRetentionDays)
	return ras.alertStore.DeleteOldAlerts(ctx, cutoffDate)
}

// GetNotificationService returns the notification service
func (ras *RiskAlertSystem) GetNotificationService() *NotificationService {
	return ras.notifications
}
