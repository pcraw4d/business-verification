package industry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

func TestIndustryModelManager(t *testing.T) {
	logger := zap.NewNop()
	manager := NewIndustryModelManager(logger)

	t.Run("initialization", func(t *testing.T) {
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.models)
		assert.Equal(t, 9, len(manager.models)) // 8 industry models + general
	})

	t.Run("get industry model", func(t *testing.T) {
		// Test existing industry
		fintechModel := manager.GetIndustryModel(IndustryFintech)
		assert.NotNil(t, fintechModel)
		assert.Equal(t, IndustryFintech, fintechModel.GetIndustryType())

		// Test non-existing industry (should return general)
		unknownModel := manager.GetIndustryModel("unknown")
		assert.NotNil(t, unknownModel)
		assert.Equal(t, IndustryGeneral, unknownModel.GetIndustryType())
	})

	t.Run("detect industry type", func(t *testing.T) {
		tests := []struct {
			name     string
			business *models.RiskAssessmentRequest
			expected IndustryType
		}{
			{
				name: "fintech business",
				business: &models.RiskAssessmentRequest{
					BusinessName: "Test Fintech",
					Industry:     "fintech",
				},
				expected: IndustryFintech,
			},
			{
				name: "healthcare business",
				business: &models.RiskAssessmentRequest{
					BusinessName: "Test Healthcare",
					Industry:     "healthcare",
				},
				expected: IndustryHealthcare,
			},
			{
				name: "technology business",
				business: &models.RiskAssessmentRequest{
					BusinessName: "Test Tech",
					Industry:     "technology",
				},
				expected: IndustryTechnology,
			},
			{
				name: "retail business",
				business: &models.RiskAssessmentRequest{
					BusinessName: "Test Retail",
					Industry:     "retail",
				},
				expected: IndustryRetail,
			},
			{
				name: "manufacturing business",
				business: &models.RiskAssessmentRequest{
					BusinessName: "Test Manufacturing",
					Industry:     "manufacturing",
				},
				expected: IndustryManufacturing,
			},
			{
				name: "real estate business",
				business: &models.RiskAssessmentRequest{
					BusinessName: "Test Real Estate",
					Industry:     "real estate",
				},
				expected: IndustryRealEstate,
			},
			{
				name: "energy business",
				business: &models.RiskAssessmentRequest{
					BusinessName: "Test Energy",
					Industry:     "energy",
				},
				expected: IndustryEnergy,
			},
			{
				name: "transportation business",
				business: &models.RiskAssessmentRequest{
					BusinessName: "Test Transportation",
					Industry:     "transportation",
				},
				expected: IndustryTransportation,
			},
			{
				name: "unknown business",
				business: &models.RiskAssessmentRequest{
					BusinessName: "Test Unknown",
					Industry:     "unknown",
				},
				expected: IndustryGeneral,
			},
			{
				name:     "nil business",
				business: nil,
				expected: IndustryGeneral,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := manager.DetectIndustryType(tt.business)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("get all industry types", func(t *testing.T) {
		types := manager.GetAllIndustryTypes()
		assert.Equal(t, 9, len(types))

		expectedTypes := []IndustryType{
			IndustryFintech, IndustryHealthcare, IndustryTechnology,
			IndustryRetail, IndustryManufacturing, IndustryRealEstate,
			IndustryEnergy, IndustryTransportation, IndustryGeneral,
		}

		for _, expectedType := range expectedTypes {
			assert.Contains(t, types, expectedType)
		}
	})

	t.Run("get industry model info", func(t *testing.T) {
		info := manager.GetIndustryModelInfo(IndustryFintech)
		assert.NotNil(t, info)
		assert.Equal(t, string(IndustryFintech), info["industry_type"])
		assert.Equal(t, true, info["model_available"])
		assert.Greater(t, info["risk_factors"], 0)
		assert.Greater(t, info["compliance_requirements"], 0)
		assert.NotNil(t, info["weightings"])
	})
}

func TestFintechModel(t *testing.T) {
	logger := zap.NewNop()
	model := NewFintechModel(logger)

	t.Run("get industry type", func(t *testing.T) {
		assert.Equal(t, IndustryFintech, model.GetIndustryType())
	})

	t.Run("get industry specific factors", func(t *testing.T) {
		factors := model.GetIndustrySpecificFactors()
		assert.Greater(t, len(factors), 0)

		// Check for key fintech factors
		factorNames := make([]string, len(factors))
		for i, factor := range factors {
			factorNames[i] = factor.FactorID
		}

		assert.Contains(t, factorNames, "fintech_regulatory_compliance")
		assert.Contains(t, factorNames, "fintech_cybersecurity")
		assert.Contains(t, factorNames, "fintech_anti_money_laundering")
	})

	t.Run("get industry weightings", func(t *testing.T) {
		weightings := model.GetIndustryWeightings()
		assert.NotNil(t, weightings)
		assert.Equal(t, 1.0, weightings["regulatory"]+weightings["compliance"]+weightings["operational"]+weightings["financial"]+weightings["reputational"]+weightings["technology"]+weightings["geopolitical"]+weightings["environmental"])
		assert.Greater(t, weightings["regulatory"], 0.0)
		assert.Greater(t, weightings["compliance"], 0.0)
	})

	t.Run("validate industry data", func(t *testing.T) {
		tests := []struct {
			name     string
			business *models.RiskAssessmentRequest
			expected int // expected number of errors
		}{
			{
				name:     "nil business",
				business: nil,
				expected: 1,
			},
			{
				name: "valid business",
				business: &models.RiskAssessmentRequest{
					BusinessName:    "Test Fintech",
					BusinessAddress: "123 Test St",
				},
				expected: 0,
			},
			{
				name: "business without address",
				business: &models.RiskAssessmentRequest{
					BusinessName: "Test Fintech",
				},
				expected: 1,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				errors := model.ValidateIndustryData(tt.business)
				assert.Equal(t, tt.expected, len(errors))
			})
		}
	})

	t.Run("get compliance requirements", func(t *testing.T) {
		requirements := model.GetIndustryComplianceRequirements()
		assert.Greater(t, len(requirements), 0)

		// Check for key fintech requirements
		requirementNames := make([]string, len(requirements))
		for i, req := range requirements {
			requirementNames[i] = req.RequirementID
		}

		assert.Contains(t, requirementNames, "fintech_licensing")
		assert.Contains(t, requirementNames, "fintech_aml")
		assert.Contains(t, requirementNames, "fintech_data_protection")
	})

	t.Run("calculate industry risk", func(t *testing.T) {
		business := &models.RiskAssessmentRequest{
			BusinessName:    "Test Fintech",
			BusinessAddress: "123 Test St",
			Industry:        "fintech",
		}

		result, err := model.CalculateIndustryRisk(context.Background(), business)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, IndustryFintech, result.IndustryType)
		assert.GreaterOrEqual(t, result.IndustryRiskScore, 0.0)
		assert.LessOrEqual(t, result.IndustryRiskScore, 1.0)
		assert.NotNil(t, result.IndustryRiskLevel)
		assert.Greater(t, len(result.IndustryFactors), 0)
		assert.Greater(t, len(result.ComplianceStatus), 0)
		assert.Greater(t, len(result.IndustryRecommendations), 0)
		assert.Greater(t, len(result.RegulatoryFactors), 0)
		assert.Greater(t, len(result.MarketFactors), 0)
		assert.Greater(t, len(result.OperationalFactors), 0)
		assert.GreaterOrEqual(t, result.ConfidenceScore, 0.0)
		assert.LessOrEqual(t, result.ConfidenceScore, 1.0)
	})
}

func TestHealthcareModel(t *testing.T) {
	logger := zap.NewNop()
	model := NewHealthcareModel(logger)

	t.Run("get industry type", func(t *testing.T) {
		assert.Equal(t, IndustryHealthcare, model.GetIndustryType())
	})

	t.Run("calculate industry risk", func(t *testing.T) {
		business := &models.RiskAssessmentRequest{
			BusinessName:    "Test Healthcare",
			BusinessAddress: "123 Test St",
			Industry:        "healthcare",
		}

		result, err := model.CalculateIndustryRisk(context.Background(), business)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, IndustryHealthcare, result.IndustryType)
		assert.GreaterOrEqual(t, result.IndustryRiskScore, 0.0)
		assert.LessOrEqual(t, result.IndustryRiskScore, 1.0)
		assert.NotNil(t, result.IndustryRiskLevel)
		assert.Greater(t, len(result.IndustryFactors), 0)
	})

	t.Run("get compliance requirements", func(t *testing.T) {
		requirements := model.GetIndustryComplianceRequirements()
		assert.Greater(t, len(requirements), 0)

		// Check for key healthcare requirements
		requirementNames := make([]string, len(requirements))
		for i, req := range requirements {
			requirementNames[i] = req.RequirementID
		}

		assert.Contains(t, requirementNames, "healthcare_hipaa")
		assert.Contains(t, requirementNames, "healthcare_medical_licensing")
	})
}

func TestTechnologyModel(t *testing.T) {
	logger := zap.NewNop()
	model := NewTechnologyModel(logger)

	t.Run("get industry type", func(t *testing.T) {
		assert.Equal(t, IndustryTechnology, model.GetIndustryType())
	})

	t.Run("calculate industry risk", func(t *testing.T) {
		business := &models.RiskAssessmentRequest{
			BusinessName:    "Test Technology",
			BusinessAddress: "123 Test St",
			Industry:        "technology",
		}

		result, err := model.CalculateIndustryRisk(context.Background(), business)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, IndustryTechnology, result.IndustryType)
		assert.GreaterOrEqual(t, result.IndustryRiskScore, 0.0)
		assert.LessOrEqual(t, result.IndustryRiskScore, 1.0)
		assert.NotNil(t, result.IndustryRiskLevel)
		assert.Greater(t, len(result.IndustryFactors), 0)
	})

	t.Run("get compliance requirements", func(t *testing.T) {
		requirements := model.GetIndustryComplianceRequirements()
		assert.Greater(t, len(requirements), 0)

		// Check for key technology requirements
		requirementNames := make([]string, len(requirements))
		for i, req := range requirements {
			requirementNames[i] = req.RequirementID
		}

		assert.Contains(t, requirementNames, "technology_gdpr")
		assert.Contains(t, requirementNames, "technology_ccpa")
	})
}

func TestRetailModel(t *testing.T) {
	logger := zap.NewNop()
	model := NewRetailModel(logger)

	t.Run("get industry type", func(t *testing.T) {
		assert.Equal(t, IndustryRetail, model.GetIndustryType())
	})

	t.Run("calculate industry risk", func(t *testing.T) {
		business := &models.RiskAssessmentRequest{
			BusinessName:    "Test Retail",
			BusinessAddress: "123 Test St",
			Industry:        "retail",
		}

		result, err := model.CalculateIndustryRisk(context.Background(), business)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, IndustryRetail, result.IndustryType)
		assert.GreaterOrEqual(t, result.IndustryRiskScore, 0.0)
		assert.LessOrEqual(t, result.IndustryRiskScore, 1.0)
		assert.NotNil(t, result.IndustryRiskLevel)
		assert.Greater(t, len(result.IndustryFactors), 0)
	})
}

func TestManufacturingModel(t *testing.T) {
	logger := zap.NewNop()
	model := NewManufacturingModel(logger)

	t.Run("get industry type", func(t *testing.T) {
		assert.Equal(t, IndustryManufacturing, model.GetIndustryType())
	})

	t.Run("calculate industry risk", func(t *testing.T) {
		business := &models.RiskAssessmentRequest{
			BusinessName:    "Test Manufacturing",
			BusinessAddress: "123 Test St",
			Industry:        "manufacturing",
		}

		result, err := model.CalculateIndustryRisk(context.Background(), business)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, IndustryManufacturing, result.IndustryType)
		assert.GreaterOrEqual(t, result.IndustryRiskScore, 0.0)
		assert.LessOrEqual(t, result.IndustryRiskScore, 1.0)
		assert.NotNil(t, result.IndustryRiskLevel)
		assert.Greater(t, len(result.IndustryFactors), 0)
	})
}

func TestRealEstateModel(t *testing.T) {
	logger := zap.NewNop()
	model := NewRealEstateModel(logger)

	t.Run("get industry type", func(t *testing.T) {
		assert.Equal(t, IndustryRealEstate, model.GetIndustryType())
	})

	t.Run("calculate industry risk", func(t *testing.T) {
		business := &models.RiskAssessmentRequest{
			BusinessName:    "Test Real Estate",
			BusinessAddress: "123 Test St",
			Industry:        "real estate",
		}

		result, err := model.CalculateIndustryRisk(context.Background(), business)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, IndustryRealEstate, result.IndustryType)
		assert.GreaterOrEqual(t, result.IndustryRiskScore, 0.0)
		assert.LessOrEqual(t, result.IndustryRiskScore, 1.0)
		assert.NotNil(t, result.IndustryRiskLevel)
		assert.Greater(t, len(result.IndustryFactors), 0)
	})
}

func TestEnergyModel(t *testing.T) {
	logger := zap.NewNop()
	model := NewEnergyModel(logger)

	t.Run("get industry type", func(t *testing.T) {
		assert.Equal(t, IndustryEnergy, model.GetIndustryType())
	})

	t.Run("calculate industry risk", func(t *testing.T) {
		business := &models.RiskAssessmentRequest{
			BusinessName:    "Test Energy",
			BusinessAddress: "123 Test St",
			Industry:        "energy",
		}

		result, err := model.CalculateIndustryRisk(context.Background(), business)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, IndustryEnergy, result.IndustryType)
		assert.GreaterOrEqual(t, result.IndustryRiskScore, 0.0)
		assert.LessOrEqual(t, result.IndustryRiskScore, 1.0)
		assert.NotNil(t, result.IndustryRiskLevel)
		assert.Greater(t, len(result.IndustryFactors), 0)
	})
}

func TestTransportationModel(t *testing.T) {
	logger := zap.NewNop()
	model := NewTransportationModel(logger)

	t.Run("get industry type", func(t *testing.T) {
		assert.Equal(t, IndustryTransportation, model.GetIndustryType())
	})

	t.Run("calculate industry risk", func(t *testing.T) {
		business := &models.RiskAssessmentRequest{
			BusinessName:    "Test Transportation",
			BusinessAddress: "123 Test St",
			Industry:        "transportation",
		}

		result, err := model.CalculateIndustryRisk(context.Background(), business)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, IndustryTransportation, result.IndustryType)
		assert.GreaterOrEqual(t, result.IndustryRiskScore, 0.0)
		assert.LessOrEqual(t, result.IndustryRiskScore, 1.0)
		assert.NotNil(t, result.IndustryRiskLevel)
		assert.Greater(t, len(result.IndustryFactors), 0)
	})
}

func TestGeneralModel(t *testing.T) {
	logger := zap.NewNop()
	model := NewGeneralModel(logger)

	t.Run("get industry type", func(t *testing.T) {
		assert.Equal(t, IndustryGeneral, model.GetIndustryType())
	})

	t.Run("calculate industry risk", func(t *testing.T) {
		business := &models.RiskAssessmentRequest{
			BusinessName:    "Test General",
			BusinessAddress: "123 Test St",
			Industry:        "general",
		}

		result, err := model.CalculateIndustryRisk(context.Background(), business)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, IndustryGeneral, result.IndustryType)
		assert.GreaterOrEqual(t, result.IndustryRiskScore, 0.0)
		assert.LessOrEqual(t, result.IndustryRiskScore, 1.0)
		assert.NotNil(t, result.IndustryRiskLevel)
		assert.Greater(t, len(result.IndustryFactors), 0)
	})
}

func TestIndustryModelManager_AnalyzeIndustryRisk(t *testing.T) {
	logger := zap.NewNop()
	manager := NewIndustryModelManager(logger)

	t.Run("analyze fintech risk", func(t *testing.T) {
		business := &models.RiskAssessmentRequest{
			BusinessName:    "Test Fintech",
			BusinessAddress: "123 Test St",
			Industry:        "fintech",
		}

		result, err := manager.AnalyzeIndustryRisk(context.Background(), business)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, IndustryFintech, result.IndustryType)
		assert.GreaterOrEqual(t, result.IndustryRiskScore, 0.0)
		assert.LessOrEqual(t, result.IndustryRiskScore, 1.0)
		assert.NotNil(t, result.IndustryRiskLevel)
		assert.Greater(t, len(result.IndustryFactors), 0)
	})

	t.Run("analyze unknown industry risk", func(t *testing.T) {
		business := &models.RiskAssessmentRequest{
			BusinessName:    "Test Unknown",
			BusinessAddress: "123 Test St",
			Industry:        "unknown",
		}

		result, err := manager.AnalyzeIndustryRisk(context.Background(), business)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, IndustryGeneral, result.IndustryType)
		assert.GreaterOrEqual(t, result.IndustryRiskScore, 0.0)
		assert.LessOrEqual(t, result.IndustryRiskScore, 1.0)
		assert.NotNil(t, result.IndustryRiskLevel)
		assert.Greater(t, len(result.IndustryFactors), 0)
	})
}

func TestIndustryRiskResult_Structure(t *testing.T) {
	logger := zap.NewNop()
	model := NewFintechModel(logger)

	business := &models.RiskAssessmentRequest{
		BusinessName:    "Test Fintech",
		BusinessAddress: "123 Test St",
		Industry:        "fintech",
	}

	result, err := model.CalculateIndustryRisk(context.Background(), business)
	require.NoError(t, err)
	require.NotNil(t, result)

	t.Run("industry factors structure", func(t *testing.T) {
		for _, factor := range result.IndustryFactors {
			assert.NotEmpty(t, factor.FactorID)
			assert.NotEmpty(t, factor.FactorName)
			assert.NotEmpty(t, factor.FactorCategory)
			assert.GreaterOrEqual(t, factor.RiskScore, 0.0)
			assert.LessOrEqual(t, factor.RiskScore, 1.0)
			assert.NotEmpty(t, factor.RiskLevel)
			assert.NotEmpty(t, factor.Description)
			assert.NotEmpty(t, factor.Impact)
			assert.NotEmpty(t, factor.Likelihood)
			assert.NotEmpty(t, factor.MitigationAdvice)
		}
	})

	t.Run("compliance status structure", func(t *testing.T) {
		for _, status := range result.ComplianceStatus {
			assert.NotEmpty(t, status.RequirementID)
			assert.NotEmpty(t, status.Status)
			assert.GreaterOrEqual(t, status.ComplianceScore, 0.0)
			assert.LessOrEqual(t, status.ComplianceScore, 1.0)
			assert.NotNil(t, status.Issues)
			assert.NotNil(t, status.Recommendations)
		}
	})

	t.Run("recommendations structure", func(t *testing.T) {
		for _, rec := range result.IndustryRecommendations {
			assert.NotEmpty(t, rec.RecommendationID)
			assert.NotEmpty(t, rec.Category)
			assert.NotEmpty(t, rec.Priority)
			assert.NotEmpty(t, rec.Title)
			assert.NotEmpty(t, rec.Description)
			assert.Greater(t, len(rec.ActionItems), 0)
			assert.NotEmpty(t, rec.ExpectedBenefit)
			assert.NotEmpty(t, rec.ImplementationCost)
			assert.NotEmpty(t, rec.Timeline)
		}
	})

	t.Run("regulatory factors structure", func(t *testing.T) {
		for _, factor := range result.RegulatoryFactors {
			assert.NotEmpty(t, factor.FactorID)
			assert.NotEmpty(t, factor.RegulationName)
			assert.NotEmpty(t, factor.RegulatoryBody)
			assert.NotEmpty(t, factor.Jurisdiction)
			assert.GreaterOrEqual(t, factor.RiskImpact, 0.0)
			assert.LessOrEqual(t, factor.RiskImpact, 1.0)
			assert.NotEmpty(t, factor.ComplianceCost)
			assert.NotEmpty(t, factor.PenaltyRisk)
			assert.NotEmpty(t, factor.Description)
		}
	})

	t.Run("market factors structure", func(t *testing.T) {
		for _, factor := range result.MarketFactors {
			assert.NotEmpty(t, factor.FactorID)
			assert.NotEmpty(t, factor.FactorName)
			assert.NotEmpty(t, factor.MarketTrend)
			assert.GreaterOrEqual(t, factor.ImpactScore, 0.0)
			assert.LessOrEqual(t, factor.ImpactScore, 1.0)
			assert.NotEmpty(t, factor.TimeHorizon)
			assert.NotEmpty(t, factor.Description)
			assert.Greater(t, len(factor.KeyDrivers), 0)
			assert.Greater(t, len(factor.RiskMitigation), 0)
		}
	})

	t.Run("operational factors structure", func(t *testing.T) {
		for _, factor := range result.OperationalFactors {
			assert.NotEmpty(t, factor.FactorID)
			assert.NotEmpty(t, factor.FactorName)
			assert.NotEmpty(t, factor.OperationalArea)
			assert.GreaterOrEqual(t, factor.RiskScore, 0.0)
			assert.LessOrEqual(t, factor.RiskScore, 1.0)
			assert.NotEmpty(t, factor.Criticality)
			assert.NotEmpty(t, factor.Description)
			assert.Greater(t, len(factor.ControlMeasures), 0)
			assert.NotEmpty(t, factor.MonitoringFrequency)
		}
	})
}
