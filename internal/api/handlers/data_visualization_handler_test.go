package handlers

import (
	"bytes"
	"context"
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

func TestNewDataVisualizationHandler(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataVisualizationHandler(logger)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
	assert.NotNil(t, handler.jobs)
	assert.Len(t, handler.jobs, 0)
}

func TestDataVisualizationHandler_GenerateVisualization(t *testing.T) {
	tests := []struct {
		name           string
		request        DataVisualizationRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful line chart generation",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeLineChart,
				ChartType:         ChartTypeLine,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful bar chart generation",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeBarChart,
				ChartType:         ChartTypeBar,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful pie chart generation",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypePieChart,
				ChartType:         ChartTypePie,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful area chart generation",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeAreaChart,
				ChartType:         ChartTypeArea,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful scatter plot generation",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeScatterPlot,
				ChartType:         ChartTypeScatter,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful heatmap generation",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeHeatmap,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful gauge generation",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeGauge,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful table generation",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeTable,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful KPI generation",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeKPI,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful dashboard generation",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeDashboard,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing visualization type",
			request: DataVisualizationRequest{
				Data: map[string]interface{}{"test": "data"},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "missing data",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeLineChart,
				ChartType:         ChartTypeLine,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "missing chart type for line chart",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeLineChart,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "unsupported visualization type",
			request: DataVisualizationRequest{
				VisualizationType: "unsupported",
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "VISUALIZATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataVisualizationHandler(zap.NewNop())

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/v1/visualize", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.GenerateVisualization(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"].(map[string]interface{})["code"], tt.expectedError)
			} else {
				var response DataVisualizationResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "success", response.Status)
				assert.True(t, response.IsSuccessful)
				assert.NotEmpty(t, response.VisualizationID)
				assert.Equal(t, tt.request.VisualizationType, response.Type)
				assert.NotNil(t, response.Data)
			}
		})
	}
}

func TestDataVisualizationHandler_CreateVisualizationJob(t *testing.T) {
	tests := []struct {
		name           string
		request        DataVisualizationRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful job creation",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeLineChart,
				ChartType:         ChartTypeLine,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "invalid request",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeLineChart,
				// Missing chart type and data
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataVisualizationHandler(zap.NewNop())

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/v1/visualize/jobs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.CreateVisualizationJob(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"].(map[string]interface{})["code"], tt.expectedError)
			} else {
				var job BackgroundVisualizationJob
				err := json.Unmarshal(w.Body.Bytes(), &job)
				require.NoError(t, err)
				assert.NotEmpty(t, job.JobID)
				assert.Equal(t, "pending", job.Status)
				assert.Equal(t, 0.0, job.Progress)
				assert.Equal(t, 5, job.TotalSteps)
				assert.Equal(t, 0, job.CurrentStep)
				assert.Equal(t, "Initializing visualization job", job.StepDescription)
			}
		})
	}
}

func TestDataVisualizationHandler_GetVisualizationJob(t *testing.T) {
	tests := []struct {
		name           string
		jobID          string
		setupJob       bool
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful job retrieval",
			jobID:          "test_job_1",
			setupJob:       true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing job ID",
			jobID:          "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_JOB_ID",
		},
		{
			name:           "job not found",
			jobID:          "nonexistent_job",
			expectedStatus: http.StatusNotFound,
			expectedError:  "JOB_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataVisualizationHandler(zap.NewNop())

			if tt.setupJob {
				job := &BackgroundVisualizationJob{
					JobID:           tt.jobID,
					Status:          "completed",
					Progress:        1.0,
					TotalSteps:      5,
					CurrentStep:     5,
					StepDescription: "Completed",
					CreatedAt:       time.Now(),
				}
				handler.jobs[tt.jobID] = job
			}

			url := fmt.Sprintf("/v1/visualize/jobs?job_id=%s", tt.jobID)
			req := httptest.NewRequest(http.MethodGet, url, nil)

			w := httptest.NewRecorder()
			handler.GetVisualizationJob(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"].(map[string]interface{})["code"], tt.expectedError)
			} else {
				var job BackgroundVisualizationJob
				err := json.Unmarshal(w.Body.Bytes(), &job)
				require.NoError(t, err)
				assert.Equal(t, tt.jobID, job.JobID)
				assert.Equal(t, "completed", job.Status)
			}
		})
	}
}

func TestDataVisualizationHandler_ListVisualizationJobs(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupJobs      []*BackgroundVisualizationJob
		expectedStatus int
		expectedCount  int
	}{
		{
			name:        "list all jobs",
			queryParams: "",
			setupJobs: []*BackgroundVisualizationJob{
				{JobID: "job_1", Status: "completed", CreatedAt: time.Now()},
				{JobID: "job_2", Status: "pending", CreatedAt: time.Now()},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:        "filter by status",
			queryParams: "?status=completed",
			setupJobs: []*BackgroundVisualizationJob{
				{JobID: "job_1", Status: "completed", CreatedAt: time.Now()},
				{JobID: "job_2", Status: "pending", CreatedAt: time.Now()},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:        "filter by business ID",
			queryParams: "?business_id=business_123",
			setupJobs: []*BackgroundVisualizationJob{
				{JobID: "job_1", BusinessID: "business_123", Status: "completed", CreatedAt: time.Now()},
				{JobID: "job_2", BusinessID: "business_456", Status: "pending", CreatedAt: time.Now()},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:        "pagination",
			queryParams: "?limit=1&offset=0",
			setupJobs: []*BackgroundVisualizationJob{
				{JobID: "job_1", Status: "completed", CreatedAt: time.Now()},
				{JobID: "job_2", Status: "pending", CreatedAt: time.Now()},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataVisualizationHandler(zap.NewNop())

			for _, job := range tt.setupJobs {
				handler.jobs[job.JobID] = job
			}

			url := fmt.Sprintf("/v1/visualize/jobs%s", tt.queryParams)
			req := httptest.NewRequest(http.MethodGet, url, nil)

			w := httptest.NewRecorder()
			handler.ListVisualizationJobs(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			jobs := response["jobs"].([]interface{})
			assert.Len(t, jobs, tt.expectedCount)
		})
	}
}

func TestDataVisualizationHandler_GetVisualizationSchema(t *testing.T) {
	tests := []struct {
		name           string
		schemaID       string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful schema retrieval",
			schemaID:       "default_line_chart",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing schema ID",
			schemaID:       "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_SCHEMA_ID",
		},
		{
			name:           "schema not found",
			schemaID:       "nonexistent_schema",
			expectedStatus: http.StatusNotFound,
			expectedError:  "SCHEMA_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataVisualizationHandler(zap.NewNop())

			url := fmt.Sprintf("/v1/visualize/schemas?schema_id=%s", tt.schemaID)
			req := httptest.NewRequest(http.MethodGet, url, nil)

			w := httptest.NewRecorder()
			handler.GetVisualizationSchema(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"].(map[string]interface{})["code"], tt.expectedError)
			} else {
				var schema VisualizationSchema
				err := json.Unmarshal(w.Body.Bytes(), &schema)
				require.NoError(t, err)
				assert.Equal(t, tt.schemaID, schema.ID)
				assert.Equal(t, "Default Line Chart", schema.Name)
			}
		})
	}
}

func TestDataVisualizationHandler_ListVisualizationSchemas(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "list all schemas",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "filter by type",
			queryParams:    "?type=line_chart",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "filter by chart type",
			queryParams:    "?chart_type=line",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "pagination",
			queryParams:    "?limit=1&offset=0",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataVisualizationHandler(zap.NewNop())

			url := fmt.Sprintf("/v1/visualize/schemas%s", tt.queryParams)
			req := httptest.NewRequest(http.MethodGet, url, nil)

			w := httptest.NewRecorder()
			handler.ListVisualizationSchemas(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			schemas := response["schemas"].([]interface{})
			assert.Len(t, schemas, tt.expectedCount)
		})
	}
}

func TestDataVisualizationHandler_GenerateDashboard(t *testing.T) {
	tests := []struct {
		name           string
		request        map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful dashboard generation",
			request: map[string]interface{}{
				"business_id": "business_123",
				"widgets": []map[string]interface{}{
					{
						"id":    "widget_1",
						"type":  "kpi",
						"title": "Test Widget",
					},
				},
				"theme": "light",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			request:        nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataVisualizationHandler(zap.NewNop())

			var body []byte
			if tt.request != nil {
				body, _ = json.Marshal(tt.request)
			} else {
				body = []byte("invalid json")
			}

			req := httptest.NewRequest(http.MethodPost, "/v1/visualize/dashboard", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.GenerateDashboard(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"].(map[string]interface{})["code"], tt.expectedError)
			} else {
				var dashboard map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &dashboard)
				require.NoError(t, err)
				assert.NotEmpty(t, dashboard["dashboard_id"])
				assert.Equal(t, "KYB Platform Dashboard", dashboard["title"])
				assert.NotNil(t, dashboard["layout"])
			}
		})
	}
}

func TestDataVisualizationHandler_validateVisualizationRequest(t *testing.T) {
	tests := []struct {
		name    string
		request DataVisualizationRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeLineChart,
				ChartType:         ChartTypeLine,
				Data:              map[string]interface{}{"test": "data"},
			},
			wantErr: false,
		},
		{
			name: "missing visualization type",
			request: DataVisualizationRequest{
				ChartType: ChartTypeLine,
				Data:      map[string]interface{}{"test": "data"},
			},
			wantErr: true,
			errMsg:  "visualization_type is required",
		},
		{
			name: "missing data",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeLineChart,
				ChartType:         ChartTypeLine,
			},
			wantErr: true,
			errMsg:  "data is required",
		},
		{
			name: "missing chart type for line chart",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeLineChart,
				Data:              map[string]interface{}{"test": "data"},
			},
			wantErr: true,
			errMsg:  "chart_type is required for chart visualizations",
		},
		{
			name: "missing chart type for bar chart",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeBarChart,
				Data:              map[string]interface{}{"test": "data"},
			},
			wantErr: true,
			errMsg:  "chart_type is required for chart visualizations",
		},
		{
			name: "missing chart type for pie chart",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypePieChart,
				Data:              map[string]interface{}{"test": "data"},
			},
			wantErr: true,
			errMsg:  "chart_type is required for chart visualizations",
		},
		{
			name: "missing chart type for area chart",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeAreaChart,
				Data:              map[string]interface{}{"test": "data"},
			},
			wantErr: true,
			errMsg:  "chart_type is required for chart visualizations",
		},
		{
			name: "missing chart type for scatter plot",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeScatterPlot,
				Data:              map[string]interface{}{"test": "data"},
			},
			wantErr: true,
			errMsg:  "chart_type is required for chart visualizations",
		},
		{
			name: "valid heatmap without chart type",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeHeatmap,
				Data:              map[string]interface{}{"test": "data"},
			},
			wantErr: false,
		},
		{
			name: "valid gauge without chart type",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeGauge,
				Data:              map[string]interface{}{"test": "data"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataVisualizationHandler(zap.NewNop())

			err := handler.validateVisualizationRequest(&tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataVisualizationHandler_generateVisualization(t *testing.T) {
	tests := []struct {
		name          string
		request       DataVisualizationRequest
		expectedType  VisualizationType
		expectedError bool
	}{
		{
			name: "generate line chart",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeLineChart,
				ChartType:         ChartTypeLine,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedType:  VisualizationTypeLineChart,
			expectedError: false,
		},
		{
			name: "generate bar chart",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeBarChart,
				ChartType:         ChartTypeBar,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedType:  VisualizationTypeBarChart,
			expectedError: false,
		},
		{
			name: "generate pie chart",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypePieChart,
				ChartType:         ChartTypePie,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedType:  VisualizationTypePieChart,
			expectedError: false,
		},
		{
			name: "generate area chart",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeAreaChart,
				ChartType:         ChartTypeArea,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedType:  VisualizationTypeAreaChart,
			expectedError: false,
		},
		{
			name: "generate scatter plot",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeScatterPlot,
				ChartType:         ChartTypeScatter,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedType:  VisualizationTypeScatterPlot,
			expectedError: false,
		},
		{
			name: "generate heatmap",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeHeatmap,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedType:  VisualizationTypeHeatmap,
			expectedError: false,
		},
		{
			name: "generate gauge",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeGauge,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedType:  VisualizationTypeGauge,
			expectedError: false,
		},
		{
			name: "generate table",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeTable,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedType:  VisualizationTypeTable,
			expectedError: false,
		},
		{
			name: "generate KPI",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeKPI,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedType:  VisualizationTypeKPI,
			expectedError: false,
		},
		{
			name: "generate dashboard",
			request: DataVisualizationRequest{
				VisualizationType: VisualizationTypeDashboard,
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedType:  VisualizationTypeDashboard,
			expectedError: false,
		},
		{
			name: "unsupported visualization type",
			request: DataVisualizationRequest{
				VisualizationType: "unsupported",
				Data:              map[string]interface{}{"test": "data"},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataVisualizationHandler(zap.NewNop())

			result, err := handler.generateVisualization(context.Background(), &tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedType, result.Type)
				assert.Equal(t, "success", result.Status)
				assert.True(t, result.IsSuccessful)
				assert.NotEmpty(t, result.VisualizationID)
				assert.NotNil(t, result.Data)
			}
		})
	}
}

func TestDataVisualizationHandler_generateLineChart(t *testing.T) {
	handler := NewDataVisualizationHandler(zap.NewNop())
	req := &DataVisualizationRequest{
		VisualizationType: VisualizationTypeLineChart,
		ChartType:         ChartTypeLine,
		Data:              map[string]interface{}{"test": "data"},
	}

	result, err := handler.generateLineChart(req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Labels, 6)
	assert.Len(t, result.Datasets, 2)
	assert.Equal(t, "Verifications", result.Datasets[0].Label)
	assert.Equal(t, "Success Rate", result.Datasets[1].Label)
}

func TestDataVisualizationHandler_generateBarChart(t *testing.T) {
	handler := NewDataVisualizationHandler(zap.NewNop())
	req := &DataVisualizationRequest{
		VisualizationType: VisualizationTypeBarChart,
		ChartType:         ChartTypeBar,
		Data:              map[string]interface{}{"test": "data"},
	}

	result, err := handler.generateBarChart(req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Labels, 5)
	assert.Len(t, result.Datasets, 1)
	assert.Equal(t, "Risk Score", result.Datasets[0].Label)
}

func TestDataVisualizationHandler_generatePieChart(t *testing.T) {
	handler := NewDataVisualizationHandler(zap.NewNop())
	req := &DataVisualizationRequest{
		VisualizationType: VisualizationTypePieChart,
		ChartType:         ChartTypePie,
		Data:              map[string]interface{}{"test": "data"},
	}

	result, err := handler.generatePieChart(req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Labels, 4)
	assert.Len(t, result.Datasets, 1)
	assert.Equal(t, "Risk Distribution", result.Datasets[0].Label)
}

func TestDataVisualizationHandler_generateAreaChart(t *testing.T) {
	handler := NewDataVisualizationHandler(zap.NewNop())
	req := &DataVisualizationRequest{
		VisualizationType: VisualizationTypeAreaChart,
		ChartType:         ChartTypeArea,
		Data:              map[string]interface{}{"test": "data"},
	}

	result, err := handler.generateAreaChart(req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Labels, 4)
	assert.Len(t, result.Datasets, 1)
	assert.Equal(t, "Revenue", result.Datasets[0].Label)
}

func TestDataVisualizationHandler_generateScatterPlot(t *testing.T) {
	handler := NewDataVisualizationHandler(zap.NewNop())
	req := &DataVisualizationRequest{
		VisualizationType: VisualizationTypeScatterPlot,
		ChartType:         ChartTypeScatter,
		Data:              map[string]interface{}{"test": "data"},
	}

	result, err := handler.generateScatterPlot(req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Datasets, 1)
	assert.Equal(t, "Risk vs Performance", result.Datasets[0].Label)
}

func TestDataVisualizationHandler_generateHeatmap(t *testing.T) {
	handler := NewDataVisualizationHandler(zap.NewNop())
	req := &DataVisualizationRequest{
		VisualizationType: VisualizationTypeHeatmap,
		Data:              map[string]interface{}{"test": "data"},
	}

	result, err := handler.generateHeatmap(req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "data")
	assert.Contains(t, result, "xLabels")
	assert.Contains(t, result, "yLabels")
	assert.Contains(t, result, "colors")
}

func TestDataVisualizationHandler_generateGauge(t *testing.T) {
	handler := NewDataVisualizationHandler(zap.NewNop())
	req := &DataVisualizationRequest{
		VisualizationType: VisualizationTypeGauge,
		Data:              map[string]interface{}{"test": "data"},
	}

	result, err := handler.generateGauge(req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 75.5, result["value"])
	assert.Equal(t, "Success Rate", result["label"])
	assert.Equal(t, "%", result["unit"])
}

func TestDataVisualizationHandler_generateTable(t *testing.T) {
	handler := NewDataVisualizationHandler(zap.NewNop())
	req := &DataVisualizationRequest{
		VisualizationType: VisualizationTypeTable,
		Data:              map[string]interface{}{"test": "data"},
	}

	result, err := handler.generateTable(req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "headers")
	assert.Contains(t, result, "rows")
	assert.Contains(t, result, "total_rows")
	assert.Equal(t, true, result["sortable"])
	assert.Equal(t, true, result["filterable"])
}

func TestDataVisualizationHandler_generateKPI(t *testing.T) {
	handler := NewDataVisualizationHandler(zap.NewNop())
	req := &DataVisualizationRequest{
		VisualizationType: VisualizationTypeKPI,
		Data:              map[string]interface{}{"test": "data"},
	}

	result, err := handler.generateKPI(req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 125000, result["value"])
	assert.Equal(t, "Total Verifications", result["label"])
	assert.Equal(t, 12.5, result["change"])
	assert.Equal(t, "increase", result["change_type"])
}

func TestDataVisualizationHandler_generateDashboard(t *testing.T) {
	handler := NewDataVisualizationHandler(zap.NewNop())
	req := &struct {
		BusinessID string                 `json:"business_id,omitempty"`
		Widgets    []DashboardWidget      `json:"widgets"`
		Layout     map[string]interface{} `json:"layout,omitempty"`
		Theme      string                 `json:"theme,omitempty"`
		Config     *VisualizationConfig   `json:"config,omitempty"`
		Metadata   map[string]interface{} `json:"metadata,omitempty"`
	}{
		BusinessID: "business_123",
		Theme:      "light",
	}

	result, err := handler.generateDashboard(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "dashboard_id")
	assert.Equal(t, "KYB Platform Dashboard", result["title"])
	assert.Contains(t, result, "layout")
	assert.Equal(t, "light", result["theme"])
}

func TestDataVisualizationHandler_createBackgroundJob(t *testing.T) {
	handler := NewDataVisualizationHandler(zap.NewNop())
	req := &DataVisualizationRequest{
		BusinessID:        "business_123",
		VisualizationType: VisualizationTypeLineChart,
		ChartType:         ChartTypeLine,
		Data:              map[string]interface{}{"test": "data"},
		Metadata:          map[string]interface{}{"key": "value"},
	}

	job := handler.createBackgroundJob(req)

	assert.NotEmpty(t, job.JobID)
	assert.Equal(t, "business_123", job.BusinessID)
	assert.Equal(t, VisualizationTypeLineChart, job.Type)
	assert.Equal(t, "pending", job.Status)
	assert.Equal(t, 0.0, job.Progress)
	assert.Equal(t, 5, job.TotalSteps)
	assert.Equal(t, 0, job.CurrentStep)
	assert.Equal(t, "Initializing visualization job", job.StepDescription)
	assert.Equal(t, req.Metadata, job.Metadata)

	// Verify job is stored
	handler.mutex.RLock()
	storedJob, exists := handler.jobs[job.JobID]
	handler.mutex.RUnlock()
	assert.True(t, exists)
	assert.Equal(t, job, storedJob)
}

func TestDataVisualizationHandler_processVisualizationJob(t *testing.T) {
	handler := NewDataVisualizationHandler(zap.NewNop())
	job := &BackgroundVisualizationJob{
		JobID:      "test_job",
		BusinessID: "business_123",
		Type:       VisualizationTypeLineChart,
		Status:     "pending",
		CreatedAt:  time.Now(),
	}

	// Store job
	handler.jobs[job.JobID] = job

	// Process job
	handler.processVisualizationJob(job)

	// Verify job completion
	handler.mutex.RLock()
	updatedJob := handler.jobs[job.JobID]
	handler.mutex.RUnlock()

	assert.Equal(t, "completed", updatedJob.Status)
	assert.Equal(t, 1.0, updatedJob.Progress)
	assert.Equal(t, 5, updatedJob.CurrentStep)
	assert.Equal(t, "Visualization completed", updatedJob.StepDescription)
	assert.NotNil(t, updatedJob.Result)
	assert.NotNil(t, updatedJob.CompletedAt)
}

func TestDataVisualizationHandler_updateJobStatus(t *testing.T) {
	handler := NewDataVisualizationHandler(zap.NewNop())
	job := &BackgroundVisualizationJob{
		JobID:      "test_job",
		Status:     "pending",
		Progress:   0.0,
		TotalSteps: 5,
		CreatedAt:  time.Now(),
	}

	// Store job
	handler.jobs[job.JobID] = job

	// Update status
	handler.updateJobStatus(job, "processing", 0.5, "Processing data")

	// Verify update
	handler.mutex.RLock()
	updatedJob := handler.jobs[job.JobID]
	handler.mutex.RUnlock()

	assert.Equal(t, "processing", updatedJob.Status)
	assert.Equal(t, 0.5, updatedJob.Progress)
	assert.Equal(t, 2, updatedJob.CurrentStep)
	assert.Equal(t, "Processing data", updatedJob.StepDescription)
	assert.NotNil(t, updatedJob.StartedAt)
}

func TestDataVisualizationHandler_getVisualizationSchema(t *testing.T) {
	handler := NewDataVisualizationHandler(zap.NewNop())

	tests := []struct {
		name     string
		schemaID string
		expected bool
	}{
		{
			name:     "existing schema",
			schemaID: "default_line_chart",
			expected: true,
		},
		{
			name:     "non-existing schema",
			schemaID: "nonexistent",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := handler.getVisualizationSchema(tt.schemaID)

			if tt.expected {
				assert.NotNil(t, schema)
				assert.Equal(t, tt.schemaID, schema.ID)
				assert.Equal(t, "Default Line Chart", schema.Name)
			} else {
				assert.Nil(t, schema)
			}
		})
	}
}

func TestDataVisualizationHandler_getVisualizationSchemas(t *testing.T) {
	handler := NewDataVisualizationHandler(zap.NewNop())

	tests := []struct {
		name              string
		visualizationType string
		chartType         string
		limit             int
		offset            int
		expectedCount     int
	}{
		{
			name:          "all schemas",
			limit:         10,
			offset:        0,
			expectedCount: 3,
		},
		{
			name:              "filter by type",
			visualizationType: "line_chart",
			limit:             10,
			offset:            0,
			expectedCount:     1,
		},
		{
			name:          "filter by chart type",
			chartType:     "line",
			limit:         10,
			offset:        0,
			expectedCount: 1,
		},
		{
			name:          "pagination",
			limit:         1,
			offset:        0,
			expectedCount: 1,
		},
		{
			name:          "pagination with offset",
			limit:         1,
			offset:        1,
			expectedCount: 1,
		},
		{
			name:          "pagination beyond available",
			limit:         10,
			offset:        10,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemas := handler.getVisualizationSchemas(tt.visualizationType, tt.chartType, tt.limit, tt.offset)

			assert.Len(t, schemas, tt.expectedCount)
		})
	}
}

func TestDataVisualizationHandler_writeJSON(t *testing.T) {
	handler := NewDataVisualizationHandler(zap.NewNop())
	w := httptest.NewRecorder()

	data := map[string]interface{}{
		"test":   "value",
		"number": 123,
	}

	handler.writeJSON(w, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "value", response["test"])
	assert.Equal(t, float64(123), response["number"])
}

func TestDataVisualizationHandler_writeError(t *testing.T) {
	handler := NewDataVisualizationHandler(zap.NewNop())
	w := httptest.NewRecorder()

	handler.writeError(w, http.StatusBadRequest, "TEST_ERROR", "Test error message")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, "TEST_ERROR", errorObj["code"])
	assert.Equal(t, "Test error message", errorObj["message"])
	assert.Contains(t, response, "timestamp")
}
