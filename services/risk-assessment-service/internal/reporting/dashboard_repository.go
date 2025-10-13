package reporting

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// SQLDashboardRepository implements DashboardRepository using SQL database
type SQLDashboardRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewSQLDashboardRepository creates a new SQL dashboard repository
func NewSQLDashboardRepository(db *sql.DB, logger *zap.Logger) *SQLDashboardRepository {
	return &SQLDashboardRepository{
		db:     db,
		logger: logger,
	}
}

// SaveDashboard saves a dashboard to the database
func (r *SQLDashboardRepository) SaveDashboard(ctx context.Context, dashboard *RiskDashboard) error {
	r.logger.Info("Saving dashboard",
		zap.String("dashboard_id", dashboard.ID),
		zap.String("tenant_id", dashboard.TenantID),
		zap.String("name", dashboard.Name))

	// Convert complex fields to JSON
	summaryJSON, err := json.Marshal(dashboard.Summary)
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	trendsJSON, err := json.Marshal(dashboard.Trends)
	if err != nil {
		return fmt.Errorf("failed to marshal trends: %w", err)
	}

	predictionsJSON, err := json.Marshal(dashboard.Predictions)
	if err != nil {
		return fmt.Errorf("failed to marshal predictions: %w", err)
	}

	chartsJSON, err := json.Marshal(dashboard.Charts)
	if err != nil {
		return fmt.Errorf("failed to marshal charts: %w", err)
	}

	filtersJSON, err := json.Marshal(dashboard.Filters)
	if err != nil {
		return fmt.Errorf("failed to marshal filters: %w", err)
	}

	metadataJSON, err := json.Marshal(dashboard.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Check if dashboard exists
	existingDashboard, err := r.GetDashboard(ctx, dashboard.TenantID, dashboard.ID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing dashboard: %w", err)
	}

	if existingDashboard != nil {
		// Update existing dashboard
		query := `
			UPDATE dashboards SET
				name = $1, type = $2, summary = $3, trends = $4, predictions = $5,
				charts = $6, filters = $7, is_public = $8, updated_at = $9, metadata = $10
			WHERE id = $11 AND tenant_id = $12
		`
		_, err = r.db.ExecContext(ctx, query,
			dashboard.Name, dashboard.Type, summaryJSON, trendsJSON, predictionsJSON,
			chartsJSON, filtersJSON, dashboard.IsPublic, dashboard.UpdatedAt, metadataJSON,
			dashboard.ID, dashboard.TenantID)
	} else {
		// Insert new dashboard
		query := `
			INSERT INTO dashboards (
				id, tenant_id, name, type, summary, trends, predictions,
				charts, filters, is_public, created_at, updated_at, created_by, metadata
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
			)
		`
		_, err = r.db.ExecContext(ctx, query,
			dashboard.ID, dashboard.TenantID, dashboard.Name, dashboard.Type,
			summaryJSON, trendsJSON, predictionsJSON, chartsJSON, filtersJSON,
			dashboard.IsPublic, dashboard.CreatedAt, dashboard.UpdatedAt,
			dashboard.CreatedBy, metadataJSON)
	}

	if err != nil {
		return fmt.Errorf("failed to save dashboard: %w", err)
	}

	r.logger.Info("Dashboard saved successfully",
		zap.String("dashboard_id", dashboard.ID),
		zap.String("tenant_id", dashboard.TenantID))

	return nil
}

// GetDashboard retrieves a dashboard by ID
func (r *SQLDashboardRepository) GetDashboard(ctx context.Context, tenantID, dashboardID string) (*RiskDashboard, error) {
	r.logger.Debug("Getting dashboard",
		zap.String("dashboard_id", dashboardID),
		zap.String("tenant_id", tenantID))

	query := `
		SELECT id, tenant_id, name, type, summary, trends, predictions,
			   charts, filters, is_public, created_at, updated_at, created_by, metadata
		FROM dashboards
		WHERE id = $1 AND tenant_id = $2
	`

	var result struct {
		ID          string    `json:"id"`
		TenantID    string    `json:"tenant_id"`
		Name        string    `json:"name"`
		Type        string    `json:"type"`
		Summary     string    `json:"summary"`
		Trends      string    `json:"trends"`
		Predictions string    `json:"predictions"`
		Charts      string    `json:"charts"`
		Filters     string    `json:"filters"`
		IsPublic    bool      `json:"is_public"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		CreatedBy   string    `json:"created_by"`
		Metadata    string    `json:"metadata"`
	}

	err := r.db.QueryRowContext(ctx, query, dashboardID, tenantID).Scan(
		&result.ID, &result.TenantID, &result.Name, &result.Type,
		&result.Summary, &result.Trends, &result.Predictions, &result.Charts,
		&result.Filters, &result.IsPublic, &result.CreatedAt, &result.UpdatedAt,
		&result.CreatedBy, &result.Metadata)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Dashboard not found
		}
		return nil, fmt.Errorf("failed to get dashboard: %w", err)
	}

	// Convert the result to RiskDashboard
	dashboard, err := r.convertToDashboard(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert result: %w", err)
	}

	r.logger.Debug("Dashboard retrieved successfully",
		zap.String("dashboard_id", dashboardID),
		zap.String("tenant_id", tenantID))

	return dashboard, nil
}

// ListDashboards lists dashboards with filters
func (r *SQLDashboardRepository) ListDashboards(ctx context.Context, filter *DashboardFilter) ([]*RiskDashboard, error) {
	r.logger.Debug("Listing dashboards",
		zap.String("tenant_id", filter.TenantID),
		zap.String("type", string(filter.Type)))

	// Build query with filters
	query := `
		SELECT id, tenant_id, name, type, summary, trends, predictions,
			   charts, filters, is_public, created_at, updated_at, created_by, metadata
		FROM dashboards
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if filter.TenantID != "" {
		query += fmt.Sprintf(" AND tenant_id = $%d", argIndex)
		args = append(args, filter.TenantID)
		argIndex++
	}

	if filter.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, string(filter.Type))
		argIndex++
	}

	if filter.CreatedBy != "" {
		query += fmt.Sprintf(" AND created_by = $%d", argIndex)
		args = append(args, filter.CreatedBy)
		argIndex++
	}

	if filter.IsPublic != nil {
		query += fmt.Sprintf(" AND is_public = $%d", argIndex)
		args = append(args, *filter.IsPublic)
		argIndex++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filter.EndDate)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
		argIndex++
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list dashboards: %w", err)
	}
	defer rows.Close()

	var dashboards []*RiskDashboard
	for rows.Next() {
		var result struct {
			ID          string    `json:"id"`
			TenantID    string    `json:"tenant_id"`
			Name        string    `json:"name"`
			Type        string    `json:"type"`
			Summary     string    `json:"summary"`
			Trends      string    `json:"trends"`
			Predictions string    `json:"predictions"`
			Charts      string    `json:"charts"`
			Filters     string    `json:"filters"`
			IsPublic    bool      `json:"is_public"`
			CreatedAt   time.Time `json:"created_at"`
			UpdatedAt   time.Time `json:"updated_at"`
			CreatedBy   string    `json:"created_by"`
			Metadata    string    `json:"metadata"`
		}

		err := rows.Scan(
			&result.ID, &result.TenantID, &result.Name, &result.Type,
			&result.Summary, &result.Trends, &result.Predictions, &result.Charts,
			&result.Filters, &result.IsPublic, &result.CreatedAt, &result.UpdatedAt,
			&result.CreatedBy, &result.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		dashboard, err := r.convertToDashboard(&result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert result: %w", err)
		}
		dashboards = append(dashboards, dashboard)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	r.logger.Debug("Dashboards listed successfully",
		zap.String("tenant_id", filter.TenantID),
		zap.Int("count", len(dashboards)))

	return dashboards, nil
}

// DeleteDashboard deletes a dashboard
func (r *SQLDashboardRepository) DeleteDashboard(ctx context.Context, tenantID, dashboardID string) error {
	r.logger.Info("Deleting dashboard",
		zap.String("dashboard_id", dashboardID),
		zap.String("tenant_id", tenantID))

	_, err := r.db.ExecContext(ctx, "DELETE FROM dashboards WHERE id = $1 AND tenant_id = $2", dashboardID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete dashboard: %w", err)
	}

	r.logger.Info("Dashboard deleted successfully",
		zap.String("dashboard_id", dashboardID),
		zap.String("tenant_id", tenantID))

	return nil
}

// GetDashboardMetrics gets dashboard usage metrics
func (r *SQLDashboardRepository) GetDashboardMetrics(ctx context.Context, tenantID string) (*DashboardMetrics, error) {
	r.logger.Debug("Getting dashboard metrics",
		zap.String("tenant_id", tenantID))

	query := `
		SELECT 
			COUNT(*) as total_dashboards,
			COUNT(CASE WHEN is_public = true THEN 1 END) as public_dashboards,
			COUNT(CASE WHEN is_public = false THEN 1 END) as private_dashboards
		FROM dashboards
		WHERE tenant_id = $1
	`

	var metrics DashboardMetrics
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&metrics.TotalDashboards, &metrics.PublicDashboards, &metrics.PrivateDashboards)

	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard metrics: %w", err)
	}

	// Get most viewed dashboards (this would require a separate views table)
	metrics.MostViewed = []DashboardViewData{}
	metrics.AverageViews = 0.0
	metrics.TotalViews = 0

	r.logger.Debug("Dashboard metrics retrieved",
		zap.String("tenant_id", tenantID),
		zap.Int("total_dashboards", metrics.TotalDashboards))

	return &metrics, nil
}

// RecordDashboardView records a dashboard view
func (r *SQLDashboardRepository) RecordDashboardView(ctx context.Context, tenantID, dashboardID string) error {
	r.logger.Debug("Recording dashboard view",
		zap.String("dashboard_id", dashboardID),
		zap.String("tenant_id", tenantID))

	// This would insert into a dashboard_views table
	// For now, we'll just log the view
	query := `
		INSERT INTO dashboard_views (dashboard_id, tenant_id, viewed_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (dashboard_id, tenant_id, viewed_at) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, dashboardID, tenantID, time.Now())
	if err != nil {
		// If the table doesn't exist, just log a warning
		r.logger.Warn("Failed to record dashboard view (table may not exist)", zap.Error(err))
		return nil
	}

	r.logger.Debug("Dashboard view recorded",
		zap.String("dashboard_id", dashboardID),
		zap.String("tenant_id", tenantID))

	return nil
}

// convertToDashboard converts a database result to RiskDashboard
func (r *SQLDashboardRepository) convertToDashboard(result interface{}) (*RiskDashboard, error) {
	// Type assertion to get the fields
	var id, tenantID, name, dashboardType, summary, trends, predictions, charts, filters, metadata string
	var isPublic bool
	var createdAt, updatedAt time.Time
	var createdBy string

	// Use reflection or type assertion based on the actual structure
	switch v := result.(type) {
	case *struct {
		ID          string    `json:"id"`
		TenantID    string    `json:"tenant_id"`
		Name        string    `json:"name"`
		Type        string    `json:"type"`
		Summary     string    `json:"summary"`
		Trends      string    `json:"trends"`
		Predictions string    `json:"predictions"`
		Charts      string    `json:"charts"`
		Filters     string    `json:"filters"`
		IsPublic    bool      `json:"is_public"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		CreatedBy   string    `json:"created_by"`
		Metadata    string    `json:"metadata"`
	}:
		id = v.ID
		tenantID = v.TenantID
		name = v.Name
		dashboardType = v.Type
		summary = v.Summary
		trends = v.Trends
		predictions = v.Predictions
		charts = v.Charts
		filters = v.Filters
		isPublic = v.IsPublic
		createdAt = v.CreatedAt
		updatedAt = v.UpdatedAt
		createdBy = v.CreatedBy
		metadata = v.Metadata
	default:
		return nil, fmt.Errorf("unsupported result type: %T", result)
	}

	// Parse JSON fields
	var summaryData DashboardSummary
	if summary != "" {
		if err := json.Unmarshal([]byte(summary), &summaryData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal summary: %w", err)
		}
	}

	var trendsData DashboardTrends
	if trends != "" {
		if err := json.Unmarshal([]byte(trends), &trendsData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal trends: %w", err)
		}
	}

	var predictionsData DashboardPredictions
	if predictions != "" {
		if err := json.Unmarshal([]byte(predictions), &predictionsData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal predictions: %w", err)
		}
	}

	var chartsData []DashboardChart
	if charts != "" {
		if err := json.Unmarshal([]byte(charts), &chartsData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal charts: %w", err)
		}
	}

	var filtersData DashboardFilters
	if filters != "" {
		if err := json.Unmarshal([]byte(filters), &filtersData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal filters: %w", err)
		}
	}

	var metadataMap map[string]interface{}
	if metadata != "" {
		if err := json.Unmarshal([]byte(metadata), &metadataMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	dashboard := &RiskDashboard{
		ID:          id,
		TenantID:    tenantID,
		Name:        name,
		Type:        DashboardType(dashboardType),
		Summary:     summaryData,
		Trends:      trendsData,
		Predictions: predictionsData,
		Charts:      chartsData,
		Filters:     filtersData,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		CreatedBy:   createdBy,
		IsPublic:    isPublic,
		Metadata:    metadataMap,
	}

	return dashboard, nil
}
