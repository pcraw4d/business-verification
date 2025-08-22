package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/external"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewVerificationBenchmarkingHandler(t *testing.T) {
	logger := zap.NewNop()
	manager := external.NewVerificationBenchmarkManager(nil, logger)

	handler := NewVerificationBenchmarkingHandler(manager, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, manager, handler.manager)
	assert.Equal(t, logger, handler.logger)
}

func TestVerificationBenchmarkingHandler_RegisterRoutes(t *testing.T) {
	logger := zap.NewNop()
	manager := external.NewVerificationBenchmarkManager(nil, logger)
	handler := NewVerificationBenchmarkingHandler(manager, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Test that routes are registered correctly
	testRoutes := []struct {
		method string
		path   string
	}{
		{"POST", "/api/benchmarking/suites"},
		{"GET", "/api/benchmarking/suites"},
		{"GET", "/api/benchmarking/suites/test-id"},
		{"POST", "/api/benchmarking/run"},
		{"GET", "/api/benchmarking/results"},
		{"POST", "/api/benchmarking/compare"},
		{"GET", "/api/benchmarking/config"},
		{"PUT", "/api/benchmarking/config"},
	}

	for _, route := range testRoutes {
		req := httptest.NewRequest(route.method, route.path, nil)
		rr := httptest.NewRecorder()

		// This should not panic, indicating the route is registered
		assert.NotPanics(t, func() {
			mux.ServeHTTP(rr, req)
		})
	}
}

func TestVerificationBenchmarkingHandler_CreateBenchmarkSuite(t *testing.T) {
	logger := zap.NewNop()
	manager := external.NewVerificationBenchmarkManager(nil, logger)
	handler := NewVerificationBenchmarkingHandler(manager, logger)

	t.Run("successful creation", func(t *testing.T) {
		reqData := CreateBenchmarkSuiteRequest{
			Name:        "Test Suite",
			Description: "A test benchmark suite",
			Category:    "verification",
			TestCases: []*external.BenchmarkTestCase{
				{
					Name:           "Test Case 1",
					Description:    "First test case",
					Input:          "test input",
					ExpectedOutput: "test output",
					GroundTruth: &external.VerificationResult{
						Status:       external.StatusPassed,
						OverallScore: 0.9,
					},
					Weight:     1.0,
					Difficulty: "easy",
				},
			},
		}

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest("POST", "/api/benchmarking/suites", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CreateBenchmarkSuite(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var response CreateBenchmarkSuiteResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.ID)
		assert.Equal(t, "Test Suite", response.Name)
		assert.Equal(t, "Benchmark suite created successfully", response.Message)
	})

	t.Run("missing name", func(t *testing.T) {
		reqData := CreateBenchmarkSuiteRequest{
			Description: "A test benchmark suite",
			Category:    "verification",
		}

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest("POST", "/api/benchmarking/suites", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CreateBenchmarkSuite(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Suite name is required")
	})

	t.Run("missing category", func(t *testing.T) {
		reqData := CreateBenchmarkSuiteRequest{
			Name:        "Test Suite",
			Description: "A test benchmark suite",
		}

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest("POST", "/api/benchmarking/suites", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CreateBenchmarkSuite(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Suite category is required")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/benchmarking/suites", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CreateBenchmarkSuite(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Invalid request format")
	})
}

func TestVerificationBenchmarkingHandler_GetBenchmarkSuite(t *testing.T) {
	logger := zap.NewNop()
	manager := external.NewVerificationBenchmarkManager(nil, logger)
	handler := NewVerificationBenchmarkingHandler(manager, logger)

	// Create a test suite first
	suite := &external.BenchmarkSuite{
		Name:        "Test Suite",
		Description: "A test benchmark suite",
		Category:    "verification",
	}
	err := manager.CreateBenchmarkSuite(suite)
	assert.NoError(t, err)

	t.Run("successful retrieval", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/benchmarking/suites/"+suite.ID, nil)
		req.SetPathValue("suiteID", suite.ID)
		rr := httptest.NewRecorder()

		handler.GetBenchmarkSuite(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response GetBenchmarkSuiteResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Suite)
		assert.Equal(t, suite.ID, response.Suite.ID)
		assert.Equal(t, "Test Suite", response.Suite.Name)
	})

	t.Run("suite not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/benchmarking/suites/non-existent", nil)
		req.SetPathValue("suiteID", "non-existent")
		rr := httptest.NewRecorder()

		handler.GetBenchmarkSuite(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Benchmark suite not found")
	})

	t.Run("missing suite ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/benchmarking/suites/", nil)
		req.SetPathValue("suiteID", "")
		rr := httptest.NewRecorder()

		handler.GetBenchmarkSuite(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Suite ID is required")
	})
}

func TestVerificationBenchmarkingHandler_ListBenchmarkSuites(t *testing.T) {
	logger := zap.NewNop()
	manager := external.NewVerificationBenchmarkManager(nil, logger)
	handler := NewVerificationBenchmarkingHandler(manager, logger)

	// Create test suites
	suite1 := &external.BenchmarkSuite{
		Name:        "Suite A",
		Description: "First test suite",
		Category:    "verification",
	}
	suite2 := &external.BenchmarkSuite{
		Name:        "Suite B",
		Description: "Second test suite",
		Category:    "performance",
	}

	err := manager.CreateBenchmarkSuite(suite1)
	assert.NoError(t, err)
	err = manager.CreateBenchmarkSuite(suite2)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/benchmarking/suites", nil)
	rr := httptest.NewRecorder()

	handler.ListBenchmarkSuites(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response ListBenchmarkSuitesResponse
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response.Suites, 2)
	assert.Equal(t, 2, response.Total)
}

func TestVerificationBenchmarkingHandler_RunBenchmark(t *testing.T) {
	logger := zap.NewNop()
	manager := external.NewVerificationBenchmarkManager(nil, logger)
	handler := NewVerificationBenchmarkingHandler(manager, logger)

	// Create a test suite with test cases
	suite := &external.BenchmarkSuite{
		Name:        "Test Suite",
		Description: "A test benchmark suite",
		Category:    "verification",
		TestCases: []*external.BenchmarkTestCase{
			{
				Name:           "Test Case 1",
				Description:    "First test case",
				Input:          "test input",
				ExpectedOutput: "test output",
				GroundTruth: &external.VerificationResult{
					Status:       external.StatusPassed,
					OverallScore: 0.9,
				},
				Weight:     1.0,
				Difficulty: "easy",
			},
		},
	}
	err := manager.CreateBenchmarkSuite(suite)
	assert.NoError(t, err)

	t.Run("successful run", func(t *testing.T) {
		reqData := RunBenchmarkRequest{
			SuiteID: suite.ID,
		}

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest("POST", "/api/benchmarking/run", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.RunBenchmark(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response RunBenchmarkResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.ID)
		assert.Equal(t, suite.ID, response.SuiteID)
		assert.Equal(t, "Test Suite", response.SuiteName)
		assert.Equal(t, "completed", response.Status)
		assert.NotNil(t, response.Metrics)
	})

	t.Run("missing suite ID", func(t *testing.T) {
		reqData := RunBenchmarkRequest{}

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest("POST", "/api/benchmarking/run", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.RunBenchmark(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Suite ID is required")
	})

	t.Run("suite not found", func(t *testing.T) {
		reqData := RunBenchmarkRequest{
			SuiteID: "non-existent",
		}

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest("POST", "/api/benchmarking/run", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.RunBenchmark(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Failed to run benchmark")
	})
}

func TestVerificationBenchmarkingHandler_GetBenchmarkResults(t *testing.T) {
	logger := zap.NewNop()
	manager := external.NewVerificationBenchmarkManager(nil, logger)
	handler := NewVerificationBenchmarkingHandler(manager, logger)

	// Create and run a benchmark to generate results
	suite := &external.BenchmarkSuite{
		Name:        "Test Suite",
		Description: "A test benchmark suite",
		Category:    "verification",
		TestCases: []*external.BenchmarkTestCase{
			{
				Name:           "Test Case 1",
				Description:    "First test case",
				Input:          "test input",
				ExpectedOutput: "test output",
				GroundTruth: &external.VerificationResult{
					Status:       external.StatusPassed,
					OverallScore: 0.9,
				},
				Weight:     1.0,
				Difficulty: "easy",
			},
		},
	}
	err := manager.CreateBenchmarkSuite(suite)
	assert.NoError(t, err)

	_, err = manager.RunBenchmark(context.Background(), suite.ID)
	assert.NoError(t, err)

	t.Run("successful retrieval", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/benchmarking/results", nil)
		rr := httptest.NewRecorder()

		handler.GetBenchmarkResults(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response GetBenchmarkResultsResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Len(t, response.Results, 1)
		assert.Equal(t, 1, response.Total)
	})

	t.Run("with limit parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/benchmarking/results?limit=5", nil)
		rr := httptest.NewRecorder()

		handler.GetBenchmarkResults(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response GetBenchmarkResultsResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(response.Results), 5)
	})
}

func TestVerificationBenchmarkingHandler_CompareBenchmarks(t *testing.T) {
	logger := zap.NewNop()
	manager := external.NewVerificationBenchmarkManager(nil, logger)
	handler := NewVerificationBenchmarkingHandler(manager, logger)

	// Create and run benchmarks to generate results
	suite := &external.BenchmarkSuite{
		Name:        "Test Suite",
		Description: "A test benchmark suite",
		Category:    "verification",
		TestCases: []*external.BenchmarkTestCase{
			{
				Name:           "Test Case 1",
				Description:    "First test case",
				Input:          "test input",
				ExpectedOutput: "test output",
				GroundTruth: &external.VerificationResult{
					Status:       external.StatusPassed,
					OverallScore: 0.9,
				},
				Weight:     1.0,
				Difficulty: "easy",
			},
		},
	}
	err := manager.CreateBenchmarkSuite(suite)
	assert.NoError(t, err)

	result1, err := manager.RunBenchmark(context.Background(), suite.ID)
	assert.NoError(t, err)

	result2, err := manager.RunBenchmark(context.Background(), suite.ID)
	assert.NoError(t, err)

	t.Run("successful comparison", func(t *testing.T) {
		reqData := CompareBenchmarksRequest{
			BaselineID:   result1.ID,
			ComparisonID: result2.ID,
		}

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest("POST", "/api/benchmarking/compare", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CompareBenchmarks(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response CompareBenchmarksResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Comparison)
		assert.Equal(t, result1.ID, response.Comparison.BaselineID)
		assert.Equal(t, result2.ID, response.Comparison.ComparisonID)
	})

	t.Run("missing baseline ID", func(t *testing.T) {
		reqData := CompareBenchmarksRequest{
			ComparisonID: result2.ID,
		}

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest("POST", "/api/benchmarking/compare", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CompareBenchmarks(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Baseline ID is required")
	})

	t.Run("missing comparison ID", func(t *testing.T) {
		reqData := CompareBenchmarksRequest{
			BaselineID: result1.ID,
		}

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest("POST", "/api/benchmarking/compare", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CompareBenchmarks(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Comparison ID is required")
	})
}

func TestVerificationBenchmarkingHandler_GetBenchmarkConfig(t *testing.T) {
	logger := zap.NewNop()
	manager := external.NewVerificationBenchmarkManager(nil, logger)
	handler := NewVerificationBenchmarkingHandler(manager, logger)

	req := httptest.NewRequest("GET", "/api/benchmarking/config", nil)
	rr := httptest.NewRecorder()

	handler.GetBenchmarkConfig(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response GetBenchmarkConfigResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.NotNil(t, response.Config)
	assert.True(t, response.Config.EnableBenchmarking)
	assert.Equal(t, 24*time.Hour, response.Config.BenchmarkInterval)
}

func TestVerificationBenchmarkingHandler_UpdateBenchmarkConfig(t *testing.T) {
	logger := zap.NewNop()
	manager := external.NewVerificationBenchmarkManager(nil, logger)
	handler := NewVerificationBenchmarkingHandler(manager, logger)

	t.Run("successful update", func(t *testing.T) {
		reqData := UpdateBenchmarkConfigRequest{
			Config: &external.BenchmarkConfig{
				EnableBenchmarking: false,
				AccuracyThreshold:  0.95,
				BenchmarkInterval:  12 * time.Hour,
				MinSampleSize:      200,
				ConfidenceLevel:    0.99,
			},
		}

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest("PUT", "/api/benchmarking/config", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.UpdateBenchmarkConfig(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response UpdateBenchmarkConfigResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "Benchmark configuration updated successfully", response.Message)
		assert.NotNil(t, response.Config)
		assert.False(t, response.Config.EnableBenchmarking)
		assert.Equal(t, 0.95, response.Config.AccuracyThreshold)
	})

	t.Run("missing config", func(t *testing.T) {
		reqData := UpdateBenchmarkConfigRequest{}

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest("PUT", "/api/benchmarking/config", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.UpdateBenchmarkConfig(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Config is required")
	})

	t.Run("invalid config", func(t *testing.T) {
		reqData := UpdateBenchmarkConfigRequest{
			Config: &external.BenchmarkConfig{
				AccuracyThreshold: 1.5, // Invalid: > 1
			},
		}

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest("PUT", "/api/benchmarking/config", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.UpdateBenchmarkConfig(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Failed to update config")
	})
}

func TestBenchmarkingHandler_ValidationErrors(t *testing.T) {
	logger := zap.NewNop()
	manager := external.NewVerificationBenchmarkManager(nil, logger)
	handler := NewVerificationBenchmarkingHandler(manager, logger)

	// Test invalid JSON in various endpoints
	endpoints := []struct {
		method   string
		path     string
		expected int
	}{
		{"POST", "/api/benchmarking/suites", http.StatusBadRequest},
		{"POST", "/api/benchmarking/run", http.StatusBadRequest},
		{"POST", "/api/benchmarking/compare", http.StatusBadRequest},
		{"PUT", "/api/benchmarking/config", http.StatusBadRequest},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.method+" "+endpoint.path, func(t *testing.T) {
			req := httptest.NewRequest(endpoint.method, endpoint.path, bytes.NewReader([]byte("invalid json")))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			switch endpoint.path {
			case "/api/benchmarking/suites":
				handler.CreateBenchmarkSuite(rr, req)
			case "/api/benchmarking/run":
				handler.RunBenchmark(rr, req)
			case "/api/benchmarking/compare":
				handler.CompareBenchmarks(rr, req)
			case "/api/benchmarking/config":
				handler.UpdateBenchmarkConfig(rr, req)
			}

			assert.Equal(t, endpoint.expected, rr.Code)

			var response map[string]string
			err := json.NewDecoder(rr.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Contains(t, response["error"], "Invalid request format")
		})
	}
}
