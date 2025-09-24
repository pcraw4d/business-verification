package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/internal/observability"
)

// AdvancedMonitoringHandler handles advanced monitoring dashboard requests
type AdvancedMonitoringHandler struct {
	dashboard *observability.AdvancedMonitoringDashboard
	logger    *zap.Logger
}

// NewAdvancedMonitoringHandler creates a new advanced monitoring handler
func NewAdvancedMonitoringHandler(
	dashboard *observability.AdvancedMonitoringDashboard,
	logger *zap.Logger,
) *AdvancedMonitoringHandler {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &AdvancedMonitoringHandler{
		dashboard: dashboard,
		logger:    logger,
	}
}

// GetDashboardData returns comprehensive dashboard data
func (h *AdvancedMonitoringHandler) GetDashboardData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get dashboard data
	data, err := h.dashboard.GetDashboardData(ctx)
	if err != nil {
		h.logger.Error("failed to get dashboard data", zap.Error(err))
		http.Error(w, "Failed to get dashboard data", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode dashboard data", zap.Error(err))
		http.Error(w, "Failed to encode dashboard data", http.StatusInternalServerError)
		return
	}

	h.logger.Info("dashboard data served successfully")
}

// GetMLModelMetrics returns ML model performance metrics
func (h *AdvancedMonitoringHandler) GetMLModelMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get dashboard data
	data, err := h.dashboard.GetDashboardData(ctx)
	if err != nil {
		h.logger.Error("failed to get ML model metrics", zap.Error(err))
		http.Error(w, "Failed to get ML model metrics", http.StatusInternalServerError)
		return
	}

	// Extract ML model metrics
	response := map[string]interface{}{
		"timestamp":        time.Now(),
		"ml_model_metrics": data.MLModelMetrics,
		"ml_model_health":  data.MLModelHealth,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode ML model metrics", zap.Error(err))
		http.Error(w, "Failed to encode ML model metrics", http.StatusInternalServerError)
		return
	}

	h.logger.Info("ML model metrics served successfully")
}

// GetEnsembleMetrics returns ensemble method contributions
func (h *AdvancedMonitoringHandler) GetEnsembleMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get dashboard data
	data, err := h.dashboard.GetDashboardData(ctx)
	if err != nil {
		h.logger.Error("failed to get ensemble metrics", zap.Error(err))
		http.Error(w, "Failed to get ensemble metrics", http.StatusInternalServerError)
		return
	}

	// Extract ensemble metrics
	response := map[string]interface{}{
		"timestamp":        time.Now(),
		"ensemble_metrics": data.EnsembleMetrics,
		"ensemble_health":  data.EnsembleHealth,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode ensemble metrics", zap.Error(err))
		http.Error(w, "Failed to encode ensemble metrics", http.StatusInternalServerError)
		return
	}

	h.logger.Info("ensemble metrics served successfully")
}

// GetUncertaintyMetrics returns uncertainty quantification metrics
func (h *AdvancedMonitoringHandler) GetUncertaintyMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get dashboard data
	data, err := h.dashboard.GetDashboardData(ctx)
	if err != nil {
		h.logger.Error("failed to get uncertainty metrics", zap.Error(err))
		http.Error(w, "Failed to get uncertainty metrics", http.StatusInternalServerError)
		return
	}

	// Extract uncertainty metrics
	response := map[string]interface{}{
		"timestamp":           time.Now(),
		"uncertainty_metrics": data.UncertaintyMetrics,
		"uncertainty_health":  data.UncertaintyHealth,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode uncertainty metrics", zap.Error(err))
		http.Error(w, "Failed to encode uncertainty metrics", http.StatusInternalServerError)
		return
	}

	h.logger.Info("uncertainty metrics served successfully")
}

// GetSecurityMetrics returns security compliance metrics
func (h *AdvancedMonitoringHandler) GetSecurityMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get dashboard data
	data, err := h.dashboard.GetDashboardData(ctx)
	if err != nil {
		h.logger.Error("failed to get security metrics", zap.Error(err))
		http.Error(w, "Failed to get security metrics", http.StatusInternalServerError)
		return
	}

	// Extract security metrics
	response := map[string]interface{}{
		"timestamp":        time.Now(),
		"security_metrics": data.SecurityMetrics,
		"security_health":  data.SecurityHealth,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode security metrics", zap.Error(err))
		http.Error(w, "Failed to encode security metrics", http.StatusInternalServerError)
		return
	}

	h.logger.Info("security metrics served successfully")
}

// GetPerformanceMetrics returns performance metrics
func (h *AdvancedMonitoringHandler) GetPerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get dashboard data
	data, err := h.dashboard.GetDashboardData(ctx)
	if err != nil {
		h.logger.Error("failed to get performance metrics", zap.Error(err))
		http.Error(w, "Failed to get performance metrics", http.StatusInternalServerError)
		return
	}

	// Extract performance metrics
	response := map[string]interface{}{
		"timestamp":           time.Now(),
		"performance_metrics": data.PerformanceMetrics,
		"performance_health":  data.PerformanceHealth,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode performance metrics", zap.Error(err))
		http.Error(w, "Failed to encode performance metrics", http.StatusInternalServerError)
		return
	}

	h.logger.Info("performance metrics served successfully")
}

// GetAlertsSummary returns alerts summary
func (h *AdvancedMonitoringHandler) GetAlertsSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get dashboard data
	data, err := h.dashboard.GetDashboardData(ctx)
	if err != nil {
		h.logger.Error("failed to get alerts summary", zap.Error(err))
		http.Error(w, "Failed to get alerts summary", http.StatusInternalServerError)
		return
	}

	// Extract alerts summary
	response := map[string]interface{}{
		"timestamp":      time.Now(),
		"alerts_summary": data.AlertsSummary,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode alerts summary", zap.Error(err))
		http.Error(w, "Failed to encode alerts summary", http.StatusInternalServerError)
		return
	}

	h.logger.Info("alerts summary served successfully")
}

// GetHealthStatus returns overall health status
func (h *AdvancedMonitoringHandler) GetHealthStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get dashboard data
	data, err := h.dashboard.GetDashboardData(ctx)
	if err != nil {
		h.logger.Error("failed to get health status", zap.Error(err))
		http.Error(w, "Failed to get health status", http.StatusInternalServerError)
		return
	}

	// Extract health status
	response := map[string]interface{}{
		"timestamp":      time.Now(),
		"overall_health": data.OverallHealth,
		"health_score":   data.HealthScore,
		"last_updated":   h.dashboard.GetLastUpdateTime(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode health status", zap.Error(err))
		http.Error(w, "Failed to encode health status", http.StatusInternalServerError)
		return
	}

	h.logger.Info("health status served successfully")
}

// GetRecommendations returns recommendations
func (h *AdvancedMonitoringHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get dashboard data
	data, err := h.dashboard.GetDashboardData(ctx)
	if err != nil {
		h.logger.Error("failed to get recommendations", zap.Error(err))
		http.Error(w, "Failed to get recommendations", http.StatusInternalServerError)
		return
	}

	// Extract recommendations
	response := map[string]interface{}{
		"timestamp":       time.Now(),
		"recommendations": data.Recommendations,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode recommendations", zap.Error(err))
		http.Error(w, "Failed to encode recommendations", http.StatusInternalServerError)
		return
	}

	h.logger.Info("recommendations served successfully")
}

// ExportDashboardData exports dashboard data in various formats
func (h *AdvancedMonitoringHandler) ExportDashboardData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get format from query parameter
	vars := mux.Vars(r)
	format := vars["format"]
	if format == "" {
		format = "json"
	}

	// Export dashboard data
	data, err := h.dashboard.ExportDashboardData(ctx, format)
	if err != nil {
		h.logger.Error("failed to export dashboard data", zap.Error(err))
		http.Error(w, "Failed to export dashboard data", http.StatusInternalServerError)
		return
	}

	// Set response headers based on format
	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=dashboard_data.json")
	case "yaml":
		w.Header().Set("Content-Type", "application/x-yaml")
		w.Header().Set("Content-Disposition", "attachment; filename=dashboard_data.yaml")
	default:
		http.Error(w, "Unsupported format", http.StatusBadRequest)
		return
	}

	// Send data
	if _, err := w.Write(data); err != nil {
		h.logger.Error("failed to write exported data", zap.Error(err))
		http.Error(w, "Failed to write exported data", http.StatusInternalServerError)
		return
	}

	h.logger.Info("dashboard data exported successfully", zap.String("format", format))
}

// GetModelDriftVisualization returns model drift visualization data
func (h *AdvancedMonitoringHandler) GetModelDriftVisualization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get dashboard data
	data, err := h.dashboard.GetDashboardData(ctx)
	if err != nil {
		h.logger.Error("failed to get model drift visualization", zap.Error(err))
		http.Error(w, "Failed to get model drift visualization", http.StatusInternalServerError)
		return
	}

	// Create drift visualization data
	driftData := make(map[string]interface{})
	for modelID, metrics := range data.MLModelMetrics {
		driftData[modelID] = map[string]interface{}{
			"model_name":   metrics.ModelName,
			"drift_score":  metrics.DriftScore,
			"drift_status": metrics.DriftStatus,
			"last_updated": metrics.LastUpdated,
		}
	}

	response := map[string]interface{}{
		"timestamp":  time.Now(),
		"drift_data": driftData,
		"drift_summary": map[string]interface{}{
			"total_models":   len(data.MLModelMetrics),
			"critical_drift": h.countDriftStatus(data.MLModelMetrics, "critical"),
			"warning_drift":  h.countDriftStatus(data.MLModelMetrics, "warning"),
			"healthy_drift":  h.countDriftStatus(data.MLModelMetrics, "healthy"),
		},
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode model drift visualization", zap.Error(err))
		http.Error(w, "Failed to encode model drift visualization", http.StatusInternalServerError)
		return
	}

	h.logger.Info("model drift visualization served successfully")
}

// GetEnsembleContributionVisualization returns ensemble contribution visualization data
func (h *AdvancedMonitoringHandler) GetEnsembleContributionVisualization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get dashboard data
	data, err := h.dashboard.GetDashboardData(ctx)
	if err != nil {
		h.logger.Error("failed to get ensemble contribution visualization", zap.Error(err))
		http.Error(w, "Failed to get ensemble contribution visualization", http.StatusInternalServerError)
		return
	}

	// Create contribution visualization data
	contributionData := make([]map[string]interface{}, 0)
	totalContribution := 0.0

	for methodID, metrics := range data.EnsembleMetrics {
		contributionData = append(contributionData, map[string]interface{}{
			"method_id":    methodID,
			"method_name":  metrics.MethodName,
			"method_type":  metrics.MethodType,
			"weight":       metrics.Weight,
			"contribution": metrics.Contribution,
			"accuracy":     metrics.Accuracy,
			"confidence":   metrics.Confidence,
			"usage_count":  metrics.UsageCount,
			"success_rate": metrics.SuccessRate,
		})
		totalContribution += metrics.Contribution
	}

	response := map[string]interface{}{
		"timestamp":         time.Now(),
		"contribution_data": contributionData,
		"contribution_summary": map[string]interface{}{
			"total_methods":      len(data.EnsembleMetrics),
			"total_contribution": totalContribution,
			"average_weight":     h.calculateAverageWeight(data.EnsembleMetrics),
			"top_contributor":    h.findTopContributor(data.EnsembleMetrics),
		},
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode ensemble contribution visualization", zap.Error(err))
		http.Error(w, "Failed to encode ensemble contribution visualization", http.StatusInternalServerError)
		return
	}

	h.logger.Info("ensemble contribution visualization served successfully")
}

// GetUncertaintyVisualization returns uncertainty quantification visualization data
func (h *AdvancedMonitoringHandler) GetUncertaintyVisualization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get dashboard data
	data, err := h.dashboard.GetDashboardData(ctx)
	if err != nil {
		h.logger.Error("failed to get uncertainty visualization", zap.Error(err))
		http.Error(w, "Failed to get uncertainty visualization", http.StatusInternalServerError)
		return
	}

	// Create uncertainty visualization data
	uncertaintyData := map[string]interface{}{
		"overall_uncertainty":   data.UncertaintyMetrics.OverallUncertainty,
		"calibration_score":     data.UncertaintyMetrics.CalibrationScore,
		"reliability_score":     data.UncertaintyMetrics.ReliabilityScore,
		"confidence_interval":   data.UncertaintyMetrics.ConfidenceInterval,
		"prediction_variance":   data.UncertaintyMetrics.PredictionVariance,
		"entropy":               data.UncertaintyMetrics.Entropy,
		"epistemic_uncertainty": data.UncertaintyMetrics.EpistemicUncertainty,
		"aleatoric_uncertainty": data.UncertaintyMetrics.AleatoricUncertainty,
		"health_status":         data.UncertaintyMetrics.HealthStatus,
		"last_updated":          data.UncertaintyMetrics.LastUpdated,
	}

	response := map[string]interface{}{
		"timestamp":        time.Now(),
		"uncertainty_data": uncertaintyData,
		"uncertainty_summary": map[string]interface{}{
			"health_status":       data.UncertaintyHealth,
			"overall_uncertainty": data.UncertaintyMetrics.OverallUncertainty,
			"calibration_quality": h.assessCalibrationQuality(data.UncertaintyMetrics.CalibrationScore),
			"reliability_quality": h.assessReliabilityQuality(data.UncertaintyMetrics.ReliabilityScore),
		},
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode uncertainty visualization", zap.Error(err))
		http.Error(w, "Failed to encode uncertainty visualization", http.StatusInternalServerError)
		return
	}

	h.logger.Info("uncertainty visualization served successfully")
}

// GetSecurityComplianceVisualization returns security compliance visualization data
func (h *AdvancedMonitoringHandler) GetSecurityComplianceVisualization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get dashboard data
	data, err := h.dashboard.GetDashboardData(ctx)
	if err != nil {
		h.logger.Error("failed to get security compliance visualization", zap.Error(err))
		http.Error(w, "Failed to get security compliance visualization", http.StatusInternalServerError)
		return
	}

	// Create security compliance visualization data
	securityData := map[string]interface{}{
		"overall_compliance":        data.SecurityMetrics.OverallCompliance,
		"data_source_trust_rate":    data.SecurityMetrics.DataSourceTrustRate,
		"website_verification_rate": data.SecurityMetrics.WebsiteVerificationRate,
		"security_violation_rate":   data.SecurityMetrics.SecurityViolationRate,
		"confidence_integrity":      data.SecurityMetrics.ConfidenceIntegrity,
		"processing_time":           data.SecurityMetrics.ProcessingTime,
		"error_rate":                data.SecurityMetrics.ErrorRate,
		"health_status":             data.SecurityMetrics.HealthStatus,
		"last_updated":              data.SecurityMetrics.LastUpdated,
	}

	response := map[string]interface{}{
		"timestamp":     time.Now(),
		"security_data": securityData,
		"security_summary": map[string]interface{}{
			"health_status":          data.SecurityHealth,
			"overall_compliance":     data.SecurityMetrics.OverallCompliance,
			"trust_rate_quality":     h.assessTrustRateQuality(data.SecurityMetrics.DataSourceTrustRate),
			"verification_quality":   h.assessVerificationQuality(data.SecurityMetrics.WebsiteVerificationRate),
			"violation_rate_quality": h.assessViolationRateQuality(data.SecurityMetrics.SecurityViolationRate),
		},
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode security compliance visualization", zap.Error(err))
		http.Error(w, "Failed to encode security compliance visualization", http.StatusInternalServerError)
		return
	}

	h.logger.Info("security compliance visualization served successfully")
}

// Helper methods
func (h *AdvancedMonitoringHandler) countDriftStatus(metrics map[string]*observability.MLModelMetrics, status string) int {
	count := 0
	for _, m := range metrics {
		if m.DriftStatus == status {
			count++
		}
	}
	return count
}

func (h *AdvancedMonitoringHandler) calculateAverageWeight(metrics map[string]*observability.EnsembleMethodMetrics) float64 {
	if len(metrics) == 0 {
		return 0.0
	}

	total := 0.0
	for _, m := range metrics {
		total += m.Weight
	}
	return total / float64(len(metrics))
}

func (h *AdvancedMonitoringHandler) findTopContributor(metrics map[string]*observability.EnsembleMethodMetrics) string {
	if len(metrics) == 0 {
		return ""
	}

	topMethod := ""
	topContribution := 0.0

	for methodID, m := range metrics {
		if m.Contribution > topContribution {
			topContribution = m.Contribution
			topMethod = methodID
		}
	}

	return topMethod
}

func (h *AdvancedMonitoringHandler) assessCalibrationQuality(score float64) string {
	if score >= 0.9 {
		return "excellent"
	} else if score >= 0.8 {
		return "good"
	} else if score >= 0.7 {
		return "fair"
	} else {
		return "poor"
	}
}

func (h *AdvancedMonitoringHandler) assessReliabilityQuality(score float64) string {
	if score >= 0.9 {
		return "excellent"
	} else if score >= 0.8 {
		return "good"
	} else if score >= 0.7 {
		return "fair"
	} else {
		return "poor"
	}
}

func (h *AdvancedMonitoringHandler) assessTrustRateQuality(rate float64) string {
	if rate >= 0.95 {
		return "excellent"
	} else if rate >= 0.9 {
		return "good"
	} else if rate >= 0.8 {
		return "fair"
	} else {
		return "poor"
	}
}

func (h *AdvancedMonitoringHandler) assessVerificationQuality(rate float64) string {
	if rate >= 0.9 {
		return "excellent"
	} else if rate >= 0.8 {
		return "good"
	} else if rate >= 0.7 {
		return "fair"
	} else {
		return "poor"
	}
}

func (h *AdvancedMonitoringHandler) assessViolationRateQuality(rate float64) string {
	if rate <= 0.01 {
		return "excellent"
	} else if rate <= 0.05 {
		return "good"
	} else if rate <= 0.1 {
		return "fair"
	} else {
		return "poor"
	}
}
