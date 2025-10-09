package engine

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// MockMLService is a mock implementation of the ML service
type MockMLService struct {
	mock.Mock
}

func (m *MockMLService) InitializeModels(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockMLService) PredictRisk(ctx context.Context, modelType string, req *models.RiskAssessmentRequest) (*models.RiskAssessment, error) {
	args := m.Called(ctx, modelType, req)
	return args.Get(0).(*models.RiskAssessment), args.Error(1)
}

func (m *MockMLService) PredictFutureRisk(ctx context.Context, modelType string, req *models.RiskAssessmentRequest, horizonMonths int) (*models.RiskPrediction, error) {
	args := m.Called(ctx, modelType, req, horizonMonths)
	return args.Get(0).(*models.RiskPrediction), args.Error(1)
}

// Test helper functions
func createTestRiskEngine() (*RiskEngine, *MockMLService) {
	logger := zap.NewNop()
	config := &Config{
		MaxConcurrentRequests: 10,
		RequestTimeout:        1 * time.Second,
		CacheTTL:              5 * time.Minute,
		CircuitBreakerConfig: CircuitBreakerConfig{
			FailureThreshold: 5,
			RecoveryTimeout:  30 * time.Second,
			HalfOpenMaxCalls: 3,
		},
		EnableMetrics: true,
		EnableCaching: true,
	}

	mockML := &MockMLService{}
	engine := NewRiskEngine(mockML, logger, config)

	return engine, mockML
}

func TestRiskEngine_AssessRisk_Success(t *testing.T) {
	engine, mockML := createTestRiskEngine()

	req := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	expectedAssessment := &models.RiskAssessment{
		ID:                "test-123",
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
		RiskScore:         0.75,
		RiskLevel:         models.RiskLevelMedium,
		ConfidenceScore:   0.85,
		Status:            models.StatusCompleted,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		RiskFactors: []models.RiskFactor{
			{
				Category:    models.RiskCategoryFinancial,
				Name:        "Credit Score",
				Score:       0.8,
				Weight:      0.3,
				Description: "Business credit score analysis",
				Source:      "internal",
				Confidence:  0.9,
			},
		},
	}

	mockML.On("PredictRisk", mock.Anything, "xgboost", req).Return(expectedAssessment, nil)

	result, err := engine.AssessRisk(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedAssessment.ID, result.ID)
	assert.Equal(t, expectedAssessment.RiskScore, result.RiskScore)
	assert.Equal(t, expectedAssessment.RiskLevel, result.RiskLevel)

	mockML.AssertExpectations(t)
}

func TestRiskEngine_AssessRisk_MLServiceError(t *testing.T) {
	engine, mockML := createTestRiskEngine()

	req := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	mockML.On("PredictRisk", mock.Anything, "xgboost", req).Return((*models.RiskAssessment)(nil), assert.AnError)

	result, err := engine.AssessRisk(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)

	mockML.AssertExpectations(t)
}

func TestRiskEngine_AssessRisk_Timeout(t *testing.T) {
	engine, mockML := createTestRiskEngine()

	req := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	// Create a context with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Mock ML service to take longer than the timeout
	mockML.On("PredictRisk", mock.Anything, "xgboost", req).Run(func(args mock.Arguments) {
		time.Sleep(10 * time.Millisecond) // Longer than context timeout
	}).Return((*models.RiskAssessment)(nil), context.DeadlineExceeded)

	result, err := engine.AssessRisk(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "context deadline exceeded")

	mockML.AssertExpectations(t)
}

func TestRiskEngine_AssessRiskBatch_Success(t *testing.T) {
	engine, mockML := createTestRiskEngine()

	requests := []*models.RiskAssessmentRequest{
		{
			BusinessName:      "Company 1",
			BusinessAddress:   "123 St 1",
			Industry:          "Tech",
			Country:           "US",
			PredictionHorizon: 3,
		},
		{
			BusinessName:      "Company 2",
			BusinessAddress:   "456 St 2",
			Industry:          "Finance",
			Country:           "US",
			PredictionHorizon: 6,
		},
	}

	expectedAssessments := []*models.RiskAssessment{
		{
			ID:              "batch-1",
			BusinessName:    "Company 1",
			RiskScore:       0.7,
			RiskLevel:       models.RiskLevelMedium,
			ConfidenceScore: 0.8,
			Status:          models.StatusCompleted,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "batch-2",
			BusinessName:    "Company 2",
			RiskScore:       0.6,
			RiskLevel:       models.RiskLevelLow,
			ConfidenceScore: 0.9,
			Status:          models.StatusCompleted,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	// Mock ML service calls for each request
	mockML.On("PredictRisk", mock.Anything, "xgboost", requests[0]).Return(expectedAssessments[0], nil)
	mockML.On("PredictRisk", mock.Anything, "xgboost", requests[1]).Return(expectedAssessments[1], nil)

	results, err := engine.AssessRiskBatch(context.Background(), requests)

	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 2)
	assert.Equal(t, expectedAssessments[0].ID, results[0].ID)
	assert.Equal(t, expectedAssessments[1].ID, results[1].ID)

	mockML.AssertExpectations(t)
}

func TestRiskEngine_AssessRiskBatch_PartialFailure(t *testing.T) {
	engine, mockML := createTestRiskEngine()

	requests := []*models.RiskAssessmentRequest{
		{
			BusinessName:      "Company 1",
			BusinessAddress:   "123 St 1",
			Industry:          "Tech",
			Country:           "US",
			PredictionHorizon: 3,
		},
		{
			BusinessName:      "Company 2",
			BusinessAddress:   "456 St 2",
			Industry:          "Finance",
			Country:           "US",
			PredictionHorizon: 6,
		},
	}

	expectedAssessment := &models.RiskAssessment{
		ID:              "batch-1",
		BusinessName:    "Company 1",
		RiskScore:       0.7,
		RiskLevel:       models.RiskLevelMedium,
		ConfidenceScore: 0.8,
		Status:          models.StatusCompleted,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Mock ML service calls - one success, one failure
	mockML.On("PredictRisk", mock.Anything, "xgboost", requests[0]).Return(expectedAssessment, nil)
	mockML.On("PredictRisk", mock.Anything, "xgboost", requests[1]).Return((*models.RiskAssessment)(nil), assert.AnError)

	results, err := engine.AssessRiskBatch(context.Background(), requests)

	assert.Error(t, err) // Should return error due to partial failure
	assert.NotNil(t, results)
	assert.Len(t, results, 1) // Only one successful result
	assert.Equal(t, expectedAssessment.ID, results[0].ID)

	mockML.AssertExpectations(t)
}

func TestRiskEngine_PredictRisk_Success(t *testing.T) {
	engine, mockML := createTestRiskEngine()

	req := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	expectedPrediction := &models.RiskPrediction{
		BusinessID:     "test-123",
		HorizonMonths:  6,
		PredictedScore: 0.8,
		Confidence:     0.85,
		RiskFactors: []models.RiskFactor{
			{
				Category:    models.RiskCategoryFinancial,
				Name:        "Future Credit Risk",
				Score:       0.8,
				Weight:      0.4,
				Description: "Predicted credit risk over 6 months",
				Source:      "ml_model",
				Confidence:  0.85,
			},
		},
		CreatedAt: time.Now(),
	}

	mockML.On("PredictFutureRisk", mock.Anything, "xgboost", req, 6).Return(expectedPrediction, nil)

	result, err := engine.PredictRisk(context.Background(), req, 6)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedPrediction.BusinessID, result.BusinessID)
	assert.Equal(t, expectedPrediction.PredictedScore, result.PredictedScore)

	mockML.AssertExpectations(t)
}

func TestRiskEngine_PredictRisk_MLServiceError(t *testing.T) {
	engine, mockML := createTestRiskEngine()

	req := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	mockML.On("PredictFutureRisk", mock.Anything, "xgboost", req, 6).Return((*models.RiskPrediction)(nil), assert.AnError)

	result, err := engine.PredictRisk(context.Background(), req, 6)

	assert.Error(t, err)
	assert.Nil(t, result)

	mockML.AssertExpectations(t)
}

func TestRiskEngine_GetCacheStats(t *testing.T) {
	engine, _ := createTestRiskEngine()

	stats := engine.GetCacheStats()

	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.HitCount, int64(0))
	assert.GreaterOrEqual(t, stats.MissCount, int64(0))
	assert.GreaterOrEqual(t, stats.Size, int64(0))
}

func TestRiskEngine_GetCircuitBreakerState(t *testing.T) {
	engine, _ := createTestRiskEngine()

	state := engine.GetCircuitBreakerState()

	assert.NotNil(t, state)
	assert.Contains(t, []CircuitBreakerState{CircuitBreakerStateClosed, CircuitBreakerStateOpen, CircuitBreakerStateHalfOpen}, state)
}

func TestRiskEngine_Shutdown_Success(t *testing.T) {
	engine, _ := createTestRiskEngine()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := engine.Shutdown(ctx)

	assert.NoError(t, err)
}

func TestRiskEngine_Shutdown_Timeout(t *testing.T) {
	engine, _ := createTestRiskEngine()

	// Create a context with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	err := engine.Shutdown(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

// Test concurrent access
func TestRiskEngine_ConcurrentAccess(t *testing.T) {
	engine, mockML := createTestRiskEngine()

	req := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	expectedAssessment := &models.RiskAssessment{
		ID:                "test-123",
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
		RiskScore:         0.75,
		RiskLevel:         models.RiskLevelMedium,
		ConfidenceScore:   0.85,
		Status:            models.StatusCompleted,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Mock multiple calls
	mockML.On("PredictRisk", mock.Anything, "xgboost", req).Return(expectedAssessment, nil).Times(10)

	// Run concurrent requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			result, err := engine.AssessRisk(context.Background(), req)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	mockML.AssertExpectations(t)
}

// Test cache behavior
func TestRiskEngine_CacheBehavior(t *testing.T) {
	engine, mockML := createTestRiskEngine()

	req := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	expectedAssessment := &models.RiskAssessment{
		ID:                "test-123",
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
		RiskScore:         0.75,
		RiskLevel:         models.RiskLevelMedium,
		ConfidenceScore:   0.85,
		Status:            models.StatusCompleted,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// First call should hit ML service
	mockML.On("PredictRisk", mock.Anything, "xgboost", req).Return(expectedAssessment, nil).Once()

	result1, err1 := engine.AssessRisk(context.Background(), req)
	assert.NoError(t, err1)
	assert.NotNil(t, result1)

	// Second call should hit cache (no additional ML service call)
	result2, err2 := engine.AssessRisk(context.Background(), req)
	assert.NoError(t, err2)
	assert.NotNil(t, result2)

	// Results should be the same
	assert.Equal(t, result1.ID, result2.ID)
	assert.Equal(t, result1.RiskScore, result2.RiskScore)

	mockML.AssertExpectations(t)
}

// Test circuit breaker behavior
func TestRiskEngine_CircuitBreakerBehavior(t *testing.T) {
	engine, mockML := createTestRiskEngine()

	req := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	// Mock multiple failures to trigger circuit breaker
	mockML.On("PredictRisk", mock.Anything, "xgboost", req).Return((*models.RiskAssessment)(nil), assert.AnError).Times(6)

	// First few calls should fail
	for i := 0; i < 5; i++ {
		result, err := engine.AssessRisk(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, result)
	}

	// Circuit breaker should be open now
	state := engine.GetCircuitBreakerState()
	assert.Equal(t, CircuitBreakerStateOpen, state)

	// Additional calls should fail immediately without calling ML service
	result, err := engine.AssessRisk(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, result)

	mockML.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkRiskEngine_AssessRisk(b *testing.B) {
	engine, mockML := createTestRiskEngine()

	req := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	expectedAssessment := &models.RiskAssessment{
		ID:                "test-123",
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
		RiskScore:         0.75,
		RiskLevel:         models.RiskLevelMedium,
		ConfidenceScore:   0.85,
		Status:            models.StatusCompleted,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	mockML.On("PredictRisk", mock.Anything, "xgboost", req).Return(expectedAssessment, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.AssessRisk(context.Background(), req)
	}
}

func BenchmarkRiskEngine_AssessRiskBatch(b *testing.B) {
	engine, mockML := createTestRiskEngine()

	requests := make([]*models.RiskAssessmentRequest, 10)
	for i := 0; i < 10; i++ {
		requests[i] = &models.RiskAssessmentRequest{
			BusinessName:      fmt.Sprintf("Company %d", i),
			BusinessAddress:   "123 Test St",
			Industry:          "Technology",
			Country:           "US",
			PredictionHorizon: 3,
		}
	}

	expectedAssessment := &models.RiskAssessment{
		ID:                "test-123",
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
		RiskScore:         0.75,
		RiskLevel:         models.RiskLevelMedium,
		ConfidenceScore:   0.85,
		Status:            models.StatusCompleted,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Mock calls for each request
	for i := 0; i < 10; i++ {
		mockML.On("PredictRisk", mock.Anything, "xgboost", requests[i]).Return(expectedAssessment, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.AssessRiskBatch(context.Background(), requests)
	}
}
