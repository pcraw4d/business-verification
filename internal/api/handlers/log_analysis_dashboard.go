package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"go.uber.org/zap"
)

// LogAnalysisDashboardHandler handles log analysis and monitoring dashboard API endpoints
type LogAnalysisDashboardHandler struct {
	logger              *observability.Logger
	logAnalysis         *observability.LogAnalysisSystem
	monitoringDashboard *observability.LogMonitoringDashboard
}

// NewLogAnalysisDashboardHandler creates a new log analysis dashboard handler
func NewLogAnalysisDashboardHandler(
	logger *observability.Logger,
	logAnalysis *observability.LogAnalysisSystem,
	monitoringDashboard *observability.LogMonitoringDashboard,
) *LogAnalysisDashboardHandler {
	return &LogAnalysisDashboardHandler{
		logger:              logger,
		logAnalysis:         logAnalysis,
		monitoringDashboard: monitoringDashboard,
	}
}

// GetDashboardData returns complete dashboard data
func (h *LogAnalysisDashboardHandler) GetDashboardData(w http.ResponseWriter, r *http.Request) {
	data, err := h.monitoringDashboard.GetDashboardData(r.Context())
	if err != nil {
		h.logger.Error("failed to get dashboard data", zap.Error(err))
		http.Error(w, "Failed to get dashboard data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// GetLogAnalysisResults returns log analysis results
func (h *LogAnalysisDashboardHandler) GetLogAnalysisResults(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	timeRange := r.URL.Query().Get("time_range")
	if timeRange == "" {
		timeRange = "1h"
	}

	// For now, return mock log analysis results
	// In a real implementation, this would analyze actual logs
	mockLogs := h.generateMockLogs()

	result, err := h.logAnalysis.AnalyzeLogs(r.Context(), mockLogs)
	if err != nil {
		h.logger.Error("failed to analyze logs", zap.Error(err))
		http.Error(w, "Failed to analyze logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetActivePatterns returns currently active log patterns
func (h *LogAnalysisDashboardHandler) GetActivePatterns(w http.ResponseWriter, r *http.Request) {
	patterns := h.logAnalysis.GetActivePatterns()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"patterns":  patterns,
		"count":     len(patterns),
		"timestamp": time.Now(),
	})
}

// GetActiveErrorGroups returns currently active error groups
func (h *LogAnalysisDashboardHandler) GetActiveErrorGroups(w http.ResponseWriter, r *http.Request) {
	errorGroups := h.logAnalysis.GetActiveErrorGroups()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error_groups": errorGroups,
		"count":        len(errorGroups),
		"timestamp":    time.Now(),
	})
}

// GetCorrelationTraces returns correlation traces
func (h *LogAnalysisDashboardHandler) GetCorrelationTraces(w http.ResponseWriter, r *http.Request) {
	correlationID := r.URL.Query().Get("correlation_id")

	traces := h.logAnalysis.GetCorrelationTraces(correlationID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"correlation_traces": traces,
		"count":              len(traces),
		"correlation_id":     correlationID,
		"timestamp":          time.Now(),
	})
}

// GetAnalysisMetrics returns log analysis metrics
func (h *LogAnalysisDashboardHandler) GetAnalysisMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.logAnalysis.GetAnalysisMetrics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetDashboardOverview returns dashboard overview data
func (h *LogAnalysisDashboardHandler) GetDashboardOverview(w http.ResponseWriter, r *http.Request) {
	data, err := h.monitoringDashboard.GetDashboardData(r.Context())
	if err != nil {
		h.logger.Error("failed to get dashboard data", zap.Error(err))
		http.Error(w, "Failed to get dashboard data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data.Overview)
}

// GetPerformanceData returns performance data
func (h *LogAnalysisDashboardHandler) GetPerformanceData(w http.ResponseWriter, r *http.Request) {
	data, err := h.monitoringDashboard.GetDashboardData(r.Context())
	if err != nil {
		h.logger.Error("failed to get dashboard data", zap.Error(err))
		http.Error(w, "Failed to get dashboard data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data.Performance)
}

// GetHealthStatus returns health status
func (h *LogAnalysisDashboardHandler) GetHealthStatus(w http.ResponseWriter, r *http.Request) {
	data, err := h.monitoringDashboard.GetDashboardData(r.Context())
	if err != nil {
		h.logger.Error("failed to get dashboard data", zap.Error(err))
		http.Error(w, "Failed to get dashboard data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data.Health)
}

// GetRealTimeMetrics returns real-time metrics
func (h *LogAnalysisDashboardHandler) GetRealTimeMetrics(w http.ResponseWriter, r *http.Request) {
	data, err := h.monitoringDashboard.GetDashboardData(r.Context())
	if err != nil {
		h.logger.Error("failed to get dashboard data", zap.Error(err))
		http.Error(w, "Failed to get dashboard data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data.RealTime)
}

// GetActiveAlerts returns active alerts
func (h *LogAnalysisDashboardHandler) GetActiveAlerts(w http.ResponseWriter, r *http.Request) {
	data, err := h.monitoringDashboard.GetDashboardData(r.Context())
	if err != nil {
		h.logger.Error("failed to get dashboard data", zap.Error(err))
		http.Error(w, "Failed to get dashboard data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alerts":    data.Alerts,
		"count":     len(data.Alerts),
		"timestamp": time.Now(),
	})
}

// AddAlert adds a new alert
func (h *LogAnalysisDashboardHandler) AddAlert(w http.ResponseWriter, r *http.Request) {
	var alert observability.LogAlert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		h.logger.Error("failed to decode alert", zap.Error(err))
		http.Error(w, "Invalid alert data", http.StatusBadRequest)
		return
	}

	h.monitoringDashboard.AddAlert(&alert)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Alert added successfully",
		"alert_id":  alert.ID,
		"timestamp": time.Now(),
	})
}

// RemoveAlert removes an alert
func (h *LogAnalysisDashboardHandler) RemoveAlert(w http.ResponseWriter, r *http.Request) {
	alertID := r.URL.Query().Get("alert_id")
	if alertID == "" {
		http.Error(w, "Alert ID is required", http.StatusBadRequest)
		return
	}

	h.monitoringDashboard.RemoveAlert(alertID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Alert removed successfully",
		"alert_id":  alertID,
		"timestamp": time.Now(),
	})
}

// GetCorrelationSummary returns correlation summary
func (h *LogAnalysisDashboardHandler) GetCorrelationSummary(w http.ResponseWriter, r *http.Request) {
	// This would typically come from the correlation tracker
	// For now, return mock data
	summary := map[string]interface{}{
		"total_traces":           150,
		"successful_traces":      135,
		"error_traces":           10,
		"warning_traces":         5,
		"average_duration":       "2.5s",
		"success_rate":           0.90,
		"error_rate":             0.07,
		"average_logs_per_trace": 8.5,
		"timestamp":              time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// GetLogInsights returns log insights
func (h *LogAnalysisDashboardHandler) GetLogInsights(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	severity := r.URL.Query().Get("severity")
	limitStr := r.URL.Query().Get("limit")

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	// Mock insights
	insights := h.generateMockInsights(severity, limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"insights":  insights,
		"count":     len(insights),
		"severity":  severity,
		"limit":     limit,
		"timestamp": time.Now(),
	})
}

// GetDashboardConfiguration returns dashboard configuration
func (h *LogAnalysisDashboardHandler) GetDashboardConfiguration(w http.ResponseWriter, r *http.Request) {
	config := observability.DefaultLogMonitoringDashboardConfig()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// generateMockLogs generates mock log entries for testing
func (h *LogAnalysisDashboardHandler) generateMockLogs() []observability.LogEntry {
	now := time.Now()

	return []observability.LogEntry{
		{
			Timestamp:     now.Add(-5 * time.Minute),
			Level:         "info",
			Message:       "Request processed successfully",
			Logger:        "api",
			TraceID:       "trace-123",
			CorrelationID: "corr-123",
			UserID:        "user-456",
			RequestID:     "req-789",
			Endpoint:      "POST /api/verify",
			Method:        "POST",
			StatusCode:    200,
			Duration:      0.5,
			IPAddress:     "192.168.1.100",
			UserAgent:     "Mozilla/5.0...",
		},
		{
			Timestamp:     now.Add(-4 * time.Minute),
			Level:         "error",
			Message:       "Database connection failed",
			Logger:        "database",
			TraceID:       "trace-123",
			CorrelationID: "corr-123",
			UserID:        "user-456",
			RequestID:     "req-789",
			Endpoint:      "POST /api/verify",
			Method:        "POST",
			StatusCode:    500,
			Duration:      2.0,
			IPAddress:     "192.168.1.100",
			UserAgent:     "Mozilla/5.0...",
			Error:         "connection timeout",
		},
		{
			Timestamp:     now.Add(-3 * time.Minute),
			Level:         "warn",
			Message:       "High response time detected",
			Logger:        "performance",
			TraceID:       "trace-124",
			CorrelationID: "corr-124",
			UserID:        "user-457",
			RequestID:     "req-790",
			Endpoint:      "GET /api/status",
			Method:        "GET",
			StatusCode:    200,
			Duration:      3.5,
			IPAddress:     "192.168.1.101",
			UserAgent:     "Mozilla/5.0...",
		},
		{
			Timestamp:     now.Add(-2 * time.Minute),
			Level:         "info",
			Message:       "Authentication successful",
			Logger:        "auth",
			TraceID:       "trace-125",
			CorrelationID: "corr-125",
			UserID:        "user-458",
			RequestID:     "req-791",
			Endpoint:      "POST /api/auth",
			Method:        "POST",
			StatusCode:    200,
			Duration:      0.8,
			IPAddress:     "192.168.1.102",
			UserAgent:     "Mozilla/5.0...",
		},
		{
			Timestamp:     now.Add(-1 * time.Minute),
			Level:         "error",
			Message:       "Validation failed: invalid email format",
			Logger:        "validation",
			TraceID:       "trace-126",
			CorrelationID: "corr-126",
			UserID:        "user-459",
			RequestID:     "req-792",
			Endpoint:      "POST /api/register",
			Method:        "POST",
			StatusCode:    400,
			Duration:      0.2,
			IPAddress:     "192.168.1.103",
			UserAgent:     "Mozilla/5.0...",
			Error:         "invalid email format",
		},
	}
}

// generateMockInsights generates mock log insights
func (h *LogAnalysisDashboardHandler) generateMockInsights(severity string, limit int) []observability.LogInsight {
	insights := []observability.LogInsight{
		{
			ID:          "insight-1",
			Type:        "high_error_rate",
			Title:       "High Error Rate Detected",
			Description: "Error rate is 15.2% which is above the 10% threshold",
			Severity:    "high",
			Confidence:  0.9,
			Timestamp:   time.Now().Add(-10 * time.Minute),
			Recommendations: []string{
				"Review error patterns and implement fixes",
				"Check system health and resource utilization",
				"Monitor error trends over time",
			},
		},
		{
			ID:          "insight-2",
			Type:        "performance_degradation",
			Title:       "Performance Degradation Detected",
			Description: "8 out of 15 traces are taking longer than 5 seconds",
			Severity:    "medium",
			Confidence:  0.8,
			Timestamp:   time.Now().Add(-15 * time.Minute),
			Recommendations: []string{
				"Investigate slow database queries",
				"Check external service response times",
				"Review resource utilization",
			},
		},
		{
			ID:          "insight-3",
			Type:        "frequent_pattern",
			Title:       "Frequent Database Connection Errors",
			Description: "Pattern 'database connection failed' occurs frequently with high severity",
			Severity:    "high",
			Confidence:  0.85,
			Timestamp:   time.Now().Add(-20 * time.Minute),
			Recommendations: []string{
				"Investigate root cause of database connection issues",
				"Implement connection pooling",
				"Monitor database server health",
			},
		},
	}

	// Filter by severity if specified
	if severity != "" {
		var filteredInsights []observability.LogInsight
		for _, insight := range insights {
			if insight.Severity == severity {
				filteredInsights = append(filteredInsights, insight)
			}
		}
		insights = filteredInsights
	}

	// Limit results
	if len(insights) > limit {
		insights = insights[:limit]
	}

	return insights
}
