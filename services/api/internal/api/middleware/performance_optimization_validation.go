package middleware

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// OptimizationValidationConfig configures the performance optimization validator
type OptimizationValidationConfig struct {
	// Validation window settings
	ValidationWindow time.Duration `json:"validation_window"`
	MinSampleSize    int           `json:"min_sample_size"`
	MaxValidationAge time.Duration `json:"max_validation_age"`

	// Improvement thresholds (percentage improvements required)
	MinResponseTimeImprovement float64 `json:"min_response_time_improvement"`
	MinThroughputImprovement   float64 `json:"min_throughput_improvement"`
	MaxErrorRateIncrease       float64 `json:"max_error_rate_increase"`
	MinResourceEfficiency      float64 `json:"min_resource_efficiency"`

	// Validation criteria
	StatisticalSignificance   float64       `json:"statistical_significance"`
	SustainabilityPeriod      time.Duration `json:"sustainability_period"`
	MaxPerformanceVariability float64       `json:"max_performance_variability"`

	// Alerting configuration
	AlertOnValidationFailure   bool `json:"alert_on_validation_failure"`
	AlertOnOptimizationSuccess bool `json:"alert_on_optimization_success"`

	// Retention settings
	ValidationRetentionDays int `json:"validation_retention_days"`
}

// DefaultOptimizationValidationConfig returns default configuration
func DefaultOptimizationValidationConfig() *OptimizationValidationConfig {
	return &OptimizationValidationConfig{
		ValidationWindow:           24 * time.Hour,
		MinSampleSize:              50,
		MaxValidationAge:           7 * 24 * time.Hour,
		MinResponseTimeImprovement: 10.0, // 10% improvement required
		MinThroughputImprovement:   5.0,  // 5% improvement required
		MaxErrorRateIncrease:       2.0,  // Max 2% error rate increase
		MinResourceEfficiency:      5.0,  // 5% resource efficiency improvement
		StatisticalSignificance:    0.05, // 5% significance level
		SustainabilityPeriod:       2 * time.Hour,
		MaxPerformanceVariability:  15.0, // 15% max variability
		AlertOnValidationFailure:   true,
		AlertOnOptimizationSuccess: false,
		ValidationRetentionDays:    30,
	}
}

// OptimizationResult represents the result of a performance optimization
type OptimizationResult struct {
	ID               string                      `json:"id"`
	OptimizationType string                      `json:"optimization_type"`
	Description      string                      `json:"description"`
	AppliedAt        time.Time                   `json:"applied_at"`
	BaselineMetrics  ValidationPerformanceMetric `json:"baseline_metrics"`
	OptimizedMetrics ValidationPerformanceMetric `json:"optimized_metrics"`
	ExpectedImpact   ExpectedImpact              `json:"expected_impact"`
	Configuration    map[string]interface{}      `json:"configuration,omitempty"`
	Tags             map[string]string           `json:"tags,omitempty"`
}

// ValidationPerformanceMetric represents performance metrics for validation
type ValidationPerformanceMetric struct {
	ResponseTime      time.Duration      `json:"response_time"`
	Throughput        float64            `json:"throughput"`
	ErrorRate         float64            `json:"error_rate"`
	CPUUsage          float64            `json:"cpu_usage"`
	MemoryUsage       float64            `json:"memory_usage"`
	DiskIO            float64            `json:"disk_io"`
	NetworkIO         float64            `json:"network_io"`
	ConcurrentUsers   int                `json:"concurrent_users"`
	Timestamp         time.Time          `json:"timestamp"`
	AdditionalMetrics map[string]float64 `json:"additional_metrics,omitempty"`
}

// ExpectedImpact represents expected performance improvements
type ExpectedImpact struct {
	ResponseTimeReduction float64 `json:"response_time_reduction"` // Percentage
	ThroughputIncrease    float64 `json:"throughput_increase"`     // Percentage
	ErrorRateReduction    float64 `json:"error_rate_reduction"`    // Percentage
	CPUReduction          float64 `json:"cpu_reduction"`           // Percentage
	MemoryReduction       float64 `json:"memory_reduction"`        // Percentage
	Confidence            float64 `json:"confidence"`              // 0-1.0
}

// ValidationResult represents the result of optimization validation
type ValidationResult struct {
	ID                  string                         `json:"id"`
	OptimizationID      string                         `json:"optimization_id"`
	ValidationStarted   time.Time                      `json:"validation_started"`
	ValidationCompleted *time.Time                     `json:"validation_completed,omitempty"`
	Status              ValidationStatus               `json:"status"`
	IsSuccess           bool                           `json:"is_success"`
	ActualImpact        ActualImpact                   `json:"actual_impact"`
	Improvements        []ImprovementMetric            `json:"improvements"`
	Regressions         []OptimizationRegressionMetric `json:"regressions"`
	Sustainability      SustainabilityAssessment       `json:"sustainability"`
	Recommendations     []string                       `json:"recommendations"`
	Confidence          float64                        `json:"confidence"`
	Score               float64                        `json:"score"`
	Metadata            map[string]interface{}         `json:"metadata,omitempty"`
}

// ValidationStatus represents validation status
type ValidationStatus string

const (
	ValidationStatusPending    ValidationStatus = "pending"
	ValidationStatusValidating ValidationStatus = "validating"
	ValidationStatusSuccess    ValidationStatus = "success"
	ValidationStatusFailure    ValidationStatus = "failure"
	ValidationStatusPartial    ValidationStatus = "partial"
	ValidationStatusError      ValidationStatus = "error"
)

// ActualImpact represents actual performance improvements achieved
type ActualImpact struct {
	ResponseTimeChange float64            `json:"response_time_change"` // Percentage change
	ThroughputChange   float64            `json:"throughput_change"`    // Percentage change
	ErrorRateChange    float64            `json:"error_rate_change"`    // Percentage change
	CPUChange          float64            `json:"cpu_change"`           // Percentage change
	MemoryChange       float64            `json:"memory_change"`        // Percentage change
	OverallImprovement float64            `json:"overall_improvement"`  // Weighted score
	StatisticalP       float64            `json:"statistical_p"`        // Statistical significance
	EffectSize         float64            `json:"effect_size"`          // Cohen's d
	AdditionalMetrics  map[string]float64 `json:"additional_metrics,omitempty"`
}

// ImprovementMetric represents a specific improvement
type ImprovementMetric struct {
	Metric             string  `json:"metric"`
	ImprovementPercent float64 `json:"improvement_percent"`
	PreviousValue      float64 `json:"previous_value"`
	CurrentValue       float64 `json:"current_value"`
	StatisticalP       float64 `json:"statistical_p"`
	EffectSize         float64 `json:"effect_size"`
	IsSignificant      bool    `json:"is_significant"`
	Sustainability     float64 `json:"sustainability"` // 0-1.0
}

// OptimizationRegressionMetric represents a performance regression in optimization validation
type OptimizationRegressionMetric struct {
	Metric            string  `json:"metric"`
	RegressionPercent float64 `json:"regression_percent"`
	PreviousValue     float64 `json:"previous_value"`
	CurrentValue      float64 `json:"current_value"`
	StatisticalP      float64 `json:"statistical_p"`
	EffectSize        float64 `json:"effect_size"`
	IsSignificant     bool    `json:"is_significant"`
	Severity          string  `json:"severity"` // low, medium, high, critical
}

// SustainabilityAssessment represents sustainability analysis
type SustainabilityAssessment struct {
	IsSustainable       bool          `json:"is_sustainable"`
	SustainabilityScore float64       `json:"sustainability_score"` // 0-1.0
	VariabilityCoeff    float64       `json:"variability_coefficient"`
	TrendAnalysis       string        `json:"trend_analysis"`
	PredictedDuration   time.Duration `json:"predicted_duration"`
	ConfidenceLevel     float64       `json:"confidence_level"`
}

// PerformanceOptimizationValidator validates optimization results
type PerformanceOptimizationValidator struct {
	config        *OptimizationValidationConfig
	logger        *zap.Logger
	optimizations map[string]*OptimizationResult
	validations   map[string]*ValidationResult
	mutex         sync.RWMutex
	stopCh        chan struct{}
}

// NewPerformanceOptimizationValidator creates a new validator
func NewPerformanceOptimizationValidator(config *OptimizationValidationConfig, logger *zap.Logger) *PerformanceOptimizationValidator {
	if config == nil {
		config = DefaultOptimizationValidationConfig()
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	return &PerformanceOptimizationValidator{
		config:        config,
		logger:        logger,
		optimizations: make(map[string]*OptimizationResult),
		validations:   make(map[string]*ValidationResult),
		stopCh:        make(chan struct{}),
	}
}

// RegisterOptimization registers a new optimization for validation
func (pov *PerformanceOptimizationValidator) RegisterOptimization(ctx context.Context, optimization *OptimizationResult) error {
	if optimization == nil {
		return errors.New("optimization cannot be nil")
	}

	if optimization.ID == "" {
		return errors.New("optimization ID is required")
	}

	pov.mutex.Lock()
	defer pov.mutex.Unlock()

	pov.optimizations[optimization.ID] = optimization

	pov.logger.Info("optimization registered for validation",
		zap.String("optimization_id", optimization.ID),
		zap.String("type", optimization.OptimizationType),
		zap.Time("applied_at", optimization.AppliedAt))

	return nil
}

// ValidateOptimization validates an optimization result
func (pov *PerformanceOptimizationValidator) ValidateOptimization(ctx context.Context, optimizationID string, currentMetrics []ValidationPerformanceMetric) (*ValidationResult, error) {
	if optimizationID == "" {
		return nil, errors.New("optimization ID is required")
	}

	if len(currentMetrics) < pov.config.MinSampleSize {
		return nil, fmt.Errorf("insufficient metrics samples: got %d, need %d", len(currentMetrics), pov.config.MinSampleSize)
	}

	pov.mutex.Lock()
	optimization, exists := pov.optimizations[optimizationID]
	pov.mutex.Unlock()

	if !exists {
		return nil, fmt.Errorf("optimization not found: %s", optimizationID)
	}

	// Create validation result
	result := &ValidationResult{
		ID:                generateValidationID(),
		OptimizationID:    optimizationID,
		ValidationStarted: time.Now(),
		Status:            ValidationStatusValidating,
		Improvements:      []ImprovementMetric{},
		Regressions:       []OptimizationRegressionMetric{},
		Recommendations:   []string{},
		Metadata:          make(map[string]interface{}),
	}

	// Calculate actual impact
	actualImpact, err := pov.calculateActualImpact(optimization.BaselineMetrics, currentMetrics)
	if err != nil {
		result.Status = ValidationStatusError
		return result, fmt.Errorf("failed to calculate actual impact: %w", err)
	}
	result.ActualImpact = *actualImpact

	// Validate improvements
	improvements, regressions := pov.validateImprovements(optimization, actualImpact)
	result.Improvements = improvements
	result.Regressions = regressions

	// Assess sustainability
	sustainability := pov.assessSustainability(currentMetrics)
	result.Sustainability = sustainability

	// Calculate overall validation score
	result.Score = pov.calculateValidationScore(optimization, result)
	result.Confidence = pov.calculateValidationConfidence(result)

	// Determine validation status
	result.IsSuccess = pov.determineValidationSuccess(result)
	if result.IsSuccess {
		result.Status = ValidationStatusSuccess
	} else if len(result.Improvements) > 0 && len(result.Regressions) > 0 {
		result.Status = ValidationStatusPartial
	} else {
		result.Status = ValidationStatusFailure
	}

	// Generate recommendations
	result.Recommendations = pov.generateValidationRecommendations(optimization, result)

	// Complete validation
	now := time.Now()
	result.ValidationCompleted = &now

	// Store result
	pov.mutex.Lock()
	pov.validations[result.ID] = result
	pov.mutex.Unlock()

	pov.logger.Info("optimization validation completed",
		zap.String("validation_id", result.ID),
		zap.String("optimization_id", optimizationID),
		zap.String("status", string(result.Status)),
		zap.Bool("is_success", result.IsSuccess),
		zap.Float64("score", result.Score))

	return result, nil
}

// GetValidationResult retrieves a validation result by ID
func (pov *PerformanceOptimizationValidator) GetValidationResult(validationID string) (*ValidationResult, error) {
	if validationID == "" {
		return nil, errors.New("validation ID is required")
	}

	pov.mutex.RLock()
	defer pov.mutex.RUnlock()

	result, exists := pov.validations[validationID]
	if !exists {
		return nil, fmt.Errorf("validation result not found: %s", validationID)
	}

	return result, nil
}

// GetOptimization retrieves an optimization by ID
func (pov *PerformanceOptimizationValidator) GetOptimization(optimizationID string) (*OptimizationResult, error) {
	if optimizationID == "" {
		return nil, errors.New("optimization ID is required")
	}

	pov.mutex.RLock()
	defer pov.mutex.RUnlock()

	optimization, exists := pov.optimizations[optimizationID]
	if !exists {
		return nil, fmt.Errorf("optimization not found: %s", optimizationID)
	}

	return optimization, nil
}

// ListValidationResults lists all validation results
func (pov *PerformanceOptimizationValidator) ListValidationResults() []*ValidationResult {
	pov.mutex.RLock()
	defer pov.mutex.RUnlock()

	results := make([]*ValidationResult, 0, len(pov.validations))
	for _, result := range pov.validations {
		results = append(results, result)
	}

	// Sort by validation started time (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].ValidationStarted.After(results[j].ValidationStarted)
	})

	return results
}

// ListOptimizations lists all registered optimizations
func (pov *PerformanceOptimizationValidator) ListOptimizations() []*OptimizationResult {
	pov.mutex.RLock()
	defer pov.mutex.RUnlock()

	optimizations := make([]*OptimizationResult, 0, len(pov.optimizations))
	for _, optimization := range pov.optimizations {
		optimizations = append(optimizations, optimization)
	}

	// Sort by applied time (newest first)
	sort.Slice(optimizations, func(i, j int) bool {
		return optimizations[i].AppliedAt.After(optimizations[j].AppliedAt)
	})

	return optimizations
}

// Cleanup removes old validations and optimizations
func (pov *PerformanceOptimizationValidator) Cleanup() error {
	pov.mutex.Lock()
	defer pov.mutex.Unlock()

	cutoff := time.Now().AddDate(0, 0, -pov.config.ValidationRetentionDays)

	// Cleanup old validations
	validationsRemoved := 0
	for id, result := range pov.validations {
		if result.ValidationStarted.Before(cutoff) {
			delete(pov.validations, id)
			validationsRemoved++
		}
	}

	// Cleanup old optimizations
	optimizationsRemoved := 0
	for id, optimization := range pov.optimizations {
		if optimization.AppliedAt.Before(cutoff) {
			delete(pov.optimizations, id)
			optimizationsRemoved++
		}
	}

	pov.logger.Info("cleanup completed",
		zap.Int("validations_removed", validationsRemoved),
		zap.Int("optimizations_removed", optimizationsRemoved))

	return nil
}

// Shutdown gracefully shuts down the validator
func (pov *PerformanceOptimizationValidator) Shutdown() error {
	close(pov.stopCh)

	pov.logger.Info("performance optimization validator shut down")
	return nil
}

// calculateActualImpact calculates the actual impact of optimization
func (pov *PerformanceOptimizationValidator) calculateActualImpact(baseline ValidationPerformanceMetric, currentMetrics []ValidationPerformanceMetric) (*ActualImpact, error) {
	if len(currentMetrics) == 0 {
		return nil, errors.New("no current metrics provided")
	}

	// Calculate average current metrics
	current := pov.calculateAverageMetrics(currentMetrics)

	// Calculate percentage changes
	responseTimeChange := calculatePercentageChange(float64(baseline.ResponseTime), float64(current.ResponseTime))
	throughputChange := calculatePercentageChange(baseline.Throughput, current.Throughput)
	errorRateChange := calculatePercentageChange(baseline.ErrorRate, current.ErrorRate)
	cpuChange := calculatePercentageChange(baseline.CPUUsage, current.CPUUsage)
	memoryChange := calculatePercentageChange(baseline.MemoryUsage, current.MemoryUsage)

	// Calculate statistical significance for response time (as example)
	responseTimeSamples := make([]float64, len(currentMetrics))
	for i, metric := range currentMetrics {
		responseTimeSamples[i] = float64(metric.ResponseTime)
	}
	statP := calculateTTestSingle(responseTimeSamples, float64(baseline.ResponseTime))
	effectSize := calculateEffectSizeSingle(responseTimeSamples, float64(baseline.ResponseTime), calculateStdDev(responseTimeSamples))

	// Calculate overall improvement (weighted score)
	overallImprovement := pov.calculateOverallImprovement(responseTimeChange, throughputChange, errorRateChange, cpuChange, memoryChange)

	return &ActualImpact{
		ResponseTimeChange: responseTimeChange,
		ThroughputChange:   throughputChange,
		ErrorRateChange:    errorRateChange,
		CPUChange:          cpuChange,
		MemoryChange:       memoryChange,
		OverallImprovement: overallImprovement,
		StatisticalP:       statP,
		EffectSize:         effectSize,
	}, nil
}

// validateImprovements validates expected vs actual improvements
func (pov *PerformanceOptimizationValidator) validateImprovements(optimization *OptimizationResult, actualImpact *ActualImpact) ([]ImprovementMetric, []OptimizationRegressionMetric) {
	improvements := []ImprovementMetric{}
	regressions := []OptimizationRegressionMetric{}

	// Validate response time improvement
	if actualImpact.ResponseTimeChange < 0 && math.Abs(actualImpact.ResponseTimeChange) >= pov.config.MinResponseTimeImprovement {
		improvements = append(improvements, ImprovementMetric{
			Metric:             "response_time",
			ImprovementPercent: math.Abs(actualImpact.ResponseTimeChange),
			StatisticalP:       actualImpact.StatisticalP,
			EffectSize:         actualImpact.EffectSize,
			IsSignificant:      actualImpact.StatisticalP < pov.config.StatisticalSignificance,
			Sustainability:     0.8, // Default value
		})
	} else if actualImpact.ResponseTimeChange > 0 {
		regressions = append(regressions, OptimizationRegressionMetric{
			Metric:            "response_time",
			RegressionPercent: actualImpact.ResponseTimeChange,
			StatisticalP:      actualImpact.StatisticalP,
			EffectSize:        actualImpact.EffectSize,
			IsSignificant:     actualImpact.StatisticalP < pov.config.StatisticalSignificance,
			Severity:          determineSeverity(actualImpact.ResponseTimeChange),
		})
	}

	// Validate throughput improvement
	if actualImpact.ThroughputChange > 0 && actualImpact.ThroughputChange >= pov.config.MinThroughputImprovement {
		improvements = append(improvements, ImprovementMetric{
			Metric:             "throughput",
			ImprovementPercent: actualImpact.ThroughputChange,
			IsSignificant:      true, // Simplified
			Sustainability:     0.8,
		})
	} else if actualImpact.ThroughputChange < 0 {
		regressions = append(regressions, OptimizationRegressionMetric{
			Metric:            "throughput",
			RegressionPercent: math.Abs(actualImpact.ThroughputChange),
			IsSignificant:     true,
			Severity:          determineSeverity(math.Abs(actualImpact.ThroughputChange)),
		})
	}

	// Validate error rate (should not increase significantly)
	if actualImpact.ErrorRateChange > pov.config.MaxErrorRateIncrease {
		regressions = append(regressions, OptimizationRegressionMetric{
			Metric:            "error_rate",
			RegressionPercent: actualImpact.ErrorRateChange,
			IsSignificant:     true,
			Severity:          determineSeverity(actualImpact.ErrorRateChange),
		})
	}

	return improvements, regressions
}

// assessSustainability assesses the sustainability of improvements
func (pov *PerformanceOptimizationValidator) assessSustainability(metrics []ValidationPerformanceMetric) SustainabilityAssessment {
	if len(metrics) < 10 {
		return SustainabilityAssessment{
			IsSustainable:       false,
			SustainabilityScore: 0.0,
			VariabilityCoeff:    1.0,
			TrendAnalysis:       "insufficient_data",
			ConfidenceLevel:     0.0,
		}
	}

	// Calculate variability coefficient for response time
	responseTimes := make([]float64, len(metrics))
	for i, metric := range metrics {
		responseTimes[i] = float64(metric.ResponseTime)
	}

	mean := calculateMean(responseTimes)
	stdDev := calculateStdDev(responseTimes)
	variabilityCoeff := stdDev / mean * 100

	// Assess sustainability
	isSustainable := variabilityCoeff <= pov.config.MaxPerformanceVariability
	sustainabilityScore := math.Max(0, 1.0-(variabilityCoeff/pov.config.MaxPerformanceVariability))

	// Simple trend analysis
	trendAnalysis := "stable"
	if variabilityCoeff > pov.config.MaxPerformanceVariability {
		trendAnalysis = "unstable"
	}

	return SustainabilityAssessment{
		IsSustainable:       isSustainable,
		SustainabilityScore: sustainabilityScore,
		VariabilityCoeff:    variabilityCoeff,
		TrendAnalysis:       trendAnalysis,
		PredictedDuration:   24 * time.Hour, // Default prediction
		ConfidenceLevel:     sustainabilityScore,
	}
}

// calculateValidationScore calculates overall validation score
func (pov *PerformanceOptimizationValidator) calculateValidationScore(optimization *OptimizationResult, result *ValidationResult) float64 {
	score := 0.0

	// Score improvements (positive contribution)
	for _, improvement := range result.Improvements {
		weight := 0.3
		if improvement.IsSignificant {
			weight = 0.5
		}
		score += improvement.ImprovementPercent * weight * improvement.Sustainability
	}

	// Penalize regressions (negative contribution)
	for _, regression := range result.Regressions {
		penalty := regression.RegressionPercent * 0.5
		if regression.IsSignificant {
			penalty *= 2.0
		}
		score -= penalty
	}

	// Sustainability bonus
	score += result.Sustainability.SustainabilityScore * 20.0

	// Normalize to 0-100 scale
	score = math.Max(0, math.Min(100, score))

	return score
}

// calculateValidationConfidence calculates validation confidence
func (pov *PerformanceOptimizationValidator) calculateValidationConfidence(result *ValidationResult) float64 {
	confidence := 0.5 // Base confidence

	// Increase confidence for significant improvements
	significantImprovements := 0
	for _, improvement := range result.Improvements {
		if improvement.IsSignificant {
			significantImprovements++
		}
	}
	confidence += float64(significantImprovements) * 0.1

	// Decrease confidence for regressions
	significantRegressions := 0
	for _, regression := range result.Regressions {
		if regression.IsSignificant {
			significantRegressions++
		}
	}
	confidence -= float64(significantRegressions) * 0.15

	// Factor in sustainability
	confidence += result.Sustainability.SustainabilityScore * 0.3

	return math.Max(0, math.Min(1.0, confidence))
}

// determineValidationSuccess determines if validation was successful
func (pov *PerformanceOptimizationValidator) determineValidationSuccess(result *ValidationResult) bool {
	// Must have at least one significant improvement
	hasSignificantImprovement := false
	for _, improvement := range result.Improvements {
		if improvement.IsSignificant {
			hasSignificantImprovement = true
			break
		}
	}

	// Must not have critical regressions
	hasCriticalRegression := false
	for _, regression := range result.Regressions {
		if regression.Severity == "critical" && regression.IsSignificant {
			hasCriticalRegression = true
			break
		}
	}

	// Must be sustainable
	isSustainable := result.Sustainability.IsSustainable

	return hasSignificantImprovement && !hasCriticalRegression && isSustainable
}

// generateValidationRecommendations generates actionable recommendations
func (pov *PerformanceOptimizationValidator) generateValidationRecommendations(optimization *OptimizationResult, result *ValidationResult) []string {
	recommendations := []string{}

	if result.IsSuccess {
		recommendations = append(recommendations, "Optimization validation successful - consider scaling this optimization to similar services")

		if result.Sustainability.SustainabilityScore < 0.8 {
			recommendations = append(recommendations, "Monitor sustainability metrics closely as variability is higher than optimal")
		}
	} else {
		// Analyze specific issues
		if len(result.Improvements) == 0 {
			recommendations = append(recommendations, "No significant improvements detected - review optimization configuration and implementation")
		}

		for _, regression := range result.Regressions {
			if regression.Severity == "critical" {
				recommendations = append(recommendations, fmt.Sprintf("Critical regression in %s detected - consider rolling back optimization", regression.Metric))
			} else if regression.Severity == "high" {
				recommendations = append(recommendations, fmt.Sprintf("High severity regression in %s - investigate and tune optimization parameters", regression.Metric))
			}
		}

		if !result.Sustainability.IsSustainable {
			recommendations = append(recommendations, "Performance improvements are not sustainable - review resource allocation and scaling policies")
		}
	}

	// Generic recommendations based on metrics
	if result.ActualImpact.OverallImprovement < 5.0 {
		recommendations = append(recommendations, "Overall improvement is minimal - consider alternative optimization strategies")
	}

	if result.Confidence < 0.7 {
		recommendations = append(recommendations, "Validation confidence is low - collect more data samples before making decisions")
	}

	return recommendations
}

// Helper functions

// calculateAverageMetrics calculates average metrics from a slice
func (pov *PerformanceOptimizationValidator) calculateAverageMetrics(metrics []ValidationPerformanceMetric) ValidationPerformanceMetric {
	if len(metrics) == 0 {
		return ValidationPerformanceMetric{}
	}

	var totalResponseTime time.Duration
	var totalThroughput, totalErrorRate, totalCPU, totalMemory, totalDiskIO, totalNetworkIO float64
	var totalUsers int

	for _, metric := range metrics {
		totalResponseTime += metric.ResponseTime
		totalThroughput += metric.Throughput
		totalErrorRate += metric.ErrorRate
		totalCPU += metric.CPUUsage
		totalMemory += metric.MemoryUsage
		totalDiskIO += metric.DiskIO
		totalNetworkIO += metric.NetworkIO
		totalUsers += metric.ConcurrentUsers
	}

	count := len(metrics)
	return ValidationPerformanceMetric{
		ResponseTime:    totalResponseTime / time.Duration(count),
		Throughput:      totalThroughput / float64(count),
		ErrorRate:       totalErrorRate / float64(count),
		CPUUsage:        totalCPU / float64(count),
		MemoryUsage:     totalMemory / float64(count),
		DiskIO:          totalDiskIO / float64(count),
		NetworkIO:       totalNetworkIO / float64(count),
		ConcurrentUsers: totalUsers / count,
		Timestamp:       time.Now(),
	}
}

// calculateOverallImprovement calculates weighted overall improvement score
func (pov *PerformanceOptimizationValidator) calculateOverallImprovement(responseTime, throughput, errorRate, cpu, memory float64) float64 {
	// Weights for different metrics
	weights := map[string]float64{
		"response_time": 0.3,
		"throughput":    0.25,
		"error_rate":    0.2,
		"cpu":           0.15,
		"memory":        0.1,
	}

	score := 0.0

	// Response time improvement (negative change is good)
	if responseTime < 0 {
		score += math.Abs(responseTime) * weights["response_time"]
	}

	// Throughput improvement (positive change is good)
	if throughput > 0 {
		score += throughput * weights["throughput"]
	}

	// Error rate improvement (negative change is good)
	if errorRate < 0 {
		score += math.Abs(errorRate) * weights["error_rate"]
	}

	// CPU improvement (negative change is good)
	if cpu < 0 {
		score += math.Abs(cpu) * weights["cpu"]
	}

	// Memory improvement (negative change is good)
	if memory < 0 {
		score += math.Abs(memory) * weights["memory"]
	}

	return score
}

// Statistical helper functions
func calculatePercentageChange(baseline, current float64) float64 {
	if baseline == 0 {
		if current == 0 {
			return 0
		}
		return 100 // Arbitrary large change
	}
	return ((current - baseline) / baseline) * 100
}

func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func calculateStdDev(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := calculateMean(values)
	variance := 0.0
	for _, value := range values {
		variance += (value - mean) * (value - mean)
	}
	variance /= float64(len(values))
	return math.Sqrt(variance)
}

func calculateTTestSingle(sample []float64, populationMean float64) float64 {
	if len(sample) < 2 {
		return 1.0
	}

	sampleMean := calculateMean(sample)
	sampleStd := calculateStdDev(sample)

	if sampleStd == 0 {
		return 1.0
	}

	t := (sampleMean - populationMean) / (sampleStd / math.Sqrt(float64(len(sample))))

	// Simplified p-value calculation (approximation)
	// In practice, you'd use a proper t-distribution table
	absT := math.Abs(t)
	if absT > 2.576 {
		return 0.01 // Highly significant
	} else if absT > 1.96 {
		return 0.05 // Significant
	} else if absT > 1.645 {
		return 0.1 // Marginally significant
	}
	return 0.2 // Not significant
}

func calculateEffectSizeSingle(sample []float64, populationMean, populationStd float64) float64 {
	if populationStd == 0 {
		return 0
	}
	sampleMean := calculateMean(sample)
	return (sampleMean - populationMean) / populationStd
}

func determineSeverity(changePercent float64) string {
	absChange := math.Abs(changePercent)
	if absChange > 50 {
		return "critical"
	} else if absChange > 20 {
		return "high"
	} else if absChange > 10 {
		return "medium"
	}
	return "low"
}

func generateValidationID() string {
	return fmt.Sprintf("val_%d", time.Now().UnixNano())
}
