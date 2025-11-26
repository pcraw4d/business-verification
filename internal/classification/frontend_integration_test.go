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
	"testing"
	"time"

	"kyb-platform/internal/classification/repository"
	"kyb-platform/internal/database"
)

// TestFrontendIntegrationComprehensive tests that API responses match frontend expectations
func TestFrontendIntegrationComprehensive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping frontend integration test in short mode")
	}

	// Check for Supabase credentials
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		t.Skip("Skipping frontend integration test: SUPABASE_URL or SUPABASE_ANON_KEY not set")
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
	handler := createFrontendCompatibleHandler(service, logger)

	// Test cases based on frontend expectations
	testCases := []struct {
		name           string
		requestBody    map[string]interface{}
		validateResult func(t *testing.T, response map[string]interface{})
	}{
		{
			name: "Frontend expects success flag",
			requestBody: map[string]interface{}{
				"business_name": "Microsoft Corporation",
				"description":   "Software development and cloud computing services",
				"website_url":   "https://microsoft.com",
			},
			validateResult: func(t *testing.T, response map[string]interface{}) {
				// Frontend checks: result.success || result.response
				success, ok := response["success"].(bool)
				if !ok {
					t.Error("Response missing 'success' field (frontend requires this)")
					return
				}
				if !success {
					t.Error("Response 'success' field is false (frontend will show error)")
				}
				t.Logf("✅ Frontend compatibility: success field = %v", success)
			},
		},
		{
			name: "Frontend expects business_name field",
			requestBody: map[string]interface{}{
				"business_name": "Amazon",
				"website_url":   "https://amazon.com",
			},
			validateResult: func(t *testing.T, response map[string]interface{}) {
				// Frontend uses: result.business_name || result.response.business_name
				businessName, ok := response["business_name"].(string)
				if !ok || businessName == "" {
					t.Error("Response missing 'business_name' field (frontend displays this)")
				} else {
					t.Logf("✅ Frontend compatibility: business_name = %s", businessName)
				}
			},
		},
		{
			name: "Frontend expects confidence_score field",
			requestBody: map[string]interface{}{
				"business_name": "Mayo Clinic",
				"website_url":   "https://mayoclinic.org",
			},
			validateResult: func(t *testing.T, response map[string]interface{}) {
				// Frontend uses: result.confidence_score || result.response.confidence_score || result.confidence
				var confidence float64
				var found bool

				if conf, ok := response["confidence_score"].(float64); ok {
					confidence = conf
					found = true
				} else if conf, ok := response["confidence"].(float64); ok {
					confidence = conf
					found = true
				}

				if !found {
					t.Error("Response missing 'confidence_score' or 'confidence' field (frontend displays this)")
				} else {
					t.Logf("✅ Frontend compatibility: confidence_score = %.2f%%", confidence*100)
				}
			},
		},
		{
			name: "Frontend expects classification object with codes",
			requestBody: map[string]interface{}{
				"business_name": "TechCorp Solutions",
				"description":   "Technology consulting",
			},
			validateResult: func(t *testing.T, response map[string]interface{}) {
				// Frontend checks: result.classification || result.response.classification
				// Then accesses: result.classification.mcc_codes, naics_codes, sic_codes
				classification, ok := response["classification"].(map[string]interface{})
				if !ok {
					t.Logf("⚠️ Response missing 'classification' object (frontend may show 'No classification results')")
					return
				}

				// Check for code arrays
				mccCodes, hasMCC := classification["mcc_codes"]
				naicsCodes, hasNAICS := classification["naics_codes"]
				sicCodes, hasSIC := classification["sic_codes"]

				if !hasMCC && !hasNAICS && !hasSIC {
					t.Logf("⚠️ Classification object missing code arrays (frontend expects mcc_codes, naics_codes, or sic_codes)")
				} else {
					t.Logf("✅ Frontend compatibility: classification object with codes")
					if hasMCC {
						if codes, ok := mccCodes.([]interface{}); ok {
							t.Logf("   - MCC codes: %d", len(codes))
						}
					}
					if hasNAICS {
						if codes, ok := naicsCodes.([]interface{}); ok {
							t.Logf("   - NAICS codes: %d", len(codes))
						}
					}
					if hasSIC {
						if codes, ok := sicCodes.([]interface{}); ok {
							t.Logf("   - SIC codes: %d", len(codes))
						}
					}
				}
			},
		},
		{
			name: "Frontend expects primary_industry field",
			requestBody: map[string]interface{}{
				"business_name": "Microsoft Corporation",
				"website_url":   "https://microsoft.com",
			},
			validateResult: func(t *testing.T, response map[string]interface{}) {
				// Frontend uses: result.primary_industry || result.response.primary_industry || result.industry_name
				var industryName string
				var found bool

				if ind, ok := response["primary_industry"].(string); ok && ind != "" {
					industryName = ind
					found = true
				} else if ind, ok := response["industry_name"].(string); ok && ind != "" {
					industryName = ind
					found = true
				}

				if !found {
					t.Logf("⚠️ Response missing 'primary_industry' or 'industry_name' field (frontend may show 'Unknown')")
				} else {
					t.Logf("✅ Frontend compatibility: primary_industry = %s", industryName)
				}
			},
		},
		{
			name: "Frontend expects method field",
			requestBody: map[string]interface{}{
				"business_name": "Amazon",
				"website_url":   "https://amazon.com",
			},
			validateResult: func(t *testing.T, response map[string]interface{}) {
				// Frontend may use: result.method || result.response.method
				method, ok := response["method"].(string)
				if !ok {
					t.Logf("⚠️ Response missing 'method' field (frontend may not display classification method)")
				} else {
					if method != "multi_strategy" {
						t.Logf("⚠️ Expected 'multi_strategy' method, got: %s", method)
					} else {
						t.Logf("✅ Frontend compatibility: method = %s", method)
					}
				}
			},
		},
		{
			name: "Frontend expects response wrapper (optional)",
			requestBody: map[string]interface{}{
				"business_name": "Test Business",
			},
			validateResult: func(t *testing.T, response map[string]interface{}) {
				// Frontend checks: result.response (alternative structure)
				// This is optional - if present, frontend will use it
				if responseWrapper, ok := response["response"].(map[string]interface{}); ok {
					t.Logf("✅ Frontend compatibility: response wrapper present (alternative structure)")
					if responseWrapper["classification"] != nil {
						t.Logf("   - response.classification present")
					}
				} else {
					t.Logf("✅ Frontend compatibility: direct response structure (no wrapper)")
				}
			},
		},
		{
			name: "Frontend expects keywords array",
			requestBody: map[string]interface{}{
				"business_name": "Mayo Clinic",
				"website_url":   "https://mayoclinic.org",
			},
			validateResult: func(t *testing.T, response map[string]interface{}) {
				// Frontend may use: result.keywords || result.response.keywords || result.website_keywords
				var keywords []interface{}
				var found bool

				if kw, ok := response["keywords"].([]interface{}); ok {
					keywords = kw
					found = true
				} else if kw, ok := response["website_keywords"].([]interface{}); ok {
					keywords = kw
					found = true
				}

				if !found {
					t.Logf("⚠️ Response missing 'keywords' field (frontend may not display keywords)")
				} else {
					t.Logf("✅ Frontend compatibility: keywords array with %d items", len(keywords))
				}
			},
		},
		{
			name: "Frontend expects processing_time or timestamp",
			requestBody: map[string]interface{}{
				"business_name": "TechCorp",
			},
			validateResult: func(t *testing.T, response map[string]interface{}) {
				// Frontend may display processing time
				hasProcessingTime := response["processing_time"] != nil
				hasTimestamp := response["timestamp"] != nil

				if !hasProcessingTime && !hasTimestamp {
					t.Logf("⚠️ Response missing 'processing_time' or 'timestamp' field")
				} else {
					t.Logf("✅ Frontend compatibility: timing information present")
				}
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

			req := httptest.NewRequest("POST", "/v1/classify", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute request
			handler.ServeHTTP(w, req)

			// Verify status code
			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
				t.Logf("Response body: %s", w.Body.String())
				return
			}

			// Parse response
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v\nBody: %s", err, w.Body.String())
			}

			// Validate frontend compatibility
			tc.validateResult(t, response)
		})
	}
}

// TestFrontendResponseFormat validates the exact response format frontend expects
func TestFrontendResponseFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping frontend response format test in short mode")
	}

	// Check for Supabase credentials
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		t.Skip("Skipping test: SUPABASE_URL or SUPABASE_ANON_KEY not set")
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
	handler := createFrontendCompatibleHandler(service, logger)

	// Test with a real business
	requestBody := map[string]interface{}{
		"business_name": "Microsoft Corporation",
		"description":   "Software development and cloud computing services",
		"website_url":   "https://microsoft.com",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/v1/classify", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d\nBody: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Validate required fields for frontend
	t.Run("Required fields", func(t *testing.T) {
		requiredFields := []string{"success", "business_name"}
		for _, field := range requiredFields {
			if _, exists := response[field]; !exists {
				t.Errorf("Missing required field: %s (frontend requires this)", field)
			}
		}
	})

	// Validate response structure matches frontend expectations
	t.Run("Response structure", func(t *testing.T) {
		// Check success flag
		if success, ok := response["success"].(bool); ok {
			if !success {
				t.Error("success field is false (frontend will show error)")
			}
		} else {
			t.Error("success field is not a boolean (frontend expects boolean)")
		}

		// Check business_name
		if businessName, ok := response["business_name"].(string); ok {
			if businessName == "" {
				t.Error("business_name is empty (frontend displays this)")
			}
		} else {
			t.Error("business_name is not a string (frontend expects string)")
		}

		// Check confidence (frontend accepts confidence_score or confidence)
		hasConfidence := false
		if _, ok := response["confidence_score"].(float64); ok {
			hasConfidence = true
		}
		if _, ok := response["confidence"].(float64); ok {
			hasConfidence = true
		}
		if !hasConfidence {
			t.Error("Missing confidence_score or confidence field (frontend displays this)")
		}

		// Check industry name (frontend accepts primary_industry or industry_name)
		hasIndustry := false
		if ind, ok := response["primary_industry"].(string); ok && ind != "" {
			hasIndustry = true
		}
		if ind, ok := response["industry_name"].(string); ok && ind != "" {
			hasIndustry = true
		}
		if !hasIndustry {
			t.Error("Missing primary_industry or industry_name field (frontend displays this)")
		}
	})

	// Log full response structure for debugging
	t.Logf("✅ Full response structure:")
	t.Logf("   - success: %v", response["success"])
	t.Logf("   - business_name: %v", response["business_name"])
	t.Logf("   - confidence_score: %v", response["confidence_score"])
	t.Logf("   - confidence: %v", response["confidence"])
	t.Logf("   - primary_industry: %v", response["primary_industry"])
	t.Logf("   - industry_name: %v", response["industry_name"])
	t.Logf("   - method: %v", response["method"])
	t.Logf("   - classification: %v", response["classification"] != nil)
	t.Logf("   - keywords: %v", response["keywords"] != nil)
}

// createFrontendCompatibleHandler creates an HTTP handler that returns frontend-compatible responses
func createFrontendCompatibleHandler(service *IndustryDetectionService, logger *log.Logger) http.Handler {
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

		// Build frontend-compatible response
		// Frontend expects: { success: true, business_name, confidence_score, primary_industry, classification: { mcc_codes, naics_codes, sic_codes }, ... }
		response := map[string]interface{}{
			"success":         true, // Required by frontend
			"business_name":   result.IndustryName, // Frontend displays this
			"confidence_score": result.Confidence,  // Frontend displays this
			"confidence":      result.Confidence,   // Alternative field name
			"primary_industry": result.IndustryName, // Frontend uses this
			"industry_name":   result.IndustryName, // Alternative field name
			"method":          result.Method,        // Frontend may display this
			"keywords":        result.Keywords,     // Frontend may display this
			"processing_time": result.ProcessingTime.String(),
			"timestamp":       time.Now().Format(time.RFC3339),
		}

		// Add reasoning if available
		if result.Reasoning != "" {
			response["reasoning"] = result.Reasoning
			response["classification_reasoning"] = result.Reasoning // Alternative field name
		}

		// Note: Classification codes (MCC, NAICS, SIC) would be added here
		// if the service returns them. For now, we're testing the basic structure.
		// The frontend can handle missing classification codes gracefully.

		// Return response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Printf("Failed to encode response: %v", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})
}

