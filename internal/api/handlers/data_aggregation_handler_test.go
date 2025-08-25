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

func TestNewDataAggregationHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataAggregationHandler(logger, mockMetrics)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
	assert.Equal(t, mockMetrics, handler.metrics)
	assert.NotNil(t, handler.aggregationJobs)
	assert.NotNil(t, handler.schemas)
	assert.Equal(t, 0, handler.jobCounter)
}

func TestDataAggregationHandler_AggregateDataHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataAggregationHandler(logger, mockMetrics)

	t.Run("successful aggregation", func(t *testing.T) {
		request := DataAggregationRequest{
			BusinessID:      "business_123",
			AggregationType: AggregationTypeBusinessMetrics,
			Data: map[string]interface{}{
				"verifications": []map[string]interface{}{
					{"status": "passed", "score": 0.95},
					{"status": "failed", "score": 0.30},
					{"status": "passed", "score": 0.88},
				},
			},
			Rules: []DataAggregationRule{
				{
					Field:       "verification_count",
					Operation:   AggregationOperationCount,
					Parameters:  map[string]interface{}{},
					Description: "Count total verifications",
					Enabled:     true,
					Order:       1,
				},
				{
					Field:       "success_rate",
					Operation:   AggregationOperationAverage,
					Parameters:  map[string]interface{}{},
					Description: "Calculate average success rate",
					Enabled:     true,
					Order:       2,
				},
			},
			IncludeMetadata: true,
		}

		requestJSON, _ := json.Marshal(request)
		req := httptest.NewRequest("POST", "/v1/aggregate", bytes.NewBuffer(requestJSON))
		w := httptest.NewRecorder()

		handler.AggregateDataHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.NotEmpty(t, w.Header().Get("X-Aggregation-ID"))
		assert.NotEmpty(t, w.Header().Get("X-Processing-Time"))

		var response DataAggregationResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.NotEmpty(t, response.AggregationID)
		assert.Equal(t, request.BusinessID, response.BusinessID)
		assert.Equal(t, request.AggregationType, response.AggregationType)
		assert.True(t, response.IsSuccessful)
		assert.NotZero(t, response.AggregatedAt)
		assert.NotZero(t, response.ProcessingTime)
		assert.NotNil(t, response.Summary)
	})

	t.Run("invalid aggregation type", func(t *testing.T) {
		request := DataAggregationRequest{
			AggregationType: "invalid_type",
			Data:            map[string]interface{}{"test": "data"},
		}

		requestJSON, _ := json.Marshal(request)
		req := httptest.NewRequest("POST", "/v1/aggregate", bytes.NewBuffer(requestJSON))
		w := httptest.NewRecorder()

		handler.AggregateDataHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing data", func(t *testing.T) {
		request := DataAggregationRequest{
			AggregationType: AggregationTypeBusinessMetrics,
			Data:            nil,
		}

		requestJSON, _ := json.Marshal(request)
		req := httptest.NewRequest("POST", "/v1/aggregate", bytes.NewBuffer(requestJSON))
		w := httptest.NewRecorder()

		handler.AggregateDataHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/aggregate", bytes.NewBufferString("invalid json"))
		w := httptest.NewRecorder()

		handler.AggregateDataHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDataAggregationHandler_CreateAggregationJobHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataAggregationHandler(logger, mockMetrics)

	t.Run("successful job creation", func(t *testing.T) {
		request := DataAggregationRequest{
			BusinessID:      "business_123",
			AggregationType: AggregationTypeBusinessMetrics,
			Data: map[string]interface{}{
				"verifications": []map[string]interface{}{
					{"status": "passed", "score": 0.95},
					{"status": "failed", "score": 0.30},
				},
			},
			Rules: []DataAggregationRule{
				{
					Field:       "verification_count",
					Operation:   AggregationOperationCount,
					Parameters:  map[string]interface{}{},
					Description: "Count total verifications",
					Enabled:     true,
					Order:       1,
				},
			},
		}

		requestJSON, _ := json.Marshal(request)
		req := httptest.NewRequest("POST", "/v1/aggregate/job", bytes.NewBuffer(requestJSON))
		w := httptest.NewRecorder()

		handler.CreateAggregationJobHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.NotEmpty(t, w.Header().Get("X-Job-ID"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.NotEmpty(t, response["job_id"])
		assert.Equal(t, "pending", response["status"])
		assert.NotNil(t, response["created_at"])
		assert.Equal(t, "Aggregation job created successfully", response["message"])
	})

	t.Run("invalid request", func(t *testing.T) {
		request := DataAggregationRequest{
			AggregationType: "invalid_type",
			Data:            nil,
		}

		requestJSON, _ := json.Marshal(request)
		req := httptest.NewRequest("POST", "/v1/aggregate/job", bytes.NewBuffer(requestJSON))
		w := httptest.NewRecorder()

		handler.CreateAggregationJobHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDataAggregationHandler_GetAggregationJobHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataAggregationHandler(logger, mockMetrics)

	// Create a test job
	job := &AggregationJob{
		ID:              "test_job_123",
		BusinessID:      "business_123",
		AggregationType: AggregationTypeBusinessMetrics,
		Status:          "completed",
		Progress:        100,
		CreatedAt:       time.Now(),
		CompletedAt:     &time.Time{},
	}

	handler.jobMutex.Lock()
	handler.aggregationJobs[job.ID] = job
	handler.jobMutex.Unlock()

	t.Run("existing job", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/aggregate/job/test_job_123", nil)
		w := httptest.NewRecorder()

		handler.GetAggregationJobHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response AggregationJob
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, job.ID, response.ID)
		assert.Equal(t, job.BusinessID, response.BusinessID)
		assert.Equal(t, job.Status, response.Status)
	})

	t.Run("non-existent job", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/aggregate/job/non_existent", nil)
		w := httptest.NewRecorder()

		handler.GetAggregationJobHandler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("missing job ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/aggregate/job/", nil)
		w := httptest.NewRecorder()

		handler.GetAggregationJobHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDataAggregationHandler_ListAggregationJobsHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataAggregationHandler(logger, mockMetrics)

	// Create test jobs
	jobs := []*AggregationJob{
		{
			ID:              "job_1",
			BusinessID:      "business_123",
			AggregationType: AggregationTypeBusinessMetrics,
			Status:          "completed",
			CreatedAt:       time.Now(),
		},
		{
			ID:              "job_2",
			BusinessID:      "business_456",
			AggregationType: AggregationTypeRiskAssessments,
			Status:          "pending",
			CreatedAt:       time.Now(),
		},
	}

	handler.jobMutex.Lock()
	for _, job := range jobs {
		handler.aggregationJobs[job.ID] = job
	}
	handler.jobMutex.Unlock()

	t.Run("list all jobs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/aggregate/jobs", nil)
		w := httptest.NewRecorder()

		handler.ListAggregationJobsHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.NotNil(t, response["jobs"])
		assert.NotNil(t, response["pagination"])
	})

	t.Run("filter by business ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/aggregate/jobs?business_id=business_123", nil)
		w := httptest.NewRecorder()

		handler.ListAggregationJobsHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		jobs := response["jobs"].([]interface{})
		assert.Len(t, jobs, 1)
	})

	t.Run("filter by status", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/aggregate/jobs?status=completed", nil)
		w := httptest.NewRecorder()

		handler.ListAggregationJobsHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		jobs := response["jobs"].([]interface{})
		assert.Len(t, jobs, 1)
	})

	t.Run("pagination", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/aggregate/jobs?page=1&limit=1", nil)
		w := httptest.NewRecorder()

		handler.ListAggregationJobsHandler(w, req)

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

func TestDataAggregationHandler_GetAggregationSchemaHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataAggregationHandler(logger, mockMetrics)

	t.Run("existing schema", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/aggregate/schema/business_metrics_default", nil)
		w := httptest.NewRecorder()

		handler.GetAggregationSchemaHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response AggregationSchema
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "business_metrics_default", response.ID)
		assert.Equal(t, "Default Business Metrics", response.Name)
		assert.Equal(t, AggregationTypeBusinessMetrics, response.Type)
	})

	t.Run("non-existent schema", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/aggregate/schema/non_existent", nil)
		w := httptest.NewRecorder()

		handler.GetAggregationSchemaHandler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("missing schema ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/aggregate/schema/", nil)
		w := httptest.NewRecorder()

		handler.GetAggregationSchemaHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDataAggregationHandler_ListAggregationSchemasHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataAggregationHandler(logger, mockMetrics)

	t.Run("list all schemas", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/aggregate/schemas", nil)
		w := httptest.NewRecorder()

		handler.ListAggregationSchemasHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.NotNil(t, response["schemas"])
		assert.Equal(t, float64(2), response["total"]) // 2 default schemas
	})

	t.Run("filter by type", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/aggregate/schemas?type=business_metrics", nil)
		w := httptest.NewRecorder()

		handler.ListAggregationSchemasHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		schemas := response["schemas"].([]interface{})
		assert.Len(t, schemas, 1)
	})
}

func TestDataAggregationHandler_ValidationAndUtilityFunctions(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataAggregationHandler(logger, mockMetrics)

	t.Run("valid aggregation type", func(t *testing.T) {
		assert.True(t, handler.isValidAggregationType(AggregationTypeBusinessMetrics))
		assert.True(t, handler.isValidAggregationType(AggregationTypeRiskAssessments))
		assert.True(t, handler.isValidAggregationType(AggregationTypeComplianceReports))
		assert.True(t, handler.isValidAggregationType(AggregationTypePerformanceAnalytics))
		assert.True(t, handler.isValidAggregationType(AggregationTypeTrendAnalysis))
		assert.True(t, handler.isValidAggregationType(AggregationTypeCustom))
		assert.True(t, handler.isValidAggregationType(AggregationTypeAll))
	})

	t.Run("invalid aggregation type", func(t *testing.T) {
		assert.False(t, handler.isValidAggregationType("invalid_type"))
		assert.False(t, handler.isValidAggregationType(""))
	})

	t.Run("validate aggregation request", func(t *testing.T) {
		validRequest := &DataAggregationRequest{
			AggregationType: AggregationTypeBusinessMetrics,
			Data:            map[string]interface{}{"test": "data"},
		}

		err := handler.validateAggregationRequest(validRequest)
		assert.NoError(t, err)
	})

	t.Run("validate aggregation request - missing type", func(t *testing.T) {
		invalidRequest := &DataAggregationRequest{
			Data: map[string]interface{}{"test": "data"},
		}

		err := handler.validateAggregationRequest(invalidRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "aggregation_type is required")
	})

	t.Run("validate aggregation request - missing data", func(t *testing.T) {
		invalidRequest := &DataAggregationRequest{
			AggregationType: AggregationTypeBusinessMetrics,
			Data:            nil,
		}

		err := handler.validateAggregationRequest(invalidRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "data is required")
	})

	t.Run("validate aggregation request - invalid type", func(t *testing.T) {
		invalidRequest := &DataAggregationRequest{
			AggregationType: "invalid_type",
			Data:            map[string]interface{}{"test": "data"},
		}

		err := handler.validateAggregationRequest(invalidRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid aggregation_type")
	})
}

func TestDataAggregationHandler_AggregationLogic(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataAggregationHandler(logger, mockMetrics)

	t.Run("aggregate data", func(t *testing.T) {
		req := &DataAggregationRequest{
			BusinessID:      "business_123",
			AggregationType: AggregationTypeBusinessMetrics,
			Data: map[string]interface{}{
				"verifications": []map[string]interface{}{
					{"status": "passed", "score": 0.95},
					{"status": "failed", "score": 0.30},
				},
			},
			Rules: []DataAggregationRule{
				{
					Field:       "verification_count",
					Operation:   AggregationOperationCount,
					Parameters:  map[string]interface{}{},
					Description: "Count total verifications",
					Enabled:     true,
					Order:       1,
				},
			},
		}

		result, err := handler.aggregateData(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.AggregationID)
		assert.Equal(t, req.BusinessID, result.BusinessID)
		assert.Equal(t, req.AggregationType, result.AggregationType)
		assert.True(t, result.IsSuccessful)
		assert.NotZero(t, result.AggregatedAt)
		assert.NotZero(t, result.ProcessingTime)
	})

	t.Run("create aggregation job", func(t *testing.T) {
		req := &DataAggregationRequest{
			BusinessID:      "business_123",
			AggregationType: AggregationTypeBusinessMetrics,
			Data:            map[string]interface{}{"test": "data"},
		}

		job, err := handler.createAggregationJob(req)
		assert.NoError(t, err)
		assert.NotNil(t, job)
		assert.NotEmpty(t, job.ID)
		assert.Equal(t, req.BusinessID, job.BusinessID)
		assert.Equal(t, req.AggregationType, job.AggregationType)
		assert.Equal(t, "pending", job.Status)
		assert.Equal(t, 0, job.Progress)
		assert.NotZero(t, job.CreatedAt)
	})

	t.Run("apply aggregation rule", func(t *testing.T) {
		rule := DataAggregationRule{
			Field:       "test_field",
			Operation:   AggregationOperationCount,
			Parameters:  map[string]interface{}{},
			Description: "Count test field",
			Enabled:     true,
			Order:       1,
		}

		data := map[string]interface{}{"test_field": "test value"}

		result, err := handler.applyAggregationRule(data, rule)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("unsupported aggregation operation", func(t *testing.T) {
		rule := DataAggregationRule{
			Field:       "test_field",
			Operation:   "unsupported_operation",
			Parameters:  map[string]interface{}{},
			Description: "Test unsupported operation",
			Enabled:     true,
			Order:       1,
		}

		data := map[string]interface{}{"test_field": "test value"}

		result, err := handler.applyAggregationRule(data, rule)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unsupported aggregation operation")
	})
}

func TestDataAggregationHandler_AggregationOperations(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataAggregationHandler(logger, mockMetrics)

	t.Run("count aggregation", func(t *testing.T) {
		rule := DataAggregationRule{
			Field:       "test_field",
			Operation:   AggregationOperationCount,
			Parameters:  map[string]interface{}{},
			Description: "Count test field",
			Enabled:     true,
			Order:       1,
		}

		data := map[string]interface{}{"test_field": "test value"}

		result, err := handler.applyCountAggregation(data, rule)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		resultMap, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "test_field", resultMap["field"])
		assert.Equal(t, float64(100), resultMap["count"])
		assert.Equal(t, "count_aggregation", resultMap["result"])
	})

	t.Run("sum aggregation", func(t *testing.T) {
		rule := DataAggregationRule{
			Field:       "test_field",
			Operation:   AggregationOperationSum,
			Parameters:  map[string]interface{}{},
			Description: "Sum test field",
			Enabled:     true,
			Order:       1,
		}

		data := map[string]interface{}{"test_field": "test value"}

		result, err := handler.applySumAggregation(data, rule)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		resultMap, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "test_field", resultMap["field"])
		assert.Equal(t, 1500.50, resultMap["sum"])
		assert.Equal(t, "sum_aggregation", resultMap["result"])
	})

	t.Run("average aggregation", func(t *testing.T) {
		rule := DataAggregationRule{
			Field:       "test_field",
			Operation:   AggregationOperationAverage,
			Parameters:  map[string]interface{}{},
			Description: "Average test field",
			Enabled:     true,
			Order:       1,
		}

		data := map[string]interface{}{"test_field": "test value"}

		result, err := handler.applyAverageAggregation(data, rule)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		resultMap, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "test_field", resultMap["field"])
		assert.Equal(t, 75.25, resultMap["average"])
		assert.Equal(t, "average_aggregation", resultMap["result"])
	})

	t.Run("percentile aggregation with parameter", func(t *testing.T) {
		rule := DataAggregationRule{
			Field:       "test_field",
			Operation:   AggregationOperationPercentile,
			Parameters:  map[string]interface{}{"percentile": 95.0},
			Description: "95th percentile of test field",
			Enabled:     true,
			Order:       1,
		}

		data := map[string]interface{}{"test_field": "test value"}

		result, err := handler.applyPercentileAggregation(data, rule)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		resultMap, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "test_field", resultMap["field"])
		assert.Equal(t, 95.0, resultMap["percentile"])
		assert.Equal(t, 88.0, resultMap["value"])
		assert.Equal(t, "percentile_aggregation", resultMap["result"])
	})
}

func TestDataAggregationHandler_UtilityFunctions(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataAggregationHandler(logger, mockMetrics)

	t.Run("generate aggregation ID", func(t *testing.T) {
		id1 := handler.generateAggregationID()
		id2 := handler.generateAggregationID()

		assert.NotEmpty(t, id1)
		assert.NotEmpty(t, id2)
		assert.NotEqual(t, id1, id2)
		assert.Contains(t, id1, "agg_")
	})

	t.Run("evaluate condition", func(t *testing.T) {
		data := map[string]interface{}{"test": "value"}
		condition := "test == 'value'"

		result := handler.evaluateCondition(data, condition)
		assert.True(t, result) // Mock implementation always returns true
	})

	t.Run("get data count", func(t *testing.T) {
		data := map[string]interface{}{"test": "value"}

		count := handler.getDataCount(data)
		assert.Equal(t, 100, count) // Mock implementation returns 100
	})
}

func TestDataAggregationHandler_DefaultSchemas(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataAggregationHandler(logger, mockMetrics)

	t.Run("default schemas initialized", func(t *testing.T) {
		handler.schemaMutex.RLock()
		defer handler.schemaMutex.RUnlock()

		// Check business metrics schema
		businessMetricsSchema, exists := handler.schemas["business_metrics_default"]
		assert.True(t, exists)
		assert.Equal(t, "business_metrics_default", businessMetricsSchema.ID)
		assert.Equal(t, "Default Business Metrics", businessMetricsSchema.Name)
		assert.Equal(t, AggregationTypeBusinessMetrics, businessMetricsSchema.Type)
		assert.Len(t, businessMetricsSchema.Rules, 2)

		// Check risk assessment schema
		riskAssessmentSchema, exists := handler.schemas["risk_assessment_default"]
		assert.True(t, exists)
		assert.Equal(t, "risk_assessment_default", riskAssessmentSchema.ID)
		assert.Equal(t, "Default Risk Assessment", riskAssessmentSchema.Name)
		assert.Equal(t, AggregationTypeRiskAssessments, riskAssessmentSchema.Type)
		assert.Len(t, riskAssessmentSchema.Rules, 2)
	})

	t.Run("schema rules validation", func(t *testing.T) {
		handler.schemaMutex.RLock()
		businessMetricsSchema := handler.schemas["business_metrics_default"]
		handler.schemaMutex.RUnlock()

		// Check first rule
		firstRule := businessMetricsSchema.Rules[0]
		assert.Equal(t, "verification_count", firstRule.Field)
		assert.Equal(t, AggregationOperationCount, firstRule.Operation)
		assert.Equal(t, "Count total verifications", firstRule.Description)
		assert.True(t, firstRule.Enabled)
		assert.Equal(t, 1, firstRule.Order)

		// Check second rule
		secondRule := businessMetricsSchema.Rules[1]
		assert.Equal(t, "success_rate", secondRule.Field)
		assert.Equal(t, AggregationOperationAverage, secondRule.Operation)
		assert.Equal(t, "Calculate average success rate", secondRule.Description)
		assert.True(t, secondRule.Enabled)
		assert.Equal(t, 2, secondRule.Order)
	})
}
