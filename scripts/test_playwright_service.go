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
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// ScrapeRequest represents a scrape request
type ScrapeRequest struct {
	URL string `json:"url"`
}

// ScrapeResponse represents the scrape response
type ScrapeResponse struct {
	HTML    string `json:"html"`
	Error   string `json:"error,omitempty"`
	Success bool   `json:"success"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test_playwright_service.go <service_url>")
		fmt.Println("Example: go run test_playwright_service.go https://playwright-scraper-production.up.railway.app")
		os.Exit(1)
	}

	serviceURL := os.Args[1]
	baseURL := serviceURL
	if baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	fmt.Println("=== Playwright Scraper Service Connectivity Test ===\n")
	fmt.Printf("Service URL: %s\n\n", baseURL)

	// Test 1: Health check
	fmt.Println("Test 1: Health Check")
	healthStatus, err := testHealthCheck(baseURL)
	if err != nil {
		fmt.Printf("❌ Health check failed: %v\n\n", err)
	} else {
		fmt.Printf("✅ Health check successful\n")
		fmt.Printf("   Status: %s\n", healthStatus.Status)
		if healthStatus.Message != "" {
			fmt.Printf("   Message: %s\n", healthStatus.Message)
		}
		fmt.Println()
	}

	// Test 2: Scrape simple website
	fmt.Println("Test 2: Scrape Simple Website (example.com)")
	if err := testScrape(baseURL, "https://example.com"); err != nil {
		fmt.Printf("❌ Scrape failed: %v\n\n", err)
	} else {
		fmt.Printf("✅ Scrape successful\n\n")
	}

	// Test 3: Scrape JavaScript-heavy website
	fmt.Println("Test 3: Scrape JavaScript-Heavy Website (github.com)")
	if err := testScrape(baseURL, "https://github.com"); err != nil {
		fmt.Printf("❌ Scrape failed: %v\n\n", err)
	} else {
		fmt.Printf("✅ Scrape successful\n\n")
	}

	// Test 4: Scrape invalid URL (should handle gracefully)
	fmt.Println("Test 4: Scrape Invalid URL (should handle gracefully)")
	if err := testScrape(baseURL, "https://invalid-url-that-does-not-exist-12345.com"); err != nil {
		fmt.Printf("⚠️ Scrape failed as expected: %v\n\n", err)
	} else {
		fmt.Printf("✅ Scrape handled gracefully\n\n")
	}

	fmt.Println("=== Test Complete ===")
}

func testHealthCheck(baseURL string) (*HealthCheckResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/health", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
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
		// If JSON decode fails, try to read as plain text
		body, _ := io.ReadAll(resp.Body)
		return &HealthCheckResponse{
			Status:  "unknown",
			Message: string(body),
		}, nil
	}

	return &healthStatus, nil
}

func testScrape(baseURL string, targetURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	reqBody := ScrapeRequest{
		URL: targetURL,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/scrape", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		return fmt.Errorf("request failed after %v: %w", duration, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d after %v: %s", resp.StatusCode, duration, string(body))
	}

	var result ScrapeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("scrape unsuccessful: %s", result.Error)
	}

	if result.Error != "" {
		return fmt.Errorf("scrape returned error: %s", result.Error)
	}

	htmlLength := len(result.HTML)
	fmt.Printf("   URL: %s\n", targetURL)
	fmt.Printf("   Duration: %v\n", duration)
	fmt.Printf("   HTML Length: %d bytes\n", htmlLength)
	if htmlLength > 0 {
		preview := result.HTML
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		fmt.Printf("   HTML Preview: %s\n", preview)
	}

	return nil
}

