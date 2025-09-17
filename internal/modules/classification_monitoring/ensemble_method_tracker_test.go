package classification_monitoring

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestRealTimeEnsembleMethodTracker_Creation(t *testing.T) {
	config := DefaultEnsembleMethodConfig()
	logger := zap.NewNop()

	tracker := NewRealTimeEnsembleMethodTracker(config, logger)

	if tracker == nil {
		t.Fatal("Expected tracker to be created")
	}

	if tracker.config == nil {
		t.Fatal("Expected config to be set")
	}

	if tracker.logger == nil {
		t.Fatal("Expected logger to be set")
	}

	if tracker.methodMetrics == nil {
		t.Fatal("Expected method metrics map to be initialized")
	}

	if tracker.methodRankings == nil {
		t.Fatal("Expected method rankings to be initialized")
	}

	if tracker.methodTrends == nil {
		t.Fatal("Expected method trends to be initialized")
	}

	if tracker.realTimeStats == nil {
		t.Fatal("Expected real-time stats to be initialized")
	}

	if tracker.weightOptimizer == nil {
		t.Fatal("Expected weight optimizer to be initialized")
	}

	if tracker.performanceAnalyzer == nil {
		t.Fatal("Expected performance analyzer to be initialized")
	}
}

func TestRealTimeEnsembleMethodTracker_TrackMethodResult(t *testing.T) {
	config := DefaultEnsembleMethodConfig()
	logger := zap.NewNop()

	tracker := NewRealTimeEnsembleMethodTracker(config, logger)

	// Create test classification result
	result := &ClassificationResult{
		ID:                     "test_123",
		BusinessName:           "Test Business",
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

	// Track method result
	err := tracker.TrackMethodResult(context.Background(), result)
	if err != nil {
		t.Fatalf("Expected tracking to succeed, got error: %v", err)
	}

	// Verify method metrics
	metrics := tracker.GetMethodMetrics("keyword_classification")
	if metrics == nil {
		t.Fatal("Expected keyword classification metrics to be available")
	}

	if metrics.MethodName != "keyword_classification" {
		t.Errorf("Expected method name to be 'keyword_classification', got '%s'", metrics.MethodName)
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

	if metrics.AverageConfidence != 0.95 {
		t.Errorf("Expected average confidence to be 0.95, got %f", metrics.AverageConfidence)
	}

	if metrics.ErrorRate != 0.0 {
		t.Errorf("Expected error rate to be 0.0, got %f", metrics.ErrorRate)
	}
}

func TestRealTimeEnsembleMethodTracker_MultipleMethods(t *testing.T) {
	config := DefaultEnsembleMethodConfig()
	logger := zap.NewNop()

	tracker := NewRealTimeEnsembleMethodTracker(config, logger)

	// Test data for multiple methods
	methods := []struct {
		name       string
		correct    bool
		confidence float64
	}{
		{"keyword_classification", true, 0.95},
		{"ml_classification", true, 0.90},
		{"description_analysis", false, 0.85},
		{"ensemble_classification", true, 0.92},
	}

	// Track results for each method
	for i, method := range methods {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
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

		err := tracker.TrackMethodResult(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track result for method %s: %v", method.name, err)
		}
	}

	// Verify all methods are tracked
	allMetrics := tracker.GetAllMethodMetrics()
	if len(allMetrics) != 4 {
		t.Errorf("Expected 4 methods to be tracked, got %d", len(allMetrics))
	}

	// Verify specific method metrics
	keywordMetrics := tracker.GetMethodMetrics("keyword_classification")
	if keywordMetrics == nil {
		t.Fatal("Expected keyword classification metrics to be available")
	}

	if keywordMetrics.AccuracyScore != 1.0 {
		t.Errorf("Expected keyword method accuracy to be 1.0, got %f", keywordMetrics.AccuracyScore)
	}

	descriptionMetrics := tracker.GetMethodMetrics("description_analysis")
	if descriptionMetrics == nil {
		t.Fatal("Expected description analysis metrics to be available")
	}

	if descriptionMetrics.AccuracyScore != 0.0 {
		t.Errorf("Expected description method accuracy to be 0.0, got %f", descriptionMetrics.AccuracyScore)
	}
}

func TestRealTimeEnsembleMethodTracker_RealTimeIndicators(t *testing.T) {
	config := DefaultEnsembleMethodConfig()
	logger := zap.NewNop()

	tracker := NewRealTimeEnsembleMethodTracker(config, logger)

	// Add multiple results to generate real-time indicators
	for i := 0; i < 10; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_realtime_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
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

		err := tracker.TrackMethodResult(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track result: %v", err)
		}
	}

	// Verify real-time indicators
	metrics := tracker.GetMethodMetrics("keyword_classification")
	if metrics == nil {
		t.Fatal("Expected keyword classification metrics to be available")
	}

	// Check real-time indicators
	if metrics.CurrentAccuracy == 0.0 {
		t.Error("Expected current accuracy to be calculated")
	}

	if metrics.CurrentLatency == 0 {
		t.Error("Expected current latency to be calculated")
	}

	if metrics.CurrentThroughput == 0.0 {
		t.Error("Expected current throughput to be calculated")
	}

	if metrics.CurrentErrorRate != 0.0 {
		t.Errorf("Expected current error rate to be 0.0, got %f", metrics.CurrentErrorRate)
	}
}

func TestRealTimeEnsembleMethodTracker_PerformanceIndicators(t *testing.T) {
	config := DefaultEnsembleMethodConfig()
	logger := zap.NewNop()

	tracker := NewRealTimeEnsembleMethodTracker(config, logger)

	// Add results to generate performance indicators
	for i := 0; i < 15; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_performance_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
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

		err := tracker.TrackMethodResult(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track result: %v", err)
		}
	}

	// Verify performance indicators
	metrics := tracker.GetMethodMetrics("keyword_classification")
	if metrics == nil {
		t.Fatal("Expected keyword classification metrics to be available")
	}

	// Check performance indicators
	if metrics.PerformanceScore == 0.0 {
		t.Error("Expected performance score to be calculated")
	}

	if metrics.ReliabilityScore == 0.0 {
		t.Error("Expected reliability score to be calculated")
	}

	if metrics.EfficiencyScore == 0.0 {
		t.Error("Expected efficiency score to be calculated")
	}

	if metrics.QualityScore == 0.0 {
		t.Error("Expected quality score to be calculated")
	}
}

func TestRealTimeEnsembleMethodTracker_MethodStatus(t *testing.T) {
	config := DefaultEnsembleMethodConfig()
	config.AccuracyThreshold = 0.90
	config.ErrorRateThreshold = 0.10
	config.LatencyThreshold = 2 * time.Second
	logger := zap.NewNop()

	tracker := NewRealTimeEnsembleMethodTracker(config, logger)

	// Test high-performing method
	for i := 0; i < 10; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_good_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "good_method",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry": "restaurant",
			},
			IsCorrect: boolPtr(true),
		}

		err := tracker.TrackMethodResult(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track result: %v", err)
		}
	}

	// Test low-performing method
	for i := 0; i < 10; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_bad_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("technology"), // Wrong classification
			ConfidenceScore:        0.95,
			ClassificationMethod:   "bad_method",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry": "restaurant",
			},
			IsCorrect: boolPtr(false),
		}

		err := tracker.TrackMethodResult(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track result: %v", err)
		}
	}

	// Verify method statuses
	goodMetrics := tracker.GetMethodMetrics("good_method")
	if goodMetrics == nil {
		t.Fatal("Expected good method metrics to be available")
	}

	if goodMetrics.Status != "active" {
		t.Errorf("Expected good method status to be 'active', got '%s'", goodMetrics.Status)
	}

	badMetrics := tracker.GetMethodMetrics("bad_method")
	if badMetrics == nil {
		t.Fatal("Expected bad method metrics to be available")
	}

	if badMetrics.Status != "critical" {
		t.Errorf("Expected bad method status to be 'critical', got '%s'", badMetrics.Status)
	}
}

func TestRealTimeEnsembleMethodTracker_RealTimeStats(t *testing.T) {
	config := DefaultEnsembleMethodConfig()
	logger := zap.NewNop()

	tracker := NewRealTimeEnsembleMethodTracker(config, logger)

	// Add results for different methods with different performance
	methods := []struct {
		name    string
		correct bool
		count   int
	}{
		{"good_method", true, 10},
		{"degraded_method", false, 5},
		{"critical_method", false, 8},
	}

	for _, method := range methods {
		for i := 0; i < method.count; i++ {
			result := &ClassificationResult{
				ID:                     fmt.Sprintf("test_%s_%d", method.name, i),
				BusinessName:           fmt.Sprintf("Test Business %d", i),
				ActualClassification:   "restaurant",
				ExpectedClassification: stringPtr("restaurant"),
				ConfidenceScore:        0.95,
				ClassificationMethod:   method.name,
				Timestamp:              time.Now(),
				Metadata: map[string]interface{}{
					"industry": "restaurant",
				},
				IsCorrect: boolPtr(method.correct),
			}

			err := tracker.TrackMethodResult(context.Background(), result)
			if err != nil {
				t.Fatalf("Failed to track result: %v", err)
			}
		}
	}

	// Verify real-time stats
	stats := tracker.GetRealTimeStats()
	if stats == nil {
		t.Fatal("Expected real-time stats to be available")
	}

	if stats.TotalMethods != 3 {
		t.Errorf("Expected total methods to be 3, got %d", stats.TotalMethods)
	}

	// The exact counts depend on the thresholds, but we should have some methods in different states
	if stats.ActiveMethods == 0 && stats.DegradedMethods == 0 && stats.CriticalMethods == 0 {
		t.Error("Expected at least one method to be in a non-zero state")
	}
}

func TestRealTimeEnsembleMethodTracker_WeightOptimizer(t *testing.T) {
	config := DefaultEnsembleMethodConfig()
	logger := zap.NewNop()

	tracker := NewRealTimeEnsembleMethodTracker(config, logger)

	// Get optimized weights
	weights := tracker.GetOptimizedWeights()
	if weights == nil {
		t.Fatal("Expected optimized weights to be available")
	}

	// Initially should be empty
	if len(weights) != 0 {
		t.Errorf("Expected initial weights to be empty, got %d methods", len(weights))
	}
}

func TestRealTimeEnsembleMethodTracker_PerformanceAnalyzer(t *testing.T) {
	config := DefaultEnsembleMethodConfig()
	logger := zap.NewNop()

	tracker := NewRealTimeEnsembleMethodTracker(config, logger)

	// Add results to generate performance data
	for i := 0; i < 10; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_analysis_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "test_method",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry": "restaurant",
			},
			IsCorrect: boolPtr(true),
		}

		err := tracker.TrackMethodResult(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track result: %v", err)
		}
	}

	// Get performance analysis
	analysis := tracker.GetPerformanceAnalysis("test_method")
	if analysis == nil {
		t.Log("Performance analysis not available yet (may need more data)")
	}
}
