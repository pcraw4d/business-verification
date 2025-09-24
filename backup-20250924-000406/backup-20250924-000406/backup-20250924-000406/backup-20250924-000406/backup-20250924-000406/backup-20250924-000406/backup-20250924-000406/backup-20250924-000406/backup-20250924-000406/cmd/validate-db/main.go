// Package main provides a command-line tool for database integrity validation
// for the KYB Platform Supabase database.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"kyb-platform/internal/database/integrity"

	_ "github.com/lib/pq"
)

func main() {
	// Command line flags
	var (
		dbURL   = flag.String("db-url", "", "Database connection URL (required)")
		output  = flag.String("output", "report.json", "Output file for validation report")
		verbose = flag.Bool("verbose", false, "Enable verbose logging")
		timeout = flag.Duration("timeout", 30*time.Minute, "Validation timeout")
		_       = flag.String("checks", "all", "Comma-separated list of checks to run (all, foreign_keys, data_types, orphaned_records, data_consistency, table_structure, indexes, constraints)")
	)

	flag.Parse()

	if *dbURL == "" {
		fmt.Fprintf(os.Stderr, "Error: Database URL is required\n")
		fmt.Fprintf(os.Stderr, "Usage: %s -db-url <connection_string> [options]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Set up logging
	logger := log.New(os.Stdout, "[DB-VALIDATOR] ", log.LstdFlags|log.Lshortfile)
	if !*verbose {
		logger.SetOutput(os.Stderr) // Only show errors in non-verbose mode
	}

	// Connect to database
	logger.Printf("Connecting to database...")
	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logger.Fatalf("Failed to ping database: %v", err)
	}

	logger.Printf("Database connection established successfully")

	// Create validator
	config := &integrity.ValidationConfig{
		CheckForeignKeys:     true,
		CheckDataTypes:       true,
		CheckOrphanedRecords: true,
		CheckDataConsistency: true,
		BatchSize:            1000,
		Timeout:              *timeout,
		ParallelValidation:   true,
		DetailedReporting:    true,
		IncludeStatistics:    true,
	}

	validator := integrity.NewValidator(db, logger, config)

	// Run validation
	logger.Printf("Starting database integrity validation...")
	startTime := time.Now()

	report, err := validator.ValidateAll(ctx)
	if err != nil {
		logger.Fatalf("Validation failed: %v", err)
	}

	duration := time.Since(startTime)
	logger.Printf("Validation completed in %v", duration)

	// Print summary
	printSummary(report, logger)

	// Save report
	if err := saveReport(report, *output); err != nil {
		logger.Fatalf("Failed to save report: %v", err)
	}

	logger.Printf("Validation report saved to %s", *output)

	// Exit with appropriate code
	if report.Summary.FailedChecks > 0 {
		logger.Printf("Validation failed with %d failed checks", report.Summary.FailedChecks)
		os.Exit(1)
	} else if report.Summary.WarningChecks > 0 {
		logger.Printf("Validation completed with %d warnings", report.Summary.WarningChecks)
		os.Exit(0)
	} else {
		logger.Printf("Validation passed successfully")
		os.Exit(0)
	}
}

// printSummary prints a summary of the validation results
func printSummary(report *integrity.IntegrityReport, logger *log.Logger) {
	logger.Printf("\n" + strings.Repeat("=", 60))
	logger.Printf("DATABASE INTEGRITY VALIDATION SUMMARY")
	logger.Printf(strings.Repeat("=", 60))

	logger.Printf("Generated At: %s", report.GeneratedAt.Format(time.RFC3339))
	logger.Printf("Database Version: %s", report.DatabaseVersion)
	logger.Printf("Validation Version: %s", report.ValidationVersion)
	logger.Printf("")

	logger.Printf("Validation Results:")
	logger.Printf("  Total Checks: %d", report.Summary.TotalChecks)
	logger.Printf("  Passed: %d", report.Summary.PassedChecks)
	logger.Printf("  Failed: %d", report.Summary.FailedChecks)
	logger.Printf("  Warnings: %d", report.Summary.WarningChecks)
	logger.Printf("  Skipped: %d", report.Summary.SkippedChecks)
	logger.Printf("")

	logger.Printf("Issues Found:")
	logger.Printf("  Total Errors: %d", report.Summary.TotalErrors)
	logger.Printf("  Total Warnings: %d", report.Summary.TotalWarnings)
	logger.Printf("")

	logger.Printf("Execution Time: %v", report.Summary.ExecutionTime)
	logger.Printf("")

	// Print failed checks
	if report.Summary.FailedChecks > 0 {
		logger.Printf("FAILED CHECKS:")
		for _, result := range report.Results {
			if result.Status == integrity.StatusFailed {
				logger.Printf("  ❌ %s: %s", result.CheckName, result.Message)
				if result.ErrorCount > 0 {
					logger.Printf("     Errors: %d", result.ErrorCount)
				}
			}
		}
		logger.Printf("")
	}

	// Print warning checks
	if report.Summary.WarningChecks > 0 {
		logger.Printf("WARNING CHECKS:")
		for _, result := range report.Results {
			if result.Status == integrity.StatusWarning {
				logger.Printf("  ⚠️  %s: %s", result.CheckName, result.Message)
				if result.WarningCount > 0 {
					logger.Printf("     Warnings: %d", result.WarningCount)
				}
			}
		}
		logger.Printf("")
	}

	// Print recommendations
	if len(report.Recommendations) > 0 {
		logger.Printf("RECOMMENDATIONS:")
		for i, rec := range report.Recommendations {
			logger.Printf("  %d. %s", i+1, rec)
		}
		logger.Printf("")
	}

	logger.Printf(strings.Repeat("=", 60))
}

// saveReport saves the validation report to a file
func saveReport(report *integrity.IntegrityReport, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to encode report: %w", err)
	}

	return nil
}
