package classification

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/shared"
)

// TestPerformanceBasedWeightAdjusterIntegration tests the integration of the performance-based weight adjustment system
func TestPerformanceBasedWeightAdjusterIntegration(t *testing.T) {
	// Setup
	logger := log.New(os.Stdout, "[INTEGRATION_TEST] ", log.LstdFlags)

	config := PerformanceWeightConfig{
		Enabled:                 true,
		AdjustmentInterval:      100 * time.Millisecond, // Fast for testing
		MinWeight:               0.05,
		MaxWeight:               0.8,
		WeightAdjustmentStep:    0.05,
		PerformanceWindow:       1 * time.Hour,
		MinSamplesForAdjustment: 5, // Low for testing
		AccuracyThreshold:       0.7,
		PerformanceDecayFactor:  0.95,
		ABTestingEnabled:        true,
		ABTestDuration:          1 * time.Minute, // Short for testing
		ABTestTrafficSplit:      0.5,
		ABTestMinSampleSize:     10,
		ABTestSignificanceLevel: 0.05,
		LearningRate:            0.1,
		AdaptiveLearningEnabled: true,
		WeightSmoothingFactor:   0.1,
		PerformanceWeightFactor: 0.7,
	}

	// Create mock weight manager
	weightManager := &MockWeightManagerForIntegration{
		configs: make(map[string]MethodConfig),
	}

	// Create performance tracker
	performanceTracker := NewMethodPerformanceTracker(config, logger)

	// Create weight adjuster
	weightAdjuster := NewPerformanceBasedWeightAdjuster(config, performanceTracker, weightManager, logger)

	// Test 1: Basic functionality
	t.Run("BasicFunctionality", func(t *testing.T) {
		// Test that the weight adjuster initializes correctly
		if weightAdjuster == nil {
			t.Fatal("Weight adjuster should not be nil")
		}

		if weightAdjuster.config.Enabled != true {
			t.Error("Weight adjuster should be enabled")
		}

		if weightAdjuster.performanceTracker == nil {
			t.Error("Performance tracker should not be nil")
		}
	})

	// Test 2: Performance tracking
	t.Run("PerformanceTracking", func(t *testing.T) {
		methodName := "test_method"

		// Record some test results
		for i := 0; i < 10; i++ {
			result := &shared.ClassificationMethodResult{
				MethodType: methodName,
				Success:    true,
				Result: &shared.IndustryClassification{
					IndustryCode:    "TEST",
					IndustryName:    "Test Industry",
					ConfidenceScore: 0.8 + float64(i)*0.02, // Varying accuracy
				},
				Confidence:     0.8 + float64(i)*0.02,
				ProcessingTime: time.Duration(100+i*10) * time.Millisecond,
				Error:          "",
			}

			weightAdjuster.RecordClassificationResult(methodName, result)
		}

		// Check that performance data was recorded
		performanceData := performanceTracker.GetAllPerformanceData()
		if len(performanceData) == 0 {
			t.Error("Performance data should be recorded")
		}

		data, exists := performanceData[methodName]
		if !exists {
			t.Error("Performance data should exist for test method")
		}

		if data.TotalRequests != 10 {
			t.Errorf("Expected 10 total requests, got %d", data.TotalRequests)
		}

		if data.SuccessfulRequests != 10 {
			t.Errorf("Expected 10 successful requests, got %d", data.SuccessfulRequests)
		}
	})

	// Test 3: Weight calculation
	t.Run("WeightCalculation", func(t *testing.T) {
		// Create performance data for multiple methods
		methods := []string{"method1", "method2", "method3"}

		for i, methodName := range methods {
			// Create test results with different performance characteristics
			accuracy := 0.6 + float64(i)*0.15 // method1: 0.6, method2: 0.75, method3: 0.9
			latency := time.Duration(100+i*50) * time.Millisecond

			for j := 0; j < 10; j++ {
				result := &shared.ClassificationMethodResult{
					MethodType: methodName,
					Success:    true,
					Result: &shared.IndustryClassification{
						IndustryCode:    "TEST",
						IndustryName:    "Test Industry",
						ConfidenceScore: accuracy,
					},
					Confidence:     accuracy,
					ProcessingTime: latency,
					Error:          "",
				}

				weightAdjuster.RecordClassificationResult(methodName, result)
			}
		}

		// Calculate optimal weights
		performanceData := performanceTracker.GetAllPerformanceData()
		optimalWeights := weightAdjuster.calculateOptimalWeights(performanceData)

		// Check that weights were calculated
		if len(optimalWeights) == 0 {
			t.Error("Optimal weights should be calculated")
		}

		// Check that method3 (highest accuracy) gets the highest weight
		if optimalWeights["method3"] <= optimalWeights["method1"] {
			t.Error("Method3 should have higher weight than method1 due to better performance")
		}

		// Check that weights sum to approximately 1.0
		var totalWeight float64
		for _, weight := range optimalWeights {
			totalWeight += weight
		}

		if totalWeight < 0.99 || totalWeight > 1.01 {
			t.Errorf("Weights should sum to 1.0, got %.3f", totalWeight)
		}
	})

	// Test 4: A/B Testing
	t.Run("ABTesting", func(t *testing.T) {
		// Start an A/B test
		test, err := weightAdjuster.abTestManager.StartABTest(
			"test_method",
			0.3, // control weight
			0.5, // treatment weight
			1*time.Minute,
		)

		if err != nil {
			t.Errorf("A/B test should start successfully: %v", err)
		}

		if test == nil {
			t.Error("A/B test should not be nil")
		}

		if test.Status != "active" {
			t.Error("A/B test should be active")
		}

		// Test weight assignment
		weight1, variant1, _ := weightAdjuster.abTestManager.GetTestWeight("test_method", "user1")
		weight2, variant2, _ := weightAdjuster.abTestManager.GetTestWeight("test_method", "user2")

		// Weights should be either control or treatment
		if weight1 != 0.3 && weight1 != 0.5 {
			t.Errorf("Weight should be either 0.3 or 0.5, got %.3f", weight1)
		}

		if weight2 != 0.3 && weight2 != 0.5 {
			t.Errorf("Weight should be either 0.3 or 0.5, got %.3f", weight2)
		}

		// Variants should be either control or treatment
		if variant1 != "control" && variant1 != "treatment" {
			t.Errorf("Variant should be either control or treatment, got %s", variant1)
		}

		if variant2 != "control" && variant2 != "treatment" {
			t.Errorf("Variant should be either control or treatment, got %s", variant2)
		}
	})

	// Test 5: Performance summary
	t.Run("PerformanceSummary", func(t *testing.T) {
		summary := weightAdjuster.GetPerformanceSummary()
		if summary == nil {
			t.Error("Performance summary should not be nil")
		}

		// Check that summary contains expected fields
		if _, exists := summary["performance_data"]; !exists {
			t.Error("Performance summary should contain performance_data")
		}

		if _, exists := summary["active_ab_tests"]; !exists {
			t.Error("Performance summary should contain active_ab_tests")
		}

		if _, exists := summary["config"]; !exists {
			t.Error("Performance summary should contain config")
		}
	})

	t.Log("✅ All performance-based weight adjustment integration tests passed!")
}

// TestPerformanceBasedWeightAdjusterSystemLearning tests the system's ability to learn and adapt
func TestPerformanceBasedWeightAdjusterSystemLearning(t *testing.T) {
	logger := log.New(os.Stdout, "[LEARNING_TEST] ", log.LstdFlags)

	config := PerformanceWeightConfig{
		Enabled:                 true,
		AdjustmentInterval:      50 * time.Millisecond, // Very fast for testing
		MinWeight:               0.05,
		MaxWeight:               0.8,
		WeightAdjustmentStep:    0.05,
		PerformanceWindow:       1 * time.Hour,
		MinSamplesForAdjustment: 3, // Very low for testing
		AccuracyThreshold:       0.7,
		PerformanceDecayFactor:  0.95,
		ABTestingEnabled:        false, // Disable for this test
		LearningRate:            0.1,
		AdaptiveLearningEnabled: true,
		WeightSmoothingFactor:   0.1,
		PerformanceWeightFactor: 0.7,
	}

	// Create mock weight manager
	weightManager := &MockWeightManagerForIntegration{
		configs: make(map[string]MethodConfig),
	}

	// Create performance tracker
	performanceTracker := NewMethodPerformanceTracker(config, logger)

	// Create weight adjuster
	weightAdjuster := NewPerformanceBasedWeightAdjuster(config, performanceTracker, weightManager, logger)

	// Set initial weights
	weightManager.SetMethodWeight("method1", 0.5)
	weightManager.SetMethodWeight("method2", 0.5)

	// Simulate method1 performing poorly and method2 performing well
	t.Run("LearningFromPerformance", func(t *testing.T) {
		// Record poor performance for method1
		for i := 0; i < 5; i++ {
			result := &shared.ClassificationMethodResult{
				MethodType: "method1",
				Success:    true,
				Result: &shared.IndustryClassification{
					IndustryCode:    "TEST",
					IndustryName:    "Test Industry",
					ConfidenceScore: 0.5, // Poor accuracy
				},
				Confidence:     0.5,
				ProcessingTime: 500 * time.Millisecond, // High latency
				Error:          "",
			}
			weightAdjuster.RecordClassificationResult("method1", result)
		}

		// Record good performance for method2
		for i := 0; i < 5; i++ {
			result := &shared.ClassificationMethodResult{
				MethodType: "method2",
				Success:    true,
				Result: &shared.IndustryClassification{
					IndustryCode:    "TEST",
					IndustryName:    "Test Industry",
					ConfidenceScore: 0.9, // Good accuracy
				},
				Confidence:     0.9,
				ProcessingTime: 100 * time.Millisecond, // Low latency
				Error:          "",
			}
			weightAdjuster.RecordClassificationResult("method2", result)
		}

		// Perform weight adjustment
		err := weightAdjuster.performWeightAdjustment()
		if err != nil {
			t.Errorf("Weight adjustment should succeed: %v", err)
		}

		// Check that weights were adjusted appropriately
		newWeight1, _ := weightManager.GetMethodWeight("method1")
		newWeight2, _ := weightManager.GetMethodWeight("method2")

		// Method2 should have higher weight due to better performance
		if newWeight2 <= newWeight1 {
			t.Errorf("Method2 should have higher weight than method1. Got: method1=%.3f, method2=%.3f", newWeight1, newWeight2)
		}

		t.Logf("✅ System learned from performance: method1=%.3f, method2=%.3f", newWeight1, newWeight2)
	})

	t.Log("✅ All system learning tests passed!")
}

// MockWeightManagerForIntegration is a mock implementation for testing
type MockWeightManagerForIntegration struct {
	configs map[string]MethodConfig
}

func (mwm *MockWeightManagerForIntegration) SetMethodWeight(methodName string, weight float64) error {
	config, exists := mwm.configs[methodName]
	if !exists {
		config = MethodConfig{
			Name:    methodName,
			Type:    "test",
			Weight:  weight,
			Enabled: true,
		}
	} else {
		config.Weight = weight
	}
	mwm.configs[methodName] = config
	return nil
}

func (mwm *MockWeightManagerForIntegration) GetMethodWeight(methodName string) (float64, error) {
	config, exists := mwm.configs[methodName]
	if !exists {
		return 0.5, nil // Default weight
	}
	return config.Weight, nil
}

func (mwm *MockWeightManagerForIntegration) GetMethodConfig(methodName string) (MethodConfig, error) {
	config, exists := mwm.configs[methodName]
	if !exists {
		return MethodConfig{
			Name:    methodName,
			Type:    "test",
			Weight:  0.5,
			Enabled: true,
		}, nil
	}
	return config, nil
}

func (mwm *MockWeightManagerForIntegration) SaveConfiguration() error {
	// Mock implementation - just return success
	return nil
}
