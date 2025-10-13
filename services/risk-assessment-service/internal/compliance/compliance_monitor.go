package compliance

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ComplianceMonitor implements continuous compliance monitoring and alerting
type ComplianceMonitor struct {
	logger *zap.Logger
	config *ComplianceMonitorConfig
}

// ComplianceMonitorConfig represents configuration for compliance monitoring
type ComplianceMonitorConfig struct {
	MonitoringInterval    time.Duration          `json:"monitoring_interval"`
	AlertThresholds       AlertThresholds        `json:"alert_thresholds"`
	NotificationChannels  []NotificationChannel  `json:"notification_channels"`
	EnableRealTimeAlerts  bool                   `json:"enable_real_time_alerts"`
	EnableScheduledChecks bool                   `json:"enable_scheduled_checks"`
	RetentionPeriod       time.Duration          `json:"retention_period"`
	Metadata              map[string]interface{} `json:"metadata"`
}

// AlertThresholds represents alert thresholds for compliance monitoring
type AlertThresholds struct {
	CriticalCompliancePercentage float64 `json:"critical_compliance_percentage"`
	HighCompliancePercentage     float64 `json:"high_compliance_percentage"`
	MediumCompliancePercentage   float64 `json:"medium_compliance_percentage"`
	LowCompliancePercentage      float64 `json:"low_compliance_percentage"`
	MaxValidationFailures        int     `json:"max_validation_failures"`
	MaxConsecutiveFailures       int     `json:"max_consecutive_failures"`
}

// NotificationChannel represents a notification channel
type NotificationChannel struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       NotificationType       `json:"type"`
	Enabled    bool                   `json:"enabled"`
	Config     map[string]interface{} `json:"config"`
	Severities []AlertSeverity        `json:"severities"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeEmail     NotificationType = "email"
	NotificationTypeSlack     NotificationType = "slack"
	NotificationTypeWebhook   NotificationType = "webhook"
	NotificationTypeSMS       NotificationType = "sms"
	NotificationTypePagerDuty NotificationType = "pagerduty"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityHigh     AlertSeverity = "high"
	AlertSeverityMedium   AlertSeverity = "medium"
	AlertSeverityLow      AlertSeverity = "low"
)

// ComplianceAlert represents a compliance alert
type ComplianceAlert struct {
	ID                   string                 `json:"id"`
	TenantID             string                 `json:"tenant_id"`
	Regulation           string                 `json:"regulation"`
	Category             RegulationCategory     `json:"category"`
	Severity             AlertSeverity          `json:"severity"`
	Type                 AlertType              `json:"type"`
	Title                string                 `json:"title"`
	Description          string                 `json:"description"`
	CompliancePercentage float64                `json:"compliance_percentage"`
	Threshold            float64                `json:"threshold"`
	CurrentValue         float64                `json:"current_value"`
	PreviousValue        float64                `json:"previous_value"`
	Trend                AlertTrend             `json:"trend"`
	Status               AlertStatus            `json:"status"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
	AcknowledgedAt       *time.Time             `json:"acknowledged_at,omitempty"`
	AcknowledgedBy       string                 `json:"acknowledged_by,omitempty"`
	ResolvedAt           *time.Time             `json:"resolved_at,omitempty"`
	ResolvedBy           string                 `json:"resolved_by,omitempty"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeComplianceDrop    AlertType = "compliance_drop"
	AlertTypeValidationFailure AlertType = "validation_failure"
	AlertTypeThresholdBreach   AlertType = "threshold_breach"
	AlertTypeTrendChange       AlertType = "trend_change"
	AlertTypeScheduledCheck    AlertType = "scheduled_check"
	AlertTypeManualCheck       AlertType = "manual_check"
)

// AlertTrend represents the trend of an alert
type AlertTrend string

const (
	AlertTrendImproving AlertTrend = "improving"
	AlertTrendDeclining AlertTrend = "declining"
	AlertTrendStable    AlertTrend = "stable"
	AlertTrendVolatile  AlertTrend = "volatile"
)

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "active"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved     AlertStatus = "resolved"
	AlertStatusSuppressed   AlertStatus = "suppressed"
)

// ComplianceMetrics represents compliance metrics
type ComplianceMetrics struct {
	ID                   string                 `json:"id"`
	TenantID             string                 `json:"tenant_id"`
	Regulation           string                 `json:"regulation"`
	Category             RegulationCategory     `json:"category"`
	Timestamp            time.Time              `json:"timestamp"`
	CompliancePercentage float64                `json:"compliance_percentage"`
	TotalValidations     int                    `json:"total_validations"`
	PassedValidations    int                    `json:"passed_validations"`
	FailedValidations    int                    `json:"failed_validations"`
	WarningValidations   int                    `json:"warning_validations"`
	AverageScore         float64                `json:"average_score"`
	Trend                AlertTrend             `json:"trend"`
	PreviousMetrics      *ComplianceMetrics     `json:"previous_metrics,omitempty"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// ComplianceDashboard represents a compliance dashboard
type ComplianceDashboard struct {
	ID              string                 `json:"id"`
	TenantID        string                 `json:"tenant_id"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Regulations     []string               `json:"regulations"`
	Categories      []RegulationCategory   `json:"categories"`
	TimeRange       string                 `json:"time_range"`
	RefreshInterval time.Duration          `json:"refresh_interval"`
	Widgets         []DashboardWidget      `json:"widgets"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	CreatedBy       string                 `json:"created_by"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// DashboardWidget represents a dashboard widget
type DashboardWidget struct {
	ID          string                 `json:"id"`
	Type        WidgetType             `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Position    WidgetPosition         `json:"position"`
	Size        WidgetSize             `json:"size"`
	Config      map[string]interface{} `json:"config"`
	Data        interface{}            `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// WidgetType represents the type of widget
type WidgetType string

const (
	WidgetTypeComplianceChart   WidgetType = "compliance_chart"
	WidgetTypeTrendChart        WidgetType = "trend_chart"
	WidgetTypeAlertList         WidgetType = "alert_list"
	WidgetTypeMetricsSummary    WidgetType = "metrics_summary"
	WidgetTypeRegulationStatus  WidgetType = "regulation_status"
	WidgetTypeValidationHistory WidgetType = "validation_history"
)

// WidgetPosition represents the position of a widget
type WidgetPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// WidgetSize represents the size of a widget
type WidgetSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// NewComplianceMonitor creates a new compliance monitor instance
func NewComplianceMonitor(config *ComplianceMonitorConfig, logger *zap.Logger) *ComplianceMonitor {
	return &ComplianceMonitor{
		logger: logger,
		config: config,
	}
}

// StartMonitoring starts continuous compliance monitoring
func (cm *ComplianceMonitor) StartMonitoring(ctx context.Context) error {
	cm.logger.Info("Starting compliance monitoring",
		zap.Duration("interval", cm.config.MonitoringInterval))

	ticker := time.NewTicker(cm.config.MonitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			cm.logger.Info("Compliance monitoring stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := cm.performMonitoringCycle(ctx); err != nil {
				cm.logger.Error("Monitoring cycle failed", zap.Error(err))
			}
		}
	}
}

// performMonitoringCycle performs a single monitoring cycle
func (cm *ComplianceMonitor) performMonitoringCycle(ctx context.Context) error {
	cm.logger.Debug("Performing monitoring cycle")

	// Get all active tenants
	tenants, err := cm.getActiveTenants(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active tenants: %w", err)
	}

	// Monitor each tenant
	for _, tenant := range tenants {
		if err := cm.monitorTenant(ctx, tenant); err != nil {
			cm.logger.Error("Failed to monitor tenant",
				zap.String("tenant_id", tenant),
				zap.Error(err))
		}
	}

	cm.logger.Debug("Monitoring cycle completed",
		zap.Int("tenants_monitored", len(tenants)))

	return nil
}

// monitorTenant monitors compliance for a specific tenant
func (cm *ComplianceMonitor) monitorTenant(ctx context.Context, tenantID string) error {
	// Get tenant's regulations
	regulations, err := cm.getTenantRegulations(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant regulations: %w", err)
	}

	// Monitor each regulation
	for _, regulation := range regulations {
		if err := cm.monitorRegulation(ctx, tenantID, regulation); err != nil {
			cm.logger.Error("Failed to monitor regulation",
				zap.String("tenant_id", tenantID),
				zap.String("regulation", regulation),
				zap.Error(err))
		}
	}

	return nil
}

// monitorRegulation monitors compliance for a specific regulation
func (cm *ComplianceMonitor) monitorRegulation(ctx context.Context, tenantID string, regulation string) error {
	// Get current compliance metrics
	metrics, err := cm.getComplianceMetrics(ctx, tenantID, regulation)
	if err != nil {
		return fmt.Errorf("failed to get compliance metrics: %w", err)
	}

	// Check for threshold breaches
	if err := cm.checkThresholdBreaches(ctx, tenantID, regulation, metrics); err != nil {
		return fmt.Errorf("failed to check threshold breaches: %w", err)
	}

	// Check for trend changes
	if err := cm.checkTrendChanges(ctx, tenantID, regulation, metrics); err != nil {
		return fmt.Errorf("failed to check trend changes: %w", err)
	}

	// Store metrics
	if err := cm.storeComplianceMetrics(ctx, metrics); err != nil {
		return fmt.Errorf("failed to store compliance metrics: %w", err)
	}

	return nil
}

// checkThresholdBreaches checks for threshold breaches
func (cm *ComplianceMonitor) checkThresholdBreaches(ctx context.Context, tenantID string, regulation string, metrics *ComplianceMetrics) error {
	thresholds := cm.config.AlertThresholds

	// Check compliance percentage thresholds
	if metrics.CompliancePercentage < thresholds.CriticalCompliancePercentage {
		alert := &ComplianceAlert{
			ID:                   fmt.Sprintf("alert_%d", time.Now().UnixNano()),
			TenantID:             tenantID,
			Regulation:           regulation,
			Category:             metrics.Category,
			Severity:             AlertSeverityCritical,
			Type:                 AlertTypeThresholdBreach,
			Title:                "Critical Compliance Drop",
			Description:          fmt.Sprintf("Compliance percentage dropped to %.1f%%", metrics.CompliancePercentage),
			CompliancePercentage: metrics.CompliancePercentage,
			Threshold:            thresholds.CriticalCompliancePercentage,
			CurrentValue:         metrics.CompliancePercentage,
			Status:               AlertStatusActive,
			CreatedAt:            time.Now(),
			Metadata:             make(map[string]interface{}),
		}

		if err := cm.createAlert(ctx, alert); err != nil {
			return fmt.Errorf("failed to create alert: %w", err)
		}
	}

	return nil
}

// checkTrendChanges checks for trend changes
func (cm *ComplianceMonitor) checkTrendChanges(ctx context.Context, tenantID string, regulation string, metrics *ComplianceMetrics) error {
	// Get previous metrics for trend analysis
	previousMetrics, err := cm.getPreviousMetrics(ctx, tenantID, regulation)
	if err != nil {
		// No previous metrics available, skip trend analysis
		return nil
	}

	// Calculate trend
	trend := cm.calculateTrend(metrics, previousMetrics)

	// Check for significant trend changes
	if trend == AlertTrendDeclining && metrics.CompliancePercentage < previousMetrics.CompliancePercentage-5.0 {
		alert := &ComplianceAlert{
			ID:                   fmt.Sprintf("alert_%d", time.Now().UnixNano()),
			TenantID:             tenantID,
			Regulation:           regulation,
			Category:             metrics.Category,
			Severity:             AlertSeverityHigh,
			Type:                 AlertTypeTrendChange,
			Title:                "Compliance Trend Declining",
			Description:          fmt.Sprintf("Compliance trend is declining: %.1f%% -> %.1f%%", previousMetrics.CompliancePercentage, metrics.CompliancePercentage),
			CompliancePercentage: metrics.CompliancePercentage,
			CurrentValue:         metrics.CompliancePercentage,
			PreviousValue:        previousMetrics.CompliancePercentage,
			Trend:                trend,
			Status:               AlertStatusActive,
			CreatedAt:            time.Now(),
			Metadata:             make(map[string]interface{}),
		}

		if err := cm.createAlert(ctx, alert); err != nil {
			return fmt.Errorf("failed to create alert: %w", err)
		}
	}

	return nil
}

// calculateTrend calculates the trend between current and previous metrics
func (cm *ComplianceMonitor) calculateTrend(current *ComplianceMetrics, previous *ComplianceMetrics) AlertTrend {
	diff := current.CompliancePercentage - previous.CompliancePercentage

	if diff > 2.0 {
		return AlertTrendImproving
	} else if diff < -2.0 {
		return AlertTrendDeclining
	} else {
		return AlertTrendStable
	}
}

// createAlert creates a new compliance alert
func (cm *ComplianceMonitor) createAlert(ctx context.Context, alert *ComplianceAlert) error {
	// Store alert in database
	cm.logger.Info("Compliance alert created",
		zap.String("alert_id", alert.ID),
		zap.String("tenant_id", alert.TenantID),
		zap.String("regulation", alert.Regulation),
		zap.String("severity", string(alert.Severity)),
		zap.String("type", string(alert.Type)))

	// Send notifications
	if cm.config.EnableRealTimeAlerts {
		if err := cm.sendNotifications(ctx, alert); err != nil {
			cm.logger.Error("Failed to send notifications", zap.Error(err))
		}
	}

	return nil
}

// sendNotifications sends notifications for an alert
func (cm *ComplianceMonitor) sendNotifications(ctx context.Context, alert *ComplianceAlert) error {
	for _, channel := range cm.config.NotificationChannels {
		if !channel.Enabled {
			continue
		}

		// Check if channel should receive this severity
		shouldNotify := false
		for _, severity := range channel.Severities {
			if severity == alert.Severity {
				shouldNotify = true
				break
			}
		}

		if !shouldNotify {
			continue
		}

		// Send notification based on channel type
		switch channel.Type {
		case NotificationTypeEmail:
			if err := cm.sendEmailNotification(ctx, channel, alert); err != nil {
				cm.logger.Error("Failed to send email notification", zap.Error(err))
			}
		case NotificationTypeSlack:
			if err := cm.sendSlackNotification(ctx, channel, alert); err != nil {
				cm.logger.Error("Failed to send Slack notification", zap.Error(err))
			}
		case NotificationTypeWebhook:
			if err := cm.sendWebhookNotification(ctx, channel, alert); err != nil {
				cm.logger.Error("Failed to send webhook notification", zap.Error(err))
			}
		}
	}

	return nil
}

// Notification sending methods
func (cm *ComplianceMonitor) sendEmailNotification(ctx context.Context, channel NotificationChannel, alert *ComplianceAlert) error {
	// Mock email notification
	cm.logger.Info("Email notification sent",
		zap.String("channel_id", channel.ID),
		zap.String("alert_id", alert.ID))
	return nil
}

func (cm *ComplianceMonitor) sendSlackNotification(ctx context.Context, channel NotificationChannel, alert *ComplianceAlert) error {
	// Mock Slack notification
	cm.logger.Info("Slack notification sent",
		zap.String("channel_id", channel.ID),
		zap.String("alert_id", alert.ID))
	return nil
}

func (cm *ComplianceMonitor) sendWebhookNotification(ctx context.Context, channel NotificationChannel, alert *ComplianceAlert) error {
	// Mock webhook notification
	cm.logger.Info("Webhook notification sent",
		zap.String("channel_id", channel.ID),
		zap.String("alert_id", alert.ID))
	return nil
}

// Mock data methods
func (cm *ComplianceMonitor) getActiveTenants(ctx context.Context) ([]string, error) {
	// Mock active tenants
	return []string{"tenant_1", "tenant_2", "tenant_3"}, nil
}

func (cm *ComplianceMonitor) getTenantRegulations(ctx context.Context, tenantID string) ([]string, error) {
	// Mock tenant regulations
	return []string{"BSA", "GDPR", "HIPAA"}, nil
}

func (cm *ComplianceMonitor) getComplianceMetrics(ctx context.Context, tenantID string, regulation string) (*ComplianceMetrics, error) {
	// Mock compliance metrics
	return &ComplianceMetrics{
		ID:                   fmt.Sprintf("metrics_%d", time.Now().UnixNano()),
		TenantID:             tenantID,
		Regulation:           regulation,
		Category:             RegulationCategoryPrivacy,
		Timestamp:            time.Now(),
		CompliancePercentage: 95.0,
		TotalValidations:     20,
		PassedValidations:    19,
		FailedValidations:    1,
		WarningValidations:   0,
		AverageScore:         94.5,
		Trend:                AlertTrendStable,
		Metadata:             make(map[string]interface{}),
	}, nil
}

func (cm *ComplianceMonitor) getPreviousMetrics(ctx context.Context, tenantID string, regulation string) (*ComplianceMetrics, error) {
	// Mock previous metrics
	return &ComplianceMetrics{
		ID:                   fmt.Sprintf("metrics_%d", time.Now().UnixNano()-86400),
		TenantID:             tenantID,
		Regulation:           regulation,
		Category:             RegulationCategoryPrivacy,
		Timestamp:            time.Now().AddDate(0, 0, -1),
		CompliancePercentage: 97.0,
		TotalValidations:     20,
		PassedValidations:    19,
		FailedValidations:    1,
		WarningValidations:   0,
		AverageScore:         96.5,
		Trend:                AlertTrendStable,
		Metadata:             make(map[string]interface{}),
	}, nil
}

func (cm *ComplianceMonitor) storeComplianceMetrics(ctx context.Context, metrics *ComplianceMetrics) error {
	// Mock storing metrics
	cm.logger.Debug("Compliance metrics stored",
		zap.String("tenant_id", metrics.TenantID),
		zap.String("regulation", metrics.Regulation),
		zap.Float64("compliance_percentage", metrics.CompliancePercentage))
	return nil
}
