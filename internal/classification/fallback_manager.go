package classification

import (
	"context"
	"fmt"

	"github.com/pcraw4d/business-verification/internal/shared"
)

// NewFallbackManager creates a new fallback manager
func NewFallbackManager(logger shared.Logger) *FallbackManager {
	manager := &FallbackManager{
		fallbackStrategies: make(map[string]*FallbackStrategy),
		logger:             logger,
	}

	// Initialize default fallback strategies
	manager.initializeDefaultStrategies()

	return manager
}

// initializeDefaultStrategies sets up default fallback strategies
func (fm *FallbackManager) initializeDefaultStrategies() {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	// Budget exceeded strategy
	fm.fallbackStrategies["budget_exceeded"] = &FallbackStrategy{
		Name:             "Budget Exceeded Fallback",
		Description:      "Use only free methods when budget is exceeded",
		TriggerCondition: "budget_exceeded",
		CostReduction:    0.80,  // 80% cost reduction
		AccuracyImpact:   -0.15, // 15% accuracy reduction
		MethodSubstitutions: map[string]string{
			"ml":           "keyword",     // Replace ML with keyword
			"external_api": "description", // Replace external API with description
		},
		Enabled: true,
	}

	// High cost strategy
	fm.fallbackStrategies["high_cost"] = &FallbackStrategy{
		Name:             "High Cost Fallback",
		Description:      "Use cheaper methods when cost is too high",
		TriggerCondition: "high_cost",
		CostReduction:    0.60,  // 60% cost reduction
		AccuracyImpact:   -0.10, // 10% accuracy reduction
		MethodSubstitutions: map[string]string{
			"external_api": "ml", // Replace external API with ML
		},
		Enabled: true,
	}

	// Performance degradation strategy
	fm.fallbackStrategies["performance_degradation"] = &FallbackStrategy{
		Name:             "Performance Degradation Fallback",
		Description:      "Use faster methods when performance is degraded",
		TriggerCondition: "performance_degradation",
		CostReduction:    0.30,  // 30% cost reduction
		AccuracyImpact:   -0.05, // 5% accuracy reduction
		MethodSubstitutions: map[string]string{
			"ml": "keyword", // Replace ML with keyword for speed
		},
		Enabled: true,
	}

	// Method failure strategy
	fm.fallbackStrategies["method_failure"] = &FallbackStrategy{
		Name:             "Method Failure Fallback",
		Description:      "Use alternative methods when primary methods fail",
		TriggerCondition: "method_failure",
		CostReduction:    0.20,  // 20% cost reduction
		AccuracyImpact:   -0.08, // 8% accuracy reduction
		MethodSubstitutions: map[string]string{
			"ml":           "keyword",     // Replace failed ML with keyword
			"external_api": "description", // Replace failed external API with description
		},
		Enabled: true,
	}

	// Emergency mode strategy
	fm.fallbackStrategies["emergency_mode"] = &FallbackStrategy{
		Name:             "Emergency Mode Fallback",
		Description:      "Use only essential methods in emergency situations",
		TriggerCondition: "emergency_mode",
		CostReduction:    0.90,  // 90% cost reduction
		AccuracyImpact:   -0.25, // 25% accuracy reduction
		MethodSubstitutions: map[string]string{
			"ml":           "keyword", // Replace ML with keyword
			"external_api": "keyword", // Replace external API with keyword
			"description":  "keyword", // Replace description with keyword
		},
		Enabled: true,
	}

	fm.logger.Log(context.Background(), shared.LogLevelInfo, "Initialized fallback strategies", map[string]interface{}{
		"count": len(fm.fallbackStrategies),
	})
}

// GetFallbackStrategy returns a fallback strategy by name
func (fm *FallbackManager) GetFallbackStrategy(name string) *FallbackStrategy {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	strategy, exists := fm.fallbackStrategies[name]
	if !exists || !strategy.Enabled {
		return nil
	}

	// Return a copy to avoid race conditions
	strategyCopy := *strategy
	return &strategyCopy
}

// GetFallbackStrategyByCondition returns a fallback strategy by trigger condition
func (fm *FallbackManager) GetFallbackStrategyByCondition(condition string) *FallbackStrategy {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	for _, strategy := range fm.fallbackStrategies {
		if strategy.TriggerCondition == condition && strategy.Enabled {
			// Return a copy to avoid race conditions
			strategyCopy := *strategy
			return &strategyCopy
		}
	}

	return nil
}

// GetAllFallbackStrategies returns all fallback strategies
func (fm *FallbackManager) GetAllFallbackStrategies() map[string]*FallbackStrategy {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	strategies := make(map[string]*FallbackStrategy)
	for name, strategy := range fm.fallbackStrategies {
		// Create a copy to avoid race conditions
		strategyCopy := *strategy
		strategies[name] = &strategyCopy
	}

	return strategies
}

// AddFallbackStrategy adds a new fallback strategy
func (fm *FallbackManager) AddFallbackStrategy(name string, strategy *FallbackStrategy) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	if name == "" {
		return fmt.Errorf("strategy name cannot be empty")
	}

	if strategy == nil {
		return fmt.Errorf("strategy cannot be nil")
	}

	// Validate strategy
	if err := fm.validateStrategy(strategy); err != nil {
		return fmt.Errorf("invalid strategy: %w", err)
	}

	fm.fallbackStrategies[name] = strategy

	fm.logger.Log(context.Background(), shared.LogLevelInfo, "Added fallback strategy", map[string]interface{}{
		"strategy_name": name,
	})

	return nil
}

// UpdateFallbackStrategy updates an existing fallback strategy
func (fm *FallbackManager) UpdateFallbackStrategy(name string, strategy *FallbackStrategy) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	if _, exists := fm.fallbackStrategies[name]; !exists {
		return fmt.Errorf("strategy %s does not exist", name)
	}

	// Validate strategy
	if err := fm.validateStrategy(strategy); err != nil {
		return fmt.Errorf("invalid strategy: %w", err)
	}

	fm.fallbackStrategies[name] = strategy

	fm.logger.Log(context.Background(), shared.LogLevelInfo, "Updated fallback strategy", map[string]interface{}{
		"strategy_name": name,
	})

	return nil
}

// RemoveFallbackStrategy removes a fallback strategy
func (fm *FallbackManager) RemoveFallbackStrategy(name string) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	if _, exists := fm.fallbackStrategies[name]; !exists {
		return fmt.Errorf("strategy %s does not exist", name)
	}

	delete(fm.fallbackStrategies, name)

	fm.logger.Log(context.Background(), shared.LogLevelInfo, "Removed fallback strategy", map[string]interface{}{
		"strategy_name": name,
	})

	return nil
}

// EnableFallbackStrategy enables a fallback strategy
func (fm *FallbackManager) EnableFallbackStrategy(name string) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	strategy, exists := fm.fallbackStrategies[name]
	if !exists {
		return fmt.Errorf("strategy %s does not exist", name)
	}

	strategy.Enabled = true

	fm.logger.Log(context.Background(), shared.LogLevelInfo, "Enabled fallback strategy", map[string]interface{}{
		"strategy_name": name,
	})

	return nil
}

// DisableFallbackStrategy disables a fallback strategy
func (fm *FallbackManager) DisableFallbackStrategy(name string) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	strategy, exists := fm.fallbackStrategies[name]
	if !exists {
		return fmt.Errorf("strategy %s does not exist", name)
	}

	strategy.Enabled = false

	fm.logger.Log(context.Background(), shared.LogLevelInfo, "Disabled fallback strategy", map[string]interface{}{
		"strategy_name": name,
	})

	return nil
}

// GetApplicableStrategies returns strategies applicable to a given condition
func (fm *FallbackManager) GetApplicableStrategies(condition string) []*FallbackStrategy {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	var applicable []*FallbackStrategy

	for _, strategy := range fm.fallbackStrategies {
		if strategy.TriggerCondition == condition && strategy.Enabled {
			// Create a copy to avoid race conditions
			strategyCopy := *strategy
			applicable = append(applicable, &strategyCopy)
		}
	}

	return applicable
}

// GetBestFallbackStrategy returns the best fallback strategy for a given condition
func (fm *FallbackManager) GetBestFallbackStrategy(condition string, prioritizeAccuracy bool) *FallbackStrategy {
	strategies := fm.GetApplicableStrategies(condition)

	if len(strategies) == 0 {
		return nil
	}

	// If only one strategy, return it
	if len(strategies) == 1 {
		return strategies[0]
	}

	// Select best strategy based on priority
	bestStrategy := strategies[0]

	for _, strategy := range strategies[1:] {
		if prioritizeAccuracy {
			// Prioritize accuracy (lower accuracy impact is better)
			if strategy.AccuracyImpact > bestStrategy.AccuracyImpact {
				bestStrategy = strategy
			}
		} else {
			// Prioritize cost reduction (higher cost reduction is better)
			if strategy.CostReduction > bestStrategy.CostReduction {
				bestStrategy = strategy
			}
		}
	}

	return bestStrategy
}

// validateStrategy validates a fallback strategy
func (fm *FallbackManager) validateStrategy(strategy *FallbackStrategy) error {
	if strategy.Name == "" {
		return fmt.Errorf("strategy name cannot be empty")
	}

	if strategy.Description == "" {
		return fmt.Errorf("strategy description cannot be empty")
	}

	if strategy.TriggerCondition == "" {
		return fmt.Errorf("strategy trigger condition cannot be empty")
	}

	if strategy.CostReduction < 0 || strategy.CostReduction > 1 {
		return fmt.Errorf("cost reduction must be between 0 and 1")
	}

	if strategy.AccuracyImpact < -1 || strategy.AccuracyImpact > 1 {
		return fmt.Errorf("accuracy impact must be between -1 and 1")
	}

	return nil
}

// GetFallbackStatistics returns statistics about fallback strategy usage
func (fm *FallbackManager) GetFallbackStatistics() *FallbackStatistics {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	stats := &FallbackStatistics{
		TotalStrategies:       len(fm.fallbackStrategies),
		EnabledStrategies:     0,
		DisabledStrategies:    0,
		StrategiesByCondition: make(map[string]int),
		AverageCostReduction:  0.0,
		AverageAccuracyImpact: 0.0,
	}

	var totalCostReduction float64
	var totalAccuracyImpact float64

	for _, strategy := range fm.fallbackStrategies {
		if strategy.Enabled {
			stats.EnabledStrategies++
		} else {
			stats.DisabledStrategies++
		}

		stats.StrategiesByCondition[strategy.TriggerCondition]++
		totalCostReduction += strategy.CostReduction
		totalAccuracyImpact += strategy.AccuracyImpact
	}

	if stats.TotalStrategies > 0 {
		stats.AverageCostReduction = totalCostReduction / float64(stats.TotalStrategies)
		stats.AverageAccuracyImpact = totalAccuracyImpact / float64(stats.TotalStrategies)
	}

	return stats
}

// FallbackStatistics represents statistics about fallback strategies
type FallbackStatistics struct {
	TotalStrategies       int            `json:"total_strategies"`
	EnabledStrategies     int            `json:"enabled_strategies"`
	DisabledStrategies    int            `json:"disabled_strategies"`
	StrategiesByCondition map[string]int `json:"strategies_by_condition"`
	AverageCostReduction  float64        `json:"average_cost_reduction"`
	AverageAccuracyImpact float64        `json:"average_accuracy_impact"`
}

// TestFallbackStrategy tests a fallback strategy with sample data
func (fm *FallbackManager) TestFallbackStrategy(name string, originalMethods []string) (*FallbackTestResult, error) {
	strategy := fm.GetFallbackStrategy(name)
	if strategy == nil {
		return nil, fmt.Errorf("strategy %s not found or disabled", name)
	}

	result := &FallbackTestResult{
		StrategyName:        strategy.Name,
		OriginalMethods:     originalMethods,
		FallbackMethods:     make([]string, 0, len(originalMethods)),
		MethodSubstitutions: make(map[string]string),
		CostReduction:       strategy.CostReduction,
		AccuracyImpact:      strategy.AccuracyImpact,
		TestPassed:          true,
	}

	// Apply method substitutions
	for _, originalMethod := range originalMethods {
		if fallbackMethod, exists := strategy.MethodSubstitutions[originalMethod]; exists {
			result.FallbackMethods = append(result.FallbackMethods, fallbackMethod)
			result.MethodSubstitutions[originalMethod] = fallbackMethod
		} else {
			result.FallbackMethods = append(result.FallbackMethods, originalMethod)
		}
	}

	// Validate that we have at least one method
	if len(result.FallbackMethods) == 0 {
		result.TestPassed = false
		result.ErrorMessage = "no methods available after fallback"
	}

	return result, nil
}

// FallbackTestResult represents the result of testing a fallback strategy
type FallbackTestResult struct {
	StrategyName        string            `json:"strategy_name"`
	OriginalMethods     []string          `json:"original_methods"`
	FallbackMethods     []string          `json:"fallback_methods"`
	MethodSubstitutions map[string]string `json:"method_substitutions"`
	CostReduction       float64           `json:"cost_reduction"`
	AccuracyImpact      float64           `json:"accuracy_impact"`
	TestPassed          bool              `json:"test_passed"`
	ErrorMessage        string            `json:"error_message,omitempty"`
}
