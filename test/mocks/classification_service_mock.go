package mocks

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/shared"
)

// MockClassificationService provides a mock implementation of classification service for E2E tests
type MockClassificationService struct {
	// Configuration for mock behavior
	ShouldFail   bool
	Delay        time.Duration
	MockResults  []shared.IndustryClassification
	ErrorMessage string
}

// NewMockClassificationService creates a new mock classification service
func NewMockClassificationService() *MockClassificationService {
	return &MockClassificationService{
		ShouldFail:   false,
		Delay:        100 * time.Millisecond,
		MockResults:  getDefaultMockResults(),
		ErrorMessage: "mock classification error",
	}
}

// ClassifyBusiness implements the classification service interface for E2E tests
func (m *MockClassificationService) ClassifyBusiness(ctx context.Context, request *shared.BusinessClassificationRequest) (*shared.BusinessClassificationResponse, error) {
	// Simulate processing delay
	if m.Delay > 0 {
		time.Sleep(m.Delay)
	}

	// Simulate failure if configured
	if m.ShouldFail {
		return nil, &ClassificationError{
			Message: m.ErrorMessage,
			Code:    "MOCK_CLASSIFICATION_ERROR",
		}
	}

	// Create mock response
	response := &shared.BusinessClassificationResponse{
		ID:                   request.ID,
		BusinessName:         request.BusinessName,
		Classifications:      m.MockResults,
		OverallConfidence:    0.85,
		ClassificationMethod: "mock_classification",
		ProcessingTime:       m.Delay,
		ModuleResults:        make(map[string]shared.ModuleResult),
		RawData:              make(map[string]interface{}),
		CreatedAt:            time.Now(),
		Metadata:             make(map[string]interface{}),
	}

	// Set primary classification if results exist
	if len(m.MockResults) > 0 {
		response.PrimaryClassification = &m.MockResults[0]
	}

	// Add module result
	response.ModuleResults["mock_module"] = shared.ModuleResult{
		ModuleID:        "mock_classification_module",
		ModuleType:      "mock",
		Success:         true,
		Classifications: m.MockResults,
		ProcessingTime:  m.Delay,
		Confidence:      0.85,
		RawData:         make(map[string]interface{}),
		Metadata:        make(map[string]interface{}),
	}

	return response, nil
}

// SetMockResults allows configuring mock results for testing
func (m *MockClassificationService) SetMockResults(results []shared.IndustryClassification) {
	m.MockResults = results
}

// SetFailureMode configures the mock to fail with a specific error
func (m *MockClassificationService) SetFailureMode(shouldFail bool, errorMessage string) {
	m.ShouldFail = shouldFail
	m.ErrorMessage = errorMessage
}

// SetDelay configures the processing delay for the mock
func (m *MockClassificationService) SetDelay(delay time.Duration) {
	m.Delay = delay
}

// getDefaultMockResults returns default mock classification results
func getDefaultMockResults() []shared.IndustryClassification {
	return []shared.IndustryClassification{
		{
			IndustryCode:         "541511",
			IndustryName:         "Custom Computer Programming Services",
			ConfidenceScore:      0.92,
			ClassificationMethod: "mock_keyword_matching",
			Keywords:             []string{"software", "development", "programming", "technology"},
			Description:          "Software development and programming services",
			Evidence:             "Mock classification based on business name and description",
			ProcessingTime:       50 * time.Millisecond,
			Metadata: map[string]interface{}{
				"source":    "mock_service",
				"algorithm": "keyword_matching",
				"version":   "1.0.0",
				"test_mode": true,
			},
		},
		{
			IndustryCode:         "541512",
			IndustryName:         "Computer Systems Design Services",
			ConfidenceScore:      0.78,
			ClassificationMethod: "mock_keyword_matching",
			Keywords:             []string{"systems", "design", "consulting"},
			Description:          "Computer systems design and consulting services",
			Evidence:             "Mock secondary classification based on business description",
			ProcessingTime:       30 * time.Millisecond,
			Metadata: map[string]interface{}{
				"source":    "mock_service",
				"algorithm": "keyword_matching",
				"version":   "1.0.0",
				"test_mode": true,
			},
		},
	}
}

// ClassificationError represents a classification-specific error
type ClassificationError struct {
	Message string
	Code    string
}

func (e *ClassificationError) Error() string {
	return e.Message
}

// MockDatabase provides a mock database implementation for E2E tests
type MockDatabase struct {
	Connected bool
	Error     error
	// Mock data
	Tables map[string][]map[string]interface{}
	// Connection pool config
	MaxOpenConns int
	MaxIdleConns int
}

// NewMockDatabase creates a new mock database
func NewMockDatabase() *MockDatabase {
	mock := &MockDatabase{
		Connected:    true,
		Error:        nil,
		Tables:       make(map[string][]map[string]interface{}),
		MaxOpenConns: 25,
		MaxIdleConns: 5,
	}

	// Initialize with mock data
	mock.initializeMockData()
	return mock
}

// Connect simulates database connection
func (m *MockDatabase) Connect() error {
	if m.Error != nil {
		return m.Error
	}
	m.Connected = true
	return nil
}

// Disconnect simulates database disconnection
func (m *MockDatabase) Disconnect() error {
	m.Connected = false
	return nil
}

// IsConnected returns the connection status
func (m *MockDatabase) IsConnected() bool {
	return m.Connected
}

// Ping simulates database ping
func (m *MockDatabase) Ping() error {
	if !m.Connected {
		return fmt.Errorf("database not connected")
	}
	return nil
}

// TableExists checks if a table exists
func (m *MockDatabase) TableExists(tableName string) (bool, error) {
	if !m.Connected {
		return false, fmt.Errorf("database not connected")
	}
	_, exists := m.Tables[tableName]
	return exists, nil
}

// ColumnExists checks if a column exists in a table
func (m *MockDatabase) ColumnExists(tableName, columnName string) (bool, error) {
	if !m.Connected {
		return false, fmt.Errorf("database not connected")
	}

	table, exists := m.Tables[tableName]
	if !exists {
		return false, nil
	}

	if len(table) == 0 {
		return false, nil
	}

	// Check if column exists in first row
	_, exists = table[0][columnName]
	return exists, nil
}

// GetTableCount returns the count of records in a table
func (m *MockDatabase) GetTableCount(tableName string) (int, error) {
	if !m.Connected {
		return 0, fmt.Errorf("database not connected")
	}

	table, exists := m.Tables[tableName]
	if !exists {
		return 0, nil
	}

	return len(table), nil
}

// ExecuteCountQuery executes a count query
func (m *MockDatabase) ExecuteCountQuery(query string) (int, error) {
	if !m.Connected {
		return 0, fmt.Errorf("database not connected")
	}

	// Mock count queries based on query content
	if strings.Contains(query, "type = 'NAICS'") && !strings.Contains(query, "code !~") {
		return 150, nil // Mock NAICS count
	}
	if strings.Contains(query, "type = 'SIC'") && !strings.Contains(query, "code !~") {
		return 80, nil // Mock SIC count
	}
	if strings.Contains(query, "type = 'MCC'") && !strings.Contains(query, "code !~") {
		return 40, nil // Mock MCC count
	}
	if strings.Contains(query, "weight <= 0") {
		return 0, nil // No negative weights
	}
	if strings.Contains(query, "name IS NULL") {
		return 0, nil // No null names
	}
	if strings.Contains(query, "LEFT JOIN industry_codes ic ON k.industry_code_id = ic.id") {
		return 0, nil // No invalid references in mock
	}
	if strings.Contains(query, "GROUP BY code, type HAVING COUNT(*) > 1") {
		return 0, nil // No duplicates in mock
	}
	if strings.Contains(query, "GROUP BY keyword HAVING COUNT(*) > 1") {
		return 0, nil // No duplicates in mock
	}
	if strings.Contains(query, "duplicate industry codes") {
		return 0, nil // No duplicates in mock
	}
	if strings.Contains(query, "duplicate keywords") {
		return 0, nil // No duplicates in mock
	}
	if strings.Contains(query, "LEFT JOIN keyword_weights kw ON k.keyword = kw.keyword WHERE kw.keyword IS NULL") {
		return 0, nil // All keywords have weights in mock
	}
	if strings.Contains(query, "code !~") {
		return 0, nil // All codes have valid format in mock
	}
	if strings.Contains(query, "invalid format") {
		return 0, nil // All codes have valid format in mock
	}

	return 10, nil // Default count
}

// ExecuteQuery executes a query and returns results
func (m *MockDatabase) ExecuteQuery(query string) ([]map[string]interface{}, error) {
	if !m.Connected {
		return nil, fmt.Errorf("database not connected")
	}

	// Check for connection error
	if m.Error != nil {
		return nil, m.Error
	}

	// Mock query results based on query content
	if strings.Contains(query, "non_existent_table") {
		return nil, fmt.Errorf("relation \"non_existent_table\" does not exist")
	}

	// Return mock results based on query content
	if strings.Contains(query, "industry_codes") && !strings.Contains(query, "GROUP BY") {
		return []map[string]interface{}{
			{"id": 1, "code": "541511", "name": "Custom Computer Programming Services", "type": "NAICS", "description": "Software development services"},
			{"id": 2, "code": "541512", "name": "Computer Systems Design Services", "type": "NAICS", "description": "Computer systems design"},
		}, nil
	}

	if strings.Contains(query, "keywords") && !strings.Contains(query, "GROUP BY") {
		return []map[string]interface{}{
			{"id": 1, "keyword": "software", "industry_code_id": 1, "weight": 1.0},
			{"id": 2, "keyword": "development", "industry_code_id": 1, "weight": 0.9},
		}, nil
	}

	// For duplicate check queries, return empty results
	if strings.Contains(query, "GROUP BY") && strings.Contains(query, "HAVING COUNT(*) > 1") {
		return []map[string]interface{}{}, nil
	}

	// Default mock results
	return []map[string]interface{}{
		{"id": 1, "name": "Test Result 1"},
		{"id": 2, "name": "Test Result 2"},
	}, nil
}

// ExecuteQueryWithContext executes a query with context
func (m *MockDatabase) ExecuteQueryWithContext(ctx context.Context, query string) ([]map[string]interface{}, error) {
	// Check for timeout
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return m.ExecuteQuery(query)
}

// BeginTransaction begins a transaction
func (m *MockDatabase) BeginTransaction() (*MockTransaction, error) {
	if !m.Connected {
		return nil, fmt.Errorf("database not connected")
	}
	return &MockTransaction{db: m, committed: false}, nil
}

// GetConnection gets a connection from the pool
func (m *MockDatabase) GetConnection() (*MockConnection, error) {
	if !m.Connected {
		return nil, fmt.Errorf("database not connected")
	}
	return &MockConnection{db: m}, nil
}

// ReleaseConnection releases a connection back to the pool
func (m *MockDatabase) ReleaseConnection(conn *MockConnection) {
	// Mock implementation - just log
}

// GetConnectionPoolConfig returns connection pool configuration
func (m *MockDatabase) GetConnectionPoolConfig() ConnectionPoolConfig {
	return ConnectionPoolConfig{
		MaxOpenConns: m.MaxOpenConns,
		MaxIdleConns: m.MaxIdleConns,
	}
}

// SetConnectionError allows configuring a connection error for testing
func (m *MockDatabase) SetConnectionError(err error) {
	m.Error = err
}

// initializeMockData initializes the mock database with sample data
func (m *MockDatabase) initializeMockData() {
	// Industry codes table
	m.Tables["industry_codes"] = []map[string]interface{}{
		{"id": 1, "code": "541511", "name": "Custom Computer Programming Services", "type": "NAICS", "description": "Software development services"},
		{"id": 2, "code": "541512", "name": "Computer Systems Design Services", "type": "NAICS", "description": "Computer systems design"},
		{"id": 3, "code": "7372", "name": "Prepackaged Software", "type": "SIC", "description": "Software publishing"},
		{"id": 4, "code": "5734", "name": "Computer Software Stores", "type": "MCC", "description": "Software retail"},
	}

	// Keywords table
	m.Tables["keywords"] = []map[string]interface{}{
		{"id": 1, "keyword": "software", "industry_code_id": 1, "weight": 1.0},
		{"id": 2, "keyword": "development", "industry_code_id": 1, "weight": 0.9},
		{"id": 3, "keyword": "programming", "industry_code_id": 1, "weight": 0.8},
		{"id": 4, "keyword": "technology", "industry_code_id": 2, "weight": 0.7},
	}

	// Keyword weights table
	m.Tables["keyword_weights"] = []map[string]interface{}{
		{"id": 1, "keyword": "software", "weight": 1.0, "category": "primary"},
		{"id": 2, "keyword": "development", "weight": 0.9, "category": "primary"},
		{"id": 3, "keyword": "programming", "weight": 0.8, "category": "secondary"},
		{"id": 4, "keyword": "technology", "weight": 0.7, "category": "secondary"},
	}

	// Classification patterns table
	m.Tables["classification_patterns"] = []map[string]interface{}{
		{"id": 1, "pattern": "software.*development", "industry_code_id": 1, "confidence": 0.9},
		{"id": 2, "pattern": "computer.*programming", "industry_code_id": 1, "confidence": 0.8},
	}
}

// MockTransaction represents a mock database transaction
type MockTransaction struct {
	db        *MockDatabase
	committed bool
}

// Exec executes a query in the transaction
func (t *MockTransaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	if t.committed {
		return nil, fmt.Errorf("transaction already committed")
	}
	return &MockResult{}, nil
}

// Commit commits the transaction
func (t *MockTransaction) Commit() error {
	if t.committed {
		return fmt.Errorf("transaction already committed")
	}
	t.committed = true
	return nil
}

// Rollback rolls back the transaction
func (t *MockTransaction) Rollback() error {
	if t.committed {
		return fmt.Errorf("transaction already committed")
	}
	return nil
}

// MockConnection represents a mock database connection
type MockConnection struct {
	db *MockDatabase
}

// Ping pings the connection
func (c *MockConnection) Ping() error {
	if !c.db.Connected {
		return fmt.Errorf("database not connected")
	}
	return nil
}

// MockResult represents a mock SQL result
type MockResult struct{}

func (r *MockResult) LastInsertId() (int64, error) {
	return 1, nil
}

func (r *MockResult) RowsAffected() (int64, error) {
	return 1, nil
}

// ConnectionPoolConfig represents connection pool configuration
type ConnectionPoolConfig struct {
	MaxOpenConns int
	MaxIdleConns int
}
