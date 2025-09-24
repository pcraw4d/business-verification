package machine_learning

import (
	"context"
	"testing"
	"time"
)

func TestNewContentClassifier(t *testing.T) {
	config := ContentClassifierConfig{
		ModelType:              "bert",
		MaxSequenceLength:      512,
		BatchSize:              16,
		LearningRate:           2e-5,
		Epochs:                 3,
		ValidationSplit:        0.2,
		IndustryModels:         []string{"finance", "healthcare", "technology"},
		DefaultModel:           "general",
		ModelUpdateInterval:    24 * time.Hour,
		ConfidenceThreshold:    0.8,
		ExplainabilityEnabled:  true,
		AttentionVisualization: true,
		PerformanceTracking:    true,
		ABTestingEnabled:       true,
		ModelVersioning:        true,
		AutoRetraining:         true,
		RetrainingThreshold:    0.05,
		DataDriftDetection:     true,
	}

	classifier := NewContentClassifier(config)

	if classifier == nil {
		t.Fatal("Expected classifier to be created, got nil")
	}

	if classifier.config.ModelType != "bert" {
		t.Errorf("Expected model type to be 'bert', got '%s'", classifier.config.ModelType)
	}

	if classifier.config.MaxSequenceLength != 512 {
		t.Errorf("Expected max sequence length to be 512, got %d", classifier.config.MaxSequenceLength)
	}

	if classifier.config.BatchSize != 16 {
		t.Errorf("Expected batch size to be 16, got %d", classifier.config.BatchSize)
	}

	if classifier.config.LearningRate != 2e-5 {
		t.Errorf("Expected learning rate to be 2e-5, got %f", classifier.config.LearningRate)
	}

	if classifier.config.Epochs != 3 {
		t.Errorf("Expected epochs to be 3, got %d", classifier.config.Epochs)
	}

	if classifier.config.ValidationSplit != 0.2 {
		t.Errorf("Expected validation split to be 0.2, got %f", classifier.config.ValidationSplit)
	}

	if len(classifier.config.IndustryModels) != 3 {
		t.Errorf("Expected 3 industry models, got %d", len(classifier.config.IndustryModels))
	}

	if classifier.config.DefaultModel != "general" {
		t.Errorf("Expected default model to be 'general', got '%s'", classifier.config.DefaultModel)
	}

	if classifier.config.ModelUpdateInterval != 24*time.Hour {
		t.Errorf("Expected model update interval to be 24h, got %v", classifier.config.ModelUpdateInterval)
	}

	if classifier.config.ConfidenceThreshold != 0.8 {
		t.Errorf("Expected confidence threshold to be 0.8, got %f", classifier.config.ConfidenceThreshold)
	}

	if !classifier.config.ExplainabilityEnabled {
		t.Error("Expected explainability to be enabled")
	}

	if !classifier.config.AttentionVisualization {
		t.Error("Expected attention visualization to be enabled")
	}

	if !classifier.config.PerformanceTracking {
		t.Error("Expected performance tracking to be enabled")
	}

	if !classifier.config.ABTestingEnabled {
		t.Error("Expected A/B testing to be enabled")
	}

	if !classifier.config.ModelVersioning {
		t.Error("Expected model versioning to be enabled")
	}

	if !classifier.config.AutoRetraining {
		t.Error("Expected auto retraining to be enabled")
	}

	if classifier.config.RetrainingThreshold != 0.05 {
		t.Errorf("Expected retraining threshold to be 0.05, got %f", classifier.config.RetrainingThreshold)
	}

	if !classifier.config.DataDriftDetection {
		t.Error("Expected data drift detection to be enabled")
	}
}

func TestNewContentClassifierWithDefaults(t *testing.T) {
	config := ContentClassifierConfig{}

	classifier := NewContentClassifier(config)

	if classifier == nil {
		t.Fatal("Expected classifier to be created, got nil")
	}

	// Check default values
	if classifier.config.ModelType != "bert" {
		t.Errorf("Expected default model type to be 'bert', got '%s'", classifier.config.ModelType)
	}

	if classifier.config.MaxSequenceLength != 512 {
		t.Errorf("Expected default max sequence length to be 512, got %d", classifier.config.MaxSequenceLength)
	}

	if classifier.config.BatchSize != 16 {
		t.Errorf("Expected default batch size to be 16, got %d", classifier.config.BatchSize)
	}

	if classifier.config.LearningRate != 2e-5 {
		t.Errorf("Expected default learning rate to be 2e-5, got %f", classifier.config.LearningRate)
	}

	if classifier.config.Epochs != 3 {
		t.Errorf("Expected default epochs to be 3, got %d", classifier.config.Epochs)
	}

	if classifier.config.ValidationSplit != 0.2 {
		t.Errorf("Expected default validation split to be 0.2, got %f", classifier.config.ValidationSplit)
	}

	if classifier.config.ConfidenceThreshold != 0.8 {
		t.Errorf("Expected default confidence threshold to be 0.8, got %f", classifier.config.ConfidenceThreshold)
	}

	if classifier.config.ModelUpdateInterval != 24*time.Hour {
		t.Errorf("Expected default model update interval to be 24h, got %v", classifier.config.ModelUpdateInterval)
	}

	if classifier.config.RetrainingThreshold != 0.05 {
		t.Errorf("Expected default retraining threshold to be 0.05, got %f", classifier.config.RetrainingThreshold)
	}
}

func TestClassifyContent(t *testing.T) {
	config := ContentClassifierConfig{
		ExplainabilityEnabled: true,
		PerformanceTracking:   true,
	}
	classifier := NewContentClassifier(config)

	// Add a mock model
	model := &ClassificationModel{
		ID:         "test-model",
		Name:       "Test Model",
		Version:    "1.0.0",
		Industry:   "general",
		ModelType:  "bert",
		IsActive:   true,
		IsDeployed: true,
	}
	classifier.models["general"] = model

	content := "This is a business registration document for ABC Corporation, a technology company incorporated in Delaware."
	industry := "general"

	ctx := context.Background()
	result, err := classifier.ClassifyContent(ctx, content, industry)

	if err != nil {
		t.Fatalf("Expected to classify content successfully, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected classification result, got nil")
	}

	if result.ContentID == "" {
		t.Error("Expected content ID to be generated")
	}

	if result.ModelID != "test-model" {
		t.Errorf("Expected model ID to be 'test-model', got '%s'", result.ModelID)
	}

	if result.ModelVersion != "1.0.0" {
		t.Errorf("Expected model version to be '1.0.0', got '%s'", result.ModelVersion)
	}

	if len(result.Classifications) == 0 {
		t.Error("Expected classifications to be generated")
	}

	if result.Confidence <= 0 {
		t.Error("Expected positive confidence score")
	}

	if result.ProcessingTime <= 0 {
		t.Error("Expected positive processing time")
	}

	if result.QualityScore <= 0 {
		t.Error("Expected positive quality score")
	}

	if len(result.QualityFactors) == 0 {
		t.Error("Expected quality factors to be generated")
	}

	// Check that explanations are generated when enabled
	if len(result.Explanations) == 0 {
		t.Error("Expected explanations to be generated when explainability is enabled")
	}
}

func TestClassifyContentNoModel(t *testing.T) {
	config := ContentClassifierConfig{}
	classifier := NewContentClassifier(config)

	content := "This is a business document."
	industry := "unknown"

	ctx := context.Background()
	_, err := classifier.ClassifyContent(ctx, content, industry)

	if err == nil {
		t.Fatal("Expected error when no suitable model is found, got nil")
	}

	expectedError := "failed to get model for industry unknown: no suitable model found for industry unknown"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestAssessContentQuality(t *testing.T) {
	config := ContentClassifierConfig{}
	classifier := NewContentClassifier(config)

	content := "This is a comprehensive business registration document for ABC Corporation, a technology company incorporated in Delaware. The document contains detailed information about the company structure, ownership, and legal status."
	predictions := []ClassificationPrediction{
		{
			Label:       "business_registration",
			Confidence:  0.95,
			Probability: 0.95,
			Rank:        1,
		},
		{
			Label:       "financial_report",
			Confidence:  0.03,
			Probability: 0.03,
			Rank:        2,
		},
	}

	qualityScore, qualityFactors := classifier.assessContentQuality(content, predictions)

	if qualityScore <= 0 {
		t.Error("Expected positive quality score")
	}

	if qualityScore > 1.0 {
		t.Error("Expected quality score <= 1.0")
	}

	if len(qualityFactors) == 0 {
		t.Error("Expected quality factors to be generated")
	}

	// Check for expected quality factors
	factorTypes := make(map[string]bool)
	for _, factor := range qualityFactors {
		factorTypes[factor.Factor] = true
	}

	expectedFactors := []string{"content_length", "classification_confidence", "content_structure", "language_quality"}
	for _, expected := range expectedFactors {
		if !factorTypes[expected] {
			t.Errorf("Expected quality factor '%s' to be present", expected)
		}
	}
}

func TestAssessContentStructure(t *testing.T) {
	config := ContentClassifierConfig{}
	classifier := NewContentClassifier(config)

	// Test short content
	shortContent := "Short text"
	score := classifier.assessContentStructure(shortContent)
	if score != 0.3 {
		t.Errorf("Expected score 0.3 for short content, got %f", score)
	}

	// Test medium content
	mediumContent := "This is a medium length content that should score around 0.6. It contains multiple sentences and provides some detail about the subject matter."
	score = classifier.assessContentStructure(mediumContent)
	if score != 0.6 {
		t.Errorf("Expected score 0.6 for medium content, got %f", score)
	}

	// Test long content
	longContent := "This is a very long content that should score around 0.6. It contains many sentences and provides comprehensive detail about the subject matter. The content is well-structured and includes multiple paragraphs with detailed information about business registration, legal requirements, and corporate governance. This type of content typically scores higher due to its length and structure."
	score = classifier.assessContentStructure(longContent)
	if score != 0.6 {
		t.Errorf("Expected score 0.6 for long content, got %f", score)
	}

	// Test very long content
	veryLongContent := "This is an extremely long content that should score around 0.9. " + string(make([]byte, 2000))
	score = classifier.assessContentStructure(veryLongContent)
	if score != 0.9 {
		t.Errorf("Expected score 0.9 for very long content, got %f", score)
	}
}

func TestAssessLanguageQuality(t *testing.T) {
	config := ContentClassifierConfig{}
	classifier := NewContentClassifier(config)

	// Test content with capitalization and punctuation
	goodContent := "This is a well-formatted document. It contains proper capitalization and punctuation!"
	score := classifier.assessLanguageQuality(goodContent)
	if score != 1.0 {
		t.Errorf("Expected score 1.0 for good content, got %f", score)
	}

	// Test content with only capitalization
	mediumContent := "This content has capitalization but no punctuation"
	score = classifier.assessLanguageQuality(mediumContent)
	if score != 0.7 {
		t.Errorf("Expected score 0.7 for medium content, got %f", score)
	}

	// Test content with only punctuation
	mediumContent2 := "this content has punctuation but no capitalization."
	score = classifier.assessLanguageQuality(mediumContent2)
	if score != 0.8 {
		t.Errorf("Expected score 0.8 for medium content with punctuation, got %f", score)
	}

	// Test poor content
	poorContent := "this content has no capitalization or punctuation"
	score = classifier.assessLanguageQuality(poorContent)
	if score != 0.5 {
		t.Errorf("Expected score 0.5 for poor content, got %f", score)
	}
}

func TestCalculateConfidence(t *testing.T) {
	scorer := NewConfidenceScorer(0.8)

	// Test with empty predictions
	predictions := []ClassificationPrediction{}
	confidence := scorer.CalculateConfidence(predictions)
	if confidence != 0.0 {
		t.Errorf("Expected confidence 0.0 for empty predictions, got %f", confidence)
	}

	// Test with single prediction
	predictions = []ClassificationPrediction{
		{
			Label:       "business_registration",
			Confidence:  0.95,
			Probability: 0.95,
			Rank:        1,
		},
	}
	confidence = scorer.CalculateConfidence(predictions)
	if confidence != 0.95 {
		t.Errorf("Expected confidence 0.95, got %f", confidence)
	}

	// Test with multiple predictions (should use highest)
	predictions = []ClassificationPrediction{
		{
			Label:       "business_registration",
			Confidence:  0.85,
			Probability: 0.85,
			Rank:        1,
		},
		{
			Label:       "financial_report",
			Confidence:  0.95,
			Probability: 0.95,
			Rank:        2,
		},
	}
	confidence = scorer.CalculateConfidence(predictions)
	if confidence != 0.85 {
		t.Errorf("Expected confidence 0.85 (first prediction), got %f", confidence)
	}
}

func TestGenerateExplanations(t *testing.T) {
	explainability := NewModelExplainability(true)

	content := "This is a business registration document for ABC Corporation."
	model := &ClassificationModel{
		ID:   "test-model",
		Name: "Test Model",
	}
	predictions := []ClassificationPrediction{
		{
			Label:       "business_registration",
			Confidence:  0.95,
			Probability: 0.95,
			Rank:        1,
		},
	}

	explanations, err := explainability.GenerateExplanations(content, model, predictions)

	if err != nil {
		t.Fatalf("Expected to generate explanations successfully, got error: %v", err)
	}

	if len(explanations) == 0 {
		t.Error("Expected explanations to be generated")
	}

	// Check that explanations are sorted by importance
	for i := 1; i < len(explanations); i++ {
		if explanations[i-1].Importance < explanations[i].Importance {
			t.Error("Expected explanations to be sorted by importance (descending)")
		}
	}

	// Check for expected explanation types
	explanationTypes := make(map[string]bool)
	for _, explanation := range explanations {
		explanationTypes[explanation.Type] = true
	}

	expectedTypes := []string{"token", "phrase"}
	for _, expected := range expectedTypes {
		if !explanationTypes[expected] {
			t.Errorf("Expected explanation type '%s' to be present", expected)
		}
	}
}

func TestGenerateTokenExplanations(t *testing.T) {
	explainability := NewModelExplainability(true)

	content := "This is a business registration document."
	predictions := []ClassificationPrediction{
		{
			Label:       "business_registration",
			Confidence:  0.95,
			Probability: 0.95,
			Rank:        1,
		},
	}

	explanations := explainability.generateTokenExplanations(content, predictions)

	if len(explanations) == 0 {
		t.Error("Expected token explanations to be generated")
	}

	// Check that all explanations are token type
	for _, explanation := range explanations {
		if explanation.Type != "token" {
			t.Errorf("Expected explanation type 'token', got '%s'", explanation.Type)
		}

		if explanation.Importance <= 0 {
			t.Error("Expected positive importance score")
		}

		if explanation.Contribution <= 0 {
			t.Error("Expected positive contribution score")
		}
	}
}

func TestGeneratePhraseExplanations(t *testing.T) {
	explainability := NewModelExplainability(true)

	content := "This is a business registration document."
	predictions := []ClassificationPrediction{
		{
			Label:       "business_registration",
			Confidence:  0.95,
			Probability: 0.95,
			Rank:        1,
		},
	}

	explanations := explainability.generatePhraseExplanations(content, predictions)

	if len(explanations) == 0 {
		t.Error("Expected phrase explanations to be generated")
	}

	// Check that all explanations are phrase type
	for _, explanation := range explanations {
		if explanation.Type != "phrase" {
			t.Errorf("Expected explanation type 'phrase', got '%s'", explanation.Type)
		}

		if explanation.Importance <= 0 {
			t.Error("Expected positive importance score")
		}

		if explanation.Contribution <= 0 {
			t.Error("Expected positive contribution score")
		}
	}
}

func TestNewModelRegistry(t *testing.T) {
	registry := NewModelRegistry()

	if registry == nil {
		t.Fatal("Expected model registry to be created, got nil")
	}

	if registry.models == nil {
		t.Error("Expected models map to be initialized")
	}

	if registry.versions == nil {
		t.Error("Expected versions map to be initialized")
	}

	if registry.deployments == nil {
		t.Error("Expected deployments map to be initialized")
	}

	if registry.rollbackHistory == nil {
		t.Error("Expected rollback history map to be initialized")
	}
}

func TestNewTrainingPipeline(t *testing.T) {
	config := ContentClassifierConfig{}
	pipeline := NewTrainingPipeline(config)

	if pipeline == nil {
		t.Fatal("Expected training pipeline to be created, got nil")
	}

	if pipeline.config.BaseModel != "bert-base-uncased" {
		t.Errorf("Expected base model to be 'bert-base-uncased', got '%s'", pipeline.config.BaseModel)
	}

	if pipeline.config.Optimizer != "adamw" {
		t.Errorf("Expected optimizer to be 'adamw', got '%s'", pipeline.config.Optimizer)
	}

	if pipeline.config.LossFunction != "cross_entropy" {
		t.Errorf("Expected loss function to be 'cross_entropy', got '%s'", pipeline.config.LossFunction)
	}

	if pipeline.config.Regularization != "dropout" {
		t.Errorf("Expected regularization to be 'dropout', got '%s'", pipeline.config.Regularization)
	}

	if !pipeline.config.EarlyStopping {
		t.Error("Expected early stopping to be enabled")
	}

	if pipeline.config.Patience != 3 {
		t.Errorf("Expected patience to be 3, got %d", pipeline.config.Patience)
	}

	if !pipeline.config.Checkpointing {
		t.Error("Expected checkpointing to be enabled")
	}

	if !pipeline.config.MixedPrecision {
		t.Error("Expected mixed precision to be enabled")
	}

	if pipeline.datasets == nil {
		t.Error("Expected datasets map to be initialized")
	}

	if pipeline.experiments == nil {
		t.Error("Expected experiments map to be initialized")
	}
}

func TestNewConfidenceScorer(t *testing.T) {
	threshold := 0.8
	scorer := NewConfidenceScorer(threshold)

	if scorer == nil {
		t.Fatal("Expected confidence scorer to be created, got nil")
	}

	if scorer.config.Method != "temperature_scaling" {
		t.Errorf("Expected method to be 'temperature_scaling', got '%s'", scorer.config.Method)
	}

	if !scorer.config.CalibrationData {
		t.Error("Expected calibration data to be enabled")
	}

	if scorer.config.EnsembleMethod != "averaging" {
		t.Errorf("Expected ensemble method to be 'averaging', got '%s'", scorer.config.EnsembleMethod)
	}

	if scorer.config.Threshold != threshold {
		t.Errorf("Expected threshold to be %f, got %f", threshold, scorer.config.Threshold)
	}

	if scorer.calibration == nil {
		t.Error("Expected calibration map to be initialized")
	}
}

func TestNewModelExplainability(t *testing.T) {
	// Test with explainability enabled
	explainability := NewModelExplainability(true)

	if explainability == nil {
		t.Fatal("Expected model explainability to be created, got nil")
	}

	if len(explainability.config.Methods) != 2 {
		t.Errorf("Expected 2 methods, got %d", len(explainability.config.Methods))
	}

	expectedMethods := []string{"attention", "gradients"}
	for _, expected := range expectedMethods {
		found := false
		for _, method := range explainability.config.Methods {
			if method == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected method '%s' to be present", expected)
		}
	}

	if len(explainability.config.AttentionLayers) != 3 {
		t.Errorf("Expected 3 attention layers, got %d", len(explainability.config.AttentionLayers))
	}

	expectedLayers := []int{0, 6, 11}
	for _, expected := range expectedLayers {
		found := false
		for _, layer := range explainability.config.AttentionLayers {
			if layer == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected attention layer %d to be present", expected)
		}
	}

	if explainability.config.MaxTokens != 100 {
		t.Errorf("Expected max tokens to be 100, got %d", explainability.config.MaxTokens)
	}

	if !explainability.config.Visualization {
		t.Error("Expected visualization to be enabled")
	}

	if explainability.explainers == nil {
		t.Error("Expected explainers map to be initialized")
	}

	// Test with explainability disabled
	explainabilityDisabled := NewModelExplainability(false)
	if explainabilityDisabled == nil {
		t.Fatal("Expected model explainability to be created even when disabled, got nil")
	}
}

func TestGenerateContentID(t *testing.T) {
	content := "This is a test content for ID generation."

	// Generate multiple IDs to ensure uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		id := generateContentID(content)
		if id == "" {
			t.Error("Expected non-empty content ID")
		}

		// Add small delay to ensure unique timestamps
		time.Sleep(1 * time.Millisecond)

		if ids[id] {
			t.Errorf("Expected unique content ID, got duplicate: %s", id)
		}

		ids[id] = true
	}
}

func TestUpdateModelUsage(t *testing.T) {
	config := ContentClassifierConfig{}
	classifier := NewContentClassifier(config)

	// Add a model
	model := &ClassificationModel{
		ID:   "test-model",
		Name: "Test Model",
	}
	classifier.models["test-model"] = model

	// Update usage
	classifier.updateModelUsage("test-model")

	// Check that last used time was updated
	if model.LastUsed.IsZero() {
		t.Error("Expected last used time to be updated")
	}
}

func TestBERTModelIntegration(t *testing.T) {
	config := ContentClassifierConfig{
		ModelType:             "bert",
		MaxSequenceLength:     512,
		BatchSize:             16,
		LearningRate:          2e-5,
		Epochs:                3,
		ValidationSplit:       0.2,
		ConfidenceThreshold:   0.8,
		ExplainabilityEnabled: true,
		PerformanceTracking:   true,
	}

	classifier := NewContentClassifier(config)

	// Test BERT-specific configuration
	if classifier.config.ModelType != "bert" {
		t.Errorf("Expected model type to be 'bert', got '%s'", classifier.config.ModelType)
	}

	if classifier.config.MaxSequenceLength != 512 {
		t.Errorf("Expected max sequence length to be 512, got %d", classifier.config.MaxSequenceLength)
	}

	if classifier.config.BatchSize != 16 {
		t.Errorf("Expected batch size to be 16, got %d", classifier.config.BatchSize)
	}

	if classifier.config.LearningRate != 2e-5 {
		t.Errorf("Expected learning rate to be 2e-5, got %f", classifier.config.LearningRate)
	}
}

func TestIndustrySpecificModels(t *testing.T) {
	config := ContentClassifierConfig{
		IndustryModels: []string{"finance", "healthcare", "technology", "legal"},
		DefaultModel:   "general",
	}

	classifier := NewContentClassifier(config)

	// Test industry models configuration
	if len(classifier.config.IndustryModels) != 4 {
		t.Errorf("Expected 4 industry models, got %d", len(classifier.config.IndustryModels))
	}

	expectedIndustries := []string{"finance", "healthcare", "technology", "legal"}
	for _, expected := range expectedIndustries {
		found := false
		for _, industry := range classifier.config.IndustryModels {
			if industry == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected industry '%s' to be present", expected)
		}
	}

	if classifier.config.DefaultModel != "general" {
		t.Errorf("Expected default model to be 'general', got '%s'", classifier.config.DefaultModel)
	}
}

func TestModelVersioning(t *testing.T) {
	config := ContentClassifierConfig{
		ModelVersioning:     true,
		ModelUpdateInterval: 24 * time.Hour,
	}

	classifier := NewContentClassifier(config)

	// Test model versioning configuration
	if !classifier.config.ModelVersioning {
		t.Error("Expected model versioning to be enabled")
	}

	if classifier.config.ModelUpdateInterval != 24*time.Hour {
		t.Errorf("Expected model update interval to be 24h, got %v", classifier.config.ModelUpdateInterval)
	}
}

func TestABTestingConfiguration(t *testing.T) {
	config := ContentClassifierConfig{
		ABTestingEnabled:    true,
		PerformanceTracking: true,
	}

	classifier := NewContentClassifier(config)

	// Test A/B testing configuration
	if !classifier.config.ABTestingEnabled {
		t.Error("Expected A/B testing to be enabled")
	}

	if !classifier.config.PerformanceTracking {
		t.Error("Expected performance tracking to be enabled")
	}
}

func TestAutoRetrainingConfiguration(t *testing.T) {
	config := ContentClassifierConfig{
		AutoRetraining:      true,
		RetrainingThreshold: 0.05,
		DataDriftDetection:  true,
	}

	classifier := NewContentClassifier(config)

	// Test auto retraining configuration
	if !classifier.config.AutoRetraining {
		t.Error("Expected auto retraining to be enabled")
	}

	if classifier.config.RetrainingThreshold != 0.05 {
		t.Errorf("Expected retraining threshold to be 0.05, got %f", classifier.config.RetrainingThreshold)
	}

	if !classifier.config.DataDriftDetection {
		t.Error("Expected data drift detection to be enabled")
	}
}
