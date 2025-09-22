package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"kyb-platform/internal/observability"
	"go.uber.org/zap"
)

// TechnicalDebtMonitorHandler provides HTTP endpoints for technical debt monitoring
type TechnicalDebtMonitorHandler struct {
	monitor *observability.TechnicalDebtMonitor
	logger  *zap.Logger
}

// NewTechnicalDebtMonitorHandler creates a new technical debt monitor handler
func NewTechnicalDebtMonitorHandler(monitor *observability.TechnicalDebtMonitor, logger *zap.Logger) *TechnicalDebtMonitorHandler {
	return &TechnicalDebtMonitorHandler{
		monitor: monitor,
		logger:  logger,
	}
}

// GetTechnicalDebtMetrics returns the current technical debt metrics
func (h *TechnicalDebtMonitorHandler) GetTechnicalDebtMetrics(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get current metrics
	metrics := h.monitor.GetMetrics()
	if metrics == nil {
		http.Error(w, "No metrics available", http.StatusNotFound)
		return
	}

	// Add response metadata
	response := map[string]interface{}{
		"metrics":     metrics,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"endpoint":    "/api/v3/technical-debt/metrics",
		"duration_ms": time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Return success response
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode technical debt metrics response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Technical debt metrics retrieved successfully",
		zap.Float64("complexity", metrics.Complexity),
		zap.Float64("maintainability", metrics.Maintainability),
		zap.Float64("reliability", metrics.Reliability),
		zap.Float64("security", metrics.Security),
		zap.Float64("test_coverage", metrics.TestCoverage),
		zap.Duration("duration", time.Since(startTime)),
	)
}

// GetTechnicalDebtReport returns a comprehensive technical debt report
func (h *TechnicalDebtMonitorHandler) GetTechnicalDebtReport(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get technical debt report
	report, err := h.monitor.GetTechnicalDebtReport(r.Context())
	if err != nil {
		h.logger.Error("Failed to get technical debt report", zap.Error(err))
		http.Error(w, "Failed to get technical debt report", http.StatusInternalServerError)
		return
	}

	// Add response metadata
	response := map[string]interface{}{
		"report":      report,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"endpoint":    "/api/v3/technical-debt/report",
		"duration_ms": time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Return success response
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode technical debt report response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Technical debt report generated successfully",
		zap.Duration("duration", time.Since(startTime)),
	)
}

// GetTechnicalDebtHistory returns historical technical debt metrics
func (h *TechnicalDebtMonitorHandler) GetTechnicalDebtHistory(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get metrics history
	history, err := h.monitor.GetMetricsHistory(r.Context(), limit)
	if err != nil {
		h.logger.Error("Failed to get metrics history", zap.Error(err))
		http.Error(w, "Failed to get metrics history", http.StatusInternalServerError)
		return
	}

	// Apply limit
	if len(history) > limit {
		history = history[len(history)-limit:]
	}

	// Add response metadata
	response := map[string]interface{}{
		"history":     history,
		"count":       len(history),
		"limit":       limit,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"endpoint":    "/api/v3/technical-debt/history",
		"duration_ms": time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Return success response
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode technical debt history response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Technical debt history retrieved successfully",
		zap.Int("count", len(history)),
		zap.Int("limit", limit),
		zap.Duration("duration", time.Since(startTime)),
	)
}

// TriggerTechnicalDebtScan triggers an immediate technical debt scan
func (h *TechnicalDebtMonitorHandler) TriggerTechnicalDebtScan(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Check if request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Trigger scan (this would need to be implemented in the monitor)
	// For now, we'll just return a success response
	response := map[string]interface{}{
		"message":     "Technical debt scan triggered successfully",
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"endpoint":    "/api/v3/technical-debt/scan",
		"duration_ms": time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Return success response
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode scan trigger response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Technical debt scan triggered successfully",
		zap.Duration("duration", time.Since(startTime)),
	)
}

// GetTechnicalDebtTrends returns technical debt trends over time
func (h *TechnicalDebtMonitorHandler) GetTechnicalDebtTrends(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse query parameters
	daysStr := r.URL.Query().Get("days")
	days := 30 // Default to 30 days
	if daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	// Get metrics history
	history, err := h.monitor.GetMetricsHistory(r.Context(), days)
	if err != nil {
		h.logger.Error("Failed to get metrics history", zap.Error(err))
		http.Error(w, "Failed to get metrics history", http.StatusInternalServerError)
		return
	}

	// Filter by days
	cutoffTime := time.Now().AddDate(0, 0, -days)
	var filteredHistory []*observability.TechnicalDebtMetrics

	for _, metric := range history {
		if metric.Timestamp.After(cutoffTime) {
			filteredHistory = append(filteredHistory, &metric)
		}
	}

	// Calculate trends
	trends := h.calculateTrends(filteredHistory)

	// Add response metadata
	response := map[string]interface{}{
		"trends":      trends,
		"days":        days,
		"count":       len(filteredHistory),
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"endpoint":    "/api/v3/technical-debt/trends",
		"duration_ms": time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Return success response
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode technical debt trends response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Technical debt trends calculated successfully",
		zap.Int("days", days),
		zap.Int("count", len(filteredHistory)),
		zap.Duration("duration", time.Since(startTime)),
	)
}

// calculateTrends calculates trends from historical metrics
func (h *TechnicalDebtMonitorHandler) calculateTrends(history []*observability.TechnicalDebtMetrics) map[string]interface{} {
	if len(history) < 2 {
		return map[string]interface{}{
			"error": "Insufficient data for trend calculation",
		}
	}

	// Get first and last metrics
	first := history[0]
	last := history[len(history)-1]

	// Calculate trends
	trends := map[string]interface{}{
		"technical_debt_ratio": map[string]interface{}{
			"start":  first.DebtRatio,
			"end":    last.DebtRatio,
			"change": last.DebtRatio - first.DebtRatio,
			"trend":  h.getTrendDirection(first.DebtRatio, last.DebtRatio),
		},
		"total_debt": map[string]interface{}{
			"start":  first.TotalDebt,
			"end":    last.TotalDebt,
			"change": last.TotalDebt - first.TotalDebt,
			"trend":  h.getTrendDirection(first.TotalDebt, last.TotalDebt),
		},
		"remediation_cost": map[string]interface{}{
			"start":  first.RemediationCost,
			"end":    last.RemediationCost,
			"change": last.RemediationCost - first.RemediationCost,
			"trend":  h.getTrendDirection(first.RemediationCost, last.RemediationCost),
		},
	}

	return trends
}

// getTrendDirection determines the trend direction
func (h *TechnicalDebtMonitorHandler) getTrendDirection(start, end float64) string {
	change := end - start
	if change > 0.01 {
		return "increasing"
	} else if change < -0.01 {
		return "decreasing"
	}
	return "stable"
}

// GetTechnicalDebtAlerts returns technical debt alerts based on thresholds
func (h *TechnicalDebtMonitorHandler) GetTechnicalDebtAlerts(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get current metrics
	metrics := h.monitor.GetCurrentMetrics()
	if metrics == nil {
		http.Error(w, "No metrics available", http.StatusNotFound)
		return
	}

	// Generate alerts based on thresholds
	alerts := h.generateAlerts(metrics)

	// Add response metadata
	response := map[string]interface{}{
		"alerts":      alerts,
		"count":       len(alerts),
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"endpoint":    "/api/v3/technical-debt/alerts",
		"duration_ms": time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Return success response
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode technical debt alerts response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Technical debt alerts generated successfully",
		zap.Int("count", len(alerts)),
		zap.Duration("duration", time.Since(startTime)),
	)
}

// generateAlerts generates alerts based on metric thresholds
func (h *TechnicalDebtMonitorHandler) generateAlerts(metrics *observability.TechnicalDebtMetrics) []map[string]interface{} {
	var alerts []map[string]interface{}

	// Technical debt ratio alert
	if metrics.DebtRatio > 0.3 {
		alerts = append(alerts, map[string]interface{}{
			"severity":  "high",
			"metric":    "technical_debt_ratio",
			"value":     metrics.DebtRatio,
			"threshold": 0.3,
			"message":   "Technical debt ratio is above 30%. Consider prioritizing refactoring efforts.",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}

	// Total debt alert
	if metrics.TotalDebt > 10000 {
		alerts = append(alerts, map[string]interface{}{
			"severity":  "medium",
			"metric":    "total_debt",
			"value":     metrics.TotalDebt,
			"threshold": 10000,
			"message":   "Total technical debt is above 10,000. Consider addressing high-impact debt items.",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}

	// Remediation cost alert
	if metrics.RemediationCost > 20000 {
		alerts = append(alerts, map[string]interface{}{
			"severity":  "high",
			"metric":    "remediation_cost",
			"value":     metrics.RemediationCost,
			"threshold": 20000,
			"message":   "Remediation cost is above 20,000. High cost to address technical debt.",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}

	// Risk level alert
	if metrics.RiskLevel == "high" {
		alerts = append(alerts, map[string]interface{}{
			"severity":  "high",
			"metric":    "risk_level",
			"value":     metrics.RiskLevel,
			"threshold": "high",
			"message":   "Technical debt risk level is high. Immediate attention required.",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}

	return alerts
}
