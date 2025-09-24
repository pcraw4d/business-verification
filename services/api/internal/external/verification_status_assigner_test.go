package external

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewStatusAssigner(t *testing.T) {
	logger := zap.NewNop()

	// Test with nil criteria (should use defaults)
	assigner := NewStatusAssigner(nil, logger)
	assert.NotNil(t, assigner)
	assert.NotNil(t, assigner.criteria)
	assert.Equal(t, 0.8, assigner.criteria.PassedThreshold)
	assert.Equal(t, 0.6, assigner.criteria.PartialThreshold)
	assert.Contains(t, assigner.criteria.CriticalFields, "business_name")
	assert.Contains(t, assigner.criteria.CriticalFields, "phone_numbers")
	assert.Contains(t, assigner.criteria.CriticalFields, "email_addresses")

	// Test with custom criteria
	customCriteria := &VerificationCriteria{
		PassedThreshold:    0.9,
		PartialThreshold:   0.7,
		CriticalFields:     []string{"business_name"},
		MaxDistanceKm:      25.0,
		MinConfidenceLevel: "high",
		FieldRequirements: map[string]FieldRequirement{
			"business_name": {
				Required:      true,
				MinScore:      0.8,
				MinConfidence: 0.7,
				Weight:        0.5,
			},
		},
	}

	assigner = NewStatusAssigner(customCriteria, logger)
	assert.NotNil(t, assigner)
	assert.Equal(t, customCriteria, assigner.criteria)
	assert.Equal(t, 0.9, assigner.criteria.PassedThreshold)
	assert.Equal(t, 0.7, assigner.criteria.PartialThreshold)
}

func TestStatusAssigner_AssignVerificationStatus(t *testing.T) {
	logger := zap.NewNop()
	assigner := NewStatusAssigner(nil, logger)

	// Test successful verification
	comparisonResult := &ComparisonResult{
		OverallScore:    0.85,
		ConfidenceLevel: "high",
		FieldResults: map[string]FieldComparison{
			"business_name": {
				Score:      0.9,
				Confidence: 0.8,
				Matched:    true,
				Reasoning:  "High similarity match",
			},
			"phone_numbers": {
				Score:      0.95,
				Confidence: 0.9,
				Matched:    true,
				Reasoning:  "Exact match",
			},
			"email_addresses": {
				Score:      0.9,
				Confidence: 0.8,
				Matched:    true,
				Reasoning:  "Exact match",
			},
		},
	}

	result, err := assigner.AssignVerificationStatus(context.Background(), comparisonResult)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, StatusPassed, result.Status)
	assert.Equal(t, 0.85, result.OverallScore)
	assert.Equal(t, "high", result.ConfidenceLevel)
	assert.NotEmpty(t, result.Reasoning)
	assert.NotEmpty(t, result.ID)
}

func TestStatusAssigner_AssignVerificationStatus_Partial(t *testing.T) {
	logger := zap.NewNop()
	assigner := NewStatusAssigner(nil, logger)

	// Test partial verification
	comparisonResult := &ComparisonResult{
		OverallScore:    0.7,
		ConfidenceLevel: "medium",
		FieldResults: map[string]FieldComparison{
			"business_name": {
				Score:      0.8,
				Confidence: 0.7,
				Matched:    true,
				Reasoning:  "Good match",
			},
			"phone_numbers": {
				Score:      0.8,
				Confidence: 0.7,
				Matched:    true,
				Reasoning:  "Good match",
			},
			"email_addresses": {
				Score:      0.9,
				Confidence: 0.8,
				Matched:    true,
				Reasoning:  "Exact match",
			},
		},
	}

	result, err := assigner.AssignVerificationStatus(context.Background(), comparisonResult)
	require.NoError(t, err)
	assert.Equal(t, StatusPartial, result.Status)
	assert.Equal(t, 0.7, result.OverallScore)
}

func TestStatusAssigner_AssignVerificationStatus_Failed(t *testing.T) {
	logger := zap.NewNop()
	assigner := NewStatusAssigner(nil, logger)

	// Test failed verification
	comparisonResult := &ComparisonResult{
		OverallScore:    0.4,
		ConfidenceLevel: "low",
		FieldResults: map[string]FieldComparison{
			"business_name": {
				Score:      0.3,
				Confidence: 0.2,
				Matched:    false,
				Reasoning:  "Poor match",
			},
			"phone_numbers": {
				Score:      0.0,
				Confidence: 0.0,
				Matched:    false,
				Reasoning:  "No match",
			},
			"email_addresses": {
				Score:      0.5,
				Confidence: 0.4,
				Matched:    false,
				Reasoning:  "Partial match",
			},
		},
	}

	result, err := assigner.AssignVerificationStatus(context.Background(), comparisonResult)
	require.NoError(t, err)
	assert.Equal(t, StatusFailed, result.Status)
	assert.Equal(t, 0.4, result.OverallScore)
}

func TestStatusAssigner_AssignVerificationStatus_MissingCriticalFields(t *testing.T) {
	logger := zap.NewNop()
	assigner := NewStatusAssigner(nil, logger)

	// Test with missing critical fields
	comparisonResult := &ComparisonResult{
		OverallScore:    0.8,
		ConfidenceLevel: "high",
		FieldResults: map[string]FieldComparison{
			"business_name": {
				Score:      0.9,
				Confidence: 0.8,
				Matched:    true,
				Reasoning:  "Good match",
			},
			// Missing phone_numbers and email_addresses (critical fields)
		},
	}

	result, err := assigner.AssignVerificationStatus(context.Background(), comparisonResult)
	require.NoError(t, err)
	assert.Equal(t, StatusFailed, result.Status)
}

func TestStatusAssigner_determineFieldStatus(t *testing.T) {
	logger := zap.NewNop()
	assigner := NewStatusAssigner(nil, logger)

	// Test required field that fails
	requirement := FieldRequirement{
		Required:      true,
		MinScore:      0.7,
		MinConfidence: 0.6,
		Weight:        0.3,
	}

	fieldResult := FieldResult{
		Score:      0.3,
		Confidence: 0.2,
		Matched:    false,
	}

	status := assigner.determineFieldStatus(fieldResult, requirement)
	assert.Equal(t, StatusFailed, status)

	// Test required field that passes
	fieldResult = FieldResult{
		Score:      0.8,
		Confidence: 0.7,
		Matched:    true,
	}

	status = assigner.determineFieldStatus(fieldResult, requirement)
	assert.Equal(t, StatusPassed, status)

	// Test optional field that should be skipped
	requirement.Required = false
	fieldResult = FieldResult{
		Score:      0.3,
		Confidence: 0.2,
		Matched:    false,
	}

	status = assigner.determineFieldStatus(fieldResult, requirement)
	assert.Equal(t, StatusSkipped, status)
}

func TestStatusAssigner_determineOverallStatus(t *testing.T) {
	logger := zap.NewNop()
	assigner := NewStatusAssigner(nil, logger)

	// Test passed status
	result := &VerificationResult{
		OverallScore: 0.85,
		FieldResults: map[string]FieldResult{
			"business_name": {
				Status: StatusPassed,
				Score:  0.9,
			},
			"phone_numbers": {
				Status: StatusPassed,
				Score:  0.8,
			},
			"email_addresses": {
				Status: StatusPassed,
				Score:  0.8,
			},
		},
	}

	status := assigner.determineOverallStatus(result)
	assert.Equal(t, StatusPassed, status)

	// Test partial status
	result.OverallScore = 0.7
	status = assigner.determineOverallStatus(result)
	assert.Equal(t, StatusPartial, status)

	// Test failed status
	result.OverallScore = 0.4
	status = assigner.determineOverallStatus(result)
	assert.Equal(t, StatusFailed, status)
}

func TestStatusAssigner_checkCriticalFields(t *testing.T) {
	logger := zap.NewNop()
	assigner := NewStatusAssigner(nil, logger)

	// Test with all critical fields present and passed
	result := &VerificationResult{
		FieldResults: map[string]FieldResult{
			"business_name": {
				Status: StatusPassed,
			},
			"phone_numbers": {
				Status: StatusPassed,
			},
			"email_addresses": {
				Status: StatusPassed,
			},
		},
	}

	passed := assigner.checkCriticalFields(result)
	assert.True(t, passed)

	// Test with missing critical field
	result = &VerificationResult{
		FieldResults: map[string]FieldResult{
			"business_name": {
				Status: StatusPassed,
			},
			// Missing phone_numbers
			"email_addresses": {
				Status: StatusPassed,
			},
		},
	}

	passed = assigner.checkCriticalFields(result)
	assert.False(t, passed)

	// Test with failed critical field
	result = &VerificationResult{
		FieldResults: map[string]FieldResult{
			"business_name": {
				Status: StatusPassed,
			},
			"phone_numbers": {
				Status: StatusFailed,
			},
			"email_addresses": {
				Status: StatusPassed,
			},
		},
	}

	passed = assigner.checkCriticalFields(result)
	assert.False(t, passed)
}

func TestStatusAssigner_generateReasoning(t *testing.T) {
	logger := zap.NewNop()
	assigner := NewStatusAssigner(nil, logger)

	result := &VerificationResult{
		Status:       StatusPassed,
		OverallScore: 0.85,
		FieldResults: map[string]FieldResult{
			"business_name": {
				Status:     StatusPassed,
				Score:      0.9,
				Confidence: 0.8,
			},
			"phone_numbers": {
				Status:     StatusPassed,
				Score:      0.8,
				Confidence: 0.7,
			},
		},
	}

	reasoning := assigner.generateReasoning(result)
	assert.Contains(t, reasoning, "Overall verification score: 0.85")
	assert.Contains(t, reasoning, "All critical fields passed verification")
	assert.Contains(t, reasoning, "business_name: PASSED")
	assert.Contains(t, reasoning, "phone_numbers: PASSED")
}

func TestStatusAssigner_generateRecommendations(t *testing.T) {
	logger := zap.NewNop()
	assigner := NewStatusAssigner(nil, logger)

	// Test with failed critical fields
	result := &VerificationResult{
		OverallScore:    0.4,
		ConfidenceLevel: "low",
		FieldResults: map[string]FieldResult{
			"business_name": {
				Status:     StatusFailed,
				Confidence: 0.2,
			},
			"phone_numbers": {
				Status:     StatusPassed,
				Confidence: 0.3, // Low confidence
			},
		},
	}

	recommendations := assigner.generateRecommendations(result)
	assert.NotEmpty(t, recommendations)

	// Check that we have the expected recommendations (order may vary)
	recommendationTexts := make([]string, len(recommendations))
	for i, rec := range recommendations {
		recommendationTexts[i] = rec
	}

	assert.Contains(t, recommendationTexts, "Critical field 'business_name' failed verification - manual review required")
	assert.Contains(t, recommendationTexts, "Field 'phone_numbers' has low confidence (0.30) - additional verification recommended")
	assert.Contains(t, recommendationTexts, "Overall verification score is low - comprehensive manual review recommended")
	assert.Contains(t, recommendationTexts, "Low confidence level - consider additional data sources for verification")
}

func TestStatusAssigner_GetCriteria(t *testing.T) {
	logger := zap.NewNop()
	assigner := NewStatusAssigner(nil, logger)

	criteria := assigner.GetCriteria()
	assert.NotNil(t, criteria)
	assert.Equal(t, 0.8, criteria.PassedThreshold)
	assert.Equal(t, 0.6, criteria.PartialThreshold)
}

func TestStatusAssigner_UpdateCriteria(t *testing.T) {
	logger := zap.NewNop()
	assigner := NewStatusAssigner(nil, logger)

	newCriteria := &VerificationCriteria{
		PassedThreshold:  0.9,
		PartialThreshold: 0.7,
		CriticalFields:   []string{"business_name"},
	}

	assigner.UpdateCriteria(newCriteria)
	updatedCriteria := assigner.GetCriteria()
	assert.Equal(t, newCriteria, updatedCriteria)
	assert.Equal(t, 0.9, updatedCriteria.PassedThreshold)
	assert.Equal(t, 0.7, updatedCriteria.PartialThreshold)
}

func TestGenerateVerificationID(t *testing.T) {
	id1 := generateVerificationID()
	id2 := generateVerificationID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "ver_")
	assert.Contains(t, id2, "ver_")
}
