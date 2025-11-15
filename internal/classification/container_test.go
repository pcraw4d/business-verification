package classification

import (
	"context"
	"log"
	"testing"

	"kyb-platform/internal/classification/repository"
)

// TestClassificationContainerStructure tests the basic structure of the container
func TestClassificationContainerStructure(t *testing.T) {
	// Test that the container struct has the expected fields
	container := &ClassificationContainer{}

	// Verify the container can be created
	if container == nil {
		t.Fatal("Expected container to be created")
	}

	// Test that the container has the expected methods
	// This is a structural test to ensure the container is properly defined
	_ = container.GetIndustryDetectionService
	_ = container.GetCodeGenerator
	_ = container.GetRepository
	_ = container.GetLogger
	_ = container.HealthCheck
	_ = container.Close
}

// TestHealthCheckStructure tests the health check response structure
func TestHealthCheckStructure(t *testing.T) {
	// Create a minimal container for testing health check
	container := &ClassificationContainer{
		industryDetectionService: &IndustryDetectionService{},
		codeGenerator:            &ClassificationCodeGenerator{},
		repository:               nil, // nil to test degraded status
		logger:                   log.Default(),
	}

	health := container.HealthCheck()

	// Verify health check structure
	if health["status"] != "degraded" {
		t.Errorf("Expected status to be 'degraded', got %v", health["status"])
	}

	services, ok := health["services"].(map[string]string)
	if !ok {
		t.Fatal("Expected services to be a map")
	}

	expectedServices := []string{"industry_detection", "code_generator", "repository"}
	for _, service := range expectedServices {
		if services[service] != "active" {
			t.Errorf("Expected service %s to be 'active', got %v", service, services[service])
		}
	}

	// Verify repository status when nil
	if health["repository_status"] != "disconnected" {
		t.Errorf("Expected repository_status to be 'disconnected', got %v", health["repository_status"])
	}
}

// TestHealthCheckHealthy tests health check when repository is available
func TestHealthCheckHealthy(t *testing.T) {
	// Create a container with a mock repository
	container := &ClassificationContainer{
		industryDetectionService: &IndustryDetectionService{},
		codeGenerator:            &ClassificationCodeGenerator{},
		repository:               &MockContainerRepository{}, // Use mock repository
		logger:                   log.Default(),
	}

	health := container.HealthCheck()

	// Verify healthy status
	if health["status"] != "healthy" {
		t.Errorf("Expected status to be 'healthy', got %v", health["status"])
	}

	if health["repository_status"] != "connected" {
		t.Errorf("Expected repository_status to be 'connected', got %v", health["repository_status"])
	}
}

// TestCloseFunctionality tests the close functionality
func TestCloseFunctionality(t *testing.T) {
	container := &ClassificationContainer{
		logger: log.Default(),
	}

	// Test that close doesn't panic and returns no error
	err := container.Close()
	if err != nil {
		t.Errorf("Expected no error from close, got %v", err)
	}
}

// MockContainerRepository is a minimal mock for testing
type MockContainerRepository struct{}

// Implement minimal interface methods needed for testing
func (m *MockContainerRepository) GetIndustryByID(ctx context.Context, id int) (*repository.Industry, error) {
	return &repository.Industry{ID: id, Name: "Test Industry"}, nil
}

func (m *MockContainerRepository) GetIndustryByName(ctx context.Context, name string) (*repository.Industry, error) {
	return &repository.Industry{ID: 1, Name: name}, nil
}

func (m *MockContainerRepository) ListIndustries(ctx context.Context, category string) ([]*repository.Industry, error) {
	return []*repository.Industry{
		{ID: 1, Name: "Test Industry 1"},
		{ID: 2, Name: "Test Industry 2"},
	}, nil
}

func (m *MockContainerRepository) GetKeywordsByIndustry(ctx context.Context, industryID int) ([]*repository.IndustryKeyword, error) {
	return []*repository.IndustryKeyword{
		{ID: 1, IndustryID: industryID, Keyword: "test", Weight: 1.0},
	}, nil
}

func (m *MockContainerRepository) SearchKeywords(ctx context.Context, query string, limit int) ([]*repository.IndustryKeyword, error) {
	return []*repository.IndustryKeyword{
		{ID: 1, IndustryID: 1, Keyword: query, Weight: 1.0},
	}, nil
}

func (m *MockContainerRepository) ClassifyBusiness(ctx context.Context, businessName, websiteURL string) (*repository.ClassificationResult, error) {
	return &repository.ClassificationResult{
		Industry:   &repository.Industry{ID: 1, Name: "Test Industry"},
		Confidence: 0.85,
		Keywords:   []string{"test"},
	}, nil
}

func (m *MockContainerRepository) GetTopIndustriesByKeywords(ctx context.Context, keywords []string, limit int) ([]*repository.Industry, error) {
	return []*repository.Industry{
		{ID: 1, Name: "Test Industry 1"},
		{ID: 2, Name: "Test Industry 2"},
	}, nil
}

// Add other required methods with minimal implementations
func (m *MockContainerRepository) CreateIndustry(ctx context.Context, industry *repository.Industry) error {
	return nil
}
func (m *MockContainerRepository) UpdateIndustry(ctx context.Context, industry *repository.Industry) error {
	return nil
}
func (m *MockContainerRepository) DeleteIndustry(ctx context.Context, id int) error { return nil }
func (m *MockContainerRepository) AddKeywordToIndustry(ctx context.Context, industryID int, keyword string, weight float64) error {
	return nil
}
func (m *MockContainerRepository) UpdateKeywordWeight(ctx context.Context, keywordID int, weight float64) error {
	return nil
}
func (m *MockContainerRepository) RemoveKeywordFromIndustry(ctx context.Context, keywordID int) error {
	return nil
}
func (m *MockContainerRepository) GetClassificationCodesByIndustry(ctx context.Context, industryID int) ([]*repository.ClassificationCode, error) {
	return nil, nil
}
func (m *MockContainerRepository) GetClassificationCodesByType(ctx context.Context, codeType string) ([]*repository.ClassificationCode, error) {
	return nil, nil
}
func (m *MockContainerRepository) AddClassificationCode(ctx context.Context, code *repository.ClassificationCode) error {
	return nil
}
func (m *MockContainerRepository) UpdateClassificationCode(ctx context.Context, code *repository.ClassificationCode) error {
	return nil
}
func (m *MockContainerRepository) DeleteClassificationCode(ctx context.Context, id int) error {
	return nil
}
func (m *MockContainerRepository) GetPatternsByIndustry(ctx context.Context, industryID int) ([]*repository.IndustryPattern, error) {
	return nil, nil
}
func (m *MockContainerRepository) AddPattern(ctx context.Context, pattern *repository.IndustryPattern) error {
	return nil
}
func (m *MockContainerRepository) UpdatePattern(ctx context.Context, pattern *repository.IndustryPattern) error {
	return nil
}
func (m *MockContainerRepository) DeletePattern(ctx context.Context, id int) error { return nil }
func (m *MockContainerRepository) GetKeywordWeights(ctx context.Context, keyword string) ([]*repository.KeywordWeight, error) {
	return nil, nil
}
func (m *MockContainerRepository) UpdateKeywordWeightByID(ctx context.Context, weight *repository.KeywordWeight) error {
	return nil
}
func (m *MockContainerRepository) IncrementUsageCount(ctx context.Context, keyword string, industryID int) error {
	return nil
}
func (m *MockContainerRepository) ClassifyBusinessByKeywords(ctx context.Context, keywords []string) (*repository.ClassificationResult, error) {
	return nil, nil
}
func (m *MockContainerRepository) SearchIndustriesByPattern(ctx context.Context, pattern string) ([]*repository.Industry, error) {
	return nil, nil
}
func (m *MockContainerRepository) GetIndustryStatistics(ctx context.Context) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockContainerRepository) GetKeywordFrequency(ctx context.Context, industryID int) (map[string]int, error) {
	return nil, nil
}
func (m *MockContainerRepository) BulkInsertKeywords(ctx context.Context, keywords []*repository.IndustryKeyword) error {
	return nil
}
func (m *MockContainerRepository) BulkUpdateKeywords(ctx context.Context, keywords []*repository.IndustryKeyword) error {
	return nil
}
func (m *MockContainerRepository) BulkDeleteKeywords(ctx context.Context, keywordIDs []int) error {
	return nil
}
func (m *MockContainerRepository) Ping(ctx context.Context) error { return nil }
func (m *MockContainerRepository) GetDatabaseStats(ctx context.Context) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockContainerRepository) CleanupInactiveData(ctx context.Context) error { return nil }

// Add missing batch methods
func (m *MockContainerRepository) GetBatchClassificationCodes(ctx context.Context, industryIDs []int) (map[int][]*repository.ClassificationCode, error) {
	result := make(map[int][]*repository.ClassificationCode)
	for _, id := range industryIDs {
		result[id] = []*repository.ClassificationCode{
			{ID: 1, Code: "541511", Description: "Custom Computer Programming Services", CodeType: "NAICS"},
		}
	}
	return result, nil
}

func (m *MockContainerRepository) GetBatchIndustries(ctx context.Context, industryIDs []int) (map[int]*repository.Industry, error) {
	result := make(map[int]*repository.Industry)
	for _, id := range industryIDs {
		result[id] = &repository.Industry{ID: id, Name: "Test Industry"}
	}
	return result, nil
}

func (m *MockContainerRepository) GetBatchKeywords(ctx context.Context, industryIDs []int) (map[int][]*repository.KeywordWeight, error) {
	result := make(map[int][]*repository.KeywordWeight)
	for _, id := range industryIDs {
		result[id] = []*repository.KeywordWeight{
			{ID: 1, Keyword: "test", BaseWeight: 1.0},
		}
	}
	return result, nil
}

func (m *MockContainerRepository) GetCachedClassificationCodes(ctx context.Context, industryID int) ([]*repository.ClassificationCode, error) {
	return []*repository.ClassificationCode{
		{ID: 1, Code: "541511", Description: "Custom Computer Programming Services", CodeType: "NAICS"},
	}, nil
}

func (m *MockContainerRepository) GetCachedClassificationCodesByType(ctx context.Context, codeType string) ([]*repository.ClassificationCode, error) {
	return []*repository.ClassificationCode{
		{ID: 1, Code: "541511", Description: "Custom Computer Programming Services", CodeType: codeType},
	}, nil
}

func (m *MockContainerRepository) InitializeIndustryCodeCache(ctx context.Context) error {
	return nil
}

func (m *MockContainerRepository) InvalidateIndustryCodeCache(ctx context.Context, rules []string) error {
	return nil
}

func (m *MockContainerRepository) GetIndustryCodeCacheStats() *repository.IndustryCodeCacheStats {
	return &repository.IndustryCodeCacheStats{
		Hits:   10,
		Misses: 2,
	}
}
