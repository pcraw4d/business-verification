package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// AuditRepositoryImpl implements the AuditRepository interface
type AuditRepositoryImpl struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *sql.DB, logger *zap.Logger) AuditRepository {
	return &AuditRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

// SaveAuditEvent saves an audit event to the database
func (r *AuditRepositoryImpl) SaveAuditEvent(ctx context.Context, event *AuditEvent) error {
	query := `
		INSERT INTO audit_events (
			id, tenant_id, user_id, session_id, action, resource, resource_id,
			method, endpoint, ip_address, user_agent, request_id, status,
			duration, request_size, response_size, metadata, hash, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)
	`

	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		event.ID, event.TenantID, event.UserID, event.SessionID, event.Action,
		event.Resource, event.ResourceID, event.Method, event.Endpoint,
		event.IPAddress, event.UserAgent, event.RequestID, event.Status,
		event.Duration, event.RequestSize, event.ResponseSize, metadataJSON,
		event.Hash, event.CreatedAt, event.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save audit event: %w", err)
	}

	return nil
}

// SaveAuditLog saves an audit log entry to the database
func (r *AuditRepositoryImpl) SaveAuditLog(ctx context.Context, log *AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			id, event_id, tenant_id, event_data, hash, prev_hash, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.EventID, log.TenantID, log.EventData, log.Hash, log.PrevHash, log.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save audit log: %w", err)
	}

	return nil
}

// GetAuditEvents retrieves audit events based on query parameters
func (r *AuditRepositoryImpl) GetAuditEvents(ctx context.Context, query AuditQuery) ([]AuditEvent, error) {
	whereClause, args := r.buildWhereClause(query)

	orderClause := "ORDER BY created_at DESC"
	if query.SortBy != "" {
		orderClause = fmt.Sprintf("ORDER BY %s %s", query.SortBy, query.SortOrder)
		if query.SortOrder == "" {
			orderClause += " DESC"
		}
	}

	limitClause := "LIMIT 1000" // Default limit
	if query.Limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", query.Limit)
	}

	offsetClause := ""
	if query.Offset > 0 {
		offsetClause = fmt.Sprintf("OFFSET %d", query.Offset)
	}

	sqlQuery := fmt.Sprintf(`
		SELECT id, tenant_id, user_id, session_id, action, resource, resource_id,
			   method, endpoint, ip_address, user_agent, request_id, status,
			   duration, request_size, response_size, metadata, hash, created_at, updated_at
		FROM audit_events
		%s
		%s
		%s
		%s
	`, whereClause, orderClause, limitClause, offsetClause)

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit events: %w", err)
	}
	defer rows.Close()

	var events []AuditEvent
	for rows.Next() {
		var event AuditEvent
		var metadataJSON []byte

		err := rows.Scan(
			&event.ID, &event.TenantID, &event.UserID, &event.SessionID, &event.Action,
			&event.Resource, &event.ResourceID, &event.Method, &event.Endpoint,
			&event.IPAddress, &event.UserAgent, &event.RequestID, &event.Status,
			&event.Duration, &event.RequestSize, &event.ResponseSize, &metadataJSON,
			&event.Hash, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit event: %w", err)
		}

		// Unmarshal metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
				r.logger.Warn("Failed to unmarshal metadata", zap.Error(err))
				event.Metadata = make(map[string]interface{})
			}
		} else {
			event.Metadata = make(map[string]interface{})
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating audit events: %w", err)
	}

	return events, nil
}

// GetAuditStats retrieves audit statistics for a tenant
func (r *AuditRepositoryImpl) GetAuditStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (*AuditStats, error) {
	stats := &AuditStats{
		EventsByAction: make(map[string]int64),
		EventsByUser:   make(map[string]int64),
		EventsByStatus: make(map[int]int64),
		EventsByDay:    make(map[string]int64),
		TopEndpoints:   []EndpointStats{},
		TopUsers:       []UserStats{},
	}

	// Total events
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM audit_events 
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3
	`, tenantID, startDate, endDate).Scan(&stats.TotalEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to get total events: %w", err)
	}

	// Events by action
	rows, err := r.db.QueryContext(ctx, `
		SELECT action, COUNT(*) FROM audit_events 
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY action
	`, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get events by action: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var action string
		var count int64
		if err := rows.Scan(&action, &count); err != nil {
			return nil, fmt.Errorf("failed to scan action stats: %w", err)
		}
		stats.EventsByAction[action] = count
	}

	// Events by user
	rows, err = r.db.QueryContext(ctx, `
		SELECT user_id, COUNT(*) FROM audit_events 
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3 AND user_id IS NOT NULL
		GROUP BY user_id
		ORDER BY COUNT(*) DESC
		LIMIT 10
	`, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get events by user: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		var count int64
		if err := rows.Scan(&userID, &count); err != nil {
			return nil, fmt.Errorf("failed to scan user stats: %w", err)
		}
		stats.EventsByUser[userID] = count
	}

	// Events by status
	rows, err = r.db.QueryContext(ctx, `
		SELECT status, COUNT(*) FROM audit_events 
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY status
	`, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get events by status: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status int
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status stats: %w", err)
		}
		stats.EventsByStatus[status] = count
	}

	// Events by day
	rows, err = r.db.QueryContext(ctx, `
		SELECT DATE(created_at), COUNT(*) FROM audit_events 
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY DATE(created_at)
		ORDER BY DATE(created_at)
	`, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get events by day: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var day string
		var count int64
		if err := rows.Scan(&day, &count); err != nil {
			return nil, fmt.Errorf("failed to scan day stats: %w", err)
		}
		stats.EventsByDay[day] = count
	}

	// Average duration
	err = r.db.QueryRowContext(ctx, `
		SELECT AVG(duration) FROM audit_events 
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3 AND duration IS NOT NULL
	`, tenantID, startDate, endDate).Scan(&stats.AverageDuration)
	if err != nil {
		r.logger.Warn("Failed to get average duration", zap.Error(err))
	}

	// Error rate
	var totalErrors int64
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM audit_events 
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3 AND status >= 400
	`, tenantID, startDate, endDate).Scan(&totalErrors)
	if err != nil {
		return nil, fmt.Errorf("failed to get error count: %w", err)
	}

	if stats.TotalEvents > 0 {
		stats.ErrorRate = float64(totalErrors) / float64(stats.TotalEvents) * 100
	}

	// Top endpoints
	rows, err = r.db.QueryContext(ctx, `
		SELECT endpoint, method, COUNT(*), AVG(duration), 
			   COUNT(CASE WHEN status >= 400 THEN 1 END)
		FROM audit_events 
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY endpoint, method
		ORDER BY COUNT(*) DESC
		LIMIT 10
	`, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get top endpoints: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var endpoint, method string
		var count int64
		var avgDuration sql.NullFloat64
		var errorCount int64

		if err := rows.Scan(&endpoint, &method, &count, &avgDuration, &errorCount); err != nil {
			return nil, fmt.Errorf("failed to scan endpoint stats: %w", err)
		}

		epStats := EndpointStats{
			Endpoint:     endpoint,
			Method:       method,
			RequestCount: count,
			ErrorCount:   errorCount,
		}

		if avgDuration.Valid {
			epStats.AverageTime = avgDuration.Float64
		}

		if count > 0 {
			epStats.ErrorRate = float64(errorCount) / float64(count) * 100
		}

		stats.TopEndpoints = append(stats.TopEndpoints, epStats)
	}

	// Top users
	rows, err = r.db.QueryContext(ctx, `
		SELECT user_id, COUNT(*), MAX(created_at),
			   COUNT(CASE WHEN status >= 400 THEN 1 END)
		FROM audit_events 
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3 AND user_id IS NOT NULL
		GROUP BY user_id
		ORDER BY COUNT(*) DESC
		LIMIT 10
	`, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get top users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		var count int64
		var lastActivity time.Time
		var errorCount int64

		if err := rows.Scan(&userID, &count, &lastActivity, &errorCount); err != nil {
			return nil, fmt.Errorf("failed to scan user stats: %w", err)
		}

		userStats := UserStats{
			UserID:       userID,
			RequestCount: count,
			LastActivity: lastActivity,
			ErrorCount:   errorCount,
		}

		if count > 0 {
			userStats.ErrorRate = float64(errorCount) / float64(count) * 100
		}

		stats.TopUsers = append(stats.TopUsers, userStats)
	}

	return stats, nil
}

// GetAuditLog retrieves an audit log entry by event ID
func (r *AuditRepositoryImpl) GetAuditLog(ctx context.Context, eventID string) (*AuditLog, error) {
	query := `
		SELECT id, event_id, tenant_id, event_data, hash, prev_hash, created_at
		FROM audit_logs
		WHERE event_id = $1
	`

	var log AuditLog
	err := r.db.QueryRowContext(ctx, query, eventID).Scan(
		&log.ID, &log.EventID, &log.TenantID, &log.EventData, &log.Hash, &log.PrevHash, &log.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("audit log not found for event %s", eventID)
		}
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	return &log, nil
}

// DeleteExpiredLogs deletes audit logs older than the specified time
func (r *AuditRepositoryImpl) DeleteExpiredLogs(ctx context.Context, before time.Time) error {
	query := `DELETE FROM audit_logs WHERE created_at < $1`

	result, err := r.db.ExecContext(ctx, query, before)
	if err != nil {
		return fmt.Errorf("failed to delete expired logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info("Deleted expired audit logs", zap.Int64("count", rowsAffected))
	return nil
}

// GetAuditLogsByHash retrieves audit logs by hash
func (r *AuditRepositoryImpl) GetAuditLogsByHash(ctx context.Context, hash string) ([]AuditLog, error) {
	query := `
		SELECT id, event_id, tenant_id, event_data, hash, prev_hash, created_at
		FROM audit_logs
		WHERE hash = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs by hash: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		err := rows.Scan(
			&log.ID, &log.EventID, &log.TenantID, &log.EventData, &log.Hash, &log.PrevHash, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating audit logs: %w", err)
	}

	return logs, nil
}

// buildWhereClause builds the WHERE clause for audit event queries
func (r *AuditRepositoryImpl) buildWhereClause(query AuditQuery) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if query.TenantID != "" {
		conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
		args = append(args, query.TenantID)
		argIndex++
	}

	if query.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, query.UserID)
		argIndex++
	}

	if query.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argIndex))
		args = append(args, query.Action)
		argIndex++
	}

	if query.Resource != "" {
		conditions = append(conditions, fmt.Sprintf("resource = $%d", argIndex))
		args = append(args, query.Resource)
		argIndex++
	}

	if query.ResourceID != "" {
		conditions = append(conditions, fmt.Sprintf("resource_id = $%d", argIndex))
		args = append(args, query.ResourceID)
		argIndex++
	}

	if query.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *query.StartDate)
		argIndex++
	}

	if query.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *query.EndDate)
		argIndex++
	}

	if query.IPAddress != "" {
		conditions = append(conditions, fmt.Sprintf("ip_address = $%d", argIndex))
		args = append(args, query.IPAddress)
		argIndex++
	}

	if query.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *query.Status)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	return whereClause, args
}
