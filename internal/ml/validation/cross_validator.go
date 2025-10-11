package validation

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// CrossValidator performs k-fold cross-validation on ML models
type CrossValidator struct {
	logger *zap.Logger
}

// ValidationResult contains the results of cross-validation
type ValidationResult struct {
	ModelName          string             `json:"model_name"`
	K                  int                `json:"k_folds"`
	TotalSamples       int                `json:"total_samples"`
	FoldResults        []FoldResult       `json:"fold_results"`
	OverallMetrics     OverallMetrics     `json:"overall_metrics"`
	ConfidenceInterval ConfidenceInterval `json:"confidence_interval"`
	ValidationTime     time.Duration      `json:"validation_time"`
	Timestamp          time.Time          `json:"timestamp"`
}

// FoldResult contains results for a single fold
type FoldResult struct {
	FoldIndex     int           `json:"fold_index"`
	TrainSamples  int           `json:"train_samples"`
	TestSamples   int           `json:"test_samples"`
	Metrics       Metrics       `json:"metrics"`
	TrainingTime  time.Duration `json:"training_time"`
	InferenceTime time.Duration `json:"inference_time"`
}

// Metrics contains standard ML performance metrics
type Metrics struct {
	Accuracy    float64 `json:"accuracy"`
	Precision   float64 `json:"precision"`
	Recall      float64 `json:"recall"`
	F1Score     float64 `json:"f1_score"`
	Specificity float64 `json:"specificity"`
	MCC         float64 `json:"matthews_correlation_coefficient"`
	AUC         float64 `json:"area_under_curve"`
	LogLoss     float64 `json:"log_loss"`
}

// OverallMetrics contains aggregated metrics across all folds
type OverallMetrics struct {
	MeanAccuracy  float64 `json:"mean_accuracy"`
	StdAccuracy   float64 `json:"std_accuracy"`
	MeanPrecision float64 `json:"mean_precision"`
	StdPrecision  float64 `json:"std_precision"`
	MeanRecall    float64 `json:"mean_recall"`
	StdRecall     float64 `json:"std_recall"`
	MeanF1Score   float64 `json:"mean_f1_score"`
	StdF1Score    float64 `json:"std_f1_score"`
	MeanAUC       float64 `json:"mean_auc"`
	StdAUC        float64 `json:"std_auc"`
	MeanLogLoss   float64 `json:"mean_log_loss"`
	StdLogLoss    float64 `json:"std_log_loss"`
}

// ConfidenceInterval contains confidence intervals for metrics
type ConfidenceInterval struct {
	Accuracy   Interval `json:"accuracy"`
	Precision  Interval `json:"precision"`
	Recall     Interval `json:"recall"`
	F1Score    Interval `json:"f1_score"`
	AUC        Interval `json:"auc"`
	Confidence float64  `json:"confidence_level"`
}

// Interval represents a confidence interval
type Interval struct {
	Lower float64 `json:"lower"`
	Upper float64 `json:"upper"`
}

// RiskSample represents a single risk assessment sample
type RiskSample struct {
	Features []float64              `json:"features"`
	Label    float64                `json:"label"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ModelValidator interface for models that can be cross-validated
type ModelValidator interface {
	Train(ctx context.Context, features [][]float64, labels []float64) error
	Predict(ctx context.Context, features [][]float64) ([]float64, error)
	PredictProba(ctx context.Context, features [][]float64) ([][]float64, error)
	GetName() string
}

// NewCrossValidator creates a new cross-validator
func NewCrossValidator(logger *zap.Logger) *CrossValidator {
	return &CrossValidator{
		logger: logger,
	}
}

// CrossValidate performs k-fold cross-validation on a model
func (cv *CrossValidator) CrossValidate(
	ctx context.Context,
	model ModelValidator,
	samples []RiskSample,
	k int,
	confidenceLevel float64,
) (*ValidationResult, error) {
	startTime := time.Now()

	cv.logger.Info("Starting cross-validation",
		zap.String("model", model.GetName()),
		zap.Int("k_folds", k),
		zap.Int("total_samples", len(samples)),
		zap.Float64("confidence_level", confidenceLevel))

	if len(samples) < k {
		return nil, fmt.Errorf("insufficient samples for %d-fold cross-validation: have %d, need at least %d", k, len(samples), k)
	}

	// Shuffle samples for random distribution
	shuffledSamples := make([]RiskSample, len(samples))
	copy(shuffledSamples, samples)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(shuffledSamples), func(i, j int) {
		shuffledSamples[i], shuffledSamples[j] = shuffledSamples[j], shuffledSamples[i]
	})

	// Create folds
	folds := cv.createFolds(shuffledSamples, k)

	// Perform cross-validation
	foldResults := make([]FoldResult, k)

	for i := 0; i < k; i++ {
		cv.logger.Debug("Processing fold",
			zap.Int("fold", i+1),
			zap.Int("total_folds", k))

		foldResult, err := cv.validateFold(ctx, model, folds, i)
		if err != nil {
			return nil, fmt.Errorf("fold %d validation failed: %w", i+1, err)
		}

		foldResults[i] = *foldResult
	}

	// Calculate overall metrics
	overallMetrics := cv.calculateOverallMetrics(foldResults)

	// Calculate confidence intervals
	confidenceInterval := cv.calculateConfidenceIntervals(foldResults, confidenceLevel)

	validationTime := time.Since(startTime)

	result := &ValidationResult{
		ModelName:          model.GetName(),
		K:                  k,
		TotalSamples:       len(samples),
		FoldResults:        foldResults,
		OverallMetrics:     overallMetrics,
		ConfidenceInterval: confidenceInterval,
		ValidationTime:     validationTime,
		Timestamp:          time.Now(),
	}

	cv.logger.Info("Cross-validation completed",
		zap.String("model", model.GetName()),
		zap.Float64("mean_accuracy", overallMetrics.MeanAccuracy),
		zap.Float64("mean_f1_score", overallMetrics.MeanF1Score),
		zap.Duration("validation_time", validationTime))

	return result, nil
}

// createFolds creates k folds from the samples
func (cv *CrossValidator) createFolds(samples []RiskSample, k int) [][]RiskSample {
	foldSize := len(samples) / k
	folds := make([][]RiskSample, k)

	start := 0
	for i := 0; i < k; i++ {
		end := start + foldSize
		if i == k-1 {
			// Last fold gets any remaining samples
			end = len(samples)
		}
		folds[i] = samples[start:end]
		start = end
	}

	return folds
}

// validateFold validates a single fold
func (cv *CrossValidator) validateFold(
	ctx context.Context,
	model ModelValidator,
	folds [][]RiskSample,
	testFoldIndex int,
) (*FoldResult, error) {
	// Prepare training and test data
	var trainSamples, testSamples []RiskSample

	for i, fold := range folds {
		if i == testFoldIndex {
			testSamples = fold
		} else {
			trainSamples = append(trainSamples, fold...)
		}
	}

	// Convert to feature/label format
	trainFeatures, trainLabels := cv.samplesToFeaturesLabels(trainSamples)
	testFeatures, testLabels := cv.samplesToFeaturesLabels(testSamples)

	// Train model
	trainStart := time.Now()
	if err := model.Train(ctx, trainFeatures, trainLabels); err != nil {
		return nil, fmt.Errorf("training failed: %w", err)
	}
	trainingTime := time.Since(trainStart)

	// Make predictions
	inferenceStart := time.Now()
	predictions, err := model.Predict(ctx, testFeatures)
	if err != nil {
		return nil, fmt.Errorf("prediction failed: %w", err)
	}
	inferenceTime := time.Since(inferenceStart)

	// Calculate metrics
	metrics := cv.calculateMetrics(testLabels, predictions)

	return &FoldResult{
		FoldIndex:     testFoldIndex + 1,
		TrainSamples:  len(trainSamples),
		TestSamples:   len(testSamples),
		Metrics:       metrics,
		TrainingTime:  trainingTime,
		InferenceTime: inferenceTime,
	}, nil
}

// samplesToFeaturesLabels converts samples to feature and label arrays
func (cv *CrossValidator) samplesToFeaturesLabels(samples []RiskSample) ([][]float64, []float64) {
	features := make([][]float64, len(samples))
	labels := make([]float64, len(samples))

	for i, sample := range samples {
		features[i] = sample.Features
		labels[i] = sample.Label
	}

	return features, labels
}

// calculateMetrics calculates performance metrics
func (cv *CrossValidator) calculateMetrics(trueLabels, predictions []float64) Metrics {
	if len(trueLabels) != len(predictions) {
		panic("true labels and predictions must have the same length")
	}

	// Convert to binary classification (0 or 1)
	trueBinary := cv.toBinary(trueLabels)
	predBinary := cv.toBinary(predictions)

	// Calculate confusion matrix
	tp, tn, fp, fn := cv.calculateConfusionMatrix(trueBinary, predBinary)

	// Calculate metrics
	accuracy := float64(tp+tn) / float64(tp+tn+fp+fn)
	precision := float64(tp) / float64(tp+fp)
	recall := float64(tp) / float64(tp+fn)
	f1Score := 2 * (precision * recall) / (precision + recall)
	specificity := float64(tn) / float64(tn+fp)

	// Matthews Correlation Coefficient
	mcc := cv.calculateMCC(tp, tn, fp, fn)

	// AUC (simplified calculation)
	auc := cv.calculateAUC(trueLabels, predictions)

	// Log Loss
	logLoss := cv.calculateLogLoss(trueLabels, predictions)

	return Metrics{
		Accuracy:    accuracy,
		Precision:   precision,
		Recall:      recall,
		F1Score:     f1Score,
		Specificity: specificity,
		MCC:         mcc,
		AUC:         auc,
		LogLoss:     logLoss,
	}
}

// toBinary converts continuous values to binary (0 or 1)
func (cv *CrossValidator) toBinary(values []float64) []int {
	binary := make([]int, len(values))
	for i, v := range values {
		if v >= 0.5 {
			binary[i] = 1
		} else {
			binary[i] = 0
		}
	}
	return binary
}

// calculateConfusionMatrix calculates true positives, true negatives, false positives, false negatives
func (cv *CrossValidator) calculateConfusionMatrix(trueLabels, predictions []int) (tp, tn, fp, fn int) {
	for i := 0; i < len(trueLabels); i++ {
		trueLabel := trueLabels[i]
		prediction := predictions[i]

		if trueLabel == 1 && prediction == 1 {
			tp++
		} else if trueLabel == 0 && prediction == 0 {
			tn++
		} else if trueLabel == 0 && prediction == 1 {
			fp++
		} else if trueLabel == 1 && prediction == 0 {
			fn++
		}
	}
	return
}

// calculateMCC calculates Matthews Correlation Coefficient
func (cv *CrossValidator) calculateMCC(tp, tn, fp, fn int) float64 {
	numerator := float64(tp*tn - fp*fn)
	denominator := math.Sqrt(float64((tp + fp) * (tp + fn) * (tn + fp) * (tn + fn)))

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

// calculateAUC calculates Area Under the Curve (simplified)
func (cv *CrossValidator) calculateAUC(trueLabels, predictions []float64) float64 {
	// Simple AUC calculation - in practice, you'd use a more sophisticated method
	// This is a placeholder implementation
	correct := 0
	total := 0

	for i := 0; i < len(trueLabels); i++ {
		for j := i + 1; j < len(trueLabels); j++ {
			if (trueLabels[i] > trueLabels[j] && predictions[i] > predictions[j]) ||
				(trueLabels[i] < trueLabels[j] && predictions[i] < predictions[j]) {
				correct++
			}
			total++
		}
	}

	if total == 0 {
		return 0.5
	}

	return float64(correct) / float64(total)
}

// calculateLogLoss calculates logarithmic loss
func (cv *CrossValidator) calculateLogLoss(trueLabels, predictions []float64) float64 {
	sum := 0.0
	for i := 0; i < len(trueLabels); i++ {
		p := math.Max(math.Min(predictions[i], 1-1e-15), 1e-15)
		sum += trueLabels[i]*math.Log(p) + (1-trueLabels[i])*math.Log(1-p)
	}
	return -sum / float64(len(trueLabels))
}

// calculateOverallMetrics calculates mean and standard deviation across folds
func (cv *CrossValidator) calculateOverallMetrics(foldResults []FoldResult) OverallMetrics {
	accuracies := make([]float64, len(foldResults))
	precisions := make([]float64, len(foldResults))
	recalls := make([]float64, len(foldResults))
	f1Scores := make([]float64, len(foldResults))
	aucs := make([]float64, len(foldResults))
	logLosses := make([]float64, len(foldResults))

	for i, result := range foldResults {
		accuracies[i] = result.Metrics.Accuracy
		precisions[i] = result.Metrics.Precision
		recalls[i] = result.Metrics.Recall
		f1Scores[i] = result.Metrics.F1Score
		aucs[i] = result.Metrics.AUC
		logLosses[i] = result.Metrics.LogLoss
	}

	return OverallMetrics{
		MeanAccuracy:  cv.mean(accuracies),
		StdAccuracy:   cv.std(accuracies),
		MeanPrecision: cv.mean(precisions),
		StdPrecision:  cv.std(precisions),
		MeanRecall:    cv.mean(recalls),
		StdRecall:     cv.std(recalls),
		MeanF1Score:   cv.mean(f1Scores),
		StdF1Score:    cv.std(f1Scores),
		MeanAUC:       cv.mean(aucs),
		StdAUC:        cv.std(aucs),
		MeanLogLoss:   cv.mean(logLosses),
		StdLogLoss:    cv.std(logLosses),
	}
}

// calculateConfidenceIntervals calculates confidence intervals for metrics
func (cv *CrossValidator) calculateConfidenceIntervals(foldResults []FoldResult, confidenceLevel float64) ConfidenceInterval {
	accuracies := make([]float64, len(foldResults))
	precisions := make([]float64, len(foldResults))
	recalls := make([]float64, len(foldResults))
	f1Scores := make([]float64, len(foldResults))
	aucs := make([]float64, len(foldResults))

	for i, result := range foldResults {
		accuracies[i] = result.Metrics.Accuracy
		precisions[i] = result.Metrics.Precision
		recalls[i] = result.Metrics.Recall
		f1Scores[i] = result.Metrics.F1Score
		aucs[i] = result.Metrics.AUC
	}

	// Calculate t-value for confidence interval (simplified)
	tValue := cv.getTValue(len(foldResults), confidenceLevel)

	return ConfidenceInterval{
		Accuracy:   cv.calculateInterval(accuracies, tValue),
		Precision:  cv.calculateInterval(precisions, tValue),
		Recall:     cv.calculateInterval(recalls, tValue),
		F1Score:    cv.calculateInterval(f1Scores, tValue),
		AUC:        cv.calculateInterval(aucs, tValue),
		Confidence: confidenceLevel,
	}
}

// calculateInterval calculates confidence interval for a metric
func (cv *CrossValidator) calculateInterval(values []float64, tValue float64) Interval {
	mean := cv.mean(values)
	std := cv.std(values)
	margin := tValue * std / math.Sqrt(float64(len(values)))

	return Interval{
		Lower: mean - margin,
		Upper: mean + margin,
	}
}

// mean calculates the mean of a slice of float64 values
func (cv *CrossValidator) mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// std calculates the standard deviation of a slice of float64 values
func (cv *CrossValidator) std(values []float64) float64 {
	if len(values) <= 1 {
		return 0
	}

	mean := cv.mean(values)
	sumSquaredDiffs := 0.0

	for _, v := range values {
		diff := v - mean
		sumSquaredDiffs += diff * diff
	}

	return math.Sqrt(sumSquaredDiffs / float64(len(values)-1))
}

// getTValue returns t-value for confidence interval (simplified)
func (cv *CrossValidator) getTValue(n int, confidenceLevel float64) float64 {
	// Simplified t-value lookup - in practice, you'd use proper t-distribution
	// For 95% confidence and n-1 degrees of freedom
	if confidenceLevel >= 0.95 {
		switch {
		case n <= 2:
			return 12.706
		case n <= 3:
			return 4.303
		case n <= 4:
			return 3.182
		case n <= 5:
			return 2.776
		case n <= 10:
			return 2.262
		default:
			return 1.96 // Approximate normal distribution
		}
	}
	return 1.96 // Default to 95% confidence
}
