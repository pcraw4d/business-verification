package repository

import (
	"context"
	"log"
	"testing"
	"time"
)

// MockPostgrestQuery is a mock implementation for PostgREST queries
type MockPostgrestQuery struct{}

func (m *MockPostgrestQuery) Select(columns, count string, head bool) PostgrestQueryInterface {
	return m
}
func (m *MockPostgrestQuery) Eq(column, value string) PostgrestQueryInterface            { return m }
func (m *MockPostgrestQuery) Ilike(column, value string) PostgrestQueryInterface         { return m }
func (m *MockPostgrestQuery) In(column string, values ...string) PostgrestQueryInterface { return m }
func (m *MockPostgrestQuery) Order(column string, ascending *map[string]string) PostgrestQueryInterface {
	return m
}
func (m *MockPostgrestQuery) Limit(count int, foreignTable string) PostgrestQueryInterface { return m }
func (m *MockPostgrestQuery) Single() PostgrestQueryInterface                              { return m }
func (m *MockPostgrestQuery) Execute() ([]byte, string, error)                             { return []byte{}, "", nil }

// MockPostgrestClient is a mock implementation for PostgREST client
type MockPostgrestClient struct{}

func (m *MockPostgrestClient) From(table string) PostgrestQueryInterface {
	return &MockPostgrestQuery{}
}

// MockSupabaseClient is a mock implementation for testing
type MockSupabaseClient struct {
	pingError error
}

func (m *MockSupabaseClient) Connect(ctx context.Context) error {
	return nil
}

func (m *MockSupabaseClient) Close() error {
	return nil
}

func (m *MockSupabaseClient) Ping(ctx context.Context) error {
	return m.pingError
}

func (m *MockSupabaseClient) GetClient() interface{} {
	return nil
}

func (m *MockSupabaseClient) GetPostgrestClient() PostgrestClientInterface {
	return &MockPostgrestClient{}
}

// TestNewSupabaseKeywordRepository tests repository creation
func TestNewSupabaseKeywordRepository(t *testing.T) {
	mockClient := &MockSupabaseClient{}
	logger := log.Default()

	repo := NewSupabaseKeywordRepositoryWithInterface(mockClient, logger)

	if repo == nil {
		t.Fatal("Expected repository to be created, got nil")
	}

	// For interface-based clients, the concrete client will be nil
	// but the clientInterface will be set
	if repo.client == nil && repo.clientInterface == nil {
		t.Error("Expected either client or clientInterface to be set")
	}

	if repo.logger != logger {
		t.Error("Expected logger to be set correctly")
	}
}

// TestNewSupabaseKeywordRepositoryWithNilLogger tests repository creation with nil logger
func TestNewSupabaseKeywordRepositoryWithNilLogger(t *testing.T) {
	mockClient := &MockSupabaseClient{}

	repo := NewSupabaseKeywordRepositoryWithInterface(mockClient, nil)

	if repo == nil {
		t.Fatal("Expected repository to be created, got nil")
	}

	if repo.logger == nil {
		t.Error("Expected default logger to be set")
	}
}

// TestSupabaseKeywordRepository_Ping tests the ping functionality
func TestSupabaseKeywordRepository_Ping(t *testing.T) {
	tests := []struct {
		name        string
		pingError   error
		expectError bool
	}{
		{
			name:        "successful ping",
			pingError:   nil,
			expectError: false,
		},
		{
			name:        "ping failure",
			pingError:   context.DeadlineExceeded,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockSupabaseClient{pingError: tt.pingError}
			repo := NewSupabaseKeywordRepositoryWithInterface(mockClient, log.Default())

			err := repo.Ping(context.Background())

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// TestSupabaseKeywordRepository_ClassifyBusiness tests basic business classification
func TestSupabaseKeywordRepository_ClassifyBusiness(t *testing.T) {
	mockClient := &MockSupabaseClient{}
	repo := NewSupabaseKeywordRepositoryWithInterface(mockClient, log.Default())

	ctx := context.Background()

	// Test with empty inputs
	result, err := repo.ClassifyBusiness(ctx, "", "")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Industry.Name != "General Business" {
		t.Errorf("Expected General Business, got: %s", result.Industry.Name)
	}

	if result.Confidence != 0.50 {
		t.Errorf("Expected confidence 0.50, got: %f", result.Confidence)
	}
}

// TestSupabaseKeywordRepository_ClassifyBusinessByKeywords tests keyword-based classification
func TestSupabaseKeywordRepository_ClassifyBusinessByKeywords(t *testing.T) {
	mockClient := &MockSupabaseClient{}
	repo := NewSupabaseKeywordRepositoryWithInterface(mockClient, log.Default())

	ctx := context.Background()

	// Test with empty keywords
	result, err := repo.ClassifyBusinessByKeywords(ctx, []string{})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Industry.Name != "General Business" {
		t.Errorf("Expected General Business, got: %s", result.Industry.Name)
	}

	// Test with some keywords
	result, err = repo.ClassifyBusinessByKeywords(ctx, []string{"software", "technology"})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Industry.Name != "Technology" {
		t.Errorf("Expected Technology, got: %s", result.Industry.Name)
	}

	if result.Confidence != 0.75 {
		t.Errorf("Expected confidence 0.75, got: %f", result.Confidence)
	}
}

// TestSupabaseKeywordRepository_extractKeywords tests keyword extraction
func TestSupabaseKeywordRepository_extractKeywords(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	mockClient := &MockSupabaseClient{}
	repo := NewSupabaseKeywordRepositoryWithInterface(mockClient, log.Default())

	// Test with business name only
	contextualKeywords := repo.extractKeywords(ctx, "Acme Software Solutions", "")
	expected := []string{"acme", "software", "solutions"}

	// The new method extracts more keywords including phrases, so we check for at least the expected ones
	if len(contextualKeywords) < len(expected) {
		t.Errorf("Expected at least %d keywords, got %d", len(expected), len(contextualKeywords))
	}

	// Check that all expected keywords are present
	keywordMap := make(map[string]bool)
	for _, ck := range contextualKeywords {
		keywordMap[ck.Keyword] = true
		if ck.Context != "business_name" {
			t.Errorf("Expected context 'business_name' for keyword %s, got %s", ck.Keyword, ck.Context)
		}
	}

	for _, expectedKeyword := range expected {
		if !keywordMap[expectedKeyword] {
			t.Errorf("Expected keyword %s not found in extracted keywords", expectedKeyword)
		}
	}

	// Test with description
	contextualKeywords = repo.extractKeywords(ctx, "", "https://example.com")
	if len(contextualKeywords) < 4 {
		t.Errorf("Expected at least 4 keywords, got %d", len(contextualKeywords))
	}
	// Verify all are from description context
	for _, ck := range contextualKeywords {
		if ck.Context != "description" {
			t.Errorf("Expected context 'description', got %s", ck.Context)
		}
	}

	// Test with website URL
	contextualKeywords = repo.extractKeywords(ctx, "", "https://www.tech-company.com")
	if len(contextualKeywords) < 1 {
		t.Errorf("Expected at least 1 keyword, got %d", len(contextualKeywords))
	}
	// Verify all are from website_url or website_content context
	for _, ck := range contextualKeywords {
		if ck.Context != "website_url" && ck.Context != "website_content" {
			t.Errorf("Expected context 'website_url' or 'website_content', got %s", ck.Context)
		}
	}

	// Test with all inputs
	contextualKeywords = repo.extractKeywords(ctx, "Tech Corp", "https://www.techcorp.com")
	if len(contextualKeywords) < 5 {
		t.Errorf("Expected at least 5 keywords, got %d", len(contextualKeywords))
	}
	// Verify we have keywords from available contexts (description removed for security)
	contexts := make(map[string]int)
	for _, ck := range contextualKeywords {
		contexts[ck.Context]++
	}
	if contexts["business_name"] == 0 {
		t.Errorf("Expected business_name keywords")
	}
	if contexts["description"] == 0 {
		t.Errorf("Expected description keywords")
	}
	if contexts["website_url"] == 0 {
		t.Errorf("Expected website_url keywords")
	}
}

// TestSupabaseKeywordRepository_InterfaceCompliance tests that the repository implements the interface
func TestSupabaseKeywordRepository_InterfaceCompliance(t *testing.T) {
	var _ KeywordRepository = (*SupabaseKeywordRepository)(nil)
}

// BenchmarkSupabaseKeywordRepository_ClassifyBusiness benchmarks business classification
func BenchmarkSupabaseKeywordRepository_ClassifyBusiness(b *testing.B) {
	mockClient := &MockSupabaseClient{}
	repo := NewSupabaseKeywordRepositoryWithInterface(mockClient, log.Default())
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.ClassifyBusiness(ctx, "Test Company", "https://test.com")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSupabaseKeywordRepository_extractKeywords benchmarks keyword extraction
func BenchmarkSupabaseKeywordRepository_extractKeywords(b *testing.B) {
	ctx := context.Background()
	mockClient := &MockSupabaseClient{}
	repo := NewSupabaseKeywordRepositoryWithInterface(mockClient, log.Default())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.extractKeywords(ctx, "Acme Software Solutions Inc", "https://www.acme-software-solutions.com")
	}
}
