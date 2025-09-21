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

// TestClassificationProcess tests the classification process
// This test covers: multi-method classification, industry code assignment, confidence scoring
func TestClassificationProcess(t *testing.T) {
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

	t.Run("Classification Process Journey", func(t *testing.T) {
		testClassificationProcessJourney(t, server, mockPortfolioService, mockClassificationService, ctx)
	})
}

// testClassificationProcessJourney tests the classification process
func testClassificationProcessJourney(
	t *testing.T,
	server *httptest.Server,
	mockPortfolioService *mocks.MockMerchantPortfolioService,
	mockClassificationService *mocks.MockClassificationService,
	ctx context.Context,
) {
	t.Log("=== Classification Process Journey ===")

	// Step 1: Create a merchant for classification
	t.Log("Step 1: Creating merchant for classification...")
	merchant := createOnboardingTestMerchant("classification_merchant_1", "Tech Startup Inc", services.PortfolioTypeProspective, services.RiskLevelMedium)

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

	t.Logf("✅ Merchant created for classification: %s (ID: %s)", createdMerchant.Name, createdMerchant.ID)

	// Step 2: Test multi-method classification
	t.Log("Step 2: Testing multi-method classification...")

	// Configure mock classification service to return comprehensive results
	// Note: For E2E testing, we'll simulate the classification response structure

	// Test classification
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

	t.Logf("✅ Multi-method classification completed")

	// Step 3: Test industry code assignment
	t.Log("Step 3: Testing industry code assignment...")

	// Verify industry codes are present
	if classificationResult.IndustryCodes == nil {
		t.Error("Expected industry codes to be present")
	} else {
		// Check MCC codes
		if mccCodes, ok := classificationResult.IndustryCodes["mcc"]; ok {
			if mccList, ok := mccCodes.([]interface{}); ok {
				if len(mccList) == 0 {
					t.Error("Expected MCC codes to be present")
				} else {
					t.Logf("✅ MCC codes assigned: %d codes", len(mccList))
				}
			}
		}

		// Check NAICS codes
		if naicsCodes, ok := classificationResult.IndustryCodes["naics"]; ok {
			if naicsList, ok := naicsCodes.([]interface{}); ok {
				if len(naicsList) == 0 {
					t.Error("Expected NAICS codes to be present")
				} else {
					t.Logf("✅ NAICS codes assigned: %d codes", len(naicsList))
				}
			}
		}

		// Check SIC codes
		if sicCodes, ok := classificationResult.IndustryCodes["sic"]; ok {
			if sicList, ok := sicCodes.([]interface{}); ok {
				if len(sicList) == 0 {
					t.Error("Expected SIC codes to be present")
				} else {
					t.Logf("✅ SIC codes assigned: %d codes", len(sicList))
				}
			}
		}
	}

	// Step 4: Test confidence scoring
	t.Log("Step 4: Testing confidence scoring...")

	// Verify confidence scores
	if classificationResult.OverallConfidence < 0.8 {
		t.Errorf("Expected overall confidence >= 0.8, got %f", classificationResult.OverallConfidence)
	}

	// Verify method breakdown confidence scores
	if classificationResult.MethodBreakdown != nil {
		if keywordMethod, ok := classificationResult.MethodBreakdown["keyword"]; ok {
			if keywordData, ok := keywordMethod.(map[string]interface{}); ok {
				if confidence, ok := keywordData["confidence"].(float64); ok {
					if confidence < 0.8 {
						t.Errorf("Expected keyword method confidence >= 0.8, got %f", confidence)
					}
				}
			}
		}
	}

	t.Logf("✅ Confidence scoring completed")

	t.Log("=== Classification Process Journey Completed ===")
}
