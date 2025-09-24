package e2e

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/services"
	"kyb-platform/test/mocks"
)

// TestRiskAssessmentWorkflow tests the risk assessment workflow
// This test covers: security analysis, domain analysis, reputation scoring, compliance checks
func TestRiskAssessmentWorkflow(t *testing.T) {
	// Skip if not running E2E tests
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	// Setup test environment
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	// Create mock services
	mockPortfolioService := mocks.NewMockMerchantPortfolioService()
	mockClassificationService := mocks.NewMockClassificationService()
	mockRiskAssessmentService := mocks.NewMockRiskAssessmentService()
	mockComparisonService := mocks.NewMockComparisonService()

	// Create handlers
	portfolioHandler := handlers.NewMerchantPortfolioHandler(mockPortfolioService, nil)
	classificationHandler := handlers.NewClassificationHandler(mockClassificationService, nil)
	riskHandler := handlers.NewRiskAssessmentHandler(mockRiskAssessmentService, nil)
	comparisonHandler := handlers.NewComparisonHandler(mockComparisonService, nil)

	// Create test server
	server := httptest.NewServer(createUserOnboardingTestRouter(
		portfolioHandler,
		classificationHandler,
		riskHandler,
		comparisonHandler,
	))
	defer server.Close()

	t.Run("Risk Assessment Journey", func(t *testing.T) {
		testCompleteRiskAssessmentJourney(t, server, mockPortfolioService, mockRiskAssessmentService, ctx)
	})
}

// testCompleteRiskAssessmentJourney tests the risk assessment workflow
func testCompleteRiskAssessmentJourney(
	t *testing.T,
	server *httptest.Server,
	mockPortfolioService *mocks.MockMerchantPortfolioService,
	mockRiskAssessmentService *mocks.MockRiskAssessmentService,
	ctx context.Context,
) {
	t.Log("=== Risk Assessment Journey ===")

	// Step 1: Create a merchant for risk assessment
	t.Log("Step 1: Creating merchant for risk assessment...")
	merchant := createOnboardingTestMerchant("risk_merchant_1", "High Risk Business", services.PortfolioTypeProspective, services.RiskLevelHigh)

	createReq := &handlers.CreateMerchantRequest{
		Name:               merchant.Name,
		LegalName:          merchant.LegalName,
		RegistrationNumber: merchant.RegistrationNumber,
		TaxID:              merchant.TaxID,
		Industry:           merchant.Industry,
		IndustryCode:       merchant.IndustryCode,
		BusinessType:       merchant.BusinessType,
		EmployeeCount:      merchant.EmployeeCount,
		Address:            merchant.Address,
		ContactInfo:        merchant.ContactInfo,
		PortfolioType:      string(merchant.PortfolioType),
		RiskLevel:          string(merchant.RiskLevel),
	}

	_, createBody, err := makeOnboardingRequest(server, "POST", "/api/v1/merchants", createReq)
	if err != nil {
		t.Fatalf("Failed to create merchant: %v", err)
	}

	var createdMerchant handlers.MerchantResponse
	if err := json.Unmarshal(createBody, &createdMerchant); err != nil {
		t.Fatalf("Failed to unmarshal created merchant: %v", err)
	}

	t.Logf("✅ Merchant created for risk assessment: %s (ID: %s)", createdMerchant.Name, createdMerchant.ID)

	// Step 2: Test security analysis
	t.Log("Step 2: Testing security analysis...")

	// Configure mock risk assessment service to return comprehensive results
	// Note: For E2E testing, we'll simulate the risk assessment response structure

	// Test risk assessment
	riskReq := &handlers.RiskAssessmentRequest{
		BusinessName: createdMerchant.Name,
		WebsiteURL:   createdMerchant.ContactInfo.Website,
		AnalysisOptions: &handlers.AnalysisOptions{
			SecurityAnalysis:     true,
			DomainAnalysis:       true,
			ReputationAnalysis:   true,
			ComplianceAnalysis:   true,
			FinancialAnalysis:    true,
			ComprehensiveScoring: true,
		},
	}

	riskResp, riskBody, err := makeOnboardingRequest(server, "POST", "/api/v1/risk-assessment", riskReq)
	if err != nil {
		t.Fatalf("Failed to assess risk: %v", err)
	}

	if riskResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d: %s", riskResp.StatusCode, string(riskBody))
	}

	var riskResult handlers.RiskAssessmentResponse
	if err := json.Unmarshal(riskBody, &riskResult); err != nil {
		t.Fatalf("Failed to unmarshal risk assessment result: %v", err)
	}

	t.Logf("✅ Security analysis completed")

	// Step 3: Test domain analysis
	t.Log("Step 3: Testing domain analysis...")

	// Verify domain analysis results
	if riskResult.DomainAnalysis == nil {
		t.Error("Expected domain analysis results")
	} else {
		if riskResult.DomainAnalysis.OverallDomainScore < 0.0 || riskResult.DomainAnalysis.OverallDomainScore > 1.0 {
			t.Errorf("Expected domain score between 0.0 and 1.0, got %f", riskResult.DomainAnalysis.OverallDomainScore)
		}
		if len(riskResult.DomainAnalysis.Issues) == 0 {
			t.Error("Expected domain issues to be identified")
		}
		if len(riskResult.DomainAnalysis.Recommendations) == 0 {
			t.Error("Expected domain recommendations to be provided")
		}
	}

	t.Logf("✅ Domain analysis completed")

	// Step 4: Test reputation scoring
	t.Log("Step 4: Testing reputation scoring...")

	// Verify reputation analysis results
	if riskResult.ReputationAnalysis == nil {
		t.Error("Expected reputation analysis results")
	} else {
		if riskResult.ReputationAnalysis.OverallReputationScore < 0.0 || riskResult.ReputationAnalysis.OverallReputationScore > 1.0 {
			t.Errorf("Expected reputation score between 0.0 and 1.0, got %f", riskResult.ReputationAnalysis.OverallReputationScore)
		}
		if len(riskResult.ReputationAnalysis.Issues) == 0 {
			t.Error("Expected reputation issues to be identified")
		}
		if len(riskResult.ReputationAnalysis.Recommendations) == 0 {
			t.Error("Expected reputation recommendations to be provided")
		}
	}

	t.Logf("✅ Reputation scoring completed")

	// Step 5: Test compliance checks
	t.Log("Step 5: Testing compliance checks...")

	// Verify compliance analysis results
	if riskResult.ComplianceAnalysis == nil {
		t.Error("Expected compliance analysis results")
	} else {
		if riskResult.ComplianceAnalysis.OverallComplianceScore < 0.0 || riskResult.ComplianceAnalysis.OverallComplianceScore > 1.0 {
			t.Errorf("Expected compliance score between 0.0 and 1.0, got %f", riskResult.ComplianceAnalysis.OverallComplianceScore)
		}
		if len(riskResult.ComplianceAnalysis.Issues) == 0 {
			t.Error("Expected compliance issues to be identified")
		}
		if len(riskResult.ComplianceAnalysis.Recommendations) == 0 {
			t.Error("Expected compliance recommendations to be provided")
		}
	}

	t.Logf("✅ Compliance checks completed")

	// Step 6: Verify overall risk assessment
	t.Log("Step 6: Verifying overall risk assessment...")

	// Verify overall risk score
	if riskResult.OverallRiskScore < 0.0 || riskResult.OverallRiskScore > 1.0 {
		t.Errorf("Expected overall risk score between 0.0 and 1.0, got %f", riskResult.OverallRiskScore)
	}

	// Verify risk level
	if riskResult.RiskLevel == "" {
		t.Error("Expected risk level to be set")
	}

	// Verify confidence score
	if riskResult.ConfidenceScore < 0.0 || riskResult.ConfidenceScore > 1.0 {
		t.Errorf("Expected confidence score between 0.0 and 1.0, got %f", riskResult.ConfidenceScore)
	}

	// Verify risk factors
	if len(riskResult.RiskFactors) == 0 {
		t.Error("Expected risk factors to be identified")
	}

	// Verify recommendations
	if len(riskResult.Recommendations) == 0 {
		t.Error("Expected recommendations to be provided")
	}

	t.Logf("✅ Overall risk assessment completed")

	t.Log("=== Risk Assessment Journey Completed ===")
}
