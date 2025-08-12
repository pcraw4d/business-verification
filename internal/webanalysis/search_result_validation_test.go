package webanalysis

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewSearchResultValidator(t *testing.T) {
	validator := NewSearchResultValidator()

	if validator == nil {
		t.Fatal("Expected validator to be created, got nil")
	}

	if len(validator.validators) == 0 {
		t.Error("Expected validators to be initialized")
	}

	if validator.httpClient == nil {
		t.Error("Expected HTTP client to be created")
	}

	config := validator.GetConfig()
	if !config.EnableURLValidation {
		t.Error("Expected EnableURLValidation true, got false")
	}

	if !config.EnableContentValidation {
		t.Error("Expected EnableContentValidation true, got false")
	}

	if config.RequestTimeout != 10*time.Second {
		t.Errorf("Expected RequestTimeout 10s, got %v", config.RequestTimeout)
	}
}

func TestSearchResultValidator_ValidateResult(t *testing.T) {
	validator := NewSearchResultValidator()

	// Create a test server for URL validation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Test content</body></html>"))
	}))
	defer server.Close()

	// Create test result
	result := &WebSearchResult{
		Title:          "Test Business - Official Website",
		URL:            server.URL,
		Description:    "Official website of Test Business, a leading technology company with comprehensive services",
		RelevanceScore: 0.9,
		Source:         "google",
	}

	ctx := context.Background()
	validationResult, err := validator.ValidateResult(ctx, result)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if validationResult == nil {
		t.Fatal("Expected validation result, got nil")
	}

	if validationResult.Result != result {
		t.Error("Expected validation result to contain the original result")
	}

	if validationResult.ValidationTime == 0 {
		t.Error("Expected validation time to be recorded")
	}

	// Check that we have validation results from all validators
	expectedValidators := 6 // URL, Content, Domain, Accessibility, Security, Freshness
	if len(validationResult.ValidationResults) != expectedValidators {
		t.Errorf("Expected %d validation results, got %d", expectedValidators, len(validationResult.ValidationResults))
	}

	// Check that the result is valid
	if !validationResult.OverallValid {
		t.Error("Expected result to be valid")
	}

	// Check overall score
	if validationResult.OverallScore < 0.5 {
		t.Errorf("Expected overall score >= 0.5, got %f", validationResult.OverallScore)
	}
}

func TestSearchResultValidator_ValidateResult_InvalidURL(t *testing.T) {
	validator := NewSearchResultValidator()

	// Create test result with invalid URL
	result := &WebSearchResult{
		Title:          "Invalid URL Test",
		URL:            "not-a-valid-url",
		Description:    "This should fail URL validation",
		RelevanceScore: 0.8,
		Source:         "google",
	}

	ctx := context.Background()
	validationResult, err := validator.ValidateResult(ctx, result)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if validationResult == nil {
		t.Fatal("Expected validation result, got nil")
	}

	// Should have errors
	if validationResult.TotalErrors == 0 {
		t.Error("Expected validation errors for invalid URL")
	}

	// Should not be overall valid
	if validationResult.OverallValid {
		t.Error("Expected result to be invalid due to URL issues")
	}
}

func TestSearchResultValidator_ValidateResult_SpamContent(t *testing.T) {
	validator := NewSearchResultValidator()

	// Create test result with spam content
	result := &WebSearchResult{
		Title:          "Spam Result - Click Here to Buy Now",
		URL:            "https://example.com",
		Description:    "Limited time offer! Click here to buy now! Make money fast!",
		RelevanceScore: 0.7,
		Source:         "google",
	}

	ctx := context.Background()
	validationResult, err := validator.ValidateResult(ctx, result)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if validationResult == nil {
		t.Fatal("Expected validation result, got nil")
	}

	// Should have errors due to spam content
	if validationResult.TotalErrors == 0 {
		t.Error("Expected validation errors for spam content")
	}

	// Should not be overall valid
	if validationResult.OverallValid {
		t.Error("Expected result to be invalid due to spam content")
	}
}

func TestSearchResultValidator_ValidateResults(t *testing.T) {
	validator := NewSearchResultValidator()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Test content</body></html>"))
	}))
	defer server.Close()

	// Create test results
	results := []WebSearchResult{
		{
			Title:          "Valid Result 1",
			URL:            server.URL,
			Description:    "This is a valid result with good content",
			RelevanceScore: 0.9,
			Source:         "google",
		},
		{
			Title:          "Valid Result 2",
			URL:            server.URL + "/page2",
			Description:    "Another valid result with comprehensive information",
			RelevanceScore: 0.8,
			Source:         "google",
		},
	}

	ctx := context.Background()
	validationResults, err := validator.ValidateResults(ctx, results)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(validationResults) != 2 {
		t.Errorf("Expected 2 validation results, got %d", len(validationResults))
	}

	// Check that both results are valid
	for i, validationResult := range validationResults {
		if !validationResult.OverallValid {
			t.Errorf("Expected result %d to be valid", i)
		}
	}
}

func TestURLValidator(t *testing.T) {
	validator := &URLValidator{
		config: ValidationConfig{
			EnableURLValidation: true,
			ValidTLDs:           []string{".com", ".org", ".net"},
			AllowedStatusCodes:  []int{200, 201, 202},
		},
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	// Test valid URL
	validResult := &WebSearchResult{
		URL: "https://example.com/page",
	}

	ctx := context.Background()
	validationResult, err := validator.Validate(ctx, validResult)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !validationResult.IsValid {
		t.Error("Expected valid URL to pass validation")
	}

	// Test invalid URL
	invalidResult := &WebSearchResult{
		URL: "not-a-valid-url",
	}

	validationResult, err = validator.Validate(ctx, invalidResult)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if validationResult.IsValid {
		t.Error("Expected invalid URL to fail validation")
	}

	if len(validationResult.Errors) == 0 {
		t.Error("Expected validation errors for invalid URL")
	}
}

func TestContentValidator(t *testing.T) {
	validator := &ContentValidator{
		config: ValidationConfig{
			EnableContentValidation: true,
			MinContentLength:        50,
			MaxContentLength:        1000,
			BlockedContentKeywords:  []string{"spam", "click here", "buy now"},
		},
	}

	// Test valid content
	validResult := &WebSearchResult{
		Title:       "Valid Content Title",
		Description: "This is a valid description with sufficient content length for validation purposes",
	}

	ctx := context.Background()
	validationResult, err := validator.Validate(ctx, validResult)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !validationResult.IsValid {
		t.Error("Expected valid content to pass validation")
	}

	// Test spam content
	spamResult := &WebSearchResult{
		Title:       "Spam Title",
		Description: "Click here to buy now! Limited time offer!",
	}

	validationResult, err = validator.Validate(ctx, spamResult)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if validationResult.IsValid {
		t.Error("Expected spam content to fail validation")
	}

	if len(validationResult.Errors) == 0 {
		t.Error("Expected validation errors for spam content")
	}
}

func TestDomainValidator(t *testing.T) {
	validator := &DomainValidator{
		config: ValidationConfig{
			EnableDomainValidation: true,
			BlockedDomains:         []string{"spam.com", "malware.com"},
		},
	}

	// Test valid domain
	validResult := &WebSearchResult{
		URL: "https://example.com/page",
	}

	ctx := context.Background()
	validationResult, err := validator.Validate(ctx, validResult)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !validationResult.IsValid {
		t.Error("Expected valid domain to pass validation")
	}

	// Test blocked domain
	blockedResult := &WebSearchResult{
		URL: "https://spam.com/bad-content",
	}

	validationResult, err = validator.Validate(ctx, blockedResult)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if validationResult.IsValid {
		t.Error("Expected blocked domain to fail validation")
	}

	if len(validationResult.Errors) == 0 {
		t.Error("Expected validation errors for blocked domain")
	}
}

func TestSecurityValidator(t *testing.T) {
	validator := &SecurityValidator{
		config: ValidationConfig{
			EnableSecurityCheck: true,
		},
	}

	// Test secure URL
	secureResult := &WebSearchResult{
		URL: "https://example.com/secure-page",
	}

	ctx := context.Background()
	validationResult, err := validator.Validate(ctx, secureResult)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !validationResult.IsValid {
		t.Error("Expected secure URL to pass validation")
	}

	// Test insecure URL
	insecureResult := &WebSearchResult{
		URL: "http://example.com/insecure-page",
	}

	validationResult, err = validator.Validate(ctx, insecureResult)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(validationResult.Warnings) == 0 {
		t.Error("Expected security warnings for insecure URL")
	}
}

func TestSearchResultValidator_UpdateConfig(t *testing.T) {
	validator := NewSearchResultValidator()

	newConfig := ValidationConfig{
		EnableURLValidation:      false,
		EnableContentValidation:  false,
		EnableDomainValidation:   false,
		EnableAccessibilityCheck: false,
		EnableSecurityCheck:      false,
		EnableFreshnessCheck:     true,
		RequestTimeout:           time.Second * 30,
		MaxRedirects:             10,
		UserAgent:                "Custom User Agent",
		MinContentLength:         100,
		MaxContentLength:         5000,
		BlockedContentKeywords:   []string{"custom", "blocked", "keywords"},
		ValidTLDs:                []string{".custom", ".test"},
		MaxValidationTime:        time.Second * 60,
	}

	validator.UpdateConfig(newConfig)

	updatedConfig := validator.GetConfig()

	if updatedConfig.EnableURLValidation {
		t.Error("Expected EnableURLValidation false, got true")
	}

	if updatedConfig.EnableContentValidation {
		t.Error("Expected EnableContentValidation false, got true")
	}

	if updatedConfig.EnableDomainValidation {
		t.Error("Expected EnableDomainValidation false, got true")
	}

	if updatedConfig.EnableAccessibilityCheck {
		t.Error("Expected EnableAccessibilityCheck false, got true")
	}

	if updatedConfig.EnableSecurityCheck {
		t.Error("Expected EnableSecurityCheck false, got true")
	}

	if !updatedConfig.EnableFreshnessCheck {
		t.Error("Expected EnableFreshnessCheck true, got false")
	}

	if updatedConfig.RequestTimeout != 30*time.Second {
		t.Errorf("Expected RequestTimeout 30s, got %v", updatedConfig.RequestTimeout)
	}

	if updatedConfig.MaxRedirects != 10 {
		t.Errorf("Expected MaxRedirects 10, got %d", updatedConfig.MaxRedirects)
	}

	if updatedConfig.UserAgent != "Custom User Agent" {
		t.Errorf("Expected UserAgent 'Custom User Agent', got %s", updatedConfig.UserAgent)
	}

	if updatedConfig.MinContentLength != 100 {
		t.Errorf("Expected MinContentLength 100, got %d", updatedConfig.MinContentLength)
	}

	if updatedConfig.MaxContentLength != 5000 {
		t.Errorf("Expected MaxContentLength 5000, got %d", updatedConfig.MaxContentLength)
	}

	if len(updatedConfig.BlockedContentKeywords) != 3 {
		t.Errorf("Expected 3 blocked content keywords, got %d", len(updatedConfig.BlockedContentKeywords))
	}

	if len(updatedConfig.ValidTLDs) != 2 {
		t.Errorf("Expected 2 valid TLDs, got %d", len(updatedConfig.ValidTLDs))
	}

	if updatedConfig.MaxValidationTime != 60*time.Second {
		t.Errorf("Expected MaxValidationTime 60s, got %v", updatedConfig.MaxValidationTime)
	}
}

func TestSearchResultValidator_GetStats(t *testing.T) {
	validator := NewSearchResultValidator()

	stats := validator.GetStats()

	if stats["total_validators"] == nil {
		t.Error("Expected total_validators in stats")
	}

	if stats["config"] == nil {
		t.Error("Expected config in stats")
	}

	totalValidators := stats["total_validators"].(int)
	if totalValidators != 6 {
		t.Errorf("Expected 6 validators, got %d", totalValidators)
	}
}

func TestValidatorPriorities(t *testing.T) {
	validator := NewSearchResultValidator()

	// Check that validators are ordered by priority
	expectedOrder := []string{
		"URLValidator",
		"ContentValidator",
		"DomainValidator",
		"AccessibilityValidator",
		"SecurityValidator",
		"FreshnessValidator",
	}

	for i, v := range validator.validators {
		if v.GetName() != expectedOrder[i] {
			t.Errorf("Expected validator %d to be %s, got %s", i, expectedOrder[i], v.GetName())
		}
	}
}
