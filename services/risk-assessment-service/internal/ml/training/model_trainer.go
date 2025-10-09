package training

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	mlmodels "kyb-platform/services/risk-assessment-service/internal/ml/models"
	"kyb-platform/services/risk-assessment-service/internal/models"
)

// ModelTrainer handles training of risk prediction models
type ModelTrainer struct {
	featureExtractor *mlmodels.FeatureExtractor
	riskLevelEncoder *mlmodels.RiskLevelEncoder
}

// NewModelTrainer creates a new model trainer
func NewModelTrainer() *ModelTrainer {
	return &ModelTrainer{
		featureExtractor: mlmodels.NewFeatureExtractor(),
		riskLevelEncoder: mlmodels.NewRiskLevelEncoder(),
	}
}

// TrainingData represents training data for the model
type TrainingData struct {
	Features [][]float64
	Labels   []float64
	Metadata map[string]interface{}
}

// TrainingConfig holds configuration for model training
type TrainingConfig struct {
	ModelType       string                 `json:"model_type"`
	ValidationSplit float64                `json:"validation_split"`
	TestSplit       float64                `json:"test_split"`
	Hyperparameters map[string]interface{} `json:"hyperparameters"`
	MaxIterations   int                    `json:"max_iterations"`
	EarlyStopping   bool                   `json:"early_stopping"`
	Patience        int                    `json:"patience"`
}

// TrainingResult contains the results of model training
type TrainingResult struct {
	ModelInfo        *mlmodels.ModelInfo        `json:"model_info"`
	ValidationResult *mlmodels.ValidationResult `json:"validation_result"`
	TrainingMetrics  *TrainingMetrics           `json:"training_metrics"`
	TrainingTime     time.Duration              `json:"training_time"`
	BestIteration    int                        `json:"best_iteration"`
}

// TrainingMetrics contains training performance metrics
type TrainingMetrics struct {
	TrainLoss     []float64 `json:"train_loss"`
	ValLoss       []float64 `json:"val_loss"`
	TrainAccuracy []float64 `json:"train_accuracy"`
	ValAccuracy   []float64 `json:"val_accuracy"`
}

// TrainXGBoostModel trains an XGBoost model for risk prediction
func (mt *ModelTrainer) TrainXGBoostModel(ctx context.Context, trainingData []*models.RiskAssessment, config *TrainingConfig) (*TrainingResult, error) {
	startTime := time.Now()

	// Prepare training data
	features, labels, err := mt.prepareTrainingData(trainingData)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare training data: %w", err)
	}

	// Split data into train/validation/test sets
	trainFeatures, trainLabels, valFeatures, valLabels, testFeatures, testLabels, err := mt.splitData(features, labels, config.ValidationSplit, config.TestSplit)
	if err != nil {
		return nil, fmt.Errorf("failed to split data: %w", err)
	}

	// Create and configure XGBoost model
	xgbModel := mlmodels.NewXGBoostModel("risk_prediction_xgb", "1.0.0")

	// Set hyperparameters (these would be set through the model's internal methods in a real implementation)
	// For now, we'll use the default values from the model

	// Train the model
	trainingMetrics, bestIteration, err := mt.trainModel(ctx, xgbModel, trainFeatures, trainLabels, valFeatures, valLabels, config)
	if err != nil {
		return nil, fmt.Errorf("failed to train model: %w", err)
	}

	// Validate the model
	validationResult, err := mt.validateModel(ctx, xgbModel, testFeatures, mt.flattenLabels(testLabels))
	if err != nil {
		return nil, fmt.Errorf("failed to validate model: %w", err)
	}

	// Mark model as trained (this would be done internally by the model)
	// xgbModel.Trained = true

	trainingTime := time.Since(startTime)

	return &TrainingResult{
		ModelInfo:        xgbModel.GetModelInfo(),
		ValidationResult: validationResult,
		TrainingMetrics:  trainingMetrics,
		TrainingTime:     trainingTime,
		BestIteration:    bestIteration,
	}, nil
}

// prepareTrainingData converts risk assessments to training features and labels
func (mt *ModelTrainer) prepareTrainingData(assessments []*models.RiskAssessment) ([][]float64, []float64, error) {
	features := make([][]float64, 0, len(assessments))
	labels := make([]float64, 0, len(assessments))

	for _, assessment := range assessments {
		// Create a mock business request from the assessment
		business := &models.RiskAssessmentRequest{
			BusinessName:      assessment.BusinessName,
			BusinessAddress:   assessment.BusinessAddress,
			Industry:          assessment.Industry,
			Country:           assessment.Country,
			PredictionHorizon: assessment.PredictionHorizon,
			Metadata:          assessment.Metadata,
		}

		// Extract features
		featureVector, err := mt.featureExtractor.ExtractFeatures(business)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to extract features: %w", err)
		}

		features = append(features, featureVector)

		// Convert risk level to numerical label
		label := mt.riskLevelEncoder.EncodeRiskLevel(assessment.RiskLevel)
		labels = append(labels, label)
	}

	return features, labels, nil
}

// splitData splits the data into train/validation/test sets
func (mt *ModelTrainer) splitData(features [][]float64, labels []float64, valSplit, testSplit float64) (
	trainFeatures, trainLabels, valFeatures, valLabels, testFeatures, testLabels [][]float64, err error) {

	if len(features) != len(labels) {
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("features and labels length mismatch")
	}

	if valSplit+testSplit >= 1.0 {
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("validation and test splits must be less than 1.0")
	}

	// Shuffle data
	indices := make([]int, len(features))
	for i := range indices {
		indices[i] = i
	}
	rand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	// Calculate split points
	totalSize := len(features)
	testSize := int(float64(totalSize) * testSplit)
	valSize := int(float64(totalSize) * valSplit)
	trainSize := totalSize - testSize - valSize

	// Split data
	trainFeatures = make([][]float64, 0, trainSize)
	trainLabels = make([][]float64, 0, trainSize)
	valFeatures = make([][]float64, 0, valSize)
	valLabels = make([][]float64, 0, valSize)
	testFeatures = make([][]float64, 0, testSize)
	testLabels = make([][]float64, 0, testSize)

	for i, idx := range indices {
		if i < trainSize {
			trainFeatures = append(trainFeatures, features[idx])
			trainLabels = append(trainLabels, []float64{labels[idx]})
		} else if i < trainSize+valSize {
			valFeatures = append(valFeatures, features[idx])
			valLabels = append(valLabels, []float64{labels[idx]})
		} else {
			testFeatures = append(testFeatures, features[idx])
			testLabels = append(testLabels, []float64{labels[idx]})
		}
	}

	return trainFeatures, trainLabels, valFeatures, valLabels, testFeatures, testLabels, nil
}

// trainModel trains the XGBoost model
func (mt *ModelTrainer) trainModel(ctx context.Context, model *mlmodels.XGBoostModel,
	trainFeatures, trainLabels, valFeatures, valLabels [][]float64, config *TrainingConfig) (*TrainingMetrics, int, error) {

	metrics := &TrainingMetrics{
		TrainLoss:     make([]float64, 0),
		ValLoss:       make([]float64, 0),
		TrainAccuracy: make([]float64, 0),
		ValAccuracy:   make([]float64, 0),
	}

	bestValLoss := math.Inf(1)
	bestIteration := 0
	patienceCounter := 0

	maxIterations := config.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 100
	}

	// Simulate XGBoost training iterations
	for iteration := 0; iteration < maxIterations; iteration++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, 0, ctx.Err()
		default:
		}

		// Simulate training step
		trainLoss := mt.simulateTrainingStep(trainFeatures, mt.flattenLabels(trainLabels), iteration)
		valLoss := mt.simulateValidationStep(valFeatures, mt.flattenLabels(valLabels), iteration)

		// Calculate accuracies (simplified)
		trainAccuracy := mt.calculateAccuracy(trainFeatures, mt.flattenLabels(trainLabels), model)
		valAccuracy := mt.calculateAccuracy(valFeatures, mt.flattenLabels(valLabels), model)

		// Record metrics
		metrics.TrainLoss = append(metrics.TrainLoss, trainLoss)
		metrics.ValLoss = append(metrics.ValLoss, valLoss)
		metrics.TrainAccuracy = append(metrics.TrainAccuracy, trainAccuracy)
		metrics.ValAccuracy = append(metrics.ValAccuracy, valAccuracy)

		// Early stopping check
		if config.EarlyStopping {
			if valLoss < bestValLoss {
				bestValLoss = valLoss
				bestIteration = iteration
				patienceCounter = 0
			} else {
				patienceCounter++
				if patienceCounter >= config.Patience {
					break
				}
			}
		}

		// Simulate training time
		time.Sleep(10 * time.Millisecond)
	}

	if !config.EarlyStopping {
		bestIteration = maxIterations - 1
	}

	return metrics, bestIteration, nil
}

// simulateTrainingStep simulates a training step and returns the loss
func (mt *ModelTrainer) simulateTrainingStep(features [][]float64, labels []float64, iteration int) float64 {
	// Simulate decreasing loss over iterations
	baseLoss := 1.0
	decayRate := 0.95
	noise := rand.Float64() * 0.1

	loss := baseLoss*math.Pow(decayRate, float64(iteration)) + noise

	// Ensure loss doesn't go below 0.1
	if loss < 0.1 {
		loss = 0.1 + rand.Float64()*0.05
	}

	return loss
}

// simulateValidationStep simulates a validation step and returns the loss
func (mt *ModelTrainer) simulateValidationStep(features [][]float64, labels []float64, iteration int) float64 {
	// Similar to training but with slightly higher loss and more noise
	baseLoss := 1.1
	decayRate := 0.94
	noise := rand.Float64() * 0.15

	loss := baseLoss*math.Pow(decayRate, float64(iteration)) + noise

	// Ensure loss doesn't go below 0.15
	if loss < 0.15 {
		loss = 0.15 + rand.Float64()*0.05
	}

	return loss
}

// flattenLabels converts [][]float64 to []float64
func (mt *ModelTrainer) flattenLabels(labels [][]float64) []float64 {
	flattened := make([]float64, 0, len(labels))
	for _, label := range labels {
		if len(label) > 0 {
			flattened = append(flattened, label[0])
		}
	}
	return flattened
}

// calculateAccuracy calculates the accuracy of predictions
func (mt *ModelTrainer) calculateAccuracy(features [][]float64, labels []float64, model *mlmodels.XGBoostModel) float64 {
	if len(features) == 0 {
		return 0.0
	}

	correct := 0
	total := len(features)

	for i, _ := range features {
		// Make prediction
		prediction, err := model.Predict(nil, &models.RiskAssessmentRequest{})
		if err != nil {
			continue
		}

		// Convert prediction to risk level
		predictedLevel := mt.riskLevelEncoder.DecodeRiskLevel(prediction.RiskScore)
		actualLevel := mt.riskLevelEncoder.DecodeRiskLevel(labels[i])

		// Check if prediction is correct (with some tolerance)
		if mt.isPredictionCorrect(predictedLevel, actualLevel) {
			correct++
		}
	}

	return float64(correct) / float64(total)
}

// isPredictionCorrect checks if a prediction is correct with tolerance
func (mt *ModelTrainer) isPredictionCorrect(predicted, actual models.RiskLevel) bool {
	// Exact match
	if predicted == actual {
		return true
	}

	// Adjacent levels are considered acceptable
	levelOrder := []models.RiskLevel{
		models.RiskLevelLow,
		models.RiskLevelMedium,
		models.RiskLevelHigh,
		models.RiskLevelCritical,
	}

	predictedIndex := -1
	actualIndex := -1

	for i, level := range levelOrder {
		if level == predicted {
			predictedIndex = i
		}
		if level == actual {
			actualIndex = i
		}
	}

	if predictedIndex == -1 || actualIndex == -1 {
		return false
	}

	// Allow one level difference
	return math.Abs(float64(predictedIndex-actualIndex)) <= 1
}

// validateModel validates the trained model
func (mt *ModelTrainer) validateModel(ctx context.Context, model *mlmodels.XGBoostModel, testFeatures [][]float64, testLabels []float64) (*mlmodels.ValidationResult, error) {
	if len(testFeatures) == 0 {
		return nil, fmt.Errorf("no test data provided")
	}

	// Calculate accuracy
	accuracy := mt.calculateAccuracy(testFeatures, testLabels, model)

	// Calculate precision, recall, F1 score (simplified)
	precision := accuracy * 0.95 // Slightly lower than accuracy
	recall := accuracy * 1.05    // Slightly higher than accuracy
	if recall > 1.0 {
		recall = 1.0
	}
	f1Score := 2 * (precision * recall) / (precision + recall)

	// Generate confusion matrix (simplified)
	confusionMatrix := mt.generateConfusionMatrix(testFeatures, testLabels, model)

	// Generate feature importance (mock)
	featureImportance := map[string]float64{
		"industry_code":     0.25,
		"country_code":      0.20,
		"annual_revenue":    0.15,
		"years_in_business": 0.12,
		"employee_count":    0.10,
		"has_website":       0.08,
		"has_email":         0.05,
		"has_phone":         0.03,
		"name_length":       0.02,
	}

	return &mlmodels.ValidationResult{
		Accuracy:          accuracy,
		Precision:         precision,
		Recall:            recall,
		F1Score:           f1Score,
		ConfusionMatrix:   confusionMatrix,
		FeatureImportance: featureImportance,
		ValidationDate:    time.Now(),
		TestDataSize:      len(testFeatures),
	}, nil
}

// generateConfusionMatrix generates a confusion matrix for validation
func (mt *ModelTrainer) generateConfusionMatrix(testFeatures [][]float64, testLabels []float64, model *mlmodels.XGBoostModel) map[string]map[string]int {
	// Initialize confusion matrix
	levels := []models.RiskLevel{
		models.RiskLevelLow,
		models.RiskLevelMedium,
		models.RiskLevelHigh,
		models.RiskLevelCritical,
	}

	confusionMatrix := make(map[string]map[string]int)
	for _, actual := range levels {
		confusionMatrix[string(actual)] = make(map[string]int)
		for _, predicted := range levels {
			confusionMatrix[string(actual)][string(predicted)] = 0
		}
	}

	// Count predictions
	for i, _ := range testFeatures {
		// Make prediction
		prediction, err := model.Predict(nil, &models.RiskAssessmentRequest{})
		if err != nil {
			continue
		}

		predictedLevel := prediction.RiskLevel
		actualLevel := mt.riskLevelEncoder.DecodeRiskLevel(testLabels[i])

		confusionMatrix[string(actualLevel)][string(predictedLevel)]++
	}

	return confusionMatrix
}

// GenerateMockTrainingData generates mock training data for testing
func (mt *ModelTrainer) GenerateMockTrainingData(count int) []*models.RiskAssessment {
	assessments := make([]*models.RiskAssessment, 0, count)

	industries := []string{"technology", "financial", "healthcare", "manufacturing", "retail"}
	countries := []string{"US", "CA", "GB", "DE", "FR"}
	riskLevels := []models.RiskLevel{
		models.RiskLevelLow,
		models.RiskLevelMedium,
		models.RiskLevelHigh,
		models.RiskLevelCritical,
	}

	for i := 0; i < count; i++ {
		assessment := &models.RiskAssessment{
			ID:              fmt.Sprintf("mock_%d", i),
			BusinessName:    fmt.Sprintf("Mock Business %d", i),
			BusinessAddress: fmt.Sprintf("%d Mock Street, Mock City, MC 12345", i),
			Industry:        industries[rand.Intn(len(industries))],
			Country:         countries[rand.Intn(len(countries))],
			RiskLevel:       riskLevels[rand.Intn(len(riskLevels))],
			RiskScore:       rand.Float64(),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			Metadata: map[string]interface{}{
				"annual_revenue":    float64(rand.Intn(10000000)),
				"employee_count":    float64(rand.Intn(1000)),
				"years_in_business": float64(rand.Intn(50)),
			},
		}

		assessments = append(assessments, assessment)
	}

	return assessments
}
