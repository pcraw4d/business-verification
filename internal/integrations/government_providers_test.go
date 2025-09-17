package integrations

import (
	"context"
	"log"
	"os"
	"testing"
	"time"
)

func TestGovernmentProvidersIntegration(t *testing.T) {
	// Skip if running in CI without API keys
	if os.Getenv("CI") == "true" && os.Getenv("COMPANIES_HOUSE_API_KEY") == "" {
		t.Skip("Skipping government API tests in CI without API keys")
	}
	
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	factory := NewGovernmentProvidersFactory(logger)
	
	t.Run("SEC EDGAR Provider", func(t *testing.T) {
		provider := factory.CreateSECEdgarProvider()
		
		// Test provider configuration
		if provider.GetName() != "SEC EDGAR" {
			t.Errorf("Expected provider name 'SEC EDGAR', got '%s'", provider.GetName())
		}
		
		if provider.GetType() != "sec_edgar" {
			t.Errorf("Expected provider type 'sec_edgar', got '%s'", provider.GetType())
		}
		
		if provider.GetCost() != 0.0 {
			t.Errorf("Expected cost 0.0, got %f", provider.GetCost())
		}
		
		if !provider.IsHealthy() {
			t.Error("Expected provider to be healthy")
		}
		
		// Test quota info
		quota := provider.GetQuota()
		if quota.DailyLimit == 0 {
			t.Error("Expected non-zero daily limit")
		}
		
		// Test business search (this will fail without real API, but tests the structure)
		ctx := context.Background()
		query := BusinessSearchQuery{
			CompanyName: "Apple Inc",
			Country:     "US",
		}
		
		_, err := provider.SearchBusiness(ctx, query)
		// We expect this to fail in test environment, but it should fail gracefully
		if err == nil {
			t.Log("SEC EDGAR search succeeded (unexpected in test environment)")
		} else {
			t.Logf("SEC EDGAR search failed as expected: %v", err)
		}
	})
	
	t.Run("Companies House Provider", func(t *testing.T) {
		apiKey := os.Getenv("COMPANIES_HOUSE_API_KEY")
		provider := factory.CreateCompaniesHouseProvider(apiKey)
		
		// Test provider configuration
		if provider.GetName() != "Companies House" {
			t.Errorf("Expected provider name 'Companies House', got '%s'", provider.GetName())
		}
		
		if provider.GetType() != "companies_house" {
			t.Errorf("Expected provider type 'companies_house', got '%s'", provider.GetType())
		}
		
		if provider.GetCost() != 0.0 {
			t.Errorf("Expected cost 0.0, got %f", provider.GetCost())
		}
		
		if !provider.IsHealthy() {
			t.Error("Expected provider to be healthy")
		}
		
		// Test quota info
		quota := provider.GetQuota()
		if quota.DailyLimit == 0 {
			t.Error("Expected non-zero daily limit")
		}
		
		// Test business search
		ctx := context.Background()
		query := BusinessSearchQuery{
			CompanyName: "Apple",
			Country:     "GB",
		}
		
		_, err := provider.SearchBusiness(ctx, query)
		if apiKey == "" {
			// Should fail without API key
			if err == nil {
				t.Error("Expected Companies House search to fail without API key")
			}
		} else {
			// May succeed or fail depending on API availability
			if err != nil {
				t.Logf("Companies House search failed: %v", err)
			} else {
				t.Log("Companies House search succeeded")
			}
		}
	})
	
	t.Run("OpenCorporates Provider", func(t *testing.T) {
		apiToken := os.Getenv("OPENCORPORATES_API_TOKEN")
		provider := factory.CreateOpenCorporatesProvider(apiToken)
		
		// Test provider configuration
		if provider.GetName() != "OpenCorporates" {
			t.Errorf("Expected provider name 'OpenCorporates', got '%s'", provider.GetName())
		}
		
		if provider.GetType() != "opencorporates" {
			t.Errorf("Expected provider type 'opencorporates', got '%s'", provider.GetType())
		}
		
		if provider.GetCost() != 0.0 {
			t.Errorf("Expected cost 0.0, got %f", provider.GetCost())
		}
		
		if !provider.IsHealthy() {
			t.Error("Expected provider to be healthy")
		}
		
		// Test quota info
		quota := provider.GetQuota()
		if quota.DailyLimit != 500 {
			t.Errorf("Expected daily limit 500, got %d", quota.DailyLimit)
		}
		
		// Test business search
		ctx := context.Background()
		query := BusinessSearchQuery{
			CompanyName: "Apple Inc",
			Country:     "US",
		}
		
		_, err := provider.SearchBusiness(ctx, query)
		// May succeed or fail depending on API availability and rate limits
		if err != nil {
			t.Logf("OpenCorporates search failed: %v", err)
		} else {
			t.Log("OpenCorporates search succeeded")
		}
	})
	
	t.Run("WHOIS Provider", func(t *testing.T) {
		provider := factory.CreateWHOISProvider()
		
		// Test provider configuration
		if provider.GetName() != "WHOIS" {
			t.Errorf("Expected provider name 'WHOIS', got '%s'", provider.GetName())
		}
		
		if provider.GetType() != "whois" {
			t.Errorf("Expected provider type 'whois', got '%s'", provider.GetType())
		}
		
		if provider.GetCost() != 0.0 {
			t.Errorf("Expected cost 0.0, got %f", provider.GetCost())
		}
		
		if !provider.IsHealthy() {
			t.Error("Expected provider to be healthy")
		}
		
		// Test quota info
		quota := provider.GetQuota()
		if quota.DailyLimit == 0 {
			t.Error("Expected non-zero daily limit")
		}
		
		// Test business search with website
		ctx := context.Background()
		query := BusinessSearchQuery{
			CompanyName: "Apple Inc",
			Website:     "https://www.apple.com",
			Country:     "US",
		}
		
		_, err := provider.SearchBusiness(ctx, query)
		// Should succeed with website provided
		if err != nil {
			t.Logf("WHOIS search failed: %v", err)
		} else {
			t.Log("WHOIS search succeeded")
		}
		
		// Test business search without website (should fail)
		queryNoWebsite := BusinessSearchQuery{
			CompanyName: "Apple Inc",
			Country:     "US",
		}
		
		_, err = provider.SearchBusiness(ctx, queryNoWebsite)
		if err == nil {
			t.Error("Expected WHOIS search to fail without website")
		}
	})
}

func TestGovernmentProvidersFactory(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	factory := NewGovernmentProvidersFactory(logger)
	
	t.Run("Create All Providers", func(t *testing.T) {
		// Test creating all providers
		secProvider := factory.CreateSECEdgarProvider()
		if secProvider == nil {
			t.Error("Failed to create SEC EDGAR provider")
		}
		
		chProvider := factory.CreateCompaniesHouseProvider("test-key")
		if chProvider == nil {
			t.Error("Failed to create Companies House provider")
		}
		
		ocProvider := factory.CreateOpenCorporatesProvider("test-token")
		if ocProvider == nil {
			t.Error("Failed to create OpenCorporates provider")
		}
		
		whoisProvider := factory.CreateWHOISProvider()
		if whoisProvider == nil {
			t.Error("Failed to create WHOIS provider")
		}
	})
	
	t.Run("Register All Providers", func(t *testing.T) {
		// Create a test service
		config := BusinessDataAPIConfig{
			CachingEnabled: true,
			CacheTTL:       1 * time.Hour,
			CacheSize:      1000,
		}
		service := NewBusinessDataAPIService(config)
		
		// Test registering all providers
		apiConfig := GovernmentAPIsConfig{
			CompaniesHouseAPIKey:    os.Getenv("COMPANIES_HOUSE_API_KEY"),
			OpenCorporatesAPIToken: os.Getenv("OPENCORPORATES_API_TOKEN"),
		}
		
		err := factory.RegisterAllGovernmentProviders(service, apiConfig)
		if err != nil {
			t.Logf("Some providers failed to register: %v", err)
		}
		
		// Verify providers were registered
		// Note: We can't directly access the providers map, but we can test
		// that the registration didn't fail completely
	})
}

func TestGovernmentAPIsConfig(t *testing.T) {
	t.Run("Default Config", func(t *testing.T) {
		config := GetDefaultGovernmentAPIsConfig()
		
		if config.CompaniesHouseAPIKey != "" {
			t.Error("Expected empty Companies House API key in default config")
		}
		
		if config.OpenCorporatesAPIToken != "" {
			t.Error("Expected empty OpenCorporates API token in default config")
		}
	})
	
	t.Run("Config Validation", func(t *testing.T) {
		config := GetDefaultGovernmentAPIsConfig()
		warnings := ValidateGovernmentAPIsConfig(config)
		
		// Should have warnings for missing API keys
		if len(warnings) == 0 {
			t.Error("Expected warnings for missing API keys")
		}
		
		// Test with API keys provided
		configWithKeys := GovernmentAPIsConfig{
			CompaniesHouseAPIKey:    "test-key",
			OpenCorporatesAPIToken: "test-token",
		}
		
		warningsWithKeys := ValidateGovernmentAPIsConfig(configWithKeys)
		if len(warningsWithKeys) > 0 {
			t.Logf("Warnings with API keys: %v", warningsWithKeys)
		}
	})
}

func TestProviderDataValidation(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	factory := NewGovernmentProvidersFactory(logger)
	
	t.Run("SEC EDGAR Data Validation", func(t *testing.T) {
		provider := factory.CreateSECEdgarProvider()
		
		// Test valid data
		validData := &BusinessData{
			CompanyName: "Apple Inc",
			ProviderID:  "0000320193",
			DataSources: []DataSource{
				{
					Name:       "SEC EDGAR",
					Type:       "government",
					TrustLevel: "high",
				},
			},
		}
		
		result, err := provider.ValidateData(validData)
		if err != nil {
			t.Errorf("Validation failed: %v", err)
		}
		
		if !result.IsValid {
			t.Error("Expected valid data to pass validation")
		}
		
		if result.QualityScore < 0.8 {
			t.Errorf("Expected quality score >= 0.8, got %f", result.QualityScore)
		}
		
		// Test invalid data
		invalidData := &BusinessData{
			// Missing required fields
		}
		
		result, err = provider.ValidateData(invalidData)
		if err != nil {
			t.Errorf("Validation failed: %v", err)
		}
		
		if result.IsValid {
			t.Error("Expected invalid data to fail validation")
		}
		
		if len(result.Issues) == 0 {
			t.Error("Expected validation issues for invalid data")
		}
	})
	
	t.Run("WHOIS Data Validation", func(t *testing.T) {
		provider := factory.CreateWHOISProvider()
		
		// Test valid data with website
		validData := &BusinessData{
			CompanyName: "Apple Inc",
			Website:     "https://www.apple.com",
			ProviderID:  "apple.com",
			DataSources: []DataSource{
				{
					Name:       "WHOIS",
					Type:       "domain_registry",
					TrustLevel: "medium",
				},
			},
		}
		
		result, err := provider.ValidateData(validData)
		if err != nil {
			t.Errorf("Validation failed: %v", err)
		}
		
		if !result.IsValid {
			t.Error("Expected valid data to pass validation")
		}
		
		// Test invalid data without website
		invalidData := &BusinessData{
			CompanyName: "Apple Inc",
			// Missing website
		}
		
		result, err = provider.ValidateData(invalidData)
		if err != nil {
			t.Errorf("Validation failed: %v", err)
		}
		
		if result.IsValid {
			t.Error("Expected invalid data to fail validation")
		}
	})
}

// Benchmark tests for performance
func BenchmarkGovernmentProviders(b *testing.B) {
	logger := log.New(os.Stdout, "[BENCH] ", log.LstdFlags)
	factory := NewGovernmentProvidersFactory(logger)
	
	b.Run("SEC EDGAR Provider Creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = factory.CreateSECEdgarProvider()
		}
	})
	
	b.Run("Companies House Provider Creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = factory.CreateCompaniesHouseProvider("test-key")
		}
	})
	
	b.Run("OpenCorporates Provider Creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = factory.CreateOpenCorporatesProvider("test-token")
		}
	})
	
	b.Run("WHOIS Provider Creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = factory.CreateWHOISProvider()
		}
	})
}
