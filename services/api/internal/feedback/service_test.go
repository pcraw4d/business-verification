package feedback

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

// MockFeedbackRepository is a mock implementation of FeedbackRepository
type MockFeedbackRepository struct {
	userFeedbacks               map[string]UserFeedback
	mlModelFeedbacks            map[string]MLModelFeedback
	securityValidationFeedbacks map[string]SecurityValidationFeedback
	feedbackTrends              []FeedbackTrend
}

func NewMockFeedbackRepository() *MockFeedbackRepository {
	return &MockFeedbackRepository{
		userFeedbacks:               make(map[string]UserFeedback),
		mlModelFeedbacks:            make(map[string]MLModelFeedback),
		securityValidationFeedbacks: make(map[string]SecurityValidationFeedback),
		feedbackTrends:              make([]FeedbackTrend, 0),
	}
}

func (m *MockFeedbackRepository) SaveUserFeedback(ctx context.Context, feedback UserFeedback) error {
	m.userFeedbacks[feedback.ID] = feedback
	return nil
}

func (m *MockFeedbackRepository) GetUserFeedback(ctx context.Context, id string) (*UserFeedback, error) {
	if feedback, exists := m.userFeedbacks[id]; exists {
		return &feedback, nil
	}
	return nil, nil
}

func (m *MockFeedbackRepository) GetUserFeedbackByClassificationID(ctx context.Context, classificationID string) ([]UserFeedback, error) {
	var feedbacks []UserFeedback
	for _, feedback := range m.userFeedbacks {
		if feedback.OriginalClassificationID == classificationID {
			feedbacks = append(feedbacks, feedback)
		}
	}
	return feedbacks, nil
}

func (m *MockFeedbackRepository) UpdateUserFeedbackStatus(ctx context.Context, id string, status FeedbackStatus) error {
	if feedback, exists := m.userFeedbacks[id]; exists {
		feedback.Status = status
		now := time.Now()
		feedback.ProcessedAt = &now
		m.userFeedbacks[id] = feedback
	}
	return nil
}

func (m *MockFeedbackRepository) SaveMLModelFeedback(ctx context.Context, feedback MLModelFeedback) error {
	m.mlModelFeedbacks[feedback.ID] = feedback
	return nil
}

func (m *MockFeedbackRepository) GetMLModelFeedback(ctx context.Context, id string) (*MLModelFeedback, error) {
	if feedback, exists := m.mlModelFeedbacks[id]; exists {
		return &feedback, nil
	}
	return nil, nil
}

func (m *MockFeedbackRepository) GetMLModelFeedbackByModelVersion(ctx context.Context, modelVersion string) ([]MLModelFeedback, error) {
	var feedbacks []MLModelFeedback
	for _, feedback := range m.mlModelFeedbacks {
		if feedback.ModelVersionID == modelVersion {
			feedbacks = append(feedbacks, feedback)
		}
	}
	return feedbacks, nil
}

func (m *MockFeedbackRepository) UpdateMLModelFeedbackStatus(ctx context.Context, id string, status FeedbackStatus) error {
	if feedback, exists := m.mlModelFeedbacks[id]; exists {
		feedback.Status = status
		now := time.Now()
		feedback.ProcessedAt = &now
		m.mlModelFeedbacks[id] = feedback
	}
	return nil
}

func (m *MockFeedbackRepository) SaveSecurityValidationFeedback(ctx context.Context, feedback SecurityValidationFeedback) error {
	m.securityValidationFeedbacks[feedback.ID] = feedback
	return nil
}

func (m *MockFeedbackRepository) GetSecurityValidationFeedback(ctx context.Context, id string) (*SecurityValidationFeedback, error) {
	if feedback, exists := m.securityValidationFeedbacks[id]; exists {
		return &feedback, nil
	}
	return nil, nil
}

func (m *MockFeedbackRepository) GetSecurityValidationFeedbackByType(ctx context.Context, validationType string) ([]SecurityValidationFeedback, error) {
	var feedbacks []SecurityValidationFeedback
	for _, feedback := range m.securityValidationFeedbacks {
		if feedback.ValidationType == validationType {
			feedbacks = append(feedbacks, feedback)
		}
	}
	return feedbacks, nil
}

func (m *MockFeedbackRepository) UpdateSecurityValidationFeedbackStatus(ctx context.Context, id string, status FeedbackStatus) error {
	if feedback, exists := m.securityValidationFeedbacks[id]; exists {
		feedback.Status = status
		now := time.Now()
		feedback.ProcessedAt = &now
		m.securityValidationFeedbacks[id] = feedback
	}
	return nil
}

func (m *MockFeedbackRepository) GetFeedbackTrends(ctx context.Context, request FeedbackAnalysisRequest) ([]FeedbackTrend, error) {
	return m.feedbackTrends, nil
}

func (m *MockFeedbackRepository) GetFeedbackStatistics(ctx context.Context, request FeedbackAnalysisRequest) (*FeedbackAnalysisResponse, error) {
	return &FeedbackAnalysisResponse{
		Trends:                m.feedbackTrends,
		TotalFeedback:         len(m.userFeedbacks) + len(m.mlModelFeedbacks) + len(m.securityValidationFeedbacks),
		AverageAccuracy:       0.85,
		AverageConfidence:     0.90,
		ErrorRate:             0.05,
		SecurityViolationRate: 0.02,
		AnalysisTime:          time.Now(),
	}, nil
}

func (m *MockFeedbackRepository) SaveBatchUserFeedback(ctx context.Context, feedbacks []UserFeedback) error {
	for _, feedback := range feedbacks {
		m.userFeedbacks[feedback.ID] = feedback
	}
	return nil
}

func (m *MockFeedbackRepository) SaveBatchMLModelFeedback(ctx context.Context, feedbacks []MLModelFeedback) error {
	for _, feedback := range feedbacks {
		m.mlModelFeedbacks[feedback.ID] = feedback
	}
	return nil
}

func (m *MockFeedbackRepository) SaveBatchSecurityValidationFeedback(ctx context.Context, feedbacks []SecurityValidationFeedback) error {
	for _, feedback := range feedbacks {
		m.securityValidationFeedbacks[feedback.ID] = feedback
	}
	return nil
}

// MockFeedbackAnalyzer is a mock implementation of FeedbackAnalyzer
type MockFeedbackAnalyzer struct{}

func (m *MockFeedbackAnalyzer) AnalyzeFeedbackTrends(ctx context.Context, request FeedbackAnalysisRequest) (*FeedbackAnalysisResponse, error) {
	return &FeedbackAnalysisResponse{
		Trends:                []FeedbackTrend{},
		TotalFeedback:         10,
		AverageAccuracy:       0.85,
		AverageConfidence:     0.90,
		ErrorRate:             0.05,
		SecurityViolationRate: 0.02,
		AnalysisTime:          time.Now(),
	}, nil
}

func (m *MockFeedbackAnalyzer) IdentifyFeedbackPatterns(ctx context.Context, method ClassificationMethod, timeWindow string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"method":      string(method),
		"time_window": timeWindow,
		"patterns":    []string{"pattern1", "pattern2"},
	}, nil
}

func (m *MockFeedbackAnalyzer) CalculateMethodPerformance(ctx context.Context, method ClassificationMethod) (map[string]float64, error) {
	return map[string]float64{
		"accuracy":   0.85,
		"confidence": 0.90,
		"error_rate": 0.05,
	}, nil
}

func (m *MockFeedbackAnalyzer) DetectAnomalies(ctx context.Context, method ClassificationMethod) ([]string, error) {
	return []string{"anomaly1", "anomaly2"}, nil
}

// MockFeedbackProcessor is a mock implementation of FeedbackProcessor
type MockFeedbackProcessor struct{}

func (m *MockFeedbackProcessor) ProcessUserFeedback(ctx context.Context, feedback UserFeedback) error {
	return nil
}

func (m *MockFeedbackProcessor) ProcessMLModelFeedback(ctx context.Context, feedback MLModelFeedback) error {
	return nil
}

func (m *MockFeedbackProcessor) ProcessSecurityValidationFeedback(ctx context.Context, feedback SecurityValidationFeedback) error {
	return nil
}

func (m *MockFeedbackProcessor) ApplyFeedbackCorrections(ctx context.Context, feedbackID string) error {
	return nil
}

func (m *MockFeedbackProcessor) GenerateFeedbackInsights(ctx context.Context, method ClassificationMethod) (map[string]interface{}, error) {
	return map[string]interface{}{
		"method":   string(method),
		"insights": []string{"insight1", "insight2"},
	}, nil
}

// MockModelVersionManager is a mock implementation of ModelVersionManager
type MockModelVersionManager struct{}

func (m *MockModelVersionManager) GetCurrentModelVersion(ctx context.Context, modelType string) (string, error) {
	return "v1.0.0", nil
}

func (m *MockModelVersionManager) GetModelVersionHistory(ctx context.Context, modelType string) ([]string, error) {
	return []string{"v1.0.0", "v0.9.0"}, nil
}

func (m *MockModelVersionManager) RegisterModelVersion(ctx context.Context, modelType, version string, metadata map[string]interface{}) error {
	return nil
}

func (m *MockModelVersionManager) GetModelVersionMetadata(ctx context.Context, modelType, version string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"version": version,
		"type":    modelType,
	}, nil
}

// MockSecurityValidator is a mock implementation of SecurityValidator
type MockSecurityValidator struct{}

func (m *MockSecurityValidator) ValidateFeedbackSource(ctx context.Context, source string) error {
	return nil
}

func (m *MockSecurityValidator) ValidateFeedbackContent(ctx context.Context, content map[string]interface{}) error {
	return nil
}

func (m *MockSecurityValidator) CheckSecurityViolations(ctx context.Context, feedback interface{}) ([]string, error) {
	return []string{}, nil
}

func (m *MockSecurityValidator) ValidateUserPermissions(ctx context.Context, userID string, feedbackType FeedbackType) error {
	return nil
}

// Test functions

func TestFeedbackService_CollectUserFeedback(t *testing.T) {
	logger := zaptest.NewLogger(t)
	repository := NewMockFeedbackRepository()
	validator := NewFeedbackValidator(&MockSecurityValidator{})
	analyzer := &MockFeedbackAnalyzer{}
	processor := &MockFeedbackProcessor{}
	modelVersionManager := &MockModelVersionManager{}
	securityValidator := &MockSecurityValidator{}

	service := NewFeedbackService(
		repository,
		validator,
		analyzer,
		processor,
		modelVersionManager,
		securityValidator,
		logger,
	)

	tests := []struct {
		name    string
		request FeedbackCollectionRequest
		wantErr bool
	}{
		{
			name: "valid user feedback",
			request: FeedbackCollectionRequest{
				UserID:                   "user123",
				BusinessName:             "Test Business",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
				FeedbackValue:            map[string]interface{}{"accuracy": 0.9},
				FeedbackText:             "This classification was accurate",
				ConfidenceScore:          0.9,
				Metadata:                 map[string]interface{}{"source": "web"},
			},
			wantErr: false,
		},
		{
			name: "invalid user ID",
			request: FeedbackCollectionRequest{
				UserID:                   "",
				BusinessName:             "Test Business",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
			},
			wantErr: true,
		},
		{
			name: "invalid business name",
			request: FeedbackCollectionRequest{
				UserID:                   "user123",
				BusinessName:             "",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
			},
			wantErr: true,
		},
		{
			name: "invalid confidence score",
			request: FeedbackCollectionRequest{
				UserID:                   "user123",
				BusinessName:             "Test Business",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
				ConfidenceScore:          1.5, // Invalid: > 1.0
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			response, err := service.CollectUserFeedback(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CollectUserFeedback() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("CollectUserFeedback() unexpected error: %v", err)
				return
			}

			if response == nil {
				t.Errorf("CollectUserFeedback() expected response but got nil")
				return
			}

			if response.ID == "" {
				t.Errorf("CollectUserFeedback() expected non-empty ID")
			}

			if response.Status != "pending" {
				t.Errorf("CollectUserFeedback() expected status 'pending', got %s", response.Status)
			}
		})
	}
}

func TestFeedbackService_CollectMLModelFeedback(t *testing.T) {
	logger := zaptest.NewLogger(t)
	repository := NewMockFeedbackRepository()
	validator := NewFeedbackValidator(&MockSecurityValidator{})
	analyzer := &MockFeedbackAnalyzer{}
	processor := &MockFeedbackProcessor{}
	modelVersionManager := &MockModelVersionManager{}
	securityValidator := &MockSecurityValidator{}

	service := NewFeedbackService(
		repository,
		validator,
		analyzer,
		processor,
		modelVersionManager,
		securityValidator,
		logger,
	)

	tests := []struct {
		name     string
		feedback MLModelFeedback
		wantErr  bool
	}{
		{
			name: "valid ML model feedback",
			feedback: MLModelFeedback{
				ID:                   "feedback123",
				ModelVersionID:       "model_v1.0.0",
				ModelType:            "bert",
				ClassificationMethod: MethodML,
				PredictionID:         "prediction123",
				ActualResult:         map[string]interface{}{"industry": "technology"},
				PredictedResult:      map[string]interface{}{"industry": "technology"},
				AccuracyScore:        0.95,
				ConfidenceScore:      0.90,
				ProcessingTimeMs:     100,
				Status:               FeedbackStatusPending,
				CreatedAt:            time.Now(),
				Metadata:             map[string]interface{}{"model_version": "v1.0.0"},
			},
			wantErr: false,
		},
		{
			name: "invalid model version ID",
			feedback: MLModelFeedback{
				ID:                   "feedback123",
				ModelVersionID:       "", // Invalid: empty
				ModelType:            "bert",
				ClassificationMethod: MethodML,
				PredictionID:         "prediction123",
				ActualResult:         map[string]interface{}{"industry": "technology"},
				PredictedResult:      map[string]interface{}{"industry": "technology"},
			},
			wantErr: true,
		},
		{
			name: "invalid accuracy score",
			feedback: MLModelFeedback{
				ID:                   "feedback123",
				ModelVersionID:       "model_v1.0.0",
				ModelType:            "bert",
				ClassificationMethod: MethodML,
				PredictionID:         "prediction123",
				ActualResult:         map[string]interface{}{"industry": "technology"},
				PredictedResult:      map[string]interface{}{"industry": "technology"},
				AccuracyScore:        1.5, // Invalid: > 1.0
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := service.CollectMLModelFeedback(ctx, tt.feedback)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CollectMLModelFeedback() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("CollectMLModelFeedback() unexpected error: %v", err)
			}
		})
	}
}

func TestFeedbackService_CollectSecurityValidationFeedback(t *testing.T) {
	logger := zaptest.NewLogger(t)
	repository := NewMockFeedbackRepository()
	validator := NewFeedbackValidator(&MockSecurityValidator{})
	analyzer := &MockFeedbackAnalyzer{}
	processor := &MockFeedbackProcessor{}
	modelVersionManager := &MockModelVersionManager{}
	securityValidator := &MockSecurityValidator{}

	service := NewFeedbackService(
		repository,
		validator,
		analyzer,
		processor,
		modelVersionManager,
		securityValidator,
		logger,
	)

	tests := []struct {
		name     string
		feedback SecurityValidationFeedback
		wantErr  bool
	}{
		{
			name: "valid security validation feedback",
			feedback: SecurityValidationFeedback{
				ID:                 "feedback123",
				ValidationType:     "website_verification",
				DataSourceType:     "domain_analysis",
				WebsiteURL:         "https://example.com",
				ValidationResult:   map[string]interface{}{"verified": true},
				TrustScore:         0.95,
				VerificationStatus: "verified",
				SecurityViolations: []string{},
				ProcessingTimeMs:   50,
				Status:             FeedbackStatusPending,
				CreatedAt:          time.Now(),
				Metadata:           map[string]interface{}{"source": "whois"},
			},
			wantErr: false,
		},
		{
			name: "invalid validation type",
			feedback: SecurityValidationFeedback{
				ID:               "feedback123",
				ValidationType:   "", // Invalid: empty
				DataSourceType:   "domain_analysis",
				ValidationResult: map[string]interface{}{"verified": true},
			},
			wantErr: true,
		},
		{
			name: "invalid trust score",
			feedback: SecurityValidationFeedback{
				ID:               "feedback123",
				ValidationType:   "website_verification",
				DataSourceType:   "domain_analysis",
				ValidationResult: map[string]interface{}{"verified": true},
				TrustScore:       1.5, // Invalid: > 1.0
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := service.CollectSecurityValidationFeedback(ctx, tt.feedback)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CollectSecurityValidationFeedback() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("CollectSecurityValidationFeedback() unexpected error: %v", err)
			}
		})
	}
}

func TestFeedbackService_CollectBatchFeedback(t *testing.T) {
	logger := zaptest.NewLogger(t)
	repository := NewMockFeedbackRepository()
	validator := NewFeedbackValidator(&MockSecurityValidator{})
	analyzer := &MockFeedbackAnalyzer{}
	processor := &MockFeedbackProcessor{}
	modelVersionManager := &MockModelVersionManager{}
	securityValidator := &MockSecurityValidator{}

	service := NewFeedbackService(
		repository,
		validator,
		analyzer,
		processor,
		modelVersionManager,
		securityValidator,
		logger,
	)

	requests := []FeedbackCollectionRequest{
		{
			UserID:                   "user1",
			BusinessName:             "Business 1",
			OriginalClassificationID: "classification1",
			FeedbackType:             FeedbackTypeAccuracy,
			FeedbackValue:            map[string]interface{}{"accuracy": 0.9},
		},
		{
			UserID:                   "user2",
			BusinessName:             "Business 2",
			OriginalClassificationID: "classification2",
			FeedbackType:             FeedbackTypeRelevance,
			FeedbackValue:            map[string]interface{}{"relevance": 0.8},
		},
	}

	ctx := context.Background()
	responses, err := service.CollectBatchFeedback(ctx, requests)

	if err != nil {
		t.Errorf("CollectBatchFeedback() unexpected error: %v", err)
		return
	}

	if len(responses) != len(requests) {
		t.Errorf("CollectBatchFeedback() expected %d responses, got %d", len(requests), len(responses))
	}

	for i, response := range responses {
		if response.ID == "" {
			t.Errorf("CollectBatchFeedback() response %d has empty ID", i)
		}
		if response.Status != "pending" {
			t.Errorf("CollectBatchFeedback() response %d expected status 'pending', got %s", i, response.Status)
		}
	}
}

func TestFeedbackService_AnalyzeFeedbackTrends(t *testing.T) {
	logger := zaptest.NewLogger(t)
	repository := NewMockFeedbackRepository()
	validator := NewFeedbackValidator(&MockSecurityValidator{})
	analyzer := &MockFeedbackAnalyzer{}
	processor := &MockFeedbackProcessor{}
	modelVersionManager := &MockModelVersionManager{}
	securityValidator := &MockSecurityValidator{}

	service := NewFeedbackService(
		repository,
		validator,
		analyzer,
		processor,
		modelVersionManager,
		securityValidator,
		logger,
	)

	request := FeedbackAnalysisRequest{
		Method:       MethodEnsemble,
		TimeWindow:   "daily",
		FeedbackType: FeedbackTypeAccuracy,
	}

	ctx := context.Background()
	response, err := service.AnalyzeFeedbackTrends(ctx, request)

	if err != nil {
		t.Errorf("AnalyzeFeedbackTrends() unexpected error: %v", err)
		return
	}

	if response == nil {
		t.Errorf("AnalyzeFeedbackTrends() expected response but got nil")
		return
	}

	if response.TotalFeedback == 0 {
		t.Errorf("AnalyzeFeedbackTrends() expected non-zero total feedback")
	}
}

func TestFeedbackService_GetServiceHealth(t *testing.T) {
	logger := zaptest.NewLogger(t)
	repository := NewMockFeedbackRepository()
	validator := NewFeedbackValidator(&MockSecurityValidator{})
	analyzer := &MockFeedbackAnalyzer{}
	processor := &MockFeedbackProcessor{}
	modelVersionManager := &MockModelVersionManager{}
	securityValidator := &MockSecurityValidator{}

	service := NewFeedbackService(
		repository,
		validator,
		analyzer,
		processor,
		modelVersionManager,
		securityValidator,
		logger,
	)

	ctx := context.Background()
	health, err := service.GetServiceHealth(ctx)

	if err != nil {
		t.Errorf("GetServiceHealth() unexpected error: %v", err)
		return
	}

	if health == nil {
		t.Errorf("GetServiceHealth() expected health data but got nil")
		return
	}

	if status, ok := health["status"]; !ok || status != "healthy" {
		t.Errorf("GetServiceHealth() expected status 'healthy', got %v", status)
	}
}

func TestFeedbackService_GetServiceMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	repository := NewMockFeedbackRepository()
	validator := NewFeedbackValidator(&MockSecurityValidator{})
	analyzer := &MockFeedbackAnalyzer{}
	processor := &MockFeedbackProcessor{}
	modelVersionManager := &MockModelVersionManager{}
	securityValidator := &MockSecurityValidator{}

	service := NewFeedbackService(
		repository,
		validator,
		analyzer,
		processor,
		modelVersionManager,
		securityValidator,
		logger,
	)

	ctx := context.Background()
	metrics, err := service.GetServiceMetrics(ctx)

	if err != nil {
		t.Errorf("GetServiceMetrics() unexpected error: %v", err)
		return
	}

	if metrics == nil {
		t.Errorf("GetServiceMetrics() expected metrics data but got nil")
		return
	}

	if service, ok := metrics["service"]; !ok || service != "feedback_collection" {
		t.Errorf("GetServiceMetrics() expected service 'feedback_collection', got %v", service)
	}
}
