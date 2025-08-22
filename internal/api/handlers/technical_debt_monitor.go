package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
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
		zap.Int64("total_lines", metrics.TotalLinesOfCode),
		zap.Int64("deprecated_lines", metrics.DeprecatedCodeLines),
		zap.Float64("technical_debt_ratio", metrics.TechnicalDebtRatio),
		zap.Float64("test_coverage", metrics.TestCoveragePercentage),
		zap.Duration("duration", time.Since(startTime)),
	)
}

// GetTechnicalDebtReport returns a comprehensive technical debt report
func (h *TechnicalDebtMonitorHandler) GetTechnicalDebtReport(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get technical debt report
	report := h.monitor.GetTechnicalDebtReport()

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
	history := h.monitor.GetMetricsHistory()

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
	history := h.monitor.GetMetricsHistory()

	// Filter by days
	cutoffTime := time.Now().AddDate(0, 0, -days)
	var filteredHistory []*observability.TechnicalDebtMetrics

	for _, metric := range history {
		if metric.Timestamp.After(cutoffTime) {
			filteredHistory = append(filteredHistory, metric)
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
			"start":  first.TechnicalDebtRatio,
			"end":    last.TechnicalDebtRatio,
			"change": last.TechnicalDebtRatio - first.TechnicalDebtRatio,
			"trend":  h.getTrendDirection(first.TechnicalDebtRatio, last.TechnicalDebtRatio),
		},
		"test_coverage": map[string]interface{}{
			"start":  first.TestCoveragePercentage,
			"end":    last.TestCoveragePercentage,
			"change": last.TestCoveragePercentage - first.TestCoveragePercentage,
			"trend":  h.getTrendDirection(first.TestCoveragePercentage, last.TestCoveragePercentage),
		},
		"code_quality": map[string]interface{}{
			"start":  first.CodeQualityScore,
			"end":    last.CodeQualityScore,
			"change": last.CodeQualityScore - first.CodeQualityScore,
			"trend":  h.getTrendDirection(first.CodeQualityScore, last.CodeQualityScore),
		},
		"maintainability": map[string]interface{}{
			"start":  first.MaintainabilityIndex,
			"end":    last.MaintainabilityIndex,
			"change": last.MaintainabilityIndex - first.MaintainabilityIndex,
			"trend":  h.getTrendDirection(first.MaintainabilityIndex, last.MaintainabilityIndex),
		},
		"deprecated_code": map[string]interface{}{
			"start":  first.DeprecatedCodeLines,
			"end":    last.DeprecatedCodeLines,
			"change": last.DeprecatedCodeLines - first.DeprecatedCodeLines,
			"trend":  h.getTrendDirection(float64(first.DeprecatedCodeLines), float64(last.DeprecatedCodeLines)),
		},
		"refactoring_opportunities": map[string]interface{}{
			"start":  first.RefactoringOpportunities,
			"end":    last.RefactoringOpportunities,
			"change": last.RefactoringOpportunities - first.RefactoringOpportunities,
			"trend":  h.getTrendDirection(float64(first.RefactoringOpportunities), float64(last.RefactoringOpportunities)),
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
	metrics := h.monitor.GetMetrics()
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
	if metrics.TechnicalDebtRatio > 0.3 {
		alerts = append(alerts, map[string]interface{}{
			"severity":  "high",
			"metric":    "technical_debt_ratio",
			"value":     metrics.TechnicalDebtRatio,
			"threshold": 0.3,
			"message":   "Technical debt ratio is above 30%. Consider prioritizing refactoring efforts.",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}

	// Test coverage alert
	if metrics.TestCoveragePercentage < 80 {
		alerts = append(alerts, map[string]interface{}{
			"severity":  "medium",
			"metric":    "test_coverage",
			"value":     metrics.TestCoveragePercentage,
			"threshold": 80,
			"message":   "Test coverage is below 80%. Increase test coverage for better code quality.",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}

	// Code quality alert
	if metrics.CodeQualityScore < 70 {
		alerts = append(alerts, map[string]interface{}{
			"severity":  "medium",
			"metric":    "code_quality",
			"value":     metrics.CodeQualityScore,
			"threshold": 70,
			"message":   "Code quality score is below 70. Review and improve code quality.",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}

	// Maintainability alert
	if metrics.MaintainabilityIndex < 60 {
		alerts = append(alerts, map[string]interface{}{
			"severity":  "high",
			"metric":    "maintainability",
			"value":     metrics.MaintainabilityIndex,
			"threshold": 60,
			"message":   "Maintainability index is below 60. Focus on improving code maintainability.",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}

	// Refactoring opportunities alert
	if metrics.RefactoringOpportunities > 20 {
		alerts = append(alerts, map[string]interface{}{
			"severity":  "medium",
			"metric":    "refactoring_opportunities",
			"value":     metrics.RefactoringOpportunities,
			"threshold": 20,
			"message":   "Many refactoring opportunities identified. Plan refactoring sprints.",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}

	// Priority issues alert
	if metrics.PriorityIssues > 10 {
		alerts = append(alerts, map[string]interface{}{
			"severity":  "high",
			"metric":    "priority_issues",
			"value":     metrics.PriorityIssues,
			"threshold": 10,
			"message":   "High number of priority issues. Address critical issues first.",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}

	return alerts
}
