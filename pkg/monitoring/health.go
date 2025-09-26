package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HealthChecker defines the interface for health checks
type HealthChecker interface {
	CheckHealth(ctx context.Context) HealthStatus
	GetName() string
}

// HealthStatus represents the status of a health check
type HealthStatus struct {
	Name      string                 `json:"name"`
	Status    string                 `json:"status"` // "healthy", "unhealthy", "degraded"
	Message   string                 `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
}

// HealthManager manages multiple health checks
type HealthManager struct {
	checkers []HealthChecker
}

// NewHealthManager creates a new health manager
func NewHealthManager() *HealthManager {
	return &HealthManager{
		checkers: make([]HealthChecker, 0),
	}
}

// AddChecker adds a health checker
func (hm *HealthManager) AddChecker(checker HealthChecker) {
	hm.checkers = append(hm.checkers, checker)
}

// CheckAll performs all health checks
func (hm *HealthManager) CheckAll(ctx context.Context) map[string]interface{} {
	results := make(map[string]interface{})
	overallStatus := "healthy"

	for _, checker := range hm.checkers {
		start := time.Now()
		status := checker.CheckHealth(ctx)
		status.Duration = time.Since(start)
		status.Timestamp = time.Now()

		results[checker.GetName()] = status

		// Update overall status
		if status.Status == "unhealthy" {
			overallStatus = "unhealthy"
		} else if status.Status == "degraded" && overallStatus == "healthy" {
			overallStatus = "degraded"
		}
	}

	return map[string]interface{}{
		"status":    overallStatus,
		"timestamp": time.Now().Format(time.RFC3339),
		"checks":    results,
	}
}

// HTTPHealthHandler provides HTTP endpoint for health checks
func (hm *HealthManager) HTTPHealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	results := hm.CheckAll(ctx)

	w.Header().Set("Content-Type", "application/json")

	// Set HTTP status based on overall health
	if results["status"] == "unhealthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else if results["status"] == "degraded" {
		w.WriteHeader(http.StatusOK) // Still OK, but degraded
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(results)
}

// DatabaseHealthChecker checks database connectivity
type DatabaseHealthChecker struct {
	name      string
	checkFunc func(ctx context.Context) error
}

// NewDatabaseHealthChecker creates a new database health checker
func NewDatabaseHealthChecker(name string, checkFunc func(ctx context.Context) error) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{
		name:      name,
		checkFunc: checkFunc,
	}
}

func (dhc *DatabaseHealthChecker) GetName() string {
	return dhc.name
}

func (dhc *DatabaseHealthChecker) CheckHealth(ctx context.Context) HealthStatus {
	err := dhc.checkFunc(ctx)
	if err != nil {
		return HealthStatus{
			Name:    dhc.name,
			Status:  "unhealthy",
			Message: fmt.Sprintf("Database check failed: %v", err),
		}
	}

	return HealthStatus{
		Name:    dhc.name,
		Status:  "healthy",
		Message: "Database connection successful",
	}
}

// CacheHealthChecker checks cache connectivity
type CacheHealthChecker struct {
	name      string
	checkFunc func(ctx context.Context) error
}

// NewCacheHealthChecker creates a new cache health checker
func NewCacheHealthChecker(name string, checkFunc func(ctx context.Context) error) *CacheHealthChecker {
	return &CacheHealthChecker{
		name:      name,
		checkFunc: checkFunc,
	}
}

func (chc *CacheHealthChecker) GetName() string {
	return chc.name
}

func (chc *CacheHealthChecker) CheckHealth(ctx context.Context) HealthStatus {
	err := chc.checkFunc(ctx)
	if err != nil {
		return HealthStatus{
			Name:    chc.name,
			Status:  "degraded", // Cache failure is degraded, not unhealthy
			Message: fmt.Sprintf("Cache check failed: %v", err),
		}
	}

	return HealthStatus{
		Name:    chc.name,
		Status:  "healthy",
		Message: "Cache connection successful",
	}
}

// ExternalServiceHealthChecker checks external service connectivity
type ExternalServiceHealthChecker struct {
	name      string
	url       string
	checkFunc func(ctx context.Context, url string) error
}

// NewExternalServiceHealthChecker creates a new external service health checker
func NewExternalServiceHealthChecker(name, url string, checkFunc func(ctx context.Context, url string) error) *ExternalServiceHealthChecker {
	return &ExternalServiceHealthChecker{
		name:      name,
		url:       url,
		checkFunc: checkFunc,
	}
}

func (eshc *ExternalServiceHealthChecker) GetName() string {
	return eshc.name
}

func (eshc *ExternalServiceHealthChecker) CheckHealth(ctx context.Context) HealthStatus {
	err := eshc.checkFunc(ctx, eshc.url)
	if err != nil {
		return HealthStatus{
			Name:    eshc.name,
			Status:  "degraded", // External service failure is degraded
			Message: fmt.Sprintf("External service check failed: %v", err),
			Details: map[string]interface{}{
				"url": eshc.url,
			},
		}
	}

	return HealthStatus{
		Name:    eshc.name,
		Status:  "healthy",
		Message: "External service connection successful",
		Details: map[string]interface{}{
			"url": eshc.url,
		},
	}
}
