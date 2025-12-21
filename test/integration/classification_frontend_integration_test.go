//go:build !comprehensive_test && !e2e_railway
// +build !comprehensive_test,!e2e_railway

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"kyb-platform/internal/classification"
	"kyb-platform/internal/classification/repository"
	"kyb-platform/internal/database"
)

// TestFrontendIntegration tests that API responses match frontend expectations
func TestFrontendIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping frontend integration test in short mode")
	}

	// Check for Supabase credentials
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		t.Skip("Skipping frontend integration test: SUPABASE_URL or SUPABASE_ANON_KEY not set")
	}

	// Create database client
	config := &database.SupabaseConfig{
		URL:    supabaseURL,
		APIKey: supabaseKey,
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	client, err := database.NewSupabaseClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create Supabase client: %v", err)
	}
	defer client.Close()

	// Create repository and service
	repo := repository.NewSupabaseKeywordRepository(client, logger)
	service := classification.NewIndustryDetectionService(repo, logger)

	// Create handler matching frontend expectations
	handler := createFrontendCompatibleHandler(service, logger)

	// Test frontend request format
	requestBody := map[string]interface{}{
		"business_name": "TechCorp Solutions",
		"description":   "Software development and AI solutions",
		"website_url":   "https://techcorp.com",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/classify", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify frontend-expected fields
	requiredFields := []string{
		"success",
		"response",
	}

	for _, field := range requiredFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Missing required field: %s", field)
		}
	}

	// Verify nested response structure
	if responseObj, ok := response["response"].(map[string]interface{}); ok {
		responseFields := []string{
			"industry_name",
			"confidence",
			"keywords",
		}

		for _, field := range responseFields {
			if _, exists := responseObj[field]; !exists {
				t.Errorf("Missing required field in response: %s", field)
			}
		}

		// Verify confidence is a number between 0 and 1
		if confidence, ok := responseObj["confidence"].(float64); ok {
			if confidence < 0 || confidence > 1 {
				t.Errorf("Invalid confidence value: %f (expected 0-1)", confidence)
			}
		}

		// Verify method indicates multi-strategy
		if method, ok := responseObj["method"].(string); ok {
			if method == "multi_strategy" {
				t.Logf("✅ Multi-strategy classification confirmed")
			} else {
				t.Logf("⚠️ Method is %s (expected multi_strategy)", method)
			}
		}
	}

	t.Logf("✅ Frontend integration test passed")
	t.Logf("   Response structure matches frontend expectations")
}

// createFrontendCompatibleHandler creates a handler that matches frontend expectations
// Frontend expects: { success: true, response: { ... } }
func createFrontendCompatibleHandler(service *classification.IndustryDetectionService, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request
		var req struct {
			BusinessName string `json:"business_name"`
			Description  string `json:"description"`
			WebsiteURL   string `json:"website_url"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate
		if req.BusinessName == "" {
			http.Error(w, "business_name is required", http.StatusBadRequest)
			return
		}

		// Process classification
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		result, err := service.DetectIndustry(ctx, req.BusinessName, req.Description, req.WebsiteURL)
		if err != nil {
			logger.Printf("Classification failed: %v", err)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		// Return response in frontend-expected format
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"response": map[string]interface{}{
				"industry_name":   result.IndustryName,
				"confidence":      result.Confidence,
				"keywords":        result.Keywords,
				"processing_time": result.ProcessingTime.String(),
				"method":          result.Method,
				"reasoning":       result.Reasoning,
			},
		})
	})
}

