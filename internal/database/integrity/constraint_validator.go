// Package integrity provides constraint validation
// for the KYB Platform database schema.
package integrity

import (
	"context"
	"fmt"
)

// ConstraintValidator validates database constraints
type ConstraintValidator struct {
	validator *Validator
}

// Name returns the name of this validation check
func (cv *ConstraintValidator) Name() string {
	return "constraints"
}

// Description returns a description of this validation check
func (cv *ConstraintValidator) Description() string {
	return "Validates database constraints for data integrity"
}

// Required returns whether this check is required
func (cv *ConstraintValidator) Required() bool {
	return false
}

// Execute performs constraint validation
func (cv *ConstraintValidator) Execute(ctx context.Context) (*ValidationResult, error) {
	details := make(map[string]interface{})
	var totalIssues int

	// Check for missing constraints
	missingConstraints, err := cv.checkMissingConstraints(ctx)
	if err != nil {
		return cv.validator.createResult(
			cv.Name(),
			StatusFailed,
			cv.validator.formatError(cv.Name(), "Failed to check missing constraints", err),
			details,
		), nil
	}

	if len(missingConstraints) > 0 {
		details["missing_constraints"] = missingConstraints
		totalIssues += len(missingConstraints)
	}

	// Check for constraint violations
	constraintViolations, err := cv.checkConstraintViolations(ctx)
	if err != nil {
		return cv.validator.createResult(
			cv.Name(),
			StatusFailed,
			cv.validator.formatError(cv.Name(), "Failed to check constraint violations", err),
			details,
		), nil
	}

	if len(constraintViolations) > 0 {
		details["constraint_violations"] = constraintViolations
		totalIssues += len(constraintViolations)
	}

	// Check for check constraint violations
	checkViolations, err := cv.checkCheckConstraintViolations(ctx)
	if err != nil {
		return cv.validator.createResult(
			cv.Name(),
			StatusFailed,
			cv.validator.formatError(cv.Name(), "Failed to check check constraint violations", err),
			details,
		), nil
	}

	if len(checkViolations) > 0 {
		details["check_constraint_violations"] = checkViolations
		totalIssues += len(checkViolations)
	}

	// Determine status
	status := StatusPassed
	message := "All constraint checks passed"

	if totalIssues > 0 {
		status = StatusFailed
		message = fmt.Sprintf("Found %d constraint issues", totalIssues)
	}

	details["total_issues"] = totalIssues

	return cv.validator.createResult(cv.Name(), status, message, details), nil
}

// checkMissingConstraints checks for missing critical constraints
func (cv *ConstraintValidator) checkMissingConstraints(ctx context.Context) ([]MissingConstraint, error) {
	var missingConstraints []MissingConstraint

	// Define expected constraints for KYB platform
	expectedConstraints := []ExpectedConstraint{
		// User table constraints
		{TableName: "users", ConstraintName: "users_email_key", ConstraintType: "UNIQUE", ColumnName: "email"},
		{TableName: "users", ConstraintName: "users_role_check", ConstraintType: "CHECK", ColumnName: "role"},
		{TableName: "users", ConstraintName: "users_status_check", ConstraintType: "CHECK", ColumnName: "status"},

		// Business table constraints
		{TableName: "businesses", ConstraintName: "businesses_user_id_fkey", ConstraintType: "FOREIGN KEY", ColumnName: "user_id"},
		{TableName: "businesses", ConstraintName: "businesses_name_check", ConstraintType: "CHECK", ColumnName: "name"},

		// Classification table constraints
		{TableName: "industries", ConstraintName: "industries_name_key", ConstraintType: "UNIQUE", ColumnName: "name"},
		{TableName: "industries", ConstraintName: "industries_category_check", ConstraintType: "CHECK", ColumnName: "category"},
		{TableName: "industry_keywords", ConstraintName: "industry_keywords_industry_id_fkey", ConstraintType: "FOREIGN KEY", ColumnName: "industry_id"},
		{TableName: "classification_codes", ConstraintName: "classification_codes_industry_id_fkey", ConstraintType: "FOREIGN KEY", ColumnName: "industry_id"},
		{TableName: "classification_codes", ConstraintName: "classification_codes_code_type_check", ConstraintType: "CHECK", ColumnName: "code_type"},

		// Risk table constraints
		{TableName: "risk_keywords", ConstraintName: "risk_keywords_risk_category_check", ConstraintType: "CHECK", ColumnName: "risk_category"},
		{TableName: "risk_keywords", ConstraintName: "risk_keywords_risk_severity_check", ConstraintType: "CHECK", ColumnName: "risk_severity"},
		{TableName: "business_risk_assessments", ConstraintName: "business_risk_assessments_business_id_fkey", ConstraintType: "FOREIGN KEY", ColumnName: "business_id"},
		{TableName: "business_risk_assessments", ConstraintName: "business_risk_assessments_risk_keyword_id_fkey", ConstraintType: "FOREIGN KEY", ColumnName: "risk_keyword_id"},
		{TableName: "business_risk_assessments", ConstraintName: "business_risk_assessments_risk_level_check", ConstraintType: "CHECK", ColumnName: "risk_level"},
	}

	// Get existing constraints
	existingConstraints, err := cv.getExistingConstraints(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing constraints: %w", err)
	}

	// Check for missing constraints
	for _, expected := range expectedConstraints {
		// Check if table exists
		exists, err := cv.validator.tableExists(ctx, expected.TableName)
		if err != nil || !exists {
			continue
		}

		found := false
		for _, existing := range existingConstraints {
			if existing.TableName == expected.TableName && existing.ConstraintName == expected.ConstraintName {
				found = true
				break
			}
		}

		if !found {
			missing := MissingConstraint{
				TableName:      expected.TableName,
				ConstraintName: expected.ConstraintName,
				ConstraintType: expected.ConstraintType,
				ColumnName:     expected.ColumnName,
				Reason:         fmt.Sprintf("Missing %s constraint %s on %s.%s", expected.ConstraintType, expected.ConstraintName, expected.TableName, expected.ColumnName),
			}
			missingConstraints = append(missingConstraints, missing)
		}
	}

	return missingConstraints, nil
}

// checkConstraintViolations checks for constraint violations
func (cv *ConstraintValidator) checkConstraintViolations(ctx context.Context) ([]ConstraintViolation, error) {
	var violations []ConstraintViolation

	// Check for unique constraint violations
	uniqueViolations, err := cv.checkUniqueConstraintViolations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check unique constraint violations: %w", err)
	}
	violations = append(violations, uniqueViolations...)

	// Check for foreign key constraint violations
	fkViolations, err := cv.checkForeignKeyConstraintViolations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check foreign key constraint violations: %w", err)
	}
	violations = append(violations, fkViolations...)

	return violations, nil
}

// checkUniqueConstraintViolations checks for unique constraint violations
func (cv *ConstraintValidator) checkUniqueConstraintViolations(ctx context.Context) ([]ConstraintViolation, error) {
	var violations []ConstraintViolation

	// Check for duplicate emails in users table
	query := `
		SELECT email, COUNT(*) as count
		FROM users
		WHERE email IS NOT NULL
		GROUP BY email
		HAVING COUNT(*) > 1
	`

	rows, err := cv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check unique constraint violations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var email string
		var count int
		if err := rows.Scan(&email, &count); err != nil {
			continue
		}

		violation := ConstraintViolation{
			TableName:      "users",
			ConstraintName: "users_email_key",
			ConstraintType: "UNIQUE",
			ColumnName:     "email",
			ViolatingValue: email,
			Description:    fmt.Sprintf("Duplicate email address: %s (appears %d times)", email, count),
		}
		violations = append(violations, violation)
	}

	return violations, nil
}

// checkForeignKeyConstraintViolations checks for foreign key constraint violations
func (cv *ConstraintValidator) checkForeignKeyConstraintViolations(ctx context.Context) ([]ConstraintViolation, error) {
	var violations []ConstraintViolation

	// Check for orphaned records in businesses table
	query := `
		SELECT id, user_id
		FROM businesses
		WHERE user_id IS NOT NULL
		AND user_id NOT IN (
			SELECT id
			FROM users
			WHERE id IS NOT NULL
		)
		LIMIT 10
	`

	rows, err := cv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check foreign key constraint violations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var businessID, userID interface{}
		if err := rows.Scan(&businessID, &userID); err != nil {
			continue
		}

		violation := ConstraintViolation{
			TableName:      "businesses",
			ConstraintName: "businesses_user_id_fkey",
			ConstraintType: "FOREIGN KEY",
			ColumnName:     "user_id",
			ViolatingValue: userID,
			Description:    fmt.Sprintf("Business %v references non-existent user %v", businessID, userID),
		}
		violations = append(violations, violation)
	}

	return violations, nil
}

// checkCheckConstraintViolations checks for check constraint violations
func (cv *ConstraintValidator) checkCheckConstraintViolations(ctx context.Context) ([]ConstraintViolation, error) {
	var violations []ConstraintViolation

	// Check for invalid user roles
	query := `
		SELECT id, role
		FROM users
		WHERE role IS NOT NULL
		AND role NOT IN ('user', 'admin', 'compliance_officer', 'risk_manager', 'business_analyst', 'developer', 'other')
	`

	rows, err := cv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check check constraint violations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userID, role interface{}
		if err := rows.Scan(&userID, &role); err != nil {
			continue
		}

		violation := ConstraintViolation{
			TableName:      "users",
			ConstraintName: "users_role_check",
			ConstraintType: "CHECK",
			ColumnName:     "role",
			ViolatingValue: role,
			Description:    fmt.Sprintf("User %v has invalid role: %v", userID, role),
		}
		violations = append(violations, violation)
	}

	// Check for invalid risk levels
	query = `
		SELECT id, risk_level
		FROM risk_assessments
		WHERE risk_level IS NOT NULL
		AND risk_level NOT IN ('low', 'medium', 'high', 'critical')
	`

	rows, err = cv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check risk level constraint violations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var assessmentID, riskLevel interface{}
		if err := rows.Scan(&assessmentID, &riskLevel); err != nil {
			continue
		}

		violation := ConstraintViolation{
			TableName:      "risk_assessments",
			ConstraintName: "risk_assessments_risk_level_check",
			ConstraintType: "CHECK",
			ColumnName:     "risk_level",
			ViolatingValue: riskLevel,
			Description:    fmt.Sprintf("Risk assessment %v has invalid risk level: %v", assessmentID, riskLevel),
		}
		violations = append(violations, violation)
	}

	return violations, nil
}

// getExistingConstraints retrieves all existing constraints
func (cv *ConstraintValidator) getExistingConstraints(ctx context.Context) ([]ConstraintInfo, error) {
	var constraints []ConstraintInfo

	query := `
		SELECT 
			tc.table_name,
			tc.constraint_name,
			tc.constraint_type,
			kcu.column_name
		FROM 
			information_schema.table_constraints tc
			LEFT JOIN information_schema.key_column_usage kcu
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
		WHERE 
			tc.table_schema = 'public'
			AND tc.constraint_type IN ('PRIMARY KEY', 'FOREIGN KEY', 'UNIQUE', 'CHECK')
		ORDER BY tc.table_name, tc.constraint_name
	`

	rows, err := cv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing constraints: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var constraint ConstraintInfo
		var columnName *string

		err := rows.Scan(&constraint.TableName, &constraint.ConstraintName, &constraint.ConstraintType, &columnName)
		if err != nil {
			continue
		}

		if columnName != nil {
			constraint.ColumnName = *columnName
		}

		constraints = append(constraints, constraint)
	}

	return constraints, nil
}

// ExpectedConstraint represents an expected constraint
type ExpectedConstraint struct {
	TableName      string
	ConstraintName string
	ConstraintType string
	ColumnName     string
}

// ConstraintInfo represents constraint information
type ConstraintInfo struct {
	TableName      string
	ConstraintName string
	ConstraintType string
	ColumnName     string
}

// MissingConstraint represents a missing constraint
type MissingConstraint struct {
	TableName      string
	ConstraintName string
	ConstraintType string
	ColumnName     string
	Reason         string
}

// ConstraintViolation represents a constraint violation
type ConstraintViolation struct {
	TableName      string
	ConstraintName string
	ConstraintType string
	ColumnName     string
	ViolatingValue interface{}
	Description    string
}
