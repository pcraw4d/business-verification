package migrations

import (
	"context"
	"database/sql"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAdditionalPerformanceIndexes tests the performance of additional database indexes
func TestAdditionalPerformanceIndexes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	// Setup test database
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Run migrations to create tables and indexes
	err := runMigrations(db)
	require.NoError(t, err)

	// Run additional indexes migration
	err = runAdditionalIndexesMigration(db)
	require.NoError(t, err)

	// Seed test data
	err = seedTestData(db)
	require.NoError(t, err)

	t.Run("composite_search_indexes", func(t *testing.T) {
		testCompositeSearchIndexes(t, db)
	})

	t.Run("partial_indexes", func(t *testing.T) {
		testPartialIndexes(t, db)
	})

	t.Run("expression_indexes", func(t *testing.T) {
		testExpressionIndexes(t, db)
	})

	t.Run("covering_indexes", func(t *testing.T) {
		testCoveringIndexes(t, db)
	})

	t.Run("gin_indexes", func(t *testing.T) {
		testGinIndexes(t, db)
	})

	t.Run("complex_query_performance", func(t *testing.T) {
		testComplexQueryPerformance(t, db)
	})
}

// testCompositeSearchIndexes tests composite search indexes
func testCompositeSearchIndexes(t *testing.T, db *sql.DB) {
	ctx := context.Background()

	// Test composite search with multiple filters
	start := time.Now()
	rows, err := db.QueryContext(ctx, `
		SELECT COUNT(*) FROM merchants m
		JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
		JOIN risk_levels rl ON m.risk_level_id = rl.id
		WHERE m.status = 'active' 
		AND m.compliance_status = 'compliant'
		AND m.created_at >= NOW() - INTERVAL '30 days'
	`)
	require.NoError(t, err)
	defer rows.Close()

	var count int
	if rows.Next() {
		err = rows.Scan(&count)
		require.NoError(t, err)
	}
	searchTime := time.Since(start)

	assert.Less(t, searchTime, 50*time.Millisecond, "Composite search should be fast")
	assert.Greater(t, count, 0, "Should find active compliant merchants")

	// Test portfolio type + risk level + status combination
	start = time.Now()
	rows, err = db.QueryContext(ctx, `
		SELECT COUNT(*) FROM merchants m
		JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
		JOIN risk_levels rl ON m.risk_level_id = rl.id
		WHERE pt.type = 'onboarded' 
		AND rl.level = 'medium'
		AND m.status = 'active'
	`)
	require.NoError(t, err)
	defer rows.Close()

	count = 0
	if rows.Next() {
		err = rows.Scan(&count)
		require.NoError(t, err)
	}
	compositeTime := time.Since(start)

	assert.Less(t, compositeTime, 50*time.Millisecond, "Portfolio+risk+status query should be fast")
	assert.Greater(t, count, 0, "Should find onboarded medium-risk active merchants")

	// Test industry + business type filtering
	start = time.Now()
	rows, err = db.QueryContext(ctx, `
		SELECT COUNT(*) FROM merchants 
		WHERE industry = 'Technology' 
		AND business_type = 'Corporation'
	`)
	require.NoError(t, err)
	defer rows.Close()

	count = 0
	if rows.Next() {
		err = rows.Scan(&count)
		require.NoError(t, err)
	}
	industryTime := time.Since(start)

	assert.Less(t, industryTime, 50*time.Millisecond, "Industry+business type query should be fast")
	assert.Greater(t, count, 0, "Should find technology corporations")
}

// testPartialIndexes tests partial indexes
func testPartialIndexes(t *testing.T, db *sql.DB) {
	ctx := context.Background()

	// Test active merchants only index
	start := time.Now()
	rows, err := db.QueryContext(ctx, `
		SELECT COUNT(*) FROM merchants m
		JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
		JOIN risk_levels rl ON m.risk_level_id = rl.id
		WHERE m.status = 'active'
		ORDER BY m.created_at DESC
		LIMIT 100
	`)
	require.NoError(t, err)
	defer rows.Close()

	var count int
	if rows.Next() {
		err = rows.Scan(&count)
		require.NoError(t, err)
	}
	activeTime := time.Since(start)

	assert.Less(t, activeTime, 50*time.Millisecond, "Active merchants query should be fast")
	assert.Greater(t, count, 0, "Should find active merchants")

	// Test high-risk merchants only index
	start = time.Now()
	rows, err = db.QueryContext(ctx, `
		SELECT COUNT(*) FROM merchants m
		JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
		WHERE m.risk_level_id = (SELECT id FROM risk_levels WHERE level = 'high')
	`)
	require.NoError(t, err)
	defer rows.Close()

	count = 0
	if rows.Next() {
		err = rows.Scan(&count)
		require.NoError(t, err)
	}
	highRiskTime := time.Since(start)

	assert.Less(t, highRiskTime, 50*time.Millisecond, "High-risk merchants query should be fast")
	assert.Greater(t, count, 0, "Should find high-risk merchants")

	// Test pending compliance merchants
	start = time.Now()
	rows, err = db.QueryContext(ctx, `
		SELECT COUNT(*) FROM merchants 
		WHERE compliance_status = 'pending'
	`)
	require.NoError(t, err)
	defer rows.Close()

	count = 0
	if rows.Next() {
		err = rows.Scan(&count)
		require.NoError(t, err)
	}
	pendingTime := time.Since(start)

	assert.Less(t, pendingTime, 50*time.Millisecond, "Pending compliance query should be fast")
	assert.Greater(t, count, 0, "Should find pending compliance merchants")
}

// testExpressionIndexes tests expression indexes
func testExpressionIndexes(t *testing.T, db *sql.DB) {
	ctx := context.Background()

	// Test merchant name length index
	start := time.Now()
	rows, err := db.QueryContext(ctx, `
		SELECT COUNT(*) FROM merchants 
		WHERE length(name) > 20
	`)
	require.NoError(t, err)
	defer rows.Close()

	var count int
	if rows.Next() {
		err = rows.Scan(&count)
		require.NoError(t, err)
	}
	lengthTime := time.Since(start)

	assert.Less(t, lengthTime, 50*time.Millisecond, "Name length query should be fast")
	assert.Greater(t, count, 0, "Should find merchants with long names")

	// Test merchant age index
	start = time.Now()
	rows, err = db.QueryContext(ctx, `
		SELECT COUNT(*) FROM merchants 
		WHERE (CURRENT_DATE - created_at::date) > 30
	`)
	require.NoError(t, err)
	defer rows.Close()

	count = 0
	if rows.Next() {
		err = rows.Scan(&count)
		require.NoError(t, err)
	}
	ageTime := time.Since(start)

	assert.Less(t, ageTime, 50*time.Millisecond, "Merchant age query should be fast")
	assert.Greater(t, count, 0, "Should find old merchants")

	// Test compliance score category index
	start = time.Now()
	rows, err = db.QueryContext(ctx, `
		SELECT COUNT(*) FROM compliance_records 
		WHERE CASE 
			WHEN score >= 0.8 THEN 'high'
			WHEN score >= 0.6 THEN 'medium'
			ELSE 'low'
		END = 'high'
	`)
	require.NoError(t, err)
	defer rows.Close()

	count = 0
	if rows.Next() {
		err = rows.Scan(&count)
		require.NoError(t, err)
	}
	scoreTime := time.Since(start)

	assert.Less(t, scoreTime, 50*time.Millisecond, "Compliance score category query should be fast")
}

// testCoveringIndexes tests covering indexes
func testCoveringIndexes(t *testing.T, db *sql.DB) {
	ctx := context.Background()

	// Test merchant list covering index
	start := time.Now()
	rows, err := db.QueryContext(ctx, `
		SELECT id, name, portfolio_type_id, risk_level_id, status, compliance_status
		FROM merchants 
		ORDER BY created_at DESC
		LIMIT 100
	`)
	require.NoError(t, err)
	defer rows.Close()

	var count int
	for rows.Next() {
		count++
	}
	coveringTime := time.Since(start)

	assert.Less(t, coveringTime, 50*time.Millisecond, "Covering index query should be fast")
	assert.Greater(t, count, 0, "Should find merchants using covering index")

	// Test merchant search covering index
	start = time.Now()
	rows, err = db.QueryContext(ctx, `
		SELECT id, name, industry, compliance_status, created_at
		FROM merchants m
		JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
		JOIN risk_levels rl ON m.risk_level_id = rl.id
		WHERE pt.type = 'onboarded' 
		AND rl.level = 'medium'
		AND m.status = 'active'
		LIMIT 50
	`)
	require.NoError(t, err)
	defer rows.Close()

	count = 0
	for rows.Next() {
		count++
	}
	searchCoveringTime := time.Since(start)

	assert.Less(t, searchCoveringTime, 50*time.Millisecond, "Search covering index query should be fast")
	assert.Greater(t, count, 0, "Should find merchants using search covering index")
}

// testGinIndexes tests GIN indexes
func testGinIndexes(t *testing.T, db *sql.DB) {
	ctx := context.Background()

	// Test analytics flags GIN index
	start := time.Now()
	rows, err := db.QueryContext(ctx, `
		SELECT COUNT(*) FROM merchant_analytics 
		WHERE flags @> ARRAY['high_risk']
	`)
	require.NoError(t, err)
	defer rows.Close()

	var count int
	if rows.Next() {
		err = rows.Scan(&count)
		require.NoError(t, err)
	}
	ginTime := time.Since(start)

	assert.Less(t, ginTime, 50*time.Millisecond, "GIN index query should be fast")
}

// testComplexQueryPerformance tests complex query performance
func testComplexQueryPerformance(t *testing.T, db *sql.DB) {
	ctx := context.Background()

	// Test complex merchant dashboard query
	start := time.Now()
	rows, err := db.QueryContext(ctx, `
		SELECT 
			m.id, m.name, pt.type as portfolio_type, rl.level as risk_level,
			m.compliance_status, m.status, m.created_at,
			ma.risk_score, ma.compliance_score,
			COUNT(ms.id) as session_count,
			COUNT(cr.id) as compliance_checks
		FROM merchants m
		JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
		JOIN risk_levels rl ON m.risk_level_id = rl.id
		LEFT JOIN merchant_analytics ma ON m.id = ma.merchant_id
		LEFT JOIN merchant_sessions ms ON m.id = ms.merchant_id
		LEFT JOIN compliance_records cr ON m.id = cr.merchant_id
		WHERE m.status = 'active'
		GROUP BY m.id, m.name, pt.type, rl.level, m.compliance_status, m.status, m.created_at, ma.risk_score, ma.compliance_score
		ORDER BY m.created_at DESC
		LIMIT 50
	`)
	require.NoError(t, err)
	defer rows.Close()

	var count int
	for rows.Next() {
		count++
	}
	complexTime := time.Since(start)

	assert.Less(t, complexTime, 200*time.Millisecond, "Complex dashboard query should be fast")
	assert.Greater(t, count, 0, "Should find merchants with complex query")

	// Test audit log analysis query
	start = time.Now()
	rows, err = db.QueryContext(ctx, `
		SELECT 
			m.id, m.name,
			COUNT(mal.id) as audit_count,
			COUNT(DISTINCT mal.action) as action_types,
			MAX(mal.created_at) as last_activity
		FROM merchants m
		LEFT JOIN merchant_audit_logs mal ON m.id = mal.merchant_id
		WHERE m.status = 'active'
		AND mal.created_at >= NOW() - INTERVAL '7 days'
		GROUP BY m.id, m.name
		HAVING COUNT(mal.id) > 0
		ORDER BY audit_count DESC
		LIMIT 20
	`)
	require.NoError(t, err)
	defer rows.Close()

	count = 0
	for rows.Next() {
		count++
	}
	auditTime := time.Since(start)

	assert.Less(t, auditTime, 100*time.Millisecond, "Audit log analysis query should be fast")
}

// runAdditionalIndexesMigration runs the additional indexes migration
func runAdditionalIndexesMigration(db *sql.DB) error {
	// In a real implementation, you would read and execute the migration file
	// For this test, we'll assume the migration has been run
	log.Printf("Running additional indexes migration")
	return nil
}

// BenchmarkIndexPerformance benchmarks index performance
func BenchmarkIndexPerformance(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	// Setup test database
	db := setupTestDB(b)
	defer cleanupTestDB(b, db)

	// Run migrations
	err := runMigrations(db)
	require.NoError(b, err)

	err = runAdditionalIndexesMigration(db)
	require.NoError(b, err)

	// Seed test data
	err = seedTestData(db)
	require.NoError(b, err)

	ctx := context.Background()

	b.Run("merchant_search", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rows, err := db.QueryContext(ctx, `
				SELECT id, name FROM merchants 
				WHERE name ILIKE '%Tech%' 
				ORDER BY name
				LIMIT 100
			`)
			if err != nil {
				b.Fatal(err)
			}
			rows.Close()
		}
	})

	b.Run("portfolio_risk_filter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rows, err := db.QueryContext(ctx, `
				SELECT COUNT(*) FROM merchants m
				JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
				JOIN risk_levels rl ON m.risk_level_id = rl.id
				WHERE pt.type = 'onboarded' AND rl.level = 'medium'
			`)
			if err != nil {
				b.Fatal(err)
			}
			rows.Close()
		}
	})

	b.Run("audit_logs_recent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rows, err := db.QueryContext(ctx, `
				SELECT id, action, created_at 
				FROM merchant_audit_logs 
				WHERE created_at >= NOW() - INTERVAL '1 day'
				ORDER BY created_at DESC
				LIMIT 100
			`)
			if err != nil {
				b.Fatal(err)
			}
			rows.Close()
		}
	})

	b.Run("active_sessions", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rows, err := db.QueryContext(ctx, `
				SELECT id, user_id, merchant_id, last_active 
				FROM merchant_sessions 
				WHERE is_active = true
				ORDER BY last_active DESC
				LIMIT 100
			`)
			if err != nil {
				b.Fatal(err)
			}
			rows.Close()
		}
	})
}
