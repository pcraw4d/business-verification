package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/external/news_api"
)

// AdverseMediaHandler handles adverse media monitoring requests
type AdverseMediaHandler struct {
	monitor *news_api.AdverseMediaMonitor
	logger  *zap.Logger
}

// NewAdverseMediaHandler creates a new adverse media handler
func NewAdverseMediaHandler(monitor *news_api.AdverseMediaMonitor, logger *zap.Logger) *AdverseMediaHandler {
	return &AdverseMediaHandler{
		monitor: monitor,
		logger:  logger,
	}
}

// StartMonitoringRequest represents a request to start monitoring
type StartMonitoringRequest struct {
	EntityNames  []string `json:"entity_names"`
	ScanInterval string   `json:"scan_interval,omitempty"` // e.g., "1h", "6h", "24h"
}

// StartMonitoringResponse represents the response to start monitoring
type StartMonitoringResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	MonitoringID string `json:"monitoring_id,omitempty"`
}

// ScanRequest represents a request to perform a scan
type ScanRequest struct {
	EntityNames []string `json:"entity_names"`
}

// ScanResponse represents the response to a scan request
type ScanResponse struct {
	Success bool                             `json:"success"`
	Data    *news_api.AdverseMediaScanResult `json:"data,omitempty"`
	Error   string                           `json:"error,omitempty"`
}

// HistoricalDataRequest represents a request for historical data
type HistoricalDataRequest struct {
	EntityName string `json:"entity_name"`
	Days       int    `json:"days"`
}

// HistoricalDataResponse represents the response to historical data request
type HistoricalDataResponse struct {
	Success bool                           `json:"success"`
	Data    []news_api.AdverseMediaArticle `json:"data,omitempty"`
	Error   string                         `json:"error,omitempty"`
}

// TrendingDataRequest represents a request for trending data
type TrendingDataRequest struct {
	EntityName string `json:"entity_name"`
	Timeframe  string `json:"timeframe"` // "24h", "7d", "30d", "90d"
}

// TrendingDataResponse represents the response to trending data request
type TrendingDataResponse struct {
	Success bool                        `json:"success"`
	Data    *news_api.AdverseMediaTrend `json:"data,omitempty"`
	Error   string                      `json:"error,omitempty"`
}

// AlertsResponse represents the response to alerts request
type AlertsResponse struct {
	Success bool                         `json:"success"`
	Data    []news_api.AdverseMediaAlert `json:"data,omitempty"`
	Error   string                       `json:"error,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

// StartMonitoring starts adverse media monitoring for specified entities
func (h *AdverseMediaHandler) StartMonitoring(w http.ResponseWriter, r *http.Request) {
	var req StartMonitoringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode start monitoring request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.EntityNames) == 0 {
		http.Error(w, "Entity names are required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Starting adverse media monitoring",
		zap.Strings("entity_names", req.EntityNames))

	// In a real implementation, this would start a background monitoring process
	// For now, we'll just return a success response
	response := StartMonitoringResponse{
		Success:      true,
		Message:      "Monitoring started successfully",
		MonitoringID: "monitor_" + strconv.FormatInt(time.Now().UnixNano(), 10),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// StopMonitoring stops adverse media monitoring
func (h *AdverseMediaHandler) StopMonitoring(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	monitoringID := vars["id"]

	h.logger.Info("Stopping adverse media monitoring",
		zap.String("monitoring_id", monitoringID))

	// In a real implementation, this would stop the background monitoring process
	response := map[string]interface{}{
		"success": true,
		"message": "Monitoring stopped successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// PerformScan performs a single adverse media scan
func (h *AdverseMediaHandler) PerformScan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode scan request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.EntityNames) == 0 {
		http.Error(w, "Entity names are required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Performing adverse media scan",
		zap.Strings("entity_names", req.EntityNames))

	result, err := h.monitor.PerformScan(ctx, req.EntityNames)
	if err != nil {
		h.logger.Error("Failed to perform adverse media scan", zap.Error(err))
		response := ScanResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ScanResponse{
		Success: true,
		Data:    result,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetHistoricalData retrieves historical adverse media data
func (h *AdverseMediaHandler) GetHistoricalData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req HistoricalDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode historical data request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.EntityName == "" {
		http.Error(w, "Entity name is required", http.StatusBadRequest)
		return
	}

	if req.Days <= 0 {
		req.Days = 30 // Default to 30 days
	}

	h.logger.Info("Retrieving historical adverse media data",
		zap.String("entity_name", req.EntityName),
		zap.Int("days", req.Days))

	articles, err := h.monitor.GetHistoricalData(ctx, req.EntityName, req.Days)
	if err != nil {
		h.logger.Error("Failed to retrieve historical data", zap.Error(err))
		response := HistoricalDataResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := HistoricalDataResponse{
		Success: true,
		Data:    articles,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetTrendingData retrieves trending adverse media data
func (h *AdverseMediaHandler) GetTrendingData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req TrendingDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode trending data request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.EntityName == "" {
		http.Error(w, "Entity name is required", http.StatusBadRequest)
		return
	}

	if req.Timeframe == "" {
		req.Timeframe = "7d" // Default to 7 days
	}

	h.logger.Info("Retrieving trending adverse media data",
		zap.String("entity_name", req.EntityName),
		zap.String("timeframe", req.Timeframe))

	trend, err := h.monitor.GetTrendingData(ctx, req.EntityName, req.Timeframe)
	if err != nil {
		h.logger.Error("Failed to retrieve trending data", zap.Error(err))
		response := TrendingDataResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := TrendingDataResponse{
		Success: true,
		Data:    trend,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetAlerts retrieves active adverse media alerts
func (h *AdverseMediaHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	entityName := r.URL.Query().Get("entity_name")
	if entityName == "" {
		http.Error(w, "Entity name is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Retrieving adverse media alerts",
		zap.String("entity_name", entityName))

	alerts, err := h.monitor.GetAlerts(ctx, entityName)
	if err != nil {
		h.logger.Error("Failed to retrieve alerts", zap.Error(err))
		response := AlertsResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := AlertsResponse{
		Success: true,
		Data:    alerts,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ResolveAlert resolves an adverse media alert
func (h *AdverseMediaHandler) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alertID := vars["id"]

	var req struct {
		ResolvedBy      string `json:"resolved_by"`
		ResolutionNotes string `json:"resolution_notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode resolve alert request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Resolving adverse media alert",
		zap.String("alert_id", alertID),
		zap.String("resolved_by", req.ResolvedBy))

	// In a real implementation, this would update the alert in the database
	response := map[string]interface{}{
		"success":  true,
		"message":  "Alert resolved successfully",
		"alert_id": alertID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetHealth checks the health of the adverse media monitoring service
func (h *AdverseMediaHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := h.monitor.IsHealthy(ctx)
	status := "healthy"
	if err != nil {
		status = "unhealthy"
		h.logger.Error("Adverse media monitoring service unhealthy", zap.Error(err))
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Service:   "adverse_media_monitor",
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(response)
}

// GetRiskFactors generates risk factors from adverse media data
func (h *AdverseMediaHandler) GetRiskFactors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		EntityNames []string `json:"entity_names"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode risk factors request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.EntityNames) == 0 {
		http.Error(w, "Entity names are required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Generating risk factors from adverse media data",
		zap.Strings("entity_names", req.EntityNames))

	// Perform a scan to get current data
	scanResult, err := h.monitor.PerformScan(ctx, req.EntityNames)
	if err != nil {
		h.logger.Error("Failed to perform scan for risk factors", zap.Error(err))
		http.Error(w, "Failed to perform scan", http.StatusInternalServerError)
		return
	}

	// Generate risk factors
	riskFactors := h.monitor.GenerateRiskFactors(scanResult)

	response := map[string]interface{}{
		"success":      true,
		"risk_factors": riskFactors,
		"scan_data":    scanResult,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetArticleDetails retrieves detailed information about a specific article
func (h *AdverseMediaHandler) GetArticleDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	articleID := vars["id"]

	h.logger.Info("Retrieving article details",
		zap.String("article_id", articleID))

	// In a real implementation, this would retrieve the article from the database
	// For now, we'll return a mock response
	article := news_api.AdverseMediaArticle{
		ArticleID:        articleID,
		Title:            "Sample Adverse Media Article",
		Content:          "This is a sample adverse media article content...",
		URL:              "https://example.com/article/" + articleID,
		Source:           "Sample News Source",
		Author:           "John Doe",
		PublishedDate:    time.Now().Add(-24 * time.Hour),
		ScrapedDate:      time.Now(),
		Language:         "en",
		Country:          "US",
		RiskScore:        0.75,
		RiskLevel:        "high",
		Severity:         "severe",
		Category:         "financial_crime",
		Subcategory:      "fraud",
		Keywords:         []string{"fraud", "investigation", "regulatory"},
		Sentiment:        "negative",
		SentimentScore:   -0.8,
		MatchedEntities:  []string{"Sample Entity"},
		EntityConfidence: 0.95,
		DataQuality:      "excellent",
		LastUpdated:      time.Now(),
	}

	response := map[string]interface{}{
		"success": true,
		"article": article,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetMonitoringStatus retrieves the current monitoring status
func (h *AdverseMediaHandler) GetMonitoringStatus(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would retrieve the actual monitoring status
	// For now, we'll return a mock response
	status := map[string]interface{}{
		"is_monitoring":        true,
		"active_monitors":      5,
		"last_scan":            time.Now().Add(-1 * time.Hour),
		"next_scan":            time.Now().Add(5 * time.Hour),
		"total_articles":       1250,
		"high_risk_articles":   15,
		"medium_risk_articles": 45,
		"low_risk_articles":    1190,
		"active_alerts":        3,
	}

	response := map[string]interface{}{
		"success": true,
		"status":  status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
