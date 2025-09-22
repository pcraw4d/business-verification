package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kyb-platform/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRailwayHealthChecker_NewRailwayHealthChecker(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)

	checker := NewRailwayHealthChecker(logger)

	assert.NotNil(t, checker)
	assert.NotNil(t, checker.moduleHealthChecks)
	assert.NotNil(t, checker.overallHealth)
	assert.NotNil(t, checker.logger)
	assert.Equal(t, 30*time.Second, checker.checkInterval)
	assert.Equal(t, 0, len(checker.moduleHealthChecks))
}

func TestRailwayHealthChecker_RegisterModuleHealthCheck(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	checker := NewRailwayHealthChecker(logger)

	// Register a module health check
	checker.RegisterModuleHealthCheck("test-module", true, func() error {
		return nil
	})

	// Verify the module is registered
	check, err := checker.GetModuleHealth("test-module")
	require.NoError(t, err)
	assert.Equal(t, "test-module", check.Name)
	assert.True(t, check.Enabled)
	assert.Equal(t, "unknown", check.Status)
}

func TestRailwayHealthChecker_GetModuleHealth_NotFound(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	checker := NewRailwayHealthChecker(logger)

	// Try to get non-existent module
	check, err := checker.GetModuleHealth("non-existent")
	assert.Error(t, err)
	assert.Nil(t, check)
	assert.Contains(t, err.Error(), "not found")
}

func TestRailwayHealthChecker_GetHealthStatus(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	checker := NewRailwayHealthChecker(logger)

	// Register some module health checks
	checker.RegisterModuleHealthCheck("healthy-module", true, func() error {
		return nil
	})

	checker.RegisterModuleHealthCheck("unhealthy-module", true, func() error {
		return assert.AnError
	})

	// Perform health checks
	checker.performHealthChecks()

	// Get health status
	status := checker.GetHealthStatus()

	assert.NotNil(t, status)
	assert.Equal(t, "unhealthy", status.Status)
	assert.False(t, status.Ready)
	assert.False(t, status.Live)
	assert.Equal(t, 2, status.OverallMetrics.TotalModules)
	assert.Equal(t, 1, status.OverallMetrics.HealthyModules)
	assert.Equal(t, 1, status.OverallMetrics.UnhealthyModules)
}

func TestRailwayHealthChecker_StartHealthCheckLoop(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	checker := NewRailwayHealthChecker(logger)

	// Register a module health check
	checker.RegisterModuleHealthCheck("test-module", true, func() error {
		return nil
	})

	// Start health check loop with short interval for testing
	checker.checkInterval = 100 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// Start the health check loop
	go checker.StartHealthCheckLoop(ctx)

	// Wait for the context to be cancelled
	<-ctx.Done()

	// Verify that health checks were performed
	status := checker.GetHealthStatus()
	assert.NotNil(t, status)
	assert.Equal(t, "healthy", status.Status)
	assert.True(t, status.Ready)
	assert.True(t, status.Live)
}

func TestRailwayHealthChecker_ForceHealthCheck(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	checker := NewRailwayHealthChecker(logger)

	// Register a module health check
	checker.RegisterModuleHealthCheck("test-module", true, func() error {
		return nil
	})

	// Force health check
	checker.ForceHealthCheck()

	// Verify health check was performed
	status := checker.GetHealthStatus()
	assert.NotNil(t, status)
	assert.Equal(t, "healthy", status.Status)
}

func TestRailwayHealthHandler_NewRailwayHealthHandler(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	checker := NewRailwayHealthChecker(logger)

	handler := NewRailwayHealthHandler(checker, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, checker, handler.healthChecker)
	assert.Equal(t, logger, handler.logger)
}

func TestRailwayHealthHandler_HandleHealth(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	checker := NewRailwayHealthChecker(logger)

	// Register a healthy module
	checker.RegisterModuleHealthCheck("test-module", true, func() error {
		return nil
	})

	// Perform health checks
	checker.performHealthChecks()

	handler := NewRailwayHealthHandler(checker, logger)

	// Create test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Handle health request
	handler.HandleHealth(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.True(t, response["ready"].(bool))
	assert.True(t, response["live"].(bool))
	assert.NotNil(t, response["modules"])
	assert.NotNil(t, response["metrics"])
}

func TestRailwayHealthHandler_HandleHealth_Unhealthy(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	checker := NewRailwayHealthChecker(logger)

	// Register an unhealthy module
	checker.RegisterModuleHealthCheck("test-module", true, func() error {
		return assert.AnError
	})

	// Perform health checks
	checker.performHealthChecks()

	handler := NewRailwayHealthHandler(checker, logger)

	// Create test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Handle health request
	handler.HandleHealth(w, req)

	// Verify response
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "unhealthy", response["status"])
	assert.False(t, response["ready"].(bool))
	assert.False(t, response["live"].(bool))
}

func TestRailwayHealthHandler_HandleReadiness(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	checker := NewRailwayHealthChecker(logger)

	// Register a healthy module
	checker.RegisterModuleHealthCheck("test-module", true, func() error {
		return nil
	})

	// Perform health checks
	checker.performHealthChecks()

	handler := NewRailwayHealthHandler(checker, logger)

	// Create test request
	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	// Handle readiness request
	handler.HandleReadiness(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["ready"].(bool))
	assert.Equal(t, "healthy", response["status"])
}

func TestRailwayHealthHandler_HandleReadiness_NotReady(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	checker := NewRailwayHealthChecker(logger)

	// Register an unhealthy module
	checker.RegisterModuleHealthCheck("test-module", true, func() error {
		return assert.AnError
	})

	// Perform health checks
	checker.performHealthChecks()

	handler := NewRailwayHealthHandler(checker, logger)

	// Create test request
	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	// Handle readiness request
	handler.HandleReadiness(w, req)

	// Verify response
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["ready"].(bool))
	assert.Equal(t, "unhealthy", response["status"])
}

func TestRailwayHealthHandler_HandleLiveness(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	checker := NewRailwayHealthChecker(logger)

	// Register a healthy module
	checker.RegisterModuleHealthCheck("test-module", true, func() error {
		return nil
	})

	// Perform health checks
	checker.performHealthChecks()

	handler := NewRailwayHealthHandler(checker, logger)

	// Create test request
	req := httptest.NewRequest("GET", "/live", nil)
	w := httptest.NewRecorder()

	// Handle liveness request
	handler.HandleLiveness(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["live"].(bool))
	assert.Equal(t, "healthy", response["status"])
}

func TestRailwayHealthHandler_HandleLiveness_NotLive(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	checker := NewRailwayHealthChecker(logger)

	// Register an unhealthy module
	checker.RegisterModuleHealthCheck("test-module", true, func() error {
		return assert.AnError
	})

	// Perform health checks
	checker.performHealthChecks()

	handler := NewRailwayHealthHandler(checker, logger)

	// Create test request
	req := httptest.NewRequest("GET", "/live", nil)
	w := httptest.NewRecorder()

	// Handle liveness request
	handler.HandleLiveness(w, req)

	// Verify response
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["live"].(bool))
	assert.Equal(t, "unhealthy", response["status"])
}

func TestRailwayHealthHandler_HandleModuleHealth(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	checker := NewRailwayHealthChecker(logger)

	// Register a module health check
	checker.RegisterModuleHealthCheck("test-module", true, func() error {
		return nil
	})

	// Perform health checks
	checker.performHealthChecks()

	handler := NewRailwayHealthHandler(checker, logger)

	// Create test request
	req := httptest.NewRequest("GET", "/module-health?module=test-module", nil)
	w := httptest.NewRecorder()

	// Handle module health request
	handler.HandleModuleHealth(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response ModuleHealthCheck
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test-module", response.Name)
	assert.True(t, response.Enabled)
	assert.Equal(t, "healthy", response.Status)
}

func TestRailwayHealthHandler_HandleModuleHealth_NotFound(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	checker := NewRailwayHealthChecker(logger)

	handler := NewRailwayHealthHandler(checker, logger)

	// Create test request for non-existent module
	req := httptest.NewRequest("GET", "/module-health?module=non-existent", nil)
	w := httptest.NewRecorder()

	// Handle module health request
	handler.HandleModuleHealth(w, req)

	// Verify response
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRailwayHealthHandler_HandleModuleHealth_MissingParameter(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	checker := NewRailwayHealthChecker(logger)

	handler := NewRailwayHealthHandler(checker, logger)

	// Create test request without module parameter
	req := httptest.NewRequest("GET", "/module-health", nil)
	w := httptest.NewRecorder()

	// Handle module health request
	handler.HandleModuleHealth(w, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRailwayHealthHandler_HandleForceHealthCheck(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	checker := NewRailwayHealthChecker(logger)

	handler := NewRailwayHealthHandler(checker, logger)

	// Create test request
	req := httptest.NewRequest("POST", "/force-health-check", nil)
	w := httptest.NewRecorder()

	// Handle force health check request
	handler.HandleForceHealthCheck(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Health check forced", response["message"])
	assert.NotNil(t, response["timestamp"])
}

// Test health check functions (placeholder implementations)
func TestCheckDatabaseHealth(t *testing.T) {
	err := CheckDatabaseHealth()
	assert.NoError(t, err)
}

func TestCheckCacheHealth(t *testing.T) {
	err := CheckCacheHealth()
	assert.NoError(t, err)
}

func TestCheckExternalAPIHealth(t *testing.T) {
	err := CheckExternalAPIHealth()
	assert.NoError(t, err)
}

func TestCheckModuleHealth(t *testing.T) {
	err := CheckModuleHealth()
	assert.NoError(t, err)
}

func TestCheckObservabilityHealth(t *testing.T) {
	err := CheckObservabilityHealth()
	assert.NoError(t, err)
}

func TestCheckErrorResilienceHealth(t *testing.T) {
	err := CheckErrorResilienceHealth()
	assert.NoError(t, err)
}

// Test concurrent health checks
func TestRailwayHealthChecker_ConcurrentHealthChecks(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	checker := NewRailwayHealthChecker(logger)

	// Register multiple module health checks
	for i := 0; i < 10; i++ {
		moduleName := fmt.Sprintf("module-%d", i)
		checker.RegisterModuleHealthCheck(moduleName, true, func() error {
			return nil
		})
	}

	// Perform health checks concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			checker.performHealthChecks()
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify no panics occurred
	status := checker.GetHealthStatus()
	assert.NotNil(t, status)
	assert.Equal(t, "healthy", status.Status)
	assert.Equal(t, 10, status.OverallMetrics.TotalModules)
	assert.Equal(t, 10, status.OverallMetrics.HealthyModules)
}

// Test health check metrics
func TestRailwayHealthChecker_HealthCheckMetrics(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	checker := NewRailwayHealthChecker(logger)

	// Register modules with different health states
	checker.RegisterModuleHealthCheck("healthy-module", true, func() error {
		return nil
	})

	checker.RegisterModuleHealthCheck("unhealthy-module", true, func() error {
		return assert.AnError
	})

	checker.RegisterModuleHealthCheck("disabled-module", false, func() error {
		return assert.AnError
	})

	// Perform health checks
	checker.performHealthChecks()

	// Get health status
	status := checker.GetHealthStatus()

	// Verify metrics
	assert.Equal(t, 3, status.OverallMetrics.TotalModules)
	assert.Equal(t, 1, status.OverallMetrics.HealthyModules)
	assert.Equal(t, 1, status.OverallMetrics.UnhealthyModules)
	assert.Equal(t, 0, status.OverallMetrics.DegradedModules)
	assert.Greater(t, status.OverallMetrics.AverageResponseTime, time.Duration(0))
	assert.NotZero(t, status.OverallMetrics.LastCheckTime)
}

// Test health check response times
func TestRailwayHealthChecker_HealthCheckResponseTimes(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	checker := NewRailwayHealthChecker(logger)

	// Register a module with a slow health check
	checker.RegisterModuleHealthCheck("slow-module", true, func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	// Perform health checks
	start := time.Now()
	checker.performHealthChecks()
	duration := time.Since(start)

	// Verify that the health check took at least 100ms
	assert.GreaterOrEqual(t, duration, 100*time.Millisecond)

	// Get health status
	status := checker.GetHealthStatus()
	check := status.Modules["slow-module"]

	// Verify response time is recorded
	assert.GreaterOrEqual(t, check.ResponseTime, 100*time.Millisecond)
}
