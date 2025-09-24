package risk

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockRiskStorageService is a mock implementation of RiskStorageService
type MockRiskStorageService struct {
	mock.Mock
}

func (m *MockRiskStorageService) StoreRiskAssessment(ctx context.Context, assessment *RiskAssessment) error {
	args := m.Called(ctx, assessment)
	return args.Error(0)
}

func (m *MockRiskStorageService) GetRiskAssessment(ctx context.Context, id string) (*RiskAssessment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RiskAssessment), args.Error(1)
}

func (m *MockRiskStorageService) GetRiskAssessmentsByBusinessID(ctx context.Context, businessID string, limit, offset int) ([]*RiskAssessment, error) {
	args := m.Called(ctx, businessID, limit, offset)
	return args.Get(0).([]*RiskAssessment), args.Error(1)
}

func (m *MockRiskStorageService) UpdateRiskAssessment(ctx context.Context, assessment *RiskAssessment) error {
	args := m.Called(ctx, assessment)
	return args.Error(0)
}

func (m *MockRiskStorageService) DeleteRiskAssessment(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestRiskHistoryTrackingService_GetRiskHistory(t *testing.T) {
	logger := zap.NewNop()
	mockStorage := new(MockRiskStorageService)
	service := NewRiskHistoryTrackingService(mockStorage, logger)

	businessID := "business-123"
	now := time.Now()

	// Create mock assessments
	assessments := []*RiskAssessment{
		{
			ID:           uuid.New().String(),
			BusinessID:   businessID,
			BusinessName: "Test Business",
			OverallScore: 75.0,
			OverallLevel: RiskLevelMedium,
			AssessedAt:   now,
			Alerts: []RiskAlert{
				{
					ID:           "alert-1",
					BusinessID:   businessID,
					RiskFactor:   "factor-1",
					Level:        RiskLevelMedium,
					Message:      "Test alert",
					Score:        75.0,
					Threshold:    70.0,
					TriggeredAt:  now,
					Acknowledged: false,
				},
			},
			Recommendations: []RiskRecommendation{
				{
					ID:          "rec-1",
					RiskFactor:  "factor-1",
					Title:       "Test Recommendation",
					Description: "Test description",
					Priority:    RiskLevelMedium,
					Action:      "Test action",
					Impact:      "Test impact",
					Timeline:    "30 days",
					CreatedAt:   now,
				},
			},
		},
		{
			ID:           uuid.New().String(),
			BusinessID:   businessID,
			BusinessName: "Test Business",
			OverallScore: 80.0,
			OverallLevel: RiskLevelHigh,
			AssessedAt:   now.Add(-24 * time.Hour),
			Alerts:       []RiskAlert{},
			Recommendations: []RiskRecommendation{
				{
					ID:          "rec-2",
					RiskFactor:  "factor-2",
					Title:       "Previous Recommendation",
					Description: "Previous description",
					Priority:    RiskLevelHigh,
					Action:      "Previous action",
					Impact:      "Previous impact",
					Timeline:    "15 days",
					CreatedAt:   now.Add(-24 * time.Hour),
				},
			},
		},
	}

	// Set up mock expectations
	mockStorage.On("GetRiskAssessmentsByBusinessID", mock.Anything, businessID, 50, 0).Return(assessments, nil)

	query := &RiskHistoryQuery{
		BusinessID: businessID,
		Limit:      50,
		Offset:     0,
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request")
	response, err := service.GetRiskHistory(ctx, query)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, businessID, response.BusinessID)
	assert.Equal(t, "Test Business", response.BusinessName)
	assert.Equal(t, 2, response.TotalAssessments)
	assert.Equal(t, 75.0, response.CurrentScore)
	assert.Equal(t, "medium", response.CurrentLevel)
	assert.Len(t, response.History, 2)
	assert.NotNil(t, response.Statistics)
	assert.Equal(t, "improving", response.Trend)

	// Verify history entries
	assert.Equal(t, "baseline", response.History[0].Trend)
	assert.Equal(t, "improving", response.History[1].Trend)
	assert.Equal(t, -5.0, response.History[1].ScoreChange)
	assert.Equal(t, "down", response.History[1].LevelChange)
	assert.Equal(t, 1, response.History[1].DaysSinceLast)
	assert.Equal(t, 1, response.History[1].AlertCount)
	assert.Equal(t, "Test Recommendation", response.History[1].Recommendation)

	mockStorage.AssertExpectations(t)
}

func TestRiskHistoryTrackingService_GetRiskHistory_NoAssessments(t *testing.T) {
	logger := zap.NewNop()
	mockStorage := new(MockRiskStorageService)
	service := NewRiskHistoryTrackingService(mockStorage, logger)

	businessID := "business-123"

	// Set up mock expectations for no assessments
	mockStorage.On("GetRiskAssessmentsByBusinessID", mock.Anything, businessID, 50, 0).Return([]*RiskAssessment{}, nil)

	query := &RiskHistoryQuery{
		BusinessID: businessID,
		Limit:      50,
		Offset:     0,
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request")
	response, err := service.GetRiskHistory(ctx, query)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, businessID, response.BusinessID)
	assert.Equal(t, 0, response.TotalAssessments)
	assert.Len(t, response.History, 0)
	assert.NotNil(t, response.Statistics)

	mockStorage.AssertExpectations(t)
}

func TestRiskHistoryTrackingService_GetRiskTrends(t *testing.T) {
	logger := zap.NewNop()
	mockStorage := new(MockRiskStorageService)
	service := NewRiskHistoryTrackingService(mockStorage, logger)

	businessID := "business-123"
	now := time.Now()

	// Create mock assessments for trend analysis
	assessments := []*RiskAssessment{
		{
			ID:           uuid.New().String(),
			BusinessID:   businessID,
			OverallScore: 70.0,
			OverallLevel: RiskLevelMedium,
			AssessedAt:   now,
		},
		{
			ID:           uuid.New().String(),
			BusinessID:   businessID,
			OverallScore: 75.0,
			OverallLevel: RiskLevelMedium,
			AssessedAt:   now.Add(-24 * time.Hour),
		},
		{
			ID:           uuid.New().String(),
			BusinessID:   businessID,
			OverallScore: 80.0,
			OverallLevel: RiskLevelHigh,
			AssessedAt:   now.Add(-48 * time.Hour),
		},
	}

	// Set up mock expectations
	mockStorage.On("GetRiskAssessmentsByBusinessID", mock.Anything, businessID, 100, 0).Return(assessments, nil)

	ctx := context.WithValue(context.Background(), "request_id", "test-request")
	trends, err := service.GetRiskTrends(ctx, businessID, 7)

	assert.NoError(t, err)
	assert.NotNil(t, trends)
	assert.Equal(t, "improving", trends["trend"])
	assert.Contains(t, trends, "volatility")
	assert.Contains(t, trends, "positive_changes")
	assert.Contains(t, trends, "negative_changes")
	assert.Contains(t, trends, "neutral_changes")
	assert.Contains(t, trends, "assessments_analyzed")

	mockStorage.AssertExpectations(t)
}

func TestRiskHistoryTrackingService_GetRiskTrends_InsufficientData(t *testing.T) {
	logger := zap.NewNop()
	mockStorage := new(MockRiskStorageService)
	service := NewRiskHistoryTrackingService(mockStorage, logger)

	businessID := "business-123"

	// Set up mock expectations for insufficient data
	mockStorage.On("GetRiskAssessmentsByBusinessID", mock.Anything, businessID, 100, 0).Return([]*RiskAssessment{}, nil)

	ctx := context.WithValue(context.Background(), "request_id", "test-request")
	trends, err := service.GetRiskTrends(ctx, businessID, 7)

	assert.NoError(t, err)
	assert.NotNil(t, trends)
	assert.Equal(t, "insufficient_data", trends["trend"])
	assert.Equal(t, 0, trends["assessments_count"])

	mockStorage.AssertExpectations(t)
}

func TestRiskHistoryTrackingService_DetermineOverallTrend(t *testing.T) {
	logger := zap.NewNop()
	service := NewRiskHistoryTrackingService(nil, logger)

	// Test improving trend
	assessments := []*RiskAssessment{
		{OverallScore: 70.0, AssessedAt: time.Now()},
		{OverallScore: 75.0, AssessedAt: time.Now().Add(-24 * time.Hour)},
		{OverallScore: 80.0, AssessedAt: time.Now().Add(-48 * time.Hour)},
	}

	trend := service.determineOverallTrend(assessments)
	assert.Equal(t, "improving", trend)

	// Test declining trend
	assessments = []*RiskAssessment{
		{OverallScore: 80.0, AssessedAt: time.Now()},
		{OverallScore: 75.0, AssessedAt: time.Now().Add(-24 * time.Hour)},
		{OverallScore: 70.0, AssessedAt: time.Now().Add(-48 * time.Hour)},
	}

	trend = service.determineOverallTrend(assessments)
	assert.Equal(t, "declining", trend)

	// Test stable trend
	assessments = []*RiskAssessment{
		{OverallScore: 75.0, AssessedAt: time.Now()},
		{OverallScore: 76.0, AssessedAt: time.Now().Add(-24 * time.Hour)},
		{OverallScore: 74.0, AssessedAt: time.Now().Add(-48 * time.Hour)},
	}

	trend = service.determineOverallTrend(assessments)
	assert.Equal(t, "stable", trend)

	// Test insufficient data
	assessments = []*RiskAssessment{
		{OverallScore: 75.0, AssessedAt: time.Now()},
	}

	trend = service.determineOverallTrend(assessments)
	assert.Equal(t, "insufficient_data", trend)
}

func TestRiskHistoryTrackingService_CalculateStatistics(t *testing.T) {
	logger := zap.NewNop()
	service := NewRiskHistoryTrackingService(nil, logger)

	assessments := []*RiskAssessment{
		{
			OverallScore: 70.0,
			OverallLevel: RiskLevelLow,
			Alerts: []RiskAlert{
				{Level: RiskLevelLow},
				{Level: RiskLevelMedium},
			},
		},
		{
			OverallScore: 80.0,
			OverallLevel: RiskLevelHigh,
			Alerts: []RiskAlert{
				{Level: RiskLevelHigh},
			},
		},
		{
			OverallScore: 75.0,
			OverallLevel: RiskLevelMedium,
			Alerts: []RiskAlert{
				{Level: RiskLevelMedium},
			},
		},
	}

	statistics := service.calculateStatistics(assessments)

	assert.Equal(t, 75.0, statistics["average_score"])
	assert.Equal(t, 70.0, statistics["min_score"])
	assert.Equal(t, 80.0, statistics["max_score"])
	assert.Equal(t, 10.0, statistics["score_range"])

	levelDistribution := statistics["level_distribution"].(map[string]int)
	assert.Equal(t, 1, levelDistribution["low"])
	assert.Equal(t, 1, levelDistribution["medium"])
	assert.Equal(t, 1, levelDistribution["high"])

	alertDistribution := statistics["alert_distribution"].(map[string]int)
	assert.Equal(t, 1, alertDistribution["low"])
	assert.Equal(t, 2, alertDistribution["medium"])
	assert.Equal(t, 1, alertDistribution["high"])
}

func TestRiskHistoryTrackingService_NewRiskHistoryTrackingService(t *testing.T) {
	logger := zap.NewNop()
	mockStorage := new(MockRiskStorageService)
	service := NewRiskHistoryTrackingService(mockStorage, logger)

	assert.NotNil(t, service)
	assert.Equal(t, mockStorage, service.storageService)
	assert.Equal(t, logger, service.logger)
}
