package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type LoadTestConfig struct {
	BaseURL         string
	ConcurrentUsers int
	RequestsPerUser int
	TestDuration    time.Duration
}

type LoadTestResult struct {
	TotalRequests       int
	SuccessfulRequests  int
	FailedRequests      int
	AverageResponseTime time.Duration
	MinResponseTime     time.Duration
	MaxResponseTime     time.Duration
	RequestsPerSecond   float64
	ErrorRate           float64
}

type ClassificationRequest struct {
	BusinessName    string `json:"business_name"`
	BusinessAddress string `json:"business_address"`
	Industry        string `json:"industry"`
}

func main() {
	config := LoadTestConfig{
		BaseURL:         "https://kyb-api-gateway-production.up.railway.app",
		ConcurrentUsers: 50, // Start with 50 concurrent users
		RequestsPerUser: 20, // 20 requests per user
		TestDuration:    30 * time.Second,
	}

	fmt.Println("🔥 KYB Platform Load Testing")
	fmt.Println("============================")
	fmt.Printf("Base URL: %s\n", config.BaseURL)
	fmt.Printf("Concurrent Users: %d\n", config.ConcurrentUsers)
	fmt.Printf("Requests per User: %d\n", config.RequestsPerUser)
	fmt.Printf("Total Requests: %d\n", config.ConcurrentUsers*config.RequestsPerUser)
	fmt.Println("")

	// Test different endpoints
	endpoints := []struct {
		name     string
		method   string
		path     string
		body     interface{}
		expected int
	}{
		{
			name:     "Health Check",
			method:   "GET",
			path:     "/health",
			expected: 200,
		},
		{
			name:     "Metrics",
			method:   "GET",
			path:     "/metrics",
			expected: 200,
		},
		{
			name:     "Analytics Overall",
			method:   "GET",
			path:     "/analytics/overall",
			expected: 200,
		},
		{
			name:   "Classification",
			method: "POST",
			path:   "/classify",
			body: ClassificationRequest{
				BusinessName:    "Test Business",
				BusinessAddress: "123 Test Street, Test City, TC 12345",
				Industry:        "Technology",
			},
			expected: 200,
		},
	}

	for _, endpoint := range endpoints {
		fmt.Printf("🧪 Testing %s...\n", endpoint.name)
		result := runLoadTest(config, endpoint.method, endpoint.path, endpoint.body, endpoint.expected)
		printResults(endpoint.name, result)
		fmt.Println("")
	}

	// Stress test with increasing load
	fmt.Println("💪 Stress Testing with Increasing Load...")
	stressTestLevels := []int{10, 25, 50, 100, 200}

	for _, users := range stressTestLevels {
		fmt.Printf("🔥 Testing with %d concurrent users...\n", users)
		stressConfig := config
		stressConfig.ConcurrentUsers = users
		stressConfig.RequestsPerUser = 10 // Reduce requests per user for stress test

		result := runLoadTest(stressConfig, "GET", "/health", nil, 200)
		printResults(fmt.Sprintf("Stress Test (%d users)", users), result)

		// Wait between stress tests
		time.Sleep(5 * time.Second)
	}
}

func runLoadTest(config LoadTestConfig, method, path string, body interface{}, expectedStatus int) LoadTestResult {
	var wg sync.WaitGroup
	var mu sync.Mutex

	results := make([]time.Duration, 0, config.ConcurrentUsers*config.RequestsPerUser)
	successCount := 0
	failureCount := 0

	startTime := time.Now()

	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			client := &http.Client{
				Timeout: 30 * time.Second,
			}

			for j := 0; j < config.RequestsPerUser; j++ {
				requestStart := time.Now()

				var req *http.Request
				var err error

				if method == "GET" {
					req, err = http.NewRequest(method, config.BaseURL+path, nil)
				} else {
					var jsonBody []byte
					if body != nil {
						jsonBody, err = json.Marshal(body)
						if err != nil {
							mu.Lock()
							failureCount++
							mu.Unlock()
							continue
						}
					}
					req, err = http.NewRequest(method, config.BaseURL+path, bytes.NewBuffer(jsonBody))
					if body != nil {
						req.Header.Set("Content-Type", "application/json")
					}
				}

				if err != nil {
					mu.Lock()
					failureCount++
					mu.Unlock()
					continue
				}

				resp, err := client.Do(req)
				responseTime := time.Since(requestStart)

				mu.Lock()
				results = append(results, responseTime)
				if err != nil || resp.StatusCode != expectedStatus {
					failureCount++
				} else {
					successCount++
				}
				mu.Unlock()

				if resp != nil {
					resp.Body.Close()
				}
			}
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	// Calculate statistics
	totalRequests := len(results)
	if totalRequests == 0 {
		return LoadTestResult{}
	}

	var totalResponseTime time.Duration
	minTime := results[0]
	maxTime := results[0]

	for _, duration := range results {
		totalResponseTime += duration
		if duration < minTime {
			minTime = duration
		}
		if duration > maxTime {
			maxTime = duration
		}
	}

	avgResponseTime := totalResponseTime / time.Duration(totalRequests)
	requestsPerSecond := float64(totalRequests) / totalTime.Seconds()
	errorRate := float64(failureCount) / float64(totalRequests) * 100

	return LoadTestResult{
		TotalRequests:       totalRequests,
		SuccessfulRequests:  successCount,
		FailedRequests:      failureCount,
		AverageResponseTime: avgResponseTime,
		MinResponseTime:     minTime,
		MaxResponseTime:     maxTime,
		RequestsPerSecond:   requestsPerSecond,
		ErrorRate:           errorRate,
	}
}

func printResults(testName string, result LoadTestResult) {
	fmt.Printf("📊 %s Results:\n", testName)
	fmt.Printf("   Total Requests: %d\n", result.TotalRequests)
	fmt.Printf("   Successful: %d\n", result.SuccessfulRequests)
	fmt.Printf("   Failed: %d\n", result.FailedRequests)
	fmt.Printf("   Success Rate: %.2f%%\n", 100-result.ErrorRate)
	fmt.Printf("   Error Rate: %.2f%%\n", result.ErrorRate)
	fmt.Printf("   Avg Response Time: %v\n", result.AverageResponseTime)
	fmt.Printf("   Min Response Time: %v\n", result.MinResponseTime)
	fmt.Printf("   Max Response Time: %v\n", result.MaxResponseTime)
	fmt.Printf("   Requests/Second: %.2f\n", result.RequestsPerSecond)

	// Performance assessment
	if result.ErrorRate > 5 {
		fmt.Printf("   ⚠️  HIGH ERROR RATE - Needs optimization\n")
	} else if result.AverageResponseTime > 500*time.Millisecond {
		fmt.Printf("   ⚠️  SLOW RESPONSE TIME - Needs optimization\n")
	} else if result.RequestsPerSecond < 10 {
		fmt.Printf("   ⚠️  LOW THROUGHPUT - Needs optimization\n")
	} else {
		fmt.Printf("   ✅ GOOD PERFORMANCE\n")
	}
}
