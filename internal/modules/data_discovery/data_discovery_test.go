package data_discovery

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDataDiscoveryService_DiscoverDataPoints(t *testing.T) {
	config := DefaultDataDiscoveryConfig()
	logger := zap.NewNop()
	service := NewDataDiscoveryService(config, logger)

	// Test content with various data types
	content := &ContentInput{
		RawContent: `
			Contact us at info@example.com or call (555) 123-4567
			Visit our website: https://example.com
			Address: 123 Main Street, Anytown, NY 12345
			Follow us on Facebook: https://facebook.com/example
			Business hours: Monday-Friday 9:00 AM - 5:00 PM
		`,
		ContentType: "html",
		URL:         "https://example.com",
	}

	ctx := context.Background()
	result, err := service.DiscoverDataPoints(ctx, content)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, len(result.DiscoveredFields), 0)
	assert.Greater(t, result.ConfidenceScore, 0.0)
	assert.Greater(t, len(result.PatternMatches), 0)
	assert.NotNil(t, result.ClassificationResult)
	assert.Greater(t, result.ProcessingTime, time.Duration(0))

	// Verify specific field discoveries
	fieldTypes := make(map[string]bool)
	for _, field := range result.DiscoveredFields {
		fieldTypes[field.FieldType] = true
	}

	// Should discover common field types
	assert.True(t, fieldTypes["email"], "Should discover email field")
	assert.True(t, fieldTypes["phone"], "Should discover phone field")
	assert.True(t, fieldTypes["url"], "Should discover URL field")
	assert.True(t, fieldTypes["address"], "Should discover address field")
	assert.True(t, fieldTypes["social_media"], "Should discover social media field")
}

func TestDataDiscoveryService_GetDiscoveredFieldsByPriority(t *testing.T) {
	config := DefaultDataDiscoveryConfig()
	logger := zap.NewNop()
	service := NewDataDiscoveryService(config, logger)

	// Create test fields with different priorities
	fields := []DiscoveredField{
		{
			FieldName:       "high_priority",
			FieldType:       "email",
			Priority:        1,
			BusinessValue:   0.9,
			ConfidenceScore: 0.8,
		},
		{
			FieldName:       "medium_priority",
			FieldType:       "phone",
			Priority:        3,
			BusinessValue:   0.7,
			ConfidenceScore: 0.8,
		},
		{
			FieldName:       "low_priority",
			FieldType:       "social_media",
			Priority:        5,
			BusinessValue:   0.5,
			ConfidenceScore: 0.8,
		},
	}

	result := &DataDiscoveryResult{
		DiscoveredFields: fields,
	}

	sortedFields := service.GetDiscoveredFieldsByPriority(result)

	// Should be sorted by priority (ascending) and then by business value (descending)
	assert.Equal(t, "high_priority", sortedFields[0].FieldName)
	assert.Equal(t, "medium_priority", sortedFields[1].FieldName)
	assert.Equal(t, "low_priority", sortedFields[2].FieldName)
}

func TestDataDiscoveryService_GetHighConfidenceFields(t *testing.T) {
	config := DefaultDataDiscoveryConfig()
	config.MinConfidenceThreshold = 0.7
	logger := zap.NewNop()
	service := NewDataDiscoveryService(config, logger)

	// Create test fields with different confidence scores
	fields := []DiscoveredField{
		{
			FieldName:       "high_confidence",
			FieldType:       "email",
			ConfidenceScore: 0.9,
		},
		{
			FieldName:       "medium_confidence",
			FieldType:       "phone",
			ConfidenceScore: 0.7,
		},
		{
			FieldName:       "low_confidence",
			FieldType:       "social_media",
			ConfidenceScore: 0.5,
		},
	}

	result := &DataDiscoveryResult{
		DiscoveredFields: fields,
	}

	highConfidenceFields := service.GetHighConfidenceFields(result)

	// Should only include fields with confidence >= 0.7
	assert.Equal(t, 2, len(highConfidenceFields))
	assert.Equal(t, "high_confidence", highConfidenceFields[0].FieldName)
	assert.Equal(t, "medium_confidence", highConfidenceFields[1].FieldName)
}

func TestDataDiscoveryService_GenerateExtractionPlan(t *testing.T) {
	config := DefaultDataDiscoveryConfig()
	logger := zap.NewNop()
	service := NewDataDiscoveryService(config, logger)

	// Create test fields
	fields := []DiscoveredField{
		{
			FieldName:        "email",
			FieldType:        "email",
			DataType:         "string",
			ConfidenceScore:  0.9,
			ExtractionMethod: "regex",
			Priority:         1,
			BusinessValue:    0.9,
		},
		{
			FieldName:        "phone",
			FieldType:        "phone",
			DataType:         "string",
			ConfidenceScore:  0.8,
			ExtractionMethod: "regex",
			Priority:         1,
			BusinessValue:    0.8,
		},
	}

	result := &DataDiscoveryResult{
		DiscoveredFields: fields,
		ExtractionRules:  []ExtractionRule{},
	}

	plan := service.GenerateExtractionPlan(result)

	assert.NotNil(t, plan)
	assert.NotEmpty(t, plan.PlanID)
	assert.Equal(t, 2, len(plan.Fields))
	assert.Equal(t, "balanced", plan.Strategy)
	assert.Greater(t, plan.EstimatedTime, time.Duration(0))
	assert.Greater(t, len(plan.FieldGroups), 0)
}

func TestDataDiscoveryService_CalculateOverallConfidence(t *testing.T) {
	config := DefaultDataDiscoveryConfig()
	logger := zap.NewNop()
	service := NewDataDiscoveryService(config, logger)

	// Test with fields of different confidence scores
	fields := []DiscoveredField{
		{
			FieldName:       "field1",
			ConfidenceScore: 0.9,
			BusinessValue:   0.8,
			Priority:        1,
		},
		{
			FieldName:       "field2",
			ConfidenceScore: 0.7,
			BusinessValue:   0.6,
			Priority:        2,
		},
	}

	result := &DataDiscoveryResult{
		DiscoveredFields: fields,
		ClassificationResult: &ClassificationResult{
			ConfidenceScore: 0.8,
		},
	}

	confidence := service.calculateOverallConfidence(result)

	assert.Greater(t, confidence, 0.0)
	assert.LessOrEqual(t, confidence, 1.0)
}

func TestDataDiscoveryService_EstimateExtractionTime(t *testing.T) {
	config := DefaultDataDiscoveryConfig()
	logger := zap.NewNop()
	service := NewDataDiscoveryService(config, logger)

	// Test with different field types
	fields := []DiscoveredField{
		{
			FieldName:        "email",
			ExtractionMethod: "regex",
			ConfidenceScore:  0.9,
		},
		{
			FieldName:        "phone",
			ExtractionMethod: "xpath",
			ConfidenceScore:  0.8,
		},
		{
			FieldName:        "address",
			ExtractionMethod: "ml_classification",
			ConfidenceScore:  0.7,
		},
	}

	result := &DataDiscoveryResult{
		DiscoveredFields: fields,
	}

	estimatedTime := service.estimateExtractionTime(result)

	assert.Greater(t, estimatedTime, time.Duration(0))
	// Should be reasonable (less than 10 seconds for 3 fields)
	assert.Less(t, estimatedTime, 10*time.Second)
}

func TestDataDiscoveryService_GroupFieldsByExtractionMethod(t *testing.T) {
	config := DefaultDataDiscoveryConfig()
	logger := zap.NewNop()
	service := NewDataDiscoveryService(config, logger)

	fields := []DiscoveredField{
		{
			FieldName:        "email",
			ExtractionMethod: "regex",
			Priority:         1,
		},
		{
			FieldName:        "phone",
			ExtractionMethod: "regex",
			Priority:         1,
		},
		{
			FieldName:        "address",
			ExtractionMethod: "xpath",
			Priority:         2,
		},
	}

	groups := service.groupFieldsByExtractionMethod(fields)

	assert.Equal(t, 2, len(groups)) // regex and xpath groups

	// Find regex group
	var regexGroup *FieldGroup
	for i := range groups {
		if groups[i].ExtractionMethod == "regex" {
			regexGroup = &groups[i]
			break
		}
	}

	assert.NotNil(t, regexGroup)
	assert.Equal(t, 2, len(regexGroup.Fields))
	assert.Equal(t, "regex", regexGroup.ExtractionMethod)
}

func TestDataDiscoveryService_Integration(t *testing.T) {
	config := DefaultDataDiscoveryConfig()
	logger := zap.NewNop()
	service := NewDataDiscoveryService(config, logger)

	// Test with comprehensive business content
	content := &ContentInput{
		RawContent: `
			Acme Corporation
			Contact: info@acme.com | Phone: (555) 123-4567
			Address: 123 Business Ave, Suite 100, Anytown, NY 12345
			Website: https://www.acme.com
			Follow us: https://facebook.com/acme | https://twitter.com/acme
			Business Hours: Monday-Friday 9:00 AM - 6:00 PM
			Services: Software Development, Consulting, Training
			Founded: 2010
			EIN: 12-3456789
		`,
		ContentType: "html",
		URL:         "https://www.acme.com",
		StructuredData: map[string]interface{}{
			"business_name": "Acme Corporation",
			"email":         "info@acme.com",
		},
	}

	ctx := context.Background()
	result, err := service.DiscoverDataPoints(ctx, content)

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Debug: Print discovered fields
	t.Logf("Discovered %d fields:", len(result.DiscoveredFields))
	for _, field := range result.DiscoveredFields {
		t.Logf("  - %s (%s): confidence=%.2f, priority=%d",
			field.FieldName, field.FieldType, field.ConfidenceScore, field.Priority)
	}

	// Should discover multiple field types
	assert.GreaterOrEqual(t, len(result.DiscoveredFields), 5)

	// Check for specific field types
	fieldTypes := make(map[string]bool)
	for _, field := range result.DiscoveredFields {
		fieldTypes[field.FieldType] = true
	}

	// Debug: Print found field types
	t.Logf("Found field types: %v", fieldTypes)

	expectedFields := []string{"email", "phone", "address", "url", "social_media"}
	for _, expected := range expectedFields {
		assert.True(t, fieldTypes[expected], "Should discover %s field", expected)
	}

	// Verify pattern matches
	assert.Greater(t, len(result.PatternMatches), 0)

	// Verify classification
	assert.NotNil(t, result.ClassificationResult)
	assert.NotEmpty(t, result.ClassificationResult.BusinessType)

	// Verify extraction rules
	assert.Greater(t, len(result.ExtractionRules), 0)

	// Test extraction plan generation
	plan := service.GenerateExtractionPlan(result)
	assert.NotNil(t, plan)
	assert.Greater(t, len(plan.Fields), 0)
	assert.Greater(t, len(plan.FieldGroups), 0)
}

func TestDataDiscoveryConfig_Default(t *testing.T) {
	config := DefaultDataDiscoveryConfig()

	assert.Equal(t, 5, config.MaxDiscoveryDepth)
	assert.Equal(t, 0.7, config.MinConfidenceThreshold)
	assert.Equal(t, 10*time.Second, config.PatternMatchTimeout)
	assert.True(t, config.EnableMLClassification)
	assert.Equal(t, 10, config.MaxPatternsPerField)
	assert.Equal(t, "balanced", config.DiscoveryStrategy)
}

func TestContentInput_Validation(t *testing.T) {
	// Test with minimal content
	content := &ContentInput{
		RawContent:  "Test content",
		ContentType: "text",
	}

	assert.NotEmpty(t, content.RawContent)
	assert.NotEmpty(t, content.ContentType)

	// Test with full content
	fullContent := &ContentInput{
		RawContent:     "Full test content",
		ContentType:    "html",
		URL:            "https://example.com",
		HTMLContent:    "<html><body>Test</body></html>",
		StructuredData: map[string]interface{}{"test": "value"},
		MetaData:       map[string]string{"title": "Test"},
		Language:       "en",
		Encoding:       "UTF-8",
	}

	assert.NotEmpty(t, fullContent.RawContent)
	assert.NotEmpty(t, fullContent.ContentType)
	assert.NotEmpty(t, fullContent.URL)
	assert.NotEmpty(t, fullContent.HTMLContent)
	assert.NotNil(t, fullContent.StructuredData)
	assert.NotNil(t, fullContent.MetaData)
	assert.NotEmpty(t, fullContent.Language)
	assert.NotEmpty(t, fullContent.Encoding)
}
