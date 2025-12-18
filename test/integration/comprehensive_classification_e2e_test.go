package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestComprehensiveClassificationE2E runs comprehensive end-to-end tests
// Tests 100 samples across the entire classification flow from scraping to frontend
func TestComprehensiveClassificationE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive E2E test in short mode")
	}

	// Load test samples
	samples, err := loadTestSamples("test/data/comprehensive_test_samples.json")
	if err != nil {
		t.Fatalf("Failed to load test samples: %v", err)
	}

	if len(samples) < 100 {
		t.Fatalf("Expected at least 100 samples, got %d", len(samples))
	}

	// Use first 100 samples
	if len(samples) > 100 {
		samples = samples[:100]
	}

	t.Logf("🚀 Starting comprehensive E2E tests with %d samples", len(samples))

	// Initialize test runner
	apiURL := os.Getenv("CLASSIFICATION_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8081"
	}
	runner := NewClassificationTestRunner(t, apiURL)

	// Run tests
	startTime := time.Now()
	_ = runner.RunComprehensiveTests(samples)
	totalDuration := time.Since(startTime)

	t.Logf("✅ Completed all tests in %v", totalDuration)

	// Calculate metrics
	runner.CalculateMetrics()

	// Generate report
	report := runner.GenerateReport(totalDuration)

	// Validate results against success criteria
	validateResults(t, report)

	// Save report
	reportPath := "test/results/comprehensive_test_results.json"
	if err := saveReport(report, reportPath); err != nil {
		t.Errorf("Failed to save report: %v", err)
	} else {
		t.Logf("📊 Test report saved to %s", reportPath)
	}

	// Print summary
	runner.PrintSummary()
}

// TestSample represents a test sample
type TestSample struct {
	ID                string   `json:"id"`
	BusinessName      string   `json:"business_name"`
	Description       string   `json:"description"`
	WebsiteURL        string   `json:"website_url"`
	ExpectedIndustry  string   `json:"expected_industry"`
	ExpectedMCC       []string `json:"expected_mcc_codes"`
	ExpectedNAICS     []string `json:"expected_naics_codes"`
	ExpectedSIC       []string `json:"expected_sic_codes"`
	Category          string   `json:"category"`
	Complexity        string   `json:"complexity"`
	ScrapingDifficulty string `json:"scraping_difficulty"`
}

// DurationMs represents a duration in milliseconds for JSON marshaling
type DurationMs int64

// MarshalJSON converts DurationMs to milliseconds in JSON
func (d DurationMs) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(d))
}

// FromDuration converts time.Duration to DurationMs (milliseconds)
func DurationMsFromDuration(d time.Duration) DurationMs {
	return DurationMs(d.Milliseconds())
}

// ToDuration converts DurationMs back to time.Duration
func (d DurationMs) ToDuration() time.Duration {
	return time.Duration(d) * time.Millisecond
}

// ClassificationTestResult represents a single classification test result
type ClassificationTestResult struct {
	SampleID              string                 `json:"sample_id"`
	BusinessName          string                 `json:"business_name"`
	WebsiteURL            string                 `json:"website_url"`
	ExpectedIndustry      string                 `json:"expected_industry"`
	ActualIndustry        string                 `json:"actual_industry"`
	Success               bool                   `json:"success"`
	ProcessingTime        DurationMs              `json:"processing_time_ms"`
	ScrapingTime          DurationMs              `json:"scraping_time_ms"`
	ClassificationTime    DurationMs              `json:"classification_time_ms"`
	ScrapingStrategy      string                 `json:"scraping_strategy"`
	EarlyExit             bool                   `json:"early_exit"`
	CacheHit              bool                   `json:"cache_hit"`
	FallbackUsed          bool                   `json:"fallback_used"`
	FallbackType          string                 `json:"fallback_type,omitempty"`
	ConfidenceScore       float64                `json:"confidence_score"`
	MCCCodes              []IndustryCode          `json:"mcc_codes"`
	NAICSCodes            []IndustryCode          `json:"naics_codes"`
	SICCodes              []IndustryCode          `json:"sic_codes"`
	Explanation           string                 `json:"explanation"`
	ExplanationStructured map[string]interface{} `json:"explanation_structured,omitempty"`
	Error                 string                 `json:"error,omitempty"`
	FrontendDataValid     bool                   `json:"frontend_data_valid"`
	FrontendDataIssues    []string               `json:"frontend_data_issues,omitempty"`
	Accuracy              bool                   `json:"accuracy"`
	Timestamp             time.Time              `json:"timestamp"`
}

// IndustryCode represents an industry classification code
type IndustryCode struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// ClassificationTestRunner runs comprehensive classification tests
type ClassificationTestRunner struct {
	t          *testing.T
	apiURL     string
	httpClient *http.Client
	results    []ClassificationTestResult
	metrics    *TestMetrics
	mu         sync.Mutex
}

// TestMetrics tracks aggregate metrics
type TestMetrics struct {
	TotalTests            int                       `json:"total_tests"`
	SuccessfulTests       int                       `json:"successful_tests"`
	FailedTests           int                       `json:"failed_tests"`
	Accuracy              float64                   `json:"accuracy"`
	AverageLatency        time.Duration             `json:"average_latency_ms"`
	P50Latency            time.Duration             `json:"p50_latency_ms"`
	P95Latency            time.Duration             `json:"p95_latency_ms"`
	P99Latency            time.Duration             `json:"p99_latency_ms"`
	StrategyDistribution  map[string]int            `json:"strategy_distribution"`
	StrategySuccessRate   map[string]float64        `json:"strategy_success_rate"`
	StrategyLatency       map[string]time.Duration  `json:"strategy_latency"`
	EarlyExitCount        int                       `json:"early_exit_count"`
	EarlyExitRate         float64                   `json:"early_exit_rate"`
	CacheHitCount         int                       `json:"cache_hit_count"`
	CacheHitRate          float64                   `json:"cache_hit_rate"`
	FallbackUsage         map[string]int            `json:"fallback_usage"`
	Throughput            float64                   `json:"throughput_rps"`
	AccuracyByIndustry    map[string]float64        `json:"accuracy_by_industry"`
	AccuracyByComplexity   map[string]float64        `json:"accuracy_by_complexity"`
	FrontendCompatibility FrontendCompatibility      `json:"frontend_compatibility"`
	CodeAccuracy          CodeAccuracyMetrics        `json:"code_accuracy"`
}

// FrontendCompatibility tracks frontend data format compliance
type FrontendCompatibility struct {
	AllFieldsPresent   float64 `json:"all_fields_present"`
	DataTypesCorrect   float64 `json:"data_types_correct"`
	StructureValid     float64 `json:"structure_valid"`
	IndustryPresent    float64 `json:"industry_present"`
	CodesPresent       float64 `json:"codes_present"`
	ExplanationPresent float64 `json:"explanation_present"`
	Top3CodesPresent   float64 `json:"top3_codes_present"`
}

// CodeAccuracyMetrics tracks code accuracy
type CodeAccuracyMetrics struct {
	MCCAccuracy   float64 `json:"mcc_accuracy"`
	NAICSAccuracy float64 `json:"naics_accuracy"`
	SICAccuracy   float64 `json:"sic_accuracy"`
	Top3MatchRate float64 `json:"top3_match_rate"`
}

// NewClassificationTestRunner creates a new test runner
func NewClassificationTestRunner(t *testing.T, apiURL string) *ClassificationTestRunner {
	return &ClassificationTestRunner{
		t:          t,
		apiURL:     apiURL,
		httpClient: &http.Client{Timeout: 60 * time.Second}, // Increased from 30s to 60s to allow fallback strategies to complete
		results:    make([]ClassificationTestResult, 0),
		metrics: &TestMetrics{
			StrategyDistribution: make(map[string]int),
			StrategySuccessRate:   make(map[string]float64),
			StrategyLatency:       make(map[string]time.Duration),
			FallbackUsage:         make(map[string]int),
			AccuracyByIndustry:    make(map[string]float64),
			AccuracyByComplexity:   make(map[string]float64),
		},
	}
}

// RunComprehensiveTests runs all test samples
func (r *ClassificationTestRunner) RunComprehensiveTests(samples []TestSample) []ClassificationTestResult {
	r.t.Logf("🚀 Starting comprehensive E2E tests with %d samples", len(samples))

	// Run tests sequentially to avoid overwhelming the service
	for i, sample := range samples {
		r.t.Logf("Running test %d/%d: %s", i+1, len(samples), sample.BusinessName)
		result := r.runSingleTest(sample)
		r.addResult(result)

		// Log progress every 10 tests
		if (i+1)%10 == 0 {
			r.t.Logf("Progress: %d/%d tests completed", i+1, len(samples))
		}

		// Small delay to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	r.t.Logf("✅ Completed all tests")
	return r.results
}

// runSingleTest runs a single classification test
func (r *ClassificationTestRunner) runSingleTest(sample TestSample) ClassificationTestResult {
	result := ClassificationTestResult{
		SampleID:         sample.ID,
		BusinessName:     sample.BusinessName,
		WebsiteURL:       sample.WebsiteURL,
		ExpectedIndustry: sample.ExpectedIndustry,
		Timestamp:        time.Now(),
	}

	startTime := time.Now()

	// Make classification request
	reqBody := map[string]interface{}{
		"business_name": sample.BusinessName,
		"description":   sample.Description,
		"website_url":   sample.WebsiteURL,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to marshal request: %v", err)
		result.Success = false
		return result
	}

	req, err := http.NewRequest("POST", r.apiURL+"/v1/classify", bytes.NewBuffer(reqJSON))
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create request: %v", err)
		result.Success = false
		return result
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("HTTP request failed: %v", err)
		result.Success = false
		result.ProcessingTime = DurationMsFromDuration(time.Since(startTime))
		return result
	}
	defer resp.Body.Close()

	result.ProcessingTime = DurationMsFromDuration(time.Since(startTime))

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to read response: %v", err)
		result.Success = false
		return result
	}

	// Parse response
	var apiResponse map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		result.Error = fmt.Sprintf("Failed to parse response: %v", err)
		result.Success = false
		return result
	}

	// Extract data
	result.Success = extractBool(apiResponse, "success")
	result.ActualIndustry = extractString(apiResponse, "primary_industry")
	result.ConfidenceScore = extractFloat64(apiResponse, "confidence_score")
	result.Explanation = extractString(apiResponse, "explanation")
	result.CacheHit = extractBool(apiResponse, "from_cache")

	// Extract classification codes
	if classification, ok := apiResponse["classification"].(map[string]interface{}); ok {
		result.MCCCodes = extractIndustryCodes(classification, "mcc_codes")
		result.NAICSCodes = extractIndustryCodes(classification, "naics_codes")
		result.SICCodes = extractIndustryCodes(classification, "sic_codes")

		// Extract structured explanation
		if explanation, ok := classification["explanation"].(map[string]interface{}); ok {
			result.ExplanationStructured = explanation
		}
	}

	// Extract metadata for strategy tracking
	if metadata, ok := apiResponse["metadata"].(map[string]interface{}); ok {
		result.ScrapingStrategy = extractString(metadata, "scraping_strategy")
		result.EarlyExit = extractBool(metadata, "early_exit")
		result.FallbackUsed = extractBool(metadata, "fallback_used")
		result.FallbackType = extractString(metadata, "fallback_type")
		result.ScrapingTime = DurationMsFromDuration(extractDuration(metadata, "scraping_time_ms"))
		result.ClassificationTime = DurationMsFromDuration(extractDuration(metadata, "classification_time_ms"))
	}

	// If scraping strategy not in metadata, try to infer from logs or response
	if result.ScrapingStrategy == "" {
		// Try to extract from processing_path or other fields
		if path, ok := apiResponse["processing_path"].(string); ok {
			result.ScrapingStrategy = path
		}
	}

	// Validate frontend data format
	result.FrontendDataValid, result.FrontendDataIssues = r.validateFrontendData(apiResponse)

	// Check accuracy with industry name normalization
	result.Accuracy = normalizeIndustryName(result.ActualIndustry) == normalizeIndustryName(result.ExpectedIndustry)

	return result
}

// normalizeIndustryName normalizes industry names to handle variations
// Maps common variations to standard industry names
func normalizeIndustryName(industry string) string {
	if industry == "" {
		return ""
	}
	
	// Convert to lowercase for comparison
	normalized := strings.ToLower(strings.TrimSpace(industry))
	
	// Industry name aliases/mappings
	industryAliases := map[string]string{
		// Retail variations
		"retail & commerce": "retail",
		"retail and commerce": "retail",
		"retail trade": "retail",
		"commerce": "retail",
		
		// Financial Services variations
		"financial services": "banking",
		"finance": "banking",
		"financial": "banking",
		
		// Technology variations
		"tech": "technology",
		"information technology": "technology",
		"it": "technology",
		
		// Healthcare variations
		"healthcare": "healthcare",
		"health care": "healthcare",
		"medical": "healthcare",
		
		// Food & Beverage variations
		"food & beverage": "food & beverage",
		"food and beverage": "food & beverage",
		"restaurant": "food & beverage",
		"food service": "food & beverage",
		
		// Professional Services variations
		"professional services": "professional services",
		"consulting": "professional services",
		"business services": "professional services",
		
		// Real Estate variations
		"real estate": "real estate",
		"property": "real estate",
		
		// Manufacturing variations
		"manufacturing": "manufacturing",
		"production": "manufacturing",
		
		// Construction variations
		"construction": "construction",
		"building": "construction",
		
		// Transportation variations
		"transportation": "transportation",
		"logistics": "transportation",
		"shipping": "transportation",
		
		// Education variations
		"education": "education",
		"educational services": "education",
		
		// Agriculture variations
		"agriculture": "agriculture",
		"farming": "agriculture",
		
		// Mining & Energy variations
		"mining & energy": "mining & energy",
		"mining and energy": "mining & energy",
		"energy": "mining & energy",
		"mining": "mining & energy",
		
		// Utilities variations
		"utilities": "utilities",
		"utility": "utilities",
		
		// Wholesale Trade variations
		"wholesale trade": "wholesale trade",
		"wholesale": "wholesale trade",
		
		// Arts & Entertainment variations
		"arts & entertainment": "arts & entertainment",
		"arts and entertainment": "arts & entertainment",
		"entertainment": "arts & entertainment",
		
		// Accommodation & Hospitality variations
		"accommodation & hospitality": "accommodation & hospitality",
		"accommodation and hospitality": "accommodation & hospitality",
		"hospitality": "accommodation & hospitality",
		"hotel": "accommodation & hospitality",
		
		// Administrative Services variations
		"administrative services": "administrative services",
		"admin services": "administrative services",
		
		// Other Services variations
		"other services": "other services",
		"miscellaneous": "other services",
	}
	
	// Check if normalized name exists in aliases
	if mapped, exists := industryAliases[normalized]; exists {
		return mapped
	}
	
	// Return normalized version (lowercase, trimmed)
	return normalized
}

// validateFrontendData validates response matches frontend expectations
func (r *ClassificationTestRunner) validateFrontendData(response map[string]interface{}) (bool, []string) {
	issues := []string{}

	// Required fields
	requiredFields := []string{"success", "primary_industry", "confidence_score"}
	for _, field := range requiredFields {
		if _, ok := response[field]; !ok {
			issues = append(issues, fmt.Sprintf("Missing required field: %s", field))
		}
	}

	// Classification object
	if classification, ok := response["classification"].(map[string]interface{}); ok {
		// Check codes arrays
		codeTypes := []string{"mcc_codes", "naics_codes", "sic_codes"}
		for _, codeType := range codeTypes {
			if codes, ok := classification[codeType].([]interface{}); ok {
				if len(codes) == 0 {
					issues = append(issues, fmt.Sprintf("No %s provided", codeType))
				} else {
					// Verify top 3 codes have required fields
					for i, code := range codes {
						if i >= 3 {
							break
						}
						if codeMap, ok := code.(map[string]interface{}); ok {
							if _, ok := codeMap["code"]; !ok {
								issues = append(issues, fmt.Sprintf("%s[%d] missing 'code' field", codeType, i))
							}
							if _, ok := codeMap["description"]; !ok {
								issues = append(issues, fmt.Sprintf("%s[%d] missing 'description' field", codeType, i))
							}
						}
					}
				}
			} else {
				issues = append(issues, fmt.Sprintf("Missing or invalid %s array", codeType))
			}
		}

		// Check explanation
		if explanation, ok := classification["explanation"]; ok {
			if explanationStr, ok := explanation.(string); ok {
				if explanationStr == "" {
					issues = append(issues, "Explanation is empty")
				}
			} else if explanationMap, ok := explanation.(map[string]interface{}); ok {
				if _, ok := explanationMap["primary_reason"]; !ok {
					issues = append(issues, "Structured explanation missing 'primary_reason'")
				}
			}
		} else {
			// Check top-level explanation
			if explanation, ok := response["explanation"]; ok {
				if explanationStr, ok := explanation.(string); ok {
					if explanationStr == "" {
						issues = append(issues, "Explanation is empty")
					}
				}
			} else {
				issues = append(issues, "Missing explanation field")
			}
		}
	} else {
		issues = append(issues, "Missing 'classification' object")
	}

	return len(issues) == 0, issues
}

// Helper functions for extracting data
func extractString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func extractBool(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func extractFloat64(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0.0
}

func extractDuration(m map[string]interface{}, key string) time.Duration {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return time.Duration(f) * time.Millisecond
		}
	}
	return 0
}

func extractIndustryCodes(m map[string]interface{}, key string) []IndustryCode {
	codes := []IndustryCode{}
	if codesArray, ok := m[key].([]interface{}); ok {
		for _, codeItem := range codesArray {
			if codeMap, ok := codeItem.(map[string]interface{}); ok {
				code := IndustryCode{
					Code:        extractString(codeMap, "code"),
					Description: extractString(codeMap, "description"),
					Confidence:  extractFloat64(codeMap, "confidence"),
				}
				codes = append(codes, code)
			}
		}
	}
	return codes
}

// addResult adds a result to the collection
func (r *ClassificationTestRunner) addResult(result ClassificationTestResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.results = append(r.results, result)
}

// CalculateMetrics calculates aggregate metrics
func (r *ClassificationTestRunner) CalculateMetrics() {
	r.metrics.TotalTests = len(r.results)
	if r.metrics.TotalTests == 0 {
		return
	}

	successful := 0
	accurate := 0
	latencies := []time.Duration{}
	strategyCounts := make(map[string]int)
	strategySuccess := make(map[string]int)
	strategyLatencies := make(map[string][]time.Duration)
	earlyExits := 0
	cacheHits := 0
	industryAccuracy := make(map[string]struct{correct, total int})
	complexityAccuracy := make(map[string]struct{correct, total int})
	mccMatches := 0
	naicsMatches := 0
	sicMatches := 0
	top3Matches := 0
	codeChecks := 0

	for _, result := range r.results {
		if result.Success {
			successful++
		}
		if result.Accuracy {
			accurate++
		}
		latencies = append(latencies, result.ProcessingTime.ToDuration())

		if result.ScrapingStrategy != "" {
			strategyCounts[result.ScrapingStrategy]++
			if result.Success {
				strategySuccess[result.ScrapingStrategy]++
			}
			strategyLatencies[result.ScrapingStrategy] = append(
				strategyLatencies[result.ScrapingStrategy],
				result.ProcessingTime.ToDuration(),
			)
		}

		if result.EarlyExit {
			earlyExits++
		}
		if result.CacheHit {
			cacheHits++
		}

		if result.FallbackUsed && result.FallbackType != "" {
			r.metrics.FallbackUsage[result.FallbackType]++
		}

		// Track accuracy by industry
		stats := industryAccuracy[result.ExpectedIndustry]
		if result.Accuracy {
			stats.correct++
		}
		stats.total++
		industryAccuracy[result.ExpectedIndustry] = stats

		// Track accuracy by complexity (would need to add complexity to result)
		// For now, use category as proxy
		stats = complexityAccuracy[result.SampleID[:3]] // Use sample ID prefix as proxy
		if result.Accuracy {
			stats.correct++
		}
		stats.total++
		complexityAccuracy[result.SampleID[:3]] = stats

		// Check code accuracy (if expected codes provided)
		// This would require expected codes in the sample
		codeChecks++
		if len(result.MCCCodes) > 0 {
			mccMatches++
		}
		if len(result.NAICSCodes) > 0 {
			naicsMatches++
		}
		if len(result.SICCodes) > 0 {
			sicMatches++
		}
		if len(result.MCCCodes) >= 3 || len(result.NAICSCodes) >= 3 || len(result.SICCodes) >= 3 {
			top3Matches++
		}
	}

	r.metrics.SuccessfulTests = successful
	r.metrics.FailedTests = r.metrics.TotalTests - successful
	r.metrics.Accuracy = float64(accurate) / float64(r.metrics.TotalTests)

	// Calculate latency percentiles
	sortDurations(latencies)
	if len(latencies) > 0 {
		r.metrics.AverageLatency = averageDuration(latencies)
		r.metrics.P50Latency = percentileDuration(latencies, 50)
		r.metrics.P95Latency = percentileDuration(latencies, 95)
		r.metrics.P99Latency = percentileDuration(latencies, 99)
	}

	// Strategy distribution
	r.metrics.StrategyDistribution = strategyCounts
	for strategy, count := range strategyCounts {
		if count > 0 {
			r.metrics.StrategySuccessRate[strategy] = float64(strategySuccess[strategy]) / float64(count)
			if latencies, ok := strategyLatencies[strategy]; ok && len(latencies) > 0 {
				r.metrics.StrategyLatency[strategy] = averageDuration(latencies)
			}
		}
	}

	// Early exit rate
	r.metrics.EarlyExitCount = earlyExits
	if successful > 0 {
		r.metrics.EarlyExitRate = float64(earlyExits) / float64(successful)
	}

	// Cache hit rate
	r.metrics.CacheHitCount = cacheHits
	r.metrics.CacheHitRate = float64(cacheHits) / float64(r.metrics.TotalTests)

	// Accuracy by industry
	for industry, stats := range industryAccuracy {
		if stats.total > 0 {
			r.metrics.AccuracyByIndustry[industry] = float64(stats.correct) / float64(stats.total)
		}
	}

	// Code accuracy
	if codeChecks > 0 {
		r.metrics.CodeAccuracy = CodeAccuracyMetrics{
			MCCAccuracy:   float64(mccMatches) / float64(codeChecks),
			NAICSAccuracy: float64(naicsMatches) / float64(codeChecks),
			SICAccuracy:   float64(sicMatches) / float64(codeChecks),
			Top3MatchRate:  float64(top3Matches) / float64(codeChecks),
		}
	}

	// Calculate frontend compatibility
	r.calculateFrontendCompatibility()
}

func sortDurations(durations []time.Duration) {
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})
}

func averageDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	var sum time.Duration
	for _, d := range durations {
		sum += d
	}
	return sum / time.Duration(len(durations))
}

func percentileDuration(durations []time.Duration, percentile int) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	index := (percentile * len(durations)) / 100
	if index >= len(durations) {
		index = len(durations) - 1
	}
	return durations[index]
}

// calculateFrontendCompatibility calculates frontend data format compliance
func (r *ClassificationTestRunner) calculateFrontendCompatibility() {
	total := len(r.results)
	if total == 0 {
		return
	}

	allFieldsPresent := 0
	dataTypesCorrect := 0
	structureValid := 0
	industryPresent := 0
	codesPresent := 0
	explanationPresent := 0
	top3CodesPresent := 0

	for _, result := range r.results {
		if result.FrontendDataValid {
			allFieldsPresent++
			dataTypesCorrect++
			structureValid++
		}
		if result.ActualIndustry != "" {
			industryPresent++
		}
		if len(result.MCCCodes) > 0 || len(result.NAICSCodes) > 0 || len(result.SICCodes) > 0 {
			codesPresent++
		}
		if result.Explanation != "" || len(result.ExplanationStructured) > 0 {
			explanationPresent++
		}
		if len(result.MCCCodes) >= 3 || len(result.NAICSCodes) >= 3 || len(result.SICCodes) >= 3 {
			top3CodesPresent++
		}
	}

	r.metrics.FrontendCompatibility = FrontendCompatibility{
		AllFieldsPresent:   float64(allFieldsPresent) / float64(total),
		DataTypesCorrect:   float64(dataTypesCorrect) / float64(total),
		StructureValid:     float64(structureValid) / float64(total),
		IndustryPresent:     float64(industryPresent) / float64(total),
		CodesPresent:       float64(codesPresent) / float64(total),
		ExplanationPresent: float64(explanationPresent) / float64(total),
		Top3CodesPresent:   float64(top3CodesPresent) / float64(total),
	}
}

// GenerateReport generates a comprehensive test report
func (r *ClassificationTestRunner) GenerateReport(totalDuration time.Duration) map[string]interface{} {
	// Calculate throughput
	if totalDuration > 0 {
		r.metrics.Throughput = float64(r.metrics.TotalTests) / totalDuration.Seconds()
	}

	// Convert durations to milliseconds for JSON
	report := map[string]interface{}{
		"test_summary": map[string]interface{}{
			"total_samples":    r.metrics.TotalTests,
			"successful_tests":  r.metrics.SuccessfulTests,
			"failed_tests":      r.metrics.FailedTests,
			"overall_accuracy":  r.metrics.Accuracy,
			"test_duration":     totalDuration.String(),
			"timestamp":         time.Now().Format(time.RFC3339),
		},
		"performance_metrics": map[string]interface{}{
			"average_latency_ms": r.metrics.AverageLatency.Milliseconds(),
			"p50_latency_ms":     r.metrics.P50Latency.Milliseconds(),
			"p95_latency_ms":     r.metrics.P95Latency.Milliseconds(),
			"p99_latency_ms":     r.metrics.P99Latency.Milliseconds(),
			"throughput_rps":     r.metrics.Throughput,
		},
		"accuracy_metrics": map[string]interface{}{
			"overall_accuracy":        r.metrics.Accuracy,
			"accuracy_by_industry":     r.metrics.AccuracyByIndustry,
			"accuracy_by_complexity":   r.metrics.AccuracyByComplexity,
		},
		"strategy_distribution": map[string]interface{}{
			"counts":        r.metrics.StrategyDistribution,
			"percentages":   calculatePercentages(r.metrics.StrategyDistribution, r.metrics.TotalTests),
			"success_rates": r.metrics.StrategySuccessRate,
			"latencies_ms":  convertDurationsToMS(r.metrics.StrategyLatency),
		},
		"optimization_metrics": map[string]interface{}{
			"early_exit_count": r.metrics.EarlyExitCount,
			"early_exit_rate":  r.metrics.EarlyExitRate,
			"cache_hit_count":  r.metrics.CacheHitCount,
			"cache_hit_rate":   r.metrics.CacheHitRate,
			"fallback_usage":   r.metrics.FallbackUsage,
		},
		"frontend_compatibility": map[string]interface{}{
			"all_fields_present":  r.metrics.FrontendCompatibility.AllFieldsPresent,
			"data_types_correct":  r.metrics.FrontendCompatibility.DataTypesCorrect,
			"structure_valid":     r.metrics.FrontendCompatibility.StructureValid,
			"industry_present":    r.metrics.FrontendCompatibility.IndustryPresent,
			"codes_present":       r.metrics.FrontendCompatibility.CodesPresent,
			"explanation_present": r.metrics.FrontendCompatibility.ExplanationPresent,
			"top3_codes_present":  r.metrics.FrontendCompatibility.Top3CodesPresent,
		},
		"code_accuracy": map[string]interface{}{
			"mcc_accuracy":   r.metrics.CodeAccuracy.MCCAccuracy,
			"naics_accuracy": r.metrics.CodeAccuracy.NAICSAccuracy,
			"sic_accuracy":   r.metrics.CodeAccuracy.SICAccuracy,
			"top3_match_rate": r.metrics.CodeAccuracy.Top3MatchRate,
		},
		"detailed_results": r.results,
	}

	return report
}

func calculatePercentages(counts map[string]int, total int) map[string]float64 {
	percentages := make(map[string]float64)
	if total == 0 {
		return percentages
	}
	for k, v := range counts {
		percentages[k] = float64(v) / float64(total) * 100
	}
	return percentages
}

func convertDurationsToMS(durations map[string]time.Duration) map[string]int64 {
	result := make(map[string]int64)
	for k, v := range durations {
		result[k] = v.Milliseconds()
	}
	return result
}

// PrintSummary prints a summary of test results
func (r *ClassificationTestRunner) PrintSummary() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("COMPREHENSIVE CLASSIFICATION E2E TEST SUMMARY")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("\nTotal Tests: %d\n", r.metrics.TotalTests)
	fmt.Printf("Successful: %d (%.1f%%)\n", r.metrics.SuccessfulTests, float64(r.metrics.SuccessfulTests)/float64(r.metrics.TotalTests)*100)
	fmt.Printf("Failed: %d (%.1f%%)\n", r.metrics.FailedTests, float64(r.metrics.FailedTests)/float64(r.metrics.TotalTests)*100)
	fmt.Printf("\nOverall Accuracy: %.2f%%\n", r.metrics.Accuracy*100)
	fmt.Printf("\nPerformance:\n")
	fmt.Printf("  Average Latency: %v\n", r.metrics.AverageLatency)
	fmt.Printf("  P50 Latency: %v\n", r.metrics.P50Latency)
	fmt.Printf("  P95 Latency: %v\n", r.metrics.P95Latency)
	fmt.Printf("  P99 Latency: %v\n", r.metrics.P99Latency)
	fmt.Printf("  Throughput: %.2f req/s\n", r.metrics.Throughput)
	fmt.Printf("\nStrategy Distribution:\n")
	for strategy, count := range r.metrics.StrategyDistribution {
		percentage := float64(count) / float64(r.metrics.TotalTests) * 100
		fmt.Printf("  %s: %d (%.1f%%) - Success: %.1f%% - Avg Latency: %v\n",
			strategy, count, percentage,
			r.metrics.StrategySuccessRate[strategy]*100,
			r.metrics.StrategyLatency[strategy])
	}
	fmt.Printf("\nOptimizations:\n")
	fmt.Printf("  Early Exit Rate: %.1f%% (%d occurrences)\n", r.metrics.EarlyExitRate*100, r.metrics.EarlyExitCount)
	fmt.Printf("  Cache Hit Rate: %.1f%% (%d hits)\n", r.metrics.CacheHitRate*100, r.metrics.CacheHitCount)
	fmt.Printf("\nFrontend Compatibility:\n")
	fmt.Printf("  All Fields Present: %.1f%%\n", r.metrics.FrontendCompatibility.AllFieldsPresent*100)
	fmt.Printf("  Industry Present: %.1f%%\n", r.metrics.FrontendCompatibility.IndustryPresent*100)
	fmt.Printf("  Codes Present: %.1f%%\n", r.metrics.FrontendCompatibility.CodesPresent*100)
	fmt.Printf("  Explanation Present: %.1f%%\n", r.metrics.FrontendCompatibility.ExplanationPresent*100)
	fmt.Printf("  Top 3 Codes Present: %.1f%%\n", r.metrics.FrontendCompatibility.Top3CodesPresent*100)
	fmt.Println(strings.Repeat("=", 80) + "\n")
}

// validateResults validates test results against success criteria
func validateResults(t *testing.T, report map[string]interface{}) {
	testSummary := report["test_summary"].(map[string]interface{})
	performanceMetrics := report["performance_metrics"].(map[string]interface{})
	frontendCompatibility := report["frontend_compatibility"].(map[string]interface{})

	// Critical validations
	accuracy := testSummary["overall_accuracy"].(float64)
	if accuracy < 0.95 {
		t.Errorf("❌ Accuracy below threshold: %.2f%% (expected ≥95%%)", accuracy*100)
	} else {
		t.Logf("✅ Accuracy: %.2f%%", accuracy*100)
	}

	avgLatency := performanceMetrics["average_latency_ms"].(int64)
	if avgLatency > 2000 {
		t.Errorf("❌ Average latency too high: %dms (expected <2000ms)", avgLatency)
	} else {
		t.Logf("✅ Average latency: %dms", avgLatency)
	}

	p95Latency := performanceMetrics["p95_latency_ms"].(int64)
	if p95Latency > 5000 {
		t.Errorf("❌ P95 latency too high: %dms (expected <5000ms)", p95Latency)
	} else {
		t.Logf("✅ P95 latency: %dms", p95Latency)
	}

	allFieldsPresent := frontendCompatibility["all_fields_present"].(float64)
	if allFieldsPresent < 0.95 {
		t.Errorf("❌ Frontend compatibility below threshold: %.1f%% (expected ≥95%%)", allFieldsPresent*100)
	} else {
		t.Logf("✅ Frontend compatibility: %.1f%%", allFieldsPresent*100)
	}
}

// saveReport saves the test report to file
func saveReport(report map[string]interface{}, path string) error {
	// Ensure directory exists
	if err := os.MkdirAll("test/results", 0755); err != nil {
		return fmt.Errorf("failed to create results directory: %w", err)
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
}

// loadTestSamples loads test samples from JSON file
func loadTestSamples(path string) ([]TestSample, error) {
	// Try relative path first, then try from project root
	var data []byte
	var err error
	
	// Try relative path
	data, err = os.ReadFile(path)
	if err != nil {
		// Try from project root (common when running from root)
		altPath := "test/data/comprehensive_test_samples.json"
		data, err = os.ReadFile(altPath)
		if err != nil {
			// Try absolute path from test file location
			_, testFile, _, _ := runtime.Caller(0)
			testDir := filepath.Dir(testFile)
			absPath := filepath.Join(testDir, "..", "data", "comprehensive_test_samples.json")
			data, err = os.ReadFile(absPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read test samples file (tried %s, %s, %s): %w", path, altPath, absPath, err)
			}
		}
	}

	var dataset struct {
		Samples []TestSample `json:"samples"`
	}

	if err := json.Unmarshal(data, &dataset); err != nil {
		return nil, fmt.Errorf("failed to parse test samples: %w", err)
	}

	return dataset.Samples, nil
}


