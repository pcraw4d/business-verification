package compatibility

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/shared"
	"github.com/pcraw4d/business-verification/pkg/validators"
)

// MockClassificationProcessor implements ClassificationProcessor for testing
type MockClassificationProcessor struct {
	shouldError bool
	response    *shared.BusinessClassificationResponse
	responses   []*shared.BusinessClassificationResponse
}

func (m *MockClassificationProcessor) ProcessClassification(
	ctx context.Context,
	request *shared.BusinessClassificationRequest,
) (*shared.BusinessClassificationResponse, error) {
	if m.shouldError {
		return nil, fmt.Errorf("Mock classification error")
	}
	return m.response, nil
}

func (m *MockClassificationProcessor) ProcessBatchClassification(
	ctx context.Context,
	requests []*shared.BusinessClassificationRequest,
) ([]*shared.BusinessClassificationResponse, error) {
	if m.shouldError {
		return nil, fmt.Errorf("Mock batch classification error")
	}
	return m.responses, nil
}

// MockLogger implements a simple logger for testing
type MockLogger struct{}

func (m *MockLogger) WithComponent(component string) interface {
	WithError(err error) interface {
		Error(msg string)
	}
	LogBusinessEvent(ctx context.Context, event string, businessID string, metadata map[string]interface{})
} {
	return &MockLoggerComponent{}
}

// MockLoggerComponent implements the component-specific logger interface
type MockLoggerComponent struct{}

func (m *MockLoggerComponent) WithError(err error) interface {
	Error(msg string)
} {
	return &MockLoggerError{}
}

func (m *MockLoggerComponent) LogBusinessEvent(ctx context.Context, event string, businessID string, metadata map[string]interface{}) {
	// Mock implementation - do nothing
}

// MockLoggerError implements the error-specific logger interface
type MockLoggerError struct{}

func (m *MockLoggerError) Error(msg string) {
	// Mock implementation - do nothing
}

// MockMetrics implements a simple metrics for testing
type MockMetrics struct{}

func (m *MockMetrics) RecordBusinessClassification(metric string, value string) {}

func TestNewBackwardCompatibilityLayer(t *testing.T) {
	featureFlagManager := config.NewFeatureFlagManager("test")
	logger := &MockLogger{}
	metrics := &MockMetrics{}
	validator := validators.NewValidator()

	bcl := NewBackwardCompatibilityLayer(featureFlagManager, logger, metrics, validator)

	if bcl == nil {
		t.Fatal("Expected BackwardCompatibilityLayer to be created")
	}

	if bcl.featureFlagManager != featureFlagManager {
		t.Error("Expected feature flag manager to be set")
	}

	if bcl.validator != validator {
		t.Error("Expected validator to be set")
	}
}

func TestBackwardCompatibilityLayer_GetAPIVersion(t *testing.T) {
	featureFlagManager := config.NewFeatureFlagManager("test")
	logger := &MockLogger{}
	metrics := &MockMetrics{}
	validator := validators.NewValidator()

	bcl := NewBackwardCompatibilityLayer(featureFlagManager, logger, metrics, validator)

	// Test Accept header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept", "application/vnd.kyb.v2+json")

	apiVersion := bcl.getAPIVersion(req)
	if apiVersion != "v2" {
		t.Errorf("Expected API version 'v2' from Accept header, got '%s'", apiVersion)
	}

	// Test X-API-Version header
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Version", "v1")

	apiVersion = bcl.getAPIVersion(req)
	if apiVersion != "v1" {
		t.Errorf("Expected API version 'v1' from X-API-Version header, got '%s'", apiVersion)
	}

	// Test query parameter
	req = httptest.NewRequest("GET", "/test?api_version=v2", nil)

	apiVersion = bcl.getAPIVersion(req)
	if apiVersion != "v2" {
		t.Errorf("Expected API version 'v2' from query parameter, got '%s'", apiVersion)
	}

	// Test default
	req = httptest.NewRequest("GET", "/test", nil)

	apiVersion = bcl.getAPIVersion(req)
	if apiVersion != "v2" {
		t.Errorf("Expected default API version 'v2', got '%s'", apiVersion)
	}
}

func TestBackwardCompatibilityLayer_ConvertToLegacyResponse(t *testing.T) {
	featureFlagManager := config.NewFeatureFlagManager("test")
	logger := &MockLogger{}
	metrics := &MockMetrics{}
	validator := validators.NewValidator()

	bcl := NewBackwardCompatibilityLayer(featureFlagManager, logger, metrics, validator)

	// Test conversion
	internalResponse := &classification.ClassificationResponse{
		BusinessID: "test-business-123",
		Classifications: []classification.IndustryClassification{
			{
				IndustryCode:    "541511",
				IndustryName:    "Custom Computer Programming Services",
				ConfidenceScore: 0.95,
			},
		},
		PrimaryClassification: &classification.IndustryClassification{
			IndustryCode:    "541511",
			IndustryName:    "Custom Computer Programming Services",
			ConfidenceScore: 0.95,
		},
		ConfidenceScore:      0.95,
		ClassificationMethod: "keyword_based",
		ProcessingTime:       time.Millisecond * 150,
	}

	legacyResponse := bcl.convertToLegacyResponse(internalResponse)

	if !legacyResponse.Success {
		t.Error("Expected legacy response to be successful")
	}

	if legacyResponse.BusinessID != "test-business-123" {
		t.Errorf("Expected BusinessID 'test-business-123', got '%s'", legacyResponse.BusinessID)
	}

	if legacyResponse.ClassificationMethod != "keyword_based" {
		t.Errorf("Expected ClassificationMethod 'keyword_based', got '%s'", legacyResponse.ClassificationMethod)
	}

	if legacyResponse.OverallConfidence != 0.95 {
		t.Errorf("Expected OverallConfidence 0.95, got %f", legacyResponse.OverallConfidence)
	}
}

func TestBackwardCompatibilityLayer_ConvertToEnhancedResponse(t *testing.T) {
	featureFlagManager := config.NewFeatureFlagManager("test")
	logger := &MockLogger{}
	metrics := &MockMetrics{}
	validator := validators.NewValidator()

	bcl := NewBackwardCompatibilityLayer(featureFlagManager, logger, metrics, validator)

	// Test conversion
	internalResponse := &classification.ClassificationResponse{
		BusinessID: "test-business-123",
		Classifications: []classification.IndustryClassification{
			{
				IndustryCode:    "541511",
				IndustryName:    "Custom Computer Programming Services",
				ConfidenceScore: 0.95,
			},
		},
		PrimaryClassification: &classification.IndustryClassification{
			IndustryCode:    "541511",
			IndustryName:    "Custom Computer Programming Services",
			ConfidenceScore: 0.95,
		},
		ConfidenceScore:      0.95,
		ClassificationMethod: "ml_enhanced",
		ProcessingTime:       time.Millisecond * 200,
		RawData: map[string]interface{}{
			"geographic_region": "North America",
			"enhanced_metadata": map[string]interface{}{
				"ml_model_version": "v2.1",
			},
			"industry_specific_data": map[string]interface{}{
				"tech_stack": "Go, React, PostgreSQL",
			},
		},
	}

	enhancedResponse := bcl.convertToEnhancedResponse(internalResponse, "v2")

	if !enhancedResponse.Success {
		t.Error("Expected enhanced response to be successful")
	}

	if enhancedResponse.APIVersion != "v2" {
		t.Errorf("Expected APIVersion 'v2', got '%s'", enhancedResponse.APIVersion)
	}

	if enhancedResponse.OverallConfidence != 0.95 {
		t.Errorf("Expected OverallConfidence 0.95, got %f", enhancedResponse.OverallConfidence)
	}
}

func TestBackwardCompatibilityLayer_CalculateRegionConfidence(t *testing.T) {
	featureFlagManager := config.NewFeatureFlagManager("test")
	logger := &MockLogger{}
	metrics := &MockMetrics{}
	validator := validators.NewValidator()

	bcl := NewBackwardCompatibilityLayer(featureFlagManager, logger, metrics, validator)

	// Test region confidence calculation
	response := &classification.ClassificationResponse{
		ConfidenceScore: 0.85,
		PrimaryClassification: &classification.IndustryClassification{
			IndustryCode: "541511",
		},
	}

	confidence := bcl.calculateRegionConfidence("North America", response)

	if confidence < 0.85 || confidence > 0.90 {
		t.Errorf("Expected confidence between 0.85 and 0.90, got %f", confidence)
	}

	// Test with no region
	confidence = bcl.calculateRegionConfidence("", response)
	if confidence != 0.85 {
		t.Errorf("Expected confidence 0.85 for no region, got %f", confidence)
	}

	// Test with high base confidence
	response.ConfidenceScore = 0.98
	confidence = bcl.calculateRegionConfidence("North America", response)
	if confidence > 1.0 {
		t.Errorf("Expected confidence <= 1.0, got %f", confidence)
	}
}

func TestBackwardCompatibilityLayer_HandleAPIVersionInfo(t *testing.T) {
	featureFlagManager := config.NewFeatureFlagManager("test")
	logger := &MockLogger{}
	metrics := &MockMetrics{}
	validator := validators.NewValidator()

	bcl := NewBackwardCompatibilityLayer(featureFlagManager, logger, metrics, validator)

	req := httptest.NewRequest("GET", "/api/versions", nil)
	w := httptest.NewRecorder()

	bcl.HandleAPIVersionInfo(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}

	// Parse response
	var versionInfo map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&versionInfo); err != nil {
		t.Fatalf("Failed to decode version info: %v", err)
	}

	if versionInfo["current_version"] != "v2" {
		t.Errorf("Expected current_version 'v2', got '%v'", versionInfo["current_version"])
	}

	supportedVersions, ok := versionInfo["supported_versions"].([]interface{})
	if !ok {
		t.Fatal("Expected supported_versions to be an array")
	}

	if len(supportedVersions) != 2 {
		t.Errorf("Expected 2 supported versions, got %d", len(supportedVersions))
	}

	featureFlags, ok := versionInfo["feature_flags"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected feature_flags to be a map")
	}

	if featureFlags["modular_architecture"] == nil {
		t.Error("Expected modular_architecture feature flag")
	}
}
