package testutil

import (
	"context"
	"time"

	"kyb-platform/internal/classification/repository"
)

// MockKeywordRepository is a comprehensive mock implementation of repository.KeywordRepository
// for use across all classification tests. This consolidates all mock implementations.
type MockKeywordRepository struct {
	// Industry data
	industries map[int]*repository.Industry
	industriesByName map[string]*repository.Industry

	// Keyword data
	keywordsByIndustry map[int][]*repository.IndustryKeyword
	keywordWeights map[string][]*repository.KeywordWeight

	// Classification code data
	codesByIndustry map[int][]*repository.ClassificationCode
	codesByType map[string][]*repository.ClassificationCode
	keywordCodes map[string][]*repository.ClassificationCodeWithMetadata

	// Error injection
	errorMap map[string]error
}

// NewMockKeywordRepository creates a new mock repository with default test data
func NewMockKeywordRepository() *MockKeywordRepository {
	mock := &MockKeywordRepository{
		industries: make(map[int]*repository.Industry),
		industriesByName: make(map[string]*repository.Industry),
		keywordsByIndustry: make(map[int][]*repository.IndustryKeyword),
		keywordWeights: make(map[string][]*repository.KeywordWeight),
		codesByIndustry: make(map[int][]*repository.ClassificationCode),
		codesByType: make(map[string][]*repository.ClassificationCode),
		keywordCodes: make(map[string][]*repository.ClassificationCodeWithMetadata),
		errorMap: make(map[string]error),
	}

	mock.initializeDefaultData()
	return mock
}

// initializeDefaultData sets up default test data
func (m *MockKeywordRepository) initializeDefaultData() {
	// Technology industry (ID: 1)
	techIndustry := &repository.Industry{ID: 1, Name: "Technology"}
	m.industries[1] = techIndustry
	m.industriesByName["Technology"] = techIndustry

	m.keywordsByIndustry[1] = []*repository.IndustryKeyword{
		{ID: 1, IndustryID: 1, Keyword: "software", Weight: 1.0, IsActive: true},
		{ID: 2, IndustryID: 1, Keyword: "technology", Weight: 1.0, IsActive: true},
		{ID: 3, IndustryID: 1, Keyword: "platform", Weight: 0.9, IsActive: true},
	}

	m.codesByIndustry[1] = []*repository.ClassificationCode{
		{ID: 1, IndustryID: 1, Code: "5734", CodeType: "MCC", Description: "Computer Software Stores", IsActive: true},
		{ID: 2, IndustryID: 1, Code: "7372", CodeType: "SIC", Description: "Prepackaged Software", IsActive: true},
		{ID: 3, IndustryID: 1, Code: "541511", CodeType: "NAICS", Description: "Custom Computer Programming Services", IsActive: true},
	}

	// Financial Services industry (ID: 2)
	financeIndustry := &repository.Industry{ID: 2, Name: "Financial Services"}
	m.industries[2] = financeIndustry
	m.industriesByName["Financial Services"] = financeIndustry

	m.keywordsByIndustry[2] = []*repository.IndustryKeyword{
		{ID: 4, IndustryID: 2, Keyword: "bank", Weight: 1.0, IsActive: true},
		{ID: 5, IndustryID: 2, Keyword: "finance", Weight: 1.0, IsActive: true},
		{ID: 6, IndustryID: 2, Keyword: "credit", Weight: 0.9, IsActive: true},
	}

	m.codesByIndustry[2] = []*repository.ClassificationCode{
		{ID: 4, IndustryID: 2, Code: "6011", CodeType: "MCC", Description: "Automated Teller Machine Services", IsActive: true},
		{ID: 5, IndustryID: 2, Code: "6021", CodeType: "SIC", Description: "National Commercial Banks", IsActive: true},
		{ID: 6, IndustryID: 2, Code: "522110", CodeType: "NAICS", Description: "Commercial Banking", IsActive: true},
	}

	// Initialize keyword codes for hybrid generation
	m.initializeKeywordCodes()
}

// initializeKeywordCodes sets up keyword-to-code mappings for hybrid generation
func (m *MockKeywordRepository) initializeKeywordCodes() {
	// Technology keywords
	m.keywordCodes["software"] = []*repository.ClassificationCodeWithMetadata{
		{
			ClassificationCode: repository.ClassificationCode{
				ID: 1, Code: "5734", CodeType: "MCC",
				Description: "Computer Software Stores", IsActive: true,
			},
			RelevanceScore: 0.9,
			MatchType: "exact",
		},
		{
			ClassificationCode: repository.ClassificationCode{
				ID: 3, Code: "541511", CodeType: "NAICS",
				Description: "Custom Computer Programming Services", IsActive: true,
			},
			RelevanceScore: 0.85,
			MatchType: "exact",
		},
	}

	m.keywordCodes["technology"] = []*repository.ClassificationCodeWithMetadata{
		{
			ClassificationCode: repository.ClassificationCode{
				ID: 2, Code: "7372", CodeType: "SIC",
				Description: "Prepackaged Software", IsActive: true,
			},
			RelevanceScore: 0.8,
			MatchType: "partial",
		},
	}

	// Financial keywords
	m.keywordCodes["bank"] = []*repository.ClassificationCodeWithMetadata{
		{
			ClassificationCode: repository.ClassificationCode{
				ID: 4, Code: "6011", CodeType: "MCC",
				Description: "Automated Teller Machine Services", IsActive: true,
			},
			RelevanceScore: 0.9,
			MatchType: "exact",
		},
	}

	m.keywordCodes["finance"] = []*repository.ClassificationCodeWithMetadata{
		{
			ClassificationCode: repository.ClassificationCode{
				ID: 6, Code: "522110", CodeType: "NAICS",
				Description: "Commercial Banking", IsActive: true,
			},
			RelevanceScore: 0.85,
			MatchType: "exact",
		},
	}
}

// SetError allows injecting errors for testing error paths
func (m *MockKeywordRepository) SetError(method string, err error) {
	m.errorMap[method] = err
}

// ClearError clears an injected error
func (m *MockKeywordRepository) ClearError(method string) {
	delete(m.errorMap, method)
}

// Implement repository.KeywordRepository interface

func (m *MockKeywordRepository) GetIndustryByID(ctx context.Context, id int) (*repository.Industry, error) {
	if err := m.errorMap["GetIndustryByID"]; err != nil {
		return nil, err
	}
	return m.industries[id], nil
}

func (m *MockKeywordRepository) GetIndustryByName(ctx context.Context, name string) (*repository.Industry, error) {
	if err := m.errorMap["GetIndustryByName"]; err != nil {
		return nil, err
	}
	industry, exists := m.industriesByName[name]
	if !exists {
		// Return a default industry if not found
		return &repository.Industry{ID: 999, Name: name}, nil
	}
	return industry, nil
}

func (m *MockKeywordRepository) ListIndustries(ctx context.Context, category string) ([]*repository.Industry, error) {
	if err := m.errorMap["ListIndustries"]; err != nil {
		return nil, err
	}
	industries := make([]*repository.Industry, 0, len(m.industries))
	for _, industry := range m.industries {
		industries = append(industries, industry)
	}
	return industries, nil
}

func (m *MockKeywordRepository) CreateIndustry(ctx context.Context, industry *repository.Industry) error {
	if err := m.errorMap["CreateIndustry"]; err != nil {
		return err
	}
	m.industries[industry.ID] = industry
	m.industriesByName[industry.Name] = industry
	return nil
}

func (m *MockKeywordRepository) UpdateIndustry(ctx context.Context, industry *repository.Industry) error {
	if err := m.errorMap["UpdateIndustry"]; err != nil {
		return err
	}
	m.industries[industry.ID] = industry
	m.industriesByName[industry.Name] = industry
	return nil
}

func (m *MockKeywordRepository) DeleteIndustry(ctx context.Context, id int) error {
	if err := m.errorMap["DeleteIndustry"]; err != nil {
		return err
	}
	if industry, exists := m.industries[id]; exists {
		delete(m.industriesByName, industry.Name)
	}
	delete(m.industries, id)
	return nil
}

func (m *MockKeywordRepository) GetKeywordsByIndustry(ctx context.Context, industryID int) ([]*repository.IndustryKeyword, error) {
	if err := m.errorMap["GetKeywordsByIndustry"]; err != nil {
		return nil, err
	}
	return m.keywordsByIndustry[industryID], nil
}

func (m *MockKeywordRepository) SearchKeywords(ctx context.Context, query string, limit int) ([]*repository.IndustryKeyword, error) {
	if err := m.errorMap["SearchKeywords"]; err != nil {
		return nil, err
	}
	results := make([]*repository.IndustryKeyword, 0)
	for _, keywords := range m.keywordsByIndustry {
		for _, keyword := range keywords {
			if len(results) >= limit {
				break
			}
			results = append(results, keyword)
		}
	}
	return results, nil
}

func (m *MockKeywordRepository) AddKeywordToIndustry(ctx context.Context, industryID int, keyword string, weight float64) error {
	if err := m.errorMap["AddKeywordToIndustry"]; err != nil {
		return err
	}
	kw := &repository.IndustryKeyword{
		ID: len(m.keywordsByIndustry[industryID]) + 1,
		IndustryID: industryID,
		Keyword: keyword,
		Weight: weight,
		IsActive: true,
	}
	m.keywordsByIndustry[industryID] = append(m.keywordsByIndustry[industryID], kw)
	return nil
}

func (m *MockKeywordRepository) UpdateKeywordWeight(ctx context.Context, keywordID int, weight float64) error {
	if err := m.errorMap["UpdateKeywordWeight"]; err != nil {
		return err
	}
	for _, keywords := range m.keywordsByIndustry {
		for _, kw := range keywords {
			if kw.ID == keywordID {
				kw.Weight = weight
				return nil
			}
		}
	}
	return nil
}

func (m *MockKeywordRepository) RemoveKeywordFromIndustry(ctx context.Context, keywordID int) error {
	if err := m.errorMap["RemoveKeywordFromIndustry"]; err != nil {
		return err
	}
	for industryID, keywords := range m.keywordsByIndustry {
		for i, kw := range keywords {
			if kw.ID == keywordID {
				m.keywordsByIndustry[industryID] = append(keywords[:i], keywords[i+1:]...)
				return nil
			}
		}
	}
	return nil
}

func (m *MockKeywordRepository) GetClassificationCodesByIndustry(ctx context.Context, industryID int) ([]*repository.ClassificationCode, error) {
	if err := m.errorMap["GetClassificationCodesByIndustry"]; err != nil {
		return nil, err
	}
	return m.codesByIndustry[industryID], nil
}

func (m *MockKeywordRepository) GetClassificationCodesByType(ctx context.Context, codeType string) ([]*repository.ClassificationCode, error) {
	if err := m.errorMap["GetClassificationCodesByType"]; err != nil {
		return nil, err
	}
	return m.codesByType[codeType], nil
}

func (m *MockKeywordRepository) AddClassificationCode(ctx context.Context, code *repository.ClassificationCode) error {
	if err := m.errorMap["AddClassificationCode"]; err != nil {
		return err
	}
	m.codesByIndustry[code.IndustryID] = append(m.codesByIndustry[code.IndustryID], code)
	return nil
}

func (m *MockKeywordRepository) UpdateClassificationCode(ctx context.Context, code *repository.ClassificationCode) error {
	if err := m.errorMap["UpdateClassificationCode"]; err != nil {
		return err
	}
	return nil
}

func (m *MockKeywordRepository) DeleteClassificationCode(ctx context.Context, id int) error {
	if err := m.errorMap["DeleteClassificationCode"]; err != nil {
		return err
	}
	return nil
}

func (m *MockKeywordRepository) GetCachedClassificationCodes(ctx context.Context, industryID int) ([]*repository.ClassificationCode, error) {
	if err := m.errorMap["GetCachedClassificationCodes"]; err != nil {
		return nil, err
	}
	return m.GetClassificationCodesByIndustry(ctx, industryID)
}

func (m *MockKeywordRepository) GetCachedClassificationCodesByType(ctx context.Context, codeType string) ([]*repository.ClassificationCode, error) {
	if err := m.errorMap["GetCachedClassificationCodesByType"]; err != nil {
		return nil, err
	}
	return m.GetClassificationCodesByType(ctx, codeType)
}

func (m *MockKeywordRepository) InitializeIndustryCodeCache(ctx context.Context) error {
	if err := m.errorMap["InitializeIndustryCodeCache"]; err != nil {
		return err
	}
	return nil
}

func (m *MockKeywordRepository) InvalidateIndustryCodeCache(ctx context.Context, patterns []string) error {
	if err := m.errorMap["InvalidateIndustryCodeCache"]; err != nil {
		return err
	}
	return nil
}

func (m *MockKeywordRepository) GetIndustryCodeCacheStats() *repository.IndustryCodeCacheStats {
	return &repository.IndustryCodeCacheStats{
		Hits: 10,
		Misses: 2,
	}
}

func (m *MockKeywordRepository) GetBatchClassificationCodes(ctx context.Context, industryIDs []int) (map[int][]*repository.ClassificationCode, error) {
	if err := m.errorMap["GetBatchClassificationCodes"]; err != nil {
		return nil, err
	}
	result := make(map[int][]*repository.ClassificationCode)
	for _, id := range industryIDs {
		result[id] = m.codesByIndustry[id]
	}
	return result, nil
}

func (m *MockKeywordRepository) GetBatchIndustries(ctx context.Context, industryIDs []int) (map[int]*repository.Industry, error) {
	if err := m.errorMap["GetBatchIndustries"]; err != nil {
		return nil, err
	}
	result := make(map[int]*repository.Industry)
	for _, id := range industryIDs {
		result[id] = m.industries[id]
	}
	return result, nil
}

func (m *MockKeywordRepository) GetBatchKeywords(ctx context.Context, industryIDs []int) (map[int][]*repository.KeywordWeight, error) {
	if err := m.errorMap["GetBatchKeywords"]; err != nil {
		return nil, err
	}
	result := make(map[int][]*repository.KeywordWeight)
	for _, id := range industryIDs {
		keywords := m.keywordsByIndustry[id]
		weights := make([]*repository.KeywordWeight, len(keywords))
		for i, kw := range keywords {
			weights[i] = &repository.KeywordWeight{
				ID: kw.ID,
				Keyword: kw.Keyword,
				BaseWeight: kw.Weight,
			}
		}
		result[id] = weights
	}
	return result, nil
}

func (m *MockKeywordRepository) GetPatternsByIndustry(ctx context.Context, industryID int) ([]*repository.IndustryPattern, error) {
	if err := m.errorMap["GetPatternsByIndustry"]; err != nil {
		return nil, err
	}
	return []*repository.IndustryPattern{}, nil
}

func (m *MockKeywordRepository) AddPattern(ctx context.Context, pattern *repository.IndustryPattern) error {
	if err := m.errorMap["AddPattern"]; err != nil {
		return err
	}
	return nil
}

func (m *MockKeywordRepository) UpdatePattern(ctx context.Context, pattern *repository.IndustryPattern) error {
	if err := m.errorMap["UpdatePattern"]; err != nil {
		return err
	}
	return nil
}

func (m *MockKeywordRepository) DeletePattern(ctx context.Context, id int) error {
	if err := m.errorMap["DeletePattern"]; err != nil {
		return err
	}
	return nil
}

func (m *MockKeywordRepository) GetKeywordWeights(ctx context.Context, keyword string) ([]*repository.KeywordWeight, error) {
	if err := m.errorMap["GetKeywordWeights"]; err != nil {
		return nil, err
	}
	return m.keywordWeights[keyword], nil
}

func (m *MockKeywordRepository) UpdateKeywordWeightByID(ctx context.Context, weight *repository.KeywordWeight) error {
	if err := m.errorMap["UpdateKeywordWeightByID"]; err != nil {
		return err
	}
	return nil
}

func (m *MockKeywordRepository) IncrementUsageCount(ctx context.Context, keyword string, industryID int) error {
	if err := m.errorMap["IncrementUsageCount"]; err != nil {
		return err
	}
	return nil
}

func (m *MockKeywordRepository) ClassifyBusiness(ctx context.Context, businessName, websiteURL string) (*repository.ClassificationResult, error) {
	if err := m.errorMap["ClassifyBusiness"]; err != nil {
		return nil, err
	}
	return nil, nil
}

func (m *MockKeywordRepository) ClassifyBusinessByKeywords(ctx context.Context, keywords []string) (*repository.ClassificationResult, error) {
	if err := m.errorMap["ClassifyBusinessByKeywords"]; err != nil {
		return nil, err
	}
	return nil, nil
}

func (m *MockKeywordRepository) GetTopIndustriesByKeywords(ctx context.Context, keywords []string, limit int) ([]*repository.Industry, error) {
	if err := m.errorMap["GetTopIndustriesByKeywords"]; err != nil {
		return nil, err
	}
	return []*repository.Industry{m.industries[1]}, nil
}

func (m *MockKeywordRepository) SearchIndustriesByPattern(ctx context.Context, pattern string) ([]*repository.Industry, error) {
	if err := m.errorMap["SearchIndustriesByPattern"]; err != nil {
		return nil, err
	}
	return []*repository.Industry{}, nil
}

func (m *MockKeywordRepository) GetIndustryStatistics(ctx context.Context) (map[string]interface{}, error) {
	if err := m.errorMap["GetIndustryStatistics"]; err != nil {
		return nil, err
	}
	return make(map[string]interface{}), nil
}

func (m *MockKeywordRepository) GetKeywordFrequency(ctx context.Context, industryID int) (map[string]int, error) {
	if err := m.errorMap["GetKeywordFrequency"]; err != nil {
		return nil, err
	}
	return make(map[string]int), nil
}

func (m *MockKeywordRepository) BulkInsertKeywords(ctx context.Context, keywords []*repository.IndustryKeyword) error {
	if err := m.errorMap["BulkInsertKeywords"]; err != nil {
		return err
	}
	return nil
}

func (m *MockKeywordRepository) BulkUpdateKeywords(ctx context.Context, keywords []*repository.IndustryKeyword) error {
	if err := m.errorMap["BulkUpdateKeywords"]; err != nil {
		return err
	}
	return nil
}

func (m *MockKeywordRepository) BulkDeleteKeywords(ctx context.Context, keywordIDs []int) error {
	if err := m.errorMap["BulkDeleteKeywords"]; err != nil {
		return err
	}
	return nil
}

func (m *MockKeywordRepository) Ping(ctx context.Context) error {
	if err := m.errorMap["Ping"]; err != nil {
		return err
	}
	return nil
}

func (m *MockKeywordRepository) GetDatabaseStats(ctx context.Context) (map[string]interface{}, error) {
	if err := m.errorMap["GetDatabaseStats"]; err != nil {
		return nil, err
	}
	return make(map[string]interface{}), nil
}

func (m *MockKeywordRepository) CleanupInactiveData(ctx context.Context) error {
	if err := m.errorMap["CleanupInactiveData"]; err != nil {
		return err
	}
	return nil
}

// GetClassificationCodesByKeywords implements the new hybrid generation method
func (m *MockKeywordRepository) GetClassificationCodesByKeywords(
	ctx context.Context,
	keywords []string,
	codeType string,
	minRelevance float64,
) ([]*repository.ClassificationCodeWithMetadata, error) {
	if err := m.errorMap["GetClassificationCodesByKeywords"]; err != nil {
		return nil, err
	}

	result := make([]*repository.ClassificationCodeWithMetadata, 0)
	seen := make(map[int]bool)

	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}

		// Look up keyword codes
		codes, exists := m.keywordCodes[keyword]
		if !exists {
			continue
		}

		for _, code := range codes {
			// Filter by code type
			if code.ClassificationCode.CodeType != codeType {
				continue
			}

			// Filter by relevance threshold
			if code.RelevanceScore < minRelevance {
				continue
			}

			// Deduplicate
			if seen[code.ClassificationCode.ID] {
				continue
			}
			seen[code.ClassificationCode.ID] = true

			result = append(result, code)
		}
	}

	return result, nil
}

// GetCalibrationStatistics retrieves calibration statistics for a date range
func (m *MockKeywordRepository) GetCalibrationStatistics(ctx context.Context, startDate, endDate time.Time) ([]*repository.CalibrationBinStatistics, error) {
	if err := m.errorMap["GetCalibrationStatistics"]; err != nil {
		return nil, err
	}
	// Return empty slice for mock - tests can override if needed
	return []*repository.CalibrationBinStatistics{}, nil
}

// SaveClassificationAccuracy saves classification accuracy tracking data
func (m *MockKeywordRepository) SaveClassificationAccuracy(ctx context.Context, tracking *repository.ClassificationAccuracyTracking) error {
	if err := m.errorMap["SaveClassificationAccuracy"]; err != nil {
		return err
	}
	// Mock implementation - just return nil
	return nil
}

// UpdateClassificationAccuracy updates classification accuracy tracking data
func (m *MockKeywordRepository) UpdateClassificationAccuracy(ctx context.Context, requestID string, actualIndustry string, validatedBy string) error {
	if err := m.errorMap["UpdateClassificationAccuracy"]; err != nil {
		return err
	}
	// Mock implementation - just return nil
	return nil
}

// ClassifyBusinessByKeywordsTrigram performs classification using trigram similarity
func (m *MockKeywordRepository) ClassifyBusinessByKeywordsTrigram(ctx context.Context, keywords []string, businessName string) (*repository.ClassificationResult, error) {
	if err := m.errorMap["ClassifyBusinessByKeywordsTrigram"]; err != nil {
		return nil, err
	}
	// Mock implementation - delegate to regular classification
	return m.ClassifyBusinessByKeywords(ctx, keywords)
}

// GetIndustryTopicsByKeywords retrieves industry-topic mappings for given keywords
func (m *MockKeywordRepository) GetIndustryTopicsByKeywords(ctx context.Context, keywords []string) (map[int]float64, error) {
	if err := m.errorMap["GetIndustryTopicsByKeywords"]; err != nil {
		return nil, err
	}
	// Mock implementation - return empty map (industry_id -> relevance_score)
	return make(map[int]float64), nil
}

// GetTopicAccuracy retrieves accuracy score for a topic-industry pair
func (m *MockKeywordRepository) GetTopicAccuracy(ctx context.Context, industryID int, topic string) (float64, error) {
	if err := m.errorMap["GetTopicAccuracy"]; err != nil {
		return 0, err
	}
	// Mock implementation - return default accuracy
	return 0.75, nil
}

// FindIndustriesByPatterns finds industries matching keyword patterns
func (m *MockKeywordRepository) FindIndustriesByPatterns(ctx context.Context, patterns []string) ([]*repository.PatternMatchResult, error) {
	if err := m.errorMap["FindIndustriesByPatterns"]; err != nil {
		return nil, err
	}
	// Mock implementation - return empty slice
	return []*repository.PatternMatchResult{}, nil
}

// GetPatternMatches retrieves specific keyword patterns
func (m *MockKeywordRepository) GetPatternMatches(ctx context.Context, industryID int, patterns []string) ([]*repository.KeywordPattern, error) {
	if err := m.errorMap["GetPatternMatches"]; err != nil {
		return nil, err
	}
	// Mock implementation - return empty slice
	return []*repository.KeywordPattern{}, nil
}

// BatchFindKeywords performs batch keyword lookup (Phase 2.2)
func (m *MockKeywordRepository) BatchFindKeywords(ctx context.Context, keywords []string) (map[string][]repository.IndustryMatch, error) {
	if err := m.errorMap["BatchFindKeywords"]; err != nil {
		return nil, err
	}
	// Mock implementation - return empty map
	result := make(map[string][]repository.IndustryMatch)
	for _, keyword := range keywords {
		result[keyword] = []repository.IndustryMatch{}
	}
	return result, nil
}

// BatchFindIndustryTopics performs batch topic lookup (Phase 2.2)
func (m *MockKeywordRepository) BatchFindIndustryTopics(ctx context.Context, keywords []string) (map[string][]repository.TopicMatch, error) {
	if err := m.errorMap["BatchFindIndustryTopics"]; err != nil {
		return nil, err
	}
	// Mock implementation - return empty map
	result := make(map[string][]repository.TopicMatch)
	for _, keyword := range keywords {
		result[keyword] = []repository.TopicMatch{}
	}
	return result, nil
}

// FindCodesByFullTextSearch performs full-text search for classification codes (Phase 4.2)
func (m *MockKeywordRepository) FindCodesByFullTextSearch(ctx context.Context, searchText string, codeType string) ([]*repository.ClassificationCode, error) {
	if err := m.errorMap["FindCodesByFullTextSearch"]; err != nil {
		return nil, err
	}
	// Mock implementation - return codes matching the type
	if codes, ok := m.codesByType[codeType]; ok {
		return codes, nil
	}
	return []*repository.ClassificationCode{}, nil
}

// Phase 2: New methods for enhanced code retrieval

// GetCodesByKeywords returns codes matching keywords with their weights (Phase 2)
func (m *MockKeywordRepository) GetCodesByKeywords(ctx context.Context, codeType string, keywords []string) []struct {
	Code        string
	Description string
	Weight      float64
} {
	result := []struct {
		Code        string
		Description string
		Weight      float64
	}{}
	
	for _, keyword := range keywords {
		if codes, exists := m.keywordCodes[keyword]; exists {
			for _, code := range codes {
				if code.ClassificationCode.CodeType == codeType {
					result = append(result, struct {
						Code        string
						Description string
						Weight      float64
					}{
						Code:        code.ClassificationCode.Code,
						Description: code.ClassificationCode.Description,
						Weight:      code.RelevanceScore,
					})
				}
			}
		}
	}
	
	return result
}

// GetCodesByTrigramSimilarity returns codes with similarity scores using trigram matching (Phase 2)
func (m *MockKeywordRepository) GetCodesByTrigramSimilarity(
	ctx context.Context,
	codeType string,
	industryName string,
	threshold float64,
	limit int,
) []struct {
	Code        string
	Description string
	Similarity  float64
} {
	result := []struct {
		Code        string
		Description string
		Similarity  float64
	}{}
	
	// Mock implementation - return codes from industries matching the name
	for _, industry := range m.industries {
		if codes, exists := m.codesByIndustry[industry.ID]; exists {
			for _, code := range codes {
				if code.CodeType == codeType && len(result) < limit {
					// Mock similarity score
					similarity := 0.75
					if similarity >= threshold {
						result = append(result, struct {
							Code        string
							Description string
							Similarity  float64
						}{
							Code:        code.Code,
							Description: code.Description,
							Similarity:  similarity,
						})
					}
				}
			}
		}
	}
	
	return result
}

// GetCrosswalks retrieves crosswalk relationships between code types (Phase 2)
func (m *MockKeywordRepository) GetCrosswalks(
	ctx context.Context,
	fromCodeType string,
	fromCode string,
	toCodeType string,
) []struct {
	ToCode        string
	ToDescription string
} {
	result := []struct {
		ToCode        string
		ToDescription string
	}{}
	
	// Mock implementation - find codes of the same industry but different type
	for _, codes := range m.codesByIndustry {
		for _, code := range codes {
			if code.CodeType == fromCodeType && code.Code == fromCode {
				// Find related codes of target type
				for _, relatedCode := range codes {
					if relatedCode.CodeType == toCodeType {
						result = append(result, struct {
							ToCode        string
							ToDescription string
						}{
							ToCode:        relatedCode.Code,
							ToDescription: relatedCode.Description,
						})
						if len(result) >= 5 {
							return result
						}
					}
				}
			}
		}
	}
	
	return result
}

// GetIndustriesByKeyword returns industries matching a keyword with minimum weight (Phase 2: Fast path)
func (m *MockKeywordRepository) GetIndustriesByKeyword(
	ctx context.Context,
	keyword string,
	minWeight float64,
) []struct {
	Name   string
	Weight float64
} {
	result := []struct {
		Name   string
		Weight float64
	}{}
	
	// Find industries with this keyword
	for industryID, keywords := range m.keywordsByIndustry {
		for _, kw := range keywords {
			if kw.Keyword == keyword && kw.Weight >= minWeight {
				if industry, exists := m.industries[industryID]; exists {
					result = append(result, struct {
						Name   string
						Weight float64
					}{
						Name:   industry.Name,
						Weight: kw.Weight,
					})
					break // One match per industry
				}
			}
		}
	}
	
	return result
}

// MatchCodeEmbeddings performs vector similarity search for code embeddings (Phase 3)
func (m *MockKeywordRepository) MatchCodeEmbeddings(
	ctx context.Context,
	embedding []float64,
	codeType string,
	threshold float64,
	limit int,
) ([]repository.CodeMatch, error) {
	if err := m.errorMap["MatchCodeEmbeddings"]; err != nil {
		return nil, err
	}
	
	// Mock implementation - return codes matching the type with mock similarity scores
	result := []repository.CodeMatch{}
	
	// Get codes by type
	codes, exists := m.codesByType[codeType]
	if !exists {
		// Fallback: get codes from all industries
		for _, codesList := range m.codesByIndustry {
			for _, code := range codesList {
				if code.CodeType == codeType {
					codes = append(codes, code)
				}
			}
		}
	}
	
	// Return mock matches with similarity scores above threshold
	for i, code := range codes {
		if i >= limit {
			break
		}
		// Mock similarity score (decreasing with index)
		similarity := 0.9 - float64(i)*0.1
		if similarity >= threshold {
			result = append(result, repository.CodeMatch{
				Code:        code.Code,
				Description: code.Description,
				Similarity:  similarity,
			})
		}
	}
	
	return result, nil
}

// Phase 5: Classification Cache Methods

// GetCachedClassification retrieves a cached classification result
func (m *MockKeywordRepository) GetCachedClassification(
	ctx context.Context,
	contentHash string,
) (*repository.CachedClassificationResult, error) {
	if err := m.errorMap["GetCachedClassification"]; err != nil {
		return nil, err
	}
	// Mock implementation - return nil (cache miss) for testing
	return nil, nil
}

// SetCachedClassification stores a cached classification result
func (m *MockKeywordRepository) SetCachedClassification(
	ctx context.Context,
	contentHash string,
	businessName string,
	websiteURL string,
	result *repository.CachedClassificationResult,
) error {
	if err := m.errorMap["SetCachedClassification"]; err != nil {
		return err
	}
	// Mock implementation - no-op for testing
	return nil
}

// GetCacheStats retrieves cache statistics
func (m *MockKeywordRepository) GetCacheStats(ctx context.Context) (*repository.CacheStats, error) {
	if err := m.errorMap["GetCacheStats"]; err != nil {
		return nil, err
	}
	return &repository.CacheStats{
		TotalEntries: 0,
		HitRate:      0.0,
		AvgAge:       0,
		ExpiringSoon: 0,
	}, nil
}

// Phase 5: Metrics Logging

// LogClassificationMetrics logs classification metrics
func (m *MockKeywordRepository) LogClassificationMetrics(
	ctx context.Context,
	metrics *repository.ClassificationMetricsRecord,
) error {
	if err := m.errorMap["LogClassificationMetrics"]; err != nil {
		return err
	}
	// Mock implementation - no-op for testing
	return nil
}

// GetDashboardSummary retrieves dashboard summary data
func (m *MockKeywordRepository) GetDashboardSummary(
	ctx context.Context,
	days int,
) ([]*repository.DashboardMetric, error) {
	if err := m.errorMap["GetDashboardSummary"]; err != nil {
		return nil, err
	}
	return []*repository.DashboardMetric{}, nil
}

// GetTimeSeriesData retrieves time series data
func (m *MockKeywordRepository) GetTimeSeriesData(
	ctx context.Context,
	days int,
) ([]*repository.TimeSeriesData, error) {
	if err := m.errorMap["GetTimeSeriesData"]; err != nil {
		return nil, err
	}
	return []*repository.TimeSeriesData{}, nil
}

