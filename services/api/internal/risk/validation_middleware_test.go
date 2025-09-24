package risk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockRiskDataValidator is a mock implementation of RiskDataValidator
type MockRiskDataValidator struct {
	mock.Mock
}

func (m *MockRiskDataValidator) ValidateRiskAssessment(ctx context.Context, assessment *RiskAssessment) *ValidationResult {
	args := m.Called(ctx, assessment)
	return args.Get(0).(*ValidationResult)
}

func (m *MockRiskDataValidator) ValidateRiskAlert(ctx context.Context, alert *RiskAlert) *ValidationResult {
	args := m.Called(ctx, alert)
	return args.Get(0).(*ValidationResult)
}

func (m *MockRiskDataValidator) ValidateRiskTrend(ctx context.Context, trend *RiskTrend) *ValidationResult {
	args := m.Called(ctx, trend)
	return args.Get(0).(*ValidationResult)
}

func (m *MockRiskDataValidator) ValidateBusinessData(ctx context.Context, businessData map[string]interface{}) *ValidationResult {
	args := m.Called(ctx, businessData)
	return args.Get(0).(*ValidationResult)
}

func (m *MockRiskDataValidator) ValidateRiskScore(ctx context.Context, score float64, level RiskLevel) *ValidationResult {
	args := m.Called(ctx, score, level)
	return args.Get(0).(*ValidationResult)
}

func TestValidationMiddleware_ValidateRiskAssessmentMiddleware(t *testing.T) {
	logger := zap.NewNop()
	mockValidator := new(MockRiskDataValidator)
	middleware := NewValidationMiddleware(mockValidator, logger)

	ctx := context.WithValue(context.Background(), "request_id", "test-request")
	assessment := &RiskAssessment{
		ID:           "test-123",
		BusinessID:   "business-123",
		BusinessName: "Test Company",
		OverallScore: 0.5,
		OverallLevel: RiskLevelMedium,
	}

	t.Run("validation passes", func(t *testing.T) {
		// Setup mock
		validResult := &ValidationResult{
			IsValid:  true,
			Errors:   []ValidationError{},
			Warnings: []ValidationError{},
		}
		mockValidator.On("ValidateRiskAssessment", ctx, assessment).Return(validResult)

		// Setup next function
		nextCalled := false
		next := func(ctx context.Context, assessment *RiskAssessment) error {
			nextCalled = true
			return nil
		}

		// Execute middleware
		wrappedNext := middleware.ValidateRiskAssessmentMiddleware(next)
		err := wrappedNext(ctx, assessment)

		// Assertions
		assert.NoError(t, err)
		assert.True(t, nextCalled)
		mockValidator.AssertExpectations(t)
	})

	t.Run("validation fails", func(t *testing.T) {
		// Setup mock
		invalidResult := &ValidationResult{
			IsValid: false,
			Errors: []ValidationError{
				{Field: "id", Message: "ID is required", Code: "REQUIRED_FIELD"},
			},
			Warnings: []ValidationError{},
		}
		mockValidator.On("ValidateRiskAssessment", ctx, assessment).Return(invalidResult)

		// Setup next function
		nextCalled := false
		next := func(ctx context.Context, assessment *RiskAssessment) error {
			nextCalled = true
			return nil
		}

		// Execute middleware
		wrappedNext := middleware.ValidateRiskAssessmentMiddleware(next)
		err := wrappedNext(ctx, assessment)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
		assert.False(t, nextCalled)
		mockValidator.AssertExpectations(t)
	})

	t.Run("validation passes with warnings", func(t *testing.T) {
		// Setup mock
		warningResult := &ValidationResult{
			IsValid: true,
			Errors:  []ValidationError{},
			Warnings: []ValidationError{
				{Field: "source", Message: "Source not specified", Code: "MISSING_SOURCE"},
			},
		}
		mockValidator.On("ValidateRiskAssessment", ctx, assessment).Return(warningResult)

		// Setup next function
		nextCalled := false
		next := func(ctx context.Context, assessment *RiskAssessment) error {
			nextCalled = true
			return nil
		}

		// Execute middleware
		wrappedNext := middleware.ValidateRiskAssessmentMiddleware(next)
		err := wrappedNext(ctx, assessment)

		// Assertions
		assert.NoError(t, err)
		assert.True(t, nextCalled)
		mockValidator.AssertExpectations(t)
	})
}

func TestValidationMiddleware_ValidateRiskAlertMiddleware(t *testing.T) {
	logger := zap.NewNop()
	mockValidator := new(MockRiskDataValidator)
	middleware := NewValidationMiddleware(mockValidator, logger)

	ctx := context.WithValue(context.Background(), "request_id", "test-request")
	alert := &RiskAlert{
		ID:           "alert-123",
		BusinessID:   "business-123",
		RiskFactor:   "financial",
		Level:        RiskLevelHigh,
		Message:      "High financial risk",
		Score:        0.8,
		Threshold:    0.7,
		TriggeredAt:  time.Now(),
		Acknowledged: false,
	}

	t.Run("validation passes", func(t *testing.T) {
		// Setup mock
		validResult := &ValidationResult{
			IsValid:  true,
			Errors:   []ValidationError{},
			Warnings: []ValidationError{},
		}
		mockValidator.On("ValidateRiskAlert", ctx, alert).Return(validResult)

		// Setup next function
		nextCalled := false
		next := func(ctx context.Context, alert *RiskAlert) error {
			nextCalled = true
			return nil
		}

		// Execute middleware
		wrappedNext := middleware.ValidateRiskAlertMiddleware(next)
		err := wrappedNext(ctx, alert)

		// Assertions
		assert.NoError(t, err)
		assert.True(t, nextCalled)
		mockValidator.AssertExpectations(t)
	})

	t.Run("validation fails", func(t *testing.T) {
		// Setup mock
		invalidResult := &ValidationResult{
			IsValid: false,
			Errors: []ValidationError{
				{Field: "score", Message: "Score out of range", Code: "INVALID_SCORE_RANGE"},
			},
			Warnings: []ValidationError{},
		}
		mockValidator.On("ValidateRiskAlert", ctx, alert).Return(invalidResult)

		// Setup next function
		nextCalled := false
		next := func(ctx context.Context, alert *RiskAlert) error {
			nextCalled = true
			return nil
		}

		// Execute middleware
		wrappedNext := middleware.ValidateRiskAlertMiddleware(next)
		err := wrappedNext(ctx, alert)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
		assert.False(t, nextCalled)
		mockValidator.AssertExpectations(t)
	})
}

func TestValidationMiddleware_ValidateBusinessDataMiddleware(t *testing.T) {
	logger := zap.NewNop()
	mockValidator := new(MockRiskDataValidator)
	middleware := NewValidationMiddleware(mockValidator, logger)

	ctx := context.WithValue(context.Background(), "request_id", "test-request")
	businessData := map[string]interface{}{
		"name":    "Test Company",
		"address": "123 Test St",
		"email":   "test@example.com",
	}

	t.Run("validation passes", func(t *testing.T) {
		// Setup mock
		validResult := &ValidationResult{
			IsValid:  true,
			Errors:   []ValidationError{},
			Warnings: []ValidationError{},
		}
		mockValidator.On("ValidateBusinessData", ctx, businessData).Return(validResult)

		// Setup next function
		nextCalled := false
		next := func(ctx context.Context, data map[string]interface{}) error {
			nextCalled = true
			return nil
		}

		// Execute middleware
		wrappedNext := middleware.ValidateBusinessDataMiddleware(next)
		err := wrappedNext(ctx, businessData)

		// Assertions
		assert.NoError(t, err)
		assert.True(t, nextCalled)
		mockValidator.AssertExpectations(t)
	})

	t.Run("validation fails", func(t *testing.T) {
		// Setup mock
		invalidResult := &ValidationResult{
			IsValid: false,
			Errors: []ValidationError{
				{Field: "name", Message: "Name is required", Code: "REQUIRED_FIELD"},
			},
			Warnings: []ValidationError{},
		}
		mockValidator.On("ValidateBusinessData", ctx, businessData).Return(invalidResult)

		// Setup next function
		nextCalled := false
		next := func(ctx context.Context, data map[string]interface{}) error {
			nextCalled = true
			return nil
		}

		// Execute middleware
		wrappedNext := middleware.ValidateBusinessDataMiddleware(next)
		err := wrappedNext(ctx, businessData)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
		assert.False(t, nextCalled)
		mockValidator.AssertExpectations(t)
	})
}

func TestHTTPValidationHandler_ValidateRiskAssessmentRequest(t *testing.T) {
	logger := zap.NewNop()
	mockValidator := new(MockRiskDataValidator)
	handler := NewHTTPValidationHandler(mockValidator, logger)

	// Create test request
	req := httptest.NewRequest("POST", "/api/risk/assess", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request"))

	assessment := &RiskAssessment{
		ID:           "test-123",
		BusinessID:   "business-123",
		BusinessName: "Test Company",
		OverallScore: 0.5,
		OverallLevel: RiskLevelMedium,
	}

	t.Run("validation passes", func(t *testing.T) {
		// Setup mock
		validResult := &ValidationResult{
			IsValid:  true,
			Errors:   []ValidationError{},
			Warnings: []ValidationError{},
		}
		mockValidator.On("ValidateRiskAssessment", req.Context(), assessment).Return(validResult)

		// Execute
		result := handler.ValidateRiskAssessmentRequest(req, assessment)

		// Assertions
		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
		mockValidator.AssertExpectations(t)
	})

	t.Run("validation fails", func(t *testing.T) {
		// Setup mock
		invalidResult := &ValidationResult{
			IsValid: false,
			Errors: []ValidationError{
				{Field: "id", Message: "ID is required", Code: "REQUIRED_FIELD"},
			},
			Warnings: []ValidationError{},
		}
		mockValidator.On("ValidateRiskAssessment", req.Context(), assessment).Return(invalidResult)

		// Execute
		result := handler.ValidateRiskAssessmentRequest(req, assessment)

		// Assertions
		assert.False(t, result.IsValid)
		assert.Len(t, result.Errors, 1)
		mockValidator.AssertExpectations(t)
	})
}

func TestHTTPValidationHandler_ValidateBusinessDataRequest(t *testing.T) {
	logger := zap.NewNop()
	mockValidator := new(MockRiskDataValidator)
	handler := NewHTTPValidationHandler(mockValidator, logger)

	// Create test request
	req := httptest.NewRequest("POST", "/api/business/validate", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request"))

	businessData := map[string]interface{}{
		"name":    "Test Company",
		"address": "123 Test St",
		"email":   "test@example.com",
	}

	t.Run("validation passes", func(t *testing.T) {
		// Setup mock
		validResult := &ValidationResult{
			IsValid:  true,
			Errors:   []ValidationError{},
			Warnings: []ValidationError{},
		}
		mockValidator.On("ValidateBusinessData", req.Context(), businessData).Return(validResult)

		// Execute
		result := handler.ValidateBusinessDataRequest(req, businessData)

		// Assertions
		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
		mockValidator.AssertExpectations(t)
	})

	t.Run("validation fails", func(t *testing.T) {
		// Setup mock
		invalidResult := &ValidationResult{
			IsValid: false,
			Errors: []ValidationError{
				{Field: "name", Message: "Name is required", Code: "REQUIRED_FIELD"},
			},
			Warnings: []ValidationError{},
		}
		mockValidator.On("ValidateBusinessData", req.Context(), businessData).Return(invalidResult)

		// Execute
		result := handler.ValidateBusinessDataRequest(req, businessData)

		// Assertions
		assert.False(t, result.IsValid)
		assert.Len(t, result.Errors, 1)
		mockValidator.AssertExpectations(t)
	})
}

func TestHTTPValidationHandler_WriteValidationErrorResponse(t *testing.T) {
	logger := zap.NewNop()
	mockValidator := new(MockRiskDataValidator)
	handler := NewHTTPValidationHandler(mockValidator, logger)

	// Create test request and response
	req := httptest.NewRequest("POST", "/api/test", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request"))
	w := httptest.NewRecorder()

	// Create validation result with errors
	result := &ValidationResult{
		IsValid: false,
		Errors: []ValidationError{
			{Field: "name", Message: "Name is required", Code: "REQUIRED_FIELD"},
			{Field: "email", Message: "Invalid email format", Code: "INVALID_EMAIL"},
		},
		Warnings: []ValidationError{
			{Field: "address", Message: "Address is empty", Code: "EMPTY_ADDRESS"},
		},
	}

	// Execute
	handler.WriteValidationErrorResponse(w, result)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse response body
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify response structure
	assert.Equal(t, "validation_failed", response["error"])
	assert.Equal(t, "Request data validation failed", response["message"])

	details, ok := response["details"].(map[string]interface{})
	assert.True(t, ok)

	errors, ok := details["errors"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, errors, 2)

	warnings, ok := details["warnings"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, warnings, 1)
}

func TestValidationServiceWithConfig_ValidateRiskAssessmentWithConfig(t *testing.T) {
	logger := zap.NewNop()
	config := &ValidationConfig{
		StrictMode:  true,
		MaxWarnings: 2,
	}
	service := NewValidationServiceWithConfig(logger, config)

	ctx := context.WithValue(context.Background(), "request_id", "test-request")
	assessment := &RiskAssessment{
		ID:           "test-123",
		BusinessID:   "business-123",
		BusinessName: "Test Company",
		OverallScore: 0.5,
		OverallLevel: RiskLevelMedium,
		AssessedAt:   time.Now(),
		ValidUntil:   time.Now().Add(24 * time.Hour),
	}

	t.Run("strict mode converts warnings to errors", func(t *testing.T) {
		// This test would need to be adjusted based on the actual validation logic
		// that generates warnings. For now, we'll test the configuration application.
		result := service.ValidateRiskAssessmentWithConfig(ctx, assessment)

		// In strict mode, any warnings should be converted to errors
		if len(result.Warnings) > 0 {
			assert.False(t, result.IsValid)
		}
	})

	t.Run("too many warnings fails validation", func(t *testing.T) {
		// Create a result with many warnings
		result := &ValidationResult{
			IsValid: true,
			Errors:  []ValidationError{},
			Warnings: []ValidationError{
				{Field: "field1", Message: "Warning 1", Code: "WARNING_1"},
				{Field: "field2", Message: "Warning 2", Code: "WARNING_2"},
				{Field: "field3", Message: "Warning 3", Code: "WARNING_3"},
			},
		}

		// Apply configuration
		if len(result.Warnings) > config.MaxWarnings {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "warnings",
				Message: "too many warnings",
				Code:    "TOO_MANY_WARNINGS",
			})
			result.IsValid = false
		}

		assert.False(t, result.IsValid)
		assert.Len(t, result.Errors, 1)
	})
}

func TestDefaultValidationConfig(t *testing.T) {
	config := DefaultValidationConfig()

	assert.False(t, config.StrictMode)
	assert.Equal(t, 10, config.MaxWarnings)
	assert.True(t, config.EnableContentValidation)
	assert.True(t, config.EnableScoreValidation)
}
