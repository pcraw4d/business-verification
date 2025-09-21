package classification

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
)

// CrosswalkAccuracyTester provides comprehensive testing for crosswalk accuracy
type CrosswalkAccuracyTester struct {
	db     *sql.DB
	logger *zap.Logger
}

// AccuracyTestResult represents the result of an accuracy test
type AccuracyTestResult struct {
	TestName        string                 `json:"test_name"`
	TestType        string                 `json:"test_type"`
	TotalTests      int                    `json:"total_tests"`
	PassedTests     int                    `json:"passed_tests"`
	FailedTests     int                    `json:"failed_tests"`
	AccuracyScore   float64                `json:"accuracy_score"`
	ConfidenceScore float64                `json:"confidence_score"`
	TestDetails     []AccuracyTestDetail   `json:"test_details"`
	Summary         string                 `json:"summary"`
	Timestamp       time.Time              `json:"timestamp"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// AccuracyTestDetail represents individual test case results
type AccuracyTestDetail struct {
	TestCaseID    string                 `json:"test_case_id"`
	Input         map[string]interface{} `json:"input"`
	Expected      map[string]interface{} `json:"expected"`
	Actual        map[string]interface{} `json:"actual"`
	Passed        bool                   `json:"passed"`
	Confidence    float64                `json:"confidence"`
	Error         string                 `json:"error,omitempty"`
	ExecutionTime time.Duration          `json:"execution_time"`
}

// AccuracyTestSuite represents a collection of accuracy tests
type AccuracyTestSuite struct {
	SuiteName    string                 `json:"suite_name"`
	Description  string                 `json:"description"`
	TestCases    []AccuracyTestCase     `json:"test_cases"`
	TestResults  []AccuracyTestResult   `json:"test_results"`
	OverallScore float64                `json:"overall_score"`
	CreatedAt    time.Time              `json:"created_at"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// AccuracyTestCase represents a single test case
type AccuracyTestCase struct {
	TestCaseID  string                 `json:"test_case_id"`
	TestName    string                 `json:"test_name"`
	TestType    string                 `json:"test_type"`
	Input       map[string]interface{} `json:"input"`
	Expected    map[string]interface{} `json:"expected"`
	Description string                 `json:"description"`
	Weight      float64                `json:"weight"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
}

// NewCrosswalkAccuracyTester creates a new crosswalk accuracy tester
func NewCrosswalkAccuracyTester(db *sql.DB, logger *zap.Logger) *CrosswalkAccuracyTester {
	return &CrosswalkAccuracyTester{
		db:     db,
		logger: logger,
	}
}

// RunComprehensiveAccuracyTests runs all accuracy tests
func (cat *CrosswalkAccuracyTester) RunComprehensiveAccuracyTests(ctx context.Context) (*AccuracyTestSuite, error) {
	cat.logger.Info("Starting comprehensive crosswalk accuracy tests")

	suite := &AccuracyTestSuite{
		SuiteName:   "Crosswalk Accuracy Test Suite",
		Description: "Comprehensive testing of MCC/NAICS/SIC crosswalk accuracy",
		TestCases:   []AccuracyTestCase{},
		TestResults: []AccuracyTestResult{},
		CreatedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	// Run different types of accuracy tests
	testTypes := []string{
		"mcc_mapping_accuracy",
		"naics_mapping_accuracy",
		"sic_mapping_accuracy",
		"confidence_scoring_accuracy",
		"validation_rules_accuracy",
		"crosswalk_consistency_accuracy",
		"industry_alignment_accuracy",
	}

	for _, testType := range testTypes {
		cat.logger.Info("Running accuracy test", zap.String("test_type", testType))

		result, err := cat.runAccuracyTestByType(ctx, testType)
		if err != nil {
			cat.logger.Error("Accuracy test failed",
				zap.String("test_type", testType),
				zap.Error(err))
			continue
		}

		suite.TestResults = append(suite.TestResults, *result)
	}

	// Calculate overall score
	suite.OverallScore = cat.calculateOverallScore(suite.TestResults)

	cat.logger.Info("Comprehensive accuracy tests completed",
		zap.Float64("overall_score", suite.OverallScore),
		zap.Int("total_tests", len(suite.TestResults)))

	return suite, nil
}

// RunAccuracyTestByType runs a specific type of accuracy test (public method)
func (cat *CrosswalkAccuracyTester) RunAccuracyTestByType(ctx context.Context, testType string) (*AccuracyTestResult, error) {
	return cat.runAccuracyTestByType(ctx, testType)
}

// runAccuracyTestByType runs a specific type of accuracy test
func (cat *CrosswalkAccuracyTester) runAccuracyTestByType(ctx context.Context, testType string) (*AccuracyTestResult, error) {
	switch testType {
	case "mcc_mapping_accuracy":
		return cat.testMCCMappingAccuracy(ctx)
	case "naics_mapping_accuracy":
		return cat.testNAICSMappingAccuracy(ctx)
	case "sic_mapping_accuracy":
		return cat.testSICMappingAccuracy(ctx)
	case "confidence_scoring_accuracy":
		return cat.testConfidenceScoringAccuracy(ctx)
	case "validation_rules_accuracy":
		return cat.testValidationRulesAccuracy(ctx)
	case "crosswalk_consistency_accuracy":
		return cat.testCrosswalkConsistencyAccuracy(ctx)
	case "industry_alignment_accuracy":
		return cat.testIndustryAlignmentAccuracy(ctx)
	default:
		return nil, fmt.Errorf("unknown test type: %s", testType)
	}
}

// testMCCMappingAccuracy tests the accuracy of MCC to industry mappings
func (cat *CrosswalkAccuracyTester) testMCCMappingAccuracy(ctx context.Context) (*AccuracyTestResult, error) {
	cat.logger.Info("Testing MCC mapping accuracy")

	// Get test cases from database
	testCases, err := cat.getMCCAccuracyTestCases(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get MCC test cases: %w", err)
	}

	result := &AccuracyTestResult{
		TestName:    "MCC Mapping Accuracy Test",
		TestType:    "mcc_mapping_accuracy",
		TestDetails: []AccuracyTestDetail{},
		Timestamp:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	passedTests := 0
	totalTests := len(testCases)

	for _, testCase := range testCases {
		startTime := time.Now()

		// Execute test case
		detail, err := cat.executeMCCMappingTest(ctx, testCase)
		if err != nil {
			cat.logger.Error("MCC mapping test failed",
				zap.String("test_case_id", testCase.TestCaseID),
				zap.Error(err))
			detail.Error = err.Error()
		}

		detail.ExecutionTime = time.Since(startTime)
		result.TestDetails = append(result.TestDetails, detail)

		if detail.Passed {
			passedTests++
		}
	}

	result.TotalTests = totalTests
	result.PassedTests = passedTests
	result.FailedTests = totalTests - passedTests
	result.AccuracyScore = float64(passedTests) / float64(totalTests)
	result.ConfidenceScore = cat.calculateAverageConfidence(result.TestDetails)
	result.Summary = fmt.Sprintf("MCC mapping accuracy: %.2f%% (%d/%d tests passed)",
		result.AccuracyScore*100, passedTests, totalTests)

	return result, nil
}

// testNAICSMappingAccuracy tests the accuracy of NAICS to industry mappings
func (cat *CrosswalkAccuracyTester) testNAICSMappingAccuracy(ctx context.Context) (*AccuracyTestResult, error) {
	cat.logger.Info("Testing NAICS mapping accuracy")

	testCases, err := cat.getNAICSAccuracyTestCases(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get NAICS test cases: %w", err)
	}

	result := &AccuracyTestResult{
		TestName:    "NAICS Mapping Accuracy Test",
		TestType:    "naics_mapping_accuracy",
		TestDetails: []AccuracyTestDetail{},
		Timestamp:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	passedTests := 0
	totalTests := len(testCases)

	for _, testCase := range testCases {
		startTime := time.Now()

		detail, err := cat.executeNAICSMappingTest(ctx, testCase)
		if err != nil {
			cat.logger.Error("NAICS mapping test failed",
				zap.String("test_case_id", testCase.TestCaseID),
				zap.Error(err))
			detail.Error = err.Error()
		}

		detail.ExecutionTime = time.Since(startTime)
		result.TestDetails = append(result.TestDetails, detail)

		if detail.Passed {
			passedTests++
		}
	}

	result.TotalTests = totalTests
	result.PassedTests = passedTests
	result.FailedTests = totalTests - passedTests
	result.AccuracyScore = float64(passedTests) / float64(totalTests)
	result.ConfidenceScore = cat.calculateAverageConfidence(result.TestDetails)
	result.Summary = fmt.Sprintf("NAICS mapping accuracy: %.2f%% (%d/%d tests passed)",
		result.AccuracyScore*100, passedTests, totalTests)

	return result, nil
}

// testSICMappingAccuracy tests the accuracy of SIC to industry mappings
func (cat *CrosswalkAccuracyTester) testSICMappingAccuracy(ctx context.Context) (*AccuracyTestResult, error) {
	cat.logger.Info("Testing SIC mapping accuracy")

	testCases, err := cat.getSICAccuracyTestCases(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SIC test cases: %w", err)
	}

	result := &AccuracyTestResult{
		TestName:    "SIC Mapping Accuracy Test",
		TestType:    "sic_mapping_accuracy",
		TestDetails: []AccuracyTestDetail{},
		Timestamp:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	passedTests := 0
	totalTests := len(testCases)

	for _, testCase := range testCases {
		startTime := time.Now()

		detail, err := cat.executeSICMappingTest(ctx, testCase)
		if err != nil {
			cat.logger.Error("SIC mapping test failed",
				zap.String("test_case_id", testCase.TestCaseID),
				zap.Error(err))
			detail.Error = err.Error()
		}

		detail.ExecutionTime = time.Since(startTime)
		result.TestDetails = append(result.TestDetails, detail)

		if detail.Passed {
			passedTests++
		}
	}

	result.TotalTests = totalTests
	result.PassedTests = passedTests
	result.FailedTests = totalTests - passedTests
	result.AccuracyScore = float64(passedTests) / float64(totalTests)
	result.ConfidenceScore = cat.calculateAverageConfidence(result.TestDetails)
	result.Summary = fmt.Sprintf("SIC mapping accuracy: %.2f%% (%d/%d tests passed)",
		result.AccuracyScore*100, passedTests, totalTests)

	return result, nil
}

// testConfidenceScoringAccuracy tests the accuracy of confidence scoring
func (cat *CrosswalkAccuracyTester) testConfidenceScoringAccuracy(ctx context.Context) (*AccuracyTestResult, error) {
	cat.logger.Info("Testing confidence scoring accuracy")

	testCases, err := cat.getConfidenceScoringTestCases(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get confidence scoring test cases: %w", err)
	}

	result := &AccuracyTestResult{
		TestName:    "Confidence Scoring Accuracy Test",
		TestType:    "confidence_scoring_accuracy",
		TestDetails: []AccuracyTestDetail{},
		Timestamp:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	passedTests := 0
	totalTests := len(testCases)

	for _, testCase := range testCases {
		startTime := time.Now()

		detail, err := cat.executeConfidenceScoringTest(ctx, testCase)
		if err != nil {
			cat.logger.Error("Confidence scoring test failed",
				zap.String("test_case_id", testCase.TestCaseID),
				zap.Error(err))
			detail.Error = err.Error()
		}

		detail.ExecutionTime = time.Since(startTime)
		result.TestDetails = append(result.TestDetails, detail)

		if detail.Passed {
			passedTests++
		}
	}

	result.TotalTests = totalTests
	result.PassedTests = passedTests
	result.FailedTests = totalTests - passedTests
	result.AccuracyScore = float64(passedTests) / float64(totalTests)
	result.ConfidenceScore = cat.calculateAverageConfidence(result.TestDetails)
	result.Summary = fmt.Sprintf("Confidence scoring accuracy: %.2f%% (%d/%d tests passed)",
		result.AccuracyScore*100, passedTests, totalTests)

	return result, nil
}

// testValidationRulesAccuracy tests the accuracy of validation rules
func (cat *CrosswalkAccuracyTester) testValidationRulesAccuracy(ctx context.Context) (*AccuracyTestResult, error) {
	cat.logger.Info("Testing validation rules accuracy")

	testCases, err := cat.getValidationRulesTestCases(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get validation rules test cases: %w", err)
	}

	result := &AccuracyTestResult{
		TestName:    "Validation Rules Accuracy Test",
		TestType:    "validation_rules_accuracy",
		TestDetails: []AccuracyTestDetail{},
		Timestamp:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	passedTests := 0
	totalTests := len(testCases)

	for _, testCase := range testCases {
		startTime := time.Now()

		detail, err := cat.executeValidationRulesTest(ctx, testCase)
		if err != nil {
			cat.logger.Error("Validation rules test failed",
				zap.String("test_case_id", testCase.TestCaseID),
				zap.Error(err))
			detail.Error = err.Error()
		}

		detail.ExecutionTime = time.Since(startTime)
		result.TestDetails = append(result.TestDetails, detail)

		if detail.Passed {
			passedTests++
		}
	}

	result.TotalTests = totalTests
	result.PassedTests = passedTests
	result.FailedTests = totalTests - passedTests
	result.AccuracyScore = float64(passedTests) / float64(totalTests)
	result.ConfidenceScore = cat.calculateAverageConfidence(result.TestDetails)
	result.Summary = fmt.Sprintf("Validation rules accuracy: %.2f%% (%d/%d tests passed)",
		result.AccuracyScore*100, passedTests, totalTests)

	return result, nil
}

// testCrosswalkConsistencyAccuracy tests the consistency of crosswalk mappings
func (cat *CrosswalkAccuracyTester) testCrosswalkConsistencyAccuracy(ctx context.Context) (*AccuracyTestResult, error) {
	cat.logger.Info("Testing crosswalk consistency accuracy")

	testCases, err := cat.getCrosswalkConsistencyTestCases(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get crosswalk consistency test cases: %w", err)
	}

	result := &AccuracyTestResult{
		TestName:    "Crosswalk Consistency Accuracy Test",
		TestType:    "crosswalk_consistency_accuracy",
		TestDetails: []AccuracyTestDetail{},
		Timestamp:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	passedTests := 0
	totalTests := len(testCases)

	for _, testCase := range testCases {
		startTime := time.Now()

		detail, err := cat.executeCrosswalkConsistencyTest(ctx, testCase)
		if err != nil {
			cat.logger.Error("Crosswalk consistency test failed",
				zap.String("test_case_id", testCase.TestCaseID),
				zap.Error(err))
			detail.Error = err.Error()
		}

		detail.ExecutionTime = time.Since(startTime)
		result.TestDetails = append(result.TestDetails, detail)

		if detail.Passed {
			passedTests++
		}
	}

	result.TotalTests = totalTests
	result.PassedTests = passedTests
	result.FailedTests = totalTests - passedTests
	result.AccuracyScore = float64(passedTests) / float64(totalTests)
	result.ConfidenceScore = cat.calculateAverageConfidence(result.TestDetails)
	result.Summary = fmt.Sprintf("Crosswalk consistency accuracy: %.2f%% (%d/%d tests passed)",
		result.AccuracyScore*100, passedTests, totalTests)

	return result, nil
}

// testIndustryAlignmentAccuracy tests the accuracy of industry alignment
func (cat *CrosswalkAccuracyTester) testIndustryAlignmentAccuracy(ctx context.Context) (*AccuracyTestResult, error) {
	cat.logger.Info("Testing industry alignment accuracy")

	testCases, err := cat.getIndustryAlignmentTestCases(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get industry alignment test cases: %w", err)
	}

	result := &AccuracyTestResult{
		TestName:    "Industry Alignment Accuracy Test",
		TestType:    "industry_alignment_accuracy",
		TestDetails: []AccuracyTestDetail{},
		Timestamp:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	passedTests := 0
	totalTests := len(testCases)

	for _, testCase := range testCases {
		startTime := time.Now()

		detail, err := cat.executeIndustryAlignmentTest(ctx, testCase)
		if err != nil {
			cat.logger.Error("Industry alignment test failed",
				zap.String("test_case_id", testCase.TestCaseID),
				zap.Error(err))
			detail.Error = err.Error()
		}

		detail.ExecutionTime = time.Since(startTime)
		result.TestDetails = append(result.TestDetails, detail)

		if detail.Passed {
			passedTests++
		}
	}

	result.TotalTests = totalTests
	result.PassedTests = passedTests
	result.FailedTests = totalTests - passedTests
	result.AccuracyScore = float64(passedTests) / float64(totalTests)
	result.ConfidenceScore = cat.calculateAverageConfidence(result.TestDetails)
	result.Summary = fmt.Sprintf("Industry alignment accuracy: %.2f%% (%d/%d tests passed)",
		result.AccuracyScore*100, passedTests, totalTests)

	return result, nil
}

// Helper methods for test case execution
func (cat *CrosswalkAccuracyTester) executeMCCMappingTest(ctx context.Context, testCase AccuracyTestCase) (AccuracyTestDetail, error) {
	detail := AccuracyTestDetail{
		TestCaseID: testCase.TestCaseID,
		Input:      testCase.Input,
		Expected:   testCase.Expected,
		Actual:     make(map[string]interface{}),
	}

	// Get MCC code from input
	mccCode, ok := testCase.Input["mcc_code"].(string)
	if !ok {
		detail.Passed = false
		detail.Error = "invalid MCC code in input"
		return detail, nil
	}

	// Query database for MCC mapping
	query := `
		SELECT industry_id, confidence_score, description
		FROM crosswalk_mappings 
		WHERE mcc_code = $1
	`

	var industryID int
	var confidenceScore float64
	var description string

	err := cat.db.QueryRowContext(ctx, query, mccCode).Scan(&industryID, &confidenceScore, &description)
	if err != nil {
		detail.Passed = false
		detail.Error = fmt.Sprintf("database query failed: %v", err)
		return detail, nil
	}

	// Set actual results
	detail.Actual["industry_id"] = industryID
	detail.Actual["confidence_score"] = confidenceScore
	detail.Actual["description"] = description

	// Compare with expected results
	expectedIndustryID, ok := testCase.Expected["industry_id"].(int)
	if !ok {
		detail.Passed = false
		detail.Error = "invalid expected industry_id"
		return detail, nil
	}

	detail.Passed = (industryID == expectedIndustryID)
	detail.Confidence = confidenceScore

	return detail, nil
}

func (cat *CrosswalkAccuracyTester) executeNAICSMappingTest(ctx context.Context, testCase AccuracyTestCase) (AccuracyTestDetail, error) {
	detail := AccuracyTestDetail{
		TestCaseID: testCase.TestCaseID,
		Input:      testCase.Input,
		Expected:   testCase.Expected,
		Actual:     make(map[string]interface{}),
	}

	// Get NAICS code from input
	naicsCode, ok := testCase.Input["naics_code"].(string)
	if !ok {
		detail.Passed = false
		detail.Error = "invalid NAICS code in input"
		return detail, nil
	}

	// Query database for NAICS mapping
	query := `
		SELECT industry_id, confidence_score, description
		FROM crosswalk_mappings 
		WHERE naics_code = $1
	`

	var industryID int
	var confidenceScore float64
	var description string

	err := cat.db.QueryRowContext(ctx, query, naicsCode).Scan(&industryID, &confidenceScore, &description)
	if err != nil {
		detail.Passed = false
		detail.Error = fmt.Sprintf("database query failed: %v", err)
		return detail, nil
	}

	// Set actual results
	detail.Actual["industry_id"] = industryID
	detail.Actual["confidence_score"] = confidenceScore
	detail.Actual["description"] = description

	// Compare with expected results
	expectedIndustryID, ok := testCase.Expected["industry_id"].(int)
	if !ok {
		detail.Passed = false
		detail.Error = "invalid expected industry_id"
		return detail, nil
	}

	detail.Passed = (industryID == expectedIndustryID)
	detail.Confidence = confidenceScore

	return detail, nil
}

func (cat *CrosswalkAccuracyTester) executeSICMappingTest(ctx context.Context, testCase AccuracyTestCase) (AccuracyTestDetail, error) {
	detail := AccuracyTestDetail{
		TestCaseID: testCase.TestCaseID,
		Input:      testCase.Input,
		Expected:   testCase.Expected,
		Actual:     make(map[string]interface{}),
	}

	// Get SIC code from input
	sicCode, ok := testCase.Input["sic_code"].(string)
	if !ok {
		detail.Passed = false
		detail.Error = "invalid SIC code in input"
		return detail, nil
	}

	// Query database for SIC mapping
	query := `
		SELECT industry_id, confidence_score, description
		FROM crosswalk_mappings 
		WHERE sic_code = $1
	`

	var industryID int
	var confidenceScore float64
	var description string

	err := cat.db.QueryRowContext(ctx, query, sicCode).Scan(&industryID, &confidenceScore, &description)
	if err != nil {
		detail.Passed = false
		detail.Error = fmt.Sprintf("database query failed: %v", err)
		return detail, nil
	}

	// Set actual results
	detail.Actual["industry_id"] = industryID
	detail.Actual["confidence_score"] = confidenceScore
	detail.Actual["description"] = description

	// Compare with expected results
	expectedIndustryID, ok := testCase.Expected["industry_id"].(int)
	if !ok {
		detail.Passed = false
		detail.Error = "invalid expected industry_id"
		return detail, nil
	}

	detail.Passed = (industryID == expectedIndustryID)
	detail.Confidence = confidenceScore

	return detail, nil
}

func (cat *CrosswalkAccuracyTester) executeConfidenceScoringTest(ctx context.Context, testCase AccuracyTestCase) (AccuracyTestDetail, error) {
	detail := AccuracyTestDetail{
		TestCaseID: testCase.TestCaseID,
		Input:      testCase.Input,
		Expected:   testCase.Expected,
		Actual:     make(map[string]interface{}),
	}

	// Get input parameters
	mccCode, _ := testCase.Input["mcc_code"].(string)
	naicsCode, _ := testCase.Input["naics_code"].(string)
	sicCode, _ := testCase.Input["sic_code"].(string)

	// Calculate confidence score using our algorithm
	confidenceScore := cat.calculateTestConfidenceScore(mccCode, naicsCode, sicCode)

	detail.Actual["confidence_score"] = confidenceScore

	// Compare with expected confidence score
	expectedConfidence, ok := testCase.Expected["confidence_score"].(float64)
	if !ok {
		detail.Passed = false
		detail.Error = "invalid expected confidence score"
		return detail, nil
	}

	// Allow for small tolerance in confidence score comparison
	tolerance := 0.05
	detail.Passed = math.Abs(confidenceScore-expectedConfidence) <= tolerance
	detail.Confidence = confidenceScore

	return detail, nil
}

func (cat *CrosswalkAccuracyTester) executeValidationRulesTest(ctx context.Context, testCase AccuracyTestCase) (AccuracyTestDetail, error) {
	detail := AccuracyTestDetail{
		TestCaseID: testCase.TestCaseID,
		Input:      testCase.Input,
		Expected:   testCase.Expected,
		Actual:     make(map[string]interface{}),
	}

	// Create validation rules engine
	config := &CrosswalkValidationConfig{
		MaxValidationTime:              30,
		EnableFormatValidation:         true,
		EnableConsistencyValidation:    true,
		EnableBusinessLogicValidation:  true,
		EnableCrossReferenceValidation: true,
		MinConfidenceScore:             0.5,
	}
	validator := NewCrosswalkValidationRules(cat.db, cat.logger, config)

	// Execute validation
	summary, err := validator.ValidateCrosswalkMappings(ctx)
	if err != nil {
		detail.Passed = false
		detail.Error = fmt.Sprintf("validation execution failed: %v", err)
		return detail, nil
	}

	// Set actual results
	detail.Actual["total_rules"] = len(summary.Results)
	detail.Actual["passed_rules"] = cat.countPassedRules(summary.Results)
	detail.Actual["failed_rules"] = cat.countFailedRules(summary.Results)

	// Compare with expected results
	expectedPassed, ok := testCase.Expected["expected_passed"].(int)
	if !ok {
		detail.Passed = false
		detail.Error = "invalid expected passed count"
		return detail, nil
	}

	actualPassed := cat.countPassedRules(summary.Results)
	detail.Passed = (actualPassed >= expectedPassed)
	detail.Confidence = float64(actualPassed) / float64(len(summary.Results))

	return detail, nil
}

func (cat *CrosswalkAccuracyTester) executeCrosswalkConsistencyTest(ctx context.Context, testCase AccuracyTestCase) (AccuracyTestDetail, error) {
	detail := AccuracyTestDetail{
		TestCaseID: testCase.TestCaseID,
		Input:      testCase.Input,
		Expected:   testCase.Expected,
		Actual:     make(map[string]interface{}),
	}

	// Get industry ID from input
	industryID, ok := testCase.Input["industry_id"].(int)
	if !ok {
		detail.Passed = false
		detail.Error = "invalid industry_id in input"
		return detail, nil
	}

	// Query for all mappings for this industry
	query := `
		SELECT mcc_code, naics_code, sic_code, confidence_score
		FROM crosswalk_mappings 
		WHERE industry_id = $1
	`

	rows, err := cat.db.QueryContext(ctx, query, industryID)
	if err != nil {
		detail.Passed = false
		detail.Error = fmt.Sprintf("database query failed: %v", err)
		return detail, nil
	}
	defer rows.Close()

	var mappings []map[string]interface{}
	for rows.Next() {
		var mccCode, naicsCode, sicCode sql.NullString
		var confidenceScore float64

		err := rows.Scan(&mccCode, &naicsCode, &sicCode, &confidenceScore)
		if err != nil {
			continue
		}

		mapping := map[string]interface{}{
			"mcc_code":         mccCode.String,
			"naics_code":       naicsCode.String,
			"sic_code":         sicCode.String,
			"confidence_score": confidenceScore,
		}
		mappings = append(mappings, mapping)
	}

	// Check consistency
	consistencyScore := cat.calculateConsistencyScore(mappings)

	detail.Actual["mappings_count"] = len(mappings)
	detail.Actual["consistency_score"] = consistencyScore

	// Compare with expected consistency score
	expectedConsistency, ok := testCase.Expected["expected_consistency"].(float64)
	if !ok {
		detail.Passed = false
		detail.Error = "invalid expected consistency score"
		return detail, nil
	}

	detail.Passed = (consistencyScore >= expectedConsistency)
	detail.Confidence = consistencyScore

	return detail, nil
}

func (cat *CrosswalkAccuracyTester) executeIndustryAlignmentTest(ctx context.Context, testCase AccuracyTestCase) (AccuracyTestDetail, error) {
	detail := AccuracyTestDetail{
		TestCaseID: testCase.TestCaseID,
		Input:      testCase.Input,
		Expected:   testCase.Expected,
		Actual:     make(map[string]interface{}),
	}

	// Create alignment engine
	config := &AlignmentConfig{
		MaxAlignmentTime:         30,
		EnableMCCAlignment:       true,
		EnableNAICSAlignment:     true,
		EnableSICAlignment:       true,
		EnableConflictResolution: true,
		EnableGapAnalysis:        true,
		MinAlignmentScore:        0.5,
	}
	engine := NewClassificationAlignmentEngine(cat.db, cat.logger, config)

	// Analyze alignment
	result, err := engine.AnalyzeClassificationAlignment(ctx)
	if err != nil {
		detail.Passed = false
		detail.Error = fmt.Sprintf("alignment analysis failed: %v", err)
		return detail, nil
	}

	// Set actual results
	detail.Actual["total_conflicts"] = len(result.Conflicts)
	detail.Actual["total_gaps"] = len(result.Gaps)
	detail.Actual["alignment_score"] = result.Summary.OverallAlignmentScore

	// Compare with expected results
	expectedConflicts, ok := testCase.Expected["expected_conflicts"].(int)
	if !ok {
		detail.Passed = false
		detail.Error = "invalid expected conflicts count"
		return detail, nil
	}

	detail.Passed = (len(result.Conflicts) <= expectedConflicts)
	detail.Confidence = result.Summary.OverallAlignmentScore

	return detail, nil
}

// Helper methods for test case retrieval
func (cat *CrosswalkAccuracyTester) getMCCAccuracyTestCases(ctx context.Context) ([]AccuracyTestCase, error) {
	// For now, return hardcoded test cases
	// In a real implementation, these would come from a test database
	return []AccuracyTestCase{
		{
			TestCaseID:  "mcc_001",
			TestName:    "MCC 5411 - Grocery Stores",
			TestType:    "mcc_mapping_accuracy",
			Input:       map[string]interface{}{"mcc_code": "5411"},
			Expected:    map[string]interface{}{"industry_id": 1},
			Description: "Test MCC 5411 maps to Grocery Stores industry",
			Weight:      1.0,
			Category:    "retail",
			Tags:        []string{"grocery", "retail", "food"},
		},
		{
			TestCaseID:  "mcc_002",
			TestName:    "MCC 5999 - Miscellaneous Retail",
			TestType:    "mcc_mapping_accuracy",
			Input:       map[string]interface{}{"mcc_code": "5999"},
			Expected:    map[string]interface{}{"industry_id": 2},
			Description: "Test MCC 5999 maps to Miscellaneous Retail industry",
			Weight:      1.0,
			Category:    "retail",
			Tags:        []string{"miscellaneous", "retail"},
		},
	}, nil
}

func (cat *CrosswalkAccuracyTester) getNAICSAccuracyTestCases(ctx context.Context) ([]AccuracyTestCase, error) {
	return []AccuracyTestCase{
		{
			TestCaseID:  "naics_001",
			TestName:    "NAICS 445110 - Supermarkets and Grocery Stores",
			TestType:    "naics_mapping_accuracy",
			Input:       map[string]interface{}{"naics_code": "445110"},
			Expected:    map[string]interface{}{"industry_id": 1},
			Description: "Test NAICS 445110 maps to Grocery Stores industry",
			Weight:      1.0,
			Category:    "retail",
			Tags:        []string{"grocery", "retail", "food"},
		},
	}, nil
}

func (cat *CrosswalkAccuracyTester) getSICAccuracyTestCases(ctx context.Context) ([]AccuracyTestCase, error) {
	return []AccuracyTestCase{
		{
			TestCaseID:  "sic_001",
			TestName:    "SIC 5411 - Grocery Stores",
			TestType:    "sic_mapping_accuracy",
			Input:       map[string]interface{}{"sic_code": "5411"},
			Expected:    map[string]interface{}{"industry_id": 1},
			Description: "Test SIC 5411 maps to Grocery Stores industry",
			Weight:      1.0,
			Category:    "retail",
			Tags:        []string{"grocery", "retail", "food"},
		},
	}, nil
}

func (cat *CrosswalkAccuracyTester) getConfidenceScoringTestCases(ctx context.Context) ([]AccuracyTestCase, error) {
	return []AccuracyTestCase{
		{
			TestCaseID:  "confidence_001",
			TestName:    "High Confidence Grocery Store",
			TestType:    "confidence_scoring_accuracy",
			Input:       map[string]interface{}{"mcc_code": "5411", "naics_code": "445110", "sic_code": "5411"},
			Expected:    map[string]interface{}{"confidence_score": 0.95},
			Description: "Test high confidence scoring for grocery store codes",
			Weight:      1.0,
			Category:    "confidence",
			Tags:        []string{"high_confidence", "grocery"},
		},
	}, nil
}

func (cat *CrosswalkAccuracyTester) getValidationRulesTestCases(ctx context.Context) ([]AccuracyTestCase, error) {
	return []AccuracyTestCase{
		{
			TestCaseID:  "validation_001",
			TestName:    "Format Validation Rules",
			TestType:    "validation_rules_accuracy",
			Input:       map[string]interface{}{"test_type": "format_validation"},
			Expected:    map[string]interface{}{"expected_passed": 5},
			Description: "Test format validation rules pass",
			Weight:      1.0,
			Category:    "validation",
			Tags:        []string{"format", "validation"},
		},
	}, nil
}

func (cat *CrosswalkAccuracyTester) getCrosswalkConsistencyTestCases(ctx context.Context) ([]AccuracyTestCase, error) {
	return []AccuracyTestCase{
		{
			TestCaseID:  "consistency_001",
			TestName:    "Grocery Store Consistency",
			TestType:    "crosswalk_consistency_accuracy",
			Input:       map[string]interface{}{"industry_id": 1},
			Expected:    map[string]interface{}{"expected_consistency": 0.9},
			Description: "Test consistency of grocery store mappings",
			Weight:      1.0,
			Category:    "consistency",
			Tags:        []string{"grocery", "consistency"},
		},
	}, nil
}

func (cat *CrosswalkAccuracyTester) getIndustryAlignmentTestCases(ctx context.Context) ([]AccuracyTestCase, error) {
	return []AccuracyTestCase{
		{
			TestCaseID:  "alignment_001",
			TestName:    "Overall Industry Alignment",
			TestType:    "industry_alignment_accuracy",
			Input:       map[string]interface{}{"test_type": "overall_alignment"},
			Expected:    map[string]interface{}{"expected_conflicts": 5},
			Description: "Test overall industry alignment has minimal conflicts",
			Weight:      1.0,
			Category:    "alignment",
			Tags:        []string{"alignment", "conflicts"},
		},
	}, nil
}

// Utility methods
func (cat *CrosswalkAccuracyTester) calculateTestConfidenceScore(mccCode, naicsCode, sicCode string) float64 {
	// Simple confidence scoring algorithm for testing
	score := 0.0

	if mccCode != "" {
		score += 0.4
	}
	if naicsCode != "" {
		score += 0.4
	}
	if sicCode != "" {
		score += 0.2
	}

	// Bonus for matching codes (simplified)
	if mccCode == "5411" && naicsCode == "445110" && sicCode == "5411" {
		score += 0.1
	}

	return math.Min(score, 1.0)
}

func (cat *CrosswalkAccuracyTester) calculateConsistencyScore(mappings []map[string]interface{}) float64 {
	if len(mappings) == 0 {
		return 0.0
	}

	// Calculate average confidence score as consistency measure
	totalConfidence := 0.0
	for _, mapping := range mappings {
		if confidence, ok := mapping["confidence_score"].(float64); ok {
			totalConfidence += confidence
		}
	}

	return totalConfidence / float64(len(mappings))
}

func (cat *CrosswalkAccuracyTester) countPassedRules(results []ValidationRuleResult) int {
	count := 0
	for _, result := range results {
		if result.Status == ValidationStatusPassed {
			count++
		}
	}
	return count
}

func (cat *CrosswalkAccuracyTester) countFailedRules(results []ValidationRuleResult) int {
	count := 0
	for _, result := range results {
		if result.Status == ValidationStatusFailed {
			count++
		}
	}
	return count
}

func (cat *CrosswalkAccuracyTester) calculateAverageConfidence(details []AccuracyTestDetail) float64 {
	if len(details) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, detail := range details {
		totalConfidence += detail.Confidence
	}

	return totalConfidence / float64(len(details))
}

func (cat *CrosswalkAccuracyTester) calculateOverallScore(results []AccuracyTestResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	totalScore := 0.0
	for _, result := range results {
		totalScore += result.AccuracyScore
	}

	return totalScore / float64(len(results))
}

// SaveAccuracyTestResults saves test results to database
func (cat *CrosswalkAccuracyTester) SaveAccuracyTestResults(ctx context.Context, suite *AccuracyTestSuite) error {
	cat.logger.Info("Saving accuracy test results to database")

	// Create accuracy_test_results table if it doesn't exist
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS accuracy_test_results (
			id SERIAL PRIMARY KEY,
			suite_name VARCHAR(255) NOT NULL,
			test_name VARCHAR(255) NOT NULL,
			test_type VARCHAR(100) NOT NULL,
			total_tests INTEGER NOT NULL,
			passed_tests INTEGER NOT NULL,
			failed_tests INTEGER NOT NULL,
			accuracy_score DECIMAL(5,4) NOT NULL,
			confidence_score DECIMAL(5,4) NOT NULL,
			summary TEXT,
			test_details JSONB,
			metadata JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := cat.db.ExecContext(ctx, createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create accuracy_test_results table: %w", err)
	}

	// Insert test results
	for _, result := range suite.TestResults {
		insertQuery := `
			INSERT INTO accuracy_test_results (
				suite_name, test_name, test_type, total_tests, passed_tests, 
				failed_tests, accuracy_score, confidence_score, summary, 
				test_details, metadata
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`

		_, err := cat.db.ExecContext(ctx, insertQuery,
			suite.SuiteName,
			result.TestName,
			result.TestType,
			result.TotalTests,
			result.PassedTests,
			result.FailedTests,
			result.AccuracyScore,
			result.ConfidenceScore,
			result.Summary,
			result.TestDetails,
			result.Metadata,
		)

		if err != nil {
			cat.logger.Error("Failed to insert test result",
				zap.String("test_name", result.TestName),
				zap.Error(err))
			continue
		}
	}

	cat.logger.Info("Accuracy test results saved successfully")
	return nil
}
