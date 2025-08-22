package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// PerformanceAlertingDashboardHandler handles HTTP requests for performance alerting
type PerformanceAlertingDashboardHandler struct {
	alertingSystem *observability.PerformanceAlertingSystem
	logger         *zap.Logger
}

// NewPerformanceAlertingDashboardHandler creates a new performance alerting dashboard handler
func NewPerformanceAlertingDashboardHandler(
	alertingSystem *observability.PerformanceAlertingSystem,
	logger *zap.Logger,
) *PerformanceAlertingDashboardHandler {
	return &PerformanceAlertingDashboardHandler{
		alertingSystem: alertingSystem,
		logger:         logger,
	}
}

// GetActiveAlerts returns all active performance alerts
func (h *PerformanceAlertingDashboardHandler) GetActiveAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	alerts, err := h.alertingSystem.GetActiveAlerts(ctx)
	if err != nil {
		h.logger.Error("Failed to get active alerts", zap.Error(err))
		http.Error(w, "Failed to get active alerts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// GetAlertHistory returns alert history with optional filtering
func (h *PerformanceAlertingDashboardHandler) GetAlertHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	severity := r.URL.Query().Get("severity")
	category := r.URL.Query().Get("category")
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	limit := 100 // default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var startTime, endTime *time.Time
	if startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = &t
		}
	}
	if endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = &t
		}
	}

	history, err := h.alertingSystem.GetAlertHistory(ctx, observability.AlertHistoryFilter{
		Limit:     limit,
		Offset:    offset,
		Severity:  severity,
		Category:  category,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		h.logger.Error("Failed to get alert history", zap.Error(err))
		http.Error(w, "Failed to get alert history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": history.Alerts,
		"total":   history.Total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetPerformanceRules returns all performance alert rules
func (h *PerformanceAlertingDashboardHandler) GetPerformanceRules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rules, err := h.alertingSystem.GetPerformanceRules(ctx)
	if err != nil {
		h.logger.Error("Failed to get performance rules", zap.Error(err))
		http.Error(w, "Failed to get performance rules", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rules": rules,
		"count": len(rules),
	})
}

// CreatePerformanceRule creates a new performance alert rule
func (h *PerformanceAlertingDashboardHandler) CreatePerformanceRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var rule observability.PerformanceAlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		h.logger.Error("Failed to decode performance rule", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if rule.Name == "" || rule.MetricType == "" || rule.Condition == "" {
		http.Error(w, "Missing required fields: name, metric_type, condition", http.StatusBadRequest)
		return
	}

	createdRule, err := h.alertingSystem.CreatePerformanceRule(ctx, &rule)
	if err != nil {
		h.logger.Error("Failed to create performance rule", zap.Error(err))
		http.Error(w, "Failed to create performance rule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdRule)
}

// UpdatePerformanceRule updates an existing performance alert rule
func (h *PerformanceAlertingDashboardHandler) UpdatePerformanceRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract rule ID from URL path
	ruleID := r.URL.Query().Get("rule_id")
	if ruleID == "" {
		http.Error(w, "Missing rule_id parameter", http.StatusBadRequest)
		return
	}

	var rule observability.PerformanceAlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		h.logger.Error("Failed to decode performance rule", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	rule.ID = ruleID // Ensure the ID matches

	updatedRule, err := h.alertingSystem.UpdatePerformanceRule(ctx, &rule)
	if err != nil {
		h.logger.Error("Failed to update performance rule", zap.Error(err))
		http.Error(w, "Failed to update performance rule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedRule)
}

// DeletePerformanceRule deletes a performance alert rule
func (h *PerformanceAlertingDashboardHandler) DeletePerformanceRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ruleID := r.URL.Query().Get("rule_id")
	if ruleID == "" {
		http.Error(w, "Missing rule_id parameter", http.StatusBadRequest)
		return
	}

	err := h.alertingSystem.DeletePerformanceRule(ctx, ruleID)
	if err != nil {
		h.logger.Error("Failed to delete performance rule", zap.Error(err))
		http.Error(w, "Failed to delete performance rule", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// EnablePerformanceRule enables a performance alert rule
func (h *PerformanceAlertingDashboardHandler) EnablePerformanceRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ruleID := r.URL.Query().Get("rule_id")
	if ruleID == "" {
		http.Error(w, "Missing rule_id parameter", http.StatusBadRequest)
		return
	}

	err := h.alertingSystem.EnablePerformanceRule(ctx, ruleID)
	if err != nil {
		h.logger.Error("Failed to enable performance rule", zap.Error(err))
		http.Error(w, "Failed to enable performance rule", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "enabled"})
}

// DisablePerformanceRule disables a performance alert rule
func (h *PerformanceAlertingDashboardHandler) DisablePerformanceRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ruleID := r.URL.Query().Get("rule_id")
	if ruleID == "" {
		http.Error(w, "Missing rule_id parameter", http.StatusBadRequest)
		return
	}

	err := h.alertingSystem.DisablePerformanceRule(ctx, ruleID)
	if err != nil {
		h.logger.Error("Failed to disable performance rule", zap.Error(err))
		http.Error(w, "Failed to disable performance rule", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "disabled"})
}

// GetNotificationChannels returns all notification channels
func (h *PerformanceAlertingDashboardHandler) GetNotificationChannels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	channels, err := h.alertingSystem.GetNotificationChannels(ctx)
	if err != nil {
		h.logger.Error("Failed to get notification channels", zap.Error(err))
		http.Error(w, "Failed to get notification channels", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"channels": channels,
		"count":    len(channels),
	})
}

// TestNotificationChannel tests a notification channel
func (h *PerformanceAlertingDashboardHandler) TestNotificationChannel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	channelName := r.URL.Query().Get("channel")
	if channelName == "" {
		http.Error(w, "Missing channel parameter", http.StatusBadRequest)
		return
	}

	var testNotification struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&testNotification); err != nil {
		h.logger.Error("Failed to decode test notification", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.alertingSystem.TestNotificationChannel(ctx, channelName, testNotification.Message)
	if err != nil {
		h.logger.Error("Failed to test notification channel", zap.Error(err))
		http.Error(w, "Failed to test notification channel", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "test_sent"})
}

// GetAlertStatistics returns alert statistics and metrics
func (h *PerformanceAlertingDashboardHandler) GetAlertStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse time range
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	var startTime, endTime time.Time
	if startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = t
		} else {
			startTime = time.Now().Add(-24 * time.Hour) // Default to last 24 hours
		}
	} else {
		startTime = time.Now().Add(-24 * time.Hour)
	}

	if endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = t
		} else {
			endTime = time.Now()
		}
	} else {
		endTime = time.Now()
	}

	stats, err := h.alertingSystem.GetAlertStatistics(ctx, startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get alert statistics", zap.Error(err))
		http.Error(w, "Failed to get alert statistics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetSystemConfiguration returns the current alerting system configuration
func (h *PerformanceAlertingDashboardHandler) GetSystemConfiguration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	config, err := h.alertingSystem.GetConfiguration(ctx)
	if err != nil {
		h.logger.Error("Failed to get system configuration", zap.Error(err))
		http.Error(w, "Failed to get system configuration", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// UpdateSystemConfiguration updates the alerting system configuration
func (h *PerformanceAlertingDashboardHandler) UpdateSystemConfiguration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var config observability.PerformanceAlertingConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		h.logger.Error("Failed to decode configuration", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.alertingSystem.UpdateConfiguration(ctx, &config)
	if err != nil {
		h.logger.Error("Failed to update system configuration", zap.Error(err))
		http.Error(w, "Failed to update system configuration", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// GetEscalationPolicies returns all escalation policies
func (h *PerformanceAlertingDashboardHandler) GetEscalationPolicies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	policies, err := h.alertingSystem.GetEscalationPolicies(ctx)
	if err != nil {
		h.logger.Error("Failed to get escalation policies", zap.Error(err))
		http.Error(w, "Failed to get escalation policies", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"policies": policies,
		"count":    len(policies),
	})
}

// CreateEscalationPolicy creates a new escalation policy
func (h *PerformanceAlertingDashboardHandler) CreateEscalationPolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var policy observability.EscalationPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		h.logger.Error("Failed to decode escalation policy", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	createdPolicy, err := h.alertingSystem.CreateEscalationPolicy(ctx, &policy)
	if err != nil {
		h.logger.Error("Failed to create escalation policy", zap.Error(err))
		http.Error(w, "Failed to create escalation policy", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPolicy)
}

// GetSystemHealth returns the health status of the alerting system
func (h *PerformanceAlertingDashboardHandler) GetSystemHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	health, err := h.alertingSystem.GetSystemHealth(ctx)
	if err != nil {
		h.logger.Error("Failed to get system health", zap.Error(err))
		http.Error(w, "Failed to get system health", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// ManualAlertTrigger manually triggers an alert for testing
func (h *PerformanceAlertingDashboardHandler) ManualAlertTrigger(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var triggerRequest struct {
		RuleID string            `json:"rule_id"`
		Labels map[string]string `json:"labels,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&triggerRequest); err != nil {
		h.logger.Error("Failed to decode trigger request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if triggerRequest.RuleID == "" {
		http.Error(w, "Missing rule_id", http.StatusBadRequest)
		return
	}

	err := h.alertingSystem.ManualAlertTrigger(ctx, triggerRequest.RuleID, triggerRequest.Labels)
	if err != nil {
		h.logger.Error("Failed to trigger manual alert", zap.Error(err))
		http.Error(w, "Failed to trigger manual alert", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "triggered"})
}

// GetAlertMetrics returns detailed alert metrics
func (h *PerformanceAlertingDashboardHandler) GetAlertMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.alertingSystem.GetAlertMetrics(ctx)
	if err != nil {
		h.logger.Error("Failed to get alert metrics", zap.Error(err))
		http.Error(w, "Failed to get alert metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
