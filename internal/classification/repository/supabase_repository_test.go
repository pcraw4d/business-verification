package repository

import (
	"context"
	"log"
	"testing"
)

// MockPostgrestQuery is a mock implementation for PostgREST queries
type MockPostgrestQuery struct{}

func (m *MockPostgrestQuery) Select(columns, count string, head bool) PostgrestQueryInterface {
	return m
}
func (m *MockPostgrestQuery) Eq(column, value string) PostgrestQueryInterface    { return m }
func (m *MockPostgrestQuery) Ilike(column, value string) PostgrestQueryInterface { return m }
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

	repo := NewSupabaseKeywordRepository(mockClient, logger)

	if repo == nil {
		t.Fatal("Expected repository to be created, got nil")
	}

	if repo.client != mockClient {
		t.Error("Expected client to be set correctly")
	}

	if repo.logger != logger {
		t.Error("Expected logger to be set correctly")
	}
}

// TestNewSupabaseKeywordRepositoryWithNilLogger tests repository creation with nil logger
func TestNewSupabaseKeywordRepositoryWithNilLogger(t *testing.T) {
	mockClient := &MockSupabaseClient{}

	repo := NewSupabaseKeywordRepository(mockClient, nil)

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
			repo := NewSupabaseKeywordRepository(mockClient, log.Default())

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
	repo := NewSupabaseKeywordRepository(mockClient, log.Default())

	ctx := context.Background()

	// Test with empty inputs
	result, err := repo.ClassifyBusiness(ctx, "", "", "")
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
	repo := NewSupabaseKeywordRepository(mockClient, log.Default())

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
	mockClient := &MockSupabaseClient{}
	repo := NewSupabaseKeywordRepository(mockClient, log.Default())

	// Test with business name only
	keywords := repo.extractKeywords("Acme Software Solutions", "", "")
	expected := []string{"acme", "software", "solutions"}

	if len(keywords) != len(expected) {
		t.Errorf("Expected %d keywords, got %d", len(expected), len(keywords))
	}

	for i, keyword := range expected {
		if keywords[i] != keyword {
			t.Errorf("Expected keyword %s at position %d, got %s", keyword, i, keywords[i])
		}
	}

	// Test with description
	keywords = repo.extractKeywords("", "We provide cloud computing services", "")
	if len(keywords) != 4 {
		t.Errorf("Expected 4 keywords, got %d", len(keywords))
	}

	// Test with website URL
	keywords = repo.extractKeywords("", "", "https://www.tech-company.com")
	if len(keywords) != 2 {
		t.Errorf("Expected 2 keywords, got %d", len(keywords))
	}

	// Test with all inputs
	keywords = repo.extractKeywords("Tech Corp", "Software development", "https://www.techcorp.com")
	if len(keywords) < 5 {
		t.Errorf("Expected at least 5 keywords, got %d", len(keywords))
	}
}

// TestSupabaseKeywordRepository_InterfaceCompliance tests that the repository implements the interface
func TestSupabaseKeywordRepository_InterfaceCompliance(t *testing.T) {
	var _ KeywordRepository = (*SupabaseKeywordRepository)(nil)
}

// BenchmarkSupabaseKeywordRepository_ClassifyBusiness benchmarks business classification
func BenchmarkSupabaseKeywordRepository_ClassifyBusiness(b *testing.B) {
	mockClient := &MockSupabaseClient{}
	repo := NewSupabaseKeywordRepository(mockClient, log.Default())
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.ClassifyBusiness(ctx, "Test Company", "Test description", "https://test.com")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSupabaseKeywordRepository_extractKeywords benchmarks keyword extraction
func BenchmarkSupabaseKeywordRepository_extractKeywords(b *testing.B) {
	mockClient := &MockSupabaseClient{}
	repo := NewSupabaseKeywordRepository(mockClient, log.Default())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.extractKeywords("Acme Software Solutions Inc", "We provide cloud computing and software development services", "https://www.acme-software-solutions.com")
	}
}
