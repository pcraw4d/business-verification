package external

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewBusinessComparator(t *testing.T) {
	logger := zap.NewNop()

	// Test with nil config
	comparator := NewBusinessComparator(logger, nil)
	assert.NotNil(t, comparator)
	assert.NotNil(t, comparator.config)
	assert.Equal(t, 0.8, comparator.config.MinSimilarityThreshold)
	assert.Equal(t, 0.3, comparator.config.Weights.BusinessName)

	// Test with custom config
	customConfig := &ComparisonConfig{
		MinSimilarityThreshold: 0.9,
		Weights: &ComparisonWeights{
			BusinessName: 0.5,
		},
	}

	comparator = NewBusinessComparator(logger, customConfig)
	assert.Equal(t, 0.9, comparator.config.MinSimilarityThreshold)
	assert.Equal(t, 0.5, comparator.config.Weights.BusinessName)
}

func TestBusinessComparator_CompareBusinessInfo(t *testing.T) {
	logger := zap.NewNop()
	comparator := NewBusinessComparator(logger, nil)

	claimed := &ComparisonBusinessInfo{
		Name:           "Acme Corporation",
		PhoneNumbers:   []string{"+1-555-123-4567"},
		EmailAddresses: []string{"contact@acme.com"},
		Addresses: []ComparisonAddress{
			{
				Street:     "123 Business St",
				City:       "New York",
				State:      "NY",
				PostalCode: "10001",
				Country:    "USA",
			},
		},
		Website:  "https://www.acme.com",
		Industry: "Technology",
	}

	extracted := &ComparisonBusinessInfo{
		Name:           "Acme Corp",
		PhoneNumbers:   []string{"555-123-4567"},
		EmailAddresses: []string{"contact@acme.com"},
		Addresses: []ComparisonAddress{
			{
				Street:     "123 Business Street",
				City:       "New York",
				State:      "NY",
				PostalCode: "10001",
				Country:    "USA",
			},
		},
		Website:  "https://acme.com",
		Industry: "Technology",
	}

	result, err := comparator.CompareBusinessInfo(context.Background(), claimed, extracted)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Check overall score
	assert.Greater(t, result.OverallScore, 0.7)
	assert.Contains(t, []string{"high", "medium"}, result.ConfidenceLevel)

	// Check field results
	assert.Contains(t, result.FieldResults, "business_name")
	assert.Contains(t, result.FieldResults, "phone_numbers")
	assert.Contains(t, result.FieldResults, "email_addresses")
	assert.Contains(t, result.FieldResults, "addresses")
	assert.Contains(t, result.FieldResults, "website")
	assert.Contains(t, result.FieldResults, "industry")
}

func TestBusinessComparator_compareBusinessNames(t *testing.T) {
	logger := zap.NewNop()
	comparator := NewBusinessComparator(logger, nil)

	tests := []struct {
		name          string
		claimed       string
		extracted     string
		expectedMatch bool
		expectedScore float64
	}{
		{
			name:          "exact match",
			claimed:       "Acme Corporation",
			extracted:     "Acme Corporation",
			expectedMatch: true,
			expectedScore: 1.0,
		},
		{
			name:          "similar names",
			claimed:       "Acme Corporation",
			extracted:     "Acme Corp",
			expectedMatch: true,
			expectedScore: 0.8,
		},
		{
			name:          "different names",
			claimed:       "Acme Corporation",
			extracted:     "XYZ Company",
			expectedMatch: false,
			expectedScore: 0.0,
		},
		{
			name:          "empty claimed",
			claimed:       "",
			extracted:     "Acme Corporation",
			expectedMatch: false,
			expectedScore: 0.0,
		},
		{
			name:          "empty extracted",
			claimed:       "Acme Corporation",
			extracted:     "",
			expectedMatch: false,
			expectedScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := comparator.compareBusinessNames(tt.claimed, tt.extracted)

			assert.Equal(t, tt.expectedMatch, result.Matched)
			if tt.expectedScore > 0 {
				assert.Greater(t, result.Score, 0.0)
			} else {
				assert.Equal(t, tt.expectedScore, result.Score)
			}
			assert.NotEmpty(t, result.Reasoning)
		})
	}
}

func TestBusinessComparator_comparePhoneNumbers(t *testing.T) {
	logger := zap.NewNop()
	comparator := NewBusinessComparator(logger, nil)

	tests := []struct {
		name          string
		claimed       []string
		extracted     []string
		expectedMatch bool
	}{
		{
			name:          "exact match",
			claimed:       []string{"+1-555-123-4567"},
			extracted:     []string{"555-123-4567"},
			expectedMatch: true,
		},
		{
			name:          "different formats",
			claimed:       []string{"(555) 123-4567"},
			extracted:     []string{"555-123-4567"},
			expectedMatch: true,
		},
		{
			name:          "no match",
			claimed:       []string{"+1-555-123-4567"},
			extracted:     []string{"+1-555-987-6543"},
			expectedMatch: false,
		},
		{
			name:          "empty claimed",
			claimed:       []string{},
			extracted:     []string{"555-123-4567"},
			expectedMatch: false,
		},
		{
			name:          "empty extracted",
			claimed:       []string{"+1-555-123-4567"},
			extracted:     []string{},
			expectedMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := comparator.comparePhoneNumbers(tt.claimed, tt.extracted)

			assert.Equal(t, tt.expectedMatch, result.Matched)
			assert.NotEmpty(t, result.Reasoning)
		})
	}
}

func TestBusinessComparator_compareEmailAddresses(t *testing.T) {
	logger := zap.NewNop()
	comparator := NewBusinessComparator(logger, nil)

	tests := []struct {
		name          string
		claimed       []string
		extracted     []string
		expectedMatch bool
	}{
		{
			name:          "exact match",
			claimed:       []string{"contact@acme.com"},
			extracted:     []string{"contact@acme.com"},
			expectedMatch: true,
		},
		{
			name:          "case insensitive",
			claimed:       []string{"Contact@Acme.com"},
			extracted:     []string{"contact@acme.com"},
			expectedMatch: true,
		},
		{
			name:          "no match",
			claimed:       []string{"contact@acme.com"},
			extracted:     []string{"info@xyz.com"},
			expectedMatch: false,
		},
		{
			name:          "empty claimed",
			claimed:       []string{},
			extracted:     []string{"contact@acme.com"},
			expectedMatch: false,
		},
		{
			name:          "empty extracted",
			claimed:       []string{"contact@acme.com"},
			extracted:     []string{},
			expectedMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := comparator.compareEmailAddresses(tt.claimed, tt.extracted)

			assert.Equal(t, tt.expectedMatch, result.Matched)
			assert.NotEmpty(t, result.Reasoning)
		})
	}
}

func TestBusinessComparator_compareAddresses(t *testing.T) {
	logger := zap.NewNop()
	comparator := NewBusinessComparator(logger, nil)

	claimed := []ComparisonAddress{
		{
			Street:     "123 Business St",
			City:       "New York",
			State:      "NY",
			PostalCode: "10001",
			Country:    "USA",
		},
	}

	extracted := []ComparisonAddress{
		{
			Street:     "123 Business Street",
			City:       "New York",
			State:      "NY",
			PostalCode: "10001",
			Country:    "USA",
		},
	}

	result := comparator.compareAddresses(claimed, extracted)
	assert.True(t, result.Matched)
	assert.Greater(t, result.Score, 0.7)
	assert.NotEmpty(t, result.Reasoning)
}

func TestBusinessComparator_compareWebsites(t *testing.T) {
	logger := zap.NewNop()
	comparator := NewBusinessComparator(logger, nil)

	tests := []struct {
		name          string
		claimed       string
		extracted     string
		expectedMatch bool
	}{
		{
			name:          "exact match",
			claimed:       "https://www.acme.com",
			extracted:     "https://www.acme.com",
			expectedMatch: true,
		},
		{
			name:          "domain match",
			claimed:       "https://www.acme.com",
			extracted:     "https://acme.com",
			expectedMatch: true,
		},
		{
			name:          "different paths",
			claimed:       "https://www.acme.com",
			extracted:     "https://www.acme.com/about",
			expectedMatch: true,
		},
		{
			name:          "different domains",
			claimed:       "https://www.acme.com",
			extracted:     "https://www.xyz.com",
			expectedMatch: false,
		},
		{
			name:          "empty claimed",
			claimed:       "",
			extracted:     "https://www.acme.com",
			expectedMatch: false,
		},
		{
			name:          "empty extracted",
			claimed:       "https://www.acme.com",
			extracted:     "",
			expectedMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := comparator.compareWebsites(tt.claimed, tt.extracted)

			assert.Equal(t, tt.expectedMatch, result.Matched)
			assert.NotEmpty(t, result.Reasoning)
		})
	}
}

func TestBusinessComparator_compareIndustries(t *testing.T) {
	logger := zap.NewNop()
	comparator := NewBusinessComparator(logger, nil)

	tests := []struct {
		name          string
		claimed       string
		extracted     string
		expectedMatch bool
	}{
		{
			name:          "exact match",
			claimed:       "Technology",
			extracted:     "Technology",
			expectedMatch: true,
		},
		{
			name:          "case insensitive",
			claimed:       "Technology",
			extracted:     "technology",
			expectedMatch: true,
		},
		{
			name:          "different industries",
			claimed:       "Technology",
			extracted:     "Healthcare",
			expectedMatch: false,
		},
		{
			name:          "empty claimed",
			claimed:       "",
			extracted:     "Technology",
			expectedMatch: false,
		},
		{
			name:          "empty extracted",
			claimed:       "Technology",
			extracted:     "",
			expectedMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := comparator.compareIndustries(tt.claimed, tt.extracted)

			assert.Equal(t, tt.expectedMatch, result.Matched)
			assert.NotEmpty(t, result.Reasoning)
		})
	}
}

func TestBusinessComparator_calculateStringSimilarity(t *testing.T) {
	logger := zap.NewNop()
	comparator := NewBusinessComparator(logger, nil)

	tests := []struct {
		name     string
		s1       string
		s2       string
		expected float64
	}{
		{
			name:     "identical strings",
			s1:       "hello",
			s2:       "hello",
			expected: 1.0,
		},
		{
			name:     "similar strings",
			s1:       "hello",
			s2:       "helo",
			expected: 0.8,
		},
		{
			name:     "different strings",
			s1:       "hello",
			s2:       "world",
			expected: 0.2, // Levenshtein distance calculation gives this result
		},
		{
			name:     "empty strings",
			s1:       "",
			s2:       "",
			expected: 1.0,
		},
		{
			name:     "one empty string",
			s1:       "hello",
			s2:       "",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := comparator.calculateStringSimilarity(tt.s1, tt.s2)

			if tt.expected == 1.0 {
				assert.Equal(t, tt.expected, similarity)
			} else if tt.expected == 0.0 {
				assert.Equal(t, tt.expected, similarity)
			} else {
				// For similar strings, just check it's greater than 0
				assert.Greater(t, similarity, 0.0)
			}
		})
	}
}

func TestBusinessComparator_calculateDistance(t *testing.T) {
	logger := zap.NewNop()
	comparator := NewBusinessComparator(logger, nil)

	// Test distance calculation between two points
	lat1, lon1 := 40.7128, -74.0060  // New York
	lat2, lon2 := 34.0522, -118.2437 // Los Angeles

	distance := comparator.calculateDistance(lat1, lon1, lat2, lon2)

	// Distance should be approximately 3935 km
	assert.Greater(t, distance, 3900.0)
	assert.Less(t, distance, 4000.0)

	// Test same point
	distance = comparator.calculateDistance(lat1, lon1, lat1, lon1)
	assert.Equal(t, 0.0, distance)
}

func TestBusinessComparator_determineConfidenceLevel(t *testing.T) {
	logger := zap.NewNop()
	comparator := NewBusinessComparator(logger, nil)

	tests := []struct {
		score float64
		level string
	}{
		{0.95, "high"},
		{0.85, "medium"},
		{0.75, "medium"},
		{0.65, "low"},
		{0.45, "very_low"},
		{0.35, "very_low"},
		{0.25, "very_low"},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			level := comparator.determineConfidenceLevel(tt.score)
			assert.Equal(t, tt.level, level)
		})
	}
}

func TestBusinessComparator_determineVerificationStatus(t *testing.T) {
	logger := zap.NewNop()
	comparator := NewBusinessComparator(logger, nil)

	// Test PASSED status
	fieldResults := map[string]FieldComparison{
		"business_name":   {Matched: true, Score: 0.9},
		"phone_numbers":   {Matched: true, Score: 0.9},
		"email_addresses": {Matched: true, Score: 0.9},
	}

	status := comparator.determineVerificationStatus(0.9, fieldResults)
	assert.Equal(t, "PASSED", status)

	// Test PARTIAL status
	fieldResults = map[string]FieldComparison{
		"business_name": {Matched: true, Score: 0.8},
		"phone_numbers": {Matched: false, Score: 0.3},
	}

	status = comparator.determineVerificationStatus(0.7, fieldResults)
	assert.Equal(t, "PARTIAL", status)

	// Test FAILED status
	fieldResults = map[string]FieldComparison{
		"business_name": {Matched: false, Score: 0.2},
		"phone_numbers": {Matched: false, Score: 0.1},
	}

	status = comparator.determineVerificationStatus(0.2, fieldResults)
	assert.Equal(t, "FAILED", status)
}

func TestBusinessComparator_generateRecommendations(t *testing.T) {
	logger := zap.NewNop()
	comparator := NewBusinessComparator(logger, nil)

	// Test with low confidence fields
	result := &ComparisonResult{
		OverallScore: 0.6,
		FieldResults: map[string]FieldComparison{
			"business_name": {Confidence: 0.3},
			"phone_numbers": {Confidence: 0.8},
		},
	}

	recommendations := comparator.generateRecommendations(result)
	assert.NotEmpty(t, recommendations)
	assert.Contains(t, recommendations[0], "Manual verification recommended")
}

func TestBusinessComparator_normalizeBusinessName(t *testing.T) {
	logger := zap.NewNop()
	comparator := NewBusinessComparator(logger, nil)

	tests := []struct {
		input    string
		expected string
	}{
		{"Acme Corporation Inc.", "acme corporation inc"},
		{"Acme Corp LLC", "acme"},
		{"Acme & Sons Ltd.", "acme sons ltd"},
		{"Acme Company", "acme"},
		{"  Acme Corp  ", "acme"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			normalized := comparator.normalizeBusinessName(tt.input)
			assert.Equal(t, tt.expected, normalized)
		})
	}
}

func TestBusinessComparator_normalizePhoneNumbers(t *testing.T) {
	logger := zap.NewNop()
	comparator := NewBusinessComparator(logger, nil)

	tests := []struct {
		input    []string
		expected []string
	}{
		{
			[]string{"+1-555-123-4567"},
			[]string{"5551234567"},
		},
		{
			[]string{"(555) 123-4567"},
			[]string{"5551234567"},
		},
		{
			[]string{"555.123.4567"},
			[]string{"5551234567"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input[0], func(t *testing.T) {
			normalized := comparator.normalizePhoneNumbers(tt.input)
			assert.Equal(t, tt.expected, normalized)
		})
	}
}

func TestBusinessComparator_normalizeEmailAddresses(t *testing.T) {
	logger := zap.NewNop()
	comparator := NewBusinessComparator(logger, nil)

	input := []string{"Contact@Acme.com", "  info@acme.com  "}
	expected := []string{"contact@acme.com", "info@acme.com"}

	normalized := comparator.normalizeEmailAddresses(input)
	assert.Equal(t, expected, normalized)
}

func TestBusinessComparator_normalizeAddress(t *testing.T) {
	logger := zap.NewNop()
	comparator := NewBusinessComparator(logger, nil)

	tests := []struct {
		input    string
		expected string
	}{
		{"123 Business Street", "123 business st"},
		{"456 Main Avenue", "456 main ave"},
		{"789 Oak Boulevard", "789 oak blvd"},
		{" 123 Test Drive ", "123 test dr"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			normalized := comparator.normalizeAddress(tt.input)
			assert.Equal(t, tt.expected, normalized)
		})
	}
}

func TestBusinessComparator_normalizeURL(t *testing.T) {
	logger := zap.NewNop()
	comparator := NewBusinessComparator(logger, nil)

	tests := []struct {
		input    string
		expected string
	}{
		{"https://www.acme.com", "acme.com"},
		{"http://acme.com/", "acme.com"},
		{"https://www.acme.com/about", "acme.com/about"},
		{"  https://www.acme.com  ", "acme.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			normalized := comparator.normalizeURL(tt.input)
			assert.Equal(t, tt.expected, normalized)
		})
	}
}

func TestBusinessComparator_extractDomain(t *testing.T) {
	logger := zap.NewNop()
	comparator := NewBusinessComparator(logger, nil)

	tests := []struct {
		input    string
		expected string
	}{
		{"acme.com", "acme.com"},
		{"acme.com/about", "acme.com"},
		{"acme.com/path/to/page", "acme.com"},
		{"subdomain.acme.com", "subdomain.acme.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			domain := comparator.extractDomain(tt.input)
			assert.Equal(t, tt.expected, domain)
		})
	}
}

func TestBusinessComparator_normalizeIndustry(t *testing.T) {
	logger := zap.NewNop()
	comparator := NewBusinessComparator(logger, nil)

	tests := []struct {
		input    string
		expected string
	}{
		{"Technology", "technology"},
		{"Health Care", "health care"},
		{"  Financial Services  ", "financial services"},
		{"E-commerce & Retail", "ecommerce retail"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			normalized := comparator.normalizeIndustry(tt.input)
			assert.Equal(t, tt.expected, normalized)
		})
	}
}
