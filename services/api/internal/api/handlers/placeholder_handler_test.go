package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"kyb-platform/internal/placeholders"
)

// mockPlaceholderService is a mock implementation of PlaceholderServiceInterface
type mockPlaceholderService struct {
	features map[string]*placeholders.Feature
	errors   map[string]error
}

func newMockPlaceholderService() *mockPlaceholderService {
	return &mockPlaceholderService{
		features: make(map[string]*placeholders.Feature),
		errors:   make(map[string]error),
	}
}

func (m *mockPlaceholderService) GetFeature(ctx context.Context, featureID string) (*placeholders.Feature, error) {
	if err, exists := m.errors["GetFeature"]; exists {
		return nil, err
	}
	if feature, exists := m.features[featureID]; exists {
		return feature, nil
	}
	return nil, fmt.Errorf("feature with ID %s not found", featureID)
}

func (m *mockPlaceholderService) ListFeatures(ctx context.Context, status *placeholders.FeatureStatus, category *string) ([]*placeholders.Feature, error) {
	if err, exists := m.errors["ListFeatures"]; exists {
		return nil, err
	}

	var features []*placeholders.Feature
	for _, feature := range m.features {
		// Apply status filter
		if status != nil && feature.Status != *status {
			continue
		}
		// Apply category filter
		if category != nil && feature.Category != *category {
			continue
		}
		features = append(features, feature)
	}
	return features, nil
}

func (m *mockPlaceholderService) CreateFeature(ctx context.Context, feature *placeholders.Feature) error {
	if err, exists := m.errors["CreateFeature"]; exists {
		return err
	}
	if feature.ID == "" {
		return fmt.Errorf("feature ID is required")
	}
	if _, exists := m.features[feature.ID]; exists {
		return fmt.Errorf("feature with ID %s already exists", feature.ID)
	}
	feature.CreatedAt = time.Now()
	feature.UpdatedAt = time.Now()
	m.features[feature.ID] = feature
	return nil
}

func (m *mockPlaceholderService) UpdateFeature(ctx context.Context, featureID string, updates *placeholders.Feature) error {
	if err, exists := m.errors["UpdateFeature"]; exists {
		return err
	}
	if _, exists := m.features[featureID]; !exists {
		return fmt.Errorf("feature with ID %s not found", featureID)
	}
	// Simple update logic for testing
	if updates.Name != "" {
		m.features[featureID].Name = updates.Name
	}
	if updates.Description != "" {
		m.features[featureID].Description = updates.Description
	}
	m.features[featureID].UpdatedAt = time.Now()
	return nil
}

func (m *mockPlaceholderService) DeleteFeature(ctx context.Context, featureID string) error {
	if err, exists := m.errors["DeleteFeature"]; exists {
		return err
	}
	if _, exists := m.features[featureID]; !exists {
		return fmt.Errorf("feature with ID %s not found", featureID)
	}
	delete(m.features, featureID)
	return nil
}

func (m *mockPlaceholderService) GetFeaturesByStatus(ctx context.Context, status placeholders.FeatureStatus) ([]*placeholders.Feature, error) {
	if err, exists := m.errors["GetFeaturesByStatus"]; exists {
		return nil, err
	}
	return m.ListFeatures(ctx, &status, nil)
}

func (m *mockPlaceholderService) GetFeaturesByCategory(ctx context.Context, category string) ([]*placeholders.Feature, error) {
	if err, exists := m.errors["GetFeaturesByCategory"]; exists {
		return nil, err
	}
	return m.ListFeatures(ctx, nil, &category)
}

func (m *mockPlaceholderService) GetComingSoonFeatures(ctx context.Context) ([]*placeholders.Feature, error) {
	return m.GetFeaturesByStatus(ctx, placeholders.StatusComingSoon)
}

func (m *mockPlaceholderService) GetInDevelopmentFeatures(ctx context.Context) ([]*placeholders.Feature, error) {
	return m.GetFeaturesByStatus(ctx, placeholders.StatusInDevelopment)
}

func (m *mockPlaceholderService) GetAvailableFeatures(ctx context.Context) ([]*placeholders.Feature, error) {
	return m.GetFeaturesByStatus(ctx, placeholders.StatusAvailable)
}

func (m *mockPlaceholderService) GetFeatureCount(ctx context.Context) int {
	return len(m.features)
}

func (m *mockPlaceholderService) GetFeatureCountByStatus(ctx context.Context, status placeholders.FeatureStatus) (int, error) {
	if err, exists := m.errors["GetFeatureCountByStatus"]; exists {
		return 0, err
	}
	features, err := m.GetFeaturesByStatus(ctx, status)
	if err != nil {
		return 0, err
	}
	return len(features), nil
}

func (m *mockPlaceholderService) GetFeatureStatistics(ctx context.Context) (map[string]interface{}, error) {
	if err, exists := m.errors["GetFeatureStatistics"]; exists {
		return nil, err
	}

	stats := make(map[string]interface{})
	stats["total_features"] = len(m.features)

	// Count by status
	statusCount := make(map[string]int)
	for _, feature := range m.features {
		statusCount[string(feature.Status)]++
	}
	stats["by_status"] = statusCount

	// Count by category
	categoryCount := make(map[string]int)
	for _, feature := range m.features {
		categoryCount[feature.Category]++
	}
	stats["by_category"] = categoryCount

	return stats, nil
}

// Helper function to create a test feature
func createTestFeature(id, name, description, status, category string) *placeholders.Feature {
	return &placeholders.Feature{
		ID:          id,
		Name:        name,
		Description: description,
		Status:      placeholders.FeatureStatus(status),
		Category:    category,
		Priority:    1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		MockData: map[string]interface{}{
			"test": true,
		},
	}
}

func TestPlaceholderHandler_GetFeature(t *testing.T) {
	tests := []struct {
		name           string
		featureID      string
		setupMock      func(*mockPlaceholderService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "successful get feature",
			featureID: "test-feature",
			setupMock: func(m *mockPlaceholderService) {
				m.features["test-feature"] = createTestFeature("test-feature", "Test Feature", "A test feature", "coming_soon", "test")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "feature not found",
			featureID:      "nonexistent",
			setupMock:      func(m *mockPlaceholderService) {},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Feature not found",
		},
		{
			name:      "service error",
			featureID: "test-feature",
			setupMock: func(m *mockPlaceholderService) {
				m.errors["GetFeature"] = fmt.Errorf("service error")
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Feature not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := newMockPlaceholderService()
			tt.setupMock(mockService)

			handler := NewPlaceholderHandler(mockService, log.Default())

			req := httptest.NewRequest("GET", "/api/v1/features/"+tt.featureID, nil)
			w := httptest.NewRecorder()

			handler.GetFeature(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var response APIResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if !strings.Contains(response.Error, tt.expectedError) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.expectedError, response.Error)
				}
			}
		})
	}
}

func TestPlaceholderHandler_ListFeatures(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*mockPlaceholderService)
		expectedStatus int
		expectedCount  int
	}{
		{
			name:        "list all features",
			queryParams: "",
			setupMock: func(m *mockPlaceholderService) {
				m.features["feature1"] = createTestFeature("feature1", "Feature 1", "Description 1", "coming_soon", "test")
				m.features["feature2"] = createTestFeature("feature2", "Feature 2", "Description 2", "available", "test")
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:        "filter by status",
			queryParams: "?status=coming_soon",
			setupMock: func(m *mockPlaceholderService) {
				m.features["feature1"] = createTestFeature("feature1", "Feature 1", "Description 1", "coming_soon", "test")
				m.features["feature2"] = createTestFeature("feature2", "Feature 2", "Description 2", "available", "test")
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:        "filter by category",
			queryParams: "?category=test",
			setupMock: func(m *mockPlaceholderService) {
				m.features["feature1"] = createTestFeature("feature1", "Feature 1", "Description 1", "coming_soon", "test")
				m.features["feature2"] = createTestFeature("feature2", "Feature 2", "Description 2", "available", "other")
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:        "pagination",
			queryParams: "?page=1&page_size=1",
			setupMock: func(m *mockPlaceholderService) {
				m.features["feature1"] = createTestFeature("feature1", "Feature 1", "Description 1", "coming_soon", "test")
				m.features["feature2"] = createTestFeature("feature2", "Feature 2", "Description 2", "available", "test")
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := newMockPlaceholderService()
			tt.setupMock(mockService)

			handler := NewPlaceholderHandler(mockService, log.Default())

			req := httptest.NewRequest("GET", "/api/v1/features"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.ListFeatures(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response APIResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if response.Success {
				var listResponse FeatureListResponse
				dataBytes, _ := json.Marshal(response.Data)
				if err := json.Unmarshal(dataBytes, &listResponse); err != nil {
					t.Fatalf("Failed to unmarshal list response: %v", err)
				}
				if listResponse.Count != tt.expectedCount {
					t.Errorf("Expected count %d, got %d", tt.expectedCount, listResponse.Count)
				}
			}
		})
	}
}

func TestPlaceholderHandler_CreateFeature(t *testing.T) {
	tests := []struct {
		name           string
		feature        *placeholders.Feature
		setupMock      func(*mockPlaceholderService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful create feature",
			feature: &placeholders.Feature{
				ID:          "new-feature",
				Name:        "New Feature",
				Description: "A new feature",
				Status:      placeholders.StatusComingSoon,
				Category:    "test",
			},
			setupMock:      func(m *mockPlaceholderService) {},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid JSON",
			feature: &placeholders.Feature{
				ID: "invalid-json",
			},
			setupMock: func(m *mockPlaceholderService) {
				// This will be handled by JSON decoding, not the service
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid JSON payload",
		},
		{
			name: "service error",
			feature: &placeholders.Feature{
				ID:          "error-feature",
				Name:        "Error Feature",
				Description: "A feature that causes error",
				Status:      placeholders.StatusComingSoon,
				Category:    "test",
			},
			setupMock: func(m *mockPlaceholderService) {
				m.errors["CreateFeature"] = fmt.Errorf("service error")
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Failed to create feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := newMockPlaceholderService()
			tt.setupMock(mockService)

			handler := NewPlaceholderHandler(mockService, log.Default())

			var reqBody []byte
			var err error
			if tt.name == "invalid JSON" {
				reqBody = []byte("invalid json")
			} else {
				reqBody, err = json.Marshal(tt.feature)
				if err != nil {
					t.Fatalf("Failed to marshal feature: %v", err)
				}
			}

			req := httptest.NewRequest("POST", "/api/v1/features", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateFeature(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var response APIResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if !strings.Contains(response.Error, tt.expectedError) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.expectedError, response.Error)
				}
			}
		})
	}
}

func TestPlaceholderHandler_UpdateFeature(t *testing.T) {
	tests := []struct {
		name           string
		featureID      string
		updates        *placeholders.Feature
		setupMock      func(*mockPlaceholderService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "successful update feature",
			featureID: "existing-feature",
			updates: &placeholders.Feature{
				Name:        "Updated Feature",
				Description: "Updated description",
			},
			setupMock: func(m *mockPlaceholderService) {
				m.features["existing-feature"] = createTestFeature("existing-feature", "Original Feature", "Original description", "coming_soon", "test")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "feature not found",
			featureID:      "nonexistent",
			updates:        &placeholders.Feature{Name: "Updated"},
			setupMock:      func(m *mockPlaceholderService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Failed to update feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := newMockPlaceholderService()
			tt.setupMock(mockService)

			handler := NewPlaceholderHandler(mockService, log.Default())

			reqBody, err := json.Marshal(tt.updates)
			if err != nil {
				t.Fatalf("Failed to marshal updates: %v", err)
			}

			req := httptest.NewRequest("PUT", "/api/v1/features/"+tt.featureID, bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.UpdateFeature(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var response APIResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if !strings.Contains(response.Error, tt.expectedError) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.expectedError, response.Error)
				}
			}
		})
	}
}

func TestPlaceholderHandler_DeleteFeature(t *testing.T) {
	tests := []struct {
		name           string
		featureID      string
		setupMock      func(*mockPlaceholderService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "successful delete feature",
			featureID: "existing-feature",
			setupMock: func(m *mockPlaceholderService) {
				m.features["existing-feature"] = createTestFeature("existing-feature", "Feature", "Description", "coming_soon", "test")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "feature not found",
			featureID:      "nonexistent",
			setupMock:      func(m *mockPlaceholderService) {},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Failed to delete feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := newMockPlaceholderService()
			tt.setupMock(mockService)

			handler := NewPlaceholderHandler(mockService, log.Default())

			req := httptest.NewRequest("DELETE", "/api/v1/features/"+tt.featureID, nil)
			w := httptest.NewRecorder()

			handler.DeleteFeature(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var response APIResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if !strings.Contains(response.Error, tt.expectedError) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.expectedError, response.Error)
				}
			}
		})
	}
}

func TestPlaceholderHandler_GetFeatureStatistics(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mockPlaceholderService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful get statistics",
			setupMock: func(m *mockPlaceholderService) {
				m.features["feature1"] = createTestFeature("feature1", "Feature 1", "Description 1", "coming_soon", "test")
				m.features["feature2"] = createTestFeature("feature2", "Feature 2", "Description 2", "available", "test")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "service error",
			setupMock: func(m *mockPlaceholderService) {
				m.errors["GetFeatureStatistics"] = fmt.Errorf("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to get feature statistics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := newMockPlaceholderService()
			tt.setupMock(mockService)

			handler := NewPlaceholderHandler(mockService, log.Default())

			req := httptest.NewRequest("GET", "/api/v1/features/statistics", nil)
			w := httptest.NewRecorder()

			handler.GetFeatureStatistics(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var response APIResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if !strings.Contains(response.Error, tt.expectedError) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.expectedError, response.Error)
				}
			}
		})
	}
}

func TestPlaceholderHandler_GetPlaceholderHealth(t *testing.T) {
	mockService := newMockPlaceholderService()
	mockService.features["feature1"] = createTestFeature("feature1", "Feature 1", "Description 1", "coming_soon", "test")

	handler := NewPlaceholderHandler(mockService, log.Default())

	req := httptest.NewRequest("GET", "/api/v1/placeholders/health", nil)
	w := httptest.NewRecorder()

	handler.GetPlaceholderHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success response, got error: %s", response.Error)
	}
}

func TestPlaceholderHandler_GetMockData(t *testing.T) {
	tests := []struct {
		name           string
		featureID      string
		setupMock      func(*mockPlaceholderService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "successful get mock data",
			featureID: "test-feature",
			setupMock: func(m *mockPlaceholderService) {
				m.features["test-feature"] = createTestFeature("test-feature", "Test Feature", "A test feature", "coming_soon", "test")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "feature not found",
			featureID:      "nonexistent",
			setupMock:      func(m *mockPlaceholderService) {},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Feature not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := newMockPlaceholderService()
			tt.setupMock(mockService)

			handler := NewPlaceholderHandler(mockService, log.Default())

			req := httptest.NewRequest("GET", "/api/v1/features/"+tt.featureID+"/mock-data", nil)
			w := httptest.NewRecorder()

			handler.GetMockData(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var response APIResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if !strings.Contains(response.Error, tt.expectedError) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.expectedError, response.Error)
				}
			}
		})
	}
}
