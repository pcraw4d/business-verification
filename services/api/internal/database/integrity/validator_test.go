// Package integrity provides comprehensive testing for database integrity validation
package integrity

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// TestValidator tests the main validator functionality
func TestValidator(t *testing.T) {
	// Skip if no database URL provided
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	// Connect to test database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Create validator
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	config := &ValidationConfig{
		CheckForeignKeys:     true,
		CheckDataTypes:       true,
		CheckOrphanedRecords: true,
		CheckDataConsistency: true,
		BatchSize:            100,
		Timeout:              5 * time.Minute,
		ParallelValidation:   false, // Disable for testing
		DetailedReporting:    true,
		IncludeStatistics:    true,
	}

	validator := NewValidator(db, logger, config)

	// Run validation
	report, err := validator.ValidateAll(ctx)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Verify report structure
	if report == nil {
		t.Fatal("Report is nil")
	}

	if report.Summary.TotalChecks == 0 {
		t.Error("No validation checks were executed")
	}

	if len(report.Results) == 0 {
		t.Error("No validation results found")
	}

	// Verify summary calculations
	expectedTotal := report.Summary.PassedChecks + report.Summary.FailedChecks +
		report.Summary.WarningChecks + report.Summary.SkippedChecks

	if report.Summary.TotalChecks != expectedTotal {
		t.Errorf("Summary total mismatch: expected %d, got %d",
			expectedTotal, report.Summary.TotalChecks)
	}

	t.Logf("Validation completed: %d checks, %d passed, %d failed, %d warnings",
		report.Summary.TotalChecks, report.Summary.PassedChecks,
		report.Summary.FailedChecks, report.Summary.WarningChecks)
}

// TestForeignKeyValidator tests foreign key validation
func TestForeignKeyValidator(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	validator := NewValidator(db, logger, nil)

	fkv := &ForeignKeyValidator{validator: validator}

	result, err := fkv.Execute(ctx)
	if err != nil {
		t.Fatalf("Foreign key validation failed: %v", err)
	}

	if result.CheckName != "foreign_key_constraints" {
		t.Errorf("Expected check name 'foreign_key_constraints', got '%s'", result.CheckName)
	}

	t.Logf("Foreign key validation result: %s - %s", result.Status, result.Message)
}

// TestDataTypeValidator tests data type validation
func TestDataTypeValidator(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	validator := NewValidator(db, logger, nil)

	dtv := &DataTypeValidator{validator: validator}

	result, err := dtv.Execute(ctx)
	if err != nil {
		t.Fatalf("Data type validation failed: %v", err)
	}

	if result.CheckName != "data_types" {
		t.Errorf("Expected check name 'data_types', got '%s'", result.CheckName)
	}

	t.Logf("Data type validation result: %s - %s", result.Status, result.Message)
}

// TestOrphanedRecordsValidator tests orphaned records validation
func TestOrphanedRecordsValidator(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	validator := NewValidator(db, logger, nil)

	orv := &OrphanedRecordsValidator{validator: validator}

	result, err := orv.Execute(ctx)
	if err != nil {
		t.Fatalf("Orphaned records validation failed: %v", err)
	}

	if result.CheckName != "orphaned_records" {
		t.Errorf("Expected check name 'orphaned_records', got '%s'", result.CheckName)
	}

	t.Logf("Orphaned records validation result: %s - %s", result.Status, result.Message)
}

// TestDataConsistencyValidator tests data consistency validation
func TestDataConsistencyValidator(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	validator := NewValidator(db, logger, nil)

	dcv := &DataConsistencyValidator{validator: validator}

	result, err := dcv.Execute(ctx)
	if err != nil {
		t.Fatalf("Data consistency validation failed: %v", err)
	}

	if result.CheckName != "data_consistency" {
		t.Errorf("Expected check name 'data_consistency', got '%s'", result.CheckName)
	}

	t.Logf("Data consistency validation result: %s - %s", result.Status, result.Message)
}

// TestTableStructureValidator tests table structure validation
func TestTableStructureValidator(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	validator := NewValidator(db, logger, nil)

	tsv := &TableStructureValidator{validator: validator}

	result, err := tsv.Execute(ctx)
	if err != nil {
		t.Fatalf("Table structure validation failed: %v", err)
	}

	if result.CheckName != "table_structure" {
		t.Errorf("Expected check name 'table_structure', got '%s'", result.CheckName)
	}

	t.Logf("Table structure validation result: %s - %s", result.Status, result.Message)
}

// TestIndexValidator tests index validation
func TestIndexValidator(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	validator := NewValidator(db, logger, nil)

	iv := &IndexValidator{validator: validator}

	result, err := iv.Execute(ctx)
	if err != nil {
		t.Fatalf("Index validation failed: %v", err)
	}

	if result.CheckName != "indexes" {
		t.Errorf("Expected check name 'indexes', got '%s'", result.CheckName)
	}

	t.Logf("Index validation result: %s - %s", result.Status, result.Message)
}

// TestConstraintValidator tests constraint validation
func TestConstraintValidator(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	validator := NewValidator(db, logger, nil)

	cv := &ConstraintValidator{validator: validator}

	result, err := cv.Execute(ctx)
	if err != nil {
		t.Fatalf("Constraint validation failed: %v", err)
	}

	if result.CheckName != "constraints" {
		t.Errorf("Expected check name 'constraints', got '%s'", result.CheckName)
	}

	t.Logf("Constraint validation result: %s - %s", result.Status, result.Message)
}

// TestValidationConfig tests validation configuration
func TestValidationConfig(t *testing.T) {
	config := &ValidationConfig{
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

	if !config.CheckForeignKeys {
		t.Error("CheckForeignKeys should be true")
	}

	if !config.CheckDataTypes {
		t.Error("CheckDataTypes should be true")
	}

	if !config.CheckOrphanedRecords {
		t.Error("CheckOrphanedRecords should be true")
	}

	if !config.CheckDataConsistency {
		t.Error("CheckDataConsistency should be true")
	}

	if config.BatchSize != 1000 {
		t.Errorf("Expected BatchSize 1000, got %d", config.BatchSize)
	}

	if config.Timeout != 30*time.Minute {
		t.Errorf("Expected Timeout 30m, got %v", config.Timeout)
	}

	if !config.ParallelValidation {
		t.Error("ParallelValidation should be true")
	}

	if !config.DetailedReporting {
		t.Error("DetailedReporting should be true")
	}

	if !config.IncludeStatistics {
		t.Error("IncludeStatistics should be true")
	}
}

// TestValidationStatus tests validation status constants
func TestValidationStatus(t *testing.T) {
	if StatusPassed != "passed" {
		t.Errorf("Expected StatusPassed 'passed', got '%s'", StatusPassed)
	}

	if StatusFailed != "failed" {
		t.Errorf("Expected StatusFailed 'failed', got '%s'", StatusFailed)
	}

	if StatusWarning != "warning" {
		t.Errorf("Expected StatusWarning 'warning', got '%s'", StatusWarning)
	}

	if StatusSkipped != "skipped" {
		t.Errorf("Expected StatusSkipped 'skipped', got '%s'", StatusSkipped)
	}
}

// BenchmarkValidator benchmarks the validator performance
func BenchmarkValidator(b *testing.B) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		b.Skip("TEST_DATABASE_URL not set, skipping benchmark tests")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		b.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	logger := log.New(os.Stdout, "[BENCH] ", log.LstdFlags)
	config := &ValidationConfig{
		CheckForeignKeys:     true,
		CheckDataTypes:       true,
		CheckOrphanedRecords: true,
		CheckDataConsistency: true,
		BatchSize:            1000,
		Timeout:              5 * time.Minute,
		ParallelValidation:   true,
		DetailedReporting:    false, // Disable for benchmarking
		IncludeStatistics:    false,
	}

	validator := NewValidator(db, logger, config)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateAll(ctx)
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}
