package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// LoadTestingAPI provides HTTP endpoints for load testing and capacity planning
type LoadTestingAPI struct {
	loadTester *LoadTester
	queue      *RequestQueue
}

// NewLoadTestingAPI creates a new load testing API
func NewLoadTestingAPI(queue *RequestQueue) *LoadTestingAPI {
	return &LoadTestingAPI{
		loadTester: NewLoadTester(nil),
		queue:      queue,
	}
}

// StartLoadTestRequest represents a request to start a load test
type StartLoadTestRequest struct {
	ConcurrentUsers    int           `json:"concurrent_users"`
	RequestsPerUser    int           `json:"requests_per_user"`
	TestDuration       time.Duration `json:"test_duration"`
	RampUpTime         time.Duration `json:"ramp_up_time"`
	RampDownTime       time.Duration `json:"ramp_down_time"`
	RequestTimeout     time.Duration `json:"request_timeout"`
	TargetEndpoint     string        `json:"target_endpoint"`
	RequestPayload     string        `json:"request_payload"`
	ExpectedStatusCode int           `json:"expected_status_code"`
}

// LoadTestResponse represents the response from a load test
type LoadTestResponse struct {
	Success   bool            `json:"success"`
	Message   string          `json:"message"`
	TestID    string          `json:"test_id"`
	StartTime time.Time       `json:"start_time"`
	EndTime   time.Time       `json:"end_time"`
	Result    *LoadTestResult `json:"result,omitempty"`
	Error     string          `json:"error,omitempty"`
}

// CapacityReportResponse represents a capacity planning report
type CapacityReportResponse struct {
	Success bool   `json:"success"`
	Report  string `json:"report"`
	Error   string `json:"error,omitempty"`
}

// TestHistoryResponse represents test history
type TestHistoryResponse struct {
	Success bool              `json:"success"`
	Tests   []*LoadTestResult `json:"tests"`
	Count   int               `json:"count"`
	Error   string            `json:"error,omitempty"`
}

// StartLoadTestHandler handles starting a new load test
func (api *LoadTestingAPI) StartLoadTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req StartLoadTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.ConcurrentUsers <= 0 {
		req.ConcurrentUsers = 50 // Default
	}
	if req.RequestsPerUser <= 0 {
		req.RequestsPerUser = 10 // Default
	}
	if req.TestDuration <= 0 {
		req.TestDuration = 5 * time.Minute // Default
	}
	if req.RequestTimeout <= 0 {
		req.RequestTimeout = 10 * time.Second // Default
	}
	if req.TargetEndpoint == "" {
		req.TargetEndpoint = "/v1/classify" // Default
	}
	if req.ExpectedStatusCode == 0 {
		req.ExpectedStatusCode = 200 // Default
	}

	// Create test configuration
	config := &LoadTestConfig{
		ConcurrentUsers:    req.ConcurrentUsers,
		RequestsPerUser:    req.RequestsPerUser,
		TestDuration:       req.TestDuration,
		RampUpTime:         req.RampUpTime,
		RampDownTime:       req.RampDownTime,
		RequestTimeout:     req.RequestTimeout,
		TargetEndpoint:     req.TargetEndpoint,
		RequestPayload:     req.RequestPayload,
		ExpectedStatusCode: req.ExpectedStatusCode,
		EnableMetrics:      true,
	}

	// Create a new load tester with the configuration
	loadTester := NewLoadTester(config)

	// Create a test handler that simulates the target endpoint
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Simulate processing time
		time.Sleep(100 * time.Millisecond)

		response := map[string]interface{}{
			"success":   true,
			"message":   "load test response",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}

		json.NewEncoder(w).Encode(response)
	}

	// Run the load test in a goroutine to avoid blocking
	go func() {
		result, err := loadTester.RunLoadTest(testHandler)
		if err != nil {
			fmt.Printf("Load test failed: %v\n", err)
		} else {
			fmt.Printf("Load test completed: %d requests, %.2f RPS, %.2f%% error rate\n",
				result.TotalRequests, result.RequestsPerSecond, result.ErrorRate*100)
		}
	}()

	// Return immediate response
	response := LoadTestResponse{
		Success:   true,
		Message:   "Load test started successfully",
		TestID:    fmt.Sprintf("test_%d", time.Now().Unix()),
		StartTime: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetCapacityReportHandler returns a comprehensive capacity planning report
func (api *LoadTestingAPI) GetCapacityReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Generate capacity report
	report := api.loadTester.GenerateCapacityReport()

	response := CapacityReportResponse{
		Success: true,
		Report:  report,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetTestHistoryHandler returns the history of load tests
func (api *LoadTestingAPI) GetTestHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters for pagination
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Get test history
	history := api.loadTester.GetTestHistory()

	// Apply limit
	if len(history) > limit {
		history = history[len(history)-limit:]
	}

	response := TestHistoryResponse{
		Success: true,
		Tests:   history,
		Count:   len(history),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetQueueMetricsHandler returns current queue metrics
func (api *LoadTestingAPI) GetQueueMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if api.queue == nil {
		http.Error(w, "Queue not available", http.StatusServiceUnavailable)
		return
	}

	metrics := api.queue.GetMetrics()

	response := map[string]interface{}{
		"success": true,
		"metrics": map[string]interface{}{
			"total_requests":       metrics.TotalRequests,
			"processed_requests":   metrics.ProcessedRequests,
			"failed_requests":      metrics.FailedRequests,
			"queue_size":           metrics.QueueSize,
			"active_workers":       metrics.ActiveWorkers,
			"average_wait_time":    metrics.AverageWaitTime.String(),
			"average_process_time": metrics.AverageProcessTime.String(),
			"last_updated":         metrics.LastUpdated.Format(time.RFC3339),
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RunQuickLoadTestHandler runs a quick load test with default settings
func (api *LoadTestingAPI) RunQuickLoadTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create a quick test configuration
	config := &LoadTestConfig{
		ConcurrentUsers: 10,               // Small number for quick test
		RequestsPerUser: 5,                // Small number for quick test
		TestDuration:    30 * time.Second, // Short duration
		RampUpTime:      5 * time.Second,  // Quick ramp up
		RampDownTime:    5 * time.Second,  // Quick ramp down
		RequestTimeout:  5 * time.Second,  // Short timeout
		TargetEndpoint:  "/v1/classify",
		EnableMetrics:   true,
	}

	// Create a new load tester
	loadTester := NewLoadTester(config)

	// Create a test handler
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Simulate processing time
		time.Sleep(50 * time.Millisecond)

		response := map[string]interface{}{
			"success":   true,
			"message":   "quick load test response",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}

		json.NewEncoder(w).Encode(response)
	}

	// Run the load test
	result, err := loadTester.RunLoadTest(testHandler)
	if err != nil {
		response := LoadTestResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Return results
	response := LoadTestResponse{
		Success:   true,
		Message:   "Quick load test completed successfully",
		TestID:    fmt.Sprintf("quick_test_%d", time.Now().Unix()),
		StartTime: result.StartTime,
		EndTime:   result.EndTime,
		Result:    result,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetLoadTestStatusHandler returns the status of load testing system
func (api *LoadTestingAPI) GetLoadTestStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	history := api.loadTester.GetTestHistory()
	lastTest := (*LoadTestResult)(nil)
	if len(history) > 0 {
		lastTest = history[len(history)-1]
	}

	status := map[string]interface{}{
		"success": true,
		"status":  "operational",
		"features": map[string]interface{}{
			"load_testing":         true,
			"capacity_planning":    true,
			"performance_analysis": true,
			"queue_monitoring":     api.queue != nil,
		},
		"statistics": map[string]interface{}{
			"total_tests":     len(history),
			"last_test_time":  lastTest != nil,
			"queue_available": api.queue != nil,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if lastTest != nil {
		status["last_test"] = map[string]interface{}{
			"concurrent_users":      lastTest.TestConfig.ConcurrentUsers,
			"requests_per_second":   lastTest.RequestsPerSecond,
			"error_rate":            lastTest.ErrorRate * 100,
			"average_response_time": lastTest.AverageResponseTime.String(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

// RegisterLoadTestingRoutes registers all load testing routes with a mux
func (api *LoadTestingAPI) RegisterLoadTestingRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/load-test/start", api.StartLoadTestHandler)
	mux.HandleFunc("POST /v1/load-test/quick", api.RunQuickLoadTestHandler)
	mux.HandleFunc("GET /v1/load-test/report", api.GetCapacityReportHandler)
	mux.HandleFunc("GET /v1/load-test/history", api.GetTestHistoryHandler)
	mux.HandleFunc("GET /v1/load-test/queue-metrics", api.GetQueueMetricsHandler)
	mux.HandleFunc("GET /v1/load-test/status", api.GetLoadTestStatusHandler)
}
