package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// ValidationResult represents the result of system validation
type ValidationResult struct {
	TestName  string    `json:"test_name"`
	Status    string    `json:"status"` // "PASS", "FAIL", "WARNING"
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// SystemValidation represents the overall validation results
type SystemValidation struct {
	ValidationResults []ValidationResult `json:"validation_results"`
	Summary           ValidationSummary  `json:"summary"`
	Timestamp         time.Time          `json:"timestamp"`
}

// ValidationSummary provides overall validation statistics
type ValidationSummary struct {
	TotalTests   int     `json:"total_tests"`
	PassedTests  int     `json:"passed_tests"`
	FailedTests  int     `json:"failed_tests"`
	WarningTests int     `json:"warning_tests"`
	PassRate     float64 `json:"pass_rate"`
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Build database connection string from environment variables
	dbURL := buildDatabaseURL()
	if dbURL == "" {
		log.Fatal("Database configuration is incomplete")
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("🔍 Starting enhanced keyword system validation...")

	// Perform validation tests
	validation, err := performSystemValidation(db)
	if err != nil {
		log.Fatalf("Failed to perform system validation: %v", err)
	}

	// Display results
	displayValidationResults(validation)

	// Save detailed results
	if err := saveValidationResults(validation); err != nil {
		log.Printf("Warning: Failed to save validation results: %v", err)
	}

	fmt.Println("✅ Enhanced keyword system validation completed successfully!")
}

// buildDatabaseURL builds a PostgreSQL connection string from environment variables
func buildDatabaseURL() string {
	// Check if DATABASE_URL is provided (Railway format)
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		return dbURL
	}

	// Build from individual environment variables
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	database := os.Getenv("DB_DATABASE")
	sslMode := os.Getenv("DB_SSL_MODE")

	if host == "" || username == "" || database == "" {
		return ""
	}

	if port == "" {
		port = "5432"
	}
	if sslMode == "" {
		sslMode = "require"
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, username, password, database, sslMode)
}

// performSystemValidation performs comprehensive system validation
func performSystemValidation(db *sql.DB) (*SystemValidation, error) {
	ctx := context.Background()

	validation := &SystemValidation{
		ValidationResults: make([]ValidationResult, 0),
		Timestamp:         time.Now(),
	}

	// Test 1: Database connectivity
	result := validateDatabaseConnectivity(ctx, db)
	validation.ValidationResults = append(validation.ValidationResults, result)

	// Test 2: Schema integrity
	result = validateSchemaIntegrity(ctx, db)
	validation.ValidationResults = append(validation.ValidationResults, result)

	// Test 3: Keyword coverage
	result = validateKeywordCoverage(ctx, db)
	validation.ValidationResults = append(validation.ValidationResults, result)

	// Test 4: Keyword weighting
	result = validateKeywordWeighting(ctx, db)
	validation.ValidationResults = append(validation.ValidationResults, result)

	// Test 5: Industry coverage
	result = validateIndustryCoverage(ctx, db)
	validation.ValidationResults = append(validation.ValidationResults, result)

	// Test 6: Performance validation
	result = validatePerformance(ctx, db)
	validation.ValidationResults = append(validation.ValidationResults, result)

	// Test 7: Data consistency
	result = validateDataConsistency(ctx, db)
	validation.ValidationResults = append(validation.ValidationResults, result)

	// Calculate summary
	validation.Summary = calculateValidationSummary(validation.ValidationResults)

	return validation, nil
}

// validateDatabaseConnectivity tests database connectivity
func validateDatabaseConnectivity(ctx context.Context, db *sql.DB) ValidationResult {
	start := time.Now()

	err := db.PingContext(ctx)
	duration := time.Since(start)

	if err != nil {
		return ValidationResult{
			TestName:  "Database Connectivity",
			Status:    "FAIL",
			Message:   "Database connection failed",
			Details:   fmt.Sprintf("Error: %v", err),
			Timestamp: time.Now(),
		}
	}

	status := "PASS"
	message := "Database connection successful"
	if duration > 1*time.Second {
		status = "WARNING"
		message = "Database connection slow"
	}

	return ValidationResult{
		TestName:  "Database Connectivity",
		Status:    status,
		Message:   message,
		Details:   fmt.Sprintf("Connection time: %v", duration),
		Timestamp: time.Now(),
	}
}

// validateSchemaIntegrity tests schema integrity
func validateSchemaIntegrity(ctx context.Context, db *sql.DB) ValidationResult {
	// Check if required tables exist
	tables := []string{"industries", "industry_keywords"}

	for _, table := range tables {
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`

		var exists bool
		err := db.QueryRowContext(ctx, query, table).Scan(&exists)
		if err != nil || !exists {
			return ValidationResult{
				TestName:  "Schema Integrity",
				Status:    "FAIL",
				Message:   "Required table missing",
				Details:   fmt.Sprintf("Table '%s' not found", table),
				Timestamp: time.Now(),
			}
		}
	}

	// Check if required columns exist
	columns := map[string][]string{
		"industries":        {"id", "name", "category"},
		"industry_keywords": {"id", "industry_id", "keyword", "weight", "is_active"},
	}

	for table, requiredColumns := range columns {
		for _, column := range requiredColumns {
			query := `SELECT EXISTS (
				SELECT FROM information_schema.columns 
				WHERE table_schema = 'public' 
				AND table_name = $1 
				AND column_name = $2
			)`

			var exists bool
			err := db.QueryRowContext(ctx, query, table, column).Scan(&exists)
			if err != nil || !exists {
				return ValidationResult{
					TestName:  "Schema Integrity",
					Status:    "FAIL",
					Message:   "Required column missing",
					Details:   fmt.Sprintf("Column '%s.%s' not found", table, column),
					Timestamp: time.Now(),
				}
			}
		}
	}

	return ValidationResult{
		TestName:  "Schema Integrity",
		Status:    "PASS",
		Message:   "Schema integrity validated",
		Details:   "All required tables and columns exist",
		Timestamp: time.Now(),
	}
}

// validateKeywordCoverage tests keyword coverage
func validateKeywordCoverage(ctx context.Context, db *sql.DB) ValidationResult {
	// Count total keywords
	var totalKeywords int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM industry_keywords WHERE is_active = true").Scan(&totalKeywords)
	if err != nil {
		return ValidationResult{
			TestName:  "Keyword Coverage",
			Status:    "FAIL",
			Message:   "Failed to count keywords",
			Details:   fmt.Sprintf("Error: %v", err),
			Timestamp: time.Now(),
		}
	}

	// Count industries with keywords
	var industriesWithKeywords int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT i.id) 
		FROM industries i 
		INNER JOIN industry_keywords ik ON i.id = ik.industry_id 
		WHERE ik.is_active = true
	`).Scan(&industriesWithKeywords)
	if err != nil {
		return ValidationResult{
			TestName:  "Keyword Coverage",
			Status:    "FAIL",
			Message:   "Failed to count industries with keywords",
			Details:   fmt.Sprintf("Error: %v", err),
			Timestamp: time.Now(),
		}
	}

	// Count total industries
	var totalIndustries int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM industries").Scan(&totalIndustries)
	if err != nil {
		return ValidationResult{
			TestName:  "Keyword Coverage",
			Status:    "FAIL",
			Message:   "Failed to count industries",
			Details:   fmt.Sprintf("Error: %v", err),
			Timestamp: time.Now(),
		}
	}

	coverageRate := float64(industriesWithKeywords) / float64(totalIndustries) * 100

	status := "PASS"
	message := "Keyword coverage adequate"
	if coverageRate < 50 {
		status = "FAIL"
		message = "Keyword coverage insufficient"
	} else if coverageRate < 80 {
		status = "WARNING"
		message = "Keyword coverage could be improved"
	}

	return ValidationResult{
		TestName: "Keyword Coverage",
		Status:   status,
		Message:  message,
		Details: fmt.Sprintf("Total keywords: %d, Industries with keywords: %d/%d (%.1f%%)",
			totalKeywords, industriesWithKeywords, totalIndustries, coverageRate),
		Timestamp: time.Now(),
	}
}

// validateKeywordWeighting tests keyword weighting
func validateKeywordWeighting(ctx context.Context, db *sql.DB) ValidationResult {
	// Check for keywords with invalid weights
	var invalidWeights int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM industry_keywords 
		WHERE is_active = true 
		AND (weight < 0 OR weight > 1)
	`).Scan(&invalidWeights)
	if err != nil {
		return ValidationResult{
			TestName:  "Keyword Weighting",
			Status:    "FAIL",
			Message:   "Failed to validate keyword weights",
			Details:   fmt.Sprintf("Error: %v", err),
			Timestamp: time.Now(),
		}
	}

	if invalidWeights > 0 {
		return ValidationResult{
			TestName:  "Keyword Weighting",
			Status:    "FAIL",
			Message:   "Invalid keyword weights found",
			Details:   fmt.Sprintf("Found %d keywords with weights outside 0-1 range", invalidWeights),
			Timestamp: time.Now(),
		}
	}

	// Check for keywords with zero weight
	var zeroWeights int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM industry_keywords 
		WHERE is_active = true 
		AND weight = 0
	`).Scan(&zeroWeights)
	if err != nil {
		return ValidationResult{
			TestName:  "Keyword Weighting",
			Status:    "FAIL",
			Message:   "Failed to check zero weights",
			Details:   fmt.Sprintf("Error: %v", err),
			Timestamp: time.Now(),
		}
	}

	status := "PASS"
	message := "Keyword weighting validated"
	if zeroWeights > 0 {
		status = "WARNING"
		message = "Keywords with zero weight found"
	}

	return ValidationResult{
		TestName:  "Keyword Weighting",
		Status:    status,
		Message:   message,
		Details:   fmt.Sprintf("Keywords with zero weight: %d", zeroWeights),
		Timestamp: time.Now(),
	}
}

// validateIndustryCoverage tests industry coverage
func validateIndustryCoverage(ctx context.Context, db *sql.DB) ValidationResult {
	// Check for industries with no keywords
	var industriesWithoutKeywords int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM industries i 
		LEFT JOIN industry_keywords ik ON i.id = ik.industry_id AND ik.is_active = true
		WHERE ik.id IS NULL
	`).Scan(&industriesWithoutKeywords)
	if err != nil {
		return ValidationResult{
			TestName:  "Industry Coverage",
			Status:    "FAIL",
			Message:   "Failed to check industry coverage",
			Details:   fmt.Sprintf("Error: %v", err),
			Timestamp: time.Now(),
		}
	}

	status := "PASS"
	message := "Industry coverage adequate"
	if industriesWithoutKeywords > 10 {
		status = "FAIL"
		message = "Too many industries without keywords"
	} else if industriesWithoutKeywords > 5 {
		status = "WARNING"
		message = "Some industries lack keyword coverage"
	}

	return ValidationResult{
		TestName:  "Industry Coverage",
		Status:    status,
		Message:   message,
		Details:   fmt.Sprintf("Industries without keywords: %d", industriesWithoutKeywords),
		Timestamp: time.Now(),
	}
}

// validatePerformance tests system performance
func validatePerformance(ctx context.Context, db *sql.DB) ValidationResult {
	// Test keyword lookup performance
	start := time.Now()

	_, err := db.QueryContext(ctx, `
		SELECT ik.industry_id, ik.keyword, ik.weight
		FROM industry_keywords ik
		WHERE ik.is_active = true
		LIMIT 100
	`)

	duration := time.Since(start)

	if err != nil {
		return ValidationResult{
			TestName:  "Performance",
			Status:    "FAIL",
			Message:   "Performance test failed",
			Details:   fmt.Sprintf("Error: %v", err),
			Timestamp: time.Now(),
		}
	}

	status := "PASS"
	message := "Performance adequate"
	if duration > 100*time.Millisecond {
		status = "WARNING"
		message = "Performance could be improved"
	}

	return ValidationResult{
		TestName:  "Performance",
		Status:    status,
		Message:   message,
		Details:   fmt.Sprintf("Query time: %v", duration),
		Timestamp: time.Now(),
	}
}

// validateDataConsistency tests data consistency
func validateDataConsistency(ctx context.Context, db *sql.DB) ValidationResult {
	// Check for orphaned keywords
	var orphanedKeywords int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM industry_keywords ik 
		LEFT JOIN industries i ON ik.industry_id = i.id
		WHERE i.id IS NULL
	`).Scan(&orphanedKeywords)
	if err != nil {
		return ValidationResult{
			TestName:  "Data Consistency",
			Status:    "FAIL",
			Message:   "Failed to check data consistency",
			Details:   fmt.Sprintf("Error: %v", err),
			Timestamp: time.Now(),
		}
	}

	if orphanedKeywords > 0 {
		return ValidationResult{
			TestName:  "Data Consistency",
			Status:    "FAIL",
			Message:   "Orphaned keywords found",
			Details:   fmt.Sprintf("Found %d keywords with invalid industry_id", orphanedKeywords),
			Timestamp: time.Now(),
		}
	}

	// Check for duplicate keywords per industry
	var totalKeywords, uniqueKeywords int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM industry_keywords
		WHERE is_active = true
	`).Scan(&totalKeywords)
	if err != nil {
		return ValidationResult{
			TestName:  "Data Consistency",
			Status:    "FAIL",
			Message:   "Failed to count total keywords",
			Details:   fmt.Sprintf("Error: %v", err),
			Timestamp: time.Now(),
		}
	}

	err = db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT industry_id, keyword)
		FROM industry_keywords
		WHERE is_active = true
	`).Scan(&uniqueKeywords)
	if err != nil {
		return ValidationResult{
			TestName:  "Data Consistency",
			Status:    "FAIL",
			Message:   "Failed to check for duplicates",
			Details:   fmt.Sprintf("Error: %v", err),
			Timestamp: time.Now(),
		}
	}

	duplicateKeywords := totalKeywords - uniqueKeywords

	status := "PASS"
	message := "Data consistency validated"
	if duplicateKeywords > 0 {
		status = "WARNING"
		message = "Duplicate keywords found"
	}

	return ValidationResult{
		TestName:  "Data Consistency",
		Status:    status,
		Message:   message,
		Details:   fmt.Sprintf("Duplicate keywords: %d", duplicateKeywords),
		Timestamp: time.Now(),
	}
}

// calculateValidationSummary calculates summary statistics
func calculateValidationSummary(results []ValidationResult) ValidationSummary {
	summary := ValidationSummary{
		TotalTests: len(results),
	}

	for _, result := range results {
		switch result.Status {
		case "PASS":
			summary.PassedTests++
		case "FAIL":
			summary.FailedTests++
		case "WARNING":
			summary.WarningTests++
		}
	}

	if summary.TotalTests > 0 {
		summary.PassRate = float64(summary.PassedTests) / float64(summary.TotalTests) * 100
	}

	return summary
}

// displayValidationResults displays the validation results
func displayValidationResults(validation *SystemValidation) {
	fmt.Println("\n📊 ENHANCED KEYWORD SYSTEM VALIDATION RESULTS")
	fmt.Println("==============================================")

	summary := validation.Summary
	fmt.Printf("Total Tests: %d\n", summary.TotalTests)
	fmt.Printf("Passed: %d\n", summary.PassedTests)
	fmt.Printf("Failed: %d\n", summary.FailedTests)
	fmt.Printf("Warnings: %d\n", summary.WarningTests)
	fmt.Printf("Pass Rate: %.1f%%\n", summary.PassRate)

	fmt.Println("\n📋 DETAILED RESULTS:")
	for _, result := range validation.ValidationResults {
		statusIcon := "✅"
		if result.Status == "FAIL" {
			statusIcon = "❌"
		} else if result.Status == "WARNING" {
			statusIcon = "⚠️"
		}

		fmt.Printf("%s %s: %s\n", statusIcon, result.TestName, result.Message)
		if result.Details != "" {
			fmt.Printf("   Details: %s\n", result.Details)
		}
	}

	// Overall assessment
	fmt.Println("\n🎯 OVERALL ASSESSMENT:")
	if summary.PassRate >= 90 {
		fmt.Println("   🎉 System validation PASSED - Ready for production")
	} else if summary.PassRate >= 70 {
		fmt.Println("   ⚠️  System validation PASSED with warnings - Review warnings")
	} else {
		fmt.Println("   ❌ System validation FAILED - Address issues before deployment")
	}
}

// saveValidationResults saves the validation results to a JSON file
func saveValidationResults(validation *SystemValidation) error {
	filename := fmt.Sprintf("keyword_system_validation_%s.json",
		time.Now().Format("2006-01-02"))

	data, err := json.MarshalIndent(validation, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
