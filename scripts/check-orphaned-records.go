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

// OrphanedRecordTest represents an orphaned record detection test
type OrphanedRecordTest struct {
	ChildTable       string
	ChildColumn      string
	ParentTable      string
	ParentColumn     string
	RelationshipType string // "foreign_key", "logical", "business"
	Description      string
}

// OrphanedRecordResult represents the result of an orphaned record test
type OrphanedRecordResult struct {
	Test           OrphanedRecordTest
	Status         string // "PASS", "FAIL", "ERROR"
	ErrorMessage   string
	OrphanedCount  int
	TotalRecords   int
	SampleOrphaned []string
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

	fmt.Println("🔍 Starting Orphaned Records Detection...")
	fmt.Println(strings.Repeat("=", 60))

	// Get all relationships to check
	relationships, err := getRelationshipsToCheck(ctx, db)
	if err != nil {
		log.Fatalf("Failed to get relationships: %v", err)
	}

	fmt.Printf("Found %d relationships to check for orphaned records\n\n", len(relationships))

	// Run orphaned record detection tests
	var results []OrphanedRecordResult
	var passed, failed, errors int

	for i, relationship := range relationships {
		fmt.Printf("[%d/%d] Checking %s.%s -> %s.%s (%s)\n",
			i+1, len(relationships), relationship.ChildTable, relationship.ChildColumn,
			relationship.ParentTable, relationship.ParentColumn, relationship.RelationshipType)

		result := checkOrphanedRecords(ctx, db, relationship)
		results = append(results, result)

		switch result.Status {
		case "PASS":
			passed++
			fmt.Printf("  ✅ PASS - No orphaned records found\n")
		case "FAIL":
			failed++
			fmt.Printf("  ❌ FAIL - Found %d orphaned records out of %d total\n",
				result.OrphanedCount, result.TotalRecords)
			if len(result.SampleOrphaned) > 0 {
				fmt.Printf("  📝 Sample orphaned values: %s\n", strings.Join(result.SampleOrphaned[:minIntOrphaned(3, len(result.SampleOrphaned))], ", "))
			}
		case "ERROR":
			errors++
			fmt.Printf("  🚨 ERROR - %s\n", result.ErrorMessage)
		}

		fmt.Printf("  ⏱️  Execution time: %v\n\n", result.ExecutionTime)
	}

	// Print summary
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("📊 ORPHANED RECORDS DETECTION SUMMARY")
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
				fmt.Printf("\n❌ %s.%s -> %s.%s (%s)\n",
					result.Test.ChildTable, result.Test.ChildColumn,
					result.Test.ParentTable, result.Test.ParentColumn, result.Test.RelationshipType)
				fmt.Printf("   Description: %s\n", result.Test.Description)

				if result.Status == "FAIL" {
					fmt.Printf("   Orphaned Records: %d out of %d total\n",
						result.OrphanedCount, result.TotalRecords)
					fmt.Printf("   Percentage: %.2f%%\n",
						float64(result.OrphanedCount)/float64(result.TotalRecords)*100)
					if len(result.SampleOrphaned) > 0 {
						fmt.Printf("   Sample Orphaned: %s\n", strings.Join(result.SampleOrphaned, ", "))
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

	fmt.Println("\n🎉 No orphaned records found!")
}

// getRelationshipsToCheck retrieves all relationships that should be checked for orphaned records
func getRelationshipsToCheck(ctx context.Context, db *sql.DB) ([]OrphanedRecordTest, error) {
	var relationships []OrphanedRecordTest

	// Get foreign key relationships
	fkRelationships, err := getForeignKeyRelationships(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to get foreign key relationships: %w", err)
	}
	relationships = append(relationships, fkRelationships...)

	// Get logical relationships (business relationships that should exist)
	logicalRelationships, err := getLogicalRelationships(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to get logical relationships: %w", err)
	}
	relationships = append(relationships, logicalRelationships...)

	return relationships, nil
}

// getForeignKeyRelationships retrieves foreign key relationships from the database
func getForeignKeyRelationships(ctx context.Context, db *sql.DB) ([]OrphanedRecordTest, error) {
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

	var relationships []OrphanedRecordTest
	for rows.Next() {
		var childTable, childColumn, parentTable, parentColumn, constraintName string
		err := rows.Scan(&childTable, &childColumn, &parentTable, &parentColumn, &constraintName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan foreign key relationship: %w", err)
		}

		relationship := OrphanedRecordTest{
			ChildTable:       childTable,
			ChildColumn:      childColumn,
			ParentTable:      parentTable,
			ParentColumn:     parentColumn,
			RelationshipType: "foreign_key",
			Description:      fmt.Sprintf("Foreign key constraint: %s", constraintName),
		}
		relationships = append(relationships, relationship)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating foreign key relationships: %w", err)
	}

	return relationships, nil
}

// getLogicalRelationships defines logical business relationships that should be checked
func getLogicalRelationships(ctx context.Context, db *sql.DB) ([]OrphanedRecordTest, error) {
	var relationships []OrphanedRecordTest

	// Check if tables exist before adding relationships
	tables, err := getExistingTables(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing tables: %w", err)
	}

	// Define logical relationships based on common business patterns
	logicalRels := []OrphanedRecordTest{
		// User-related relationships
		{
			ChildTable:       "merchants",
			ChildColumn:      "user_id",
			ParentTable:      "users",
			ParentColumn:     "id",
			RelationshipType: "logical",
			Description:      "Merchants should reference existing users",
		},
		{
			ChildTable:       "business_verifications",
			ChildColumn:      "user_id",
			ParentTable:      "users",
			ParentColumn:     "id",
			RelationshipType: "logical",
			Description:      "Business verifications should reference existing users",
		},
		{
			ChildTable:       "audit_logs",
			ChildColumn:      "user_id",
			ParentTable:      "users",
			ParentColumn:     "id",
			RelationshipType: "logical",
			Description:      "Audit logs should reference existing users",
		},
		// Merchant-related relationships
		{
			ChildTable:       "business_verifications",
			ChildColumn:      "merchant_id",
			ParentTable:      "merchants",
			ParentColumn:     "id",
			RelationshipType: "logical",
			Description:      "Business verifications should reference existing merchants",
		},
		{
			ChildTable:       "classification_results",
			ChildColumn:      "merchant_id",
			ParentTable:      "merchants",
			ParentColumn:     "id",
			RelationshipType: "logical",
			Description:      "Classification results should reference existing merchants",
		},
		{
			ChildTable:       "risk_assessments",
			ChildColumn:      "merchant_id",
			ParentTable:      "merchants",
			ParentColumn:     "id",
			RelationshipType: "logical",
			Description:      "Risk assessments should reference existing merchants",
		},
		{
			ChildTable:       "merchant_audit_logs",
			ChildColumn:      "merchant_id",
			ParentTable:      "merchants",
			ParentColumn:     "id",
			RelationshipType: "logical",
			Description:      "Merchant audit logs should reference existing merchants",
		},
		// Industry-related relationships
		{
			ChildTable:       "classification_results",
			ChildColumn:      "industry_id",
			ParentTable:      "industries",
			ParentColumn:     "id",
			RelationshipType: "logical",
			Description:      "Classification results should reference existing industries",
		},
		{
			ChildTable:       "industry_keywords",
			ChildColumn:      "industry_id",
			ParentTable:      "industries",
			ParentColumn:     "id",
			RelationshipType: "logical",
			Description:      "Industry keywords should reference existing industries",
		},
		// Risk-related relationships
		{
			ChildTable:       "business_risk_assessments",
			ChildColumn:      "risk_keyword_id",
			ParentTable:      "risk_keywords",
			ParentColumn:     "id",
			RelationshipType: "logical",
			Description:      "Business risk assessments should reference existing risk keywords",
		},
		{
			ChildTable:       "business_risk_assessments",
			ChildColumn:      "business_id",
			ParentTable:      "merchants",
			ParentColumn:     "id",
			RelationshipType: "logical",
			Description:      "Business risk assessments should reference existing merchants",
		},
	}

	// Only add relationships for tables that exist
	for _, rel := range logicalRels {
		if tableExists(rel.ChildTable, tables) && tableExists(rel.ParentTable, tables) {
			relationships = append(relationships, rel)
		}
	}

	return relationships, nil
}

// getExistingTables gets a list of all existing tables
func getExistingTables(ctx context.Context, db *sql.DB) ([]string, error) {
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_type = 'BASE TABLE'
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query existing tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tables: %w", err)
	}

	return tables, nil
}

// tableExists checks if a table exists in the list of tables
func tableExists(tableName string, tables []string) bool {
	for _, table := range tables {
		if table == tableName {
			return true
		}
	}
	return false
}

// checkOrphanedRecords checks for orphaned records in a specific relationship
func checkOrphanedRecords(ctx context.Context, db *sql.DB, test OrphanedRecordTest) OrphanedRecordResult {
	startTime := time.Now()
	result := OrphanedRecordResult{
		Test: test,
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
		test.ParentColumn,
		test.ChildTable,
		test.ParentTable,
		test.ChildColumn,
		test.ParentColumn,
		test.ChildColumn,
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

	// Get sample orphaned records if any exist
	if orphanedRecords > 0 {
		sampleQuery := fmt.Sprintf(`
			SELECT child.%s
			FROM %s child
			LEFT JOIN %s parent ON child.%s = parent.%s
			WHERE child.%s IS NOT NULL
			AND parent.%s IS NULL
			LIMIT 5
		`,
			test.ChildColumn,
			test.ChildTable,
			test.ParentTable,
			test.ChildColumn,
			test.ParentColumn,
			test.ChildColumn,
			test.ParentColumn,
		)

		rows, err := db.QueryContext(ctx, sampleQuery)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var value sql.NullString
				if err := rows.Scan(&value); err == nil && value.Valid {
					result.SampleOrphaned = append(result.SampleOrphaned, value.String)
				}
			}
		}
	}

	if orphanedRecords > 0 {
		result.Status = "FAIL"
	} else {
		result.Status = "PASS"
	}

	return result
}

// minIntOrphaned returns the minimum of two integers
func minIntOrphaned(a, b int) int {
	if a < b {
		return a
	}
	return b
}
