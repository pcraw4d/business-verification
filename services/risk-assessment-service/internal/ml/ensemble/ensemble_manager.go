package ensemble

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// EnsembleManager manages multiple models and combines their predictions
type EnsembleManager struct {
	logger *zap.Logger

	// Models in the ensemble
	models map[string]EnsembleModel

	// Ensemble configuration
	config EnsembleConfig

	// Model weights (learned from validation data)
	weights map[string]float64

	// Performance tracking
	performance map[string]ModelPerformance

	// Thread safety
	mu sync.RWMutex
}

// EnsembleModel represents a model that can be part of an ensemble
type EnsembleModel interface {
	Predict(ctx context.Context, request *models.RiskAssessmentRequest) (*models.RiskAssessment, error)
	GetModelInfo() map[string]interface{}
	ValidateModel(ctx context.Context, data []models.RiskAssessmentRequest) (map[string]interface{}, error)
}

// EnsembleConfig holds configuration for the ensemble
type EnsembleConfig struct {
	Method              string             `json:"method"`               // "weighted_average", "stacking", "voting", "bayesian"
	WeightOptimization  bool               `json:"weight_optimization"`  // Whether to optimize weights
	ValidationSplit     float64            `json:"validation_split"`     // Fraction of data for validation
	MinModels           int                `json:"min_models"`           // Minimum models required
	MaxModels           int                `json:"max_models"`           // Maximum models allowed
	ConfidenceThreshold float64            `json:"confidence_threshold"` // Minimum confidence for prediction
	FallbackModel       string             `json:"fallback_model"`       // Fallback model if ensemble fails
	CustomWeights       map[string]float64 `json:"custom_weights"`       // Custom model weights
}

// ModelPerformance tracks performance metrics for each model
type ModelPerformance struct {
	Accuracy    float64       `json:"accuracy"`
	Precision   float64       `json:"precision"`
	Recall      float64       `json:"recall"`
	F1Score     float64       `json:"f1_score"`
	Confidence  float64       `json:"confidence"`
	Latency     time.Duration `json:"latency"`
	LastUpdated time.Time     `json:"last_updated"`
}

// EnsembleResult contains the result of ensemble prediction
type EnsembleResult struct {
	Prediction            *models.RiskAssessment            `json:"prediction"`
	ModelContributions    map[string]float64                `json:"model_contributions"`
	EnsembleConfidence    float64                           `json:"ensemble_confidence"`
	ModelAgreement        float64                           `json:"model_agreement"`
	PredictionMethod      string                            `json:"prediction_method"`
	FallbackUsed          bool                              `json:"fallback_used"`
	ProcessingTime        time.Duration                     `json:"processing_time"`
	IndividualPredictions map[string]*models.RiskAssessment `json:"individual_predictions"`
}

// WeightOptimizationResult contains the result of weight optimization
type WeightOptimizationResult struct {
	OptimizedWeights      map[string]float64 `json:"optimized_weights"`
	ValidationScore       float64            `json:"validation_score"`
	ImprovementScore      float64            `json:"improvement_score"`
	OptimizationMethod    string             `json:"optimization_method"`
	OptimizationTime      time.Duration      `json:"optimization_time"`
	ConvergenceIterations int                `json:"convergence_iterations"`
}

// NewEnsembleManager creates a new ensemble manager
func NewEnsembleManager(logger *zap.Logger, config EnsembleConfig) *EnsembleManager {
	return &EnsembleManager{
		logger:      logger,
		models:      make(map[string]EnsembleModel),
		config:      config,
		weights:     make(map[string]float64),
		performance: make(map[string]ModelPerformance),
	}
}

// AddModel adds a model to the ensemble
func (em *EnsembleManager) AddModel(name string, model EnsembleModel) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if len(em.models) >= em.config.MaxModels {
		return fmt.Errorf("maximum number of models (%d) reached", em.config.MaxModels)
	}

	em.models[name] = model
	em.weights[name] = 1.0 / float64(len(em.models)) // Equal weights initially

	em.logger.Info("Model added to ensemble",
		zap.String("model_name", name),
		zap.Int("total_models", len(em.models)))

	return nil
}

// RemoveModel removes a model from the ensemble
func (em *EnsembleManager) RemoveModel(name string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if len(em.models) <= em.config.MinModels {
		return fmt.Errorf("minimum number of models (%d) required", em.config.MinModels)
	}

	delete(em.models, name)
	delete(em.weights, name)
	delete(em.performance, name)

	// Rebalance weights
	em.rebalanceWeights()

	em.logger.Info("Model removed from ensemble",
		zap.String("model_name", name),
		zap.Int("total_models", len(em.models)))

	return nil
}

// Predict performs ensemble prediction
func (em *EnsembleManager) Predict(ctx context.Context, request *models.RiskAssessmentRequest) (*EnsembleResult, error) {
	startTime := time.Now()

	em.mu.RLock()
	defer em.mu.RUnlock()

	if len(em.models) < em.config.MinModels {
		return nil, fmt.Errorf("insufficient models in ensemble (have %d, need %d)", len(em.models), em.config.MinModels)
	}

	// Get predictions from all models
	predictions := make(map[string]*models.RiskAssessment)
	var errors []error

	for name, model := range em.models {
		pred, err := model.Predict(ctx, request)
		if err != nil {
			em.logger.Warn("Model prediction failed",
				zap.String("model_name", name),
				zap.Error(err))
			errors = append(errors, fmt.Errorf("model %s: %w", name, err))
			continue
		}
		predictions[name] = pred
	}

	// Check if we have enough successful predictions
	if len(predictions) < em.config.MinModels {
		// Try fallback model if specified
		if em.config.FallbackModel != "" {
			if fallbackModel, exists := em.models[em.config.FallbackModel]; exists {
				em.logger.Info("Using fallback model",
					zap.String("fallback_model", em.config.FallbackModel))

				pred, err := fallbackModel.Predict(ctx, request)
				if err != nil {
					return nil, fmt.Errorf("fallback model also failed: %w", err)
				}

				return &EnsembleResult{
					Prediction:            pred,
					ModelContributions:    map[string]float64{em.config.FallbackModel: 1.0},
					EnsembleConfidence:    pred.ConfidenceScore,
					ModelAgreement:        1.0,
					PredictionMethod:      "fallback",
					FallbackUsed:          true,
					ProcessingTime:        time.Since(startTime),
					IndividualPredictions: map[string]*models.RiskAssessment{em.config.FallbackModel: pred},
				}, nil
			}
		}

		return nil, fmt.Errorf("insufficient successful predictions: %v", errors)
	}

	// Combine predictions based on method
	var ensemblePrediction *models.RiskAssessment
	var modelContributions map[string]float64
	var ensembleConfidence float64

	switch em.config.Method {
	case "weighted_average":
		ensemblePrediction, modelContributions, ensembleConfidence = em.weightedAverage(predictions)
	case "stacking":
		ensemblePrediction, modelContributions, ensembleConfidence = em.stacking(predictions)
	case "voting":
		ensemblePrediction, modelContributions, ensembleConfidence = em.voting(predictions)
	case "bayesian":
		ensemblePrediction, modelContributions, ensembleConfidence = em.bayesianCombination(predictions)
	default:
		ensemblePrediction, modelContributions, ensembleConfidence = em.weightedAverage(predictions)
	}

	// Calculate model agreement
	modelAgreement := em.calculateModelAgreement(predictions)

	// Check confidence threshold
	if ensembleConfidence < em.config.ConfidenceThreshold {
		em.logger.Warn("Ensemble confidence below threshold",
			zap.Float64("confidence", ensembleConfidence),
			zap.Float64("threshold", em.config.ConfidenceThreshold))
	}

	result := &EnsembleResult{
		Prediction:            ensemblePrediction,
		ModelContributions:    modelContributions,
		EnsembleConfidence:    ensembleConfidence,
		ModelAgreement:        modelAgreement,
		PredictionMethod:      em.config.Method,
		FallbackUsed:          false,
		ProcessingTime:        time.Since(startTime),
		IndividualPredictions: predictions,
	}

	em.logger.Debug("Ensemble prediction completed",
		zap.Duration("processing_time", result.ProcessingTime),
		zap.Float64("ensemble_confidence", ensembleConfidence),
		zap.Float64("model_agreement", modelAgreement))

	return result, nil
}

// OptimizeWeights optimizes model weights based on validation data
func (em *EnsembleManager) OptimizeWeights(ctx context.Context, validationData []models.RiskAssessmentRequest, actuals []float64) (*WeightOptimizationResult, error) {
	startTime := time.Now()

	if len(validationData) != len(actuals) {
		return nil, fmt.Errorf("validation data and actuals must have the same length")
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	em.logger.Info("Starting weight optimization",
		zap.Int("validation_samples", len(validationData)),
		zap.Int("num_models", len(em.models)))

	// Get predictions from all models for validation data
	modelPredictions := make(map[string][]float64)

	for name, model := range em.models {
		predictions := make([]float64, len(validationData))
		for i, request := range validationData {
			pred, err := model.Predict(ctx, &request)
			if err != nil {
				em.logger.Warn("Model prediction failed during optimization",
					zap.String("model_name", name),
					zap.Int("sample_index", i),
					zap.Error(err))
				// Use previous prediction or default
				predictions[i] = 0.5
			} else {
				predictions[i] = pred.RiskScore
			}
		}
		modelPredictions[name] = predictions
	}

	// Optimize weights using different methods
	var bestWeights map[string]float64
	var bestScore float64
	var bestMethod string

	// Method 1: Performance-based weights
	perfWeights := em.optimizePerformanceBasedWeights(modelPredictions, actuals)
	perfScore := em.evaluateWeights(perfWeights, modelPredictions, actuals)
	if perfScore > bestScore {
		bestWeights = perfWeights
		bestScore = perfScore
		bestMethod = "performance_based"
	}

	// Method 2: Correlation-based weights
	corrWeights := em.optimizeCorrelationBasedWeights(modelPredictions, actuals)
	corrScore := em.evaluateWeights(corrWeights, modelPredictions, actuals)
	if corrScore > bestScore {
		bestWeights = corrWeights
		bestScore = corrScore
		bestMethod = "correlation_based"
	}

	// Method 3: Stacking weights (simplified)
	stackWeights := em.optimizeStackingWeights(modelPredictions, actuals)
	stackScore := em.evaluateWeights(stackWeights, modelPredictions, actuals)
	if stackScore > bestScore {
		bestWeights = stackWeights
		bestScore = stackScore
		bestMethod = "stacking"
	}

	// Update ensemble weights
	em.weights = bestWeights

	// Calculate improvement score
	baselineScore := em.evaluateWeights(em.getEqualWeights(), modelPredictions, actuals)
	improvementScore := bestScore - baselineScore

	result := &WeightOptimizationResult{
		OptimizedWeights:      bestWeights,
		ValidationScore:       bestScore,
		ImprovementScore:      improvementScore,
		OptimizationMethod:    bestMethod,
		OptimizationTime:      time.Since(startTime),
		ConvergenceIterations: 3, // Simplified
	}

	em.logger.Info("Weight optimization completed",
		zap.Duration("optimization_time", result.OptimizationTime),
		zap.Float64("validation_score", bestScore),
		zap.Float64("improvement_score", improvementScore),
		zap.String("best_method", bestMethod))

	return result, nil
}

// weightedAverage combines predictions using weighted average
func (em *EnsembleManager) weightedAverage(predictions map[string]*models.RiskAssessment) (ensemblePrediction *models.RiskAssessment, modelContributions map[string]float64, confidence float64) {
	modelContributions = make(map[string]float64)

	var weightedScore, weightedConfidence, totalWeight float64

	for name, prediction := range predictions {
		weight := em.weights[name]
		modelContributions[name] = weight

		weightedScore += prediction.RiskScore * weight
		weightedConfidence += prediction.ConfidenceScore * weight
		totalWeight += weight
	}

	// Normalize by total weight
	weightedScore /= totalWeight
	weightedConfidence /= totalWeight

	// Create ensemble prediction
	ensemblePrediction = &models.RiskAssessment{
		RiskScore:       weightedScore,
		RiskLevel:       em.determineRiskLevel(weightedScore),
		ConfidenceScore: weightedConfidence,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return ensemblePrediction, modelContributions, weightedConfidence
}

// stacking combines predictions using stacking (simplified implementation)
func (em *EnsembleManager) stacking(predictions map[string]*models.RiskAssessment) (ensemblePrediction *models.RiskAssessment, modelContributions map[string]float64, confidence float64) {
	// Simplified stacking implementation
	// In practice, you would train a meta-learner on the predictions

	modelContributions = make(map[string]float64)

	// Use performance-based weights for stacking
	var weightedScore, weightedConfidence, totalWeight float64

	for name, prediction := range predictions {
		// Use performance metrics to determine weight
		perf, exists := em.performance[name]
		var weight float64
		if exists {
			weight = perf.F1Score * perf.Confidence
		} else {
			weight = em.weights[name]
		}

		modelContributions[name] = weight
		weightedScore += prediction.RiskScore * weight
		weightedConfidence += prediction.ConfidenceScore * weight
		totalWeight += weight
	}

	// Normalize
	weightedScore /= totalWeight
	weightedConfidence /= totalWeight

	ensemblePrediction = &models.RiskAssessment{
		RiskScore:       weightedScore,
		RiskLevel:       em.determineRiskLevel(weightedScore),
		ConfidenceScore: weightedConfidence,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return ensemblePrediction, modelContributions, weightedConfidence
}

// voting combines predictions using voting
func (em *EnsembleManager) voting(predictions map[string]*models.RiskAssessment) (ensemblePrediction *models.RiskAssessment, modelContributions map[string]float64, confidence float64) {
	modelContributions = make(map[string]float64)

	// Count votes for each risk level
	votes := make(map[models.RiskLevel]int)
	confidences := make(map[models.RiskLevel][]float64)

	for name, prediction := range predictions {
		weight := em.weights[name]
		modelContributions[name] = weight

		// Weighted voting
		voteCount := int(weight * 100) // Convert to integer votes
		votes[prediction.RiskLevel] += voteCount
		confidences[prediction.RiskLevel] = append(confidences[prediction.RiskLevel], prediction.ConfidenceScore)
	}

	// Find the risk level with most votes
	var winningLevel models.RiskLevel
	maxVotes := 0
	for level, count := range votes {
		if count > maxVotes {
			maxVotes = count
			winningLevel = level
		}
	}

	// Calculate average confidence for winning level
	var avgConfidence float64
	if confs, exists := confidences[winningLevel]; exists && len(confs) > 0 {
		var sum float64
		for _, conf := range confs {
			sum += conf
		}
		avgConfidence = sum / float64(len(confs))
	}

	// Convert risk level back to score
	score := em.riskLevelToScore(winningLevel)

	ensemblePrediction = &models.RiskAssessment{
		RiskScore:       score,
		RiskLevel:       winningLevel,
		ConfidenceScore: avgConfidence,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return ensemblePrediction, modelContributions, avgConfidence
}

// bayesianCombination combines predictions using Bayesian model averaging
func (em *EnsembleManager) bayesianCombination(predictions map[string]*models.RiskAssessment) (ensemblePrediction *models.RiskAssessment, modelContributions map[string]float64, confidence float64) {
	modelContributions = make(map[string]float64)

	// Bayesian model averaging
	var totalWeight float64
	weightedScore := 0.0

	for name, prediction := range predictions {
		// Use model performance as prior
		perf, exists := em.performance[name]
		var priorWeight float64
		if exists {
			priorWeight = perf.F1Score * perf.Confidence
		} else {
			priorWeight = em.weights[name]
		}

		// Calculate likelihood based on confidence
		likelihood := prediction.ConfidenceScore
		posteriorWeight := priorWeight * likelihood

		modelContributions[name] = posteriorWeight
		weightedScore += prediction.RiskScore * posteriorWeight
		totalWeight += posteriorWeight
	}

	// Normalize
	weightedScore /= totalWeight

	// Calculate ensemble confidence
	var weightedConfidence float64
	for name, prediction := range predictions {
		weight := modelContributions[name] / totalWeight
		weightedConfidence += prediction.ConfidenceScore * weight
	}

	ensemblePrediction = &models.RiskAssessment{
		RiskScore:       weightedScore,
		RiskLevel:       em.determineRiskLevel(weightedScore),
		ConfidenceScore: weightedConfidence,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return ensemblePrediction, modelContributions, weightedConfidence
}

// Helper methods

func (em *EnsembleManager) predictWithFallback(ctx context.Context, request *models.RiskAssessmentRequest) (*EnsembleResult, error) {
	// Implementation for fallback prediction
	return nil, fmt.Errorf("fallback prediction not implemented")
}

func (em *EnsembleManager) calculateModelAgreement(predictions map[string]*models.RiskAssessment) float64 {
	if len(predictions) < 2 {
		return 1.0
	}

	var totalAgreement float64
	var comparisons int

	predList := make([]*models.RiskAssessment, 0, len(predictions))
	for _, pred := range predictions {
		predList = append(predList, pred)
	}

	for i := 0; i < len(predList); i++ {
		for j := i + 1; j < len(predList); j++ {
			agreement := 1.0 - math.Abs(predList[i].RiskScore-predList[j].RiskScore)
			totalAgreement += agreement
			comparisons++
		}
	}

	if comparisons == 0 {
		return 1.0
	}

	return totalAgreement / float64(comparisons)
}

func (em *EnsembleManager) determineRiskLevel(score float64) models.RiskLevel {
	switch {
	case score < 0.2:
		return models.RiskLevelLow
	case score < 0.4:
		return models.RiskLevelMedium
	case score < 0.7:
		return models.RiskLevelHigh
	default:
		return models.RiskLevelCritical
	}
}

func (em *EnsembleManager) riskLevelToScore(level models.RiskLevel) float64 {
	switch level {
	case models.RiskLevelLow:
		return 0.1
	case models.RiskLevelMedium:
		return 0.3
	case models.RiskLevelHigh:
		return 0.6
	case models.RiskLevelCritical:
		return 0.9
	default:
		return 0.5
	}
}

func (em *EnsembleManager) rebalanceWeights() {
	if len(em.weights) == 0 {
		return
	}

	equalWeight := 1.0 / float64(len(em.weights))
	for name := range em.weights {
		em.weights[name] = equalWeight
	}
}

func (em *EnsembleManager) getEqualWeights() map[string]float64 {
	weights := make(map[string]float64)
	if len(em.models) == 0 {
		return weights
	}

	equalWeight := 1.0 / float64(len(em.models))
	for name := range em.models {
		weights[name] = equalWeight
	}
	return weights
}

// Weight optimization methods

func (em *EnsembleManager) optimizePerformanceBasedWeights(modelPredictions map[string][]float64, actuals []float64) map[string]float64 {
	weights := make(map[string]float64)
	var totalPerformance float64

	// Calculate performance-based weights
	for name, predictions := range modelPredictions {
		// Calculate accuracy as performance metric
		accuracy := em.calculateAccuracy(predictions, actuals)
		weights[name] = accuracy
		totalPerformance += accuracy
	}

	// Normalize weights
	if totalPerformance > 0 {
		for name := range weights {
			weights[name] /= totalPerformance
		}
	} else {
		// Fallback to equal weights
		equalWeight := 1.0 / float64(len(weights))
		for name := range weights {
			weights[name] = equalWeight
		}
	}

	return weights
}

func (em *EnsembleManager) optimizeCorrelationBasedWeights(modelPredictions map[string][]float64, actuals []float64) map[string]float64 {
	weights := make(map[string]float64)
	var totalCorrelation float64

	// Calculate correlation-based weights
	for name, predictions := range modelPredictions {
		correlation := em.calculateCorrelation(predictions, actuals)
		// Use absolute correlation and ensure positive weights
		weight := math.Abs(correlation)
		weights[name] = weight
		totalCorrelation += weight
	}

	// Normalize weights
	if totalCorrelation > 0 {
		for name := range weights {
			weights[name] /= totalCorrelation
		}
	} else {
		// Fallback to equal weights
		equalWeight := 1.0 / float64(len(weights))
		for name := range weights {
			weights[name] = equalWeight
		}
	}

	return weights
}

func (em *EnsembleManager) optimizeStackingWeights(modelPredictions map[string][]float64, actuals []float64) map[string]float64 {
	// Simplified stacking optimization
	// In practice, this would train a meta-learner

	// Use a simple linear combination optimization
	// This is a placeholder for more sophisticated stacking

	// For now, use performance-based weights as stacking approximation
	return em.optimizePerformanceBasedWeights(modelPredictions, actuals)
}

func (em *EnsembleManager) evaluateWeights(weights map[string]float64, modelPredictions map[string][]float64, actuals []float64) float64 {
	if len(modelPredictions) == 0 || len(actuals) == 0 {
		return 0.0
	}

	// Calculate ensemble predictions using given weights
	ensemblePredictions := make([]float64, len(actuals))

	for i := range actuals {
		var weightedSum, totalWeight float64

		for name, predictions := range modelPredictions {
			if i < len(predictions) {
				weight := weights[name]
				weightedSum += predictions[i] * weight
				totalWeight += weight
			}
		}

		if totalWeight > 0 {
			ensemblePredictions[i] = weightedSum / totalWeight
		} else {
			ensemblePredictions[i] = 0.5 // Default
		}
	}

	// Calculate accuracy as evaluation metric
	return em.calculateAccuracy(ensemblePredictions, actuals)
}

func (em *EnsembleManager) calculateAccuracy(predictions, actuals []float64) float64 {
	if len(predictions) != len(actuals) || len(predictions) == 0 {
		return 0.0
	}

	var correct float64
	for i := range predictions {
		// Convert to binary classification for accuracy calculation
		predBinary := 0.0
		if predictions[i] > 0.5 {
			predBinary = 1.0
		}

		actualBinary := 0.0
		if actuals[i] > 0.5 {
			actualBinary = 1.0
		}

		if predBinary == actualBinary {
			correct++
		}
	}

	return correct / float64(len(predictions))
}

func (em *EnsembleManager) calculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0.0
	}

	// Calculate means
	var sumX, sumY float64
	for i := range x {
		sumX += x[i]
		sumY += y[i]
	}
	meanX := sumX / float64(len(x))
	meanY := sumY / float64(len(y))

	// Calculate correlation coefficient
	var numerator, sumXSquared, sumYSquared float64
	for i := range x {
		dx := x[i] - meanX
		dy := y[i] - meanY
		numerator += dx * dy
		sumXSquared += dx * dx
		sumYSquared += dy * dy
	}

	if sumXSquared == 0 || sumYSquared == 0 {
		return 0.0
	}

	return numerator / math.Sqrt(sumXSquared*sumYSquared)
}

// GetEnsembleInfo returns information about the ensemble
func (em *EnsembleManager) GetEnsembleInfo() map[string]interface{} {
	em.mu.RLock()
	defer em.mu.RUnlock()

	modelInfo := make(map[string]interface{})
	for name, model := range em.models {
		modelInfo[name] = model.GetModelInfo()
	}

	return map[string]interface{}{
		"num_models":  len(em.models),
		"method":      em.config.Method,
		"weights":     em.weights,
		"performance": em.performance,
		"model_info":  modelInfo,
		"config":      em.config,
	}
}
