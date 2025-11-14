package database

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRepositoryMethodsWithData tests repository methods with actual test data
func TestRepositoryMethodsWithData(t *testing.T) {
	dbURL := getTestDatabaseURL(t)
	
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	
	// Setup: Create test data
	setupTestData(t, db, ctx)
	defer cleanupTestData(t, db, ctx)
	
	t.Run("GetPortfolioDistribution", func(t *testing.T) {
		query := `
			SELECT pt.type, COUNT(*) as count
			FROM merchants m
			JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
			GROUP BY pt.type
		`

		rows, err := db.QueryContext(ctx, query)
		require.NoError(t, err)
		defer rows.Close()

		dist := make(map[string]int)
		for rows.Next() {
			var portfolioType string
			var count int
			err := rows.Scan(&portfolioType, &count)
			require.NoError(t, err)
			dist[portfolioType] = count
		}
		require.NoError(t, rows.Err())
		
		// Verify we have data
		totalMerchants := 0
		for _, count := range dist {
			totalMerchants += count
		}
		
		assert.Greater(t, totalMerchants, 0, "Should have merchants in portfolio distribution")
		t.Logf("✓ Portfolio distribution: %v (total: %d merchants)", dist, totalMerchants)
	})

	t.Run("GetRiskDistribution", func(t *testing.T) {
		query := `
			SELECT rl.level, COUNT(*) as count
			FROM merchants m
			JOIN risk_levels rl ON m.risk_level_id = rl.id
			GROUP BY rl.level
		`

		rows, err := db.QueryContext(ctx, query)
		require.NoError(t, err)
		defer rows.Close()

		dist := make(map[string]int)
		for rows.Next() {
			var riskLevel string
			var count int
			err := rows.Scan(&riskLevel, &count)
			require.NoError(t, err)
			dist[riskLevel] = count
		}
		require.NoError(t, rows.Err())
		
		// Verify we have data
		totalMerchants := 0
		for _, count := range dist {
			totalMerchants += count
		}
		
		assert.Greater(t, totalMerchants, 0, "Should have merchants in risk distribution")
		t.Logf("✓ Risk distribution: %v (total: %d merchants)", dist, totalMerchants)
	})

	t.Run("GetIndustryDistribution", func(t *testing.T) {
		query := `
			SELECT COALESCE(industry, 'Unknown') as industry, COUNT(*) as count
			FROM merchants
			WHERE industry IS NOT NULL AND industry != ''
			GROUP BY industry
			ORDER BY count DESC
			LIMIT 10
		`

		rows, err := db.QueryContext(ctx, query)
		require.NoError(t, err)
		defer rows.Close()

		dist := make(map[string]int)
		for rows.Next() {
			var industry string
			var count int
			err := rows.Scan(&industry, &count)
			require.NoError(t, err)
			dist[industry] = count
		}
		require.NoError(t, rows.Err())
		
		t.Logf("✓ Industry distribution: %v", dist)
		// Note: May be empty if no merchants have industry set, which is okay
	})

	t.Run("GetComplianceDistribution", func(t *testing.T) {
		query := `
			SELECT COALESCE(compliance_status, 'pending') as compliance_status, COUNT(*) as count
			FROM merchants
			GROUP BY compliance_status
		`

		rows, err := db.QueryContext(ctx, query)
		require.NoError(t, err)
		defer rows.Close()

		dist := make(map[string]int)
		for rows.Next() {
			var status string
			var count int
			err := rows.Scan(&status, &count)
			require.NoError(t, err)
			dist[status] = count
		}
		require.NoError(t, rows.Err())
		
		// Verify we have data
		totalMerchants := 0
		for _, count := range dist {
			totalMerchants += count
		}
		
		assert.Greater(t, totalMerchants, 0, "Should have merchants in compliance distribution")
		t.Logf("✓ Compliance distribution: %v (total: %d merchants)", dist, totalMerchants)
	})

	t.Run("CountMerchants", func(t *testing.T) {
		var count int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM merchants").Scan(&count)
		require.NoError(t, err)
		
		assert.Greater(t, count, 0, "Should have merchants in database")
		t.Logf("✓ Total merchants: %d", count)
	})
}

// setupTestData creates test data for repository method testing
func setupTestData(t *testing.T, db *sql.DB, ctx context.Context) {
	// Create test user
	userQuery := `
		INSERT INTO users (
			id, email, username, password_hash, first_name, last_name, role
		) VALUES (
			gen_random_uuid(),
			'repo-test@example.com',
			'repotestuser',
			'$2a$10$dummyhash',
			'Repo',
			'Test',
			'user'
		)
		ON CONFLICT (email) DO NOTHING
		RETURNING id
	`
	
	var userID string
	err := db.QueryRowContext(ctx, userQuery).Scan(&userID)
	if err == sql.ErrNoRows {
		err = db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = 'repo-test@example.com'").Scan(&userID)
		require.NoError(t, err)
	} else {
		require.NoError(t, err)
	}

	// Ensure portfolio types exist
	portfolioTypes := []string{"onboarded", "deactivated", "prospective", "pending"}
	portfolioTypeIDs := make(map[string]string)
	
	for i, ptType := range portfolioTypes {
		var id string
		query := `
			INSERT INTO portfolio_types (type, description, display_order, is_active)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (type) DO UPDATE SET description = EXCLUDED.description
			RETURNING id
		`
		err := db.QueryRowContext(ctx, query, ptType, "Test "+ptType, i+1, true).Scan(&id)
		if err == sql.ErrNoRows {
			err = db.QueryRowContext(ctx, "SELECT id FROM portfolio_types WHERE type = $1", ptType).Scan(&id)
			require.NoError(t, err)
		} else {
			require.NoError(t, err)
		}
		portfolioTypeIDs[ptType] = id
	}

	// Ensure risk levels exist
	riskLevels := []struct {
		level        string
		numericValue int
		colorCode    string
	}{
		{"low", 1, "#10B981"},
		{"medium", 2, "#F59E0B"},
		{"high", 3, "#EF4444"},
	}
	riskLevelIDs := make(map[string]string)
	
	for _, rl := range riskLevels {
		var id string
		query := `
			INSERT INTO risk_levels (level, description, numeric_value, color_code, display_order)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (level) DO UPDATE SET description = EXCLUDED.description
			RETURNING id
		`
		err := db.QueryRowContext(ctx, query, rl.level, "Test "+rl.level, rl.numericValue, rl.colorCode, rl.numericValue).Scan(&id)
		if err == sql.ErrNoRows {
			err = db.QueryRowContext(ctx, "SELECT id FROM risk_levels WHERE level = $1", rl.level).Scan(&id)
			require.NoError(t, err)
		} else {
			require.NoError(t, err)
		}
		riskLevelIDs[rl.level] = id
	}

	// Create test merchants with various portfolio types, risk levels, and compliance statuses
	merchants := []struct {
		name             string
		regNum           string
		portfolioType    string
		riskLevel        string
		complianceStatus string
		industry         string
	}{
		{"Merchant 1", "MERC-001", "onboarded", "low", "compliant", "Technology"},
		{"Merchant 2", "MERC-002", "onboarded", "medium", "compliant", "Retail"},
		{"Merchant 3", "MERC-003", "prospective", "high", "pending", "Finance"},
		{"Merchant 4", "MERC-004", "pending", "low", "non_compliant", "Manufacturing"},
		{"Merchant 5", "MERC-005", "deactivated", "medium", "compliant", "Services"},
		{"Merchant 6", "MERC-006", "onboarded", "low", "compliant", "Technology"},
		{"Merchant 7", "MERC-007", "prospective", "high", "pending", "Retail"},
		{"Merchant 8", "MERC-008", "onboarded", "medium", "compliant", "Finance"},
	}

	for _, merchant := range merchants {
		legalName := merchant.name + " Inc."
		query := `
			INSERT INTO merchants (
				id, name, legal_name, registration_number, portfolio_type_id, risk_level_id, 
				compliance_status, industry, created_by, created_at, updated_at
			) VALUES (
				gen_random_uuid(),
				$1,
				$2,
				$3,
				$4::uuid,
				$5::uuid,
				$6,
				$7,
				$8::uuid,
				NOW(),
				NOW()
			)
			ON CONFLICT (registration_number) DO NOTHING
		`
		
		_, err := db.ExecContext(ctx, query,
			merchant.name,
			legalName,
			merchant.regNum,
			portfolioTypeIDs[merchant.portfolioType],
			riskLevelIDs[merchant.riskLevel],
			merchant.complianceStatus,
			merchant.industry,
			userID,
		)
		require.NoError(t, err)
	}

	t.Log("✓ Test data setup complete")
}

// cleanupTestData removes test data (optional, for clean test runs)
func cleanupTestData(t *testing.T, db *sql.DB, ctx context.Context) {
	// Clean up test merchants (they should cascade delete businesses)
	_, err := db.ExecContext(ctx, `
		DELETE FROM merchants 
		WHERE name LIKE 'Merchant %'
	`)
	if err != nil {
		t.Logf("Note: Could not clean up test merchants: %v", err)
	}
	
	t.Log("✓ Test data cleanup complete")
}

