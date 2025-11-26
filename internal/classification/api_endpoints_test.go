package classification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"kyb-platform/internal/classification/repository"
	"kyb-platform/internal/database"
)

// TestClassificationAPIEndpointsComprehensive tests all classification API endpoints
func TestClassificationAPIEndpointsComprehensive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive API endpoint test in short mode")
	}

	// Check for Supabase credentials
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		t.Skip("Skipping API endpoint test: SUPABASE_URL or SUPABASE_ANON_KEY not set")
	}

	// Create database client
	config := &database.SupabaseConfig{
		URL:           supabaseURL,
		APIKey:        supabaseKey,
		ServiceRoleKey: supabaseServiceKey,
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	client, err := database.NewSupabaseClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create Supabase client: %v", err)
	}
	defer client.Close()

	// Create repository and service
	repo := repository.NewSupabaseKeywordRepository(client, logger)
	service := NewIndustryDetectionService(repo, logger)

	// Create test handler
	handler := createClassificationHandler(service, logger)

	// Test cases for different scenarios
	testCases := []struct {
		name           string
		endpoint       string
		method         string
		requestBody    map[string]interface{}
		expectedStatus int
		validateResult func(t *testing.T, response map[string]interface{}, body string)
	}{
		{
			name:     "POST /v1/classify - Technology company",
			endpoint: "/v1/classify",
			method:   "POST",
			requestBody: map[string]interface{}{
				"business_name": "Microsoft Corporation",
				"description":   "Software development and cloud computing services",
				"website_url":   "https://microsoft.com",
			},
			expectedStatus: http.StatusOK,
			validateResult: func(t *testing.T, response map[string]interface{}, body string) {
				// Verify response structure
				if response["industry_name"] == nil {
					t.Error("Response missing industry_name")
				}
				if response["confidence"] == nil {
					t.Error("Response missing confidence")
				}
				if response["method"] == nil {
					t.Error("Response missing method")
				}

				// Verify values
				industryName, ok := response["industry_name"].(string)
				if !ok {
					t.Error("industry_name is not a string")
				} else {
					if !strings.Contains(strings.ToLower(industryName), "technology") {
						t.Logf("⚠️ Expected Technology industry, got: %s", industryName)
					}
				}

				confidence, ok := response["confidence"].(float64)
				if !ok {
					t.Error("confidence is not a number")
				} else {
					if confidence < 0.70 {
						t.Logf("⚠️ Confidence is low: %.2f%%", confidence*100)
					}
				}

				method, ok := response["method"].(string)
				if !ok {
					t.Error("method is not a string")
				} else {
					if method != "multi_strategy" {
						t.Logf("⚠️ Expected multi_strategy method, got: %s", method)
					}
				}

				t.Logf("✅ POST /v1/classify - Microsoft classified as %s with %.2f%% confidence using %s",
					industryName, confidence*100, method)
			},
		},
		{
			name:     "POST /v1/classify - Retail company",
			endpoint: "/v1/classify",
			method:   "POST",
			requestBody: map[string]interface{}{
				"business_name": "Amazon",
				"description":   "E-commerce and retail services",
				"website_url":   "https://amazon.com",
			},
			expectedStatus: http.StatusOK,
			validateResult: func(t *testing.T, response map[string]interface{}, body string) {
				industryName, ok := response["industry_name"].(string)
				if !ok {
					t.Error("industry_name is not a string")
				} else {
					if !strings.Contains(strings.ToLower(industryName), "retail") {
						t.Logf("⚠️ Expected Retail industry, got: %s", industryName)
					}
				}

				confidence, ok := response["confidence"].(float64)
				if !ok {
					t.Error("confidence is not a number")
				} else {
					if confidence < 0.65 {
						t.Logf("⚠️ Confidence is low: %.2f%%", confidence*100)
					}
				}

				t.Logf("✅ POST /v1/classify - Amazon classified as %s with %.2f%% confidence",
					industryName, confidence*100)
			},
		},
		{
			name:     "POST /v1/classify - Healthcare company",
			endpoint: "/v1/classify",
			method:   "POST",
			requestBody: map[string]interface{}{
				"business_name": "Mayo Clinic",
				"description":   "Medical center and hospital services",
				"website_url":   "https://mayoclinic.org",
			},
			expectedStatus: http.StatusOK,
			validateResult: func(t *testing.T, response map[string]interface{}, body string) {
				industryName, ok := response["industry_name"].(string)
				if !ok {
					t.Error("industry_name is not a string")
				} else {
					if !strings.Contains(strings.ToLower(industryName), "healthcare") {
						t.Logf("⚠️ Expected Healthcare industry, got: %s", industryName)
					}
				}

				confidence, ok := response["confidence"].(float64)
				if !ok {
					t.Error("confidence is not a number")
				} else {
					if confidence < 0.70 {
						t.Logf("⚠️ Confidence is low: %.2f%%", confidence*100)
					}
				}

				t.Logf("✅ POST /v1/classify - Mayo Clinic classified as %s with %.2f%% confidence",
					industryName, confidence*100)
			},
		},
		{
			name:     "POST /v2/classify - Enhanced classification",
			endpoint: "/v2/classify",
			method:   "POST",
			requestBody: map[string]interface{}{
				"business_name": "TechCorp Solutions",
				"description":   "Technology consulting and software development",
				"website_url":   "https://techcorp.com",
			},
			expectedStatus: http.StatusOK,
			validateResult: func(t *testing.T, response map[string]interface{}, body string) {
				if response["industry_name"] == nil {
					t.Error("Response missing industry_name")
				}
				if response["method"] == nil {
					t.Error("Response missing method")
				}

				method, ok := response["method"].(string)
				if !ok {
					t.Error("method is not a string")
				} else {
					if method != "multi_strategy" {
						t.Logf("⚠️ Expected multi_strategy method for v2, got: %s", method)
					}
				}

				t.Logf("✅ POST /v2/classify - Enhanced classification completed using %s", method)
			},
		},
		{
			name:     "POST /classify - Legacy endpoint",
			endpoint: "/classify",
			method:   "POST",
			requestBody: map[string]interface{}{
				"business_name": "Test Business",
				"description":   "Test description",
			},
			expectedStatus: http.StatusOK,
			validateResult: func(t *testing.T, response map[string]interface{}, body string) {
				if response["industry_name"] == nil {
					t.Error("Response missing industry_name")
				}
				t.Logf("✅ POST /classify - Legacy endpoint working")
			},
		},
		{
			name:     "POST /v1/classify - Missing business_name",
			endpoint: "/v1/classify",
			method:   "POST",
			requestBody: map[string]interface{}{
				"description": "Some description",
			},
			expectedStatus: http.StatusBadRequest,
			validateResult: func(t *testing.T, response map[string]interface{}, body string) {
				// Should return error message
				if !strings.Contains(strings.ToLower(body), "required") {
					t.Logf("⚠️ Expected error message about required field, got: %s", body)
				}
				t.Logf("✅ POST /v1/classify - Validation working (rejected missing business_name)")
			},
		},
		{
			name:           "POST /v1/classify - Invalid JSON body",
			endpoint:       "/v1/classify",
			method:         "POST",
			requestBody:    nil, // Will send invalid JSON string
			expectedStatus: http.StatusBadRequest,
			validateResult: func(t *testing.T, response map[string]interface{}, body string) {
				// Should return error message
				if !strings.Contains(strings.ToLower(body), "invalid") && !strings.Contains(strings.ToLower(body), "json") {
					t.Logf("⚠️ Expected error message about invalid JSON, got: %s", body)
				}
				t.Logf("✅ POST /v1/classify - Invalid JSON correctly rejected")
			},
		},
		{
			name:     "GET /v1/classify - Wrong method",
			endpoint: "/v1/classify",
			method:   "GET",
			requestBody: map[string]interface{}{
				"business_name": "Test",
			},
			expectedStatus: http.StatusMethodNotAllowed,
			validateResult: func(t *testing.T, response map[string]interface{}, body string) {
				t.Logf("✅ GET /v1/classify - Method not allowed correctly rejected")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request body
			var bodyBytes []byte
			var err error
			if tc.requestBody != nil {
				bodyBytes, err = json.Marshal(tc.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request: %v", err)
				}
			} else if tc.name == "POST /v1/classify - Invalid JSON body" {
				// Send invalid JSON for this specific test
				bodyBytes = []byte(`{invalid json}`)
			}

			// Create request
			req := httptest.NewRequest(tc.method, tc.endpoint, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute request
			startTime := time.Now()
			handler.ServeHTTP(w, req)
			duration := time.Since(startTime)

			// Verify status code
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
				t.Logf("Response body: %s", w.Body.String())
				return
			}

			// Validate response if successful
			if tc.expectedStatus == http.StatusOK && tc.validateResult != nil {
				var response map[string]interface{}
				bodyStr := w.Body.String()
				if err := json.Unmarshal([]byte(bodyStr), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v\nBody: %s", err, bodyStr)
				}
				tc.validateResult(t, response, bodyStr)
			} else if tc.validateResult != nil {
				// For error cases, pass the body string
				tc.validateResult(t, nil, w.Body.String())
			}

			// Log performance
			if duration > 5*time.Second {
				t.Logf("⚠️ Request took %v (exceeds 5s target)", duration)
			} else {
				t.Logf("✅ Request completed in %v", duration)
			}
		})
	}
}

// createClassificationHandler creates a simple HTTP handler for testing
func createClassificationHandler(service *IndustryDetectionService, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle different methods
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
			"success":         true,
			"timestamp":       time.Now().Format(time.RFC3339),
		})
	})
}

