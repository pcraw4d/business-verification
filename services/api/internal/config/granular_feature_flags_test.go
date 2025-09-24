package config

import (
	"context"
	"log"
	"os"
	"testing"
	"time"
)

func TestNewGranularFeatureFlagManager(t *testing.T) {
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
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)

	// Create manager
	manager := NewGranularFeatureFlagManager(config, logger)

	// Test that manager was created successfully
	if manager == nil {
		t.Fatal("Expected manager to be created, got nil")
	}

	// Test that default flags are set
	flags := manager.GetFlags()
	if flags == nil {
		t.Fatal("Expected flags to be set, got nil")
	}

	// Test default service flags
	if !flags.Services.PythonMLServiceEnabled {
		t.Error("Expected Python ML Service to be enabled by default")
	}
	if !flags.Services.GoRuleEngineEnabled {
		t.Error("Expected Go Rule Engine to be enabled by default")
	}
	if !flags.Services.APIGatewayEnabled {
		t.Error("Expected API Gateway to be enabled by default")
	}

	// Test default model flags
	if !flags.Models.BERTClassificationEnabled {
		t.Error("Expected BERT Classification to be enabled by default")
	}
	if !flags.Models.DistilBERTClassificationEnabled {
		t.Error("Expected DistilBERT Classification to be enabled by default")
	}
	if flags.Models.CustomNeuralNetEnabled {
		t.Error("Expected Custom Neural Net to be disabled by default")
	}

	// Test default rule-based flags
	if !flags.Models.KeywordMatchingEnabled {
		t.Error("Expected Keyword Matching to be enabled by default")
	}
	if !flags.Models.MCCCodeLookupEnabled {
		t.Error("Expected MCC Code Lookup to be enabled by default")
	}
	if !flags.Models.BlacklistCheckEnabled {
		t.Error("Expected Blacklist Check to be enabled by default")
	}
}

func TestIsServiceEnabled(t *testing.T) {
	config := GranularFeatureFlagConfig{
		Environment: "test",
	}
	manager := NewGranularFeatureFlagManager(config, nil)

	// Test valid service names
	if !manager.IsServiceEnabled("python_ml_service") {
		t.Error("Expected python_ml_service to be enabled")
	}
	if !manager.IsServiceEnabled("go_rule_engine") {
		t.Error("Expected go_rule_engine to be enabled")
	}
	if !manager.IsServiceEnabled("api_gateway") {
		t.Error("Expected api_gateway to be enabled")
	}

	// Test invalid service name
	if manager.IsServiceEnabled("invalid_service") {
		t.Error("Expected invalid_service to be disabled")
	}
}

func TestIsModelEnabled(t *testing.T) {
	config := GranularFeatureFlagConfig{
		Environment: "test",
	}
	manager := NewGranularFeatureFlagManager(config, nil)

	// Test valid model names
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

	// Test invalid model name
	if manager.IsModelEnabled("invalid_model") {
		t.Error("Expected invalid_model to be disabled")
	}
}

func TestGetOptimalModel(t *testing.T) {
	config := GranularFeatureFlagConfig{
		Environment: "test",
	}
	manager := NewGranularFeatureFlagManager(config, nil)

	ctx := context.Background()

	// Test classification request
	model, err := manager.GetOptimalModel(ctx, "classification")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if model == "" {
		t.Error("Expected model to be returned")
	}

	// Test risk detection request
	model, err = manager.GetOptimalModel(ctx, "risk_detection")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if model == "" {
		t.Error("Expected model to be returned")
	}

	// Test invalid request type
	model, err = manager.GetOptimalModel(ctx, "invalid_type")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if model == "" {
		t.Error("Expected fallback model to be returned")
	}
}

func TestUpdateFlags(t *testing.T) {
	config := GranularFeatureFlagConfig{
		Environment: "test",
	}
	manager := NewGranularFeatureFlagManager(config, nil)

	// Create new flags
	newFlags := &GranularFeatureFlags{
		Services: struct {
			PythonMLServiceEnabled bool `json:"python_ml_service_enabled"`
			GoRuleEngineEnabled    bool `json:"go_rule_engine_enabled"`
			APIGatewayEnabled      bool `json:"api_gateway_enabled"`
		}{
			PythonMLServiceEnabled: false,
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
			BERTClassificationEnabled:       false,
			DistilBERTClassificationEnabled: true,
			CustomNeuralNetEnabled:          true,
			BERTRiskDetectionEnabled:        false,
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
			UpdatedBy:   "test",
			Environment: "test",
			Description: "Test configuration",
		},
	}

	// Update flags
	err := manager.UpdateFlags(newFlags)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify flags were updated
	updatedFlags := manager.GetFlags()
	if updatedFlags.Services.PythonMLServiceEnabled {
		t.Error("Expected Python ML Service to be disabled")
	}
	if !updatedFlags.Services.GoRuleEngineEnabled {
		t.Error("Expected Go Rule Engine to be enabled")
	}
	if !updatedFlags.Models.DistilBERTClassificationEnabled {
		t.Error("Expected DistilBERT Classification to be enabled")
	}
	if !updatedFlags.Models.CustomNeuralNetEnabled {
		t.Error("Expected Custom Neural Net to be enabled")
	}
	if updatedFlags.Models.BERTClassificationEnabled {
		t.Error("Expected BERT Classification to be disabled")
	}
	if updatedFlags.Models.BERTRiskDetectionEnabled {
		t.Error("Expected BERT Risk Detection to be disabled")
	}
}

func TestValidateFlags(t *testing.T) {
	config := GranularFeatureFlagConfig{
		Environment: "test",
	}
	manager := NewGranularFeatureFlagManager(config, nil)

	// Test valid flags
	validFlags := &GranularFeatureFlags{
		Rollout: struct {
			GradualRolloutEnabled bool          `json:"gradual_rollout_enabled"`
			RolloutPercentage     int           `json:"rollout_percentage"`
			RolloutIncrement      int           `json:"rollout_increment"`
			RolloutInterval       time.Duration `json:"rollout_interval"`
			AutoRollbackEnabled   bool          `json:"auto_rollback_enabled"`
			RollbackThreshold     float64       `json:"rollback_threshold"`
		}{
			RolloutPercentage: 50,
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
			ControlGroupPercentage:  50,
			TestGroupPercentage:     50,
			StatisticalSignificance: 0.95,
		},
	}

	err := manager.UpdateFlags(validFlags)
	if err != nil {
		t.Fatalf("Expected no error for valid flags, got: %v", err)
	}

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

	err = manager.UpdateFlags(invalidFlags)
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
}

func TestFeatureFlagPerformanceTracker(t *testing.T) {
	config := PerformanceTrackingConfig{
		Enabled:        true,
		RetentionDays:  30,
		UpdateInterval: time.Minute,
	}

	tracker := NewFeatureFlagPerformanceTracker(config)
	if tracker == nil {
		t.Fatal("Expected tracker to be created, got nil")
	}

	// Test that metrics map is initialized
	if tracker.metrics == nil {
		t.Error("Expected metrics map to be initialized")
	}
}

func TestABTester(t *testing.T) {
	config := ABTestingConfig{
		Enabled:                 true,
		DefaultTrafficSplit:     0.5,
		MinimumSampleSize:       1000,
		StatisticalSignificance: 0.95,
		TestDuration:            time.Hour * 24,
	}

	tester := NewABTester(config)
	if tester == nil {
		t.Fatal("Expected tester to be created, got nil")
	}

	// Test that tests map is initialized
	if tester.tests == nil {
		t.Error("Expected tests map to be initialized")
	}
}

func TestRolloutManager(t *testing.T) {
	config := RolloutConfig{
		Enabled:             true,
		IncrementPercentage: 10,
		IncrementInterval:   time.Hour,
		MaxPercentage:       100,
		AutoRollbackEnabled: true,
		RollbackThreshold:   0.05,
	}

	manager := NewRolloutManager(config)
	if manager == nil {
		t.Fatal("Expected manager to be created, got nil")
	}

	// Test that rollouts map is initialized
	if manager.rollouts == nil {
		t.Error("Expected rollouts map to be initialized")
	}
}

func TestConcurrentAccess(t *testing.T) {
	config := GranularFeatureFlagConfig{
		Environment: "test",
	}
	manager := NewGranularFeatureFlagManager(config, nil)

	// Test concurrent access to IsServiceEnabled
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				manager.IsServiceEnabled("python_ml_service")
				manager.IsModelEnabled("bert_classification")
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestGetOptimalModelWithContext(t *testing.T) {
	config := GranularFeatureFlagConfig{
		Environment: "test",
	}
	manager := NewGranularFeatureFlagManager(config, nil)

	// Test with context that has timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	model, err := manager.GetOptimalModel(ctx, "classification")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if model == "" {
		t.Error("Expected model to be returned")
	}

	// Test with cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	model, err = manager.GetOptimalModel(cancelledCtx, "classification")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if model == "" {
		t.Error("Expected fallback model to be returned")
	}
}

func BenchmarkIsServiceEnabled(b *testing.B) {
	config := GranularFeatureFlagConfig{
		Environment: "test",
	}
	manager := NewGranularFeatureFlagManager(config, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.IsServiceEnabled("python_ml_service")
	}
}

func BenchmarkIsModelEnabled(b *testing.B) {
	config := GranularFeatureFlagConfig{
		Environment: "test",
	}
	manager := NewGranularFeatureFlagManager(config, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.IsModelEnabled("bert_classification")
	}
}

func BenchmarkGetOptimalModel(b *testing.B) {
	config := GranularFeatureFlagConfig{
		Environment: "test",
	}
	manager := NewGranularFeatureFlagManager(config, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetOptimalModel(ctx, "classification")
	}
}
