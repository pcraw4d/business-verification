package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/health"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewComprehensiveMetricsHandler(t *testing.T) {
	tests := []struct {
		name          string
		logger        *zap.Logger
		healthChecker *health.RailwayHealthChecker
		version       string
		environment   string
		expectPanic   bool
	}{
		{
			name:          "valid handler creation",
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
			expectPanic:   true,
		},
		{
			name:          "nil health checker",
			logger:        zap.NewNop(),
			healthChecker: nil,
			version:       "1.0.0",
			environment:   "test",
			expectPanic:   true,
		},
		{
			name:          "empty version",
			logger:        zap.NewNop(),
			healthChecker: health.NewRailwayHealthChecker(nil),
			version:       "",
			environment:   "test",
			expectPanic:   true,
		},
		{
			name:          "empty environment",
			logger:        zap.NewNop(),
			healthChecker: health.NewRailwayHealthChecker(nil),
			version:       "1.0.0",
			environment:   "",
			expectPanic:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.Panics(t, func() {
					NewComprehensiveMetricsHandler(tt.logger, tt.healthChecker, tt.version, tt.environment)
				})
			} else {
				handler := NewComprehensiveMetricsHandler(tt.logger, tt.healthChecker, tt.version, tt.environment)

				assert.NotNil(t, handler)
				assert.Equal(t, tt.version, handler.version)
				assert.Equal(t, tt.environment, handler.environment)
				assert.NotNil(t, handler.metrics)
				assert.NotNil(t, handler.collectors)
				assert.Equal(t, 30*time.Second, handler.updateInterval)
			}
		})
	}
}

func TestComprehensiveMetricsHandler_HandleComprehensiveMetrics(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewComprehensiveMetricsHandler(logger, healthChecker, "1.0.0", "test")

	tests := []struct {
		name           string
		requestMethod  string
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:           "GET request",
			requestMethod:  "GET",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.NotNil(t, response["timestamp"])
				assert.Equal(t, "1.0.0", response["version"])
				assert.Equal(t, "test", response["environment"])
				assert.NotNil(t, response["system_metrics"])
				assert.NotNil(t, response["api_metrics"])
				assert.NotNil(t, response["business_metrics"])
				assert.NotNil(t, response["performance_metrics"])
				assert.NotNil(t, response["resource_metrics"])
				assert.NotNil(t, response["error_metrics"])
			},
		},
		{
			name:           "POST request",
			requestMethod:  "POST",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.NotNil(t, response["timestamp"])
				assert.Equal(t, "1.0.0", response["version"])
				assert.Equal(t, "test", response["environment"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.requestMethod, "/metrics/comprehensive", nil)
			w := httptest.NewRecorder()

			handler.HandleComprehensiveMetrics(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			tt.checkResponse(t, response)
		})
	}
}

func TestComprehensiveMetricsHandler_HandlePrometheusMetrics(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewComprehensiveMetricsHandler(logger, healthChecker, "1.0.0", "test")

	tests := []struct {
		name           string
		requestMethod  string
		expectedStatus int
		checkResponse  func(*testing.T, string)
	}{
		{
			name:           "GET request",
			requestMethod:  "GET",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response string) {
				assert.Contains(t, response, "# HELP go_goroutines")
				assert.Contains(t, response, "# TYPE go_goroutines gauge")
				assert.Contains(t, response, "go_goroutines{version=\"1.0.0\",environment=\"test\"}")
				assert.Contains(t, response, "# HELP api_requests_total")
				assert.Contains(t, response, "# TYPE api_requests_total counter")
				assert.Contains(t, response, "api_requests_total{version=\"1.0.0\",environment=\"test\"}")
				assert.Contains(t, response, "# HELP business_verifications_total")
				assert.Contains(t, response, "# TYPE business_verifications_total counter")
				assert.Contains(t, response, "business_verifications_total{version=\"1.0.0\",environment=\"test\"}")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.requestMethod, "/metrics/prometheus", nil)
			w := httptest.NewRecorder()

			handler.HandlePrometheusMetrics(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8", w.Header().Get("Content-Type"))

			tt.checkResponse(t, w.Body.String())
		})
	}
}

func TestComprehensiveMetricsHandler_HandleSystemMetrics(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewComprehensiveMetricsHandler(logger, healthChecker, "1.0.0", "test")

	tests := []struct {
		name           string
		requestMethod  string
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:           "GET request",
			requestMethod:  "GET",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.NotNil(t, response["timestamp"])
				assert.Equal(t, "1.0.0", response["version"])
				assert.Equal(t, "test", response["environment"])

				systemMetrics, ok := response["system_metrics"].(map[string]interface{})
				assert.True(t, ok)
				assert.NotNil(t, systemMetrics["uptime"])
				assert.NotNil(t, systemMetrics["start_time"])
				assert.NotNil(t, systemMetrics["go_version"])
				assert.NotNil(t, systemMetrics["num_cpu"])
				assert.NotNil(t, systemMetrics["num_goroutines"])
				assert.NotNil(t, systemMetrics["memory_stats"])
				assert.NotNil(t, systemMetrics["gc_stats"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.requestMethod, "/metrics/system", nil)
			w := httptest.NewRecorder()

			handler.HandleSystemMetrics(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			tt.checkResponse(t, response)
		})
	}
}

func TestComprehensiveMetricsHandler_HandleAPIMetrics(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewComprehensiveMetricsHandler(logger, healthChecker, "1.0.0", "test")

	tests := []struct {
		name           string
		requestMethod  string
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:           "GET request",
			requestMethod:  "GET",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.NotNil(t, response["timestamp"])
				assert.Equal(t, "1.0.0", response["version"])
				assert.Equal(t, "test", response["environment"])

				apiMetrics, ok := response["api_metrics"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, float64(1000), apiMetrics["total_requests"])
				assert.Equal(t, 50.5, apiMetrics["requests_per_second"])
				assert.NotNil(t, apiMetrics["average_response_time"])
				assert.NotNil(t, apiMetrics["requests_by_method"])
				assert.NotNil(t, apiMetrics["requests_by_endpoint"])
				assert.NotNil(t, apiMetrics["requests_by_status"])
				assert.NotNil(t, apiMetrics["errors_by_type"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.requestMethod, "/metrics/api", nil)
			w := httptest.NewRecorder()

			handler.HandleAPIMetrics(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			tt.checkResponse(t, response)
		})
	}
}

func TestComprehensiveMetricsHandler_HandleBusinessMetrics(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewComprehensiveMetricsHandler(logger, healthChecker, "1.0.0", "test")

	tests := []struct {
		name           string
		requestMethod  string
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:           "GET request",
			requestMethod:  "GET",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.NotNil(t, response["timestamp"])
				assert.Equal(t, "1.0.0", response["version"])
				assert.Equal(t, "test", response["environment"])

				businessMetrics, ok := response["business_metrics"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, float64(5000), businessMetrics["total_verifications"])
				assert.Equal(t, 25.3, businessMetrics["verifications_per_second"])
				assert.Equal(t, 96.0, businessMetrics["success_rate"])
				assert.Equal(t, 4.0, businessMetrics["error_rate"])
				assert.Equal(t, 0.85, businessMetrics["average_confidence_score"])
				assert.NotNil(t, businessMetrics["verifications_by_status"])
				assert.NotNil(t, businessMetrics["verifications_by_type"])
				assert.NotNil(t, businessMetrics["verifications_by_industry"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.requestMethod, "/metrics/business", nil)
			w := httptest.NewRecorder()

			handler.HandleBusinessMetrics(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			tt.checkResponse(t, response)
		})
	}
}

func TestComprehensiveMetricsHandler_RegisterCollector(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewComprehensiveMetricsHandler(logger, healthChecker, "1.0.0", "test")

	// Create a test collector
	testCollector := &testMetricsCollector{
		name: "test-collector",
		metrics: map[string]interface{}{
			"custom_metric_1": 42,
			"custom_metric_2": "test-value",
		},
	}

	// Register the collector
	handler.RegisterCollector(testCollector)

	// Verify the collector was registered
	assert.Contains(t, handler.collectors, "test-collector")
	assert.Equal(t, testCollector, handler.collectors["test-collector"])
}

func TestComprehensiveMetricsHandler_CollectMetrics(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewComprehensiveMetricsHandler(logger, healthChecker, "1.0.0", "test")

	// Test metrics collection
	ctx := context.Background()
	metrics := handler.collectMetrics(ctx)

	assert.NotNil(t, metrics)
	assert.Equal(t, "1.0.0", metrics.Version)
	assert.Equal(t, "test", metrics.Environment)
	assert.NotNil(t, metrics.SystemMetrics)
	assert.NotNil(t, metrics.APIMetrics)
	assert.NotNil(t, metrics.BusinessMetrics)
	assert.NotNil(t, metrics.PerformanceMetrics)
	assert.NotNil(t, metrics.ResourceMetrics)
	assert.NotNil(t, metrics.ErrorMetrics)
	assert.NotNil(t, metrics.CustomMetrics)
}

func TestComprehensiveMetricsHandler_CollectSystemMetrics(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewComprehensiveMetricsHandler(logger, healthChecker, "1.0.0", "test")

	// Test system metrics collection
	systemMetrics := handler.collectSystemMetrics()

	assert.NotNil(t, systemMetrics)
	assert.Equal(t, "1.0.0", handler.version)
	assert.Equal(t, "test", handler.environment)
	assert.NotZero(t, systemMetrics.NumCPU)
	assert.NotZero(t, systemMetrics.NumGoroutines)
	assert.NotEmpty(t, systemMetrics.GoVersion)
	assert.NotNil(t, systemMetrics.MemoryStats)
	assert.NotNil(t, systemMetrics.GCStats)
}

func TestComprehensiveMetricsHandler_CollectAPIMetrics(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewComprehensiveMetricsHandler(logger, healthChecker, "1.0.0", "test")

	// Test API metrics collection
	apiMetrics := handler.collectAPIMetrics()

	assert.NotNil(t, apiMetrics)
	assert.Equal(t, int64(1000), apiMetrics.TotalRequests)
	assert.Equal(t, 50.5, apiMetrics.RequestsPerSecond)
	assert.NotNil(t, apiMetrics.RequestsByMethod)
	assert.NotNil(t, apiMetrics.RequestsByEndpoint)
	assert.NotNil(t, apiMetrics.RequestsByStatus)
	assert.NotNil(t, apiMetrics.ErrorsByType)
}

func TestComprehensiveMetricsHandler_CollectBusinessMetrics(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewComprehensiveMetricsHandler(logger, healthChecker, "1.0.0", "test")

	// Test business metrics collection
	businessMetrics := handler.collectBusinessMetrics()

	assert.NotNil(t, businessMetrics)
	assert.Equal(t, int64(5000), businessMetrics.TotalVerifications)
	assert.Equal(t, 25.3, businessMetrics.VerificationsPerSecond)
	assert.Equal(t, 96.0, businessMetrics.SuccessRate)
	assert.Equal(t, 4.0, businessMetrics.ErrorRate)
	assert.Equal(t, 0.85, businessMetrics.AverageConfidenceScore)
	assert.NotNil(t, businessMetrics.VerificationsByStatus)
	assert.NotNil(t, businessMetrics.VerificationsByType)
	assert.NotNil(t, businessMetrics.VerificationsByIndustry)
}

func TestComprehensiveMetricsHandler_CollectPerformanceMetrics(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewComprehensiveMetricsHandler(logger, healthChecker, "1.0.0", "test")

	// Test performance metrics collection
	performanceMetrics := handler.collectPerformanceMetrics()

	assert.NotNil(t, performanceMetrics)
	assert.Equal(t, 45.2, performanceMetrics.CPUUsagePercent)
	assert.Equal(t, 62.8, performanceMetrics.MemoryUsagePercent)
	assert.Equal(t, 35.5, performanceMetrics.DiskUsagePercent)
	assert.NotNil(t, performanceMetrics.NetworkLatency)
	assert.Equal(t, 850.5, performanceMetrics.NetworkThroughput)
}

func TestComprehensiveMetricsHandler_CollectResourceMetrics(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewComprehensiveMetricsHandler(logger, healthChecker, "1.0.0", "test")

	// Test resource metrics collection
	resourceMetrics := handler.collectResourceMetrics()

	assert.NotNil(t, resourceMetrics)
	assert.Equal(t, 125, resourceMetrics.OpenFiles)
	assert.Equal(t, 1024, resourceMetrics.MaxFiles)
	assert.Equal(t, 125, resourceMetrics.FileDescriptors)
	assert.Equal(t, 45, resourceMetrics.Threads)
	assert.Equal(t, 1000, resourceMetrics.MaxThreads)
	assert.NotNil(t, resourceMetrics.LoadAverage)
	assert.NotZero(t, resourceMetrics.DiskIOReadBytes)
	assert.NotZero(t, resourceMetrics.DiskIOWriteBytes)
}

func TestComprehensiveMetricsHandler_CollectErrorMetrics(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewComprehensiveMetricsHandler(logger, healthChecker, "1.0.0", "test")

	// Test error metrics collection
	errorMetrics := handler.collectErrorMetrics()

	assert.NotNil(t, errorMetrics)
	assert.Equal(t, int64(50), errorMetrics.TotalErrors)
	assert.Equal(t, 2.5, errorMetrics.ErrorsPerSecond)
	assert.Equal(t, 5.0, errorMetrics.ErrorRate)
	assert.NotNil(t, errorMetrics.ErrorsByType)
	assert.NotNil(t, errorMetrics.ErrorsByEndpoint)
	assert.NotNil(t, errorMetrics.ErrorsByStatus)
	assert.NotNil(t, errorMetrics.LastErrorTime)
	assert.NotEmpty(t, errorMetrics.LastErrorMessage)
}

func TestComprehensiveMetricsHandler_ConvertToPrometheusFormat(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewComprehensiveMetricsHandler(logger, healthChecker, "1.0.0", "test")

	// Collect metrics
	ctx := context.Background()
	metrics := handler.collectMetrics(ctx)

	// Convert to Prometheus format
	prometheusMetrics := handler.convertToPrometheusFormat(metrics)

	assert.NotEmpty(t, prometheusMetrics)
	assert.Contains(t, prometheusMetrics, "# HELP go_goroutines")
	assert.Contains(t, prometheusMetrics, "# TYPE go_goroutines gauge")
	assert.Contains(t, prometheusMetrics, "go_goroutines{version=\"1.0.0\",environment=\"test\"}")
	assert.Contains(t, prometheusMetrics, "# HELP api_requests_total")
	assert.Contains(t, prometheusMetrics, "# TYPE api_requests_total counter")
	assert.Contains(t, prometheusMetrics, "api_requests_total{version=\"1.0.0\",environment=\"test\"}")
	assert.Contains(t, prometheusMetrics, "# HELP business_verifications_total")
	assert.Contains(t, prometheusMetrics, "# TYPE business_verifications_total counter")
	assert.Contains(t, prometheusMetrics, "business_verifications_total{version=\"1.0.0\",environment=\"test\"}")
	assert.Contains(t, prometheusMetrics, "# HELP performance_cpu_usage_percent")
	assert.Contains(t, prometheusMetrics, "# TYPE performance_cpu_usage_percent gauge")
	assert.Contains(t, prometheusMetrics, "performance_cpu_usage_percent{version=\"1.0.0\",environment=\"test\"}")
	assert.Contains(t, prometheusMetrics, "# HELP error_total")
	assert.Contains(t, prometheusMetrics, "# TYPE error_total counter")
	assert.Contains(t, prometheusMetrics, "error_total{version=\"1.0.0\",environment=\"test\"}")
}

func TestComprehensiveMetricsHandler_Caching(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewComprehensiveMetricsHandler(logger, healthChecker, "1.0.0", "test")

	t.Run("cache functionality", func(t *testing.T) {
		// First request should populate cache
		req1 := httptest.NewRequest("GET", "/metrics/comprehensive", nil)
		w1 := httptest.NewRecorder()
		handler.HandleComprehensiveMetrics(w1, req1)

		assert.Equal(t, http.StatusOK, w1.Code)

		// Second request should use cache
		req2 := httptest.NewRequest("GET", "/metrics/comprehensive", nil)
		w2 := httptest.NewRecorder()
		handler.HandleComprehensiveMetrics(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)

		// Both responses should be identical
		assert.Equal(t, w1.Body.String(), w2.Body.String())
	})
}

func TestComprehensiveMetricsHandler_Logging(t *testing.T) {
	core, obs := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	healthChecker := health.NewRailwayHealthChecker(nil)

	handler := NewComprehensiveMetricsHandler(logger, healthChecker, "1.0.0", "test")

	t.Run("comprehensive metrics logging", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/metrics/comprehensive", nil)
		w := httptest.NewRecorder()

		handler.HandleComprehensiveMetrics(w, req)

		// Check that debug logs were generated
		logs := obs.FilterMessage("Comprehensive metrics request served").All()
		assert.NotEmpty(t, logs)

		if len(logs) > 0 {
			log := logs[0]
			assert.Equal(t, "success", log.ContextMap()["status"])
			assert.NotNil(t, log.ContextMap()["client_ip"])
			assert.NotNil(t, log.ContextMap()["response_time"])
			assert.NotNil(t, log.ContextMap()["user_agent"])
		}
	})

	t.Run("prometheus metrics logging", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/metrics/prometheus", nil)
		w := httptest.NewRecorder()

		handler.HandlePrometheusMetrics(w, req)

		// Check that debug logs were generated
		logs := obs.FilterMessage("Prometheus metrics served").All()
		assert.NotEmpty(t, logs)

		if len(logs) > 0 {
			log := logs[0]
			assert.NotNil(t, log.ContextMap()["client_ip"])
			assert.NotNil(t, log.ContextMap()["response_time"])
		}
	})

	t.Run("system metrics logging", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/metrics/system", nil)
		w := httptest.NewRecorder()

		handler.HandleSystemMetrics(w, req)

		// Check that debug logs were generated
		logs := obs.FilterMessage("System metrics served").All()
		assert.NotEmpty(t, logs)

		if len(logs) > 0 {
			log := logs[0]
			assert.NotNil(t, log.ContextMap()["client_ip"])
			assert.NotNil(t, log.ContextMap()["response_time"])
		}
	})

	t.Run("api metrics logging", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/metrics/api", nil)
		w := httptest.NewRecorder()

		handler.HandleAPIMetrics(w, req)

		// Check that debug logs were generated
		logs := obs.FilterMessage("API metrics served").All()
		assert.NotEmpty(t, logs)

		if len(logs) > 0 {
			log := logs[0]
			assert.NotNil(t, log.ContextMap()["client_ip"])
			assert.NotNil(t, log.ContextMap()["response_time"])
		}
	})

	t.Run("business metrics logging", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/metrics/business", nil)
		w := httptest.NewRecorder()

		handler.HandleBusinessMetrics(w, req)

		// Check that debug logs were generated
		logs := obs.FilterMessage("Business metrics served").All()
		assert.NotEmpty(t, logs)

		if len(logs) > 0 {
			log := logs[0]
			assert.NotNil(t, log.ContextMap()["client_ip"])
			assert.NotNil(t, log.ContextMap()["response_time"])
		}
	})
}

// testMetricsCollector is a test implementation of MetricsCollector
type testMetricsCollector struct {
	name    string
	metrics map[string]interface{}
}

func (t *testMetricsCollector) Collect(ctx context.Context) (map[string]interface{}, error) {
	return t.metrics, nil
}

func (t *testMetricsCollector) Name() string {
	return t.name
}
