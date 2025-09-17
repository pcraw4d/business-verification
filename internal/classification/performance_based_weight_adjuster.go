package classification

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/shared"
)

// PerformanceBasedWeightAdjuster manages dynamic weight adjustment based on method performance
type PerformanceBasedWeightAdjuster struct {
	// Configuration
	config PerformanceWeightConfig

	// Performance tracking
	performanceTracker *MethodPerformanceTracker
	weightManager      *WeightConfigurationManager

	// A/B Testing
	abTestManager *ABTestManager

	// Historical data
	historicalData *HistoricalPerformanceData

	// Thread safety
	mutex sync.RWMutex

	// Logging
	logger *log.Logger

	// Control
	ctx    context.Context
	cancel context.CancelFunc
}

// PerformanceWeightConfig holds configuration for performance-based weight adjustment
type PerformanceWeightConfig struct {
	// Weight adjustment settings
	Enabled                 bool          `json:"enabled"`
	AdjustmentInterval      time.Duration `json:"adjustment_interval"`
	MinWeight               float64       `json:"min_weight"`
	MaxWeight               float64       `json:"max_weight"`
	WeightAdjustmentStep    float64       `json:"weight_adjustment_step"`
	PerformanceWindow       time.Duration `json:"performance_window"`
	MinSamplesForAdjustment int           `json:"min_samples_for_adjustment"`
	AccuracyThreshold       float64       `json:"accuracy_threshold"`
	PerformanceDecayFactor  float64       `json:"performance_decay_factor"`

	// A/B Testing settings
	ABTestingEnabled        bool          `json:"ab_testing_enabled"`
	ABTestDuration          time.Duration `json:"ab_test_duration"`
	ABTestTrafficSplit      float64       `json:"ab_test_traffic_split"`
	ABTestMinSampleSize     int           `json:"ab_test_min_sample_size"`
	ABTestSignificanceLevel float64       `json:"ab_test_significance_level"`

	// Learning settings
	LearningRate            float64 `json:"learning_rate"`
	AdaptiveLearningEnabled bool    `json:"adaptive_learning_enabled"`
	WeightSmoothingFactor   float64 `json:"weight_smoothing_factor"`
	PerformanceWeightFactor float64 `json:"performance_weight_factor"`
}

// MethodPerformanceTracker tracks performance metrics for each classification method
type MethodPerformanceTracker struct {
	// Performance data by method
	performanceData map[string]*MethodPerformanceData
	mutex           sync.RWMutex

	// Configuration
	config PerformanceWeightConfig
	logger *log.Logger
}

// MethodPerformanceData holds performance data for a single method
type MethodPerformanceData struct {
	MethodName         string              `json:"method_name"`
	TotalRequests      int64               `json:"total_requests"`
	SuccessfulRequests int64               `json:"successful_requests"`
	FailedRequests     int64               `json:"failed_requests"`
	AverageAccuracy    float64             `json:"average_accuracy"`
	AverageLatency     time.Duration       `json:"average_latency"`
	LastAccuracy       float64             `json:"last_accuracy"`
	LastLatency        time.Duration       `json:"last_latency"`
	LastUpdated        time.Time           `json:"last_updated"`
	AccuracyHistory    []AccuracyDataPoint `json:"accuracy_history"`
	LatencyHistory     []LatencyDataPoint  `json:"latency_history"`
	CurrentWeight      float64             `json:"current_weight"`
	WeightHistory      []WeightDataPoint   `json:"weight_history"`
}

// AccuracyDataPoint represents a single accuracy measurement
type AccuracyDataPoint struct {
	Timestamp  time.Time `json:"timestamp"`
	Accuracy   float64   `json:"accuracy"`
	SampleSize int       `json:"sample_size"`
}

// LatencyDataPoint represents a single latency measurement
type LatencyDataPoint struct {
	Timestamp time.Time     `json:"timestamp"`
	Latency   time.Duration `json:"latency"`
}

// WeightDataPoint represents a single weight adjustment
type WeightDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Weight    float64   `json:"weight"`
	Reason    string    `json:"reason"`
}

// HistoricalPerformanceData stores historical performance data
type HistoricalPerformanceData struct {
	DataFile string
	Data     map[string]*MethodPerformanceData
	mutex    sync.RWMutex
}

// ABTestManager manages A/B testing for weight optimization
type ABTestManager struct {
	// Active tests
	activeTests map[string]*ABTest
	mutex       sync.RWMutex

	// Configuration
	config PerformanceWeightConfig
	logger *log.Logger
}

// ABTest represents an A/B test for weight optimization
type ABTest struct {
	TestID          string    `json:"test_id"`
	MethodName      string    `json:"method_name"`
	ControlWeight   float64   `json:"control_weight"`
	TreatmentWeight float64   `json:"treatment_weight"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	TrafficSplit    float64   `json:"traffic_split"`
	Status          string    `json:"status"` // "active", "completed", "cancelled"

	// Results
	ControlResults    *ABTestResults `json:"control_results"`
	TreatmentResults  *ABTestResults `json:"treatment_results"`
	SignificanceLevel float64        `json:"significance_level"`
	IsSignificant     bool           `json:"is_significant"`
	Winner            string         `json:"winner"` // "control", "treatment", "inconclusive"
}

// ABTestResults holds results for an A/B test
type ABTestResults struct {
	SampleSize      int           `json:"sample_size"`
	AverageAccuracy float64       `json:"average_accuracy"`
	AverageLatency  time.Duration `json:"average_latency"`
	SuccessRate     float64       `json:"success_rate"`
	ErrorRate       float64       `json:"error_rate"`
}

// NewPerformanceBasedWeightAdjuster creates a new performance-based weight adjuster
func NewPerformanceBasedWeightAdjuster(
	config PerformanceWeightConfig,
	performanceTracker *MethodPerformanceTracker,
	weightManager *WeightConfigurationManager,
	logger *log.Logger,
) *PerformanceBasedWeightAdjuster {
	if logger == nil {
		logger = log.Default()
	}

	ctx, cancel := context.WithCancel(context.Background())

	adjuster := &PerformanceBasedWeightAdjuster{
		config:             config,
		performanceTracker: performanceTracker,
		weightManager:      weightManager,
		abTestManager:      NewABTestManager(config, logger),
		historicalData:     NewHistoricalPerformanceData("data/performance_history.json", logger),
		logger:             logger,
		ctx:                ctx,
		cancel:             cancel,
	}

	// Set default configuration if not provided
	adjuster.setDefaultConfig()

	return adjuster
}

// setDefaultConfig sets default configuration values
func (pbwa *PerformanceBasedWeightAdjuster) setDefaultConfig() {
	if pbwa.config.AdjustmentInterval == 0 {
		pbwa.config.AdjustmentInterval = 1 * time.Hour
	}
	if pbwa.config.MinWeight == 0 {
		pbwa.config.MinWeight = 0.05
	}
	if pbwa.config.MaxWeight == 0 {
		pbwa.config.MaxWeight = 0.8
	}
	if pbwa.config.WeightAdjustmentStep == 0 {
		pbwa.config.WeightAdjustmentStep = 0.05
	}
	if pbwa.config.PerformanceWindow == 0 {
		pbwa.config.PerformanceWindow = 24 * time.Hour
	}
	if pbwa.config.MinSamplesForAdjustment == 0 {
		pbwa.config.MinSamplesForAdjustment = 100
	}
	if pbwa.config.AccuracyThreshold == 0 {
		pbwa.config.AccuracyThreshold = 0.7
	}
	if pbwa.config.PerformanceDecayFactor == 0 {
		pbwa.config.PerformanceDecayFactor = 0.95
	}
	if pbwa.config.LearningRate == 0 {
		pbwa.config.LearningRate = 0.1
	}
	if pbwa.config.WeightSmoothingFactor == 0 {
		pbwa.config.WeightSmoothingFactor = 0.1
	}
	if pbwa.config.PerformanceWeightFactor == 0 {
		pbwa.config.PerformanceWeightFactor = 0.7
	}
}

// Start begins the performance-based weight adjustment process
func (pbwa *PerformanceBasedWeightAdjuster) Start() error {
	if !pbwa.config.Enabled {
		pbwa.logger.Printf("📊 Performance-based weight adjustment is disabled")
		return nil
	}

	pbwa.logger.Printf("🚀 Starting performance-based weight adjustment system")
	pbwa.logger.Printf("📊 Configuration: interval=%v, min_weight=%.3f, max_weight=%.3f, step=%.3f",
		pbwa.config.AdjustmentInterval, pbwa.config.MinWeight, pbwa.config.MaxWeight, pbwa.config.WeightAdjustmentStep)

	// Start the adjustment loop
	go pbwa.adjustmentLoop()

	// Start A/B testing if enabled
	if pbwa.config.ABTestingEnabled {
		go pbwa.abTestManager.manageABTests()
	}

	return nil
}

// Stop stops the performance-based weight adjustment process
func (pbwa *PerformanceBasedWeightAdjuster) Stop() {
	pbwa.logger.Printf("🛑 Stopping performance-based weight adjustment system")
	pbwa.cancel()
}

// adjustmentLoop runs the main weight adjustment loop
func (pbwa *PerformanceBasedWeightAdjuster) adjustmentLoop() {
	ticker := time.NewTicker(pbwa.config.AdjustmentInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pbwa.ctx.Done():
			pbwa.logger.Printf("📊 Weight adjustment loop stopped")
			return
		case <-ticker.C:
			if err := pbwa.performWeightAdjustment(); err != nil {
				pbwa.logger.Printf("❌ Error during weight adjustment: %v", err)
			}
		}
	}
}

// performWeightAdjustment performs the actual weight adjustment based on performance
func (pbwa *PerformanceBasedWeightAdjuster) performWeightAdjustment() error {
	pbwa.mutex.Lock()
	defer pbwa.mutex.Unlock()

	pbwa.logger.Printf("📊 Performing weight adjustment based on performance data")

	// Get current performance data
	performanceData := pbwa.performanceTracker.GetAllPerformanceData()
	if len(performanceData) == 0 {
		pbwa.logger.Printf("⚠️ No performance data available for weight adjustment")
		return nil
	}

	// Calculate new weights based on performance
	newWeights := pbwa.calculateOptimalWeights(performanceData)
	if len(newWeights) == 0 {
		pbwa.logger.Printf("⚠️ No weight adjustments calculated")
		return nil
	}

	// Apply weight adjustments
	adjustments := make(map[string]WeightAdjustment)
	for methodName, newWeight := range newWeights {
		currentWeight := pbwa.getCurrentWeight(methodName)
		adjustment := WeightAdjustment{
			MethodName: methodName,
			OldWeight:  currentWeight,
			NewWeight:  newWeight,
			Adjustment: newWeight - currentWeight,
			Reason:     "performance_based_adjustment",
			Timestamp:  time.Now(),
		}
		adjustments[methodName] = adjustment

		// Apply the weight adjustment
		if err := pbwa.applyWeightAdjustment(adjustment); err != nil {
			pbwa.logger.Printf("❌ Failed to apply weight adjustment for method '%s': %v", methodName, err)
			continue
		}

		pbwa.logger.Printf("✅ Adjusted weight for method '%s': %.3f → %.3f (Δ%.3f)",
			methodName, currentWeight, newWeight, newWeight-currentWeight)
	}

	// Save historical data
	if err := pbwa.historicalData.SaveWeightAdjustments(adjustments); err != nil {
		pbwa.logger.Printf("⚠️ Failed to save weight adjustment history: %v", err)
	}

	// Log summary
	pbwa.logger.Printf("📊 Weight adjustment completed: %d methods adjusted", len(adjustments))

	return nil
}

// WeightAdjustment represents a single weight adjustment
type WeightAdjustment struct {
	MethodName string    `json:"method_name"`
	OldWeight  float64   `json:"old_weight"`
	NewWeight  float64   `json:"new_weight"`
	Adjustment float64   `json:"adjustment"`
	Reason     string    `json:"reason"`
	Timestamp  time.Time `json:"timestamp"`
}

// calculateOptimalWeights calculates optimal weights based on performance data
func (pbwa *PerformanceBasedWeightAdjuster) calculateOptimalWeights(
	performanceData map[string]*MethodPerformanceData,
) map[string]float64 {
	optimalWeights := make(map[string]float64)

	// Calculate performance scores for each method
	performanceScores := make(map[string]float64)
	totalScore := 0.0

	for methodName, data := range performanceData {
		// Skip methods with insufficient data
		if data.TotalRequests < int64(pbwa.config.MinSamplesForAdjustment) {
			pbwa.logger.Printf("⚠️ Method '%s' has insufficient samples (%d < %d), skipping",
				methodName, data.TotalRequests, pbwa.config.MinSamplesForAdjustment)
			continue
		}

		// Calculate performance score
		score := pbwa.calculatePerformanceScore(data)
		performanceScores[methodName] = score
		totalScore += score

		pbwa.logger.Printf("📊 Method '%s' performance score: %.3f (accuracy=%.3f, latency=%v, requests=%d)",
			methodName, score, data.AverageAccuracy, data.AverageLatency, data.TotalRequests)
	}

	if totalScore == 0 {
		pbwa.logger.Printf("⚠️ No valid performance scores calculated")
		return optimalWeights
	}

	// Calculate optimal weights based on performance scores
	for methodName, score := range performanceScores {
		// Calculate base weight from performance score
		baseWeight := score / totalScore

		// Apply performance-based adjustments
		adjustedWeight := pbwa.applyPerformanceAdjustments(methodName, baseWeight, performanceData[methodName])

		// Ensure weight is within bounds
		finalWeight := pbwa.clampWeight(adjustedWeight)

		optimalWeights[methodName] = finalWeight
	}

	// Normalize weights to ensure they sum to 1.0
	optimalWeights = pbwa.normalizeWeights(optimalWeights)

	return optimalWeights
}

// calculatePerformanceScore calculates a performance score for a method
func (pbwa *PerformanceBasedWeightAdjuster) calculatePerformanceScore(
	data *MethodPerformanceData,
) float64 {
	// Base score from accuracy
	accuracyScore := data.AverageAccuracy

	// Latency penalty (lower latency = higher score)
	latencyPenalty := pbwa.calculateLatencyPenalty(data.AverageLatency)

	// Reliability bonus (higher success rate = higher score)
	reliabilityBonus := float64(data.SuccessfulRequests) / float64(data.TotalRequests)

	// Sample size confidence (more samples = more confidence)
	sampleConfidence := pbwa.calculateSampleConfidence(data.TotalRequests)

	// Calculate weighted performance score
	performanceScore := (accuracyScore * pbwa.config.PerformanceWeightFactor) +
		(reliabilityBonus * 0.2) +
		(sampleConfidence * 0.1) -
		(latencyPenalty * 0.1)

	// Ensure score is non-negative
	if performanceScore < 0 {
		performanceScore = 0
	}

	return performanceScore
}

// calculateLatencyPenalty calculates penalty for high latency
func (pbwa *PerformanceBasedWeightAdjuster) calculateLatencyPenalty(latency time.Duration) float64 {
	// Penalty increases with latency
	// Target latency: 100ms, penalty starts at 500ms
	targetLatency := 100 * time.Millisecond
	penaltyThreshold := 500 * time.Millisecond

	if latency <= targetLatency {
		return 0.0
	}

	if latency >= penaltyThreshold {
		return 1.0 // Maximum penalty
	}

	// Linear penalty between target and threshold
	penalty := float64(latency-targetLatency) / float64(penaltyThreshold-targetLatency)
	return penalty
}

// calculateSampleConfidence calculates confidence based on sample size
func (pbwa *PerformanceBasedWeightAdjuster) calculateSampleConfidence(sampleSize int64) float64 {
	// Confidence increases with sample size, capped at 1.0
	// Target: 1000 samples for full confidence
	targetSamples := int64(1000)

	if sampleSize >= targetSamples {
		return 1.0
	}

	return float64(sampleSize) / float64(targetSamples)
}

// applyPerformanceAdjustments applies additional adjustments based on performance trends
func (pbwa *PerformanceBasedWeightAdjuster) applyPerformanceAdjustments(
	methodName string,
	baseWeight float64,
	data *MethodPerformanceData,
) float64 {
	adjustedWeight := baseWeight

	// Trend analysis
	if len(data.AccuracyHistory) >= 2 {
		trend := pbwa.calculateAccuracyTrend(data.AccuracyHistory)

		// Boost weight for improving methods
		if trend > 0.05 { // 5% improvement trend
			adjustedWeight *= 1.1
			pbwa.logger.Printf("📈 Method '%s' showing improvement trend (%.3f), boosting weight", methodName, trend)
		}

		// Reduce weight for declining methods
		if trend < -0.05 { // 5% decline trend
			adjustedWeight *= 0.9
			pbwa.logger.Printf("📉 Method '%s' showing decline trend (%.3f), reducing weight", methodName, trend)
		}
	}

	// Stability bonus (consistent performance)
	stability := pbwa.calculatePerformanceStability(data.AccuracyHistory)
	if stability > 0.8 { // High stability
		adjustedWeight *= 1.05
		pbwa.logger.Printf("🎯 Method '%s' showing high stability (%.3f), slight weight boost", methodName, stability)
	}

	return adjustedWeight
}

// calculateAccuracyTrend calculates the trend in accuracy over time
func (pbwa *PerformanceBasedWeightAdjuster) calculateAccuracyTrend(history []AccuracyDataPoint) float64 {
	if len(history) < 2 {
		return 0.0
	}

	// Simple linear regression slope
	n := len(history)
	var sumX, sumY, sumXY, sumXX float64

	for i, point := range history {
		x := float64(i)
		y := point.Accuracy
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	// Calculate slope
	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumXX - sumX*sumX)
	return slope
}

// calculatePerformanceStability calculates how stable the performance is
func (pbwa *PerformanceBasedWeightAdjuster) calculatePerformanceStability(history []AccuracyDataPoint) float64 {
	if len(history) < 2 {
		return 1.0
	}

	// Calculate coefficient of variation (lower = more stable)
	var sum, sumSquares float64
	for _, point := range history {
		sum += point.Accuracy
		sumSquares += point.Accuracy * point.Accuracy
	}

	mean := sum / float64(len(history))
	variance := (sumSquares / float64(len(history))) - (mean * mean)
	stdDev := math.Sqrt(variance)

	if mean == 0 {
		return 1.0
	}

	coefficientOfVariation := stdDev / mean
	stability := 1.0 - coefficientOfVariation

	// Ensure stability is between 0 and 1
	if stability < 0 {
		stability = 0
	}
	if stability > 1 {
		stability = 1
	}

	return stability
}

// clampWeight ensures weight is within configured bounds
func (pbwa *PerformanceBasedWeightAdjuster) clampWeight(weight float64) float64 {
	if weight < pbwa.config.MinWeight {
		return pbwa.config.MinWeight
	}
	if weight > pbwa.config.MaxWeight {
		return pbwa.config.MaxWeight
	}
	return weight
}

// normalizeWeights ensures weights sum to 1.0
func (pbwa *PerformanceBasedWeightAdjuster) normalizeWeights(weights map[string]float64) map[string]float64 {
	var total float64
	for _, weight := range weights {
		total += weight
	}

	if total == 0 {
		return weights
	}

	normalized := make(map[string]float64)
	for methodName, weight := range weights {
		normalized[methodName] = weight / total
	}

	return normalized
}

// getCurrentWeight gets the current weight for a method
func (pbwa *PerformanceBasedWeightAdjuster) getCurrentWeight(methodName string) float64 {
	if pbwa.weightManager == nil {
		return 0.5 // Default weight
	}

	// Get weight from weight manager
	weight, err := pbwa.weightManager.GetMethodWeight(methodName)
	if err != nil {
		return 0.5 // Default weight
	}

	return weight
}

// applyWeightAdjustment applies a weight adjustment to the system
func (pbwa *PerformanceBasedWeightAdjuster) applyWeightAdjustment(adjustment WeightAdjustment) error {
	if pbwa.weightManager == nil {
		return fmt.Errorf("weight manager not available")
	}

	// Update weight in the weight manager
	if err := pbwa.weightManager.SetMethodWeight(adjustment.MethodName, adjustment.NewWeight); err != nil {
		return fmt.Errorf("failed to set method weight: %w", err)
	}

	// Update performance tracker
	if pbwa.performanceTracker != nil {
		pbwa.performanceTracker.UpdateMethodWeight(adjustment.MethodName, adjustment.NewWeight)
	}

	// Save configuration
	if err := pbwa.weightManager.SaveConfiguration(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// RecordClassificationResult records a classification result for performance tracking
func (pbwa *PerformanceBasedWeightAdjuster) RecordClassificationResult(
	methodName string,
	result *shared.ClassificationMethodResult,
) {
	if pbwa.performanceTracker == nil {
		return
	}

	pbwa.performanceTracker.RecordResult(methodName, result)
}

// GetPerformanceSummary returns a summary of current performance data
func (pbwa *PerformanceBasedWeightAdjuster) GetPerformanceSummary() map[string]interface{} {
	pbwa.mutex.RLock()
	defer pbwa.mutex.RUnlock()

	summary := make(map[string]interface{})

	if pbwa.performanceTracker != nil {
		summary["performance_data"] = pbwa.performanceTracker.GetAllPerformanceData()
	}

	if pbwa.abTestManager != nil {
		summary["active_ab_tests"] = pbwa.abTestManager.GetActiveTests()
	}

	summary["config"] = pbwa.config
	summary["enabled"] = pbwa.config.Enabled

	return summary
}
