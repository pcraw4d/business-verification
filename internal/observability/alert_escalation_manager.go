package observability

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// NewAlertEscalationManager creates a new alert escalation manager
func NewAlertEscalationManager(config PerformanceAlertingConfig, logger *zap.Logger) *AlertEscalationManager {
	return &AlertEscalationManager{
		escalationPolicies: make(map[string]*EscalationPolicy),
		activeEscalations:  make(map[string]*EscalationEvent),
		config:             config,
		logger:             logger,
	}
}

// StartEscalation starts an escalation for an alert
func (aem *AlertEscalationManager) StartEscalation(alert *PerformanceAlert, policy *EscalationPolicy) {
	aem.mu.Lock()
	defer aem.mu.Unlock()

	// Check if escalation already exists for this alert
	if _, exists := aem.activeEscalations[alert.ID]; exists {
		aem.logger.Warn("Escalation already exists for alert", zap.String("alert_id", alert.ID))
		return
	}

	// Create escalation event
	escalation := &EscalationEvent{
		ID:        fmt.Sprintf("esc-%s-%d", alert.ID, time.Now().Unix()),
		AlertID:   alert.ID,
		PolicyID:  policy.ID,
		Level:     1, // Start at level 1
		Status:    "active",
		StartedAt: time.Now().UTC(),
	}

	// Add to active escalations
	aem.activeEscalations[alert.ID] = escalation

	// Store policy if not already stored
	aem.escalationPolicies[policy.ID] = policy

	aem.logger.Info("Escalation started",
		zap.String("escalation_id", escalation.ID),
		zap.String("alert_id", alert.ID),
		zap.String("policy_id", policy.ID),
		zap.Int("level", escalation.Level))
}

// StopEscalation stops an escalation for an alert
func (aem *AlertEscalationManager) StopEscalation(alertID string) {
	aem.mu.Lock()
	defer aem.mu.Unlock()

	escalation, exists := aem.activeEscalations[alertID]
	if !exists {
		return
	}

	// Mark as completed
	now := time.Now().UTC()
	escalation.Status = "completed"
	escalation.CompletedAt = &now

	// Remove from active escalations
	delete(aem.activeEscalations, alertID)

	aem.logger.Info("Escalation stopped",
		zap.String("escalation_id", escalation.ID),
		zap.String("alert_id", alertID),
		zap.Int("final_level", escalation.Level))
}

// ProcessEscalations processes active escalations
func (aem *AlertEscalationManager) ProcessEscalations(ctx context.Context) {
	aem.mu.RLock()
	escalations := make([]*EscalationEvent, 0, len(aem.activeEscalations))
	for _, escalation := range aem.activeEscalations {
		if escalation.Status == "active" {
			escalations = append(escalations, escalation)
		}
	}
	aem.mu.RUnlock()

	for _, escalation := range escalations {
		aem.processEscalation(ctx, escalation)
	}
}

// processEscalation processes a single escalation
func (aem *AlertEscalationManager) processEscalation(ctx context.Context, escalation *EscalationEvent) {
	policy, exists := aem.escalationPolicies[escalation.PolicyID]
	if !exists {
		aem.logger.Error("Escalation policy not found", zap.String("policy_id", escalation.PolicyID))
		return
	}

	// Check if it's time for the next escalation level
	if aem.shouldEscalate(escalation, policy) {
		aem.escalateToNextLevel(ctx, escalation, policy)
	}
}

// shouldEscalate determines if an escalation should proceed to the next level
func (aem *AlertEscalationManager) shouldEscalate(escalation *EscalationEvent, policy *EscalationPolicy) bool {
	// Find the current level configuration
	var currentLevel *EscalationLevel
	for _, level := range policy.Levels {
		if level.Level == escalation.Level {
			currentLevel = &level
			break
		}
	}

	if currentLevel == nil {
		// No more levels to escalate to
		return false
	}

	// Check if enough time has passed since the escalation started
	timeSinceStart := time.Since(escalation.StartedAt)
	return timeSinceStart >= currentLevel.Delay
}

// escalateToNextLevel escalates an alert to the next level
func (aem *AlertEscalationManager) escalateToNextLevel(ctx context.Context, escalation *EscalationEvent, policy *EscalationPolicy) {
	aem.mu.Lock()
	defer aem.mu.Unlock()

	// Find the next level
	var nextLevel *EscalationLevel
	for _, level := range policy.Levels {
		if level.Level == escalation.Level+1 {
			nextLevel = &level
			break
		}
	}

	if nextLevel == nil {
		// No more levels to escalate to
		aem.logger.Info("Escalation reached maximum level",
			zap.String("escalation_id", escalation.ID),
			zap.String("alert_id", escalation.AlertID),
			zap.Int("final_level", escalation.Level))
		return
	}

	// Increment level
	escalation.Level++
	now := time.Now().UTC()
	escalation.EscalatedAt = &now

	// Send notifications for this level
	aem.sendEscalationNotifications(ctx, escalation, nextLevel)

	aem.logger.Info("Alert escalated to next level",
		zap.String("escalation_id", escalation.ID),
		zap.String("alert_id", escalation.AlertID),
		zap.Int("new_level", escalation.Level),
		zap.Strings("notifications", nextLevel.Notifications))
}

// sendEscalationNotifications sends notifications for an escalation level
func (aem *AlertEscalationManager) sendEscalationNotifications(ctx context.Context, escalation *EscalationEvent, level *EscalationLevel) {
	// In a real implementation, this would send notifications through the notification system
	// For now, we'll just log the notifications
	for _, notificationType := range level.Notifications {
		aem.logger.Info("Escalation notification sent",
			zap.String("escalation_id", escalation.ID),
			zap.String("alert_id", escalation.AlertID),
			zap.Int("level", level.Level),
			zap.String("notification_type", notificationType))
	}

	escalation.NotificationsSent++
}

// AddEscalationPolicy adds an escalation policy
func (aem *AlertEscalationManager) AddEscalationPolicy(policy *EscalationPolicy) error {
	aem.mu.Lock()
	defer aem.mu.Unlock()

	if policy.ID == "" {
		return fmt.Errorf("policy ID is required")
	}

	if _, exists := aem.escalationPolicies[policy.ID]; exists {
		return fmt.Errorf("escalation policy with ID %s already exists", policy.ID)
	}

	// Validate policy
	if err := aem.validateEscalationPolicy(policy); err != nil {
		return fmt.Errorf("invalid escalation policy: %w", err)
	}

	aem.escalationPolicies[policy.ID] = policy
	aem.logger.Info("Escalation policy added", zap.String("policy_id", policy.ID))
	return nil
}

// UpdateEscalationPolicy updates an escalation policy
func (aem *AlertEscalationManager) UpdateEscalationPolicy(policyID string, policy *EscalationPolicy) error {
	aem.mu.Lock()
	defer aem.mu.Unlock()

	if _, exists := aem.escalationPolicies[policyID]; !exists {
		return fmt.Errorf("escalation policy with ID %s does not exist", policyID)
	}

	// Validate policy
	if err := aem.validateEscalationPolicy(policy); err != nil {
		return fmt.Errorf("invalid escalation policy: %w", err)
	}

	policy.ID = policyID
	aem.escalationPolicies[policyID] = policy
	aem.logger.Info("Escalation policy updated", zap.String("policy_id", policyID))
	return nil
}

// DeleteEscalationPolicy deletes an escalation policy
func (aem *AlertEscalationManager) DeleteEscalationPolicy(policyID string) error {
	aem.mu.Lock()
	defer aem.mu.Unlock()

	if _, exists := aem.escalationPolicies[policyID]; !exists {
		return fmt.Errorf("escalation policy with ID %s does not exist", policyID)
	}

	delete(aem.escalationPolicies, policyID)
	aem.logger.Info("Escalation policy deleted", zap.String("policy_id", policyID))
	return nil
}

// GetEscalationPolicy returns an escalation policy
func (aem *AlertEscalationManager) GetEscalationPolicy(policyID string) (*EscalationPolicy, bool) {
	aem.mu.RLock()
	defer aem.mu.RUnlock()

	policy, exists := aem.escalationPolicies[policyID]
	return policy, exists
}

// GetActiveEscalations returns all active escalations
func (aem *AlertEscalationManager) GetActiveEscalations() []*EscalationEvent {
	aem.mu.RLock()
	defer aem.mu.RUnlock()

	escalations := make([]*EscalationEvent, 0, len(aem.activeEscalations))
	for _, escalation := range aem.activeEscalations {
		if escalation.Status == "active" {
			escalations = append(escalations, escalation)
		}
	}
	return escalations
}

// GetEscalationForAlert returns the escalation for a specific alert
func (aem *AlertEscalationManager) GetEscalationForAlert(alertID string) (*EscalationEvent, bool) {
	aem.mu.RLock()
	defer aem.mu.RUnlock()

	escalation, exists := aem.activeEscalations[alertID]
	return escalation, exists
}

// validateEscalationPolicy validates an escalation policy
func (aem *AlertEscalationManager) validateEscalationPolicy(policy *EscalationPolicy) error {
	if policy.ID == "" {
		return fmt.Errorf("policy ID is required")
	}

	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}

	if len(policy.Levels) == 0 {
		return fmt.Errorf("policy must have at least one escalation level")
	}

	// Validate levels
	levelNumbers := make(map[int]bool)
	for i, level := range policy.Levels {
		if level.Level <= 0 {
			return fmt.Errorf("level %d: level number must be positive", i+1)
		}

		if levelNumbers[level.Level] {
			return fmt.Errorf("level %d: duplicate level number %d", i+1, level.Level)
		}
		levelNumbers[level.Level] = true

		if level.Delay < 0 {
			return fmt.Errorf("level %d: delay must be non-negative", i+1)
		}

		if len(level.Notifications) == 0 {
			return fmt.Errorf("level %d: must have at least one notification type", i+1)
		}

		// Validate notification types
		validNotificationTypes := []string{"email", "slack", "pagerduty", "webhook", "sms"}
		for _, notificationType := range level.Notifications {
			valid := false
			for _, validType := range validNotificationTypes {
				if notificationType == validType {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("level %d: invalid notification type %s", i+1, notificationType)
			}
		}
	}

	// Check if levels are sequential
	expectedLevel := 1
	for _, level := range policy.Levels {
		if level.Level != expectedLevel {
			return fmt.Errorf("escalation levels must be sequential starting from 1, found level %d", level.Level)
		}
		expectedLevel++
	}

	return nil
}

// CreateDefaultEscalationPolicies creates default escalation policies
func (aem *AlertEscalationManager) CreateDefaultEscalationPolicies() {
	// Critical alerts escalation policy
	criticalPolicy := &EscalationPolicy{
		ID:              "critical_alerts",
		Name:            "Critical Alerts Escalation",
		Description:     "Escalation policy for critical performance alerts",
		MaxEscalations:  3,
		EscalationDelay: 15 * time.Minute,
		Levels: []EscalationLevel{
			{
				Level:         1,
				Delay:         5 * time.Minute,
				Notifications: []string{"email", "slack"},
				Recipients:    []string{"oncall@company.com", "#alerts"},
			},
			{
				Level:         2,
				Delay:         15 * time.Minute,
				Notifications: []string{"email", "slack", "pagerduty"},
				Recipients:    []string{"oncall@company.com", "#alerts", "pagerduty"},
			},
			{
				Level:         3,
				Delay:         30 * time.Minute,
				Notifications: []string{"email", "slack", "pagerduty", "sms"},
				Recipients:    []string{"oncall@company.com", "#alerts", "pagerduty", "+1234567890"},
			},
		},
	}

	// Warning alerts escalation policy
	warningPolicy := &EscalationPolicy{
		ID:              "warning_alerts",
		Name:            "Warning Alerts Escalation",
		Description:     "Escalation policy for warning performance alerts",
		MaxEscalations:  2,
		EscalationDelay: 30 * time.Minute,
		Levels: []EscalationLevel{
			{
				Level:         1,
				Delay:         15 * time.Minute,
				Notifications: []string{"email", "slack"},
				Recipients:    []string{"oncall@company.com", "#alerts"},
			},
			{
				Level:         2,
				Delay:         30 * time.Minute,
				Notifications: []string{"email", "slack", "pagerduty"},
				Recipients:    []string{"oncall@company.com", "#alerts", "pagerduty"},
			},
		},
	}

	// Add policies
	aem.AddEscalationPolicy(criticalPolicy)
	aem.AddEscalationPolicy(warningPolicy)
}

// GetEscalationStatistics returns statistics for escalations
func (aem *AlertEscalationManager) GetEscalationStatistics() *EscalationStatistics {
	aem.mu.RLock()
	defer aem.mu.RUnlock()

	stats := &EscalationStatistics{
		TotalPolicies:     len(aem.escalationPolicies),
		ActiveEscalations: len(aem.activeEscalations),
		PolicyStats:       make(map[string]*PolicyStatistics),
	}

	// Calculate statistics for each policy
	for policyID, policy := range aem.escalationPolicies {
		policyStats := &PolicyStatistics{
			PolicyID:          policyID,
			PolicyName:        policy.Name,
			MaxLevels:         len(policy.Levels),
			ActiveEscalations: 0,
		}

		// Count active escalations for this policy
		for _, escalation := range aem.activeEscalations {
			if escalation.PolicyID == policyID && escalation.Status == "active" {
				policyStats.ActiveEscalations++
			}
		}

		stats.PolicyStats[policyID] = policyStats
	}

	return stats
}

// EscalationStatistics represents escalation statistics
type EscalationStatistics struct {
	TotalPolicies     int                          `json:"total_policies"`
	ActiveEscalations int                          `json:"active_escalations"`
	PolicyStats       map[string]*PolicyStatistics `json:"policy_stats"`
}

// PolicyStatistics represents statistics for a policy
type PolicyStatistics struct {
	PolicyID          string `json:"policy_id"`
	PolicyName        string `json:"policy_name"`
	MaxLevels         int    `json:"max_levels"`
	ActiveEscalations int    `json:"active_escalations"`
}
