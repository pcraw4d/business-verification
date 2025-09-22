package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"kyb-platform/internal/observability"
	"kyb-platform/internal/services"
)

// UnifiedComplianceRepository implements the UnifiedComplianceRepository interface
type UnifiedComplianceRepository struct {
	db     *sql.DB
	logger *observability.Logger
}

// NewUnifiedComplianceRepository creates a new unified compliance repository
func NewUnifiedComplianceRepository(db *sql.DB, logger *observability.Logger) *UnifiedComplianceRepository {
	return &UnifiedComplianceRepository{
		db:     db,
		logger: logger,
	}
}

// SaveComplianceTracking saves a compliance tracking record
func (ucr *UnifiedComplianceRepository) SaveComplianceTracking(ctx context.Context, tracking *services.ComplianceTracking) error {
	query := `
		INSERT INTO compliance_tracking (
			id, merchant_id, compliance_type, compliance_framework, check_type,
			status, score, risk_level, requirements, check_method, source,
			raw_data, result, findings, recommendations, evidence,
			checked_by, checked_at, reviewed_by, reviewed_at, approved_by, approved_at,
			due_date, expires_at, next_review_date, priority, assigned_to,
			tags, notes, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32
		)`

	_, err := ucr.db.ExecContext(ctx, query,
		tracking.ID,
		tracking.MerchantID,
		tracking.ComplianceType,
		tracking.ComplianceFramework,
		tracking.CheckType,
		tracking.Status,
		tracking.Score,
		tracking.RiskLevel,
		toJSONB(tracking.Requirements),
		tracking.CheckMethod,
		tracking.Source,
		toJSONB(tracking.RawData),
		toJSONB(tracking.Result),
		toJSONB(tracking.Findings),
		toJSONB(tracking.Recommendations),
		toJSONB(tracking.Evidence),
		tracking.CheckedBy,
		tracking.CheckedAt,
		tracking.ReviewedBy,
		tracking.ReviewedAt,
		tracking.ApprovedBy,
		tracking.ApprovedAt,
		tracking.DueDate,
		tracking.ExpiresAt,
		tracking.NextReviewDate,
		tracking.Priority,
		tracking.AssignedTo,
		toStringArray(tracking.Tags),
		tracking.Notes,
		toJSONB(tracking.Metadata),
		tracking.CreatedAt,
		tracking.UpdatedAt,
	)

	if err != nil {
		ucr.logger.Error("failed to save compliance tracking", map[string]interface{}{
			"error":       err.Error(),
			"tracking_id": tracking.ID,
		})
		return fmt.Errorf("failed to save compliance tracking: %w", err)
	}

	return nil
}

// GetComplianceTracking retrieves compliance tracking records with filtering
func (ucr *UnifiedComplianceRepository) GetComplianceTracking(ctx context.Context, filters *services.ComplianceTrackingFilters) ([]*services.ComplianceTracking, error) {
	query, args := ucr.buildComplianceTrackingQuery(filters)

	rows, err := ucr.db.QueryContext(ctx, query, args...)
	if err != nil {
		ucr.logger.Error("failed to query compliance tracking", map[string]interface{}{
			"error":   err.Error(),
			"filters": filters,
		})
		return nil, fmt.Errorf("failed to query compliance tracking: %w", err)
	}
	defer rows.Close()

	var records []*services.ComplianceTracking
	for rows.Next() {
		record, err := ucr.scanComplianceTracking(rows)
		if err != nil {
			ucr.logger.Error("failed to scan compliance tracking", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, fmt.Errorf("failed to scan compliance tracking: %w", err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		ucr.logger.Error("error iterating compliance tracking rows", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("error iterating compliance tracking rows: %w", err)
	}

	return records, nil
}

// GetComplianceTrackingByID retrieves a specific compliance tracking record by ID
func (ucr *UnifiedComplianceRepository) GetComplianceTrackingByID(ctx context.Context, id string) (*services.ComplianceTracking, error) {
	query := `
		SELECT id, merchant_id, compliance_type, compliance_framework, check_type,
			   status, score, risk_level, requirements, check_method, source,
			   raw_data, result, findings, recommendations, evidence,
			   checked_by, checked_at, reviewed_by, reviewed_at, approved_by, approved_at,
			   due_date, expires_at, next_review_date, priority, assigned_to,
			   tags, notes, metadata, created_at, updated_at
		FROM compliance_tracking
		WHERE id = $1`

	row := ucr.db.QueryRowContext(ctx, query, id)

	record, err := ucr.scanComplianceTrackingRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("compliance tracking record not found: %s", id)
		}
		ucr.logger.Error("failed to get compliance tracking by ID", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return nil, fmt.Errorf("failed to get compliance tracking by ID: %w", err)
	}

	return record, nil
}

// UpdateComplianceTracking updates an existing compliance tracking record
func (ucr *UnifiedComplianceRepository) UpdateComplianceTracking(ctx context.Context, tracking *services.ComplianceTracking) error {
	query := `
		UPDATE compliance_tracking SET
			status = $2, score = $3, risk_level = $4, requirements = $5,
			result = $6, findings = $7, recommendations = $8, evidence = $9,
			reviewed_by = $10, reviewed_at = $11, approved_by = $12, approved_at = $13,
			due_date = $14, expires_at = $15, next_review_date = $16, priority = $17,
			assigned_to = $18, tags = $19, notes = $20, metadata = $21, updated_at = $22
		WHERE id = $1`

	_, err := ucr.db.ExecContext(ctx, query,
		tracking.ID,
		tracking.Status,
		tracking.Score,
		tracking.RiskLevel,
		toJSONB(tracking.Requirements),
		toJSONB(tracking.Result),
		toJSONB(tracking.Findings),
		toJSONB(tracking.Recommendations),
		toJSONB(tracking.Evidence),
		tracking.ReviewedBy,
		tracking.ReviewedAt,
		tracking.ApprovedBy,
		tracking.ApprovedAt,
		tracking.DueDate,
		tracking.ExpiresAt,
		tracking.NextReviewDate,
		tracking.Priority,
		tracking.AssignedTo,
		toStringArray(tracking.Tags),
		tracking.Notes,
		toJSONB(tracking.Metadata),
		tracking.UpdatedAt,
	)

	if err != nil {
		ucr.logger.Error("failed to update compliance tracking", map[string]interface{}{
			"error":       err.Error(),
			"tracking_id": tracking.ID,
		})
		return fmt.Errorf("failed to update compliance tracking: %w", err)
	}

	return nil
}

// DeleteComplianceTracking deletes a compliance tracking record
func (ucr *UnifiedComplianceRepository) DeleteComplianceTracking(ctx context.Context, id string) error {
	query := `DELETE FROM compliance_tracking WHERE id = $1`

	result, err := ucr.db.ExecContext(ctx, query, id)
	if err != nil {
		ucr.logger.Error("failed to delete compliance tracking", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return fmt.Errorf("failed to delete compliance tracking: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("compliance tracking record not found: %s", id)
	}

	return nil
}

// GetMerchantComplianceSummary retrieves compliance summary for a merchant
func (ucr *UnifiedComplianceRepository) GetMerchantComplianceSummary(ctx context.Context, merchantID string) (*services.MerchantComplianceSummary, error) {
	query := `
		SELECT 
			merchant_id,
			COUNT(*) as total_checks,
			COUNT(*) FILTER (WHERE status = 'completed') as completed_checks,
			COUNT(*) FILTER (WHERE status = 'pending') as pending_checks,
			COUNT(*) FILTER (WHERE status = 'failed') as failed_checks,
			COUNT(*) FILTER (WHERE status = 'overdue') as overdue_checks,
			COUNT(*) FILTER (WHERE due_date < CURRENT_TIMESTAMP AND status NOT IN ('completed', 'cancelled')) as past_due_checks,
			AVG(score) as average_score,
			MAX(checked_at) as last_check_date,
			MIN(next_review_date) as next_review_date,
			COUNT(DISTINCT compliance_type) as compliance_types_covered
		FROM compliance_tracking
		WHERE merchant_id = $1
		GROUP BY merchant_id`

	row := ucr.db.QueryRowContext(ctx, query, merchantID)

	var summary services.MerchantComplianceSummary
	var averageScore sql.NullFloat64
	var lastCheckDate sql.NullTime
	var nextReviewDate sql.NullTime

	err := row.Scan(
		&summary.MerchantID,
		&summary.TotalChecks,
		&summary.CompletedChecks,
		&summary.PendingChecks,
		&summary.FailedChecks,
		&summary.OverdueChecks,
		&summary.PastDueChecks,
		&averageScore,
		&lastCheckDate,
		&nextReviewDate,
		&summary.ComplianceTypesCovered,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Return empty summary if no records found
			summary.MerchantID = merchantID
			summary.GeneratedAt = time.Now()
			summary.ComplianceScore = 0.0
			summary.RiskLevel = "unknown"
			return &summary, nil
		}
		ucr.logger.Error("failed to get merchant compliance summary", map[string]interface{}{
			"error":       err.Error(),
			"merchant_id": merchantID,
		})
		return nil, fmt.Errorf("failed to get merchant compliance summary: %w", err)
	}

	// Set nullable fields
	if averageScore.Valid {
		summary.AverageScore = &averageScore.Float64
	}
	if lastCheckDate.Valid {
		summary.LastCheckDate = &lastCheckDate.Time
	}
	if nextReviewDate.Valid {
		summary.NextReviewDate = &nextReviewDate.Time
	}

	// Calculate compliance score
	if summary.TotalChecks > 0 {
		summary.ComplianceScore = float64(summary.CompletedChecks) / float64(summary.TotalChecks)
	} else {
		summary.ComplianceScore = 0.0
	}

	// Determine risk level
	if summary.ComplianceScore >= 0.9 {
		summary.RiskLevel = "low"
	} else if summary.ComplianceScore >= 0.7 {
		summary.RiskLevel = "medium"
	} else if summary.ComplianceScore >= 0.5 {
		summary.RiskLevel = "high"
	} else {
		summary.RiskLevel = "critical"
	}

	summary.GeneratedAt = time.Now()

	return &summary, nil
}

// GetComplianceAlerts retrieves compliance alerts for monitoring
func (ucr *UnifiedComplianceRepository) GetComplianceAlerts(ctx context.Context, filters *services.ComplianceAlertFilters) ([]*services.UnifiedComplianceAlert, error) {
	query, args := ucr.buildComplianceAlertsQuery(filters)

	rows, err := ucr.db.QueryContext(ctx, query, args...)
	if err != nil {
		ucr.logger.Error("failed to query compliance alerts", map[string]interface{}{
			"error":   err.Error(),
			"filters": filters,
		})
		return nil, fmt.Errorf("failed to query compliance alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*services.UnifiedComplianceAlert
	for rows.Next() {
		alert := &services.UnifiedComplianceAlert{}
		err := rows.Scan(
			&alert.ID,
			&alert.MerchantID,
			&alert.ComplianceType,
			&alert.ComplianceFramework,
			&alert.Status,
			&alert.Priority,
			&alert.RiskLevel,
			&alert.AlertType,
			&alert.DueDate,
			&alert.ExpiresAt,
			&alert.AssignedTo,
			&alert.CreatedAt,
			&alert.UpdatedAt,
		)
		if err != nil {
			ucr.logger.Error("failed to scan compliance alert", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, fmt.Errorf("failed to scan compliance alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		ucr.logger.Error("error iterating compliance alert rows", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("error iterating compliance alert rows: %w", err)
	}

	return alerts, nil
}

// GetComplianceTrends retrieves compliance trends for reporting
func (ucr *UnifiedComplianceRepository) GetComplianceTrends(ctx context.Context, filters *services.ComplianceTrendFilters) ([]*services.UnifiedComplianceTrend, error) {
	query, args := ucr.buildComplianceTrendsQuery(filters)

	rows, err := ucr.db.QueryContext(ctx, query, args...)
	if err != nil {
		ucr.logger.Error("failed to query compliance trends", map[string]interface{}{
			"error":   err.Error(),
			"filters": filters,
		})
		return nil, fmt.Errorf("failed to query compliance trends: %w", err)
	}
	defer rows.Close()

	var trends []*services.UnifiedComplianceTrend
	for rows.Next() {
		trend := &services.UnifiedComplianceTrend{}
		var averageScore sql.NullFloat64
		err := rows.Scan(
			&trend.Date,
			&trend.MerchantID,
			&trend.ComplianceType,
			&trend.TotalChecks,
			&trend.CompletedChecks,
			&trend.FailedChecks,
			&averageScore,
			&trend.ComplianceScore,
		)
		if err != nil {
			ucr.logger.Error("failed to scan compliance trend", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, fmt.Errorf("failed to scan compliance trend: %w", err)
		}

		if averageScore.Valid {
			trend.AverageScore = &averageScore.Float64
		}

		trends = append(trends, trend)
	}

	if err := rows.Err(); err != nil {
		ucr.logger.Error("error iterating compliance trend rows", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("error iterating compliance trend rows: %w", err)
	}

	return trends, nil
}

// Helper methods

func (ucr *UnifiedComplianceRepository) buildComplianceTrackingQuery(filters *services.ComplianceTrackingFilters) (string, []interface{}) {
	query := `
		SELECT id, merchant_id, compliance_type, compliance_framework, check_type,
			   status, score, risk_level, requirements, check_method, source,
			   raw_data, result, findings, recommendations, evidence,
			   checked_by, checked_at, reviewed_by, reviewed_at, approved_by, approved_at,
			   due_date, expires_at, next_review_date, priority, assigned_to,
			   tags, notes, metadata, created_at, updated_at
		FROM compliance_tracking
		WHERE 1=1`

	var args []interface{}
	argIndex := 1

	// Add filters
	if filters.MerchantID != "" {
		query += fmt.Sprintf(" AND merchant_id = $%d", argIndex)
		args = append(args, filters.MerchantID)
		argIndex++
	}

	if filters.ComplianceType != "" {
		query += fmt.Sprintf(" AND compliance_type = $%d", argIndex)
		args = append(args, filters.ComplianceType)
		argIndex++
	}

	if filters.ComplianceFramework != "" {
		query += fmt.Sprintf(" AND compliance_framework = $%d", argIndex)
		args = append(args, filters.ComplianceFramework)
		argIndex++
	}

	if filters.CheckType != "" {
		query += fmt.Sprintf(" AND check_type = $%d", argIndex)
		args = append(args, filters.CheckType)
		argIndex++
	}

	if filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filters.Status)
		argIndex++
	}

	if filters.RiskLevel != "" {
		query += fmt.Sprintf(" AND risk_level = $%d", argIndex)
		args = append(args, filters.RiskLevel)
		argIndex++
	}

	if filters.Priority != "" {
		query += fmt.Sprintf(" AND priority = $%d", argIndex)
		args = append(args, filters.Priority)
		argIndex++
	}

	if filters.CheckedBy != "" {
		query += fmt.Sprintf(" AND checked_by = $%d", argIndex)
		args = append(args, filters.CheckedBy)
		argIndex++
	}

	if filters.AssignedTo != "" {
		query += fmt.Sprintf(" AND assigned_to = $%d", argIndex)
		args = append(args, filters.AssignedTo)
		argIndex++
	}

	if filters.DueDateAfter != nil {
		query += fmt.Sprintf(" AND due_date >= $%d", argIndex)
		args = append(args, filters.DueDateAfter)
		argIndex++
	}

	if filters.DueDateBefore != nil {
		query += fmt.Sprintf(" AND due_date <= $%d", argIndex)
		args = append(args, filters.DueDateBefore)
		argIndex++
	}

	if filters.ExpiresAtAfter != nil {
		query += fmt.Sprintf(" AND expires_at >= $%d", argIndex)
		args = append(args, filters.ExpiresAtAfter)
		argIndex++
	}

	if filters.ExpiresAtBefore != nil {
		query += fmt.Sprintf(" AND expires_at <= $%d", argIndex)
		args = append(args, filters.ExpiresAtBefore)
		argIndex++
	}

	if filters.CreatedAtAfter != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, filters.CreatedAtAfter)
		argIndex++
	}

	if filters.CreatedAtBefore != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, filters.CreatedAtBefore)
		argIndex++
	}

	if len(filters.Tags) > 0 {
		query += fmt.Sprintf(" AND tags && $%d", argIndex)
		args = append(args, toStringArray(filters.Tags))
		argIndex++
	}

	if filters.Overdue {
		query += " AND due_date < CURRENT_TIMESTAMP AND status NOT IN ('completed', 'cancelled')"
	}

	if filters.ExpiringSoon {
		query += " AND expires_at < CURRENT_TIMESTAMP + INTERVAL '7 days'"
	}

	// Add ordering
	query += " ORDER BY created_at DESC"

	// Add pagination
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
		argIndex++
	}

	return query, args
}

func (ucr *UnifiedComplianceRepository) buildComplianceAlertsQuery(filters *services.ComplianceAlertFilters) (string, []interface{}) {
	query := `
		SELECT 
			id, merchant_id, compliance_type, compliance_framework, status,
			priority, risk_level, alert_type, due_date, expires_at,
			assigned_to, created_at, updated_at
		FROM compliance_alerts
		WHERE 1=1`

	var args []interface{}
	argIndex := 1

	// Add filters
	if filters.MerchantID != "" {
		query += fmt.Sprintf(" AND merchant_id = $%d", argIndex)
		args = append(args, filters.MerchantID)
		argIndex++
	}

	if filters.ComplianceType != "" {
		query += fmt.Sprintf(" AND compliance_type = $%d", argIndex)
		args = append(args, filters.ComplianceType)
		argIndex++
	}

	if filters.AlertType != "" {
		query += fmt.Sprintf(" AND alert_type = $%d", argIndex)
		args = append(args, filters.AlertType)
		argIndex++
	}

	if filters.Priority != "" {
		query += fmt.Sprintf(" AND priority = $%d", argIndex)
		args = append(args, filters.Priority)
		argIndex++
	}

	if filters.RiskLevel != "" {
		query += fmt.Sprintf(" AND risk_level = $%d", argIndex)
		args = append(args, filters.RiskLevel)
		argIndex++
	}

	if filters.AssignedTo != "" {
		query += fmt.Sprintf(" AND assigned_to = $%d", argIndex)
		args = append(args, filters.AssignedTo)
		argIndex++
	}

	if filters.CreatedAtAfter != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, filters.CreatedAtAfter)
		argIndex++
	}

	if filters.CreatedAtBefore != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, filters.CreatedAtBefore)
		argIndex++
	}

	// Add ordering
	query += " ORDER BY created_at DESC"

	// Add pagination
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
		argIndex++
	}

	return query, args
}

func (ucr *UnifiedComplianceRepository) buildComplianceTrendsQuery(filters *services.ComplianceTrendFilters) (string, []interface{}) {
	// Determine date grouping
	dateGrouping := "DATE(created_at)"
	if filters.GroupBy == "week" {
		dateGrouping = "DATE_TRUNC('week', created_at)"
	} else if filters.GroupBy == "month" {
		dateGrouping = "DATE_TRUNC('month', created_at)"
	}

	query := fmt.Sprintf(`
		SELECT 
			%s as date,
			merchant_id,
			compliance_type,
			COUNT(*) as total_checks,
			COUNT(*) FILTER (WHERE status = 'completed') as completed_checks,
			COUNT(*) FILTER (WHERE status = 'failed') as failed_checks,
			AVG(score) as average_score,
			COUNT(*) FILTER (WHERE status = 'completed')::float / COUNT(*) as compliance_score
		FROM compliance_tracking
		WHERE created_at >= $1 AND created_at <= $2`, dateGrouping)

	var args []interface{}
	args = append(args, filters.StartDate, filters.EndDate)
	argIndex := 3

	// Add filters
	if filters.MerchantID != "" {
		query += fmt.Sprintf(" AND merchant_id = $%d", argIndex)
		args = append(args, filters.MerchantID)
		argIndex++
	}

	if filters.ComplianceType != "" {
		query += fmt.Sprintf(" AND compliance_type = $%d", argIndex)
		args = append(args, filters.ComplianceType)
		argIndex++
	}

	// Add grouping
	query += fmt.Sprintf(" GROUP BY %s, merchant_id, compliance_type", dateGrouping)

	// Add ordering
	query += " ORDER BY date DESC"

	// Add pagination
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
		argIndex++
	}

	return query, args
}

func (ucr *UnifiedComplianceRepository) scanComplianceTracking(rows *sql.Rows) (*services.ComplianceTracking, error) {
	record := &services.ComplianceTracking{}
	var requirements, rawData, result, findings, recommendations, evidence, metadata sql.NullString
	var tags sql.NullString
	var checkedBy, reviewedBy, approvedBy, assignedTo, notes sql.NullString
	var score sql.NullFloat64
	var reviewedAt, approvedAt, dueDate, expiresAt, nextReviewDate sql.NullTime

	err := rows.Scan(
		&record.ID,
		&record.MerchantID,
		&record.ComplianceType,
		&record.ComplianceFramework,
		&record.CheckType,
		&record.Status,
		&score,
		&record.RiskLevel,
		&requirements,
		&record.CheckMethod,
		&record.Source,
		&rawData,
		&result,
		&findings,
		&recommendations,
		&evidence,
		&checkedBy,
		&record.CheckedAt,
		&reviewedBy,
		&reviewedAt,
		&approvedBy,
		&approvedAt,
		&dueDate,
		&expiresAt,
		&nextReviewDate,
		&record.Priority,
		&assignedTo,
		&tags,
		&notes,
		&metadata,
		&record.CreatedAt,
		&record.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Set nullable fields
	if score.Valid {
		record.Score = &score.Float64
	}
	if checkedBy.Valid {
		record.CheckedBy = &checkedBy.String
	}
	if reviewedBy.Valid {
		record.ReviewedBy = &reviewedBy.String
	}
	if reviewedAt.Valid {
		record.ReviewedAt = &reviewedAt.Time
	}
	if approvedBy.Valid {
		record.ApprovedBy = &approvedBy.String
	}
	if approvedAt.Valid {
		record.ApprovedAt = &approvedAt.Time
	}
	if dueDate.Valid {
		record.DueDate = &dueDate.Time
	}
	if expiresAt.Valid {
		record.ExpiresAt = &expiresAt.Time
	}
	if nextReviewDate.Valid {
		record.NextReviewDate = &nextReviewDate.Time
	}
	if assignedTo.Valid {
		record.AssignedTo = &assignedTo.String
	}
	if notes.Valid {
		record.Notes = &notes.String
	}

	// Parse JSONB fields
	record.Requirements = parseJSONB(requirements)
	record.RawData = parseJSONB(rawData)
	record.Result = parseJSONB(result)
	record.Findings = parseJSONB(findings)
	record.Recommendations = parseJSONB(recommendations)
	record.Evidence = parseJSONB(evidence)
	record.Metadata = parseJSONB(metadata)

	// Parse tags array
	if tags.Valid {
		record.Tags = parseStringArray(tags.String)
	}

	return record, nil
}

func (ucr *UnifiedComplianceRepository) scanComplianceTrackingRow(row *sql.Row) (*services.ComplianceTracking, error) {
	record := &services.ComplianceTracking{}
	var requirements, rawData, result, findings, recommendations, evidence, metadata sql.NullString
	var tags sql.NullString
	var checkedBy, reviewedBy, approvedBy, assignedTo, notes sql.NullString
	var score sql.NullFloat64
	var reviewedAt, approvedAt, dueDate, expiresAt, nextReviewDate sql.NullTime

	err := row.Scan(
		&record.ID,
		&record.MerchantID,
		&record.ComplianceType,
		&record.ComplianceFramework,
		&record.CheckType,
		&record.Status,
		&score,
		&record.RiskLevel,
		&requirements,
		&record.CheckMethod,
		&record.Source,
		&rawData,
		&result,
		&findings,
		&recommendations,
		&evidence,
		&checkedBy,
		&record.CheckedAt,
		&reviewedBy,
		&reviewedAt,
		&approvedBy,
		&approvedAt,
		&dueDate,
		&expiresAt,
		&nextReviewDate,
		&record.Priority,
		&assignedTo,
		&tags,
		&notes,
		&metadata,
		&record.CreatedAt,
		&record.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Set nullable fields (same logic as scanComplianceTracking)
	if score.Valid {
		record.Score = &score.Float64
	}
	if checkedBy.Valid {
		record.CheckedBy = &checkedBy.String
	}
	if reviewedBy.Valid {
		record.ReviewedBy = &reviewedBy.String
	}
	if reviewedAt.Valid {
		record.ReviewedAt = &reviewedAt.Time
	}
	if approvedBy.Valid {
		record.ApprovedBy = &approvedBy.String
	}
	if approvedAt.Valid {
		record.ApprovedAt = &approvedAt.Time
	}
	if dueDate.Valid {
		record.DueDate = &dueDate.Time
	}
	if expiresAt.Valid {
		record.ExpiresAt = &expiresAt.Time
	}
	if nextReviewDate.Valid {
		record.NextReviewDate = &nextReviewDate.Time
	}
	if assignedTo.Valid {
		record.AssignedTo = &assignedTo.String
	}
	if notes.Valid {
		record.Notes = &notes.String
	}

	// Parse JSONB fields
	record.Requirements = parseJSONB(requirements)
	record.RawData = parseJSONB(rawData)
	record.Result = parseJSONB(result)
	record.Findings = parseJSONB(findings)
	record.Recommendations = parseJSONB(recommendations)
	record.Evidence = parseJSONB(evidence)
	record.Metadata = parseJSONB(metadata)

	// Parse tags array
	if tags.Valid {
		record.Tags = parseStringArray(tags.String)
	}

	return record, nil
}

// Utility functions for data conversion

func toJSONB(data map[string]interface{}) interface{} {
	if data == nil {
		return nil
	}
	// In a real implementation, you would use a JSON library to convert to JSONB
	// For now, we'll return the data as-is and let the database driver handle it
	return data
}

func toStringArray(tags []string) interface{} {
	if tags == nil {
		return nil
	}
	return tags
}

func parseJSONB(nullStr sql.NullString) map[string]interface{} {
	if !nullStr.Valid {
		return nil
	}
	// In a real implementation, you would parse the JSON string
	// For now, we'll return an empty map
	return make(map[string]interface{})
}

func parseStringArray(str string) []string {
	if str == "" {
		return nil
	}
	// Remove curly braces and split by comma
	str = strings.Trim(str, "{}")
	if str == "" {
		return nil
	}
	return strings.Split(str, ",")
}
