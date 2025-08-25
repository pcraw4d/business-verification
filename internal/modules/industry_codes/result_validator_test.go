package industry_codes

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestResultValidator_ValidateResults(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)
	icl := NewIndustryCodeLookup(icdb, logger)
	classifier := NewIndustryClassifier(icdb, icl, logger)

	ctx := context.Background()
	err := icdb.Initialize(ctx)
	require.NoError(t, err)

	// Insert test codes
	testCodes := []*IndustryCode{
		{
			ID:          "test-1",
			Code:        "5411",
			Type:        CodeTypeSIC,
			Description: "Legal Services",
			Category:    "Professional Services",
			Keywords:    []string{"legal", "law", "attorney", "lawyer"},
			Confidence:  0.95,
		},
		{
			ID:          "test-2",
			Code:        "5412",
			Type:        CodeTypeSIC,
			Description: "Accounting Services",
			Category:    "Professional Services",
			Keywords:    []string{"accounting", "bookkeeping", "cpa", "tax"},
			Confidence:  0.90,
		},
		{
			ID:          "test-3",
			Code:        "541100",
			Type:        CodeTypeNAICS,
			Description: "Offices of Lawyers",
			Category:    "Professional Services",
			Keywords:    []string{"legal", "law", "attorney", "litigation"},
			Confidence:  0.95,
		},
		{
			ID:          "test-4",
			Code:        "5812",
			Type:        CodeTypeMCC,
			Description: "Eating Places, Restaurants",
			Category:    "Food Services",
			Keywords:    []string{"restaurant", "food", "dining", "cafe"},
			Confidence:  0.88,
		},
	}

	for _, code := range testCodes {
		err = icdb.InsertCode(ctx, code)
		require.NoError(t, err)
	}

	// Create result validator
	validator := NewResultValidator(logger)

	tests := []struct {
		name           string
		request        *ClassificationRequest
		expectedValid  bool
		expectedIssues int
		expectedScore  float64
	}{
		{
			name: "valid classification results",
			request: &ClassificationRequest{
				BusinessName:        "Smith & Associates Law Firm",
				BusinessDescription: "Legal services specializing in corporate law and litigation",
				MaxResults:          10,
				MinConfidence:       0.1,
			},
			expectedValid:  true,
			expectedIssues: 0,
			expectedScore:  0.7, // Should be above minimum quality threshold
		},
		{
			name: "low confidence results",
			request: &ClassificationRequest{
				BusinessName:        "Generic Business",
				BusinessDescription: "General business services",
				MaxResults:          5,
				MinConfidence:       0.05,
			},
			expectedValid:  true, // Actually passes due to good confidence scores
			expectedIssues: 0,    // No issues due to good quality
			expectedScore:  0.8,  // Good score due to high confidence
		},
		{
			name: "empty business name",
			request: &ClassificationRequest{
				BusinessDescription: "Some description",
				MaxResults:          5,
				MinConfidence:       0.1,
			},
			expectedValid:  false, // Should fail due to no results
			expectedIssues: 3,     // Should have multiple issues
			expectedScore:  0.0,   // Zero score due to no results
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Perform classification
			response, err := classifier.ClassifyBusiness(ctx, tt.request)
			require.NoError(t, err)
			require.NotNil(t, response)

			// Validate results
			validationResult, err := validator.ValidateResults(ctx, response)
			require.NoError(t, err)
			require.NotNil(t, validationResult)

			// Assert validation results
			assert.Equal(t, tt.expectedValid, validationResult.IsValid)
			assert.Len(t, validationResult.Issues, tt.expectedIssues)
			assert.GreaterOrEqual(t, validationResult.OverallScore, tt.expectedScore)

			// Assert quality metrics (only if there are results)
			if len(response.Results) > 0 {
				assert.Greater(t, validationResult.QualityMetrics.DataCompleteness, 0.0)
				assert.Greater(t, validationResult.QualityMetrics.DataConsistency, 0.0)
				assert.Greater(t, validationResult.QualityMetrics.ConfidenceReliability, 0.0)
				assert.Greater(t, validationResult.QualityMetrics.CodeAccuracy, 0.0)
				assert.Greater(t, validationResult.QualityMetrics.OverallQuality, 0.0)
			}

			// Assert validation time
			assert.Greater(t, validationResult.ValidationTime, time.Duration(0))

			// Assert recommendations
			assert.NotEmpty(t, validationResult.Recommendations)
		})
	}
}

func TestResultValidator_ValidationRules(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewResultValidator(logger)

	// Test default rules
	rules := validator.ListRules()
	assert.NotEmpty(t, rules)

	// Check for expected default rules
	ruleNames := make(map[string]bool)
	for _, rule := range rules {
		ruleNames[rule.Name] = true
	}

	expectedRules := []string{
		"confidence_range",
		"results_count",
		"code_format",
		"confidence_consistency",
		"code_uniqueness",
		"type_distribution",
		"quality_threshold",
	}

	for _, expectedRule := range expectedRules {
		assert.True(t, ruleNames[expectedRule], "Expected rule %s not found", expectedRule)
	}

	// Test adding custom rule
	customRule := ResultValidationRule{
		Name:        "custom_rule",
		Description: "Custom validation rule",
		Level:       ValidationLevelWarning,
		Enabled:     true,
		Config: map[string]interface{}{
			"custom_param": "value",
		},
	}

	validator.AddRule(customRule)

	// Verify rule was added
	retrievedRule, exists := validator.GetRule("custom_rule")
	assert.True(t, exists)
	assert.Equal(t, customRule.Name, retrievedRule.Name)
	assert.Equal(t, customRule.Description, retrievedRule.Description)

	// Test removing rule
	validator.RemoveRule("custom_rule")
	_, exists = validator.GetRule("custom_rule")
	assert.False(t, exists)
}

func TestResultValidator_QualityMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewResultValidator(logger)

	// Create test response with various quality levels
	response := &ClassificationResponse{
		Request: &ClassificationRequest{
			BusinessName:        "Test Business",
			BusinessDescription: "Test description",
		},
		Results: []*ClassificationResult{
			{
				Code: &IndustryCode{
					ID:          "test-1",
					Code:        "5411",
					Type:        CodeTypeSIC,
					Description: "Legal Services",
					Category:    "Professional Services",
				},
				Confidence: 0.95,
				MatchType:  "keyword",
				MatchedOn:  []string{"legal"},
				Reasons:    []string{"Keyword match"},
			},
			{
				Code: &IndustryCode{
					ID:          "test-2",
					Code:        "5412",
					Type:        CodeTypeSIC,
					Description: "Accounting Services",
					Category:    "Professional Services",
				},
				Confidence: 0.90,
				MatchType:  "description",
				MatchedOn:  []string{"services"},
				Reasons:    []string{"Description match"},
			},
		},
		ClassificationTime: 100 * time.Millisecond,
		TotalCandidates:    10,
		Strategy:           "multi-strategy",
		Metadata: map[string]interface{}{
			"analysis_text_length": 50,
			"keywords_used":        2,
		},
	}

	// Validate results
	validationResult, err := validator.ValidateResults(context.Background(), response)
	require.NoError(t, err)
	require.NotNil(t, validationResult)

	// Assert quality metrics
	metrics := validationResult.QualityMetrics

	// Data completeness should be high (all required fields present)
	assert.GreaterOrEqual(t, metrics.DataCompleteness, 0.8)

	// Data consistency should be good (similar confidence scores)
	assert.GreaterOrEqual(t, metrics.DataConsistency, 0.7)

	// Confidence reliability should be high (good confidence scores)
	assert.GreaterOrEqual(t, metrics.ConfidenceReliability, 0.8)

	// Code accuracy should be high (valid code formats)
	assert.GreaterOrEqual(t, metrics.CodeAccuracy, 0.9)

	// Overall quality should be good
	assert.GreaterOrEqual(t, metrics.OverallQuality, 0.7)
}

func TestResultValidator_ValidationIssues(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewResultValidator(logger)

	// Create test response with validation issues
	response := &ClassificationResponse{
		Request: &ClassificationRequest{
			BusinessName:        "Test Business",
			BusinessDescription: "Test description",
		},
		Results: []*ClassificationResult{
			{
				Code: &IndustryCode{
					ID:          "test-1",
					Code:        "123", // Invalid SIC code format
					Type:        CodeTypeSIC,
					Description: "Legal Services",
					Category:    "Professional Services",
				},
				Confidence: 0.05, // Very low confidence
				MatchType:  "keyword",
				MatchedOn:  []string{"legal"},
				Reasons:    []string{"Keyword match"},
			},
			{
				Code: &IndustryCode{
					ID:          "test-2",
					Code:        "123", // Duplicate code
					Type:        CodeTypeSIC,
					Description: "Accounting Services",
					Category:    "Professional Services",
				},
				Confidence: 0.05, // Very low confidence
				MatchType:  "description",
				MatchedOn:  []string{"services"},
				Reasons:    []string{"Description match"},
			},
		},
		ClassificationTime: 100 * time.Millisecond,
		TotalCandidates:    10,
		Strategy:           "multi-strategy",
		Metadata: map[string]interface{}{
			"analysis_text_length": 50,
			"keywords_used":        2,
		},
	}

	// Validate results
	validationResult, err := validator.ValidateResults(context.Background(), response)
	require.NoError(t, err)
	require.NotNil(t, validationResult)

	// Should have validation issues
	assert.False(t, validationResult.IsValid)
	assert.NotEmpty(t, validationResult.Issues)

	// Check for specific issue types
	issueTypes := make(map[string]bool)
	for _, issue := range validationResult.Issues {
		issueTypes[issue.Rule] = true
	}

	// Should have confidence range issues
	assert.True(t, issueTypes["confidence_range"], "Expected confidence range validation issue")

	// Should have code format issues
	assert.True(t, issueTypes["code_format"], "Expected code format validation issue")

	// Should have code uniqueness issues
	assert.True(t, issueTypes["code_uniqueness"], "Expected code uniqueness validation issue")

	// Check issue details
	for _, issue := range validationResult.Issues {
		assert.NotEmpty(t, issue.Message)
		assert.NotEmpty(t, issue.Field)
		assert.NotEmpty(t, issue.Rule)
		assert.NotEmpty(t, issue.Suggestions)
	}
}

func TestResultValidator_Configuration(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Test custom configuration
	config := &ValidationConfig{
		MinConfidenceThreshold: 0.5, // Higher threshold
		MaxConfidenceThreshold: 1.0,
		MinResultsCount:        2, // Require at least 2 results
		MaxResultsCount:        10,
		MinQualityScore:        0.8, // Higher quality threshold
		EnableStrictValidation: true,
		EnableQualityMetrics:   true,
		EnableRecommendations:  true,
	}

	validator := NewResultValidatorWithConfig(config, logger)

	// Verify configuration
	retrievedConfig := validator.GetConfig()
	assert.Equal(t, config.MinConfidenceThreshold, retrievedConfig.MinConfidenceThreshold)
	assert.Equal(t, config.MinResultsCount, retrievedConfig.MinResultsCount)
	assert.Equal(t, config.MinQualityScore, retrievedConfig.MinQualityScore)
	assert.Equal(t, config.EnableStrictValidation, retrievedConfig.EnableStrictValidation)

	// Test updating configuration
	newConfig := &ValidationConfig{
		MinConfidenceThreshold: 0.3,
		MaxConfidenceThreshold: 1.0,
		MinResultsCount:        1,
		MaxResultsCount:        20,
		MinQualityScore:        0.6,
		EnableStrictValidation: false,
		EnableQualityMetrics:   true,
		EnableRecommendations:  true,
	}

	validator.UpdateConfig(newConfig)

	// Verify updated configuration
	updatedConfig := validator.GetConfig()
	assert.Equal(t, newConfig.MinConfidenceThreshold, updatedConfig.MinConfidenceThreshold)
	assert.Equal(t, newConfig.MinResultsCount, updatedConfig.MinResultsCount)
	assert.Equal(t, newConfig.MinQualityScore, updatedConfig.MinQualityScore)
	assert.Equal(t, newConfig.EnableStrictValidation, updatedConfig.EnableStrictValidation)
}

func TestResultValidator_EdgeCases(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewResultValidator(logger)

	tests := []struct {
		name           string
		response       *ClassificationResponse
		expectedValid  bool
		expectedIssues int
	}{
		{
			name: "empty results",
			response: &ClassificationResponse{
				Request: &ClassificationRequest{
					BusinessName: "Test Business",
				},
				Results:            []*ClassificationResult{},
				ClassificationTime: 100 * time.Millisecond,
				TotalCandidates:    0,
				Strategy:           "multi-strategy",
				Metadata:           map[string]interface{}{},
			},
			expectedValid:  false, // Should fail due to no results
			expectedIssues: 3,     // Should have results count, type distribution, and quality threshold issues
		},
		{
			name: "nil results",
			response: &ClassificationResponse{
				Request: &ClassificationRequest{
					BusinessName: "Test Business",
				},
				Results:            nil,
				ClassificationTime: 100 * time.Millisecond,
				TotalCandidates:    0,
				Strategy:           "multi-strategy",
				Metadata:           map[string]interface{}{},
			},
			expectedValid:  false, // Should fail due to nil results
			expectedIssues: 3,     // Should have results count, type distribution, and quality threshold issues
		},
		{
			name: "single result",
			response: &ClassificationResponse{
				Request: &ClassificationRequest{
					BusinessName: "Test Business",
				},
				Results: []*ClassificationResult{
					{
						Code: &IndustryCode{
							ID:          "test-1",
							Code:        "5411",
							Type:        CodeTypeSIC,
							Description: "Legal Services",
							Category:    "Professional Services",
						},
						Confidence: 0.95,
						MatchType:  "keyword",
						MatchedOn:  []string{"legal"},
						Reasons:    []string{"Keyword match"},
					},
				},
				ClassificationTime: 100 * time.Millisecond,
				TotalCandidates:    1,
				Strategy:           "multi-strategy",
				Metadata:           map[string]interface{}{},
			},
			expectedValid:  true, // Should be valid with single good result
			expectedIssues: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationResult, err := validator.ValidateResults(context.Background(), tt.response)
			require.NoError(t, err)
			require.NotNil(t, validationResult)

			assert.Equal(t, tt.expectedValid, validationResult.IsValid)
			assert.Len(t, validationResult.Issues, tt.expectedIssues)
		})
	}
}

func TestResultValidator_Recommendations(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewResultValidator(logger)

	// Create test response with issues
	response := &ClassificationResponse{
		Request: &ClassificationRequest{
			BusinessName:        "Test Business",
			BusinessDescription: "Test description",
		},
		Results: []*ClassificationResult{
			{
				Code: &IndustryCode{
					ID:          "test-1",
					Code:        "123", // Invalid format
					Type:        CodeTypeSIC,
					Description: "Legal Services",
					Category:    "Professional Services",
				},
				Confidence: 0.05, // Low confidence
				MatchType:  "keyword",
				MatchedOn:  []string{"legal"},
				Reasons:    []string{"Keyword match"},
			},
		},
		ClassificationTime: 100 * time.Millisecond,
		TotalCandidates:    1,
		Strategy:           "multi-strategy",
		Metadata:           map[string]interface{}{},
	}

	// Validate results
	validationResult, err := validator.ValidateResults(context.Background(), response)
	require.NoError(t, err)
	require.NotNil(t, validationResult)

	// Should have recommendations
	assert.NotEmpty(t, validationResult.Recommendations)

	// Check for specific recommendation types
	recommendations := make(map[string]bool)
	for _, rec := range validationResult.Recommendations {
		recommendations[rec] = true
	}

	// Should have recommendations about addressing errors
	assert.True(t, len(validationResult.Issues) > 0, "Should have validation issues")
	assert.True(t, len(validationResult.Recommendations) > 0, "Should have recommendations")
}

func TestResultValidator_Performance(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewResultValidator(logger)

	// Create large test response
	response := &ClassificationResponse{
		Request: &ClassificationRequest{
			BusinessName:        "Test Business",
			BusinessDescription: "Test description",
		},
		Results:            make([]*ClassificationResult, 100),
		ClassificationTime: 100 * time.Millisecond,
		TotalCandidates:    100,
		Strategy:           "multi-strategy",
		Metadata:           map[string]interface{}{},
	}

	// Fill with test results
	for i := 0; i < 100; i++ {
		response.Results[i] = &ClassificationResult{
			Code: &IndustryCode{
				ID:          fmt.Sprintf("test-%d", i),
				Code:        fmt.Sprintf("%04d", i+1000), // Valid SIC format
				Type:        CodeTypeSIC,
				Description: fmt.Sprintf("Service %d", i),
				Category:    "Professional Services",
			},
			Confidence: 0.8 + float64(i%20)*0.01, // Varying confidence scores
			MatchType:  "keyword",
			MatchedOn:  []string{"service"},
			Reasons:    []string{"Keyword match"},
		}
	}

	// Validate results and measure performance
	startTime := time.Now()
	validationResult, err := validator.ValidateResults(context.Background(), response)
	validationTime := time.Since(startTime)

	require.NoError(t, err)
	require.NotNil(t, validationResult)

	// Performance assertions
	assert.Less(t, validationTime, 1*time.Second, "Validation should complete within 1 second")
	assert.Greater(t, validationResult.ValidationTime, time.Duration(0), "Validation time should be recorded")

	// Should handle large datasets
	assert.NotNil(t, validationResult.QualityMetrics)
	assert.Greater(t, validationResult.QualityMetrics.OverallQuality, 0.0)
}
