package repository

import (
	"context"
	"encoding/json"
	"log"
	"testing"
)

// TestGetClassificationCodesByKeywords tests the new keyword-based code retrieval method
func TestGetClassificationCodesByKeywords(t *testing.T) {
	tests := []struct {
		name          string
		keywords      []string
		codeType      string
		minRelevance  float64
		mockResponse  []byte
		expectError   bool
		expectedCount int
	}{
		{
			name:         "successful keyword match",
			keywords:     []string{"software", "technology"},
			codeType:     "MCC",
			minRelevance: 0.5,
			mockResponse: jsonResponse([]map[string]interface{}{
				{
					"id":             1,
					"code_id":        101,
					"keyword":        "software",
					"relevance_score": 0.85,
					"match_type":     "exact",
				},
				{
					"id":             2,
					"code_id":        102,
					"keyword":        "technology",
					"relevance_score": 0.75,
					"match_type":     "partial",
				},
			}),
			expectError:   false,
			expectedCount: 2,
		},
		{
			name:         "empty keywords",
			keywords:     []string{},
			codeType:     "MCC",
			minRelevance:  0.5,
			expectError:  false,
			expectedCount: 0,
		},
		{
			name:         "no matches found",
			keywords:     []string{"nonexistent"},
			codeType:     "MCC",
			minRelevance: 0.5,
			mockResponse: jsonResponse([]map[string]interface{}{}),
			expectError:  false,
			expectedCount: 0,
		},
		{
			name:         "filter by relevance threshold",
			keywords:     []string{"test"},
			codeType:     "SIC",
			minRelevance:  0.7,
			mockResponse: jsonResponse([]map[string]interface{}{
				{
					"id":             1,
					"code_id":        201,
					"keyword":        "test",
					"relevance_score": 0.6, // Below threshold
					"match_type":     "exact",
				},
				{
					"id":             2,
					"code_id":        202,
					"keyword":        "test",
					"relevance_score": 0.8, // Above threshold
					"match_type":     "exact",
				},
			}),
			expectError:   false,
			expectedCount: 1, // Only one above threshold
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockSupabaseClientWithResponse{
				response: tt.mockResponse,
			}
			repo := NewSupabaseKeywordRepositoryWithInterface(mockClient, log.Default())

			ctx := context.Background()
			results, err := repo.GetClassificationCodesByKeywords(ctx, tt.keywords, tt.codeType, tt.minRelevance)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(results))
			}

			// Verify results have correct structure
			for _, result := range results {
				if result.ClassificationCode.CodeType != tt.codeType {
					t.Errorf("Expected code type %s, got %s", tt.codeType, result.ClassificationCode.CodeType)
				}
				if result.RelevanceScore < tt.minRelevance {
					t.Errorf("Result relevance score %.2f is below threshold %.2f", result.RelevanceScore, tt.minRelevance)
				}
			}
		})
	}
}

// TestGetClassificationCodesByKeywords_CodeTypeFiltering tests filtering by code type
func TestGetClassificationCodesByKeywords_CodeTypeFiltering(t *testing.T) {
	mockClient := &MockSupabaseClientWithResponse{
		response: jsonResponse([]map[string]interface{}{
			{
				"id":             1,
				"code_id":        101,
				"keyword":        "test",
				"relevance_score": 0.8,
				"match_type":     "exact",
			},
		}),
	}
	repo := NewSupabaseKeywordRepositoryWithInterface(mockClient, log.Default())

	ctx := context.Background()
	
	// Test different code types
	codeTypes := []string{"MCC", "SIC", "NAICS"}
	for _, codeType := range codeTypes {
		results, err := repo.GetClassificationCodesByKeywords(ctx, []string{"test"}, codeType, 0.5)
		if err != nil {
			t.Errorf("Unexpected error for code type %s: %v", codeType, err)
			continue
		}

		for _, result := range results {
			if result.ClassificationCode.CodeType != codeType {
				t.Errorf("Expected code type %s, got %s", codeType, result.ClassificationCode.CodeType)
			}
		}
	}
}

// MockSupabaseClientWithResponse extends MockSupabaseClient to return custom responses
type MockSupabaseClientWithResponse struct {
	*MockSupabaseClient
	response []byte
}

func (m *MockSupabaseClientWithResponse) GetPostgrestClient() PostgrestClientInterface {
	return &MockPostgrestClientWithResponse{response: m.response}
}

// MockPostgrestClientWithResponse extends MockPostgrestClient to return custom responses
type MockPostgrestClientWithResponse struct {
	*MockPostgrestClient
	response []byte
}

func (m *MockPostgrestClientWithResponse) From(table string) PostgrestQueryInterface {
	return &MockPostgrestQueryWithResponse{response: m.response}
}

// MockPostgrestQueryWithResponse extends MockPostgrestQuery to return custom responses
type MockPostgrestQueryWithResponse struct {
	*MockPostgrestQuery
	response []byte
}

func (m *MockPostgrestQueryWithResponse) Execute() ([]byte, string, error) {
	return m.response, "", nil
}

// Helper function to create JSON response
func jsonResponse(data interface{}) []byte {
	jsonData, _ := json.Marshal(data)
	return jsonData
}

