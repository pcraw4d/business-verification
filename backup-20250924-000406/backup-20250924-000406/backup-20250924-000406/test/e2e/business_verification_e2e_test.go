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

// TestBusinessVerificationWorkflow tests the business verification workflow
// This test covers: website scraping, ownership verification, data validation
func TestBusinessVerificationWorkflow(t *testing.T) {
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

	t.Run("Business Verification Journey", func(t *testing.T) {
		testBusinessVerificationJourney(t, server, mockPortfolioService, mockClassificationService, ctx)
	})
}

// testBusinessVerificationJourney tests the business verification workflow
func testBusinessVerificationJourney(
	t *testing.T,
	server *httptest.Server,
	mockPortfolioService *mocks.MockMerchantPortfolioService,
	mockClassificationService *mocks.MockClassificationService,
	ctx context.Context,
) {
	t.Log("=== Business Verification Journey ===")

	// Step 1: Create a merchant for verification
	t.Log("Step 1: Creating merchant for verification...")
	merchant := createOnboardingTestMerchant("verification_merchant_1", "Test Business", services.PortfolioTypeProspective, services.RiskLevelMedium)

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

	t.Logf("✅ Merchant created for verification: %s (ID: %s)", createdMerchant.Name, createdMerchant.ID)

	// Step 2: Test website scraping and ownership verification
	t.Log("Step 2: Testing website scraping and ownership verification...")

	// Configure mock classification service to return verification results
	// Note: We would normally use the classification package here, but for E2E testing
	// we'll use a simple mock structure to avoid complex imports

	// Test classification with website verification
	classificationReq := &handlers.ClassificationRequest{
		BusinessName: createdMerchant.Name,
		Country:      "USA",
		WebsiteURL:   createdMerchant.ContactInfo.Website,
	}

	classificationResp, classificationBody, err := makeOnboardingRequest(server, "POST", "/api/v1/classification", classificationReq)
	if err != nil {
		t.Fatalf("Failed to classify merchant: %v", err)
	}

	if classificationResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d: %s", classificationResp.StatusCode, string(classificationBody))
	}

	var classificationResult handlers.ClassificationResponse
	if err := json.Unmarshal(classificationBody, &classificationResult); err != nil {
		t.Fatalf("Failed to unmarshal classification result: %v", err)
	}

	// Verify website verification results
	if classificationResult.WebsiteVerification == nil {
		t.Error("Expected website verification results")
	} else {
		if classificationResult.WebsiteVerification.Status != "PASSED" {
			t.Errorf("Expected verification status 'PASSED', got %s", classificationResult.WebsiteVerification.Status)
		}
		if classificationResult.WebsiteVerification.DataMatchScore < 0.9 {
			t.Errorf("Expected data match score >= 0.9, got %f", classificationResult.WebsiteVerification.DataMatchScore)
		}
	}

	t.Logf("✅ Website scraping and ownership verification completed")

	// Step 3: Test data validation
	t.Log("Step 3: Testing data validation...")

	// Verify classification results structure
	if classificationResult.PrimaryIndustry == "" {
		t.Error("Expected primary industry to be set")
	}
	if classificationResult.OverallConfidence < 0.8 {
		t.Errorf("Expected overall confidence >= 0.8, got %f", classificationResult.OverallConfidence)
	}
	if classificationResult.MethodBreakdown == nil {
		t.Error("Expected method breakdown to be present")
	}

	t.Logf("✅ Data validation completed")

	t.Log("=== Business Verification Journey Completed ===")
}
