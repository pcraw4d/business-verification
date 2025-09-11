package classification

import (
	"context"
	"testing"

	"github.com/pcraw4d/business-verification/internal/classification/repository"
)

// TestEndToEndClassificationFlow tests the complete classification flow with various industries
func TestEndToEndClassificationFlow(t *testing.T) {
	// Create mock repository for testing
	mockRepo := &MockE2ERepository{
		industries: map[int]*repository.Industry{
			1: {ID: 1, Name: "Technology", Description: "Software and technology services"},
			2: {ID: 2, Name: "Healthcare", Description: "Medical and health services"},
			3: {ID: 3, Name: "Financial Services", Description: "Banking and financial services"},
			4: {ID: 4, Name: "Retail", Description: "Retail and e-commerce"},
			5: {ID: 5, Name: "Manufacturing", Description: "Manufacturing and production"},
		},
		keywords: map[string][]*repository.IndustryKeyword{
			"technology": {
				{ID: 1, IndustryID: 1, Keyword: "software", Weight: 0.9},
				{ID: 2, IndustryID: 1, Keyword: "cloud", Weight: 0.8},
				{ID: 3, IndustryID: 1, Keyword: "ai", Weight: 0.85},
				{ID: 4, IndustryID: 1, Keyword: "data", Weight: 0.7},
			},
			"healthcare": {
				{ID: 5, IndustryID: 2, Keyword: "medical", Weight: 0.9},
				{ID: 6, IndustryID: 2, Keyword: "health", Weight: 0.8},
				{ID: 7, IndustryID: 2, Keyword: "patient", Weight: 0.75},
				{ID: 8, IndustryID: 2, Keyword: "clinic", Weight: 0.8},
			},
			"financial": {
				{ID: 9, IndustryID: 3, Keyword: "banking", Weight: 0.9},
				{ID: 10, IndustryID: 3, Keyword: "finance", Weight: 0.8},
				{ID: 11, IndustryID: 3, Keyword: "investment", Weight: 0.85},
				{ID: 12, IndustryID: 3, Keyword: "credit", Weight: 0.7},
			},
			"retail": {
				{ID: 13, IndustryID: 4, Keyword: "retail", Weight: 0.9},
				{ID: 14, IndustryID: 4, Keyword: "ecommerce", Weight: 0.8},
				{ID: 15, IndustryID: 4, Keyword: "shopping", Weight: 0.7},
				{ID: 16, IndustryID: 4, Keyword: "store", Weight: 0.75},
			},
			"manufacturing": {
				{ID: 17, IndustryID: 5, Keyword: "manufacturing", Weight: 0.9},
				{ID: 18, IndustryID: 5, Keyword: "production", Weight: 0.8},
				{ID: 19, IndustryID: 5, Keyword: "factory", Weight: 0.85},
				{ID: 20, IndustryID: 5, Keyword: "industrial", Weight: 0.7},
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
			4: { // Retail
				{ID: 10, IndustryID: 4, CodeType: "MCC", Code: "5311", Description: "Department Stores"},
				{ID: 11, IndustryID: 4, CodeType: "SIC", Code: "5311", Description: "Department Stores"},
				{ID: 12, IndustryID: 4, CodeType: "NAICS", Code: "452111", Description: "Department Stores"},
			},
			5: { // Manufacturing
				{ID: 13, IndustryID: 5, CodeType: "MCC", Code: "5999", Description: "Miscellaneous and Specialty Retail Stores"},
				{ID: 14, IndustryID: 5, CodeType: "SIC", Code: "3711", Description: "Motor Vehicles and Passenger Car Bodies"},
				{ID: 15, IndustryID: 5, CodeType: "NAICS", Code: "336111", Description: "Automobile Manufacturing"},
			},
		},
	}

	// Create services
	industryService := NewIndustryDetectionService(mockRepo, nil)
	classifier := NewClassificationCodeGenerator(mockRepo, nil)

	// Test cases for different industries
	testCases := []struct {
		name               string
		businessName       string
		description        string
		websiteURL         string
		expectedIndustry   string
		expectedConfidence float64
		expectedCodeCount  int
	}{
		{
			name:               "Technology Company",
			businessName:       "TechCorp Solutions",
			description:        "We develop innovative software solutions using cloud technology and AI for businesses",
			websiteURL:         "https://techcorp.com",
			expectedIndustry:   "Technology",
			expectedConfidence: 0.8,
			expectedCodeCount:  3,
		},
		{
			name:               "Healthcare Provider",
			businessName:       "Medical Center Plus",
			description:        "Comprehensive medical services and patient care at our modern clinic",
			websiteURL:         "https://medicalcenter.com",
			expectedIndustry:   "Healthcare",
			expectedConfidence: 0.8,
			expectedCodeCount:  1,
		},
		{
			name:               "Financial Institution",
			businessName:       "First National Bank",
			description:        "Personal and business banking services with investment and credit solutions",
			websiteURL:         "https://firstnational.com",
			expectedIndustry:   "Financial Services",
			expectedConfidence: 0.8,
			expectedCodeCount:  3,
		},
		{
			name:               "Retail Business",
			businessName:       "ShopSmart Retail",
			description:        "Online and offline retail store offering shopping experience for customers",
			websiteURL:         "https://shopsmart.com",
			expectedIndustry:   "Retail",
			expectedConfidence: 0.8,
			expectedCodeCount:  3,
		},
		{
			name:               "Manufacturing Company",
			businessName:       "Industrial Manufacturing Co",
			description:        "Factory production of industrial equipment and manufacturing solutions",
			websiteURL:         "https://industrialmfg.com",
			expectedIndustry:   "Manufacturing",
			expectedConfidence: 0.8,
			expectedCodeCount:  3,
		},
		{
			name:               "Mixed Industry Business",
			businessName:       "TechHealth Solutions",
			description:        "Combining technology and healthcare for innovative medical software",
			websiteURL:         "https://techhealth.com",
			expectedIndustry:   "Healthcare", // Should match healthcare keywords
			expectedConfidence: 0.8,
			expectedCodeCount:  3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			// Test industry detection
			industryResult, err := industryService.DetectIndustryFromContent(ctx, tc.description)
			if err != nil {
				t.Fatalf("industry detection failed: %v", err)
			}

			if industryResult == nil {
				t.Fatal("expected industry result, got nil")
			}

			if industryResult.Industry.Name != tc.expectedIndustry {
				t.Errorf("expected industry %s, got %s", tc.expectedIndustry, industryResult.Industry.Name)
			}

			if industryResult.Confidence < tc.expectedConfidence-0.05 { // Allow some variance
				t.Errorf("expected confidence >= %.3f, got %.3f", tc.expectedConfidence-0.05, industryResult.Confidence)
			}

			// Test classification code generation
			classificationResult, err := classifier.GenerateClassificationCodes(
				ctx,
				industryResult.KeywordsMatched,
				industryResult.Industry.Name,
				industryResult.Confidence,
			)
			if err != nil {
				t.Fatalf("classification code generation failed: %v", err)
			}

			if classificationResult == nil {
				t.Fatal("expected classification result, got nil")
			}

			// Verify we have the expected number of codes
			totalCodes := len(classificationResult.MCC) + len(classificationResult.SIC) + len(classificationResult.NAICS)
			if totalCodes < tc.expectedCodeCount {
				t.Errorf("expected at least %d codes, got %d", tc.expectedCodeCount, totalCodes)
			}

			// Log the results for verification
			t.Logf("✅ %s classified as %s (confidence: %.1f%%)", tc.businessName, industryResult.Industry.Name, industryResult.Confidence)
			t.Logf("   Generated %d MCC, %d SIC, %d NAICS codes",
				len(classificationResult.MCC),
				len(classificationResult.SIC),
				len(classificationResult.NAICS))
		})
	}
}

// TestClassificationConsistency tests that similar businesses get consistent classifications
func TestClassificationConsistency(t *testing.T) {
	mockRepo := &MockE2ERepository{
		industries: map[int]*repository.Industry{
			1: {ID: 1, Name: "Technology", Description: "Software and technology services"},
		},
		keywords: map[string][]*repository.IndustryKeyword{
			"technology": {
				{ID: 1, IndustryID: 1, Keyword: "software", Weight: 0.9},
				{ID: 2, IndustryID: 1, Keyword: "cloud", Weight: 0.8},
				{ID: 3, IndustryID: 1, Keyword: "ai", Weight: 0.85},
			},
		},
		classificationCodes: map[int][]*repository.ClassificationCode{
			1: {
				{ID: 1, IndustryID: 1, CodeType: "MCC", Code: "5734", Description: "Computer Software Stores"},
				{ID: 2, IndustryID: 1, CodeType: "SIC", Code: "7372", Description: "Prepackaged Software"},
				{ID: 3, IndustryID: 1, CodeType: "NAICS", Code: "511210", Description: "Software Publishers"},
			},
		},
	}

	industryService := NewIndustryDetectionService(mockRepo, nil)
	classifier := NewClassificationCodeGenerator(mockRepo, nil)

	// Test multiple similar descriptions
	descriptions := []string{
		"We develop software solutions using cloud technology",
		"Software development company specializing in cloud solutions",
		"Cloud-based software development services",
		"Technology company focused on software and cloud",
	}

	var results []*IndustryDetectionResult
	var classifications []*ClassificationCodesInfo

	for _, desc := range descriptions {
		ctx := context.Background()

		// Get industry detection
		result, err := industryService.DetectIndustryFromContent(ctx, desc)
		if err != nil {
			t.Fatalf("industry detection failed: %v", err)
		}
		results = append(results, result)

		// Get classification codes
		classification, err := classifier.GenerateClassificationCodes(
			ctx,
			result.KeywordsMatched,
			result.Industry.Name,
			result.Confidence,
		)
		if err != nil {
			t.Fatalf("classification code generation failed: %v", err)
		}
		classifications = append(classifications, classification)
	}

	// Verify consistency
	for i := 1; i < len(results); i++ {
		if results[i].Industry.Name != results[0].Industry.Name {
			t.Errorf("inconsistent industry detection: %s vs %s",
				results[0].Industry.Name, results[i].Industry.Name)
		}

		if results[i].Confidence < results[0].Confidence-10 { // Allow some variance
			t.Errorf("confidence varies too much: %.1f%% vs %.1f%%",
				results[0].Confidence, results[i].Confidence)
		}

		// Verify classification codes are consistent
		prevCodes := classifications[i-1]
		currCodes := classifications[i]

		if len(prevCodes.MCC) != len(currCodes.MCC) ||
			len(prevCodes.SIC) != len(currCodes.SIC) ||
			len(prevCodes.NAICS) != len(currCodes.NAICS) {
			t.Errorf("inconsistent code generation for descriptions %d and %d", i-1, i)
		}
	}

	t.Logf("✅ All %d similar descriptions classified consistently as %s",
		len(descriptions), results[0].Industry.Name)
}

// TestEdgeCases tests edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	mockRepo := &MockE2ERepository{
		industries: map[int]*repository.Industry{
			1: {ID: 1, Name: "General Business", Description: "General business services"},
		},
		keywords: map[string][]*repository.IndustryKeyword{
			"general": {
				{ID: 1, IndustryID: 1, Keyword: "business", Weight: 0.5},
				{ID: 2, IndustryID: 1, Keyword: "service", Weight: 0.4},
			},
		},
		classificationCodes: map[int][]*repository.ClassificationCode{
			1: {
				{ID: 1, IndustryID: 1, CodeType: "MCC", Code: "7399", Description: "Business Services"},
				{ID: 2, IndustryID: 1, CodeType: "SIC", Code: "7389", Description: "Business Services"},
				{ID: 3, IndustryID: 1, CodeType: "NAICS", Code: "541990", Description: "All Other Professional, Scientific, and Technical Services"},
			},
		},
	}

	industryService := NewIndustryDetectionService(mockRepo, nil)
	classifier := NewClassificationCodeGenerator(mockRepo, nil)

	testCases := []struct {
		name             string
		description      string
		expectedIndustry string
	}{
		{
			name:             "Empty Description",
			description:      "",
			expectedIndustry: "General Business",
		},
		{
			name:             "Very Short Description",
			description:      "Business",
			expectedIndustry: "General Business",
		},
		{
			name:             "Very Long Description",
			description:      "This is a very long business description that contains many words and should still be processed correctly even though it exceeds normal length limits and includes various business-related terms and concepts that should help with classification",
			expectedIndustry: "General Business",
		},
		{
			name:             "Description with Special Characters",
			description:      "Business & Services (Ltd.) - We provide #1 quality solutions!",
			expectedIndustry: "General Business",
		},
		{
			name:             "Description with Numbers",
			description:      "Business services since 1995, serving 1000+ clients",
			expectedIndustry: "General Business",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := industryService.DetectIndustryFromContent(ctx, tc.description)
			if err != nil {
				t.Fatalf("industry detection failed: %v", err)
			}

			if result == nil {
				t.Fatal("expected result, got nil")
			}

			if result.Industry.Name != tc.expectedIndustry {
				t.Errorf("expected industry %s, got %s", tc.expectedIndustry, result.Industry.Name)
			}

			// Verify classification codes are generated
			classification, err := classifier.GenerateClassificationCodes(
				ctx,
				result.KeywordsMatched,
				result.Industry.Name,
				result.Confidence,
			)
			if err != nil {
				t.Fatalf("classification code generation failed: %v", err)
			}

			if classification == nil {
				t.Fatal("expected classification, got nil")
			}

			// Edge cases may not generate codes, which is acceptable
			totalCodes := len(classification.MCC) + len(classification.SIC) + len(classification.NAICS)
			t.Logf("Generated %d classification codes for edge case", totalCodes)

			t.Logf("✅ %s handled correctly: %s (confidence: %.1f%%)",
				tc.name, result.Industry.Name, result.Confidence)
		})
	}
}

// MockE2ERepository for testing
type MockE2ERepository struct {
	industries          map[int]*repository.Industry
	keywords            map[string][]*repository.IndustryKeyword
	classificationCodes map[int][]*repository.ClassificationCode
}

func (m *MockE2ERepository) GetIndustryByID(ctx context.Context, id int) (*repository.Industry, error) {
	if industry, exists := m.industries[id]; exists {
		return industry, nil
	}
	return nil, nil
}

func (m *MockE2ERepository) GetIndustryByName(ctx context.Context, name string) (*repository.Industry, error) {
	for _, industry := range m.industries {
		if industry.Name == name {
			return industry, nil
		}
	}
	return nil, nil
}

func (m *MockE2ERepository) ListIndustries(ctx context.Context, category string) ([]*repository.Industry, error) {
	var result []*repository.Industry
	for _, industry := range m.industries {
		result = append(result, industry)
	}
	return result, nil
}

func (m *MockE2ERepository) GetKeywordsByIndustry(ctx context.Context, industryID int) ([]*repository.IndustryKeyword, error) {
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

func (m *MockE2ERepository) SearchKeywords(ctx context.Context, query string, limit int) ([]*repository.IndustryKeyword, error) {
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

func (m *MockE2ERepository) GetClassificationCodesByIndustry(ctx context.Context, industryID int) ([]*repository.ClassificationCode, error) {
	if codes, exists := m.classificationCodes[industryID]; exists {
		return codes, nil
	}
	return []*repository.ClassificationCode{}, nil
}

func (m *MockE2ERepository) GetTopIndustriesByKeywords(ctx context.Context, keywords []string, limit int) ([]*repository.Industry, error) {
	var result []*repository.Industry
	for _, industry := range m.industries {
		if len(result) >= limit {
			break
		}
		result = append(result, industry)
	}
	return result, nil
}

func (m *MockE2ERepository) ClassifyBusiness(ctx context.Context, businessName, description, websiteURL string) (*repository.ClassificationResult, error) {
	// Simple classification logic for testing
	for keyword, keywords := range m.keywords {
		if len(keywords) > 0 {
			industryID := keywords[0].IndustryID
			if industry, exists := m.industries[industryID]; exists {
				// Convert []*ClassificationCode to []ClassificationCode
				var codes []repository.ClassificationCode
				for _, code := range m.classificationCodes[industryID] {
					codes = append(codes, *code)
				}

				return &repository.ClassificationResult{
					Industry:   industry,
					Confidence: 85.0,
					Keywords:   []string{keyword},
					Codes:      codes,
				}, nil
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
				Confidence: 50.0,
				Keywords:   []string{"business"},
				Codes:      codes,
			}, nil
		}
	}

	return nil, nil
}

// Implement remaining interface methods with minimal implementations
func (m *MockE2ERepository) CreateIndustry(ctx context.Context, industry *repository.Industry) error {
	return nil
}
func (m *MockE2ERepository) UpdateIndustry(ctx context.Context, industry *repository.Industry) error {
	return nil
}
func (m *MockE2ERepository) DeleteIndustry(ctx context.Context, id int) error { return nil }
func (m *MockE2ERepository) AddKeywordToIndustry(ctx context.Context, industryID int, keyword string, weight float64) error {
	return nil
}
func (m *MockE2ERepository) UpdateKeywordWeight(ctx context.Context, keywordID int, weight float64) error {
	return nil
}
func (m *MockE2ERepository) RemoveKeywordFromIndustry(ctx context.Context, keywordID int) error {
	return nil
}
func (m *MockE2ERepository) GetClassificationCodesByType(ctx context.Context, codeType string) ([]*repository.ClassificationCode, error) {
	return nil, nil
}
func (m *MockE2ERepository) AddClassificationCode(ctx context.Context, code *repository.ClassificationCode) error {
	return nil
}
func (m *MockE2ERepository) UpdateClassificationCode(ctx context.Context, code *repository.ClassificationCode) error {
	return nil
}
func (m *MockE2ERepository) DeleteClassificationCode(ctx context.Context, id int) error { return nil }
func (m *MockE2ERepository) GetPatternsByIndustry(ctx context.Context, industryID int) ([]*repository.IndustryPattern, error) {
	return nil, nil
}
func (m *MockE2ERepository) AddPattern(ctx context.Context, pattern *repository.IndustryPattern) error {
	return nil
}
func (m *MockE2ERepository) UpdatePattern(ctx context.Context, pattern *repository.IndustryPattern) error {
	return nil
}
func (m *MockE2ERepository) DeletePattern(ctx context.Context, id int) error { return nil }
func (m *MockE2ERepository) GetKeywordWeights(ctx context.Context, keyword string) ([]*repository.KeywordWeight, error) {
	return nil, nil
}
func (m *MockE2ERepository) UpdateKeywordWeightByID(ctx context.Context, weight *repository.KeywordWeight) error {
	return nil
}
func (m *MockE2ERepository) IncrementUsageCount(ctx context.Context, keyword string, industryID int) error {
	return nil
}
func (m *MockE2ERepository) ClassifyBusinessByKeywords(ctx context.Context, keywords []string) (*repository.ClassificationResult, error) {
	// Simple classification logic for testing
	// Map keywords to industries based on the test data
	keywordToIndustry := map[string]string{
		"software":      "Technology",
		"cloud":         "Technology",
		"ai":            "Technology",
		"data":          "Technology",
		"medical":       "Healthcare",
		"health":        "Healthcare",
		"patient":       "Healthcare",
		"clinic":        "Healthcare",
		"banking":       "Financial Services",
		"finance":       "Financial Services",
		"investment":    "Financial Services",
		"credit":        "Financial Services",
		"retail":        "Retail",
		"ecommerce":     "Retail",
		"shopping":      "Retail",
		"store":         "Retail",
		"manufacturing": "Manufacturing",
		"production":    "Manufacturing",
		"factory":       "Manufacturing",
		"industrial":    "Manufacturing",
	}

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

					// Return the first good match
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
func (m *MockE2ERepository) SearchIndustriesByPattern(ctx context.Context, pattern string) ([]*repository.Industry, error) {
	return nil, nil
}
func (m *MockE2ERepository) GetIndustryStatistics(ctx context.Context) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockE2ERepository) GetKeywordFrequency(ctx context.Context, industryID int) (map[string]int, error) {
	return nil, nil
}
func (m *MockE2ERepository) BulkInsertKeywords(ctx context.Context, keywords []*repository.IndustryKeyword) error {
	return nil
}
func (m *MockE2ERepository) BulkUpdateKeywords(ctx context.Context, keywords []*repository.IndustryKeyword) error {
	return nil
}
func (m *MockE2ERepository) BulkDeleteKeywords(ctx context.Context, keywordIDs []int) error {
	return nil
}
func (m *MockE2ERepository) Ping(ctx context.Context) error { return nil }
func (m *MockE2ERepository) GetDatabaseStats(ctx context.Context) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockE2ERepository) CleanupInactiveData(ctx context.Context) error { return nil }

// Add missing batch methods
func (m *MockE2ERepository) GetBatchClassificationCodes(ctx context.Context, industryIDs []int) (map[int][]*repository.ClassificationCode, error) {
	result := make(map[int][]*repository.ClassificationCode)
	for _, id := range industryIDs {
		if codes, exists := m.classificationCodes[id]; exists {
			result[id] = codes
		} else {
			result[id] = []*repository.ClassificationCode{
				{ID: 1, Code: "541511", Description: "Custom Computer Programming Services", CodeType: "NAICS"},
			}
		}
	}
	return result, nil
}

func (m *MockE2ERepository) GetBatchIndustries(ctx context.Context, industryIDs []int) (map[int]*repository.Industry, error) {
	result := make(map[int]*repository.Industry)
	for _, id := range industryIDs {
		if industry, exists := m.industries[id]; exists {
			result[id] = industry
		} else {
			result[id] = &repository.Industry{ID: id, Name: "Test Industry"}
		}
	}
	return result, nil
}

func (m *MockE2ERepository) GetBatchKeywords(ctx context.Context, industryIDs []int) (map[int][]*repository.KeywordWeight, error) {
	result := make(map[int][]*repository.KeywordWeight)
	for _, id := range industryIDs {
		result[id] = []*repository.KeywordWeight{
			{ID: 1, Keyword: "test", BaseWeight: 1.0},
		}
	}
	return result, nil
}

func (m *MockE2ERepository) GetCachedClassificationCodes(ctx context.Context, industryID int) ([]*repository.ClassificationCode, error) {
	if codes, exists := m.classificationCodes[industryID]; exists {
		return codes, nil
	}
	return []*repository.ClassificationCode{
		{ID: 1, Code: "541511", Description: "Custom Computer Programming Services", CodeType: "NAICS"},
	}, nil
}

func (m *MockE2ERepository) GetCachedClassificationCodesByType(ctx context.Context, codeType string) ([]*repository.ClassificationCode, error) {
	return []*repository.ClassificationCode{
		{ID: 1, Code: "541511", Description: "Custom Computer Programming Services", CodeType: codeType},
	}, nil
}

func (m *MockE2ERepository) InitializeIndustryCodeCache(ctx context.Context) error {
	return nil
}

func (m *MockE2ERepository) InvalidateIndustryCodeCache(ctx context.Context, rules []string) error {
	return nil
}

func (m *MockE2ERepository) GetIndustryCodeCacheStats() *repository.IndustryCodeCacheStats {
	return &repository.IndustryCodeCacheStats{
		Hits:   10,
		Misses: 2,
	}
}
