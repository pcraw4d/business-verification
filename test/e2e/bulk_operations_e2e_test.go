package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/database"
	"kyb-platform/internal/services"
	"kyb-platform/test/mocks"
)

// TestBulkOperationsE2E tests bulk operations functionality end-to-end
func TestBulkOperationsE2E(t *testing.T) {
	// Skip if not running E2E tests
	if os.Getenv("E2E_TESTS") != "true" {
		t.Skip("Skipping E2E tests - set E2E_TESTS=true to run")
	}

	// Setup test environment
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Create mock services
	mockService := mocks.NewMockMerchantPortfolioService()
	handler := handlers.NewMerchantPortfolioHandler(mockService, nil)

	// Create test server
	server := httptest.NewServer(createTestRouter(handler))
	defer server.Close()

	// Test 1: Bulk portfolio type update
	t.Run("BulkPortfolioTypeUpdate", func(t *testing.T) {
		testBulkPortfolioTypeUpdate(t, server, mockService, ctx)
	})

	// Test 2: Bulk risk level update
	t.Run("BulkRiskLevelUpdate", func(t *testing.T) {
		testBulkRiskLevelUpdate(t, server, mockService, ctx)
	})

	// Test 3: Bulk operations with progress tracking
	t.Run("BulkOperationsWithProgressTracking", func(t *testing.T) {
		testBulkOperationsWithProgressTracking(t, server, mockService, ctx)
	})

	// Test 4: Bulk operations error handling
	t.Run("BulkOperationsErrorHandling", func(t *testing.T) {
		testBulkOperationsErrorHandling(t, server, mockService, ctx)
	})
}

// testBulkPortfolioTypeUpdate tests bulk portfolio type updates
func testBulkPortfolioTypeUpdate(t *testing.T, server *httptest.Server, mockService *mocks.MockMerchantPortfolioService, ctx context.Context) {
	// Create test merchants
	merchants := createTestMerchantsForBulkOps(5)
	for _, merchant := range merchants {
		mockService.AddMerchant(merchant)
	}

	// Extract merchant IDs
	merchantIDs := make([]string, len(merchants))
	for i, merchant := range merchants {
		merchantIDs[i] = merchant.ID
	}

	// Test bulk portfolio type update
	t.Log("Testing bulk portfolio type update...")
	bulkReq := &handlers.BulkUpdateRequest{
		MerchantIDs:   merchantIDs,
		PortfolioType: string(services.PortfolioTypeOnboarded),
	}

	bulkResp, err := makeRequest(t, server, "POST", "/api/v1/merchants/bulk/portfolio-type", bulkReq)
	if err != nil {
		t.Fatalf("Failed to perform bulk portfolio type update: %v", err)
	}

	if bulkResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", bulkResp.StatusCode, bulkResp.Body)
	}

	var bulkResult handlers.BulkOperationResponse
	if err := json.Unmarshal(bulkResp.Body, &bulkResult); err != nil {
		t.Fatalf("Failed to unmarshal bulk operation result: %v", err)
	}

	// Verify bulk operation result
	if bulkResult.OperationID == "" {
		t.Fatal("Expected operation ID to be generated")
	}

	if bulkResult.TotalMerchants != len(merchantIDs) {
		t.Errorf("Expected total merchants %d, got %d", len(merchantIDs), bulkResult.TotalMerchants)
	}

	if bulkResult.SuccessfulUpdates != len(merchantIDs) {
		t.Errorf("Expected successful updates %d, got %d", len(merchantIDs), bulkResult.SuccessfulUpdates)
	}

	if bulkResult.FailedUpdates != 0 {
		t.Errorf("Expected failed updates 0, got %d", bulkResult.FailedUpdates)
	}

	if bulkResult.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", bulkResult.Status)
	}

	t.Logf("✅ Bulk portfolio type update completed successfully: %d merchants updated", bulkResult.SuccessfulUpdates)

	// Verify individual merchant updates
	for _, merchantID := range merchantIDs {
		merchant, err := mockService.GetMerchant(ctx, merchantID)
		if err != nil {
			t.Errorf("Failed to get merchant %s: %v", merchantID, err)
			continue
		}

		if merchant.PortfolioType != services.PortfolioTypeOnboarded {
			t.Errorf("Expected merchant %s portfolio type to be 'onboarded', got '%s'", merchantID, merchant.PortfolioType)
		}
	}

	t.Logf("✅ All merchants verified to have updated portfolio type")
}

// testBulkRiskLevelUpdate tests bulk risk level updates
func testBulkRiskLevelUpdate(t *testing.T, server *httptest.Server, mockService *mocks.MockMerchantPortfolioService, ctx context.Context) {
	// Create test merchants
	merchants := createTestMerchantsForBulkOps(3)
	for _, merchant := range merchants {
		mockService.AddMerchant(merchant)
	}

	// Extract merchant IDs
	merchantIDs := make([]string, len(merchants))
	for i, merchant := range merchants {
		merchantIDs[i] = merchant.ID
	}

	// Test bulk risk level update
	t.Log("Testing bulk risk level update...")
	bulkReq := &handlers.BulkUpdateRequest{
		MerchantIDs: merchantIDs,
		RiskLevel:   string(services.RiskLevelHigh),
	}

	bulkResp, err := makeRequest(t, server, "POST", "/api/v1/merchants/bulk/risk-level", bulkReq)
	if err != nil {
		t.Fatalf("Failed to perform bulk risk level update: %v", err)
	}

	if bulkResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", bulkResp.StatusCode, bulkResp.Body)
	}

	var bulkResult handlers.BulkOperationResponse
	if err := json.Unmarshal(bulkResp.Body, &bulkResult); err != nil {
		t.Fatalf("Failed to unmarshal bulk operation result: %v", err)
	}

	// Verify bulk operation result
	if bulkResult.OperationID == "" {
		t.Fatal("Expected operation ID to be generated")
	}

	if bulkResult.TotalMerchants != len(merchantIDs) {
		t.Errorf("Expected total merchants %d, got %d", len(merchantIDs), bulkResult.TotalMerchants)
	}

	if bulkResult.SuccessfulUpdates != len(merchantIDs) {
		t.Errorf("Expected successful updates %d, got %d", len(merchantIDs), bulkResult.SuccessfulUpdates)
	}

	if bulkResult.FailedUpdates != 0 {
		t.Errorf("Expected failed updates 0, got %d", bulkResult.FailedUpdates)
	}

	if bulkResult.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", bulkResult.Status)
	}

	t.Logf("✅ Bulk risk level update completed successfully: %d merchants updated", bulkResult.SuccessfulUpdates)

	// Verify individual merchant updates
	for _, merchantID := range merchantIDs {
		merchant, err := mockService.GetMerchant(ctx, merchantID)
		if err != nil {
			t.Errorf("Failed to get merchant %s: %v", merchantID, err)
			continue
		}

		if merchant.RiskLevel != services.RiskLevelHigh {
			t.Errorf("Expected merchant %s risk level to be 'high', got '%s'", merchantID, merchant.RiskLevel)
		}
	}

	t.Logf("✅ All merchants verified to have updated risk level")
}

// testBulkOperationsWithProgressTracking tests bulk operations with progress tracking
func testBulkOperationsWithProgressTracking(t *testing.T, server *httptest.Server, mockService *mocks.MockMerchantPortfolioService, ctx context.Context) {
	// Create a larger set of test merchants for progress tracking
	merchants := createTestMerchantsForBulkOps(10)
	for _, merchant := range merchants {
		mockService.AddMerchant(merchant)
	}

	// Extract merchant IDs
	merchantIDs := make([]string, len(merchants))
	for i, merchant := range merchants {
		merchantIDs[i] = merchant.ID
	}

	// Start bulk operation
	t.Log("Starting bulk operation with progress tracking...")
	bulkReq := &handlers.BulkUpdateRequest{
		MerchantIDs:   merchantIDs,
		PortfolioType: string(services.PortfolioTypePending),
	}

	bulkResp, err := makeRequest(t, server, "POST", "/api/v1/merchants/bulk/portfolio-type", bulkReq)
	if err != nil {
		t.Fatalf("Failed to start bulk operation: %v", err)
	}

	if bulkResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", bulkResp.StatusCode, bulkResp.Body)
	}

	var bulkResult handlers.BulkOperationResponse
	if err := json.Unmarshal(bulkResp.Body, &bulkResult); err != nil {
		t.Fatalf("Failed to unmarshal bulk operation result: %v", err)
	}

	operationID := bulkResult.OperationID
	if operationID == "" {
		t.Fatal("Expected operation ID to be generated")
	}

	t.Logf("✅ Bulk operation started with ID: %s", operationID)

	// Track progress
	t.Log("Tracking bulk operation progress...")
	maxAttempts := 10
	attempt := 0

	for attempt < maxAttempts {
		progressResp, err := makeRequest(t, server, "GET", fmt.Sprintf("/api/v1/merchants/bulk/operations/%s/progress", operationID), nil)
		if err != nil {
			t.Fatalf("Failed to get operation progress: %v", err)
		}

		if progressResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d: %s", progressResp.StatusCode, progressResp.Body)
		}

		var progress handlers.BulkOperationProgressResponse
		if err := json.Unmarshal(progressResp.Body, &progress); err != nil {
			t.Fatalf("Failed to unmarshal progress response: %v", err)
		}

		t.Logf("Progress: %d/%d completed, Status: %s", progress.CompletedUpdates, progress.TotalMerchants, progress.Status)

		if progress.Status == "completed" {
			// Verify final results
			if progress.CompletedUpdates != len(merchantIDs) {
				t.Errorf("Expected completed updates %d, got %d", len(merchantIDs), progress.CompletedUpdates)
			}

			if progress.FailedUpdates != 0 {
				t.Errorf("Expected failed updates 0, got %d", progress.FailedUpdates)
			}

			t.Logf("✅ Bulk operation completed successfully: %d merchants updated", progress.CompletedUpdates)
			break
		}

		if progress.Status == "failed" {
			t.Fatalf("Bulk operation failed: %s", progress.ErrorMessage)
		}

		// Wait before next check
		time.Sleep(100 * time.Millisecond)
		attempt++
	}

	if attempt >= maxAttempts {
		t.Fatal("Bulk operation did not complete within expected time")
	}
}

// testBulkOperationsErrorHandling tests bulk operations error handling
func testBulkOperationsErrorHandling(t *testing.T, server *httptest.Server, mockService *mocks.MockMerchantPortfolioService, ctx context.Context) {
	// Test 1: Bulk operation with non-existent merchants
	t.Log("Test 1: Bulk operation with non-existent merchants...")
	bulkReq := &handlers.BulkUpdateRequest{
		MerchantIDs:   []string{"nonexistent_1", "nonexistent_2", "nonexistent_3"},
		PortfolioType: string(services.PortfolioTypeOnboarded),
	}

	bulkResp, err := makeRequest(t, server, "POST", "/api/v1/merchants/bulk/portfolio-type", bulkReq)
	if err != nil {
		t.Fatalf("Failed to perform bulk operation with non-existent merchants: %v", err)
	}

	if bulkResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", bulkResp.StatusCode, bulkResp.Body)
	}

	var bulkResult handlers.BulkOperationResponse
	if err := json.Unmarshal(bulkResp.Body, &bulkResult); err != nil {
		t.Fatalf("Failed to unmarshal bulk operation result: %v", err)
	}

	// Should have all failed updates
	if bulkResult.SuccessfulUpdates != 0 {
		t.Errorf("Expected successful updates 0, got %d", bulkResult.SuccessfulUpdates)
	}

	if bulkResult.FailedUpdates != 3 {
		t.Errorf("Expected failed updates 3, got %d", bulkResult.FailedUpdates)
	}

	if bulkResult.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", bulkResult.Status)
	}

	t.Logf("✅ Bulk operation with non-existent merchants handled correctly: %d failed", bulkResult.FailedUpdates)

	// Test 2: Bulk operation with empty merchant list
	t.Log("Test 2: Bulk operation with empty merchant list...")
	emptyReq := &handlers.BulkUpdateRequest{
		MerchantIDs:   []string{},
		PortfolioType: string(services.PortfolioTypeOnboarded),
	}

	emptyResp, err := makeRequest(t, server, "POST", "/api/v1/merchants/bulk/portfolio-type", emptyReq)
	if err != nil {
		t.Fatalf("Failed to perform bulk operation with empty merchant list: %v", err)
	}

	if emptyResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d: %s", emptyResp.StatusCode, emptyResp.Body)
	}

	t.Logf("✅ Bulk operation with empty merchant list handled correctly")

	// Test 3: Bulk operation with invalid portfolio type
	t.Log("Test 3: Bulk operation with invalid portfolio type...")
	// Create a test merchant first
	testMerchant := createTestMerchantForBulkOps("invalid_type_test")
	mockService.AddMerchant(testMerchant)

	invalidReq := &handlers.BulkUpdateRequest{
		MerchantIDs:   []string{testMerchant.ID},
		PortfolioType: "invalid_type",
	}

	invalidResp, err := makeRequest(t, server, "POST", "/api/v1/merchants/bulk/portfolio-type", invalidReq)
	if err != nil {
		t.Fatalf("Failed to perform bulk operation with invalid portfolio type: %v", err)
	}

	if invalidResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d: %s", invalidResp.StatusCode, invalidResp.Body)
	}

	t.Logf("✅ Bulk operation with invalid portfolio type handled correctly")

	// Test 4: Get progress for non-existent operation
	t.Log("Test 4: Get progress for non-existent operation...")
	progressResp, err := makeRequest(t, server, "GET", "/api/v1/merchants/bulk/operations/nonexistent_operation/progress", nil)
	if err != nil {
		t.Fatalf("Failed to get progress for non-existent operation: %v", err)
	}

	if progressResp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d: %s", progressResp.StatusCode, progressResp.Body)
	}

	t.Logf("✅ Progress request for non-existent operation handled correctly")
}

// Helper functions for bulk operations testing

func createTestMerchantsForBulkOps(count int) []*services.Merchant {
	merchants := make([]*services.Merchant, count)
	for i := 0; i < count; i++ {
		merchants[i] = createTestMerchantForBulkOps(fmt.Sprintf("bulk_test_merchant_%d", i+1))
	}
	return merchants
}

func createTestMerchantForBulkOps(id string) *services.Merchant {
	now := time.Now()
	return &services.Merchant{
		ID:                 id,
		Name:               fmt.Sprintf("Bulk Test Company %s", id),
		LegalName:          fmt.Sprintf("Bulk Test Company %s LLC", id),
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
		PortfolioType:    services.PortfolioTypeProspective,
		RiskLevel:        services.RiskLevelMedium,
		ComplianceStatus: "pending",
		Status:           "active",
		CreatedBy:        "test_user",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}
