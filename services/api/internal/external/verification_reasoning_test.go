package external

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVerificationReasoningGenerator(t *testing.T) {
	tests := []struct {
		name   string
		config *VerificationReasoningConfig
		want   *VerificationReasoningConfig
	}{
		{
			name:   "nil config uses defaults",
			config: nil,
			want: &VerificationReasoningConfig{
				EnableDetailedExplanations: true,
				EnableRiskAnalysis:         true,
				EnableRecommendations:      true,
				EnableAuditTrail:           true,
				MinConfidenceThreshold:     0.6,
				MaxRiskProbability:         0.8,
				Language:                   "en",
			},
		},
		{
			name: "custom config preserved",
			config: &VerificationReasoningConfig{
				EnableDetailedExplanations: false,
				EnableRiskAnalysis:         false,
				EnableRecommendations:      false,
				EnableAuditTrail:           false,
				MinConfidenceThreshold:     0.8,
				MaxRiskProbability:         0.9,
				Language:                   "es",
			},
			want: &VerificationReasoningConfig{
				EnableDetailedExplanations: false,
				EnableRiskAnalysis:         false,
				EnableRecommendations:      false,
				EnableAuditTrail:           false,
				MinConfidenceThreshold:     0.8,
				MaxRiskProbability:         0.9,
				Language:                   "es",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewVerificationReasoningGenerator(tt.config)
			require.NotNil(t, generator)
			assert.Equal(t, tt.want, generator.GetConfig())
		})
	}
}

func TestVerificationReasoningGenerator_CalculateConfidenceLevel(t *testing.T) {
	generator := NewVerificationReasoningGenerator(nil)

	tests := []struct {
		name  string
		score float64
		want  string
	}{
		{"high confidence", 0.9, "high"},
		{"high confidence boundary", 0.8, "high"},
		{"medium confidence", 0.7, "medium"},
		{"medium confidence boundary", 0.6, "medium"},
		{"low confidence", 0.5, "low"},
		{"low confidence boundary", 0.4, "low"},
		{"very low confidence", 0.3, "very_low"},
		{"zero score", 0.0, "very_low"},
		{"negative score", -0.1, "very_low"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.calculateConfidenceLevel(tt.score)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestVerificationReasoningGenerator_GenerateReasoning(t *testing.T) {
	generator := NewVerificationReasoningGenerator(nil)

	tests := []struct {
		name           string
		verificationID string
		businessName   string
		websiteURL     string
		result         *VerificationResult
		comparison     *ComparisonResult
		wantErr        bool
		checkFields    func(*testing.T, *VerificationReasoning)
	}{
		{
			name:           "nil result returns error",
			verificationID: "test-123",
			businessName:   "Test Business",
			websiteURL:     "https://test.com",
			result:         nil,
			comparison:     nil,
			wantErr:        true,
		},
		{
			name:           "passed verification with comparison",
			verificationID: "test-123",
			businessName:   "Test Business",
			websiteURL:     "https://test.com",
			result: &VerificationResult{
				ID:           "test-123",
				Status:       StatusPassed,
				OverallScore: 0.85,
				FieldResults: map[string]FieldResult{
					"business_name": {
						Status:     StatusPassed,
						Score:      0.9,
						Confidence: 0.8,
						Matched:    true,
					},
				},
			},
			comparison: &ComparisonResult{
				OverallScore: 0.85,
				FieldResults: map[string]FieldComparison{
					"business_name": {
						Score:      0.9,
						Confidence: 0.8,
						Matched:    true,
					},
				},
			},
			wantErr: false,
			checkFields: func(t *testing.T, reasoning *VerificationReasoning) {
				assert.Equal(t, "PASSED", reasoning.Status)
				assert.Equal(t, 0.85, reasoning.OverallScore)
				assert.Equal(t, "high", reasoning.ConfidenceLevel)
				assert.Contains(t, reasoning.Explanation, "Verification PASSED")
				assert.Len(t, reasoning.FieldAnalysis, 1)
				assert.Len(t, reasoning.Recommendations, 1)
				assert.Equal(t, "test-123", reasoning.VerificationID)
				assert.Equal(t, "Test Business", reasoning.BusinessName)
				assert.Equal(t, "https://test.com", reasoning.WebsiteURL)
			},
		},
		{
			name:           "failed verification with risk factors",
			verificationID: "test-456",
			businessName:   "Failed Business",
			websiteURL:     "https://failed.com",
			result: &VerificationResult{
				ID:           "test-456",
				Status:       StatusFailed,
				OverallScore: 0.3,
				FieldResults: map[string]FieldResult{
					"business_name": {
						Status:     StatusFailed,
						Score:      0.2,
						Confidence: 0.3,
						Matched:    false,
					},
				},
			},
			comparison: &ComparisonResult{
				OverallScore: 0.3,
				FieldResults: map[string]FieldComparison{
					"business_name": {
						Score:      0.2,
						Confidence: 0.3,
						Matched:    false,
					},
				},
			},
			wantErr: false,
			checkFields: func(t *testing.T, reasoning *VerificationReasoning) {
				assert.Equal(t, "FAILED", reasoning.Status)
				assert.Equal(t, 0.3, reasoning.OverallScore)
				assert.Equal(t, "very_low", reasoning.ConfidenceLevel)
				assert.Contains(t, reasoning.Explanation, "Verification FAILED")
				assert.Len(t, reasoning.FieldAnalysis, 1)
				assert.Len(t, reasoning.Recommendations, 2) // Overall + field-specific
				assert.Len(t, reasoning.RiskFactors, 2)     // Overall + field-specific
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reasoning, err := generator.GenerateReasoning(
				tt.verificationID,
				tt.businessName,
				tt.websiteURL,
				tt.result,
				tt.comparison,
			)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, reasoning)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reasoning)
				if tt.checkFields != nil {
					tt.checkFields(t, reasoning)
				}
			}
		})
	}
}

func TestVerificationReasoningGenerator_GenerateOverallExplanation(t *testing.T) {
	generator := NewVerificationReasoningGenerator(nil)

	tests := []struct {
		name       string
		result     *VerificationResult
		comparison *ComparisonResult
		want       string
	}{
		{
			name: "passed verification",
			result: &VerificationResult{
				Status:       StatusPassed,
				OverallScore: 0.85,
			},
			comparison: nil,
			want:       "Verification PASSED with high confidence. Overall score: 0.85 (85%). All critical fields matched successfully with strong evidence.",
		},
		{
			name: "partial verification",
			result: &VerificationResult{
				Status:       StatusPartial,
				OverallScore: 0.65,
			},
			comparison: nil,
			want:       "Verification PARTIALLY PASSED with moderate confidence. Overall score: 0.65 (65%). Some fields matched while others require manual review.",
		},
		{
			name: "failed verification",
			result: &VerificationResult{
				Status:       StatusFailed,
				OverallScore: 0.35,
			},
			comparison: nil,
			want:       "Verification FAILED with low confidence. Overall score: 0.35 (35%). Multiple fields failed to match or had insufficient evidence.",
		},
		{
			name: "skipped verification",
			result: &VerificationResult{
				Status:       StatusSkipped,
				OverallScore: 0.0,
			},
			comparison: nil,
			want:       "Verification SKIPPED due to insufficient data or technical issues. Manual verification recommended.",
		},
		{
			name: "passed with field insights",
			result: &VerificationResult{
				Status:       StatusPassed,
				OverallScore: 0.85,
			},
			comparison: &ComparisonResult{
				FieldResults: map[string]FieldComparison{
					"business_name":   {Score: 0.9},
					"phone_numbers":   {Score: 0.7},
					"email_addresses": {Score: 0.3},
				},
			},
			want: "Verification PASSED with high confidence. Overall score: 0.85 (85%). All critical fields matched successfully with strong evidence. 1 fields matched strongly, 1 fields matched partially, 1 fields failed to match. ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			explanation := generator.generateOverallExplanation(tt.result, tt.comparison)
			assert.Equal(t, tt.want, explanation)
		})
	}
}

func TestVerificationReasoningGenerator_GenerateFieldAnalysis(t *testing.T) {
	generator := NewVerificationReasoningGenerator(nil)

	tests := []struct {
		name       string
		comparison *ComparisonResult
		want       []FieldAnalysis
	}{
		{
			name:       "nil comparison returns empty slice",
			comparison: nil,
			want:       []FieldAnalysis{},
		},
		{
			name: "single field analysis",
			comparison: &ComparisonResult{
				FieldResults: map[string]FieldComparison{
					"business_name": {
						Score:      0.9,
						Confidence: 0.8,
					},
				},
			},
			want: []FieldAnalysis{
				{
					FieldName:    "business_name",
					Score:        0.9,
					Status:       "passed",
					Confidence:   0.8,
					Weight:       0.0,
					Contribution: 0.0,
				},
			},
		},
		{
			name: "multiple fields analysis",
			comparison: &ComparisonResult{
				FieldResults: map[string]FieldComparison{
					"business_name": {
						Score:      0.9,
						Confidence: 0.8,
					},
					"phone_numbers": {
						Score:      0.5,
						Confidence: 0.6,
					},
				},
			},
			want: []FieldAnalysis{
				{
					FieldName:    "business_name",
					Score:        0.9,
					Status:       "passed",
					Confidence:   0.8,
					Weight:       0.0,
					Contribution: 0.0,
				},
				{
					FieldName:    "phone_numbers",
					Score:        0.5,
					Status:       "failed",
					Confidence:   0.6,
					Weight:       0.0,
					Contribution: 0.0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := generator.generateFieldAnalysis(tt.comparison)
			assert.Len(t, analysis, len(tt.want))

			// Create a map of expected results by field name for order-independent comparison
			expectedMap := make(map[string]FieldAnalysis)
			for _, expected := range tt.want {
				expectedMap[expected.FieldName] = expected
			}

			// Check each field in the analysis
			for _, field := range analysis {
				expected, exists := expectedMap[field.FieldName]
				assert.True(t, exists, "Field %s not found in expected results", field.FieldName)
				if exists {
					assert.Equal(t, expected.Score, field.Score)
					assert.Equal(t, expected.Status, field.Status)
					assert.Equal(t, expected.Confidence, field.Confidence)
					assert.Equal(t, expected.Weight, field.Weight)
					assert.Equal(t, expected.Contribution, field.Contribution)
					assert.NotEmpty(t, field.Explanation)
					assert.NotEmpty(t, field.Evidence)
				}
			}
		})
	}
}

func TestVerificationReasoningGenerator_GetFieldStatus(t *testing.T) {
	generator := NewVerificationReasoningGenerator(nil)

	tests := []struct {
		name  string
		score float64
		want  string
	}{
		{"passed", 0.9, "passed"},
		{"passed boundary", 0.8, "passed"},
		{"partial", 0.7, "partial"},
		{"partial boundary", 0.6, "partial"},
		{"failed", 0.5, "failed"},
		{"failed boundary", 0.0, "failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := generator.getFieldStatus(tt.score)
			assert.Equal(t, tt.want, status)
		})
	}
}

func TestVerificationReasoningGenerator_GenerateFieldExplanation(t *testing.T) {
	generator := NewVerificationReasoningGenerator(nil)

	tests := []struct {
		name      string
		fieldName string
		field     FieldComparison
		want      string
	}{
		{
			name:      "business name high score",
			fieldName: "business_name",
			field:     FieldComparison{Score: 0.9},
			want:      "Business name comparison using fuzzy string matching. Names matched with high similarity.",
		},
		{
			name:      "business name medium score",
			fieldName: "business_name",
			field:     FieldComparison{Score: 0.7},
			want:      "Business name comparison using fuzzy string matching. Names showed moderate similarity with minor differences.",
		},
		{
			name:      "business name low score",
			fieldName: "business_name",
			field:     FieldComparison{Score: 0.3},
			want:      "Business name comparison using fuzzy string matching. Names showed significant differences or no match found.",
		},
		{
			name:      "phone numbers high score",
			fieldName: "phone_numbers",
			field:     FieldComparison{Score: 0.9},
			want:      "Phone number comparison using normalized format matching. Phone numbers matched exactly or with minor formatting differences.",
		},
		{
			name:      "email addresses medium score",
			fieldName: "email_addresses",
			field:     FieldComparison{Score: 0.7},
			want:      "Email address comparison using domain and local part analysis. Email addresses showed partial match or similar patterns.",
		},
		{
			name:      "addresses low score",
			fieldName: "addresses",
			field:     FieldComparison{Score: 0.3},
			want:      "Address comparison using geographic and component matching. Addresses did not match or geographic data insufficient.",
		},
		{
			name:      "unknown field",
			fieldName: "unknown_field",
			field:     FieldComparison{Score: 0.8},
			want:      "unknown_field field comparison. Field matched successfully.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			explanation := generator.generateFieldExplanation(tt.fieldName, tt.field)
			assert.Equal(t, tt.want, explanation)
		})
	}
}

func TestVerificationReasoningGenerator_GenerateFieldEvidence(t *testing.T) {
	generator := NewVerificationReasoningGenerator(nil)

	tests := []struct {
		name  string
		field FieldComparison
		want  string
	}{
		{
			name:  "high score evidence",
			field: FieldComparison{Score: 0.9, Confidence: 0.8},
			want:  "Score: 0.90, Confidence: 0.80. Strong evidence of match.",
		},
		{
			name:  "medium score evidence",
			field: FieldComparison{Score: 0.7, Confidence: 0.6},
			want:  "Score: 0.70, Confidence: 0.60. Moderate evidence of match.",
		},
		{
			name:  "low score evidence",
			field: FieldComparison{Score: 0.3, Confidence: 0.4},
			want:  "Score: 0.30, Confidence: 0.40. Weak or no evidence of match.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evidence := generator.generateFieldEvidence(tt.field)
			assert.Equal(t, tt.want, evidence)
		})
	}
}

func TestVerificationReasoningGenerator_GenerateRecommendations(t *testing.T) {
	generator := NewVerificationReasoningGenerator(nil)

	tests := []struct {
		name       string
		result     *VerificationResult
		comparison *ComparisonResult
		wantCount  int
		checkTypes []string
	}{
		{
			name: "passed verification",
			result: &VerificationResult{
				Status:       StatusPassed,
				OverallScore: 0.85,
			},
			comparison: nil,
			wantCount:  1,
			checkTypes: []string{"verification"},
		},
		{
			name: "partial verification",
			result: &VerificationResult{
				Status:       StatusPartial,
				OverallScore: 0.65,
			},
			comparison: nil,
			wantCount:  1,
			checkTypes: []string{"manual_review"},
		},
		{
			name: "failed verification",
			result: &VerificationResult{
				Status:       StatusFailed,
				OverallScore: 0.35,
			},
			comparison: nil,
			wantCount:  1,
			checkTypes: []string{"investigation"},
		},
		{
			name: "failed verification with field issues",
			result: &VerificationResult{
				Status:       StatusFailed,
				OverallScore: 0.35,
			},
			comparison: &ComparisonResult{
				FieldResults: map[string]FieldComparison{
					"business_name": {Score: 0.3},
					"phone_numbers": {Score: 0.4},
				},
			},
			wantCount:  3, // Overall + 2 field-specific
			checkTypes: []string{"investigation", "field_verification", "field_verification"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendations := generator.generateRecommendations(tt.result, tt.comparison)
			assert.Len(t, recommendations, tt.wantCount)

			for i, expectedType := range tt.checkTypes {
				if i < len(recommendations) {
					assert.Equal(t, expectedType, recommendations[i].Type)
					assert.NotEmpty(t, recommendations[i].Description)
					assert.NotEmpty(t, recommendations[i].Action)
					assert.NotEmpty(t, recommendations[i].Reason)
					assert.NotEmpty(t, recommendations[i].Impact)
					assert.Contains(t, []string{"high", "medium", "low"}, recommendations[i].Priority)
				}
			}
		})
	}
}

func TestVerificationReasoningGenerator_GenerateRiskFactors(t *testing.T) {
	generator := NewVerificationReasoningGenerator(nil)

	tests := []struct {
		name       string
		result     *VerificationResult
		comparison *ComparisonResult
		wantCount  int
	}{
		{
			name: "high score no risks",
			result: &VerificationResult{
				Status:       StatusPassed,
				OverallScore: 0.85,
			},
			comparison: nil,
			wantCount:  0,
		},
		{
			name: "low score with overall risk",
			result: &VerificationResult{
				Status:       StatusFailed,
				OverallScore: 0.35,
			},
			comparison: nil,
			wantCount:  1,
		},
		{
			name: "low score with field risks",
			result: &VerificationResult{
				Status:       StatusFailed,
				OverallScore: 0.35,
			},
			comparison: &ComparisonResult{
				FieldResults: map[string]FieldComparison{
					"business_name": {Score: 0.2},
					"phone_numbers": {Score: 0.3},
				},
			},
			wantCount: 3, // Overall + 2 field-specific
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			riskFactors := generator.generateRiskFactors(tt.result, tt.comparison)
			assert.Len(t, riskFactors, tt.wantCount)

			for _, risk := range riskFactors {
				assert.NotEmpty(t, risk.Factor)
				assert.NotEmpty(t, risk.Description)
				assert.NotEmpty(t, risk.Impact)
				assert.NotEmpty(t, risk.Mitigation)
				assert.Contains(t, []string{"low", "medium", "high", "critical"}, risk.Severity)
				assert.GreaterOrEqual(t, risk.Probability, 0.0)
				assert.LessOrEqual(t, risk.Probability, 1.0)
			}
		})
	}
}

func TestVerificationReasoning_StructFields(t *testing.T) {
	reasoning := &VerificationReasoning{
		Status:          "PASSED",
		OverallScore:    0.85,
		ConfidenceLevel: "high",
		Explanation:     "Test explanation",
		FieldAnalysis:   []FieldAnalysis{},
		Recommendations: []Recommendation{},
		RiskFactors:     []RiskFactor{},
		GeneratedAt:     time.Now(),
		VerificationID:  "test-123",
		BusinessName:    "Test Business",
		WebsiteURL:      "https://test.com",
	}

	assert.Equal(t, "PASSED", reasoning.Status)
	assert.Equal(t, 0.85, reasoning.OverallScore)
	assert.Equal(t, "high", reasoning.ConfidenceLevel)
	assert.Equal(t, "Test explanation", reasoning.Explanation)
	assert.Equal(t, "test-123", reasoning.VerificationID)
	assert.Equal(t, "Test Business", reasoning.BusinessName)
	assert.Equal(t, "https://test.com", reasoning.WebsiteURL)
}

func TestFieldAnalysis_StructFields(t *testing.T) {
	analysis := &FieldAnalysis{
		FieldName:    "business_name",
		Score:        0.9,
		Status:       "passed",
		Explanation:  "Test explanation",
		Evidence:     "Test evidence",
		Confidence:   0.8,
		Weight:       0.3,
		Contribution: 0.27,
	}

	assert.Equal(t, "business_name", analysis.FieldName)
	assert.Equal(t, 0.9, analysis.Score)
	assert.Equal(t, "passed", analysis.Status)
	assert.Equal(t, "Test explanation", analysis.Explanation)
	assert.Equal(t, "Test evidence", analysis.Evidence)
	assert.Equal(t, 0.8, analysis.Confidence)
	assert.Equal(t, 0.3, analysis.Weight)
	assert.Equal(t, 0.27, analysis.Contribution)
}

func TestRecommendation_StructFields(t *testing.T) {
	recommendation := &Recommendation{
		Type:        "verification",
		Priority:    "high",
		Description: "Test description",
		Action:      "Test action",
		Reason:      "Test reason",
		Impact:      "Test impact",
	}

	assert.Equal(t, "verification", recommendation.Type)
	assert.Equal(t, "high", recommendation.Priority)
	assert.Equal(t, "Test description", recommendation.Description)
	assert.Equal(t, "Test action", recommendation.Action)
	assert.Equal(t, "Test reason", recommendation.Reason)
	assert.Equal(t, "Test impact", recommendation.Impact)
}

func TestRiskFactor_StructFields(t *testing.T) {
	riskFactor := &RiskFactor{
		Factor:      "test_factor",
		Severity:    "high",
		Description: "Test description",
		Impact:      "Test impact",
		Mitigation:  "Test mitigation",
		Probability: 0.7,
	}

	assert.Equal(t, "test_factor", riskFactor.Factor)
	assert.Equal(t, "high", riskFactor.Severity)
	assert.Equal(t, "Test description", riskFactor.Description)
	assert.Equal(t, "Test impact", riskFactor.Impact)
	assert.Equal(t, "Test mitigation", riskFactor.Mitigation)
	assert.Equal(t, 0.7, riskFactor.Probability)
}

func TestVerificationReasoningGenerator_GenerateVerificationReport(t *testing.T) {
	generator := NewVerificationReasoningGenerator(nil)

	tests := []struct {
		name           string
		verificationID string
		businessName   string
		websiteURL     string
		result         *VerificationResult
		comparison     *ComparisonResult
		includeAudit   bool
		metadata       map[string]interface{}
		expectError    bool
		checkReport    func(*testing.T, *VerificationReport)
	}{
		{
			name:           "nil result returns error",
			verificationID: "test-123",
			businessName:   "Test Business",
			websiteURL:     "https://test.com",
			result:         nil,
			comparison:     nil,
			includeAudit:   true,
			metadata:       nil,
			expectError:    true,
		},
		{
			name:           "successful report generation with audit",
			verificationID: "test-123",
			businessName:   "Test Business",
			websiteURL:     "https://test.com",
			result: &VerificationResult{
				ID:           "test-123",
				Status:       StatusPassed,
				OverallScore: 0.85,
				FieldResults: map[string]FieldResult{
					"business_name": {
						Status:     StatusPassed,
						Score:      0.9,
						Confidence: 0.8,
						Matched:    true,
					},
				},
			},
			comparison: &ComparisonResult{
				OverallScore: 0.85,
				FieldResults: map[string]FieldComparison{
					"business_name": {
						Score:      0.9,
						Confidence: 0.8,
						Matched:    true,
					},
				},
			},
			includeAudit: true,
			metadata: map[string]interface{}{
				"source": "test",
			},
			expectError: false,
			checkReport: func(t *testing.T, report *VerificationReport) {
				assert.NotNil(t, report)
				assert.Equal(t, "test-123", report.VerificationID)
				assert.Equal(t, "Test Business", report.BusinessName)
				assert.Equal(t, "https://test.com", report.WebsiteURL)
				assert.Equal(t, "PASSED", report.Status)
				assert.Equal(t, 0.85, report.OverallScore)
				assert.Equal(t, "high", report.ConfidenceLevel)
				assert.NotNil(t, report.Reasoning)
				assert.NotNil(t, report.ComparisonDetails)
				assert.NotEmpty(t, report.AuditTrail)
				assert.Equal(t, "test", report.Metadata["source"])
				assert.True(t, len(report.AuditTrail) >= 5) // Should have multiple audit events
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := generator.GenerateVerificationReport(
				tt.verificationID,
				tt.businessName,
				tt.websiteURL,
				tt.result,
				tt.comparison,
				tt.includeAudit,
				tt.metadata,
			)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, report)
			} else {
				assert.NoError(t, err)
				if tt.checkReport != nil {
					tt.checkReport(t, report)
				}
			}
		})
	}
}

func TestVerificationReasoningGenerator_EnhancedRecommendations(t *testing.T) {
	generator := NewVerificationReasoningGenerator(nil)

	tests := []struct {
		name          string
		result        *VerificationResult
		comparison    *ComparisonResult
		expectedTypes []string // Expected recommendation types
		minCount      int      // Minimum number of recommendations
	}{
		{
			name: "high confidence passed verification",
			result: &VerificationResult{
				Status:       StatusPassed,
				OverallScore: 0.96,
			},
			comparison: &ComparisonResult{
				FieldResults: map[string]FieldComparison{
					"business_name": {Score: 0.95, Confidence: 0.9, Matched: true},
				},
			},
			expectedTypes: []string{"verification_complete"},
			minCount:      1,
		},
		{
			name: "medium confidence passed verification",
			result: &VerificationResult{
				Status:       StatusPassed,
				OverallScore: 0.87,
			},
			comparison: &ComparisonResult{
				FieldResults: map[string]FieldComparison{
					"business_name": {Score: 0.85, Confidence: 0.8, Matched: true},
				},
			},
			expectedTypes: []string{"verification_passed"},
			minCount:      1,
		},
		{
			name: "low confidence passed verification",
			result: &VerificationResult{
				Status:       StatusPassed,
				OverallScore: 0.75,
			},
			comparison: &ComparisonResult{
				FieldResults: map[string]FieldComparison{
					"business_name": {Score: 0.7, Confidence: 0.6, Matched: true},
				},
			},
			expectedTypes: []string{"monitoring"},
			minCount:      1,
		},
		{
			name: "partial verification with failed fields",
			result: &VerificationResult{
				Status:       StatusPartial,
				OverallScore: 0.65,
			},
			comparison: &ComparisonResult{
				FieldResults: map[string]FieldComparison{
					"business_name":   {Score: 0.9, Confidence: 0.8, Matched: true},
					"phone_numbers":   {Score: 0.4, Confidence: 0.5, Matched: false},
					"email_addresses": {Score: 0.3, Confidence: 0.4, Matched: false},
				},
			},
			expectedTypes: []string{"manual_review", "contact_verification", "email_verification"},
			minCount:      3,
		},
		{
			name: "failed verification with multiple issues",
			result: &VerificationResult{
				Status:       StatusFailed,
				OverallScore: 0.35,
			},
			comparison: &ComparisonResult{
				FieldResults: map[string]FieldComparison{
					"business_name": {Score: 0.3, Confidence: 0.4, Matched: false},
					"website_url":   {Score: 0.2, Confidence: 0.3, Matched: false},
				},
			},
			expectedTypes: []string{"investigation", "documentation_request", "risk_assessment"},
			minCount:      4, // Should include field-specific recommendations too
		},
		{
			name: "verification with missing critical data",
			result: &VerificationResult{
				Status:       StatusPartial,
				OverallScore: 0.6,
			},
			comparison: &ComparisonResult{
				FieldResults: map[string]FieldComparison{
					// Missing business_name, phone_numbers, addresses, website_url
					"industry": {Score: 0.8, Confidence: 0.7, Matched: true},
				},
			},
			expectedTypes: []string{"manual_review", "missing_data"},
			minCount:      2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendations := generator.generateRecommendations(tt.result, tt.comparison)

			assert.True(t, len(recommendations) >= tt.minCount,
				"Expected at least %d recommendations, got %d", tt.minCount, len(recommendations))

			// Check that expected recommendation types are present
			foundTypes := make(map[string]bool)
			for _, rec := range recommendations {
				foundTypes[rec.Type] = true
			}

			for _, expectedType := range tt.expectedTypes {
				assert.True(t, foundTypes[expectedType],
					"Expected recommendation type '%s' not found", expectedType)
			}

			// Verify all recommendations have required fields
			for _, rec := range recommendations {
				assert.NotEmpty(t, rec.Type, "Recommendation type should not be empty")
				assert.NotEmpty(t, rec.Priority, "Recommendation priority should not be empty")
				assert.NotEmpty(t, rec.Description, "Recommendation description should not be empty")
				assert.NotEmpty(t, rec.Action, "Recommendation action should not be empty")
				assert.NotEmpty(t, rec.Reason, "Recommendation reason should not be empty")
				assert.NotEmpty(t, rec.Impact, "Recommendation impact should not be empty")
				assert.Contains(t, []string{"low", "medium", "high"}, rec.Priority,
					"Priority should be low, medium, or high")
			}
		})
	}
}

func TestVerificationReasoningGenerator_CreateFieldRecommendation(t *testing.T) {
	generator := NewVerificationReasoningGenerator(nil)

	tests := []struct {
		name             string
		fieldName        string
		score            float64
		confidence       float64
		expectedType     string
		expectedPriority string
	}{
		{
			name:             "business name high priority",
			fieldName:        "business_name",
			score:            0.2,
			confidence:       0.3,
			expectedType:     "business_name_verification",
			expectedPriority: "high",
		},
		{
			name:             "phone number medium priority",
			fieldName:        "phone_numbers",
			score:            0.4,
			confidence:       0.5,
			expectedType:     "contact_verification",
			expectedPriority: "medium",
		},
		{
			name:             "email low priority",
			fieldName:        "email_addresses",
			score:            0.55,
			confidence:       0.6,
			expectedType:     "email_verification",
			expectedPriority: "low",
		},
		{
			name:             "address verification",
			fieldName:        "addresses",
			score:            0.3,
			confidence:       0.4,
			expectedType:     "address_verification",
			expectedPriority: "medium", // 0.3 is not < 0.3, so it's medium priority
		},
		{
			name:             "website verification",
			fieldName:        "website_url",
			score:            0.25,
			confidence:       0.35,
			expectedType:     "website_verification",
			expectedPriority: "high",
		},
		{
			name:             "industry classification",
			fieldName:        "industry",
			score:            0.4,
			confidence:       0.5,
			expectedType:     "industry_classification",
			expectedPriority: "low", // Always low priority for industry
		},
		{
			name:             "unknown field",
			fieldName:        "unknown_field",
			score:            0.3,
			confidence:       0.4,
			expectedType:     "field_verification",
			expectedPriority: "medium", // 0.3 is not < 0.3, so it's medium priority
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendation := generator.createFieldRecommendation(tt.fieldName, tt.score, tt.confidence)

			assert.NotNil(t, recommendation)
			assert.Equal(t, tt.expectedType, recommendation.Type)
			assert.Equal(t, tt.expectedPriority, recommendation.Priority)
			assert.NotEmpty(t, recommendation.Description)
			assert.NotEmpty(t, recommendation.Action)
			assert.NotEmpty(t, recommendation.Reason)
			assert.NotEmpty(t, recommendation.Impact)
			assert.Contains(t, recommendation.Reason, fmt.Sprintf("%.2f", tt.score))
			assert.Contains(t, recommendation.Reason, fmt.Sprintf("%.2f", tt.confidence))
		})
	}
}

func TestVerificationReasoningGenerator_GenerateDataQualityRecommendations(t *testing.T) {
	generator := NewVerificationReasoningGenerator(nil)

	tests := []struct {
		name          string
		comparison    *ComparisonResult
		expectedTypes []string
	}{
		{
			name:          "nil comparison returns empty",
			comparison:    nil,
			expectedTypes: []string{},
		},
		{
			name: "high quality data with all critical fields",
			comparison: &ComparisonResult{
				FieldResults: map[string]FieldComparison{
					"business_name": {Score: 0.9, Confidence: 0.8, Matched: true},
					"phone_numbers": {Score: 0.85, Confidence: 0.75, Matched: true},
					"addresses":     {Score: 0.8, Confidence: 0.7, Matched: true},
					"website_url":   {Score: 0.9, Confidence: 0.85, Matched: true},
				},
			},
			expectedTypes: []string{},
		},
		{
			name: "low quality data triggers improvement recommendation",
			comparison: &ComparisonResult{
				FieldResults: map[string]FieldComparison{
					"business_name":   {Score: 0.9, Confidence: 0.5, Matched: true},  // Low confidence
					"phone_numbers":   {Score: 0.85, Confidence: 0.4, Matched: true}, // Low confidence
					"addresses":       {Score: 0.8, Confidence: 0.5, Matched: true},  // Low confidence
					"website_url":     {Score: 0.9, Confidence: 0.85, Matched: true}, // Good confidence
					"email_addresses": {Score: 0.7, Confidence: 0.8, Matched: true},  // Good confidence
				},
			},
			expectedTypes: []string{"data_source_improvement"},
		},
		{
			name: "missing critical fields",
			comparison: &ComparisonResult{
				FieldResults: map[string]FieldComparison{
					// Missing business_name, phone_numbers, addresses, website_url
					"industry": {Score: 0.8, Confidence: 0.7, Matched: true},
				},
			},
			expectedTypes: []string{"missing_data"},
		},
		{
			name: "both low quality and missing data",
			comparison: &ComparisonResult{
				FieldResults: map[string]FieldComparison{
					// Missing critical fields, and has low confidence
					"industry": {Score: 0.8, Confidence: 0.3, Matched: true},
					"services": {Score: 0.6, Confidence: 0.2, Matched: true},
				},
			},
			expectedTypes: []string{"data_source_improvement", "missing_data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendations := generator.generateDataQualityRecommendations(tt.comparison)

			// Check expected types
			foundTypes := make(map[string]bool)
			for _, rec := range recommendations {
				foundTypes[rec.Type] = true
			}

			assert.Equal(t, len(tt.expectedTypes), len(recommendations),
				"Expected %d recommendations, got %d", len(tt.expectedTypes), len(recommendations))

			for _, expectedType := range tt.expectedTypes {
				assert.True(t, foundTypes[expectedType],
					"Expected recommendation type '%s' not found", expectedType)
			}
		})
	}
}
