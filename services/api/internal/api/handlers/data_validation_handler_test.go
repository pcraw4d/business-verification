package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewDataValidationHandler(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataValidationHandler(logger)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
	assert.NotNil(t, handler.validations)
	assert.NotNil(t, handler.jobs)
	assert.NotNil(t, handler.reports)
	assert.Len(t, handler.validations, 0)
	assert.Len(t, handler.jobs, 0)
	assert.Len(t, handler.reports, 0)
}

func TestDataValidationHandler_CreateValidation(t *testing.T) {
	tests := []struct {
		name           string
		request        DataValidationRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful validation creation",
			request: DataValidationRequest{
				Name:        "Test Validation",
				Description: "Test data validation",
				Dataset:     "customer_data",
				Data: map[string]interface{}{
					"name":  "John Doe",
					"email": "john@example.com",
					"age":   30,
				},
				Schema: &ValidationSchema{
					Type:    "object",
					Version: "1.0",
					Properties: map[string]SchemaProperty{
						"name": {
							Type:        "string",
							Description: "Customer name",
							Required:    true,
							MinLength:   2,
							MaxLength:   100,
						},
						"email": {
							Type:        "string",
							Description: "Customer email",
							Required:    true,
							Format:      "email",
						},
						"age": {
							Type:        "integer",
							Description: "Customer age",
							Required:    true,
							MinValue:    18,
							MaxValue:    120,
						},
					},
					Required: []string{"name", "email", "age"},
				},
				Rules: []ValidationRule{
					{
						Name:        "email_format_rule",
						Type:        ValidationTypeFormat,
						Description: "Validate email format",
						Severity:    ValidationSeverityHigh,
						Expression:  "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
						Parameters: map[string]interface{}{
							"field": "email",
						},
						Enabled: true,
					},
					{
						Name:        "age_business_rule",
						Type:        ValidationTypeBusiness,
						Description: "Validate age is 18 or older",
						Severity:    ValidationSeverityCritical,
						Expression:  "age >= 18",
						Parameters: map[string]interface{}{
							"field": "age",
						},
						Enabled: true,
					},
				},
				Validators: []CustomValidator{
					{
						Name:        "custom_email_validator",
						Description: "Custom email validation",
						Type:        "javascript",
						Code:        "function validate(data) { return data.email.includes('@'); }",
						Language:    "javascript",
						Parameters:  map[string]interface{}{},
						Timeout:     5 * time.Second,
						Enabled:     true,
					},
				},
				Options: ValidationOptions{
					StopOnFirstError: false,
					ContinueOnError:  true,
					MaxErrors:        100,
					Timeout:          30 * time.Second,
					Parallel:         true,
					BatchSize:        1000,
					CacheResults:     true,
					LogLevel:         "info",
					Custom:           map[string]interface{}{},
				},
				Metadata: map[string]interface{}{
					"department": "data_team",
					"priority":   "high",
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			request: DataValidationRequest{
				Description: "Test data validation",
				Dataset:     "customer_data",
				Data:        map[string]interface{}{},
				Rules: []ValidationRule{
					{
						Name:        "test_rule",
						Type:        ValidationTypeRule,
						Description: "Test rule",
						Severity:    ValidationSeverityMedium,
						Expression:  "true",
						Parameters:  map[string]interface{}{},
						Enabled:     true,
					},
				},
				Options: ValidationOptions{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name is required",
		},
		{
			name: "missing dataset",
			request: DataValidationRequest{
				Name:        "Test Validation",
				Description: "Test data validation",
				Data:        map[string]interface{}{},
				Rules: []ValidationRule{
					{
						Name:        "test_rule",
						Type:        ValidationTypeRule,
						Description: "Test rule",
						Severity:    ValidationSeverityMedium,
						Expression:  "true",
						Parameters:  map[string]interface{}{},
						Enabled:     true,
					},
				},
				Options: ValidationOptions{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "dataset is required",
		},
		{
			name: "missing data",
			request: DataValidationRequest{
				Name:        "Test Validation",
				Description: "Test data validation",
				Dataset:     "customer_data",
				Rules: []ValidationRule{
					{
						Name:        "test_rule",
						Type:        ValidationTypeRule,
						Description: "Test rule",
						Severity:    ValidationSeverityMedium,
						Expression:  "true",
						Parameters:  map[string]interface{}{},
						Enabled:     true,
					},
				},
				Options: ValidationOptions{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "data is required",
		},
		{
			name: "missing rules",
			request: DataValidationRequest{
				Name:        "Test Validation",
				Description: "Test data validation",
				Dataset:     "customer_data",
				Data:        map[string]interface{}{},
				Rules:       []ValidationRule{},
				Options:     ValidationOptions{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "at least one validation rule is required",
		},
		{
			name: "invalid rule - missing name",
			request: DataValidationRequest{
				Name:        "Test Validation",
				Description: "Test data validation",
				Dataset:     "customer_data",
				Data:        map[string]interface{}{},
				Rules: []ValidationRule{
					{
						Type:        ValidationTypeRule,
						Description: "Test rule",
						Severity:    ValidationSeverityMedium,
						Expression:  "true",
						Parameters:  map[string]interface{}{},
						Enabled:     true,
					},
				},
				Options: ValidationOptions{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "rule 1: name is required",
		},
		{
			name: "invalid rule - missing type",
			request: DataValidationRequest{
				Name:        "Test Validation",
				Description: "Test data validation",
				Dataset:     "customer_data",
				Data:        map[string]interface{}{},
				Rules: []ValidationRule{
					{
						Name:        "test_rule",
						Description: "Test rule",
						Severity:    ValidationSeverityMedium,
						Expression:  "true",
						Parameters:  map[string]interface{}{},
						Enabled:     true,
					},
				},
				Options: ValidationOptions{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "rule 1: type is required",
		},
		{
			name: "invalid rule - missing severity",
			request: DataValidationRequest{
				Name:        "Test Validation",
				Description: "Test data validation",
				Dataset:     "customer_data",
				Data:        map[string]interface{}{},
				Rules: []ValidationRule{
					{
						Name:        "test_rule",
						Type:        ValidationTypeRule,
						Description: "Test rule",
						Expression:  "true",
						Parameters:  map[string]interface{}{},
						Enabled:     true,
					},
				},
				Options: ValidationOptions{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "rule 1: severity is required",
		},
		{
			name: "invalid rule - missing expression",
			request: DataValidationRequest{
				Name:        "Test Validation",
				Description: "Test data validation",
				Dataset:     "customer_data",
				Data:        map[string]interface{}{},
				Rules: []ValidationRule{
					{
						Name:        "test_rule",
						Type:        ValidationTypeRule,
						Description: "Test rule",
						Severity:    ValidationSeverityMedium,
						Parameters:  map[string]interface{}{},
						Enabled:     true,
					},
				},
				Options: ValidationOptions{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "rule 1: expression is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataValidationHandler(zap.NewNop())

			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/validation", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.CreateValidation(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response DataValidationResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.NotEmpty(t, response.ID)
				assert.Equal(t, tt.request.Name, response.Name)
				assert.Equal(t, "completed", response.Status)
				assert.Greater(t, response.OverallScore, 0.0)
				assert.NotEmpty(t, response.Validations)
				assert.NotNil(t, response.Summary)
				assert.Equal(t, tt.request.Metadata, response.Metadata)
			}
		})
	}
}

func TestDataValidationHandler_GetValidation(t *testing.T) {
	handler := NewDataValidationHandler(zap.NewNop())

	// Create a test validation
	request := DataValidationRequest{
		Name:        "Test Validation",
		Description: "Test data validation",
		Dataset:     "customer_data",
		Data:        map[string]interface{}{"test": "data"},
		Rules: []ValidationRule{
			{
				Name:        "test_rule",
				Type:        ValidationTypeRule,
				Description: "Test rule",
				Severity:    ValidationSeverityMedium,
				Expression:  "true",
				Parameters:  map[string]interface{}{},
				Enabled:     true,
			},
		},
		Options: ValidationOptions{},
	}

	body, err := json.Marshal(request)
	require.NoError(t, err)

	createReq := httptest.NewRequest("POST", "/validation", bytes.NewBuffer(body))
	createW := httptest.NewRecorder()
	handler.CreateValidation(createW, createReq)

	var createResponse DataValidationResponse
	err = json.Unmarshal(createW.Body.Bytes(), &createResponse)
	require.NoError(t, err)

	tests := []struct {
		name           string
		id             string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful validation retrieval",
			id:             createResponse.ID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing id",
			id:             "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Validation ID is required",
		},
		{
			name:           "validation not found",
			id:             "non_existent_id",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Validation not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/validation"
			if tt.id != "" {
				url = fmt.Sprintf("/validation?id=%s", tt.id)
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.GetValidation(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response DataValidationResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, createResponse.ID, response.ID)
				assert.Equal(t, createResponse.Name, response.Name)
			}
		})
	}
}

func TestDataValidationHandler_ListValidations(t *testing.T) {
	handler := NewDataValidationHandler(zap.NewNop())

	// Create test validations
	request := DataValidationRequest{
		Name:        "Test Validation",
		Description: "Test data validation",
		Dataset:     "customer_data",
		Data:        map[string]interface{}{"test": "data"},
		Rules: []ValidationRule{
			{
				Name:        "test_rule",
				Type:        ValidationTypeRule,
				Description: "Test rule",
				Severity:    ValidationSeverityMedium,
				Expression:  "true",
				Parameters:  map[string]interface{}{},
				Enabled:     true,
			},
		},
		Options: ValidationOptions{},
	}

	body, err := json.Marshal(request)
	require.NoError(t, err)

	// Create first validation
	createReq1 := httptest.NewRequest("POST", "/validation", bytes.NewBuffer(body))
	createW1 := httptest.NewRecorder()
	handler.CreateValidation(createW1, createReq1)

	// Create second validation
	createReq2 := httptest.NewRequest("POST", "/validation", bytes.NewBuffer(body))
	createW2 := httptest.NewRecorder()
	handler.CreateValidation(createW2, createReq2)

	// Test list validations
	req := httptest.NewRequest("GET", "/validation", nil)
	w := httptest.NewRecorder()

	handler.ListValidations(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response["validations"])
	assert.Equal(t, float64(2), response["total"])
}

func TestDataValidationHandler_CreateValidationJob(t *testing.T) {
	tests := []struct {
		name           string
		request        DataValidationRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful job creation",
			request: DataValidationRequest{
				Name:        "Test Validation Job",
				Description: "Test data validation job",
				Dataset:     "customer_data",
				Data:        map[string]interface{}{"test": "data"},
				Rules: []ValidationRule{
					{
						Name:        "test_rule",
						Type:        ValidationTypeRule,
						Description: "Test rule",
						Severity:    ValidationSeverityMedium,
						Expression:  "true",
						Parameters:  map[string]interface{}{},
						Enabled:     true,
					},
				},
				Options: ValidationOptions{},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request - missing name",
			request: DataValidationRequest{
				Description: "Test data validation job",
				Dataset:     "customer_data",
				Data:        map[string]interface{}{},
				Rules: []ValidationRule{
					{
						Name:        "test_rule",
						Type:        ValidationTypeRule,
						Description: "Test rule",
						Severity:    ValidationSeverityMedium,
						Expression:  "true",
						Parameters:  map[string]interface{}{},
						Enabled:     true,
					},
				},
				Options: ValidationOptions{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataValidationHandler(zap.NewNop())

			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/validation/jobs", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.CreateValidationJob(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response ValidationJob
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.NotEmpty(t, response.ID)
				assert.NotEmpty(t, response.RequestID)
				assert.Equal(t, "pending", response.Status)
				assert.Equal(t, 0, response.Progress)
				assert.Equal(t, tt.request.Metadata, response.Metadata)
			}
		})
	}
}

func TestDataValidationHandler_GetValidationJob(t *testing.T) {
	handler := NewDataValidationHandler(zap.NewNop())

	// Create a test job
	request := DataValidationRequest{
		Name:        "Test Validation Job",
		Description: "Test data validation job",
		Dataset:     "customer_data",
		Data:        map[string]interface{}{"test": "data"},
		Rules: []ValidationRule{
			{
				Name:        "test_rule",
				Type:        ValidationTypeRule,
				Description: "Test rule",
				Severity:    ValidationSeverityMedium,
				Expression:  "true",
				Parameters:  map[string]interface{}{},
				Enabled:     true,
			},
		},
		Options: ValidationOptions{},
	}

	body, err := json.Marshal(request)
	require.NoError(t, err)

	createReq := httptest.NewRequest("POST", "/validation/jobs", bytes.NewBuffer(body))
	createW := httptest.NewRecorder()
	handler.CreateValidationJob(createW, createReq)

	var createResponse ValidationJob
	err = json.Unmarshal(createW.Body.Bytes(), &createResponse)
	require.NoError(t, err)

	tests := []struct {
		name           string
		id             string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful job retrieval",
			id:             createResponse.ID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing id",
			id:             "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Job ID is required",
		},
		{
			name:           "job not found",
			id:             "non_existent_id",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Job not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/validation/jobs"
			if tt.id != "" {
				url = fmt.Sprintf("/validation/jobs?id=%s", tt.id)
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.GetValidationJob(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response ValidationJob
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, createResponse.ID, response.ID)
				assert.Equal(t, createResponse.RequestID, response.RequestID)
			}
		})
	}
}

func TestDataValidationHandler_ListValidationJobs(t *testing.T) {
	handler := NewDataValidationHandler(zap.NewNop())

	// Create test jobs
	request := DataValidationRequest{
		Name:        "Test Validation Job",
		Description: "Test data validation job",
		Dataset:     "customer_data",
		Data:        map[string]interface{}{"test": "data"},
		Rules: []ValidationRule{
			{
				Name:        "test_rule",
				Type:        ValidationTypeRule,
				Description: "Test rule",
				Severity:    ValidationSeverityMedium,
				Expression:  "true",
				Parameters:  map[string]interface{}{},
				Enabled:     true,
			},
		},
		Options: ValidationOptions{},
	}

	body, err := json.Marshal(request)
	require.NoError(t, err)

	// Create first job
	createReq1 := httptest.NewRequest("POST", "/validation/jobs", bytes.NewBuffer(body))
	createW1 := httptest.NewRecorder()
	handler.CreateValidationJob(createW1, createReq1)

	// Create second job
	createReq2 := httptest.NewRequest("POST", "/validation/jobs", bytes.NewBuffer(body))
	createW2 := httptest.NewRecorder()
	handler.CreateValidationJob(createW2, createReq2)

	// Test list jobs
	req := httptest.NewRequest("GET", "/validation/jobs", nil)
	w := httptest.NewRecorder()

	handler.ListValidationJobs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response["jobs"])
	assert.Equal(t, float64(2), response["total"])
}

func TestDataValidationHandler_ValidationLogic(t *testing.T) {
	handler := NewDataValidationHandler(zap.NewNop())

	tests := []struct {
		name           string
		request        DataValidationRequest
		expectedScore  float64
		expectedStatus ValidationStatus
	}{
		{
			name: "schema validation",
			request: DataValidationRequest{
				Name:        "Schema Validation",
				Description: "Test schema validation",
				Dataset:     "customer_data",
				Data:        map[string]interface{}{"test": "data"},
				Schema: &ValidationSchema{
					Type:    "object",
					Version: "1.0",
					Properties: map[string]SchemaProperty{
						"test": {
							Type:        "string",
							Description: "Test field",
							Required:    true,
						},
					},
				},
				Rules: []ValidationRule{
					{
						Name:        "test_rule",
						Type:        ValidationTypeSchema,
						Description: "Test rule",
						Severity:    ValidationSeverityHigh,
						Expression:  "true",
						Parameters:  map[string]interface{}{},
						Enabled:     true,
					},
				},
				Options: ValidationOptions{},
			},
			expectedScore:  0.95,
			expectedStatus: ValidationStatusPassed,
		},
		{
			name: "rule validation",
			request: DataValidationRequest{
				Name:        "Rule Validation",
				Description: "Test rule validation",
				Dataset:     "customer_data",
				Data:        map[string]interface{}{"test": "data"},
				Rules: []ValidationRule{
					{
						Name:        "test_rule",
						Type:        ValidationTypeRule,
						Description: "Test rule",
						Severity:    ValidationSeverityMedium,
						Expression:  "true",
						Parameters:  map[string]interface{}{},
						Enabled:     true,
					},
				},
				Options: ValidationOptions{},
			},
			expectedScore:  0.92,
			expectedStatus: ValidationStatusPassed,
		},
		{
			name: "custom validation",
			request: DataValidationRequest{
				Name:        "Custom Validation",
				Description: "Test custom validation",
				Dataset:     "customer_data",
				Data:        map[string]interface{}{"test": "data"},
				Rules: []ValidationRule{
					{
						Name:        "test_rule",
						Type:        ValidationTypeCustom,
						Description: "Test rule",
						Severity:    ValidationSeverityMedium,
						Expression:  "true",
						Parameters:  map[string]interface{}{},
						Enabled:     true,
					},
				},
				Options: ValidationOptions{},
			},
			expectedScore:  0.88,
			expectedStatus: ValidationStatusWarning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation request validation
			err := handler.validateValidationRequest(tt.request)
			assert.NoError(t, err)

			// Test overall score calculation
			score := handler.calculateOverallScore(tt.request)
			assert.Greater(t, score, 0.0)

			// Test validation performance
			results := handler.performValidations(tt.request)
			assert.NotEmpty(t, results)

			// Test summary generation
			summary := handler.generateValidationSummary(tt.request)
			assert.NotNil(t, summary)
			assert.Greater(t, summary.TotalValidations, 0)
		})
	}
}

func TestDataValidationHandler_UtilityFunctions(t *testing.T) {
	handler := NewDataValidationHandler(zap.NewNop())

	// Test severity weight calculation
	tests := []struct {
		severity ValidationSeverity
		expected float64
	}{
		{ValidationSeverityCritical, 4.0},
		{ValidationSeverityHigh, 3.0},
		{ValidationSeverityMedium, 2.0},
		{ValidationSeverityLow, 1.0},
	}

	for _, tt := range tests {
		weight := handler.getSeverityWeight(tt.severity)
		assert.Equal(t, tt.expected, weight)
	}

	// Test validation score simulation
	rule := ValidationRule{Type: ValidationTypeSchema}
	score := handler.simulateValidationScore(rule)
	assert.Greater(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)

	// Test validation status determination
	status := handler.determineValidationStatus(ValidationTypeSchema)
	assert.Contains(t, []ValidationStatus{
		ValidationStatusPassed,
		ValidationStatusWarning,
		ValidationStatusFailed,
		ValidationStatusError,
	}, status)
}

func TestDataValidationHandler_StringConversions(t *testing.T) {
	// Test ValidationType string conversion
	assert.Equal(t, "schema", ValidationTypeSchema.String())
	assert.Equal(t, "rule", ValidationTypeRule.String())
	assert.Equal(t, "custom", ValidationTypeCustom.String())

	// Test ValidationStatus string conversion
	assert.Equal(t, "passed", ValidationStatusPassed.String())
	assert.Equal(t, "failed", ValidationStatusFailed.String())
	assert.Equal(t, "warning", ValidationStatusWarning.String())
	assert.Equal(t, "error", ValidationStatusError.String())

	// Test ValidationSeverity string conversion
	assert.Equal(t, "low", ValidationSeverityLow.String())
	assert.Equal(t, "medium", ValidationSeverityMedium.String())
	assert.Equal(t, "high", ValidationSeverityHigh.String())
	assert.Equal(t, "critical", ValidationSeverityCritical.String())
}

func TestDataValidationHandler_ErrorGeneration(t *testing.T) {
	handler := NewDataValidationHandler(zap.NewNop())

	// Test schema error generation
	req := DataValidationRequest{
		Name:        "Test",
		Description: "Test",
		Dataset:     "test",
		Data:        map[string]interface{}{},
		Schema: &ValidationSchema{
			Type:    "object",
			Version: "1.0",
			Properties: map[string]SchemaProperty{
				"required_field": {
					Type:     "string",
					Required: true,
				},
			},
		},
		Rules: []ValidationRule{
			{
				Name:        "test_rule",
				Type:        ValidationTypeRule,
				Description: "Test rule",
				Severity:    ValidationSeverityMedium,
				Expression:  "true",
				Parameters:  map[string]interface{}{},
				Enabled:     true,
			},
		},
		Options: ValidationOptions{},
	}

	errors := handler.generateSchemaErrors(req)
	assert.NotEmpty(t, errors)

	warnings := handler.generateSchemaWarnings(req)
	assert.NotEmpty(t, warnings)

	metrics := handler.generateSchemaMetrics(req)
	assert.NotNil(t, metrics)

	// Test rule error generation
	rule := ValidationRule{
		Name:        "test_rule",
		Type:        ValidationTypeFormat,
		Description: "Test rule",
		Severity:    ValidationSeverityHigh,
		Expression:  "test",
		Parameters:  map[string]interface{}{},
		Enabled:     true,
	}

	ruleErrors := handler.generateRuleErrors(rule)
	assert.NotEmpty(t, ruleErrors)

	ruleWarnings := handler.generateRuleWarnings(rule)
	assert.NotEmpty(t, ruleWarnings)

	ruleMetrics := handler.generateRuleMetrics(rule)
	assert.NotNil(t, ruleMetrics)

	// Test custom error generation
	validator := CustomValidator{
		Name:        "test_validator",
		Description: "Test validator",
		Type:        "javascript",
		Code:        "function test() {}",
		Language:    "javascript",
		Parameters:  map[string]interface{}{},
		Timeout:     5 * time.Second,
		Enabled:     true,
	}

	customErrors := handler.generateCustomErrors(validator)
	assert.NotEmpty(t, customErrors)

	customWarnings := handler.generateCustomWarnings(validator)
	assert.NotEmpty(t, customWarnings)

	customMetrics := handler.generateCustomMetrics(validator)
	assert.NotNil(t, customMetrics)
}
