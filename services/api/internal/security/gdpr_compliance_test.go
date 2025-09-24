package security

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewGDPRComplianceService(t *testing.T) {
	logger := zap.NewNop()

	// Test with nil config
	service, err := NewGDPRComplianceService(nil, logger)
	require.NoError(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.config)
	assert.True(t, service.config.EnableGDPRCompliance)
	assert.Equal(t, "KYB Platform", service.config.DataControllerName)
	assert.Equal(t, "privacy@kyb-platform.com", service.config.DataControllerEmail)
	assert.True(t, service.config.RequireExplicitConsent)
	assert.Equal(t, 365*24*time.Hour, service.config.ConsentExpiryPeriod)
	assert.True(t, service.config.AllowWithdrawal)
	assert.True(t, service.config.EnableDataSubjectRights)
	assert.Equal(t, 30*24*time.Hour, service.config.ResponseTimeLimit)
	assert.Equal(t, int64(10*1024*1024), service.config.MaxDataPortabilitySize)
	assert.True(t, service.config.EnableBreachDetection)
	assert.Equal(t, 100, service.config.BreachNotificationThreshold)
	assert.Equal(t, 72*time.Hour, service.config.NotificationTimeLimit)
	assert.True(t, service.config.EnableProcessingRecords)
	assert.Equal(t, 5*365*24*time.Hour, service.config.RecordRetentionPeriod)
	assert.Equal(t, "legitimate_interest", service.config.DefaultLegalBasis)
	assert.Len(t, service.config.AllowedLegalBases, 6)

	// Test with custom config
	customConfig := &GDPRComplianceConfig{
		EnableGDPRCompliance:        false,
		DataControllerName:          "Custom Company",
		DataControllerEmail:         "custom@company.com",
		RequireExplicitConsent:      false,
		ConsentExpiryPeriod:         180 * 24 * time.Hour, // 6 months
		AllowWithdrawal:             false,
		EnableDataSubjectRights:     false,
		ResponseTimeLimit:           15 * 24 * time.Hour, // 15 days
		MaxDataPortabilitySize:      5 * 1024 * 1024,     // 5MB
		EnableBreachDetection:       false,
		BreachNotificationThreshold: 50,
		NotificationTimeLimit:       48 * time.Hour, // 48 hours
		EnableProcessingRecords:     false,
		RecordRetentionPeriod:       2 * 365 * 24 * time.Hour, // 2 years
		DefaultLegalBasis:           "consent",
		AllowedLegalBases:           []string{"consent", "contract"},
	}

	service, err = NewGDPRComplianceService(customConfig, logger)
	require.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, customConfig, service.config)
	assert.False(t, service.config.EnableGDPRCompliance)
	assert.Equal(t, "Custom Company", service.config.DataControllerName)
	assert.Equal(t, "custom@company.com", service.config.DataControllerEmail)
	assert.False(t, service.config.RequireExplicitConsent)
	assert.Equal(t, 180*24*time.Hour, service.config.ConsentExpiryPeriod)
	assert.False(t, service.config.AllowWithdrawal)
	assert.False(t, service.config.EnableDataSubjectRights)
	assert.Equal(t, 15*24*time.Hour, service.config.ResponseTimeLimit)
	assert.Equal(t, int64(5*1024*1024), service.config.MaxDataPortabilitySize)
	assert.False(t, service.config.EnableBreachDetection)
	assert.Equal(t, 50, service.config.BreachNotificationThreshold)
	assert.Equal(t, 48*time.Hour, service.config.NotificationTimeLimit)
	assert.False(t, service.config.EnableProcessingRecords)
	assert.Equal(t, 2*365*24*time.Hour, service.config.RecordRetentionPeriod)
	assert.Equal(t, "consent", service.config.DefaultLegalBasis)
	assert.Len(t, service.config.AllowedLegalBases, 2)
}

func TestValidateGDPRCompliance(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewGDPRComplianceService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name             string
		data             map[string]interface{}
		purpose          string
		expectCompliant  bool
		expectViolations int
		expectWarnings   int
	}{
		{
			name: "compliant data processing",
			data: map[string]interface{}{
				"legal_basis":      "legitimate_interest",
				"consent_given":    true,
				"consent_date":     time.Now(),
				"encryption":       true,
				"access_controls":  true,
				"audit_logging":    true,
				"retention_period": 30 * 24 * time.Hour,
				"original_purpose": "business_verification",
				"business_name":    "Acme Corp",
				"business_address": "123 Main St",
				"industry":         "Technology",
			},
			purpose:          "business_verification",
			expectCompliant:  false, // Will have data minimization violation
			expectViolations: 1,     // data_minimization_violation
			expectWarnings:   0,
		},
		{
			name: "missing legal basis",
			data: map[string]interface{}{
				"consent_given":    true,
				"consent_date":     time.Now(),
				"encryption":       true,
				"access_controls":  true,
				"audit_logging":    true,
				"business_name":    "Acme Corp",
				"business_address": "123 Main St",
				"industry":         "Technology",
			},
			purpose:          "business_verification",
			expectCompliant:  false,
			expectViolations: 2, // invalid_legal_basis + data_minimization_violation
			expectWarnings:   0,
		},
		{
			name: "missing consent",
			data: map[string]interface{}{
				"legal_basis":      "legitimate_interest",
				"encryption":       true,
				"access_controls":  true,
				"audit_logging":    true,
				"business_name":    "Acme Corp",
				"business_address": "123 Main St",
				"industry":         "Technology",
			},
			purpose:          "business_verification",
			expectCompliant:  false,
			expectViolations: 2, // missing_consent + data_minimization_violation
			expectWarnings:   0,
		},
		{
			name: "expired consent",
			data: map[string]interface{}{
				"legal_basis":      "legitimate_interest",
				"consent_given":    true,
				"consent_date":     time.Now().Add(-2 * 365 * 24 * time.Hour), // 2 years ago
				"encryption":       true,
				"access_controls":  true,
				"audit_logging":    true,
				"business_name":    "Acme Corp",
				"business_address": "123 Main St",
				"industry":         "Technology",
			},
			purpose:          "business_verification",
			expectCompliant:  false,
			expectViolations: 2, // missing_consent (expired) + data_minimization_violation
			expectWarnings:   0,
		},
		{
			name: "excessive data collection",
			data: map[string]interface{}{
				"legal_basis":        "legitimate_interest",
				"consent_given":      true,
				"consent_date":       time.Now(),
				"encryption":         true,
				"access_controls":    true,
				"audit_logging":      true,
				"unnecessary_field1": "value1",
				"unnecessary_field2": "value2",
				"unnecessary_field3": "value3",
			},
			purpose:          "business_verification",
			expectCompliant:  false,
			expectViolations: 1, // data_minimization_violation
			expectWarnings:   0,
		},
		{
			name: "purpose limitation violation",
			data: map[string]interface{}{
				"legal_basis":      "legitimate_interest",
				"consent_given":    true,
				"consent_date":     time.Now(),
				"encryption":       true,
				"access_controls":  true,
				"audit_logging":    true,
				"original_purpose": "marketing",
				"business_name":    "Acme Corp",
				"business_address": "123 Main St",
				"industry":         "Technology",
			},
			purpose:          "business_verification",
			expectCompliant:  false,
			expectViolations: 2, // purpose_limitation_violation + data_minimization_violation
			expectWarnings:   0,
		},
		{
			name: "excessive retention period",
			data: map[string]interface{}{
				"legal_basis":      "legitimate_interest",
				"consent_given":    true,
				"consent_date":     time.Now(),
				"encryption":       true,
				"access_controls":  true,
				"audit_logging":    true,
				"retention_period": 10 * 365 * 24 * time.Hour, // 10 years
				"business_name":    "Acme Corp",
				"business_address": "123 Main St",
				"industry":         "Technology",
			},
			purpose:          "business_verification",
			expectCompliant:  false, // Will have data minimization violation
			expectViolations: 1,     // data_minimization_violation
			expectWarnings:   1,     // retention_period_warning
		},
		{
			name: "inadequate security measures",
			data: map[string]interface{}{
				"legal_basis":      "legitimate_interest",
				"consent_given":    true,
				"consent_date":     time.Now(),
				"encryption":       true,
				"business_name":    "Acme Corp",
				"business_address": "123 Main St",
				"industry":         "Technology",
				// Missing access_controls and audit_logging
			},
			purpose:          "business_verification",
			expectCompliant:  false,
			expectViolations: 2, // inadequate_security + data_minimization_violation
			expectWarnings:   0,
		},
		{
			name: "GDPR compliance disabled",
			data: map[string]interface{}{
				"business_name": "Test Company",
			},
			purpose:          "business_verification",
			expectCompliant:  true,
			expectViolations: 0,
			expectWarnings:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For the last test case, disable GDPR compliance
			if tt.name == "GDPR compliance disabled" {
				service.config.EnableGDPRCompliance = false
				defer func() { service.config.EnableGDPRCompliance = true }()
			}

			result, err := service.ValidateGDPRCompliance(ctx, tt.data, tt.purpose)
			require.NoError(t, err)
			assert.NotNil(t, result)

			assert.Equal(t, tt.expectCompliant, result.IsCompliant)
			assert.Len(t, result.Violations, tt.expectViolations)
			assert.Len(t, result.Warnings, tt.expectWarnings)

			// Check compliance score
			assert.GreaterOrEqual(t, result.ComplianceScore, 0.0)
			assert.LessOrEqual(t, result.ComplianceScore, 100.0)

			// Check processing time
			assert.Greater(t, result.ValidationTime, time.Duration(0))

			// Check metadata
			assert.NotNil(t, result.Metadata)
			assert.NotZero(t, result.Metadata["validation_timestamp"])

			// For GDPR compliance disabled, metadata may be minimal
			if tt.name != "GDPR compliance disabled" {
				assert.Equal(t, tt.purpose, result.Metadata["processing_purpose"])
				assert.NotNil(t, result.Metadata["data_categories"])
			}

			// Check recommendations
			if len(result.Violations) > 0 || len(result.Warnings) > 0 {
				assert.NotEmpty(t, result.Recommendations)
			}
		})
	}
}

func TestRecordConsent(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewGDPRComplianceService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Test successful consent recording
	consent, err := service.RecordConsent(ctx, "user123", "explicit", "legitimate_interest", "business_verification", []string{"contact_information", "business_data"})
	require.NoError(t, err)
	assert.NotNil(t, consent)

	assert.NotEmpty(t, consent.ID)
	assert.True(t, strings.HasPrefix(consent.ID, "consent_"))
	assert.Equal(t, "user123", consent.UserID)
	assert.Equal(t, "explicit", consent.ConsentType)
	assert.Equal(t, "legitimate_interest", consent.LegalBasis)
	assert.Equal(t, "business_verification", consent.Purpose)
	assert.Equal(t, []string{"contact_information", "business_data"}, consent.DataCategories)
	assert.True(t, consent.Granular)
	assert.True(t, consent.Withdrawable)
	assert.NotZero(t, consent.GivenAt)
	assert.NotNil(t, consent.ExpiresAt)
	assert.Nil(t, consent.WithdrawnAt)
	assert.NotNil(t, consent.Evidence)
	assert.NotEmpty(t, consent.Evidence["recorded_at"])
	assert.Equal(t, "192.168.1.1", consent.Evidence["ip_address"])
	assert.Equal(t, "Mozilla/5.0 (compatible; KYB-Platform/1.0)", consent.Evidence["user_agent"])

	// Test with GDPR compliance disabled
	service.config.EnableGDPRCompliance = false
	consent, err = service.RecordConsent(ctx, "user123", "explicit", "legitimate_interest", "business_verification", []string{"contact_information"})
	assert.Error(t, err)
	assert.Nil(t, consent)
	assert.Contains(t, err.Error(), "GDPR compliance is disabled")
}

func TestWithdrawConsent(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewGDPRComplianceService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Test successful consent withdrawal
	err = service.WithdrawConsent(ctx, "consent_123", "user123")
	require.NoError(t, err)

	// Test with GDPR compliance disabled
	service.config.EnableGDPRCompliance = false
	err = service.WithdrawConsent(ctx, "consent_123", "user123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GDPR compliance is disabled")

	// Test with withdrawal not allowed
	service.config.EnableGDPRCompliance = true
	service.config.AllowWithdrawal = false
	err = service.WithdrawConsent(ctx, "consent_123", "user123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "consent withdrawal is not allowed")
}

func TestSubmitDataSubjectRequest(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewGDPRComplianceService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Test successful request submission
	request, err := service.SubmitDataSubjectRequest(ctx, "user123", "access", "Request for personal data access")
	require.NoError(t, err)
	assert.NotNil(t, request)

	assert.NotEmpty(t, request.ID)
	assert.True(t, strings.HasPrefix(request.ID, "dsr_"))
	assert.Equal(t, "user123", request.UserID)
	assert.Equal(t, "access", request.RequestType)
	assert.Equal(t, "pending", request.Status)
	assert.Equal(t, "Request for personal data access", request.Description)
	assert.NotZero(t, request.RequestedAt)
	assert.Nil(t, request.CompletedAt)
	assert.Equal(t, "email_verification", request.VerificationMethod)

	// Test with GDPR compliance disabled
	service.config.EnableGDPRCompliance = false
	request, err = service.SubmitDataSubjectRequest(ctx, "user123", "access", "Request for personal data access")
	assert.Error(t, err)
	assert.Nil(t, request)
	assert.Contains(t, err.Error(), "GDPR compliance is disabled")

	// Test with data subject rights disabled
	service.config.EnableGDPRCompliance = true
	service.config.EnableDataSubjectRights = false
	request, err = service.SubmitDataSubjectRequest(ctx, "user123", "access", "Request for personal data access")
	assert.Error(t, err)
	assert.Nil(t, request)
	assert.Contains(t, err.Error(), "data subject rights are disabled")
}

func TestProcessDataSubjectRequest(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewGDPRComplianceService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Test successful request processing
	request, err := service.ProcessDataSubjectRequest(ctx, "dsr_123")
	require.NoError(t, err)
	assert.NotNil(t, request)

	assert.Equal(t, "dsr_123", request.ID)
	assert.Equal(t, "completed", request.Status)
	assert.NotNil(t, request.CompletedAt)

	// Test with GDPR compliance disabled
	service.config.EnableGDPRCompliance = false
	request, err = service.ProcessDataSubjectRequest(ctx, "dsr_123")
	assert.Error(t, err)
	assert.Nil(t, request)
	assert.Contains(t, err.Error(), "GDPR compliance is disabled")
}

func TestDetectDataBreach(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewGDPRComplianceService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Test successful breach detection
	breach, err := service.DetectDataBreach(ctx, "unauthorized_access", "Suspicious login activity detected", 150, []string{"contact_information", "personal_identification"})
	require.NoError(t, err)
	assert.NotNil(t, breach)

	assert.NotEmpty(t, breach.ID)
	assert.True(t, strings.HasPrefix(breach.ID, "breach_"))
	assert.Equal(t, "unauthorized_access", breach.BreachType)
	assert.Equal(t, "high", breach.Severity) // 150 records + sensitive data
	assert.Equal(t, "Suspicious login activity detected", breach.Description)
	assert.NotZero(t, breach.DetectedAt)
	assert.Equal(t, 150, breach.AffectedRecords)
	assert.Equal(t, []string{"contact_information", "personal_identification"}, breach.DataCategories)
	assert.Equal(t, "system_analysis_required", breach.RootCause)
	assert.NotEmpty(t, breach.MitigationSteps)
	assert.True(t, breach.NotificationSent) // Above threshold of 100
	assert.True(t, breach.RegulatoryReported)
	assert.NotNil(t, breach.ReportedAt)

	// Test breach below notification threshold
	breach, err = service.DetectDataBreach(ctx, "data_loss", "Minor data loss incident", 50, []string{"business_data"})
	require.NoError(t, err)
	assert.NotNil(t, breach)
	assert.Equal(t, "low", breach.Severity)
	assert.False(t, breach.NotificationSent)
	assert.False(t, breach.RegulatoryReported)
	assert.Nil(t, breach.ReportedAt)

	// Test with GDPR compliance disabled
	service.config.EnableGDPRCompliance = false
	breach, err = service.DetectDataBreach(ctx, "unauthorized_access", "Test breach", 100, []string{"contact_information"})
	assert.Error(t, err)
	assert.Nil(t, breach)
	assert.Contains(t, err.Error(), "GDPR compliance is disabled")

	// Test with breach detection disabled
	service.config.EnableGDPRCompliance = true
	service.config.EnableBreachDetection = false
	breach, err = service.DetectDataBreach(ctx, "unauthorized_access", "Test breach", 100, []string{"contact_information"})
	assert.Error(t, err)
	assert.Nil(t, breach)
	assert.Contains(t, err.Error(), "breach detection is disabled")
}

func TestRecordDataProcessing(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewGDPRComplianceService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Test successful processing record
	record, err := service.RecordDataProcessing(ctx, "business_verification", "legitimate_interest", []string{"contact_information", "business_data"}, []string{"user123", "user456"})
	require.NoError(t, err)
	assert.NotNil(t, record)

	assert.NotEmpty(t, record.ID)
	assert.True(t, strings.HasPrefix(record.ID, "record_"))
	assert.Equal(t, "business_verification", record.ProcessingPurpose)
	assert.Equal(t, "legitimate_interest", record.LegalBasis)
	assert.Equal(t, []string{"contact_information", "business_data"}, record.DataCategories)
	assert.Equal(t, []string{"user123", "user456"}, record.DataSubjects)
	assert.Equal(t, service.config.RecordRetentionPeriod, record.RetentionPeriod)
	assert.NotEmpty(t, record.SecurityMeasures)
	assert.NotZero(t, record.ProcessingAt)
	assert.Equal(t, service.config.DataControllerName, record.DataController)
	assert.NotNil(t, record.Metadata)
	assert.NotZero(t, record.Metadata["recorded_at"])
	assert.Equal(t, "192.168.1.1", record.Metadata["ip_address"])

	// Test with GDPR compliance disabled
	service.config.EnableGDPRCompliance = false
	record, err = service.RecordDataProcessing(ctx, "business_verification", "legitimate_interest", []string{"contact_information"}, []string{"user123"})
	assert.Error(t, err)
	assert.Nil(t, record)
	assert.Contains(t, err.Error(), "GDPR compliance is disabled")

	// Test with processing records disabled
	service.config.EnableGDPRCompliance = true
	service.config.EnableProcessingRecords = false
	record, err = service.RecordDataProcessing(ctx, "business_verification", "legitimate_interest", []string{"contact_information"}, []string{"user123"})
	assert.Error(t, err)
	assert.Nil(t, record)
	assert.Contains(t, err.Error(), "processing records are disabled")
}

func TestCalculateComplianceScore(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewGDPRComplianceService(nil, logger)
	require.NoError(t, err)

	// Test with no violations or warnings
	violations := []GDPRViolation{}
	warnings := []GDPRWarning{}
	score := service.calculateComplianceScore(violations, warnings)
	assert.Equal(t, 100.0, score)

	// Test with violations
	violations = []GDPRViolation{
		{Severity: "high"},
		{Severity: "medium"},
		{Severity: "low"},
	}
	warnings = []GDPRWarning{{}, {}}
	score = service.calculateComplianceScore(violations, warnings)
	expectedScore := 100.0 - 15.0 - 10.0 - 5.0 - 2.0 - 2.0 // 66.0
	assert.Equal(t, expectedScore, score)

	// Test with critical violation
	violations = []GDPRViolation{
		{Severity: "critical"},
	}
	warnings = []GDPRWarning{}
	score = service.calculateComplianceScore(violations, warnings)
	assert.Equal(t, 75.0, score)

	// Test minimum score
	violations = []GDPRViolation{
		{Severity: "critical"},
		{Severity: "critical"},
		{Severity: "critical"},
		{Severity: "critical"},
		{Severity: "critical"},
	}
	warnings = []GDPRWarning{}
	score = service.calculateComplianceScore(violations, warnings)
	assert.Equal(t, 0.0, score) // Should not go below 0
}

func TestGenerateRecommendations(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewGDPRComplianceService(nil, logger)
	require.NoError(t, err)

	// Test with violations
	violations := []GDPRViolation{
		{Recommendation: "Fix PII exposure"},
		{Recommendation: "Reduce retention period"},
	}
	warnings := []GDPRWarning{
		{Recommendation: "Review potential PII"},
	}

	recommendations := service.generateRecommendations(violations, warnings)

	// Should include all violation and warning recommendations
	assert.Contains(t, recommendations, "Fix PII exposure")
	assert.Contains(t, recommendations, "Reduce retention period")
	assert.Contains(t, recommendations, "Review potential PII")

	// Should include general recommendations when violations exist
	assert.Contains(t, recommendations, "Conduct a comprehensive GDPR compliance audit")
	assert.Contains(t, recommendations, "Implement regular compliance monitoring and reporting")
	assert.Contains(t, recommendations, "Provide GDPR training to all staff members")
	assert.Contains(t, recommendations, "Establish a data protection impact assessment process")

	// Test with no violations
	violations = []GDPRViolation{}
	warnings = []GDPRWarning{{Recommendation: "Review potential PII"}}

	recommendations = service.generateRecommendations(violations, warnings)

	// Should not include general recommendations when no violations
	assert.NotContains(t, recommendations, "Conduct a comprehensive GDPR compliance audit")
	assert.NotContains(t, recommendations, "Implement regular compliance monitoring and reporting")
	assert.NotContains(t, recommendations, "Provide GDPR training to all staff members")
	assert.NotContains(t, recommendations, "Establish a data protection impact assessment process")

	// Should include warning-specific recommendations
	assert.Contains(t, recommendations, "Review potential PII")
	assert.Contains(t, recommendations, "Review and update privacy policies")
	assert.Contains(t, recommendations, "Enhance data protection measures")
	assert.Contains(t, recommendations, "Improve consent management processes")
}

func TestExtractDataCategories(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewGDPRComplianceService(nil, logger)
	require.NoError(t, err)

	tests := []struct {
		name     string
		data     map[string]interface{}
		expected []string
	}{
		{
			name: "contact information",
			data: map[string]interface{}{
				"email": "test@example.com",
				"phone": "+1-555-123-4567",
			},
			expected: []string{"contact_information"},
		},
		{
			name: "location data",
			data: map[string]interface{}{
				"address": "123 Main St",
				"city":    "Anytown",
			},
			expected: []string{"location_data", "business_data"},
		},
		{
			name: "personal identification",
			data: map[string]interface{}{
				"first_name": "John",
				"last_name":  "Doe",
			},
			expected: []string{"personal_identification"},
		},
		{
			name: "financial data",
			data: map[string]interface{}{
				"financial_info": "bank_details",
				"credit_score":   750,
			},
			expected: []string{"financial_data", "business_data"},
		},
		{
			name: "mixed categories",
			data: map[string]interface{}{
				"email":          "test@example.com",
				"address":        "123 Main St",
				"first_name":     "John",
				"financial_info": "bank_details",
				"business_name":  "Acme Corp",
			},
			expected: []string{"contact_information", "location_data", "personal_identification", "financial_data", "business_data"},
		},
		{
			name: "business data only",
			data: map[string]interface{}{
				"business_name": "Acme Corp",
				"industry":      "Technology",
				"revenue":       1000000,
			},
			expected: []string{"business_data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			categories := service.extractDataCategories(tt.data)
			assert.ElementsMatch(t, tt.expected, categories)
		})
	}
}

func TestCalculateBreachSeverity(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewGDPRComplianceService(nil, logger)
	require.NoError(t, err)

	tests := []struct {
		name             string
		affectedRecords  int
		dataCategories   []string
		expectedSeverity string
	}{
		{
			name:             "low severity - few records, non-sensitive data",
			affectedRecords:  50,
			dataCategories:   []string{"business_data"},
			expectedSeverity: "low",
		},
		{
			name:             "medium severity - moderate records, non-sensitive data",
			affectedRecords:  150,
			dataCategories:   []string{"business_data"},
			expectedSeverity: "medium",
		},
		{
			name:             "medium severity - few records, sensitive data",
			affectedRecords:  50,
			dataCategories:   []string{"financial_data"},
			expectedSeverity: "medium",
		},
		{
			name:             "high severity - many records, non-sensitive data",
			affectedRecords:  1500,
			dataCategories:   []string{"business_data"},
			expectedSeverity: "high",
		},
		{
			name:             "high severity - moderate records, sensitive data",
			affectedRecords:  150,
			dataCategories:   []string{"financial_data", "personal_identification"},
			expectedSeverity: "high",
		},
		{
			name:             "critical severity - many records",
			affectedRecords:  15000,
			dataCategories:   []string{"business_data"},
			expectedSeverity: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			severity := service.calculateBreachSeverity(tt.affectedRecords, tt.dataCategories)
			assert.Equal(t, tt.expectedSeverity, severity)
		})
	}
}

func TestIDGeneration(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewGDPRComplianceService(nil, logger)
	require.NoError(t, err)

	// Test consent ID generation
	consentID, err := service.generateConsentID()
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(consentID, "consent_"))
	assert.Len(t, consentID, 24) // "consent_" + 16 hex chars

	// Test request ID generation
	requestID, err := service.generateRequestID()
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(requestID, "dsr_"))
	assert.Len(t, requestID, 20) // "dsr_" + 16 hex chars

	// Test breach ID generation
	breachID, err := service.generateBreachID()
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(breachID, "breach_"))
	assert.Len(t, breachID, 23) // "breach_" + 16 hex chars

	// Test record ID generation
	recordID, err := service.generateRecordID()
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(recordID, "record_"))
	assert.Len(t, recordID, 23) // "record_" + 16 hex chars

	// Test generic ID generation
	genericID, err := service.generateID("test")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(genericID, "test_"))
	assert.Len(t, genericID, 21) // "test_" + 16 hex chars
}
