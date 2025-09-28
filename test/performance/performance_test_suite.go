package performance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// PerformanceTestSuite provides comprehensive performance testing for all KYB services
type PerformanceTestSuite struct {
	suite.Suite
	baseURL    string
	apiGateway *APIGatewayClient
	httpClient *http.Client
	metrics    *PerformanceMetrics
	config     *PerformanceConfig
}

// PerformanceConfig contains performance test configuration
type PerformanceConfig struct {
	BaseURL                 string
	Timeout                 time.Duration
	MaxResponseTime         time.Duration
	MaxResponseTimeCritical time.Duration
	MinThroughput           int
	MaxConcurrentRequests   int
	LoadTestDuration        time.Duration
	RampUpDuration          time.Duration
	RampDownDuration        time.Duration
}

// PerformanceMetrics tracks performance metrics during tests
type PerformanceMetrics struct {
	ResponseTimes      []time.Duration
	Throughput         float64
	ErrorRate          float64
	ConcurrentUsers    int
	TotalRequests      int
	SuccessfulRequests int
	FailedRequests     int
	StartTime          time.Time
	EndTime            time.Time
	mu                 sync.RWMutex
}

// LoadTestResult represents the result of a load test
type LoadTestResult struct {
	TestName            string                 `json:"test_name"`
	Duration            time.Duration          `json:"duration"`
	TotalRequests       int                    `json:"total_requests"`
	SuccessfulRequests  int                    `json:"successful_requests"`
	FailedRequests      int                    `json:"failed_requests"`
	AverageResponseTime time.Duration          `json:"average_response_time"`
	MinResponseTime     time.Duration          `json:"min_response_time"`
	MaxResponseTime     time.Duration          `json:"max_response_time"`
	P95ResponseTime     time.Duration          `json:"p95_response_time"`
	P99ResponseTime     time.Duration          `json:"p99_response_time"`
	Throughput          float64                `json:"throughput"`
	ErrorRate           float64                `json:"error_rate"`
	ConcurrentUsers     int                    `json:"concurrent_users"`
	Status              string                 `json:"status"`
	Details             map[string]interface{} `json:"details"`
}

// BusinessVerificationRequest represents a business verification request
type BusinessVerificationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	Industry    string `json:"industry"`
	Website     string `json:"website,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Email       string `json:"email,omitempty"`
}

// HTTPResponse represents an HTTP response
type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
	Duration   time.Duration
}

// SetupSuite initializes the performance test suite
func (suite *PerformanceTestSuite) SetupSuite() {
	// Load performance test configuration
	suite.config = suite.loadPerformanceConfig()

	// Initialize HTTP client with appropriate timeouts
	suite.httpClient = &http.Client{
		Timeout: suite.config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Initialize API Gateway client
	suite.apiGateway = &APIGatewayClient{
		baseURL: suite.config.BaseURL,
		client:  suite.httpClient,
	}

	// Initialize performance metrics
	suite.metrics = &PerformanceMetrics{
		ResponseTimes: make([]time.Duration, 0),
		StartTime:     time.Now(),
	}

	suite.baseURL = suite.config.BaseURL
}

// loadPerformanceConfig loads performance test configuration
func (suite *PerformanceTestSuite) loadPerformanceConfig() *PerformanceConfig {
	config := &PerformanceConfig{
		BaseURL:                 getEnv("PERF_TEST_BASE_URL", "https://kyb-api-gateway-production.up.railway.app"),
		Timeout:                 30 * time.Second,
		MaxResponseTime:         1 * time.Second,
		MaxResponseTimeCritical: 5 * time.Second,
		MinThroughput:           100, // requests per second
		MaxConcurrentRequests:   100,
		LoadTestDuration:        5 * time.Minute,
		RampUpDuration:          1 * time.Minute,
		RampDownDuration:        1 * time.Minute,
	}

	return config
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestResponseTimeRequirements tests response time requirements
func (suite *PerformanceTestSuite) TestResponseTimeRequirements() {
	suite.Run("Single_Request_Response_Time", func() {
		testBusiness := BusinessVerificationRequest{
			Name:        "Performance Test Business",
			Description: "A test business for performance testing",
			Address:     "789 Performance Street, Test City, TC 12345",
			Industry:    "Technology",
		}

		start := time.Now()
		resp, err := suite.apiGateway.VerifyBusiness(testBusiness)
		duration := time.Since(start)

		require.NoError(suite.T(), err, "Performance test request failed")
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "Performance test returned non-200 status")
		assert.Less(suite.T(), duration, suite.config.MaxResponseTime,
			"Response time %v should be less than %v", duration, suite.config.MaxResponseTime)

		// Log performance metrics
		suite.T().Logf("Response time: %v", duration)
		suite.T().Logf("Status code: %d", resp.StatusCode)
	})

	suite.Run("Multiple_Requests_Response_Time", func() {
		testBusinesses := []BusinessVerificationRequest{
			{
				Name:        "Performance Test Business 1",
				Description: "A test business for performance testing",
				Address:     "789 Performance Street 1, Test City, TC 12345",
				Industry:    "Technology",
			},
			{
				Name:        "Performance Test Business 2",
				Description: "A test business for performance testing",
				Address:     "789 Performance Street 2, Test City, TC 12345",
				Industry:    "Restaurant",
			},
			{
				Name:        "Performance Test Business 3",
				Description: "A test business for performance testing",
				Address:     "789 Performance Street 3, Test City, TC 12345",
				Industry:    "Retail",
			},
		}

		var totalDuration time.Duration
		var successCount int

		for i, business := range testBusinesses {
			start := time.Now()
			resp, err := suite.apiGateway.VerifyBusiness(business)
			duration := time.Since(start)

			if err == nil && resp.StatusCode == http.StatusOK {
				successCount++
				totalDuration += duration
			}

			assert.Less(suite.T(), duration, suite.config.MaxResponseTime,
				"Request %d response time %v should be less than %v", i+1, duration, suite.config.MaxResponseTime)
		}

		if successCount > 0 {
			avgDuration := totalDuration / time.Duration(successCount)
			suite.T().Logf("Average response time: %v", avgDuration)
			suite.T().Logf("Success rate: %d/%d", successCount, len(testBusinesses))
		}
	})
}

// TestConcurrentRequests tests concurrent request handling
func (suite *PerformanceTestSuite) TestConcurrentRequests() {
	suite.Run("Low_Concurrency", func() {
		suite.runConcurrencyTest(10, "Low_Concurrency")
	})

	suite.Run("Medium_Concurrency", func() {
		suite.runConcurrencyTest(50, "Medium_Concurrency")
	})

	suite.Run("High_Concurrency", func() {
		suite.runConcurrencyTest(100, "High_Concurrency")
	})
}

// runConcurrencyTest runs a concurrency test with specified number of concurrent requests
func (suite *PerformanceTestSuite) runConcurrencyTest(concurrency int, testName string) {
	suite.T().Logf("Running %s test with %d concurrent requests", testName, concurrency)

	results := make(chan *HTTPResponse, concurrency)
	errors := make(chan error, concurrency)

	start := time.Now()

	// Launch concurrent requests
	for i := 0; i < concurrency; i++ {
		go func(index int) {
			testBusiness := BusinessVerificationRequest{
				Name:        fmt.Sprintf("Concurrent Test Business %d", index),
				Description: "A test business for concurrent testing",
				Address:     fmt.Sprintf("%d Concurrent Street, Test City, TC 12345", index),
				Industry:    "Technology",
			}

			resp, err := suite.apiGateway.VerifyBusiness(testBusiness)
			if err != nil {
				errors <- err
			} else {
				results <- resp
			}
		}(i)
	}

	// Collect results
	var successfulRequests int
	var failedRequests int
	var totalResponseTime time.Duration
	var maxResponseTime time.Duration
	var minResponseTime time.Duration = time.Hour // Initialize with large value

	for i := 0; i < concurrency; i++ {
		select {
		case resp := <-results:
			successfulRequests++
			if resp.Duration > maxResponseTime {
				maxResponseTime = resp.Duration
			}
			if resp.Duration < minResponseTime {
				minResponseTime = resp.Duration
			}
			totalResponseTime += resp.Duration

		case err := <-errors:
			failedRequests++
			suite.T().Logf("Request failed: %v", err)
		}
	}

	totalDuration := time.Since(start)

	// Calculate metrics
	throughput := float64(successfulRequests) / totalDuration.Seconds()
	errorRate := float64(failedRequests) / float64(concurrency) * 100
	avgResponseTime := totalResponseTime / time.Duration(successfulRequests)

	// Log results
	suite.T().Logf("Concurrency Test Results:")
	suite.T().Logf("  Concurrent Requests: %d", concurrency)
	suite.T().Logf("  Successful Requests: %d", successfulRequests)
	suite.T().Logf("  Failed Requests: %d", failedRequests)
	suite.T().Logf("  Total Duration: %v", totalDuration)
	suite.T().Logf("  Average Response Time: %v", avgResponseTime)
	suite.T().Logf("  Min Response Time: %v", minResponseTime)
	suite.T().Logf("  Max Response Time: %v", maxResponseTime)
	suite.T().Logf("  Throughput: %.2f req/s", throughput)
	suite.T().Logf("  Error Rate: %.2f%%", errorRate)

	// Assertions
	assert.Greater(suite.T(), successfulRequests, 0, "Should have at least one successful request")
	assert.Less(suite.T(), errorRate, 10.0, "Error rate should be less than 10%%")
	assert.Less(suite.T(), avgResponseTime, suite.config.MaxResponseTime,
		"Average response time should be less than %v", suite.config.MaxResponseTime)
	assert.Greater(suite.T(), throughput, float64(suite.config.MinThroughput),
		"Throughput should be greater than %d req/s", suite.config.MinThroughput)
}

// TestLoadTesting tests system under sustained load
func (suite *PerformanceTestSuite) TestLoadTesting() {
	suite.Run("Sustained_Load", func() {
		suite.runLoadTest(50, 2*time.Minute, "Sustained_Load")
	})

	suite.Run("Peak_Load", func() {
		suite.runLoadTest(100, 1*time.Minute, "Peak_Load")
	})
}

// runLoadTest runs a sustained load test
func (suite *PerformanceTestSuite) runLoadTest(concurrency int, duration time.Duration, testName string) {
	suite.T().Logf("Running %s test with %d concurrent users for %v", testName, concurrency, duration)

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	results := make(chan *HTTPResponse, concurrency*10)
	errors := make(chan error, concurrency*10)

	// Start load test
	start := time.Now()

	// Launch workers
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			requestCount := 0
			for {
				select {
				case <-ctx.Done():
					return
				default:
					testBusiness := BusinessVerificationRequest{
						Name:        fmt.Sprintf("Load Test Business %d-%d", workerID, requestCount),
						Description: "A test business for load testing",
						Address:     fmt.Sprintf("%d Load Test Street, Test City, TC 12345", workerID),
						Industry:    "Technology",
					}

					resp, err := suite.apiGateway.VerifyBusiness(testBusiness)
					if err != nil {
						errors <- err
					} else {
						results <- resp
					}

					requestCount++

					// Small delay to prevent overwhelming the system
					time.Sleep(100 * time.Millisecond)
				}
			}
		}(i)
	}

	// Wait for load test to complete
	wg.Wait()

	totalDuration := time.Since(start)

	// Collect results
	var successfulRequests int
	var failedRequests int
	var totalResponseTime time.Duration
	var maxResponseTime time.Duration
	var minResponseTime time.Duration = time.Hour

	// Process results
	close(results)
	close(errors)

	for resp := range results {
		successfulRequests++
		if resp.Duration > maxResponseTime {
			maxResponseTime = resp.Duration
		}
		if resp.Duration < minResponseTime {
			minResponseTime = resp.Duration
		}
		totalResponseTime += resp.Duration
	}

	for range errors {
		failedRequests++
	}

	// Calculate metrics
	throughput := float64(successfulRequests) / totalDuration.Seconds()
	errorRate := float64(failedRequests) / float64(successfulRequests+failedRequests) * 100
	avgResponseTime := totalResponseTime / time.Duration(successfulRequests)

	// Log results
	suite.T().Logf("Load Test Results:")
	suite.T().Logf("  Test Name: %s", testName)
	suite.T().Logf("  Duration: %v", totalDuration)
	suite.T().Logf("  Concurrent Users: %d", concurrency)
	suite.T().Logf("  Total Requests: %d", successfulRequests+failedRequests)
	suite.T().Logf("  Successful Requests: %d", successfulRequests)
	suite.T().Logf("  Failed Requests: %d", failedRequests)
	suite.T().Logf("  Average Response Time: %v", avgResponseTime)
	suite.T().Logf("  Min Response Time: %v", minResponseTime)
	suite.T().Logf("  Max Response Time: %v", maxResponseTime)
	suite.T().Logf("  Throughput: %.2f req/s", throughput)
	suite.T().Logf("  Error Rate: %.2f%%", errorRate)

	// Assertions
	assert.Greater(suite.T(), successfulRequests, 0, "Should have at least one successful request")
	assert.Less(suite.T(), errorRate, 5.0, "Error rate should be less than 5%%")
	assert.Less(suite.T(), avgResponseTime, suite.config.MaxResponseTime,
		"Average response time should be less than %v", suite.config.MaxResponseTime)
	assert.Greater(suite.T(), throughput, float64(suite.config.MinThroughput),
		"Throughput should be greater than %d req/s", suite.config.MinThroughput)
}

// TestStressTesting tests system under stress conditions
func (suite *PerformanceTestSuite) TestStressTesting() {
	suite.Run("Stress_Test", func() {
		suite.runStressTest(200, 30*time.Second, "Stress_Test")
	})
}

// runStressTest runs a stress test with high concurrency
func (suite *PerformanceTestSuite) runStressTest(concurrency int, duration time.Duration, testName string) {
	suite.T().Logf("Running %s with %d concurrent users for %v", testName, concurrency, duration)

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	results := make(chan *HTTPResponse, concurrency*5)
	errors := make(chan error, concurrency*5)

	start := time.Now()

	// Launch stress test workers
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					testBusiness := BusinessVerificationRequest{
						Name:        fmt.Sprintf("Stress Test Business %d", workerID),
						Description: "A test business for stress testing",
						Address:     fmt.Sprintf("%d Stress Street, Test City, TC 12345", workerID),
						Industry:    "Technology",
					}

					resp, err := suite.apiGateway.VerifyBusiness(testBusiness)
					if err != nil {
						errors <- err
					} else {
						results <- resp
					}

					// No delay for stress testing
				}
			}
		}(i)
	}

	wg.Wait()

	totalDuration := time.Since(start)

	// Collect results
	var successfulRequests int
	var failedRequests int
	var totalResponseTime time.Duration
	var maxResponseTime time.Duration

	close(results)
	close(errors)

	for resp := range results {
		successfulRequests++
		if resp.Duration > maxResponseTime {
			maxResponseTime = resp.Duration
		}
		totalResponseTime += resp.Duration
	}

	for range errors {
		failedRequests++
	}

	// Calculate metrics
	throughput := float64(successfulRequests) / totalDuration.Seconds()
	errorRate := float64(failedRequests) / float64(successfulRequests+failedRequests) * 100
	avgResponseTime := totalResponseTime / time.Duration(successfulRequests)

	// Log results
	suite.T().Logf("Stress Test Results:")
	suite.T().Logf("  Test Name: %s", testName)
	suite.T().Logf("  Duration: %v", totalDuration)
	suite.T().Logf("  Concurrent Users: %d", concurrency)
	suite.T().Logf("  Total Requests: %d", successfulRequests+failedRequests)
	suite.T().Logf("  Successful Requests: %d", successfulRequests)
	suite.T().Logf("  Failed Requests: %d", failedRequests)
	suite.T().Logf("  Average Response Time: %v", avgResponseTime)
	suite.T().Logf("  Max Response Time: %v", maxResponseTime)
	suite.T().Logf("  Throughput: %.2f req/s", throughput)
	suite.T().Logf("  Error Rate: %.2f%%", errorRate)

	// Stress test assertions are more lenient
	assert.Greater(suite.T(), successfulRequests, 0, "Should have at least one successful request")
	assert.Less(suite.T(), errorRate, 20.0, "Error rate should be less than 20%% under stress")
	assert.Less(suite.T(), avgResponseTime, suite.config.MaxResponseTimeCritical,
		"Average response time should be less than %v under stress", suite.config.MaxResponseTimeCritical)
}

// TestMemoryUsage tests memory usage under load
func (suite *PerformanceTestSuite) TestMemoryUsage() {
	suite.Run("Memory_Usage_Under_Load", func() {
		// This test would typically require monitoring tools
		// For now, we'll just ensure the system doesn't crash under load

		concurrency := 50
		duration := 1 * time.Minute

		ctx, cancel := context.WithTimeout(context.Background(), duration)
		defer cancel()

		var wg sync.WaitGroup
		successCount := 0
		var mu sync.Mutex

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for {
					select {
					case <-ctx.Done():
						return
					default:
						testBusiness := BusinessVerificationRequest{
							Name:        fmt.Sprintf("Memory Test Business %d", workerID),
							Description: "A test business for memory testing",
							Address:     fmt.Sprintf("%d Memory Street, Test City, TC 12345", workerID),
							Industry:    "Technology",
						}

						_, err := suite.apiGateway.VerifyBusiness(testBusiness)
						if err == nil {
							mu.Lock()
							successCount++
							mu.Unlock()
						}
					}
				}
			}(i)
		}

		wg.Wait()

		suite.T().Logf("Memory test completed with %d successful requests", successCount)
		assert.Greater(suite.T(), successCount, 0, "Should have at least one successful request")
	})
}

// APIGatewayClient represents the API Gateway client for testing
type APIGatewayClient struct {
	baseURL string
	client  *http.Client
}

// VerifyBusiness sends a business verification request
func (c *APIGatewayClient) VerifyBusiness(business BusinessVerificationRequest) (*HTTPResponse, error) {
	jsonData, err := json.Marshal(business)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	req, err := http.NewRequest("POST", c.baseURL+"/verify", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
		Duration:   duration,
	}, nil
}

// TestPerformanceSuite runs the complete performance test suite
func TestPerformanceSuite(t *testing.T) {
	suite.Run(t, new(PerformanceTestSuite))
}
