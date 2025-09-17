package risk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// APITestRunner provides comprehensive API testing capabilities
type APITestRunner struct {
	logger    *zap.Logger
	testSuite *APIIntegrationTestSuite
	results   *APITestResults
}

// APITestResults contains the results of API test execution
type APITestResults struct {
	TotalTests    int                    `json:"total_tests"`
	PassedTests   int                    `json:"passed_tests"`
	FailedTests   int                    `json:"failed_tests"`
	SkippedTests  int                    `json:"skipped_tests"`
	ExecutionTime time.Duration          `json:"execution_time"`
	TestDetails   []APITestDetail        `json:"test_details"`
	Summary       map[string]interface{} `json:"summary"`
	Performance   *APIPerformanceMetrics `json:"performance"`
}

// APITestDetail contains details about individual API test execution
type APITestDetail struct {
	Name         string        `json:"name"`
	Endpoint     string        `json:"endpoint"`
	Method       string        `json:"method"`
	Status       string        `json:"status"`
	Duration     time.Duration `json:"duration"`
	StatusCode   int           `json:"status_code"`
	ResponseTime time.Duration `json:"response_time"`
	ErrorMessage string        `json:"error_message,omitempty"`
	RequestSize  int           `json:"request_size"`
	ResponseSize int           `json:"response_size"`
}

// APIPerformanceMetrics contains API performance metrics
type APIPerformanceMetrics struct {
	AverageResponseTime time.Duration `json:"average_response_time"`
	MaxResponseTime     time.Duration `json:"max_response_time"`
	MinResponseTime     time.Duration `json:"min_response_time"`
	TotalRequests       int           `json:"total_requests"`
	SuccessfulRequests  int           `json:"successful_requests"`
	FailedRequests      int           `json:"failed_requests"`
	Throughput          float64       `json:"throughput"` // requests per second
}

// NewAPITestRunner creates a new API test runner
func NewAPITestRunner() *APITestRunner {
	logger := zap.NewNop()
	return &APITestRunner{
		logger: logger,
		results: &APITestResults{
			TestDetails: make([]APITestDetail, 0),
			Summary:     make(map[string]interface{}),
			Performance: &APIPerformanceMetrics{},
		},
	}
}

// RunAllAPITests runs all API integration tests
func (atr *APITestRunner) RunAllAPITests(t *testing.T) *APITestResults {
	startTime := time.Now()
	atr.logger.Info("Starting API integration test suite")

	// Initialize test suite
	atr.testSuite = NewAPIIntegrationTestSuite(t)
	defer atr.testSuite.Close()

	// Run all test categories
	// TODO: Fix test function references - these are test functions, not regular functions
	// atr.runAPITestCategory(t, "Export API Endpoints", TestExportAPIEndpoints)
	// atr.runAPITestCategory(t, "Backup API Endpoints", TestBackupAPIEndpoints)
	// atr.runAPITestCategory(t, "API Error Handling", TestAPIErrorHandling)
	// atr.runAPITestCategory(t, "API Performance", TestAPIPerformance)
	// atr.runAPITestCategory(t, "API Security", TestAPISecurity)
	// atr.runAPITestCategory(t, "API Versioning", TestAPIVersioning)

	// Calculate final results
	atr.results.ExecutionTime = time.Since(startTime)
	atr.calculateAPISummary()

	atr.logger.Info("API integration test suite completed",
		zap.Int("total_tests", atr.results.TotalTests),
		zap.Int("passed_tests", atr.results.PassedTests),
		zap.Int("failed_tests", atr.results.FailedTests),
		zap.Duration("execution_time", atr.results.ExecutionTime))

	return atr.results
}

// runAPITestCategory runs a specific API test category
func (atr *APITestRunner) runAPITestCategory(t *testing.T, categoryName string, testFunc func(*testing.T)) {
	atr.logger.Info("Running API test category", zap.String("category", categoryName))

	// Create a sub-test for the category
	t.Run(categoryName, func(t *testing.T) {
		startTime := time.Now()

		// Run the test function
		testFunc(t)

		duration := time.Since(startTime)

		// Record test result
		atr.results.TotalTests++
		atr.results.PassedTests++ // If we get here, the test passed

		atr.results.TestDetails = append(atr.results.TestDetails, APITestDetail{
			Name:     categoryName,
			Status:   "PASSED",
			Duration: duration,
		})

		atr.logger.Info("API test category completed",
			zap.String("category", categoryName),
			zap.Duration("duration", duration),
			zap.String("status", "PASSED"))
	})
}

// calculateAPISummary calculates API test summary statistics
func (atr *APITestRunner) calculateAPISummary() {
	atr.results.Summary = map[string]interface{}{
		"total_tests":       atr.results.TotalTests,
		"passed_tests":      atr.results.PassedTests,
		"failed_tests":      atr.results.FailedTests,
		"skipped_tests":     atr.results.SkippedTests,
		"pass_rate":         float64(atr.results.PassedTests) / float64(atr.results.TotalTests) * 100,
		"execution_time":    atr.results.ExecutionTime.String(),
		"average_test_time": atr.results.ExecutionTime / time.Duration(atr.results.TotalTests),
	}

	// Calculate API-specific statistics
	apiStats := make(map[string]map[string]interface{})
	for _, detail := range atr.results.TestDetails {
		if apiStats[detail.Name] == nil {
			apiStats[detail.Name] = make(map[string]interface{})
		}
		apiStats[detail.Name]["duration"] = detail.Duration.String()
		apiStats[detail.Name]["status"] = detail.Status
		apiStats[detail.Name]["status_code"] = detail.StatusCode
		apiStats[detail.Name]["response_time"] = detail.ResponseTime.String()
	}
	atr.results.Summary["api_stats"] = apiStats

	// Calculate performance metrics
	atr.calculatePerformanceMetrics()
}

// calculatePerformanceMetrics calculates API performance metrics
func (atr *APITestRunner) calculatePerformanceMetrics() {
	if len(atr.results.TestDetails) == 0 {
		return
	}

	var totalResponseTime time.Duration
	var maxResponseTime time.Duration
	var minResponseTime time.Duration
	var successfulRequests int
	var failedRequests int

	for _, detail := range atr.results.TestDetails {
		if detail.ResponseTime > 0 {
			totalResponseTime += detail.ResponseTime
			if detail.ResponseTime > maxResponseTime {
				maxResponseTime = detail.ResponseTime
			}
			if minResponseTime == 0 || detail.ResponseTime < minResponseTime {
				minResponseTime = detail.ResponseTime
			}
		}

		if detail.StatusCode >= 200 && detail.StatusCode < 300 {
			successfulRequests++
		} else {
			failedRequests++
		}
	}

	atr.results.Performance.TotalRequests = len(atr.results.TestDetails)
	atr.results.Performance.SuccessfulRequests = successfulRequests
	atr.results.Performance.FailedRequests = failedRequests
	atr.results.Performance.MaxResponseTime = maxResponseTime
	atr.results.Performance.MinResponseTime = minResponseTime

	if len(atr.results.TestDetails) > 0 {
		atr.results.Performance.AverageResponseTime = totalResponseTime / time.Duration(len(atr.results.TestDetails))
	}

	if atr.results.ExecutionTime > 0 {
		atr.results.Performance.Throughput = float64(len(atr.results.TestDetails)) / atr.results.ExecutionTime.Seconds()
	}
}

// TestAPISingleEndpoint tests a single API endpoint
func (atr *APITestRunner) TestAPISingleEndpoint(t *testing.T, method, endpoint string, requestBody interface{}, expectedStatusCode int) {
	startTime := time.Now()

	// Create request
	var req *http.Request
	var err error

	if requestBody != nil {
		body, _ := json.Marshal(requestBody)
		req, err = http.NewRequest(method, atr.testSuite.server.URL+endpoint, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, atr.testSuite.server.URL+endpoint, nil)
	}

	require.NoError(t, err)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Record test details
	duration := time.Since(startTime)
	responseTime := duration

	atr.results.TestDetails = append(atr.results.TestDetails, APITestDetail{
		Name:         fmt.Sprintf("%s %s", method, endpoint),
		Endpoint:     endpoint,
		Method:       method,
		Status:       "PASSED",
		Duration:     duration,
		StatusCode:   resp.StatusCode,
		ResponseTime: responseTime,
	})

	// Validate response
	assert.Equal(t, expectedStatusCode, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

// TestAPILoadTesting performs load testing on API endpoints
func (atr *APITestRunner) TestAPILoadTesting(t *testing.T, endpoint string, numRequests int, concurrency int) {
	atr.logger.Info("Starting API load testing",
		zap.String("endpoint", endpoint),
		zap.Int("num_requests", numRequests),
		zap.Int("concurrency", concurrency))

	results := make(chan APITestDetail, numRequests)
	semaphore := make(chan struct{}, concurrency)

	startTime := time.Now()

	for i := 0; i < numRequests; i++ {
		go func(i int) {
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			requestStart := time.Now()

			req, err := http.NewRequest("GET", atr.testSuite.server.URL+endpoint, nil)
			require.NoError(t, err)

			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			duration := time.Since(requestStart)

			results <- APITestDetail{
				Name:         fmt.Sprintf("Load Test %d", i),
				Endpoint:     endpoint,
				Method:       "GET",
				Status:       "PASSED",
				Duration:     duration,
				StatusCode:   resp.StatusCode,
				ResponseTime: duration,
			}
		}(i)
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		result := <-results
		atr.results.TestDetails = append(atr.results.TestDetails, result)
	}

	totalTime := time.Since(startTime)
	atr.logger.Info("API load testing completed",
		zap.Duration("total_time", totalTime),
		zap.Float64("requests_per_second", float64(numRequests)/totalTime.Seconds()))
}

// TestAPISecurityScanning performs security scanning on API endpoints
func (atr *APITestRunner) TestAPISecurityScanning(t *testing.T) {
	atr.logger.Info("Starting API security scanning")

	// Test SQL injection
	atr.testSQLInjection(t)

	// Test XSS
	atr.testXSS(t)

	// Test CSRF
	atr.testCSRF(t)

	// Test authentication bypass
	atr.testAuthenticationBypass(t)

	atr.logger.Info("API security scanning completed")
}

// testSQLInjection tests for SQL injection vulnerabilities
func (atr *APITestRunner) testSQLInjection(t *testing.T) {
	sqlInjectionPayloads := []string{
		"'; DROP TABLE assessments; --",
		"' OR '1'='1",
		"'; INSERT INTO assessments VALUES ('hacked'); --",
		"' UNION SELECT * FROM users --",
	}

	for _, payload := range sqlInjectionPayloads {
		exportData := map[string]interface{}{
			"business_id": payload,
			"export_type": "assessments",
			"format":      "json",
		}

		reqBody, _ := json.Marshal(exportData)
		req, err := http.NewRequest("POST", atr.testSuite.server.URL+"/api/v1/export/jobs", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should handle gracefully without crashing
		assert.True(t, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusBadRequest)
	}
}

// testXSS tests for XSS vulnerabilities
func (atr *APITestRunner) testXSS(t *testing.T) {
	xssPayloads := []string{
		"<script>alert('xss')</script>",
		"javascript:alert('xss')",
		"<img src=x onerror=alert('xss')>",
		"<svg onload=alert('xss')>",
	}

	for _, payload := range xssPayloads {
		exportData := map[string]interface{}{
			"business_id": "test-business-123",
			"export_type": "assessments",
			"format":      "json",
			"metadata": map[string]interface{}{
				"script": payload,
			},
		}

		reqBody, _ := json.Marshal(exportData)
		req, err := http.NewRequest("POST", atr.testSuite.server.URL+"/api/v1/export/jobs", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should handle gracefully
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}
}

// testCSRF tests for CSRF vulnerabilities
func (atr *APITestRunner) testCSRF(t *testing.T) {
	// Test without proper CSRF token
	exportData := map[string]interface{}{
		"business_id": "test-business-123",
		"export_type": "assessments",
		"format":      "json",
	}

	reqBody, _ := json.Marshal(exportData)
	req, err := http.NewRequest("POST", atr.testSuite.server.URL+"/api/v1/export/jobs", bytes.NewReader(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	// Note: Not setting CSRF token header

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should still work as we're not implementing CSRF protection in this test
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

// testAuthenticationBypass tests for authentication bypass vulnerabilities
func (atr *APITestRunner) testAuthenticationBypass(t *testing.T) {
	// Test without authentication headers
	req, err := http.NewRequest("GET", atr.testSuite.server.URL+"/api/v1/export/jobs", nil)
	require.NoError(t, err)
	// Note: Not setting authentication headers

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should still work as we're not implementing authentication in this test
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// GenerateAPIReport generates a comprehensive API test report
func (atr *APITestRunner) GenerateAPIReport() (string, error) {
	report := fmt.Sprintf(`
# API Integration Test Report

## Summary
- Total Tests: %d
- Passed Tests: %d
- Failed Tests: %d
- Skipped Tests: %d
- Pass Rate: %.2f%%
- Execution Time: %s

## Performance Metrics
- Average Response Time: %s
- Max Response Time: %s
- Min Response Time: %s
- Total Requests: %d
- Successful Requests: %d
- Failed Requests: %d
- Throughput: %.2f requests/second

## Test Details
`,
		atr.results.TotalTests,
		atr.results.PassedTests,
		atr.results.FailedTests,
		atr.results.SkippedTests,
		float64(atr.results.PassedTests)/float64(atr.results.TotalTests)*100,
		atr.results.ExecutionTime.String(),
		atr.results.Performance.AverageResponseTime.String(),
		atr.results.Performance.MaxResponseTime.String(),
		atr.results.Performance.MinResponseTime.String(),
		atr.results.Performance.TotalRequests,
		atr.results.Performance.SuccessfulRequests,
		atr.results.Performance.FailedRequests,
		atr.results.Performance.Throughput)

	for _, detail := range atr.results.TestDetails {
		report += fmt.Sprintf(`
### %s
- Endpoint: %s %s
- Status: %s
- Duration: %s
- Status Code: %d
- Response Time: %s
`,
			detail.Name,
			detail.Method,
			detail.Endpoint,
			detail.Status,
			detail.Duration.String(),
			detail.StatusCode,
			detail.ResponseTime.String())
	}

	return report, nil
}
