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

func TestNewDataCatalogHandler(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataCatalogHandler(logger)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
	assert.NotNil(t, handler.catalogs)
	assert.NotNil(t, handler.jobs)
	assert.Len(t, handler.catalogs, 0)
	assert.Len(t, handler.jobs, 0)
}

func TestDataCatalogHandler_CreateCatalog(t *testing.T) {
	tests := []struct {
		name           string
		request        DataCatalogRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful catalog creation",
			request: DataCatalogRequest{
				Name:        "Enterprise Data Catalog",
				Description: "Comprehensive enterprise data catalog",
				Type:        CatalogTypeDatabase,
				Category:    "enterprise",
				Assets: []CatalogAsset{
					{
						ID:          "asset_1",
						Name:        "Customer Database",
						Type:        AssetTypeDataset,
						Description: "Main customer database",
						Location:    "postgres://localhost:5432/customers",
						Format:      "postgresql",
						Size:        1024000000,
						Schema: AssetSchema{
							Type:    "relational",
							Version: "1.0",
							Columns: []SchemaColumn{
								{
									Name:        "customer_id",
									Type:        "integer",
									Description: "Unique customer identifier",
									Nullable:    false,
									PrimaryKey:  true,
									Properties:  make(map[string]interface{}),
								},
								{
									Name:        "name",
									Type:        "varchar",
									Description: "Customer name",
									Nullable:    false,
									Length:      255,
									Properties:  make(map[string]interface{}),
								},
								{
									Name:        "email",
									Type:        "varchar",
									Description: "Customer email",
									Nullable:    true,
									Length:      255,
									Unique:      true,
									Properties:  make(map[string]interface{}),
								},
							},
							Constraints: []SchemaConstraint{
								{
									Name:    "pk_customer",
									Type:    "primary_key",
									Columns: []string{"customer_id"},
									Enabled: true,
								},
								{
									Name:    "uk_email",
									Type:    "unique",
									Columns: []string{"email"},
									Enabled: true,
								},
							},
							Indexes: []SchemaIndex{
								{
									Name:    "idx_customer_email",
									Type:    "btree",
									Columns: []string{"email"},
									Unique:  true,
								},
							},
							Properties: make(map[string]interface{}),
						},
						Connection: AssetConnection{
							ID:         "conn_1",
							Name:       "Customer DB Connection",
							Type:       "postgresql",
							Protocol:   "postgresql",
							Host:       "localhost",
							Port:       5432,
							Database:   "customers",
							Schema:     "public",
							Properties: make(map[string]interface{}),
						},
						Classification: "confidential",
						Sensitivity:    "high",
						Tags:           []string{"customer", "pii", "production"},
						Properties:     make(map[string]interface{}),
						Metadata:       make(map[string]interface{}),
						CreatedAt:      time.Now(),
						UpdatedAt:      time.Now(),
					},
				},
				Collections: []CatalogCollection{
					{
						ID:          "collection_1",
						Name:        "Customer Data Collection",
						Description: "Collection of customer-related datasets",
						Type:        "business",
						Assets:      []string{"asset_1"},
						Owner:       "data_team",
						Tags:        []string{"customer", "core"},
						Properties:  make(map[string]interface{}),
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				},
				Schemas: []CatalogSchema{
					{
						ID:          "schema_1",
						Name:        "Customer Schema",
						Version:     "1.0",
						Description: "Customer data schema definition",
						Type:        "json_schema",
						Content: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"customer_id": map[string]interface{}{
									"type": "integer",
								},
								"name": map[string]interface{}{
									"type": "string",
								},
								"email": map[string]interface{}{
									"type":   "string",
									"format": "email",
								},
							},
						},
						Assets:     []string{"asset_1"},
						Properties: make(map[string]interface{}),
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					},
				},
				Connections: []CatalogConnection{
					{
						ID:          "conn_1",
						Name:        "Customer DB Connection",
						Type:        "postgresql",
						Description: "Connection to customer database",
						Protocol:    "postgresql",
						Host:        "localhost",
						Port:        5432,
						Database:    "customers",
						Schema:      "public",
						Credentials: map[string]interface{}{
							"username":  "app_user",
							"auth_type": "password",
						},
						Properties: map[string]interface{}{
							"pool_size": 10,
							"timeout":   30,
						},
						Status:    "active",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				},
				Tags:     []string{"enterprise", "production", "customer"},
				Owners:   []string{"data_team", "engineering"},
				Stewards: []string{"data_steward_1", "data_steward_2"},
				Domains:  []string{"customer", "finance", "operations"},
				Options: CatalogOptions{
					AutoDiscovery:   true,
					IncludeMetadata: true,
					IncludeSchema:   true,
					IncludeLineage:  true,
					IncludeUsage:    true,
					IncludeQuality:  true,
					ScanFrequency:   "daily",
					NotifyChanges:   true,
					Custom:          make(map[string]interface{}),
				},
				Metadata: map[string]interface{}{
					"priority":    "high",
					"environment": "production",
					"region":      "us-east-1",
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			request: DataCatalogRequest{
				Type:     CatalogTypeDatabase,
				Category: "enterprise",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name is required",
		},
		{
			name: "missing type",
			request: DataCatalogRequest{
				Name:     "Test Catalog",
				Category: "enterprise",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "type is required",
		},
		{
			name: "missing category",
			request: DataCatalogRequest{
				Name: "Test Catalog",
				Type: CatalogTypeDatabase,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "category is required",
		},
	}

	handler := NewDataCatalogHandler(zap.NewNop())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/catalog", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.CreateCatalog(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response DataCatalogResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.NotEmpty(t, response.ID)
				assert.Equal(t, tt.request.Name, response.Name)
				assert.Equal(t, tt.request.Type, response.Type)
				assert.Equal(t, CatalogStatusActive, response.Status)
				assert.Equal(t, tt.request.Category, response.Category)
				assert.NotNil(t, response.Assets)
				assert.NotNil(t, response.Collections)
				assert.NotNil(t, response.Schemas)
				assert.NotNil(t, response.Connections)
				assert.NotNil(t, response.Summary)
				assert.NotNil(t, response.Statistics)
				assert.NotNil(t, response.Health)
				assert.NotZero(t, response.CreatedAt)
				assert.NotZero(t, response.UpdatedAt)
			}
		})
	}
}

func TestDataCatalogHandler_GetCatalog(t *testing.T) {
	handler := NewDataCatalogHandler(zap.NewNop())

	// Create a test catalog
	catalog := &DataCatalogResponse{
		ID:          "test_catalog_123",
		Name:        "Test Catalog",
		Type:        CatalogTypeDatabase,
		Status:      CatalogStatusActive,
		Category:    "test",
		Assets:      []CatalogAsset{},
		Collections: []CatalogCollection{},
		Schemas:     []CatalogSchema{},
		Connections: []CatalogConnection{},
		Summary:     CatalogSummary{},
		Statistics:  CatalogStatistics{},
		Health:      CatalogHealth{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	handler.catalogs["test_catalog_123"] = catalog

	tests := []struct {
		name           string
		id             string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful retrieval",
			id:             "test_catalog_123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing id",
			id:             "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Catalog ID is required",
		},
		{
			name:           "not found",
			id:             "nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Catalog not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/catalog"
			if tt.id != "" {
				url = fmt.Sprintf("/catalog?id=%s", tt.id)
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.GetCatalog(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response DataCatalogResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, catalog.ID, response.ID)
				assert.Equal(t, catalog.Name, response.Name)
				assert.Equal(t, catalog.Type, response.Type)
				assert.Equal(t, catalog.Status, response.Status)
			}
		})
	}
}

func TestDataCatalogHandler_ListCatalogs(t *testing.T) {
	handler := NewDataCatalogHandler(zap.NewNop())

	// Create test catalogs
	catalog1 := &DataCatalogResponse{
		ID:        "catalog_1",
		Name:      "Catalog 1",
		Type:      CatalogTypeDatabase,
		Status:    CatalogStatusActive,
		Category:  "test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	catalog2 := &DataCatalogResponse{
		ID:        "catalog_2",
		Name:      "Catalog 2",
		Type:      CatalogTypeTable,
		Status:    CatalogStatusActive,
		Category:  "test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	handler.catalogs["catalog_1"] = catalog1
	handler.catalogs["catalog_2"] = catalog2

	req := httptest.NewRequest("GET", "/catalog", nil)
	w := httptest.NewRecorder()

	handler.ListCatalogs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(2), response["total"])

	catalogs := response["catalogs"].([]interface{})
	assert.Len(t, catalogs, 2)
}

func TestDataCatalogHandler_CreateCatalogJob(t *testing.T) {
	tests := []struct {
		name           string
		request        DataCatalogRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful job creation",
			request: DataCatalogRequest{
				Name:        "Test Catalog Job",
				Description: "Test background catalog job",
				Type:        CatalogTypeDatabase,
				Category:    "test",
				Assets: []CatalogAsset{
					{
						ID:          "asset_1",
						Name:        "Test Asset",
						Type:        AssetTypeDataset,
						Description: "Test dataset",
						Location:    "test://localhost:5432/test",
						Format:      "postgresql",
						Size:        1000000,
						Schema:      AssetSchema{Type: "relational", Version: "1.0"},
						Connection:  AssetConnection{ID: "conn_1", Name: "Test Connection", Type: "postgresql"},
						Tags:        []string{"test"},
						Properties:  make(map[string]interface{}),
						Metadata:    make(map[string]interface{}),
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				},
				Collections: []CatalogCollection{},
				Schemas:     []CatalogSchema{},
				Connections: []CatalogConnection{},
				Tags:        []string{"test"},
				Owners:      []string{"test_team"},
				Stewards:    []string{"test_steward"},
				Domains:     []string{"test"},
				Options: CatalogOptions{
					AutoDiscovery:   true,
					IncludeMetadata: true,
					IncludeSchema:   true,
					IncludeLineage:  false,
					IncludeUsage:    false,
					IncludeQuality:  false,
					ScanFrequency:   "manual",
					NotifyChanges:   false,
					Custom:          make(map[string]interface{}),
				},
				Metadata: map[string]interface{}{
					"priority": "low",
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request",
			request: DataCatalogRequest{
				Name:     "",
				Type:     CatalogTypeDatabase,
				Category: "test",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name is required",
		},
	}

	handler := NewDataCatalogHandler(zap.NewNop())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/catalog/jobs", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.CreateCatalogJob(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var job CatalogJob
				err := json.Unmarshal(w.Body.Bytes(), &job)
				require.NoError(t, err)

				assert.NotEmpty(t, job.ID)
				assert.NotEmpty(t, job.RequestID)
				assert.Equal(t, "catalog_creation", job.Type)
				assert.Equal(t, "pending", job.Status)
				assert.Equal(t, 0, job.Progress)
				assert.NotZero(t, job.CreatedAt)
				assert.NotZero(t, job.UpdatedAt)
				assert.Nil(t, job.CompletedAt)
			}
		})
	}
}

func TestDataCatalogHandler_GetCatalogJob(t *testing.T) {
	handler := NewDataCatalogHandler(zap.NewNop())

	// Create a test job
	job := &CatalogJob{
		ID:        "test_job_123",
		RequestID: "req_123",
		Type:      "catalog_creation",
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
			url := "/catalog/jobs"
			if tt.id != "" {
				url = fmt.Sprintf("/catalog/jobs?id=%s", tt.id)
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.GetCatalogJob(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response CatalogJob
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, job.ID, response.ID)
				assert.Equal(t, job.RequestID, response.RequestID)
				assert.Equal(t, job.Type, response.Type)
				assert.Equal(t, job.Status, response.Status)
				assert.Equal(t, job.Progress, response.Progress)
			}
		})
	}
}

func TestDataCatalogHandler_ListCatalogJobs(t *testing.T) {
	handler := NewDataCatalogHandler(zap.NewNop())

	// Create test jobs
	job1 := &CatalogJob{
		ID:        "job_1",
		RequestID: "req_1",
		Type:      "catalog_creation",
		Status:    "completed",
		Progress:  100,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	job2 := &CatalogJob{
		ID:        "job_2",
		RequestID: "req_2",
		Type:      "catalog_creation",
		Status:    "running",
		Progress:  50,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	handler.jobs["job_1"] = job1
	handler.jobs["job_2"] = job2

	req := httptest.NewRequest("GET", "/catalog/jobs", nil)
	w := httptest.NewRecorder()

	handler.ListCatalogJobs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(2), response["total"])

	jobs := response["jobs"].([]interface{})
	assert.Len(t, jobs, 2)
}

func TestDataCatalogHandler_validateCatalogRequest(t *testing.T) {
	handler := NewDataCatalogHandler(zap.NewNop())

	tests := []struct {
		name    string
		request DataCatalogRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: DataCatalogRequest{
				Name:     "Test Catalog",
				Type:     CatalogTypeDatabase,
				Category: "test",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			request: DataCatalogRequest{
				Type:     CatalogTypeDatabase,
				Category: "test",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing type",
			request: DataCatalogRequest{
				Name:     "Test Catalog",
				Category: "test",
			},
			wantErr: true,
			errMsg:  "type is required",
		},
		{
			name: "missing category",
			request: DataCatalogRequest{
				Name: "Test Catalog",
				Type: CatalogTypeDatabase,
			},
			wantErr: true,
			errMsg:  "category is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateCatalogRequest(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataCatalogHandler_processAssets(t *testing.T) {
	handler := NewDataCatalogHandler(zap.NewNop())

	assets := []CatalogAsset{
		{
			ID:          "asset_1",
			Name:        "Test Asset",
			Type:        AssetTypeDataset,
			Description: "Test dataset",
			Location:    "test://localhost:5432/test",
			Format:      "postgresql",
			Size:        1000000,
			Schema: AssetSchema{
				Type:    "relational",
				Version: "1.0",
				Columns: []SchemaColumn{
					{
						Name:        "id",
						Type:        "integer",
						Description: "Primary key",
						Nullable:    false,
						PrimaryKey:  true,
						Properties:  make(map[string]interface{}),
					},
				},
				Properties: make(map[string]interface{}),
			},
			Connection: AssetConnection{
				ID:         "conn_1",
				Name:       "Test Connection",
				Type:       "postgresql",
				Protocol:   "postgresql",
				Host:       "localhost",
				Port:       5432,
				Database:   "test",
				Properties: make(map[string]interface{}),
			},
			Tags:       []string{"test"},
			Properties: make(map[string]interface{}),
			Metadata:   make(map[string]interface{}),
		},
	}

	processedAssets := handler.processAssets(assets)

	assert.Len(t, processedAssets, 1)

	asset := processedAssets[0]
	assert.Equal(t, "asset_1", asset.ID)
	assert.Equal(t, "Test Asset", asset.Name)
	assert.Equal(t, AssetTypeDataset, asset.Type)
	assert.NotZero(t, asset.CreatedAt)
	assert.NotZero(t, asset.UpdatedAt)
	assert.NotZero(t, asset.Quality.Score)
	assert.NotZero(t, asset.Usage.AccessCount)
}

func TestDataCatalogHandler_processCollections(t *testing.T) {
	handler := NewDataCatalogHandler(zap.NewNop())

	collections := []CatalogCollection{
		{
			ID:          "collection_1",
			Name:        "Test Collection",
			Description: "Test collection description",
			Type:        "business",
			Assets:      []string{"asset_1"},
			Owner:       "test_team",
			Tags:        []string{"test"},
			Properties:  make(map[string]interface{}),
		},
	}

	processedCollections := handler.processCollections(collections)

	assert.Len(t, processedCollections, 1)

	collection := processedCollections[0]
	assert.Equal(t, "collection_1", collection.ID)
	assert.Equal(t, "Test Collection", collection.Name)
	assert.NotZero(t, collection.CreatedAt)
	assert.NotZero(t, collection.UpdatedAt)
}

func TestDataCatalogHandler_processSchemas(t *testing.T) {
	handler := NewDataCatalogHandler(zap.NewNop())

	schemas := []CatalogSchema{
		{
			ID:          "schema_1",
			Name:        "Test Schema",
			Version:     "1.0",
			Description: "Test schema description",
			Type:        "json_schema",
			Content: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type": "integer",
					},
				},
			},
			Assets:     []string{"asset_1"},
			Properties: make(map[string]interface{}),
		},
	}

	processedSchemas := handler.processSchemas(schemas)

	assert.Len(t, processedSchemas, 1)

	schema := processedSchemas[0]
	assert.Equal(t, "schema_1", schema.ID)
	assert.Equal(t, "Test Schema", schema.Name)
	assert.NotZero(t, schema.CreatedAt)
	assert.NotZero(t, schema.UpdatedAt)
}

func TestDataCatalogHandler_processConnections(t *testing.T) {
	handler := NewDataCatalogHandler(zap.NewNop())

	connections := []CatalogConnection{
		{
			ID:          "conn_1",
			Name:        "Test Connection",
			Type:        "postgresql",
			Description: "Test connection description",
			Protocol:    "postgresql",
			Host:        "localhost",
			Port:        5432,
			Database:    "test",
			Schema:      "public",
			Credentials: map[string]interface{}{
				"username": "test_user",
			},
			Properties: make(map[string]interface{}),
		},
	}

	processedConnections := handler.processConnections(connections)

	assert.Len(t, processedConnections, 1)

	connection := processedConnections[0]
	assert.Equal(t, "conn_1", connection.ID)
	assert.Equal(t, "Test Connection", connection.Name)
	assert.Equal(t, "active", connection.Status)
	assert.NotZero(t, connection.CreatedAt)
	assert.NotZero(t, connection.UpdatedAt)
}

func TestDataCatalogHandler_generateCatalogSummary(t *testing.T) {
	handler := NewDataCatalogHandler(zap.NewNop())

	request := DataCatalogRequest{
		Name:     "Test Catalog",
		Type:     CatalogTypeDatabase,
		Category: "test",
		Assets: []CatalogAsset{
			{
				ID:   "asset_1",
				Name: "Asset 1",
				Type: AssetTypeDataset,
			},
			{
				ID:   "asset_2",
				Name: "Asset 2",
				Type: AssetTypeSchema,
			},
		},
		Collections: []CatalogCollection{
			{
				ID:   "collection_1",
				Name: "Collection 1",
			},
		},
		Schemas: []CatalogSchema{
			{
				ID:   "schema_1",
				Name: "Schema 1",
			},
		},
		Connections: []CatalogConnection{
			{
				ID:   "conn_1",
				Name: "Connection 1",
			},
		},
	}

	summary := handler.generateCatalogSummary(request)

	assert.Equal(t, 2, summary.TotalAssets)
	assert.Equal(t, 1, summary.TotalCollections)
	assert.Equal(t, 1, summary.TotalSchemas)
	assert.Equal(t, 1, summary.TotalConnections)
	assert.Equal(t, 1, summary.AssetTypes[string(AssetTypeDataset)])
	assert.Equal(t, 1, summary.AssetTypes[string(AssetTypeSchema)])
	assert.Equal(t, 2, summary.AssetStatuses["active"])
	assert.Equal(t, "2.5TB", summary.DataVolume)
	assert.Equal(t, 0.85, summary.Coverage)
	assert.Equal(t, 0.90, summary.Completeness)
	assert.NotZero(t, summary.LastUpdate)
	assert.NotNil(t, summary.Metrics)
}

func TestDataCatalogHandler_generateCatalogStatistics(t *testing.T) {
	handler := NewDataCatalogHandler(zap.NewNop())

	request := DataCatalogRequest{
		Name:     "Test Catalog",
		Type:     CatalogTypeDatabase,
		Category: "test",
		Assets: []CatalogAsset{
			{ID: "asset_1", Name: "Asset 1", Type: AssetTypeDataset},
			{ID: "asset_2", Name: "Asset 2", Type: AssetTypeSchema},
		},
	}

	statistics := handler.generateCatalogStatistics(request)

	assert.Equal(t, int64(15000), statistics.AccessStats.TotalAccess)
	assert.Equal(t, 85, statistics.AccessStats.UniqueUsers)
	assert.NotEmpty(t, statistics.AccessStats.PopularAssets)
	assert.Equal(t, 0.85, statistics.QualityStats.OverallScore)
	assert.Equal(t, 0, statistics.QualityStats.PassingAssets)
	assert.Equal(t, 2, statistics.QualityStats.FailingAssets)
	assert.Equal(t, 2, statistics.LineageStats.TrackedAssets)
	assert.Equal(t, 25, statistics.LineageStats.LineagePaths)
	assert.Equal(t, 3, statistics.LineageStats.OrphanAssets)
	assert.Equal(t, 2, statistics.GovernanceStats.ManagedAssets)
	assert.Equal(t, 2, statistics.GovernanceStats.PolicyViolations)
	assert.Equal(t, 0.92, statistics.GovernanceStats.ComplianceScore)
	assert.Equal(t, 125.5, statistics.PerformanceStats.AvgResponseTime)
	assert.Equal(t, int64(45000), statistics.PerformanceStats.TotalQueries)
	assert.Len(t, statistics.Trends, 1)
}

func TestDataCatalogHandler_generateCatalogHealth(t *testing.T) {
	handler := NewDataCatalogHandler(zap.NewNop())

	request := DataCatalogRequest{
		Name:     "Test Catalog",
		Type:     CatalogTypeDatabase,
		Category: "test",
		Assets: []CatalogAsset{
			{ID: "asset_1", Name: "Asset 1", Type: AssetTypeDataset},
		},
	}

	health := handler.generateCatalogHealth(request)

	assert.Equal(t, "healthy", health.OverallStatus)
	assert.Len(t, health.ComponentHealth, 3)
	assert.Equal(t, "metadata", health.ComponentHealth[0].Component)
	assert.Equal(t, "healthy", health.ComponentHealth[0].Status)
	assert.Equal(t, 0.92, health.ComponentHealth[0].Score)
	assert.Equal(t, "discovery", health.ComponentHealth[1].Component)
	assert.Equal(t, "healthy", health.ComponentHealth[1].Status)
	assert.Equal(t, 0.88, health.ComponentHealth[1].Score)
	assert.Equal(t, "lineage", health.ComponentHealth[2].Component)
	assert.Equal(t, "warning", health.ComponentHealth[2].Status)
	assert.Equal(t, 0.75, health.ComponentHealth[2].Score)
	assert.Len(t, health.Issues, 1)
	assert.Equal(t, "issue_1", health.Issues[0].ID)
	assert.Equal(t, "metadata", health.Issues[0].Type)
	assert.Equal(t, "low", health.Issues[0].Severity)
	assert.Equal(t, "lineage", health.Issues[0].Component)
	assert.Len(t, health.Recommendations, 3)
	assert.NotZero(t, health.LastCheck)
	assert.NotZero(t, health.NextCheck)
}

func TestDataCatalogHandler_StringConversions(t *testing.T) {
	// Test CatalogType string conversion
	catalogType := CatalogTypeDatabase
	assert.Equal(t, "database", catalogType.String())

	// Test CatalogStatus string conversion
	catalogStatus := CatalogStatusActive
	assert.Equal(t, "active", catalogStatus.String())

	// Test AssetType string conversion
	assetType := AssetTypeDataset
	assert.Equal(t, "dataset", assetType.String())
}
