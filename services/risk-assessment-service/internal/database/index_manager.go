package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// IndexManager handles automated index creation and management
type IndexManager struct {
	db     *sql.DB
	logger *zap.Logger
}

// IndexInfo represents information about a database index
type IndexInfo struct {
	Name      string    `json:"name"`
	Table     string    `json:"table"`
	Columns   []string  `json:"columns"`
	Type      string    `json:"type"`
	Size      int64     `json:"size_bytes"`
	Usage     int64     `json:"usage_count"`
	LastUsed  time.Time `json:"last_used"`
	IsUnique  bool      `json:"is_unique"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
}

// IndexStats represents statistics about index usage
type IndexStats struct {
	TotalIndexes     int     `json:"total_indexes"`
	UnusedIndexes    int     `json:"unused_indexes"`
	DuplicateIndexes int     `json:"duplicate_indexes"`
	TotalSize        int64   `json:"total_size_bytes"`
	AverageSize      float64 `json:"average_size_bytes"`
	UsageEfficiency  float64 `json:"usage_efficiency_percent"`
}

// NewIndexManager creates a new index manager
func NewIndexManager(db *sql.DB, logger *zap.Logger) *IndexManager {
	return &IndexManager{
		db:     db,
		logger: logger,
	}
}

// CreateIndex creates a new database index
func (im *IndexManager) CreateIndex(ctx context.Context, recommendation *IndexRecommendation) error {
	indexName := im.generateIndexName(recommendation.Table, recommendation.Columns)

	// Check if index already exists
	exists, err := im.indexExists(ctx, indexName)
	if err != nil {
		return fmt.Errorf("failed to check if index exists: %w", err)
	}

	if exists {
		im.logger.Info("Index already exists, skipping creation",
			zap.String("index_name", indexName))
		return nil
	}

	// Build CREATE INDEX statement
	createSQL := im.buildCreateIndexSQL(indexName, recommendation)

	im.logger.Info("Creating index",
		zap.String("index_name", indexName),
		zap.String("table", recommendation.Table),
		zap.Strings("columns", recommendation.Columns),
		zap.String("type", recommendation.Type))

	// Execute CREATE INDEX
	_, err = im.db.ExecContext(ctx, createSQL)
	if err != nil {
		return fmt.Errorf("failed to create index %s: %w", indexName, err)
	}

	im.logger.Info("Index created successfully",
		zap.String("index_name", indexName))

	return nil
}

// CreateIndexes creates multiple indexes from recommendations
func (im *IndexManager) CreateIndexes(ctx context.Context, recommendations []*IndexRecommendation) error {
	// Sort recommendations by priority (highest first)
	sortedRecs := im.sortByPriority(recommendations)

	var errors []string
	created := 0

	for _, rec := range sortedRecs {
		err := im.CreateIndex(ctx, rec)
		if err != nil {
			im.logger.Error("Failed to create index",
				zap.String("table", rec.Table),
				zap.Strings("columns", rec.Columns),
				zap.Error(err))
			errors = append(errors, fmt.Sprintf("%s: %v", rec.Table, err))
		} else {
			created++
		}
	}

	im.logger.Info("Index creation completed",
		zap.Int("created", created),
		zap.Int("failed", len(errors)))

	if len(errors) > 0 {
		return fmt.Errorf("failed to create %d indexes: %s", len(errors), strings.Join(errors, "; "))
	}

	return nil
}

// DropIndex drops a database index
func (im *IndexManager) DropIndex(ctx context.Context, indexName string) error {
	// Check if index exists
	exists, err := im.indexExists(ctx, indexName)
	if err != nil {
		return fmt.Errorf("failed to check if index exists: %w", err)
	}

	if !exists {
		im.logger.Info("Index does not exist, skipping drop",
			zap.String("index_name", indexName))
		return nil
	}

	dropSQL := fmt.Sprintf("DROP INDEX IF EXISTS %s", indexName)

	im.logger.Info("Dropping index",
		zap.String("index_name", indexName))

	_, err = im.db.ExecContext(ctx, dropSQL)
	if err != nil {
		return fmt.Errorf("failed to drop index %s: %w", indexName, err)
	}

	im.logger.Info("Index dropped successfully",
		zap.String("index_name", indexName))

	return nil
}

// GetIndexes retrieves all indexes for a table
func (im *IndexManager) GetIndexes(ctx context.Context, tableName string) ([]*IndexInfo, error) {
	query := `
		SELECT 
			i.indexname,
			i.tablename,
			array_agg(a.attname ORDER BY a.attnum) as columns,
			am.amname as type,
			pg_relation_size(i.indexrelid) as size,
			COALESCE(s.idx_tup_read, 0) as usage,
			COALESCE(s.last_idx_tup_read, '1970-01-01'::timestamp) as last_used,
			i.indexdef LIKE '%UNIQUE%' as is_unique,
			i.indexdef LIKE '%PRIMARY KEY%' as is_primary,
			pg_stat_get_tuples_inserted(c.oid) as created_at
		FROM pg_indexes i
		JOIN pg_class c ON c.relname = i.tablename
		JOIN pg_index idx ON idx.indexrelid = (i.schemaname||'.'||i.indexname)::regclass
		JOIN pg_attribute a ON a.attrelid = idx.indrelid AND a.attnum = ANY(idx.indkey)
		JOIN pg_am am ON am.oid = (SELECT relam FROM pg_class WHERE relname = i.indexname)
		LEFT JOIN pg_stat_user_indexes s ON s.indexrelname = i.indexname
		WHERE i.tablename = $1
		GROUP BY i.indexname, i.tablename, am.amname, s.idx_tup_read, s.last_idx_tup_read, i.indexdef, c.oid
		ORDER BY i.indexname
	`

	rows, err := im.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes: %w", err)
	}
	defer rows.Close()

	var indexes []*IndexInfo
	for rows.Next() {
		var index IndexInfo
		var columnsStr string

		err := rows.Scan(
			&index.Name,
			&index.Table,
			&columnsStr,
			&index.Type,
			&index.Size,
			&index.Usage,
			&index.LastUsed,
			&index.IsUnique,
			&index.IsPrimary,
			&index.CreatedAt,
		)
		if err != nil {
			continue
		}

		// Parse columns from string array
		index.Columns = im.parseColumns(columnsStr)
		indexes = append(indexes, &index)
	}

	return indexes, nil
}

// GetIndexStats retrieves statistics about all indexes
func (im *IndexManager) GetIndexStats(ctx context.Context) (*IndexStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_indexes,
			COUNT(CASE WHEN s.idx_tup_read = 0 THEN 1 END) as unused_indexes,
			SUM(pg_relation_size(i.indexrelid)) as total_size,
			AVG(pg_relation_size(i.indexrelid)) as average_size
		FROM pg_indexes i
		JOIN pg_class c ON c.relname = i.tablename
		LEFT JOIN pg_stat_user_indexes s ON s.indexrelname = i.indexname
		WHERE i.schemaname = 'public'
	`

	var stats IndexStats
	err := im.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalIndexes,
		&stats.UnusedIndexes,
		&stats.TotalSize,
		&stats.AverageSize,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get index stats: %w", err)
	}

	// Calculate usage efficiency
	if stats.TotalIndexes > 0 {
		usedIndexes := stats.TotalIndexes - stats.UnusedIndexes
		stats.UsageEfficiency = float64(usedIndexes) / float64(stats.TotalIndexes) * 100
	}

	return &stats, nil
}

// FindUnusedIndexes identifies indexes that are not being used
func (im *IndexManager) FindUnusedIndexes(ctx context.Context) ([]*IndexInfo, error) {
	query := `
		SELECT 
			i.indexname,
			i.tablename,
			array_agg(a.attname ORDER BY a.attnum) as columns,
			am.amname as type,
			pg_relation_size(i.indexrelid) as size,
			COALESCE(s.idx_tup_read, 0) as usage,
			COALESCE(s.last_idx_tup_read, '1970-01-01'::timestamp) as last_used,
			i.indexdef LIKE '%UNIQUE%' as is_unique,
			i.indexdef LIKE '%PRIMARY KEY%' as is_primary
		FROM pg_indexes i
		JOIN pg_class c ON c.relname = i.tablename
		JOIN pg_index idx ON idx.indexrelid = (i.schemaname||'.'||i.indexname)::regclass
		JOIN pg_attribute a ON a.attrelid = idx.indrelid AND a.attnum = ANY(idx.indkey)
		JOIN pg_am am ON am.oid = (SELECT relam FROM pg_class WHERE relname = i.indexname)
		LEFT JOIN pg_stat_user_indexes s ON s.indexrelname = i.indexname
		WHERE i.schemaname = 'public'
		AND (s.idx_tup_read = 0 OR s.idx_tup_read IS NULL)
		AND i.indexdef NOT LIKE '%PRIMARY KEY%'
		GROUP BY i.indexname, i.tablename, am.amname, s.idx_tup_read, s.last_idx_tup_read, i.indexdef
		ORDER BY pg_relation_size(i.indexrelid) DESC
	`

	rows, err := im.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find unused indexes: %w", err)
	}
	defer rows.Close()

	var unusedIndexes []*IndexInfo
	for rows.Next() {
		var index IndexInfo
		var columnsStr string

		err := rows.Scan(
			&index.Name,
			&index.Table,
			&columnsStr,
			&index.Type,
			&index.Size,
			&index.Usage,
			&index.LastUsed,
			&index.IsUnique,
			&index.IsPrimary,
		)
		if err != nil {
			continue
		}

		index.Columns = im.parseColumns(columnsStr)
		unusedIndexes = append(unusedIndexes, &index)
	}

	return unusedIndexes, nil
}

// FindDuplicateIndexes identifies duplicate or redundant indexes
func (im *IndexManager) FindDuplicateIndexes(ctx context.Context) ([][]*IndexInfo, error) {
	// This is a simplified implementation
	// A full implementation would need to analyze index column overlap
	query := `
		SELECT 
			i1.indexname as index1,
			i1.tablename as table1,
			i2.indexname as index2,
			i2.tablename as table2
		FROM pg_indexes i1
		JOIN pg_indexes i2 ON i1.tablename = i2.tablename 
		WHERE i1.indexname < i2.indexname
		AND i1.indexdef = i2.indexdef
		AND i1.schemaname = 'public'
		AND i2.schemaname = 'public'
	`

	rows, err := im.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find duplicate indexes: %w", err)
	}
	defer rows.Close()

	var duplicates [][]*IndexInfo
	for rows.Next() {
		var index1Name, table1Name, index2Name, table2Name string

		err := rows.Scan(&index1Name, &table1Name, &index2Name, &table2Name)
		if err != nil {
			continue
		}

		// Get full index info for both indexes
		index1, err := im.getIndexInfo(ctx, index1Name)
		if err != nil {
			continue
		}

		index2, err := im.getIndexInfo(ctx, index2Name)
		if err != nil {
			continue
		}

		duplicates = append(duplicates, []*IndexInfo{index1, index2})
	}

	return duplicates, nil
}

// OptimizeIndexes performs automatic index optimization
func (im *IndexManager) OptimizeIndexes(ctx context.Context) error {
	im.logger.Info("Starting automatic index optimization")

	// Get current index stats
	stats, err := im.GetIndexStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get index stats: %w", err)
	}

	im.logger.Info("Current index statistics",
		zap.Int("total_indexes", stats.TotalIndexes),
		zap.Int("unused_indexes", stats.UnusedIndexes),
		zap.Float64("usage_efficiency", stats.UsageEfficiency))

	// Find and report unused indexes
	unusedIndexes, err := im.FindUnusedIndexes(ctx)
	if err != nil {
		return fmt.Errorf("failed to find unused indexes: %w", err)
	}

	if len(unusedIndexes) > 0 {
		im.logger.Warn("Found unused indexes",
			zap.Int("count", len(unusedIndexes)))

		for _, index := range unusedIndexes {
			im.logger.Info("Unused index",
				zap.String("name", index.Name),
				zap.String("table", index.Table),
				zap.Int64("size_bytes", index.Size))
		}
	}

	// Find and report duplicate indexes
	duplicates, err := im.FindDuplicateIndexes(ctx)
	if err != nil {
		return fmt.Errorf("failed to find duplicate indexes: %w", err)
	}

	if len(duplicates) > 0 {
		im.logger.Warn("Found duplicate indexes",
			zap.Int("count", len(duplicates)))

		for _, pair := range duplicates {
			im.logger.Info("Duplicate indexes",
				zap.String("index1", pair[0].Name),
				zap.String("index2", pair[1].Name),
				zap.String("table", pair[0].Table))
		}
	}

	im.logger.Info("Index optimization analysis completed")

	return nil
}

// Helper methods

func (im *IndexManager) generateIndexName(table string, columns []string) string {
	// Create a standardized index name
	name := fmt.Sprintf("idx_%s_%s", table, strings.Join(columns, "_"))

	// Ensure name is not too long (PostgreSQL limit is 63 characters)
	if len(name) > 60 {
		name = name[:60]
	}

	return name
}

func (im *IndexManager) buildCreateIndexSQL(indexName string, rec *IndexRecommendation) string {
	columns := strings.Join(rec.Columns, ", ")

	var sql strings.Builder
	sql.WriteString(fmt.Sprintf("CREATE INDEX %s ON %s (%s)", indexName, rec.Table, columns))

	// Add index type if specified
	if rec.Type != "btree" {
		sql.WriteString(fmt.Sprintf(" USING %s", rec.Type))
	}

	return sql.String()
}

func (im *IndexManager) indexExists(ctx context.Context, indexName string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM pg_indexes WHERE indexname = $1)"

	var exists bool
	err := im.db.QueryRowContext(ctx, query, indexName).Scan(&exists)
	return exists, err
}

func (im *IndexManager) sortByPriority(recommendations []*IndexRecommendation) []*IndexRecommendation {
	// Simple bubble sort by priority (highest first)
	sorted := make([]*IndexRecommendation, len(recommendations))
	copy(sorted, recommendations)

	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Priority < sorted[j+1].Priority {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

func (im *IndexManager) parseColumns(columnsStr string) []string {
	// Remove curly braces and split by comma
	columnsStr = strings.Trim(columnsStr, "{}")
	if columnsStr == "" {
		return []string{}
	}

	columns := strings.Split(columnsStr, ",")
	for i, col := range columns {
		columns[i] = strings.TrimSpace(col)
	}

	return columns
}

func (im *IndexManager) getIndexInfo(ctx context.Context, indexName string) (*IndexInfo, error) {
	query := `
		SELECT 
			i.indexname,
			i.tablename,
			array_agg(a.attname ORDER BY a.attnum) as columns,
			am.amname as type,
			pg_relation_size(i.indexrelid) as size,
			COALESCE(s.idx_tup_read, 0) as usage,
			COALESCE(s.last_idx_tup_read, '1970-01-01'::timestamp) as last_used,
			i.indexdef LIKE '%UNIQUE%' as is_unique,
			i.indexdef LIKE '%PRIMARY KEY%' as is_primary
		FROM pg_indexes i
		JOIN pg_class c ON c.relname = i.tablename
		JOIN pg_index idx ON idx.indexrelid = (i.schemaname||'.'||i.indexname)::regclass
		JOIN pg_attribute a ON a.attrelid = idx.indrelid AND a.attnum = ANY(idx.indkey)
		JOIN pg_am am ON am.oid = (SELECT relam FROM pg_class WHERE relname = i.indexname)
		LEFT JOIN pg_stat_user_indexes s ON s.indexrelname = i.indexname
		WHERE i.indexname = $1
		GROUP BY i.indexname, i.tablename, am.amname, s.idx_tup_read, s.last_idx_tup_read, i.indexdef
	`

	var index IndexInfo
	var columnsStr string

	err := im.db.QueryRowContext(ctx, query, indexName).Scan(
		&index.Name,
		&index.Table,
		&columnsStr,
		&index.Type,
		&index.Size,
		&index.Usage,
		&index.LastUsed,
		&index.IsUnique,
		&index.IsPrimary,
	)
	if err != nil {
		return nil, err
	}

	index.Columns = im.parseColumns(columnsStr)
	return &index, nil
}
