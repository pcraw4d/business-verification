package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// DataTypeTest represents a data type validation test
type DataTypeTest struct {
	TableName   string
	ColumnName  string
	DataType    string
	IsNullable  bool
	MaxLength   *int
	Pattern     string
	Description string
}

// DataTypeValidationResult represents the result of a data type validation
type DataTypeValidationResult struct {
	Test          DataTypeTest
	Status        string // "PASS", "FAIL", "ERROR"
	ErrorMessage  string
	InvalidCount  int
	TotalRecords  int
	SampleInvalid []string
	ExecutionTime time.Duration
}

// Common validation patterns
var validationPatterns = map[string]string{
	"email":     `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
	"uuid":      `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
	"phone":     `^\+?[1-9]\d{1,14}$`,
	"url":       `^https?://[^\s/$.?#].[^\s]*$`,
	"date":      `^\d{4}-\d{2}-\d{2}$`,
	"datetime":  `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`,
	"timestamp": `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`,
	"ipv4":      `^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`,
	"ipv6":      `^(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$`,
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

	fmt.Println("🔍 Starting Data Type and Format Validation...")
	fmt.Println(strings.Repeat("=", 60))

	// Get all columns and their data types
	columns, err := getTableColumns(ctx, db)
	if err != nil {
		log.Fatalf("Failed to get table columns: %v", err)
	}

	fmt.Printf("Found %d columns to validate\n\n", len(columns))

	// Run validation tests
	var results []DataTypeValidationResult
	var passed, failed, errors int

	for i, column := range columns {
		fmt.Printf("[%d/%d] Validating %s.%s (%s)\n",
			i+1, len(columns), column.TableName, column.ColumnName, column.DataType)

		result := validateDataType(ctx, db, column)
		results = append(results, result)

		switch result.Status {
		case "PASS":
			passed++
			fmt.Printf("  ✅ PASS - All values are valid\n")
		case "FAIL":
			failed++
			fmt.Printf("  ❌ FAIL - Found %d invalid values out of %d total\n",
				result.InvalidCount, result.TotalRecords)
			if len(result.SampleInvalid) > 0 {
				fmt.Printf("  📝 Sample invalid values: %s\n", strings.Join(result.SampleInvalid[:minInt(3, len(result.SampleInvalid))], ", "))
			}
		case "ERROR":
			errors++
			fmt.Printf("  🚨 ERROR - %s\n", result.ErrorMessage)
		}

		fmt.Printf("  ⏱️  Execution time: %v\n\n", result.ExecutionTime)
	}

	// Print summary
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("📊 DATA TYPE VALIDATION SUMMARY")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total Tests: %d\n", len(results))
	fmt.Printf("✅ Passed: %d\n", passed)
	fmt.Printf("❌ Failed: %d\n", failed)
	fmt.Printf("🚨 Errors: %d\n", errors)
	fmt.Printf("Success Rate: %.1f%%\n", float64(passed)/float64(len(results))*100)

	// Print detailed results for failures
	if failed > 0 || errors > 0 {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("🔍 DETAILED FAILURE REPORT")
		fmt.Println(strings.Repeat("=", 60))

		for _, result := range results {
			if result.Status != "PASS" {
				fmt.Printf("\n❌ %s.%s (%s)\n",
					result.Test.TableName, result.Test.ColumnName, result.Test.DataType)
				fmt.Printf("   Description: %s\n", result.Test.Description)

				if result.Status == "FAIL" {
					fmt.Printf("   Invalid Values: %d out of %d total\n",
						result.InvalidCount, result.TotalRecords)
					fmt.Printf("   Percentage: %.2f%%\n",
						float64(result.InvalidCount)/float64(result.TotalRecords)*100)
					if len(result.SampleInvalid) > 0 {
						fmt.Printf("   Sample Invalid: %s\n", strings.Join(result.SampleInvalid, ", "))
					}
				} else {
					fmt.Printf("   Error: %s\n", result.ErrorMessage)
				}
			}
		}
	}

	// Exit with appropriate code
	if failed > 0 || errors > 0 {
		os.Exit(1)
	}

	fmt.Println("\n🎉 All data types and formats are valid!")
}

// getTableColumns retrieves all columns from all tables
func getTableColumns(ctx context.Context, db *sql.DB) ([]DataTypeTest, error) {
	query := `
		SELECT 
			t.table_name,
			c.column_name,
			c.data_type,
			c.is_nullable = 'YES' as is_nullable,
			c.character_maximum_length,
			c.numeric_precision,
			c.numeric_scale
		FROM 
			information_schema.tables t
			JOIN information_schema.columns c ON t.table_name = c.table_name
		WHERE 
			t.table_schema = 'public'
			AND t.table_type = 'BASE TABLE'
			AND c.table_schema = 'public'
		ORDER BY 
			t.table_name, c.ordinal_position
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table columns: %w", err)
	}
	defer rows.Close()

	var columns []DataTypeTest
	for rows.Next() {
		var tableName, columnName, dataType, isNullable string
		var maxLength, precision, scale sql.NullInt64

		err := rows.Scan(
			&tableName,
			&columnName,
			&dataType,
			&isNullable,
			&maxLength,
			&precision,
			&scale,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		// Determine validation pattern based on column name and data type
		pattern := getValidationPattern(columnName, dataType)
		description := getColumnDescription(columnName, dataType, maxLength)

		column := DataTypeTest{
			TableName:   tableName,
			ColumnName:  columnName,
			DataType:    dataType,
			IsNullable:  isNullable == "YES",
			Description: description,
			Pattern:     pattern,
		}

		if maxLength.Valid {
			maxLen := int(maxLength.Int64)
			column.MaxLength = &maxLen
		}

		columns = append(columns, column)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating columns: %w", err)
	}

	return columns, nil
}

// getValidationPattern determines the appropriate validation pattern for a column
func getValidationPattern(columnName, dataType string) string {
	columnName = strings.ToLower(columnName)

	// Check for specific patterns based on column name
	if strings.Contains(columnName, "email") {
		return validationPatterns["email"]
	}
	if strings.Contains(columnName, "uuid") || strings.Contains(columnName, "id") && dataType == "uuid" {
		return validationPatterns["uuid"]
	}
	if strings.Contains(columnName, "phone") {
		return validationPatterns["phone"]
	}
	if strings.Contains(columnName, "url") || strings.Contains(columnName, "website") {
		return validationPatterns["url"]
	}
	if strings.Contains(columnName, "date") && dataType == "date" {
		return validationPatterns["date"]
	}
	if strings.Contains(columnName, "created_at") || strings.Contains(columnName, "updated_at") {
		return validationPatterns["timestamp"]
	}
	if strings.Contains(columnName, "ip") {
		return validationPatterns["ipv4"] // Default to IPv4, could be enhanced
	}

	// Return empty pattern for columns that don't need format validation
	return ""
}

// getColumnDescription generates a description for the column
func getColumnDescription(columnName, dataType string, maxLength sql.NullInt64) string {
	desc := fmt.Sprintf("%s (%s)", columnName, dataType)
	if maxLength.Valid {
		desc += fmt.Sprintf(" max length: %d", maxLength.Int64)
	}
	return desc
}

// validateDataType validates a specific column's data type and format
func validateDataType(ctx context.Context, db *sql.DB, test DataTypeTest) DataTypeValidationResult {
	startTime := time.Now()
	result := DataTypeValidationResult{
		Test: test,
	}

	// Skip validation for certain data types that don't need format checking
	if shouldSkipValidation(test.DataType) {
		result.Status = "PASS"
		result.ExecutionTime = time.Since(startTime)
		return result
	}

	// Build validation query
	query := buildValidationQuery(test)
	if query == "" {
		result.Status = "PASS"
		result.ExecutionTime = time.Since(startTime)
		return result
	}

	// Execute validation
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		result.Status = "ERROR"
		result.ErrorMessage = err.Error()
		result.ExecutionTime = time.Since(startTime)
		return result
	}
	defer rows.Close()

	var invalidValues []string
	var totalCount, invalidCount int

	for rows.Next() {
		var value sql.NullString
		err := rows.Scan(&value)
		if err != nil {
			result.Status = "ERROR"
			result.ErrorMessage = err.Error()
			result.ExecutionTime = time.Since(startTime)
			return result
		}

		totalCount++
		if !value.Valid {
			invalidCount++
			if len(invalidValues) < 5 { // Keep sample of invalid values
				invalidValues = append(invalidValues, "NULL")
			}
		} else if !isValidValue(value.String, test) {
			invalidCount++
			if len(invalidValues) < 5 {
				invalidValues = append(invalidValues, value.String)
			}
		}
	}

	if err := rows.Err(); err != nil {
		result.Status = "ERROR"
		result.ErrorMessage = err.Error()
		result.ExecutionTime = time.Since(startTime)
		return result
	}

	result.TotalRecords = totalCount
	result.InvalidCount = invalidCount
	result.SampleInvalid = invalidValues
	result.ExecutionTime = time.Since(startTime)

	if invalidCount > 0 {
		result.Status = "FAIL"
	} else {
		result.Status = "PASS"
	}

	return result
}

// shouldSkipValidation determines if a data type should skip validation
func shouldSkipValidation(dataType string) bool {
	skipTypes := []string{
		"integer", "bigint", "smallint", "decimal", "numeric", "real", "double precision",
		"boolean", "text", "bytea", "json", "jsonb", "array",
	}

	for _, skipType := range skipTypes {
		if strings.Contains(dataType, skipType) {
			return true
		}
	}
	return false
}

// buildValidationQuery builds the appropriate validation query for a column
func buildValidationQuery(test DataTypeTest) string {
	// For columns with specific patterns, we need to validate format
	if test.Pattern != "" {
		return fmt.Sprintf(`
			SELECT %s 
			FROM %s 
			WHERE %s IS NOT NULL
		`, test.ColumnName, test.TableName, test.ColumnName)
	}

	// For varchar columns, check length constraints
	if strings.Contains(test.DataType, "character varying") && test.MaxLength != nil {
		return fmt.Sprintf(`
			SELECT %s 
			FROM %s 
			WHERE %s IS NOT NULL 
			AND LENGTH(%s) > %d
		`, test.ColumnName, test.TableName, test.ColumnName, test.ColumnName, *test.MaxLength)
	}

	// For timestamp columns, check for valid timestamps
	if strings.Contains(test.DataType, "timestamp") {
		return fmt.Sprintf(`
			SELECT %s::text 
			FROM %s 
			WHERE %s IS NOT NULL
		`, test.ColumnName, test.TableName, test.ColumnName)
	}

	return ""
}

// isValidValue checks if a value matches the expected pattern
func isValidValue(value string, test DataTypeTest) bool {
	// If no pattern specified, assume valid
	if test.Pattern == "" {
		return true
	}

	// Check length constraints
	if test.MaxLength != nil && len(value) > *test.MaxLength {
		return false
	}

	// Check pattern matching
	matched, err := regexp.MatchString(test.Pattern, value)
	if err != nil {
		return false
	}

	return matched
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
