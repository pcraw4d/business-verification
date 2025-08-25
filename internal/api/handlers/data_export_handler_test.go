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
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewDataExportHandler(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataExportHandler(logger)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
	assert.NotNil(t, handler.jobs)
	assert.Len(t, handler.jobs, 0)
}

func TestDataExportHandler_ExportData(t *testing.T) {
	tests := []struct {
		name           string
		request        DataExportRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful CSV export",
			request: DataExportRequest{
				BusinessID: "business_123",
				ExportType: ExportTypeVerifications,
				Format:     ExportFormatCSV,
				Columns:    []string{"id", "name", "status"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful JSON export",
			request: DataExportRequest{
				BusinessID:      "business_123",
				ExportType:      ExportTypeAnalytics,
				Format:          ExportFormatJSON,
				IncludeMetadata: true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful Excel export",
			request: DataExportRequest{
				BusinessID: "business_123",
				ExportType: ExportTypeReports,
				Format:     ExportFormatExcel,
				Filters: map[string]interface{}{
					"status": "completed",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing business_id",
			request: DataExportRequest{
				ExportType: ExportTypeVerifications,
				Format:     ExportFormatCSV,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "business_id is required",
		},
		{
			name: "missing export_type",
			request: DataExportRequest{
				BusinessID: "business_123",
				Format:     ExportFormatCSV,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "export_type is required",
		},
		{
			name: "missing format",
			request: DataExportRequest{
				BusinessID: "business_123",
				ExportType: ExportTypeVerifications,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "format is required",
		},
		{
			name: "unsupported format",
			request: DataExportRequest{
				BusinessID: "business_123",
				ExportType: ExportTypeVerifications,
				Format:     "unsupported",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "unsupported format",
		},
		{
			name: "unsupported export type",
			request: DataExportRequest{
				BusinessID: "business_123",
				ExportType: "unsupported",
				Format:     ExportFormatCSV,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "unsupported export type",
		},
		{
			name: "batch size too large",
			request: DataExportRequest{
				BusinessID: "business_123",
				ExportType: ExportTypeVerifications,
				Format:     ExportFormatCSV,
				BatchSize:  15000,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "batch_size cannot exceed 10000",
		},
		{
			name: "max rows too large",
			request: DataExportRequest{
				BusinessID: "business_123",
				ExportType: ExportTypeVerifications,
				Format:     ExportFormatCSV,
				MaxRows:    2000000,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "max_rows cannot exceed 1000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataExportHandler(zap.NewNop())

			// Create request body
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			// Create HTTP request
			req := httptest.NewRequest(http.MethodPost, "/v1/export", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.ExportData(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Parse response
				var response DataExportResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				// Verify response fields
				assert.NotEmpty(t, response.ExportID)
				assert.Equal(t, tt.request.BusinessID, response.BusinessID)
				assert.Equal(t, string(tt.request.ExportType), string(response.Type))
				assert.Equal(t, string(tt.request.Format), string(response.Format))
				assert.Equal(t, "success", response.Status)
				assert.True(t, response.IsSuccessful)
				assert.NotEmpty(t, response.FileURL)
				assert.Greater(t, response.FileSize, int64(0))
				assert.Greater(t, response.RowCount, 0)
				assert.NotEmpty(t, response.Columns)
				assert.NotZero(t, response.GeneratedAt)
				assert.NotEmpty(t, response.ProcessingTime)
			} else {
				// Check error message
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestDataExportHandler_CreateExportJob(t *testing.T) {
	tests := []struct {
		name           string
		request        DataExportRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful job creation",
			request: DataExportRequest{
				BusinessID: "business_123",
				ExportType: ExportTypeVerifications,
				Format:     ExportFormatCSV,
				Columns:    []string{"id", "name", "status"},
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "job creation with filters",
			request: DataExportRequest{
				BusinessID: "business_123",
				ExportType: ExportTypeAnalytics,
				Format:     ExportFormatExcel,
				Filters: map[string]interface{}{
					"status": "completed",
				},
				TimeRange: &TimeRange{
					Start: time.Now().AddDate(0, -1, 0),
					End:   time.Now(),
				},
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "validation error",
			request: DataExportRequest{
				ExportType: ExportTypeVerifications,
				Format:     ExportFormatCSV,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "business_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataExportHandler(zap.NewNop())

			// Create request body
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			// Create HTTP request
			req := httptest.NewRequest(http.MethodPost, "/v1/export/jobs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.CreateExportJob(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusAccepted {
				// Parse response
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				// Verify response fields
				assert.NotEmpty(t, response["job_id"])
				assert.Equal(t, tt.request.BusinessID, response["business_id"])
				assert.Equal(t, string(tt.request.ExportType), response["type"])
				assert.Equal(t, "pending", response["status"])
				assert.Equal(t, float64(0), response["progress"])
				assert.Equal(t, float64(5), response["total_steps"])
				assert.Equal(t, float64(0), response["current_step"])
				assert.Equal(t, "Initializing export job", response["step_description"])
				assert.NotNil(t, response["created_at"])

				// Verify job was created
				jobID := response["job_id"].(string)
				handler.mutex.RLock()
				job, exists := handler.jobs[jobID]
				handler.mutex.RUnlock()
				assert.True(t, exists)
				assert.Equal(t, tt.request.BusinessID, job.BusinessID)
				assert.Equal(t, tt.request.ExportType, job.Type)
				assert.Equal(t, tt.request.Format, job.Format)
			} else {
				// Check error message
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestDataExportHandler_GetExportJob(t *testing.T) {
	tests := []struct {
		name           string
		jobID          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "existing job",
			jobID:          "test_job_123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing job_id",
			jobID:          "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "job_id is required",
		},
		{
			name:           "non-existent job",
			jobID:          "non_existent_job",
			expectedStatus: http.StatusNotFound,
			expectedError:  "export job not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataExportHandler(zap.NewNop())

			// Create test job if needed
			if tt.jobID == "test_job_123" {
				job := &ExportJob{
					JobID:           tt.jobID,
					BusinessID:      "business_123",
					Type:            ExportTypeVerifications,
					Format:          ExportFormatCSV,
					Status:          JobStatusPending,
					Progress:        0.0,
					TotalSteps:      5,
					CurrentStep:     0,
					StepDescription: "Initializing export job",
					CreatedAt:       time.Now(),
				}
				handler.mutex.Lock()
				handler.jobs[tt.jobID] = job
				handler.mutex.Unlock()
			}

			// Create HTTP request
			url := "/v1/export/jobs"
			if tt.jobID != "" {
				url += "?job_id=" + tt.jobID
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.GetExportJob(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Parse response
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				// Verify response fields
				assert.Equal(t, tt.jobID, response["job_id"])
				assert.Equal(t, "business_123", response["business_id"])
				assert.Equal(t, "verifications", response["type"])
				assert.Equal(t, "csv", response["format"])
				assert.Equal(t, "pending", response["status"])
				assert.Equal(t, float64(0), response["progress"])
				assert.Equal(t, float64(5), response["total_steps"])
				assert.Equal(t, float64(0), response["current_step"])
				assert.Equal(t, "Initializing export job", response["step_description"])
				assert.NotNil(t, response["created_at"])
			} else {
				// Check error message
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestDataExportHandler_ListExportJobs(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "list all jobs",
			queryParams:    map[string]string{},
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name: "filter by status",
			queryParams: map[string]string{
				"status": "pending",
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "filter by business_id",
			queryParams: map[string]string{
				"business_id": "business_123",
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "pagination",
			queryParams: map[string]string{
				"limit":  "2",
				"offset": "1",
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "invalid limit",
			queryParams: map[string]string{
				"limit": "invalid",
			},
			expectedStatus: http.StatusOK,
			expectedCount:  3, // Should use default limit
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataExportHandler(zap.NewNop())

			// Create test jobs
			testJobs := []*ExportJob{
				{
					JobID:      "job_1",
					BusinessID: "business_123",
					Type:       ExportTypeVerifications,
					Format:     ExportFormatCSV,
					Status:     JobStatusPending,
					CreatedAt:  time.Now(),
				},
				{
					JobID:      "job_2",
					BusinessID: "business_123",
					Type:       ExportTypeAnalytics,
					Format:     ExportFormatExcel,
					Status:     JobStatusPending,
					CreatedAt:  time.Now(),
				},
				{
					JobID:      "job_3",
					BusinessID: "business_456",
					Type:       ExportTypeReports,
					Format:     ExportFormatPDF,
					Status:     JobStatusCompleted,
					CreatedAt:  time.Now(),
				},
			}

			handler.mutex.Lock()
			for _, job := range testJobs {
				handler.jobs[job.JobID] = job
			}
			handler.mutex.Unlock()

			// Build query string
			query := ""
			for key, value := range tt.queryParams {
				if query != "" {
					query += "&"
				}
				query += key + "=" + value
			}

			// Create HTTP request
			url := "/v1/export/jobs"
			if query != "" {
				url += "?" + query
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.ListExportJobs(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Verify response structure
			assert.NotNil(t, response["jobs"])
			assert.NotNil(t, response["total_count"])
			assert.NotNil(t, response["limit"])
			assert.NotNil(t, response["offset"])

			// Verify job count
			jobs := response["jobs"].([]interface{})
			assert.Len(t, jobs, tt.expectedCount)
		})
	}
}

func TestDataExportHandler_GetExportTemplate(t *testing.T) {
	tests := []struct {
		name           string
		templateID     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "existing template",
			templateID:     "verifications_csv",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing template_id",
			templateID:     "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "template_id is required",
		},
		{
			name:           "non-existent template",
			templateID:     "non_existent_template",
			expectedStatus: http.StatusNotFound,
			expectedError:  "export template not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataExportHandler(zap.NewNop())

			// Create HTTP request
			url := "/v1/export/templates"
			if tt.templateID != "" {
				url += "?template_id=" + tt.templateID
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.GetExportTemplate(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Parse response
				var template ExportTemplate
				err := json.Unmarshal(w.Body.Bytes(), &template)
				require.NoError(t, err)

				// Verify template fields
				assert.Equal(t, tt.templateID, template.ID)
				assert.NotEmpty(t, template.Name)
				assert.NotEmpty(t, template.Description)
				assert.NotEmpty(t, template.Type)
				assert.NotEmpty(t, template.Format)
				assert.NotEmpty(t, template.Columns)
				assert.NotZero(t, template.CreatedAt)
				assert.NotZero(t, template.UpdatedAt)
			} else {
				// Check error message
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestDataExportHandler_ListExportTemplates(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "list all templates",
			queryParams:    map[string]string{},
			expectedStatus: http.StatusOK,
			expectedCount:  4,
		},
		{
			name: "filter by type",
			queryParams: map[string]string{
				"type": "verifications",
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name: "filter by format",
			queryParams: map[string]string{
				"format": "csv",
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name: "pagination",
			queryParams: map[string]string{
				"limit":  "2",
				"offset": "1",
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataExportHandler(zap.NewNop())

			// Build query string
			query := ""
			for key, value := range tt.queryParams {
				if query != "" {
					query += "&"
				}
				query += key + "=" + value
			}

			// Create HTTP request
			url := "/v1/export/templates"
			if query != "" {
				url += "?" + query
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.ListExportTemplates(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Verify response structure
			assert.NotNil(t, response["templates"])
			assert.NotNil(t, response["total_count"])
			assert.NotNil(t, response["limit"])
			assert.NotNil(t, response["offset"])

			// Verify template count
			templates := response["templates"].([]interface{})
			assert.Len(t, templates, tt.expectedCount)
		})
	}
}

func TestDataExportHandler_validateExportRequest(t *testing.T) {
	tests := []struct {
		name    string
		request DataExportRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: DataExportRequest{
				BusinessID: "business_123",
				ExportType: ExportTypeVerifications,
				Format:     ExportFormatCSV,
			},
			wantErr: false,
		},
		{
			name: "missing business_id",
			request: DataExportRequest{
				ExportType: ExportTypeVerifications,
				Format:     ExportFormatCSV,
			},
			wantErr: true,
			errMsg:  "business_id is required",
		},
		{
			name: "missing export_type",
			request: DataExportRequest{
				BusinessID: "business_123",
				Format:     ExportFormatCSV,
			},
			wantErr: true,
			errMsg:  "export_type is required",
		},
		{
			name: "missing format",
			request: DataExportRequest{
				BusinessID: "business_123",
				ExportType: ExportTypeVerifications,
			},
			wantErr: true,
			errMsg:  "format is required",
		},
		{
			name: "unsupported format",
			request: DataExportRequest{
				BusinessID: "business_123",
				ExportType: ExportTypeVerifications,
				Format:     "unsupported",
			},
			wantErr: true,
			errMsg:  "unsupported format",
		},
		{
			name: "unsupported export type",
			request: DataExportRequest{
				BusinessID: "business_123",
				ExportType: "unsupported",
				Format:     ExportFormatCSV,
			},
			wantErr: true,
			errMsg:  "unsupported export type",
		},
		{
			name: "batch size too large",
			request: DataExportRequest{
				BusinessID: "business_123",
				ExportType: ExportTypeVerifications,
				Format:     ExportFormatCSV,
				BatchSize:  15000,
			},
			wantErr: true,
			errMsg:  "batch_size cannot exceed 10000",
		},
		{
			name: "max rows too large",
			request: DataExportRequest{
				BusinessID: "business_123",
				ExportType: ExportTypeVerifications,
				Format:     ExportFormatCSV,
				MaxRows:    2000000,
			},
			wantErr: true,
			errMsg:  "max_rows cannot exceed 1000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataExportHandler(zap.NewNop())

			err := handler.validateExportRequest(&tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataExportHandler_generateExportID(t *testing.T) {
	handler := NewDataExportHandler(zap.NewNop())

	// Generate multiple IDs
	id1 := handler.generateExportID()
	id2 := handler.generateExportID()
	id3 := handler.generateExportID()

	// Verify IDs are unique
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id1, id3)
	assert.NotEqual(t, id2, id3)

	// Verify ID format
	assert.Contains(t, id1, "export_")
	assert.Contains(t, id2, "export_")
	assert.Contains(t, id3, "export_")
}

func TestDataExportHandler_generateJobID(t *testing.T) {
	handler := NewDataExportHandler(zap.NewNop())

	// Generate multiple IDs
	id1 := handler.generateJobID()
	id2 := handler.generateJobID()
	id3 := handler.generateJobID()

	// Verify IDs are unique
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id1, id3)
	assert.NotEqual(t, id2, id3)

	// Verify ID format
	assert.Contains(t, id1, "export_job_")
	assert.Contains(t, id2, "export_job_")
	assert.Contains(t, id3, "export_job_")
}

func TestDataExportHandler_processExport(t *testing.T) {
	handler := NewDataExportHandler(zap.NewNop())

	request := &DataExportRequest{
		BusinessID: "business_123",
		ExportType: ExportTypeVerifications,
		Format:     ExportFormatCSV,
		Columns:    []string{"id", "name", "status"},
	}

	ctx := context.Background()
	result, err := handler.processExport(ctx, request, "test_export_123")

	// Verify no error
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify result fields
	assert.Equal(t, "test_export_123", result.ExportID)
	assert.Equal(t, "business_123", result.BusinessID)
	assert.Equal(t, ExportTypeVerifications, result.Type)
	assert.Equal(t, ExportFormatCSV, result.Format)
	assert.Equal(t, "success", result.Status)
	assert.True(t, result.IsSuccessful)
	assert.NotEmpty(t, result.FileURL)
	assert.Greater(t, result.FileSize, int64(0))
	assert.Greater(t, result.RowCount, 0)
	assert.NotEmpty(t, result.Columns)
	assert.NotNil(t, result.ExpiresAt)
	assert.NotZero(t, result.GeneratedAt)
}

func TestDataExportHandler_processExportJob(t *testing.T) {
	handler := NewDataExportHandler(zap.NewNop())

	job := &ExportJob{
		JobID:           "test_job_123",
		BusinessID:      "business_123",
		Type:            ExportTypeVerifications,
		Format:          ExportFormatCSV,
		Status:          JobStatusPending,
		Progress:        0.0,
		TotalSteps:      5,
		CurrentStep:     0,
		StepDescription: "Initializing export job",
		CreatedAt:       time.Now(),
	}

	request := &DataExportRequest{
		BusinessID: "business_123",
		ExportType: ExportTypeVerifications,
		Format:     ExportFormatCSV,
		Columns:    []string{"id", "name", "status"},
	}

	// Store job
	handler.mutex.Lock()
	handler.jobs[job.JobID] = job
	handler.mutex.Unlock()

	// Process job
	handler.processExportJob(job, request)

	// Verify job was updated
	handler.mutex.RLock()
	updatedJob := handler.jobs[job.JobID]
	handler.mutex.RUnlock()

	assert.Equal(t, JobStatusCompleted, updatedJob.Status)
	assert.Equal(t, 1.0, updatedJob.Progress)
	assert.Equal(t, 5, updatedJob.CurrentStep)
	assert.Equal(t, "Export completed successfully", updatedJob.StepDescription)
	assert.NotNil(t, updatedJob.StartedAt)
	assert.NotNil(t, updatedJob.CompletedAt)
	assert.NotNil(t, updatedJob.Result)

	// Verify result
	result := updatedJob.Result
	assert.Equal(t, job.JobID, result.ExportID)
	assert.Equal(t, job.BusinessID, result.BusinessID)
	assert.Equal(t, job.Type, result.Type)
	assert.Equal(t, job.Format, result.Format)
	assert.Equal(t, "success", result.Status)
	assert.True(t, result.IsSuccessful)
	assert.NotEmpty(t, result.FileURL)
	assert.Greater(t, result.FileSize, int64(0))
	assert.Greater(t, result.RowCount, 0)
	assert.NotEmpty(t, result.Columns)
	assert.NotNil(t, result.ExpiresAt)
	assert.NotZero(t, result.GeneratedAt)
}

func TestDataExportHandler_updateJobProgress(t *testing.T) {
	handler := NewDataExportHandler(zap.NewNop())

	job := &ExportJob{
		JobID:           "test_job_123",
		BusinessID:      "business_123",
		Type:            ExportTypeVerifications,
		Format:          ExportFormatCSV,
		Status:          JobStatusPending,
		Progress:        0.0,
		TotalSteps:      5,
		CurrentStep:     0,
		StepDescription: "Initializing export job",
		CreatedAt:       time.Now(),
	}

	// Update progress
	handler.updateJobProgress(job, 2, "Processing data")

	// Verify job was updated
	assert.Equal(t, 2, job.CurrentStep)
	assert.Equal(t, "Processing data", job.StepDescription)
	assert.Equal(t, 0.4, job.Progress) // 2/5 = 0.4
}

func TestDataExportHandler_completeJob(t *testing.T) {
	handler := NewDataExportHandler(zap.NewNop())

	job := &ExportJob{
		JobID:           "test_job_123",
		BusinessID:      "business_123",
		Type:            ExportTypeVerifications,
		Format:          ExportFormatCSV,
		Status:          JobStatusProcessing,
		Progress:        0.8,
		TotalSteps:      5,
		CurrentStep:     4,
		StepDescription: "Almost done",
		CreatedAt:       time.Now(),
	}

	request := &DataExportRequest{
		BusinessID: "business_123",
		ExportType: ExportTypeVerifications,
		Format:     ExportFormatCSV,
		Columns:    []string{"id", "name", "status"},
	}

	// Complete job
	handler.completeJob(job, request)

	// Verify job was completed
	assert.Equal(t, JobStatusCompleted, job.Status)
	assert.Equal(t, 1.0, job.Progress)
	assert.Equal(t, 5, job.CurrentStep)
	assert.Equal(t, "Export completed successfully", job.StepDescription)
	assert.NotNil(t, job.Result)
	assert.NotNil(t, job.CompletedAt)

	// Verify result
	result := job.Result
	assert.Equal(t, job.JobID, result.ExportID)
	assert.Equal(t, job.BusinessID, result.BusinessID)
	assert.Equal(t, job.Type, result.Type)
	assert.Equal(t, job.Format, result.Format)
	assert.Equal(t, "success", result.Status)
	assert.True(t, result.IsSuccessful)
	assert.NotEmpty(t, result.FileURL)
	assert.Greater(t, result.FileSize, int64(0))
	assert.Greater(t, result.RowCount, 0)
	assert.NotEmpty(t, result.Columns)
	assert.NotNil(t, result.ExpiresAt)
	assert.NotZero(t, result.GeneratedAt)
}

func TestDataExportHandler_getDefaultTemplate(t *testing.T) {
	handler := NewDataExportHandler(zap.NewNop())

	tests := []struct {
		name       string
		templateID string
		wantNil    bool
	}{
		{
			name:       "existing template",
			templateID: "verifications_csv",
			wantNil:    false,
		},
		{
			name:       "non-existent template",
			templateID: "non_existent",
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := handler.getDefaultTemplate(tt.templateID)

			if tt.wantNil {
				assert.Nil(t, template)
			} else {
				assert.NotNil(t, template)
				assert.Equal(t, tt.templateID, template.ID)
			}
		})
	}
}

func TestDataExportHandler_getDefaultTemplates(t *testing.T) {
	handler := NewDataExportHandler(zap.NewNop())

	templates := handler.getDefaultTemplates()

	// Verify templates were returned
	assert.Len(t, templates, 4)

	// Verify each template has required fields
	for _, template := range templates {
		assert.NotEmpty(t, template.ID)
		assert.NotEmpty(t, template.Name)
		assert.NotEmpty(t, template.Description)
		assert.NotEmpty(t, template.Type)
		assert.NotEmpty(t, template.Format)
		assert.NotEmpty(t, template.Columns)
		assert.NotZero(t, template.CreatedAt)
		assert.NotZero(t, template.UpdatedAt)
	}

	// Verify specific templates exist
	templateIDs := make(map[string]bool)
	for _, template := range templates {
		templateIDs[template.ID] = true
	}

	assert.True(t, templateIDs["verifications_csv"])
	assert.True(t, templateIDs["analytics_excel"])
	assert.True(t, templateIDs["audit_logs_json"])
	assert.True(t, templateIDs["business_data_pdf"])
}

func TestExportFormat_String(t *testing.T) {
	tests := []struct {
		format ExportFormat
		want   string
	}{
		{ExportFormatCSV, "csv"},
		{ExportFormatJSON, "json"},
		{ExportFormatExcel, "excel"},
		{ExportFormatPDF, "pdf"},
		{ExportFormatXML, "xml"},
		{ExportFormatTSV, "tsv"},
		{ExportFormatYAML, "yaml"},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.format))
		})
	}
}

func TestExportType_String(t *testing.T) {
	tests := []struct {
		exportType ExportType
		want       string
	}{
		{ExportTypeVerifications, "verifications"},
		{ExportTypeAnalytics, "analytics"},
		{ExportTypeReports, "reports"},
		{ExportTypeAuditLogs, "audit_logs"},
		{ExportTypeUserData, "user_data"},
		{ExportTypeBusinessData, "business_data"},
		{ExportTypeCustom, "custom"},
	}

	for _, tt := range tests {
		t.Run(string(tt.exportType), func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.exportType))
		})
	}
}

func TestJobStatus_String(t *testing.T) {
	tests := []struct {
		status JobStatus
		want   string
	}{
		{JobStatusPending, "pending"},
		{JobStatusProcessing, "processing"},
		{JobStatusCompleted, "completed"},
		{JobStatusFailed, "failed"},
		{JobStatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.status))
		})
	}
}
