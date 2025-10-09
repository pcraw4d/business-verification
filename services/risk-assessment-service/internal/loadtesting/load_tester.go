package loadtesting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// LoadTester performs load testing on the risk assessment service
type LoadTester struct {
	logger     *zap.Logger
	baseURL    string
	httpClient *http.Client
}

// LoadTestConfig represents load test configuration
type LoadTestConfig struct {
	Duration        time.Duration `json:"duration"`
	ConcurrentUsers int           `json:"concurrent_users"`
	RequestsPerUser int           `json:"requests_per_user"`
	RampUpTime      time.Duration `json:"ramp_up_time"`
	TargetRPS       float64       `json:"target_rps"`
	Timeout         time.Duration `json:"timeout"`
}

// LoadTestResult represents the result of a load test
type LoadTestResult struct {
	TotalRequests       int64           `json:"total_requests"`
	SuccessfulRequests  int64           `json:"successful_requests"`
	FailedRequests      int64           `json:"failed_requests"`
	TotalDuration       time.Duration   `json:"total_duration"`
	AverageResponseTime time.Duration   `json:"average_response_time"`
	MinResponseTime     time.Duration   `json:"min_response_time"`
	MaxResponseTime     time.Duration   `json:"max_response_time"`
	RequestsPerSecond   float64         `json:"requests_per_second"`
	RequestsPerMinute   float64         `json:"requests_per_minute"`
	ErrorRate           float64         `json:"error_rate"`
	ResponseTimes       []time.Duration `json:"response_times"`
	Errors              []LoadTestError `json:"errors"`
	Timestamp           time.Time       `json:"timestamp"`
}

// LoadTestError represents an error during load testing
type LoadTestError struct {
	RequestID string        `json:"request_id"`
	Error     string        `json:"error"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
}

// LoadTestRequest represents a single load test request
type LoadTestRequest struct {
	ID        string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Success   bool
	Error     error
	Response  *http.Response
}

// NewLoadTester creates a new load tester
func NewLoadTester(logger *zap.Logger, baseURL string) *LoadTester {
	return &LoadTester{
		logger:  logger,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RunLoadTest runs a comprehensive load test
func (lt *LoadTester) RunLoadTest(ctx context.Context, config LoadTestConfig) (*LoadTestResult, error) {
	lt.logger.Info("Starting load test",
		zap.Duration("duration", config.Duration),
		zap.Int("concurrent_users", config.ConcurrentUsers),
		zap.Int("requests_per_user", config.RequestsPerUser),
		zap.Float64("target_rps", config.TargetRPS))

	// Create channels for coordination
	requestChan := make(chan LoadTestRequest, config.ConcurrentUsers*config.RequestsPerUser)
	resultChan := make(chan LoadTestResult, 1)

	// Start result collector
	go lt.collectResults(requestChan, resultChan, config)

	// Start load test
	lt.startLoadTest(ctx, config, requestChan)

	// Wait for completion
	select {
	case result := <-resultChan:
		lt.logger.Info("Load test completed",
			zap.Int64("total_requests", result.TotalRequests),
			zap.Int64("successful_requests", result.SuccessfulRequests),
			zap.Float64("requests_per_second", result.RequestsPerSecond),
			zap.Float64("error_rate", result.ErrorRate))
		return &result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// startLoadTest starts the actual load test
func (lt *LoadTester) startLoadTest(ctx context.Context, config LoadTestConfig, requestChan chan<- LoadTestRequest) {
	var wg sync.WaitGroup

	// Calculate ramp-up delay
	rampUpDelay := time.Duration(0)
	if config.RampUpTime > 0 && config.ConcurrentUsers > 1 {
		rampUpDelay = config.RampUpTime / time.Duration(config.ConcurrentUsers-1)
	}

	// Start concurrent users
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			// Ramp up delay
			if userID > 0 {
				time.Sleep(rampUpDelay)
			}

			// Send requests for this user
			for j := 0; j < config.RequestsPerUser; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					request := lt.sendRequest(ctx, userID, j)
					requestChan <- request
				}
			}
		}(i)
	}

	// Wait for all users to complete
	go func() {
		wg.Wait()
		close(requestChan)
	}()
}

// sendRequest sends a single request
func (lt *LoadTester) sendRequest(ctx context.Context, userID, requestID int) LoadTestRequest {
	startTime := time.Now()
	request := LoadTestRequest{
		ID:        fmt.Sprintf("user_%d_request_%d", userID, requestID),
		StartTime: startTime,
	}

	// Create test request body
	reqBody := models.RiskAssessmentRequest{
		BusinessName:      fmt.Sprintf("Load Test Company %d-%d", userID, requestID),
		BusinessAddress:   "123 Load Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		Email:             fmt.Sprintf("test%d-%d@loadtest.com", userID, requestID),
		Phone:             "+1-555-123-4567",
		Website:           "https://loadtest.com",
		PredictionHorizon: 3,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		request.Error = fmt.Errorf("failed to marshal request: %w", err)
		request.EndTime = time.Now()
		request.Duration = request.EndTime.Sub(request.StartTime)
		return request
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", lt.baseURL+"/api/v1/assess", bytes.NewBuffer(jsonBody))
	if err != nil {
		request.Error = fmt.Errorf("failed to create HTTP request: %w", err)
		request.EndTime = time.Now()
		request.Duration = request.EndTime.Sub(request.StartTime)
		return request
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := lt.httpClient.Do(httpReq)
	request.EndTime = time.Now()
	request.Duration = request.EndTime.Sub(request.StartTime)

	if err != nil {
		request.Error = err
		return request
	}

	request.Response = resp
	request.Success = resp.StatusCode >= 200 && resp.StatusCode < 300

	if !request.Success {
		request.Error = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return request
}

// collectResults collects and processes load test results
func (lt *LoadTester) collectResults(requestChan <-chan LoadTestRequest, resultChan chan<- LoadTestResult, config LoadTestConfig) {
	var totalRequests, successfulRequests, failedRequests int64
	var totalDuration time.Duration
	var responseTimes []time.Duration
	var errors []LoadTestError
	var minResponseTime, maxResponseTime time.Duration

	startTime := time.Now()

	for request := range requestChan {
		totalRequests++
		totalDuration += request.Duration

		if request.Success {
			successfulRequests++
		} else {
			failedRequests++
			errors = append(errors, LoadTestError{
				RequestID: request.ID,
				Error:     request.Error.Error(),
				Timestamp: request.StartTime,
				Duration:  request.Duration,
			})
		}

		// Track response times
		responseTimes = append(responseTimes, request.Duration)
		if minResponseTime == 0 || request.Duration < minResponseTime {
			minResponseTime = request.Duration
		}
		if request.Duration > maxResponseTime {
			maxResponseTime = request.Duration
		}
	}

	endTime := time.Now()
	testDuration := endTime.Sub(startTime)

	// Calculate metrics
	var averageResponseTime time.Duration
	if len(responseTimes) > 0 {
		var totalResponseTime time.Duration
		for _, duration := range responseTimes {
			totalResponseTime += duration
		}
		averageResponseTime = totalResponseTime / time.Duration(len(responseTimes))
	}

	requestsPerSecond := float64(totalRequests) / testDuration.Seconds()
	requestsPerMinute := requestsPerSecond * 60
	errorRate := float64(failedRequests) / float64(totalRequests)

	result := LoadTestResult{
		TotalRequests:       totalRequests,
		SuccessfulRequests:  successfulRequests,
		FailedRequests:      failedRequests,
		TotalDuration:       testDuration,
		AverageResponseTime: averageResponseTime,
		MinResponseTime:     minResponseTime,
		MaxResponseTime:     maxResponseTime,
		RequestsPerSecond:   requestsPerSecond,
		RequestsPerMinute:   requestsPerMinute,
		ErrorRate:           errorRate,
		ResponseTimes:       responseTimes,
		Errors:              errors,
		Timestamp:           time.Now(),
	}

	resultChan <- result
}

// RunStressTest runs a stress test to find breaking points
func (lt *LoadTester) RunStressTest(ctx context.Context, baseConfig LoadTestConfig) (*LoadTestResult, error) {
	lt.logger.Info("Starting stress test")

	// Start with base configuration
	currentConfig := baseConfig
	bestResult := &LoadTestResult{}

	// Gradually increase load until we hit limits
	for {
		lt.logger.Info("Running stress test iteration",
			zap.Int("concurrent_users", currentConfig.ConcurrentUsers),
			zap.Float64("target_rps", currentConfig.TargetRPS))

		result, err := lt.RunLoadTest(ctx, currentConfig)
		if err != nil {
			return bestResult, err
		}

		// Check if we've hit our limits
		if result.ErrorRate > 0.05 || result.AverageResponseTime > 5*time.Second {
			lt.logger.Info("Stress test limits reached",
				zap.Float64("error_rate", result.ErrorRate),
				zap.Duration("avg_response_time", result.AverageResponseTime))
			break
		}

		bestResult = result

		// Increase load
		currentConfig.ConcurrentUsers = int(float64(currentConfig.ConcurrentUsers) * 1.5)
		currentConfig.TargetRPS = currentConfig.TargetRPS * 1.5

		// Check if we should continue
		select {
		case <-ctx.Done():
			return bestResult, ctx.Err()
		default:
		}
	}

	return bestResult, nil
}

// RunSpikeTest runs a spike test to test system recovery
func (lt *LoadTester) RunSpikeTest(ctx context.Context, baseConfig LoadTestConfig) (*LoadTestResult, error) {
	lt.logger.Info("Starting spike test")

	// Phase 1: Normal load
	lt.logger.Info("Phase 1: Normal load")
	normalConfig := baseConfig
	normalResult, err := lt.RunLoadTest(ctx, normalConfig)
	if err != nil {
		return nil, err
	}

	// Phase 2: Spike load (10x normal)
	lt.logger.Info("Phase 2: Spike load")
	spikeConfig := baseConfig
	spikeConfig.ConcurrentUsers *= 10
	spikeConfig.TargetRPS *= 10
	spikeConfig.Duration = 2 * time.Minute // Shorter spike duration

	spikeResult, err := lt.RunLoadTest(ctx, spikeConfig)
	if err != nil {
		return normalResult, err
	}

	// Phase 3: Recovery (back to normal)
	lt.logger.Info("Phase 3: Recovery")
	recoveryConfig := baseConfig
	recoveryConfig.Duration = 5 * time.Minute // Longer recovery period

	recoveryResult, err := lt.RunLoadTest(ctx, recoveryConfig)
	if err != nil {
		return spikeResult, err
	}

	// Combine results
	combinedResult := &LoadTestResult{
		TotalRequests:       normalResult.TotalRequests + spikeResult.TotalRequests + recoveryResult.TotalRequests,
		SuccessfulRequests:  normalResult.SuccessfulRequests + spikeResult.SuccessfulRequests + recoveryResult.SuccessfulRequests,
		FailedRequests:      normalResult.FailedRequests + spikeResult.FailedRequests + recoveryResult.FailedRequests,
		TotalDuration:       normalResult.TotalDuration + spikeResult.TotalDuration + recoveryResult.TotalDuration,
		AverageResponseTime: (normalResult.AverageResponseTime + spikeResult.AverageResponseTime + recoveryResult.AverageResponseTime) / 3,
		MinResponseTime:     minDuration(normalResult.MinResponseTime, spikeResult.MinResponseTime, recoveryResult.MinResponseTime),
		MaxResponseTime:     maxDuration(normalResult.MaxResponseTime, spikeResult.MaxResponseTime, recoveryResult.MaxResponseTime),
		RequestsPerSecond:   (normalResult.RequestsPerSecond + spikeResult.RequestsPerSecond + recoveryResult.RequestsPerSecond) / 3,
		RequestsPerMinute:   (normalResult.RequestsPerMinute + spikeResult.RequestsPerMinute + recoveryResult.RequestsPerMinute) / 3,
		ErrorRate:           (normalResult.ErrorRate + spikeResult.ErrorRate + recoveryResult.ErrorRate) / 3,
		Timestamp:           time.Now(),
	}

	return combinedResult, nil
}

// Helper functions
func minDuration(a, b, c time.Duration) time.Duration {
	min := a
	if b < min {
		min = b
	}
	if c < min {
		min = c
	}
	return min
}

func maxDuration(a, b, c time.Duration) time.Duration {
	max := a
	if b > max {
		max = b
	}
	if c > max {
		max = c
	}
	return max
}
