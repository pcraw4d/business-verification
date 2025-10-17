//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/config"
	"kyb-platform/services/risk-assessment-service/internal/ml"
	"kyb-platform/services/risk-assessment-service/internal/ml/service"
	"kyb-platform/services/risk-assessment-service/internal/models"
)

// TestMLService represents a test ML service
type TestMLService struct {
	service *service.MLService
	logger  *zap.Logger
}

// SetupTestMLService creates a test ML service
func SetupTestMLService(t *testing.T) *TestMLService {
	logger := zap.NewNop()

	// Load test configuration
	cfg, err := config.Load()
	require.NoError(t, err)

	// Override with test values
	cfg.ML.ModelPath = "./test/models"
	cfg.ML.TrainingData = "./test/data"
	cfg.ML.BatchSize = 16
	cfg.ML.LearningRate = 0.001
	cfg.ML.MaxIterations = 100

	// Create ML service
	service, err := ml.NewService(&cfg.ML, logger)
	if err != nil {
		t.Skipf("Skipping ML integration test: ML service not available: %v", err)
	}

	return &TestMLService{
		service: service,
		logger:  logger,
	}
}

// TeardownTestMLService cleans up test ML service
func (tms *TestMLService) TeardownTestMLService() {
	if tms.service != nil {
		tms.service.Close()
	}
}

func TestMLService_RiskAssessment_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test ML service
	mlService := SetupTestMLService(t)
	defer mlService.TeardownTestMLService()

	ctx := context.Background()

	tests := []struct {
		name           string
		request        *models.RiskAssessmentRequest
		expectedError  bool
		validateResult func(*testing.T, *models.RiskAssessmentResponse)
	}{
		{
			name: "valid risk assessment request",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "Test Company Inc",
				BusinessAddress:   "123 Test Street, Test City, TC 12345",
				Industry:          "technology",
				Country:           "US",
				Phone:             "+1-555-123-4567",
				Email:             "test@company.com",
				Website:           "https://testcompany.com",
				PredictionHorizon: 3,
				ModelType:         "xgboost",
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *models.RiskAssessmentResponse) {
				assert.NotEmpty(t, result.ID)
				assert.NotEmpty(t, result.BusinessID)
				assert.GreaterOrEqual(t, result.RiskScore, 0.0)
				assert.LessOrEqual(t, result.RiskScore, 1.0)
				assert.NotEmpty(t, result.RiskLevel)
				assert.NotEmpty(t, result.RiskFactors)
				assert.Equal(t, 3, result.PredictionHorizon)
				assert.GreaterOrEqual(t, result.ConfidenceScore, 0.0)
				assert.LessOrEqual(t, result.ConfidenceScore, 1.0)
				assert.Equal(t, models.StatusCompleted, result.Status)
				assert.Equal(t, "xgboost", result.ModelType)
				assert.NotZero(t, result.CreatedAt)
				assert.NotZero(t, result.UpdatedAt)
			},
		},
		{
			name: "risk assessment with custom model",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "Custom Model Company",
				BusinessAddress:   "456 Custom Street, Test City, TC 12345",
				Industry:          "finance",
				Country:           "US",
				PredictionHorizon: 6,
				ModelType:         "custom",
				CustomModelID:     "custom-model-123",
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *models.RiskAssessmentResponse) {
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, 6, result.PredictionHorizon)
				assert.Equal(t, "custom", result.ModelType)
				assert.Equal(t, "custom-model-123", result.CustomModelID)
			},
		},
		{
			name: "risk assessment with ensemble model",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "Ensemble Company",
				BusinessAddress:   "789 Ensemble Street, Test City, TC 12345",
				Industry:          "healthcare",
				Country:           "US",
				PredictionHorizon: 12,
				ModelType:         "ensemble",
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *models.RiskAssessmentResponse) {
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, 12, result.PredictionHorizon)
				assert.Equal(t, "ensemble", result.ModelType)
			},
		},
		{
			name: "risk assessment with temporal analysis",
			request: &models.RiskAssessmentRequest{
				BusinessName:            "Temporal Company",
				BusinessAddress:         "321 Temporal Street, Test City, TC 12345",
				Industry:                "manufacturing",
				Country:                 "US",
				PredictionHorizon:       6,
				ModelType:               "lstm",
				IncludeTemporalAnalysis: true,
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *models.RiskAssessmentResponse) {
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, 6, result.PredictionHorizon)
				assert.Equal(t, "lstm", result.ModelType)
			},
		},
		{
			name: "invalid request - missing required fields",
			request: &models.RiskAssessmentRequest{
				BusinessName: "", // Invalid: empty name
				Industry:     "technology",
				Country:      "US",
			},
			expectedError: true,
		},
		{
			name: "invalid request - invalid model type",
			request: &models.RiskAssessmentRequest{
				BusinessName:    "Invalid Model Company",
				BusinessAddress: "123 Invalid Street, Test City, TC 12345",
				Industry:        "technology",
				Country:         "US",
				ModelType:       "invalid-model",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test risk assessment
			result, err := mlService.service.AssessRisk(ctx, tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestMLService_BatchRiskAssessment_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test ML service
	mlService := SetupTestMLService(t)
	defer mlService.TeardownTestMLService()

	ctx := context.Background()

	tests := []struct {
		name           string
		requests       []*models.RiskAssessmentRequest
		expectedError  bool
		validateResult func(*testing.T, *models.BatchRiskAssessmentResponse)
	}{
		{
			name: "valid batch request",
			requests: []*models.RiskAssessmentRequest{
				{
					BusinessName:      "Batch Company 1",
					BusinessAddress:   "123 Batch Street 1, Test City, TC 12345",
					Industry:          "technology",
					Country:           "US",
					PredictionHorizon: 3,
					ModelType:         "xgboost",
				},
				{
					BusinessName:      "Batch Company 2",
					BusinessAddress:   "456 Batch Street 2, Test City, TC 12345",
					Industry:          "finance",
					Country:           "US",
					PredictionHorizon: 6,
					ModelType:         "ensemble",
				},
				{
					BusinessName:      "Batch Company 3",
					BusinessAddress:   "789 Batch Street 3, Test City, TC 12345",
					Industry:          "healthcare",
					Country:           "US",
					PredictionHorizon: 12,
					ModelType:         "lstm",
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *models.BatchRiskAssessmentResponse) {
				assert.NotEmpty(t, result.BatchID)
				assert.Equal(t, models.StatusCompleted, result.Status)
				assert.Len(t, result.Assessments, 3)
				assert.Equal(t, 3, result.TotalCount)
				assert.Equal(t, 3, result.SuccessCount)
				assert.Equal(t, 0, result.ErrorCount)
				assert.NotZero(t, result.CreatedAt)
				assert.NotZero(t, result.CompletedAt)

				// Validate individual assessments
				for i, assessment := range result.Assessments {
					assert.NotEmpty(t, assessment.ID)
					assert.Equal(t, fmt.Sprintf("Batch Company %d", i+1), assessment.BusinessName)
					assert.GreaterOrEqual(t, assessment.RiskScore, 0.0)
					assert.LessOrEqual(t, assessment.RiskScore, 1.0)
					assert.NotEmpty(t, assessment.RiskLevel)
					assert.Equal(t, models.StatusCompleted, assessment.Status)
				}
			},
		},
		{
			name: "batch request with mixed valid and invalid",
			requests: []*models.RiskAssessmentRequest{
				{
					BusinessName:      "Valid Company",
					BusinessAddress:   "123 Valid Street, Test City, TC 12345",
					Industry:          "technology",
					Country:           "US",
					PredictionHorizon: 3,
					ModelType:         "xgboost",
				},
				{
					BusinessName: "", // Invalid: empty name
					Industry:     "finance",
					Country:      "US",
				},
				{
					BusinessName:      "Another Valid Company",
					BusinessAddress:   "789 Another Street, Test City, TC 12345",
					Industry:          "healthcare",
					Country:           "US",
					PredictionHorizon: 6,
					ModelType:         "ensemble",
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *models.BatchRiskAssessmentResponse) {
				assert.NotEmpty(t, result.BatchID)
				assert.Equal(t, models.StatusCompleted, result.Status)
				assert.Len(t, result.Assessments, 2) // Only valid assessments
				assert.Equal(t, 3, result.TotalCount)
				assert.Equal(t, 2, result.SuccessCount)
				assert.Equal(t, 1, result.ErrorCount)
				assert.NotEmpty(t, result.Errors)
				assert.Len(t, result.Errors, 1)
			},
		},
		{
			name:          "empty batch request",
			requests:      []*models.RiskAssessmentRequest{},
			expectedError: true,
		},
		{
			name: "batch request with all invalid",
			requests: []*models.RiskAssessmentRequest{
				{
					BusinessName: "", // Invalid: empty name
					Industry:     "technology",
					Country:      "US",
				},
				{
					BusinessName: "", // Invalid: empty name
					Industry:     "finance",
					Country:      "US",
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *models.BatchRiskAssessmentResponse) {
				assert.NotEmpty(t, result.BatchID)
				assert.Equal(t, models.StatusCompleted, result.Status)
				assert.Len(t, result.Assessments, 0) // No valid assessments
				assert.Equal(t, 2, result.TotalCount)
				assert.Equal(t, 0, result.SuccessCount)
				assert.Equal(t, 2, result.ErrorCount)
				assert.NotEmpty(t, result.Errors)
				assert.Len(t, result.Errors, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test batch risk assessment
			result, err := mlService.service.AssessRiskBatch(ctx, tt.requests)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestMLService_ModelTraining_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test ML service
	mlService := SetupTestMLService(t)
	defer mlService.TeardownTestMLService()

	ctx := context.Background()

	tests := []struct {
		name           string
		trainingData   *models.TrainingData
		expectedError  bool
		validateResult func(*testing.T, *models.TrainingResult)
	}{
		{
			name: "valid training data",
			trainingData: &models.TrainingData{
				ModelType:         "xgboost",
				TrainingSamples:   generateMockTrainingSamples(100),
				ValidationSamples: generateMockTrainingSamples(20),
				TestSamples:       generateMockTrainingSamples(20),
				Features:          []string{"business_name_length", "industry_risk", "country_risk"},
				Target:            "risk_score",
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *models.TrainingResult) {
				assert.NotEmpty(t, result.ModelID)
				assert.Equal(t, "xgboost", result.ModelType)
				assert.Equal(t, models.StatusCompleted, result.Status)
				assert.GreaterOrEqual(t, result.Accuracy, 0.0)
				assert.LessOrEqual(t, result.Accuracy, 1.0)
				assert.GreaterOrEqual(t, result.Precision, 0.0)
				assert.LessOrEqual(t, result.Precision, 1.0)
				assert.GreaterOrEqual(t, result.Recall, 0.0)
				assert.LessOrEqual(t, result.Recall, 1.0)
				assert.GreaterOrEqual(t, result.F1Score, 0.0)
				assert.LessOrEqual(t, result.F1Score, 1.0)
				assert.NotZero(t, result.TrainingTime)
				assert.NotZero(t, result.CreatedAt)
				assert.NotZero(t, result.CompletedAt)
			},
		},
		{
			name: "training with insufficient data",
			trainingData: &models.TrainingData{
				ModelType:         "xgboost",
				TrainingSamples:   generateMockTrainingSamples(5), // Too few samples
				ValidationSamples: generateMockTrainingSamples(2),
				TestSamples:       generateMockTrainingSamples(2),
				Features:          []string{"business_name_length", "industry_risk"},
				Target:            "risk_score",
			},
			expectedError: true,
		},
		{
			name: "training with invalid model type",
			trainingData: &models.TrainingData{
				ModelType:         "invalid-model",
				TrainingSamples:   generateMockTrainingSamples(100),
				ValidationSamples: generateMockTrainingSamples(20),
				TestSamples:       generateMockTrainingSamples(20),
				Features:          []string{"business_name_length", "industry_risk"},
				Target:            "risk_score",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test model training
			result, err := mlService.service.TrainModel(ctx, tt.trainingData)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestMLService_ModelValidation_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test ML service
	mlService := SetupTestMLService(t)
	defer mlService.TeardownTestMLService()

	ctx := context.Background()

	tests := []struct {
		name           string
		modelID        string
		validationData *models.ValidationData
		expectedError  bool
		validateResult func(*testing.T, *models.ValidationResult)
	}{
		{
			name:    "valid model validation",
			modelID: "test-model-123",
			validationData: &models.ValidationData{
				TestSamples: generateMockTrainingSamples(50),
				Metrics:     []string{"accuracy", "precision", "recall", "f1_score"},
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *models.ValidationResult) {
				assert.NotEmpty(t, result.ModelID)
				assert.Equal(t, "test-model-123", result.ModelID)
				assert.Equal(t, models.StatusCompleted, result.Status)
				assert.GreaterOrEqual(t, result.Accuracy, 0.0)
				assert.LessOrEqual(t, result.Accuracy, 1.0)
				assert.GreaterOrEqual(t, result.Precision, 0.0)
				assert.LessOrEqual(t, result.Precision, 1.0)
				assert.GreaterOrEqual(t, result.Recall, 0.0)
				assert.LessOrEqual(t, result.Recall, 1.0)
				assert.GreaterOrEqual(t, result.F1Score, 0.0)
				assert.LessOrEqual(t, result.F1Score, 1.0)
				assert.NotZero(t, result.ValidationTime)
				assert.NotZero(t, result.CreatedAt)
				assert.NotZero(t, result.CompletedAt)
			},
		},
		{
			name:    "validation with non-existent model",
			modelID: "non-existent-model",
			validationData: &models.ValidationData{
				TestSamples: generateMockTrainingSamples(50),
				Metrics:     []string{"accuracy", "precision"},
			},
			expectedError: true,
		},
		{
			name:    "validation with empty test data",
			modelID: "test-model-123",
			validationData: &models.ValidationData{
				TestSamples: []*models.TrainingSample{},
				Metrics:     []string{"accuracy"},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test model validation
			result, err := mlService.service.ValidateModel(ctx, tt.modelID, tt.validationData)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestMLService_Concurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test ML service
	mlService := SetupTestMLService(t)
	defer mlService.TeardownTestMLService()

	ctx := context.Background()

	// Test concurrent risk assessments
	numGoroutines := 5
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			request := &models.RiskAssessmentRequest{
				BusinessName:      fmt.Sprintf("Concurrent Company %d", id),
				BusinessAddress:   fmt.Sprintf("123 Concurrent Street %d, Test City, TC 12345", id),
				Industry:          "technology",
				Country:           "US",
				PredictionHorizon: 3,
				ModelType:         "xgboost",
			}

			result, err := mlService.service.AssessRisk(ctx, request)
			if err != nil {
				results <- err
				return
			}

			if result == nil {
				results <- fmt.Errorf("result is nil")
				return
			}

			if result.ID == "" {
				results <- fmt.Errorf("result ID is empty")
				return
			}

			results <- nil
		}(i)
	}

	// Wait for all assessments to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err, "Concurrent assessment %d failed", i)
	}
}

func TestMLService_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test ML service
	mlService := SetupTestMLService(t)
	defer mlService.TeardownTestMLService()

	ctx := context.Background()

	tests := []struct {
		name         string
		testFunction func() error
		expectError  bool
	}{
		{
			name: "assess risk with nil request",
			testFunction: func() error {
				_, err := mlService.service.AssessRisk(ctx, nil)
				return err
			},
			expectError: true,
		},
		{
			name: "assess risk batch with nil requests",
			testFunction: func() error {
				_, err := mlService.service.AssessRiskBatch(ctx, nil)
				return err
			},
			expectError: true,
		},
		{
			name: "train model with nil training data",
			testFunction: func() error {
				_, err := mlService.service.TrainModel(ctx, nil)
				return err
			},
			expectError: true,
		},
		{
			name: "validate model with nil validation data",
			testFunction: func() error {
				_, err := mlService.service.ValidateModel(ctx, "test-model", nil)
				return err
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testFunction()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper functions
func generateMockTrainingSamples(count int) []*models.TrainingSample {
	samples := make([]*models.TrainingSample, count)

	for i := 0; i < count; i++ {
		samples[i] = &models.TrainingSample{
			Features: map[string]interface{}{
				"business_name_length": float64(10 + i%20),
				"industry_risk":        float64(0.1 + float64(i%10)*0.1),
				"country_risk":         float64(0.2 + float64(i%5)*0.1),
			},
			Target: float64(0.1 + float64(i%10)*0.1),
		}
	}

	return samples
}

// Mock models for testing
type BatchRiskAssessmentResponse struct {
	BatchID      string                           `json:"batch_id"`
	Status       models.AssessmentStatus          `json:"status"`
	Assessments  []*models.RiskAssessmentResponse `json:"assessments"`
	TotalCount   int                              `json:"total_count"`
	SuccessCount int                              `json:"success_count"`
	ErrorCount   int                              `json:"error_count"`
	Errors       []string                         `json:"errors"`
	CreatedAt    time.Time                        `json:"created_at"`
	CompletedAt  time.Time                        `json:"completed_at"`
}

type TrainingData struct {
	ModelType         string            `json:"model_type"`
	TrainingSamples   []*TrainingSample `json:"training_samples"`
	ValidationSamples []*TrainingSample `json:"validation_samples"`
	TestSamples       []*TrainingSample `json:"test_samples"`
	Features          []string          `json:"features"`
	Target            string            `json:"target"`
}

type TrainingSample struct {
	Features map[string]interface{} `json:"features"`
	Target   float64                `json:"target"`
}

type TrainingResult struct {
	ModelID      string                  `json:"model_id"`
	ModelType    string                  `json:"model_type"`
	Status       models.AssessmentStatus `json:"status"`
	Accuracy     float64                 `json:"accuracy"`
	Precision    float64                 `json:"precision"`
	Recall       float64                 `json:"recall"`
	F1Score      float64                 `json:"f1_score"`
	TrainingTime time.Duration           `json:"training_time"`
	CreatedAt    time.Time               `json:"created_at"`
	CompletedAt  time.Time               `json:"completed_at"`
}

type ValidationData struct {
	TestSamples []*TrainingSample `json:"test_samples"`
	Metrics     []string          `json:"metrics"`
}

type ValidationResult struct {
	ModelID        string                  `json:"model_id"`
	Status         models.AssessmentStatus `json:"status"`
	Accuracy       float64                 `json:"accuracy"`
	Precision      float64                 `json:"precision"`
	Recall         float64                 `json:"recall"`
	F1Score        float64                 `json:"f1_score"`
	ValidationTime time.Duration           `json:"validation_time"`
	CreatedAt      time.Time               `json:"created_at"`
	CompletedAt    time.Time               `json:"completed_at"`
}

// Benchmark tests
func BenchmarkMLService_AssessRisk(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping integration benchmark")
	}

	// Setup test ML service
	mlService := SetupTestMLService(&testing.T{})
	defer mlService.TeardownTestMLService()

	ctx := context.Background()

	request := &models.RiskAssessmentRequest{
		BusinessName:      "Benchmark Company",
		BusinessAddress:   "123 Benchmark Street, Test City, TC 12345",
		Industry:          "technology",
		Country:           "US",
		PredictionHorizon: 3,
		ModelType:         "xgboost",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mlService.service.AssessRisk(ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMLService_AssessRiskBatch(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping integration benchmark")
	}

	// Setup test ML service
	mlService := SetupTestMLService(&testing.T{})
	defer mlService.TeardownTestMLService()

	ctx := context.Background()

	requests := []*models.RiskAssessmentRequest{
		{
			BusinessName:      "Batch Company 1",
			BusinessAddress:   "123 Batch Street 1, Test City, TC 12345",
			Industry:          "technology",
			Country:           "US",
			PredictionHorizon: 3,
			ModelType:         "xgboost",
		},
		{
			BusinessName:      "Batch Company 2",
			BusinessAddress:   "456 Batch Street 2, Test City, TC 12345",
			Industry:          "finance",
			Country:           "US",
			PredictionHorizon: 6,
			ModelType:         "ensemble",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mlService.service.AssessRiskBatch(ctx, requests)
		if err != nil {
			b.Fatal(err)
		}
	}
}
