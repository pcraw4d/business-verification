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

// ErrorScenarioRunner provides execution and management of error scenario tests
type ErrorScenarioRunner struct {
	logger  *zap.Logger
	config  *ErrorScenarioConfig
	testing *ErrorScenarioTesting
}

// NewErrorScenarioRunner creates a new error scenario runner
func NewErrorScenarioRunner(config *ErrorScenarioConfig) *ErrorScenarioRunner {
	logger := zap.NewNop()
	if config.LogLevel == "debug" {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}

	testing := NewErrorScenarioTesting(config)

	// Load KYB error scenarios
	scenarios := CreateKYBErrorScenarios()
	for _, scenario := range scenarios {
		testing.AddScenario(scenario)
	}

	return &ErrorScenarioRunner{
		logger:  logger,
		config:  config,
		testing: testing,
	}
}

// RunScenarioSuite runs the complete error scenario suite
func (esr *ErrorScenarioRunner) RunScenarioSuite(ctx context.Context) (*ErrorScenarioResults, error) {
	esr.logger.Info("Starting error scenario suite execution")

	results, err := esr.testing.RunScenarioSuite(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run error scenario suite: %w", err)
	}

	// Print summary
	esr.PrintSummary(results)

	return results, nil
}

// RunSpecificScenario runs a specific error scenario
func (esr *ErrorScenarioRunner) RunSpecificScenario(ctx context.Context, scenarioID string) (*ErrorScenarioResult, error) {
	esr.logger.Info("Running specific error scenario", zap.String("scenario_id", scenarioID))

	result, err := esr.testing.RunScenario(ctx, scenarioID)
	if err != nil {
		return nil, fmt.Errorf("failed to run error scenario: %w", err)
	}

	esr.logger.Info("Error scenario completed",
		zap.String("scenario_id", scenarioID),
		zap.Bool("success", result.Success),
		zap.Duration("duration", result.Duration))

	return result, nil
}

// PrintSummary prints a summary of the error scenario results
func (esr *ErrorScenarioRunner) PrintSummary(results *ErrorScenarioResults) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("ERROR SCENARIO TEST SUMMARY")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Session ID: %s\n", results.SessionID)
	fmt.Printf("Environment: %s\n", results.Environment)
	fmt.Printf("Start Time: %s\n", results.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("End Time: %s\n", results.EndTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Total Duration: %s\n", results.TotalDuration)
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total Scenarios: %d\n", results.TotalScenarios)
	fmt.Printf("Passed Scenarios: %d\n", results.PassedScenarios)
	fmt.Printf("Failed Scenarios: %d\n", results.FailedScenarios)
	fmt.Printf("Skipped Scenarios: %d\n", results.SkippedScenarios)
	fmt.Printf("Pass Rate: %.2f%%\n", results.PassRate)
	fmt.Println(strings.Repeat("-", 80))

	// Print error scenario metrics
	if results.Summary != nil {
		fmt.Println("ERROR SCENARIO METRICS:")
		fmt.Printf("  Overall Pass Rate: %.2f%%\n", results.Summary.OverallPassRate)
		fmt.Printf("  Critical Failures: %d\n", results.Summary.CriticalFailures)
		fmt.Printf("  High Severity Failures: %d\n", results.Summary.HighSeverityFailures)
		fmt.Printf("  Medium Severity Failures: %d\n", results.Summary.MediumSeverityFailures)
		fmt.Printf("  Low Severity Failures: %d\n", results.Summary.LowSeverityFailures)
		fmt.Printf("  Recovery Success Rate: %.2f%%\n", results.Summary.RecoverySuccessRate)
		fmt.Printf("  Average Recovery Time: %s\n", results.Summary.AverageRecoveryTime)
		fmt.Printf("  Data Loss Incidents: %d\n", results.Summary.DataLossIncidents)
		fmt.Printf("  Service Downtime: %s\n", results.Summary.ServiceDowntime)
		fmt.Println(strings.Repeat("-", 80))
	}

	// Print scenario results
	fmt.Println("ERROR SCENARIO RESULTS:")
	for scenarioID, scenarioResults := range results.ScenarioResults {
		if len(scenarioResults) == 0 {
			continue
		}

		// Calculate averages for this scenario
		successCount := 0
		recoveryCount := 0
		totalRecoveryTime := time.Duration(0)
		dataLossCount := 0

		for _, result := range scenarioResults {
			if result.Success {
				successCount++
			}
			if result.RecoveryAttempted && result.RecoverySuccess {
				recoveryCount++
				totalRecoveryTime += result.RecoveryTime
			}
			if result.Impact != nil && result.Impact.DataLoss {
				dataLossCount++
			}
		}

		passRate := float64(successCount) / float64(len(scenarioResults)) * 100
		avgRecoveryTime := time.Duration(0)
		if recoveryCount > 0 {
			avgRecoveryTime = totalRecoveryTime / time.Duration(recoveryCount)
		}

		status := "✅ PASSED"
		if successCount == 0 {
			status = "❌ FAILED"
		}

		fmt.Printf("  %s: %s (%d iterations)\n",
			scenarioID, status, len(scenarioResults))
		fmt.Printf("    Pass Rate: %.2f%%, Recovery Time: %s, Data Loss: %d\n",
			passRate, avgRecoveryTime, dataLossCount)
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
	fmt.Printf("Reports generated in: %s\n", esr.config.ReportOutputPath)
	fmt.Println(strings.Repeat("=", 80))
}

// GetResults returns the error scenario results
func (esr *ErrorScenarioRunner) GetResults() *ErrorScenarioResults {
	return esr.testing.GetResults()
}

// GetScenarios returns all error scenarios
func (esr *ErrorScenarioRunner) GetScenarios() map[string]*ErrorScenario {
	return esr.testing.GetScenarios()
}

// parseErrorScenarioCommandLineFlags parses command line flags for error scenario configuration
func parseErrorScenarioCommandLineFlags() *ErrorScenarioConfig {
	var (
		testEnvironment      = flag.String("environment", "test", "Test environment (test, staging, production)")
		testTimeout          = flag.Duration("timeout", 30*time.Minute, "Test timeout duration")
		reportOutputPath     = flag.String("reports", "./error-scenario-reports", "Path to report output directory")
		logLevel             = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		enableErrorInjection = flag.Bool("error-injection", true, "Enable error injection")
		errorInjectionRate   = flag.Float64("injection-rate", 0.1, "Error injection rate (0.0-1.0)")
		recoveryTimeout      = flag.Duration("recovery-timeout", 5*time.Minute, "Recovery timeout duration")
		maxRetryAttempts     = flag.Int("max-retries", 3, "Maximum retry attempts")
	)

	flag.Parse()

	return &ErrorScenarioConfig{
		TestEnvironment:      *testEnvironment,
		TestTimeout:          *testTimeout,
		ReportOutputPath:     *reportOutputPath,
		LogLevel:             *logLevel,
		EnableErrorInjection: *enableErrorInjection,
		ErrorInjectionRate:   *errorInjectionRate,
		RecoveryTimeout:      *recoveryTimeout,
		MaxRetryAttempts:     *maxRetryAttempts,
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

// Main function for running error scenarios from command line
func RunErrorScenariosFromCommandLine() {
	config := parseErrorScenarioCommandLineFlags()
	runner := NewErrorScenarioRunner(config)

	ctx, cancel := context.WithTimeout(context.Background(), config.TestTimeout)
	defer cancel()

	results, err := runner.RunScenarioSuite(ctx)
	if err != nil {
		fmt.Printf("Error scenario execution failed: %v\n", err)
		os.Exit(1)
	}

	if results.PassRate < 80 {
		fmt.Printf("Low pass rate detected: %.2f%%\n", results.PassRate)
		os.Exit(1)
	}

	fmt.Println("All error scenarios passed successfully!")
}
