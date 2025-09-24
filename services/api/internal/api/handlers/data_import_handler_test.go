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

	"kyb-platform/internal/observability"
)

func TestNewDataImportHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataImportHandler(logger, mockMetrics)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
	assert.Equal(t, mockMetrics, handler.metrics)
	assert.NotNil(t, handler.importJobs)
	assert.Equal(t, 0, handler.jobCounter)
}

func TestDataImportHandler_ImportDataHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataImportHandler(logger, mockMetrics)

	t.Run("successful import", func(t *testing.T) {
		request := ImportRequest{
			BusinessID: "business_123",
			ImportType: ImportTypeBusinessVerifications,
			Format:     ImportFormatJSON,
			Mode:       ImportModeUpsert,
			Data: map[string]interface{}{
				"business_name": "Test Company",
				"address":       "123 Test St",
			},
		}

		requestJSON, _ := json.Marshal(request)
		req := httptest.NewRequest("POST", "/v1/import", bytes.NewBuffer(requestJSON))
		w := httptest.NewRecorder()

		handler.ImportDataHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.NotEmpty(t, w.Header().Get("X-Import-ID"))

		var response ImportResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.NotEmpty(t, response.ImportID)
		assert.Equal(t, request.BusinessID, response.BusinessID)
		assert.Equal(t, request.ImportType, response.ImportType)
		assert.Equal(t, request.Format, response.Format)
		assert.Equal(t, request.Mode, response.Mode)
		assert.Equal(t, "completed", response.Status)
		assert.Greater(t, response.RecordCount, 0)
		assert.NotZero(t, response.ProcessedAt)
	})

	t.Run("invalid method", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/import", nil)
		w := httptest.NewRecorder()

		handler.ImportDataHandler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/import", bytes.NewBufferString("invalid json"))
		w := httptest.NewRecorder()

		handler.ImportDataHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing import type", func(t *testing.T) {
		request := ImportRequest{
			BusinessID: "business_123",
			Format:     ImportFormatJSON,
			Data:       map[string]interface{}{},
		}

		requestJSON, _ := json.Marshal(request)
		req := httptest.NewRequest("POST", "/v1/import", bytes.NewBuffer(requestJSON))
		w := httptest.NewRecorder()

		handler.ImportDataHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "import_type is required")
	})

	t.Run("invalid import type", func(t *testing.T) {
		request := ImportRequest{
			BusinessID: "business_123",
			ImportType: "invalid_type",
			Format:     ImportFormatJSON,
			Data:       map[string]interface{}{},
		}

		requestJSON, _ := json.Marshal(request)
		req := httptest.NewRequest("POST", "/v1/import", bytes.NewBuffer(requestJSON))
		w := httptest.NewRecorder()

		handler.ImportDataHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid import_type")
	})

	t.Run("missing data", func(t *testing.T) {
		request := ImportRequest{
			BusinessID: "business_123",
			ImportType: ImportTypeBusinessVerifications,
			Format:     ImportFormatJSON,
		}

		requestJSON, _ := json.Marshal(request)
		req := httptest.NewRequest("POST", "/v1/import", bytes.NewBuffer(requestJSON))
		w := httptest.NewRecorder()

		handler.ImportDataHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "data is required")
	})
}

func TestDataImportHandler_CreateImportJobHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataImportHandler(logger, mockMetrics)

	t.Run("successful job creation", func(t *testing.T) {
		request := ImportRequest{
			BusinessID: "business_123",
			ImportType: ImportTypeClassifications,
			Format:     ImportFormatCSV,
			Mode:       ImportModeCreate,
			Data:       "csv,data,here",
		}

		requestJSON, _ := json.Marshal(request)
		req := httptest.NewRequest("POST", "/v1/import/job", bytes.NewBuffer(requestJSON))
		w := httptest.NewRecorder()

		handler.CreateImportJobHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.NotEmpty(t, w.Header().Get("X-Job-ID"))

		var job ImportJob
		err := json.Unmarshal(w.Body.Bytes(), &job)
		assert.NoError(t, err)

		assert.NotEmpty(t, job.ID)
		assert.Equal(t, request.BusinessID, job.BusinessID)
		assert.Equal(t, request.ImportType, job.ImportType)
		assert.Equal(t, request.Format, job.Format)
		assert.Equal(t, request.Mode, job.Mode)
		assert.Equal(t, "pending", job.Status)
		assert.Equal(t, 0, job.Progress)
		assert.NotZero(t, job.CreatedAt)
	})

	t.Run("invalid method", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/import/job", nil)
		w := httptest.NewRecorder()

		handler.CreateImportJobHandler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/import/job", bytes.NewBufferString("invalid json"))
		w := httptest.NewRecorder()

		handler.CreateImportJobHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDataImportHandler_GetImportJobHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataImportHandler(logger, mockMetrics)

	// Create a test job
	job := &ImportJob{
		ID:          "test_job_123",
		BusinessID:  "business_123",
		ImportType:  ImportTypeBusinessVerifications,
		Format:      ImportFormatJSON,
		Mode:        ImportModeUpsert,
		Status:      "completed",
		Progress:    100,
		RecordCount: 50,
		CreatedAt:   time.Now(),
	}

	handler.jobMutex.Lock()
	handler.importJobs[job.ID] = job
	handler.jobMutex.Unlock()

	t.Run("successful job retrieval", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/import/job/test_job_123", nil)
		w := httptest.NewRecorder()

		handler.GetImportJobHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var retrievedJob ImportJob
		err := json.Unmarshal(w.Body.Bytes(), &retrievedJob)
		assert.NoError(t, err)

		assert.Equal(t, job.ID, retrievedJob.ID)
		assert.Equal(t, job.BusinessID, retrievedJob.BusinessID)
		assert.Equal(t, job.ImportType, retrievedJob.ImportType)
		assert.Equal(t, job.Status, retrievedJob.Status)
	})

	t.Run("invalid method", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/import/job/test_job_123", nil)
		w := httptest.NewRecorder()

		handler.GetImportJobHandler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("missing job ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/import/job/", nil)
		w := httptest.NewRecorder()

		handler.GetImportJobHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("job not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/import/job/nonexistent_job", nil)
		w := httptest.NewRecorder()

		handler.GetImportJobHandler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestDataImportHandler_ListImportJobsHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataImportHandler(logger, mockMetrics)

	// Create test jobs
	jobs := []*ImportJob{
		{
			ID:         "job_1",
			BusinessID: "business_123",
			ImportType: ImportTypeBusinessVerifications,
			Status:     "completed",
			CreatedAt:  time.Now(),
		},
		{
			ID:         "job_2",
			BusinessID: "business_123",
			ImportType: ImportTypeClassifications,
			Status:     "processing",
			CreatedAt:  time.Now(),
		},
		{
			ID:         "job_3",
			BusinessID: "business_456",
			ImportType: ImportTypeRiskAssessments,
			Status:     "completed",
			CreatedAt:  time.Now(),
		},
	}

	handler.jobMutex.Lock()
	for _, job := range jobs {
		handler.importJobs[job.ID] = job
	}
	handler.jobMutex.Unlock()

	t.Run("list all jobs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/import/jobs", nil)
		w := httptest.NewRecorder()

		handler.ListImportJobsHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, float64(3), response["total"])
		assert.Equal(t, float64(50), response["limit"])
		assert.Equal(t, float64(0), response["offset"])
	})

	t.Run("filter by business ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/import/jobs?business_id=business_123", nil)
		w := httptest.NewRecorder()

		handler.ListImportJobsHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, float64(2), response["total"])
	})

	t.Run("filter by status", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/import/jobs?status=completed", nil)
		w := httptest.NewRecorder()

		handler.ListImportJobsHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, float64(2), response["total"])
	})

	t.Run("pagination", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/import/jobs?limit=2&offset=1", nil)
		w := httptest.NewRecorder()

		handler.ListImportJobsHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, float64(2), response["limit"])
		assert.Equal(t, float64(1), response["offset"])
	})

	t.Run("invalid method", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/import/jobs", nil)
		w := httptest.NewRecorder()

		handler.ListImportJobsHandler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestDataImportHandler_ValidationFunctions(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataImportHandler(logger, mockMetrics)

	t.Run("isValidImportType", func(t *testing.T) {
		assert.True(t, handler.isValidImportType(ImportTypeBusinessVerifications))
		assert.True(t, handler.isValidImportType(ImportTypeClassifications))
		assert.True(t, handler.isValidImportType(ImportTypeRiskAssessments))
		assert.True(t, handler.isValidImportType(ImportTypeComplianceReports))
		assert.True(t, handler.isValidImportType(ImportTypeAuditTrails))
		assert.True(t, handler.isValidImportType(ImportTypeMetrics))
		assert.True(t, handler.isValidImportType(ImportTypeAll))
		assert.False(t, handler.isValidImportType("invalid_type"))
	})

	t.Run("isValidImportFormat", func(t *testing.T) {
		assert.True(t, handler.isValidImportFormat(ImportFormatJSON))
		assert.True(t, handler.isValidImportFormat(ImportFormatCSV))
		assert.True(t, handler.isValidImportFormat(ImportFormatXML))
		assert.True(t, handler.isValidImportFormat(ImportFormatXLSX))
		assert.False(t, handler.isValidImportFormat("invalid_format"))
	})

	t.Run("isValidImportMode", func(t *testing.T) {
		assert.True(t, handler.isValidImportMode(ImportModeCreate))
		assert.True(t, handler.isValidImportMode(ImportModeUpdate))
		assert.True(t, handler.isValidImportMode(ImportModeUpsert))
		assert.True(t, handler.isValidImportMode(ImportModeReplace))
		assert.False(t, handler.isValidImportMode("invalid_mode"))
	})

	t.Run("isValidConflictPolicy", func(t *testing.T) {
		assert.True(t, handler.isValidConflictPolicy("skip"))
		assert.True(t, handler.isValidConflictPolicy("update"))
		assert.True(t, handler.isValidConflictPolicy("error"))
		assert.False(t, handler.isValidConflictPolicy("invalid_policy"))
	})
}

func TestDataImportHandler_DataProcessingFunctions(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataImportHandler(logger, mockMetrics)

	request := ImportRequest{
		BusinessID: "business_123",
		ImportType: ImportTypeBusinessVerifications,
		Format:     ImportFormatJSON,
		Mode:       ImportModeUpsert,
		Data:       map[string]interface{}{},
	}

	t.Run("parseCSVData", func(t *testing.T) {
		data, errors, warnings := handler.parseCSVData("csv_data")

		assert.NotNil(t, data)
		assert.Empty(t, errors)
		assert.Empty(t, warnings)
	})

	t.Run("parseXMLData", func(t *testing.T) {
		data, errors, warnings := handler.parseXMLData("xml_data")

		assert.NotNil(t, data)
		assert.Empty(t, errors)
		assert.Empty(t, warnings)
	})

	t.Run("parseXLSXData", func(t *testing.T) {
		data, errors, warnings := handler.parseXLSXData("xlsx_data")

		assert.NotNil(t, data)
		assert.Empty(t, errors)
		assert.Empty(t, warnings)
	})

	t.Run("applyValidationRules", func(t *testing.T) {
		data := map[string]interface{}{"test": "value"}
		rules := map[string]interface{}{"rule": "value"}

		errors, warnings := handler.applyValidationRules(data, rules)

		assert.Empty(t, errors)
		assert.Empty(t, warnings)
	})

	t.Run("applyTransformRules", func(t *testing.T) {
		data := map[string]interface{}{"test": "value"}
		rules := map[string]interface{}{"rule": "value"}

		result := handler.applyTransformRules(data, rules)

		assert.Equal(t, data, result)
	})

	t.Run("processBusinessVerifications", func(t *testing.T) {
		data, errors, warnings := handler.processBusinessVerifications(map[string]interface{}{}, request)

		assert.NotNil(t, data)
		assert.Empty(t, errors)
		assert.Empty(t, warnings)
	})

	t.Run("processClassifications", func(t *testing.T) {
		data, errors, warnings := handler.processClassifications(map[string]interface{}{}, request)

		assert.NotNil(t, data)
		assert.Empty(t, errors)
		assert.Empty(t, warnings)
	})

	t.Run("processRiskAssessments", func(t *testing.T) {
		data, errors, warnings := handler.processRiskAssessments(map[string]interface{}{}, request)

		assert.NotNil(t, data)
		assert.Empty(t, errors)
		assert.Empty(t, warnings)
	})

	t.Run("processComplianceReports", func(t *testing.T) {
		data, errors, warnings := handler.processComplianceReports(map[string]interface{}{}, request)

		assert.NotNil(t, data)
		assert.Empty(t, errors)
		assert.Empty(t, warnings)
	})

	t.Run("processAuditTrails", func(t *testing.T) {
		data, errors, warnings := handler.processAuditTrails(map[string]interface{}{}, request)

		assert.NotNil(t, data)
		assert.Empty(t, errors)
		assert.Empty(t, warnings)
	})

	t.Run("processMetrics", func(t *testing.T) {
		data, errors, warnings := handler.processMetrics(map[string]interface{}{}, request)

		assert.NotNil(t, data)
		assert.Empty(t, errors)
		assert.Empty(t, warnings)
	})

	t.Run("processAllData", func(t *testing.T) {
		data, errors, warnings := handler.processAllData(map[string]interface{}{}, request)

		assert.NotNil(t, data)
		assert.Empty(t, errors)
		assert.Empty(t, warnings)
	})

	t.Run("countRecords", func(t *testing.T) {
		count := handler.countRecords(map[string]interface{}{})

		assert.Equal(t, 150, count)
	})
}

func TestDataImportHandler_UtilityFunctions(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataImportHandler(logger, mockMetrics)

	t.Run("extractPathParam", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/import/job/test_job_123", nil)

		jobID := handler.extractPathParam(req, "job_id")
		assert.Equal(t, "test_job_123", jobID)

		// Test with missing parameter
		req2 := httptest.NewRequest("GET", "/v1/import/job/", nil)
		jobID2 := handler.extractPathParam(req2, "job_id")
		assert.Equal(t, "", jobID2)
	})
}

func TestDataImportHandler_ImportDataProcessing(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataImportHandler(logger, mockMetrics)

	t.Run("successful import data processing", func(t *testing.T) {
		request := ImportRequest{
			BusinessID: "business_123",
			ImportType: ImportTypeBusinessVerifications,
			Format:     ImportFormatJSON,
			Mode:       ImportModeUpsert,
			Data: map[string]interface{}{
				"business_name": "Test Company",
				"address":       "123 Test St",
			},
		}

		response, err := handler.importData(context.Background(), request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.ImportID)
		assert.Equal(t, request.BusinessID, response.BusinessID)
		assert.Equal(t, request.ImportType, response.ImportType)
		assert.Equal(t, "completed", response.Status)
		assert.Greater(t, response.RecordCount, 0)
		assert.NotZero(t, response.ProcessedAt)
	})

	t.Run("unsupported import type", func(t *testing.T) {
		request := ImportRequest{
			BusinessID: "business_123",
			ImportType: "unsupported_type",
			Format:     ImportFormatJSON,
			Mode:       ImportModeUpsert,
			Data:       map[string]interface{}{},
		}

		response, err := handler.importData(context.Background(), request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "unsupported import type")
	})

	t.Run("unsupported format", func(t *testing.T) {
		request := ImportRequest{
			BusinessID: "business_123",
			ImportType: ImportTypeBusinessVerifications,
			Format:     "unsupported_format",
			Mode:       ImportModeUpsert,
			Data:       map[string]interface{}{},
		}

		response, err := handler.importData(context.Background(), request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "unsupported format")
	})
}

func TestDataImportHandler_JobManagement(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockMetrics := &observability.Metrics{}

	handler := NewDataImportHandler(logger, mockMetrics)

	t.Run("create import job", func(t *testing.T) {
		request := ImportRequest{
			BusinessID: "business_123",
			ImportType: ImportTypeClassifications,
			Format:     ImportFormatCSV,
			Mode:       ImportModeCreate,
			Data:       "csv_data",
		}

		job, err := handler.createImportJob(context.Background(), request)

		assert.NoError(t, err)
		assert.NotNil(t, job)
		assert.NotEmpty(t, job.ID)
		assert.Equal(t, request.BusinessID, job.BusinessID)
		assert.Equal(t, request.ImportType, job.ImportType)
		assert.Equal(t, "pending", job.Status)
		assert.Equal(t, 0, job.Progress)
		assert.NotZero(t, job.CreatedAt)

		// Verify job is stored
		handler.jobMutex.RLock()
		storedJob, exists := handler.importJobs[job.ID]
		handler.jobMutex.RUnlock()

		assert.True(t, exists)
		assert.Equal(t, job.ID, storedJob.ID)
	})

	t.Run("job counter increments", func(t *testing.T) {
		initialCounter := handler.jobCounter

		request := ImportRequest{
			BusinessID: "business_456",
			ImportType: ImportTypeRiskAssessments,
			Format:     ImportFormatJSON,
			Mode:       ImportModeUpsert,
			Data:       map[string]interface{}{},
		}

		job, err := handler.createImportJob(context.Background(), request)

		assert.NoError(t, err)
		assert.NotNil(t, job)
		assert.Equal(t, initialCounter+1, handler.jobCounter)
	})
}
