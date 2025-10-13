package reporting

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// SQLReportRepository implements ReportRepository using SQL database
type SQLReportRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewSQLReportRepository creates a new SQL report repository
func NewSQLReportRepository(db *sql.DB, logger *zap.Logger) *SQLReportRepository {
	return &SQLReportRepository{
		db:     db,
		logger: logger,
	}
}

// SaveReport saves a report to the database
func (r *SQLReportRepository) SaveReport(ctx context.Context, report *Report) error {
	r.logger.Info("Saving report",
		zap.String("report_id", report.ID),
		zap.String("tenant_id", report.TenantID),
		zap.String("name", report.Name))

	// Convert complex fields to JSON
	dataJSON, err := json.Marshal(report.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	filtersJSON, err := json.Marshal(report.Filters)
	if err != nil {
		return fmt.Errorf("failed to marshal filters: %w", err)
	}

	metadataJSON, err := json.Marshal(report.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Check if report exists
	existingReport, err := r.GetReport(ctx, report.TenantID, report.ID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing report: %w", err)
	}

	if existingReport != nil {
		// Update existing report
		query := `
			UPDATE reports SET
				name = $1, type = $2, status = $3, format = $4, template_id = $5,
				data = $6, filters = $7, generated_at = $8, expires_at = $9,
				file_size = $10, download_url = $11, updated_at = $12, metadata = $13, error = $14
			WHERE id = $15 AND tenant_id = $16
		`
		_, err = r.db.ExecContext(ctx, query,
			report.Name, report.Type, report.Status, report.Format, report.TemplateID,
			dataJSON, filtersJSON, report.GeneratedAt, report.ExpiresAt,
			report.FileSize, report.DownloadURL, report.UpdatedAt, metadataJSON, report.Error,
			report.ID, report.TenantID)
	} else {
		// Insert new report
		query := `
			INSERT INTO reports (
				id, tenant_id, name, type, status, format, template_id, data, filters,
				generated_at, expires_at, file_size, download_url, created_by, created_at, updated_at, metadata, error
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
			)
		`
		_, err = r.db.ExecContext(ctx, query,
			report.ID, report.TenantID, report.Name, report.Type, report.Status, report.Format,
			report.TemplateID, dataJSON, filtersJSON, report.GeneratedAt, report.ExpiresAt,
			report.FileSize, report.DownloadURL, report.CreatedBy, report.CreatedAt, report.UpdatedAt,
			metadataJSON, report.Error)
	}

	if err != nil {
		return fmt.Errorf("failed to save report: %w", err)
	}

	r.logger.Info("Report saved successfully",
		zap.String("report_id", report.ID),
		zap.String("tenant_id", report.TenantID))

	return nil
}

// GetReport retrieves a report by ID
func (r *SQLReportRepository) GetReport(ctx context.Context, tenantID, reportID string) (*Report, error) {
	r.logger.Debug("Getting report",
		zap.String("report_id", reportID),
		zap.String("tenant_id", tenantID))

	query := `
		SELECT id, tenant_id, name, type, status, format, template_id, data, filters,
			   generated_at, expires_at, file_size, download_url, created_by, created_at, updated_at, metadata, error
		FROM reports
		WHERE id = $1 AND tenant_id = $2
	`

	var result struct {
		ID          string     `json:"id"`
		TenantID    string     `json:"tenant_id"`
		Name        string     `json:"name"`
		Type        string     `json:"type"`
		Status      string     `json:"status"`
		Format      string     `json:"format"`
		TemplateID  string     `json:"template_id"`
		Data        string     `json:"data"`
		Filters     string     `json:"filters"`
		GeneratedAt *time.Time `json:"generated_at"`
		ExpiresAt   *time.Time `json:"expires_at"`
		FileSize    int64      `json:"file_size"`
		DownloadURL string     `json:"download_url"`
		CreatedBy   string     `json:"created_by"`
		CreatedAt   time.Time  `json:"created_at"`
		UpdatedAt   time.Time  `json:"updated_at"`
		Metadata    string     `json:"metadata"`
		Error       string     `json:"error"`
	}

	err := r.db.QueryRowContext(ctx, query, reportID, tenantID).Scan(
		&result.ID, &result.TenantID, &result.Name, &result.Type, &result.Status, &result.Format,
		&result.TemplateID, &result.Data, &result.Filters, &result.GeneratedAt, &result.ExpiresAt,
		&result.FileSize, &result.DownloadURL, &result.CreatedBy, &result.CreatedAt, &result.UpdatedAt,
		&result.Metadata, &result.Error)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Report not found
		}
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	// Convert the result to Report
	report, err := r.convertToReport(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert result: %w", err)
	}

	r.logger.Debug("Report retrieved successfully",
		zap.String("report_id", reportID),
		zap.String("tenant_id", tenantID))

	return report, nil
}

// ListReports lists reports with filters
func (r *SQLReportRepository) ListReports(ctx context.Context, filter *ReportFilter) ([]*Report, error) {
	r.logger.Debug("Listing reports",
		zap.String("tenant_id", filter.TenantID),
		zap.String("type", string(filter.Type)))

	// Build query with filters
	query := `
		SELECT id, tenant_id, name, type, status, format, template_id, data, filters,
			   generated_at, expires_at, file_size, download_url, created_by, created_at, updated_at, metadata, error
		FROM reports
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

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, string(filter.Status))
		argIndex++
	}

	if filter.Format != "" {
		query += fmt.Sprintf(" AND format = $%d", argIndex)
		args = append(args, string(filter.Format))
		argIndex++
	}

	if filter.CreatedBy != "" {
		query += fmt.Sprintf(" AND created_by = $%d", argIndex)
		args = append(args, filter.CreatedBy)
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
		return nil, fmt.Errorf("failed to list reports: %w", err)
	}
	defer rows.Close()

	var reports []*Report
	for rows.Next() {
		var result struct {
			ID          string     `json:"id"`
			TenantID    string     `json:"tenant_id"`
			Name        string     `json:"name"`
			Type        string     `json:"type"`
			Status      string     `json:"status"`
			Format      string     `json:"format"`
			TemplateID  string     `json:"template_id"`
			Data        string     `json:"data"`
			Filters     string     `json:"filters"`
			GeneratedAt *time.Time `json:"generated_at"`
			ExpiresAt   *time.Time `json:"expires_at"`
			FileSize    int64      `json:"file_size"`
			DownloadURL string     `json:"download_url"`
			CreatedBy   string     `json:"created_by"`
			CreatedAt   time.Time  `json:"created_at"`
			UpdatedAt   time.Time  `json:"updated_at"`
			Metadata    string     `json:"metadata"`
			Error       string     `json:"error"`
		}

		err := rows.Scan(
			&result.ID, &result.TenantID, &result.Name, &result.Type, &result.Status, &result.Format,
			&result.TemplateID, &result.Data, &result.Filters, &result.GeneratedAt, &result.ExpiresAt,
			&result.FileSize, &result.DownloadURL, &result.CreatedBy, &result.CreatedAt, &result.UpdatedAt,
			&result.Metadata, &result.Error)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		report, err := r.convertToReport(&result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert result: %w", err)
		}
		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	r.logger.Debug("Reports listed successfully",
		zap.String("tenant_id", filter.TenantID),
		zap.Int("count", len(reports)))

	return reports, nil
}

// DeleteReport deletes a report
func (r *SQLReportRepository) DeleteReport(ctx context.Context, tenantID, reportID string) error {
	r.logger.Info("Deleting report",
		zap.String("report_id", reportID),
		zap.String("tenant_id", tenantID))

	_, err := r.db.ExecContext(ctx, "DELETE FROM reports WHERE id = $1 AND tenant_id = $2", reportID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	r.logger.Info("Report deleted successfully",
		zap.String("report_id", reportID),
		zap.String("tenant_id", tenantID))

	return nil
}

// UpdateReportStatus updates the status of a report
func (r *SQLReportRepository) UpdateReportStatus(ctx context.Context, tenantID, reportID string, status ReportStatus, errorMsg string) error {
	r.logger.Debug("Updating report status",
		zap.String("report_id", reportID),
		zap.String("status", string(status)))

	query := `
		UPDATE reports SET
			status = $1, error = $2, updated_at = $3
		WHERE id = $4 AND tenant_id = $5
	`

	_, err := r.db.ExecContext(ctx, query, status, errorMsg, time.Now(), reportID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update report status: %w", err)
	}

	r.logger.Debug("Report status updated successfully",
		zap.String("report_id", reportID),
		zap.String("status", string(status)))

	return nil
}

// UpdateReportFile updates the file information of a report
func (r *SQLReportRepository) UpdateReportFile(ctx context.Context, tenantID, reportID string, fileSize int64, downloadURL string) error {
	r.logger.Debug("Updating report file",
		zap.String("report_id", reportID),
		zap.Int64("file_size", fileSize))

	query := `
		UPDATE reports SET
			file_size = $1, download_url = $2, updated_at = $3
		WHERE id = $4 AND tenant_id = $5
	`

	_, err := r.db.ExecContext(ctx, query, fileSize, downloadURL, time.Now(), reportID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update report file: %w", err)
	}

	r.logger.Debug("Report file updated successfully",
		zap.String("report_id", reportID),
		zap.Int64("file_size", fileSize))

	return nil
}

// GetReportMetrics gets report usage metrics
func (r *SQLReportRepository) GetReportMetrics(ctx context.Context, tenantID string) (*ReportMetrics, error) {
	r.logger.Debug("Getting report metrics",
		zap.String("tenant_id", tenantID))

	// Get basic metrics
	query := `
		SELECT 
			COUNT(*) as total_reports,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_reports,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_reports,
			COUNT(CASE WHEN status = 'generating' THEN 1 END) as generating_reports,
			COALESCE(SUM(file_size), 0) as total_file_size,
			COALESCE(AVG(file_size), 0) as average_file_size
		FROM reports
		WHERE tenant_id = $1
	`

	var metrics ReportMetrics
	var completedReports, failedReports, generatingReports int

	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&metrics.TotalReports, &completedReports, &failedReports, &generatingReports,
		&metrics.TotalFileSize, &metrics.AverageFileSize)

	if err != nil {
		return nil, fmt.Errorf("failed to get report metrics: %w", err)
	}

	// Get reports by type
	typeQuery := `
		SELECT type, COUNT(*) as count
		FROM reports
		WHERE tenant_id = $1
		GROUP BY type
	`

	rows, err := r.db.QueryContext(ctx, typeQuery, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reports by type: %w", err)
	}
	defer rows.Close()

	metrics.ReportsByType = make(map[ReportType]int)
	for rows.Next() {
		var reportType string
		var count int
		if err := rows.Scan(&reportType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan type row: %w", err)
		}
		metrics.ReportsByType[ReportType(reportType)] = count
	}

	// Get reports by status
	metrics.ReportsByStatus = map[ReportStatus]int{
		ReportStatusCompleted:  completedReports,
		ReportStatusFailed:     failedReports,
		ReportStatusGenerating: generatingReports,
	}

	// Get reports by format
	formatQuery := `
		SELECT format, COUNT(*) as count
		FROM reports
		WHERE tenant_id = $1
		GROUP BY format
	`

	rows, err = r.db.QueryContext(ctx, formatQuery, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reports by format: %w", err)
	}
	defer rows.Close()

	metrics.ReportsByFormat = make(map[ReportFormat]int)
	for rows.Next() {
		var format string
		var count int
		if err := rows.Scan(&format, &count); err != nil {
			return nil, fmt.Errorf("failed to scan format row: %w", err)
		}
		metrics.ReportsByFormat[ReportFormat(format)] = count
	}

	// Get most used templates
	templateQuery := `
		SELECT template_id, COUNT(*) as usage_count, MAX(created_at) as last_used
		FROM reports
		WHERE tenant_id = $1 AND template_id IS NOT NULL AND template_id != ''
		GROUP BY template_id
		ORDER BY usage_count DESC
		LIMIT 10
	`

	rows, err = r.db.QueryContext(ctx, templateQuery, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template usage: %w", err)
	}
	defer rows.Close()

	metrics.MostUsedTemplates = []TemplateUsageData{}
	for rows.Next() {
		var templateID string
		var usageCount int
		var lastUsed time.Time
		if err := rows.Scan(&templateID, &usageCount, &lastUsed); err != nil {
			return nil, fmt.Errorf("failed to scan template row: %w", err)
		}
		metrics.MostUsedTemplates = append(metrics.MostUsedTemplates, TemplateUsageData{
			TemplateID:   templateID,
			TemplateName: templateID, // This would be looked up from templates table
			UsageCount:   usageCount,
			LastUsed:     lastUsed,
		})
	}

	// Get generation time metrics
	timeQuery := `
		SELECT 
			COALESCE(AVG(EXTRACT(EPOCH FROM (generated_at - created_at))), 0) as average_seconds,
			COALESCE(MIN(EXTRACT(EPOCH FROM (generated_at - created_at))), 0) as min_seconds,
			COALESCE(MAX(EXTRACT(EPOCH FROM (generated_at - created_at))), 0) as max_seconds
		FROM reports
		WHERE tenant_id = $1 AND generated_at IS NOT NULL
	`

	err = r.db.QueryRowContext(ctx, timeQuery, tenantID).Scan(
		&metrics.GenerationTime.Average,
		&metrics.GenerationTime.Min,
		&metrics.GenerationTime.Max)

	if err != nil {
		return nil, fmt.Errorf("failed to get generation time metrics: %w", err)
	}

	// Calculate P95 and P99 (simplified)
	metrics.GenerationTime.P95 = metrics.GenerationTime.Average * 1.5
	metrics.GenerationTime.P99 = metrics.GenerationTime.Average * 2.0

	r.logger.Debug("Report metrics retrieved",
		zap.String("tenant_id", tenantID),
		zap.Int("total_reports", metrics.TotalReports))

	return &metrics, nil
}

// convertToReport converts a database result to Report
func (r *SQLReportRepository) convertToReport(result interface{}) (*Report, error) {
	// Type assertion to get the fields
	var id, tenantID, name, reportType, status, format, templateID, data, filters, metadata, errorMsg string
	var generatedAt, expiresAt *time.Time
	var fileSize int64
	var downloadURL, createdBy string
	var createdAt, updatedAt time.Time

	// Use reflection or type assertion based on the actual structure
	switch v := result.(type) {
	case *struct {
		ID          string     `json:"id"`
		TenantID    string     `json:"tenant_id"`
		Name        string     `json:"name"`
		Type        string     `json:"type"`
		Status      string     `json:"status"`
		Format      string     `json:"format"`
		TemplateID  string     `json:"template_id"`
		Data        string     `json:"data"`
		Filters     string     `json:"filters"`
		GeneratedAt *time.Time `json:"generated_at"`
		ExpiresAt   *time.Time `json:"expires_at"`
		FileSize    int64      `json:"file_size"`
		DownloadURL string     `json:"download_url"`
		CreatedBy   string     `json:"created_by"`
		CreatedAt   time.Time  `json:"created_at"`
		UpdatedAt   time.Time  `json:"updated_at"`
		Metadata    string     `json:"metadata"`
		Error       string     `json:"error"`
	}:
		id = v.ID
		tenantID = v.TenantID
		name = v.Name
		reportType = v.Type
		status = v.Status
		format = v.Format
		templateID = v.TemplateID
		data = v.Data
		filters = v.Filters
		generatedAt = v.GeneratedAt
		expiresAt = v.ExpiresAt
		fileSize = v.FileSize
		downloadURL = v.DownloadURL
		createdBy = v.CreatedBy
		createdAt = v.CreatedAt
		updatedAt = v.UpdatedAt
		metadata = v.Metadata
		errorMsg = v.Error
	default:
		return nil, fmt.Errorf("unsupported result type: %T", result)
	}

	// Parse JSON fields
	var dataObj ReportData
	if data != "" {
		if err := json.Unmarshal([]byte(data), &dataObj); err != nil {
			return nil, fmt.Errorf("failed to unmarshal data: %w", err)
		}
	}

	var filtersObj ReportFilters
	if filters != "" {
		if err := json.Unmarshal([]byte(filters), &filtersObj); err != nil {
			return nil, fmt.Errorf("failed to unmarshal filters: %w", err)
		}
	}

	var metadataMap map[string]interface{}
	if metadata != "" {
		if err := json.Unmarshal([]byte(metadata), &metadataMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	report := &Report{
		ID:          id,
		TenantID:    tenantID,
		Name:        name,
		Type:        ReportType(reportType),
		Status:      ReportStatus(status),
		Format:      ReportFormat(format),
		TemplateID:  templateID,
		Data:        dataObj,
		Filters:     filtersObj,
		GeneratedAt: generatedAt,
		ExpiresAt:   expiresAt,
		FileSize:    fileSize,
		DownloadURL: downloadURL,
		CreatedBy:   createdBy,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Metadata:    metadataMap,
		Error:       errorMsg,
	}

	return report, nil
}
