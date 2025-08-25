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

func TestNewDataReportingHandler(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataReportingHandler(logger)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
	assert.NotNil(t, handler.jobs)
	assert.Len(t, handler.jobs, 0)
}

func TestDataReportingHandler_GenerateReport(t *testing.T) {
	tests := []struct {
		name           string
		request        DataReportingRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful PDF report",
			request: DataReportingRequest{
				BusinessID:    "business_123",
				ReportType:    ReportTypeVerificationSummary,
				Format:        ReportFormatPDF,
				Title:         "Verification Summary Report",
				IncludeCharts: true,
				IncludeTables: true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful HTML report",
			request: DataReportingRequest{
				BusinessID:     "business_123",
				ReportType:     ReportTypeAnalytics,
				Format:         ReportFormatHTML,
				Title:          "Analytics Dashboard",
				IncludeCharts:  true,
				IncludeTables:  true,
				IncludeSummary: true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing business_id",
			request: DataReportingRequest{
				ReportType: ReportTypeVerificationSummary,
				Format:     ReportFormatPDF,
				Title:      "Test Report",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "business_id is required",
		},
		{
			name: "missing report_type",
			request: DataReportingRequest{
				BusinessID: "business_123",
				Format:     ReportFormatPDF,
				Title:      "Test Report",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "report_type is required",
		},
		{
			name: "missing format",
			request: DataReportingRequest{
				BusinessID: "business_123",
				ReportType: ReportTypeVerificationSummary,
				Title:      "Test Report",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "format is required",
		},
		{
			name: "missing title",
			request: DataReportingRequest{
				BusinessID: "business_123",
				ReportType: ReportTypeVerificationSummary,
				Format:     ReportFormatPDF,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "title is required",
		},
		{
			name: "unsupported report type",
			request: DataReportingRequest{
				BusinessID: "business_123",
				ReportType: "unsupported",
				Format:     ReportFormatPDF,
				Title:      "Test Report",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "unsupported report type",
		},
		{
			name: "unsupported format",
			request: DataReportingRequest{
				BusinessID: "business_123",
				ReportType: ReportTypeVerificationSummary,
				Format:     "unsupported",
				Title:      "Test Report",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "unsupported format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataReportingHandler(zap.NewNop())

			// Create request body
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			// Create HTTP request
			req := httptest.NewRequest(http.MethodPost, "/v1/reports", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.GenerateReport(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Parse response
				var response DataReportingResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				// Verify response fields
				assert.NotEmpty(t, response.ReportID)
				assert.Equal(t, tt.request.BusinessID, response.BusinessID)
				assert.Equal(t, string(tt.request.ReportType), string(response.Type))
				assert.Equal(t, string(tt.request.Format), string(response.Format))
				assert.Equal(t, tt.request.Title, response.Title)
				assert.Equal(t, "success", response.Status)
				assert.True(t, response.IsSuccessful)
				assert.NotEmpty(t, response.FileURL)
				assert.Greater(t, response.FileSize, int64(0))
				assert.Greater(t, response.PageCount, 0)
				assert.NotZero(t, response.GeneratedAt)
				assert.NotEmpty(t, response.ProcessingTime)
				assert.NotNil(t, response.Summary)
				assert.NotNil(t, response.ExpiresAt)
			} else {
				// Check error message
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestDataReportingHandler_CreateReportJob(t *testing.T) {
	tests := []struct {
		name           string
		request        DataReportingRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful job creation",
			request: DataReportingRequest{
				BusinessID:    "business_123",
				ReportType:    ReportTypeVerificationSummary,
				Format:        ReportFormatPDF,
				Title:         "Verification Summary Report",
				IncludeCharts: true,
				IncludeTables: true,
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "job creation with schedule",
			request: DataReportingRequest{
				BusinessID: "business_123",
				ReportType: ReportTypeAnalytics,
				Format:     ReportFormatHTML,
				Title:      "Analytics Dashboard",
				Schedule: &ReportSchedule{
					Type:      ScheduleTypeDaily,
					StartDate: time.Now(),
					TimeOfDay: "09:00",
					Enabled:   true,
				},
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "validation error",
			request: DataReportingRequest{
				ReportType: ReportTypeVerificationSummary,
				Format:     ReportFormatPDF,
				Title:      "Test Report",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "business_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataReportingHandler(zap.NewNop())

			// Create request body
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			// Create HTTP request
			req := httptest.NewRequest(http.MethodPost, "/v1/reports/jobs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.CreateReportJob(w, req)

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
				assert.Equal(t, string(tt.request.ReportType), response["type"])
				assert.Equal(t, tt.request.Title, response["title"])
				assert.Equal(t, "pending", response["status"])
				assert.Equal(t, float64(0), response["progress"])
				assert.Equal(t, float64(6), response["total_steps"])
				assert.Equal(t, float64(0), response["current_step"])
				assert.Equal(t, "Initializing report generation", response["step_description"])
				assert.NotNil(t, response["created_at"])

				// Verify job was created
				jobID := response["job_id"].(string)
				handler.mutex.RLock()
				job, exists := handler.jobs[jobID]
				handler.mutex.RUnlock()
				assert.True(t, exists)
				assert.Equal(t, tt.request.BusinessID, job.BusinessID)
				assert.Equal(t, tt.request.ReportType, job.Type)
				assert.Equal(t, tt.request.Format, job.Format)
				assert.Equal(t, tt.request.Title, job.Title)
			} else {
				// Check error message
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestDataReportingHandler_GetReportJob(t *testing.T) {
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
			expectedError:  "report job not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataReportingHandler(zap.NewNop())

			// Create test job if needed
			if tt.jobID == "test_job_123" {
				job := &ReportJob{
					JobID:           tt.jobID,
					BusinessID:      "business_123",
					Type:            ReportTypeVerificationSummary,
					Format:          ReportFormatPDF,
					Title:           "Test Report",
					Status:          ReportStatusPending,
					Progress:        0.0,
					TotalSteps:      6,
					CurrentStep:     0,
					StepDescription: "Initializing report generation",
					CreatedAt:       time.Now(),
				}
				handler.mutex.Lock()
				handler.jobs[tt.jobID] = job
				handler.mutex.Unlock()
			}

			// Create HTTP request
			url := "/v1/reports/jobs"
			if tt.jobID != "" {
				url += "?job_id=" + tt.jobID
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.GetReportJob(w, req)

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
				assert.Equal(t, "verification_summary", response["type"])
				assert.Equal(t, "pdf", response["format"])
				assert.Equal(t, "Test Report", response["title"])
				assert.Equal(t, "pending", response["status"])
				assert.Equal(t, float64(0), response["progress"])
				assert.Equal(t, float64(6), response["total_steps"])
				assert.Equal(t, float64(0), response["current_step"])
				assert.Equal(t, "Initializing report generation", response["step_description"])
				assert.NotNil(t, response["created_at"])
			} else {
				// Check error message
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestDataReportingHandler_ListReportJobs(t *testing.T) {
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
			name: "filter by report_type",
			queryParams: map[string]string{
				"report_type": "verification_summary",
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
			handler := NewDataReportingHandler(zap.NewNop())

			// Create test jobs
			testJobs := []*ReportJob{
				{
					JobID:      "job_1",
					BusinessID: "business_123",
					Type:       ReportTypeVerificationSummary,
					Format:     ReportFormatPDF,
					Title:      "Report 1",
					Status:     ReportStatusPending,
					CreatedAt:  time.Now(),
				},
				{
					JobID:      "job_2",
					BusinessID: "business_123",
					Type:       ReportTypeAnalytics,
					Format:     ReportFormatHTML,
					Title:      "Report 2",
					Status:     ReportStatusPending,
					CreatedAt:  time.Now(),
				},
				{
					JobID:      "job_3",
					BusinessID: "business_456",
					Type:       ReportTypeCompliance,
					Format:     ReportFormatPDF,
					Title:      "Report 3",
					Status:     ReportStatusCompleted,
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
			url := "/v1/reports/jobs"
			if query != "" {
				url += "?" + query
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.ListReportJobs(w, req)

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

func TestDataReportingHandler_GetReportTemplate(t *testing.T) {
	tests := []struct {
		name           string
		templateID     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "existing template",
			templateID:     "verification_summary_pdf",
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
			expectedError:  "report template not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataReportingHandler(zap.NewNop())

			// Create HTTP request
			url := "/v1/reports/templates"
			if tt.templateID != "" {
				url += "?template_id=" + tt.templateID
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.GetReportTemplate(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Parse response
				var template ReportTemplate
				err := json.Unmarshal(w.Body.Bytes(), &template)
				require.NoError(t, err)

				// Verify template fields
				assert.Equal(t, tt.templateID, template.ID)
				assert.NotEmpty(t, template.Name)
				assert.NotEmpty(t, template.Description)
				assert.NotEmpty(t, template.Type)
				assert.NotEmpty(t, template.Format)
				assert.NotZero(t, template.CreatedAt)
				assert.NotZero(t, template.UpdatedAt)
			} else {
				// Check error message
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestDataReportingHandler_ListReportTemplates(t *testing.T) {
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
			name: "filter by report_type",
			queryParams: map[string]string{
				"report_type": "verification_summary",
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name: "filter by format",
			queryParams: map[string]string{
				"format": "pdf",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataReportingHandler(zap.NewNop())

			// Build query string
			query := ""
			for key, value := range tt.queryParams {
				if query != "" {
					query += "&"
				}
				query += key + "=" + value
			}

			// Create HTTP request
			url := "/v1/reports/templates"
			if query != "" {
				url += "?" + query
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.ListReportTemplates(w, req)

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

func TestDataReportingHandler_validateReportRequest(t *testing.T) {
	tests := []struct {
		name    string
		request DataReportingRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: DataReportingRequest{
				BusinessID: "business_123",
				ReportType: ReportTypeVerificationSummary,
				Format:     ReportFormatPDF,
				Title:      "Test Report",
			},
			wantErr: false,
		},
		{
			name: "missing business_id",
			request: DataReportingRequest{
				ReportType: ReportTypeVerificationSummary,
				Format:     ReportFormatPDF,
				Title:      "Test Report",
			},
			wantErr: true,
			errMsg:  "business_id is required",
		},
		{
			name: "missing report_type",
			request: DataReportingRequest{
				BusinessID: "business_123",
				Format:     ReportFormatPDF,
				Title:      "Test Report",
			},
			wantErr: true,
			errMsg:  "report_type is required",
		},
		{
			name: "missing format",
			request: DataReportingRequest{
				BusinessID: "business_123",
				ReportType: ReportTypeVerificationSummary,
				Title:      "Test Report",
			},
			wantErr: true,
			errMsg:  "format is required",
		},
		{
			name: "missing title",
			request: DataReportingRequest{
				BusinessID: "business_123",
				ReportType: ReportTypeVerificationSummary,
				Format:     ReportFormatPDF,
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "unsupported report type",
			request: DataReportingRequest{
				BusinessID: "business_123",
				ReportType: "unsupported",
				Format:     ReportFormatPDF,
				Title:      "Test Report",
			},
			wantErr: true,
			errMsg:  "unsupported report type",
		},
		{
			name: "unsupported format",
			request: DataReportingRequest{
				BusinessID: "business_123",
				ReportType: ReportTypeVerificationSummary,
				Format:     "unsupported",
				Title:      "Test Report",
			},
			wantErr: true,
			errMsg:  "unsupported format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataReportingHandler(zap.NewNop())

			err := handler.validateReportRequest(&tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataReportingHandler_validateSchedule(t *testing.T) {
	tests := []struct {
		name     string
		schedule *ReportSchedule
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid daily schedule",
			schedule: &ReportSchedule{
				Type:      ScheduleTypeDaily,
				StartDate: time.Now(),
				TimeOfDay: "09:00",
				Enabled:   true,
			},
			wantErr: false,
		},
		{
			name: "valid weekly schedule",
			schedule: &ReportSchedule{
				Type:      ScheduleTypeWeekly,
				StartDate: time.Now(),
				DayOfWeek: 1, // Monday
				Enabled:   true,
			},
			wantErr: false,
		},
		{
			name: "unsupported schedule type",
			schedule: &ReportSchedule{
				Type:      "unsupported",
				StartDate: time.Now(),
				Enabled:   true,
			},
			wantErr: true,
			errMsg:  "unsupported schedule type",
		},
		{
			name: "missing start_date",
			schedule: &ReportSchedule{
				Type:    ScheduleTypeDaily,
				Enabled: true,
			},
			wantErr: true,
			errMsg:  "start_date is required",
		},
		{
			name: "invalid day_of_week",
			schedule: &ReportSchedule{
				Type:      ScheduleTypeWeekly,
				StartDate: time.Now(),
				DayOfWeek: 10, // Invalid
				Enabled:   true,
			},
			wantErr: true,
			errMsg:  "day_of_week must be between 0 and 6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataReportingHandler(zap.NewNop())

			err := handler.validateSchedule(tt.schedule)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataReportingHandler_generateReportID(t *testing.T) {
	handler := NewDataReportingHandler(zap.NewNop())

	// Generate multiple IDs
	id1 := handler.generateReportID()
	id2 := handler.generateReportID()
	id3 := handler.generateReportID()

	// Verify IDs are unique
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id1, id3)
	assert.NotEqual(t, id2, id3)

	// Verify ID format
	assert.Contains(t, id1, "report_")
	assert.Contains(t, id2, "report_")
	assert.Contains(t, id3, "report_")
}

func TestDataReportingHandler_generateJobID(t *testing.T) {
	handler := NewDataReportingHandler(zap.NewNop())

	// Generate multiple IDs
	id1 := handler.generateJobID()
	id2 := handler.generateJobID()
	id3 := handler.generateJobID()

	// Verify IDs are unique
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id1, id3)
	assert.NotEqual(t, id2, id3)

	// Verify ID format
	assert.Contains(t, id1, "report_job_")
	assert.Contains(t, id2, "report_job_")
	assert.Contains(t, id3, "report_job_")
}

func TestDataReportingHandler_processReport(t *testing.T) {
	handler := NewDataReportingHandler(zap.NewNop())

	request := &DataReportingRequest{
		BusinessID:    "business_123",
		ReportType:    ReportTypeVerificationSummary,
		Format:        ReportFormatPDF,
		Title:         "Test Report",
		IncludeCharts: true,
		IncludeTables: true,
	}

	ctx := context.Background()
	result, err := handler.processReport(ctx, request, "test_report_123")

	// Verify no error
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify result fields
	assert.Equal(t, "test_report_123", result.ReportID)
	assert.Equal(t, "business_123", result.BusinessID)
	assert.Equal(t, ReportTypeVerificationSummary, result.Type)
	assert.Equal(t, ReportFormatPDF, result.Format)
	assert.Equal(t, "Test Report", result.Title)
	assert.Equal(t, "success", result.Status)
	assert.True(t, result.IsSuccessful)
	assert.NotEmpty(t, result.FileURL)
	assert.Greater(t, result.FileSize, int64(0))
	assert.Greater(t, result.PageCount, 0)
	assert.NotZero(t, result.GeneratedAt)
	assert.NotNil(t, result.Summary)
	assert.NotNil(t, result.ExpiresAt)

	// Verify summary
	assert.Greater(t, result.Summary.TotalRecords, 0)
	assert.NotNil(t, result.Summary.DateRange)
	assert.NotEmpty(t, result.Summary.KeyMetrics)
	assert.NotEmpty(t, result.Summary.Charts)
	assert.NotEmpty(t, result.Summary.Tables)
	assert.NotEmpty(t, result.Summary.Recommendations)
}

func TestDataReportingHandler_processReportJob(t *testing.T) {
	handler := NewDataReportingHandler(zap.NewNop())

	job := &ReportJob{
		JobID:           "test_job_123",
		BusinessID:      "business_123",
		Type:            ReportTypeVerificationSummary,
		Format:          ReportFormatPDF,
		Title:           "Test Report",
		Status:          ReportStatusPending,
		Progress:        0.0,
		TotalSteps:      6,
		CurrentStep:     0,
		StepDescription: "Initializing report generation",
		CreatedAt:       time.Now(),
	}

	request := &DataReportingRequest{
		BusinessID:    "business_123",
		ReportType:    ReportTypeVerificationSummary,
		Format:        ReportFormatPDF,
		Title:         "Test Report",
		IncludeCharts: true,
		IncludeTables: true,
	}

	// Store job
	handler.mutex.Lock()
	handler.jobs[job.JobID] = job
	handler.mutex.Unlock()

	// Process job
	handler.processReportJob(job, request)

	// Verify job was updated
	handler.mutex.RLock()
	updatedJob := handler.jobs[job.JobID]
	handler.mutex.RUnlock()

	assert.Equal(t, ReportStatusCompleted, updatedJob.Status)
	assert.Equal(t, 1.0, updatedJob.Progress)
	assert.Equal(t, 6, updatedJob.CurrentStep)
	assert.Equal(t, "Report generation completed successfully", updatedJob.StepDescription)
	assert.NotNil(t, updatedJob.StartedAt)
	assert.NotNil(t, updatedJob.CompletedAt)
	assert.NotNil(t, updatedJob.Result)

	// Verify result
	result := updatedJob.Result
	assert.Equal(t, job.JobID, result.ReportID)
	assert.Equal(t, job.BusinessID, result.BusinessID)
	assert.Equal(t, job.Type, result.Type)
	assert.Equal(t, job.Format, result.Format)
	assert.Equal(t, job.Title, result.Title)
	assert.Equal(t, "success", result.Status)
	assert.True(t, result.IsSuccessful)
	assert.NotEmpty(t, result.FileURL)
	assert.Greater(t, result.FileSize, int64(0))
	assert.Greater(t, result.PageCount, 0)
	assert.NotZero(t, result.GeneratedAt)
	assert.NotNil(t, result.Summary)
	assert.NotNil(t, result.ExpiresAt)
}

func TestDataReportingHandler_updateJobProgress(t *testing.T) {
	handler := NewDataReportingHandler(zap.NewNop())

	job := &ReportJob{
		JobID:           "test_job_123",
		BusinessID:      "business_123",
		Type:            ReportTypeVerificationSummary,
		Format:          ReportFormatPDF,
		Title:           "Test Report",
		Status:          ReportStatusPending,
		Progress:        0.0,
		TotalSteps:      6,
		CurrentStep:     0,
		StepDescription: "Initializing report generation",
		CreatedAt:       time.Now(),
	}

	// Update progress
	handler.updateJobProgress(job, 2, "Processing data")

	// Verify job was updated
	assert.Equal(t, 2, job.CurrentStep)
	assert.Equal(t, "Processing data", job.StepDescription)
	assert.Equal(t, 0.3333333333333333, job.Progress) // 2/6 = 0.333...
}

func TestDataReportingHandler_completeJob(t *testing.T) {
	handler := NewDataReportingHandler(zap.NewNop())

	job := &ReportJob{
		JobID:           "test_job_123",
		BusinessID:      "business_123",
		Type:            ReportTypeVerificationSummary,
		Format:          ReportFormatPDF,
		Title:           "Test Report",
		Status:          ReportStatusProcessing,
		Progress:        0.8,
		TotalSteps:      6,
		CurrentStep:     5,
		StepDescription: "Almost done",
		CreatedAt:       time.Now(),
	}

	request := &DataReportingRequest{
		BusinessID:    "business_123",
		ReportType:    ReportTypeVerificationSummary,
		Format:        ReportFormatPDF,
		Title:         "Test Report",
		IncludeCharts: true,
		IncludeTables: true,
	}

	// Complete job
	handler.completeJob(job, request)

	// Verify job was completed
	assert.Equal(t, ReportStatusCompleted, job.Status)
	assert.Equal(t, 1.0, job.Progress)
	assert.Equal(t, 6, job.CurrentStep)
	assert.Equal(t, "Report generation completed successfully", job.StepDescription)
	assert.NotNil(t, job.Result)
	assert.NotNil(t, job.CompletedAt)

	// Verify result
	result := job.Result
	assert.Equal(t, job.JobID, result.ReportID)
	assert.Equal(t, job.BusinessID, result.BusinessID)
	assert.Equal(t, job.Type, result.Type)
	assert.Equal(t, job.Format, result.Format)
	assert.Equal(t, job.Title, result.Title)
	assert.Equal(t, "success", result.Status)
	assert.True(t, result.IsSuccessful)
	assert.NotEmpty(t, result.FileURL)
	assert.Greater(t, result.FileSize, int64(0))
	assert.Greater(t, result.PageCount, 0)
	assert.NotZero(t, result.GeneratedAt)
	assert.NotNil(t, result.Summary)
	assert.NotNil(t, result.ExpiresAt)
}

func TestDataReportingHandler_calculateNextRunTime(t *testing.T) {
	handler := NewDataReportingHandler(zap.NewNop())

	tests := []struct {
		name     string
		schedule *ReportSchedule
		want     time.Time
	}{
		{
			name: "one time schedule",
			schedule: &ReportSchedule{
				Type:      ScheduleTypeOneTime,
				StartDate: time.Date(2024, 12, 25, 9, 0, 0, 0, time.UTC),
			},
			want: time.Date(2024, 12, 25, 9, 0, 0, 0, time.UTC),
		},
		{
			name: "daily schedule",
			schedule: &ReportSchedule{
				Type:      ScheduleTypeDaily,
				StartDate: time.Now(),
				TimeOfDay: "09:00",
			},
			want: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 9, 0, 0, 0, time.Now().Location()),
		},
		{
			name: "weekly schedule",
			schedule: &ReportSchedule{
				Type:      ScheduleTypeWeekly,
				StartDate: time.Now(),
				DayOfWeek: 1, // Monday
			},
			want: time.Now().AddDate(0, 0, 1), // Simplified expectation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.calculateNextRunTime(tt.schedule)
			assert.NotZero(t, result)
		})
	}
}

func TestDataReportingHandler_getDefaultTemplate(t *testing.T) {
	handler := NewDataReportingHandler(zap.NewNop())

	tests := []struct {
		name       string
		templateID string
		wantNil    bool
	}{
		{
			name:       "existing template",
			templateID: "verification_summary_pdf",
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

func TestDataReportingHandler_getDefaultTemplates(t *testing.T) {
	handler := NewDataReportingHandler(zap.NewNop())

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
		assert.NotZero(t, template.CreatedAt)
		assert.NotZero(t, template.UpdatedAt)
	}

	// Verify specific templates exist
	templateIDs := make(map[string]bool)
	for _, template := range templates {
		templateIDs[template.ID] = true
	}

	assert.True(t, templateIDs["verification_summary_pdf"])
	assert.True(t, templateIDs["analytics_dashboard_html"])
	assert.True(t, templateIDs["compliance_report_pdf"])
	assert.True(t, templateIDs["risk_assessment_excel"])
}

func TestReportType_String(t *testing.T) {
	tests := []struct {
		reportType ReportType
		want       string
	}{
		{ReportTypeVerificationSummary, "verification_summary"},
		{ReportTypeAnalytics, "analytics"},
		{ReportTypeCompliance, "compliance"},
		{ReportTypeRiskAssessment, "risk_assessment"},
		{ReportTypeAuditTrail, "audit_trail"},
		{ReportTypePerformance, "performance"},
		{ReportTypeCustom, "custom"},
	}

	for _, tt := range tests {
		t.Run(string(tt.reportType), func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.reportType))
		})
	}
}

func TestReportFormat_String(t *testing.T) {
	tests := []struct {
		format ReportFormat
		want   string
	}{
		{ReportFormatPDF, "pdf"},
		{ReportFormatHTML, "html"},
		{ReportFormatJSON, "json"},
		{ReportFormatExcel, "excel"},
		{ReportFormatCSV, "csv"},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.format))
		})
	}
}

func TestReportStatus_String(t *testing.T) {
	tests := []struct {
		status ReportStatus
		want   string
	}{
		{ReportStatusPending, "pending"},
		{ReportStatusProcessing, "processing"},
		{ReportStatusCompleted, "completed"},
		{ReportStatusFailed, "failed"},
		{ReportStatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.status))
		})
	}
}

func TestScheduleType_String(t *testing.T) {
	tests := []struct {
		scheduleType ScheduleType
		want         string
	}{
		{ScheduleTypeOneTime, "one_time"},
		{ScheduleTypeDaily, "daily"},
		{ScheduleTypeWeekly, "weekly"},
		{ScheduleTypeMonthly, "monthly"},
		{ScheduleTypeQuarterly, "quarterly"},
		{ScheduleTypeYearly, "yearly"},
	}

	for _, tt := range tests {
		t.Run(string(tt.scheduleType), func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.scheduleType))
		})
	}
}
