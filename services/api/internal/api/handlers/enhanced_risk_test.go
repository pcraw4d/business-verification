package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kyb-platform/internal/risk"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockEnhancedRiskService is a mock implementation of the enhanced risk service
type MockEnhancedRiskService struct {
	mock.Mock
}

func (m *MockEnhancedRiskService) PerformEnhancedRiskAssessment(ctx context.Context, request *risk.EnhancedRiskAssessmentRequest) (*risk.EnhancedRiskAssessmentResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*risk.EnhancedRiskAssessmentResponse), args.Error(1)
}

func (m *MockEnhancedRiskService) GetRiskFactorHistory(ctx context.Context, businessID string, factorType string, timeRange *risk.TimeRange) ([]risk.RiskHistoryEntry, error) {
	args := m.Called(ctx, businessID, factorType, timeRange)
	return args.Get(0).([]risk.RiskHistoryEntry), args.Error(1)
}

func (m *MockEnhancedRiskService) GetActiveAlerts(ctx context.Context, businessID string) ([]risk.AlertDetail, error) {
	args := m.Called(ctx, businessID)
	return args.Get(0).([]risk.AlertDetail), args.Error(1)
}

func (m *MockEnhancedRiskService) AcknowledgeAlert(ctx context.Context, alertID string, userID string, notes string) error {
	args := m.Called(ctx, alertID, userID, notes)
	return args.Error(0)
}

func (m *MockEnhancedRiskService) ResolveAlert(ctx context.Context, alertID string, userID string, resolutionNotes string) error {
	args := m.Called(ctx, alertID, userID, resolutionNotes)
	return args.Error(0)
}

func TestEnhancedRiskHandler_EnhancedRiskAssessmentHandler(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockEnhancedRiskService)
	handler := NewEnhancedRiskHandler(logger, mockService)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful risk assessment",
			requestBody: risk.EnhancedRiskAssessmentRequest{
				AssessmentID: "test-assessment-123",
				BusinessID:   "test-business-456",
				RiskFactorInputs: []risk.RiskFactorInput{
					{
						FactorType: "financial",
						Data: map[string]interface{}{
							"revenue": 1000000,
							"debt":    500000,
						},
						Weight: 0.3,
					},
				},
				IncludeTrendAnalysis:       true,
				IncludeCorrelationAnalysis: true,
			},
			mockSetup: func() {
				mockService.On("PerformEnhancedRiskAssessment", mock.Anything, mock.AnythingOfType("*risk.EnhancedRiskAssessmentRequest")).
					Return(&risk.EnhancedRiskAssessmentResponse{
						AssessmentID:     "test-assessment-123",
						BusinessID:       "test-business-456",
						Timestamp:        time.Now(),
						OverallRiskScore: 0.65,
						OverallRiskLevel: risk.RiskLevelHigh,
						RiskFactors: []risk.RiskFactorDetail{
							{
								FactorType:  "financial",
								Score:       0.65,
								RiskLevel:   risk.RiskLevelHigh,
								Confidence:  0.85,
								Weight:      0.3,
								Description: "Financial risk assessment",
								LastUpdated: time.Now(),
							},
						},
						Recommendations:  []risk.RecommendationDetail{},
						Alerts:           []risk.AlertDetail{},
						ConfidenceScore:  0.85,
						ProcessingTimeMs: 150,
						Metadata: map[string]interface{}{
							"version": "2.0",
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"invalid_field": "invalid_value",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "service error",
			requestBody: risk.EnhancedRiskAssessmentRequest{
				AssessmentID: "test-assessment-123",
				BusinessID:   "test-business-456",
				RiskFactorInputs: []risk.RiskFactorInput{
					{
						FactorType: "financial",
						Data:       map[string]interface{}{},
						Weight:     0.3,
					},
				},
			},
			mockSetup: func() {
				mockService.On("PerformEnhancedRiskAssessment", mock.Anything, mock.AnythingOfType("*risk.EnhancedRiskAssessmentRequest")).
					Return((*risk.EnhancedRiskAssessmentResponse)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			tt.mockSetup()

			// Create request
			requestBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/v1/risk/enhanced/assess", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.EnhancedRiskAssessmentHandler(rr, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var response risk.EnhancedRiskAssessmentResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.AssessmentID)
				assert.NotEmpty(t, response.BusinessID)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestEnhancedRiskHandler_RiskFactorCalculationHandler(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockEnhancedRiskService)
	handler := NewEnhancedRiskHandler(logger, mockService)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful factor calculation",
			requestBody: map[string]interface{}{
				"factor_type": "financial",
				"data": map[string]interface{}{
					"revenue": 1000000,
					"debt":    500000,
				},
				"weight": 0.3,
			},
			mockSetup: func() {
				// Mock the risk factor calculation
				mockService.On("PerformEnhancedRiskAssessment", mock.Anything, mock.AnythingOfType("*risk.EnhancedRiskAssessmentRequest")).
					Return(&risk.EnhancedRiskAssessmentResponse{
						AssessmentID:     "calc-test-123",
						BusinessID:       "test-business-456",
						Timestamp:        time.Now(),
						OverallRiskScore: 0.65,
						OverallRiskLevel: risk.RiskLevelHigh,
						RiskFactors: []risk.RiskFactorDetail{
							{
								FactorType:  "financial",
								Score:       0.65,
								RiskLevel:   risk.RiskLevelHigh,
								Confidence:  0.85,
								Weight:      0.3,
								Description: "Financial risk calculation",
								LastUpdated: time.Now(),
							},
						},
						Recommendations:  []risk.RecommendationDetail{},
						Alerts:           []risk.AlertDetail{},
						ConfidenceScore:  0.85,
						ProcessingTimeMs: 100,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "invalid factor type",
			requestBody: map[string]interface{}{
				"factor_type": "",
				"data":        map[string]interface{}{},
				"weight":      0.3,
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			tt.mockSetup()

			// Create request
			requestBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/v1/risk/factors/calculate", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.RiskFactorCalculationHandler(rr, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestEnhancedRiskHandler_RiskRecommendationsHandler(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockEnhancedRiskService)
	handler := NewEnhancedRiskHandler(logger, mockService)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful recommendations",
			requestBody: map[string]interface{}{
				"business_id": "test-business-456",
				"risk_factors": []map[string]interface{}{
					{
						"factor_type": "financial",
						"score":       0.75,
						"risk_level":  "high",
					},
				},
			},
			mockSetup: func() {
				mockService.On("PerformEnhancedRiskAssessment", mock.Anything, mock.AnythingOfType("*risk.EnhancedRiskAssessmentRequest")).
					Return(&risk.EnhancedRiskAssessmentResponse{
						AssessmentID:     "rec-test-123",
						BusinessID:       "test-business-456",
						Timestamp:        time.Now(),
						OverallRiskScore: 0.75,
						OverallRiskLevel: risk.RiskLevelHigh,
						RiskFactors: []risk.RiskFactorDetail{
							{
								FactorType:  "financial",
								Score:       0.75,
								RiskLevel:   risk.RiskLevelHigh,
								Confidence:  0.85,
								Weight:      0.3,
								Description: "Financial risk assessment",
								LastUpdated: time.Now(),
							},
						},
						Recommendations: []risk.RecommendationDetail{
							{
								ID:          "rec-1",
								Title:       "Improve Financial Health",
								Description: "Reduce debt-to-revenue ratio",
								Priority:    risk.PriorityHigh,
								Category:    "financial",
								Impact:      "high",
								Effort:      "medium",
								Timeline:    "3-6 months",
								CreatedAt:   time.Now(),
								UpdatedAt:   time.Now(),
							},
						},
						Alerts:           []risk.AlertDetail{},
						ConfidenceScore:  0.85,
						ProcessingTimeMs: 120,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			tt.mockSetup()

			// Create request
			requestBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/v1/risk/recommendations", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.RiskRecommendationsHandler(rr, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestEnhancedRiskHandler_RiskTrendAnalysisHandler(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockEnhancedRiskService)
	handler := NewEnhancedRiskHandler(logger, mockService)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful trend analysis",
			requestBody: map[string]interface{}{
				"business_id": "test-business-456",
				"time_range": map[string]interface{}{
					"start_time": "2024-01-01T00:00:00Z",
					"end_time":   "2024-12-31T23:59:59Z",
					"duration":   "1 year",
				},
			},
			mockSetup: func() {
				mockService.On("PerformEnhancedRiskAssessment", mock.Anything, mock.AnythingOfType("*risk.EnhancedRiskAssessmentRequest")).
					Return(&risk.EnhancedRiskAssessmentResponse{
						AssessmentID:     "trend-test-123",
						BusinessID:       "test-business-456",
						Timestamp:        time.Now(),
						OverallRiskScore: 0.60,
						OverallRiskLevel: risk.RiskLevelMedium,
						RiskFactors: []risk.RiskFactorDetail{
							{
								FactorType:  "financial",
								Score:       0.60,
								RiskLevel:   risk.RiskLevelMedium,
								Confidence:  0.80,
								Weight:      0.3,
								Description: "Financial risk trend analysis",
								LastUpdated: time.Now(),
							},
						},
						Recommendations: []risk.RecommendationDetail{},
						TrendData: &risk.RiskTrendData{
							BusinessID: "test-business-456",
							Trends: []risk.RiskTrend{
								{
									FactorType:  "financial",
									Direction:   risk.TrendDirectionImproving,
									Magnitude:   0.15,
									Confidence:  0.75,
									Timeframe:   "6 months",
									Description: "Financial risk showing improvement",
									DataPoints: []risk.TrendDataPoint{
										{
											Timestamp: time.Now().AddDate(0, -6, 0),
											Value:     0.75,
											Score:     0.75,
										},
										{
											Timestamp: time.Now(),
											Value:     0.60,
											Score:     0.60,
										},
									},
								},
							},
							LastAnalyzed: time.Now(),
							DataPoints:   2,
							TrendSummary: "Overall improving trend",
						},
						Alerts:           []risk.AlertDetail{},
						ConfidenceScore:  0.80,
						ProcessingTimeMs: 200,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			tt.mockSetup()

			// Create request
			requestBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/v1/risk/trends/analyze", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.RiskTrendAnalysisHandler(rr, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestEnhancedRiskHandler_RiskAlertsHandler(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockEnhancedRiskService)
	handler := NewEnhancedRiskHandler(logger, mockService)

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:        "successful alerts retrieval",
			queryParams: "?business_id=test-business-456",
			mockSetup: func() {
				mockService.On("GetActiveAlerts", mock.Anything, "test-business-456").
					Return([]risk.AlertDetail{
						{
							ID:           "alert-1",
							BusinessID:   "test-business-456",
							AlertType:    "threshold_exceeded",
							Severity:     risk.AlertSeverityHigh,
							Title:        "High Financial Risk",
							Description:  "Financial risk score exceeds threshold",
							RiskFactor:   "financial",
							Threshold:    0.7,
							CurrentValue: 0.85,
							Status:       risk.AlertStatusActive,
							CreatedAt:    time.Now(),
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "missing business_id parameter",
			queryParams:    "",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			tt.mockSetup()

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/v1/risk/alerts"+tt.queryParams, nil)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.RiskAlertsHandler(rr, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestEnhancedRiskHandler_AcknowledgeAlertHandler(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockEnhancedRiskService)
	handler := NewEnhancedRiskHandler(logger, mockService)

	tests := []struct {
		name           string
		alertID        string
		requestBody    interface{}
		mockSetup      func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:    "successful alert acknowledgment",
			alertID: "alert-123",
			requestBody: map[string]interface{}{
				"user_id": "user-456",
				"notes":   "Alert acknowledged by user",
			},
			mockSetup: func() {
				mockService.On("AcknowledgeAlert", mock.Anything, "alert-123", "user-456", "Alert acknowledged by user").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "invalid alert ID",
			alertID:        "",
			requestBody:    map[string]interface{}{},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			tt.mockSetup()

			// Create request
			requestBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/v1/risk/alerts/"+tt.alertID+"/acknowledge", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.AcknowledgeAlertHandler(rr, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestEnhancedRiskHandler_ResolveAlertHandler(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockEnhancedRiskService)
	handler := NewEnhancedRiskHandler(logger, mockService)

	tests := []struct {
		name           string
		alertID        string
		requestBody    interface{}
		mockSetup      func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:    "successful alert resolution",
			alertID: "alert-123",
			requestBody: map[string]interface{}{
				"user_id":          "user-456",
				"resolution_notes": "Alert resolved - risk mitigated",
			},
			mockSetup: func() {
				mockService.On("ResolveAlert", mock.Anything, "alert-123", "user-456", "Alert resolved - risk mitigated").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "invalid alert ID",
			alertID:        "",
			requestBody:    map[string]interface{}{},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			tt.mockSetup()

			// Create request
			requestBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/v1/risk/alerts/"+tt.alertID+"/resolve", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.ResolveAlertHandler(rr, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestEnhancedRiskHandler_RiskFactorHistoryHandler(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockEnhancedRiskService)
	handler := NewEnhancedRiskHandler(logger, mockService)

	tests := []struct {
		name           string
		factorID       string
		queryParams    string
		mockSetup      func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:        "successful history retrieval",
			factorID:    "financial",
			queryParams: "?business_id=test-business-456&start_time=2024-01-01T00:00:00Z&end_time=2024-12-31T23:59:59Z",
			mockSetup: func() {
				mockService.On("GetRiskFactorHistory", mock.Anything, "test-business-456", "financial", mock.AnythingOfType("*risk.TimeRange")).
					Return([]risk.RiskHistoryEntry{
						{
							ID:           "history-1",
							BusinessID:   "test-business-456",
							AssessmentID: "assessment-1",
							Timestamp:    time.Now().AddDate(0, -1, 0),
							RiskScore:    0.75,
							RiskLevel:    risk.RiskLevelHigh,
							FactorScores: map[string]float64{
								"financial": 0.75,
							},
							Confidence: 0.80,
						},
						{
							ID:           "history-2",
							BusinessID:   "test-business-456",
							AssessmentID: "assessment-2",
							Timestamp:    time.Now(),
							RiskScore:    0.65,
							RiskLevel:    risk.RiskLevelMedium,
							FactorScores: map[string]float64{
								"financial": 0.65,
							},
							Confidence: 0.85,
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "missing business_id parameter",
			factorID:       "financial",
			queryParams:    "",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			tt.mockSetup()

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/v1/risk/factors/"+tt.factorID+"/history"+tt.queryParams, nil)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.RiskFactorHistoryHandler(rr, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}
