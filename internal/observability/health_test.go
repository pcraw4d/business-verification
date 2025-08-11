package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
)

// createTestLogger creates a logger for testing
func createTestLogger() *Logger {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "json",
	}
	return NewLogger(cfg)
}

func TestNewHealthChecker(t *testing.T) {
	logger := createTestLogger()
	hc := NewHealthChecker(nil, nil, logger, "1.0.0")

	if hc == nil {
		t.Fatal("Expected health checker to be created")
	}

	if hc.version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", hc.version)
	}

	if hc.logger != logger {
		t.Error("Expected logger to be set")
	}
}

func TestHealthChecker_ApplicationHealthCheck(t *testing.T) {
	logger := createTestLogger()
	hc := NewHealthChecker(nil, nil, logger, "1.0.0")

	ctx := context.Background()
	check := hc.applicationHealthCheck(ctx)

	if check.Name != "application" {
		t.Errorf("Expected name 'application', got %s", check.Name)
	}

	if check.Status != HealthStatusHealthy {
		t.Errorf("Expected status 'healthy', got %s", check.Status)
	}

	if check.Message != "Application is running" {
		t.Errorf("Expected message 'Application is running', got %s", check.Message)
	}

	if check.Duration < 0 {
		t.Error("Expected duration to be positive")
	}

	// Check details
	if _, ok := check.Details["uptime"]; !ok {
		t.Error("Expected uptime in details")
	}

	if _, ok := check.Details["version"]; !ok {
		t.Error("Expected version in details")
	}
}

func TestHealthChecker_DatabaseHealthCheck_NoDB(t *testing.T) {
	logger := createTestLogger()
	hc := NewHealthChecker(nil, nil, logger, "1.0.0")

	ctx := context.Background()
	check := hc.databaseHealthCheck(ctx)

	if check.Status != HealthStatusUnhealthy {
		t.Errorf("Expected status 'unhealthy', got %s", check.Status)
	}

	if check.Message != "Database connection not available" {
		t.Errorf("Expected message about database not available, got %s", check.Message)
	}
}

func TestHealthChecker_RedisHealthCheck_NoRedis(t *testing.T) {
	logger := NewLogger("test", "debug", "json")
	hc := NewHealthChecker(nil, nil, logger, "1.0.0")

	ctx := context.Background()
	check := hc.redisHealthCheck(ctx)

	if check.Status != HealthStatusUnhealthy {
		t.Errorf("Expected status 'unhealthy', got %s", check.Status)
	}

	if check.Message != "Redis connection not available" {
		t.Errorf("Expected message about Redis not available, got %s", check.Message)
	}
}

func TestHealthChecker_CheckHealth(t *testing.T) {
	logger := NewLogger("test", "debug", "json")
	hc := NewHealthChecker(nil, nil, logger, "1.0.0")

	ctx := context.Background()
	response := hc.CheckHealth(ctx)

	if response.Status != HealthStatusUnhealthy {
		t.Errorf("Expected status 'unhealthy' (due to missing DB/Redis), got %s", response.Status)
	}

	if response.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", response.Version)
	}

	if response.Uptime < 0 {
		t.Error("Expected uptime to be positive")
	}

	// Check that we have the expected checks
	expectedChecks := []string{"application", "database", "redis", "memory", "disk", "external_services"}
	for _, expected := range expectedChecks {
		if _, exists := response.Checks[expected]; !exists {
			t.Errorf("Expected check '%s' to exist", expected)
		}
	}

	// Check summary
	if response.Summary.Total != len(expectedChecks) {
		t.Errorf("Expected %d total checks, got %d", len(expectedChecks), response.Summary.Total)
	}
}

func TestHealthChecker_HealthHandler(t *testing.T) {
	logger := NewLogger("test", "debug", "json")
	hc := NewHealthChecker(nil, nil, logger, "1.0.0")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	hc.HealthHandler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status code %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Status != HealthStatusUnhealthy {
		t.Errorf("Expected status 'unhealthy', got %s", response.Status)
	}

	// Check headers
	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected Content-Type to be application/json")
	}

	if w.Header().Get("Cache-Control") != "no-cache, no-store, must-revalidate" {
		t.Error("Expected Cache-Control header")
	}
}

func TestHealthChecker_LivenessHandler(t *testing.T) {
	logger := NewLogger("test", "debug", "json")
	hc := NewHealthChecker(nil, nil, logger, "1.0.0")

	req := httptest.NewRequest("GET", "/live", nil)
	w := httptest.NewRecorder()

	hc.LivenessHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "alive" {
		t.Errorf("Expected status 'alive', got %v", response["status"])
	}
}

func TestHealthChecker_ReadinessHandler(t *testing.T) {
	logger := NewLogger("test", "debug", "json")
	hc := NewHealthChecker(nil, nil, logger, "1.0.0")

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	hc.ReadinessHandler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status code %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "not_ready" {
		t.Errorf("Expected status 'not_ready', got %v", response["status"])
	}

	checks, ok := response["checks"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected checks to be a map")
	}

	if checks["database"] != "unhealthy" {
		t.Errorf("Expected database to be unhealthy, got %v", checks["database"])
	}

	if checks["redis"] != "not_configured" {
		t.Errorf("Expected redis to be not_configured, got %v", checks["redis"])
	}
}

func TestHealthChecker_RegisterCheck(t *testing.T) {
	logger := NewLogger("test", "debug", "json")
	hc := NewHealthChecker(nil, nil, logger, "1.0.0")

	// Register a custom check
	customCheck := func(ctx context.Context) HealthCheck {
		return HealthCheck{
			Name:      "custom",
			Status:    HealthStatusHealthy,
			Message:   "Custom check passed",
			Timestamp: time.Now(),
		}
	}

	hc.RegisterCheck("custom", customCheck)

	ctx := context.Background()
	response := hc.CheckHealth(ctx)

	if _, exists := response.Checks["custom"]; !exists {
		t.Error("Expected custom check to exist")
	}

	check := response.Checks["custom"]
	if check.Status != HealthStatusHealthy {
		t.Errorf("Expected custom check to be healthy, got %s", check.Status)
	}

	if check.Message != "Custom check passed" {
		t.Errorf("Expected custom check message, got %s", check.Message)
	}
}

func TestHealthChecker_Timeout(t *testing.T) {
	logger := NewLogger("test", "debug", "json")
	hc := NewHealthChecker(nil, nil, logger, "1.0.0")

	// Register a slow check
	slowCheck := func(ctx context.Context) HealthCheck {
		select {
		case <-ctx.Done():
			return HealthCheck{
				Name:      "slow",
				Status:    HealthStatusUnhealthy,
				Message:   "Health check cancelled",
				Timestamp: time.Now(),
			}
		case <-time.After(2 * time.Second):
			return HealthCheck{
				Name:      "slow",
				Status:    HealthStatusHealthy,
				Message:   "Slow check completed",
				Timestamp: time.Now(),
			}
		}
	}

	hc.RegisterCheck("slow", slowCheck)

	// Test with timeout
	req := httptest.NewRequest("GET", "/health?timeout=1s", nil)
	w := httptest.NewRecorder()

	hc.HealthHandler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status code %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Status != HealthStatusUnhealthy {
		t.Errorf("Expected status 'unhealthy' due to timeout, got %s", response.Status)
	}
}

func TestHealthChecker_ConcurrentChecks(t *testing.T) {
	logger := NewLogger("test", "debug", "json")
	hc := NewHealthChecker(nil, nil, logger, "1.0.0")

	// Register multiple checks that take time
	for i := 0; i < 5; i++ {
		checkName := fmt.Sprintf("concurrent_%d", i)
		checkFunc := func(name string) HealthCheckFunc {
			return func(ctx context.Context) HealthCheck {
				time.Sleep(10 * time.Millisecond) // Simulate work
				return HealthCheck{
					Name:      name,
					Status:    HealthStatusHealthy,
					Message:   "Concurrent check completed",
					Timestamp: time.Now(),
				}
			}
		}(checkName)
		hc.RegisterCheck(checkName, checkFunc)
	}

	ctx := context.Background()
	start := time.Now()
	response := hc.CheckHealth(ctx)
	duration := time.Since(start)

	// All checks should complete quickly due to concurrency
	if duration > 100*time.Millisecond {
		t.Errorf("Expected concurrent checks to complete quickly, took %v", duration)
	}

	// All checks should be present
	for i := 0; i < 5; i++ {
		checkName := fmt.Sprintf("concurrent_%d", i)
		if _, exists := response.Checks[checkName]; !exists {
			t.Errorf("Expected check '%s' to exist", checkName)
		}
	}
}

func TestHealthStatus_String(t *testing.T) {
	tests := []struct {
		status HealthStatus
		want   string
	}{
		{HealthStatusHealthy, "healthy"},
		{HealthStatusDegraded, "degraded"},
		{HealthStatusUnhealthy, "unhealthy"},
	}

	for _, tt := range tests {
		if got := string(tt.status); got != tt.want {
			t.Errorf("HealthStatus.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestHealthChecker_SummaryCalculation(t *testing.T) {
	logger := NewLogger("test", "debug", "json")
	hc := NewHealthChecker(nil, nil, logger, "1.0.0")

	// Register checks with different statuses
	hc.RegisterCheck("healthy", func(ctx context.Context) HealthCheck {
		return HealthCheck{
			Name:      "healthy",
			Status:    HealthStatusHealthy,
			Timestamp: time.Now(),
		}
	})

	hc.RegisterCheck("degraded", func(ctx context.Context) HealthCheck {
		return HealthCheck{
			Name:      "degraded",
			Status:    HealthStatusDegraded,
			Timestamp: time.Now(),
		}
	})

	hc.RegisterCheck("unhealthy", func(ctx context.Context) HealthCheck {
		return HealthCheck{
			Name:      "unhealthy",
			Status:    HealthStatusUnhealthy,
			Timestamp: time.Now(),
		}
	})

	ctx := context.Background()
	response := hc.CheckHealth(ctx)

	// Check summary
	if response.Summary.Total != 9 { // 6 default + 3 custom
		t.Errorf("Expected 9 total checks, got %d", response.Summary.Total)
	}

	if response.Summary.Healthy < 1 {
		t.Error("Expected at least 1 healthy check")
	}

	if response.Summary.Degraded < 1 {
		t.Error("Expected at least 1 degraded check")
	}

	if response.Summary.Unhealthy < 1 {
		t.Error("Expected at least 1 unhealthy check")
	}

	// Overall status should be unhealthy
	if response.Status != HealthStatusUnhealthy {
		t.Errorf("Expected overall status 'unhealthy', got %s", response.Status)
	}
}
