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

func TestNewDataMiningHandler(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataMiningHandler(logger)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
	assert.NotNil(t, handler.jobs)
	assert.Len(t, handler.jobs, 0)
}

func TestDataMiningHandler_MineData(t *testing.T) {
	tests := []struct {
		name           string
		request        DataMiningRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful clustering mining",
			request: DataMiningRequest{
				BusinessID:           "business_123",
				MiningType:           MiningTypeClustering,
				Algorithm:            MiningAlgorithmKMeans,
				Dataset:              "verification_data",
				Features:             []string{"score", "age", "income"},
				IncludeModel:         true,
				IncludeMetrics:       true,
				IncludeVisualization: true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful classification mining",
			request: DataMiningRequest{
				BusinessID:     "business_123",
				MiningType:     MiningTypeClassification,
				Algorithm:      MiningAlgorithmRandomForest,
				Dataset:        "verification_data",
				Features:       []string{"score", "age", "income"},
				Target:         "status",
				IncludeModel:   true,
				IncludeMetrics: true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful association rules mining",
			request: DataMiningRequest{
				BusinessID:   "business_123",
				MiningType:   MiningTypeAssociationRules,
				Algorithm:    MiningAlgorithmApriori,
				Dataset:      "verification_data",
				Features:     []string{"status", "industry", "region"},
				IncludeModel: true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing business_id",
			request: DataMiningRequest{
				MiningType: MiningTypeClustering,
				Algorithm:  MiningAlgorithmKMeans,
				Dataset:    "verification_data",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "business_id is required",
		},
		{
			name: "missing mining_type",
			request: DataMiningRequest{
				BusinessID: "business_123",
				Algorithm:  MiningAlgorithmKMeans,
				Dataset:    "verification_data",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "mining_type is required",
		},
		{
			name: "missing algorithm",
			request: DataMiningRequest{
				BusinessID: "business_123",
				MiningType: MiningTypeClustering,
				Dataset:    "verification_data",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "algorithm is required",
		},
		{
			name: "missing dataset",
			request: DataMiningRequest{
				BusinessID: "business_123",
				MiningType: MiningTypeClustering,
				Algorithm:  MiningAlgorithmKMeans,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "dataset is required",
		},
		{
			name: "invalid mining_type",
			request: DataMiningRequest{
				BusinessID: "business_123",
				MiningType: "invalid_type",
				Algorithm:  MiningAlgorithmKMeans,
				Dataset:    "verification_data",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid mining_type",
		},
		{
			name: "invalid algorithm",
			request: DataMiningRequest{
				BusinessID: "business_123",
				MiningType: MiningTypeClustering,
				Algorithm:  "invalid_algorithm",
				Dataset:    "verification_data",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid algorithm",
		},
		{
			name: "classification without features",
			request: DataMiningRequest{
				BusinessID: "business_123",
				MiningType: MiningTypeClassification,
				Algorithm:  MiningAlgorithmRandomForest,
				Dataset:    "verification_data",
				Target:     "status",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "features are required for classification",
		},
		{
			name: "classification without target",
			request: DataMiningRequest{
				BusinessID: "business_123",
				MiningType: MiningTypeClassification,
				Algorithm:  MiningAlgorithmRandomForest,
				Dataset:    "verification_data",
				Features:   []string{"score", "age"},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "target is required for classification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataMiningHandler(zap.NewNop())

			// Create request body
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			// Create HTTP request
			req := httptest.NewRequest(http.MethodPost, "/v1/mining", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.MineData(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				// Check error message
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				// Parse response
				var response DataMiningResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				// Assert response fields
				assert.NotEmpty(t, response.MiningID)
				assert.Equal(t, tt.request.BusinessID, response.BusinessID)
				assert.Equal(t, tt.request.MiningType, response.Type)
				assert.Equal(t, tt.request.Algorithm, response.Algorithm)
				assert.Equal(t, "success", response.Status)
				assert.True(t, response.IsSuccessful)
				assert.NotNil(t, response.Results)
				assert.NotZero(t, response.GeneratedAt)
				assert.NotEmpty(t, response.ProcessingTime)

				// Assert results based on mining type
				if tt.request.MiningType == MiningTypeClustering {
					assert.NotEmpty(t, response.Results.Clusters)
				} else if tt.request.MiningType == MiningTypeClassification {
					assert.NotEmpty(t, response.Results.Classifications)
				} else if tt.request.MiningType == MiningTypeAssociationRules {
					assert.NotEmpty(t, response.Results.Associations)
				}

				// Assert optional fields based on request
				if tt.request.IncludeModel {
					assert.NotNil(t, response.Model)
				}
				if tt.request.IncludeMetrics {
					assert.NotNil(t, response.Metrics)
				}
				if tt.request.IncludeVisualization {
					assert.NotNil(t, response.Visualization)
				}
			}
		})
	}
}

func TestDataMiningHandler_CreateMiningJob(t *testing.T) {
	tests := []struct {
		name           string
		request        DataMiningRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful job creation",
			request: DataMiningRequest{
				BusinessID: "business_123",
				MiningType: MiningTypeClustering,
				Algorithm:  MiningAlgorithmKMeans,
				Dataset:    "verification_data",
				Features:   []string{"score", "age", "income"},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request",
			request: DataMiningRequest{
				BusinessID: "business_123",
				// Missing required fields
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "mining_type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataMiningHandler(zap.NewNop())

			// Create request body
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			// Create HTTP request
			req := httptest.NewRequest(http.MethodPost, "/v1/mining/jobs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.CreateMiningJob(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				// Check error message
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				// Parse response
				var job MiningJob
				err := json.Unmarshal(w.Body.Bytes(), &job)
				require.NoError(t, err)

				// Assert job fields
				assert.NotEmpty(t, job.JobID)
				assert.Equal(t, tt.request.BusinessID, job.BusinessID)
				assert.Equal(t, tt.request.MiningType, job.Type)
				assert.Equal(t, tt.request.Algorithm, job.Algorithm)
				assert.Equal(t, JobStatusPending, job.Status)
				assert.Equal(t, 0.0, job.Progress)
				assert.Equal(t, 8, job.TotalSteps)
				assert.Equal(t, 0, job.CurrentStep)
				assert.NotZero(t, job.CreatedAt)

				// Wait a bit for background processing to start
				time.Sleep(100 * time.Millisecond)

				// Check that job status has been updated
				handler.mu.RLock()
				updatedJob, exists := handler.jobs[job.JobID]
				handler.mu.RUnlock()

				assert.True(t, exists)
				assert.Equal(t, JobStatusProcessing, updatedJob.Status)
				assert.Greater(t, updatedJob.Progress, 0.0)
			}
		})
	}
}

func TestDataMiningHandler_GetMiningJob(t *testing.T) {
	tests := []struct {
		name           string
		jobID          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing job_id",
			jobID:          "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "job_id is required",
		},
		{
			name:           "job not found",
			jobID:          "nonexistent_job",
			expectedStatus: http.StatusNotFound,
			expectedError:  "mining job not found",
		},
		{
			name:           "existing job",
			jobID:          "test_job",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataMiningHandler(zap.NewNop())

			// Create test job if needed
			if tt.jobID == "test_job" {
				job := &MiningJob{
					JobID:      tt.jobID,
					BusinessID: "business_123",
					Type:       MiningTypeClustering,
					Algorithm:  MiningAlgorithmKMeans,
					Status:     JobStatusCompleted,
					CreatedAt:  time.Now(),
				}
				handler.mu.Lock()
				handler.jobs[tt.jobID] = job
				handler.mu.Unlock()
			}

			// Create HTTP request
			url := "/v1/mining/jobs"
			if tt.jobID != "" {
				url += "?job_id=" + tt.jobID
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.GetMiningJob(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				// Check error message
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				// Parse response
				var job MiningJob
				err := json.Unmarshal(w.Body.Bytes(), &job)
				require.NoError(t, err)

				// Assert job fields
				assert.Equal(t, tt.jobID, job.JobID)
				assert.Equal(t, "business_123", job.BusinessID)
				assert.Equal(t, MiningTypeClustering, job.Type)
				assert.Equal(t, MiningAlgorithmKMeans, job.Algorithm)
			}
		})
	}
}

func TestDataMiningHandler_ListMiningJobs(t *testing.T) {
	handler := NewDataMiningHandler(zap.NewNop())

	// Create test jobs
	testJobs := []*MiningJob{
		{
			JobID:      "job_1",
			BusinessID: "business_123",
			Type:       MiningTypeClustering,
			Algorithm:  MiningAlgorithmKMeans,
			Status:     JobStatusCompleted,
			CreatedAt:  time.Now(),
		},
		{
			JobID:      "job_2",
			BusinessID: "business_456",
			Type:       MiningTypeClassification,
			Algorithm:  MiningAlgorithmRandomForest,
			Status:     JobStatusProcessing,
			CreatedAt:  time.Now(),
		},
		{
			JobID:      "job_3",
			BusinessID: "business_123",
			Type:       MiningTypeAssociationRules,
			Algorithm:  MiningAlgorithmApriori,
			Status:     JobStatusFailed,
			CreatedAt:  time.Now(),
		},
	}

	// Add jobs to handler
	handler.mu.Lock()
	for _, job := range testJobs {
		handler.jobs[job.JobID] = job
	}
	handler.mu.Unlock()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "list all jobs",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "filter by business_id",
			queryParams:    "?business_id=business_123",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "filter by status",
			queryParams:    "?status=completed",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "filter by mining_type",
			queryParams:    "?mining_type=clustering",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "filter by algorithm",
			queryParams:    "?algorithm=random_forest",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "multiple filters",
			queryParams:    "?business_id=business_123&status=completed",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "pagination",
			queryParams:    "?limit=2&offset=1",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create HTTP request
			req := httptest.NewRequest(http.MethodGet, "/v1/mining/jobs"+tt.queryParams, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.ListMiningJobs(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Assert response structure
			assert.Contains(t, response, "jobs")
			assert.Contains(t, response, "total_count")
			assert.Contains(t, response, "limit")
			assert.Contains(t, response, "offset")

			// Assert job count
			jobs := response["jobs"].([]interface{})
			assert.Len(t, jobs, tt.expectedCount)
		})
	}
}

func TestDataMiningHandler_GetMiningSchema(t *testing.T) {
	tests := []struct {
		name           string
		schemaID       string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing schema_id",
			schemaID:       "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "schema_id is required",
		},
		{
			name:           "schema not found",
			schemaID:       "nonexistent_schema",
			expectedStatus: http.StatusNotFound,
			expectedError:  "mining schema not found",
		},
		{
			name:           "existing clustering schema",
			schemaID:       "clustering_schema",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "existing classification schema",
			schemaID:       "classification_schema",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataMiningHandler(zap.NewNop())

			// Create HTTP request
			url := "/v1/mining/schemas"
			if tt.schemaID != "" {
				url += "?schema_id=" + tt.schemaID
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.GetMiningSchema(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				// Check error message
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				// Parse response
				var schema MiningSchema
				err := json.Unmarshal(w.Body.Bytes(), &schema)
				require.NoError(t, err)

				// Assert schema fields
				assert.Equal(t, tt.schemaID, schema.ID)
				assert.NotEmpty(t, schema.Name)
				assert.NotEmpty(t, schema.Description)
				assert.NotEmpty(t, schema.Type)
				assert.NotEmpty(t, schema.Algorithm)
				assert.NotZero(t, schema.CreatedAt)
				assert.NotZero(t, schema.UpdatedAt)
			}
		})
	}
}

func TestDataMiningHandler_ListMiningSchemas(t *testing.T) {
	handler := NewDataMiningHandler(zap.NewNop())

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
			name:           "filter by mining_type",
			queryParams:    "?mining_type=clustering",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "filter by algorithm",
			queryParams:    "?algorithm=random_forest",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "multiple filters",
			queryParams:    "?mining_type=clustering&algorithm=kmeans",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "pagination",
			queryParams:    "?limit=2&offset=1",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create HTTP request
			req := httptest.NewRequest(http.MethodGet, "/v1/mining/schemas"+tt.queryParams, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.ListMiningSchemas(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Assert response structure
			assert.Contains(t, response, "schemas")
			assert.Contains(t, response, "total_count")
			assert.Contains(t, response, "limit")
			assert.Contains(t, response, "offset")

			// Assert schema count
			schemas := response["schemas"].([]interface{})
			assert.Len(t, schemas, tt.expectedCount)
		})
	}
}

func TestDataMiningHandler_validateMiningRequest(t *testing.T) {
	handler := NewDataMiningHandler(zap.NewNop())

	tests := []struct {
		name    string
		request DataMiningRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid clustering request",
			request: DataMiningRequest{
				BusinessID: "business_123",
				MiningType: MiningTypeClustering,
				Algorithm:  MiningAlgorithmKMeans,
				Dataset:    "verification_data",
			},
			wantErr: false,
		},
		{
			name: "valid classification request",
			request: DataMiningRequest{
				BusinessID: "business_123",
				MiningType: MiningTypeClassification,
				Algorithm:  MiningAlgorithmRandomForest,
				Dataset:    "verification_data",
				Features:   []string{"score", "age"},
				Target:     "status",
			},
			wantErr: false,
		},
		{
			name: "missing business_id",
			request: DataMiningRequest{
				MiningType: MiningTypeClustering,
				Algorithm:  MiningAlgorithmKMeans,
				Dataset:    "verification_data",
			},
			wantErr: true,
			errMsg:  "business_id is required",
		},
		{
			name: "missing mining_type",
			request: DataMiningRequest{
				BusinessID: "business_123",
				Algorithm:  MiningAlgorithmKMeans,
				Dataset:    "verification_data",
			},
			wantErr: true,
			errMsg:  "mining_type is required",
		},
		{
			name: "missing algorithm",
			request: DataMiningRequest{
				BusinessID: "business_123",
				MiningType: MiningTypeClustering,
				Dataset:    "verification_data",
			},
			wantErr: true,
			errMsg:  "algorithm is required",
		},
		{
			name: "missing dataset",
			request: DataMiningRequest{
				BusinessID: "business_123",
				MiningType: MiningTypeClustering,
				Algorithm:  MiningAlgorithmKMeans,
			},
			wantErr: true,
			errMsg:  "dataset is required",
		},
		{
			name: "invalid mining_type",
			request: DataMiningRequest{
				BusinessID: "business_123",
				MiningType: "invalid_type",
				Algorithm:  MiningAlgorithmKMeans,
				Dataset:    "verification_data",
			},
			wantErr: true,
			errMsg:  "invalid mining_type",
		},
		{
			name: "invalid algorithm",
			request: DataMiningRequest{
				BusinessID: "business_123",
				MiningType: MiningTypeClustering,
				Algorithm:  "invalid_algorithm",
				Dataset:    "verification_data",
			},
			wantErr: true,
			errMsg:  "invalid algorithm",
		},
		{
			name: "classification without features",
			request: DataMiningRequest{
				BusinessID: "business_123",
				MiningType: MiningTypeClassification,
				Algorithm:  MiningAlgorithmRandomForest,
				Dataset:    "verification_data",
				Target:     "status",
			},
			wantErr: true,
			errMsg:  "features are required for classification",
		},
		{
			name: "classification without target",
			request: DataMiningRequest{
				BusinessID: "business_123",
				MiningType: MiningTypeClassification,
				Algorithm:  MiningAlgorithmRandomForest,
				Dataset:    "verification_data",
				Features:   []string{"score", "age"},
			},
			wantErr: true,
			errMsg:  "target is required for classification",
		},
		{
			name: "regression without features",
			request: DataMiningRequest{
				BusinessID: "business_123",
				MiningType: MiningTypeRegression,
				Algorithm:  MiningAlgorithmLinearRegression,
				Dataset:    "verification_data",
				Target:     "score",
			},
			wantErr: true,
			errMsg:  "features are required for regression",
		},
		{
			name: "regression without target",
			request: DataMiningRequest{
				BusinessID: "business_123",
				MiningType: MiningTypeRegression,
				Algorithm:  MiningAlgorithmLinearRegression,
				Dataset:    "verification_data",
				Features:   []string{"age", "income"},
			},
			wantErr: true,
			errMsg:  "target is required for regression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateMiningRequest(&tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataMiningHandler_performMining(t *testing.T) {
	handler := NewDataMiningHandler(zap.NewNop())

	request := &DataMiningRequest{
		BusinessID: "business_123",
		MiningType: MiningTypeClustering,
		Algorithm:  MiningAlgorithmKMeans,
		Dataset:    "verification_data",
		Features:   []string{"score", "age", "income"},
	}

	results, model, metrics, visualization, insights, recommendations, err := handler.performMining(context.Background(), request)

	// Assert no error
	assert.NoError(t, err)

	// Assert results
	assert.NotNil(t, results)
	assert.NotEmpty(t, results.Patterns)
	assert.NotEmpty(t, results.Clusters)
	assert.NotEmpty(t, results.Associations)
	assert.NotNil(t, results.Summary)

	// Assert model
	assert.NotNil(t, model)
	assert.Equal(t, request.MiningType, model.Type)
	assert.Equal(t, request.Algorithm, model.Algorithm)
	assert.NotEmpty(t, model.ID)
	assert.NotEmpty(t, model.Version)
	assert.NotNil(t, model.Parameters)
	assert.NotNil(t, model.Performance)
	assert.NotZero(t, model.CreatedAt)

	// Assert metrics
	assert.NotNil(t, metrics)
	assert.Greater(t, metrics.ProcessingTime, 0.0)
	assert.Greater(t, metrics.MemoryUsage, 0.0)
	assert.Greater(t, metrics.DataSize, 0)
	assert.Equal(t, len(request.Features), metrics.FeatureCount)

	// Assert visualization
	assert.NotNil(t, visualization)
	assert.NotEmpty(t, visualization.Type)
	assert.NotNil(t, visualization.Data)
	assert.NotEmpty(t, visualization.Format)

	// Assert insights
	assert.NotEmpty(t, insights)
	for _, insight := range insights {
		assert.NotEmpty(t, insight.ID)
		assert.NotEmpty(t, insight.Type)
		assert.NotEmpty(t, insight.Title)
		assert.NotEmpty(t, insight.Description)
		assert.Greater(t, insight.Confidence, 0.0)
		assert.NotEmpty(t, insight.Impact)
		assert.NotEmpty(t, insight.Category)
	}

	// Assert recommendations
	assert.NotEmpty(t, recommendations)
	for _, recommendation := range recommendations {
		assert.NotEmpty(t, recommendation)
	}
}

func TestDataMiningHandler_processMiningJob(t *testing.T) {
	handler := NewDataMiningHandler(zap.NewNop())

	request := &DataMiningRequest{
		BusinessID: "business_123",
		MiningType: MiningTypeClustering,
		Algorithm:  MiningAlgorithmKMeans,
		Dataset:    "verification_data",
		Features:   []string{"score", "age", "income"},
	}

	job := &MiningJob{
		JobID:      "test_job",
		BusinessID: "business_123",
		Type:       MiningTypeClustering,
		Algorithm:  MiningAlgorithmKMeans,
		Status:     JobStatusPending,
		CreatedAt:  time.Now(),
	}

	// Add job to handler
	handler.mu.Lock()
	handler.jobs[job.JobID] = job
	handler.mu.Unlock()

	// Process job
	handler.processMiningJob(context.Background(), job, request)

	// Check job status
	handler.mu.RLock()
	updatedJob := handler.jobs[job.JobID]
	handler.mu.RUnlock()

	// Assert job completion
	assert.Equal(t, JobStatusCompleted, updatedJob.Status)
	assert.Equal(t, 1.0, updatedJob.Progress)
	assert.Equal(t, 8, updatedJob.CurrentStep)
	assert.NotNil(t, updatedJob.Result)
	assert.NotNil(t, updatedJob.StartedAt)
	assert.NotNil(t, updatedJob.CompletedAt)

	// Assert result
	result := updatedJob.Result
	assert.Equal(t, job.JobID, result.MiningID)
	assert.Equal(t, job.BusinessID, result.BusinessID)
	assert.Equal(t, job.Type, result.Type)
	assert.Equal(t, job.Algorithm, result.Algorithm)
	assert.Equal(t, "success", result.Status)
	assert.True(t, result.IsSuccessful)
	assert.NotNil(t, result.Results)
	assert.NotZero(t, result.GeneratedAt)
	assert.NotEmpty(t, result.ProcessingTime)
}

func TestDataMiningHandler_updateJobStatus(t *testing.T) {
	handler := NewDataMiningHandler(zap.NewNop())

	job := &MiningJob{
		JobID:      "test_job",
		BusinessID: "business_123",
		Type:       MiningTypeClustering,
		Algorithm:  MiningAlgorithmKMeans,
		Status:     JobStatusPending,
		CreatedAt:  time.Now(),
	}

	// Add job to handler
	handler.mu.Lock()
	handler.jobs[job.JobID] = job
	handler.mu.Unlock()

	// Update job status
	handler.updateJobStatus(job, JobStatusProcessing, 0.5, 4, "Processing data")

	// Check updated job
	handler.mu.RLock()
	updatedJob := handler.jobs[job.JobID]
	handler.mu.RUnlock()

	// Assert status update
	assert.Equal(t, JobStatusProcessing, updatedJob.Status)
	assert.Equal(t, 0.5, updatedJob.Progress)
	assert.Equal(t, 4, updatedJob.CurrentStep)
	assert.Equal(t, "Processing data", updatedJob.StepDescription)
	assert.NotNil(t, updatedJob.StartedAt)
}

func TestDataMiningHandler_updateJobWithResult(t *testing.T) {
	handler := NewDataMiningHandler(zap.NewNop())

	job := &MiningJob{
		JobID:      "test_job",
		BusinessID: "business_123",
		Type:       MiningTypeClustering,
		Algorithm:  MiningAlgorithmKMeans,
		Status:     JobStatusProcessing,
		CreatedAt:  time.Now(),
	}

	result := &DataMiningResponse{
		MiningID:     "test_mining",
		BusinessID:   "business_123",
		Type:         MiningTypeClustering,
		Algorithm:    MiningAlgorithmKMeans,
		Status:       "success",
		IsSuccessful: true,
		GeneratedAt:  time.Now(),
	}

	// Add job to handler
	handler.mu.Lock()
	handler.jobs[job.JobID] = job
	handler.mu.Unlock()

	// Update job with result
	handler.updateJobWithResult(job, result, JobStatusCompleted, 1.0, 8, "Completed")

	// Check updated job
	handler.mu.RLock()
	updatedJob := handler.jobs[job.JobID]
	handler.mu.RUnlock()

	// Assert result update
	assert.Equal(t, JobStatusCompleted, updatedJob.Status)
	assert.Equal(t, 1.0, updatedJob.Progress)
	assert.Equal(t, 8, updatedJob.CurrentStep)
	assert.Equal(t, "Completed", updatedJob.StepDescription)
	assert.Equal(t, result, updatedJob.Result)
	assert.NotNil(t, updatedJob.CompletedAt)
}

func TestDataMiningHandler_getMiningSchema(t *testing.T) {
	handler := NewDataMiningHandler(zap.NewNop())

	tests := []struct {
		name     string
		schemaID string
		expected *MiningSchema
	}{
		{
			name:     "clustering schema",
			schemaID: "clustering_schema",
			expected: &MiningSchema{
				ID:                   "clustering_schema",
				Name:                 "Customer Segmentation Clustering",
				Type:                 MiningTypeClustering,
				Algorithm:            MiningAlgorithmKMeans,
				IncludeModel:         true,
				IncludeMetrics:       true,
				IncludeVisualization: true,
			},
		},
		{
			name:     "classification schema",
			schemaID: "classification_schema",
			expected: &MiningSchema{
				ID:                   "classification_schema",
				Name:                 "Risk Classification",
				Type:                 MiningTypeClassification,
				Algorithm:            MiningAlgorithmRandomForest,
				IncludeModel:         true,
				IncludeMetrics:       true,
				IncludeVisualization: true,
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
			schema := handler.getMiningSchema(tt.schemaID)

			if tt.expected == nil {
				assert.Nil(t, schema)
			} else {
				assert.NotNil(t, schema)
				assert.Equal(t, tt.expected.ID, schema.ID)
				assert.Equal(t, tt.expected.Name, schema.Name)
				assert.Equal(t, tt.expected.Type, schema.Type)
				assert.Equal(t, tt.expected.Algorithm, schema.Algorithm)
				assert.Equal(t, tt.expected.IncludeModel, schema.IncludeModel)
				assert.Equal(t, tt.expected.IncludeMetrics, schema.IncludeMetrics)
				assert.Equal(t, tt.expected.IncludeVisualization, schema.IncludeVisualization)
			}
		})
	}
}

func TestDataMiningHandler_getMiningSchemas(t *testing.T) {
	handler := NewDataMiningHandler(zap.NewNop())

	tests := []struct {
		name       string
		miningType string
		algorithm  string
		limit      int
		offset     int
		expected   int
	}{
		{
			name:       "all schemas",
			miningType: "",
			algorithm:  "",
			limit:      10,
			offset:     0,
			expected:   3,
		},
		{
			name:       "filter by mining type",
			miningType: "clustering",
			algorithm:  "",
			limit:      10,
			offset:     0,
			expected:   1,
		},
		{
			name:       "filter by algorithm",
			miningType: "",
			algorithm:  "random_forest",
			limit:      10,
			offset:     0,
			expected:   1,
		},
		{
			name:       "multiple filters",
			miningType: "clustering",
			algorithm:  "kmeans",
			limit:      10,
			offset:     0,
			expected:   1,
		},
		{
			name:       "pagination",
			miningType: "",
			algorithm:  "",
			limit:      2,
			offset:     1,
			expected:   2,
		},
		{
			name:       "pagination beyond available",
			miningType: "",
			algorithm:  "",
			limit:      10,
			offset:     5,
			expected:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemas := handler.getMiningSchemas(tt.miningType, tt.algorithm, tt.limit, tt.offset)

			assert.Len(t, schemas, tt.expected)

			// Verify schema properties
			for _, schema := range schemas {
				assert.NotEmpty(t, schema.ID)
				assert.NotEmpty(t, schema.Name)
				assert.NotEmpty(t, schema.Description)
				assert.NotEmpty(t, schema.Type)
				assert.NotEmpty(t, schema.Algorithm)
				assert.NotZero(t, schema.CreatedAt)
				assert.NotZero(t, schema.UpdatedAt)

				// Verify filters if applied
				if tt.miningType != "" {
					assert.Equal(t, MiningType(tt.miningType), schema.Type)
				}
				if tt.algorithm != "" {
					assert.Equal(t, MiningAlgorithm(tt.algorithm), schema.Algorithm)
				}
			}
		})
	}
}

func TestMiningType_String(t *testing.T) {
	tests := []struct {
		name       string
		miningType MiningType
		expected   string
	}{
		{
			name:       "pattern discovery",
			miningType: MiningTypePatternDiscovery,
			expected:   "pattern_discovery",
		},
		{
			name:       "clustering",
			miningType: MiningTypeClustering,
			expected:   "clustering",
		},
		{
			name:       "classification",
			miningType: MiningTypeClassification,
			expected:   "classification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.miningType.String())
		})
	}
}

func TestMiningAlgorithm_String(t *testing.T) {
	tests := []struct {
		name      string
		algorithm MiningAlgorithm
		expected  string
	}{
		{
			name:      "kmeans",
			algorithm: MiningAlgorithmKMeans,
			expected:  "kmeans",
		},
		{
			name:      "random forest",
			algorithm: MiningAlgorithmRandomForest,
			expected:  "random_forest",
		},
		{
			name:      "apriori",
			algorithm: MiningAlgorithmApriori,
			expected:  "apriori",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.algorithm.String())
		})
	}
}
