package data_discovery

import (
	"fmt"
	"sort"
	"time"

	"go.uber.org/zap"
)

// NewExtractionOptimizer creates a new extraction optimizer
func NewExtractionOptimizer(config *ExtractionMonitorConfig, logger *zap.Logger, metrics *ExtractionMetrics) *ExtractionOptimizer {
	optimizer := &ExtractionOptimizer{
		config:     config,
		logger:     logger,
		metrics:    metrics,
		strategies: make([]OptimizationStrategy, 0),
	}

	// Initialize default optimization strategies
	optimizer.initializeDefaultStrategies()

	return optimizer
}

// initializeDefaultStrategies sets up default optimization strategies
func (eo *ExtractionOptimizer) initializeDefaultStrategies() {
	eo.strategies = []OptimizationStrategy{
		{
			Name:        "pattern_optimization",
			Description: "Optimize pattern detection for better accuracy and performance",
			Enabled:     true,
			Priority:    1,
			Parameters: map[string]interface{}{
				"confidence_threshold": 0.7,
				"max_patterns":         20,
				"pattern_timeout":      100 * time.Millisecond,
			},
			Effectiveness: 0.0,
		},
		{
			Name:        "field_prioritization",
			Description: "Prioritize high-value fields for extraction",
			Enabled:     true,
			Priority:    2,
			Parameters: map[string]interface{}{
				"priority_fields":       []string{"email", "phone", "address"},
				"quality_threshold":     0.8,
				"business_value_weight": 0.7,
			},
			Effectiveness: 0.0,
		},
		{
			Name:        "resource_optimization",
			Description: "Optimize resource usage and processing efficiency",
			Enabled:     true,
			Priority:    3,
			Parameters: map[string]interface{}{
				"max_concurrent":   10,
				"memory_limit_mb":  256,
				"cpu_threshold":    0.8,
				"timeout_duration": 5 * time.Second,
			},
			Effectiveness: 0.0,
		},
		{
			Name:        "quality_improvement",
			Description: "Improve data quality through enhanced validation and scoring",
			Enabled:     true,
			Priority:    4,
			Parameters: map[string]interface{}{
				"min_quality_score":     0.7,
				"validation_strictness": 0.8,
				"cross_validation":      true,
			},
			Effectiveness: 0.0,
		},
		{
			Name:        "error_reduction",
			Description: "Reduce errors through improved error handling and retry logic",
			Enabled:     true,
			Priority:    5,
			Parameters: map[string]interface{}{
				"max_retries":     3,
				"retry_delay":     100 * time.Millisecond,
				"error_threshold": 0.05,
				"circuit_breaker": true,
			},
			Effectiveness: 0.0,
		},
	}
}

// RunOptimization executes optimization strategies based on current metrics
func (eo *ExtractionOptimizer) RunOptimization() {
	eo.logger.Info("Starting extraction optimization")

	// Get current metrics
	metrics := eo.metrics
	if metrics == nil {
		eo.logger.Warn("No metrics available for optimization")
		return
	}

	// Analyze current performance
	analysis := eo.analyzePerformance(metrics)

	// Select and execute optimization strategies
	appliedStrategies := eo.selectAndExecuteStrategies(analysis)

	// Measure effectiveness of applied strategies
	eo.measureEffectiveness(appliedStrategies)

	eo.logger.Info("Extraction optimization completed",
		zap.Int("strategies_applied", len(appliedStrategies)))
}

// PerformanceAnalysis represents analysis of current performance
type PerformanceAnalysis struct {
	OverallScore              float64                      `json:"overall_score"`
	PerformanceIssues         []PerformanceIssue           `json:"performance_issues"`
	QualityIssues             []QualityIssue               `json:"quality_issues"`
	ResourceIssues            []ResourceIssue              `json:"resource_issues"`
	OptimizationOpportunities []OptimizationOpportunity    `json:"optimization_opportunities"`
	Recommendations           []OptimizationRecommendation `json:"recommendations"`
}

// PerformanceIssue represents a performance-related issue
type PerformanceIssue struct {
	Type         string      `json:"type"`
	Description  string      `json:"description"`
	Severity     string      `json:"severity"`
	Impact       float64     `json:"impact"`
	Metric       string      `json:"metric"`
	CurrentValue interface{} `json:"current_value"`
	TargetValue  interface{} `json:"target_value"`
}

// QualityIssue represents a quality-related issue
type QualityIssue struct {
	Type         string  `json:"type"`
	Description  string  `json:"description"`
	Severity     string  `json:"severity"`
	Impact       float64 `json:"impact"`
	FieldType    string  `json:"field_type"`
	CurrentScore float64 `json:"current_score"`
	TargetScore  float64 `json:"target_score"`
}

// ResourceIssue represents a resource-related issue
type ResourceIssue struct {
	Type         string      `json:"type"`
	Description  string      `json:"description"`
	Severity     string      `json:"severity"`
	Impact       float64     `json:"impact"`
	Resource     string      `json:"resource"`
	CurrentUsage interface{} `json:"current_usage"`
	Threshold    interface{} `json:"threshold"`
}

// OptimizationOpportunity represents an opportunity for optimization
type OptimizationOpportunity struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Potential   float64 `json:"potential_improvement"`
	Effort      string  `json:"effort_required"`
	Priority    int     `json:"priority"`
	Strategy    string  `json:"recommended_strategy"`
}

// analyzePerformance analyzes current performance and identifies issues
func (eo *ExtractionOptimizer) analyzePerformance(metrics *ExtractionMetrics) *PerformanceAnalysis {
	analysis := &PerformanceAnalysis{
		PerformanceIssues:         make([]PerformanceIssue, 0),
		QualityIssues:             make([]QualityIssue, 0),
		ResourceIssues:            make([]ResourceIssue, 0),
		OptimizationOpportunities: make([]OptimizationOpportunity, 0),
		Recommendations:           make([]OptimizationRecommendation, 0),
	}

	// Analyze performance metrics
	eo.analyzePerformanceMetrics(metrics, analysis)

	// Analyze quality metrics
	eo.analyzeQualityMetrics(metrics, analysis)

	// Analyze resource metrics
	eo.analyzeResourceMetrics(metrics, analysis)

	// Calculate overall score
	analysis.OverallScore = eo.calculateOverallScore(analysis)

	// Generate optimization opportunities
	eo.generateOptimizationOpportunities(analysis)

	// Generate recommendations
	eo.generateRecommendations(analysis)

	return analysis
}

// analyzePerformanceMetrics analyzes performance-related metrics
func (eo *ExtractionOptimizer) analyzePerformanceMetrics(metrics *ExtractionMetrics, analysis *PerformanceAnalysis) {
	// Check processing time
	if metrics.AverageProcessingTime > eo.config.PerformanceThresholds.MaxProcessingTime {
		analysis.PerformanceIssues = append(analysis.PerformanceIssues, PerformanceIssue{
			Type:         "processing_time",
			Description:  "Average processing time exceeds threshold",
			Severity:     "warning",
			Impact:       0.3,
			Metric:       "average_processing_time",
			CurrentValue: metrics.AverageProcessingTime,
			TargetValue:  eo.config.PerformanceThresholds.MaxProcessingTime,
		})
	}

	// Check success rate
	successRate := float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests)
	if successRate < eo.config.PerformanceThresholds.MinSuccessRate {
		analysis.PerformanceIssues = append(analysis.PerformanceIssues, PerformanceIssue{
			Type:         "success_rate",
			Description:  "Success rate below threshold",
			Severity:     "critical",
			Impact:       0.8,
			Metric:       "success_rate",
			CurrentValue: successRate,
			TargetValue:  eo.config.PerformanceThresholds.MinSuccessRate,
		})
	}

	// Check fields discovered per request
	if metrics.FieldsDiscoveredPerRequest < float64(eo.config.PerformanceThresholds.MinDataPointsPerBusiness) {
		analysis.PerformanceIssues = append(analysis.PerformanceIssues, PerformanceIssue{
			Type:         "data_points",
			Description:  "Insufficient data points discovered per request",
			Severity:     "medium",
			Impact:       0.5,
			Metric:       "fields_discovered_per_request",
			CurrentValue: metrics.FieldsDiscoveredPerRequest,
			TargetValue:  eo.config.PerformanceThresholds.MinDataPointsPerBusiness,
		})
	}
}

// analyzeQualityMetrics analyzes quality-related metrics
func (eo *ExtractionOptimizer) analyzeQualityMetrics(metrics *ExtractionMetrics, analysis *PerformanceAnalysis) {
	// Check overall quality score
	if metrics.AverageQualityScore < eo.config.PerformanceThresholds.MinQualityScore {
		analysis.QualityIssues = append(analysis.QualityIssues, QualityIssue{
			Type:         "overall_quality",
			Description:  "Average quality score below threshold",
			Severity:     "warning",
			Impact:       0.4,
			FieldType:    "overall",
			CurrentScore: metrics.AverageQualityScore,
			TargetScore:  eo.config.PerformanceThresholds.MinQualityScore,
		})
	}

	// Check quality distribution
	excellentCount := metrics.QualityScoreDistribution["excellent"]
	goodCount := metrics.QualityScoreDistribution["good"]
	totalAssessments := excellentCount + goodCount

	if totalAssessments > 0 {
		excellentRatio := float64(excellentCount) / float64(totalAssessments)
		if excellentRatio < 0.6 {
			analysis.QualityIssues = append(analysis.QualityIssues, QualityIssue{
				Type:         "quality_distribution",
				Description:  "Low ratio of excellent quality assessments",
				Severity:     "medium",
				Impact:       0.3,
				FieldType:    "distribution",
				CurrentScore: excellentRatio,
				TargetScore:  0.6,
			})
		}
	}

	// Check field-specific quality scores
	for fieldType, qualityScore := range metrics.FieldQualityScores {
		if qualityScore < 0.7 {
			analysis.QualityIssues = append(analysis.QualityIssues, QualityIssue{
				Type:         "field_quality",
				Description:  fmt.Sprintf("Low quality score for field type: %s", fieldType),
				Severity:     "medium",
				Impact:       0.3,
				FieldType:    fieldType,
				CurrentScore: qualityScore,
				TargetScore:  0.7,
			})
		}
	}
}

// analyzeResourceMetrics analyzes resource-related metrics
func (eo *ExtractionOptimizer) analyzeResourceMetrics(metrics *ExtractionMetrics, analysis *PerformanceAnalysis) {
	// Check memory usage
	if metrics.MemoryUsage > eo.config.PerformanceThresholds.MaxMemoryUsage {
		analysis.ResourceIssues = append(analysis.ResourceIssues, ResourceIssue{
			Type:         "memory_usage",
			Description:  "Memory usage exceeds threshold",
			Severity:     "warning",
			Impact:       0.4,
			Resource:     "memory",
			CurrentUsage: metrics.MemoryUsage,
			Threshold:    eo.config.PerformanceThresholds.MaxMemoryUsage,
		})
	}

	// Check CPU usage
	if metrics.CPUUsage > 80.0 {
		analysis.ResourceIssues = append(analysis.ResourceIssues, ResourceIssue{
			Type:         "cpu_usage",
			Description:  "CPU usage is high",
			Severity:     "medium",
			Impact:       0.3,
			Resource:     "cpu",
			CurrentUsage: metrics.CPUUsage,
			Threshold:    80.0,
		})
	}

	// Check concurrent requests
	if metrics.ConcurrentRequests > 20 {
		analysis.ResourceIssues = append(analysis.ResourceIssues, ResourceIssue{
			Type:         "concurrent_requests",
			Description:  "High number of concurrent requests",
			Severity:     "medium",
			Impact:       0.3,
			Resource:     "concurrency",
			CurrentUsage: metrics.ConcurrentRequests,
			Threshold:    20,
		})
	}
}

// calculateOverallScore calculates an overall performance score
func (eo *ExtractionOptimizer) calculateOverallScore(analysis *PerformanceAnalysis) float64 {
	score := 1.0

	// Deduct points for issues
	for _, issue := range analysis.PerformanceIssues {
		score -= issue.Impact * 0.1
	}

	for _, issue := range analysis.QualityIssues {
		score -= issue.Impact * 0.1
	}

	for _, issue := range analysis.ResourceIssues {
		score -= issue.Impact * 0.1
	}

	// Ensure score is between 0 and 1
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}

// generateOptimizationOpportunities identifies optimization opportunities
func (eo *ExtractionOptimizer) generateOptimizationOpportunities(analysis *PerformanceAnalysis) {
	// Performance optimization opportunities
	if len(analysis.PerformanceIssues) > 0 {
		analysis.OptimizationOpportunities = append(analysis.OptimizationOpportunities, OptimizationOpportunity{
			Type:        "performance",
			Description: "Improve processing performance and success rates",
			Potential:   0.3,
			Effort:      "medium",
			Priority:    1,
			Strategy:    "pattern_optimization",
		})
	}

	// Quality optimization opportunities
	if len(analysis.QualityIssues) > 0 {
		analysis.OptimizationOpportunities = append(analysis.OptimizationOpportunities, OptimizationOpportunity{
			Type:        "quality",
			Description: "Enhance data quality and validation",
			Potential:   0.4,
			Effort:      "high",
			Priority:    2,
			Strategy:    "quality_improvement",
		})
	}

	// Resource optimization opportunities
	if len(analysis.ResourceIssues) > 0 {
		analysis.OptimizationOpportunities = append(analysis.OptimizationOpportunities, OptimizationOpportunity{
			Type:        "resource",
			Description: "Optimize resource usage and efficiency",
			Potential:   0.2,
			Effort:      "low",
			Priority:    3,
			Strategy:    "resource_optimization",
		})
	}
}

// generateRecommendations generates optimization recommendations
func (eo *ExtractionOptimizer) generateRecommendations(analysis *PerformanceAnalysis) {
	// Sort optimization opportunities by priority
	sort.Slice(analysis.OptimizationOpportunities, func(i, j int) bool {
		return analysis.OptimizationOpportunities[i].Priority < analysis.OptimizationOpportunities[j].Priority
	})

	// Generate recommendations based on opportunities
	for _, opportunity := range analysis.OptimizationOpportunities {
		recommendation := OptimizationRecommendation{
			Type:               opportunity.Type,
			Description:        opportunity.Description,
			Priority:           eo.getPriorityString(opportunity.Priority),
			ExpectedImpact:     opportunity.Potential,
			ImplementationCost: opportunity.Effort,
			Actions:            eo.getActionsForStrategy(opportunity.Strategy),
		}
		analysis.Recommendations = append(analysis.Recommendations, recommendation)
	}
}

// selectAndExecuteStrategies selects and executes optimization strategies
func (eo *ExtractionOptimizer) selectAndExecuteStrategies(analysis *PerformanceAnalysis) []OptimizationStrategy {
	var appliedStrategies []OptimizationStrategy

	// Sort strategies by priority
	sort.Slice(eo.strategies, func(i, j int) bool {
		return eo.strategies[i].Priority < eo.strategies[j].Priority
	})

	// Apply strategies based on analysis
	for i := range eo.strategies {
		strategy := &eo.strategies[i]
		if !strategy.Enabled {
			continue
		}

		// Check if strategy should be applied based on analysis
		if eo.shouldApplyStrategy(strategy, analysis) {
			eo.applyStrategy(strategy)
			appliedStrategies = append(appliedStrategies, *strategy)
		}
	}

	return appliedStrategies
}

// shouldApplyStrategy determines if a strategy should be applied
func (eo *ExtractionOptimizer) shouldApplyStrategy(strategy *OptimizationStrategy, analysis *PerformanceAnalysis) bool {
	switch strategy.Name {
	case "pattern_optimization":
		return len(analysis.PerformanceIssues) > 0
	case "field_prioritization":
		return len(analysis.QualityIssues) > 0
	case "resource_optimization":
		return len(analysis.ResourceIssues) > 0
	case "quality_improvement":
		return len(analysis.QualityIssues) > 0
	case "error_reduction":
		return len(analysis.PerformanceIssues) > 0
	default:
		return false
	}
}

// applyStrategy applies an optimization strategy
func (eo *ExtractionOptimizer) applyStrategy(strategy *OptimizationStrategy) {
	eo.logger.Info("Applying optimization strategy",
		zap.String("strategy", strategy.Name),
		zap.String("description", strategy.Description))

	// Update strategy parameters based on current performance
	eo.updateStrategyParameters(strategy)

	// Mark strategy as applied
	strategy.LastApplied = time.Now()

	eo.logger.Info("Optimization strategy applied successfully",
		zap.String("strategy", strategy.Name))
}

// updateStrategyParameters updates strategy parameters based on current performance
func (eo *ExtractionOptimizer) updateStrategyParameters(strategy *OptimizationStrategy) {
	switch strategy.Name {
	case "pattern_optimization":
		// Adjust confidence threshold based on current success rate
		if eo.metrics.TotalRequests > 0 {
			successRate := float64(eo.metrics.SuccessfulRequests) / float64(eo.metrics.TotalRequests)
			if successRate < 0.9 {
				strategy.Parameters["confidence_threshold"] = 0.8
			} else {
				strategy.Parameters["confidence_threshold"] = 0.7
			}
		}

	case "field_prioritization":
		// Update priority fields based on business value
		priorityFields := []string{"email", "phone", "address"}
		if eo.metrics.AverageQualityScore < 0.8 {
			priorityFields = append(priorityFields, "tax_id", "url")
		}
		strategy.Parameters["priority_fields"] = priorityFields

	case "resource_optimization":
		// Adjust resource limits based on current usage
		if eo.metrics.MemoryUsage > 200 {
			strategy.Parameters["memory_limit_mb"] = 512
		}
		if eo.metrics.CPUUsage > 70 {
			strategy.Parameters["max_concurrent"] = 5
		}
	}
}

// measureEffectiveness measures the effectiveness of applied strategies
func (eo *ExtractionOptimizer) measureEffectiveness(appliedStrategies []OptimizationStrategy) {
	for i := range appliedStrategies {
		strategy := &appliedStrategies[i]

		// Calculate effectiveness based on performance improvement
		effectiveness := eo.calculateStrategyEffectiveness(strategy)
		strategy.Effectiveness = effectiveness

		eo.logger.Info("Strategy effectiveness measured",
			zap.String("strategy", strategy.Name),
			zap.Float64("effectiveness", effectiveness))
	}
}

// calculateStrategyEffectiveness calculates the effectiveness of a strategy
func (eo *ExtractionOptimizer) calculateStrategyEffectiveness(strategy *OptimizationStrategy) float64 {
	// This would typically compare before/after metrics
	// For now, return a placeholder value
	return 0.75
}

// getPriorityString converts priority number to string
func (eo *ExtractionOptimizer) getPriorityString(priority int) string {
	switch priority {
	case 1:
		return "high"
	case 2:
		return "medium"
	case 3:
		return "low"
	default:
		return "low"
	}
}

// getActionsForStrategy returns actions for a specific strategy
func (eo *ExtractionOptimizer) getActionsForStrategy(strategyName string) []string {
	switch strategyName {
	case "pattern_optimization":
		return []string{
			"Review and update pattern detection algorithms",
			"Adjust confidence thresholds",
			"Optimize pattern matching performance",
		}
	case "field_prioritization":
		return []string{
			"Identify high-value fields",
			"Update field priority weights",
			"Implement business value scoring",
		}
	case "resource_optimization":
		return []string{
			"Monitor resource usage",
			"Adjust concurrency limits",
			"Optimize memory allocation",
		}
	case "quality_improvement":
		return []string{
			"Enhance validation rules",
			"Implement cross-validation",
			"Improve quality scoring algorithms",
		}
	case "error_reduction":
		return []string{
			"Implement retry logic",
			"Add circuit breaker patterns",
			"Improve error handling",
		}
	default:
		return []string{"Review and optimize strategy"}
	}
}

// GetOptimizationStrategies returns current optimization strategies
func (eo *ExtractionOptimizer) GetOptimizationStrategies() []OptimizationStrategy {
	// eo.mu.RLock()
	// defer eo.mu.RUnlock()

	strategies := make([]OptimizationStrategy, len(eo.strategies))
	copy(strategies, eo.strategies)
	return strategies
}

// EnableStrategy enables or disables an optimization strategy
func (eo *ExtractionOptimizer) EnableStrategy(strategyName string, enabled bool) error {
	// eo.mu.Lock()
	// defer eo.mu.Unlock()

	for i := range eo.strategies {
		if eo.strategies[i].Name == strategyName {
			eo.strategies[i].Enabled = enabled
			eo.logger.Info("Strategy enabled/disabled",
				zap.String("strategy", strategyName),
				zap.Bool("enabled", enabled))
			return nil
		}
	}

	return fmt.Errorf("strategy not found: %s", strategyName)
}

// UpdateStrategyParameters updates parameters for a specific strategy
func (eo *ExtractionOptimizer) UpdateStrategyParameters(strategyName string, parameters map[string]interface{}) error {
	// eo.mu.Lock()
	// defer eo.mu.Unlock()

	for i := range eo.strategies {
		if eo.strategies[i].Name == strategyName {
			eo.strategies[i].Parameters = parameters
			eo.logger.Info("Strategy parameters updated",
				zap.String("strategy", strategyName))
			return nil
		}
	}

	return fmt.Errorf("strategy not found: %s", strategyName)
}
