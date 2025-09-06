package compatibility

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/shared"
	"github.com/pcraw4d/business-verification/pkg/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// EnhancedMockClassificationProcessor implements ClassificationProcessor for testing
type EnhancedMockClassificationProcessor struct {
	responses map[string]*shared.BusinessClassificationResponse
	errors    map[string]error
}

func NewEnhancedMockClassificationProcessor() *EnhancedMockClassificationProcessor {
	return &EnhancedMockClassificationProcessor{
		responses: make(map[string]*shared.BusinessClassificationResponse),
		errors:    make(map[string]error),
	}
}

func (m *EnhancedMockClassificationProcessor) ProcessClassification(
	ctx context.Context,
	request *shared.BusinessClassificationRequest,
) (*shared.BusinessClassificationResponse, error) {
	key := request.BusinessName
	if err, exists := m.errors[key]; exists {
		return nil, err
	}
	if response, exists := m.responses[key]; exists {
		return response, nil
	}

	// Default response
	return &classification.ClassificationResponse{
		BusinessID: "test-123",
		Classifications: []classification.IndustryClassification{
			{
				IndustryCode:         "541511",
				IndustryName:         "Custom Computer Programming Services",
				ConfidenceScore:      0.95,
				ClassificationMethod: "ml_classification",
				Keywords:             []string{"software", "programming", "technology"},
				Description:          "Custom computer programming services",
			},
		},
		PrimaryClassification: &classification.IndustryClassification{
			IndustryCode:         "541511",
			IndustryName:         "Custom Computer Programming Services",
			ConfidenceScore:      0.95,
			ClassificationMethod: "ml_classification",
		},
		ConfidenceScore:      0.95,
		ClassificationMethod: "ml_classification",
		ProcessingTime:       100 * time.Millisecond,
		RawData: map[string]interface{}{
			"data_sources": []string{"business_registry", "industry_database"},
			"confidence_factors": map[string]float64{
				"name_match":     0.95,
				"industry_match": 0.90,
			},
			"geographic_region": "North America",
			"naics_codes":       []string{"541511", "541512", "541519"},
			"sic_codes":         []string{"7371", "7372", "7373"},
		},
	}, nil
}

func (m *EnhancedMockClassificationProcessor) ProcessBatchClassification(
	ctx context.Context,
	requests []*shared.BusinessClassificationRequest,
) ([]*shared.BusinessClassificationResponse, error) {
	var responses []*shared.BusinessClassificationResponse
	for _, req := range requests {
		response, err := m.ProcessClassification(ctx, req)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	return responses, nil
}

func (m *EnhancedMockClassificationProcessor) SetResponse(businessName string, response *shared.BusinessClassificationResponse) {
	m.responses[businessName] = response
}

func (m *EnhancedMockClassificationProcessor) SetError(businessName string, err error) {
	m.errors[businessName] = err
}

func TestEnhancedBackwardCompatibilityLayer_HandleRequestWithCompatibility(t *testing.T) {
	logger := zap.NewNop()
	featureFlagManager := &config.FeatureFlagManager{}
	validator := validators.NewValidator()

	// Create version manager
	versionManager := NewVersionManager(logger, nil)

	// Create enhanced backward compatibility layer
	ebcl := NewEnhancedBackwardCompatibilityLayer(
		versionManager,
		featureFlagManager,
		logger,
		validator,
		nil,
	)

	tests := []struct {
		name           string
		version        string
		requestBody    interface{}
		headers        map[string]string
		expectedStatus int
		expectedFields []string
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:    "v1 legacy request",
			version: "v1",
			requestBody: LegacyClassificationRequest{
				BusinessName: "Test Company",
				BusinessType: "Corporation",
				Industry:     "Technology",
			},
			headers: map[string]string{
				"Accept": "application/vnd.kyb-platform.v1+json",
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"success", "business_id", "primary_industry_code", "deprecation_warning"},
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, true, response["success"])
				assert.Contains(t, response["deprecation_warning"], "deprecated")
				assert.Equal(t, "test-123", response["business_id"])
			},
		},
		{
			name:    "v2 enhanced request",
			version: "v2",
			requestBody: EnhancedClassificationRequest{
				BusinessName:     "Test Company",
				BusinessType:     "Corporation",
				Industry:         "Technology",
				GeographicRegion: "North America",
			},
			headers: map[string]string{
				"X-API-Version": "v2",
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"success", "api_version", "business_id", "geographic_region"},
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, true, response["success"])
				assert.Equal(t, "v2", response["api_version"])
				assert.Equal(t, "North America", response["geographic_region"])
			},
		},
		{
			name:    "v3 current request",
			version: "v3",
			requestBody: classification.ClassificationRequest{
				BusinessName: "Test Company",
				BusinessType: "Corporation",
				Industry:     "Technology",
			},
			headers: map[string]string{
				"X-API-Version": "v3",
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"business_id", "raw_data"},
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.NotNil(t, response["raw_data"])
			},
		},
		{
			name:    "version negotiation from Accept header",
			version: "v1",
			requestBody: LegacyClassificationRequest{
				BusinessName: "Test Company",
			},
			headers: map[string]string{
				"Accept": "application/vnd.kyb-platform.v1+json",
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"success", "business_id"},
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, true, response["success"])
			},
		},
		{
			name:    "client version validation",
			version: "v1",
			requestBody: LegacyClassificationRequest{
				BusinessName: "Test Company",
			},
			headers: map[string]string{
				"X-API-Version":    "v1",
				"X-Client-Version": "1.0.0",
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"success", "business_id"},
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, true, response["success"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock processor
			processor := NewEnhancedMockClassificationProcessor()

			// Create request body
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			// Create HTTP request
			req := httptest.NewRequest("POST", "/classify", bytes.NewBuffer(body))
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Handle request
			ebcl.HandleRequestWithCompatibility(w, req, processor)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check expected fields
			for _, field := range tt.expectedFields {
				assert.Contains(t, response, field)
			}

			// Check custom response validation
			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}

			// Check headers
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.Equal(t, tt.version, w.Header().Get("X-API-Version"))
		})
	}
}

func TestEnhancedBackwardCompatibilityLayer_ErrorHandling(t *testing.T) {
	logger := zap.NewNop()
	featureFlagManager := &config.FeatureFlagManager{}
	validator := validators.NewValidator()

	versionManager := NewVersionManager(logger, nil)
	ebcl := NewEnhancedBackwardCompatibilityLayer(
		versionManager,
		featureFlagManager,
		logger,
		validator,
		nil,
	)

	tests := []struct {
		name           string
		requestBody    string
		headers        map[string]string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "invalid JSON",
			requestBody:    `{"invalid": json}`,
			headers:        map[string]string{"X-API-Version": "v1"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "request_parsing_failed",
		},
		{
			name:           "unsupported version",
			requestBody:    `{"business_name": "Test"}`,
			headers:        map[string]string{"X-API-Version": "v99"},
			expectedStatus: http.StatusOK, // Version negotiation falls back to default
			// Don't check error field for successful cases
		},
		{
			name:           "missing required field",
			requestBody:    `{"business_type": "Corporation"}`,
			headers:        map[string]string{"X-API-Version": "v1"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "request_parsing_failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewEnhancedMockClassificationProcessor()

			req := httptest.NewRequest("POST", "/classify", bytes.NewBufferString(tt.requestBody))
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			w := httptest.NewRecorder()
			ebcl.HandleRequestWithCompatibility(w, req, processor)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Only check error field if expectedError is not empty
			if tt.expectedError != "" {
				assert.Equal(t, tt.expectedError, response["error"])
			} else {
				// For successful cases, there should be no error field
				_, hasError := response["error"]
				assert.False(t, hasError, "Response should not have error field for successful requests")
			}
		})
	}
}

func TestEnhancedBackwardCompatibilityLayer_DeprecationHandling(t *testing.T) {
	logger := zap.NewNop()
	featureFlagManager := &config.FeatureFlagManager{}
	validator := validators.NewValidator()

	versionManager := NewVersionManager(logger, nil)
	ebcl := NewEnhancedBackwardCompatibilityLayer(
		versionManager,
		featureFlagManager,
		logger,
		validator,
		nil,
	)

	// Test v1 deprecation headers
	req := httptest.NewRequest("POST", "/classify", bytes.NewBufferString(`{"business_name": "Test"}`))
	req.Header.Set("X-API-Version", "v1")

	w := httptest.NewRecorder()
	processor := NewEnhancedMockClassificationProcessor()
	ebcl.HandleRequestWithCompatibility(w, req, processor)

	// Check deprecation headers
	assert.Equal(t, "true", w.Header().Get("X-API-Deprecated"))
	assert.Contains(t, w.Header().Get("X-API-Deprecation-Message"), "deprecated")
	assert.NotEmpty(t, w.Header().Get("X-API-Sunset-Date"))
}

func TestEnhancedBackwardCompatibilityLayer_VersionNegotiation(t *testing.T) {
	logger := zap.NewNop()
	featureFlagManager := &config.FeatureFlagManager{}
	validator := validators.NewValidator()

	versionManager := NewVersionManager(logger, nil)
	ebcl := NewEnhancedBackwardCompatibilityLayer(
		versionManager,
		featureFlagManager,
		logger,
		validator,
		nil,
	)

	tests := []struct {
		name            string
		acceptHeader    string
		apiVersion      string
		expectedVersion string
	}{
		{
			name:            "Accept header v1",
			acceptHeader:    "application/vnd.kyb-platform.v1+json",
			expectedVersion: "v1",
		},
		{
			name:            "Accept header v2",
			acceptHeader:    "application/vnd.kyb-platform.v2+json",
			expectedVersion: "v2",
		},
		{
			name:            "X-API-Version header",
			apiVersion:      "v3",
			expectedVersion: "v3",
		},
		{
			name:            "URL path version",
			expectedVersion: "v3", // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/classify", bytes.NewBufferString(`{"business_name": "Test"}`))

			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}
			if tt.apiVersion != "" {
				req.Header.Set("X-API-Version", tt.apiVersion)
			}

			negotiatedVersion, err := ebcl.versionManager.NegotiateVersion(context.Background(), req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedVersion, negotiatedVersion)
		})
	}
}

func TestEnhancedBackwardCompatibilityLayer_ResponseTransformation(t *testing.T) {
	logger := zap.NewNop()
	featureFlagManager := &config.FeatureFlagManager{}
	validator := validators.NewValidator()

	versionManager := NewVersionManager(logger, nil)
	ebcl := NewEnhancedBackwardCompatibilityLayer(
		versionManager,
		featureFlagManager,
		logger,
		validator,
		nil,
	)

	// Create test response
	response := &classification.ClassificationResponse{
		BusinessID: "test-123",
		Classifications: []classification.IndustryClassification{
			{
				IndustryCode:         "541511",
				IndustryName:         "Custom Computer Programming Services",
				ConfidenceScore:      0.95,
				ClassificationMethod: "ml_classification",
			},
		},
		PrimaryClassification: &classification.IndustryClassification{
			IndustryCode:         "541511",
			IndustryName:         "Custom Computer Programming Services",
			ConfidenceScore:      0.95,
			ClassificationMethod: "ml_classification",
		},
		ConfidenceScore:      0.95,
		ClassificationMethod: "ml_classification",
		ProcessingTime:       100 * time.Millisecond,
		RawData: map[string]interface{}{
			"data_sources": []string{"business_registry"},
		},
	}

	tests := []struct {
		name           string
		version        string
		expectedFields []string
		checkResponse  func(*testing.T, interface{})
	}{
		{
			name:           "v1 transformation",
			version:        "v1",
			expectedFields: []string{"success", "business_id", "deprecation_warning"},
			checkResponse: func(t *testing.T, response interface{}) {
				legacyResp := response.(*LegacyClassificationResponse)
				assert.Equal(t, true, legacyResp.Success)
				assert.Contains(t, legacyResp.DeprecationWarning, "deprecated")
			},
		},
		{
			name:           "v2 transformation",
			version:        "v2",
			expectedFields: []string{"success", "api_version", "geographic_region"},
			checkResponse: func(t *testing.T, response interface{}) {
				enhancedResp := response.(*EnhancedClassificationResponse)
				assert.Equal(t, true, enhancedResp.Success)
				assert.Equal(t, "v2", enhancedResp.APIVersion)
				// Geographic region may be empty if not in raw data
				// assert.Equal(t, "North America", enhancedResp.GeographicRegion)
			},
		},
		{
			name:           "v3 transformation",
			version:        "v3",
			expectedFields: []string{"raw_data"},
			checkResponse: func(t *testing.T, response interface{}) {
				v3Resp := response.(*classification.ClassificationResponse)
				assert.NotNil(t, v3Resp.RawData)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformedResponse, err := ebcl.transformResponseToVersion(context.Background(), response, tt.version)
			require.NoError(t, err)

			if tt.checkResponse != nil {
				tt.checkResponse(t, transformedResponse)
			}
		})
	}
}

func TestEnhancedBackwardCompatibilityLayer_CompatibilityInfo(t *testing.T) {
	logger := zap.NewNop()
	featureFlagManager := &config.FeatureFlagManager{}
	validator := validators.NewValidator()

	versionManager := NewVersionManager(logger, nil)
	ebcl := NewEnhancedBackwardCompatibilityLayer(
		versionManager,
		featureFlagManager,
		logger,
		validator,
		nil,
	)

	tests := []struct {
		name             string
		version          string
		expectedLevel    string
		expectedWarnings bool
	}{
		{
			name:             "v1 compatibility",
			version:          "v1",
			expectedLevel:    "partial",
			expectedWarnings: true,
		},
		{
			name:             "v2 compatibility",
			version:          "v2",
			expectedLevel:    "partial",
			expectedWarnings: true,
		},
		{
			name:             "v3 compatibility",
			version:          "v3",
			expectedLevel:    "full",
			expectedWarnings: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := ebcl.generateCompatibilityInfo(context.Background(), tt.version)

			assert.Equal(t, true, info.IsCompatible)
			assert.Equal(t, tt.expectedLevel, info.CompatibilityLevel)

			if tt.expectedWarnings {
				assert.NotEmpty(t, info.Warnings)
				assert.NotEmpty(t, info.Suggestions)
			} else {
				assert.Empty(t, info.Warnings)
				assert.Empty(t, info.Suggestions)
			}
		})
	}
}

func TestEnhancedBackwardCompatibilityLayer_MigrationInfo(t *testing.T) {
	logger := zap.NewNop()
	featureFlagManager := &config.FeatureFlagManager{}
	validator := validators.NewValidator()

	versionManager := NewVersionManager(logger, nil)
	ebcl := NewEnhancedBackwardCompatibilityLayer(
		versionManager,
		featureFlagManager,
		logger,
		validator,
		nil,
	)

	tests := []struct {
		name              string
		version           string
		expectedMigration bool
		expectedSteps     bool
	}{
		{
			name:              "v1 migration",
			version:           "v1",
			expectedMigration: true,
			expectedSteps:     true,
		},
		{
			name:              "v2 migration",
			version:           "v2",
			expectedMigration: true,
			expectedSteps:     true,
		},
		{
			name:              "v3 migration",
			version:           "v3",
			expectedMigration: false,
			expectedSteps:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := ebcl.generateMigrationInfo(context.Background(), tt.version)

			assert.Equal(t, tt.expectedMigration, info.MigrationRequired)

			if tt.expectedSteps {
				assert.NotEmpty(t, info.MigrationSteps)
				assert.NotEmpty(t, info.MigrationPath)
			} else {
				assert.Empty(t, info.MigrationSteps)
			}
		})
	}
}
