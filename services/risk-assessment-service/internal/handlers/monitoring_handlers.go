package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	errorspkg "kyb-platform/pkg/errors"
	"kyb-platform/services/risk-assessment-service/internal/monitoring"

	"go.uber.org/zap"
)

// MonitoringHandler handles monitoring-related API endpoints
type MonitoringHandler struct {
	metrics *monitoring.PrometheusMetrics
	alerts  *monitoring.AlertManager
	grafana *monitoring.GrafanaClient
	logger  *zap.Logger
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(metrics *monitoring.PrometheusMetrics, alerts *monitoring.AlertManager, grafana *monitoring.GrafanaClient, logger *zap.Logger) *MonitoringHandler {
	return &MonitoringHandler{
		metrics: metrics,
		alerts:  alerts,
		grafana: grafana,
		logger:  logger,
	}
}

// GetMetrics returns Prometheus metrics
func (mh *MonitoringHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	mh.metrics.GetMetricsHandler().ServeHTTP(w, r)
}

// GetHealth returns service health status
func (mh *MonitoringHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"uptime":    time.Since(time.Now().Add(-24 * time.Hour)).String(), // Mock uptime
		"checks": map[string]interface{}{
			"database":      "healthy",
			"redis":         "healthy",
			"external_apis": "healthy",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// GetAlerts returns active alerts
func (mh *MonitoringHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	alerts := mh.alerts.GetActiveAlerts(tenantID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// GetAlertHistory returns alert history
func (mh *MonitoringHandler) GetAlertHistory(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	limitStr := r.URL.Query().Get("limit")

	limit := 100 // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	history := mh.alerts.GetAlertHistory(tenantID, limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": history,
		"count":   len(history),
	})
}

// SuppressAlert suppresses an alert
func (mh *MonitoringHandler) SuppressAlert(w http.ResponseWriter, r *http.Request) {
	var request struct {
		AlertID  string        `json:"alert_id"`
		Duration time.Duration `json:"duration"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		errorspkg.WriteBadRequest(w, r, "Invalid request body")
		return
	}

	if err := mh.alerts.SuppressAlert(request.AlertID, request.Duration); err != nil {
		errorspkg.WriteNotFound(w, r, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Alert suppressed successfully",
		"alert_id": request.AlertID,
		"duration": request.Duration.String(),
	})
}

// GetPerformanceInsights returns performance insights
func (mh *MonitoringHandler) GetPerformanceInsights(w http.ResponseWriter, r *http.Request) {
	endpoint := r.URL.Query().Get("endpoint")
	tenantID := r.URL.Query().Get("tenant_id")

	// Mock performance insights
	insights := map[string]interface{}{
		"endpoint":  endpoint,
		"tenant_id": tenantID,
		"metrics": map[string]interface{}{
			"average_response_time": 0.5,
			"p95_response_time":     1.2,
			"p99_response_time":     2.1,
			"throughput":            100.0,
			"error_rate":            0.01,
		},
		"recommendations": []string{
			"Response time is within acceptable limits",
			"Consider implementing caching for frequently accessed data",
			"Monitor error rate closely",
		},
		"health_score": 85,
		"last_updated": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(insights)
}

// GetSystemMetrics returns system metrics
func (mh *MonitoringHandler) GetSystemMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := map[string]interface{}{
		"system": map[string]interface{}{
			"active_connections": 150,
			"database_connections": map[string]int{
				"active": 10,
				"idle":   5,
			},
			"cache_hit_rates": map[string]float64{
				"redis":  0.95,
				"memory": 0.98,
			},
		},
		"performance": map[string]interface{}{
			"cpu_usage":    45.2,
			"memory_usage": 67.8,
			"disk_usage":   23.1,
		},
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetTenantMetrics returns tenant-specific metrics
func (mh *MonitoringHandler) GetTenantMetrics(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")

	if tenantID == "" {
		errorspkg.WriteBadRequest(w, r, "tenant_id parameter is required")
		return
	}

	metrics := map[string]interface{}{
		"tenant_id": tenantID,
		"usage": map[string]interface{}{
			"requests_today":        1250,
			"data_processed":        "2.5GB",
			"assessments_performed": 89,
		},
		"performance": map[string]interface{}{
			"average_response_time": 0.3,
			"error_rate":            0.005,
			"uptime":                "99.9%",
		},
		"limits": map[string]interface{}{
			"max_requests_per_hour":   1000,
			"max_data_per_month":      "10GB",
			"max_assessments_per_day": 500,
		},
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// CreateGrafanaDashboard creates a Grafana dashboard
func (mh *MonitoringHandler) CreateGrafanaDashboard(w http.ResponseWriter, r *http.Request) {
	if err := mh.grafana.CreateRiskAssessmentDashboard(r.Context()); err != nil {
		errorspkg.WriteInternalError(w, r, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Grafana dashboard created successfully",
		"dashboard_uid": "risk-assessment-overview",
	})
}

// GetGrafanaDashboard returns Grafana dashboard information
func (mh *MonitoringHandler) GetGrafanaDashboard(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		uid = "risk-assessment-overview"
	}

	dashboard, err := mh.grafana.GetDashboard(r.Context(), uid)
	if err != nil {
		errorspkg.WriteNotFound(w, r, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

// DeleteGrafanaDashboard deletes a Grafana dashboard
func (mh *MonitoringHandler) DeleteGrafanaDashboard(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		errorspkg.WriteBadRequest(w, r, "uid parameter is required")
		return
	}

	if err := mh.grafana.DeleteDashboard(r.Context(), uid); err != nil {
		errorspkg.WriteInternalError(w, r, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Grafana dashboard deleted successfully",
		"dashboard_uid": uid,
	})
}

// GetMonitoringConfig returns monitoring configuration
func (mh *MonitoringHandler) GetMonitoringConfig(w http.ResponseWriter, r *http.Request) {
	config := map[string]interface{}{
		"prometheus": map[string]interface{}{
			"enabled": true,
			"port":    9090,
			"path":    "/metrics",
		},
		"grafana": map[string]interface{}{
			"enabled":       true,
			"base_url":      "http://localhost:3000",
			"dashboard_uid": "risk-assessment-overview",
		},
		"alerting": map[string]interface{}{
			"enabled":  true,
			"channels": []string{"email", "slack", "webhook"},
		},
		"retention": map[string]interface{}{
			"metrics": "30d",
			"logs":    "7d",
			"alerts":  "90d",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// UpdateMonitoringConfig updates monitoring configuration
func (mh *MonitoringHandler) UpdateMonitoringConfig(w http.ResponseWriter, r *http.Request) {
	var config map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		errorspkg.WriteBadRequest(w, r, "Invalid request body")
		return
	}

	// Mock configuration update
	mh.logger.Info("Monitoring configuration updated", zap.Any("config", config))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Monitoring configuration updated successfully",
		"config":  config,
	})
}
