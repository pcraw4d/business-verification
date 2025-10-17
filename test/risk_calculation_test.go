package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"kyb-platform/internal/risk"
)

// TestRiskFactorCalculator_CalculateFactor tests the main calculation function
func TestRiskFactorCalculator_CalculateFactor(t *testing.T) {
	registry := createTestRegistry()
	calculator := risk.NewRiskFactorCalculator(registry)

	tests := []struct {
		name               string
		input              risk.RiskFactorInput
		expectedError      bool
		expectedScore      float64
		expectedLevel      risk.RiskLevel
		expectedConfidence float64
	}{
		{
			name: "Valid Direct Calculation",
			input: risk.RiskFactorInput{
				FactorID: "cash_flow_coverage",
				Data: map[string]interface{}{
					"cash_flow_coverage": 2.5,
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
			},
			expectedError:      false,
			expectedScore:      25.0, // Low risk for high coverage
			expectedLevel:      risk.RiskLevelLow,
			expectedConfidence: 0.9,
		},
		{
			name: "Valid Derived Calculation",
			input: risk.RiskFactorInput{
				FactorID: "debt_ratio",
				Data: map[string]interface{}{
					"total_debt":   1000000,
					"total_assets": 2000000,
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.8,
			},
			expectedError:      false,
			expectedLevel:      risk.RiskLevelMedium,
			expectedConfidence: 0.8,
		},
		{
			name: "Invalid Factor ID",
			input: risk.RiskFactorInput{
				FactorID: "nonexistent_factor",
				Data: map[string]interface{}{
					"value": 100,
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
			},
			expectedError: true,
		},
		{
			name: "Invalid Reliability Score",
			input: risk.RiskFactorInput{
				FactorID: "cash_flow_coverage",
				Data: map[string]interface{}{
					"cash_flow_coverage": 2.5,
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 1.5, // Invalid: > 1.0
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calculator.CalculateFactor(tt.input)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.input.FactorID, result.FactorID)
				assert.Equal(t, tt.expectedLevel, result.Level)
				assert.InDelta(t, tt.expectedConfidence, result.Confidence, 0.1)
				assert.GreaterOrEqual(t, result.Score, 0.0)
				assert.LessOrEqual(t, result.Score, 100.0)
				assert.NotEmpty(t, result.Explanation)
				assert.NotEmpty(t, result.Evidence)
			}
		})
	}
}

// TestRiskFactorCalculator_CalculateDirectScore tests direct score calculation
func TestRiskFactorCalculator_CalculateDirectScore(t *testing.T) {
	registry := createTestRegistry()
	calculator := risk.NewRiskFactorCalculator(registry)

	tests := []struct {
		name          string
		input         risk.RiskFactorInput
		expectedScore float64
		expectedError bool
	}{
		{
			name: "Direct Value Found",
			input: risk.RiskFactorInput{
				FactorID: "cash_flow_coverage",
				Data: map[string]interface{}{
					"cash_flow_coverage": 2.5,
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
			},
			expectedScore: 2.5,
			expectedError: false,
		},
		{
			name: "Numeric Value Found",
			input: risk.RiskFactorInput{
				FactorID: "cash_flow_coverage",
				Data: map[string]interface{}{
					"coverage_ratio": 1.8,
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
			},
			expectedScore: 1.8,
			expectedError: false,
		},
		{
			name: "No Relevant Data",
			input: risk.RiskFactorInput{
				FactorID: "cash_flow_coverage",
				Data: map[string]interface{}{
					"unrelated_field": "text",
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
			},
			expectedScore: 0,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factorDef, _ := registry.GetFactor(tt.input.FactorID)
			score, explanation, evidence, err := calculator.calculateDirectScore(tt.input, factorDef)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedScore, score)
				assert.NotEmpty(t, explanation)
				assert.NotNil(t, evidence)
			}
		})
	}
}

// TestRiskFactorCalculator_CalculateDerivedScore tests derived score calculation
func TestRiskFactorCalculator_CalculateDerivedScore(t *testing.T) {
	registry := createTestRegistry()
	calculator := risk.NewRiskFactorCalculator(registry)

	tests := []struct {
		name          string
		input         risk.RiskFactorInput
		expectedScore float64
		expectedError bool
	}{
		{
			name: "Single Value Calculation",
			input: risk.RiskFactorInput{
				FactorID: "debt_ratio",
				Data: map[string]interface{}{
					"debt_ratio": 0.5,
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
			},
			expectedScore: 0.5,
			expectedError: false,
		},
		{
			name: "Multiple Values Average",
			input: risk.RiskFactorInput{
				FactorID: "debt_ratio",
				Data: map[string]interface{}{
					"debt_ratio_1": 0.3,
					"debt_ratio_2": 0.7,
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
			},
			expectedScore: 0.5, // Average of 0.3 and 0.7
			expectedError: false,
		},
		{
			name: "No Numeric Data",
			input: risk.RiskFactorInput{
				FactorID: "debt_ratio",
				Data: map[string]interface{}{
					"text_field": "not numeric",
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
			},
			expectedScore: 0,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factorDef, _ := registry.GetFactor(tt.input.FactorID)
			score, explanation, evidence, err := calculator.calculateDerivedScore(tt.input, factorDef)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedScore, score)
				assert.NotEmpty(t, explanation)
				assert.NotNil(t, evidence)
			}
		})
	}
}

// TestRiskFactorCalculator_CalculateCompositeScore tests composite score calculation
func TestRiskFactorCalculator_CalculateCompositeScore(t *testing.T) {
	registry := createTestRegistry()
	calculator := risk.NewRiskFactorCalculator(registry)

	tests := []struct {
		name          string
		input         risk.RiskFactorInput
		expectedScore float64
		expectedError bool
	}{
		{
			name: "Multiple Components",
			input: risk.RiskFactorInput{
				FactorID: "operational_efficiency",
				Data: map[string]interface{}{
					"turnover_rate": 0.1,
					"uptime":        0.98,
					"concentration": 0.4,
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
			},
			expectedError: false,
		},
		{
			name: "No Valid Components",
			input: risk.RiskFactorInput{
				FactorID: "operational_efficiency",
				Data: map[string]interface{}{
					"text_field": "not numeric",
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
			},
			expectedScore: 0,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factorDef, _ := registry.GetFactor(tt.input.FactorID)
			score, explanation, evidence, err := calculator.calculateCompositeScore(tt.input, factorDef)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, score, 0.0)
				assert.LessOrEqual(t, score, 100.0)
				assert.NotEmpty(t, explanation)
				assert.NotNil(t, evidence)
			}
		})
	}
}

// TestRiskFactorCalculator_CalculateFinancialComponent tests financial component calculations
func TestRiskFactorCalculator_CalculateFinancialComponent(t *testing.T) {
	registry := createTestRegistry()
	calculator := risk.NewRiskFactorCalculator(registry)

	tests := []struct {
		name          string
		value         float64
		componentKey  string
		expectedScore float64
	}{
		{
			name:          "High Ratio (Low Risk)",
			value:         2.5,
			componentKey:  "debt_to_equity_ratio",
			expectedScore: 20.0,
		},
		{
			name:          "Medium Ratio (Medium Risk)",
			value:         1.2,
			componentKey:  "debt_to_equity_ratio",
			expectedScore: 60.0,
		},
		{
			name:          "Low Ratio (High Risk)",
			value:         0.5,
			componentKey:  "debt_to_equity_ratio",
			expectedScore: 80.0,
		},
		{
			name:          "High Score (Low Risk)",
			value:         85.0,
			componentKey:  "credit_score",
			expectedScore: 20.0,
		},
		{
			name:          "Positive Trend (Low Risk)",
			value:         0.1,
			componentKey:  "revenue_trend",
			expectedScore: 30.0,
		},
		{
			name:          "Negative Trend (High Risk)",
			value:         -0.2,
			componentKey:  "revenue_trend",
			expectedScore: 90.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculator.calculateFinancialComponent(tt.value, tt.componentKey)
			assert.Equal(t, tt.expectedScore, score)
		})
	}
}

// TestRiskFactorCalculator_CalculateOperationalComponent tests operational component calculations
func TestRiskFactorCalculator_CalculateOperationalComponent(t *testing.T) {
	registry := createTestRegistry()
	calculator := risk.NewRiskFactorCalculator(registry)

	tests := []struct {
		name          string
		value         float64
		componentKey  string
		expectedScore float64
	}{
		{
			name:          "Low Turnover (Low Risk)",
			value:         0.03,
			componentKey:  "employee_turnover",
			expectedScore: 20.0,
		},
		{
			name:          "High Turnover (High Risk)",
			value:         0.3,
			componentKey:  "employee_turnover",
			expectedScore: 80.0,
		},
		{
			name:          "High Uptime (Low Risk)",
			value:         0.995,
			componentKey:  "system_uptime",
			expectedScore: 20.0,
		},
		{
			name:          "Low Uptime (High Risk)",
			value:         0.85,
			componentKey:  "system_uptime",
			expectedScore: 80.0,
		},
		{
			name:          "Low Concentration (Low Risk)",
			value:         0.2,
			componentKey:  "customer_concentration",
			expectedScore: 20.0,
		},
		{
			name:          "High Concentration (High Risk)",
			value:         0.8,
			componentKey:  "customer_concentration",
			expectedScore: 80.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculator.calculateOperationalComponent(tt.value, tt.componentKey)
			assert.Equal(t, tt.expectedScore, score)
		})
	}
}

// TestRiskFactorCalculator_CalculateRegulatoryComponent tests regulatory component calculations
func TestRiskFactorCalculator_CalculateRegulatoryComponent(t *testing.T) {
	registry := createTestRegistry()
	calculator := risk.NewRiskFactorCalculator(registry)

	tests := []struct {
		name          string
		value         float64
		componentKey  string
		expectedScore float64
	}{
		{
			name:          "High Compliance (Low Risk)",
			value:         95.0,
			componentKey:  "compliance_score",
			expectedScore: 20.0,
		},
		{
			name:          "Low Compliance (High Risk)",
			value:         50.0,
			componentKey:  "compliance_score",
			expectedScore: 80.0,
		},
		{
			name:          "No Violations (Low Risk)",
			value:         0.0,
			componentKey:  "regulatory_violations",
			expectedScore: 20.0,
		},
		{
			name:          "Many Violations (High Risk)",
			value:         5.0,
			componentKey:  "regulatory_violations",
			expectedScore: 80.0,
		},
		{
			name:          "Valid License (Low Risk)",
			value:         1.0,
			componentKey:  "license_status",
			expectedScore: 20.0,
		},
		{
			name:          "Invalid License (High Risk)",
			value:         0.3,
			componentKey:  "license_status",
			expectedScore: 80.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculator.calculateRegulatoryComponent(tt.value, tt.componentKey)
			assert.Equal(t, tt.expectedScore, score)
		})
	}
}

// TestRiskFactorCalculator_CalculateReputationalComponent tests reputational component calculations
func TestRiskFactorCalculator_CalculateReputationalComponent(t *testing.T) {
	registry := createTestRegistry()
	calculator := risk.NewRiskFactorCalculator(registry)

	tests := []struct {
		name          string
		value         float64
		componentKey  string
		expectedScore float64
	}{
		{
			name:          "High Sentiment (Low Risk)",
			value:         0.8,
			componentKey:  "customer_sentiment",
			expectedScore: 20.0,
		},
		{
			name:          "Low Sentiment (High Risk)",
			value:         0.2,
			componentKey:  "customer_sentiment",
			expectedScore: 80.0,
		},
		{
			name:          "High Satisfaction (Low Risk)",
			value:         4.5,
			componentKey:  "customer_satisfaction",
			expectedScore: 20.0,
		},
		{
			name:          "Low Satisfaction (High Risk)",
			value:         2.5,
			componentKey:  "customer_satisfaction",
			expectedScore: 80.0,
		},
		{
			name:          "No Negative Mentions (Low Risk)",
			value:         0.0,
			componentKey:  "negative_mentions",
			expectedScore: 20.0,
		},
		{
			name:          "Many Negative Mentions (High Risk)",
			value:         20.0,
			componentKey:  "negative_mentions",
			expectedScore: 80.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculator.calculateReputationalComponent(tt.value, tt.componentKey)
			assert.Equal(t, tt.expectedScore, score)
		})
	}
}

// TestRiskFactorCalculator_CalculateCybersecurityComponent tests cybersecurity component calculations
func TestRiskFactorCalculator_CalculateCybersecurityComponent(t *testing.T) {
	registry := createTestRegistry()
	calculator := risk.NewRiskFactorCalculator(registry)

	tests := []struct {
		name          string
		value         float64
		componentKey  string
		expectedScore float64
	}{
		{
			name:          "High Security Score (Low Risk)",
			value:         90.0,
			componentKey:  "security_score",
			expectedScore: 20.0,
		},
		{
			name:          "Low Security Score (High Risk)",
			value:         40.0,
			componentKey:  "security_score",
			expectedScore: 80.0,
		},
		{
			name:          "No Breaches (Low Risk)",
			value:         0.0,
			componentKey:  "security_breaches",
			expectedScore: 20.0,
		},
		{
			name:          "Many Breaches (High Risk)",
			value:         5.0,
			componentKey:  "security_breaches",
			expectedScore: 80.0,
		},
		{
			name:          "High Patch Compliance (Low Risk)",
			value:         0.98,
			componentKey:  "patch_compliance",
			expectedScore: 20.0,
		},
		{
			name:          "Low Patch Compliance (High Risk)",
			value:         0.60,
			componentKey:  "patch_compliance",
			expectedScore: 80.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculator.calculateCybersecurityComponent(tt.value, tt.componentKey)
			assert.Equal(t, tt.expectedScore, score)
		})
	}
}

// TestRiskFactorCalculator_NormalizeScore tests score normalization
func TestRiskFactorCalculator_NormalizeScore(t *testing.T) {
	registry := createTestRegistry()
	calculator := risk.NewRiskFactorCalculator(registry)

	tests := []struct {
		name          string
		rawScore      float64
		factorID      string
		expectedScore float64
	}{
		{
			name:          "Score Within Range",
			rawScore:      50.0,
			factorID:      "cash_flow_coverage",
			expectedScore: 50.0,
		},
		{
			name:          "Score Above Range",
			rawScore:      150.0,
			factorID:      "cash_flow_coverage",
			expectedScore: 100.0,
		},
		{
			name:          "Score Below Range",
			rawScore:      -10.0,
			factorID:      "cash_flow_coverage",
			expectedScore: 0.0,
		},
		{
			name:          "High Coverage (Low Risk)",
			rawScore:      3.0,
			factorID:      "cash_flow_coverage",
			expectedScore: 25.0, // Should be normalized to low risk
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factorDef, _ := registry.GetFactor(tt.factorID)
			normalizedScore := calculator.normalizeScore(tt.rawScore, factorDef)
			assert.Equal(t, tt.expectedScore, normalizedScore)
		})
	}
}

// TestRiskFactorCalculator_DetermineRiskLevel tests risk level determination
func TestRiskFactorCalculator_DetermineRiskLevel(t *testing.T) {
	registry := createTestRegistry()
	calculator := risk.NewRiskFactorCalculator(registry)

	tests := []struct {
		name          string
		score         float64
		expectedLevel risk.RiskLevel
	}{
		{
			name:          "Low Risk",
			score:         20.0,
			expectedLevel: risk.RiskLevelLow,
		},
		{
			name:          "Medium Risk",
			score:         40.0,
			expectedLevel: risk.RiskLevelMedium,
		},
		{
			name:          "High Risk",
			score:         60.0,
			expectedLevel: risk.RiskLevelHigh,
		},
		{
			name:          "Critical Risk",
			score:         80.0,
			expectedLevel: risk.RiskLevelCritical,
		},
		{
			name:          "Boundary Low-Medium",
			score:         25.0,
			expectedLevel: risk.RiskLevelLow,
		},
		{
			name:          "Boundary Medium-High",
			score:         50.0,
			expectedLevel: risk.RiskLevelMedium,
		},
		{
			name:          "Boundary High-Critical",
			score:         75.0,
			expectedLevel: risk.RiskLevelHigh,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factorDef, _ := registry.GetFactor("cash_flow_coverage")
			level := calculator.determineRiskLevel(tt.score, factorDef.Thresholds)
			assert.Equal(t, tt.expectedLevel, level)
		})
	}
}

// TestRiskFactorCalculator_CalculateConfidence tests confidence calculation
func TestRiskFactorCalculator_CalculateConfidence(t *testing.T) {
	registry := createTestRegistry()
	calculator := risk.NewRiskFactorCalculator(registry)

	tests := []struct {
		name               string
		input              risk.RiskFactorInput
		expectedConfidence float64
	}{
		{
			name: "High Reliability, Recent Data",
			input: risk.RiskFactorInput{
				FactorID:    "cash_flow_coverage",
				Data:        map[string]interface{}{"coverage": 2.0},
				Timestamp:   time.Now().Add(-1 * time.Hour), // 1 hour ago
				Source:      "test",
				Reliability: 0.9,
			},
			expectedConfidence: 0.9,
		},
		{
			name: "Low Reliability, Old Data",
			input: risk.RiskFactorInput{
				FactorID:    "cash_flow_coverage",
				Data:        map[string]interface{}{"coverage": 2.0},
				Timestamp:   time.Now().Add(-60 * 24 * time.Hour), // 60 days ago
				Source:      "test",
				Reliability: 0.5,
			},
			expectedConfidence: 0.3, // 0.5 * 0.6 (old data penalty)
		},
		{
			name: "Incomplete Data",
			input: risk.RiskFactorInput{
				FactorID:    "cash_flow_coverage",
				Data:        map[string]interface{}{"coverage": 2.0},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.8,
			},
			expectedConfidence: 0.8, // Should be adjusted based on data completeness
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factorDef, _ := registry.GetFactor(tt.input.FactorID)
			confidence := calculator.calculateConfidence(tt.input, factorDef)
			assert.InDelta(t, tt.expectedConfidence, confidence, 0.1)
			assert.GreaterOrEqual(t, confidence, 0.0)
			assert.LessOrEqual(t, confidence, 1.0)
		})
	}
}

// TestRiskFactorCalculator_ToFloat64 tests type conversion
func TestRiskFactorCalculator_ToFloat64(t *testing.T) {
	registry := createTestRegistry()
	calculator := risk.NewRiskFactorCalculator(registry)

	tests := []struct {
		name          string
		value         interface{}
		expectedFloat float64
		expectedOk    bool
	}{
		{
			name:          "Float64",
			value:         float64(42.5),
			expectedFloat: 42.5,
			expectedOk:    true,
		},
		{
			name:          "Float32",
			value:         float32(42.5),
			expectedFloat: 42.5,
			expectedOk:    true,
		},
		{
			name:          "Int",
			value:         int(42),
			expectedFloat: 42.0,
			expectedOk:    true,
		},
		{
			name:          "Int32",
			value:         int32(42),
			expectedFloat: 42.0,
			expectedOk:    true,
		},
		{
			name:          "Int64",
			value:         int64(42),
			expectedFloat: 42.0,
			expectedOk:    true,
		},
		{
			name:          "Valid String",
			value:         "42.5",
			expectedFloat: 42.5,
			expectedOk:    true,
		},
		{
			name:          "Invalid String",
			value:         "not a number",
			expectedFloat: 0.0,
			expectedOk:    false,
		},
		{
			name:          "Boolean",
			value:         true,
			expectedFloat: 0.0,
			expectedOk:    false,
		},
		{
			name:          "Nil",
			value:         nil,
			expectedFloat: 0.0,
			expectedOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := calculator.toFloat64(tt.value)
			assert.Equal(t, tt.expectedOk, ok)
			assert.Equal(t, tt.expectedFloat, result)
		})
	}
}

// TestRiskFactorCalculator_ApplyFormula tests formula application
func TestRiskFactorCalculator_ApplyFormula(t *testing.T) {
	registry := createTestRegistry()
	calculator := risk.NewRiskFactorCalculator(registry)

	tests := []struct {
		name          string
		values        []float64
		formula       string
		expectedScore float64
		expectedError bool
	}{
		{
			name:          "Average Formula",
			values:        []float64{10.0, 20.0, 30.0},
			formula:       "average",
			expectedScore: 20.0,
			expectedError: false,
		},
		{
			name:          "Sum Formula",
			values:        []float64{10.0, 20.0, 30.0},
			formula:       "sum",
			expectedScore: 60.0,
			expectedError: false,
		},
		{
			name:          "Max Formula",
			values:        []float64{10.0, 20.0, 30.0},
			formula:       "max",
			expectedScore: 30.0,
			expectedError: false,
		},
		{
			name:          "Min Formula",
			values:        []float64{10.0, 20.0, 30.0},
			formula:       "min",
			expectedScore: 10.0,
			expectedError: false,
		},
		{
			name:          "Unknown Formula (Default to Average)",
			values:        []float64{10.0, 20.0, 30.0},
			formula:       "unknown",
			expectedScore: 20.0,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evidence := []string{"test evidence"}
			score, explanation, resultEvidence, err := calculator.applyFormula(tt.values, tt.formula, evidence)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedScore, score)
				assert.NotEmpty(t, explanation)
				assert.Equal(t, evidence, resultEvidence)
			}
		})
	}
}

// TestRiskFactorCalculator_ValidateInput tests input validation
func TestRiskFactorCalculator_ValidateInput(t *testing.T) {
	registry := createTestRegistry()
	calculator := risk.NewRiskFactorCalculator(registry)

	tests := []struct {
		name          string
		input         risk.RiskFactorInput
		expectedError bool
	}{
		{
			name: "Valid Input",
			input: risk.RiskFactorInput{
				FactorID:    "cash_flow_coverage",
				Data:        map[string]interface{}{"coverage": 2.0},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
			},
			expectedError: false,
		},
		{
			name: "Invalid Reliability - Too High",
			input: risk.RiskFactorInput{
				FactorID:    "cash_flow_coverage",
				Data:        map[string]interface{}{"coverage": 2.0},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 1.5,
			},
			expectedError: true,
		},
		{
			name: "Invalid Reliability - Too Low",
			input: risk.RiskFactorInput{
				FactorID:    "cash_flow_coverage",
				Data:        map[string]interface{}{"coverage": 2.0},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: -0.1,
			},
			expectedError: true,
		},
		{
			name: "Missing Required Data Source",
			input: risk.RiskFactorInput{
				FactorID:    "cash_flow_coverage",
				Data:        map[string]interface{}{}, // Empty data
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
			},
			expectedError: false, // Should not error for empty data
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factorDef, _ := registry.GetFactor(tt.input.FactorID)
			err := calculator.validateInput(tt.input, factorDef)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to create a test registry with sample data
func createTestRegistry() *risk.RiskCategoryRegistry {
	registry := risk.NewRiskCategoryRegistry()

	// Create test factor definitions
	factors := []*risk.RiskFactorDefinition{
		{
			ID:              "cash_flow_coverage",
			Name:            "Cash Flow Coverage",
			Description:     "Ratio of cash flow to debt obligations",
			Category:        risk.RiskCategoryFinancial,
			Subcategory:     "liquidity",
			Weight:          0.3,
			CalculationType: "direct",
			DataSources:     []string{"cash_flow", "debt_obligations"},
			Thresholds: map[risk.RiskLevel]float64{
				risk.RiskLevelLow:      2.0,
				risk.RiskLevelMedium:   1.5,
				risk.RiskLevelHigh:     1.0,
				risk.RiskLevelCritical: 0.5,
			},
			Formula:   "",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:              "debt_ratio",
			Name:            "Debt to Equity Ratio",
			Description:     "Ratio of total debt to total equity",
			Category:        risk.RiskCategoryFinancial,
			Subcategory:     "leverage",
			Weight:          0.25,
			CalculationType: "derived",
			DataSources:     []string{"total_debt", "total_equity"},
			Thresholds: map[risk.RiskLevel]float64{
				risk.RiskLevelLow:      0.3,
				risk.RiskLevelMedium:   0.6,
				risk.RiskLevelHigh:     1.0,
				risk.RiskLevelCritical: 2.0,
			},
			Formula:   "average",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:              "operational_efficiency",
			Name:            "Operational Efficiency",
			Description:     "Composite measure of operational performance",
			Category:        risk.RiskCategoryOperational,
			Subcategory:     "performance",
			Weight:          0.2,
			CalculationType: "composite",
			DataSources:     []string{"turnover_rate", "uptime", "concentration"},
			Thresholds: map[risk.RiskLevel]float64{
				risk.RiskLevelLow:      20.0,
				risk.RiskLevelMedium:   40.0,
				risk.RiskLevelHigh:     60.0,
				risk.RiskLevelCritical: 80.0,
			},
			Formula:   "average",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Register factors directly
	for _, factor := range factors {
		registry.RegisterFactor(factor)
	}

	return registry
}
