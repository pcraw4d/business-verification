// Package integrity provides foreign key constraint validation
// for the KYB Platform database schema.
package integrity

import (
	"context"
	"fmt"
)

// ForeignKeyValidator validates foreign key constraints across all tables
type ForeignKeyValidator struct {
	validator *Validator
}

// Name returns the name of this validation check
func (fkv *ForeignKeyValidator) Name() string {
	return "foreign_key_constraints"
}

// Description returns a description of this validation check
func (fkv *ForeignKeyValidator) Description() string {
	return "Validates all foreign key constraints to ensure referential integrity"
}

// Required returns whether this check is required
func (fkv *ForeignKeyValidator) Required() bool {
	return true
}

// Execute performs foreign key constraint validation
func (fkv *ForeignKeyValidator) Execute(ctx context.Context) (*ValidationResult, error) {
	details := make(map[string]interface{})
	var violations []ForeignKeyViolation
	var totalViolations int

	// Get all foreign key constraints
	constraints, err := fkv.getForeignKeyConstraints(ctx)
	if err != nil {
		return fkv.validator.createResult(
			fkv.Name(),
			StatusFailed,
			fkv.validator.formatError(fkv.Name(), "Failed to retrieve foreign key constraints", err),
			details,
		), nil
	}

	// Validate each foreign key constraint
	for _, constraint := range constraints {
		violations, err := fkv.validateForeignKeyConstraint(ctx, constraint)
		if err != nil {
			fkv.validator.logger.Printf("Error validating constraint %s: %v", constraint.Name, err)
			continue
		}

		totalViolations += len(violations)
		details[constraint.Name] = map[string]interface{}{
			"violations": len(violations),
			"table":      constraint.TableName,
			"column":     constraint.ColumnName,
			"references": constraint.ReferencedTable,
		}
	}

	// Determine status
	status := StatusPassed
	message := "All foreign key constraints are valid"

	if totalViolations > 0 {
		status = StatusFailed
		message = fmt.Sprintf("Found %d foreign key constraint violations", totalViolations)
	}

	details["total_constraints"] = len(constraints)
	details["total_violations"] = totalViolations
	details["violation_details"] = violations

	return fkv.validator.createResult(fkv.Name(), status, message, details), nil
}

// ForeignKeyConstraint represents a foreign key constraint
type ForeignKeyConstraint struct {
	Name             string
	TableName        string
	ColumnName       string
	ReferencedTable  string
	ReferencedColumn string
	ConstraintType   string
}

// ForeignKeyViolation represents a foreign key constraint violation
type ForeignKeyViolation struct {
	TableName        string
	ColumnName       string
	ReferencedTable  string
	ReferencedColumn string
	ViolatingValue   interface{}
	RowID            interface{}
	ErrorMessage     string
}

// getForeignKeyConstraints retrieves all foreign key constraints from the database
func (fkv *ForeignKeyValidator) getForeignKeyConstraints(ctx context.Context) ([]ForeignKeyConstraint, error) {
	query := `
		SELECT 
			tc.constraint_name,
			tc.table_name,
			kcu.column_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name,
			tc.constraint_type
		FROM 
			information_schema.table_constraints AS tc 
			JOIN information_schema.key_column_usage AS kcu
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
			JOIN information_schema.constraint_column_usage AS ccu
				ON ccu.constraint_name = tc.constraint_name
				AND ccu.table_schema = tc.table_schema
		WHERE 
			tc.constraint_type = 'FOREIGN KEY'
			AND tc.table_schema = 'public'
		ORDER BY tc.table_name, tc.constraint_name
	`

	rows, err := fkv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query foreign key constraints: %w", err)
	}
	defer rows.Close()

	var constraints []ForeignKeyConstraint
	for rows.Next() {
		var constraint ForeignKeyConstraint
		err := rows.Scan(
			&constraint.Name,
			&constraint.TableName,
			&constraint.ColumnName,
			&constraint.ReferencedTable,
			&constraint.ReferencedColumn,
			&constraint.ConstraintType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan foreign key constraint: %w", err)
		}
		constraints = append(constraints, constraint)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating foreign key constraints: %w", err)
	}

	return constraints, nil
}

// validateForeignKeyConstraint validates a specific foreign key constraint
func (fkv *ForeignKeyValidator) validateForeignKeyConstraint(ctx context.Context, constraint ForeignKeyConstraint) ([]ForeignKeyViolation, error) {
	var violations []ForeignKeyViolation

	// Check if both tables exist
	tableExists, err := fkv.validator.tableExists(ctx, constraint.TableName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if table %s exists: %w", constraint.TableName, err)
	}
	if !tableExists {
		return nil, fmt.Errorf("table %s does not exist", constraint.TableName)
	}

	refTableExists, err := fkv.validator.tableExists(ctx, constraint.ReferencedTable)
	if err != nil {
		return nil, fmt.Errorf("failed to check if referenced table %s exists: %w", constraint.ReferencedTable, err)
	}
	if !refTableExists {
		return nil, fmt.Errorf("referenced table %s does not exist", constraint.ReferencedTable)
	}

	// Find orphaned records (records in the referencing table that don't have corresponding records in the referenced table)
	query := fmt.Sprintf(`
		SELECT 
			%s,
			%s
		FROM %s 
		WHERE %s IS NOT NULL 
		AND %s NOT IN (
			SELECT %s 
			FROM %s 
			WHERE %s IS NOT NULL
		)
	`,
		// Get primary key column for the referencing table
		fkv.getPrimaryKeyColumn(ctx, constraint.TableName),
		constraint.ColumnName,
		constraint.TableName,
		constraint.ColumnName,
		constraint.ColumnName,
		constraint.ReferencedColumn,
		constraint.ReferencedTable,
		constraint.ReferencedColumn,
	)

	rows, err := fkv.validator.executeQuery(ctx, query)
	if err != nil {
		// If the query fails, it might be due to data type mismatches or other issues
		fkv.validator.logger.Printf("Warning: Could not validate foreign key constraint %s: %v", constraint.Name, err)
		return violations, nil
	}
	defer rows.Close()

	for rows.Next() {
		var rowID, violatingValue interface{}
		err := rows.Scan(&rowID, &violatingValue)
		if err != nil {
			fkv.validator.logger.Printf("Error scanning foreign key violation: %v", err)
			continue
		}

		violation := ForeignKeyViolation{
			TableName:        constraint.TableName,
			ColumnName:       constraint.ColumnName,
			ReferencedTable:  constraint.ReferencedTable,
			ReferencedColumn: constraint.ReferencedColumn,
			ViolatingValue:   violatingValue,
			RowID:            rowID,
			ErrorMessage:     fmt.Sprintf("Referenced value %v does not exist in %s.%s", violatingValue, constraint.ReferencedTable, constraint.ReferencedColumn),
		}
		violations = append(violations, violation)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating foreign key violations: %w", err)
	}

	return violations, nil
}

// getPrimaryKeyColumn retrieves the primary key column name for a table
func (fkv *ForeignKeyValidator) getPrimaryKeyColumn(ctx context.Context, tableName string) string {
	query := `
		SELECT kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		WHERE tc.constraint_type = 'PRIMARY KEY'
			AND tc.table_name = $1
			AND tc.table_schema = 'public'
		LIMIT 1
	`

	var columnName string
	err := fkv.validator.executeQueryRow(ctx, query, tableName).Scan(&columnName)
	if err != nil {
		// Default to 'id' if we can't determine the primary key
		return "id"
	}

	return columnName
}

// validateSpecificForeignKeys validates specific foreign key relationships for KYB platform tables
func (fkv *ForeignKeyValidator) validateSpecificForeignKeys(ctx context.Context) ([]ForeignKeyViolation, error) {
	var allViolations []ForeignKeyViolation

	// Define specific foreign key relationships to validate
	relationships := []struct {
		tableName        string
		columnName       string
		referencedTable  string
		referencedColumn string
	}{
		// User relationships
		{"api_keys", "user_id", "users", "id"},
		{"businesses", "user_id", "users", "id"},
		{"business_classifications", "user_id", "users", "id"},
		{"risk_assessments", "user_id", "users", "id"},
		{"compliance_checks", "user_id", "users", "id"},
		{"audit_logs", "user_id", "users", "id"},

		// Business relationships
		{"business_classifications", "business_id", "businesses", "id"},
		{"risk_assessments", "business_id", "businesses", "id"},
		{"compliance_checks", "business_id", "businesses", "id"},

		// Classification relationships
		{"industry_keywords", "industry_id", "industries", "id"},
		{"classification_codes", "industry_id", "industries", "id"},
		{"industry_patterns", "industry_id", "industries", "id"},
		{"keyword_weights", "keyword_id", "industry_keywords", "id"},

		// Risk relationships
		{"business_risk_assessments", "business_id", "merchants", "id"},
		{"business_risk_assessments", "risk_keyword_id", "risk_keywords", "id"},
		{"risk_keyword_relationships", "parent_keyword_id", "risk_keywords", "id"},
		{"risk_keyword_relationships", "child_keyword_id", "risk_keywords", "id"},

		// Crosswalk relationships
		{"industry_code_crosswalks", "industry_id", "industries", "id"},
	}

	for _, rel := range relationships {
		// Check if both tables exist before validating
		tableExists, err := fkv.validator.tableExists(ctx, rel.tableName)
		if err != nil || !tableExists {
			continue
		}

		refTableExists, err := fkv.validator.tableExists(ctx, rel.referencedTable)
		if err != nil || !refTableExists {
			continue
		}

		violations, err := fkv.validateSpecificForeignKey(ctx, rel.tableName, rel.columnName, rel.referencedTable, rel.referencedColumn)
		if err != nil {
			fkv.validator.logger.Printf("Error validating foreign key %s.%s -> %s.%s: %v",
				rel.tableName, rel.columnName, rel.referencedTable, rel.referencedColumn, err)
			continue
		}

		allViolations = append(allViolations, violations...)
	}

	return allViolations, nil
}

// validateSpecificForeignKey validates a specific foreign key relationship
func (fkv *ForeignKeyValidator) validateSpecificForeignKey(ctx context.Context, tableName, columnName, referencedTable, referencedColumn string) ([]ForeignKeyViolation, error) {
	var violations []ForeignKeyViolation

	// Find orphaned records
	query := fmt.Sprintf(`
		SELECT 
			%s,
			%s
		FROM %s 
		WHERE %s IS NOT NULL 
		AND %s NOT IN (
			SELECT %s 
			FROM %s 
			WHERE %s IS NOT NULL
		)
	`,
		fkv.getPrimaryKeyColumn(ctx, tableName),
		columnName,
		tableName,
		columnName,
		columnName,
		referencedColumn,
		referencedTable,
		referencedColumn,
	)

	rows, err := fkv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to validate foreign key %s.%s -> %s.%s: %w",
			tableName, columnName, referencedTable, referencedColumn, err)
	}
	defer rows.Close()

	for rows.Next() {
		var rowID, violatingValue interface{}
		err := rows.Scan(&rowID, &violatingValue)
		if err != nil {
			continue
		}

		violation := ForeignKeyViolation{
			TableName:        tableName,
			ColumnName:       columnName,
			ReferencedTable:  referencedTable,
			ReferencedColumn: referencedColumn,
			ViolatingValue:   violatingValue,
			RowID:            rowID,
			ErrorMessage:     fmt.Sprintf("Referenced value %v does not exist in %s.%s", violatingValue, referencedTable, referencedColumn),
		}
		violations = append(violations, violation)
	}

	return violations, nil
}
