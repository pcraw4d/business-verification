package risk

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

// UATRunner provides execution and management of user acceptance tests
type UATRunner struct {
	logger  *zap.Logger
	config  *UATConfig
	testing *UserAcceptanceTesting
}

// NewUATRunner creates a new UAT runner
func NewUATRunner(config *UATConfig) *UATRunner {
	logger := zap.NewNop()
	if config.LogLevel == "debug" {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}

	testing := NewUserAcceptanceTesting(config)

	// Load KYB UAT test cases
	testCases := CreateKYBUATTestCases()
	for _, testCase := range testCases {
		testing.AddTestCase(testCase)
	}

	return &UATRunner{
		logger:  logger,
		config:  config,
		testing: testing,
	}
}

// RunUATSuite runs the complete UAT suite
func (uatr *UATRunner) RunUATSuite(ctx context.Context) (*UATResults, error) {
	uatr.logger.Info("Starting UAT suite execution")

	results, err := uatr.testing.RunUATSuite(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run UAT suite: %w", err)
	}

	// Print summary
	uatr.PrintSummary(results)

	return results, nil
}

// RunSpecificTestCase runs a specific UAT test case
func (uatr *UATRunner) RunSpecificTestCase(ctx context.Context, testCaseID string, userID string, userRole string) (*UATResult, error) {
	uatr.logger.Info("Running specific UAT test case", zap.String("test_case_id", testCaseID), zap.String("user_id", userID))

	result, err := uatr.testing.RunTestCase(ctx, testCaseID, userID, userRole)
	if err != nil {
		return nil, fmt.Errorf("failed to run UAT test case: %w", err)
	}

	uatr.logger.Info("UAT test case completed",
		zap.String("test_case_id", testCaseID),
		zap.String("user_id", userID),
		zap.Bool("success", result.Success),
		zap.Duration("duration", result.Duration))

	return result, nil
}

// PrintSummary prints a summary of the UAT results
func (uatr *UATRunner) PrintSummary(results *UATResults) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("USER ACCEPTANCE TESTING SUMMARY")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Session ID: %s\n", results.SessionID)
	fmt.Printf("Environment: %s\n", results.Environment)
	fmt.Printf("Start Time: %s\n", results.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("End Time: %s\n", results.EndTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Total Duration: %s\n", results.TotalDuration)
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total Test Cases: %d\n", results.TotalTestCases)
	fmt.Printf("Passed Test Cases: %d\n", results.PassedTestCases)
	fmt.Printf("Failed Test Cases: %d\n", results.FailedTestCases)
	fmt.Printf("Skipped Test Cases: %d\n", results.SkippedTestCases)
	fmt.Printf("Pass Rate: %.2f%%\n", results.PassRate)
	fmt.Println(strings.Repeat("-", 80))

	// Print UAT metrics
	if results.Summary != nil {
		fmt.Println("USER EXPERIENCE METRICS:")
		fmt.Printf("  Overall User Satisfaction: %.2f/10\n", results.Summary.OverallUserSatisfaction)
		fmt.Printf("  Overall Usability Score: %.2f/10\n", results.Summary.OverallUsabilityScore)
		fmt.Printf("  Average Completion Time: %s\n", results.Summary.AverageCompletionTime)
		fmt.Printf("  Average Error Rate: %.2f%%\n", results.Summary.AverageErrorRate)
		fmt.Printf("  Recommendation Rate: %.2f%%\n", results.Summary.RecommendationRate)
		fmt.Println(strings.Repeat("-", 80))
	}

	// Print test case results
	fmt.Println("UAT TEST CASE RESULTS:")
	for testCaseID, testCaseResults := range results.TestCaseResults {
		if len(testCaseResults) == 0 {
			continue
		}

		// Calculate averages for this test case
		successCount := 0
		totalSatisfaction := 0.0
		totalUsability := 0.0
		totalCompletionTime := time.Duration(0)
		totalErrorRate := 0.0
		recommendationCount := 0

		for _, result := range testCaseResults {
			if result.Success {
				successCount++
			}
			if result.UserSatisfaction != nil {
				totalSatisfaction += result.UserSatisfaction.OverallRating
				if result.UserSatisfaction.WouldRecommend {
					recommendationCount++
				}
			}
			if result.UsabilityMetrics != nil {
				totalUsability += result.UsabilityMetrics.SatisfactionScore
				totalCompletionTime += result.UsabilityMetrics.TimeToComplete
				totalErrorRate += result.UsabilityMetrics.ErrorRate
			}
		}

		passRate := float64(successCount) / float64(len(testCaseResults)) * 100
		avgSatisfaction := totalSatisfaction / float64(len(testCaseResults))
		avgUsability := totalUsability / float64(len(testCaseResults))
		avgCompletionTime := totalCompletionTime / time.Duration(len(testCaseResults))
		avgErrorRate := totalErrorRate / float64(len(testCaseResults))
		recommendationRate := float64(recommendationCount) / float64(len(testCaseResults)) * 100

		status := "✅ PASSED"
		if successCount == 0 {
			status = "❌ FAILED"
		}

		fmt.Printf("  %s: %s (%d users)\n",
			testCaseID, status, len(testCaseResults))
		fmt.Printf("    Pass Rate: %.2f%%, Satisfaction: %.2f/10, Usability: %.2f/10\n",
			passRate, avgSatisfaction, avgUsability)
		fmt.Printf("    Completion Time: %s, Error Rate: %.2f%%, Recommendations: %.2f%%\n",
			avgCompletionTime, avgErrorRate, recommendationRate)
	}

	// Print recommendations
	if len(results.Recommendations) > 0 {
		fmt.Println(strings.Repeat("-", 80))
		fmt.Println("RECOMMENDATIONS:")
		for i, recommendation := range results.Recommendations {
			fmt.Printf("  %d. %s\n", i+1, recommendation)
		}
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Reports generated in: %s\n", uatr.config.ReportOutputPath)
	fmt.Println(strings.Repeat("=", 80))
}

// GetResults returns the UAT results
func (uatr *UATRunner) GetResults() *UATResults {
	return uatr.testing.GetResults()
}

// GetTestCases returns all UAT test cases
func (uatr *UATRunner) GetTestCases() map[string]*UATTestCase {
	return uatr.testing.GetTestCases()
}

// parseUATCommandLineFlags parses command line flags for UAT configuration
func parseUATCommandLineFlags() *UATConfig {
	var (
		testEnvironment      = flag.String("environment", "test", "Test environment (test, staging, production)")
		testTimeout          = flag.Duration("timeout", 30*time.Minute, "Test timeout duration")
		reportOutputPath     = flag.String("reports", "./uat-reports", "Path to report output directory")
		logLevel             = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		enableUserSimulation = flag.Bool("user-simulation", true, "Enable user simulation")
		userCount            = flag.Int("user-count", 3, "Number of users per test case")
		testDuration         = flag.Duration("test-duration", 10*time.Minute, "Test duration")
		feedbackCollection   = flag.Bool("feedback-collection", true, "Enable feedback collection")
	)

	flag.Parse()

	return &UATConfig{
		TestEnvironment:      *testEnvironment,
		TestTimeout:          *testTimeout,
		ReportOutputPath:     *reportOutputPath,
		LogLevel:             *logLevel,
		EnableUserSimulation: *enableUserSimulation,
		UserCount:            *userCount,
		TestDuration:         *testDuration,
		FeedbackCollection:   *feedbackCollection,
		EnvironmentVariables: make(map[string]string),
		DatabaseConfig: &DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "kyb_test",
			Username: "test_user",
			Password: "test_password",
			SSLMode:  "disable",
		},
		APIConfig: &APIConfig{
			BaseURL:   "http://localhost:8080",
			Timeout:   30 * time.Second,
			Headers:   make(map[string]string),
			RateLimit: 1000,
		},
		ResourceLimits: &ResourceLimits{
			MaxMemoryMB:     1024,
			MaxCPUPercents:  80,
			MaxGoroutines:   1000,
			MaxConnections:  100,
			MaxFileHandles:  1000,
			TimeoutDuration: 30 * time.Minute,
		},
	}
}

// Main function for running UAT from command line
func RunUATFromCommandLine() {
	config := parseUATCommandLineFlags()
	runner := NewUATRunner(config)

	ctx, cancel := context.WithTimeout(context.Background(), config.TestTimeout)
	defer cancel()

	results, err := runner.RunUATSuite(ctx)
	if err != nil {
		fmt.Printf("UAT execution failed: %v\n", err)
		os.Exit(1)
	}

	if results.PassRate < 80 {
		fmt.Printf("Low pass rate detected: %.2f%%\n", results.PassRate)
		os.Exit(1)
	}

	if results.Summary.OverallUserSatisfaction < 7.0 {
		fmt.Printf("Low user satisfaction detected: %.2f/10\n", results.Summary.OverallUserSatisfaction)
		os.Exit(1)
	}

	fmt.Println("All UAT tests passed successfully!")
}
