package classification_monitoring

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestAdvancedAccuracyTrackingIntegration tests the complete advanced accuracy tracking system
func TestAdvancedAccuracyTrackingIntegration(t *testing.T) {
	// Create all components
	config := DefaultAdvancedAccuracyConfig()
	logger := zap.NewNop()

	// Create main advanced accuracy tracker
	advancedTracker := NewAdvancedAccuracyTracker(config, logger)

	// Create industry accuracy monitor
	industryConfig := DefaultIndustryAccuracyConfig()
	industryMonitor := NewIndustryAccuracyMonitor(industryConfig, logger)

	// Create ensemble method tracker
	ensembleConfig := DefaultEnsembleMethodConfig()
	ensembleTracker := NewRealTimeEnsembleMethodTracker(ensembleConfig, logger)

	// Create ML model accuracy monitor
	mlConfig := DefaultMLModelMonitorConfig()
	mlMonitor := NewMLModelAccuracyMonitor(mlConfig, logger)

	// Create security metrics accuracy tracker
	securityTracker := NewSecurityMetricsAccuracyTracker(DefaultSecurityMetricsConfig(), logger)

	// Test data representing different business classifications
	testData := []struct {
		businessName    string
		industry        string
		expectedClass   string
		actualClass     string
		confidence      float64
		method          string
		modelVersion    string
		trustedSource   string
		websiteVerified bool
		isCorrect       bool
	}{
		{"McDonald's", "restaurant", "restaurant", "restaurant", 0.98, "ensemble", "v1.2.3", "government_database", true, true},
		{"Apple Inc", "technology", "technology", "technology", 0.95, "ml_model", "v1.2.3", "business_registry", true, true},
		{"Local Pizza Shop", "restaurant", "restaurant", "food_service", 0.85, "rule_based", "v1.1.0", "manual_entry", false, false},
		{"Microsoft Corp", "technology", "technology", "technology", 0.97, "ensemble", "v1.2.3", "government_database", true, true},
		{"Corner Store", "retail", "retail", "retail", 0.92, "ml_model", "v1.2.3", "business_registry", true, true},
		{"Fake Business", "unknown", "unknown", "technology", 0.60, "ensemble", "v1.1.0", "untrusted_source", false, false},
	}

	// Track classifications across all systems
	for i, data := range testData {
		// Create classification result
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("integration_test_%d", i),
			BusinessName:           data.businessName,
			ActualClassification:   data.actualClass,
			ExpectedClassification: &data.expectedClass,
			ConfidenceScore:        data.confidence,
			ClassificationMethod:   data.method,
			Timestamp:              time.Now().Add(-time.Duration(i) * time.Minute),
			Metadata: map[string]interface{}{
				"industry":         data.industry,
				"model_version":    data.modelVersion,
				"trusted_source":   data.trustedSource,
				"website_verified": data.websiteVerified,
			},
			IsCorrect: &data.isCorrect,
		}

		// Track in advanced accuracy tracker
		err := advancedTracker.TrackClassification(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track classification in advanced tracker: %v", err)
		}

		// Track in industry monitor
		err = industryMonitor.TrackClassification(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track industry classification: %v", err)
		}

		// Track in ensemble method tracker
		err = ensembleTracker.TrackMethodResult(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track ensemble method result: %v", err)
		}

		// Track in ML model monitor
		err = mlMonitor.TrackModelPrediction(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track ML model prediction: %v", err)
		}

		// Track in security metrics tracker
		err = securityTracker.TrackTrustedDataSourceResult(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track trusted data source result: %v", err)
		}

		// Track website verification if applicable
		if data.websiteVerified {
			err = securityTracker.TrackWebsiteVerification(context.Background(), fmt.Sprintf("%s.com", data.businessName), "ssl_certificate", true, 2*time.Second)
			if err != nil {
				t.Fatalf("Failed to track website verification: %v", err)
			}
		}
	}

	// Verify overall accuracy tracking
	overallAccuracy := advancedTracker.GetOverallAccuracy()

	// Calculate expected accuracy (4 correct out of 6 total)
	expectedAccuracy := 4.0 / 6.0
	if overallAccuracy != expectedAccuracy {
		t.Errorf("Expected overall accuracy to be %.2f, got %.2f", expectedAccuracy, overallAccuracy)
	}

	// Verify industry-specific tracking
	restaurantMetrics := industryMonitor.GetIndustryMetrics("restaurant")
	if restaurantMetrics == nil {
		t.Fatal("Expected restaurant industry metrics to be available")
	}

	// Restaurant should have 1 correct out of 2 (McDonald's correct, Local Pizza Shop incorrect)
	expectedRestaurantAccuracy := 1.0 / 2.0
	if restaurantMetrics.AccuracyScore != expectedRestaurantAccuracy {
		t.Errorf("Expected restaurant accuracy to be %.2f, got %.2f", expectedRestaurantAccuracy, restaurantMetrics.AccuracyScore)
	}

	// Verify ensemble method tracking
	ensembleMetrics := ensembleTracker.GetMethodMetrics("ensemble")
	if ensembleMetrics == nil {
		t.Fatal("Expected ensemble method metrics to be available")
	}

	// Ensemble should have 2 correct out of 3 (McDonald's, Microsoft correct, Fake Business incorrect)
	expectedEnsembleAccuracy := 2.0 / 3.0
	if ensembleMetrics.AccuracyScore != expectedEnsembleAccuracy {
		t.Errorf("Expected ensemble accuracy to be %.2f, got %.2f", expectedEnsembleAccuracy, ensembleMetrics.AccuracyScore)
	}

	// Verify ML model tracking
	mlMetrics := mlMonitor.GetModelMetrics("ml_model", "v1.2.3")
	if mlMetrics == nil {
		t.Fatal("Expected ML model v1.2.3 metrics to be available")
	}

	// v1.2.3 should have 3 correct out of 4 (Apple, Microsoft, Corner Store correct, Fake Business incorrect)
	expectedMLAccuracy := 3.0 / 4.0
	if mlMetrics.AccuracyScore != expectedMLAccuracy {
		t.Errorf("Expected ML model v1.2.3 accuracy to be %.2f, got %.2f", expectedMLAccuracy, mlMetrics.AccuracyScore)
	}

	// Verify security metrics tracking
	governmentMetrics := securityTracker.GetTrustedDataSourceMetrics("government_database")
	if governmentMetrics == nil {
		t.Fatal("Expected government database metrics to be available")
	}

	// Government database should have 2 correct out of 2 (McDonald's, Microsoft)
	expectedGovernmentAccuracy := 2.0 / 2.0
	if governmentMetrics.AccuracyScore != expectedGovernmentAccuracy {
		t.Errorf("Expected government database accuracy to be %.2f, got %.2f", expectedGovernmentAccuracy, governmentMetrics.AccuracyScore)
	}

	// Verify website verification tracking
	websiteMetrics := securityTracker.GetWebsiteVerificationMetrics("McDonald's.com")
	if websiteMetrics == nil {
		t.Fatal("Expected McDonald's website verification metrics to be available")
	}

	if websiteMetrics.VerificationAccuracy != 1.0 {
		t.Errorf("Expected website verification accuracy to be 1.0, got %.2f", websiteMetrics.VerificationAccuracy)
	}
}

// TestAdvancedAccuracyTrackingPerformance tests the performance of the tracking system
func TestAdvancedAccuracyTrackingPerformance(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	logger := zap.NewNop()

	advancedTracker := NewAdvancedAccuracyTracker(config, logger)
	industryConfig := DefaultIndustryAccuracyConfig()
	industryMonitor := NewIndustryAccuracyMonitor(industryConfig, logger)
	ensembleConfig := DefaultEnsembleMethodConfig()
	ensembleTracker := NewRealTimeEnsembleMethodTracker(ensembleConfig, logger)
	mlConfig := DefaultMLModelMonitorConfig()
	mlMonitor := NewMLModelAccuracyMonitor(mlConfig, logger)
	securityTracker := NewSecurityMetricsAccuracyTracker(DefaultSecurityMetricsConfig(), logger)

	// Test with a large number of classifications
	numClassifications := 1000
	startTime := time.Now()

	for i := 0; i < numClassifications; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("perf_test_%d", i),
			BusinessName:           fmt.Sprintf("Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "ensemble",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry":       "restaurant",
				"model_version":  "v1.2.3",
				"trusted_source": "government_database",
			},
			IsCorrect: boolPtr(true),
		}

		// Track in all systems
		advancedTracker.TrackClassification(context.Background(), result)
		industryMonitor.TrackClassification(context.Background(), result)
		ensembleTracker.TrackMethodResult(context.Background(), result)
		mlMonitor.TrackModelPrediction(context.Background(), result)
		securityTracker.TrackTrustedDataSourceResult(context.Background(), result)
	}

	elapsed := time.Since(startTime)

	// Performance should be reasonable (less than 1 second for 1000 classifications)
	if elapsed > time.Second {
		t.Errorf("Performance test took too long: %v for %d classifications", elapsed, numClassifications)
	}

	// Verify all systems have the expected data
	overallAccuracy := advancedTracker.GetOverallAccuracy()
	if overallAccuracy == 0 {
		t.Error("Expected overall accuracy to be calculated")
	}

	industryMetrics := industryMonitor.GetIndustryMetrics("restaurant")
	if industryMetrics.TotalClassifications != int64(numClassifications) {
		t.Errorf("Expected %d restaurant classifications, got %d", numClassifications, industryMetrics.TotalClassifications)
	}

	ensembleMetrics := ensembleTracker.GetMethodMetrics("ensemble")
	if ensembleMetrics.TotalClassifications != int64(numClassifications) {
		t.Errorf("Expected %d ensemble classifications, got %d", numClassifications, ensembleMetrics.TotalClassifications)
	}

	mlMetrics := mlMonitor.GetModelMetrics("ml_model", "v1.2.3")
	if mlMetrics == nil {
		t.Error("Expected ML model metrics to be available")
	}

	securityMetrics := securityTracker.GetTrustedDataSourceMetrics("government_database")
	if securityMetrics.TotalRequests != int64(numClassifications) {
		t.Errorf("Expected %d security requests, got %d", numClassifications, securityMetrics.TotalRequests)
	}
}

// TestAdvancedAccuracyTrackingConcurrency tests concurrent access to the tracking system
func TestAdvancedAccuracyTrackingConcurrency(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	logger := zap.NewNop()

	advancedTracker := NewAdvancedAccuracyTracker(config, logger)
	industryConfig := DefaultIndustryAccuracyConfig()
	industryMonitor := NewIndustryAccuracyMonitor(industryConfig, logger)
	ensembleConfig := DefaultEnsembleMethodConfig()
	ensembleTracker := NewRealTimeEnsembleMethodTracker(ensembleConfig, logger)
	mlConfig := DefaultMLModelMonitorConfig()
	mlMonitor := NewMLModelAccuracyMonitor(mlConfig, logger)
	securityTracker := NewSecurityMetricsAccuracyTracker(DefaultSecurityMetricsConfig(), logger)

	// Number of concurrent goroutines
	numGoroutines := 10
	classificationsPerGoroutine := 100

	// Channel to collect errors
	errorChan := make(chan error, numGoroutines)

	// Start concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < classificationsPerGoroutine; j++ {
				result := &ClassificationResult{
					ID:                     fmt.Sprintf("concurrent_test_%d_%d", goroutineID, j),
					BusinessName:           fmt.Sprintf("Business %d_%d", goroutineID, j),
					ActualClassification:   "restaurant",
					ExpectedClassification: stringPtr("restaurant"),
					ConfidenceScore:        0.95,
					ClassificationMethod:   "ensemble",
					Timestamp:              time.Now(),
					Metadata: map[string]interface{}{
						"industry":       "restaurant",
						"model_version":  "v1.2.3",
						"trusted_source": "government_database",
					},
					IsCorrect: boolPtr(true),
				}

				// Track in all systems concurrently
				if err := advancedTracker.TrackClassification(context.Background(), result); err != nil {
					errorChan <- fmt.Errorf("advanced tracker error: %v", err)
					return
				}

				if err := industryMonitor.TrackClassification(context.Background(), result); err != nil {
					errorChan <- fmt.Errorf("industry monitor error: %v", err)
					return
				}

				if err := ensembleTracker.TrackMethodResult(context.Background(), result); err != nil {
					errorChan <- fmt.Errorf("ensemble tracker error: %v", err)
					return
				}

				if err := mlMonitor.TrackModelPrediction(context.Background(), result); err != nil {
					errorChan <- fmt.Errorf("ML monitor error: %v", err)
					return
				}

				if err := securityTracker.TrackTrustedDataSourceResult(context.Background(), result); err != nil {
					errorChan <- fmt.Errorf("security tracker error: %v", err)
					return
				}
			}
			errorChan <- nil // Signal completion
		}(i)
	}

	// Collect errors
	var errors []error
	for i := 0; i < numGoroutines; i++ {
		if err := <-errorChan; err != nil {
			errors = append(errors, err)
		}
	}

	// Check for errors
	if len(errors) > 0 {
		t.Fatalf("Concurrency test failed with errors: %v", errors)
	}

	// Verify final counts
	expectedTotal := int64(numGoroutines * classificationsPerGoroutine)

	overallAccuracy := advancedTracker.GetOverallAccuracy()
	if overallAccuracy == 0 {
		t.Error("Expected overall accuracy to be calculated")
	}

	industryMetrics := industryMonitor.GetIndustryMetrics("restaurant")
	if industryMetrics.TotalClassifications != expectedTotal {
		t.Errorf("Expected %d restaurant classifications, got %d", expectedTotal, industryMetrics.TotalClassifications)
	}

	ensembleMetrics := ensembleTracker.GetMethodMetrics("ensemble")
	if ensembleMetrics.TotalClassifications != expectedTotal {
		t.Errorf("Expected %d ensemble classifications, got %d", expectedTotal, ensembleMetrics.TotalClassifications)
	}

	mlMetrics := mlMonitor.GetModelMetrics("ml_model", "v1.2.3")
	if mlMetrics == nil {
		t.Error("Expected ML model metrics to be available")
	}

	securityMetrics := securityTracker.GetTrustedDataSourceMetrics("government_database")
	if securityMetrics.TotalRequests != expectedTotal {
		t.Errorf("Expected %d security requests, got %d", expectedTotal, securityMetrics.TotalRequests)
	}
}

// TestAdvancedAccuracyTrackingDataConsistency tests data consistency across all tracking systems
func TestAdvancedAccuracyTrackingDataConsistency(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	logger := zap.NewNop()

	advancedTracker := NewAdvancedAccuracyTracker(config, logger)
	industryConfig := DefaultIndustryAccuracyConfig()
	industryMonitor := NewIndustryAccuracyMonitor(industryConfig, logger)
	ensembleConfig := DefaultEnsembleMethodConfig()
	ensembleTracker := NewRealTimeEnsembleMethodTracker(ensembleConfig, logger)
	mlConfig := DefaultMLModelMonitorConfig()
	mlMonitor := NewMLModelAccuracyMonitor(mlConfig, logger)
	securityTracker := NewSecurityMetricsAccuracyTracker(DefaultSecurityMetricsConfig(), logger)

	// Create test data with known values
	testResults := []*ClassificationResult{
		{
			ID:                     "consistency_test_1",
			BusinessName:           "Test Restaurant",
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "ensemble",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry":       "restaurant",
				"model_version":  "v1.2.3",
				"trusted_source": "government_database",
			},
			IsCorrect: boolPtr(true),
		},
		{
			ID:                     "consistency_test_2",
			BusinessName:           "Test Tech Company",
			ActualClassification:   "technology",
			ExpectedClassification: stringPtr("technology"),
			ConfidenceScore:        0.90,
			ClassificationMethod:   "ml_model",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry":       "technology",
				"model_version":  "v1.2.3",
				"trusted_source": "business_registry",
			},
			IsCorrect: boolPtr(true),
		},
		{
			ID:                     "consistency_test_3",
			BusinessName:           "Test Retail Store",
			ActualClassification:   "retail",
			ExpectedClassification: stringPtr("retail"),
			ConfidenceScore:        0.85,
			ClassificationMethod:   "rule_based",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry":       "retail",
				"model_version":  "v1.1.0",
				"trusted_source": "manual_entry",
			},
			IsCorrect: boolPtr(false), // Incorrect classification
		},
	}

	// Track all results
	for _, result := range testResults {
		advancedTracker.TrackClassification(context.Background(), result)
		industryMonitor.TrackClassification(context.Background(), result)
		ensembleTracker.TrackMethodResult(context.Background(), result)
		mlMonitor.TrackModelPrediction(context.Background(), result)
		securityTracker.TrackTrustedDataSourceResult(context.Background(), result)
	}

	// Verify data consistency across all systems
	overallAccuracy := advancedTracker.GetOverallAccuracy()

	// Check accuracy is consistent (2 correct out of 3)
	expectedAccuracy := 2.0 / 3.0
	if overallAccuracy != expectedAccuracy {
		t.Errorf("Expected accuracy %.2f, got %.2f", expectedAccuracy, overallAccuracy)
	}

	// Verify industry-specific consistency
	restaurantMetrics := industryMonitor.GetIndustryMetrics("restaurant")
	if restaurantMetrics.TotalClassifications != 1 {
		t.Errorf("Expected 1 restaurant classification, got %d", restaurantMetrics.TotalClassifications)
	}

	if restaurantMetrics.CorrectClassifications != 1 {
		t.Errorf("Expected 1 correct restaurant classification, got %d", restaurantMetrics.CorrectClassifications)
	}

	// Verify method-specific consistency
	ensembleMetrics := ensembleTracker.GetMethodMetrics("ensemble")
	if ensembleMetrics.TotalClassifications != 1 {
		t.Errorf("Expected 1 ensemble classification, got %d", ensembleMetrics.TotalClassifications)
	}

	if ensembleMetrics.CorrectClassifications != 1 {
		t.Errorf("Expected 1 correct ensemble classification, got %d", ensembleMetrics.CorrectClassifications)
	}

	// Verify ML model consistency
	mlMetrics := mlMonitor.GetModelMetrics("ml_model", "v1.2.3")
	if mlMetrics == nil {
		t.Error("Expected ML model metrics to be available")
	}

	// Verify security metrics consistency
	governmentMetrics := securityTracker.GetTrustedDataSourceMetrics("government_database")
	if governmentMetrics.TotalRequests != 1 {
		t.Errorf("Expected 1 government database request, got %d", governmentMetrics.TotalRequests)
	}

	if governmentMetrics.SuccessfulRequests != 1 {
		t.Errorf("Expected 1 successful government database request, got %d", governmentMetrics.SuccessfulRequests)
	}
}

// TestAdvancedAccuracyTrackingErrorHandling tests error handling in the tracking system
func TestAdvancedAccuracyTrackingErrorHandling(t *testing.T) {
	config := DefaultAdvancedAccuracyConfig()
	logger := zap.NewNop()

	advancedTracker := NewAdvancedAccuracyTracker(config, logger)
	industryConfig := DefaultIndustryAccuracyConfig()
	industryMonitor := NewIndustryAccuracyMonitor(industryConfig, logger)
	ensembleConfig := DefaultEnsembleMethodConfig()
	ensembleTracker := NewRealTimeEnsembleMethodTracker(ensembleConfig, logger)
	mlConfig := DefaultMLModelMonitorConfig()
	mlMonitor := NewMLModelAccuracyMonitor(mlConfig, logger)
	securityTracker := NewSecurityMetricsAccuracyTracker(DefaultSecurityMetricsConfig(), logger)

	// Test with nil result
	err := advancedTracker.TrackClassification(context.Background(), nil)
	if err == nil {
		t.Error("Expected error when tracking nil result")
	}

	// Test with invalid industry
	invalidResult := &ClassificationResult{
		ID:                   "error_test",
		BusinessName:         "Test Business",
		ActualClassification: "restaurant",
		ConfidenceScore:      0.95,
		ClassificationMethod: "ensemble",
		Timestamp:            time.Now(),
		Metadata: map[string]interface{}{
			"industry": "invalid_industry",
		},
		IsCorrect: boolPtr(true),
	}

	err = industryMonitor.TrackClassification(context.Background(), invalidResult)
	if err == nil {
		t.Error("Expected error when tracking with invalid result")
	}

	// Test with invalid method
	err = ensembleTracker.TrackMethodResult(context.Background(), invalidResult)
	if err == nil {
		t.Error("Expected error when tracking with invalid result")
	}

	// Test with invalid model version
	err = mlMonitor.TrackModelPrediction(context.Background(), invalidResult)
	if err == nil {
		t.Error("Expected error when tracking with invalid result")
	}

	// Test with invalid trusted source
	err = securityTracker.TrackTrustedDataSourceResult(context.Background(), invalidResult)
	if err == nil {
		t.Error("Expected error when tracking with invalid trusted source")
	}
}

// TestAdvancedAccuracyTrackingConfiguration tests configuration handling
func TestAdvancedAccuracyTrackingConfiguration(t *testing.T) {
	// Test with nil config
	tracker := NewAdvancedAccuracyTracker(nil, zap.NewNop())
	if tracker == nil {
		t.Fatal("Expected tracker to be created with nil config")
	}

	// Test with nil logger
	tracker = NewAdvancedAccuracyTracker(DefaultAdvancedAccuracyConfig(), nil)
	if tracker == nil {
		t.Fatal("Expected tracker to be created with nil logger")
	}

	// Test with custom config
	customConfig := &AdvancedAccuracyConfig{
		EnableRealTimeTracking:    true,
		TargetAccuracy:            0.90,
		CriticalAccuracyThreshold: 0.85,
		WarningAccuracyThreshold:  0.88,
		CollectionInterval:        5 * time.Minute,
		AlertCheckInterval:        15 * time.Minute,
		TrendAnalysisInterval:     5 * time.Minute,
		SampleWindowSize:          10,
		MinSamplesForAnalysis:     5,
		EnableSecurityTracking:    true,
		SecurityTrustTarget:       0.95,
		EnablePerformanceTracking: true,
		MaxProcessingTime:         2 * time.Minute,
	}

	tracker = NewAdvancedAccuracyTracker(customConfig, zap.NewNop())
	if tracker == nil {
		t.Fatal("Expected tracker to be created with custom config")
	}

	// Verify config is set correctly
	if tracker.config.TargetAccuracy != 0.90 {
		t.Errorf("Expected target accuracy 0.90, got %f", tracker.config.TargetAccuracy)
	}

	if tracker.config.CriticalAccuracyThreshold != 0.85 {
		t.Errorf("Expected critical accuracy threshold 0.85, got %f", tracker.config.CriticalAccuracyThreshold)
	}
}
