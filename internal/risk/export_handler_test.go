package risk

import (
	"bytes"
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

func TestExportHandler_CreateExportJob(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	mockJobManager := &MockExportJobManager{}
	handler := NewExportHandler(logger, mockExportSvc, mockJobManager)

	// Mock the job manager
	mockJob := &ExportJob{
		ID:         "test-job-123",
		BusinessID: "test-business-123",
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
		Status:     "pending",
		Progress:   0,
		CreatedAt:  time.Now(),
		Metadata:   map[string]interface{}{"test": "value"},
	}

	mockJobManager.On("CreateExportJob", mock.Anything, mock.Anything).Return(mockJob, nil)

	// Create request
	reqBody := CreateExportJobRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
		Metadata:   map[string]interface{}{"test": "value"},
	}

	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/export/jobs", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.CreateExportJob(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response CreateExportJobResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-job-123", response.JobID)
	assert.Equal(t, "pending", response.Status)
	assert.Equal(t, "Export job created successfully", response.Message)

	mockJobManager.AssertExpectations(t)
}

func TestExportHandler_CreateExportJob_InvalidRequest(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	mockJobManager := &MockExportJobManager{}
	handler := NewExportHandler(logger, mockExportSvc, mockJobManager)

	// Create request with missing business_id
	reqBody := CreateExportJobRequest{
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/export/jobs", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.CreateExportJob(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "business_id is required")
}

func TestExportHandler_CreateExportJob_JobManagerError(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	mockJobManager := &MockExportJobManager{}
	handler := NewExportHandler(logger, mockExportSvc, mockJobManager)

	// Mock the job manager to return an error
	mockJobManager.On("CreateExportJob", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	// Create request
	reqBody := CreateExportJobRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/export/jobs", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.CreateExportJob(w, req)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to create export job")

	mockJobManager.AssertExpectations(t)
}

func TestExportHandler_GetExportJob(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	mockJobManager := &MockExportJobManager{}
	handler := NewExportHandler(logger, mockExportSvc, mockJobManager)

	// Mock the job manager
	mockJob := &ExportJob{
		ID:          "test-job-123",
		BusinessID:  "test-business-123",
		ExportType:  ExportTypeAssessments,
		Format:      ExportFormatJSON,
		Status:      "completed",
		Progress:    100,
		CreatedAt:   time.Now(),
		CompletedAt: &[]time.Time{time.Now()}[0],
	}

	mockJobManager.On("GetExportJob", "test-job-123").Return(mockJob, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/export/jobs/test-job-123", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.GetExportJob(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response GetExportJobResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-job-123", response.Job.ID)
	assert.Equal(t, "completed", response.Job.Status)

	mockJobManager.AssertExpectations(t)
}

func TestExportHandler_GetExportJob_NotFound(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	mockJobManager := &MockExportJobManager{}
	handler := NewExportHandler(logger, mockExportSvc, mockJobManager)

	// Mock the job manager to return an error
	mockJobManager.On("GetExportJob", "non-existent-job").Return(nil, assert.AnError)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/export/jobs/non-existent-job", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.GetExportJob(w, req)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Export job not found")

	mockJobManager.AssertExpectations(t)
}

func TestExportHandler_ListExportJobs(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	mockJobManager := &MockExportJobManager{}
	handler := NewExportHandler(logger, mockExportSvc, mockJobManager)

	// Mock the job manager
	mockJobs := []*ExportJob{
		{
			ID:         "test-job-1",
			BusinessID: "test-business-123",
			ExportType: ExportTypeAssessments,
			Format:     ExportFormatJSON,
			Status:     "completed",
			Progress:   100,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "test-job-2",
			BusinessID: "test-business-123",
			ExportType: ExportTypeFactors,
			Format:     ExportFormatCSV,
			Status:     "pending",
			Progress:   0,
			CreatedAt:  time.Now(),
		},
	}

	mockJobManager.On("ListExportJobs", "test-business-123").Return(mockJobs, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/export/jobs?business_id=test-business-123", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.ListExportJobs(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response ListExportJobsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Jobs, 2)
	assert.Equal(t, 2, response.Total)

	mockJobManager.AssertExpectations(t)
}

func TestExportHandler_ListExportJobs_MissingBusinessID(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	mockJobManager := &MockExportJobManager{}
	handler := NewExportHandler(logger, mockExportSvc, mockJobManager)

	// Create request without business_id
	req := httptest.NewRequest("GET", "/api/v1/export/jobs", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.ListExportJobs(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "business_id query parameter is required")
}

func TestExportHandler_CancelExportJob(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	mockJobManager := &MockExportJobManager{}
	handler := NewExportHandler(logger, mockExportSvc, mockJobManager)

	// Mock the job manager
	mockJobManager.On("CancelExportJob", "test-job-123").Return(nil)

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/export/jobs/test-job-123", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.CancelExportJob(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Export job cancelled successfully", response["message"])
	assert.Equal(t, "test-job-123", response["job_id"])

	mockJobManager.AssertExpectations(t)
}

func TestExportHandler_CancelExportJob_Error(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	mockJobManager := &MockExportJobManager{}
	handler := NewExportHandler(logger, mockExportSvc, mockJobManager)

	// Mock the job manager to return an error
	mockJobManager.On("CancelExportJob", "test-job-123").Return(assert.AnError)

	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/export/jobs/test-job-123", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.CancelExportJob(w, req)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to cancel export job")

	mockJobManager.AssertExpectations(t)
}

func TestExportHandler_GetExportJobStatistics(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	mockJobManager := &MockExportJobManager{}
	handler := NewExportHandler(logger, mockExportSvc, mockJobManager)

	// Mock the job manager
	mockStats := map[string]interface{}{
		"total_jobs":      5,
		"pending_jobs":    1,
		"processing_jobs": 1,
		"completed_jobs":  2,
		"failed_jobs":     1,
		"cancelled_jobs":  0,
	}

	mockJobManager.On("GetJobStatistics").Return(mockStats)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/export/jobs/statistics", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.GetExportJobStatistics(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response GetExportJobStatisticsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 5, response.Statistics["total_jobs"])
	assert.Equal(t, 1, response.Statistics["pending_jobs"])

	mockJobManager.AssertExpectations(t)
}

func TestExportHandler_CleanupOldJobs(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	mockJobManager := &MockExportJobManager{}
	handler := NewExportHandler(logger, mockExportSvc, mockJobManager)

	// Mock the job manager
	mockJobManager.On("CleanupOldJobs", mock.Anything).Return(nil)

	// Create request
	req := httptest.NewRequest("POST", "/api/v1/export/jobs/cleanup?hours=24", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.CleanupOldJobs(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Old jobs cleaned up successfully", response["message"])
	assert.Equal(t, 24, response["hours"])

	mockJobManager.AssertExpectations(t)
}

func TestExportHandler_CleanupOldJobs_InvalidHours(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	mockJobManager := &MockExportJobManager{}
	handler := NewExportHandler(logger, mockExportSvc, mockJobManager)

	// Create request with invalid hours
	req := httptest.NewRequest("POST", "/api/v1/export/jobs/cleanup?hours=invalid", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.CleanupOldJobs(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid hours parameter")
}

func TestExportHandler_ExportData_Assessments(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	mockJobManager := &MockExportJobManager{}
	handler := NewExportHandler(logger, mockExportSvc, mockJobManager)

	// Mock the export service
	mockResponse := &ExportResponse{
		ExportID:    "test-export-123",
		BusinessID:  "test-business-123",
		ExportType:  ExportTypeAssessments,
		Format:      ExportFormatJSON,
		Data:        map[string]interface{}{"test": "data"},
		RecordCount: 1,
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	mockExportSvc.On("ExportRiskAssessments", mock.Anything, mock.Anything, ExportFormatJSON).Return(mockResponse, nil)

	// Create request
	reqBody := CreateExportJobRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/export/data", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.ExportData(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response ExportResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-export-123", response.ExportID)
	assert.Equal(t, ExportTypeAssessments, response.ExportType)

	mockExportSvc.AssertExpectations(t)
}

func TestExportHandler_ExportData_Factors(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	mockJobManager := &MockExportJobManager{}
	handler := NewExportHandler(logger, mockExportSvc, mockJobManager)

	// Mock the export service
	mockResponse := &ExportResponse{
		ExportID:    "test-export-123",
		BusinessID:  "test-business-123",
		ExportType:  ExportTypeFactors,
		Format:      ExportFormatJSON,
		Data:        map[string]interface{}{"test": "data"},
		RecordCount: 1,
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	mockExportSvc.On("ExportRiskFactors", mock.Anything, mock.Anything, ExportFormatJSON).Return(mockResponse, nil)

	// Create request
	reqBody := CreateExportJobRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeFactors,
		Format:     ExportFormatJSON,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/export/data", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.ExportData(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response ExportResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-export-123", response.ExportID)
	assert.Equal(t, ExportTypeFactors, response.ExportType)

	mockExportSvc.AssertExpectations(t)
}

func TestExportHandler_ExportData_Trends(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	mockJobManager := &MockExportJobManager{}
	handler := NewExportHandler(logger, mockExportSvc, mockJobManager)

	// Mock the export service
	mockResponse := &ExportResponse{
		ExportID:    "test-export-123",
		BusinessID:  "test-business-123",
		ExportType:  ExportTypeTrends,
		Format:      ExportFormatJSON,
		Data:        map[string]interface{}{"test": "data"},
		RecordCount: 1,
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	mockExportSvc.On("ExportRiskTrends", mock.Anything, mock.Anything, ExportFormatJSON).Return(mockResponse, nil)

	// Create request
	reqBody := CreateExportJobRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeTrends,
		Format:     ExportFormatJSON,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/export/data", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.ExportData(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response ExportResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-export-123", response.ExportID)
	assert.Equal(t, ExportTypeTrends, response.ExportType)

	mockExportSvc.AssertExpectations(t)
}

func TestExportHandler_ExportData_Alerts(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	mockJobManager := &MockExportJobManager{}
	handler := NewExportHandler(logger, mockExportSvc, mockJobManager)

	// Mock the export service
	mockResponse := &ExportResponse{
		ExportID:    "test-export-123",
		BusinessID:  "test-business-123",
		ExportType:  ExportTypeAlerts,
		Format:      ExportFormatJSON,
		Data:        map[string]interface{}{"test": "data"},
		RecordCount: 1,
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	mockExportSvc.On("ExportRiskAlerts", mock.Anything, mock.Anything, ExportFormatJSON).Return(mockResponse, nil)

	// Create request
	reqBody := CreateExportJobRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeAlerts,
		Format:     ExportFormatJSON,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/export/data", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.ExportData(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response ExportResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-export-123", response.ExportID)
	assert.Equal(t, ExportTypeAlerts, response.ExportType)

	mockExportSvc.AssertExpectations(t)
}

func TestExportHandler_ExportData_UnsupportedType(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	mockJobManager := &MockExportJobManager{}
	handler := NewExportHandler(logger, mockExportSvc, mockJobManager)

	// Create request with unsupported export type
	reqBody := CreateExportJobRequest{
		BusinessID: "test-business-123",
		ExportType: "unsupported_type",
		Format:     ExportFormatJSON,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/export/data", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.ExportData(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Unsupported export type")
}

func TestExportHandler_ExportData_ExportError(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	mockJobManager := &MockExportJobManager{}
	handler := NewExportHandler(logger, mockExportSvc, mockJobManager)

	// Mock the export service to return an error
	mockExportSvc.On("ExportRiskAssessments", mock.Anything, mock.Anything, ExportFormatJSON).Return(nil, assert.AnError)

	// Create request
	reqBody := CreateExportJobRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/export/data", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.ExportData(w, req)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to export data")

	mockExportSvc.AssertExpectations(t)
}

func TestExportHandler_RegisterRoutes(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	mockJobManager := &MockExportJobManager{}
	handler := NewExportHandler(logger, mockExportSvc, mockJobManager)

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register routes
	handler.RegisterRoutes(mux)

	// Test that routes are registered by making requests
	// Note: This is a basic test to ensure routes are registered
	// In a real implementation, you might want to test the actual route handling

	// Test POST /api/v1/export/jobs
	req := httptest.NewRequest("POST", "/api/v1/export/jobs", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	// Should not return 404 (route not found)
	assert.NotEqual(t, http.StatusNotFound, w.Code)

	// Test GET /api/v1/export/jobs/{job_id}
	req = httptest.NewRequest("GET", "/api/v1/export/jobs/test-job-123", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	// Should not return 404 (route not found)
	assert.NotEqual(t, http.StatusNotFound, w.Code)

	// Test GET /api/v1/export/jobs
	req = httptest.NewRequest("GET", "/api/v1/export/jobs", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	// Should not return 404 (route not found)
	assert.NotEqual(t, http.StatusNotFound, w.Code)

	// Test DELETE /api/v1/export/jobs/{job_id}
	req = httptest.NewRequest("DELETE", "/api/v1/export/jobs/test-job-123", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	// Should not return 404 (route not found)
	assert.NotEqual(t, http.StatusNotFound, w.Code)

	// Test GET /api/v1/export/jobs/statistics
	req = httptest.NewRequest("GET", "/api/v1/export/jobs/statistics", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	// Should not return 404 (route not found)
	assert.NotEqual(t, http.StatusNotFound, w.Code)

	// Test POST /api/v1/export/jobs/cleanup
	req = httptest.NewRequest("POST", "/api/v1/export/jobs/cleanup", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	// Should not return 404 (route not found)
	assert.NotEqual(t, http.StatusNotFound, w.Code)

	// Test POST /api/v1/export/data
	req = httptest.NewRequest("POST", "/api/v1/export/data", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	// Should not return 404 (route not found)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}
