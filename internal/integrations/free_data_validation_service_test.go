package integrations

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestFreeDataValidationService_ValidateBusinessData(t *testing.T) {
	// Setup
	config := GetDefaultFreeDataValidationConfig()
	logger := zap.NewNop()
	governmentAPIs := &BusinessDataAPIService{} // Mock service
	service := NewFreeDataValidationService(config, logger, governmentAPIs)

	tests := []struct {
		name             string
		data             BusinessDataForValidation
		expectedValid    bool
		expectedMinScore float64
		expectedErrors   int
		expectedWarnings int
	}{
		{
			name: "complete_valid_data",
			data: BusinessDataForValidation{
				BusinessID:         "test-001",
				Name:               "Acme Corporation",
				Description:        "A leading technology company specializing in software development",
				Address:            "123 Main Street, New York, NY 10001",
				Phone:              "+1-555-123-4567",
				Email:              "contact@acme.com",
				Website:            "https://www.acme.com",
				Industry:           "Technology",
				Country:            "US",
				RegistrationNumber: "1234567890",
				TaxID:              "12-3456789",
			},
			expectedValid:    true,
			expectedMinScore: 0.7,
			expectedErrors:   0,
			expectedWarnings: 2, // Address-country consistency and name-description consistency warnings
		},
		{
			name: "missing_required_fields",
			data: BusinessDataForValidation{
				BusinessID:  "test-002",
				Name:        "", // Missing required field
				Description: "",
				Address:     "",
				Country:     "",
			},
			expectedValid:    false,
			expectedMinScore: 0.0,
			expectedErrors:   4, // 4 missing required fields
			expectedWarnings: 4, // 4 missing optional fields
		},
		{
			name: "inconsistent_email_domain",
			data: BusinessDataForValidation{
				BusinessID:  "test-003",
				Name:        "Test Company",
				Description: "A test company",
				Address:     "123 Test St, Test City, TC 12345",
				Country:     "US",
				Email:       "contact@different-domain.com",
				Website:     "https://www.test-company.com",
			},
			expectedValid:    true,
			expectedMinScore: 0.5,
			expectedErrors:   0,
			expectedWarnings: 4, // Email domain mismatch + 3 other warnings
		},
		{
			name: "invalid_phone_format",
			data: BusinessDataForValidation{
				BusinessID:  "test-004",
				Name:        "Test Company",
				Description: "A test company",
				Address:     "123 Test St, Test City, TC 12345",
				Country:     "US",
				Phone:       "invalid-phone",
			},
			expectedValid:    true,
			expectedMinScore: 0.5,
			expectedErrors:   0,
			expectedWarnings: 5, // Invalid phone format + 4 other warnings
		},
		{
			name: "uk_company_data",
			data: BusinessDataForValidation{
				BusinessID:         "test-005",
				Name:               "British Tech Ltd",
				Description:        "A UK technology company",
				Address:            "456 London Road, London, UK",
				Phone:              "+44-20-1234-5678",
				Email:              "info@britishtech.co.uk",
				Website:            "https://www.britishtech.co.uk",
				Industry:           "Technology",
				Country:            "UK",
				RegistrationNumber: "12345678",
			},
			expectedValid:    true,
			expectedMinScore: 0.7,
			expectedErrors:   0,
			expectedWarnings: 1, // Name-description consistency warning
		},
		{
			name: "minimal_valid_data",
			data: BusinessDataForValidation{
				BusinessID:  "test-006",
				Name:        "Minimal Company",
				Description: "A minimal company",
				Address:     "789 Simple St, Simple City, SC 54321",
				Country:     "US",
			},
			expectedValid:    true,
			expectedMinScore: 0.5,
			expectedErrors:   0,
			expectedWarnings: 5, // Missing optional fields + consistency warnings
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := service.ValidateBusinessData(ctx, tt.data)

			if err != nil {
				t.Errorf("ValidateBusinessData() error = %v, wantErr false", err)
				return
			}

			if result.IsValid != tt.expectedValid {
				t.Errorf("ValidateBusinessData() IsValid = %v, want %v", result.IsValid, tt.expectedValid)
			}

			if result.QualityScore < tt.expectedMinScore {
				t.Errorf("ValidateBusinessData() QualityScore = %v, want >= %v", result.QualityScore, tt.expectedMinScore)
			}

			if len(result.ValidationErrors) != tt.expectedErrors {
				t.Errorf("ValidateBusinessData() ValidationErrors count = %v, want %v", len(result.ValidationErrors), tt.expectedErrors)
			}

			if len(result.ValidationWarnings) != tt.expectedWarnings {
				t.Errorf("ValidateBusinessData() ValidationWarnings count = %v, want %v", len(result.ValidationWarnings), tt.expectedWarnings)
			}

			// Verify cost is always 0.0 (free validation)
			if result.Cost != 0.0 {
				t.Errorf("ValidateBusinessData() Cost = %v, want 0.0", result.Cost)
			}

			// Verify validation time is reasonable
			if result.ValidationTime > 5*time.Second {
				t.Errorf("ValidateBusinessData() ValidationTime = %v, want < 5s", result.ValidationTime)
			}
		})
	}
}

func TestFreeDataValidationService_ValidateCompleteness(t *testing.T) {
	service := &FreeDataValidationService{
		config: GetDefaultFreeDataValidationConfig(),
		logger: zap.NewNop(),
	}

	tests := []struct {
		name             string
		data             BusinessDataForValidation
		expectedScore    float64
		expectedErrors   int
		expectedWarnings int
	}{
		{
			name: "complete_data",
			data: BusinessDataForValidation{
				Name:               "Complete Company",
				Description:        "A complete company",
				Address:            "123 Complete St",
				Country:            "US",
				Phone:              "+1-555-123-4567",
				Email:              "contact@complete.com",
				Website:            "https://www.complete.com",
				RegistrationNumber: "1234567890",
			},
			expectedScore:    2.5, // 4 required fields * 0.5 + 4 optional fields * 0.125
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "missing_required_fields",
			data: BusinessDataForValidation{
				Name:        "", // Missing
				Description: "", // Missing
				Address:     "", // Missing
				Country:     "", // Missing
			},
			expectedScore:    0.0,
			expectedErrors:   4,
			expectedWarnings: 4,
		},
		{
			name: "partial_data",
			data: BusinessDataForValidation{
				Name:        "Partial Company",
				Description: "A partial company",
				Address:     "123 Partial St",
				Country:     "US",
				// Missing optional fields
			},
			expectedScore:    2.0, // 4 required fields * 0.5
			expectedErrors:   0,
			expectedWarnings: 4, // Missing optional fields
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				ValidationErrors:   []ValidationError{},
				ValidationWarnings: []ValidationWarning{},
			}

			err := service.validateCompleteness(tt.data, result)
			if err != nil {
				t.Errorf("validateCompleteness() error = %v", err)
			}

			if result.CompletenessScore != tt.expectedScore {
				t.Errorf("validateCompleteness() CompletenessScore = %v, want %v", result.CompletenessScore, tt.expectedScore)
			}

			if len(result.ValidationErrors) != tt.expectedErrors {
				t.Errorf("validateCompleteness() ValidationErrors count = %v, want %v", len(result.ValidationErrors), tt.expectedErrors)
			}

			if len(result.ValidationWarnings) != tt.expectedWarnings {
				t.Errorf("validateCompleteness() ValidationWarnings count = %v, want %v", len(result.ValidationWarnings), tt.expectedWarnings)
			}
		})
	}
}

func TestFreeDataValidationService_ValidateConsistency(t *testing.T) {
	service := &FreeDataValidationService{
		config: GetDefaultFreeDataValidationConfig(),
		logger: zap.NewNop(),
	}

	tests := []struct {
		name             string
		data             BusinessDataForValidation
		expectedScore    float64
		expectedWarnings int
	}{
		{
			name: "consistent_data",
			data: BusinessDataForValidation{
				Email:       "contact@acme.com",
				Website:     "https://www.acme.com",
				Phone:       "+1-555-123-4567",
				Name:        "Acme Corporation",
				Description: "Acme Corporation is a leading technology company",
				Address:     "123 Main St, New York, NY 10001",
				Country:     "US",
			},
			expectedScore:    0.8, // 0.3 (email/website) + 0.2 (phone) + 0.3 (name/description) + 0.0 (address/country - no match)
			expectedWarnings: 1,   // Address-country mismatch warning
		},
		{
			name: "inconsistent_email_domain",
			data: BusinessDataForValidation{
				Email:   "contact@different.com",
				Website: "https://www.acme.com",
			},
			expectedScore:    0.0, // No consistency checks pass
			expectedWarnings: 1,   // Email domain mismatch
		},
		{
			name: "invalid_phone_format",
			data: BusinessDataForValidation{
				Phone: "invalid-phone",
			},
			expectedScore:    0.0, // No consistency checks pass
			expectedWarnings: 1,   // Invalid phone format
		},
		{
			name: "inconsistent_name_description",
			data: BusinessDataForValidation{
				Name:        "Acme Corporation",
				Description: "A completely different company description",
			},
			expectedScore:    0.0, // No consistency checks pass
			expectedWarnings: 1,   // Name description mismatch
		},
		{
			name: "inconsistent_address_country",
			data: BusinessDataForValidation{
				Address: "123 Main St, London, UK",
				Country: "US",
			},
			expectedScore:    0.0, // No consistency checks pass
			expectedWarnings: 1,   // Address country mismatch
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				ValidationWarnings: []ValidationWarning{},
			}

			err := service.validateConsistency(tt.data, result)
			if err != nil {
				t.Errorf("validateConsistency() error = %v", err)
			}

			if result.ConsistencyScore != tt.expectedScore {
				t.Errorf("validateConsistency() ConsistencyScore = %v, want %v", result.ConsistencyScore, tt.expectedScore)
			}

			if len(result.ValidationWarnings) != tt.expectedWarnings {
				t.Errorf("validateConsistency() ValidationWarnings count = %v, want %v", len(result.ValidationWarnings), tt.expectedWarnings)
			}
		})
	}
}

func TestFreeDataValidationService_HelperMethods(t *testing.T) {
	service := &FreeDataValidationService{
		config: GetDefaultFreeDataValidationConfig(),
		logger: zap.NewNop(),
	}

	t.Run("isValidAddressFormat", func(t *testing.T) {
		tests := []struct {
			address  string
			expected bool
		}{
			{"123 Main Street, New York, NY 10001", true},
			{"456 Oak Avenue", true},
			{"Short", false},
			{"", false},
		}

		for _, tt := range tests {
			result := service.isValidAddressFormat(tt.address)
			if result != tt.expected {
				t.Errorf("isValidAddressFormat(%q) = %v, want %v", tt.address, result, tt.expected)
			}
		}
	})

	t.Run("isEmailDomainConsistent", func(t *testing.T) {
		tests := []struct {
			email    string
			website  string
			expected bool
		}{
			{"contact@acme.com", "https://www.acme.com", true},
			{"contact@acme.com", "https://acme.com", true},
			{"contact@acme.com", "http://www.acme.com", true},
			{"contact@different.com", "https://www.acme.com", false},
			{"invalid-email", "https://www.acme.com", false},
		}

		for _, tt := range tests {
			result := service.isEmailDomainConsistent(tt.email, tt.website)
			if result != tt.expected {
				t.Errorf("isEmailDomainConsistent(%q, %q) = %v, want %v", tt.email, tt.website, result, tt.expected)
			}
		}
	})

	t.Run("isValidPhoneFormat", func(t *testing.T) {
		tests := []struct {
			phone    string
			expected bool
		}{
			{"+1-555-123-4567", true},
			{"(555) 123-4567", true},
			{"555-123-4567", true},
			{"5551234567", true},
			{"+44-20-1234-5678", true},
			{"invalid-phone", false},
			{"123", false},
			{"", false},
		}

		for _, tt := range tests {
			result := service.isValidPhoneFormat(tt.phone)
			if result != tt.expected {
				t.Errorf("isValidPhoneFormat(%q) = %v, want %v", tt.phone, result, tt.expected)
			}
		}
	})

	t.Run("isNameDescriptionConsistent", func(t *testing.T) {
		tests := []struct {
			name        string
			description string
			expected    bool
		}{
			{"Acme Corporation", "Acme Corporation is a leading technology company", true},
			{"Acme Corp", "Acme Corporation provides software solutions", true},
			{"Acme Corporation", "A completely different company description", false},
			{"Short", "Short description", true},
			{"", "Some description", false},
		}

		for _, tt := range tests {
			result := service.isNameDescriptionConsistent(tt.name, tt.description)
			if result != tt.expected {
				t.Errorf("isNameDescriptionConsistent(%q, %q) = %v, want %v", tt.name, tt.description, result, tt.expected)
			}
		}
	})

	t.Run("isAddressCountryConsistent", func(t *testing.T) {
		tests := []struct {
			address  string
			country  string
			expected bool
		}{
			{"123 Main St, New York, NY, USA", "US", true},
			{"456 London Road, London, UK", "UK", true},
			{"789 Paris Street, Paris, France", "FR", true},
			{"123 Main St, New York, NY, USA", "UK", false},
			{"456 London Road, London, UK", "US", false},
		}

		for _, tt := range tests {
			result := service.isAddressCountryConsistent(tt.address, tt.country)
			if result != tt.expected {
				t.Errorf("isAddressCountryConsistent(%q, %q) = %v, want %v", tt.address, tt.country, result, tt.expected)
			}
		}
	})
}

func TestFreeDataValidationService_CacheManagement(t *testing.T) {
	service := &FreeDataValidationService{
		config:          GetDefaultFreeDataValidationConfig(),
		logger:          zap.NewNop(),
		validationCache: make(map[string]*ValidationResult),
	}

	// Test caching
	businessID := "test-cache-001"
	result := &ValidationResult{
		BusinessID:   businessID,
		IsValid:      true,
		QualityScore: 0.85,
		ValidatedAt:  time.Now(),
	}

	// Cache result
	service.cacheValidationResult(businessID, result)

	// Retrieve cached result
	cached := service.getCachedValidation(businessID)
	if cached == nil {
		t.Error("Expected cached result, got nil")
	}

	if cached.QualityScore != result.QualityScore {
		t.Errorf("Cached result QualityScore = %v, want %v", cached.QualityScore, result.QualityScore)
	}

	// Test cache expiration
	expiredResult := &ValidationResult{
		BusinessID:   "test-expired",
		IsValid:      true,
		QualityScore: 0.75,
		ValidatedAt:  time.Now().Add(-2 * time.Hour), // 2 hours ago
	}

	service.cacheValidationResult("test-expired", expiredResult)
	cached = service.getCachedValidation("test-expired")
	if cached != nil {
		t.Error("Expected nil for expired cache entry, got result")
	}
}

func TestFreeDataValidationService_GetValidationStats(t *testing.T) {
	service := &FreeDataValidationService{
		config:          GetDefaultFreeDataValidationConfig(),
		logger:          zap.NewNop(),
		validationCache: make(map[string]*ValidationResult),
	}

	// Add some test data
	service.cacheValidationResult("test-001", &ValidationResult{
		BusinessID:   "test-001",
		IsValid:      true,
		QualityScore: 0.9,
		ValidatedAt:  time.Now(),
	})

	service.cacheValidationResult("test-002", &ValidationResult{
		BusinessID:   "test-002",
		IsValid:      false,
		QualityScore: 0.6,
		ValidatedAt:  time.Now(),
	})

	service.cacheValidationResult("test-003", &ValidationResult{
		BusinessID:   "test-003",
		IsValid:      true,
		QualityScore: 0.8,
		ValidatedAt:  time.Now(),
	})

	stats := service.GetValidationStats()

	if stats["total_validations"] != 3 {
		t.Errorf("Expected total_validations = 3, got %v", stats["total_validations"])
	}

	if stats["valid_count"] != 2 {
		t.Errorf("Expected valid_count = 2, got %v", stats["valid_count"])
	}

	if stats["invalid_count"] != 1 {
		t.Errorf("Expected invalid_count = 1, got %v", stats["invalid_count"])
	}

	expectedAvgScore := (0.9 + 0.6 + 0.8) / 3.0
	actualScore := stats["average_quality_score"].(float64)
	if actualScore != expectedAvgScore {
		t.Errorf("Expected average_quality_score = %v, got %v", expectedAvgScore, actualScore)
	}

	if stats["cost_per_validation"] != 0.0 {
		t.Errorf("Expected cost_per_validation = 0.0, got %v", stats["cost_per_validation"])
	}
}

func TestFreeDataValidationService_Configuration(t *testing.T) {
	config := GetDefaultFreeDataValidationConfig()

	// Test default configuration values
	if config.MinQualityScore != 0.7 {
		t.Errorf("Expected MinQualityScore = 0.7, got %v", config.MinQualityScore)
	}

	if config.ConsistencyWeight != 0.3 {
		t.Errorf("Expected ConsistencyWeight = 0.3, got %v", config.ConsistencyWeight)
	}

	if config.CompletenessWeight != 0.25 {
		t.Errorf("Expected CompletenessWeight = 0.25, got %v", config.CompletenessWeight)
	}

	if config.AccuracyWeight != 0.25 {
		t.Errorf("Expected AccuracyWeight = 0.25, got %v", config.AccuracyWeight)
	}

	if config.FreshnessWeight != 0.2 {
		t.Errorf("Expected FreshnessWeight = 0.2, got %v", config.FreshnessWeight)
	}

	if !config.EnableCrossReference {
		t.Error("Expected EnableCrossReference = true")
	}

	if config.MaxValidationTime != 30 {
		t.Errorf("Expected MaxValidationTime = 30, got %v", config.MaxValidationTime)
	}

	if !config.CacheValidationResults {
		t.Error("Expected CacheValidationResults = true")
	}

	if config.MaxAPICallsPerValidation != 10 {
		t.Errorf("Expected MaxAPICallsPerValidation = 10, got %v", config.MaxAPICallsPerValidation)
	}

	if config.RateLimitDelay != 100 {
		t.Errorf("Expected RateLimitDelay = 100, got %v", config.RateLimitDelay)
	}

	// Test that weights sum to 1.0
	totalWeight := config.ConsistencyWeight + config.CompletenessWeight + config.AccuracyWeight + config.FreshnessWeight
	if totalWeight != 1.0 {
		t.Errorf("Expected weights to sum to 1.0, got %v", totalWeight)
	}
}

// Benchmark tests
func BenchmarkFreeDataValidationService_ValidateBusinessData(b *testing.B) {
	config := GetDefaultFreeDataValidationConfig()
	logger := zap.NewNop()
	governmentAPIs := &BusinessDataAPIService{}
	service := NewFreeDataValidationService(config, logger, governmentAPIs)

	data := BusinessDataForValidation{
		BusinessID:         "benchmark-test",
		Name:               "Benchmark Company",
		Description:        "A company for benchmarking",
		Address:            "123 Benchmark St, Benchmark City, BC 12345",
		Phone:              "+1-555-123-4567",
		Email:              "contact@benchmark.com",
		Website:            "https://www.benchmark.com",
		Industry:           "Technology",
		Country:            "US",
		RegistrationNumber: "1234567890",
		TaxID:              "12-3456789",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ValidateBusinessData(ctx, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
