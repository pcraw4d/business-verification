// Package integrity provides data type validation
// for the KYB Platform database schema.
package integrity

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// DataTypeValidator validates data types and formats across all tables
type DataTypeValidator struct {
	validator *Validator
}

// Name returns the name of this validation check
func (dtv *DataTypeValidator) Name() string {
	return "data_types"
}

// Description returns a description of this validation check
func (dtv *DataTypeValidator) Description() string {
	return "Validates data types and formats to ensure data consistency"
}

// Required returns whether this check is required
func (dtv *DataTypeValidator) Required() bool {
	return true
}

// Execute performs data type validation
func (dtv *DataTypeValidator) Execute(ctx context.Context) (*ValidationResult, error) {
	details := make(map[string]interface{})
	var totalViolations int

	// Get all tables and their columns
	tables, err := dtv.getTableSchemas(ctx)
	if err != nil {
		return dtv.validator.createResult(
			dtv.Name(),
			StatusFailed,
			dtv.validator.formatError(dtv.Name(), "Failed to retrieve table schemas", err),
			details,
		), nil
	}

	// Validate each table
	for _, table := range tables {
		violations, err := dtv.validateTableDataTypes(ctx, table)
		if err != nil {
			dtv.validator.logger.Printf("Error validating table %s: %v", table.Name, err)
			continue
		}

		if len(violations) > 0 {
			details[table.Name] = map[string]interface{}{
				"violations": len(violations),
				"details":    violations,
			}
			totalViolations += len(violations)
		}
	}

	// Determine status
	status := StatusPassed
	message := "All data types and formats are valid"

	if totalViolations > 0 {
		status = StatusFailed
		message = fmt.Sprintf("Found %d data type violations", totalViolations)
	}

	details["total_tables"] = len(tables)
	details["total_violations"] = totalViolations

	return dtv.validator.createResult(dtv.Name(), status, message, details), nil
}

// TableSchema represents a table schema with its columns
type TableSchema struct {
	Name    string
	Columns []ColumnSchema
}

// ColumnSchema represents a column schema
type ColumnSchema struct {
	Name         string
	DataType     string
	IsNullable   bool
	DefaultValue *string
	MaxLength    *int
	IsPrimaryKey bool
	IsForeignKey bool
}

// DataTypeViolation represents a data type violation
type DataTypeViolation struct {
	TableName      string
	ColumnName     string
	DataType       string
	ViolatingValue interface{}
	RowID          interface{}
	ErrorMessage   string
	ViolationType  string
}

// getTableSchemas retrieves all table schemas from the database
func (dtv *DataTypeValidator) getTableSchemas(ctx context.Context) ([]TableSchema, error) {
	query := `
		SELECT 
			t.table_name,
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			c.character_maximum_length,
			CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END as is_primary_key,
			CASE WHEN fk.column_name IS NOT NULL THEN true ELSE false END as is_foreign_key
		FROM 
			information_schema.tables t
			JOIN information_schema.columns c ON t.table_name = c.table_name
			LEFT JOIN (
				SELECT ku.table_name, ku.column_name
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage ku ON tc.constraint_name = ku.constraint_name
				WHERE tc.constraint_type = 'PRIMARY KEY'
			) pk ON c.table_name = pk.table_name AND c.column_name = pk.column_name
			LEFT JOIN (
				SELECT ku.table_name, ku.column_name
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage ku ON tc.constraint_name = ku.constraint_name
				WHERE tc.constraint_type = 'FOREIGN KEY'
			) fk ON c.table_name = fk.table_name AND c.column_name = fk.column_name
		WHERE 
			t.table_schema = 'public'
			AND t.table_type = 'BASE TABLE'
		ORDER BY t.table_name, c.ordinal_position
	`

	rows, err := dtv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table schemas: %w", err)
	}
	defer rows.Close()

	tableMap := make(map[string]*TableSchema)

	for rows.Next() {
		var tableName, columnName, dataType, isNullable string
		var defaultValue, maxLength sql.NullString
		var isPrimaryKey, isForeignKey bool

		err := rows.Scan(
			&tableName,
			&columnName,
			&dataType,
			&isNullable,
			&defaultValue,
			&maxLength,
			&isPrimaryKey,
			&isForeignKey,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table schema: %w", err)
		}

		// Get or create table schema
		table, exists := tableMap[tableName]
		if !exists {
			table = &TableSchema{Name: tableName, Columns: make([]ColumnSchema, 0)}
			tableMap[tableName] = table
		}

		// Create column schema
		column := ColumnSchema{
			Name:         columnName,
			DataType:     dataType,
			IsNullable:   isNullable == "YES",
			IsPrimaryKey: isPrimaryKey,
			IsForeignKey: isForeignKey,
		}

		if defaultValue.Valid {
			column.DefaultValue = &defaultValue.String
		}

		if maxLength.Valid && maxLength.String != "" {
			if length, err := fmt.Sscanf(maxLength.String, "%d", &column.MaxLength); err == nil && length == 1 {
				// Successfully parsed max length
			}
		}

		table.Columns = append(table.Columns, column)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating table schemas: %w", err)
	}

	// Convert map to slice
	tables := make([]TableSchema, 0, len(tableMap))
	for _, table := range tableMap {
		tables = append(tables, *table)
	}

	return tables, nil
}

// validateTableDataTypes validates data types for a specific table
func (dtv *DataTypeValidator) validateTableDataTypes(ctx context.Context, table TableSchema) ([]DataTypeViolation, error) {
	var violations []DataTypeViolation

	// Get table row count
	rowCount, err := dtv.validator.getTableRowCount(ctx, table.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get row count for table %s: %w", table.Name, err)
	}

	if rowCount == 0 {
		// No data to validate
		return violations, nil
	}

	// Validate each column
	for _, column := range table.Columns {
		columnViolations, err := dtv.validateColumnDataType(ctx, table.Name, column)
		if err != nil {
			dtv.validator.logger.Printf("Error validating column %s.%s: %v", table.Name, column.Name, err)
			continue
		}
		violations = append(violations, columnViolations...)
	}

	return violations, nil
}

// validateColumnDataType validates data type for a specific column
func (dtv *DataTypeValidator) validateColumnDataType(ctx context.Context, tableName string, column ColumnSchema) ([]DataTypeViolation, error) {
	var violations []DataTypeViolation

	switch column.DataType {
	case "uuid":
		violations = append(violations, dtv.validateUUIDFormat(ctx, tableName, column)...)
	case "character varying", "varchar", "text":
		violations = append(violations, dtv.validateStringFormat(ctx, tableName, column)...)
	case "integer", "bigint", "smallint":
		violations = append(violations, dtv.validateIntegerFormat(ctx, tableName, column)...)
	case "numeric", "decimal":
		violations = append(violations, dtv.validateNumericFormat(ctx, tableName, column)...)
	case "boolean":
		violations = append(violations, dtv.validateBooleanFormat(ctx, tableName, column)...)
	case "timestamp with time zone", "timestamp without time zone":
		violations = append(violations, dtv.validateTimestampFormat(ctx, tableName, column)...)
	case "date":
		violations = append(violations, dtv.validateDateFormat(ctx, tableName, column)...)
	case "jsonb", "json":
		violations = append(violations, dtv.validateJSONFormat(ctx, tableName, column)...)
	}

	// Validate specific business rules for KYB platform columns
	violations = append(violations, dtv.validateBusinessRules(ctx, tableName, column)...)

	return violations, nil
}

// validateUUIDFormat validates UUID format
func (dtv *DataTypeValidator) validateUUIDFormat(ctx context.Context, tableName string, column ColumnSchema) []DataTypeViolation {
	var violations []DataTypeViolation

	query := fmt.Sprintf(`
		SELECT id, %s 
		FROM %s 
		WHERE %s IS NOT NULL 
		AND %s::text !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
	`,
		column.Name,
		tableName,
		column.Name,
		column.Name,
	)

	rows, err := dtv.validator.executeQuery(ctx, query)
	if err != nil {
		return violations
	}
	defer rows.Close()

	for rows.Next() {
		var rowID, value interface{}
		if err := rows.Scan(&rowID, &value); err != nil {
			continue
		}

		violation := DataTypeViolation{
			TableName:      tableName,
			ColumnName:     column.Name,
			DataType:       "uuid",
			ViolatingValue: value,
			RowID:          rowID,
			ErrorMessage:   fmt.Sprintf("Invalid UUID format: %v", value),
			ViolationType:  "format",
		}
		violations = append(violations, violation)
	}

	return violations
}

// validateStringFormat validates string format and length
func (dtv *DataTypeValidator) validateStringFormat(ctx context.Context, tableName string, column ColumnSchema) []DataTypeViolation {
	var violations []DataTypeViolation

	// Check for null values in non-nullable columns
	if !column.IsNullable {
		query := fmt.Sprintf(`
			SELECT id, %s 
			FROM %s 
			WHERE %s IS NULL
		`,
			column.Name,
			tableName,
			column.Name,
		)

		rows, err := dtv.validator.executeQuery(ctx, query)
		if err == nil {
			defer rows.Close()

			for rows.Next() {
				var rowID, value interface{}
				if err := rows.Scan(&rowID, &value); err != nil {
					continue
				}

				violation := DataTypeViolation{
					TableName:      tableName,
					ColumnName:     column.Name,
					DataType:       "string",
					ViolatingValue: value,
					RowID:          rowID,
					ErrorMessage:   "Non-nullable string column contains NULL value",
					ViolationType:  "null_constraint",
				}
				violations = append(violations, violation)
			}
		}
	}

	// Check string length if max length is defined
	if column.MaxLength != nil {
		query := fmt.Sprintf(`
			SELECT id, %s 
			FROM %s 
			WHERE %s IS NOT NULL 
			AND LENGTH(%s::text) > %d
		`,
			column.Name,
			tableName,
			column.Name,
			column.Name,
			*column.MaxLength,
		)

		rows, err := dtv.validator.executeQuery(ctx, query)
		if err == nil {
			defer rows.Close()

			for rows.Next() {
				var rowID, value interface{}
				if err := rows.Scan(&rowID, &value); err != nil {
					continue
				}

				violation := DataTypeViolation{
					TableName:      tableName,
					ColumnName:     column.Name,
					DataType:       "string",
					ViolatingValue: value,
					RowID:          rowID,
					ErrorMessage:   fmt.Sprintf("String length %d exceeds maximum length %d", len(value.(string)), *column.MaxLength),
					ViolationType:  "length_constraint",
				}
				violations = append(violations, violation)
			}
		}
	}

	return violations
}

// validateIntegerFormat validates integer format
func (dtv *DataTypeValidator) validateIntegerFormat(ctx context.Context, tableName string, column ColumnSchema) []DataTypeViolation {
	var violations []DataTypeViolation

	// Check for null values in non-nullable columns
	if !column.IsNullable {
		query := fmt.Sprintf(`
			SELECT id, %s 
			FROM %s 
			WHERE %s IS NULL
		`,
			column.Name,
			tableName,
			column.Name,
		)

		rows, err := dtv.validator.executeQuery(ctx, query)
		if err == nil {
			defer rows.Close()

			for rows.Next() {
				var rowID, value interface{}
				if err := rows.Scan(&rowID, &value); err != nil {
					continue
				}

				violation := DataTypeViolation{
					TableName:      tableName,
					ColumnName:     column.Name,
					DataType:       "integer",
					ViolatingValue: value,
					RowID:          rowID,
					ErrorMessage:   "Non-nullable integer column contains NULL value",
					ViolationType:  "null_constraint",
				}
				violations = append(violations, violation)
			}
		}
	}

	return violations
}

// validateNumericFormat validates numeric format
func (dtv *DataTypeValidator) validateNumericFormat(ctx context.Context, tableName string, column ColumnSchema) []DataTypeViolation {
	var violations []DataTypeViolation

	// Check for null values in non-nullable columns
	if !column.IsNullable {
		query := fmt.Sprintf(`
			SELECT id, %s 
			FROM %s 
			WHERE %s IS NULL
		`,
			column.Name,
			tableName,
			column.Name,
		)

		rows, err := dtv.validator.executeQuery(ctx, query)
		if err == nil {
			defer rows.Close()

			for rows.Next() {
				var rowID, value interface{}
				if err := rows.Scan(&rowID, &value); err != nil {
					continue
				}

				violation := DataTypeViolation{
					TableName:      tableName,
					ColumnName:     column.Name,
					DataType:       "numeric",
					ViolatingValue: value,
					RowID:          rowID,
					ErrorMessage:   "Non-nullable numeric column contains NULL value",
					ViolationType:  "null_constraint",
				}
				violations = append(violations, violation)
			}
		}
	}

	return violations
}

// validateBooleanFormat validates boolean format
func (dtv *DataTypeValidator) validateBooleanFormat(ctx context.Context, tableName string, column ColumnSchema) []DataTypeViolation {
	var violations []DataTypeViolation

	// Check for null values in non-nullable columns
	if !column.IsNullable {
		query := fmt.Sprintf(`
			SELECT id, %s 
			FROM %s 
			WHERE %s IS NULL
		`,
			column.Name,
			tableName,
			column.Name,
		)

		rows, err := dtv.validator.executeQuery(ctx, query)
		if err == nil {
			defer rows.Close()

			for rows.Next() {
				var rowID, value interface{}
				if err := rows.Scan(&rowID, &value); err != nil {
					continue
				}

				violation := DataTypeViolation{
					TableName:      tableName,
					ColumnName:     column.Name,
					DataType:       "boolean",
					ViolatingValue: value,
					RowID:          rowID,
					ErrorMessage:   "Non-nullable boolean column contains NULL value",
					ViolationType:  "null_constraint",
				}
				violations = append(violations, violation)
			}
		}
	}

	return violations
}

// validateTimestampFormat validates timestamp format
func (dtv *DataTypeValidator) validateTimestampFormat(ctx context.Context, tableName string, column ColumnSchema) []DataTypeViolation {
	var violations []DataTypeViolation

	// Check for null values in non-nullable columns
	if !column.IsNullable {
		query := fmt.Sprintf(`
			SELECT id, %s 
			FROM %s 
			WHERE %s IS NULL
		`,
			column.Name,
			tableName,
			column.Name,
		)

		rows, err := dtv.validator.executeQuery(ctx, query)
		if err == nil {
			defer rows.Close()

			for rows.Next() {
				var rowID, value interface{}
				if err := rows.Scan(&rowID, &value); err != nil {
					continue
				}

				violation := DataTypeViolation{
					TableName:      tableName,
					ColumnName:     column.Name,
					DataType:       "timestamp",
					ViolatingValue: value,
					RowID:          rowID,
					ErrorMessage:   "Non-nullable timestamp column contains NULL value",
					ViolationType:  "null_constraint",
				}
				violations = append(violations, violation)
			}
		}
	}

	return violations
}

// validateDateFormat validates date format
func (dtv *DataTypeValidator) validateDateFormat(ctx context.Context, tableName string, column ColumnSchema) []DataTypeViolation {
	var violations []DataTypeViolation

	// Check for null values in non-nullable columns
	if !column.IsNullable {
		query := fmt.Sprintf(`
			SELECT id, %s 
			FROM %s 
			WHERE %s IS NULL
		`,
			column.Name,
			tableName,
			column.Name,
		)

		rows, err := dtv.validator.executeQuery(ctx, query)
		if err == nil {
			defer rows.Close()

			for rows.Next() {
				var rowID, value interface{}
				if err := rows.Scan(&rowID, &value); err != nil {
					continue
				}

				violation := DataTypeViolation{
					TableName:      tableName,
					ColumnName:     column.Name,
					DataType:       "date",
					ViolatingValue: value,
					RowID:          rowID,
					ErrorMessage:   "Non-nullable date column contains NULL value",
					ViolationType:  "null_constraint",
				}
				violations = append(violations, violation)
			}
		}
	}

	return violations
}

// validateJSONFormat validates JSON format
func (dtv *DataTypeValidator) validateJSONFormat(ctx context.Context, tableName string, column ColumnSchema) []DataTypeViolation {
	var violations []DataTypeViolation

	// Check for null values in non-nullable columns
	if !column.IsNullable {
		query := fmt.Sprintf(`
			SELECT id, %s 
			FROM %s 
			WHERE %s IS NULL
		`,
			column.Name,
			tableName,
			column.Name,
		)

		rows, err := dtv.validator.executeQuery(ctx, query)
		if err == nil {
			defer rows.Close()

			for rows.Next() {
				var rowID, value interface{}
				if err := rows.Scan(&rowID, &value); err != nil {
					continue
				}

				violation := DataTypeViolation{
					TableName:      tableName,
					ColumnName:     column.Name,
					DataType:       "json",
					ViolatingValue: value,
					RowID:          rowID,
					ErrorMessage:   "Non-nullable JSON column contains NULL value",
					ViolationType:  "null_constraint",
				}
				violations = append(violations, violation)
			}
		}
	}

	return violations
}

// validateBusinessRules validates specific business rules for KYB platform columns
func (dtv *DataTypeValidator) validateBusinessRules(ctx context.Context, tableName string, column ColumnSchema) []DataTypeViolation {
	var violations []DataTypeViolation

	// Email validation
	if strings.Contains(column.Name, "email") && column.DataType == "character varying" {
		violations = append(violations, dtv.validateEmailFormat(ctx, tableName, column)...)
	}

	// URL validation
	if strings.Contains(column.Name, "url") && column.DataType == "character varying" {
		violations = append(violations, dtv.validateURLFormat(ctx, tableName, column)...)
	}

	// Phone number validation
	if strings.Contains(column.Name, "phone") && column.DataType == "character varying" {
		violations = append(violations, dtv.validatePhoneFormat(ctx, tableName, column)...)
	}

	// Confidence score validation (should be between 0 and 1)
	if strings.Contains(column.Name, "confidence") && column.DataType == "numeric" {
		violations = append(violations, dtv.validateConfidenceScore(ctx, tableName, column)...)
	}

	// Risk score validation (should be between 0 and 1)
	if strings.Contains(column.Name, "risk_score") && column.DataType == "numeric" {
		violations = append(violations, dtv.validateRiskScore(ctx, tableName, column)...)
	}

	return violations
}

// validateEmailFormat validates email format
func (dtv *DataTypeValidator) validateEmailFormat(ctx context.Context, tableName string, column ColumnSchema) []DataTypeViolation {
	var violations []DataTypeViolation

	query := fmt.Sprintf(`
		SELECT id, %s 
		FROM %s 
		WHERE %s IS NOT NULL 
		AND %s::text !~ '^[a-zA-Z0-9._%%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$'
	`,
		column.Name,
		tableName,
		column.Name,
		column.Name,
	)

	rows, err := dtv.validator.executeQuery(ctx, query)
	if err != nil {
		return violations
	}
	defer rows.Close()

	for rows.Next() {
		var rowID, value interface{}
		if err := rows.Scan(&rowID, &value); err != nil {
			continue
		}

		violation := DataTypeViolation{
			TableName:      tableName,
			ColumnName:     column.Name,
			DataType:       "email",
			ViolatingValue: value,
			RowID:          rowID,
			ErrorMessage:   fmt.Sprintf("Invalid email format: %v", value),
			ViolationType:  "format",
		}
		violations = append(violations, violation)
	}

	return violations
}

// validateURLFormat validates URL format
func (dtv *DataTypeValidator) validateURLFormat(ctx context.Context, tableName string, column ColumnSchema) []DataTypeViolation {
	var violations []DataTypeViolation

	query := fmt.Sprintf(`
		SELECT id, %s 
		FROM %s 
		WHERE %s IS NOT NULL 
		AND %s::text !~ '^https?://[^\\s/$.?#].[^\\s]*$'
	`,
		column.Name,
		tableName,
		column.Name,
		column.Name,
	)

	rows, err := dtv.validator.executeQuery(ctx, query)
	if err != nil {
		return violations
	}
	defer rows.Close()

	for rows.Next() {
		var rowID, value interface{}
		if err := rows.Scan(&rowID, &value); err != nil {
			continue
		}

		violation := DataTypeViolation{
			TableName:      tableName,
			ColumnName:     column.Name,
			DataType:       "url",
			ViolatingValue: value,
			RowID:          rowID,
			ErrorMessage:   fmt.Sprintf("Invalid URL format: %v", value),
			ViolationType:  "format",
		}
		violations = append(violations, violation)
	}

	return violations
}

// validatePhoneFormat validates phone number format
func (dtv *DataTypeValidator) validatePhoneFormat(ctx context.Context, tableName string, column ColumnSchema) []DataTypeViolation {
	var violations []DataTypeViolation

	query := fmt.Sprintf(`
		SELECT id, %s 
		FROM %s 
		WHERE %s IS NOT NULL 
		AND %s::text !~ '^\\+?[1-9]\\d{1,14}$'
	`,
		column.Name,
		tableName,
		column.Name,
		column.Name,
	)

	rows, err := dtv.validator.executeQuery(ctx, query)
	if err != nil {
		return violations
	}
	defer rows.Close()

	for rows.Next() {
		var rowID, value interface{}
		if err := rows.Scan(&rowID, &value); err != nil {
			continue
		}

		violation := DataTypeViolation{
			TableName:      tableName,
			ColumnName:     column.Name,
			DataType:       "phone",
			ViolatingValue: value,
			RowID:          rowID,
			ErrorMessage:   fmt.Sprintf("Invalid phone format: %v", value),
			ViolationType:  "format",
		}
		violations = append(violations, violation)
	}

	return violations
}

// validateConfidenceScore validates confidence score range (0-1)
func (dtv *DataTypeValidator) validateConfidenceScore(ctx context.Context, tableName string, column ColumnSchema) []DataTypeViolation {
	var violations []DataTypeViolation

	query := fmt.Sprintf(`
		SELECT id, %s 
		FROM %s 
		WHERE %s IS NOT NULL 
		AND (%s < 0 OR %s > 1)
	`,
		column.Name,
		tableName,
		column.Name,
		column.Name,
		column.Name,
	)

	rows, err := dtv.validator.executeQuery(ctx, query)
	if err != nil {
		return violations
	}
	defer rows.Close()

	for rows.Next() {
		var rowID, value interface{}
		if err := rows.Scan(&rowID, &value); err != nil {
			continue
		}

		violation := DataTypeViolation{
			TableName:      tableName,
			ColumnName:     column.Name,
			DataType:       "confidence_score",
			ViolatingValue: value,
			RowID:          rowID,
			ErrorMessage:   fmt.Sprintf("Confidence score %v is outside valid range [0, 1]", value),
			ViolationType:  "range",
		}
		violations = append(violations, violation)
	}

	return violations
}

// validateRiskScore validates risk score range (0-1)
func (dtv *DataTypeValidator) validateRiskScore(ctx context.Context, tableName string, column ColumnSchema) []DataTypeViolation {
	var violations []DataTypeViolation

	query := fmt.Sprintf(`
		SELECT id, %s 
		FROM %s 
		WHERE %s IS NOT NULL 
		AND (%s < 0 OR %s > 1)
	`,
		column.Name,
		tableName,
		column.Name,
		column.Name,
		column.Name,
	)

	rows, err := dtv.validator.executeQuery(ctx, query)
	if err != nil {
		return violations
	}
	defer rows.Close()

	for rows.Next() {
		var rowID, value interface{}
		if err := rows.Scan(&rowID, &value); err != nil {
			continue
		}

		violation := DataTypeViolation{
			TableName:      tableName,
			ColumnName:     column.Name,
			DataType:       "risk_score",
			ViolatingValue: value,
			RowID:          rowID,
			ErrorMessage:   fmt.Sprintf("Risk score %v is outside valid range [0, 1]", value),
			ViolationType:  "range",
		}
		violations = append(violations, violation)
	}

	return violations
}
