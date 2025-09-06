package mocks

import (
	"context"
	"time"

	"github.com/pcraw4d/business-verification/internal/shared"
)

// MockClassificationService provides a mock implementation of classification service for E2E tests
type MockClassificationService struct {
	// Configuration for mock behavior
	ShouldFail    bool
	Delay         time.Duration
	MockResults   []shared.IndustryClassification
	ErrorMessage  string
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
		ID:                    request.ID,
		BusinessName:          request.BusinessName,
		Classifications:       m.MockResults,
		OverallConfidence:     0.85,
		ClassificationMethod:  "mock_classification",
		ProcessingTime:        m.Delay,
		ModuleResults:         make(map[string]shared.ModuleResult),
		RawData:               make(map[string]interface{}),
		CreatedAt:             time.Now(),
		Metadata:              make(map[string]interface{}),
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
				"source":        "mock_service",
				"algorithm":     "keyword_matching",
				"version":       "1.0.0",
				"test_mode":     true,
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
				"source":        "mock_service",
				"algorithm":     "keyword_matching",
				"version":       "1.0.0",
				"test_mode":     true,
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
}

// NewMockDatabase creates a new mock database
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		Connected: true,
		Error:     nil,
	}
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

// SetConnectionError allows configuring a connection error for testing
func (m *MockDatabase) SetConnectionError(err error) {
	m.Error = err
}
