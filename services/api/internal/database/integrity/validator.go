// Package integrity provides comprehensive database integrity validation
// for the KYB Platform Supabase database schema.
//
// This package implements a modular, extensible validation system that
// ensures data integrity across all database tables and relationships.
package integrity

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
)

// Validator provides comprehensive database integrity validation
type Validator struct {
	db     *sql.DB
	logger *log.Logger
	config *ValidationConfig
}

// ValidationConfig contains configuration for the validation process
type ValidationConfig struct {
	// Validation settings
	CheckForeignKeys     bool
	CheckDataTypes       bool
	CheckOrphanedRecords bool
	CheckDataConsistency bool

	// Performance settings
	BatchSize          int
	Timeout            time.Duration
	ParallelValidation bool

	// Reporting settings
	DetailedReporting bool
	IncludeStatistics bool
}

// ValidationResult represents the result of a validation check
type ValidationResult struct {
	CheckName     string                 `json:"check_name"`
	Status        ValidationStatus       `json:"status"`
	Message       string                 `json:"message"`
	Details       map[string]interface{} `json:"details,omitempty"`
	ErrorCount    int                    `json:"error_count,omitempty"`
	WarningCount  int                    `json:"warning_count,omitempty"`
	ExecutionTime time.Duration          `json:"execution_time"`
	Timestamp     time.Time              `json:"timestamp"`
}

// ValidationStatus represents the status of a validation check
type ValidationStatus string

const (
	StatusPassed  ValidationStatus = "passed"
	StatusFailed  ValidationStatus = "failed"
	StatusWarning ValidationStatus = "warning"
	StatusSkipped ValidationStatus = "skipped"
)

// IntegrityReport contains the complete validation report
type IntegrityReport struct {
	Summary           ValidationSummary  `json:"summary"`
	Results           []ValidationResult `json:"results"`
	Recommendations   []string           `json:"recommendations"`
	GeneratedAt       time.Time          `json:"generated_at"`
	DatabaseVersion   string             `json:"database_version,omitempty"`
	ValidationVersion string             `json:"validation_version"`
}

// ValidationSummary provides a high-level overview of validation results
type ValidationSummary struct {
	TotalChecks   int           `json:"total_checks"`
	PassedChecks  int           `json:"passed_checks"`
	FailedChecks  int           `json:"failed_checks"`
	WarningChecks int           `json:"warning_checks"`
	SkippedChecks int           `json:"skipped_checks"`
	TotalErrors   int           `json:"total_errors"`
	TotalWarnings int           `json:"total_warnings"`
	ExecutionTime time.Duration `json:"execution_time"`
}

// NewValidator creates a new database integrity validator
func NewValidator(db *sql.DB, logger *log.Logger, config *ValidationConfig) *Validator {
	if config == nil {
		config = &ValidationConfig{
			CheckForeignKeys:     true,
			CheckDataTypes:       true,
			CheckOrphanedRecords: true,
			CheckDataConsistency: true,
			BatchSize:            1000,
			Timeout:              30 * time.Minute,
			ParallelValidation:   true,
			DetailedReporting:    true,
			IncludeStatistics:    true,
		}
	}

	return &Validator{
		db:     db,
		logger: logger,
		config: config,
	}
}

// ValidateAll performs comprehensive database integrity validation
func (v *Validator) ValidateAll(ctx context.Context) (*IntegrityReport, error) {
	startTime := time.Now()
	v.logger.Printf("Starting comprehensive database integrity validation")

	report := &IntegrityReport{
		Results:           make([]ValidationResult, 0),
		Recommendations:   make([]string, 0),
		GeneratedAt:       time.Now(),
		ValidationVersion: "1.0.0",
	}

	// Get database version
	if version, err := v.getDatabaseVersion(ctx); err == nil {
		report.DatabaseVersion = version
	}

	// Define all validation checks
	checks := []ValidationCheck{
		// Foreign Key Constraints
		&ForeignKeyValidator{validator: v},

		// Data Type Validation
		&DataTypeValidator{validator: v},

		// Orphaned Records Check
		&OrphanedRecordsValidator{validator: v},

		// Data Consistency Check
		&DataConsistencyValidator{validator: v},

		// Table Structure Validation
		&TableStructureValidator{validator: v},

		// Index Validation
		&IndexValidator{validator: v},

		// Constraint Validation
		&ConstraintValidator{validator: v},
	}

	// Execute validation checks
	for _, check := range checks {
		if v.shouldRunCheck(check) {
			result, err := v.executeCheck(ctx, check)
			if err != nil {
				v.logger.Printf("Error executing check %s: %v", check.Name(), err)
				result = &ValidationResult{
					CheckName:     check.Name(),
					Status:        StatusFailed,
					Message:       fmt.Sprintf("Check execution failed: %v", err),
					ExecutionTime: time.Since(startTime),
					Timestamp:     time.Now(),
				}
			}
			report.Results = append(report.Results, *result)
		}
	}

	// Generate summary
	report.Summary = v.generateSummary(report.Results, time.Since(startTime))

	// Generate recommendations
	report.Recommendations = v.generateRecommendations(report.Results)

	v.logger.Printf("Database integrity validation completed in %v", time.Since(startTime))
	return report, nil
}

// ValidationCheck interface defines the contract for validation checks
type ValidationCheck interface {
	Name() string
	Description() string
	Execute(ctx context.Context) (*ValidationResult, error)
	Required() bool
}

// shouldRunCheck determines if a validation check should be executed
func (v *Validator) shouldRunCheck(check ValidationCheck) bool {
	// Always run required checks
	if check.Required() {
		return true
	}

	// Check configuration-based execution
	switch check.Name() {
	case "foreign_key_constraints":
		return v.config.CheckForeignKeys
	case "data_types":
		return v.config.CheckDataTypes
	case "orphaned_records":
		return v.config.CheckOrphanedRecords
	case "data_consistency":
		return v.config.CheckDataConsistency
	}

	return true
}

// executeCheck executes a single validation check
func (v *Validator) executeCheck(ctx context.Context, check ValidationCheck) (*ValidationResult, error) {
	startTime := time.Now()
	v.logger.Printf("Executing validation check: %s", check.Name())

	// Create timeout context for individual checks
	checkCtx, cancel := context.WithTimeout(ctx, v.config.Timeout)
	defer cancel()

	result, err := check.Execute(checkCtx)
	if err != nil {
		return nil, fmt.Errorf("check %s failed: %w", check.Name(), err)
	}

	result.ExecutionTime = time.Since(startTime)
	result.Timestamp = time.Now()

	v.logger.Printf("Validation check %s completed with status: %s", check.Name(), result.Status)
	return result, nil
}

// generateSummary creates a summary of validation results
func (v *Validator) generateSummary(results []ValidationResult, totalTime time.Duration) ValidationSummary {
	summary := ValidationSummary{
		TotalChecks:   len(results),
		ExecutionTime: totalTime,
	}

	for _, result := range results {
		switch result.Status {
		case StatusPassed:
			summary.PassedChecks++
		case StatusFailed:
			summary.FailedChecks++
		case StatusWarning:
			summary.WarningChecks++
		case StatusSkipped:
			summary.SkippedChecks++
		}

		summary.TotalErrors += result.ErrorCount
		summary.TotalWarnings += result.WarningCount
	}

	return summary
}

// generateRecommendations creates recommendations based on validation results
func (v *Validator) generateRecommendations(results []ValidationResult) []string {
	recommendations := make([]string, 0)

	for _, result := range results {
		if result.Status == StatusFailed || result.Status == StatusWarning {
			switch result.CheckName {
			case "foreign_key_constraints":
				recommendations = append(recommendations,
					"Review and fix foreign key constraint violations to ensure referential integrity")
			case "data_types":
				recommendations = append(recommendations,
					"Validate and correct data type mismatches to prevent runtime errors")
			case "orphaned_records":
				recommendations = append(recommendations,
					"Clean up orphaned records to maintain data consistency")
			case "data_consistency":
				recommendations = append(recommendations,
					"Review data consistency issues and implement data validation rules")
			case "table_structure":
				recommendations = append(recommendations,
					"Review table structure issues and consider schema updates")
			case "indexes":
				recommendations = append(recommendations,
					"Optimize database indexes for better query performance")
			case "constraints":
				recommendations = append(recommendations,
					"Review and update database constraints for data integrity")
			}
		}
	}

	// Add general recommendations
	if len(recommendations) == 0 {
		recommendations = append(recommendations,
			"Database integrity validation passed successfully. Continue regular monitoring.")
	} else {
		recommendations = append(recommendations,
			"Schedule regular integrity validation checks to maintain database health")
	}

	return recommendations
}

// getDatabaseVersion retrieves the database version information
func (v *Validator) getDatabaseVersion(ctx context.Context) (string, error) {
	var version string
	err := v.db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		return "", fmt.Errorf("failed to get database version: %w", err)
	}
	return version, nil
}

// Helper function to execute a query and return results
func (v *Validator) executeQuery(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := v.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	return rows, nil
}

// Helper function to execute a query and return a single row
func (v *Validator) executeQueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return v.db.QueryRowContext(ctx, query, args...)
}

// Helper function to check if a table exists
func (v *Validator) tableExists(ctx context.Context, tableName string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)
	`
	err := v.executeQueryRow(ctx, query, tableName).Scan(&exists)
	return exists, err
}

// Helper function to get table row count
func (v *Validator) getTableRowCount(ctx context.Context, tableName string) (int64, error) {
	var count int64
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", pq.QuoteIdentifier(tableName))
	err := v.executeQueryRow(ctx, query).Scan(&count)
	return count, err
}

// Helper function to format error messages
func (v *Validator) formatError(checkName, message string, err error) string {
	if err != nil {
		return fmt.Sprintf("%s: %s - %v", checkName, message, err)
	}
	return fmt.Sprintf("%s: %s", checkName, message)
}

// Helper function to create validation result
func (v *Validator) createResult(checkName string, status ValidationStatus, message string, details map[string]interface{}) *ValidationResult {
	result := &ValidationResult{
		CheckName: checkName,
		Status:    status,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}

	// Count errors and warnings from details
	if details != nil {
		if errorCount, ok := details["error_count"].(int); ok {
			result.ErrorCount = errorCount
		}
		if warningCount, ok := details["warning_count"].(int); ok {
			result.WarningCount = warningCount
		}
	}

	return result
}
