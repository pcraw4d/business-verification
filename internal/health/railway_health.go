package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"kyb-platform/internal/observability"
)

// RailwayHealthChecker provides Railway-specific health checks for modular architecture
type RailwayHealthChecker struct {
	moduleHealthChecks map[string]ModuleHealthCheck
	overallHealth      *RailwayHealthStatus
	mu                 sync.RWMutex
	logger             *observability.Logger
	checkInterval      time.Duration
	lastCheckTime      time.Time
}

// ModuleHealthCheck represents a health check for a specific module
type ModuleHealthCheck struct {
	Name          string                 `json:"name"`
	Enabled       bool                   `json:"enabled"`
	Status        string                 `json:"status"`
	LastCheck     time.Time              `json:"last_check"`
	ResponseTime  time.Duration          `json:"response_time"`
	Error         string                 `json:"error,omitempty"`
	Details       map[string]interface{} `json:"details,omitempty"`
	CheckFunction func() error           `json:"-"`
}

// RailwayHealthStatus represents the overall health status for Railway
type RailwayHealthStatus struct {
	Status         string                       `json:"status"`
	Timestamp      time.Time                    `json:"timestamp"`
	Version        string                       `json:"version"`
	Environment    string                       `json:"environment"`
	Modules        map[string]ModuleHealthCheck `json:"modules"`
	OverallMetrics RailwayMetrics               `json:"metrics"`
	Ready          bool                         `json:"ready"`
	Live           bool                         `json:"live"`
}

// RailwayMetrics represents metrics for Railway health monitoring
type RailwayMetrics struct {
	TotalModules        int           `json:"total_modules"`
	HealthyModules      int           `json:"healthy_modules"`
	UnhealthyModules    int           `json:"unhealthy_modules"`
	DegradedModules     int           `json:"degraded_modules"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastCheckTime       time.Time     `json:"last_check_time"`
}

// NewRailwayHealthChecker creates a new Railway health checker
func NewRailwayHealthChecker(logger *observability.Logger) *RailwayHealthChecker {
	return &RailwayHealthChecker{
		moduleHealthChecks: make(map[string]ModuleHealthCheck),
		overallHealth: &RailwayHealthStatus{
			Status:      "unknown",
			Timestamp:   time.Now(),
			Version:     "1.0.0",
			Environment: "production",
			Modules:     make(map[string]ModuleHealthCheck),
			OverallMetrics: RailwayMetrics{
				LastCheckTime: time.Now(),
			},
			Ready: false,
			Live:  false,
		},
		logger:        logger,
		checkInterval: 30 * time.Second,
		lastCheckTime: time.Now(),
	}
}

// RegisterModuleHealthCheck registers a health check for a specific module
func (rhc *RailwayHealthChecker) RegisterModuleHealthCheck(
	name string,
	enabled bool,
	checkFunction func() error,
) {
	rhc.mu.Lock()
	defer rhc.mu.Unlock()

	rhc.moduleHealthChecks[name] = ModuleHealthCheck{
		Name:          name,
		Enabled:       enabled,
		Status:        "unknown",
		LastCheck:     time.Now(),
		CheckFunction: checkFunction,
		Details:       make(map[string]interface{}),
	}

	rhc.logger.Info("Module health check registered", map[string]interface{}{
		"module_name": name,
		"enabled":     enabled,
	})
}

// StartHealthCheckLoop starts the background health check loop
func (rhc *RailwayHealthChecker) StartHealthCheckLoop(ctx context.Context) {
	rhc.logger.Info("Starting Railway health check loop", map[string]interface{}{})

	ticker := time.NewTicker(rhc.checkInterval)
	defer ticker.Stop()

	// Run initial health check
	rhc.performHealthChecks()

	for {
		select {
		case <-ctx.Done():
			rhc.logger.Info("Stopping Railway health check loop", map[string]interface{}{})
			return
		case <-ticker.C:
			rhc.performHealthChecks()
		}
	}
}

// performHealthChecks performs health checks for all registered modules
func (rhc *RailwayHealthChecker) performHealthChecks() {
	rhc.mu.Lock()
	defer rhc.mu.Unlock()

	startTime := time.Now()
	healthyCount := 0
	unhealthyCount := 0
	degradedCount := 0
	totalResponseTime := time.Duration(0)

	// Perform health checks for each module
	for name, check := range rhc.moduleHealthChecks {
		if !check.Enabled {
			continue
		}

		checkStart := time.Now()
		err := check.CheckFunction()
		responseTime := time.Since(checkStart)

		// Update module health status
		check.LastCheck = time.Now()
		check.ResponseTime = responseTime
		totalResponseTime += responseTime

		if err != nil {
			check.Status = "unhealthy"
			check.Error = err.Error()
			unhealthyCount++
			rhc.logger.Warn("Module health check failed", map[string]interface{}{
				"module_name":   name,
				"error":         err.Error(),
				"response_time": responseTime,
			})
		} else {
			check.Status = "healthy"
			check.Error = ""
			healthyCount++
			rhc.logger.Debug("Module health check passed", map[string]interface{}{
				"module_name":   name,
				"response_time": responseTime,
			})
		}

		rhc.moduleHealthChecks[name] = check
	}

	// Update overall health status
	rhc.overallHealth.Timestamp = time.Now()
	rhc.overallHealth.Modules = rhc.moduleHealthChecks
	rhc.overallHealth.OverallMetrics = RailwayMetrics{
		TotalModules:        len(rhc.moduleHealthChecks),
		HealthyModules:      healthyCount,
		UnhealthyModules:    unhealthyCount,
		DegradedModules:     degradedCount,
		AverageResponseTime: totalResponseTime / time.Duration(len(rhc.moduleHealthChecks)),
		LastCheckTime:       time.Now(),
	}

	// Determine overall status
	if unhealthyCount == 0 && degradedCount == 0 {
		rhc.overallHealth.Status = "healthy"
		rhc.overallHealth.Ready = true
		rhc.overallHealth.Live = true
	} else if unhealthyCount == 0 {
		rhc.overallHealth.Status = "degraded"
		rhc.overallHealth.Ready = true
		rhc.overallHealth.Live = true
	} else {
		rhc.overallHealth.Status = "unhealthy"
		rhc.overallHealth.Ready = false
		rhc.overallHealth.Live = false
	}

	rhc.lastCheckTime = time.Now()

	rhc.logger.Info("Health checks completed", map[string]interface{}{
		"total_modules":  len(rhc.moduleHealthChecks),
		"healthy":        healthyCount,
		"unhealthy":      unhealthyCount,
		"degraded":       degradedCount,
		"overall_status": rhc.overallHealth.Status,
		"total_time":     time.Since(startTime),
	})
}

// GetHealthStatus returns the current health status
func (rhc *RailwayHealthChecker) GetHealthStatus() *RailwayHealthStatus {
	rhc.mu.RLock()
	defer rhc.mu.RUnlock()

	// Return a copy to avoid race conditions
	status := *rhc.overallHealth
	status.Modules = make(map[string]ModuleHealthCheck)
	for name, check := range rhc.overallHealth.Modules {
		status.Modules[name] = check
	}

	return &status
}

// GetModuleHealth returns the health status of a specific module
func (rhc *RailwayHealthChecker) GetModuleHealth(moduleName string) (*ModuleHealthCheck, error) {
	rhc.mu.RLock()
	defer rhc.mu.RUnlock()

	check, exists := rhc.moduleHealthChecks[moduleName]
	if !exists {
		return nil, fmt.Errorf("module health check not found: %s", moduleName)
	}

	return &check, nil
}

// ForceHealthCheck forces an immediate health check
func (rhc *RailwayHealthChecker) ForceHealthCheck() {
	rhc.logger.Info("Forcing immediate health check", map[string]interface{}{})
	rhc.performHealthChecks()
}

// Railway Health Check Handlers

// HealthHandler handles Railway health check endpoints
type RailwayHealthHandler struct {
	healthChecker *RailwayHealthChecker
	logger        *observability.Logger
}

// NewRailwayHealthHandler creates a new Railway health handler
func NewRailwayHealthHandler(healthChecker *RailwayHealthChecker, logger *observability.Logger) *RailwayHealthHandler {
	return &RailwayHealthHandler{
		healthChecker: healthChecker,
		logger:        logger,
	}
}

// HandleHealth handles the main health check endpoint
func (rhh *RailwayHealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	status := rhh.healthChecker.GetHealthStatus()

	// Set appropriate HTTP status code
	var httpStatus int
	switch status.Status {
	case "healthy":
		httpStatus = http.StatusOK
	case "degraded":
		httpStatus = http.StatusOK // Still OK but with warnings
	case "unhealthy":
		httpStatus = http.StatusServiceUnavailable
	default:
		httpStatus = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	response := map[string]interface{}{
		"status":      status.Status,
		"timestamp":   status.Timestamp,
		"version":     status.Version,
		"environment": status.Environment,
		"ready":       status.Ready,
		"live":        status.Live,
		"modules":     status.Modules,
		"metrics":     status.OverallMetrics,
	}

	json.NewEncoder(w).Encode(response)

	rhh.logger.Debug("Health check request served", map[string]interface{}{
		"status":      status.Status,
		"http_status": httpStatus,
		"client_ip":   r.RemoteAddr,
	})
}

// HandleReadiness handles the readiness probe endpoint
func (rhh *RailwayHealthHandler) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	status := rhh.healthChecker.GetHealthStatus()

	var httpStatus int
	if status.Ready {
		httpStatus = http.StatusOK
	} else {
		httpStatus = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	response := map[string]interface{}{
		"ready":     status.Ready,
		"timestamp": status.Timestamp,
		"status":    status.Status,
	}

	json.NewEncoder(w).Encode(response)

	rhh.logger.Debug("Readiness probe served", map[string]interface{}{
		"ready":       status.Ready,
		"http_status": httpStatus,
	})
}

// HandleLiveness handles the liveness probe endpoint
func (rhh *RailwayHealthHandler) HandleLiveness(w http.ResponseWriter, r *http.Request) {
	status := rhh.healthChecker.GetHealthStatus()

	var httpStatus int
	if status.Live {
		httpStatus = http.StatusOK
	} else {
		httpStatus = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	response := map[string]interface{}{
		"live":      status.Live,
		"timestamp": status.Timestamp,
		"status":    status.Status,
	}

	json.NewEncoder(w).Encode(response)

	rhh.logger.Debug("Liveness probe served", map[string]interface{}{
		"live":        status.Live,
		"http_status": httpStatus,
	})
}

// HandleModuleHealth handles module-specific health checks
func (rhh *RailwayHealthHandler) HandleModuleHealth(w http.ResponseWriter, r *http.Request) {
	moduleName := r.URL.Query().Get("module")
	if moduleName == "" {
		http.Error(w, "module parameter is required", http.StatusBadRequest)
		return
	}

	check, err := rhh.healthChecker.GetModuleHealth(moduleName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var httpStatus int
	switch check.Status {
	case "healthy":
		httpStatus = http.StatusOK
	case "degraded":
		httpStatus = http.StatusOK
	case "unhealthy":
		httpStatus = http.StatusServiceUnavailable
	default:
		httpStatus = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	json.NewEncoder(w).Encode(check)

	rhh.logger.Debug("Module health check served", map[string]interface{}{
		"module_name": moduleName,
		"status":      check.Status,
		"http_status": httpStatus,
	})
}

// HandleForceHealthCheck forces an immediate health check
func (rhh *RailwayHealthHandler) HandleForceHealthCheck(w http.ResponseWriter, r *http.Request) {
	rhh.healthChecker.ForceHealthCheck()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"message":   "Health check forced",
		"timestamp": time.Now(),
	}

	json.NewEncoder(w).Encode(response)

	rhh.logger.Info("Health check forced via HTTP request", map[string]interface{}{})
}

// Railway-specific health check functions

// CheckDatabaseHealth checks database connectivity
func CheckDatabaseHealth() error {
	// TODO: Implement actual database health check
	// This is a placeholder implementation
	return nil
}

// CheckCacheHealth checks cache connectivity
func CheckCacheHealth() error {
	// TODO: Implement actual cache health check
	// This is a placeholder implementation
	return nil
}

// CheckExternalAPIHealth checks external API connectivity
func CheckExternalAPIHealth() error {
	// TODO: Implement actual external API health check
	// This is a placeholder implementation
	return nil
}

// CheckModuleHealth checks if all modules are functioning
func CheckModuleHealth() error {
	// TODO: Implement actual module health check
	// This is a placeholder implementation
	return nil
}

// CheckObservabilityHealth checks observability systems
func CheckObservabilityHealth() error {
	// TODO: Implement actual observability health check
	// This is a placeholder implementation
	return nil
}

// CheckErrorResilienceHealth checks error resilience systems
func CheckErrorResilienceHealth() error {
	// TODO: Implement actual error resilience health check
	// This is a placeholder implementation
	return nil
}
