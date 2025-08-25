package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/pcraw4d/business-verification/internal/observability"
)

func TestNewDataTransformationHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataTransformationHandler(logger, mockMetrics)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
	assert.Equal(t, mockMetrics, handler.metrics)
	assert.NotNil(t, handler.transformationJobs)
	assert.NotNil(t, handler.schemas)
	assert.Equal(t, 0, handler.jobCounter)
}

func TestDataTransformationHandler_TransformDataHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataTransformationHandler(logger, mockMetrics)

	t.Run("successful transformation", func(t *testing.T) {
		request := DataTransformationRequest{
			BusinessID:         "business_123",
			TransformationType: TransformationTypeDataCleaning,
			Data: map[string]interface{}{
				"business_name": "  Test Company  ",
				"email":         "TEST@EXAMPLE.COM",
				"phone":         "+1-555-123-4567",
			},
			Rules: []DataTransformationRule{
				{
					Field:       "business_name",
					Operation:   TransformationOperationTrim,
					Parameters:  map[string]interface{}{},
					Description: "Trim whitespace from business name",
					Enabled:     true,
					Order:       1,
				},
				{
					Field:       "email",
					Operation:   TransformationOperationToLower,
					Parameters:  map[string]interface{}{},
					Description: "Convert email to lowercase",
					Enabled:     true,
					Order:       2,
				},
			},
			ValidateBefore:  false,
			ValidateAfter:   false,
			IncludeMetadata: true,
		}

		requestJSON, _ := json.Marshal(request)
		req := httptest.NewRequest("POST", "/v1/transform", bytes.NewBuffer(requestJSON))
		w := httptest.NewRecorder()

		handler.TransformDataHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.NotEmpty(t, w.Header().Get("X-Transformation-ID"))
		assert.NotEmpty(t, w.Header().Get("X-Processing-Time"))

		var response DataTransformationResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.NotEmpty(t, response.TransformationID)
		assert.Equal(t, request.BusinessID, response.BusinessID)
		assert.Equal(t, request.TransformationType, response.TransformationType)
		assert.True(t, response.IsSuccessful)
		assert.NotZero(t, response.TransformedAt)
		assert.NotZero(t, response.ProcessingTime)
		assert.NotNil(t, response.Summary)
	})

	t.Run("invalid transformation type", func(t *testing.T) {
		request := DataTransformationRequest{
			BusinessID:         "business_123",
			TransformationType: "invalid_type",
			Data: map[string]interface{}{
				"business_name": "Test Company",
			},
		}

		requestJSON, _ := json.Marshal(request)
		req := httptest.NewRequest("POST", "/v1/transform", bytes.NewBuffer(requestJSON))
		w := httptest.NewRecorder()

		handler.TransformDataHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing data", func(t *testing.T) {
		request := DataTransformationRequest{
			BusinessID:         "business_123",
			TransformationType: TransformationTypeDataCleaning,
			Data:               nil,
		}

		requestJSON, _ := json.Marshal(request)
		req := httptest.NewRequest("POST", "/v1/transform", bytes.NewBuffer(requestJSON))
		w := httptest.NewRecorder()

		handler.TransformDataHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/transform", bytes.NewBufferString("invalid json"))
		w := httptest.NewRecorder()

		handler.TransformDataHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDataTransformationHandler_CreateTransformationJobHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataTransformationHandler(logger, mockMetrics)

	t.Run("successful job creation", func(t *testing.T) {
		request := DataTransformationRequest{
			BusinessID:         "business_123",
			TransformationType: TransformationTypeNormalization,
			Data: map[string]interface{}{
				"phone": "+1-555-123-4567",
			},
			Rules: []DataTransformationRule{
				{
					Field:       "phone",
					Operation:   TransformationOperationFormat,
					Parameters:  map[string]interface{}{"format": "E.164"},
					Description: "Format phone number to E.164",
					Enabled:     true,
					Order:       1,
				},
			},
		}

		requestJSON, _ := json.Marshal(request)
		req := httptest.NewRequest("POST", "/v1/transform/job", bytes.NewBuffer(requestJSON))
		w := httptest.NewRecorder()

		handler.CreateTransformationJobHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.NotEmpty(t, w.Header().Get("X-Job-ID"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.NotEmpty(t, response["job_id"])
		assert.Equal(t, "pending", response["status"])
		assert.NotNil(t, response["created_at"])
		assert.Equal(t, "Transformation job created successfully", response["message"])
	})

	t.Run("invalid request", func(t *testing.T) {
		request := DataTransformationRequest{
			TransformationType: "invalid_type",
			Data:               nil,
		}

		requestJSON, _ := json.Marshal(request)
		req := httptest.NewRequest("POST", "/v1/transform/job", bytes.NewBuffer(requestJSON))
		w := httptest.NewRecorder()

		handler.CreateTransformationJobHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDataTransformationHandler_GetTransformationJobHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataTransformationHandler(logger, mockMetrics)

	// Create a test job
	job := &TransformationJob{
		ID:                 "test_job_123",
		BusinessID:         "business_123",
		TransformationType: TransformationTypeDataCleaning,
		Status:             "completed",
		Progress:           100,
		CreatedAt:          time.Now(),
		CompletedAt:        &time.Time{},
	}

	handler.jobMutex.Lock()
	handler.transformationJobs[job.ID] = job
	handler.jobMutex.Unlock()

	t.Run("existing job", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/transform/job/test_job_123", nil)
		w := httptest.NewRecorder()

		handler.GetTransformationJobHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response TransformationJob
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, job.ID, response.ID)
		assert.Equal(t, job.BusinessID, response.BusinessID)
		assert.Equal(t, job.Status, response.Status)
	})

	t.Run("non-existent job", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/transform/job/non_existent", nil)
		w := httptest.NewRecorder()

		handler.GetTransformationJobHandler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("missing job ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/transform/job/", nil)
		w := httptest.NewRecorder()

		handler.GetTransformationJobHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDataTransformationHandler_ListTransformationJobsHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataTransformationHandler(logger, mockMetrics)

	// Create test jobs
	jobs := []*TransformationJob{
		{
			ID:                 "job_1",
			BusinessID:         "business_123",
			TransformationType: TransformationTypeDataCleaning,
			Status:             "completed",
			CreatedAt:          time.Now(),
		},
		{
			ID:                 "job_2",
			BusinessID:         "business_456",
			TransformationType: TransformationTypeNormalization,
			Status:             "processing",
			CreatedAt:          time.Now(),
		},
	}

	handler.jobMutex.Lock()
	for _, job := range jobs {
		handler.transformationJobs[job.ID] = job
	}
	handler.jobMutex.Unlock()

	t.Run("list all jobs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/transform/jobs", nil)
		w := httptest.NewRecorder()

		handler.ListTransformationJobsHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.NotNil(t, response["jobs"])
		assert.NotNil(t, response["pagination"])
	})

	t.Run("filter by business ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/transform/jobs?business_id=business_123", nil)
		w := httptest.NewRecorder()

		handler.ListTransformationJobsHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		jobs := response["jobs"].([]interface{})
		assert.Len(t, jobs, 1)
	})

	t.Run("filter by status", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/transform/jobs?status=completed", nil)
		w := httptest.NewRecorder()

		handler.ListTransformationJobsHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		jobs := response["jobs"].([]interface{})
		assert.Len(t, jobs, 1)
	})

	t.Run("pagination", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/transform/jobs?page=1&limit=1", nil)
		w := httptest.NewRecorder()

		handler.ListTransformationJobsHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		jobs := response["jobs"].([]interface{})
		assert.Len(t, jobs, 1)

		pagination := response["pagination"].(map[string]interface{})
		assert.Equal(t, float64(1), pagination["page"])
		assert.Equal(t, float64(1), pagination["limit"])
		assert.Equal(t, float64(2), pagination["total"])
	})
}

func TestDataTransformationHandler_GetTransformationSchemaHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataTransformationHandler(logger, mockMetrics)

	t.Run("existing schema", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/transform/schema/data_cleaning_default", nil)
		w := httptest.NewRecorder()

		handler.GetTransformationSchemaHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response TransformationSchema
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "data_cleaning_default", response.ID)
		assert.Equal(t, "Default Data Cleaning", response.Name)
		assert.Equal(t, TransformationTypeDataCleaning, response.Type)
	})

	t.Run("non-existent schema", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/transform/schema/non_existent", nil)
		w := httptest.NewRecorder()

		handler.GetTransformationSchemaHandler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("missing schema ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/transform/schema/", nil)
		w := httptest.NewRecorder()

		handler.GetTransformationSchemaHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDataTransformationHandler_ListTransformationSchemasHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataTransformationHandler(logger, mockMetrics)

	t.Run("list all schemas", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/transform/schemas", nil)
		w := httptest.NewRecorder()

		handler.ListTransformationSchemasHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.NotNil(t, response["schemas"])
		assert.Equal(t, float64(2), response["total"]) // 2 default schemas
	})

	t.Run("filter by type", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/transform/schemas?type=data_cleaning", nil)
		w := httptest.NewRecorder()

		handler.ListTransformationSchemasHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		schemas := response["schemas"].([]interface{})
		assert.Len(t, schemas, 1)
	})
}

func TestDataTransformationHandler_ValidationFunctions(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataTransformationHandler(logger, mockMetrics)

	t.Run("valid transformation type", func(t *testing.T) {
		assert.True(t, handler.isValidTransformationType(TransformationTypeDataCleaning))
		assert.True(t, handler.isValidTransformationType(TransformationTypeNormalization))
		assert.True(t, handler.isValidTransformationType(TransformationTypeEnrichment))
		assert.True(t, handler.isValidTransformationType(TransformationTypeAll))
	})

	t.Run("invalid transformation type", func(t *testing.T) {
		assert.False(t, handler.isValidTransformationType("invalid_type"))
		assert.False(t, handler.isValidTransformationType(""))
	})

	t.Run("validate transformation request", func(t *testing.T) {
		validRequest := &DataTransformationRequest{
			TransformationType: TransformationTypeDataCleaning,
			Data:               map[string]interface{}{"test": "data"},
		}

		err := handler.validateTransformationRequest(validRequest)
		assert.NoError(t, err)
	})

	t.Run("validate transformation request - missing type", func(t *testing.T) {
		invalidRequest := &DataTransformationRequest{
			Data: map[string]interface{}{"test": "data"},
		}

		err := handler.validateTransformationRequest(invalidRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transformation_type is required")
	})

	t.Run("validate transformation request - missing data", func(t *testing.T) {
		invalidRequest := &DataTransformationRequest{
			TransformationType: TransformationTypeDataCleaning,
			Data:               nil,
		}

		err := handler.validateTransformationRequest(invalidRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "data is required")
	})

	t.Run("validate transformation request - invalid type", func(t *testing.T) {
		invalidRequest := &DataTransformationRequest{
			TransformationType: "invalid_type",
			Data:               map[string]interface{}{"test": "data"},
		}

		err := handler.validateTransformationRequest(invalidRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid transformation_type")
	})
}

func TestDataTransformationHandler_TransformationFunctions(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataTransformationHandler(logger, mockMetrics)

	t.Run("transform data", func(t *testing.T) {
		req := &DataTransformationRequest{
			BusinessID:         "business_123",
			TransformationType: TransformationTypeDataCleaning,
			Data: map[string]interface{}{
				"business_name": "Test Company",
			},
			Rules: []DataTransformationRule{
				{
					Field:       "business_name",
					Operation:   TransformationOperationTrim,
					Parameters:  map[string]interface{}{},
					Description: "Trim whitespace",
					Enabled:     true,
					Order:       1,
				},
			},
		}

		result, err := handler.transformData(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.TransformationID)
		assert.Equal(t, req.BusinessID, result.BusinessID)
		assert.Equal(t, req.TransformationType, result.TransformationType)
		assert.True(t, result.IsSuccessful)
		assert.NotZero(t, result.TransformedAt)
		assert.NotZero(t, result.ProcessingTime)
	})

	t.Run("create transformation job", func(t *testing.T) {
		req := &DataTransformationRequest{
			BusinessID:         "business_123",
			TransformationType: TransformationTypeNormalization,
			Data:               map[string]interface{}{"test": "data"},
		}

		job, err := handler.createTransformationJob(req)
		assert.NoError(t, err)
		assert.NotNil(t, job)
		assert.NotEmpty(t, job.ID)
		assert.Equal(t, req.BusinessID, job.BusinessID)
		assert.Equal(t, req.TransformationType, job.TransformationType)
		assert.Equal(t, "pending", job.Status)
		assert.Equal(t, 0, job.Progress)
		assert.NotZero(t, job.CreatedAt)
	})

	t.Run("apply transformation rule", func(t *testing.T) {
		rule := DataTransformationRule{
			Field:       "test_field",
			Operation:   TransformationOperationTrim,
			Parameters:  map[string]interface{}{},
			Description: "Test transformation",
			Enabled:     true,
			Order:       1,
		}

		data := map[string]interface{}{"test_field": "  test value  "}

		result, err := handler.applyTransformationRule(data, rule)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("unsupported transformation operation", func(t *testing.T) {
		rule := DataTransformationRule{
			Field:       "test_field",
			Operation:   "unsupported_operation",
			Parameters:  map[string]interface{}{},
			Description: "Test transformation",
			Enabled:     true,
			Order:       1,
		}

		data := map[string]interface{}{"test_field": "test value"}

		result, err := handler.applyTransformationRule(data, rule)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported transformation operation")
		assert.Equal(t, data, result)
	})
}

func TestDataTransformationHandler_UtilityFunctions(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataTransformationHandler(logger, mockMetrics)

	t.Run("generate transformation ID", func(t *testing.T) {
		id1 := handler.generateTransformationID()
		id2 := handler.generateTransformationID()

		assert.NotEmpty(t, id1)
		assert.NotEmpty(t, id2)
		assert.NotEqual(t, id1, id2)
		assert.Contains(t, id1, "transform_")
	})

	t.Run("evaluate condition", func(t *testing.T) {
		data := map[string]interface{}{"test": "value"}
		condition := "test == 'value'"

		result := handler.evaluateCondition(data, condition)
		assert.True(t, result) // Mock implementation always returns true
	})

	t.Run("pre-transformation validation", func(t *testing.T) {
		data := map[string]interface{}{"test": "value"}

		result := handler.performPreTransformationValidation(data)
		assert.NotNil(t, result)

		validation, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, true, validation["valid"])
	})

	t.Run("post-transformation validation", func(t *testing.T) {
		data := map[string]interface{}{"test": "value"}

		result := handler.performPostTransformationValidation(data)
		assert.NotNil(t, result)

		validation, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, true, validation["valid"])
	})

	t.Run("extract path param", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/transform/job/test_job_123", nil)

		jobID := extractPathParam(req, "job_id")
		assert.Equal(t, "test_job_123", jobID)

		nonExistent := extractPathParam(req, "non_existent")
		assert.Equal(t, "", nonExistent)
	})
}

func TestDataTransformationHandler_DefaultSchemas(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataTransformationHandler(logger, mockMetrics)

	t.Run("default schemas initialized", func(t *testing.T) {
		handler.schemaMutex.RLock()
		defer handler.schemaMutex.RUnlock()

		// Check data cleaning schema
		dataCleaningSchema, exists := handler.schemas["data_cleaning_default"]
		assert.True(t, exists)
		assert.Equal(t, "Default Data Cleaning", dataCleaningSchema.Name)
		assert.Equal(t, TransformationTypeDataCleaning, dataCleaningSchema.Type)
		assert.Len(t, dataCleaningSchema.Rules, 2)

		// Check normalization schema
		normalizationSchema, exists := handler.schemas["normalization_default"]
		assert.True(t, exists)
		assert.Equal(t, "Default Normalization", normalizationSchema.Name)
		assert.Equal(t, TransformationTypeNormalization, normalizationSchema.Type)
		assert.Len(t, normalizationSchema.Rules, 1)
	})

	t.Run("schema rules validation", func(t *testing.T) {
		handler.schemaMutex.RLock()
		dataCleaningSchema := handler.schemas["data_cleaning_default"]
		handler.schemaMutex.RUnlock()

		// Check first rule
		firstRule := dataCleaningSchema.Rules[0]
		assert.Equal(t, "business_name", firstRule.Field)
		assert.Equal(t, TransformationOperationTrim, firstRule.Operation)
		assert.True(t, firstRule.Enabled)
		assert.Equal(t, 1, firstRule.Order)

		// Check second rule
		secondRule := dataCleaningSchema.Rules[1]
		assert.Equal(t, "email", secondRule.Field)
		assert.Equal(t, TransformationOperationToLower, secondRule.Operation)
		assert.True(t, secondRule.Enabled)
		assert.Equal(t, 2, secondRule.Order)
	})
}
