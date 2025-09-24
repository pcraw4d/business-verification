package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pcraw4d/business-verification/test"
)

func main() {
	// Command line flags
	var (
		configFile = flag.String("config", "", "Path to validation configuration file (JSON)")
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
	validator := test.NewIndustryCodeMappingValidator(testRunner, logger, config)

	logger.Printf("🔍 Starting Industry Code Mapping Validation...")

	// Run validation
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	result, err := validator.ValidateCodeMapping(ctx)
	if err != nil {
		logger.Fatalf("❌ Validation failed: %v", err)
	}

	// Save report
	if err := validator.SaveValidationReport(result); err != nil {
		logger.Printf("⚠️  Failed to save report: %v", err)
	}

	// Print summary
	printValidationSummary(result, *verbose)

	logger.Printf("✅ Industry Code Mapping Validation Completed!")
}

func loadConfig(configFile string) (*test.CodeMappingValidationConfig, error) {
	if configFile == "" {
		return getDefaultConfig(), nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config test.CodeMappingValidationConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func getDefaultConfig() *test.CodeMappingValidationConfig {
	return &test.CodeMappingValidationConfig{
		SessionName:                     "Industry Code Mapping Validation",
		ValidationDirectory:             "./code-mapping-validation",
		SampleSize:                      50,
		Timeout:                         30 * time.Minute,
		MinAccuracyThreshold:            0.80,
		IncludeFormatValidation:         true,
		IncludeStructureValidation:      true,
		IncludeCrossReferenceValidation: true,
		GenerateDetailedReport:          true,
	}
}

func printConfig(config *test.CodeMappingValidationConfig) {
	fmt.Printf("📋 Industry Code Mapping Validation Configuration:\n")
	fmt.Printf("   Session Name: %s\n", config.SessionName)
	fmt.Printf("   Validation Directory: %s\n", config.ValidationDirectory)
	fmt.Printf("   Sample Size: %d\n", config.SampleSize)
	fmt.Printf("   Timeout: %v\n", config.Timeout)
	fmt.Printf("   Min Accuracy Threshold: %.2f\n", config.MinAccuracyThreshold)
	fmt.Printf("   Format Validation: %t\n", config.IncludeFormatValidation)
	fmt.Printf("   Structure Validation: %t\n", config.IncludeStructureValidation)
	fmt.Printf("   Cross-Reference Validation: %t\n", config.IncludeCrossReferenceValidation)
	fmt.Printf("   Detailed Report: %t\n", config.GenerateDetailedReport)
	fmt.Println()
}

func printValidationSummary(result *test.CodeMappingValidationResult, verbose bool) {
	fmt.Printf("🏁 Industry Code Mapping Validation Completed!\n")
	fmt.Printf("⏱️  Duration: %v\n", result.Duration)
	fmt.Printf("📊 Session ID: %s\n", result.SessionID)
	fmt.Printf("📁 Validation Directory: %s\n", result.ValidationSummary)
	fmt.Printf("📋 Validation Summary:\n")

	if result.ValidationSummary != nil {
		fmt.Printf("   Overall Accuracy: %.2f%%\n", result.ValidationSummary.OverallAccuracy*100)
		fmt.Printf("   MCC Accuracy: %.2f%%\n", result.ValidationSummary.MCCAccuracy*100)
		fmt.Printf("   SIC Accuracy: %.2f%%\n", result.ValidationSummary.SICAccuracy*100)
		fmt.Printf("   NAICS Accuracy: %.2f%%\n", result.ValidationSummary.NAICSAccuracy*100)
		fmt.Printf("   Format Validation: %t\n", result.ValidationSummary.FormatValidationPassed)
		fmt.Printf("   Structure Validation: %t\n", result.ValidationSummary.StructureValidationPassed)
		fmt.Printf("   Cross-Reference Validation: %t\n", result.ValidationSummary.CrossReferencePassed)
	}

	fmt.Printf("📊 Issues Found:\n")
	if result.ValidationSummary != nil {
		fmt.Printf("   Critical: %d\n", result.ValidationSummary.CriticalIssues)
		fmt.Printf("   High: %d\n", result.ValidationSummary.HighIssues)
		fmt.Printf("   Medium: %d\n", result.ValidationSummary.MediumIssues)
		fmt.Printf("   Low: %d\n", result.ValidationSummary.LowIssues)
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
	fmt.Printf("   - code_mapping_validation_report.json (comprehensive JSON report)\n")
	fmt.Printf("   - code_mapping_validation_report.html (human-readable HTML report)\n")
	fmt.Printf("   - code_mapping_validation_summary.json (session summary)\n")
}

func showHelp() {
	fmt.Printf("Industry Code Mapping Validator\n")
	fmt.Printf("===============================\n\n")
	fmt.Printf("Usage: code-mapping-validator [options]\n\n")
	fmt.Printf("Options:\n")
	fmt.Printf("  -config string\n")
	fmt.Printf("        Path to validation configuration file (JSON)\n")
	fmt.Printf("  -format string\n")
	fmt.Printf("        Output format: json, html, text (default: json)\n")
	fmt.Printf("  -verbose\n")
	fmt.Printf("        Enable verbose output\n")
	fmt.Printf("  -help\n")
	fmt.Printf("        Show this help information\n\n")
	fmt.Printf("Examples:\n")
	fmt.Printf("  code-mapping-validator\n")
	fmt.Printf("  code-mapping-validator -config configs/code-mapping-config.json\n")
	fmt.Printf("  code-mapping-validator -verbose -format html\n\n")
	fmt.Printf("Configuration File Format:\n")
	fmt.Printf("  {\n")
	fmt.Printf("    \"session_name\": \"Industry Code Mapping Validation\",\n")
	fmt.Printf("    \"validation_directory\": \"./code-mapping-validation\",\n")
	fmt.Printf("    \"sample_size\": 50,\n")
	fmt.Printf("    \"timeout\": \"30m\",\n")
	fmt.Printf("    \"min_accuracy_threshold\": 0.80,\n")
	fmt.Printf("    \"include_format_validation\": true,\n")
	fmt.Printf("    \"include_structure_validation\": true,\n")
	fmt.Printf("    \"include_cross_reference_validation\": true,\n")
	fmt.Printf("    \"generate_detailed_report\": true\n")
	fmt.Printf("  }\n")
}
