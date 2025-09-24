package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"kyb-platform/internal/database"
	"kyb-platform/internal/models"
	"kyb-platform/internal/services"
)

// MockMerchantPortfolioService is a mock implementation of MerchantPortfolioService
type MockMerchantPortfolioService struct {
	merchants     map[string]*services.Merchant
	sessions      map[string]*services.MerchantSession
	bulkResults   map[string]*services.BulkOperationResult
	searchResults []*services.Merchant
	searchTotal   int
	errors        map[string]error
}

func NewMockMerchantPortfolioService() *MockMerchantPortfolioService {
	return &MockMerchantPortfolioService{
		merchants:   make(map[string]*services.Merchant),
		sessions:    make(map[string]*services.MerchantSession),
		bulkResults: make(map[string]*services.BulkOperationResult),
		errors:      make(map[string]error),
	}
}

func (m *MockMerchantPortfolioService) CreateMerchant(ctx context.Context, merchant *services.Merchant, userID string) (*services.Merchant, error) {
	if err, exists := m.errors["create"]; exists {
		return nil, err
	}
	m.merchants[merchant.ID] = merchant
	return merchant, nil
}

func (m *MockMerchantPortfolioService) GetMerchant(ctx context.Context, id string) (*services.Merchant, error) {
	if err, exists := m.errors["get"]; exists {
		return nil, err
	}
	if merchant, exists := m.merchants[id]; exists {
		return merchant, nil
	}
	return nil, database.ErrMerchantNotFound
}

func (m *MockMerchantPortfolioService) UpdateMerchant(ctx context.Context, merchant *services.Merchant, userID string) (*services.Merchant, error) {
	if err, exists := m.errors["update"]; exists {
		return nil, err
	}
	if _, exists := m.merchants[merchant.ID]; !exists {
		return nil, database.ErrMerchantNotFound
	}
	m.merchants[merchant.ID] = merchant
	return merchant, nil
}

func (m *MockMerchantPortfolioService) DeleteMerchant(ctx context.Context, id string, userID string) error {
	if err, exists := m.errors["delete"]; exists {
		return err
	}
	if _, exists := m.merchants[id]; !exists {
		return database.ErrMerchantNotFound
	}
	delete(m.merchants, id)
	return nil
}

func (m *MockMerchantPortfolioService) SearchMerchants(ctx context.Context, filters *services.MerchantSearchFilters, page, pageSize int) (*services.MerchantListResult, error) {
	if err, exists := m.errors["search"]; exists {
		return nil, err
	}
	return &services.MerchantListResult{
		Merchants: m.searchResults,
		Total:     m.searchTotal,
		Page:      page,
		PageSize:  pageSize,
		HasMore:   page*pageSize < m.searchTotal,
	}, nil
}

func (m *MockMerchantPortfolioService) BulkUpdatePortfolioType(ctx context.Context, merchantIDs []string, portfolioType services.PortfolioType, userID string) (*services.BulkOperationResult, error) {
	if err, exists := m.errors["bulk"]; exists {
		return nil, err
	}
	operationID := fmt.Sprintf("bulk_%d", time.Now().UnixNano())
	result := &services.BulkOperationResult{
		OperationID: operationID,
		TotalItems:  len(merchantIDs),
		Processed:   len(merchantIDs),
		Successful:  len(merchantIDs),
		Failed:      0,
		Status:      "completed",
		StartedAt:   time.Now(),
	}
	m.bulkResults[operationID] = result
	return result, nil
}

func (m *MockMerchantPortfolioService) BulkUpdateRiskLevel(ctx context.Context, merchantIDs []string, riskLevel services.RiskLevel, userID string) (*services.BulkOperationResult, error) {
	if err, exists := m.errors["bulk"]; exists {
		return nil, err
	}
	operationID := fmt.Sprintf("bulk_%d", time.Now().UnixNano())
	result := &services.BulkOperationResult{
		OperationID: operationID,
		TotalItems:  len(merchantIDs),
		Processed:   len(merchantIDs),
		Successful:  len(merchantIDs),
		Failed:      0,
		Status:      "completed",
		StartedAt:   time.Now(),
	}
	m.bulkResults[operationID] = result
	return result, nil
}

func (m *MockMerchantPortfolioService) StartMerchantSession(ctx context.Context, userID, merchantID string) (*services.MerchantSession, error) {
	if err, exists := m.errors["start_session"]; exists {
		return nil, err
	}
	if _, exists := m.merchants[merchantID]; !exists {
		return nil, database.ErrMerchantNotFound
	}
	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())
	session := &services.MerchantSession{
		ID:         sessionID,
		MerchantID: merchantID,
		UserID:     userID,
		StartedAt:  time.Now(),
		LastActive: time.Now(),
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	m.sessions[sessionID] = session
	return session, nil
}

func (m *MockMerchantPortfolioService) EndMerchantSession(ctx context.Context, userID string) error {
	if err, exists := m.errors["end_session"]; exists {
		return err
	}
	// Find and remove session
	for sessionID, session := range m.sessions {
		if session.UserID == userID && session.IsActive {
			delete(m.sessions, sessionID)
			return nil
		}
	}
	return database.ErrSessionNotFound
}

func (m *MockMerchantPortfolioService) GetActiveMerchantSession(ctx context.Context, userID string) (*services.MerchantSession, error) {
	if err, exists := m.errors["get_session"]; exists {
		return nil, err
	}
	for _, session := range m.sessions {
		if session.UserID == userID && session.IsActive {
			return session, nil
		}
	}
	return nil, database.ErrSessionNotFound
}

// Helper methods for test setup
func (m *MockMerchantPortfolioService) SetError(operation string, err error) {
	m.errors[operation] = err
}

func (m *MockMerchantPortfolioService) SetSearchResults(merchants []*services.Merchant, total int) {
	m.searchResults = merchants
	m.searchTotal = total
}

func (m *MockMerchantPortfolioService) AddMerchant(merchant *services.Merchant) {
	m.merchants[merchant.ID] = merchant
}

// Test helper functions
func createTestMerchant() *services.Merchant {
	return &services.Merchant{
		ID:                 "test_merchant_1",
		Name:               "Test Company",
		LegalName:          "Test Company LLC",
		RegistrationNumber: "REG123456",
		TaxID:              "TAX123456",
		Industry:           "Technology",
		IndustryCode:       "7372",
		BusinessType:       "LLC",
		FoundedDate:        &[]time.Time{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)}[0],
		EmployeeCount:      50,
		AnnualRevenue:      &[]float64{1000000.0}[0],
		Address: database.Address{
			Street1:     "123 Test St",
			City:        "Test City",
			State:       "TS",
			PostalCode:  "12345",
			Country:     "United States",
			CountryCode: "US",
		},
		ContactInfo: database.ContactInfo{
			Phone:          "+1-555-123-4567",
			Email:          "test@testcompany.com",
			Website:        "https://testcompany.com",
			PrimaryContact: "John Doe",
		},
		PortfolioType:    services.PortfolioTypeOnboarded,
		RiskLevel:        services.RiskLevelMedium,
		ComplianceStatus: "compliant",
		Status:           "active",
		CreatedBy:        "test_user",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

func createTestRequest(method, url string, body interface{}) *http.Request {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "user_id", "test_user")
	return req.WithContext(ctx)
}

// =============================================================================
// Test Cases
// =============================================================================

func TestCreateMerchant(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	tests := []struct {
		name           string
		request        CreateMerchantRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful creation",
			request: CreateMerchantRequest{
				Name:      "Test Company",
				LegalName: "Test Company LLC",
				Address: models.Address{
					Street1: "123 Test St",
					City:    "Test City",
					State:   "TS",
				},
				ContactInfo: models.ContactInfo{
					Email: "test@testcompany.com",
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing required fields",
			request: CreateMerchantRequest{
				Name: "",
				Address: models.Address{
					Street1: "123 Test St",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Name and legal name are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest("POST", "/api/v1/merchants", tt.request)
			w := httptest.NewRecorder()

			handler.CreateMerchant(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				if !strings.Contains(w.Body.String(), tt.expectedError) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedError, w.Body.String())
				}
			}

			if tt.expectedStatus == http.StatusCreated {
				var response MerchantResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if response.Name != tt.request.Name {
					t.Errorf("Expected name %s, got %s", tt.request.Name, response.Name)
				}
			}
		})
	}
}

func TestGetMerchant(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	// Add test merchant
	testMerchant := createTestMerchant()
	mockService.AddMerchant(testMerchant)

	tests := []struct {
		name           string
		merchantID     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful retrieval",
			merchantID:     "test_merchant_1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "merchant not found",
			merchantID:     "nonexistent_merchant",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Merchant not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest("GET", fmt.Sprintf("/api/v1/merchants/%s", tt.merchantID), nil)
			w := httptest.NewRecorder()

			handler.GetMerchant(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				if !strings.Contains(w.Body.String(), tt.expectedError) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedError, w.Body.String())
				}
			}

			if tt.expectedStatus == http.StatusOK {
				var response MerchantResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if response.ID != tt.merchantID {
					t.Errorf("Expected ID %s, got %s", tt.merchantID, response.ID)
				}
			}
		})
	}
}

func TestUpdateMerchant(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	// Add test merchant
	testMerchant := createTestMerchant()
	mockService.AddMerchant(testMerchant)

	tests := []struct {
		name           string
		merchantID     string
		request        UpdateMerchantRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name:       "successful update",
			merchantID: "test_merchant_1",
			request: UpdateMerchantRequest{
				Name: &[]string{"Updated Company"}[0],
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "merchant not found",
			merchantID:     "nonexistent_merchant",
			request:        UpdateMerchantRequest{},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Merchant not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest("PUT", fmt.Sprintf("/api/v1/merchants/%s", tt.merchantID), tt.request)
			w := httptest.NewRecorder()

			handler.UpdateMerchant(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				if !strings.Contains(w.Body.String(), tt.expectedError) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedError, w.Body.String())
				}
			}

			if tt.expectedStatus == http.StatusOK && tt.request.Name != nil {
				var response MerchantResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if response.Name != *tt.request.Name {
					t.Errorf("Expected name %s, got %s", *tt.request.Name, response.Name)
				}
			}
		})
	}
}

func TestDeleteMerchant(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	// Add test merchant
	testMerchant := createTestMerchant()
	mockService.AddMerchant(testMerchant)

	tests := []struct {
		name           string
		merchantID     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful deletion",
			merchantID:     "test_merchant_1",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "merchant not found",
			merchantID:     "nonexistent_merchant",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Merchant not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest("DELETE", fmt.Sprintf("/api/v1/merchants/%s", tt.merchantID), nil)
			w := httptest.NewRecorder()

			handler.DeleteMerchant(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				if !strings.Contains(w.Body.String(), tt.expectedError) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedError, w.Body.String())
				}
			}
		})
	}
}

func TestListMerchants(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	// Set up search results
	testMerchants := []*services.Merchant{createTestMerchant()}
	mockService.SetSearchResults(testMerchants, 1)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "successful listing",
			queryParams:    "?page=1&page_size=20",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "listing with filters",
			queryParams:    "?portfolio_type=onboarded&risk_level=medium",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest("GET", "/api/v1/merchants"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.ListMerchants(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response MerchantListResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if len(response.Merchants) != 1 {
					t.Errorf("Expected 1 merchant, got %d", len(response.Merchants))
				}
				if response.Total != 1 {
					t.Errorf("Expected total 1, got %d", response.Total)
				}
			}
		})
	}
}

func TestSearchMerchants(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	// Set up search results
	testMerchants := []*services.Merchant{createTestMerchant()}
	mockService.SetSearchResults(testMerchants, 1)

	tests := []struct {
		name           string
		request        MerchantSearchRequest
		expectedStatus int
	}{
		{
			name: "successful search",
			request: MerchantSearchRequest{
				Query:    "Test Company",
				Page:     1,
				PageSize: 20,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "search with filters",
			request: MerchantSearchRequest{
				Query:         "Test",
				PortfolioType: &[]services.PortfolioType{services.PortfolioTypeOnboarded}[0],
				RiskLevel:     &[]services.RiskLevel{services.RiskLevelMedium}[0],
				Page:          1,
				PageSize:      20,
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest("POST", "/api/v1/merchants/search", tt.request)
			w := httptest.NewRecorder()

			handler.SearchMerchants(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response MerchantListResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if len(response.Merchants) != 1 {
					t.Errorf("Expected 1 merchant, got %d", len(response.Merchants))
				}
			}
		})
	}
}

func TestBulkUpdateMerchants(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	tests := []struct {
		name           string
		request        BulkOperationRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful bulk update",
			request: BulkOperationRequest{
				MerchantIDs:   []string{"merchant_1", "merchant_2"},
				Operation:     "update_portfolio_type",
				PortfolioType: &[]services.PortfolioType{services.PortfolioTypeOnboarded}[0],
				Reason:        "Bulk onboarding",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing merchant IDs",
			request: BulkOperationRequest{
				MerchantIDs: []string{},
				Operation:   "update_portfolio_type",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Merchant IDs are required",
		},
		{
			name: "missing operation",
			request: BulkOperationRequest{
				MerchantIDs: []string{"merchant_1"},
				Operation:   "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Operation is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest("POST", "/api/v1/merchants/bulk/update", tt.request)
			w := httptest.NewRecorder()

			handler.BulkUpdateMerchants(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				if !strings.Contains(w.Body.String(), tt.expectedError) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedError, w.Body.String())
				}
			}

			if tt.expectedStatus == http.StatusOK {
				var response BulkOperationResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if response.TotalMerchants != len(tt.request.MerchantIDs) {
					t.Errorf("Expected total merchants %d, got %d", len(tt.request.MerchantIDs), response.TotalMerchants)
				}
			}
		})
	}
}

func TestStartMerchantSession(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	// Add test merchant
	testMerchant := createTestMerchant()
	mockService.AddMerchant(testMerchant)

	tests := []struct {
		name           string
		merchantID     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful session start",
			merchantID:     "test_merchant_1",
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "merchant not found",
			merchantID:     "nonexistent_merchant",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Merchant not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest("POST", fmt.Sprintf("/api/v1/merchants/%s/session", tt.merchantID), nil)
			w := httptest.NewRecorder()

			handler.StartMerchantSession(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				if !strings.Contains(w.Body.String(), tt.expectedError) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedError, w.Body.String())
				}
			}

			if tt.expectedStatus == http.StatusCreated {
				var response SessionResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if response.MerchantID != tt.merchantID {
					t.Errorf("Expected merchant ID %s, got %s", tt.merchantID, response.MerchantID)
				}
				if !response.IsActive {
					t.Errorf("Expected session to be active")
				}
			}
		})
	}
}

func TestEndMerchantSession(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	// Add test merchant and session
	testMerchant := createTestMerchant()
	mockService.AddMerchant(testMerchant)
	mockService.StartMerchantSession(context.Background(), "test_user", testMerchant.ID)

	tests := []struct {
		name           string
		merchantID     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful session end",
			merchantID:     "test_merchant_1",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "session not found",
			merchantID:     "nonexistent_merchant",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Session not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest("DELETE", fmt.Sprintf("/api/v1/merchants/%s/session", tt.merchantID), nil)
			w := httptest.NewRecorder()

			handler.EndMerchantSession(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				if !strings.Contains(w.Body.String(), tt.expectedError) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedError, w.Body.String())
				}
			}
		})
	}
}

func TestGetActiveSession(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	// Add test merchant and session
	testMerchant := createTestMerchant()
	mockService.AddMerchant(testMerchant)
	session, _ := mockService.StartMerchantSession(context.Background(), "test_user", testMerchant.ID)

	tests := []struct {
		name           string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful session retrieval",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest("GET", "/api/v1/merchants/session/active", nil)
			w := httptest.NewRecorder()

			handler.GetActiveSession(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				if !strings.Contains(w.Body.String(), tt.expectedError) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedError, w.Body.String())
				}
			}

			if tt.expectedStatus == http.StatusOK {
				var response SessionResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if response.SessionID != session.ID {
					t.Errorf("Expected session ID %s, got %s", session.ID, response.SessionID)
				}
				if !response.IsActive {
					t.Errorf("Expected session to be active")
				}
			}
		})
	}
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestExtractMerchantIDFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/api/v1/merchants/merchant_123", "merchant_123"},
		{"/merchants/merchant_456", "merchant_456"},
		{"/api/v1/merchants/merchant_789/session", "merchant_789"},
		{"/api/v1/merchants/", ""},
		{"/api/v1/merchants", ""},
		{"/invalid/path", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := extractMerchantIDFromPath(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestConvertMerchantToResponse(t *testing.T) {
	handler := NewMerchantPortfolioHandler(nil, log.Default())

	merchant := createTestMerchant()
	response := handler.convertMerchantToResponse(merchant)

	if response.ID != merchant.ID {
		t.Errorf("Expected ID %s, got %s", merchant.ID, response.ID)
	}
	if response.Name != merchant.Name {
		t.Errorf("Expected Name %s, got %s", merchant.Name, response.Name)
	}
	if response.PortfolioType != models.PortfolioType(merchant.PortfolioType) {
		t.Errorf("Expected PortfolioType %s, got %s", merchant.PortfolioType, response.PortfolioType)
	}
	if response.RiskLevel != models.RiskLevel(merchant.RiskLevel) {
		t.Errorf("Expected RiskLevel %s, got %s", merchant.RiskLevel, response.RiskLevel)
	}
}

// =============================================================================
// Additional Comprehensive API Handler Tests
// =============================================================================

func TestMerchantPortfolioHandler_ErrorHandling(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	// Test service error handling
	mockService.SetError("create", errors.New("database connection failed"))

	req := createTestRequest("POST", "/api/v1/merchants", CreateMerchantRequest{
		Name:      "Test Company",
		LegalName: "Test Company LLC",
		Address: models.Address{
			Street1: "123 Test St",
			City:    "Test City",
			State:   "TS",
		},
		ContactInfo: models.ContactInfo{
			Email: "test@testcompany.com",
		},
	})
	w := httptest.NewRecorder()

	handler.CreateMerchant(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestMerchantPortfolioHandler_RequestValidation(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	tests := []struct {
		name           string
		request        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "invalid JSON",
			request:        "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name: "missing required fields",
			request: CreateMerchantRequest{
				Name: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Name and legal name are required",
		},
		{
			name:           "empty request body",
			request:        nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.request == nil {
				req = createTestRequest("POST", "/api/v1/merchants", nil)
			} else if tt.request == "invalid json" {
				req = httptest.NewRequest("POST", "/api/v1/merchants", strings.NewReader("invalid json"))
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), "user_id", "test_user")
				req = req.WithContext(ctx)
			} else {
				req = createTestRequest("POST", "/api/v1/merchants", tt.request)
			}

			w := httptest.NewRecorder()
			handler.CreateMerchant(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				if !strings.Contains(w.Body.String(), tt.expectedError) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedError, w.Body.String())
				}
			}
		})
	}
}

func TestMerchantPortfolioHandler_Pagination(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	// Set up search results
	testMerchants := []*services.Merchant{createTestMerchant()}
	mockService.SetSearchResults(testMerchants, 1)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedPage   int
		expectedSize   int
	}{
		{
			name:           "default pagination",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedPage:   1,
			expectedSize:   20,
		},
		{
			name:           "custom pagination",
			queryParams:    "?page=2&page_size=10",
			expectedStatus: http.StatusOK,
			expectedPage:   2,
			expectedSize:   10,
		},
		{
			name:           "invalid page",
			queryParams:    "?page=0&page_size=10",
			expectedStatus: http.StatusOK,
			expectedPage:   1, // Should default to 1
			expectedSize:   10,
		},
		{
			name:           "invalid page size",
			queryParams:    "?page=1&page_size=0",
			expectedStatus: http.StatusOK,
			expectedPage:   1,
			expectedSize:   20, // Should default to 20
		},
		{
			name:           "page size too large",
			queryParams:    "?page=1&page_size=1000",
			expectedStatus: http.StatusOK,
			expectedPage:   1,
			expectedSize:   100, // Should cap at 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest("GET", "/api/v1/merchants"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.ListMerchants(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response MerchantListResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if response.Page != tt.expectedPage {
					t.Errorf("Expected page %d, got %d", tt.expectedPage, response.Page)
				}
				if response.PageSize != tt.expectedSize {
					t.Errorf("Expected page size %d, got %d", tt.expectedSize, response.PageSize)
				}
			}
		})
	}
}

func TestMerchantPortfolioHandler_SearchFilters(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	// Set up search results
	testMerchants := []*services.Merchant{createTestMerchant()}
	mockService.SetSearchResults(testMerchants, 1)

	tests := []struct {
		name           string
		request        MerchantSearchRequest
		expectedStatus int
	}{
		{
			name: "search with query only",
			request: MerchantSearchRequest{
				Query: "Test Company",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "search with portfolio type filter",
			request: MerchantSearchRequest{
				Query:         "Test",
				PortfolioType: &[]services.PortfolioType{services.PortfolioTypeOnboarded}[0],
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "search with risk level filter",
			request: MerchantSearchRequest{
				Query:     "Test",
				RiskLevel: &[]services.RiskLevel{services.RiskLevelMedium}[0],
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "search with industry filter",
			request: MerchantSearchRequest{
				Query:    "Test",
				Industry: "Technology",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "search with status filter",
			request: MerchantSearchRequest{
				Query:  "Test",
				Status: "active",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "search with all filters",
			request: MerchantSearchRequest{
				Query:         "Test",
				PortfolioType: &[]services.PortfolioType{services.PortfolioTypeOnboarded}[0],
				RiskLevel:     &[]services.RiskLevel{services.RiskLevelMedium}[0],
				Industry:      "Technology",
				Status:        "active",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest("POST", "/api/v1/merchants/search", tt.request)
			w := httptest.NewRecorder()

			handler.SearchMerchants(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response MerchantListResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if len(response.Merchants) != 1 {
					t.Errorf("Expected 1 merchant, got %d", len(response.Merchants))
				}
			}
		})
	}
}

func TestMerchantPortfolioHandler_BulkOperations_Validation(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	tests := []struct {
		name           string
		request        BulkOperationRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid bulk update",
			request: BulkOperationRequest{
				MerchantIDs:   []string{"merchant_1", "merchant_2"},
				Operation:     "update_portfolio_type",
				PortfolioType: &[]services.PortfolioType{services.PortfolioTypeOnboarded}[0],
				Reason:        "Bulk onboarding",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "valid bulk risk level update",
			request: BulkOperationRequest{
				MerchantIDs: []string{"merchant_1", "merchant_2"},
				Operation:   "update_risk_level",
				RiskLevel:   &[]services.RiskLevel{services.RiskLevelHigh}[0],
				Reason:      "Risk assessment update",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "empty merchant IDs",
			request: BulkOperationRequest{
				MerchantIDs: []string{},
				Operation:   "update_portfolio_type",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Merchant IDs are required",
		},
		{
			name: "missing operation",
			request: BulkOperationRequest{
				MerchantIDs: []string{"merchant_1"},
				Operation:   "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Operation is required",
		},
		{
			name: "invalid operation",
			request: BulkOperationRequest{
				MerchantIDs: []string{"merchant_1"},
				Operation:   "invalid_operation",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid operation",
		},
		{
			name: "missing portfolio type for portfolio update",
			request: BulkOperationRequest{
				MerchantIDs: []string{"merchant_1"},
				Operation:   "update_portfolio_type",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Portfolio type is required",
		},
		{
			name: "missing risk level for risk update",
			request: BulkOperationRequest{
				MerchantIDs: []string{"merchant_1"},
				Operation:   "update_risk_level",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Risk level is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest("POST", "/api/v1/merchants/bulk/update", tt.request)
			w := httptest.NewRecorder()

			handler.BulkUpdateMerchants(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				if !strings.Contains(w.Body.String(), tt.expectedError) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedError, w.Body.String())
				}
			}

			if tt.expectedStatus == http.StatusOK {
				var response BulkOperationResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if response.TotalMerchants != len(tt.request.MerchantIDs) {
					t.Errorf("Expected total merchants %d, got %d", len(tt.request.MerchantIDs), response.TotalMerchants)
				}
			}
		})
	}
}

func TestMerchantPortfolioHandler_SessionManagement_EdgeCases(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	// Test starting session with service error
	mockService.SetError("start_session", errors.New("session creation failed"))

	req := createTestRequest("POST", "/api/v1/merchants/test_merchant_1/session", nil)
	w := httptest.NewRecorder()

	handler.StartMerchantSession(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	// Test ending session with service error
	mockService.SetError("end_session", errors.New("session termination failed"))

	req = createTestRequest("DELETE", "/api/v1/merchants/test_merchant_1/session", nil)
	w = httptest.NewRecorder()

	handler.EndMerchantSession(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	// Test getting active session with service error
	mockService.SetError("get_session", errors.New("session retrieval failed"))

	req = createTestRequest("GET", "/api/v1/merchants/session/active", nil)
	w = httptest.NewRecorder()

	handler.GetActiveSession(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestMerchantPortfolioHandler_ResponseFormatting(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	// Test merchant response formatting
	testMerchant := createTestMerchant()
	mockService.AddMerchant(testMerchant)

	req := createTestRequest("GET", "/api/v1/merchants/test_merchant_1", nil)
	w := httptest.NewRecorder()

	handler.GetMerchant(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response MerchantResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	// Verify all fields are properly formatted
	if response.ID != testMerchant.ID {
		t.Errorf("Expected ID %s, got %s", testMerchant.ID, response.ID)
	}
	if response.Name != testMerchant.Name {
		t.Errorf("Expected Name %s, got %s", testMerchant.Name, response.Name)
	}
	if response.LegalName != testMerchant.LegalName {
		t.Errorf("Expected LegalName %s, got %s", testMerchant.LegalName, response.LegalName)
	}
	if response.Industry != testMerchant.Industry {
		t.Errorf("Expected Industry %s, got %s", testMerchant.Industry, response.Industry)
	}
	if response.PortfolioType != models.PortfolioType(testMerchant.PortfolioType) {
		t.Errorf("Expected PortfolioType %s, got %s", testMerchant.PortfolioType, response.PortfolioType)
	}
	if response.RiskLevel != models.RiskLevel(testMerchant.RiskLevel) {
		t.Errorf("Expected RiskLevel %s, got %s", testMerchant.RiskLevel, response.RiskLevel)
	}
}

func TestMerchantPortfolioHandler_ContextHandling(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	// Test request without user context
	req := httptest.NewRequest("GET", "/api/v1/merchants/test_merchant_1", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.GetMerchant(w, req)

	// Should still work but with empty user ID
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestMerchantPortfolioHandler_ContentTypeHandling(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	// Test request with wrong content type
	req := httptest.NewRequest("POST", "/api/v1/merchants", strings.NewReader(`{"name":"Test"}`))
	req.Header.Set("Content-Type", "text/plain")
	ctx := context.WithValue(req.Context(), "user_id", "test_user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.CreateMerchant(w, req)

	// Should still work as we don't strictly enforce content type
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestMerchantPortfolioHandler_ConcurrentRequests(t *testing.T) {
	mockService := NewMockMerchantPortfolioService()
	handler := NewMerchantPortfolioHandler(mockService, log.Default())

	// Test concurrent requests
	const numGoroutines = 10
	results := make(chan int, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			req := createTestRequest("GET", fmt.Sprintf("/api/v1/merchants/merchant_%d", id), nil)
			w := httptest.NewRecorder()
			handler.GetMerchant(w, req)
			results <- w.Code
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		select {
		case code := <-results:
			if code != http.StatusNotFound {
				t.Errorf("Expected status %d, got %d", http.StatusNotFound, code)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent requests timed out")
		}
	}
}
