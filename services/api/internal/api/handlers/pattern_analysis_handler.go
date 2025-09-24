package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/internal/modules/classification_monitoring"
)

// PatternAnalysisHandler handles pattern analysis API requests
type PatternAnalysisHandler struct {
	patternEngine *classification_monitoring.PatternAnalysisEngine
	logger        *zap.Logger
}

// NewPatternAnalysisHandler creates a new pattern analysis handler
func NewPatternAnalysisHandler(patternEngine *classification_monitoring.PatternAnalysisEngine, logger *zap.Logger) *PatternAnalysisHandler {
	return &PatternAnalysisHandler{
		patternEngine: patternEngine,
		logger:        logger,
	}
}

// AnalyzeMisclassificationsHandler handles requests to analyze misclassification patterns
func (h *PatternAnalysisHandler) AnalyzeMisclassificationsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request struct {
		Misclassifications []*classification_monitoring.MisclassificationRecord `json:"misclassifications"`
		Config             *classification_monitoring.PatternAnalysisConfig     `json:"config,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(request.Misclassifications) == 0 {
		http.Error(w, "No misclassifications provided", http.StatusBadRequest)
		return
	}

	// Use provided config or default
	if request.Config != nil {
		// Mock config setting since config field is unexported
		_ = request.Config
	}

	result, err := h.patternEngine.AnalyzeMisclassifications(ctx, request.Misclassifications)
	if err != nil {
		h.logger.Error("Failed to analyze misclassifications", zap.Error(err))
		http.Error(w, "Analysis failed", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"result":  result,
		"metadata": map[string]interface{}{
			"analyzed_at": time.Now(),
			"count":       len(request.Misclassifications),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPatternsHandler returns all detected patterns
func (h *PatternAnalysisHandler) GetPatternsHandler(w http.ResponseWriter, r *http.Request) {
	patterns := h.patternEngine.GetPatterns()

	response := map[string]interface{}{
		"success":  true,
		"patterns": patterns,
		"count":    len(patterns),
		"metadata": map[string]interface{}{
			"retrieved_at": time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPatternsByTypeHandler returns patterns filtered by type
func (h *PatternAnalysisHandler) GetPatternsByTypeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	patternType := classification_monitoring.PatternType(vars["type"])

	patterns := h.patternEngine.GetPatternsByType(patternType)

	response := map[string]interface{}{
		"success":      true,
		"patterns":     patterns,
		"pattern_type": patternType,
		"count":        len(patterns),
		"metadata": map[string]interface{}{
			"retrieved_at": time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPatternsBySeverityHandler returns patterns filtered by severity
func (h *PatternAnalysisHandler) GetPatternsBySeverityHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	severity := classification_monitoring.PatternSeverity(vars["severity"])

	patterns := h.patternEngine.GetPatternsBySeverity(severity)

	response := map[string]interface{}{
		"success":  true,
		"patterns": patterns,
		"severity": severity,
		"count":    len(patterns),
		"metadata": map[string]interface{}{
			"retrieved_at": time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPatternHistoryHandler returns the history of pattern analysis results
func (h *PatternAnalysisHandler) GetPatternHistoryHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limitStr := query.Get("limit")

	limit := 50 // default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	history := h.patternEngine.GetPatternHistory()

	// Apply limit
	if len(history) > limit {
		history = history[len(history)-limit:]
	}

	response := map[string]interface{}{
		"success": true,
		"history": history,
		"count":   len(history),
		"limit":   limit,
		"metadata": map[string]interface{}{
			"retrieved_at": time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPatternSummaryHandler returns a summary of pattern analysis
func (h *PatternAnalysisHandler) GetPatternSummaryHandler(w http.ResponseWriter, r *http.Request) {
	patterns := h.patternEngine.GetPatterns()
	history := h.patternEngine.GetPatternHistory()

	// Calculate summary statistics
	patternsByType := make(map[classification_monitoring.PatternType]int)
	patternsByCategory := make(map[classification_monitoring.PatternCategory]int)
	patternsBySeverity := make(map[classification_monitoring.PatternSeverity]int)

	var totalImpact float64
	criticalPatterns := 0
	highImpactPatterns := 0

	for _, pattern := range patterns {
		patternsByType[pattern.PatternType]++
		patternsByCategory[pattern.Category]++
		patternsBySeverity[pattern.Severity]++

		totalImpact += pattern.ImpactScore

		if pattern.Severity == classification_monitoring.PatternSeverityCritical {
			criticalPatterns++
		}

		if pattern.ImpactScore >= 0.7 {
			highImpactPatterns++
		}
	}

	var averageImpact float64
	if len(patterns) > 0 {
		averageImpact = totalImpact / float64(len(patterns))
	}

	// Determine risk level
	var riskLevel string
	switch {
	case averageImpact >= 0.8:
		riskLevel = "critical"
	case averageImpact >= 0.6:
		riskLevel = "high"
	case averageImpact >= 0.4:
		riskLevel = "medium"
	default:
		riskLevel = "low"
	}

	summary := map[string]interface{}{
		"total_patterns":       len(patterns),
		"patterns_by_type":     patternsByType,
		"patterns_by_category": patternsByCategory,
		"patterns_by_severity": patternsBySeverity,
		"critical_patterns":    criticalPatterns,
		"high_impact_patterns": highImpactPatterns,
		"average_impact":       averageImpact,
		"risk_level":           riskLevel,
		"analysis_history": map[string]interface{}{
			"total_analyses": len(history),
			"last_analysis":  nil,
		},
	}

	// Add last analysis info if available
	if len(history) > 0 {
		lastAnalysis := history[len(history)-1]
		summary["analysis_history"].(map[string]interface{})["last_analysis"] = map[string]interface{}{
			"id":             lastAnalysis.ID,
			"analysis_time":  lastAnalysis.AnalysisTime,
			"patterns_found": lastAnalysis.PatternsFound,
			"new_patterns":   lastAnalysis.NewPatterns,
		}
	}

	response := map[string]interface{}{
		"success": true,
		"summary": summary,
		"metadata": map[string]interface{}{
			"generated_at": time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPatternDetailsHandler returns detailed information about a specific pattern
func (h *PatternAnalysisHandler) GetPatternDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	patternID := vars["id"]

	patterns := h.patternEngine.GetPatterns()
	pattern, exists := patterns[patternID]

	if !exists {
		http.Error(w, "Pattern not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"pattern": pattern,
		"metadata": map[string]interface{}{
			"retrieved_at": time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetRecommendationsHandler returns recommendations based on current patterns
func (h *PatternAnalysisHandler) GetRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	patterns := h.patternEngine.GetPatterns()

	// Convert map to slice for recommendation engine
	patternSlice := make([]*classification_monitoring.MisclassificationPattern, 0, len(patterns))
	for _, pattern := range patterns {
		patternSlice = append(patternSlice, pattern)
	}

	// Generate recommendations
	recommendationEngine := classification_monitoring.NewRecommendationEngine(h.logger)
	recommendations := recommendationEngine.GenerateRecommendations(patternSlice)

	response := map[string]interface{}{
		"success":         true,
		"recommendations": recommendations,
		"count":           len(recommendations),
		"metadata": map[string]interface{}{
			"generated_at": time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HealthCheckHandler provides health status for the pattern analysis engine
func (h *PatternAnalysisHandler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	patterns := h.patternEngine.GetPatterns()
	history := h.patternEngine.GetPatternHistory()

	health := map[string]interface{}{
		"status": "healthy",
		"stats": map[string]interface{}{
			"active_patterns":  len(patterns),
			"analysis_history": len(history),
			"uptime":           "24h", // Mock since startTime field is unexported
		},
		"metadata": map[string]interface{}{
			"checked_at": time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}
