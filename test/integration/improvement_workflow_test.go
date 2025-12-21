//go:build !comprehensive_test && !e2e_railway
// +build !comprehensive_test,!e2e_railway

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/api/routes"
	"kyb-platform/internal/modules/classification_optimization"
)

func TestImprovementWorkflowAPI_StartContinuousImprovement(t *testing.T) {
	logger := zap.NewNop()

	// Set up dependencies
	algorithmRegistry := classification_optimization.NewAlgorithmRegistry(logger)
	performanceTracker := classification_optimization.NewPerformanceTracker(logger)
	accuracyValidator := classification_optimization.NewAccuracyValidator(nil, logger)
	accuracyValidator.SetAlgorithmRegistry(algorithmRegistry)

	// Create workflow
	workflow := classification_optimization.NewImprovementWorkflow(nil, logger)
	workflow.SetDependencies(algorithmRegistry, performanceTracker, accuracyValidator, nil)

	// Register test algorithm
	algorithm := &classification_optimization.ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	algorithmRegistry.RegisterAlgorithm(algorithm)

	// Create handler
	handler := handlers.NewImprovementWorkflowHandler(workflow, logger)

	// Create router
	router := mux.NewRouter()
	routes.RegisterImprovementWorkflowRoutes(router, handler)

	// Test request
	requestBody := map[string]interface{}{
		"algorithm_id": "test-algorithm",
	}
	requestJSON, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/v1/workflows/continuous-improvement", bytes.NewBuffer(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "test-algorithm", response["algorithm_id"])
	assert.Equal(t, "completed", response["status"])
	assert.Equal(t, "continuous_improvement", response["type"])
}

func TestImprovementWorkflowAPI_StartABTesting(t *testing.T) {
	logger := zap.NewNop()

	// Set up dependencies
	algorithmRegistry := classification_optimization.NewAlgorithmRegistry(logger)
	performanceTracker := classification_optimization.NewPerformanceTracker(logger)
	accuracyValidator := classification_optimization.NewAccuracyValidator(nil, logger)
	accuracyValidator.SetAlgorithmRegistry(algorithmRegistry)

	// Create workflow
	workflow := classification_optimization.NewImprovementWorkflow(nil, logger)
	workflow.SetDependencies(algorithmRegistry, performanceTracker, accuracyValidator, nil)

	// Register test algorithms
	algorithmA := &classification_optimization.ClassificationAlgorithm{
		ID:                  "algorithm-a",
		Name:                "Algorithm A",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	algorithmB := &classification_optimization.ClassificationAlgorithm{
		ID:                  "algorithm-b",
		Name:                "Algorithm B",
		Category:            "test-category",
		ConfidenceThreshold: 0.8,
		IsActive:            true,
	}
	algorithmRegistry.RegisterAlgorithm(algorithmA)
	algorithmRegistry.RegisterAlgorithm(algorithmB)

	// Create handler
	handler := handlers.NewImprovementWorkflowHandler(workflow, logger)

	// Create router
	router := mux.NewRouter()
	routes.RegisterImprovementWorkflowRoutes(router, handler)

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

	// Test request
	requestBody := map[string]interface{}{
		"algorithm_a": "algorithm-a",
		"algorithm_b": "algorithm-b",
		"test_cases":  testCases,
	}
	requestJSON, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/v1/workflows/ab-testing", bytes.NewBuffer(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "algorithm-a", response["algorithm_id"])
	assert.Equal(t, "completed", response["status"])
	assert.Equal(t, "ab_testing", response["type"])
}

func TestImprovementWorkflowAPI_GetWorkflowHistory(t *testing.T) {
	logger := zap.NewNop()

	// Set up dependencies
	algorithmRegistry := classification_optimization.NewAlgorithmRegistry(logger)
	performanceTracker := classification_optimization.NewPerformanceTracker(logger)
	accuracyValidator := classification_optimization.NewAccuracyValidator(nil, logger)
	accuracyValidator.SetAlgorithmRegistry(algorithmRegistry)

	// Create workflow
	workflow := classification_optimization.NewImprovementWorkflow(nil, logger)
	workflow.SetDependencies(algorithmRegistry, performanceTracker, accuracyValidator, nil)

	// Register test algorithm and run a workflow
	algorithm := &classification_optimization.ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	algorithmRegistry.RegisterAlgorithm(algorithm)

	// Run a workflow to populate history
	_, err := workflow.StartContinuousImprovement(context.Background(), "test-algorithm")
	assert.NoError(t, err)

	// Create handler
	handler := handlers.NewImprovementWorkflowHandler(workflow, logger)

	// Create router
	router := mux.NewRouter()
	routes.RegisterImprovementWorkflowRoutes(router, handler)

	// Test request
	req := httptest.NewRequest("GET", "/api/v1/workflows/history", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "workflows")
	assert.Contains(t, response, "total")
	assert.Contains(t, response, "limit")

	workflows := response["workflows"].([]interface{})
	assert.GreaterOrEqual(t, len(workflows), 1)
}

func TestImprovementWorkflowAPI_GetActiveWorkflows(t *testing.T) {
	logger := zap.NewNop()

	// Set up dependencies
	algorithmRegistry := classification_optimization.NewAlgorithmRegistry(logger)
	performanceTracker := classification_optimization.NewPerformanceTracker(logger)
	accuracyValidator := classification_optimization.NewAccuracyValidator(nil, logger)
	accuracyValidator.SetAlgorithmRegistry(algorithmRegistry)

	// Create workflow
	workflow := classification_optimization.NewImprovementWorkflow(nil, logger)
	workflow.SetDependencies(algorithmRegistry, performanceTracker, accuracyValidator, nil)

	// Create handler
	handler := handlers.NewImprovementWorkflowHandler(workflow, logger)

	// Create router
	router := mux.NewRouter()
	routes.RegisterImprovementWorkflowRoutes(router, handler)

	// Test request
	req := httptest.NewRequest("GET", "/api/v1/workflows/active", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "active_workflows")
	assert.Contains(t, response, "count")

	// Should be empty initially
	assert.Equal(t, float64(0), response["count"])
}

func TestImprovementWorkflowAPI_GetWorkflowExecution(t *testing.T) {
	logger := zap.NewNop()

	// Set up dependencies
	algorithmRegistry := classification_optimization.NewAlgorithmRegistry(logger)
	performanceTracker := classification_optimization.NewPerformanceTracker(logger)
	accuracyValidator := classification_optimization.NewAccuracyValidator(nil, logger)
	accuracyValidator.SetAlgorithmRegistry(algorithmRegistry)

	// Create workflow
	workflow := classification_optimization.NewImprovementWorkflow(nil, logger)
	workflow.SetDependencies(algorithmRegistry, performanceTracker, accuracyValidator, nil)

	// Register test algorithm and run a workflow
	algorithm := &classification_optimization.ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	algorithmRegistry.RegisterAlgorithm(algorithm)

	// Run a workflow to get execution ID
	execution, err := workflow.StartContinuousImprovement(context.Background(), "test-algorithm")
	assert.NoError(t, err)

	// Create handler
	handler := handlers.NewImprovementWorkflowHandler(workflow, logger)

	// Create router
	router := mux.NewRouter()
	routes.RegisterImprovementWorkflowRoutes(router, handler)

	// Test request
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/workflows/%s", execution.ID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, execution.ID, response["id"])
	assert.Equal(t, "test-algorithm", response["algorithm_id"])
	assert.Equal(t, "completed", response["status"])
}

func TestImprovementWorkflowAPI_GetWorkflowStatistics(t *testing.T) {
	logger := zap.NewNop()

	// Set up dependencies
	algorithmRegistry := classification_optimization.NewAlgorithmRegistry(logger)
	performanceTracker := classification_optimization.NewPerformanceTracker(logger)
	accuracyValidator := classification_optimization.NewAccuracyValidator(nil, logger)
	accuracyValidator.SetAlgorithmRegistry(algorithmRegistry)

	// Create workflow
	workflow := classification_optimization.NewImprovementWorkflow(nil, logger)
	workflow.SetDependencies(algorithmRegistry, performanceTracker, accuracyValidator, nil)

	// Register test algorithm and run a workflow
	algorithm := &classification_optimization.ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	algorithmRegistry.RegisterAlgorithm(algorithm)

	// Run a workflow to populate statistics
	_, err := workflow.StartContinuousImprovement(context.Background(), "test-algorithm")
	assert.NoError(t, err)

	// Create handler
	handler := handlers.NewImprovementWorkflowHandler(workflow, logger)

	// Create router
	router := mux.NewRouter()
	routes.RegisterImprovementWorkflowRoutes(router, handler)

	// Test request
	req := httptest.NewRequest("GET", "/api/v1/workflows/statistics", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "total_executions")
	assert.Contains(t, response, "active_workflows")
	assert.Contains(t, response, "completed_executions")
	assert.Contains(t, response, "failed_executions")
	assert.Contains(t, response, "success_rate")
	assert.Contains(t, response, "average_improvement")
	assert.Contains(t, response, "workflow_types")
	assert.Contains(t, response, "status_distribution")

	// Should have at least one execution
	assert.GreaterOrEqual(t, response["total_executions"], float64(1))
}

func TestImprovementWorkflowAPI_GetWorkflowRecommendations(t *testing.T) {
	logger := zap.NewNop()

	// Set up dependencies
	algorithmRegistry := classification_optimization.NewAlgorithmRegistry(logger)
	performanceTracker := classification_optimization.NewPerformanceTracker(logger)
	accuracyValidator := classification_optimization.NewAccuracyValidator(nil, logger)
	accuracyValidator.SetAlgorithmRegistry(algorithmRegistry)

	// Create workflow
	workflow := classification_optimization.NewImprovementWorkflow(nil, logger)
	workflow.SetDependencies(algorithmRegistry, performanceTracker, accuracyValidator, nil)

	// Create handler
	handler := handlers.NewImprovementWorkflowHandler(workflow, logger)

	// Create router
	router := mux.NewRouter()
	routes.RegisterImprovementWorkflowRoutes(router, handler)

	// Test request
	req := httptest.NewRequest("GET", "/api/v1/workflows/recommendations", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "recommendations")
	assert.Contains(t, response, "total")

	recommendations := response["recommendations"].([]interface{})
	assert.GreaterOrEqual(t, len(recommendations), 1)
}

func TestImprovementWorkflowAPI_GetWorkflowMetrics(t *testing.T) {
	logger := zap.NewNop()

	// Set up dependencies
	algorithmRegistry := classification_optimization.NewAlgorithmRegistry(logger)
	performanceTracker := classification_optimization.NewPerformanceTracker(logger)
	accuracyValidator := classification_optimization.NewAccuracyValidator(nil, logger)
	accuracyValidator.SetAlgorithmRegistry(algorithmRegistry)

	// Create workflow
	workflow := classification_optimization.NewImprovementWorkflow(nil, logger)
	workflow.SetDependencies(algorithmRegistry, performanceTracker, accuracyValidator, nil)

	// Register test algorithm and run a workflow
	algorithm := &classification_optimization.ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	algorithmRegistry.RegisterAlgorithm(algorithm)

	// Run a workflow to populate metrics
	_, err := workflow.StartContinuousImprovement(context.Background(), "test-algorithm")
	assert.NoError(t, err)

	// Create handler
	handler := handlers.NewImprovementWorkflowHandler(workflow, logger)

	// Create router
	router := mux.NewRouter()
	routes.RegisterImprovementWorkflowRoutes(router, handler)

	// Test request
	req := httptest.NewRequest("GET", "/api/v1/workflows/metrics", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "total_workflows")
	assert.Contains(t, response, "completed_workflows")
	assert.Contains(t, response, "average_duration_seconds")
	assert.Contains(t, response, "average_iterations")
	assert.Contains(t, response, "average_accuracy_improvement")
	assert.Contains(t, response, "average_f1_improvement")
	assert.Contains(t, response, "average_confidence_improvement")
	assert.Contains(t, response, "total_iterations")
	assert.Contains(t, response, "total_duration_seconds")

	// Should have at least one workflow
	assert.GreaterOrEqual(t, response["total_workflows"], float64(1))
}

func TestImprovementWorkflowAPI_ErrorHandling(t *testing.T) {
	logger := zap.NewNop()

	// Set up dependencies
	algorithmRegistry := classification_optimization.NewAlgorithmRegistry(logger)
	performanceTracker := classification_optimization.NewPerformanceTracker(logger)
	accuracyValidator := classification_optimization.NewAccuracyValidator(nil, logger)
	accuracyValidator.SetAlgorithmRegistry(algorithmRegistry)

	// Create workflow
	workflow := classification_optimization.NewImprovementWorkflow(nil, logger)
	workflow.SetDependencies(algorithmRegistry, performanceTracker, accuracyValidator, nil)

	// Create handler
	handler := handlers.NewImprovementWorkflowHandler(workflow, logger)

	// Create router
	router := mux.NewRouter()
	routes.RegisterImprovementWorkflowRoutes(router, handler)

	// Test invalid request body
	req := httptest.NewRequest("POST", "/api/v1/workflows/continuous-improvement", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test missing algorithm_id
	requestBody := map[string]interface{}{}
	requestJSON, _ := json.Marshal(requestBody)

	req = httptest.NewRequest("POST", "/api/v1/workflows/continuous-improvement", bytes.NewBuffer(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test non-existent algorithm
	requestBody = map[string]interface{}{
		"algorithm_id": "non-existent",
	}
	requestJSON, _ = json.Marshal(requestBody)

	req = httptest.NewRequest("POST", "/api/v1/workflows/continuous-improvement", bytes.NewBuffer(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Test non-existent workflow
	req = httptest.NewRequest("GET", "/api/v1/workflows/non-existent", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test stopping non-existent workflow
	req = httptest.NewRequest("POST", "/api/v1/workflows/non-existent/stop", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
