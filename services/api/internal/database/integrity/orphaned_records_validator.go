// Package integrity provides orphaned records validation
// for the KYB Platform database schema.
package integrity

import (
	"context"
	"fmt"
)

// OrphanedRecordsValidator validates orphaned records across all tables
type OrphanedRecordsValidator struct {
	validator *Validator
}

// Name returns the name of this validation check
func (orv *OrphanedRecordsValidator) Name() string {
	return "orphaned_records"
}

// Description returns a description of this validation check
func (orv *OrphanedRecordsValidator) Description() string {
	return "Checks for orphaned records that violate referential integrity"
}

// Required returns whether this check is required
func (orv *OrphanedRecordsValidator) Required() bool {
	return true
}

// Execute performs orphaned records validation
func (orv *OrphanedRecordsValidator) Execute(ctx context.Context) (*ValidationResult, error) {
	details := make(map[string]interface{})
	var totalOrphanedRecords int

	// Define specific relationships to check for orphaned records
	relationships := []OrphanedRecordCheck{
		// User relationships
		{
			TableName:        "api_keys",
			ColumnName:       "user_id",
			ReferencedTable:  "users",
			ReferencedColumn: "id",
			Description:      "API keys referencing non-existent users",
		},
		{
			TableName:        "businesses",
			ColumnName:       "user_id",
			ReferencedTable:  "users",
			ReferencedColumn: "id",
			Description:      "Businesses referencing non-existent users",
		},
		{
			TableName:        "business_classifications",
			ColumnName:       "user_id",
			ReferencedTable:  "users",
			ReferencedColumn: "id",
			Description:      "Business classifications referencing non-existent users",
		},
		{
			TableName:        "risk_assessments",
			ColumnName:       "user_id",
			ReferencedTable:  "users",
			ReferencedColumn: "id",
			Description:      "Risk assessments referencing non-existent users",
		},
		{
			TableName:        "compliance_checks",
			ColumnName:       "user_id",
			ReferencedTable:  "users",
			ReferencedColumn: "id",
			Description:      "Compliance checks referencing non-existent users",
		},
		{
			TableName:        "audit_logs",
			ColumnName:       "user_id",
			ReferencedTable:  "users",
			ReferencedColumn: "id",
			Description:      "Audit logs referencing non-existent users",
		},

		// Business relationships
		{
			TableName:        "business_classifications",
			ColumnName:       "business_id",
			ReferencedTable:  "businesses",
			ReferencedColumn: "id",
			Description:      "Business classifications referencing non-existent businesses",
		},
		{
			TableName:        "risk_assessments",
			ColumnName:       "business_id",
			ReferencedTable:  "businesses",
			ReferencedColumn: "id",
			Description:      "Risk assessments referencing non-existent businesses",
		},
		{
			TableName:        "compliance_checks",
			ColumnName:       "business_id",
			ReferencedTable:  "businesses",
			ReferencedColumn: "id",
			Description:      "Compliance checks referencing non-existent businesses",
		},

		// Classification relationships
		{
			TableName:        "industry_keywords",
			ColumnName:       "industry_id",
			ReferencedTable:  "industries",
			ReferencedColumn: "id",
			Description:      "Industry keywords referencing non-existent industries",
		},
		{
			TableName:        "classification_codes",
			ColumnName:       "industry_id",
			ReferencedTable:  "industries",
			ReferencedColumn: "id",
			Description:      "Classification codes referencing non-existent industries",
		},
		{
			TableName:        "industry_patterns",
			ColumnName:       "industry_id",
			ReferencedTable:  "industries",
			ReferencedColumn: "id",
			Description:      "Industry patterns referencing non-existent industries",
		},
		{
			TableName:        "keyword_weights",
			ColumnName:       "keyword_id",
			ReferencedTable:  "industry_keywords",
			ReferencedColumn: "id",
			Description:      "Keyword weights referencing non-existent industry keywords",
		},

		// Risk relationships
		{
			TableName:        "business_risk_assessments",
			ColumnName:       "business_id",
			ReferencedTable:  "merchants",
			ReferencedColumn: "id",
			Description:      "Business risk assessments referencing non-existent merchants",
		},
		{
			TableName:        "business_risk_assessments",
			ColumnName:       "risk_keyword_id",
			ReferencedTable:  "risk_keywords",
			ReferencedColumn: "id",
			Description:      "Business risk assessments referencing non-existent risk keywords",
		},
		{
			TableName:        "risk_keyword_relationships",
			ColumnName:       "parent_keyword_id",
			ReferencedTable:  "risk_keywords",
			ReferencedColumn: "id",
			Description:      "Risk keyword relationships referencing non-existent parent keywords",
		},
		{
			TableName:        "risk_keyword_relationships",
			ColumnName:       "child_keyword_id",
			ReferencedTable:  "risk_keywords",
			ReferencedColumn: "id",
			Description:      "Risk keyword relationships referencing non-existent child keywords",
		},

		// Crosswalk relationships
		{
			TableName:        "industry_code_crosswalks",
			ColumnName:       "industry_id",
			ReferencedTable:  "industries",
			ReferencedColumn: "id",
			Description:      "Industry code crosswalks referencing non-existent industries",
		},

		// Performance metrics relationships
		{
			TableName:        "classification_performance_metrics",
			ColumnName:       "industry_id",
			ReferencedTable:  "industries",
			ReferencedColumn: "id",
			Description:      "Classification performance metrics referencing non-existent industries",
		},
	}

	// Check each relationship
	for _, relationship := range relationships {
		orphanedCount, err := orv.checkOrphanedRecords(ctx, relationship)
		if err != nil {
			orv.validator.logger.Printf("Error checking orphaned records for %s.%s -> %s.%s: %v",
				relationship.TableName, relationship.ColumnName,
				relationship.ReferencedTable, relationship.ReferencedColumn, err)
			continue
		}

		if orphanedCount > 0 {
			details[fmt.Sprintf("%s.%s", relationship.TableName, relationship.ColumnName)] = map[string]interface{}{
				"orphaned_count":   orphanedCount,
				"referenced_table": relationship.ReferencedTable,
				"description":      relationship.Description,
			}
			totalOrphanedRecords += orphanedCount
		}
	}

	// Determine status
	status := StatusPassed
	message := "No orphaned records found"

	if totalOrphanedRecords > 0 {
		status = StatusFailed
		message = fmt.Sprintf("Found %d orphaned records", totalOrphanedRecords)
	}

	details["total_orphaned_records"] = totalOrphanedRecords
	details["total_relationships_checked"] = len(relationships)

	return orv.validator.createResult(orv.Name(), status, message, details), nil
}

// OrphanedRecordCheck represents a check for orphaned records
type OrphanedRecordCheck struct {
	TableName        string
	ColumnName       string
	ReferencedTable  string
	ReferencedColumn string
	Description      string
}

// checkOrphanedRecords checks for orphaned records in a specific relationship
func (orv *OrphanedRecordsValidator) checkOrphanedRecords(ctx context.Context, check OrphanedRecordCheck) (int, error) {
	// Check if both tables exist
	tableExists, err := orv.validator.tableExists(ctx, check.TableName)
	if err != nil {
		return 0, fmt.Errorf("failed to check if table %s exists: %w", check.TableName, err)
	}
	if !tableExists {
		// Table doesn't exist, so no orphaned records
		return 0, nil
	}

	refTableExists, err := orv.validator.tableExists(ctx, check.ReferencedTable)
	if err != nil {
		return 0, fmt.Errorf("failed to check if referenced table %s exists: %w", check.ReferencedTable, err)
	}
	if !refTableExists {
		// Referenced table doesn't exist, so all records are orphaned
		// Get count of non-null values in the referencing column
		count, err := orv.validator.getTableRowCount(ctx, check.TableName)
		if err != nil {
			return 0, fmt.Errorf("failed to get row count for table %s: %w", check.TableName, err)
		}
		return int(count), nil
	}

	// Count orphaned records
	query := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM %s 
		WHERE %s IS NOT NULL 
		AND %s NOT IN (
			SELECT %s 
			FROM %s 
			WHERE %s IS NOT NULL
		)
	`,
		check.TableName,
		check.ColumnName,
		check.ColumnName,
		check.ReferencedColumn,
		check.ReferencedTable,
		check.ReferencedColumn,
	)

	var count int
	err = orv.validator.executeQueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count orphaned records: %w", err)
	}

	return count, nil
}

// getOrphanedRecordDetails retrieves detailed information about orphaned records
func (orv *OrphanedRecordsValidator) getOrphanedRecordDetails(ctx context.Context, check OrphanedRecordCheck, limit int) ([]OrphanedRecordDetail, error) {
	var details []OrphanedRecordDetail

	// Get primary key column for the referencing table
	pkColumn := orv.getPrimaryKeyColumn(ctx, check.TableName)

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
		LIMIT %d
	`,
		pkColumn,
		check.ColumnName,
		check.TableName,
		check.ColumnName,
		check.ColumnName,
		check.ReferencedColumn,
		check.ReferencedTable,
		check.ReferencedColumn,
		limit,
	)

	rows, err := orv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get orphaned record details: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var detail OrphanedRecordDetail
		err := rows.Scan(&detail.RowID, &detail.OrphanedValue)
		if err != nil {
			continue
		}

		detail.TableName = check.TableName
		detail.ColumnName = check.ColumnName
		detail.ReferencedTable = check.ReferencedTable
		detail.ReferencedColumn = check.ReferencedColumn
		detail.Description = check.Description

		details = append(details, detail)
	}

	return details, nil
}

// OrphanedRecordDetail represents detailed information about an orphaned record
type OrphanedRecordDetail struct {
	TableName        string
	ColumnName       string
	ReferencedTable  string
	ReferencedColumn string
	RowID            interface{}
	OrphanedValue    interface{}
	Description      string
}

// getPrimaryKeyColumn retrieves the primary key column name for a table
func (orv *OrphanedRecordsValidator) getPrimaryKeyColumn(ctx context.Context, tableName string) string {
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
	err := orv.validator.executeQueryRow(ctx, query, tableName).Scan(&columnName)
	if err != nil {
		// Default to 'id' if we can't determine the primary key
		return "id"
	}

	return columnName
}

// validateSpecificOrphanedRecords validates specific orphaned record scenarios
func (orv *OrphanedRecordsValidator) validateSpecificOrphanedRecords(ctx context.Context) ([]OrphanedRecordDetail, error) {
	var allDetails []OrphanedRecordDetail

	// Check for specific problematic scenarios
	scenarios := []OrphanedRecordCheck{
		// Check for businesses without users
		{
			TableName:        "businesses",
			ColumnName:       "user_id",
			ReferencedTable:  "users",
			ReferencedColumn: "id",
			Description:      "Businesses without valid user references",
		},

		// Check for classifications without businesses
		{
			TableName:        "business_classifications",
			ColumnName:       "business_id",
			ReferencedTable:  "businesses",
			ReferencedColumn: "id",
			Description:      "Business classifications without valid business references",
		},

		// Check for risk assessments without businesses
		{
			TableName:        "risk_assessments",
			ColumnName:       "business_id",
			ReferencedTable:  "businesses",
			ReferencedColumn: "id",
			Description:      "Risk assessments without valid business references",
		},

		// Check for industry keywords without industries
		{
			TableName:        "industry_keywords",
			ColumnName:       "industry_id",
			ReferencedTable:  "industries",
			ReferencedColumn: "id",
			Description:      "Industry keywords without valid industry references",
		},

		// Check for classification codes without industries
		{
			TableName:        "classification_codes",
			ColumnName:       "industry_id",
			ReferencedTable:  "industries",
			ReferencedColumn: "id",
			Description:      "Classification codes without valid industry references",
		},
	}

	for _, scenario := range scenarios {
		details, err := orv.getOrphanedRecordDetails(ctx, scenario, 10) // Limit to 10 details per scenario
		if err != nil {
			orv.validator.logger.Printf("Error getting orphaned record details for %s: %v", scenario.Description, err)
			continue
		}

		allDetails = append(allDetails, details...)
	}

	return allDetails, nil
}
