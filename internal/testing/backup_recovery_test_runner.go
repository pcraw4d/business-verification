package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// BackupRecoveryTestRunner orchestrates the complete backup and recovery testing process
type BackupRecoveryTestRunner struct {
	config  *BackupTestConfig
	tester  *BackupRecoveryTester
	logger  *log.Logger
	results []*BackupTestResult
}

// NewBackupRecoveryTestRunner creates a new test runner
func NewBackupRecoveryTestRunner(config *BackupTestConfig) (*BackupRecoveryTestRunner, error) {
	logger := log.New(os.Stdout, "[BACKUP_RECOVERY_RUNNER] ", log.LstdFlags|log.Lshortfile)

	tester, err := NewBackupRecoveryTester(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup recovery tester: %w", err)
	}

	return &BackupRecoveryTestRunner{
		config:  config,
		tester:  tester,
		logger:  logger,
		results: make([]*BackupTestResult, 0),
	}, nil
}

// RunAllTests runs all backup and recovery tests
func (runner *BackupRecoveryTestRunner) RunAllTests(ctx context.Context) error {
	runner.logger.Printf("Starting comprehensive backup and recovery testing")

	// Test 1: Backup Procedures
	runner.logger.Printf("=== Running Backup Procedures Test ===")
	result, err := runner.tester.TestBackupProcedures(ctx)
	if err != nil {
		runner.logger.Printf("Backup procedures test failed: %v", err)
	} else {
		runner.logger.Printf("Backup procedures test completed successfully")
	}
	runner.results = append(runner.results, result)

	// Test 2: Recovery Scenarios
	runner.logger.Printf("=== Running Recovery Scenarios Test ===")
	result, err = runner.tester.TestRecoveryScenarios(ctx)
	if err != nil {
		runner.logger.Printf("Recovery scenarios test failed: %v", err)
	} else {
		runner.logger.Printf("Recovery scenarios test completed successfully")
	}
	runner.results = append(runner.results, result)

	// Test 3: Data Restoration Validation
	runner.logger.Printf("=== Running Data Restoration Validation ===")
	result, err = runner.tester.TestDataRestoration(ctx)
	if err != nil {
		runner.logger.Printf("Data restoration validation failed: %v", err)
	} else {
		runner.logger.Printf("Data restoration validation completed successfully")
	}
	runner.results = append(runner.results, result)

	// Test 4: Point-in-Time Recovery
	runner.logger.Printf("=== Running Point-in-Time Recovery Test ===")
	result, err = runner.tester.TestPointInTimeRecovery(ctx)
	if err != nil {
		runner.logger.Printf("Point-in-time recovery test failed: %v", err)
	} else {
		runner.logger.Printf("Point-in-time recovery test completed successfully")
	}
	runner.results = append(runner.results, result)

	// Generate comprehensive report
	if err := runner.generateComprehensiveReport(); err != nil {
		return fmt.Errorf("failed to generate comprehensive report: %w", err)
	}

	runner.logger.Printf("All backup and recovery tests completed")
	return nil
}

// generateComprehensiveReport generates a comprehensive test report
func (runner *BackupRecoveryTestRunner) generateComprehensiveReport() error {
	runner.logger.Printf("Generating comprehensive backup and recovery test report")

	report := &BackupRecoveryTestReport{
		TestDate:        time.Now(),
		Config:          runner.config,
		Results:         runner.results,
		Summary:         runner.generateSummary(),
		Recommendations: runner.generateRecommendations(),
	}

	// Save report to file
	reportFile := filepath.Join(runner.config.BackupDirectory, "backup_recovery_test_report.json")
	reportData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(reportFile, reportData, 0644); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	// Generate human-readable summary
	summaryFile := filepath.Join(runner.config.BackupDirectory, "backup_recovery_test_summary.txt")
	if err := runner.generateHumanReadableSummary(summaryFile, report); err != nil {
		return fmt.Errorf("failed to generate human-readable summary: %w", err)
	}

	runner.logger.Printf("Comprehensive report generated: %s", reportFile)
	runner.logger.Printf("Human-readable summary generated: %s", summaryFile)

	return nil
}

// generateSummary generates a test summary
func (runner *BackupRecoveryTestRunner) generateSummary() *TestSummary {
	summary := &TestSummary{
		TotalTests:    len(runner.results),
		PassedTests:   0,
		FailedTests:   0,
		TotalDuration: 0,
		AverageScore:  0,
	}

	var totalScore float64
	for _, result := range runner.results {
		summary.TotalDuration += result.Duration
		if result.Success {
			summary.PassedTests++
		} else {
			summary.FailedTests++
		}
		totalScore += result.ValidationScore
	}

	if len(runner.results) > 0 {
		summary.AverageScore = totalScore / float64(len(runner.results))
	}

	return summary
}

// generateRecommendations generates recommendations based on test results
func (runner *BackupRecoveryTestRunner) generateRecommendations() []string {
	var recommendations []string

	for _, result := range runner.results {
		if !result.Success {
			switch result.TestName {
			case "Backup Procedures Test":
				recommendations = append(recommendations,
					"Review and fix backup procedures. Ensure pg_dump is properly configured and accessible.")
			case "Recovery Scenarios Test":
				recommendations = append(recommendations,
					"Improve recovery procedures. Test with different failure scenarios and ensure proper error handling.")
			case "Data Restoration Validation":
				recommendations = append(recommendations,
					"Address data integrity issues. Review foreign key constraints and data validation procedures.")
			case "Point-in-Time Recovery Test":
				recommendations = append(recommendations,
					"Enhance point-in-time recovery capabilities. Consider implementing WAL-based recovery for better precision.")
			}
		}

		if result.ValidationScore < 0.95 {
			recommendations = append(recommendations,
				fmt.Sprintf("Improve data validation for %s. Current score: %.2f%%, target: 95%%",
					result.TestName, result.ValidationScore*100))
		}

		if result.RecoveryTime > 5*time.Minute {
			recommendations = append(recommendations,
				fmt.Sprintf("Optimize recovery time for %s. Current time: %v, target: <5 minutes",
					result.TestName, result.RecoveryTime))
		}
	}

	// Add general recommendations
	recommendations = append(recommendations,
		"Implement automated backup testing in CI/CD pipeline",
		"Set up monitoring and alerting for backup failures",
		"Document recovery procedures and create runbooks",
		"Conduct regular disaster recovery drills",
		"Consider implementing backup encryption for sensitive data")

	return recommendations
}

// generateHumanReadableSummary generates a human-readable test summary
func (runner *BackupRecoveryTestRunner) generateHumanReadableSummary(filename string, report *BackupRecoveryTestReport) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create summary file: %w", err)
	}
	defer file.Close()

	// Write header
	fmt.Fprintf(file, "BACKUP AND RECOVERY TEST REPORT\n")
	fmt.Fprintf(file, "================================\n\n")
	fmt.Fprintf(file, "Test Date: %s\n", report.TestDate.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "Total Tests: %d\n", report.Summary.TotalTests)
	fmt.Fprintf(file, "Passed Tests: %d\n", report.Summary.PassedTests)
	fmt.Fprintf(file, "Failed Tests: %d\n", report.Summary.FailedTests)
	fmt.Fprintf(file, "Total Duration: %v\n", report.Summary.TotalDuration)
	fmt.Fprintf(file, "Average Validation Score: %.2f%%\n\n", report.Summary.AverageScore*100)

	// Write test results
	fmt.Fprintf(file, "TEST RESULTS\n")
	fmt.Fprintf(file, "============\n\n")
	for _, result := range report.Results {
		status := "PASS"
		if !result.Success {
			status = "FAIL"
		}

		fmt.Fprintf(file, "Test: %s\n", result.TestName)
		fmt.Fprintf(file, "Status: %s\n", status)
		fmt.Fprintf(file, "Duration: %v\n", result.Duration)
		fmt.Fprintf(file, "Validation Score: %.2f%%\n", result.ValidationScore*100)
		if result.RecoveryTime > 0 {
			fmt.Fprintf(file, "Recovery Time: %v\n", result.RecoveryTime)
		}
		if result.ErrorMessage != "" {
			fmt.Fprintf(file, "Error: %s\n", result.ErrorMessage)
		}
		fmt.Fprintf(file, "\n")
	}

	// Write recommendations
	fmt.Fprintf(file, "RECOMMENDATIONS\n")
	fmt.Fprintf(file, "===============\n\n")
	for i, rec := range report.Recommendations {
		fmt.Fprintf(file, "%d. %s\n", i+1, rec)
	}

	return nil
}

// Close closes the test runner and cleans up resources
func (runner *BackupRecoveryTestRunner) Close() error {
	if runner.tester != nil {
		return runner.tester.Close()
	}
	return nil
}

// BackupRecoveryTestReport represents the complete test report
type BackupRecoveryTestReport struct {
	TestDate        time.Time           `json:"test_date"`
	Config          *BackupTestConfig   `json:"config"`
	Results         []*BackupTestResult `json:"results"`
	Summary         *TestSummary        `json:"summary"`
	Recommendations []string            `json:"recommendations"`
}

// TestSummary contains summary statistics
type TestSummary struct {
	TotalTests    int           `json:"total_tests"`
	PassedTests   int           `json:"passed_tests"`
	FailedTests   int           `json:"failed_tests"`
	TotalDuration time.Duration `json:"total_duration"`
	AverageScore  float64       `json:"average_score"`
}
