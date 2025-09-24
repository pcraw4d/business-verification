package integration

import (
	"bytes"

	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/api/routes"
	"github.com/pcraw4d/business-verification/internal/modules/classification_monitoring"
	"github.com/pcraw4d/business-verification/internal/modules/classification_optimization"
)

func TestAlgorithmOptimizationIntegration(t *testing.T) {
	logger := zap.NewNop()

	// Create algorithm optimizer
	config := &classification_optimization.OptimizationConfig{
		MinPatternsForOptimization: 2,
		OptimizationWindowHours:    24,
		MaxOptimizationsPerDay:     10,
		ConfidenceThreshold:        0.7,
		PerformanceThreshold:       0.05,
		OptimizationTimeout:        30 * time.Minute,
		EnableAutoOptimization:     true,
	}

	optimizer := classification_optimization.NewAlgorithmOptimizer(config, logger)

	// Create pattern analyzer and set it in optimizer
	patternAnalyzer := classification_monitoring.NewPatternAnalysisEngine(nil, logger)
	optimizer.SetPatternAnalyzer(patternAnalyzer)

	// Create handler
	handler := handlers.NewAlgorithmOptimizationHandler(optimizer, logger)

	// Create router and register routes
	router := mux.NewRouter()
	routes.RegisterAlgorithmOptimizationRoutes(router, handler)

	t.Run("test analyze and optimize endpoint", func(t *testing.T) {
		// Create request body
		requestBody := map[string]interface{}{
			"force_optimization": false,
		}

		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		// Create request
		req := httptest.NewRequest("POST", "/api/v1/algorithm-optimization/analyze", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Serve request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "success", response["status"])
		assert.Contains(t, response["message"], "completed")
	})

	t.Run("test get optimization history endpoint", func(t *testing.T) {
		// Create request
		req := httptest.NewRequest("GET", "/api/v1/algorithm-optimization/history?limit=10", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Serve request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "history")
		assert.Contains(t, response, "total")
		assert.Contains(t, response, "limit")
		assert.Equal(t, float64(10), response["limit"])
	})

	t.Run("test get active optimizations endpoint", func(t *testing.T) {
		// Create request
		req := httptest.NewRequest("GET", "/api/v1/algorithm-optimization/active", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Serve request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "active_optimizations")
		assert.Contains(t, response, "count")
	})

	t.Run("test get optimization summary endpoint", func(t *testing.T) {
		// Create request
		req := httptest.NewRequest("GET", "/api/v1/algorithm-optimization/summary", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Serve request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "total_optimizations")
		assert.Contains(t, response, "successful_optimizations")
		assert.Contains(t, response, "failed_optimizations")
		assert.Contains(t, response, "average_improvement")
	})

	t.Run("test get optimization recommendations endpoint", func(t *testing.T) {
		// Create request
		req := httptest.NewRequest("GET", "/api/v1/algorithm-optimization/recommendations", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Serve request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "recommendations")
		assert.Contains(t, response, "count")
		assert.Contains(t, response, "generated_at")
	})

	t.Run("test get optimizations by type endpoint", func(t *testing.T) {
		// Create request
		req := httptest.NewRequest("GET", "/api/v1/algorithm-optimization/type/threshold", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Serve request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "optimizations")
		assert.Contains(t, response, "type")
		assert.Contains(t, response, "count")
		assert.Equal(t, "threshold", response["type"])
	})

	t.Run("test get optimizations by algorithm endpoint", func(t *testing.T) {
		// Create request
		req := httptest.NewRequest("GET", "/api/v1/algorithm-optimization/algorithm/test-algorithm", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Serve request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "optimizations")
		assert.Contains(t, response, "algorithm_id")
		assert.Contains(t, response, "count")
		assert.Equal(t, "test-algorithm", response["algorithm_id"])
	})

	t.Run("test cancel optimization endpoint", func(t *testing.T) {
		// Create request
		req := httptest.NewRequest("POST", "/api/v1/algorithm-optimization/test-id/cancel", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Serve request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Should return 404 since optimization doesn't exist
	})

	t.Run("test rollback optimization endpoint", func(t *testing.T) {
		// Create request body
		requestBody := map[string]interface{}{
			"reason": "Performance degradation detected",
		}

		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		// Create request
		req := httptest.NewRequest("POST", "/api/v1/algorithm-optimization/test-id/rollback", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Serve request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "success", response["status"])
		assert.Contains(t, response["message"], "rollback")
		assert.Equal(t, "test-id", response["id"])
		assert.Equal(t, "Performance degradation detected", response["reason"])
	})

	t.Run("test get optimization by id endpoint - not found", func(t *testing.T) {
		// Create request
		req := httptest.NewRequest("GET", "/api/v1/algorithm-optimization/non-existent-id", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Serve request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("test invalid request body", func(t *testing.T) {
		// Create invalid request body
		body := bytes.NewBufferString("invalid json")

		// Create request
		req := httptest.NewRequest("POST", "/api/v1/algorithm-optimization/analyze", body)
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Serve request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAlgorithmOptimizationWithRealData(t *testing.T) {
	logger := zap.NewNop()

	// Create algorithm optimizer
	config := &classification_optimization.OptimizationConfig{
		MinPatternsForOptimization: 1, // Lower threshold for testing
		OptimizationWindowHours:    24,
		MaxOptimizationsPerDay:     10,
		ConfidenceThreshold:        0.7,
		PerformanceThreshold:       0.05,
		OptimizationTimeout:        30 * time.Minute,
		EnableAutoOptimization:     true,
	}

	optimizer := classification_optimization.NewAlgorithmOptimizer(config, logger)

	// Create pattern analyzer
	patternAnalyzer := classification_monitoring.NewPatternAnalysisEngine(nil, logger)
	optimizer.SetPatternAnalyzer(patternAnalyzer)

	// Note: In a real implementation, algorithms would be registered through the optimizer
	// For testing purposes, we'll work with the default state

	// Create handler
	handler := handlers.NewAlgorithmOptimizationHandler(optimizer, logger)

	// Create router and register routes
	router := mux.NewRouter()
	routes.RegisterAlgorithmOptimizationRoutes(router, handler)

	t.Run("test complete optimization workflow", func(t *testing.T) {
		// Step 1: Trigger analysis and optimization
		requestBody := map[string]interface{}{
			"force_optimization": true,
		}

		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/algorithm-optimization/analyze", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Step 2: Check optimization history
		req = httptest.NewRequest("GET", "/api/v1/algorithm-optimization/history", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var historyResponse map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &historyResponse)
		require.NoError(t, err)

		// Step 3: Check optimization summary
		req = httptest.NewRequest("GET", "/api/v1/algorithm-optimization/summary", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var summaryResponse map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &summaryResponse)
		require.NoError(t, err)

		// Step 4: Check recommendations
		req = httptest.NewRequest("GET", "/api/v1/algorithm-optimization/recommendations", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var recommendationsResponse map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &recommendationsResponse)
		require.NoError(t, err)

		assert.Contains(t, recommendationsResponse, "recommendations")
		assert.Contains(t, recommendationsResponse, "count")
	})
}
