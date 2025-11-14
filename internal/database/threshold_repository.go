package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// ThresholdConfigData represents threshold configuration data for database operations
// This avoids importing risk package directly to break import cycles
type ThresholdConfigData struct {
	ID             string
	Name           string
	Description    string
	Category       string
	IndustryCode   string
	BusinessType   string
	RiskLevels     map[string]float64 // JSON-serialized risk levels
	IsDefault      bool
	IsActive       bool
	Priority       int
	Metadata       map[string]interface{}
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CreatedBy      string
	LastModifiedBy string
}

// ThresholdRepository provides data access operations for risk thresholds
type ThresholdRepository struct {
	db     *sql.DB
	logger *log.Logger
}

// NewThresholdRepository creates a new threshold repository
func NewThresholdRepository(db *sql.DB, logger *log.Logger) *ThresholdRepository {
	if logger == nil {
		logger = log.Default()
	}

	return &ThresholdRepository{
		db:     db,
		logger: logger,
	}
}

// CreateThreshold creates a new threshold configuration in the database
func (r *ThresholdRepository) CreateThreshold(ctx context.Context, config *ThresholdConfigData) error {
	// Serialize risk_levels to JSONB
	riskLevelsJSON, err := json.Marshal(config.RiskLevels)
	if err != nil {
		return fmt.Errorf("failed to marshal risk levels: %w", err)
	}

	// Serialize metadata to JSONB - always provide a value (empty object if nil)
	var metadataJSON []byte
	if config.Metadata != nil && len(config.Metadata) > 0 {
		metadataJSON, err = json.Marshal(config.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	} else {
		// Use empty JSON object instead of NULL to avoid type inference issues
		metadataJSON = []byte("{}")
	}

	query := `
		INSERT INTO risk_thresholds (
			id, name, description, category, industry_code, business_type,
			risk_levels, is_default, is_active, priority, metadata,
			created_by, last_modified_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	now := time.Now()
	// Convert JSON bytes to string - PostgreSQL JSONB columns accept text/string input
	// lib/pq will handle the conversion to JSONB type
	riskLevelsJSONStr := string(riskLevelsJSON)
	metadataJSONStr := string(metadataJSON)

	_, err = r.db.ExecContext(ctx, query,
		config.ID,
		config.Name,
		config.Description,
		config.Category,
		config.IndustryCode,
		config.BusinessType,
		riskLevelsJSONStr,
		config.IsDefault,
		config.IsActive,
		config.Priority,
		metadataJSONStr,
		config.CreatedBy,
		config.LastModifiedBy,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create threshold: %w", err)
	}

	r.logger.Printf("Created threshold: %s (%s)", config.ID, config.Name)
	return nil
}

// GetThreshold retrieves a threshold configuration by ID
func (r *ThresholdRepository) GetThreshold(ctx context.Context, id string) (*ThresholdConfigData, error) {
	query := `
		SELECT id, name, description, category, industry_code, business_type,
		       risk_levels, is_default, is_active, priority, metadata,
		       created_by, last_modified_by, created_at, updated_at
		FROM risk_thresholds
		WHERE id = $1
	`

	var config ThresholdConfigData
	var riskLevelsJSON []byte
	var metadataJSON []byte
	var industryCode, businessType sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&config.ID,
		&config.Name,
		&config.Description,
		&config.Category,
		&industryCode,
		&businessType,
		&riskLevelsJSON,
		&config.IsDefault,
		&config.IsActive,
		&config.Priority,
		&metadataJSON,
		&config.CreatedBy,
		&config.LastModifiedBy,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("threshold not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get threshold: %w", err)
	}

	// Parse optional fields
	if industryCode.Valid {
		config.IndustryCode = industryCode.String
	}
	if businessType.Valid {
		config.BusinessType = businessType.String
	}

	// Deserialize risk_levels (as map[string]float64)
	if err := json.Unmarshal(riskLevelsJSON, &config.RiskLevels); err != nil {
		return nil, fmt.Errorf("failed to unmarshal risk levels: %w", err)
	}

	// Deserialize metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &config.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &config, nil
}

// UpdateThreshold updates an existing threshold configuration
func (r *ThresholdRepository) UpdateThreshold(ctx context.Context, config *ThresholdConfigData) error {
	// Serialize risk_levels to JSONB
	riskLevelsJSON, err := json.Marshal(config.RiskLevels)
	if err != nil {
		return fmt.Errorf("failed to marshal risk levels: %w", err)
	}

	// Serialize metadata to JSONB - always provide a value (empty object if nil)
	var metadataJSON []byte
	if config.Metadata != nil && len(config.Metadata) > 0 {
		metadataJSON, err = json.Marshal(config.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	} else {
		// Use empty JSON object instead of NULL
		metadataJSON = []byte("{}")
	}

	query := `
		UPDATE risk_thresholds
		SET name = $2, description = $3, category = $4, industry_code = $5, business_type = $6,
		    risk_levels = $7, is_default = $8, is_active = $9, priority = $10, metadata = $11,
		    last_modified_by = $12, updated_at = $13
		WHERE id = $1
	`

	// Convert JSON bytes to string for PostgreSQL JSONB
	riskLevelsJSONStr := string(riskLevelsJSON)
	metadataJSONStr := string(metadataJSON)

	result, err := r.db.ExecContext(ctx, query,
		config.ID,
		config.Name,
		config.Description,
		config.Category,
		config.IndustryCode,
		config.BusinessType,
		riskLevelsJSONStr,
		config.IsDefault,
		config.IsActive,
		config.Priority,
		metadataJSONStr,
		config.LastModifiedBy,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update threshold: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("threshold not found: %s", config.ID)
	}

	r.logger.Printf("Updated threshold: %s (%s)", config.ID, config.Name)
	return nil
}

// DeleteThreshold deletes a threshold configuration
func (r *ThresholdRepository) DeleteThreshold(ctx context.Context, id string) error {
	query := `DELETE FROM risk_thresholds WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete threshold: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("threshold not found: %s", id)
	}

	r.logger.Printf("Deleted threshold: %s", id)
	return nil
}

// ListThresholds retrieves all threshold configurations with optional filters
func (r *ThresholdRepository) ListThresholds(ctx context.Context, category *string, industryCode *string, activeOnly bool) ([]*ThresholdConfigData, error) {
	query := `
		SELECT id, name, description, category, industry_code, business_type,
		       risk_levels, is_default, is_active, priority, metadata,
		       created_by, last_modified_by, created_at, updated_at
		FROM risk_thresholds
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if category != nil {
		query += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, *category)
		argIndex++
	}

	if industryCode != nil {
		query += fmt.Sprintf(" AND industry_code = $%d", argIndex)
		args = append(args, *industryCode)
		argIndex++
	}

	if activeOnly {
		query += fmt.Sprintf(" AND is_active = $%d", argIndex)
		args = append(args, true)
		argIndex++
	}

	query += " ORDER BY priority DESC, created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list thresholds: %w", err)
	}
	defer rows.Close()

	var configs []*ThresholdConfigData
	for rows.Next() {
		var config ThresholdConfigData
		var riskLevelsJSON []byte
		var metadataJSON []byte
		var industryCode, businessType sql.NullString

		err := rows.Scan(
			&config.ID,
			&config.Name,
			&config.Description,
			&config.Category,
			&industryCode,
			&businessType,
			&riskLevelsJSON,
			&config.IsDefault,
			&config.IsActive,
			&config.Priority,
			&metadataJSON,
			&config.CreatedBy,
			&config.LastModifiedBy,
			&config.CreatedAt,
			&config.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan threshold: %w", err)
		}

		// Parse optional fields
		if industryCode.Valid {
			config.IndustryCode = industryCode.String
		}
		if businessType.Valid {
			config.BusinessType = businessType.String
		}

		// Deserialize risk_levels (as map[string]float64)
		if err := json.Unmarshal(riskLevelsJSON, &config.RiskLevels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal risk levels: %w", err)
		}

		// Deserialize metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &config.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		configs = append(configs, &config)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating thresholds: %w", err)
	}

	return configs, nil
}

// LoadAllThresholds loads all active thresholds from the database
func (r *ThresholdRepository) LoadAllThresholds(ctx context.Context) ([]*ThresholdConfigData, error) {
	return r.ListThresholds(ctx, nil, nil, true)
}

