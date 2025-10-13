package reporting

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// SQLReportTemplateRepository implements ReportTemplateRepository using SQL database
type SQLReportTemplateRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewSQLReportTemplateRepository creates a new SQL report template repository
func NewSQLReportTemplateRepository(db *sql.DB, logger *zap.Logger) *SQLReportTemplateRepository {
	return &SQLReportTemplateRepository{
		db:     db,
		logger: logger,
	}
}

// SaveTemplate saves a report template to the database
func (r *SQLReportTemplateRepository) SaveTemplate(ctx context.Context, template *ReportTemplate) error {
	r.logger.Info("Saving report template",
		zap.String("template_id", template.ID),
		zap.String("tenant_id", template.TenantID),
		zap.String("name", template.Name))

	// Convert complex fields to JSON
	templateJSON, err := json.Marshal(template.Template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	metadataJSON, err := json.Marshal(template.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Check if template exists
	existingTemplate, err := r.GetTemplate(ctx, template.TenantID, template.ID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing template: %w", err)
	}

	if existingTemplate != nil {
		// Update existing template
		query := `
			UPDATE report_templates SET
				name = $1, type = $2, description = $3, template = $4,
				is_public = $5, is_default = $6, updated_at = $7, metadata = $8
			WHERE id = $9 AND tenant_id = $10
		`
		_, err = r.db.ExecContext(ctx, query,
			template.Name, template.Type, template.Description, templateJSON,
			template.IsPublic, template.IsDefault, time.Now(), metadataJSON,
			template.ID, template.TenantID)
	} else {
		// Insert new template
		query := `
			INSERT INTO report_templates (
				id, tenant_id, name, type, description, template,
				is_public, is_default, created_at, updated_at, created_by, metadata
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
			)
		`
		_, err = r.db.ExecContext(ctx, query,
			template.ID, template.TenantID, template.Name, template.Type,
			template.Description, templateJSON, template.IsPublic, template.IsDefault,
			template.CreatedAt, template.UpdatedAt, template.CreatedBy, metadataJSON)
	}

	if err != nil {
		return fmt.Errorf("failed to save template: %w", err)
	}

	r.logger.Info("Report template saved successfully",
		zap.String("template_id", template.ID),
		zap.String("tenant_id", template.TenantID))

	return nil
}

// GetTemplate retrieves a report template by ID
func (r *SQLReportTemplateRepository) GetTemplate(ctx context.Context, tenantID, templateID string) (*ReportTemplate, error) {
	r.logger.Debug("Getting report template",
		zap.String("template_id", templateID),
		zap.String("tenant_id", tenantID))

	query := `
		SELECT id, tenant_id, name, type, description, template,
			   is_public, is_default, created_at, updated_at, created_by, metadata
		FROM report_templates
		WHERE id = $1 AND tenant_id = $2
	`

	var result struct {
		ID          string    `json:"id"`
		TenantID    string    `json:"tenant_id"`
		Name        string    `json:"name"`
		Type        string    `json:"type"`
		Description string    `json:"description"`
		Template    string    `json:"template"`
		IsPublic    bool      `json:"is_public"`
		IsDefault   bool      `json:"is_default"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		CreatedBy   string    `json:"created_by"`
		Metadata    string    `json:"metadata"`
	}

	err := r.db.QueryRowContext(ctx, query, templateID, tenantID).Scan(
		&result.ID, &result.TenantID, &result.Name, &result.Type,
		&result.Description, &result.Template, &result.IsPublic, &result.IsDefault,
		&result.CreatedAt, &result.UpdatedAt, &result.CreatedBy, &result.Metadata)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Template not found
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Convert the result to ReportTemplate
	template, err := r.convertToReportTemplate(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert result: %w", err)
	}

	r.logger.Debug("Report template retrieved successfully",
		zap.String("template_id", templateID),
		zap.String("tenant_id", tenantID))

	return template, nil
}

// ListTemplates lists report templates with filters
func (r *SQLReportTemplateRepository) ListTemplates(ctx context.Context, filter *ReportTemplateFilter) ([]*ReportTemplate, error) {
	r.logger.Debug("Listing report templates",
		zap.String("tenant_id", filter.TenantID),
		zap.String("type", string(filter.Type)))

	// Build query with filters
	query := `
		SELECT id, tenant_id, name, type, description, template,
			   is_public, is_default, created_at, updated_at, created_by, metadata
		FROM report_templates
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

	if filter.IsPublic != nil {
		query += fmt.Sprintf(" AND is_public = $%d", argIndex)
		args = append(args, *filter.IsPublic)
		argIndex++
	}

	if filter.IsDefault != nil {
		query += fmt.Sprintf(" AND is_default = $%d", argIndex)
		args = append(args, *filter.IsDefault)
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

	var templates []*ReportTemplate
	for rows.Next() {
		var result struct {
			ID          string    `json:"id"`
			TenantID    string    `json:"tenant_id"`
			Name        string    `json:"name"`
			Type        string    `json:"type"`
			Description string    `json:"description"`
			Template    string    `json:"template"`
			IsPublic    bool      `json:"is_public"`
			IsDefault   bool      `json:"is_default"`
			CreatedAt   time.Time `json:"created_at"`
			UpdatedAt   time.Time `json:"updated_at"`
			CreatedBy   string    `json:"created_by"`
			Metadata    string    `json:"metadata"`
		}

		err := rows.Scan(
			&result.ID, &result.TenantID, &result.Name, &result.Type,
			&result.Description, &result.Template, &result.IsPublic, &result.IsDefault,
			&result.CreatedAt, &result.UpdatedAt, &result.CreatedBy, &result.Metadata)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		template, err := r.convertToReportTemplate(&result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert result: %w", err)
		}

		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	r.logger.Debug("Report templates listed successfully",
		zap.String("tenant_id", filter.TenantID),
		zap.Int("count", len(templates)))

	return templates, nil
}

// UpdateTemplate updates a report template
func (r *SQLReportTemplateRepository) UpdateTemplate(ctx context.Context, template *ReportTemplate) error {
	return r.SaveTemplate(ctx, template) // SaveTemplate handles both insert and update
}

// DeleteTemplate deletes a report template
func (r *SQLReportTemplateRepository) DeleteTemplate(ctx context.Context, tenantID, templateID string) error {
	r.logger.Info("Deleting report template",
		zap.String("template_id", templateID),
		zap.String("tenant_id", tenantID))

	_, err := r.db.ExecContext(ctx, "DELETE FROM report_templates WHERE id = $1 AND tenant_id = $2", templateID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	r.logger.Info("Report template deleted successfully",
		zap.String("template_id", templateID),
		zap.String("tenant_id", tenantID))

	return nil
}

// Helper methods

func (r *SQLReportTemplateRepository) convertToReportTemplate(result *struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Template    string    `json:"template"`
	IsPublic    bool      `json:"is_public"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   string    `json:"created_by"`
	Metadata    string    `json:"metadata"`
}) (*ReportTemplate, error) {
	// Parse template JSON
	var templateConfig ReportTemplateConfig
	if err := json.Unmarshal([]byte(result.Template), &templateConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template: %w", err)
	}

	// Parse metadata JSON
	var metadata map[string]interface{}
	if result.Metadata != "" {
		if err := json.Unmarshal([]byte(result.Metadata), &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &ReportTemplate{
		ID:          result.ID,
		TenantID:    result.TenantID,
		Name:        result.Name,
		Type:        ReportType(result.Type),
		Description: result.Description,
		Template:    templateConfig,
		IsPublic:    result.IsPublic,
		IsDefault:   result.IsDefault,
		CreatedAt:   result.CreatedAt,
		UpdatedAt:   result.UpdatedAt,
		CreatedBy:   result.CreatedBy,
		Metadata:    metadata,
	}, nil
}
