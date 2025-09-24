package classification

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// ValidationTools provides comprehensive keyword testing and validation functionality
type ValidationTools struct {
	db *sql.DB
}

// NewValidationTools creates a new instance of ValidationTools
func NewValidationTools(db *sql.DB) *ValidationTools {
	return &ValidationTools{
		db: db,
	}
}

// TestResult represents the result of a keyword classification test
type TestResult struct {
	IndustryName        string                 `json:"industry_name"`
	ConfidenceScore     float64                `json:"confidence_score"`
	MatchedKeywords     []string               `json:"matched_keywords"`
	ClassificationCodes map[string]interface{} `json:"classification_codes"`
}

// CoverageResult represents keyword coverage validation results
type CoverageResult struct {
	IndustryName         string  `json:"industry_name"`
	TotalKeywords        int     `json:"total_keywords"`
	HighWeightKeywords   int     `json:"high_weight_keywords"`
	MediumWeightKeywords int     `json:"medium_weight_keywords"`
	LowWeightKeywords    int     `json:"low_weight_keywords"`
	CoverageScore        float64 `json:"coverage_score"`
}

// DuplicateKeyword represents duplicate keyword analysis results
type DuplicateKeyword struct {
	Keyword       string   `json:"keyword"`
	IndustryCount int      `json:"industry_count"`
	Industries    []string `json:"industries"`
	ConflictLevel string   `json:"conflict_level"`
}

// PatternTestResult represents keyword pattern testing results
type PatternTestResult struct {
	ExtractedKeyword  string    `json:"extracted_keyword"`
	MatchedIndustries []string  `json:"matched_industries"`
	ConfidenceScores  []float64 `json:"confidence_scores"`
	BestMatch         string    `json:"best_match"`
}

// EffectivenessResult represents keyword effectiveness analysis
type EffectivenessResult struct {
	Keyword            string  `json:"keyword"`
	IndustryName       string  `json:"industry_name"`
	UsageCount         int     `json:"usage_count"`
	SuccessRate        float64 `json:"success_rate"`
	AvgConfidence      float64 `json:"avg_confidence"`
	EffectivenessScore float64 `json:"effectiveness_score"`
}

// ImprovementSuggestion represents keyword improvement suggestions
type ImprovementSuggestion struct {
	IndustryName       string   `json:"industry_name"`
	CurrentKeywords    int      `json:"current_keywords"`
	SuggestedAdditions []string `json:"suggested_additions"`
	SuggestedRemovals  []string `json:"suggested_removals"`
	ImprovementScore   float64  `json:"improvement_score"`
}

// ConsistencyResult represents classification consistency validation
type ConsistencyResult struct {
	IndustryName     string  `json:"industry_name"`
	TotalCodes       int     `json:"total_codes"`
	MCCCodes         int     `json:"mcc_codes"`
	NAICSCodes       int     `json:"naics_codes"`
	SICCodes         int     `json:"sic_codes"`
	ConsistencyScore float64 `json:"consistency_score"`
}

// TestReport represents a comprehensive test report
type TestReport struct {
	TestCase         string   `json:"test_case"`
	ExpectedIndustry string   `json:"expected_industry"`
	ActualIndustry   string   `json:"actual_industry"`
	ConfidenceScore  float64  `json:"confidence_score"`
	MatchedKeywords  []string `json:"matched_keywords"`
	TestResult       string   `json:"test_result"`
	Recommendations  []string `json:"recommendations"`
}

// PerformanceResult represents keyword performance monitoring results
type PerformanceResult struct {
	Keyword          string  `json:"keyword"`
	IndustryName     string  `json:"industry_name"`
	RequestCount     int     `json:"request_count"`
	SuccessCount     int     `json:"success_count"`
	AvgResponseTime  float64 `json:"avg_response_time"`
	ErrorCount       int     `json:"error_count"`
	PerformanceScore float64 `json:"performance_score"`
}

// WeightOptimization represents keyword weight optimization results
type WeightOptimization struct {
	IndustryName    string  `json:"industry_name"`
	Keyword         string  `json:"keyword"`
	CurrentWeight   float64 `json:"current_weight"`
	SuggestedWeight float64 `json:"suggested_weight"`
	WeightChange    float64 `json:"weight_change"`
	Reason          string  `json:"reason"`
}

// CompletenessResult represents keyword completeness validation
type CompletenessResult struct {
	IndustryName      string   `json:"industry_name"`
	MissingKeywords   []string `json:"missing_keywords"`
	CompletenessScore float64  `json:"completeness_score"`
	Recommendations   []string `json:"recommendations"`
}

// EdgeCaseTest represents edge case testing results
type EdgeCaseTest struct {
	TestCase       string `json:"test_case"`
	InputData      string `json:"input_data"`
	ExpectedResult string `json:"expected_result"`
	ActualResult   string `json:"actual_result"`
	TestStatus     string `json:"test_status"`
	Notes          string `json:"notes"`
}

// Statistics represents keyword statistics
type Statistics struct {
	MetricName  string `json:"metric_name"`
	MetricValue string `json:"metric_value"`
	Description string `json:"description"`
}

// CompletionValidation represents testing completion validation
type CompletionValidation struct {
	TestCategory    string `json:"test_category"`
	TestName        string `json:"test_name"`
	Status          string `json:"status"`
	Details         string `json:"details"`
	Recommendations string `json:"recommendations"`
}

// ComprehensiveTestSuite represents comprehensive test suite results
type ComprehensiveTestSuite struct {
	TestSuite     string  `json:"test_suite"`
	TotalTests    int     `json:"total_tests"`
	PassedTests   int     `json:"passed_tests"`
	FailedTests   int     `json:"failed_tests"`
	SuccessRate   float64 `json:"success_rate"`
	OverallStatus string  `json:"overall_status"`
}

// TestKeywordClassification tests keyword classification functionality
func (vt *ValidationTools) TestKeywordClassification(ctx context.Context, keywords []string, businessName, description string) (*TestResult, error) {
	query := `
		SELECT industry_name, confidence_score, matched_keywords, classification_codes
		FROM test_keyword_classification($1, $2, $3)
	`

	var result TestResult
	var matchedKeywordsStr string
	var codesStr string

	err := vt.db.QueryRowContext(ctx, query, keywords, businessName, description).Scan(
		&result.IndustryName,
		&result.ConfidenceScore,
		&matchedKeywordsStr,
		&codesStr,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to test keyword classification: %w", err)
	}

	// Parse matched keywords
	if matchedKeywordsStr != "" {
		result.MatchedKeywords = strings.Split(matchedKeywordsStr, ",")
	}

	// Parse classification codes
	if codesStr != "" {
		if err := json.Unmarshal([]byte(codesStr), &result.ClassificationCodes); err != nil {
			log.Printf("Warning: failed to parse classification codes: %v", err)
		}
	}

	return &result, nil
}

// ValidateKeywordCoverage validates keyword coverage across industries
func (vt *ValidationTools) ValidateKeywordCoverage(ctx context.Context, industryFilter string) ([]CoverageResult, error) {
	query := `
		SELECT industry_name, total_keywords, high_weight_keywords, 
		       medium_weight_keywords, low_weight_keywords, coverage_score
		FROM validate_keyword_coverage($1)
		ORDER BY coverage_score DESC
	`

	rows, err := vt.db.QueryContext(ctx, query, industryFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to validate keyword coverage: %w", err)
	}
	defer rows.Close()

	var results []CoverageResult
	for rows.Next() {
		var result CoverageResult
		err := rows.Scan(
			&result.IndustryName,
			&result.TotalKeywords,
			&result.HighWeightKeywords,
			&result.MediumWeightKeywords,
			&result.LowWeightKeywords,
			&result.CoverageScore,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan coverage result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// FindDuplicateKeywords finds keywords used in multiple industries
func (vt *ValidationTools) FindDuplicateKeywords(ctx context.Context) ([]DuplicateKeyword, error) {
	query := `
		SELECT keyword, industry_count, industries, conflict_level
		FROM find_duplicate_keywords()
		ORDER BY industry_count DESC, keyword
	`

	rows, err := vt.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find duplicate keywords: %w", err)
	}
	defer rows.Close()

	var results []DuplicateKeyword
	for rows.Next() {
		var result DuplicateKeyword
		var industriesStr string

		err := rows.Scan(
			&result.Keyword,
			&result.IndustryCount,
			&industriesStr,
			&result.ConflictLevel,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan duplicate keyword result: %w", err)
		}

		// Parse industries
		if industriesStr != "" {
			result.Industries = strings.Split(industriesStr, ",")
		}

		results = append(results, result)
	}

	return results, nil
}

// TestKeywordPatterns tests keyword pattern extraction and matching
func (vt *ValidationTools) TestKeywordPatterns(ctx context.Context, testText string) ([]PatternTestResult, error) {
	query := `
		SELECT extracted_keyword, matched_industries, confidence_scores, best_match
		FROM test_keyword_patterns($1)
	`

	rows, err := vt.db.QueryContext(ctx, query, testText)
	if err != nil {
		return nil, fmt.Errorf("failed to test keyword patterns: %w", err)
	}
	defer rows.Close()

	var results []PatternTestResult
	for rows.Next() {
		var result PatternTestResult
		var industriesStr, scoresStr string

		err := rows.Scan(
			&result.ExtractedKeyword,
			&industriesStr,
			&scoresStr,
			&result.BestMatch,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pattern test result: %w", err)
		}

		// Parse industries
		if industriesStr != "" {
			result.MatchedIndustries = strings.Split(industriesStr, ",")
		}

		// Parse confidence scores
		if scoresStr != "" {
			scoreStrs := strings.Split(scoresStr, ",")
			for _, scoreStr := range scoreStrs {
				var score float64
				if _, err := fmt.Sscanf(scoreStr, "%f", &score); err == nil {
					result.ConfidenceScores = append(result.ConfidenceScores, score)
				}
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// AnalyzeKeywordEffectiveness analyzes keyword effectiveness over time
func (vt *ValidationTools) AnalyzeKeywordEffectiveness(ctx context.Context, daysBack int) ([]EffectivenessResult, error) {
	query := `
		SELECT keyword, industry_name, usage_count, success_rate, 
		       avg_confidence, effectiveness_score
		FROM analyze_keyword_effectiveness($1)
		ORDER BY effectiveness_score DESC
	`

	rows, err := vt.db.QueryContext(ctx, query, daysBack)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze keyword effectiveness: %w", err)
	}
	defer rows.Close()

	var results []EffectivenessResult
	for rows.Next() {
		var result EffectivenessResult
		err := rows.Scan(
			&result.Keyword,
			&result.IndustryName,
			&result.UsageCount,
			&result.SuccessRate,
			&result.AvgConfidence,
			&result.EffectivenessScore,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan effectiveness result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// SuggestKeywordImprovements suggests keyword improvements for industries
func (vt *ValidationTools) SuggestKeywordImprovements(ctx context.Context, industryFilter string) ([]ImprovementSuggestion, error) {
	query := `
		SELECT industry_name, current_keywords, suggested_additions, 
		       suggested_removals, improvement_score
		FROM suggest_keyword_improvements($1)
		ORDER BY improvement_score DESC
	`

	rows, err := vt.db.QueryContext(ctx, query, industryFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to suggest keyword improvements: %w", err)
	}
	defer rows.Close()

	var results []ImprovementSuggestion
	for rows.Next() {
		var result ImprovementSuggestion
		var additionsStr, removalsStr string

		err := rows.Scan(
			&result.IndustryName,
			&result.CurrentKeywords,
			&additionsStr,
			&removalsStr,
			&result.ImprovementScore,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan improvement suggestion: %w", err)
		}

		// Parse suggested additions
		if additionsStr != "" {
			result.SuggestedAdditions = strings.Split(additionsStr, ",")
		}

		// Parse suggested removals
		if removalsStr != "" {
			result.SuggestedRemovals = strings.Split(removalsStr, ",")
		}

		results = append(results, result)
	}

	return results, nil
}

// ValidateClassificationConsistency validates classification code consistency
func (vt *ValidationTools) ValidateClassificationConsistency(ctx context.Context) ([]ConsistencyResult, error) {
	query := `
		SELECT industry_name, total_codes, mcc_codes, naics_codes, 
		       sic_codes, consistency_score
		FROM validate_classification_consistency()
		ORDER BY consistency_score DESC
	`

	rows, err := vt.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to validate classification consistency: %w", err)
	}
	defer rows.Close()

	var results []ConsistencyResult
	for rows.Next() {
		var result ConsistencyResult
		err := rows.Scan(
			&result.IndustryName,
			&result.TotalCodes,
			&result.MCCCodes,
			&result.NAICSCodes,
			&result.SICCodes,
			&result.ConsistencyScore,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan consistency result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GenerateKeywordTestReport generates comprehensive test reports
func (vt *ValidationTools) GenerateKeywordTestReport(ctx context.Context, testCases []map[string]interface{}) ([]TestReport, error) {
	// Convert test cases to JSON
	testCasesJSON, err := json.Marshal(testCases)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal test cases: %w", err)
	}

	query := `
		SELECT test_case, expected_industry, actual_industry, confidence_score,
		       matched_keywords, test_result, recommendations
		FROM generate_keyword_test_report($1)
	`

	rows, err := vt.db.QueryContext(ctx, query, string(testCasesJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to generate keyword test report: %w", err)
	}
	defer rows.Close()

	var results []TestReport
	for rows.Next() {
		var result TestReport
		var keywordsStr, recommendationsStr string

		err := rows.Scan(
			&result.TestCase,
			&result.ExpectedIndustry,
			&result.ActualIndustry,
			&result.ConfidenceScore,
			&keywordsStr,
			&result.TestResult,
			&recommendationsStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan test report result: %w", err)
		}

		// Parse matched keywords
		if keywordsStr != "" {
			result.MatchedKeywords = strings.Split(keywordsStr, ",")
		}

		// Parse recommendations
		if recommendationsStr != "" {
			result.Recommendations = strings.Split(recommendationsStr, ",")
		}

		results = append(results, result)
	}

	return results, nil
}

// MonitorKeywordPerformance monitors keyword performance metrics
func (vt *ValidationTools) MonitorKeywordPerformance(ctx context.Context, hoursBack int) ([]PerformanceResult, error) {
	query := `
		SELECT keyword, industry_name, request_count, success_count,
		       avg_response_time, error_count, performance_score
		FROM monitor_keyword_performance($1)
		ORDER BY performance_score DESC
	`

	rows, err := vt.db.QueryContext(ctx, query, hoursBack)
	if err != nil {
		return nil, fmt.Errorf("failed to monitor keyword performance: %w", err)
	}
	defer rows.Close()

	var results []PerformanceResult
	for rows.Next() {
		var result PerformanceResult
		err := rows.Scan(
			&result.Keyword,
			&result.IndustryName,
			&result.RequestCount,
			&result.SuccessCount,
			&result.AvgResponseTime,
			&result.ErrorCount,
			&result.PerformanceScore,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan performance result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// OptimizeKeywordWeights optimizes keyword weights based on performance
func (vt *ValidationTools) OptimizeKeywordWeights(ctx context.Context, industryFilter string) ([]WeightOptimization, error) {
	query := `
		SELECT industry_name, keyword, current_weight, suggested_weight,
		       weight_change, reason
		FROM optimize_keyword_weights($1)
		ORDER BY ABS(weight_change) DESC
	`

	rows, err := vt.db.QueryContext(ctx, query, industryFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to optimize keyword weights: %w", err)
	}
	defer rows.Close()

	var results []WeightOptimization
	for rows.Next() {
		var result WeightOptimization
		err := rows.Scan(
			&result.IndustryName,
			&result.Keyword,
			&result.CurrentWeight,
			&result.SuggestedWeight,
			&result.WeightChange,
			&result.Reason,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan weight optimization result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// ValidateKeywordCompleteness validates keyword completeness across industries
func (vt *ValidationTools) ValidateKeywordCompleteness(ctx context.Context) ([]CompletenessResult, error) {
	query := `
		SELECT industry_name, missing_keywords, completeness_score, recommendations
		FROM validate_keyword_completeness()
		ORDER BY completeness_score DESC
	`

	rows, err := vt.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to validate keyword completeness: %w", err)
	}
	defer rows.Close()

	var results []CompletenessResult
	for rows.Next() {
		var result CompletenessResult
		var missingStr, recommendationsStr string

		err := rows.Scan(
			&result.IndustryName,
			&missingStr,
			&result.CompletenessScore,
			&recommendationsStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan completeness result: %w", err)
		}

		// Parse missing keywords
		if missingStr != "" {
			result.MissingKeywords = strings.Split(missingStr, ",")
		}

		// Parse recommendations
		if recommendationsStr != "" {
			result.Recommendations = strings.Split(recommendationsStr, ",")
		}

		results = append(results, result)
	}

	return results, nil
}

// TestKeywordEdgeCases tests edge case handling and error conditions
func (vt *ValidationTools) TestKeywordEdgeCases(ctx context.Context) ([]EdgeCaseTest, error) {
	query := `
		SELECT test_case, input_data, expected_result, actual_result, test_status, notes
		FROM test_keyword_edge_cases()
	`

	rows, err := vt.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to test keyword edge cases: %w", err)
	}
	defer rows.Close()

	var results []EdgeCaseTest
	for rows.Next() {
		var result EdgeCaseTest
		err := rows.Scan(
			&result.TestCase,
			&result.InputData,
			&result.ExpectedResult,
			&result.ActualResult,
			&result.TestStatus,
			&result.Notes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan edge case test result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GenerateKeywordStatistics generates comprehensive keyword statistics
func (vt *ValidationTools) GenerateKeywordStatistics(ctx context.Context) ([]Statistics, error) {
	query := `
		SELECT metric_name, metric_value, description
		FROM generate_keyword_statistics()
	`

	rows, err := vt.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate keyword statistics: %w", err)
	}
	defer rows.Close()

	var results []Statistics
	for rows.Next() {
		var result Statistics
		err := rows.Scan(
			&result.MetricName,
			&result.MetricValue,
			&result.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan statistics result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// ValidateKeywordTestingCompletion validates all testing functions are working
func (vt *ValidationTools) ValidateKeywordTestingCompletion(ctx context.Context) ([]CompletionValidation, error) {
	query := `
		SELECT test_category, test_name, status, details, recommendations
		FROM validate_keyword_testing_completion()
	`

	rows, err := vt.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to validate keyword testing completion: %w", err)
	}
	defer rows.Close()

	var results []CompletionValidation
	for rows.Next() {
		var result CompletionValidation
		err := rows.Scan(
			&result.TestCategory,
			&result.TestName,
			&result.Status,
			&result.Details,
			&result.Recommendations,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan completion validation result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// RunComprehensiveKeywordTests runs comprehensive test suite and reports results
func (vt *ValidationTools) RunComprehensiveKeywordTests(ctx context.Context) (*ComprehensiveTestSuite, error) {
	query := `
		SELECT test_suite, total_tests, passed_tests, failed_tests, success_rate, overall_status
		FROM run_comprehensive_keyword_tests()
	`

	var result ComprehensiveTestSuite
	err := vt.db.QueryRowContext(ctx, query).Scan(
		&result.TestSuite,
		&result.TotalTests,
		&result.PassedTests,
		&result.FailedTests,
		&result.SuccessRate,
		&result.OverallStatus,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to run comprehensive keyword tests: %w", err)
	}

	return &result, nil
}

// ExecuteKeywordTest executes a specific keyword test by name
func (vt *ValidationTools) ExecuteKeywordTest(ctx context.Context, testName string) (string, error) {
	query := `SELECT execute_keyword_test($1)`

	var result string
	err := vt.db.QueryRowContext(ctx, query, testName).Scan(&result)
	if err != nil {
		return "", fmt.Errorf("failed to execute keyword test %s: %w", testName, err)
	}

	return result, nil
}

// GetTestingDashboard returns the testing dashboard information
func (vt *ValidationTools) GetTestingDashboard(ctx context.Context) ([]map[string]string, error) {
	query := `
		SELECT test_name, test_query, description
		FROM keyword_testing_dashboard
		ORDER BY test_name
	`

	rows, err := vt.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get testing dashboard: %w", err)
	}
	defer rows.Close()

	var results []map[string]string
	for rows.Next() {
		var testName, testQuery, description string
		err := rows.Scan(&testName, &testQuery, &description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dashboard result: %w", err)
		}

		results = append(results, map[string]string{
			"test_name":   testName,
			"test_query":  testQuery,
			"description": description,
		})
	}

	return results, nil
}

// LogKeywordTest logs a keyword test for monitoring and analysis
func (vt *ValidationTools) LogKeywordTest(ctx context.Context, keyword string, industry string, success bool, confidence float64, responseTime time.Duration) error {
	query := `
		INSERT INTO keyword_logs (keyword, industry, classification_success, confidence_score, response_time_ms, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`

	_, err := vt.db.ExecContext(ctx, query, keyword, industry, success, confidence, responseTime.Milliseconds())
	if err != nil {
		return fmt.Errorf("failed to log keyword test: %w", err)
	}

	return nil
}
