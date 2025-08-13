package integrations

import (
	"context"
	"testing"
	"time"
)

func TestNewBusinessDataAPIService(t *testing.T) {
	config := BusinessDataAPIConfig{
		DefaultProvider:     "dnb",
		FallbackProvider:    "experian",
		RateLimiting:        true,
		GlobalRateLimit:     1000,
		ProviderRateLimit:   100,
		CachingEnabled:      true,
		CacheTTL:            1 * time.Hour,
		CacheSize:           10000,
		CostTracking:        true,
		BudgetLimit:         1000.0,
		CostOptimization:    true,
		DataValidation:      true,
		QualityThreshold:    0.8,
		DuplicateCheck:      true,
		MonitoringEnabled:   true,
		AlertThreshold:      0.9,
		HealthCheckInterval: 5 * time.Minute,
	}

	service := NewBusinessDataAPIService(config)

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	if service.config.DefaultProvider != "dnb" {
		t.Errorf("Expected default provider to be 'dnb', got '%s'", service.config.DefaultProvider)
	}

	if service.config.FallbackProvider != "experian" {
		t.Errorf("Expected fallback provider to be 'experian', got '%s'", service.config.FallbackProvider)
	}

	if !service.config.RateLimiting {
		t.Error("Expected rate limiting to be enabled")
	}

	if service.config.GlobalRateLimit != 1000 {
		t.Errorf("Expected global rate limit to be 1000, got %d", service.config.GlobalRateLimit)
	}

	if service.config.ProviderRateLimit != 100 {
		t.Errorf("Expected provider rate limit to be 100, got %d", service.config.ProviderRateLimit)
	}

	if !service.config.CachingEnabled {
		t.Error("Expected caching to be enabled")
	}

	if service.config.CacheTTL != 1*time.Hour {
		t.Errorf("Expected cache TTL to be 1h, got %v", service.config.CacheTTL)
	}

	if service.config.CacheSize != 10000 {
		t.Errorf("Expected cache size to be 10000, got %d", service.config.CacheSize)
	}

	if !service.config.CostTracking {
		t.Error("Expected cost tracking to be enabled")
	}

	if service.config.BudgetLimit != 1000.0 {
		t.Errorf("Expected budget limit to be 1000.0, got %f", service.config.BudgetLimit)
	}

	if !service.config.CostOptimization {
		t.Error("Expected cost optimization to be enabled")
	}

	if !service.config.DataValidation {
		t.Error("Expected data validation to be enabled")
	}

	if service.config.QualityThreshold != 0.8 {
		t.Errorf("Expected quality threshold to be 0.8, got %f", service.config.QualityThreshold)
	}

	if !service.config.DuplicateCheck {
		t.Error("Expected duplicate check to be enabled")
	}

	if !service.config.MonitoringEnabled {
		t.Error("Expected monitoring to be enabled")
	}

	if service.config.AlertThreshold != 0.9 {
		t.Errorf("Expected alert threshold to be 0.9, got %f", service.config.AlertThreshold)
	}

	if service.config.HealthCheckInterval != 5*time.Minute {
		t.Errorf("Expected health check interval to be 5m, got %v", service.config.HealthCheckInterval)
	}
}

func TestNewBusinessDataAPIServiceWithDefaults(t *testing.T) {
	config := BusinessDataAPIConfig{}

	service := NewBusinessDataAPIService(config)

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	// Check default values
	if service.config.GlobalRateLimit != 1000 {
		t.Errorf("Expected default global rate limit to be 1000, got %d", service.config.GlobalRateLimit)
	}

	if service.config.ProviderRateLimit != 100 {
		t.Errorf("Expected default provider rate limit to be 100, got %d", service.config.ProviderRateLimit)
	}

	if service.config.CacheTTL != 1*time.Hour {
		t.Errorf("Expected default cache TTL to be 1h, got %v", service.config.CacheTTL)
	}

	if service.config.CacheSize != 10000 {
		t.Errorf("Expected default cache size to be 10000, got %d", service.config.CacheSize)
	}

	if service.config.BudgetLimit != 1000.0 {
		t.Errorf("Expected default budget limit to be 1000.0, got %f", service.config.BudgetLimit)
	}

	if service.config.QualityThreshold != 0.8 {
		t.Errorf("Expected default quality threshold to be 0.8, got %f", service.config.QualityThreshold)
	}

	if service.config.AlertThreshold != 0.9 {
		t.Errorf("Expected default alert threshold to be 0.9, got %f", service.config.AlertThreshold)
	}

	if service.config.HealthCheckInterval != 5*time.Minute {
		t.Errorf("Expected default health check interval to be 5m, got %v", service.config.HealthCheckInterval)
	}
}

func TestRegisterProvider(t *testing.T) {
	config := BusinessDataAPIConfig{}
	service := NewBusinessDataAPIService(config)

	// Create a test provider
	providerConfig := ProviderConfig{
		Name:           "test-provider",
		Type:           "test",
		BaseURL:        "https://api.test.com",
		APIKey:         "test-key",
		RateLimit:      100,
		BurstLimit:     10,
		CostPerRequest: 0.01,
		Timeout:        30 * time.Second,
		DataQuality:    0.9,
	}

	provider := NewDunBradstreetProvider(providerConfig)

	err := service.RegisterProvider(provider)
	if err != nil {
		t.Fatalf("Expected to register provider successfully, got error: %v", err)
	}

	// Check if provider was registered
	registeredProvider := service.getProvider("test-provider")
	if registeredProvider == nil {
		t.Error("Expected provider to be registered, got nil")
	}

	if registeredProvider.GetName() != "test-provider" {
		t.Errorf("Expected provider name to be 'test-provider', got '%s'", registeredProvider.GetName())
	}
}

func TestSearchBusiness(t *testing.T) {
	config := BusinessDataAPIConfig{
		CachingEnabled: true,
		DataValidation: true,
		CostTracking:   true,
	}
	service := NewBusinessDataAPIService(config)

	// Register a test provider
	providerConfig := ProviderConfig{
		Name:           "test-provider",
		Type:           "test",
		BaseURL:        "https://api.test.com",
		APIKey:         "test-key",
		RateLimit:      100,
		BurstLimit:     10,
		CostPerRequest: 0.01,
		CostPerSearch:  0.02,
		Timeout:        30 * time.Second,
		DataQuality:    0.9,
		Coverage: map[string]float64{
			"US": 0.95,
		},
		Features: []string{"search", "details", "financial", "compliance", "news"},
	}

	provider := NewDunBradstreetProvider(providerConfig)
	service.RegisterProvider(provider)

	query := BusinessSearchQuery{
		CompanyName:       "Test Company",
		Country:           "US",
		Industry:          "Technology",
		IncludeFinancial:  true,
		IncludeCompliance: true,
		IncludeNews:       true,
		MaxResults:        10,
	}

	ctx := context.Background()
	result, err := service.SearchBusiness(ctx, query)

	if err != nil {
		t.Fatalf("Expected to search business successfully, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected search result, got nil")
	}

	if result.CompanyName != "Test Company" {
		t.Errorf("Expected company name to be 'Test Company', got '%s'", result.CompanyName)
	}

	if result.ProviderName != "test-provider" {
		t.Errorf("Expected provider name to be 'test-provider', got '%s'", result.ProviderName)
	}

	if result.DataQuality <= 0 {
		t.Error("Expected positive data quality score")
	}

	if result.Confidence <= 0 {
		t.Error("Expected positive confidence score")
	}
}

func TestSearchBusinessNoProvider(t *testing.T) {
	config := BusinessDataAPIConfig{}
	service := NewBusinessDataAPIService(config)

	query := BusinessSearchQuery{
		CompanyName: "Test Company",
		Country:     "US",
	}

	ctx := context.Background()
	_, err := service.SearchBusiness(ctx, query)

	if err == nil {
		t.Fatal("Expected error when no provider is available, got nil")
	}

	expectedError := "no suitable provider found for query"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestGetBusinessDetails(t *testing.T) {
	config := BusinessDataAPIConfig{
		CachingEnabled: true,
		CostTracking:   true,
	}
	service := NewBusinessDataAPIService(config)

	// Register a test provider
	providerConfig := ProviderConfig{
		Name:           "test-provider",
		Type:           "test",
		BaseURL:        "https://api.test.com",
		APIKey:         "test-key",
		RateLimit:      100,
		BurstLimit:     10,
		CostPerRequest: 0.01,
		CostPerDetail:  0.05,
		Timeout:        30 * time.Second,
		DataQuality:    0.9,
	}

	provider := NewDunBradstreetProvider(providerConfig)
	service.RegisterProvider(provider)

	businessID := "test-business-123"
	providerName := "test-provider"

	ctx := context.Background()
	result, err := service.GetBusinessDetails(ctx, businessID, providerName)

	if err != nil {
		t.Fatalf("Expected to get business details successfully, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected business details, got nil")
	}

	if result.ID != businessID {
		t.Errorf("Expected business ID to be '%s', got '%s'", businessID, result.ID)
	}

	if result.ProviderName != providerName {
		t.Errorf("Expected provider name to be '%s', got '%s'", providerName, result.ProviderName)
	}
}

func TestGetBusinessDetailsProviderNotFound(t *testing.T) {
	config := BusinessDataAPIConfig{}
	service := NewBusinessDataAPIService(config)

	businessID := "test-business-123"
	providerName := "non-existent-provider"

	ctx := context.Background()
	_, err := service.GetBusinessDetails(ctx, businessID, providerName)

	if err == nil {
		t.Fatal("Expected error when provider not found, got nil")
	}

	expectedError := "provider non-existent-provider not found"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestGetFinancialData(t *testing.T) {
	config := BusinessDataAPIConfig{
		CachingEnabled: true,
		CostTracking:   true,
	}
	service := NewBusinessDataAPIService(config)

	// Register a test provider
	providerConfig := ProviderConfig{
		Name:             "test-provider",
		Type:             "test",
		BaseURL:          "https://api.test.com",
		APIKey:           "test-key",
		RateLimit:        100,
		BurstLimit:       10,
		CostPerRequest:   0.01,
		CostPerFinancial: 0.10,
		Timeout:          30 * time.Second,
		DataQuality:      0.9,
	}

	provider := NewDunBradstreetProvider(providerConfig)
	service.RegisterProvider(provider)

	businessID := "test-business-123"
	providerName := "test-provider"

	ctx := context.Background()
	result, err := service.GetFinancialData(ctx, businessID, providerName)

	if err != nil {
		t.Fatalf("Expected to get financial data successfully, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected financial data, got nil")
	}

	if result.FiscalYear != 2023 {
		t.Errorf("Expected fiscal year to be 2023, got %d", result.FiscalYear)
	}

	if result.Revenue <= 0 {
		t.Error("Expected positive revenue")
	}

	if result.NetIncome <= 0 {
		t.Error("Expected positive net income")
	}

	if result.TotalAssets <= 0 {
		t.Error("Expected positive total assets")
	}
}

func TestGetComplianceData(t *testing.T) {
	config := BusinessDataAPIConfig{
		CostTracking: true,
	}
	service := NewBusinessDataAPIService(config)

	// Register a test provider
	providerConfig := ProviderConfig{
		Name:           "test-provider",
		Type:           "test",
		BaseURL:        "https://api.test.com",
		APIKey:         "test-key",
		RateLimit:      100,
		BurstLimit:     10,
		CostPerRequest: 0.01,
		Timeout:        30 * time.Second,
		DataQuality:    0.9,
	}

	provider := NewDunBradstreetProvider(providerConfig)
	service.RegisterProvider(provider)

	businessID := "test-business-123"
	providerName := "test-provider"

	ctx := context.Background()
	result, err := service.GetComplianceData(ctx, businessID, providerName)

	if err != nil {
		t.Fatalf("Expected to get compliance data successfully, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected compliance data, got nil")
	}

	if result.RegulatoryStatus != "compliant" {
		t.Errorf("Expected regulatory status to be 'compliant', got '%s'", result.RegulatoryStatus)
	}

	if result.ComplianceScore <= 0 {
		t.Error("Expected positive compliance score")
	}

	if result.RiskLevel != "low" {
		t.Errorf("Expected risk level to be 'low', got '%s'", result.RiskLevel)
	}
}

func TestGetNewsData(t *testing.T) {
	config := BusinessDataAPIConfig{
		CostTracking: true,
	}
	service := NewBusinessDataAPIService(config)

	// Register a test provider
	providerConfig := ProviderConfig{
		Name:           "test-provider",
		Type:           "test",
		BaseURL:        "https://api.test.com",
		APIKey:         "test-key",
		RateLimit:      100,
		BurstLimit:     10,
		CostPerRequest: 0.01,
		Timeout:        30 * time.Second,
		DataQuality:    0.9,
	}

	provider := NewDunBradstreetProvider(providerConfig)
	service.RegisterProvider(provider)

	businessID := "test-business-123"
	providerName := "test-provider"

	ctx := context.Background()
	result, err := service.GetNewsData(ctx, businessID, providerName)

	if err != nil {
		t.Fatalf("Expected to get news data successfully, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected news data, got nil")
	}

	if len(result) == 0 {
		t.Error("Expected non-empty news data")
	}

	newsItem := result[0]
	if newsItem.Title == "" {
		t.Error("Expected non-empty news title")
	}

	if newsItem.Source == "" {
		t.Error("Expected non-empty news source")
	}

	if newsItem.PublishedDate.IsZero() {
		t.Error("Expected non-zero published date")
	}
}

func TestProviderSelection(t *testing.T) {
	config := BusinessDataAPIConfig{
		CostOptimization: true,
	}
	service := NewBusinessDataAPIService(config)

	// Register multiple providers with different characteristics
	dnbConfig := ProviderConfig{
		Name:          "dnb",
		Type:          "dnb",
		DataQuality:   0.95,
		CostPerSearch: 0.02,
		Coverage: map[string]float64{
			"US": 0.95,
		},
		Features: []string{"search", "details", "financial", "compliance", "news"},
	}

	experianConfig := ProviderConfig{
		Name:          "experian",
		Type:          "experian",
		DataQuality:   0.93,
		CostPerSearch: 0.01,
		Coverage: map[string]float64{
			"US": 0.90,
		},
		Features: []string{"search", "details", "financial"},
	}

	dnbProvider := NewDunBradstreetProvider(dnbConfig)
	experianProvider := NewExperianProvider(experianConfig)

	service.RegisterProvider(dnbProvider)
	service.RegisterProvider(experianProvider)

	query := BusinessSearchQuery{
		CompanyName:       "Test Company",
		Country:           "US",
		IncludeFinancial:  true,
		IncludeCompliance: true,
		IncludeNews:       true,
	}

	// Test provider selection
	selectedProvider := service.selectBestProvider(query)
	if selectedProvider == nil {
		t.Fatal("Expected provider to be selected, got nil")
	}

	// D&B should be selected due to better feature coverage
	if selectedProvider.GetName() != "dnb" {
		t.Errorf("Expected D&B provider to be selected, got '%s'", selectedProvider.GetName())
	}
}

func TestRateLimiting(t *testing.T) {
	config := BusinessDataAPIConfig{
		RateLimiting: true,
	}
	service := NewBusinessDataAPIService(config)

	// Register a provider with low rate limit
	providerConfig := ProviderConfig{
		Name:        "test-provider",
		Type:        "test",
		RateLimit:   1, // 1 request per minute
		BurstLimit:  1,
		Timeout:     30 * time.Second,
		DataQuality: 0.9,
	}

	provider := NewDunBradstreetProvider(providerConfig)
	service.RegisterProvider(provider)

	// First request should succeed
	if !service.checkRateLimit("test-provider") {
		t.Error("Expected first request to be allowed")
	}

	// Second request should be rate limited
	if service.checkRateLimit("test-provider") {
		t.Error("Expected second request to be rate limited")
	}
}

func TestCaching(t *testing.T) {
	config := BusinessDataAPIConfig{
		CachingEnabled: true,
		CacheTTL:       1 * time.Hour,
		CacheSize:      1000,
	}
	service := NewBusinessDataAPIService(config)

	// Test cache operations
	cacheKey := "test-key"
	testData := "test-value"

	// Set cache entry
	service.cache.Set(cacheKey, testData, 1*time.Hour)

	// Get cache entry
	cachedData := service.cache.Get(cacheKey)
	if cachedData == nil {
		t.Error("Expected cached data, got nil")
	}

	if cachedData.(string) != testData {
		t.Errorf("Expected cached data to be '%s', got '%s'", testData, cachedData.(string))
	}

	// Test cache expiration
	expiredKey := "expired-key"
	service.cache.Set(expiredKey, testData, 1*time.Nanosecond)
	time.Sleep(1 * time.Millisecond)

	expiredData := service.cache.Get(expiredKey)
	if expiredData != nil {
		t.Error("Expected expired data to be nil")
	}
}

func TestDunBradstreetProvider(t *testing.T) {
	config := ProviderConfig{
		Name:           "dnb",
		Type:           "dnb",
		BaseURL:        "https://api.dnb.com",
		APIKey:         "test-key",
		RateLimit:      100,
		BurstLimit:     10,
		CostPerRequest: 0.01,
		Timeout:        30 * time.Second,
		DataQuality:    0.95,
	}

	provider := NewDunBradstreetProvider(config)

	if provider.GetName() != "dnb" {
		t.Errorf("Expected provider name to be 'dnb', got '%s'", provider.GetName())
	}

	if provider.GetType() != "dnb" {
		t.Errorf("Expected provider type to be 'dnb', got '%s'", provider.GetType())
	}

	if !provider.IsHealthy() {
		t.Error("Expected provider to be healthy")
	}

	if provider.GetCost() != 0.01 {
		t.Errorf("Expected cost to be 0.01, got %f", provider.GetCost())
	}

	quota := provider.GetQuota()
	if quota.DailyLimit != 1000 {
		t.Errorf("Expected daily limit to be 1000, got %d", quota.DailyLimit)
	}

	// Test search
	query := BusinessSearchQuery{
		CompanyName: "Test Company",
		Country:     "US",
	}

	ctx := context.Background()
	result, err := provider.SearchBusiness(ctx, query)
	if err != nil {
		t.Fatalf("Expected search to succeed, got error: %v", err)
	}

	if result.CompanyName != "Test Company" {
		t.Errorf("Expected company name to be 'Test Company', got '%s'", result.CompanyName)
	}

	if result.ProviderName != "dnb" {
		t.Errorf("Expected provider name to be 'dnb', got '%s'", result.ProviderName)
	}
}

func TestExperianProvider(t *testing.T) {
	config := ProviderConfig{
		Name:           "experian",
		Type:           "experian",
		BaseURL:        "https://api.experian.com",
		APIKey:         "test-key",
		RateLimit:      120,
		BurstLimit:     12,
		CostPerRequest: 0.015,
		Timeout:        30 * time.Second,
		DataQuality:    0.93,
	}

	provider := NewExperianProvider(config)

	if provider.GetName() != "experian" {
		t.Errorf("Expected provider name to be 'experian', got '%s'", provider.GetName())
	}

	if provider.GetType() != "experian" {
		t.Errorf("Expected provider type to be 'experian', got '%s'", provider.GetType())
	}

	if !provider.IsHealthy() {
		t.Error("Expected provider to be healthy")
	}

	if provider.GetCost() != 0.015 {
		t.Errorf("Expected cost to be 0.015, got %f", provider.GetCost())
	}

	quota := provider.GetQuota()
	if quota.DailyLimit != 1200 {
		t.Errorf("Expected daily limit to be 1200, got %d", quota.DailyLimit)
	}

	// Test search
	query := BusinessSearchQuery{
		CompanyName: "Test Company",
		Country:     "US",
	}

	ctx := context.Background()
	result, err := provider.SearchBusiness(ctx, query)
	if err != nil {
		t.Fatalf("Expected search to succeed, got error: %v", err)
	}

	if result.CompanyName != "Test Company" {
		t.Errorf("Expected company name to be 'Test Company', got '%s'", result.CompanyName)
	}

	if result.ProviderName != "experian" {
		t.Errorf("Expected provider name to be 'experian', got '%s'", result.ProviderName)
	}
}

func TestSECProvider(t *testing.T) {
	config := ProviderConfig{
		Name:           "sec",
		Type:           "sec",
		BaseURL:        "https://api.sec.gov",
		APIKey:         "test-key",
		RateLimit:      500,
		BurstLimit:     50,
		CostPerRequest: 0.005,
		Timeout:        30 * time.Second,
		DataQuality:    0.98,
	}

	provider := NewSECProvider(config)

	if provider.GetName() != "sec" {
		t.Errorf("Expected provider name to be 'sec', got '%s'", provider.GetName())
	}

	if provider.GetType() != "sec" {
		t.Errorf("Expected provider type to be 'sec', got '%s'", provider.GetType())
	}

	if !provider.IsHealthy() {
		t.Error("Expected provider to be healthy")
	}

	if provider.GetCost() != 0.005 {
		t.Errorf("Expected cost to be 0.005, got %f", provider.GetCost())
	}

	quota := provider.GetQuota()
	if quota.DailyLimit != 5000 {
		t.Errorf("Expected daily limit to be 5000, got %d", quota.DailyLimit)
	}

	// Test search
	query := BusinessSearchQuery{
		CompanyName: "Test Company",
		Country:     "US",
	}

	ctx := context.Background()
	result, err := provider.SearchBusiness(ctx, query)
	if err != nil {
		t.Fatalf("Expected search to succeed, got error: %v", err)
	}

	if result.CompanyName != "Test Company" {
		t.Errorf("Expected company name to be 'Test Company', got '%s'", result.CompanyName)
	}

	if result.ProviderName != "sec" {
		t.Errorf("Expected provider name to be 'sec', got '%s'", result.ProviderName)
	}
}

func TestBloombergProvider(t *testing.T) {
	config := ProviderConfig{
		Name:           "bloomberg",
		Type:           "bloomberg",
		BaseURL:        "https://api.bloomberg.com",
		APIKey:         "test-key",
		RateLimit:      200,
		BurstLimit:     20,
		CostPerRequest: 0.02,
		Timeout:        30 * time.Second,
		DataQuality:    0.96,
	}

	provider := NewBloombergProvider(config)

	if provider.GetName() != "bloomberg" {
		t.Errorf("Expected provider name to be 'bloomberg', got '%s'", provider.GetName())
	}

	if provider.GetType() != "bloomberg" {
		t.Errorf("Expected provider type to be 'bloomberg', got '%s'", provider.GetType())
	}

	if !provider.IsHealthy() {
		t.Error("Expected provider to be healthy")
	}

	if provider.GetCost() != 0.02 {
		t.Errorf("Expected cost to be 0.02, got %f", provider.GetCost())
	}

	quota := provider.GetQuota()
	if quota.DailyLimit != 2000 {
		t.Errorf("Expected daily limit to be 2000, got %d", quota.DailyLimit)
	}

	// Test search
	query := BusinessSearchQuery{
		CompanyName: "Test Company",
		Country:     "UK",
	}

	ctx := context.Background()
	result, err := provider.SearchBusiness(ctx, query)
	if err != nil {
		t.Fatalf("Expected search to succeed, got error: %v", err)
	}

	if result.CompanyName != "Test Company" {
		t.Errorf("Expected company name to be 'Test Company', got '%s'", result.CompanyName)
	}

	if result.ProviderName != "bloomberg" {
		t.Errorf("Expected provider name to be 'bloomberg', got '%s'", result.ProviderName)
	}
}

func TestFactivaProvider(t *testing.T) {
	config := ProviderConfig{
		Name:           "factiva",
		Type:           "factiva",
		BaseURL:        "https://api.factiva.com",
		APIKey:         "test-key",
		RateLimit:      300,
		BurstLimit:     30,
		CostPerRequest: 0.01,
		Timeout:        30 * time.Second,
		DataQuality:    0.85,
	}

	provider := NewFactivaProvider(config)

	if provider.GetName() != "factiva" {
		t.Errorf("Expected provider name to be 'factiva', got '%s'", provider.GetName())
	}

	if provider.GetType() != "factiva" {
		t.Errorf("Expected provider type to be 'factiva', got '%s'", provider.GetType())
	}

	if !provider.IsHealthy() {
		t.Error("Expected provider to be healthy")
	}

	if provider.GetCost() != 0.01 {
		t.Errorf("Expected cost to be 0.01, got %f", provider.GetCost())
	}

	quota := provider.GetQuota()
	if quota.DailyLimit != 3000 {
		t.Errorf("Expected daily limit to be 3000, got %d", quota.DailyLimit)
	}

	// Test news data (Factiva's primary feature)
	businessID := "test-business-123"
	ctx := context.Background()
	newsData, err := provider.GetNewsData(ctx, businessID)
	if err != nil {
		t.Fatalf("Expected news data to succeed, got error: %v", err)
	}

	if len(newsData) == 0 {
		t.Error("Expected non-empty news data")
	}

	// Test that financial data is not available
	_, err = provider.GetFinancialData(ctx, businessID)
	if err == nil {
		t.Error("Expected error for financial data, got nil")
	}

	expectedError := "financial data not available from Factiva"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestDataValidation(t *testing.T) {
	config := ProviderConfig{
		Name:        "test-provider",
		Type:        "test",
		DataQuality: 0.95,
	}

	provider := NewDunBradstreetProvider(config)

	// Test valid data
	validData := &BusinessData{
		CompanyName:    "Valid Company",
		BusinessNumber: "123456789",
		Address: Address{
			Street1: "123 Main St",
			City:    "New York",
			State:   "NY",
			Country: "US",
		},
	}

	validation, err := provider.ValidateData(validData)
	if err != nil {
		t.Fatalf("Expected validation to succeed, got error: %v", err)
	}

	if !validation.IsValid {
		t.Error("Expected valid data to pass validation")
	}

	if validation.QualityScore <= 0 {
		t.Error("Expected positive quality score")
	}

	// Test invalid data
	invalidData := &BusinessData{
		// Missing company name
		BusinessNumber: "123456789",
	}

	validation, err = provider.ValidateData(invalidData)
	if err != nil {
		t.Fatalf("Expected validation to succeed, got error: %v", err)
	}

	if validation.IsValid {
		t.Error("Expected invalid data to fail validation")
	}

	if len(validation.Issues) == 0 {
		t.Error("Expected validation issues to be found")
	}

	// Check for specific issue
	foundIssue := false
	for _, issue := range validation.Issues {
		if issue.Field == "company_name" && issue.Type == "missing" {
			foundIssue = true
			break
		}
	}

	if !foundIssue {
		t.Error("Expected missing company name issue to be found")
	}
}

func TestGenerateCacheKey(t *testing.T) {
	// Test simple key
	key1 := generateCacheKey("search", "test-query")
	if key1 != "search:test-query" {
		t.Errorf("Expected key 'search:test-query', got '%s'", key1)
	}

	// Test key with multiple parameters
	key2 := generateCacheKey("details", "business-123", "provider-name")
	if key2 != "details:business-123:provider-name" {
		t.Errorf("Expected key 'details:business-123:provider-name', got '%s'", key2)
	}

	// Test key with complex object
	query := BusinessSearchQuery{
		CompanyName: "Test Company",
		Country:     "US",
	}
	key3 := generateCacheKey("search", query)
	if key3 == "" {
		t.Error("Expected non-empty cache key")
	}
}

func TestContains(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	// Test existing item
	if !contains(slice, "banana") {
		t.Error("Expected 'banana' to be found in slice")
	}

	// Test non-existing item
	if contains(slice, "orange") {
		t.Error("Expected 'orange' not to be found in slice")
	}

	// Test empty slice
	emptySlice := []string{}
	if contains(emptySlice, "apple") {
		t.Error("Expected item not to be found in empty slice")
	}
}
