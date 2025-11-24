package jobs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"kyb-platform/services/merchant-service/internal/config"
	"kyb-platform/services/merchant-service/internal/supabase"
)

// mockSupabaseClient is a mock Supabase client for testing
type mockSupabaseClient struct {
	records map[string]map[string]interface{}
}

func newMockSupabaseClient() *mockSupabaseClient {
	return &mockSupabaseClient{
		records: make(map[string]map[string]interface{}),
	}
}

func (m *mockSupabaseClient) GetClient() *supabase.Client {
	// Return nil for testing - we'll mock the database operations
	return nil
}

func TestClassificationJob_GetID(t *testing.T) {
	job := NewClassificationJob(
		"merchant_123",
		"Test Business",
		"Test Description",
		"https://test.com",
		nil,
		&config.Config{Environment: "test"},
		zaptest.NewLogger(t),
	)

	assert.NotEmpty(t, job.GetID())
	assert.Contains(t, job.GetID(), "classification_")
	assert.Contains(t, job.GetID(), "merchant_123")
}

func TestClassificationJob_GetMerchantID(t *testing.T) {
	job := NewClassificationJob(
		"merchant_123",
		"Test Business",
		"Test Description",
		"",
		nil,
		&config.Config{Environment: "test"},
		zaptest.NewLogger(t),
	)

	assert.Equal(t, "merchant_123", job.GetMerchantID())
}

func TestClassificationJob_GetType(t *testing.T) {
	job := NewClassificationJob(
		"merchant_123",
		"Test Business",
		"Test Description",
		"",
		nil,
		&config.Config{Environment: "test"},
		zaptest.NewLogger(t),
	)

	assert.Equal(t, "classification", job.GetType())
}

func TestClassificationJob_SetStatus(t *testing.T) {
	job := NewClassificationJob(
		"merchant_123",
		"Test Business",
		"Test Description",
		"",
		nil,
		&config.Config{Environment: "test"},
		zaptest.NewLogger(t),
	)

	assert.Equal(t, StatusPending, job.GetStatus())
	
	job.SetStatus(StatusProcessing)
	assert.Equal(t, StatusProcessing, job.GetStatus())
	
	job.SetStatus(StatusCompleted)
	assert.Equal(t, StatusCompleted, job.GetStatus())
}

func TestClassificationJob_extractClassificationFromResponse(t *testing.T) {
	job := NewClassificationJob(
		"merchant_123",
		"Test Business",
		"Test Description",
		"",
		nil,
		&config.Config{Environment: "test"},
		zaptest.NewLogger(t),
	)

	tests := []struct {
		name     string
		response map[string]interface{}
		expected *ClassificationResult
	}{
		{
			name: "standard response format",
			response: map[string]interface{}{
				"primary_industry": "Technology",
				"confidence_score": 0.95,
				"risk_level":       "low",
				"mcc_codes": []interface{}{
					map[string]interface{}{
						"code":        "5734",
						"description": "Computer Software Stores",
						"confidence":  0.9,
					},
				},
				"sic_codes": []interface{}{
					map[string]interface{}{
						"code":        "7372",
						"description": "Prepackaged Software",
						"confidence":  0.85,
					},
				},
				"naics_codes": []interface{}{
					map[string]interface{}{
						"code":        "541511",
						"description": "Custom Computer Programming Services",
						"confidence":  0.92,
					},
				},
			},
			expected: &ClassificationResult{
				PrimaryIndustry: "Technology",
				ConfidenceScore: 0.95,
				RiskLevel:       "low",
				Status:          "completed",
				MCCCodes: []IndustryCode{
					{Code: "5734", Description: "Computer Software Stores", Confidence: 0.9},
				},
				SICCodes: []IndustryCode{
					{Code: "7372", Description: "Prepackaged Software", Confidence: 0.85},
				},
				NAICSCodes: []IndustryCode{
					{Code: "541511", Description: "Custom Computer Programming Services", Confidence: 0.92},
				},
			},
		},
		{
			name: "alternative response format",
			response: map[string]interface{}{
				"industry":   "Retail",
				"confidence": 0.8,
			},
			expected: &ClassificationResult{
				PrimaryIndustry: "Retail",
				ConfidenceScore: 0.8,
				RiskLevel:       "medium",
				Status:          "completed",
			},
		},
		{
			name: "enhanced classification format",
			response: map[string]interface{}{
				"enhanced_classification": map[string]interface{}{
					"primary_industry": "Finance",
					"mcc_codes": []interface{}{
						map[string]interface{}{
							"code":        "6012",
							"description": "Financial Institutions",
							"confidence":  0.88,
						},
					},
				},
			},
			expected: &ClassificationResult{
				PrimaryIndustry: "Finance",
				RiskLevel:       "medium",
				Status:          "completed",
				MCCCodes: []IndustryCode{
					{Code: "6012", Description: "Financial Institutions", Confidence: 0.88},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := job.extractClassificationFromResponse(tt.response)
			
			assert.Equal(t, tt.expected.PrimaryIndustry, result.PrimaryIndustry)
			assert.Equal(t, tt.expected.ConfidenceScore, result.ConfidenceScore)
			assert.Equal(t, tt.expected.RiskLevel, result.RiskLevel)
			assert.Equal(t, tt.expected.Status, result.Status)
			
			if len(tt.expected.MCCCodes) > 0 {
				assert.Equal(t, len(tt.expected.MCCCodes), len(result.MCCCodes))
				for i, expectedCode := range tt.expected.MCCCodes {
					assert.Equal(t, expectedCode.Code, result.MCCCodes[i].Code)
					assert.Equal(t, expectedCode.Description, result.MCCCodes[i].Description)
					assert.Equal(t, expectedCode.Confidence, result.MCCCodes[i].Confidence)
				}
			}
		})
	}
}

func TestClassificationJob_callClassificationService(t *testing.T) {
	// Create a mock classification service server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/classify", r.URL.Path)
		
		var reqBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		
		assert.Equal(t, "Test Business", reqBody["business_name"])
		assert.Equal(t, "Test Description", reqBody["description"])
		assert.Equal(t, "https://test.com", reqBody["website_url"])
		
		response := map[string]interface{}{
			"primary_industry": "Technology",
			"confidence_score": 0.95,
			"risk_level":       "low",
			"mcc_codes": []interface{}{
				map[string]interface{}{
					"code":        "5734",
					"description": "Computer Software Stores",
					"confidence":  0.9,
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// Create job with mock server URL
	cfg := &config.Config{
		Environment: "test",
	}
	
	// Set environment variable to use mock server
	t.Setenv("CLASSIFICATION_SERVICE_URL", mockServer.URL)
	
	job := NewClassificationJob(
		"merchant_123",
		"Test Business",
		"Test Description",
		"https://test.com",
		nil, // Supabase client not needed for this test
		cfg,
		zaptest.NewLogger(t),
	)

	ctx := context.Background()
	result, err := job.callClassificationService(ctx)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Technology", result.PrimaryIndustry)
	assert.Equal(t, 0.95, result.ConfidenceScore)
	assert.Equal(t, "low", result.RiskLevel)
	assert.Len(t, result.MCCCodes, 1)
	assert.Equal(t, "5734", result.MCCCodes[0].Code)
}

func TestClassificationJob_callClassificationService_Error(t *testing.T) {
	// Create a mock server that returns an error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockServer.Close()

	cfg := &config.Config{
		Environment: "test",
	}
	
	t.Setenv("CLASSIFICATION_SERVICE_URL", mockServer.URL)
	
	job := NewClassificationJob(
		"merchant_123",
		"Test Business",
		"Test Description",
		"",
		nil,
		cfg,
		zaptest.NewLogger(t),
	)

	ctx := context.Background()
	result, err := job.callClassificationService(ctx)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "classification service returned status 500")
}

func TestClassificationJob_parseIndustryCodesArray(t *testing.T) {
	job := NewClassificationJob(
		"merchant_123",
		"Test Business",
		"Test Description",
		"",
		nil,
		&config.Config{Environment: "test"},
		zaptest.NewLogger(t),
	)

	tests := []struct {
		name     string
		data     []interface{}
		expected []IndustryCode
	}{
		{
			name: "valid industry codes",
			data: []interface{}{
				map[string]interface{}{
					"code":        "5734",
					"description": "Computer Software Stores",
					"confidence":  0.9,
				},
				map[string]interface{}{
					"code":        "7372",
					"description": "Prepackaged Software",
					"confidence":  0.85,
				},
			},
			expected: []IndustryCode{
				{Code: "5734", Description: "Computer Software Stores", Confidence: 0.9},
				{Code: "7372", Description: "Prepackaged Software", Confidence: 0.85},
			},
		},
		{
			name: "codes with confidence_score field",
			data: []interface{}{
				map[string]interface{}{
					"code":            "6012",
					"description":     "Financial Institutions",
					"confidence_score": 0.88,
				},
			},
			expected: []IndustryCode{
				{Code: "6012", Description: "Financial Institutions", Confidence: 0.88},
			},
		},
		{
			name:     "empty array",
			data:     []interface{}{},
			expected: []IndustryCode{},
		},
		{
			name: "invalid code (missing code field)",
			data: []interface{}{
				map[string]interface{}{
					"description": "Missing Code",
					"confidence":  0.5,
				},
			},
			expected: []IndustryCode{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := job.parseIndustryCodesArray(tt.data)
			assert.Equal(t, len(tt.expected), len(result))
			
			for i, expectedCode := range tt.expected {
				assert.Equal(t, expectedCode.Code, result[i].Code)
				assert.Equal(t, expectedCode.Description, result[i].Description)
				assert.Equal(t, expectedCode.Confidence, result[i].Confidence)
			}
		})
	}
}

func TestClassificationJob_getClassificationServiceURL(t *testing.T) {
	tests := []struct {
		name        string
		envVar      string
		environment string
		expected    string
	}{
		{
			name:        "environment variable set",
			envVar:      "https://custom-classification-service.com",
			environment: "production",
			expected:    "https://custom-classification-service.com",
		},
		{
			name:        "development environment",
			envVar:      "",
			environment: "development",
			expected:    "http://localhost:8081",
		},
		{
			name:        "production environment default",
			envVar:      "",
			environment: "production",
			expected:    "https://classification-service-production.up.railway.app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				t.Setenv("CLASSIFICATION_SERVICE_URL", tt.envVar)
			} else {
				t.Setenv("CLASSIFICATION_SERVICE_URL", "")
			}

			job := NewClassificationJob(
				"merchant_123",
				"Test Business",
				"Test Description",
				"",
				nil,
				&config.Config{Environment: tt.environment},
				zaptest.NewLogger(t),
			)

			url := job.getClassificationServiceURL()
			assert.Equal(t, tt.expected, url)
		})
	}
}

