package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsHandler_GetMetricsSummary(t *testing.T) {
	// Create test dependencies
	logger := observability.NewLogger("test", "info")
	metricsAggregator := observability.NewMetricsAggregator(logger, nil)

	// Start the metrics aggregator
	metricsAggregator.Start()
	defer metricsAggregator.Stop()

	// Create handler
	handler := NewMetricsHandler(metricsAggregator, logger)

	// Create test request
	req := httptest.NewRequest("GET", "/api/v3/metrics/summary", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.GetMetricsSummary(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Assert response structure
	assert.Contains(t, response, "metrics")
	assert.Contains(t, response, "timestamp")
	assert.Contains(t, response, "endpoint")
	assert.Contains(t, response, "duration_ms")
	assert.Equal(t, "/api/v3/metrics/summary", response["endpoint"])
}

func TestMetricsHandler_GetAggregatedMetrics(t *testing.T) {
	// Create test dependencies
	logger := observability.NewLogger("test", "info")
	metricsAggregator := observability.NewMetricsAggregator(logger, nil)

	// Start the metrics aggregator
	metricsAggregator.Start()
	defer metricsAggregator.Stop()

	// Create handler
	handler := NewMetricsHandler(metricsAggregator, logger)

	// Create test request
	req := httptest.NewRequest("GET", "/api/v3/metrics/aggregated", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.GetAggregatedMetrics(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Assert response structure
	assert.Contains(t, response, "metrics")
	assert.Contains(t, response, "timestamp")
	assert.Contains(t, response, "endpoint")
	assert.Contains(t, response, "duration_ms")
	assert.Equal(t, "/api/v3/metrics/aggregated", response["endpoint"])
}

func TestMetricsHandler_GetModuleMetrics(t *testing.T) {
	// Create test dependencies
	logger := observability.NewLogger("test", "info")
	metricsAggregator := observability.NewMetricsAggregator(logger, nil)

	// Start the metrics aggregator
	metricsAggregator.Start()
	defer metricsAggregator.Stop()

	// Create handler
	handler := NewMetricsHandler(metricsAggregator, logger)

	t.Run("with valid module_id", func(t *testing.T) {
		// Create test request with module_id
		req := httptest.NewRequest("GET", "/api/v3/metrics/module?module_id=test-module", nil)
		w := httptest.NewRecorder()

		// Call handler
		handler.GetModuleMetrics(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		// Parse response
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Assert response structure
		assert.Contains(t, response, "module_id")
		assert.Contains(t, response, "metrics")
		assert.Contains(t, response, "timestamp")
		assert.Contains(t, response, "endpoint")
		assert.Equal(t, "test-module", response["module_id"])
		assert.Equal(t, "/api/v3/metrics/module", response["endpoint"])
	})

	t.Run("without module_id", func(t *testing.T) {
		// Create test request without module_id
		req := httptest.NewRequest("GET", "/api/v3/metrics/module", nil)
		w := httptest.NewRecorder()

		// Call handler
		handler.GetModuleMetrics(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "module_id parameter is required")
	})
}

func TestMetricsHandler_GetModuleList(t *testing.T) {
	// Create test dependencies
	logger := observability.NewLogger("test", "info")
	metricsAggregator := observability.NewMetricsAggregator(logger, nil)

	// Start the metrics aggregator
	metricsAggregator.Start()
	defer metricsAggregator.Stop()

	// Create handler
	handler := NewMetricsHandler(metricsAggregator, logger)

	// Create test request
	req := httptest.NewRequest("GET", "/api/v3/metrics/modules", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.GetModuleList(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Assert response structure
	assert.Contains(t, response, "modules")
	assert.Contains(t, response, "count")
	assert.Contains(t, response, "timestamp")
	assert.Contains(t, response, "endpoint")
	assert.Contains(t, response, "duration_ms")
	assert.Equal(t, "/api/v3/metrics/modules", response["endpoint"])

	// Assert count is a number
	count, ok := response["count"].(float64)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, count, float64(0))
}

func TestMetricsHandler_GetHealthMetrics(t *testing.T) {
	// Create test dependencies
	logger := observability.NewLogger("test", "info")
	metricsAggregator := observability.NewMetricsAggregator(logger, nil)

	// Start the metrics aggregator
	metricsAggregator.Start()
	defer metricsAggregator.Stop()

	// Create handler
	handler := NewMetricsHandler(metricsAggregator, logger)

	// Create test request
	req := httptest.NewRequest("GET", "/api/v3/metrics/health", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.GetHealthMetrics(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Assert response structure
	assert.Contains(t, response, "health")
	assert.Contains(t, response, "timestamp")
	assert.Contains(t, response, "endpoint")
	assert.Contains(t, response, "duration_ms")
	assert.Equal(t, "/api/v3/metrics/health", response["endpoint"])

	// Assert health metrics structure
	health, ok := response["health"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, health, "overall_health")
	assert.Contains(t, health, "health_score")
	assert.Contains(t, health, "degraded_modules")
	assert.Contains(t, health, "critical_modules")
	assert.Contains(t, health, "overall_success_rate")
	assert.Contains(t, health, "overall_error_rate")
	assert.Contains(t, health, "total_requests")
	assert.Contains(t, health, "successful_requests")
	assert.Contains(t, health, "failed_requests")
	assert.Contains(t, health, "timestamp")
}

func TestMetricsHandler_GetPerformanceMetrics(t *testing.T) {
	// Create test dependencies
	logger := observability.NewLogger("test", "info")
	metricsAggregator := observability.NewMetricsAggregator(logger, nil)

	// Start the metrics aggregator
	metricsAggregator.Start()
	defer metricsAggregator.Stop()

	// Create handler
	handler := NewMetricsHandler(metricsAggregator, logger)

	// Create test request
	req := httptest.NewRequest("GET", "/api/v3/metrics/performance", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.GetPerformanceMetrics(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Assert response structure
	assert.Contains(t, response, "performance")
	assert.Contains(t, response, "timestamp")
	assert.Contains(t, response, "endpoint")
	assert.Contains(t, response, "duration_ms")
	assert.Equal(t, "/api/v3/metrics/performance", response["endpoint"])

	// Assert performance metrics structure
	performance, ok := response["performance"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, performance, "average_response_time")
	assert.Contains(t, performance, "p95_response_time")
	assert.Contains(t, performance, "p99_response_time")
	assert.Contains(t, performance, "overall_throughput")
	assert.Contains(t, performance, "total_memory_usage")
	assert.Contains(t, performance, "average_cpu_usage")
	assert.Contains(t, performance, "total_goroutines")
	assert.Contains(t, performance, "database_connections")
	assert.Contains(t, performance, "timestamp")
}

func TestMetricsHandler_GetPrometheusMetrics(t *testing.T) {
	// Create test dependencies
	logger := observability.NewLogger("test", "info")
	metricsAggregator := observability.NewMetricsAggregator(logger, nil)

	// Start the metrics aggregator
	metricsAggregator.Start()
	defer metricsAggregator.Stop()

	// Create handler
	handler := NewMetricsHandler(metricsAggregator, logger)

	// Create test request
	req := httptest.NewRequest("GET", "/api/v3/metrics/prometheus", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.GetPrometheusMetrics(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))

	// Assert response body contains Prometheus metrics format
	body := w.Body.String()
	assert.Contains(t, body, "# HELP")
	assert.Contains(t, body, "# TYPE")
}

func TestMetricsHandler_GetMetricsHistory(t *testing.T) {
	// Create test dependencies
	logger := observability.NewLogger("test", "info")
	metricsAggregator := observability.NewMetricsAggregator(logger, nil)

	// Start the metrics aggregator
	metricsAggregator.Start()
	defer metricsAggregator.Stop()

	// Create handler
	handler := NewMetricsHandler(metricsAggregator, logger)

	t.Run("with module_id", func(t *testing.T) {
		// Create test request with module_id
		req := httptest.NewRequest("GET", "/api/v3/metrics/history?module_id=test-module&limit=50", nil)
		w := httptest.NewRecorder()

		// Call handler
		handler.GetMetricsHistory(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		// Parse response
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Assert response structure
		assert.Contains(t, response, "module_id")
		assert.Contains(t, response, "limit")
		assert.Contains(t, response, "history")
		assert.Contains(t, response, "timestamp")
		assert.Contains(t, response, "endpoint")
		assert.Equal(t, "test-module", response["module_id"])
		assert.Equal(t, float64(50), response["limit"])
		assert.Equal(t, "/api/v3/metrics/history", response["endpoint"])
	})

	t.Run("without module_id", func(t *testing.T) {
		// Create test request without module_id
		req := httptest.NewRequest("GET", "/api/v3/metrics/history?limit=25", nil)
		w := httptest.NewRecorder()

		// Call handler
		handler.GetMetricsHistory(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		// Parse response
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Assert response structure
		assert.Contains(t, response, "module_id")
		assert.Contains(t, response, "limit")
		assert.Contains(t, response, "history")
		assert.Equal(t, "", response["module_id"])
		assert.Equal(t, float64(25), response["limit"])
	})

	t.Run("with default limit", func(t *testing.T) {
		// Create test request without limit
		req := httptest.NewRequest("GET", "/api/v3/metrics/history", nil)
		w := httptest.NewRecorder()

		// Call handler
		handler.GetMetricsHistory(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Assert default limit
		assert.Equal(t, float64(100), response["limit"])
	})

	t.Run("with invalid limit", func(t *testing.T) {
		// Create test request with invalid limit
		req := httptest.NewRequest("GET", "/api/v3/metrics/history?limit=invalid", nil)
		w := httptest.NewRecorder()

		// Call handler
		handler.GetMetricsHistory(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Assert default limit is used
		assert.Equal(t, float64(100), response["limit"])
	})
}

func TestMetricsHandler_ErrorHandling(t *testing.T) {
	// Create test dependencies
	logger := observability.NewLogger("test", "info")
	metricsAggregator := observability.NewMetricsAggregator(logger, nil)

	// Start the metrics aggregator
	metricsAggregator.Start()
	defer metricsAggregator.Stop()

	// Create handler
	handler := NewMetricsHandler(metricsAggregator, logger)

	t.Run("module metrics without module_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v3/metrics/module", nil)
		w := httptest.NewRecorder()

		handler.GetModuleMetrics(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "module_id parameter is required")
	})

	t.Run("module metrics with non-existent module", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v3/metrics/module?module_id=non-existent", nil)
		w := httptest.NewRecorder()

		handler.GetModuleMetrics(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Module not found")
	})
}
