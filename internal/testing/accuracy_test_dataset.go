package testing

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// AccuracyTestDataset manages the accuracy test dataset for comprehensive testing
type AccuracyTestDataset struct {
	db     *sql.DB
	logger *log.Logger
}

// NewAccuracyTestDataset creates a new accuracy test dataset manager
func NewAccuracyTestDataset(db *sql.DB, logger *log.Logger) *AccuracyTestDataset {
	if logger == nil {
		logger = log.Default()
	}
	return &AccuracyTestDataset{
		db:     db,
		logger: logger,
	}
}

// TestCase represents a single test case in the accuracy test dataset
type TestCase struct {
	ID                        int       `json:"id"`
	BusinessName              string    `json:"business_name"`
	BusinessDescription       string    `json:"business_description"`
	WebsiteURL                string    `json:"website_url"`
	ExpectedPrimaryIndustry   string    `json:"expected_primary_industry"`
	ExpectedIndustryConfidence float64   `json:"expected_industry_confidence"`
	ExpectedMCCCodes          []string  `json:"expected_mcc_codes"`
	ExpectedNAICSCodes        []string  `json:"expected_naics_codes"`
	ExpectedSICCodes          []string  `json:"expected_sic_codes"`
	TestCategory              string    `json:"test_category"`
	TestSubcategory           string    `json:"test_subcategory"`
	IsEdgeCase                bool      `json:"is_edge_case"`
	IsHighConfidence          bool      `json:"is_high_confidence"`
	ExpectedConfidenceMin     float64   `json:"expected_confidence_min"`
	BusinessSize              string    `json:"business_size"`
	BusinessType              string    `json:"business_type"`
	LocationCountry           string    `json:"location_country"`
	LocationState             string    `json:"location_state"`
	ManuallyVerified          bool      `json:"manually_verified"`
	VerifiedBy                string    `json:"verified_by"`
	VerifiedAt                time.Time `json:"verified_at"`
	Notes                     string    `json:"notes"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
	IsActive                  bool      `json:"is_active"`
}

// scanTestCase scans a row into a TestCase, handling NULL values
func scanTestCase(rows *sql.Rows) (*TestCase, error) {
	tc := &TestCase{}
	var verifiedAt sql.NullTime
	var verifiedBy, businessDescription, websiteURL, testSubcategory, businessSize, businessType, locationCountry, locationState, notes sql.NullString

	err := rows.Scan(
		&tc.ID, &tc.BusinessName, &businessDescription, &websiteURL,
		&tc.ExpectedPrimaryIndustry, &tc.ExpectedIndustryConfidence,
		pq.Array(&tc.ExpectedMCCCodes), pq.Array(&tc.ExpectedNAICSCodes), pq.Array(&tc.ExpectedSICCodes),
		&tc.TestCategory, &testSubcategory, &tc.IsEdgeCase, &tc.IsHighConfidence,
		&tc.ExpectedConfidenceMin, &businessSize, &businessType,
		&locationCountry, &locationState, &tc.ManuallyVerified,
		&verifiedBy, &verifiedAt, &notes, &tc.CreatedAt, &tc.UpdatedAt, &tc.IsActive,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan test case: %w", err)
	}

	// Handle NULL values
	if businessDescription.Valid {
		tc.BusinessDescription = businessDescription.String
	}
	if websiteURL.Valid {
		tc.WebsiteURL = websiteURL.String
	}
	if testSubcategory.Valid {
		tc.TestSubcategory = testSubcategory.String
	}
	if businessSize.Valid {
		tc.BusinessSize = businessSize.String
	}
	if businessType.Valid {
		tc.BusinessType = businessType.String
	}
	if locationCountry.Valid {
		tc.LocationCountry = locationCountry.String
	}
	if locationState.Valid {
		tc.LocationState = locationState.String
	}
	if notes.Valid {
		tc.Notes = notes.String
	}
	if verifiedBy.Valid {
		tc.VerifiedBy = verifiedBy.String
	}
	if verifiedAt.Valid {
		tc.VerifiedAt = verifiedAt.Time
	}

	return tc, nil
}

// LoadAllTestCases loads all active test cases from the database
func (atd *AccuracyTestDataset) LoadAllTestCases(ctx context.Context) ([]*TestCase, error) {
	query := `
		SELECT 
			id, business_name, business_description, website_url,
			expected_primary_industry, expected_industry_confidence,
			expected_mcc_codes, expected_naics_codes, expected_sic_codes,
			test_category, test_subcategory, is_edge_case, is_high_confidence,
			expected_confidence_min, business_size, business_type,
			location_country, location_state, manually_verified,
			verified_by, verified_at, notes, created_at, updated_at, is_active
		FROM accuracy_test_dataset
		WHERE is_active = true
		ORDER BY test_category, business_name
	`

	rows, err := atd.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query test cases: %w", err)
	}
	defer rows.Close()

	var testCases []*TestCase
	for rows.Next() {
		tc, err := scanTestCase(rows)
		if err != nil {
			return nil, err
		}
		testCases = append(testCases, tc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating test cases: %w", err)
	}

	atd.logger.Printf("✅ Loaded %d test cases from database", len(testCases))
	return testCases, nil
}

// LoadTestCasesByCategory loads test cases filtered by category
func (atd *AccuracyTestDataset) LoadTestCasesByCategory(ctx context.Context, category string) ([]*TestCase, error) {
	query := `
		SELECT 
			id, business_name, business_description, website_url,
			expected_primary_industry, expected_industry_confidence,
			expected_mcc_codes, expected_naics_codes, expected_sic_codes,
			test_category, test_subcategory, is_edge_case, is_high_confidence,
			expected_confidence_min, business_size, business_type,
			location_country, location_state, manually_verified,
			verified_by, verified_at, notes, created_at, updated_at, is_active
		FROM accuracy_test_dataset
		WHERE is_active = true AND test_category = $1
		ORDER BY business_name
	`

	rows, err := atd.db.QueryContext(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to query test cases by category: %w", err)
	}
	defer rows.Close()

	var testCases []*TestCase
	for rows.Next() {
		tc, err := scanTestCase(rows)
		if err != nil {
			return nil, err
		}
		testCases = append(testCases, tc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating test cases: %w", err)
	}

	atd.logger.Printf("✅ Loaded %d test cases for category: %s", len(testCases), category)
	return testCases, nil
}

// LoadTestCasesByIndustry loads test cases filtered by expected primary industry
func (atd *AccuracyTestDataset) LoadTestCasesByIndustry(ctx context.Context, industry string) ([]*TestCase, error) {
	query := `
		SELECT 
			id, business_name, business_description, website_url,
			expected_primary_industry, expected_industry_confidence,
			expected_mcc_codes, expected_naics_codes, expected_sic_codes,
			test_category, test_subcategory, is_edge_case, is_high_confidence,
			expected_confidence_min, business_size, business_type,
			location_country, location_state, manually_verified,
			verified_by, verified_at, notes, created_at, updated_at, is_active
		FROM accuracy_test_dataset
		WHERE is_active = true AND expected_primary_industry = $1
		ORDER BY business_name
	`

	rows, err := atd.db.QueryContext(ctx, query, industry)
	if err != nil {
		return nil, fmt.Errorf("failed to query test cases by industry: %w", err)
	}
	defer rows.Close()

	var testCases []*TestCase
	for rows.Next() {
		tc, err := scanTestCase(rows)
		if err != nil {
			return nil, err
		}
		testCases = append(testCases, tc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating test cases: %w", err)
	}

	atd.logger.Printf("✅ Loaded %d test cases for industry: %s", len(testCases), industry)
	return testCases, nil
}

// GetDatasetStatistics returns statistics about the test dataset
func (atd *AccuracyTestDataset) GetDatasetStatistics(ctx context.Context) (*DatasetStatistics, error) {
	stats := &DatasetStatistics{}

	// Total count
	err := atd.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM accuracy_test_dataset WHERE is_active = true
	`).Scan(&stats.TotalTestCases)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Count by category
	categoryRows, err := atd.db.QueryContext(ctx, `
		SELECT test_category, COUNT(*) 
		FROM accuracy_test_dataset 
		WHERE is_active = true 
		GROUP BY test_category
		ORDER BY COUNT(*) DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get category counts: %w", err)
	}
	defer categoryRows.Close()

	stats.CategoryCounts = make(map[string]int)
	for categoryRows.Next() {
		var category string
		var count int
		if err := categoryRows.Scan(&category, &count); err != nil {
			return nil, fmt.Errorf("failed to scan category count: %w", err)
		}
		stats.CategoryCounts[category] = count
	}

	// Count by industry
	industryRows, err := atd.db.QueryContext(ctx, `
		SELECT expected_primary_industry, COUNT(*) 
		FROM accuracy_test_dataset 
		WHERE is_active = true 
		GROUP BY expected_primary_industry
		ORDER BY COUNT(*) DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get industry counts: %w", err)
	}
	defer industryRows.Close()

	stats.IndustryCounts = make(map[string]int)
	for industryRows.Next() {
		var industry string
		var count int
		if err := industryRows.Scan(&industry, &count); err != nil {
			return nil, fmt.Errorf("failed to scan industry count: %w", err)
		}
		stats.IndustryCounts[industry] = count
	}

	// Edge cases count
	err = atd.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM accuracy_test_dataset 
		WHERE is_active = true AND is_edge_case = true
	`).Scan(&stats.EdgeCaseCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get edge case count: %w", err)
	}

	// High confidence count
	err = atd.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM accuracy_test_dataset 
		WHERE is_active = true AND is_high_confidence = true
	`).Scan(&stats.HighConfidenceCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get high confidence count: %w", err)
	}

	// Verified count
	err = atd.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM accuracy_test_dataset 
		WHERE is_active = true AND manually_verified = true
	`).Scan(&stats.VerifiedCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get verified count: %w", err)
	}

	return stats, nil
}

// DatasetStatistics represents statistics about the test dataset
type DatasetStatistics struct {
	TotalTestCases      int            `json:"total_test_cases"`
	CategoryCounts      map[string]int `json:"category_counts"`
	IndustryCounts      map[string]int `json:"industry_counts"`
	EdgeCaseCount       int            `json:"edge_case_count"`
	HighConfidenceCount int            `json:"high_confidence_count"`
	VerifiedCount       int            `json:"verified_count"`
}

// ValidateTestCase validates that a test case has all required fields
func (atd *AccuracyTestDataset) ValidateTestCase(tc *TestCase) error {
	if tc.BusinessName == "" {
		return fmt.Errorf("business_name is required")
	}
	if tc.TestCategory == "" {
		return fmt.Errorf("test_category is required")
	}
	if tc.ExpectedPrimaryIndustry == "" {
		return fmt.Errorf("expected_primary_industry is required")
	}
	if len(tc.ExpectedMCCCodes) == 0 && len(tc.ExpectedNAICSCodes) == 0 && len(tc.ExpectedSICCodes) == 0 {
		return fmt.Errorf("at least one expected code type (MCC, NAICS, or SIC) is required")
	}
	if tc.ExpectedIndustryConfidence < 0.0 || tc.ExpectedIndustryConfidence > 1.0 {
		return fmt.Errorf("expected_industry_confidence must be between 0.0 and 1.0")
	}
	return nil
}

// AddTestCase adds a new test case to the dataset
func (atd *AccuracyTestDataset) AddTestCase(ctx context.Context, tc *TestCase) error {
	if err := atd.ValidateTestCase(tc); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO accuracy_test_dataset (
			business_name, business_description, website_url,
			expected_primary_industry, expected_industry_confidence,
			expected_mcc_codes, expected_naics_codes, expected_sic_codes,
			test_category, test_subcategory, is_edge_case, is_high_confidence,
			expected_confidence_min, business_size, business_type,
			location_country, location_state, notes, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		ON CONFLICT (business_name) DO UPDATE SET
			business_description = EXCLUDED.business_description,
			website_url = EXCLUDED.website_url,
			expected_primary_industry = EXCLUDED.expected_primary_industry,
			expected_industry_confidence = EXCLUDED.expected_industry_confidence,
			expected_mcc_codes = EXCLUDED.expected_mcc_codes,
			expected_naics_codes = EXCLUDED.expected_naics_codes,
			expected_sic_codes = EXCLUDED.expected_sic_codes,
			test_category = EXCLUDED.test_category,
			test_subcategory = EXCLUDED.test_subcategory,
			is_edge_case = EXCLUDED.is_edge_case,
			is_high_confidence = EXCLUDED.is_high_confidence,
			expected_confidence_min = EXCLUDED.expected_confidence_min,
			business_size = EXCLUDED.business_size,
			business_type = EXCLUDED.business_type,
			location_country = EXCLUDED.location_country,
			location_state = EXCLUDED.location_state,
			notes = EXCLUDED.notes,
			updated_at = NOW()
		RETURNING id
	`

	var id int
	err := atd.db.QueryRowContext(ctx, query,
		tc.BusinessName, tc.BusinessDescription, tc.WebsiteURL,
		tc.ExpectedPrimaryIndustry, tc.ExpectedIndustryConfidence,
		tc.ExpectedMCCCodes, tc.ExpectedNAICSCodes, tc.ExpectedSICCodes,
		tc.TestCategory, tc.TestSubcategory, tc.IsEdgeCase, tc.IsHighConfidence,
		tc.ExpectedConfidenceMin, tc.BusinessSize, tc.BusinessType,
		tc.LocationCountry, tc.LocationState, tc.Notes, tc.IsActive,
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to insert test case: %w", err)
	}

	tc.ID = id
	atd.logger.Printf("✅ Added test case: %s (ID: %d)", tc.BusinessName, id)
	return nil
}

// ExportTestCases exports test cases to JSON format
func (atd *AccuracyTestDataset) ExportTestCases(ctx context.Context, testCases []*TestCase) ([]byte, error) {
	data := map[string]interface{}{
		"exported_at": time.Now().Format(time.RFC3339),
		"total_cases": len(testCases),
		"test_cases":  testCases,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal test cases: %w", err)
	}

	return jsonData, nil
}

