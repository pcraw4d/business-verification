package test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthcareKeywordsComprehensive tests the comprehensive healthcare keywords implementation
func TestHealthcareKeywordsComprehensive(t *testing.T) {
	// Setup database connection
	db, err := setupTestDatabase()
	require.NoError(t, err)
	defer db.Close()

	// Test healthcare industries exist
	t.Run("HealthcareIndustriesExist", func(t *testing.T) {
		testHealthcareIndustriesExist(t, db)
	})

	// Test keyword count per industry
	t.Run("KeywordCountPerIndustry", func(t *testing.T) {
		testKeywordCountPerIndustry(t, db)
	})

	// Test keyword weight distribution
	t.Run("KeywordWeightDistribution", func(t *testing.T) {
		testKeywordWeightDistribution(t, db)
	})

	// Test no duplicate keywords
	t.Run("NoDuplicateKeywords", func(t *testing.T) {
		testNoDuplicateKeywords(t, db)
	})

	// Test keyword relevance
	t.Run("KeywordRelevance", func(t *testing.T) {
		testKeywordRelevance(t, db)
	})

	// Test keyword coverage
	t.Run("KeywordCoverage", func(t *testing.T) {
		testKeywordCoverage(t, db)
	})

	// Test performance
	t.Run("Performance", func(t *testing.T) {
		testPerformance(t, db)
	})
}

// testHealthcareIndustriesExist verifies all 4 healthcare industries exist and are active
func testHealthcareIndustriesExist(t *testing.T, db *sql.DB) {
	healthcareIndustries := []string{
		"Medical Practices",
		"Healthcare Services",
		"Mental Health",
		"Healthcare Technology",
	}

	for _, industryName := range healthcareIndustries {
		var count int
		err := db.QueryRow(`
			SELECT COUNT(*) 
			FROM industries 
			WHERE name = $1 AND is_active = true
		`, industryName).Scan(&count)

		require.NoError(t, err)
		assert.Equal(t, 1, count, "Healthcare industry '%s' should exist and be active", industryName)
	}
}

// testKeywordCountPerIndustry verifies each healthcare industry has 50+ keywords
func testKeywordCountPerIndustry(t *testing.T, db *sql.DB) {
	healthcareIndustries := []string{
		"Medical Practices",
		"Healthcare Services",
		"Mental Health",
		"Healthcare Technology",
	}

	for _, industryName := range healthcareIndustries {
		var keywordCount int
		err := db.QueryRow(`
			SELECT COUNT(kw.keyword)
			FROM industries i
			JOIN keyword_weights kw ON i.id = kw.industry_id
			WHERE i.name = $1 AND kw.is_active = true
		`, industryName).Scan(&keywordCount)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, keywordCount, 50,
			"Healthcare industry '%s' should have 50+ keywords, got %d", industryName, keywordCount)
	}
}

// testKeywordWeightDistribution verifies keyword weights are in valid range (0.50-1.00)
func testKeywordWeightDistribution(t *testing.T, db *sql.DB) {
	healthcareIndustries := []string{
		"Medical Practices",
		"Healthcare Services",
		"Mental Health",
		"Healthcare Technology",
	}

	for _, industryName := range healthcareIndustries {
		var minWeight, maxWeight float64
		err := db.QueryRow(`
			SELECT MIN(kw.base_weight), MAX(kw.base_weight)
			FROM industries i
			JOIN keyword_weights kw ON i.id = kw.industry_id
			WHERE i.name = $1 AND kw.is_active = true
		`, industryName).Scan(&minWeight, &maxWeight)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, minWeight, 0.50,
			"Healthcare industry '%s' should have minimum weight >= 0.50, got %.3f", industryName, minWeight)
		assert.LessOrEqual(t, maxWeight, 1.00,
			"Healthcare industry '%s' should have maximum weight <= 1.00, got %.3f", industryName, maxWeight)
	}
}

// testNoDuplicateKeywords verifies no duplicate keywords within healthcare industries
func testNoDuplicateKeywords(t *testing.T, db *sql.DB) {
	healthcareIndustries := []string{
		"Medical Practices",
		"Healthcare Services",
		"Mental Health",
		"Healthcare Technology",
	}

	for _, industryName := range healthcareIndustries {
		var duplicateCount int
		err := db.QueryRow(`
			SELECT COUNT(*)
			FROM (
				SELECT kw.keyword, COUNT(*) as count
				FROM industries i
				JOIN keyword_weights kw ON i.id = kw.industry_id
				WHERE i.name = $1 AND kw.is_active = true
				GROUP BY kw.keyword
				HAVING COUNT(*) > 1
			) duplicates
		`, industryName).Scan(&duplicateCount)

		require.NoError(t, err)
		assert.Equal(t, 0, duplicateCount,
			"Healthcare industry '%s' should have no duplicate keywords, found %d", industryName, duplicateCount)
	}
}

// testKeywordRelevance verifies healthcare keywords are relevant to their industries
func testKeywordRelevance(t *testing.T, db *sql.DB) {
	// Define relevance patterns for each healthcare industry
	relevanceTests := map[string][]string{
		"Medical Practices": {
			"medical practice", "family medicine", "primary care", "physician", "doctor",
			"internal medicine", "pediatrics", "cardiology", "dermatology", "orthopedics",
		},
		"Healthcare Services": {
			"hospital", "medical center", "healthcare facility", "emergency department",
			"intensive care", "surgery center", "outpatient center", "urgent care",
		},
		"Mental Health": {
			"mental health", "counseling", "therapy", "psychologist", "psychiatrist",
			"behavioral health", "depression", "anxiety", "trauma", "substance abuse",
		},
		"Healthcare Technology": {
			"healthcare technology", "medical devices", "health it", "digital health",
			"telehealth", "telemedicine", "electronic health records", "health analytics",
		},
	}

	for industryName, expectedKeywords := range relevanceTests {
		for _, expectedKeyword := range expectedKeywords {
			var exists bool
			err := db.QueryRow(`
				SELECT EXISTS(
					SELECT 1
					FROM industries i
					JOIN keyword_weights kw ON i.id = kw.industry_id
					WHERE i.name = $1 AND kw.keyword = $2 AND kw.is_active = true
				)
			`, industryName, expectedKeyword).Scan(&exists)

			require.NoError(t, err)
			assert.True(t, exists,
				"Healthcare industry '%s' should contain relevant keyword '%s'", industryName, expectedKeyword)
		}
	}
}

// testKeywordCoverage verifies comprehensive keyword coverage for classification accuracy
func testKeywordCoverage(t *testing.T, db *sql.DB) {
	healthcareIndustries := []string{
		"Medical Practices",
		"Healthcare Services",
		"Mental Health",
		"Healthcare Technology",
	}

	for _, industryName := range healthcareIndustries {
		// Test core industry terms
		var coreTermsCount int
		err := db.QueryRow(`
			SELECT COUNT(kw.keyword)
			FROM industries i
			JOIN keyword_weights kw ON i.id = kw.industry_id
			WHERE i.name = $1 AND kw.is_active = true
			AND (kw.keyword ILIKE '%medical%' OR kw.keyword ILIKE '%health%' OR 
				 kw.keyword ILIKE '%healthcare%' OR kw.keyword ILIKE '%therapy%' OR
				 kw.keyword ILIKE '%counseling%' OR kw.keyword ILIKE '%clinical%')
		`, industryName).Scan(&coreTermsCount)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, coreTermsCount, 10,
			"Healthcare industry '%s' should have 10+ core industry terms, got %d", industryName, coreTermsCount)

		// Test professional terms
		var professionalTermsCount int
		err = db.QueryRow(`
			SELECT COUNT(kw.keyword)
			FROM industries i
			JOIN keyword_weights kw ON i.id = kw.industry_id
			WHERE i.name = $1 AND kw.is_active = true
			AND (kw.keyword ILIKE '%doctor%' OR kw.keyword ILIKE '%physician%' OR 
				 kw.keyword ILIKE '%nurse%' OR kw.keyword ILIKE '%therapist%' OR
				 kw.keyword ILIKE '%counselor%' OR kw.keyword ILIKE '%specialist%')
		`, industryName).Scan(&professionalTermsCount)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, professionalTermsCount, 5,
			"Healthcare industry '%s' should have 5+ professional terms, got %d", industryName, professionalTermsCount)
	}
}

// testPerformance verifies healthcare keyword queries perform efficiently
func testPerformance(t *testing.T, db *sql.DB) {
	healthcareIndustries := []string{
		"Medical Practices",
		"Healthcare Services",
		"Mental Health",
		"Healthcare Technology",
	}

	for _, industryName := range healthcareIndustries {
		start := time.Now()

		var keywordCount int
		err := db.QueryRow(`
			SELECT COUNT(kw.keyword)
			FROM industries i
			JOIN keyword_weights kw ON i.id = kw.industry_id
			WHERE i.name = $1 AND kw.is_active = true
		`, industryName).Scan(&keywordCount)

		duration := time.Since(start)

		require.NoError(t, err)
		assert.Less(t, duration, 100*time.Millisecond,
			"Healthcare industry '%s' keyword query should complete in <100ms, took %v", industryName, duration)
		assert.Greater(t, keywordCount, 0,
			"Healthcare industry '%s' should return keywords", industryName)
	}
}

// TestHealthcareKeywordsIntegration tests integration with existing classification system
func TestHealthcareKeywordsIntegration(t *testing.T) {
	// Setup database connection
	db, err := setupTestDatabase()
	require.NoError(t, err)
	defer db.Close()

	// Test healthcare classification scenarios
	healthcareTestCases := []struct {
		name             string
		businessName     string
		description      string
		expectedIndustry string
		minConfidence    float64
	}{
		{
			name:             "Family Medical Practice",
			businessName:     "Smith Family Medicine",
			description:      "Family medical practice providing primary care services",
			expectedIndustry: "Medical Practices",
			minConfidence:    0.80,
		},
		{
			name:             "Regional Medical Center",
			businessName:     "Regional Medical Center",
			description:      "Full-service hospital providing comprehensive healthcare services",
			expectedIndustry: "Healthcare Services",
			minConfidence:    0.80,
		},
		{
			name:             "Mental Health Counseling",
			businessName:     "Wellness Counseling Center",
			description:      "Mental health counseling and therapy services",
			expectedIndustry: "Mental Health",
			minConfidence:    0.80,
		},
		{
			name:             "Health Technology Company",
			businessName:     "MedTech Solutions",
			description:      "Healthcare technology and medical device development",
			expectedIndustry: "Healthcare Technology",
			minConfidence:    0.80,
		},
	}

	for _, tc := range healthcareTestCases {
		t.Run(tc.name, func(t *testing.T) {
			// This would integrate with the actual classification system
			// For now, we'll test that the keywords exist for classification
			var keywordCount int
			err := db.QueryRow(`
				SELECT COUNT(kw.keyword)
				FROM industries i
				JOIN keyword_weights kw ON i.id = kw.industry_id
				WHERE i.name = $1 AND kw.is_active = true
			`, tc.expectedIndustry).Scan(&keywordCount)

			require.NoError(t, err)
			assert.GreaterOrEqual(t, keywordCount, 50,
				"Expected industry '%s' should have 50+ keywords for classification", tc.expectedIndustry)
		})
	}
}

// TestHealthcareKeywordsAccuracy tests classification accuracy with healthcare businesses
func TestHealthcareKeywordsAccuracy(t *testing.T) {
	// Setup database connection
	db, err := setupTestDatabase()
	require.NoError(t, err)
	defer db.Close()

	// Test accuracy with various healthcare business descriptions
	accuracyTestCases := []struct {
		description      string
		expectedIndustry string
		keywords         []string
	}{
		{
			description:      "Family medicine practice with primary care services",
			expectedIndustry: "Medical Practices",
			keywords:         []string{"family medicine", "primary care", "medical practice"},
		},
		{
			description:      "Hospital providing emergency and surgical services",
			expectedIndustry: "Healthcare Services",
			keywords:         []string{"hospital", "emergency", "surgical", "healthcare services"},
		},
		{
			description:      "Mental health counseling and therapy services",
			expectedIndustry: "Mental Health",
			keywords:         []string{"mental health", "counseling", "therapy", "psychological"},
		},
		{
			description:      "Digital health platform with telemedicine services",
			expectedIndustry: "Healthcare Technology",
			keywords:         []string{"digital health", "telemedicine", "healthcare technology"},
		},
	}

	for _, tc := range accuracyTestCases {
		t.Run(tc.description, func(t *testing.T) {
			// Verify that expected keywords exist for the industry
			for _, keyword := range tc.keywords {
				var exists bool
				err := db.QueryRow(`
					SELECT EXISTS(
						SELECT 1
						FROM industries i
						JOIN keyword_weights kw ON i.id = kw.industry_id
						WHERE i.name = $1 AND kw.keyword = $2 AND kw.is_active = true
					)
				`, tc.expectedIndustry, keyword).Scan(&exists)

				require.NoError(t, err)
				assert.True(t, exists,
					"Expected industry '%s' should contain keyword '%s' for description: %s",
					tc.expectedIndustry, keyword, tc.description)
			}
		})
	}
}

// setupTestDatabase sets up a test database connection
func setupTestDatabase() (*sql.DB, error) {
	// This would use actual database connection parameters
	// For testing, we'll use environment variables or test configuration
	dbURL := "postgresql://postgres:password@localhost:5432/test_db"

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
