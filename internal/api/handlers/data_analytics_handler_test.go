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

func TestNewDataAnalyticsHandler(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataAnalyticsHandler(logger)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
	assert.NotNil(t, handler.jobs)
	assert.Len(t, handler.jobs, 0)
}

func TestDataAnalyticsHandler_AnalyzeData(t *testing.T) {
	tests := []struct {
		name           string
		request        DataAnalyticsRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful verification trends analytics",
			request: DataAnalyticsRequest{
				BusinessID:      "business_123",
				AnalyticsType:   AnalyticsTypeVerificationTrends,
				Operations:      []AnalyticsOperation{AnalyticsOperationCount, AnalyticsOperationTrend},
				IncludeInsights: true,
				IncludeTrends:   true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful success rates analytics",
			request: DataAnalyticsRequest{
				BusinessID:         "business_123",
				AnalyticsType:      AnalyticsTypeSuccessRates,
				Operations:         []AnalyticsOperation{AnalyticsOperationAverage, AnalyticsOperationPercentage},
				IncludeInsights:    true,
				IncludePredictions: true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful custom query analytics",
			request: DataAnalyticsRequest{
				BusinessID:    "business_123",
				AnalyticsType: AnalyticsTypeCustomQuery,
				Operations:    []AnalyticsOperation{AnalyticsOperationCount},
				CustomQuery:   "SELECT COUNT(*) FROM verifications WHERE status = 'completed'",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing business_id",
			request: DataAnalyticsRequest{
				AnalyticsType: AnalyticsTypeVerificationTrends,
				Operations:    []AnalyticsOperation{AnalyticsOperationCount},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "business_id is required",
		},
		{
			name: "missing analytics_type",
			request: DataAnalyticsRequest{
				BusinessID: "business_123",
				Operations: []AnalyticsOperation{AnalyticsOperationCount},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "analytics_type is required",
		},
		{
			name: "missing operations",
			request: DataAnalyticsRequest{
				BusinessID:    "business_123",
				AnalyticsType: AnalyticsTypeVerificationTrends,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "at least one operation is required",
		},
		{
			name: "invalid analytics_type",
			request: DataAnalyticsRequest{
				BusinessID:    "business_123",
				AnalyticsType: "invalid_type",
				Operations:    []AnalyticsOperation{AnalyticsOperationCount},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid analytics_type",
		},
		{
			name: "invalid operation",
			request: DataAnalyticsRequest{
				BusinessID:    "business_123",
				AnalyticsType: AnalyticsTypeVerificationTrends,
				Operations:    []AnalyticsOperation{"invalid_operation"},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid operation",
		},
		{
			name: "custom query without query",
			request: DataAnalyticsRequest{
				BusinessID:    "business_123",
				AnalyticsType: AnalyticsTypeCustomQuery,
				Operations:    []AnalyticsOperation{AnalyticsOperationCount},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "custom_query is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataAnalyticsHandler(zap.NewNop())

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/v1/analytics", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.AnalyzeData(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response DataAnalyticsResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.NotEmpty(t, response.AnalyticsID)
				assert.Equal(t, tt.request.BusinessID, response.BusinessID)
				assert.Equal(t, string(tt.request.AnalyticsType), string(response.Type))
				assert.Equal(t, "success", response.Status)
				assert.True(t, response.IsSuccessful)
				assert.NotEmpty(t, response.Results)
				assert.NotZero(t, response.GeneratedAt)
				assert.NotEmpty(t, response.ProcessingTime)
			}
		})
	}
}

func TestDataAnalyticsHandler_CreateAnalyticsJob(t *testing.T) {
	tests := []struct {
		name           string
		request        DataAnalyticsRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful job creation",
			request: DataAnalyticsRequest{
				BusinessID:      "business_123",
				AnalyticsType:   AnalyticsTypeVerificationTrends,
				Operations:      []AnalyticsOperation{AnalyticsOperationCount, AnalyticsOperationTrend},
				IncludeInsights: true,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "validation error",
			request: DataAnalyticsRequest{
				BusinessID:    "business_123",
				AnalyticsType: "invalid_type",
				Operations:    []AnalyticsOperation{AnalyticsOperationCount},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid analytics_type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataAnalyticsHandler(zap.NewNop())

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/v1/analytics/jobs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.CreateAnalyticsJob(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var job AnalyticsJob
				err := json.Unmarshal(w.Body.Bytes(), &job)
				require.NoError(t, err)

				assert.NotEmpty(t, job.JobID)
				assert.Equal(t, tt.request.BusinessID, job.BusinessID)
				assert.Equal(t, string(tt.request.AnalyticsType), string(job.Type))
				assert.Equal(t, JobStatusPending, job.Status)
				assert.Equal(t, 0.0, job.Progress)
				assert.Equal(t, 6, job.TotalSteps)
				assert.Equal(t, 0, job.CurrentStep)
				assert.NotZero(t, job.CreatedAt)

				// Verify job is stored
				handler.mu.RLock()
				storedJob, exists := handler.jobs[job.JobID]
				handler.mu.RUnlock()

				assert.True(t, exists)
				assert.Equal(t, job.JobID, storedJob.JobID)
			}
		})
	}
}

func TestDataAnalyticsHandler_GetAnalyticsJob(t *testing.T) {
	tests := []struct {
		name           string
		jobID          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "job found",
			jobID:          "test_job_123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "job not found",
			jobID:          "nonexistent_job",
			expectedStatus: http.StatusNotFound,
			expectedError:  "analytics job not found",
		},
		{
			name:           "missing job_id",
			jobID:          "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "job_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataAnalyticsHandler(zap.NewNop())

			// Create a test job if needed
			if tt.jobID == "test_job_123" {
				handler.mu.Lock()
				handler.jobs[tt.jobID] = &AnalyticsJob{
					JobID:      tt.jobID,
					BusinessID: "business_123",
					Type:       AnalyticsTypeVerificationTrends,
					Status:     JobStatusCompleted,
					CreatedAt:  time.Now(),
				}
				handler.mu.Unlock()
			}

			url := "/v1/analytics/jobs"
			if tt.jobID != "" {
				url += "?job_id=" + tt.jobID
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			handler.GetAnalyticsJob(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var job AnalyticsJob
				err := json.Unmarshal(w.Body.Bytes(), &job)
				require.NoError(t, err)

				assert.Equal(t, tt.jobID, job.JobID)
				assert.Equal(t, "business_123", job.BusinessID)
				assert.Equal(t, AnalyticsTypeVerificationTrends, job.Type)
			}
		})
	}
}

func TestDataAnalyticsHandler_ListAnalyticsJobs(t *testing.T) {
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
			expectedCount:  2,
		},
		{
			name:           "filter by status",
			queryParams:    map[string]string{"status": "completed"},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "filter by business_id",
			queryParams:    map[string]string{"business_id": "business_123"},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "filter by analytics_type",
			queryParams:    map[string]string{"analytics_type": "verification_trends"},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "with pagination",
			queryParams:    map[string]string{"limit": "1", "offset": "0"},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataAnalyticsHandler(zap.NewNop())

			// Create test jobs
			handler.mu.Lock()
			handler.jobs["job_1"] = &AnalyticsJob{
				JobID:      "job_1",
				BusinessID: "business_123",
				Type:       AnalyticsTypeVerificationTrends,
				Status:     JobStatusCompleted,
				CreatedAt:  time.Now(),
			}
			handler.jobs["job_2"] = &AnalyticsJob{
				JobID:      "job_2",
				BusinessID: "business_123",
				Type:       AnalyticsTypeSuccessRates,
				Status:     JobStatusPending,
				CreatedAt:  time.Now(),
			}
			handler.mu.Unlock()

			// Build query string
			query := ""
			for key, value := range tt.queryParams {
				if query != "" {
					query += "&"
				}
				query += key + "=" + value
			}

			url := "/v1/analytics/jobs"
			if query != "" {
				url += "?" + query
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			handler.ListAnalyticsJobs(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.NotNil(t, response["jobs"])
			assert.Equal(t, float64(tt.expectedCount), float64(len(response["jobs"].([]interface{}))))
			assert.Equal(t, float64(2), response["total_count"])
		})
	}
}

func TestDataAnalyticsHandler_GetAnalyticsSchema(t *testing.T) {
	tests := []struct {
		name           string
		schemaID       string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "schema found",
			schemaID:       "verification_trends_schema",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "schema not found",
			schemaID:       "nonexistent_schema",
			expectedStatus: http.StatusNotFound,
			expectedError:  "analytics schema not found",
		},
		{
			name:           "missing schema_id",
			schemaID:       "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "schema_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataAnalyticsHandler(zap.NewNop())

			url := "/v1/analytics/schemas"
			if tt.schemaID != "" {
				url += "?schema_id=" + tt.schemaID
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			handler.GetAnalyticsSchema(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var schema AnalyticsSchema
				err := json.Unmarshal(w.Body.Bytes(), &schema)
				require.NoError(t, err)

				assert.Equal(t, tt.schemaID, schema.ID)
				assert.NotEmpty(t, schema.Name)
				assert.NotEmpty(t, schema.Description)
				assert.NotEmpty(t, schema.Operations)
			}
		})
	}
}

func TestDataAnalyticsHandler_ListAnalyticsSchemas(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "list all schemas",
			queryParams:    map[string]string{},
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "filter by analytics_type",
			queryParams:    map[string]string{"analytics_type": "verification_trends"},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "with pagination",
			queryParams:    map[string]string{"limit": "2", "offset": "0"},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataAnalyticsHandler(zap.NewNop())

			// Build query string
			query := ""
			for key, value := range tt.queryParams {
				if query != "" {
					query += "&"
				}
				query += key + "=" + value
			}

			url := "/v1/analytics/schemas"
			if query != "" {
				url += "?" + query
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			handler.ListAnalyticsSchemas(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.NotNil(t, response["schemas"])
			assert.Equal(t, float64(tt.expectedCount), float64(len(response["schemas"].([]interface{}))))
		})
	}
}

func TestDataAnalyticsHandler_validateAnalyticsRequest(t *testing.T) {
	tests := []struct {
		name    string
		request DataAnalyticsRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: DataAnalyticsRequest{
				BusinessID:    "business_123",
				AnalyticsType: AnalyticsTypeVerificationTrends,
				Operations:    []AnalyticsOperation{AnalyticsOperationCount},
			},
			wantErr: false,
		},
		{
			name: "missing business_id",
			request: DataAnalyticsRequest{
				AnalyticsType: AnalyticsTypeVerificationTrends,
				Operations:    []AnalyticsOperation{AnalyticsOperationCount},
			},
			wantErr: true,
			errMsg:  "business_id is required",
		},
		{
			name: "missing analytics_type",
			request: DataAnalyticsRequest{
				BusinessID: "business_123",
				Operations: []AnalyticsOperation{AnalyticsOperationCount},
			},
			wantErr: true,
			errMsg:  "analytics_type is required",
		},
		{
			name: "missing operations",
			request: DataAnalyticsRequest{
				BusinessID:    "business_123",
				AnalyticsType: AnalyticsTypeVerificationTrends,
			},
			wantErr: true,
			errMsg:  "at least one operation is required",
		},
		{
			name: "invalid analytics_type",
			request: DataAnalyticsRequest{
				BusinessID:    "business_123",
				AnalyticsType: "invalid_type",
				Operations:    []AnalyticsOperation{AnalyticsOperationCount},
			},
			wantErr: true,
			errMsg:  "invalid analytics_type",
		},
		{
			name: "invalid operation",
			request: DataAnalyticsRequest{
				BusinessID:    "business_123",
				AnalyticsType: AnalyticsTypeVerificationTrends,
				Operations:    []AnalyticsOperation{"invalid_operation"},
			},
			wantErr: true,
			errMsg:  "invalid operation",
		},
		{
			name: "custom query without query",
			request: DataAnalyticsRequest{
				BusinessID:    "business_123",
				AnalyticsType: AnalyticsTypeCustomQuery,
				Operations:    []AnalyticsOperation{AnalyticsOperationCount},
			},
			wantErr: true,
			errMsg:  "custom_query is required",
		},
	}

	handler := NewDataAnalyticsHandler(zap.NewNop())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateAnalyticsRequest(&tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataAnalyticsHandler_performAnalytics(t *testing.T) {
	handler := NewDataAnalyticsHandler(zap.NewNop())

	req := &DataAnalyticsRequest{
		BusinessID:      "business_123",
		AnalyticsType:   AnalyticsTypeVerificationTrends,
		Operations:      []AnalyticsOperation{AnalyticsOperationCount, AnalyticsOperationTrend},
		IncludeInsights: true,
		IncludeTrends:   true,
	}

	results, insights, predictions, trends, correlations, summary, err := handler.performAnalytics(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, results)
	assert.NotEmpty(t, insights)
	assert.NotEmpty(t, predictions)
	assert.NotEmpty(t, trends)
	assert.NotEmpty(t, correlations)
	assert.NotNil(t, summary)

	// Verify results structure
	assert.Len(t, results, 2)
	assert.Equal(t, AnalyticsOperationCount, results[0].Operation)
	assert.Equal(t, "verifications", results[0].Field)
	assert.Equal(t, float64(1500), results[0].Value)

	// Verify insights structure
	assert.Len(t, insights, 1)
	assert.Equal(t, "trend", insights[0].Type)
	assert.NotEmpty(t, insights[0].Title)
	assert.NotEmpty(t, insights[0].Description)

	// Verify predictions structure
	assert.Len(t, predictions, 1)
	assert.Equal(t, "monthly_verifications", predictions[0].Field)
	assert.Equal(t, float64(1800), predictions[0].PredictedValue)
	assert.Equal(t, 0.92, predictions[0].Confidence)

	// Verify trends structure
	assert.Len(t, trends, 1)
	assert.Equal(t, "verification_volume", trends[0].Field)
	assert.Equal(t, "increasing", trends[0].Direction)
	assert.Equal(t, 0.15, trends[0].Slope)

	// Verify correlations structure
	assert.Len(t, correlations, 1)
	assert.Equal(t, "verification_volume", correlations[0].Field1)
	assert.Equal(t, "success_rate", correlations[0].Field2)
	assert.Equal(t, 0.65, correlations[0].Coefficient)

	// Verify summary structure
	assert.Equal(t, 1500, summary.TotalRecords)
	assert.NotNil(t, summary.TimeRange)
	assert.NotEmpty(t, summary.KeyMetrics)
	assert.NotEmpty(t, summary.TopInsights)
	assert.NotEmpty(t, summary.Recommendations)
}

func TestDataAnalyticsHandler_processAnalyticsJob(t *testing.T) {
	handler := NewDataAnalyticsHandler(zap.NewNop())

	req := &DataAnalyticsRequest{
		BusinessID:    "business_123",
		AnalyticsType: AnalyticsTypeVerificationTrends,
		Operations:    []AnalyticsOperation{AnalyticsOperationCount},
	}

	job := &AnalyticsJob{
		JobID:      "test_job_123",
		BusinessID: "business_123",
		Type:       AnalyticsTypeVerificationTrends,
		Status:     JobStatusPending,
		CreatedAt:  time.Now(),
	}

	// Store job
	handler.mu.Lock()
	handler.jobs[job.JobID] = job
	handler.mu.Unlock()

	// Process job
	handler.processAnalyticsJob(context.Background(), job, req)

	// Verify job completion
	handler.mu.RLock()
	updatedJob := handler.jobs[job.JobID]
	handler.mu.RUnlock()

	assert.Equal(t, JobStatusCompleted, updatedJob.Status)
	assert.Equal(t, 1.0, updatedJob.Progress)
	assert.Equal(t, 6, updatedJob.CurrentStep)
	assert.NotNil(t, updatedJob.Result)
	assert.NotNil(t, updatedJob.CompletedAt)

	// Verify result
	result := updatedJob.Result
	assert.Equal(t, job.JobID, result.AnalyticsID)
	assert.Equal(t, "business_123", result.BusinessID)
	assert.Equal(t, AnalyticsTypeVerificationTrends, result.Type)
	assert.Equal(t, "success", result.Status)
	assert.True(t, result.IsSuccessful)
	assert.NotEmpty(t, result.Results)
	assert.NotZero(t, result.GeneratedAt)
	assert.NotEmpty(t, result.ProcessingTime)
}

func TestDataAnalyticsHandler_getAnalyticsSchema(t *testing.T) {
	handler := NewDataAnalyticsHandler(zap.NewNop())

	tests := []struct {
		name     string
		schemaID string
		expected *AnalyticsSchema
	}{
		{
			name:     "verification trends schema",
			schemaID: "verification_trends_schema",
			expected: &AnalyticsSchema{
				ID:              "verification_trends_schema",
				Name:            "Verification Trends Analysis",
				Type:            AnalyticsTypeVerificationTrends,
				Operations:      []AnalyticsOperation{AnalyticsOperationCount, AnalyticsOperationTrend},
				IncludeInsights: true,
				IncludeTrends:   true,
			},
		},
		{
			name:     "success rates schema",
			schemaID: "success_rates_schema",
			expected: &AnalyticsSchema{
				ID:                 "success_rates_schema",
				Name:               "Success Rate Analysis",
				Type:               AnalyticsTypeSuccessRates,
				Operations:         []AnalyticsOperation{AnalyticsOperationAverage, AnalyticsOperationPercentage},
				IncludeInsights:    true,
				IncludePredictions: true,
			},
		},
		{
			name:     "nonexistent schema",
			schemaID: "nonexistent_schema",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := handler.getAnalyticsSchema(tt.schemaID)

			if tt.expected == nil {
				assert.Nil(t, schema)
			} else {
				assert.NotNil(t, schema)
				assert.Equal(t, tt.expected.ID, schema.ID)
				assert.Equal(t, tt.expected.Name, schema.Name)
				assert.Equal(t, tt.expected.Type, schema.Type)
				assert.Equal(t, tt.expected.Operations, schema.Operations)
				assert.Equal(t, tt.expected.IncludeInsights, schema.IncludeInsights)
				assert.Equal(t, tt.expected.IncludeTrends, schema.IncludeTrends)
				assert.Equal(t, tt.expected.IncludePredictions, schema.IncludePredictions)
			}
		})
	}
}

func TestDataAnalyticsHandler_getAnalyticsSchemas(t *testing.T) {
	handler := NewDataAnalyticsHandler(zap.NewNop())

	tests := []struct {
		name          string
		analyticsType string
		limit         int
		offset        int
		expectedCount int
	}{
		{
			name:          "all schemas",
			analyticsType: "",
			limit:         10,
			offset:        0,
			expectedCount: 3,
		},
		{
			name:          "filtered by type",
			analyticsType: "verification_trends",
			limit:         10,
			offset:        0,
			expectedCount: 1,
		},
		{
			name:          "with pagination",
			analyticsType: "",
			limit:         2,
			offset:        0,
			expectedCount: 2,
		},
		{
			name:          "pagination offset",
			analyticsType: "",
			limit:         2,
			offset:        1,
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemas := handler.getAnalyticsSchemas(tt.analyticsType, tt.limit, tt.offset)

			assert.Len(t, schemas, tt.expectedCount)

			// Verify schema structure
			for _, schema := range schemas {
				assert.NotEmpty(t, schema.ID)
				assert.NotEmpty(t, schema.Name)
				assert.NotEmpty(t, schema.Description)
				assert.NotEmpty(t, schema.Type)
				assert.NotEmpty(t, schema.Operations)
				assert.NotZero(t, schema.CreatedAt)
				assert.NotZero(t, schema.UpdatedAt)
			}
		})
	}
}

func TestAnalyticsType_String(t *testing.T) {
	tests := []struct {
		analyticsType AnalyticsType
		expected      string
	}{
		{AnalyticsTypeVerificationTrends, "verification_trends"},
		{AnalyticsTypeSuccessRates, "success_rates"},
		{AnalyticsTypeRiskDistribution, "risk_distribution"},
		{AnalyticsTypeIndustryAnalysis, "industry_analysis"},
		{AnalyticsTypeGeographicAnalysis, "geographic_analysis"},
		{AnalyticsTypePerformanceMetrics, "performance_metrics"},
		{AnalyticsTypeComplianceMetrics, "compliance_metrics"},
		{AnalyticsTypeCustomQuery, "custom_query"},
		{AnalyticsTypePredictiveAnalysis, "predictive_analysis"},
	}

	for _, tt := range tests {
		t.Run(string(tt.analyticsType), func(t *testing.T) {
			result := tt.analyticsType.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalyticsOperation_String(t *testing.T) {
	tests := []struct {
		operation AnalyticsOperation
		expected  string
	}{
		{AnalyticsOperationCount, "count"},
		{AnalyticsOperationSum, "sum"},
		{AnalyticsOperationAverage, "average"},
		{AnalyticsOperationMedian, "median"},
		{AnalyticsOperationMin, "min"},
		{AnalyticsOperationMax, "max"},
		{AnalyticsOperationPercentage, "percentage"},
		{AnalyticsOperationTrend, "trend"},
		{AnalyticsOperationCorrelation, "correlation"},
		{AnalyticsOperationPrediction, "prediction"},
		{AnalyticsOperationAnomalyDetection, "anomaly_detection"},
	}

	for _, tt := range tests {
		t.Run(string(tt.operation), func(t *testing.T) {
			result := tt.operation.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}
