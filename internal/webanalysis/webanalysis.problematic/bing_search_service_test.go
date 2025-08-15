package webanalysis

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewBingSearchService(t *testing.T) {
	apiKey := "test-bing-api-key"

	service := NewBingSearchService(apiKey)

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	if service.apiKey != apiKey {
		t.Errorf("Expected API key %s, got %s", apiKey, service.apiKey)
	}

	if service.httpClient == nil {
		t.Error("Expected HTTP client to be created")
	}

	if service.quotaManager == nil {
		t.Error("Expected quota manager to be created")
	}
}

func TestBingSearchService_Search_Success(t *testing.T) {
	// Create a test server that returns a valid Bing search response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a mock Bing search response
		response := `{
			"_type": "SearchResponse",
			"queryContext": {
				"originalQuery": "test business"
			},
			"webPages": {
				"_type": "SearchResponse",
				"webSearchUrl": "https://www.bing.com/search?q=test+business",
				"totalEstimatedMatches": 1234,
				"value": [
					{
						"id": "https://api.bing.microsoft.com/api/v7/#Webpages.0",
						"name": "Test Business - Official Website",
						"url": "https://testbusiness.com",
						"displayUrl": "testbusiness.com",
						"snippet": "Test Business is a leading company in the technology sector...",
						"dateLastCrawled": "2023-01-01T00:00:00.0000000Z"
					},
					{
						"id": "https://api.bing.microsoft.com/api/v7/#Webpages.1",
						"name": "Test Business - About Us",
						"url": "https://testbusiness.com/about",
						"displayUrl": "testbusiness.com/about",
						"snippet": "Learn more about Test Business and our mission...",
						"dateLastCrawled": "2023-01-01T00:00:00.0000000Z"
					}
				]
			}
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create service and test the makeSearchRequest method directly
	service := NewBingSearchService("test-bing-api-key")
	service.httpClient = &http.Client{Timeout: 5 * time.Second}

	// Test the makeSearchRequest method directly
	ctx := context.Background()
	response, err := service.makeSearchRequest(ctx, server.URL+"/v7.0/search?q=test&count=10")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.WebPages.Value == nil {
		t.Fatal("Expected web pages, got nil")
	}

	if len(response.WebPages.Value) != 2 {
		t.Errorf("Expected 2 items, got %d", len(response.WebPages.Value))
	}

	// Check first item
	firstItem := response.WebPages.Value[0]
	if firstItem.Name != "Test Business - Official Website" {
		t.Errorf("Expected name 'Test Business - Official Website', got %s", firstItem.Name)
	}

	if firstItem.URL != "https://testbusiness.com" {
		t.Errorf("Expected URL 'https://testbusiness.com', got %s", firstItem.URL)
	}

	// Check search information
	if response.WebPages.TotalEstimatedMatches != 1234 {
		t.Errorf("Expected total estimated matches 1234, got %d", response.WebPages.TotalEstimatedMatches)
	}
}

func TestBingSearchService_Search_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse := `{
			"error": {
				"code": 401,
				"message": "Access denied due to invalid subscription key",
				"status": "UNAUTHORIZED"
			}
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(errorResponse))
	}))
	defer server.Close()

	// Create service
	service := NewBingSearchService("invalid-bing-api-key")
	service.httpClient = &http.Client{Timeout: 5 * time.Second}

	// Test the makeSearchRequest method directly
	ctx := context.Background()
	response, err := service.makeSearchRequest(ctx, server.URL+"/v7.0/search?q=test&count=10")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if response != nil {
		t.Error("Expected nil response on error")
	}

	// Check error message
	expectedError := "Bing API error: Access denied due to invalid subscription key (code: 401)"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestBingSearchQuotaManager_CheckQuota(t *testing.T) {
	quotaManager := NewBingSearchQuotaManager()

	// Test initial quota check
	err := quotaManager.CheckQuota()
	if err != nil {
		t.Errorf("Expected no error on initial quota check, got %v", err)
	}

	// Test quota after multiple uses
	for i := 0; i < 3; i++ {
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

func TestBingSearchQuotaManager_GetQuotaStatus(t *testing.T) {
	quotaManager := NewBingSearchQuotaManager()

	// Add some usage
	quotaManager.IncrementUsage()
	quotaManager.IncrementUsage()

	status := quotaManager.GetQuotaStatus()

	if status["daily_quota_used"] != 2 {
		t.Errorf("Expected daily quota used 2, got %v", status["daily_quota_used"])
	}

	if status["daily_quota_limit"] != 3000 {
		t.Errorf("Expected daily quota limit 3000, got %v", status["daily_quota_limit"])
	}

	if status["queries_this_second"] != 2 {
		t.Errorf("Expected queries this second 2, got %v", status["queries_this_second"])
	}

	if status["per_second_limit"] != 3 {
		t.Errorf("Expected per second limit 3, got %v", status["per_second_limit"])
	}

	if status["daily_quota_remaining"] != 2998 {
		t.Errorf("Expected daily quota remaining 2998, got %v", status["daily_quota_remaining"])
	}
}

func TestBingSearchService_ValidateAPIKey(t *testing.T) {
	// Create a test server that returns success
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"_type": "SearchResponse",
			"webPages": {
				"_type": "SearchResponse",
				"value": [
					{
						"id": "https://api.bing.microsoft.com/api/v7/#Webpages.0",
						"name": "Test Result",
						"url": "https://example.com",
						"snippet": "Test snippet"
					}
				]
			}
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create service
	service := NewBingSearchService("valid-bing-api-key")
	service.httpClient = &http.Client{Timeout: 5 * time.Second}

	// Test the makeSearchRequest method directly to simulate API key validation
	ctx := context.Background()
	response, err := service.makeSearchRequest(ctx, server.URL+"/v7.0/search?q=test&count=1")

	if err != nil {
		t.Errorf("Expected no error for valid API key, got %v", err)
	}

	if response == nil {
		t.Error("Expected response for valid API key")
	}
}

func TestBingSearchService_ValidateAPIKey_Invalid(t *testing.T) {
	// Create a test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse := `{
			"error": {
				"code": 401,
				"message": "Access denied due to invalid subscription key",
				"status": "UNAUTHORIZED"
			}
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(errorResponse))
	}))
	defer server.Close()

	// Create service
	service := NewBingSearchService("invalid-bing-api-key")
	service.httpClient = &http.Client{Timeout: 5 * time.Second}

	// Test the makeSearchRequest method directly to simulate API key validation
	ctx := context.Background()
	response, err := service.makeSearchRequest(ctx, server.URL+"/v7.0/search?q=test&count=1")

	if err == nil {
		t.Error("Expected error for invalid API key, got nil")
	}

	expectedError := "Bing API error: Access denied due to invalid subscription key (code: 401)"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}

	if response != nil {
		t.Error("Expected nil response for invalid API key")
	}
}

func TestBingSearchService_GetConfig(t *testing.T) {
	service := NewBingSearchService("test-bing-api-key")

	config := service.GetConfig()

	if config.MaxResultsPerQuery != 10 {
		t.Errorf("Expected MaxResultsPerQuery 10, got %d", config.MaxResultsPerQuery)
	}

	if config.MaxQueriesPerDay != 3000 {
		t.Errorf("Expected MaxQueriesPerDay 3000, got %d", config.MaxQueriesPerDay)
	}

	if config.RequestTimeout != 30*time.Second {
		t.Errorf("Expected RequestTimeout 30s, got %v", config.RequestTimeout)
	}

	if !config.EnableSafeSearch {
		t.Error("Expected EnableSafeSearch true, got false")
	}

	if config.Market != "en-US" {
		t.Errorf("Expected Market 'en-US', got %s", config.Market)
	}
}

func TestBingSearchService_UpdateConfig(t *testing.T) {
	service := NewBingSearchService("test-bing-api-key")

	newConfig := BingCustomSearchConfig{
		MaxResultsPerQuery: 20,
		MaxQueriesPerDay:   5000,
		RequestTimeout:     60 * time.Second,
		EnableSafeSearch:   false,
		EnableImageSearch:  true,
		EnableNewsSearch:   true,
		EnableVideoSearch:  true,
		EnableSpellCheck:   false,
		EnableSuggestions:  false,
		Market:             "en-GB",
		ResponseFilter:     "Webpages,News",
		SafeSearch:         "Strict",
		SetLang:            "en-GB",
		TextFormat:         "HTML",
	}

	service.UpdateConfig(newConfig)

	updatedConfig := service.GetConfig()

	if updatedConfig.MaxResultsPerQuery != 20 {
		t.Errorf("Expected MaxResultsPerQuery 20, got %d", updatedConfig.MaxResultsPerQuery)
	}

	if updatedConfig.MaxQueriesPerDay != 5000 {
		t.Errorf("Expected MaxQueriesPerDay 5000, got %d", updatedConfig.MaxQueriesPerDay)
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

	if !updatedConfig.EnableNewsSearch {
		t.Error("Expected EnableNewsSearch true, got false")
	}

	if !updatedConfig.EnableVideoSearch {
		t.Error("Expected EnableVideoSearch true, got false")
	}

	if updatedConfig.Market != "en-GB" {
		t.Errorf("Expected Market 'en-GB', got %s", updatedConfig.Market)
	}

	if updatedConfig.ResponseFilter != "Webpages,News" {
		t.Errorf("Expected ResponseFilter 'Webpages,News', got %s", updatedConfig.ResponseFilter)
	}

	if updatedConfig.SafeSearch != "Strict" {
		t.Errorf("Expected SafeSearch 'Strict', got %s", updatedConfig.SafeSearch)
	}

	if updatedConfig.SetLang != "en-GB" {
		t.Errorf("Expected SetLang 'en-GB', got %s", updatedConfig.SetLang)
	}

	if updatedConfig.TextFormat != "HTML" {
		t.Errorf("Expected TextFormat 'HTML', got %s", updatedConfig.TextFormat)
	}
}

func TestBingSearchService_GetQuotaStatus(t *testing.T) {
	service := NewBingSearchService("test-bing-api-key")

	// Add some usage
	service.quotaManager.IncrementUsage()
	service.quotaManager.IncrementUsage()

	status := service.GetQuotaStatus()

	if status["daily_quota_used"] != 2 {
		t.Errorf("Expected daily quota used 2, got %v", status["daily_quota_used"])
	}

	if status["daily_quota_remaining"] != 2998 {
		t.Errorf("Expected daily quota remaining 2998, got %v", status["daily_quota_remaining"])
	}
}
