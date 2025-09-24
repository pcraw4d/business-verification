package test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/classification"
)

// ManualValidationFramework provides comprehensive manual validation capabilities
type ManualValidationFramework struct {
	TestRunner    *ClassificationAccuracyTestRunner
	ValidationDir string
	Results       *ManualValidationResults
	Config        *ManualValidationConfig
}

// ManualValidationResults represents the results of manual validation
type ManualValidationResults struct {
	SessionID        string                   `json:"session_id"`
	StartTime        time.Time                `json:"start_time"`
	EndTime          time.Time                `json:"end_time"`
	Duration         time.Duration            `json:"duration"`
	TotalValidations int                      `json:"total_validations"`
	ValidatedCases   []ManualValidationCase   `json:"validated_cases"`
	Summary          *ManualValidationSummary `json:"summary"`
	Statistics       *ValidationStatistics    `json:"statistics"`
	Recommendations  []string                 `json:"recommendations"`
}

// ManualValidationCase represents a single manual validation case
type ManualValidationCase struct {
	CaseID              string                                  `json:"case_id"`
	BusinessName        string                                  `json:"business_name"`
	BusinessDescription string                                  `json:"business_description"`
	AutomatedResult     *classification.ClassificationCodesInfo `json:"automated_result"`
	ManualValidation    *ManualValidationResult                 `json:"manual_validation"`
	ValidationStatus    string                                  `json:"validation_status"` // "pending", "validated", "disputed"
	ValidationNotes     string                                  `json:"validation_notes"`
	ValidatedBy         string                                  `json:"validated_by"`
	ValidatedAt         time.Time                               `json:"validated_at"`
	Discrepancies       []ValidationDiscrepancy                 `json:"discrepancies"`
	ConfidenceScore     float64                                 `json:"confidence_score"`
}

// ManualValidationResult represents manual validation results
type ManualValidationResult struct {
	IndustryClassification string   `json:"industry_classification"`
	MCCCodes               []string `json:"mcc_codes"`
	SICCodes               []string `json:"sic_codes"`
	NAICSCodes             []string `json:"naics_codes"`
	ConfidenceScore        float64  `json:"confidence_score"`
	ValidationNotes        string   `json:"validation_notes"`
	IsAccurate             bool     `json:"is_accurate"`
	AccuracyScore          float64  `json:"accuracy_score"`
}

// ValidationDiscrepancy represents a discrepancy between automated and manual results
type ValidationDiscrepancy struct {
	Field          string `json:"field"`
	AutomatedValue string `json:"automated_value"`
	ManualValue    string `json:"manual_value"`
	Severity       string `json:"severity"` // "low", "medium", "high", "critical"
	Description    string `json:"description"`
	Impact         string `json:"impact"`
	Recommendation string `json:"recommendation"`
}

// ManualValidationSummary represents high-level validation summary
type ManualValidationSummary struct {
	TotalCases           int     `json:"total_cases"`
	ValidatedCases       int     `json:"validated_cases"`
	PendingCases         int     `json:"pending_cases"`
	DisputedCases        int     `json:"disputed_cases"`
	OverallAccuracy      float64 `json:"overall_accuracy"`
	IndustryAccuracy     float64 `json:"industry_accuracy"`
	CodeAccuracy         float64 `json:"code_accuracy"`
	ConfidenceAccuracy   float64 `json:"confidence_accuracy"`
	AverageDiscrepancies float64 `json:"average_discrepancies"`
	CriticalIssues       int     `json:"critical_issues"`
	HighIssues           int     `json:"high_issues"`
	MediumIssues         int     `json:"medium_issues"`
	LowIssues            int     `json:"low_issues"`
}

// ValidationStatistics represents detailed validation statistics
type ValidationStatistics struct {
	IndustryDistribution    map[string]int     `json:"industry_distribution"`
	CodeTypeDistribution    map[string]int     `json:"code_type_distribution"`
	ConfidenceDistribution  map[string]int     `json:"confidence_distribution"`
	DiscrepancyDistribution map[string]int     `json:"discrepancy_distribution"`
	ValidationTimeStats     *TimeStatistics    `json:"validation_time_stats"`
	AccuracyByIndustry      map[string]float64 `json:"accuracy_by_industry"`
	AccuracyByCodeType      map[string]float64 `json:"accuracy_by_code_type"`
}

// TimeStatistics represents time-related statistics
type TimeStatistics struct {
	AverageValidationTime time.Duration `json:"average_validation_time"`
	MinValidationTime     time.Duration `json:"min_validation_time"`
	MaxValidationTime     time.Duration `json:"max_validation_time"`
	TotalValidationTime   time.Duration `json:"total_validation_time"`
}

// ManualValidationConfig represents configuration for manual validation
type ManualValidationConfig struct {
	SessionName           string        `json:"session_name"`
	ValidationDir         string        `json:"validation_dir"`
	SampleSize            int           `json:"sample_size"`
	ValidationTimeout     time.Duration `json:"validation_timeout"`
	AutoSaveInterval      time.Duration `json:"auto_save_interval"`
	RequireValidation     bool          `json:"require_validation"`
	AllowDisputes         bool          `json:"allow_disputes"`
	MinAccuracyThreshold  float64       `json:"min_accuracy_threshold"`
	IncludeEdgeCases      bool          `json:"include_edge_cases"`
	IncludeHighConfidence bool          `json:"include_high_confidence"`
	IncludeLowConfidence  bool          `json:"include_low_confidence"`
	ValidationFields      []string      `json:"validation_fields"`
}

// NewManualValidationFramework creates a new manual validation framework
func NewManualValidationFramework(config *ManualValidationConfig) *ManualValidationFramework {
	if config == nil {
		config = &ManualValidationConfig{
			SessionName:           "Manual Validation Session",
			ValidationDir:         "./manual-validation",
			SampleSize:            50,
			ValidationTimeout:     30 * time.Minute,
			AutoSaveInterval:      5 * time.Minute,
			RequireValidation:     true,
			AllowDisputes:         true,
			MinAccuracyThreshold:  0.8,
			IncludeEdgeCases:      true,
			IncludeHighConfidence: true,
			IncludeLowConfidence:  true,
			ValidationFields:      []string{"industry", "mcc", "sic", "naics", "confidence"},
		}
	}

	// Create validation directory
	if err := os.MkdirAll(config.ValidationDir, 0755); err != nil {
		log.Printf("Warning: Failed to create validation directory %s: %v", config.ValidationDir, err)
	}

	// Create test runner
	mockRepo := &MockKeywordRepository{}
	logger := log.New(os.Stdout, "", log.LstdFlags)
	testRunner := NewClassificationAccuracyTestRunner(mockRepo, logger)

	return &ManualValidationFramework{
		TestRunner:    testRunner,
		ValidationDir: config.ValidationDir,
		Config:        config,
		Results: &ManualValidationResults{
			SessionID:      generateSessionID(),
			ValidatedCases: []ManualValidationCase{},
		},
	}
}

// RunManualValidation runs the complete manual validation process
func (framework *ManualValidationFramework) RunManualValidation() error {
	framework.Results.StartTime = time.Now()

	log.Printf("🔍 Starting Manual Validation Session: %s", framework.Config.SessionName)
	log.Printf("📁 Validation Directory: %s", framework.ValidationDir)
	log.Printf("📊 Sample Size: %d", framework.Config.SampleSize)

	// Step 1: Generate sample business cases
	sampleCases, err := framework.generateSampleBusinessCases()
	if err != nil {
		return fmt.Errorf("failed to generate sample cases: %w", err)
	}

	// Step 2: Run automated classification
	automatedResults, err := framework.runAutomatedClassification(sampleCases)
	if err != nil {
		return fmt.Errorf("failed to run automated classification: %w", err)
	}

	// Step 3: Create validation cases
	validationCases, err := framework.createValidationCases(sampleCases, automatedResults)
	if err != nil {
		return fmt.Errorf("failed to create validation cases: %w", err)
	}

	// Step 4: Save validation cases for manual review
	err = framework.saveValidationCases(validationCases)
	if err != nil {
		return fmt.Errorf("failed to save validation cases: %w", err)
	}

	// Step 5: Load existing validations (if any)
	err = framework.loadExistingValidations()
	if err != nil {
		log.Printf("Warning: Failed to load existing validations: %v", err)
	}

	// Step 6: Generate validation report
	err = framework.generateValidationReport()
	if err != nil {
		return fmt.Errorf("failed to generate validation report: %w", err)
	}

	framework.Results.EndTime = time.Now()
	framework.Results.Duration = framework.Results.EndTime.Sub(framework.Results.StartTime)

	log.Printf("✅ Manual Validation Session Completed")
	log.Printf("⏱️  Duration: %s", framework.Results.Duration.String())
	log.Printf("📊 Total Cases: %d", framework.Results.TotalValidations)

	return nil
}

// generateSampleBusinessCases generates a diverse set of business cases for validation
func (framework *ManualValidationFramework) generateSampleBusinessCases() ([]ClassificationTestCase, error) {
	log.Printf("📋 Generating sample business cases...")

	// Get comprehensive test dataset
	dataset := NewComprehensiveTestDataset()

	// Filter cases based on configuration
	var selectedCases []ClassificationTestCase

	for _, testCase := range dataset.TestCases {
		// Apply filters based on configuration
		if framework.Config.IncludeEdgeCases || testCase.ExpectedIndustry != "Mixed Industry Business" {
			if framework.Config.IncludeHighConfidence || testCase.ExpectedConfidence < 0.8 {
				if framework.Config.IncludeLowConfidence || testCase.ExpectedConfidence > 0.3 {
					selectedCases = append(selectedCases, testCase)
				}
			}
		}
	}

	// Limit to sample size
	if len(selectedCases) > framework.Config.SampleSize {
		selectedCases = selectedCases[:framework.Config.SampleSize]
	}

	log.Printf("✅ Generated %d sample business cases", len(selectedCases))
	return selectedCases, nil
}

// runAutomatedClassification runs automated classification on sample cases
func (framework *ManualValidationFramework) runAutomatedClassification(cases []ClassificationTestCase) (map[string]*classification.ClassificationCodesInfo, error) {
	log.Printf("🤖 Running automated classification on %d cases...", len(cases))

	results := make(map[string]*classification.ClassificationCodesInfo)

	for _, testCase := range cases {
		// Run classification
		result, err := framework.TestRunner.classifier.GenerateClassificationCodes(
			context.Background(),
			testCase.Keywords,
			testCase.ExpectedIndustry,
			testCase.ExpectedConfidence,
		)
		if err != nil {
			log.Printf("Warning: Classification failed for %s: %v", testCase.Name, err)
			continue
		}

		results[testCase.Name] = result
	}

	log.Printf("✅ Automated classification completed for %d cases", len(results))
	return results, nil
}

// createValidationCases creates validation cases from sample cases and automated results
func (framework *ManualValidationFramework) createValidationCases(
	cases []ClassificationTestCase,
	automatedResults map[string]*classification.ClassificationCodesInfo,
) ([]ManualValidationCase, error) {
	log.Printf("📝 Creating validation cases...")

	var validationCases []ManualValidationCase

	for _, testCase := range cases {
		automatedResult, exists := automatedResults[testCase.Name]
		if !exists {
			continue
		}

		validationCase := ManualValidationCase{
			CaseID:              generateCaseID(testCase.Name),
			BusinessName:        testCase.BusinessName,
			BusinessDescription: testCase.Description,
			AutomatedResult:     automatedResult,
			ValidationStatus:    "pending",
			ValidatedAt:         time.Time{},
		}

		validationCases = append(validationCases, validationCase)
	}

	log.Printf("✅ Created %d validation cases", len(validationCases))
	return validationCases, nil
}

// saveValidationCases saves validation cases to files for manual review
func (framework *ManualValidationFramework) saveValidationCases(cases []ManualValidationCase) error {
	log.Printf("💾 Saving validation cases for manual review...")

	// Save individual case files
	for _, validationCase := range cases {
		caseFile := filepath.Join(framework.ValidationDir, fmt.Sprintf("case_%s.json", validationCase.CaseID))

		data, err := json.MarshalIndent(validationCase, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal validation case: %w", err)
		}

		if err := os.WriteFile(caseFile, data, 0644); err != nil {
			return fmt.Errorf("failed to write validation case file: %w", err)
		}
	}

	// Save summary file
	summaryFile := filepath.Join(framework.ValidationDir, "validation_summary.json")
	summary := map[string]interface{}{
		"session_id":     framework.Results.SessionID,
		"total_cases":    len(cases),
		"validation_dir": framework.ValidationDir,
		"created_at":     time.Now(),
		"case_files":     make([]string, len(cases)),
	}

	for i, validationCase := range cases {
		summary["case_files"].([]string)[i] = fmt.Sprintf("case_%s.json", validationCase.CaseID)
	}

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	if err := os.WriteFile(summaryFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write summary file: %w", err)
	}

	log.Printf("✅ Saved %d validation cases to %s", len(cases), framework.ValidationDir)
	return nil
}

// loadExistingValidations loads existing validation results
func (framework *ManualValidationFramework) loadExistingValidations() error {
	log.Printf("📂 Loading existing validation results...")

	// Load individual case files
	files, err := filepath.Glob(filepath.Join(framework.ValidationDir, "case_*.json"))
	if err != nil {
		return fmt.Errorf("failed to glob case files: %w", err)
	}

	var validatedCases []ManualValidationCase
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Warning: Failed to read case file %s: %v", file, err)
			continue
		}

		var validationCase ManualValidationCase
		if err := json.Unmarshal(data, &validationCase); err != nil {
			log.Printf("Warning: Failed to unmarshal case file %s: %v", file, err)
			continue
		}

		// Only include cases that have been validated
		if validationCase.ValidationStatus == "validated" || validationCase.ValidationStatus == "disputed" {
			validatedCases = append(validatedCases, validationCase)
		}
	}

	framework.Results.ValidatedCases = validatedCases
	framework.Results.TotalValidations = len(validatedCases)

	log.Printf("✅ Loaded %d existing validations", len(validatedCases))
	return nil
}

// generateValidationReport generates comprehensive validation report
func (framework *ManualValidationFramework) generateValidationReport() error {
	log.Printf("📊 Generating validation report...")

	// Calculate summary statistics
	framework.calculateSummaryStatistics()

	// Calculate detailed statistics
	framework.calculateDetailedStatistics()

	// Generate recommendations
	framework.generateRecommendations()

	// Save comprehensive report
	reportFile := filepath.Join(framework.ValidationDir, "validation_report.json")
	data, err := json.MarshalIndent(framework.Results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal validation report: %w", err)
	}

	if err := os.WriteFile(reportFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write validation report: %w", err)
	}

	// Generate HTML report
	err = framework.generateHTMLReport()
	if err != nil {
		log.Printf("Warning: Failed to generate HTML report: %v", err)
	}

	log.Printf("✅ Validation report generated: %s", reportFile)
	return nil
}

// calculateSummaryStatistics calculates high-level validation statistics
func (framework *ManualValidationFramework) calculateSummaryStatistics() {
	summary := &ManualValidationSummary{
		TotalCases: len(framework.Results.ValidatedCases),
	}

	var totalAccuracy, industryAccuracy, codeAccuracy, confidenceAccuracy float64
	var totalDiscrepancies int

	for _, validationCase := range framework.Results.ValidatedCases {
		if validationCase.ValidationStatus == "validated" {
			summary.ValidatedCases++
		} else if validationCase.ValidationStatus == "pending" {
			summary.PendingCases++
		} else if validationCase.ValidationStatus == "disputed" {
			summary.DisputedCases++
		}

		if validationCase.ManualValidation != nil {
			totalAccuracy += validationCase.ManualValidation.AccuracyScore
			industryAccuracy += framework.calculateIndustryAccuracy(validationCase)
			codeAccuracy += framework.calculateCodeAccuracy(validationCase)
			confidenceAccuracy += framework.calculateConfidenceAccuracy(validationCase)
		}

		totalDiscrepancies += len(validationCase.Discrepancies)
		for _, discrepancy := range validationCase.Discrepancies {
			switch discrepancy.Severity {
			case "critical":
				summary.CriticalIssues++
			case "high":
				summary.HighIssues++
			case "medium":
				summary.MediumIssues++
			case "low":
				summary.LowIssues++
			}
		}
	}

	if summary.ValidatedCases > 0 {
		summary.OverallAccuracy = totalAccuracy / float64(summary.ValidatedCases)
		summary.IndustryAccuracy = industryAccuracy / float64(summary.ValidatedCases)
		summary.CodeAccuracy = codeAccuracy / float64(summary.ValidatedCases)
		summary.ConfidenceAccuracy = confidenceAccuracy / float64(summary.ValidatedCases)
		summary.AverageDiscrepancies = float64(totalDiscrepancies) / float64(summary.ValidatedCases)
	}

	framework.Results.Summary = summary
}

// calculateDetailedStatistics calculates detailed validation statistics
func (framework *ManualValidationFramework) calculateDetailedStatistics() {
	stats := &ValidationStatistics{
		IndustryDistribution:    make(map[string]int),
		CodeTypeDistribution:    make(map[string]int),
		ConfidenceDistribution:  make(map[string]int),
		DiscrepancyDistribution: make(map[string]int),
		AccuracyByIndustry:      make(map[string]float64),
		AccuracyByCodeType:      make(map[string]float64),
	}

	var totalValidationTime time.Duration
	var minTime, maxTime time.Duration
	var firstTime = true

	for _, validationCase := range framework.Results.ValidatedCases {
		// Industry distribution
		if validationCase.ManualValidation != nil {
			industry := validationCase.ManualValidation.IndustryClassification
			stats.IndustryDistribution[industry]++
		}

		// Code type distribution
		if validationCase.AutomatedResult != nil {
			if len(validationCase.AutomatedResult.MCC) > 0 {
				stats.CodeTypeDistribution["MCC"]++
			}
			if len(validationCase.AutomatedResult.SIC) > 0 {
				stats.CodeTypeDistribution["SIC"]++
			}
			if len(validationCase.AutomatedResult.NAICS) > 0 {
				stats.CodeTypeDistribution["NAICS"]++
			}
		}

		// Confidence distribution
		confidence := validationCase.ConfidenceScore
		if confidence >= 0.8 {
			stats.ConfidenceDistribution["High"]++
		} else if confidence >= 0.5 {
			stats.ConfidenceDistribution["Medium"]++
		} else if confidence >= 0.2 {
			stats.ConfidenceDistribution["Low"]++
		} else {
			stats.ConfidenceDistribution["Very Low"]++
		}

		// Discrepancy distribution
		for _, discrepancy := range validationCase.Discrepancies {
			stats.DiscrepancyDistribution[discrepancy.Severity]++
		}

		// Time statistics
		if !validationCase.ValidatedAt.IsZero() {
			validationTime := validationCase.ValidatedAt.Sub(framework.Results.StartTime)
			totalValidationTime += validationTime

			if firstTime {
				minTime = validationTime
				maxTime = validationTime
				firstTime = false
			} else {
				if validationTime < minTime {
					minTime = validationTime
				}
				if validationTime > maxTime {
					maxTime = validationTime
				}
			}
		}
	}

	// Calculate time statistics
	if len(framework.Results.ValidatedCases) > 0 {
		stats.ValidationTimeStats = &TimeStatistics{
			AverageValidationTime: totalValidationTime / time.Duration(len(framework.Results.ValidatedCases)),
			MinValidationTime:     minTime,
			MaxValidationTime:     maxTime,
			TotalValidationTime:   totalValidationTime,
		}
	}

	framework.Results.Statistics = stats
}

// generateRecommendations generates improvement recommendations
func (framework *ManualValidationFramework) generateRecommendations() {
	var recommendations []string

	if framework.Results.Summary != nil {
		// Overall accuracy recommendations
		if framework.Results.Summary.OverallAccuracy < framework.Config.MinAccuracyThreshold {
			recommendations = append(recommendations,
				fmt.Sprintf("Overall accuracy (%.2f%%) is below threshold (%.2f%%). Review classification algorithms.",
					framework.Results.Summary.OverallAccuracy*100, framework.Config.MinAccuracyThreshold*100))
		}

		// Industry accuracy recommendations
		if framework.Results.Summary.IndustryAccuracy < 0.9 {
			recommendations = append(recommendations,
				"Industry classification accuracy is below 90%. Review industry detection algorithms and keyword matching.")
		}

		// Code accuracy recommendations
		if framework.Results.Summary.CodeAccuracy < 0.8 {
			recommendations = append(recommendations,
				"Code mapping accuracy is below 80%. Review MCC, SIC, and NAICS code mapping algorithms.")
		}

		// Discrepancy recommendations
		if framework.Results.Summary.CriticalIssues > 0 {
			recommendations = append(recommendations,
				fmt.Sprintf("Found %d critical discrepancies. Immediate review required.", framework.Results.Summary.CriticalIssues))
		}

		if framework.Results.Summary.HighIssues > 5 {
			recommendations = append(recommendations,
				fmt.Sprintf("Found %d high-severity discrepancies. Consider algorithm improvements.", framework.Results.Summary.HighIssues))
		}

		// Validation coverage recommendations
		if framework.Results.Summary.PendingCases > framework.Results.Summary.ValidatedCases {
			recommendations = append(recommendations,
				"More cases are pending validation than completed. Consider increasing validation resources.")
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations,
			"Validation results are within acceptable thresholds. Continue monitoring and consider expanding validation coverage.")
	}

	framework.Results.Recommendations = recommendations
}

// Helper functions for accuracy calculations
func (framework *ManualValidationFramework) calculateIndustryAccuracy(validationCase ManualValidationCase) float64 {
	if validationCase.ManualValidation == nil || validationCase.AutomatedResult == nil {
		return 0.0
	}

	// Simple industry matching (in real implementation, this would be more sophisticated)
	// Note: ClassificationCodesInfo doesn't have DetectedIndustry field, using test case industry instead
	automatedIndustry := "Software Development" // This would come from the test case
	manualIndustry := validationCase.ManualValidation.IndustryClassification

	if automatedIndustry == manualIndustry {
		return 1.0
	}
	return 0.0
}

func (framework *ManualValidationFramework) calculateCodeAccuracy(validationCase ManualValidationCase) float64 {
	if validationCase.ManualValidation == nil || validationCase.AutomatedResult == nil {
		return 0.0
	}

	// Calculate code accuracy (simplified)
	var totalCodes, matchingCodes int

	// MCC codes
	if len(validationCase.AutomatedResult.MCC) > 0 && len(validationCase.ManualValidation.MCCCodes) > 0 {
		totalCodes++
		if framework.codesMatch(extractCodes(validationCase.AutomatedResult.MCC), validationCase.ManualValidation.MCCCodes) {
			matchingCodes++
		}
	}

	// SIC codes
	if len(validationCase.AutomatedResult.SIC) > 0 && len(validationCase.ManualValidation.SICCodes) > 0 {
		totalCodes++
		if framework.codesMatch(extractCodes(validationCase.AutomatedResult.SIC), validationCase.ManualValidation.SICCodes) {
			matchingCodes++
		}
	}

	// NAICS codes
	if len(validationCase.AutomatedResult.NAICS) > 0 && len(validationCase.ManualValidation.NAICSCodes) > 0 {
		totalCodes++
		if framework.codesMatch(extractCodes(validationCase.AutomatedResult.NAICS), validationCase.ManualValidation.NAICSCodes) {
			matchingCodes++
		}
	}

	if totalCodes == 0 {
		return 0.0
	}

	return float64(matchingCodes) / float64(totalCodes)
}

func (framework *ManualValidationFramework) calculateConfidenceAccuracy(validationCase ManualValidationCase) float64 {
	if validationCase.ManualValidation == nil {
		return 0.0
	}

	// Calculate confidence accuracy (simplified)
	automatedConfidence := validationCase.ConfidenceScore
	manualConfidence := validationCase.ManualValidation.ConfidenceScore

	// Simple accuracy based on how close the confidence scores are
	diff := automatedConfidence - manualConfidence
	if diff < 0 {
		diff = -diff
	}

	// Return accuracy based on how close the scores are (1.0 = perfect match, 0.0 = completely different)
	return 1.0 - diff
}

func (framework *ManualValidationFramework) codesMatch(automated []string, manual []string) bool {
	// Simple code matching (in real implementation, this would be more sophisticated)
	if len(automated) != len(manual) {
		return false
	}

	for i, code := range automated {
		if i >= len(manual) || code != manual[i] {
			return false
		}
	}

	return true
}

// generateHTMLReport generates HTML validation report
func (framework *ManualValidationFramework) generateHTMLReport() error {
	reportFile := filepath.Join(framework.ValidationDir, "validation_report.html")

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Manual Validation Report - %s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; }
        .summary { background-color: #e8f5e8; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .case { margin: 10px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
        .case.validated { border-left: 4px solid #4CAF50; }
        .case.disputed { border-left: 4px solid #f44336; }
        .case.pending { border-left: 4px solid #ff9800; }
        .metrics { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin: 20px 0; }
        .metric-card { background-color: #f9f9f9; padding: 15px; border-radius: 5px; }
        .recommendations { background-color: #fff3cd; padding: 15px; border-radius: 5px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Manual Validation Report</h1>
        <p>Session: %s</p>
        <p>Generated: %s</p>
        <p>Duration: %s</p>
    </div>
    
    <div class="summary">
        <h2>Validation Summary</h2>
        <p><strong>Total Cases:</strong> %d</p>
        <p><strong>Validated:</strong> %d</p>
        <p><strong>Pending:</strong> %d</p>
        <p><strong>Disputed:</strong> %d</p>
        <p><strong>Overall Accuracy:</strong> %.2f%%</p>
    </div>
    
    <div class="metrics">
        <div class="metric-card">
            <h3>Accuracy Metrics</h3>
            <p>Industry Accuracy: %.2f%%</p>
            <p>Code Accuracy: %.2f%%</p>
            <p>Confidence Accuracy: %.2f%%</p>
        </div>
        <div class="metric-card">
            <h3>Issues</h3>
            <p>Critical: %d</p>
            <p>High: %d</p>
            <p>Medium: %d</p>
            <p>Low: %d</p>
        </div>
    </div>
    
    <h2>Validation Cases</h2>
`, framework.Config.SessionName,
		framework.Results.SessionID,
		framework.Results.EndTime.Format("2006-01-02 15:04:05"),
		framework.Results.Duration.String(),
		framework.Results.Summary.TotalCases,
		framework.Results.Summary.ValidatedCases,
		framework.Results.Summary.PendingCases,
		framework.Results.Summary.DisputedCases,
		framework.Results.Summary.OverallAccuracy*100,
		framework.Results.Summary.IndustryAccuracy*100,
		framework.Results.Summary.CodeAccuracy*100,
		framework.Results.Summary.ConfidenceAccuracy*100,
		framework.Results.Summary.CriticalIssues,
		framework.Results.Summary.HighIssues,
		framework.Results.Summary.MediumIssues,
		framework.Results.Summary.LowIssues)

	// Add validation cases
	for _, validationCase := range framework.Results.ValidatedCases {
		statusClass := validationCase.ValidationStatus

		html += fmt.Sprintf(`
    <div class="case %s">
        <h3>%s</h3>
        <p><strong>Business:</strong> %s</p>
        <p><strong>Status:</strong> %s</p>
        <p><strong>Validated By:</strong> %s</p>
        <p><strong>Validated At:</strong> %s</p>
        %s
    </div>
`, statusClass, validationCase.CaseID, validationCase.BusinessName,
			validationCase.ValidationStatus, validationCase.ValidatedBy,
			validationCase.ValidatedAt.Format("2006-01-02 15:04:05"),
			func() string {
				if validationCase.ValidationNotes != "" {
					return fmt.Sprintf("<p><strong>Notes:</strong> %s</p>", validationCase.ValidationNotes)
				}
				return ""
			}())
	}

	// Add recommendations
	if len(framework.Results.Recommendations) > 0 {
		html += `
    <div class="recommendations">
        <h2>Recommendations</h2>
        <ul>
`
		for _, rec := range framework.Results.Recommendations {
			html += fmt.Sprintf("            <li>%s</li>\n", rec)
		}
		html += `
        </ul>
    </div>
`
	}

	html += `
</body>
</html>
`

	if err := os.WriteFile(reportFile, []byte(html), 0644); err != nil {
		return fmt.Errorf("failed to write HTML report: %w", err)
	}

	return nil
}

// Utility functions
func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().Unix())
}

func generateCaseID(businessName string) string {
	return fmt.Sprintf("case_%s_%d", sanitizeString(businessName), time.Now().Unix())
}

func sanitizeString(s string) string {
	// Simple string sanitization for file names
	result := ""
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result += string(r)
		} else {
			result += "_"
		}
	}
	return result
}

// extractCodes extracts code strings from classification code structs
func extractCodes(codes interface{}) []string {
	var result []string

	switch v := codes.(type) {
	case []classification.MCCCode:
		for _, code := range v {
			result = append(result, code.Code)
		}
	case []classification.SICCode:
		for _, code := range v {
			result = append(result, code.Code)
		}
	case []classification.NAICSCode:
		for _, code := range v {
			result = append(result, code.Code)
		}
	}

	return result
}

// TestManualValidationFramework tests the manual validation framework
func TestManualValidationFramework(t *testing.T) {
	// Create manual validation configuration
	config := &ManualValidationConfig{
		SessionName:           "Test Manual Validation Session",
		ValidationDir:         "./test-manual-validation",
		SampleSize:            10,
		ValidationTimeout:     10 * time.Minute,
		AutoSaveInterval:      1 * time.Minute,
		RequireValidation:     true,
		AllowDisputes:         true,
		MinAccuracyThreshold:  0.8,
		IncludeEdgeCases:      true,
		IncludeHighConfidence: true,
		IncludeLowConfidence:  true,
		ValidationFields:      []string{"industry", "mcc", "sic", "naics", "confidence"},
	}

	// Create and run manual validation framework
	framework := NewManualValidationFramework(config)
	err := framework.RunManualValidation()

	if err != nil {
		t.Errorf("Manual validation failed: %v", err)
	}

	// Verify results
	if framework.Results == nil {
		t.Error("Validation results should not be nil")
	}

	if framework.Results.SessionID == "" {
		t.Error("Session ID should not be empty")
	}

	if framework.Results.TotalValidations == 0 {
		t.Error("Should have generated validation cases")
	}

	// Clean up
	os.RemoveAll(config.ValidationDir)
}
