package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/modules/classification_optimization"
)

// TestAccuracyValidationIntegration tests the full accuracy validation API flow
func TestAccuracyValidationIntegration(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	validator := classification_optimization.NewAccuracyValidator(nil, logger)
	handler := handlers.NewClassificationOptimizationValidationHandler(validator, logger)
	router := mux.NewRouter()

	// Register routes
	RegisterAccuracyValidationRoutes(router, handler)

	// Register a test algorithm
	algorithm := &classification_optimization.ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	algorithmRegistry := classification_optimization.NewAlgorithmRegistry(logger)
	validator.SetAlgorithmRegistry(algorithmRegistry)
	algorithmRegistry.RegisterAlgorithm(algorithm)

	// Create test cases
	testCases := make([]*classification_optimization.TestCase, 100)
	for i := 0; i < 100; i++ {
		testCases[i] = &classification_optimization.TestCase{
			ID:             fmt.Sprintf("test-%d", i),
			Input:          map[string]interface{}{"name": fmt.Sprintf("Company %d", i)},
			ExpectedOutput: "technology",
			TestCaseType:   "standard",
			Difficulty:     "easy",
		}
	}

	t.Run("validate accuracy", func(t *testing.T) {
		// Create request
		reqBody := handlers.ValidateAccuracyRequest{
			AlgorithmID: "test-algorithm",
			TestCases:   testCases,
		}
		reqJSON, _ := json.Marshal(reqBody)

		// Create HTTP request
		req := httptest.NewRequest("POST", "/api/v1/accuracy-validation/validate", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Execute request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.ValidateAccuracyResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.NotNil(t, response.Result)
		assert.Equal(t, "test-algorithm", response.Result.AlgorithmID)
		assert.Equal(t, classification_optimization.ValidationTypeAccuracy, response.Result.ValidationType)
		assert.Equal(t, classification_optimization.ValidationStatusCompleted, response.Result.Status)
		assert.Len(t, response.Result.TestCases, 100)
		assert.NotNil(t, response.Result.Metrics)
		assert.NotNil(t, response.Result.Recommendations)
	})

	t.Run("perform cross validation", func(t *testing.T) {
		// Create request
		reqBody := handlers.CrossValidationRequest{
			AlgorithmID: "test-algorithm",
			TestCases:   testCases,
		}
		reqJSON, _ := json.Marshal(reqBody)

		// Create HTTP request
		req := httptest.NewRequest("POST", "/api/v1/accuracy-validation/cross-validation", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Execute request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.CrossValidationResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.NotNil(t, response.Result)
		assert.Equal(t, "test-algorithm", response.Result.AlgorithmID)
		assert.Equal(t, classification_optimization.ValidationTypeCrossValidation, response.Result.ValidationType)
		assert.Equal(t, classification_optimization.ValidationStatusCompleted, response.Result.Status)
		assert.Len(t, response.Result.TestCases, 100)
		assert.NotNil(t, response.Result.Metrics)
		assert.NotNil(t, response.Result.Recommendations)
	})

	t.Run("get validation history", func(t *testing.T) {
		// Create HTTP request
		req := httptest.NewRequest("GET", "/api/v1/accuracy-validation/history", nil)
		w := httptest.NewRecorder()

		// Execute request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.ValidationHistoryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.GreaterOrEqual(t, response.Count, 2) // Should have at least 2 validations from previous tests
		assert.NotNil(t, response.History)
	})

	t.Run("get validation summary", func(t *testing.T) {
		// Create HTTP request
		req := httptest.NewRequest("GET", "/api/v1/accuracy-validation/summary", nil)
		w := httptest.NewRecorder()

		// Execute request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.ValidationSummaryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.NotNil(t, response.Summary)
		assert.GreaterOrEqual(t, response.Summary.TotalValidations, 2)
		assert.Equal(t, 0, response.Summary.ActiveValidations)
		assert.Greater(t, response.Summary.AverageAccuracy, 0.0)
		assert.Greater(t, response.Summary.AverageF1Score, 0.0)
	})

	t.Run("get active validations", func(t *testing.T) {
		// Create HTTP request
		req := httptest.NewRequest("GET", "/api/v1/accuracy-validation/active", nil)
		w := httptest.NewRecorder()

		// Execute request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.ActiveValidationsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, 0, response.Count) // No active validations after completion
		assert.NotNil(t, response.Active)
	})

	t.Run("get validations by algorithm", func(t *testing.T) {
		// Create HTTP request
		req := httptest.NewRequest("GET", "/api/v1/accuracy-validation/algorithm/test-algorithm", nil)
		w := httptest.NewRecorder()

		// Execute request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.ValidationHistoryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.GreaterOrEqual(t, response.Count, 2) // Should have at least 2 validations for this algorithm
		assert.NotNil(t, response.History)
	})

	t.Run("get validations by type", func(t *testing.T) {
		// Create HTTP request
		req := httptest.NewRequest("GET", "/api/v1/accuracy-validation/type/accuracy", nil)
		w := httptest.NewRecorder()

		// Execute request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.ValidationHistoryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.GreaterOrEqual(t, response.Count, 1) // Should have at least 1 accuracy validation
		assert.NotNil(t, response.History)
	})

	t.Run("health check", func(t *testing.T) {
		// Create HTTP request
		req := httptest.NewRequest("GET", "/api/v1/accuracy-validation/health", nil)
		w := httptest.NewRecorder()

		// Execute request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "healthy", response["status"])
		assert.Equal(t, "classification-optimization-validation", response["service"])
		assert.NotNil(t, response["timestamp"])
	})
}

// TestAccuracyValidationWithRealData tests accuracy validation with realistic test data
func TestAccuracyValidationWithRealData(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	validator := classification_optimization.NewAccuracyValidator(nil, logger)
	handler := handlers.NewClassificationOptimizationValidationHandler(validator, logger)
	router := mux.NewRouter()

	// Register routes
	RegisterAccuracyValidationRoutes(router, handler)

	// Register a test algorithm
	algorithm := &classification_optimization.ClassificationAlgorithm{
		ID:                  "real-algorithm",
		Name:                "Real Algorithm",
		Category:            "real-category",
		ConfidenceThreshold: 0.8,
		IsActive:            true,
	}
	algorithmRegistry := classification_optimization.NewAlgorithmRegistry(logger)
	validator.SetAlgorithmRegistry(algorithmRegistry)
	algorithmRegistry.RegisterAlgorithm(algorithm)

	// Create realistic test cases
	testCases := []*classification_optimization.TestCase{
		{
			ID:             "tech-1",
			Input:          map[string]interface{}{"name": "Microsoft Corporation"},
			ExpectedOutput: "technology",
			TestCaseType:   "standard",
			Difficulty:     "easy",
		},
		{
			ID:             "tech-2",
			Input:          map[string]interface{}{"name": "Apple Inc."},
			ExpectedOutput: "technology",
			TestCaseType:   "standard",
			Difficulty:     "easy",
		},
		{
			ID:             "retail-1",
			Input:          map[string]interface{}{"name": "Walmart Stores Inc."},
			ExpectedOutput: "retail",
			TestCaseType:   "standard",
			Difficulty:     "easy",
		},
		{
			ID:             "retail-2",
			Input:          map[string]interface{}{"name": "Target Corporation"},
			ExpectedOutput: "retail",
			TestCaseType:   "standard",
			Difficulty:     "easy",
		},
		{
			ID:             "finance-1",
			Input:          map[string]interface{}{"name": "JPMorgan Chase & Co."},
			ExpectedOutput: "finance",
			TestCaseType:   "standard",
			Difficulty:     "medium",
		},
	}

	// Add more test cases to meet minimum requirement
	for i := 5; i < 100; i++ {
		testCases = append(testCases, &classification_optimization.TestCase{
			ID:             fmt.Sprintf("test-%d", i),
			Input:          map[string]interface{}{"name": fmt.Sprintf("Company %d", i)},
			ExpectedOutput: "technology",
			TestCaseType:   "standard",
			Difficulty:     "easy",
		})
	}

	t.Run("validate accuracy with real data", func(t *testing.T) {
		// Create request
		reqBody := handlers.ValidateAccuracyRequest{
			AlgorithmID: "real-algorithm",
			TestCases:   testCases,
		}
		reqJSON, _ := json.Marshal(reqBody)

		// Create HTTP request
		req := httptest.NewRequest("POST", "/api/v1/accuracy-validation/validate", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Execute request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.ValidateAccuracyResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.NotNil(t, response.Result)
		assert.Equal(t, "real-algorithm", response.Result.AlgorithmID)
		assert.Equal(t, classification_optimization.ValidationTypeAccuracy, response.Result.ValidationType)
		assert.Equal(t, classification_optimization.ValidationStatusCompleted, response.Result.Status)
		assert.Len(t, response.Result.TestCases, 100)
		assert.NotNil(t, response.Result.Metrics)
		assert.NotNil(t, response.Result.Recommendations)

		// Check metrics
		metrics := response.Result.Metrics
		assert.Equal(t, 100, metrics.TotalTestCases)
		assert.GreaterOrEqual(t, metrics.PassedTestCases, 0)
		assert.GreaterOrEqual(t, metrics.FailedTestCases, 0)
		assert.GreaterOrEqual(t, metrics.Accuracy, 0.0)
		assert.LessOrEqual(t, metrics.Accuracy, 1.0)
		assert.GreaterOrEqual(t, metrics.F1Score, 0.0)
		assert.LessOrEqual(t, metrics.F1Score, 1.0)
		assert.GreaterOrEqual(t, metrics.AverageConfidence, 0.0)
		assert.LessOrEqual(t, metrics.AverageConfidence, 1.0)
	})

	t.Run("cross validation with real data", func(t *testing.T) {
		// Create request
		reqBody := handlers.CrossValidationRequest{
			AlgorithmID: "real-algorithm",
			TestCases:   testCases,
		}
		reqJSON, _ := json.Marshal(reqBody)

		// Create HTTP request
		req := httptest.NewRequest("POST", "/api/v1/accuracy-validation/cross-validation", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Execute request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.CrossValidationResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.NotNil(t, response.Result)
		assert.Equal(t, "real-algorithm", response.Result.AlgorithmID)
		assert.Equal(t, classification_optimization.ValidationTypeCrossValidation, response.Result.ValidationType)
		assert.Equal(t, classification_optimization.ValidationStatusCompleted, response.Result.Status)
		assert.Len(t, response.Result.TestCases, 100)
		assert.NotNil(t, response.Result.Metrics)
		assert.NotNil(t, response.Result.Recommendations)

		// Check cross validation metrics
		metrics := response.Result.Metrics
		assert.Equal(t, 100, metrics.TotalTestCases)
		assert.GreaterOrEqual(t, metrics.Accuracy, 0.0)
		assert.LessOrEqual(t, metrics.Accuracy, 1.0)
		assert.GreaterOrEqual(t, metrics.F1Score, 0.0)
		assert.LessOrEqual(t, metrics.F1Score, 1.0)
	})
}

// RegisterAccuracyValidationRoutes is a helper function to register routes for testing
func RegisterAccuracyValidationRoutes(router *mux.Router, handler *handlers.ClassificationOptimizationValidationHandler) {
	// Base path for accuracy validation endpoints
	basePath := "/api/v1/accuracy-validation"

	// Accuracy validation endpoints
	router.HandleFunc(basePath+"/validate", handler.ValidateAccuracy).Methods("POST")
	router.HandleFunc(basePath+"/cross-validation", handler.PerformCrossValidation).Methods("POST")

	// History and summary endpoints
	router.HandleFunc(basePath+"/history", handler.GetValidationHistory).Methods("GET")
	router.HandleFunc(basePath+"/summary", handler.GetValidationSummary).Methods("GET")
	router.HandleFunc(basePath+"/active", handler.GetActiveValidations).Methods("GET")

	// Specific validation endpoints
	router.HandleFunc(basePath+"/validation/{id}", handler.GetValidationByID).Methods("GET")
	router.HandleFunc(basePath+"/validation/{id}/cancel", handler.CancelValidation).Methods("POST")

	// Filtered history endpoints
	router.HandleFunc(basePath+"/algorithm/{algorithm_id}", handler.GetValidationsByAlgorithm).Methods("GET")
	router.HandleFunc(basePath+"/type/{type}", handler.GetValidationsByType).Methods("GET")

	// Health check endpoint
	router.HandleFunc(basePath+"/health", handler.HealthCheck).Methods("GET")
}
