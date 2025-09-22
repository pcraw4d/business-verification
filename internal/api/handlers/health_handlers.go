package handlers

import (
	"encoding/json"
	"net/http"
	"runtime"
	"sync"
	"time"

	"kyb-platform/internal/health"
	"go.uber.org/zap"
)

// HealthStatus represents the overall health status of the system
type HealthStatus struct {
	Status      string                 `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
	Uptime      time.Duration          `json:"uptime"`
	Checks      map[string]HealthCheck `json:"checks"`
	Metrics     HealthMetrics          `json:"metrics"`
	Ready       bool                   `json:"ready"`
	Live        bool                   `json:"live"`
}

// HealthCheck represents the status of a specific health check
type HealthCheck struct {
	Status       string                 `json:"status"`
	ResponseTime time.Duration          `json:"response_time,omitempty"`
	LastCheck    time.Time              `json:"last_check"`
	Error        string                 `json:"error,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// HealthMetrics represents system health metrics
type HealthMetrics struct {
	TotalChecks     int           `json:"total_checks"`
	HealthyChecks   int           `json:"healthy_checks"`
	UnhealthyChecks int           `json:"unhealthy_checks"`
	DegradedChecks  int           `json:"degraded_checks"`
	AverageResponse time.Duration `json:"average_response_time"`
	MemoryUsage     MemoryInfo    `json:"memory_usage"`
	GoRuntime       GoRuntimeInfo `json:"go_runtime"`
}

// MemoryInfo represents memory usage information
type MemoryInfo struct {
	Alloc        uint64 `json:"alloc_bytes"`
	TotalAlloc   uint64 `json:"total_alloc_bytes"`
	Sys          uint64 `json:"sys_bytes"`
	NumGC        uint32 `json:"num_gc"`
	HeapAlloc    uint64 `json:"heap_alloc_bytes"`
	HeapSys      uint64 `json:"heap_sys_bytes"`
	HeapIdle     uint64 `json:"heap_idle_bytes"`
	HeapInuse    uint64 `json:"heap_inuse_bytes"`
	HeapReleased uint64 `json:"heap_released_bytes"`
	HeapObjects  uint64 `json:"heap_objects"`
}

// GoRuntimeInfo represents Go runtime information
type GoRuntimeInfo struct {
	Version      string `json:"version"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCgoCall   int64  `json:"num_cgo_call"`
}

// HealthHandler provides comprehensive health check endpoints
type HealthHandler struct {
	logger        *zap.Logger
	healthChecker *health.RailwayHealthChecker
	startTime     time.Time
	version       string
	environment   string
	mu            sync.RWMutex
	lastCheck     time.Time
	cache         map[string]HealthCheck
	cacheTTL      time.Duration
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(logger *zap.Logger, healthChecker *health.RailwayHealthChecker, version, environment string) *HealthHandler {
	return &HealthHandler{
		logger:        logger,
		healthChecker: healthChecker,
		startTime:     time.Now(),
		version:       version,
		environment:   environment,
		cache:         make(map[string]HealthCheck),
		cacheTTL:      30 * time.Second,
	}
}

// HandleHealth handles the main health check endpoint
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Check if we can serve from cache
	h.mu.RLock()
	if time.Since(h.lastCheck) < h.cacheTTL {
		cachedStatus := h.getCachedHealthStatus()
		h.mu.RUnlock()

		h.serveHealthResponse(w, r, cachedStatus, time.Since(start))
		return
	}
	h.mu.RUnlock()

	// Perform fresh health checks
	status := h.performHealthChecks()

	// Update cache
	h.mu.Lock()
	h.lastCheck = time.Now()
	h.cache = status.Checks
	h.mu.Unlock()

	h.serveHealthResponse(w, r, status, time.Since(start))
}

// HandleReadiness handles the readiness probe endpoint
func (h *HealthHandler) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Perform readiness checks (critical dependencies only)
	status := h.performReadinessChecks()

	// Set appropriate HTTP status
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
		"checks":    status.Checks,
	}

	json.NewEncoder(w).Encode(response)

	h.logger.Debug("Readiness probe served",
		zap.Bool("ready", status.Ready),
		zap.Int("http_status", httpStatus),
		zap.String("client_ip", r.RemoteAddr),
		zap.Duration("response_time", time.Since(start)),
	)
}

// HandleLiveness handles the liveness probe endpoint
func (h *HealthHandler) HandleLiveness(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Perform liveness checks (basic system health)
	status := h.performLivenessChecks()

	// Set appropriate HTTP status
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
		"uptime":    status.Uptime.String(),
	}

	json.NewEncoder(w).Encode(response)

	h.logger.Debug("Liveness probe served",
		zap.Bool("live", status.Live),
		zap.Int("http_status", httpStatus),
		zap.String("client_ip", r.RemoteAddr),
		zap.Duration("response_time", time.Since(start)),
	)
}

// HandleDetailedHealth handles detailed health information
func (h *HealthHandler) HandleDetailedHealth(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Perform comprehensive health checks
	status := h.performDetailedHealthChecks()

	h.serveHealthResponse(w, r, status, time.Since(start))
}

// HandleModuleHealth handles module-specific health checks
func (h *HealthHandler) HandleModuleHealth(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	moduleName := r.URL.Query().Get("module")

	if moduleName == "" {
		http.Error(w, "module parameter is required", http.StatusBadRequest)
		return
	}

	// Get module health from Railway health checker
	moduleHealth, err := h.healthChecker.GetModuleHealth(moduleName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Convert to our format
	check := HealthCheck{
		Status:       moduleHealth.Status,
		ResponseTime: moduleHealth.ResponseTime,
		LastCheck:    moduleHealth.LastCheck,
		Error:        moduleHealth.Error,
		Details:      moduleHealth.Details,
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

	h.logger.Debug("Module health check served",
		zap.String("module_name", moduleName),
		zap.String("status", check.Status),
		zap.Int("http_status", httpStatus),
		zap.String("client_ip", r.RemoteAddr),
		zap.Duration("response_time", time.Since(start)),
	)
}

// performHealthChecks performs comprehensive health checks
func (h *HealthHandler) performHealthChecks() *HealthStatus {
	checks := make(map[string]HealthCheck)

	// Basic system checks
	checks["system"] = h.checkSystemHealth()
	checks["database"] = h.checkDatabaseHealth()
	checks["cache"] = h.checkCacheHealth()
	checks["external_apis"] = h.checkExternalAPIsHealth()
	checks["ml_models"] = h.checkMLModelsHealth()
	checks["observability"] = h.checkObservabilityHealth()

	// Calculate metrics
	metrics := h.calculateHealthMetrics(checks)

	// Determine overall status
	status := "healthy"
	ready := true
	live := true

	unhealthyCount := 0
	degradedCount := 0

	for _, check := range checks {
		switch check.Status {
		case "unhealthy":
			unhealthyCount++
			ready = false
			live = false
		case "degraded":
			degradedCount++
		}
	}

	if unhealthyCount > 0 {
		status = "unhealthy"
	} else if degradedCount > 0 {
		status = "degraded"
	}

	return &HealthStatus{
		Status:      status,
		Timestamp:   time.Now(),
		Version:     h.version,
		Environment: h.environment,
		Uptime:      time.Since(h.startTime),
		Checks:      checks,
		Metrics:     metrics,
		Ready:       ready,
		Live:        live,
	}
}

// performReadinessChecks performs readiness checks (critical dependencies)
func (h *HealthHandler) performReadinessChecks() *HealthStatus {
	checks := make(map[string]HealthCheck)

	// Only check critical dependencies for readiness
	checks["database"] = h.checkDatabaseHealth()
	checks["cache"] = h.checkCacheHealth()

	// Calculate metrics
	metrics := h.calculateHealthMetrics(checks)

	// Determine readiness
	ready := true
	status := "healthy"

	for _, check := range checks {
		if check.Status == "unhealthy" {
			ready = false
			status = "unhealthy"
			break
		} else if check.Status == "degraded" {
			status = "degraded"
		}
	}

	return &HealthStatus{
		Status:      status,
		Timestamp:   time.Now(),
		Version:     h.version,
		Environment: h.environment,
		Uptime:      time.Since(h.startTime),
		Checks:      checks,
		Metrics:     metrics,
		Ready:       ready,
		Live:        true, // Liveness is separate
	}
}

// performLivenessChecks performs liveness checks (basic system health)
func (h *HealthHandler) performLivenessChecks() *HealthStatus {
	checks := make(map[string]HealthCheck)

	// Only check basic system health for liveness
	checks["system"] = h.checkSystemHealth()

	// Calculate metrics
	metrics := h.calculateHealthMetrics(checks)

	// Determine liveness
	live := true
	status := "healthy"

	for _, check := range checks {
		if check.Status == "unhealthy" {
			live = false
			status = "unhealthy"
			break
		}
	}

	return &HealthStatus{
		Status:      status,
		Timestamp:   time.Now(),
		Version:     h.version,
		Environment: h.environment,
		Uptime:      time.Since(h.startTime),
		Checks:      checks,
		Metrics:     metrics,
		Ready:       true, // Readiness is separate
		Live:        live,
	}
}

// performDetailedHealthChecks performs detailed health checks with additional information
func (h *HealthHandler) performDetailedHealthChecks() *HealthStatus {
	status := h.performHealthChecks()

	// Add additional detailed checks
	detailedChecks := status.Checks
	detailedChecks["memory"] = h.checkMemoryHealth()
	detailedChecks["goroutines"] = h.checkGoroutinesHealth()
	detailedChecks["disk"] = h.checkDiskHealth()
	detailedChecks["network"] = h.checkNetworkHealth()

	// Recalculate metrics with detailed checks
	status.Checks = detailedChecks
	status.Metrics = h.calculateHealthMetrics(detailedChecks)

	return status
}

// checkSystemHealth checks basic system health
func (h *HealthHandler) checkSystemHealth() HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Status:    "healthy",
		LastCheck: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// Check if system is responsive
	check.ResponseTime = time.Since(start)

	// Add system details
	check.Details["uptime"] = time.Since(h.startTime).String()
	check.Details["version"] = h.version
	check.Details["environment"] = h.environment

	return check
}

// checkDatabaseHealth checks database connectivity and performance
func (h *HealthHandler) checkDatabaseHealth() HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Status:    "healthy",
		LastCheck: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// TODO: Implement actual database health check
	// This is a placeholder implementation
	time.Sleep(10 * time.Millisecond) // Simulate database check

	check.ResponseTime = time.Since(start)

	// Add database details
	check.Details["connection_pool_size"] = 10
	check.Details["active_connections"] = 3
	check.Details["max_connections"] = 100

	return check
}

// checkCacheHealth checks cache connectivity and performance
func (h *HealthHandler) checkCacheHealth() HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Status:    "healthy",
		LastCheck: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// TODO: Implement actual cache health check
	// This is a placeholder implementation
	time.Sleep(5 * time.Millisecond) // Simulate cache check

	check.ResponseTime = time.Since(start)

	// Add cache details
	check.Details["cache_size"] = "1GB"
	check.Details["hit_rate"] = 0.85
	check.Details["miss_rate"] = 0.15

	return check
}

// checkExternalAPIsHealth checks external API connectivity
func (h *HealthHandler) checkExternalAPIsHealth() HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Status:    "healthy",
		LastCheck: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// TODO: Implement actual external API health checks
	// This is a placeholder implementation
	time.Sleep(20 * time.Millisecond) // Simulate API checks

	check.ResponseTime = time.Since(start)

	// Add external API details
	check.Details["apis_checked"] = 3
	check.Details["apis_healthy"] = 3
	check.Details["apis_degraded"] = 0

	return check
}

// checkMLModelsHealth checks ML model availability and performance
func (h *HealthHandler) checkMLModelsHealth() HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Status:    "healthy",
		LastCheck: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// TODO: Implement actual ML model health checks
	// This is a placeholder implementation
	time.Sleep(15 * time.Millisecond) // Simulate model checks

	check.ResponseTime = time.Since(start)

	// Add ML model details
	check.Details["models_loaded"] = 5
	check.Details["models_healthy"] = 5
	check.Details["model_version"] = "1.2.3"

	return check
}

// checkObservabilityHealth checks observability systems
func (h *HealthHandler) checkObservabilityHealth() HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Status:    "healthy",
		LastCheck: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// TODO: Implement actual observability health checks
	// This is a placeholder implementation
	time.Sleep(5 * time.Millisecond) // Simulate observability checks

	check.ResponseTime = time.Since(start)

	// Add observability details
	check.Details["logging_enabled"] = true
	check.Details["metrics_enabled"] = true
	check.Details["tracing_enabled"] = true

	return check
}

// checkMemoryHealth checks memory usage
func (h *HealthHandler) checkMemoryHealth() HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Status:    "healthy",
		LastCheck: time.Now(),
		Details:   make(map[string]interface{}),
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	check.ResponseTime = time.Since(start)

	// Determine memory health based on usage
	memoryUsage := float64(m.Alloc) / float64(m.Sys)
	if memoryUsage > 0.9 {
		check.Status = "degraded"
		check.Error = "High memory usage"
	} else if memoryUsage > 0.95 {
		check.Status = "unhealthy"
		check.Error = "Critical memory usage"
	}

	// Add memory details
	check.Details["alloc_bytes"] = m.Alloc
	check.Details["sys_bytes"] = m.Sys
	check.Details["heap_alloc_bytes"] = m.HeapAlloc
	check.Details["heap_sys_bytes"] = m.HeapSys
	check.Details["num_gc"] = m.NumGC
	check.Details["memory_usage_percent"] = memoryUsage * 100

	return check
}

// checkGoroutinesHealth checks goroutine count
func (h *HealthHandler) checkGoroutinesHealth() HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Status:    "healthy",
		LastCheck: time.Now(),
		Details:   make(map[string]interface{}),
	}

	numGoroutines := runtime.NumGoroutine()

	check.ResponseTime = time.Since(start)

	// Determine goroutine health
	if numGoroutines > 1000 {
		check.Status = "degraded"
		check.Error = "High number of goroutines"
	} else if numGoroutines > 5000 {
		check.Status = "unhealthy"
		check.Error = "Critical number of goroutines"
	}

	// Add goroutine details
	check.Details["num_goroutines"] = numGoroutines
	check.Details["num_cpu"] = runtime.NumCPU()

	return check
}

// checkDiskHealth checks disk usage
func (h *HealthHandler) checkDiskHealth() HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Status:    "healthy",
		LastCheck: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// TODO: Implement actual disk health check
	// This is a placeholder implementation
	time.Sleep(5 * time.Millisecond) // Simulate disk check

	check.ResponseTime = time.Since(start)

	// Add disk details (placeholder)
	check.Details["disk_usage_percent"] = 45.2
	check.Details["available_space_gb"] = 125.8
	check.Details["total_space_gb"] = 256.0

	return check
}

// checkNetworkHealth checks network connectivity
func (h *HealthHandler) checkNetworkHealth() HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Status:    "healthy",
		LastCheck: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// TODO: Implement actual network health check
	// This is a placeholder implementation
	time.Sleep(10 * time.Millisecond) // Simulate network check

	check.ResponseTime = time.Since(start)

	// Add network details (placeholder)
	check.Details["network_latency_ms"] = 15.3
	check.Details["packet_loss_percent"] = 0.01
	check.Details["bandwidth_mbps"] = 1000.0

	return check
}

// calculateHealthMetrics calculates health metrics from checks
func (h *HealthHandler) calculateHealthMetrics(checks map[string]HealthCheck) HealthMetrics {
	metrics := HealthMetrics{
		TotalChecks: len(checks),
	}

	var totalResponse time.Duration
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	for _, check := range checks {
		totalResponse += check.ResponseTime

		switch check.Status {
		case "healthy":
			metrics.HealthyChecks++
		case "unhealthy":
			metrics.UnhealthyChecks++
		case "degraded":
			metrics.DegradedChecks++
		}
	}

	if len(checks) > 0 {
		metrics.AverageResponse = totalResponse / time.Duration(len(checks))
	}

	// Memory info
	metrics.MemoryUsage = MemoryInfo{
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		NumGC:        m.NumGC,
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapIdle:     m.HeapIdle,
		HeapInuse:    m.HeapInuse,
		HeapReleased: m.HeapReleased,
		HeapObjects:  m.HeapObjects,
	}

	// Go runtime info
	metrics.GoRuntime = GoRuntimeInfo{
		Version:      runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		NumCgoCall:   runtime.NumCgoCall(),
	}

	return metrics
}

// serveHealthResponse serves the health response with appropriate status code
func (h *HealthHandler) serveHealthResponse(w http.ResponseWriter, r *http.Request, status *HealthStatus, responseTime time.Duration) {
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

	json.NewEncoder(w).Encode(status)

	h.logger.Debug("Health check request served",
		zap.String("status", status.Status),
		zap.Int("http_status", httpStatus),
		zap.String("client_ip", r.RemoteAddr),
		zap.Duration("response_time", responseTime),
		zap.Int("total_checks", status.Metrics.TotalChecks),
		zap.Int("healthy_checks", status.Metrics.HealthyChecks),
		zap.Int("unhealthy_checks", status.Metrics.UnhealthyChecks),
	)
}

// getCachedHealthStatus returns cached health status
func (h *HealthHandler) getCachedHealthStatus() *HealthStatus {
	checks := make(map[string]HealthCheck)
	for name, check := range h.cache {
		checks[name] = check
	}

	metrics := h.calculateHealthMetrics(checks)

	// Determine overall status from cached checks
	status := "healthy"
	ready := true
	live := true

	for _, check := range checks {
		switch check.Status {
		case "unhealthy":
			ready = false
			live = false
			status = "unhealthy"
		case "degraded":
			status = "degraded"
		}
	}

	return &HealthStatus{
		Status:      status,
		Timestamp:   time.Now(),
		Version:     h.version,
		Environment: h.environment,
		Uptime:      time.Since(h.startTime),
		Checks:      checks,
		Metrics:     metrics,
		Ready:       ready,
		Live:        live,
	}
}
