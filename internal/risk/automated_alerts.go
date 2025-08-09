package risk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// AutomatedAlertService provides comprehensive automated risk alerting functionality
type AutomatedAlertService struct {
	logger                *observability.Logger
	alertRules            map[string]*AutomatedAlertRule
	notificationProviders map[string]NotificationProvider
	alertHistory          map[string][]AutomatedAlert
	alertQueue            chan AutomatedAlert
	processingWorkers     int
	enabled               bool
	mutex                 sync.RWMutex
}

// AutomatedAlertRule represents a rule for automated alert generation
type AutomatedAlertRule struct {
	ID                   string                    `json:"id"`
	Name                 string                    `json:"name"`
	Description          string                    `json:"description"`
	Category             RiskCategory              `json:"category"`
	FactorID             string                    `json:"factor_id,omitempty"`
	TriggerCondition     AutomatedTriggerCondition `json:"trigger_condition"`
	Threshold            float64                   `json:"threshold"`
	Level                RiskLevel                 `json:"level"`
	Message              string                    `json:"message"`
	NotificationChannels []string                  `json:"notification_channels"`
	Recipients           []string                  `json:"recipients"`
	Template             string                    `json:"template"`
	Enabled              bool                      `json:"enabled"`
	CooldownPeriod       time.Duration             `json:"cooldown_period"`
	EscalationRules      []EscalationRule          `json:"escalation_rules"`
	Metadata             map[string]interface{}    `json:"metadata,omitempty"`
	CreatedAt            time.Time                 `json:"created_at"`
	UpdatedAt            time.Time                 `json:"updated_at"`
}

// AutomatedTriggerCondition represents the condition for triggering an automated alert
type AutomatedTriggerCondition string

const (
	AutomatedTriggerConditionThresholdExceeded AutomatedTriggerCondition = "threshold_exceeded"
	AutomatedTriggerConditionRapidIncrease     AutomatedTriggerCondition = "rapid_increase"
	AutomatedTriggerConditionTrending          AutomatedTriggerCondition = "trending"
	AutomatedTriggerConditionAnomaly           AutomatedTriggerCondition = "anomaly"
	AutomatedTriggerConditionVolatility        AutomatedTriggerCondition = "volatility"
	AutomatedTriggerConditionPattern           AutomatedTriggerCondition = "pattern"
	AutomatedTriggerConditionCombination       AutomatedTriggerCondition = "combination"
)

// AutomatedAlert represents an automated alert
type AutomatedAlert struct {
	ID              string                 `json:"id"`
	RuleID          string                 `json:"rule_id"`
	BusinessID      string                 `json:"business_id"`
	Category        RiskCategory           `json:"category"`
	FactorID        string                 `json:"factor_id,omitempty"`
	Level           RiskLevel              `json:"level"`
	Message         string                 `json:"message"`
	CurrentValue    float64                `json:"current_value"`
	ThresholdValue  float64                `json:"threshold_value"`
	ExceededBy      float64                `json:"exceeded_by"`
	TriggeredAt     time.Time              `json:"triggered_at"`
	Acknowledged    bool                   `json:"acknowledged"`
	AcknowledgedAt  *time.Time             `json:"acknowledged_at,omitempty"`
	Notifications   []AlertNotification    `json:"notifications"`
	EscalationLevel int                    `json:"escalation_level"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// EscalationRule represents a rule for alert escalation
type EscalationRule struct {
	ID                   string                 `json:"id"`
	Level                int                    `json:"level"`
	Delay                time.Duration          `json:"delay"`
	AdditionalRecipients []string               `json:"additional_recipients"`
	Channels             []string               `json:"channels"`
	Message              string                 `json:"message"`
	Enabled              bool                   `json:"enabled"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// NotificationProvider represents a notification service provider
type NotificationProvider interface {
	SendNotification(ctx context.Context, notification AlertNotification) error
	GetProviderName() string
	IsAvailable() bool
}

// NewAutomatedAlertService creates a new automated alert service
func NewAutomatedAlertService(logger *observability.Logger) *AutomatedAlertService {
	service := &AutomatedAlertService{
		logger:                logger,
		alertRules:            make(map[string]*AutomatedAlertRule),
		notificationProviders: make(map[string]NotificationProvider),
		alertHistory:          make(map[string][]AutomatedAlert),
		alertQueue:            make(chan AutomatedAlert, 1000),
		processingWorkers:     5,
		enabled:               true,
	}

	// Start alert processing workers
	for i := 0; i < service.processingWorkers; i++ {
		go service.processAlertWorker(i)
	}

	// Initialize default alert rules
	service.initializeDefaultRules()

	return service
}

// ProcessAssessment processes a risk assessment and generates automated alerts
func (s *AutomatedAlertService) ProcessAssessment(ctx context.Context, assessment *RiskAssessment) ([]AutomatedAlert, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Processing assessment for automated alerts",
		"request_id", requestID,
		"business_id", assessment.BusinessID,
		"overall_score", assessment.OverallScore,
	)

	var alerts []AutomatedAlert

	// Process overall risk alerts
	overallAlerts, err := s.processOverallRiskAlerts(ctx, assessment)
	if err != nil {
		s.logger.Error("Failed to process overall risk alerts",
			"request_id", requestID,
			"business_id", assessment.BusinessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to process overall risk alerts: %w", err)
	}
	alerts = append(alerts, overallAlerts...)

	// Process category-specific alerts
	categoryAlerts, err := s.processCategoryAlerts(ctx, assessment)
	if err != nil {
		s.logger.Error("Failed to process category alerts",
			"request_id", requestID,
			"business_id", assessment.BusinessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to process category alerts: %w", err)
	}
	alerts = append(alerts, categoryAlerts...)

	// Process factor-specific alerts
	factorAlerts, err := s.processFactorAlerts(ctx, assessment)
	if err != nil {
		s.logger.Error("Failed to process factor alerts",
			"request_id", requestID,
			"business_id", assessment.BusinessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to process factor alerts: %w", err)
	}
	alerts = append(alerts, factorAlerts...)

	// Process trend-based alerts
	trendAlerts, err := s.processTrendAlerts(ctx, assessment)
	if err != nil {
		s.logger.Error("Failed to process trend alerts",
			"request_id", requestID,
			"business_id", assessment.BusinessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to process trend alerts: %w", err)
	}
	alerts = append(alerts, trendAlerts...)

	// Process anomaly alerts
	anomalyAlerts, err := s.processAnomalyAlerts(ctx, assessment)
	if err != nil {
		s.logger.Error("Failed to process anomaly alerts",
			"request_id", requestID,
			"business_id", assessment.BusinessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to process anomaly alerts: %w", err)
	}
	alerts = append(alerts, anomalyAlerts...)

	// Queue alerts for processing
	for _, alert := range alerts {
		select {
		case s.alertQueue <- alert:
			s.logger.Debug("Alert queued for processing",
				"request_id", requestID,
				"alert_id", alert.ID,
				"rule_id", alert.RuleID,
			)
		default:
			s.logger.Warn("Alert queue full, dropping alert",
				"request_id", requestID,
				"alert_id", alert.ID,
			)
		}
	}

	// Store alerts in history
	s.storeAlertHistory(assessment.BusinessID, alerts)

	s.logger.Info("Automated alert processing completed",
		"request_id", requestID,
		"business_id", assessment.BusinessID,
		"total_alerts", len(alerts),
	)

	return alerts, nil
}

// processOverallRiskAlerts processes overall risk alerts
func (s *AutomatedAlertService) processOverallRiskAlerts(ctx context.Context, assessment *RiskAssessment) ([]AutomatedAlert, error) {
	var alerts []AutomatedAlert

	// Check for critical overall risk
	if assessment.OverallScore >= 85 {
		rule := s.getAlertRule("overall_critical_risk")
		if rule != nil && rule.Enabled {
			alert := AutomatedAlert{
				ID:              fmt.Sprintf("auto_alert_%s_overall_critical", assessment.ID),
				RuleID:          rule.ID,
				BusinessID:      assessment.BusinessID,
				Category:        RiskCategoryOperational,
				Level:           RiskLevelCritical,
				Message:         fmt.Sprintf("Critical overall risk detected: %.1f", assessment.OverallScore),
				CurrentValue:    assessment.OverallScore,
				ThresholdValue:  85.0,
				ExceededBy:      assessment.OverallScore - 85.0,
				TriggeredAt:     time.Now(),
				Acknowledged:    false,
				EscalationLevel: 1,
			}
			alerts = append(alerts, alert)
		}
	}

	// Check for high overall risk
	if assessment.OverallScore >= 75 && assessment.OverallScore < 85 {
		rule := s.getAlertRule("overall_high_risk")
		if rule != nil && rule.Enabled {
			alert := AutomatedAlert{
				ID:              fmt.Sprintf("auto_alert_%s_overall_high", assessment.ID),
				RuleID:          rule.ID,
				BusinessID:      assessment.BusinessID,
				Category:        RiskCategoryOperational,
				Level:           RiskLevelHigh,
				Message:         fmt.Sprintf("High overall risk detected: %.1f", assessment.OverallScore),
				CurrentValue:    assessment.OverallScore,
				ThresholdValue:  75.0,
				ExceededBy:      assessment.OverallScore - 75.0,
				TriggeredAt:     time.Now(),
				Acknowledged:    false,
				EscalationLevel: 1,
			}
			alerts = append(alerts, alert)
		}
	}

	// Check for rapid risk increase
	if assessment.OverallScore > 60 {
		highRiskFactors := 0
		for _, factor := range assessment.FactorScores {
			if factor.Score > 70 {
				highRiskFactors++
			}
		}

		if highRiskFactors >= 3 {
			rule := s.getAlertRule("rapid_risk_increase")
			if rule != nil && rule.Enabled {
				alert := AutomatedAlert{
					ID:              fmt.Sprintf("auto_alert_%s_rapid_increase", assessment.ID),
					RuleID:          rule.ID,
					BusinessID:      assessment.BusinessID,
					Category:        RiskCategoryOperational,
					Level:           RiskLevelHigh,
					Message:         fmt.Sprintf("Rapid risk increase detected: %d factors above 70", highRiskFactors),
					CurrentValue:    assessment.OverallScore,
					ThresholdValue:  60.0,
					ExceededBy:      assessment.OverallScore - 60.0,
					TriggeredAt:     time.Now(),
					Acknowledged:    false,
					EscalationLevel: 1,
				}
				alerts = append(alerts, alert)
			}
		}
	}

	return alerts, nil
}

// processCategoryAlerts processes category-specific alerts
func (s *AutomatedAlertService) processCategoryAlerts(ctx context.Context, assessment *RiskAssessment) ([]AutomatedAlert, error) {
	var alerts []AutomatedAlert

	for category, score := range assessment.CategoryScores {
		// Check for critical category risk
		if score.Score >= 90 {
			rule := s.getAlertRule(fmt.Sprintf("%s_critical_risk", category))
			if rule != nil && rule.Enabled {
				alert := AutomatedAlert{
					ID:              fmt.Sprintf("auto_alert_%s_%s_critical", assessment.ID, category),
					RuleID:          rule.ID,
					BusinessID:      assessment.BusinessID,
					Category:        category,
					Level:           RiskLevelCritical,
					Message:         fmt.Sprintf("Critical %s risk detected: %.1f", category, score.Score),
					CurrentValue:    score.Score,
					ThresholdValue:  90.0,
					ExceededBy:      score.Score - 90.0,
					TriggeredAt:     time.Now(),
					Acknowledged:    false,
					EscalationLevel: 1,
				}
				alerts = append(alerts, alert)
			}
		}

		// Check for high category risk
		if score.Score >= 80 && score.Score < 90 {
			rule := s.getAlertRule(fmt.Sprintf("%s_high_risk", category))
			if rule != nil && rule.Enabled {
				alert := AutomatedAlert{
					ID:              fmt.Sprintf("auto_alert_%s_%s_high", assessment.ID, category),
					RuleID:          rule.ID,
					BusinessID:      assessment.BusinessID,
					Category:        category,
					Level:           RiskLevelHigh,
					Message:         fmt.Sprintf("High %s risk detected: %.1f", category, score.Score),
					CurrentValue:    score.Score,
					ThresholdValue:  80.0,
					ExceededBy:      score.Score - 80.0,
					TriggeredAt:     time.Now(),
					Acknowledged:    false,
					EscalationLevel: 1,
				}
				alerts = append(alerts, alert)
			}
		}
	}

	return alerts, nil
}

// processFactorAlerts processes factor-specific alerts
func (s *AutomatedAlertService) processFactorAlerts(ctx context.Context, assessment *RiskAssessment) ([]AutomatedAlert, error) {
	var alerts []AutomatedAlert

	for _, factorScore := range assessment.FactorScores {
		// Check for critical factor risk
		if factorScore.Score >= 85 {
			rule := s.getAlertRule(fmt.Sprintf("%s_critical_risk", factorScore.FactorID))
			if rule != nil && rule.Enabled {
				alert := AutomatedAlert{
					ID:              fmt.Sprintf("auto_alert_%s_%s_critical", assessment.ID, factorScore.FactorID),
					RuleID:          rule.ID,
					BusinessID:      assessment.BusinessID,
					Category:        factorScore.Category,
					FactorID:        factorScore.FactorID,
					Level:           RiskLevelCritical,
					Message:         fmt.Sprintf("Critical %s risk detected: %.1f", factorScore.FactorName, factorScore.Score),
					CurrentValue:    factorScore.Score,
					ThresholdValue:  85.0,
					ExceededBy:      factorScore.Score - 85.0,
					TriggeredAt:     time.Now(),
					Acknowledged:    false,
					EscalationLevel: 1,
				}
				alerts = append(alerts, alert)
			}
		}

		// Check for high factor risk
		if factorScore.Score >= 75 && factorScore.Score < 85 {
			rule := s.getAlertRule(fmt.Sprintf("%s_high_risk", factorScore.FactorID))
			if rule != nil && rule.Enabled {
				alert := AutomatedAlert{
					ID:              fmt.Sprintf("auto_alert_%s_%s_high", assessment.ID, factorScore.FactorID),
					RuleID:          rule.ID,
					BusinessID:      assessment.BusinessID,
					Category:        factorScore.Category,
					FactorID:        factorScore.FactorID,
					Level:           RiskLevelHigh,
					Message:         fmt.Sprintf("High %s risk detected: %.1f", factorScore.FactorName, factorScore.Score),
					CurrentValue:    factorScore.Score,
					ThresholdValue:  75.0,
					ExceededBy:      factorScore.Score - 75.0,
					TriggeredAt:     time.Now(),
					Acknowledged:    false,
					EscalationLevel: 1,
				}
				alerts = append(alerts, alert)
			}
		}
	}

	return alerts, nil
}

// processTrendAlerts processes trend-based alerts
func (s *AutomatedAlertService) processTrendAlerts(ctx context.Context, assessment *RiskAssessment) ([]AutomatedAlert, error) {
	var alerts []AutomatedAlert

	// Check for upward trend
	if assessment.OverallScore > 50 {
		trendingFactors := 0
		for _, factor := range assessment.FactorScores {
			if factor.Score > 60 {
				trendingFactors++
			}
		}

		if trendingFactors >= 4 {
			rule := s.getAlertRule("upward_trend")
			if rule != nil && rule.Enabled {
				alert := AutomatedAlert{
					ID:              fmt.Sprintf("auto_alert_%s_upward_trend", assessment.ID),
					RuleID:          rule.ID,
					BusinessID:      assessment.BusinessID,
					Category:        RiskCategoryOperational,
					Level:           RiskLevelMedium,
					Message:         fmt.Sprintf("Upward risk trend detected: %d factors trending", trendingFactors),
					CurrentValue:    assessment.OverallScore,
					ThresholdValue:  50.0,
					ExceededBy:      assessment.OverallScore - 50.0,
					TriggeredAt:     time.Now(),
					Acknowledged:    false,
					EscalationLevel: 1,
				}
				alerts = append(alerts, alert)
			}
		}
	}

	return alerts, nil
}

// processAnomalyAlerts processes anomaly-based alerts
func (s *AutomatedAlertService) processAnomalyAlerts(ctx context.Context, assessment *RiskAssessment) ([]AutomatedAlert, error) {
	var alerts []AutomatedAlert

	// Check for unusual patterns
	volatilityCount := 0
	for _, factor := range assessment.FactorScores {
		if factor.Score > 50 && factor.Score < 80 {
			volatilityCount++
		}
	}

	if volatilityCount >= 5 {
		rule := s.getAlertRule("high_volatility")
		if rule != nil && rule.Enabled {
			alert := AutomatedAlert{
				ID:              fmt.Sprintf("auto_alert_%s_high_volatility", assessment.ID),
				RuleID:          rule.ID,
				BusinessID:      assessment.BusinessID,
				Category:        RiskCategoryOperational,
				Level:           RiskLevelMedium,
				Message:         fmt.Sprintf("High risk volatility detected: %d factors in mid-range", volatilityCount),
				CurrentValue:    assessment.OverallScore,
				ThresholdValue:  50.0,
				ExceededBy:      assessment.OverallScore - 50.0,
				TriggeredAt:     time.Now(),
				Acknowledged:    false,
				EscalationLevel: 1,
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// processAlertWorker processes alerts from the queue
func (s *AutomatedAlertService) processAlertWorker(workerID int) {
	s.logger.Info("Starting alert processing worker",
		"worker_id", workerID,
	)

	for alert := range s.alertQueue {
		ctx := context.Background()
		ctx = context.WithValue(ctx, "request_id", fmt.Sprintf("worker_%d_%s", workerID, alert.ID))

		s.logger.Info("Processing automated alert",
			"worker_id", workerID,
			"alert_id", alert.ID,
			"business_id", alert.BusinessID,
			"level", alert.Level,
		)

		// Send notifications
		if err := s.sendAlertNotifications(ctx, alert); err != nil {
			s.logger.Error("Failed to send alert notifications",
				"worker_id", workerID,
				"alert_id", alert.ID,
				"error", err.Error(),
			)
		}

		// Handle escalation if needed
		if err := s.handleAlertEscalation(ctx, alert); err != nil {
			s.logger.Error("Failed to handle alert escalation",
				"worker_id", workerID,
				"alert_id", alert.ID,
				"error", err.Error(),
			)
		}
	}
}

// sendAlertNotifications sends notifications for an alert
func (s *AutomatedAlertService) sendAlertNotifications(ctx context.Context, alert AutomatedAlert) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Sending alert notifications",
		"request_id", requestID,
		"alert_id", alert.ID,
		"business_id", alert.BusinessID,
	)

	rule := s.getAlertRule(alert.RuleID)
	if rule == nil {
		return fmt.Errorf("alert rule not found: %s", alert.RuleID)
	}

	var notifications []AlertNotification

	// Create notifications for each channel
	for _, channel := range rule.NotificationChannels {
		provider := s.getNotificationProvider(channel)
		if provider == nil {
			s.logger.Warn("Notification provider not found",
				"request_id", requestID,
				"channel", channel,
			)
			continue
		}

		notification := AlertNotification{
			ID:        fmt.Sprintf("notification_%s_%s", alert.ID, channel),
			AlertID:   alert.ID,
			Type:      channel,
			Recipient: s.getRecipientForChannel(channel, rule.Recipients),
			Message:   s.formatAlertMessage(alert, rule.Template),
			Status:    "pending",
			CreatedAt: time.Now(),
		}

		// Send notification
		if err := provider.SendNotification(ctx, notification); err != nil {
			s.logger.Error("Failed to send notification",
				"request_id", requestID,
				"alert_id", alert.ID,
				"channel", channel,
				"error", err.Error(),
			)
			notification.Status = "failed"
		} else {
			now := time.Now()
			notification.Status = "sent"
			notification.SentAt = &now
		}

		notifications = append(notifications, notification)
	}

	// Update alert with notifications
	alert.Notifications = notifications

	s.logger.Info("Alert notifications sent",
		"request_id", requestID,
		"alert_id", alert.ID,
		"notification_count", len(notifications),
	)

	return nil
}

// handleAlertEscalation handles alert escalation
func (s *AutomatedAlertService) handleAlertEscalation(ctx context.Context, alert AutomatedAlert) error {
	requestID := ctx.Value("request_id").(string)

	rule := s.getAlertRule(alert.RuleID)
	if rule == nil {
		return fmt.Errorf("alert rule not found: %s", alert.RuleID)
	}

	// Check if escalation is needed
	if len(rule.EscalationRules) > 0 && alert.EscalationLevel < len(rule.EscalationRules) {
		escalationRule := rule.EscalationRules[alert.EscalationLevel-1]
		if escalationRule.Enabled {
			s.logger.Info("Handling alert escalation",
				"request_id", requestID,
				"alert_id", alert.ID,
				"escalation_level", alert.EscalationLevel,
			)

			// Schedule escalation
			go func() {
				time.Sleep(escalationRule.Delay)
				s.escalateAlert(ctx, alert, escalationRule)
			}()
		}
	}

	return nil
}

// escalateAlert escalates an alert
func (s *AutomatedAlertService) escalateAlert(ctx context.Context, alert AutomatedAlert, escalationRule EscalationRule) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Escalating alert",
		"request_id", requestID,
		"alert_id", alert.ID,
		"escalation_level", alert.EscalationLevel,
	)

	// Create escalation notification
	notification := AlertNotification{
		ID:        fmt.Sprintf("escalation_%s_%d", alert.ID, alert.EscalationLevel),
		AlertID:   alert.ID,
		Type:      "escalation",
		Recipient: strings.Join(escalationRule.AdditionalRecipients, ","),
		Message:   escalationRule.Message,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	// Send escalation notifications
	for _, channel := range escalationRule.Channels {
		provider := s.getNotificationProvider(channel)
		if provider != nil {
			if err := provider.SendNotification(ctx, notification); err != nil {
				s.logger.Error("Failed to send escalation notification",
					"request_id", requestID,
					"alert_id", alert.ID,
					"channel", channel,
					"error", err.Error(),
				)
			}
		}
	}
}

// Helper methods
func (s *AutomatedAlertService) getAlertRule(ruleID string) *AutomatedAlertRule {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.alertRules[ruleID]
}

func (s *AutomatedAlertService) getNotificationProvider(channel string) NotificationProvider {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.notificationProviders[channel]
}

func (s *AutomatedAlertService) getRecipientForChannel(channel string, recipients []string) string {
	if len(recipients) == 0 {
		return ""
	}

	// Simple round-robin for now
	index := time.Now().Unix() % int64(len(recipients))
	return recipients[index]
}

func (s *AutomatedAlertService) formatAlertMessage(alert AutomatedAlert, template string) string {
	if template == "" {
		return alert.Message
	}

	// Simple template replacement
	message := template
	message = strings.ReplaceAll(message, "{{business_id}}", alert.BusinessID)
	message = strings.ReplaceAll(message, "{{level}}", string(alert.Level))
	message = strings.ReplaceAll(message, "{{current_value}}", fmt.Sprintf("%.1f", alert.CurrentValue))
	message = strings.ReplaceAll(message, "{{threshold_value}}", fmt.Sprintf("%.1f", alert.ThresholdValue))
	message = strings.ReplaceAll(message, "{{exceeded_by}}", fmt.Sprintf("%.1f", alert.ExceededBy))

	return message
}

func (s *AutomatedAlertService) storeAlertHistory(businessID string, alerts []AutomatedAlert) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.alertHistory[businessID] == nil {
		s.alertHistory[businessID] = []AutomatedAlert{}
	}

	s.alertHistory[businessID] = append(s.alertHistory[businessID], alerts...)

	// Keep only last 100 alerts per business
	if len(s.alertHistory[businessID]) > 100 {
		s.alertHistory[businessID] = s.alertHistory[businessID][len(s.alertHistory[businessID])-100:]
	}
}

func (s *AutomatedAlertService) initializeDefaultRules() {
	now := time.Now()

	defaultRules := []*AutomatedAlertRule{
		{
			ID:                   "overall_critical_risk",
			Name:                 "Overall Critical Risk",
			Description:          "Alert when overall risk score exceeds 85",
			Category:             RiskCategoryOperational,
			TriggerCondition:     AutomatedTriggerConditionThresholdExceeded,
			Threshold:            85.0,
			Level:                RiskLevelCritical,
			Message:              "Critical overall risk detected",
			NotificationChannels: []string{"email", "webhook", "sms"},
			Recipients:           []string{"admin@company.com", "risk@company.com"},
			Template:             "CRITICAL: Overall risk score {{current_value}} exceeds threshold {{threshold_value}}",
			Enabled:              true,
			CooldownPeriod:       1 * time.Hour,
			CreatedAt:            now,
			UpdatedAt:            now,
		},
		{
			ID:                   "overall_high_risk",
			Name:                 "Overall High Risk",
			Description:          "Alert when overall risk score exceeds 75",
			Category:             RiskCategoryOperational,
			TriggerCondition:     AutomatedTriggerConditionThresholdExceeded,
			Threshold:            75.0,
			Level:                RiskLevelHigh,
			Message:              "High overall risk detected",
			NotificationChannels: []string{"email", "dashboard"},
			Recipients:           []string{"risk@company.com"},
			Template:             "HIGH: Overall risk score {{current_value}} exceeds threshold {{threshold_value}}",
			Enabled:              true,
			CooldownPeriod:       2 * time.Hour,
			CreatedAt:            now,
			UpdatedAt:            now,
		},
		{
			ID:                   "rapid_risk_increase",
			Name:                 "Rapid Risk Increase",
			Description:          "Alert when multiple risk factors increase rapidly",
			Category:             RiskCategoryOperational,
			TriggerCondition:     AutomatedTriggerConditionRapidIncrease,
			Threshold:            60.0,
			Level:                RiskLevelHigh,
			Message:              "Rapid risk increase detected",
			NotificationChannels: []string{"email", "webhook"},
			Recipients:           []string{"risk@company.com"},
			Template:             "RAPID INCREASE: Multiple risk factors trending upward",
			Enabled:              true,
			CooldownPeriod:       30 * time.Minute,
			CreatedAt:            now,
			UpdatedAt:            now,
		},
		{
			ID:                   "upward_trend",
			Name:                 "Upward Risk Trend",
			Description:          "Alert when risk shows upward trend",
			Category:             RiskCategoryOperational,
			TriggerCondition:     AutomatedTriggerConditionTrending,
			Threshold:            50.0,
			Level:                RiskLevelMedium,
			Message:              "Upward risk trend detected",
			NotificationChannels: []string{"email", "dashboard"},
			Recipients:           []string{"risk@company.com"},
			Template:             "TREND: Risk trending upward with score {{current_value}}",
			Enabled:              true,
			CooldownPeriod:       1 * time.Hour,
			CreatedAt:            now,
			UpdatedAt:            now,
		},
		{
			ID:                   "high_volatility",
			Name:                 "High Risk Volatility",
			Description:          "Alert when risk shows high volatility",
			Category:             RiskCategoryOperational,
			TriggerCondition:     AutomatedTriggerConditionVolatility,
			Threshold:            50.0,
			Level:                RiskLevelMedium,
			Message:              "High risk volatility detected",
			NotificationChannels: []string{"email", "dashboard"},
			Recipients:           []string{"risk@company.com"},
			Template:             "VOLATILITY: High risk volatility detected",
			Enabled:              true,
			CooldownPeriod:       1 * time.Hour,
			CreatedAt:            now,
			UpdatedAt:            now,
		},
	}

	for _, rule := range defaultRules {
		s.alertRules[rule.ID] = rule
	}
}

// RegisterNotificationProvider registers a notification provider
func (s *AutomatedAlertService) RegisterNotificationProvider(channel string, provider NotificationProvider) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.notificationProviders[channel] = provider

	s.logger.Info("Notification provider registered",
		"channel", channel,
		"provider", provider.GetProviderName(),
	)
}

// GetAlertRules returns all alert rules
func (s *AutomatedAlertService) GetAlertRules() ([]*AutomatedAlertRule, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var rules []*AutomatedAlertRule
	for _, rule := range s.alertRules {
		rules = append(rules, rule)
	}

	return rules, nil
}

// CreateAlertRule creates a new alert rule
func (s *AutomatedAlertService) CreateAlertRule(rule *AutomatedAlertRule) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	rule.ID = fmt.Sprintf("rule_%d", time.Now().UnixNano())
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	s.alertRules[rule.ID] = rule

	s.logger.Info("Alert rule created",
		"rule_id", rule.ID,
		"rule_name", rule.Name,
	)

	return nil
}

// UpdateAlertRule updates an existing alert rule
func (s *AutomatedAlertService) UpdateAlertRule(ruleID string, rule *AutomatedAlertRule) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.alertRules[ruleID]; !exists {
		return fmt.Errorf("alert rule not found: %s", ruleID)
	}

	rule.ID = ruleID
	rule.UpdatedAt = time.Now()

	s.alertRules[ruleID] = rule

	s.logger.Info("Alert rule updated",
		"rule_id", ruleID,
		"rule_name", rule.Name,
	)

	return nil
}

// DeleteAlertRule deletes an alert rule
func (s *AutomatedAlertService) DeleteAlertRule(ruleID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.alertRules[ruleID]; !exists {
		return fmt.Errorf("alert rule not found: %s", ruleID)
	}

	delete(s.alertRules, ruleID)

	s.logger.Info("Alert rule deleted",
		"rule_id", ruleID,
	)

	return nil
}

// GetAlertHistory returns alert history for a business
func (s *AutomatedAlertService) GetAlertHistory(businessID string) ([]AutomatedAlert, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	history, exists := s.alertHistory[businessID]
	if !exists {
		return []AutomatedAlert{}, nil
	}

	return history, nil
}

// RealNotificationProvider represents a real notification provider with API integration
type RealNotificationProvider struct {
	name          string
	apiKey        string
	baseURL       string
	timeout       time.Duration
	retryAttempts int
	available     bool
	logger        *observability.Logger
	httpClient    *http.Client
}

// NewRealNotificationProvider creates a new real notification provider
func NewRealNotificationProvider(name, apiKey, baseURL string, logger *observability.Logger) *RealNotificationProvider {
	return &RealNotificationProvider{
		name:          name,
		apiKey:        apiKey,
		baseURL:       baseURL,
		timeout:       30 * time.Second,
		retryAttempts: 3,
		available:     true,
		logger:        logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendNotification implements NotificationProvider interface for real providers
func (p *RealNotificationProvider) SendNotification(ctx context.Context, notification AlertNotification) error {
	requestID := ctx.Value("request_id").(string)

	p.logger.Info("Sending notification with real provider",
		"request_id", requestID,
		"provider", p.name,
		"notification_id", notification.ID,
		"type", notification.Type,
	)

	url := fmt.Sprintf("%s/notifications", p.baseURL)

	// Create request body
	requestBody := map[string]interface{}{
		"notification_id": notification.ID,
		"type":            notification.Type,
		"recipient":       notification.Recipient,
		"message":         notification.Message,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	for attempt := 0; attempt < p.retryAttempts; attempt++ {
		resp, err = p.httpClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}

		if attempt < p.retryAttempts-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	if err != nil {
		p.logger.Error("Failed to send notification with real provider",
			"request_id", requestID,
			"provider", p.name,
			"notification_id", notification.ID,
			"error", err.Error(),
		)
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.logger.Error("Real provider returned error status for notification",
			"request_id", requestID,
			"provider", p.name,
			"notification_id", notification.ID,
			"status_code", resp.StatusCode,
		)
		return fmt.Errorf("provider returned status %d", resp.StatusCode)
	}

	p.logger.Info("Successfully sent notification with real provider",
		"request_id", requestID,
		"provider", p.name,
		"notification_id", notification.ID,
	)

	return nil
}

func (p *RealNotificationProvider) GetProviderName() string {
	return p.name
}

func (p *RealNotificationProvider) IsAvailable() bool {
	return p.available
}

func (p *RealNotificationProvider) SetAvailable(available bool) {
	p.available = available
}

// Specialized notification providers
type EmailNotificationProvider struct {
	*RealNotificationProvider
}

func NewEmailNotificationProvider(apiKey, baseURL string, logger *observability.Logger) *EmailNotificationProvider {
	return &EmailNotificationProvider{
		RealNotificationProvider: NewRealNotificationProvider("email_provider", apiKey, baseURL, logger),
	}
}

type WebhookNotificationProvider struct {
	*RealNotificationProvider
}

func NewWebhookNotificationProvider(apiKey, baseURL string, logger *observability.Logger) *WebhookNotificationProvider {
	return &WebhookNotificationProvider{
		RealNotificationProvider: NewRealNotificationProvider("webhook_provider", apiKey, baseURL, logger),
	}
}

type SMSNotificationProvider struct {
	*RealNotificationProvider
}

func NewSMSNotificationProvider(apiKey, baseURL string, logger *observability.Logger) *SMSNotificationProvider {
	return &SMSNotificationProvider{
		RealNotificationProvider: NewRealNotificationProvider("sms_provider", apiKey, baseURL, logger),
	}
}

type DashboardNotificationProvider struct {
	*RealNotificationProvider
}

func NewDashboardNotificationProvider(apiKey, baseURL string, logger *observability.Logger) *DashboardNotificationProvider {
	return &DashboardNotificationProvider{
		RealNotificationProvider: NewRealNotificationProvider("dashboard_provider", apiKey, baseURL, logger),
	}
}
