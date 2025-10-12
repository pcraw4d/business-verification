package testing

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ABTestManager manages A/B testing for ML models
type ABTestManager struct {
	experiments map[string]*Experiment
	metrics     *MetricsTracker
	logger      *zap.Logger
	mu          sync.RWMutex
}

// Experiment represents an A/B test experiment
type Experiment struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Status          ExperimentStatus       `json:"status"`
	TrafficSplit    map[string]float64     `json:"traffic_split"` // model_id -> percentage
	Models          map[string]ModelConfig `json:"models"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         *time.Time             `json:"end_time,omitempty"`
	SuccessMetrics  []string               `json:"success_metrics"`
	MinSampleSize   int                    `json:"min_sample_size"`
	ConfidenceLevel float64                `json:"confidence_level"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// ExperimentResult represents the results of an experiment
type ExperimentResult struct {
	ExperimentID    string                  `json:"experiment_id"`
	TotalRequests   int64                   `json:"total_requests"`
	Duration        time.Duration           `json:"duration"`
	ModelResults    map[string]*ModelResult `json:"model_results"`
	StatisticalTest *StatisticalTestResult  `json:"statistical_test"`
	Winner          string                  `json:"winner,omitempty"`
	Confidence      float64                 `json:"confidence"`
	IsSignificant   bool                    `json:"is_significant"`
	Recommendation  string                  `json:"recommendation"`
	CompletedAt     time.Time               `json:"completed_at"`
}

// ModelResult represents results for a specific model in an experiment
type ModelResult struct {
	ModelID         string             `json:"model_id"`
	RequestCount    int64              `json:"request_count"`
	Metrics         map[string]float64 `json:"metrics"`
	AverageLatency  time.Duration      `json:"average_latency"`
	ErrorRate       float64            `json:"error_rate"`
	ConfidenceScore float64            `json:"confidence_score"`
	Accuracy        float64            `json:"accuracy"`
	Precision       float64            `json:"precision"`
	Recall          float64            `json:"recall"`
	F1Score         float64            `json:"f1_score"`
}

// StatisticalTestResult represents the results of statistical significance testing
type StatisticalTestResult struct {
	TestType           string     `json:"test_type"`
	PValue             float64    `json:"p_value"`
	IsSignificant      bool       `json:"is_significant"`
	EffectSize         float64    `json:"effect_size"`
	ConfidenceInterval [2]float64 `json:"confidence_interval"`
	Power              float64    `json:"power"`
}

// NewABTestManager creates a new A/B test manager
func NewABTestManager(logger *zap.Logger) *ABTestManager {
	return &ABTestManager{
		experiments: make(map[string]*Experiment),
		metrics:     NewMetricsTracker(logger),
		logger:      logger,
	}
}

// CreateExperiment creates a new A/B test experiment
func (abm *ABTestManager) CreateExperiment(ctx context.Context, config *ExperimentConfig) (*Experiment, error) {
	abm.mu.Lock()
	defer abm.mu.Unlock()

	// Validate traffic split
	totalSplit := 0.0
	for _, split := range config.TrafficSplit {
		totalSplit += split
	}
	if totalSplit != 1.0 {
		return nil, fmt.Errorf("traffic split must sum to 1.0, got %.2f", totalSplit)
	}

	// Validate minimum sample size
	if config.MinSampleSize < 100 {
		return nil, fmt.Errorf("minimum sample size must be at least 100")
	}

	experiment := &Experiment{
		ID:              config.ID,
		Name:            config.Name,
		Description:     config.Description,
		Status:          StatusDraft,
		TrafficSplit:    config.TrafficSplit,
		Models:          config.Models,
		SuccessMetrics:  config.SuccessMetrics,
		MinSampleSize:   config.MinSampleSize,
		ConfidenceLevel: config.ConfidenceLevel,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	abm.experiments[experiment.ID] = experiment

	abm.logger.Info("A/B test experiment created",
		zap.String("experiment_id", experiment.ID),
		zap.String("name", experiment.Name),
		zap.Int("model_count", len(experiment.Models)))

	return experiment, nil
}

// StartExperiment starts an A/B test experiment
func (abm *ABTestManager) StartExperiment(ctx context.Context, experimentID string) error {
	abm.mu.Lock()
	defer abm.mu.Unlock()

	experiment, exists := abm.experiments[experimentID]
	if !exists {
		return fmt.Errorf("experiment %s not found", experimentID)
	}

	if experiment.Status != StatusDraft {
		return fmt.Errorf("experiment %s is not in draft status", experimentID)
	}

	experiment.Status = StatusRunning
	experiment.StartTime = time.Now()
	experiment.UpdatedAt = time.Now()

	abm.logger.Info("A/B test experiment started",
		zap.String("experiment_id", experimentID),
		zap.Time("start_time", experiment.StartTime))

	return nil
}

// StopExperiment stops an A/B test experiment
func (abm *ABTestManager) StopExperiment(ctx context.Context, experimentID string) error {
	abm.mu.Lock()
	defer abm.mu.Unlock()

	experiment, exists := abm.experiments[experimentID]
	if !exists {
		return fmt.Errorf("experiment %s not found", experimentID)
	}

	if experiment.Status != StatusRunning {
		return fmt.Errorf("experiment %s is not running", experimentID)
	}

	experiment.Status = StatusCompleted
	now := time.Now()
	experiment.EndTime = &now
	experiment.UpdatedAt = now

	abm.logger.Info("A/B test experiment stopped",
		zap.String("experiment_id", experimentID),
		zap.Time("end_time", now))

	return nil
}

// GetExperiment retrieves an experiment by ID
func (abm *ABTestManager) GetExperiment(experimentID string) (*Experiment, error) {
	abm.mu.RLock()
	defer abm.mu.RUnlock()

	experiment, exists := abm.experiments[experimentID]
	if !exists {
		return nil, fmt.Errorf("experiment %s not found", experimentID)
	}

	return experiment, nil
}

// ListExperiments returns all experiments
func (abm *ABTestManager) ListExperiments() []*Experiment {
	abm.mu.RLock()
	defer abm.mu.RUnlock()

	experiments := make([]*Experiment, 0, len(abm.experiments))
	for _, experiment := range abm.experiments {
		experiments = append(experiments, experiment)
	}

	return experiments
}

// SelectModel selects a model for a request based on traffic splitting
func (abm *ABTestManager) SelectModel(ctx context.Context, experimentID string, requestID string) (string, error) {
	abm.mu.RLock()
	defer abm.mu.RUnlock()

	experiment, exists := abm.experiments[experimentID]
	if !exists {
		return "", fmt.Errorf("experiment %s not found", experimentID)
	}

	if experiment.Status != StatusRunning {
		return "", fmt.Errorf("experiment %s is not running", experimentID)
	}

	// Use request ID as seed for consistent assignment
	rand.Seed(hashString(requestID))
	randomValue := rand.Float64()

	cumulativeSplit := 0.0
	for modelID, split := range experiment.TrafficSplit {
		cumulativeSplit += split
		if randomValue <= cumulativeSplit {
			return modelID, nil
		}
	}

	// Fallback to first model (should never happen with valid traffic split)
	for modelID := range experiment.TrafficSplit {
		return modelID, nil
	}

	return "", fmt.Errorf("no models configured for experiment %s", experimentID)
}

// RecordPrediction records a prediction result for an experiment
func (abm *ABTestManager) RecordPrediction(ctx context.Context, experimentID, modelID string, prediction *PredictionRecord) error {
	abm.mu.RLock()
	experiment, exists := abm.experiments[experimentID]
	abm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("experiment %s not found", experimentID)
	}

	if experiment.Status != StatusRunning {
		return fmt.Errorf("experiment %s is not running", experimentID)
	}

	// Record metrics
	err := abm.metrics.RecordPrediction(experimentID, modelID, prediction)
	if err != nil {
		abm.logger.Error("Failed to record prediction",
			zap.String("experiment_id", experimentID),
			zap.String("model_id", modelID),
			zap.Error(err))
		return err
	}

	return nil
}

// GetExperimentResults retrieves results for a completed experiment
func (abm *ABTestManager) GetExperimentResults(ctx context.Context, experimentID string) (*ExperimentResult, error) {
	abm.mu.RLock()
	experiment, exists := abm.experiments[experimentID]
	abm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("experiment %s not found", experimentID)
	}

	if experiment.Status != StatusCompleted {
		return nil, fmt.Errorf("experiment %s is not completed", experimentID)
	}

	// Get metrics for all models
	modelResults := make(map[string]*ModelResult)
	for modelID := range experiment.Models {
		metrics, err := abm.metrics.GetModelMetrics(experimentID, modelID)
		if err != nil {
			abm.logger.Error("Failed to get model metrics",
				zap.String("experiment_id", experimentID),
				zap.String("model_id", modelID),
				zap.Error(err))
			continue
		}

		modelResults[modelID] = &ModelResult{
			ModelID:         modelID,
			RequestCount:    metrics.RequestCount,
			Metrics:         metrics.Metrics,
			AverageLatency:  metrics.AverageLatency,
			ErrorRate:       metrics.ErrorRate,
			ConfidenceScore: metrics.ConfidenceScore,
			Accuracy:        metrics.Accuracy,
			Precision:       metrics.Precision,
			Recall:          metrics.Recall,
			F1Score:         metrics.F1Score,
		}
	}

	// Perform statistical testing
	statisticalTest, err := abm.performStatisticalTest(modelResults, experiment.SuccessMetrics[0])
	if err != nil {
		abm.logger.Error("Failed to perform statistical test",
			zap.String("experiment_id", experimentID),
			zap.Error(err))
	}

	// Determine winner
	winner := abm.determineWinner(modelResults, experiment.SuccessMetrics[0])

	// Calculate confidence
	confidence := 0.0
	if statisticalTest != nil {
		confidence = 1.0 - statisticalTest.PValue
	}

	// Generate recommendation
	recommendation := abm.generateRecommendation(modelResults, statisticalTest, winner)

	duration := time.Since(experiment.StartTime)
	if experiment.EndTime != nil {
		duration = experiment.EndTime.Sub(experiment.StartTime)
	}

	result := &ExperimentResult{
		ExperimentID:    experimentID,
		TotalRequests:   abm.calculateTotalRequests(modelResults),
		Duration:        duration,
		ModelResults:    modelResults,
		StatisticalTest: statisticalTest,
		Winner:          winner,
		Confidence:      confidence,
		IsSignificant:   statisticalTest != nil && statisticalTest.IsSignificant,
		Recommendation:  recommendation,
		CompletedAt:     time.Now(),
	}

	return result, nil
}

// performStatisticalTest performs statistical significance testing
func (abm *ABTestManager) performStatisticalTest(modelResults map[string]*ModelResult, primaryMetric string) (*StatisticalTestResult, error) {
	if len(modelResults) < 2 {
		return nil, fmt.Errorf("need at least 2 models for statistical testing")
	}

	// Get primary metric values for all models
	values := make([]float64, 0, len(modelResults))
	for _, result := range modelResults {
		if value, exists := result.Metrics[primaryMetric]; exists {
			values = append(values, value)
		}
	}

	if len(values) < 2 {
		return nil, fmt.Errorf("insufficient data for statistical testing")
	}

	// Perform t-test (simplified implementation)
	// In a real implementation, you would use a proper statistical library
	pValue := abm.calculateTTest(values)
	effectSize := abm.calculateEffectSize(values)
	confidenceInterval := abm.calculateConfidenceInterval(values)
	power := abm.calculatePower(values, pValue)

	isSignificant := pValue < 0.05 // Standard significance level

	return &StatisticalTestResult{
		TestType:           "t-test",
		PValue:             pValue,
		IsSignificant:      isSignificant,
		EffectSize:         effectSize,
		ConfidenceInterval: confidenceInterval,
		Power:              power,
	}, nil
}

// determineWinner determines the winning model based on primary metric
func (abm *ABTestManager) determineWinner(modelResults map[string]*ModelResult, primaryMetric string) string {
	bestModel := ""
	bestValue := -1.0

	for modelID, result := range modelResults {
		if value, exists := result.Metrics[primaryMetric]; exists {
			if value > bestValue {
				bestValue = value
				bestModel = modelID
			}
		}
	}

	return bestModel
}

// generateRecommendation generates a recommendation based on results
func (abm *ABTestManager) generateRecommendation(modelResults map[string]*ModelResult, statisticalTest *StatisticalTestResult, winner string) string {
	if statisticalTest == nil {
		return "Insufficient data for recommendation"
	}

	if !statisticalTest.IsSignificant {
		return fmt.Sprintf("No significant difference found. Consider running experiment longer or increasing sample size.")
	}

	if statisticalTest.EffectSize < 0.2 {
		return fmt.Sprintf("Winner: %s, but effect size is small. Consider practical significance.", winner)
	}

	return fmt.Sprintf("Winner: %s with significant improvement (p=%.4f, effect size=%.3f)", winner, statisticalTest.PValue, statisticalTest.EffectSize)
}

// Helper functions for statistical calculations
func (abm *ABTestManager) calculateTTest(values []float64) float64 {
	// Simplified t-test calculation
	// In production, use a proper statistical library
	if len(values) < 2 {
		return 1.0
	}

	// Calculate mean
	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	// Calculate standard deviation
	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(values) - 1)
	stdDev := math.Sqrt(variance)

	// Simplified p-value calculation
	// This is a placeholder - use proper statistical library in production
	if stdDev == 0 {
		return 0.0
	}

	// Approximate p-value based on t-statistic
	tStat := math.Abs(values[0]-values[1]) / (stdDev * math.Sqrt(2.0/float64(len(values))))
	pValue := 2.0 * (1.0 - abm.normalCDF(tStat))

	return pValue
}

func (abm *ABTestManager) calculateEffectSize(values []float64) float64 {
	if len(values) < 2 {
		return 0.0
	}

	// Cohen's d calculation
	mean1 := values[0]
	mean2 := values[1]

	// Pooled standard deviation (simplified)
	stdDev := math.Abs(mean1-mean2) / 2.0

	if stdDev == 0 {
		return 0.0
	}

	return math.Abs(mean1-mean2) / stdDev
}

func (abm *ABTestManager) calculateConfidenceInterval(values []float64) [2]float64 {
	if len(values) < 2 {
		return [2]float64{0.0, 0.0}
	}

	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	// Simplified confidence interval (95%)
	margin := 1.96 * math.Sqrt(mean*(1.0-mean)/float64(len(values)))

	lower := mean - margin
	upper := mean + margin

	// Clamp values to [0, 1] range
	if lower < 0 {
		lower = 0
	}
	if upper > 1 {
		upper = 1
	}

	return [2]float64{lower, upper}
}

func (abm *ABTestManager) calculatePower(values []float64, pValue float64) float64 {
	// Simplified power calculation
	// In production, use proper statistical library
	if pValue < 0.05 {
		return 0.8 // Assume 80% power for significant results
	}
	return 0.2 // Assume 20% power for non-significant results
}

func (abm *ABTestManager) normalCDF(x float64) float64 {
	// Simplified normal CDF approximation
	// In production, use proper statistical library
	return 0.5 * (1.0 + math.Erf(x/math.Sqrt2))
}

func (abm *ABTestManager) calculateTotalRequests(modelResults map[string]*ModelResult) int64 {
	total := int64(0)
	for _, result := range modelResults {
		total += result.RequestCount
	}
	return total
}

// GetModelMetrics returns metrics for a specific model in an experiment
func (abm *ABTestManager) GetModelMetrics(experimentID, modelID string) (*ModelMetrics, error) {
	return abm.metrics.GetModelMetrics(experimentID, modelID)
}

// GetExperimentMetrics returns metrics for all models in an experiment
func (abm *ABTestManager) GetExperimentMetrics(experimentID string) (map[string]*ModelMetrics, error) {
	return abm.metrics.GetExperimentMetrics(experimentID)
}

// hashString creates a hash from a string for consistent random assignment
func hashString(s string) int64 {
	hash := int64(0)
	for _, c := range s {
		hash = hash*31 + int64(c)
	}
	return hash
}
