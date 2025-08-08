package risk

import (
	"strings"
	"testing"
	"time"
)

func TestNewRiskFactorCalculator(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	if calculator == nil {
		t.Fatal("Expected calculator to be created")
	}

	if calculator.registry != registry {
		t.Error("Expected calculator to use the provided registry")
	}
}

func TestRiskFactorCalculator_CalculateFactor_Direct(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	// Test direct calculation for cash flow coverage ratio
	input := RiskFactorInput{
		FactorID:    "cash_flow_coverage",
		Data:        map[string]interface{}{"cash_flow_coverage": 2.5, "financial_statements": "data"},
		Timestamp:   time.Now(),
		Source:      "financial_statements",
		Reliability: 0.9,
	}

	result, err := calculator.CalculateFactor(input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.FactorID != "cash_flow_coverage" {
		t.Errorf("Expected factor ID 'cash_flow_coverage', got '%s'", result.FactorID)
	}

	if result.Category != RiskCategoryFinancial {
		t.Errorf("Expected category Financial, got %s", result.Category)
	}

	if result.Subcategory != "financial_liquidity" {
		t.Errorf("Expected subcategory 'financial_liquidity', got '%s'", result.Subcategory)
	}

	// Should be low risk for 2.5 ratio (good cash flow coverage)
	if result.Level != RiskLevelLow && result.Level != RiskLevelMedium {
		t.Errorf("Expected risk level Low or Medium for good ratio, got %s (score: %f)", result.Level, result.Score)
	}

	if result.Score < 0 || result.Score > 100 {
		t.Errorf("Expected score between 0 and 100, got %f", result.Score)
	}

	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got %f", result.Confidence)
	}
}

func TestRiskFactorCalculator_CalculateFactor_Derived(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	// Test derived calculation for profit margin trend
	input := RiskFactorInput{
		FactorID:    "profit_margin",
		Data:        map[string]interface{}{"current_margin": 0.15, "previous_margin": 0.12, "income_statements": "data"},
		Timestamp:   time.Now(),
		Source:      "income_statements",
		Reliability: 0.8,
	}

	result, err := calculator.CalculateFactor(input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.FactorID != "profit_margin" {
		t.Errorf("Expected factor ID 'profit_margin', got '%s'", result.FactorID)
	}

	if result.Category != RiskCategoryFinancial {
		t.Errorf("Expected category Financial, got %s", result.Category)
	}

	if result.Subcategory != "financial_performance" {
		t.Errorf("Expected subcategory 'financial_performance', got '%s'", result.Subcategory)
	}

	// Should be low risk for positive trend
	if result.Level != RiskLevelLow && result.Level != RiskLevelMedium {
		t.Errorf("Expected risk level Low or Medium for positive trend, got %s (score: %f)", result.Level, result.Score)
	}
}

func TestRiskFactorCalculator_CalculateFactor_Composite(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	// Test composite calculation for security score
	input := RiskFactorInput{
		FactorID: "security_score",
		Data: map[string]interface{}{
			"vulnerability_count":  2,
			"patch_compliance":     0.95,
			"incident_count":       0,
			"security_assessments": "data",
		},
		Timestamp:   time.Now(),
		Source:      "security_assessments",
		Reliability: 0.85,
	}

	result, err := calculator.CalculateFactor(input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.FactorID != "security_score" {
		t.Errorf("Expected factor ID 'security_score', got '%s'", result.FactorID)
	}

	if result.Category != RiskCategoryCybersecurity {
		t.Errorf("Expected category Cybersecurity, got %s", result.Category)
	}

	if result.Subcategory != "cybersecurity_technical" {
		t.Errorf("Expected subcategory 'cybersecurity_technical', got '%s'", result.Subcategory)
	}
}

func TestRiskFactorCalculator_CalculateFactor_NotFound(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	input := RiskFactorInput{
		FactorID:    "non_existent_factor",
		Data:        map[string]interface{}{"value": 10.0},
		Timestamp:   time.Now(),
		Source:      "test",
		Reliability: 0.9,
	}

	_, err := calculator.CalculateFactor(input)
	if err == nil {
		t.Error("Expected error for non-existent factor")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got %v", err)
	}
}

func TestRiskFactorCalculator_CalculateFactor_InvalidInput(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	// Test with invalid reliability score
	input := RiskFactorInput{
		FactorID:    "cash_flow_coverage",
		Data:        map[string]interface{}{"cash_flow_coverage": 2.5},
		Timestamp:   time.Now(),
		Source:      "financial_statements",
		Reliability: 1.5, // Invalid: > 1.0
	}

	_, err := calculator.CalculateFactor(input)
	if err == nil {
		t.Error("Expected error for invalid reliability score")
	}

	// Test with missing required data source (should not fail if we have data)
	input = RiskFactorInput{
		FactorID:    "cash_flow_coverage",
		Data:        map[string]interface{}{"unrelated_field": 10.0},
		Timestamp:   time.Now(),
		Source:      "financial_statements",
		Reliability: 0.9,
	}

	_, err = calculator.CalculateFactor(input)
	if err != nil {
		t.Errorf("Expected no error for missing data source when we have data, got %v", err)
	}
}

func TestRiskFactorCalculator_CalculateDirectScore(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	factorDef, _ := registry.GetFactor("cash_flow_coverage")

	// Test with exact field match
	input := RiskFactorInput{
		FactorID:    "cash_flow_coverage",
		Data:        map[string]interface{}{"cash_flow_coverage": 1.8, "financial_statements": "data"},
		Timestamp:   time.Now(),
		Source:      "financial_statements",
		Reliability: 0.9,
	}

	score, explanation, evidence, err := calculator.calculateDirectScore(input, factorDef)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if score != 1.8 {
		t.Errorf("Expected score 1.8, got %f", score)
	}

	if !strings.Contains(explanation, "Direct measurement") {
		t.Errorf("Expected explanation to contain 'Direct measurement', got '%s'", explanation)
	}

	if len(evidence) == 0 {
		t.Error("Expected evidence to be provided")
	}
}

func TestRiskFactorCalculator_CalculateDerivedScore(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	factorDef, _ := registry.GetFactor("profit_margin")

	// Test with multiple values
	input := RiskFactorInput{
		FactorID: "profit_margin",
		Data: map[string]interface{}{
			"current_margin":    0.15,
			"previous_margin":   0.12,
			"income_statements": "data",
		},
		Timestamp:   time.Now(),
		Source:      "income_statements",
		Reliability: 0.8,
	}

	score, explanation, evidence, err := calculator.calculateDerivedScore(input, factorDef)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should calculate average
	expectedScore := (0.15 + 0.12) / 2
	if score != expectedScore {
		t.Errorf("Expected score %f, got %f", expectedScore, score)
	}

	if !strings.Contains(explanation, "Average calculation") && !strings.Contains(explanation, "Default calculation") {
		t.Errorf("Expected explanation to contain 'Average calculation' or 'Default calculation', got '%s'", explanation)
	}

	if len(evidence) != 2 {
		t.Errorf("Expected 2 evidence items, got %d", len(evidence))
	}
}

func TestRiskFactorCalculator_CalculateCompositeScore(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	factorDef, _ := registry.GetFactor("security_score")

	// Test with multiple components
	input := RiskFactorInput{
		FactorID: "security_score",
		Data: map[string]interface{}{
			"vulnerability_count":  3,
			"patch_compliance":     0.92,
			"incident_count":       1,
			"security_assessments": "data",
		},
		Timestamp:   time.Now(),
		Source:      "security_assessments",
		Reliability: 0.85,
	}

	score, explanation, evidence, err := calculator.calculateCompositeScore(input, factorDef)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if score < 0 || score > 100 {
		t.Errorf("Expected score between 0 and 100, got %f", score)
	}

	if !strings.Contains(explanation, "Composite calculation") {
		t.Errorf("Expected explanation to contain 'Composite calculation', got '%s'", explanation)
	}

	if len(evidence) != 3 {
		t.Errorf("Expected 3 evidence items, got %d", len(evidence))
	}
}

func TestRiskFactorCalculator_CalculateFinancialComponent(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	// Test ratio calculation
	score := calculator.calculateFinancialComponent(2.5, "debt_to_equity_ratio")
	if score != 20.0 {
		t.Errorf("Expected score 20.0 for good ratio, got %f", score)
	}

	score = calculator.calculateFinancialComponent(1.2, "debt_to_equity_ratio")
	if score != 60.0 {
		t.Errorf("Expected score 60.0 for medium ratio, got %f", score)
	}

	// Test score calculation
	score = calculator.calculateFinancialComponent(85, "credit_score")
	if score != 20.0 {
		t.Errorf("Expected score 20.0 for good credit score, got %f", score)
	}

	// Test trend calculation
	score = calculator.calculateFinancialComponent(0.05, "profit_trend")
	if score != 30.0 {
		t.Errorf("Expected score 30.0 for positive trend, got %f", score)
	}
}

func TestRiskFactorCalculator_CalculateOperationalComponent(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	// Test turnover calculation
	score := calculator.calculateOperationalComponent(0.03, "employee_turnover")
	if score != 20.0 {
		t.Errorf("Expected score 20.0 for low turnover, got %f", score)
	}

	score = calculator.calculateOperationalComponent(0.20, "employee_turnover")
	if score != 60.0 {
		t.Errorf("Expected score 60.0 for medium turnover, got %f", score)
	}

	// Test uptime calculation
	score = calculator.calculateOperationalComponent(0.99, "system_uptime")
	if score != 20.0 {
		t.Errorf("Expected score 20.0 for high uptime, got %f", score)
	}

	// Test concentration calculation
	score = calculator.calculateOperationalComponent(0.25, "supplier_concentration")
	if score != 20.0 {
		t.Errorf("Expected score 20.0 for low concentration, got %f", score)
	}
}

func TestRiskFactorCalculator_CalculateRegulatoryComponent(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	// Test compliance score calculation
	score := calculator.calculateRegulatoryComponent(95, "compliance_score")
	if score != 20.0 {
		t.Errorf("Expected score 20.0 for high compliance, got %f", score)
	}

	// Test violation count calculation
	score = calculator.calculateRegulatoryComponent(0, "regulatory_violations")
	if score != 20.0 {
		t.Errorf("Expected score 20.0 for no violations, got %f", score)
	}

	score = calculator.calculateRegulatoryComponent(2, "regulatory_violations")
	if score != 60.0 {
		t.Errorf("Expected score 60.0 for medium violations, got %f", score)
	}

	// Test license status calculation
	score = calculator.calculateRegulatoryComponent(1.0, "license_status")
	if score != 20.0 {
		t.Errorf("Expected score 20.0 for valid licenses, got %f", score)
	}
}

func TestRiskFactorCalculator_CalculateReputationalComponent(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	// Test sentiment calculation
	score := calculator.calculateReputationalComponent(0.8, "sentiment_score")
	if score != 20.0 {
		t.Errorf("Expected score 20.0 for positive sentiment, got %f", score)
	}

	// Test satisfaction calculation
	score = calculator.calculateReputationalComponent(4.5, "customer_satisfaction")
	if score != 20.0 {
		t.Errorf("Expected score 20.0 for high satisfaction, got %f", score)
	}

	// Test negative mentions calculation
	score = calculator.calculateReputationalComponent(0, "negative_mentions")
	if score != 20.0 {
		t.Errorf("Expected score 20.0 for no negative mentions, got %f", score)
	}

	score = calculator.calculateReputationalComponent(10, "negative_mentions")
	if score != 60.0 {
		t.Errorf("Expected score 60.0 for medium negative mentions, got %f", score)
	}
}

func TestRiskFactorCalculator_CalculateCybersecurityComponent(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	// Test security score calculation
	score := calculator.calculateCybersecurityComponent(90, "security_score")
	if score != 20.0 {
		t.Errorf("Expected score 20.0 for high security score, got %f", score)
	}

	// Test breach count calculation
	score = calculator.calculateCybersecurityComponent(0, "data_breaches")
	if score != 20.0 {
		t.Errorf("Expected score 20.0 for no breaches, got %f", score)
	}

	score = calculator.calculateCybersecurityComponent(2, "data_breaches")
	if score != 60.0 {
		t.Errorf("Expected score 60.0 for medium breaches, got %f", score)
	}

	// Test patch compliance calculation
	score = calculator.calculateCybersecurityComponent(0.98, "patch_compliance")
	if score != 20.0 {
		t.Errorf("Expected score 20.0 for high patch compliance, got %f", score)
	}
}

func TestRiskFactorCalculator_NormalizeScore(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	// Test with default normalization
	factorDef, _ := registry.GetFactor("cash_flow_coverage")

	// Test with value above highest threshold (should be low risk)
	score := calculator.normalizeScore(3.0, factorDef)
	if score != 25.0 {
		t.Errorf("Expected score 25.0 for value above highest threshold, got %f", score)
	}

	// Test with value below lowest threshold (should be critical risk)
	score = calculator.normalizeScore(0.3, factorDef)
	if score != 100.0 {
		t.Errorf("Expected score 100.0 for value below lowest threshold, got %f", score)
	}

	// Test with value in middle range
	score = calculator.normalizeScore(1.25, factorDef)
	if score < 0 || score > 100 {
		t.Errorf("Expected score between 0 and 100, got %f", score)
	}
}

func TestRiskFactorCalculator_NormalizeWithThresholds(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	thresholds := map[RiskLevel]float64{
		RiskLevelLow:      2.0,
		RiskLevelMedium:   1.5,
		RiskLevelHigh:     1.0,
		RiskLevelCritical: 0.5,
	}

		// Test score below lowest threshold (critical risk)
	score := calculator.normalizeWithThresholds(0.3, thresholds)
	if score != 100.0 {
		t.Errorf("Expected score 100.0 for value below lowest threshold, got %f", score)
	}
	
	// Test score above highest threshold (low risk)
	score = calculator.normalizeWithThresholds(3.0, thresholds)
	if score != 25.0 {
		t.Errorf("Expected score 25.0 for value above highest threshold, got %f", score)
	}

	// Test score in middle range
	score = calculator.normalizeWithThresholds(1.25, thresholds)
	if score < 0 || score > 100 {
		t.Errorf("Expected score between 0 and 100, got %f", score)
	}
}

func TestRiskFactorCalculator_DetermineRiskLevel(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	// Test with default thresholds
	level := calculator.determineRiskLevel(20.0, nil)
	if level != RiskLevelLow {
		t.Errorf("Expected risk level Low for score 20.0, got %s", level)
	}

	level = calculator.determineRiskLevel(60.0, nil)
	if level != RiskLevelHigh {
		t.Errorf("Expected risk level High for score 60.0, got %s", level)
	}

	level = calculator.determineRiskLevel(90.0, nil)
	if level != RiskLevelCritical {
		t.Errorf("Expected risk level Critical for score 90.0, got %s", level)
	}

	// Test with custom thresholds
	customThresholds := map[RiskLevel]float64{
		RiskLevelLow:      30.0,
		RiskLevelMedium:   60.0,
		RiskLevelHigh:     80.0,
		RiskLevelCritical: 90.0,
	}

	level = calculator.determineRiskLevel(25.0, customThresholds)
	if level != RiskLevelLow {
		t.Errorf("Expected risk level Low for score 25.0 with custom thresholds, got %s", level)
	}

	level = calculator.determineRiskLevel(70.0, customThresholds)
	if level != RiskLevelHigh {
		t.Errorf("Expected risk level High for score 70.0 with custom thresholds, got %s", level)
	}
}

func TestRiskFactorCalculator_CalculateConfidence(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	factorDef, _ := registry.GetFactor("cash_flow_coverage")

	// Test with high reliability and complete data
	input := RiskFactorInput{
		FactorID:    "cash_flow_coverage",
		Data:        map[string]interface{}{"cash_flow_coverage": 2.5, "financial_statements": "data"},
		Timestamp:   time.Now(),
		Source:      "financial_statements",
		Reliability: 0.9,
	}

	confidence := calculator.calculateConfidence(input, factorDef)
	if confidence < 0.8 {
		t.Errorf("Expected high confidence for reliable and complete data, got %f", confidence)
	}

	// Test with low reliability
	input.Reliability = 0.5
	confidence = calculator.calculateConfidence(input, factorDef)
	if confidence > 0.6 {
		t.Errorf("Expected lower confidence for low reliability, got %f", confidence)
	}

	// Test with old data
	input.Timestamp = time.Now().Add(-30 * 24 * time.Hour) // 30 days old
	confidence = calculator.calculateConfidence(input, factorDef)
	if confidence > 0.5 {
		t.Errorf("Expected lower confidence for old data, got %f", confidence)
	}
}

func TestRiskFactorCalculator_ToFloat64(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	// Test various types
	testCases := []struct {
		input    interface{}
		expected float64
		valid    bool
	}{
		{float64(10.5), 10.5, true},
		{float32(10.5), 10.5, true},
		{int(10), 10.0, true},
		{int32(10), 10.0, true},
		{int64(10), 10.0, true},
		{"10.5", 10.5, true},
		{"invalid", 0.0, false},
		{true, 0.0, false},
		{nil, 0.0, false},
	}

	for _, tc := range testCases {
		result, valid := calculator.toFloat64(tc.input)
		if valid != tc.valid {
			t.Errorf("Expected valid=%v for input %v, got %v", tc.valid, tc.input, valid)
		}
		if valid && result != tc.expected {
			t.Errorf("Expected %f for input %v, got %f", tc.expected, tc.input, result)
		}
	}
}

func TestRiskFactorCalculator_ApplyFormula(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	values := []float64{10.0, 20.0, 30.0}
	evidence := []string{"test evidence"}

	// Test average formula
	score, explanation, resultEvidence, err := calculator.applyFormula(values, "average", evidence)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedScore := (10.0 + 20.0 + 30.0) / 3
	if score != expectedScore {
		t.Errorf("Expected score %f, got %f", expectedScore, score)
	}

	if !strings.Contains(explanation, "Average calculation") {
		t.Errorf("Expected explanation to contain 'Average calculation', got '%s'", explanation)
	}

	if len(resultEvidence) != 1 {
		t.Errorf("Expected 1 evidence item, got %d", len(resultEvidence))
	}

	// Test sum formula
	score, explanation, _, err = calculator.applyFormula(values, "sum", evidence)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedScore = 10.0 + 20.0 + 30.0
	if score != expectedScore {
		t.Errorf("Expected score %f, got %f", expectedScore, score)
	}

	// Test max formula
	score, explanation, _, err = calculator.applyFormula(values, "max", evidence)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if score != 30.0 {
		t.Errorf("Expected score 30.0, got %f", score)
	}

	// Test min formula
	score, explanation, _, err = calculator.applyFormula(values, "min", evidence)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if score != 10.0 {
		t.Errorf("Expected score 10.0, got %f", score)
	}

	// Test unknown formula (should default to average)
	score, explanation, _, err = calculator.applyFormula(values, "unknown", evidence)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedScore = (10.0 + 20.0 + 30.0) / 3
	if score != expectedScore {
		t.Errorf("Expected score %f for unknown formula, got %f", expectedScore, score)
	}
}

func TestRiskFactorCalculator_Performance(t *testing.T) {
	registry := CreateDefaultRiskCategories()
	calculator := NewRiskFactorCalculator(registry)

	// Test performance with multiple calculations
	start := time.Now()

	for i := 0; i < 1000; i++ {
		input := RiskFactorInput{
			FactorID:    "cash_flow_coverage",
			Data:        map[string]interface{}{"cash_flow_coverage": float64(i % 100), "financial_statements": "data"},
			Timestamp:   time.Now(),
			Source:      "financial_statements",
			Reliability: 0.9,
		}

		_, err := calculator.CalculateFactor(input)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	}

	duration := time.Since(start)

	// Should complete within reasonable time (less than 1 second for 1000 calculations)
	if duration > time.Second {
		t.Errorf("Performance test took too long: %v", duration)
	}
}
