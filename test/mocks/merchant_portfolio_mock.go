package mocks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"kyb-platform/internal/database"
	"kyb-platform/internal/services"
)

// MockMerchantPortfolioService provides a mock implementation for E2E tests
type MockMerchantPortfolioService struct {
	merchants     map[string]*services.Merchant
	sessions      map[string]*services.MerchantSession
	bulkResults   map[string]*services.BulkOperationResult
	searchResults []*services.Merchant
	searchTotal   int
	errors        map[string]error
}

// NewMockMerchantPortfolioService creates a new mock merchant portfolio service
func NewMockMerchantPortfolioService() *MockMerchantPortfolioService {
	return &MockMerchantPortfolioService{
		merchants:   make(map[string]*services.Merchant),
		sessions:    make(map[string]*services.MerchantSession),
		bulkResults: make(map[string]*services.BulkOperationResult),
		errors:      make(map[string]error),
	}
}

// CreateMerchant creates a new merchant
func (m *MockMerchantPortfolioService) CreateMerchant(ctx context.Context, merchant *services.Merchant, userID string) (*services.Merchant, error) {
	if err, exists := m.errors["create"]; exists {
		return nil, err
	}

	// Generate ID if not provided
	if merchant.ID == "" {
		merchant.ID = fmt.Sprintf("merchant_%d", time.Now().UnixNano())
	}

	// Set timestamps
	now := time.Now()
	merchant.CreatedAt = now
	merchant.UpdatedAt = now
	merchant.CreatedBy = userID

	m.merchants[merchant.ID] = merchant
	return merchant, nil
}

// GetMerchant retrieves a merchant by ID
func (m *MockMerchantPortfolioService) GetMerchant(ctx context.Context, id string) (*services.Merchant, error) {
	if err, exists := m.errors["get"]; exists {
		return nil, err
	}

	if merchant, exists := m.merchants[id]; exists {
		return merchant, nil
	}
	return nil, database.ErrMerchantNotFound
}

// UpdateMerchant updates an existing merchant
func (m *MockMerchantPortfolioService) UpdateMerchant(ctx context.Context, merchant *services.Merchant, userID string) (*services.Merchant, error) {
	if err, exists := m.errors["update"]; exists {
		return nil, err
	}

	if _, exists := m.merchants[merchant.ID]; !exists {
		return nil, database.ErrMerchantNotFound
	}

	merchant.UpdatedAt = time.Now()
	m.merchants[merchant.ID] = merchant
	return merchant, nil
}

// DeleteMerchant deletes a merchant
func (m *MockMerchantPortfolioService) DeleteMerchant(ctx context.Context, merchantID string, userID string) error {
	if err, exists := m.errors["delete"]; exists {
		return err
	}

	if _, exists := m.merchants[merchantID]; !exists {
		return database.ErrMerchantNotFound
	}

	delete(m.merchants, merchantID)
	return nil
}

// SearchMerchants searches for merchants with filters
func (m *MockMerchantPortfolioService) SearchMerchants(ctx context.Context, filters *services.MerchantSearchFilters, page, pageSize int) (*services.MerchantListResult, error) {
	if err, exists := m.errors["search"]; exists {
		return nil, err
	}

	// Return mock search results
	return &services.MerchantListResult{
		Merchants: m.searchResults,
		Total:     m.searchTotal,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// BulkUpdatePortfolioType updates portfolio type for multiple merchants
func (m *MockMerchantPortfolioService) BulkUpdatePortfolioType(ctx context.Context, merchantIDs []string, portfolioType services.PortfolioType, userID string) (*services.BulkOperationResult, error) {
	if err, exists := m.errors["bulk_portfolio"]; exists {
		return nil, err
	}

	operationID := fmt.Sprintf("bulk_op_%d", time.Now().UnixNano())
	successful := 0
	failed := 0

	for _, merchantID := range merchantIDs {
		if merchant, exists := m.merchants[merchantID]; exists {
			merchant.PortfolioType = portfolioType
			merchant.UpdatedAt = time.Now()
			successful++
		} else {
			failed++
		}
	}

	result := &services.BulkOperationResult{
		OperationID:       operationID,
		TotalMerchants:    len(merchantIDs),
		SuccessfulUpdates: successful,
		FailedUpdates:     failed,
		Status:            "completed",
		CreatedAt:         time.Now(),
	}

	m.bulkResults[operationID] = result
	return result, nil
}

// BulkUpdateRiskLevel updates risk level for multiple merchants
func (m *MockMerchantPortfolioService) BulkUpdateRiskLevel(ctx context.Context, merchantIDs []string, riskLevel services.RiskLevel, userID string) (*services.BulkOperationResult, error) {
	if err, exists := m.errors["bulk_risk"]; exists {
		return nil, err
	}

	operationID := fmt.Sprintf("bulk_op_%d", time.Now().UnixNano())
	successful := 0
	failed := 0

	for _, merchantID := range merchantIDs {
		if merchant, exists := m.merchants[merchantID]; exists {
			merchant.RiskLevel = riskLevel
			merchant.UpdatedAt = time.Now()
			successful++
		} else {
			failed++
		}
	}

	result := &services.BulkOperationResult{
		OperationID:       operationID,
		TotalMerchants:    len(merchantIDs),
		SuccessfulUpdates: successful,
		FailedUpdates:     failed,
		Status:            "completed",
		CreatedAt:         time.Now(),
	}

	m.bulkResults[operationID] = result
	return result, nil
}

// StartMerchantSession starts a new merchant session
func (m *MockMerchantPortfolioService) StartMerchantSession(ctx context.Context, userID, merchantID string) (*services.MerchantSession, error) {
	if err, exists := m.errors["start_session"]; exists {
		return nil, err
	}

	// Check if merchant exists
	if _, exists := m.merchants[merchantID]; !exists {
		return nil, database.ErrMerchantNotFound
	}

	// End any existing active session for this user
	for _, session := range m.sessions {
		if session.UserID == userID && session.IsActive {
			session.IsActive = false
			session.EndedAt = &[]time.Time{time.Now()}[0]
		}
	}

	// Create new session
	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())
	session := &services.MerchantSession{
		ID:         sessionID,
		UserID:     userID,
		MerchantID: merchantID,
		StartedAt:  time.Now(),
		LastActive: time.Now(),
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	m.sessions[sessionID] = session
	return session, nil
}

// EndMerchantSession ends the active merchant session for a user
func (m *MockMerchantPortfolioService) EndMerchantSession(ctx context.Context, userID string) error {
	if err, exists := m.errors["end_session"]; exists {
		return err
	}

	for _, session := range m.sessions {
		if session.UserID == userID && session.IsActive {
			session.IsActive = false
			endedAt := time.Now()
			session.EndedAt = &endedAt
			session.UpdatedAt = time.Now()
			return nil
		}
	}

	return errors.New("no active session found")
}

// GetActiveMerchantSession gets the active merchant session for a user
func (m *MockMerchantPortfolioService) GetActiveMerchantSession(ctx context.Context, userID string) (*services.MerchantSession, error) {
	if err, exists := m.errors["get_session"]; exists {
		return nil, err
	}

	for _, session := range m.sessions {
		if session.UserID == userID && session.IsActive {
			return session, nil
		}
	}

	return nil, errors.New("no active session found")
}

// Helper methods for testing

// AddMerchant adds a merchant to the mock service
func (m *MockMerchantPortfolioService) AddMerchant(merchant *services.Merchant) {
	m.merchants[merchant.ID] = merchant
}

// SetSearchResults sets the search results for testing
func (m *MockMerchantPortfolioService) SetSearchResults(merchants []*services.Merchant, total int) {
	m.searchResults = merchants
	m.searchTotal = total
}

// SetError sets an error for a specific operation
func (m *MockMerchantPortfolioService) SetError(operation string, err error) {
	m.errors[operation] = err
}

// ClearErrors clears all errors
func (m *MockMerchantPortfolioService) ClearErrors() {
	m.errors = make(map[string]error)
}

// GetMerchantCount returns the number of merchants
func (m *MockMerchantPortfolioService) GetMerchantCount() int {
	return len(m.merchants)
}

// GetSessionCount returns the number of sessions
func (m *MockMerchantPortfolioService) GetSessionCount() int {
	return len(m.sessions)
}
