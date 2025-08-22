package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"time"
)

// LoadTestConfig holds configuration for load testing
type LoadTestConfig struct {
	ConcurrentUsers    int           // Number of concurrent users to simulate
	RequestsPerUser    int           // Number of requests per user
	TestDuration       time.Duration // Total test duration
	RampUpTime         time.Duration // Time to ramp up to full load
	RampDownTime       time.Duration // Time to ramp down from full load
	RequestTimeout     time.Duration // Timeout for individual requests
	TargetEndpoint     string        // Target endpoint to test
	RequestPayload     string        // JSON payload for requests
	ExpectedStatusCode int           // Expected HTTP status code
	EnableMetrics      bool          // Enable detailed metrics collection
}

// LoadTestResult holds the results of a load test
type LoadTestResult struct {
	TestConfig          *LoadTestConfig
	StartTime           time.Time
	EndTime             time.Time
	TotalRequests       int64
	SuccessfulRequests  int64
	FailedRequests      int64
	TimeoutRequests     int64
	RateLimitedRequests int64
	AverageResponseTime time.Duration
	MinResponseTime     time.Duration
	MaxResponseTime     time.Duration
	P50ResponseTime     time.Duration
	P95ResponseTime     time.Duration
	P99ResponseTime     time.Duration
	RequestsPerSecond   float64
	ErrorRate           float64
	TimeoutRate         float64
	RateLimitRate       float64
	QueueMetrics        *QueueMetrics
	CapacityAnalysis    *CapacityAnalysis
}

// CapacityAnalysis provides capacity planning insights
type CapacityAnalysis struct {
	MaxConcurrentUsers     int                // Maximum users the system can handle
	OptimalConcurrentUsers int                // Optimal number of concurrent users
	BottleneckType         string             // Type of bottleneck (CPU, Memory, Network, Queue)
	BottleneckSeverity     string             // Severity of bottleneck (Low, Medium, High, Critical)
	Recommendations        []string           // Capacity planning recommendations
	ScalingFactor          float64            // Recommended scaling factor
	ResourceUtilization    map[string]float64 // Resource utilization percentages
}

// LoadTester manages load testing operations
type LoadTester struct {
	config     *LoadTestConfig
	queue      *RequestQueue
	results    []*LoadTestResult
	mu         sync.RWMutex
	httpClient *http.Client
}

// DefaultLoadTestConfig returns default configuration for load testing
func DefaultLoadTestConfig() *LoadTestConfig {
	return &LoadTestConfig{
		ConcurrentUsers:    50,                                                            // Start with 50 concurrent users
		RequestsPerUser:    10,                                                            // 10 requests per user
		TestDuration:       5 * time.Minute,                                               // 5 minute test
		RampUpTime:         30 * time.Second,                                              // 30 second ramp up
		RampDownTime:       30 * time.Second,                                              // 30 second ramp down
		RequestTimeout:     10 * time.Second,                                              // 10 second timeout
		TargetEndpoint:     "/v1/classify",                                                // Test classification endpoint
		RequestPayload:     `{"business_name":"Test Company","location":"United States"}`, // Sample payload
		ExpectedStatusCode: 200,                                                           // Expect 200 OK
		EnableMetrics:      true,                                                          // Enable metrics
	}
}

// NewLoadTester creates a new load tester
func NewLoadTester(config *LoadTestConfig) *LoadTester {
	if config == nil {
		config = DefaultLoadTestConfig()
	}

	return &LoadTester{
		config: config,
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
		},
	}
}

// RunLoadTest executes a comprehensive load test
func (lt *LoadTester) RunLoadTest(handler http.HandlerFunc) (*LoadTestResult, error) {
	log.Printf("Starting load test with %d concurrent users for %v",
		lt.config.ConcurrentUsers, lt.config.TestDuration)

	startTime := time.Now()
	result := &LoadTestResult{
		TestConfig: lt.config,
		StartTime:  startTime,
	}

	// Create a test server
	server := httptest.NewServer(handler)
	defer server.Close()

	// Create channels for collecting results
	responseTimes := make(chan time.Duration, lt.config.ConcurrentUsers*lt.config.RequestsPerUser)
	errors := make(chan error, lt.config.ConcurrentUsers*lt.config.RequestsPerUser)
	statusCodes := make(chan int, lt.config.ConcurrentUsers*lt.config.RequestsPerUser)

	// Start the load test
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), lt.config.TestDuration)
	defer cancel()

	// Calculate ramp up and ramp down intervals
	rampUpInterval := lt.config.RampUpTime / time.Duration(lt.config.ConcurrentUsers)
	rampDownInterval := lt.config.RampDownTime / time.Duration(lt.config.ConcurrentUsers)

	// Start user goroutines with ramp up
	for i := 0; i < lt.config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			// Ramp up delay
			rampUpDelay := time.Duration(userID) * rampUpInterval
			select {
			case <-time.After(rampUpDelay):
			case <-ctx.Done():
				return
			}

			// Send requests for this user
			for j := 0; j < lt.config.RequestsPerUser; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					lt.sendRequest(server.URL+lt.config.TargetEndpoint, responseTimes, errors, statusCodes)
				}
			}

			// Ramp down delay
			rampDownDelay := time.Duration(lt.config.ConcurrentUsers-userID) * rampDownInterval
			select {
			case <-time.After(rampDownDelay):
			case <-ctx.Done():
				return
			}
		}(i)
	}

	// Wait for all users to complete
	wg.Wait()
	close(responseTimes)
	close(errors)
	close(statusCodes)

	// Collect results
	result.EndTime = time.Now()
	result.collectResults(responseTimes, errors, statusCodes)

	// Perform capacity analysis
	result.CapacityAnalysis = lt.analyzeCapacity(result)

	// Store result
	lt.mu.Lock()
	lt.results = append(lt.results, result)
	lt.mu.Unlock()

	log.Printf("Load test completed: %d requests, %.2f RPS, %.2f%% error rate",
		result.TotalRequests, result.RequestsPerSecond, result.ErrorRate*100)

	return result, nil
}

// sendRequest sends a single request and collects metrics
func (lt *LoadTester) sendRequest(url string, responseTimes chan<- time.Duration, errors chan<- error, statusCodes chan<- int) {
	startTime := time.Now()

	// Create request
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		errors <- err
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := lt.httpClient.Do(req)
	if err != nil {
		errors <- err
		return
	}
	defer resp.Body.Close()

	// Record response time
	responseTime := time.Since(startTime)
	responseTimes <- responseTime

	// Record status code
	statusCodes <- resp.StatusCode
}

// collectResults processes all collected metrics
func (result *LoadTestResult) collectResults(responseTimes <-chan time.Duration, errors <-chan error, statusCodes <-chan int) {
	var times []time.Duration
	var totalRequests int64
	var successfulRequests int64
	var failedRequests int64
	var timeoutRequests int64
	var rateLimitedRequests int64

	// Collect response times
	for responseTime := range responseTimes {
		times = append(times, responseTime)
		totalRequests++
	}

	// Collect errors
	for err := range errors {
		failedRequests++
		if err.Error() == "timeout" {
			timeoutRequests++
		}
	}

	// Collect status codes
	for statusCode := range statusCodes {
		if statusCode == 200 {
			successfulRequests++
		} else if statusCode == 429 {
			rateLimitedRequests++
		} else {
			failedRequests++
		}
	}

	// Calculate statistics
	result.TotalRequests = totalRequests
	result.SuccessfulRequests = successfulRequests
	result.FailedRequests = failedRequests
	result.TimeoutRequests = timeoutRequests
	result.RateLimitedRequests = rateLimitedRequests

	// Calculate response time statistics
	if len(times) > 0 {
		result.calculateResponseTimeStats(times)
	}

	// Calculate rates
	duration := result.EndTime.Sub(result.StartTime).Seconds()
	result.RequestsPerSecond = float64(totalRequests) / duration
	result.ErrorRate = float64(failedRequests) / float64(totalRequests)
	result.TimeoutRate = float64(timeoutRequests) / float64(totalRequests)
	result.RateLimitRate = float64(rateLimitedRequests) / float64(totalRequests)
}

// calculateResponseTimeStats calculates response time percentiles
func (result *LoadTestResult) calculateResponseTimeStats(times []time.Duration) {
	if len(times) == 0 {
		return
	}

	// Sort times for percentile calculation
	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})

	result.MinResponseTime = times[0]
	result.MaxResponseTime = times[len(times)-1]

	// Calculate average
	var total time.Duration
	for _, t := range times {
		total += t
	}
	result.AverageResponseTime = total / time.Duration(len(times))

	// Calculate percentiles
	result.P50ResponseTime = times[len(times)*50/100]
	result.P95ResponseTime = times[len(times)*95/100]
	result.P99ResponseTime = times[len(times)*99/100]
}

// analyzeCapacity performs capacity analysis based on test results
func (lt *LoadTester) analyzeCapacity(result *LoadTestResult) *CapacityAnalysis {
	analysis := &CapacityAnalysis{
		ResourceUtilization: make(map[string]float64),
	}

	// Analyze performance characteristics
	if result.ErrorRate > 0.05 { // More than 5% errors
		analysis.BottleneckType = "Queue"
		analysis.BottleneckSeverity = "High"
		analysis.MaxConcurrentUsers = int(float64(lt.config.ConcurrentUsers) * 0.8)
		analysis.Recommendations = append(analysis.Recommendations,
			"Reduce concurrent users or increase queue capacity")
	} else if result.AverageResponseTime > 5*time.Second {
		analysis.BottleneckType = "Processing"
		analysis.BottleneckSeverity = "Medium"
		analysis.MaxConcurrentUsers = int(float64(lt.config.ConcurrentUsers) * 0.9)
		analysis.Recommendations = append(analysis.Recommendations,
			"Optimize request processing or increase worker count")
	} else {
		analysis.BottleneckType = "None"
		analysis.BottleneckSeverity = "Low"
		analysis.MaxConcurrentUsers = int(float64(lt.config.ConcurrentUsers) * 1.2)
		analysis.Recommendations = append(analysis.Recommendations,
			"System performing well, can handle more load")
	}

	// Calculate optimal concurrent users
	analysis.OptimalConcurrentUsers = int(float64(analysis.MaxConcurrentUsers) * 0.8)

	// Calculate scaling factor
	analysis.ScalingFactor = float64(analysis.MaxConcurrentUsers) / float64(lt.config.ConcurrentUsers)

	// Estimate resource utilization
	analysis.ResourceUtilization["CPU"] = lt.estimateCPUUtilization(result)
	analysis.ResourceUtilization["Memory"] = lt.estimateMemoryUtilization(result)
	analysis.ResourceUtilization["Network"] = lt.estimateNetworkUtilization(result)

	// Add capacity planning recommendations
	analysis.addCapacityRecommendations(result)

	return analysis
}

// estimateCPUUtilization estimates CPU utilization based on test results
func (lt *LoadTester) estimateCPUUtilization(result *LoadTestResult) float64 {
	// Simple estimation based on response time and request rate
	baseCPU := 20.0      // Base CPU usage
	cpuPerRequest := 0.1 // CPU per request
	cpuPerRPS := 2.0     // CPU per request per second

	estimatedCPU := baseCPU + (float64(result.TotalRequests) * cpuPerRequest) + (result.RequestsPerSecond * cpuPerRPS)
	if estimatedCPU > 100.0 {
		estimatedCPU = 100.0
	}
	return estimatedCPU
}

// estimateMemoryUtilization estimates memory utilization
func (lt *LoadTester) estimateMemoryUtilization(result *LoadTestResult) float64 {
	// Simple estimation based on concurrent users and queue size
	baseMemory := 30.0        // Base memory usage
	memoryPerUser := 0.5      // Memory per concurrent user
	memoryPerQueueItem := 0.1 // Memory per queued request

	estimatedMemory := baseMemory + (float64(lt.config.ConcurrentUsers) * memoryPerUser) + (float64(result.TotalRequests) * memoryPerQueueItem)
	if estimatedMemory > 100.0 {
		estimatedMemory = 100.0
	}
	return estimatedMemory
}

// estimateNetworkUtilization estimates network utilization
func (lt *LoadTester) estimateNetworkUtilization(result *LoadTestResult) float64 {
	// Simple estimation based on request rate and payload size
	baseNetwork := 10.0       // Base network usage
	networkPerRequest := 0.05 // Network per request

	estimatedNetwork := baseNetwork + (float64(result.TotalRequests) * networkPerRequest)
	if estimatedNetwork > 100.0 {
		estimatedNetwork = 100.0
	}
	return estimatedNetwork
}

// addCapacityRecommendations adds specific capacity planning recommendations
func (analysis *CapacityAnalysis) addCapacityRecommendations(result *LoadTestResult) {
	if result.ErrorRate > 0.1 {
		analysis.Recommendations = append(analysis.Recommendations,
			"Error rate too high - implement circuit breakers and retry logic")
	}

	if result.TimeoutRate > 0.05 {
		analysis.Recommendations = append(analysis.Recommendations,
			"Timeout rate high - increase request timeout or optimize processing")
	}

	if result.RateLimitRate > 0.1 {
		analysis.Recommendations = append(analysis.Recommendations,
			"Rate limiting frequent - increase rate limits or implement better caching")
	}

	if result.AverageResponseTime > 3*time.Second {
		analysis.Recommendations = append(analysis.Recommendations,
			"Response time slow - optimize database queries and external API calls")
	}

	if analysis.ResourceUtilization["CPU"] > 80 {
		analysis.Recommendations = append(analysis.Recommendations,
			"High CPU usage - consider horizontal scaling or code optimization")
	}

	if analysis.ResourceUtilization["Memory"] > 80 {
		analysis.Recommendations = append(analysis.Recommendations,
			"High memory usage - implement memory pooling or increase memory limits")
	}
}

// GetTestHistory returns all previous test results
func (lt *LoadTester) GetTestHistory() []*LoadTestResult {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	// Return a copy to avoid race conditions
	results := make([]*LoadTestResult, len(lt.results))
	copy(results, lt.results)
	return results
}

// GenerateCapacityReport generates a comprehensive capacity planning report
func (lt *LoadTester) GenerateCapacityReport() string {
	history := lt.GetTestHistory()
	if len(history) == 0 {
		return "No load test results available"
	}

	report := fmt.Sprintf("# Capacity Planning Report\n\n")
	report += fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339))

	// Summary statistics
	report += "## Test Summary\n\n"
	report += fmt.Sprintf("- Total Tests: %d\n", len(history))
	report += fmt.Sprintf("- Average RPS: %.2f\n", lt.calculateAverageRPS(history))
	report += fmt.Sprintf("- Average Error Rate: %.2f%%\n", lt.calculateAverageErrorRate(history)*100)
	report += fmt.Sprintf("- Average Response Time: %v\n", lt.calculateAverageResponseTime(history))

	// Latest test analysis
	latest := history[len(history)-1]
	report += "\n## Latest Test Analysis\n\n"
	report += fmt.Sprintf("- Concurrent Users: %d\n", latest.TestConfig.ConcurrentUsers)
	report += fmt.Sprintf("- Requests Per Second: %.2f\n", latest.RequestsPerSecond)
	report += fmt.Sprintf("- Error Rate: %.2f%%\n", latest.ErrorRate*100)
	report += fmt.Sprintf("- Average Response Time: %v\n", latest.AverageResponseTime)
	report += fmt.Sprintf("- P95 Response Time: %v\n", latest.P95ResponseTime)
	report += fmt.Sprintf("- P99 Response Time: %v\n", latest.P99ResponseTime)

	// Capacity analysis
	if latest.CapacityAnalysis != nil {
		report += "\n## Capacity Analysis\n\n"
		report += fmt.Sprintf("- Max Concurrent Users: %d\n", latest.CapacityAnalysis.MaxConcurrentUsers)
		report += fmt.Sprintf("- Optimal Concurrent Users: %d\n", latest.CapacityAnalysis.OptimalConcurrentUsers)
		report += fmt.Sprintf("- Bottleneck Type: %s\n", latest.CapacityAnalysis.BottleneckType)
		report += fmt.Sprintf("- Bottleneck Severity: %s\n", latest.CapacityAnalysis.BottleneckSeverity)
		report += fmt.Sprintf("- Scaling Factor: %.2f\n", latest.CapacityAnalysis.ScalingFactor)

		report += "\n### Resource Utilization\n\n"
		for resource, utilization := range latest.CapacityAnalysis.ResourceUtilization {
			report += fmt.Sprintf("- %s: %.1f%%\n", resource, utilization)
		}

		report += "\n### Recommendations\n\n"
		for i, recommendation := range latest.CapacityAnalysis.Recommendations {
			report += fmt.Sprintf("%d. %s\n", i+1, recommendation)
		}
	}

	return report
}

// calculateAverageRPS calculates average requests per second across all tests
func (lt *LoadTester) calculateAverageRPS(history []*LoadTestResult) float64 {
	if len(history) == 0 {
		return 0
	}

	var total float64
	for _, result := range history {
		total += result.RequestsPerSecond
	}
	return total / float64(len(history))
}

// calculateAverageErrorRate calculates average error rate across all tests
func (lt *LoadTester) calculateAverageErrorRate(history []*LoadTestResult) float64 {
	if len(history) == 0 {
		return 0
	}

	var total float64
	for _, result := range history {
		total += result.ErrorRate
	}
	return total / float64(len(history))
}

// calculateAverageResponseTime calculates average response time across all tests
func (lt *LoadTester) calculateAverageResponseTime(history []*LoadTestResult) time.Duration {
	if len(history) == 0 {
		return 0
	}

	var total time.Duration
	for _, result := range history {
		total += result.AverageResponseTime
	}
	return total / time.Duration(len(history))
}
