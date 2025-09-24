package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/database"
	"kyb-platform/internal/services"
	"kyb-platform/test/mocks"
)

// TestCompleteUserOnboardingWorkflow tests the complete user onboarding journey
// This test covers: user registration, first merchant creation, initial setup
func TestCompleteUserOnboardingWorkflow(t *testing.T) {
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

	t.Run("Complete User Onboarding Journey", func(t *testing.T) {
		testCompleteUserOnboardingJourney(t, server, mockPortfolioService, ctx)
	})
}

// testCompleteUserOnboardingJourney tests the complete user onboarding journey
func testCompleteUserOnboardingJourney(
	t *testing.T,
	server *httptest.Server,
	mockPortfolioService *mocks.MockMerchantPortfolioService,
	ctx context.Context,
) {
	t.Log("=== Complete User Onboarding Journey ===")

	// Step 1: User registration (simulated)
	t.Log("Step 1: User registration...")
	userID := "test_user_123"
	t.Logf("✅ User registered: %s", userID)

	// Step 2: Create first merchant
	t.Log("Step 2: Creating first merchant...")
	merchant := createOnboardingTestMerchant("onboarding_merchant_1", "New User Company", services.PortfolioTypeProspective, services.RiskLevelMedium)

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

	createResp, respBody, err := makeOnboardingRequest(server, "POST", "/api/v1/merchants", createReq)
	if err != nil {
		t.Fatalf("Failed to create first merchant: %v", err)
	}

	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d: %s", createResp.StatusCode, string(respBody))
	}

	var createdMerchant handlers.MerchantResponse
	if err := json.Unmarshal(respBody, &createdMerchant); err != nil {
		t.Fatalf("Failed to unmarshal created merchant: %v", err)
	}

	t.Logf("✅ First merchant created: %s (ID: %s)", createdMerchant.Name, createdMerchant.ID)

	// Step 3: Start merchant session
	t.Log("Step 3: Starting merchant session...")
	sessionResp, sessionBody, err := makeOnboardingRequest(server, "POST", fmt.Sprintf("/api/v1/merchants/%s/session", createdMerchant.ID), nil)
	if err != nil {
		t.Fatalf("Failed to start merchant session: %v", err)
	}

	if sessionResp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d: %s", sessionResp.StatusCode, string(sessionBody))
	}

	var session handlers.SessionResponse
	if err := json.Unmarshal(sessionBody, &session); err != nil {
		t.Fatalf("Failed to unmarshal session: %v", err)
	}

	t.Logf("✅ Merchant session started: %s", session.ID)

	// Step 4: View merchant dashboard
	t.Log("Step 4: Viewing merchant dashboard...")
	dashboardResp, dashboardBody, err := makeOnboardingRequest(server, "GET", fmt.Sprintf("/api/v1/merchants/%s/dashboard", createdMerchant.ID), nil)
	if err != nil {
		t.Fatalf("Failed to get merchant dashboard: %v", err)
	}

	if dashboardResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", dashboardResp.StatusCode, string(dashboardBody))
	}

	var dashboard handlers.MerchantDashboardResponse
	if err := json.Unmarshal(dashboardBody, &dashboard); err != nil {
		t.Fatalf("Failed to unmarshal dashboard: %v", err)
	}

	if dashboard.MerchantID != createdMerchant.ID {
		t.Errorf("Expected merchant ID %s, got %s", createdMerchant.ID, dashboard.MerchantID)
	}

	t.Logf("✅ Merchant dashboard viewed successfully")

	// Step 5: Update merchant information
	t.Log("Step 5: Updating merchant information...")
	updateReq := &handlers.UpdateMerchantRequest{
		Name:          "Updated New User Company",
		EmployeeCount: 75,
		PortfolioType: string(services.PortfolioTypePending),
		RiskLevel:     string(services.RiskLevelLow),
	}

	updateResp, updateBody, err := makeOnboardingRequest(server, "PUT", fmt.Sprintf("/api/v1/merchants/%s", createdMerchant.ID), updateReq)
	if err != nil {
		t.Fatalf("Failed to update merchant: %v", err)
	}

	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", updateResp.StatusCode, string(updateBody))
	}

	t.Logf("✅ Merchant information updated successfully")

	// Step 6: End the session
	t.Log("Step 6: Ending merchant session...")
	endSessionResp, endSessionBody, err := makeOnboardingRequest(server, "DELETE", "/api/v1/merchants/session/active", nil)
	if err != nil {
		t.Fatalf("Failed to end merchant session: %v", err)
	}

	if endSessionResp.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected status 204, got %d: %s", endSessionResp.StatusCode, string(endSessionBody))
	}

	t.Logf("✅ Merchant session ended successfully")

	t.Log("=== Complete User Onboarding Journey Completed ===")
}

// Helper functions

func createOnboardingTestMerchant(id, name string, portfolioType services.PortfolioType, riskLevel services.RiskLevel) *services.Merchant {
	now := time.Now()
	return &services.Merchant{
		ID:                 id,
		Name:               name,
		LegalName:          fmt.Sprintf("%s LLC", name),
		RegistrationNumber: fmt.Sprintf("REG%s", id),
		TaxID:              fmt.Sprintf("TAX%s", id),
		Industry:           "Technology",
		IndustryCode:       "541511",
		BusinessType:       "LLC",
		FoundedDate:        &now,
		EmployeeCount:      50,
		Address: database.Address{
			Street1:    "123 Test St",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "USA",
		},
		ContactInfo: database.ContactInfo{
			Phone:   "+1-555-123-4567",
			Email:   fmt.Sprintf("test@%s.com", id),
			Website: fmt.Sprintf("https://%s.com", id),
		},
		PortfolioType:    portfolioType,
		RiskLevel:        riskLevel,
		ComplianceStatus: "pending",
		Status:           "active",
		CreatedBy:        "test_user",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func createUserOnboardingTestRouter(
	portfolioHandler *handlers.MerchantPortfolioHandler,
	classificationHandler *handlers.ClassificationHandler,
	riskHandler *handlers.RiskAssessmentHandler,
	comparisonHandler *handlers.ComparisonHandler,
) http.Handler {
	mux := http.NewServeMux()

	// Merchant portfolio routes
	mux.HandleFunc("POST /api/v1/merchants", portfolioHandler.CreateMerchant)
	mux.HandleFunc("GET /api/v1/merchants/{id}", portfolioHandler.GetMerchant)
	mux.HandleFunc("PUT /api/v1/merchants/{id}", portfolioHandler.UpdateMerchant)
	mux.HandleFunc("DELETE /api/v1/merchants/{id}", portfolioHandler.DeleteMerchant)
	mux.HandleFunc("GET /api/v1/merchants", portfolioHandler.ListMerchants)
	mux.HandleFunc("POST /api/v1/merchants/search", portfolioHandler.SearchMerchants)
	mux.HandleFunc("GET /api/v1/merchants/{id}/dashboard", portfolioHandler.GetMerchantDashboard)

	// Session management routes
	mux.HandleFunc("POST /api/v1/merchants/{id}/session", portfolioHandler.StartMerchantSession)
	mux.HandleFunc("GET /api/v1/merchants/session/active", portfolioHandler.GetActiveMerchantSession)
	mux.HandleFunc("DELETE /api/v1/merchants/session/active", portfolioHandler.EndMerchantSession)

	// Classification routes
	mux.HandleFunc("POST /api/v1/classification", classificationHandler.ClassifyBusiness)

	// Risk assessment routes
	mux.HandleFunc("POST /api/v1/risk-assessment", riskHandler.AssessRisk)

	// Comparison routes
	mux.HandleFunc("POST /api/v1/merchants/compare", comparisonHandler.CompareMerchants)
	mux.HandleFunc("GET /api/v1/merchants/compare/{id}", comparisonHandler.GetComparison)
	mux.HandleFunc("POST /api/v1/merchants/compare/report", comparisonHandler.GenerateReport)

	return mux
}

func makeOnboardingRequest(server *httptest.Server, method, path string, body interface{}) (*http.Response, []byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, server.URL+path, reqBody)
	if err != nil {
		return nil, nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, nil, err
	}

	return resp, respBody, nil
}
