package external

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/external/ofac"
	"kyb-platform/services/risk-assessment-service/internal/external/thomson_reuters"
	"kyb-platform/services/risk-assessment-service/internal/external/worldcheck"
	"kyb-platform/services/risk-assessment-service/internal/models"
)

func TestExternalAPIManager_Initialization(t *testing.T) {
	logger := zap.NewNop()

	config := &ExternalAPIManagerConfig{
		ThomsonReuters: &ThomsonReutersConfig{
			APIKey:    "test_key",
			BaseURL:   "https://api.test.com",
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   true,
		},
		OFAC: &OFACConfig{
			APIKey:    "test_key",
			BaseURL:   "https://api.test.com",
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   true,
		},
		WorldCheck: &WorldCheckConfig{
			APIKey:    "test_key",
			BaseURL:   "https://api.test.com",
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   true,
		},
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		EnableCache: true,
		CacheTTL:    1 * time.Hour,
	}

	manager := NewExternalAPIManager(config, logger)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.thomsonReuters)
	assert.NotNil(t, manager.ofac)
	assert.NotNil(t, manager.worldCheck)
	assert.Equal(t, config, manager.config)
}

func TestExternalAPIManager_GetComprehensiveData(t *testing.T) {
	logger := zap.NewNop()

	config := &ExternalAPIManagerConfig{
		ThomsonReuters: &ThomsonReutersConfig{
			APIKey:    "test_key",
			BaseURL:   "https://api.test.com",
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   true,
		},
		OFAC: &OFACConfig{
			APIKey:    "test_key",
			BaseURL:   "https://api.test.com",
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   true,
		},
		WorldCheck: &WorldCheckConfig{
			APIKey:    "test_key",
			BaseURL:   "https://api.test.com",
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   true,
		},
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		EnableCache: true,
		CacheTTL:    1 * time.Hour,
	}

	manager := NewExternalAPIManager(config, logger)

	business := &models.RiskAssessmentRequest{
		BusinessName:    "Test Company",
		BusinessAddress: "123 Test St",
		Industry:        "technology",
		Country:         "US",
	}

	ctx := context.Background()
	result, err := manager.GetComprehensiveData(ctx, business)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, business.BusinessName, result.BusinessName)
	assert.Equal(t, business.Country, result.Country)
	assert.Equal(t, business.Industry, result.Industry)
	assert.NotZero(t, result.ProcessingTime)
	assert.NotZero(t, result.LastChecked)
	assert.NotEmpty(t, result.DataQuality)
	assert.GreaterOrEqual(t, result.OverallRiskScore, 0.0)
	assert.LessOrEqual(t, result.OverallRiskScore, 1.0)
	assert.NotNil(t, result.APIStatus)
	assert.NotEmpty(t, result.RiskFactors)
}

func TestExternalAPIManager_GetThomsonReutersData(t *testing.T) {
	logger := zap.NewNop()

	config := &ExternalAPIManagerConfig{
		ThomsonReuters: &ThomsonReutersConfig{
			APIKey:    "test_key",
			BaseURL:   "https://api.test.com",
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   true,
		},
		Timeout: 30 * time.Second,
	}

	manager := NewExternalAPIManager(config, logger)

	business := &models.RiskAssessmentRequest{
		BusinessName:    "Test Company",
		BusinessAddress: "123 Test St",
		Industry:        "technology",
		Country:         "US",
	}

	ctx := context.Background()
	result, err := manager.GetThomsonReutersData(ctx, business)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.CompanyProfile)
	assert.Equal(t, business.BusinessName, result.CompanyProfile.CompanyName)
	assert.NotZero(t, result.ProcessingTime)
	assert.NotEmpty(t, result.DataQuality)
}

func TestExternalAPIManager_GetOFACData(t *testing.T) {
	logger := zap.NewNop()

	config := &ExternalAPIManagerConfig{
		OFAC: &OFACConfig{
			APIKey:    "test_key",
			BaseURL:   "https://api.test.com",
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   true,
		},
		Timeout: 30 * time.Second,
	}

	manager := NewExternalAPIManager(config, logger)

	business := &models.RiskAssessmentRequest{
		BusinessName:    "Test Company",
		BusinessAddress: "123 Test St",
		Industry:        "technology",
		Country:         "US",
	}

	ctx := context.Background()
	result, err := manager.GetOFACData(ctx, business)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.SanctionsSearch)
	assert.NotNil(t, result.ComplianceStatus)
	assert.NotNil(t, result.EntityVerification)
	assert.NotZero(t, result.ProcessingTime)
	assert.NotEmpty(t, result.DataQuality)
}

func TestExternalAPIManager_GetWorldCheckData(t *testing.T) {
	logger := zap.NewNop()

	config := &ExternalAPIManagerConfig{
		WorldCheck: &WorldCheckConfig{
			APIKey:    "test_key",
			BaseURL:   "https://api.test.com",
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   true,
		},
		Timeout: 30 * time.Second,
	}

	manager := NewExternalAPIManager(config, logger)

	business := &models.RiskAssessmentRequest{
		BusinessName:    "Test Company",
		BusinessAddress: "123 Test St",
		Industry:        "technology",
		Country:         "US",
	}

	ctx := context.Background()
	result, err := manager.GetWorldCheckData(ctx, business)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Profile)
	assert.Equal(t, business.BusinessName, result.Profile.EntityName)
	assert.NotZero(t, result.ProcessingTime)
	assert.NotEmpty(t, result.DataQuality)
}

func TestExternalAPIManager_GetAPIStatus(t *testing.T) {
	logger := zap.NewNop()

	config := &ExternalAPIManagerConfig{
		ThomsonReuters: &ThomsonReutersConfig{
			APIKey:    "test_key",
			BaseURL:   "https://api.test.com",
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   true,
		},
		OFAC: &OFACConfig{
			APIKey:    "test_key",
			BaseURL:   "https://api.test.com",
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   true,
		},
		WorldCheck: &WorldCheckConfig{
			APIKey:    "test_key",
			BaseURL:   "https://api.test.com",
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   true,
		},
		Timeout: 30 * time.Second,
	}

	manager := NewExternalAPIManager(config, logger)

	status := manager.GetAPIStatus()

	assert.Equal(t, "enabled", status["thomson_reuters"])
	assert.Equal(t, "enabled", status["ofac"])
	assert.Equal(t, "enabled", status["worldcheck"])
}

func TestExternalAPIManager_GetSupportedAPIs(t *testing.T) {
	logger := zap.NewNop()

	config := &ExternalAPIManagerConfig{
		ThomsonReuters: &ThomsonReutersConfig{
			APIKey:    "test_key",
			BaseURL:   "https://api.test.com",
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   true,
		},
		OFAC: &OFACConfig{
			APIKey:    "test_key",
			BaseURL:   "https://api.test.com",
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   true,
		},
		WorldCheck: &WorldCheckConfig{
			APIKey:    "test_key",
			BaseURL:   "https://api.test.com",
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   true,
		},
		Timeout: 30 * time.Second,
	}

	manager := NewExternalAPIManager(config, logger)

	apis := manager.GetSupportedAPIs()

	assert.Contains(t, apis, "thomson_reuters")
	assert.Contains(t, apis, "ofac")
	assert.Contains(t, apis, "worldcheck")
	assert.Len(t, apis, 3)
}

func TestExternalAPIManager_HealthCheck(t *testing.T) {
	logger := zap.NewNop()

	config := &ExternalAPIManagerConfig{
		ThomsonReuters: &ThomsonReutersConfig{
			APIKey:    "test_key",
			BaseURL:   "https://api.test.com",
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   true,
		},
		OFAC: &OFACConfig{
			APIKey:    "test_key",
			BaseURL:   "https://api.test.com",
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   true,
		},
		WorldCheck: &WorldCheckConfig{
			APIKey:    "test_key",
			BaseURL:   "https://api.test.com",
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   true,
		},
		Timeout: 30 * time.Second,
	}

	manager := NewExternalAPIManager(config, logger)

	ctx := context.Background()
	health := manager.HealthCheck(ctx)

	assert.True(t, health["thomson_reuters"])
	assert.True(t, health["ofac"])
	assert.True(t, health["worldcheck"])
}

func TestThomsonReutersMock_GetCompanyProfile(t *testing.T) {
	logger := zap.NewNop()
	config := &thomson_reuters.ThomsonReutersConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := thomson_reuters.NewThomsonReutersMock(config, logger)

	ctx := context.Background()
	profile, err := client.GetCompanyProfile(ctx, "Test Company", "US")

	require.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, "Test Company", profile.CompanyName)
	assert.Equal(t, "US", profile.Country)
	assert.NotEmpty(t, profile.CompanyID)
	assert.NotEmpty(t, profile.Industry)
	assert.NotZero(t, profile.FoundedYear)
	assert.NotZero(t, profile.EmployeeCount)
	assert.NotZero(t, profile.Revenue)
	assert.NotEmpty(t, profile.BusinessDescription)
	assert.NotEmpty(t, profile.Website)
	assert.NotEmpty(t, profile.Address)
	assert.NotEmpty(t, profile.Phone)
	assert.NotEmpty(t, profile.Email)
	assert.NotEmpty(t, profile.DataQuality)
	assert.NotZero(t, profile.LastUpdated)
}

func TestThomsonReutersMock_GetFinancialData(t *testing.T) {
	logger := zap.NewNop()
	config := &thomson_reuters.ThomsonReutersConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := thomson_reuters.NewThomsonReutersMock(config, logger)

	ctx := context.Background()
	financialData, err := client.GetFinancialData(ctx, "TEST_COMPANY_123")

	require.NoError(t, err)
	assert.NotNil(t, financialData)
	assert.Equal(t, "TEST_COMPANY_123", financialData.CompanyID)
	assert.NotZero(t, financialData.FiscalYear)
	assert.NotZero(t, financialData.Revenue)
	assert.NotZero(t, financialData.TotalAssets)
	assert.NotZero(t, financialData.TotalLiabilities)
	assert.NotZero(t, financialData.ShareholdersEquity)
	assert.NotZero(t, financialData.LastUpdated)
}

func TestThomsonReutersMock_GetFinancialRatios(t *testing.T) {
	logger := zap.NewNop()
	config := &thomson_reuters.ThomsonReutersConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := thomson_reuters.NewThomsonReutersMock(config, logger)

	ctx := context.Background()
	ratios, err := client.GetFinancialRatios(ctx, "TEST_COMPANY_123")

	require.NoError(t, err)
	assert.NotNil(t, ratios)
	assert.Equal(t, "TEST_COMPANY_123", ratios.CompanyID)
	assert.Greater(t, ratios.CurrentRatio, 0.0)
	assert.Greater(t, ratios.QuickRatio, 0.0)
	assert.Greater(t, ratios.DebtToEquityRatio, 0.0)
	assert.Greater(t, ratios.ReturnOnEquity, 0.0)
	assert.Greater(t, ratios.ReturnOnAssets, 0.0)
	assert.Greater(t, ratios.GrossProfitMargin, 0.0)
	assert.Greater(t, ratios.NetProfitMargin, 0.0)
	assert.Greater(t, ratios.AssetTurnover, 0.0)
	assert.Greater(t, ratios.InventoryTurnover, 0.0)
	assert.NotZero(t, ratios.LastUpdated)
}

func TestThomsonReutersMock_GetRiskMetrics(t *testing.T) {
	logger := zap.NewNop()
	config := &thomson_reuters.ThomsonReutersConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := thomson_reuters.NewThomsonReutersMock(config, logger)

	ctx := context.Background()
	riskMetrics, err := client.GetRiskMetrics(ctx, "TEST_COMPANY_123")

	require.NoError(t, err)
	assert.NotNil(t, riskMetrics)
	assert.Equal(t, "TEST_COMPANY_123", riskMetrics.CompanyID)
	assert.GreaterOrEqual(t, riskMetrics.OverallRiskScore, 0.0)
	assert.LessOrEqual(t, riskMetrics.OverallRiskScore, 1.0)
	assert.GreaterOrEqual(t, riskMetrics.FinancialRisk, 0.0)
	assert.LessOrEqual(t, riskMetrics.FinancialRisk, 1.0)
	assert.GreaterOrEqual(t, riskMetrics.OperationalRisk, 0.0)
	assert.LessOrEqual(t, riskMetrics.OperationalRisk, 1.0)
	assert.GreaterOrEqual(t, riskMetrics.MarketRisk, 0.0)
	assert.LessOrEqual(t, riskMetrics.MarketRisk, 1.0)
	assert.GreaterOrEqual(t, riskMetrics.CreditRisk, 0.0)
	assert.LessOrEqual(t, riskMetrics.CreditRisk, 1.0)
	assert.GreaterOrEqual(t, riskMetrics.LiquidityRisk, 0.0)
	assert.LessOrEqual(t, riskMetrics.LiquidityRisk, 1.0)
	assert.GreaterOrEqual(t, riskMetrics.RegulatoryRisk, 0.0)
	assert.LessOrEqual(t, riskMetrics.RegulatoryRisk, 1.0)
	assert.GreaterOrEqual(t, riskMetrics.ESGRisk, 0.0)
	assert.LessOrEqual(t, riskMetrics.ESGRisk, 1.0)
	assert.NotEmpty(t, riskMetrics.RiskTrend)
	assert.NotZero(t, riskMetrics.LastUpdated)
}

func TestThomsonReutersMock_GetESGScore(t *testing.T) {
	logger := zap.NewNop()
	config := &thomson_reuters.ThomsonReutersConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := thomson_reuters.NewThomsonReutersMock(config, logger)

	ctx := context.Background()
	esgScore, err := client.GetESGScore(ctx, "TEST_COMPANY_123")

	require.NoError(t, err)
	assert.NotNil(t, esgScore)
	assert.Equal(t, "TEST_COMPANY_123", esgScore.CompanyID)
	assert.GreaterOrEqual(t, esgScore.OverallESGScore, 0.0)
	assert.LessOrEqual(t, esgScore.OverallESGScore, 100.0)
	assert.GreaterOrEqual(t, esgScore.Environmental, 0.0)
	assert.LessOrEqual(t, esgScore.Environmental, 100.0)
	assert.GreaterOrEqual(t, esgScore.Social, 0.0)
	assert.LessOrEqual(t, esgScore.Social, 100.0)
	assert.GreaterOrEqual(t, esgScore.Governance, 0.0)
	assert.LessOrEqual(t, esgScore.Governance, 100.0)
	assert.NotEmpty(t, esgScore.ESGLevel)
	assert.NotZero(t, esgScore.LastUpdated)
}

func TestThomsonReutersMock_GetExecutiveInfo(t *testing.T) {
	logger := zap.NewNop()
	config := &thomson_reuters.ThomsonReutersConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := thomson_reuters.NewThomsonReutersMock(config, logger)

	ctx := context.Background()
	executiveInfo, err := client.GetExecutiveInfo(ctx, "TEST_COMPANY_123")

	require.NoError(t, err)
	assert.NotNil(t, executiveInfo)
	assert.Equal(t, "TEST_COMPANY_123", executiveInfo.CompanyID)
	assert.NotEmpty(t, executiveInfo.Executives)
	assert.Len(t, executiveInfo.Executives, 3)

	for _, executive := range executiveInfo.Executives {
		assert.NotEmpty(t, executive.Name)
		assert.NotEmpty(t, executive.Title)
		assert.Greater(t, executive.Tenure, 0)
		assert.NotEmpty(t, executive.Experience)
		assert.NotEmpty(t, executive.Education)
		assert.Greater(t, executive.Compensation, 0.0)
	}

	assert.NotZero(t, executiveInfo.LastUpdated)
}

func TestThomsonReutersMock_GetOwnershipStructure(t *testing.T) {
	logger := zap.NewNop()
	config := &thomson_reuters.ThomsonReutersConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := thomson_reuters.NewThomsonReutersMock(config, logger)

	ctx := context.Background()
	ownershipStructure, err := client.GetOwnershipStructure(ctx, "TEST_COMPANY_123")

	require.NoError(t, err)
	assert.NotNil(t, ownershipStructure)
	assert.Equal(t, "TEST_COMPANY_123", ownershipStructure.CompanyID)
	assert.NotEmpty(t, ownershipStructure.OwnershipData)
	assert.Len(t, ownershipStructure.OwnershipData, 3)

	for _, ownership := range ownershipStructure.OwnershipData {
		assert.NotEmpty(t, ownership.OwnerName)
		assert.NotEmpty(t, ownership.OwnerType)
		assert.Greater(t, ownership.OwnershipPercentage, 0.0)
		assert.LessOrEqual(t, ownership.OwnershipPercentage, 100.0)
		assert.Greater(t, ownership.Shares, int64(0))
	}

	assert.NotZero(t, ownershipStructure.LastUpdated)
}

func TestThomsonReutersMock_GetComprehensiveData(t *testing.T) {
	logger := zap.NewNop()
	config := &thomson_reuters.ThomsonReutersConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := thomson_reuters.NewThomsonReutersMock(config, logger)

	ctx := context.Background()
	result, err := client.GetComprehensiveData(ctx, "Test Company", "US")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.CompanyProfile)
	assert.NotNil(t, result.FinancialData)
	assert.NotNil(t, result.FinancialRatios)
	assert.NotNil(t, result.RiskMetrics)
	assert.NotNil(t, result.ESGScore)
	assert.NotNil(t, result.ExecutiveInfo)
	assert.NotNil(t, result.OwnershipStructure)
	assert.NotEmpty(t, result.DataQuality)
	assert.NotZero(t, result.LastChecked)
	assert.NotZero(t, result.ProcessingTime)
}

func TestThomsonReutersMock_GenerateRiskFactors(t *testing.T) {
	logger := zap.NewNop()
	config := &thomson_reuters.ThomsonReutersConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := thomson_reuters.NewThomsonReutersMock(config, logger)

	// Create a mock result
	result := &thomson_reuters.ThomsonReutersResult{
		FinancialData: &thomson_reuters.FinancialData{
			Revenue:   1000000,
			NetIncome: 100000,
		},
		RiskMetrics: &thomson_reuters.RiskMetrics{
			OverallRiskScore: 0.5,
		},
		ESGScore: &thomson_reuters.ESGScore{
			OverallESGScore: 75.0,
		},
	}

	riskFactors := client.GenerateRiskFactors(result)

	assert.NotEmpty(t, riskFactors)
	assert.Len(t, riskFactors, 4) // revenue_growth_risk, profitability_risk, thomson_reuters_risk_score, esg_risk

	for _, factor := range riskFactors {
		assert.NotEmpty(t, factor.Category)
		assert.NotEmpty(t, factor.Subcategory)
		assert.NotEmpty(t, factor.Name)
		assert.GreaterOrEqual(t, factor.Score, 0.0)
		assert.LessOrEqual(t, factor.Score, 1.0)
		assert.Greater(t, factor.Weight, 0.0)
		assert.NotEmpty(t, factor.Description)
		assert.Equal(t, "thomson_reuters", factor.Source)
		assert.Greater(t, factor.Confidence, 0.0)
		assert.NotEmpty(t, factor.Impact)
		assert.NotEmpty(t, factor.Mitigation)
		assert.NotNil(t, factor.LastUpdated)
	}
}

func TestOFACMock_SearchSanctions(t *testing.T) {
	logger := zap.NewNop()
	config := &ofac.OFACConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := ofac.NewOFACMock(config, logger)

	ctx := context.Background()
	result, err := client.SearchSanctions(ctx, "Test Entity", "US")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Entity", result.Query)
	assert.GreaterOrEqual(t, result.TotalMatches, 0)
	assert.NotZero(t, result.SearchTime)
	assert.NotZero(t, result.LastChecked)
	assert.NotEmpty(t, result.DataQuality)
}

func TestOFACMock_VerifyEntity(t *testing.T) {
	logger := zap.NewNop()
	config := &ofac.OFACConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := ofac.NewOFACMock(config, logger)

	ctx := context.Background()
	result, err := client.VerifyEntity(ctx, "Test Entity", "US")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Entity", result.EntityName)
	assert.GreaterOrEqual(t, result.VerificationScore, 0.0)
	assert.LessOrEqual(t, result.VerificationScore, 1.0)
	assert.NotEmpty(t, result.MatchType)
	assert.GreaterOrEqual(t, result.Confidence, 0.0)
	assert.LessOrEqual(t, result.Confidence, 1.0)
	assert.NotZero(t, result.VerificationDate)
	assert.NotEmpty(t, result.Notes)
}

func TestOFACMock_GetComplianceStatus(t *testing.T) {
	logger := zap.NewNop()
	config := &ofac.OFACConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := ofac.NewOFACMock(config, logger)

	ctx := context.Background()
	result, err := client.GetComplianceStatus(ctx, "Test Entity", "US")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Entity", result.EntityName)
	assert.GreaterOrEqual(t, result.SanctionsMatches, 0)
	assert.GreaterOrEqual(t, result.ComplianceScore, 0.0)
	assert.LessOrEqual(t, result.ComplianceScore, 1.0)
	assert.NotZero(t, result.LastScreened)
	assert.NotZero(t, result.NextScreening)
	assert.NotEmpty(t, result.ComplianceNotes)
}

func TestOFACMock_GetComprehensiveData(t *testing.T) {
	logger := zap.NewNop()
	config := &ofac.OFACConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := ofac.NewOFACMock(config, logger)

	ctx := context.Background()
	result, err := client.GetComprehensiveData(ctx, "Test Entity", "US")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.SanctionsSearch)
	assert.NotNil(t, result.ComplianceStatus)
	assert.NotNil(t, result.EntityVerification)
	assert.NotEmpty(t, result.DataQuality)
	assert.NotZero(t, result.LastChecked)
	assert.NotZero(t, result.ProcessingTime)
}

func TestOFACMock_GenerateRiskFactors(t *testing.T) {
	logger := zap.NewNop()
	config := &ofac.OFACConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := ofac.NewOFACMock(config, logger)

	// Create a mock result
	result := &ofac.OFACResult{
		SanctionsSearch: &ofac.SanctionsSearchResult{
			TotalMatches: 0,
		},
		ComplianceStatus: &ofac.ComplianceStatus{
			IsCompliant: true,
			RiskLevel:   "low",
		},
		EntityVerification: &ofac.EntityVerification{
			IsVerified: true,
			Confidence: 0.9,
		},
	}

	riskFactors := client.GenerateRiskFactors(result)

	assert.NotEmpty(t, riskFactors)
	assert.Len(t, riskFactors, 3) // sanctions_risk, compliance_risk, entity_verification_risk

	for _, factor := range riskFactors {
		assert.NotEmpty(t, factor.Category)
		assert.NotEmpty(t, factor.Subcategory)
		assert.NotEmpty(t, factor.Name)
		assert.GreaterOrEqual(t, factor.Score, 0.0)
		assert.LessOrEqual(t, factor.Score, 1.0)
		assert.Greater(t, factor.Weight, 0.0)
		assert.NotEmpty(t, factor.Description)
		assert.Equal(t, "ofac", factor.Source)
		assert.Greater(t, factor.Confidence, 0.0)
		assert.NotEmpty(t, factor.Impact)
		assert.NotEmpty(t, factor.Mitigation)
		assert.NotNil(t, factor.LastUpdated)
	}
}

func TestWorldCheckMock_SearchProfile(t *testing.T) {
	logger := zap.NewNop()
	config := &worldcheck.WorldCheckConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := worldcheck.NewWorldCheckMock(config, logger)

	ctx := context.Background()
	profile, err := client.SearchProfile(ctx, "Test Entity", "US")

	require.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, "Test Entity", profile.EntityName)
	assert.Equal(t, "US", profile.Country)
	assert.NotEmpty(t, profile.ProfileID)
	assert.NotEmpty(t, profile.EntityType)
	assert.NotEmpty(t, profile.RiskLevel)
	assert.NotEmpty(t, profile.Category)
	assert.NotEmpty(t, profile.SubCategory)
	assert.NotEmpty(t, profile.Source)
	assert.NotZero(t, profile.LastUpdated)
	assert.NotEmpty(t, profile.ProfileStatus)
	assert.GreaterOrEqual(t, profile.MatchScore, 0.0)
	assert.LessOrEqual(t, profile.MatchScore, 1.0)
	assert.GreaterOrEqual(t, profile.Confidence, 0.0)
	assert.LessOrEqual(t, profile.Confidence, 1.0)
}

func TestWorldCheckMock_GetAdverseMedia(t *testing.T) {
	logger := zap.NewNop()
	config := &worldcheck.WorldCheckConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := worldcheck.NewWorldCheckMock(config, logger)

	ctx := context.Background()
	adverseMedia, err := client.GetAdverseMedia(ctx, "Test Entity")

	require.NoError(t, err)
	assert.NotNil(t, adverseMedia)
	assert.GreaterOrEqual(t, len(adverseMedia), 0)
	assert.LessOrEqual(t, len(adverseMedia), 3)

	for _, media := range adverseMedia {
		assert.NotEmpty(t, media.MediaID)
		assert.NotEmpty(t, media.Title)
		assert.NotEmpty(t, media.Source)
		assert.NotEmpty(t, media.URL)
		assert.NotZero(t, media.PublishedDate)
		assert.NotEmpty(t, media.Content)
		assert.NotEmpty(t, media.Sentiment)
		assert.GreaterOrEqual(t, media.Relevance, 0.0)
		assert.LessOrEqual(t, media.Relevance, 1.0)
		assert.NotEmpty(t, media.RiskLevel)
	}
}

func TestWorldCheckMock_GetPEPStatus(t *testing.T) {
	logger := zap.NewNop()
	config := &worldcheck.WorldCheckConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := worldcheck.NewWorldCheckMock(config, logger)

	ctx := context.Background()
	pepStatus, err := client.GetPEPStatus(ctx, "Test Entity")

	require.NoError(t, err)
	assert.NotNil(t, pepStatus)
	assert.NotEmpty(t, pepStatus.PEPLevel)
}

func TestWorldCheckMock_GetSanctionsInfo(t *testing.T) {
	logger := zap.NewNop()
	config := &worldcheck.WorldCheckConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := worldcheck.NewWorldCheckMock(config, logger)

	ctx := context.Background()
	sanctionsInfo, err := client.GetSanctionsInfo(ctx, "Test Entity")

	require.NoError(t, err)
	assert.NotNil(t, sanctionsInfo)
}

func TestWorldCheckMock_GetRiskAssessment(t *testing.T) {
	logger := zap.NewNop()
	config := &worldcheck.WorldCheckConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := worldcheck.NewWorldCheckMock(config, logger)

	ctx := context.Background()
	riskAssessment, err := client.GetRiskAssessment(ctx, "Test Entity")

	require.NoError(t, err)
	assert.NotNil(t, riskAssessment)
	assert.GreaterOrEqual(t, riskAssessment.OverallRiskScore, 0.0)
	assert.LessOrEqual(t, riskAssessment.OverallRiskScore, 1.0)
	assert.GreaterOrEqual(t, riskAssessment.FinancialRisk, 0.0)
	assert.LessOrEqual(t, riskAssessment.FinancialRisk, 1.0)
	assert.GreaterOrEqual(t, riskAssessment.ReputationalRisk, 0.0)
	assert.LessOrEqual(t, riskAssessment.ReputationalRisk, 1.0)
	assert.GreaterOrEqual(t, riskAssessment.RegulatoryRisk, 0.0)
	assert.LessOrEqual(t, riskAssessment.RegulatoryRisk, 1.0)
	assert.GreaterOrEqual(t, riskAssessment.OperationalRisk, 0.0)
	assert.LessOrEqual(t, riskAssessment.OperationalRisk, 1.0)
	assert.NotEmpty(t, riskAssessment.RiskFactors)
	assert.NotEmpty(t, riskAssessment.Recommendations)
}

func TestWorldCheckMock_GetComprehensiveData(t *testing.T) {
	logger := zap.NewNop()
	config := &worldcheck.WorldCheckConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := worldcheck.NewWorldCheckMock(config, logger)

	ctx := context.Background()
	result, err := client.GetComprehensiveData(ctx, "Test Entity", "US")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Profile)
	assert.NotNil(t, result.AdverseMedia)
	assert.NotNil(t, result.PEPStatus)
	assert.NotNil(t, result.SanctionsInfo)
	assert.NotNil(t, result.RiskAssessment)
	assert.NotEmpty(t, result.DataQuality)
	assert.NotZero(t, result.LastChecked)
	assert.NotZero(t, result.ProcessingTime)
}

func TestWorldCheckMock_GenerateRiskFactors(t *testing.T) {
	logger := zap.NewNop()
	config := &worldcheck.WorldCheckConfig{
		APIKey:    "test_key",
		BaseURL:   "https://api.test.com",
		Timeout:   30 * time.Second,
		RateLimit: 100,
		Enabled:   true,
	}

	client := worldcheck.NewWorldCheckMock(config, logger)

	// Create a mock result
	result := &worldcheck.WorldCheckResult{
		Profile: &worldcheck.WorldCheckProfile{
			RiskLevel: "low",
		},
		AdverseMedia: []worldcheck.AdverseMedia{
			{Title: "Test Article", Sentiment: "negative"},
		},
		PEPStatus: &worldcheck.PEPStatus{
			IsPEP: false,
		},
		SanctionsInfo: &worldcheck.SanctionsInfo{
			IsSanctioned: false,
		},
		RiskAssessment: &worldcheck.RiskAssessment{
			OverallRiskScore: 0.3,
		},
	}

	riskFactors := client.GenerateRiskFactors(result)

	assert.NotEmpty(t, riskFactors)
	assert.Len(t, riskFactors, 3) // worldcheck_profile_risk, adverse_media_risk, worldcheck_risk_assessment

	for _, factor := range riskFactors {
		assert.NotEmpty(t, factor.Category)
		assert.NotEmpty(t, factor.Subcategory)
		assert.NotEmpty(t, factor.Name)
		assert.GreaterOrEqual(t, factor.Score, 0.0)
		assert.LessOrEqual(t, factor.Score, 1.0)
		assert.Greater(t, factor.Weight, 0.0)
		assert.NotEmpty(t, factor.Description)
		assert.Equal(t, "worldcheck", factor.Source)
		assert.Greater(t, factor.Confidence, 0.0)
		assert.NotEmpty(t, factor.Impact)
		assert.NotEmpty(t, factor.Mitigation)
		assert.NotNil(t, factor.LastUpdated)
	}
}

func TestExternalAPIManager_DisabledAPIs(t *testing.T) {
	logger := zap.NewNop()

	config := &ExternalAPIManagerConfig{
		ThomsonReuters: &ThomsonReutersConfig{
			Enabled: false,
		},
		OFAC: &OFACConfig{
			Enabled: false,
		},
		WorldCheck: &WorldCheckConfig{
			Enabled: false,
		},
		Timeout: 30 * time.Second,
	}

	manager := NewExternalAPIManager(config, logger)

	assert.Nil(t, manager.thomsonReuters)
	assert.Nil(t, manager.ofac)
	assert.Nil(t, manager.worldCheck)

	status := manager.GetAPIStatus()
	assert.Equal(t, "disabled", status["thomson_reuters"])
	assert.Equal(t, "disabled", status["ofac"])
	assert.Equal(t, "disabled", status["worldcheck"])

	apis := manager.GetSupportedAPIs()
	assert.Empty(t, apis)

	business := &models.RiskAssessmentRequest{
		BusinessName:    "Test Company",
		BusinessAddress: "123 Test St",
		Industry:        "technology",
		Country:         "US",
	}

	ctx := context.Background()
	result, err := manager.GetComprehensiveData(ctx, business)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, business.BusinessName, result.BusinessName)
	assert.Empty(t, result.RiskFactors) // No APIs enabled, so no risk factors
}

func TestExternalAPIManager_PartialAPIs(t *testing.T) {
	logger := zap.NewNop()

	config := &ExternalAPIManagerConfig{
		ThomsonReuters: &ThomsonReutersConfig{
			APIKey:    "test_key",
			BaseURL:   "https://api.test.com",
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   true,
		},
		OFAC: &OFACConfig{
			Enabled: false,
		},
		WorldCheck: &WorldCheckConfig{
			Enabled: false,
		},
		Timeout: 30 * time.Second,
	}

	manager := NewExternalAPIManager(config, logger)

	assert.NotNil(t, manager.thomsonReuters)
	assert.Nil(t, manager.ofac)
	assert.Nil(t, manager.worldCheck)

	status := manager.GetAPIStatus()
	assert.Equal(t, "enabled", status["thomson_reuters"])
	assert.Equal(t, "disabled", status["ofac"])
	assert.Equal(t, "disabled", status["worldcheck"])

	apis := manager.GetSupportedAPIs()
	assert.Len(t, apis, 1)
	assert.Contains(t, apis, "thomson_reuters")

	business := &models.RiskAssessmentRequest{
		BusinessName:    "Test Company",
		BusinessAddress: "123 Test St",
		Industry:        "technology",
		Country:         "US",
	}

	ctx := context.Background()
	result, err := manager.GetComprehensiveData(ctx, business)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, business.BusinessName, result.BusinessName)
	assert.NotNil(t, result.ThomsonReuters)
	assert.Nil(t, result.OFAC)
	assert.Nil(t, result.WorldCheck)
	assert.NotEmpty(t, result.RiskFactors) // Should have Thomson Reuters risk factors
}
