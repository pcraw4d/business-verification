package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"kyb-platform/internal/models"
)

// SeedTestMerchant creates a test merchant in the database
func SeedTestMerchant(db *sql.DB, merchantID, status string) error {
	now := time.Now()
	merchant := &models.Merchant{
		ID:               merchantID,
		Name:             "Test Business " + merchantID,
		LegalName:        "Test Business Legal Name " + merchantID,
		RegistrationNumber: "REG-" + merchantID,
		TaxID:            "TAX-" + merchantID,
		Industry:         "Technology",
		IndustryCode:     "541511",
		BusinessType:     "Corporation",
		EmployeeCount:    100,
		Address: models.Address{
			Street1:    "123 Test St",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "United States",
			CountryCode: "US",
		},
		ContactInfo: models.ContactInfo{
			Phone:   "+1-555-123-4567",
			Email:   "test@" + merchantID + ".com",
			Website: "https://test-" + merchantID + ".com",
		},
		PortfolioType:    models.PortfolioTypeOnboarded,
		RiskLevel:        models.RiskLevelMedium,
		ComplianceStatus: "compliant",
		Status:           status,
		CreatedBy:        "test-user",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Get portfolio type and risk level IDs
	ctx := context.Background()
	portfolioTypeID, err := getPortfolioTypeID(ctx, db, string(merchant.PortfolioType))
	if err != nil {
		return fmt.Errorf("failed to get portfolio type ID: %w", err)
	}

	riskLevelID, err := getRiskLevelID(ctx, db, string(merchant.RiskLevel))
	if err != nil {
		return fmt.Errorf("failed to get risk level ID: %w", err)
	}

	query := `
		INSERT INTO merchants (
			id, name, legal_name, registration_number, tax_id, industry, industry_code,
			business_type, founded_date, employee_count, annual_revenue,
			address_street1, address_street2, address_city, address_state, 
			address_postal_code, address_country, address_country_code,
			contact_phone, contact_email, contact_website, contact_primary_contact,
			portfolio_type_id, risk_level_id, compliance_status, status,
			created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17, $18,
			$19, $20, $21, $22,
			$23, $24, $25, $26, $27, $28, $29
		)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`

	_, err = db.ExecContext(ctx, query,
		merchant.ID,
		merchant.Name,
		merchant.LegalName,
		merchant.RegistrationNumber,
		merchant.TaxID,
		merchant.Industry,
		merchant.IndustryCode,
		merchant.BusinessType,
		merchant.FoundedDate,
		merchant.EmployeeCount,
		merchant.AnnualRevenue,
		merchant.Address.Street1,
		merchant.Address.Street2,
		merchant.Address.City,
		merchant.Address.State,
		merchant.Address.PostalCode,
		merchant.Address.Country,
		merchant.Address.CountryCode,
		merchant.ContactInfo.Phone,
		merchant.ContactInfo.Email,
		merchant.ContactInfo.Website,
		merchant.ContactInfo.PrimaryContact,
		portfolioTypeID,
		riskLevelID,
		merchant.ComplianceStatus,
		merchant.Status,
		merchant.CreatedBy,
		merchant.CreatedAt,
		merchant.UpdatedAt,
	)

	return err
}

// SeedTestAnalytics creates analytics data for a merchant
func SeedTestAnalytics(db *sql.DB, merchantID string, classification, security, quality, intelligence map[string]interface{}) error {
	ctx := context.Background()

	// Default values if not provided
	if classification == nil {
		classification = map[string]interface{}{
			"primaryIndustry": "Technology",
			"confidenceScore": 0.95,
			"riskLevel":       "low",
		}
	}
	if security == nil {
		security = map[string]interface{}{
			"trustScore": 0.8,
			"sslValid":   true,
		}
	}
	if quality == nil {
		quality = map[string]interface{}{
			"completenessScore": 0.9,
			"dataPoints":        100,
		}
	}
	if intelligence == nil {
		intelligence = map[string]interface{}{}
	}

	classificationJSON, _ := json.Marshal(classification)
	securityJSON, _ := json.Marshal(security)
	qualityJSON, _ := json.Marshal(quality)
	intelligenceJSON, _ := json.Marshal(intelligence)

	query := `
		INSERT INTO merchant_analytics (
			merchant_id, classification_data, security_data, quality_data, intelligence_data,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (merchant_id) DO UPDATE SET
			classification_data = EXCLUDED.classification_data,
			security_data = EXCLUDED.security_data,
			quality_data = EXCLUDED.quality_data,
			intelligence_data = EXCLUDED.intelligence_data,
			updated_at = EXCLUDED.updated_at
	`

	_, err := db.ExecContext(ctx, query,
		merchantID,
		classificationJSON,
		securityJSON,
		qualityJSON,
		intelligenceJSON,
	)

	return err
}

// SeedTestRiskAssessment creates a risk assessment for a merchant
func SeedTestRiskAssessment(db *sql.DB, merchantID, assessmentID, status string, result map[string]interface{}) error {
	ctx := context.Background()

	if result == nil {
		result = map[string]interface{}{
			"overallScore": 0.7,
			"riskLevel":    "medium",
			"factors":       []interface{}{},
		}
	}

	resultJSON, _ := json.Marshal(result)

	query := `
		INSERT INTO risk_assessments (
			id, merchant_id, status, result, progress, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			result = EXCLUDED.result,
			progress = EXCLUDED.progress,
			updated_at = EXCLUDED.updated_at
	`

	progress := 0
	if status == "completed" {
		progress = 100
	} else if status == "processing" {
		progress = 50
	}

	_, err := db.ExecContext(ctx, query,
		assessmentID,
		merchantID,
		status,
		resultJSON,
		progress,
	)

	return err
}

// SeedTestRiskIndicators creates risk indicators for a merchant
func SeedTestRiskIndicators(db *sql.DB, merchantID string, count int, severity string) error {
	ctx := context.Background()

	for i := 0; i < count; i++ {
		indicatorID := fmt.Sprintf("indicator-%s-%d", merchantID, i)
		indicatorType := "financial"
		if i%2 == 0 {
			indicatorType = "operational"
		}
		indicatorSeverity := severity
		if severity == "" {
			severities := []string{"low", "medium", "high", "critical"}
			indicatorSeverity = severities[i%len(severities)]
		}
		status := "active"
		if i%3 == 0 {
			status = "resolved"
		}

		query := `
			INSERT INTO risk_indicators (
				id, merchant_id, type, name, severity, status, description, score, detected_at, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW(), NOW())
			ON CONFLICT (id) DO UPDATE SET
				severity = EXCLUDED.severity,
				status = EXCLUDED.status,
				updated_at = EXCLUDED.updated_at
		`

		score := 0.5 + float64(i%5)*0.1

		_, err := db.ExecContext(ctx, query,
			indicatorID,
			merchantID,
			indicatorType,
			fmt.Sprintf("Test Indicator %d", i),
			indicatorSeverity,
			status,
			fmt.Sprintf("Test description for indicator %d", i),
			score,
		)

		if err != nil {
			return fmt.Errorf("failed to seed indicator %d: %w", i, err)
		}
	}

	return nil
}

// CleanupTestData removes all test data for a merchant
func CleanupTestData(db *sql.DB, merchantID string) error {
	ctx := context.Background()

	// Delete in order to respect foreign keys
	tables := []string{
		"risk_indicators",
		"risk_assessments",
		"merchant_analytics",
		"merchants",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s WHERE merchant_id = $1", table)
		if _, err := db.ExecContext(ctx, query, merchantID); err != nil {
			// Log but don't fail - table might not exist or have different structure
			log.Printf("Warning: Could not cleanup %s for merchant %s: %v", table, merchantID, err)
		}
	}

	return nil
}

// Helper functions to get portfolio type and risk level IDs
func getPortfolioTypeID(ctx context.Context, db *sql.DB, portfolioType string) (int, error) {
	var id int
	query := `SELECT id FROM portfolio_types WHERE name = $1 LIMIT 1`
	err := db.QueryRowContext(ctx, query, portfolioType).Scan(&id)
	if err == sql.ErrNoRows {
		// Create default portfolio type if it doesn't exist
		insertQuery := `INSERT INTO portfolio_types (name, description, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id`
		err = db.QueryRowContext(ctx, insertQuery, portfolioType, "Test portfolio type").Scan(&id)
		if err != nil {
			return 1, nil // Default to 1 if creation fails
		}
		return id, nil
	}
	return id, err
}

func getRiskLevelID(ctx context.Context, db *sql.DB, riskLevel string) (int, error) {
	var id int
	query := `SELECT id FROM risk_levels WHERE name = $1 LIMIT 1`
	err := db.QueryRowContext(ctx, query, riskLevel).Scan(&id)
	if err == sql.ErrNoRows {
		// Create default risk level if it doesn't exist
		insertQuery := `INSERT INTO risk_levels (name, description, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id`
		err = db.QueryRowContext(ctx, insertQuery, riskLevel, "Test risk level").Scan(&id)
		if err != nil {
			return 1, nil // Default to 1 if creation fails
		}
		return id, nil
	}
	return id, err
}

