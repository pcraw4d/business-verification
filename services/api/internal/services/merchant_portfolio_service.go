package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"kyb-platform/internal/database"
)

// MerchantPortfolioService provides core business logic for merchant portfolio management
type MerchantPortfolioService struct {
	db     database.Database
	logger *log.Logger
	mu     sync.RWMutex // For session management
}

// NewMerchantPortfolioService creates a new merchant portfolio service
func NewMerchantPortfolioService(db database.Database, logger *log.Logger) *MerchantPortfolioService {
	if logger == nil {
		logger = log.Default()
	}

	return &MerchantPortfolioService{
		db:     db,
		logger: logger,
	}
}

// Merchant represents a merchant in the portfolio
type Merchant struct {
	ID                 string               `json:"id" db:"id"`
	Name               string               `json:"name" db:"name"`
	LegalName          string               `json:"legal_name" db:"legal_name"`
	RegistrationNumber string               `json:"registration_number" db:"registration_number"`
	TaxID              string               `json:"tax_id" db:"tax_id"`
	Industry           string               `json:"industry" db:"industry"`
	IndustryCode       string               `json:"industry_code" db:"industry_code"`
	BusinessType       string               `json:"business_type" db:"business_type"`
	FoundedDate        *time.Time           `json:"founded_date" db:"founded_date"`
	EmployeeCount      int                  `json:"employee_count" db:"employee_count"`
	AnnualRevenue      *float64             `json:"annual_revenue" db:"annual_revenue"`
	Address            database.Address     `json:"address" db:"address"`
	ContactInfo        database.ContactInfo `json:"contact_info" db:"contact_info"`
	PortfolioType      PortfolioType        `json:"portfolio_type" db:"portfolio_type"`
	RiskLevel          RiskLevel            `json:"risk_level" db:"risk_level"`
	ComplianceStatus   string               `json:"compliance_status" db:"compliance_status"`
	Status             string               `json:"status" db:"status"`
	CreatedBy          string               `json:"created_by" db:"created_by"`
	CreatedAt          time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at" db:"updated_at"`
}

// PortfolioType represents the type of merchant in the portfolio
type PortfolioType string

const (
	PortfolioTypeOnboarded   PortfolioType = "onboarded"
	PortfolioTypeDeactivated PortfolioType = "deactivated"
	PortfolioTypeProspective PortfolioType = "prospective"
	PortfolioTypePending     PortfolioType = "pending"
)

// RiskLevel represents the risk level of a merchant
type RiskLevel string

const (
	RiskLevelHigh   RiskLevel = "high"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelLow    RiskLevel = "low"
)

// MerchantSession represents an active merchant session
type MerchantSession struct {
	ID         string    `json:"id" db:"id"`
	UserID     string    `json:"user_id" db:"user_id"`
	MerchantID string    `json:"merchant_id" db:"merchant_id"`
	StartedAt  time.Time `json:"started_at" db:"started_at"`
	LastActive time.Time `json:"last_active" db:"last_active"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// AuditLog represents an audit log entry for merchant operations
type AuditLog struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	MerchantID   string    `json:"merchant_id" db:"merchant_id"`
	Action       string    `json:"action" db:"action"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
	ResourceID   string    `json:"resource_id" db:"resource_id"`
	Details      string    `json:"details" db:"details"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	RequestID    string    `json:"request_id" db:"request_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// MerchantSearchFilters represents filters for merchant search
type MerchantSearchFilters struct {
	PortfolioType *PortfolioType `json:"portfolio_type,omitempty"`
	RiskLevel     *RiskLevel     `json:"risk_level,omitempty"`
	Industry      string         `json:"industry,omitempty"`
	Status        string         `json:"status,omitempty"`
	SearchQuery   string         `json:"search_query,omitempty"`
}

// MerchantListResult represents the result of a merchant list operation
type MerchantListResult struct {
	Merchants []*Merchant `json:"merchants"`
	Total     int         `json:"total"`
	Page      int         `json:"page"`
	PageSize  int         `json:"page_size"`
	HasMore   bool        `json:"has_more"`
}

// BulkOperationResult represents the result of a bulk operation
type BulkOperationResult struct {
	OperationID string                 `json:"operation_id"`
	Status      string                 `json:"status"`
	TotalItems  int                    `json:"total_items"`
	Processed   int                    `json:"processed"`
	Successful  int                    `json:"successful"`
	Failed      int                    `json:"failed"`
	Errors      []string               `json:"errors"`
	Results     []BulkOperationItem    `json:"results"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// BulkOperationItem represents a single item in a bulk operation
type BulkOperationItem struct {
	MerchantID string `json:"merchant_id"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
}

// Common errors
var (
	ErrInvalidPortfolioType  = errors.New("invalid portfolio type")
	ErrInvalidRiskLevel      = errors.New("invalid risk level")
	ErrSessionAlreadyActive  = errors.New("user already has an active merchant session")
	ErrNoActiveSession       = errors.New("no active merchant session found")
	ErrBulkOperationNotFound = errors.New("bulk operation not found")
)

// =============================================================================
// Merchant CRUD Operations
// =============================================================================

// CreateMerchant creates a new merchant in the portfolio
func (s *MerchantPortfolioService) CreateMerchant(ctx context.Context, merchant *Merchant, userID string) (*Merchant, error) {
	// Validate merchant data first
	if err := s.validateMerchant(merchant); err != nil {
		return nil, fmt.Errorf("merchant validation failed: %w", err)
	}

	s.logger.Printf("Creating merchant: %s", merchant.Name)

	// Set default values
	merchant.ID = s.generateID()
	merchant.CreatedBy = userID
	merchant.CreatedAt = time.Now()
	merchant.UpdatedAt = time.Now()

	// Set default portfolio type if not specified
	if merchant.PortfolioType == "" {
		merchant.PortfolioType = PortfolioTypeProspective
	}

	// Set default risk level if not specified
	if merchant.RiskLevel == "" {
		merchant.RiskLevel = RiskLevelMedium
	}

	// Set default status if not specified
	if merchant.Status == "" {
		merchant.Status = "active"
	}

	// Convert to database Business model
	business := s.merchantToBusiness(merchant)

	// Create in database
	if err := s.db.CreateBusiness(ctx, business); err != nil {
		return nil, fmt.Errorf("failed to create merchant in database: %w", err)
	}

	// Log audit trail
	if err := s.logAuditEvent(ctx, userID, merchant.ID, "CREATE_MERCHANT", "merchant", merchant.ID, "Merchant created", "", ""); err != nil {
		s.logger.Printf("Warning: failed to log audit event: %v", err)
	}

	s.logger.Printf("Successfully created merchant: %s (ID: %s)", merchant.Name, merchant.ID)
	return merchant, nil
}

// GetMerchant retrieves a merchant by ID
func (s *MerchantPortfolioService) GetMerchant(ctx context.Context, merchantID string) (*Merchant, error) {
	s.logger.Printf("Retrieving merchant: %s", merchantID)

	business, err := s.db.GetBusinessByID(ctx, merchantID)
	if err != nil {
		if errors.Is(err, database.ErrUserNotFound) {
			return nil, database.ErrMerchantNotFound
		}
		return nil, fmt.Errorf("failed to retrieve merchant from database: %w", err)
	}

	merchant := s.businessToMerchant(business)
	return merchant, nil
}

// UpdateMerchant updates an existing merchant
func (s *MerchantPortfolioService) UpdateMerchant(ctx context.Context, merchant *Merchant, userID string) (*Merchant, error) {
	s.logger.Printf("Updating merchant: %s", merchant.ID)

	// Validate merchant data
	if err := s.validateMerchant(merchant); err != nil {
		return nil, fmt.Errorf("merchant validation failed: %w", err)
	}

	// Check if merchant exists
	existing, err := s.GetMerchant(ctx, merchant.ID)
	if err != nil {
		return nil, err
	}

	// Update timestamp
	merchant.UpdatedAt = time.Now()
	merchant.CreatedBy = existing.CreatedBy // Preserve original creator
	merchant.CreatedAt = existing.CreatedAt // Preserve original creation time

	// Convert to database Business model
	business := s.merchantToBusiness(merchant)

	// Update in database
	if err := s.db.UpdateBusiness(ctx, business); err != nil {
		return nil, fmt.Errorf("failed to update merchant in database: %w", err)
	}

	// Log audit trail
	if err := s.logAuditEvent(ctx, userID, merchant.ID, "UPDATE_MERCHANT", "merchant", merchant.ID, "Merchant updated", "", ""); err != nil {
		s.logger.Printf("Warning: failed to log audit event: %v", err)
	}

	s.logger.Printf("Successfully updated merchant: %s", merchant.ID)
	return merchant, nil
}

// DeleteMerchant deletes a merchant from the portfolio
func (s *MerchantPortfolioService) DeleteMerchant(ctx context.Context, merchantID string, userID string) error {
	s.logger.Printf("Deleting merchant: %s", merchantID)

	// Check if merchant exists
	_, err := s.GetMerchant(ctx, merchantID)
	if err != nil {
		return err
	}

	// Delete from database
	if err := s.db.DeleteBusiness(ctx, merchantID); err != nil {
		return fmt.Errorf("failed to delete merchant from database: %w", err)
	}

	// Log audit trail
	if err := s.logAuditEvent(ctx, userID, merchantID, "DELETE_MERCHANT", "merchant", merchantID, "Merchant deleted", "", ""); err != nil {
		s.logger.Printf("Warning: failed to log audit event: %v", err)
	}

	s.logger.Printf("Successfully deleted merchant: %s", merchantID)
	return nil
}

// =============================================================================
// Portfolio Type Management
// =============================================================================

// UpdateMerchantPortfolioType updates the portfolio type of a merchant
func (s *MerchantPortfolioService) UpdateMerchantPortfolioType(ctx context.Context, merchantID string, portfolioType PortfolioType, userID string) error {
	s.logger.Printf("Updating portfolio type for merchant %s to %s", merchantID, portfolioType)

	// Validate portfolio type
	if !s.isValidPortfolioType(portfolioType) {
		return ErrInvalidPortfolioType
	}

	// Get existing merchant
	merchant, err := s.GetMerchant(ctx, merchantID)
	if err != nil {
		return err
	}

	// Update portfolio type
	merchant.PortfolioType = portfolioType
	merchant.UpdatedAt = time.Now()

	// Save changes
	_, err = s.UpdateMerchant(ctx, merchant, userID)
	if err != nil {
		return fmt.Errorf("failed to update merchant portfolio type: %w", err)
	}

	// Log audit trail
	if err := s.logAuditEvent(ctx, userID, merchantID, "UPDATE_PORTFOLIO_TYPE", "merchant", merchantID,
		fmt.Sprintf("Portfolio type updated to %s", portfolioType), "", ""); err != nil {
		s.logger.Printf("Warning: failed to log audit event: %v", err)
	}

	s.logger.Printf("Successfully updated portfolio type for merchant %s", merchantID)
	return nil
}

// GetMerchantsByPortfolioType retrieves merchants by portfolio type
func (s *MerchantPortfolioService) GetMerchantsByPortfolioType(ctx context.Context, portfolioType PortfolioType, page, pageSize int) (*MerchantListResult, error) {
	s.logger.Printf("Retrieving merchants by portfolio type: %s (page: %d, size: %d)", portfolioType, page, pageSize)

	// Validate portfolio type
	if !s.isValidPortfolioType(portfolioType) {
		return nil, ErrInvalidPortfolioType
	}

	// For now, we'll use the existing business search and filter by portfolio type
	// In a real implementation, this would be optimized with proper database queries
	businesses, err := s.db.ListBusinesses(ctx, pageSize*10, 0) // Get more to filter
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve businesses: %w", err)
	}

	// Filter by portfolio type (this is a simplified implementation)
	var filteredMerchants []*Merchant
	for _, business := range businesses {
		merchant := s.businessToMerchant(business)
		if merchant.PortfolioType == portfolioType {
			filteredMerchants = append(filteredMerchants, merchant)
		}
	}

	// Apply pagination
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(filteredMerchants) {
		return &MerchantListResult{
			Merchants: []*Merchant{},
			Total:     len(filteredMerchants),
			Page:      page,
			PageSize:  pageSize,
			HasMore:   false,
		}, nil
	}

	if end > len(filteredMerchants) {
		end = len(filteredMerchants)
	}

	result := &MerchantListResult{
		Merchants: filteredMerchants[start:end],
		Total:     len(filteredMerchants),
		Page:      page,
		PageSize:  pageSize,
		HasMore:   end < len(filteredMerchants),
	}

	s.logger.Printf("Retrieved %d merchants by portfolio type %s", len(result.Merchants), portfolioType)
	return result, nil
}

// =============================================================================
// Risk Level Assignment
// =============================================================================

// UpdateMerchantRiskLevel updates the risk level of a merchant
func (s *MerchantPortfolioService) UpdateMerchantRiskLevel(ctx context.Context, merchantID string, riskLevel RiskLevel, userID string) error {
	s.logger.Printf("Updating risk level for merchant %s to %s", merchantID, riskLevel)

	// Validate risk level
	if !s.isValidRiskLevel(riskLevel) {
		return ErrInvalidRiskLevel
	}

	// Get existing merchant
	merchant, err := s.GetMerchant(ctx, merchantID)
	if err != nil {
		return err
	}

	// Update risk level
	merchant.RiskLevel = riskLevel
	merchant.UpdatedAt = time.Now()

	// Save changes
	_, err = s.UpdateMerchant(ctx, merchant, userID)
	if err != nil {
		return fmt.Errorf("failed to update merchant risk level: %w", err)
	}

	// Log audit trail
	if err := s.logAuditEvent(ctx, userID, merchantID, "UPDATE_RISK_LEVEL", "merchant", merchantID,
		fmt.Sprintf("Risk level updated to %s", riskLevel), "", ""); err != nil {
		s.logger.Printf("Warning: failed to log audit event: %v", err)
	}

	s.logger.Printf("Successfully updated risk level for merchant %s", merchantID)
	return nil
}

// GetMerchantsByRiskLevel retrieves merchants by risk level
func (s *MerchantPortfolioService) GetMerchantsByRiskLevel(ctx context.Context, riskLevel RiskLevel, page, pageSize int) (*MerchantListResult, error) {
	s.logger.Printf("Retrieving merchants by risk level: %s (page: %d, size: %d)", riskLevel, page, pageSize)

	// Validate risk level
	if !s.isValidRiskLevel(riskLevel) {
		return nil, ErrInvalidRiskLevel
	}

	// For now, we'll use the existing business search and filter by risk level
	businesses, err := s.db.ListBusinesses(ctx, pageSize*10, 0) // Get more to filter
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve businesses: %w", err)
	}

	// Filter by risk level (this is a simplified implementation)
	var filteredMerchants []*Merchant
	for _, business := range businesses {
		merchant := s.businessToMerchant(business)
		if merchant.RiskLevel == riskLevel {
			filteredMerchants = append(filteredMerchants, merchant)
		}
	}

	// Apply pagination
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(filteredMerchants) {
		return &MerchantListResult{
			Merchants: []*Merchant{},
			Total:     len(filteredMerchants),
			Page:      page,
			PageSize:  pageSize,
			HasMore:   false,
		}, nil
	}

	if end > len(filteredMerchants) {
		end = len(filteredMerchants)
	}

	result := &MerchantListResult{
		Merchants: filteredMerchants[start:end],
		Total:     len(filteredMerchants),
		Page:      page,
		PageSize:  pageSize,
		HasMore:   end < len(filteredMerchants),
	}

	s.logger.Printf("Retrieved %d merchants by risk level %s", len(result.Merchants), riskLevel)
	return result, nil
}

// =============================================================================
// Session Management (Single Merchant Active at a Time)
// =============================================================================

// StartMerchantSession starts a new merchant session for a user
func (s *MerchantPortfolioService) StartMerchantSession(ctx context.Context, userID, merchantID string) (*MerchantSession, error) {
	s.logger.Printf("Starting merchant session for user %s with merchant %s", userID, merchantID)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user already has an active session
	existingSession, err := s.getActiveSessionByUserID(ctx, userID)
	if err != nil && !errors.Is(err, ErrNoActiveSession) {
		return nil, fmt.Errorf("failed to check existing session: %w", err)
	}

	if existingSession != nil {
		// Deactivate existing session
		if err := s.deactivateSession(ctx, existingSession.ID); err != nil {
			s.logger.Printf("Warning: failed to deactivate existing session: %v", err)
		}
	}

	// Verify merchant exists
	_, err = s.GetMerchant(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("merchant not found: %w", err)
	}

	// Create new session
	session := &MerchantSession{
		ID:         s.generateID(),
		UserID:     userID,
		MerchantID: merchantID,
		StartedAt:  time.Now(),
		LastActive: time.Now(),
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save session (for now, we'll use a simple in-memory approach)
	// In a real implementation, this would be stored in the database
	if err := s.saveSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Log audit trail
	if err := s.logAuditEvent(ctx, userID, merchantID, "START_SESSION", "session", session.ID,
		"Merchant session started", "", ""); err != nil {
		s.logger.Printf("Warning: failed to log audit event: %v", err)
	}

	s.logger.Printf("Successfully started merchant session: %s", session.ID)
	return session, nil
}

// GetActiveMerchantSession retrieves the active merchant session for a user
func (s *MerchantPortfolioService) GetActiveMerchantSession(ctx context.Context, userID string) (*MerchantSession, error) {
	s.logger.Printf("Retrieving active merchant session for user: %s", userID)

	s.mu.RLock()
	defer s.mu.RUnlock()

	session, err := s.getActiveSessionByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update last active timestamp
	session.LastActive = time.Now()
	if err := s.saveSession(ctx, session); err != nil {
		s.logger.Printf("Warning: failed to update session last active time: %v", err)
	}

	return session, nil
}

// EndMerchantSession ends the active merchant session for a user
func (s *MerchantPortfolioService) EndMerchantSession(ctx context.Context, userID string) error {
	s.logger.Printf("Ending merchant session for user: %s", userID)

	s.mu.Lock()
	defer s.mu.Unlock()

	session, err := s.getActiveSessionByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Deactivate session
	if err := s.deactivateSession(ctx, session.ID); err != nil {
		return fmt.Errorf("failed to deactivate session: %w", err)
	}

	// Log audit trail
	if err := s.logAuditEvent(ctx, userID, session.MerchantID, "END_SESSION", "session", session.ID,
		"Merchant session ended", "", ""); err != nil {
		s.logger.Printf("Warning: failed to log audit event: %v", err)
	}

	s.logger.Printf("Successfully ended merchant session: %s", session.ID)
	return nil
}

// =============================================================================
// Search and Filtering
// =============================================================================

// SearchMerchants searches merchants with filters and pagination
func (s *MerchantPortfolioService) SearchMerchants(ctx context.Context, filters *MerchantSearchFilters, page, pageSize int) (*MerchantListResult, error) {
	s.logger.Printf("Searching merchants with filters (page: %d, size: %d)", page, pageSize)

	// Get all businesses (simplified implementation)
	businesses, err := s.db.ListBusinesses(ctx, pageSize*10, 0) // Get more to filter
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve businesses: %w", err)
	}

	// Convert to merchants
	var allMerchants []*Merchant
	for _, business := range businesses {
		merchant := s.businessToMerchant(business)
		allMerchants = append(allMerchants, merchant)
	}

	// Apply filters
	filteredMerchants := s.applyFilters(allMerchants, filters)

	// Apply pagination
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(filteredMerchants) {
		return &MerchantListResult{
			Merchants: []*Merchant{},
			Total:     len(filteredMerchants),
			Page:      page,
			PageSize:  pageSize,
			HasMore:   false,
		}, nil
	}

	if end > len(filteredMerchants) {
		end = len(filteredMerchants)
	}

	result := &MerchantListResult{
		Merchants: filteredMerchants[start:end],
		Total:     len(filteredMerchants),
		Page:      page,
		PageSize:  pageSize,
		HasMore:   end < len(filteredMerchants),
	}

	s.logger.Printf("Search completed: %d merchants found", len(result.Merchants))
	return result, nil
}

// =============================================================================
// Bulk Operations
// =============================================================================

// BulkUpdatePortfolioType updates portfolio type for multiple merchants
func (s *MerchantPortfolioService) BulkUpdatePortfolioType(ctx context.Context, merchantIDs []string, portfolioType PortfolioType, userID string) (*BulkOperationResult, error) {
	s.logger.Printf("Bulk updating portfolio type for %d merchants to %s", len(merchantIDs), portfolioType)

	// Validate portfolio type
	if !s.isValidPortfolioType(portfolioType) {
		return nil, ErrInvalidPortfolioType
	}

	operationID := s.generateID()
	result := &BulkOperationResult{
		OperationID: operationID,
		Status:      "processing",
		TotalItems:  len(merchantIDs),
		Processed:   0,
		Successful:  0,
		Failed:      0,
		Errors:      []string{},
		Results:     []BulkOperationItem{},
		StartedAt:   time.Now(),
		Metadata:    map[string]interface{}{"operation": "bulk_update_portfolio_type", "portfolio_type": string(portfolioType)},
	}

	// Process each merchant
	for _, merchantID := range merchantIDs {
		item := BulkOperationItem{
			MerchantID: merchantID,
			Status:     "pending",
		}

		err := s.UpdateMerchantPortfolioType(ctx, merchantID, portfolioType, userID)
		if err != nil {
			item.Status = "failed"
			item.Error = err.Error()
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("Merchant %s: %v", merchantID, err))
		} else {
			item.Status = "success"
			result.Successful++
		}

		result.Results = append(result.Results, item)
		result.Processed++
	}

	// Mark operation as completed
	result.Status = "completed"
	now := time.Now()
	result.CompletedAt = &now

	s.logger.Printf("Bulk operation completed: %d successful, %d failed", result.Successful, result.Failed)
	return result, nil
}

// BulkUpdateRiskLevel updates risk level for multiple merchants
func (s *MerchantPortfolioService) BulkUpdateRiskLevel(ctx context.Context, merchantIDs []string, riskLevel RiskLevel, userID string) (*BulkOperationResult, error) {
	s.logger.Printf("Bulk updating risk level for %d merchants to %s", len(merchantIDs), riskLevel)

	// Validate risk level
	if !s.isValidRiskLevel(riskLevel) {
		return nil, ErrInvalidRiskLevel
	}

	operationID := s.generateID()
	result := &BulkOperationResult{
		OperationID: operationID,
		Status:      "processing",
		TotalItems:  len(merchantIDs),
		Processed:   0,
		Successful:  0,
		Failed:      0,
		Errors:      []string{},
		Results:     []BulkOperationItem{},
		StartedAt:   time.Now(),
		Metadata:    map[string]interface{}{"operation": "bulk_update_risk_level", "risk_level": string(riskLevel)},
	}

	// Process each merchant
	for _, merchantID := range merchantIDs {
		item := BulkOperationItem{
			MerchantID: merchantID,
			Status:     "pending",
		}

		err := s.UpdateMerchantRiskLevel(ctx, merchantID, riskLevel, userID)
		if err != nil {
			item.Status = "failed"
			item.Error = err.Error()
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("Merchant %s: %v", merchantID, err))
		} else {
			item.Status = "success"
			result.Successful++
		}

		result.Results = append(result.Results, item)
		result.Processed++
	}

	// Mark operation as completed
	result.Status = "completed"
	now := time.Now()
	result.CompletedAt = &now

	s.logger.Printf("Bulk operation completed: %d successful, %d failed", result.Successful, result.Failed)
	return result, nil
}

// =============================================================================
// Helper Methods
// =============================================================================

// validateMerchant validates merchant data
func (s *MerchantPortfolioService) validateMerchant(merchant *Merchant) error {
	if merchant == nil {
		return errors.New("merchant cannot be nil")
	}

	if strings.TrimSpace(merchant.Name) == "" {
		return errors.New("merchant name is required")
	}

	if merchant.PortfolioType != "" && !s.isValidPortfolioType(merchant.PortfolioType) {
		return ErrInvalidPortfolioType
	}

	if merchant.RiskLevel != "" && !s.isValidRiskLevel(merchant.RiskLevel) {
		return ErrInvalidRiskLevel
	}

	return nil
}

// isValidPortfolioType checks if a portfolio type is valid
func (s *MerchantPortfolioService) isValidPortfolioType(portfolioType PortfolioType) bool {
	switch portfolioType {
	case PortfolioTypeOnboarded, PortfolioTypeDeactivated, PortfolioTypeProspective, PortfolioTypePending:
		return true
	default:
		return false
	}
}

// isValidRiskLevel checks if a risk level is valid
func (s *MerchantPortfolioService) isValidRiskLevel(riskLevel RiskLevel) bool {
	switch riskLevel {
	case RiskLevelHigh, RiskLevelMedium, RiskLevelLow:
		return true
	default:
		return false
	}
}

// applyFilters applies search filters to merchant list
func (s *MerchantPortfolioService) applyFilters(merchants []*Merchant, filters *MerchantSearchFilters) []*Merchant {
	if filters == nil {
		return merchants
	}

	var filtered []*Merchant
	for _, merchant := range merchants {
		// Portfolio type filter
		if filters.PortfolioType != nil && merchant.PortfolioType != *filters.PortfolioType {
			continue
		}

		// Risk level filter
		if filters.RiskLevel != nil && merchant.RiskLevel != *filters.RiskLevel {
			continue
		}

		// Industry filter
		if filters.Industry != "" && !strings.Contains(strings.ToLower(merchant.Industry), strings.ToLower(filters.Industry)) {
			continue
		}

		// Status filter
		if filters.Status != "" && merchant.Status != filters.Status {
			continue
		}

		// Search query filter
		if filters.SearchQuery != "" {
			searchLower := strings.ToLower(filters.SearchQuery)
			if !strings.Contains(strings.ToLower(merchant.Name), searchLower) &&
				!strings.Contains(strings.ToLower(merchant.LegalName), searchLower) &&
				!strings.Contains(strings.ToLower(merchant.Industry), searchLower) {
				continue
			}
		}

		filtered = append(filtered, merchant)
	}

	return filtered
}

// merchantToBusiness converts a Merchant to a database Business
func (s *MerchantPortfolioService) merchantToBusiness(merchant *Merchant) *database.Business {
	// Encode portfolio type in status field for storage
	status := string(merchant.PortfolioType)
	if status == "" {
		status = "prospective"
	}

	return &database.Business{
		ID:                 merchant.ID,
		Name:               merchant.Name,
		LegalName:          merchant.LegalName,
		RegistrationNumber: merchant.RegistrationNumber,
		TaxID:              merchant.TaxID,
		Industry:           merchant.Industry,
		IndustryCode:       merchant.IndustryCode,
		BusinessType:       merchant.BusinessType,
		FoundedDate:        merchant.FoundedDate,
		EmployeeCount:      merchant.EmployeeCount,
		AnnualRevenue:      merchant.AnnualRevenue,
		Address:            merchant.Address,
		ContactInfo:        merchant.ContactInfo,
		Status:             status,
		RiskLevel:          string(merchant.RiskLevel),
		ComplianceStatus:   merchant.ComplianceStatus,
		CreatedBy:          merchant.CreatedBy,
		CreatedAt:          merchant.CreatedAt,
		UpdatedAt:          merchant.UpdatedAt,
	}
}

// businessToMerchant converts a database Business to a Merchant
func (s *MerchantPortfolioService) businessToMerchant(business *database.Business) *Merchant {
	// For MVP, we'll use a simple mapping approach
	// In a real implementation, we'd have a separate portfolio_type field
	portfolioType := PortfolioTypeProspective
	status := business.Status

	// Map status to portfolio type for MVP
	switch status {
	case "onboarded":
		portfolioType = PortfolioTypeOnboarded
	case "deactivated":
		portfolioType = PortfolioTypeDeactivated
	case "pending":
		portfolioType = PortfolioTypePending
	default:
		portfolioType = PortfolioTypeProspective
	}

	// Parse risk level
	riskLevel := RiskLevelMedium
	switch strings.ToLower(business.RiskLevel) {
	case "high":
		riskLevel = RiskLevelHigh
	case "medium":
		riskLevel = RiskLevelMedium
	case "low":
		riskLevel = RiskLevelLow
	}

	return &Merchant{
		ID:                 business.ID,
		Name:               business.Name,
		LegalName:          business.LegalName,
		RegistrationNumber: business.RegistrationNumber,
		TaxID:              business.TaxID,
		Industry:           business.Industry,
		IndustryCode:       business.IndustryCode,
		BusinessType:       business.BusinessType,
		FoundedDate:        business.FoundedDate,
		EmployeeCount:      business.EmployeeCount,
		AnnualRevenue:      business.AnnualRevenue,
		Address:            business.Address,
		ContactInfo:        business.ContactInfo,
		PortfolioType:      portfolioType,
		RiskLevel:          riskLevel,
		ComplianceStatus:   business.ComplianceStatus,
		Status:             business.Status,
		CreatedBy:          business.CreatedBy,
		CreatedAt:          business.CreatedAt,
		UpdatedAt:          business.UpdatedAt,
	}
}

// generateID generates a unique ID
func (s *MerchantPortfolioService) generateID() string {
	return fmt.Sprintf("merchant_%d", time.Now().UnixNano())
}

// logAuditEvent logs an audit event
func (s *MerchantPortfolioService) logAuditEvent(ctx context.Context, userID, merchantID, action, resourceType, resourceID, details, ipAddress, userAgent string) error {
	auditLog := &database.AuditLog{
		ID:           s.generateID(),
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		RequestID:    s.generateID(),
		CreatedAt:    time.Now(),
	}

	return s.db.CreateAuditLog(ctx, auditLog)
}

// Session management helpers (simplified in-memory implementation)
var activeSessions = make(map[string]*MerchantSession)
var sessionMutex sync.RWMutex

// saveSession saves a session (simplified in-memory implementation)
func (s *MerchantPortfolioService) saveSession(ctx context.Context, session *MerchantSession) error {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	activeSessions[session.ID] = session
	return nil
}

// getActiveSessionByUserID gets the active session for a user
func (s *MerchantPortfolioService) getActiveSessionByUserID(ctx context.Context, userID string) (*MerchantSession, error) {
	sessionMutex.RLock()
	defer sessionMutex.RUnlock()

	for _, session := range activeSessions {
		if session.UserID == userID && session.IsActive {
			return session, nil
		}
	}

	return nil, ErrNoActiveSession
}

// deactivateSession deactivates a session
func (s *MerchantPortfolioService) deactivateSession(ctx context.Context, sessionID string) error {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	if session, exists := activeSessions[sessionID]; exists {
		session.IsActive = false
		session.UpdatedAt = time.Now()
		activeSessions[sessionID] = session
	}

	return nil
}
