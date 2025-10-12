package testing

import (
	"context"
	"math"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestABTestManager_CreateExperiment(t *testing.T) {
	logger := zap.NewNop()
	abm := NewABTestManager(logger)

	config := &ExperimentConfig{
		ID:          "test_experiment",
		Name:        "Test Experiment",
		Description: "A test experiment",
		TrafficSplit: map[string]float64{
			"model_a": 0.5,
			"model_b": 0.5,
		},
		Models: map[string]ModelConfig{
			"model_a": {
				ID:          "model_a",
				Name:        "Model A",
				Type:        "xgboost",
				Version:     "1.0",
				Description: "Test model A",
			},
			"model_b": {
				ID:          "model_b",
				Name:        "Model B",
				Type:        "lstm",
				Version:     "1.0",
				Description: "Test model B",
			},
		},
		SuccessMetrics:  []string{"accuracy", "f1_score"},
		MinSampleSize:   1000,
		ConfidenceLevel: 0.95,
	}

	experiment, err := abm.CreateExperiment(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create experiment: %v", err)
	}

	if experiment.ID != config.ID {
		t.Errorf("Expected experiment ID %s, got %s", config.ID, experiment.ID)
	}

	if experiment.Status != StatusDraft {
		t.Errorf("Expected status %s, got %s", StatusDraft, experiment.Status)
	}

	if len(experiment.Models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(experiment.Models))
	}
}

func TestABTestManager_InvalidTrafficSplit(t *testing.T) {
	logger := zap.NewNop()
	abm := NewABTestManager(logger)

	config := &ExperimentConfig{
		ID:          "test_experiment",
		Name:        "Test Experiment",
		Description: "A test experiment",
		TrafficSplit: map[string]float64{
			"model_a": 0.3,
			"model_b": 0.3, // Total is 0.6, not 1.0
		},
		Models: map[string]ModelConfig{
			"model_a": {ID: "model_a", Name: "Model A", Type: "xgboost"},
			"model_b": {ID: "model_b", Name: "Model B", Type: "lstm"},
		},
		SuccessMetrics:  []string{"accuracy"},
		MinSampleSize:   1000,
		ConfidenceLevel: 0.95,
	}

	_, err := abm.CreateExperiment(context.Background(), config)
	if err == nil {
		t.Fatal("Expected error for invalid traffic split, got nil")
	}
}

func TestABTestManager_StartStopExperiment(t *testing.T) {
	logger := zap.NewNop()
	abm := NewABTestManager(logger)

	// Create experiment
	config := &ExperimentConfig{
		ID:          "test_experiment",
		Name:        "Test Experiment",
		Description: "A test experiment",
		TrafficSplit: map[string]float64{
			"model_a": 0.5,
			"model_b": 0.5,
		},
		Models: map[string]ModelConfig{
			"model_a": {ID: "model_a", Name: "Model A", Type: "xgboost"},
			"model_b": {ID: "model_b", Name: "Model B", Type: "lstm"},
		},
		SuccessMetrics:  []string{"accuracy"},
		MinSampleSize:   1000,
		ConfidenceLevel: 0.95,
	}

	experiment, err := abm.CreateExperiment(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create experiment: %v", err)
	}

	// Start experiment
	err = abm.StartExperiment(context.Background(), experiment.ID)
	if err != nil {
		t.Fatalf("Failed to start experiment: %v", err)
	}

	// Verify status
	experiment, err = abm.GetExperiment(experiment.ID)
	if err != nil {
		t.Fatalf("Failed to get experiment: %v", err)
	}

	if experiment.Status != StatusRunning {
		t.Errorf("Expected status %s, got %s", StatusRunning, experiment.Status)
	}

	// Stop experiment
	err = abm.StopExperiment(context.Background(), experiment.ID)
	if err != nil {
		t.Fatalf("Failed to stop experiment: %v", err)
	}

	// Verify status
	experiment, err = abm.GetExperiment(experiment.ID)
	if err != nil {
		t.Fatalf("Failed to get experiment: %v", err)
	}

	if experiment.Status != StatusCompleted {
		t.Errorf("Expected status %s, got %s", StatusCompleted, experiment.Status)
	}
}

func TestABTestManager_SelectModel(t *testing.T) {
	logger := zap.NewNop()
	abm := NewABTestManager(logger)

	// Create and start experiment
	config := &ExperimentConfig{
		ID:          "test_experiment",
		Name:        "Test Experiment",
		Description: "A test experiment",
		TrafficSplit: map[string]float64{
			"model_a": 0.5,
			"model_b": 0.5,
		},
		Models: map[string]ModelConfig{
			"model_a": {ID: "model_a", Name: "Model A", Type: "xgboost"},
			"model_b": {ID: "model_b", Name: "Model B", Type: "lstm"},
		},
		SuccessMetrics:  []string{"accuracy"},
		MinSampleSize:   1000,
		ConfidenceLevel: 0.95,
	}

	experiment, err := abm.CreateExperiment(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create experiment: %v", err)
	}

	err = abm.StartExperiment(context.Background(), experiment.ID)
	if err != nil {
		t.Fatalf("Failed to start experiment: %v", err)
	}

	// Test model selection consistency
	requestID := "test_request_1"
	modelID1, err := abm.SelectModel(context.Background(), experiment.ID, requestID)
	if err != nil {
		t.Fatalf("Failed to select model: %v", err)
	}

	// Same request ID should always return same model
	modelID2, err := abm.SelectModel(context.Background(), experiment.ID, requestID)
	if err != nil {
		t.Fatalf("Failed to select model: %v", err)
	}

	if modelID1 != modelID2 {
		t.Errorf("Expected consistent model selection, got %s and %s", modelID1, modelID2)
	}

	// Different request ID might return different model
	modelID3, err := abm.SelectModel(context.Background(), experiment.ID, "test_request_2")
	if err != nil {
		t.Fatalf("Failed to select model: %v", err)
	}

	// Verify model ID is valid
	if modelID1 != "model_a" && modelID1 != "model_b" {
		t.Errorf("Expected model_a or model_b, got %s", modelID1)
	}

	if modelID3 != "model_a" && modelID3 != "model_b" {
		t.Errorf("Expected model_a or model_b, got %s", modelID3)
	}
}

func TestABTestManager_RecordPrediction(t *testing.T) {
	logger := zap.NewNop()
	abm := NewABTestManager(logger)

	// Create and start experiment
	config := &ExperimentConfig{
		ID:          "test_experiment",
		Name:        "Test Experiment",
		Description: "A test experiment",
		TrafficSplit: map[string]float64{
			"model_a": 1.0, // Use only one model for testing
		},
		Models: map[string]ModelConfig{
			"model_a": {ID: "model_a", Name: "Model A", Type: "xgboost"},
		},
		SuccessMetrics:  []string{"accuracy"},
		MinSampleSize:   1000,
		ConfidenceLevel: 0.95,
	}

	experiment, err := abm.CreateExperiment(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create experiment: %v", err)
	}

	err = abm.StartExperiment(context.Background(), experiment.ID)
	if err != nil {
		t.Fatalf("Failed to start experiment: %v", err)
	}

	// Record predictions
	predictions := []*PredictionRecord{
		{
			RequestID:    "req_1",
			ModelID:      "model_a",
			ExperimentID: experiment.ID,
			Input:        map[string]interface{}{"feature1": 1.0},
			Prediction:   0.8,
			Confidence:   0.9,
			Latency:      100 * time.Millisecond,
			Timestamp:    time.Now(),
			IsError:      false,
		},
		{
			RequestID:    "req_2",
			ModelID:      "model_a",
			ExperimentID: experiment.ID,
			Input:        map[string]interface{}{"feature1": 2.0},
			Prediction:   0.7,
			Confidence:   0.8,
			Latency:      120 * time.Millisecond,
			Timestamp:    time.Now(),
			IsError:      false,
		},
	}

	for _, prediction := range predictions {
		err := abm.RecordPrediction(context.Background(), experiment.ID, "model_a", prediction)
		if err != nil {
			t.Fatalf("Failed to record prediction: %v", err)
		}
	}

	// Stop experiment to get results
	err = abm.StopExperiment(context.Background(), experiment.ID)
	if err != nil {
		t.Fatalf("Failed to stop experiment: %v", err)
	}

	// Get results
	results, err := abm.GetExperimentResults(context.Background(), experiment.ID)
	if err != nil {
		t.Fatalf("Failed to get experiment results: %v", err)
	}

	if results.TotalRequests != 2 {
		t.Errorf("Expected 2 total requests, got %d", results.TotalRequests)
	}

	modelResult, exists := results.ModelResults["model_a"]
	if !exists {
		t.Fatal("Expected model_a results")
	}

	if modelResult.RequestCount != 2 {
		t.Errorf("Expected 2 requests for model_a, got %d", modelResult.RequestCount)
	}

	if modelResult.ErrorRate != 0 {
		t.Errorf("Expected 0 error rate for model_a, got %f", modelResult.ErrorRate)
	}
}

func TestABTestManager_StatisticalTest(t *testing.T) {
	logger := zap.NewNop()
	abm := NewABTestManager(logger)

	// Test statistical test calculation
	modelResults := map[string]*ModelResult{
		"model_a": {
			ModelID:      "model_a",
			RequestCount: 1000,
			Metrics: map[string]float64{
				"accuracy": 0.85,
			},
		},
		"model_b": {
			ModelID:      "model_b",
			RequestCount: 1000,
			Metrics: map[string]float64{
				"accuracy": 0.90,
			},
		},
	}

	statisticalTest, err := abm.performStatisticalTest(modelResults, "accuracy")
	if err != nil {
		t.Fatalf("Failed to perform statistical test: %v", err)
	}

	if statisticalTest == nil {
		t.Fatal("Expected statistical test result")
	}

	if statisticalTest.TestType != "t-test" {
		t.Errorf("Expected test type 't-test', got %s", statisticalTest.TestType)
	}

	// Test winner determination
	winner := abm.determineWinner(modelResults, "accuracy")
	if winner != "model_b" {
		t.Errorf("Expected winner 'model_b', got %s", winner)
	}
}

func TestABTestManager_Recommendation(t *testing.T) {
	logger := zap.NewNop()
	abm := NewABTestManager(logger)

	// Test recommendation generation
	modelResults := map[string]*ModelResult{
		"model_a": {
			ModelID:      "model_a",
			RequestCount: 1000,
			Metrics: map[string]float64{
				"accuracy": 0.85,
			},
		},
		"model_b": {
			ModelID:      "model_b",
			RequestCount: 1000,
			Metrics: map[string]float64{
				"accuracy": 0.90,
			},
		},
	}

	statisticalTest := &StatisticalTestResult{
		TestType:      "t-test",
		PValue:        0.01,
		IsSignificant: true,
		EffectSize:    0.5,
	}

	winner := "model_b"
	recommendation := abm.generateRecommendation(modelResults, statisticalTest, winner)

	if recommendation == "" {
		t.Fatal("Expected non-empty recommendation")
	}

	// Test with non-significant result
	statisticalTest.IsSignificant = false
	recommendation = abm.generateRecommendation(modelResults, statisticalTest, winner)

	if recommendation == "" {
		t.Fatal("Expected non-empty recommendation")
	}
}

func TestABTestManager_ListExperiments(t *testing.T) {
	logger := zap.NewNop()
	abm := NewABTestManager(logger)

	// Create multiple experiments
	configs := []*ExperimentConfig{
		{
			ID:              "exp_1",
			Name:            "Experiment 1",
			Description:     "First experiment",
			TrafficSplit:    map[string]float64{"model_a": 1.0},
			Models:          map[string]ModelConfig{"model_a": {ID: "model_a", Name: "Model A", Type: "xgboost"}},
			SuccessMetrics:  []string{"accuracy"},
			MinSampleSize:   1000,
			ConfidenceLevel: 0.95,
		},
		{
			ID:              "exp_2",
			Name:            "Experiment 2",
			Description:     "Second experiment",
			TrafficSplit:    map[string]float64{"model_b": 1.0},
			Models:          map[string]ModelConfig{"model_b": {ID: "model_b", Name: "Model B", Type: "lstm"}},
			SuccessMetrics:  []string{"f1_score"},
			MinSampleSize:   1000,
			ConfidenceLevel: 0.95,
		},
	}

	for _, config := range configs {
		_, err := abm.CreateExperiment(context.Background(), config)
		if err != nil {
			t.Fatalf("Failed to create experiment: %v", err)
		}
	}

	// List experiments
	experiments := abm.ListExperiments()
	if len(experiments) != 2 {
		t.Errorf("Expected 2 experiments, got %d", len(experiments))
	}

	// Verify experiment IDs
	ids := make(map[string]bool)
	for _, exp := range experiments {
		ids[exp.ID] = true
	}

	if !ids["exp_1"] || !ids["exp_2"] {
		t.Error("Expected experiments exp_1 and exp_2")
	}
}

func TestABTestManager_ErrorHandling(t *testing.T) {
	logger := zap.NewNop()
	abm := NewABTestManager(logger)

	// Test getting non-existent experiment
	_, err := abm.GetExperiment("non_existent")
	if err == nil {
		t.Fatal("Expected error for non-existent experiment")
	}

	// Test starting non-existent experiment
	err = abm.StartExperiment(context.Background(), "non_existent")
	if err == nil {
		t.Fatal("Expected error for starting non-existent experiment")
	}

	// Test selecting model for non-existent experiment
	_, err = abm.SelectModel(context.Background(), "non_existent", "request_1")
	if err == nil {
		t.Fatal("Expected error for selecting model in non-existent experiment")
	}

	// Test getting results for non-existent experiment
	_, err = abm.GetExperimentResults(context.Background(), "non_existent")
	if err == nil {
		t.Fatal("Expected error for getting results of non-existent experiment")
	}
}

func TestHashString(t *testing.T) {
	// Test hash consistency
	hash1 := hashString("test_string")
	hash2 := hashString("test_string")

	if hash1 != hash2 {
		t.Error("Expected consistent hash for same string")
	}

	// Test hash difference
	hash3 := hashString("different_string")
	if hash1 == hash3 {
		t.Error("Expected different hash for different strings")
	}
}

func TestNormalCDF(t *testing.T) {
	logger := zap.NewNop()
	abm := NewABTestManager(logger)

	// Test normal CDF values
	testCases := []struct {
		input     float64
		expected  float64
		tolerance float64
	}{
		{0.0, 0.5, 0.1},
		{1.0, 0.84, 0.1},
		{-1.0, 0.16, 0.1},
	}

	for _, tc := range testCases {
		result := abm.normalCDF(tc.input)
		if math.Abs(result-tc.expected) > tc.tolerance {
			t.Errorf("Expected normalCDF(%.1f) ≈ %.2f, got %.2f", tc.input, tc.expected, result)
		}
	}
}

func TestEffectSizeCalculation(t *testing.T) {
	logger := zap.NewNop()
	abm := NewABTestManager(logger)

	// Test effect size calculation
	values := []float64{0.85, 0.90}
	effectSize := abm.calculateEffectSize(values)

	if effectSize <= 0 {
		t.Error("Expected positive effect size")
	}

	// Test with identical values
	identicalValues := []float64{0.85, 0.85}
	effectSize = abm.calculateEffectSize(identicalValues)

	if effectSize != 0 {
		t.Errorf("Expected effect size 0 for identical values, got %f", effectSize)
	}
}

func TestConfidenceIntervalCalculation(t *testing.T) {
	logger := zap.NewNop()
	abm := NewABTestManager(logger)

	// Test confidence interval calculation
	values := []float64{0.85, 0.90}
	ci := abm.calculateConfidenceInterval(values)

	if ci[0] >= ci[1] {
		t.Error("Expected lower bound < upper bound")
	}

	if ci[0] < 0 || ci[1] > 1 {
		t.Error("Expected confidence interval bounds between 0 and 1")
	}
}
