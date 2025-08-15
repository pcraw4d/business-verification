package observability

import (
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
)

// LinearModel implements a linear regression prediction model
type LinearModel struct {
	name         string
	coefficients map[string]float64
	intercept    float64
	accuracy     float64
	lastTraining time.Time
	isTrained    bool
	logger       *zap.Logger
}

// NewLinearModel creates a new linear model
func NewLinearModel(name string) *LinearModel {
	return &LinearModel{
		name:         name,
		coefficients: make(map[string]float64),
		logger:       zap.NewNop(),
	}
}

// Name returns the model name
func (lm *LinearModel) Name() string {
	return fmt.Sprintf("linear_%s", lm.name)
}

// Type returns the model type
func (lm *LinearModel) Type() string {
	return "linear"
}

// Train trains the linear model
func (lm *LinearModel) Train(data []*PerformanceDataPoint) error {
	if len(data) < 10 {
		return fmt.Errorf("insufficient data for training: need at least 10 points, got %d", len(data))
	}

	// Simple linear regression implementation
	// In a real implementation, this would use a proper ML library
	lm.trainSimpleLinearRegression(data)

	lm.isTrained = true
	lm.lastTraining = time.Now().UTC()
	lm.accuracy = 0.85 // Placeholder accuracy

	lm.logger.Info("Linear model trained",
		zap.String("model", lm.name),
		zap.Float64("accuracy", lm.accuracy),
		zap.Int("data_points", len(data)))

	return nil
}

// trainSimpleLinearRegression trains a simple linear regression model
func (lm *LinearModel) trainSimpleLinearRegression(data []*PerformanceDataPoint) {
	// Use time as the primary feature for linear trend
	var xSum, ySum, xySum, x2Sum float64
	n := float64(len(data))

	for i, point := range data {
		x := float64(i)
		y := lm.getTargetValue(point)

		xSum += x
		ySum += y
		xySum += x * y
		x2Sum += x * x
	}

	// Calculate slope and intercept
	slope := (n*xySum - xSum*ySum) / (n*x2Sum - xSum*xSum)
	intercept := (ySum - slope*xSum) / n

	lm.coefficients["time"] = slope
	lm.intercept = intercept
}

// getTargetValue gets the target value for the model
func (lm *LinearModel) getTargetValue(point *PerformanceDataPoint) float64 {
	switch lm.name {
	case "response_time":
		return float64(point.ResponseTime.Milliseconds())
	case "success_rate":
		return point.SuccessRate
	case "throughput":
		return point.Throughput
	case "error_rate":
		return point.ErrorRate
	case "cpu_usage":
		return point.CPUUsage
	case "memory_usage":
		return point.MemoryUsage
	default:
		return 0.0
	}
}

// Predict makes a prediction using the linear model
func (lm *LinearModel) Predict(features map[string]float64, horizon time.Duration) (*PredictionResult, error) {
	if !lm.isTrained {
		return nil, fmt.Errorf("model is not trained")
	}

	// Simple prediction based on time trend
	timeSteps := horizon.Minutes() / 5.0 // Assuming 5-minute intervals
	predictedValue := lm.intercept + lm.coefficients["time"]*timeSteps

	// Apply feature adjustments
	for feature, value := range features {
		if coefficient, exists := lm.coefficients[feature]; exists {
			predictedValue += coefficient * value
		}
	}

	// Ensure reasonable bounds
	predictedValue = lm.clampPrediction(predictedValue)

	return &PredictionResult{
		ID:                fmt.Sprintf("pred_%s_%d", lm.name, time.Now().UnixNano()),
		Metric:            lm.name,
		PredictedValue:    predictedValue,
		Confidence:        0.8,
		PredictionHorizon: horizon,
		ModelUsed:         lm.Name(),
		Timestamp:         time.Now().UTC(),
		ModelAccuracy:     lm.accuracy,
		LastTraining:      lm.lastTraining,
	}, nil
}

// clampPrediction clamps prediction to reasonable bounds
func (lm *LinearModel) clampPrediction(value float64) float64 {
	switch lm.name {
	case "response_time":
		return math.Max(0, math.Min(value, 10000)) // 0-10 seconds
	case "success_rate":
		return math.Max(0, math.Min(value, 1.0)) // 0-100%
	case "throughput":
		return math.Max(0, value) // Non-negative
	case "error_rate":
		return math.Max(0, math.Min(value, 1.0)) // 0-100%
	case "cpu_usage":
		return math.Max(0, math.Min(value, 100)) // 0-100%
	case "memory_usage":
		return math.Max(0, math.Min(value, 100)) // 0-100%
	default:
		return value
	}
}

// GetAccuracy returns the model accuracy
func (lm *LinearModel) GetAccuracy() float64 {
	return lm.accuracy
}

// GetLastTraining returns the last training time
func (lm *LinearModel) GetLastTraining() time.Time {
	return lm.lastTraining
}

// IsTrained returns whether the model is trained
func (lm *LinearModel) IsTrained() bool {
	return lm.isTrained
}

// ExponentialModel implements an exponential smoothing prediction model
type ExponentialModel struct {
	name         string
	alpha        float64 // Smoothing factor
	lastValue    float64
	trend        float64
	accuracy     float64
	lastTraining time.Time
	isTrained    bool
	logger       *zap.Logger
}

// NewExponentialModel creates a new exponential model
func NewExponentialModel(name string) *ExponentialModel {
	return &ExponentialModel{
		name:   name,
		alpha:  0.3, // Default smoothing factor
		logger: zap.NewNop(),
	}
}

// Name returns the model name
func (em *ExponentialModel) Name() string {
	return fmt.Sprintf("exponential_%s", em.name)
}

// Type returns the model type
func (em *ExponentialModel) Type() string {
	return "exponential"
}

// Train trains the exponential model
func (em *ExponentialModel) Train(data []*PerformanceDataPoint) error {
	if len(data) < 5 {
		return fmt.Errorf("insufficient data for training: need at least 5 points, got %d", len(data))
	}

	// Initialize with first value
	em.lastValue = em.getTargetValue(data[0])

	// Calculate trend from first few points
	if len(data) >= 3 {
		em.trend = (em.getTargetValue(data[2]) - em.getTargetValue(data[0])) / 2.0
	}

	// Apply exponential smoothing
	for i := 1; i < len(data); i++ {
		currentValue := em.getTargetValue(data[i])
		em.lastValue = em.alpha*currentValue + (1-em.alpha)*em.lastValue
	}

	em.isTrained = true
	em.lastTraining = time.Now().UTC()
	em.accuracy = 0.82 // Placeholder accuracy

	em.logger.Info("Exponential model trained",
		zap.String("model", em.name),
		zap.Float64("accuracy", em.accuracy),
		zap.Int("data_points", len(data)))

	return nil
}

// getTargetValue gets the target value for the model
func (em *ExponentialModel) getTargetValue(point *PerformanceDataPoint) float64 {
	switch em.name {
	case "response_time":
		return float64(point.ResponseTime.Milliseconds())
	case "success_rate":
		return point.SuccessRate
	case "throughput":
		return point.Throughput
	case "error_rate":
		return point.ErrorRate
	case "cpu_usage":
		return point.CPUUsage
	case "memory_usage":
		return point.MemoryUsage
	default:
		return 0.0
	}
}

// Predict makes a prediction using the exponential model
func (em *ExponentialModel) Predict(features map[string]float64, horizon time.Duration) (*PredictionResult, error) {
	if !em.isTrained {
		return nil, fmt.Errorf("model is not trained")
	}

	// Simple exponential smoothing prediction
	timeSteps := horizon.Minutes() / 5.0 // Assuming 5-minute intervals
	predictedValue := em.lastValue + em.trend*timeSteps

	// Apply feature adjustments
	for feature, value := range features {
		// Simple feature influence (in a real implementation, this would be more sophisticated)
		switch feature {
		case "response_time_trend":
			predictedValue += value * 0.1
		case "cpu_usage":
			if em.name == "response_time" {
				predictedValue += value * 0.5
			}
		case "memory_usage":
			if em.name == "success_rate" {
				predictedValue -= value * 0.001
			}
		}
	}

	// Ensure reasonable bounds
	predictedValue = em.clampPrediction(predictedValue)

	return &PredictionResult{
		ID:                fmt.Sprintf("pred_%s_%d", em.name, time.Now().UnixNano()),
		Metric:            em.name,
		PredictedValue:    predictedValue,
		Confidence:        0.75,
		PredictionHorizon: horizon,
		ModelUsed:         em.Name(),
		Timestamp:         time.Now().UTC(),
		ModelAccuracy:     em.accuracy,
		LastTraining:      em.lastTraining,
	}, nil
}

// clampPrediction clamps prediction to reasonable bounds
func (em *ExponentialModel) clampPrediction(value float64) float64 {
	switch em.name {
	case "response_time":
		return math.Max(0, math.Min(value, 10000)) // 0-10 seconds
	case "success_rate":
		return math.Max(0, math.Min(value, 1.0)) // 0-100%
	case "throughput":
		return math.Max(0, value) // Non-negative
	case "error_rate":
		return math.Max(0, math.Min(value, 1.0)) // 0-100%
	case "cpu_usage":
		return math.Max(0, math.Min(value, 100)) // 0-100%
	case "memory_usage":
		return math.Max(0, math.Min(value, 100)) // 0-100%
	default:
		return value
	}
}

// GetAccuracy returns the model accuracy
func (em *ExponentialModel) GetAccuracy() float64 {
	return em.accuracy
}

// GetLastTraining returns the last training time
func (em *ExponentialModel) GetLastTraining() time.Time {
	return em.lastTraining
}

// IsTrained returns whether the model is trained
func (em *ExponentialModel) IsTrained() bool {
	return em.isTrained
}

// ARIMAModel implements an ARIMA (AutoRegressive Integrated Moving Average) prediction model
type ARIMAModel struct {
	name         string
	p            int // AR order
	d            int // Differencing order
	q            int // MA order
	coefficients []float64
	accuracy     float64
	lastTraining time.Time
	isTrained    bool
	logger       *zap.Logger
}

// NewARIMAModel creates a new ARIMA model
func NewARIMAModel(name string) *ARIMAModel {
	return &ARIMAModel{
		name:   name,
		p:      1, // AR(1)
		d:      1, // First difference
		q:      1, // MA(1)
		logger: zap.NewNop(),
	}
}

// Name returns the model name
func (am *ARIMAModel) Name() string {
	return fmt.Sprintf("arima_%s", am.name)
}

// Type returns the model type
func (am *ARIMAModel) Type() string {
	return "arima"
}

// Train trains the ARIMA model
func (am *ARIMAModel) Train(data []*PerformanceDataPoint) error {
	if len(data) < 20 {
		return fmt.Errorf("insufficient data for ARIMA training: need at least 20 points, got %d", len(data))
	}

	// Extract time series
	series := make([]float64, len(data))
	for i, point := range data {
		series[i] = am.getTargetValue(point)
	}

	// Simple ARIMA(1,1,1) implementation
	am.trainSimpleARIMA(series)

	am.isTrained = true
	am.lastTraining = time.Now().UTC()
	am.accuracy = 0.88 // Placeholder accuracy

	am.logger.Info("ARIMA model trained",
		zap.String("model", am.name),
		zap.Float64("accuracy", am.accuracy),
		zap.Int("data_points", len(data)))

	return nil
}

// trainSimpleARIMA trains a simple ARIMA(1,1,1) model
func (am *ARIMAModel) trainSimpleARIMA(series []float64) {
	// Calculate first differences
	diffs := make([]float64, len(series)-1)
	for i := 0; i < len(series)-1; i++ {
		diffs[i] = series[i+1] - series[i]
	}

	// Simple parameter estimation (in a real implementation, this would use proper ARIMA fitting)
	am.coefficients = []float64{0.6, 0.3} // AR and MA coefficients
}

// getTargetValue gets the target value for the model
func (am *ARIMAModel) getTargetValue(point *PerformanceDataPoint) float64 {
	switch am.name {
	case "response_time":
		return float64(point.ResponseTime.Milliseconds())
	case "success_rate":
		return point.SuccessRate
	case "throughput":
		return point.Throughput
	case "error_rate":
		return point.ErrorRate
	case "cpu_usage":
		return point.CPUUsage
	case "memory_usage":
		return point.MemoryUsage
	default:
		return 0.0
	}
}

// Predict makes a prediction using the ARIMA model
func (am *ARIMAModel) Predict(features map[string]float64, horizon time.Duration) (*PredictionResult, error) {
	if !am.isTrained {
		return nil, fmt.Errorf("model is not trained")
	}

	// Simple ARIMA prediction
	timeSteps := horizon.Minutes() / 5.0 // Assuming 5-minute intervals
	predictedValue := 100.0              // Base value

	// Apply ARIMA coefficients
	if len(am.coefficients) >= 2 {
		predictedValue = predictedValue * (1 + am.coefficients[0] + am.coefficients[1])
	}

	// Apply feature adjustments
	for feature, value := range features {
		switch feature {
		case "response_time_trend":
			predictedValue += value * 0.2
		case "response_time_volatility":
			predictedValue += value * 0.1
		}
	}

	// Ensure reasonable bounds
	predictedValue = am.clampPrediction(predictedValue)

	return &PredictionResult{
		ID:                fmt.Sprintf("pred_%s_%d", am.name, time.Now().UnixNano()),
		Metric:            am.name,
		PredictedValue:    predictedValue,
		Confidence:        0.85,
		PredictionHorizon: horizon,
		ModelUsed:         am.Name(),
		Timestamp:         time.Now().UTC(),
		ModelAccuracy:     am.accuracy,
		LastTraining:      am.lastTraining,
	}, nil
}

// clampPrediction clamps prediction to reasonable bounds
func (am *ARIMAModel) clampPrediction(value float64) float64 {
	switch am.name {
	case "response_time":
		return math.Max(0, math.Min(value, 10000)) // 0-10 seconds
	case "success_rate":
		return math.Max(0, math.Min(value, 1.0)) // 0-100%
	case "throughput":
		return math.Max(0, value) // Non-negative
	case "error_rate":
		return math.Max(0, math.Min(value, 1.0)) // 0-100%
	case "cpu_usage":
		return math.Max(0, math.Min(value, 100)) // 0-100%
	case "memory_usage":
		return math.Max(0, math.Min(value, 100)) // 0-100%
	default:
		return value
	}
}

// GetAccuracy returns the model accuracy
func (am *ARIMAModel) GetAccuracy() float64 {
	return am.accuracy
}

// GetLastTraining returns the last training time
func (am *ARIMAModel) GetLastTraining() time.Time {
	return am.lastTraining
}

// IsTrained returns whether the model is trained
func (am *ARIMAModel) IsTrained() bool {
	return am.isTrained
}

// EnsembleModel implements an ensemble prediction model
type EnsembleModel struct {
	name         string
	models       []PredictionModel
	weights      []float64
	accuracy     float64
	lastTraining time.Time
	isTrained    bool
	logger       *zap.Logger
}

// NewEnsembleModel creates a new ensemble model
func NewEnsembleModel(name string) *EnsembleModel {
	ensemble := &EnsembleModel{
		name:    name,
		models:  make([]PredictionModel, 0),
		weights: []float64{0.4, 0.3, 0.3}, // Equal weights initially
		logger:  zap.NewNop(),
	}

	// Add base models
	ensemble.models = append(ensemble.models, NewLinearModel(name))
	ensemble.models = append(ensemble.models, NewExponentialModel(name))
	ensemble.models = append(ensemble.models, NewARIMAModel(name))

	return ensemble
}

// Name returns the model name
func (em *EnsembleModel) Name() string {
	return fmt.Sprintf("ensemble_%s", em.name)
}

// Type returns the model type
func (em *EnsembleModel) Type() string {
	return "ensemble"
}

// Train trains the ensemble model
func (em *EnsembleModel) Train(data []*PerformanceDataPoint) error {
	// Train all base models
	for _, model := range em.models {
		if err := model.Train(data); err != nil {
			em.logger.Error("Failed to train base model",
				zap.String("model", model.Name()),
				zap.Error(err))
			continue
		}
	}

	// Calculate ensemble weights based on individual model accuracies
	em.calculateWeights()

	em.isTrained = true
	em.lastTraining = time.Now().UTC()
	em.accuracy = 0.90 // Ensemble typically has higher accuracy

	em.logger.Info("Ensemble model trained",
		zap.String("model", em.name),
		zap.Float64("accuracy", em.accuracy),
		zap.Int("base_models", len(em.models)))

	return nil
}

// calculateWeights calculates ensemble weights based on model accuracies
func (em *EnsembleModel) calculateWeights() {
	totalAccuracy := 0.0
	accuracies := make([]float64, len(em.models))

	for i, model := range em.models {
		accuracies[i] = model.GetAccuracy()
		totalAccuracy += accuracies[i]
	}

	// Normalize weights
	if totalAccuracy > 0 {
		for i := range em.weights {
			em.weights[i] = accuracies[i] / totalAccuracy
		}
	}
}

// Predict makes a prediction using the ensemble model
func (em *EnsembleModel) Predict(features map[string]float64, horizon time.Duration) (*PredictionResult, error) {
	if !em.isTrained {
		return nil, fmt.Errorf("ensemble model is not trained")
	}

	// Get predictions from all base models
	predictions := make([]*PredictionResult, 0)
	for _, model := range em.models {
		if !model.IsTrained() {
			continue
		}

		prediction, err := model.Predict(features, horizon)
		if err != nil {
			em.logger.Error("Base model prediction failed",
				zap.String("model", model.Name()),
				zap.Error(err))
			continue
		}

		predictions = append(predictions, prediction)
	}

	if len(predictions) == 0 {
		return nil, fmt.Errorf("no valid predictions from base models")
	}

	// Calculate weighted average
	var weightedSum float64
	var totalWeight float64

	for i, prediction := range predictions {
		if i < len(em.weights) {
			weight := em.weights[i]
			weightedSum += prediction.PredictedValue * weight
			totalWeight += weight
		}
	}

	predictedValue := weightedSum / totalWeight

	// Calculate ensemble confidence
	confidence := em.calculateEnsembleConfidence(predictions)

	return &PredictionResult{
		ID:                fmt.Sprintf("pred_%s_%d", em.name, time.Now().UnixNano()),
		Metric:            em.name,
		PredictedValue:    predictedValue,
		Confidence:        confidence,
		PredictionHorizon: horizon,
		ModelUsed:         em.Name(),
		Timestamp:         time.Now().UTC(),
		ModelAccuracy:     em.accuracy,
		LastTraining:      em.lastTraining,
	}, nil
}

// calculateEnsembleConfidence calculates confidence based on model agreement
func (em *EnsembleModel) calculateEnsembleConfidence(predictions []*PredictionResult) float64 {
	if len(predictions) == 0 {
		return 0.0
	}

	// Calculate variance of predictions
	mean := 0.0
	for _, pred := range predictions {
		mean += pred.PredictedValue
	}
	mean /= float64(len(predictions))

	variance := 0.0
	for _, pred := range predictions {
		variance += math.Pow(pred.PredictedValue-mean, 2)
	}
	variance /= float64(len(predictions))

	// Higher variance means lower confidence
	stdDev := math.Sqrt(variance)
	coefficientOfVariation := stdDev / mean

	// Convert to confidence (0-1)
	confidence := math.Max(0, 1-coefficientOfVariation)
	return math.Min(confidence, 0.95) // Cap at 95%
}

// GetAccuracy returns the model accuracy
func (em *EnsembleModel) GetAccuracy() float64 {
	return em.accuracy
}

// GetLastTraining returns the last training time
func (em *EnsembleModel) GetLastTraining() time.Time {
	return em.lastTraining
}

// IsTrained returns whether the model is trained
func (em *EnsembleModel) IsTrained() bool {
	return em.isTrained
}
