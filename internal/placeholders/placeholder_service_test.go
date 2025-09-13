package placeholders

import (
	"context"
	"testing"
	"time"
)

func TestNewPlaceholderService(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{
			name:   "nil config should use defaults",
			config: nil,
			want:   true,
		},
		{
			name: "custom config should be used",
			config: &Config{
				DefaultMockDataEnabled: false,
				CacheTimeout:           10 * time.Minute,
				MaxFeatures:            50,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewPlaceholderService(tt.config)
			if service == nil {
				t.Fatal("service should not be nil")
			}

			if tt.config == nil {
				// Check default config
				if !service.config.DefaultMockDataEnabled {
					t.Error("default mock data should be enabled")
				}
				if service.config.CacheTimeout != 5*time.Minute {
					t.Error("default cache timeout should be 5 minutes")
				}
				if service.config.MaxFeatures != 100 {
					t.Error("default max features should be 100")
				}
			} else {
				// Check custom config
				if service.config.DefaultMockDataEnabled != tt.config.DefaultMockDataEnabled {
					t.Error("custom mock data setting should be preserved")
				}
				if service.config.CacheTimeout != tt.config.CacheTimeout {
					t.Error("custom cache timeout should be preserved")
				}
				if service.config.MaxFeatures != tt.config.MaxFeatures {
					t.Error("custom max features should be preserved")
				}
			}

			// Check that default features are initialized
			count := service.GetFeatureCount(context.Background())
			if count == 0 {
				t.Error("default features should be initialized")
			}
		})
	}
}

func TestPlaceholderService_GetFeature(t *testing.T) {
	service := NewPlaceholderService(nil)
	ctx := context.Background()

	tests := []struct {
		name      string
		featureID string
		wantErr   bool
	}{
		{
			name:      "existing feature should be returned",
			featureID: "advanced_analytics",
			wantErr:   false,
		},
		{
			name:      "non-existing feature should return error",
			featureID: "non_existing",
			wantErr:   true,
		},
		{
			name:      "empty feature ID should return error",
			featureID: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feature, err := service.GetFeature(ctx, tt.featureID)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if feature != nil {
					t.Error("feature should be nil when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if feature == nil {
					t.Error("feature should not be nil")
				}
				if feature.ID != tt.featureID {
					t.Errorf("expected feature ID %s, got %s", tt.featureID, feature.ID)
				}
			}
		})
	}
}

func TestPlaceholderService_ListFeatures(t *testing.T) {
	service := NewPlaceholderService(nil)
	ctx := context.Background()

	tests := []struct {
		name     string
		status   *FeatureStatus
		category *string
		wantMin  int
	}{
		{
			name:     "list all features",
			status:   nil,
			category: nil,
			wantMin:  5, // Should have at least 5 default features
		},
		{
			name:     "filter by status",
			status:   func() *FeatureStatus { s := StatusComingSoon; return &s }(),
			category: nil,
			wantMin:  1,
		},
		{
			name:     "filter by category",
			status:   nil,
			category: stringPtr("analytics"),
			wantMin:  1,
		},
		{
			name:     "filter by both status and category",
			status:   func() *FeatureStatus { s := StatusComingSoon; return &s }(),
			category: stringPtr("analytics"),
			wantMin:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features, err := service.ListFeatures(ctx, tt.status, tt.category)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(features) < tt.wantMin {
				t.Errorf("expected at least %d features, got %d", tt.wantMin, len(features))
			}

			// Verify filters are applied correctly
			for _, feature := range features {
				if tt.status != nil && feature.Status != *tt.status {
					t.Errorf("feature %s has status %s, expected %s", feature.ID, feature.Status, *tt.status)
				}
				if tt.category != nil && feature.Category != *tt.category {
					t.Errorf("feature %s has category %s, expected %s", feature.ID, feature.Category, *tt.category)
				}
			}
		})
	}
}

func TestPlaceholderService_CreateFeature(t *testing.T) {
	service := NewPlaceholderService(&Config{MaxFeatures: 20}) // Allow room for default features
	ctx := context.Background()

	tests := []struct {
		name    string
		feature *Feature
		wantErr bool
	}{
		{
			name: "valid feature should be created",
			feature: &Feature{
				ID:          "test_feature",
				Name:        "Test Feature",
				Description: "A test feature",
				Status:      StatusComingSoon,
				Category:    "test",
				Priority:    1,
			},
			wantErr: false,
		},
		{
			name:    "nil feature should return error",
			feature: nil,
			wantErr: true,
		},
		{
			name: "empty ID should return error",
			feature: &Feature{
				ID:          "",
				Name:        "Test Feature",
				Description: "A test feature",
				Status:      StatusComingSoon,
				Category:    "test",
			},
			wantErr: true,
		},
		{
			name: "empty name should return error",
			feature: &Feature{
				ID:          "test_feature_2",
				Name:        "",
				Description: "A test feature",
				Status:      StatusComingSoon,
				Category:    "test",
			},
			wantErr: true,
		},
		{
			name: "invalid status should return error",
			feature: &Feature{
				ID:          "test_feature_3",
				Name:        "Test Feature",
				Description: "A test feature",
				Status:      FeatureStatus("invalid"),
				Category:    "test",
			},
			wantErr: true,
		},
		{
			name: "duplicate ID should return error",
			feature: &Feature{
				ID:          "advanced_analytics", // This should already exist
				Name:        "Duplicate Feature",
				Description: "A duplicate feature",
				Status:      StatusComingSoon,
				Category:    "test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateFeature(ctx, tt.feature)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				// Verify feature was created
				if tt.feature != nil {
					created, err := service.GetFeature(ctx, tt.feature.ID)
					if err != nil {
						t.Errorf("failed to retrieve created feature: %v", err)
					}
					if created.Name != tt.feature.Name {
						t.Errorf("expected name %s, got %s", tt.feature.Name, created.Name)
					}
				}
			}
		})
	}
}

func TestPlaceholderService_UpdateFeature(t *testing.T) {
	service := NewPlaceholderService(nil)
	ctx := context.Background()

	tests := []struct {
		name      string
		featureID string
		updates   *Feature
		wantErr   bool
	}{
		{
			name:      "valid update should succeed",
			featureID: "advanced_analytics",
			updates: &Feature{
				Name:        "Updated Analytics",
				Description: "Updated description",
				Status:      StatusInDevelopment,
			},
			wantErr: false,
		},
		{
			name:      "nil updates should return error",
			featureID: "advanced_analytics",
			updates:   nil,
			wantErr:   true,
		},
		{
			name:      "non-existing feature should return error",
			featureID: "non_existing",
			updates: &Feature{
				Name: "Updated Name",
			},
			wantErr: true,
		},
		{
			name:      "invalid status should return error",
			featureID: "advanced_analytics",
			updates: &Feature{
				Status: FeatureStatus("invalid"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateFeature(ctx, tt.featureID, tt.updates)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				// Verify feature was updated
				updated, err := service.GetFeature(ctx, tt.featureID)
				if err != nil {
					t.Errorf("failed to retrieve updated feature: %v", err)
				}
				if tt.updates.Name != "" && updated.Name != tt.updates.Name {
					t.Errorf("expected name %s, got %s", tt.updates.Name, updated.Name)
				}
				if tt.updates.Status != "" && updated.Status != tt.updates.Status {
					t.Errorf("expected status %s, got %s", tt.updates.Status, updated.Status)
				}
			}
		})
	}
}

func TestPlaceholderService_DeleteFeature(t *testing.T) {
	service := NewPlaceholderService(nil)
	ctx := context.Background()

	// Create a test feature first
	testFeature := &Feature{
		ID:          "test_delete_feature",
		Name:        "Test Delete Feature",
		Description: "A feature to test deletion",
		Status:      StatusComingSoon,
		Category:    "test",
	}
	err := service.CreateFeature(ctx, testFeature)
	if err != nil {
		t.Fatalf("failed to create test feature: %v", err)
	}

	tests := []struct {
		name      string
		featureID string
		wantErr   bool
	}{
		{
			name:      "existing feature should be deleted",
			featureID: "test_delete_feature",
			wantErr:   false,
		},
		{
			name:      "non-existing feature should return error",
			featureID: "non_existing",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteFeature(ctx, tt.featureID)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				// Verify feature was deleted
				_, err := service.GetFeature(ctx, tt.featureID)
				if err == nil {
					t.Error("feature should not exist after deletion")
				}
			}
		})
	}
}

func TestPlaceholderService_GetFeaturesByStatus(t *testing.T) {
	service := NewPlaceholderService(nil)
	ctx := context.Background()

	statuses := []FeatureStatus{StatusComingSoon, StatusInDevelopment, StatusAvailable, StatusDeprecated}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			features, err := service.GetFeaturesByStatus(ctx, status)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			for _, feature := range features {
				if feature.Status != status {
					t.Errorf("feature %s has status %s, expected %s", feature.ID, feature.Status, status)
				}
			}
		})
	}
}

func TestPlaceholderService_GetFeaturesByCategory(t *testing.T) {
	service := NewPlaceholderService(nil)
	ctx := context.Background()

	categories := []string{"analytics", "automation", "integration", "reporting", "monitoring", "security", "mobile"}

	for _, category := range categories {
		t.Run(category, func(t *testing.T) {
			features, err := service.GetFeaturesByCategory(ctx, category)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			for _, feature := range features {
				if feature.Category != category {
					t.Errorf("feature %s has category %s, expected %s", feature.ID, feature.Category, category)
				}
			}
		})
	}
}

func TestPlaceholderService_GetFeatureCount(t *testing.T) {
	service := NewPlaceholderService(nil)
	ctx := context.Background()

	count := service.GetFeatureCount(ctx)
	if count == 0 {
		t.Error("should have at least one feature")
	}

	// Create a new feature and verify count increases
	testFeature := &Feature{
		ID:          "test_count_feature",
		Name:        "Test Count Feature",
		Description: "A feature to test counting",
		Status:      StatusComingSoon,
		Category:    "test",
	}
	err := service.CreateFeature(ctx, testFeature)
	if err != nil {
		t.Fatalf("failed to create test feature: %v", err)
	}

	newCount := service.GetFeatureCount(ctx)
	if newCount != count+1 {
		t.Errorf("expected count %d, got %d", count+1, newCount)
	}
}

func TestPlaceholderService_GetFeatureCountByStatus(t *testing.T) {
	service := NewPlaceholderService(nil)
	ctx := context.Background()

	statuses := []FeatureStatus{StatusComingSoon, StatusInDevelopment, StatusAvailable, StatusDeprecated}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			count, err := service.GetFeatureCountByStatus(ctx, status)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if count < 0 {
				t.Error("count should not be negative")
			}
		})
	}
}

func TestPlaceholderService_GetFeatureStatistics(t *testing.T) {
	service := NewPlaceholderService(nil)
	ctx := context.Background()

	stats, err := service.GetFeatureStatistics(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check required fields
	requiredFields := []string{"total_features", "by_status", "by_category"}
	for _, field := range requiredFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("statistics should contain field: %s", field)
		}
	}

	// Check by_status structure
	byStatus, ok := stats["by_status"].(map[string]int)
	if !ok {
		t.Error("by_status should be a map[string]int")
	}

	expectedStatuses := []string{"coming_soon", "in_development", "available", "deprecated"}
	for _, status := range expectedStatuses {
		if _, exists := byStatus[status]; !exists {
			t.Errorf("by_status should contain status: %s", status)
		}
	}
}

func TestPlaceholderService_GetComingSoonFeatures(t *testing.T) {
	service := NewPlaceholderService(nil)
	ctx := context.Background()

	features, err := service.GetComingSoonFeatures(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	for _, feature := range features {
		if feature.Status != StatusComingSoon {
			t.Errorf("feature %s has status %s, expected %s", feature.ID, feature.Status, StatusComingSoon)
		}
	}
}

func TestPlaceholderService_GetInDevelopmentFeatures(t *testing.T) {
	service := NewPlaceholderService(nil)
	ctx := context.Background()

	features, err := service.GetInDevelopmentFeatures(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	for _, feature := range features {
		if feature.Status != StatusInDevelopment {
			t.Errorf("feature %s has status %s, expected %s", feature.ID, feature.Status, StatusInDevelopment)
		}
	}
}

func TestPlaceholderService_GetAvailableFeatures(t *testing.T) {
	service := NewPlaceholderService(nil)
	ctx := context.Background()

	features, err := service.GetAvailableFeatures(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	for _, feature := range features {
		if feature.Status != StatusAvailable {
			t.Errorf("feature %s has status %s, expected %s", feature.ID, feature.Status, StatusAvailable)
		}
	}
}

func TestPlaceholderService_ConcurrentAccess(t *testing.T) {
	service := NewPlaceholderService(nil)
	ctx := context.Background()

	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			_, err := service.ListFeatures(ctx, nil, nil)
			if err != nil {
				t.Errorf("concurrent read failed: %v", err)
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestPlaceholderService_MaxFeaturesLimit(t *testing.T) {
	// Create a service with a very small limit
	service := NewPlaceholderService(&Config{MaxFeatures: 1})
	ctx := context.Background()

	// The service initializes with default features, so we need to clear them first
	// or create a service without default initialization
	service.features = make(map[string]*Feature) // Clear default features

	// Create first feature (should succeed)
	firstFeature := &Feature{
		ID:          "test_max_limit_1",
		Name:        "Test Max Limit 1",
		Description: "First feature to test max limit",
		Status:      StatusComingSoon,
		Category:    "test",
	}

	err := service.CreateFeature(ctx, firstFeature)
	if err != nil {
		t.Fatalf("failed to create first feature: %v", err)
	}

	// Try to create a second feature (should fail due to max limit)
	secondFeature := &Feature{
		ID:          "test_max_limit_2",
		Name:        "Test Max Limit 2",
		Description: "Second feature to test max limit",
		Status:      StatusComingSoon,
		Category:    "test",
	}

	err = service.CreateFeature(ctx, secondFeature)
	if err == nil {
		t.Error("expected error when exceeding max features limit")
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func TestPlaceholderService_GenerateMockData(t *testing.T) {
	service := NewPlaceholderService(nil)

	tests := []struct {
		name     string
		category string
		wantKeys []string
	}{
		{
			name:     "analytics category",
			category: "analytics",
			wantKeys: []string{"sample_charts", "mock_metrics"},
		},
		{
			name:     "reporting category",
			category: "reporting",
			wantKeys: []string{"sample_reports", "mock_data"},
		},
		{
			name:     "integration category",
			category: "integration",
			wantKeys: []string{"available_apis", "mock_status"},
		},
		{
			name:     "automation category",
			category: "automation",
			wantKeys: []string{"workflow_templates", "mock_automation"},
		},
		{
			name:     "default category",
			category: "unknown",
			wantKeys: []string{"message", "mock_data", "last_updated"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feature := &Feature{Category: tt.category}
			mockData := service.generateMockData(feature)

			mockMap, ok := mockData.(map[string]interface{})
			if !ok {
				t.Error("mock data should be a map")
			}

			for _, key := range tt.wantKeys {
				if _, exists := mockMap[key]; !exists {
					t.Errorf("mock data should contain key: %s", key)
				}
			}
		})
	}
}
