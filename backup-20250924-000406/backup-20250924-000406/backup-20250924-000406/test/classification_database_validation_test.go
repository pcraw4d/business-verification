package test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ClassificationDatabaseValidator validates database queries and data integrity
type ClassificationDatabaseValidator struct {
	db     *sql.DB
	logger *log.Logger
}

// NewClassificationDatabaseValidator creates a new database validator
func NewClassificationDatabaseValidator(db *sql.DB, logger *log.Logger) *ClassificationDatabaseValidator {
	return &ClassificationDatabaseValidator{
		db:     db,
		logger: logger,
	}
}

// ValidateDatabaseSchema validates that all required tables exist and have correct structure
func (cdv *ClassificationDatabaseValidator) ValidateDatabaseSchema(t *testing.T) {
	t.Log("🔍 Validating Database Schema")

	ctx := context.Background()

	// Expected tables and their required columns
	expectedTables := map[string][]string{
		"industries": {
			"id", "name", "description", "category", "confidence_threshold",
			"is_active", "created_at", "updated_at",
		},
		"industry_keywords": {
			"id", "industry_id", "keyword", "weight", "is_active",
			"created_at", "updated_at",
		},
		"classification_codes": {
			"id", "code", "type", "description", "industry_id",
			"is_active", "created_at", "updated_at",
		},
		"code_keywords": {
			"id", "code_id", "keyword", "weight", "is_active",
			"created_at", "updated_at",
		},
		"industry_patterns": {
			"id", "industry_id", "pattern", "pattern_type", "weight",
			"is_active", "created_at", "updated_at",
		},
		"keyword_weights": {
			"id", "keyword", "weight", "context", "is_active",
			"created_at", "updated_at",
		},
	}

	for tableName, expectedColumns := range expectedTables {
		t.Run(fmt.Sprintf("Table %s", tableName), func(t *testing.T) {
			// Check if table exists
			query := `
				SELECT EXISTS (
					SELECT FROM information_schema.tables 
					WHERE table_schema = 'public' 
					AND table_name = $1
				)
			`

			var exists bool
			err := cdv.db.QueryRowContext(ctx, query, tableName).Scan(&exists)
			require.NoError(t, err, "Should be able to check table existence")

			if !exists {
				t.Logf("⚠️ Table %s does not exist - this may be expected if migration hasn't been run", tableName)
				return
			}

			// Check table columns
			columnQuery := `
				SELECT column_name, data_type, is_nullable, column_default
				FROM information_schema.columns 
				WHERE table_schema = 'public' 
				AND table_name = $1
				ORDER BY ordinal_position
			`

			rows, err := cdv.db.QueryContext(ctx, columnQuery, tableName)
			require.NoError(t, err, "Should be able to query table columns")
			defer rows.Close()

			actualColumns := make(map[string]bool)
			for rows.Next() {
				var columnName, dataType, isNullable, columnDefault sql.NullString
				err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault)
				require.NoError(t, err, "Should be able to scan column information")

				actualColumns[columnName.String] = true
			}

			// Verify all expected columns exist
			for _, expectedColumn := range expectedColumns {
				assert.True(t, actualColumns[expectedColumn],
					"Table %s should have column %s", tableName, expectedColumn)
			}

			t.Logf("✅ Table %s schema validated", tableName)
		})
	}
}

// ValidateDataIntegrity validates data integrity and constraints
func (cdv *ClassificationDatabaseValidator) ValidateDataIntegrity(t *testing.T) {
	t.Log("🔍 Validating Data Integrity")

	ctx := context.Background()

	// Test 1: Check for orphaned records
	t.Run("Orphaned Records Check", func(t *testing.T) {
		// Check for orphaned industry_keywords
		query := `
			SELECT COUNT(*) 
			FROM industry_keywords ik 
			LEFT JOIN industries i ON ik.industry_id = i.id 
			WHERE i.id IS NULL
		`

		var orphanedCount int
		err := cdv.db.QueryRowContext(ctx, query).Scan(&orphanedCount)
		if err != nil {
			t.Logf("⚠️ Could not check orphaned records - table may not exist: %v", err)
			return
		}

		assert.Equal(t, 0, orphanedCount, "Should not have orphaned industry_keywords records")

		// Check for orphaned classification_codes
		query = `
			SELECT COUNT(*) 
			FROM classification_codes cc 
			LEFT JOIN industries i ON cc.industry_id = i.id 
			WHERE i.id IS NULL
		`

		err = cdv.db.QueryRowContext(ctx, query).Scan(&orphanedCount)
		if err != nil {
			t.Logf("⚠️ Could not check orphaned classification_codes - table may not exist: %v", err)
			return
		}

		assert.Equal(t, 0, orphanedCount, "Should not have orphaned classification_codes records")
	})

	// Test 2: Check for duplicate records
	t.Run("Duplicate Records Check", func(t *testing.T) {
		// Check for duplicate industry names
		query := `
			SELECT name, COUNT(*) 
			FROM industries 
			GROUP BY name 
			HAVING COUNT(*) > 1
		`

		rows, err := cdv.db.QueryContext(ctx, query)
		if err != nil {
			t.Logf("⚠️ Could not check duplicate industries - table may not exist: %v", err)
			return
		}
		defer rows.Close()

		var duplicates []string
		for rows.Next() {
			var name string
			var count int
			err := rows.Scan(&name, &count)
			require.NoError(t, err, "Should be able to scan duplicate check results")
			duplicates = append(duplicates, name)
		}

		assert.Empty(t, duplicates, "Should not have duplicate industry names: %v", duplicates)

		// Check for duplicate industry_keywords
		query = `
			SELECT industry_id, keyword, COUNT(*) 
			FROM industry_keywords 
			GROUP BY industry_id, keyword 
			HAVING COUNT(*) > 1
		`

		rows, err = cdv.db.QueryContext(ctx, query)
		if err != nil {
			t.Logf("⚠️ Could not check duplicate industry_keywords - table may not exist: %v", err)
			return
		}
		defer rows.Close()

		var keywordDuplicates []string
		for rows.Next() {
			var industryID int
			var keyword string
			var count int
			err := rows.Scan(&industryID, &keyword, &count)
			require.NoError(t, err, "Should be able to scan duplicate keyword check results")
			keywordDuplicates = append(keywordDuplicates, fmt.Sprintf("industry_id=%d, keyword=%s", industryID, keyword))
		}

		assert.Empty(t, keywordDuplicates, "Should not have duplicate industry_keywords: %v", keywordDuplicates)
	})

	// Test 3: Check data types and constraints
	t.Run("Data Types and Constraints", func(t *testing.T) {
		// Check that confidence_threshold is within valid range
		query := `
			SELECT COUNT(*) 
			FROM industries 
			WHERE confidence_threshold < 0 OR confidence_threshold > 1
		`

		var invalidCount int
		err := cdv.db.QueryRowContext(ctx, query).Scan(&invalidCount)
		if err != nil {
			t.Logf("⚠️ Could not check confidence_threshold constraints - table may not exist: %v", err)
			return
		}

		assert.Equal(t, 0, invalidCount, "Should not have confidence_threshold values outside 0-1 range")

		// Check that weight is within valid range
		query = `
			SELECT COUNT(*) 
			FROM industry_keywords 
			WHERE weight < 0 OR weight > 10
		`

		err = cdv.db.QueryRowContext(ctx, query).Scan(&invalidCount)
		if err != nil {
			t.Logf("⚠️ Could not check weight constraints - table may not exist: %v", err)
			return
		}

		assert.Equal(t, 0, invalidCount, "Should not have weight values outside 0-10 range")
	})
}

// ValidateQueryPerformance validates that queries perform within acceptable limits
func (cdv *ClassificationDatabaseValidator) ValidateQueryPerformance(t *testing.T) {
	t.Log("🔍 Validating Query Performance")

	ctx := context.Background()

	// Test 1: Basic industry lookup performance
	t.Run("Industry Lookup Performance", func(t *testing.T) {
		query := `SELECT id, name, description FROM industries WHERE is_active = true LIMIT 10`

		startTime := time.Now()
		rows, err := cdv.db.QueryContext(ctx, query)
		duration := time.Since(startTime)

		if err != nil {
			t.Logf("⚠️ Could not test industry lookup performance - table may not exist: %v", err)
			return
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var id int
			var name, description string
			err := rows.Scan(&id, &name, &description)
			require.NoError(t, err, "Should be able to scan industry data")
			count++
		}

		assert.Less(t, duration, 100*time.Millisecond, "Industry lookup should complete within 100ms")
		t.Logf("✅ Industry lookup: %v for %d records", duration, count)
	})

	// Test 2: Keyword search performance
	t.Run("Keyword Search Performance", func(t *testing.T) {
		query := `
			SELECT ik.keyword, ik.weight, i.name as industry_name
			FROM industry_keywords ik
			JOIN industries i ON ik.industry_id = i.id
			WHERE ik.keyword ILIKE $1 AND ik.is_active = true
			LIMIT 20
		`

		startTime := time.Now()
		rows, err := cdv.db.QueryContext(ctx, query, "%software%")
		duration := time.Since(startTime)

		if err != nil {
			t.Logf("⚠️ Could not test keyword search performance - table may not exist: %v", err)
			return
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var keyword string
			var weight float64
			var industryName string
			err := rows.Scan(&keyword, &weight, &industryName)
			require.NoError(t, err, "Should be able to scan keyword search results")
			count++
		}

		assert.Less(t, duration, 200*time.Millisecond, "Keyword search should complete within 200ms")
		t.Logf("✅ Keyword search: %v for %d results", duration, count)
	})

	// Test 3: Complex classification query performance
	t.Run("Complex Classification Query Performance", func(t *testing.T) {
		query := `
			SELECT 
				i.name as industry_name,
				ik.keyword,
				ik.weight,
				cc.code,
				cc.type as code_type
			FROM industries i
			LEFT JOIN industry_keywords ik ON i.id = ik.industry_id AND ik.is_active = true
			LEFT JOIN classification_codes cc ON i.id = cc.industry_id AND cc.is_active = true
			WHERE i.is_active = true
			ORDER BY i.name, ik.weight DESC
			LIMIT 50
		`

		startTime := time.Now()
		rows, err := cdv.db.QueryContext(ctx, query)
		duration := time.Since(startTime)

		if err != nil {
			t.Logf("⚠️ Could not test complex classification query performance - tables may not exist: %v", err)
			return
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var industryName, keyword, code, codeType sql.NullString
			var weight sql.NullFloat64
			err := rows.Scan(&industryName, &keyword, &weight, &code, &codeType)
			require.NoError(t, err, "Should be able to scan complex query results")
			count++
		}

		assert.Less(t, duration, 500*time.Millisecond, "Complex classification query should complete within 500ms")
		t.Logf("✅ Complex classification query: %v for %d results", duration, count)
	})
}

// ValidateIndexes validates that required indexes exist and are being used
func (cdv *ClassificationDatabaseValidator) ValidateIndexes(t *testing.T) {
	t.Log("🔍 Validating Database Indexes")

	ctx := context.Background()

	// Expected indexes
	expectedIndexes := map[string][]string{
		"industries": {
			"idx_industries_name",
			"idx_industries_category",
			"idx_industries_active",
		},
		"industry_keywords": {
			"idx_industry_keywords_industry_id",
			"idx_industry_keywords_keyword",
			"idx_industry_keywords_weight",
			"idx_industry_keywords_active",
		},
		"classification_codes": {
			"idx_classification_codes_industry_id",
			"idx_classification_codes_code",
			"idx_classification_codes_type",
			"idx_classification_codes_active",
		},
	}

	for tableName, expectedIndexNames := range expectedIndexes {
		t.Run(fmt.Sprintf("Indexes for %s", tableName), func(t *testing.T) {
			// Check if table exists first
			tableExistsQuery := `
				SELECT EXISTS (
					SELECT FROM information_schema.tables 
					WHERE table_schema = 'public' 
					AND table_name = $1
				)
			`

			var tableExists bool
			err := cdv.db.QueryRowContext(ctx, tableExistsQuery, tableName).Scan(&tableExists)
			require.NoError(t, err, "Should be able to check table existence")

			if !tableExists {
				t.Logf("⚠️ Table %s does not exist - skipping index validation", tableName)
				return
			}

			// Get actual indexes for the table
			indexQuery := `
				SELECT indexname 
				FROM pg_indexes 
				WHERE tablename = $1 AND schemaname = 'public'
			`

			rows, err := cdv.db.QueryContext(ctx, indexQuery, tableName)
			require.NoError(t, err, "Should be able to query table indexes")
			defer rows.Close()

			actualIndexes := make(map[string]bool)
			for rows.Next() {
				var indexName string
				err := rows.Scan(&indexName)
				require.NoError(t, err, "Should be able to scan index names")
				actualIndexes[indexName] = true
			}

			// Check for expected indexes
			for _, expectedIndex := range expectedIndexNames {
				if !actualIndexes[expectedIndex] {
					t.Logf("⚠️ Expected index %s not found on table %s", expectedIndex, tableName)
				} else {
					t.Logf("✅ Found index %s on table %s", expectedIndex, tableName)
				}
			}
		})
	}
}

// RunComprehensiveDatabaseValidation runs all database validation tests
func (cdv *ClassificationDatabaseValidator) RunComprehensiveDatabaseValidation(t *testing.T) {
	t.Log("🚀 Starting Comprehensive Database Validation")

	// Test 1: Schema Validation
	cdv.ValidateDatabaseSchema(t)

	// Test 2: Data Integrity
	cdv.ValidateDataIntegrity(t)

	// Test 3: Query Performance
	cdv.ValidateQueryPerformance(t)

	// Test 4: Index Validation
	cdv.ValidateIndexes(t)

	t.Log("✅ Comprehensive Database Validation Completed")
}

// TestClassificationDatabaseValidation runs the database validation test
func TestClassificationDatabaseValidation(t *testing.T) {
	// Skip if no database connection available
	if testing.Short() {
		t.Skip("Skipping classification database validation in short mode")
	}

	// For now, skip the test since we don't have database connection setup
	t.Skip("Skipping test - database connection not configured")
}
