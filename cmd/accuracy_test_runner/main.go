package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"kyb-platform/internal/classification"
)

func main() {
	// Command line flags
	var (
		dbHost     = flag.String("db-host", "localhost", "Database host")
		dbPort     = flag.Int("db-port", 5432, "Database port")
		dbUser     = flag.String("db-user", "postgres", "Database user")
		dbPassword = flag.String("db-password", "", "Database password")
		dbName     = flag.String("db-name", "business_verification", "Database name")
		testType   = flag.String("test-type", "all", "Type of test to run (all, mcc, naics, sic, confidence, validation, consistency, alignment)")
		outputFile = flag.String("output", "", "Output file for test results (JSON format)")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Setup logging
	var logger *zap.Logger
	var err error

	if *verbose {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting crosswalk accuracy test runner",
		zap.String("test_type", *testType),
		zap.String("db_host", *dbHost),
		zap.Int("db_port", *dbPort),
		zap.String("db_name", *dbName))

	// Connect to database
	dbURL := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		*dbHost, *dbPort, *dbUser, *dbPassword, *dbName)

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}

	logger.Info("Successfully connected to database")

	// Create accuracy tester
	tester := classification.NewCrosswalkAccuracyTester(db, logger)

	// Run tests based on test type
	var suite *classification.AccuracyTestSuite

	switch *testType {
	case "all":
		suite, err = tester.RunComprehensiveAccuracyTests(ctx)
	case "mcc":
		suite, err = runSingleTestType(ctx, tester, "mcc_mapping_accuracy")
	case "naics":
		suite, err = runSingleTestType(ctx, tester, "naics_mapping_accuracy")
	case "sic":
		suite, err = runSingleTestType(ctx, tester, "sic_mapping_accuracy")
	case "confidence":
		suite, err = runSingleTestType(ctx, tester, "confidence_scoring_accuracy")
	case "validation":
		suite, err = runSingleTestType(ctx, tester, "validation_rules_accuracy")
	case "consistency":
		suite, err = runSingleTestType(ctx, tester, "crosswalk_consistency_accuracy")
	case "alignment":
		suite, err = runSingleTestType(ctx, tester, "industry_alignment_accuracy")
	default:
		logger.Fatal("Invalid test type", zap.String("test_type", *testType))
	}

	if err != nil {
		logger.Fatal("Failed to run accuracy tests", zap.Error(err))
	}

	// Display results
	displayTestResults(logger, suite)

	// Save results to database
	if err := tester.SaveAccuracyTestResults(ctx, suite); err != nil {
		logger.Error("Failed to save test results to database", zap.Error(err))
	}

	// Save results to file if specified
	if *outputFile != "" {
		if err := saveResultsToFile(suite, *outputFile); err != nil {
			logger.Error("Failed to save results to file",
				zap.String("output_file", *outputFile),
				zap.Error(err))
		} else {
			logger.Info("Results saved to file", zap.String("output_file", *outputFile))
		}
	}

	// Exit with appropriate code based on overall score
	if suite.OverallScore >= 0.8 {
		logger.Info("Accuracy tests completed successfully",
			zap.Float64("overall_score", suite.OverallScore))
		os.Exit(0)
	} else {
		logger.Warn("Accuracy tests completed with low score",
			zap.Float64("overall_score", suite.OverallScore))
		os.Exit(1)
	}
}

func runSingleTestType(ctx context.Context, tester *classification.CrosswalkAccuracyTester, testType string) (*classification.AccuracyTestSuite, error) {
	logger := zap.L()

	logger.Info("Running single test type", zap.String("test_type", testType))

	// Create a minimal suite for single test type
	suite := &classification.AccuracyTestSuite{
		SuiteName:   fmt.Sprintf("Single Test Suite - %s", testType),
		Description: fmt.Sprintf("Running %s accuracy test", testType),
		TestCases:   []classification.AccuracyTestCase{},
		TestResults: []classification.AccuracyTestResult{},
		CreatedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	// Run the specific test type
	result, err := tester.RunAccuracyTestByType(ctx, testType)
	if err != nil {
		return nil, fmt.Errorf("failed to run test type %s: %w", testType, err)
	}

	suite.TestResults = append(suite.TestResults, *result)
	suite.OverallScore = result.AccuracyScore

	return suite, nil
}

func displayTestResults(logger *zap.Logger, suite *classification.AccuracyTestSuite) {
	logger.Info("=== ACCURACY TEST RESULTS ===",
		zap.String("suite_name", suite.SuiteName),
		zap.Float64("overall_score", suite.OverallScore),
		zap.Int("total_test_types", len(suite.TestResults)))

	fmt.Printf("\n=== CROSSWALK ACCURACY TEST RESULTS ===\n")
	fmt.Printf("Suite: %s\n", suite.SuiteName)
	fmt.Printf("Description: %s\n", suite.Description)
	fmt.Printf("Overall Score: %.2f%%\n", suite.OverallScore*100)
	fmt.Printf("Total Test Types: %d\n", len(suite.TestResults))
	fmt.Printf("Created At: %s\n\n", suite.CreatedAt.Format(time.RFC3339))

	for i, result := range suite.TestResults {
		fmt.Printf("Test %d: %s\n", i+1, result.TestName)
		fmt.Printf("  Type: %s\n", result.TestType)
		fmt.Printf("  Total Tests: %d\n", result.TotalTests)
		fmt.Printf("  Passed: %d\n", result.PassedTests)
		fmt.Printf("  Failed: %d\n", result.FailedTests)
		fmt.Printf("  Accuracy: %.2f%%\n", result.AccuracyScore*100)
		fmt.Printf("  Confidence: %.2f%%\n", result.ConfidenceScore*100)
		fmt.Printf("  Summary: %s\n", result.Summary)
		fmt.Printf("  Timestamp: %s\n\n", result.Timestamp.Format(time.RFC3339))

		// Log individual test details if verbose
		if len(result.TestDetails) > 0 {
			logger.Info("Test details",
				zap.String("test_name", result.TestName),
				zap.Int("detail_count", len(result.TestDetails)))
		}
	}

	// Display summary statistics
	totalTests := 0
	totalPassed := 0
	totalFailed := 0

	for _, result := range suite.TestResults {
		totalTests += result.TotalTests
		totalPassed += result.PassedTests
		totalFailed += result.FailedTests
	}

	fmt.Printf("=== SUMMARY STATISTICS ===\n")
	fmt.Printf("Total Tests Run: %d\n", totalTests)
	fmt.Printf("Total Passed: %d\n", totalPassed)
	fmt.Printf("Total Failed: %d\n", totalFailed)
	fmt.Printf("Overall Pass Rate: %.2f%%\n", float64(totalPassed)/float64(totalTests)*100)
	fmt.Printf("Overall Score: %.2f%%\n", suite.OverallScore*100)

	// Display recommendations
	fmt.Printf("\n=== RECOMMENDATIONS ===\n")
	if suite.OverallScore >= 0.9 {
		fmt.Printf("✅ Excellent accuracy! The crosswalk system is performing very well.\n")
	} else if suite.OverallScore >= 0.8 {
		fmt.Printf("✅ Good accuracy! Minor improvements may be beneficial.\n")
	} else if suite.OverallScore >= 0.7 {
		fmt.Printf("⚠️  Moderate accuracy. Consider reviewing failed test cases.\n")
	} else {
		fmt.Printf("❌ Low accuracy. Significant improvements needed.\n")
	}

	// Identify areas for improvement
	for _, result := range suite.TestResults {
		if result.AccuracyScore < 0.8 {
			fmt.Printf("• Review %s - accuracy: %.2f%%\n", result.TestName, result.AccuracyScore*100)
		}
	}
}

func saveResultsToFile(suite *classification.AccuracyTestSuite, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// For now, save as simple text format
	// In a real implementation, this would be JSON or another structured format
	fmt.Fprintf(file, "Crosswalk Accuracy Test Results\n")
	fmt.Fprintf(file, "Generated: %s\n\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "Suite: %s\n", suite.SuiteName)
	fmt.Fprintf(file, "Overall Score: %.2f%%\n\n", suite.OverallScore*100)

	for _, result := range suite.TestResults {
		fmt.Fprintf(file, "Test: %s\n", result.TestName)
		fmt.Fprintf(file, "  Accuracy: %.2f%% (%d/%d)\n",
			result.AccuracyScore*100, result.PassedTests, result.TotalTests)
		fmt.Fprintf(file, "  Confidence: %.2f%%\n", result.ConfidenceScore*100)
		fmt.Fprintf(file, "  Summary: %s\n\n", result.Summary)
	}

	return nil
}
