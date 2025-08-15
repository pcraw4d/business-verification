package classification

import (
	"context"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

func TestNewDynamicConfidenceAdjuster(t *testing.T) {
	logger := observability.NewLogger("test")
	metrics := observability.NewMetrics()

	adjuster := NewDynamicConfidenceAdjuster(logger, metrics)

	if adjuster == nil {
		t.Fatal("Expected non-nil adjuster")
	}

	if adjuster.logger != logger {
		t.Error("Expected logger to be set")
	}

	if adjuster.metrics != metrics {
		t.Error("Expected metrics to be set")
	}

	// Check default weights
	if adjuster.contentQualityWeight != 0.25 {
		t.Errorf("Expected contentQualityWeight to be 0.25, got %f", adjuster.contentQualityWeight)
	}

	if adjuster.geographicRegionWeight != 0.20 {
		t.Errorf("Expected geographicRegionWeight to be 0.20, got %f", adjuster.geographicRegionWeight)
	}
}

func TestAdjustConfidence(t *testing.T) {
	logger := observability.NewLogger("test")
	metrics := observability.NewMetrics()
	adjuster := NewDynamicConfidenceAdjuster(logger, metrics)

	ctx := context.Background()
	baseConfidence := 0.8

	// Test with nil factors (should use defaults)
	adjusted := adjuster.AdjustConfidence(ctx, baseConfidence, nil)
	if adjusted < 0.0 || adjusted > 1.0 {
		t.Errorf("Expected adjusted confidence to be between 0.0 and 1.0, got %f", adjusted)
	}

	// Test with complete factors
	factors := &ConfidenceAdjustmentFactors{
		ContentQuality: &ContentQuality{
			Completeness:      0.9,
			Relevance:         0.9,
			Freshness:         0.8,
			Accuracy:          0.9,
			Consistency:       0.8,
			SourceReliability: 0.9,
		},
		GeographicRegion: &GeographicRegion{
			Country:     "US",
			State:       "CA",
			City:        "San Francisco",
			Region:      "West Coast",
			Confidence:  0.9,
			DataQuality: 0.9,
		},
		IndustryFactors: &IndustrySpecificFactors{
			IndustryCode:     "51",
			IndustryCategory: "Information",
			CodeDensity:      0.3,
			ValidationRate:   0.9,
			Popularity:       0.7,
			Complexity:       0.4,
		},
		BusinessSize:      "medium",
		BusinessAge:       5,
		DataSourceQuality: 0.9,
		CrossValidation:   0.8,
		LastUpdated:       time.Now(),
	}

	adjusted = adjuster.AdjustConfidence(ctx, baseConfidence, factors)
	if adjusted < 0.0 || adjusted > 1.0 {
		t.Errorf("Expected adjusted confidence to be between 0.0 and 1.0, got %f", adjusted)
	}

	// Test that adjustment is reasonable (not too extreme)
	if adjusted < 0.4 || adjusted > 1.0 {
		t.Errorf("Expected reasonable adjustment, got %f", adjusted)
	}
}

func TestCalculateContentQualityAdjustment(t *testing.T) {
	adjuster := &DynamicConfidenceAdjuster{}

	// Test with nil quality
	adjustment := adjuster.calculateContentQualityAdjustment(nil)
	if adjustment != 0.75 {
		t.Errorf("Expected default adjustment of 0.75 for nil quality, got %f", adjustment)
	}

	// Test with high quality
	highQuality := &ContentQuality{
		Completeness:      1.0,
		Relevance:         1.0,
		Freshness:         1.0,
		Accuracy:          1.0,
		Consistency:       1.0,
		SourceReliability: 1.0,
	}
	adjustment = adjuster.calculateContentQualityAdjustment(highQuality)
	if adjustment < 1.4 || adjustment > 1.5 {
		t.Errorf("Expected high quality adjustment to be around 1.5, got %f", adjustment)
	}

	// Test with low quality
	lowQuality := &ContentQuality{
		Completeness:      0.2,
		Relevance:         0.2,
		Freshness:         0.2,
		Accuracy:          0.2,
		Consistency:       0.2,
		SourceReliability: 0.2,
	}
	adjustment = adjuster.calculateContentQualityAdjustment(lowQuality)
	if adjustment < 0.5 || adjustment > 0.7 {
		t.Errorf("Expected low quality adjustment to be around 0.6, got %f", adjustment)
	}
}

func TestCalculateGeographicAdjustment(t *testing.T) {
	adjuster := &DynamicConfidenceAdjuster{}
	adjuster.initializeDefaultValues()

	// Test with nil region
	adjustment := adjuster.calculateGeographicAdjustment(nil)
	if adjustment != 0.85 {
		t.Errorf("Expected default adjustment of 0.85 for nil region, got %f", adjustment)
	}

	// Test with US region (baseline)
	usRegion := &GeographicRegion{
		Country:     "US",
		State:       "CA",
		City:        "San Francisco",
		Confidence:  0.9,
		DataQuality: 0.9,
	}
	adjustment = adjuster.calculateGeographicAdjustment(usRegion)
	if adjustment < 0.8 || adjustment > 1.2 {
		t.Errorf("Expected US region adjustment to be reasonable, got %f", adjustment)
	}

	// Test with unknown country
	unknownRegion := &GeographicRegion{
		Country:     "XX",
		Confidence:  0.8,
		DataQuality: 0.8,
	}
	adjustment = adjuster.calculateGeographicAdjustment(unknownRegion)
	if adjustment < 0.5 || adjustment > 1.5 {
		t.Errorf("Expected unknown region adjustment to be reasonable, got %f", adjustment)
	}
}

func TestCalculateIndustryAdjustment(t *testing.T) {
	adjuster := &DynamicConfidenceAdjuster{}
	adjuster.initializeDefaultValues()

	// Test with nil factors
	adjustment := adjuster.calculateIndustryAdjustment(nil)
	if adjustment != 0.90 {
		t.Errorf("Expected default adjustment of 0.90 for nil factors, got %f", adjustment)
	}

	// Test with known industry (Finance and Insurance)
	financeFactors := &IndustrySpecificFactors{
		IndustryCode:     "52",
		IndustryCategory: "Finance and Insurance",
		CodeDensity:      0.3,
		ValidationRate:   0.9,
		Popularity:       0.8,
		Complexity:       0.4,
	}
	adjustment = adjuster.calculateIndustryAdjustment(financeFactors)
	if adjustment < 0.5 || adjustment > 1.5 {
		t.Errorf("Expected finance industry adjustment to be reasonable, got %f", adjustment)
	}

	// Test with unknown industry
	unknownFactors := &IndustrySpecificFactors{
		IndustryCode:     "99",
		IndustryCategory: "Unknown",
		CodeDensity:      0.5,
		ValidationRate:   0.8,
		Popularity:       0.5,
		Complexity:       0.5,
	}
	adjustment = adjuster.calculateIndustryAdjustment(unknownFactors)
	if adjustment < 0.5 || adjustment > 1.5 {
		t.Errorf("Expected unknown industry adjustment to be reasonable, got %f", adjustment)
	}
}

func TestCalculateBusinessSizeAdjustment(t *testing.T) {
	adjuster := &DynamicConfidenceAdjuster{}
	adjuster.initializeDefaultValues()

	// Test with empty size
	adjustment := adjuster.calculateBusinessSizeAdjustment("")
	if adjustment < 0.5 || adjustment > 1.5 {
		t.Errorf("Expected empty size adjustment to be reasonable, got %f", adjustment)
	}

	// Test with known sizes
	testCases := []struct {
		size        string
		expectedMin float64
		expectedMax float64
	}{
		{"large", 0.9, 1.3},
		{"medium", 0.8, 1.2},
		{"small", 0.7, 1.1},
		{"micro", 0.6, 1.0},
		{"startup", 0.5, 0.9},
		{"unknown", 0.5, 0.9},
	}

	for _, tc := range testCases {
		adjustment := adjuster.calculateBusinessSizeAdjustment(tc.size)
		if adjustment < tc.expectedMin || adjustment > tc.expectedMax {
			t.Errorf("Expected %s size adjustment to be between %f and %f, got %f",
				tc.size, tc.expectedMin, tc.expectedMax, adjustment)
		}
	}
}

func TestCalculateBusinessAgeAdjustment(t *testing.T) {
	adjuster := &DynamicConfidenceAdjuster{}

	// Test with zero age
	adjustment := adjuster.calculateBusinessAgeAdjustment(0)
	if adjustment != 0.80 {
		t.Errorf("Expected zero age adjustment to be 0.80, got %f", adjustment)
	}

	// Test with negative age
	adjustment = adjuster.calculateBusinessAgeAdjustment(-1)
	if adjustment != 0.80 {
		t.Errorf("Expected negative age adjustment to be 0.80, got %f", adjustment)
	}

	// Test with different age ranges
	testCases := []struct {
		age       int
		expected  float64
		tolerance float64
	}{
		{1, 0.85, 0.01},
		{5, 1.0, 0.01},
		{15, 1.05, 0.01},
	}

	for _, tc := range testCases {
		adjustment := adjuster.calculateBusinessAgeAdjustment(tc.age)
		if adjustment < tc.expected-tc.tolerance || adjustment > tc.expected+tc.tolerance {
			t.Errorf("Expected age %d adjustment to be around %f, got %f",
				tc.age, tc.expected, adjustment)
		}
	}
}

func TestCalculateDataSourceAdjustment(t *testing.T) {
	adjuster := &DynamicConfidenceAdjuster{}

	// Test with different data source qualities
	testCases := []struct {
		quality     float64
		expectedMin float64
		expectedMax float64
	}{
		{0.0, 0.5, 0.7},
		{0.5, 0.9, 1.1},
		{1.0, 1.3, 1.5},
	}

	for _, tc := range testCases {
		adjustment := adjuster.calculateDataSourceAdjustment(tc.quality)
		if adjustment < tc.expectedMin || adjustment > tc.expectedMax {
			t.Errorf("Expected quality %f adjustment to be between %f and %f, got %f",
				tc.quality, tc.expectedMin, tc.expectedMax, adjustment)
		}
	}
}

func TestCalculateCrossValidationAdjustment(t *testing.T) {
	adjuster := &DynamicConfidenceAdjuster{}

	// Test with different cross-validation scores
	testCases := []struct {
		score       float64
		expectedMin float64
		expectedMax float64
	}{
		{0.0, 0.7, 0.9},
		{0.5, 0.9, 1.1},
		{1.0, 1.1, 1.3},
	}

	for _, tc := range testCases {
		adjustment := adjuster.calculateCrossValidationAdjustment(tc.score)
		if adjustment < tc.expectedMin || adjustment > tc.expectedMax {
			t.Errorf("Expected score %f adjustment to be between %f and %f, got %f",
				tc.score, tc.expectedMin, tc.expectedMax, adjustment)
		}
	}
}

func TestCalculateRecencyAdjustment(t *testing.T) {
	adjuster := &DynamicConfidenceAdjuster{}

	// Test with zero time
	adjustment := adjuster.calculateRecencyAdjustment(time.Time{})
	if adjustment != 0.85 {
		t.Errorf("Expected zero time adjustment to be 0.85, got %f", adjustment)
	}

	// Test with recent time
	recentTime := time.Now().Add(-15 * 24 * time.Hour) // 15 days ago
	adjustment = adjuster.calculateRecencyAdjustment(recentTime)
	if adjustment < 1.0 || adjustment > 1.1 {
		t.Errorf("Expected recent time adjustment to be around 1.05, got %f", adjustment)
	}

	// Test with old time
	oldTime := time.Now().Add(-400 * 24 * time.Hour) // 400 days ago
	adjustment = adjuster.calculateRecencyAdjustment(oldTime)
	if adjustment < 0.8 || adjustment > 0.9 {
		t.Errorf("Expected old time adjustment to be around 0.85, got %f", adjustment)
	}
}

func TestUpdateMethods(t *testing.T) {
	adjuster := &DynamicConfidenceAdjuster{}
	adjuster.initializeDefaultValues()

	// Test updating geographic modifier
	adjuster.UpdateGeographicConfidenceModifier("TEST", 1.2)
	modifier, exists := adjuster.geographicConfidenceModifiers["TEST"]
	if !exists {
		t.Error("Expected TEST modifier to exist after update")
	}
	if modifier != 1.2 {
		t.Errorf("Expected TEST modifier to be 1.2, got %f", modifier)
	}

	// Test updating industry adjustment
	adjuster.UpdateIndustryConfidenceAdjustment("TEST", 0.8)
	adjustment, exists := adjuster.industryConfidenceAdjustments["TEST"]
	if !exists {
		t.Error("Expected TEST adjustment to exist after update")
	}
	if adjustment != 0.8 {
		t.Errorf("Expected TEST adjustment to be 0.8, got %f", adjustment)
	}

	// Test updating business size modifier
	adjuster.UpdateBusinessSizeModifier("TEST", 0.9)
	sizeModifier, exists := adjuster.businessSizeModifiers["test"] // Should be lowercase
	if !exists {
		t.Error("Expected test modifier to exist after update")
	}
	if sizeModifier != 0.9 {
		t.Errorf("Expected test modifier to be 0.9, got %f", sizeModifier)
	}
}

func TestGetAdjustmentFactors(t *testing.T) {
	adjuster := &DynamicConfidenceAdjuster{}
	adjuster.initializeDefaultValues()

	factors := adjuster.GetAdjustmentFactors()

	// Check that all expected keys exist
	expectedKeys := []string{"geographic_modifiers", "industry_adjustments", "business_size_modifiers", "data_source_thresholds", "weights"}
	for _, key := range expectedKeys {
		if _, exists := factors[key]; !exists {
			t.Errorf("Expected key %s to exist in adjustment factors", key)
		}
	}

	// Check weights structure
	weights, ok := factors["weights"].(map[string]float64)
	if !ok {
		t.Fatal("Expected weights to be map[string]float64")
	}

	expectedWeightKeys := []string{"content_quality", "geographic_region", "industry_specific", "business_size", "business_age", "data_source_quality", "cross_validation", "recency"}
	for _, key := range expectedWeightKeys {
		if _, exists := weights[key]; !exists {
			t.Errorf("Expected weight key %s to exist", key)
		}
	}
}
