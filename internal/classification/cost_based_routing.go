package classification

import (
	"context"
	"fmt"
	"sync"
	"time"

	"kyb-platform/internal/shared"
)

// CustomerTier represents different customer service tiers
type CustomerTier string

const (
	// Free tier - basic classification with minimal cost
	CustomerTierFree CustomerTier = "free"

	// Standard tier - balanced accuracy and cost
	CustomerTierStandard CustomerTier = "standard"

	// Premium tier - high accuracy with premium methods
	CustomerTierPremium CustomerTier = "premium"

	// Enterprise tier - maximum accuracy and features
	CustomerTierEnterprise CustomerTier = "enterprise"
)

// TierConfig defines configuration for each customer tier
type TierConfig struct {
	Tier        CustomerTier `json:"tier"`
	Name        string       `json:"name"`
	Description string       `json:"description"`

	// Cost constraints
	MaxCostPerCall     float64 `json:"max_cost_per_call"`    // Maximum cost per classification call
	MonthlyBudgetLimit float64 `json:"monthly_budget_limit"` // Monthly budget limit
	CostOptimization   bool    `json:"cost_optimization"`    // Enable aggressive cost optimization

	// Method constraints
	AllowedMethods  []string `json:"allowed_methods"`  // Methods allowed for this tier
	RequiredMethods []string `json:"required_methods"` // Methods that must be used
	ExcludedMethods []string `json:"excluded_methods"` // Methods excluded for this tier

	// Quality constraints
	MinAccuracyTarget   float64 `json:"min_accuracy_target"`   // Minimum accuracy target
	MinConfidenceTarget float64 `json:"min_confidence_target"` // Minimum confidence target

	// Performance constraints
	MaxResponseTime time.Duration `json:"max_response_time"` // Maximum response time
	CachePriority   bool          `json:"cache_priority"`    // Prioritize cached results

	// Features
	AdvancedFeatures   bool `json:"advanced_features"`    // Enable advanced features
	RealTimeMonitoring bool `json:"real_time_monitoring"` // Enable real-time monitoring
	CustomWeighting    bool `json:"custom_weighting"`     // Allow custom method weighting
}

// CostBasedRouter manages cost-based routing for different customer tiers
type CostBasedRouter struct {
	tierConfigs     map[CustomerTier]*TierConfig
	costTracker     *CostTracker
	fallbackManager *FallbackManager
	registry        CostRoutingMethodRegistry
	logger          shared.Logger
	mutex           sync.RWMutex
}

// CostTracker tracks costs and budgets for different tiers
type CostTracker struct {
	tierBudgets  map[CustomerTier]*TierBudget
	globalBudget *GlobalBudget
	logger       shared.Logger
	mutex        sync.RWMutex
}

// TierBudget tracks budget usage for a specific tier
type TierBudget struct {
	Tier               CustomerTier `json:"tier"`
	MonthlyBudget      float64      `json:"monthly_budget"`
	UsedBudget         float64      `json:"used_budget"`
	RemainingBudget    float64      `json:"remaining_budget"`
	CallCount          int64        `json:"call_count"`
	AverageCostPerCall float64      `json:"average_cost_per_call"`
	LastReset          time.Time    `json:"last_reset"`
	NextReset          time.Time    `json:"next_reset"`
}

// GlobalBudget tracks global cost constraints
type GlobalBudget struct {
	DailyBudgetLimit   float64 `json:"daily_budget_limit"`
	UsedDailyBudget    float64 `json:"used_daily_budget"`
	MonthlyBudgetLimit float64 `json:"monthly_budget_limit"`
	UsedMonthlyBudget  float64 `json:"used_monthly_budget"`
	EmergencyThreshold float64 `json:"emergency_threshold"` // Threshold for emergency cost controls
}

// FallbackManager manages fallback strategies for cost optimization
type FallbackManager struct {
	fallbackStrategies map[string]*FallbackStrategy
	logger             shared.Logger
	mutex              sync.RWMutex
}

// FallbackStrategy defines a fallback strategy for cost optimization
type FallbackStrategy struct {
	Name                string            `json:"name"`
	Description         string            `json:"description"`
	TriggerCondition    string            `json:"trigger_condition"`    // When to trigger this strategy
	CostReduction       float64           `json:"cost_reduction"`       // Expected cost reduction percentage
	AccuracyImpact      float64           `json:"accuracy_impact"`      // Expected accuracy impact
	MethodSubstitutions map[string]string `json:"method_substitutions"` // Method substitutions
	Enabled             bool              `json:"enabled"`
}

// RoutingDecision represents a routing decision for a classification request
type RoutingDecision struct {
	Tier             CustomerTier           `json:"tier"`
	SelectedMethods  []string               `json:"selected_methods"`
	MethodWeights    map[string]float64     `json:"method_weights"`
	CostEstimate     float64                `json:"cost_estimate"`
	ExpectedAccuracy float64                `json:"expected_accuracy"`
	FallbackStrategy *FallbackStrategy      `json:"fallback_strategy,omitempty"`
	Reasoning        string                 `json:"reasoning"`
	Constraints      map[string]interface{} `json:"constraints"`
}

// NewCostBasedRouter creates a new cost-based router
func NewCostBasedRouter(registry CostRoutingMethodRegistry, logger shared.Logger) *CostBasedRouter {
	router := &CostBasedRouter{
		tierConfigs:     make(map[CustomerTier]*TierConfig),
		registry:        registry,
		logger:          logger,
		costTracker:     NewCostTracker(logger),
		fallbackManager: NewFallbackManager(logger),
	}

	// Initialize default tier configurations
	router.initializeDefaultTierConfigs()

	return router
}

// initializeDefaultTierConfigs sets up default tier configurations
func (cbr *CostBasedRouter) initializeDefaultTierConfigs() {
	cbr.mutex.Lock()
	defer cbr.mutex.Unlock()

	// Free tier - minimal cost, basic accuracy
	cbr.tierConfigs[CustomerTierFree] = &TierConfig{
		Tier:                CustomerTierFree,
		Name:                "Free Tier",
		Description:         "Basic classification with minimal cost",
		MaxCostPerCall:      0.001, // $0.001 per call
		MonthlyBudgetLimit:  10.0,  // $10/month
		CostOptimization:    true,
		AllowedMethods:      []string{"keyword", "description"},
		RequiredMethods:     []string{"keyword"},
		ExcludedMethods:     []string{"ml", "external_api"},
		MinAccuracyTarget:   0.70, // 70% accuracy
		MinConfidenceTarget: 0.60, // 60% confidence
		MaxResponseTime:     2 * time.Second,
		CachePriority:       true,
		AdvancedFeatures:    false,
		RealTimeMonitoring:  false,
		CustomWeighting:     false,
	}

	// Standard tier - balanced cost and accuracy
	cbr.tierConfigs[CustomerTierStandard] = &TierConfig{
		Tier:                CustomerTierStandard,
		Name:                "Standard Tier",
		Description:         "Balanced accuracy and cost",
		MaxCostPerCall:      0.005, // $0.005 per call
		MonthlyBudgetLimit:  50.0,  // $50/month
		CostOptimization:    true,
		AllowedMethods:      []string{"keyword", "ml", "description"},
		RequiredMethods:     []string{"keyword", "ml"},
		ExcludedMethods:     []string{"external_api"},
		MinAccuracyTarget:   0.85, // 85% accuracy
		MinConfidenceTarget: 0.75, // 75% confidence
		MaxResponseTime:     3 * time.Second,
		CachePriority:       true,
		AdvancedFeatures:    false,
		RealTimeMonitoring:  false,
		CustomWeighting:     false,
	}

	// Premium tier - high accuracy with some premium methods
	cbr.tierConfigs[CustomerTierPremium] = &TierConfig{
		Tier:                CustomerTierPremium,
		Name:                "Premium Tier",
		Description:         "High accuracy with premium methods",
		MaxCostPerCall:      0.020, // $0.020 per call
		MonthlyBudgetLimit:  200.0, // $200/month
		CostOptimization:    false,
		AllowedMethods:      []string{"keyword", "ml", "external_api", "description"},
		RequiredMethods:     []string{"keyword", "ml"},
		ExcludedMethods:     []string{},
		MinAccuracyTarget:   0.90, // 90% accuracy
		MinConfidenceTarget: 0.80, // 80% confidence
		MaxResponseTime:     5 * time.Second,
		CachePriority:       false,
		AdvancedFeatures:    true,
		RealTimeMonitoring:  true,
		CustomWeighting:     true,
	}

	// Enterprise tier - maximum accuracy and features
	cbr.tierConfigs[CustomerTierEnterprise] = &TierConfig{
		Tier:                CustomerTierEnterprise,
		Name:                "Enterprise Tier",
		Description:         "Maximum accuracy and features",
		MaxCostPerCall:      0.050,  // $0.050 per call
		MonthlyBudgetLimit:  1000.0, // $1000/month
		CostOptimization:    false,
		AllowedMethods:      []string{"keyword", "ml", "external_api", "description"},
		RequiredMethods:     []string{"keyword", "ml", "external_api"},
		ExcludedMethods:     []string{},
		MinAccuracyTarget:   0.95, // 95% accuracy
		MinConfidenceTarget: 0.85, // 85% confidence
		MaxResponseTime:     10 * time.Second,
		CachePriority:       false,
		AdvancedFeatures:    true,
		RealTimeMonitoring:  true,
		CustomWeighting:     true,
	}
}

// RouteRequest determines the optimal routing for a classification request
func (cbr *CostBasedRouter) RouteRequest(ctx context.Context, tier CustomerTier, businessName, description, websiteURL string) (*RoutingDecision, error) {
	cbr.mutex.RLock()
	config, exists := cbr.tierConfigs[tier]
	cbr.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown customer tier: %s", tier)
	}

	// Check budget constraints
	if err := cbr.costTracker.CheckBudgetConstraints(tier, config); err != nil {
		// Try fallback strategy
		fallbackStrategy := cbr.fallbackManager.GetFallbackStrategy("budget_exceeded")
		if fallbackStrategy != nil {
			cbr.logger.Log(context.Background(), shared.LogLevelWarning, "Budget exceeded, using fallback strategy", map[string]interface{}{
				"tier":              tier,
				"fallback_strategy": fallbackStrategy.Name,
			})
			return cbr.createFallbackRoutingDecision(config, fallbackStrategy, businessName, description, websiteURL)
		}
		return nil, fmt.Errorf("budget constraints exceeded: %w", err)
	}

	// Get available methods for this tier
	availableMethods := cbr.getAvailableMethods(config)
	if len(availableMethods) == 0 {
		return nil, fmt.Errorf("no available methods for tier %s", tier)
	}

	// Calculate optimal method selection and weights
	selectedMethods, methodWeights, err := cbr.calculateOptimalMethods(ctx, config, availableMethods, businessName, description, websiteURL)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate optimal methods: %w", err)
	}

	// Estimate cost
	costEstimate := cbr.estimateCost(selectedMethods, methodWeights)

	// Estimate accuracy
	expectedAccuracy := cbr.estimateAccuracy(selectedMethods, methodWeights)

	// Create routing decision
	decision := &RoutingDecision{
		Tier:             tier,
		SelectedMethods:  selectedMethods,
		MethodWeights:    methodWeights,
		CostEstimate:     costEstimate,
		ExpectedAccuracy: expectedAccuracy,
		Reasoning:        cbr.generateReasoning(config, selectedMethods, costEstimate, expectedAccuracy),
		Constraints: map[string]interface{}{
			"max_cost_per_call":   config.MaxCostPerCall,
			"min_accuracy_target": config.MinAccuracyTarget,
			"max_response_time":   config.MaxResponseTime,
			"cost_optimization":   config.CostOptimization,
		},
	}

	cbr.logger.Log(context.Background(), shared.LogLevelInfo, "Routing decision", map[string]interface{}{
		"tier":     tier,
		"methods":  selectedMethods,
		"cost":     costEstimate,
		"accuracy": expectedAccuracy,
	})

	return decision, nil
}

// getAvailableMethods returns methods available for the given tier configuration
func (cbr *CostBasedRouter) getAvailableMethods(config *TierConfig) []string {
	var availableMethods []string

	// Get all registered methods
	allMethods := cbr.registry.GetAllMethods()

	for _, method := range allMethods {
		methodName := method.GetName()
		methodType := method.GetType()

		// Check if method is allowed
		if !cbr.isMethodAllowed(methodName, methodType, config) {
			continue
		}

		// Check if method is enabled
		if !method.IsEnabled() {
			continue
		}

		availableMethods = append(availableMethods, methodName)
	}

	return availableMethods
}

// isMethodAllowed checks if a method is allowed for the given tier configuration
func (cbr *CostBasedRouter) isMethodAllowed(methodName, methodType string, config *TierConfig) bool {
	// Check if method is explicitly excluded
	for _, excluded := range config.ExcludedMethods {
		if methodName == excluded || methodType == excluded {
			return false
		}
	}

	// Check if method is in allowed list
	for _, allowed := range config.AllowedMethods {
		if methodName == allowed || methodType == allowed {
			return true
		}
	}

	// If no allowed methods specified, allow all non-excluded methods
	return len(config.AllowedMethods) == 0
}

// calculateOptimalMethods calculates the optimal method selection and weights
func (cbr *CostBasedRouter) calculateOptimalMethods(ctx context.Context, config *TierConfig, availableMethods []string, businessName, description, websiteURL string) ([]string, map[string]float64, error) {
	// Start with required methods
	selectedMethods := make([]string, 0, len(config.RequiredMethods))
	methodWeights := make(map[string]float64)

	// Add required methods
	for _, required := range config.RequiredMethods {
		if cbr.isMethodAvailable(required, availableMethods) {
			selectedMethods = append(selectedMethods, required)
			methodWeights[required] = 0.0 // Will be calculated later
		}
	}

	// Add additional methods based on tier and cost optimization
	if config.CostOptimization {
		// For cost-optimized tiers, prefer cheaper methods
		selectedMethods = cbr.selectCostOptimizedMethods(availableMethods, selectedMethods, config)
	} else {
		// For premium tiers, use all available methods
		for _, method := range availableMethods {
			if !cbr.containsMethod(selectedMethods, method) {
				selectedMethods = append(selectedMethods, method)
			}
		}
	}

	// Calculate weights based on method performance and tier requirements
	weights, err := cbr.calculateMethodWeights(selectedMethods, config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to calculate method weights: %w", err)
	}

	// Normalize weights
	normalizedWeights := cbr.normalizeWeights(weights)

	return selectedMethods, normalizedWeights, nil
}

// selectCostOptimizedMethods selects methods optimized for cost
func (cbr *CostBasedRouter) selectCostOptimizedMethods(availableMethods, requiredMethods []string, config *TierConfig) []string {
	// Method cost rankings (lower is cheaper)
	methodCosts := map[string]float64{
		"keyword":      0.001, // Cheapest
		"description":  0.002, // Cheap
		"ml":           0.005, // Moderate
		"external_api": 0.020, // Expensive
	}

	selectedMethods := make([]string, len(requiredMethods))
	copy(selectedMethods, requiredMethods)

	// Add cheapest available methods first
	for _, method := range availableMethods {
		if cbr.containsMethod(selectedMethods, method) {
			continue
		}

		cost := methodCosts[method]
		if cost == 0 {
			cost = 0.010 // Default cost for unknown methods
		}

		// Only add if within cost constraints
		if cost <= config.MaxCostPerCall {
			selectedMethods = append(selectedMethods, method)
		}
	}

	return selectedMethods
}

// calculateMethodWeights calculates weights for selected methods
func (cbr *CostBasedRouter) calculateMethodWeights(methods []string, config *TierConfig) (map[string]float64, error) {
	weights := make(map[string]float64)

	// Base weights by method type
	baseWeights := map[string]float64{
		"keyword":      0.5, // Primary method
		"ml":           0.4, // Secondary method
		"description":  0.1, // Tertiary method
		"external_api": 0.3, // External validation
	}

	// Adjust weights based on tier
	weightMultipliers := map[CustomerTier]map[string]float64{
		CustomerTierFree: {
			"keyword":      1.0, // Emphasize keyword
			"ml":           0.0, // Disable ML for free tier
			"description":  1.0, // Use description
			"external_api": 0.0, // Disable external APIs
		},
		CustomerTierStandard: {
			"keyword":      1.0, // Emphasize keyword
			"ml":           1.0, // Use ML
			"description":  0.5, // Reduce description weight
			"external_api": 0.0, // Disable external APIs
		},
		CustomerTierPremium: {
			"keyword":      1.0, // Use keyword
			"ml":           1.0, // Use ML
			"description":  0.3, // Reduce description weight
			"external_api": 1.0, // Use external APIs
		},
		CustomerTierEnterprise: {
			"keyword":      1.0, // Use keyword
			"ml":           1.0, // Use ML
			"description":  0.2, // Minimal description weight
			"external_api": 1.0, // Use external APIs
		},
	}

	// Calculate weights
	for _, method := range methods {
		baseWeight := baseWeights[method]
		if baseWeight == 0 {
			baseWeight = 0.1 // Default weight
		}

		multiplier := weightMultipliers[config.Tier][method]
		if multiplier == 0 {
			multiplier = 0.1 // Default multiplier
		}

		weights[method] = baseWeight * multiplier
	}

	return weights, nil
}

// normalizeWeights normalizes weights so they sum to 1.0
func (cbr *CostBasedRouter) normalizeWeights(weights map[string]float64) map[string]float64 {
	var totalWeight float64
	for _, weight := range weights {
		totalWeight += weight
	}

	if totalWeight == 0 {
		// If all weights are zero, set equal weights
		equalWeight := 1.0 / float64(len(weights))
		for method := range weights {
			weights[method] = equalWeight
		}
		return weights
	}

	// Normalize weights
	normalized := make(map[string]float64)
	for method, weight := range weights {
		normalized[method] = weight / totalWeight
	}

	return normalized
}

// estimateCost estimates the cost of using the selected methods
func (cbr *CostBasedRouter) estimateCost(methods []string, weights map[string]float64) float64 {
	// Method cost estimates (per call)
	methodCosts := map[string]float64{
		"keyword":      0.001, // $0.001 per call
		"description":  0.002, // $0.002 per call
		"ml":           0.005, // $0.005 per call
		"external_api": 0.020, // $0.020 per call
	}

	var totalCost float64
	for _, method := range methods {
		cost := methodCosts[method]
		if cost == 0 {
			cost = 0.010 // Default cost
		}

		weight := weights[method]
		if weight == 0 {
			weight = 1.0 / float64(len(methods)) // Equal weight if not specified
		}

		totalCost += cost * weight
	}

	return totalCost
}

// estimateAccuracy estimates the expected accuracy of the selected methods
func (cbr *CostBasedRouter) estimateAccuracy(methods []string, weights map[string]float64) float64 {
	// Method accuracy estimates
	methodAccuracies := map[string]float64{
		"keyword":      0.75, // 75% accuracy
		"description":  0.60, // 60% accuracy
		"ml":           0.85, // 85% accuracy
		"external_api": 0.90, // 90% accuracy
	}

	var weightedAccuracy float64
	var totalWeight float64

	for _, method := range methods {
		accuracy := methodAccuracies[method]
		if accuracy == 0 {
			accuracy = 0.70 // Default accuracy
		}

		weight := weights[method]
		if weight == 0 {
			weight = 1.0 / float64(len(methods)) // Equal weight if not specified
		}

		weightedAccuracy += accuracy * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.70 // Default accuracy
	}

	return weightedAccuracy / totalWeight
}

// generateReasoning generates human-readable reasoning for the routing decision
func (cbr *CostBasedRouter) generateReasoning(config *TierConfig, methods []string, cost, accuracy float64) string {
	reasoning := fmt.Sprintf("Selected %d methods for %s tier: %v. ", len(methods), config.Name, methods)
	reasoning += fmt.Sprintf("Estimated cost: $%.4f (limit: $%.4f). ", cost, config.MaxCostPerCall)
	reasoning += fmt.Sprintf("Expected accuracy: %.1f%% (target: %.1f%%). ", accuracy*100, config.MinAccuracyTarget*100)

	if config.CostOptimization {
		reasoning += "Cost optimization enabled - prioritizing efficient methods. "
	}

	if config.CachePriority {
		reasoning += "Cache priority enabled - will use cached results when available. "
	}

	return reasoning
}

// Helper functions
func (cbr *CostBasedRouter) isMethodAvailable(method string, availableMethods []string) bool {
	for _, available := range availableMethods {
		if available == method {
			return true
		}
	}
	return false
}

func (cbr *CostBasedRouter) containsMethod(methods []string, method string) bool {
	for _, m := range methods {
		if m == method {
			return true
		}
	}
	return false
}

// createFallbackRoutingDecision creates a routing decision using fallback strategy
func (cbr *CostBasedRouter) createFallbackRoutingDecision(config *TierConfig, strategy *FallbackStrategy, businessName, description, websiteURL string) (*RoutingDecision, error) {
	// Use fallback method substitutions
	selectedMethods := make([]string, 0, len(strategy.MethodSubstitutions))
	methodWeights := make(map[string]float64)

	for _, fallbackMethod := range strategy.MethodSubstitutions {
		selectedMethods = append(selectedMethods, fallbackMethod)
		methodWeights[fallbackMethod] = 0.5 // Default weight for fallback methods
	}

	// If no substitutions, use minimal methods
	if len(selectedMethods) == 0 {
		selectedMethods = []string{"keyword"}
		methodWeights["keyword"] = 1.0
	}

	// Normalize weights
	normalizedWeights := cbr.normalizeWeights(methodWeights)

	// Estimate cost and accuracy
	costEstimate := cbr.estimateCost(selectedMethods, normalizedWeights)
	expectedAccuracy := cbr.estimateAccuracy(selectedMethods, normalizedWeights)

	return &RoutingDecision{
		Tier:             config.Tier,
		SelectedMethods:  selectedMethods,
		MethodWeights:    normalizedWeights,
		CostEstimate:     costEstimate,
		ExpectedAccuracy: expectedAccuracy,
		FallbackStrategy: strategy,
		Reasoning:        fmt.Sprintf("Using fallback strategy '%s' due to budget constraints. %s", strategy.Name, strategy.Description),
		Constraints: map[string]interface{}{
			"fallback_strategy": strategy.Name,
			"cost_reduction":    strategy.CostReduction,
			"accuracy_impact":   strategy.AccuracyImpact,
		},
	}, nil
}

// GetTierConfig returns the configuration for a specific tier
func (cbr *CostBasedRouter) GetTierConfig(tier CustomerTier) (*TierConfig, error) {
	cbr.mutex.RLock()
	defer cbr.mutex.RUnlock()

	config, exists := cbr.tierConfigs[tier]
	if !exists {
		return nil, fmt.Errorf("unknown customer tier: %s", tier)
	}

	return config, nil
}

// UpdateTierConfig updates the configuration for a specific tier
func (cbr *CostBasedRouter) UpdateTierConfig(tier CustomerTier, config *TierConfig) error {
	cbr.mutex.Lock()
	defer cbr.mutex.Unlock()

	config.Tier = tier
	cbr.tierConfigs[tier] = config

	cbr.logger.Log(context.Background(), shared.LogLevelInfo, "Updated tier configuration", map[string]interface{}{
		"tier": tier,
	})
	return nil
}

// GetAllTierConfigs returns all tier configurations
func (cbr *CostBasedRouter) GetAllTierConfigs() map[CustomerTier]*TierConfig {
	cbr.mutex.RLock()
	defer cbr.mutex.RUnlock()

	configs := make(map[CustomerTier]*TierConfig)
	for tier, config := range cbr.tierConfigs {
		configs[tier] = config
	}

	return configs
}
