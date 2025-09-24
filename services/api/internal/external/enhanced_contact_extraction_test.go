package external

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewEnhancedContactExtractorV2(t *testing.T) {
	logger := zap.NewNop()

	t.Run("with default config", func(t *testing.T) {
		extractor := NewEnhancedContactExtractorV2(nil, logger)
		assert.NotNil(t, extractor)
		assert.NotNil(t, extractor.config)
		assert.True(t, extractor.config.EnableInternationalPhones)
		assert.True(t, extractor.config.EnableTollFreeNumbers)
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &EnhancedExtractionConfig{
			ContactExtractionConfig:   getDefaultContactExtractionConfig(),
			EnableInternationalPhones: false,
			PhoneValidationStrict:     false,
		}

		extractor := NewEnhancedContactExtractorV2(config, logger)
		assert.NotNil(t, extractor)
		assert.False(t, extractor.config.EnableInternationalPhones)
		assert.False(t, extractor.config.PhoneValidationStrict)
	})
}

func TestExtractPhoneNumbersAdvanced(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewEnhancedContactExtractorV2(nil, logger)
	ctx := context.Background()

	t.Run("extract various phone formats", func(t *testing.T) {
		content := `
			Call us at (555) 123-4567 or 555-987-6543.
			International: +1-555-111-2222
			Toll-free: 800-555-0123
			UK number: +44 20 7946 0958
		`

		result, err := extractor.ExtractPhoneNumbersAdvanced(ctx, content)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.GreaterOrEqual(t, len(result.PhoneNumbers), 4)
		assert.Greater(t, result.ExtractionStats.ValidNumbers, 0)
		assert.Greater(t, result.ExtractionStats.InternationalNums, 0)
		assert.Greater(t, result.ExtractionStats.TollFreeNumbers, 0)
		assert.Greater(t, result.ExtractionStats.AverageConfidence, 0.0)

		// Check that toll-free number is identified correctly
		foundTollFree := false
		for _, phone := range result.PhoneNumbers {
			if phone.Type == "toll_free" {
				foundTollFree = true
				break
			}
		}
		assert.True(t, foundTollFree)
	})

	t.Run("handle duplicate phone numbers", func(t *testing.T) {
		content := `
			Call (555) 123-4567 or 555-123-4567 for support.
			Phone: 555.123.4567
		`

		result, err := extractor.ExtractPhoneNumbersAdvanced(ctx, content)
		require.NoError(t, err)

		// Should deduplicate the same number in different formats
		assert.LessOrEqual(t, len(result.PhoneNumbers), 2)
	})

	t.Run("context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(2 * time.Nanosecond) // Ensure timeout

		content := "Call (555) 123-4567"
		result, err := extractor.ExtractPhoneNumbersAdvanced(ctx, content)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
		assert.Nil(t, result)
	})
}

func TestExtractEmailAddressesAdvanced(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewEnhancedContactExtractorV2(nil, logger)
	ctx := context.Background()

	t.Run("extract various email types", func(t *testing.T) {
		content := `
			Contact us at info@company.com or sales@business.org.
			Support: support@example.net
			CEO: john.doe@startup.co
			Admin: admin@corporation.biz
		`

		result, err := extractor.ExtractEmailAddressesAdvanced(ctx, content)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.GreaterOrEqual(t, len(result.EmailAddresses), 3)
		assert.Greater(t, result.ExtractionStats.ValidEmails, 0)
		assert.Greater(t, result.ExtractionStats.RoleBasedEmails, 0)
		assert.Greater(t, result.ExtractionStats.AverageConfidence, 0.0)

		// Check that role-based emails are identified correctly
		foundContact := false
		foundSales := false
		for _, email := range result.EmailAddresses {
			if email.Type == "contact" || email.Type == "general" {
				foundContact = true
			}
			if email.Type == "sales" {
				foundSales = true
			}
		}
		assert.True(t, foundContact)
		assert.True(t, foundSales)
	})

	t.Run("filter blacklisted domains", func(t *testing.T) {
		config := getDefaultEnhancedExtractionConfig()
		config.EmailDomainBlacklist = []string{"test.com"}
		extractor := NewEnhancedContactExtractorV2(config, logger)

		content := `
			Valid: info@company.com
			Blacklisted: user@test.com
		`

		result, err := extractor.ExtractEmailAddressesAdvanced(ctx, content)
		require.NoError(t, err)

		// Should not include blacklisted domain
		for _, email := range result.EmailAddresses {
			assert.NotContains(t, email.Address, "test.com")
		}
	})

	t.Run("respect confidence threshold", func(t *testing.T) {
		config := getDefaultEnhancedExtractionConfig()
		config.MinConfidenceThreshold = 0.95
		extractor := NewEnhancedContactExtractorV2(config, logger)

		content := "Contact: info@company.com"

		result, err := extractor.ExtractEmailAddressesAdvanced(ctx, content)
		require.NoError(t, err)

		// Only high-confidence emails should be included
		for _, email := range result.EmailAddresses {
			assert.GreaterOrEqual(t, email.ConfidenceScore, 0.95)
		}
	})
}

func TestExtractPhysicalAddressesAdvanced(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewEnhancedContactExtractorV2(nil, logger)
	ctx := context.Background()

	t.Run("extract various address formats", func(t *testing.T) {
		content := `
			Visit us at 123 Main Street, New York, NY 10001.
			Mailing address: 456 Business Ave, Los Angeles, CA 90210-1234
			Office: 789 Corporate Blvd, Chicago, IL 60601
		`

		result, err := extractor.ExtractPhysicalAddressesAdvanced(ctx, content)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.GreaterOrEqual(t, len(result.Addresses), 2)
		assert.Greater(t, result.ExtractionStats.ValidAddresses, 0)
		assert.Greater(t, result.ExtractionStats.CompleteAddresses, 0)
		assert.Greater(t, result.ExtractionStats.AverageConfidence, 0.0)

		// Check address parsing
		for _, addr := range result.Addresses {
			assert.NotEmpty(t, addr.StreetAddress)
			assert.NotEmpty(t, addr.City)
			if addr.State != "" {
				assert.Len(t, addr.State, 2) // US state codes are 2 letters
			}
		}
	})

	t.Run("require postal code when configured", func(t *testing.T) {
		config := getDefaultEnhancedExtractionConfig()
		config.RequirePostalCode = true
		extractor := NewEnhancedContactExtractorV2(config, logger)

		content := `
			Complete: 123 Main St, New York, NY 10001
			Incomplete: 456 Business Ave, Los Angeles, CA
		`

		result, err := extractor.ExtractPhysicalAddressesAdvanced(ctx, content)
		require.NoError(t, err)

		// Should only include addresses with postal codes
		for _, addr := range result.Addresses {
			assert.NotEmpty(t, addr.PostalCode)
		}
	})
}

func TestAdvancedPhoneValidation(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewEnhancedContactExtractorV2(nil, logger)

	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		{"valid US number", "5551234567", true},
		{"valid international", "+14155551234", true},
		{"too short", "555123", false},
		{"too long", "555123456789012345", false},
		{"starts with 0", "0551234567", false},
		{"starts with 1 (10 digits)", "1234567890", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.isValidPhoneNumberStrict(tt.number)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAdvancedEmailValidation(t *testing.T) {
	logger := zap.NewNop()
	config := getDefaultEnhancedExtractionConfig()
	config.EmailDomainBlacklist = []string{"blocked.com"}
	config.EmailDomainWhitelist = []string{"allowed.com"}
	extractor := NewEnhancedContactExtractorV2(config, logger)

	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"valid allowed email", "user@allowed.com", true},
		{"blocked domain", "user@blocked.com", false},
		{"not in whitelist", "user@other.com", false},
		{"invalid format", "invalid-email", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.isValidEmailAddressAdvanced(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPhoneTypeDetection(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewEnhancedContactExtractorV2(nil, logger)

	tests := []struct {
		name     string
		number   string
		content  string
		expected string
	}{
		{"toll free", "800-555-0123", "Call our toll-free line", "toll_free"},
		{"international", "+44 20 7946 0958", "International office", "international"},
		{"mobile context", "555-123-4567", "Mobile: 555-123-4567", "mobile"},
		{"office context", "555-123-4567", "Main office: 555-123-4567", "office"},
		{"fax context", "555-123-4567", "Fax: 555-123-4567", "fax"},
		{"general", "555-123-4567", "Contact number", "general"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.determineAdvancedPhoneType(tt.number, tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEmailTypeDetection(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewEnhancedContactExtractorV2(nil, logger)

	tests := []struct {
		name        string
		email       string
		patternType string
		expected    string
	}{
		{"sales pattern", "sales@company.com", "sales", "sales"},
		{"support pattern", "support@company.com", "support", "support"},
		{"contact pattern", "contact@company.com", "contact", "contact"},
		{"ceo executive", "ceo@company.com", "general", "executive"},
		{"founder executive", "founder@company.com", "general", "executive"},
		{"admin type", "admin@company.com", "general", "admin"},
		{"marketing type", "marketing@company.com", "general", "marketing"},
		{"hr type", "hr@company.com", "general", "hr"},
		{"general fallback", "john@company.com", "general", "general"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.determineAdvancedEmailType(tt.email, tt.patternType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCountryCodeDetection(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewEnhancedContactExtractorV2(nil, logger)

	tests := []struct {
		name     string
		number   string
		hint     string
		expected string
	}{
		{"US hint", "555-123-4567", "US", "US"},
		{"UK hint", "020 7946 0958", "UK", "UK"},
		{"international US", "+1-555-123-4567", "international", "US"},
		{"international UK", "+44 20 7946 0958", "international", "UK"},
		{"international AU", "+61 2 1234 5678", "international", "AU"},
		{"international DE", "+49 30 12345678", "international", "DE"},
		{"international FR", "+33 1 23 45 67 89", "international", "FR"},
		{"unknown international", "+999 123 456 789", "international", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.determineAdvancedCountryCode(tt.number, tt.hint)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAddressParsing(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewEnhancedContactExtractorV2(nil, logger)

	tests := []struct {
		name     string
		address  string
		expected EnhancedPhysicalAddress
	}{
		{
			name:    "full US address",
			address: "123 Main St, New York, NY 10001",
			expected: EnhancedPhysicalAddress{
				StreetAddress: "123 Main St",
				City:          "New York",
				State:         "NY",
				PostalCode:    "10001",
			},
		},
		{
			name:    "address with ZIP+4",
			address: "456 Business Ave, Los Angeles, CA 90210-1234",
			expected: EnhancedPhysicalAddress{
				StreetAddress: "456 Business Ave",
				City:          "Los Angeles",
				State:         "CA",
				PostalCode:    "90210-1234",
			},
		},
		{
			name:    "minimal address",
			address: "789 Corporate Blvd, Chicago",
			expected: EnhancedPhysicalAddress{
				StreetAddress: "789 Corporate Blvd",
				City:          "Chicago",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.parseAdvancedAddress(tt.address)
			assert.Equal(t, tt.expected.StreetAddress, result.StreetAddress)
			assert.Equal(t, tt.expected.City, result.City)
			assert.Equal(t, tt.expected.State, result.State)
			assert.Equal(t, tt.expected.PostalCode, result.PostalCode)
		})
	}
}

func TestUtilityFunctions(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewEnhancedContactExtractorV2(nil, logger)

	t.Run("clean phone number", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"(555) 123-4567", "5551234567"},
			{"+1-555-123-4567", "+15551234567"},
			{"555.123.4567", "5551234567"},
			{"555 123 4567", "5551234567"},
		}

		for _, tt := range tests {
			result := extractor.cleanPhoneNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("toll free detection", func(t *testing.T) {
		tests := []struct {
			number   string
			expected bool
		}{
			{"800-555-0123", true},
			{"888-555-0123", true},
			{"877-555-0123", true},
			{"555-123-4567", false},
			{"+1-800-555-0123", true},
		}

		for _, tt := range tests {
			result := extractor.isTollFreeNumber(tt.number)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("role-based email detection", func(t *testing.T) {
		tests := []struct {
			email    string
			expected bool
		}{
			{"info@company.com", true},
			{"contact@company.com", true},
			{"sales@company.com", true},
			{"support@company.com", true},
			{"john.doe@company.com", false},
			{"admin@company.com", true},
		}

		for _, tt := range tests {
			result := extractor.isRoleBasedEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("business domain detection", func(t *testing.T) {
		tests := []struct {
			domain   string
			expected bool
		}{
			{"company.com", true},
			{"business.org", true},
			{"gmail.com", false},
			{"yahoo.com", false},
			{"hotmail.com", false},
			{"startup.co", true},
		}

		for _, tt := range tests {
			result := extractor.isBusinessDomain(tt.domain)
			assert.Equal(t, tt.expected, result)
		}
	})
}

func TestConfidenceCalculation(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewEnhancedContactExtractorV2(nil, logger)

	t.Run("phone confidence", func(t *testing.T) {
		pattern := PhonePattern{
			Name:       "test_pattern",
			Confidence: 0.8,
		}

		tests := []struct {
			name     string
			number   string
			expected float64
		}{
			{"international number", "+1-555-123-4567", 0.87}, // 0.8 + 0.05 + 0.02
			{"toll free number", "800-555-0123", 0.87},        // 0.8 + 0.05 + 0.02
			{"regular number", "555-123-4567", 0.82},          // 0.8 + 0.02
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := extractor.calculatePhoneConfidence(tt.number, pattern)
				assert.InDelta(t, tt.expected, result, 0.01)
			})
		}
	})

	t.Run("email confidence", func(t *testing.T) {
		pattern := EmailPattern{
			Name:       "test_pattern",
			Confidence: 0.9,
		}

		tests := []struct {
			name     string
			email    string
			expected float64
		}{
			{"role-based email", "info@company.com", 0.98}, // 0.9 + 0.05 + 0.03
			{"business domain", "user@company.com", 0.93},  // 0.9 + 0.03
			{"personal domain", "user@gmail.com", 0.9},     // 0.9 (no bonuses)
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := extractor.calculateEmailConfidence(tt.email, pattern)
				assert.InDelta(t, tt.expected, result, 0.01)
			})
		}
	})

	t.Run("address confidence", func(t *testing.T) {
		pattern := AddressPattern{
			Name:       "test_pattern",
			Confidence: 0.8,
		}

		tests := []struct {
			name     string
			address  string
			expected float64
		}{
			{"full address with postal", "123 Main St, New York, NY 10001", 0.87}, // 0.8 + 0.02 + 0.05
			{"address with commas", "123 Main St, New York, NY", 0.82},            // 0.8 + 0.02
			{"simple address", "123 Main St", 0.8},                                // 0.8
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := extractor.calculateAddressConfidence(tt.address, pattern)
				assert.InDelta(t, tt.expected, result, 0.01)
			})
		}
	})
}

func TestStatisticsCalculation(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewEnhancedContactExtractorV2(nil, logger)

	t.Run("phone statistics", func(t *testing.T) {
		phones := []EnhancedPhoneNumber{
			{Number: "+1-555-123-4567", ConfidenceScore: 0.9, Type: "international"},
			{Number: "800-555-0123", ConfidenceScore: 0.95, Type: "toll_free"},
			{Number: "555-987-6543", ConfidenceScore: 0.8, Type: "general"},
		}

		stats := extractor.calculatePhoneStats(phones)
		assert.Equal(t, 3, stats.TotalMatches)
		assert.Greater(t, stats.ValidNumbers, 0)
		assert.Equal(t, 1, stats.InternationalNums)
		assert.Equal(t, 1, stats.TollFreeNumbers)
		assert.InDelta(t, 0.88, stats.AverageConfidence, 0.02) // (0.9 + 0.95 + 0.8) / 3
	})

	t.Run("email statistics", func(t *testing.T) {
		emails := []EnhancedEmailAddress{
			{Address: "info@company.com", ConfidenceScore: 0.95, Type: "contact"},
			{Address: "sales@company.com", ConfidenceScore: 0.9, Type: "sales"},
			{Address: "john@company.com", ConfidenceScore: 0.85, Type: "general"},
		}

		stats := extractor.calculateEmailStats(emails)
		assert.Equal(t, 3, stats.TotalMatches)
		assert.Greater(t, stats.ValidEmails, 0)
		assert.Equal(t, 2, stats.RoleBasedEmails)             // info and sales
		assert.Equal(t, 1, stats.PersonalEmails)              // john
		assert.InDelta(t, 0.9, stats.AverageConfidence, 0.02) // (0.95 + 0.9 + 0.85) / 3
	})

	t.Run("address statistics", func(t *testing.T) {
		addresses := []EnhancedPhysicalAddress{
			{
				StreetAddress:   "123 Main St",
				City:            "New York",
				State:           "NY",
				PostalCode:      "10001",
				ConfidenceScore: 0.9,
			},
			{
				StreetAddress:   "456 Business Ave",
				City:            "Los Angeles",
				ConfidenceScore: 0.8,
			},
		}

		stats := extractor.calculateAddressStats(addresses)
		assert.Equal(t, 2, stats.TotalMatches)
		assert.Equal(t, 2, stats.ValidAddresses)               // Both have street and city
		assert.Equal(t, 1, stats.CompleteAddresses)            // Only first has all fields
		assert.InDelta(t, 0.85, stats.AverageConfidence, 0.01) // (0.9 + 0.8) / 2
	})
}

func TestDeduplication(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewEnhancedContactExtractorV2(nil, logger)

	t.Run("phone normalization", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"(555) 123-4567", "5551234567"},
			{"+1-555-123-4567", "5551234567"},
			{"555.123.4567", "5551234567"},
			{"+44 20 7946 0958", "+442079460958"},
		}

		for _, tt := range tests {
			result := extractor.normalizePhoneForDeduplication(tt.input)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("address normalization", func(t *testing.T) {
		address := EnhancedPhysicalAddress{
			StreetAddress: "123 Main Street",
			City:          "New York",
			State:         "NY",
			PostalCode:    "10001",
		}

		result := extractor.normalizeAddressForDeduplication(address)
		expected := "123 main street|new york|ny|10001"
		assert.Equal(t, expected, result)
	})
}

func TestConfigurationMethods(t *testing.T) {
	t.Run("default enhanced config", func(t *testing.T) {
		config := getDefaultEnhancedExtractionConfig()
		assert.NotNil(t, config)
		assert.True(t, config.EnableInternationalPhones)
		assert.True(t, config.EnableTollFreeNumbers)
		assert.True(t, config.PhoneValidationStrict)
		assert.True(t, config.EnableRoleBasedEmails)
		assert.True(t, config.EnablePersonalEmails)
		assert.Contains(t, config.EmailDomainBlacklist, "example.com")
		assert.Equal(t, 0.7, config.MinConfidenceThreshold)
	})
}
