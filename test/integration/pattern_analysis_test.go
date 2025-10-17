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

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/api/routes"
	"kyb-platform/internal/modules/classification_monitoring"
)

func TestPatternAnalysisIntegration(t *testing.T) {
	logger := zap.NewNop()

	// Create pattern analysis engine
	config := &classification_monitoring.PatternAnalysisConfig{
		MinConfidenceThreshold:   0.5,
		MaxPatternsToTrack:       100,
		AnalysisWindowHours:      24,
		MinOccurrencesForPattern: 3,
	}

	patternEngine := classification_monitoring.NewPatternAnalysisEngine(config, logger)
	handler := handlers.NewPatternAnalysisHandler(patternEngine, logger)

	// Create router and register routes
	router := mux.NewRouter()
	routes.RegisterPatternAnalysisRoutes(router, handler)

	t.Run("analyze misclassifications", func(t *testing.T) {
		misclassifications := []*classification_monitoring.MisclassificationRecord{
			{
				ID:                     "test-1",
				Timestamp:              time.Now(),
				BusinessName:           "Test Business 1",
				ExpectedClassification: "Technology",
				ActualClassification:   "Finance",
				ConfidenceScore:        0.9,
				ClassificationMethod:   "ml",
				InputData:              map[string]interface{}{"text": "software company"},
				ErrorType:              "misclassification",
				Severity:               "high",
			},
			{
				ID:                     "test-2",
				Timestamp:              time.Now(),
				BusinessName:           "Test Business 2",
				ExpectedClassification: "Technology",
				ActualClassification:   "Finance",
				ConfidenceScore:        0.8,
				ClassificationMethod:   "ml",
				InputData:              map[string]interface{}{"text": "tech startup"},
				ErrorType:              "misclassification",
				Severity:               "high",
			},
		}

		requestBody := map[string]interface{}{
			"misclassifications": misclassifications,
		}

		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/pattern-analysis/analyze", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.NotNil(t, response["result"])
		assert.Equal(t, float64(2), response["metadata"].(map[string]interface{})["count"])
	})

	t.Run("get patterns", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/pattern-analysis/patterns", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.NotNil(t, response["patterns"])
	})

	t.Run("get patterns by type", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/pattern-analysis/patterns/type/temporal", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Equal(t, "temporal", response["pattern_type"])
	})

	t.Run("get patterns by severity", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/pattern-analysis/patterns/severity/high", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Equal(t, "high", response["severity"])
	})

	t.Run("get pattern history", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/pattern-analysis/history?limit=10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.NotNil(t, response["history"])
		assert.Equal(t, float64(10), response["limit"])
	})

	t.Run("get pattern summary", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/pattern-analysis/summary", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.NotNil(t, response["summary"])

		summary := response["summary"].(map[string]interface{})
		assert.NotNil(t, summary["total_patterns"])
		assert.NotNil(t, summary["risk_level"])
	})

	t.Run("get recommendations", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/pattern-analysis/recommendations", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.NotNil(t, response["recommendations"])
	})

	t.Run("get pattern details", func(t *testing.T) {
		// First analyze some misclassifications to create patterns
		misclassifications := []*classification_monitoring.MisclassificationRecord{
			{
				ID:                     "test-3",
				Timestamp:              time.Now(),
				BusinessName:           "Test Business 3",
				ExpectedClassification: "Technology",
				ActualClassification:   "Finance",
				ConfidenceScore:        0.9,
				ClassificationMethod:   "ml",
				InputData:              map[string]interface{}{"text": "software company"},
				ErrorType:              "misclassification",
				Severity:               "high",
			},
		}

		requestBody := map[string]interface{}{
			"misclassifications": misclassifications,
		}

		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/pattern-analysis/analyze", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Get patterns to find a pattern ID
		req = httptest.NewRequest("GET", "/api/v1/pattern-analysis/patterns", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var patternsResponse map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &patternsResponse)
		require.NoError(t, err)

		patterns := patternsResponse["patterns"].(map[string]interface{})
		if len(patterns) > 0 {
			// Get the first pattern ID
			var patternID string
			for id := range patterns {
				patternID = id
				break
			}

			// Get pattern details
			req = httptest.NewRequest("GET", "/api/v1/pattern-analysis/patterns/"+patternID, nil)
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.True(t, response["success"].(bool))
			assert.NotNil(t, response["pattern"])
		}
	})

	t.Run("get pattern details - not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/pattern-analysis/patterns/nonexistent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("health check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/pattern-analysis/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "healthy", response["status"])
		assert.NotNil(t, response["stats"])
	})

	t.Run("analyze misclassifications - empty request", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"misclassifications": []interface{}{},
		}

		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/pattern-analysis/analyze", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("analyze misclassifications - invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/pattern-analysis/analyze", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
