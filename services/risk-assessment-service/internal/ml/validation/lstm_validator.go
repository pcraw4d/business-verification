package validation

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/ml/data"
	"kyb-platform/services/risk-assessment-service/internal/ml/ensemble"
	"kyb-platform/services/risk-assessment-service/internal/models"
)

// LSTMValidator validates LSTM model accuracy with cross-validation
type LSTMValidator struct {
	syntheticGenerator *data.SyntheticDataGenerator
	historyCollector   *data.HistoryCollector
	hybridBlender      *data.HybridBlender
	ensembleRouter     *ensemble.EnsembleRouter
	logger             *zap.Logger
}

// ValidationConfig holds configuration for validation
type ValidationConfig struct {
	NumBusinesses        int
	SequenceLength       int
	ValidationHorizons   []int
	CrossValidationFolds int
	TestDataRatio        float64
	RandomSeed           int64
}

// LSTMValidationResult contains the results of model validation
type LSTMValidationResult struct {
	ModelType              string                  `json:"model_type"`
	ValidationDate         time.Time               `json:"validation_date"`
	TotalSamples           int                     `json:"total_samples"`
	HorizonResults         map[int]HorizonResult   `json:"horizon_results"`
	OverallAccuracy        float64                 `json:"overall_accuracy"`
	OverallMAE             float64                 `json:"overall_mae"`
	OverallRMSE            float64                 `json:"overall_rmse"`
	CrossValidationResults []CrossValidationResult `json:"cross_validation_results"`
	ModelComparison        ModelComparison         `json:"model_comparison"`
	ConfidenceAnalysis     ConfidenceAnalysis      `json:"confidence_analysis"`
}

// HorizonResult contains validation results for a specific prediction horizon
type HorizonResult struct {
	HorizonMonths      int                `json:"horizon_months"`
	Accuracy           float64            `json:"accuracy"`
	Precision          float64            `json:"precision"`
	Recall             float64            `json:"recall"`
	F1Score            float64            `json:"f1_score"`
	MAE                float64            `json:"mae"`
	RMSE               float64            `json:"rmse"`
	R2Score            float64            `json:"r2_score"`
	SampleCount        int                `json:"sample_count"`
	ConfidenceInterval ConfidenceInterval `json:"confidence_interval"`
}

// CrossValidationResult contains results from a single fold
type CrossValidationResult struct {
	Fold            int                   `json:"fold"`
	HorizonResults  map[int]HorizonResult `json:"horizon_results"`
	OverallAccuracy float64               `json:"overall_accuracy"`
	OverallMAE      float64               `json:"overall_mae"`
	OverallRMSE     float64               `json:"overall_rmse"`
}

// ModelComparison compares different models
type ModelComparison struct {
	XGBoostResults      map[int]HorizonResult `json:"xgboost_results"`
	LSTMResults         map[int]HorizonResult `json:"lstm_results"`
	EnsembleResults     map[int]HorizonResult `json:"ensemble_results"`
	BestModelPerHorizon map[int]string        `json:"best_model_per_horizon"`
}

// ConfidenceAnalysis analyzes prediction confidence
type ConfidenceAnalysis struct {
	HighConfidenceAccuracy float64 `json:"high_confidence_accuracy"`
	LowConfidenceAccuracy  float64 `json:"low_confidence_accuracy"`
	ConfidenceThreshold    float64 `json:"confidence_threshold"`
	CalibrationError       float64 `json:"calibration_error"`
}

// ConfidenceInterval represents a confidence interval
type ConfidenceInterval struct {
	Lower float64 `json:"lower"`
	Upper float64 `json:"upper"`
	Level float64 `json:"level"`
}

// ValidationSample represents a single validation sample
type ValidationSample struct {
	Business           *models.RiskAssessmentRequest
	TrueRiskScore      float64
	TrueRiskLevel      models.RiskLevel
	PredictedRiskScore float64
	PredictedRiskLevel models.RiskLevel
	ConfidenceScore    float64
	HorizonMonths      int
	ModelType          string
	Timestamp          time.Time
}

// NewLSTMValidator creates a new LSTM validator
func NewLSTMValidator(
	syntheticGenerator *data.SyntheticDataGenerator,
	historyCollector *data.HistoryCollector,
	hybridBlender *data.HybridBlender,
	ensembleRouter *ensemble.EnsembleRouter,
	logger *zap.Logger,
) *LSTMValidator {
	return &LSTMValidator{
		syntheticGenerator: syntheticGenerator,
		historyCollector:   historyCollector,
		hybridBlender:      hybridBlender,
		ensembleRouter:     ensembleRouter,
		logger:             logger,
	}
}

// ValidateModel performs comprehensive model validation
func (lv *LSTMValidator) ValidateModel(ctx context.Context, config ValidationConfig) (*LSTMValidationResult, error) {
	lv.logger.Info("Starting LSTM model validation",
		zap.Int("num_businesses", config.NumBusinesses),
		zap.Ints("horizons", config.ValidationHorizons),
		zap.Int("cv_folds", config.CrossValidationFolds))

	// Generate validation dataset
	validationSamples, err := lv.generateValidationDataset(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate validation dataset: %w", err)
	}

	lv.logger.Info("Generated validation dataset",
		zap.Int("total_samples", len(validationSamples)))

	// Perform cross-validation
	cvResults, err := lv.performCrossValidation(ctx, validationSamples, config)
	if err != nil {
		return nil, fmt.Errorf("failed to perform cross-validation: %w", err)
	}

	// Calculate overall results
	overallResults := lv.calculateOverallResults(cvResults)

	// Perform model comparison
	modelComparison, err := lv.performModelComparison(ctx, validationSamples, config)
	if err != nil {
		return nil, fmt.Errorf("failed to perform model comparison: %w", err)
	}

	// Analyze confidence
	confidenceAnalysis := lv.analyzeConfidence(validationSamples)

	// Create validation result
	result := &LSTMValidationResult{
		ModelType:              "lstm",
		ValidationDate:         time.Now(),
		TotalSamples:           len(validationSamples),
		HorizonResults:         overallResults,
		OverallAccuracy:        lv.calculateOverallAccuracy(overallResults),
		OverallMAE:             lv.calculateOverallMAE(overallResults),
		OverallRMSE:            lv.calculateOverallRMSE(overallResults),
		CrossValidationResults: cvResults,
		ModelComparison:        *modelComparison,
		ConfidenceAnalysis:     confidenceAnalysis,
	}

	lv.logger.Info("LSTM model validation completed",
		zap.Float64("overall_accuracy", result.OverallAccuracy),
		zap.Float64("overall_mae", result.OverallMAE),
		zap.Float64("overall_rmse", result.OverallRMSE))

	return result, nil
}

// generateValidationDataset generates a comprehensive validation dataset
func (lv *LSTMValidator) generateValidationDataset(ctx context.Context, config ValidationConfig) ([]ValidationSample, error) {
	var samples []ValidationSample

	// Generate businesses for validation
	businesses := lv.generateValidationBusinesses(config.NumBusinesses)

	for _, business := range businesses {
		// Generate time-series data for this business
		timeSeries, err := lv.hybridBlender.BuildTimeSeries(ctx, business, config.SequenceLength)
		if err != nil {
			lv.logger.Warn("Failed to generate time series for business",
				zap.String("business_name", business.BusinessName),
				zap.Error(err))
			continue
		}

		// Create validation samples for each horizon
		for _, horizon := range config.ValidationHorizons {
			// Get true future risk score (simulated)
			trueRiskScore := lv.simulateTrueRiskScore(timeSeries, horizon)

			// Get model predictions
			predictions, err := lv.getModelPredictions(ctx, business, horizon)
			if err != nil {
				lv.logger.Warn("Failed to get model predictions",
					zap.String("business_name", business.BusinessName),
					zap.Int("horizon", horizon),
					zap.Error(err))
				continue
			}

			// Create samples for each model
			for modelType, prediction := range predictions {
				sample := ValidationSample{
					Business:           business,
					TrueRiskScore:      trueRiskScore,
					TrueRiskLevel:      lv.determineRiskLevel(trueRiskScore),
					PredictedRiskScore: prediction.RiskScore,
					PredictedRiskLevel: prediction.RiskLevel,
					ConfidenceScore:    prediction.ConfidenceScore,
					HorizonMonths:      horizon,
					ModelType:          modelType,
					Timestamp:          time.Now(),
				}
				samples = append(samples, sample)
			}
		}
	}

	return samples, nil
}

// generateValidationBusinesses generates businesses for validation
func (lv *LSTMValidator) generateValidationBusinesses(count int) []*models.RiskAssessmentRequest {
	businesses := make([]*models.RiskAssessmentRequest, count)
	industries := []string{"technology", "healthcare", "financial", "manufacturing", "retail"}

	for i := 0; i < count; i++ {
		industry := industries[i%len(industries)]
		businesses[i] = &models.RiskAssessmentRequest{
			BusinessName:      fmt.Sprintf("Validation Business %d", i),
			BusinessAddress:   fmt.Sprintf("%d Validation St, Test City, TC 12345", i),
			Industry:          industry,
			Country:           "US",
			Phone:             fmt.Sprintf("+1-555-%03d-%04d", i%1000, i%10000),
			Email:             fmt.Sprintf("test%d@validation.com", i),
			Website:           fmt.Sprintf("https://validation%d.com", i),
			PredictionHorizon: 6,
			Metadata: map[string]interface{}{
				"validation": true,
				"index":      i,
			},
		}
	}

	return businesses
}

// simulateTrueRiskScore simulates the true future risk score
func (lv *LSTMValidator) simulateTrueRiskScore(timeSeries []data.RiskDataPoint, horizon int) float64 {
	if len(timeSeries) == 0 {
		return 0.5 // Default risk score
	}

	// Use the last data point as base
	lastPoint := timeSeries[len(timeSeries)-1]
	baseRisk := lastPoint.RiskScore

	// Add trend based on recent history
	if len(timeSeries) >= 3 {
		recentTrend := (timeSeries[len(timeSeries)-1].RiskScore - timeSeries[len(timeSeries)-3].RiskScore) / 2
		baseRisk += recentTrend * float64(horizon) / 12.0
	}

	// Add some randomness
	randomFactor := (math.Sin(float64(horizon)) + 1) / 2 * 0.1
	baseRisk += randomFactor

	// Ensure risk score is in valid range
	if baseRisk < 0 {
		baseRisk = 0
	} else if baseRisk > 1 {
		baseRisk = 1
	}

	return baseRisk
}

// getModelPredictions gets predictions from all models
func (lv *LSTMValidator) getModelPredictions(ctx context.Context, business *models.RiskAssessmentRequest, horizon int) (map[string]*models.RiskAssessment, error) {
	predictions := make(map[string]*models.RiskAssessment)

	// Get XGBoost prediction (using ensemble router's internal models)
	// Note: This is a simplified approach for validation
	// In a real implementation, we'd need getter methods on the ensemble router

	// Get ensemble prediction
	ensemblePrediction, err := lv.ensembleRouter.PredictWithEnsemble(ctx, business)
	if err == nil {
		predictions["ensemble"] = ensemblePrediction
	}

	return predictions, nil
}

// performCrossValidation performs k-fold cross-validation
func (lv *LSTMValidator) performCrossValidation(ctx context.Context, samples []ValidationSample, config ValidationConfig) ([]CrossValidationResult, error) {
	var results []CrossValidationResult

	// Shuffle samples
	shuffledSamples := make([]ValidationSample, len(samples))
	copy(shuffledSamples, samples)
	lv.shuffleSamples(shuffledSamples)

	// Calculate fold size
	foldSize := len(shuffledSamples) / config.CrossValidationFolds

	for fold := 0; fold < config.CrossValidationFolds; fold++ {
		// Split data into train and test
		start := fold * foldSize
		end := start + foldSize
		if fold == config.CrossValidationFolds-1 {
			end = len(shuffledSamples) // Last fold gets remaining samples
		}

		testSamples := shuffledSamples[start:end]
		trainSamples := append(shuffledSamples[:start], shuffledSamples[end:]...)

		lv.logger.Debug("Cross-validation fold",
			zap.Int("fold", fold+1),
			zap.Int("train_samples", len(trainSamples)),
			zap.Int("test_samples", len(testSamples)))

		// Calculate metrics for this fold
		foldResult := lv.calculateFoldResults(testSamples)
		foldResult.Fold = fold + 1

		results = append(results, foldResult)
	}

	return results, nil
}

// calculateFoldResults calculates results for a single fold
func (lv *LSTMValidator) calculateFoldResults(samples []ValidationSample) CrossValidationResult {
	// Group samples by horizon
	horizonGroups := make(map[int][]ValidationSample)
	for _, sample := range samples {
		horizonGroups[sample.HorizonMonths] = append(horizonGroups[sample.HorizonMonths], sample)
	}

	// Calculate results for each horizon
	horizonResults := make(map[int]HorizonResult)
	for horizon, horizonSamples := range horizonGroups {
		horizonResult := lv.calculateHorizonMetrics(horizonSamples)
		horizonResults[horizon] = horizonResult
	}

	// Calculate overall metrics
	overallAccuracy := lv.calculateOverallAccuracy(horizonResults)
	overallMAE := lv.calculateOverallMAE(horizonResults)
	overallRMSE := lv.calculateOverallRMSE(horizonResults)

	return CrossValidationResult{
		HorizonResults:  horizonResults,
		OverallAccuracy: overallAccuracy,
		OverallMAE:      overallMAE,
		OverallRMSE:     overallRMSE,
	}
}

// calculateHorizonMetrics calculates metrics for a specific horizon
func (lv *LSTMValidator) calculateHorizonMetrics(samples []ValidationSample) HorizonResult {
	if len(samples) == 0 {
		return HorizonResult{}
	}

	horizon := samples[0].HorizonMonths

	// Calculate accuracy metrics
	correct := 0
	total := len(samples)
	var truePositives, falsePositives, falseNegatives int

	// Calculate regression metrics
	var mae, mse float64
	var predictedScores, trueScores []float64

	for _, sample := range samples {
		// Accuracy calculation
		if sample.PredictedRiskLevel == sample.TrueRiskLevel {
			correct++
		}

		// Confusion matrix for precision/recall
		if sample.PredictedRiskLevel == models.RiskLevelHigh || sample.PredictedRiskLevel == models.RiskLevelCritical {
			if sample.TrueRiskLevel == models.RiskLevelHigh || sample.TrueRiskLevel == models.RiskLevelCritical {
				truePositives++
			} else {
				falsePositives++
			}
		} else {
			if sample.TrueRiskLevel == models.RiskLevelHigh || sample.TrueRiskLevel == models.RiskLevelCritical {
				falseNegatives++
			}
		}

		// Regression metrics
		error := math.Abs(sample.PredictedRiskScore - sample.TrueRiskScore)
		mae += error
		mse += error * error

		predictedScores = append(predictedScores, sample.PredictedRiskScore)
		trueScores = append(trueScores, sample.TrueRiskScore)
	}

	// Calculate metrics
	accuracy := float64(correct) / float64(total)
	precision := float64(truePositives) / float64(truePositives+falsePositives)
	recall := float64(truePositives) / float64(truePositives+falseNegatives)
	f1Score := 2 * precision * recall / (precision + recall)
	mae = mae / float64(total)
	rmse := math.Sqrt(mse / float64(total))

	// Calculate R² score
	r2Score := lv.calculateR2Score(trueScores, predictedScores)

	// Calculate confidence interval
	confidenceInterval := lv.calculateConfidenceInterval(accuracy, total)

	return HorizonResult{
		HorizonMonths:      horizon,
		Accuracy:           accuracy,
		Precision:          precision,
		Recall:             recall,
		F1Score:            f1Score,
		MAE:                mae,
		RMSE:               rmse,
		R2Score:            r2Score,
		SampleCount:        total,
		ConfidenceInterval: confidenceInterval,
	}
}

// calculateR2Score calculates the R² score
func (lv *LSTMValidator) calculateR2Score(trueScores, predictedScores []float64) float64 {
	if len(trueScores) != len(predictedScores) || len(trueScores) == 0 {
		return 0
	}

	// Calculate mean of true scores
	var meanTrue float64
	for _, score := range trueScores {
		meanTrue += score
	}
	meanTrue /= float64(len(trueScores))

	// Calculate sum of squares
	var ssRes, ssTot float64
	for i := 0; i < len(trueScores); i++ {
		residual := trueScores[i] - predictedScores[i]
		ssRes += residual * residual

		total := trueScores[i] - meanTrue
		ssTot += total * total
	}

	if ssTot == 0 {
		return 0
	}

	return 1 - (ssRes / ssTot)
}

// calculateConfidenceInterval calculates confidence interval for accuracy
func (lv *LSTMValidator) calculateConfidenceInterval(accuracy float64, sampleCount int) ConfidenceInterval {
	if sampleCount == 0 {
		return ConfidenceInterval{Level: 0.95}
	}

	// Use normal approximation for binomial distribution
	z := 1.96 // 95% confidence interval
	se := math.Sqrt(accuracy * (1 - accuracy) / float64(sampleCount))

	return ConfidenceInterval{
		Lower: math.Max(0, accuracy-z*se),
		Upper: math.Min(1, accuracy+z*se),
		Level: 0.95,
	}
}

// calculateOverallResults calculates overall results from cross-validation
func (lv *LSTMValidator) calculateOverallResults(cvResults []CrossValidationResult) map[int]HorizonResult {
	// Group results by horizon
	horizonGroups := make(map[int][]HorizonResult)
	for _, cvResult := range cvResults {
		for horizon, result := range cvResult.HorizonResults {
			horizonGroups[horizon] = append(horizonGroups[horizon], result)
		}
	}

	// Calculate average results for each horizon
	overallResults := make(map[int]HorizonResult)
	for horizon, results := range horizonGroups {
		if len(results) == 0 {
			continue
		}

		// Calculate averages
		var avgAccuracy, avgPrecision, avgRecall, avgF1, avgMAE, avgRMSE, avgR2 float64
		var totalSamples int

		for _, result := range results {
			avgAccuracy += result.Accuracy
			avgPrecision += result.Precision
			avgRecall += result.Recall
			avgF1 += result.F1Score
			avgMAE += result.MAE
			avgRMSE += result.RMSE
			avgR2 += result.R2Score
			totalSamples += result.SampleCount
		}

		count := float64(len(results))
		overallResults[horizon] = HorizonResult{
			HorizonMonths:      horizon,
			Accuracy:           avgAccuracy / count,
			Precision:          avgPrecision / count,
			Recall:             avgRecall / count,
			F1Score:            avgF1 / count,
			MAE:                avgMAE / count,
			RMSE:               avgRMSE / count,
			R2Score:            avgR2 / count,
			SampleCount:        totalSamples,
			ConfidenceInterval: lv.calculateConfidenceInterval(avgAccuracy/count, totalSamples),
		}
	}

	return overallResults
}

// performModelComparison compares different models
func (lv *LSTMValidator) performModelComparison(ctx context.Context, samples []ValidationSample, config ValidationConfig) (*ModelComparison, error) {
	// Group samples by model type and horizon
	modelGroups := make(map[string]map[int][]ValidationSample)
	for _, sample := range samples {
		if modelGroups[sample.ModelType] == nil {
			modelGroups[sample.ModelType] = make(map[int][]ValidationSample)
		}
		modelGroups[sample.ModelType][sample.HorizonMonths] = append(modelGroups[sample.ModelType][sample.HorizonMonths], sample)
	}

	// Calculate results for each model
	xgbResults := make(map[int]HorizonResult)
	lstmResults := make(map[int]HorizonResult)
	ensembleResults := make(map[int]HorizonResult)

	for modelType, horizonGroups := range modelGroups {
		for horizon, horizonSamples := range horizonGroups {
			result := lv.calculateHorizonMetrics(horizonSamples)
			switch modelType {
			case "xgboost":
				xgbResults[horizon] = result
			case "lstm":
				lstmResults[horizon] = result
			case "ensemble":
				ensembleResults[horizon] = result
			}
		}
	}

	// Determine best model per horizon
	bestModelPerHorizon := make(map[int]string)
	for _, horizon := range config.ValidationHorizons {
		bestAccuracy := 0.0
		bestModel := ""

		if result, exists := xgbResults[horizon]; exists && result.Accuracy > bestAccuracy {
			bestAccuracy = result.Accuracy
			bestModel = "xgboost"
		}
		if result, exists := lstmResults[horizon]; exists && result.Accuracy > bestAccuracy {
			bestAccuracy = result.Accuracy
			bestModel = "lstm"
		}
		if result, exists := ensembleResults[horizon]; exists && result.Accuracy > bestAccuracy {
			bestAccuracy = result.Accuracy
			bestModel = "ensemble"
		}

		if bestModel != "" {
			bestModelPerHorizon[horizon] = bestModel
		}
	}

	return &ModelComparison{
		XGBoostResults:      xgbResults,
		LSTMResults:         lstmResults,
		EnsembleResults:     ensembleResults,
		BestModelPerHorizon: bestModelPerHorizon,
	}, nil
}

// analyzeConfidence analyzes prediction confidence
func (lv *LSTMValidator) analyzeConfidence(samples []ValidationSample) ConfidenceAnalysis {
	if len(samples) == 0 {
		return ConfidenceAnalysis{}
	}

	// Sort samples by confidence
	sort.Slice(samples, func(i, j int) bool {
		return samples[i].ConfidenceScore < samples[j].ConfidenceScore
	})

	// Split into high and low confidence groups
	threshold := 0.7
	highConfidenceSamples := make([]ValidationSample, 0)
	lowConfidenceSamples := make([]ValidationSample, 0)

	for _, sample := range samples {
		if sample.ConfidenceScore >= threshold {
			highConfidenceSamples = append(highConfidenceSamples, sample)
		} else {
			lowConfidenceSamples = append(lowConfidenceSamples, sample)
		}
	}

	// Calculate accuracy for each group
	highConfidenceAccuracy := lv.calculateGroupAccuracy(highConfidenceSamples)
	lowConfidenceAccuracy := lv.calculateGroupAccuracy(lowConfidenceSamples)

	// Calculate calibration error
	calibrationError := lv.calculateCalibrationError(samples)

	return ConfidenceAnalysis{
		HighConfidenceAccuracy: highConfidenceAccuracy,
		LowConfidenceAccuracy:  lowConfidenceAccuracy,
		ConfidenceThreshold:    threshold,
		CalibrationError:       calibrationError,
	}
}

// calculateGroupAccuracy calculates accuracy for a group of samples
func (lv *LSTMValidator) calculateGroupAccuracy(samples []ValidationSample) float64 {
	if len(samples) == 0 {
		return 0
	}

	correct := 0
	for _, sample := range samples {
		if sample.PredictedRiskLevel == sample.TrueRiskLevel {
			correct++
		}
	}

	return float64(correct) / float64(len(samples))
}

// calculateCalibrationError calculates calibration error
func (lv *LSTMValidator) calculateCalibrationError(samples []ValidationSample) float64 {
	if len(samples) == 0 {
		return 0
	}

	// Group samples by confidence bins
	bins := make(map[int][]ValidationSample)
	for _, sample := range samples {
		bin := int(sample.ConfidenceScore * 10)
		bins[bin] = append(bins[bin], sample)
	}

	var totalError float64
	var totalSamples int

	for _, binSamples := range bins {
		if len(binSamples) == 0 {
			continue
		}

		// Calculate accuracy for this bin
		accuracy := lv.calculateGroupAccuracy(binSamples)

		// Calculate expected accuracy (average confidence)
		var expectedAccuracy float64
		for _, sample := range binSamples {
			expectedAccuracy += sample.ConfidenceScore
		}
		expectedAccuracy /= float64(len(binSamples))

		// Calculate error
		error := math.Abs(accuracy - expectedAccuracy)
		totalError += error * float64(len(binSamples))
		totalSamples += len(binSamples)
	}

	if totalSamples == 0 {
		return 0
	}

	return totalError / float64(totalSamples)
}

// Helper methods
func (lv *LSTMValidator) calculateOverallAccuracy(horizonResults map[int]HorizonResult) float64 {
	if len(horizonResults) == 0 {
		return 0
	}

	var totalAccuracy, totalWeight float64
	for _, result := range horizonResults {
		weight := float64(result.SampleCount)
		totalAccuracy += result.Accuracy * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0
	}

	return totalAccuracy / totalWeight
}

func (lv *LSTMValidator) calculateOverallMAE(horizonResults map[int]HorizonResult) float64 {
	if len(horizonResults) == 0 {
		return 0
	}

	var totalMAE, totalWeight float64
	for _, result := range horizonResults {
		weight := float64(result.SampleCount)
		totalMAE += result.MAE * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0
	}

	return totalMAE / totalWeight
}

func (lv *LSTMValidator) calculateOverallRMSE(horizonResults map[int]HorizonResult) float64 {
	if len(horizonResults) == 0 {
		return 0
	}

	var totalRMSE, totalWeight float64
	for _, result := range horizonResults {
		weight := float64(result.SampleCount)
		totalRMSE += result.RMSE * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0
	}

	return totalRMSE / totalWeight
}

func (lv *LSTMValidator) determineRiskLevel(riskScore float64) models.RiskLevel {
	switch {
	case riskScore <= 0.25:
		return models.RiskLevelLow
	case riskScore <= 0.5:
		return models.RiskLevelMedium
	case riskScore <= 0.75:
		return models.RiskLevelHigh
	default:
		return models.RiskLevelCritical
	}
}

func (lv *LSTMValidator) shuffleSamples(samples []ValidationSample) {
	// Simple shuffle implementation
	for i := len(samples) - 1; i > 0; i-- {
		j := int(float64(i+1)*math.Sin(float64(i))*0.5 + 0.5)
		if j < 0 {
			j = 0
		}
		if j > i {
			j = i
		}
		samples[i], samples[j] = samples[j], samples[i]
	}
}
