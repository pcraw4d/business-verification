package classification

import (
	"context"
	"log"
	"strings"
	"testing"

	"github.com/pcraw4d/business-verification/internal/classification/repository"
)

// MockKeywordRepository is a mock implementation for testing
type MockKeywordRepository struct {
	classifyBusinessResult       *repository.ClassificationResult
	classifyBusinessError        error
	classifyByKeywordsResult     *repository.ClassificationResult
	classifyByKeywordsError      error
	getClassificationCodesResult []*repository.ClassificationCode
	getClassificationCodesError  error
}

func (m *MockKeywordRepository) GetIndustryByID(ctx context.Context, id int) (*repository.Industry, error) {
	return &repository.Industry{ID: id, Name: "Test Industry"}, nil
}

func (m *MockKeywordRepository) GetIndustryByName(ctx context.Context, name string) (*repository.Industry, error) {
	return &repository.Industry{ID: 1, Name: name}, nil
}

func (m *MockKeywordRepository) ListIndustries(ctx context.Context, category string) ([]*repository.Industry, error) {
	return []*repository.Industry{{ID: 1, Name: "Test Industry"}}, nil
}

func (m *MockKeywordRepository) CreateIndustry(ctx context.Context, industry *repository.Industry) error {
	return nil
}

func (m *MockKeywordRepository) UpdateIndustry(ctx context.Context, industry *repository.Industry) error {
	return nil
}

func (m *MockKeywordRepository) DeleteIndustry(ctx context.Context, id int) error {
	return nil
}

func (m *MockKeywordRepository) GetKeywordsByIndustry(ctx context.Context, industryID int) ([]*repository.IndustryKeyword, error) {
	return []*repository.IndustryKeyword{{ID: 1, Keyword: "test"}}, nil
}

func (m *MockKeywordRepository) SearchKeywords(ctx context.Context, query string, limit int) ([]*repository.IndustryKeyword, error) {
	return []*repository.IndustryKeyword{{ID: 1, Keyword: "test"}}, nil
}

func (m *MockKeywordRepository) AddKeywordToIndustry(ctx context.Context, industryID int, keyword string, weight float64) error {
	return nil
}

func (m *MockKeywordRepository) UpdateKeywordWeight(ctx context.Context, keywordID int, weight float64) error {
	return nil
}

func (m *MockKeywordRepository) RemoveKeywordFromIndustry(ctx context.Context, keywordID int) error {
	return nil
}

func (m *MockKeywordRepository) GetClassificationCodesByIndustry(ctx context.Context, industryID int) ([]*repository.ClassificationCode, error) {
	if m.getClassificationCodesError != nil {
		return nil, m.getClassificationCodesError
	}
	return m.getClassificationCodesResult, nil
}

func (m *MockKeywordRepository) GetClassificationCodesByType(ctx context.Context, codeType string) ([]*repository.ClassificationCode, error) {
	return []*repository.ClassificationCode{{ID: 1, Code: "TEST123"}}, nil
}

func (m *MockKeywordRepository) AddClassificationCode(ctx context.Context, code *repository.ClassificationCode) error {
	return nil
}

func (m *MockKeywordRepository) UpdateClassificationCode(ctx context.Context, code *repository.ClassificationCode) error {
	return nil
}

func (m *MockKeywordRepository) DeleteClassificationCode(ctx context.Context, id int) error {
	return nil
}

func (m *MockKeywordRepository) GetPatternsByIndustry(ctx context.Context, industryID int) ([]*repository.IndustryPattern, error) {
	return []*repository.IndustryPattern{{ID: 1, Pattern: "test pattern"}}, nil
}

func (m *MockKeywordRepository) AddPattern(ctx context.Context, pattern *repository.IndustryPattern) error {
	return nil
}

func (m *MockKeywordRepository) UpdatePattern(ctx context.Context, pattern *repository.IndustryPattern) error {
	return nil
}

func (m *MockKeywordRepository) DeletePattern(ctx context.Context, id int) error {
	return nil
}

func (m *MockKeywordRepository) GetKeywordWeights(ctx context.Context, keyword string) ([]*repository.KeywordWeight, error) {
	return []*repository.KeywordWeight{{ID: 1, Keyword: keyword}}, nil
}

func (m *MockKeywordRepository) UpdateKeywordWeightByID(ctx context.Context, weight *repository.KeywordWeight) error {
	return nil
}

func (m *MockKeywordRepository) IncrementUsageCount(ctx context.Context, keyword string, industryID int) error {
	return nil
}

func (m *MockKeywordRepository) ClassifyBusiness(ctx context.Context, businessName, description, websiteURL string) (*repository.ClassificationResult, error) {
	if m.classifyBusinessError != nil {
		return nil, m.classifyBusinessError
	}
	return m.classifyBusinessResult, nil
}

func (m *MockKeywordRepository) ClassifyBusinessByKeywords(ctx context.Context, keywords []string) (*repository.ClassificationResult, error) {
	if m.classifyByKeywordsError != nil {
		return nil, m.classifyByKeywordsError
	}
	return m.classifyByKeywordsResult, nil
}

func (m *MockKeywordRepository) GetTopIndustriesByKeywords(ctx context.Context, keywords []string, limit int) ([]*repository.Industry, error) {
	return []*repository.Industry{{ID: 1, Name: "Test Industry"}}, nil
}

func (m *MockKeywordRepository) SearchIndustriesByPattern(ctx context.Context, pattern string) ([]*repository.Industry, error) {
	return []*repository.Industry{{ID: 1, Name: "Test Industry"}}, nil
}

func (m *MockKeywordRepository) GetIndustryStatistics(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{"total_industries": 1}, nil
}

func (m *MockKeywordRepository) GetKeywordFrequency(ctx context.Context, industryID int) (map[string]int, error) {
	return map[string]int{"test": 1}, nil
}

func (m *MockKeywordRepository) BulkInsertKeywords(ctx context.Context, keywords []*repository.IndustryKeyword) error {
	return nil
}

func (m *MockKeywordRepository) BulkUpdateKeywords(ctx context.Context, keywords []*repository.IndustryKeyword) error {
	return nil
}

func (m *MockKeywordRepository) BulkDeleteKeywords(ctx context.Context, keywordIDs []int) error {
	return nil
}

func (m *MockKeywordRepository) GetDatabaseStats(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{"status": "ok"}, nil
}

func (m *MockKeywordRepository) CleanupInactiveData(ctx context.Context) error {
	return nil
}

func (m *MockKeywordRepository) Ping(ctx context.Context) error {
	return nil
}

// TestNewIndustryDetectionService tests service creation
func TestNewIndustryDetectionService(t *testing.T) {
	mockRepo := &MockKeywordRepository{}
	logger := log.Default()

	service := NewIndustryDetectionService(mockRepo, logger)

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	if service.repo != mockRepo {
		t.Error("Expected repository to be set correctly")
	}

	if service.logger != logger {
		t.Error("Expected logger to be set correctly")
	}
}

// TestNewIndustryDetectionServiceWithNilLogger tests service creation with nil logger
func TestNewIndustryDetectionServiceWithNilLogger(t *testing.T) {
	mockRepo := &MockKeywordRepository{}

	service := NewIndustryDetectionService(mockRepo, nil)

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	if service.logger == nil {
		t.Error("Expected default logger to be set")
	}
}

// TestDetectIndustryFromContent tests content-based industry detection
func TestDetectIndustryFromContent(t *testing.T) {
	mockRepo := &MockKeywordRepository{
		classifyByKeywordsResult: &repository.ClassificationResult{
			Industry:   &repository.Industry{ID: 1, Name: "Technology"},
			Confidence: 0.85,
			Keywords:   []string{"software", "technology"},
			Reasoning:  "Technology keywords detected",
		},
		getClassificationCodesResult: []*repository.ClassificationCode{
			{ID: 1, Code: "511210", Description: "Software Publishers"},
		},
	}

	service := NewIndustryDetectionService(mockRepo, log.Default())
	ctx := context.Background()

	// Test with valid content
	result, err := service.DetectIndustryFromContent(ctx, "We develop software solutions for businesses")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Industry.Name != "Technology" {
		t.Errorf("Expected Technology industry, got: %s", result.Industry.Name)
	}

	if result.Confidence != 0.85 {
		t.Errorf("Expected confidence 0.85, got: %f", result.Confidence)
	}

	if len(result.ClassificationCodes) != 1 {
		t.Errorf("Expected 1 classification code, got: %d", len(result.ClassificationCodes))
	}

	// Test with empty content
	result, err = service.DetectIndustryFromContent(ctx, "")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Industry.Name != "General Business" {
		t.Errorf("Expected General Business industry, got: %s", result.Industry.Name)
	}
}

// TestDetectIndustryFromBusinessInfo tests business info-based industry detection
func TestDetectIndustryFromBusinessInfo(t *testing.T) {
	mockRepo := &MockKeywordRepository{
		classifyBusinessResult: &repository.ClassificationResult{
			Industry:   &repository.Industry{ID: 2, Name: "Healthcare"},
			Confidence: 0.90,
			Keywords:   []string{"medical", "healthcare"},
			Reasoning:  "Healthcare keywords detected",
		},
		getClassificationCodesResult: []*repository.ClassificationCode{
			{ID: 2, Code: "621111", Description: "Offices of Physicians"},
		},
	}

	service := NewIndustryDetectionService(mockRepo, log.Default())
	ctx := context.Background()

	result, err := service.DetectIndustryFromBusinessInfo(ctx, "Medical Center", "Healthcare services", "https://medical.com")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Industry.Name != "Healthcare" {
		t.Errorf("Expected Healthcare industry, got: %s", result.Industry.Name)
	}

	if result.Confidence != 0.90 {
		t.Errorf("Expected confidence 0.90, got: %f", result.Confidence)
	}
}

// TestExtractKeywordsFromContent tests keyword extraction from content
func TestExtractKeywordsFromContent(t *testing.T) {
	mockRepo := &MockKeywordRepository{}
	service := NewIndustryDetectionService(mockRepo, log.Default())

	// Test with valid content
	content := "We develop software solutions for businesses using cloud technology"
	keywords := service.extractKeywordsFromContent(content)

	if len(keywords) == 0 {
		t.Error("Expected keywords to be extracted")
	}

	// Debug: print extracted keywords
	t.Logf("Extracted keywords: %v", keywords)

	// Check that common words are filtered out
	hasCommonWord := false
	for _, keyword := range keywords {
		if keyword == "the" || keyword == "for" || keyword == "using" {
			hasCommonWord = true
			break
		}
	}

	if hasCommonWord {
		t.Error("Expected common words to be filtered out")
	}

	// Check that meaningful keywords are extracted
	expectedKeywords := []string{"develop", "software", "solutions", "businesses", "cloud", "technology"}
	hasMeaningfulKeywords := false
	for _, expected := range expectedKeywords {
		for _, keyword := range keywords {
			if keyword == expected {
				hasMeaningfulKeywords = true
				break
			}
		}
		if hasMeaningfulKeywords {
			break
		}
	}

	if !hasMeaningfulKeywords {
		t.Error("Expected meaningful keywords to be extracted")
	}

	// Test with empty content
	keywords = service.extractKeywordsFromContent("")
	if len(keywords) != 0 {
		t.Error("Expected no keywords from empty content")
	}
}

// TestExtractKeywordsFromBusinessInfo tests keyword extraction from business info
func TestExtractKeywordsFromBusinessInfo(t *testing.T) {
	mockRepo := &MockKeywordRepository{}
	service := NewIndustryDetectionService(mockRepo, log.Default())

	// Test with all fields
	keywords := service.extractKeywordsFromBusinessInfo("Tech Solutions Inc", "Software development company", "https://tech-solutions.com")

	if len(keywords) == 0 {
		t.Error("Expected keywords to be extracted")
	}

	// Test with only business name
	keywords = service.extractKeywordsFromBusinessInfo("Healthcare Center", "", "")
	if len(keywords) == 0 {
		t.Error("Expected keywords from business name")
	}

	// Test with empty inputs
	keywords = service.extractKeywordsFromBusinessInfo("", "", "")
	if len(keywords) != 0 {
		t.Error("Expected no keywords from empty inputs")
	}
}

// TestIsCommonWord tests common word filtering
func TestIsCommonWord(t *testing.T) {
	mockRepo := &MockKeywordRepository{}
	service := NewIndustryDetectionService(mockRepo, log.Default())

	// Test common words
	commonWords := []string{"the", "and", "or", "but", "in", "on", "at"}
	for _, word := range commonWords {
		if !service.isCommonWord(word) {
			t.Errorf("Expected '%s' to be identified as common word", word)
		}
	}

	// Test non-common words
	nonCommonWords := []string{"software", "technology", "healthcare", "business"}
	for _, word := range nonCommonWords {
		if service.isCommonWord(word) {
			t.Errorf("Expected '%s' to not be identified as common word", word)
		}
	}
}

// TestBuildEvidenceString tests evidence string building
func TestBuildEvidenceString(t *testing.T) {
	mockRepo := &MockKeywordRepository{}
	service := NewIndustryDetectionService(mockRepo, log.Default())

	// Test with all components
	evidence := service.buildEvidenceString([]string{"software", "technology"}, []string{"tech"}, "Technology keywords detected")

	if evidence == "" {
		t.Error("Expected evidence string to be built")
	}

	if !strings.Contains(evidence, "2 extracted keywords") {
		t.Error("Expected keyword count in evidence")
	}

	if !strings.Contains(evidence, "1 matching industry indicators") {
		t.Error("Expected industry indicator count in evidence")
	}

	if !strings.Contains(evidence, "Technology keywords detected") {
		t.Error("Expected reasoning in evidence")
	}

	// Test with empty keywords
	evidence = service.buildEvidenceString([]string{}, []string{}, "")
	if evidence != "No keywords found for analysis" {
		t.Errorf("Expected default evidence for empty keywords, got: %s", evidence)
	}
}

// TestGetDefaultResult tests default result generation
func TestGetDefaultResult(t *testing.T) {
	mockRepo := &MockKeywordRepository{}
	service := NewIndustryDetectionService(mockRepo, log.Default())

	reason := "Test reason"
	result := service.getDefaultResult(reason)

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Industry.Name != "General Business" {
		t.Errorf("Expected General Business industry, got: %s", result.Industry.Name)
	}

	if result.Confidence != 0.50 {
		t.Errorf("Expected confidence 0.50, got: %f", result.Confidence)
	}

	if result.Evidence != reason {
		t.Errorf("Expected evidence '%s', got: %s", reason, result.Evidence)
	}

	if result.AnalysisMethod != "default_fallback" {
		t.Errorf("Expected analysis method 'default_fallback', got: %s", result.AnalysisMethod)
	}
}
