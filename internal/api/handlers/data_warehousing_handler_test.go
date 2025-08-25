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

func TestNewDataWarehousingHandler(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataWarehousingHandler(logger)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
	assert.NotNil(t, handler.warehouses)
	assert.NotNil(t, handler.etlProcesses)
	assert.NotNil(t, handler.pipelines)
	assert.NotNil(t, handler.jobs)
	assert.Len(t, handler.warehouses, 0)
	assert.Len(t, handler.etlProcesses, 0)
	assert.Len(t, handler.pipelines, 0)
	assert.Len(t, handler.jobs, 0)
}

func TestDataWarehousingHandler_CreateWarehouse(t *testing.T) {
	tests := []struct {
		name           string
		request        DataWarehouseRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful warehouse creation",
			request: DataWarehouseRequest{
				Name:        "Test Warehouse",
				Type:        WarehouseTypeOLAP,
				Description: "Test data warehouse",
				StorageConfig: StorageConfiguration{
					StorageType: "postgresql",
					Capacity:    "1TB",
					Compression: "gzip",
				},
				SecurityConfig: SecurityConfiguration{
					Encryption: EncryptionConfig{
						Algorithm:     "AES-256",
						KeyManagement: "AWS KMS",
						AtRest:        true,
						InTransit:     true,
					},
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing warehouse name",
			request: DataWarehouseRequest{
				Type: WarehouseTypeOLAP,
				StorageConfig: StorageConfiguration{
					Capacity: "1TB",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "warehouse name is required",
		},
		{
			name: "missing warehouse type",
			request: DataWarehouseRequest{
				Name: "Test Warehouse",
				StorageConfig: StorageConfiguration{
					Capacity: "1TB",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "warehouse type is required",
		},
		{
			name: "missing storage capacity",
			request: DataWarehouseRequest{
				Name: "Test Warehouse",
				Type: WarehouseTypeOLAP,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "storage capacity is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataWarehousingHandler(zap.NewNop())

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/warehouses", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.CreateWarehouse(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response DataWarehouseResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, tt.request.Name, response.Name)
				assert.Equal(t, tt.request.Type, response.Type)
				assert.Equal(t, "creating", response.Status)
			}
		})
	}
}

func TestDataWarehousingHandler_GetWarehouse(t *testing.T) {
	handler := NewDataWarehousingHandler(zap.NewNop())

	// Create a test warehouse
	warehouse := &DataWarehouseResponse{
		ID:     "test_warehouse_123",
		Name:   "Test Warehouse",
		Type:   WarehouseTypeOLAP,
		Status: "active",
	}
	handler.warehouses[warehouse.ID] = warehouse

	tests := []struct {
		name           string
		warehouseID    string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful warehouse retrieval",
			warehouseID:    "test_warehouse_123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing warehouse ID",
			warehouseID:    "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Warehouse ID is required",
		},
		{
			name:           "warehouse not found",
			warehouseID:    "nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Warehouse not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/warehouses?id=%s", tt.warehouseID), nil)
			w := httptest.NewRecorder()

			handler.GetWarehouse(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response DataWarehouseResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, warehouse.ID, response.ID)
				assert.Equal(t, warehouse.Name, response.Name)
			}
		})
	}
}

func TestDataWarehousingHandler_ListWarehouses(t *testing.T) {
	handler := NewDataWarehousingHandler(zap.NewNop())

	// Create test warehouses
	warehouse1 := &DataWarehouseResponse{
		ID:     "warehouse_1",
		Name:   "Warehouse 1",
		Type:   WarehouseTypeOLAP,
		Status: "active",
	}
	warehouse2 := &DataWarehouseResponse{
		ID:     "warehouse_2",
		Name:   "Warehouse 2",
		Type:   WarehouseTypeDataLake,
		Status: "active",
	}
	handler.warehouses[warehouse1.ID] = warehouse1
	handler.warehouses[warehouse2.ID] = warehouse2

	req := httptest.NewRequest("GET", "/warehouses", nil)
	w := httptest.NewRecorder()

	handler.ListWarehouses(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(2), response["count"])

	warehouses := response["warehouses"].([]interface{})
	assert.Len(t, warehouses, 2)
}

func TestDataWarehousingHandler_CreateETLProcess(t *testing.T) {
	tests := []struct {
		name           string
		request        ETLProcessRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful ETL process creation",
			request: ETLProcessRequest{
				Name:        "Test ETL",
				Type:        ETLProcessTypeFull,
				Description: "Test ETL process",
				SourceConfig: SourceConfiguration{
					SourceType:       "postgresql",
					ConnectionString: "postgres://user:pass@localhost:5432/db",
					Query:            "SELECT * FROM source_table",
					BatchSize:        1000,
				},
				TransformConfig: TransformConfiguration{
					Transformations: []TransformationRule{
						{
							Name:       "clean_data",
							Type:       "filter",
							Expression: "status = 'active'",
						},
					},
				},
				TargetConfig: TargetConfiguration{
					TargetType:       "postgresql",
					ConnectionString: "postgres://user:pass@localhost:5432/warehouse",
					TableName:        "target_table",
					LoadStrategy:     "insert",
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing ETL process name",
			request: ETLProcessRequest{
				Type: ETLProcessTypeFull,
				SourceConfig: SourceConfiguration{
					SourceType: "postgresql",
				},
				TargetConfig: TargetConfiguration{
					TargetType: "postgresql",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "ETL process name is required",
		},
		{
			name: "missing ETL process type",
			request: ETLProcessRequest{
				Name: "Test ETL",
				SourceConfig: SourceConfiguration{
					SourceType: "postgresql",
				},
				TargetConfig: TargetConfiguration{
					TargetType: "postgresql",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "ETL process type is required",
		},
		{
			name: "missing source type",
			request: ETLProcessRequest{
				Name: "Test ETL",
				Type: ETLProcessTypeFull,
				TargetConfig: TargetConfiguration{
					TargetType: "postgresql",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "source type is required",
		},
		{
			name: "missing target type",
			request: ETLProcessRequest{
				Name: "Test ETL",
				Type: ETLProcessTypeFull,
				SourceConfig: SourceConfiguration{
					SourceType: "postgresql",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "target type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataWarehousingHandler(zap.NewNop())

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/etl", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.CreateETLProcess(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response ETLProcessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, tt.request.Name, response.Name)
				assert.Equal(t, tt.request.Type, response.Type)
				assert.Equal(t, PipelineStatusPending, response.Status)
			}
		})
	}
}

func TestDataWarehousingHandler_GetETLProcess(t *testing.T) {
	handler := NewDataWarehousingHandler(zap.NewNop())

	// Create a test ETL process
	etlProcess := &ETLProcessResponse{
		ID:     "test_etl_123",
		Name:   "Test ETL",
		Type:   ETLProcessTypeFull,
		Status: PipelineStatusPending,
	}
	handler.etlProcesses[etlProcess.ID] = etlProcess

	tests := []struct {
		name           string
		etlID          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful ETL process retrieval",
			etlID:          "test_etl_123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing ETL process ID",
			etlID:          "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "ETL process ID is required",
		},
		{
			name:           "ETL process not found",
			etlID:          "nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedError:  "ETL process not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/etl?id=%s", tt.etlID), nil)
			w := httptest.NewRecorder()

			handler.GetETLProcess(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response ETLProcessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, etlProcess.ID, response.ID)
				assert.Equal(t, etlProcess.Name, response.Name)
			}
		})
	}
}

func TestDataWarehousingHandler_ListETLProcesses(t *testing.T) {
	handler := NewDataWarehousingHandler(zap.NewNop())

	// Create test ETL processes
	etl1 := &ETLProcessResponse{
		ID:     "etl_1",
		Name:   "ETL Process 1",
		Type:   ETLProcessTypeFull,
		Status: PipelineStatusPending,
	}
	etl2 := &ETLProcessResponse{
		ID:     "etl_2",
		Name:   "ETL Process 2",
		Type:   ETLProcessTypeIncremental,
		Status: PipelineStatusCompleted,
	}
	handler.etlProcesses[etl1.ID] = etl1
	handler.etlProcesses[etl2.ID] = etl2

	req := httptest.NewRequest("GET", "/etl", nil)
	w := httptest.NewRecorder()

	handler.ListETLProcesses(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(2), response["count"])

	etlProcesses := response["etl_processes"].([]interface{})
	assert.Len(t, etlProcesses, 2)
}

func TestDataWarehousingHandler_CreatePipeline(t *testing.T) {
	tests := []struct {
		name           string
		request        DataPipelineRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful pipeline creation",
			request: DataPipelineRequest{
				Name:        "Test Pipeline",
				Description: "Test data pipeline",
				Stages: []PipelineStage{
					{
						Name:  "extract",
						Type:  "extract",
						Order: 1,
						Configuration: map[string]interface{}{
							"source": "database",
						},
					},
					{
						Name:  "transform",
						Type:  "transform",
						Order: 2,
						Configuration: map[string]interface{}{
							"operations": []string{"clean", "validate"},
						},
					},
					{
						Name:  "load",
						Type:  "load",
						Order: 3,
						Configuration: map[string]interface{}{
							"target": "warehouse",
						},
					},
				},
				Triggers: []PipelineTrigger{
					{
						Name:     "daily",
						Type:     "schedule",
						Schedule: "0 0 * * *",
					},
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing pipeline name",
			request: DataPipelineRequest{
				Stages: []PipelineStage{
					{
						Name:  "extract",
						Type:  "extract",
						Order: 1,
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "pipeline name is required",
		},
		{
			name: "missing pipeline stages",
			request: DataPipelineRequest{
				Name: "Test Pipeline",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "at least one pipeline stage is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataWarehousingHandler(zap.NewNop())

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/pipelines", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.CreatePipeline(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response DataPipelineResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, tt.request.Name, response.Name)
				assert.Equal(t, PipelineStatusPending, response.Status)
				assert.Len(t, response.Stages, len(tt.request.Stages))
			}
		})
	}
}

func TestDataWarehousingHandler_GetPipeline(t *testing.T) {
	handler := NewDataWarehousingHandler(zap.NewNop())

	// Create a test pipeline
	pipeline := &DataPipelineResponse{
		ID:     "test_pipeline_123",
		Name:   "Test Pipeline",
		Status: PipelineStatusPending,
		Stages: []PipelineStageStatus{
			{
				Name:   "extract",
				Status: PipelineStatusPending,
			},
		},
	}
	handler.pipelines[pipeline.ID] = pipeline

	tests := []struct {
		name           string
		pipelineID     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful pipeline retrieval",
			pipelineID:     "test_pipeline_123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing pipeline ID",
			pipelineID:     "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Pipeline ID is required",
		},
		{
			name:           "pipeline not found",
			pipelineID:     "nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Pipeline not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/pipelines?id=%s", tt.pipelineID), nil)
			w := httptest.NewRecorder()

			handler.GetPipeline(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response DataPipelineResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, pipeline.ID, response.ID)
				assert.Equal(t, pipeline.Name, response.Name)
			}
		})
	}
}

func TestDataWarehousingHandler_ListPipelines(t *testing.T) {
	handler := NewDataWarehousingHandler(zap.NewNop())

	// Create test pipelines
	pipeline1 := &DataPipelineResponse{
		ID:     "pipeline_1",
		Name:   "Pipeline 1",
		Status: PipelineStatusPending,
	}
	pipeline2 := &DataPipelineResponse{
		ID:     "pipeline_2",
		Name:   "Pipeline 2",
		Status: PipelineStatusCompleted,
	}
	handler.pipelines[pipeline1.ID] = pipeline1
	handler.pipelines[pipeline2.ID] = pipeline2

	req := httptest.NewRequest("GET", "/pipelines", nil)
	w := httptest.NewRecorder()

	handler.ListPipelines(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(2), response["count"])

	pipelines := response["pipelines"].([]interface{})
	assert.Len(t, pipelines, 2)
}

func TestDataWarehousingHandler_CreateWarehouseJob(t *testing.T) {
	tests := []struct {
		name           string
		request        map[string]interface{}
		expectedStatus int
	}{
		{
			name: "successful job creation",
			request: map[string]interface{}{
				"type":         "backup",
				"warehouse_id": "warehouse_123",
				"configuration": map[string]interface{}{
					"backup_type": "full",
					"compression": true,
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "job creation with minimal config",
			request: map[string]interface{}{
				"type":         "maintenance",
				"warehouse_id": "warehouse_456",
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataWarehousingHandler(zap.NewNop())

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/warehouse/jobs", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.CreateWarehouseJob(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.NotEmpty(t, response["job_id"])
			assert.Equal(t, "pending", response["status"])
			assert.NotEmpty(t, response["created_at"])
		})
	}
}

func TestDataWarehousingHandler_GetWarehouseJob(t *testing.T) {
	handler := NewDataWarehousingHandler(zap.NewNop())

	// Create a test job
	job := &WarehouseJob{
		ID:        "test_job_123",
		Type:      "backup",
		Status:    "running",
		Progress:  50,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	handler.jobs[job.ID] = job

	tests := []struct {
		name           string
		jobID          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful job retrieval",
			jobID:          "test_job_123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing job ID",
			jobID:          "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Job ID is required",
		},
		{
			name:           "job not found",
			jobID:          "nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Job not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/warehouse/jobs?id=%s", tt.jobID), nil)
			w := httptest.NewRecorder()

			handler.GetWarehouseJob(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response WarehouseJob
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, job.ID, response.ID)
				assert.Equal(t, job.Type, response.Type)
				assert.Equal(t, job.Status, response.Status)
			}
		})
	}
}

func TestDataWarehousingHandler_ListWarehouseJobs(t *testing.T) {
	handler := NewDataWarehousingHandler(zap.NewNop())

	// Create test jobs
	job1 := &WarehouseJob{
		ID:        "job_1",
		Type:      "backup",
		Status:    "completed",
		Progress:  100,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	job2 := &WarehouseJob{
		ID:        "job_2",
		Type:      "maintenance",
		Status:    "running",
		Progress:  75,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	handler.jobs[job1.ID] = job1
	handler.jobs[job2.ID] = job2

	req := httptest.NewRequest("GET", "/warehouse/jobs", nil)
	w := httptest.NewRecorder()

	handler.ListWarehouseJobs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(2), response["count"])

	jobs := response["jobs"].([]interface{})
	assert.Len(t, jobs, 2)
}

func TestDataWarehousingHandler_validateWarehouseRequest(t *testing.T) {
	handler := NewDataWarehousingHandler(zap.NewNop())

	tests := []struct {
		name    string
		request DataWarehouseRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: DataWarehouseRequest{
				Name: "Test Warehouse",
				Type: WarehouseTypeOLAP,
				StorageConfig: StorageConfiguration{
					Capacity: "1TB",
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			request: DataWarehouseRequest{
				Type: WarehouseTypeOLAP,
				StorageConfig: StorageConfiguration{
					Capacity: "1TB",
				},
			},
			wantErr: true,
			errMsg:  "warehouse name is required",
		},
		{
			name: "missing type",
			request: DataWarehouseRequest{
				Name: "Test Warehouse",
				StorageConfig: StorageConfiguration{
					Capacity: "1TB",
				},
			},
			wantErr: true,
			errMsg:  "warehouse type is required",
		},
		{
			name: "missing storage capacity",
			request: DataWarehouseRequest{
				Name: "Test Warehouse",
				Type: WarehouseTypeOLAP,
			},
			wantErr: true,
			errMsg:  "storage capacity is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateWarehouseRequest(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataWarehousingHandler_validateETLRequest(t *testing.T) {
	handler := NewDataWarehousingHandler(zap.NewNop())

	tests := []struct {
		name    string
		request ETLProcessRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: ETLProcessRequest{
				Name: "Test ETL",
				Type: ETLProcessTypeFull,
				SourceConfig: SourceConfiguration{
					SourceType: "postgresql",
				},
				TargetConfig: TargetConfiguration{
					TargetType: "postgresql",
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			request: ETLProcessRequest{
				Type: ETLProcessTypeFull,
				SourceConfig: SourceConfiguration{
					SourceType: "postgresql",
				},
				TargetConfig: TargetConfiguration{
					TargetType: "postgresql",
				},
			},
			wantErr: true,
			errMsg:  "ETL process name is required",
		},
		{
			name: "missing type",
			request: ETLProcessRequest{
				Name: "Test ETL",
				SourceConfig: SourceConfiguration{
					SourceType: "postgresql",
				},
				TargetConfig: TargetConfiguration{
					TargetType: "postgresql",
				},
			},
			wantErr: true,
			errMsg:  "ETL process type is required",
		},
		{
			name: "missing source type",
			request: ETLProcessRequest{
				Name: "Test ETL",
				Type: ETLProcessTypeFull,
				TargetConfig: TargetConfiguration{
					TargetType: "postgresql",
				},
			},
			wantErr: true,
			errMsg:  "source type is required",
		},
		{
			name: "missing target type",
			request: ETLProcessRequest{
				Name: "Test ETL",
				Type: ETLProcessTypeFull,
				SourceConfig: SourceConfiguration{
					SourceType: "postgresql",
				},
			},
			wantErr: true,
			errMsg:  "target type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateETLRequest(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataWarehousingHandler_validatePipelineRequest(t *testing.T) {
	handler := NewDataWarehousingHandler(zap.NewNop())

	tests := []struct {
		name    string
		request DataPipelineRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: DataPipelineRequest{
				Name: "Test Pipeline",
				Stages: []PipelineStage{
					{
						Name:  "extract",
						Type:  "extract",
						Order: 1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			request: DataPipelineRequest{
				Stages: []PipelineStage{
					{
						Name:  "extract",
						Type:  "extract",
						Order: 1,
					},
				},
			},
			wantErr: true,
			errMsg:  "pipeline name is required",
		},
		{
			name: "missing stages",
			request: DataPipelineRequest{
				Name: "Test Pipeline",
			},
			wantErr: true,
			errMsg:  "at least one pipeline stage is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validatePipelineRequest(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWarehouseType_String(t *testing.T) {
	tests := []struct {
		warehouseType WarehouseType
		expected      string
	}{
		{WarehouseTypeOLTP, "oltp"},
		{WarehouseTypeOLAP, "olap"},
		{WarehouseTypeDataLake, "data_lake"},
		{WarehouseTypeDataMart, "data_mart"},
		{WarehouseTypeHybrid, "hybrid"},
	}

	for _, tt := range tests {
		t.Run(string(tt.warehouseType), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.warehouseType))
		})
	}
}

func TestETLProcessType_String(t *testing.T) {
	tests := []struct {
		etlType  ETLProcessType
		expected string
	}{
		{ETLProcessTypeExtract, "extract"},
		{ETLProcessTypeTransform, "transform"},
		{ETLProcessTypeLoad, "load"},
		{ETLProcessTypeFull, "full"},
		{ETLProcessTypeIncremental, "incremental"},
	}

	for _, tt := range tests {
		t.Run(string(tt.etlType), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.etlType))
		})
	}
}

func TestPipelineStatus_String(t *testing.T) {
	tests := []struct {
		status   PipelineStatus
		expected string
	}{
		{PipelineStatusPending, "pending"},
		{PipelineStatusRunning, "running"},
		{PipelineStatusCompleted, "completed"},
		{PipelineStatusFailed, "failed"},
		{PipelineStatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}
