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

func TestNewDataLineageHandler(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataLineageHandler(logger)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
	assert.NotNil(t, handler.lineages)
	assert.NotNil(t, handler.jobs)
	assert.NotNil(t, handler.reports)
	assert.Len(t, handler.lineages, 0)
	assert.Len(t, handler.jobs, 0)
	assert.Len(t, handler.reports, 0)
}

func TestDataLineageHandler_CreateLineage(t *testing.T) {
	tests := []struct {
		name           string
		request        DataLineageRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful lineage creation",
			request: DataLineageRequest{
				Name:        "Test Lineage",
				Description: "Test data lineage",
				Dataset:     "customer_data",
				Type:        LineageTypeDataFlow,
				Direction:   LineageDirectionDownstream,
				Depth:       3,
				Sources: []LineageSource{
					{
						ID:       "source_1",
						Name:     "Customer Database",
						Type:     "database",
						Location: "postgres://localhost:5432/customers",
						Format:   "postgresql",
						Schema: map[string]interface{}{
							"table": "customers",
						},
						Connection: LineageConnection{
							ID:       "conn_1",
							Name:     "Customer DB Connection",
							Type:     "postgresql",
							Protocol: "postgresql",
							Host:     "localhost",
							Port:     5432,
							Database: "customers",
							Schema:   "public",
							Table:    "customers",
						},
						Properties: map[string]interface{}{
							"refresh_rate": "daily",
						},
						Metadata: map[string]interface{}{
							"owner": "data_team",
						},
					},
				},
				Targets: []LineageTarget{
					{
						ID:       "target_1",
						Name:     "Analytics Warehouse",
						Type:     "warehouse",
						Location: "bigquery://project/dataset",
						Format:   "bigquery",
						Schema: map[string]interface{}{
							"dataset": "analytics",
							"table":   "customer_analytics",
						},
						Connection: LineageConnection{
							ID:       "conn_2",
							Name:     "BigQuery Connection",
							Type:     "bigquery",
							Protocol: "https",
							Host:     "bigquery.googleapis.com",
							Port:     443,
							Database: "project",
							Schema:   "dataset",
							Table:    "customer_analytics",
						},
						Properties: map[string]interface{}{
							"refresh_rate": "hourly",
						},
						Metadata: map[string]interface{}{
							"owner": "analytics_team",
						},
					},
				},
				Processes: []LineageProcess{
					{
						ID:          "process_1",
						Name:        "Data Transformation",
						Type:        "etl",
						Description: "Transform customer data for analytics",
						Inputs:      []string{"source_1"},
						Outputs:     []string{"target_1"},
						Logic:       "SELECT * FROM customers WHERE active = true",
						Parameters: map[string]interface{}{
							"filter": "active_customers",
						},
						Schedule: "0 2 * * *",
						Status:   "active",
						Metadata: map[string]interface{}{
							"owner": "etl_team",
						},
					},
				},
				Transformations: []LineageTransformation{
					{
						ID:           "transform_1",
						Name:         "Customer Filter",
						Type:         "filter",
						Description:  "Filter active customers only",
						InputFields:  []string{"customer_id", "name", "email", "active"},
						OutputFields: []string{"customer_id", "name", "email"},
						Logic:        "active = true",
						Rules: []LineageTransformationRule{
							{
								ID:          "rule_1",
								Name:        "Active Customer Rule",
								Type:        "filter",
								Description: "Only include active customers",
								Expression:  "active = true",
								Parameters:  map[string]interface{}{},
								Priority:    1,
								Enabled:     true,
							},
						},
						Conditions: []TransformationCondition{
							{
								ID:          "condition_1",
								Name:        "Active Status Check",
								Type:        "validation",
								Description: "Check if customer is active",
								Expression:  "active",
								Parameters:  map[string]interface{}{},
								Operator:    "equals",
								Value:       true,
							},
						},
						Metadata: map[string]interface{}{
							"owner": "data_team",
						},
					},
				},
				Filters: LineageFilters{
					Types:    []string{"data_flow", "transformation"},
					Statuses: []string{"active"},
					DateRange: DateRange{
						Start: time.Now().AddDate(0, 0, -30),
						End:   time.Now(),
					},
					Tags:   []string{"customer", "analytics"},
					Owners: []string{"data_team"},
					Custom: map[string]interface{}{
						"priority": "high",
					},
				},
				Options: LineageOptions{
					IncludeMetadata: true,
					IncludeSchema:   true,
					IncludeStats:    true,
					MaxDepth:        5,
					MaxNodes:        100,
					Format:          "json",
					Direction:       LineageDirectionDownstream,
					Custom: map[string]interface{}{
						"visualization": "graph",
					},
				},
				Metadata: map[string]interface{}{
					"priority": "high",
					"tags":     []string{"customer", "analytics"},
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			request: DataLineageRequest{
				Dataset:   "customer_data",
				Type:      LineageTypeDataFlow,
				Direction: LineageDirectionDownstream,
				Depth:     3,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name is required",
		},
		{
			name: "missing dataset",
			request: DataLineageRequest{
				Name:      "Test Lineage",
				Type:      LineageTypeDataFlow,
				Direction: LineageDirectionDownstream,
				Depth:     3,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "dataset is required",
		},
		{
			name: "missing type",
			request: DataLineageRequest{
				Name:      "Test Lineage",
				Dataset:   "customer_data",
				Direction: LineageDirectionDownstream,
				Depth:     3,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "type is required",
		},
		{
			name: "missing direction",
			request: DataLineageRequest{
				Name:    "Test Lineage",
				Dataset: "customer_data",
				Type:    LineageTypeDataFlow,
				Depth:   3,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "direction is required",
		},
		{
			name: "invalid depth",
			request: DataLineageRequest{
				Name:      "Test Lineage",
				Dataset:   "customer_data",
				Type:      LineageTypeDataFlow,
				Direction: LineageDirectionDownstream,
				Depth:     0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "depth must be greater than 0",
		},
	}

	handler := NewDataLineageHandler(zap.NewNop())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/lineage", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.CreateLineage(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response DataLineageResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.NotEmpty(t, response.ID)
				assert.Equal(t, tt.request.Name, response.Name)
				assert.Equal(t, tt.request.Type, response.Type)
				assert.Equal(t, LineageStatusActive, response.Status)
				assert.Equal(t, tt.request.Dataset, response.Dataset)
				assert.Equal(t, tt.request.Direction, response.Direction)
				assert.Equal(t, tt.request.Depth, response.Depth)
				assert.NotNil(t, response.Nodes)
				assert.NotNil(t, response.Edges)
				assert.NotNil(t, response.Paths)
				assert.NotNil(t, response.Impact)
				assert.NotNil(t, response.Summary)
				assert.NotZero(t, response.CreatedAt)
				assert.NotZero(t, response.UpdatedAt)
			}
		})
	}
}

func TestDataLineageHandler_GetLineage(t *testing.T) {
	handler := NewDataLineageHandler(zap.NewNop())

	// Create a test lineage
	lineage := &DataLineageResponse{
		ID:        "test_lineage_123",
		Name:      "Test Lineage",
		Type:      LineageTypeDataFlow,
		Status:    LineageStatusActive,
		Dataset:   "test_dataset",
		Direction: LineageDirectionDownstream,
		Depth:     3,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	handler.lineages["test_lineage_123"] = lineage

	tests := []struct {
		name           string
		id             string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful retrieval",
			id:             "test_lineage_123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing id",
			id:             "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Lineage ID is required",
		},
		{
			name:           "not found",
			id:             "nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Lineage not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/lineage"
			if tt.id != "" {
				url = fmt.Sprintf("/lineage?id=%s", tt.id)
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.GetLineage(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response DataLineageResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, lineage.ID, response.ID)
				assert.Equal(t, lineage.Name, response.Name)
				assert.Equal(t, lineage.Type, response.Type)
				assert.Equal(t, lineage.Status, response.Status)
			}
		})
	}
}

func TestDataLineageHandler_ListLineages(t *testing.T) {
	handler := NewDataLineageHandler(zap.NewNop())

	// Create test lineages
	lineage1 := &DataLineageResponse{
		ID:        "lineage_1",
		Name:      "Lineage 1",
		Type:      LineageTypeDataFlow,
		Status:    LineageStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	lineage2 := &DataLineageResponse{
		ID:        "lineage_2",
		Name:      "Lineage 2",
		Type:      LineageTypeTransformation,
		Status:    LineageStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	handler.lineages["lineage_1"] = lineage1
	handler.lineages["lineage_2"] = lineage2

	req := httptest.NewRequest("GET", "/lineage", nil)
	w := httptest.NewRecorder()

	handler.ListLineages(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(2), response["total"])

	lineages := response["lineages"].([]interface{})
	assert.Len(t, lineages, 2)
}

func TestDataLineageHandler_CreateLineageJob(t *testing.T) {
	tests := []struct {
		name           string
		request        DataLineageRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful job creation",
			request: DataLineageRequest{
				Name:        "Test Lineage Job",
				Description: "Test background lineage job",
				Dataset:     "customer_data",
				Type:        LineageTypeDataFlow,
				Direction:   LineageDirectionDownstream,
				Depth:       3,
				Sources: []LineageSource{
					{
						ID:       "source_1",
						Name:     "Customer Database",
						Type:     "database",
						Location: "postgres://localhost:5432/customers",
						Format:   "postgresql",
						Schema:   map[string]interface{}{},
						Connection: LineageConnection{
							ID:       "conn_1",
							Name:     "Customer DB Connection",
							Type:     "postgresql",
							Protocol: "postgresql",
							Host:     "localhost",
							Port:     5432,
							Database: "customers",
						},
						Properties: map[string]interface{}{},
						Metadata:   map[string]interface{}{},
					},
				},
				Targets: []LineageTarget{
					{
						ID:       "target_1",
						Name:     "Analytics Warehouse",
						Type:     "warehouse",
						Location: "bigquery://project/dataset",
						Format:   "bigquery",
						Schema:   map[string]interface{}{},
						Connection: LineageConnection{
							ID:       "conn_2",
							Name:     "BigQuery Connection",
							Type:     "bigquery",
							Protocol: "https",
							Host:     "bigquery.googleapis.com",
							Port:     443,
							Database: "project",
						},
						Properties: map[string]interface{}{},
						Metadata:   map[string]interface{}{},
					},
				},
				Processes: []LineageProcess{
					{
						ID:          "process_1",
						Name:        "Data Transformation",
						Type:        "etl",
						Description: "Transform customer data for analytics",
						Inputs:      []string{"source_1"},
						Outputs:     []string{"target_1"},
						Logic:       "SELECT * FROM customers WHERE active = true",
						Parameters:  map[string]interface{}{},
						Schedule:    "0 2 * * *",
						Status:      "active",
						Metadata:    map[string]interface{}{},
					},
				},
				Transformations: []LineageTransformation{},
				Filters: LineageFilters{
					Types:    []string{"data_flow"},
					Statuses: []string{"active"},
					DateRange: DateRange{
						Start: time.Now().AddDate(0, 0, -30),
						End:   time.Now(),
					},
					Tags:   []string{"customer"},
					Owners: []string{"data_team"},
					Custom: map[string]interface{}{},
				},
				Options: LineageOptions{
					IncludeMetadata: true,
					IncludeSchema:   true,
					IncludeStats:    true,
					MaxDepth:        5,
					MaxNodes:        100,
					Format:          "json",
					Direction:       LineageDirectionDownstream,
					Custom:          map[string]interface{}{},
				},
				Metadata: map[string]interface{}{
					"priority": "high",
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request",
			request: DataLineageRequest{
				Name:      "",
				Dataset:   "customer_data",
				Type:      LineageTypeDataFlow,
				Direction: LineageDirectionDownstream,
				Depth:     3,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name is required",
		},
	}

	handler := NewDataLineageHandler(zap.NewNop())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/lineage/jobs", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.CreateLineageJob(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var job LineageJob
				err := json.Unmarshal(w.Body.Bytes(), &job)
				require.NoError(t, err)

				assert.NotEmpty(t, job.ID)
				assert.NotEmpty(t, job.RequestID)
				assert.Equal(t, "pending", job.Status)
				assert.Equal(t, 0, job.Progress)
				assert.NotZero(t, job.CreatedAt)
				assert.NotZero(t, job.UpdatedAt)
				assert.Nil(t, job.CompletedAt)
			}
		})
	}
}

func TestDataLineageHandler_GetLineageJob(t *testing.T) {
	handler := NewDataLineageHandler(zap.NewNop())

	// Create a test job
	job := &LineageJob{
		ID:        "test_job_123",
		RequestID: "req_123",
		Status:    "completed",
		Progress:  100,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	handler.jobs["test_job_123"] = job

	tests := []struct {
		name           string
		id             string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful retrieval",
			id:             "test_job_123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing id",
			id:             "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Job ID is required",
		},
		{
			name:           "not found",
			id:             "nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Job not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/lineage/jobs"
			if tt.id != "" {
				url = fmt.Sprintf("/lineage/jobs?id=%s", tt.id)
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.GetLineageJob(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response LineageJob
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, job.ID, response.ID)
				assert.Equal(t, job.RequestID, response.RequestID)
				assert.Equal(t, job.Status, response.Status)
				assert.Equal(t, job.Progress, response.Progress)
			}
		})
	}
}

func TestDataLineageHandler_ListLineageJobs(t *testing.T) {
	handler := NewDataLineageHandler(zap.NewNop())

	// Create test jobs
	job1 := &LineageJob{
		ID:        "job_1",
		RequestID: "req_1",
		Status:    "completed",
		Progress:  100,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	job2 := &LineageJob{
		ID:        "job_2",
		RequestID: "req_2",
		Status:    "running",
		Progress:  50,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	handler.jobs["job_1"] = job1
	handler.jobs["job_2"] = job2

	req := httptest.NewRequest("GET", "/lineage/jobs", nil)
	w := httptest.NewRecorder()

	handler.ListLineageJobs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(2), response["total"])

	jobs := response["jobs"].([]interface{})
	assert.Len(t, jobs, 2)
}

func TestDataLineageHandler_validateLineageRequest(t *testing.T) {
	handler := NewDataLineageHandler(zap.NewNop())

	tests := []struct {
		name    string
		request DataLineageRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: DataLineageRequest{
				Name:      "Test Lineage",
				Dataset:   "customer_data",
				Type:      LineageTypeDataFlow,
				Direction: LineageDirectionDownstream,
				Depth:     3,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			request: DataLineageRequest{
				Dataset:   "customer_data",
				Type:      LineageTypeDataFlow,
				Direction: LineageDirectionDownstream,
				Depth:     3,
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing dataset",
			request: DataLineageRequest{
				Name:      "Test Lineage",
				Type:      LineageTypeDataFlow,
				Direction: LineageDirectionDownstream,
				Depth:     3,
			},
			wantErr: true,
			errMsg:  "dataset is required",
		},
		{
			name: "missing type",
			request: DataLineageRequest{
				Name:      "Test Lineage",
				Dataset:   "customer_data",
				Direction: LineageDirectionDownstream,
				Depth:     3,
			},
			wantErr: true,
			errMsg:  "type is required",
		},
		{
			name: "missing direction",
			request: DataLineageRequest{
				Name:    "Test Lineage",
				Dataset: "customer_data",
				Type:    LineageTypeDataFlow,
				Depth:   3,
			},
			wantErr: true,
			errMsg:  "direction is required",
		},
		{
			name: "invalid depth",
			request: DataLineageRequest{
				Name:      "Test Lineage",
				Dataset:   "customer_data",
				Type:      LineageTypeDataFlow,
				Direction: LineageDirectionDownstream,
				Depth:     0,
			},
			wantErr: true,
			errMsg:  "depth must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateLineageRequest(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataLineageHandler_generateLineageNodes(t *testing.T) {
	handler := NewDataLineageHandler(zap.NewNop())

	request := DataLineageRequest{
		Name:      "Test Lineage",
		Dataset:   "customer_data",
		Type:      LineageTypeDataFlow,
		Direction: LineageDirectionDownstream,
		Depth:     3,
		Sources: []LineageSource{
			{
				ID:       "source_1",
				Name:     "Customer Database",
				Type:     "database",
				Location: "postgres://localhost:5432/customers",
				Format:   "postgresql",
				Schema:   map[string]interface{}{},
				Connection: LineageConnection{
					ID:       "conn_1",
					Name:     "Customer DB Connection",
					Type:     "postgresql",
					Protocol: "postgresql",
					Host:     "localhost",
					Port:     5432,
					Database: "customers",
				},
				Properties: map[string]interface{}{},
				Metadata:   map[string]interface{}{},
			},
		},
		Processes: []LineageProcess{
			{
				ID:          "process_1",
				Name:        "Data Transformation",
				Type:        "etl",
				Description: "Transform customer data for analytics",
				Inputs:      []string{"source_1"},
				Outputs:     []string{"target_1"},
				Logic:       "SELECT * FROM customers WHERE active = true",
				Parameters:  map[string]interface{}{},
				Schedule:    "0 2 * * *",
				Status:      "active",
				Metadata:    map[string]interface{}{},
			},
		},
		Targets: []LineageTarget{
			{
				ID:       "target_1",
				Name:     "Analytics Warehouse",
				Type:     "warehouse",
				Location: "bigquery://project/dataset",
				Format:   "bigquery",
				Schema:   map[string]interface{}{},
				Connection: LineageConnection{
					ID:       "conn_2",
					Name:     "BigQuery Connection",
					Type:     "bigquery",
					Protocol: "https",
					Host:     "bigquery.googleapis.com",
					Port:     443,
					Database: "project",
				},
				Properties: map[string]interface{}{},
				Metadata:   map[string]interface{}{},
			},
		},
	}

	nodes := handler.generateLineageNodes(request)

	assert.Len(t, nodes, 3)

	// Check source node
	sourceNode := nodes[0]
	assert.Equal(t, "source_1", sourceNode.ID)
	assert.Equal(t, "Customer Database", sourceNode.Name)
	assert.Equal(t, "database", sourceNode.Type)
	assert.Equal(t, "source", sourceNode.Category)
	assert.Equal(t, "postgres://localhost:5432/customers", sourceNode.Location)
	assert.Equal(t, "active", sourceNode.Status)
	assert.Equal(t, float64(0), sourceNode.Position.X)
	assert.Equal(t, float64(0), sourceNode.Position.Y)

	// Check process node
	processNode := nodes[1]
	assert.Equal(t, "process_1", processNode.ID)
	assert.Equal(t, "Data Transformation", processNode.Name)
	assert.Equal(t, "etl", processNode.Type)
	assert.Equal(t, "process", processNode.Category)
	assert.Equal(t, "internal", processNode.Location)
	assert.Equal(t, "active", processNode.Status)
	assert.Equal(t, float64(0), processNode.Position.X)
	assert.Equal(t, float64(100), processNode.Position.Y)

	// Check target node
	targetNode := nodes[2]
	assert.Equal(t, "target_1", targetNode.ID)
	assert.Equal(t, "Analytics Warehouse", targetNode.Name)
	assert.Equal(t, "warehouse", targetNode.Type)
	assert.Equal(t, "target", targetNode.Category)
	assert.Equal(t, "bigquery://project/dataset", targetNode.Location)
	assert.Equal(t, "active", targetNode.Status)
	assert.Equal(t, float64(0), targetNode.Position.X)
	assert.Equal(t, float64(200), targetNode.Position.Y)
}

func TestDataLineageHandler_generateLineageEdges(t *testing.T) {
	handler := NewDataLineageHandler(zap.NewNop())

	request := DataLineageRequest{
		Name:      "Test Lineage",
		Dataset:   "customer_data",
		Type:      LineageTypeDataFlow,
		Direction: LineageDirectionDownstream,
		Depth:     3,
		Sources: []LineageSource{
			{
				ID:       "source_1",
				Name:     "Customer Database",
				Type:     "database",
				Location: "postgres://localhost:5432/customers",
				Format:   "postgresql",
				Schema:   map[string]interface{}{},
				Connection: LineageConnection{
					ID:       "conn_1",
					Name:     "Customer DB Connection",
					Type:     "postgresql",
					Protocol: "postgresql",
					Host:     "localhost",
					Port:     5432,
					Database: "customers",
				},
				Properties: map[string]interface{}{},
				Metadata:   map[string]interface{}{},
			},
		},
		Processes: []LineageProcess{
			{
				ID:          "process_1",
				Name:        "Data Transformation",
				Type:        "etl",
				Description: "Transform customer data for analytics",
				Inputs:      []string{"source_1"},
				Outputs:     []string{"target_1"},
				Logic:       "SELECT * FROM customers WHERE active = true",
				Parameters:  map[string]interface{}{},
				Schedule:    "0 2 * * *",
				Status:      "active",
				Metadata:    map[string]interface{}{},
			},
		},
		Targets: []LineageTarget{
			{
				ID:       "target_1",
				Name:     "Analytics Warehouse",
				Type:     "warehouse",
				Location: "bigquery://project/dataset",
				Format:   "bigquery",
				Schema:   map[string]interface{}{},
				Connection: LineageConnection{
					ID:       "conn_2",
					Name:     "BigQuery Connection",
					Type:     "bigquery",
					Protocol: "https",
					Host:     "bigquery.googleapis.com",
					Port:     443,
					Database: "project",
				},
				Properties: map[string]interface{}{},
				Metadata:   map[string]interface{}{},
			},
		},
	}

	edges := handler.generateLineageEdges(request)

	assert.Len(t, edges, 2)

	// Check source to process edge
	sourceProcessEdge := edges[0]
	assert.Equal(t, "edge_source_1_process_1", sourceProcessEdge.ID)
	assert.Equal(t, "source_1", sourceProcessEdge.Source)
	assert.Equal(t, "process_1", sourceProcessEdge.Target)
	assert.Equal(t, "data_flow", sourceProcessEdge.Type)
	assert.Equal(t, LineageDirectionDownstream, sourceProcessEdge.Direction)
	assert.Equal(t, "extract", sourceProcessEdge.Properties["flow_type"])
	assert.Equal(t, "daily", sourceProcessEdge.Properties["frequency"])

	// Check process to target edge
	processTargetEdge := edges[1]
	assert.Equal(t, "edge_process_1_target_1", processTargetEdge.ID)
	assert.Equal(t, "process_1", processTargetEdge.Source)
	assert.Equal(t, "target_1", processTargetEdge.Target)
	assert.Equal(t, "data_flow", processTargetEdge.Type)
	assert.Equal(t, LineageDirectionDownstream, processTargetEdge.Direction)
	assert.Equal(t, "load", processTargetEdge.Properties["flow_type"])
	assert.Equal(t, "hourly", processTargetEdge.Properties["frequency"])
}

func TestDataLineageHandler_generateLineagePaths(t *testing.T) {
	handler := NewDataLineageHandler(zap.NewNop())

	request := DataLineageRequest{
		Name:      "Test Lineage",
		Dataset:   "customer_data",
		Type:      LineageTypeDataFlow,
		Direction: LineageDirectionDownstream,
		Depth:     3,
		Sources: []LineageSource{
			{
				ID:       "source_1",
				Name:     "Customer Database",
				Type:     "database",
				Location: "postgres://localhost:5432/customers",
				Format:   "postgresql",
				Schema:   map[string]interface{}{},
				Connection: LineageConnection{
					ID:       "conn_1",
					Name:     "Customer DB Connection",
					Type:     "postgresql",
					Protocol: "postgresql",
					Host:     "localhost",
					Port:     5432,
					Database: "customers",
				},
				Properties: map[string]interface{}{},
				Metadata:   map[string]interface{}{},
			},
		},
		Targets: []LineageTarget{
			{
				ID:       "target_1",
				Name:     "Analytics Warehouse",
				Type:     "warehouse",
				Location: "bigquery://project/dataset",
				Format:   "bigquery",
				Schema:   map[string]interface{}{},
				Connection: LineageConnection{
					ID:       "conn_2",
					Name:     "BigQuery Connection",
					Type:     "bigquery",
					Protocol: "https",
					Host:     "bigquery.googleapis.com",
					Port:     443,
					Database: "project",
				},
				Properties: map[string]interface{}{},
				Metadata:   map[string]interface{}{},
			},
		},
	}

	paths := handler.generateLineagePaths(request)

	assert.Len(t, paths, 1)

	path := paths[0]
	assert.Equal(t, "path_source_1_target_1", path.ID)
	assert.Equal(t, "Path from Customer Database to Analytics Warehouse", path.Name)
	assert.Equal(t, []string{"source_1", "target_1"}, path.Nodes)
	assert.Equal(t, []string{"edge_source_1_target_1"}, path.Edges)
	assert.Equal(t, 2, path.Length)
	assert.Equal(t, "data_flow", path.Type)
	assert.Equal(t, "direct", path.Properties["path_type"])
	assert.Equal(t, "low", path.Properties["complexity"])
}

func TestDataLineageHandler_generateLineageImpact(t *testing.T) {
	handler := NewDataLineageHandler(zap.NewNop())

	request := DataLineageRequest{
		Name:      "Test Lineage",
		Dataset:   "customer_data",
		Type:      LineageTypeDataFlow,
		Direction: LineageDirectionDownstream,
		Depth:     3,
		Sources: []LineageSource{
			{
				ID:       "source_1",
				Name:     "Customer Database",
				Type:     "database",
				Location: "postgres://localhost:5432/customers",
				Format:   "postgresql",
				Schema:   map[string]interface{}{},
				Connection: LineageConnection{
					ID:       "conn_1",
					Name:     "Customer DB Connection",
					Type:     "postgresql",
					Protocol: "postgresql",
					Host:     "localhost",
					Port:     5432,
					Database: "customers",
				},
				Properties: map[string]interface{}{},
				Metadata:   map[string]interface{}{},
			},
		},
		Targets: []LineageTarget{
			{
				ID:       "target_1",
				Name:     "Analytics Warehouse",
				Type:     "warehouse",
				Location: "bigquery://project/dataset",
				Format:   "bigquery",
				Schema:   map[string]interface{}{},
				Connection: LineageConnection{
					ID:       "conn_2",
					Name:     "BigQuery Connection",
					Type:     "bigquery",
					Protocol: "https",
					Host:     "bigquery.googleapis.com",
					Port:     443,
					Database: "project",
				},
				Properties: map[string]interface{}{},
				Metadata:   map[string]interface{}{},
			},
		},
	}

	impact := handler.generateLineageImpact(request)

	assert.Equal(t, []string{"source_1", "target_1"}, impact.AffectedNodes)
	assert.Equal(t, []string{}, impact.AffectedEdges)
	assert.Equal(t, []string{}, impact.AffectedPaths)
	assert.Equal(t, 0.75, impact.ImpactScore)
	assert.Equal(t, "medium", impact.RiskLevel)
	assert.Len(t, impact.Recommendations, 3)
	assert.Contains(t, impact.Recommendations, "Monitor data quality metrics")
	assert.Contains(t, impact.Recommendations, "Implement data validation checks")
	assert.Contains(t, impact.Recommendations, "Set up automated alerts")
	assert.Equal(t, float64(2), impact.Analysis["critical_paths"])
	assert.Equal(t, float64(1), impact.Analysis["bottlenecks"])
	assert.Equal(t, float64(5), impact.Analysis["dependencies"])
}

func TestDataLineageHandler_generateLineageSummary(t *testing.T) {
	handler := NewDataLineageHandler(zap.NewNop())

	request := DataLineageRequest{
		Name:      "Test Lineage",
		Dataset:   "customer_data",
		Type:      LineageTypeDataFlow,
		Direction: LineageDirectionDownstream,
		Depth:     3,
		Sources: []LineageSource{
			{
				ID:       "source_1",
				Name:     "Customer Database",
				Type:     "database",
				Location: "postgres://localhost:5432/customers",
				Format:   "postgresql",
				Schema:   map[string]interface{}{},
				Connection: LineageConnection{
					ID:       "conn_1",
					Name:     "Customer DB Connection",
					Type:     "postgresql",
					Protocol: "postgresql",
					Host:     "localhost",
					Port:     5432,
					Database: "customers",
				},
				Properties: map[string]interface{}{},
				Metadata:   map[string]interface{}{},
			},
		},
		Processes: []LineageProcess{
			{
				ID:          "process_1",
				Name:        "Data Transformation",
				Type:        "etl",
				Description: "Transform customer data for analytics",
				Inputs:      []string{"source_1"},
				Outputs:     []string{"target_1"},
				Logic:       "SELECT * FROM customers WHERE active = true",
				Parameters:  map[string]interface{}{},
				Schedule:    "0 2 * * *",
				Status:      "active",
				Metadata:    map[string]interface{}{},
			},
		},
		Targets: []LineageTarget{
			{
				ID:       "target_1",
				Name:     "Analytics Warehouse",
				Type:     "warehouse",
				Location: "bigquery://project/dataset",
				Format:   "bigquery",
				Schema:   map[string]interface{}{},
				Connection: LineageConnection{
					ID:       "conn_2",
					Name:     "BigQuery Connection",
					Type:     "bigquery",
					Protocol: "https",
					Host:     "bigquery.googleapis.com",
					Port:     443,
					Database: "project",
				},
				Properties: map[string]interface{}{},
				Metadata:   map[string]interface{}{},
			},
		},
	}

	summary := handler.generateLineageSummary(request)

	assert.Equal(t, 3, summary.TotalNodes)
	assert.Equal(t, 2, summary.TotalEdges)
	assert.Equal(t, 1, summary.TotalPaths)
	assert.Equal(t, 1, summary.NodeTypes["source"])
	assert.Equal(t, 1, summary.NodeTypes["process"])
	assert.Equal(t, 1, summary.NodeTypes["target"])
	assert.Equal(t, 2, summary.EdgeTypes["data_flow"])
	assert.Equal(t, 1, summary.PathTypes["data_flow"])
	assert.Equal(t, 3, summary.MaxDepth)
	assert.Equal(t, 2.0, summary.AvgPathLength)
	assert.Equal(t, "medium", summary.Complexity)
	assert.Equal(t, "1.5TB", summary.Metrics["data_volume"])
	assert.Equal(t, "hourly", summary.Metrics["refresh_frequency"])
	assert.Equal(t, 0.95, summary.Metrics["data_quality"])
}

func TestDataLineageHandler_StringConversions(t *testing.T) {
	// Test LineageType string conversion
	lineageType := LineageTypeDataFlow
	assert.Equal(t, "data_flow", lineageType.String())

	// Test LineageStatus string conversion
	lineageStatus := LineageStatusActive
	assert.Equal(t, "active", lineageStatus.String())

	// Test LineageDirection string conversion
	lineageDirection := LineageDirectionDownstream
	assert.Equal(t, "downstream", lineageDirection.String())
}
