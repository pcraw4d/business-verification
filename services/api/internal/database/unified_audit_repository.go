package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"kyb-platform/internal/models"
)

// UnifiedAuditRepository handles unified audit log operations
type UnifiedAuditRepository struct {
	db     *sql.DB
	logger *log.Logger
}

// NewUnifiedAuditRepository creates a new unified audit repository
func NewUnifiedAuditRepository(db *sql.DB, logger *log.Logger) *UnifiedAuditRepository {
	return &UnifiedAuditRepository{
		db:     db,
		logger: logger,
	}
}

// SaveAuditLog saves a unified audit log entry
func (r *UnifiedAuditRepository) SaveAuditLog(ctx context.Context, auditLog *models.UnifiedAuditLog) error {
	r.logger.Printf("Saving unified audit log: %s", auditLog.ID)

	query := `
		INSERT INTO unified_audit_logs (
			id, user_id, api_key_id, merchant_id, session_id,
			event_type, event_category, action,
			resource_type, resource_id, table_name,
			old_values, new_values, details,
			request_id, ip_address, user_agent,
			metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`

	_, err := r.db.ExecContext(ctx, query,
		auditLog.ID,
		auditLog.UserID,
		auditLog.APIKeyID,
		auditLog.MerchantID,
		auditLog.SessionID,
		auditLog.EventType,
		auditLog.EventCategory,
		auditLog.Action,
		auditLog.ResourceType,
		auditLog.ResourceID,
		auditLog.TableName,
		auditLog.OldValues,
		auditLog.NewValues,
		auditLog.Details,
		auditLog.RequestID,
		auditLog.IPAddress,
		auditLog.UserAgent,
		auditLog.Metadata,
		auditLog.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save unified audit log: %w", err)
	}

	r.logger.Printf("Successfully saved unified audit log: %s", auditLog.ID)
	return nil
}

// GetAuditLogs retrieves unified audit logs with filtering
func (r *UnifiedAuditRepository) GetAuditLogs(ctx context.Context, filters *models.UnifiedAuditLogFilters) (*models.UnifiedAuditLogResult, error) {
	r.logger.Printf("Getting unified audit logs with filters")

	// Build the WHERE clause
	whereClause, args := r.buildWhereClause(filters)

	// Build the query
	query := fmt.Sprintf(`
		SELECT 
			id, user_id, api_key_id, merchant_id, session_id,
			event_type, event_category, action,
			resource_type, resource_id, table_name,
			old_values, new_values, details,
			request_id, ip_address, user_agent,
			metadata, created_at
		FROM unified_audit_logs
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, len(args)+1, len(args)+2)

	// Add pagination parameters
	limit := filters.Limit
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	offset := filters.Offset
	if offset < 0 {
		offset = 0
	}

	args = append(args, limit, offset)

	// Execute the query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get unified audit logs: %w", err)
	}
	defer rows.Close()

	var auditLogs []*models.UnifiedAuditLog
	for rows.Next() {
		auditLog, err := r.scanUnifiedAuditLog(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan unified audit log: %w", err)
		}
		auditLogs = append(auditLogs, auditLog)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating unified audit logs: %w", err)
	}

	// Get total count
	total, err := r.getAuditLogsCount(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs count: %w", err)
	}

	hasMore := int64(offset+len(auditLogs)) < total

	result := &models.UnifiedAuditLogResult{
		AuditLogs: auditLogs,
		Total:     total,
		Page:      (offset / limit) + 1,
		PageSize:  limit,
		HasMore:   hasMore,
	}

	r.logger.Printf("Retrieved %d unified audit logs (total: %d)", len(auditLogs), total)
	return result, nil
}

// GetAuditLogByID retrieves a specific unified audit log by ID
func (r *UnifiedAuditRepository) GetAuditLogByID(ctx context.Context, id string) (*models.UnifiedAuditLog, error) {
	r.logger.Printf("Getting unified audit log by ID: %s", id)

	query := `
		SELECT 
			id, user_id, api_key_id, merchant_id, session_id,
			event_type, event_category, action,
			resource_type, resource_id, table_name,
			old_values, new_values, details,
			request_id, ip_address, user_agent,
			metadata, created_at
		FROM unified_audit_logs
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	auditLog, err := r.scanUnifiedAuditLog(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("unified audit log not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get unified audit log: %w", err)
	}

	r.logger.Printf("Successfully retrieved unified audit log: %s", id)
	return auditLog, nil
}

// GetAuditTrail retrieves audit trail for a specific merchant
func (r *UnifiedAuditRepository) GetAuditTrail(ctx context.Context, merchantID string, limit int, offset int) ([]*models.UnifiedAuditLog, error) {
	r.logger.Printf("Getting audit trail for merchant: %s", merchantID)

	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT 
			id, user_id, api_key_id, merchant_id, session_id,
			event_type, event_category, action,
			resource_type, resource_id, table_name,
			old_values, new_values, details,
			request_id, ip_address, user_agent,
			metadata, created_at
		FROM unified_audit_logs
		WHERE merchant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, merchantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit trail: %w", err)
	}
	defer rows.Close()

	var auditLogs []*models.UnifiedAuditLog
	for rows.Next() {
		auditLog, err := r.scanUnifiedAuditLog(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		auditLogs = append(auditLogs, auditLog)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating audit trail: %w", err)
	}

	r.logger.Printf("Retrieved %d audit trail entries for merchant %s", len(auditLogs), merchantID)
	return auditLogs, nil
}

// GetAuditLogsByUser retrieves audit logs for a specific user
func (r *UnifiedAuditRepository) GetAuditLogsByUser(ctx context.Context, userID string, limit int, offset int) ([]*models.UnifiedAuditLog, error) {
	r.logger.Printf("Getting audit logs for user: %s", userID)

	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT 
			id, user_id, api_key_id, merchant_id, session_id,
			event_type, event_category, action,
			resource_type, resource_id, table_name,
			old_values, new_values, details,
			request_id, ip_address, user_agent,
			metadata, created_at
		FROM unified_audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs by user: %w", err)
	}
	defer rows.Close()

	var auditLogs []*models.UnifiedAuditLog
	for rows.Next() {
		auditLog, err := r.scanUnifiedAuditLog(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		auditLogs = append(auditLogs, auditLog)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user audit logs: %w", err)
	}

	r.logger.Printf("Retrieved %d audit logs for user %s", len(auditLogs), userID)
	return auditLogs, nil
}

// GetAuditLogsByAction retrieves audit logs for a specific action
func (r *UnifiedAuditRepository) GetAuditLogsByAction(ctx context.Context, action string, limit int, offset int) ([]*models.UnifiedAuditLog, error) {
	r.logger.Printf("Getting audit logs for action: %s", action)

	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT 
			id, user_id, api_key_id, merchant_id, session_id,
			event_type, event_category, action,
			resource_type, resource_id, table_name,
			old_values, new_values, details,
			request_id, ip_address, user_agent,
			metadata, created_at
		FROM unified_audit_logs
		WHERE action = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, action, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs by action: %w", err)
	}
	defer rows.Close()

	var auditLogs []*models.UnifiedAuditLog
	for rows.Next() {
		auditLog, err := r.scanUnifiedAuditLog(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		auditLogs = append(auditLogs, auditLog)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating action audit logs: %w", err)
	}

	r.logger.Printf("Retrieved %d audit logs for action %s", len(auditLogs), action)
	return auditLogs, nil
}

// DeleteOldAuditLogs deletes audit logs older than the specified duration
func (r *UnifiedAuditRepository) DeleteOldAuditLogs(ctx context.Context, olderThan time.Duration) (int64, error) {
	r.logger.Printf("Deleting audit logs older than %v", olderThan)

	cutoffTime := time.Now().Add(-olderThan)

	query := `
		DELETE FROM unified_audit_logs
		WHERE created_at < $1
	`

	result, err := r.db.ExecContext(ctx, query, cutoffTime)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old audit logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Printf("Deleted %d old audit logs", rowsAffected)
	return rowsAffected, nil
}

// buildWhereClause builds the WHERE clause for filtering
func (r *UnifiedAuditRepository) buildWhereClause(filters *models.UnifiedAuditLogFilters) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filters.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *filters.UserID)
		argIndex++
	}

	if filters.APIKeyID != nil {
		conditions = append(conditions, fmt.Sprintf("api_key_id = $%d", argIndex))
		args = append(args, *filters.APIKeyID)
		argIndex++
	}

	if filters.MerchantID != nil {
		conditions = append(conditions, fmt.Sprintf("merchant_id = $%d", argIndex))
		args = append(args, *filters.MerchantID)
		argIndex++
	}

	if filters.SessionID != nil {
		conditions = append(conditions, fmt.Sprintf("session_id = $%d", argIndex))
		args = append(args, *filters.SessionID)
		argIndex++
	}

	if filters.EventType != nil {
		conditions = append(conditions, fmt.Sprintf("event_type = $%d", argIndex))
		args = append(args, *filters.EventType)
		argIndex++
	}

	if filters.EventCategory != nil {
		conditions = append(conditions, fmt.Sprintf("event_category = $%d", argIndex))
		args = append(args, *filters.EventCategory)
		argIndex++
	}

	if filters.Action != nil {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argIndex))
		args = append(args, *filters.Action)
		argIndex++
	}

	if filters.ResourceType != nil {
		conditions = append(conditions, fmt.Sprintf("resource_type = $%d", argIndex))
		args = append(args, *filters.ResourceType)
		argIndex++
	}

	if filters.ResourceID != nil {
		conditions = append(conditions, fmt.Sprintf("resource_id = $%d", argIndex))
		args = append(args, *filters.ResourceID)
		argIndex++
	}

	if filters.TableName != nil {
		conditions = append(conditions, fmt.Sprintf("table_name = $%d", argIndex))
		args = append(args, *filters.TableName)
		argIndex++
	}

	if filters.RequestID != nil {
		conditions = append(conditions, fmt.Sprintf("request_id = $%d", argIndex))
		args = append(args, *filters.RequestID)
		argIndex++
	}

	if filters.IPAddress != nil {
		conditions = append(conditions, fmt.Sprintf("ip_address = $%d", argIndex))
		args = append(args, *filters.IPAddress)
		argIndex++
	}

	if filters.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *filters.StartDate)
		argIndex++
	}

	if filters.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *filters.EndDate)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	return whereClause, args
}

// getAuditLogsCount gets the total count of audit logs matching the filters
func (r *UnifiedAuditRepository) getAuditLogsCount(ctx context.Context, filters *models.UnifiedAuditLogFilters) (int64, error) {
	whereClause, args := r.buildWhereClause(filters)

	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM unified_audit_logs
		%s
	`, whereClause)

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get audit logs count: %w", err)
	}

	return count, nil
}

// scanUnifiedAuditLog scans a row into a UnifiedAuditLog struct
func (r *UnifiedAuditRepository) scanUnifiedAuditLog(scanner interface{}) (*models.UnifiedAuditLog, error) {
	auditLog := &models.UnifiedAuditLog{}

	var (
		userID, apiKeyID, merchantID, sessionID sql.NullString
		resourceType, resourceID, tableName     sql.NullString
		requestID, ipAddress, userAgent         sql.NullString
		oldValues, newValues, details, metadata sql.NullString
	)

	var err error
	switch s := scanner.(type) {
	case *sql.Row:
		err = s.Scan(
			&auditLog.ID,
			&userID, &apiKeyID, &merchantID, &sessionID,
			&auditLog.EventType, &auditLog.EventCategory, &auditLog.Action,
			&resourceType, &resourceID, &tableName,
			&oldValues, &newValues, &details,
			&requestID, &ipAddress, &userAgent,
			&metadata, &auditLog.CreatedAt,
		)
	case *sql.Rows:
		err = s.Scan(
			&auditLog.ID,
			&userID, &apiKeyID, &merchantID, &sessionID,
			&auditLog.EventType, &auditLog.EventCategory, &auditLog.Action,
			&resourceType, &resourceID, &tableName,
			&oldValues, &newValues, &details,
			&requestID, &ipAddress, &userAgent,
			&metadata, &auditLog.CreatedAt,
		)
	default:
		return nil, fmt.Errorf("unsupported scanner type")
	}

	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if userID.Valid {
		auditLog.UserID = &userID.String
	}
	if apiKeyID.Valid {
		auditLog.APIKeyID = &apiKeyID.String
	}
	if merchantID.Valid {
		auditLog.MerchantID = &merchantID.String
	}
	if sessionID.Valid {
		auditLog.SessionID = &sessionID.String
	}
	if resourceType.Valid {
		auditLog.ResourceType = &resourceType.String
	}
	if resourceID.Valid {
		auditLog.ResourceID = &resourceID.String
	}
	if tableName.Valid {
		auditLog.TableName = &tableName.String
	}
	if requestID.Valid {
		auditLog.RequestID = &requestID.String
	}
	if ipAddress.Valid {
		auditLog.IPAddress = &ipAddress.String
	}
	if userAgent.Valid {
		auditLog.UserAgent = &userAgent.String
	}
	if oldValues.Valid {
		oldData := json.RawMessage(oldValues.String)
		auditLog.OldValues = &oldData
	}
	if newValues.Valid {
		newData := json.RawMessage(newValues.String)
		auditLog.NewValues = &newData
	}
	if details.Valid {
		detailsData := json.RawMessage(details.String)
		auditLog.Details = &detailsData
	}
	if metadata.Valid {
		metadataData := json.RawMessage(metadata.String)
		auditLog.Metadata = &metadataData
	}

	return auditLog, nil
}
