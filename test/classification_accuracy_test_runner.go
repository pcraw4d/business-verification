package test

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/classification"
	"github.com/pcraw4d/business-verification/internal/classification/repository"
)

// KeywordRepository interface for testing
type KeywordRepository interface {
	GetKeywordsByIndustry(ctx context.Context, industryID int) ([]*IndustryKeyword, error)
	GetClassificationCodesByIndustry(ctx context.Context, industryID int) ([]*ClassificationCode, error)
	GetClassificationCodesByType(ctx context.Context, codeType string) ([]*ClassificationCode, error)
}

// IndustryKeyword represents a keyword associated with an industry
type IndustryKeyword struct {
	ID         int     `json:"id"`
	IndustryID int     `json:"industry_id"`
	Keyword    string  `json:"keyword"`
	Weight     float64 `json:"weight"`
	IsActive   bool    `json:"is_active"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// ClassificationCode represents industry classification codes (NAICS, MCC, SIC)
type ClassificationCode struct {
	ID          int    `json:"id"`
	IndustryID  int    `json:"industry_id"`
	CodeType    string `json:"code_type"`
	Code        string `json:"code"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ClassificationAccuracyTestRunner runs comprehensive classification accuracy tests
type ClassificationAccuracyTestRunner struct {
	classifier *classification.ClassificationCodeGenerator
	dataset    *ComprehensiveTestDataset
	logger     *log.Logger
}

// NewClassificationAccuracyTestRunner creates a new test runner
func NewClassificationAccuracyTestRunner(repo KeywordRepository, logger *log.Logger) *ClassificationAccuracyTestRunner {
	if logger == nil {
		logger = log.Default()
	}

	// Create a wrapper that implements the repository interface
	wrapper := &repositoryWrapper{repo: repo}
	classifier := classification.NewClassificationCodeGenerator(wrapper, logger)
	dataset := NewComprehensiveTestDataset()

	return &ClassificationAccuracyTestRunner{
		classifier: classifier,
		dataset:    dataset,
		logger:     logger,
	}
}

// NewClassificationAccuracyTestRunnerWithMock creates a new test runner with mock repository
func NewClassificationAccuracyTestRunnerWithMock(logger *log.Logger) *ClassificationAccuracyTestRunner {
	if logger == nil {
		logger = log.Default()
	}

	// Create mock repository
	mockRepo := NewMockKeywordRepository()

	// Create a wrapper that implements the repository interface
	wrapper := &repositoryWrapper{repo: mockRepo}
	classifier := classification.NewClassificationCodeGenerator(wrapper, logger)
	dataset := NewComprehensiveTestDataset()

	return &ClassificationAccuracyTestRunner{
		classifier: classifier,
		dataset:    dataset,
		logger:     logger,
	}
}

// repositoryWrapper wraps our test repository to implement the classification repository interface
type repositoryWrapper struct {
	repo KeywordRepository
}

func (w *repositoryWrapper) GetKeywordsByIndustry(ctx context.Context, industryID int) ([]*repository.IndustryKeyword, error) {
	keywords, err := w.repo.GetKeywordsByIndustry(ctx, industryID)
	if err != nil {
		return nil, err
	}

	// Convert to repository types
	result := make([]*repository.IndustryKeyword, len(keywords))
	for i, kw := range keywords {
		result[i] = &repository.IndustryKeyword{
			ID:         kw.ID,
			IndustryID: kw.IndustryID,
			Keyword:    kw.Keyword,
			Weight:     kw.Weight,
			IsActive:   kw.IsActive,
			CreatedAt:  kw.CreatedAt,
			UpdatedAt:  kw.UpdatedAt,
		}
	}

	return result, nil
}

func (w *repositoryWrapper) GetClassificationCodesByIndustry(ctx context.Context, industryID int) ([]*repository.ClassificationCode, error) {
	codes, err := w.repo.GetClassificationCodesByIndustry(ctx, industryID)
	if err != nil {
		return nil, err
	}

	// Convert to repository types
	result := make([]*repository.ClassificationCode, len(codes))
	for i, code := range codes {
		result[i] = &repository.ClassificationCode{
			ID:          code.ID,
			IndustryID:  code.IndustryID,
			CodeType:    code.CodeType,
			Code:        code.Code,
			Description: code.Description,
			IsActive:    code.IsActive,
			CreatedAt:   code.CreatedAt,
			UpdatedAt:   code.UpdatedAt,
		}
	}

	return result, nil
}

func (w *repositoryWrapper) GetClassificationCodesByType(ctx context.Context, codeType string) ([]*repository.ClassificationCode, error) {
	codes, err := w.repo.GetClassificationCodesByType(ctx, codeType)
	if err != nil {
		return nil, err
	}

	// Convert to repository types
	result := make([]*repository.ClassificationCode, len(codes))
	for i, code := range codes {
		result[i] = &repository.ClassificationCode{
			ID:          code.ID,
			IndustryID:  code.IndustryID,
			CodeType:    code.CodeType,
			Code:        code.Code,
			Description: code.Description,
			IsActive:    code.IsActive,
			CreatedAt:   code.CreatedAt,
			UpdatedAt:   code.UpdatedAt,
		}
	}

	return result, nil
}

// AddClassificationCode adds a new classification code (required by repository interface)
func (w *repositoryWrapper) AddClassificationCode(ctx context.Context, code *repository.ClassificationCode) error {
	// Convert to test type
	testCode := &ClassificationCode{
		ID:          code.ID,
		IndustryID:  code.IndustryID,
		CodeType:    code.CodeType,
		Code:        code.Code,
		Description: code.Description,
		IsActive:    code.IsActive,
		CreatedAt:   code.CreatedAt,
		UpdatedAt:   code.UpdatedAt,
	}

	// Add to mock repository if it supports it
	if mockRepo, ok := w.repo.(*MockKeywordRepository); ok {
		mockRepo.AddClassificationCode(code.IndustryID, testCode)
	}

	return nil
}

// UpdateClassificationCode updates an existing classification code (required by repository interface)
func (w *repositoryWrapper) UpdateClassificationCode(ctx context.Context, code *repository.ClassificationCode) error {
	// Mock implementation - just return success
	return nil
}

// DeleteClassificationCode deletes a classification code (required by repository interface)
func (w *repositoryWrapper) DeleteClassificationCode(ctx context.Context, id int) error {
	// Mock implementation - just return success
	return nil
}

// AddKeywordToIndustry adds a keyword to an industry (required by repository interface)
func (w *repositoryWrapper) AddKeywordToIndustry(ctx context.Context, industryID int, keyword string, weight float64) error {
	// Create test keyword
	testKeyword := &IndustryKeyword{
		ID:         0, // Will be set by mock repository
		IndustryID: industryID,
		Keyword:    keyword,
		Weight:     weight,
		IsActive:   true,
		CreatedAt:  "2025-01-01T00:00:00Z",
		UpdatedAt:  "2025-01-01T00:00:00Z",
	}

	// Add to mock repository if it supports it
	if mockRepo, ok := w.repo.(*MockKeywordRepository); ok {
		mockRepo.AddKeyword(industryID, testKeyword)
	}

	return nil
}

// AddPattern adds a pattern to an industry (required by repository interface)
func (w *repositoryWrapper) AddPattern(ctx context.Context, pattern *repository.IndustryPattern) error {
	// Mock implementation - just return success
	return nil
}

// BulkDeleteKeywords deletes multiple keywords (required by repository interface)
func (w *repositoryWrapper) BulkDeleteKeywords(ctx context.Context, keywordIDs []int) error {
	// Mock implementation - just return success
	return nil
}

// BulkInsertKeywords inserts multiple keywords (required by repository interface)
func (w *repositoryWrapper) BulkInsertKeywords(ctx context.Context, keywords []*repository.IndustryKeyword) error {
	// Mock implementation - just return success
	return nil
}

// BulkUpdateKeywords updates multiple keywords (required by repository interface)
func (w *repositoryWrapper) BulkUpdateKeywords(ctx context.Context, keywords []*repository.IndustryKeyword) error {
	// Mock implementation - just return success
	return nil
}

// ClassifyBusiness classifies a business using keywords (required by repository interface)
func (w *repositoryWrapper) ClassifyBusiness(ctx context.Context, businessName, description, websiteURL string) (*repository.ClassificationResult, error) {
	// Mock implementation - return empty result
	return &repository.ClassificationResult{}, nil
}

// ClassifyBusinessByKeywords classifies a business by keywords (required by repository interface)
func (w *repositoryWrapper) ClassifyBusinessByKeywords(ctx context.Context, keywords []string) (*repository.ClassificationResult, error) {
	// Mock implementation - return empty result
	return &repository.ClassificationResult{}, nil
}

// CleanupInactiveData cleans up inactive data (required by repository interface)
func (w *repositoryWrapper) CleanupInactiveData(ctx context.Context) error {
	// Mock implementation - just return success
	return nil
}

// CreateIndustry creates a new industry (required by repository interface)
func (w *repositoryWrapper) CreateIndustry(ctx context.Context, industry *repository.Industry) error {
	// Mock implementation - just return success
	return nil
}

// DeleteIndustry deletes an industry (required by repository interface)
func (w *repositoryWrapper) DeleteIndustry(ctx context.Context, industryID int) error {
	// Mock implementation - just return success
	return nil
}

// DeletePattern deletes a pattern (required by repository interface)
func (w *repositoryWrapper) DeletePattern(ctx context.Context, patternID int) error {
	// Mock implementation - just return success
	return nil
}

// GetBatchClassificationCodes gets classification codes in batch (required by repository interface)
func (w *repositoryWrapper) GetBatchClassificationCodes(ctx context.Context, industryIDs []int) (map[int][]*repository.ClassificationCode, error) {
	// Mock implementation - return empty map
	return make(map[int][]*repository.ClassificationCode), nil
}

// GetBatchIndustries gets industries in batch (required by repository interface)
func (w *repositoryWrapper) GetBatchIndustries(ctx context.Context, industryIDs []int) (map[int]*repository.Industry, error) {
	// Mock implementation - return empty map
	return make(map[int]*repository.Industry), nil
}

// GetBatchKeywords gets keywords in batch (required by repository interface)
func (w *repositoryWrapper) GetBatchKeywords(ctx context.Context, industryIDs []int) (map[int][]*repository.KeywordWeight, error) {
	// Mock implementation - return empty map
	return make(map[int][]*repository.KeywordWeight), nil
}

// GetCachedClassificationCodes gets cached classification codes (required by repository interface)
func (w *repositoryWrapper) GetCachedClassificationCodes(ctx context.Context, industryID int) ([]*repository.ClassificationCode, error) {
	// Mock implementation - return empty slice
	return []*repository.ClassificationCode{}, nil
}

// GetCachedClassificationCodesByType gets cached classification codes by type (required by repository interface)
func (w *repositoryWrapper) GetCachedClassificationCodesByType(ctx context.Context, codeType string) ([]*repository.ClassificationCode, error) {
	// Mock implementation - return empty slice
	return []*repository.ClassificationCode{}, nil
}

// GetDatabaseStats gets database statistics (required by repository interface)
func (w *repositoryWrapper) GetDatabaseStats(ctx context.Context) (map[string]interface{}, error) {
	// Mock implementation - return empty map
	return make(map[string]interface{}), nil
}

// GetIndustryByID gets an industry by ID (required by repository interface)
func (w *repositoryWrapper) GetIndustryByID(ctx context.Context, industryID int) (*repository.Industry, error) {
	// Mock implementation - return a test industry
	return &repository.Industry{
		ID:          industryID,
		Name:        "Test Industry",
		Description: "Test industry for classification testing",
		IsActive:    true,
		CreatedAt:   "2025-01-01T00:00:00Z",
		UpdatedAt:   "2025-01-01T00:00:00Z",
	}, nil
}

// GetIndustryByName gets an industry by name (required by repository interface)
func (w *repositoryWrapper) GetIndustryByName(ctx context.Context, name string) (*repository.Industry, error) {
	// Mock implementation - return a test industry based on the name
	return &repository.Industry{
		ID:          1,
		Name:        name,
		Description: "Test industry for classification testing",
		IsActive:    true,
		CreatedAt:   "2025-01-01T00:00:00Z",
		UpdatedAt:   "2025-01-01T00:00:00Z",
	}, nil
}

// GetIndustryCodeCacheStats gets industry code cache statistics (required by repository interface)
func (w *repositoryWrapper) GetIndustryCodeCacheStats() *repository.IndustryCodeCacheStats {
	// Mock implementation - return nil
	return nil
}

// GetIndustryStatistics gets industry statistics (required by repository interface)
func (w *repositoryWrapper) GetIndustryStatistics(ctx context.Context) (map[string]interface{}, error) {
	// Mock implementation - return empty map
	return make(map[string]interface{}), nil
}

// GetKeywordFrequency gets keyword frequency (required by repository interface)
func (w *repositoryWrapper) GetKeywordFrequency(ctx context.Context, industryID int) (map[string]int, error) {
	// Mock implementation - return empty map
	return make(map[string]int), nil
}

// GetKeywordWeights gets keyword weights (required by repository interface)
func (w *repositoryWrapper) GetKeywordWeights(ctx context.Context, keyword string) ([]*repository.KeywordWeight, error) {
	// Mock implementation - return empty slice
	return []*repository.KeywordWeight{}, nil
}

// GetPatternsByIndustry gets patterns by industry (required by repository interface)
func (w *repositoryWrapper) GetPatternsByIndustry(ctx context.Context, industryID int) ([]*repository.IndustryPattern, error) {
	// Mock implementation - return empty slice
	return []*repository.IndustryPattern{}, nil
}

// GetTopIndustriesByKeywords gets top industries by keywords (required by repository interface)
func (w *repositoryWrapper) GetTopIndustriesByKeywords(ctx context.Context, keywords []string, limit int) ([]*repository.Industry, error) {
	// Mock implementation - return empty slice
	return []*repository.Industry{}, nil
}

// IncrementUsageCount increments usage count (required by repository interface)
func (w *repositoryWrapper) IncrementUsageCount(ctx context.Context, keyword string, count int) error {
	// Mock implementation - just return success
	return nil
}

// InitializeIndustryCodeCache initializes industry code cache (required by repository interface)
func (w *repositoryWrapper) InitializeIndustryCodeCache(ctx context.Context) error {
	// Mock implementation - just return success
	return nil
}

// InvalidateIndustryCodeCache invalidates industry code cache (required by repository interface)
func (w *repositoryWrapper) InvalidateIndustryCodeCache(ctx context.Context, keywords []string) error {
	// Mock implementation - just return success
	return nil
}

// ListIndustries lists all industries (required by repository interface)
func (w *repositoryWrapper) ListIndustries(ctx context.Context, searchTerm string) ([]*repository.Industry, error) {
	// Mock implementation - return empty slice
	return []*repository.Industry{}, nil
}

// Ping checks database connectivity (required by repository interface)
func (w *repositoryWrapper) Ping(ctx context.Context) error {
	// Mock implementation - just return success
	return nil
}

// RemoveKeywordFromIndustry removes a keyword from an industry (required by repository interface)
func (w *repositoryWrapper) RemoveKeywordFromIndustry(ctx context.Context, keywordID int) error {
	// Mock implementation - just return success
	return nil
}

// SearchIndustriesByPattern searches industries by pattern (required by repository interface)
func (w *repositoryWrapper) SearchIndustriesByPattern(ctx context.Context, pattern string) ([]*repository.Industry, error) {
	// Mock implementation - return empty slice
	return []*repository.Industry{}, nil
}

// SearchKeywords searches keywords (required by repository interface)
func (w *repositoryWrapper) SearchKeywords(ctx context.Context, query string, limit int) ([]*repository.IndustryKeyword, error) {
	// Mock implementation - return empty slice
	return []*repository.IndustryKeyword{}, nil
}

// UpdateIndustry updates an industry (required by repository interface)
func (w *repositoryWrapper) UpdateIndustry(ctx context.Context, industry *repository.Industry) error {
	// Mock implementation - just return success
	return nil
}

// UpdateKeywordWeight updates keyword weight (required by repository interface)
func (w *repositoryWrapper) UpdateKeywordWeight(ctx context.Context, keywordID int, weight float64) error {
	// Mock implementation - just return success
	return nil
}

// UpdateKeywordWeightByID updates keyword weight by ID (required by repository interface)
func (w *repositoryWrapper) UpdateKeywordWeightByID(ctx context.Context, weight *repository.KeywordWeight) error {
	// Mock implementation - just return success
	return nil
}

// UpdatePattern updates a pattern (required by repository interface)
func (w *repositoryWrapper) UpdatePattern(ctx context.Context, pattern *repository.IndustryPattern) error {
	// Mock implementation - just return success
	return nil
}

// ClassificationAccuracyResult represents the result of accuracy validation
type ClassificationAccuracyResult struct {
	IsAccurate            bool
	AccuracyScore         float64
	MatchedIndustry       string
	ExpectedIndustry      string
	ConfidenceScore       float64
	CodeMatches           map[string]bool // MCC, SIC, NAICS
	CodeMappingValidation *CodeMappingValidation
}

// CodeMappingValidation represents validation results for industry code mappings
type CodeMappingValidation struct {
	MCCValidation   *CodeValidation
	SICValidation   *CodeValidation
	NAICSValidation *CodeValidation
	OverallScore    float64
}

// CodeValidation represents validation results for a specific code type
type CodeValidation struct {
	CodeType       string
	IsValid        bool
	ExpectedCodes  []string
	ActualCodes    []string
	MatchedCodes   []string
	MissingCodes   []string
	ExtraCodes     []string
	AccuracyScore  float64
	FormatValid    bool
	StructureValid bool
}

// ValidateClassificationAccuracy validates the accuracy of a classification result
func (runner *ClassificationAccuracyTestRunner) ValidateClassificationAccuracy(tc ClassificationTestCase, result *classification.ClassificationCodesInfo) ClassificationAccuracyResult {
	accuracy := ClassificationAccuracyResult{
		ExpectedIndustry: tc.ExpectedIndustry,
		CodeMatches:      make(map[string]bool),
	}

	// Calculate confidence score
	accuracy.ConfidenceScore = calculateOverallConfidence(result)

	// Determine matched industry based on the result
	// For now, we'll use a simple keyword matching approach
	matchedIndustry := runner.determineMatchedIndustry(tc, result)
	accuracy.MatchedIndustry = matchedIndustry

	// Check if the matched industry matches the expected industry
	accuracy.IsAccurate = runner.isIndustryMatch(matchedIndustry, tc.ExpectedIndustry)

	// Calculate accuracy score based on industry match and confidence
	if accuracy.IsAccurate {
		accuracy.AccuracyScore = 100.0
	} else {
		// Partial credit based on confidence and keyword overlap
		accuracy.AccuracyScore = accuracy.ConfidenceScore * 0.8
	}

	// Validate code mapping accuracy
	accuracy.CodeMappingValidation = runner.ValidateCodeMapping(tc, result)

	// Validate code matches based on mapping validation
	accuracy.CodeMatches["MCC"] = accuracy.CodeMappingValidation.MCCValidation.IsValid
	accuracy.CodeMatches["SIC"] = accuracy.CodeMappingValidation.SICValidation.IsValid
	accuracy.CodeMatches["NAICS"] = accuracy.CodeMappingValidation.NAICSValidation.IsValid

	return accuracy
}

// determineMatchedIndustry determines which industry was matched based on keywords and result
func (runner *ClassificationAccuracyTestRunner) determineMatchedIndustry(tc ClassificationTestCase, result *classification.ClassificationCodesInfo) string {
	// Simple keyword-based industry detection
	keywordsStr := strings.Join(tc.Keywords, " ")
	keywords := strings.ToLower(keywordsStr)
	businessName := strings.ToLower(tc.BusinessName)
	description := strings.ToLower(tc.Description)

	// Combine all text for analysis
	combinedText := keywords + " " + businessName + " " + description

	// Industry keyword mappings
	industryKeywords := map[string][]string{
		"Software Development":     {"software", "development", "programming", "tech", "technology", "cloud", "ai", "artificial intelligence", "machine learning"},
		"Cloud Computing":          {"cloud", "computing", "infrastructure", "platform", "saas", "paas", "iaas"},
		"AI/ML Startup":            {"ai", "artificial intelligence", "machine learning", "neural", "algorithm", "data science"},
		"Medical Center":           {"medical", "healthcare", "clinic", "hospital", "patient", "health"},
		"Medical Technology":       {"medical", "technology", "devices", "healthcare", "medtech", "biomedical"},
		"Pharmaceutical":           {"pharmaceutical", "pharmacy", "drug", "medicine", "research", "clinical"},
		"Commercial Bank":          {"bank", "banking", "financial", "loans", "deposits", "credit"},
		"Fintech Startup":          {"fintech", "financial technology", "payment", "digital banking", "fintech"},
		"Insurance Company":        {"insurance", "coverage", "policy", "claims", "underwriting"},
		"Online Retail":            {"retail", "shopping", "e-commerce", "online", "consumer", "goods"},
		"E-commerce Platform":      {"e-commerce", "marketplace", "online", "platform", "retail"},
		"Industrial Manufacturing": {"manufacturing", "industrial", "factory", "production", "machinery"},
		"Food Manufacturing":       {"food", "manufacturing", "processing", "packaged", "beverages"},
		"Management Consulting":    {"consulting", "management", "strategy", "business", "advisory"},
		"Legal Services":           {"legal", "law", "attorney", "litigation", "corporate", "lawyer"},
		"Real Estate Agency":       {"real estate", "property", "residential", "commercial", "sales"},
		"Educational Technology":   {"education", "e-learning", "online learning", "educational", "learning"},
		"Renewable Energy":         {"renewable energy", "solar", "wind power", "clean technology", "sustainable"},
		"Mixed Industry Business":  {"mixed", "diversified", "multiple", "various"},
		"Generic Business":         {"business", "services", "consulting", "solutions"},
		"Very Short Description":   {"short", "brief", "minimal"},
	}

	// Find the best matching industry
	bestMatch := ""
	maxScore := 0.0

	for industry, keywords := range industryKeywords {
		score := 0.0
		for _, keyword := range keywords {
			if strings.Contains(combinedText, keyword) {
				score += 1.0
			}
		}

		// Normalize score by number of keywords
		normalizedScore := score / float64(len(keywords))

		if normalizedScore > maxScore {
			maxScore = normalizedScore
			bestMatch = industry
		}
	}

	// If no good match found, return a default
	if maxScore < 0.1 {
		return "Generic Business"
	}

	return bestMatch
}

// ValidateCodeMapping validates the accuracy of industry code mappings
func (runner *ClassificationAccuracyTestRunner) ValidateCodeMapping(tc ClassificationTestCase, result *classification.ClassificationCodesInfo) *CodeMappingValidation {
	validation := &CodeMappingValidation{}

	// Get expected codes for the industry
	expectedCodes := runner.getExpectedCodesForIndustry(tc.ExpectedIndustry)

	// Convert code structs to string slices
	mccCodes := make([]string, len(result.MCC))
	for i, code := range result.MCC {
		mccCodes[i] = code.Code
	}

	sicCodes := make([]string, len(result.SIC))
	for i, code := range result.SIC {
		sicCodes[i] = code.Code
	}

	naicsCodes := make([]string, len(result.NAICS))
	for i, code := range result.NAICS {
		naicsCodes[i] = code.Code
	}

	// Validate MCC codes
	validation.MCCValidation = runner.validateCodeType("MCC", expectedCodes.MCC, mccCodes)

	// Validate SIC codes
	validation.SICValidation = runner.validateCodeType("SIC", expectedCodes.SIC, sicCodes)

	// Validate NAICS codes
	validation.NAICSValidation = runner.validateCodeType("NAICS", expectedCodes.NAICS, naicsCodes)

	// Calculate overall score
	validation.OverallScore = (validation.MCCValidation.AccuracyScore +
		validation.SICValidation.AccuracyScore +
		validation.NAICSValidation.AccuracyScore) / 3.0

	return validation
}

// ExpectedCodes represents expected codes for an industry
type ExpectedCodes struct {
	MCC   []string
	SIC   []string
	NAICS []string
}

// getExpectedCodesForIndustry returns expected codes for a given industry
func (runner *ClassificationAccuracyTestRunner) getExpectedCodesForIndustry(industry string) ExpectedCodes {
	// Define expected codes for each industry
	expectedCodesMap := map[string]ExpectedCodes{
		"Software Development": {
			MCC:   []string{"5734", "7372", "7373"},
			SIC:   []string{"7372", "7373", "7374"},
			NAICS: []string{"541511", "541512", "541513"},
		},
		"Cloud Computing": {
			MCC:   []string{"5734", "7372", "7373"},
			SIC:   []string{"7372", "7373", "7374"},
			NAICS: []string{"541511", "541512", "518210"},
		},
		"AI/ML Startup": {
			MCC:   []string{"5734", "7372", "7373"},
			SIC:   []string{"7372", "7373", "7374"},
			NAICS: []string{"541511", "541512", "541330"},
		},
		"Medical Center": {
			MCC:   []string{"8062", "8069", "8071"},
			SIC:   []string{"8062", "8069", "8071"},
			NAICS: []string{"621111", "621112", "621210"},
		},
		"Medical Technology": {
			MCC:   []string{"5047", "5045", "5046"},
			SIC:   []string{"5047", "5045", "5046"},
			NAICS: []string{"334510", "334511", "334512"},
		},
		"Pharmaceutical": {
			MCC:   []string{"5122", "5912", "5047"},
			SIC:   []string{"2834", "2835", "2836"},
			NAICS: []string{"325412", "325411", "325413"},
		},
		"Commercial Bank": {
			MCC:   []string{"6010", "6011", "6012"},
			SIC:   []string{"6021", "6022", "6029"},
			NAICS: []string{"522110", "522120", "522130"},
		},
		"Fintech Startup": {
			MCC:   []string{"6010", "6011", "6012"},
			SIC:   []string{"6021", "6022", "6029"},
			NAICS: []string{"522110", "522120", "523110"},
		},
		"Insurance Company": {
			MCC:   []string{"6300", "6310", "6321"},
			SIC:   []string{"6311", "6321", "6331"},
			NAICS: []string{"524113", "524114", "524126"},
		},
		"Online Retail": {
			MCC:   []string{"5310", "5311", "5312"},
			SIC:   []string{"5311", "5312", "5313"},
			NAICS: []string{"454110", "454111", "454112"},
		},
		"E-commerce Platform": {
			MCC:   []string{"5310", "5311", "5312"},
			SIC:   []string{"5311", "5312", "5313"},
			NAICS: []string{"454110", "454111", "454112"},
		},
		"Restaurant": {
			MCC:   []string{"5812", "5813", "5814"},
			SIC:   []string{"5812", "5813", "5814"},
			NAICS: []string{"722511", "722512", "722513"},
		},
		"Food Truck": {
			MCC:   []string{"5812", "5813", "5814"},
			SIC:   []string{"5812", "5813", "5814"},
			NAICS: []string{"722330", "722511", "722512"},
		},
		"Construction Company": {
			MCC:   []string{"1771", "1799", "1521"},
			SIC:   []string{"1521", "1522", "1541"},
			NAICS: []string{"236116", "236117", "236118"},
		},
		"Real Estate Agency": {
			MCC:   []string{"6513", "6514", "6515"},
			SIC:   []string{"6531", "6532", "6533"},
			NAICS: []string{"531210", "531311", "531312"},
		},
		"Law Firm": {
			MCC:   []string{"8111", "8112", "8113"},
			SIC:   []string{"8111", "8112", "8113"},
			NAICS: []string{"541110", "541191", "541199"},
		},
		"Consulting Firm": {
			MCC:   []string{"7392", "7393", "7394"},
			SIC:   []string{"8742", "8748", "8999"},
			NAICS: []string{"541611", "541612", "541613"},
		},
		"Marketing Agency": {
			MCC:   []string{"7310", "7311", "7312"},
			SIC:   []string{"7311", "7312", "7313"},
			NAICS: []string{"541810", "541820", "541830"},
		},
		"Manufacturing Company": {
			MCC:   []string{"5085", "5087", "5088"},
			SIC:   []string{"3089", "3081", "3082"},
			NAICS: []string{"311111", "311112", "311113"},
		},
		"Transportation Company": {
			MCC:   []string{"4111", "4119", "4121"},
			SIC:   []string{"4111", "4119", "4121"},
			NAICS: []string{"484110", "484121", "484122"},
		},
		"Generic Business": {
			MCC:   []string{"7399", "8999", "9999"},
			SIC:   []string{"8999", "9999", "0000"},
			NAICS: []string{"999999", "000000", "111111"},
		},
	}

	if codes, exists := expectedCodesMap[industry]; exists {
		return codes
	}

	// Return default codes for unknown industries
	return ExpectedCodes{
		MCC:   []string{"7399", "8999"},
		SIC:   []string{"8999", "9999"},
		NAICS: []string{"999999", "000000"},
	}
}

// validateCodeType validates a specific type of industry code
func (runner *ClassificationAccuracyTestRunner) validateCodeType(codeType string, expectedCodes []string, actualCodes []string) *CodeValidation {
	validation := &CodeValidation{
		CodeType:      codeType,
		ExpectedCodes: expectedCodes,
		ActualCodes:   actualCodes,
	}

	// Check format validity
	validation.FormatValid = runner.validateCodeFormat(codeType, actualCodes)

	// Check structure validity
	validation.StructureValid = runner.validateCodeStructure(codeType, actualCodes)

	// Find matched, missing, and extra codes
	validation.MatchedCodes = runner.findMatchedCodes(expectedCodes, actualCodes)
	validation.MissingCodes = runner.findMissingCodes(expectedCodes, actualCodes)
	validation.ExtraCodes = runner.findExtraCodes(expectedCodes, actualCodes)

	// Calculate accuracy score
	validation.AccuracyScore = runner.calculateCodeAccuracyScore(validation)

	// Determine if validation passed
	validation.IsValid = validation.AccuracyScore >= 0.7 && validation.FormatValid && validation.StructureValid

	return validation
}

// validateCodeFormat validates the format of industry codes
func (runner *ClassificationAccuracyTestRunner) validateCodeFormat(codeType string, codes []string) bool {
	for _, code := range codes {
		switch codeType {
		case "MCC":
			// MCC codes should be 4 digits
			if len(code) != 4 {
				return false
			}
			// Check if all characters are digits
			for _, char := range code {
				if char < '0' || char > '9' {
					return false
				}
			}
		case "SIC":
			// SIC codes should be 4 digits
			if len(code) != 4 {
				return false
			}
			// Check if all characters are digits
			for _, char := range code {
				if char < '0' || char > '9' {
					return false
				}
			}
		case "NAICS":
			// NAICS codes should be 6 digits
			if len(code) != 6 {
				return false
			}
			// Check if all characters are digits
			for _, char := range code {
				if char < '0' || char > '9' {
					return false
				}
			}
		}
	}
	return true
}

// validateCodeStructure validates the structure of industry codes
func (runner *ClassificationAccuracyTestRunner) validateCodeStructure(codeType string, codes []string) bool {
	// Basic structure validation - codes should not be empty and should be unique
	if len(codes) == 0 {
		return false
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for _, code := range codes {
		if seen[code] {
			return false
		}
		seen[code] = true
	}

	return true
}

// findMatchedCodes finds codes that match between expected and actual
func (runner *ClassificationAccuracyTestRunner) findMatchedCodes(expected, actual []string) []string {
	var matched []string
	actualMap := make(map[string]bool)

	for _, code := range actual {
		actualMap[code] = true
	}

	for _, code := range expected {
		if actualMap[code] {
			matched = append(matched, code)
		}
	}

	return matched
}

// findMissingCodes finds codes that are expected but not present
func (runner *ClassificationAccuracyTestRunner) findMissingCodes(expected, actual []string) []string {
	var missing []string
	actualMap := make(map[string]bool)

	for _, code := range actual {
		actualMap[code] = true
	}

	for _, code := range expected {
		if !actualMap[code] {
			missing = append(missing, code)
		}
	}

	return missing
}

// findExtraCodes finds codes that are present but not expected
func (runner *ClassificationAccuracyTestRunner) findExtraCodes(expected, actual []string) []string {
	var extra []string
	expectedMap := make(map[string]bool)

	for _, code := range expected {
		expectedMap[code] = true
	}

	for _, code := range actual {
		if !expectedMap[code] {
			extra = append(extra, code)
		}
	}

	return extra
}

// calculateCodeAccuracyScore calculates the accuracy score for code validation
func (runner *ClassificationAccuracyTestRunner) calculateCodeAccuracyScore(validation *CodeValidation) float64 {
	if len(validation.ExpectedCodes) == 0 {
		return 0.0
	}

	// Calculate precision (matched / actual)
	precision := 0.0
	if len(validation.ActualCodes) > 0 {
		precision = float64(len(validation.MatchedCodes)) / float64(len(validation.ActualCodes))
	}

	// Calculate recall (matched / expected)
	recall := float64(len(validation.MatchedCodes)) / float64(len(validation.ExpectedCodes))

	// Calculate F1 score (harmonic mean of precision and recall)
	if precision+recall == 0 {
		return 0.0
	}

	return 2 * (precision * recall) / (precision + recall)
}

// RunClassification runs classification for a test case
func (runner *ClassificationAccuracyTestRunner) RunClassification(tc ClassificationTestCase) (*classification.ClassificationCodesInfo, error) {
	// Generate classification codes using the keywords and expected industry
	result, err := runner.classifier.GenerateClassificationCodes(
		context.Background(),
		tc.Keywords,
		tc.ExpectedIndustry,
		0.8, // Default confidence score
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// isIndustryMatch checks if the matched industry matches the expected industry
func (runner *ClassificationAccuracyTestRunner) isIndustryMatch(matched, expected string) bool {
	// Exact match
	if matched == expected {
		return true
	}

	// Fuzzy matching for similar industries
	similarIndustries := map[string][]string{
		"Software Development":     {"Cloud Computing", "AI/ML Startup"},
		"Cloud Computing":          {"Software Development", "AI/ML Startup"},
		"AI/ML Startup":            {"Software Development", "Cloud Computing"},
		"Medical Center":           {"Medical Technology", "Pharmaceutical"},
		"Medical Technology":       {"Medical Center", "Pharmaceutical"},
		"Pharmaceutical":           {"Medical Center", "Medical Technology"},
		"Commercial Bank":          {"Fintech Startup", "Insurance Company"},
		"Fintech Startup":          {"Commercial Bank", "Insurance Company"},
		"Insurance Company":        {"Commercial Bank", "Fintech Startup"},
		"Online Retail":            {"E-commerce Platform"},
		"E-commerce Platform":      {"Online Retail"},
		"Industrial Manufacturing": {"Food Manufacturing"},
		"Food Manufacturing":       {"Industrial Manufacturing"},
		"Management Consulting":    {"Legal Services"},
		"Legal Services":           {"Management Consulting"},
		"Educational Technology":   {"Software Development"},
		"Renewable Energy":         {"Industrial Manufacturing"},
	}

	if similar, exists := similarIndustries[expected]; exists {
		for _, similarIndustry := range similar {
			if matched == similarIndustry {
				return true
			}
		}
	}

	return false
}

// ConfidenceScoreReliabilityResult represents the result of confidence score reliability testing
type ConfidenceScoreReliabilityResult struct {
	TestName          string
	TotalTests        int
	ValidScores       int
	InvalidScores     int
	ScoreRange        *ScoreRange
	DistributionStats *DistributionStats
	ConsistencyStats  *ConsistencyStats
	ReliabilityScore  float64
	IsReliable        bool
}

// ScoreRange represents the range of confidence scores
type ScoreRange struct {
	MinScore    float64
	MaxScore    float64
	AvgScore    float64
	MedianScore float64
}

// DistributionStats represents the distribution of confidence scores
type DistributionStats struct {
	HighConfidence    int // >= 0.8
	MediumConfidence  int // 0.5 - 0.79
	LowConfidence     int // 0.2 - 0.49
	VeryLowConfidence int // < 0.2
}

// ConsistencyStats represents consistency metrics for confidence scores
type ConsistencyStats struct {
	Variance               float64
	StandardDeviation      float64
	CoefficientOfVariation float64
	ConsistentScores       int
	InconsistentScores     int
}

// ManualClassificationResult represents manual classification results for comparison
type ManualClassificationResult struct {
	TestCaseName        string
	ManualIndustry      string
	ManualMCC           []string
	ManualSIC           []string
	ManualNAICS         []string
	ManualConfidence    float64
	ClassificationNotes string
	Difficulty          string
}

// ClassificationComparisonResult represents comparison between automated and manual classification
type ClassificationComparisonResult struct {
	TestCaseName         string
	AutomatedIndustry    string
	ManualIndustry       string
	IndustryMatch        bool
	IndustryAccuracy     float64
	MCCComparison        *CodeComparison
	SICComparison        *CodeComparison
	NAICSComparison      *CodeComparison
	ConfidenceComparison *ConfidenceComparison
	OverallAccuracy      float64
	DiscrepancyLevel     string // "Low", "Medium", "High"
	ComparisonNotes      string
}

// CodeComparison represents comparison between automated and manual codes
type CodeComparison struct {
	CodeType       string
	AutomatedCodes []string
	ManualCodes    []string
	MatchedCodes   []string
	MissingCodes   []string
	ExtraCodes     []string
	Precision      float64
	Recall         float64
	F1Score        float64
	Accuracy       float64
}

// ConfidenceComparison represents comparison between automated and manual confidence scores
type ConfidenceComparison struct {
	AutomatedConfidence float64
	ManualConfidence    float64
	ConfidenceDiff      float64
	ConfidenceMatch     bool
	ConfidenceAccuracy  float64
}

// RunManualClassificationComparisonTest runs comprehensive manual classification comparison tests
func (runner *ClassificationAccuracyTestRunner) RunManualClassificationComparisonTest(t *testing.T) {
	t.Log("📋 Running Manual Classification Comparison Tests...")

	dataset := NewComprehensiveTestDataset()
	manualBaselines := runner.getManualClassificationBaselines()

	totalTests := 0
	accurateTests := 0
	totalAccuracy := 0.0

	var comparisonResults []ClassificationComparisonResult

	for _, tc := range dataset.TestCases {
		// Get manual baseline for this test case
		manualBaseline, exists := manualBaselines[tc.Name]
		if !exists {
			t.Logf("⚠️ No manual baseline found for %s, skipping comparison", tc.Name)
			continue
		}

		// Run automated classification
		automatedResult, err := runner.RunClassification(tc)
		if err != nil {
			t.Errorf("❌ Automated classification failed for %s: %v", tc.Name, err)
			continue
		}

		// Compare results
		comparison := runner.compareClassificationResults(tc.Name, tc.ExpectedIndustry, automatedResult, manualBaseline)
		comparisonResults = append(comparisonResults, comparison)

		// Update statistics
		totalTests++
		totalAccuracy += comparison.OverallAccuracy

		if comparison.OverallAccuracy >= 0.8 {
			accurateTests++
		}

		// Log detailed comparison
		t.Logf("📊 %s Comparison:", tc.Name)
		t.Logf("   Industry Match: %v (Auto: %s, Manual: %s)",
			comparison.IndustryMatch, comparison.AutomatedIndustry, comparison.ManualIndustry)
		t.Logf("   MCC Accuracy: %.2f (Precision: %.2f, Recall: %.2f, F1: %.2f)",
			comparison.MCCComparison.Accuracy, comparison.MCCComparison.Precision,
			comparison.MCCComparison.Recall, comparison.MCCComparison.F1Score)
		t.Logf("   SIC Accuracy: %.2f (Precision: %.2f, Recall: %.2f, F1: %.2f)",
			comparison.SICComparison.Accuracy, comparison.SICComparison.Precision,
			comparison.SICComparison.Recall, comparison.SICComparison.F1Score)
		t.Logf("   NAICS Accuracy: %.2f (Precision: %.2f, Recall: %.2f, F1: %.2f)",
			comparison.NAICSComparison.Accuracy, comparison.NAICSComparison.Precision,
			comparison.NAICSComparison.Recall, comparison.NAICSComparison.F1Score)
		t.Logf("   Confidence Match: %v (Auto: %.2f, Manual: %.2f, Diff: %.2f)",
			comparison.ConfidenceComparison.ConfidenceMatch,
			comparison.ConfidenceComparison.AutomatedConfidence,
			comparison.ConfidenceComparison.ManualConfidence,
			comparison.ConfidenceComparison.ConfidenceDiff)
		t.Logf("   Overall Accuracy: %.2f (Discrepancy: %s)",
			comparison.OverallAccuracy, comparison.DiscrepancyLevel)
	}

	// Calculate and log final statistics
	avgAccuracy := totalAccuracy / float64(totalTests)
	passRate := float64(accurateTests) / float64(totalTests) * 100

	t.Logf("📊 Manual Classification Comparison Results:")
	t.Logf("   Total Tests: %d", totalTests)
	t.Logf("   Accurate Tests: %d", accurateTests)
	t.Logf("   Pass Rate: %.1f%%", passRate)
	t.Logf("   Average Accuracy: %.2f", avgAccuracy)

	// Analyze discrepancies
	discrepancyAnalysis := runner.analyzeDiscrepancies(comparisonResults)
	t.Logf("📈 Discrepancy Analysis:")
	t.Logf("   Low Discrepancy: %d (%.1f%%)", discrepancyAnalysis.LowDiscrepancy,
		float64(discrepancyAnalysis.LowDiscrepancy)/float64(totalTests)*100)
	t.Logf("   Medium Discrepancy: %d (%.1f%%)", discrepancyAnalysis.MediumDiscrepancy,
		float64(discrepancyAnalysis.MediumDiscrepancy)/float64(totalTests)*100)
	t.Logf("   High Discrepancy: %d (%.1f%%)", discrepancyAnalysis.HighDiscrepancy,
		float64(discrepancyAnalysis.HighDiscrepancy)/float64(totalTests)*100)

	// Assert minimum accuracy threshold
	if avgAccuracy < 0.7 {
		t.Errorf("❌ Manual comparison accuracy too low: %.2f (expected >= 0.7)", avgAccuracy)
	}

	if passRate < 60.0 {
		t.Errorf("❌ Manual comparison pass rate too low: %.1f%% (expected >= 60%%)", passRate)
	}
}

// getManualClassificationBaselines returns manual classification baselines for test cases
func (runner *ClassificationAccuracyTestRunner) getManualClassificationBaselines() map[string]ManualClassificationResult {
	return map[string]ManualClassificationResult{
		"Software Development Company": {
			TestCaseName:        "Software Development Company",
			ManualIndustry:      "Technology",
			ManualMCC:           []string{"7372", "7373", "7374"},
			ManualSIC:           []string{"7372", "7373", "7374"},
			ManualNAICS:         []string{"541511", "541512", "541513"},
			ManualConfidence:    0.95,
			ClassificationNotes: "Clear software development business with technology focus",
			Difficulty:          "Easy",
		},
		"Cloud Computing Provider": {
			TestCaseName:        "Cloud Computing Provider",
			ManualIndustry:      "Cloud Computing",
			ManualMCC:           []string{"7372", "7373", "7374"},
			ManualSIC:           []string{"7372", "7373", "7374"},
			ManualNAICS:         []string{"541511", "541512", "541513"},
			ManualConfidence:    0.90,
			ClassificationNotes: "Cloud infrastructure and services provider",
			Difficulty:          "Easy",
		},
		"AI/ML Startup": {
			TestCaseName:        "AI/ML Startup",
			ManualIndustry:      "Artificial Intelligence",
			ManualMCC:           []string{"7372", "7373", "7374"},
			ManualSIC:           []string{"7372", "7373", "7374"},
			ManualNAICS:         []string{"541511", "541512", "541513"},
			ManualConfidence:    0.85,
			ClassificationNotes: "AI and machine learning technology startup",
			Difficulty:          "Medium",
		},
		"Medical Center": {
			TestCaseName:        "Medical Center",
			ManualIndustry:      "Healthcare",
			ManualMCC:           []string{"8062", "8069", "8071"},
			ManualSIC:           []string{"8062", "8069", "8071"},
			ManualNAICS:         []string{"621111", "621112", "621113"},
			ManualConfidence:    0.95,
			ClassificationNotes: "Healthcare facility providing medical services",
			Difficulty:          "Easy",
		},
		"Medical Technology Company": {
			TestCaseName:        "Medical Technology Company",
			ManualIndustry:      "Medical Technology",
			ManualMCC:           []string{"5047", "5048", "5049"},
			ManualSIC:           []string{"5047", "5048", "5049"},
			ManualNAICS:         []string{"334510", "334511", "334512"},
			ManualConfidence:    0.90,
			ClassificationNotes: "Medical device and technology manufacturer",
			Difficulty:          "Medium",
		},
		"Pharmaceutical Company": {
			TestCaseName:        "Pharmaceutical Company",
			ManualIndustry:      "Pharmaceuticals",
			ManualMCC:           []string{"5122", "5123", "5124"},
			ManualSIC:           []string{"5122", "5123", "5124"},
			ManualNAICS:         []string{"325412", "325413", "325414"},
			ManualConfidence:    0.95,
			ClassificationNotes: "Pharmaceutical drug manufacturer and distributor",
			Difficulty:          "Easy",
		},
		"Commercial Bank": {
			TestCaseName:        "Commercial Bank",
			ManualIndustry:      "Finance",
			ManualMCC:           []string{"6010", "6011", "6012"},
			ManualSIC:           []string{"6010", "6011", "6012"},
			ManualNAICS:         []string{"522110", "522111", "522112"},
			ManualConfidence:    0.95,
			ClassificationNotes: "Traditional commercial banking services",
			Difficulty:          "Easy",
		},
		"Fintech Startup": {
			TestCaseName:        "Fintech Startup",
			ManualIndustry:      "Fintech",
			ManualMCC:           []string{"6010", "6011", "6012"},
			ManualSIC:           []string{"6010", "6011", "6012"},
			ManualNAICS:         []string{"522110", "522111", "522112"},
			ManualConfidence:    0.85,
			ClassificationNotes: "Financial technology startup with digital services",
			Difficulty:          "Medium",
		},
		"Insurance Company": {
			TestCaseName:        "Insurance Company",
			ManualIndustry:      "Insurance",
			ManualMCC:           []string{"6300", "6301", "6302"},
			ManualSIC:           []string{"6300", "6301", "6302"},
			ManualNAICS:         []string{"524113", "524114", "524115"},
			ManualConfidence:    0.90,
			ClassificationNotes: "Insurance services provider",
			Difficulty:          "Easy",
		},
		"Online Retail Store": {
			TestCaseName:        "Online Retail Store",
			ManualIndustry:      "Retail",
			ManualMCC:           []string{"5310", "5311", "5312"},
			ManualSIC:           []string{"5310", "5311", "5312"},
			ManualNAICS:         []string{"454111", "454112", "454113"},
			ManualConfidence:    0.90,
			ClassificationNotes: "Online retail and e-commerce business",
			Difficulty:          "Easy",
		},
		"E-commerce Platform": {
			TestCaseName:        "E-commerce Platform",
			ManualIndustry:      "Retail",
			ManualMCC:           []string{"5310", "5311", "5312"},
			ManualSIC:           []string{"5310", "5311", "5312"},
			ManualNAICS:         []string{"454111", "454112", "454113"},
			ManualConfidence:    0.85,
			ClassificationNotes: "E-commerce platform and marketplace",
			Difficulty:          "Medium",
		},
		"Industrial Manufacturing": {
			TestCaseName:        "Industrial Manufacturing",
			ManualIndustry:      "Manufacturing",
			ManualMCC:           []string{"5087", "5088", "5089"},
			ManualSIC:           []string{"5087", "5088", "5089"},
			ManualNAICS:         []string{"331110", "331111", "331112"},
			ManualConfidence:    0.90,
			ClassificationNotes: "Industrial manufacturing and production",
			Difficulty:          "Easy",
		},
		"Food Manufacturing": {
			TestCaseName:        "Food Manufacturing",
			ManualIndustry:      "Manufacturing",
			ManualMCC:           []string{"5087", "5088", "5089"},
			ManualSIC:           []string{"5087", "5088", "5089"},
			ManualNAICS:         []string{"311111", "311112", "311113"},
			ManualConfidence:    0.90,
			ClassificationNotes: "Food production and manufacturing",
			Difficulty:          "Easy",
		},
		"Management Consulting": {
			TestCaseName:        "Management Consulting",
			ManualIndustry:      "Professional Services",
			ManualMCC:           []string{"7392", "7393", "7394"},
			ManualSIC:           []string{"7392", "7393", "7394"},
			ManualNAICS:         []string{"541611", "541612", "541613"},
			ManualConfidence:    0.85,
			ClassificationNotes: "Management consulting and advisory services",
			Difficulty:          "Medium",
		},
		"Legal Services": {
			TestCaseName:        "Legal Services",
			ManualIndustry:      "Legal Services",
			ManualMCC:           []string{"7392", "7393", "7394"},
			ManualSIC:           []string{"7392", "7393", "7394"},
			ManualNAICS:         []string{"541110", "541111", "541112"},
			ManualConfidence:    0.95,
			ClassificationNotes: "Legal services and law firm",
			Difficulty:          "Easy",
		},
		"Real Estate Agency": {
			TestCaseName:        "Real Estate Agency",
			ManualIndustry:      "Real Estate",
			ManualMCC:           []string{"6513", "6514", "6515"},
			ManualSIC:           []string{"6513", "6514", "6515"},
			ManualNAICS:         []string{"531210", "531211", "531212"},
			ManualConfidence:    0.90,
			ClassificationNotes: "Real estate sales and property management",
			Difficulty:          "Easy",
		},
		"Educational Technology": {
			TestCaseName:        "Educational Technology",
			ManualIndustry:      "Education",
			ManualMCC:           []string{"7372", "7373", "7374"},
			ManualSIC:           []string{"7372", "7373", "7374"},
			ManualNAICS:         []string{"611710", "611711", "611712"},
			ManualConfidence:    0.85,
			ClassificationNotes: "Educational technology and e-learning platform",
			Difficulty:          "Medium",
		},
		"Renewable Energy": {
			TestCaseName:        "Renewable Energy",
			ManualIndustry:      "Energy",
			ManualMCC:           []string{"4911", "4912", "4913"},
			ManualSIC:           []string{"4911", "4912", "4913"},
			ManualNAICS:         []string{"221115", "221116", "221117"},
			ManualConfidence:    0.90,
			ClassificationNotes: "Renewable energy generation and distribution",
			Difficulty:          "Easy",
		},
		"Mixed Industry Business": {
			TestCaseName:        "Mixed Industry Business",
			ManualIndustry:      "Healthcare",
			ManualMCC:           []string{"8062", "8069", "8071"},
			ManualSIC:           []string{"8062", "8069", "8071"},
			ManualNAICS:         []string{"621111", "621112", "621113"},
			ManualConfidence:    0.70,
			ClassificationNotes: "Mixed business with healthcare focus but technology elements",
			Difficulty:          "Hard",
		},
		"Generic Business": {
			TestCaseName:        "Generic Business",
			ManualIndustry:      "General Business",
			ManualMCC:           []string{"7399", "7398", "7397"},
			ManualSIC:           []string{"7399", "7398", "7397"},
			ManualNAICS:         []string{"541990", "541991", "541992"},
			ManualConfidence:    0.60,
			ClassificationNotes: "Generic business with unclear industry classification",
			Difficulty:          "Hard",
		},
		"Very Short Description": {
			TestCaseName:        "Very Short Description",
			ManualIndustry:      "General Business",
			ManualMCC:           []string{"7399", "7398", "7397"},
			ManualSIC:           []string{"7399", "7398", "7397"},
			ManualNAICS:         []string{"541990", "541991", "541992"},
			ManualConfidence:    0.50,
			ClassificationNotes: "Insufficient information for accurate classification",
			Difficulty:          "Hard",
		},
	}
}

// compareClassificationResults compares automated and manual classification results
func (runner *ClassificationAccuracyTestRunner) compareClassificationResults(
	testCaseName string,
	automatedIndustry string,
	automatedResult *classification.ClassificationCodesInfo,
	manualBaseline ManualClassificationResult,
) ClassificationComparisonResult {
	comparison := ClassificationComparisonResult{
		TestCaseName:      testCaseName,
		AutomatedIndustry: automatedIndustry,
		ManualIndustry:    manualBaseline.ManualIndustry,
	}

	// Compare industry classification
	comparison.IndustryMatch = comparison.AutomatedIndustry == comparison.ManualIndustry
	if comparison.IndustryMatch {
		comparison.IndustryAccuracy = 1.0
	} else {
		comparison.IndustryAccuracy = 0.0
	}

	// Convert automated codes to string slices
	automatedMCC := make([]string, len(automatedResult.MCC))
	for i, code := range automatedResult.MCC {
		automatedMCC[i] = code.Code
	}

	automatedSIC := make([]string, len(automatedResult.SIC))
	for i, code := range automatedResult.SIC {
		automatedSIC[i] = code.Code
	}

	automatedNAICS := make([]string, len(automatedResult.NAICS))
	for i, code := range automatedResult.NAICS {
		automatedNAICS[i] = code.Code
	}

	// Compare MCC codes
	comparison.MCCComparison = runner.compareCodes("MCC", automatedMCC, manualBaseline.ManualMCC)

	// Compare SIC codes
	comparison.SICComparison = runner.compareCodes("SIC", automatedSIC, manualBaseline.ManualSIC)

	// Compare NAICS codes
	comparison.NAICSComparison = runner.compareCodes("NAICS", automatedNAICS, manualBaseline.ManualNAICS)

	// Compare confidence scores
	automatedConfidence := calculateOverallConfidence(automatedResult)
	comparison.ConfidenceComparison = &ConfidenceComparison{
		AutomatedConfidence: automatedConfidence,
		ManualConfidence:    manualBaseline.ManualConfidence,
		ConfidenceDiff:      math.Abs(automatedConfidence - manualBaseline.ManualConfidence),
		ConfidenceMatch:     math.Abs(automatedConfidence-manualBaseline.ManualConfidence) <= 0.1,
		ConfidenceAccuracy:  1.0 - math.Abs(automatedConfidence-manualBaseline.ManualConfidence),
	}

	// Calculate overall accuracy
	comparison.OverallAccuracy = (comparison.IndustryAccuracy + comparison.MCCComparison.Accuracy + comparison.SICComparison.Accuracy + comparison.NAICSComparison.Accuracy + comparison.ConfidenceComparison.ConfidenceAccuracy) / 5.0

	// Determine discrepancy level
	if comparison.OverallAccuracy >= 0.8 {
		comparison.DiscrepancyLevel = "Low"
	} else if comparison.OverallAccuracy >= 0.6 {
		comparison.DiscrepancyLevel = "Medium"
	} else {
		comparison.DiscrepancyLevel = "High"
	}

	// Generate comparison notes
	comparison.ComparisonNotes = runner.generateComparisonNotes(comparison, manualBaseline)

	return comparison
}

// compareCodes compares automated and manual codes
func (runner *ClassificationAccuracyTestRunner) compareCodes(codeType string, automated, manual []string) *CodeComparison {
	comparison := &CodeComparison{
		CodeType:       codeType,
		AutomatedCodes: automated,
		ManualCodes:    manual,
	}

	// Find matched codes
	comparison.MatchedCodes = runner.findMatchedCodes(automated, manual)

	// Find missing codes (in manual but not in automated)
	comparison.MissingCodes = runner.findMissingCodes(automated, manual)

	// Find extra codes (in automated but not in manual)
	comparison.ExtraCodes = runner.findExtraCodes(automated, manual)

	// Calculate precision, recall, and F1 score
	if len(automated) > 0 {
		comparison.Precision = float64(len(comparison.MatchedCodes)) / float64(len(automated))
	} else {
		comparison.Precision = 0.0
	}

	if len(manual) > 0 {
		comparison.Recall = float64(len(comparison.MatchedCodes)) / float64(len(manual))
	} else {
		comparison.Recall = 0.0
	}

	if comparison.Precision+comparison.Recall > 0 {
		comparison.F1Score = 2 * (comparison.Precision * comparison.Recall) / (comparison.Precision + comparison.Recall)
	} else {
		comparison.F1Score = 0.0
	}

	// Calculate overall accuracy
	comparison.Accuracy = (comparison.Precision + comparison.Recall) / 2.0

	return comparison
}

// generateComparisonNotes generates detailed comparison notes
func (runner *ClassificationAccuracyTestRunner) generateComparisonNotes(
	comparison ClassificationComparisonResult,
	manualBaseline ManualClassificationResult,
) string {
	notes := []string{}

	// Industry comparison
	if comparison.IndustryMatch {
		notes = append(notes, "Industry classification matches manual baseline")
	} else {
		notes = append(notes, fmt.Sprintf("Industry mismatch: automated=%s, manual=%s",
			comparison.AutomatedIndustry, comparison.ManualIndustry))
	}

	// Code comparison
	if len(comparison.MCCComparison.MatchedCodes) > 0 {
		notes = append(notes, fmt.Sprintf("MCC codes matched: %v", comparison.MCCComparison.MatchedCodes))
	}
	if len(comparison.MCCComparison.MissingCodes) > 0 {
		notes = append(notes, fmt.Sprintf("MCC codes missing: %v", comparison.MCCComparison.MissingCodes))
	}
	if len(comparison.MCCComparison.ExtraCodes) > 0 {
		notes = append(notes, fmt.Sprintf("MCC codes extra: %v", comparison.MCCComparison.ExtraCodes))
	}

	// Confidence comparison
	if comparison.ConfidenceComparison.ConfidenceMatch {
		notes = append(notes, "Confidence scores match within acceptable range")
	} else {
		notes = append(notes, fmt.Sprintf("Confidence difference: %.2f",
			comparison.ConfidenceComparison.ConfidenceDiff))
	}

	// Overall assessment
	notes = append(notes, fmt.Sprintf("Overall accuracy: %.2f (Discrepancy: %s)",
		comparison.OverallAccuracy, comparison.DiscrepancyLevel))

	return strings.Join(notes, "; ")
}

// DiscrepancyAnalysis represents analysis of classification discrepancies
type DiscrepancyAnalysis struct {
	LowDiscrepancy    int
	MediumDiscrepancy int
	HighDiscrepancy   int
	TotalTests        int
}

// analyzeDiscrepancies analyzes discrepancies across all test cases
func (runner *ClassificationAccuracyTestRunner) analyzeDiscrepancies(
	comparisons []ClassificationComparisonResult,
) DiscrepancyAnalysis {
	analysis := DiscrepancyAnalysis{
		TotalTests: len(comparisons),
	}

	for _, comparison := range comparisons {
		switch comparison.DiscrepancyLevel {
		case "Low":
			analysis.LowDiscrepancy++
		case "Medium":
			analysis.MediumDiscrepancy++
		case "High":
			analysis.HighDiscrepancy++
		}
	}

	return analysis
}

// RunConfidenceScoreReliabilityTest runs comprehensive confidence score reliability tests
func (runner *ClassificationAccuracyTestRunner) RunConfidenceScoreReliabilityTest(t *testing.T) {
	t.Log("🎯 Running Confidence Score Reliability Tests...")

	dataset := NewComprehensiveTestDataset()

	// Test 1: Basic Confidence Score Validation
	basicResult := runner.testBasicConfidenceScoreReliability(t, dataset)

	// Test 2: Confidence Score Distribution Analysis
	distributionResult := runner.testConfidenceScoreDistribution(t, dataset)

	// Test 3: Confidence Score Consistency Testing
	consistencyResult := runner.testConfidenceScoreConsistency(t, dataset)

	// Test 4: Confidence Score Range Validation
	rangeResult := runner.testConfidenceScoreRange(t, dataset)

	// Test 5: Confidence Score Reliability Under Stress
	stressResult := runner.testConfidenceScoreUnderStress(t, dataset)

	// Calculate overall reliability score
	overallReliability := (basicResult.ReliabilityScore +
		distributionResult.ReliabilityScore +
		consistencyResult.ReliabilityScore +
		rangeResult.ReliabilityScore +
		stressResult.ReliabilityScore) / 5.0

	// Log comprehensive results
	t.Logf("📊 Confidence Score Reliability Results:")
	t.Logf("   Basic Validation: %.2f%% reliable", basicResult.ReliabilityScore*100)
	t.Logf("   Distribution Analysis: %.2f%% reliable", distributionResult.ReliabilityScore*100)
	t.Logf("   Consistency Testing: %.2f%% reliable", consistencyResult.ReliabilityScore*100)
	t.Logf("   Range Validation: %.2f%% reliable", rangeResult.ReliabilityScore*100)
	t.Logf("   Stress Testing: %.2f%% reliable", stressResult.ReliabilityScore*100)
	t.Logf("   Overall Reliability: %.2f%%", overallReliability*100)

	// Assert minimum reliability threshold
	if overallReliability < 0.8 {
		t.Errorf("❌ Confidence score reliability too low: %.2f%% (expected >= 80%%)", overallReliability*100)
	}
}

// testBasicConfidenceScoreReliability tests basic confidence score validation
func (runner *ClassificationAccuracyTestRunner) testBasicConfidenceScoreReliability(t *testing.T, dataset *ComprehensiveTestDataset) *ConfidenceScoreReliabilityResult {
	t.Log("🔍 Testing Basic Confidence Score Reliability...")

	result := &ConfidenceScoreReliabilityResult{
		TestName:          "Basic Confidence Score Reliability",
		ScoreRange:        &ScoreRange{},
		DistributionStats: &DistributionStats{},
		ConsistencyStats:  &ConsistencyStats{},
	}

	var allScores []float64
	validScores := 0
	invalidScores := 0

	for _, tc := range dataset.TestCases {
		// Run classification
		classificationResult, err := runner.RunClassification(tc)
		if err != nil {
			t.Errorf("❌ Classification failed for %s: %v", tc.Name, err)
			invalidScores++
			continue
		}

		// Calculate confidence score
		confidence := calculateOverallConfidence(classificationResult)
		allScores = append(allScores, confidence)

		// Validate confidence score
		if confidence >= 0.0 && confidence <= 1.0 {
			validScores++

			// Update distribution stats
			if confidence >= 0.8 {
				result.DistributionStats.HighConfidence++
			} else if confidence >= 0.5 {
				result.DistributionStats.MediumConfidence++
			} else if confidence >= 0.2 {
				result.DistributionStats.LowConfidence++
			} else {
				result.DistributionStats.VeryLowConfidence++
			}
		} else {
			invalidScores++
			t.Errorf("❌ Invalid confidence score for %s: %.2f (expected 0.0-1.0)", tc.Name, confidence)
		}
	}

	// Calculate statistics
	result.TotalTests = len(dataset.TestCases)
	result.ValidScores = validScores
	result.InvalidScores = invalidScores

	if len(allScores) > 0 {
		result.ScoreRange = runner.calculateScoreRange(allScores)
		result.ConsistencyStats = runner.calculateConsistencyStats(allScores)
	}

	// Calculate reliability score
	result.ReliabilityScore = float64(validScores) / float64(result.TotalTests)
	result.IsReliable = result.ReliabilityScore >= 0.95

	t.Logf("✅ Basic Confidence Score Reliability: %.2f%% (Valid: %d/%d)",
		result.ReliabilityScore*100, validScores, result.TotalTests)

	return result
}

// testConfidenceScoreDistribution tests confidence score distribution
func (runner *ClassificationAccuracyTestRunner) testConfidenceScoreDistribution(t *testing.T, dataset *ComprehensiveTestDataset) *ConfidenceScoreReliabilityResult {
	t.Log("📊 Testing Confidence Score Distribution...")

	result := &ConfidenceScoreReliabilityResult{
		TestName:          "Confidence Score Distribution",
		DistributionStats: &DistributionStats{},
	}

	var allScores []float64

	for _, tc := range dataset.TestCases {
		classificationResult, err := runner.RunClassification(tc)
		if err != nil {
			continue
		}

		confidence := calculateOverallConfidence(classificationResult)
		allScores = append(allScores, confidence)
	}

	if len(allScores) > 0 {
		result.DistributionStats = runner.calculateDistributionStats(allScores)
		result.ScoreRange = runner.calculateScoreRange(allScores)
	}

	// Validate distribution - should have reasonable spread
	expectedHighConfidence := float64(len(allScores)) * 0.3   // 30% should be high confidence
	expectedMediumConfidence := float64(len(allScores)) * 0.4 // 40% should be medium confidence
	expectedLowConfidence := float64(len(allScores)) * 0.3    // 30% should be low confidence

	highConfidenceDiff := abs(float64(result.DistributionStats.HighConfidence) - expectedHighConfidence)
	mediumConfidenceDiff := abs(float64(result.DistributionStats.MediumConfidence) - expectedMediumConfidence)
	lowConfidenceDiff := abs(float64(result.DistributionStats.LowConfidence) - expectedLowConfidence)

	// Calculate reliability based on distribution balance
	distributionReliability := 1.0 - (highConfidenceDiff+mediumConfidenceDiff+lowConfidenceDiff)/float64(len(allScores))
	result.ReliabilityScore = max(0.0, distributionReliability)
	result.IsReliable = result.ReliabilityScore >= 0.7

	t.Logf("✅ Confidence Score Distribution: %.2f%% reliable", result.ReliabilityScore*100)
	t.Logf("   High Confidence (≥0.8): %d (%.1f%%)",
		result.DistributionStats.HighConfidence,
		float64(result.DistributionStats.HighConfidence)/float64(len(allScores))*100)
	t.Logf("   Medium Confidence (0.5-0.79): %d (%.1f%%)",
		result.DistributionStats.MediumConfidence,
		float64(result.DistributionStats.MediumConfidence)/float64(len(allScores))*100)
	t.Logf("   Low Confidence (0.2-0.49): %d (%.1f%%)",
		result.DistributionStats.LowConfidence,
		float64(result.DistributionStats.LowConfidence)/float64(len(allScores))*100)

	return result
}

// testConfidenceScoreConsistency tests confidence score consistency
func (runner *ClassificationAccuracyTestRunner) testConfidenceScoreConsistency(t *testing.T, dataset *ComprehensiveTestDataset) *ConfidenceScoreReliabilityResult {
	t.Log("🔄 Testing Confidence Score Consistency...")

	result := &ConfidenceScoreReliabilityResult{
		TestName:         "Confidence Score Consistency",
		ConsistencyStats: &ConsistencyStats{},
	}

	// Test consistency by running the same classification multiple times
	consistencyTests := 0
	consistentResults := 0

	for _, tc := range dataset.TestCases[:5] { // Test first 5 cases for consistency
		var scores []float64

		// Run classification 3 times
		for i := 0; i < 3; i++ {
			classificationResult, err := runner.RunClassification(tc)
			if err != nil {
				continue
			}

			confidence := calculateOverallConfidence(classificationResult)
			scores = append(scores, confidence)
		}

		if len(scores) >= 2 {
			consistencyTests++

			// Check if scores are consistent (within 0.1 of each other)
			maxScore := scores[0]
			minScore := scores[0]
			for _, score := range scores {
				if score > maxScore {
					maxScore = score
				}
				if score < minScore {
					minScore = score
				}
			}

			if maxScore-minScore <= 0.1 {
				consistentResults++
			}
		}
	}

	if consistencyTests > 0 {
		result.ReliabilityScore = float64(consistentResults) / float64(consistencyTests)
		result.IsReliable = result.ReliabilityScore >= 0.8
	}

	t.Logf("✅ Confidence Score Consistency: %.2f%% reliable (%d/%d consistent)",
		result.ReliabilityScore*100, consistentResults, consistencyTests)

	return result
}

// testConfidenceScoreRange tests confidence score range validation
func (runner *ClassificationAccuracyTestRunner) testConfidenceScoreRange(t *testing.T, dataset *ComprehensiveTestDataset) *ConfidenceScoreReliabilityResult {
	t.Log("📏 Testing Confidence Score Range...")

	result := &ConfidenceScoreReliabilityResult{
		TestName:   "Confidence Score Range",
		ScoreRange: &ScoreRange{},
	}

	var allScores []float64

	for _, tc := range dataset.TestCases {
		classificationResult, err := runner.RunClassification(tc)
		if err != nil {
			continue
		}

		confidence := calculateOverallConfidence(classificationResult)
		allScores = append(allScores, confidence)
	}

	if len(allScores) > 0 {
		result.ScoreRange = runner.calculateScoreRange(allScores)

		// Validate range - should be within reasonable bounds
		rangeValid := result.ScoreRange.MinScore >= 0.0 &&
			result.ScoreRange.MaxScore <= 1.0 &&
			result.ScoreRange.AvgScore >= 0.0 &&
			result.ScoreRange.AvgScore <= 1.0

		if rangeValid {
			result.ReliabilityScore = 1.0
		} else {
			result.ReliabilityScore = 0.0
		}

		result.IsReliable = result.ReliabilityScore >= 0.95
	}

	t.Logf("✅ Confidence Score Range: %.2f%% reliable", result.ReliabilityScore*100)
	t.Logf("   Min Score: %.3f", result.ScoreRange.MinScore)
	t.Logf("   Max Score: %.3f", result.ScoreRange.MaxScore)
	t.Logf("   Avg Score: %.3f", result.ScoreRange.AvgScore)
	t.Logf("   Median Score: %.3f", result.ScoreRange.MedianScore)

	return result
}

// testConfidenceScoreUnderStress tests confidence score reliability under stress conditions
func (runner *ClassificationAccuracyTestRunner) testConfidenceScoreUnderStress(t *testing.T, dataset *ComprehensiveTestDataset) *ConfidenceScoreReliabilityResult {
	t.Log("💪 Testing Confidence Score Under Stress...")

	result := &ConfidenceScoreReliabilityResult{
		TestName: "Confidence Score Under Stress",
	}

	// Test with edge cases and difficult scenarios
	stressTests := 0
	reliableTests := 0

	// Test edge cases
	edgeCases := []ClassificationTestCase{
		{
			Name:             "Empty Description",
			BusinessName:     "Test Company",
			Description:      "",
			Keywords:         []string{"test"},
			ExpectedIndustry: "Generic Business",
		},
		{
			Name:             "Very Long Description",
			BusinessName:     "Test Company",
			Description:      strings.Repeat("This is a very long description that goes on and on. ", 50),
			Keywords:         []string{"test", "long", "description"},
			ExpectedIndustry: "Generic Business",
		},
		{
			Name:             "Mixed Languages",
			BusinessName:     "Test Company",
			Description:      "This is a test company that does business in multiple languages. Esta es una empresa de prueba.",
			Keywords:         []string{"test", "multilingual"},
			ExpectedIndustry: "Generic Business",
		},
	}

	for _, tc := range edgeCases {
		classificationResult, err := runner.RunClassification(tc)
		if err != nil {
			continue
		}

		confidence := calculateOverallConfidence(classificationResult)
		stressTests++

		// Check if confidence score is reasonable under stress
		if confidence >= 0.0 && confidence <= 1.0 {
			reliableTests++
		}
	}

	if stressTests > 0 {
		result.ReliabilityScore = float64(reliableTests) / float64(stressTests)
		result.IsReliable = result.ReliabilityScore >= 0.8
	}

	t.Logf("✅ Confidence Score Under Stress: %.2f%% reliable (%d/%d reliable)",
		result.ReliabilityScore*100, reliableTests, stressTests)

	return result
}

// Helper functions for confidence score analysis
func (runner *ClassificationAccuracyTestRunner) calculateScoreRange(scores []float64) *ScoreRange {
	if len(scores) == 0 {
		return &ScoreRange{}
	}

	minScore := scores[0]
	maxScore := scores[0]
	sum := 0.0

	for _, score := range scores {
		if score < minScore {
			minScore = score
		}
		if score > maxScore {
			maxScore = score
		}
		sum += score
	}

	// Calculate median
	sortedScores := make([]float64, len(scores))
	copy(sortedScores, scores)
	sort.Float64s(sortedScores)

	medianScore := 0.0
	if len(sortedScores) > 0 {
		if len(sortedScores)%2 == 0 {
			medianScore = (sortedScores[len(sortedScores)/2-1] + sortedScores[len(sortedScores)/2]) / 2
		} else {
			medianScore = sortedScores[len(sortedScores)/2]
		}
	}

	return &ScoreRange{
		MinScore:    minScore,
		MaxScore:    maxScore,
		AvgScore:    sum / float64(len(scores)),
		MedianScore: medianScore,
	}
}

func (runner *ClassificationAccuracyTestRunner) calculateDistributionStats(scores []float64) *DistributionStats {
	stats := &DistributionStats{}

	for _, score := range scores {
		if score >= 0.8 {
			stats.HighConfidence++
		} else if score >= 0.5 {
			stats.MediumConfidence++
		} else if score >= 0.2 {
			stats.LowConfidence++
		} else {
			stats.VeryLowConfidence++
		}
	}

	return stats
}

func (runner *ClassificationAccuracyTestRunner) calculateConsistencyStats(scores []float64) *ConsistencyStats {
	if len(scores) == 0 {
		return &ConsistencyStats{}
	}

	// Calculate mean
	sum := 0.0
	for _, score := range scores {
		sum += score
	}
	mean := sum / float64(len(scores))

	// Calculate variance
	variance := 0.0
	for _, score := range scores {
		variance += (score - mean) * (score - mean)
	}
	variance /= float64(len(scores))

	// Calculate standard deviation
	stdDev := math.Sqrt(variance)

	// Calculate coefficient of variation
	coefficientOfVariation := 0.0
	if mean != 0 {
		coefficientOfVariation = stdDev / mean
	}

	// Count consistent scores (within 1 standard deviation of mean)
	consistentScores := 0
	for _, score := range scores {
		if math.Abs(score-mean) <= stdDev {
			consistentScores++
		}
	}

	return &ConsistencyStats{
		Variance:               variance,
		StandardDeviation:      stdDev,
		CoefficientOfVariation: coefficientOfVariation,
		ConsistentScores:       consistentScores,
		InconsistentScores:     len(scores) - consistentScores,
	}
}

// Utility functions
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// RunCodeMappingValidationTest runs comprehensive code mapping validation tests
func (runner *ClassificationAccuracyTestRunner) RunCodeMappingValidationTest(t *testing.T) {
	t.Log("🔍 Running Code Mapping Validation Tests...")

	dataset := NewComprehensiveTestDataset()
	totalTests := 0
	passedTests := 0
	totalMappingScore := 0.0

	for _, tc := range dataset.TestCases {
		t.Run("CodeMapping_"+tc.Name, func(t *testing.T) {
			// Run classification
			result, err := runner.RunClassification(tc)
			if err != nil {
				t.Errorf("❌ Classification failed for %s: %v", tc.Name, err)
				return
			}

			// Validate code mapping
			validation := runner.ValidateCodeMapping(tc, result)

			// Check overall mapping score
			if validation.OverallScore < 0.7 {
				t.Errorf("❌ Low mapping score for %s: expected >= 0.7, got %.2f",
					tc.Name, validation.OverallScore)
			}

			// Check individual code type validations
			if !validation.MCCValidation.IsValid {
				t.Errorf("❌ MCC validation failed for %s: accuracy %.2f, format %v, structure %v",
					tc.Name, validation.MCCValidation.AccuracyScore,
					validation.MCCValidation.FormatValid, validation.MCCValidation.StructureValid)
			}

			if !validation.SICValidation.IsValid {
				t.Errorf("❌ SIC validation failed for %s: accuracy %.2f, format %v, structure %v",
					tc.Name, validation.SICValidation.AccuracyScore,
					validation.SICValidation.FormatValid, validation.SICValidation.StructureValid)
			}

			if !validation.NAICSValidation.IsValid {
				t.Errorf("❌ NAICS validation failed for %s: accuracy %.2f, format %v, structure %v",
					tc.Name, validation.NAICSValidation.AccuracyScore,
					validation.NAICSValidation.FormatValid, validation.NAICSValidation.StructureValid)
			}

			// Update statistics
			totalTests++
			totalMappingScore += validation.OverallScore

			if validation.OverallScore >= 0.7 {
				passedTests++
			}

			// Log detailed results
			t.Logf("✅ %s: Overall Score %.2f (MCC: %.2f, SIC: %.2f, NAICS: %.2f)",
				tc.Name, validation.OverallScore,
				validation.MCCValidation.AccuracyScore,
				validation.SICValidation.AccuracyScore,
				validation.NAICSValidation.AccuracyScore)
		})
	}

	// Calculate and log final statistics
	avgMappingScore := totalMappingScore / float64(totalTests)
	passRate := float64(passedTests) / float64(totalTests) * 100

	t.Logf("📊 Code Mapping Validation Results:")
	t.Logf("   Total Tests: %d", totalTests)
	t.Logf("   Passed Tests: %d", passedTests)
	t.Logf("   Pass Rate: %.1f%%", passRate)
	t.Logf("   Average Mapping Score: %.2f", avgMappingScore)

	// Assert minimum pass rate
	if passRate < 70.0 {
		t.Errorf("❌ Code mapping validation pass rate too low: %.1f%% (expected >= 70%%)", passRate)
	}
}

// RunAllTests runs all classification accuracy tests
func (runner *ClassificationAccuracyTestRunner) RunAllTests(t *testing.T) {
	t.Log("🚀 Starting Comprehensive Classification Accuracy Tests")

	// Test 1: Basic Classification Accuracy
	runner.RunBasicAccuracyTest(t)

	// Test 2: Industry-Specific Accuracy
	runner.RunIndustrySpecificTest(t)

	// Test 3: Difficulty-Based Accuracy
	runner.RunDifficultyBasedTest(t)

	// Test 4: Edge Case Handling
	runner.RunEdgeCaseTest(t)

	// Test 5: Performance and Response Time
	runner.RunPerformanceTest(t)

	// Test 6: Confidence Score Validation
	runner.RunConfidenceValidationTest(t)

	// Test 7: Code Mapping Validation
	runner.RunCodeMappingValidationTest(t)

	// Test 8: Code Mapping Accuracy
	runner.RunCodeMappingTest(t)

	// Test 9: Confidence Score Reliability
	runner.RunConfidenceScoreReliabilityTest(t)

	// Test 10: Manual Classification Comparison
	runner.RunManualClassificationComparisonTest(t)

	t.Log("✅ All Classification Accuracy Tests Completed")
}

// RunBasicAccuracyTest tests basic classification accuracy across all test cases
func (runner *ClassificationAccuracyTestRunner) RunBasicAccuracyTest(t *testing.T) {
	t.Run("Basic Classification Accuracy", func(t *testing.T) {
		ctx := context.Background()

		var totalTests int
		var passedTests int
		var totalConfidence float64
		var totalResponseTime time.Duration

		for _, tc := range runner.dataset.TestCases {
			t.Run(tc.Name, func(t *testing.T) {
				// Run classification
				startTime := time.Now()
				result, err := runner.classifier.GenerateClassificationCodes(
					ctx,
					tc.Keywords,
					tc.BusinessName,
					tc.ExpectedConfidence,
				)
				responseTime := time.Since(startTime)

				// Validate results
				if err != nil {
					t.Errorf("❌ Classification failed for %s: %v", tc.Name, err)
					return
				}

				if result == nil {
					t.Errorf("❌ No classification result for %s", tc.Name)
					return
				}

				// Validate classification accuracy
				accuracy := runner.ValidateClassificationAccuracy(tc, result)

				// Check confidence score
				actualConfidence := accuracy.ConfidenceScore
				if actualConfidence < tc.ExpectedConfidence-0.1 {
					t.Errorf("❌ Low confidence for %s: expected >= %.2f, got %.2f",
						tc.Name, tc.ExpectedConfidence-0.1, actualConfidence)
				}

				// Update statistics
				totalTests++
				totalConfidence += actualConfidence
				totalResponseTime += responseTime

				if accuracy.IsAccurate {
					passedTests++
				}

				// Log results
				if accuracy.IsAccurate {
					t.Logf("✅ %s: Accuracy %.1f%%, Confidence %.2f, Response time %v",
						tc.Name, accuracy.AccuracyScore, actualConfidence, responseTime)
				} else {
					t.Logf("⚠️ %s: Accuracy %.1f%% (Expected: %s, Got: %s), Confidence %.2f, Response time %v",
						tc.Name, accuracy.AccuracyScore, tc.ExpectedIndustry, accuracy.MatchedIndustry, actualConfidence, responseTime)
				}
			})
		}

		// Calculate and log overall statistics
		overallAccuracy := float64(passedTests) / float64(totalTests) * 100
		averageConfidence := totalConfidence / float64(totalTests)
		averageResponseTime := totalResponseTime / time.Duration(totalTests)

		t.Logf("📊 Basic Accuracy Test Results:")
		t.Logf("   Total tests: %d", totalTests)
		t.Logf("   Passed tests: %d", passedTests)
		t.Logf("   Overall accuracy: %.1f%%", overallAccuracy)
		t.Logf("   Average confidence: %.2f", averageConfidence)
		t.Logf("   Average response time: %v", averageResponseTime)

		// Assert minimum accuracy threshold
		if overallAccuracy < 70.0 {
			t.Errorf("❌ Overall accuracy %.1f%% is below minimum threshold of 70%%", overallAccuracy)
		}

		// Assert maximum response time
		if averageResponseTime > 2*time.Second {
			t.Errorf("❌ Average response time %v exceeds maximum threshold of 2 seconds", averageResponseTime)
		}
	})
}

// RunIndustrySpecificTest tests accuracy for specific industries
func (runner *ClassificationAccuracyTestRunner) RunIndustrySpecificTest(t *testing.T) {
	t.Run("Industry-Specific Accuracy", func(t *testing.T) {
		ctx := context.Background()

		// Test each industry category
		industries := []string{"Technology", "Healthcare", "Finance", "Retail", "Manufacturing", "Professional Services", "Real Estate", "Education", "Energy"}

		for _, industry := range industries {
			t.Run(industry, func(t *testing.T) {
				testCases := runner.dataset.GetTestCasesByIndustry(industry)
				if len(testCases) == 0 {
					t.Skipf("No test cases found for industry: %s", industry)
					return
				}

				var passedTests int
				var totalConfidence float64

				for _, tc := range testCases {
					result, err := runner.classifier.GenerateClassificationCodes(
						ctx,
						tc.Keywords,
						tc.BusinessName,
						tc.ExpectedConfidence,
					)

					if err != nil {
						t.Errorf("❌ Classification failed for %s: %v", tc.Name, err)
						continue
					}

					if result == nil {
						t.Errorf("❌ No classification result for %s", tc.Name)
						continue
					}

					actualConfidence := calculateOverallConfidence(result)
					totalConfidence += actualConfidence

					if actualConfidence >= tc.ExpectedConfidence-0.1 {
						passedTests++
					}

					t.Logf("   %s: Confidence %.2f", tc.Name, actualConfidence)
				}

				accuracy := float64(passedTests) / float64(len(testCases)) * 100
				averageConfidence := totalConfidence / float64(len(testCases))

				t.Logf("📊 %s Industry Results:", industry)
				t.Logf("   Test cases: %d", len(testCases))
				t.Logf("   Passed: %d", passedTests)
				t.Logf("   Accuracy: %.1f%%", accuracy)
				t.Logf("   Average confidence: %.2f", averageConfidence)

				// Assert minimum accuracy for each industry
				if accuracy < 60.0 {
					t.Errorf("❌ %s industry accuracy %.1f%% is below minimum threshold of 60%%", industry, accuracy)
				}
			})
		}
	})
}

// RunDifficultyBasedTest tests accuracy based on difficulty levels
func (runner *ClassificationAccuracyTestRunner) RunDifficultyBasedTest(t *testing.T) {
	t.Run("Difficulty-Based Accuracy", func(t *testing.T) {
		ctx := context.Background()

		difficulties := []string{"Easy", "Medium", "Hard"}

		for _, difficulty := range difficulties {
			t.Run(difficulty, func(t *testing.T) {
				testCases := runner.dataset.GetTestCasesByDifficulty(difficulty)
				if len(testCases) == 0 {
					t.Skipf("No test cases found for difficulty: %s", difficulty)
					return
				}

				var passedTests int
				var totalConfidence float64

				for _, tc := range testCases {
					result, err := runner.classifier.GenerateClassificationCodes(
						ctx,
						tc.Keywords,
						tc.BusinessName,
						tc.ExpectedConfidence,
					)

					if err != nil {
						t.Errorf("❌ Classification failed for %s: %v", tc.Name, err)
						continue
					}

					if result == nil {
						t.Errorf("❌ No classification result for %s", tc.Name)
						continue
					}

					actualConfidence := calculateOverallConfidence(result)
					totalConfidence += actualConfidence

					if actualConfidence >= tc.ExpectedConfidence-0.1 {
						passedTests++
					}

					t.Logf("   %s: Confidence %.2f", tc.Name, actualConfidence)
				}

				accuracy := float64(passedTests) / float64(len(testCases)) * 100
				averageConfidence := totalConfidence / float64(len(testCases))

				t.Logf("📊 %s Difficulty Results:", difficulty)
				t.Logf("   Test cases: %d", len(testCases))
				t.Logf("   Passed: %d", passedTests)
				t.Logf("   Accuracy: %.1f%%", accuracy)
				t.Logf("   Average confidence: %.2f", averageConfidence)

				// Set different thresholds based on difficulty
				var minThreshold float64
				switch difficulty {
				case "Easy":
					minThreshold = 80.0
				case "Medium":
					minThreshold = 70.0
				case "Hard":
					minThreshold = 50.0
				}

				if accuracy < minThreshold {
					t.Errorf("❌ %s difficulty accuracy %.1f%% is below minimum threshold of %.1f%%",
						difficulty, accuracy, minThreshold)
				}
			})
		}
	})
}

// RunEdgeCaseTest tests edge cases and challenging scenarios
func (runner *ClassificationAccuracyTestRunner) RunEdgeCaseTest(t *testing.T) {
	t.Run("Edge Case Handling", func(t *testing.T) {
		ctx := context.Background()

		edgeCases := runner.dataset.GetTestCasesByCategory("Edge Cases")
		if len(edgeCases) == 0 {
			t.Skip("No edge cases found")
			return
		}

		var passedTests int
		var totalConfidence float64

		for _, tc := range edgeCases {
			t.Run(tc.Name, func(t *testing.T) {
				result, err := runner.classifier.GenerateClassificationCodes(
					ctx,
					tc.Keywords,
					tc.BusinessName,
					tc.ExpectedConfidence,
				)

				if err != nil {
					t.Errorf("❌ Classification failed for %s: %v", tc.Name, err)
					return
				}

				if result == nil {
					t.Errorf("❌ No classification result for %s", tc.Name)
					return
				}

				actualConfidence := calculateOverallConfidence(result)
				totalConfidence += actualConfidence

				// For edge cases, we're more lenient with confidence thresholds
				if actualConfidence >= tc.ExpectedConfidence-0.2 {
					passedTests++
				}

				t.Logf("   %s: Confidence %.2f (Expected: %.2f)", tc.Name, actualConfidence, tc.ExpectedConfidence)
			})
		}

		accuracy := float64(passedTests) / float64(len(edgeCases)) * 100
		averageConfidence := totalConfidence / float64(len(edgeCases))

		t.Logf("📊 Edge Case Results:")
		t.Logf("   Test cases: %d", len(edgeCases))
		t.Logf("   Passed: %d", passedTests)
		t.Logf("   Accuracy: %.1f%%", accuracy)
		t.Logf("   Average confidence: %.2f", averageConfidence)

		// Edge cases have lower accuracy expectations
		if accuracy < 40.0 {
			t.Errorf("❌ Edge case accuracy %.1f%% is below minimum threshold of 40%%", accuracy)
		}
	})
}

// RunPerformanceTest tests performance and response times
func (runner *ClassificationAccuracyTestRunner) RunPerformanceTest(t *testing.T) {
	t.Run("Performance and Response Time", func(t *testing.T) {
		ctx := context.Background()

		// Test with a subset of test cases for performance testing
		testCases := runner.dataset.TestCases[:10] // Use first 10 test cases

		var totalResponseTime time.Duration
		var maxResponseTime time.Duration
		var minResponseTime time.Duration = time.Hour // Initialize with high value

		for _, tc := range testCases {
			startTime := time.Now()
			_, err := runner.classifier.GenerateClassificationCodes(
				ctx,
				tc.Keywords,
				tc.BusinessName,
				tc.ExpectedConfidence,
			)
			responseTime := time.Since(startTime)

			if err != nil {
				t.Errorf("❌ Performance test failed for %s: %v", tc.Name, err)
				continue
			}

			totalResponseTime += responseTime

			if responseTime > maxResponseTime {
				maxResponseTime = responseTime
			}

			if responseTime < minResponseTime {
				minResponseTime = responseTime
			}

			t.Logf("   %s: Response time %v", tc.Name, responseTime)
		}

		averageResponseTime := totalResponseTime / time.Duration(len(testCases))

		t.Logf("📊 Performance Test Results:")
		t.Logf("   Test cases: %d", len(testCases))
		t.Logf("   Average response time: %v", averageResponseTime)
		t.Logf("   Max response time: %v", maxResponseTime)
		t.Logf("   Min response time: %v", minResponseTime)

		// Assert performance thresholds
		if averageResponseTime > 1*time.Second {
			t.Errorf("❌ Average response time %v exceeds maximum threshold of 1 second", averageResponseTime)
		}

		if maxResponseTime > 3*time.Second {
			t.Errorf("❌ Max response time %v exceeds maximum threshold of 3 seconds", maxResponseTime)
		}
	})
}

// RunConfidenceValidationTest tests confidence score validation
func (runner *ClassificationAccuracyTestRunner) RunConfidenceValidationTest(t *testing.T) {
	t.Run("Confidence Score Validation", func(t *testing.T) {
		ctx := context.Background()

		var totalTests int
		var validConfidenceTests int
		var confidenceRangeTests int

		for _, tc := range runner.dataset.TestCases {
			result, err := runner.classifier.GenerateClassificationCodes(
				ctx,
				tc.Keywords,
				tc.BusinessName,
				tc.ExpectedConfidence,
			)

			if err != nil {
				t.Errorf("❌ Confidence validation failed for %s: %v", tc.Name, err)
				continue
			}

			if result == nil {
				t.Errorf("❌ No classification result for %s", tc.Name)
				continue
			}

			totalTests++

			// Check if confidence scores are valid (0.0 to 1.0)
			validConfidence := true
			for _, mcc := range result.MCC {
				if mcc.Confidence < 0.0 || mcc.Confidence > 1.0 {
					validConfidence = false
					break
				}
			}
			for _, sic := range result.SIC {
				if sic.Confidence < 0.0 || sic.Confidence > 1.0 {
					validConfidence = false
					break
				}
			}
			for _, naics := range result.NAICS {
				if naics.Confidence < 0.0 || naics.Confidence > 1.0 {
					validConfidence = false
					break
				}
			}

			if validConfidence {
				validConfidenceTests++
			}

			// Check if confidence scores are within expected range
			overallConfidence := calculateOverallConfidence(result)
			if overallConfidence >= 0.0 && overallConfidence <= 1.0 {
				confidenceRangeTests++
			}

			t.Logf("   %s: Overall confidence %.2f", tc.Name, overallConfidence)
		}

		validConfidenceRate := float64(validConfidenceTests) / float64(totalTests) * 100
		confidenceRangeRate := float64(confidenceRangeTests) / float64(totalTests) * 100

		t.Logf("📊 Confidence Validation Results:")
		t.Logf("   Total tests: %d", totalTests)
		t.Logf("   Valid confidence scores: %d (%.1f%%)", validConfidenceTests, validConfidenceRate)
		t.Logf("   Confidence in range: %d (%.1f%%)", confidenceRangeTests, confidenceRangeRate)

		// Assert confidence validation thresholds
		if validConfidenceRate < 95.0 {
			t.Errorf("❌ Valid confidence rate %.1f%% is below minimum threshold of 95%%", validConfidenceRate)
		}

		if confidenceRangeRate < 100.0 {
			t.Errorf("❌ Confidence range rate %.1f%% is below minimum threshold of 100%%", confidenceRangeRate)
		}
	})
}

// RunCodeMappingTest tests accuracy of industry code mapping
func (runner *ClassificationAccuracyTestRunner) RunCodeMappingTest(t *testing.T) {
	t.Run("Industry Code Mapping Accuracy", func(t *testing.T) {
		ctx := context.Background()

		var totalTests int
		var mccMappingTests int
		var sicMappingTests int
		var naicsMappingTests int

		for _, tc := range runner.dataset.TestCases {
			result, err := runner.classifier.GenerateClassificationCodes(
				ctx,
				tc.Keywords,
				tc.BusinessName,
				tc.ExpectedConfidence,
			)

			if err != nil {
				t.Errorf("❌ Code mapping test failed for %s: %v", tc.Name, err)
				continue
			}

			if result == nil {
				t.Errorf("❌ No classification result for %s", tc.Name)
				continue
			}

			totalTests++

			// Check MCC code mapping
			if len(result.MCC) > 0 {
				mccMappingTests++
			}

			// Check SIC code mapping
			if len(result.SIC) > 0 {
				sicMappingTests++
			}

			// Check NAICS code mapping
			if len(result.NAICS) > 0 {
				naicsMappingTests++
			}

			t.Logf("   %s: MCC=%d, SIC=%d, NAICS=%d",
				tc.Name, len(result.MCC), len(result.SIC), len(result.NAICS))
		}

		mccMappingRate := float64(mccMappingTests) / float64(totalTests) * 100
		sicMappingRate := float64(sicMappingTests) / float64(totalTests) * 100
		naicsMappingRate := float64(naicsMappingTests) / float64(totalTests) * 100

		t.Logf("📊 Code Mapping Results:")
		t.Logf("   Total tests: %d", totalTests)
		t.Logf("   MCC mapping: %d (%.1f%%)", mccMappingTests, mccMappingRate)
		t.Logf("   SIC mapping: %d (%.1f%%)", sicMappingTests, sicMappingRate)
		t.Logf("   NAICS mapping: %d (%.1f%%)", naicsMappingTests, naicsMappingRate)

		// Assert code mapping thresholds
		if mccMappingRate < 80.0 {
			t.Errorf("❌ MCC mapping rate %.1f%% is below minimum threshold of 80%%", mccMappingRate)
		}

		if sicMappingRate < 80.0 {
			t.Errorf("❌ SIC mapping rate %.1f%% is below minimum threshold of 80%%", sicMappingRate)
		}

		if naicsMappingRate < 80.0 {
			t.Errorf("❌ NAICS mapping rate %.1f%% is below minimum threshold of 80%%", naicsMappingRate)
		}
	})
}

// GetClassifier returns the classifier for external use
func (runner *ClassificationAccuracyTestRunner) GetClassifier() *classification.ClassificationCodeGenerator {
	return runner.classifier
}

// GetDataset returns the test dataset for external use
func (runner *ClassificationAccuracyTestRunner) GetDataset() *ComprehensiveTestDataset {
	return runner.dataset
}

// GenerateTestReport generates a comprehensive test report
func (runner *ClassificationAccuracyTestRunner) GenerateTestReport() map[string]interface{} {
	report := make(map[string]interface{})

	// Dataset statistics
	report["dataset_statistics"] = runner.dataset.GetStatistics()

	// Test categories
	report["test_categories"] = []string{
		"Basic Classification Accuracy",
		"Industry-Specific Accuracy",
		"Difficulty-Based Accuracy",
		"Edge Case Handling",
		"Performance and Response Time",
		"Confidence Score Validation",
		"Industry Code Mapping Accuracy",
	}

	// Expected outcomes
	report["expected_outcomes"] = map[string]interface{}{
		"overall_accuracy_threshold":   70.0,
		"industry_accuracy_threshold":  60.0,
		"easy_difficulty_threshold":    80.0,
		"medium_difficulty_threshold":  70.0,
		"hard_difficulty_threshold":    50.0,
		"edge_case_threshold":          40.0,
		"max_response_time":            "2 seconds",
		"max_individual_response_time": "3 seconds",
		"valid_confidence_threshold":   95.0,
		"code_mapping_threshold":       80.0,
	}

	return report
}
