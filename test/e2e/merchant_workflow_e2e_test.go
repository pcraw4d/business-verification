package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/database"
	"kyb-platform/internal/models"
	"kyb-platform/internal/services"
	"kyb-platform/test/mocks"
)

// TestMerchantWorkflowE2E tests complete merchant workflows end-to-end
func TestMerchantWorkflowE2E(t *testing.T) {
	// Skip if not running E2E tests
	if os.Getenv("E2E_TESTS") != "true" {
		t.Skip("Skipping E2E tests - set E2E_TESTS=true to run")
	}

	// Setup test environment
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create mock services
	mockService := mocks.NewMockMerchantPortfolioService()
	handler := handlers.NewMerchantPortfolioHandler(mockService, nil)

	// Create test server
	server := httptest.NewServer(createTestRouter(handler))
	defer server.Close()

	// Test 1: Complete merchant lifecycle workflow
	t.Run("CompleteMerchantLifecycle", func(t *testing.T) {
		testCompleteMerchantLifecycle(t, server, mockService, ctx)
	})

	// Test 2: Merchant session management workflow
	t.Run("MerchantSessionManagement", func(t *testing.T) {
		testMerchantSessionManagement(t, server, mockService, ctx)
	})

	// Test 3: Merchant search and filtering workflow
	t.Run("MerchantSearchAndFiltering", func(t *testing.T) {
		testMerchantSearchAndFiltering(t, server, mockService, ctx)
	})

	// Test 4: Error handling and edge cases
	t.Run("ErrorHandlingAndEdgeCases", func(t *testing.T) {
		testErrorHandlingAndEdgeCases(t, server, mockService, ctx)
	})
}

// testCompleteMerchantLifecycle tests the complete merchant lifecycle
func testCompleteMerchantLifecycle(t *testing.T, server *httptest.Server, mockService *mocks.MockMerchantPortfolioService, ctx context.Context) {
	// Step 1: Create a new merchant
	t.Log("Step 1: Creating new merchant...")
	merchant := createTestMerchant()
	merchant.ID = "lifecycle_test_merchant_1"
	merchant.Name = "Lifecycle Test Company"
	merchant.LegalName = "Lifecycle Test Company LLC"

	createReq := &handlers.CreateMerchantRequest{
		Name:               merchant.Name,
		LegalName:          merchant.LegalName,
		RegistrationNumber: merchant.RegistrationNumber,
		TaxID:              merchant.TaxID,
		Industry:           merchant.Industry,
		IndustryCode:       merchant.IndustryCode,
		BusinessType:       merchant.BusinessType,
		EmployeeCount:      merchant.EmployeeCount,
		Address:            models.Address(merchant.Address),
		ContactInfo:        models.ContactInfo(merchant.ContactInfo),
		PortfolioType:      models.PortfolioType(merchant.PortfolioType),
		RiskLevel:          models.RiskLevel(merchant.RiskLevel),
	}

	createResp, respBody, err := makeRequest(t, server, "POST", "/api/v1/merchants", createReq)
	if err != nil {
		t.Fatalf("Failed to create merchant: %v", err)
	}

	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d: %s", createResp.StatusCode, string(respBody))
	}

	var createdMerchant handlers.MerchantResponse
	if err := json.Unmarshal(respBody, &createdMerchant); err != nil {
		t.Fatalf("Failed to unmarshal created merchant: %v", err)
	}

	if createdMerchant.ID == "" {
		t.Fatal("Expected merchant ID to be generated")
	}

	t.Logf("✅ Merchant created successfully with ID: %s", createdMerchant.ID)

	// Step 2: Retrieve the merchant
	t.Log("Step 2: Retrieving merchant...")
	getResp, getBody, err := makeRequest(t, server, "GET", fmt.Sprintf("/api/v1/merchants/%s", createdMerchant.ID), nil)
	if err != nil {
		t.Fatalf("Failed to get merchant: %v", err)
	}

	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", getResp.StatusCode, string(getBody))
	}

	var retrievedMerchant handlers.MerchantResponse
	if err := json.Unmarshal(getBody, &retrievedMerchant); err != nil {
		t.Fatalf("Failed to unmarshal retrieved merchant: %v", err)
	}

	if retrievedMerchant.Name != merchant.Name {
		t.Errorf("Expected name %s, got %s", merchant.Name, retrievedMerchant.Name)
	}

	t.Logf("✅ Merchant retrieved successfully: %s", retrievedMerchant.Name)

	// Step 3: Update the merchant
	t.Log("Step 3: Updating merchant...")
	updateReq := &handlers.UpdateMerchantRequest{
		Name:          stringPtr("Updated Lifecycle Test Company"),
		EmployeeCount: intPtr(100),
		PortfolioType: (*models.PortfolioType)(&[]services.PortfolioType{services.PortfolioTypeOnboarded}[0]),
		RiskLevel:     (*models.RiskLevel)(&[]services.RiskLevel{services.RiskLevelLow}[0]),
	}

	updateResp, updateBody, err := makeRequest(t, server, "PUT", fmt.Sprintf("/api/v1/merchants/%s", createdMerchant.ID), updateReq)
	if err != nil {
		t.Fatalf("Failed to update merchant: %v", err)
	}

	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", updateResp.StatusCode, string(updateBody))
	}

	var updatedMerchant handlers.MerchantResponse
	if err := json.Unmarshal(updateBody, &updatedMerchant); err != nil {
		t.Fatalf("Failed to unmarshal updated merchant: %v", err)
	}

	if updatedMerchant.Name != *updateReq.Name {
		t.Errorf("Expected updated name %s, got %s", *updateReq.Name, updatedMerchant.Name)
	}

	if updatedMerchant.EmployeeCount != *updateReq.EmployeeCount {
		t.Errorf("Expected updated employee count %d, got %d", *updateReq.EmployeeCount, updatedMerchant.EmployeeCount)
	}

	t.Logf("✅ Merchant updated successfully: %s", updatedMerchant.Name)

	// Step 4: Delete the merchant
	t.Log("Step 4: Deleting merchant...")
	deleteResp, _, err := makeRequest(t, server, "DELETE", fmt.Sprintf("/api/v1/merchants/%s", createdMerchant.ID), nil)
	if err != nil {
		t.Fatalf("Failed to delete merchant: %v", err)
	}

	if deleteResp.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected status 204, got %d", deleteResp.StatusCode)
	}

	t.Logf("✅ Merchant deleted successfully")

	// Step 5: Verify merchant is deleted
	t.Log("Step 5: Verifying merchant deletion...")
	getDeletedResp, _, err := makeRequest(t, server, "GET", fmt.Sprintf("/api/v1/merchants/%s", createdMerchant.ID), nil)
	if err != nil {
		t.Fatalf("Failed to get deleted merchant: %v", err)
	}

	if getDeletedResp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", getDeletedResp.StatusCode)
	}

	t.Logf("✅ Merchant deletion verified - merchant not found")
}

// testMerchantSessionManagement tests merchant session management
func testMerchantSessionManagement(t *testing.T, server *httptest.Server, mockService *mocks.MockMerchantPortfolioService, ctx context.Context) {
	// Create test merchant
	merchant := createTestMerchant()
	merchant.ID = "session_test_merchant_1"
	mockService.AddMerchant(merchant)

	// Step 1: Start a merchant session
	t.Log("Step 1: Starting merchant session...")
	startSessionResp, startBody, err := makeRequest(t, server, "POST", fmt.Sprintf("/api/v1/merchants/%s/session", merchant.ID), nil)
	if err != nil {
		t.Fatalf("Failed to start merchant session: %v", err)
	}

	if startSessionResp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d: %s", startSessionResp.StatusCode, string(startBody))
	}

	var session handlers.SessionResponse
	if err := json.Unmarshal(startBody, &session); err != nil {
		t.Fatalf("Failed to unmarshal session: %v", err)
	}

	if session.MerchantID != merchant.ID {
		t.Errorf("Expected merchant ID %s, got %s", merchant.ID, session.MerchantID)
	}

	if !session.IsActive {
		t.Error("Expected session to be active")
	}

	t.Logf("✅ Merchant session started successfully: %s", session.SessionID)

	// Step 2: Get active session
	t.Log("Step 2: Getting active session...")
	getSessionResp, getBody, err := makeRequest(t, server, "GET", "/api/v1/merchants/session/active", nil)
	if err != nil {
		t.Fatalf("Failed to get active session: %v", err)
	}

	if getSessionResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", getSessionResp.StatusCode, string(getBody))
	}

	var activeSession handlers.SessionResponse
	if err := json.Unmarshal(getBody, &activeSession); err != nil {
		t.Fatalf("Failed to unmarshal active session: %v", err)
	}

	if activeSession.SessionID != session.SessionID {
		t.Errorf("Expected session ID %s, got %s", session.SessionID, activeSession.SessionID)
	}

	t.Logf("✅ Active session retrieved successfully: %s", activeSession.SessionID)

	// Step 3: End the session
	t.Log("Step 3: Ending merchant session...")
	endSessionResp, _, err := makeRequest(t, server, "DELETE", "/api/v1/merchants/session/active", nil)
	if err != nil {
		t.Fatalf("Failed to end merchant session: %v", err)
	}

	if endSessionResp.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected status 204, got %d", endSessionResp.StatusCode)
	}

	t.Logf("✅ Merchant session ended successfully")

	// Step 4: Verify no active session
	t.Log("Step 4: Verifying no active session...")
	getNoSessionResp, _, err := makeRequest(t, server, "GET", "/api/v1/merchants/session/active", nil)
	if err != nil {
		t.Fatalf("Failed to get active session: %v", err)
	}

	if getNoSessionResp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", getNoSessionResp.StatusCode)
	}

	t.Logf("✅ No active session verified")
}

// testMerchantSearchAndFiltering tests merchant search and filtering
func testMerchantSearchAndFiltering(t *testing.T, server *httptest.Server, mockService *mocks.MockMerchantPortfolioService, ctx context.Context) {
	// Create test merchants with different portfolio types and risk levels
	merchants := []*services.Merchant{
		createTestMerchantWithType("search_test_1", "Tech Company A", services.PortfolioTypeOnboarded, services.RiskLevelLow),
		createTestMerchantWithType("search_test_2", "Retail Company B", services.PortfolioTypeProspective, services.RiskLevelMedium),
		createTestMerchantWithType("search_test_3", "Service Company C", services.PortfolioTypeOnboarded, services.RiskLevelHigh),
		createTestMerchantWithType("search_test_4", "Manufacturing Company D", services.PortfolioTypePending, services.RiskLevelLow),
	}

	for _, merchant := range merchants {
		mockService.AddMerchant(merchant)
	}

	// Set up search results
	mockService.SetSearchResults(merchants, len(merchants))

	// Test 1: List all merchants
	t.Log("Test 1: Listing all merchants...")
	listResp, listBody, err := makeRequest(t, server, "GET", "/api/v1/merchants?page=1&page_size=10", nil)
	if err != nil {
		t.Fatalf("Failed to list merchants: %v", err)
	}

	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", listResp.StatusCode, string(listBody))
	}

	var listResult handlers.MerchantListResponse
	if err := json.Unmarshal(listBody, &listResult); err != nil {
		t.Fatalf("Failed to unmarshal merchant list: %v", err)
	}

	if len(listResult.Merchants) != 4 {
		t.Errorf("Expected 4 merchants, got %d", len(listResult.Merchants))
	}

	if listResult.Total != 4 {
		t.Errorf("Expected total 4, got %d", listResult.Total)
	}

	t.Logf("✅ Listed all merchants successfully: %d merchants", len(listResult.Merchants))

	// Test 2: Filter by portfolio type
	t.Log("Test 2: Filtering by portfolio type...")
	filterResp, filterBody, err := makeRequest(t, server, "GET", "/api/v1/merchants?portfolio_type=onboarded&page=1&page_size=10", nil)
	if err != nil {
		t.Fatalf("Failed to filter merchants: %v", err)
	}

	if filterResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", filterResp.StatusCode, string(filterBody))
	}

	var filterResult handlers.MerchantListResponse
	if err := json.Unmarshal(filterBody, &filterResult); err != nil {
		t.Fatalf("Failed to unmarshal filtered merchant list: %v", err)
	}

	// Should have 2 onboarded merchants
	if len(filterResult.Merchants) != 2 {
		t.Errorf("Expected 2 onboarded merchants, got %d", len(filterResult.Merchants))
	}

	t.Logf("✅ Filtered merchants by portfolio type successfully: %d onboarded merchants", len(filterResult.Merchants))

	// Test 3: Search by name
	t.Log("Test 3: Searching merchants by name...")
	searchReq := &handlers.MerchantSearchRequest{
		Query:    "Tech Company",
		Page:     1,
		PageSize: 10,
	}

	searchResp, searchBody, err := makeRequest(t, server, "POST", "/api/v1/merchants/search", searchReq)
	if err != nil {
		t.Fatalf("Failed to search merchants: %v", err)
	}

	if searchResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", searchResp.StatusCode, string(searchBody))
	}

	var searchResult handlers.MerchantListResponse
	if err := json.Unmarshal(searchBody, &searchResult); err != nil {
		t.Fatalf("Failed to unmarshal search results: %v", err)
	}

	// Should find at least one tech company
	if len(searchResult.Merchants) == 0 {
		t.Error("Expected to find at least one tech company")
	}

	t.Logf("✅ Searched merchants by name successfully: %d results", len(searchResult.Merchants))
}

// testErrorHandlingAndEdgeCases tests error handling and edge cases
func testErrorHandlingAndEdgeCases(t *testing.T, server *httptest.Server, mockService *mocks.MockMerchantPortfolioService, ctx context.Context) {
	// Test 1: Get non-existent merchant
	t.Log("Test 1: Getting non-existent merchant...")
	getResp, _, err := makeRequest(t, server, "GET", "/api/v1/merchants/nonexistent_merchant", nil)
	if err != nil {
		t.Fatalf("Failed to get non-existent merchant: %v", err)
	}

	if getResp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", getResp.StatusCode)
	}

	t.Logf("✅ Non-existent merchant handled correctly")

	// Test 2: Create merchant with invalid data
	t.Log("Test 2: Creating merchant with invalid data...")
	invalidReq := &handlers.CreateMerchantRequest{
		Name: "", // Invalid: empty name
	}

	createResp, _, err := makeRequest(t, server, "POST", "/api/v1/merchants", invalidReq)
	if err != nil {
		t.Fatalf("Failed to create invalid merchant: %v", err)
	}

	if createResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", createResp.StatusCode)
	}

	t.Logf("✅ Invalid merchant data handled correctly")

	// Test 3: Update non-existent merchant
	t.Log("Test 3: Updating non-existent merchant...")
	updateReq := &handlers.UpdateMerchantRequest{
		Name: stringPtr("Updated Name"),
	}

	updateResp, _, err := makeRequest(t, server, "PUT", "/api/v1/merchants/nonexistent_merchant", updateReq)
	if err != nil {
		t.Fatalf("Failed to update non-existent merchant: %v", err)
	}

	if updateResp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", updateResp.StatusCode)
	}

	t.Logf("✅ Update of non-existent merchant handled correctly")

	// Test 4: Delete non-existent merchant
	t.Log("Test 4: Deleting non-existent merchant...")
	deleteResp, _, err := makeRequest(t, server, "DELETE", "/api/v1/merchants/nonexistent_merchant", nil)
	if err != nil {
		t.Fatalf("Failed to delete non-existent merchant: %v", err)
	}

	if deleteResp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", deleteResp.StatusCode)
	}

	t.Logf("✅ Delete of non-existent merchant handled correctly")
}

// Helper functions

func createTestMerchant() *services.Merchant {
	now := time.Now()
	return &services.Merchant{
		ID:                 "test_merchant_1",
		Name:               "Test Company",
		LegalName:          "Test Company LLC",
		RegistrationNumber: "REG123456",
		TaxID:              "TAX123456",
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
			Email:   "test@testcompany.com",
			Website: "https://testcompany.com",
		},
		PortfolioType:    services.PortfolioTypeProspective,
		RiskLevel:        services.RiskLevelMedium,
		ComplianceStatus: "pending",
		Status:           "active",
		CreatedBy:        "test_user",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func createTestMerchantWithType(id, name string, portfolioType services.PortfolioType, riskLevel services.RiskLevel) *services.Merchant {
	merchant := createTestMerchant()
	merchant.ID = id
	merchant.Name = name
	merchant.PortfolioType = portfolioType
	merchant.RiskLevel = riskLevel
	return merchant
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func makeRequest(t *testing.T, server *httptest.Server, method, path string, body interface{}) (*http.Response, []byte, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal request body: %v", err)
		}
	}

	req, err := http.NewRequest(method, server.URL+path, bytes.NewReader(reqBody))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "test_user") // Mock user ID for testing

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to make request: %v", err)
	}

	respBody := make([]byte, 0)
	if resp.Body != nil {
		respBody, err = readAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return resp, respBody, fmt.Errorf("failed to read response body: %v", err)
		}
	}

	return resp, respBody, nil
}

func readAll(body io.Reader) ([]byte, error) {
	return io.ReadAll(body)
}

func createTestRouter(handler *handlers.MerchantPortfolioHandler) http.Handler {
	mux := http.NewServeMux()

	// Merchant portfolio routes
	mux.HandleFunc("POST /api/v1/merchants", handler.CreateMerchant)
	mux.HandleFunc("GET /api/v1/merchants/{id}", handler.GetMerchant)
	mux.HandleFunc("PUT /api/v1/merchants/{id}", handler.UpdateMerchant)
	mux.HandleFunc("DELETE /api/v1/merchants/{id}", handler.DeleteMerchant)
	mux.HandleFunc("GET /api/v1/merchants", handler.ListMerchants)
	mux.HandleFunc("POST /api/v1/merchants/search", handler.SearchMerchants)

	// Session management routes
	mux.HandleFunc("POST /api/v1/merchants/{id}/session", handler.StartMerchantSession)
	mux.HandleFunc("GET /api/v1/merchants/session/active", handler.GetActiveSession)
	mux.HandleFunc("DELETE /api/v1/merchants/session/active", handler.EndMerchantSession)

	return mux
}
