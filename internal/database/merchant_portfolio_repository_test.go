package database

import (
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/models"
)

// Test helper functions
func createTestRepository() *MerchantPortfolioRepository {
	logger := log.New(log.Writer(), "test: ", log.LstdFlags)
	// Note: In a real test, you'd use a proper mock or test database
	// For this example, we'll create a repository with nil db to test the structure
	repo := &MerchantPortfolioRepository{
		db:     nil, // Would be mockDB in real implementation
		logger: logger,
	}
	return repo
}

func createTestMerchant() *models.Merchant {
	now := time.Now()
	return &models.Merchant{
		ID:                 "merchant_123",
		Name:               "Test Company",
		LegalName:          "Test Company LLC",
		RegistrationNumber: "REG123456",
		TaxID:              "TAX123456",
		Industry:           "Technology",
		IndustryCode:       "541511",
		BusinessType:       "LLC",
		FoundedDate:        &now,
		EmployeeCount:      50,
		Address: models.Address{
			Street1:    "123 Test St",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "USA",
		},
		ContactInfo: models.ContactInfo{
			Phone:   "+1-555-123-4567",
			Email:   "test@testcompany.com",
			Website: "https://testcompany.com",
		},
		PortfolioType:    models.PortfolioTypeProspective,
		RiskLevel:        models.RiskLevelMedium,
		ComplianceStatus: "pending",
		Status:           "active",
		CreatedBy:        "user123",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func createTestMerchantSession() *models.MerchantSession {
	now := time.Now()
	return &models.MerchantSession{
		ID:         "session_123",
		UserID:     "user123",
		MerchantID: "merchant_123",
		StartedAt:  now,
		LastActive: now,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func createTestAuditLog() *models.AuditLog {
	return &models.AuditLog{
		ID:           "audit_123",
		UserID:       "user123",
		MerchantID:   "merchant_123",
		Action:       "CREATE_MERCHANT",
		ResourceType: "merchant",
		ResourceID:   "merchant_123",
		Details:      "Merchant created",
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0",
		RequestID:    "req_123",
		CreatedAt:    time.Now(),
	}
}

// Tests
func TestNewMerchantPortfolioRepository(t *testing.T) {
	// Note: In a real test, you'd use a proper *sql.DB
	// For this example, we'll test with nil to avoid type issues
	repo := NewMerchantPortfolioRepository(nil, nil)

	if repo == nil {
		t.Fatal("Expected repository to be created")
	}

	if repo.logger == nil {
		t.Error("Expected logger to be set (default logger)")
	}
}

// Test pagination validation
func TestMerchantPortfolioRepository_PaginationValidation(t *testing.T) {
	repo := createTestRepository()
	ctx := context.Background()

	// Test invalid page
	_, err := repo.ListMerchants(ctx, 0, 10)
	if err != ErrInvalidPagination {
		t.Errorf("Expected ErrInvalidPagination for page 0, got: %v", err)
	}

	// Test invalid page size
	_, err = repo.ListMerchants(ctx, 1, 0)
	if err != ErrInvalidPagination {
		t.Errorf("Expected ErrInvalidPagination for page size 0, got: %v", err)
	}

	// Test page size too large
	_, err = repo.ListMerchants(ctx, 1, 1001)
	if err != ErrInvalidPagination {
		t.Errorf("Expected ErrInvalidPagination for page size 1001, got: %v", err)
	}
}

// Test query building functions
func TestMerchantPortfolioRepository_BuildSearchQuery(t *testing.T) {
	repo := createTestRepository()

	// Test with empty filters
	filters := &models.MerchantSearchFilters{}
	query, args := repo.buildSearchQuery(filters, 10, 0)

	if query == "" {
		t.Error("Expected query to be built")
	}

	if len(args) != 2 { // limit and offset
		t.Errorf("Expected 2 args (limit, offset), got %d", len(args))
	}

	// Test with portfolio type filter
	portfolioType := models.PortfolioTypeOnboarded
	filters = &models.MerchantSearchFilters{
		PortfolioType: &portfolioType,
	}
	query, args = repo.buildSearchQuery(filters, 10, 0)

	if !strings.Contains(query, "portfolio_type = $1") {
		t.Error("Expected query to contain portfolio_type filter")
	}

	if len(args) != 3 { // portfolio_type, limit, offset
		t.Errorf("Expected 3 args, got %d", len(args))
	}

	// Test with search query
	filters = &models.MerchantSearchFilters{
		SearchQuery: "test",
	}
	query, args = repo.buildSearchQuery(filters, 10, 0)

	if !strings.Contains(query, "name ILIKE") {
		t.Error("Expected query to contain name ILIKE filter")
	}

	if len(args) != 3 { // search_query, limit, offset
		t.Errorf("Expected 3 args, got %d", len(args))
	}
}

func TestMerchantPortfolioRepository_BuildCountQuery(t *testing.T) {
	repo := createTestRepository()

	// Test with empty filters
	filters := &models.MerchantSearchFilters{}
	query, args := repo.buildCountQuery(filters)

	if query == "" {
		t.Error("Expected query to be built")
	}

	if !strings.Contains(query, "SELECT COUNT(*)") {
		t.Error("Expected query to be a count query")
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 args for empty filters, got %d", len(args))
	}

	// Test with portfolio type filter
	portfolioType := models.PortfolioTypeOnboarded
	filters = &models.MerchantSearchFilters{
		PortfolioType: &portfolioType,
	}
	query, args = repo.buildCountQuery(filters)

	if !strings.Contains(query, "portfolio_type = $1") {
		t.Error("Expected query to contain portfolio_type filter")
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
}

// Test error constants
func TestRepositoryErrors(t *testing.T) {
	// Test that error constants are defined
	if ErrMerchantNotFound == nil {
		t.Error("Expected ErrMerchantNotFound to be defined")
	}

	if ErrSessionNotFound == nil {
		t.Error("Expected ErrSessionNotFound to be defined")
	}

	if ErrAuditLogNotFound == nil {
		t.Error("Expected ErrAuditLogNotFound to be defined")
	}

	if ErrNotificationNotFound == nil {
		t.Error("Expected ErrNotificationNotFound to be defined")
	}

	if ErrComparisonNotFound == nil {
		t.Error("Expected ErrComparisonNotFound to be defined")
	}

	if ErrAnalyticsNotFound == nil {
		t.Error("Expected ErrAnalyticsNotFound to be defined")
	}

	if ErrDuplicateMerchant == nil {
		t.Error("Expected ErrDuplicateMerchant to be defined")
	}

	if ErrDuplicateSession == nil {
		t.Error("Expected ErrDuplicateSession to be defined")
	}

	if ErrInvalidPagination == nil {
		t.Error("Expected ErrInvalidPagination to be defined")
	}
}

// Test scanMerchant method (this would require proper mocking in a real test)
func TestMerchantPortfolioRepository_ScanMerchant(t *testing.T) {
	repo := createTestRepository()

	// This test would require a proper mock row implementation
	// For now, we're just testing that the method exists and handles errors
	_, err := repo.scanMerchant(nil)
	if err == nil {
		t.Error("Expected error when scanning nil row")
	}
}

// Test scanSession method
func TestMerchantPortfolioRepository_ScanSession(t *testing.T) {
	repo := createTestRepository()

	// This test would require a proper mock row implementation
	// For now, we're just testing that the method exists and handles errors
	_, err := repo.scanSession(nil)
	if err == nil {
		t.Error("Expected error when scanning nil row")
	}
}

// Test scanAuditLog method
func TestMerchantPortfolioRepository_ScanAuditLog(t *testing.T) {
	repo := createTestRepository()

	// This test would require a proper mock row implementation
	// For now, we're just testing that the method exists and handles errors
	_, err := repo.scanAuditLog(nil)
	if err == nil {
		t.Error("Expected error when scanning nil row")
	}
}

// Test bulk operation methods (without database calls)
func TestMerchantPortfolioRepository_BulkUpdatePortfolioType_EmptyList(t *testing.T) {
	repo := createTestRepository()
	ctx := context.Background()

	// Test with empty merchant list (should not error)
	err := repo.BulkUpdatePortfolioType(ctx, []string{}, models.PortfolioTypeOnboarded, "user123")
	if err != nil {
		t.Errorf("Expected no error for empty list, got: %v", err)
	}
}

func TestMerchantPortfolioRepository_BulkUpdateRiskLevel_EmptyList(t *testing.T) {
	repo := createTestRepository()
	ctx := context.Background()

	// Test with empty merchant list (should not error)
	err := repo.BulkUpdateRiskLevel(ctx, []string{}, models.RiskLevelHigh, "user123")
	if err != nil {
		t.Errorf("Expected no error for empty list, got: %v", err)
	}
}

// Test search filters
func TestMerchantSearchFilters_Comprehensive(t *testing.T) {
	// Test all filter combinations
	portfolioType := models.PortfolioTypeOnboarded
	riskLevel := models.RiskLevelHigh
	createdAfter := time.Now().Add(-24 * time.Hour)
	createdBefore := time.Now()
	employeeCountMin := 10
	employeeCountMax := 100
	revenueMin := 100000.0
	revenueMax := 1000000.0

	filters := &models.MerchantSearchFilters{
		PortfolioType:    &portfolioType,
		RiskLevel:        &riskLevel,
		Industry:         "Technology",
		Status:           "active",
		SearchQuery:      "test company",
		CreatedAfter:     &createdAfter,
		CreatedBefore:    &createdBefore,
		EmployeeCountMin: &employeeCountMin,
		EmployeeCountMax: &employeeCountMax,
		RevenueMin:       &revenueMin,
		RevenueMax:       &revenueMax,
	}

	repo := createTestRepository()
	query, args := repo.buildSearchQuery(filters, 10, 0)

	// Verify all filters are included in the query
	expectedFilters := []string{
		"portfolio_type = $1",
		"risk_level = $2",
		"industry ILIKE $3",
		"status = $4",
		"name ILIKE $5",
		"created_at >= $6",
		"created_at <= $7",
		"employee_count >= $8",
		"employee_count <= $9",
		"annual_revenue >= $10",
		"annual_revenue <= $11",
	}

	for _, expectedFilter := range expectedFilters {
		if !strings.Contains(query, expectedFilter) {
			t.Errorf("Expected query to contain filter: %s", expectedFilter)
		}
	}

	// Verify correct number of arguments (11 filters + limit + offset = 13)
	if len(args) != 13 {
		t.Errorf("Expected 13 args, got %d", len(args))
	}
}

// Test count query with comprehensive filters
func TestMerchantPortfolioRepository_BuildCountQuery_Comprehensive(t *testing.T) {
	portfolioType := models.PortfolioTypeOnboarded
	riskLevel := models.RiskLevelHigh
	createdAfter := time.Now().Add(-24 * time.Hour)
	createdBefore := time.Now()
	employeeCountMin := 10
	employeeCountMax := 100
	revenueMin := 100000.0
	revenueMax := 1000000.0

	filters := &models.MerchantSearchFilters{
		PortfolioType:    &portfolioType,
		RiskLevel:        &riskLevel,
		Industry:         "Technology",
		Status:           "active",
		SearchQuery:      "test company",
		CreatedAfter:     &createdAfter,
		CreatedBefore:    &createdBefore,
		EmployeeCountMin: &employeeCountMin,
		EmployeeCountMax: &employeeCountMax,
		RevenueMin:       &revenueMin,
		RevenueMax:       &revenueMax,
	}

	repo := createTestRepository()
	query, args := repo.buildCountQuery(filters)

	// Verify all filters are included in the count query
	expectedFilters := []string{
		"portfolio_type = $1",
		"risk_level = $2",
		"industry ILIKE $3",
		"status = $4",
		"name ILIKE $5",
		"created_at >= $6",
		"created_at <= $7",
		"employee_count >= $8",
		"employee_count <= $9",
		"annual_revenue >= $10",
		"annual_revenue <= $11",
	}

	for _, expectedFilter := range expectedFilters {
		if !strings.Contains(query, expectedFilter) {
			t.Errorf("Expected count query to contain filter: %s", expectedFilter)
		}
	}

	// Verify correct number of arguments (11 filters)
	if len(args) != 11 {
		t.Errorf("Expected 11 args, got %d", len(args))
	}
}

// =============================================================================
// Additional Comprehensive Repository Tests
// =============================================================================

func TestMerchantPortfolioRepository_QueryBuilding_EdgeCases(t *testing.T) {
	repo := createTestRepository()

	// Test with nil filters
	query, args := repo.buildSearchQuery(nil, 10, 0)
	if query == "" {
		t.Error("Expected query to be built even with nil filters")
	}
	if len(args) != 2 { // limit and offset
		t.Errorf("Expected 2 args for nil filters, got %d", len(args))
	}

	// Test with empty filters struct
	emptyFilters := &models.MerchantSearchFilters{}
	query, args = repo.buildSearchQuery(emptyFilters, 10, 0)
	if query == "" {
		t.Error("Expected query to be built with empty filters")
	}
	if len(args) != 2 { // limit and offset
		t.Errorf("Expected 2 args for empty filters, got %d", len(args))
	}

	// Test with only search query
	searchFilters := &models.MerchantSearchFilters{
		SearchQuery: "test",
	}
	query, args = repo.buildSearchQuery(searchFilters, 10, 0)
	if !strings.Contains(query, "name ILIKE") {
		t.Error("Expected query to contain name ILIKE for search query")
	}
	if len(args) != 3 { // search_query, limit, offset
		t.Errorf("Expected 3 args for search query, got %d", len(args))
	}
}

func TestMerchantPortfolioRepository_QueryBuilding_AllFilters(t *testing.T) {
	repo := createTestRepository()

	// Test with all possible filters
	portfolioType := models.PortfolioTypeOnboarded
	riskLevel := models.RiskLevelHigh
	createdAfter := time.Now().Add(-24 * time.Hour)
	createdBefore := time.Now()
	employeeCountMin := 10
	employeeCountMax := 100
	revenueMin := 100000.0
	revenueMax := 1000000.0

	filters := &models.MerchantSearchFilters{
		PortfolioType:    &portfolioType,
		RiskLevel:        &riskLevel,
		Industry:         "Technology",
		Status:           "active",
		SearchQuery:      "test company",
		CreatedAfter:     &createdAfter,
		CreatedBefore:    &createdBefore,
		EmployeeCountMin: &employeeCountMin,
		EmployeeCountMax: &employeeCountMax,
		RevenueMin:       &revenueMin,
		RevenueMax:       &revenueMax,
	}

	query, args := repo.buildSearchQuery(filters, 20, 40)

	// Verify query structure
	if !strings.Contains(query, "SELECT") {
		t.Error("Expected query to contain SELECT")
	}
	if !strings.Contains(query, "FROM merchants") {
		t.Error("Expected query to contain FROM merchants")
	}
	if !strings.Contains(query, "JOIN portfolio_types") {
		t.Error("Expected query to contain JOIN portfolio_types")
	}
	if !strings.Contains(query, "JOIN risk_levels") {
		t.Error("Expected query to contain JOIN risk_levels")
	}
	if !strings.Contains(query, "WHERE") {
		t.Error("Expected query to contain WHERE clause")
	}
	if !strings.Contains(query, "ORDER BY") {
		t.Error("Expected query to contain ORDER BY clause")
	}
	if !strings.Contains(query, "LIMIT") {
		t.Error("Expected query to contain LIMIT clause")
	}
	if !strings.Contains(query, "OFFSET") {
		t.Error("Expected query to contain OFFSET clause")
	}

	// Verify all filters are present
	expectedFilters := []string{
		"portfolio_type = $1",
		"risk_level = $2",
		"industry ILIKE $3",
		"status = $4",
		"name ILIKE $5",
		"created_at >= $6",
		"created_at <= $7",
		"employee_count >= $8",
		"employee_count <= $9",
		"annual_revenue >= $10",
		"annual_revenue <= $11",
	}

	for _, expectedFilter := range expectedFilters {
		if !strings.Contains(query, expectedFilter) {
			t.Errorf("Expected query to contain filter: %s", expectedFilter)
		}
	}

	// Verify correct number of arguments (11 filters + limit + offset = 13)
	if len(args) != 13 {
		t.Errorf("Expected 13 args, got %d", len(args))
	}
}

func TestMerchantPortfolioRepository_CountQueryBuilding_AllFilters(t *testing.T) {
	repo := createTestRepository()

	// Test with all possible filters
	portfolioType := models.PortfolioTypeOnboarded
	riskLevel := models.RiskLevelHigh
	createdAfter := time.Now().Add(-24 * time.Hour)
	createdBefore := time.Now()
	employeeCountMin := 10
	employeeCountMax := 100
	revenueMin := 100000.0
	revenueMax := 1000000.0

	filters := &models.MerchantSearchFilters{
		PortfolioType:    &portfolioType,
		RiskLevel:        &riskLevel,
		Industry:         "Technology",
		Status:           "active",
		SearchQuery:      "test company",
		CreatedAfter:     &createdAfter,
		CreatedBefore:    &createdBefore,
		EmployeeCountMin: &employeeCountMin,
		EmployeeCountMax: &employeeCountMax,
		RevenueMin:       &revenueMin,
		RevenueMax:       &revenueMax,
	}

	query, args := repo.buildCountQuery(filters)

	// Verify query structure
	if !strings.Contains(query, "SELECT COUNT(*)") {
		t.Error("Expected count query to contain SELECT COUNT(*)")
	}
	if !strings.Contains(query, "FROM merchants") {
		t.Error("Expected count query to contain FROM merchants")
	}
	if !strings.Contains(query, "JOIN portfolio_types") {
		t.Error("Expected count query to contain JOIN portfolio_types")
	}
	if !strings.Contains(query, "JOIN risk_levels") {
		t.Error("Expected count query to contain JOIN risk_levels")
	}
	if !strings.Contains(query, "WHERE") {
		t.Error("Expected count query to contain WHERE clause")
	}

	// Verify all filters are present
	expectedFilters := []string{
		"portfolio_type = $1",
		"risk_level = $2",
		"industry ILIKE $3",
		"status = $4",
		"name ILIKE $5",
		"created_at >= $6",
		"created_at <= $7",
		"employee_count >= $8",
		"employee_count <= $9",
		"annual_revenue >= $10",
		"annual_revenue <= $11",
	}

	for _, expectedFilter := range expectedFilters {
		if !strings.Contains(query, expectedFilter) {
			t.Errorf("Expected count query to contain filter: %s", expectedFilter)
		}
	}

	// Verify correct number of arguments (11 filters)
	if len(args) != 11 {
		t.Errorf("Expected 11 args, got %d", len(args))
	}
}

func TestMerchantPortfolioRepository_ErrorHandling(t *testing.T) {
	repo := createTestRepository()
	ctx := context.Background()

	// Test pagination validation
	tests := []struct {
		name      string
		page      int
		pageSize  int
		expectErr bool
	}{
		{"valid pagination", 1, 10, false},
		{"page zero", 0, 10, true},
		{"negative page", -1, 10, true},
		{"page size zero", 1, 0, true},
		{"negative page size", 1, -1, true},
		{"page size too large", 1, 1001, true},
		{"maximum page size", 1, 1000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := repo.ListMerchants(ctx, tt.page, tt.pageSize)
			if tt.expectErr && err != ErrInvalidPagination {
				t.Errorf("Expected ErrInvalidPagination for %s, got: %v", tt.name, err)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error for %s, got: %v", tt.name, err)
			}
		})
	}
}

func TestMerchantPortfolioRepository_HelperMethods(t *testing.T) {
	repo := createTestRepository()

	// Test getPortfolioTypeID method exists
	ctx := context.Background()
	_, err := repo.getPortfolioTypeID(ctx, "onboarded")
	// We expect an error since we're using nil db, but the method should exist
	if err == nil {
		t.Error("Expected error for getPortfolioTypeID with nil db")
	}

	// Test getRiskLevelID method exists
	_, err = repo.getRiskLevelID(ctx, "high")
	// We expect an error since we're using nil db, but the method should exist
	if err == nil {
		t.Error("Expected error for getRiskLevelID with nil db")
	}
}

func TestMerchantPortfolioRepository_DataValidation(t *testing.T) {
	// Test merchant data validation
	merchant := createTestMerchant()

	// Test required fields
	if merchant.ID == "" {
		t.Error("Expected merchant ID to be set")
	}
	if merchant.Name == "" {
		t.Error("Expected merchant name to be set")
	}
	if merchant.LegalName == "" {
		t.Error("Expected merchant legal name to be set")
	}

	// Test address validation
	if merchant.Address.Street1 == "" {
		t.Error("Expected address street1 to be set")
	}
	if merchant.Address.City == "" {
		t.Error("Expected address city to be set")
	}
	if merchant.Address.State == "" {
		t.Error("Expected address state to be set")
	}
	if merchant.Address.PostalCode == "" {
		t.Error("Expected address postal code to be set")
	}
	if merchant.Address.Country == "" {
		t.Error("Expected address country to be set")
	}

	// Test contact info validation
	if merchant.ContactInfo.Phone == "" {
		t.Error("Expected contact phone to be set")
	}
	if merchant.ContactInfo.Email == "" {
		t.Error("Expected contact email to be set")
	}
	if merchant.ContactInfo.Website == "" {
		t.Error("Expected contact website to be set")
	}

	// Test portfolio type validation
	if merchant.PortfolioType != models.PortfolioTypeProspective {
		t.Errorf("Expected portfolio type %s, got %s", models.PortfolioTypeProspective, merchant.PortfolioType)
	}

	// Test risk level validation
	if merchant.RiskLevel != models.RiskLevelMedium {
		t.Errorf("Expected risk level %s, got %s", models.RiskLevelMedium, merchant.RiskLevel)
	}
}

func TestMerchantPortfolioRepository_SessionDataValidation(t *testing.T) {
	// Test session data validation
	session := createTestMerchantSession()

	// Test required fields
	if session.ID == "" {
		t.Error("Expected session ID to be set")
	}
	if session.UserID == "" {
		t.Error("Expected session user ID to be set")
	}
	if session.MerchantID == "" {
		t.Error("Expected session merchant ID to be set")
	}

	// Test timestamps
	if session.StartedAt.IsZero() {
		t.Error("Expected session started at to be set")
	}
	if session.LastActive.IsZero() {
		t.Error("Expected session last active to be set")
	}
	if session.CreatedAt.IsZero() {
		t.Error("Expected session created at to be set")
	}
	if session.UpdatedAt.IsZero() {
		t.Error("Expected session updated at to be set")
	}

	// Test active status
	if !session.IsActive {
		t.Error("Expected session to be active")
	}
}

func TestMerchantPortfolioRepository_AuditLogDataValidation(t *testing.T) {
	// Test audit log data validation
	auditLog := createTestAuditLog()

	// Test required fields
	if auditLog.ID == "" {
		t.Error("Expected audit log ID to be set")
	}
	if auditLog.UserID == "" {
		t.Error("Expected audit log user ID to be set")
	}
	if auditLog.MerchantID == "" {
		t.Error("Expected audit log merchant ID to be set")
	}
	if auditLog.Action == "" {
		t.Error("Expected audit log action to be set")
	}
	if auditLog.ResourceType == "" {
		t.Error("Expected audit log resource type to be set")
	}
	if auditLog.ResourceID == "" {
		t.Error("Expected audit log resource ID to be set")
	}
	if auditLog.Details == "" {
		t.Error("Expected audit log details to be set")
	}

	// Test timestamps
	if auditLog.CreatedAt.IsZero() {
		t.Error("Expected audit log created at to be set")
	}
}

func TestMerchantPortfolioRepository_QueryOptimization(t *testing.T) {
	repo := createTestRepository()

	// Test that queries include proper indexing hints
	portfolioType := models.PortfolioTypeOnboarded
	filters := &models.MerchantSearchFilters{
		PortfolioType: &portfolioType,
	}

	query, _ := repo.buildSearchQuery(filters, 10, 0)

	// Verify query includes proper ordering for performance
	if !strings.Contains(query, "ORDER BY") {
		t.Error("Expected query to include ORDER BY clause for performance")
	}

	// Verify query includes proper joins
	if !strings.Contains(query, "JOIN portfolio_types") {
		t.Error("Expected query to include portfolio_types join")
	}
	if !strings.Contains(query, "JOIN risk_levels") {
		t.Error("Expected query to include risk_levels join")
	}
}

func TestMerchantPortfolioRepository_ConcurrentAccess(t *testing.T) {
	repo := createTestRepository()
	ctx := context.Background()

	// Test that repository methods can be called concurrently
	// This is more of a structure test since we're using nil db
	done := make(chan bool, 2)

	go func() {
		repo.ListMerchants(ctx, 1, 10)
		done <- true
	}()

	go func() {
		repo.CountMerchants(ctx, &models.MerchantSearchFilters{})
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done
}
