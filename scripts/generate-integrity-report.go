package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// IntegrityReport represents a comprehensive data integrity report
type IntegrityReport struct {
	GeneratedAt     time.Time
	DatabaseURL     string
	Summary         ReportSummary
	ForeignKeys     []ForeignKeyResult
	DataTypes       []DataTypeResult
	OrphanedRecords []OrphanedRecordResult
	Consistency     []ConsistencyResult
	Recommendations []string
}

// ReportSummary provides an overview of all validation results
type ReportSummary struct {
	TotalTests       int
	PassedTests      int
	FailedTests      int
	ErrorTests       int
	SuccessRate      float64
	CriticalFailures int
	OverallStatus    string
}

// ForeignKeyResult represents foreign key validation results
type ForeignKeyResult struct {
	TableName        string
	ColumnName       string
	ReferencedTable  string
	ReferencedColumn string
	Status           string
	ErrorMessage     string
	OrphanedCount    int
	TotalRecords     int
	ExecutionTime    time.Duration
}

// DataTypeResult represents data type validation results
type DataTypeResult struct {
	TableName     string
	ColumnName    string
	DataType      string
	Status        string
	ErrorMessage  string
	InvalidCount  int
	TotalRecords  int
	SampleInvalid []string
	ExecutionTime time.Duration
}

// OrphanedRecordResult represents orphaned record detection results
type OrphanedRecordResult struct {
	ChildTable       string
	ChildColumn      string
	ParentTable      string
	ParentColumn     string
	RelationshipType string
	Status           string
	ErrorMessage     string
	OrphanedCount    int
	TotalRecords     int
	SampleOrphaned   []string
	ExecutionTime    time.Duration
}

// ConsistencyResult represents data consistency verification results
type ConsistencyResult struct {
	TestName       string
	Description    string
	TestType       string
	Status         string
	ErrorMessage   string
	ActualResult   interface{}
	ExpectedResult interface{}
	Difference     float64
	Critical       bool
	ExecutionTime  time.Duration
}

func main() {
	// Get database URL from environment or command line
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		if len(os.Args) > 1 {
			dbURL = os.Args[1]
		} else {
			log.Fatal("Database URL required. Set DATABASE_URL environment variable or pass as first argument")
		}
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("📊 Generating Comprehensive Data Integrity Report...")
	fmt.Println(strings.Repeat("=", 60))

	// Generate the comprehensive report
	report, err := generateIntegrityReport(ctx, db, dbURL)
	if err != nil {
		log.Fatalf("Failed to generate integrity report: %v", err)
	}

	// Print summary
	printReportSummary(report)

	// Generate detailed report files
	err = generateReportFiles(report)
	if err != nil {
		log.Fatalf("Failed to generate report files: %v", err)
	}

	fmt.Println("\n🎉 Comprehensive data integrity report generated successfully!")
}

// generateIntegrityReport creates a comprehensive data integrity report
func generateIntegrityReport(ctx context.Context, db *sql.DB, dbURL string) (*IntegrityReport, error) {
	report := &IntegrityReport{
		GeneratedAt: time.Now(),
		DatabaseURL: maskDatabaseURLForReport(dbURL),
	}

	// Run all validation tests
	fmt.Println("🔍 Running foreign key constraint tests...")
	fkResults, err := runForeignKeyTests(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to run foreign key tests: %w", err)
	}
	report.ForeignKeys = fkResults

	fmt.Println("🔍 Running data type validation tests...")
	dtResults, err := runDataTypeTests(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to run data type tests: %w", err)
	}
	report.DataTypes = dtResults

	fmt.Println("🔍 Running orphaned records detection tests...")
	orResults, err := runOrphanedRecordsTests(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to run orphaned records tests: %w", err)
	}
	report.OrphanedRecords = orResults

	fmt.Println("🔍 Running data consistency verification tests...")
	consistencyResults, err := runConsistencyTests(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to run consistency tests: %w", err)
	}
	report.Consistency = consistencyResults

	// Generate summary
	report.Summary = generateReportSummary(report)

	// Generate recommendations
	report.Recommendations = generateIntegrityRecommendations(report)

	return report, nil
}

// runForeignKeyTests runs foreign key constraint tests
func runForeignKeyTests(ctx context.Context, db *sql.DB) ([]ForeignKeyResult, error) {
	var results []ForeignKeyResult

	// Get foreign key relationships
	query := `
		SELECT 
			tc.table_name as child_table,
			kcu.column_name as child_column,
			ccu.table_name as parent_table,
			ccu.column_name as parent_column,
			tc.constraint_name
		FROM 
			information_schema.table_constraints AS tc 
			JOIN information_schema.key_column_usage AS kcu
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
			JOIN information_schema.constraint_column_usage AS ccu
				ON ccu.constraint_name = tc.constraint_name
				AND ccu.table_schema = tc.table_schema
		WHERE 
			tc.constraint_type = 'FOREIGN KEY'
			AND tc.table_schema = 'public'
		ORDER BY 
			tc.table_name, kcu.column_name
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query foreign key relationships: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var childTable, childColumn, parentTable, parentColumn, constraintName string
		err := rows.Scan(&childTable, &childColumn, &parentTable, &parentColumn, &constraintName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan foreign key relationship: %w", err)
		}

		// Test the foreign key constraint
		result := testForeignKeyConstraint(ctx, db, childTable, childColumn, parentTable, parentColumn)
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating foreign key relationships: %w", err)
	}

	return results, nil
}

// testForeignKeyConstraint tests a specific foreign key constraint
func testForeignKeyConstraint(ctx context.Context, db *sql.DB, childTable, childColumn, parentTable, parentColumn string) ForeignKeyResult {
	startTime := time.Now()
	result := ForeignKeyResult{
		TableName:        childTable,
		ColumnName:       childColumn,
		ReferencedTable:  parentTable,
		ReferencedColumn: parentColumn,
	}

	// Build query to find orphaned records
	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_records,
			COUNT(CASE WHEN parent.%s IS NULL THEN 1 END) as orphaned_records
		FROM %s child
		LEFT JOIN %s parent ON child.%s = parent.%s
		WHERE child.%s IS NOT NULL
	`,
		parentColumn,
		childTable,
		parentTable,
		childColumn,
		parentColumn,
		childColumn,
	)

	var totalRecords, orphanedRecords int
	err := db.QueryRowContext(ctx, query).Scan(&totalRecords, &orphanedRecords)
	if err != nil {
		result.Status = "ERROR"
		result.ErrorMessage = err.Error()
		result.ExecutionTime = time.Since(startTime)
		return result
	}

	result.TotalRecords = totalRecords
	result.OrphanedCount = orphanedRecords
	result.ExecutionTime = time.Since(startTime)

	if orphanedRecords > 0 {
		result.Status = "FAIL"
	} else {
		result.Status = "PASS"
	}

	return result
}

// runDataTypeTests runs data type validation tests
func runDataTypeTests(ctx context.Context, db *sql.DB) ([]DataTypeResult, error) {
	var results []DataTypeResult

	// Get columns to validate
	query := `
		SELECT 
			table_name,
			column_name,
			data_type,
			is_nullable,
			character_maximum_length
		FROM information_schema.columns
		WHERE table_schema = 'public'
		AND table_name IN ('users', 'merchants', 'business_verifications', 'classification_results', 'risk_assessments')
		ORDER BY table_name, column_name
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName, columnName, dataType, isNullable string
		var charLength sql.NullInt64
		err := rows.Scan(&tableName, &columnName, &dataType, &isNullable, &charLength)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		// Validate the column data
		result := validateColumnData(ctx, db, tableName, columnName, dataType, isNullable == "YES", charLength)
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating columns: %w", err)
	}

	return results, nil
}

// validateColumnData validates data in a specific column
func validateColumnData(ctx context.Context, db *sql.DB, tableName, columnName, dataType string, isNullable bool, charLength sql.NullInt64) DataTypeResult {
	startTime := time.Now()
	result := DataTypeResult{
		TableName:  tableName,
		ColumnName: columnName,
		DataType:   dataType,
	}

	// Build validation query based on data type
	var validationQuery string
	switch dataType {
	case "uuid":
		validationQuery = fmt.Sprintf(`
			SELECT 
				COUNT(*) as total_records,
				COUNT(CASE WHEN %s IS NOT NULL AND %s !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$' THEN 1 END) as invalid_records
			FROM %s
			WHERE %s IS NOT NULL
		`, columnName, columnName, tableName, columnName)
	case "character varying", "text":
		if charLength.Valid {
			validationQuery = fmt.Sprintf(`
				SELECT 
					COUNT(*) as total_records,
					COUNT(CASE WHEN LENGTH(%s) > %d THEN 1 END) as invalid_records
				FROM %s
				WHERE %s IS NOT NULL
			`, columnName, charLength.Int64, tableName, columnName)
		} else {
			validationQuery = fmt.Sprintf(`
				SELECT 
					COUNT(*) as total_records,
					0 as invalid_records
				FROM %s
				WHERE %s IS NOT NULL
			`, tableName, columnName)
		}
	case "integer", "bigint", "smallint":
		validationQuery = fmt.Sprintf(`
			SELECT 
				COUNT(*) as total_records,
				COUNT(CASE WHEN %s IS NOT NULL AND %s::text !~ '^-?[0-9]+$' THEN 1 END) as invalid_records
			FROM %s
			WHERE %s IS NOT NULL
		`, columnName, columnName, tableName, columnName)
	case "numeric", "decimal":
		validationQuery = fmt.Sprintf(`
			SELECT 
				COUNT(*) as total_records,
				COUNT(CASE WHEN %s IS NOT NULL AND %s::text !~ '^-?[0-9]+(\.[0-9]+)?$' THEN 1 END) as invalid_records
			FROM %s
			WHERE %s IS NOT NULL
		`, columnName, columnName, tableName, columnName)
	default:
		validationQuery = fmt.Sprintf(`
			SELECT 
				COUNT(*) as total_records,
				0 as invalid_records
			FROM %s
			WHERE %s IS NOT NULL
		`, tableName, columnName)
	}

	var totalRecords, invalidRecords int
	err := db.QueryRowContext(ctx, validationQuery).Scan(&totalRecords, &invalidRecords)
	if err != nil {
		result.Status = "ERROR"
		result.ErrorMessage = err.Error()
		result.ExecutionTime = time.Since(startTime)
		return result
	}

	result.TotalRecords = totalRecords
	result.InvalidCount = invalidRecords
	result.ExecutionTime = time.Since(startTime)

	if invalidRecords > 0 {
		result.Status = "FAIL"
	} else {
		result.Status = "PASS"
	}

	return result
}

// runOrphanedRecordsTests runs orphaned records detection tests
func runOrphanedRecordsTests(ctx context.Context, db *sql.DB) ([]OrphanedRecordResult, error) {
	var results []OrphanedRecordResult

	// Define logical relationships to check
	relationships := []struct {
		ChildTable       string
		ChildColumn      string
		ParentTable      string
		ParentColumn     string
		RelationshipType string
	}{
		{"merchants", "user_id", "users", "id", "logical"},
		{"business_verifications", "merchant_id", "merchants", "id", "logical"},
		{"classification_results", "merchant_id", "merchants", "id", "logical"},
		{"risk_assessments", "merchant_id", "merchants", "id", "logical"},
		{"audit_logs", "user_id", "users", "id", "logical"},
	}

	for _, rel := range relationships {
		result := checkOrphanedRecordsForReport(ctx, db, rel.ChildTable, rel.ChildColumn, rel.ParentTable, rel.ParentColumn, rel.RelationshipType)
		results = append(results, result)
	}

	return results, nil
}

// checkOrphanedRecordsForReport checks for orphaned records in a specific relationship
func checkOrphanedRecordsForReport(ctx context.Context, db *sql.DB, childTable, childColumn, parentTable, parentColumn, relationshipType string) OrphanedRecordResult {
	startTime := time.Now()
	result := OrphanedRecordResult{
		ChildTable:       childTable,
		ChildColumn:      childColumn,
		ParentTable:      parentTable,
		ParentColumn:     parentColumn,
		RelationshipType: relationshipType,
	}

	// Build query to find orphaned records
	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_records,
			COUNT(CASE WHEN parent.%s IS NULL THEN 1 END) as orphaned_records
		FROM %s child
		LEFT JOIN %s parent ON child.%s = parent.%s
		WHERE child.%s IS NOT NULL
	`,
		parentColumn,
		childTable,
		parentTable,
		childColumn,
		parentColumn,
		childColumn,
	)

	var totalRecords, orphanedRecords int
	err := db.QueryRowContext(ctx, query).Scan(&totalRecords, &orphanedRecords)
	if err != nil {
		result.Status = "ERROR"
		result.ErrorMessage = err.Error()
		result.ExecutionTime = time.Since(startTime)
		return result
	}

	result.TotalRecords = totalRecords
	result.OrphanedCount = orphanedRecords
	result.ExecutionTime = time.Since(startTime)

	if orphanedRecords > 0 {
		result.Status = "FAIL"
	} else {
		result.Status = "PASS"
	}

	return result
}

// runConsistencyTests runs data consistency verification tests
func runConsistencyTests(ctx context.Context, db *sql.DB) ([]ConsistencyResult, error) {
	var results []ConsistencyResult

	// Define consistency tests
	tests := []struct {
		TestName       string
		Description    string
		TestType       string
		Query          string
		ExpectedResult int
		Critical       bool
	}{
		{
			TestName:       "Business Verification Status Consistency",
			Description:    "Business verifications should have valid status values",
			TestType:       "custom",
			Query:          "SELECT COUNT(*) FROM business_verifications WHERE status NOT IN ('pending', 'approved', 'rejected', 'in_progress', 'completed')",
			ExpectedResult: 0,
			Critical:       true,
		},
		{
			TestName:       "Classification Confidence Score Consistency",
			Description:    "Classification results should have confidence scores between 0 and 1",
			TestType:       "custom",
			Query:          "SELECT COUNT(*) FROM classification_results WHERE confidence_score < 0 OR confidence_score > 1",
			ExpectedResult: 0,
			Critical:       true,
		},
		{
			TestName:       "Date Consistency - Created Before Updated",
			Description:    "Created dates should be before updated dates",
			TestType:       "custom",
			Query:          "SELECT COUNT(*) FROM merchants WHERE created_at > updated_at",
			ExpectedResult: 0,
			Critical:       true,
		},
	}

	for _, test := range tests {
		result := runConsistencyTestForReport(ctx, db, test.TestName, test.Description, test.TestType, test.Query, test.ExpectedResult, test.Critical)
		results = append(results, result)
	}

	return results, nil
}

// runConsistencyTestForReport runs a specific consistency test
func runConsistencyTestForReport(ctx context.Context, db *sql.DB, testName, description, testType, query string, expectedResult int, critical bool) ConsistencyResult {
	startTime := time.Now()
	result := ConsistencyResult{
		TestName:       testName,
		Description:    description,
		TestType:       testType,
		ExpectedResult: expectedResult,
		Critical:       critical,
	}

	var actualValue int
	err := db.QueryRowContext(ctx, query).Scan(&actualValue)
	if err != nil {
		result.Status = "ERROR"
		result.ErrorMessage = err.Error()
		result.ExecutionTime = time.Since(startTime)
		return result
	}

	result.ActualResult = actualValue
	result.ExecutionTime = time.Since(startTime)

	if actualValue == expectedResult {
		result.Status = "PASS"
	} else {
		result.Status = "FAIL"
		result.Difference = float64(actualValue - expectedResult)
	}

	return result
}

// generateReportSummary creates a summary of all validation results
func generateReportSummary(report *IntegrityReport) ReportSummary {
	var totalTests, passedTests, failedTests, errorTests, criticalFailures int

	// Count foreign key results
	for _, result := range report.ForeignKeys {
		totalTests++
		switch result.Status {
		case "PASS":
			passedTests++
		case "FAIL":
			failedTests++
		case "ERROR":
			errorTests++
		}
	}

	// Count data type results
	for _, result := range report.DataTypes {
		totalTests++
		switch result.Status {
		case "PASS":
			passedTests++
		case "FAIL":
			failedTests++
		case "ERROR":
			errorTests++
		}
	}

	// Count orphaned records results
	for _, result := range report.OrphanedRecords {
		totalTests++
		switch result.Status {
		case "PASS":
			passedTests++
		case "FAIL":
			failedTests++
		case "ERROR":
			errorTests++
		}
	}

	// Count consistency results
	for _, result := range report.Consistency {
		totalTests++
		switch result.Status {
		case "PASS":
			passedTests++
		case "FAIL":
			failedTests++
			if result.Critical {
				criticalFailures++
			}
		case "ERROR":
			errorTests++
		}
	}

	successRate := float64(passedTests) / float64(totalTests) * 100

	var overallStatus string
	if criticalFailures > 0 {
		overallStatus = "CRITICAL ISSUES FOUND"
	} else if failedTests > 0 {
		overallStatus = "ISSUES FOUND"
	} else if errorTests > 0 {
		overallStatus = "ERRORS ENCOUNTERED"
	} else {
		overallStatus = "ALL TESTS PASSED"
	}

	return ReportSummary{
		TotalTests:       totalTests,
		PassedTests:      passedTests,
		FailedTests:      failedTests,
		ErrorTests:       errorTests,
		SuccessRate:      successRate,
		CriticalFailures: criticalFailures,
		OverallStatus:    overallStatus,
	}
}

// generateIntegrityRecommendations generates recommendations based on validation results
func generateIntegrityRecommendations(report *IntegrityReport) []string {
	var recommendations []string

	// Check for foreign key issues
	for _, result := range report.ForeignKeys {
		if result.Status == "FAIL" {
			recommendations = append(recommendations,
				fmt.Sprintf("Fix orphaned records in %s.%s -> %s.%s (%d orphaned records)",
					result.TableName, result.ColumnName, result.ReferencedTable, result.ReferencedColumn, result.OrphanedCount))
		}
	}

	// Check for data type issues
	for _, result := range report.DataTypes {
		if result.Status == "FAIL" {
			recommendations = append(recommendations,
				fmt.Sprintf("Fix invalid data types in %s.%s (%d invalid records)",
					result.TableName, result.ColumnName, result.InvalidCount))
		}
	}

	// Check for orphaned records
	for _, result := range report.OrphanedRecords {
		if result.Status == "FAIL" {
			recommendations = append(recommendations,
				fmt.Sprintf("Clean up orphaned records in %s.%s -> %s.%s (%d orphaned records)",
					result.ChildTable, result.ChildColumn, result.ParentTable, result.ParentColumn, result.OrphanedCount))
		}
	}

	// Check for consistency issues
	for _, result := range report.Consistency {
		if result.Status == "FAIL" {
			recommendations = append(recommendations,
				fmt.Sprintf("Fix consistency issue: %s", result.Description))
		}
	}

	// Add general recommendations
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "No issues found - database integrity is excellent!")
	} else {
		recommendations = append(recommendations, "Consider implementing automated data integrity monitoring")
		recommendations = append(recommendations, "Add database constraints to prevent future integrity issues")
		recommendations = append(recommendations, "Implement application-level validation for data entry")
	}

	return recommendations
}

// printReportSummary prints a summary of the report
func printReportSummary(report *IntegrityReport) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 COMPREHENSIVE DATA INTEGRITY REPORT SUMMARY")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Generated At: %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Database: %s\n", report.DatabaseURL)
	fmt.Printf("Overall Status: %s\n", report.Summary.OverallStatus)
	fmt.Println()
	fmt.Printf("Total Tests: %d\n", report.Summary.TotalTests)
	fmt.Printf("✅ Passed: %d\n", report.Summary.PassedTests)
	fmt.Printf("❌ Failed: %d\n", report.Summary.FailedTests)
	fmt.Printf("🚨 Errors: %d\n", report.Summary.ErrorTests)
	fmt.Printf("Success Rate: %.1f%%\n", report.Summary.SuccessRate)

	if report.Summary.CriticalFailures > 0 {
		fmt.Printf("⚠️  Critical Failures: %d\n", report.Summary.CriticalFailures)
	}
}

// generateReportFiles generates detailed report files
func generateReportFiles(report *IntegrityReport) error {
	// Generate HTML report
	err := generateHTMLReport(report)
	if err != nil {
		return fmt.Errorf("failed to generate HTML report: %w", err)
	}

	// Generate JSON report
	err = generateJSONReport(report)
	if err != nil {
		return fmt.Errorf("failed to generate JSON report: %w", err)
	}

	// Generate Markdown report
	err = generateMarkdownReport(report)
	if err != nil {
		return fmt.Errorf("failed to generate Markdown report: %w", err)
	}

	return nil
}

// generateHTMLReport generates an HTML report
func generateHTMLReport(report *IntegrityReport) error {
	filename := fmt.Sprintf("data_integrity_report_%s.html", report.GeneratedAt.Format("20060102_150405"))
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write HTML header
	fmt.Fprintf(file, `<!DOCTYPE html>
<html>
<head>
    <title>Data Integrity Report - %s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; }
        .summary { background-color: #e8f5e8; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .failed { background-color: #ffe8e8; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .passed { background-color: #e8f5e8; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .error { background-color: #fff3cd; padding: 15px; border-radius: 5px; margin: 20px 0; }
        table { border-collapse: collapse; width: 100%%; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .status-pass { color: green; font-weight: bold; }
        .status-fail { color: red; font-weight: bold; }
        .status-error { color: orange; font-weight: bold; }
    </style>
</head>
<body>
    <div class="header">
        <h1>📊 Data Integrity Report</h1>
        <p><strong>Generated:</strong> %s</p>
        <p><strong>Database:</strong> %s</p>
        <p><strong>Overall Status:</strong> %s</p>
    </div>
`, report.GeneratedAt.Format("2006-01-02 15:04:05"), report.GeneratedAt.Format("2006-01-02 15:04:05"), report.DatabaseURL, report.Summary.OverallStatus)

	// Write summary
	fmt.Fprintf(file, `
    <div class="summary">
        <h2>📈 Summary</h2>
        <p><strong>Total Tests:</strong> %d</p>
        <p><strong>Passed:</strong> %d</p>
        <p><strong>Failed:</strong> %d</p>
        <p><strong>Errors:</strong> %d</p>
        <p><strong>Success Rate:</strong> %.1f%%</p>
`, report.Summary.TotalTests, report.Summary.PassedTests, report.Summary.FailedTests, report.Summary.ErrorTests, report.Summary.SuccessRate)

	if report.Summary.CriticalFailures > 0 {
		fmt.Fprintf(file, `        <p><strong>Critical Failures:</strong> %d</p>`, report.Summary.CriticalFailures)
	}

	fmt.Fprintf(file, `    </div>`)

	// Write recommendations
	if len(report.Recommendations) > 0 {
		fmt.Fprintf(file, `
    <div class="failed">
        <h2>🔧 Recommendations</h2>
        <ul>`)
		for _, rec := range report.Recommendations {
			fmt.Fprintf(file, `<li>%s</li>`, rec)
		}
		fmt.Fprintf(file, `        </ul>
    </div>`)
	}

	// Write detailed results
	fmt.Fprintf(file, `
    <h2>🔍 Detailed Results</h2>
    <h3>Foreign Key Constraints</h3>
    <table>
        <tr>
            <th>Child Table</th>
            <th>Child Column</th>
            <th>Parent Table</th>
            <th>Parent Column</th>
            <th>Status</th>
            <th>Orphaned Count</th>
            <th>Total Records</th>
        </tr>`)

	for _, result := range report.ForeignKeys {
		statusClass := "status-pass"
		if result.Status == "FAIL" {
			statusClass = "status-fail"
		} else if result.Status == "ERROR" {
			statusClass = "status-error"
		}
		fmt.Fprintf(file, `
        <tr>
            <td>%s</td>
            <td>%s</td>
            <td>%s</td>
            <td>%s</td>
            <td class="%s">%s</td>
            <td>%d</td>
            <td>%d</td>
        </tr>`, result.TableName, result.ColumnName, result.ReferencedTable, result.ReferencedColumn, statusClass, result.Status, result.OrphanedCount, result.TotalRecords)
	}

	fmt.Fprintf(file, `    </table>`)

	fmt.Fprintf(file, `
</body>
</html>`)

	fmt.Printf("📄 HTML report generated: %s\n", filename)
	return nil
}

// generateJSONReport generates a JSON report
func generateJSONReport(report *IntegrityReport) error {
	filename := fmt.Sprintf("data_integrity_report_%s.json", report.GeneratedAt.Format("20060102_150405"))
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Simple JSON output (in a real implementation, you'd use json.Marshal)
	fmt.Fprintf(file, `{
    "generated_at": "%s",
    "database_url": "%s",
    "summary": {
        "total_tests": %d,
        "passed_tests": %d,
        "failed_tests": %d,
        "error_tests": %d,
        "success_rate": %.1f,
        "critical_failures": %d,
        "overall_status": "%s"
    },
    "recommendations": [`,
		report.GeneratedAt.Format("2006-01-02T15:04:05Z"),
		report.DatabaseURL,
		report.Summary.TotalTests,
		report.Summary.PassedTests,
		report.Summary.FailedTests,
		report.Summary.ErrorTests,
		report.Summary.SuccessRate,
		report.Summary.CriticalFailures,
		report.Summary.OverallStatus)

	for i, rec := range report.Recommendations {
		if i > 0 {
			fmt.Fprintf(file, ",")
		}
		fmt.Fprintf(file, `"%s"`, rec)
	}

	fmt.Fprintf(file, `    ]
}`)

	fmt.Printf("📄 JSON report generated: %s\n", filename)
	return nil
}

// generateMarkdownReport generates a Markdown report
func generateMarkdownReport(report *IntegrityReport) error {
	filename := fmt.Sprintf("data_integrity_report_%s.md", report.GeneratedAt.Format("20060102_150405"))
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, `# Data Integrity Report

**Generated:** %s  
**Database:** %s  
**Overall Status:** %s

## Summary

- **Total Tests:** %d
- **Passed:** %d
- **Failed:** %d
- **Errors:** %d
- **Success Rate:** %.1f%%
`,
		report.GeneratedAt.Format("2006-01-02 15:04:05"),
		report.DatabaseURL,
		report.Summary.OverallStatus,
		report.Summary.TotalTests,
		report.Summary.PassedTests,
		report.Summary.FailedTests,
		report.Summary.ErrorTests,
		report.Summary.SuccessRate)

	if report.Summary.CriticalFailures > 0 {
		fmt.Fprintf(file, "- **Critical Failures:** %d\n", report.Summary.CriticalFailures)
	}

	fmt.Fprintf(file, `
## Recommendations

`)
	for _, rec := range report.Recommendations {
		fmt.Fprintf(file, "- %s\n", rec)
	}

	fmt.Fprintf(file, `
## Detailed Results

### Foreign Key Constraints

| Child Table | Child Column | Parent Table | Parent Column | Status | Orphaned Count | Total Records |
|-------------|--------------|--------------|---------------|--------|----------------|---------------|
`)

	for _, result := range report.ForeignKeys {
		fmt.Fprintf(file, "| %s | %s | %s | %s | %s | %d | %d |\n",
			result.TableName, result.ColumnName, result.ReferencedTable, result.ReferencedColumn, result.Status, result.OrphanedCount, result.TotalRecords)
	}

	fmt.Printf("📄 Markdown report generated: %s\n", filename)
	return nil
}

// maskDatabaseURLForReport masks sensitive information in database URL
func maskDatabaseURLForReport(dbURL string) string {
	// Simple masking - in production, you'd want more sophisticated masking
	if len(dbURL) > 20 {
		return dbURL[:10] + "****" + dbURL[len(dbURL)-6:]
	}
	return "****"
}
