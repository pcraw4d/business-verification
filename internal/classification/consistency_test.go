package classification

import (
	"context"
	"testing"

	"github.com/pcraw4d/business-verification/internal/classification/repository"
)

// TestKeywordClassificationConsistency tests that keywords and classification codes are properly aligned
func TestKeywordClassificationConsistency(t *testing.T) {
	// Create mock repository with consistent data
	mockRepo := &MockConsistencyRepository{
		industries: map[int]*repository.Industry{
			1: {ID: 1, Name: "Technology", Description: "Software and technology services"},
			2: {ID: 2, Name: "Healthcare", Description: "Medical and health services"},
			3: {ID: 3, Name: "Financial Services", Description: "Banking and financial services"},
		},
		keywords: map[string][]*repository.IndustryKeyword{
			"software": {
				{ID: 1, IndustryID: 1, Keyword: "software", Weight: 0.9},
			},
			"cloud": {
				{ID: 2, IndustryID: 1, Keyword: "cloud", Weight: 0.8},
			},
			"medical": {
				{ID: 3, IndustryID: 2, Keyword: "medical", Weight: 0.9},
			},
			"health": {
				{ID: 4, IndustryID: 2, Keyword: "health", Weight: 0.8},
			},
			"banking": {
				{ID: 5, IndustryID: 3, Keyword: "banking", Weight: 0.9},
			},
			"finance": {
				{ID: 6, IndustryID: 3, Keyword: "finance", Weight: 0.8},
			},
		},
		classificationCodes: map[int][]*repository.ClassificationCode{
			1: { // Technology
				{ID: 1, IndustryID: 1, CodeType: "MCC", Code: "5734", Description: "Computer Software Stores"},
				{ID: 2, IndustryID: 1, CodeType: "SIC", Code: "7372", Description: "Prepackaged Software"},
				{ID: 3, IndustryID: 1, CodeType: "NAICS", Code: "511210", Description: "Software Publishers"},
			},
			2: { // Healthcare
				{ID: 4, IndustryID: 2, CodeType: "MCC", Code: "8011", Description: "Doctors"},
				{ID: 5, IndustryID: 2, CodeType: "SIC", Code: "8011", Description: "Offices and Clinics of Doctors of Medicine"},
				{ID: 6, IndustryID: 2, CodeType: "NAICS", Code: "621111", Description: "Offices of Physicians"},
			},
			3: { // Financial Services
				{ID: 7, IndustryID: 3, CodeType: "MCC", Code: "6011", Description: "Automated Cash Disbursement"},
				{ID: 8, IndustryID: 3, CodeType: "SIC", Code: "6021", Description: "National Commercial Banks"},
				{ID: 9, IndustryID: 3, CodeType: "NAICS", Code: "522110", Description: "Commercial Banking"},
			},
		},
	}

	// Test cases for keyword consistency
	testCases := []struct {
		name              string
		keywords          []string
		expectedIndustry  string
		expectedCodeCount int
	}{
		{
			name:              "Technology Keywords",
			keywords:          []string{"software", "cloud"},
			expectedIndustry:  "Technology",
			expectedCodeCount: 3,
		},
		{
			name:              "Healthcare Keywords",
			keywords:          []string{"medical", "health"},
			expectedIndustry:  "Healthcare",
			expectedCodeCount: 3,
		},
		{
			name:              "Financial Keywords",
			keywords:          []string{"banking", "finance"},
			expectedIndustry:  "Financial Services",
			expectedCodeCount: 3,
		},
		{
			name:              "Mixed Keywords",
			keywords:          []string{"software", "medical"},
			expectedIndustry:  "Technology", // First match
			expectedCodeCount: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			// Test that keywords map to the correct industry
			result, err := mockRepo.ClassifyBusinessByKeywords(ctx, tc.keywords)
			if err != nil {
				t.Fatalf("classification failed: %v", err)
			}

			if result == nil {
				t.Fatal("expected classification result, got nil")
			}

			if result.Industry.Name != tc.expectedIndustry {
				t.Errorf("expected industry %s, got %s", tc.expectedIndustry, result.Industry.Name)
			}

			// Test that the industry has the expected number of classification codes
			codes, err := mockRepo.GetClassificationCodesByIndustry(ctx, result.Industry.ID)
			if err != nil {
				t.Fatalf("failed to get classification codes: %v", err)
			}

			if len(codes) != tc.expectedCodeCount {
				t.Errorf("expected %d classification codes, got %d", tc.expectedCodeCount, len(codes))
			}

			// Verify that all codes belong to the correct industry
			for _, code := range codes {
				if code.IndustryID != result.Industry.ID {
					t.Errorf("classification code %s belongs to industry %d, expected %d",
						code.Code, code.IndustryID, result.Industry.ID)
				}
			}

			t.Logf("✅ %s: keywords %v correctly map to %s with %d classification codes",
				tc.name, tc.keywords, result.Industry.Name, len(codes))
		})
	}
}

// TestClassificationCodeAlignment tests that classification codes are properly aligned with industries
func TestClassificationCodeAlignment(t *testing.T) {
	mockRepo := &MockConsistencyRepository{
		industries: map[int]*repository.Industry{
			1: {ID: 1, Name: "Technology", Description: "Software and technology services"},
			2: {ID: 2, Name: "Healthcare", Description: "Medical and health services"},
		},
		classificationCodes: map[int][]*repository.ClassificationCode{
			1: { // Technology
				{ID: 1, IndustryID: 1, CodeType: "MCC", Code: "5734", Description: "Computer Software Stores"},
				{ID: 2, IndustryID: 1, CodeType: "SIC", Code: "7372", Description: "Prepackaged Software"},
				{ID: 3, IndustryID: 1, CodeType: "NAICS", Code: "511210", Description: "Software Publishers"},
			},
			2: { // Healthcare
				{ID: 4, IndustryID: 2, CodeType: "MCC", Code: "8011", Description: "Doctors"},
				{ID: 5, IndustryID: 2, CodeType: "SIC", Code: "8011", Description: "Offices and Clinics of Doctors of Medicine"},
				{ID: 6, IndustryID: 2, CodeType: "NAICS", Code: "621111", Description: "Offices of Physicians"},
			},
		},
	}

	ctx := context.Background()

	// Test Technology industry codes
	techCodes, err := mockRepo.GetClassificationCodesByIndustry(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get technology codes: %v", err)
	}

	if len(techCodes) != 3 {
		t.Errorf("expected 3 technology codes, got %d", len(techCodes))
	}

	// Verify Technology codes are appropriate
	expectedTechCodes := map[string]string{
		"MCC":   "5734",
		"SIC":   "7372",
		"NAICS": "511210",
	}

	for _, code := range techCodes {
		if expectedCode, exists := expectedTechCodes[code.CodeType]; exists {
			if code.Code != expectedCode {
				t.Errorf("expected %s code %s for Technology, got %s",
					code.CodeType, expectedCode, code.Code)
			}
		}
	}

	// Test Healthcare industry codes
	healthCodes, err := mockRepo.GetClassificationCodesByIndustry(ctx, 2)
	if err != nil {
		t.Fatalf("failed to get healthcare codes: %v", err)
	}

	if len(healthCodes) != 3 {
		t.Errorf("expected 3 healthcare codes, got %d", len(healthCodes))
	}

	// Verify Healthcare codes are appropriate
	expectedHealthCodes := map[string]string{
		"MCC":   "8011",
		"SIC":   "8011",
		"NAICS": "621111",
	}

	for _, code := range healthCodes {
		if expectedCode, exists := expectedHealthCodes[code.CodeType]; exists {
			if code.Code != expectedCode {
				t.Errorf("expected %s code %s for Healthcare, got %s",
					code.CodeType, expectedCode, code.Code)
			}
		}
	}

	t.Logf("✅ All classification codes are properly aligned with their industries")
}

// TestKeywordWeightConsistency tests that keyword weights are consistent across industries
func TestKeywordWeightConsistency(t *testing.T) {
	mockRepo := &MockConsistencyRepository{
		keywords: map[string][]*repository.IndustryKeyword{
			"software": {
				{ID: 1, IndustryID: 1, Keyword: "software", Weight: 0.9},
			},
			"medical": {
				{ID: 2, IndustryID: 2, Keyword: "medical", Weight: 0.9},
			},
			"banking": {
				{ID: 3, IndustryID: 3, Keyword: "banking", Weight: 0.9},
			},
		},
	}

	ctx := context.Background()

	// Test that keywords have consistent weights within their industries
	keywords := []string{"software", "medical", "banking"}

	for _, keyword := range keywords {
		weights, err := mockRepo.GetKeywordWeights(ctx, keyword)
		if err != nil {
			t.Fatalf("failed to get keyword weights for %s: %v", keyword, err)
		}

		if len(weights) == 0 {
			t.Errorf("no weights found for keyword %s", keyword)
			continue
		}

		// All weights for the same keyword should be consistent
		expectedWeight := weights[0].BaseWeight
		for _, weight := range weights {
			if weight.BaseWeight != expectedWeight {
				t.Errorf("inconsistent weight for keyword %s: expected %.2f, got %.2f",
					keyword, expectedWeight, weight.BaseWeight)
			}
		}

		t.Logf("✅ Keyword '%s' has consistent weights: %.2f", keyword, expectedWeight)
	}
}

// MockConsistencyRepository for testing consistency
type MockConsistencyRepository struct {
	industries          map[int]*repository.Industry
	keywords            map[string][]*repository.IndustryKeyword
	classificationCodes map[int][]*repository.ClassificationCode
}

func (m *MockConsistencyRepository) GetIndustryByID(ctx context.Context, id int) (*repository.Industry, error) {
	if industry, exists := m.industries[id]; exists {
		return industry, nil
	}
	return nil, nil
}

func (m *MockConsistencyRepository) GetIndustryByName(ctx context.Context, name string) (*repository.Industry, error) {
	for _, industry := range m.industries {
		if industry.Name == name {
			return industry, nil
		}
	}
	return nil, nil
}

func (m *MockConsistencyRepository) ListIndustries(ctx context.Context, category string) ([]*repository.Industry, error) {
	var result []*repository.Industry
	for _, industry := range m.industries {
		result = append(result, industry)
	}
	return result, nil
}

func (m *MockConsistencyRepository) GetKeywordsByIndustry(ctx context.Context, industryID int) ([]*repository.IndustryKeyword, error) {
	var result []*repository.IndustryKeyword
	for _, keywords := range m.keywords {
		for _, keyword := range keywords {
			if keyword.IndustryID == industryID {
				result = append(result, keyword)
			}
		}
	}
	return result, nil
}

func (m *MockConsistencyRepository) SearchKeywords(ctx context.Context, query string, limit int) ([]*repository.IndustryKeyword, error) {
	var result []*repository.IndustryKeyword
	for keyword, keywords := range m.keywords {
		if keyword == query {
			for i, kw := range keywords {
				if i >= limit {
					break
				}
				result = append(result, kw)
			}
		}
	}
	return result, nil
}

func (m *MockConsistencyRepository) GetClassificationCodesByIndustry(ctx context.Context, industryID int) ([]*repository.ClassificationCode, error) {
	if codes, exists := m.classificationCodes[industryID]; exists {
		return codes, nil
	}
	return []*repository.ClassificationCode{}, nil
}

func (m *MockConsistencyRepository) GetTopIndustriesByKeywords(ctx context.Context, keywords []string, limit int) ([]*repository.Industry, error) {
	var result []*repository.Industry
	for _, industry := range m.industries {
		if len(result) >= limit {
			break
		}
		result = append(result, industry)
	}
	return result, nil
}

func (m *MockConsistencyRepository) ClassifyBusinessByKeywords(ctx context.Context, keywords []string) (*repository.ClassificationResult, error) {
	// Simple classification logic for testing
	keywordToIndustry := map[string]string{
		"software": "Technology",
		"cloud":    "Technology",
		"medical":  "Healthcare",
		"health":   "Healthcare",
		"banking":  "Financial Services",
		"finance":  "Financial Services",
	}

	// Find the first matching keyword
	for _, keyword := range keywords {
		if industryName, exists := keywordToIndustry[keyword]; exists {
			for _, industry := range m.industries {
				if industry.Name == industryName {
					// Convert []*ClassificationCode to []ClassificationCode
					var codes []repository.ClassificationCode
					if codesList, exists := m.classificationCodes[industry.ID]; exists {
						for _, code := range codesList {
							codes = append(codes, *code)
						}
					}

					return &repository.ClassificationResult{
						Industry:   industry,
						Confidence: 0.85,
						Keywords:   []string{keyword},
						Codes:      codes,
					}, nil
				}
			}
		}
	}

	// Default fallback
	if len(m.industries) > 0 {
		for _, industry := range m.industries {
			// Convert []*ClassificationCode to []ClassificationCode
			var codes []repository.ClassificationCode
			if codesList, exists := m.classificationCodes[industry.ID]; exists {
				for _, code := range codesList {
					codes = append(codes, *code)
				}
			}

			return &repository.ClassificationResult{
				Industry:   industry,
				Confidence: 0.5,
				Keywords:   []string{"business"},
				Codes:      codes,
			}, nil
		}
	}

	return nil, nil
}

func (m *MockConsistencyRepository) GetKeywordWeights(ctx context.Context, keyword string) ([]*repository.KeywordWeight, error) {
	var result []*repository.KeywordWeight
	for keywordName, keywords := range m.keywords {
		if keywordName == keyword {
			for _, kw := range keywords {
				result = append(result, &repository.KeywordWeight{
					ID:         kw.ID,
					Keyword:    kw.Keyword,
					IndustryID: kw.IndustryID,
					BaseWeight: kw.Weight,
				})
			}
		}
	}
	return result, nil
}

// Implement remaining interface methods with minimal implementations
func (m *MockConsistencyRepository) CreateIndustry(ctx context.Context, industry *repository.Industry) error {
	return nil
}
func (m *MockConsistencyRepository) UpdateIndustry(ctx context.Context, industry *repository.Industry) error {
	return nil
}
func (m *MockConsistencyRepository) DeleteIndustry(ctx context.Context, id int) error { return nil }
func (m *MockConsistencyRepository) AddKeywordToIndustry(ctx context.Context, industryID int, keyword string, weight float64) error {
	return nil
}
func (m *MockConsistencyRepository) UpdateKeywordWeight(ctx context.Context, keywordID int, weight float64) error {
	return nil
}
func (m *MockConsistencyRepository) RemoveKeywordFromIndustry(ctx context.Context, keywordID int) error {
	return nil
}
func (m *MockConsistencyRepository) GetClassificationCodesByType(ctx context.Context, codeType string) ([]*repository.ClassificationCode, error) {
	return nil, nil
}
func (m *MockConsistencyRepository) AddClassificationCode(ctx context.Context, code *repository.ClassificationCode) error {
	return nil
}
func (m *MockConsistencyRepository) UpdateClassificationCode(ctx context.Context, code *repository.ClassificationCode) error {
	return nil
}
func (m *MockConsistencyRepository) DeleteClassificationCode(ctx context.Context, id int) error {
	return nil
}
func (m *MockConsistencyRepository) GetPatternsByIndustry(ctx context.Context, industryID int) ([]*repository.IndustryPattern, error) {
	return nil, nil
}
func (m *MockConsistencyRepository) AddPattern(ctx context.Context, pattern *repository.IndustryPattern) error {
	return nil
}
func (m *MockConsistencyRepository) UpdatePattern(ctx context.Context, pattern *repository.IndustryPattern) error {
	return nil
}
func (m *MockConsistencyRepository) DeletePattern(ctx context.Context, id int) error { return nil }
func (m *MockConsistencyRepository) UpdateKeywordWeightByID(ctx context.Context, weight *repository.KeywordWeight) error {
	return nil
}
func (m *MockConsistencyRepository) IncrementUsageCount(ctx context.Context, keyword string, industryID int) error {
	return nil
}
func (m *MockConsistencyRepository) ClassifyBusiness(ctx context.Context, businessName, description, websiteURL string) (*repository.ClassificationResult, error) {
	return nil, nil
}
func (m *MockConsistencyRepository) SearchIndustriesByPattern(ctx context.Context, pattern string) ([]*repository.Industry, error) {
	return nil, nil
}
func (m *MockConsistencyRepository) GetIndustryStatistics(ctx context.Context) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockConsistencyRepository) GetKeywordFrequency(ctx context.Context, industryID int) (map[string]int, error) {
	return nil, nil
}
func (m *MockConsistencyRepository) BulkInsertKeywords(ctx context.Context, keywords []*repository.IndustryKeyword) error {
	return nil
}
func (m *MockConsistencyRepository) BulkUpdateKeywords(ctx context.Context, keywords []*repository.IndustryKeyword) error {
	return nil
}
func (m *MockConsistencyRepository) BulkDeleteKeywords(ctx context.Context, keywordIDs []int) error {
	return nil
}
func (m *MockConsistencyRepository) Ping(ctx context.Context) error { return nil }
func (m *MockConsistencyRepository) GetDatabaseStats(ctx context.Context) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockConsistencyRepository) CleanupInactiveData(ctx context.Context) error { return nil }
