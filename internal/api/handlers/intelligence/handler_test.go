package intelligence

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
)

func TestNewDataIntelligencePlatformHandler(t *testing.T) {
	handler := NewDataIntelligencePlatformHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.jobs)
}

func TestDataIntelligencePlatformHandler_CreateIntelligenceAnalysis(t *testing.T) {
	handler := NewDataIntelligencePlatformHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful intelligence analysis creation",
			requestBody: IntelligenceAnalysisRequest{
				PlatformID: "platform-1",
				AnalysisID: "analysis-1",
				Type:       AnalysisTypeTrend,
				Parameters: map[string]interface{}{
					"data_source": "business_metrics",
					"time_range":  "3_months",
				},
				DataRange: DataRange{
					StartDate: time.Now().AddDate(0, -3, 0),
					EndDate:   time.Now(),
					TimeZone:  "UTC",
				},
				Options: AnalysisOptions{
					RealTime:      false,
					BatchMode:     true,
					Parallel:      false,
					Notifications: true,
					AuditTrail:    true,
					Monitoring:    true,
					Validation:    true,
				},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response IntelligenceAnalysisResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, "completed", response.Status)
				assert.NotNil(t, response.Analysis)
				assert.NotEmpty(t, response.Insights)
				assert.NotEmpty(t, response.Predictions)
				assert.NotEmpty(t, response.Recommendations)
				assert.NotNil(t, response.Statistics)
				assert.NotNil(t, response.Timeline)
			},
		},
		{
			name: "pattern analysis",
			requestBody: IntelligenceAnalysisRequest{
				PlatformID: "platform-1",
				AnalysisID: "analysis-2",
				Type:       AnalysisTypePattern,
				Parameters: map[string]interface{}{
					"data_source":  "customer_behavior",
					"pattern_type": "seasonal",
				},
				DataRange: DataRange{
					StartDate: time.Now().AddDate(0, -6, 0),
					EndDate:   time.Now(),
					TimeZone:  "UTC",
				},
				Options: AnalysisOptions{
					RealTime:      false,
					BatchMode:     true,
					Parallel:      false,
					Notifications: true,
					AuditTrail:    true,
					Monitoring:    true,
					Validation:    true,
				},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response IntelligenceAnalysisResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, AnalysisTypePattern, response.Analysis.Type)
				assert.NotEmpty(t, response.Insights)
			},
		},
		{
			name: "anomaly analysis",
			requestBody: IntelligenceAnalysisRequest{
				PlatformID: "platform-1",
				AnalysisID: "analysis-3",
				Type:       AnalysisTypeAnomaly,
				Parameters: map[string]interface{}{
					"data_source": "system_metrics",
					"threshold":   0.95,
				},
				DataRange: DataRange{
					StartDate: time.Now().AddDate(0, -1, 0),
					EndDate:   time.Now(),
					TimeZone:  "UTC",
				},
				Options: AnalysisOptions{
					RealTime:      true,
					BatchMode:     false,
					Parallel:      false,
					Notifications: true,
					AuditTrail:    true,
					Monitoring:    true,
					Validation:    true,
				},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response IntelligenceAnalysisResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, AnalysisTypeAnomaly, response.Analysis.Type)
				assert.NotEmpty(t, response.Insights)
			},
		},
		{
			name: "missing platform ID",
			requestBody: IntelligenceAnalysisRequest{
				AnalysisID: "analysis-1",
				Type:       AnalysisTypeTrend,
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Body.String(), "platform ID is required")
			},
		},
		{
			name: "missing analysis ID",
			requestBody: IntelligenceAnalysisRequest{
				PlatformID: "platform-1",
				Type:       AnalysisTypeTrend,
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Body.String(), "analysis ID is required")
			},
		},
		{
			name: "missing analysis type",
			requestBody: IntelligenceAnalysisRequest{
				PlatformID: "platform-1",
				AnalysisID: "analysis-1",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Body.String(), "analysis type is required")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/intelligence", bytes.NewBuffer(body))
			recorder := httptest.NewRecorder()

			handler.CreateIntelligenceAnalysis(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			tt.checkResponse(t, recorder)
		})
	}
}

func TestDataIntelligencePlatformHandler_GetIntelligenceAnalysis(t *testing.T) {
	handler := NewDataIntelligencePlatformHandler()

	tests := []struct {
		name           string
		analysisID     string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "successful analysis retrieval",
			analysisID:     "analysis-123",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response IntelligenceAnalysisResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "analysis-123", response.ID)
				assert.Equal(t, "retrieved", response.Status)
				assert.NotNil(t, response.Analysis)
				assert.NotEmpty(t, response.Insights)
				assert.NotEmpty(t, response.Predictions)
				assert.NotEmpty(t, response.Recommendations)
				assert.NotNil(t, response.Statistics)
				assert.NotNil(t, response.Timeline)
			},
		},
		{
			name:           "missing analysis ID",
			analysisID:     "",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Body.String(), "Analysis ID is required")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/intelligence"
			if tt.analysisID != "" {
				url += "?id=" + tt.analysisID
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			recorder := httptest.NewRecorder()

			handler.GetIntelligenceAnalysis(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			tt.checkResponse(t, recorder)
		})
	}
}

func TestDataIntelligencePlatformHandler_ListIntelligenceAnalyses(t *testing.T) {
	handler := NewDataIntelligencePlatformHandler()

	req := httptest.NewRequest(http.MethodGet, "/intelligence", nil)
	recorder := httptest.NewRecorder()

	handler.ListIntelligenceAnalyses(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "analyses")
	assert.Contains(t, response, "total")
	assert.Contains(t, response, "timestamp")

	analyses := response["analyses"].([]interface{})
	assert.Len(t, analyses, 3)
}

func TestDataIntelligencePlatformHandler_CreateIntelligenceJob(t *testing.T) {
	handler := NewDataIntelligencePlatformHandler()

	requestBody := IntelligenceAnalysisRequest{
		PlatformID: "platform-1",
		AnalysisID: "analysis-1",
		Type:       AnalysisTypePrediction,
		Parameters: map[string]interface{}{
			"data_source": "business_metrics",
			"horizon":     "30_days",
		},
		DataRange: DataRange{
			StartDate: time.Now().AddDate(0, -3, 0),
			EndDate:   time.Now(),
			TimeZone:  "UTC",
		},
		Options: AnalysisOptions{
			RealTime:      false,
			BatchMode:     true,
			Parallel:      false,
			Notifications: true,
			AuditTrail:    true,
			Monitoring:    true,
			Validation:    true,
		},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/intelligence/jobs", bytes.NewBuffer(body))
	recorder := httptest.NewRecorder()

	handler.CreateIntelligenceJob(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "job_id")
	assert.Contains(t, response, "status")
	assert.Contains(t, response, "created_at")
	assert.Equal(t, "created", response["status"])

	jobID := response["job_id"].(string)
	assert.NotEmpty(t, jobID)
}

func TestDataIntelligencePlatformHandler_GetIntelligenceJob(t *testing.T) {
	handler := NewDataIntelligencePlatformHandler()

	// Create a job first
	requestBody := IntelligenceAnalysisRequest{
		PlatformID: "platform-1",
		AnalysisID: "analysis-1",
		Type:       AnalysisTypeCorrelation,
	}
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/intelligence/jobs", bytes.NewBuffer(body))
	recorder := httptest.NewRecorder()
	handler.CreateIntelligenceJob(recorder, req)

	var createResponse map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &createResponse)
	jobID := createResponse["job_id"].(string)

	tests := []struct {
		name           string
		jobID          string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "successful job retrieval",
			jobID:          jobID,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var job IntelligenceJob
				err := json.Unmarshal(recorder.Body.Bytes(), &job)
				require.NoError(t, err)
				assert.Equal(t, jobID, job.ID)
				assert.Equal(t, "intelligence_analysis", job.Type)
			},
		},
		{
			name:           "missing job ID",
			jobID:          "",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Body.String(), "Job ID is required")
			},
		},
		{
			name:           "non-existent job ID",
			jobID:          "non-existent",
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Body.String(), "Job not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/intelligence/jobs"
			if tt.jobID != "" {
				url += "?id=" + tt.jobID
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			recorder := httptest.NewRecorder()

			handler.GetIntelligenceJob(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			tt.checkResponse(t, recorder)
		})
	}
}

func TestDataIntelligencePlatformHandler_ListIntelligenceJobs(t *testing.T) {
	handler := NewDataIntelligencePlatformHandler()

	// Create some jobs first
	for i := 0; i < 3; i++ {
		requestBody := IntelligenceAnalysisRequest{
			PlatformID: fmt.Sprintf("platform-%d", i),
			AnalysisID: fmt.Sprintf("analysis-%d", i),
			Type:       AnalysisTypeTrend,
		}
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/intelligence/jobs", bytes.NewBuffer(body))
		recorder := httptest.NewRecorder()
		handler.CreateIntelligenceJob(recorder, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/intelligence/jobs", nil)
	recorder := httptest.NewRecorder()

	handler.ListIntelligenceJobs(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "jobs")
	assert.Contains(t, response, "total")
	assert.Contains(t, response, "timestamp")

	jobs := response["jobs"].([]interface{})
	assert.Len(t, jobs, 3)
}

func TestDataIntelligencePlatformHandler_validateIntelligenceRequest(t *testing.T) {
	handler := NewDataIntelligencePlatformHandler()

	tests := []struct {
		name    string
		request IntelligenceAnalysisRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: IntelligenceAnalysisRequest{
				PlatformID: "platform-1",
				AnalysisID: "analysis-1",
				Type:       AnalysisTypeTrend,
			},
			wantErr: false,
		},
		{
			name: "missing platform ID",
			request: IntelligenceAnalysisRequest{
				AnalysisID: "analysis-1",
				Type:       AnalysisTypeTrend,
			},
			wantErr: true,
			errMsg:  "platform ID is required",
		},
		{
			name: "missing analysis ID",
			request: IntelligenceAnalysisRequest{
				PlatformID: "platform-1",
				Type:       AnalysisTypeTrend,
			},
			wantErr: true,
			errMsg:  "analysis ID is required",
		},
		{
			name: "missing analysis type",
			request: IntelligenceAnalysisRequest{
				PlatformID: "platform-1",
				AnalysisID: "analysis-1",
			},
			wantErr: true,
			errMsg:  "analysis type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateIntelligenceRequest(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataIntelligencePlatformHandler_processIntelligenceAnalysis(t *testing.T) {
	handler := NewDataIntelligencePlatformHandler()

	request := IntelligenceAnalysisRequest{
		PlatformID: "platform-1",
		AnalysisID: "analysis-1",
		Type:       AnalysisTypeClustering,
		Parameters: map[string]interface{}{
			"data_source":   "customer_data",
			"cluster_count": 4,
		},
		DataRange: DataRange{
			StartDate: time.Now().AddDate(0, -1, 0),
			EndDate:   time.Now(),
			TimeZone:  "UTC",
		},
		Options: AnalysisOptions{
			RealTime:      false,
			BatchMode:     true,
			Parallel:      false,
			Notifications: true,
			AuditTrail:    true,
			Monitoring:    true,
			Validation:    true,
		},
	}

	analysis := handler.processIntelligenceAnalysis(&request)

	assert.NotEmpty(t, analysis.ID)
	assert.Equal(t, "analysis-1", analysis.ID)
	assert.Equal(t, "clustering Analysis", analysis.Name)
	assert.Equal(t, AnalysisTypeClustering, analysis.Type)
	assert.Equal(t, IntelligenceStatusCompleted, analysis.Status)
	assert.NotZero(t, analysis.StartedAt)
	assert.NotZero(t, analysis.CompletedAt)
	assert.NotZero(t, analysis.Duration)
	assert.NotNil(t, analysis.Parameters)
	assert.NotNil(t, analysis.Results)
	assert.Empty(t, analysis.Errors)
	assert.NotNil(t, analysis.Metadata)
}

func TestDataIntelligencePlatformHandler_generateAnalysisResults(t *testing.T) {
	handler := NewDataIntelligencePlatformHandler()

	tests := []struct {
		name     string
		reqType  IntelligenceAnalysisType
		expected map[string]interface{}
	}{
		{
			name:    "trend analysis results",
			reqType: AnalysisTypeTrend,
			expected: map[string]interface{}{
				"trend_direction":  "upward",
				"trend_strength":   0.85,
				"trend_confidence": 0.92,
				"data_points":      1250,
			},
		},
		{
			name:    "pattern analysis results",
			reqType: AnalysisTypePattern,
			expected: map[string]interface{}{
				"pattern_type":       "seasonal",
				"pattern_strength":   0.78,
				"pattern_confidence": 0.88,
				"seasonality_period": "monthly",
			},
		},
		{
			name:    "anomaly analysis results",
			reqType: AnalysisTypeAnomaly,
			expected: map[string]interface{}{
				"anomaly_count":      3,
				"anomaly_severity":   "medium",
				"anomaly_confidence": 0.95,
				"affected_periods":   []string{"2024-01-15", "2024-02-03", "2024-02-18"},
			},
		},
		{
			name:    "prediction analysis results",
			reqType: AnalysisTypePrediction,
			expected: map[string]interface{}{
				"prediction_horizon":    "30 days",
				"prediction_confidence": 0.87,
				"predicted_value":       1250.5,
				"confidence_interval":   []float64{1150.2, 1350.8},
			},
		},
		{
			name:    "correlation analysis results",
			reqType: AnalysisTypeCorrelation,
			expected: map[string]interface{}{
				"correlation_coefficient":  0.76,
				"correlation_significance": 0.001,
				"correlated_variables":     []string{"revenue", "customer_count"},
			},
		},
		{
			name:    "clustering analysis results",
			reqType: AnalysisTypeClustering,
			expected: map[string]interface{}{
				"cluster_count":   4,
				"cluster_quality": 0.82,
				"cluster_sizes":   []int{150, 320, 180, 95},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := IntelligenceAnalysisRequest{
				PlatformID: "platform-1",
				AnalysisID: "analysis-1",
				Type:       tt.reqType,
			}

			results := handler.generateAnalysisResults(&req)

			for key, expectedValue := range tt.expected {
				assert.Contains(t, results, key)
				assert.Equal(t, expectedValue, results[key])
			}
		})
	}
}

func TestDataIntelligencePlatformHandler_generateInsights(t *testing.T) {
	handler := NewDataIntelligencePlatformHandler()

	analysis := &IntelligenceAnalysis{
		ID:   "test-analysis",
		Type: AnalysisTypeTrend,
	}

	insights := handler.generateInsights(analysis)

	assert.Len(t, insights, 3)

	// Check first insight
	assert.Equal(t, "insight-1", insights[0].ID)
	assert.Equal(t, "Strong Upward Trend Detected", insights[0].Title)
	assert.Equal(t, "trend", insights[0].Type)
	assert.Equal(t, "performance", insights[0].Category)
	assert.Equal(t, 0.92, insights[0].Confidence)
	assert.Equal(t, "high", insights[0].Impact)
	assert.NotNil(t, insights[0].Data)
	assert.NotZero(t, insights[0].CreatedAt)

	// Check second insight
	assert.Equal(t, "insight-2", insights[1].ID)
	assert.Equal(t, "Seasonal Pattern Identified", insights[1].Title)
	assert.Equal(t, "pattern", insights[1].Type)
	assert.Equal(t, "behavior", insights[1].Category)
	assert.Equal(t, 0.88, insights[1].Confidence)
	assert.Equal(t, "medium", insights[1].Impact)

	// Check third insight
	assert.Equal(t, "insight-3", insights[2].ID)
	assert.Equal(t, "Anomaly Detection Alert", insights[2].Title)
	assert.Equal(t, "anomaly", insights[2].Type)
	assert.Equal(t, "alert", insights[2].Category)
	assert.Equal(t, 0.95, insights[2].Confidence)
	assert.Equal(t, "high", insights[2].Impact)
}

func TestDataIntelligencePlatformHandler_generatePredictions(t *testing.T) {
	handler := NewDataIntelligencePlatformHandler()

	analysis := &IntelligenceAnalysis{
		ID:   "test-analysis",
		Type: AnalysisTypePrediction,
	}

	predictions := handler.generatePredictions(analysis)

	assert.Len(t, predictions, 3)

	// Check first prediction
	assert.Equal(t, "prediction-1", predictions[0].ID)
	assert.Equal(t, "Revenue Forecast", predictions[0].Title)
	assert.Equal(t, "revenue", predictions[0].Type)
	assert.Equal(t, 1250.5, predictions[0].Value)
	assert.Equal(t, 0.87, predictions[0].Confidence)
	assert.Equal(t, time.Hour*24*30, predictions[0].Horizon)
	assert.Len(t, predictions[0].Factors, 3)
	assert.NotZero(t, predictions[0].CreatedAt)

	// Check second prediction
	assert.Equal(t, "prediction-2", predictions[1].ID)
	assert.Equal(t, "Customer Growth", predictions[1].Title)
	assert.Equal(t, "customers", predictions[1].Type)
	assert.Equal(t, 150, predictions[1].Value)
	assert.Equal(t, 0.82, predictions[1].Confidence)

	// Check third prediction
	assert.Equal(t, "prediction-3", predictions[2].ID)
	assert.Equal(t, "Risk Assessment", predictions[2].Title)
	assert.Equal(t, "risk", predictions[2].Type)
	assert.Equal(t, "low", predictions[2].Value)
	assert.Equal(t, 0.91, predictions[2].Confidence)
}

func TestDataIntelligencePlatformHandler_generateRecommendations(t *testing.T) {
	handler := NewDataIntelligencePlatformHandler()

	analysis := &IntelligenceAnalysis{
		ID:   "test-analysis",
		Type: AnalysisTypeTrend,
	}

	recommendations := handler.generateRecommendations(analysis)

	assert.Len(t, recommendations, 3)

	// Check first recommendation
	assert.Equal(t, "rec-1", recommendations[0].ID)
	assert.Equal(t, "Optimize Marketing Strategy", recommendations[0].Title)
	assert.Equal(t, "strategy", recommendations[0].Type)
	assert.Equal(t, "high", recommendations[0].Priority)
	assert.Equal(t, "high", recommendations[0].Impact)
	assert.Equal(t, "medium", recommendations[0].Effort)
	assert.Len(t, recommendations[0].Actions, 3)
	assert.NotZero(t, recommendations[0].CreatedAt)

	// Check second recommendation
	assert.Equal(t, "rec-2", recommendations[1].ID)
	assert.Equal(t, "Investigate Anomalies", recommendations[1].Title)
	assert.Equal(t, "investigation", recommendations[1].Type)
	assert.Equal(t, "high", recommendations[1].Priority)
	assert.Equal(t, "medium", recommendations[1].Impact)
	assert.Equal(t, "high", recommendations[1].Effort)

	// Check third recommendation
	assert.Equal(t, "rec-3", recommendations[2].ID)
	assert.Equal(t, "Enhance Monitoring", recommendations[2].Title)
	assert.Equal(t, "monitoring", recommendations[2].Type)
	assert.Equal(t, "medium", recommendations[2].Priority)
	assert.Equal(t, "medium", recommendations[2].Impact)
	assert.Equal(t, "low", recommendations[2].Effort)
}

func TestDataIntelligencePlatformHandler_generateIntelligenceStatistics(t *testing.T) {
	handler := NewDataIntelligencePlatformHandler()

	analysis := &IntelligenceAnalysis{
		ID:        "test-analysis",
		Type:      AnalysisTypeTrend,
		StartedAt: time.Now(),
		Duration:  time.Minute * 5,
	}

	statistics := handler.generateIntelligenceStatistics(analysis)

	assert.Equal(t, 15, statistics.TotalAnalyses)
	assert.Equal(t, 12, statistics.CompletedAnalyses)
	assert.Equal(t, 2, statistics.FailedAnalyses)
	assert.Equal(t, 1, statistics.ActiveAnalyses)
	assert.Equal(t, 45, statistics.TotalInsights)
	assert.Equal(t, 28, statistics.TotalPredictions)
	assert.Equal(t, 32, statistics.TotalRecommendations)

	assert.NotNil(t, statistics.PerformanceMetrics)
	assert.Equal(t, 2.5, statistics.PerformanceMetrics["avg_processing_time"])
	assert.Equal(t, 0.93, statistics.PerformanceMetrics["success_rate"])
	assert.Equal(t, 0.89, statistics.PerformanceMetrics["accuracy"])

	assert.NotNil(t, statistics.AccuracyMetrics)
	assert.Equal(t, 0.87, statistics.AccuracyMetrics["prediction_accuracy"])
	assert.Equal(t, 0.92, statistics.AccuracyMetrics["insight_relevance"])
	assert.Equal(t, 0.85, statistics.AccuracyMetrics["recommendation_quality"])

	assert.Len(t, statistics.TimelineEvents, 1)
}

func TestDataIntelligencePlatformHandler_generateIntelligenceTimeline(t *testing.T) {
	handler := NewDataIntelligencePlatformHandler()

	analysis := &IntelligenceAnalysis{
		ID:          "test-analysis",
		Type:        AnalysisTypeTrend,
		StartedAt:   time.Now(),
		CompletedAt: time.Now().Add(time.Minute * 5),
		Duration:    time.Minute * 5,
	}

	timeline := handler.generateIntelligenceTimeline(analysis)

	assert.NotZero(t, timeline.StartDate)
	assert.NotZero(t, timeline.EndDate)
	assert.Equal(t, float64(300000), timeline.Duration) // 5 minutes in milliseconds

	assert.Len(t, timeline.Milestones, 3)
	assert.Equal(t, "milestone-1", timeline.Milestones[0].ID)
	assert.Equal(t, "Analysis Started", timeline.Milestones[0].Name)
	assert.Equal(t, "completed", timeline.Milestones[0].Status)
	assert.Equal(t, "start", timeline.Milestones[0].Type)

	assert.Len(t, timeline.Events, 1)
	assert.Equal(t, "event-1", timeline.Events[0].ID)
	assert.Equal(t, "analysis_started", timeline.Events[0].Type)
	assert.Equal(t, "test-analysis", timeline.Events[0].Analysis)

	assert.Len(t, timeline.Projections, 1)
	assert.Equal(t, "performance", timeline.Projections[0].Type)
	assert.Equal(t, 0.85, timeline.Projections[0].Confidence)
}

func TestDataIntelligencePlatformHandler_generateSampleAnalysis(t *testing.T) {
	handler := NewDataIntelligencePlatformHandler()

	analysis := handler.generateSampleAnalysis("test-id")

	assert.Equal(t, "test-id", analysis.ID)
	assert.Equal(t, "Sample Intelligence Analysis", analysis.Name)
	assert.Equal(t, AnalysisTypeTrend, analysis.Type)
	assert.Equal(t, "Sample intelligence analysis for demonstration", analysis.Description)
	assert.Equal(t, IntelligenceStatusCompleted, analysis.Status)
	assert.NotZero(t, analysis.StartedAt)
	assert.NotZero(t, analysis.CompletedAt)
	assert.Equal(t, time.Minute*5, analysis.Duration)
	assert.NotNil(t, analysis.Parameters)
	assert.NotNil(t, analysis.Results)
	assert.Empty(t, analysis.Errors)
	assert.NotNil(t, analysis.Metadata)
}

func TestDataIntelligencePlatformHandler_processIntelligenceJob(t *testing.T) {
	handler := NewDataIntelligencePlatformHandler()

	request := IntelligenceAnalysisRequest{
		PlatformID: "platform-1",
		AnalysisID: "analysis-1",
		Type:       AnalysisTypeTrend,
	}

	jobID := "test-job-123"
	job := &IntelligenceJob{
		ID:        jobID,
		Type:      "intelligence_analysis",
		Status:    "pending",
		Progress:  0.0,
		CreatedAt: time.Now(),
	}

	handler.mu.Lock()
	handler.jobs[jobID] = job
	handler.mu.Unlock()

	// Start processing
	go handler.processIntelligenceJob(jobID, &request)

	// Wait for processing to complete
	time.Sleep(1200 * time.Millisecond)

	handler.mu.RLock()
	processedJob := handler.jobs[jobID]
	handler.mu.RUnlock()

	assert.Equal(t, "completed", processedJob.Status)
	assert.Equal(t, 1.0, processedJob.Progress)
	assert.NotNil(t, processedJob.Result)
	assert.NotZero(t, processedJob.CompletedAt)
}

// String conversion tests for enums
func TestIntelligenceAnalysisType_String(t *testing.T) {
	tests := []struct {
		analysisType IntelligenceAnalysisType
		expected     string
	}{
		{AnalysisTypeTrend, "trend"},
		{AnalysisTypePattern, "pattern"},
		{AnalysisTypeAnomaly, "anomaly"},
		{AnalysisTypePrediction, "prediction"},
		{AnalysisTypeCorrelation, "correlation"},
		{AnalysisTypeClustering, "clustering"},
	}

	for _, tt := range tests {
		t.Run(string(tt.analysisType), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.analysisType))
		})
	}
}

func TestIntelligenceStatus_String(t *testing.T) {
	tests := []struct {
		status   IntelligenceStatus
		expected string
	}{
		{IntelligenceStatusPending, "pending"},
		{IntelligenceStatusRunning, "running"},
		{IntelligenceStatusCompleted, "completed"},
		{IntelligenceStatusFailed, "failed"},
		{IntelligenceStatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

func TestDataSourceType_String(t *testing.T) {
	tests := []struct {
		sourceType DataSourceType
		expected   string
	}{
		{DataSourceInternal, "internal"},
		{DataSourceExternal, "external"},
		{DataSourceAPI, "api"},
		{DataSourceDatabase, "database"},
		{DataSourceFile, "file"},
		{DataSourceStream, "stream"},
	}

	for _, tt := range tests {
		t.Run(string(tt.sourceType), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.sourceType))
		})
	}
}

func TestIntelligenceModelType_String(t *testing.T) {
	tests := []struct {
		modelType IntelligenceModelType
		expected  string
	}{
		{ModelTypeML, "machine_learning"},
		{ModelTypeStatistical, "statistical"},
		{ModelTypeRuleBased, "rule_based"},
		{ModelTypeHybrid, "hybrid"},
		{ModelTypeCustom, "custom"},
	}

	for _, tt := range tests {
		t.Run(string(tt.modelType), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.modelType))
		})
	}
}
