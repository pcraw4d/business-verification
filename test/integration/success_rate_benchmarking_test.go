//go:build !comprehensive_test && !e2e_railway
// +build !comprehensive_test,!e2e_railway

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/modules/success_monitoring"
)

func TestSuccessRateBenchmarkingIntegration(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	benchmarkManager := success_monitoring.NewSuccessRateBenchmarkManager(nil, logger)
	benchmarkHandler := handlers.NewSuccessRateBenchmarkingHandler(benchmarkManager, logger)

	// Create router and register routes
	router := mux.NewRouter()
	router.HandleFunc("/api/v3/benchmarking/suites", benchmarkHandler.CreateBenchmarkSuite).Methods(http.MethodPost)
	router.HandleFunc("/api/v3/benchmarking/suites/{suiteId}/execute", benchmarkHandler.ExecuteBenchmark).Methods(http.MethodPost)
	router.HandleFunc("/api/v3/benchmarking/suites/{suiteId}/results", benchmarkHandler.GetBenchmarkResults).Methods(http.MethodGet)
	router.HandleFunc("/api/v3/benchmarking/suites/{suiteId}/report", benchmarkHandler.GenerateBenchmarkReport).Methods(http.MethodGet)
	router.HandleFunc("/api/v3/benchmarking/baselines", benchmarkHandler.UpdateBaseline).Methods(http.MethodPost)
	router.HandleFunc("/api/v3/benchmarking/baselines/{category}", benchmarkHandler.GetBaselineMetrics).Methods(http.MethodGet)
	router.HandleFunc("/api/v3/benchmarking/config", benchmarkHandler.GetBenchmarkConfiguration).Methods(http.MethodGet)

	t.Run("Create and Execute Benchmark Suite", func(t *testing.T) {
		// Create a benchmark suite
		suite := &success_monitoring.BenchmarkSuite{
			ID:          "test-suite-1",
			Name:        "Test Business Verification Suite",
			Description: "Comprehensive test suite for business verification",
			Category:    "business_verification",
			TestCases: []success_monitoring.BenchmarkTestCase{
				{
					ID:                  "test-case-1",
					Name:                "Basic Business Verification",
					Description:         "Test basic business verification functionality",
					Input:               map[string]interface{}{"business_name": "Test Corp", "address": "123 Test St"},
					ExpectedSuccessRate: 0.95,
					MaxDuration:         5 * time.Second,
				},
				{
					ID:                  "test-case-2",
					Name:                "Complex Business Verification",
					Description:         "Test complex business verification with multiple data sources",
					Input:               map[string]interface{}{"business_name": "Complex Corp", "address": "456 Complex Ave", "industry": "technology"},
					ExpectedSuccessRate: 0.90,
					MaxDuration:         10 * time.Second,
				},
			},
			SampleSize:    100,
			MaxIterations: 3,
		}

		suiteJSON, err := json.Marshal(suite)
		require.NoError(t, err)

		// Create benchmark suite
		req := httptest.NewRequest(http.MethodPost, "/api/v3/benchmarking/suites", nil)
		req.Body = httptest.NewRecorder().Result().Body
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Execute benchmark
		req = httptest.NewRequest(http.MethodPost, "/api/v3/benchmarking/suites/test-suite-1/execute", nil)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.ExecuteBenchmarkResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.NotNil(t, response.Result)
		assert.Equal(t, "test-suite-1", response.Result.SuiteID)
	})

	t.Run("Get Benchmark Results", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v3/benchmarking/suites/test-suite-1/results?limit=10", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.GetBenchmarkResultsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "test-suite-1", response.SuiteID)
		assert.GreaterOrEqual(t, response.Count, 0)
	})

	t.Run("Generate Benchmark Report", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v3/benchmarking/suites/test-suite-1/report", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.GenerateBenchmarkReportResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.NotNil(t, response.Report)
		assert.Equal(t, "test-suite-1", response.Report.SuiteID)
	})

	t.Run("Update and Get Baseline Metrics", func(t *testing.T) {
		// Update baseline
		baselineRequest := handlers.UpdateBaselineRequest{
			Category:    "business_verification",
			SuccessRate: 0.92,
			SampleCount: 1000,
		}

		baselineJSON, err := json.Marshal(baselineRequest)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/benchmarking/baselines", nil)
		req.Body = httptest.NewRecorder().Result().Body
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Get baseline metrics
		req = httptest.NewRequest(http.MethodGet, "/api/v3/benchmarking/baselines/business_verification", nil)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.GetBaselineMetricsResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.NotNil(t, response.Baseline)
		assert.Equal(t, "business_verification", response.Category)
		assert.Equal(t, 0.92, response.Baseline.SuccessRate)
		assert.Equal(t, 1000, response.Baseline.SampleCount)
	})

	t.Run("Get Benchmark Configuration", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v3/benchmarking/config", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.GetBenchmarkConfigurationResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.NotNil(t, response.Config)
		assert.Equal(t, 0.95, response.Config.TargetSuccessRate)
		assert.Equal(t, 0.95, response.Config.ConfidenceLevel)
		assert.Equal(t, 100, response.Config.MinSampleSize)
		assert.Equal(t, 10000, response.Config.MaxSampleSize)
	})
}

func TestSuccessRateBenchmarkingValidation(t *testing.T) {
	logger := zap.NewNop()
	benchmarkManager := success_monitoring.NewSuccessRateBenchmarkManager(nil, logger)

	t.Run("Statistical Validation", func(t *testing.T) {
		// Create test results with known success rates
		results := []*success_monitoring.BenchmarkResult{
			{
				SuiteID:     "test-suite",
				SuccessRate: 0.95,
				SampleSize:  100,
				Duration:    2 * time.Second,
				Timestamp:   time.Now(),
			},
			{
				SuiteID:     "test-suite",
				SuccessRate: 0.94,
				SampleSize:  100,
				Duration:    2 * time.Second,
				Timestamp:   time.Now(),
			},
			{
				SuiteID:     "test-suite",
				SuccessRate: 0.96,
				SampleSize:  100,
				Duration:    2 * time.Second,
				Timestamp:   time.Now(),
			},
		}

		// Test statistical validation
		validation := benchmarkManager.validateResults(results)
		assert.NotNil(t, validation)
		assert.True(t, validation.IsStatisticallySignificant)
		assert.Greater(t, validation.ConfidenceInterval, 0.0)
		assert.Less(t, validation.ConfidenceInterval, 1.0)
	})

	t.Run("Baseline Comparison", func(t *testing.T) {
		// Set up baseline
		ctx := context.Background()
		err := benchmarkManager.UpdateBaseline(ctx, "test_category", 0.90, 1000)
		require.NoError(t, err)

		// Create result to compare against baseline
		result := &success_monitoring.BenchmarkResult{
			SuiteID:     "test-suite",
			SuccessRate: 0.95,
			SampleSize:  100,
			Duration:    2 * time.Second,
			Timestamp:   time.Now(),
		}

		// Test baseline comparison
		comparison := benchmarkManager.compareWithBaseline(result, "test_category")
		assert.NotNil(t, comparison)
		assert.True(t, comparison.ExceedsBaseline)
		assert.Greater(t, comparison.ImprovementPercentage, 0.0)
	})

	t.Run("Trend Analysis", func(t *testing.T) {
		// Create time series data
		now := time.Now()
		results := []*success_monitoring.BenchmarkResult{
			{
				SuiteID:     "test-suite",
				SuccessRate: 0.90,
				SampleSize:  100,
				Duration:    2 * time.Second,
				Timestamp:   now.Add(-2 * time.Hour),
			},
			{
				SuiteID:     "test-suite",
				SuccessRate: 0.92,
				SampleSize:  100,
				Duration:    2 * time.Second,
				Timestamp:   now.Add(-1 * time.Hour),
			},
			{
				SuiteID:     "test-suite",
				SuccessRate: 0.95,
				SampleSize:  100,
				Duration:    2 * time.Second,
				Timestamp:   now,
			},
		}

		// Test trend analysis
		trend := benchmarkManager.calculateTrendAnalysis(results)
		assert.NotNil(t, trend)
		assert.Equal(t, "improving", trend.TrendDirection)
		assert.Greater(t, trend.SuccessRateTrend, 0.0)
	})
}

func TestSuccessRateBenchmarkingErrorHandling(t *testing.T) {
	logger := zap.NewNop()
	benchmarkManager := success_monitoring.NewSuccessRateBenchmarkManager(nil, logger)
	benchmarkHandler := handlers.NewSuccessRateBenchmarkingHandler(benchmarkManager, logger)

	router := mux.NewRouter()
	router.HandleFunc("/api/v3/benchmarking/suites", benchmarkHandler.CreateBenchmarkSuite).Methods(http.MethodPost)
	router.HandleFunc("/api/v3/benchmarking/suites/{suiteId}/execute", benchmarkHandler.ExecuteBenchmark).Methods(http.MethodPost)
	router.HandleFunc("/api/v3/benchmarking/baselines", benchmarkHandler.UpdateBaseline).Methods(http.MethodPost)

	t.Run("Invalid Request Body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v3/benchmarking/suites", nil)
		req.Body = httptest.NewRecorder().Result().Body
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid Suite ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v3/benchmarking/suites//execute", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid Baseline Data", func(t *testing.T) {
		invalidRequest := handlers.UpdateBaselineRequest{
			Category:    "",
			SuccessRate: 1.5, // Invalid: > 1.0
			SampleCount: -1,  // Invalid: negative
		}

		invalidJSON, err := json.Marshal(invalidRequest)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/benchmarking/baselines", nil)
		req.Body = httptest.NewRecorder().Result().Body
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v3/benchmarking/suites", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestSuccessRateBenchmarkingPerformance(t *testing.T) {
	logger := zap.NewNop()
	benchmarkManager := success_monitoring.NewSuccessRateBenchmarkManager(nil, logger)

	t.Run("Large Dataset Processing", func(t *testing.T) {
		// Create a large number of benchmark results
		results := make([]*success_monitoring.BenchmarkResult, 1000)
		now := time.Now()

		for i := 0; i < 1000; i++ {
			results[i] = &success_monitoring.BenchmarkResult{
				SuiteID:     fmt.Sprintf("suite-%d", i%10),
				SuccessRate: 0.90 + float64(i%10)*0.01,
				SampleSize:  100,
				Duration:    time.Duration(i%1000) * time.Millisecond,
				Timestamp:   now.Add(time.Duration(-i) * time.Minute),
			}
		}

		// Test performance of processing large dataset
		start := time.Now()
		summary := benchmarkManager.generateSummary(results)
		duration := time.Since(start)

		assert.NotNil(t, summary)
		assert.Less(t, duration, 100*time.Millisecond, "Processing should be fast")
		assert.Equal(t, 1000, summary.TotalResults)
	})

	t.Run("Concurrent Benchmark Execution", func(t *testing.T) {
		// Test concurrent execution of multiple benchmarks
		suite := &success_monitoring.BenchmarkSuite{
			ID:       "concurrent-test",
			Name:     "Concurrent Test Suite",
			Category: "concurrent_testing",
			TestCases: []success_monitoring.BenchmarkTestCase{
				{
					ID:                  "test-1",
					Name:                "Test 1",
					Input:               map[string]interface{}{"test": "data1"},
					ExpectedSuccessRate: 0.95,
					MaxDuration:         1 * time.Second,
				},
				{
					ID:                  "test-2",
					Name:                "Test 2",
					Input:               map[string]interface{}{"test": "data2"},
					ExpectedSuccessRate: 0.95,
					MaxDuration:         1 * time.Second,
				},
			},
			SampleSize:    10,
			MaxIterations: 2,
		}

		ctx := context.Background()
		err := benchmarkManager.CreateBenchmarkSuite(ctx, suite)
		require.NoError(t, err)

		// Execute benchmark and measure performance
		start := time.Now()
		result, err := benchmarkManager.ExecuteBenchmark(ctx, "concurrent-test")
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Less(t, duration, 5*time.Second, "Benchmark execution should complete within reasonable time")
	})
}
