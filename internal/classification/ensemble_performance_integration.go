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

// EnsemblePerformanceIntegration integrates performance-based weight adjustment with the ensemble system
type EnsemblePerformanceIntegration struct {
	// Core components
	weightAdjuster     *PerformanceBasedWeightAdjuster
	performanceTracker *MethodPerformanceTracker
	abTestManager      *ABTestManager
	weightManager      *WeightConfigurationManager

	// Ensemble system
	multiMethodClassifier *MultiMethodClassifier

	// Configuration
	config EnsemblePerformanceConfig

	// Thread safety
	mutex sync.RWMutex

	// Logging
	logger *log.Logger

	// Control
	ctx    context.Context
	cancel context.CancelFunc
}

// EnsemblePerformanceConfig holds configuration for ensemble performance integration
type EnsemblePerformanceConfig struct {
	// Performance tracking
	PerformanceTrackingEnabled bool          `json:"performance_tracking_enabled"`
	PerformanceUpdateInterval  time.Duration `json:"performance_update_interval"`

	// Weight adjustment
	WeightAdjustmentEnabled  bool          `json:"weight_adjustment_enabled"`
	WeightAdjustmentInterval time.Duration `json:"weight_adjustment_interval"`

	// A/B Testing
	ABTestingEnabled bool          `json:"ab_testing_enabled"`
	ABTestAutoStart  bool          `json:"ab_test_auto_start"`
	ABTestDuration   time.Duration `json:"ab_test_duration"`

	// Learning
	LearningEnabled         bool    `json:"learning_enabled"`
	LearningRate            float64 `json:"learning_rate"`
	AdaptiveLearningEnabled bool    `json:"adaptive_learning_enabled"`

	// Performance thresholds
	MinAccuracyForBoost   float64       `json:"min_accuracy_for_boost"`
	MaxLatencyForPenalty  time.Duration `json:"max_latency_for_penalty"`
	MinSamplesForLearning int           `json:"min_samples_for_learning"`
}

// NewEnsemblePerformanceIntegration creates a new ensemble performance integration
func NewEnsemblePerformanceIntegration(
	multiMethodClassifier *MultiMethodClassifier,
	weightManager *WeightConfigurationManager,
	config EnsemblePerformanceConfig,
	logger *log.Logger,
) *EnsemblePerformanceIntegration {
	if logger == nil {
		logger = log.Default()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Set default configuration
	if config.PerformanceUpdateInterval == 0 {
		config.PerformanceUpdateInterval = 30 * time.Second
	}
	if config.WeightAdjustmentInterval == 0 {
		config.WeightAdjustmentInterval = 1 * time.Hour
	}
	if config.ABTestDuration == 0 {
		config.ABTestDuration = 24 * time.Hour
	}
	if config.LearningRate == 0 {
		config.LearningRate = 0.1
	}
	if config.MinAccuracyForBoost == 0 {
		config.MinAccuracyForBoost = 0.8
	}
	if config.MaxLatencyForPenalty == 0 {
		config.MaxLatencyForPenalty = 2 * time.Second
	}
	if config.MinSamplesForLearning == 0 {
		config.MinSamplesForLearning = 100
	}

	// Create performance tracker
	performanceConfig := PerformanceWeightConfig{
		Enabled:                 config.PerformanceTrackingEnabled,
		AdjustmentInterval:      config.WeightAdjustmentInterval,
		MinWeight:               0.05,
		MaxWeight:               0.8,
		WeightAdjustmentStep:    0.05,
		PerformanceWindow:       24 * time.Hour,
		MinSamplesForAdjustment: config.MinSamplesForLearning,
		AccuracyThreshold:       config.MinAccuracyForBoost,
		PerformanceDecayFactor:  0.95,
		ABTestingEnabled:        config.ABTestingEnabled,
		ABTestDuration:          config.ABTestDuration,
		ABTestTrafficSplit:      0.5,
		ABTestMinSampleSize:     100,
		ABTestSignificanceLevel: 0.05,
		LearningRate:            config.LearningRate,
		AdaptiveLearningEnabled: config.AdaptiveLearningEnabled,
		WeightSmoothingFactor:   0.1,
		PerformanceWeightFactor: 0.7,
	}

	performanceTracker := NewMethodPerformanceTracker(performanceConfig, logger)
	abTestManager := NewABTestManager(performanceConfig, logger)
	weightAdjuster := NewPerformanceBasedWeightAdjuster(performanceConfig, performanceTracker, weightManager, logger)

	integration := &EnsemblePerformanceIntegration{
		weightAdjuster:        weightAdjuster,
		performanceTracker:    performanceTracker,
		abTestManager:         abTestManager,
		weightManager:         weightManager,
		multiMethodClassifier: multiMethodClassifier,
		config:                config,
		logger:                logger,
		ctx:                   ctx,
		cancel:                cancel,
	}

	return integration
}

// Start starts the ensemble performance integration system
func (epi *EnsemblePerformanceIntegration) Start() error {
	epi.logger.Printf("🚀 Starting ensemble performance integration system")

	// Start performance tracking
	if epi.config.PerformanceTrackingEnabled {
		epi.logger.Printf("📊 Starting performance tracking")
		go epi.performanceTrackingLoop()
	}

	// Start weight adjustment
	if epi.config.WeightAdjustmentEnabled {
		epi.logger.Printf("⚖️ Starting weight adjustment system")
		if err := epi.weightAdjuster.Start(); err != nil {
			return fmt.Errorf("failed to start weight adjuster: %w", err)
		}
	}

	// Start A/B testing
	if epi.config.ABTestingEnabled {
		epi.logger.Printf("🧪 Starting A/B testing system")
		go epi.abTestManager.manageABTests()
	}

	// Start learning system
	if epi.config.LearningEnabled {
		epi.logger.Printf("🧠 Starting learning system")
		go epi.learningLoop()
	}

	epi.logger.Printf("✅ Ensemble performance integration system started successfully")
	return nil
}

// Stop stops the ensemble performance integration system
func (epi *EnsemblePerformanceIntegration) Stop() {
	epi.logger.Printf("🛑 Stopping ensemble performance integration system")
	epi.cancel()

	if epi.weightAdjuster != nil {
		epi.weightAdjuster.Stop()
	}
}

// ClassifyWithPerformanceTracking performs classification with performance tracking
func (epi *EnsemblePerformanceIntegration) ClassifyWithPerformanceTracking(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*MultiMethodClassificationResult, error) {
	startTime := time.Now()

	// Perform classification using the multi-method classifier
	result, err := epi.multiMethodClassifier.ClassifyWithMultipleMethods(ctx, businessName, description, websiteURL)
	if err != nil {
		epi.logger.Printf("❌ Classification failed: %v", err)
		return nil, err
	}

	processingTime := time.Since(startTime)

	// Track performance if enabled
	if epi.config.PerformanceTrackingEnabled {
		epi.trackClassificationPerformance(result, processingTime, err)
	}

	return result, nil
}

// trackClassificationPerformance tracks the performance of classification results
func (epi *EnsemblePerformanceIntegration) trackClassificationPerformance(
	result *MultiMethodClassificationResult,
	processingTime time.Duration,
	err error,
) {
	// Extract method results from ensemble result
	if result.MethodResults == nil {
		return
	}

	// Record performance for each method
	for _, methodResult := range result.MethodResults {
		epi.performanceTracker.RecordResult(methodResult.MethodType, &methodResult)
	}

	epi.logger.Printf("📊 Tracked performance for ensemble classification: processing_time=%v, confidence=%.3f",
		processingTime, result.PrimaryClassification.ConfidenceScore)
}

// performanceTrackingLoop runs the performance tracking loop
func (epi *EnsemblePerformanceIntegration) performanceTrackingLoop() {
	ticker := time.NewTicker(epi.config.PerformanceUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-epi.ctx.Done():
			epi.logger.Printf("📊 Performance tracking loop stopped")
			return
		case <-ticker.C:
			epi.updatePerformanceMetrics()
		}
	}
}

// updatePerformanceMetrics updates performance metrics
func (epi *EnsemblePerformanceIntegration) updatePerformanceMetrics() {
	epi.mutex.Lock()
	defer epi.mutex.Unlock()

	// Get performance summary
	summary := epi.performanceTracker.GetPerformanceSummary()

	// Log performance metrics
	epi.logger.Printf("📊 Performance metrics update:")
	epi.logger.Printf("   Total methods: %v", summary["total_methods"])
	epi.logger.Printf("   Total requests: %v", summary["total_requests"])
	epi.logger.Printf("   Overall success rate: %.3f", summary["overall_success_rate"])
	epi.logger.Printf("   Overall error rate: %.3f", summary["overall_error_rate"])

	// Check for methods that need attention
	if methods, ok := summary["methods"].(map[string]interface{}); ok {
		for methodName, methodData := range methods {
			if data, ok := methodData.(map[string]interface{}); ok {
				successRate := data["success_rate"].(float64)
				errorRate := data["error_rate"].(float64)

				// Alert on high error rates
				if errorRate > 0.1 { // 10% error rate
					epi.logger.Printf("⚠️ High error rate for method '%s': %.3f", methodName, errorRate)
				}

				// Alert on low success rates
				if successRate < 0.8 { // 80% success rate
					epi.logger.Printf("⚠️ Low success rate for method '%s': %.3f", methodName, successRate)
				}
			}
		}
	}
}

// learningLoop runs the learning loop for adaptive weight adjustment
func (epi *EnsemblePerformanceIntegration) learningLoop() {
	ticker := time.NewTicker(1 * time.Hour) // Learn every hour
	defer ticker.Stop()

	for {
		select {
		case <-epi.ctx.Done():
			epi.logger.Printf("🧠 Learning loop stopped")
			return
		case <-ticker.C:
			epi.performLearning()
		}
	}
}

// performLearning performs learning-based weight adjustments
func (epi *EnsemblePerformanceIntegration) performLearning() {
	epi.mutex.Lock()
	defer epi.mutex.Unlock()

	epi.logger.Printf("🧠 Performing learning-based weight adjustments")

	// Get current performance data
	performanceData := epi.performanceTracker.GetAllPerformanceData()
	if len(performanceData) == 0 {
		epi.logger.Printf("⚠️ No performance data available for learning")
		return
	}

	// Identify methods that need weight adjustments
	adjustments := make(map[string]float64)

	for methodName, data := range performanceData {
		// Skip methods with insufficient data
		if data.TotalRequests < int64(epi.config.MinSamplesForLearning) {
			continue
		}

		currentWeight := epi.getCurrentMethodWeight(methodName)
		newWeight := currentWeight

		// Boost weight for high-performing methods
		if data.AverageAccuracy >= epi.config.MinAccuracyForBoost {
			boost := epi.config.LearningRate * 0.1 // Small boost
			newWeight = currentWeight + boost
			epi.logger.Printf("📈 Boosting weight for high-performing method '%s': %.3f → %.3f",
				methodName, currentWeight, newWeight)
		}

		// Reduce weight for low-performing methods
		if data.AverageAccuracy < epi.config.MinAccuracyForBoost*0.8 {
			penalty := epi.config.LearningRate * 0.1 // Small penalty
			newWeight = currentWeight - penalty
			epi.logger.Printf("📉 Reducing weight for low-performing method '%s': %.3f → %.3f",
				methodName, currentWeight, newWeight)
		}

		// Penalize high-latency methods
		if data.AverageLatency > epi.config.MaxLatencyForPenalty {
			penalty := epi.config.LearningRate * 0.05 // Small penalty
			newWeight = newWeight - penalty
			epi.logger.Printf("⏱️ Penalizing high-latency method '%s': latency=%v, weight=%.3f",
				methodName, data.AverageLatency, newWeight)
		}

		// Ensure weight is within bounds
		if newWeight < 0.05 {
			newWeight = 0.05
		}
		if newWeight > 0.8 {
			newWeight = 0.8
		}

		// Only apply if there's a meaningful change
		if math.Abs(newWeight-currentWeight) > 0.01 {
			adjustments[methodName] = newWeight
		}
	}

	// Apply adjustments
	if len(adjustments) > 0 {
		epi.applyWeightAdjustments(adjustments)
		epi.logger.Printf("🧠 Applied %d learning-based weight adjustments", len(adjustments))
	} else {
		epi.logger.Printf("🧠 No weight adjustments needed based on learning")
	}
}

// applyWeightAdjustments applies weight adjustments to the system
func (epi *EnsemblePerformanceIntegration) applyWeightAdjustments(adjustments map[string]float64) {
	for methodName, newWeight := range adjustments {
		// Update weight in weight manager
		if err := epi.weightManager.SetMethodWeight(methodName, newWeight); err != nil {
			epi.logger.Printf("❌ Failed to set weight for method '%s': %v", methodName, err)
			continue
		}

		// Update performance tracker
		epi.performanceTracker.UpdateMethodWeight(methodName, newWeight)

		epi.logger.Printf("✅ Updated weight for method '%s': %.3f", methodName, newWeight)
	}

	// Save configuration
	if err := epi.weightManager.SaveConfiguration(); err != nil {
		epi.logger.Printf("❌ Failed to save weight configuration: %v", err)
	}
}

// getCurrentMethodWeight gets the current weight for a method
func (epi *EnsemblePerformanceIntegration) getCurrentMethodWeight(methodName string) float64 {
	weight, err := epi.weightManager.GetMethodWeight(methodName)
	if err != nil {
		return 0.5 // Default weight
	}
	return weight
}

// StartABTest starts an A/B test for a method
func (epi *EnsemblePerformanceIntegration) StartABTest(
	methodName string,
	controlWeight float64,
	treatmentWeight float64,
) (*ABTest, error) {
	if !epi.config.ABTestingEnabled {
		return nil, fmt.Errorf("A/B testing is disabled")
	}

	return epi.abTestManager.StartABTest(methodName, controlWeight, treatmentWeight, epi.config.ABTestDuration)
}

// GetPerformanceSummary returns a comprehensive performance summary
func (epi *EnsemblePerformanceIntegration) GetPerformanceSummary() map[string]interface{} {
	epi.mutex.RLock()
	defer epi.mutex.RUnlock()

	summary := make(map[string]interface{})

	// Performance tracking summary
	if epi.performanceTracker != nil {
		summary["performance_tracking"] = epi.performanceTracker.GetPerformanceSummary()
	}

	// A/B testing summary
	if epi.abTestManager != nil {
		summary["ab_testing"] = epi.abTestManager.GetTestSummary()
	}

	// Weight adjustment summary
	if epi.weightAdjuster != nil {
		summary["weight_adjustment"] = epi.weightAdjuster.GetPerformanceSummary()
	}

	// Configuration
	summary["config"] = epi.config

	// System status
	summary["status"] = map[string]interface{}{
		"performance_tracking_enabled": epi.config.PerformanceTrackingEnabled,
		"weight_adjustment_enabled":    epi.config.WeightAdjustmentEnabled,
		"ab_testing_enabled":           epi.config.ABTestingEnabled,
		"learning_enabled":             epi.config.LearningEnabled,
	}

	return summary
}

// GetMethodPerformanceData returns performance data for a specific method
func (epi *EnsemblePerformanceIntegration) GetMethodPerformanceData(methodName string) (*MethodPerformanceData, bool) {
	return epi.performanceTracker.GetMethodPerformanceData(methodName)
}

// GetActiveABTests returns all active A/B tests
func (epi *EnsemblePerformanceIntegration) GetActiveABTests() map[string]*ABTest {
	return epi.abTestManager.GetActiveTests()
}

// RecordUserFeedback records user feedback for learning
func (epi *EnsemblePerformanceIntegration) RecordUserFeedback(
	methodName string,
	classification *shared.IndustryClassification,
	userRating float64, // 0.0 to 1.0
	feedback string,
) {
	epi.mutex.Lock()
	defer epi.mutex.Unlock()

	// Create a mock classification result for tracking
	result := &shared.ClassificationMethodResult{
		MethodType:     methodName,
		Success:        true,
		Result:         classification,
		Confidence:     userRating, // Use user rating as confidence
		ProcessingTime: 0,          // Not applicable for user feedback
		Error:          "",
	}

	// Record the feedback
	epi.performanceTracker.RecordResult(methodName, result)

	epi.logger.Printf("📝 Recorded user feedback for method '%s': rating=%.3f, feedback='%s'",
		methodName, userRating, feedback)
}
