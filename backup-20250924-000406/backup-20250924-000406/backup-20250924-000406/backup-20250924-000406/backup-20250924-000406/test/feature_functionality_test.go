package test

import (
	"net/http/httptest"
	"testing"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/classification"
	"github.com/pcraw4d/business-verification/internal/modules/risk_assessment"
)

// FeatureFunctionalityTestSuite provides comprehensive testing for all feature functionality
type FeatureFunctionalityTestSuite struct {
	classificationService *classification.Service
	riskAssessmentService *risk_assessment.RiskAssessmentService
	merchantHandler       *handlers.MerchantPortfolioHandler
	testServer            *httptest.Server
}

// NewFeatureFunctionalityTestSuite creates a new test suite
func NewFeatureFunctionalityTestSuite() *FeatureFunctionalityTestSuite {
	return &FeatureFunctionalityTestSuite{}
}

// SetupTestSuite initializes the test suite with mock services
func (suite *FeatureFunctionalityTestSuite) SetupTestSuite(t *testing.T) {
	// Initialize mock services
	suite.classificationService = createMockClassificationService()
	suite.riskAssessmentService = createMockRiskAssessmentService()
	suite.merchantHandler = createMockMerchantHandler()

	// Setup test server
	suite.setupTestServer()
}

// TestBusinessClassificationFeatures tests all business classification functionality
func (suite *FeatureFunctionalityTestSuite) TestBusinessClassificationFeatures(t *testing.T) {
	t.Run("MultiMethodClassification", func(t *testing.T) {
		suite.testMultiMethodClassification(t)
	})

	t.Run("KeywordBasedClassification", func(t *testing.T) {
		suite.testKeywordBasedClassification(t)
	})

	t.Run("MLBasedClassification", func(t *testing.T) {
		suite.testMLBasedClassification(t)
	})

	t.Run("EnsembleClassification", func(t *testing.T) {
		suite.testEnsembleClassification(t)
	})

	t.Run("ConfidenceScoring", func(t *testing.T) {
		suite.testConfidenceScoring(t)
	})
}

// TestRiskAssessmentFeatures tests all risk assessment functionality
func (suite *FeatureFunctionalityTestSuite) TestRiskAssessmentFeatures(t *testing.T) {
	t.Run("ComprehensiveRiskAssessment", func(t *testing.T) {
		suite.testComprehensiveRiskAssessment(t)
	})

	t.Run("SecurityAnalysis", func(t *testing.T) {
		suite.testSecurityAnalysis(t)
	})

	t.Run("DomainAnalysis", func(t *testing.T) {
		suite.testDomainAnalysis(t)
	})

	t.Run("ReputationAnalysis", func(t *testing.T) {
		suite.testReputationAnalysis(t)
	})

	t.Run("ComplianceAnalysis", func(t *testing.T) {
		suite.testComplianceAnalysis(t)
	})

	t.Run("FinancialAnalysis", func(t *testing.T) {
		suite.testFinancialAnalysis(t)
	})

	t.Run("RiskScoring", func(t *testing.T) {
		suite.testRiskScoring(t)
	})
}

// TestComplianceCheckingFeatures tests all compliance checking functionality
func (suite *FeatureFunctionalityTestSuite) TestComplianceCheckingFeatures(t *testing.T) {
	t.Run("AMLCompliance", func(t *testing.T) {
		suite.testAMLCompliance(t)
	})

	t.Run("KYCCompliance", func(t *testing.T) {
		suite.testKYCCompliance(t)
	})

	t.Run("KYBCompliance", func(t *testing.T) {
		suite.testKYBCompliance(t)
	})

	t.Run("GDPRCompliance", func(t *testing.T) {
		suite.testGDPRCompliance(t)
	})

	t.Run("PCICompliance", func(t *testing.T) {
		suite.testPCICompliance(t)
	})

	t.Run("SOC2Compliance", func(t *testing.T) {
		suite.testSOC2Compliance(t)
	})
}

// TestMerchantManagementFeatures tests all merchant management functionality
func (suite *FeatureFunctionalityTestSuite) TestMerchantManagementFeatures(t *testing.T) {
	t.Run("CreateMerchant", func(t *testing.T) {
		suite.testCreateMerchant(t)
	})

	t.Run("GetMerchant", func(t *testing.T) {
		suite.testGetMerchant(t)
	})

	t.Run("UpdateMerchant", func(t *testing.T) {
		suite.testUpdateMerchant(t)
	})

	t.Run("DeleteMerchant", func(t *testing.T) {
		suite.testDeleteMerchant(t)
	})

	t.Run("SearchMerchants", func(t *testing.T) {
		suite.testSearchMerchants(t)
	})

	t.Run("BulkOperations", func(t *testing.T) {
		suite.testBulkOperations(t)
	})

	t.Run("PortfolioManagement", func(t *testing.T) {
		suite.testPortfolioManagement(t)
	})
}

// RunAllTests runs all feature functionality tests
func (suite *FeatureFunctionalityTestSuite) RunAllTests(t *testing.T) {
	suite.SetupTestSuite(t)

	t.Run("BusinessClassification", suite.TestBusinessClassificationFeatures)
	t.Run("RiskAssessment", suite.TestRiskAssessmentFeatures)
	t.Run("ComplianceChecking", suite.TestComplianceCheckingFeatures)
	t.Run("MerchantManagement", suite.TestMerchantManagementFeatures)
}
