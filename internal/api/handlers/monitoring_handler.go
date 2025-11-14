package handlers

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"go.uber.org/zap"
)

// MonitoringHandler handles monitoring API endpoints
type MonitoringHandler struct {
	logger *zap.Logger
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(logger *zap.Logger) *MonitoringHandler {
	return &MonitoringHandler{
		logger: logger,
	}
}

// DashboardMetrics represents dashboard metrics
type DashboardMetrics struct {
	RequestRate        float64 `json:"request_rate"`
	ResponseTime       float64 `json:"response_time"`
	ErrorRate          float64 `json:"error_rate"`
	ActiveUsers        int64   `json:"active_users"`
	MemoryUsage        float64 `json:"memory_usage"`
	CPUUsage           float64 `json:"cpu_usage"`
	RequestRateChange  float64 `json:"request_rate_change"`
	ResponseTimeChange float64 `json:"response_time_change"`
	ErrorRateChange    float64 `json:"error_rate_change"`
	ActiveUsersChange  int64   `json:"active_users_change"`
	MemoryUsageChange  float64 `json:"memory_usage_change"`
	CPUUsageChange     float64 `json:"cpu_usage_change"`
	Timestamp          int64   `json:"timestamp"`
}

// Alert represents an alert
type MonitoringAlert struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Severity    string            `json:"severity"`
	Status      string            `json:"status"`
	Timestamp   int64             `json:"timestamp"`
	Labels      map[string]string `json:"labels"`
}

// HealthCheck represents a health check result
type MonitoringHealthCheck struct {
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	Message     string  `json:"message"`
	LastChecked int64   `json:"last_checked"`
	Duration    float64 `json:"duration"`
}

// GetMetrics returns current dashboard metrics
func (h *MonitoringHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Fetching dashboard metrics")

	metrics := h.collectMetrics()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		h.logger.Error("Failed to encode metrics response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetAlerts returns active alerts
func (h *MonitoringHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Fetching active alerts")

	alerts := h.getMockAlerts()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(alerts); err != nil {
		h.logger.Error("Failed to encode alerts response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetHealthChecks returns system health check results
func (h *MonitoringHandler) GetHealthChecks(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Fetching health checks")

	healthChecks := h.getMockHealthChecks()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(healthChecks); err != nil {
		h.logger.Error("Failed to encode health checks response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// collectMetrics collects current application metrics
func (h *MonitoringHandler) collectMetrics() DashboardMetrics {
	now := time.Now().Unix()

	// Get memory usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryUsage := float64(m.Alloc) / 1024 / 1024 / 1024 * 100 // Convert to percentage of 1GB

	// Mock metrics for demonstration
	requestRate := 75.0 + (float64(now%100)-50.0)/10.0
	responseTime := 150.0 + (float64(now%50)-25.0)/2.0
	errorRate := 1.5 + (float64(now%20)-10.0)/10.0
	activeUsers := 12 + int64(now%10) - 5
	cpuUsage := 35.0 + (float64(now%30)-15.0)/2.0

	// Calculate changes (simplified)
	requestRateChange := (requestRate - 75.0) / 75.0 * 100
	responseTimeChange := responseTime - 150.0
	errorRateChange := errorRate - 1.5
	activeUsersChange := activeUsers - 12
	memoryUsageChange := memoryUsage - 45.0
	cpuUsageChange := cpuUsage - 35.0

	return DashboardMetrics{
		RequestRate:        requestRate,
		ResponseTime:       responseTime,
		ErrorRate:          errorRate,
		ActiveUsers:        activeUsers,
		MemoryUsage:        memoryUsage,
		CPUUsage:           cpuUsage,
		RequestRateChange:  requestRateChange,
		ResponseTimeChange: responseTimeChange,
		ErrorRateChange:    errorRateChange,
		ActiveUsersChange:  activeUsersChange,
		MemoryUsageChange:  memoryUsageChange,
		CPUUsageChange:     cpuUsageChange,
		Timestamp:          now,
	}
}

// getMockAlerts returns mock alerts for demonstration
func (h *MonitoringHandler) getMockAlerts() []MonitoringAlert {
	alerts := []MonitoringAlert{}

	// Randomly show 0-2 alerts
	numAlerts := int(time.Now().Unix() % 3)

	alertTypes := []MonitoringAlert{
		{
			ID:          "alert-1",
			Title:       "High Error Rate",
			Description: "Error rate exceeded 5% threshold",
			Severity:    "critical",
			Status:      "active",
			Timestamp:   time.Now().Unix() - 300,
			Labels:      map[string]string{"service": "api"},
		},
		{
			ID:          "alert-2",
			Title:       "High Response Time",
			Description: "95th percentile response time is 2.5s",
			Severity:    "warning",
			Status:      "active",
			Timestamp:   time.Now().Unix() - 600,
			Labels:      map[string]string{"service": "api"},
		},
		{
			ID:          "alert-3",
			Title:       "Scheduled Maintenance",
			Description: "Database maintenance scheduled for tonight",
			Severity:    "info",
			Status:      "active",
			Timestamp:   time.Now().Unix() - 1800,
			Labels:      map[string]string{"service": "database"},
		},
	}

	for i := 0; i < numAlerts; i++ {
		alerts = append(alerts, alertTypes[i])
	}

	return alerts
}

// getMockHealthChecks returns mock health checks for demonstration
func (h *MonitoringHandler) getMockHealthChecks() []MonitoringHealthCheck {
	return []MonitoringHealthCheck{
		{
			Name:        "API Server",
			Status:      "healthy",
			Message:     "All endpoints responding",
			LastChecked: time.Now().Unix(),
			Duration:    0.05,
		},
		{
			Name:        "Database",
			Status:      "healthy",
			Message:     "Connection pool healthy",
			LastChecked: time.Now().Unix(),
			Duration:    0.12,
		},
		{
			Name:        "Redis Cache",
			Status:      "warning",
			Message:     "High memory usage",
			LastChecked: time.Now().Unix(),
			Duration:    0.08,
		},
		{
			Name:        "External APIs",
			Status:      "healthy",
			Message:     "All external services responding",
			LastChecked: time.Now().Unix(),
			Duration:    0.25,
		},
		{
			Name:        "File System",
			Status:      "healthy",
			Message:     "Disk space available",
			LastChecked: time.Now().Unix(),
			Duration:    0.02,
		},
		{
			Name:        "Memory",
			Status:      "healthy",
			Message:     "Memory usage normal",
			LastChecked: time.Now().Unix(),
			Duration:    0.01,
		},
	}
}
