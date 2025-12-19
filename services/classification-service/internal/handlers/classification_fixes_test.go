package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap/zaptest"
)

// TestCacheKeyConsistency tests that cache keys are generated consistently
// FIX: Verifies cache key generation matches between handler and service
func TestCacheKeyConsistency(t *testing.T) {
	handler := &ClassificationHandler{}
	
	req1 := &ClassificationRequest{
		BusinessName: "Acme Corp",
		Description:  "Technology company",
		WebsiteURL:   "https://acme.com",
	}
	
	req2 := &ClassificationRequest{
		BusinessName: "acme corp", // Different case
		Description:  "Technology company",
		WebsiteURL:   "https://acme.com",
	}
	
	req3 := &ClassificationRequest{
		BusinessName: "  Acme Corp  ", // With whitespace
		Description:  "Technology company",
		WebsiteURL:   "https://acme.com",
	}
	
	key1 := handler.getCacheKey(req1)
	key2 := handler.getCacheKey(req2)
	key3 := handler.getCacheKey(req3)
	
	// All should generate the same key (normalized)
	if key1 != key2 {
		t.Errorf("Cache keys should match for different case: %s != %s", key1, key2)
	}
	
	if key1 != key3 {
		t.Errorf("Cache keys should match with whitespace: %s != %s", key1, key3)
	}
	
	// Key should have classification: prefix
	if !strings.HasPrefix(key1, "classification:") {
		t.Errorf("Cache key should start with 'classification:' prefix, got: %s", key1)
	}
	
	// Different requests should generate different keys
	req4 := &ClassificationRequest{
		BusinessName: "Different Company",
		Description:  "Technology company",
		WebsiteURL:   "https://acme.com",
	}
	key4 := handler.getCacheKey(req4)
	if key1 == key4 {
		t.Errorf("Different business names should generate different keys")
	}
}

// TestErrorResponseStructure tests that error responses include all required frontend fields
// FIX: Verifies sendErrorResponse returns ClassificationResponse structure
func TestErrorResponseStructure(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := &ClassificationHandler{
		logger: logger,
	}
	
	req := &ClassificationRequest{
		RequestID:    "test-123",
		BusinessName: "Test Company",
		Description:  "Test description",
	}
	
	rr := httptest.NewRecorder()
	httpReq := httptest.NewRequest("POST", "/classify", nil)
	
	handler.sendErrorResponse(rr, httpReq, req, &testError{message: "test error"}, http.StatusInternalServerError)
	
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	
	var response ClassificationResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}
	
	// Verify required fields are present
	if response.RequestID != req.RequestID {
		t.Errorf("Expected RequestID %s, got %s", req.RequestID, response.RequestID)
	}
	
	if response.BusinessName != req.BusinessName {
		t.Errorf("Expected BusinessName %s, got %s", req.BusinessName, response.BusinessName)
	}
	
	// PrimaryIndustry should be present (empty string is valid for errors)
	_ = response.PrimaryIndustry // Field is always present as it's a string, not pointer
	
	if response.Classification == nil {
		t.Error("Classification field should be present")
	}
	
	if response.Classification.MCCCodes == nil {
		t.Error("MCCCodes should be present (even if empty)")
	}
	
	if response.Classification.NAICSCodes == nil {
		t.Error("NAICSCodes should be present (even if empty)")
	}
	
	if response.Classification.SICCodes == nil {
		t.Error("SICCodes should be present (even if empty)")
	}
	
	if response.ConfidenceScore != 0.0 {
		t.Errorf("Expected ConfidenceScore 0.0, got %f", response.ConfidenceScore)
	}
	
	if response.Explanation == "" {
		t.Error("Explanation should be present")
	}
	
	if response.Status != "error" {
		t.Errorf("Expected Status 'error', got %s", response.Status)
	}
	
	if response.Success != false {
		t.Error("Success should be false for error responses")
	}
	
	if response.Metadata == nil {
		t.Error("Metadata should be present")
	}
	
	if response.Metadata["error"] == nil {
		t.Error("Metadata should include error information")
	}
}

// TestMetadataFallback tests that metadata is populated from fallback sources
// FIX: Verifies inferStrategyFromPath helper function
func TestInferStrategyFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"layer1", "early_exit"},
		{"layer1_high_conf", "early_exit"},
		{"layer2", "standard_scraping"},
		{"layer2_better", "standard_scraping"},
		{"layer3", "deep_scraping"},
		{"layer3_llm", "deep_scraping"},
		{"unknown", "unknown"},
		{"", "unknown"},
	}
	
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := inferStrategyFromPath(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s for path %s", tt.expected, result, tt.path)
			}
		})
	}
}

// testError is a simple error type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

