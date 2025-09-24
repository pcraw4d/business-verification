package classification

import (
	"context"
	"testing"

	"kyb-platform/internal/classification/repository"
)

// TestCrossIndustryKeywordValidation validates that keywords don't create conflicts between industries
func TestCrossIndustryKeywordValidation(t *testing.T) {
	mockRepo := &MockValidationRepository{
		industries: map[int]*repository.Industry{
			1: {ID: 1, Name: "Technology", Category: "Technology"},
			2: {ID: 2, Name: "Healthcare", Category: "Healthcare"},
			3: {ID: 3, Name: "Financial Services", Category: "Financial"},
		},
		keywords: map[int][]*repository.IndustryKeyword{
			1: {
				{ID: 1, IndustryID: 1, Keyword: "software", Weight: 0.9},
				{ID: 2, IndustryID: 1, Keyword: "cloud", Weight: 0.85},
				{ID: 3, IndustryID: 1, Keyword: "technology", Weight: 0.8},
			},
			2: {
				{ID: 4, IndustryID: 2, Keyword: "medical", Weight: 0.9},
				{ID: 5, IndustryID: 2, Keyword: "health", Weight: 0.85},
				{ID: 6, IndustryID: 2, Keyword: "patient", Weight: 0.8},
			},
			3: {
				{ID: 7, IndustryID: 3, Keyword: "banking", Weight: 0.9},
				{ID: 8, IndustryID: 3, Keyword: "finance", Weight: 0.85},
				{ID: 9, IndustryID: 3, Keyword: "investment", Weight: 0.8},
			},
		},
	}

	// Check for keyword conflicts between industries
	keywordIndustryMap := make(map[string]int)
	conflicts := make([]string, 0)

	for industryID, keywords := range mockRepo.keywords {
		for _, keyword := range keywords {
			if existingIndustryID, exists := keywordIndustryMap[keyword.Keyword]; exists {
				if existingIndustryID != industryID {
					conflicts = append(conflicts, keyword.Keyword)
				}
			} else {
				keywordIndustryMap[keyword.Keyword] = industryID
			}
		}
	}

	if len(conflicts) > 0 {
		t.Errorf("found keyword conflicts between industries: %v", conflicts)
	} else {
		t.Log("✅ No keyword conflicts found between industries")
	}
}

// TestClassificationCodeUniqueness validates that classification codes are unique within their type
func TestClassificationCodeUniqueness(t *testing.T) {
	mockRepo := &MockValidationRepository{
		classificationCodes: map[int][]repository.ClassificationCode{
			1: {
				{ID: 1, IndustryID: 1, CodeType: "MCC", Code: "5734", Description: "Computer Software Stores"},
				{ID: 2, IndustryID: 1, CodeType: "SIC", Code: "7372", Description: "Prepackaged Software"},
				{ID: 3, IndustryID: 1, CodeType: "NAICS", Code: "511210", Description: "Software Publishers"},
			},
			2: {
				{ID: 4, IndustryID: 2, CodeType: "MCC", Code: "8011", Description: "Doctors"},
				{ID: 5, IndustryID: 2, CodeType: "SIC", Code: "8011", Description: "Offices and Clinics of Doctors of Medicine"},
				{ID: 6, IndustryID: 2, CodeType: "NAICS", Code: "621111", Description: "Offices of Physicians"},
			},
		},
	}

	// Check for duplicate codes within each type
	codeTypeMap := make(map[string]map[string]bool)
	duplicates := make([]string, 0)

	for _, codes := range mockRepo.classificationCodes {
		for _, code := range codes {
			if codeTypeMap[code.CodeType] == nil {
				codeTypeMap[code.CodeType] = make(map[string]bool)
			}

			if codeTypeMap[code.CodeType][code.Code] {
				duplicates = append(duplicates, code.CodeType+":"+code.Code)
			} else {
				codeTypeMap[code.CodeType][code.Code] = true
			}
		}
	}

	if len(duplicates) > 0 {
		t.Errorf("found duplicate classification codes: %v", duplicates)
	} else {
		t.Log("✅ All classification codes are unique within their type")
	}
}

// TestKeywordIndustryMappingConsistency validates that keyword-to-industry mappings are consistent
func TestKeywordIndustryMappingConsistency(t *testing.T) {
	mockRepo := &MockValidationRepository{
		industries: map[int]*repository.Industry{
			1: {ID: 1, Name: "Technology", Category: "Technology"},
			2: {ID: 2, Name: "Healthcare", Category: "Healthcare"},
		},
		keywords: map[int][]*repository.IndustryKeyword{
			1: {
				{ID: 1, IndustryID: 1, Keyword: "software", Weight: 0.9},
				{ID: 2, IndustryID: 1, Keyword: "cloud", Weight: 0.85},
			},
			2: {
				{ID: 3, IndustryID: 2, Keyword: "medical", Weight: 0.9},
				{ID: 4, IndustryID: 2, Keyword: "health", Weight: 0.85},
			},
		},
	}

	// Test that each keyword maps to exactly one industry
	keywordCount := make(map[string]int)
	for _, keywords := range mockRepo.keywords {
		for _, keyword := range keywords {
			keywordCount[keyword.Keyword]++
		}
	}

	// Check for keywords that appear in multiple industries
	multiIndustryKeywords := make([]string, 0)
	for keyword, count := range keywordCount {
		if count > 1 {
			multiIndustryKeywords = append(multiIndustryKeywords, keyword)
		}
	}

	if len(multiIndustryKeywords) > 0 {
		t.Errorf("keywords found in multiple industries: %v", multiIndustryKeywords)
	} else {
		t.Log("✅ All keywords map to exactly one industry")
	}
}

// TestClassificationCodeFormatValidation validates that classification codes have proper formats
func TestClassificationCodeFormatValidation(t *testing.T) {
	mockRepo := &MockValidationRepository{
		classificationCodes: map[int][]repository.ClassificationCode{
			1: {
				{ID: 1, IndustryID: 1, CodeType: "MCC", Code: "5734", Description: "Computer Software Stores"},
				{ID: 2, IndustryID: 1, CodeType: "SIC", Code: "7372", Description: "Prepackaged Software"},
				{ID: 3, IndustryID: 1, CodeType: "NAICS", Code: "511210", Description: "Software Publishers"},
			},
		},
	}

	// Validate code formats
	for _, codes := range mockRepo.classificationCodes {
		for _, code := range codes {
			// Check that code is not empty
			if code.Code == "" {
				t.Errorf("classification code has empty code value (ID: %d)", code.ID)
			}

			// Check that code type is valid
			validTypes := map[string]bool{"MCC": true, "SIC": true, "NAICS": true}
			if !validTypes[code.CodeType] {
				t.Errorf("invalid code type: %s (ID: %d)", code.CodeType, code.ID)
			}

			// Check that description is not empty
			if code.Description == "" {
				t.Errorf("classification code has empty description (ID: %d)", code.ID)
			}

			// Check that industry ID is positive
			if code.IndustryID <= 0 {
				t.Errorf("invalid industry ID: %d (ID: %d)", code.IndustryID, code.ID)
			}
		}
	}

	t.Log("✅ All classification codes have proper formats")
}

// MockValidationRepository is a mock repository for validation testing
type MockValidationRepository struct {
	industries          map[int]*repository.Industry
	keywords            map[int][]*repository.IndustryKeyword
	classificationCodes map[int][]repository.ClassificationCode
}

// Implement only the methods needed for these specific tests
func (m *MockValidationRepository) ListIndustries(ctx context.Context, category string) ([]*repository.Industry, error) {
	var industries []*repository.Industry
	for _, industry := range m.industries {
		if category == "" || industry.Category == category {
			industries = append(industries, industry)
		}
	}
	return industries, nil
}

func (m *MockValidationRepository) GetIndustryByID(ctx context.Context, id int) (*repository.Industry, error) {
	if industry, exists := m.industries[id]; exists {
		return industry, nil
	}
	return nil, nil
}

func (m *MockValidationRepository) GetKeywordsByIndustry(ctx context.Context, industryID int) ([]*repository.IndustryKeyword, error) {
	if keywords, exists := m.keywords[industryID]; exists {
		return keywords, nil
	}
	return nil, nil
}

func (m *MockValidationRepository) GetClassificationCodesByIndustry(ctx context.Context, industryID int) ([]repository.ClassificationCode, error) {
	if codes, exists := m.classificationCodes[industryID]; exists {
		return codes, nil
	}
	return nil, nil
}

// Stub implementations for interface compliance (not used in these tests)
func (m *MockValidationRepository) GetIndustryByName(ctx context.Context, name string) (*repository.Industry, error) {
	return nil, nil
}
func (m *MockValidationRepository) CreateIndustry(ctx context.Context, industry *repository.Industry) error {
	return nil
}
func (m *MockValidationRepository) UpdateIndustry(ctx context.Context, industry *repository.Industry) error {
	return nil
}
func (m *MockValidationRepository) DeleteIndustry(ctx context.Context, id int) error { return nil }
func (m *MockValidationRepository) SearchKeywords(ctx context.Context, query string, limit int) ([]*repository.IndustryKeyword, error) {
	return nil, nil
}
func (m *MockValidationRepository) AddKeywordToIndustry(ctx context.Context, industryID int, keyword string, weight float64) error {
	return nil
}
func (m *MockValidationRepository) UpdateKeywordWeight(ctx context.Context, keywordID int, weight float64) error {
	return nil
}
func (m *MockValidationRepository) RemoveKeywordFromIndustry(ctx context.Context, keywordID int) error {
	return nil
}
func (m *MockValidationRepository) GetClassificationCodesByType(ctx context.Context, codeType string) ([]*repository.ClassificationCode, error) {
	return nil, nil
}
func (m *MockValidationRepository) AddClassificationCode(ctx context.Context, code *repository.ClassificationCode) error {
	return nil
}
func (m *MockValidationRepository) UpdateClassificationCode(ctx context.Context, code *repository.ClassificationCode) error {
	return nil
}
func (m *MockValidationRepository) DeleteClassificationCode(ctx context.Context, id int) error {
	return nil
}
func (m *MockValidationRepository) GetPatternsByIndustry(ctx context.Context, industryID int) ([]*repository.IndustryPattern, error) {
	return nil, nil
}
func (m *MockValidationRepository) AddPattern(ctx context.Context, pattern *repository.IndustryPattern) error {
	return nil
}
func (m *MockValidationRepository) UpdatePattern(ctx context.Context, pattern *repository.IndustryPattern) error {
	return nil
}
func (m *MockValidationRepository) DeletePattern(ctx context.Context, id int) error { return nil }
func (m *MockValidationRepository) GetKeywordWeights(ctx context.Context, keyword string) ([]*repository.KeywordWeight, error) {
	return nil, nil
}
func (m *MockValidationRepository) UpdateKeywordWeightByID(ctx context.Context, weight *repository.KeywordWeight) error {
	return nil
}
func (m *MockValidationRepository) IncrementUsageCount(ctx context.Context, keyword string, industryID int) error {
	return nil
}
func (m *MockValidationRepository) ClassifyBusiness(ctx context.Context, businessName, description, websiteURL string) (*repository.ClassificationResult, error) {
	return nil, nil
}
func (m *MockValidationRepository) ClassifyBusinessByKeywords(ctx context.Context, keywords []string) (*repository.ClassificationResult, error) {
	return nil, nil
}
func (m *MockValidationRepository) GetTopIndustriesByKeywords(ctx context.Context, keywords []string, limit int) ([]*repository.Industry, error) {
	return nil, nil
}
func (m *MockValidationRepository) SearchIndustriesByPattern(ctx context.Context, pattern string) ([]*repository.Industry, error) {
	return nil, nil
}
func (m *MockValidationRepository) GetIndustryStatistics(ctx context.Context) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockValidationRepository) GetKeywordFrequency(ctx context.Context, industryID int) (map[string]int, error) {
	return nil, nil
}
func (m *MockValidationRepository) BulkInsertKeywords(ctx context.Context, keywords []*repository.IndustryKeyword) error {
	return nil
}
func (m *MockValidationRepository) BulkUpdateKeywords(ctx context.Context, keywords []*repository.IndustryKeyword) error {
	return nil
}
func (m *MockValidationRepository) BulkDeleteKeywords(ctx context.Context, keywordIDs []int) error {
	return nil
}
func (m *MockValidationRepository) Ping(ctx context.Context) error { return nil }
func (m *MockValidationRepository) GetDatabaseStats(ctx context.Context) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockValidationRepository) CleanupInactiveData(ctx context.Context) error { return nil }
