package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

// Phase2TestResult holds test results
type Phase2TestResult struct {
	TestName      string
	Passed        bool
	Message       string
	Details       map[string]interface{}
	ExecutionTime time.Duration
}

// TestPhase2_Comprehensive runs comprehensive Phase 2 tests
func TestPhase2_Comprehensive(t *testing.T) {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	logger.Println("=== Phase 2 Comprehensive Testing ===")
	logger.Println("")

	results := []Phase2TestResult{}

	// Test 1: Top 3 Codes Per Type
	t.Run("Top3CodesPerType", func(t *testing.T) {
		result := testTop3CodesPerType(baseURL, logger)
		results = append(results, result)
		if !result.Passed {
			t.Errorf("Test failed: %s", result.Message)
		}
	})

	// Test 2: Confidence Calibration
	t.Run("ConfidenceCalibration", func(t *testing.T) {
		result := testConfidenceCalibration(baseURL, logger)
		results = append(results, result)
		if !result.Passed {
			t.Errorf("Test failed: %s", result.Message)
		}
	})

	// Test 3: Fast Path
	t.Run("FastPath", func(t *testing.T) {
		result := testFastPath(baseURL, logger)
		results = append(results, result)
		if !result.Passed {
			t.Errorf("Test failed: %s", result.Message)
		}
	})

	// Test 4: Structured Explanations
	t.Run("StructuredExplanations", func(t *testing.T) {
		result := testStructuredExplanations(baseURL, logger)
		results = append(results, result)
		if !result.Passed {
			t.Errorf("Test failed: %s", result.Message)
		}
	})

	// Test 5: Generic Fallback
	t.Run("GenericFallback", func(t *testing.T) {
		result := testGenericFallback(baseURL, logger)
		results = append(results, result)
		if !result.Passed {
			t.Errorf("Test failed: %s", result.Message)
		}
	})

	// Test 6: Performance
	t.Run("Performance", func(t *testing.T) {
		result := testPerformance(baseURL, logger)
		results = append(results, result)
		if !result.Passed {
			t.Errorf("Test failed: %s", result.Message)
		}
	})

	// Print summary
	printSummary(results, logger)
}

func testTop3CodesPerType(baseURL string, logger *log.Logger) Phase2TestResult {
	start := time.Now()
	logger.Println("📋 Test 1: Top 3 Codes Per Type")

	testCases := []struct {
		name        string
		business    string
		description string
	}{
		{"Restaurant", "Joe's Pizza Restaurant", "Family pizza restaurant serving authentic Italian cuisine"},
		{"Tech Company", "Tech Startup Inc", "Software development and cloud services"},
		{"Retail Store", "Fashion Boutique", "Clothing and accessories retail store"},
	}

	allPassed := true
	details := make(map[string]interface{})
	details["test_cases"] = []map[string]interface{}{}

	for _, tc := range testCases {
		result := makeClassificationRequest(baseURL, tc.business, tc.description, "")
		if result == nil {
			allPassed = false
			continue
		}

		if result.Classification == nil {
			allPassed = false
			logger.Printf("  ❌ %s: Missing classification result", tc.name)
			continue
		}

		testResult := map[string]interface{}{
			"business":     tc.name,
			"mcc_count":    len(result.Classification.MCCCodes),
			"sic_count":    len(result.Classification.SICCodes),
			"naics_count":  len(result.Classification.NAICSCodes),
			"all_have_3":   len(result.Classification.MCCCodes) == 3 && len(result.Classification.SICCodes) == 3 && len(result.Classification.NAICSCodes) == 3,
			"all_have_source": checkAllCodesHaveSource(result),
		}

		details["test_cases"] = append(details["test_cases"].([]map[string]interface{}), testResult)

		if !testResult["all_have_3"].(bool) {
			allPassed = false
			logger.Printf("  ❌ %s: Expected 3 codes per type, got MCC:%d SIC:%d NAICS:%d",
				tc.name, len(result.Classification.MCCCodes), len(result.Classification.SICCodes), len(result.Classification.NAICSCodes))
		} else if !testResult["all_have_source"].(bool) {
			allPassed = false
			logger.Printf("  ❌ %s: Some codes missing Source field", tc.name)
		} else {
			logger.Printf("  ✅ %s: All 3 code types returned with Source fields", tc.name)
		}
	}

	return Phase2TestResult{
		TestName:      "Top3CodesPerType",
		Passed:        allPassed,
		Message:       fmt.Sprintf("Top 3 codes test: %v", allPassed),
		Details:       details,
		ExecutionTime: time.Since(start),
	}
}

func testConfidenceCalibration(baseURL string, logger *log.Logger) Phase2TestResult {
	start := time.Now()
	logger.Println("📊 Test 2: Confidence Calibration")

	testCases := []struct {
		name        string
		business    string
		description string
		expectedMin float64
		expectedMax float64
	}{
		{"High Quality", "Starbucks Coffee", "Coffee shop and cafe", 0.70, 0.95},
		{"Medium Quality", "ABC Services", "General business services", 0.60, 0.90},
		{"Low Quality", "XYZ Corp", "Business", 0.50, 0.85},
	}

	allPassed := true
	details := make(map[string]interface{})
	details["confidences"] = []float64{}

	for _, tc := range testCases {
		result := makeClassificationRequest(baseURL, tc.business, tc.description, "")
		if result == nil {
			allPassed = false
			continue
		}

		confidence := result.ConfidenceScore
		details["confidences"] = append(details["confidences"].([]float64), confidence)

		if confidence < tc.expectedMin || confidence > tc.expectedMax {
			allPassed = false
			logger.Printf("  ❌ %s: Confidence %.2f%% outside expected range [%.0f%%, %.0f%%]",
				tc.name, confidence*100, tc.expectedMin*100, tc.expectedMax*100)
		} else {
			logger.Printf("  ✅ %s: Confidence %.2f%% in range [%.0f%%, %.0f%%]",
				tc.name, confidence*100, tc.expectedMin*100, tc.expectedMax*100)
		}
	}

	return Phase2TestResult{
		TestName:      "ConfidenceCalibration",
		Passed:        allPassed,
		Message:       fmt.Sprintf("Confidence calibration test: %v", allPassed),
		Details:       details,
		ExecutionTime: time.Since(start),
	}
}

func testFastPath(baseURL string, logger *log.Logger) Phase2TestResult {
	start := time.Now()
	logger.Println("⚡ Test 3: Fast Path Performance")

	obviousCases := []struct {
		name        string
		business    string
		description string
	}{
		{"Restaurant", "Pizza Hut", "Pizza restaurant"},
		{"Coffee Shop", "Starbucks Coffee", "Coffee shop"},
		{"Hotel", "Hilton Hotel", "Hotel and lodging"},
		{"Bank", "Chase Bank", "Banking services"},
		{"Hospital", "General Hospital", "Medical hospital"},
	}

	fastPathCount := 0
	totalTime := time.Duration(0)
	details := make(map[string]interface{})
	details["latencies"] = []float64{}

	for _, tc := range obviousCases {
		reqStart := time.Now()
		result := makeClassificationRequest(baseURL, tc.business, tc.description, "")
		latency := time.Since(reqStart)
		totalTime += latency
		details["latencies"] = append(details["latencies"].([]float64), latency.Seconds()*1000)

		if result != nil && latency < 100*time.Millisecond {
			fastPathCount++
			logger.Printf("  ✅ %s: Fast path (<100ms) - %.0fms", tc.name, latency.Seconds()*1000)
		} else if result != nil {
			logger.Printf("  ⚠️  %s: Slow path (>=100ms) - %.0fms", tc.name, latency.Seconds()*1000)
		} else {
			logger.Printf("  ❌ %s: Request failed", tc.name)
		}
	}

	fastPathRate := float64(fastPathCount) / float64(len(obviousCases)) * 100
	avgLatency := totalTime / time.Duration(len(obviousCases))
	details["fast_path_rate"] = fastPathRate
	details["avg_latency_ms"] = avgLatency.Seconds() * 1000

	passed := fastPathRate >= 60.0 && avgLatency < 200*time.Millisecond

	logger.Printf("  Fast Path Hit Rate: %.1f%% (target: >=60%%)", fastPathRate)
	logger.Printf("  Average Latency: %.0fms (target: <200ms)", avgLatency.Seconds()*1000)

	return Phase2TestResult{
		TestName:      "FastPath",
		Passed:        passed,
		Message:       fmt.Sprintf("Fast path: %.1f%% hit rate, %.0fms avg latency", fastPathRate, avgLatency.Seconds()*1000),
		Details:       details,
		ExecutionTime: time.Since(start),
	}
}

func testStructuredExplanations(baseURL string, logger *log.Logger) Phase2TestResult {
	start := time.Now()
	logger.Println("📝 Test 4: Structured Explanations")

	testCases := []struct {
		name        string
		business    string
		description string
	}{
		{"Restaurant", "Mario's Italian Restaurant", "Authentic Italian restaurant"},
		{"Tech Company", "Cloud Services Inc", "Cloud computing and SaaS platform"},
	}

	allPassed := true
	details := make(map[string]interface{})
	details["explanations"] = []map[string]interface{}{}

	for _, tc := range testCases {
		result := makeClassificationRequest(baseURL, tc.business, tc.description, "")
		if result == nil || result.Classification == nil || result.Classification.Explanation == nil {
			allPassed = false
			logger.Printf("  ❌ %s: Missing explanation", tc.name)
			continue
		}

		expl := result.Classification.Explanation
		explResult := map[string]interface{}{
			"business":            tc.name,
			"has_primary_reason":  expl.PrimaryReason != "",
			"has_supporting":      len(expl.SupportingFactors) >= 3,
			"has_key_terms":       len(expl.KeyTermsFound) > 0,
			"has_method":          expl.MethodUsed != "",
			"has_processing_path":  expl.ProcessingPath != "",
			"supporting_count":    len(expl.SupportingFactors),
			"key_terms_count":     len(expl.KeyTermsFound),
		}

		details["explanations"] = append(details["explanations"].([]map[string]interface{}), explResult)

		if !explResult["has_primary_reason"].(bool) ||
			!explResult["has_supporting"].(bool) ||
			!explResult["has_key_terms"].(bool) {
			allPassed = false
			logger.Printf("  ❌ %s: Incomplete explanation", tc.name)
		} else {
			logger.Printf("  ✅ %s: Complete explanation (%d factors, %d key terms)",
				tc.name, len(expl.SupportingFactors), len(expl.KeyTermsFound))
		}
	}

	return Phase2TestResult{
		TestName:      "StructuredExplanations",
		Passed:        allPassed,
		Message:       fmt.Sprintf("Structured explanations test: %v", allPassed),
		Details:       details,
		ExecutionTime: time.Since(start),
	}
}

func testGenericFallback(baseURL string, logger *log.Logger) Phase2TestResult {
	start := time.Now()
	logger.Println("🔄 Test 5: Generic Fallback Fix")

	ambiguousCases := []struct {
		name        string
		business    string
		description string
	}{
		{"Ambiguous 1", "ABC Corporation", "General business services"},
		{"Ambiguous 2", "XYZ Services", "Business services"},
		{"Ambiguous 3", "Global Enterprises", "Corporate services"},
		{"Ambiguous 4", "Universal Business", "General services"},
		{"Ambiguous 5", "Main Street Company", "Local business"},
	}

	genericCount := 0
	details := make(map[string]interface{})
	details["industries"] = []string{}

	for _, tc := range ambiguousCases {
		result := makeClassificationRequest(baseURL, tc.business, tc.description, "")
		if result == nil {
			continue
		}

		var industry string
		if result.Classification != nil {
			industry = result.Classification.Industry
		} else if result.PrimaryIndustry != "" {
			industry = result.PrimaryIndustry
		} else {
			industry = "Unknown"
		}
		details["industries"] = append(details["industries"].([]string), industry)

		if industry == "General Business" {
			genericCount++
			logger.Printf("  ⚠️  %s: Classified as 'General Business'", tc.name)
		} else {
			logger.Printf("  ✅ %s: Classified as '%s' (specific)", tc.name, industry)
		}
	}

	genericRate := float64(genericCount) / float64(len(ambiguousCases)) * 100
	details["generic_rate"] = genericRate
	passed := genericRate < 10.0

	logger.Printf("  Generic Business Rate: %.1f%% (target: <10%%)", genericRate)

	return Phase2TestResult{
		TestName:      "GenericFallback",
		Passed:        passed,
		Message:       fmt.Sprintf("Generic fallback: %.1f%% (target: <10%%)", genericRate),
		Details:       details,
		ExecutionTime: time.Since(start),
	}
}

func testPerformance(baseURL string, logger *log.Logger) Phase2TestResult {
	start := time.Now()
	logger.Println("⚙️  Test 6: Overall Performance")

	testCases := []struct {
		name        string
		business    string
		description string
	}{
		{"Restaurant", "Joe's Pizza", "Pizza restaurant"},
		{"Tech", "Software Inc", "Software development"},
		{"Retail", "Fashion Store", "Clothing store"},
		{"Healthcare", "Medical Clinic", "Healthcare services"},
		{"Finance", "Investment Bank", "Financial services"},
	}

	var latencies []time.Duration
	allPassed := true

	for _, tc := range testCases {
		reqStart := time.Now()
		result := makeClassificationRequest(baseURL, tc.business, tc.description, "")
		latency := time.Since(reqStart)
		latencies = append(latencies, latency)

		if result == nil {
			allPassed = false
			logger.Printf("  ❌ %s: Request failed", tc.name)
		} else if latency > 500*time.Millisecond {
			logger.Printf("  ⚠️  %s: Slow (%.0fms)", tc.name, latency.Seconds()*1000)
		} else {
			logger.Printf("  ✅ %s: Fast (%.0fms)", tc.name, latency.Seconds()*1000)
		}
	}

	// Calculate percentiles
	sortDurations(latencies)
	p50 := latencies[len(latencies)/2]
	p90 := latencies[int(float64(len(latencies))*0.9)]
	p95 := latencies[int(float64(len(latencies))*0.95)]

	details := map[string]interface{}{
		"p50_ms": p50.Seconds() * 1000,
		"p90_ms": p90.Seconds() * 1000,
		"p95_ms": p95.Seconds() * 1000,
	}

	logger.Printf("  P50 Latency: %.0fms", p50.Seconds()*1000)
	logger.Printf("  P90 Latency: %.0fms", p90.Seconds()*1000)
	logger.Printf("  P95 Latency: %.0fms", p95.Seconds()*1000)

	passed := p95 < 500*time.Millisecond

	return Phase2TestResult{
		TestName:      "Performance",
		Passed:        passed && allPassed,
		Message:       fmt.Sprintf("Performance: P95=%.0fms (target: <500ms)", p95.Seconds()*1000),
		Details:       details,
		ExecutionTime: time.Since(start),
	}
}

// Helper functions

type ClassificationResponse struct {
	RequestID      string                 `json:"request_id"`
	BusinessName   string                 `json:"business_name"`
	Description    string                 `json:"description"`
	PrimaryIndustry string                `json:"primary_industry,omitempty"`
	Classification *ClassificationResult  `json:"classification"`
	ConfidenceScore float64               `json:"confidence_score"`
	Status          string                `json:"status"`
	Success         bool                  `json:"success"`
}

type ClassificationResult struct {
	Industry    string                 `json:"industry"`
	MCCCodes    []CodeInfo             `json:"mcc_codes"`
	SICCodes    []CodeInfo             `json:"sic_codes"`
	NAICSCodes  []CodeInfo             `json:"naics_codes"`
	Explanation *ExplanationInfo       `json:"explanation,omitempty"`
}

type CodeInfo struct {
	Code           string   `json:"code"`
	Description    string   `json:"description"`
	Confidence     float64  `json:"confidence"`
	Source         []string `json:"source,omitempty"` // Phase 2: Source is an array
	MatchType      string   `json:"matchType,omitempty"`
	RelevanceScore float64  `json:"relevanceScore,omitempty"`
}

type ExplanationInfo struct {
	PrimaryReason     string   `json:"primary_reason"`
	SupportingFactors []string `json:"supporting_factors"`
	KeyTermsFound     []string `json:"key_terms_found"`
	MethodUsed        string   `json:"method_used"`
	ProcessingPath    string   `json:"processing_path"`
}

func makeClassificationRequest(baseURL, businessName, description, websiteURL string) *ClassificationResponse {
	payload := map[string]interface{}{
		"business_name": businessName,
		"description":   description,
	}
	if websiteURL != "" {
		payload["website_url"] = websiteURL
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", baseURL+"/api/classify", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result ClassificationResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}

	return &result
}

func checkAllCodesHaveSource(result *ClassificationResponse) bool {
	if result.Classification == nil {
		return false
	}
	for _, code := range result.Classification.MCCCodes {
		if len(code.Source) == 0 {
			return false
		}
	}
	for _, code := range result.Classification.SICCodes {
		if len(code.Source) == 0 {
			return false
		}
	}
	for _, code := range result.Classification.NAICSCodes {
		if len(code.Source) == 0 {
			return false
		}
	}
	return true
}

func sortDurations(durations []time.Duration) {
	for i := 0; i < len(durations)-1; i++ {
		for j := i + 1; j < len(durations); j++ {
			if durations[i] > durations[j] {
				durations[i], durations[j] = durations[j], durations[i]
			}
		}
	}
}

func printSummary(results []Phase2TestResult, logger *log.Logger) {
	logger.Println("")
	logger.Println("=== Test Summary ===")
	logger.Println("")

	passed := 0
	total := len(results)

	for _, result := range results {
		status := "❌ FAIL"
		if result.Passed {
			status = "✅ PASS"
			passed++
		}
		logger.Printf("%s %s (%.2fs) - %s",
			status, result.TestName, result.ExecutionTime.Seconds(), result.Message)
	}

	logger.Println("")
	logger.Printf("Results: %d/%d tests passed (%.1f%%)", passed, total, float64(passed)/float64(total)*100)
	logger.Println("")
}
