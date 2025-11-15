package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"kyb-platform/internal/models"
)

// RiskIndicatorsRepository provides data access operations for risk indicators
type RiskIndicatorsRepository struct {
	db     *sql.DB
	logger *log.Logger
}

// NewRiskIndicatorsRepository creates a new risk indicators repository
func NewRiskIndicatorsRepository(db *sql.DB, logger *log.Logger) *RiskIndicatorsRepository {
	if logger == nil {
		logger = log.Default()
	}

	return &RiskIndicatorsRepository{
		db:     db,
		logger: logger,
	}
}

// RiskIndicatorFilters represents filters for risk indicators
type RiskIndicatorFilters struct {
	Severity string // low, medium, high, critical
	Status   string // active, resolved, dismissed
}

// GetByMerchantID retrieves risk indicators for a merchant with optional filters
func (r *RiskIndicatorsRepository) GetByMerchantID(ctx context.Context, merchantID string, filters *RiskIndicatorFilters) ([]models.RiskIndicator, error) {
	r.logger.Printf("Getting risk indicators for merchant: %s", merchantID)

	// Build query with filters
	query := `
		SELECT 
			id, type, name, severity, status, description, detected_at, score
		FROM risk_indicators
		WHERE merchant_id = $1
	`

	args := []interface{}{merchantID}
	argIndex := 2

	// Add severity filter
	if filters != nil && filters.Severity != "" {
		query += fmt.Sprintf(" AND severity = $%d", argIndex)
		args = append(args, filters.Severity)
		argIndex++
	}

	// Add status filter
	if filters != nil && filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filters.Status)
		argIndex++
	}

	query += " ORDER BY detected_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query indicators: %w", err)
	}
	defer rows.Close()

	var indicators []models.RiskIndicator
	for rows.Next() {
		var indicator models.RiskIndicator
		var detectedAt time.Time

		err := rows.Scan(
			&indicator.ID,
			&indicator.Type,
			&indicator.Name,
			&indicator.Severity,
			&indicator.Status,
			&indicator.Description,
			&detectedAt,
			&indicator.Score,
		)
		if err != nil {
			r.logger.Printf("Error scanning indicator: %v", err)
			continue
		}

		indicator.DetectedAt = detectedAt
		indicators = append(indicators, indicator)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating indicators: %w", err)
	}

	// If no indicators found in database, return empty array (not an error)
	// In production, indicators would be populated from risk assessment results
	if len(indicators) == 0 {
		r.logger.Printf("No risk indicators found for merchant: %s", merchantID)
		return []models.RiskIndicator{}, nil
	}

	return indicators, nil
}

