package reporting

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// SQLScheduledReportRepository implements ScheduledReportRepository using SQL database
type SQLScheduledReportRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewSQLScheduledReportRepository creates a new SQL scheduled report repository
func NewSQLScheduledReportRepository(db *sql.DB, logger *zap.Logger) *SQLScheduledReportRepository {
	return &SQLScheduledReportRepository{
		db:     db,
		logger: logger,
	}
}

// SaveScheduledReport saves a scheduled report to the database
func (r *SQLScheduledReportRepository) SaveScheduledReport(ctx context.Context, scheduledReport *ScheduledReport) error {
	r.logger.Info("Saving scheduled report",
		zap.String("scheduled_report_id", scheduledReport.ID),
		zap.String("tenant_id", scheduledReport.TenantID),
		zap.String("name", scheduledReport.Name))

	// Convert complex fields to JSON
	scheduleJSON, err := json.Marshal(scheduledReport.Schedule)
	if err != nil {
		return fmt.Errorf("failed to marshal schedule: %w", err)
	}

	filtersJSON, err := json.Marshal(scheduledReport.Filters)
	if err != nil {
		return fmt.Errorf("failed to marshal filters: %w", err)
	}

	recipientsJSON, err := json.Marshal(scheduledReport.Recipients)
	if err != nil {
		return fmt.Errorf("failed to marshal recipients: %w", err)
	}

	metadataJSON, err := json.Marshal(scheduledReport.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Check if scheduled report exists
	existingReport, err := r.GetScheduledReport(ctx, scheduledReport.TenantID, scheduledReport.ID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing scheduled report: %w", err)
	}

	if existingReport != nil {
		// Update existing scheduled report
		query := `
			UPDATE scheduled_reports SET
				name = $1, template_id = $2, schedule = $3, filters = $4,
				recipients = $5, is_active = $6, last_run_at = $7, next_run_at = $8,
				updated_at = $9, metadata = $10
			WHERE id = $11 AND tenant_id = $12
		`
		_, err = r.db.ExecContext(ctx, query,
			scheduledReport.Name, scheduledReport.TemplateID, scheduleJSON, filtersJSON,
			recipientsJSON, scheduledReport.IsActive, scheduledReport.LastRunAt, scheduledReport.NextRunAt,
			time.Now(), metadataJSON, scheduledReport.ID, scheduledReport.TenantID)
	} else {
		// Insert new scheduled report
		query := `
			INSERT INTO scheduled_reports (
				id, tenant_id, name, template_id, schedule, filters,
				recipients, is_active, last_run_at, next_run_at,
				created_at, updated_at, created_by, metadata
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
			)
		`
		_, err = r.db.ExecContext(ctx, query,
			scheduledReport.ID, scheduledReport.TenantID, scheduledReport.Name, scheduledReport.TemplateID,
			scheduleJSON, filtersJSON, recipientsJSON, scheduledReport.IsActive,
			scheduledReport.LastRunAt, scheduledReport.NextRunAt, scheduledReport.CreatedAt,
			scheduledReport.UpdatedAt, scheduledReport.CreatedBy, metadataJSON)
	}

	if err != nil {
		return fmt.Errorf("failed to save scheduled report: %w", err)
	}

	r.logger.Info("Scheduled report saved successfully",
		zap.String("scheduled_report_id", scheduledReport.ID),
		zap.String("tenant_id", scheduledReport.TenantID))

	return nil
}

// GetScheduledReport retrieves a scheduled report by ID
func (r *SQLScheduledReportRepository) GetScheduledReport(ctx context.Context, tenantID, scheduledReportID string) (*ScheduledReport, error) {
	r.logger.Debug("Getting scheduled report",
		zap.String("scheduled_report_id", scheduledReportID),
		zap.String("tenant_id", tenantID))

	query := `
		SELECT id, tenant_id, name, template_id, schedule, filters,
			   recipients, is_active, last_run_at, next_run_at,
			   created_at, updated_at, created_by, metadata
		FROM scheduled_reports
		WHERE id = $1 AND tenant_id = $2
	`

	var result struct {
		ID         string     `json:"id"`
		TenantID   string     `json:"tenant_id"`
		Name       string     `json:"name"`
		TemplateID string     `json:"template_id"`
		Schedule   string     `json:"schedule"`
		Filters    string     `json:"filters"`
		Recipients string     `json:"recipients"`
		IsActive   bool       `json:"is_active"`
		LastRunAt  *time.Time `json:"last_run_at"`
		NextRunAt  *time.Time `json:"next_run_at"`
		CreatedAt  time.Time  `json:"created_at"`
		UpdatedAt  time.Time  `json:"updated_at"`
		CreatedBy  string     `json:"created_by"`
		Metadata   string     `json:"metadata"`
	}

	err := r.db.QueryRowContext(ctx, query, scheduledReportID, tenantID).Scan(
		&result.ID, &result.TenantID, &result.Name, &result.TemplateID,
		&result.Schedule, &result.Filters, &result.Recipients, &result.IsActive,
		&result.LastRunAt, &result.NextRunAt, &result.CreatedAt, &result.UpdatedAt,
		&result.CreatedBy, &result.Metadata)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Scheduled report not found
		}
		return nil, fmt.Errorf("failed to get scheduled report: %w", err)
	}

	// Convert the result to ScheduledReport
	scheduledReport, err := r.convertToScheduledReport(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert result: %w", err)
	}

	r.logger.Debug("Scheduled report retrieved successfully",
		zap.String("scheduled_report_id", scheduledReportID),
		zap.String("tenant_id", tenantID))

	return scheduledReport, nil
}

// ListScheduledReports lists scheduled reports with filters
func (r *SQLScheduledReportRepository) ListScheduledReports(ctx context.Context, filter *ScheduledReportFilter) ([]*ScheduledReport, error) {
	r.logger.Debug("Listing scheduled reports",
		zap.String("tenant_id", filter.TenantID))

	// Build query with filters
	query := `
		SELECT id, tenant_id, name, template_id, schedule, filters,
			   recipients, is_active, last_run_at, next_run_at,
			   created_at, updated_at, created_by, metadata
		FROM scheduled_reports
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if filter.TenantID != "" {
		query += fmt.Sprintf(" AND tenant_id = $%d", argIndex)
		args = append(args, filter.TenantID)
		argIndex++
	}

	if filter.IsActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argIndex)
		args = append(args, *filter.IsActive)
		argIndex++
	}

	if filter.CreatedBy != "" {
		query += fmt.Sprintf(" AND created_by = $%d", argIndex)
		args = append(args, filter.CreatedBy)
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
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var scheduledReports []*ScheduledReport
	for rows.Next() {
		var result struct {
			ID         string     `json:"id"`
			TenantID   string     `json:"tenant_id"`
			Name       string     `json:"name"`
			TemplateID string     `json:"template_id"`
			Schedule   string     `json:"schedule"`
			Filters    string     `json:"filters"`
			Recipients string     `json:"recipients"`
			IsActive   bool       `json:"is_active"`
			LastRunAt  *time.Time `json:"last_run_at"`
			NextRunAt  *time.Time `json:"next_run_at"`
			CreatedAt  time.Time  `json:"created_at"`
			UpdatedAt  time.Time  `json:"updated_at"`
			CreatedBy  string     `json:"created_by"`
			Metadata   string     `json:"metadata"`
		}

		err := rows.Scan(
			&result.ID, &result.TenantID, &result.Name, &result.TemplateID,
			&result.Schedule, &result.Filters, &result.Recipients, &result.IsActive,
			&result.LastRunAt, &result.NextRunAt, &result.CreatedAt, &result.UpdatedAt,
			&result.CreatedBy, &result.Metadata)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		scheduledReport, err := r.convertToScheduledReport(&result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert result: %w", err)
		}

		scheduledReports = append(scheduledReports, scheduledReport)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	r.logger.Debug("Scheduled reports listed successfully",
		zap.String("tenant_id", filter.TenantID),
		zap.Int("count", len(scheduledReports)))

	return scheduledReports, nil
}

// UpdateScheduledReport updates a scheduled report
func (r *SQLScheduledReportRepository) UpdateScheduledReport(ctx context.Context, scheduledReport *ScheduledReport) error {
	return r.SaveScheduledReport(ctx, scheduledReport) // SaveScheduledReport handles both insert and update
}

// DeleteScheduledReport deletes a scheduled report
func (r *SQLScheduledReportRepository) DeleteScheduledReport(ctx context.Context, tenantID, scheduledReportID string) error {
	r.logger.Info("Deleting scheduled report",
		zap.String("scheduled_report_id", scheduledReportID),
		zap.String("tenant_id", tenantID))

	_, err := r.db.ExecContext(ctx, "DELETE FROM scheduled_reports WHERE id = $1 AND tenant_id = $2", scheduledReportID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete scheduled report: %w", err)
	}

	r.logger.Info("Scheduled report deleted successfully",
		zap.String("scheduled_report_id", scheduledReportID),
		zap.String("tenant_id", tenantID))

	return nil
}

// GetScheduledReportsToRun returns scheduled reports that are due to run
func (r *SQLScheduledReportRepository) GetScheduledReportsToRun(ctx context.Context) ([]*ScheduledReport, error) {
	r.logger.Debug("Getting scheduled reports to run")

	query := `
		SELECT id, tenant_id, name, template_id, schedule, filters,
			   recipients, is_active, last_run_at, next_run_at,
			   created_at, updated_at, created_by, metadata
		FROM scheduled_reports
		WHERE is_active = TRUE AND next_run_at <= NOW()
		ORDER BY next_run_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var scheduledReports []*ScheduledReport
	for rows.Next() {
		var result struct {
			ID         string     `json:"id"`
			TenantID   string     `json:"tenant_id"`
			Name       string     `json:"name"`
			TemplateID string     `json:"template_id"`
			Schedule   string     `json:"schedule"`
			Filters    string     `json:"filters"`
			Recipients string     `json:"recipients"`
			IsActive   bool       `json:"is_active"`
			LastRunAt  *time.Time `json:"last_run_at"`
			NextRunAt  *time.Time `json:"next_run_at"`
			CreatedAt  time.Time  `json:"created_at"`
			UpdatedAt  time.Time  `json:"updated_at"`
			CreatedBy  string     `json:"created_by"`
			Metadata   string     `json:"metadata"`
		}

		err := rows.Scan(
			&result.ID, &result.TenantID, &result.Name, &result.TemplateID,
			&result.Schedule, &result.Filters, &result.Recipients, &result.IsActive,
			&result.LastRunAt, &result.NextRunAt, &result.CreatedAt, &result.UpdatedAt,
			&result.CreatedBy, &result.Metadata)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		scheduledReport, err := r.convertToScheduledReport(&result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert result: %w", err)
		}

		scheduledReports = append(scheduledReports, scheduledReport)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	r.logger.Debug("Scheduled reports to run retrieved successfully",
		zap.Int("count", len(scheduledReports)))

	return scheduledReports, nil
}

// UpdateScheduledReportLastRun updates the last run time and next run time for a scheduled report
func (r *SQLScheduledReportRepository) UpdateScheduledReportLastRun(ctx context.Context, tenantID, scheduledReportID string, lastRunAt time.Time, nextRunAt *time.Time) error {
	r.logger.Debug("Updating scheduled report last run",
		zap.String("scheduled_report_id", scheduledReportID),
		zap.String("tenant_id", tenantID))

	query := `
		UPDATE scheduled_reports SET
			last_run_at = $1, next_run_at = $2, updated_at = $3
		WHERE id = $4 AND tenant_id = $5
	`

	_, err := r.db.ExecContext(ctx, query, lastRunAt, nextRunAt, time.Now(), scheduledReportID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update scheduled report last run: %w", err)
	}

	r.logger.Debug("Scheduled report last run updated successfully",
		zap.String("scheduled_report_id", scheduledReportID),
		zap.String("tenant_id", tenantID))

	return nil
}

// Helper methods

func (r *SQLScheduledReportRepository) convertToScheduledReport(result *struct {
	ID         string     `json:"id"`
	TenantID   string     `json:"tenant_id"`
	Name       string     `json:"name"`
	TemplateID string     `json:"template_id"`
	Schedule   string     `json:"schedule"`
	Filters    string     `json:"filters"`
	Recipients string     `json:"recipients"`
	IsActive   bool       `json:"is_active"`
	LastRunAt  *time.Time `json:"last_run_at"`
	NextRunAt  *time.Time `json:"next_run_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	CreatedBy  string     `json:"created_by"`
	Metadata   string     `json:"metadata"`
}) (*ScheduledReport, error) {
	// Parse schedule JSON
	var schedule ReportSchedule
	if err := json.Unmarshal([]byte(result.Schedule), &schedule); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schedule: %w", err)
	}

	// Parse filters JSON
	var filters ReportFilters
	if result.Filters != "" {
		if err := json.Unmarshal([]byte(result.Filters), &filters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal filters: %w", err)
		}
	}

	// Parse recipients JSON
	var recipients []ReportRecipient
	if result.Recipients != "" {
		if err := json.Unmarshal([]byte(result.Recipients), &recipients); err != nil {
			return nil, fmt.Errorf("failed to unmarshal recipients: %w", err)
		}
	}

	// Parse metadata JSON
	var metadata map[string]interface{}
	if result.Metadata != "" {
		if err := json.Unmarshal([]byte(result.Metadata), &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &ScheduledReport{
		ID:         result.ID,
		TenantID:   result.TenantID,
		Name:       result.Name,
		TemplateID: result.TemplateID,
		Schedule:   schedule,
		Filters:    filters,
		Recipients: recipients,
		IsActive:   result.IsActive,
		LastRunAt:  result.LastRunAt,
		NextRunAt:  result.NextRunAt,
		CreatedAt:  result.CreatedAt,
		UpdatedAt:  result.UpdatedAt,
		CreatedBy:  result.CreatedBy,
		Metadata:   metadata,
	}, nil
}
