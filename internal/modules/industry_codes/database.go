package industry_codes

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// IndustryCode represents a single industry classification code
type IndustryCode struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Type        CodeType  `json:"type"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Subcategory string    `json:"subcategory,omitempty"`
	Keywords    []string  `json:"keywords,omitempty"`
	Confidence  float64   `json:"confidence"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CodeType represents the type of industry classification code
type CodeType string

const (
	CodeTypeMCC   CodeType = "mcc"
	CodeTypeSIC   CodeType = "sic"
	CodeTypeNAICS CodeType = "naics"
)

// IndustryCodeDatabase provides database operations for industry codes
type IndustryCodeDatabase struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewIndustryCodeDatabase creates a new industry code database instance
func NewIndustryCodeDatabase(db *sql.DB, logger *zap.Logger) *IndustryCodeDatabase {
	return &IndustryCodeDatabase{
		db:     db,
		logger: logger,
	}
}

// Initialize creates the industry codes table if it doesn't exist
func (icdb *IndustryCodeDatabase) Initialize(ctx context.Context) error {
	// Detect database type for compatibility
	dbType := icdb.getDatabaseType()

	var query string
	if dbType == "sqlite3" {
		query = `
			CREATE TABLE IF NOT EXISTS industry_codes (
				id VARCHAR(50) PRIMARY KEY,
				code VARCHAR(20) NOT NULL,
				type VARCHAR(10) NOT NULL,
				description TEXT NOT NULL,
				category VARCHAR(255) NOT NULL,
				subcategory VARCHAR(255),
				keywords TEXT,
				confidence REAL DEFAULT 0.00,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(code, type)
			);
			
			CREATE INDEX IF NOT EXISTS idx_industry_codes_type ON industry_codes(type);
			CREATE INDEX IF NOT EXISTS idx_industry_codes_category ON industry_codes(category);
			CREATE INDEX IF NOT EXISTS idx_industry_codes_keywords ON industry_codes(keywords);
		`
	} else {
		// PostgreSQL syntax
		query = `
			CREATE TABLE IF NOT EXISTS industry_codes (
				id VARCHAR(50) PRIMARY KEY,
				code VARCHAR(20) NOT NULL,
				type VARCHAR(10) NOT NULL,
				description TEXT NOT NULL,
				category VARCHAR(255) NOT NULL,
				subcategory VARCHAR(255),
				keywords TEXT,
				confidence DECIMAL(3,2) DEFAULT 0.00,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(code, type)
			);
			
			CREATE INDEX IF NOT EXISTS idx_industry_codes_type ON industry_codes(type);
			CREATE INDEX IF NOT EXISTS idx_industry_codes_category ON industry_codes(category);
			CREATE INDEX IF NOT EXISTS idx_industry_codes_keywords ON industry_codes USING GIN(to_tsvector('english', keywords));
		`
	}

	_, err := icdb.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create industry_codes table: %w", err)
	}

	icdb.logger.Info("Industry codes table initialized successfully")
	return nil
}

// getDatabaseType detects the database type for compatibility
func (icdb *IndustryCodeDatabase) getDatabaseType() string {
	driver := icdb.db.Driver()
	driverType := fmt.Sprintf("%T", driver)

	if strings.Contains(driverType, "sqlite3") {
		return "sqlite3"
	}
	return "postgres"
}

// InsertCode adds a new industry code to the database
func (icdb *IndustryCodeDatabase) InsertCode(ctx context.Context, code *IndustryCode) error {
	query := `
		INSERT INTO industry_codes (id, code, type, description, category, subcategory, keywords, confidence)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (code, type) DO UPDATE SET
			description = EXCLUDED.description,
			category = EXCLUDED.category,
			subcategory = EXCLUDED.subcategory,
			keywords = EXCLUDED.keywords,
			confidence = EXCLUDED.confidence,
			updated_at = CURRENT_TIMESTAMP
	`

	keywordsStr := strings.Join(code.Keywords, ",")
	_, err := icdb.db.ExecContext(ctx, query,
		code.ID, code.Code, code.Type, code.Description,
		code.Category, code.Subcategory, keywordsStr, code.Confidence)

	if err != nil {
		return fmt.Errorf("failed to insert industry code: %w", err)
	}

	icdb.logger.Debug("Industry code inserted", zap.String("code", code.Code), zap.String("type", string(code.Type)))
	return nil
}

// GetCodeByID retrieves an industry code by its ID
func (icdb *IndustryCodeDatabase) GetCodeByID(ctx context.Context, id string) (*IndustryCode, error) {
	query := `
		SELECT id, code, type, description, category, subcategory, keywords, confidence, created_at, updated_at
		FROM industry_codes
		WHERE id = $1
	`

	var code IndustryCode
	var keywordsStr string

	err := icdb.db.QueryRowContext(ctx, query, id).Scan(
		&code.ID, &code.Code, &code.Type, &code.Description,
		&code.Category, &code.Subcategory, &keywordsStr, &code.Confidence,
		&code.CreatedAt, &code.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("industry code not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get industry code: %w", err)
	}

	if keywordsStr != "" {
		code.Keywords = strings.Split(keywordsStr, ",")
	} else {
		code.Keywords = []string{}
	}

	return &code, nil
}

// GetCodeByCodeAndType retrieves an industry code by its code and type
func (icdb *IndustryCodeDatabase) GetCodeByCodeAndType(ctx context.Context, code string, codeType CodeType) (*IndustryCode, error) {
	query := `
		SELECT id, code, type, description, category, subcategory, keywords, confidence, created_at, updated_at
		FROM industry_codes
		WHERE code = $1 AND type = $2
	`

	var industryCode IndustryCode
	var keywordsStr string

	err := icdb.db.QueryRowContext(ctx, query, code, codeType).Scan(
		&industryCode.ID, &industryCode.Code, &industryCode.Type, &industryCode.Description,
		&industryCode.Category, &industryCode.Subcategory, &keywordsStr, &industryCode.Confidence,
		&industryCode.CreatedAt, &industryCode.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("industry code not found: %s (%s)", code, codeType)
		}
		return nil, fmt.Errorf("failed to get industry code: %w", err)
	}

	if keywordsStr != "" {
		industryCode.Keywords = strings.Split(keywordsStr, ",")
	}

	return &industryCode, nil
}

// SearchCodes searches for industry codes by description, category, or keywords
func (icdb *IndustryCodeDatabase) SearchCodes(ctx context.Context, query string, codeType *CodeType, limit int) ([]*IndustryCode, error) {
	dbType := icdb.getDatabaseType()

	var baseQuery string
	if dbType == "sqlite3" {
		baseQuery = `
			SELECT id, code, type, description, category, subcategory, keywords, confidence, created_at, updated_at
			FROM industry_codes
			WHERE (
				description LIKE ? OR
				category LIKE ? OR
				keywords LIKE ?
			)
		`
	} else {
		baseQuery = `
			SELECT id, code, type, description, category, subcategory, keywords, confidence, created_at, updated_at
			FROM industry_codes
			WHERE (
				to_tsvector('english', description) @@ plainto_tsquery('english', $1) OR
				to_tsvector('english', category) @@ plainto_tsquery('english', $1) OR
				to_tsvector('english', keywords) @@ plainto_tsquery('english', $1) OR
				description ILIKE $2 OR
				category ILIKE $2 OR
				keywords ILIKE $2
			)
		`
	}

	var args []interface{}
	if dbType == "sqlite3" {
		args = append(args, "%"+query+"%", "%"+query+"%", "%"+query+"%")
	} else {
		args = append(args, query, "%"+query+"%")
	}
	argIndex := len(args) + 1

	if codeType != nil {
		if dbType == "sqlite3" {
			baseQuery += " AND type = ?"
			args = append(args, *codeType)
		} else {
			baseQuery += fmt.Sprintf(" AND type = $%d", argIndex)
			args = append(args, *codeType)
			argIndex++
		}
	}

	if dbType == "sqlite3" {
		baseQuery += " ORDER BY confidence DESC, description LIMIT ?"
		args = append(args, limit)
	} else {
		baseQuery += fmt.Sprintf(" ORDER BY confidence DESC, description LIMIT $%d", argIndex)
		args = append(args, limit)
	}

	rows, err := icdb.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search industry codes: %w", err)
	}
	defer rows.Close()

	var codes []*IndustryCode
	for rows.Next() {
		var code IndustryCode
		var keywordsStr string

		err := rows.Scan(
			&code.ID, &code.Code, &code.Type, &code.Description,
			&code.Category, &code.Subcategory, &keywordsStr, &code.Confidence,
			&code.CreatedAt, &code.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan industry code: %w", err)
		}

		if keywordsStr != "" {
			code.Keywords = strings.Split(keywordsStr, ",")
		}

		codes = append(codes, &code)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over industry codes: %w", err)
	}

	icdb.logger.Debug("Industry codes search completed",
		zap.String("query", query),
		zap.Int("results", len(codes)))

	return codes, nil
}

// GetCodesByType retrieves all industry codes of a specific type
func (icdb *IndustryCodeDatabase) GetCodesByType(ctx context.Context, codeType CodeType, limit int, offset int) ([]*IndustryCode, error) {
	query := `
		SELECT id, code, type, description, category, subcategory, keywords, confidence, created_at, updated_at
		FROM industry_codes
		WHERE type = $1
		ORDER BY code
		LIMIT $2 OFFSET $3
	`

	rows, err := icdb.db.QueryContext(ctx, query, codeType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get industry codes by type: %w", err)
	}
	defer rows.Close()

	var codes []*IndustryCode
	for rows.Next() {
		var code IndustryCode
		var keywordsStr string

		err := rows.Scan(
			&code.ID, &code.Code, &code.Type, &code.Description,
			&code.Category, &code.Subcategory, &keywordsStr, &code.Confidence,
			&code.CreatedAt, &code.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan industry code: %w", err)
		}

		if keywordsStr != "" {
			code.Keywords = strings.Split(keywordsStr, ",")
		}

		codes = append(codes, &code)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over industry codes: %w", err)
	}

	return codes, nil
}

// GetCodesByCategory retrieves industry codes by category
func (icdb *IndustryCodeDatabase) GetCodesByCategory(ctx context.Context, category string, codeType *CodeType, limit int) ([]*IndustryCode, error) {
	dbType := icdb.getDatabaseType()

	var baseQuery string
	if dbType == "sqlite3" {
		baseQuery = `
			SELECT id, code, type, description, category, subcategory, keywords, confidence, created_at, updated_at
			FROM industry_codes
			WHERE category LIKE ?
		`
	} else {
		baseQuery = `
			SELECT id, code, type, description, category, subcategory, keywords, confidence, created_at, updated_at
			FROM industry_codes
			WHERE category ILIKE $1
		`
	}

	var args []interface{}
	if dbType == "sqlite3" {
		args = append(args, "%"+category+"%")
	} else {
		args = append(args, "%"+category+"%")
	}
	argIndex := len(args) + 1

	if codeType != nil {
		if dbType == "sqlite3" {
			baseQuery += " AND type = ?"
			args = append(args, *codeType)
		} else {
			baseQuery += fmt.Sprintf(" AND type = $%d", argIndex)
			args = append(args, *codeType)
			argIndex++
		}
	}

	if dbType == "sqlite3" {
		baseQuery += " ORDER BY confidence DESC LIMIT ?"
		args = append(args, limit)
	} else {
		baseQuery += fmt.Sprintf(" ORDER BY confidence DESC LIMIT $%d", argIndex)
		args = append(args, limit)
	}

	rows, err := icdb.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get industry codes by category: %w", err)
	}
	defer rows.Close()

	var codes []*IndustryCode
	for rows.Next() {
		var code IndustryCode
		var keywordsStr string

		err := rows.Scan(
			&code.ID, &code.Code, &code.Type, &code.Description,
			&code.Category, &code.Subcategory, &keywordsStr, &code.Confidence,
			&code.CreatedAt, &code.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan industry code: %w", err)
		}

		if keywordsStr != "" {
			code.Keywords = strings.Split(keywordsStr, ",")
		}

		codes = append(codes, &code)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over industry codes: %w", err)
	}

	return codes, nil
}

// UpdateCodeConfidence updates the confidence score for an industry code
func (icdb *IndustryCodeDatabase) UpdateCodeConfidence(ctx context.Context, id string, confidence float64) error {
	query := `
		UPDATE industry_codes
		SET confidence = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := icdb.db.ExecContext(ctx, query, confidence, id)
	if err != nil {
		return fmt.Errorf("failed to update code confidence: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("industry code not found: %s", id)
	}

	icdb.logger.Debug("Industry code confidence updated",
		zap.String("id", id),
		zap.Float64("confidence", confidence))

	return nil
}

// GetCodeStats returns statistics about the industry codes database
func (icdb *IndustryCodeDatabase) GetCodeStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT 
			type,
			COUNT(*) as count,
			AVG(confidence) as avg_confidence,
			MIN(confidence) as min_confidence,
			MAX(confidence) as max_confidence
		FROM industry_codes
		GROUP BY type
	`

	rows, err := icdb.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get code stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]interface{})
	totalCodes := 0

	for rows.Next() {
		var codeType string
		var count int
		var avgConfidence, minConfidence, maxConfidence sql.NullFloat64

		err := rows.Scan(&codeType, &count, &avgConfidence, &minConfidence, &maxConfidence)
		if err != nil {
			return nil, fmt.Errorf("failed to scan code stats: %w", err)
		}

		typeStats := map[string]interface{}{
			"count": count,
		}

		if avgConfidence.Valid {
			typeStats["avg_confidence"] = avgConfidence.Float64
		}
		if minConfidence.Valid {
			typeStats["min_confidence"] = minConfidence.Float64
		}
		if maxConfidence.Valid {
			typeStats["max_confidence"] = maxConfidence.Float64
		}

		stats[codeType] = typeStats
		totalCodes += count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over code stats: %w", err)
	}

	stats["total_codes"] = totalCodes

	return stats, nil
}

// Close closes the database connection
func (icdb *IndustryCodeDatabase) Close() error {
	return icdb.db.Close()
}
