// Package integrity provides table structure validation
// for the KYB Platform database schema.
package integrity

import (
	"context"
	"fmt"
)

// TableStructureValidator validates table structure and schema consistency
type TableStructureValidator struct {
	validator *Validator
}

// Name returns the name of this validation check
func (tsv *TableStructureValidator) Name() string {
	return "table_structure"
}

// Description returns a description of this validation check
func (tsv *TableStructureValidator) Description() string {
	return "Validates table structure and schema consistency"
}

// Required returns whether this check is required
func (tsv *TableStructureValidator) Required() bool {
	return false
}

// Execute performs table structure validation
func (tsv *TableStructureValidator) Execute(ctx context.Context) (*ValidationResult, error) {
	details := make(map[string]interface{})
	var totalIssues int

	// Define expected tables for KYB platform
	expectedTables := []string{
		"users",
		"api_keys",
		"businesses",
		"business_classifications",
		"risk_assessments",
		"compliance_checks",
		"audit_logs",
		"industries",
		"industry_keywords",
		"classification_codes",
		"industry_patterns",
		"keyword_weights",
		"risk_keywords",
		"business_risk_assessments",
		"industry_code_crosswalks",
		"risk_keyword_relationships",
		"classification_performance_metrics",
	}

	// Check for missing tables
	missingTables, err := tsv.checkMissingTables(ctx, expectedTables)
	if err != nil {
		return tsv.validator.createResult(
			tsv.Name(),
			StatusFailed,
			tsv.validator.formatError(tsv.Name(), "Failed to check missing tables", err),
			details,
		), nil
	}

	if len(missingTables) > 0 {
		details["missing_tables"] = missingTables
		totalIssues += len(missingTables)
	}

	// Check for unexpected tables
	unexpectedTables, err := tsv.checkUnexpectedTables(ctx, expectedTables)
	if err != nil {
		return tsv.validator.createResult(
			tsv.Name(),
			StatusFailed,
			tsv.validator.formatError(tsv.Name(), "Failed to check unexpected tables", err),
			details,
		), nil
	}

	if len(unexpectedTables) > 0 {
		details["unexpected_tables"] = unexpectedTables
		totalIssues += len(unexpectedTables)
	}

	// Check table column structure
	columnIssues, err := tsv.checkTableColumns(ctx)
	if err != nil {
		return tsv.validator.createResult(
			tsv.Name(),
			StatusFailed,
			tsv.validator.formatError(tsv.Name(), "Failed to check table columns", err),
			details,
		), nil
	}

	if len(columnIssues) > 0 {
		details["column_issues"] = columnIssues
		totalIssues += len(columnIssues)
	}

	// Determine status
	status := StatusPassed
	message := "All table structure checks passed"

	if totalIssues > 0 {
		status = StatusWarning
		message = fmt.Sprintf("Found %d table structure issues", totalIssues)
	}

	details["total_issues"] = totalIssues
	details["expected_tables"] = len(expectedTables)

	return tsv.validator.createResult(tsv.Name(), status, message, details), nil
}

// checkMissingTables checks for missing expected tables
func (tsv *TableStructureValidator) checkMissingTables(ctx context.Context, expectedTables []string) ([]string, error) {
	var missingTables []string

	for _, tableName := range expectedTables {
		exists, err := tsv.validator.tableExists(ctx, tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to check if table %s exists: %w", tableName, err)
		}

		if !exists {
			missingTables = append(missingTables, tableName)
		}
	}

	return missingTables, nil
}

// checkUnexpectedTables checks for unexpected tables
func (tsv *TableStructureValidator) checkUnexpectedTables(ctx context.Context, expectedTables []string) ([]string, error) {
	var unexpectedTables []string

	// Get all tables in the public schema
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := tsv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get table list: %w", err)
	}
	defer rows.Close()

	expectedMap := make(map[string]bool)
	for _, table := range expectedTables {
		expectedMap[table] = true
	}

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}

		if !expectedMap[tableName] {
			unexpectedTables = append(unexpectedTables, tableName)
		}
	}

	return unexpectedTables, nil
}

// checkTableColumns checks table column structure
func (tsv *TableStructureValidator) checkTableColumns(ctx context.Context) ([]TableColumnIssue, error) {
	var issues []TableColumnIssue

	// Define expected column structures for key tables
	expectedColumns := map[string][]ExpectedColumn{
		"users": {
			{Name: "id", DataType: "uuid", IsNullable: false, IsPrimaryKey: true},
			{Name: "email", DataType: "character varying", IsNullable: false},
			{Name: "name", DataType: "character varying", IsNullable: true},
			{Name: "created_at", DataType: "timestamp with time zone", IsNullable: false},
			{Name: "updated_at", DataType: "timestamp with time zone", IsNullable: false},
		},
		"businesses": {
			{Name: "id", DataType: "uuid", IsNullable: false, IsPrimaryKey: true},
			{Name: "user_id", DataType: "uuid", IsNullable: false},
			{Name: "name", DataType: "character varying", IsNullable: false},
			{Name: "created_at", DataType: "timestamp with time zone", IsNullable: false},
			{Name: "updated_at", DataType: "timestamp with time zone", IsNullable: false},
		},
		"industries": {
			{Name: "id", DataType: "integer", IsNullable: false, IsPrimaryKey: true},
			{Name: "name", DataType: "character varying", IsNullable: false},
			{Name: "is_active", DataType: "boolean", IsNullable: false},
			{Name: "created_at", DataType: "timestamp with time zone", IsNullable: false},
			{Name: "updated_at", DataType: "timestamp with time zone", IsNullable: false},
		},
	}

	for tableName, expectedCols := range expectedColumns {
		// Check if table exists
		exists, err := tsv.validator.tableExists(ctx, tableName)
		if err != nil || !exists {
			continue
		}

		// Get actual columns
		actualColumns, err := tsv.getTableColumns(ctx, tableName)
		if err != nil {
			continue
		}

		// Check for missing columns
		for _, expectedCol := range expectedCols {
			found := false
			for _, actualCol := range actualColumns {
				if actualCol.Name == expectedCol.Name {
					found = true

					// Check data type
					if actualCol.DataType != expectedCol.DataType {
						issue := TableColumnIssue{
							TableName:     tableName,
							ColumnName:    expectedCol.Name,
							IssueType:     "data_type_mismatch",
							Description:   fmt.Sprintf("Expected data type %s, got %s", expectedCol.DataType, actualCol.DataType),
							ExpectedValue: expectedCol.DataType,
							ActualValue:   actualCol.DataType,
						}
						issues = append(issues, issue)
					}

					// Check nullable constraint
					if actualCol.IsNullable != expectedCol.IsNullable {
						issue := TableColumnIssue{
							TableName:     tableName,
							ColumnName:    expectedCol.Name,
							IssueType:     "nullable_mismatch",
							Description:   fmt.Sprintf("Expected nullable %v, got %v", expectedCol.IsNullable, actualCol.IsNullable),
							ExpectedValue: expectedCol.IsNullable,
							ActualValue:   actualCol.IsNullable,
						}
						issues = append(issues, issue)
					}

					break
				}
			}

			if !found {
				issue := TableColumnIssue{
					TableName:     tableName,
					ColumnName:    expectedCol.Name,
					IssueType:     "missing_column",
					Description:   fmt.Sprintf("Required column %s is missing", expectedCol.Name),
					ExpectedValue: "column exists",
					ActualValue:   "column missing",
				}
				issues = append(issues, issue)
			}
		}
	}

	return issues, nil
}

// getTableColumns retrieves column information for a table
func (tsv *TableStructureValidator) getTableColumns(ctx context.Context, tableName string) ([]ActualColumn, error) {
	query := `
		SELECT 
			column_name,
			data_type,
			is_nullable,
			CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END as is_primary_key
		FROM 
			information_schema.columns c
			LEFT JOIN (
				SELECT ku.table_name, ku.column_name
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage ku ON tc.constraint_name = ku.constraint_name
				WHERE tc.constraint_type = 'PRIMARY KEY'
			) pk ON c.table_name = pk.table_name AND c.column_name = pk.column_name
		WHERE 
			c.table_name = $1
			AND c.table_schema = 'public'
		ORDER BY c.ordinal_position
	`

	rows, err := tsv.validator.executeQuery(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
	}
	defer rows.Close()

	var columns []ActualColumn
	for rows.Next() {
		var column ActualColumn
		var isNullable string
		var isPrimaryKey bool

		err := rows.Scan(&column.Name, &column.DataType, &isNullable, &isPrimaryKey)
		if err != nil {
			continue
		}

		column.IsNullable = isNullable == "YES"
		column.IsPrimaryKey = isPrimaryKey
		columns = append(columns, column)
	}

	return columns, nil
}

// ExpectedColumn represents an expected column structure
type ExpectedColumn struct {
	Name         string
	DataType     string
	IsNullable   bool
	IsPrimaryKey bool
}

// ActualColumn represents an actual column structure
type ActualColumn struct {
	Name         string
	DataType     string
	IsNullable   bool
	IsPrimaryKey bool
}

// TableColumnIssue represents a table column issue
type TableColumnIssue struct {
	TableName     string
	ColumnName    string
	IssueType     string
	Description   string
	ExpectedValue interface{}
	ActualValue   interface{}
}
