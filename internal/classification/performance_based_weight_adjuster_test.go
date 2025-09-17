package classification

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/shared"
)

func TestPerformanceBasedWeightAdjuster(t *testing.T) {
	// Setup
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

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
	weightManager := &MockWeightManager{
		configs: make(map[string]MethodConfig),
	}

	// Create performance tracker
	performanceTracker := NewMethodPerformanceTracker(config, logger)

	// Create weight adjuster
	weightAdjuster := NewPerformanceBasedWeightAdjuster(config, performanceTracker, weightManager, logger)

	// Test cases
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "test_initialization",
			testFunc: func(t *testing.T) {
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
			},
		},
		{
			name: "test_performance_tracking",
			testFunc: func(t *testing.T) {
				// Record some test results
				methodName := "test_method"

				// Create test results
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
			},
		},
		{
			name: "test_weight_calculation",
			testFunc: func(t *testing.T) {
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
			},
		},
		{
			name: "test_weight_adjustment",
			testFunc: func(t *testing.T) {
				// Set up initial weights
				weightManager.SetMethodWeight("method1", 0.3)
				weightManager.SetMethodWeight("method2", 0.3)
				weightManager.SetMethodWeight("method3", 0.4)

				// Create performance data
				methods := []string{"method1", "method2", "method3"}
				accuracies := []float64{0.6, 0.75, 0.9} // method3 is best

				for i, methodName := range methods {
					for j := 0; j < 10; j++ {
						result := &shared.ClassificationMethodResult{
							MethodType: methodName,
							Success:    true,
							Result: &shared.IndustryClassification{
								IndustryCode:    "TEST",
								IndustryName:    "Test Industry",
								ConfidenceScore: accuracies[i],
							},
							Confidence:     accuracies[i],
							ProcessingTime: 100 * time.Millisecond,
							Error:          "",
						}

						weightAdjuster.RecordClassificationResult(methodName, result)
					}
				}

				// Perform weight adjustment
				err := weightAdjuster.performWeightAdjustment()
				if err != nil {
					t.Errorf("Weight adjustment should succeed: %v", err)
				}

				// Check that weights were adjusted
				newWeight1, _ := weightManager.GetMethodWeight("method1")
				newWeight2, _ := weightManager.GetMethodWeight("method2")
				newWeight3, _ := weightManager.GetMethodWeight("method3")

				// Method3 should have the highest weight due to best performance
				if newWeight3 <= newWeight1 || newWeight3 <= newWeight2 {
					t.Error("Method3 should have the highest weight after adjustment")
				}
			},
		},
		{
			name: "test_ab_testing",
			testFunc: func(t *testing.T) {
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
			},
		},
		{
			name: "test_performance_score_calculation",
			testFunc: func(t *testing.T) {
				// Test performance score calculation
				data := &MethodPerformanceData{
					MethodName:         "test_method",
					TotalRequests:      100,
					SuccessfulRequests: 95,
					FailedRequests:     5,
					AverageAccuracy:    0.85,
					AverageLatency:     150 * time.Millisecond,
					LastAccuracy:       0.85,
					LastLatency:        150 * time.Millisecond,
					LastUpdated:        time.Now(),
				}

				score := weightAdjuster.calculatePerformanceScore(data)

				if score <= 0 {
					t.Error("Performance score should be positive")
				}

				// Test with different performance characteristics
				data2 := &MethodPerformanceData{
					MethodName:         "test_method2",
					TotalRequests:      100,
					SuccessfulRequests: 80,
					FailedRequests:     20,
					AverageAccuracy:    0.6,
					AverageLatency:     500 * time.Millisecond,
					LastAccuracy:       0.6,
					LastLatency:        500 * time.Millisecond,
					LastUpdated:        time.Now(),
				}

				score2 := weightAdjuster.calculatePerformanceScore(data2)

				// First method should have higher score
				if score <= score2 {
					t.Error("Better performing method should have higher score")
				}
			},
		},
		{
			name: "test_weight_bounds",
			testFunc: func(t *testing.T) {
				// Test weight clamping
				testCases := []struct {
					input    float64
					expected float64
				}{
					{0.0, 0.05},  // Below minimum
					{0.03, 0.05}, // Below minimum
					{0.1, 0.1},   // Within bounds
					{0.5, 0.5},   // Within bounds
					{0.9, 0.8},   // Above maximum
					{1.0, 0.8},   // Above maximum
				}

				for _, tc := range testCases {
					result := weightAdjuster.clampWeight(tc.input)
					if result != tc.expected {
						t.Errorf("clampWeight(%.3f) = %.3f, expected %.3f", tc.input, result, tc.expected)
					}
				}
			},
		},
		{
			name: "test_weight_normalization",
			testFunc: func(t *testing.T) {
				// Test weight normalization
				weights := map[string]float64{
					"method1": 0.2,
					"method2": 0.3,
					"method3": 0.5,
				}

				normalized := weightAdjuster.normalizeWeights(weights)

				// Check that weights sum to 1.0
				var total float64
				for _, weight := range normalized {
					total += weight
				}

				if total < 0.99 || total > 1.01 {
					t.Errorf("Normalized weights should sum to 1.0, got %.3f", total)
				}

				// Check that relative proportions are maintained
				ratio1 := normalized["method1"] / normalized["method2"]
				expectedRatio1 := 0.2 / 0.3

				if ratio1 < expectedRatio1*0.99 || ratio1 > expectedRatio1*1.01 {
					t.Errorf("Relative proportions should be maintained")
				}
			},
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

// MockWeightManager is a mock implementation of WeightConfigurationManager for testing
type MockWeightManager struct {
	configs map[string]MethodConfig
}

func (mwm *MockWeightManager) SetMethodWeight(methodName string, weight float64) error {
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

func (mwm *MockWeightManager) GetMethodWeight(methodName string) (float64, error) {
	config, exists := mwm.configs[methodName]
	if !exists {
		return 0.5, nil // Default weight
	}
	return config.Weight, nil
}

func (mwm *MockWeightManager) GetMethodConfig(methodName string) (MethodConfig, error) {
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

func (mwm *MockWeightManager) SaveConfiguration() error {
	// Mock implementation - just return success
	return nil
}

func TestEnsemblePerformanceIntegration(t *testing.T) {
	// Setup
	logger := log.New(os.Stdout, "[INTEGRATION_TEST] ", log.LstdFlags)

	config := EnsemblePerformanceConfig{
		PerformanceTrackingEnabled: true,
		PerformanceUpdateInterval:  100 * time.Millisecond,
		WeightAdjustmentEnabled:    true,
		WeightAdjustmentInterval:   100 * time.Millisecond,
		ABTestingEnabled:           true,
		ABTestAutoStart:            false,
		ABTestDuration:             1 * time.Minute,
		LearningEnabled:            true,
		LearningRate:               0.1,
		AdaptiveLearningEnabled:    true,
		MinAccuracyForBoost:        0.8,
		MaxLatencyForPenalty:       2 * time.Second,
		MinSamplesForLearning:      5,
	}

	// Create mock components
	weightManager := &MockWeightManager{configs: make(map[string]MethodConfig)}
	multiMethodClassifier := &MockMultiMethodClassifier{}

	// Create integration
	integration := NewEnsemblePerformanceIntegration(
		multiMethodClassifier,
		weightManager,
		config,
		logger,
	)

	// Test initialization
	if integration == nil {
		t.Fatal("Integration should not be nil")
	}

	// Test starting the system
	err := integration.Start()
	if err != nil {
		t.Errorf("Integration should start successfully: %v", err)
	}

	// Test classification with performance tracking
	ctx := context.Background()
	result, err := integration.ClassifyWithPerformanceTracking(
		ctx,
		"Test Business",
		"Test description",
		"https://test.com",
	)

	if err != nil {
		t.Errorf("Classification should succeed: %v", err)
	}

	if result == nil {
		t.Error("Classification result should not be nil")
	}

	// Test A/B test creation
	abTest, err := integration.StartABTest("test_method", 0.3, 0.5)
	if err != nil {
		t.Errorf("A/B test should start successfully: %v", err)
	}

	if abTest == nil {
		t.Error("A/B test should not be nil")
	}

	// Test performance summary
	summary := integration.GetPerformanceSummary()
	if summary == nil {
		t.Error("Performance summary should not be nil")
	}

	// Test user feedback recording
	integration.RecordUserFeedback(
		"test_method",
		result,
		0.9, // High rating
		"Great classification!",
	)

	// Stop the system
	integration.Stop()
}

// MockMultiMethodClassifier is a mock implementation for testing
type MockMultiMethodClassifier struct{}

func (mmc *MockMultiMethodClassifier) Classify(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*shared.IndustryClassification, error) {
	return &shared.IndustryClassification{
		IndustryCode:         "TEST",
		IndustryName:         "Test Industry",
		ConfidenceScore:      0.85,
		ClassificationMethod: "mock",
		Description:          "Mock classification",
		Evidence:             "Mock evidence",
		Metadata: map[string]interface{}{
			"method_results": []shared.ClassificationMethodResult{
				{
					MethodType:     "keyword",
					Success:        true,
					Result:         &shared.IndustryClassification{IndustryCode: "TEST", ConfidenceScore: 0.8},
					Confidence:     0.8,
					ProcessingTime: 100 * time.Millisecond,
					Error:          "",
				},
				{
					MethodType:     "ml",
					Success:        true,
					Result:         &shared.IndustryClassification{IndustryCode: "TEST", ConfidenceScore: 0.9},
					Confidence:     0.9,
					ProcessingTime: 200 * time.Millisecond,
					Error:          "",
				},
			},
		},
	}, nil
}

func TestPerformanceMetricsCalculation(t *testing.T) {
	logger := log.New(os.Stdout, "[METRICS_TEST] ", log.LstdFlags)

	config := PerformanceWeightConfig{
		Enabled:                 true,
		MinSamplesForAdjustment: 5,
		AccuracyThreshold:       0.7,
		PerformanceWeightFactor: 0.7,
	}

	performanceTracker := NewMethodPerformanceTracker(config, logger)

	// Test accuracy trend calculation
	adjuster := &PerformanceBasedWeightAdjuster{
		config: config,
		logger: logger,
	}

	// Create test accuracy history
	history := []AccuracyDataPoint{
		{Timestamp: time.Now().Add(-5 * time.Minute), Accuracy: 0.7, SampleSize: 10},
		{Timestamp: time.Now().Add(-4 * time.Minute), Accuracy: 0.72, SampleSize: 10},
		{Timestamp: time.Now().Add(-3 * time.Minute), Accuracy: 0.74, SampleSize: 10},
		{Timestamp: time.Now().Add(-2 * time.Minute), Accuracy: 0.76, SampleSize: 10},
		{Timestamp: time.Now().Add(-1 * time.Minute), Accuracy: 0.78, SampleSize: 10},
	}

	trend := adjuster.calculateAccuracyTrend(history)
	if trend <= 0 {
		t.Error("Improving accuracy should have positive trend")
	}

	// Test performance stability calculation
	stability := adjuster.calculatePerformanceStability(history)
	if stability <= 0 || stability > 1 {
		t.Error("Stability should be between 0 and 1")
	}

	// Test with more variable data
	variableHistory := []AccuracyDataPoint{
		{Timestamp: time.Now().Add(-5 * time.Minute), Accuracy: 0.5, SampleSize: 10},
		{Timestamp: time.Now().Add(-4 * time.Minute), Accuracy: 0.9, SampleSize: 10},
		{Timestamp: time.Now().Add(-3 * time.Minute), Accuracy: 0.3, SampleSize: 10},
		{Timestamp: time.Now().Add(-2 * time.Minute), Accuracy: 0.8, SampleSize: 10},
		{Timestamp: time.Now().Add(-1 * time.Minute), Accuracy: 0.4, SampleSize: 10},
	}

	variableStability := adjuster.calculatePerformanceStability(variableHistory)
	if variableStability >= stability {
		t.Error("More variable data should have lower stability")
	}
}

func BenchmarkPerformanceBasedWeightAdjuster(b *testing.B) {
	logger := log.New(os.Stdout, "[BENCHMARK] ", log.LstdFlags)

	config := PerformanceWeightConfig{
		Enabled:                 true,
		MinSamplesForAdjustment: 100,
		AccuracyThreshold:       0.7,
		PerformanceWeightFactor: 0.7,
	}

	weightManager := &MockWeightManager{configs: make(map[string]MethodConfig)}
	performanceTracker := NewMethodPerformanceTracker(config, logger)
	weightAdjuster := NewPerformanceBasedWeightAdjuster(config, performanceTracker, weightManager, logger)

	// Benchmark weight calculation
	b.Run("WeightCalculation", func(b *testing.B) {
		// Create test data
		performanceData := make(map[string]*MethodPerformanceData)
		for i := 0; i < 10; i++ {
			methodName := fmt.Sprintf("method_%d", i)
			performanceData[methodName] = &MethodPerformanceData{
				MethodName:         methodName,
				TotalRequests:      1000,
				SuccessfulRequests: 900,
				FailedRequests:     100,
				AverageAccuracy:    0.8 + float64(i)*0.02,
				AverageLatency:     time.Duration(100+i*10) * time.Millisecond,
				LastUpdated:        time.Now(),
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			weightAdjuster.calculateOptimalWeights(performanceData)
		}
	})

	// Benchmark performance score calculation
	b.Run("PerformanceScoreCalculation", func(b *testing.B) {
		data := &MethodPerformanceData{
			MethodName:         "test_method",
			TotalRequests:      1000,
			SuccessfulRequests: 950,
			FailedRequests:     50,
			AverageAccuracy:    0.85,
			AverageLatency:     150 * time.Millisecond,
			LastUpdated:        time.Now(),
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			weightAdjuster.calculatePerformanceScore(data)
		}
	})
}
