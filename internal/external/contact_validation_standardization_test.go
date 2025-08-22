package external

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewContactValidationStandardizer(t *testing.T) {
	logger := zap.NewNop()

	t.Run("with default config", func(t *testing.T) {
		standardizer := NewContactValidationStandardizer(nil, logger)
		assert.NotNil(t, standardizer)
		assert.NotNil(t, standardizer.config)
		assert.True(t, standardizer.config.EnablePhoneValidation)
		assert.True(t, standardizer.config.EnableEmailValidation)
		assert.Equal(t, 0.7, standardizer.config.MinValidationConfidence)
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &ContactValidationConfig{
			EnablePhoneValidation:   false,
			EnableEmailValidation:   true,
			MinValidationConfidence: 0.9,
		}

		standardizer := NewContactValidationStandardizer(config, logger)
		assert.NotNil(t, standardizer)
		assert.False(t, standardizer.config.EnablePhoneValidation)
		assert.True(t, standardizer.config.EnableEmailValidation)
		assert.Equal(t, 0.9, standardizer.config.MinValidationConfidence)
	})

	t.Run("with nil logger", func(t *testing.T) {
		standardizer := NewContactValidationStandardizer(nil, nil)
		assert.NotNil(t, standardizer)
		assert.NotNil(t, standardizer.logger)
	})
}

func TestValidatePhoneNumber(t *testing.T) {
	logger := zap.NewNop()
	standardizer := NewContactValidationStandardizer(nil, logger)
	ctx := context.Background()

	t.Run("valid E.164 format", func(t *testing.T) {
		result, err := standardizer.ValidatePhoneNumber(ctx, "+1234567890")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.True(t, result.IsValid)
		assert.Equal(t, "+1234567890", result.OriginalValue)
		assert.Equal(t, "+1234567890", result.StandardizedValue)
		assert.Greater(t, result.ValidationScore, 0.9)
		assert.Equal(t, "E.164", result.TechnicalInfo.Format)
		assert.Empty(t, result.ValidationErrors)
	})

	t.Run("valid US format with parentheses", func(t *testing.T) {
		result, err := standardizer.ValidatePhoneNumber(ctx, "(123) 456-7890")
		require.NoError(t, err)

		assert.True(t, result.IsValid)
		assert.Greater(t, result.ValidationScore, 0.7)
		assert.Contains(t, result.StandardizedValue, "123")
	})

	t.Run("invalid phone number", func(t *testing.T) {
		result, err := standardizer.ValidatePhoneNumber(ctx, "invalid-phone")
		require.NoError(t, err)

		assert.False(t, result.IsValid)
		assert.Greater(t, len(result.ValidationErrors), 0)
		assert.Equal(t, "INVALID_FORMAT", result.ValidationErrors[0].Code)
	})

	t.Run("empty phone number", func(t *testing.T) {
		result, err := standardizer.ValidatePhoneNumber(ctx, "")
		require.NoError(t, err)

		assert.False(t, result.IsValid)
		assert.Greater(t, len(result.ValidationErrors), 0)
		assert.Equal(t, "EMPTY_PHONE", result.ValidationErrors[0].Code)
	})

	t.Run("phone validation disabled", func(t *testing.T) {
		config := getDefaultContactValidationConfig()
		config.EnablePhoneValidation = false
		standardizer := NewContactValidationStandardizer(config, logger)

		result, err := standardizer.ValidatePhoneNumber(ctx, "invalid-phone")
		require.NoError(t, err)

		assert.True(t, result.IsValid)
		assert.Equal(t, 1.0, result.ValidationScore)
	})

	t.Run("country code validation", func(t *testing.T) {
		config := getDefaultContactValidationConfig()
		config.AllowedCountryCodes = []string{"1", "44"}
		standardizer := NewContactValidationStandardizer(config, logger)

		// Valid country code
		result, err := standardizer.ValidatePhoneNumber(ctx, "+1234567890")
		require.NoError(t, err)
		assert.Equal(t, "1", result.GeographicInfo.CountryCode)
		assert.Equal(t, "US", result.GeographicInfo.Country)

		// Invalid country code (should have warning)
		result, err = standardizer.ValidatePhoneNumber(ctx, "+49123456789")
		require.NoError(t, err)
		assert.Greater(t, len(result.ValidationWarnings), 0)
		assert.Equal(t, "COUNTRY_NOT_ALLOWED", result.ValidationWarnings[0].Code)
	})

	t.Run("context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(2 * time.Nanosecond) // Ensure timeout

		result, err := standardizer.ValidatePhoneNumber(ctx, "+1234567890")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation cancelled")
		assert.Nil(t, result)
	})
}

func TestValidateEmailAddress(t *testing.T) {
	logger := zap.NewNop()
	standardizer := NewContactValidationStandardizer(nil, logger)
	ctx := context.Background()

	t.Run("valid email address", func(t *testing.T) {
		result, err := standardizer.ValidateEmailAddress(ctx, "test@example.com")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.True(t, result.IsValid)
		assert.Equal(t, "test@example.com", result.OriginalValue)
		assert.Equal(t, "test@example.com", result.StandardizedValue)
		assert.Greater(t, result.ValidationScore, 0.7)
		assert.Equal(t, "RFC5322", result.TechnicalInfo.Format)
		assert.Equal(t, "SMTP", result.TechnicalInfo.Protocol)
		assert.Empty(t, result.ValidationErrors)
	})

	t.Run("email with common domain", func(t *testing.T) {
		result, err := standardizer.ValidateEmailAddress(ctx, "user@gmail.com")
		require.NoError(t, err)

		assert.True(t, result.IsValid)
		assert.Greater(t, result.ValidationScore, 0.8) // Bonus for .com TLD
		assert.Equal(t, "Google", result.TechnicalInfo.Provider)
	})

	t.Run("invalid email format", func(t *testing.T) {
		result, err := standardizer.ValidateEmailAddress(ctx, "invalid-email")
		require.NoError(t, err)

		assert.False(t, result.IsValid)
		assert.Greater(t, len(result.ValidationErrors), 0)
		assert.Equal(t, "INVALID_FORMAT", result.ValidationErrors[0].Code)
	})

	t.Run("missing domain", func(t *testing.T) {
		result, err := standardizer.ValidateEmailAddress(ctx, "user@")
		require.NoError(t, err)

		assert.False(t, result.IsValid)
		assert.Greater(t, len(result.ValidationErrors), 0)
		assert.Equal(t, "MISSING_DOMAIN", result.ValidationErrors[0].Code)
	})

	t.Run("blocked domain", func(t *testing.T) {
		result, err := standardizer.ValidateEmailAddress(ctx, "user@tempmail.com")
		require.NoError(t, err)

		assert.False(t, result.IsValid)
		assert.Greater(t, len(result.ValidationErrors), 0)
		assert.Equal(t, "BLOCKED_DOMAIN", result.ValidationErrors[0].Code)
	})

	t.Run("email standardization", func(t *testing.T) {
		config := getDefaultContactValidationConfig()
		config.EnableEmailStandardization = true
		standardizer := NewContactValidationStandardizer(config, logger)

		// Gmail alias removal
		result, err := standardizer.ValidateEmailAddress(ctx, "User.Name+alias@gmail.com")
		require.NoError(t, err)

		assert.True(t, result.IsValid)
		assert.Equal(t, "username@gmail.com", result.StandardizedValue)
	})

	t.Run("email validation disabled", func(t *testing.T) {
		config := getDefaultContactValidationConfig()
		config.EnableEmailValidation = false
		standardizer := NewContactValidationStandardizer(config, logger)

		result, err := standardizer.ValidateEmailAddress(ctx, "invalid-email")
		require.NoError(t, err)

		assert.True(t, result.IsValid)
		assert.Equal(t, 1.0, result.ValidationScore)
	})

	t.Run("suspicious email patterns", func(t *testing.T) {
		result, err := standardizer.ValidateEmailAddress(ctx, "user..name@example.com")
		require.NoError(t, err)

		// Should be valid but with lower score due to double dots
		assert.True(t, result.IsValid)
		assert.Less(t, result.QualityMetrics.FormatCompliance, 0.8)
	})
}

func TestValidatePhysicalAddress(t *testing.T) {
	logger := zap.NewNop()
	standardizer := NewContactValidationStandardizer(nil, logger)
	ctx := context.Background()

	t.Run("valid complete address", func(t *testing.T) {
		address := "123 Main St, Anytown, CA 12345"
		result, err := standardizer.ValidatePhysicalAddress(ctx, address)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.True(t, result.IsValid)
		assert.Equal(t, address, result.OriginalValue)
		assert.Greater(t, result.ValidationScore, 0.7)
		assert.Equal(t, "COMMA_SEPARATED", result.TechnicalInfo.Format)
		assert.Equal(t, "123 Main St", result.GeographicInfo.City)
		assert.Equal(t, "CA", result.GeographicInfo.Region)
		assert.Equal(t, "12345", result.GeographicInfo.PostalCode)
		assert.Empty(t, result.ValidationErrors)
	})

	t.Run("address with country", func(t *testing.T) {
		address := "123 Main St, London, GB, SW1A 1AA, United Kingdom"
		result, err := standardizer.ValidatePhysicalAddress(ctx, address)
		require.NoError(t, err)

		assert.True(t, result.IsValid)
		assert.Equal(t, "United Kingdom", result.GeographicInfo.Country)
	})

	t.Run("incomplete address", func(t *testing.T) {
		address := "123 Main St"
		result, err := standardizer.ValidatePhysicalAddress(ctx, address)
		require.NoError(t, err)

		// Should be valid but with lower score
		assert.True(t, result.IsValid)
		assert.Less(t, result.ValidationScore, 0.7)
		assert.Less(t, result.QualityMetrics.DataCompleteness, 1.0)
	})

	t.Run("empty address", func(t *testing.T) {
		result, err := standardizer.ValidatePhysicalAddress(ctx, "")
		require.NoError(t, err)

		assert.False(t, result.IsValid)
		assert.Greater(t, len(result.ValidationErrors), 0)
		assert.Equal(t, "EMPTY_ADDRESS", result.ValidationErrors[0].Code)
	})

	t.Run("address standardization", func(t *testing.T) {
		config := getDefaultContactValidationConfig()
		config.EnableAddressStandardization = true
		standardizer := NewContactValidationStandardizer(config, logger)

		address := "  123   Main   St  ,  Anytown  ,  CA   12345  "
		result, err := standardizer.ValidatePhysicalAddress(ctx, address)
		require.NoError(t, err)

		assert.True(t, result.IsValid)
		assert.Contains(t, result.StandardizedValue, "123 Main St")
		assert.NotContains(t, result.StandardizedValue, "   ") // No extra spaces
	})

	t.Run("postal code validation", func(t *testing.T) {
		config := getDefaultContactValidationConfig()
		config.EnablePostalCodeValidation = true
		standardizer := NewContactValidationStandardizer(config, logger)

		// Valid US postal code
		result, err := standardizer.ValidatePhysicalAddress(ctx, "123 Main St, Anytown, CA 12345")
		require.NoError(t, err)
		assert.True(t, result.IsValid)
		assert.Empty(t, result.ValidationWarnings) // No warnings for valid postal code

		// Invalid postal code
		result, err = standardizer.ValidatePhysicalAddress(ctx, "123 Main St, Anytown, CA ABC")
		require.NoError(t, err)
		assert.Greater(t, len(result.ValidationWarnings), 0)
		assert.Equal(t, "INVALID_POSTAL_CODE", result.ValidationWarnings[0].Code)
	})

	t.Run("address validation disabled", func(t *testing.T) {
		config := getDefaultContactValidationConfig()
		config.EnableAddressValidation = false
		standardizer := NewContactValidationStandardizer(config, logger)

		result, err := standardizer.ValidatePhysicalAddress(ctx, "invalid address")
		require.NoError(t, err)

		assert.True(t, result.IsValid)
		assert.Equal(t, 1.0, result.ValidationScore)
	})
}

func TestValidateBatch(t *testing.T) {
	logger := zap.NewNop()
	standardizer := NewContactValidationStandardizer(nil, logger)
	ctx := context.Background()

	t.Run("batch phone validation", func(t *testing.T) {
		phones := []string{
			"+1234567890",
			"(123) 456-7890",
			"invalid-phone",
			"+44123456789",
		}

		result, err := standardizer.ValidateBatch(ctx, phones, "phone")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, 4, result.TotalProcessed)
		assert.Equal(t, 4, len(result.Results))
		assert.Equal(t, 4, result.ProcessingStats.PhoneValidated)
		assert.Greater(t, result.TotalValid, 0)
		assert.Greater(t, result.TotalInvalid, 0)
		assert.Greater(t, result.ValidationTime, 0)
	})

	t.Run("batch email validation", func(t *testing.T) {
		emails := []string{
			"valid@example.com",
			"user@gmail.com",
			"invalid-email",
			"user@tempmail.com", // Blocked domain
		}

		result, err := standardizer.ValidateBatch(ctx, emails, "email")
		require.NoError(t, err)

		assert.Equal(t, 4, result.TotalProcessed)
		assert.Equal(t, 4, result.ProcessingStats.EmailValidated)
		assert.Greater(t, result.TotalValid, 0)
		assert.Greater(t, result.TotalInvalid, 0)
	})

	t.Run("batch address validation", func(t *testing.T) {
		addresses := []string{
			"123 Main St, Anytown, CA 12345",
			"456 Oak Ave, Somewhere, NY 67890",
			"", // Empty address
		}

		result, err := standardizer.ValidateBatch(ctx, addresses, "address")
		require.NoError(t, err)

		assert.Equal(t, 3, result.TotalProcessed)
		assert.Equal(t, 3, result.ProcessingStats.AddressValidated)
		assert.Equal(t, 2, result.TotalValid)
		assert.Equal(t, 1, result.TotalInvalid)
	})

	t.Run("unsupported contact type", func(t *testing.T) {
		contacts := []string{"test"}

		result, err := standardizer.ValidateBatch(ctx, contacts, "unsupported")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported contact type")
		assert.Nil(t, result)
	})

	t.Run("batch size limit exceeded", func(t *testing.T) {
		config := getDefaultContactValidationConfig()
		config.MaxBatchSize = 2
		standardizer := NewContactValidationStandardizer(config, logger)

		contacts := []string{"1", "2", "3"} // Exceeds limit of 2

		result, err := standardizer.ValidateBatch(ctx, contacts, "phone")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "batch size")
		assert.Nil(t, result)
	})

	t.Run("context timeout during batch", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(2 * time.Nanosecond) // Ensure timeout

		contacts := []string{"+1234567890"}

		result, err := standardizer.ValidateBatch(ctx, contacts, "phone")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "batch validation cancelled")
		assert.Nil(t, result)
	})
}

func TestPhoneValidationHelpers(t *testing.T) {
	logger := zap.NewNop()
	standardizer := NewContactValidationStandardizer(nil, logger)

	t.Run("clean phone number", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"+1 (234) 567-8900", "+1 (234) 567-8900"},
			{"234.567.8900", "234.567.8900"},
			{"234@567#8900", "2345678900"},
			{"  +1 234 567 8900  ", "+1 234 567 8900"},
		}

		for _, test := range tests {
			result := standardizer.cleanPhoneNumber(test.input)
			assert.Equal(t, test.expected, result)
		}
	})

	t.Run("validate phone format", func(t *testing.T) {
		tests := []struct {
			input         string
			expectedValid bool
			minScore      float64
		}{
			{"+1234567890", true, 1.0},
			{"+1 234 567 8900", true, 0.9},
			{"(123) 456-7890", true, 0.8},
			{"123-456-7890", true, 0.7},
			{"1234567890", true, 0.6},
			{"invalid", false, 0.0},
		}

		for _, test := range tests {
			valid, score := standardizer.validatePhoneFormat(test.input)
			assert.Equal(t, test.expectedValid, valid, "Input: %s", test.input)
			if test.expectedValid {
				assert.GreaterOrEqual(t, score, test.minScore, "Input: %s", test.input)
			}
		}
	})

	t.Run("standardize phone to E.164", func(t *testing.T) {
		config := getDefaultContactValidationConfig()
		config.DefaultCountryCode = "1"
		standardizer := NewContactValidationStandardizer(config, logger)

		tests := []struct {
			input    string
			expected string
		}{
			{"+1234567890", "+1234567890"},
			{"1234567890", "+11234567890"},
			{"+49123456789", "+49123456789"},
		}

		for _, test := range tests {
			result := standardizer.standardizePhoneToE164(test.input)
			assert.Equal(t, test.expected, result)
		}
	})

	t.Run("extract country code", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"+1234567890", "1"},
			{"+44123456789", "44"},
			{"+49123456789", "49"},
			{"1234567890", ""},
		}

		for _, test := range tests {
			result := standardizer.extractCountryCode(test.input)
			assert.Equal(t, test.expected, result)
		}
	})

	t.Run("detect phone format", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"+1234567890", "E.164"},
			{"(123) 456-7890", "US_STANDARD"},
			{"123-456-7890", "US_HYPHENATED"},
			{"invalid", "UNKNOWN"},
		}

		for _, test := range tests {
			result := standardizer.detectPhoneFormat(test.input)
			assert.Equal(t, test.expected, result)
		}
	})

	t.Run("detect line type", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"+18001234567", "TOLL_FREE"},
			{"+18881234567", "TOLL_FREE"},
			{"+1234567890", "STANDARD"},
		}

		for _, test := range tests {
			result := standardizer.detectLineType(test.input)
			assert.Equal(t, test.expected, result)
		}
	})
}

func TestEmailValidationHelpers(t *testing.T) {
	logger := zap.NewNop()
	standardizer := NewContactValidationStandardizer(nil, logger)

	t.Run("clean email address", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"User@Example.COM", "user@example.com"},
			{"  test@domain.org  ", "test@domain.org"},
			{"Mixed.Case@Domain.Net", "mixed.case@domain.net"},
		}

		for _, test := range tests {
			result := standardizer.cleanEmailAddress(test.input)
			assert.Equal(t, test.expected, result)
		}
	})

	t.Run("validate email format", func(t *testing.T) {
		tests := []struct {
			input         string
			expectedValid bool
			minScore      float64
		}{
			{"user@example.com", true, 0.8},
			{"test@domain.org", true, 0.8},
			{"user+alias@domain.net", true, 0.8},
			{"user..name@domain.com", true, 0.6}, // Valid but suspicious
			{"invalid-email", false, 0.0},
			{"user@", false, 0.0},
			{"@domain.com", false, 0.0},
		}

		for _, test := range tests {
			valid, score := standardizer.validateEmailFormat(test.input)
			assert.Equal(t, test.expectedValid, valid, "Input: %s", test.input)
			if test.expectedValid {
				assert.GreaterOrEqual(t, score, test.minScore, "Input: %s", test.input)
			}
		}
	})

	t.Run("standardize email", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"User@Domain.COM", "user@domain.com"},
			{"user.name+alias@gmail.com", "username@gmail.com"},
			{"test@otherdomain.com", "test@otherdomain.com"},
		}

		for _, test := range tests {
			result := standardizer.standardizeEmail(test.input)
			assert.Equal(t, test.expected, result)
		}
	})

	t.Run("extract domain", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"user@example.com", "example.com"},
			{"test@sub.domain.org", "sub.domain.org"},
			{"invalid-email", ""},
		}

		for _, test := range tests {
			result := standardizer.extractDomain(test.input)
			assert.Equal(t, test.expected, result)
		}
	})

	t.Run("domain blocked/trusted checks", func(t *testing.T) {
		assert.True(t, standardizer.isDomainBlocked("tempmail.com"))
		assert.False(t, standardizer.isDomainBlocked("example.com"))

		assert.True(t, standardizer.isDomainTrusted("gmail.com"))
		assert.False(t, standardizer.isDomainTrusted("tempmail.com"))
	})

	t.Run("detect email provider", func(t *testing.T) {
		tests := []struct {
			domain   string
			expected string
		}{
			{"gmail.com", "Google"},
			{"yahoo.com", "Yahoo"},
			{"outlook.com", "Microsoft"},
			{"example.com", "UNKNOWN"},
		}

		for _, test := range tests {
			result := standardizer.detectEmailProvider(test.domain)
			assert.Equal(t, test.expected, result)
		}
	})
}

func TestAddressValidationHelpers(t *testing.T) {
	logger := zap.NewNop()
	standardizer := NewContactValidationStandardizer(nil, logger)

	t.Run("clean address", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"  123   Main   St  ", "123 Main St"},
			{"456\tOak\nAve", "456 Oak Ave"},
			{"123 Main St, City, State", "123 Main St, City, State"},
		}

		for _, test := range tests {
			result := standardizer.cleanAddress(test.input)
			assert.Equal(t, test.expected, result)
		}
	})

	t.Run("parse address components", func(t *testing.T) {
		address := "123 Main St, Anytown, CA 12345"
		components := standardizer.parseAddressComponents(address)

		assert.Equal(t, "123 Main St", components["street"])
		assert.Equal(t, "Anytown", components["city"])
		assert.Equal(t, "CA", components["region"])
		assert.Equal(t, "12345", components["postal_code"])
		assert.Equal(t, "US", components["country"])
	})

	t.Run("validate address format", func(t *testing.T) {
		tests := []struct {
			address    string
			components map[string]string
			minScore   float64
		}{
			{
				"123 Main St, Anytown, CA 12345",
				map[string]string{
					"street":      "123 Main St",
					"city":        "Anytown",
					"region":      "CA",
					"postal_code": "12345",
				},
				0.9,
			},
			{
				"123 Main St",
				map[string]string{
					"street": "123 Main St",
				},
				0.4,
			},
		}

		for _, test := range tests {
			valid, score := standardizer.validateAddressFormat(test.address, test.components)
			assert.True(t, valid == (score >= 0.5))
			assert.GreaterOrEqual(t, score, test.minScore)
		}
	})

	t.Run("validate postal code", func(t *testing.T) {
		tests := []struct {
			postalCode string
			country    string
			expected   bool
		}{
			{"12345", "US", true},
			{"12345-6789", "US", true},
			{"ABC123", "US", false},
			{"SW1A 1AA", "GB", true},
			{"12345", "DE", true},
			{"INVALID", "XX", false},
		}

		for _, test := range tests {
			result := standardizer.validatePostalCode(test.postalCode, test.country)
			assert.Equal(t, test.expected, result, "Postal: %s, Country: %s", test.postalCode, test.country)
		}
	})

	t.Run("detect address format", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"123 Main St, City, State", "COMMA_SEPARATED"},
			{"123 Main St\nCity\nState", "MULTI_LINE"},
			{"123 Main St City State", "SINGLE_LINE"},
		}

		for _, test := range tests {
			result := standardizer.detectAddressFormat(test.input)
			assert.Equal(t, test.expected, result)
		}
	})
}

func TestQualityMetrics(t *testing.T) {
	logger := zap.NewNop()
	standardizer := NewContactValidationStandardizer(nil, logger)

	t.Run("phone quality metrics", func(t *testing.T) {
		tests := []struct {
			phone      string
			minQuality float64
		}{
			{"+1234567890", 0.8},
			{"1234567890", 0.6},
			{"123456", 0.3},
		}

		for _, test := range tests {
			completeness := standardizer.calculatePhoneCompleteness(test.phone)
			accuracy := standardizer.calculatePhoneAccuracy(test.phone)
			trustScore := standardizer.calculatePhoneTrustScore(test.phone)

			metrics := ContactQualityMetrics{
				DataCompleteness: completeness,
				Accuracy:         accuracy,
				TrustScore:       trustScore,
			}

			quality := standardizer.calculateOverallPhoneQuality(metrics)
			assert.GreaterOrEqual(t, quality, test.minQuality, "Phone: %s", test.phone)
		}
	})

	t.Run("email quality metrics", func(t *testing.T) {
		tests := []struct {
			email      string
			minQuality float64
		}{
			{"user@gmail.com", 0.7},
			{"test@example.com", 0.6},
			{"user@tempmail.com", 0.3},
		}

		for _, test := range tests {
			completeness := standardizer.calculateEmailCompleteness(test.email)
			accuracy := standardizer.calculateEmailAccuracy(test.email, true)
			domain := standardizer.extractDomain(test.email)
			trustScore := standardizer.calculateEmailTrustScore(domain, standardizer.isDomainTrusted(domain))

			metrics := ContactQualityMetrics{
				DataCompleteness: completeness,
				Accuracy:         accuracy,
				TrustScore:       trustScore,
				Deliverability:   0.8, // Assume good deliverability
			}

			quality := standardizer.calculateOverallEmailQuality(metrics)
			assert.GreaterOrEqual(t, quality, test.minQuality, "Email: %s", test.email)
		}
	})

	t.Run("address quality metrics", func(t *testing.T) {
		addresses := []string{
			"123 Main St, Anytown, CA 12345",
			"123 Main St",
		}

		for _, address := range addresses {
			components := standardizer.parseAddressComponents(address)
			completeness := standardizer.calculateAddressCompleteness(components)
			accuracy := standardizer.calculateAddressAccuracy(address, components)
			trustScore := standardizer.calculateAddressTrustScore(components)

			metrics := ContactQualityMetrics{
				DataCompleteness: completeness,
				Accuracy:         accuracy,
				TrustScore:       trustScore,
			}

			quality := standardizer.calculateOverallAddressQuality(metrics)
			assert.GreaterOrEqual(t, quality, 0.0)
			assert.LessOrEqual(t, quality, 1.0)
		}
	})
}

func TestConfigurationManagement(t *testing.T) {
	logger := zap.NewNop()
	standardizer := NewContactValidationStandardizer(nil, logger)

	t.Run("update config", func(t *testing.T) {
		newConfig := &ContactValidationConfig{
			EnablePhoneValidation:   false,
			MinValidationConfidence: 0.9,
		}

		err := standardizer.UpdateConfig(newConfig)
		assert.NoError(t, err)
		assert.False(t, standardizer.config.EnablePhoneValidation)
		assert.Equal(t, 0.9, standardizer.config.MinValidationConfidence)
	})

	t.Run("update config with nil", func(t *testing.T) {
		err := standardizer.UpdateConfig(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("get config", func(t *testing.T) {
		config := standardizer.GetConfig()
		assert.NotNil(t, config)
		assert.Equal(t, standardizer.config, config)
	})
}

func TestDefaultConfiguration(t *testing.T) {
	config := getDefaultContactValidationConfig()

	t.Run("default validation settings", func(t *testing.T) {
		assert.True(t, config.EnablePhoneValidation)
		assert.True(t, config.EnableE164Format)
		assert.True(t, config.EnableEmailValidation)
		assert.True(t, config.EnableDomainValidation)
		assert.False(t, config.EnableMXValidation) // Disabled by default
		assert.True(t, config.EnableAddressValidation)
		assert.False(t, config.EnableGeocoding) // Disabled by default
		assert.True(t, config.EnablePostalCodeValidation)
	})

	t.Run("default standardization settings", func(t *testing.T) {
		assert.True(t, config.EnablePhoneStandardization)
		assert.True(t, config.EnableEmailStandardization)
		assert.True(t, config.EnableAddressStandardization)
	})

	t.Run("default quality settings", func(t *testing.T) {
		assert.Equal(t, 0.7, config.MinValidationConfidence)
		assert.False(t, config.EnableFuzzyMatching)
		assert.False(t, config.EnableAutoCorrection)
	})

	t.Run("default performance settings", func(t *testing.T) {
		assert.Equal(t, 30*time.Second, config.ValidationTimeout)
		assert.Equal(t, 1000, config.MaxBatchSize)
		assert.True(t, config.EnableCaching)
	})

	t.Run("default domain lists", func(t *testing.T) {
		assert.Contains(t, config.BlockedDomains, "tempmail.com")
		assert.Contains(t, config.TrustedDomains, "gmail.com")
		assert.Contains(t, config.SupportedCountries, "US")
	})

	t.Run("default country settings", func(t *testing.T) {
		assert.Equal(t, "1", config.DefaultCountryCode)
		assert.Empty(t, config.AllowedCountryCodes) // Empty means all allowed
	})
}
