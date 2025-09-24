package classification_monitoring

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestIndustryAccuracyMonitor_Creation(t *testing.T) {
	config := DefaultIndustryAccuracyConfig()
	logger := zap.NewNop()

	monitor := NewIndustryAccuracyMonitor(config, logger)

	if monitor == nil {
		t.Fatal("Expected monitor to be created")
	}

	if monitor.config == nil {
		t.Fatal("Expected config to be set")
	}

	if monitor.logger == nil {
		t.Fatal("Expected logger to be set")
	}

	if monitor.industryMetrics == nil {
		t.Fatal("Expected industry metrics map to be initialized")
	}

	if monitor.industryRankings == nil {
		t.Fatal("Expected industry rankings to be initialized")
	}

	if monitor.industryTrends == nil {
		t.Fatal("Expected industry trends to be initialized")
	}
}

func TestIndustryAccuracyMonitor_TrackClassification(t *testing.T) {
	config := DefaultIndustryAccuracyConfig()
	logger := zap.NewNop()

	monitor := NewIndustryAccuracyMonitor(config, logger)

	// Create test classification result
	result := &ClassificationResult{
		ID:                     "test_123",
		BusinessName:           "Test Restaurant",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "keyword_classification",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"industry": "restaurant",
		},
		IsCorrect: boolPtr(true),
	}

	// Track classification
	err := monitor.TrackClassification(context.Background(), result)
	if err != nil {
		t.Fatalf("Expected tracking to succeed, got error: %v", err)
	}

	// Verify industry metrics
	metrics := monitor.GetIndustryMetrics("restaurant")
	if metrics == nil {
		t.Fatal("Expected restaurant metrics to be available")
	}

	if metrics.IndustryName != "restaurant" {
		t.Errorf("Expected industry name to be 'restaurant', got '%s'", metrics.IndustryName)
	}

	if metrics.TotalClassifications != 1 {
		t.Errorf("Expected total classifications to be 1, got %d", metrics.TotalClassifications)
	}

	if metrics.CorrectClassifications != 1 {
		t.Errorf("Expected correct classifications to be 1, got %d", metrics.CorrectClassifications)
	}

	if metrics.AccuracyScore != 1.0 {
		t.Errorf("Expected accuracy score to be 1.0, got %f", metrics.AccuracyScore)
	}

	if metrics.ConfidenceScore != 0.95 {
		t.Errorf("Expected confidence score to be 0.95, got %f", metrics.ConfidenceScore)
	}
}

func TestIndustryAccuracyMonitor_MultipleIndustries(t *testing.T) {
	config := DefaultIndustryAccuracyConfig()
	logger := zap.NewNop()

	monitor := NewIndustryAccuracyMonitor(config, logger)

	// Test data for multiple industries
	industries := []struct {
		name       string
		correct    bool
		confidence float64
	}{
		{"restaurant", true, 0.95},
		{"technology", true, 0.90},
		{"healthcare", false, 0.85},
		{"retail", true, 0.88},
	}

	// Track classifications for each industry
	for i, industry := range industries {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_%d", i),
			BusinessName:           fmt.Sprintf("Test %s Business", industry.name),
			ActualClassification:   industry.name,
			ExpectedClassification: stringPtr(industry.name),
			ConfidenceScore:        industry.confidence,
			ClassificationMethod:   "keyword_classification",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry": industry.name,
			},
			IsCorrect: boolPtr(industry.correct),
		}

		err := monitor.TrackClassification(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track classification for %s: %v", industry.name, err)
		}
	}

	// Verify all industries are tracked
	allMetrics := monitor.GetAllIndustryMetrics()
	if len(allMetrics) != 4 {
		t.Errorf("Expected 4 industries to be tracked, got %d", len(allMetrics))
	}

	// Verify specific industry metrics
	restaurantMetrics := monitor.GetIndustryMetrics("restaurant")
	if restaurantMetrics == nil {
		t.Fatal("Expected restaurant metrics to be available")
	}

	if restaurantMetrics.AccuracyScore != 1.0 {
		t.Errorf("Expected restaurant accuracy to be 1.0, got %f", restaurantMetrics.AccuracyScore)
	}

	healthcareMetrics := monitor.GetIndustryMetrics("healthcare")
	if healthcareMetrics == nil {
		t.Fatal("Expected healthcare metrics to be available")
	}

	if healthcareMetrics.AccuracyScore != 0.0 {
		t.Errorf("Expected healthcare accuracy to be 0.0, got %f", healthcareMetrics.AccuracyScore)
	}
}

func TestIndustryAccuracyMonitor_MethodPerformance(t *testing.T) {
	config := DefaultIndustryAccuracyConfig()
	logger := zap.NewNop()

	monitor := NewIndustryAccuracyMonitor(config, logger)

	// Test data for different methods within the same industry
	methods := []struct {
		name       string
		correct    bool
		confidence float64
	}{
		{"keyword_classification", true, 0.95},
		{"ml_classification", true, 0.90},
		{"description_analysis", false, 0.85},
	}

	// Track classifications for each method
	for i, method := range methods {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_method_%d", i),
			BusinessName:           fmt.Sprintf("Test Restaurant %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        method.confidence,
			ClassificationMethod:   method.name,
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry": "restaurant",
			},
			IsCorrect: boolPtr(method.correct),
		}

		err := monitor.TrackClassification(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track classification for method %s: %v", method.name, err)
		}
	}

	// Verify method performance tracking
	metrics := monitor.GetIndustryMetrics("restaurant")
	if metrics == nil {
		t.Fatal("Expected restaurant metrics to be available")
	}

	if len(metrics.MethodPerformance) != 3 {
		t.Errorf("Expected 3 methods to be tracked, got %d", len(metrics.MethodPerformance))
	}

	// Verify specific method performance
	keywordMethod := metrics.MethodPerformance["keyword_classification"]
	if keywordMethod == nil {
		t.Fatal("Expected keyword classification method to be tracked")
	}

	if keywordMethod.AccuracyScore != 1.0 {
		t.Errorf("Expected keyword method accuracy to be 1.0, got %f", keywordMethod.AccuracyScore)
	}

	descriptionMethod := metrics.MethodPerformance["description_analysis"]
	if descriptionMethod == nil {
		t.Fatal("Expected description analysis method to be tracked")
	}

	if descriptionMethod.AccuracyScore != 0.0 {
		t.Errorf("Expected description method accuracy to be 0.0, got %f", descriptionMethod.AccuracyScore)
	}
}

func TestIndustryAccuracyMonitor_ConfidenceDistribution(t *testing.T) {
	config := DefaultIndustryAccuracyConfig()
	logger := zap.NewNop()

	monitor := NewIndustryAccuracyMonitor(config, logger)

	// Test data with different confidence levels
	confidenceLevels := []float64{0.95, 0.85, 0.75, 0.65, 0.55}

	for i, confidence := range confidenceLevels {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_confidence_%d", i),
			BusinessName:           fmt.Sprintf("Test Restaurant %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        confidence,
			ClassificationMethod:   "keyword_classification",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry": "restaurant",
			},
			IsCorrect: boolPtr(true),
		}

		err := monitor.TrackClassification(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track classification: %v", err)
		}
	}

	// Verify confidence distribution
	metrics := monitor.GetIndustryMetrics("restaurant")
	if metrics == nil {
		t.Fatal("Expected restaurant metrics to be available")
	}

	// Check that confidence distribution is populated
	if len(metrics.ConfidenceDistribution) == 0 {
		t.Error("Expected confidence distribution to be populated")
	}

	// Verify specific confidence ranges
	highConfidence := metrics.ConfidenceDistribution["high"]
	if highConfidence != 1 {
		t.Errorf("Expected 1 high confidence classification, got %d", highConfidence)
	}

	mediumConfidence := metrics.ConfidenceDistribution["medium"]
	if mediumConfidence != 2 {
		t.Errorf("Expected 2 medium confidence classifications, got %d", mediumConfidence)
	}

	lowConfidence := metrics.ConfidenceDistribution["low"]
	if lowConfidence != 2 {
		t.Errorf("Expected 2 low confidence classifications, got %d", lowConfidence)
	}
}

func TestIndustryAccuracyMonitor_PerformanceIndicators(t *testing.T) {
	config := DefaultIndustryAccuracyConfig()
	logger := zap.NewNop()

	monitor := NewIndustryAccuracyMonitor(config, logger)

	// Add multiple results to generate performance indicators
	for i := 0; i < 15; i++ {
		accuracy := 0.8 + float64(i)*0.01 // Improving accuracy
		isCorrect := accuracy > 0.85      // Most results are correct

		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_performance_%d", i),
			BusinessName:           fmt.Sprintf("Test Restaurant %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        accuracy,
			ClassificationMethod:   "keyword_classification",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry": "restaurant",
			},
			IsCorrect: boolPtr(isCorrect),
		}

		err := monitor.TrackClassification(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track classification: %v", err)
		}
	}

	// Verify performance indicators
	metrics := monitor.GetIndustryMetrics("restaurant")
	if metrics == nil {
		t.Fatal("Expected restaurant metrics to be available")
	}

	// Check performance grade
	if metrics.PerformanceGrade == "" {
		t.Error("Expected performance grade to be set")
	}

	// Check reliability score
	if metrics.ReliabilityScore == 0.0 {
		t.Error("Expected reliability score to be calculated")
	}

	// Check consistency score
	if metrics.ConsistencyScore == 0.0 {
		t.Error("Expected consistency score to be calculated")
	}

	// Check data quality score
	if metrics.DataQualityScore == 0.0 {
		t.Error("Expected data quality score to be calculated")
	}
}

func TestIndustryAccuracyMonitor_TimeBasedMetrics(t *testing.T) {
	config := DefaultIndustryAccuracyConfig()
	logger := zap.NewNop()

	monitor := NewIndustryAccuracyMonitor(config, logger)

	// Add results at different times
	baseTime := time.Now()

	for i := 0; i < 5; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_time_%d", i),
			BusinessName:           fmt.Sprintf("Test Restaurant %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "keyword_classification",
			Timestamp:              baseTime.Add(time.Duration(i) * time.Hour),
			Metadata: map[string]interface{}{
				"industry": "restaurant",
			},
			IsCorrect: boolPtr(true),
		}

		err := monitor.TrackClassification(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track classification: %v", err)
		}
	}

	// Verify time-based metrics
	metrics := monitor.GetIndustryMetrics("restaurant")
	if metrics == nil {
		t.Fatal("Expected restaurant metrics to be available")
	}

	// Check hourly accuracy
	if len(metrics.HourlyAccuracy) == 0 {
		t.Error("Expected hourly accuracy to be populated")
	}

	// Check daily accuracy
	if len(metrics.DailyAccuracy) == 0 {
		t.Error("Expected daily accuracy to be populated")
	}

	// Check weekly accuracy
	if len(metrics.WeeklyAccuracy) == 0 {
		t.Error("Expected weekly accuracy to be populated")
	}
}

func TestIndustryAccuracyMonitor_HistoricalData(t *testing.T) {
	config := DefaultIndustryAccuracyConfig()
	logger := zap.NewNop()

	monitor := NewIndustryAccuracyMonitor(config, logger)

	// Add multiple results to generate historical data
	for i := 0; i < 10; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_historical_%d", i),
			BusinessName:           fmt.Sprintf("Test Restaurant %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "keyword_classification",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry": "restaurant",
			},
			IsCorrect: boolPtr(true),
		}

		err := monitor.TrackClassification(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track classification: %v", err)
		}
	}

	// Verify historical data
	metrics := monitor.GetIndustryMetrics("restaurant")
	if metrics == nil {
		t.Fatal("Expected restaurant metrics to be available")
	}

	// Check historical accuracy
	if len(metrics.HistoricalAccuracy) == 0 {
		t.Error("Expected historical accuracy data to be populated")
	}

	// Check historical confidence
	if len(metrics.HistoricalConfidence) == 0 {
		t.Error("Expected historical confidence data to be populated")
	}

	// Verify data points have correct structure
	if len(metrics.HistoricalAccuracy) > 0 {
		point := metrics.HistoricalAccuracy[0]
		if point.Accuracy != 1.0 {
			t.Errorf("Expected accuracy point to be 1.0, got %f", point.Accuracy)
		}
		if point.SampleSize != 1 {
			t.Errorf("Expected sample size to be 1, got %d", point.SampleSize)
		}
	}
}

func TestIndustryAccuracyMonitor_PerformanceSummary(t *testing.T) {
	config := DefaultIndustryAccuracyConfig()
	logger := zap.NewNop()

	monitor := NewIndustryAccuracyMonitor(config, logger)

	// Add results to generate performance data
	for i := 0; i < 10; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_summary_%d", i),
			BusinessName:           fmt.Sprintf("Test Restaurant %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "keyword_classification",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry": "restaurant",
			},
			IsCorrect: boolPtr(true),
		}

		err := monitor.TrackClassification(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track classification: %v", err)
		}
	}

	// Get performance summary
	summary := monitor.GetIndustryPerformanceSummary("restaurant")
	if summary == nil {
		t.Fatal("Expected performance summary to be available")
	}

	// Verify summary fields
	if summary.IndustryName != "restaurant" {
		t.Errorf("Expected industry name to be 'restaurant', got '%s'", summary.IndustryName)
	}

	if summary.AccuracyScore != 1.0 {
		t.Errorf("Expected accuracy score to be 1.0, got %f", summary.AccuracyScore)
	}

	if summary.ConfidenceScore != 0.95 {
		t.Errorf("Expected confidence score to be 0.95, got %f", summary.ConfidenceScore)
	}

	if summary.TotalClassifications != 10 {
		t.Errorf("Expected total classifications to be 10, got %d", summary.TotalClassifications)
	}

	if summary.PerformanceGrade == "" {
		t.Error("Expected performance grade to be set")
	}

	if summary.ReliabilityScore == 0.0 {
		t.Error("Expected reliability score to be calculated")
	}

	if summary.ConsistencyScore == 0.0 {
		t.Error("Expected consistency score to be calculated")
	}

	if summary.DataQualityScore == 0.0 {
		t.Error("Expected data quality score to be calculated")
	}
}
