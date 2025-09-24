package risk

import (
	"testing"
	"time"
)

func TestRiskCategoryRegistry_Creation(t *testing.T) {
	registry := NewRiskCategoryRegistry()

	if registry == nil {
		t.Fatal("Expected registry to be created")
	}

	if len(registry.categories) != 0 {
		t.Errorf("Expected empty categories map, got %d", len(registry.categories))
	}

	if len(registry.factors) != 0 {
		t.Errorf("Expected empty factors map, got %d", len(registry.factors))
	}
}

func TestRiskCategoryRegistry_RegisterCategory(t *testing.T) {
	registry := NewRiskCategoryRegistry()

	// Create a test category definition
	categoryDef := &RiskCategoryDefinition{
		Category:    RiskCategoryFinancial,
		Name:        "Test Financial Risk",
		Description: "Test financial risk category",
		Weight:      0.25,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Subcategories: []RiskSubcategory{
			{
				ID:          "test_subcategory",
				Name:        "Test Subcategory",
				Description: "Test subcategory description",
				Weight:      1.0,
			},
		},
		Factors: []RiskFactorDefinition{
			{
				ID:              "test_factor",
				Name:            "Test Factor",
				Description:     "Test factor description",
				Category:        RiskCategoryFinancial,
				Subcategory:     "test_subcategory",
				Weight:          1.0,
				CalculationType: "direct",
				DataSources:     []string{"test_source"},
				Thresholds: map[RiskLevel]float64{
					RiskLevelLow:      10,
					RiskLevelMedium:   20,
					RiskLevelHigh:     30,
					RiskLevelCritical: 40,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	// Register the category
	registry.RegisterCategory(categoryDef)

	// Verify category was registered
	if len(registry.categories) != 1 {
		t.Errorf("Expected 1 category, got %d", len(registry.categories))
	}

	// Verify factor was registered
	if len(registry.factors) != 1 {
		t.Errorf("Expected 1 factor, got %d", len(registry.factors))
	}

	// Verify we can retrieve the category
	retrieved, exists := registry.GetCategory(RiskCategoryFinancial)
	if !exists {
		t.Error("Expected category to exist")
	}

	if retrieved.Name != "Test Financial Risk" {
		t.Errorf("Expected name 'Test Financial Risk', got '%s'", retrieved.Name)
	}

	// Verify we can retrieve the factor
	factor, exists := registry.GetFactor("test_factor")
	if !exists {
		t.Error("Expected factor to exist")
	}

	if factor.Name != "Test Factor" {
		t.Errorf("Expected factor name 'Test Factor', got '%s'", factor.Name)
	}
}

func TestRiskCategoryRegistry_GetCategory(t *testing.T) {
	registry := NewRiskCategoryRegistry()

	// Test getting non-existent category
	_, exists := registry.GetCategory(RiskCategoryFinancial)
	if exists {
		t.Error("Expected category to not exist")
	}

	// Register a category
	categoryDef := &RiskCategoryDefinition{
		Category:    RiskCategoryFinancial,
		Name:        "Financial Risk",
		Description: "Financial risk category",
		Weight:      0.25,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	registry.RegisterCategory(categoryDef)

	// Test getting existing category
	retrieved, exists := registry.GetCategory(RiskCategoryFinancial)
	if !exists {
		t.Error("Expected category to exist")
	}

	if retrieved.Category != RiskCategoryFinancial {
		t.Errorf("Expected category %s, got %s", RiskCategoryFinancial, retrieved.Category)
	}
}

func TestRiskCategoryRegistry_GetFactor(t *testing.T) {
	registry := NewRiskCategoryRegistry()

	// Test getting non-existent factor
	_, exists := registry.GetFactor("non_existent")
	if exists {
		t.Error("Expected factor to not exist")
	}

	// Register a category with factors
	categoryDef := &RiskCategoryDefinition{
		Category:    RiskCategoryFinancial,
		Name:        "Financial Risk",
		Description: "Financial risk category",
		Weight:      0.25,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Factors: []RiskFactorDefinition{
			{
				ID:              "cash_flow",
				Name:            "Cash Flow Risk",
				Description:     "Cash flow risk factor",
				Category:        RiskCategoryFinancial,
				Weight:          1.0,
				CalculationType: "direct",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
		},
	}

	registry.RegisterCategory(categoryDef)

	// Test getting existing factor
	factor, exists := registry.GetFactor("cash_flow")
	if !exists {
		t.Error("Expected factor to exist")
	}

	if factor.Name != "Cash Flow Risk" {
		t.Errorf("Expected factor name 'Cash Flow Risk', got '%s'", factor.Name)
	}
}

func TestRiskCategoryRegistry_ListCategories(t *testing.T) {
	registry := NewRiskCategoryRegistry()

	// Test empty registry
	categories := registry.ListCategories()
	if len(categories) != 0 {
		t.Errorf("Expected 0 categories, got %d", len(categories))
	}

	// Register categories
	financialDef := &RiskCategoryDefinition{
		Category:    RiskCategoryFinancial,
		Name:        "Financial Risk",
		Description: "Financial risk category",
		Weight:      0.25,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	operationalDef := &RiskCategoryDefinition{
		Category:    RiskCategoryOperational,
		Name:        "Operational Risk",
		Description: "Operational risk category",
		Weight:      0.2,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	registry.RegisterCategory(financialDef)
	registry.RegisterCategory(operationalDef)

	// Test listing categories
	categories = registry.ListCategories()
	if len(categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(categories))
	}

	// Verify both categories are present
	foundFinancial := false
	foundOperational := false

	for _, category := range categories {
		if category == RiskCategoryFinancial {
			foundFinancial = true
		}
		if category == RiskCategoryOperational {
			foundOperational = true
		}
	}

	if !foundFinancial {
		t.Error("Expected to find Financial category")
	}

	if !foundOperational {
		t.Error("Expected to find Operational category")
	}
}

func TestRiskCategoryRegistry_ListFactors(t *testing.T) {
	registry := NewRiskCategoryRegistry()

	// Test empty registry
	factors := registry.ListFactors()
	if len(factors) != 0 {
		t.Errorf("Expected 0 factors, got %d", len(factors))
	}

	// Register category with factors
	categoryDef := &RiskCategoryDefinition{
		Category:    RiskCategoryFinancial,
		Name:        "Financial Risk",
		Description: "Financial risk category",
		Weight:      0.25,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Factors: []RiskFactorDefinition{
			{
				ID:              "factor1",
				Name:            "Factor 1",
				Description:     "First factor",
				Category:        RiskCategoryFinancial,
				Weight:          1.0,
				CalculationType: "direct",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			{
				ID:              "factor2",
				Name:            "Factor 2",
				Description:     "Second factor",
				Category:        RiskCategoryFinancial,
				Weight:          1.0,
				CalculationType: "direct",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
		},
	}

	registry.RegisterCategory(categoryDef)

	// Test listing factors
	factors = registry.ListFactors()
	if len(factors) != 2 {
		t.Errorf("Expected 2 factors, got %d", len(factors))
	}

	// Verify both factors are present
	foundFactor1 := false
	foundFactor2 := false

	for _, factor := range factors {
		if factor.ID == "factor1" {
			foundFactor1 = true
		}
		if factor.ID == "factor2" {
			foundFactor2 = true
		}
	}

	if !foundFactor1 {
		t.Error("Expected to find factor1")
	}

	if !foundFactor2 {
		t.Error("Expected to find factor2")
	}
}

func TestRiskCategoryRegistry_GetFactorsByCategory(t *testing.T) {
	registry := NewRiskCategoryRegistry()

	// Register categories with factors
	financialDef := &RiskCategoryDefinition{
		Category:    RiskCategoryFinancial,
		Name:        "Financial Risk",
		Description: "Financial risk category",
		Weight:      0.25,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Factors: []RiskFactorDefinition{
			{
				ID:              "financial_factor1",
				Name:            "Financial Factor 1",
				Description:     "First financial factor",
				Category:        RiskCategoryFinancial,
				Weight:          1.0,
				CalculationType: "direct",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			{
				ID:              "financial_factor2",
				Name:            "Financial Factor 2",
				Description:     "Second financial factor",
				Category:        RiskCategoryFinancial,
				Weight:          1.0,
				CalculationType: "direct",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
		},
	}

	operationalDef := &RiskCategoryDefinition{
		Category:    RiskCategoryOperational,
		Name:        "Operational Risk",
		Description: "Operational risk category",
		Weight:      0.2,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Factors: []RiskFactorDefinition{
			{
				ID:              "operational_factor1",
				Name:            "Operational Factor 1",
				Description:     "First operational factor",
				Category:        RiskCategoryOperational,
				Weight:          1.0,
				CalculationType: "direct",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
		},
	}

	registry.RegisterCategory(financialDef)
	registry.RegisterCategory(operationalDef)

	// Test getting factors by category
	financialFactors := registry.GetFactorsByCategory(RiskCategoryFinancial)
	if len(financialFactors) != 2 {
		t.Errorf("Expected 2 financial factors, got %d", len(financialFactors))
	}

	operationalFactors := registry.GetFactorsByCategory(RiskCategoryOperational)
	if len(operationalFactors) != 1 {
		t.Errorf("Expected 1 operational factor, got %d", len(operationalFactors))
	}

	// Test getting factors for non-existent category
	nonExistentFactors := registry.GetFactorsByCategory(RiskCategoryRegulatory)
	if len(nonExistentFactors) != 0 {
		t.Errorf("Expected 0 factors for non-existent category, got %d", len(nonExistentFactors))
	}
}

func TestRiskCategoryRegistry_GetFactorsBySubcategory(t *testing.T) {
	registry := NewRiskCategoryRegistry()

	// Register category with subcategories
	categoryDef := &RiskCategoryDefinition{
		Category:    RiskCategoryFinancial,
		Name:        "Financial Risk",
		Description: "Financial risk category",
		Weight:      0.25,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Factors: []RiskFactorDefinition{
			{
				ID:              "liquidity_factor1",
				Name:            "Liquidity Factor 1",
				Description:     "First liquidity factor",
				Category:        RiskCategoryFinancial,
				Subcategory:     "financial_liquidity",
				Weight:          1.0,
				CalculationType: "direct",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			{
				ID:              "liquidity_factor2",
				Name:            "Liquidity Factor 2",
				Description:     "Second liquidity factor",
				Category:        RiskCategoryFinancial,
				Subcategory:     "financial_liquidity",
				Weight:          1.0,
				CalculationType: "direct",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			{
				ID:              "credit_factor1",
				Name:            "Credit Factor 1",
				Description:     "First credit factor",
				Category:        RiskCategoryFinancial,
				Subcategory:     "financial_credit",
				Weight:          1.0,
				CalculationType: "direct",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
		},
	}

	registry.RegisterCategory(categoryDef)

	// Test getting factors by subcategory
	liquidityFactors := registry.GetFactorsBySubcategory(RiskCategoryFinancial, "financial_liquidity")
	if len(liquidityFactors) != 2 {
		t.Errorf("Expected 2 liquidity factors, got %d", len(liquidityFactors))
	}

	creditFactors := registry.GetFactorsBySubcategory(RiskCategoryFinancial, "financial_credit")
	if len(creditFactors) != 1 {
		t.Errorf("Expected 1 credit factor, got %d", len(creditFactors))
	}

	// Test getting factors for non-existent subcategory
	nonExistentFactors := registry.GetFactorsBySubcategory(RiskCategoryFinancial, "non_existent")
	if len(nonExistentFactors) != 0 {
		t.Errorf("Expected 0 factors for non-existent subcategory, got %d", len(nonExistentFactors))
	}
}

func TestCreateDefaultRiskCategories(t *testing.T) {
	registry := CreateDefaultRiskCategories()

	if registry == nil {
		t.Fatal("Expected registry to be created")
	}

	// Verify all 5 categories are registered
	categories := registry.ListCategories()
	if len(categories) != 5 {
		t.Errorf("Expected 5 categories, got %d", len(categories))
	}

	// Verify expected categories are present
	expectedCategories := []RiskCategory{
		RiskCategoryFinancial,
		RiskCategoryOperational,
		RiskCategoryRegulatory,
		RiskCategoryReputational,
		RiskCategoryCybersecurity,
	}

	for _, expected := range expectedCategories {
		found := false
		for _, actual := range categories {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find category %s", expected)
		}
	}

	// Verify factors are registered
	factors := registry.ListFactors()
	if len(factors) == 0 {
		t.Error("Expected factors to be registered")
	}

	// Test specific category details
	financialCategory, exists := registry.GetCategory(RiskCategoryFinancial)
	if !exists {
		t.Error("Expected Financial category to exist")
	}

	if financialCategory.Name != "Financial Risk" {
		t.Errorf("Expected name 'Financial Risk', got '%s'", financialCategory.Name)
	}

	if financialCategory.Weight != 0.25 {
		t.Errorf("Expected weight 0.25, got %f", financialCategory.Weight)
	}

	// Verify subcategories
	if len(financialCategory.Subcategories) != 4 {
		t.Errorf("Expected 4 subcategories, got %d", len(financialCategory.Subcategories))
	}

	// Verify factors
	if len(financialCategory.Factors) != 4 {
		t.Errorf("Expected 4 factors, got %d", len(financialCategory.Factors))
	}

	// Test specific factor
	cashFlowFactor, exists := registry.GetFactor("cash_flow_coverage")
	if !exists {
		t.Error("Expected cash_flow_coverage factor to exist")
	}

	if cashFlowFactor.Name != "Cash Flow Coverage Ratio" {
		t.Errorf("Expected name 'Cash Flow Coverage Ratio', got '%s'", cashFlowFactor.Name)
	}

	if cashFlowFactor.Category != RiskCategoryFinancial {
		t.Errorf("Expected category Financial, got %s", cashFlowFactor.Category)
	}

	if cashFlowFactor.Subcategory != "financial_liquidity" {
		t.Errorf("Expected subcategory 'financial_liquidity', got '%s'", cashFlowFactor.Subcategory)
	}

	// Verify thresholds
	if len(cashFlowFactor.Thresholds) != 4 {
		t.Errorf("Expected 4 thresholds, got %d", len(cashFlowFactor.Thresholds))
	}

	// Verify data sources
	if len(cashFlowFactor.DataSources) != 2 {
		t.Errorf("Expected 2 data sources, got %d", len(cashFlowFactor.DataSources))
	}
}

func TestRiskCategoryDefinition_Validation(t *testing.T) {
	// Test valid category definition
	validCategory := &RiskCategoryDefinition{
		Category:    RiskCategoryFinancial,
		Name:        "Financial Risk",
		Description: "Financial risk category",
		Weight:      0.25,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Subcategories: []RiskSubcategory{
			{
				ID:          "test_sub",
				Name:        "Test Subcategory",
				Description: "Test description",
				Weight:      1.0,
			},
		},
		Factors: []RiskFactorDefinition{
			{
				ID:              "test_factor",
				Name:            "Test Factor",
				Description:     "Test factor description",
				Category:        RiskCategoryFinancial,
				Subcategory:     "test_sub",
				Weight:          1.0,
				CalculationType: "direct",
				DataSources:     []string{"test_source"},
				Thresholds: map[RiskLevel]float64{
					RiskLevelLow:      10,
					RiskLevelMedium:   20,
					RiskLevelHigh:     30,
					RiskLevelCritical: 40,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	registry := NewRiskCategoryRegistry()
	registry.RegisterCategory(validCategory)

	// Verify registration was successful
	if len(registry.categories) != 1 {
		t.Errorf("Expected 1 category, got %d", len(registry.categories))
	}

	if len(registry.factors) != 1 {
		t.Errorf("Expected 1 factor, got %d", len(registry.factors))
	}
}

func TestRiskFactorDefinition_Validation(t *testing.T) {
	// Test factor with all required fields
	factor := RiskFactorDefinition{
		ID:              "test_factor",
		Name:            "Test Factor",
		Description:     "Test factor description",
		Category:        RiskCategoryFinancial,
		Subcategory:     "test_subcategory",
		Weight:          1.0,
		CalculationType: "direct",
		DataSources:     []string{"test_source"},
		Thresholds: map[RiskLevel]float64{
			RiskLevelLow:      10,
			RiskLevelMedium:   20,
			RiskLevelHigh:     30,
			RiskLevelCritical: 40,
		},
		Formula:   "Test Formula",
		Unit:      "test_unit",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Verify all fields are set correctly
	if factor.ID != "test_factor" {
		t.Errorf("Expected ID 'test_factor', got '%s'", factor.ID)
	}

	if factor.Name != "Test Factor" {
		t.Errorf("Expected name 'Test Factor', got '%s'", factor.Name)
	}

	if factor.Category != RiskCategoryFinancial {
		t.Errorf("Expected category Financial, got %s", factor.Category)
	}

	if factor.CalculationType != "direct" {
		t.Errorf("Expected calculation type 'direct', got '%s'", factor.CalculationType)
	}

	if len(factor.DataSources) != 1 {
		t.Errorf("Expected 1 data source, got %d", len(factor.DataSources))
	}

	if len(factor.Thresholds) != 4 {
		t.Errorf("Expected 4 thresholds, got %d", len(factor.Thresholds))
	}

	if factor.Formula != "Test Formula" {
		t.Errorf("Expected formula 'Test Formula', got '%s'", factor.Formula)
	}

	if factor.Unit != "test_unit" {
		t.Errorf("Expected unit 'test_unit', got '%s'", factor.Unit)
	}
}

func TestRiskCategoryRegistry_Performance(t *testing.T) {
	registry := CreateDefaultRiskCategories()

	// Test performance of category retrieval
	start := time.Now()
	for i := 0; i < 1000; i++ {
		_, exists := registry.GetCategory(RiskCategoryFinancial)
		if !exists {
			t.Error("Expected category to exist")
		}
	}
	duration := time.Since(start)

	// Should complete within reasonable time (less than 1ms for 1000 lookups)
	if duration > time.Millisecond {
		t.Errorf("Category retrieval took too long: %v", duration)
	}

	// Test performance of factor retrieval
	start = time.Now()
	for i := 0; i < 1000; i++ {
		_, exists := registry.GetFactor("cash_flow_coverage")
		if !exists {
			t.Error("Expected factor to exist")
		}
	}
	duration = time.Since(start)

	// Should complete within reasonable time
	if duration > time.Millisecond {
		t.Errorf("Factor retrieval took too long: %v", duration)
	}
}
