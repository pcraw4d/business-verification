package main

import (
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
		configFile        = flag.String("config", "", "Path to validation configuration file (JSON)")
		validationDir     = flag.String("dir", "./manual-validation", "Validation directory")
		sampleSize        = flag.Int("sample-size", 50, "Number of sample business cases")
		sessionName       = flag.String("session", "Manual Validation Session", "Session name")
		timeout           = flag.Duration("timeout", 30*time.Minute, "Validation timeout duration")
		autoSave          = flag.Duration("auto-save", 5*time.Minute, "Auto-save interval")
		requireValidation = flag.Bool("require-validation", true, "Require manual validation")
		allowDisputes     = flag.Bool("allow-disputes", true, "Allow disputed validations")
		minAccuracy       = flag.Float64("min-accuracy", 0.8, "Minimum accuracy threshold")
		includeEdgeCases  = flag.Bool("include-edge-cases", true, "Include edge cases")
		includeHighConf   = flag.Bool("include-high-confidence", true, "Include high confidence cases")
		includeLowConf    = flag.Bool("include-low-confidence", true, "Include low confidence cases")
		help              = flag.Bool("help", false, "Show help message")
	)

	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Load configuration
	config := loadConfiguration(*configFile, &test.ManualValidationConfig{
		SessionName:           *sessionName,
		ValidationDir:         *validationDir,
		SampleSize:            *sampleSize,
		ValidationTimeout:     *timeout,
		AutoSaveInterval:      *autoSave,
		RequireValidation:     *requireValidation,
		AllowDisputes:         *allowDisputes,
		MinAccuracyThreshold:  *minAccuracy,
		IncludeEdgeCases:      *includeEdgeCases,
		IncludeHighConfidence: *includeHighConf,
		IncludeLowConfidence:  *includeLowConf,
		ValidationFields:      []string{"industry", "mcc", "sic", "naics", "confidence"},
	})

	// Print configuration
	printConfiguration(config)

	// Create and run manual validation framework
	fmt.Println("🔍 Starting Manual Validation Framework...")
	fmt.Println()

	framework := test.NewManualValidationFramework(config)

	// Run the validation framework
	startTime := time.Now()

	err := framework.RunManualValidation()
	if err != nil {
		log.Fatalf("Manual validation failed: %v", err)
	}

	duration := time.Since(startTime)

	// Print final results
	printFinalResults(framework, duration)
}

// loadConfiguration loads configuration from file or uses defaults
func loadConfiguration(configFile string, defaultConfig *test.ManualValidationConfig) *test.ManualValidationConfig {
	if configFile == "" {
		return defaultConfig
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Printf("Warning: Failed to read config file %s: %v", configFile, err)
		return defaultConfig
	}

	var config test.ManualValidationConfig
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("Warning: Failed to parse config file %s: %v", configFile, err)
		return defaultConfig
	}

	// Merge with defaults for missing fields
	if config.SessionName == "" {
		config.SessionName = defaultConfig.SessionName
	}
	if config.ValidationDir == "" {
		config.ValidationDir = defaultConfig.ValidationDir
	}
	if config.ValidationTimeout == 0 {
		config.ValidationTimeout = defaultConfig.ValidationTimeout
	}

	return &config
}

// printConfiguration prints the validation configuration
func printConfiguration(config *test.ManualValidationConfig) {
	fmt.Println("📋 Manual Validation Configuration:")
	fmt.Printf("   Session Name: %s\n", config.SessionName)
	fmt.Printf("   Validation Directory: %s\n", config.ValidationDir)
	fmt.Printf("   Sample Size: %d\n", config.SampleSize)
	fmt.Printf("   Timeout: %v\n", config.ValidationTimeout)
	fmt.Printf("   Auto-save Interval: %v\n", config.AutoSaveInterval)
	fmt.Printf("   Require Validation: %v\n", config.RequireValidation)
	fmt.Printf("   Allow Disputes: %v\n", config.AllowDisputes)
	fmt.Printf("   Min Accuracy Threshold: %.2f\n", config.MinAccuracyThreshold)
	fmt.Printf("   Include Edge Cases: %v\n", config.IncludeEdgeCases)
	fmt.Printf("   Include High Confidence: %v\n", config.IncludeHighConfidence)
	fmt.Printf("   Include Low Confidence: %v\n", config.IncludeLowConfidence)
	fmt.Printf("   Validation Fields: %v\n", config.ValidationFields)
	fmt.Println()
}

// printFinalResults prints final validation results
func printFinalResults(framework *test.ManualValidationFramework, duration time.Duration) {
	fmt.Println()
	fmt.Println("🏁 Manual Validation Completed!")
	fmt.Printf("⏱️  Total Duration: %s\n", duration.String())
	fmt.Printf("📊 Session ID: %s\n", framework.Results.SessionID)
	fmt.Printf("📁 Validation Directory: %s\n", framework.ValidationDir)

	if framework.Results.Summary != nil {
		fmt.Printf("📋 Validation Summary:\n")
		fmt.Printf("   Total Cases: %d\n", framework.Results.Summary.TotalCases)
		fmt.Printf("   Validated: %d\n", framework.Results.Summary.ValidatedCases)
		fmt.Printf("   Pending: %d\n", framework.Results.Summary.PendingCases)
		fmt.Printf("   Disputed: %d\n", framework.Results.Summary.DisputedCases)
		fmt.Printf("   Overall Accuracy: %.2f%%\n", framework.Results.Summary.OverallAccuracy*100)
		fmt.Printf("   Industry Accuracy: %.2f%%\n", framework.Results.Summary.IndustryAccuracy*100)
		fmt.Printf("   Code Accuracy: %.2f%%\n", framework.Results.Summary.CodeAccuracy*100)
		fmt.Printf("   Confidence Accuracy: %.2f%%\n", framework.Results.Summary.ConfidenceAccuracy*100)

		fmt.Printf("📊 Issues Found:\n")
		fmt.Printf("   Critical: %d\n", framework.Results.Summary.CriticalIssues)
		fmt.Printf("   High: %d\n", framework.Results.Summary.HighIssues)
		fmt.Printf("   Medium: %d\n", framework.Results.Summary.MediumIssues)
		fmt.Printf("   Low: %d\n", framework.Results.Summary.LowIssues)
	}

	if len(framework.Results.Recommendations) > 0 {
		fmt.Printf("💡 Recommendations: %d suggestions generated\n", len(framework.Results.Recommendations))
		fmt.Println()
		fmt.Println("📋 Recommendations:")
		for i, rec := range framework.Results.Recommendations {
			fmt.Printf("   %d. %s\n", i+1, rec)
		}
	}

	fmt.Println()
	fmt.Printf("📁 Validation files generated in: %s\n", framework.ValidationDir)
	fmt.Println("📄 Reports available:")
	fmt.Printf("   - validation_report.json (comprehensive JSON report)\n")
	fmt.Printf("   - validation_report.html (human-readable HTML report)\n")
	fmt.Printf("   - validation_summary.json (session summary)\n")
	fmt.Printf("   - case_*.json (individual validation cases)\n")
}

// showHelp shows help message
func showHelp() {
	fmt.Println("KYB Manual Validation Framework")
	fmt.Println("===============================")
	fmt.Println()
	fmt.Println("Usage: manual-validator [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -config string")
	fmt.Println("        Path to validation configuration file (JSON)")
	fmt.Println("  -dir string")
	fmt.Println("        Validation directory (default: ./manual-validation)")
	fmt.Println("  -sample-size int")
	fmt.Println("        Number of sample business cases (default: 50)")
	fmt.Println("  -session string")
	fmt.Println("        Session name (default: Manual Validation Session)")
	fmt.Println("  -timeout duration")
	fmt.Println("        Validation timeout duration (default: 30m)")
	fmt.Println("  -auto-save duration")
	fmt.Println("        Auto-save interval (default: 5m)")
	fmt.Println("  -require-validation")
	fmt.Println("        Require manual validation (default: true)")
	fmt.Println("  -allow-disputes")
	fmt.Println("        Allow disputed validations (default: true)")
	fmt.Println("  -min-accuracy float")
	fmt.Println("        Minimum accuracy threshold (default: 0.8)")
	fmt.Println("  -include-edge-cases")
	fmt.Println("        Include edge cases (default: true)")
	fmt.Println("  -include-high-confidence")
	fmt.Println("        Include high confidence cases (default: true)")
	fmt.Println("  -include-low-confidence")
	fmt.Println("        Include low confidence cases (default: true)")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  manual-validator")
	fmt.Println("  manual-validator -sample-size 100 -dir ./validation-results")
	fmt.Println("  manual-validator -config validation-config.json")
	fmt.Println("  manual-validator -min-accuracy 0.9 -include-edge-cases=false")
	fmt.Println()
	fmt.Println("Configuration File Format:")
	fmt.Println("  {")
	fmt.Println("    \"session_name\": \"My Validation Session\",")
	fmt.Println("    \"validation_dir\": \"./my-validation\",")
	fmt.Println("    \"sample_size\": 100,")
	fmt.Println("    \"validation_timeout\": \"1h\",")
	fmt.Println("    \"auto_save_interval\": \"10m\",")
	fmt.Println("    \"require_validation\": true,")
	fmt.Println("    \"allow_disputes\": true,")
	fmt.Println("    \"min_accuracy_threshold\": 0.85,")
	fmt.Println("    \"include_edge_cases\": true,")
	fmt.Println("    \"include_high_confidence\": true,")
	fmt.Println("    \"include_low_confidence\": true,")
	fmt.Println("    \"validation_fields\": [\"industry\", \"mcc\", \"sic\", \"naics\", \"confidence\"]")
	fmt.Println("  }")
}
