package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/internal/modules/classification_monitoring"
)

// ClassificationMonitoringHandler handles HTTP requests for classification monitoring
type ClassificationMonitoringHandler struct {
	accuracyTracker           *classification_monitoring.AccuracyTracker
	misclassificationDetector *classification_monitoring.MisclassificationDetector
	metricsCollector          *classification_monitoring.AccuracyMetricsCollector
	alertingSystem            *classification_monitoring.AccuracyAlertingSystem
	logger                    *zap.Logger
}

// NewClassificationMonitoringHandler creates a new monitoring handler
func NewClassificationMonitoringHandler(
	accuracyTracker *classification_monitoring.AccuracyTracker,
	misclassificationDetector *classification_monitoring.MisclassificationDetector,
	metricsCollector *classification_monitoring.AccuracyMetricsCollector,
	alertingSystem *classification_monitoring.AccuracyAlertingSystem,
	logger *zap.Logger,
) *ClassificationMonitoringHandler {
	return &ClassificationMonitoringHandler{
		accuracyTracker:           accuracyTracker,
		misclassificationDetector: misclassificationDetector,
		metricsCollector:          metricsCollector,
		alertingSystem:            alertingSystem,
		logger:                    logger,
	}
}

// GetAccuracyMetrics returns current accuracy metrics
func (cmh *ClassificationMonitoringHandler) GetAccuracyMetrics(w http.ResponseWriter, r *http.Request) {

	cmh.logger.Info("Getting accuracy metrics")

	// Get query parameters
	dimension := r.URL.Query().Get("dimension")
	includeHistorical := r.URL.Query().Get("include_historical") == "true"

	var response map[string]interface{}

	if dimension != "" {
		// Get specific dimension metrics
		metrics := cmh.accuracyTracker.GetAccuracyMetrics()
		if dimMetrics, exists := metrics[dimension]; exists {
			response = map[string]interface{}{
				"dimension": dimension,
				"metrics":   dimMetrics,
			}
		} else {
			http.Error(w, "Dimension not found", http.StatusNotFound)
			return
		}
	} else {
		// Get all metrics
		metrics := cmh.accuracyTracker.GetAccuracyMetrics()
		response = map[string]interface{}{
			"metrics": metrics,
		}

		// Include historical data if requested
		if includeHistorical {
			since := time.Now().Add(-24 * time.Hour) // Last 24 hours by default
			if sinceParam := r.URL.Query().Get("since"); sinceParam != "" {
				if parsedTime, err := time.Parse(time.RFC3339, sinceParam); err == nil {
					since = parsedTime
				}
			}

			historical := cmh.accuracyTracker.GetHistoricalData(since)
			response["historical"] = historical
		}
	}

	// Add metadata
	response["timestamp"] = time.Now()
	response["overall_accuracy"] = cmh.accuracyTracker.GetOverallAccuracy()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		cmh.logger.Error("Failed to encode accuracy metrics response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetMisclassifications returns misclassification records
func (cmh *ClassificationMonitoringHandler) GetMisclassifications(w http.ResponseWriter, r *http.Request) {

	cmh.logger.Info("Getting misclassifications")

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	var misclassifications interface{}
	if startTimeStr != "" && endTimeStr != "" {
		// Get misclassifications by time range
		startTime, err1 := time.Parse(time.RFC3339, startTimeStr)
		endTime, err2 := time.Parse(time.RFC3339, endTimeStr)

		if err1 != nil || err2 != nil {
			http.Error(w, "Invalid time format. Use RFC3339 format", http.StatusBadRequest)
			return
		}

		misclassifications = cmh.misclassificationDetector.GetMisclassificationsByTimeRange(startTime, endTime)
	} else {
		// Get recent misclassifications
		misclassifications = cmh.accuracyTracker.GetMisclassifications(limit)
	}

	response := map[string]interface{}{
		"misclassifications": misclassifications,
		"timestamp":          time.Now(),
		"limit":              limit,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		cmh.logger.Error("Failed to encode misclassifications response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetErrorPatterns returns detected error patterns
func (cmh *ClassificationMonitoringHandler) GetErrorPatterns(w http.ResponseWriter, r *http.Request) {
	cmh.logger.Info("Getting error patterns")

	// Parse query parameters
	severity := r.URL.Query().Get("severity")
	patternType := r.URL.Query().Get("type")

	patterns := cmh.misclassificationDetector.GetDetectedPatterns()

	// Filter patterns if requested
	filteredPatterns := make(map[string]*classification_monitoring.ErrorPattern)
	for id, pattern := range patterns {
		include := true

		if severity != "" && pattern.Severity != severity {
			include = false
		}

		if patternType != "" && pattern.Type != patternType {
			include = false
		}

		if include {
			filteredPatterns[id] = pattern
		}
	}

	response := map[string]interface{}{
		"patterns":  filteredPatterns,
		"timestamp": time.Now(),
		"filters": map[string]interface{}{
			"severity": severity,
			"type":     patternType,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		cmh.logger.Error("Failed to encode error patterns response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetErrorStatistics returns comprehensive error statistics
func (cmh *ClassificationMonitoringHandler) GetErrorStatistics(w http.ResponseWriter, r *http.Request) {
	cmh.logger.Info("Getting error statistics")

	statistics := cmh.misclassificationDetector.GetErrorStatistics()

	response := map[string]interface{}{
		"statistics": statistics,
		"timestamp":  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		cmh.logger.Error("Failed to encode error statistics response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// TrackClassification manually tracks a classification result
func (cmh *ClassificationMonitoringHandler) TrackClassification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cmh.logger.Info("Tracking classification")

	var request classification_monitoring.ClassificationResult
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		cmh.logger.Error("Failed to decode track classification request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if request.BusinessName == "" || request.ActualClassification == "" {
		http.Error(w, "business_name and actual_classification are required", http.StatusBadRequest)
		return
	}

	// Set timestamp if not provided
	if request.Timestamp.IsZero() {
		request.Timestamp = time.Now()
	}

	// Track the classification
	if err := cmh.accuracyTracker.TrackClassification(ctx, &request); err != nil {
		cmh.logger.Error("Failed to track classification", zap.Error(err))
		http.Error(w, "Failed to track classification", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message":   "Classification tracked successfully",
		"id":        request.ID,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		cmh.logger.Error("Failed to encode track classification response", zap.Error(err))
	}
}

// GetActiveAlerts returns currently active alerts
func (cmh *ClassificationMonitoringHandler) GetActiveAlerts(w http.ResponseWriter, r *http.Request) {
	cmh.logger.Info("Getting active alerts")

	var alerts interface{}

	if cmh.alertingSystem != nil {
		alerts = cmh.alertingSystem.GetActiveAlerts()
	} else {
		alerts = cmh.accuracyTracker.GetActiveAlerts()
	}

	response := map[string]interface{}{
		"alerts":    alerts,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		cmh.logger.Error("Failed to encode active alerts response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetAlertHistory returns alert history
func (cmh *ClassificationMonitoringHandler) GetAlertHistory(w http.ResponseWriter, r *http.Request) {
	cmh.logger.Info("Getting alert history")

	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	var alerts interface{}

	if cmh.alertingSystem != nil {
		alerts = cmh.alertingSystem.GetAlertHistory(limit)
	} else {
		// Fallback to accuracy tracker if alerting system not available
		alerts = []interface{}{} // Empty for now
	}

	response := map[string]interface{}{
		"alerts":    alerts,
		"limit":     limit,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		cmh.logger.Error("Failed to encode alert history response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ResolveAlert resolves an active alert
func (cmh *ClassificationMonitoringHandler) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	cmh.logger.Info("Resolving alert")

	vars := mux.Vars(r)
	alertID := vars["alertId"]

	if alertID == "" {
		http.Error(w, "Alert ID is required", http.StatusBadRequest)
		return
	}

	// Try to resolve the alert
	err := cmh.accuracyTracker.ResolveAlert(alertID)
	if err != nil {
		cmh.logger.Error("Failed to resolve alert", zap.String("alert_id", alertID), zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"message":   "Alert resolved successfully",
		"alert_id":  alertID,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		cmh.logger.Error("Failed to encode resolve alert response", zap.Error(err))
	}
}

// CollectMetrics triggers metrics collection
func (cmh *ClassificationMonitoringHandler) CollectMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cmh.logger.Info("Collecting metrics")

	if cmh.metricsCollector == nil {
		http.Error(w, "Metrics collector not available", http.StatusServiceUnavailable)
		return
	}

	// Collect and aggregate metrics
	result, err := cmh.metricsCollector.CollectAndAggregateMetrics(ctx)
	if err != nil {
		cmh.logger.Error("Failed to collect metrics", zap.Error(err))
		http.Error(w, "Failed to collect metrics", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message":         "Metrics collected successfully",
		"collection_time": result.LastUpdated,
		"metrics_count":   len(result.DimensionalMetrics),
		"period": map[string]interface{}{
			"start_time": result.StartTime,
			"end_time":   result.EndTime,
			"duration":   result.AggregationPeriod,
		},
		"overall_accuracy": func() float64 {
			if result.OverallMetrics != nil {
				return result.OverallMetrics.AccuracyRate
			}
			return 0.0
		}(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		cmh.logger.Error("Failed to encode collect metrics response", zap.Error(err))
	}
}

// GenerateReport generates an accuracy report
func (cmh *ClassificationMonitoringHandler) GenerateReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cmh.logger.Info("Generating accuracy report")

	if cmh.alertingSystem == nil || cmh.metricsCollector == nil {
		http.Error(w, "Reporting system not available", http.StatusServiceUnavailable)
		return
	}

	// Parse period parameters
	periodStr := r.URL.Query().Get("period")
	var period classification_monitoring.ReportPeriod

	switch periodStr {
	case "hour":
		period = classification_monitoring.ReportPeriod{
			StartTime: time.Now().Add(-1 * time.Hour),
			EndTime:   time.Now(),
			Duration:  1 * time.Hour,
			Label:     "Last Hour",
		}
	case "day":
		period = classification_monitoring.ReportPeriod{
			StartTime: time.Now().Add(-24 * time.Hour),
			EndTime:   time.Now(),
			Duration:  24 * time.Hour,
			Label:     "Last 24 Hours",
		}
	case "week":
		period = classification_monitoring.ReportPeriod{
			StartTime: time.Now().Add(-7 * 24 * time.Hour),
			EndTime:   time.Now(),
			Duration:  7 * 24 * time.Hour,
			Label:     "Last Week",
		}
	default:
		// Default to last 24 hours
		period = classification_monitoring.ReportPeriod{
			StartTime: time.Now().Add(-24 * time.Hour),
			EndTime:   time.Now(),
			Duration:  24 * time.Hour,
			Label:     "Last 24 Hours",
		}
	}

	// Collect current metrics
	metrics, err := cmh.metricsCollector.CollectAndAggregateMetrics(ctx)
	if err != nil {
		cmh.logger.Error("Failed to collect metrics for report", zap.Error(err))
		http.Error(w, "Failed to collect metrics for report", http.StatusInternalServerError)
		return
	}

	// Generate report
	report, err := cmh.alertingSystem.GenerateReport(ctx, period, metrics)
	if err != nil {
		cmh.logger.Error("Failed to generate report", zap.Error(err))
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		cmh.logger.Error("Failed to encode report response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetAlertRules returns configured alert rules
func (cmh *ClassificationMonitoringHandler) GetAlertRules(w http.ResponseWriter, r *http.Request) {
	cmh.logger.Info("Getting alert rules")

	if cmh.alertingSystem == nil {
		http.Error(w, "Alerting system not available", http.StatusServiceUnavailable)
		return
	}

	rules := cmh.alertingSystem.GetAlertRules()

	response := map[string]interface{}{
		"rules":     rules,
		"count":     len(rules),
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		cmh.logger.Error("Failed to encode alert rules response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// CreateAlertRule creates a new alert rule
func (cmh *ClassificationMonitoringHandler) CreateAlertRule(w http.ResponseWriter, r *http.Request) {
	cmh.logger.Info("Creating alert rule")

	if cmh.alertingSystem == nil {
		http.Error(w, "Alerting system not available", http.StatusServiceUnavailable)
		return
	}

	var rule classification_monitoring.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		cmh.logger.Error("Failed to decode alert rule request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if rule.Name == "" || rule.Threshold == 0 || rule.ComparisonOperator == "" {
		http.Error(w, "name, threshold, and comparison_operator are required", http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if rule.ID == "" {
		rule.ID = fmt.Sprintf("rule_%d", time.Now().UnixNano())
	}

	if err := cmh.alertingSystem.AddAlertRule(&rule); err != nil {
		cmh.logger.Error("Failed to add alert rule", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"message":   "Alert rule created successfully",
		"rule_id":   rule.ID,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		cmh.logger.Error("Failed to encode create alert rule response", zap.Error(err))
	}
}

// UpdateAlertRule updates an existing alert rule
func (cmh *ClassificationMonitoringHandler) UpdateAlertRule(w http.ResponseWriter, r *http.Request) {
	cmh.logger.Info("Updating alert rule")

	if cmh.alertingSystem == nil {
		http.Error(w, "Alerting system not available", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	ruleID := vars["ruleId"]

	if ruleID == "" {
		http.Error(w, "Rule ID is required", http.StatusBadRequest)
		return
	}

	var rule classification_monitoring.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		cmh.logger.Error("Failed to decode alert rule update request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := cmh.alertingSystem.UpdateAlertRule(ruleID, &rule); err != nil {
		cmh.logger.Error("Failed to update alert rule", zap.String("rule_id", ruleID), zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"message":   "Alert rule updated successfully",
		"rule_id":   ruleID,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		cmh.logger.Error("Failed to encode update alert rule response", zap.Error(err))
	}
}

// DeleteAlertRule deletes an alert rule
func (cmh *ClassificationMonitoringHandler) DeleteAlertRule(w http.ResponseWriter, r *http.Request) {
	cmh.logger.Info("Deleting alert rule")

	if cmh.alertingSystem == nil {
		http.Error(w, "Alerting system not available", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	ruleID := vars["ruleId"]

	if ruleID == "" {
		http.Error(w, "Rule ID is required", http.StatusBadRequest)
		return
	}

	if err := cmh.alertingSystem.DeleteAlertRule(ruleID); err != nil {
		cmh.logger.Error("Failed to delete alert rule", zap.String("rule_id", ruleID), zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"message":   "Alert rule deleted successfully",
		"rule_id":   ruleID,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		cmh.logger.Error("Failed to encode delete alert rule response", zap.Error(err))
	}
}

// GetHealthStatus returns the health status of the monitoring system
func (cmh *ClassificationMonitoringHandler) GetHealthStatus(w http.ResponseWriter, r *http.Request) {
	cmh.logger.Info("Getting monitoring system health status")

	status := map[string]interface{}{
		"timestamp": time.Now(),
		"status":    "healthy",
		"components": map[string]interface{}{
			"accuracy_tracker": map[string]interface{}{
				"status":           "healthy",
				"overall_accuracy": cmh.accuracyTracker.GetOverallAccuracy(),
				"active_alerts":    len(cmh.accuracyTracker.GetActiveAlerts()),
			},
			"misclassification_detector": map[string]interface{}{
				"status":            "healthy",
				"detected_patterns": len(cmh.misclassificationDetector.GetDetectedPatterns()),
			},
		},
	}

	if cmh.metricsCollector != nil {
		status["components"].(map[string]interface{})["metrics_collector"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	if cmh.alertingSystem != nil {
		status["components"].(map[string]interface{})["alerting_system"] = map[string]interface{}{
			"status":        "healthy",
			"active_alerts": len(cmh.alertingSystem.GetActiveAlerts()),
			"alert_rules":   len(cmh.alertingSystem.GetAlertRules()),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		cmh.logger.Error("Failed to encode health status response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
