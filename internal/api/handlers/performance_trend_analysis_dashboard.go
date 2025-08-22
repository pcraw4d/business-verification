package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/company/kyb-platform/internal/observability"
)

// PerformanceTrendAnalysisDashboardHandler provides HTTP handlers for trend analysis
type PerformanceTrendAnalysisDashboardHandler struct {
	trendAnalysisSystem *observability.PerformanceTrendAnalysisSystem
	logger              *zap.Logger
}

// NewPerformanceTrendAnalysisDashboardHandler creates a new trend analysis dashboard handler
func NewPerformanceTrendAnalysisDashboardHandler(
	trendAnalysisSystem *observability.PerformanceTrendAnalysisSystem,
	logger *zap.Logger,
) *PerformanceTrendAnalysisDashboardHandler {
	return &PerformanceTrendAnalysisDashboardHandler{
		trendAnalysisSystem: trendAnalysisSystem,
		logger:              logger,
	}
}

// AnalyzeTrendRequest represents a trend analysis request
type AnalyzeTrendRequest struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// GenerateReportRequest represents a report generation request
type GenerateReportRequest struct {
	ReportType string    `json:"report_type"` // trend, performance, forecast
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Format     string    `json:"format"` // json, csv, pdf
}

// GenerateForecastRequest represents a forecast generation request
type GenerateForecastRequest struct {
	Horizon time.Duration `json:"horizon"`
}

// UpdateConfigRequest represents a configuration update request
type UpdateConfigRequest struct {
	Config *observability.TrendAnalysisConfig `json:"config"`
}

// AnalyzeTrend performs trend analysis on historical data
func (h *PerformanceTrendAnalysisDashboardHandler) AnalyzeTrend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AnalyzeTrendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	timeRange := observability.TimeRange{
		Start: req.StartTime,
		End:   req.EndTime,
	}

	result, err := h.trendAnalysisSystem.AnalyzeTrend(r.Context(), timeRange)
	if err != nil {
		h.logger.Error("failed to analyze trend", zap.Error(err))
		http.Error(w, "Failed to analyze trend", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GenerateTrendReport generates a comprehensive trend report
func (h *PerformanceTrendAnalysisDashboardHandler) GenerateTrendReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GenerateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	timeRange := observability.TimeRange{
		Start: req.StartTime,
		End:   req.EndTime,
	}

	report, err := h.trendAnalysisSystem.GenerateTrendReport(r.Context(), timeRange)
	if err != nil {
		h.logger.Error("failed to generate trend report", zap.Error(err))
		http.Error(w, "Failed to generate trend report", http.StatusInternalServerError)
		return
	}

	// Export if format is specified
	if req.Format != "" && req.Format != "json" {
		exported, err := h.trendAnalysisSystem.ExportReport(r.Context(), report, req.Format)
		if err != nil {
			h.logger.Error("failed to export trend report", zap.Error(err))
			http.Error(w, "Failed to export trend report", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", getContentType(req.Format))
		w.Write(exported)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GeneratePerformanceReport generates a performance report
func (h *PerformanceTrendAnalysisDashboardHandler) GeneratePerformanceReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GenerateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	timeRange := observability.TimeRange{
		Start: req.StartTime,
		End:   req.EndTime,
	}

	report, err := h.trendAnalysisSystem.GeneratePerformanceReport(r.Context(), timeRange)
	if err != nil {
		h.logger.Error("failed to generate performance report", zap.Error(err))
		http.Error(w, "Failed to generate performance report", http.StatusInternalServerError)
		return
	}

	// Export if format is specified
	if req.Format != "" && req.Format != "json" {
		exported, err := h.trendAnalysisSystem.ExportReport(r.Context(), report, req.Format)
		if err != nil {
			h.logger.Error("failed to export performance report", zap.Error(err))
			http.Error(w, "Failed to export performance report", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", getContentType(req.Format))
		w.Write(exported)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GenerateForecast generates a performance forecast
func (h *PerformanceTrendAnalysisDashboardHandler) GenerateForecast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GenerateForecastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	forecast, err := h.trendAnalysisSystem.GenerateForecast(r.Context(), req.Horizon)
	if err != nil {
		h.logger.Error("failed to generate forecast", zap.Error(err))
		http.Error(w, "Failed to generate forecast", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(forecast)
}

// GetTrendHistory retrieves trend analysis history
func (h *PerformanceTrendAnalysisDashboardHandler) GetTrendHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		http.Error(w, "Invalid start_time parameter", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		http.Error(w, "Invalid end_time parameter", http.StatusBadRequest)
		return
	}

	timeRange := observability.TimeRange{
		Start: startTime,
		End:   endTime,
	}

	history, err := h.trendAnalysisSystem.GetTrendHistory(r.Context(), timeRange)
	if err != nil {
		h.logger.Error("failed to get trend history", zap.Error(err))
		http.Error(w, "Failed to get trend history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// GetReports retrieves generated reports
func (h *PerformanceTrendAnalysisDashboardHandler) GetReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	reportType := r.URL.Query().Get("report_type")
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	if reportType == "" {
		http.Error(w, "report_type parameter is required", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		http.Error(w, "Invalid start_time parameter", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		http.Error(w, "Invalid end_time parameter", http.StatusBadRequest)
		return
	}

	timeRange := observability.TimeRange{
		Start: startTime,
		End:   endTime,
	}

	reports, err := h.trendAnalysisSystem.GetReports(r.Context(), reportType, timeRange)
	if err != nil {
		h.logger.Error("failed to get reports", zap.Error(err))
		http.Error(w, "Failed to get reports", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}

// ExportReport exports a report in the specified format
func (h *PerformanceTrendAnalysisDashboardHandler) ExportReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Report interface{} `json:"report"`
		Format string      `json:"format"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Format == "" {
		http.Error(w, "format parameter is required", http.StatusBadRequest)
		return
	}

	exported, err := h.trendAnalysisSystem.ExportReport(r.Context(), req.Report, req.Format)
	if err != nil {
		h.logger.Error("failed to export report", zap.Error(err))
		http.Error(w, "Failed to export report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", getContentType(req.Format))
	w.Write(exported)
}

// UpdateConfiguration updates the system configuration
func (h *PerformanceTrendAnalysisDashboardHandler) UpdateConfiguration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Config == nil {
		http.Error(w, "config is required", http.StatusBadRequest)
		return
	}

	if err := h.trendAnalysisSystem.UpdateConfiguration(req.Config); err != nil {
		h.logger.Error("failed to update configuration", zap.Error(err))
		http.Error(w, "Failed to update configuration", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "configuration updated"})
}

// GetConfiguration returns the current configuration
func (h *PerformanceTrendAnalysisDashboardHandler) GetConfiguration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config := h.trendAnalysisSystem.GetConfiguration()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// GetStatus returns the system status
func (h *PerformanceTrendAnalysisDashboardHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := h.trendAnalysisSystem.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// GetTrendMetrics returns trend metrics for dashboard
func (h *PerformanceTrendAnalysisDashboardHandler) GetTrendMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	metricType := r.URL.Query().Get("metric_type")
	timeRangeStr := r.URL.Query().Get("time_range") // e.g., "1h", "24h", "7d"

	if metricType == "" {
		http.Error(w, "metric_type parameter is required", http.StatusBadRequest)
		return
	}

	// Calculate time range based on parameter
	var timeRange observability.TimeRange
	endTime := time.Now()

	switch timeRangeStr {
	case "1h":
		timeRange = observability.TimeRange{
			Start: endTime.Add(-1 * time.Hour),
			End:   endTime,
		}
	case "24h":
		timeRange = observability.TimeRange{
			Start: endTime.Add(-24 * time.Hour),
			End:   endTime,
		}
	case "7d":
		timeRange = observability.TimeRange{
			Start: endTime.Add(-7 * 24 * time.Hour),
			End:   endTime,
		}
	default:
		timeRange = observability.TimeRange{
			Start: endTime.Add(-24 * time.Hour),
			End:   endTime,
		}
	}

	// Get trend analysis for the specified metric type
	result, err := h.trendAnalysisSystem.AnalyzeTrend(r.Context(), timeRange)
	if err != nil {
		h.logger.Error("failed to get trend metrics", zap.Error(err))
		http.Error(w, "Failed to get trend metrics", http.StatusInternalServerError)
		return
	}

	// Filter metrics based on type
	var metrics map[string]interface{}
	switch metricType {
	case "overall":
		metrics = map[string]interface{}{
			"overall_trend":     result.OverallTrend,
			"trend_strength":    result.TrendStrength,
			"trend_confidence":  result.TrendConfidence,
			"data_points_count": result.DataPointsCount,
		}
	case "response_time":
		if trend, exists := result.MetricTrends["response_time"]; exists {
			metrics = map[string]interface{}{
				"direction":   trend.Direction,
				"strength":    trend.Strength,
				"confidence":  trend.Confidence,
				"change_rate": trend.ChangeRate,
			}
		}
	case "throughput":
		if trend, exists := result.MetricTrends["throughput"]; exists {
			metrics = map[string]interface{}{
				"direction":   trend.Direction,
				"strength":    trend.Strength,
				"confidence":  trend.Confidence,
				"change_rate": trend.ChangeRate,
			}
		}
	case "error_rate":
		if trend, exists := result.MetricTrends["error_rate"]; exists {
			metrics = map[string]interface{}{
				"direction":   trend.Direction,
				"strength":    trend.Strength,
				"confidence":  trend.Confidence,
				"change_rate": trend.ChangeRate,
			}
		}
	default:
		metrics = result.MetricTrends
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetTrendAlerts returns trend-based alerts
func (h *PerformanceTrendAnalysisDashboardHandler) GetTrendAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	severity := r.URL.Query().Get("severity")
	limitStr := r.URL.Query().Get("limit")

	limit := 50 // default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Get recent trend analysis
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	timeRange := observability.TimeRange{
		Start: startTime,
		End:   endTime,
	}

	result, err := h.trendAnalysisSystem.AnalyzeTrend(r.Context(), timeRange)
	if err != nil {
		h.logger.Error("failed to get trend alerts", zap.Error(err))
		http.Error(w, "Failed to get trend alerts", http.StatusInternalServerError)
		return
	}

	// Generate alerts based on trend analysis
	var alerts []map[string]interface{}

	// Overall trend alerts
	if result.TrendConfidence > 0.8 {
		if result.OverallTrend == observability.TrendDirectionDecreasing {
			alerts = append(alerts, map[string]interface{}{
				"id":          "trend_decreasing",
				"type":        "trend",
				"severity":    "warning",
				"title":       "Performance Trend Decreasing",
				"description": "Overall performance trend is decreasing",
				"confidence":  result.TrendConfidence,
				"timestamp":   result.AnalysisTime,
			})
		}
	}

	// Metric-specific alerts
	for metricName, trend := range result.MetricTrends {
		if trend.Confidence > 0.8 {
			if trend.Direction == observability.TrendDirectionDecreasing && trend.Strength > 0.5 {
				alerts = append(alerts, map[string]interface{}{
					"id":          "metric_decreasing_" + metricName,
					"type":        "metric_trend",
					"severity":    "warning",
					"title":       metricName + " Trend Decreasing",
					"description": metricName + " is showing a decreasing trend",
					"confidence":  trend.Confidence,
					"metric":      metricName,
					"timestamp":   result.AnalysisTime,
				})
			}
		}
	}

	// Anomaly alerts
	for _, anomaly := range result.Anomalies {
		alerts = append(alerts, map[string]interface{}{
			"id":           "anomaly_" + anomaly.ID,
			"type":         "anomaly",
			"severity":     "critical",
			"title":        "Performance Anomaly Detected",
			"description":  "Anomaly detected in performance metrics",
			"confidence":   anomaly.Confidence,
			"anomaly_type": anomaly.Type,
			"timestamp":    anomaly.DetectedAt,
		})
	}

	// Filter by severity if specified
	if severity != "" {
		var filteredAlerts []map[string]interface{}
		for _, alert := range alerts {
			if alert["severity"] == severity {
				filteredAlerts = append(filteredAlerts, alert)
			}
		}
		alerts = filteredAlerts
	}

	// Limit results
	if len(alerts) > limit {
		alerts = alerts[:limit]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

// GetTrendRecommendations returns trend-based recommendations
func (h *PerformanceTrendAnalysisDashboardHandler) GetTrendRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get recent trend analysis
	endTime := time.Now()
	startTime := endTime.Add(-7 * 24 * time.Hour) // Last 7 days

	timeRange := observability.TimeRange{
		Start: startTime,
		End:   endTime,
	}

	result, err := h.trendAnalysisSystem.AnalyzeTrend(r.Context(), timeRange)
	if err != nil {
		h.logger.Error("failed to get trend recommendations", zap.Error(err))
		http.Error(w, "Failed to get trend recommendations", http.StatusInternalServerError)
		return
	}

	var recommendations []map[string]interface{}

	// Generate recommendations based on trend analysis
	if result.OverallTrend == observability.TrendDirectionDecreasing {
		recommendations = append(recommendations, map[string]interface{}{
			"id":          "improve_overall_performance",
			"priority":    "high",
			"title":       "Improve Overall Performance",
			"description": "Overall performance is trending downward. Consider investigating root causes.",
			"actions":     []string{"Review recent deployments", "Check resource utilization", "Analyze error patterns"},
			"impact":      "high",
			"effort":      "medium",
		})
	}

	// Metric-specific recommendations
	for metricName, trend := range result.MetricTrends {
		if trend.Direction == observability.TrendDirectionDecreasing && trend.Strength > 0.6 {
			recommendations = append(recommendations, map[string]interface{}{
				"id":          "optimize_" + metricName,
				"priority":    "medium",
				"title":       "Optimize " + metricName,
				"description": metricName + " is showing a strong decreasing trend",
				"actions":     []string{"Investigate " + metricName + " bottlenecks", "Review related configurations"},
				"impact":      "medium",
				"effort":      "low",
			})
		}
	}

	// Correlation-based recommendations
	if result.Correlations != nil {
		for _, correlation := range result.Correlations.Significant {
			if correlation.Coefficient > 0.8 {
				recommendations = append(recommendations, map[string]interface{}{
					"id":          "correlation_" + correlation.Metric1 + "_" + correlation.Metric2,
					"priority":    "low",
					"title":       "Investigate Correlation",
					"description": "Strong correlation detected between " + correlation.Metric1 + " and " + correlation.Metric2,
					"actions":     []string{"Analyze relationship between metrics", "Consider optimization opportunities"},
					"impact":      "low",
					"effort":      "low",
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recommendations)
}

// getContentType returns the appropriate content type for export formats
func getContentType(format string) string {
	switch format {
	case "csv":
		return "text/csv"
	case "pdf":
		return "application/pdf"
	case "excel":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/json"
	}
}
