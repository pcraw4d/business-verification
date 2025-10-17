package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"kyb-platform/test"
)

func main() {
	// Command line flags
	var (
		configFile = flag.String("config", "", "Path to calibration configuration file (JSON)")
		verbose    = flag.Bool("verbose", false, "Enable verbose output")
		help       = flag.Bool("help", false, "Show help information")
	)

	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Load configuration
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Printf("Warning: Failed to parse config file %s: %v", *configFile, err)
		config = getDefaultConfig()
	}

	if *verbose {
		printConfig(config)
	}

	// Create logger
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Create test runner
	mockRepo := &test.MockKeywordRepository{}
	testRunner := test.NewClassificationAccuracyTestRunner(mockRepo, logger)

	// Create validator
	validator := test.NewConfidenceScoreCalibrationValidator(testRunner, logger, config)

	logger.Printf("🎯 Starting Confidence Score Calibration Validation...")

	// Run validation
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	result, err := validator.ValidateCalibration(ctx)
	if err != nil {
		logger.Fatalf("❌ Calibration validation failed: %v", err)
	}

	// Save report
	if err := validator.SaveCalibrationReport(result); err != nil {
		logger.Printf("⚠️  Failed to save report: %v", err)
	}

	// Print summary
	printCalibrationSummary(result, *verbose)

	logger.Printf("✅ Confidence Score Calibration Validation Completed!")
}

func loadConfig(configFile string) (*test.CalibrationValidationConfig, error) {
	if configFile == "" {
		return getDefaultConfig(), nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config test.CalibrationValidationConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func getDefaultConfig() *test.CalibrationValidationConfig {
	return &test.CalibrationValidationConfig{
		SessionName:                     "Confidence Score Calibration Validation",
		ValidationDirectory:             "./confidence-calibration-validation",
		SampleSize:                      100,
		Timeout:                         30 * time.Minute,
		MinCalibrationThreshold:         0.80,
		IncludeReliabilityDiagram:       true,
		IncludeCalibrationCurve:         true,
		IncludeBrierScore:               true,
		IncludeExpectedCalibrationError: true,
		IncludeTemperatureScaling:       true,
		GenerateDetailedReport:          true,
	}
}

func printConfig(config *test.CalibrationValidationConfig) {
	fmt.Printf("📋 Confidence Score Calibration Configuration:\n")
	fmt.Printf("   Session Name: %s\n", config.SessionName)
	fmt.Printf("   Validation Directory: %s\n", config.ValidationDirectory)
	fmt.Printf("   Sample Size: %d\n", config.SampleSize)
	fmt.Printf("   Timeout: %v\n", config.Timeout)
	fmt.Printf("   Min Calibration Threshold: %.2f\n", config.MinCalibrationThreshold)
	fmt.Printf("   Reliability Diagram: %t\n", config.IncludeReliabilityDiagram)
	fmt.Printf("   Calibration Curve: %t\n", config.IncludeCalibrationCurve)
	fmt.Printf("   Brier Score: %t\n", config.IncludeBrierScore)
	fmt.Printf("   Expected Calibration Error: %t\n", config.IncludeExpectedCalibrationError)
	fmt.Printf("   Temperature Scaling: %t\n", config.IncludeTemperatureScaling)
	fmt.Printf("   Detailed Report: %t\n", config.GenerateDetailedReport)
	fmt.Println()
}

func printCalibrationSummary(result *test.CalibrationValidationResult, verbose bool) {
	fmt.Printf("🏁 Confidence Score Calibration Validation Completed!\n")
	fmt.Printf("⏱️  Duration: %v\n", result.Duration)
	fmt.Printf("📊 Session ID: %s\n", result.SessionID)
	fmt.Printf("📁 Validation Directory: %s\n", result.SessionID)
	fmt.Printf("📋 Calibration Summary:\n")

	if result.CalibrationSummary != nil {
		fmt.Printf("   Overall Calibration: %.3f\n", result.CalibrationSummary.OverallCalibration)
		fmt.Printf("   Calibration Quality: %s\n", result.CalibrationSummary.CalibrationQuality)
		fmt.Printf("   Is Well Calibrated: %t\n", result.CalibrationSummary.IsWellCalibrated)
		fmt.Printf("   Calibration Error: %.3f\n", result.CalibrationSummary.CalibrationError)
		fmt.Printf("   Brier Score: %.3f\n", result.CalibrationSummary.BrierScore)
		fmt.Printf("   Expected Calibration Error: %.3f\n", result.CalibrationSummary.ExpectedCalibrationError)
	}

	fmt.Printf("📊 Confidence Bins: %d bins analyzed\n", len(result.ConfidenceBins))

	if verbose && len(result.ConfidenceBins) > 0 {
		fmt.Printf("📊 Bin Details:\n")
		for _, bin := range result.ConfidenceBins {
			fmt.Printf("   Bin %d (%.2f-%.2f): %d samples, accuracy=%.3f, confidence=%.3f, error=%.3f\n",
				bin.BinIndex, bin.ConfidenceMin, bin.ConfidenceMax, bin.SampleCount,
				bin.ActualAccuracy, bin.PredictedConfidence, bin.CalibrationError)
		}
	}

	if len(result.Recommendations) > 0 {
		fmt.Printf("💡 Recommendations: %d suggestions generated\n", len(result.Recommendations))
		fmt.Printf("📋 Recommendations:\n")
		for i, rec := range result.Recommendations {
			fmt.Printf("   %d. %s\n", i+1, rec)
		}
	}

	fmt.Printf("📁 Validation files generated in: %s\n", result.SessionID)
	fmt.Printf("📄 Reports available:\n")
	fmt.Printf("   - confidence_calibration_report.json (comprehensive JSON report)\n")
	fmt.Printf("   - confidence_calibration_report.html (human-readable HTML report)\n")
	fmt.Printf("   - confidence_calibration_summary.json (session summary)\n")
}

func showHelp() {
	fmt.Printf("Confidence Score Calibration Validator\n")
	fmt.Printf("=====================================\n\n")
	fmt.Printf("Usage: confidence-calibration-validator [options]\n\n")
	fmt.Printf("Options:\n")
	fmt.Printf("  -config string\n")
	fmt.Printf("        Path to calibration configuration file (JSON)\n")
	fmt.Printf("  -verbose\n")
	fmt.Printf("        Enable verbose output\n")
	fmt.Printf("  -help\n")
	fmt.Printf("        Show this help information\n\n")
	fmt.Printf("Examples:\n")
	fmt.Printf("  confidence-calibration-validator\n")
	fmt.Printf("  confidence-calibration-validator -config configs/calibration-config.json\n")
	fmt.Printf("  confidence-calibration-validator -verbose\n\n")
	fmt.Printf("Configuration File Format:\n")
	fmt.Printf("  {\n")
	fmt.Printf("    \"session_name\": \"Confidence Score Calibration Validation\",\n")
	fmt.Printf("    \"validation_directory\": \"./confidence-calibration-validation\",\n")
	fmt.Printf("    \"sample_size\": 100,\n")
	fmt.Printf("    \"timeout\": \"30m\",\n")
	fmt.Printf("    \"min_calibration_threshold\": 0.80,\n")
	fmt.Printf("    \"include_reliability_diagram\": true,\n")
	fmt.Printf("    \"include_calibration_curve\": true,\n")
	fmt.Printf("    \"include_brier_score\": true,\n")
	fmt.Printf("    \"include_expected_calibration_error\": true,\n")
	fmt.Printf("    \"include_temperature_scaling\": true,\n")
	fmt.Printf("    \"generate_detailed_report\": true\n")
	fmt.Printf("  }\n")
}
