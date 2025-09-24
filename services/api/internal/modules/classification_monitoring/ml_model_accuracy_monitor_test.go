package classification_monitoring

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestMLModelAccuracyMonitor_Creation(t *testing.T) {
	config := DefaultMLModelMonitorConfig()
	logger := zap.NewNop()

	monitor := NewMLModelAccuracyMonitor(config, logger)

	if monitor == nil {
		t.Fatal("Expected monitor to be created")
	}

	if monitor.config == nil {
		t.Fatal("Expected config to be set")
	}

	if monitor.logger == nil {
		t.Fatal("Expected logger to be set")
	}

	if monitor.models == nil {
		t.Fatal("Expected models map to be initialized")
	}

	if monitor.driftAlerts == nil {
		t.Fatal("Expected drift alerts to be initialized")
	}

	if monitor.trendAnalyzer == nil {
		t.Fatal("Expected trend analyzer to be initialized")
	}

	if monitor.driftDetector == nil {
		t.Fatal("Expected drift detector to be initialized")
	}

	if monitor.performanceTracker == nil {
		t.Fatal("Expected performance tracker to be initialized")
	}
}

func TestMLModelAccuracyMonitor_TrackModelPrediction(t *testing.T) {
	config := DefaultMLModelMonitorConfig()
	logger := zap.NewNop()

	monitor := NewMLModelAccuracyMonitor(config, logger)

	// Create test classification result
	result := &ClassificationResult{
		ID:                     "test_123",
		BusinessName:           "Test Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "ml_classification",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"industry":      "restaurant",
			"model_version": "v1.0",
		},
		IsCorrect: boolPtr(true),
	}

	// Track model prediction
	err := monitor.TrackModelPrediction(context.Background(), result)
	if err != nil {
		t.Fatalf("Expected tracking to succeed, got error: %v", err)
	}

	// Verify model metrics
	metrics := monitor.GetModelMetrics("ml_classification", "v1.0")
	if metrics == nil {
		t.Fatal("Expected ML model metrics to be available")
	}

	if metrics.ModelName != "ml_classification" {
		t.Errorf("Expected model name to be 'ml_classification', got '%s'", metrics.ModelName)
	}

	if metrics.ModelVersion != "v1.0" {
		t.Errorf("Expected model version to be 'v1.0', got '%s'", metrics.ModelVersion)
	}

	if metrics.TotalPredictions != 1 {
		t.Errorf("Expected total predictions to be 1, got %d", metrics.TotalPredictions)
	}

	if metrics.CorrectPredictions != 1 {
		t.Errorf("Expected correct predictions to be 1, got %d", metrics.CorrectPredictions)
	}

	if metrics.AccuracyScore != 1.0 {
		t.Errorf("Expected accuracy score to be 1.0, got %f", metrics.AccuracyScore)
	}

	if metrics.AverageConfidence != 0.95 {
		t.Errorf("Expected average confidence to be 0.95, got %f", metrics.AverageConfidence)
	}

	if metrics.DriftStatus != "stable" {
		t.Errorf("Expected drift status to be 'stable', got '%s'", metrics.DriftStatus)
	}

	if metrics.PerformanceTrend != "stable" {
		t.Errorf("Expected performance trend to be 'stable', got '%s'", metrics.PerformanceTrend)
	}
}

func TestMLModelAccuracyMonitor_MultipleModels(t *testing.T) {
	config := DefaultMLModelMonitorConfig()
	logger := zap.NewNop()

	monitor := NewMLModelAccuracyMonitor(config, logger)

	// Test data for multiple models
	models := []struct {
		name       string
		version    string
		correct    bool
		confidence float64
	}{
		{"ml_classification", "v1.0", true, 0.95},
		{"ml_classification", "v1.1", true, 0.90},
		{"deep_learning", "v2.0", false, 0.85},
		{"ensemble_ml", "v1.5", true, 0.92},
	}

	// Track predictions for each model
	for i, model := range models {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        model.confidence,
			ClassificationMethod:   model.name,
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry":      "restaurant",
				"model_version": model.version,
			},
			IsCorrect: boolPtr(model.correct),
		}

		err := monitor.TrackModelPrediction(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track prediction for model %s %s: %v", model.name, model.version, err)
		}
	}

	// Verify all models are tracked
	allMetrics := monitor.GetAllModelMetrics()
	if len(allMetrics) != 4 {
		t.Errorf("Expected 4 models to be tracked, got %d", len(allMetrics))
	}

	// Verify specific model metrics
	mlV1Metrics := monitor.GetModelMetrics("ml_classification", "v1.0")
	if mlV1Metrics == nil {
		t.Fatal("Expected ml_classification v1.0 metrics to be available")
	}

	if mlV1Metrics.AccuracyScore != 1.0 {
		t.Errorf("Expected ml_classification v1.0 accuracy to be 1.0, got %f", mlV1Metrics.AccuracyScore)
	}

	deepLearningMetrics := monitor.GetModelMetrics("deep_learning", "v2.0")
	if deepLearningMetrics == nil {
		t.Fatal("Expected deep_learning v2.0 metrics to be available")
	}

	if deepLearningMetrics.AccuracyScore != 0.0 {
		t.Errorf("Expected deep_learning v2.0 accuracy to be 0.0, got %f", deepLearningMetrics.AccuracyScore)
	}
}

func TestMLModelAccuracyMonitor_HistoricalData(t *testing.T) {
	config := DefaultMLModelMonitorConfig()
	logger := zap.NewNop()

	monitor := NewMLModelAccuracyMonitor(config, logger)

	// Add multiple predictions to generate historical data
	for i := 0; i < 15; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_historical_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "ml_classification",
			Timestamp:              time.Now().Add(-time.Duration(i) * time.Minute),
			Metadata: map[string]interface{}{
				"industry":      "restaurant",
				"model_version": "v1.0",
			},
			IsCorrect: boolPtr(true),
		}

		err := monitor.TrackModelPrediction(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track prediction: %v", err)
		}
	}

	// Verify historical data
	metrics := monitor.GetModelMetrics("ml_classification", "v1.0")
	if metrics == nil {
		t.Fatal("Expected ml_classification metrics to be available")
	}

	if len(metrics.HistoricalAccuracy) == 0 {
		t.Error("Expected historical accuracy data to be populated")
	}

	if len(metrics.HistoricalConfidence) == 0 {
		t.Error("Expected historical confidence data to be populated")
	}

	if len(metrics.HistoricalLatency) == 0 {
		t.Error("Expected historical latency data to be populated")
	}

	// Verify data points have correct values
	if len(metrics.HistoricalAccuracy) != 15 {
		t.Errorf("Expected 15 historical accuracy points, got %d", len(metrics.HistoricalAccuracy))
	}

	if len(metrics.HistoricalConfidence) != 15 {
		t.Errorf("Expected 15 historical confidence points, got %d", len(metrics.HistoricalConfidence))
	}

	if len(metrics.HistoricalLatency) != 15 {
		t.Errorf("Expected 15 historical latency points, got %d", len(metrics.HistoricalLatency))
	}
}

func TestMLModelAccuracyMonitor_PerformanceScores(t *testing.T) {
	config := DefaultMLModelMonitorConfig()
	logger := zap.NewNop()

	monitor := NewMLModelAccuracyMonitor(config, logger)

	// Add predictions to generate performance scores
	for i := 0; i < 20; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_performance_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "ml_classification",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry":      "restaurant",
				"model_version": "v1.0",
			},
			IsCorrect: boolPtr(true),
		}

		err := monitor.TrackModelPrediction(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track prediction: %v", err)
		}
	}

	// Verify performance scores
	metrics := monitor.GetModelMetrics("ml_classification", "v1.0")
	if metrics == nil {
		t.Fatal("Expected ml_classification metrics to be available")
	}

	// Check performance scores
	if metrics.ReliabilityScore == 0.0 {
		t.Error("Expected reliability score to be calculated")
	}

	if metrics.StabilityScore == 0.0 {
		t.Error("Expected stability score to be calculated")
	}

	// Scores should be positive
	if metrics.ReliabilityScore < 0 || metrics.ReliabilityScore > 1 {
		t.Errorf("Expected reliability score to be between 0 and 1, got %f", metrics.ReliabilityScore)
	}

	if metrics.StabilityScore < 0 || metrics.StabilityScore > 1 {
		t.Errorf("Expected stability score to be between 0 and 1, got %f", metrics.StabilityScore)
	}
}

func TestMLModelAccuracyMonitor_DriftDetection(t *testing.T) {
	config := DefaultMLModelMonitorConfig()
	config.BaselineWindowSize = 20      // Smaller window for testing
	config.AccuracyDriftThreshold = 0.1 // 10% threshold
	logger := zap.NewNop()

	monitor := NewMLModelAccuracyMonitor(config, logger)

	// Add baseline predictions (all correct)
	for i := 0; i < 10; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("baseline_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "ml_classification",
			Timestamp:              time.Now().Add(-time.Duration(20-i) * time.Minute),
			Metadata: map[string]interface{}{
				"industry":      "restaurant",
				"model_version": "v1.0",
			},
			IsCorrect: boolPtr(true),
		}

		err := monitor.TrackModelPrediction(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track baseline prediction: %v", err)
		}
	}

	// Add recent predictions (all incorrect to create drift)
	for i := 0; i < 10; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("recent_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("technology"), // Wrong classification
			ConfidenceScore:        0.95,
			ClassificationMethod:   "ml_classification",
			Timestamp:              time.Now().Add(-time.Duration(10-i) * time.Minute),
			Metadata: map[string]interface{}{
				"industry":      "restaurant",
				"model_version": "v1.0",
			},
			IsCorrect: boolPtr(false),
		}

		err := monitor.TrackModelPrediction(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track recent prediction: %v", err)
		}
	}

	// Verify drift detection
	metrics := monitor.GetModelMetrics("ml_classification", "v1.0")
	if metrics == nil {
		t.Fatal("Expected ml_classification metrics to be available")
	}

	// Should detect accuracy drift
	if metrics.AccuracyDrift == 0.0 {
		t.Error("Expected accuracy drift to be detected")
	}

	// Drift should be negative (performance degraded)
	if metrics.AccuracyDrift >= 0 {
		t.Errorf("Expected negative accuracy drift, got %f", metrics.AccuracyDrift)
	}

	// Should have drift status
	if metrics.DriftStatus == "stable" {
		t.Error("Expected drift status to indicate drift detected")
	}
}

func TestMLModelAccuracyMonitor_DriftAlerts(t *testing.T) {
	config := DefaultMLModelMonitorConfig()
	config.BaselineWindowSize = 20
	config.AccuracyDriftThreshold = 0.1
	logger := zap.NewNop()

	monitor := NewMLModelAccuracyMonitor(config, logger)

	// Create significant drift to trigger alerts
	// Add baseline predictions (all correct)
	for i := 0; i < 10; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("baseline_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "ml_classification",
			Timestamp:              time.Now().Add(-time.Duration(20-i) * time.Minute),
			Metadata: map[string]interface{}{
				"industry":      "restaurant",
				"model_version": "v1.0",
			},
			IsCorrect: boolPtr(true),
		}

		err := monitor.TrackModelPrediction(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track baseline prediction: %v", err)
		}
	}

	// Add recent predictions (all incorrect to create significant drift)
	for i := 0; i < 10; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("recent_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("technology"), // Wrong classification
			ConfidenceScore:        0.95,
			ClassificationMethod:   "ml_classification",
			Timestamp:              time.Now().Add(-time.Duration(10-i) * time.Minute),
			Metadata: map[string]interface{}{
				"industry":      "restaurant",
				"model_version": "v1.0",
			},
			IsCorrect: boolPtr(false),
		}

		err := monitor.TrackModelPrediction(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track recent prediction: %v", err)
		}
	}

	// Verify drift alerts
	alerts := monitor.GetDriftAlerts()
	if len(alerts) == 0 {
		t.Error("Expected drift alerts to be generated")
	}

	// Check for accuracy drift alert
	accuracyAlertFound := false
	for _, alert := range alerts {
		if alert.AlertType == "accuracy_drift" {
			accuracyAlertFound = true
			if alert.Severity == "" {
				t.Error("Expected alert to have severity")
			}
			if alert.Message == "" {
				t.Error("Expected alert to have message")
			}
			if alert.ModelName == "" {
				t.Error("Expected alert to have model name")
			}
			break
		}
	}

	if !accuracyAlertFound {
		t.Error("Expected accuracy drift alert to be generated")
	}

	// Verify active alerts
	activeAlerts := monitor.GetActiveDriftAlerts()
	if len(activeAlerts) == 0 {
		t.Error("Expected active drift alerts")
	}

	// All alerts should be unresolved initially
	for _, alert := range activeAlerts {
		if alert.Resolved {
			t.Error("Expected active alerts to be unresolved")
		}
	}
}

func TestMLModelAccuracyMonitor_TrendAnalysis(t *testing.T) {
	config := DefaultMLModelMonitorConfig()
	config.MinDataPointsForTrend = 10
	logger := zap.NewNop()

	monitor := NewMLModelAccuracyMonitor(config, logger)

	// Add predictions with improving accuracy trend
	for i := 0; i < 15; i++ {
		// Simulate improving accuracy over time
		accuracy := 0.7 + float64(i)*0.02 // 0.7 to 0.98
		correct := accuracy > 0.8         // More correct predictions over time

		result := &ClassificationResult{
			ID:                     fmt.Sprintf("trend_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "ml_classification",
			Timestamp:              time.Now().Add(-time.Duration(15-i) * time.Minute),
			Metadata: map[string]interface{}{
				"industry":      "restaurant",
				"model_version": "v1.0",
			},
			IsCorrect: boolPtr(correct),
		}

		err := monitor.TrackModelPrediction(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track prediction: %v", err)
		}
	}

	// Verify trend analysis
	metrics := monitor.GetModelMetrics("ml_classification", "v1.0")
	if metrics == nil {
		t.Fatal("Expected ml_classification metrics to be available")
	}

	// Should have performance trend
	if metrics.PerformanceTrend == "" {
		t.Error("Expected performance trend to be calculated")
	}

	// With improving accuracy, trend should be improving or stable
	if metrics.PerformanceTrend != "improving" && metrics.PerformanceTrend != "stable" {
		t.Errorf("Expected performance trend to be 'improving' or 'stable', got '%s'", metrics.PerformanceTrend)
	}
}
