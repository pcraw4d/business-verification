package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"go.uber.org/zap"
)

// CodeQualityValidatorHandler handles code quality validation API requests
type CodeQualityValidatorHandler struct {
	validator *observability.CodeQualityValidator
	logger    *zap.Logger
}

// NewCodeQualityValidatorHandler creates a new code quality validator handler
func NewCodeQualityValidatorHandler(validator *observability.CodeQualityValidator, logger *zap.Logger) *CodeQualityValidatorHandler {
	return &CodeQualityValidatorHandler{
		validator: validator,
		logger:    logger,
	}
}

// GetCodeQualityMetrics handles GET /api/v3/code-quality/metrics
func (h *CodeQualityValidatorHandler) GetCodeQualityMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	h.logger.Info("Handling code quality metrics request")

	metrics, err := h.validator.ValidateCodeQuality(ctx)
	if err != nil {
		h.logger.Error("Failed to validate code quality", zap.Error(err))
		http.Error(w, "Failed to validate code quality", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"data":      metrics,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCodeQualityReport handles GET /api/v3/code-quality/report
func (h *CodeQualityValidatorHandler) GetCodeQualityReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	h.logger.Info("Handling code quality report request")

	// Get format parameter
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	metrics, err := h.validator.ValidateCodeQuality(ctx)
	if err != nil {
		h.logger.Error("Failed to validate code quality", zap.Error(err))
		http.Error(w, "Failed to validate code quality", http.StatusInternalServerError)
		return
	}

	switch format {
	case "markdown", "md":
		report, err := h.validator.GenerateQualityReport(metrics)
		if err != nil {
			h.logger.Error("Failed to generate quality report", zap.Error(err))
			http.Error(w, "Failed to generate quality report", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/markdown")
		w.Write([]byte(report))

	case "json":
		fallthrough
	default:
		response := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"metrics": metrics,
				"report": map[string]interface{}{
					"summary": map[string]interface{}{
						"quality_score":         metrics.CodeQualityScore,
						"maintainability_index": metrics.MaintainabilityIndex,
						"technical_debt_ratio":  metrics.TechnicalDebtRatio,
						"test_coverage":         metrics.TestCoverage,
						"improvement_score":     metrics.ImprovementScore,
						"trend_direction":       metrics.TrendDirection,
					},
					"recommendations": h.generateRecommendations(metrics),
					"trends":          h.generateTrends(),
				},
			},
			"timestamp": time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// GetCodeQualityHistory handles GET /api/v3/code-quality/history
func (h *CodeQualityValidatorHandler) GetCodeQualityHistory(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling code quality history request")

	// Get limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 30 // default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	history := h.validator.GetMetricsHistory()

	// Apply limit
	if len(history) > limit {
		history = history[len(history)-limit:]
	}

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"history": history,
			"count":   len(history),
			"limit":   limit,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCodeQualityTrends handles GET /api/v3/code-quality/trends
func (h *CodeQualityValidatorHandler) GetCodeQualityTrends(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling code quality trends request")

	// Get period parameter
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "7d" // default to 7 days
	}

	history := h.validator.GetMetricsHistory()

	// Calculate trends based on period
	trends := h.calculateTrends(history, period)

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"trends": trends,
			"period": period,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TriggerCodeQualityValidation handles POST /api/v3/code-quality/validate
func (h *CodeQualityValidatorHandler) TriggerCodeQualityValidation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	h.logger.Info("Handling code quality validation trigger request")

	// Parse request body for options
	var request struct {
		IncludePatterns []string `json:"include_patterns"`
		ExcludePatterns []string `json:"exclude_patterns"`
		GenerateReport  bool     `json:"generate_report"`
	}

	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			h.logger.Warn("Failed to decode request body", zap.Error(err))
		}
	}

	// Perform validation
	metrics, err := h.validator.ValidateCodeQuality(ctx)
	if err != nil {
		h.logger.Error("Failed to validate code quality", zap.Error(err))
		http.Error(w, "Failed to validate code quality", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"metrics": metrics,
			"validation": map[string]interface{}{
				"status":        "completed",
				"duration":      metrics.ScanDuration,
				"files_scanned": metrics.TotalFiles,
				"timestamp":     metrics.Timestamp.Format(time.RFC3339),
			},
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Include report if requested
	if request.GenerateReport {
		report, err := h.validator.GenerateQualityReport(metrics)
		if err == nil {
			response["data"].(map[string]interface{})["report"] = report
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCodeQualityAlerts handles GET /api/v3/code-quality/alerts
func (h *CodeQualityValidatorHandler) GetCodeQualityAlerts(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling code quality alerts request")

	// Get severity parameter
	severity := r.URL.Query().Get("severity")
	if severity == "" {
		severity = "all"
	}

	// Get latest metrics
	ctx := r.Context()
	metrics, err := h.validator.ValidateCodeQuality(ctx)
	if err != nil {
		h.logger.Error("Failed to validate code quality for alerts", zap.Error(err))
		http.Error(w, "Failed to validate code quality", http.StatusInternalServerError)
		return
	}

	// Generate alerts based on metrics
	alerts := h.generateAlerts(metrics, severity)

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"alerts":   alerts,
			"count":    len(alerts),
			"severity": severity,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper methods

func (h *CodeQualityValidatorHandler) generateRecommendations(metrics *observability.CodeQualityMetrics) []string {
	var recommendations []string

	if metrics.CyclomaticComplexity > 10 {
		recommendations = append(recommendations, "High cyclomatic complexity detected. Consider refactoring complex functions.")
	}

	if metrics.AverageFunctionSize > 30 {
		recommendations = append(recommendations, "Large average function size. Break down large functions into smaller, focused functions.")
	}

	if metrics.TestCoverage < 80 {
		recommendations = append(recommendations, "Test coverage below 80%. Increase test coverage for better code quality.")
	}

	if metrics.TechnicalDebtRatio > 0.3 {
		recommendations = append(recommendations, "High technical debt ratio. Prioritize debt reduction in upcoming sprints.")
	}

	if metrics.CodeSmells > 10 {
		recommendations = append(recommendations, "Multiple code smells detected. Review and refactor problematic code.")
	}

	if metrics.DocumentationCoverage < 70 {
		recommendations = append(recommendations, "Low documentation coverage. Improve code documentation.")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Code quality is good. Continue maintaining current standards.")
	}

	return recommendations
}

func (h *CodeQualityValidatorHandler) generateTrends() map[string]interface{} {
	history := h.validator.GetMetricsHistory()

	if len(history) < 2 {
		return map[string]interface{}{
			"status":  "insufficient_data",
			"message": "Insufficient historical data for trend analysis",
		}
	}

	recent := history[len(history)-1]
	previous := history[len(history)-2]

	return map[string]interface{}{
		"status": "available",
		"changes": map[string]interface{}{
			"quality_score": map[string]interface{}{
				"current":  recent.CodeQualityScore,
				"previous": previous.CodeQualityScore,
				"change":   recent.CodeQualityScore - previous.CodeQualityScore,
			},
			"maintainability": map[string]interface{}{
				"current":  recent.MaintainabilityIndex,
				"previous": previous.MaintainabilityIndex,
				"change":   recent.MaintainabilityIndex - previous.MaintainabilityIndex,
			},
			"test_coverage": map[string]interface{}{
				"current":  recent.TestCoverage,
				"previous": previous.TestCoverage,
				"change":   recent.TestCoverage - previous.TestCoverage,
			},
			"technical_debt": map[string]interface{}{
				"current":  recent.TechnicalDebtRatio,
				"previous": previous.TechnicalDebtRatio,
				"change":   previous.TechnicalDebtRatio - recent.TechnicalDebtRatio, // Lower is better
			},
		},
		"trend_direction":   recent.TrendDirection,
		"improvement_score": recent.ImprovementScore,
	}
}

func (h *CodeQualityValidatorHandler) calculateTrends(history []observability.CodeQualityMetrics, period string) map[string]interface{} {
	if len(history) < 2 {
		return map[string]interface{}{
			"status":  "insufficient_data",
			"message": "Insufficient historical data for trend analysis",
		}
	}

	// Filter history based on period
	var filteredHistory []observability.CodeQualityMetrics
	now := time.Now()

	for _, metric := range history {
		switch period {
		case "1d":
			if now.Sub(metric.Timestamp) <= 24*time.Hour {
				filteredHistory = append(filteredHistory, metric)
			}
		case "7d":
			if now.Sub(metric.Timestamp) <= 7*24*time.Hour {
				filteredHistory = append(filteredHistory, metric)
			}
		case "30d":
			if now.Sub(metric.Timestamp) <= 30*24*time.Hour {
				filteredHistory = append(filteredHistory, metric)
			}
		default:
			filteredHistory = append(filteredHistory, metric)
		}
	}

	if len(filteredHistory) < 2 {
		return map[string]interface{}{
			"status":  "insufficient_data",
			"message": "Insufficient data for the specified period",
		}
	}

	// Calculate trends
	first := filteredHistory[0]
	last := filteredHistory[len(filteredHistory)-1]

	return map[string]interface{}{
		"status": "available",
		"period": period,
		"trends": map[string]interface{}{
			"quality_score": map[string]interface{}{
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
			"test_coverage": map[string]interface{}{
				"start":  first.TestCoverage,
				"end":    last.TestCoverage,
				"change": last.TestCoverage - first.TestCoverage,
				"trend":  h.getTrendDirection(first.TestCoverage, last.TestCoverage),
			},
			"technical_debt": map[string]interface{}{
				"start":  first.TechnicalDebtRatio,
				"end":    last.TechnicalDebtRatio,
				"change": first.TechnicalDebtRatio - last.TechnicalDebtRatio,                     // Lower is better
				"trend":  h.getTrendDirection(last.TechnicalDebtRatio, first.TechnicalDebtRatio), // Reversed
			},
		},
		"data_points": len(filteredHistory),
	}
}

func (h *CodeQualityValidatorHandler) getTrendDirection(start, end float64) string {
	change := end - start
	if change > 1.0 {
		return "improving"
	} else if change < -1.0 {
		return "declining"
	}
	return "stable"
}

func (h *CodeQualityValidatorHandler) generateAlerts(metrics *observability.CodeQualityMetrics, severity string) []map[string]interface{} {
	var alerts []map[string]interface{}

	// Critical alerts
	if severity == "all" || severity == "critical" {
		if metrics.TechnicalDebtRatio > 0.5 {
			alerts = append(alerts, map[string]interface{}{
				"severity":  "critical",
				"title":     "High Technical Debt",
				"message":   "Technical debt ratio is above 50%",
				"value":     metrics.TechnicalDebtRatio,
				"threshold": 0.5,
			})
		}

		if metrics.TestCoverage < 50 {
			alerts = append(alerts, map[string]interface{}{
				"severity":  "critical",
				"title":     "Low Test Coverage",
				"message":   "Test coverage is below 50%",
				"value":     metrics.TestCoverage,
				"threshold": 50.0,
			})
		}
	}

	// High severity alerts
	if severity == "all" || severity == "high" {
		if metrics.CyclomaticComplexity > 15 {
			alerts = append(alerts, map[string]interface{}{
				"severity":  "high",
				"title":     "High Complexity",
				"message":   "Cyclomatic complexity is above 15",
				"value":     metrics.CyclomaticComplexity,
				"threshold": 15.0,
			})
		}

		if metrics.CodeQualityScore < 60 {
			alerts = append(alerts, map[string]interface{}{
				"severity":  "high",
				"title":     "Low Code Quality",
				"message":   "Code quality score is below 60",
				"value":     metrics.CodeQualityScore,
				"threshold": 60.0,
			})
		}
	}

	// Medium severity alerts
	if severity == "all" || severity == "medium" {
		if metrics.AverageFunctionSize > 50 {
			alerts = append(alerts, map[string]interface{}{
				"severity":  "medium",
				"title":     "Large Functions",
				"message":   "Average function size is above 50 lines",
				"value":     metrics.AverageFunctionSize,
				"threshold": 50.0,
			})
		}

		if metrics.DocumentationCoverage < 60 {
			alerts = append(alerts, map[string]interface{}{
				"severity":  "medium",
				"title":     "Low Documentation",
				"message":   "Documentation coverage is below 60%",
				"value":     metrics.DocumentationCoverage,
				"threshold": 60.0,
			})
		}
	}

	// Low severity alerts
	if severity == "all" || severity == "low" {
		if metrics.CodeSmells > 5 {
			alerts = append(alerts, map[string]interface{}{
				"severity":  "low",
				"title":     "Code Smells",
				"message":   "Multiple code smells detected",
				"value":     metrics.CodeSmells,
				"threshold": 5.0,
			})
		}

		if metrics.CommentRatio < 10 {
			alerts = append(alerts, map[string]interface{}{
				"severity":  "low",
				"title":     "Low Comment Ratio",
				"message":   "Comment ratio is below 10%",
				"value":     metrics.CommentRatio,
				"threshold": 10.0,
			})
		}
	}

	return alerts
}
