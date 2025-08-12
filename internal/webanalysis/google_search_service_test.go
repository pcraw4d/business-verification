package webanalysis

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewGoogleCustomSearchService(t *testing.T) {
	apiKey := "test-api-key"
	searchEngineID := "test-search-engine-id"

	service := NewGoogleCustomSearchService(apiKey, searchEngineID)

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	if service.apiKey != apiKey {
		t.Errorf("Expected API key %s, got %s", apiKey, service.apiKey)
	}

	if service.searchEngineID != searchEngineID {
		t.Errorf("Expected search engine ID %s, got %s", searchEngineID, service.searchEngineID)
	}

	if service.httpClient == nil {
		t.Error("Expected HTTP client to be created")
	}

	if service.quotaManager == nil {
		t.Error("Expected quota manager to be created")
	}
}

func TestGoogleCustomSearchService_Search_Success(t *testing.T) {
	// Create a test server that returns a valid Google search response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a mock Google search response
		response := `{
			"items": [
				{
					"title": "Test Business - Official Website",
					"link": "https://testbusiness.com",
					"snippet": "Test Business is a leading company in the technology sector...",
					"displayLink": "testbusiness.com"
				},
				{
					"title": "Test Business - About Us",
					"link": "https://testbusiness.com/about",
					"snippet": "Learn more about Test Business and our mission...",
					"displayLink": "testbusiness.com"
				}
			],
			"searchInformation": {
				"searchTime": 0.123,
				"formattedSearchTime": "0.12",
				"totalResults": "1,234",
				"formattedTotalResults": "1,234"
			}
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create service and test the makeSearchRequest method directly
	service := NewGoogleCustomSearchService("test-api-key", "test-search-engine-id")
	service.httpClient = &http.Client{Timeout: 5 * time.Second}

	// Test the makeSearchRequest method directly
	ctx := context.Background()
	response, err := service.makeSearchRequest(ctx, server.URL+"/customsearch/v1?key=test&cx=test&q=test")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(response.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(response.Items))
	}

	// Check first item
	firstItem := response.Items[0]
	if firstItem.Title != "Test Business - Official Website" {
		t.Errorf("Expected title 'Test Business - Official Website', got %s", firstItem.Title)
	}

	if firstItem.Link != "https://testbusiness.com" {
		t.Errorf("Expected link 'https://testbusiness.com', got %s", firstItem.Link)
	}

	// Check search information
	if response.SearchInformation.TotalResults != "1,234" {
		t.Errorf("Expected total results '1,234', got %s", response.SearchInformation.TotalResults)
	}
}

func TestGoogleCustomSearchService_Search_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse := `{
			"error": {
				"code": 403,
				"message": "API key not valid",
				"status": "PERMISSION_DENIED"
			}
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(errorResponse))
	}))
	defer server.Close()

	// Create service
	service := NewGoogleCustomSearchService("invalid-api-key", "test-search-engine-id")
	service.httpClient = &http.Client{Timeout: 5 * time.Second}

	// Test the makeSearchRequest method directly
	ctx := context.Background()
	response, err := service.makeSearchRequest(ctx, server.URL+"/customsearch/v1?key=test&cx=test&q=test")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if response != nil {
		t.Error("Expected nil response on error")
	}

	// Check error message
	expectedError := "Google API error: API key not valid (code: 403)"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestGoogleSearchQuotaManager_CheckQuota(t *testing.T) {
	quotaManager := NewGoogleSearchQuotaManager()

	// Test initial quota check
	err := quotaManager.CheckQuota()
	if err != nil {
		t.Errorf("Expected no error on initial quota check, got %v", err)
	}

	// Test quota after multiple uses
	for i := 0; i < 10; i++ {
		quotaManager.IncrementUsage()
	}

	err = quotaManager.CheckQuota()
	if err == nil {
		t.Error("Expected error when rate limit exceeded, got nil")
	}

	if err.Error() != "rate limit exceeded" {
		t.Errorf("Expected 'rate limit exceeded' error, got %s", err.Error())
	}
}

func TestGoogleSearchQuotaManager_GetQuotaStatus(t *testing.T) {
	quotaManager := NewGoogleSearchQuotaManager()

	// Add some usage
	quotaManager.IncrementUsage()
	quotaManager.IncrementUsage()

	status := quotaManager.GetQuotaStatus()

	if status["daily_quota_used"] != 2 {
		t.Errorf("Expected daily quota used 2, got %v", status["daily_quota_used"])
	}

	if status["daily_quota_limit"] != 10000 {
		t.Errorf("Expected daily quota limit 10000, got %v", status["daily_quota_limit"])
	}

	if status["queries_this_second"] != 2 {
		t.Errorf("Expected queries this second 2, got %v", status["queries_this_second"])
	}

	if status["per_second_limit"] != 10 {
		t.Errorf("Expected per second limit 10, got %v", status["per_second_limit"])
	}

	if status["daily_quota_remaining"] != 9998 {
		t.Errorf("Expected daily quota remaining 9998, got %v", status["daily_quota_remaining"])
	}
}

func TestGoogleCustomSearchService_ValidateAPIKey(t *testing.T) {
	// Create a test server that returns success
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"items": [
				{
					"title": "Test Result",
					"link": "https://example.com",
					"snippet": "Test snippet"
				}
			],
			"searchInformation": {
				"searchTime": 0.1,
				"totalResults": "1"
			}
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create service
	service := NewGoogleCustomSearchService("valid-api-key", "test-search-engine-id")
	service.httpClient = &http.Client{Timeout: 5 * time.Second}

	// Test the makeSearchRequest method directly to simulate API key validation
	ctx := context.Background()
	response, err := service.makeSearchRequest(ctx, server.URL+"/customsearch/v1?key=test&cx=test&q=test")

	if err != nil {
		t.Errorf("Expected no error for valid API key, got %v", err)
	}

	if response == nil {
		t.Error("Expected response for valid API key")
	}
}

func TestGoogleCustomSearchService_ValidateAPIKey_Invalid(t *testing.T) {
	// Create a test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse := `{
			"error": {
				"code": 403,
				"message": "API key not valid",
				"status": "PERMISSION_DENIED"
			}
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(errorResponse))
	}))
	defer server.Close()

	// Create service
	service := NewGoogleCustomSearchService("invalid-api-key", "test-search-engine-id")
	service.httpClient = &http.Client{Timeout: 5 * time.Second}

	// Test the makeSearchRequest method directly to simulate API key validation
	ctx := context.Background()
	response, err := service.makeSearchRequest(ctx, server.URL+"/customsearch/v1?key=test&cx=test&q=test")

	if err == nil {
		t.Error("Expected error for invalid API key, got nil")
	}

	expectedError := "Google API error: API key not valid (code: 403)"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}

	if response != nil {
		t.Error("Expected nil response for invalid API key")
	}
}

func TestGoogleCustomSearchService_GetConfig(t *testing.T) {
	service := NewGoogleCustomSearchService("test-api-key", "test-search-engine-id")

	config := service.GetConfig()

	if config.MaxResultsPerQuery != 10 {
		t.Errorf("Expected MaxResultsPerQuery 10, got %d", config.MaxResultsPerQuery)
	}

	if config.MaxQueriesPerDay != 10000 {
		t.Errorf("Expected MaxQueriesPerDay 10000, got %d", config.MaxQueriesPerDay)
	}

	if config.RequestTimeout != 30*time.Second {
		t.Errorf("Expected RequestTimeout 30s, got %v", config.RequestTimeout)
	}

	if !config.EnableSafeSearch {
		t.Error("Expected EnableSafeSearch true, got false")
	}
}

func TestGoogleCustomSearchService_UpdateConfig(t *testing.T) {
	service := NewGoogleCustomSearchService("test-api-key", "test-search-engine-id")

	newConfig := GoogleCustomSearchConfig{
		MaxResultsPerQuery:    20,
		MaxQueriesPerDay:      20000,
		RequestTimeout:        60 * time.Second,
		EnableSafeSearch:      false,
		EnableImageSearch:     true,
		EnableSiteRestriction: true,
		SiteRestriction:       "example.com",
	}

	service.UpdateConfig(newConfig)

	updatedConfig := service.GetConfig()

	if updatedConfig.MaxResultsPerQuery != 20 {
		t.Errorf("Expected MaxResultsPerQuery 20, got %d", updatedConfig.MaxResultsPerQuery)
	}

	if updatedConfig.MaxQueriesPerDay != 20000 {
		t.Errorf("Expected MaxQueriesPerDay 20000, got %d", updatedConfig.MaxQueriesPerDay)
	}

	if updatedConfig.RequestTimeout != 60*time.Second {
		t.Errorf("Expected RequestTimeout 60s, got %v", updatedConfig.RequestTimeout)
	}

	if updatedConfig.EnableSafeSearch {
		t.Error("Expected EnableSafeSearch false, got true")
	}

	if !updatedConfig.EnableImageSearch {
		t.Error("Expected EnableImageSearch true, got false")
	}

	if !updatedConfig.EnableSiteRestriction {
		t.Error("Expected EnableSiteRestriction true, got false")
	}

	if updatedConfig.SiteRestriction != "example.com" {
		t.Errorf("Expected SiteRestriction 'example.com', got %s", updatedConfig.SiteRestriction)
	}
}

func TestGoogleCustomSearchService_GetQuotaStatus(t *testing.T) {
	service := NewGoogleCustomSearchService("test-api-key", "test-search-engine-id")

	// Add some usage
	service.quotaManager.IncrementUsage()
	service.quotaManager.IncrementUsage()

	status := service.GetQuotaStatus()

	if status["daily_quota_used"] != 2 {
		t.Errorf("Expected daily quota used 2, got %v", status["daily_quota_used"])
	}

	if status["daily_quota_remaining"] != 9998 {
		t.Errorf("Expected daily quota remaining 9998, got %v", status["daily_quota_remaining"])
	}
}
