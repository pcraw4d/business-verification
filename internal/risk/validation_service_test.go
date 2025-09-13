package risk

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRiskValidationService_ValidateRiskAssessment(t *testing.T) {
	logger := zap.NewNop()
	service := NewRiskValidationService(logger)
	ctx := context.WithValue(context.Background(), "request_id", "test-request")

	tests := []struct {
		name             string
		assessment       *RiskAssessment
		expectedValid    bool
		expectedErrors   int
		expectedWarnings int
	}{
		{
			name: "valid assessment",
			assessment: &RiskAssessment{
				ID:               "test-123",
				BusinessID:       "business-123",
				BusinessName:     "Test Company",
				OverallScore:     0.5,
				OverallLevel:     RiskLevelMedium,
				CategoryScores:   map[string]interface{}{"financial": 0.6, "operational": 0.4},
				FactorScores:     []string{"factor1", "factor2"},
				Recommendations:  []string{"rec1", "rec2"},
				Alerts:           []string{"alert1"},
				AssessmentMethod: "automated",
				Source:           "api",
				AssessedAt:       time.Now(),
				ValidUntil:       time.Now().Add(24 * time.Hour),
			},
			expectedValid:    true,
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "missing required fields",
			assessment: &RiskAssessment{
				ID:           "",
				BusinessID:   "",
				BusinessName: "",
				OverallScore: 0.5,
				OverallLevel: RiskLevelMedium,
			},
			expectedValid:    false,
			expectedErrors:   3, // ID, BusinessID, BusinessName
			expectedWarnings: 0,
		},
		{
			name: "invalid score range",
			assessment: &RiskAssessment{
				ID:           "test-123",
				BusinessID:   "business-123",
				BusinessName: "Test Company",
				OverallScore: 1.5, // Invalid: > 1
				OverallLevel: RiskLevelMedium,
				AssessedAt:   time.Now(),
				ValidUntil:   time.Now().Add(24 * time.Hour),
			},
			expectedValid:    false,
			expectedErrors:   1, // Invalid score
			expectedWarnings: 0,
		},
		{
			name: "invalid timestamp logic",
			assessment: &RiskAssessment{
				ID:           "test-123",
				BusinessID:   "business-123",
				BusinessName: "Test Company",
				OverallScore: 0.5,
				OverallLevel: RiskLevelMedium,
				AssessedAt:   time.Now(),
				ValidUntil:   time.Now().Add(-24 * time.Hour), // Invalid: before assessed_at
			},
			expectedValid:    false,
			expectedErrors:   1, // Invalid date range
			expectedWarnings: 0,
		},
		{
			name: "missing timestamps",
			assessment: &RiskAssessment{
				ID:           "test-123",
				BusinessID:   "business-123",
				BusinessName: "Test Company",
				OverallScore: 0.5,
				OverallLevel: RiskLevelMedium,
				AssessedAt:   time.Time{}, // Zero time
				ValidUntil:   time.Time{}, // Zero time
			},
			expectedValid:    false,
			expectedErrors:   2, // Missing timestamps
			expectedWarnings: 0,
		},
		{
			name: "level score mismatch",
			assessment: &RiskAssessment{
				ID:           "test-123",
				BusinessID:   "business-123",
				BusinessName: "Test Company",
				OverallScore: 0.9,          // High score
				OverallLevel: RiskLevelLow, // But low level
				AssessedAt:   time.Now(),
				ValidUntil:   time.Now().Add(24 * time.Hour),
			},
			expectedValid:    true,
			expectedErrors:   0,
			expectedWarnings: 1, // Level score mismatch
		},
		{
			name: "empty collections with warnings",
			assessment: &RiskAssessment{
				ID:              "test-123",
				BusinessID:      "business-123",
				BusinessName:    "Test Company",
				OverallScore:    0.5,
				OverallLevel:    RiskLevelMedium,
				CategoryScores:  map[string]interface{}{}, // Empty
				FactorScores:    []string{},               // Empty
				Recommendations: []string{},               // Empty
				Alerts:          []string{},               // Empty
				AssessedAt:      time.Now(),
				ValidUntil:      time.Now().Add(24 * time.Hour),
			},
			expectedValid:    true,
			expectedErrors:   0,
			expectedWarnings: 4, // Empty collections
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ValidateRiskAssessment(ctx, tt.assessment)

			assert.Equal(t, tt.expectedValid, result.IsValid)
			assert.Equal(t, tt.expectedErrors, len(result.Errors))
			assert.Equal(t, tt.expectedWarnings, len(result.Warnings))

			if tt.expectedErrors > 0 {
				assert.False(t, result.IsValid)
			}
		})
	}
}

func TestRiskValidationService_ValidateRiskAlert(t *testing.T) {
	logger := zap.NewNop()
	service := NewRiskValidationService(logger)
	ctx := context.WithValue(context.Background(), "request_id", "test-request")

	tests := []struct {
		name             string
		alert            *RiskAlert
		expectedValid    bool
		expectedErrors   int
		expectedWarnings int
	}{
		{
			name: "valid alert",
			alert: &RiskAlert{
				ID:           "alert-123",
				BusinessID:   "business-123",
				RiskFactor:   "financial",
				Level:        RiskLevelHigh,
				Message:      "High financial risk detected",
				Score:        0.8,
				Threshold:    0.7,
				TriggeredAt:  time.Now(),
				Acknowledged: false,
			},
			expectedValid:    true,
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "missing required fields",
			alert: &RiskAlert{
				ID:         "",
				BusinessID: "",
				RiskFactor: "",
				Message:    "",
			},
			expectedValid:    false,
			expectedErrors:   4, // ID, BusinessID, RiskFactor, Message
			expectedWarnings: 0,
		},
		{
			name: "invalid risk level",
			alert: &RiskAlert{
				ID:          "alert-123",
				BusinessID:  "business-123",
				RiskFactor:  "financial",
				Level:       "invalid", // Invalid level
				Message:     "Test message",
				Score:       0.8,
				Threshold:   0.7,
				TriggeredAt: time.Now(),
			},
			expectedValid:    false,
			expectedErrors:   1, // Invalid level
			expectedWarnings: 0,
		},
		{
			name: "invalid score range",
			alert: &RiskAlert{
				ID:          "alert-123",
				BusinessID:  "business-123",
				RiskFactor:  "financial",
				Level:       RiskLevelHigh,
				Message:     "Test message",
				Score:       1.5, // Invalid: > 1
				Threshold:   0.7,
				TriggeredAt: time.Now(),
			},
			expectedValid:    false,
			expectedErrors:   1, // Invalid score
			expectedWarnings: 0,
		},
		{
			name: "invalid threshold range",
			alert: &RiskAlert{
				ID:          "alert-123",
				BusinessID:  "business-123",
				RiskFactor:  "financial",
				Level:       RiskLevelHigh,
				Message:     "Test message",
				Score:       0.8,
				Threshold:   -0.1, // Invalid: < 0
				TriggeredAt: time.Now(),
			},
			expectedValid:    false,
			expectedErrors:   1, // Invalid threshold
			expectedWarnings: 0,
		},
		{
			name: "missing triggered timestamp",
			alert: &RiskAlert{
				ID:          "alert-123",
				BusinessID:  "business-123",
				RiskFactor:  "financial",
				Level:       RiskLevelHigh,
				Message:     "Test message",
				Score:       0.8,
				Threshold:   0.7,
				TriggeredAt: time.Time{}, // Zero time
			},
			expectedValid:    false,
			expectedErrors:   1, // Missing timestamp
			expectedWarnings: 0,
		},
		{
			name: "acknowledged without timestamp",
			alert: &RiskAlert{
				ID:           "alert-123",
				BusinessID:   "business-123",
				RiskFactor:   "financial",
				Level:        RiskLevelHigh,
				Message:      "Test message",
				Score:        0.8,
				Threshold:    0.7,
				TriggeredAt:  time.Now(),
				Acknowledged: true, // Acknowledged but no timestamp
			},
			expectedValid:    false,
			expectedErrors:   1, // Missing acknowledgment timestamp
			expectedWarnings: 0,
		},
		{
			name: "inconsistent acknowledgment",
			alert: &RiskAlert{
				ID:             "alert-123",
				BusinessID:     "business-123",
				RiskFactor:     "financial",
				Level:          RiskLevelHigh,
				Message:        "Test message",
				Score:          0.8,
				Threshold:      0.7,
				TriggeredAt:    time.Now(),
				Acknowledged:   false,                       // Not acknowledged
				AcknowledgedAt: &[]time.Time{time.Now()}[0], // But has timestamp
			},
			expectedValid:    true,
			expectedErrors:   0,
			expectedWarnings: 1, // Inconsistent acknowledgment
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ValidateRiskAlert(ctx, tt.alert)

			assert.Equal(t, tt.expectedValid, result.IsValid)
			assert.Equal(t, tt.expectedErrors, len(result.Errors))
			assert.Equal(t, tt.expectedWarnings, len(result.Warnings))

			if tt.expectedErrors > 0 {
				assert.False(t, result.IsValid)
			}
		})
	}
}

func TestRiskValidationService_ValidateRiskTrend(t *testing.T) {
	logger := zap.NewNop()
	service := NewRiskValidationService(logger)
	ctx := context.WithValue(context.Background(), "request_id", "test-request")

	tests := []struct {
		name             string
		trend            *RiskTrend
		expectedValid    bool
		expectedErrors   int
		expectedWarnings int
	}{
		{
			name: "valid trend",
			trend: &RiskTrend{
				BusinessID: "business-123",
				Direction:  "improving",
				Confidence: 0.8,
				Period:     30,
				AnalyzedAt: time.Now(),
			},
			expectedValid:    true,
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "missing business ID",
			trend: &RiskTrend{
				BusinessID: "", // Missing
				Direction:  "improving",
				Confidence: 0.8,
				Period:     30,
				AnalyzedAt: time.Now(),
			},
			expectedValid:    false,
			expectedErrors:   1, // Missing business ID
			expectedWarnings: 0,
		},
		{
			name: "invalid direction",
			trend: &RiskTrend{
				BusinessID: "business-123",
				Direction:  "invalid", // Invalid direction
				Confidence: 0.8,
				Period:     30,
				AnalyzedAt: time.Now(),
			},
			expectedValid:    false,
			expectedErrors:   1, // Invalid direction
			expectedWarnings: 0,
		},
		{
			name: "invalid confidence range",
			trend: &RiskTrend{
				BusinessID: "business-123",
				Direction:  "improving",
				Confidence: 1.5, // Invalid: > 1
				Period:     30,
				AnalyzedAt: time.Now(),
			},
			expectedValid:    false,
			expectedErrors:   1, // Invalid confidence
			expectedWarnings: 0,
		},
		{
			name: "invalid period",
			trend: &RiskTrend{
				BusinessID: "business-123",
				Direction:  "improving",
				Confidence: 0.8,
				Period:     -5, // Invalid: negative
				AnalyzedAt: time.Now(),
			},
			expectedValid:    false,
			expectedErrors:   1, // Invalid period
			expectedWarnings: 0,
		},
		{
			name: "missing analyzed timestamp",
			trend: &RiskTrend{
				BusinessID: "business-123",
				Direction:  "improving",
				Confidence: 0.8,
				Period:     30,
				AnalyzedAt: time.Time{}, // Zero time
			},
			expectedValid:    false,
			expectedErrors:   1, // Missing timestamp
			expectedWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ValidateRiskTrend(ctx, tt.trend)

			assert.Equal(t, tt.expectedValid, result.IsValid)
			assert.Equal(t, tt.expectedErrors, len(result.Errors))
			assert.Equal(t, tt.expectedWarnings, len(result.Warnings))

			if tt.expectedErrors > 0 {
				assert.False(t, result.IsValid)
			}
		})
	}
}

func TestRiskValidationService_ValidateBusinessData(t *testing.T) {
	logger := zap.NewNop()
	service := NewRiskValidationService(logger)
	ctx := context.WithValue(context.Background(), "request_id", "test-request")

	tests := []struct {
		name             string
		businessData     map[string]interface{}
		expectedValid    bool
		expectedErrors   int
		expectedWarnings int
	}{
		{
			name: "valid business data",
			businessData: map[string]interface{}{
				"name":    "Test Company",
				"address": "123 Test St, Test City, TC 12345",
				"email":   "test@example.com",
				"phone":   "+1-555-123-4567",
				"website": "https://www.example.com",
			},
			expectedValid:    true,
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "missing required name",
			businessData: map[string]interface{}{
				"address": "123 Test St",
				"email":   "test@example.com",
			},
			expectedValid:    false,
			expectedErrors:   1, // Missing name
			expectedWarnings: 0,
		},
		{
			name: "empty name",
			businessData: map[string]interface{}{
				"name":    "", // Empty name
				"address": "123 Test St",
			},
			expectedValid:    false,
			expectedErrors:   1, // Empty name
			expectedWarnings: 0,
		},
		{
			name: "name too long",
			businessData: map[string]interface{}{
				"name": strings.Repeat("A", 256), // Too long
			},
			expectedValid:    false,
			expectedErrors:   1, // Name too long
			expectedWarnings: 0,
		},
		{
			name: "invalid email format",
			businessData: map[string]interface{}{
				"name":  "Test Company",
				"email": "invalid-email", // Invalid format
			},
			expectedValid:    false,
			expectedErrors:   1, // Invalid email
			expectedWarnings: 0,
		},
		{
			name: "invalid phone format",
			businessData: map[string]interface{}{
				"name":  "Test Company",
				"phone": "123-456-7890", // Invalid format (not E.164)
			},
			expectedValid:    false,
			expectedErrors:   1, // Invalid phone
			expectedWarnings: 0,
		},
		{
			name: "invalid website URL",
			businessData: map[string]interface{}{
				"name":    "Test Company",
				"website": "not-a-url", // Invalid URL
			},
			expectedValid:    false,
			expectedErrors:   1, // Invalid URL
			expectedWarnings: 0,
		},
		{
			name: "empty address warning",
			businessData: map[string]interface{}{
				"name":    "Test Company",
				"address": "", // Empty address
			},
			expectedValid:    true,
			expectedErrors:   0,
			expectedWarnings: 1, // Empty address warning
		},
		{
			name: "suspicious content warning",
			businessData: map[string]interface{}{
				"name":    "Test Company",
				"address": "123 Test St <script>alert('xss')</script>", // Suspicious content
			},
			expectedValid:    true,
			expectedErrors:   0,
			expectedWarnings: 1, // Suspicious content warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ValidateBusinessData(ctx, tt.businessData)

			assert.Equal(t, tt.expectedValid, result.IsValid)
			assert.Equal(t, tt.expectedErrors, len(result.Errors))
			assert.Equal(t, tt.expectedWarnings, len(result.Warnings))

			if tt.expectedErrors > 0 {
				assert.False(t, result.IsValid)
			}
		})
	}
}

func TestRiskValidationService_ValidateRiskScore(t *testing.T) {
	logger := zap.NewNop()
	service := NewRiskValidationService(logger)
	ctx := context.WithValue(context.Background(), "request_id", "test-request")

	tests := []struct {
		name             string
		score            float64
		level            RiskLevel
		expectedValid    bool
		expectedErrors   int
		expectedWarnings int
	}{
		{
			name:             "valid score and level match",
			score:            0.5,
			level:            RiskLevelMedium,
			expectedValid:    true,
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name:             "score out of range - negative",
			score:            -0.1,
			level:            RiskLevelLow,
			expectedValid:    false,
			expectedErrors:   1, // Invalid score range
			expectedWarnings: 0,
		},
		{
			name:             "score out of range - too high",
			score:            1.1,
			level:            RiskLevelCritical,
			expectedValid:    false,
			expectedErrors:   1, // Invalid score range
			expectedWarnings: 0,
		},
		{
			name:             "level score mismatch",
			score:            0.9,          // High score
			level:            RiskLevelLow, // But low level
			expectedValid:    true,
			expectedErrors:   0,
			expectedWarnings: 1, // Level score mismatch
		},
		{
			name:             "zero score warning",
			score:            0.0,
			level:            RiskLevelLow,
			expectedValid:    true,
			expectedErrors:   0,
			expectedWarnings: 1, // Zero score warning
		},
		{
			name:             "maximum score warning",
			score:            1.0,
			level:            RiskLevelCritical,
			expectedValid:    true,
			expectedErrors:   0,
			expectedWarnings: 1, // Maximum score warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ValidateRiskScore(ctx, tt.score, tt.level)

			assert.Equal(t, tt.expectedValid, result.IsValid)
			assert.Equal(t, tt.expectedErrors, len(result.Errors))
			assert.Equal(t, tt.expectedWarnings, len(result.Warnings))

			if tt.expectedErrors > 0 {
				assert.False(t, result.IsValid)
			}
		})
	}
}

func TestRiskValidationService_HelperMethods(t *testing.T) {
	logger := zap.NewNop()
	service := NewRiskValidationService(logger)

	t.Run("calculateExpectedRiskLevel", func(t *testing.T) {
		tests := []struct {
			score    float64
			expected RiskLevel
		}{
			{0.0, RiskLevelLow},
			{0.2, RiskLevelLow},
			{0.3, RiskLevelMedium},
			{0.5, RiskLevelMedium},
			{0.6, RiskLevelHigh},
			{0.7, RiskLevelHigh},
			{0.8, RiskLevelCritical},
			{1.0, RiskLevelCritical},
		}

		for _, tt := range tests {
			result := service.calculateExpectedRiskLevel(tt.score)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("isValidEmail", func(t *testing.T) {
		tests := []struct {
			email    string
			expected bool
		}{
			{"test@example.com", true},
			{"user.name@domain.co.uk", true},
			{"invalid-email", false},
			{"@domain.com", false},
			{"user@", false},
			{"", false},
		}

		for _, tt := range tests {
			result := service.isValidEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("isValidPhone", func(t *testing.T) {
		tests := []struct {
			phone    string
			expected bool
		}{
			{"+1-555-123-4567", true},
			{"+44-20-7946-0958", true},
			{"123-456-7890", false}, // Not E.164
			{"+", false},
			{"", false},
		}

		for _, tt := range tests {
			result := service.isValidPhone(tt.phone)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("isValidURL", func(t *testing.T) {
		tests := []struct {
			url      string
			expected bool
		}{
			{"https://www.example.com", true},
			{"http://example.com", true},
			{"https://subdomain.example.com/path", true},
			{"not-a-url", false},
			{"ftp://example.com", false}, // Not HTTP/HTTPS
			{"", false},
		}

		for _, tt := range tests {
			result := service.isValidURL(tt.url)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("containsSuspiciousContent", func(t *testing.T) {
		tests := []struct {
			content  string
			expected bool
		}{
			{"<script>alert('xss')</script>", true},
			{"SELECT * FROM users", true},
			{"'; DROP TABLE users; --", true},
			{"<div>Hello World</div>", true},
			{"Normal business content", false},
			{"", false},
		}

		for _, tt := range tests {
			result := service.containsSuspiciousContent(tt.content)
			assert.Equal(t, tt.expected, result)
		}
	})
}

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{
		Field:   "test_field",
		Message: "test message",
		Code:    "TEST_CODE",
	}

	expected := "validation error in field 'test_field': test message"
	assert.Equal(t, expected, err.Error())
}
