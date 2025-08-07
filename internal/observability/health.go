package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	Healthy   HealthStatus = "healthy"
	Unhealthy HealthStatus = "unhealthy"
	Degraded  HealthStatus = "degraded"
)

// HealthCheck represents a health check result
type HealthCheck struct {
	Component   string                 `json:"component"`
	Status      HealthStatus           `json:"status"`
	Message     string                 `json:"message,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	LastChecked time.Time              `json:"last_checked"`
	Duration    time.Duration          `json:"duration_ms"`
}

// HealthResponse represents the overall health response
type HealthResponse struct {
	Status      HealthStatus           `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
	Checks      map[string]HealthCheck `json:"checks"`
	Summary     HealthSummary          `json:"summary"`
}

// HealthSummary provides a summary of health checks
type HealthSummary struct {
	Total     int `json:"total"`
	Healthy   int `json:"healthy"`
	Unhealthy int `json:"unhealthy"`
	Degraded  int `json:"degraded"`
}

// HealthChecker defines the interface for health checks
type HealthChecker interface {
	Check(ctx context.Context) HealthCheck
}

// HealthManager manages health checks
type HealthManager struct {
	config      *config.ObservabilityConfig
	logger      *Logger
	checkers    map[string]HealthChecker
	version     string
	environment string
}

// NewHealthManager creates a new health manager
func NewHealthManager(cfg *config.ObservabilityConfig, logger *Logger, version, environment string) *HealthManager {
	return &HealthManager{
		config:      cfg,
		logger:      logger,
		checkers:    make(map[string]HealthChecker),
		version:     version,
		environment: environment,
	}
}

// AddChecker adds a health checker
func (hm *HealthManager) AddChecker(name string, checker HealthChecker) {
	hm.checkers[name] = checker
}

// RemoveChecker removes a health checker
func (hm *HealthManager) RemoveChecker(name string) {
	delete(hm.checkers, name)
}

// CheckAll performs all health checks
func (hm *HealthManager) CheckAll(ctx context.Context) HealthResponse {
	start := time.Now()
	checks := make(map[string]HealthCheck)

	// Perform all health checks
	for name, checker := range hm.checkers {
		checkStart := time.Now()
		check := checker.Check(ctx)
		check.Duration = time.Since(checkStart)
		check.LastChecked = time.Now()
		checks[name] = check

		// Log health check result
		hm.logger.LogHealthCheck(name, string(check.Status), check.Details)
	}

	// Calculate summary
	summary := hm.calculateSummary(checks)

	// Determine overall status
	overallStatus := hm.determineOverallStatus(summary)

	response := HealthResponse{
		Status:      overallStatus,
		Timestamp:   time.Now(),
		Version:     hm.version,
		Environment: hm.environment,
		Checks:      checks,
		Summary:     summary,
	}

	// Log overall health status
	hm.logger.WithFields(map[string]interface{}{
		"overall_status": string(overallStatus),
		"total_checks":   summary.Total,
		"healthy":        summary.Healthy,
		"unhealthy":      summary.Unhealthy,
		"degraded":       summary.Degraded,
		"duration_ms":    time.Since(start).Milliseconds(),
	}).Info("Health check completed")

	return response
}

// calculateSummary calculates the summary of health checks
func (hm *HealthManager) calculateSummary(checks map[string]HealthCheck) HealthSummary {
	summary := HealthSummary{
		Total: len(checks),
	}

	for _, check := range checks {
		switch check.Status {
		case Healthy:
			summary.Healthy++
		case Unhealthy:
			summary.Unhealthy++
		case Degraded:
			summary.Degraded++
		}
	}

	return summary
}

// determineOverallStatus determines the overall health status
func (hm *HealthManager) determineOverallStatus(summary HealthSummary) HealthStatus {
	if summary.Unhealthy > 0 {
		return Unhealthy
	}
	if summary.Degraded > 0 {
		return Degraded
	}
	return Healthy
}

// ServeHTTP serves the health check endpoint
func (hm *HealthManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")

	// Perform health checks
	response := hm.CheckAll(ctx)

	// Set status code based on overall health
	switch response.Status {
	case Healthy:
		w.WriteHeader(http.StatusOK)
	case Degraded:
		w.WriteHeader(http.StatusOK) // Still OK but degraded
	case Unhealthy:
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode health response", http.StatusInternalServerError)
		return
	}
}

// BasicHealthChecker provides basic health checks
type BasicHealthChecker struct {
	config *config.ObservabilityConfig
}

// NewBasicHealthChecker creates a new basic health checker
func NewBasicHealthChecker(cfg *config.ObservabilityConfig) *BasicHealthChecker {
	return &BasicHealthChecker{
		config: cfg,
	}
}

// Check performs basic health checks
func (bhc *BasicHealthChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()

	// Check runtime statistics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	details := map[string]interface{}{
		"goroutines":      runtime.NumGoroutine(),
		"memory_alloc":    m.Alloc,
		"memory_sys":      m.Sys,
		"memory_heap":     m.HeapAlloc,
		"memory_heap_sys": m.HeapSys,
		"gc_cycles":       m.NumGC,
		"uptime_seconds":  time.Since(start).Seconds(),
	}

	// Determine status based on memory usage
	status := Healthy
	message := "Application is healthy"

	// Check if memory usage is high (simple heuristic)
	if m.Sys > 1<<30 { // 1GB
		status = Degraded
		message = "High memory usage detected"
	}

	// Check if too many goroutines
	if runtime.NumGoroutine() > 1000 {
		status = Unhealthy
		message = "Too many goroutines"
	}

	return HealthCheck{
		Component:   "application",
		Status:      status,
		Message:     message,
		Details:     details,
		LastChecked: time.Now(),
		Duration:    time.Since(start),
	}
}

// DatabaseHealthChecker provides database health checks
type DatabaseHealthChecker struct {
	pingFunc func(context.Context) error
}

// NewDatabaseHealthChecker creates a new database health checker
func NewDatabaseHealthChecker(pingFunc func(context.Context) error) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{
		pingFunc: pingFunc,
	}
}

// Check performs database health checks
func (dhc *DatabaseHealthChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()

	if dhc.pingFunc == nil {
		return HealthCheck{
			Component:   "database",
			Status:      Unhealthy,
			Message:     "Database checker not configured",
			LastChecked: time.Now(),
			Duration:    time.Since(start),
		}
	}

	err := dhc.pingFunc(ctx)
	if err != nil {
		return HealthCheck{
			Component: "database",
			Status:    Unhealthy,
			Message:   fmt.Sprintf("Database connection failed: %v", err),
			Details: map[string]interface{}{
				"error": err.Error(),
			},
			LastChecked: time.Now(),
			Duration:    time.Since(start),
		}
	}

	return HealthCheck{
		Component:   "database",
		Status:      Healthy,
		Message:     "Database connection is healthy",
		LastChecked: time.Now(),
		Duration:    time.Since(start),
	}
}

// ExternalServiceHealthChecker provides external service health checks
type ExternalServiceHealthChecker struct {
	serviceName string
	checkFunc   func(context.Context) error
}

// NewExternalServiceHealthChecker creates a new external service health checker
func NewExternalServiceHealthChecker(serviceName string, checkFunc func(context.Context) error) *ExternalServiceHealthChecker {
	return &ExternalServiceHealthChecker{
		serviceName: serviceName,
		checkFunc:   checkFunc,
	}
}

// Check performs external service health checks
func (eshc *ExternalServiceHealthChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()

	if eshc.checkFunc == nil {
		return HealthCheck{
			Component:   eshc.serviceName,
			Status:      Unhealthy,
			Message:     "External service checker not configured",
			LastChecked: time.Now(),
			Duration:    time.Since(start),
		}
	}

	err := eshc.checkFunc(ctx)
	if err != nil {
		return HealthCheck{
			Component: eshc.serviceName,
			Status:    Unhealthy,
			Message:   fmt.Sprintf("External service check failed: %v", err),
			Details: map[string]interface{}{
				"error": err.Error(),
			},
			LastChecked: time.Now(),
			Duration:    time.Since(start),
		}
	}

	return HealthCheck{
		Component:   eshc.serviceName,
		Status:      Healthy,
		Message:     fmt.Sprintf("%s is healthy", eshc.serviceName),
		LastChecked: time.Now(),
		Duration:    time.Since(start),
	}
}

// StartHealthServer starts the health check server
func (hm *HealthManager) StartHealthServer(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.Handle(hm.config.HealthCheckPath, hm)

	server := &http.Server{
		Addr:    ":8081", // Health check port
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Health server error: %v\n", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return server.Shutdown(shutdownCtx)
}
