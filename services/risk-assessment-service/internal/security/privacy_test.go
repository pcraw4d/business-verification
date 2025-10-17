package security

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPrivacyManager(t *testing.T) {
	tests := []struct {
		name   string
		config *PrivacyConfig
	}{
		{
			name:   "default config",
			config: nil,
		},
		{
			name: "custom config",
			config: &PrivacyConfig{
				DataRetentionPeriod:     24 * time.Hour,
				AnonymizationEnabled:    true,
				PseudonymizationEnabled: true,
				ConsentRequired:         true,
				RightToErasureEnabled:   true,
				DataPortabilityEnabled:  true,
				AuditLogRetention:       7 * 24 * time.Hour,
				EncryptionAtRest:        true,
				EncryptionInTransit:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			pm := NewPrivacyManager(tt.config, mockLogger)
			assert.NotNil(t, pm)
			assert.NotNil(t, pm.config)
		})
	}
}

func TestPrivacyManager_RegisterDataSubject(t *testing.T) {
	mockLogger := &MockLogger{}
	pm := NewPrivacyManager(nil, mockLogger)

	tests := []struct {
		name        string
		subject     *DataSubject
		expectError bool
	}{
		{
			name: "valid data subject",
			subject: &DataSubject{
				Email:        "test@example.com",
				ConsentGiven: true,
				DataTypes:    []string{"personal_data", "contact_info"},
			},
			expectError: false,
		},
		{
			name: "data subject with ID",
			subject: &DataSubject{
				ID:           "ds_12345678",
				Email:        "test@example.com",
				ConsentGiven: true,
			},
			expectError: false,
		},
		{
			name:        "nil data subject",
			subject:     nil,
			expectError: true,
		},
		{
			name: "empty email",
			subject: &DataSubject{
				Email: "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := pm.RegisterDataSubject(ctx, tt.subject)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, tt.subject.ID)
				assert.False(t, tt.subject.CreatedAt.IsZero())
				assert.False(t, tt.subject.UpdatedAt.IsZero())
			}
		})
	}
}

func TestPrivacyManager_RecordConsent(t *testing.T) {
	mockLogger := &MockLogger{}
	pm := NewPrivacyManager(nil, mockLogger)

	tests := []struct {
		name          string
		dataSubjectID string
		purpose       string
		consentGiven  bool
		expectError   bool
	}{
		{
			name:          "valid consent",
			dataSubjectID: "ds_12345678",
			purpose:       "marketing",
			consentGiven:  true,
			expectError:   false,
		},
		{
			name:          "consent withdrawal",
			dataSubjectID: "ds_12345678",
			purpose:       "marketing",
			consentGiven:  false,
			expectError:   false,
		},
		{
			name:          "empty data subject ID",
			dataSubjectID: "",
			purpose:       "marketing",
			consentGiven:  true,
			expectError:   true,
		},
		{
			name:          "empty purpose",
			dataSubjectID: "ds_12345678",
			purpose:       "",
			consentGiven:  true,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			consent, err := pm.RecordConsent(ctx, tt.dataSubjectID, tt.purpose, tt.consentGiven)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, consent)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, consent)
				assert.Equal(t, tt.dataSubjectID, consent.DataSubjectID)
				assert.Equal(t, tt.purpose, consent.Purpose)
				assert.Equal(t, tt.consentGiven, consent.ConsentGiven)
				assert.False(t, consent.ConsentDate.IsZero())
				assert.Equal(t, 1, consent.Version)
			}
		})
	}
}

func TestPrivacyManager_WithdrawConsent(t *testing.T) {
	mockLogger := &MockLogger{}
	pm := NewPrivacyManager(nil, mockLogger)

	tests := []struct {
		name          string
		dataSubjectID string
		purpose       string
		expectError   bool
	}{
		{
			name:          "valid withdrawal",
			dataSubjectID: "ds_12345678",
			purpose:       "marketing",
			expectError:   false,
		},
		{
			name:          "empty data subject ID",
			dataSubjectID: "",
			purpose:       "marketing",
			expectError:   true,
		},
		{
			name:          "empty purpose",
			dataSubjectID: "ds_12345678",
			purpose:       "",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := pm.WithdrawConsent(ctx, tt.dataSubjectID, tt.purpose)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPrivacyManager_RequestDataErasure(t *testing.T) {
	tests := []struct {
		name          string
		config        *PrivacyConfig
		dataSubjectID string
		reason        string
		expectError   bool
	}{
		{
			name: "erasure enabled",
			config: &PrivacyConfig{
				RightToErasureEnabled: true,
			},
			dataSubjectID: "ds_12345678",
			reason:        "withdrawal of consent",
			expectError:   false,
		},
		{
			name: "erasure disabled",
			config: &PrivacyConfig{
				RightToErasureEnabled: false,
			},
			dataSubjectID: "ds_12345678",
			reason:        "withdrawal of consent",
			expectError:   true,
		},
		{
			name:          "empty data subject ID",
			config:        &PrivacyConfig{RightToErasureEnabled: true},
			dataSubjectID: "",
			reason:        "withdrawal of consent",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			pm := NewPrivacyManager(tt.config, mockLogger)

			ctx := context.Background()
			request, err := pm.RequestDataErasure(ctx, tt.dataSubjectID, tt.reason)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, request)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, request)
				assert.Equal(t, tt.dataSubjectID, request.DataSubjectID)
				assert.Equal(t, tt.reason, request.Reason)
				assert.Equal(t, "PENDING", request.Status)
				assert.False(t, request.RequestDate.IsZero())
			}
		})
	}
}

func TestPrivacyManager_ProcessDataErasure(t *testing.T) {
	mockLogger := &MockLogger{}
	pm := NewPrivacyManager(nil, mockLogger)

	tests := []struct {
		name        string
		requestID   string
		processedBy string
		expectError bool
	}{
		{
			name:        "valid processing",
			requestID:   "erasure_12345678",
			processedBy: "admin_user",
			expectError: false,
		},
		{
			name:        "empty request ID",
			requestID:   "",
			processedBy: "admin_user",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := pm.ProcessDataErasure(ctx, tt.requestID, tt.processedBy)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPrivacyManager_RequestDataPortability(t *testing.T) {
	tests := []struct {
		name          string
		config        *PrivacyConfig
		dataSubjectID string
		dataTypes     []string
		format        string
		expectError   bool
	}{
		{
			name: "portability enabled",
			config: &PrivacyConfig{
				DataPortabilityEnabled: true,
			},
			dataSubjectID: "ds_12345678",
			dataTypes:     []string{"personal_data", "contact_info"},
			format:        "JSON",
			expectError:   false,
		},
		{
			name: "portability disabled",
			config: &PrivacyConfig{
				DataPortabilityEnabled: false,
			},
			dataSubjectID: "ds_12345678",
			dataTypes:     []string{"personal_data"},
			format:        "JSON",
			expectError:   true,
		},
		{
			name:          "empty data subject ID",
			config:        &PrivacyConfig{DataPortabilityEnabled: true},
			dataSubjectID: "",
			dataTypes:     []string{"personal_data"},
			format:        "JSON",
			expectError:   true,
		},
		{
			name:          "empty data types",
			config:        &PrivacyConfig{DataPortabilityEnabled: true},
			dataSubjectID: "ds_12345678",
			dataTypes:     []string{},
			format:        "JSON",
			expectError:   true,
		},
		{
			name:          "default format",
			config:        &PrivacyConfig{DataPortabilityEnabled: true},
			dataSubjectID: "ds_12345678",
			dataTypes:     []string{"personal_data"},
			format:        "",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			pm := NewPrivacyManager(tt.config, mockLogger)

			ctx := context.Background()
			request, err := pm.RequestDataPortability(ctx, tt.dataSubjectID, tt.dataTypes, tt.format)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, request)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, request)
				assert.Equal(t, tt.dataSubjectID, request.DataSubjectID)
				assert.Equal(t, tt.dataTypes, request.DataTypes)
				if tt.format == "" {
					assert.Equal(t, "JSON", request.Format)
				} else {
					assert.Equal(t, tt.format, request.Format)
				}
				assert.Equal(t, "PENDING", request.Status)
				assert.False(t, request.RequestDate.IsZero())
			}
		})
	}
}

func TestPrivacyManager_ProcessDataPortability(t *testing.T) {
	mockLogger := &MockLogger{}
	pm := NewPrivacyManager(nil, mockLogger)

	tests := []struct {
		name        string
		requestID   string
		processedBy string
		expectError bool
	}{
		{
			name:        "valid processing",
			requestID:   "portability_12345678",
			processedBy: "admin_user",
			expectError: false,
		},
		{
			name:        "empty request ID",
			requestID:   "",
			processedBy: "admin_user",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := pm.ProcessDataPortability(ctx, tt.requestID, tt.processedBy)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPrivacyManager_AnonymizeData(t *testing.T) {
	tests := []struct {
		name        string
		config      *PrivacyConfig
		data        map[string]interface{}
		expectError bool
	}{
		{
			name: "anonymization enabled",
			config: &PrivacyConfig{
				AnonymizationEnabled: true,
			},
			data: map[string]interface{}{
				"name":       "John Doe",
				"email":      "john@example.com",
				"age":        30,
				"company":    "Acme Corp",
				"department": "Engineering",
			},
			expectError: false,
		},
		{
			name: "anonymization disabled",
			config: &PrivacyConfig{
				AnonymizationEnabled: false,
			},
			data: map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
				"age":   30,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			pm := NewPrivacyManager(tt.config, mockLogger)

			ctx := context.Background()
			anonymized, err := pm.AnonymizeData(ctx, tt.data)

			require.NoError(t, err)
			assert.NotNil(t, anonymized)

			if tt.config.AnonymizationEnabled {
				// Check that personal data is anonymized
				assert.Equal(t, "***ANONYMIZED***", anonymized["name"])
				assert.Equal(t, "***ANONYMIZED***", anonymized["email"])
				assert.Equal(t, 0, anonymized["age"])

				// Check that non-personal data is preserved
				assert.Equal(t, "Acme Corp", anonymized["company"])
				assert.Equal(t, "Engineering", anonymized["department"])
			} else {
				// Check that data is unchanged
				assert.Equal(t, tt.data, anonymized)
			}
		})
	}
}

func TestPrivacyManager_PseudonymizeData(t *testing.T) {
	tests := []struct {
		name        string
		config      *PrivacyConfig
		data        map[string]interface{}
		expectError bool
	}{
		{
			name: "pseudonymization enabled",
			config: &PrivacyConfig{
				PseudonymizationEnabled: true,
			},
			data: map[string]interface{}{
				"name":       "John Doe",
				"email":      "john@example.com",
				"age":        30,
				"company":    "Acme Corp",
				"department": "Engineering",
			},
			expectError: false,
		},
		{
			name: "pseudonymization disabled",
			config: &PrivacyConfig{
				PseudonymizationEnabled: false,
			},
			data: map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
				"age":   30,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			pm := NewPrivacyManager(tt.config, mockLogger)

			ctx := context.Background()
			pseudonymized, err := pm.PseudonymizeData(ctx, tt.data)

			require.NoError(t, err)
			assert.NotNil(t, pseudonymized)

			if tt.config.PseudonymizationEnabled {
				// Check that personal data is pseudonymized
				assert.Contains(t, pseudonymized["name"].(string), "pseudo_")
				assert.Contains(t, pseudonymized["email"].(string), "pseudo_")
				assert.Equal(t, 30, pseudonymized["age"]) // Non-personal data unchanged

				// Check that non-personal data is preserved
				assert.Equal(t, "Acme Corp", pseudonymized["company"])
				assert.Equal(t, "Engineering", pseudonymized["department"])
			} else {
				// Check that data is unchanged
				assert.Equal(t, tt.data, pseudonymized)
			}
		})
	}
}

func TestPrivacyManager_CheckDataRetention(t *testing.T) {
	mockLogger := &MockLogger{}
	pm := NewPrivacyManager(nil, mockLogger)

	tests := []struct {
		name         string
		dataType     string
		createdAt    time.Time
		expectRetain bool
		expectError  bool
	}{
		{
			name:         "data within retention period",
			dataType:     "personal_data",
			createdAt:    time.Now().Add(-24 * time.Hour), // 1 day ago
			expectRetain: true,
			expectError:  false,
		},
		{
			name:         "data beyond retention period",
			dataType:     "personal_data",
			createdAt:    time.Now().Add(-8 * 24 * time.Hour), // 8 days ago
			expectRetain: false,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			shouldRetain, err := pm.CheckDataRetention(ctx, tt.dataType, tt.createdAt)

			require.NoError(t, err)
			assert.Equal(t, tt.expectRetain, shouldRetain)
		})
	}
}

func TestPrivacyManager_GetDataSubjectRights(t *testing.T) {
	tests := []struct {
		name           string
		config         *PrivacyConfig
		dataSubjectID  string
		expectedRights []string
	}{
		{
			name: "all rights enabled",
			config: &PrivacyConfig{
				RightToErasureEnabled:  true,
				DataPortabilityEnabled: true,
			},
			dataSubjectID: "ds_12345678",
			expectedRights: []string{
				"right_to_access",
				"right_to_rectification",
				"right_to_erasure",
				"right_to_restrict_processing",
				"right_to_data_portability",
				"right_to_object",
				"rights_related_to_automated_decision_making",
			},
		},
		{
			name: "some rights disabled",
			config: &PrivacyConfig{
				RightToErasureEnabled:  false,
				DataPortabilityEnabled: false,
			},
			dataSubjectID: "ds_12345678",
			expectedRights: []string{
				"right_to_access",
				"right_to_rectification",
				"right_to_restrict_processing",
				"right_to_object",
				"rights_related_to_automated_decision_making",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			pm := NewPrivacyManager(tt.config, mockLogger)

			ctx := context.Background()
			rights, err := pm.GetDataSubjectRights(ctx, tt.dataSubjectID)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedRights, rights)
		})
	}
}

func TestPrivacyManager_ValidateDataProcessing(t *testing.T) {
	mockLogger := &MockLogger{}
	pm := NewPrivacyManager(nil, mockLogger)

	tests := []struct {
		name        string
		activity    *DataProcessingActivity
		expectError bool
	}{
		{
			name: "valid activity",
			activity: &DataProcessingActivity{
				Purpose:          "risk assessment",
				LegalBasis:       "legitimate_interest",
				DataCategories:   []string{"personal_data", "financial_data"},
				RetentionPeriod:  30 * 24 * time.Hour,
				SecurityMeasures: []string{"encryption", "access_control"},
			},
			expectError: false,
		},
		{
			name:        "nil activity",
			activity:    nil,
			expectError: true,
		},
		{
			name: "empty purpose",
			activity: &DataProcessingActivity{
				Purpose:         "",
				LegalBasis:      "legitimate_interest",
				DataCategories:  []string{"personal_data"},
				RetentionPeriod: 30 * 24 * time.Hour,
			},
			expectError: true,
		},
		{
			name: "empty legal basis",
			activity: &DataProcessingActivity{
				Purpose:         "risk assessment",
				LegalBasis:      "",
				DataCategories:  []string{"personal_data"},
				RetentionPeriod: 30 * 24 * time.Hour,
			},
			expectError: true,
		},
		{
			name: "empty data categories",
			activity: &DataProcessingActivity{
				Purpose:         "risk assessment",
				LegalBasis:      "legitimate_interest",
				DataCategories:  []string{},
				RetentionPeriod: 30 * 24 * time.Hour,
			},
			expectError: true,
		},
		{
			name: "zero retention period",
			activity: &DataProcessingActivity{
				Purpose:         "risk assessment",
				LegalBasis:      "legitimate_interest",
				DataCategories:  []string{"personal_data"},
				RetentionPeriod: 0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := pm.ValidateDataProcessing(ctx, tt.activity)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPrivacyManager_GeneratePrivacyReport(t *testing.T) {
	mockLogger := &MockLogger{}
	pm := NewPrivacyManager(nil, mockLogger)

	ctx := context.Background()
	report, err := pm.GeneratePrivacyReport(ctx)

	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Contains(t, report, "generated_at")
	assert.Contains(t, report, "config")
	assert.Contains(t, report, "compliance_status")
	assert.Contains(t, report, "recommendations")

	// Check compliance status
	assert.Equal(t, "COMPLIANT", report["compliance_status"])

	// Check recommendations
	recommendations, ok := report["recommendations"].([]string)
	assert.True(t, ok)
	assert.NotEmpty(t, recommendations)
}

func TestIsPersonalData(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		expected  bool
	}{
		{"email field", "email", true},
		{"name field", "name", true},
		{"address field", "address", true},
		{"phone field", "phone", true},
		{"ssn field", "ssn", true},
		{"passport field", "passport", true},
		{"id_number field", "id_number", true},
		{"date_of_birth field", "date_of_birth", true},
		{"gender field", "gender", true},
		{"nationality field", "nationality", true},
		{"ip_address field", "ip_address", true},
		{"user_agent field", "user_agent", true},
		{"location field", "location", true},
		{"biometric_data field", "biometric_data", true},
		{"health_data field", "health_data", true},
		{"financial_data field", "financial_data", true},
		{"company field", "company", false},
		{"department field", "department", false},
		{"role field", "role", false},
		{"status field", "status", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPersonalData(tt.fieldName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnonymizeValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
	}{
		{"string value", "John Doe", "***ANONYMIZED***"},
		{"empty string", "", ""},
		{"int value", 30, 0},
		{"int64 value", int64(30), 0},
		{"float64 value", 30.5, 0},
		{"bool value", true, false},
		{"nil value", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := anonymizeValue(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPseudonymizeValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
	}{
		{"string value", "John Doe", "pseudo_"},
		{"empty string", "", ""},
		{"int value", 30, 30},
		{"bool value", true, true},
		{"nil value", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pseudonymizeValue(tt.value)
			if tt.name == "string value" {
				assert.Contains(t, result.(string), "pseudo_")
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestRemoveString(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected []string
	}{
		{
			name:     "remove existing item",
			slice:    []string{"a", "b", "c"},
			item:     "b",
			expected: []string{"a", "c"},
		},
		{
			name:     "remove non-existing item",
			slice:    []string{"a", "b", "c"},
			item:     "d",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "remove from empty slice",
			slice:    []string{},
			item:     "a",
			expected: []string{},
		},
		{
			name:     "remove from single item slice",
			slice:    []string{"a"},
			item:     "a",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeString(tt.slice, tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}
