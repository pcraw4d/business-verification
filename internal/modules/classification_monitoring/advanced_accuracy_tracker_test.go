package classification_monitoring

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestAdvancedAccuracyTracker_Creation(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	logger := zap.NewNop()

	tracker := NewAdvancedAccuracyTracker(config, logger)

	if tracker == nil {
		t.Fatal("Expected tracker to be created")
	}

	if tracker.config == nil {
		t.Fatal("Expected config to be set")
	}

	if tracker.logger == nil {
		t.Fatal("Expected logger to be set")
	}

	if tracker.overallTracker == nil {
		t.Fatal("Expected overall tracker to be initialized")
	}

	if tracker.industryTracker == nil {
		t.Fatal("Expected industry tracker to be initialized")
	}

	if tracker.ensembleTracker == nil {
		t.Fatal("Expected ensemble tracker to be initialized")
	}

	if tracker.mlModelTracker == nil {
		t.Fatal("Expected ML model tracker to be initialized")
	}

	if tracker.securityTracker == nil {
		t.Fatal("Expected security tracker to be initialized")
	}
}

func TestAdvancedAccuracyTracker_StartStop(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	config.CollectionInterval = 100 * time.Millisecond // Fast for testing
	logger := zap.NewNop()

	tracker := NewAdvancedAccuracyTracker(config, logger)

	// Test start
	err := tracker.Start()
	if err != nil {
		t.Fatalf("Expected start to succeed, got error: %v", err)
	}

	if !tracker.active {
		t.Fatal("Expected tracker to be active after start")
	}

	// Test stop
	tracker.Stop()

	if tracker.active {
		t.Fatal("Expected tracker to be inactive after stop")
	}
}

func TestAdvancedAccuracyTracker_TrackClassification(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	logger := zap.NewNop()

	tracker := NewAdvancedAccuracyTracker(config, logger)

	// Start tracker
	err := tracker.Start()
	if err != nil {
		t.Fatalf("Failed to start tracker: %v", err)
	}
	defer tracker.Stop()

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

	// Track classification
	err = tracker.TrackClassification(context.Background(), result)
	if err != nil {
		t.Fatalf("Expected tracking to succeed, got error: %v", err)
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Verify overall accuracy
	overallAccuracy := tracker.GetOverallAccuracy()
	if overallAccuracy != 1.0 {
		t.Errorf("Expected overall accuracy to be 1.0, got %f", overallAccuracy)
	}

	// Verify industry accuracy
	industryAccuracy := tracker.GetIndustryAccuracy("restaurant")
	if industryAccuracy != 1.0 {
		t.Errorf("Expected restaurant accuracy to be 1.0, got %f", industryAccuracy)
	}

	// Verify method accuracy
	methodAccuracy := tracker.GetMethodAccuracy("keyword_classification")
	if methodAccuracy != 1.0 {
		t.Errorf("Expected method accuracy to be 1.0, got %f", methodAccuracy)
	}
}

func TestAdvancedAccuracyTracker_TargetAccuracyMet(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	config.TargetAccuracy = 0.95
	logger := zap.NewNop()

	tracker := NewAdvancedAccuracyTracker(config, logger)

	// Start tracker
	err := tracker.Start()
	if err != nil {
		t.Fatalf("Failed to start tracker: %v", err)
	}
	defer tracker.Stop()

	// Add high-accuracy results
	for i := 0; i < 10; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_%d", i),
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

		err = tracker.TrackClassification(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track classification: %v", err)
		}
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Check if target accuracy is met
	if !tracker.IsTargetAccuracyMet() {
		t.Error("Expected target accuracy to be met")
	}

	// Check accuracy status
	status := tracker.GetAccuracyStatus()
	if status != "excellent" {
		t.Errorf("Expected status to be 'excellent', got '%s'", status)
	}
}

func TestAdvancedAccuracyTracker_CriticalThresholdBreached(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	config.CriticalAccuracyThreshold = 0.90
	logger := zap.NewNop()

	tracker := NewAdvancedAccuracyTracker(config, logger)

	// Start tracker
	err := tracker.Start()
	if err != nil {
		t.Fatalf("Failed to start tracker: %v", err)
	}
	defer tracker.Stop()

	// Add low-accuracy results
	for i := 0; i < 10; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("technology"), // Wrong classification
			ConfidenceScore:        0.95,
			ClassificationMethod:   "keyword_classification",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry": "restaurant",
			},
			IsCorrect: boolPtr(false),
		}

		err = tracker.TrackClassification(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track classification: %v", err)
		}
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Check if critical threshold is breached
	if !tracker.IsCriticalThresholdBreached() {
		t.Error("Expected critical threshold to be breached")
	}

	// Check accuracy status
	status := tracker.GetAccuracyStatus()
	if status != "critical" {
		t.Errorf("Expected status to be 'critical', got '%s'", status)
	}
}

func TestAdvancedAccuracyTracker_IndustrySpecificTracking(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	logger := zap.NewNop()

	tracker := NewAdvancedAccuracyTracker(config, logger)

	// Start tracker
	err := tracker.Start()
	if err != nil {
		t.Fatalf("Failed to start tracker: %v", err)
	}
	defer tracker.Stop()

	// Add results for different industries
	industries := []string{"restaurant", "technology", "healthcare"}

	for _, industry := range industries {
		for i := 0; i < 5; i++ {
			result := &ClassificationResult{
				ID:                     fmt.Sprintf("test_%s_%d", industry, i),
				BusinessName:           fmt.Sprintf("Test %s Business %d", industry, i),
				ActualClassification:   industry,
				ExpectedClassification: stringPtr(industry),
				ConfidenceScore:        0.95,
				ClassificationMethod:   "keyword_classification",
				Timestamp:              time.Now(),
				Metadata: map[string]interface{}{
					"industry": industry,
				},
				IsCorrect: boolPtr(true),
			}

			err = tracker.TrackClassification(context.Background(), result)
			if err != nil {
				t.Fatalf("Failed to track classification: %v", err)
			}
		}
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Verify all industries have 100% accuracy
	for _, industry := range industries {
		accuracy := tracker.GetIndustryAccuracy(industry)
		if accuracy != 1.0 {
			t.Errorf("Expected %s accuracy to be 1.0, got %f", industry, accuracy)
		}
	}
}

func TestAdvancedAccuracyTracker_EnsembleMethodTracking(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	logger := zap.NewNop()

	tracker := NewAdvancedAccuracyTracker(config, logger)

	// Start tracker
	err := tracker.Start()
	if err != nil {
		t.Fatalf("Failed to start tracker: %v", err)
	}
	defer tracker.Stop()

	// Add results for different methods
	methods := []string{"keyword_classification", "ml_classification", "description_analysis"}

	for _, method := range methods {
		for i := 0; i < 5; i++ {
			result := &ClassificationResult{
				ID:                     fmt.Sprintf("test_%s_%d", method, i),
				BusinessName:           fmt.Sprintf("Test Business %d", i),
				ActualClassification:   "restaurant",
				ExpectedClassification: stringPtr("restaurant"),
				ConfidenceScore:        0.95,
				ClassificationMethod:   method,
				Timestamp:              time.Now(),
				Metadata: map[string]interface{}{
					"industry": "restaurant",
				},
				IsCorrect: boolPtr(true),
			}

			err = tracker.TrackClassification(context.Background(), result)
			if err != nil {
				t.Fatalf("Failed to track classification: %v", err)
			}
		}
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Verify all methods have 100% accuracy
	for _, method := range methods {
		accuracy := tracker.GetMethodAccuracy(method)
		if accuracy != 1.0 {
			t.Errorf("Expected %s accuracy to be 1.0, got %f", method, accuracy)
		}
	}
}

func TestAdvancedAccuracyTracker_MLModelTracking(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	logger := zap.NewNop()

	tracker := NewAdvancedAccuracyTracker(config, logger)

	// Start tracker
	err := tracker.Start()
	if err != nil {
		t.Fatalf("Failed to start tracker: %v", err)
	}
	defer tracker.Stop()

	// Add results for ML models
	models := []string{"bert_model", "transformer_model", "ensemble_model"}

	for _, model := range models {
		for i := 0; i < 5; i++ {
			result := &ClassificationResult{
				ID:                     fmt.Sprintf("test_%s_%d", model, i),
				BusinessName:           fmt.Sprintf("Test Business %d", i),
				ActualClassification:   "restaurant",
				ExpectedClassification: stringPtr("restaurant"),
				ConfidenceScore:        0.95,
				ClassificationMethod:   "ml_classification",
				Timestamp:              time.Now(),
				Metadata: map[string]interface{}{
					"industry":      "restaurant",
					"model_name":    model,
					"model_version": "v1.0",
				},
				IsCorrect: boolPtr(true),
			}

			err = tracker.TrackClassification(context.Background(), result)
			if err != nil {
				t.Fatalf("Failed to track classification: %v", err)
			}
		}
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Verify all models have 100% accuracy
	for _, model := range models {
		modelKey := fmt.Sprintf("%s_v1.0", model)
		accuracy := tracker.GetMLModelAccuracy(modelKey)
		if accuracy != 1.0 {
			t.Errorf("Expected %s accuracy to be 1.0, got %f", model, accuracy)
		}
	}
}

func TestAdvancedAccuracyTracker_SecurityMetrics(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	config.EnableSecurityTracking = true
	logger := zap.NewNop()

	tracker := NewAdvancedAccuracyTracker(config, logger)

	// Start tracker
	err := tracker.Start()
	if err != nil {
		t.Fatalf("Failed to start tracker: %v", err)
	}
	defer tracker.Stop()

	// Add results with security metadata
	for i := 0; i < 5; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_security_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "keyword_classification",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry":            "restaurant",
				"trusted_data_source": true,
				"website_verified":    true,
			},
			IsCorrect: boolPtr(true),
		}

		err = tracker.TrackClassification(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track classification: %v", err)
		}
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Verify security metrics
	securityMetrics := tracker.GetSecurityMetrics()
	if securityMetrics == nil {
		t.Fatal("Expected security metrics to be available")
	}

	if securityMetrics.TrustedDataSourceRate != 1.0 {
		t.Errorf("Expected trusted data source rate to be 1.0, got %f", securityMetrics.TrustedDataSourceRate)
	}

	if securityMetrics.WebsiteVerificationRate != 1.0 {
		t.Errorf("Expected website verification rate to be 1.0, got %f", securityMetrics.WebsiteVerificationRate)
	}
}

func TestAdvancedAccuracyTracker_RealTimeMetrics(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	logger := zap.NewNop()

	tracker := NewAdvancedAccuracyTracker(config, logger)

	// Start tracker
	err := tracker.Start()
	if err != nil {
		t.Fatalf("Failed to start tracker: %v", err)
	}
	defer tracker.Stop()

	// Add some results
	for i := 0; i < 5; i++ {
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

		err = tracker.TrackClassification(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track classification: %v", err)
		}
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Verify real-time metrics
	realTimeMetrics := tracker.GetRealTimeMetrics()
	if realTimeMetrics == nil {
		t.Fatal("Expected real-time metrics to be available")
	}

	if realTimeMetrics.CurrentAccuracy != 1.0 {
		t.Errorf("Expected current accuracy to be 1.0, got %f", realTimeMetrics.CurrentAccuracy)
	}

	if realTimeMetrics.HealthStatus != "healthy" {
		t.Errorf("Expected health status to be 'healthy', got '%s'", realTimeMetrics.HealthStatus)
	}
}

func TestAdvancedAccuracyTracker_TrendAnalysis(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	logger := zap.NewNop()

	tracker := NewAdvancedAccuracyTracker(config, logger)

	// Start tracker
	err := tracker.Start()
	if err != nil {
		t.Fatalf("Failed to start tracker: %v", err)
	}
	defer tracker.Stop()

	// Add results with improving accuracy over time
	for i := 0; i < 20; i++ {
		accuracy := 0.8 + float64(i)*0.01 // Improving from 80% to 99%
		isCorrect := accuracy > 0.9       // Most results are correct

		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_trend_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
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

		err = tracker.TrackClassification(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track classification: %v", err)
		}

		// Small delay to simulate time progression
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for processing and trend analysis
	time.Sleep(1 * time.Second) // Wait longer for trend analysis to run

	// Verify trend analysis
	trends := tracker.GetTrendAnalysis()
	if trends == nil {
		t.Fatal("Expected trend analysis to be available")
	}

	// Check if we have trend data (may be empty if not enough data yet)
	if len(trends) == 0 {
		t.Log("No trend data available yet (may need more data points)")
	}
}
