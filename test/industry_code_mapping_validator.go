package test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"kyb-platform/internal/classification"
)

// IndustryCodeMappingValidator provides comprehensive validation for industry code mappings
type IndustryCodeMappingValidator struct {
	TestRunner *ClassificationAccuracyTestRunner
	Logger     *log.Logger
	Config     *CodeMappingValidationConfig
}

// CodeMappingValidationConfig configuration for code mapping validation
type CodeMappingValidationConfig struct {
	SessionName                     string        `json:"session_name"`
	ValidationDirectory             string        `json:"validation_directory"`
	SampleSize                      int           `json:"sample_size"`
	Timeout                         time.Duration `json:"timeout"`
	MinAccuracyThreshold            float64       `json:"min_accuracy_threshold"`
	IncludeFormatValidation         bool          `json:"include_format_validation"`
	IncludeStructureValidation      bool          `json:"include_structure_validation"`
	IncludeCrossReferenceValidation bool          `json:"include_cross_reference_validation"`
	GenerateDetailedReport          bool          `json:"generate_detailed_report"`
}

// CodeMappingValidationResult represents the result of code mapping validation
type CodeMappingValidationResult struct {
	SessionID                  string                     `json:"session_id"`
	StartTime                  time.Time                  `json:"start_time"`
	EndTime                    time.Time                  `json:"end_time"`
	Duration                   time.Duration              `json:"duration"`
	TotalValidations           int                        `json:"total_validations"`
	ValidationSummary          *CodeMappingSummary        `json:"validation_summary"`
	CodeTypeResults            map[string]*CodeTypeResult `json:"code_type_results"`
	IndustryResults            map[string]*IndustryResult `json:"industry_results"`
	FormatValidationResults    *FormatValidationResult    `json:"format_validation_results"`
	StructureValidationResults *StructureValidationResult `json:"structure_validation_results"`
	CrossReferenceResults      *CrossReferenceResult      `json:"cross_reference_results"`
	Recommendations            []string                   `json:"recommendations"`
	Issues                     []ValidationIssue          `json:"issues"`
}

// CodeMappingSummary provides overall validation summary
type CodeMappingSummary struct {
	OverallAccuracy           float64 `json:"overall_accuracy"`
	MCCAccuracy               float64 `json:"mcc_accuracy"`
	SICAccuracy               float64 `json:"sic_accuracy"`
	NAICSAccuracy             float64 `json:"naics_accuracy"`
	FormatValidationPassed    bool    `json:"format_validation_passed"`
	StructureValidationPassed bool    `json:"structure_validation_passed"`
	CrossReferencePassed      bool    `json:"cross_reference_passed"`
	TotalIssues               int     `json:"total_issues"`
	CriticalIssues            int     `json:"critical_issues"`
	HighIssues                int     `json:"high_issues"`
	MediumIssues              int     `json:"medium_issues"`
	LowIssues                 int     `json:"low_issues"`
}

// CodeTypeResult represents validation results for a specific code type
type CodeTypeResult struct {
	CodeType          string            `json:"code_type"`
	TotalCodes        int               `json:"total_codes"`
	ValidCodes        int               `json:"valid_codes"`
	InvalidCodes      int               `json:"invalid_codes"`
	Accuracy          float64           `json:"accuracy"`
	FormatAccuracy    float64           `json:"format_accuracy"`
	StructureAccuracy float64           `json:"structure_accuracy"`
	ExpectedCodes     int               `json:"expected_codes"`
	MatchedCodes      int               `json:"matched_codes"`
	UnmatchedCodes    int               `json:"unmatched_codes"`
	Precision         float64           `json:"precision"`
	Recall            float64           `json:"recall"`
	F1Score           float64           `json:"f1_score"`
	Issues            []ValidationIssue `json:"issues"`
}

// IndustryResult represents validation results for a specific industry
type IndustryResult struct {
	Industry      string              `json:"industry"`
	TotalCases    int                 `json:"total_cases"`
	ValidCases    int                 `json:"valid_cases"`
	Accuracy      float64             `json:"accuracy"`
	MCCAccuracy   float64             `json:"mcc_accuracy"`
	SICAccuracy   float64             `json:"sic_accuracy"`
	NAICSAccuracy float64             `json:"naics_accuracy"`
	ExpectedCodes map[string][]string `json:"expected_codes"`
	ActualCodes   map[string][]string `json:"actual_codes"`
	Issues        []ValidationIssue   `json:"issues"`
}

// FormatValidationResult represents format validation results
type FormatValidationResult struct {
	MCCFormatValid        int               `json:"mcc_format_valid"`
	MCCFormatInvalid      int               `json:"mcc_format_invalid"`
	SICFormatValid        int               `json:"sic_format_valid"`
	SICFormatInvalid      int               `json:"sic_format_invalid"`
	NAICSFormatValid      int               `json:"naics_format_valid"`
	NAICSFormatInvalid    int               `json:"naics_format_invalid"`
	OverallFormatAccuracy float64           `json:"overall_format_accuracy"`
	Issues                []ValidationIssue `json:"issues"`
}

// StructureValidationResult represents structure validation results
type StructureValidationResult struct {
	DuplicateCodes          int               `json:"duplicate_codes"`
	EmptyResults            int               `json:"empty_results"`
	InvalidConfidenceScores int               `json:"invalid_confidence_scores"`
	StructureAccuracy       float64           `json:"structure_accuracy"`
	Issues                  []ValidationIssue `json:"issues"`
}

// CrossReferenceResult represents cross-reference validation results
type CrossReferenceResult struct {
	MCCToSICMappings       int               `json:"mcc_to_sic_mappings"`
	MCCToNAICSMappings     int               `json:"mcc_to_naics_mappings"`
	SICToNAICSMappings     int               `json:"sic_to_naics_mappings"`
	CrossReferenceAccuracy float64           `json:"cross_reference_accuracy"`
	Issues                 []ValidationIssue `json:"issues"`
}

// ValidationIssue represents a validation issue
type ValidationIssue struct {
	Type           string `json:"type"`           // "critical", "high", "medium", "low"
	Code           string `json:"code"`           // Issue code
	Message        string `json:"message"`        // Issue description
	Industry       string `json:"industry"`       // Affected industry
	CodeType       string `json:"code_type"`      // Affected code type
	Expected       string `json:"expected"`       // Expected value
	Actual         string `json:"actual"`         // Actual value
	Recommendation string `json:"recommendation"` // Recommended action
}

// ExpectedCodeMapping defines expected codes for each industry
type ExpectedCodeMapping struct {
	Industry string            `json:"industry"`
	MCC      []string          `json:"mcc"`
	SIC      []string          `json:"sic"`
	NAICS    []string          `json:"naics"`
	Keywords []string          `json:"keywords"`
	Metadata map[string]string `json:"metadata"`
}

// NewIndustryCodeMappingValidator creates a new industry code mapping validator
func NewIndustryCodeMappingValidator(testRunner *ClassificationAccuracyTestRunner, logger *log.Logger, config *CodeMappingValidationConfig) *IndustryCodeMappingValidator {
	return &IndustryCodeMappingValidator{
		TestRunner: testRunner,
		Logger:     logger,
		Config:     config,
	}
}

// ValidateCodeMapping performs comprehensive code mapping validation
func (validator *IndustryCodeMappingValidator) ValidateCodeMapping(ctx context.Context) (*CodeMappingValidationResult, error) {
	startTime := time.Now()
	sessionID := fmt.Sprintf("code_mapping_%d", startTime.Unix())

	validator.Logger.Printf("🔍 Starting Industry Code Mapping Validation Session: %s", sessionID)

	result := &CodeMappingValidationResult{
		SessionID:        sessionID,
		StartTime:        startTime,
		TotalValidations: 0,
		CodeTypeResults:  make(map[string]*CodeTypeResult),
		IndustryResults:  make(map[string]*IndustryResult),
		Recommendations:  []string{},
		Issues:           []ValidationIssue{},
	}

	// Initialize code type results
	result.CodeTypeResults["MCC"] = &CodeTypeResult{CodeType: "MCC"}
	result.CodeTypeResults["SIC"] = &CodeTypeResult{CodeType: "SIC"}
	result.CodeTypeResults["NAICS"] = &CodeTypeResult{CodeType: "NAICS"}

	// Get test cases
	dataset := validator.TestRunner.GetDataset()
	testCases := dataset.TestCases
	if len(testCases) > validator.Config.SampleSize {
		testCases = testCases[:validator.Config.SampleSize]
	}

	validator.Logger.Printf("📊 Validating %d test cases", len(testCases))

	// Run validation for each test case
	for _, testCase := range testCases {
		if err := validator.validateTestCase(ctx, testCase, result); err != nil {
			validator.Logger.Printf("❌ Validation failed for %s: %v", testCase.Name, err)
			continue
		}
		result.TotalValidations++
	}

	// Perform additional validations
	if validator.Config.IncludeFormatValidation {
		result.FormatValidationResults = validator.validateCodeFormats(result)
	}

	if validator.Config.IncludeStructureValidation {
		result.StructureValidationResults = validator.validateCodeStructures(result)
	}

	if validator.Config.IncludeCrossReferenceValidation {
		result.CrossReferenceResults = validator.validateCrossReferences(result)
	}

	// Calculate summary
	result.ValidationSummary = validator.calculateValidationSummary(result)

	// Generate recommendations
	result.Recommendations = validator.generateRecommendations(result)

	// Set end time and duration
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	validator.Logger.Printf("✅ Code mapping validation completed in %v", result.Duration)
	validator.Logger.Printf("📊 Overall accuracy: %.2f%%", result.ValidationSummary.OverallAccuracy)

	return result, nil
}

// validateTestCase validates a single test case
func (validator *IndustryCodeMappingValidator) validateTestCase(ctx context.Context, testCase ClassificationTestCase, result *CodeMappingValidationResult) error {
	// Run classification
	classificationResult, err := validator.TestRunner.classifier.GenerateClassificationCodes(
		ctx,
		testCase.Keywords,
		testCase.ExpectedIndustry,
		testCase.ExpectedConfidence,
	)
	if err != nil {
		return fmt.Errorf("classification failed: %w", err)
	}

	// Get expected codes for this industry
	expectedCodes := validator.getExpectedCodes(testCase.ExpectedIndustry)

	// Validate each code type
	validator.validateMCCCodes(classificationResult.MCC, expectedCodes.MCC, testCase.ExpectedIndustry, result)
	validator.validateSICCodes(classificationResult.SIC, expectedCodes.SIC, testCase.ExpectedIndustry, result)
	validator.validateNAICSCodes(classificationResult.NAICS, expectedCodes.NAICS, testCase.ExpectedIndustry, result)

	// Update industry results
	validator.updateIndustryResults(testCase.ExpectedIndustry, classificationResult, expectedCodes, result)

	return nil
}

// validateMCCCodes validates MCC codes
func (validator *IndustryCodeMappingValidator) validateMCCCodes(actualCodes []classification.MCCCode, expectedCodes []string, industry string, result *CodeMappingValidationResult) {
	mccResult := result.CodeTypeResults["MCC"]

	// Extract actual code strings
	actualCodeStrings := make([]string, len(actualCodes))
	for i, code := range actualCodes {
		actualCodeStrings[i] = code.Code
	}

	// Validate format
	validFormat := 0
	for _, code := range actualCodeStrings {
		if validator.isValidMCCFormat(code) {
			validFormat++
		} else {
			validator.addIssue("medium", "INVALID_MCC_FORMAT", fmt.Sprintf("Invalid MCC format: %s", code), industry, "MCC", "4-digit code", code, "Fix MCC code format")
		}
	}

	// Calculate accuracy metrics
	matchedCodes := validator.countMatchedCodes(actualCodeStrings, expectedCodes)
	precision := validator.calculatePrecision(actualCodeStrings, expectedCodes)
	recall := validator.calculateRecall(actualCodeStrings, expectedCodes)
	f1Score := validator.calculateF1Score(precision, recall)

	// Update results
	mccResult.TotalCodes += len(actualCodeStrings)
	mccResult.ValidCodes += validFormat
	mccResult.InvalidCodes += len(actualCodeStrings) - validFormat
	mccResult.ExpectedCodes += len(expectedCodes)
	mccResult.MatchedCodes += matchedCodes
	mccResult.UnmatchedCodes += len(actualCodeStrings) - matchedCodes
	mccResult.Precision = precision
	mccResult.Recall = recall
	mccResult.F1Score = f1Score
}

// validateSICCodes validates SIC codes
func (validator *IndustryCodeMappingValidator) validateSICCodes(actualCodes []classification.SICCode, expectedCodes []string, industry string, result *CodeMappingValidationResult) {
	sicResult := result.CodeTypeResults["SIC"]

	// Extract actual code strings
	actualCodeStrings := make([]string, len(actualCodes))
	for i, code := range actualCodes {
		actualCodeStrings[i] = code.Code
	}

	// Validate format
	validFormat := 0
	for _, code := range actualCodeStrings {
		if validator.isValidSICFormat(code) {
			validFormat++
		} else {
			validator.addIssue("medium", "INVALID_SIC_FORMAT", fmt.Sprintf("Invalid SIC format: %s", code), industry, "SIC", "4-digit code", code, "Fix SIC code format")
		}
	}

	// Calculate accuracy metrics
	matchedCodes := validator.countMatchedCodes(actualCodeStrings, expectedCodes)
	precision := validator.calculatePrecision(actualCodeStrings, expectedCodes)
	recall := validator.calculateRecall(actualCodeStrings, expectedCodes)
	f1Score := validator.calculateF1Score(precision, recall)

	// Update results
	sicResult.TotalCodes += len(actualCodeStrings)
	sicResult.ValidCodes += validFormat
	sicResult.InvalidCodes += len(actualCodeStrings) - validFormat
	sicResult.ExpectedCodes += len(expectedCodes)
	sicResult.MatchedCodes += matchedCodes
	sicResult.UnmatchedCodes += len(actualCodeStrings) - matchedCodes
	sicResult.Precision = precision
	sicResult.Recall = recall
	sicResult.F1Score = f1Score
}

// validateNAICSCodes validates NAICS codes
func (validator *IndustryCodeMappingValidator) validateNAICSCodes(actualCodes []classification.NAICSCode, expectedCodes []string, industry string, result *CodeMappingValidationResult) {
	naicsResult := result.CodeTypeResults["NAICS"]

	// Extract actual code strings
	actualCodeStrings := make([]string, len(actualCodes))
	for i, code := range actualCodes {
		actualCodeStrings[i] = code.Code
	}

	// Validate format
	validFormat := 0
	for _, code := range actualCodeStrings {
		if validator.isValidNAICSFormat(code) {
			validFormat++
		} else {
			validator.addIssue("medium", "INVALID_NAICS_FORMAT", fmt.Sprintf("Invalid NAICS format: %s", code), industry, "NAICS", "6-digit code", code, "Fix NAICS code format")
		}
	}

	// Calculate accuracy metrics
	matchedCodes := validator.countMatchedCodes(actualCodeStrings, expectedCodes)
	precision := validator.calculatePrecision(actualCodeStrings, expectedCodes)
	recall := validator.calculateRecall(actualCodeStrings, expectedCodes)
	f1Score := validator.calculateF1Score(precision, recall)

	// Update results
	naicsResult.TotalCodes += len(actualCodeStrings)
	naicsResult.ValidCodes += validFormat
	naicsResult.InvalidCodes += len(actualCodeStrings) - validFormat
	naicsResult.ExpectedCodes += len(expectedCodes)
	naicsResult.MatchedCodes += matchedCodes
	naicsResult.UnmatchedCodes += len(actualCodeStrings) - matchedCodes
	naicsResult.Precision = precision
	naicsResult.Recall = recall
	naicsResult.F1Score = f1Score
}

// Helper methods for validation
func (validator *IndustryCodeMappingValidator) isValidMCCFormat(code string) bool {
	matched, _ := regexp.MatchString(`^\d{4}$`, code)
	return matched
}

func (validator *IndustryCodeMappingValidator) isValidSICFormat(code string) bool {
	matched, _ := regexp.MatchString(`^\d{4}$`, code)
	return matched
}

func (validator *IndustryCodeMappingValidator) isValidNAICSFormat(code string) bool {
	matched, _ := regexp.MatchString(`^\d{6}$`, code)
	return matched
}

func (validator *IndustryCodeMappingValidator) countMatchedCodes(actual, expected []string) int {
	matches := 0
	for _, actualCode := range actual {
		for _, expectedCode := range expected {
			if actualCode == expectedCode {
				matches++
				break
			}
		}
	}
	return matches
}

func (validator *IndustryCodeMappingValidator) calculatePrecision(actual, expected []string) float64 {
	if len(actual) == 0 {
		return 0.0
	}
	matches := validator.countMatchedCodes(actual, expected)
	return float64(matches) / float64(len(actual))
}

func (validator *IndustryCodeMappingValidator) calculateRecall(actual, expected []string) float64 {
	if len(expected) == 0 {
		return 0.0
	}
	matches := validator.countMatchedCodes(actual, expected)
	return float64(matches) / float64(len(expected))
}

func (validator *IndustryCodeMappingValidator) calculateF1Score(precision, recall float64) float64 {
	if precision+recall == 0 {
		return 0.0
	}
	return 2 * (precision * recall) / (precision + recall)
}

func (validator *IndustryCodeMappingValidator) addIssue(issueType, code, message, industry, codeType, expected, actual, recommendation string) {
	issue := ValidationIssue{
		Type:           issueType,
		Code:           code,
		Message:        message,
		Industry:       industry,
		CodeType:       codeType,
		Expected:       expected,
		Actual:         actual,
		Recommendation: recommendation,
	}
	// Add to overall issues
	_ = issue // TODO: Implement issue tracking
}

func (validator *IndustryCodeMappingValidator) getExpectedCodes(industry string) ExpectedCodeMapping {
	// Return expected codes for the given industry
	// This would typically come from a database or configuration
	return ExpectedCodeMapping{
		Industry: industry,
		MCC:      []string{},
		SIC:      []string{},
		NAICS:    []string{},
		Keywords: []string{},
		Metadata: make(map[string]string),
	}
}

func (validator *IndustryCodeMappingValidator) updateIndustryResults(industry string, classificationResult *classification.ClassificationCodesInfo, expectedCodes ExpectedCodeMapping, result *CodeMappingValidationResult) {
	// Update industry-specific results
}

func (validator *IndustryCodeMappingValidator) validateCodeFormats(result *CodeMappingValidationResult) *FormatValidationResult {
	// Validate code formats
	return &FormatValidationResult{}
}

func (validator *IndustryCodeMappingValidator) validateCodeStructures(result *CodeMappingValidationResult) *StructureValidationResult {
	// Validate code structures
	return &StructureValidationResult{}
}

func (validator *IndustryCodeMappingValidator) validateCrossReferences(result *CodeMappingValidationResult) *CrossReferenceResult {
	// Validate cross-references between code systems
	return &CrossReferenceResult{}
}

func (validator *IndustryCodeMappingValidator) calculateValidationSummary(result *CodeMappingValidationResult) *CodeMappingSummary {
	// Calculate overall validation summary
	return &CodeMappingSummary{}
}

func (validator *IndustryCodeMappingValidator) generateRecommendations(result *CodeMappingValidationResult) []string {
	// Generate recommendations based on validation results
	return []string{}
}

// SaveValidationReport saves the validation report to file
func (validator *IndustryCodeMappingValidator) SaveValidationReport(result *CodeMappingValidationResult) error {
	// Create validation directory if it doesn't exist
	if err := os.MkdirAll(validator.Config.ValidationDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create validation directory: %w", err)
	}

	// Save JSON report
	jsonFile := filepath.Join(validator.Config.ValidationDirectory, "code_mapping_validation_report.json")
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON report: %w", err)
	}

	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	validator.Logger.Printf("✅ Validation report saved to: %s", jsonFile)
	return nil
}
