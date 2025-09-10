package compatibility

import (
	"context"
	"fmt"
	"testing"

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

// Debug logs a debug level message
func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	// Mock implementation - do nothing
}

// Info logs an info level message
func (m *MockLogger) Info(msg string, fields ...interface{}) {
	// Mock implementation - do nothing
}

// Warn logs a warning level message
func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	// Mock implementation - do nothing
}

// Error logs an error level message
func (m *MockLogger) Error(msg string, fields ...interface{}) {
	// Mock implementation - do nothing
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

// IncrementCounter increments a counter metric
func (m *MockMetrics) IncrementCounter(name string, labels map[string]string) {
	// Mock implementation - do nothing
}

// RecordHistogram records a histogram metric
func (m *MockMetrics) RecordHistogram(name string, value float64, labels map[string]string) {
	// Mock implementation - do nothing
}

// SetGauge sets a gauge metric
func (m *MockMetrics) SetGauge(name string, value float64, labels map[string]string) {
	// Mock implementation - do nothing
}

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

	// Test API version (the method returns a fixed version)
	apiVersion := bcl.GetAPIVersion()
	if apiVersion != "v1.0.0" {
		t.Errorf("Expected API version 'v1.0.0', got '%s'", apiVersion)
	}
}

func TestBackwardCompatibilityLayer_ProcessLegacyRequest(t *testing.T) {
	featureFlagManager := config.NewFeatureFlagManager("test")
	logger := &MockLogger{}
	metrics := &MockMetrics{}
	validator := validators.NewValidator()

	bcl := NewBackwardCompatibilityLayer(featureFlagManager, logger, metrics, validator)

	// Test legacy request processing
	req := &ClassificationRequest{
		BusinessName:    "Test Company",
		BusinessAddress: "123 Test St, Test City, TC 12345",
		Industry:        "Technology",
		Metadata: map[string]interface{}{
			"test": true,
		},
	}

	// Enable legacy mode
	legacyFlag := &config.FeatureFlag{
		Name:    "legacy_classification",
		Enabled: true,
	}
	featureFlagManager.SetFlag(legacyFlag)

	legacyResponse, err := bcl.ProcessRequest(context.Background(), req)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if legacyResponse == nil {
		t.Error("Expected legacy response to not be nil")
	}

	if legacyResponse.Version != "legacy" {
		t.Errorf("Expected version 'legacy', got '%s'", legacyResponse.Version)
	}

	if len(legacyResponse.ClassificationCodes) == 0 {
		t.Error("Expected classification codes to be present")
	}

	if legacyResponse.Confidence <= 0 {
		t.Errorf("Expected confidence > 0, got %f", legacyResponse.Confidence)
	}
}

func TestBackwardCompatibilityLayer_ProcessCurrentRequest(t *testing.T) {
	featureFlagManager := config.NewFeatureFlagManager("test")
	logger := &MockLogger{}
	metrics := &MockMetrics{}
	validator := validators.NewValidator()

	bcl := NewBackwardCompatibilityLayer(featureFlagManager, logger, metrics, validator)

	// Test current request processing
	req := &ClassificationRequest{
		BusinessName:    "Test Company",
		BusinessAddress: "123 Test St, Test City, TC 12345",
		Industry:        "Technology",
		Metadata: map[string]interface{}{
			"test": true,
		},
	}

	// Disable legacy mode to use current processing
	legacyFlag := &config.FeatureFlag{
		Name:    "legacy_classification",
		Enabled: false,
	}
	featureFlagManager.SetFlag(legacyFlag)

	enhancedResponse, err := bcl.ProcessRequest(context.Background(), req)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if enhancedResponse == nil {
		t.Error("Expected enhanced response to not be nil")
	}

	if enhancedResponse.Version != "v1.0.0" {
		t.Errorf("Expected version 'v1.0.0', got '%s'", enhancedResponse.Version)
	}

	if len(enhancedResponse.ClassificationCodes) == 0 {
		t.Error("Expected classification codes to be present")
	}

	if enhancedResponse.Confidence <= 0 {
		t.Errorf("Expected confidence > 0, got %f", enhancedResponse.Confidence)
	}
}

func TestBackwardCompatibilityLayer_IsLegacyModeEnabled(t *testing.T) {
	featureFlagManager := config.NewFeatureFlagManager("test")
	logger := &MockLogger{}
	metrics := &MockMetrics{}
	validator := validators.NewValidator()

	bcl := NewBackwardCompatibilityLayer(featureFlagManager, logger, metrics, validator)

	// Test legacy mode detection
	legacyFlag := &config.FeatureFlag{
		Name:    "legacy_classification",
		Enabled: true,
	}
	featureFlagManager.SetFlag(legacyFlag)
	isLegacy := bcl.IsLegacyModeEnabled()

	if !isLegacy {
		t.Error("Expected legacy mode to be enabled")
	}

	// Test with legacy mode disabled
	legacyFlag.Enabled = false
	featureFlagManager.SetFlag(legacyFlag)
	isLegacy = bcl.IsLegacyModeEnabled()

	if isLegacy {
		t.Error("Expected legacy mode to be disabled")
	}
}

func TestBackwardCompatibilityLayer_GetSupportedVersions(t *testing.T) {
	featureFlagManager := config.NewFeatureFlagManager("test")
	logger := &MockLogger{}
	metrics := &MockMetrics{}
	validator := validators.NewValidator()

	bcl := NewBackwardCompatibilityLayer(featureFlagManager, logger, metrics, validator)

	// Test getting supported versions
	versions := bcl.GetSupportedVersions()

	if len(versions) == 0 {
		t.Error("Expected supported versions to be returned")
	}

	// Check if expected versions are present
	expectedVersions := []string{"v1.0.0", "legacy"}
	for _, expected := range expectedVersions {
		found := false
		for _, version := range versions {
			if version == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected version '%s' to be in supported versions", expected)
		}
	}
}
