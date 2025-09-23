package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kyb-platform/internal/health"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewHealthHandler(t *testing.T) {
	tests := []struct {
		name          string
		logger        *zap.Logger
		healthChecker *health.RailwayHealthChecker
		version       string
		environment   string
		expectPanic   bool
	}{
		{
			name:          "valid configuration",
			logger:        zap.NewNop(),
			healthChecker: health.NewRailwayHealthChecker(nil),
			version:       "1.0.0",
			environment:   "test",
			expectPanic:   false,
		},
		{
			name:          "nil logger",
			logger:        nil,
			healthChecker: health.NewRailwayHealthChecker(nil),
			version:       "1.0.0",
			environment:   "test",
			expectPanic:   false, // Should not panic with nil logger
		},
		{
			name:          "nil health checker",
			logger:        zap.NewNop(),
			healthChecker: nil,
			version:       "1.0.0",
			environment:   "test",
			expectPanic:   false, // Should not panic with nil health checker
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHealthHandler(tt.logger, tt.healthChecker, tt.version, tt.environment)

			assert.NotNil(t, handler)
			assert.Equal(t, tt.version, handler.version)
			assert.Equal(t, tt.environment, handler.environment)
			assert.Equal(t, 30*time.Second, handler.cacheTTL)
			assert.NotNil(t, handler.cache)
		})
	}
}

func TestHealthHandler_HandleHealth(t *testing.T) {
	core, _ := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewHealthHandler(logger, healthChecker, "1.0.0", "test")

	tests := []struct {
		name           string
		requestMethod  string
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:           "successful health check",
			requestMethod:  "GET",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "healthy", response["status"])
				assert.Equal(t, "1.0.0", response["version"])
				assert.Equal(t, "test", response["environment"])
				assert.True(t, response["ready"].(bool))
				assert.True(t, response["live"].(bool))

				// Check that checks are present
				checks, ok := response["checks"].(map[string]interface{})
				assert.True(t, ok)
				assert.NotEmpty(t, checks)

				// Check that metrics are present
				metrics, ok := response["metrics"].(map[string]interface{})
				assert.True(t, ok)
				assert.NotNil(t, metrics)
			},
		},
		{
			name:           "cached response",
			requestMethod:  "GET",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "healthy", response["status"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.requestMethod, "/health", nil)
			w := httptest.NewRecorder()

			handler.HandleHealth(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			tt.checkResponse(t, response)
		})
	}
}

func TestHealthHandler_HandleReadiness(t *testing.T) {
	core, _ := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewHealthHandler(logger, healthChecker, "1.0.0", "test")

	tests := []struct {
		name           string
		requestMethod  string
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:           "successful readiness check",
			requestMethod:  "GET",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.True(t, response["ready"].(bool))
				assert.Equal(t, "healthy", response["status"])

				// Check that checks are present
				checks, ok := response["checks"].(map[string]interface{})
				assert.True(t, ok)
				assert.NotEmpty(t, checks)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.requestMethod, "/ready", nil)
			w := httptest.NewRecorder()

			handler.HandleReadiness(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			tt.checkResponse(t, response)
		})
	}
}

func TestHealthHandler_HandleLiveness(t *testing.T) {
	core, _ := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewHealthHandler(logger, healthChecker, "1.0.0", "test")

	tests := []struct {
		name           string
		requestMethod  string
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:           "successful liveness check",
			requestMethod:  "GET",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.True(t, response["live"].(bool))
				assert.Equal(t, "healthy", response["status"])
				assert.NotEmpty(t, response["uptime"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.requestMethod, "/live", nil)
			w := httptest.NewRecorder()

			handler.HandleLiveness(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			tt.checkResponse(t, response)
		})
	}
}

func TestHealthHandler_HandleDetailedHealth(t *testing.T) {
	core, _ := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewHealthHandler(logger, healthChecker, "1.0.0", "test")

	tests := []struct {
		name           string
		requestMethod  string
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:           "successful detailed health check",
			requestMethod:  "GET",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "healthy", response["status"])
				assert.Equal(t, "1.0.0", response["version"])
				assert.Equal(t, "test", response["environment"])

				// Check that detailed checks are present
				checks, ok := response["checks"].(map[string]interface{})
				assert.True(t, ok)
				assert.NotEmpty(t, checks)

				// Should include additional detailed checks
				assert.Contains(t, checks, "memory")
				assert.Contains(t, checks, "goroutines")
				assert.Contains(t, checks, "disk")
				assert.Contains(t, checks, "network")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.requestMethod, "/health/detailed", nil)
			w := httptest.NewRecorder()

			handler.HandleDetailedHealth(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			tt.checkResponse(t, response)
		})
	}
}

func TestHealthHandler_HandleModuleHealth(t *testing.T) {
	core, _ := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	// Register a test module
	healthChecker.RegisterModuleHealthCheck("test-module", true, func() error {
		return nil
	})

	handler := NewHealthHandler(logger, healthChecker, "1.0.0", "test")

	tests := []struct {
		name           string
		moduleName     string
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:           "valid module health check",
			moduleName:     "test-module",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "unknown", response["status"]) // Initial status
				assert.NotNil(t, response["last_check"])
			},
		},
		{
			name:           "missing module parameter",
			moduleName:     "",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				// Should return error message
			},
		},
		{
			name:           "non-existent module",
			moduleName:     "non-existent",
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				// Should return error message
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/health/module"
			if tt.moduleName != "" {
				url += "?module=" + tt.moduleName
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.HandleModuleHealth(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				tt.checkResponse(t, response)
			}
		})
	}
}

func TestHealthHandler_HealthChecks(t *testing.T) {
	core, obs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewHealthHandler(logger, healthChecker, "1.0.0", "test")

	t.Run("checkSystemHealth", func(t *testing.T) {
		check := handler.checkSystemHealth()

		assert.Equal(t, "healthy", check.Status)
		assert.NotZero(t, check.ResponseTime)
		assert.NotNil(t, check.Details)
		assert.Equal(t, "1.0.0", check.Details["version"])
		assert.Equal(t, "test", check.Details["environment"])
	})

	t.Run("checkDatabaseHealth", func(t *testing.T) {
		check := handler.checkDatabaseHealth()

		assert.Equal(t, "healthy", check.Status)
		assert.NotZero(t, check.ResponseTime)
		assert.NotNil(t, check.Details)
		assert.Equal(t, 10, check.Details["connection_pool_size"])
		assert.Equal(t, 3, check.Details["active_connections"])
	})

	t.Run("checkCacheHealth", func(t *testing.T) {
		check := handler.checkCacheHealth()

		assert.Equal(t, "healthy", check.Status)
		assert.NotZero(t, check.ResponseTime)
		assert.NotNil(t, check.Details)
		assert.Equal(t, "1GB", check.Details["cache_size"])
		assert.Equal(t, 0.85, check.Details["hit_rate"])
	})

	t.Run("checkExternalAPIsHealth", func(t *testing.T) {
		check := handler.checkExternalAPIsHealth()

		assert.Equal(t, "healthy", check.Status)
		assert.NotZero(t, check.ResponseTime)
		assert.NotNil(t, check.Details)
		assert.Equal(t, 3, check.Details["apis_checked"])
		assert.Equal(t, 3, check.Details["apis_healthy"])
	})

	t.Run("checkMLModelsHealth", func(t *testing.T) {
		check := handler.checkMLModelsHealth()

		assert.Equal(t, "healthy", check.Status)
		assert.NotZero(t, check.ResponseTime)
		assert.NotNil(t, check.Details)
		assert.Equal(t, 5, check.Details["models_loaded"])
		assert.Equal(t, "1.2.3", check.Details["model_version"])
	})

	t.Run("checkObservabilityHealth", func(t *testing.T) {
		check := handler.checkObservabilityHealth()

		assert.Equal(t, "healthy", check.Status)
		assert.NotZero(t, check.ResponseTime)
		assert.NotNil(t, check.Details)
		assert.Equal(t, true, check.Details["logging_enabled"])
		assert.Equal(t, true, check.Details["metrics_enabled"])
	})

	t.Run("checkMemoryHealth", func(t *testing.T) {
		check := handler.checkMemoryHealth()

		assert.Equal(t, "healthy", check.Status)
		assert.NotZero(t, check.ResponseTime)
		assert.NotNil(t, check.Details)
		assert.NotZero(t, check.Details["alloc_bytes"])
		assert.NotZero(t, check.Details["sys_bytes"])
	})

	t.Run("checkGoroutinesHealth", func(t *testing.T) {
		check := handler.checkGoroutinesHealth()

		assert.Equal(t, "healthy", check.Status)
		assert.NotZero(t, check.ResponseTime)
		assert.NotNil(t, check.Details)
		assert.NotZero(t, check.Details["num_goroutines"])
		assert.NotZero(t, check.Details["num_cpu"])
	})

	t.Run("checkDiskHealth", func(t *testing.T) {
		check := handler.checkDiskHealth()

		assert.Equal(t, "healthy", check.Status)
		assert.NotZero(t, check.ResponseTime)
		assert.NotNil(t, check.Details)
		assert.Equal(t, 45.2, check.Details["disk_usage_percent"])
		assert.Equal(t, 125.8, check.Details["available_space_gb"])
	})

	t.Run("checkNetworkHealth", func(t *testing.T) {
		check := handler.checkNetworkHealth()

		assert.Equal(t, "healthy", check.Status)
		assert.NotZero(t, check.ResponseTime)
		assert.NotNil(t, check.Details)
		assert.Equal(t, 15.3, check.Details["network_latency_ms"])
		assert.Equal(t, 0.01, check.Details["packet_loss_percent"])
	})
}

func TestHealthHandler_CalculateHealthMetrics(t *testing.T) {
	core, obs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewHealthHandler(logger, healthChecker, "1.0.0", "test")

	tests := []struct {
		name              string
		checks            map[string]HealthCheck
		expectedTotal     int
		expectedHealthy   int
		expectedUnhealthy int
		expectedDegraded  int
	}{
		{
			name: "all healthy checks",
			checks: map[string]HealthCheck{
				"check1": {Status: "healthy", ResponseTime: 10 * time.Millisecond},
				"check2": {Status: "healthy", ResponseTime: 20 * time.Millisecond},
				"check3": {Status: "healthy", ResponseTime: 15 * time.Millisecond},
			},
			expectedTotal:     3,
			expectedHealthy:   3,
			expectedUnhealthy: 0,
			expectedDegraded:  0,
		},
		{
			name: "mixed status checks",
			checks: map[string]HealthCheck{
				"check1": {Status: "healthy", ResponseTime: 10 * time.Millisecond},
				"check2": {Status: "unhealthy", ResponseTime: 20 * time.Millisecond},
				"check3": {Status: "degraded", ResponseTime: 15 * time.Millisecond},
			},
			expectedTotal:     3,
			expectedHealthy:   1,
			expectedUnhealthy: 1,
			expectedDegraded:  1,
		},
		{
			name:              "empty checks",
			checks:            map[string]HealthCheck{},
			expectedTotal:     0,
			expectedHealthy:   0,
			expectedUnhealthy: 0,
			expectedDegraded:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := handler.calculateHealthMetrics(tt.checks)

			assert.Equal(t, tt.expectedTotal, metrics.TotalChecks)
			assert.Equal(t, tt.expectedHealthy, metrics.HealthyChecks)
			assert.Equal(t, tt.expectedUnhealthy, metrics.UnhealthyChecks)
			assert.Equal(t, tt.expectedDegraded, metrics.DegradedChecks)

			// Check that memory and runtime info are populated
			assert.NotZero(t, metrics.MemoryUsage.Alloc)
			assert.NotZero(t, metrics.MemoryUsage.Sys)
			assert.NotEmpty(t, metrics.GoRuntime.Version)
			assert.NotZero(t, metrics.GoRuntime.NumCPU)
			assert.NotZero(t, metrics.GoRuntime.NumGoroutine)
		})
	}
}

func TestHealthHandler_PerformHealthChecks(t *testing.T) {
	core, obs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewHealthHandler(logger, healthChecker, "1.0.0", "test")

	t.Run("performHealthChecks", func(t *testing.T) {
		status := handler.performHealthChecks()

		assert.Equal(t, "healthy", status.Status)
		assert.Equal(t, "1.0.0", status.Version)
		assert.Equal(t, "test", status.Environment)
		assert.True(t, status.Ready)
		assert.True(t, status.Live)
		assert.NotZero(t, status.Uptime)
		assert.NotEmpty(t, status.Checks)
		assert.NotNil(t, status.Metrics)

		// Check that all expected checks are present
		expectedChecks := []string{"system", "database", "cache", "external_apis", "ml_models", "observability"}
		for _, checkName := range expectedChecks {
			assert.Contains(t, status.Checks, checkName)
		}
	})

	t.Run("performReadinessChecks", func(t *testing.T) {
		status := handler.performReadinessChecks()

		assert.Equal(t, "healthy", status.Status)
		assert.True(t, status.Ready)
		assert.True(t, status.Live) // Liveness is separate
		assert.NotEmpty(t, status.Checks)

		// Should only include critical dependencies
		expectedChecks := []string{"database", "cache"}
		for _, checkName := range expectedChecks {
			assert.Contains(t, status.Checks, checkName)
		}
	})

	t.Run("performLivenessChecks", func(t *testing.T) {
		status := handler.performLivenessChecks()

		assert.Equal(t, "healthy", status.Status)
		assert.True(t, status.Ready) // Readiness is separate
		assert.True(t, status.Live)
		assert.NotEmpty(t, status.Checks)

		// Should only include basic system health
		assert.Contains(t, status.Checks, "system")
		assert.Len(t, status.Checks, 1)
	})

	t.Run("performDetailedHealthChecks", func(t *testing.T) {
		status := handler.performDetailedHealthChecks()

		assert.Equal(t, "healthy", status.Status)
		assert.True(t, status.Ready)
		assert.True(t, status.Live)
		assert.NotEmpty(t, status.Checks)

		// Should include additional detailed checks
		expectedDetailedChecks := []string{"memory", "goroutines", "disk", "network"}
		for _, checkName := range expectedDetailedChecks {
			assert.Contains(t, status.Checks, checkName)
		}
	})
}

func TestHealthHandler_CacheFunctionality(t *testing.T) {
	core, obs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewHealthHandler(logger, healthChecker, "1.0.0", "test")

	t.Run("cache functionality", func(t *testing.T) {
		// First request should populate cache
		req1 := httptest.NewRequest("GET", "/health", nil)
		w1 := httptest.NewRecorder()
		handler.HandleHealth(w1, req1)

		assert.Equal(t, http.StatusOK, w1.Code)

		// Second request should use cache
		req2 := httptest.NewRequest("GET", "/health", nil)
		w2 := httptest.NewRecorder()
		handler.HandleHealth(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)

		// Both responses should be identical
		assert.Equal(t, w1.Body.String(), w2.Body.String())
	})

	t.Run("getCachedHealthStatus", func(t *testing.T) {
		// Populate cache
		handler.cache["test-check"] = HealthCheck{
			Status:    "healthy",
			LastCheck: time.Now(),
		}

		status := handler.getCachedHealthStatus()

		assert.Equal(t, "healthy", status.Status)
		assert.Equal(t, "1.0.0", status.Version)
		assert.Equal(t, "test", status.Environment)
		assert.True(t, status.Ready)
		assert.True(t, status.Live)
		assert.Contains(t, status.Checks, "test-check")
	})
}

func TestHealthHandler_Logging(t *testing.T) {
	core, obs := observer.New(zap.DebugLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewHealthHandler(logger, healthChecker, "1.0.0", "test")

	t.Run("health check logging", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		handler.HandleHealth(w, req)

		// Check that debug logs were generated
		logs := obs.FilterMessage("Health check request served").All()
		assert.NotEmpty(t, logs)

		if len(logs) > 0 {
			log := logs[0]
			assert.Equal(t, "healthy", log.ContextMap()["status"])
			assert.Equal(t, 200, log.ContextMap()["http_status"])
		}
	})

	t.Run("readiness probe logging", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()

		handler.HandleReadiness(w, req)

		// Check that debug logs were generated
		logs := obs.FilterMessage("Readiness probe served").All()
		assert.NotEmpty(t, logs)

		if len(logs) > 0 {
			log := logs[0]
			assert.Equal(t, true, log.ContextMap()["ready"])
			assert.Equal(t, 200, log.ContextMap()["http_status"])
		}
	})

	t.Run("liveness probe logging", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/live", nil)
		w := httptest.NewRecorder()

		handler.HandleLiveness(w, req)

		// Check that debug logs were generated
		logs := obs.FilterMessage("Liveness probe served").All()
		assert.NotEmpty(t, logs)

		if len(logs) > 0 {
			log := logs[0]
			assert.Equal(t, true, log.ContextMap()["live"])
			assert.Equal(t, 200, log.ContextMap()["http_status"])
		}
	})
}
