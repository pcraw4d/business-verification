package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewDataDiscoveryHandler(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataDiscoveryHandler(logger)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.logger)
	assert.NotNil(t, handler.discoveries)
	assert.NotNil(t, handler.jobs)
}

func TestDataDiscoveryHandler_CreateDiscovery(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataDiscoveryHandler(logger)

	tests := []struct {
		name           string
		request        DataDiscoveryRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid discovery request",
			request: DataDiscoveryRequest{
				Name:        "Test Discovery",
				Description: "Test discovery description",
				Type:        DiscoveryTypeAuto,
				Sources: []DiscoverySource{
					{
						ID:       "source_1",
						Name:     "Test Source",
						Type:     "database",
						Location: "postgres://localhost:5432/test",
						Enabled:  true,
					},
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			request: DataDiscoveryRequest{
				Type: DiscoveryTypeAuto,
				Sources: []DiscoverySource{
					{
						ID:       "source_1",
						Name:     "Test Source",
						Type:     "database",
						Location: "postgres://localhost:5432/test",
						Enabled:  true,
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name is required",
		},
		{
			name: "missing type",
			request: DataDiscoveryRequest{
				Name: "Test Discovery",
				Sources: []DiscoverySource{
					{
						ID:       "source_1",
						Name:     "Test Source",
						Type:     "database",
						Location: "postgres://localhost:5432/test",
						Enabled:  true,
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "type is required",
		},
		{
			name: "missing sources",
			request: DataDiscoveryRequest{
				Name: "Test Discovery",
				Type: DiscoveryTypeAuto,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "at least one source is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/discovery", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.CreateDiscovery(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response DataDiscoveryResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, tt.request.Name, response.Name)
				assert.Equal(t, tt.request.Type, response.Type)
				assert.Equal(t, DiscoveryStatusCompleted, response.Status)
			}
		})
	}
}

func TestDataDiscoveryHandler_GetDiscovery(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataDiscoveryHandler(logger)

	// Create a test discovery
	discovery := &DataDiscoveryResponse{
		ID:     "test_discovery_1",
		Name:   "Test Discovery",
		Type:   DiscoveryTypeAuto,
		Status: DiscoveryStatusCompleted,
	}
	handler.discoveries["test_discovery_1"] = discovery

	tests := []struct {
		name           string
		id             string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "existing discovery",
			id:             "test_discovery_1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing id",
			id:             "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Discovery ID is required",
		},
		{
			name:           "non-existent discovery",
			id:             "non_existent",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Discovery not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/discovery"
			if tt.id != "" {
				url += "?id=" + tt.id
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.GetDiscovery(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response DataDiscoveryResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, discovery.ID, response.ID)
				assert.Equal(t, discovery.Name, response.Name)
			}
		})
	}
}

func TestDataDiscoveryHandler_ListDiscoveries(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataDiscoveryHandler(logger)

	// Add test discoveries
	discovery1 := &DataDiscoveryResponse{ID: "discovery_1", Name: "Discovery 1"}
	discovery2 := &DataDiscoveryResponse{ID: "discovery_2", Name: "Discovery 2"}
	handler.discoveries["discovery_1"] = discovery1
	handler.discoveries["discovery_2"] = discovery2

	req := httptest.NewRequest("GET", "/discovery", nil)
	w := httptest.NewRecorder()

	handler.ListDiscoveries(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(2), response["total"])
	discoveries := response["discoveries"].([]interface{})
	assert.Len(t, discoveries, 2)
}

func TestDataDiscoveryHandler_CreateDiscoveryJob(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataDiscoveryHandler(logger)

	request := DataDiscoveryRequest{
		Name:        "Test Discovery Job",
		Description: "Test discovery job description",
		Type:        DiscoveryTypeScheduled,
		Sources: []DiscoverySource{
			{
				ID:       "source_1",
				Name:     "Test Source",
				Type:     "database",
				Location: "postgres://localhost:5432/test",
				Enabled:  true,
			},
		},
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/discovery/jobs", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CreateDiscoveryJob(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var job DiscoveryJob
	err := json.Unmarshal(w.Body.Bytes(), &job)
	require.NoError(t, err)

	assert.NotEmpty(t, job.ID)
	assert.Equal(t, "pending", job.Status)
	assert.Equal(t, 0, job.Progress)
}

func TestDataDiscoveryHandler_GetDiscoveryJob(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataDiscoveryHandler(logger)

	// Create a test job
	job := &DiscoveryJob{
		ID:        "test_job_1",
		RequestID: "req_1",
		Type:      "discovery_creation",
		Status:    "completed",
		Progress:  100,
	}
	handler.jobs["test_job_1"] = job

	tests := []struct {
		name           string
		id             string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "existing job",
			id:             "test_job_1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing id",
			id:             "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Job ID is required",
		},
		{
			name:           "non-existent job",
			id:             "non_existent",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Job not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/discovery/jobs"
			if tt.id != "" {
				url += "?id=" + tt.id
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.GetDiscoveryJob(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response DiscoveryJob
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, job.ID, response.ID)
				assert.Equal(t, job.Status, response.Status)
			}
		})
	}
}

func TestDataDiscoveryHandler_ListDiscoveryJobs(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataDiscoveryHandler(logger)

	// Add test jobs
	job1 := &DiscoveryJob{ID: "job_1", Status: "completed"}
	job2 := &DiscoveryJob{ID: "job_2", Status: "pending"}
	handler.jobs["job_1"] = job1
	handler.jobs["job_2"] = job2

	req := httptest.NewRequest("GET", "/discovery/jobs", nil)
	w := httptest.NewRecorder()

	handler.ListDiscoveryJobs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(2), response["total"])
	jobs := response["jobs"].([]interface{})
	assert.Len(t, jobs, 2)
}

func TestDataDiscoveryHandler_validateDiscoveryRequest(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataDiscoveryHandler(logger)

	tests := []struct {
		name    string
		request DataDiscoveryRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: DataDiscoveryRequest{
				Name: "Test Discovery",
				Type: DiscoveryTypeAuto,
				Sources: []DiscoverySource{
					{ID: "source_1", Name: "Test Source", Type: "database"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			request: DataDiscoveryRequest{
				Type: DiscoveryTypeAuto,
				Sources: []DiscoverySource{
					{ID: "source_1", Name: "Test Source", Type: "database"},
				},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing type",
			request: DataDiscoveryRequest{
				Name: "Test Discovery",
				Sources: []DiscoverySource{
					{ID: "source_1", Name: "Test Source", Type: "database"},
				},
			},
			wantErr: true,
			errMsg:  "type is required",
		},
		{
			name: "missing sources",
			request: DataDiscoveryRequest{
				Name: "Test Discovery",
				Type: DiscoveryTypeAuto,
			},
			wantErr: true,
			errMsg:  "at least one source is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateDiscoveryRequest(tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataDiscoveryHandler_processSources(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataDiscoveryHandler(logger)

	sources := []DiscoverySource{
		{
			ID:       "source_1",
			Name:     "Test Source 1",
			Type:     "database",
			Location: "postgres://localhost:5432/test1",
			Enabled:  true,
		},
		{
			ID:       "source_2",
			Name:     "Test Source 2",
			Type:     "file",
			Location: "/path/to/file",
			Enabled:  false,
		},
	}

	processed := handler.processSources(sources)

	assert.Len(t, processed, 2)
	assert.Equal(t, sources[0].ID, processed[0].ID)
	assert.Equal(t, sources[1].ID, processed[1].ID)
}

func TestDataDiscoveryHandler_processRules(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataDiscoveryHandler(logger)

	rules := []DiscoveryRule{
		{
			ID:          "rule_1",
			Name:        "Test Rule 1",
			Type:        "validation",
			Description: "Test rule description",
			Enabled:     true,
		},
	}

	processed := handler.processRules(rules)

	assert.Len(t, processed, 1)
	assert.Equal(t, rules[0].ID, processed[0].ID)
}

func TestDataDiscoveryHandler_processProfiles(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataDiscoveryHandler(logger)

	profiles := []DiscoveryProfile{
		{
			ID:          "profile_1",
			Name:        "Test Profile",
			Type:        ProfileTypeStatistical,
			Description: "Test profile description",
			Enabled:     true,
		},
	}

	processed := handler.processProfiles(profiles)

	assert.Len(t, processed, 1)
	assert.Equal(t, profiles[0].ID, processed[0].ID)
}

func TestDataDiscoveryHandler_processPatterns(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataDiscoveryHandler(logger)

	patterns := []DiscoveryPattern{
		{
			ID:          "pattern_1",
			Name:        "Test Pattern",
			Type:        PatternTypeTemporal,
			Description: "Test pattern description",
			Enabled:     true,
		},
	}

	processed := handler.processPatterns(patterns)

	assert.Len(t, processed, 1)
	assert.Equal(t, patterns[0].ID, processed[0].ID)
}

func TestDataDiscoveryHandler_generateDiscoveryResults(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataDiscoveryHandler(logger)

	request := DataDiscoveryRequest{
		Name: "Test Discovery",
		Type: DiscoveryTypeAuto,
		Sources: []DiscoverySource{
			{ID: "source_1", Name: "Test Source", Type: "database"},
		},
	}

	results := handler.generateDiscoveryResults(request)

	assert.NotNil(t, results)
	assert.Len(t, results.Assets, 1)
	assert.Len(t, results.Profiles, 1)
	assert.Len(t, results.Patterns, 1)
	assert.Len(t, results.Anomalies, 1)
	assert.Len(t, results.Recommendations, 1)

	// Check asset details
	asset := results.Assets[0]
	assert.Equal(t, "asset_1", asset.ID)
	assert.Equal(t, "Customer Database", asset.Name)
	assert.Equal(t, "database", asset.Type)
	assert.Equal(t, "postgres://localhost:5432/customers", asset.Location)
	assert.Equal(t, int64(1024000000), asset.Size)
	assert.Equal(t, "postgresql", asset.Format)

	// Check schema
	assert.Equal(t, "relational", asset.Schema.Type)
	assert.Len(t, asset.Schema.Columns, 2)

	// Check quality
	assert.Equal(t, 0.85, asset.Quality.Score)
	assert.Equal(t, 0.90, asset.Quality.Completeness)

	// Check patterns
	assert.Len(t, asset.Patterns, 1)
	assert.Equal(t, PatternTypeTemporal, asset.Patterns[0].Type)

	// Check anomalies
	assert.Len(t, asset.Anomalies, 1)
	assert.Equal(t, "outlier", asset.Anomalies[0].Type)
}

func TestDataDiscoveryHandler_generateDiscoverySummary(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataDiscoveryHandler(logger)

	request := DataDiscoveryRequest{
		Name: "Test Discovery",
		Type: DiscoveryTypeAuto,
		Sources: []DiscoverySource{
			{ID: "source_1", Name: "Test Source", Type: "database"},
		},
	}

	summary := handler.generateDiscoverySummary(request)

	assert.Equal(t, 1, summary.TotalAssets)
	assert.Equal(t, 1, summary.TotalProfiles)
	assert.Equal(t, 1, summary.TotalPatterns)
	assert.Equal(t, 1, summary.TotalAnomalies)
	assert.Equal(t, "1GB", summary.DataVolume)
	assert.Equal(t, 0.85, summary.Coverage)
	assert.Equal(t, 0.90, summary.Completeness)

	// Check asset types
	assert.Equal(t, 1, summary.AssetTypes["database"])

	// Check quality scores
	assert.Equal(t, 0.85, summary.QualityScores["overall"])

	// Check pattern types
	assert.Equal(t, 1, summary.PatternTypes["temporal"])

	// Check anomaly types
	assert.Equal(t, 1, summary.AnomalyTypes["outlier"])
}

func TestDataDiscoveryHandler_generateDiscoveryStatistics(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataDiscoveryHandler(logger)

	request := DataDiscoveryRequest{
		Name: "Test Discovery",
		Type: DiscoveryTypeAuto,
		Sources: []DiscoverySource{
			{ID: "source_1", Name: "Test Source", Type: "database"},
		},
	}

	statistics := handler.generateDiscoveryStatistics(request)

	// Check performance stats
	assert.Equal(t, 120.5, statistics.PerformanceStats.TotalTime)
	assert.Equal(t, 120.5, statistics.PerformanceStats.AvgTimePerAsset)
	assert.Equal(t, 1.0, statistics.PerformanceStats.SuccessRate)
	assert.Equal(t, 0.0, statistics.PerformanceStats.ErrorRate)

	// Check quality stats
	assert.Equal(t, 0.85, statistics.QualityStats.OverallScore)
	assert.Equal(t, 0, statistics.QualityStats.HighQuality)
	assert.Equal(t, 1, statistics.QualityStats.MediumQuality)
	assert.Equal(t, 0, statistics.QualityStats.LowQuality)
	assert.Equal(t, "stable", statistics.QualityStats.TrendDirection)

	// Check pattern stats
	assert.Equal(t, 1, statistics.PatternStats.TotalPatterns)
	assert.Equal(t, 0.85, statistics.PatternStats.AvgConfidence)
	assert.Equal(t, 1, statistics.PatternStats.HighConfidence)

	// Check anomaly stats
	assert.Equal(t, 1, statistics.AnomalyStats.TotalAnomalies)
	assert.Equal(t, 0.95, statistics.AnomalyStats.AvgScore)
	assert.Equal(t, 1, statistics.AnomalyStats.HighSeverity)

	// Check trends
	assert.Len(t, statistics.Trends, 1)
	assert.Equal(t, "assets_discovered", statistics.Trends[0].Metric)
	assert.Equal(t, "daily", statistics.Trends[0].Period)
	assert.Equal(t, "increasing", statistics.Trends[0].Direction)
}

func TestDataDiscoveryHandler_generateDiscoveryInsights(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataDiscoveryHandler(logger)

	request := DataDiscoveryRequest{
		Name: "Test Discovery",
		Type: DiscoveryTypeAuto,
		Sources: []DiscoverySource{
			{ID: "source_1", Name: "Test Source", Type: "database"},
		},
	}

	insights := handler.generateDiscoveryInsights(request)

	assert.Len(t, insights, 2)

	// Check first insight
	insight1 := insights[0]
	assert.Equal(t, "insight_1", insight1.ID)
	assert.Equal(t, "quality", insight1.Type)
	assert.Equal(t, "Data Quality Issues Detected", insight1.Title)
	assert.Equal(t, "medium", insight1.Severity)
	assert.Equal(t, 0.85, insight1.Confidence)

	// Check second insight
	insight2 := insights[1]
	assert.Equal(t, "insight_2", insight2.ID)
	assert.Equal(t, "pattern", insight2.Type)
	assert.Equal(t, "Temporal Patterns Identified", insight2.Title)
	assert.Equal(t, "low", insight2.Severity)
	assert.Equal(t, 0.90, insight2.Confidence)
}

func TestDataDiscoveryHandler_processDiscoveryJob(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataDiscoveryHandler(logger)

	request := DataDiscoveryRequest{
		Name: "Test Discovery Job",
		Type: DiscoveryTypeScheduled,
		Sources: []DiscoverySource{
			{ID: "source_1", Name: "Test Source", Type: "database"},
		},
	}

	job := &DiscoveryJob{
		ID:        "test_job_1",
		RequestID: "req_1",
		Type:      "discovery_creation",
		Status:    "pending",
		Progress:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Start processing in background
	go handler.processDiscoveryJob(job, request)

	// Wait a bit for processing to start
	time.Sleep(100 * time.Millisecond)

	// Check initial state
	assert.Equal(t, "running", job.Status)
	assert.Greater(t, job.Progress, 0)

	// Wait for completion
	time.Sleep(5 * time.Second)

	// Check final state
	assert.Equal(t, "completed", job.Status)
	assert.Equal(t, 100, job.Progress)
	assert.NotNil(t, job.Result)
	assert.NotNil(t, job.CompletedAt)
}

// String conversion tests
func TestDiscoveryType_String(t *testing.T) {
	tests := []struct {
		discoveryType DiscoveryType
		expected      string
	}{
		{DiscoveryTypeAuto, "auto"},
		{DiscoveryTypeManual, "manual"},
		{DiscoveryTypeScheduled, "scheduled"},
		{DiscoveryTypeIncremental, "incremental"},
		{DiscoveryTypeFull, "full"},
	}

	for _, tt := range tests {
		t.Run(string(tt.discoveryType), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.discoveryType.String())
		})
	}
}

func TestDiscoveryStatus_String(t *testing.T) {
	tests := []struct {
		status   DiscoveryStatus
		expected string
	}{
		{DiscoveryStatusPending, "pending"},
		{DiscoveryStatusRunning, "running"},
		{DiscoveryStatusCompleted, "completed"},
		{DiscoveryStatusFailed, "failed"},
		{DiscoveryStatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestProfileType_String(t *testing.T) {
	tests := []struct {
		profileType ProfileType
		expected    string
	}{
		{ProfileTypeStatistical, "statistical"},
		{ProfileTypeQuality, "quality"},
		{ProfileTypePattern, "pattern"},
		{ProfileTypeAnomaly, "anomaly"},
		{ProfileTypeComprehensive, "comprehensive"},
	}

	for _, tt := range tests {
		t.Run(string(tt.profileType), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.profileType.String())
		})
	}
}

func TestPatternType_String(t *testing.T) {
	tests := []struct {
		patternType PatternType
		expected    string
	}{
		{PatternTypeTemporal, "temporal"},
		{PatternTypeSequential, "sequential"},
		{PatternTypeCorrelation, "correlation"},
		{PatternTypeOutlier, "outlier"},
		{PatternTypeTrend, "trend"},
		{PatternTypeSeasonal, "seasonal"},
		{PatternTypeCyclic, "cyclic"},
		{PatternTypeCustom, "custom"},
	}

	for _, tt := range tests {
		t.Run(string(tt.patternType), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.patternType.String())
		})
	}
}
