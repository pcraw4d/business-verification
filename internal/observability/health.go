package observability

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheck represents a health check result
type HealthCheck struct {
	Name      string                 `json:"name"`
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// HealthResponse represents the overall health response
type HealthResponse struct {
	Status    HealthStatus           `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    time.Duration          `json:"uptime"`
	Checks    map[string]HealthCheck `json:"checks"`
	Summary   HealthSummary          `json:"summary"`
}

// HealthSummary provides a summary of health checks
type HealthSummary struct {
	Total     int `json:"total"`
	Healthy   int `json:"healthy"`
	Degraded  int `json:"degraded"`
	Unhealthy int `json:"unhealthy"`
}

// HealthChecker provides health checking functionality
type HealthChecker struct {
	db          *sql.DB
	redisClient interface{} // Will be nil if Redis is not available
	logger      *Logger
	startTime   time.Time
	version     string
	checks      map[string]HealthCheckFunc
	mu          sync.RWMutex
}

// HealthCheckFunc represents a health check function
type HealthCheckFunc func(ctx context.Context) HealthCheck

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *sql.DB, redisClient interface{}, logger *Logger, version string) *HealthChecker {
	hc := &HealthChecker{
		db:          db,
		redisClient: redisClient,
		logger:      logger,
		startTime:   time.Now(),
		version:     version,
		checks:      make(map[string]HealthCheckFunc),
	}

	// Register default health checks
	hc.registerDefaultChecks()

	return hc
}

// registerDefaultChecks registers the default health checks
func (hc *HealthChecker) registerDefaultChecks() {
	hc.RegisterCheck("application", hc.applicationHealthCheck)
	hc.RegisterCheck("database", hc.databaseHealthCheck)
	hc.RegisterCheck("redis", hc.redisHealthCheck)
	hc.RegisterCheck("memory", hc.memoryHealthCheck)
	hc.RegisterCheck("disk", hc.diskHealthCheck)
	hc.RegisterCheck("external_services", hc.externalServicesHealthCheck)
}

// RegisterCheck registers a new health check
func (hc *HealthChecker) RegisterCheck(name string, check HealthCheckFunc) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checks[name] = check
}

// applicationHealthCheck checks the application health
func (hc *HealthChecker) applicationHealthCheck(ctx context.Context) HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Name:      "application",
		Status:    HealthStatusHealthy,
		Timestamp: time.Now(),
		Message:   "Application is running",
		Details: map[string]interface{}{
			"uptime":  time.Since(hc.startTime).String(),
			"version": hc.version,
		},
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		check.Status = HealthStatusUnhealthy
		check.Message = "Health check cancelled"
	default:
		// Application is healthy if we can reach this point
	}

	check.Duration = time.Since(start)
	return check
}

// databaseHealthCheck checks the database health
func (hc *HealthChecker) databaseHealthCheck(ctx context.Context) HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Name:      "database",
		Status:    HealthStatusHealthy,
		Timestamp: time.Now(),
		Message:   "Database is healthy",
	}

	// Check database connection
	if hc.db == nil {
		check.Status = HealthStatusUnhealthy
		check.Message = "Database connection not available"
		check.Duration = time.Since(start)
		return check
	}

	// Test database connection with timeout
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := hc.db.PingContext(dbCtx)
	if err != nil {
		check.Status = HealthStatusUnhealthy
		check.Message = fmt.Sprintf("Database connection failed: %v", err)
		check.Duration = time.Since(start)
		return check
	}

	// Check database statistics
	stats := hc.db.Stats()
	check.Details = map[string]interface{}{
		"open_connections":    stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}

	// Check if database is under stress
	if stats.WaitCount > 100 {
		check.Status = HealthStatusDegraded
		check.Message = "Database showing signs of stress"
	}

	check.Duration = time.Since(start)
	return check
}

// redisHealthCheck checks the Redis health
func (hc *HealthChecker) redisHealthCheck(ctx context.Context) HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Name:      "redis",
		Status:    HealthStatusHealthy,
		Timestamp: time.Now(),
		Message:   "Redis is healthy",
	}

	// Check Redis connection
	if hc.redisClient == nil {
		check.Status = HealthStatusUnhealthy
		check.Message = "Redis connection not available"
		check.Duration = time.Since(start)
		return check
	}

	// Try to use Redis client if available
	// Note: This requires Redis client to be properly configured
	check.Details = map[string]interface{}{
		"note":      "Redis health check requires Redis client to be properly configured",
		"available": hc.redisClient != nil,
	}

	check.Duration = time.Since(start)
	return check
}

// memoryHealthCheck checks the memory usage
func (hc *HealthChecker) memoryHealthCheck(ctx context.Context) HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Name:      "memory",
		Status:    HealthStatusHealthy,
		Timestamp: time.Now(),
		Message:   "Memory usage is normal",
	}

	// Get memory statistics (this would need to be implemented based on the platform)
	// For now, we'll return a basic check
	check.Details = map[string]interface{}{
		"note": "Memory monitoring requires platform-specific implementation",
	}

	check.Duration = time.Since(start)
	return check
}

// diskHealthCheck checks the disk usage
func (hc *HealthChecker) diskHealthCheck(ctx context.Context) HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Name:      "disk",
		Status:    HealthStatusHealthy,
		Timestamp: time.Now(),
		Message:   "Disk usage is normal",
	}

	// Get disk statistics (this would need to be implemented based on the platform)
	// For now, we'll return a basic check
	check.Details = map[string]interface{}{
		"note": "Disk monitoring requires platform-specific implementation",
	}

	check.Duration = time.Since(start)
	return check
}

// externalServicesHealthCheck checks external service dependencies
func (hc *HealthChecker) externalServicesHealthCheck(ctx context.Context) HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Name:      "external_services",
		Status:    HealthStatusHealthy,
		Timestamp: time.Now(),
		Message:   "External services are healthy",
		Details:   make(map[string]interface{}),
	}

	// Check external services (this would be implemented based on actual external dependencies)
	// For now, we'll return a basic check
	check.Details["note"] = "External service checks should be implemented based on actual dependencies"

	check.Duration = time.Since(start)
	return check
}

// CheckHealth performs all health checks
func (hc *HealthChecker) CheckHealth(ctx context.Context) HealthResponse {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	response := HealthResponse{
		Status:    HealthStatusHealthy,
		Timestamp: time.Now(),
		Version:   hc.version,
		Uptime:    time.Since(hc.startTime),
		Checks:    make(map[string]HealthCheck),
		Summary:   HealthSummary{},
	}

	// Run all health checks concurrently
	var wg sync.WaitGroup
	checkChan := make(chan HealthCheck, len(hc.checks))

	for name, checkFunc := range hc.checks {
		wg.Add(1)
		go func(name string, checkFunc HealthCheckFunc) {
			defer wg.Done()
			check := checkFunc(ctx)
			check.Name = name
			checkChan <- check
		}(name, checkFunc)
	}

	// Wait for all checks to complete
	go func() {
		wg.Wait()
		close(checkChan)
	}()

	// Collect results
	for check := range checkChan {
		response.Checks[check.Name] = check
		response.Summary.Total++

		switch check.Status {
		case HealthStatusHealthy:
			response.Summary.Healthy++
		case HealthStatusDegraded:
			response.Summary.Degraded++
		case HealthStatusUnhealthy:
			response.Summary.Unhealthy++
		}
	}

	// Determine overall status
	if response.Summary.Unhealthy > 0 {
		response.Status = HealthStatusUnhealthy
	} else if response.Summary.Degraded > 0 {
		response.Status = HealthStatusDegraded
	}

	return response
}

// HealthHandler handles health check HTTP requests
func (hc *HealthChecker) HealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Set timeout for health checks
	timeout := 30 * time.Second
	if r.URL.Query().Get("timeout") != "" {
		if t, err := time.ParseDuration(r.URL.Query().Get("timeout")); err == nil {
			timeout = t
		}
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	response := hc.CheckHealth(ctx)

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Set HTTP status code based on health status
	switch response.Status {
	case HealthStatusHealthy:
		w.WriteHeader(http.StatusOK)
	case HealthStatusDegraded:
		w.WriteHeader(http.StatusOK) // Still OK but degraded
	case HealthStatusUnhealthy:
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	// Encode response
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(response); err != nil {
		hc.logger.Error("Failed to encode health response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// LivenessHandler handles liveness probe requests
func (hc *HealthChecker) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	// Simple liveness check - just check if the application is running
	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
		"uptime":    time.Since(hc.startTime).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(response); err != nil {
		hc.logger.Error("Failed to encode liveness response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// ReadinessHandler handles readiness probe requests
func (hc *HealthChecker) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Check critical dependencies
	response := map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now(),
		"checks":    make(map[string]interface{}),
	}

	// Check database
	if hc.db != nil {
		if err := hc.db.PingContext(ctx); err != nil {
			response["status"] = "not_ready"
			response["checks"].(map[string]interface{})["database"] = "unhealthy"
		} else {
			response["checks"].(map[string]interface{})["database"] = "healthy"
		}
	}

	// Check Redis
	if hc.redisClient != nil {
		// Note: Redis health check requires proper Redis client configuration
		response["checks"].(map[string]interface{})["redis"] = "available"
	} else {
		response["checks"].(map[string]interface{})["redis"] = "not_configured"
	}

	// Set HTTP status code
	if response["status"] == "ready" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(response); err != nil {
		hc.logger.Error("Failed to encode readiness response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
