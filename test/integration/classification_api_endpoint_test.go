package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

// TestClassificationAPIEndpoint tests the classification API endpoint with updated service
func TestClassificationAPIEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API endpoint test in short mode")
	}

	// Check for Supabase credentials
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		t.Skip("Skipping API endpoint test: SUPABASE_URL or SUPABASE_ANON_KEY not set")
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

	// Create test handler
	handler := createClassificationHandler(service, logger)

	// Test cases
	testCases := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		validateResult func(t *testing.T, response map[string]interface{})
	}{
		{
			name: "Valid classification request",
			requestBody: map[string]interface{}{
				"business_name": "Microsoft Corporation",
				"description":   "Software development and cloud computing services",
				"website_url":   "https://microsoft.com",
			},
			expectedStatus: http.StatusOK,
			validateResult: func(t *testing.T, response map[string]interface{}) {
				if response["industry_name"] == nil {
					t.Error("Response missing industry_name")
				}
				if response["confidence"] == nil {
					t.Error("Response missing confidence")
				}
				if response["method"] == nil {
					t.Error("Response missing method")
				}
				// Verify method is multi_strategy
				if method, ok := response["method"].(string); ok {
					if method != "multi_strategy" {
						t.Logf("⚠️ Expected multi_strategy method, got: %s", method)
					}
				}
			},
		},
		{
			name: "Request with only business name",
			requestBody: map[string]interface{}{
				"business_name": "TechCorp Solutions",
			},
			expectedStatus: http.StatusOK,
			validateResult: func(t *testing.T, response map[string]interface{}) {
				if response["industry_name"] == nil {
					t.Error("Response missing industry_name")
				}
			},
		},
		{
			name: "Request with missing business_name",
			requestBody: map[string]interface{}{
				"description": "Some description",
			},
			expectedStatus: http.StatusBadRequest,
			validateResult: func(t *testing.T, response map[string]interface{}) {
				// Should return error
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			body, err := json.Marshal(tc.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			req := httptest.NewRequest("POST", "/classify", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute request
			handler.ServeHTTP(w, req)

			// Verify status code
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
				t.Logf("Response body: %s", w.Body.String())
				return
			}

			// Validate response if successful
			if tc.expectedStatus == http.StatusOK && tc.validateResult != nil {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				tc.validateResult(t, response)
			}
		})
	}
}

// createClassificationHandler creates a simple HTTP handler for testing
func createClassificationHandler(service *classification.IndustryDetectionService, logger *log.Logger) http.Handler {
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
			http.Error(w, fmt.Sprintf("Classification failed: %v", err), http.StatusInternalServerError)
			return
		}

		// Return response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"industry_name":   result.IndustryName,
			"confidence":      result.Confidence,
			"keywords":        result.Keywords,
			"processing_time": result.ProcessingTime.String(),
			"method":          result.Method,
			"reasoning":       result.Reasoning,
		})
	})
}

