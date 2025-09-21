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

// ForeignKeyTest represents a foreign key constraint test
type ForeignKeyTest struct {
	TableName        string
	ColumnName       string
	ReferencedTable  string
	ReferencedColumn string
	ConstraintName   string
}

// FKTestResult represents the result of a foreign key test
type FKTestResult struct {
	Test          ForeignKeyTest
	Status        string // "PASS", "FAIL", "ERROR"
	ErrorMessage  string
	OrphanedCount int
	TotalRecords  int
	ExecutionTime time.Duration
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

	fmt.Println("🔍 Starting Foreign Key Constraint Testing...")
	fmt.Println(strings.Repeat("=", 60))

	// Get all foreign key constraints
	fkTests, err := getForeignKeyConstraints(ctx, db)
	if err != nil {
		log.Fatalf("Failed to get foreign key constraints: %v", err)
	}

	fmt.Printf("Found %d foreign key constraints to test\n\n", len(fkTests))

	// Run tests
	var results []FKTestResult
	var passed, failed, errors int

	for i, test := range fkTests {
		fmt.Printf("[%d/%d] Testing %s.%s -> %s.%s\n",
			i+1, len(fkTests), test.TableName, test.ColumnName,
			test.ReferencedTable, test.ReferencedColumn)

		result := testForeignKeyConstraint(ctx, db, test)
		results = append(results, result)

		switch result.Status {
		case "PASS":
			passed++
			fmt.Printf("  ✅ PASS - No orphaned records found\n")
		case "FAIL":
			failed++
			fmt.Printf("  ❌ FAIL - Found %d orphaned records out of %d total\n",
				result.OrphanedCount, result.TotalRecords)
		case "ERROR":
			errors++
			fmt.Printf("  🚨 ERROR - %s\n", result.ErrorMessage)
		}

		fmt.Printf("  ⏱️  Execution time: %v\n\n", result.ExecutionTime)
	}

	// Print summary
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("📊 FOREIGN KEY CONSTRAINT TEST SUMMARY")
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
				fmt.Printf("\n❌ %s.%s -> %s.%s\n",
					result.Test.TableName, result.Test.ColumnName,
					result.Test.ReferencedTable, result.Test.ReferencedColumn)
				fmt.Printf("   Constraint: %s\n", result.Test.ConstraintName)

				if result.Status == "FAIL" {
					fmt.Printf("   Orphaned Records: %d out of %d total\n",
						result.OrphanedCount, result.TotalRecords)
					fmt.Printf("   Percentage: %.2f%%\n",
						float64(result.OrphanedCount)/float64(result.TotalRecords)*100)
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

	fmt.Println("\n🎉 All foreign key constraints are valid!")
}

// getForeignKeyConstraints retrieves all foreign key constraints from the database
func getForeignKeyConstraints(ctx context.Context, db *sql.DB) ([]ForeignKeyTest, error) {
	query := `
		SELECT 
			tc.table_name,
			kcu.column_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name,
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
		return nil, fmt.Errorf("failed to query foreign key constraints: %w", err)
	}
	defer rows.Close()

	var tests []ForeignKeyTest
	for rows.Next() {
		var test ForeignKeyTest
		err := rows.Scan(
			&test.TableName,
			&test.ColumnName,
			&test.ReferencedTable,
			&test.ReferencedColumn,
			&test.ConstraintName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan foreign key constraint: %w", err)
		}
		tests = append(tests, test)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating foreign key constraints: %w", err)
	}

	return tests, nil
}

// testForeignKeyConstraint tests a specific foreign key constraint
func testForeignKeyConstraint(ctx context.Context, db *sql.DB, test ForeignKeyTest) FKTestResult {
	startTime := time.Now()
	result := FKTestResult{
		Test: test,
	}

	// Build query to find orphaned records
	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_records,
			COUNT(CASE WHEN ref.%s IS NULL THEN 1 END) as orphaned_records
		FROM %s t
		LEFT JOIN %s ref ON t.%s = ref.%s
		WHERE t.%s IS NOT NULL
	`,
		test.ReferencedColumn,
		test.TableName,
		test.ReferencedTable,
		test.ColumnName,
		test.ReferencedColumn,
		test.ColumnName,
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
