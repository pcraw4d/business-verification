package classification_monitoring

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestAdvancedAccuracyTracker_EdgeCases tests edge cases and boundary conditions
func TestAdvancedAccuracyTracker_EdgeCases(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	logger := zap.NewNop()
	tracker := NewAdvancedAccuracyTracker(config, logger)

	// Test with zero confidence score
	result := &ClassificationResult{
		ID:                     "edge_case_1",
		BusinessName:           "Zero Confidence Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.0,
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"industry": "restaurant",
		},
		IsCorrect: boolPtr(true),
	}

	err := tracker.TrackClassification(context.Background(), result)
	if err != nil {
		t.Fatalf("Expected zero confidence to be handled gracefully: %v", err)
	}

	// Test with maximum confidence score
	result = &ClassificationResult{
		ID:                     "edge_case_2",
		BusinessName:           "Max Confidence Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        1.0,
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"industry": "restaurant",
		},
		IsCorrect: boolPtr(true),
	}

	err = tracker.TrackClassification(context.Background(), result)
	if err != nil {
		t.Fatalf("Expected maximum confidence to be handled gracefully: %v", err)
	}

	// Test with very long business name
	longName := string(make([]byte, 1000))
	for i := range longName {
		longName = longName[:i] + "A" + longName[i+1:]
	}

	result = &ClassificationResult{
		ID:                     "edge_case_3",
		BusinessName:           longName,
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"industry": "restaurant",
		},
		IsCorrect: boolPtr(true),
	}

	err = tracker.TrackClassification(context.Background(), result)
	if err != nil {
		t.Fatalf("Expected long business name to be handled gracefully: %v", err)
	}

	// Test with special characters in business name
	result = &ClassificationResult{
		ID:                     "edge_case_4",
		BusinessName:           "Business with Special Chars: !@#$%^&*()_+-=[]{}|;':\",./<>?",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"industry": "restaurant",
		},
		IsCorrect: boolPtr(true),
	}

	err = tracker.TrackClassification(context.Background(), result)
	if err != nil {
		t.Fatalf("Expected special characters to be handled gracefully: %v", err)
	}

	// Test with empty business name
	result = &ClassificationResult{
		ID:                     "edge_case_5",
		BusinessName:           "",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"industry": "restaurant",
		},
		IsCorrect: boolPtr(true),
	}

	err = tracker.TrackClassification(context.Background(), result)
	if err != nil {
		t.Fatalf("Expected empty business name to be handled gracefully: %v", err)
	}

	// Test with very old timestamp
	result = &ClassificationResult{
		ID:                     "edge_case_6",
		BusinessName:           "Old Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now().Add(-365 * 24 * time.Hour), // 1 year ago
		Metadata: map[string]interface{}{
			"industry": "restaurant",
		},
		IsCorrect: boolPtr(true),
	}

	err = tracker.TrackClassification(context.Background(), result)
	if err != nil {
		t.Fatalf("Expected old timestamp to be handled gracefully: %v", err)
	}

	// Test with future timestamp
	result = &ClassificationResult{
		ID:                     "edge_case_7",
		BusinessName:           "Future Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now().Add(24 * time.Hour), // 1 day in future
		Metadata: map[string]interface{}{
			"industry": "restaurant",
		},
		IsCorrect: boolPtr(true),
	}

	err = tracker.TrackClassification(context.Background(), result)
	if err != nil {
		t.Fatalf("Expected future timestamp to be handled gracefully: %v", err)
	}
}

// TestAdvancedAccuracyTracker_InvalidInputs tests invalid input handling
func TestAdvancedAccuracyTracker_InvalidInputs(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	logger := zap.NewNop()
	tracker := NewAdvancedAccuracyTracker(config, logger)

	// Test with nil result
	err := tracker.TrackClassification(context.Background(), nil)
	if err == nil {
		t.Error("Expected error when tracking nil result")
	}

	// Test with empty ID
	result := &ClassificationResult{
		ID:                     "",
		BusinessName:           "Test Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"industry": "restaurant",
		},
		IsCorrect: boolPtr(true),
	}

	err = tracker.TrackClassification(context.Background(), result)
	if err == nil {
		t.Error("Expected error when tracking result with empty ID")
	}

	// Test with negative confidence score
	result = &ClassificationResult{
		ID:                     "invalid_confidence",
		BusinessName:           "Test Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        -0.1,
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"industry": "restaurant",
		},
		IsCorrect: boolPtr(true),
	}

	err = tracker.TrackClassification(context.Background(), result)
	if err == nil {
		t.Error("Expected error when tracking result with negative confidence")
	}

	// Test with confidence score > 1.0
	result = &ClassificationResult{
		ID:                     "invalid_confidence_high",
		BusinessName:           "Test Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        1.1,
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"industry": "restaurant",
		},
		IsCorrect: boolPtr(true),
	}

	err = tracker.TrackClassification(context.Background(), result)
	if err == nil {
		t.Error("Expected error when tracking result with confidence > 1.0")
	}

	// Test with empty classification method
	result = &ClassificationResult{
		ID:                     "empty_method",
		BusinessName:           "Test Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"industry": "restaurant",
		},
		IsCorrect: boolPtr(true),
	}

	err = tracker.TrackClassification(context.Background(), result)
	if err == nil {
		t.Error("Expected error when tracking result with empty classification method")
	}

	// Test with empty actual classification
	result = &ClassificationResult{
		ID:                     "empty_actual",
		BusinessName:           "Test Business",
		ActualClassification:   "",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"industry": "restaurant",
		},
		IsCorrect: boolPtr(true),
	}

	err = tracker.TrackClassification(context.Background(), result)
	if err == nil {
		t.Error("Expected error when tracking result with empty actual classification")
	}
}

// TestAdvancedAccuracyTracker_MetadataHandling tests metadata handling
func TestAdvancedAccuracyTracker_MetadataHandling(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	logger := zap.NewNop()
	tracker := NewAdvancedAccuracyTracker(config, logger)

	// Test with nil metadata
	result := &ClassificationResult{
		ID:                     "nil_metadata",
		BusinessName:           "Test Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now(),
		Metadata:               nil,
		IsCorrect:              boolPtr(true),
	}

	err := tracker.TrackClassification(context.Background(), result)
	if err != nil {
		t.Fatalf("Expected nil metadata to be handled gracefully: %v", err)
	}

	// Test with empty metadata
	result = &ClassificationResult{
		ID:                     "empty_metadata",
		BusinessName:           "Test Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now(),
		Metadata:               make(map[string]interface{}),
		IsCorrect:              boolPtr(true),
	}

	err = tracker.TrackClassification(context.Background(), result)
	if err != nil {
		t.Fatalf("Expected empty metadata to be handled gracefully: %v", err)
	}

	// Test with complex metadata types
	result = &ClassificationResult{
		ID:                     "complex_metadata",
		BusinessName:           "Test Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"industry":    "restaurant",
			"score":       95,
			"active":      true,
			"tags":        []string{"food", "service", "local"},
			"coordinates": map[string]float64{"lat": 40.7128, "lng": -74.0060},
			"nested":      map[string]interface{}{"level1": map[string]interface{}{"level2": "value"}},
		},
		IsCorrect: boolPtr(true),
	}

	err = tracker.TrackClassification(context.Background(), result)
	if err != nil {
		t.Fatalf("Expected complex metadata to be handled gracefully: %v", err)
	}

	// Test with very large metadata
	largeMetadata := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		largeMetadata[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
	}

	result = &ClassificationResult{
		ID:                     "large_metadata",
		BusinessName:           "Test Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now(),
		Metadata:               largeMetadata,
		IsCorrect:              boolPtr(true),
	}

	err = tracker.TrackClassification(context.Background(), result)
	if err != nil {
		t.Fatalf("Expected large metadata to be handled gracefully: %v", err)
	}
}

// TestAdvancedAccuracyTracker_ConcurrentAccess tests concurrent access scenarios
func TestAdvancedAccuracyTracker_ConcurrentAccess(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	logger := zap.NewNop()
	tracker := NewAdvancedAccuracyTracker(config, logger)

	// Test concurrent tracking and retrieval
	done := make(chan bool)

	// Goroutine 1: Track classifications
	go func() {
		for i := 0; i < 100; i++ {
			result := &ClassificationResult{
				ID:                     fmt.Sprintf("concurrent_track_%d", i),
				BusinessName:           fmt.Sprintf("Business %d", i),
				ActualClassification:   "restaurant",
				ExpectedClassification: stringPtr("restaurant"),
				ConfidenceScore:        0.95,
				ClassificationMethod:   "ensemble",
				Timestamp:              time.Now(),
				Metadata: map[string]interface{}{
					"industry": "restaurant",
				},
				IsCorrect: boolPtr(true),
			}
			tracker.TrackClassification(context.Background(), result)
		}
		done <- true
	}()

	// Goroutine 2: Retrieve metrics
	go func() {
		for i := 0; i < 100; i++ {
			tracker.GetOverallAccuracy()
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify final state
	accuracy := tracker.GetOverallAccuracy()
	if accuracy == 0 {
		t.Error("Expected accuracy to be calculated")
	}
}

// TestAdvancedAccuracyTracker_MemoryLeaks tests for potential memory leaks
func TestAdvancedAccuracyTracker_MemoryLeaks(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	logger := zap.NewNop()
	tracker := NewAdvancedAccuracyTracker(config, logger)

	// Add many classifications
	for i := 0; i < 10000; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("memory_test_%d", i),
			BusinessName:           fmt.Sprintf("Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "ensemble",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry": "restaurant",
			},
			IsCorrect: boolPtr(true),
		}
		tracker.TrackClassification(context.Background(), result)
	}

	// Verify metrics are still accessible
	accuracy := tracker.GetOverallAccuracy()
	if accuracy == 0 {
		t.Error("Expected accuracy to be calculated")
	}
}

// TestAdvancedAccuracyTracker_ConfigurationValidation tests configuration validation
func TestAdvancedAccuracyTracker_ConfigurationValidation(t *testing.T) {
	// Test with invalid thresholds
	invalidConfig := &AdvancedAccuracyConfig{
		EnableRealTimeTracking:    true,
		TargetAccuracy:            -0.1, // Invalid negative threshold
		CriticalAccuracyThreshold: 1.1,  // Invalid > 1.0 threshold
		WarningAccuracyThreshold:  0.5,
		CollectionInterval:        -time.Minute, // Invalid negative duration
		AlertCheckInterval:        time.Minute,
		TrendAnalysisInterval:     time.Minute,
		SampleWindowSize:          0, // Invalid zero size
		MinSamplesForAnalysis:     0, // Invalid zero count
		EnableSecurityTracking:    true,
		SecurityTrustTarget:       -0.1, // Invalid negative threshold
		EnablePerformanceTracking: true,
		MaxProcessingTime:         time.Minute,
	}

	logger := zap.NewNop()
	tracker := NewAdvancedAccuracyTracker(invalidConfig, logger)

	// Tracker should still be created but with validated config
	if tracker == nil {
		t.Fatal("Expected tracker to be created even with invalid config")
	}

	// Test with extreme values
	extremeConfig := &AdvancedAccuracyConfig{
		EnableRealTimeTracking:    true,
		TargetAccuracy:            0.999999,
		CriticalAccuracyThreshold: 0.999999,
		WarningAccuracyThreshold:  0.999999,
		CollectionInterval:        24 * time.Hour,
		AlertCheckInterval:        time.Hour,
		TrendAnalysisInterval:     time.Hour,
		SampleWindowSize:          1000,
		MinSamplesForAnalysis:     1000,
		EnableSecurityTracking:    true,
		SecurityTrustTarget:       0.999999,
		EnablePerformanceTracking: true,
		MaxProcessingTime:         time.Hour,
	}

	tracker = NewAdvancedAccuracyTracker(extremeConfig, logger)
	if tracker == nil {
		t.Fatal("Expected tracker to be created with extreme config")
	}
}

// TestAdvancedAccuracyTracker_ErrorRecovery tests error recovery scenarios
func TestAdvancedAccuracyTracker_ErrorRecovery(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	logger := zap.NewNop()
	tracker := NewAdvancedAccuracyTracker(config, logger)

	// Test recovery from invalid input
	invalidResult := &ClassificationResult{
		ID:                     "invalid_recovery",
		BusinessName:           "Test Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        -0.1, // Invalid confidence
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"industry": "restaurant",
		},
		IsCorrect: boolPtr(true),
	}

	// This should return an error
	err := tracker.TrackClassification(context.Background(), invalidResult)
	if err == nil {
		t.Error("Expected error for invalid input")
	}

	// Valid input should still work after error
	validResult := &ClassificationResult{
		ID:                     "valid_after_error",
		BusinessName:           "Test Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"industry": "restaurant",
		},
		IsCorrect: boolPtr(true),
	}

	err = tracker.TrackClassification(context.Background(), validResult)
	if err != nil {
		t.Fatalf("Expected valid input to work after error: %v", err)
	}

	// Verify metrics are still accessible
	accuracy := tracker.GetOverallAccuracy()
	if accuracy == 0 {
		t.Error("Expected accuracy to be calculated after error recovery")
	}
}
