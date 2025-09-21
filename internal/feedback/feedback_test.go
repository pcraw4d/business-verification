package feedback

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

// MockFeedbackStorage implements FeedbackStorage for testing
type MockFeedbackStorage struct {
	feedback []*UserFeedback
	stats    *FeedbackStats
}

func (m *MockFeedbackStorage) StoreFeedback(ctx context.Context, feedback *UserFeedback) error {
	m.feedback = append(m.feedback, feedback)
	return nil
}

func (m *MockFeedbackStorage) GetFeedbackByCategory(ctx context.Context, category string) ([]*UserFeedback, error) {
	var result []*UserFeedback
	for _, f := range m.feedback {
		if string(f.Category) == category {
			result = append(result, f)
		}
	}
	return result, nil
}

func (m *MockFeedbackStorage) GetFeedbackByTimeRange(ctx context.Context, start, end time.Time) ([]*UserFeedback, error) {
	var result []*UserFeedback
	for _, f := range m.feedback {
		if f.SubmittedAt.After(start) && f.SubmittedAt.Before(end) {
			result = append(result, f)
		}
	}
	return result, nil
}

func (m *MockFeedbackStorage) GetFeedbackStats(ctx context.Context) (*FeedbackStats, error) {
	return m.stats, nil
}

func TestUserFeedbackCollector_CollectFeedback(t *testing.T) {
	// Setup
	mockStorage := &MockFeedbackStorage{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	collector := NewUserFeedbackCollector(mockStorage, logger)

	// Test data
	feedback := &UserFeedback{
		UserID:                 "test-user-123",
		Category:               CategoryDatabasePerformance,
		Rating:                 4,
		Comments:               "Great improvement in performance!",
		SpecificFeatures:       []string{"query_speed", "response_time"},
		ImprovementAreas:       []string{"caching", "indexing"},
		ClassificationAccuracy: 0.95,
		PerformanceRating:      4,
		UsabilityRating:        5,
		BusinessImpact: BusinessImpactRating{
			TimeSaved:        30,
			CostReduction:    "25%",
			ErrorReduction:   40,
			ProductivityGain: 35,
			ROI:              "High",
		},
	}

	// Test
	ctx := context.Background()
	err := collector.CollectFeedback(ctx, feedback)

	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if feedback.ID.String() == "" {
		t.Error("Expected feedback ID to be set")
	}

	if feedback.SubmittedAt.IsZero() {
		t.Error("Expected submitted timestamp to be set")
	}

	if len(mockStorage.feedback) != 1 {
		t.Errorf("Expected 1 feedback item, got %d", len(mockStorage.feedback))
	}
}

func TestUserFeedbackCollector_GetFeedbackAnalysis(t *testing.T) {
	// Setup
	mockStorage := &MockFeedbackStorage{
		feedback: []*UserFeedback{
			{
				ID:                     uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
				UserID:                 "user-1",
				Category:               CategoryDatabasePerformance,
				Rating:                 4,
				PerformanceRating:      4,
				UsabilityRating:        5,
				ClassificationAccuracy: 0.9,
				ImprovementAreas:       []string{"caching", "indexing"},
				SpecificFeatures:       []string{"query_speed"},
				BusinessImpact: BusinessImpactRating{
					TimeSaved:        30,
					ErrorReduction:   40,
					ProductivityGain: 35,
				},
			},
			{
				ID:                     uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
				UserID:                 "user-2",
				Category:               CategoryDatabasePerformance,
				Rating:                 5,
				PerformanceRating:      5,
				UsabilityRating:        4,
				ClassificationAccuracy: 0.95,
				ImprovementAreas:       []string{"caching"},
				SpecificFeatures:       []string{"response_time"},
				BusinessImpact: BusinessImpactRating{
					TimeSaved:        45,
					ErrorReduction:   50,
					ProductivityGain: 40,
				},
			},
		},
	}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	collector := NewUserFeedbackCollector(mockStorage, logger)

	// Test
	ctx := context.Background()
	analysis, err := collector.GetFeedbackAnalysis(ctx, CategoryDatabasePerformance)

	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if analysis == nil {
		t.Fatal("Expected analysis to be returned")
	}

	if analysis.TotalResponses != 2 {
		t.Errorf("Expected 2 total responses, got %d", analysis.TotalResponses)
	}

	if analysis.AverageRating != 4.5 {
		t.Errorf("Expected average rating of 4.5, got %f", analysis.AverageRating)
	}

	if analysis.AveragePerformance != 4.5 {
		t.Errorf("Expected average performance of 4.5, got %f", analysis.AveragePerformance)
	}

	if analysis.AverageUsability != 4.5 {
		t.Errorf("Expected average usability of 4.5, got %f", analysis.AverageUsability)
	}

	if analysis.AverageAccuracy != 0.925 {
		t.Errorf("Expected average accuracy of 0.925, got %f", analysis.AverageAccuracy)
	}

	if len(analysis.TopImprovements) == 0 {
		t.Error("Expected top improvements to be populated")
	}

	if len(analysis.TopFeatures) == 0 {
		t.Error("Expected top features to be populated")
	}

	if len(analysis.Recommendations) == 0 {
		t.Error("Expected recommendations to be generated")
	}
}

func TestUserFeedbackCollector_ValidateFeedback(t *testing.T) {
	// Setup
	mockStorage := &MockFeedbackStorage{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	collector := NewUserFeedbackCollector(mockStorage, logger)

	tests := []struct {
		name      string
		feedback  *UserFeedback
		expectErr bool
	}{
		{
			name: "valid feedback",
			feedback: &UserFeedback{
				UserID:                 "test-user",
				Category:               CategoryDatabasePerformance,
				Rating:                 4,
				PerformanceRating:      4,
				UsabilityRating:        5,
				ClassificationAccuracy: 0.9,
			},
			expectErr: false,
		},
		{
			name: "missing user ID",
			feedback: &UserFeedback{
				Category:               CategoryDatabasePerformance,
				Rating:                 4,
				PerformanceRating:      4,
				UsabilityRating:        5,
				ClassificationAccuracy: 0.9,
			},
			expectErr: true,
		},
		{
			name: "invalid rating",
			feedback: &UserFeedback{
				UserID:                 "test-user",
				Category:               CategoryDatabasePerformance,
				Rating:                 6, // Invalid: > 5
				PerformanceRating:      4,
				UsabilityRating:        5,
				ClassificationAccuracy: 0.9,
			},
			expectErr: true,
		},
		{
			name: "invalid classification accuracy",
			feedback: &UserFeedback{
				UserID:                 "test-user",
				Category:               CategoryDatabasePerformance,
				Rating:                 4,
				PerformanceRating:      4,
				UsabilityRating:        5,
				ClassificationAccuracy: 1.5, // Invalid: > 1
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := collector.CollectFeedback(ctx, tt.feedback)

			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
		})
	}
}

func TestUserFeedbackCollector_ExportFeedback(t *testing.T) {
	// Setup
	mockStorage := &MockFeedbackStorage{
		feedback: []*UserFeedback{
			{
				ID:                     uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
				UserID:                 "user-1",
				Category:               CategoryDatabasePerformance,
				Rating:                 4,
				Comments:               "Great performance!",
				PerformanceRating:      4,
				UsabilityRating:        5,
				ClassificationAccuracy: 0.9,
				SubmittedAt:            time.Now(),
			},
		},
	}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	collector := NewUserFeedbackCollector(mockStorage, logger)

	// Test JSON export
	ctx := context.Background()
	jsonData, err := collector.ExportFeedback(ctx, "json", CategoryDatabasePerformance)
	if err != nil {
		t.Fatalf("Expected no error for JSON export, got %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Expected JSON data to be returned")
	}

	// Test CSV export
	csvData, err := collector.ExportFeedback(ctx, "csv", CategoryDatabasePerformance)
	if err != nil {
		t.Fatalf("Expected no error for CSV export, got %v", err)
	}

	if len(csvData) == 0 {
		t.Error("Expected CSV data to be returned")
	}

	// Test invalid format
	_, err = collector.ExportFeedback(ctx, "invalid", CategoryDatabasePerformance)
	if err == nil {
		t.Error("Expected error for invalid export format")
	}
}
