package testing

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"kyb-platform/internal/classification"
)

// ComprehensiveAccuracyTester provides comprehensive accuracy testing for classification system
type ComprehensiveAccuracyTester struct {
	datasetManager    *AccuracyTestDataset
	industryService   *classification.IndustryDetectionService
	codeGenerator     *classification.ClassificationCodeGenerator
	logger            *log.Logger
	concurrencyLimit  int
}

// NewComprehensiveAccuracyTester creates a new comprehensive accuracy tester
func NewComprehensiveAccuracyTester(
	datasetManager *AccuracyTestDataset,
	industryService *classification.IndustryDetectionService,
	codeGenerator *classification.ClassificationCodeGenerator,
	logger *log.Logger,
) *ComprehensiveAccuracyTester {
	if logger == nil {
		logger = log.Default()
	}
	return &ComprehensiveAccuracyTester{
		datasetManager:   datasetManager,
		industryService:  industryService,
		codeGenerator:    codeGenerator,
		logger:           logger,
		concurrencyLimit: 10, // Process 10 test cases concurrently
	}
}

// AccuracyTestResult represents the result of testing a single test case
type AccuracyTestResult struct {
	TestCaseID              int       `json:"test_case_id"`
	BusinessName            string    `json:"business_name"`
	TestCategory            string    `json:"test_category"`
	
	// Expected values
	ExpectedIndustry        string    `json:"expected_industry"`
	ExpectedIndustryConfidence float64 `json:"expected_industry_confidence"`
	ExpectedMCCCodes        []string  `json:"expected_mcc_codes"`
	ExpectedNAICSCodes      []string  `json:"expected_naics_codes"`
	ExpectedSICCodes        []string  `json:"expected_sic_codes"`
	
	// Actual values
	ActualIndustry          string    `json:"actual_industry"`
	ActualIndustryConfidence float64  `json:"actual_industry_confidence"`
	ActualMCCCodes          []string  `json:"actual_mcc_codes"`
	ActualNAICSCodes        []string  `json:"actual_naics_codes"`
	ActualSICCodes          []string  `json:"actual_sic_codes"`
	
	// Accuracy metrics
	IndustryMatch            bool      `json:"industry_match"`
	IndustryAccuracy         float64   `json:"industry_accuracy"` // 1.0 if exact match, 0.0 if no match
	MCCAccuracy              float64   `json:"mcc_accuracy"`     // Percentage of expected MCC codes found in top 3
	NAICSAccuracy            float64   `json:"naics_accuracy"`   // Percentage of expected NAICS codes found in top 3
	SICAccuracy              float64   `json:"sic_accuracy"`     // Percentage of expected SIC codes found in top 3
	OverallAccuracy          float64   `json:"overall_accuracy"` // Weighted average of all metrics
	
	// Metadata
	ProcessingTime           time.Duration `json:"processing_time"`
	Error                    string        `json:"error,omitempty"`
	Timestamp                time.Time     `json:"timestamp"`
}

// ComprehensiveAccuracyMetrics represents aggregated accuracy metrics for comprehensive testing
type ComprehensiveAccuracyMetrics struct {
	TotalTestCases           int     `json:"total_test_cases"`
	PassedTestCases          int     `json:"passed_test_cases"`
	FailedTestCases          int     `json:"failed_test_cases"`
	
	// Industry accuracy
	IndustryAccuracy         float64 `json:"industry_accuracy"`         // Percentage of correct industry classifications
	IndustryExactMatches     int     `json:"industry_exact_matches"`    // Number of exact industry matches
	IndustryPartialMatches   int     `json:"industry_partial_matches"` // Number of partial matches (if applicable)
	
	// Code accuracy
	MCCAccuracy              float64 `json:"mcc_accuracy"`              // Average MCC code accuracy
	NAICSAccuracy            float64 `json:"naics_accuracy"`           // Average NAICS code accuracy
	SICAccuracy              float64 `json:"sic_accuracy"`              // Average SIC code accuracy
	CodeAccuracy             float64 `json:"code_accuracy"`             // Overall code accuracy (average of MCC, NAICS, SIC)
	
	// Overall accuracy
	OverallAccuracy          float64 `json:"overall_accuracy"`          // Weighted overall accuracy
	
	// Accuracy by category
	AccuracyByCategory       map[string]float64 `json:"accuracy_by_category"`
	
	// Accuracy by industry
	AccuracyByIndustry       map[string]float64 `json:"accuracy_by_industry"`
	
	// Edge case performance
	EdgeCaseAccuracy         float64 `json:"edge_case_accuracy"`
	HighConfidenceAccuracy   float64 `json:"high_confidence_accuracy"`
	
	// Performance metrics
	AverageProcessingTime    time.Duration `json:"average_processing_time"`
	TotalProcessingTime      time.Duration `json:"total_processing_time"`
	
	// Test results
	TestResults              []*AccuracyTestResult `json:"test_results"`
	
	// Timestamp
	GeneratedAt              time.Time `json:"generated_at"`
}

// RunAccuracyTests runs comprehensive accuracy tests on all test cases
func (cat *ComprehensiveAccuracyTester) RunAccuracyTests(ctx context.Context) (*ComprehensiveAccuracyMetrics, error) {
	cat.logger.Println("🚀 Starting comprehensive accuracy tests...")
	startTime := time.Now()

	// Load all test cases
	testCases, err := cat.datasetManager.LoadAllTestCases(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load test cases: %w", err)
	}

	if len(testCases) == 0 {
		return nil, fmt.Errorf("no test cases found in dataset")
	}

	cat.logger.Printf("📊 Loaded %d test cases", len(testCases))

	// Run tests with concurrency control
	testResults := cat.runTestsConcurrently(ctx, testCases)

	// Calculate metrics
	metrics := cat.calculateMetrics(testResults, time.Since(startTime))

	cat.logger.Printf("✅ Accuracy tests completed: Overall Accuracy: %.2f%%, Industry: %.2f%%, Codes: %.2f%%",
		metrics.OverallAccuracy*100, metrics.IndustryAccuracy*100, metrics.CodeAccuracy*100)

	return metrics, nil
}

// RunAccuracyTestsByCategory runs accuracy tests for a specific category
func (cat *ComprehensiveAccuracyTester) RunAccuracyTestsByCategory(ctx context.Context, category string) (*ComprehensiveAccuracyMetrics, error) {
	cat.logger.Printf("🚀 Starting accuracy tests for category: %s", category)

	// Load test cases by category
	testCases, err := cat.datasetManager.LoadTestCasesByCategory(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("failed to load test cases by category: %w", err)
	}

	if len(testCases) == 0 {
		return nil, fmt.Errorf("no test cases found for category: %s", category)
	}

	cat.logger.Printf("📊 Loaded %d test cases for category: %s", len(testCases), category)

	// Run tests
	testResults := cat.runTestsConcurrently(ctx, testCases)

	// Calculate metrics
	metrics := cat.calculateMetrics(testResults, time.Duration(0))

	cat.logger.Printf("✅ Category accuracy tests completed: Overall Accuracy: %.2f%%", metrics.OverallAccuracy*100)

	return metrics, nil
}

// runTestsConcurrently runs tests with controlled concurrency
func (cat *ComprehensiveAccuracyTester) runTestsConcurrently(ctx context.Context, testCases []*TestCase) []*AccuracyTestResult {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, cat.concurrencyLimit)
	results := make([]*AccuracyTestResult, len(testCases))
	var mu sync.Mutex

	for i, tc := range testCases {
		wg.Add(1)
		go func(index int, testCase *TestCase) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Run test
			result := cat.runSingleTest(ctx, testCase)
			
			// Store result
			mu.Lock()
			results[index] = result
			mu.Unlock()
		}(i, tc)
	}

	wg.Wait()
	return results
}

// runSingleTest runs a single test case
func (cat *ComprehensiveAccuracyTester) runSingleTest(ctx context.Context, testCase *TestCase) *AccuracyTestResult {
	startTime := time.Now()
	result := &AccuracyTestResult{
		TestCaseID:              testCase.ID,
		BusinessName:            testCase.BusinessName,
		TestCategory:            testCase.TestCategory,
		ExpectedIndustry:        testCase.ExpectedPrimaryIndustry,
		ExpectedIndustryConfidence: testCase.ExpectedIndustryConfidence,
		ExpectedMCCCodes:        testCase.ExpectedMCCCodes,
		ExpectedNAICSCodes:      testCase.ExpectedNAICSCodes,
		ExpectedSICCodes:        testCase.ExpectedSICCodes,
		Timestamp:               time.Now(),
	}

	// Run industry detection
	industryResult, err := cat.industryService.DetectIndustry(
		ctx,
		testCase.BusinessName,
		testCase.BusinessDescription,
		testCase.WebsiteURL,
	)
	if err != nil {
		result.Error = fmt.Sprintf("industry detection failed: %v", err)
		result.ProcessingTime = time.Since(startTime)
		return result
	}

	result.ActualIndustry = industryResult.IndustryName
	result.ActualIndustryConfidence = industryResult.Confidence

	// Normalize industry names for comparison using IndustryNameNormalizer
	normalizer := classification.NewIndustryNameNormalizer()
	normalizedExpected, _ := normalizer.NormalizeIndustryName(result.ExpectedIndustry)
	normalizedActual, _ := normalizer.NormalizeIndustryName(result.ActualIndustry)

	// Check industry match (using normalized names for better accuracy)
	result.IndustryMatch = strings.EqualFold(normalizedActual, normalizedExpected)
	if result.IndustryMatch {
		result.IndustryAccuracy = 1.0
	} else {
		result.IndustryAccuracy = 0.0
	}

	// Generate classification codes
	keywords := industryResult.Keywords
	if len(keywords) == 0 {
		// Extract basic keywords from business name and description
		keywords = extractBasicKeywords(testCase.BusinessName, testCase.BusinessDescription)
	}

	codeInfo, err := cat.codeGenerator.GenerateClassificationCodes(
		ctx,
		keywords,
		industryResult.IndustryName,
		industryResult.Confidence,
	)
	if err != nil {
		result.Error = fmt.Sprintf("code generation failed: %v", err)
		result.ProcessingTime = time.Since(startTime)
		return result
	}

	// Extract actual codes (top 3 per type)
	result.ActualMCCCodes = extractTopCodes(codeInfo.MCC, 3)
	result.ActualNAICSCodes = extractTopCodes(codeInfo.NAICS, 3)
	result.ActualSICCodes = extractTopCodes(codeInfo.SIC, 3)

	// Calculate code accuracy
	result.MCCAccuracy = calculateCodeAccuracy(result.ExpectedMCCCodes, result.ActualMCCCodes)
	result.NAICSAccuracy = calculateCodeAccuracy(result.ExpectedNAICSCodes, result.ActualNAICSCodes)
	result.SICAccuracy = calculateCodeAccuracy(result.ExpectedSICCodes, result.ActualSICCodes)

	// Calculate overall accuracy (weighted: 40% industry, 60% codes)
	codeAccuracy := (result.MCCAccuracy + result.NAICSAccuracy + result.SICAccuracy) / 3.0
	result.OverallAccuracy = 0.4*result.IndustryAccuracy + 0.6*codeAccuracy

	result.ProcessingTime = time.Since(startTime)
	return result
}

// calculateMetrics calculates aggregated accuracy metrics
func (cat *ComprehensiveAccuracyTester) calculateMetrics(results []*AccuracyTestResult, totalTime time.Duration) *ComprehensiveAccuracyMetrics {
	metrics := &ComprehensiveAccuracyMetrics{
		TotalTestCases:        len(results),
		AccuracyByCategory:    make(map[string]float64),
		AccuracyByIndustry:    make(map[string]float64),
		TestResults:           results,
		GeneratedAt:           time.Now(),
		TotalProcessingTime:   totalTime,
	}

	if len(results) == 0 {
		return metrics
	}

	// Category and industry tracking
	categoryCounts := make(map[string]int)
	categoryAccuracies := make(map[string]float64)
	industryCounts := make(map[string]int)
	industryAccuracies := make(map[string]float64)

	var totalIndustryAccuracy float64
	var totalMCCAccuracy float64
	var totalNAICSAccuracy float64
	var totalSICAccuracy float64
	var totalOverallAccuracy float64
	var totalProcessingTime time.Duration
	var edgeCaseAccuracySum float64
	var edgeCaseCount int
	var highConfidenceAccuracySum float64
	var highConfidenceCount int

	for _, result := range results {
		if result.Error != "" {
			metrics.FailedTestCases++
			continue
		}

		metrics.PassedTestCases++

		// Industry accuracy
		totalIndustryAccuracy += result.IndustryAccuracy
		if result.IndustryMatch {
			metrics.IndustryExactMatches++
		}

		// Code accuracy
		totalMCCAccuracy += result.MCCAccuracy
		totalNAICSAccuracy += result.NAICSAccuracy
		totalSICAccuracy += result.SICAccuracy

		// Overall accuracy
		totalOverallAccuracy += result.OverallAccuracy

		// Processing time
		totalProcessingTime += result.ProcessingTime

		// Category tracking
		categoryCounts[result.TestCategory]++
		categoryAccuracies[result.TestCategory] += result.OverallAccuracy

		// Industry tracking
		industryCounts[result.ExpectedIndustry]++
		industryAccuracies[result.ExpectedIndustry] += result.OverallAccuracy

		// Edge case and high confidence tracking (would need to check test case metadata)
		// For now, we'll skip this as we don't have direct access to test case in result
	}

	// Calculate averages
	if metrics.PassedTestCases > 0 {
		metrics.IndustryAccuracy = totalIndustryAccuracy / float64(metrics.PassedTestCases)
		metrics.MCCAccuracy = totalMCCAccuracy / float64(metrics.PassedTestCases)
		metrics.NAICSAccuracy = totalNAICSAccuracy / float64(metrics.PassedTestCases)
		metrics.SICAccuracy = totalSICAccuracy / float64(metrics.PassedTestCases)
		metrics.CodeAccuracy = (metrics.MCCAccuracy + metrics.NAICSAccuracy + metrics.SICAccuracy) / 3.0
		metrics.OverallAccuracy = totalOverallAccuracy / float64(metrics.PassedTestCases)
		metrics.AverageProcessingTime = totalProcessingTime / time.Duration(metrics.PassedTestCases)

		if edgeCaseCount > 0 {
			metrics.EdgeCaseAccuracy = edgeCaseAccuracySum / float64(edgeCaseCount)
		}
		if highConfidenceCount > 0 {
			metrics.HighConfidenceAccuracy = highConfidenceAccuracySum / float64(highConfidenceCount)
		}
	}

	metrics.FailedTestCases = metrics.TotalTestCases - metrics.PassedTestCases

	// Calculate accuracy by category
	for category, count := range categoryCounts {
		if count > 0 {
			metrics.AccuracyByCategory[category] = categoryAccuracies[category] / float64(count)
		}
	}

	// Calculate accuracy by industry
	for industry, count := range industryCounts {
		if count > 0 {
			metrics.AccuracyByIndustry[industry] = industryAccuracies[industry] / float64(count)
		}
	}

	return metrics
}

// Helper functions

// extractTopCodes extracts top N codes from classification code results
func extractTopCodes(codes interface{}, topN int) []string {
	result := []string{}
	
	// Handle different code types
	switch v := codes.(type) {
	case []classification.MCCCode:
		for i, code := range v {
			if i >= topN {
				break
			}
			result = append(result, code.Code)
		}
	case []classification.SICCode:
		for i, code := range v {
			if i >= topN {
				break
			}
			result = append(result, code.Code)
		}
	case []classification.NAICSCode:
		for i, code := range v {
			if i >= topN {
				break
			}
			result = append(result, code.Code)
		}
	}
	
	return result
}

// calculateCodeAccuracy calculates the accuracy of code matching
// Returns the percentage of expected codes found in actual codes
func calculateCodeAccuracy(expected, actual []string) float64 {
	if len(expected) == 0 {
		return 1.0 // If no expected codes, consider it a pass
	}

	if len(actual) == 0 {
		return 0.0 // If no actual codes but expected codes exist, it's a fail
	}

	// Create a map of actual codes for quick lookup
	actualMap := make(map[string]bool)
	for _, code := range actual {
		actualMap[code] = true
	}

	// Count how many expected codes are found
	matches := 0
	for _, code := range expected {
		if actualMap[code] {
			matches++
		}
	}

	// Return percentage of expected codes found
	return float64(matches) / float64(len(expected))
}

// extractBasicKeywords extracts basic keywords from business name and description
func extractBasicKeywords(businessName, description string) []string {
	keywords := []string{}
	
	// Simple keyword extraction (can be enhanced)
	text := strings.ToLower(businessName + " " + description)
	words := strings.Fields(text)
	
	// Filter out common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
		"is": true, "was": true, "are": true, "were": true, "be": true,
		"been": true, "being": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true,
		"should": true, "could": true, "may": true, "might": true, "must": true,
		"can": true, "this": true, "that": true, "these": true, "those": true,
		"it": true, "its": true, "they": true, "them": true, "their": true,
		"our": true, "your": true, "my": true, "his": true, "her": true,
		"he": true, "she": true, "we": true, "you": true, "i": true, "me": true, "us": true,
	}
	
	for _, word := range words {
		word = strings.Trim(word, ".,!?;:()[]{}'\"")
		if len(word) >= 3 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}
	
	return keywords
}

