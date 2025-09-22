package database

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"kyb-platform/internal/models"
)

// MockMerchantDatabase provides an in-memory mock database for testing
type MockMerchantDatabase struct {
	merchants map[string]*models.Merchant
	sessions  map[string]*models.MerchantSession
	auditLogs []*models.AuditLog
	nextID    int
	logger    *log.Logger
}

// NewMockMerchantDatabase creates a new mock merchant database with realistic test data
func NewMockMerchantDatabase(logger *log.Logger) *MockMerchantDatabase {
	if logger == nil {
		logger = log.Default()
	}

	mdb := &MockMerchantDatabase{
		merchants: make(map[string]*models.Merchant),
		sessions:  make(map[string]*models.MerchantSession),
		auditLogs: make([]*models.AuditLog, 0),
		nextID:    1,
		logger:    logger,
	}

	// Initialize with realistic mock data
	mdb.initializeMockData()
	return mdb
}

// initializeMockData populates the mock database with 100 realistic merchants
func (mdb *MockMerchantDatabase) initializeMockData() {
	mdb.logger.Printf("Initializing mock database with 100 realistic merchants")

	// Realistic business data
	businesses := []struct {
		name          string
		legalName     string
		industry      string
		industryCode  string
		businessType  string
		employeeCount int
		annualRevenue float64
		address       models.Address
		contact       models.ContactInfo
	}{
		// Technology Companies
		{"TechFlow Solutions", "TechFlow Solutions Inc.", "Technology", "541511", "Corporation", 45, 2500000,
			models.Address{Street1: "123 Innovation Drive", City: "San Francisco", State: "CA", PostalCode: "94105", Country: "USA"},
			models.ContactInfo{Phone: "+1-415-555-0101", Email: "info@techflow.com", Website: "https://techflow.com"}},

		{"DataSync Analytics", "DataSync Analytics LLC", "Technology", "541512", "LLC", 28, 1800000,
			models.Address{Street1: "456 Data Street", City: "Austin", State: "TX", PostalCode: "78701", Country: "USA"},
			models.ContactInfo{Phone: "+1-512-555-0102", Email: "contact@datasync.com", Website: "https://datasync.com"}},

		{"CloudScale Systems", "CloudScale Systems Inc.", "Technology", "541511", "Corporation", 67, 4200000,
			models.Address{Street1: "789 Cloud Avenue", City: "Seattle", State: "WA", PostalCode: "98101", Country: "USA"},
			models.ContactInfo{Phone: "+1-206-555-0103", Email: "hello@cloudscale.com", Website: "https://cloudscale.com"}},

		// Financial Services
		{"Metro Credit Union", "Metro Credit Union", "Finance", "522110", "Credit Union", 125, 8500000,
			models.Address{Street1: "321 Financial Plaza", City: "Chicago", State: "IL", PostalCode: "60601", Country: "USA"},
			models.ContactInfo{Phone: "+1-312-555-0104", Email: "info@metrocu.org", Website: "https://metrocu.org"}},

		{"Premier Investment Group", "Premier Investment Group LLC", "Finance", "523920", "LLC", 89, 12000000,
			models.Address{Street1: "654 Investment Way", City: "New York", State: "NY", PostalCode: "10001", Country: "USA"},
			models.ContactInfo{Phone: "+1-212-555-0105", Email: "contact@premierinvest.com", Website: "https://premierinvest.com"}},

		// Healthcare
		{"Wellness Medical Center", "Wellness Medical Center PLLC", "Healthcare", "621111", "PLLC", 156, 9800000,
			models.Address{Street1: "987 Health Boulevard", City: "Denver", State: "CO", PostalCode: "80201", Country: "USA"},
			models.ContactInfo{Phone: "+1-303-555-0106", Email: "info@wellnessmed.com", Website: "https://wellnessmed.com"}},

		{"Advanced Dental Care", "Advanced Dental Care PC", "Healthcare", "621210", "PC", 34, 2100000,
			models.Address{Street1: "147 Dental Drive", City: "Phoenix", State: "AZ", PostalCode: "85001", Country: "USA"},
			models.ContactInfo{Phone: "+1-602-555-0107", Email: "appointments@advanceddental.com", Website: "https://advanceddental.com"}},

		// Retail
		{"Urban Fashion Co.", "Urban Fashion Company Inc.", "Retail", "448140", "Corporation", 78, 5600000,
			models.Address{Street1: "258 Fashion Street", City: "Los Angeles", State: "CA", PostalCode: "90001", Country: "USA"},
			models.ContactInfo{Phone: "+1-213-555-0108", Email: "orders@urbanfashion.com", Website: "https://urbanfashion.com"}},

		{"Green Earth Organics", "Green Earth Organics LLC", "Retail", "445299", "LLC", 42, 3200000,
			models.Address{Street1: "369 Organic Lane", City: "Portland", State: "OR", PostalCode: "97201", Country: "USA"},
			models.ContactInfo{Phone: "+1-503-555-0109", Email: "info@greenearth.com", Website: "https://greenearth.com"}},

		// Manufacturing
		{"Precision Manufacturing", "Precision Manufacturing Corp", "Manufacturing", "332710", "Corporation", 234, 18500000,
			models.Address{Street1: "741 Industrial Park", City: "Detroit", State: "MI", PostalCode: "48201", Country: "USA"},
			models.ContactInfo{Phone: "+1-313-555-0110", Email: "sales@precisionmfg.com", Website: "https://precisionmfg.com"}},

		{"EcoTech Solutions", "EcoTech Solutions Inc.", "Manufacturing", "334419", "Corporation", 167, 12800000,
			models.Address{Street1: "852 Green Tech Way", City: "Boston", State: "MA", PostalCode: "02101", Country: "USA"},
			models.ContactInfo{Phone: "+1-617-555-0111", Email: "contact@ecotech.com", Website: "https://ecotech.com"}},

		// Professional Services
		{"Legal Associates Group", "Legal Associates Group LLP", "Professional Services", "541110", "LLP", 89, 7500000,
			models.Address{Street1: "963 Legal Plaza", City: "Miami", State: "FL", PostalCode: "33101", Country: "USA"},
			models.ContactInfo{Phone: "+1-305-555-0112", Email: "info@legalassociates.com", Website: "https://legalassociates.com"}},

		{"Strategic Consulting", "Strategic Consulting LLC", "Professional Services", "541612", "LLC", 56, 4200000,
			models.Address{Street1: "174 Strategy Street", City: "Atlanta", State: "GA", PostalCode: "30301", Country: "USA"},
			models.ContactInfo{Phone: "+1-404-555-0113", Email: "hello@strategicconsulting.com", Website: "https://strategicconsulting.com"}},

		// Construction
		{"Metro Builders", "Metro Builders Inc.", "Construction", "236220", "Corporation", 145, 11200000,
			models.Address{Street1: "285 Construction Way", City: "Houston", State: "TX", PostalCode: "77001", Country: "USA"},
			models.ContactInfo{Phone: "+1-713-555-0114", Email: "projects@metrobuilders.com", Website: "https://metrobuilders.com"}},

		{"Elite Renovations", "Elite Renovations LLC", "Construction", "238220", "LLC", 67, 4800000,
			models.Address{Street1: "396 Renovation Road", City: "Nashville", State: "TN", PostalCode: "37201", Country: "USA"},
			models.ContactInfo{Phone: "+1-615-555-0115", Email: "info@eliterenovations.com", Website: "https://eliterenovations.com"}},

		// Food & Beverage
		{"Artisan Bakery Co.", "Artisan Bakery Company LLC", "Food & Beverage", "311812", "LLC", 23, 1800000,
			models.Address{Street1: "507 Bakery Lane", City: "Portland", State: "OR", PostalCode: "97201", Country: "USA"},
			models.ContactInfo{Phone: "+1-503-555-0116", Email: "orders@artisanbakery.com", Website: "https://artisanbakery.com"}},

		{"Craft Brewery", "Craft Brewery Inc.", "Food & Beverage", "312120", "Corporation", 45, 3200000,
			models.Address{Street1: "618 Brewery Street", City: "Denver", State: "CO", PostalCode: "80201", Country: "USA"},
			models.ContactInfo{Phone: "+1-303-555-0117", Email: "info@craftbrewery.com", Website: "https://craftbrewery.com"}},

		// Transportation
		{"City Logistics", "City Logistics LLC", "Transportation", "484121", "LLC", 78, 6500000,
			models.Address{Street1: "729 Logistics Drive", City: "Kansas City", State: "MO", PostalCode: "64101", Country: "USA"},
			models.ContactInfo{Phone: "+1-816-555-0118", Email: "dispatch@citylogistics.com", Website: "https://citylogistics.com"}},

		{"Express Delivery", "Express Delivery Inc.", "Transportation", "492110", "Corporation", 134, 9800000,
			models.Address{Street1: "830 Delivery Way", City: "Memphis", State: "TN", PostalCode: "38101", Country: "USA"},
			models.ContactInfo{Phone: "+1-901-555-0119", Email: "support@expressdelivery.com", Website: "https://expressdelivery.com"}},
	}

	// Portfolio types and risk levels
	portfolioTypes := []models.PortfolioType{
		models.PortfolioTypeOnboarded,
		models.PortfolioTypeProspective,
		models.PortfolioTypePending,
		models.PortfolioTypeDeactivated,
	}

	riskLevels := []models.RiskLevel{
		models.RiskLevelLow,
		models.RiskLevelMedium,
		models.RiskLevelHigh,
	}

	// Generate 100 merchants
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 100; i++ {
		business := businesses[i%len(businesses)]

		// Create variations for each business
		merchant := &models.Merchant{
			ID:                 fmt.Sprintf("merchant_%03d", i+1),
			Name:               business.name,
			LegalName:          business.legalName,
			RegistrationNumber: fmt.Sprintf("REG%06d", 100000+i),
			TaxID:              fmt.Sprintf("TAX%09d", 100000000+i),
			Industry:           business.industry,
			IndustryCode:       business.industryCode,
			BusinessType:       business.businessType,
			FoundedDate:        generateRandomDate(),
			EmployeeCount:      business.employeeCount + rand.Intn(50) - 25, // Add variation
			AnnualRevenue:      &business.annualRevenue,
			Address:            business.address,
			ContactInfo:        business.contact,
			PortfolioType:      portfolioTypes[rand.Intn(len(portfolioTypes))],
			RiskLevel:          riskLevels[rand.Intn(len(riskLevels))],
			ComplianceStatus:   generateComplianceStatus(),
			Status:             "active",
			CreatedBy:          "system",
			CreatedAt:          time.Now().Add(-time.Duration(rand.Intn(365)) * 24 * time.Hour),
			UpdatedAt:          time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour),
		}

		// Ensure employee count is positive
		if merchant.EmployeeCount < 1 {
			merchant.EmployeeCount = 1
		}

		// Add some variation to revenue
		revenueVariation := 0.8 + rand.Float64()*0.4 // 80% to 120% of base
		*merchant.AnnualRevenue *= revenueVariation

		mdb.merchants[merchant.ID] = merchant
	}

	mdb.logger.Printf("Successfully initialized %d mock merchants", len(mdb.merchants))
}

// generateRandomDate generates a random date within the last 20 years
func generateRandomDate() *time.Time {
	now := time.Now()
	yearsAgo := rand.Intn(20)
	monthsAgo := rand.Intn(12)
	daysAgo := rand.Intn(30)

	date := now.AddDate(-yearsAgo, -monthsAgo, -daysAgo)
	return &date
}

// generateComplianceStatus generates a random compliance status
func generateComplianceStatus() string {
	statuses := []string{"compliant", "pending", "non_compliant", "under_review"}
	return statuses[rand.Intn(len(statuses))]
}

// =============================================================================
// Merchant CRUD Operations
// =============================================================================

// CreateMerchant creates a new merchant in the mock database
func (mdb *MockMerchantDatabase) CreateMerchant(ctx context.Context, merchant *models.Merchant) error {
	mdb.logger.Printf("Creating merchant: %s", merchant.ID)

	if _, exists := mdb.merchants[merchant.ID]; exists {
		return ErrDuplicateMerchant
	}

	// Set timestamps if not provided
	now := time.Now()
	if merchant.CreatedAt.IsZero() {
		merchant.CreatedAt = now
	}
	if merchant.UpdatedAt.IsZero() {
		merchant.UpdatedAt = now
	}

	mdb.merchants[merchant.ID] = merchant
	mdb.logger.Printf("Successfully created merchant: %s", merchant.ID)
	return nil
}

// GetMerchant retrieves a merchant by ID
func (mdb *MockMerchantDatabase) GetMerchant(ctx context.Context, merchantID string) (*models.Merchant, error) {
	mdb.logger.Printf("Retrieving merchant: %s", merchantID)

	merchant, exists := mdb.merchants[merchantID]
	if !exists {
		return nil, ErrMerchantNotFound
	}

	return merchant, nil
}

// UpdateMerchant updates an existing merchant
func (mdb *MockMerchantDatabase) UpdateMerchant(ctx context.Context, merchant *models.Merchant) error {
	mdb.logger.Printf("Updating merchant: %s", merchant.ID)

	if _, exists := mdb.merchants[merchant.ID]; !exists {
		return ErrMerchantNotFound
	}

	merchant.UpdatedAt = time.Now()
	mdb.merchants[merchant.ID] = merchant
	mdb.logger.Printf("Successfully updated merchant: %s", merchant.ID)
	return nil
}

// DeleteMerchant deletes a merchant from the mock database
func (mdb *MockMerchantDatabase) DeleteMerchant(ctx context.Context, merchantID string) error {
	mdb.logger.Printf("Deleting merchant: %s", merchantID)

	if _, exists := mdb.merchants[merchantID]; !exists {
		return ErrMerchantNotFound
	}

	delete(mdb.merchants, merchantID)
	mdb.logger.Printf("Successfully deleted merchant: %s", merchantID)
	return nil
}

// ListMerchants retrieves merchants with pagination
func (mdb *MockMerchantDatabase) ListMerchants(ctx context.Context, page, pageSize int) ([]*models.Merchant, error) {
	if page < 1 || pageSize < 1 || pageSize > 1000 {
		return nil, ErrInvalidPagination
	}

	offset := (page - 1) * pageSize
	mdb.logger.Printf("Listing merchants (page: %d, size: %d, offset: %d)", page, pageSize, offset)

	var merchants []*models.Merchant
	count := 0
	for _, merchant := range mdb.merchants {
		if count >= offset && len(merchants) < pageSize {
			merchants = append(merchants, merchant)
		}
		count++
	}

	mdb.logger.Printf("Retrieved %d merchants", len(merchants))
	return merchants, nil
}

// SearchMerchants searches merchants with filters
func (mdb *MockMerchantDatabase) SearchMerchants(ctx context.Context, filters *models.MerchantSearchFilters, page, pageSize int) ([]*models.Merchant, error) {
	if page < 1 || pageSize < 1 || pageSize > 1000 {
		return nil, ErrInvalidPagination
	}

	offset := (page - 1) * pageSize
	mdb.logger.Printf("Searching merchants with filters (page: %d, size: %d)", page, pageSize)

	var filteredMerchants []*models.Merchant
	for _, merchant := range mdb.merchants {
		if mdb.matchesFilters(merchant, filters) {
			filteredMerchants = append(filteredMerchants, merchant)
		}
	}

	// Apply pagination
	var merchants []*models.Merchant
	for i, merchant := range filteredMerchants {
		if i >= offset && len(merchants) < pageSize {
			merchants = append(merchants, merchant)
		}
	}

	mdb.logger.Printf("Found %d merchants matching filters", len(merchants))
	return merchants, nil
}

// CountMerchants counts total merchants matching filters
func (mdb *MockMerchantDatabase) CountMerchants(ctx context.Context, filters *models.MerchantSearchFilters) (int, error) {
	mdb.logger.Printf("Counting merchants with filters")

	count := 0
	for _, merchant := range mdb.merchants {
		if mdb.matchesFilters(merchant, filters) {
			count++
		}
	}

	mdb.logger.Printf("Found %d merchants matching filters", count)
	return count, nil
}

// matchesFilters checks if a merchant matches the given filters
func (mdb *MockMerchantDatabase) matchesFilters(merchant *models.Merchant, filters *models.MerchantSearchFilters) bool {
	if filters == nil {
		return true
	}

	if filters.PortfolioType != nil && merchant.PortfolioType != *filters.PortfolioType {
		return false
	}

	if filters.RiskLevel != nil && merchant.RiskLevel != *filters.RiskLevel {
		return false
	}

	if filters.Industry != "" && merchant.Industry != filters.Industry {
		return false
	}

	if filters.Status != "" && merchant.Status != filters.Status {
		return false
	}

	if filters.SearchQuery != "" {
		searchLower := strings.ToLower(filters.SearchQuery)
		if !strings.Contains(strings.ToLower(merchant.Name), searchLower) &&
			!strings.Contains(strings.ToLower(merchant.LegalName), searchLower) &&
			!strings.Contains(strings.ToLower(merchant.Industry), searchLower) {
			return false
		}
	}

	if filters.CreatedAfter != nil && merchant.CreatedAt.Before(*filters.CreatedAfter) {
		return false
	}

	if filters.CreatedBefore != nil && merchant.CreatedAt.After(*filters.CreatedBefore) {
		return false
	}

	if filters.EmployeeCountMin != nil && merchant.EmployeeCount < *filters.EmployeeCountMin {
		return false
	}

	if filters.EmployeeCountMax != nil && merchant.EmployeeCount > *filters.EmployeeCountMax {
		return false
	}

	if filters.RevenueMin != nil && merchant.AnnualRevenue != nil && *merchant.AnnualRevenue < *filters.RevenueMin {
		return false
	}

	if filters.RevenueMax != nil && merchant.AnnualRevenue != nil && *merchant.AnnualRevenue > *filters.RevenueMax {
		return false
	}

	return true
}

// GetMerchantsByPortfolioType retrieves merchants by portfolio type
func (mdb *MockMerchantDatabase) GetMerchantsByPortfolioType(ctx context.Context, portfolioType models.PortfolioType, page, pageSize int) ([]*models.Merchant, error) {
	filters := &models.MerchantSearchFilters{
		PortfolioType: &portfolioType,
	}
	return mdb.SearchMerchants(ctx, filters, page, pageSize)
}

// GetMerchantsByRiskLevel retrieves merchants by risk level
func (mdb *MockMerchantDatabase) GetMerchantsByRiskLevel(ctx context.Context, riskLevel models.RiskLevel, page, pageSize int) ([]*models.Merchant, error) {
	filters := &models.MerchantSearchFilters{
		RiskLevel: &riskLevel,
	}
	return mdb.SearchMerchants(ctx, filters, page, pageSize)
}

// =============================================================================
// Session Management
// =============================================================================

// CreateSession creates a new merchant session
func (mdb *MockMerchantDatabase) CreateSession(ctx context.Context, session *models.MerchantSession) error {
	mdb.logger.Printf("Creating session: %s", session.ID)

	if _, exists := mdb.sessions[session.ID]; exists {
		return ErrDuplicateSession
	}

	// Set timestamps if not provided
	now := time.Now()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	if session.UpdatedAt.IsZero() {
		session.UpdatedAt = now
	}

	mdb.sessions[session.ID] = session
	mdb.logger.Printf("Successfully created session: %s", session.ID)
	return nil
}

// GetActiveSessionByUserID retrieves the active session for a user
func (mdb *MockMerchantDatabase) GetActiveSessionByUserID(ctx context.Context, userID string) (*models.MerchantSession, error) {
	mdb.logger.Printf("Getting active session for user: %s", userID)

	for _, session := range mdb.sessions {
		if session.UserID == userID && session.IsActive {
			return session, nil
		}
	}

	return nil, ErrSessionNotFound
}

// UpdateSession updates a merchant session
func (mdb *MockMerchantDatabase) UpdateSession(ctx context.Context, session *models.MerchantSession) error {
	mdb.logger.Printf("Updating session: %s", session.ID)

	if _, exists := mdb.sessions[session.ID]; !exists {
		return ErrSessionNotFound
	}

	session.UpdatedAt = time.Now()
	mdb.sessions[session.ID] = session
	mdb.logger.Printf("Successfully updated session: %s", session.ID)
	return nil
}

// DeactivateSession deactivates a session
func (mdb *MockMerchantDatabase) DeactivateSession(ctx context.Context, sessionID string) error {
	mdb.logger.Printf("Deactivating session: %s", sessionID)

	session, exists := mdb.sessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}

	session.IsActive = false
	session.UpdatedAt = time.Now()
	mdb.sessions[sessionID] = session
	mdb.logger.Printf("Successfully deactivated session: %s", sessionID)
	return nil
}

// =============================================================================
// Audit Logging
// =============================================================================

// CreateAuditLog creates a new audit log entry
func (mdb *MockMerchantDatabase) CreateAuditLog(ctx context.Context, auditLog *models.AuditLog) error {
	mdb.logger.Printf("Creating audit log: %s", auditLog.ID)

	// Set timestamp if not provided
	if auditLog.CreatedAt.IsZero() {
		auditLog.CreatedAt = time.Now()
	}

	mdb.auditLogs = append(mdb.auditLogs, auditLog)
	mdb.logger.Printf("Successfully created audit log: %s", auditLog.ID)
	return nil
}

// GetAuditLogsByMerchantID retrieves audit logs for a merchant
func (mdb *MockMerchantDatabase) GetAuditLogsByMerchantID(ctx context.Context, merchantID string, page, pageSize int) ([]*models.AuditLog, error) {
	if page < 1 || pageSize < 1 || pageSize > 1000 {
		return nil, ErrInvalidPagination
	}

	offset := (page - 1) * pageSize
	mdb.logger.Printf("Getting audit logs for merchant: %s (page: %d, size: %d)", merchantID, page, pageSize)

	var filteredLogs []*models.AuditLog
	for _, log := range mdb.auditLogs {
		if log.MerchantID == merchantID {
			filteredLogs = append(filteredLogs, log)
		}
	}

	// Apply pagination
	var auditLogs []*models.AuditLog
	for i, log := range filteredLogs {
		if i >= offset && len(auditLogs) < pageSize {
			auditLogs = append(auditLogs, log)
		}
	}

	mdb.logger.Printf("Retrieved %d audit logs for merchant %s", len(auditLogs), merchantID)
	return auditLogs, nil
}

// =============================================================================
// Bulk Operations
// =============================================================================

// BulkUpdatePortfolioType updates portfolio type for multiple merchants
func (mdb *MockMerchantDatabase) BulkUpdatePortfolioType(ctx context.Context, merchantIDs []string, portfolioType models.PortfolioType, userID string) error {
	if len(merchantIDs) == 0 {
		return nil
	}

	mdb.logger.Printf("Bulk updating portfolio type for %d merchants to %s", len(merchantIDs), portfolioType)

	updated := 0
	for _, merchantID := range merchantIDs {
		if merchant, exists := mdb.merchants[merchantID]; exists {
			merchant.PortfolioType = portfolioType
			merchant.UpdatedAt = time.Now()
			mdb.merchants[merchantID] = merchant
			updated++
		}
	}

	mdb.logger.Printf("Successfully updated %d merchants", updated)
	return nil
}

// BulkUpdateRiskLevel updates risk level for multiple merchants
func (mdb *MockMerchantDatabase) BulkUpdateRiskLevel(ctx context.Context, merchantIDs []string, riskLevel models.RiskLevel, userID string) error {
	if len(merchantIDs) == 0 {
		return nil
	}

	mdb.logger.Printf("Bulk updating risk level for %d merchants to %s", len(merchantIDs), riskLevel)

	updated := 0
	for _, merchantID := range merchantIDs {
		if merchant, exists := mdb.merchants[merchantID]; exists {
			merchant.RiskLevel = riskLevel
			merchant.UpdatedAt = time.Now()
			mdb.merchants[merchantID] = merchant
			updated++
		}
	}

	mdb.logger.Printf("Successfully updated %d merchants", updated)
	return nil
}

// =============================================================================
// Utility Methods
// =============================================================================

// GetMerchantCount returns the total number of merchants
func (mdb *MockMerchantDatabase) GetMerchantCount() int {
	return len(mdb.merchants)
}

// GetSessionCount returns the total number of sessions
func (mdb *MockMerchantDatabase) GetSessionCount() int {
	return len(mdb.sessions)
}

// GetAuditLogCount returns the total number of audit logs
func (mdb *MockMerchantDatabase) GetAuditLogCount() int {
	return len(mdb.auditLogs)
}

// ClearAllData clears all data from the mock database
func (mdb *MockMerchantDatabase) ClearAllData() {
	mdb.merchants = make(map[string]*models.Merchant)
	mdb.sessions = make(map[string]*models.MerchantSession)
	mdb.auditLogs = make([]*models.AuditLog, 0)
	mdb.logger.Printf("Cleared all data from mock database")
}

// ResetToInitialData resets the database to initial mock data
func (mdb *MockMerchantDatabase) ResetToInitialData() {
	mdb.ClearAllData()
	mdb.initializeMockData()
	mdb.logger.Printf("Reset mock database to initial data")
}
