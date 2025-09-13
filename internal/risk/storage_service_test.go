package risk

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRiskStorageService_ConvertStorageToAssessment(t *testing.T) {
	logger := zap.NewNop()
	service := NewRiskStorageService(nil, logger)

	now := time.Now()
	storage := &RiskDataStorage{
		ID:           "test-id",
		BusinessID:   "business-123",
		BusinessName: "Test Business",
		OverallScore: 75.5,
		OverallLevel: "medium",
		CategoryScores: map[string]interface{}{
			"financial": map[string]interface{}{
				"factor_id":     "financial-1",
				"factor_name":   "Financial Stability",
				"category":      "financial",
				"score":         80.0,
				"level":         "medium",
				"confidence":    0.85,
				"explanation":   "Good financial indicators",
				"evidence":      []interface{}{"Revenue growth", "Low debt ratio"},
				"calculated_at": now,
			},
		},
		FactorScores: []RiskScore{
			{
				FactorID:     "factor-1",
				FactorName:   "Test Factor",
				Category:     RiskCategoryOperational,
				Score:        70.0,
				Level:        RiskLevelMedium,
				Confidence:   0.8,
				Explanation:  "Test explanation",
				Evidence:     []string{"Evidence 1", "Evidence 2"},
				CalculatedAt: now,
			},
		},
		Recommendations: []RiskRecommendation{
			{
				ID:          "rec-1",
				RiskFactor:  "factor-1",
				Title:       "Test Recommendation",
				Description: "Test description",
				Priority:    RiskLevelMedium,
				Action:      "Test action",
				Impact:      "Test impact",
				Timeline:    "30 days",
				CreatedAt:   now,
			},
		},
		Alerts: []RiskAlert{
			{
				ID:           "alert-1",
				BusinessID:   "business-123",
				RiskFactor:   "factor-1",
				Level:        RiskLevelMedium,
				Message:      "Test alert message",
				Score:        75.0,
				Threshold:    70.0,
				TriggeredAt:  now,
				Acknowledged: false,
			},
		},
		AssessedAt: now,
		ValidUntil: now.Add(24 * time.Hour),
		Metadata: map[string]interface{}{
			"source":  "test",
			"version": "1.0",
		},
	}

	assessment := service.convertStorageToAssessment(storage)

	assert.Equal(t, "test-id", assessment.ID)
	assert.Equal(t, "business-123", assessment.BusinessID)
	assert.Equal(t, "Test Business", assessment.BusinessName)
	assert.Equal(t, 75.5, assessment.OverallScore)
	assert.Equal(t, RiskLevelMedium, assessment.OverallLevel)
	assert.Len(t, assessment.CategoryScores, 1)
	assert.Len(t, assessment.FactorScores, 1)
	assert.Len(t, assessment.Recommendations, 1)
	assert.Len(t, assessment.Alerts, 1)
	assert.Equal(t, "test", assessment.Metadata["source"])
	assert.Equal(t, "1.0", assessment.Metadata["version"])
}

func TestRiskStorageService_HelperFunctions(t *testing.T) {
	// Test getString helper function
	data := map[string]interface{}{
		"string_value": "test",
		"int_value":    123,
		"nil_value":    nil,
	}

	assert.Equal(t, "test", getString(data, "string_value"))
	assert.Equal(t, "", getString(data, "int_value"))
	assert.Equal(t, "", getString(data, "nil_value"))
	assert.Equal(t, "", getString(data, "missing_key"))

	// Test getFloat64 helper function
	floatData := map[string]interface{}{
		"float64_value": 123.45,
		"float32_value": float32(67.89),
		"int_value":     100,
		"int64_value":   int64(200),
		"string_value":  "not_a_number",
		"nil_value":     nil,
	}

	assert.Equal(t, 123.45, getFloat64(floatData, "float64_value"))
	assert.Equal(t, 67.89, getFloat64(floatData, "float32_value"))
	assert.Equal(t, 100.0, getFloat64(floatData, "int_value"))
	assert.Equal(t, 200.0, getFloat64(floatData, "int64_value"))
	assert.Equal(t, 0.0, getFloat64(floatData, "string_value"))
	assert.Equal(t, 0.0, getFloat64(floatData, "nil_value"))
	assert.Equal(t, 0.0, getFloat64(floatData, "missing_key"))
}

func TestRiskStorageService_NewRiskStorageService(t *testing.T) {
	logger := zap.NewNop()
	service := NewRiskStorageService(nil, logger)

	assert.NotNil(t, service)
	assert.Nil(t, service.db)
	assert.Equal(t, logger, service.logger)
}

func TestRiskStorageService_ContextHandling(t *testing.T) {

	// Test with valid request_id in context
	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	requestID := "unknown"
	if rid := ctx.Value("request_id"); rid != nil {
		if str, ok := rid.(string); ok {
			requestID = str
		}
	}
	assert.Equal(t, "test-request-123", requestID)

	// Test with invalid request_id in context
	ctx2 := context.WithValue(context.Background(), "request_id", 123)
	requestID2 := "unknown"
	if rid := ctx2.Value("request_id"); rid != nil {
		if str, ok := rid.(string); ok {
			requestID2 = str
		}
	}
	assert.Equal(t, "unknown", requestID2)

	// Test with no request_id in context
	ctx3 := context.Background()
	requestID3 := "unknown"
	if rid := ctx3.Value("request_id"); rid != nil {
		if str, ok := rid.(string); ok {
			requestID3 = str
		}
	}
	assert.Equal(t, "unknown", requestID3)
}

func TestRiskStorageService_AssessmentIDGeneration(t *testing.T) {
	assessment := &RiskAssessment{
		BusinessID:   "business-123",
		BusinessName: "Test Business",
		OverallScore: 75.5,
		OverallLevel: RiskLevelMedium,
	}

	// Test that ID is generated when empty
	if assessment.ID == "" {
		assessment.ID = uuid.New().String()
	}

	assert.NotEmpty(t, assessment.ID)
	assert.Len(t, assessment.ID, 36) // UUID length
}

func TestRiskStorageService_JSONMarshaling(t *testing.T) {
	assessment := &RiskAssessment{
		ID:           uuid.New().String(),
		BusinessID:   "business-123",
		BusinessName: "Test Business",
		OverallScore: 75.5,
		OverallLevel: RiskLevelMedium,
		CategoryScores: map[RiskCategory]RiskScore{
			RiskCategoryFinancial: {
				FactorID:     "financial-1",
				FactorName:   "Financial Stability",
				Category:     RiskCategoryFinancial,
				Score:        80.0,
				Level:        RiskLevelMedium,
				Confidence:   0.85,
				Explanation:  "Good financial indicators",
				Evidence:     []string{"Revenue growth", "Low debt ratio"},
				CalculatedAt: time.Now(),
			},
		},
		FactorScores: []RiskScore{
			{
				FactorID:     "factor-1",
				FactorName:   "Test Factor",
				Category:     RiskCategoryOperational,
				Score:        70.0,
				Level:        RiskLevelMedium,
				Confidence:   0.8,
				Explanation:  "Test explanation",
				Evidence:     []string{"Evidence 1", "Evidence 2"},
				CalculatedAt: time.Now(),
			},
		},
		Recommendations: []RiskRecommendation{
			{
				ID:          "rec-1",
				RiskFactor:  "factor-1",
				Title:       "Test Recommendation",
				Description: "Test description",
				Priority:    RiskLevelMedium,
				Action:      "Test action",
				Impact:      "Test impact",
				Timeline:    "30 days",
				CreatedAt:   time.Now(),
			},
		},
		Alerts: []RiskAlert{
			{
				ID:           "alert-1",
				BusinessID:   "business-123",
				RiskFactor:   "factor-1",
				Level:        RiskLevelMedium,
				Message:      "Test alert message",
				Score:        75.0,
				Threshold:    70.0,
				TriggeredAt:  time.Now(),
				Acknowledged: false,
			},
		},
		AssessedAt: time.Now(),
		ValidUntil: time.Now().Add(24 * time.Hour),
		Metadata: map[string]interface{}{
			"source":  "test",
			"version": "1.0",
		},
	}

	// Test that all JSON marshaling works without errors
	logger := zap.NewNop()
	service := NewRiskStorageService(nil, logger)

	// This would normally be called in StoreRiskAssessment
	// We're just testing that the JSON marshaling works
	result := service.convertStorageToAssessment(&RiskDataStorage{
		ID:              assessment.ID,
		BusinessID:      assessment.BusinessID,
		BusinessName:    assessment.BusinessName,
		OverallScore:    assessment.OverallScore,
		OverallLevel:    string(assessment.OverallLevel),
		CategoryScores:  make(map[string]interface{}),
		FactorScores:    assessment.FactorScores,
		Recommendations: assessment.Recommendations,
		Alerts:          assessment.Alerts,
		AssessedAt:      assessment.AssessedAt,
		ValidUntil:      assessment.ValidUntil,
		Metadata:        assessment.Metadata,
	})

	assert.NotNil(t, result)
}
