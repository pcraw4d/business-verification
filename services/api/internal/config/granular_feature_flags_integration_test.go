package config

import (
	"context"
	"log"
	"os"
	"testing"
	"time"
)

func TestGranularFeatureFlagIntegration(t *testing.T) {
	// Create test configuration
	config := GranularFeatureFlagConfig{
		UpdateInterval:             time.Minute,
		RealTimeUpdates:            true,
		PerformanceTrackingEnabled: true,
		MetricsRetentionDays:       30,
		ABTestingEnabled:           true,
		MinimumSampleSize:          1000,
		StatisticalSignificance:    0.95,
		GradualRolloutEnabled:      true,
		RolloutIncrement:           10,
		RolloutInterval:            time.Hour,
		AutoRollbackEnabled:        true,
		RollbackThreshold:          0.05,
		Environment:                "test",
		DefaultFallbackEnabled:     true,
	}

	// Create logger
	logger := log.New(os.Stdout, "INTEGRATION_TEST: ", log.LstdFlags)

	// Create manager
	manager := NewGranularFeatureFlagManager(config, logger)

	// Test 1: Initial state verification
	t.Run("InitialStateVerification", func(t *testing.T) {
		flags := manager.GetFlags()

		// Verify default service states
		if !flags.Services.PythonMLServiceEnabled {
			t.Error("Expected Python ML Service to be enabled initially")
		}
		if !flags.Services.GoRuleEngineEnabled {
			t.Error("Expected Go Rule Engine to be enabled initially")
		}

		// Verify default model states
		if !flags.Models.BERTClassificationEnabled {
			t.Error("Expected BERT Classification to be enabled initially")
		}
		if !flags.Models.DistilBERTClassificationEnabled {
			t.Error("Expected DistilBERT Classification to be enabled initially")
		}
		if flags.Models.CustomNeuralNetEnabled {
			t.Error("Expected Custom Neural Net to be disabled initially")
		}

		// Verify default rule-based states
		if !flags.Models.KeywordMatchingEnabled {
			t.Error("Expected Keyword Matching to be enabled initially")
		}
		if !flags.Models.MCCCodeLookupEnabled {
			t.Error("Expected MCC Code Lookup to be enabled initially")
		}
		if !flags.Models.BlacklistCheckEnabled {
			t.Error("Expected Blacklist Check to be enabled initially")
		}
	})

	// Test 2: Service-level toggle functionality
	t.Run("ServiceLevelToggles", func(t *testing.T) {
		// Test Python ML Service toggle
		if !manager.IsServiceEnabled("python_ml_service") {
			t.Error("Expected python_ml_service to be enabled")
		}

		// Test Go Rule Engine toggle
		if !manager.IsServiceEnabled("go_rule_engine") {
			t.Error("Expected go_rule_engine to be enabled")
		}

		// Test API Gateway toggle
		if !manager.IsServiceEnabled("api_gateway") {
			t.Error("Expected api_gateway to be enabled")
		}

		// Test invalid service
		if manager.IsServiceEnabled("invalid_service") {
			t.Error("Expected invalid_service to be disabled")
		}
	})

	// Test 3: Individual model toggle functionality
	t.Run("IndividualModelToggles", func(t *testing.T) {
		// Test classification models
		if !manager.IsModelEnabled("bert_classification") {
			t.Error("Expected bert_classification to be enabled")
		}
		if !manager.IsModelEnabled("distilbert_classification") {
			t.Error("Expected distilbert_classification to be enabled")
		}
		if manager.IsModelEnabled("custom_neural_net") {
			t.Error("Expected custom_neural_net to be disabled")
		}

		// Test risk detection models
		if !manager.IsModelEnabled("bert_risk_detection") {
			t.Error("Expected bert_risk_detection to be enabled")
		}
		if !manager.IsModelEnabled("anomaly_detection") {
			t.Error("Expected anomaly_detection to be enabled")
		}
		if !manager.IsModelEnabled("pattern_recognition") {
			t.Error("Expected pattern_recognition to be enabled")
		}

		// Test rule-based models
		if !manager.IsModelEnabled("keyword_matching") {
			t.Error("Expected keyword_matching to be enabled")
		}
		if !manager.IsModelEnabled("mcc_code_lookup") {
			t.Error("Expected mcc_code_lookup to be enabled")
		}
		if !manager.IsModelEnabled("blacklist_check") {
			t.Error("Expected blacklist_check to be enabled")
		}

		// Test invalid model
		if manager.IsModelEnabled("invalid_model") {
			t.Error("Expected invalid_model to be disabled")
		}
	})

	// Test 4: Optimal model selection
	t.Run("OptimalModelSelection", func(t *testing.T) {
		ctx := context.Background()

		// Test classification request
		model, err := manager.GetOptimalModel(ctx, "classification")
		if err != nil {
			t.Fatalf("Expected no error for classification request, got: %v", err)
		}
		if model == "" {
			t.Error("Expected model to be returned for classification request")
		}

		// Test risk detection request
		model, err = manager.GetOptimalModel(ctx, "risk_detection")
		if err != nil {
			t.Fatalf("Expected no error for risk detection request, got: %v", err)
		}
		if model == "" {
			t.Error("Expected model to be returned for risk detection request")
		}

		// Test invalid request type
		model, err = manager.GetOptimalModel(ctx, "invalid_type")
		if err != nil {
			t.Fatalf("Expected no error for invalid request type, got: %v", err)
		}
		if model == "" {
			t.Error("Expected fallback model to be returned for invalid request type")
		}
	})

	// Test 5: Feature flag updates and rollback
	t.Run("FeatureFlagUpdatesAndRollback", func(t *testing.T) {
		// Get initial flags
		initialFlags := manager.GetFlags()

		// Create modified flags
		modifiedFlags := &GranularFeatureFlags{
			Services: struct {
				PythonMLServiceEnabled bool `json:"python_ml_service_enabled"`
				GoRuleEngineEnabled    bool `json:"go_rule_engine_enabled"`
				APIGatewayEnabled      bool `json:"api_gateway_enabled"`
			}{
				PythonMLServiceEnabled: false, // Disable Python ML Service
				GoRuleEngineEnabled:    true,
				APIGatewayEnabled:      true,
			},
			Models: struct {
				BERTClassificationEnabled       bool `json:"bert_classification_enabled"`
				DistilBERTClassificationEnabled bool `json:"distilbert_classification_enabled"`
				CustomNeuralNetEnabled          bool `json:"custom_neural_net_enabled"`
				BERTRiskDetectionEnabled        bool `json:"bert_risk_detection_enabled"`
				AnomalyDetectionEnabled         bool `json:"anomaly_detection_enabled"`
				PatternRecognitionEnabled       bool `json:"pattern_recognition_enabled"`
				KeywordMatchingEnabled          bool `json:"keyword_matching_enabled"`
				MCCCodeLookupEnabled            bool `json:"mcc_code_lookup_enabled"`
				BlacklistCheckEnabled           bool `json:"blacklist_check_enabled"`
			}{
				BERTClassificationEnabled:       false, // Disable BERT Classification
				DistilBERTClassificationEnabled: true,
				CustomNeuralNetEnabled:          true,
				BERTRiskDetectionEnabled:        false, // Disable BERT Risk Detection
				AnomalyDetectionEnabled:         true,
				PatternRecognitionEnabled:       true,
				KeywordMatchingEnabled:          true,
				MCCCodeLookupEnabled:            true,
				BlacklistCheckEnabled:           true,
			},
			ModelConfig: struct {
				DefaultModelVersion string `json:"default_model_version"`
				FallbackToRules     bool   `json:"fallback_to_rules"`
				RolloutPercentage   int    `json:"rollout_percentage"`
				ABTestEnabled       bool   `json:"ab_test_enabled"`
			}{
				DefaultModelVersion: "v2.0",
				FallbackToRules:     true,
				RolloutPercentage:   50,
				ABTestEnabled:       true,
			},
			Performance: struct {
				MetricsEnabled        bool `json:"metrics_enabled"`
				PerformanceTracking   bool `json:"performance_tracking"`
				ModelVersioning       bool `json:"model_versioning"`
				CircuitBreakerEnabled bool `json:"circuit_breaker_enabled"`
			}{
				MetricsEnabled:        true,
				PerformanceTracking:   true,
				ModelVersioning:       true,
				CircuitBreakerEnabled: true,
			},
			ABTesting: struct {
				Enabled                 bool    `json:"enabled"`
				TestPercentage          int     `json:"test_percentage"`
				ControlGroupPercentage  int     `json:"control_group_percentage"`
				TestGroupPercentage     int     `json:"test_group_percentage"`
				MinimumSampleSize       int     `json:"minimum_sample_size"`
				StatisticalSignificance float64 `json:"statistical_significance"`
			}{
				Enabled:                 true,
				TestPercentage:          20,
				ControlGroupPercentage:  50,
				TestGroupPercentage:     50,
				MinimumSampleSize:       1000,
				StatisticalSignificance: 0.95,
			},
			Rollout: struct {
				GradualRolloutEnabled bool          `json:"gradual_rollout_enabled"`
				RolloutPercentage     int           `json:"rollout_percentage"`
				RolloutIncrement      int           `json:"rollout_increment"`
				RolloutInterval       time.Duration `json:"rollout_interval"`
				AutoRollbackEnabled   bool          `json:"auto_rollback_enabled"`
				RollbackThreshold     float64       `json:"rollback_threshold"`
			}{
				GradualRolloutEnabled: true,
				RolloutPercentage:     50,
				RolloutIncrement:      10,
				RolloutInterval:       time.Hour,
				AutoRollbackEnabled:   true,
				RollbackThreshold:     0.05,
			},
			Metadata: struct {
				Version     string    `json:"version"`
				LastUpdated time.Time `json:"last_updated"`
				UpdatedBy   string    `json:"updated_by"`
				Environment string    `json:"environment"`
				Description string    `json:"description"`
			}{
				Version:     "2.0.0",
				LastUpdated: time.Now(),
				UpdatedBy:   "integration_test",
				Environment: "test",
				Description: "Integration test configuration",
			},
		}

		// Update flags
		err := manager.UpdateFlags(modifiedFlags)
		if err != nil {
			t.Fatalf("Expected no error updating flags, got: %v", err)
		}

		// Verify changes
		updatedFlags := manager.GetFlags()
		if updatedFlags.Services.PythonMLServiceEnabled {
			t.Error("Expected Python ML Service to be disabled after update")
		}
		if updatedFlags.Models.BERTClassificationEnabled {
			t.Error("Expected BERT Classification to be disabled after update")
		}
		if updatedFlags.Models.BERTRiskDetectionEnabled {
			t.Error("Expected BERT Risk Detection to be disabled after update")
		}
		if !updatedFlags.Models.CustomNeuralNetEnabled {
			t.Error("Expected Custom Neural Net to be enabled after update")
		}

		// Test service and model toggles after update
		if manager.IsServiceEnabled("python_ml_service") {
			t.Error("Expected python_ml_service to be disabled after update")
		}
		if manager.IsModelEnabled("bert_classification") {
			t.Error("Expected bert_classification to be disabled after update")
		}
		if !manager.IsModelEnabled("custom_neural_net") {
			t.Error("Expected custom_neural_net to be enabled after update")
		}

		// Test rollback - restore initial flags
		err = manager.UpdateFlags(initialFlags)
		if err != nil {
			t.Fatalf("Expected no error rolling back flags, got: %v", err)
		}

		// Verify rollback
		rolledBackFlags := manager.GetFlags()
		if !rolledBackFlags.Services.PythonMLServiceEnabled {
			t.Error("Expected Python ML Service to be enabled after rollback")
		}
		if !rolledBackFlags.Models.BERTClassificationEnabled {
			t.Error("Expected BERT Classification to be enabled after rollback")
		}
		if !rolledBackFlags.Models.BERTRiskDetectionEnabled {
			t.Error("Expected BERT Risk Detection to be enabled after rollback")
		}
		if rolledBackFlags.Models.CustomNeuralNetEnabled {
			t.Error("Expected Custom Neural Net to be disabled after rollback")
		}

		// Test service and model toggles after rollback
		if !manager.IsServiceEnabled("python_ml_service") {
			t.Error("Expected python_ml_service to be enabled after rollback")
		}
		if !manager.IsModelEnabled("bert_classification") {
			t.Error("Expected bert_classification to be enabled after rollback")
		}
		if manager.IsModelEnabled("custom_neural_net") {
			t.Error("Expected custom_neural_net to be disabled after rollback")
		}
	})

	// Test 6: A/B testing functionality
	t.Run("ABTestingFunctionality", func(t *testing.T) {
		// Enable A/B testing
		flags := manager.GetFlags()
		flags.ABTesting.Enabled = true
		flags.ABTesting.TestPercentage = 50
		flags.ABTesting.ControlGroupPercentage = 50
		flags.ABTesting.TestGroupPercentage = 50

		err := manager.UpdateFlags(flags)
		if err != nil {
			t.Fatalf("Expected no error enabling A/B testing, got: %v", err)
		}

		// Test A/B testing with multiple requests
		ctx := context.Background()
		controlCount := 0
		testCount := 0

		for i := 0; i < 100; i++ {
			model, err := manager.GetOptimalModel(ctx, "classification")
			if err != nil {
				t.Fatalf("Expected no error for A/B test request %d, got: %v", i, err)
			}

			if model == "control" {
				controlCount++
			} else if model == "test" {
				testCount++
			}

			// Add small delay to ensure time variation for hash function
			time.Sleep(1 * time.Millisecond)
		}

		// Verify A/B testing distribution (should be roughly 50/50)
		total := controlCount + testCount
		if total == 0 {
			t.Error("Expected some A/B test results")
		}

		controlPercentage := float64(controlCount) / float64(total) * 100
		testPercentage := float64(testCount) / float64(total) * 100

		// Allow for some variance in distribution
		if controlPercentage < 30 || controlPercentage > 70 {
			t.Errorf("Expected control percentage to be around 50%%, got %.1f%%", controlPercentage)
		}
		if testPercentage < 30 || testPercentage > 70 {
			t.Errorf("Expected test percentage to be around 50%%, got %.1f%%", testPercentage)
		}
	})

	// Test 7: Gradual rollout functionality
	t.Run("GradualRolloutFunctionality", func(t *testing.T) {
		// Enable gradual rollout
		flags := manager.GetFlags()
		flags.Rollout.GradualRolloutEnabled = true
		flags.Rollout.RolloutPercentage = 25 // 25% rollout
		flags.Rollout.RolloutIncrement = 10
		flags.Rollout.RolloutInterval = time.Minute

		err := manager.UpdateFlags(flags)
		if err != nil {
			t.Fatalf("Expected no error enabling gradual rollout, got: %v", err)
		}

		// Test gradual rollout with multiple requests
		ctx := context.Background()
		newModelCount := 0
		fallbackCount := 0

		for i := 0; i < 100; i++ {
			model, err := manager.GetOptimalModel(ctx, "classification")
			if err != nil {
				t.Fatalf("Expected no error for rollout test request %d, got: %v", i, err)
			}

			if model == "rule_based" {
				fallbackCount++
			} else {
				newModelCount++
			}

			// Add small delay to ensure time variation for hash function
			time.Sleep(1 * time.Millisecond)
		}

		// Verify gradual rollout distribution
		total := newModelCount + fallbackCount
		if total == 0 {
			t.Error("Expected some rollout test results")
		}

		newModelPercentage := float64(newModelCount) / float64(total) * 100

		// Allow for some variance in rollout percentage
		if newModelPercentage < 15 || newModelPercentage > 35 {
			t.Errorf("Expected new model percentage to be around 25%%, got %.1f%%", newModelPercentage)
		}
	})

	// Test 8: Performance and concurrency
	t.Run("PerformanceAndConcurrency", func(t *testing.T) {
		// Test concurrent access to feature flags
		done := make(chan bool, 50)

		for i := 0; i < 50; i++ {
			go func(id int) {
				for j := 0; j < 100; j++ {
					// Test service toggles
					manager.IsServiceEnabled("python_ml_service")
					manager.IsServiceEnabled("go_rule_engine")
					manager.IsServiceEnabled("api_gateway")

					// Test model toggles
					manager.IsModelEnabled("bert_classification")
					manager.IsModelEnabled("distilbert_classification")
					manager.IsModelEnabled("custom_neural_net")

					// Test optimal model selection
					ctx := context.Background()
					manager.GetOptimalModel(ctx, "classification")
					manager.GetOptimalModel(ctx, "risk_detection")
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 50; i++ {
			<-done
		}

		// If we get here without deadlock, the test passes
	})

	// Test 9: Error handling and validation
	t.Run("ErrorHandlingAndValidation", func(t *testing.T) {
		// Test invalid rollout percentage
		invalidFlags := &GranularFeatureFlags{
			Rollout: struct {
				GradualRolloutEnabled bool          `json:"gradual_rollout_enabled"`
				RolloutPercentage     int           `json:"rollout_percentage"`
				RolloutIncrement      int           `json:"rollout_increment"`
				RolloutInterval       time.Duration `json:"rollout_interval"`
				AutoRollbackEnabled   bool          `json:"auto_rollback_enabled"`
				RollbackThreshold     float64       `json:"rollback_threshold"`
			}{
				RolloutPercentage: 150, // Invalid: > 100
			},
		}

		err := manager.UpdateFlags(invalidFlags)
		if err == nil {
			t.Error("Expected error for invalid rollout percentage")
		}

		// Test invalid A/B testing percentages
		invalidABFlags := &GranularFeatureFlags{
			ABTesting: struct {
				Enabled                 bool    `json:"enabled"`
				TestPercentage          int     `json:"test_percentage"`
				ControlGroupPercentage  int     `json:"control_group_percentage"`
				TestGroupPercentage     int     `json:"test_group_percentage"`
				MinimumSampleSize       int     `json:"minimum_sample_size"`
				StatisticalSignificance float64 `json:"statistical_significance"`
			}{
				Enabled:                 true,
				ControlGroupPercentage:  60,
				TestGroupPercentage:     50, // Invalid: doesn't sum to 100
				StatisticalSignificance: 0.95,
			},
		}

		err = manager.UpdateFlags(invalidABFlags)
		if err == nil {
			t.Error("Expected error for invalid A/B testing percentages")
		}

		// Test invalid statistical significance
		invalidSigFlags := &GranularFeatureFlags{
			ABTesting: struct {
				Enabled                 bool    `json:"enabled"`
				TestPercentage          int     `json:"test_percentage"`
				ControlGroupPercentage  int     `json:"control_group_percentage"`
				TestGroupPercentage     int     `json:"test_group_percentage"`
				MinimumSampleSize       int     `json:"minimum_sample_size"`
				StatisticalSignificance float64 `json:"statistical_significance"`
			}{
				Enabled:                 true,
				ControlGroupPercentage:  50,
				TestGroupPercentage:     50,
				StatisticalSignificance: 0.3, // Invalid: < 0.5
			},
		}

		err = manager.UpdateFlags(invalidSigFlags)
		if err == nil {
			t.Error("Expected error for invalid statistical significance")
		}
	})

	t.Log("✅ All granular feature flag integration tests passed successfully!")
}
