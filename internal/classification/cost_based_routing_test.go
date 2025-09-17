package classification

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/shared"
)

// MockMethodRegistry is a mock implementation of MethodRegistry for testing
type MockMethodRegistry struct {
	methods map[string]*MockMethod
}

type MockMethod struct {
	name        string
	methodType  string
	enabled     bool
	weight      float64
	description string
}

func (m *MockMethod) GetName() string          { return m.name }
func (m *MockMethod) GetType() string          { return m.methodType }
func (m *MockMethod) GetDescription() string   { return m.description }
func (m *MockMethod) GetWeight() float64       { return m.weight }
func (m *MockMethod) SetWeight(weight float64) { m.weight = weight }
func (m *MockMethod) IsEnabled() bool          { return m.enabled }
func (m *MockMethod) SetEnabled(enabled bool)  { m.enabled = enabled }
func (m *MockMethod) Classify(ctx context.Context, businessName, description, websiteURL string) (*shared.ClassificationMethodResult, error) {
	return &shared.ClassificationMethodResult{
		Success:        true,
		MethodType:     m.methodType,
		Confidence:     0.8,
		ProcessingTime: time.Second,
	}, nil
}
func (m *MockMethod) GetPerformanceMetrics() interface{}                               { return nil }
func (m *MockMethod) ValidateInput(businessName, description, websiteURL string) error { return nil }
func (m *MockMethod) GetRequiredDependencies() []string                                { return nil }
func (m *MockMethod) Initialize(ctx context.Context) error                             { return nil }
func (m *MockMethod) Cleanup() error                                                   { return nil }

func (mr *MockMethodRegistry) GetAllMethods() []ClassificationMethod {
	var methods []ClassificationMethod
	for _, method := range mr.methods {
		methods = append(methods, method)
	}
	return methods
}

func (mr *MockMethodRegistry) GetMethod(name string) (ClassificationMethod, error) {
	if method, exists := mr.methods[name]; exists {
		return method, nil
	}
	return nil, fmt.Errorf("method %s not found", name)
}

func (mr *MockMethodRegistry) RegisterMethod(method ClassificationMethod, config MethodConfig) error {
	mr.methods[method.GetName()] = method.(*MockMethod)
	return nil
}

func (mr *MockMethodRegistry) UnregisterMethod(name string) error {
	delete(mr.methods, name)
	return nil
}

func (mr *MockMethodRegistry) UpdateMethodConfig(name string, config MethodConfig) error {
	return nil
}

// MockLogger is a mock implementation of shared.Logger for testing
type MockLogger struct{}

func (ml *MockLogger) Log(ctx context.Context, level shared.LogLevel, message string, fields map[string]interface{}) error {
	return nil
}

func (ml *MockLogger) LogClassification(ctx context.Context, req *shared.BusinessClassificationRequest, resp *shared.BusinessClassificationResponse, err error) error {
	return nil
}

func (ml *MockLogger) HealthCheck(ctx context.Context) error {
	return nil
}

func TestCostBasedRouter_InitializeDefaultTierConfigs(t *testing.T) {
	registry := &MockMethodRegistry{methods: make(map[string]*MockMethod)}
	logger := &MockLogger{}
	router := NewCostBasedRouter(registry, logger)

	// Test that all default tiers are initialized
	expectedTiers := []CustomerTier{
		CustomerTierFree,
		CustomerTierStandard,
		CustomerTierPremium,
		CustomerTierEnterprise,
	}

	for _, tier := range expectedTiers {
		config, err := router.GetTierConfig(tier)
		if err != nil {
			t.Errorf("Expected tier %s to be initialized, got error: %v", tier, err)
		}

		if config.Tier != tier {
			t.Errorf("Expected tier %s, got %s", tier, config.Tier)
		}
	}

	// Test specific tier configurations
	freeConfig, _ := router.GetTierConfig(CustomerTierFree)
	if freeConfig.MaxCostPerCall != 0.001 {
		t.Errorf("Expected free tier max cost per call to be 0.001, got %f", freeConfig.MaxCostPerCall)
	}

	if freeConfig.MonthlyBudgetLimit != 10.0 {
		t.Errorf("Expected free tier monthly budget to be 10.0, got %f", freeConfig.MonthlyBudgetLimit)
	}

	enterpriseConfig, _ := router.GetTierConfig(CustomerTierEnterprise)
	if enterpriseConfig.MaxCostPerCall != 0.050 {
		t.Errorf("Expected enterprise tier max cost per call to be 0.050, got %f", enterpriseConfig.MaxCostPerCall)
	}

	if enterpriseConfig.MonthlyBudgetLimit != 1000.0 {
		t.Errorf("Expected enterprise tier monthly budget to be 1000.0, got %f", enterpriseConfig.MonthlyBudgetLimit)
	}
}

func TestCostBasedRouter_RouteRequest_FreeTier(t *testing.T) {
	// Setup mock registry with methods
	registry := &MockMethodRegistry{methods: make(map[string]*MockMethod)}
	registry.methods["keyword"] = &MockMethod{
		name: "keyword", methodType: "keyword", enabled: true, weight: 0.5,
		description: "Keyword-based classification",
	}
	registry.methods["description"] = &MockMethod{
		name: "description", methodType: "description", enabled: true, weight: 0.1,
		description: "Description-based classification",
	}
	registry.methods["ml"] = &MockMethod{
		name: "ml", methodType: "ml", enabled: true, weight: 0.4,
		description: "Machine learning classification",
	}

	logger := &MockLogger{}
	router := NewCostBasedRouter(registry, logger)

	// Test free tier routing
	decision, err := router.RouteRequest(context.Background(), CustomerTierFree, "Test Business", "A test business", "https://test.com")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Free tier should only use keyword and description methods
	expectedMethods := []string{"keyword", "description"}
	if len(decision.SelectedMethods) != len(expectedMethods) {
		t.Errorf("Expected %d methods for free tier, got %d", len(expectedMethods), len(decision.SelectedMethods))
	}

	// Check that ML method is not included (excluded for free tier)
	for _, method := range decision.SelectedMethods {
		if method == "ml" {
			t.Errorf("ML method should not be included in free tier")
		}
	}

	// Check cost estimate is within limits
	if decision.CostEstimate > 0.001 {
		t.Errorf("Expected cost estimate to be <= 0.001 for free tier, got %f", decision.CostEstimate)
	}

	// Check reasoning includes cost optimization
	if decision.Reasoning == "" {
		t.Errorf("Expected reasoning to be provided")
	}
}

func TestCostBasedRouter_RouteRequest_PremiumTier(t *testing.T) {
	// Setup mock registry with all methods
	registry := &MockMethodRegistry{methods: make(map[string]*MockMethod)}
	registry.methods["keyword"] = &MockMethod{
		name: "keyword", methodType: "keyword", enabled: true, weight: 0.5,
		description: "Keyword-based classification",
	}
	registry.methods["ml"] = &MockMethod{
		name: "ml", methodType: "ml", enabled: true, weight: 0.4,
		description: "Machine learning classification",
	}
	registry.methods["external_api"] = &MockMethod{
		name: "external_api", methodType: "external_api", enabled: true, weight: 0.3,
		description: "External API classification",
	}
	registry.methods["description"] = &MockMethod{
		name: "description", methodType: "description", enabled: true, weight: 0.1,
		description: "Description-based classification",
	}

	logger := &MockLogger{}
	router := NewCostBasedRouter(registry, logger)

	// Test premium tier routing
	decision, err := router.RouteRequest(context.Background(), CustomerTierPremium, "Test Business", "A test business", "https://test.com")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Premium tier should use multiple methods including external API
	if len(decision.SelectedMethods) < 2 {
		t.Errorf("Expected premium tier to use multiple methods, got %d", len(decision.SelectedMethods))
	}

	// Check that external API method is included (allowed for premium tier)
	hasExternalAPI := false
	for _, method := range decision.SelectedMethods {
		if method == "external_api" {
			hasExternalAPI = true
			break
		}
	}
	if !hasExternalAPI {
		t.Errorf("Expected premium tier to include external API method")
	}

	// Check cost estimate is within limits
	if decision.CostEstimate > 0.020 {
		t.Errorf("Expected cost estimate to be <= 0.020 for premium tier, got %f", decision.CostEstimate)
	}

	// Check expected accuracy is high
	if decision.ExpectedAccuracy < 0.85 {
		t.Errorf("Expected accuracy to be >= 0.85 for premium tier, got %f", decision.ExpectedAccuracy)
	}
}

func TestCostBasedRouter_GetAvailableMethods(t *testing.T) {
	// Setup mock registry
	registry := &MockMethodRegistry{methods: make(map[string]*MockMethod)}
	registry.methods["keyword"] = &MockMethod{
		name: "keyword", methodType: "keyword", enabled: true, weight: 0.5,
		description: "Keyword-based classification",
	}
	registry.methods["ml"] = &MockMethod{
		name: "ml", methodType: "ml", enabled: true, weight: 0.4,
		description: "Machine learning classification",
	}
	registry.methods["disabled_method"] = &MockMethod{
		name: "disabled_method", methodType: "keyword", enabled: false, weight: 0.5,
		description: "Disabled method",
	}

	logger := &MockLogger{}
	router := NewCostBasedRouter(registry, logger)

	// Test free tier configuration
	freeConfig, _ := router.GetTierConfig(CustomerTierFree)
	availableMethods := router.getAvailableMethods(freeConfig)

	// Free tier should only have keyword and description methods available
	expectedCount := 1 // Only keyword (description not in registry)
	if len(availableMethods) != expectedCount {
		t.Errorf("Expected %d available methods for free tier, got %d", expectedCount, len(availableMethods))
	}

	// Check that disabled method is not included
	for _, method := range availableMethods {
		if method == "disabled_method" {
			t.Errorf("Disabled method should not be available")
		}
	}
}

func TestCostBasedRouter_EstimateCost(t *testing.T) {
	registry := &MockMethodRegistry{methods: make(map[string]*MockMethod)}
	logger := &MockLogger{}
	router := NewCostBasedRouter(registry, logger)

	// Test cost estimation
	methods := []string{"keyword", "ml", "external_api"}
	weights := map[string]float64{
		"keyword":      0.5,
		"ml":           0.3,
		"external_api": 0.2,
	}

	cost := router.estimateCost(methods, weights)

	// Expected cost: (0.001 * 0.5) + (0.005 * 0.3) + (0.020 * 0.2) = 0.0005 + 0.0015 + 0.004 = 0.006
	expectedCost := 0.006
	if cost != expectedCost {
		t.Errorf("Expected cost %f, got %f", expectedCost, cost)
	}
}

func TestCostBasedRouter_EstimateAccuracy(t *testing.T) {
	registry := &MockMethodRegistry{methods: make(map[string]*MockMethod)}
	logger := &MockLogger{}
	router := NewCostBasedRouter(registry, logger)

	// Test accuracy estimation
	methods := []string{"keyword", "ml", "external_api"}
	weights := map[string]float64{
		"keyword":      0.5,
		"ml":           0.3,
		"external_api": 0.2,
	}

	accuracy := router.estimateAccuracy(methods, weights)

	// Expected accuracy: (0.75 * 0.5) + (0.85 * 0.3) + (0.90 * 0.2) = 0.375 + 0.255 + 0.18 = 0.81
	expectedAccuracy := 0.81
	if accuracy != expectedAccuracy {
		t.Errorf("Expected accuracy %f, got %f", expectedAccuracy, accuracy)
	}
}

func TestCostTracker_CheckBudgetConstraints(t *testing.T) {
	logger := &MockLogger{}
	tracker := NewCostTracker(logger)

	// Test free tier budget constraints
	freeConfig := &TierConfig{
		Tier:               CustomerTierFree,
		MaxCostPerCall:     0.001,
		MonthlyBudgetLimit: 10.0,
	}

	// Should pass with fresh budget
	err := tracker.CheckBudgetConstraints(CustomerTierFree, freeConfig)
	if err != nil {
		t.Errorf("Expected no error with fresh budget, got: %v", err)
	}

	// Record some costs
	tracker.RecordCost(CustomerTierFree, 5.0, "keyword")
	tracker.RecordCost(CustomerTierFree, 5.0, "keyword")

	// Should still pass (within budget)
	err = tracker.CheckBudgetConstraints(CustomerTierFree, freeConfig)
	if err != nil {
		t.Errorf("Expected no error within budget, got: %v", err)
	}

	// Record more costs to exceed budget
	tracker.RecordCost(CustomerTierFree, 1.0, "keyword")

	// Should fail (exceeded budget)
	err = tracker.CheckBudgetConstraints(CustomerTierFree, freeConfig)
	if err == nil {
		t.Errorf("Expected error when budget exceeded, got none")
	}
}

func TestCostTracker_RecordCost(t *testing.T) {
	logger := &MockLogger{}
	tracker := NewCostTracker(logger)

	// Record initial cost
	err := tracker.RecordCost(CustomerTierStandard, 0.005, "ml")
	if err != nil {
		t.Errorf("Expected no error recording cost, got: %v", err)
	}

	// Check budget status
	status, err := tracker.GetBudgetStatus(CustomerTierStandard)
	if err != nil {
		t.Errorf("Expected no error getting budget status, got: %v", err)
	}

	if status.UsedBudget != 0.005 {
		t.Errorf("Expected used budget to be 0.005, got %f", status.UsedBudget)
	}

	if status.CallCount != 1 {
		t.Errorf("Expected call count to be 1, got %d", status.CallCount)
	}

	if status.AverageCostPerCall != 0.005 {
		t.Errorf("Expected average cost per call to be 0.005, got %f", status.AverageCostPerCall)
	}
}

func TestFallbackManager_GetFallbackStrategy(t *testing.T) {
	logger := &MockLogger{}
	manager := NewFallbackManager(logger)

	// Test getting existing strategy
	strategy := manager.GetFallbackStrategy("budget_exceeded")
	if strategy == nil {
		t.Errorf("Expected to get budget_exceeded strategy, got nil")
	}

	if strategy.Name != "Budget Exceeded Fallback" {
		t.Errorf("Expected strategy name 'Budget Exceeded Fallback', got '%s'", strategy.Name)
	}

	// Test getting non-existent strategy
	strategy = manager.GetFallbackStrategy("non_existent")
	if strategy != nil {
		t.Errorf("Expected nil for non-existent strategy, got %v", strategy)
	}
}

func TestFallbackManager_GetFallbackStrategyByCondition(t *testing.T) {
	logger := &MockLogger{}
	manager := NewFallbackManager(logger)

	// Test getting strategy by condition
	strategy := manager.GetFallbackStrategyByCondition("budget_exceeded")
	if strategy == nil {
		t.Errorf("Expected to get strategy for budget_exceeded condition, got nil")
	}

	if strategy.TriggerCondition != "budget_exceeded" {
		t.Errorf("Expected trigger condition 'budget_exceeded', got '%s'", strategy.TriggerCondition)
	}

	// Test getting strategy for non-existent condition
	strategy = manager.GetFallbackStrategyByCondition("non_existent")
	if strategy != nil {
		t.Errorf("Expected nil for non-existent condition, got %v", strategy)
	}
}

func TestFallbackManager_TestFallbackStrategy(t *testing.T) {
	logger := &MockLogger{}
	manager := NewFallbackManager(logger)

	// Test fallback strategy
	originalMethods := []string{"ml", "external_api", "keyword"}
	result, err := manager.TestFallbackStrategy("budget_exceeded", originalMethods)
	if err != nil {
		t.Errorf("Expected no error testing strategy, got: %v", err)
	}

	if result.StrategyName != "Budget Exceeded Fallback" {
		t.Errorf("Expected strategy name 'Budget Exceeded Fallback', got '%s'", result.StrategyName)
	}

	// Check that ML was substituted with keyword
	if result.MethodSubstitutions["ml"] != "keyword" {
		t.Errorf("Expected ML to be substituted with keyword, got '%s'", result.MethodSubstitutions["ml"])
	}

	// Check that external_api was substituted with description
	if result.MethodSubstitutions["external_api"] != "description" {
		t.Errorf("Expected external_api to be substituted with description, got '%s'", result.MethodSubstitutions["external_api"])
	}

	// Check that keyword was not substituted
	if result.MethodSubstitutions["keyword"] != "" {
		t.Errorf("Expected keyword to not be substituted, got '%s'", result.MethodSubstitutions["keyword"])
	}

	if !result.TestPassed {
		t.Errorf("Expected test to pass, got failed")
	}
}

func TestCostBasedRouter_UpdateTierConfig(t *testing.T) {
	registry := &MockMethodRegistry{methods: make(map[string]*MockMethod)}
	logger := &MockLogger{}
	router := NewCostBasedRouter(registry, logger)

	// Update free tier configuration
	newConfig := &TierConfig{
		Tier:                CustomerTierFree,
		Name:                "Updated Free Tier",
		Description:         "Updated free tier description",
		MaxCostPerCall:      0.002, // Increased from 0.001
		MonthlyBudgetLimit:  20.0,  // Increased from 10.0
		CostOptimization:    true,
		AllowedMethods:      []string{"keyword", "description"},
		RequiredMethods:     []string{"keyword"},
		ExcludedMethods:     []string{"ml", "external_api"},
		MinAccuracyTarget:   0.75,            // Increased from 0.70
		MinConfidenceTarget: 0.65,            // Increased from 0.60
		MaxResponseTime:     3 * time.Second, // Increased from 2 seconds
		CachePriority:       true,
		AdvancedFeatures:    false,
		RealTimeMonitoring:  false,
		CustomWeighting:     false,
	}

	err := router.UpdateTierConfig(CustomerTierFree, newConfig)
	if err != nil {
		t.Errorf("Expected no error updating tier config, got: %v", err)
	}

	// Verify the update
	updatedConfig, err := router.GetTierConfig(CustomerTierFree)
	if err != nil {
		t.Errorf("Expected no error getting updated config, got: %v", err)
	}

	if updatedConfig.MaxCostPerCall != 0.002 {
		t.Errorf("Expected max cost per call to be 0.002, got %f", updatedConfig.MaxCostPerCall)
	}

	if updatedConfig.MonthlyBudgetLimit != 20.0 {
		t.Errorf("Expected monthly budget limit to be 20.0, got %f", updatedConfig.MonthlyBudgetLimit)
	}

	if updatedConfig.MinAccuracyTarget != 0.75 {
		t.Errorf("Expected min accuracy target to be 0.75, got %f", updatedConfig.MinAccuracyTarget)
	}
}

func TestCostBasedRouter_GetAllTierConfigs(t *testing.T) {
	registry := &MockMethodRegistry{methods: make(map[string]*MockMethod)}
	logger := &MockLogger{}
	router := NewCostBasedRouter(registry, logger)

	configs := router.GetAllTierConfigs()

	expectedTiers := []CustomerTier{
		CustomerTierFree,
		CustomerTierStandard,
		CustomerTierPremium,
		CustomerTierEnterprise,
	}

	if len(configs) != len(expectedTiers) {
		t.Errorf("Expected %d tier configs, got %d", len(expectedTiers), len(configs))
	}

	for _, tier := range expectedTiers {
		if _, exists := configs[tier]; !exists {
			t.Errorf("Expected tier %s to be in configs", tier)
		}
	}
}

func TestFallbackManager_GetFallbackStatistics(t *testing.T) {
	logger := &MockLogger{}
	manager := NewFallbackManager(logger)

	stats := manager.GetFallbackStatistics()

	if stats.TotalStrategies == 0 {
		t.Errorf("Expected total strategies to be > 0, got %d", stats.TotalStrategies)
	}

	if stats.EnabledStrategies == 0 {
		t.Errorf("Expected enabled strategies to be > 0, got %d", stats.EnabledStrategies)
	}

	if stats.AverageCostReduction <= 0 {
		t.Errorf("Expected average cost reduction to be > 0, got %f", stats.AverageCostReduction)
	}

	if stats.AverageAccuracyImpact >= 0 {
		t.Errorf("Expected average accuracy impact to be < 0 (negative), got %f", stats.AverageAccuracyImpact)
	}
}

func TestCostBasedRouter_IntegrationTest(t *testing.T) {
	// Comprehensive integration test
	registry := &MockMethodRegistry{methods: make(map[string]*MockMethod)}

	// Add all methods
	registry.methods["keyword"] = &MockMethod{
		name: "keyword", methodType: "keyword", enabled: true, weight: 0.5,
		description: "Keyword-based classification",
	}
	registry.methods["ml"] = &MockMethod{
		name: "ml", methodType: "ml", enabled: true, weight: 0.4,
		description: "Machine learning classification",
	}
	registry.methods["external_api"] = &MockMethod{
		name: "external_api", methodType: "external_api", enabled: true, weight: 0.3,
		description: "External API classification",
	}
	registry.methods["description"] = &MockMethod{
		name: "description", methodType: "description", enabled: true, weight: 0.1,
		description: "Description-based classification",
	}

	logger := &MockLogger{}
	router := NewCostBasedRouter(registry, logger)

	// Test all tiers
	tiers := []CustomerTier{
		CustomerTierFree,
		CustomerTierStandard,
		CustomerTierPremium,
		CustomerTierEnterprise,
	}

	for _, tier := range tiers {
		t.Run(string(tier), func(t *testing.T) {
			decision, err := router.RouteRequest(context.Background(), tier, "Test Business", "A test business", "https://test.com")
			if err != nil {
				t.Errorf("Expected no error for tier %s, got: %v", tier, err)
				return
			}

			// Verify decision structure
			if decision.Tier != tier {
				t.Errorf("Expected tier %s, got %s", tier, decision.Tier)
			}

			if len(decision.SelectedMethods) == 0 {
				t.Errorf("Expected at least one selected method for tier %s", tier)
			}

			if decision.CostEstimate <= 0 {
				t.Errorf("Expected positive cost estimate for tier %s, got %f", tier, decision.CostEstimate)
			}

			if decision.ExpectedAccuracy <= 0 {
				t.Errorf("Expected positive accuracy estimate for tier %s, got %f", tier, decision.ExpectedAccuracy)
			}

			if decision.Reasoning == "" {
				t.Errorf("Expected reasoning for tier %s", tier)
			}

			// Verify method weights sum to approximately 1.0
			var totalWeight float64
			for _, weight := range decision.MethodWeights {
				totalWeight += weight
			}

			if totalWeight < 0.99 || totalWeight > 1.01 {
				t.Errorf("Expected method weights to sum to ~1.0 for tier %s, got %f", tier, totalWeight)
			}
		})
	}
}
