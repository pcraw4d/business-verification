package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// MockModelManager is a mock implementation of the model manager
type MockModelManager struct {
	mock.Mock
}

func (m *MockModelManager) RegisterModel(name string, model mlmodels.RiskModel) error {
	args := m.Called(name, model)
	return args.Error(0)
}

func (m *MockModelManager) GetModel(name string) (mlmodels.RiskModel, error) {
	args := m.Called(name)
	return args.Get(0).(mlmodels.RiskModel), args.Error(1)
}

func (m *MockModelManager) ListModels() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockModelManager) RemoveModel(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// MockModelTrainer is a mock implementation of the model trainer
type MockModelTrainer struct {
	mock.Mock
}

func (m *MockModelTrainer) TrainModel(ctx context.Context, modelType string, data []mlmodels.TrainingData) (*mlmodels.TrainingResult, error) {
	args := m.Called(ctx, modelType, data)
	return args.Get(0).(*mlmodels.TrainingResult), args.Error(1)
}

func (m *MockModelTrainer) ValidateModel(ctx context.Context, modelType string, data []mlmodels.TrainingData) (*mlmodels.ValidationResult, error) {
	args := m.Called(ctx, modelType, data)
	return args.Get(0).(*mlmodels.ValidationResult), args.Error(1)
}

func (m *MockModelTrainer) GenerateMockTrainingData(count int) []mlmodels.TrainingData {
	args := m.Called(count)
	return args.Get(0).([]mlmodels.TrainingData)
}

// MockRiskModel is a mock implementation of the risk model
type MockRiskModel struct {
	mock.Mock
}

func (m *MockRiskModel) Predict(ctx context.Context, features []float64) (*mlmodels.PredictionResult, error) {
	args := m.Called(ctx, features)
	return args.Get(0).(*mlmodels.PredictionResult), args.Error(1)
}

func (m *MockRiskModel) Train(ctx context.Context, data []mlmodels.TrainingData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockRiskModel) Validate(ctx context.Context, data []mlmodels.TrainingData) (*mlmodels.ValidationResult, error) {
	args := m.Called(ctx, data)
	return args.Get(0).(*mlmodels.ValidationResult), args.Error(1)
}

func (m *MockRiskModel) GetModelInfo() *mlmodels.ModelInfo {
	args := m.Called()
	return args.Get(0).(*mlmodels.ModelInfo)
}

// Test helper functions
func createTestMLService() (*MLService, *MockModelManager, *MockModelTrainer) {
	logger := zap.NewNop()
	service := NewMLService(logger)

	mockManager := &MockModelManager{}
	mockTrainer := &MockModelTrainer{}

	// Replace the internal components with mocks
	service.modelManager = mockManager
	service.trainer = mockTrainer

	return service, mockManager, mockTrainer
}

func TestMLService_InitializeModels_Success(t *testing.T) {
	service, mockManager, _ := createTestMLService()

	mockModel := &MockRiskModel{}
	mockManager.On("RegisterModel", "xgboost", mock.AnythingOfType("*xgboost.XGBoostModel")).Return(nil)

	err := service.InitializeModels(context.Background())

	assert.NoError(t, err)
	mockManager.AssertExpectations(t)
}

func TestMLService_InitializeModels_RegistrationError(t *testing.T) {
	service, mockManager, _ := createTestMLService()

	mockManager.On("RegisterModel", "xgboost", mock.AnythingOfType("*xgboost.XGBoostModel")).Return(assert.AnError)

	err := service.InitializeModels(context.Background())

	assert.Error(t, err)
	mockManager.AssertExpectations(t)
}

func TestMLService_PredictRisk_Success(t *testing.T) {
	service, mockManager, _ := createTestMLService()

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

	mockModel := &MockRiskModel{}
	mockManager.On("GetModel", "xgboost").Return(mockModel, nil)
	mockModel.On("Predict", mock.Anything, mock.AnythingOfType("[]float64")).Return(&mlmodels.PredictionResult{
		RiskScore:       0.75,
		ConfidenceScore: 0.85,
		RiskFactors: []mlmodels.RiskFactor{
			{
				Category:    "financial",
				Name:        "Credit Score",
				Score:       0.8,
				Weight:      0.3,
				Description: "Business credit score analysis",
				Source:      "internal",
				Confidence:  0.9,
			},
		},
	}, nil)

	result, err := service.PredictRisk(context.Background(), "xgboost", req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedAssessment.RiskScore, result.RiskScore)
	assert.Equal(t, expectedAssessment.RiskLevel, result.RiskLevel)
	assert.Equal(t, expectedAssessment.ConfidenceScore, result.ConfidenceScore)

	mockManager.AssertExpectations(t)
	mockModel.AssertExpectations(t)
}

func TestMLService_PredictRisk_ModelNotFound(t *testing.T) {
	service, mockManager, _ := createTestMLService()

	req := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	mockManager.On("GetModel", "nonexistent").Return((*MockRiskModel)(nil), assert.AnError)

	result, err := service.PredictRisk(context.Background(), "nonexistent", req)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockManager.AssertExpectations(t)
}

func TestMLService_PredictRisk_PredictionError(t *testing.T) {
	service, mockManager, _ := createTestMLService()

	req := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	mockModel := &MockRiskModel{}
	mockManager.On("GetModel", "xgboost").Return(mockModel, nil)
	mockModel.On("Predict", mock.Anything, mock.AnythingOfType("[]float64")).Return((*mlmodels.PredictionResult)(nil), assert.AnError)

	result, err := service.PredictRisk(context.Background(), "xgboost", req)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockManager.AssertExpectations(t)
	mockModel.AssertExpectations(t)
}

func TestMLService_PredictFutureRisk_Success(t *testing.T) {
	service, mockManager, _ := createTestMLService()

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

	mockModel := &MockRiskModel{}
	mockManager.On("GetModel", "xgboost").Return(mockModel, nil)
	mockModel.On("Predict", mock.Anything, mock.AnythingOfType("[]float64")).Return(&mlmodels.PredictionResult{
		RiskScore:       0.8,
		ConfidenceScore: 0.85,
		RiskFactors: []mlmodels.RiskFactor{
			{
				Category:    "financial",
				Name:        "Future Credit Risk",
				Score:       0.8,
				Weight:      0.4,
				Description: "Predicted credit risk over 6 months",
				Source:      "ml_model",
				Confidence:  0.85,
			},
		},
	}, nil)

	result, err := service.PredictFutureRisk(context.Background(), "xgboost", req, 6)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedPrediction.PredictedScore, result.PredictedScore)
	assert.Equal(t, expectedPrediction.Confidence, result.Confidence)
	assert.Equal(t, expectedPrediction.HorizonMonths, result.HorizonMonths)

	mockManager.AssertExpectations(t)
	mockModel.AssertExpectations(t)
}

func TestMLService_PredictFutureRisk_ModelNotFound(t *testing.T) {
	service, mockManager, _ := createTestMLService()

	req := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	mockManager.On("GetModel", "nonexistent").Return((*MockRiskModel)(nil), assert.AnError)

	result, err := service.PredictFutureRisk(context.Background(), "nonexistent", req, 6)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockManager.AssertExpectations(t)
}

func TestMLService_PredictFutureRisk_PredictionError(t *testing.T) {
	service, mockManager, _ := createTestMLService()

	req := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	mockModel := &MockRiskModel{}
	mockManager.On("GetModel", "xgboost").Return(mockModel, nil)
	mockModel.On("Predict", mock.Anything, mock.AnythingOfType("[]float64")).Return((*mlmodels.PredictionResult)(nil), assert.AnError)

	result, err := service.PredictFutureRisk(context.Background(), "xgboost", req, 6)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockManager.AssertExpectations(t)
	mockModel.AssertExpectations(t)
}

func TestMLService_TrainModel_Success(t *testing.T) {
	service, _, mockTrainer := createTestMLService()

	trainingData := []mlmodels.TrainingData{
		{
			Features: []float64{1.0, 2.0, 3.0},
			Label:    []float64{0.5},
		},
		{
			Features: []float64{4.0, 5.0, 6.0},
			Label:    []float64{0.7},
		},
	}

	expectedResult := &mlmodels.TrainingResult{
		ModelType:      "xgboost",
		Accuracy:       0.85,
		Loss:           0.15,
		TrainingTime:   time.Second,
		ValidationLoss: 0.12,
		Epochs:         100,
		Status:         "completed",
	}

	mockTrainer.On("TrainModel", mock.Anything, "xgboost", trainingData).Return(expectedResult, nil)

	result, err := service.TrainModel(context.Background(), "xgboost", trainingData)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedResult.ModelType, result.ModelType)
	assert.Equal(t, expectedResult.Accuracy, result.Accuracy)
	assert.Equal(t, expectedResult.Loss, result.Loss)

	mockTrainer.AssertExpectations(t)
}

func TestMLService_TrainModel_Error(t *testing.T) {
	service, _, mockTrainer := createTestMLService()

	trainingData := []mlmodels.TrainingData{
		{
			Features: []float64{1.0, 2.0, 3.0},
			Label:    []float64{0.5},
		},
	}

	mockTrainer.On("TrainModel", mock.Anything, "xgboost", trainingData).Return((*mlmodels.TrainingResult)(nil), assert.AnError)

	result, err := service.TrainModel(context.Background(), "xgboost", trainingData)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockTrainer.AssertExpectations(t)
}

func TestMLService_ValidateModel_Success(t *testing.T) {
	service, _, mockTrainer := createTestMLService()

	validationData := []mlmodels.TrainingData{
		{
			Features: []float64{1.0, 2.0, 3.0},
			Label:    []float64{0.5},
		},
		{
			Features: []float64{4.0, 5.0, 6.0},
			Label:    []float64{0.7},
		},
	}

	expectedResult := &mlmodels.ValidationResult{
		ModelType:    "xgboost",
		Accuracy:     0.82,
		Precision:    0.85,
		Recall:       0.80,
		F1Score:      0.82,
		ConfusionMatrix: mlmodels.ConfusionMatrix{
			TruePositives:  80,
			TrueNegatives:  15,
			FalsePositives: 5,
			FalseNegatives: 0,
		},
		ValidationTime: time.Second,
		Status:         "completed",
	}

	mockTrainer.On("ValidateModel", mock.Anything, "xgboost", validationData).Return(expectedResult, nil)

	result, err := service.ValidateModel(context.Background(), "xgboost", validationData)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedResult.ModelType, result.ModelType)
	assert.Equal(t, expectedResult.Accuracy, result.Accuracy)
	assert.Equal(t, expectedResult.Precision, result.Precision)
	assert.Equal(t, expectedResult.Recall, result.Recall)

	mockTrainer.AssertExpectations(t)
}

func TestMLService_ValidateModel_Error(t *testing.T) {
	service, _, mockTrainer := createTestMLService()

	validationData := []mlmodels.TrainingData{
		{
			Features: []float64{1.0, 2.0, 3.0},
			Label:    []float64{0.5},
		},
	}

	mockTrainer.On("ValidateModel", mock.Anything, "xgboost", validationData).Return((*mlmodels.ValidationResult)(nil), assert.AnError)

	result, err := service.ValidateModel(context.Background(), "xgboost", validationData)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockTrainer.AssertExpectations(t)
}

func TestMLService_GenerateMockTrainingData_Success(t *testing.T) {
	service, _, mockTrainer := createTestMLService()

	expectedData := []mlmodels.TrainingData{
		{
			Features: []float64{1.0, 2.0, 3.0},
			Label:    []float64{0.5},
		},
		{
			Features: []float64{4.0, 5.0, 6.0},
			Label:    []float64{0.7},
		},
	}

	mockTrainer.On("GenerateMockTrainingData", 100).Return(expectedData)

	result := service.GenerateMockTrainingData(100)

	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, expectedData[0].Features, result[0].Features)
	assert.Equal(t, expectedData[0].Label, result[0].Label)

	mockTrainer.AssertExpectations(t)
}

func TestMLService_GetModelInfo_Success(t *testing.T) {
	service, mockManager, _ := createTestMLService()

	expectedInfo := &mlmodels.ModelInfo{
		ModelType:    "xgboost",
		Version:      "1.0.0",
		LastTrained:  time.Now(),
		Accuracy:     0.85,
		Status:       "active",
		Features:     []string{"feature1", "feature2", "feature3"},
		Description:  "XGBoost risk assessment model",
	}

	mockModel := &MockRiskModel{}
	mockManager.On("GetModel", "xgboost").Return(mockModel, nil)
	mockModel.On("GetModelInfo").Return(expectedInfo)

	result, err := service.GetModelInfo("xgboost")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedInfo.ModelType, result.ModelType)
	assert.Equal(t, expectedInfo.Version, result.Version)
	assert.Equal(t, expectedInfo.Accuracy, result.Accuracy)

	mockManager.AssertExpectations(t)
	mockModel.AssertExpectations(t)
}

func TestMLService_GetModelInfo_ModelNotFound(t *testing.T) {
	service, mockManager, _ := createTestMLService()

	mockManager.On("GetModel", "nonexistent").Return((*MockRiskModel)(nil), assert.AnError)

	result, err := service.GetModelInfo("nonexistent")

	assert.Error(t, err)
	assert.Nil(t, result)
	mockManager.AssertExpectations(t)
}

func TestMLService_ListModels_Success(t *testing.T) {
	service, mockManager, _ := createTestMLService()

	expectedModels := []string{"xgboost", "lstm", "random_forest"}

	mockManager.On("ListModels").Return(expectedModels)

	result := service.ListModels()

	assert.NotNil(t, result)
	assert.Len(t, result, 3)
	assert.Contains(t, result, "xgboost")
	assert.Contains(t, result, "lstm")
	assert.Contains(t, result, "random_forest")

	mockManager.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkMLService_PredictRisk(b *testing.B) {
	service, mockManager, _ := createTestMLService()

	req := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	mockModel := &MockRiskModel{}
	mockManager.On("GetModel", "xgboost").Return(mockModel, nil)
	mockModel.On("Predict", mock.Anything, mock.AnythingOfType("[]float64")).Return(&mlmodels.PredictionResult{
		RiskScore:       0.75,
		ConfidenceScore: 0.85,
		RiskFactors:     []mlmodels.RiskFactor{},
	}, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.PredictRisk(context.Background(), "xgboost", req)
	}
}

func BenchmarkMLService_TrainModel(b *testing.B) {
	service, _, mockTrainer := createTestMLService()

	trainingData := make([]mlmodels.TrainingData, 1000)
	for i := 0; i < 1000; i++ {
		trainingData[i] = mlmodels.TrainingData{
			Features: []float64{float64(i), float64(i + 1), float64(i + 2)},
			Label:    []float64{float64(i) / 1000.0},
		}
	}

	expectedResult := &mlmodels.TrainingResult{
		ModelType:    "xgboost",
		Accuracy:     0.85,
		Loss:         0.15,
		TrainingTime: time.Second,
		Status:       "completed",
	}

	mockTrainer.On("TrainModel", mock.Anything, "xgboost", trainingData).Return(expectedResult, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.TrainModel(context.Background(), "xgboost", trainingData)
	}
}
