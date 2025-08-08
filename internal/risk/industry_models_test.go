package risk

import (
	"testing"
)

func TestIndustryModelRegistry_RegisterModel(t *testing.T) {
	registry := NewIndustryModelRegistry()
	
	// Test valid model registration
	model := &IndustryModel{
		IndustryCode: "52",
		IndustryName: "Finance and Insurance",
		RiskFactors: []RiskFactor{
			{
				ID:       "regulatory_compliance",
				Name:     "Regulatory Compliance",
				Category: RiskCategoryRegulatory,
				Weight:   0.4,
			},
		},
		ModelVersion: "1.0",
		LastUpdated:  "2024-01-01",
	}
	
	err := registry.RegisterModel(model)
	if err != nil {
		t.Errorf("Expected no error when registering valid model, got %v", err)
	}
	
	// Test invalid model (empty industry code)
	invalidModel := &IndustryModel{
		IndustryCode: "",
		IndustryName: "Invalid Model",
		RiskFactors:  []RiskFactor{},
	}
	
	err = registry.RegisterModel(invalidModel)
	if err == nil {
		t.Error("Expected error when registering model with empty industry code")
	}
	
	// Test invalid model (no risk factors)
	invalidModel2 := &IndustryModel{
		IndustryCode: "99",
		IndustryName: "Invalid Model",
		RiskFactors:  []RiskFactor{},
	}
	
	err = registry.RegisterModel(invalidModel2)
	if err == nil {
		t.Error("Expected error when registering model with no risk factors")
	}
}

func TestIndustryModelRegistry_GetModel(t *testing.T) {
	registry := NewIndustryModelRegistry()
	
	// Register a test model
	model := &IndustryModel{
		IndustryCode: "52",
		IndustryName: "Finance and Insurance",
		RiskFactors: []RiskFactor{
			{
				ID:       "regulatory_compliance",
				Name:     "Regulatory Compliance",
				Category: RiskCategoryRegulatory,
				Weight:   0.4,
			},
		},
		ModelVersion: "1.0",
		LastUpdated:  "2024-01-01",
	}
	registry.RegisterModel(model)
	
	// Test exact match
	retrievedModel, exists := registry.GetModel("52")
	if !exists {
		t.Error("Expected model to exist for exact match")
	}
	if retrievedModel.IndustryCode != "52" {
		t.Errorf("Expected industry code 52, got %s", retrievedModel.IndustryCode)
	}
	
	// Test non-existent model
	_, exists = registry.GetModel("99")
	if exists {
		t.Error("Expected model to not exist for non-existent code")
	}
}

func TestIndustryModelRegistry_GetModelByNAICS(t *testing.T) {
	registry := NewIndustryModelRegistry()
	
	// Register models with different NAICS codes
	models := []*IndustryModel{
		{
			IndustryCode: "52",
			IndustryName: "Finance and Insurance",
			RiskFactors: []RiskFactor{
				{ID: "test", Name: "Test", Category: RiskCategoryFinancial, Weight: 1.0},
			},
		},
		{
			IndustryCode: "5415",
			IndustryName: "Computer Systems Design",
			RiskFactors: []RiskFactor{
				{ID: "test", Name: "Test", Category: RiskCategoryCybersecurity, Weight: 1.0},
			},
		},
	}
	
	for _, model := range models {
		registry.RegisterModel(model)
	}
	
	// Test exact match
	model, exists := registry.GetModelByNAICS("52")
	if !exists {
		t.Error("Expected model to exist for exact match")
	}
	if model.IndustryCode != "52" {
		t.Errorf("Expected industry code 52, got %s", model.IndustryCode)
	}
	
	// Test partial match
	model, exists = registry.GetModelByNAICS("541511")
	if !exists {
		t.Error("Expected model to exist for partial match")
	}
	if model.IndustryCode != "5415" {
		t.Errorf("Expected industry code 5415, got %s", model.IndustryCode)
	}
	
	// Test no match
	_, exists = registry.GetModelByNAICS("99")
	if exists {
		t.Error("Expected no model to exist for non-existent code")
	}
}

func TestIndustryModelRegistry_ListModels(t *testing.T) {
	registry := NewIndustryModelRegistry()
	
	// Register multiple models
	models := []*IndustryModel{
		{
			IndustryCode: "52",
			IndustryName: "Finance and Insurance",
			RiskFactors: []RiskFactor{
				{ID: "test1", Name: "Test1", Category: RiskCategoryFinancial, Weight: 1.0},
			},
		},
		{
			IndustryCode: "54",
			IndustryName: "Professional Services",
			RiskFactors: []RiskFactor{
				{ID: "test2", Name: "Test2", Category: RiskCategoryOperational, Weight: 1.0},
			},
		},
	}
	
	for _, model := range models {
		registry.RegisterModel(model)
	}
	
	// Test listing models
	listedModels := registry.ListModels()
	if len(listedModels) != 2 {
		t.Errorf("Expected 2 models, got %d", len(listedModels))
	}
	
	// Verify all models are present
	codes := make(map[string]bool)
	for _, model := range listedModels {
		codes[model.IndustryCode] = true
	}
	
	if !codes["52"] || !codes["54"] {
		t.Error("Expected both industry codes to be present in listed models")
	}
}

func TestIndustrySpecificScoringAlgorithm_CalculateScore(t *testing.T) {
	registry := CreateDefaultIndustryModels()
	algorithm := NewIndustrySpecificScoringAlgorithm(registry)
	
	// Test with financial industry data
	data := map[string]interface{}{
		"industry_code": "52",
		"regulatory_compliance": map[string]interface{}{
			"compliance_violations": 2.0,
			"regulatory_fines":      5000.0,
			"license_status":        "active",
		},
		"financial_stability": map[string]interface{}{
			"revenue":      1000000.0,
			"debt_ratio":   0.4,
			"cash_flow":    100000.0,
			"profit_margin": 0.15,
		},
		"cybersecurity": map[string]interface{}{
			"security_incidents": 1.0,
			"data_breaches":      0.0,
			"security_maturity":  3.5,
		},
	}
	
	score, confidence, err := algorithm.CalculateScore([]RiskFactor{}, data)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if score < 0 || score > 100 {
		t.Errorf("Score should be between 0 and 100, got %f", score)
	}
	
	if confidence < 0 || confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", confidence)
	}
	
	t.Logf("Financial industry score: %f, confidence: %f", score, confidence)
}

func TestIndustrySpecificScoringAlgorithm_CalculateScore_Technology(t *testing.T) {
	registry := CreateDefaultIndustryModels()
	algorithm := NewIndustrySpecificScoringAlgorithm(registry)
	
	// Test with technology industry data
	data := map[string]interface{}{
		"industry_code": "5415",
		"cybersecurity": map[string]interface{}{
			"security_incidents": 3.0,
			"data_breaches":      0.0,
			"security_maturity":  4.0,
		},
		"operational_efficiency": map[string]interface{}{
			"employee_turnover":     0.15,
			"operational_efficiency": 0.8,
			"process_maturity":      3.5,
		},
		"financial_stability": map[string]interface{}{
			"revenue":      2000000.0,
			"debt_ratio":   0.3,
			"cash_flow":    200000.0,
			"profit_margin": 0.2,
		},
	}
	
	score, confidence, err := algorithm.CalculateScore([]RiskFactor{}, data)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if score < 0 || score > 100 {
		t.Errorf("Score should be between 0 and 100, got %f", score)
	}
	
	if confidence < 0 || confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", confidence)
	}
	
	t.Logf("Technology industry score: %f, confidence: %f", score, confidence)
}

func TestIndustrySpecificScoringAlgorithm_CalculateScore_Healthcare(t *testing.T) {
	registry := CreateDefaultIndustryModels()
	algorithm := NewIndustrySpecificScoringAlgorithm(registry)
	
	// Test with healthcare industry data
	data := map[string]interface{}{
		"industry_code": "62",
		"regulatory_compliance": map[string]interface{}{
			"compliance_violations": 1.0,
			"regulatory_fines":      0.0,
			"license_status":        "active",
		},
		"cybersecurity": map[string]interface{}{
			"security_incidents": 2.0,
			"data_breaches":      0.0,
			"security_maturity":  4.5,
		},
		"operational_efficiency": map[string]interface{}{
			"employee_turnover":     0.1,
			"operational_efficiency": 0.85,
			"process_maturity":      4.0,
		},
	}
	
	score, confidence, err := algorithm.CalculateScore([]RiskFactor{}, data)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if score < 0 || score > 100 {
		t.Errorf("Score should be between 0 and 100, got %f", score)
	}
	
	if confidence < 0 || confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", confidence)
	}
	
	t.Logf("Healthcare industry score: %f, confidence: %f", score, confidence)
}

func TestIndustrySpecificScoringAlgorithm_CalculateScore_NoIndustryCode(t *testing.T) {
	registry := CreateDefaultIndustryModels()
	algorithm := NewIndustrySpecificScoringAlgorithm(registry)
	
	// Test with no industry code (should fall back to base algorithm)
	data := map[string]interface{}{
		"financial_stability": map[string]interface{}{
			"revenue":      1000000.0,
			"debt_ratio":   0.4,
			"cash_flow":    100000.0,
			"profit_margin": 0.15,
		},
	}
	
	score, confidence, err := algorithm.CalculateScore([]RiskFactor{}, data)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if score < 0 || score > 100 {
		t.Errorf("Score should be between 0 and 100, got %f", score)
	}
	
	if confidence < 0 || confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", confidence)
	}
	
	t.Logf("No industry code score: %f, confidence: %f", score, confidence)
}

func TestIndustrySpecificScoringAlgorithm_CalculateScore_UnknownIndustry(t *testing.T) {
	registry := CreateDefaultIndustryModels()
	algorithm := NewIndustrySpecificScoringAlgorithm(registry)
	
	// Test with unknown industry code (should fall back to base algorithm)
	data := map[string]interface{}{
		"industry_code": "99",
		"financial_stability": map[string]interface{}{
			"revenue":      1000000.0,
			"debt_ratio":   0.4,
			"cash_flow":    100000.0,
			"profit_margin": 0.15,
		},
	}
	
	score, confidence, err := algorithm.CalculateScore([]RiskFactor{}, data)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if score < 0 || score > 100 {
		t.Errorf("Score should be between 0 and 100, got %f", score)
	}
	
	if confidence < 0 || confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", confidence)
	}
	
	t.Logf("Unknown industry score: %f, confidence: %f", score, confidence)
}

func TestIndustrySpecificScoringAlgorithm_CalculateLevel(t *testing.T) {
	registry := CreateDefaultIndustryModels()
	algorithm := NewIndustrySpecificScoringAlgorithm(registry)
	
	// Test with custom thresholds
	customThresholds := map[RiskLevel]float64{
		RiskLevelLow:      20.0,
		RiskLevelMedium:   45.0,
		RiskLevelHigh:     75.0,
		RiskLevelCritical: 90.0,
	}
	
	testCases := []struct {
		score    float64
		expected RiskLevel
	}{
		{15.0, RiskLevelLow},
		{35.0, RiskLevelLow},
		{55.0, RiskLevelMedium},
		{85.0, RiskLevelHigh},
		{95.0, RiskLevelCritical},
	}
	
	for i, tc := range testCases {
		result := algorithm.CalculateLevel(tc.score, customThresholds)
		if result != tc.expected {
			t.Errorf("Test case %d: Expected %s, got %s for score %f", i+1, tc.expected, result, tc.score)
		}
	}
}

func TestIndustrySpecificScoringAlgorithm_CalculateConfidence(t *testing.T) {
	registry := CreateDefaultIndustryModels()
	algorithm := NewIndustrySpecificScoringAlgorithm(registry)
	
	// Test with financial industry data
	data := map[string]interface{}{
		"industry_code": "52",
		"regulatory_compliance": map[string]interface{}{
			"compliance_violations": 2.0,
		},
		"financial_stability": map[string]interface{}{
			"revenue": 1000000.0,
		},
		// Missing cybersecurity data
	}
	
	confidence := algorithm.CalculateConfidence([]RiskFactor{}, data)
	if confidence < 0 || confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", confidence)
	}
	
	t.Logf("Industry-specific confidence: %f", confidence)
}

func TestCreateDefaultIndustryModels(t *testing.T) {
	registry := CreateDefaultIndustryModels()
	
	// Test that all expected models are created
	expectedModels := []string{"52", "54", "62", "31", "44"}
	
	for _, expectedCode := range expectedModels {
		model, exists := registry.GetModel(expectedCode)
		if !exists {
			t.Errorf("Expected model for industry code %s to exist", expectedCode)
		}
		if model.IndustryCode != expectedCode {
			t.Errorf("Expected industry code %s, got %s", expectedCode, model.IndustryCode)
		}
	}
	
	// Test model properties
	financialModel, exists := registry.GetModel("52")
	if !exists {
		t.Fatal("Financial model should exist")
	}
	
	if financialModel.IndustryName != "Finance and Insurance" {
		t.Errorf("Expected industry name 'Finance and Insurance', got %s", financialModel.IndustryName)
	}
	
	if len(financialModel.RiskFactors) != 3 {
		t.Errorf("Expected 3 risk factors, got %d", len(financialModel.RiskFactors))
	}
	
	// Test special factors
	if specialFactors, exists := financialModel.SpecialFactors["regulatory_compliance"]; !exists {
		t.Error("Expected special factors for regulatory compliance")
	} else {
		specialMap, ok := specialFactors.(map[string]interface{})
		if !ok {
			t.Error("Expected special factors to be a map")
		}
		if _, exists := specialMap["required_licenses"]; !exists {
			t.Error("Expected required_licenses in special factors")
		}
	}
}

func TestIndustryModel_Validation(t *testing.T) {
	// Test valid model
	validModel := &IndustryModel{
		IndustryCode: "52",
		IndustryName: "Finance and Insurance",
		RiskFactors: []RiskFactor{
			{
				ID:       "test",
				Name:     "Test Factor",
				Category: RiskCategoryFinancial,
				Weight:   1.0,
			},
		},
		ModelVersion: "1.0",
		LastUpdated:  "2024-01-01",
	}
	
	registry := NewIndustryModelRegistry()
	err := registry.RegisterModel(validModel)
	if err != nil {
		t.Errorf("Expected no error for valid model, got %v", err)
	}
	
	// Test model with empty industry code
	invalidModel := &IndustryModel{
		IndustryCode: "",
		IndustryName: "Invalid Model",
		RiskFactors:  []RiskFactor{},
	}
	
	err = registry.RegisterModel(invalidModel)
	if err == nil {
		t.Error("Expected error for model with empty industry code")
	}
	
	// Test model with no risk factors
	invalidModel2 := &IndustryModel{
		IndustryCode: "99",
		IndustryName: "Invalid Model",
		RiskFactors:  []RiskFactor{},
	}
	
	err = registry.RegisterModel(invalidModel2)
	if err == nil {
		t.Error("Expected error for model with no risk factors")
	}
}

func TestIndustrySpecificScoringAlgorithm_Performance(t *testing.T) {
	registry := CreateDefaultIndustryModels()
	algorithm := NewIndustrySpecificScoringAlgorithm(registry)
	
	// Create test data for all industries
	industries := []string{"52", "54", "62", "31", "44"}
	
	for _, industryCode := range industries {
		data := map[string]interface{}{
			"industry_code": industryCode,
			"financial_stability": map[string]interface{}{
				"revenue":      1000000.0,
				"debt_ratio":   0.4,
				"cash_flow":    100000.0,
				"profit_margin": 0.15,
			},
			"operational_efficiency": map[string]interface{}{
				"employee_turnover":     0.15,
				"operational_efficiency": 0.8,
				"process_maturity":      3.5,
			},
			"regulatory_compliance": map[string]interface{}{
				"compliance_violations": 1.0,
				"regulatory_fines":      0.0,
				"license_status":        "active",
			},
			"cybersecurity": map[string]interface{}{
				"security_incidents": 1.0,
				"data_breaches":      0.0,
				"security_maturity":  3.5,
			},
		}
		
		score, confidence, err := algorithm.CalculateScore([]RiskFactor{}, data)
		if err != nil {
			t.Errorf("Expected no error for industry %s, got %v", industryCode, err)
		}
		
		if score < 0 || score > 100 {
			t.Errorf("Score should be between 0 and 100 for industry %s, got %f", industryCode, score)
		}
		
		if confidence < 0 || confidence > 1 {
			t.Errorf("Confidence should be between 0 and 1 for industry %s, got %f", industryCode, confidence)
		}
		
		t.Logf("Industry %s score: %f, confidence: %f", industryCode, score, confidence)
	}
}
