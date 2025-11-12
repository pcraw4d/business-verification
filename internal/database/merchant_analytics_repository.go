package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"kyb-platform/internal/models"
)

// MerchantAnalyticsRepository provides data access operations for merchant analytics
type MerchantAnalyticsRepository struct {
	db     *sql.DB
	logger *log.Logger
}

// NewMerchantAnalyticsRepository creates a new merchant analytics repository
func NewMerchantAnalyticsRepository(db *sql.DB, logger *log.Logger) *MerchantAnalyticsRepository {
	if logger == nil {
		logger = log.Default()
	}

	return &MerchantAnalyticsRepository{
		db:     db,
		logger: logger,
	}
}

// Repository errors
var (
	ErrClassificationNotFound = errors.New("classification not found")
	ErrSecurityDataNotFound   = errors.New("security data not found")
	ErrQualityDataNotFound    = errors.New("quality data not found")
)

// GetClassificationByMerchantID retrieves classification data for a merchant
func (r *MerchantAnalyticsRepository) GetClassificationByMerchantID(ctx context.Context, merchantID string) (*models.ClassificationData, error) {
	r.logger.Printf("Retrieving classification data for merchant: %s", merchantID)

	// Try to find classification by business_name matching merchant name
	// First, get merchant name
	var merchantName string
	merchantQuery := `SELECT name FROM merchants WHERE id = $1`
	err := r.db.QueryRowContext(ctx, merchantQuery, merchantID).Scan(&merchantName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrClassificationNotFound
		}
		return nil, fmt.Errorf("failed to get merchant: %w", err)
	}

	// Query business_classifications table
	query := `
		SELECT 
			primary_industry,
			secondary_industries,
			confidence_score,
			classification_metadata
		FROM business_classifications
		WHERE business_name = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var primaryIndustryJSON []byte
	var secondaryIndustriesJSON []byte
	var confidenceScore sql.NullFloat64
	var metadataJSON []byte

	err = r.db.QueryRowContext(ctx, query, merchantName).Scan(
		&primaryIndustryJSON,
		&secondaryIndustriesJSON,
		&confidenceScore,
		&metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Return empty classification with default values
			return &models.ClassificationData{
				PrimaryIndustry: "",
				ConfidenceScore: 0.0,
				RiskLevel:       "medium",
				MCCCodes:        []models.IndustryCode{},
				SICCodes:        []models.IndustryCode{},
				NAICSCodes:      []models.IndustryCode{},
			}, nil
		}
		return nil, fmt.Errorf("failed to retrieve classification: %w", err)
	}

	// Parse primary_industry JSON
	var primaryIndustry map[string]interface{}
	if len(primaryIndustryJSON) > 0 {
		if err := json.Unmarshal(primaryIndustryJSON, &primaryIndustry); err != nil {
			r.logger.Printf("Warning: failed to parse primary_industry JSON: %v", err)
		}
	}

	// Parse classification_metadata for codes
	var mccCodes, sicCodes, naicsCodes []models.IndustryCode
	if len(metadataJSON) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(metadataJSON, &metadata); err == nil {
			// Extract codes from metadata
			if mccData, ok := metadata["mcc_codes"].([]interface{}); ok {
				mccCodes = parseIndustryCodes(mccData)
			}
			if sicData, ok := metadata["sic_codes"].([]interface{}); ok {
				sicCodes = parseIndustryCodes(sicData)
			}
			if naicsData, ok := metadata["naics_codes"].([]interface{}); ok {
				naicsCodes = parseIndustryCodes(naicsData)
			}
		}
	}

	// Determine risk level from confidence score
	riskLevel := "medium"
	if confidenceScore.Valid {
		if confidenceScore.Float64 >= 0.8 {
			riskLevel = "low"
		} else if confidenceScore.Float64 < 0.5 {
			riskLevel = "high"
		}
	}

	primaryIndustryName := ""
	if primaryIndustry != nil {
		if name, ok := primaryIndustry["name"].(string); ok {
			primaryIndustryName = name
		} else if name, ok := primaryIndustry["industry"].(string); ok {
			primaryIndustryName = name
		}
	}

	confidence := 0.0
	if confidenceScore.Valid {
		confidence = confidenceScore.Float64
	}

	return &models.ClassificationData{
		PrimaryIndustry: primaryIndustryName,
		ConfidenceScore: confidence,
		RiskLevel:       riskLevel,
		MCCCodes:        mccCodes,
		SICCodes:        sicCodes,
		NAICSCodes:      naicsCodes,
	}, nil
}

// parseIndustryCodes parses industry codes from JSON metadata
func parseIndustryCodes(data []interface{}) []models.IndustryCode {
	codes := make([]models.IndustryCode, 0, len(data))
	for _, item := range data {
		if codeMap, ok := item.(map[string]interface{}); ok {
			code := models.IndustryCode{}
			if c, ok := codeMap["code"].(string); ok {
				code.Code = c
			}
			if d, ok := codeMap["description"].(string); ok {
				code.Description = d
			} else if d, ok := codeMap["desc"].(string); ok {
				code.Description = d
			}
			if conf, ok := codeMap["confidence"].(float64); ok {
				code.Confidence = conf
			} else {
				code.Confidence = 0.8 // Default confidence
			}
			codes = append(codes, code)
		}
	}
	return codes
}

// GetSecurityDataByMerchantID retrieves security data for a merchant
func (r *MerchantAnalyticsRepository) GetSecurityDataByMerchantID(ctx context.Context, merchantID string) (*models.SecurityData, error) {
	r.logger.Printf("Retrieving security data for merchant: %s", merchantID)

	// Get merchant website URL
	var websiteURL sql.NullString
	query := `SELECT contact_website FROM merchants WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, merchantID).Scan(&websiteURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSecurityDataNotFound
		}
		return nil, fmt.Errorf("failed to get merchant: %w", err)
	}

	// For now, return default security data
	// In a full implementation, this would query a security_checks table
	// or call an external service to check SSL and headers
	securityData := &models.SecurityData{
		TrustScore:      0.85, // Default trust score
		SSLValid:        true,
		SecurityHeaders: []models.SecurityHeader{},
	}

	// If website exists, we could check SSL here
	// For MVP, return default values
	if websiteURL.Valid && websiteURL.String != "" {
		// In production, this would make an HTTP request to check SSL
		// For now, assume valid SSL
		expiryDate := time.Now().Add(365 * 24 * time.Hour) // 1 year from now
		securityData.SSLExpiryDate = &expiryDate
	}

	return securityData, nil
}

// GetQualityMetricsByMerchantID retrieves quality metrics for a merchant
func (r *MerchantAnalyticsRepository) GetQualityMetricsByMerchantID(ctx context.Context, merchantID string) (*models.QualityData, error) {
	r.logger.Printf("Retrieving quality metrics for merchant: %s", merchantID)

	// Query merchant table to calculate completeness
	query := `
		SELECT 
			CASE WHEN name IS NOT NULL THEN 1 ELSE 0 END +
			CASE WHEN legal_name IS NOT NULL THEN 1 ELSE 0 END +
			CASE WHEN industry IS NOT NULL THEN 1 ELSE 0 END +
			CASE WHEN contact_email IS NOT NULL THEN 1 ELSE 0 END +
			CASE WHEN contact_phone IS NOT NULL THEN 1 ELSE 0 END +
			CASE WHEN contact_website IS NOT NULL THEN 1 ELSE 0 END +
			CASE WHEN address_street1 IS NOT NULL THEN 1 ELSE 0 END +
			CASE WHEN address_city IS NOT NULL THEN 1 ELSE 0 END +
			CASE WHEN address_state IS NOT NULL THEN 1 ELSE 0 END +
			CASE WHEN address_postal_code IS NOT NULL THEN 1 ELSE 0 END +
			CASE WHEN address_country IS NOT NULL THEN 1 ELSE 0 END +
			CASE WHEN employee_count IS NOT NULL THEN 1 ELSE 0 END +
			CASE WHEN annual_revenue IS NOT NULL THEN 1 ELSE 0 END +
			CASE WHEN founded_date IS NOT NULL THEN 1 ELSE 0 END +
			CASE WHEN tax_id IS NOT NULL THEN 1 ELSE 0 END +
			CASE WHEN registration_number IS NOT NULL THEN 1 ELSE 0 END
		AS filled_fields,
		16 AS total_fields
		FROM merchants
		WHERE id = $1
	`

	var filledFields, totalFields int
	err := r.db.QueryRowContext(ctx, query, merchantID).Scan(&filledFields, &totalFields)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrQualityDataNotFound
		}
		return nil, fmt.Errorf("failed to calculate quality metrics: %w", err)
	}

	completenessScore := float64(filledFields) / float64(totalFields)
	if totalFields == 0 {
		completenessScore = 0.0
	}

	// Get missing fields
	missingFieldsQuery := `
		SELECT 
			CASE WHEN name IS NULL THEN 'name' END,
			CASE WHEN legal_name IS NULL THEN 'legal_name' END,
			CASE WHEN industry IS NULL THEN 'industry' END,
			CASE WHEN contact_email IS NULL THEN 'email' END,
			CASE WHEN contact_phone IS NULL THEN 'phone' END,
			CASE WHEN contact_website IS NULL THEN 'website' END,
			CASE WHEN address_street1 IS NULL THEN 'address_street1' END,
			CASE WHEN address_city IS NULL THEN 'address_city' END,
			CASE WHEN address_state IS NULL THEN 'address_state' END,
			CASE WHEN address_postal_code IS NULL THEN 'address_postal_code' END,
			CASE WHEN address_country IS NULL THEN 'address_country' END,
			CASE WHEN employee_count IS NULL THEN 'employee_count' END,
			CASE WHEN annual_revenue IS NULL THEN 'annual_revenue' END,
			CASE WHEN founded_date IS NULL THEN 'founded_date' END,
			CASE WHEN tax_id IS NULL THEN 'tax_id' END,
			CASE WHEN registration_number IS NULL THEN 'registration_number' END
		FROM merchants
		WHERE id = $1
	`

	rows, err := r.db.QueryContext(ctx, missingFieldsQuery, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get missing fields: %w", err)
	}
	defer rows.Close()

	missingFields := []string{}
	for rows.Next() {
		var field sql.NullString
		if err := rows.Scan(&field); err == nil && field.Valid {
			missingFields = append(missingFields, field.String)
		}
	}

	// Count total data points (non-null fields)
	dataPoints := filledFields

	return &models.QualityData{
		CompletenessScore: completenessScore,
		DataPoints:        dataPoints,
		MissingFields:     missingFields,
	}, nil
}

// GetIntelligenceDataByMerchantID retrieves business intelligence data for a merchant
func (r *MerchantAnalyticsRepository) GetIntelligenceDataByMerchantID(ctx context.Context, merchantID string) (*models.IntelligenceData, error) {
	r.logger.Printf("Retrieving intelligence data for merchant: %s", merchantID)

	query := `
		SELECT 
			employee_count,
			annual_revenue,
			founded_date
		FROM merchants
		WHERE id = $1
	`

	var employeeCount sql.NullInt64
	var annualRevenue sql.NullFloat64
	var foundedDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, merchantID).Scan(
		&employeeCount,
		&annualRevenue,
		&foundedDate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return &models.IntelligenceData{}, nil
		}
		return nil, fmt.Errorf("failed to retrieve intelligence data: %w", err)
	}

	intelligence := &models.IntelligenceData{}

	if employeeCount.Valid {
		count := int(employeeCount.Int64)
		intelligence.EmployeeCount = &count
	}

	if annualRevenue.Valid {
		revenue := annualRevenue.Float64
		intelligence.AnnualRevenue = &revenue
	}

	if foundedDate.Valid {
		now := time.Now()
		age := int(now.Year() - foundedDate.Time.Year())
		if now.YearDay() < foundedDate.Time.YearDay() {
			age--
		}
		intelligence.BusinessAge = &age
	}

	return intelligence, nil
}

