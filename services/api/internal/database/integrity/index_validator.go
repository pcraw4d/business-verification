// Package integrity provides index validation
// for the KYB Platform database schema.
package integrity

import (
	"context"
	"fmt"
)

// IndexValidator validates database indexes
type IndexValidator struct {
	validator *Validator
}

// Name returns the name of this validation check
func (iv *IndexValidator) Name() string {
	return "indexes"
}

// Description returns a description of this validation check
func (iv *IndexValidator) Description() string {
	return "Validates database indexes for optimal performance"
}

// Required returns whether this check is required
func (iv *IndexValidator) Required() bool {
	return false
}

// Execute performs index validation
func (iv *IndexValidator) Execute(ctx context.Context) (*ValidationResult, error) {
	details := make(map[string]interface{})
	var totalIssues int

	// Check for missing indexes
	missingIndexes, err := iv.checkMissingIndexes(ctx)
	if err != nil {
		return iv.validator.createResult(
			iv.Name(),
			StatusFailed,
			iv.validator.formatError(iv.Name(), "Failed to check missing indexes", err),
			details,
		), nil
	}

	if len(missingIndexes) > 0 {
		details["missing_indexes"] = missingIndexes
		totalIssues += len(missingIndexes)
	}

	// Check for duplicate indexes
	duplicateIndexes, err := iv.checkDuplicateIndexes(ctx)
	if err != nil {
		return iv.validator.createResult(
			iv.Name(),
			StatusFailed,
			iv.validator.formatError(iv.Name(), "Failed to check duplicate indexes", err),
			details,
		), nil
	}

	if len(duplicateIndexes) > 0 {
		details["duplicate_indexes"] = duplicateIndexes
		totalIssues += len(duplicateIndexes)
	}

	// Check for unused indexes
	unusedIndexes, err := iv.checkUnusedIndexes(ctx)
	if err != nil {
		return iv.validator.createResult(
			iv.Name(),
			StatusFailed,
			iv.validator.formatError(iv.Name(), "Failed to check unused indexes", err),
			details,
		), nil
	}

	if len(unusedIndexes) > 0 {
		details["unused_indexes"] = unusedIndexes
		totalIssues += len(unusedIndexes)
	}

	// Determine status
	status := StatusPassed
	message := "All index checks passed"

	if totalIssues > 0 {
		status = StatusWarning
		message = fmt.Sprintf("Found %d index issues", totalIssues)
	}

	details["total_issues"] = totalIssues

	return iv.validator.createResult(iv.Name(), status, message, details), nil
}

// checkMissingIndexes checks for missing critical indexes
func (iv *IndexValidator) checkMissingIndexes(ctx context.Context) ([]MissingIndex, error) {
	var missingIndexes []MissingIndex

	// Define expected indexes for optimal performance
	expectedIndexes := []ExpectedIndex{
		// User table indexes
		{TableName: "users", ColumnName: "email", IndexType: "unique"},
		{TableName: "users", ColumnName: "created_at", IndexType: "btree"},
		{TableName: "users", ColumnName: "is_active", IndexType: "btree"},

		// Business table indexes
		{TableName: "businesses", ColumnName: "user_id", IndexType: "btree"},
		{TableName: "businesses", ColumnName: "name", IndexType: "btree"},
		{TableName: "businesses", ColumnName: "created_at", IndexType: "btree"},

		// Classification table indexes
		{TableName: "industries", ColumnName: "name", IndexType: "unique"},
		{TableName: "industries", ColumnName: "is_active", IndexType: "btree"},
		{TableName: "industry_keywords", ColumnName: "industry_id", IndexType: "btree"},
		{TableName: "industry_keywords", ColumnName: "keyword", IndexType: "btree"},
		{TableName: "classification_codes", ColumnName: "industry_id", IndexType: "btree"},
		{TableName: "classification_codes", ColumnName: "code_type", IndexType: "btree"},
		{TableName: "classification_codes", ColumnName: "code", IndexType: "btree"},

		// Risk table indexes
		{TableName: "risk_keywords", ColumnName: "keyword", IndexType: "btree"},
		{TableName: "risk_keywords", ColumnName: "risk_category", IndexType: "btree"},
		{TableName: "risk_keywords", ColumnName: "risk_severity", IndexType: "btree"},
		{TableName: "business_risk_assessments", ColumnName: "business_id", IndexType: "btree"},
		{TableName: "business_risk_assessments", ColumnName: "risk_keyword_id", IndexType: "btree"},
		{TableName: "business_risk_assessments", ColumnName: "risk_level", IndexType: "btree"},

		// Performance table indexes
		{TableName: "business_classifications", ColumnName: "user_id", IndexType: "btree"},
		{TableName: "business_classifications", ColumnName: "business_id", IndexType: "btree"},
		{TableName: "business_classifications", ColumnName: "created_at", IndexType: "btree"},
		{TableName: "risk_assessments", ColumnName: "user_id", IndexType: "btree"},
		{TableName: "risk_assessments", ColumnName: "business_id", IndexType: "btree"},
		{TableName: "risk_assessments", ColumnName: "created_at", IndexType: "btree"},
	}

	// Get existing indexes
	existingIndexes, err := iv.getExistingIndexes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing indexes: %w", err)
	}

	// Check for missing indexes
	for _, expected := range expectedIndexes {
		// Check if table exists
		exists, err := iv.validator.tableExists(ctx, expected.TableName)
		if err != nil || !exists {
			continue
		}

		found := false
		for _, existing := range existingIndexes {
			if existing.TableName == expected.TableName && existing.ColumnName == expected.ColumnName {
				found = true
				break
			}
		}

		if !found {
			missingIndex := MissingIndex{
				TableName:  expected.TableName,
				ColumnName: expected.ColumnName,
				IndexType:  expected.IndexType,
				Reason:     fmt.Sprintf("Missing %s index on %s.%s for optimal query performance", expected.IndexType, expected.TableName, expected.ColumnName),
			}
			missingIndexes = append(missingIndexes, missingIndex)
		}
	}

	return missingIndexes, nil
}

// checkDuplicateIndexes checks for duplicate indexes
func (iv *IndexValidator) checkDuplicateIndexes(ctx context.Context) ([]DuplicateIndex, error) {
	var duplicateIndexes []DuplicateIndex

	// Get all indexes grouped by table and columns
	query := `
		SELECT 
			t.relname as table_name,
			i.relname as index_name,
			a.attname as column_name,
			am.amname as index_type
		FROM 
			pg_class t,
			pg_class i,
			pg_index ix,
			pg_attribute a,
			pg_am am
		WHERE 
			t.oid = ix.indrelid
			AND i.oid = ix.indexrelid
			AND a.attrelid = t.oid
			AND a.attnum = ANY(ix.indkey)
			AND t.relkind = 'r'
			AND am.oid = i.relam
			AND t.relname NOT LIKE 'pg_%'
		ORDER BY t.relname, i.relname, a.attnum
	`

	rows, err := iv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get index information: %w", err)
	}
	defer rows.Close()

	indexMap := make(map[string][]IndexInfo)

	for rows.Next() {
		var indexInfo IndexInfo
		err := rows.Scan(&indexInfo.TableName, &indexInfo.IndexName, &indexInfo.ColumnName, &indexInfo.IndexType)
		if err != nil {
			continue
		}

		key := fmt.Sprintf("%s.%s", indexInfo.TableName, indexInfo.ColumnName)
		indexMap[key] = append(indexMap[key], indexInfo)
	}

	// Find duplicates
	for key, indexes := range indexMap {
		if len(indexes) > 1 {
			duplicate := DuplicateIndex{
				TableName:   indexes[0].TableName,
				ColumnName:  indexes[0].ColumnName,
				IndexCount:  len(indexes),
				IndexNames:  make([]string, len(indexes)),
				Description: fmt.Sprintf("Multiple indexes found on %s: %d indexes", key, len(indexes)),
			}

			for i, index := range indexes {
				duplicate.IndexNames[i] = index.IndexName
			}

			duplicateIndexes = append(duplicateIndexes, duplicate)
		}
	}

	return duplicateIndexes, nil
}

// checkUnusedIndexes checks for unused indexes
func (iv *IndexValidator) checkUnusedIndexes(ctx context.Context) ([]UnusedIndex, error) {
	var unusedIndexes []UnusedIndex

	// Check for indexes with zero usage
	query := `
		SELECT 
			schemaname,
			tablename,
			indexname,
			idx_tup_read,
			idx_tup_fetch
		FROM 
			pg_stat_user_indexes
		WHERE 
			idx_tup_read = 0
			AND idx_tup_fetch = 0
			AND schemaname = 'public'
		ORDER BY tablename, indexname
	`

	rows, err := iv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get index usage statistics: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var unused UnusedIndex
		err := rows.Scan(&unused.SchemaName, &unused.TableName, &unused.IndexName, &unused.TupleReads, &unused.TupleFetches)
		if err != nil {
			continue
		}

		unused.Description = fmt.Sprintf("Index %s on table %s has never been used (0 reads, 0 fetches)", unused.IndexName, unused.TableName)
		unusedIndexes = append(unusedIndexes, unused)
	}

	return unusedIndexes, nil
}

// getExistingIndexes retrieves all existing indexes
func (iv *IndexValidator) getExistingIndexes(ctx context.Context) ([]IndexInfo, error) {
	var indexes []IndexInfo

	query := `
		SELECT 
			t.relname as table_name,
			i.relname as index_name,
			a.attname as column_name,
			am.amname as index_type
		FROM 
			pg_class t,
			pg_class i,
			pg_index ix,
			pg_attribute a,
			pg_am am
		WHERE 
			t.oid = ix.indrelid
			AND i.oid = ix.indexrelid
			AND a.attrelid = t.oid
			AND a.attnum = ANY(ix.indkey)
			AND t.relkind = 'r'
			AND am.oid = i.relam
			AND t.relname NOT LIKE 'pg_%'
			AND t.relname NOT LIKE 'sql_%'
		ORDER BY t.relname, i.relname, a.attnum
	`

	rows, err := iv.validator.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing indexes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var index IndexInfo
		err := rows.Scan(&index.TableName, &index.IndexName, &index.ColumnName, &index.IndexType)
		if err != nil {
			continue
		}

		indexes = append(indexes, index)
	}

	return indexes, nil
}

// ExpectedIndex represents an expected index
type ExpectedIndex struct {
	TableName  string
	ColumnName string
	IndexType  string
}

// IndexInfo represents index information
type IndexInfo struct {
	TableName  string
	IndexName  string
	ColumnName string
	IndexType  string
}

// MissingIndex represents a missing index
type MissingIndex struct {
	TableName  string
	ColumnName string
	IndexType  string
	Reason     string
}

// DuplicateIndex represents duplicate indexes
type DuplicateIndex struct {
	TableName   string
	ColumnName  string
	IndexCount  int
	IndexNames  []string
	Description string
}

// UnusedIndex represents an unused index
type UnusedIndex struct {
	SchemaName   string
	TableName    string
	IndexName    string
	TupleReads   int64
	TupleFetches int64
	Description  string
}
