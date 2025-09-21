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

// ConsistencyTest represents a data consistency verification test
type ConsistencyTest struct {
	TestName       string
	Description    string
	TestType       string // "count", "sum", "avg", "min", "max", "custom"
	Query          string
	ExpectedResult interface{}
	Tolerance      float64 // For numeric comparisons
	Critical       bool    // Whether this is a critical consistency check
}

// ConsistencyResult represents the result of a consistency test
type ConsistencyResult struct {
	Test           ConsistencyTest
	Status         string // "PASS", "FAIL", "ERROR"
	ErrorMessage   string
	ActualResult   interface{}
	ExpectedResult interface{}
	Difference     float64
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

	fmt.Println("🔍 Starting Data Consistency Verification...")
	fmt.Println(strings.Repeat("=", 60))

	// Get all consistency tests to run
	tests, err := getConsistencyTests(ctx, db)
	if err != nil {
		log.Fatalf("Failed to get consistency tests: %v", err)
	}

	fmt.Printf("Found %d consistency tests to run\n\n", len(tests))

	// Run consistency tests
	var results []ConsistencyResult
	var passed, failed, errors int

	for i, test := range tests {
		fmt.Printf("[%d/%d] %s (%s)\n",
			i+1, len(tests), test.TestName, test.TestType)

		result := runConsistencyTest(ctx, db, test)
		results = append(results, result)

		switch result.Status {
		case "PASS":
			passed++
			fmt.Printf("  ✅ PASS - %s\n", test.Description)
		case "FAIL":
			failed++
			fmt.Printf("  ❌ FAIL - %s\n", test.Description)
			if result.Difference != 0 {
				fmt.Printf("  📊 Expected: %v, Actual: %v, Difference: %.2f\n",
					result.ExpectedResult, result.ActualResult, result.Difference)
			}
		case "ERROR":
			errors++
			fmt.Printf("  🚨 ERROR - %s\n", result.ErrorMessage)
		}

		if test.Critical && result.Status != "PASS" {
			fmt.Printf("  ⚠️  CRITICAL TEST FAILED\n")
		}

		fmt.Printf("  ⏱️  Execution time: %v\n\n", result.ExecutionTime)
	}

	// Print summary
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("📊 DATA CONSISTENCY VERIFICATION SUMMARY")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total Tests: %d\n", len(results))
	fmt.Printf("✅ Passed: %d\n", passed)
	fmt.Printf("❌ Failed: %d\n", failed)
	fmt.Printf("🚨 Errors: %d\n", errors)
	fmt.Printf("Success Rate: %.1f%%\n", float64(passed)/float64(len(results))*100)

	// Check for critical failures
	criticalFailures := 0
	for _, result := range results {
		if result.Test.Critical && result.Status != "PASS" {
			criticalFailures++
		}
	}

	if criticalFailures > 0 {
		fmt.Printf("⚠️  Critical Failures: %d\n", criticalFailures)
	}

	// Print detailed results for failures
	if failed > 0 || errors > 0 {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("🔍 DETAILED FAILURE REPORT")
		fmt.Println(strings.Repeat("=", 60))

		for _, result := range results {
			if result.Status != "PASS" {
				fmt.Printf("\n❌ %s (%s)\n", result.Test.TestName, result.Test.TestType)
				fmt.Printf("   Description: %s\n", result.Test.Description)

				if result.Status == "FAIL" {
					fmt.Printf("   Expected: %v\n", result.ExpectedResult)
					fmt.Printf("   Actual: %v\n", result.ActualResult)
					if result.Difference != 0 {
						fmt.Printf("   Difference: %.2f\n", result.Difference)
					}
				} else {
					fmt.Printf("   Error: %s\n", result.ErrorMessage)
				}

				if result.Test.Critical {
					fmt.Printf("   ⚠️  CRITICAL TEST\n")
				}
			}
		}
	}

	// Exit with appropriate code
	if failed > 0 || errors > 0 {
		os.Exit(1)
	}

	fmt.Println("\n🎉 All data consistency checks passed!")
}

// getConsistencyTests retrieves all consistency tests to run
func getConsistencyTests(ctx context.Context, db *sql.DB) ([]ConsistencyTest, error) {
	var tests []ConsistencyTest

	// Get table existence tests
	tableTests, err := getTableExistenceTests(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to get table existence tests: %w", err)
	}
	tests = append(tests, tableTests...)

	// Get count consistency tests
	countTests, err := getCountConsistencyTests(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to get count consistency tests: %w", err)
	}
	tests = append(tests, countTests...)

	// Get business logic consistency tests
	businessTests, err := getBusinessLogicConsistencyTests(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to get business logic consistency tests: %w", err)
	}
	tests = append(tests, businessTests...)

	// Get data integrity consistency tests
	integrityTests, err := getDataIntegrityConsistencyTests(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to get data integrity consistency tests: %w", err)
	}
	tests = append(tests, integrityTests...)

	return tests, nil
}

// getTableExistenceTests creates tests to verify table existence and basic structure
func getTableExistenceTests(ctx context.Context, db *sql.DB) ([]ConsistencyTest, error) {
	var tests []ConsistencyTest

	// Check if core tables exist
	coreTables := []string{
		"users", "merchants", "business_verifications",
		"classification_results", "risk_assessments", "audit_logs",
		"industries", "industry_keywords", "risk_keywords",
		"business_risk_assessments", "merchant_audit_logs",
	}

	for _, table := range coreTables {
		tests = append(tests, ConsistencyTest{
			TestName:       fmt.Sprintf("Table Exists: %s", table),
			Description:    fmt.Sprintf("Verify that table %s exists", table),
			TestType:       "count",
			Query:          fmt.Sprintf("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = '%s' AND table_schema = 'public'", table),
			ExpectedResult: 1,
			Critical:       true,
		})
	}

	return tests, nil
}

// getCountConsistencyTests creates tests to verify count consistency across related tables
func getCountConsistencyTests(ctx context.Context, db *sql.DB) ([]ConsistencyTest, error) {
	var tests []ConsistencyTest

	// Test: Users should have at least one merchant (if merchants exist)
	tests = append(tests, ConsistencyTest{
		TestName:    "User-Merchant Count Consistency",
		Description: "Users with merchants should have at least one merchant record",
		TestType:    "custom",
		Query: `
			SELECT COUNT(*) 
			FROM users u 
			WHERE NOT EXISTS (
				SELECT 1 FROM merchants m WHERE m.user_id = u.id
			)
			AND EXISTS (SELECT 1 FROM merchants LIMIT 1)
		`,
		ExpectedResult: 0,
		Critical:       false,
	})

	// Test: Merchants should have at least one verification (if verifications exist)
	tests = append(tests, ConsistencyTest{
		TestName:    "Merchant-Verification Count Consistency",
		Description: "Merchants should have at least one verification record",
		TestType:    "custom",
		Query: `
			SELECT COUNT(*) 
			FROM merchants m 
			WHERE NOT EXISTS (
				SELECT 1 FROM business_verifications bv WHERE bv.merchant_id = m.id
			)
			AND EXISTS (SELECT 1 FROM business_verifications LIMIT 1)
		`,
		ExpectedResult: 0,
		Critical:       false,
	})

	// Test: Merchants should have classification results (if classifications exist)
	tests = append(tests, ConsistencyTest{
		TestName:    "Merchant-Classification Count Consistency",
		Description: "Merchants should have classification results",
		TestType:    "custom",
		Query: `
			SELECT COUNT(*) 
			FROM merchants m 
			WHERE NOT EXISTS (
				SELECT 1 FROM classification_results cr WHERE cr.merchant_id = m.id
			)
			AND EXISTS (SELECT 1 FROM classification_results LIMIT 1)
		`,
		ExpectedResult: 0,
		Critical:       false,
	})

	return tests, nil
}

// getBusinessLogicConsistencyTests creates tests to verify business logic consistency
func getBusinessLogicConsistencyTests(ctx context.Context, db *sql.DB) ([]ConsistencyTest, error) {
	var tests []ConsistencyTest

	// Test: Business verifications should have valid status values
	tests = append(tests, ConsistencyTest{
		TestName:    "Business Verification Status Consistency",
		Description: "Business verifications should have valid status values",
		TestType:    "custom",
		Query: `
			SELECT COUNT(*) 
			FROM business_verifications 
			WHERE status NOT IN ('pending', 'approved', 'rejected', 'in_progress', 'completed')
		`,
		ExpectedResult: 0,
		Critical:       true,
	})

	// Test: Classification results should have valid confidence scores
	tests = append(tests, ConsistencyTest{
		TestName:    "Classification Confidence Score Consistency",
		Description: "Classification results should have confidence scores between 0 and 1",
		TestType:    "custom",
		Query: `
			SELECT COUNT(*) 
			FROM classification_results 
			WHERE confidence_score < 0 OR confidence_score > 1
		`,
		ExpectedResult: 0,
		Critical:       true,
	})

	// Test: Risk assessments should have valid risk levels
	tests = append(tests, ConsistencyTest{
		TestName:    "Risk Assessment Level Consistency",
		Description: "Risk assessments should have valid risk levels",
		TestType:    "custom",
		Query: `
			SELECT COUNT(*) 
			FROM risk_assessments 
			WHERE risk_level NOT IN ('low', 'medium', 'high', 'critical')
		`,
		ExpectedResult: 0,
		Critical:       true,
	})

	// Test: Users should have valid email formats
	tests = append(tests, ConsistencyTest{
		TestName:    "User Email Format Consistency",
		Description: "Users should have valid email formats",
		TestType:    "custom",
		Query: `
			SELECT COUNT(*) 
			FROM users 
			WHERE email IS NOT NULL 
			AND email !~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
		`,
		ExpectedResult: 0,
		Critical:       true,
	})

	return tests, nil
}

// getDataIntegrityConsistencyTests creates tests to verify data integrity consistency
func getDataIntegrityConsistencyTests(ctx context.Context, db *sql.DB) ([]ConsistencyTest, error) {
	var tests []ConsistencyTest

	// Test: Created dates should be before updated dates
	tests = append(tests, ConsistencyTest{
		TestName:    "Date Consistency - Created Before Updated",
		Description: "Created dates should be before or equal to updated dates",
		TestType:    "custom",
		Query: `
			SELECT COUNT(*) 
			FROM merchants 
			WHERE created_at > updated_at
		`,
		ExpectedResult: 0,
		Critical:       true,
	})

	// Test: Business verifications should have created dates
	tests = append(tests, ConsistencyTest{
		TestName:    "Business Verification Date Consistency",
		Description: "Business verifications should have valid created dates",
		TestType:    "custom",
		Query: `
			SELECT COUNT(*) 
			FROM business_verifications 
			WHERE created_at IS NULL
		`,
		ExpectedResult: 0,
		Critical:       true,
	})

	// Test: Classification results should have valid timestamps
	tests = append(tests, ConsistencyTest{
		TestName:    "Classification Timestamp Consistency",
		Description: "Classification results should have valid timestamps",
		TestType:    "custom",
		Query: `
			SELECT COUNT(*) 
			FROM classification_results 
			WHERE created_at IS NULL OR updated_at IS NULL
		`,
		ExpectedResult: 0,
		Critical:       true,
	})

	// Test: Risk assessments should have assessment dates
	tests = append(tests, ConsistencyTest{
		TestName:    "Risk Assessment Date Consistency",
		Description: "Risk assessments should have valid assessment dates",
		TestType:    "custom",
		Query: `
			SELECT COUNT(*) 
			FROM risk_assessments 
			WHERE assessment_date IS NULL
		`,
		ExpectedResult: 0,
		Critical:       true,
	})

	return tests, nil
}

// runConsistencyTest runs a specific consistency test
func runConsistencyTest(ctx context.Context, db *sql.DB, test ConsistencyTest) ConsistencyResult {
	startTime := time.Now()
	result := ConsistencyResult{
		Test:           test,
		ExpectedResult: test.ExpectedResult,
	}

	// Execute the test query
	var actualValue interface{}
	err := db.QueryRowContext(ctx, test.Query).Scan(&actualValue)
	if err != nil {
		result.Status = "ERROR"
		result.ErrorMessage = err.Error()
		result.ExecutionTime = time.Since(startTime)
		return result
	}

	result.ActualResult = actualValue
	result.ExecutionTime = time.Since(startTime)

	// Compare results based on test type
	switch test.TestType {
	case "count", "custom":
		expectedInt, ok1 := test.ExpectedResult.(int)
		actualInt, ok2 := actualValue.(int64)
		if ok1 && ok2 {
			if int64(expectedInt) == actualInt {
				result.Status = "PASS"
			} else {
				result.Status = "FAIL"
				result.Difference = float64(actualInt - int64(expectedInt))
			}
		} else {
			result.Status = "ERROR"
			result.ErrorMessage = "Type mismatch in count comparison"
		}

	case "sum", "avg", "min", "max":
		expectedFloat, ok1 := test.ExpectedResult.(float64)
		actualFloat, ok2 := actualValue.(float64)
		if ok1 && ok2 {
			diff := actualFloat - expectedFloat
			if diff <= test.Tolerance && diff >= -test.Tolerance {
				result.Status = "PASS"
			} else {
				result.Status = "FAIL"
				result.Difference = diff
			}
		} else {
			result.Status = "ERROR"
			result.ErrorMessage = "Type mismatch in numeric comparison"
		}

	default:
		result.Status = "ERROR"
		result.ErrorMessage = "Unknown test type"
	}

	return result
}
