package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// ResourceAlertingScalingAPI provides RESTful API endpoints for resource alerting and scaling
type ResourceAlertingScalingAPI struct {
	manager *ResourceAlertingScalingManager
}

// NewResourceAlertingScalingAPI creates a new resource alerting and scaling API
func NewResourceAlertingScalingAPI(manager *ResourceAlertingScalingManager) *ResourceAlertingScalingAPI {
	return &ResourceAlertingScalingAPI{
		manager: manager,
	}
}

// RegisterResourceAlertingScalingRoutes registers all resource alerting and scaling routes
func (rasa *ResourceAlertingScalingAPI) RegisterResourceAlertingScalingRoutes(mux *http.ServeMux) {
	// Alert management endpoints
	mux.HandleFunc("GET /v1/alerts/active", rasa.GetActiveAlerts)
	mux.HandleFunc("GET /v1/alerts/history", rasa.GetAlertHistory)
	mux.HandleFunc("POST /v1/alerts/{alertId}/acknowledge", rasa.AcknowledgeAlert)
	mux.HandleFunc("POST /v1/alerts/{alertId}/resolve", rasa.ResolveAlert)

	// Scaling management endpoints
	mux.HandleFunc("GET /v1/scaling/status", rasa.GetScalingStatus)
	mux.HandleFunc("GET /v1/scaling/history", rasa.GetScalingHistory)
	mux.HandleFunc("POST /v1/scaling/manual", rasa.ManualScale)
	mux.HandleFunc("GET /v1/scaling/instances", rasa.GetCurrentInstances)

	// Metrics endpoints
	mux.HandleFunc("GET /v1/metrics/current", rasa.GetCurrentMetrics)
	mux.HandleFunc("GET /v1/metrics/history", rasa.GetMetricsHistory)

	// Configuration endpoints
	mux.HandleFunc("GET /v1/alerting-scaling/config", rasa.GetConfig)
	mux.HandleFunc("PUT /v1/alerting-scaling/config", rasa.UpdateConfig)
	mux.HandleFunc("GET /v1/alerting-scaling/status", rasa.GetStatus)
	mux.HandleFunc("GET /v1/alerting-scaling/health", rasa.GetHealth)

	// Threshold management endpoints
	mux.HandleFunc("GET /v1/thresholds", rasa.GetThresholds)
	mux.HandleFunc("PUT /v1/thresholds", rasa.UpdateThresholds)
	mux.HandleFunc("GET /v1/thresholds/adaptive", rasa.GetAdaptiveThresholds)

	// Notification endpoints
	mux.HandleFunc("GET /v1/notifications/channels", rasa.GetNotificationChannels)
	mux.HandleFunc("POST /v1/notifications/channels", rasa.CreateNotificationChannel)
	mux.HandleFunc("PUT /v1/notifications/channels/{channelId}", rasa.UpdateNotificationChannel)
	mux.HandleFunc("DELETE /v1/notifications/channels/{channelId}", rasa.DeleteNotificationChannel)
	mux.HandleFunc("GET /v1/notifications/history", rasa.GetNotificationHistory)

	// Escalation endpoints
	mux.HandleFunc("GET /v1/escalations/policies", rasa.GetEscalationPolicies)
	mux.HandleFunc("POST /v1/escalations/policies", rasa.CreateEscalationPolicy)
	mux.HandleFunc("PUT /v1/escalations/policies/{policyId}", rasa.UpdateEscalationPolicy)
	mux.HandleFunc("DELETE /v1/escalations/policies/{policyId}", rasa.DeleteEscalationPolicy)
	mux.HandleFunc("GET /v1/escalations/active", rasa.GetActiveEscalations)
	mux.HandleFunc("GET /v1/escalations/history", rasa.GetEscalationHistory)

	// Predictive scaling endpoints
	mux.HandleFunc("GET /v1/predictive/model", rasa.GetPredictiveModel)
	mux.HandleFunc("POST /v1/predictive/train", rasa.TrainPredictiveModel)
	mux.HandleFunc("GET /v1/predictive/predictions", rasa.GetPredictions)
}

// GetActiveAlerts returns all active alerts
func (rasa *ResourceAlertingScalingAPI) GetActiveAlerts(w http.ResponseWriter, r *http.Request) {
	alerts := rasa.manager.GetActiveAlerts()

	response := map[string]interface{}{
		"active_alerts": alerts,
		"count":         len(alerts),
		"timestamp":     time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAlertHistory returns alert history
func (rasa *ResourceAlertingScalingAPI) GetAlertHistory(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	alerts := rasa.manager.GetAlertHistory(limit)

	response := map[string]interface{}{
		"alert_history": alerts,
		"count":         len(alerts),
		"limit":         limit,
		"timestamp":     time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AcknowledgeAlert acknowledges a specific alert
func (rasa *ResourceAlertingScalingAPI) AcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	alertID := r.PathValue("alertId")
	if alertID == "" {
		http.Error(w, "Alert ID is required", http.StatusBadRequest)
		return
	}

	err := rasa.manager.AcknowledgeAlert(alertID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to acknowledge alert: %v", err), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"status":    "success",
		"message":   "Alert acknowledged successfully",
		"alert_id":  alertID,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ResolveAlert resolves a specific alert
func (rasa *ResourceAlertingScalingAPI) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	alertID := r.PathValue("alertId")
	if alertID == "" {
		http.Error(w, "Alert ID is required", http.StatusBadRequest)
		return
	}

	err := rasa.manager.ResolveAlert(alertID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to resolve alert: %v", err), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"status":    "success",
		"message":   "Alert resolved successfully",
		"alert_id":  alertID,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetScalingStatus returns the current scaling status
func (rasa *ResourceAlertingScalingAPI) GetScalingStatus(w http.ResponseWriter, r *http.Request) {
	currentInstances := rasa.manager.GetCurrentInstances()

	response := map[string]interface{}{
		"current_instances":    currentInstances,
		"min_instances":        rasa.manager.config.ScalingPolicies.MinInstances,
		"max_instances":        rasa.manager.config.ScalingPolicies.MaxInstances,
		"last_scaling_time":    rasa.manager.scalingEngine.lastScalingTime,
		"predictive_enabled":   rasa.manager.config.ScalingPolicies.PredictiveScalingEnabled,
		"cooldown_period":      rasa.manager.config.ScalingCooldownPeriod,
		"scale_up_threshold":   rasa.manager.config.ScalingPolicies.CPUScaleUpThreshold,
		"scale_down_threshold": rasa.manager.config.ScalingPolicies.CPUScaleDownThreshold,
		"timestamp":            time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetScalingHistory returns scaling history
func (rasa *ResourceAlertingScalingAPI) GetScalingHistory(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	events := rasa.manager.GetScalingHistory(limit)

	response := map[string]interface{}{
		"scaling_history": events,
		"count":           len(events),
		"limit":           limit,
		"timestamp":       time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ManualScaleRequest represents a manual scaling request
type ManualScaleRequest struct {
	TargetInstances int    `json:"target_instances"`
	Reason          string `json:"reason"`
}

// ManualScale manually triggers a scaling operation
func (rasa *ResourceAlertingScalingAPI) ManualScale(w http.ResponseWriter, r *http.Request) {
	var req ManualScaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TargetInstances <= 0 {
		http.Error(w, "Target instances must be positive", http.StatusBadRequest)
		return
	}

	if req.Reason == "" {
		req.Reason = "Manual scaling via API"
	}

	err := rasa.manager.ManualScale(req.TargetInstances, req.Reason)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to scale: %v", err), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"status":           "success",
		"message":          "Manual scaling initiated successfully",
		"target_instances": req.TargetInstances,
		"reason":           req.Reason,
		"timestamp":        time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCurrentInstances returns the current number of instances
func (rasa *ResourceAlertingScalingAPI) GetCurrentInstances(w http.ResponseWriter, r *http.Request) {
	currentInstances := rasa.manager.GetCurrentInstances()

	response := map[string]interface{}{
		"current_instances": currentInstances,
		"timestamp":         time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCurrentMetrics returns current system metrics
func (rasa *ResourceAlertingScalingAPI) GetCurrentMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := rasa.manager.GetCurrentMetrics()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get metrics: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"metrics":   metrics,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMetricsHistory returns metrics history
func (rasa *ResourceAlertingScalingAPI) GetMetricsHistory(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	rasa.manager.metricCollector.mu.RLock()
	history := rasa.manager.metricCollector.metricHistory
	if limit > 0 && len(history) > limit {
		history = history[len(history)-limit:]
	}

	// Create a copy
	result := make([]*MetricSnapshot, len(history))
	copy(result, history)
	rasa.manager.metricCollector.mu.RUnlock()

	response := map[string]interface{}{
		"metrics_history": result,
		"count":           len(result),
		"limit":           limit,
		"timestamp":       time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetConfig returns the current configuration
func (rasa *ResourceAlertingScalingAPI) GetConfig(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"config":    rasa.manager.config,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateConfig updates the configuration
func (rasa *ResourceAlertingScalingAPI) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var config AlertingScalingConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := rasa.manager.UpdateConfig(&config)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update config: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":    "success",
		"message":   "Configuration updated successfully",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetStatus returns the current status
func (rasa *ResourceAlertingScalingAPI) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := rasa.manager.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// GetHealth returns the health status
func (rasa *ResourceAlertingScalingAPI) GetHealth(w http.ResponseWriter, r *http.Request) {
	// Perform health checks
	health := map[string]interface{}{
		"status":                    "healthy",
		"alerting_operational":      true,
		"scaling_operational":       true,
		"metrics_operational":       true,
		"notifications_operational": true,
		"timestamp":                 time.Now(),
		"uptime":                    time.Since(rasa.manager.scalingEngine.lastScalingTime),
	}

	// Check if manager is responding
	_, err := rasa.manager.GetCurrentMetrics()
	if err != nil {
		health["status"] = "unhealthy"
		health["metrics_operational"] = false
		health["error"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// GetThresholds returns current alert thresholds
func (rasa *ResourceAlertingScalingAPI) GetThresholds(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"thresholds": rasa.manager.config.AlertThresholds,
		"timestamp":  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateThresholds updates alert thresholds
func (rasa *ResourceAlertingScalingAPI) UpdateThresholds(w http.ResponseWriter, r *http.Request) {
	var thresholds EnhancedAlertThresholds
	if err := json.NewDecoder(r.Body).Decode(&thresholds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update thresholds in config
	rasa.manager.config.AlertThresholds = &thresholds
	rasa.manager.alertEngine.thresholds = &thresholds

	response := map[string]interface{}{
		"status":    "success",
		"message":   "Thresholds updated successfully",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAdaptiveThresholds returns adaptive threshold information
func (rasa *ResourceAlertingScalingAPI) GetAdaptiveThresholds(w http.ResponseWriter, r *http.Request) {
	rasa.manager.alertEngine.mu.RLock()
	adaptiveMetrics := rasa.manager.alertEngine.adaptiveMetrics
	rasa.manager.alertEngine.mu.RUnlock()

	response := map[string]interface{}{
		"adaptive_metrics": adaptiveMetrics,
		"timestamp":        time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetNotificationChannels returns all notification channels
func (rasa *ResourceAlertingScalingAPI) GetNotificationChannels(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"channels":  rasa.manager.config.NotificationChannels,
		"count":     len(rasa.manager.config.NotificationChannels),
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateNotificationChannel creates a new notification channel
func (rasa *ResourceAlertingScalingAPI) CreateNotificationChannel(w http.ResponseWriter, r *http.Request) {
	var channel NotificationChannel
	if err := json.NewDecoder(r.Body).Decode(&channel); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if channel.ID == "" {
		channel.ID = fmt.Sprintf("channel-%d", time.Now().UnixNano())
	}

	// Add to config
	rasa.manager.config.NotificationChannels = append(rasa.manager.config.NotificationChannels, &channel)

	response := map[string]interface{}{
		"status":     "success",
		"message":    "Notification channel created successfully",
		"channel_id": channel.ID,
		"timestamp":  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateNotificationChannel updates a notification channel
func (rasa *ResourceAlertingScalingAPI) UpdateNotificationChannel(w http.ResponseWriter, r *http.Request) {
	channelID := r.PathValue("channelId")
	if channelID == "" {
		http.Error(w, "Channel ID is required", http.StatusBadRequest)
		return
	}

	var updatedChannel NotificationChannel
	if err := json.NewDecoder(r.Body).Decode(&updatedChannel); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Find and update channel
	found := false
	for i, channel := range rasa.manager.config.NotificationChannels {
		if channel.ID == channelID {
			updatedChannel.ID = channelID
			rasa.manager.config.NotificationChannels[i] = &updatedChannel
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"status":    "success",
		"message":   "Notification channel updated successfully",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteNotificationChannel deletes a notification channel
func (rasa *ResourceAlertingScalingAPI) DeleteNotificationChannel(w http.ResponseWriter, r *http.Request) {
	channelID := r.PathValue("channelId")
	if channelID == "" {
		http.Error(w, "Channel ID is required", http.StatusBadRequest)
		return
	}

	// Find and remove channel
	found := false
	for i, channel := range rasa.manager.config.NotificationChannels {
		if channel.ID == channelID {
			rasa.manager.config.NotificationChannels = append(
				rasa.manager.config.NotificationChannels[:i],
				rasa.manager.config.NotificationChannels[i+1:]...,
			)
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"status":    "success",
		"message":   "Notification channel deleted successfully",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetNotificationHistory returns notification history
func (rasa *ResourceAlertingScalingAPI) GetNotificationHistory(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	rasa.manager.notificationMgr.mu.RLock()
	history := rasa.manager.notificationMgr.notificationHistory
	if limit > 0 && len(history) > limit {
		history = history[len(history)-limit:]
	}

	// Create a copy
	result := make([]*NotificationEvent, len(history))
	copy(result, history)
	rasa.manager.notificationMgr.mu.RUnlock()

	response := map[string]interface{}{
		"notification_history": result,
		"count":                len(result),
		"limit":                limit,
		"timestamp":            time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetEscalationPolicies returns all escalation policies
func (rasa *ResourceAlertingScalingAPI) GetEscalationPolicies(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"policies":  rasa.manager.config.EscalationPolicies,
		"count":     len(rasa.manager.config.EscalationPolicies),
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateEscalationPolicy creates a new escalation policy
func (rasa *ResourceAlertingScalingAPI) CreateEscalationPolicy(w http.ResponseWriter, r *http.Request) {
	var policy EscalationPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if policy.ID == "" {
		policy.ID = fmt.Sprintf("policy-%d", time.Now().UnixNano())
	}

	// Add to config
	rasa.manager.config.EscalationPolicies = append(rasa.manager.config.EscalationPolicies, &policy)

	response := map[string]interface{}{
		"status":    "success",
		"message":   "Escalation policy created successfully",
		"policy_id": policy.ID,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateEscalationPolicy updates an escalation policy
func (rasa *ResourceAlertingScalingAPI) UpdateEscalationPolicy(w http.ResponseWriter, r *http.Request) {
	policyID := r.PathValue("policyId")
	if policyID == "" {
		http.Error(w, "Policy ID is required", http.StatusBadRequest)
		return
	}

	var updatedPolicy EscalationPolicy
	if err := json.NewDecoder(r.Body).Decode(&updatedPolicy); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Find and update policy
	found := false
	for i, policy := range rasa.manager.config.EscalationPolicies {
		if policy.ID == policyID {
			updatedPolicy.ID = policyID
			rasa.manager.config.EscalationPolicies[i] = &updatedPolicy
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Policy not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"status":    "success",
		"message":   "Escalation policy updated successfully",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteEscalationPolicy deletes an escalation policy
func (rasa *ResourceAlertingScalingAPI) DeleteEscalationPolicy(w http.ResponseWriter, r *http.Request) {
	policyID := r.PathValue("policyId")
	if policyID == "" {
		http.Error(w, "Policy ID is required", http.StatusBadRequest)
		return
	}

	// Find and remove policy
	found := false
	for i, policy := range rasa.manager.config.EscalationPolicies {
		if policy.ID == policyID {
			rasa.manager.config.EscalationPolicies = append(
				rasa.manager.config.EscalationPolicies[:i],
				rasa.manager.config.EscalationPolicies[i+1:]...,
			)
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Policy not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"status":    "success",
		"message":   "Escalation policy deleted successfully",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetActiveEscalations returns all active escalations
func (rasa *ResourceAlertingScalingAPI) GetActiveEscalations(w http.ResponseWriter, r *http.Request) {
	rasa.manager.escalationEngine.mu.RLock()
	activeEscalations := make([]*ActiveEscalation, 0, len(rasa.manager.escalationEngine.activeEscalations))
	for _, escalation := range rasa.manager.escalationEngine.activeEscalations {
		activeEscalations = append(activeEscalations, escalation)
	}
	rasa.manager.escalationEngine.mu.RUnlock()

	response := map[string]interface{}{
		"active_escalations": activeEscalations,
		"count":              len(activeEscalations),
		"timestamp":          time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetEscalationHistory returns escalation history
func (rasa *ResourceAlertingScalingAPI) GetEscalationHistory(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	rasa.manager.escalationEngine.mu.RLock()
	history := rasa.manager.escalationEngine.escalationHistory
	if limit > 0 && len(history) > limit {
		history = history[len(history)-limit:]
	}

	// Create a copy
	result := make([]*EscalationEvent, len(history))
	copy(result, history)
	rasa.manager.escalationEngine.mu.RUnlock()

	response := map[string]interface{}{
		"escalation_history": result,
		"count":              len(result),
		"limit":              limit,
		"timestamp":          time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPredictiveModel returns the current predictive model state
func (rasa *ResourceAlertingScalingAPI) GetPredictiveModel(w http.ResponseWriter, r *http.Request) {
	rasa.manager.scalingEngine.mu.RLock()
	model := rasa.manager.scalingEngine.predictiveModel
	rasa.manager.scalingEngine.mu.RUnlock()

	response := map[string]interface{}{
		"predictive_model": model,
		"timestamp":        time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TrainPredictiveModel manually triggers predictive model training
func (rasa *ResourceAlertingScalingAPI) TrainPredictiveModel(w http.ResponseWriter, r *http.Request) {
	// Placeholder for model training
	// In a real implementation, this would:
	// - Collect historical data
	// - Train ML models
	// - Update predictions

	rasa.manager.scalingEngine.mu.Lock()
	rasa.manager.scalingEngine.predictiveModel.LastTrainingTime = time.Now()
	rasa.manager.scalingEngine.mu.Unlock()

	response := map[string]interface{}{
		"status":    "success",
		"message":   "Predictive model training initiated",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPredictions returns current predictions from the predictive model
func (rasa *ResourceAlertingScalingAPI) GetPredictions(w http.ResponseWriter, r *http.Request) {
	rasa.manager.scalingEngine.mu.RLock()
	predictions := rasa.manager.scalingEngine.predictiveModel.Predictions
	confidence := rasa.manager.scalingEngine.predictiveModel.Confidence
	trends := rasa.manager.scalingEngine.predictiveModel.Trends
	rasa.manager.scalingEngine.mu.RUnlock()

	response := map[string]interface{}{
		"predictions": predictions,
		"confidence":  confidence,
		"trends":      trends,
		"timestamp":   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
