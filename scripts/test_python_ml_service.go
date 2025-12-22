package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// HealthCheckResponse represents the health check response
type HealthCheckResponse struct {
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	Timestamp string                 `json:"timestamp"`
	Checks    map[string]interface{} `json:"checks,omitempty"`
}

// ClassificationRequest represents a test classification request
type ClassificationRequest struct {
	BusinessName string `json:"business_name"`
	Description  string `json:"description"`
	WebsiteURL   string `json:"website_url,omitempty"`
}

// ClassificationResponse represents the classification response
type ClassificationResponse struct {
	Success        bool                   `json:"success"`
	Classifications []ClassificationResult `json:"classifications"`
	Message        string                 `json:"message,omitempty"`
}

// ClassificationResult represents a single classification result
type ClassificationResult struct {
	Label      string  `json:"label"`
	Confidence float64 `json:"confidence"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test_python_ml_service.go <service_url>")
		fmt.Println("Example: go run test_python_ml_service.go https://python-ml-service-production.up.railway.app")
		os.Exit(1)
	}

	serviceURL := os.Args[1]
	baseURL := serviceURL
	if baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	fmt.Println("=== Python ML Service Connectivity Test ===\n")
	fmt.Printf("Service URL: %s\n\n", baseURL)

	// Test 1: Ping endpoint
	fmt.Println("Test 1: Ping Endpoint")
	if err := testPing(baseURL); err != nil {
		fmt.Printf("❌ Ping failed: %v\n\n", err)
	} else {
		fmt.Printf("✅ Ping successful\n\n")
	}

	// Test 2: Health check
	fmt.Println("Test 2: Health Check")
	healthStatus, err := testHealthCheck(baseURL)
	if err != nil {
		fmt.Printf("❌ Health check failed: %v\n\n", err)
	} else {
		fmt.Printf("✅ Health check successful\n")
		fmt.Printf("   Status: %s\n", healthStatus.Status)
		fmt.Printf("   Message: %s\n", healthStatus.Message)
		if healthStatus.Checks != nil {
			fmt.Printf("   Checks: %v\n", healthStatus.Checks)
		}
		fmt.Println()
	}

	// Test 3: Fast classification
	fmt.Println("Test 3: Fast Classification")
	if err := testFastClassification(baseURL); err != nil {
		fmt.Printf("❌ Fast classification failed: %v\n\n", err)
	} else {
		fmt.Printf("✅ Fast classification successful\n\n")
	}

	// Test 4: Enhanced classification
	fmt.Println("Test 4: Enhanced Classification")
	if err := testEnhancedClassification(baseURL); err != nil {
		fmt.Printf("❌ Enhanced classification failed: %v\n\n", err)
	} else {
		fmt.Printf("✅ Enhanced classification successful\n\n")
	}

	fmt.Println("=== Test Complete ===")
}

func testPing(baseURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/ping", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func testHealthCheck(baseURL string) (*HealthCheckResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/health", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var healthStatus HealthCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthStatus); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &healthStatus, nil
}

func testFastClassification(baseURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqBody := ClassificationRequest{
		BusinessName: "Test Bank",
		Description:  "A financial services company providing banking and investment services",
		WebsiteURL:   "https://example.com",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/classify-fast", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result ClassificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("classification unsuccessful: %s", result.Message)
	}

	fmt.Printf("   Classifications: %d\n", len(result.Classifications))
	if len(result.Classifications) > 0 {
		fmt.Printf("   Top result: %s (%.2f%%)\n", result.Classifications[0].Label, result.Classifications[0].Confidence*100)
	}

	return nil
}

func testEnhancedClassification(baseURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reqBody := ClassificationRequest{
		BusinessName: "Tech Startup Inc",
		Description:  "A technology company developing software solutions for businesses",
		WebsiteURL:   "https://example.com",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/classify-enhanced", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result ClassificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("classification unsuccessful: %s", result.Message)
	}

	fmt.Printf("   Classifications: %d\n", len(result.Classifications))
	if len(result.Classifications) > 0 {
		fmt.Printf("   Top result: %s (%.2f%%)\n", result.Classifications[0].Label, result.Classifications[0].Confidence*100)
	}

	return nil
}

