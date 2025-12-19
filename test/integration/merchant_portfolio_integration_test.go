//go:build !comprehensive_test
// +build !comprehensive_test

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/api/routes"
	"kyb-platform/internal/database"
	"kyb-platform/internal/models"
	"kyb-platform/internal/services"
)

// TestDatabase represents the test database connection
type TestDatabase struct {
	db     *sql.DB
	logger *log.Logger
}

// IntegrationTestConfig holds test configuration
type IntegrationTestConfig struct {
	DatabaseURL string
	TestUserID  string
}

var (
	testDB     *TestDatabase
	testConfig *IntegrationTestConfig
)

// SetupTestDatabase initializes the test database
func SetupTestDatabase() (*TestDatabase, error) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:password@localhost:5432/kyb_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open test database: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping test database: %w", err)
	}

	logger := log.New(os.Stdout, "[TEST-DB] ", log.LstdFlags)

	return &TestDatabase{
		db:     db,
		logger: logger,
	}, nil
}

// CleanupTestData removes all test data from the database
func (tdb *TestDatabase) CleanupTestData() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Clean up in reverse order of dependencies
	queries := []string{
		"DELETE FROM audit_logs WHERE user_id = $1",
		"DELETE FROM merchant_sessions WHERE user_id = $1",
		"DELETE FROM merchants WHERE user_id = $1",
		"DELETE FROM users WHERE id = $1",
	}

	for _, query := range queries {
		if _, err := tdb.db.ExecContext(ctx, query, testConfig.TestUserID); err != nil {
			tdb.logger.Printf("Warning: failed to cleanup %s: %v", query, err)
		}
	}

	return nil
}

// SetupTestUser creates a test user in the database
func (tdb *TestDatabase) SetupTestUser() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO users (id, email, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO NOTHING
	`

	now := time.Now()
	_, err := tdb.db.ExecContext(ctx, query,
		testConfig.TestUserID,
		"test@example.com",
		"Test User",
		now,
		now,
	)

	return err
}

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	var err error

	// Setup test configuration
	testConfig = &IntegrationTestConfig{
		DatabaseURL: os.Getenv("TEST_DATABASE_URL"),
		TestUserID:  "test-user-123",
	}

	if testConfig.DatabaseURL == "" {
		testConfig.DatabaseURL = "postgres://postgres:password@localhost:5432/kyb_test?sslmode=disable"
	}

	// Setup test database
	testDB, err = SetupTestDatabase()
	if err != nil {
		log.Fatalf("Failed to setup test database: %v", err)
	}
	defer testDB.db.Close()

	// Setup test user
	if err := testDB.SetupTestUser(); err != nil {
		log.Fatalf("Failed to setup test user: %v", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if err := testDB.CleanupTestData(); err != nil {
		log.Printf("Failed to cleanup test data: %v", err)
	}

	os.Exit(code)
}

// =============================================================================
// Database Integration Tests
// =============================================================================

func TestMerchantPortfolioRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	repo := database.NewMerchantPortfolioRepository(testDB.db, testDB.logger)

	t.Run("CreateMerchant", func(t *testing.T) {
		ctx := context.Background()

		merchant := &models.Merchant{
			ID:                 "test-merchant-1",
			Name:               "Test Merchant 1",
			LegalName:          "Test Merchant 1 LLC",
			RegistrationNumber: "REG123456",
			TaxID:              "TAX123456",
			Industry:           "Technology",
			IndustryCode:       "541511",
			BusinessType:       "LLC",
			EmployeeCount:      10,
			CreatedBy:          testConfig.TestUserID,
			PortfolioType:      models.PortfolioTypeOnboarded,
			RiskLevel:          models.RiskLevelMedium,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		// Create merchant
		err := repo.CreateMerchant(ctx, merchant)
		if err != nil {
			t.Fatalf("Failed to create merchant: %v", err)
		}

		// Verify merchant was created
		retrievedMerchant, err := repo.GetMerchant(ctx, merchant.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve merchant: %v", err)
		}

		if retrievedMerchant.Name != merchant.Name {
			t.Errorf("Expected merchant name %s, got %s", merchant.Name, retrievedMerchant.Name)
		}
	})

	t.Run("SearchMerchants", func(t *testing.T) {
		ctx := context.Background()

		// Create test merchants
		merchants := []*models.Merchant{
			{
				ID:            "test-merchant-search-1",
				Name:          "Search Test Merchant 1",
				Industry:      "Technology",
				PortfolioType: models.PortfolioTypeOnboarded,
				RiskLevel:     models.RiskLevelHigh,
				CreatedBy:     testConfig.TestUserID,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
			{
				ID:            "test-merchant-search-2",
				Name:          "Search Test Merchant 2",
				Industry:      "Finance",
				PortfolioType: models.PortfolioTypeProspective,
				RiskLevel:     models.RiskLevelLow,
				CreatedBy:     testConfig.TestUserID,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		}

		for _, merchant := range merchants {
			err := repo.CreateMerchant(ctx, merchant)
			if err != nil {
				t.Fatalf("Failed to create test merchant: %v", err)
			}
		}

		// Test search by name
		filters := &models.MerchantSearchFilters{
			SearchQuery: "Search Test",
		}

		results, err := repo.SearchMerchants(ctx, filters, 1, 10)
		if err != nil {
			t.Fatalf("Failed to search merchants: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 merchants, got %d", len(results))
		}

		// Test search by portfolio type
		portfolioType := models.PortfolioTypeOnboarded
		filters = &models.MerchantSearchFilters{
			PortfolioType: &portfolioType,
		}

		results, err = repo.SearchMerchants(ctx, filters, 1, 10)
		if err != nil {
			t.Fatalf("Failed to search merchants by portfolio type: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 onboarded merchant, got %d", len(results))
		}
	})

	t.Run("BulkUpdatePortfolioType", func(t *testing.T) {
		ctx := context.Background()

		// Create test merchants for bulk update
		merchantIDs := []string{}
		for i := 1; i <= 3; i++ {
			merchant := &models.Merchant{
				ID:            fmt.Sprintf("test-bulk-merchant-%d", i),
				Name:          fmt.Sprintf("Bulk Test Merchant %d", i),
				PortfolioType: models.PortfolioTypePending,
				RiskLevel:     models.RiskLevelMedium,
				CreatedBy:     testConfig.TestUserID,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}

			err := repo.CreateMerchant(ctx, merchant)
			if err != nil {
				t.Fatalf("Failed to create bulk test merchant: %v", err)
			}

			merchantIDs = append(merchantIDs, merchant.ID)
		}

		// Perform bulk update
		err := repo.BulkUpdatePortfolioType(ctx, merchantIDs, models.PortfolioTypeOnboarded, testConfig.TestUserID)
		if err != nil {
			t.Fatalf("Failed to bulk update portfolio type: %v", err)
		}

		// Verify updates
		for _, merchantID := range merchantIDs {
			merchant, err := repo.GetMerchant(ctx, merchantID)
			if err != nil {
				t.Fatalf("Failed to retrieve updated merchant: %v", err)
			}

			if merchant.PortfolioType != models.PortfolioTypeOnboarded {
				t.Errorf("Expected portfolio type %s, got %s", models.PortfolioTypeOnboarded, merchant.PortfolioType)
			}
		}
	})
}

// =============================================================================
// Service Integration Tests
// =============================================================================

func TestMerchantPortfolioService_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	mockDB := NewMockDatabase(testDB.db, testDB.logger)
	service := services.NewMerchantPortfolioService(mockDB, testDB.logger)

	t.Run("CreateAndRetrieveMerchant", func(t *testing.T) {
		ctx := context.Background()

		merchant := &services.Merchant{
			ID:                 "test-service-merchant-1",
			Name:               "Service Test Merchant 1",
			LegalName:          "Service Test Merchant 1 LLC",
			RegistrationNumber: "SVC123456",
			TaxID:              "SVC123456",
			Industry:           "Technology",
			IndustryCode:       "541511",
			BusinessType:       "LLC",
			EmployeeCount:      15,
			PortfolioType:      services.PortfolioTypeOnboarded,
			RiskLevel:          services.RiskLevelMedium,
		}

		// Create merchant through service
		createdMerchant, err := service.CreateMerchant(ctx, merchant, testConfig.TestUserID)
		if err != nil {
			t.Fatalf("Failed to create merchant through service: %v", err)
		}

		if createdMerchant.ID != merchant.ID {
			t.Errorf("Expected merchant ID %s, got %s", merchant.ID, createdMerchant.ID)
		}

		// Retrieve merchant through service
		retrievedMerchant, err := service.GetMerchant(ctx, merchant.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve merchant through service: %v", err)
		}

		if retrievedMerchant.Name != merchant.Name {
			t.Errorf("Expected merchant name %s, got %s", merchant.Name, retrievedMerchant.Name)
		}
	})

	t.Run("SearchMerchants", func(t *testing.T) {
		ctx := context.Background()

		// Create test merchants
		merchants := []*services.Merchant{
			{
				ID:            "test-service-search-1",
				Name:          "Service Search Test 1",
				Industry:      "Technology",
				PortfolioType: services.PortfolioTypeOnboarded,
				RiskLevel:     services.RiskLevelHigh,
			},
			{
				ID:            "test-service-search-2",
				Name:          "Service Search Test 2",
				Industry:      "Finance",
				PortfolioType: services.PortfolioTypeProspective,
				RiskLevel:     services.RiskLevelLow,
			},
		}

		for _, merchant := range merchants {
			_, err := service.CreateMerchant(ctx, merchant, testConfig.TestUserID)
			if err != nil {
				t.Fatalf("Failed to create service test merchant: %v", err)
			}
		}

		// Test search
		filters := &services.MerchantSearchFilters{
			SearchQuery: "Service Search",
		}

		results, err := service.SearchMerchants(ctx, filters, 1, 10)
		if err != nil {
			t.Fatalf("Failed to search merchants through service: %v", err)
		}

		if len(results.Merchants) != 2 {
			t.Errorf("Expected 2 merchants, got %d", len(results.Merchants))
		}
	})

	t.Run("BulkOperations", func(t *testing.T) {
		ctx := context.Background()

		// Create test merchants for bulk operations
		merchantIDs := []string{}
		for i := 1; i <= 5; i++ {
			merchant := &services.Merchant{
				ID:            fmt.Sprintf("test-bulk-service-%d", i),
				Name:          fmt.Sprintf("Bulk Service Test %d", i),
				PortfolioType: services.PortfolioTypePending,
				RiskLevel:     services.RiskLevelMedium,
			}

			createdMerchant, err := service.CreateMerchant(ctx, merchant, testConfig.TestUserID)
			if err != nil {
				t.Fatalf("Failed to create bulk service test merchant: %v", err)
			}

			merchantIDs = append(merchantIDs, createdMerchant.ID)
		}

		// Test bulk update portfolio type
		result, err := service.BulkUpdatePortfolioType(ctx, merchantIDs, services.PortfolioTypeOnboarded, testConfig.TestUserID)
		if err != nil {
			t.Fatalf("Failed to bulk update portfolio type through service: %v", err)
		}

		if result.Successful != 5 {
			t.Errorf("Expected 5 successful updates, got %d", result.Successful)
		}

		// Test bulk update risk level
		result, err = service.BulkUpdateRiskLevel(ctx, merchantIDs, services.RiskLevelHigh, testConfig.TestUserID)
		if err != nil {
			t.Fatalf("Failed to bulk update risk level through service: %v", err)
		}

		if result.Successful != 5 {
			t.Errorf("Expected 5 successful risk level updates, got %d", result.Successful)
		}
	})

	t.Run("SessionManagement", func(t *testing.T) {
		ctx := context.Background()

		// Create a test merchant
		merchant := &services.Merchant{
			ID:            "test-session-merchant",
			Name:          "Session Test Merchant",
			PortfolioType: services.PortfolioTypeOnboarded,
			RiskLevel:     services.RiskLevelMedium,
		}

		createdMerchant, err := service.CreateMerchant(ctx, merchant, testConfig.TestUserID)
		if err != nil {
			t.Fatalf("Failed to create session test merchant: %v", err)
		}

		// Start session
		session, err := service.StartMerchantSession(ctx, testConfig.TestUserID, createdMerchant.ID)
		if err != nil {
			t.Fatalf("Failed to start merchant session: %v", err)
		}

		if session.MerchantID != createdMerchant.ID {
			t.Errorf("Expected session merchant ID %s, got %s", createdMerchant.ID, session.MerchantID)
		}

		// Get active session
		activeSession, err := service.GetActiveMerchantSession(ctx, testConfig.TestUserID)
		if err != nil {
			t.Fatalf("Failed to get active merchant session: %v", err)
		}

		if activeSession.MerchantID != createdMerchant.ID {
			t.Errorf("Expected active session merchant ID %s, got %s", createdMerchant.ID, activeSession.MerchantID)
		}

		// End session
		err = service.EndMerchantSession(ctx, testConfig.TestUserID)
		if err != nil {
			t.Fatalf("Failed to end merchant session: %v", err)
		}
	})
}

// =============================================================================
// API Integration Tests
// =============================================================================

func TestMerchantPortfolioAPI_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup service and handler
	mockDB := NewMockDatabase(testDB.db, testDB.logger)
	service := services.NewMerchantPortfolioService(mockDB, testDB.logger)
	handler := handlers.NewMerchantPortfolioHandler(service, testDB.logger)

	// Setup router
	mux := http.NewServeMux()
	config := &routes.MerchantRouteConfig{
		MerchantPortfolioHandler: handler,
		Logger:                   nil, // Use default logger for testing
		EnableBulkOperations:     true,
		EnableSessionManagement:  true,
		MaxBulkOperationSize:     1000,
	}
	routes.RegisterMerchantRoutes(mux, config)

	t.Run("CreateMerchantAPI", func(t *testing.T) {
		merchantData := map[string]interface{}{
			"id":                  "test-api-merchant-1",
			"name":                "API Test Merchant 1",
			"legal_name":          "API Test Merchant 1 LLC",
			"registration_number": "API123456",
			"tax_id":              "API123456",
			"industry":            "Technology",
			"industry_code":       "541511",
			"business_type":       "LLC",
			"employee_count":      20,
			"portfolio_type":      "onboarded",
			"risk_level":          "medium",
		}

		jsonData, err := json.Marshal(merchantData)
		if err != nil {
			t.Fatalf("Failed to marshal merchant data: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/v1/merchants", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", testConfig.TestUserID)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d. Response: %s", http.StatusCreated, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["id"] != merchantData["id"] {
			t.Errorf("Expected merchant ID %s, got %s", merchantData["id"], response["id"])
		}
	})

	t.Run("GetMerchantAPI", func(t *testing.T) {
		// First create a merchant
		merchantData := map[string]interface{}{
			"id":             "test-api-get-merchant",
			"name":           "API Get Test Merchant",
			"portfolio_type": "onboarded",
			"risk_level":     "medium",
		}

		jsonData, err := json.Marshal(merchantData)
		if err != nil {
			t.Fatalf("Failed to marshal merchant data: %v", err)
		}

		createReq := httptest.NewRequest("POST", "/api/v1/merchants", bytes.NewBuffer(jsonData))
		createReq.Header.Set("Content-Type", "application/json")
		createReq.Header.Set("X-User-ID", testConfig.TestUserID)

		createW := httptest.NewRecorder()
		mux.ServeHTTP(createW, createReq)

		if createW.Code != http.StatusCreated {
			t.Fatalf("Failed to create merchant for get test: %s", createW.Body.String())
		}

		// Now get the merchant
		getReq := httptest.NewRequest("GET", "/api/v1/merchants/test-api-get-merchant", nil)
		getReq.Header.Set("X-User-ID", testConfig.TestUserID)

		getW := httptest.NewRecorder()
		mux.ServeHTTP(getW, getReq)

		if getW.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, getW.Code, getW.Body.String())
		}

		var response map[string]interface{}
		if err := json.Unmarshal(getW.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["id"] != merchantData["id"] {
			t.Errorf("Expected merchant ID %s, got %s", merchantData["id"], response["id"])
		}
	})
}

// =============================================================================
// Error Handling Integration Tests
// =============================================================================

func TestMerchantPortfolioErrorHandling_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	mockDB := NewMockDatabase(testDB.db, testDB.logger)
	service := services.NewMerchantPortfolioService(mockDB, testDB.logger)

	t.Run("GetNonExistentMerchant", func(t *testing.T) {
		ctx := context.Background()

		_, err := service.GetMerchant(ctx, "non-existent-merchant")
		if err == nil {
			t.Error("Expected error when getting non-existent merchant")
		}
	})

	t.Run("CreateMerchantWithInvalidData", func(t *testing.T) {
		ctx := context.Background()

		// Create merchant with invalid data
		merchant := &services.Merchant{
			ID:   "", // Invalid: empty ID
			Name: "", // Invalid: empty name
		}

		_, err := service.CreateMerchant(ctx, merchant, testConfig.TestUserID)
		if err == nil {
			t.Error("Expected error when creating merchant with invalid data")
		}
	})
}

// =============================================================================
// Performance Integration Tests
// =============================================================================

func TestMerchantPortfolioPerformance_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	mockDB := NewMockDatabase(testDB.db, testDB.logger)
	service := services.NewMerchantPortfolioService(mockDB, testDB.logger)

	t.Run("BulkCreatePerformance", func(t *testing.T) {
		ctx := context.Background()
		start := time.Now()

		// Create 100 merchants
		for i := 1; i <= 100; i++ {
			merchant := &services.Merchant{
				ID:            fmt.Sprintf("perf-test-merchant-%d", i),
				Name:          fmt.Sprintf("Performance Test Merchant %d", i),
				PortfolioType: services.PortfolioTypeOnboarded,
				RiskLevel:     services.RiskLevelMedium,
			}

			_, err := service.CreateMerchant(ctx, merchant, testConfig.TestUserID)
			if err != nil {
				t.Fatalf("Failed to create performance test merchant %d: %v", i, err)
			}
		}

		duration := time.Since(start)
		t.Logf("Created 100 merchants in %v", duration)

		// Should complete within 10 seconds
		if duration > 10*time.Second {
			t.Errorf("Bulk create took too long: %v", duration)
		}
	})

	t.Run("SearchPerformance", func(t *testing.T) {
		ctx := context.Background()

		// Search with pagination
		filters := &services.MerchantSearchFilters{
			SearchQuery: "Performance",
		}

		start := time.Now()
		results, err := service.SearchMerchants(ctx, filters, 1, 50)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Failed to search merchants: %v", err)
		}

		t.Logf("Searched %d merchants in %v", len(results.Merchants), duration)

		// Should complete within 1 second
		if duration > 1*time.Second {
			t.Errorf("Search took too long: %v", duration)
		}
	})
}
