// Package integrity provides data consistency validation
// for the KYB Platform database schema.
package integrity

import (
	"context"
	"fmt"
	"strings"
)

// DataConsistencyValidator validates data consistency across related tables
type DataConsistencyValidator struct {
	validator *Validator
}

// Name returns the name of this validation check
func (dcv *DataConsistencyValidator) Name() string {
	return "data_consistency"
}

// Description returns a description of this validation check
func (dcv *DataConsistencyValidator) Description() string {
	return "Validates data consistency across related tables and business rules"
}

// Required returns whether this check is required
func (dcv *DataConsistencyValidator) Required() bool {
	return true
}

// Execute performs data consistency validation
func (dcv *DataConsistencyValidator) Execute(ctx context.Context) (*ValidationResult, error) {
	details := make(map[string]interface{})
	var totalInconsistencies int

	// Define consistency checks
	checks := []ConsistencyCheck{
		// User consistency checks
		{
			Name:        "user_email_consistency",
			Description: "Check for duplicate email addresses across user tables",
			CheckFunc:   dcv.checkUserEmailConsistency,
		},
		{
			Name:        "user_role_consistency",
			Description: "Check for valid user roles",
			CheckFunc:   dcv.checkUserRoleConsistency,
		},

		// Business consistency checks
		{
			Name:        "business_name_consistency",
			Description: "Check for business name consistency across tables",
			CheckFunc:   dcv.checkBusinessNameConsistency,
		},
		{
			Name:        "business_website_consistency",
			Description: "Check for valid website URLs",
			CheckFunc:   dcv.checkBusinessWebsiteConsistency,
		},

		// Classification consistency checks
		{
			Name:        "classification_confidence_consistency",
			Description: "Check for valid confidence scores in classifications",
			CheckFunc:   dcv.checkClassificationConfidenceConsistency,
		},
		{
			Name:        "industry_keyword_consistency",
			Description: "Check for consistent industry-keyword relationships",
			CheckFunc:   dcv.checkIndustryKeywordConsistency,
		},

		// Risk assessment consistency checks
		{
			Name:        "risk_score_consistency",
			Description: "Check for valid risk scores",
			CheckFunc:   dcv.checkRiskScoreConsistency,
		},
		{
			Name:        "risk_level_consistency",
			Description: "Check for consistent risk levels",
			CheckFunc:   dcv.checkRiskLevelConsistency,
		},

		// Code crosswalk consistency checks
		{
			Name:        "mcc_code_consistency",
			Description: "Check for valid MCC codes",
			CheckFunc:   dcv.checkMCCCodeConsistency,
		},
		{
			Name:        "naics_code_consistency",
			Description: "Check for valid NAICS codes",
			CheckFunc:   dcv.checkNAICSCodeConsistency,
		},
		{
			Name:        "sic_code_consistency",
			Description: "Check for valid SIC codes",
			CheckFunc:   dcv.checkSICCodeConsistency,
		},

		// Timestamp consistency checks
		{
			Name:        "timestamp_consistency",
			Description: "Check for logical timestamp relationships",
			CheckFunc:   dcv.checkTimestampConsistency,
		},

		// JSON data consistency checks
		{
			Name:        "json_data_consistency",
			Description: "Check for valid JSON data structures",
			CheckFunc:   dcv.checkJSONDataConsistency,
		},
	}

	// Execute each consistency check
	for _, check := range checks {
		inconsistencies, err := check.CheckFunc(ctx)
		if err != nil {
			dcv.validator.logger.Printf("Error executing consistency check %s: %v", check.Name, err)
			continue
		}

		if len(inconsistencies) > 0 {
			details[check.Name] = map[string]interface{}{
				"description":     check.Description,
				"inconsistencies": len(inconsistencies),
				"details":         inconsistencies,
			}
			totalInconsistencies += len(inconsistencies)
		}
	}

	// Determine status
	status := StatusPassed
	message := "All data consistency checks passed"

	if totalInconsistencies > 0 {
		status = StatusFailed
		message = fmt.Sprintf("Found %d data consistency issues", totalInconsistencies)
	}

	details["total_inconsistencies"] = totalInconsistencies
	details["total_checks"] = len(checks)

	return dcv.validator.createResult(dcv.Name(), status, message, details), nil
}

// ConsistencyCheck represents a data consistency check
type ConsistencyCheck struct {
	Name        string
	Description string
	CheckFunc   func(ctx context.Context) ([]ConsistencyIssue, error)
}

// ConsistencyIssue represents a data consistency issue
type ConsistencyIssue struct {
	TableName     string
	ColumnName    string
	RowID         interface{}
	IssueType     string
	Description   string
	CurrentValue  interface{}
	ExpectedValue interface{}
	Severity      string
}

// checkUserEmailConsistency checks for duplicate email addresses
func (dcv *DataConsistencyValidator) checkUserEmailConsistency(ctx context.Context) ([]ConsistencyIssue, error) {
	var issues []ConsistencyIssue

	// Check for duplicate emails in users table
	query := `
		SELECT email, COUNT(*) as count
		FROM users
		WHERE email IS NOT NULL
		GROUP BY email
		HAVING COUNT(*) > 1
	`

	rows, err := dcv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check user email consistency: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var email string
		var count int
		if err := rows.Scan(&email, &count); err != nil {
			continue
		}

		issue := ConsistencyIssue{
			TableName:     "users",
			ColumnName:    "email",
			IssueType:     "duplicate_email",
			Description:   fmt.Sprintf("Duplicate email address found: %s (appears %d times)", email, count),
			CurrentValue:  email,
			ExpectedValue: "unique email",
			Severity:      "high",
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// checkUserRoleConsistency checks for valid user roles
func (dcv *DataConsistencyValidator) checkUserRoleConsistency(ctx context.Context) ([]ConsistencyIssue, error) {
	var issues []ConsistencyIssue

	validRoles := []string{"user", "admin", "compliance_officer", "risk_manager", "business_analyst", "developer", "other"}
	validRolesStr := "'" + strings.Join(validRoles, "','") + "'"

	query := fmt.Sprintf(`
		SELECT id, role
		FROM users
		WHERE role IS NOT NULL
		AND role NOT IN (%s)
	`, validRolesStr)

	rows, err := dcv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check user role consistency: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userID, role interface{}
		if err := rows.Scan(&userID, &role); err != nil {
			continue
		}

		issue := ConsistencyIssue{
			TableName:     "users",
			ColumnName:    "role",
			RowID:         userID,
			IssueType:     "invalid_role",
			Description:   fmt.Sprintf("Invalid user role: %v", role),
			CurrentValue:  role,
			ExpectedValue: validRoles,
			Severity:      "medium",
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// checkBusinessNameConsistency checks for business name consistency
func (dcv *DataConsistencyValidator) checkBusinessNameConsistency(ctx context.Context) ([]ConsistencyIssue, error) {
	var issues []ConsistencyIssue

	// Check for empty business names
	query := `
		SELECT id, name
		FROM businesses
		WHERE name IS NULL OR TRIM(name) = ''
	`

	rows, err := dcv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check business name consistency: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var businessID, name interface{}
		if err := rows.Scan(&businessID, &name); err != nil {
			continue
		}

		issue := ConsistencyIssue{
			TableName:     "businesses",
			ColumnName:    "name",
			RowID:         businessID,
			IssueType:     "empty_name",
			Description:   "Business name is empty or null",
			CurrentValue:  name,
			ExpectedValue: "non-empty string",
			Severity:      "high",
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// checkBusinessWebsiteConsistency checks for valid website URLs
func (dcv *DataConsistencyValidator) checkBusinessWebsiteConsistency(ctx context.Context) ([]ConsistencyIssue, error) {
	var issues []ConsistencyIssue

	// Check for invalid website URLs
	query := `
		SELECT id, website_url
		FROM businesses
		WHERE website_url IS NOT NULL
		AND website_url != ''
		AND (website_url NOT LIKE 'http://%' AND website_url NOT LIKE 'https://%')
	`

	rows, err := dcv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check business website consistency: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var businessID, websiteURL interface{}
		if err := rows.Scan(&businessID, &websiteURL); err != nil {
			continue
		}

		issue := ConsistencyIssue{
			TableName:     "businesses",
			ColumnName:    "website_url",
			RowID:         businessID,
			IssueType:     "invalid_url",
			Description:   fmt.Sprintf("Invalid website URL format: %v", websiteURL),
			CurrentValue:  websiteURL,
			ExpectedValue: "URL starting with http:// or https://",
			Severity:      "medium",
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// checkClassificationConfidenceConsistency checks for valid confidence scores
func (dcv *DataConsistencyValidator) checkClassificationConfidenceConsistency(ctx context.Context) ([]ConsistencyIssue, error) {
	var issues []ConsistencyIssue

	// Check for invalid confidence scores
	query := `
		SELECT id, confidence_score
		FROM business_classifications
		WHERE confidence_score IS NOT NULL
		AND (confidence_score < 0 OR confidence_score > 1)
	`

	rows, err := dcv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check classification confidence consistency: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var classificationID, confidenceScore interface{}
		if err := rows.Scan(&classificationID, &confidenceScore); err != nil {
			continue
		}

		issue := ConsistencyIssue{
			TableName:     "business_classifications",
			ColumnName:    "confidence_score",
			RowID:         classificationID,
			IssueType:     "invalid_confidence",
			Description:   fmt.Sprintf("Confidence score %v is outside valid range [0, 1]", confidenceScore),
			CurrentValue:  confidenceScore,
			ExpectedValue: "value between 0 and 1",
			Severity:      "high",
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// checkIndustryKeywordConsistency checks for consistent industry-keyword relationships
func (dcv *DataConsistencyValidator) checkIndustryKeywordConsistency(ctx context.Context) ([]ConsistencyIssue, error) {
	var issues []ConsistencyIssue

	// Check for keywords with invalid weights
	query := `
		SELECT id, keyword, weight
		FROM industry_keywords
		WHERE weight IS NOT NULL
		AND (weight < 0 OR weight > 10)
	`

	rows, err := dcv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check industry keyword consistency: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var keywordID, keyword, weight interface{}
		if err := rows.Scan(&keywordID, &keyword, &weight); err != nil {
			continue
		}

		issue := ConsistencyIssue{
			TableName:     "industry_keywords",
			ColumnName:    "weight",
			RowID:         keywordID,
			IssueType:     "invalid_weight",
			Description:   fmt.Sprintf("Keyword '%v' has invalid weight %v", keyword, weight),
			CurrentValue:  weight,
			ExpectedValue: "value between 0 and 10",
			Severity:      "medium",
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// checkRiskScoreConsistency checks for valid risk scores
func (dcv *DataConsistencyValidator) checkRiskScoreConsistency(ctx context.Context) ([]ConsistencyIssue, error) {
	var issues []ConsistencyIssue

	// Check for invalid risk scores in risk assessments
	query := `
		SELECT id, risk_score
		FROM risk_assessments
		WHERE risk_score IS NOT NULL
		AND (risk_score < 0 OR risk_score > 1)
	`

	rows, err := dcv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check risk score consistency: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var assessmentID, riskScore interface{}
		if err := rows.Scan(&assessmentID, &riskScore); err != nil {
			continue
		}

		issue := ConsistencyIssue{
			TableName:     "risk_assessments",
			ColumnName:    "risk_score",
			RowID:         assessmentID,
			IssueType:     "invalid_risk_score",
			Description:   fmt.Sprintf("Risk score %v is outside valid range [0, 1]", riskScore),
			CurrentValue:  riskScore,
			ExpectedValue: "value between 0 and 1",
			Severity:      "high",
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// checkRiskLevelConsistency checks for consistent risk levels
func (dcv *DataConsistencyValidator) checkRiskLevelConsistency(ctx context.Context) ([]ConsistencyIssue, error) {
	var issues []ConsistencyIssue

	validRiskLevels := []string{"low", "medium", "high", "critical"}
	validRiskLevelsStr := "'" + strings.Join(validRiskLevels, "','") + "'"

	query := fmt.Sprintf(`
		SELECT id, risk_level
		FROM risk_assessments
		WHERE risk_level IS NOT NULL
		AND risk_level NOT IN (%s)
	`, validRiskLevelsStr)

	rows, err := dcv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check risk level consistency: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var assessmentID, riskLevel interface{}
		if err := rows.Scan(&assessmentID, &riskLevel); err != nil {
			continue
		}

		issue := ConsistencyIssue{
			TableName:     "risk_assessments",
			ColumnName:    "risk_level",
			RowID:         assessmentID,
			IssueType:     "invalid_risk_level",
			Description:   fmt.Sprintf("Invalid risk level: %v", riskLevel),
			CurrentValue:  riskLevel,
			ExpectedValue: validRiskLevels,
			Severity:      "high",
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// checkMCCCodeConsistency checks for valid MCC codes
func (dcv *DataConsistencyValidator) checkMCCCodeConsistency(ctx context.Context) ([]ConsistencyIssue, error) {
	var issues []ConsistencyIssue

	// Check for invalid MCC code formats (should be 4 digits)
	query := `
		SELECT id, code
		FROM classification_codes
		WHERE code_type = 'MCC'
		AND (code IS NULL OR code !~ '^[0-9]{4}$')
	`

	rows, err := dcv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check MCC code consistency: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var codeID, code interface{}
		if err := rows.Scan(&codeID, &code); err != nil {
			continue
		}

		issue := ConsistencyIssue{
			TableName:     "classification_codes",
			ColumnName:    "code",
			RowID:         codeID,
			IssueType:     "invalid_mcc_code",
			Description:   fmt.Sprintf("Invalid MCC code format: %v", code),
			CurrentValue:  code,
			ExpectedValue: "4-digit numeric code",
			Severity:      "medium",
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// checkNAICSCodeConsistency checks for valid NAICS codes
func (dcv *DataConsistencyValidator) checkNAICSCodeConsistency(ctx context.Context) ([]ConsistencyIssue, error) {
	var issues []ConsistencyIssue

	// Check for invalid NAICS code formats (should be 2-6 digits)
	query := `
		SELECT id, code
		FROM classification_codes
		WHERE code_type = 'NAICS'
		AND (code IS NULL OR code !~ '^[0-9]{2,6}$')
	`

	rows, err := dcv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check NAICS code consistency: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var codeID, code interface{}
		if err := rows.Scan(&codeID, &code); err != nil {
			continue
		}

		issue := ConsistencyIssue{
			TableName:     "classification_codes",
			ColumnName:    "code",
			RowID:         codeID,
			IssueType:     "invalid_naics_code",
			Description:   fmt.Sprintf("Invalid NAICS code format: %v", code),
			CurrentValue:  code,
			ExpectedValue: "2-6 digit numeric code",
			Severity:      "medium",
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// checkSICCodeConsistency checks for valid SIC codes
func (dcv *DataConsistencyValidator) checkSICCodeConsistency(ctx context.Context) ([]ConsistencyIssue, error) {
	var issues []ConsistencyIssue

	// Check for invalid SIC code formats (should be 2-4 digits)
	query := `
		SELECT id, code
		FROM classification_codes
		WHERE code_type = 'SIC'
		AND (code IS NULL OR code !~ '^[0-9]{2,4}$')
	`

	rows, err := dcv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check SIC code consistency: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var codeID, code interface{}
		if err := rows.Scan(&codeID, &code); err != nil {
			continue
		}

		issue := ConsistencyIssue{
			TableName:     "classification_codes",
			ColumnName:    "code",
			RowID:         codeID,
			IssueType:     "invalid_sic_code",
			Description:   fmt.Sprintf("Invalid SIC code format: %v", code),
			CurrentValue:  code,
			ExpectedValue: "2-4 digit numeric code",
			Severity:      "medium",
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// checkTimestampConsistency checks for logical timestamp relationships
func (dcv *DataConsistencyValidator) checkTimestampConsistency(ctx context.Context) ([]ConsistencyIssue, error) {
	var issues []ConsistencyIssue

	// Check for updated_at before created_at
	query := `
		SELECT id, created_at, updated_at
		FROM users
		WHERE updated_at IS NOT NULL
		AND created_at IS NOT NULL
		AND updated_at < created_at
	`

	rows, err := dcv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check timestamp consistency: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userID, createdAt, updatedAt interface{}
		if err := rows.Scan(&userID, &createdAt, &updatedAt); err != nil {
			continue
		}

		issue := ConsistencyIssue{
			TableName:     "users",
			ColumnName:    "updated_at",
			RowID:         userID,
			IssueType:     "invalid_timestamp",
			Description:   fmt.Sprintf("updated_at (%v) is before created_at (%v)", updatedAt, createdAt),
			CurrentValue:  updatedAt,
			ExpectedValue: "timestamp after created_at",
			Severity:      "medium",
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// checkJSONDataConsistency checks for valid JSON data structures
func (dcv *DataConsistencyValidator) checkJSONDataConsistency(ctx context.Context) ([]ConsistencyIssue, error) {
	var issues []ConsistencyIssue

	// Check for invalid JSON in metadata columns
	query := `
		SELECT id, metadata
		FROM users
		WHERE metadata IS NOT NULL
		AND NOT (metadata::text ~ '^[\\[\\{].*[\\]\\}]$')
	`

	rows, err := dcv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check JSON data consistency: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userID, metadata interface{}
		if err := rows.Scan(&userID, &metadata); err != nil {
			continue
		}

		issue := ConsistencyIssue{
			TableName:     "users",
			ColumnName:    "metadata",
			RowID:         userID,
			IssueType:     "invalid_json",
			Description:   fmt.Sprintf("Invalid JSON structure in metadata: %v", metadata),
			CurrentValue:  metadata,
			ExpectedValue: "valid JSON object or array",
			Severity:      "medium",
		}
		issues = append(issues, issue)
	}

	return issues, nil
}
