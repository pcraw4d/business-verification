package classification

import (
	"context"
	"testing"
	"time"

	"kyb-platform/internal/classification/repository"
)

// BenchmarkIndustryDetection benchmarks the industry detection performance
func BenchmarkIndustryDetection(b *testing.B) {
	mockRepo := createMockPerformanceRepository()
	service := NewIndustryDetectionService(mockRepo, nil)
	ctx := context.Background()

	testContent := "We develop innovative software solutions for businesses using cloud technology and artificial intelligence to transform digital experiences."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.DetectIndustryFromContent(ctx, testContent)
		if err != nil {
			b.Fatalf("industry detection failed: %v", err)
		}
	}
}

// BenchmarkClassificationCodeGeneration benchmarks the classification code generation performance
func BenchmarkClassificationCodeGeneration(b *testing.B) {
	mockRepo := createMockPerformanceRepository()
	classifier := NewClassificationCodeGenerator(mockRepo, nil)
	ctx := context.Background()

	keywords := []string{"software", "cloud", "technology", "business", "solutions"}
	industry := "Technology"
	confidence := 0.85

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := classifier.GenerateClassificationCodes(ctx, keywords, industry, confidence)
		if err != nil {
			b.Fatalf("classification code generation failed: %v", err)
		}
	}
}

// BenchmarkKeywordExtraction benchmarks the keyword extraction performance
func BenchmarkKeywordExtraction(b *testing.B) {
	service := NewIndustryDetectionService(nil, nil)

	testContent := "We develop innovative software solutions for businesses using cloud technology and artificial intelligence to transform digital experiences. Our platform leverages cutting-edge machine learning algorithms to provide intelligent business insights and automation capabilities."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.extractKeywordsFromContent(testContent)
	}
}

// TestConcurrentClassificationPerformance tests performance under concurrent load
func TestConcurrentClassificationPerformance(t *testing.T) {
	mockRepo := createMockPerformanceRepository()
	service := NewIndustryDetectionService(mockRepo, nil)
	classifier := NewClassificationCodeGenerator(mockRepo, nil)
	ctx := context.Background()

	testCases := []struct {
		name             string
		content          string
		expectedIndustry string
	}{
		{
			name:             "Technology Company",
			content:          "We develop innovative software solutions for businesses using cloud technology",
			expectedIndustry: "Technology",
		},
		{
			name:             "Healthcare Provider",
			content:          "Medical center providing comprehensive healthcare services and patient care",
			expectedIndustry: "Healthcare",
		},
		{
			name:             "Financial Institution",
			content:          "Banking services including loans, investments, and financial planning",
			expectedIndustry: "Financial Services",
		},
		{
			name:             "Retail Business",
			content:          "Retail store offering consumer products and customer service",
			expectedIndustry: "Retail",
		},
		{
			name:             "Manufacturing Company",
			content:          "Industrial manufacturing of machinery and equipment",
			expectedIndustry: "Manufacturing",
		},
	}

	// Test concurrent classification
	start := time.Now()
	results := make(chan *repository.ClassificationResult, len(testCases))
	errors := make(chan error, len(testCases))

	// Launch concurrent classification requests
	for _, tc := range testCases {
		go func(content, expectedIndustry string) {
			// Industry detection
			industryResult, err := service.DetectIndustryFromContent(ctx, content)
			if err != nil {
				errors <- err
				return
			}

			// Classification code generation
			classificationCodes, err := classifier.GenerateClassificationCodes(
				ctx,
				industryResult.KeywordsMatched,
				industryResult.Industry.Name,
				industryResult.Confidence,
			)
			if err != nil {
				errors <- err
				return
			}

			// Create a ClassificationResult for the channel
			classificationResult := &repository.ClassificationResult{
				Industry:   industryResult.Industry,
				Confidence: industryResult.Confidence,
				Keywords:   industryResult.KeywordsMatched,
				Patterns:   []string{},
				Codes:      []repository.ClassificationCode{},
				Reasoning:  "Performance test classification",
			}

			// Convert classification codes to the expected format
			if classificationCodes != nil {
				for _, code := range classificationCodes.MCC {
					classificationResult.Codes = append(classificationResult.Codes, repository.ClassificationCode{
						ID:          0, // Mock ID
						IndustryID:  industryResult.Industry.ID,
						CodeType:    "MCC",
						Code:        code.Code,
						Description: code.Description,
					})
				}
				for _, code := range classificationCodes.SIC {
					classificationResult.Codes = append(classificationResult.Codes, repository.ClassificationCode{
						ID:          0, // Mock ID
						IndustryID:  industryResult.Industry.ID,
						CodeType:    "SIC",
						Code:        code.Code,
						Description: code.Description,
					})
				}
				for _, code := range classificationCodes.NAICS {
					classificationResult.Codes = append(classificationResult.Codes, repository.ClassificationCode{
						ID:          0, // Mock ID
						IndustryID:  industryResult.Industry.ID,
						CodeType:    "NAICS",
						Code:        code.Code,
						Description: code.Description,
					})
				}
			}

			results <- classificationResult
		}(tc.content, tc.expectedIndustry)
	}

	// Collect results
	var successfulResults []*repository.ClassificationResult
	for i := 0; i < len(testCases); i++ {
		select {
		case result := <-results:
			successfulResults = append(successfulResults, result)
		case err := <-errors:
			t.Errorf("concurrent classification failed: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("concurrent classification timed out")
		}
	}

	duration := time.Since(start)
	t.Logf("✅ Concurrent classification completed in %v", duration)
	t.Logf("   Processed %d requests successfully", len(successfulResults))

	// Performance threshold: should complete within 2 seconds for 5 concurrent requests
	if duration > 2*time.Second {
		t.Errorf("concurrent classification took too long: %v (expected < 2s)", duration)
	}
}

// TestMemoryUsagePerformance tests memory usage during classification operations
func TestMemoryUsagePerformance(t *testing.T) {
	mockRepo := createMockPerformanceRepository()
	service := NewIndustryDetectionService(mockRepo, nil)
	classifier := NewClassificationCodeGenerator(mockRepo, nil)
	ctx := context.Background()

	// Large content to test memory usage
	largeContent := generateLargeContent(10000) // 10KB content

	start := time.Now()

	// Perform industry detection
	industryResult, err := service.DetectIndustryFromContent(ctx, largeContent)
	if err != nil {
		t.Fatalf("industry detection failed: %v", err)
	}

	// Generate classification codes
	classificationResult, err := classifier.GenerateClassificationCodes(
		ctx,
		industryResult.KeywordsMatched,
		industryResult.Industry.Name,
		industryResult.Confidence,
	)
	if err != nil {
		t.Fatalf("classification code generation failed: %v", err)
	}

	duration := time.Since(start)
	t.Logf("✅ Large content processing completed in %v", duration)
	t.Logf("   Content size: %d characters", len(largeContent))
	t.Logf("   Keywords extracted: %d", len(industryResult.KeywordsMatched))
	t.Logf("   Classification codes generated: %d",
		len(classificationResult.MCC)+len(classificationResult.SIC)+len(classificationResult.NAICS))

	// Performance threshold: should complete within 1 second for large content
	if duration > time.Second {
		t.Errorf("large content processing took too long: %v (expected < 1s)", duration)
	}
}

// TestDatabaseQueryPerformance tests database query performance with mock repository
func TestDatabaseQueryPerformance(t *testing.T) {
	mockRepo := createMockPerformanceRepository()
	ctx := context.Background()

	// Test industry listing performance
	start := time.Now()
	industries, err := mockRepo.ListIndustries(ctx, "")
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("failed to list industries: %v", err)
	}

	t.Logf("✅ Industry listing completed in %v", duration)
	t.Logf("   Retrieved %d industries", len(industries))

	// Performance threshold: should complete within 100ms
	if duration > 100*time.Millisecond {
		t.Errorf("industry listing took too long: %v (expected < 100ms)", duration)
	}

	// Test keyword search performance
	start = time.Now()
	keywords, err := mockRepo.SearchKeywords(ctx, "technology", 50)
	duration = time.Since(start)

	if err != nil {
		t.Fatalf("failed to search keywords: %v", err)
	}

	t.Logf("✅ Keyword search completed in %v", duration)
	t.Logf("   Found %d keywords", len(keywords))

	// Performance threshold: should complete within 50ms
	if duration > 50*time.Millisecond {
		t.Errorf("keyword search took too long: %v (expected < 50ms)", duration)
	}
}

// TestCachingPerformance tests the performance impact of caching (if implemented)
func TestCachingPerformance(t *testing.T) {
	mockRepo := createMockPerformanceRepository()
	service := NewIndustryDetectionService(mockRepo, nil)
	ctx := context.Background()

	testContent := "We develop innovative software solutions for businesses using cloud technology"

	// First request (cold start)
	start := time.Now()
	result1, err := service.DetectIndustryFromContent(ctx, testContent)
	coldStartDuration := time.Since(start)

	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}

	// Second request (should be faster if caching is implemented)
	start = time.Now()
	result2, err := service.DetectIndustryFromContent(ctx, testContent)
	cachedDuration := time.Since(start)

	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}

	// Verify results are consistent
	if result1.Industry.Name != result2.Industry.Name {
		t.Errorf("cached results inconsistent: %s vs %s",
			result1.Industry.Name, result2.Industry.Name)
	}

	t.Logf("✅ Cold start request: %v", coldStartDuration)
	t.Logf("✅ Cached request: %v", cachedDuration)

	// If caching is implemented, second request should be faster
	if cachedDuration >= coldStartDuration {
		t.Logf("ℹ️ No caching implemented or minimal performance improvement")
	} else {
		improvement := float64(coldStartDuration-cachedDuration) / float64(coldStartDuration) * 100
		t.Logf("✅ Caching performance improvement: %.1f%%", improvement)
	}
}

// Helper function to create a mock repository for performance testing
func createMockPerformanceRepository() *MockPerformanceRepository {
	return &MockPerformanceRepository{
		industries: map[int]*repository.Industry{
			1: {ID: 1, Name: "Technology", Category: "Technology"},
			2: {ID: 2, Name: "Healthcare", Category: "Healthcare"},
			3: {ID: 3, Name: "Financial Services", Category: "Financial"},
			4: {ID: 4, Name: "Retail", Category: "Retail"},
			5: {ID: 5, Name: "Manufacturing", Category: "Manufacturing"},
		},
		keywords: map[int][]*repository.IndustryKeyword{
			1: {
				{ID: 1, IndustryID: 1, Keyword: "software", Weight: 0.9},
				{ID: 2, IndustryID: 1, Keyword: "cloud", Weight: 0.85},
				{ID: 3, IndustryID: 1, Keyword: "technology", Weight: 0.8},
				{ID: 4, IndustryID: 1, Keyword: "digital", Weight: 0.75},
				{ID: 5, IndustryID: 1, Keyword: "innovation", Weight: 0.7},
			},
			2: {
				{ID: 6, IndustryID: 2, Keyword: "medical", Weight: 0.9},
				{ID: 7, IndustryID: 2, Keyword: "health", Weight: 0.85},
				{ID: 8, IndustryID: 2, Keyword: "patient", Weight: 0.8},
				{ID: 9, IndustryID: 2, Keyword: "care", Weight: 0.75},
				{ID: 10, IndustryID: 2, Keyword: "treatment", Weight: 0.7},
			},
			3: {
				{ID: 11, IndustryID: 3, Keyword: "banking", Weight: 0.9},
				{ID: 12, IndustryID: 3, Keyword: "finance", Weight: 0.85},
				{ID: 13, IndustryID: 3, Keyword: "investment", Weight: 0.8},
				{ID: 14, IndustryID: 3, Keyword: "loan", Weight: 0.75},
				{ID: 15, IndustryID: 3, Keyword: "credit", Weight: 0.7},
			},
		},
		classificationCodes: map[int][]*repository.ClassificationCode{
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
}

// Helper function to generate large content for testing
func generateLargeContent(size int) string {
	baseContent := "We develop innovative software solutions for businesses using cloud technology and artificial intelligence to transform digital experiences. Our platform leverages cutting-edge machine learning algorithms to provide intelligent business insights and automation capabilities. "

	content := ""
	for len(content) < size {
		content += baseContent
	}

	return content[:size]
}

// MockPerformanceRepository is a mock repository optimized for performance testing
type MockPerformanceRepository struct {
	industries          map[int]*repository.Industry
	keywords            map[int][]*repository.IndustryKeyword
	classificationCodes map[int][]*repository.ClassificationCode
}

// Implement only the methods needed for performance testing
func (m *MockPerformanceRepository) ListIndustries(ctx context.Context, category string) ([]*repository.Industry, error) {
	var industries []*repository.Industry
	for _, industry := range m.industries {
		if category == "" || industry.Category == category {
			industries = append(industries, industry)
		}
	}
	return industries, nil
}

func (m *MockPerformanceRepository) SearchKeywords(ctx context.Context, query string, limit int) ([]*repository.IndustryKeyword, error) {
	var results []*repository.IndustryKeyword
	count := 0

	for _, keywords := range m.keywords {
		for _, keyword := range keywords {
			if count >= limit {
				break
			}
			// Simple keyword matching for performance testing
			if len(results) < limit {
				results = append(results, keyword)
				count++
			}
		}
	}

	return results, nil
}

// Stub implementations for interface compliance
func (m *MockPerformanceRepository) GetIndustryByID(ctx context.Context, id int) (*repository.Industry, error) {
	return nil, nil
}
func (m *MockPerformanceRepository) GetIndustryByName(ctx context.Context, name string) (*repository.Industry, error) {
	return nil, nil
}
func (m *MockPerformanceRepository) CreateIndustry(ctx context.Context, industry *repository.Industry) error {
	return nil
}
func (m *MockPerformanceRepository) UpdateIndustry(ctx context.Context, industry *repository.Industry) error {
	return nil
}
func (m *MockPerformanceRepository) DeleteIndustry(ctx context.Context, id int) error { return nil }
func (m *MockPerformanceRepository) GetKeywordsByIndustry(ctx context.Context, industryID int) ([]*repository.IndustryKeyword, error) {
	return nil, nil
}
func (m *MockPerformanceRepository) AddKeywordToIndustry(ctx context.Context, industryID int, keyword string, weight float64) error {
	return nil
}
func (m *MockPerformanceRepository) UpdateKeywordWeight(ctx context.Context, keywordID int, weight float64) error {
	return nil
}
func (m *MockPerformanceRepository) RemoveKeywordFromIndustry(ctx context.Context, keywordID int) error {
	return nil
}
func (m *MockPerformanceRepository) GetClassificationCodesByIndustry(ctx context.Context, industryID int) ([]*repository.ClassificationCode, error) {
	return nil, nil
}
func (m *MockPerformanceRepository) GetClassificationCodesByType(ctx context.Context, codeType string) ([]*repository.ClassificationCode, error) {
	return nil, nil
}
func (m *MockPerformanceRepository) AddClassificationCode(ctx context.Context, code *repository.ClassificationCode) error {
	return nil
}
func (m *MockPerformanceRepository) UpdateClassificationCode(ctx context.Context, code *repository.ClassificationCode) error {
	return nil
}
func (m *MockPerformanceRepository) DeleteClassificationCode(ctx context.Context, id int) error {
	return nil
}
func (m *MockPerformanceRepository) GetPatternsByIndustry(ctx context.Context, industryID int) ([]*repository.IndustryPattern, error) {
	return nil, nil
}
func (m *MockPerformanceRepository) AddPattern(ctx context.Context, pattern *repository.IndustryPattern) error {
	return nil
}
func (m *MockPerformanceRepository) UpdatePattern(ctx context.Context, pattern *repository.IndustryPattern) error {
	return nil
}
func (m *MockPerformanceRepository) DeletePattern(ctx context.Context, id int) error { return nil }
func (m *MockPerformanceRepository) GetKeywordWeights(ctx context.Context, keyword string) ([]*repository.KeywordWeight, error) {
	return nil, nil
}
func (m *MockPerformanceRepository) UpdateKeywordWeightByID(ctx context.Context, weight *repository.KeywordWeight) error {
	return nil
}
func (m *MockPerformanceRepository) IncrementUsageCount(ctx context.Context, keyword string, industryID int) error {
	return nil
}
func (m *MockPerformanceRepository) ClassifyBusiness(ctx context.Context, businessName, websiteURL string) (*repository.ClassificationResult, error) {
	return nil, nil
}
func (m *MockPerformanceRepository) ClassifyBusinessByKeywords(ctx context.Context, keywords []string) (*repository.ClassificationResult, error) {
	if len(keywords) == 0 {
		return &repository.ClassificationResult{
			Industry:   &repository.Industry{ID: 0, Name: "General Business", Category: "General"},
			Confidence: 0.5,
			Keywords:   []string{},
			Patterns:   []string{},
			Codes:      []repository.ClassificationCode{},
			Reasoning:  "No keywords provided",
		}, nil
	}

	// Find best matching industry based on keyword weights
	var bestIndustry *repository.Industry
	var bestScore float64

	for industryID, industryKeywords := range m.keywords {
		score := 0.0
		matchedKeywords := []string{}

		for _, keyword := range industryKeywords {
			for _, inputKeyword := range keywords {
				if keyword.Keyword == inputKeyword {
					score += keyword.Weight
					matchedKeywords = append(matchedKeywords, keyword.Keyword)
				}
			}
		}

		if score > bestScore {
			bestScore = score
			bestIndustry = m.industries[industryID]
		}
	}

	if bestIndustry == nil {
		bestIndustry = &repository.Industry{ID: 0, Name: "General Business", Category: "General"}
		bestScore = 0.5
	}

	// Get classification codes for the best industry
	var codes []repository.ClassificationCode
	if bestIndustry.ID > 0 {
		if industryCodes, exists := m.classificationCodes[bestIndustry.ID]; exists {
			for _, code := range industryCodes {
				codes = append(codes, repository.ClassificationCode{
					ID:          code.ID,
					IndustryID:  code.IndustryID,
					CodeType:    code.CodeType,
					Code:        code.Code,
					Description: code.Description,
				})
			}
		}
	}

	return &repository.ClassificationResult{
		Industry:   bestIndustry,
		Confidence: bestScore,
		Keywords:   keywords,
		Patterns:   []string{},
		Codes:      codes,
		Reasoning:  "Classification based on keyword matching",
	}, nil
}
func (m *MockPerformanceRepository) GetTopIndustriesByKeywords(ctx context.Context, keywords []string, limit int) ([]*repository.Industry, error) {
	return nil, nil
}
func (m *MockPerformanceRepository) SearchIndustriesByPattern(ctx context.Context, pattern string) ([]*repository.Industry, error) {
	return nil, nil
}
func (m *MockPerformanceRepository) GetIndustryStatistics(ctx context.Context) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockPerformanceRepository) GetKeywordFrequency(ctx context.Context, industryID int) (map[string]int, error) {
	return nil, nil
}
func (m *MockPerformanceRepository) BulkInsertKeywords(ctx context.Context, keywords []*repository.IndustryKeyword) error {
	return nil
}
func (m *MockPerformanceRepository) BulkUpdateKeywords(ctx context.Context, keywords []*repository.IndustryKeyword) error {
	return nil
}
func (m *MockPerformanceRepository) BulkDeleteKeywords(ctx context.Context, keywordIDs []int) error {
	return nil
}
func (m *MockPerformanceRepository) Ping(ctx context.Context) error { return nil }
func (m *MockPerformanceRepository) GetDatabaseStats(ctx context.Context) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockPerformanceRepository) CleanupInactiveData(ctx context.Context) error { return nil }

// Add missing batch methods
func (m *MockPerformanceRepository) GetBatchClassificationCodes(ctx context.Context, industryIDs []int) (map[int][]*repository.ClassificationCode, error) {
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

func (m *MockPerformanceRepository) GetBatchIndustries(ctx context.Context, industryIDs []int) (map[int]*repository.Industry, error) {
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

func (m *MockPerformanceRepository) GetBatchKeywords(ctx context.Context, industryIDs []int) (map[int][]*repository.KeywordWeight, error) {
	result := make(map[int][]*repository.KeywordWeight)
	for _, id := range industryIDs {
		result[id] = []*repository.KeywordWeight{
			{ID: 1, Keyword: "test", BaseWeight: 1.0},
		}
	}
	return result, nil
}

func (m *MockPerformanceRepository) GetCachedClassificationCodes(ctx context.Context, industryID int) ([]*repository.ClassificationCode, error) {
	if codes, exists := m.classificationCodes[industryID]; exists {
		return codes, nil
	}
	return []*repository.ClassificationCode{
		{ID: 1, Code: "541511", Description: "Custom Computer Programming Services", CodeType: "NAICS"},
	}, nil
}

func (m *MockPerformanceRepository) GetCachedClassificationCodesByType(ctx context.Context, codeType string) ([]*repository.ClassificationCode, error) {
	return []*repository.ClassificationCode{
		{ID: 1, Code: "541511", Description: "Custom Computer Programming Services", CodeType: codeType},
	}, nil
}

func (m *MockPerformanceRepository) InitializeIndustryCodeCache(ctx context.Context) error {
	return nil
}

func (m *MockPerformanceRepository) InvalidateIndustryCodeCache(ctx context.Context, rules []string) error {
	return nil
}

func (m *MockPerformanceRepository) GetIndustryCodeCacheStats() *repository.IndustryCodeCacheStats {
	return &repository.IndustryCodeCacheStats{
		Hits:   10,
		Misses: 2,
	}
}
