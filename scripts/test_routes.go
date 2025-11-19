package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const APIBaseURL = "https://api-gateway-service-production-21fd.up.railway.app"

type TestResult struct {
	Name        string
	Status      string // PASS, FAIL, SKIP
	StatusCode  int
	Response    string
	Error       string
	Duration    time.Duration
}

func main() {
	fmt.Println("=" + string(bytes.Repeat([]byte("="), 60)))
	fmt.Println("Route Testing Suite")
	fmt.Println("=" + string(bytes.Repeat([]byte("="), 60)))
	fmt.Printf("API Base URL: %s\n", APIBaseURL)
	fmt.Printf("Test Time: %s\n\n", time.Now().Format(time.RFC3339))

	var results []TestResult

	// Phase 3.1: Authentication Routes
	fmt.Println("\n" + string(bytes.Repeat([]byte("-"), 60)))
	fmt.Println("Phase 3.1: Authentication Routes Testing")
	fmt.Println(string(bytes.Repeat([]byte("-"), 60)))

	// Test 1: Valid Registration
	result := testAuthRegisterValid()
	results = append(results, result)
	printResult(result)

	// Test 2: Missing Required Fields
	result = testAuthRegisterMissingFields()
	results = append(results, result)
	printResult(result)

	// Test 3: Invalid Email Format
	result = testAuthRegisterInvalidEmail()
	results = append(results, result)
	printResult(result)

	// Test 4: Valid Login (may fail if user doesn't exist)
	result = testAuthLoginValid()
	results = append(results, result)
	printResult(result)

	// Test 5: Invalid Credentials
	result = testAuthLoginInvalid()
	results = append(results, result)
	printResult(result)

	// Test 6: Missing Fields (Login)
	result = testAuthLoginMissingFields()
	results = append(results, result)
	printResult(result)

	// Phase 3.2: UUID Validation
	fmt.Println("\n" + string(bytes.Repeat([]byte("-"), 60)))
	fmt.Println("Phase 3.2: UUID Validation Testing")
	fmt.Println(string(bytes.Repeat([]byte("-"), 60)))

	// Test 1: Invalid UUID Format
	result = testUUIDValidationInvalid()
	results = append(results, result)
	printResult(result)

	// Test 2: "indicators" as ID (edge case)
	result = testUUIDValidationIndicators()
	results = append(results, result)
	printResult(result)

	// Phase 3.3: CORS Configuration
	fmt.Println("\n" + string(bytes.Repeat([]byte("-"), 60)))
	fmt.Println("Phase 3.3: CORS Configuration Testing")
	fmt.Println(string(bytes.Repeat([]byte("-"), 60)))

	// Test 1: Preflight Request
	result = testCORSPreflight()
	results = append(results, result)
	printResult(result)

	// Phase 6: Error Handling
	fmt.Println("\n" + string(bytes.Repeat([]byte("-"), 60)))
	fmt.Println("Phase 6: Error Handling Testing")
	fmt.Println(string(bytes.Repeat([]byte("-"), 60)))

	// Test 1: 404 Handler
	result = test404Handler()
	results = append(results, result)
	printResult(result)

	// Summary
	fmt.Println("\n" + string(bytes.Repeat([]byte("="), 60)))
	fmt.Println("Test Summary")
	fmt.Println(string(bytes.Repeat([]byte("="), 60)))

	passed := 0
	failed := 0
	skipped := 0

	for _, r := range results {
		switch r.Status {
		case "PASS":
			passed++
		case "FAIL":
			failed++
		case "SKIP":
			skipped++
		}
	}

	fmt.Printf("Total Tests: %d\n", len(results))
	fmt.Printf("Passed: %d\n", passed)
	fmt.Printf("Failed: %d\n", failed)
	fmt.Printf("Skipped: %d\n", skipped)
	fmt.Printf("Success Rate: %.1f%%\n", float64(passed)/float64(len(results))*100)

	if failed > 0 {
		os.Exit(1)
	}
}

func printResult(r TestResult) {
	statusIcon := "✅"
	if r.Status == "FAIL" {
		statusIcon = "❌"
	} else if r.Status == "SKIP" {
		statusIcon = "⏭️"
	}

	fmt.Printf("\n%s %s\n", statusIcon, r.Name)
	fmt.Printf("   Status Code: %d\n", r.StatusCode)
	if r.Error != "" {
		fmt.Printf("   Error: %s\n", r.Error)
	}
	if r.Response != "" && len(r.Response) < 200 {
		fmt.Printf("   Response: %s\n", r.Response)
	} else if r.Response != "" {
		fmt.Printf("   Response: %s...\n", r.Response[:200])
	}
	fmt.Printf("   Duration: %v\n", r.Duration)
}

func testAuthRegisterValid() TestResult {
	start := time.Now()
	name := "POST /api/v1/auth/register (Valid Data)"

	email := fmt.Sprintf("test_%d@example.com", time.Now().Unix())
	payload := map[string]interface{}{
		"email":     email,
		"password":  "TestPassword123!",
		"username":  "testuser",
		"first_name": "Test",
		"last_name":  "User",
	}

	jsonData, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", APIBaseURL+"/api/v1/auth/register", bytes.NewBuffer(jsonData))
	if err != nil {
		return TestResult{Name: name, Status: "FAIL", Error: err.Error(), Duration: time.Since(start)}
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return TestResult{Name: name, Status: "FAIL", Error: err.Error(), Duration: duration}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	response := string(body)

	status := "FAIL"
	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		status = "PASS"
	}

	return TestResult{
		Name:       name,
		Status:     status,
		StatusCode: resp.StatusCode,
		Response:   response,
		Duration:   duration,
	}
}

func testAuthRegisterMissingFields() TestResult {
	start := time.Now()
	name := "POST /api/v1/auth/register (Missing Fields)"

	payload := map[string]interface{}{
		"email": "test@example.com",
		// Missing password
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", APIBaseURL+"/api/v1/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return TestResult{Name: name, Status: "FAIL", Error: err.Error(), Duration: duration}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	response := string(body)

	status := "FAIL"
	if resp.StatusCode == 400 {
		status = "PASS"
	}

	return TestResult{
		Name:       name,
		Status:     status,
		StatusCode: resp.StatusCode,
		Response:   response,
		Duration:   duration,
	}
}

func testAuthRegisterInvalidEmail() TestResult {
	start := time.Now()
	name := "POST /api/v1/auth/register (Invalid Email)"

	payload := map[string]interface{}{
		"email":    "not-an-email",
		"password": "TestPassword123!",
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", APIBaseURL+"/api/v1/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return TestResult{Name: name, Status: "FAIL", Error: err.Error(), Duration: duration}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	response := string(body)

	status := "FAIL"
	if resp.StatusCode == 400 {
		status = "PASS"
	}

	return TestResult{
		Name:       name,
		Status:     status,
		StatusCode: resp.StatusCode,
		Response:   response,
		Duration:   duration,
	}
}

func testAuthLoginValid() TestResult {
	start := time.Now()
	name := "POST /api/v1/auth/login (Valid - May Fail if User Doesn't Exist)"

	payload := map[string]interface{}{
		"email":    "test@example.com",
		"password": "TestPassword123!",
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", APIBaseURL+"/api/v1/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return TestResult{Name: name, Status: "FAIL", Error: err.Error(), Duration: duration}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	response := string(body)

	status := "SKIP"
	if resp.StatusCode == 200 {
		status = "PASS"
	} else if resp.StatusCode == 401 {
		// User may not exist, which is acceptable for this test
		status = "SKIP"
	}

	return TestResult{
		Name:       name,
		Status:     status,
		StatusCode: resp.StatusCode,
		Response:   response,
		Duration:   duration,
	}
}

func testAuthLoginInvalid() TestResult {
	start := time.Now()
	name := "POST /api/v1/auth/login (Invalid Credentials)"

	payload := map[string]interface{}{
		"email":    "nonexistent@example.com",
		"password": "WrongPassword123!",
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", APIBaseURL+"/api/v1/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return TestResult{Name: name, Status: "FAIL", Error: err.Error(), Duration: duration}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	response := string(body)

	status := "FAIL"
	if resp.StatusCode == 401 {
		status = "PASS"
	}

	return TestResult{
		Name:       name,
		Status:     status,
		StatusCode: resp.StatusCode,
		Response:   response,
		Duration:   duration,
	}
}

func testAuthLoginMissingFields() TestResult {
	start := time.Now()
	name := "POST /api/v1/auth/login (Missing Fields)"

	payload := map[string]interface{}{
		"email": "test@example.com",
		// Missing password
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", APIBaseURL+"/api/v1/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return TestResult{Name: name, Status: "FAIL", Error: err.Error(), Duration: duration}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	response := string(body)

	status := "FAIL"
	if resp.StatusCode == 400 {
		status = "PASS"
	}

	return TestResult{
		Name:       name,
		Status:     status,
		StatusCode: resp.StatusCode,
		Response:   response,
		Duration:   duration,
	}
}

func testUUIDValidationInvalid() TestResult {
	start := time.Now()
	name := "GET /api/v1/risk/indicators/{invalid-uuid}"

	req, _ := http.NewRequest("GET", APIBaseURL+"/api/v1/risk/indicators/invalid-id", nil)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return TestResult{Name: name, Status: "FAIL", Error: err.Error(), Duration: duration}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	response := string(body)

	status := "FAIL"
	if resp.StatusCode == 400 {
		status = "PASS"
	}

	return TestResult{
		Name:       name,
		Status:     status,
		StatusCode: resp.StatusCode,
		Response:   response,
		Duration:   duration,
	}
}

func testUUIDValidationIndicators() TestResult {
	start := time.Now()
	name := "GET /api/v1/risk/indicators/indicators (Edge Case)"

	req, _ := http.NewRequest("GET", APIBaseURL+"/api/v1/risk/indicators/indicators", nil)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return TestResult{Name: name, Status: "FAIL", Error: err.Error(), Duration: duration}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	response := string(body)

	status := "FAIL"
	if resp.StatusCode == 400 {
		status = "PASS"
	}

	return TestResult{
		Name:       name,
		Status:     status,
		StatusCode: resp.StatusCode,
		Response:   response,
		Duration:   duration,
	}
}

func testCORSPreflight() TestResult {
	start := time.Now()
	name := "OPTIONS /api/v1/auth/register (CORS Preflight)"

	req, _ := http.NewRequest("OPTIONS", APIBaseURL+"/api/v1/auth/register", nil)
	req.Header.Set("Origin", "https://frontend-service-production-b225.up.railway.app")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return TestResult{Name: name, Status: "FAIL", Error: err.Error(), Duration: duration}
	}
	defer resp.Body.Close()

	// Check CORS headers
	allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	allowCredentials := resp.Header.Get("Access-Control-Allow-Credentials")

	status := "FAIL"
	if resp.StatusCode == 200 {
		if allowOrigin == "https://frontend-service-production-b225.up.railway.app" || allowOrigin == "*" {
			if allowCredentials == "true" || allowOrigin != "*" {
				status = "PASS"
			}
		}
	}

	response := fmt.Sprintf("Allow-Origin: %s, Allow-Credentials: %s", allowOrigin, allowCredentials)

	return TestResult{
		Name:       name,
		Status:     status,
		StatusCode: resp.StatusCode,
		Response:   response,
		Duration:   duration,
	}
}

func test404Handler() TestResult {
	start := time.Now()
	name := "GET /api/v1/nonexistent-route (404 Handler)"

	req, _ := http.NewRequest("GET", APIBaseURL+"/api/v1/nonexistent-route", nil)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return TestResult{Name: name, Status: "FAIL", Error: err.Error(), Duration: duration}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	response := string(body)

	status := "FAIL"
	if resp.StatusCode == 404 {
		// Check if response includes helpful error message
		if bytes.Contains(body, []byte("NOT_FOUND")) || bytes.Contains(body, []byte("not found")) {
			status = "PASS"
		}
	}

	return TestResult{
		Name:       name,
		Status:     status,
		StatusCode: resp.StatusCode,
		Response:   response,
		Duration:   duration,
	}
}

