package services

import (
	"context"
	"errors"
	"log"
	"testing"
	"time"

	"kyb-platform/internal/database"
	"kyb-platform/internal/models"
)

// mockRiskAssessmentRepository is a mock implementation of RiskAssessmentRepository
type mockRiskAssessmentRepository struct {
	assessments []*models.RiskAssessment
	assessment  *models.RiskAssessment
	err         error
}

func (m *mockRiskAssessmentRepository) CreateAssessment(ctx context.Context, merchantID string, options models.AssessmentOptions) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return "assessment-123", nil
}

func (m *mockRiskAssessmentRepository) GetAssessmentByID(ctx context.Context, assessmentID string) (*models.RiskAssessment, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.assessment, nil
}

func (m *mockRiskAssessmentRepository) GetAssessmentsByMerchantID(ctx context.Context, merchantID string) ([]*models.RiskAssessment, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.assessments, nil
}

func (m *mockRiskAssessmentRepository) UpdateAssessmentStatus(ctx context.Context, assessmentID string, status models.AssessmentStatus, progress int) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *mockRiskAssessmentRepository) UpdateAssessmentResult(ctx context.Context, assessmentID string, result *models.RiskAssessmentResult) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func TestRiskAssessmentService_GetRiskHistory(t *testing.T) {
	tests := []struct {
		name        string
		merchantID  string
		limit       int
		offset      int
		assessments []*models.RiskAssessment
		wantErr     bool
		wantCount   int
	}{
		{
			name:       "successful fetch with pagination",
			merchantID: "merchant-123",
			limit:      10,
			offset:     0,
			assessments: []*models.RiskAssessment{
				{ID: "assessment-1", MerchantID: "merchant-123"},
				{ID: "assessment-2", MerchantID: "merchant-123"},
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:       "pagination with offset",
			merchantID: "merchant-123",
			limit:      10,
			offset:     5,
			assessments: []*models.RiskAssessment{
				{ID: "assessment-1", MerchantID: "merchant-123"},
				{ID: "assessment-2", MerchantID: "merchant-123"},
				{ID: "assessment-3", MerchantID: "merchant-123"},
				{ID: "assessment-4", MerchantID: "merchant-123"},
				{ID: "assessment-5", MerchantID: "merchant-123"},
				{ID: "assessment-6", MerchantID: "merchant-123"},
				{ID: "assessment-7", MerchantID: "merchant-123"},
			},
			wantErr:   false,
			wantCount: 2, // offset 5, limit 10, so should return 2 items
		},
		{
			name:        "empty history",
			merchantID:  "merchant-123",
			limit:       10,
			offset:      0,
			assessments: []*models.RiskAssessment{},
			wantErr:     false,
			wantCount:   0,
		},
		{
			name:       "repository error",
			merchantID: "merchant-123",
			limit:      10,
			offset:     0,
			assessments: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			logger := log.New(log.Writer(), "", 0)

			repo := &mockRiskAssessmentRepository{
				assessments: tt.assessments,
				err:         nil,
			}

			if tt.wantErr {
				repo.err = errors.New("database error")
			}

			service := NewRiskAssessmentService(repo, nil, logger)

			result, err := service.GetRiskHistory(ctx, tt.merchantID, tt.limit, tt.offset)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("Expected %d assessments, got %d", tt.wantCount, len(result))
			}
		})
	}
}

func TestRiskAssessmentService_GetPredictions(t *testing.T) {
	ctx := context.Background()
	logger := log.New(log.Writer(), "", 0)

	assessments := []*models.RiskAssessment{
		{
			ID:         "assessment-1",
			MerchantID: "merchant-123",
			Result: &models.RiskAssessmentResult{
				OverallScore: 0.7,
				RiskLevel:    "medium",
			},
		},
	}

	repo := &mockRiskAssessmentRepository{
		assessments: assessments,
	}

	service := NewRiskAssessmentService(repo, nil, logger)

	result, err := service.GetPredictions(ctx, "merchant-123", []int{3, 6, 12}, true, true)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result but got nil")
	}

	// Verify structure
	if result["merchantId"] != "merchant-123" {
		t.Errorf("Expected merchantId, got %v", result["merchantId"])
	}

	predictions, ok := result["predictions"].([]map[string]interface{})
	if !ok {
		t.Fatal("Expected predictions array")
	}

	if len(predictions) != 3 {
		t.Errorf("Expected 3 predictions, got %d", len(predictions))
	}
}

func TestRiskAssessmentService_ExplainAssessment(t *testing.T) {
	ctx := context.Background()
	logger := log.New(log.Writer(), "", 0)

	assessment := &models.RiskAssessment{
		ID:         "assessment-123",
		MerchantID: "merchant-123",
		Result: &models.RiskAssessmentResult{
			OverallScore: 0.7,
			Factors: []models.RiskFactor{
				{Name: "Factor1", Score: 0.8, Weight: 0.5},
				{Name: "Factor2", Score: 0.6, Weight: 0.5},
			},
		},
	}

	repo := &mockRiskAssessmentRepository{
		assessment: assessment,
	}

	service := NewRiskAssessmentService(repo, nil, logger)

	result, err := service.ExplainAssessment(ctx, "assessment-123")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result but got nil")
	}

	if result["assessmentId"] != "assessment-123" {
		t.Errorf("Expected assessmentId, got %v", result["assessmentId"])
	}

	if result["prediction"].(float64) != 0.7 {
		t.Errorf("Expected prediction 0.7, got %v", result["prediction"])
	}

	factors, ok := result["factors"].([]interface{})
	if !ok || len(factors) != 2 {
		t.Errorf("Expected 2 factors, got %d", len(factors))
	}
}

func TestRiskAssessmentService_GetRecommendations(t *testing.T) {
	ctx := context.Background()
	logger := log.New(log.Writer(), "", 0)

	assessments := []*models.RiskAssessment{
		{
			ID:         "assessment-1",
			MerchantID: "merchant-123",
			Result: &models.RiskAssessmentResult{
				OverallScore: 0.8, // High risk
				Factors: []models.RiskFactor{
					{Name: "Factor1", Score: 0.7, Weight: 0.5},
				},
			},
		},
	}

	repo := &mockRiskAssessmentRepository{
		assessments: assessments,
	}

	service := NewRiskAssessmentService(repo, nil, logger)

	result, err := service.GetRecommendations(ctx, "merchant-123")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result but got nil")
	}

	if len(result) == 0 {
		t.Error("Expected at least one recommendation")
	}

	// Verify recommendation structure
	rec := result[0]
	if rec["merchantId"] != "merchant-123" {
		t.Errorf("Expected merchantId in recommendation")
	}
}

