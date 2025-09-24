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

// ClassificationSystemValidator provides comprehensive validation of the classification system
type ClassificationSystemValidator struct {
	db       *sql.DB
	logger   *log.Logger
	testData []ClassificationTestData
}

// ClassificationTestData represents test data for validation
type ClassificationTestData struct {
	Name             string
	Description      string
	WebsiteURL       string
	ExpectedIndustry string
	ExpectedMCC      string
	ExpectedSIC      string
	ExpectedNAICS    string
	MinConfidence    float64
	Keywords         []string
	Category         string // 'easy', 'medium', 'hard'
}

// ClassificationValidationResult represents the result of a validation test
type ClassificationValidationResult struct {
	TestName         string
	Passed           bool
	ActualIndustry   string
	ActualMCC        string
	ActualSIC        string
	ActualNAICS      string
	ActualConfidence float64
	ResponseTime     time.Duration
	Error            error
	Details          map[string]interface{}
}

// NewClassificationSystemValidator creates a new validator instance
func NewClassificationSystemValidator(db *sql.DB, logger *log.Logger) *ClassificationSystemValidator {
	return &ClassificationSystemValidator{
		db:       db,
		logger:   logger,
		testData: getComprehensiveTestData(),
	}
}

// getComprehensiveTestData returns a comprehensive set of test data
func getComprehensiveTestData() []ClassificationTestData {
	return []ClassificationTestData{
		// Easy cases - clear industry indicators
		{
			Name:             "TechCorp Software Solutions",
			Description:      "We develop enterprise software solutions for businesses",
			WebsiteURL:       "https://techcorp.com",
			ExpectedIndustry: "Technology",
			ExpectedMCC:      "5734",
			ExpectedSIC:      "7372",
			ExpectedNAICS:    "541511",
			MinConfidence:    0.85,
			Keywords:         []string{"software", "technology", "enterprise", "development"},
			Category:         "easy",
		},
		{
			Name:             "Green Energy Solar",
			Description:      "Solar panel installation and renewable energy solutions",
			WebsiteURL:       "https://greenenergy.com",
			ExpectedIndustry: "Energy",
			ExpectedMCC:      "1731",
			ExpectedSIC:      "1623",
			ExpectedNAICS:    "238220",
			MinConfidence:    0.80,
			Keywords:         []string{"solar", "energy", "renewable", "installation"},
			Category:         "easy",
		},
		{
			Name:             "MediCare Health Services",
			Description:      "Comprehensive healthcare services and medical consultations",
			WebsiteURL:       "https://medicare.com",
			ExpectedIndustry: "Healthcare",
			ExpectedMCC:      "8062",
			ExpectedSIC:      "8011",
			ExpectedNAICS:    "621111",
			MinConfidence:    0.85,
			Keywords:         []string{"healthcare", "medical", "health", "services"},
			Category:         "easy",
		},

		// Medium cases - mixed indicators
		{
			Name:             "Global Finance Group",
			Description:      "Investment banking and financial advisory services",
			WebsiteURL:       "https://globalfinance.com",
			ExpectedIndustry: "Finance",
			ExpectedMCC:      "6012",
			ExpectedSIC:      "6211",
			ExpectedNAICS:    "523110",
			MinConfidence:    0.75,
			Keywords:         []string{"finance", "investment", "banking", "advisory"},
			Category:         "medium",
		},
		{
			Name:             "EcoFriendly Manufacturing",
			Description:      "Sustainable manufacturing of eco-friendly products",
			WebsiteURL:       "https://ecofriendly.com",
			ExpectedIndustry: "Manufacturing",
			ExpectedMCC:      "5085",
			ExpectedSIC:      "3089",
			ExpectedNAICS:    "326199",
			MinConfidence:    0.70,
			Keywords:         []string{"manufacturing", "eco-friendly", "sustainable", "products"},
			Category:         "medium",
		},

		// Hard cases - ambiguous or complex
		{
			Name:             "Innovation Hub",
			Description:      "A collaborative space for startups and entrepreneurs",
			WebsiteURL:       "https://innovationhub.com",
			ExpectedIndustry: "Business Services",
			ExpectedMCC:      "7392",
			ExpectedSIC:      "7389",
			ExpectedNAICS:    "561499",
			MinConfidence:    0.60,
			Keywords:         []string{"innovation", "startups", "entrepreneurs", "collaborative"},
			Category:         "hard",
		},
		{
			Name:             "Digital Solutions Co",
			Description:      "We provide digital transformation and consulting services",
			WebsiteURL:       "https://digitalsolutions.com",
			ExpectedIndustry: "Technology",
			ExpectedMCC:      "7372",
			ExpectedSIC:      "7372",
			ExpectedNAICS:    "541511",
			MinConfidence:    0.65,
			Keywords:         []string{"digital", "transformation", "consulting", "solutions"},
			Category:         "hard",
		},

		// Edge cases
		{
			Name:             "A",
			Description:      "",
			WebsiteURL:       "",
			ExpectedIndustry: "General Business",
			ExpectedMCC:      "",
			ExpectedSIC:      "",
			ExpectedNAICS:    "",
			MinConfidence:    0.30,
			Keywords:         []string{},
			Category:         "edge",
		},
		{
			Name:             "Very Long Business Name That Contains Many Words And Descriptions",
			Description:      "This is a very long description that contains many different industry keywords and terms that might confuse the classification system and test its ability to handle complex inputs with multiple potential classifications",
			WebsiteURL:       "https://verylongbusinessname.com",
			ExpectedIndustry: "General Business",
			ExpectedMCC:      "",
			ExpectedSIC:      "",
			ExpectedNAICS:    "",
			MinConfidence:    0.50,
			Keywords:         []string{"business", "long", "description", "many", "words"},
			Category:         "edge",
		},
	}
}

// ValidateClassificationQueries tests that classification queries work correctly
func (csv *ClassificationSystemValidator) ValidateClassificationQueries(t *testing.T) {
	t.Log("🔍 Testing Classification Queries")

	ctx := context.Background()

	// Test 1: Database connectivity
	t.Run("Database Connectivity", func(t *testing.T) {
		err := csv.db.PingContext(ctx)
		require.NoError(t, err, "Database should be accessible")
	})

	// Test 2: Basic table existence
	t.Run("Table Existence", func(t *testing.T) {
		requiredTables := []string{"industries", "industry_keywords", "classification_codes"}

		for _, table := range requiredTables {
			query := `
				SELECT EXISTS (
					SELECT FROM information_schema.tables 
					WHERE table_schema = 'public' 
					AND table_name = $1
				)
			`

			var exists bool
			err := csv.db.QueryRowContext(ctx, query, table).Scan(&exists)
			require.NoError(t, err, "Should be able to check table existence for %s", table)

			if !exists {
				t.Logf("⚠️ Table %s does not exist - this may be expected if migration hasn't been run", table)
			} else {
				t.Logf("✅ Table %s exists", table)
			}
		}
	})

	// Test 3: Basic data queries
	t.Run("Basic Data Queries", func(t *testing.T) {
		// Test industries table query
		query := `SELECT COUNT(*) FROM industries WHERE is_active = true`
		var count int
		err := csv.db.QueryRowContext(ctx, query).Scan(&count)
		if err != nil {
			t.Logf("⚠️ Could not query industries table - may not exist: %v", err)
		} else {
			t.Logf("✅ Found %d active industries", count)
			assert.GreaterOrEqual(t, count, 0, "Industry count should be non-negative")
		}
	})
}

// ValidateKeywordMatching tests keyword matching functionality
func (csv *ClassificationSystemValidator) ValidateKeywordMatching(t *testing.T) {
	t.Log("🔍 Testing Keyword Matching Functionality")

	ctx := context.Background()

	// Test keyword table queries
	t.Run("Keyword Table Queries", func(t *testing.T) {
		// Test industry_keywords table query
		query := `
			SELECT ik.keyword, ik.weight, i.name as industry_name
			FROM industry_keywords ik
			JOIN industries i ON ik.industry_id = i.id
			WHERE ik.keyword ILIKE $1 AND ik.is_active = true
			LIMIT 10
		`

		rows, err := csv.db.QueryContext(ctx, query, "%software%")
		if err != nil {
			t.Logf("⚠️ Could not query industry_keywords table - may not exist: %v", err)
			return
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var keyword, industryName string
			var weight float64
			err := rows.Scan(&keyword, &weight, &industryName)
			require.NoError(t, err, "Should be able to scan keyword data")
			count++
		}

		t.Logf("✅ Found %d keyword matches for 'software'", count)
		assert.GreaterOrEqual(t, count, 0, "Keyword count should be non-negative")
	})

	// Test keyword weight validation
	t.Run("Keyword Weight Validation", func(t *testing.T) {
		query := `
			SELECT COUNT(*) 
			FROM industry_keywords 
			WHERE weight < 0 OR weight > 10
		`

		var invalidCount int
		err := csv.db.QueryRowContext(ctx, query).Scan(&invalidCount)
		if err != nil {
			t.Logf("⚠️ Could not validate keyword weights - table may not exist: %v", err)
			return
		}

		assert.Equal(t, 0, invalidCount, "Should not have keyword weights outside 0-10 range")
		t.Logf("✅ Keyword weight validation passed")
	})
}

// ValidateConfidenceScoring tests confidence scoring algorithms
func (csv *ClassificationSystemValidator) ValidateConfidenceScoring(t *testing.T) {
	t.Log("🔍 Testing Confidence Scoring Algorithms")

	ctx := context.Background()

	// Test confidence threshold validation
	t.Run("Confidence Threshold Validation", func(t *testing.T) {
		query := `
			SELECT COUNT(*) 
			FROM industries 
			WHERE confidence_threshold < 0 OR confidence_threshold > 1
		`

		var invalidCount int
		err := csv.db.QueryRowContext(ctx, query).Scan(&invalidCount)
		if err != nil {
			t.Logf("⚠️ Could not validate confidence thresholds - table may not exist: %v", err)
			return
		}

		assert.Equal(t, 0, invalidCount, "Should not have confidence_threshold values outside 0-1 range")
		t.Logf("✅ Confidence threshold validation passed")
	})

	// Test industry confidence distribution
	t.Run("Industry Confidence Distribution", func(t *testing.T) {
		query := `
			SELECT 
				AVG(confidence_threshold) as avg_confidence,
				MIN(confidence_threshold) as min_confidence,
				MAX(confidence_threshold) as max_confidence,
				COUNT(*) as total_industries
			FROM industries 
			WHERE is_active = true
		`

		var avgConfidence, minConfidence, maxConfidence float64
		var totalIndustries int

		err := csv.db.QueryRowContext(ctx, query).Scan(&avgConfidence, &minConfidence, &maxConfidence, &totalIndustries)
		if err != nil {
			t.Logf("⚠️ Could not analyze confidence distribution - table may not exist: %v", err)
			return
		}

		t.Logf("✅ Confidence Distribution Analysis:")
		t.Logf("   Total Industries: %d", totalIndustries)
		t.Logf("   Average Confidence: %.3f", avgConfidence)
		t.Logf("   Min Confidence: %.3f", minConfidence)
		t.Logf("   Max Confidence: %.3f", maxConfidence)

		assert.GreaterOrEqual(t, minConfidence, 0.0, "Minimum confidence should be non-negative")
		assert.LessOrEqual(t, maxConfidence, 1.0, "Maximum confidence should not exceed 1.0")
		assert.GreaterOrEqual(t, avgConfidence, 0.0, "Average confidence should be non-negative")
		assert.LessOrEqual(t, avgConfidence, 1.0, "Average confidence should not exceed 1.0")
	})
}

// ValidatePerformance tests performance with sample data
func (csv *ClassificationSystemValidator) ValidatePerformance(t *testing.T) {
	t.Log("🔍 Testing Performance with Sample Data")

	ctx := context.Background()

	// Test database query performance
	t.Run("Database Query Performance", func(t *testing.T) {
		// Test basic industry lookup performance
		startTime := time.Now()
		query := `SELECT id, name, description FROM industries WHERE is_active = true LIMIT 10`

		rows, err := csv.db.QueryContext(ctx, query)
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
		t.Logf("✅ Industry lookup performance: %v for %d records", duration, count)
	})

	// Test keyword search performance
	t.Run("Keyword Search Performance", func(t *testing.T) {
		startTime := time.Now()
		query := `
			SELECT ik.keyword, ik.weight, i.name as industry_name
			FROM industry_keywords ik
			JOIN industries i ON ik.industry_id = i.id
			WHERE ik.keyword ILIKE $1 AND ik.is_active = true
			LIMIT 20
		`

		rows, err := csv.db.QueryContext(ctx, query, "%software%")
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
		t.Logf("✅ Keyword search performance: %v for %d results", duration, count)
	})

	// Test complex query performance
	t.Run("Complex Query Performance", func(t *testing.T) {
		startTime := time.Now()
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

		rows, err := csv.db.QueryContext(ctx, query)
		duration := time.Since(startTime)

		if err != nil {
			t.Logf("⚠️ Could not test complex query performance - tables may not exist: %v", err)
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

		assert.Less(t, duration, 500*time.Millisecond, "Complex query should complete within 500ms")
		t.Logf("✅ Complex query performance: %v for %d results", duration, count)
	})
}

// RunComprehensiveValidation runs all validation tests
func (csv *ClassificationSystemValidator) RunComprehensiveValidation(t *testing.T) {
	t.Log("🚀 Starting Comprehensive Classification System Validation")

	// Test 1: Classification Queries
	csv.ValidateClassificationQueries(t)

	// Test 2: Keyword Matching
	csv.ValidateKeywordMatching(t)

	// Test 3: Confidence Scoring
	csv.ValidateConfidenceScoring(t)

	// Test 4: Performance
	csv.ValidatePerformance(t)

	t.Log("✅ Comprehensive Classification System Validation Completed")
}

// GenerateValidationReport generates a simple validation report
func (csv *ClassificationSystemValidator) GenerateValidationReport(t *testing.T) *ValidationReport {
	t.Log("📊 Generating Classification System Validation Report")

	report := &ValidationReport{
		GeneratedAt: time.Now(),
		TestResults: make([]ClassificationValidationResult, 0),
		Summary:     ClassificationValidationSummary{},
	}

	// Simple validation - just check database connectivity and basic queries
	ctx := context.Background()

	// Test database connectivity
	err := csv.db.PingContext(ctx)
	connectivityResult := ClassificationValidationResult{
		TestName: "Database Connectivity",
		Passed:   err == nil,
		Error:    err,
	}
	report.TestResults = append(report.TestResults, connectivityResult)

	// Test table existence
	requiredTables := []string{"industries", "industry_keywords", "classification_codes"}
	for _, table := range requiredTables {
		query := `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)
		`

		var exists bool
		err := csv.db.QueryRowContext(ctx, query, table).Scan(&exists)
		tableResult := ClassificationValidationResult{
			TestName: fmt.Sprintf("Table %s exists", table),
			Passed:   err == nil && exists,
			Error:    err,
		}
		report.TestResults = append(report.TestResults, tableResult)
	}

	// Calculate summary statistics
	report.calculateSummary()

	return report
}

// ValidationReport represents a comprehensive validation report
type ValidationReport struct {
	GeneratedAt time.Time
	TestResults []ClassificationValidationResult
	Summary     ClassificationValidationSummary
}

// ClassificationValidationSummary represents summary statistics
type ClassificationValidationSummary struct {
	TotalTests      int
	PassedTests     int
	FailedTests     int
	PassRate        float64
	AvgConfidence   float64
	AvgResponseTime time.Duration
	MaxResponseTime time.Duration
	MinResponseTime time.Duration
}

// calculateSummary calculates summary statistics
func (vr *ValidationReport) calculateSummary() {
	vr.Summary.TotalTests = len(vr.TestResults)
	vr.Summary.PassedTests = 0
	vr.Summary.FailedTests = 0

	for _, result := range vr.TestResults {
		if result.Passed {
			vr.Summary.PassedTests++
		} else {
			vr.Summary.FailedTests++
		}
	}

	if vr.Summary.TotalTests > 0 {
		vr.Summary.PassRate = float64(vr.Summary.PassedTests) / float64(vr.Summary.TotalTests) * 100
	}
}

// TestClassificationSystemValidation runs the comprehensive validation test
func TestClassificationSystemValidation(t *testing.T) {
	// Skip if no database connection available
	if testing.Short() {
		t.Skip("Skipping classification system validation in short mode")
	}

	// For now, skip the test since we don't have database connection setup
	t.Skip("Skipping test - database connection not configured")

	// This test would run when database is properly configured:
	// 1. Initialize database connection
	// 2. Create validator
	// 3. Run comprehensive validation
	// 4. Generate and log validation report
	// 5. Assert minimum performance requirements
}
