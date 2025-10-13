package businessintelligence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// SQLBIRepository implements BIRepository using SQL database
type SQLBIRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewSQLBIRepository creates a new SQL BI repository
func NewSQLBIRepository(db *sql.DB, logger *zap.Logger) *SQLBIRepository {
	return &SQLBIRepository{
		db:     db,
		logger: logger,
	}
}

// Data Sync Methods

// SaveDataSync saves a data sync to the database
func (r *SQLBIRepository) SaveDataSync(ctx context.Context, sync *DataSync) error {
	r.logger.Info("Saving data sync",
		zap.String("sync_id", sync.ID),
		zap.String("tenant_id", sync.TenantID),
		zap.String("name", sync.Name))

	// Convert complex fields to JSON
	sourceConfigJSON, err := json.Marshal(sync.SourceConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal source config: %w", err)
	}

	destinationConfigJSON, err := json.Marshal(sync.DestinationConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal destination config: %w", err)
	}

	syncScheduleJSON, err := json.Marshal(sync.SyncSchedule)
	if err != nil {
		return fmt.Errorf("failed to marshal sync schedule: %w", err)
	}

	metadataJSON, err := json.Marshal(sync.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Check if sync exists
	existingSync, err := r.GetDataSync(ctx, sync.TenantID, sync.ID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing sync: %w", err)
	}

	if existingSync != nil {
		// Update existing sync
		query := `
			UPDATE data_syncs SET
				name = $1, data_source_type = $2, source_config = $3, destination_config = $4,
				sync_schedule = $5, status = $6, last_sync_at = $7, next_sync_at = $8,
				records_synced = $9, records_failed = $10, error = $11, updated_at = $12, metadata = $13
			WHERE id = $14 AND tenant_id = $15
		`
		_, err = r.db.ExecContext(ctx, query,
			sync.Name, sync.DataSourceType, sourceConfigJSON, destinationConfigJSON,
			syncScheduleJSON, sync.Status, sync.LastSyncAt, sync.NextSyncAt,
			sync.RecordsSynced, sync.RecordsFailed, sync.Error, sync.UpdatedAt, metadataJSON,
			sync.ID, sync.TenantID)
	} else {
		// Insert new sync
		query := `
			INSERT INTO data_syncs (
				id, tenant_id, name, data_source_type, source_config, destination_config,
				sync_schedule, status, last_sync_at, next_sync_at, records_synced, records_failed,
				error, created_by, created_at, updated_at, metadata
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
			)
		`
		_, err = r.db.ExecContext(ctx, query,
			sync.ID, sync.TenantID, sync.Name, sync.DataSourceType, sourceConfigJSON, destinationConfigJSON,
			syncScheduleJSON, sync.Status, sync.LastSyncAt, sync.NextSyncAt, sync.RecordsSynced, sync.RecordsFailed,
			sync.Error, sync.CreatedBy, sync.CreatedAt, sync.UpdatedAt, metadataJSON)
	}

	if err != nil {
		return fmt.Errorf("failed to save data sync: %w", err)
	}

	r.logger.Info("Data sync saved successfully",
		zap.String("sync_id", sync.ID),
		zap.String("tenant_id", sync.TenantID))

	return nil
}

// GetDataSync retrieves a data sync by ID
func (r *SQLBIRepository) GetDataSync(ctx context.Context, tenantID, syncID string) (*DataSync, error) {
	r.logger.Debug("Getting data sync",
		zap.String("sync_id", syncID),
		zap.String("tenant_id", tenantID))

	query := `
		SELECT id, tenant_id, name, data_source_type, source_config, destination_config,
			   sync_schedule, status, last_sync_at, next_sync_at, records_synced, records_failed,
			   error, created_by, created_at, updated_at, metadata
		FROM data_syncs
		WHERE id = $1 AND tenant_id = $2
	`

	var result struct {
		ID                string     `json:"id"`
		TenantID          string     `json:"tenant_id"`
		Name              string     `json:"name"`
		DataSourceType    string     `json:"data_source_type"`
		SourceConfig      string     `json:"source_config"`
		DestinationConfig string     `json:"destination_config"`
		SyncSchedule      string     `json:"sync_schedule"`
		Status            string     `json:"status"`
		LastSyncAt        *time.Time `json:"last_sync_at"`
		NextSyncAt        *time.Time `json:"next_sync_at"`
		RecordsSynced     int64      `json:"records_synced"`
		RecordsFailed     int64      `json:"records_failed"`
		Error             string     `json:"error"`
		CreatedBy         string     `json:"created_by"`
		CreatedAt         time.Time  `json:"created_at"`
		UpdatedAt         time.Time  `json:"updated_at"`
		Metadata          string     `json:"metadata"`
	}

	err := r.db.QueryRowContext(ctx, query, syncID, tenantID).Scan(
		&result.ID, &result.TenantID, &result.Name, &result.DataSourceType, &result.SourceConfig,
		&result.DestinationConfig, &result.SyncSchedule, &result.Status, &result.LastSyncAt,
		&result.NextSyncAt, &result.RecordsSynced, &result.RecordsFailed, &result.Error,
		&result.CreatedBy, &result.CreatedAt, &result.UpdatedAt, &result.Metadata)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Sync not found
		}
		return nil, fmt.Errorf("failed to get data sync: %w", err)
	}

	// Convert the result to DataSync
	sync, err := r.convertToDataSync(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert result: %w", err)
	}

	r.logger.Debug("Data sync retrieved successfully",
		zap.String("sync_id", syncID),
		zap.String("tenant_id", tenantID))

	return sync, nil
}

// ListDataSyncs lists data syncs with filters
func (r *SQLBIRepository) ListDataSyncs(ctx context.Context, filter *BIFilter) ([]*DataSync, error) {
	r.logger.Debug("Listing data syncs",
		zap.String("tenant_id", filter.TenantID))

	// Build query with filters
	query := `
		SELECT id, tenant_id, name, data_source_type, source_config, destination_config,
			   sync_schedule, status, last_sync_at, next_sync_at, records_synced, records_failed,
			   error, created_by, created_at, updated_at, metadata
		FROM data_syncs
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if filter.TenantID != "" {
		query += fmt.Sprintf(" AND tenant_id = $%d", argIndex)
		args = append(args, filter.TenantID)
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
		return nil, fmt.Errorf("failed to list data syncs: %w", err)
	}
	defer rows.Close()

	var syncs []*DataSync
	for rows.Next() {
		var result struct {
			ID                string     `json:"id"`
			TenantID          string     `json:"tenant_id"`
			Name              string     `json:"name"`
			DataSourceType    string     `json:"data_source_type"`
			SourceConfig      string     `json:"source_config"`
			DestinationConfig string     `json:"destination_config"`
			SyncSchedule      string     `json:"sync_schedule"`
			Status            string     `json:"status"`
			LastSyncAt        *time.Time `json:"last_sync_at"`
			NextSyncAt        *time.Time `json:"next_sync_at"`
			RecordsSynced     int64      `json:"records_synced"`
			RecordsFailed     int64      `json:"records_failed"`
			Error             string     `json:"error"`
			CreatedBy         string     `json:"created_by"`
			CreatedAt         time.Time  `json:"created_at"`
			UpdatedAt         time.Time  `json:"updated_at"`
			Metadata          string     `json:"metadata"`
		}

		err := rows.Scan(
			&result.ID, &result.TenantID, &result.Name, &result.DataSourceType, &result.SourceConfig,
			&result.DestinationConfig, &result.SyncSchedule, &result.Status, &result.LastSyncAt,
			&result.NextSyncAt, &result.RecordsSynced, &result.RecordsFailed, &result.Error,
			&result.CreatedBy, &result.CreatedAt, &result.UpdatedAt, &result.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		sync, err := r.convertToDataSync(&result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert result: %w", err)
		}
		syncs = append(syncs, sync)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	r.logger.Debug("Data syncs listed successfully",
		zap.String("tenant_id", filter.TenantID),
		zap.Int("count", len(syncs)))

	return syncs, nil
}

// UpdateDataSync updates a data sync
func (r *SQLBIRepository) UpdateDataSync(ctx context.Context, sync *DataSync) error {
	r.logger.Debug("Updating data sync",
		zap.String("sync_id", sync.ID),
		zap.String("tenant_id", sync.TenantID))

	// Convert complex fields to JSON
	sourceConfigJSON, err := json.Marshal(sync.SourceConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal source config: %w", err)
	}

	destinationConfigJSON, err := json.Marshal(sync.DestinationConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal destination config: %w", err)
	}

	syncScheduleJSON, err := json.Marshal(sync.SyncSchedule)
	if err != nil {
		return fmt.Errorf("failed to marshal sync schedule: %w", err)
	}

	metadataJSON, err := json.Marshal(sync.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE data_syncs SET
			name = $1, data_source_type = $2, source_config = $3, destination_config = $4,
			sync_schedule = $5, status = $6, last_sync_at = $7, next_sync_at = $8,
			records_synced = $9, records_failed = $10, error = $11, updated_at = $12, metadata = $13
		WHERE id = $14 AND tenant_id = $15
	`

	_, err = r.db.ExecContext(ctx, query,
		sync.Name, sync.DataSourceType, sourceConfigJSON, destinationConfigJSON,
		syncScheduleJSON, sync.Status, sync.LastSyncAt, sync.NextSyncAt,
		sync.RecordsSynced, sync.RecordsFailed, sync.Error, sync.UpdatedAt, metadataJSON,
		sync.ID, sync.TenantID)

	if err != nil {
		return fmt.Errorf("failed to update data sync: %w", err)
	}

	r.logger.Debug("Data sync updated successfully",
		zap.String("sync_id", sync.ID),
		zap.String("tenant_id", sync.TenantID))

	return nil
}

// DeleteDataSync deletes a data sync
func (r *SQLBIRepository) DeleteDataSync(ctx context.Context, tenantID, syncID string) error {
	r.logger.Info("Deleting data sync",
		zap.String("sync_id", syncID),
		zap.String("tenant_id", tenantID))

	_, err := r.db.ExecContext(ctx, "DELETE FROM data_syncs WHERE id = $1 AND tenant_id = $2", syncID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete data sync: %w", err)
	}

	r.logger.Info("Data sync deleted successfully",
		zap.String("sync_id", syncID),
		zap.String("tenant_id", tenantID))

	return nil
}

// GetDataSyncsToRun retrieves data syncs that need to be run
func (r *SQLBIRepository) GetDataSyncsToRun(ctx context.Context) ([]*DataSync, error) {
	r.logger.Debug("Getting data syncs to run")

	query := `
		SELECT id, tenant_id, name, data_source_type, source_config, destination_config,
			   sync_schedule, status, last_sync_at, next_sync_at, records_synced, records_failed,
			   error, created_by, created_at, updated_at, metadata
		FROM data_syncs
		WHERE status = 'pending' AND next_sync_at <= $1
		ORDER BY next_sync_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get data syncs to run: %w", err)
	}
	defer rows.Close()

	var syncs []*DataSync
	for rows.Next() {
		var result struct {
			ID                string     `json:"id"`
			TenantID          string     `json:"tenant_id"`
			Name              string     `json:"name"`
			DataSourceType    string     `json:"data_source_type"`
			SourceConfig      string     `json:"source_config"`
			DestinationConfig string     `json:"destination_config"`
			SyncSchedule      string     `json:"sync_schedule"`
			Status            string     `json:"status"`
			LastSyncAt        *time.Time `json:"last_sync_at"`
			NextSyncAt        *time.Time `json:"next_sync_at"`
			RecordsSynced     int64      `json:"records_synced"`
			RecordsFailed     int64      `json:"records_failed"`
			Error             string     `json:"error"`
			CreatedBy         string     `json:"created_by"`
			CreatedAt         time.Time  `json:"created_at"`
			UpdatedAt         time.Time  `json:"updated_at"`
			Metadata          string     `json:"metadata"`
		}

		err := rows.Scan(
			&result.ID, &result.TenantID, &result.Name, &result.DataSourceType, &result.SourceConfig,
			&result.DestinationConfig, &result.SyncSchedule, &result.Status, &result.LastSyncAt,
			&result.NextSyncAt, &result.RecordsSynced, &result.RecordsFailed, &result.Error,
			&result.CreatedBy, &result.CreatedAt, &result.UpdatedAt, &result.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		sync, err := r.convertToDataSync(&result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert result: %w", err)
		}
		syncs = append(syncs, sync)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	r.logger.Debug("Data syncs to run retrieved successfully",
		zap.Int("count", len(syncs)))

	return syncs, nil
}

// UpdateDataSyncStatus updates the status of a data sync
func (r *SQLBIRepository) UpdateDataSyncStatus(ctx context.Context, tenantID, syncID string, status DataSyncStatus, errorMsg string) error {
	r.logger.Debug("Updating data sync status",
		zap.String("sync_id", syncID),
		zap.String("status", string(status)))

	query := `
		UPDATE data_syncs SET
			status = $1, error = $2, updated_at = $3
		WHERE id = $4 AND tenant_id = $5
	`

	_, err := r.db.ExecContext(ctx, query, status, errorMsg, time.Now(), syncID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update data sync status: %w", err)
	}

	r.logger.Debug("Data sync status updated successfully",
		zap.String("sync_id", syncID),
		zap.String("status", string(status)))

	return nil
}

// UpdateDataSyncLastRun updates the last run information of a data sync
func (r *SQLBIRepository) UpdateDataSyncLastRun(ctx context.Context, tenantID, syncID string, lastRunAt time.Time, nextRunAt *time.Time, recordsSynced, recordsFailed int64) error {
	r.logger.Debug("Updating data sync last run",
		zap.String("sync_id", syncID),
		zap.Time("last_run_at", lastRunAt))

	query := `
		UPDATE data_syncs SET
			last_sync_at = $1, next_sync_at = $2, records_synced = $3, records_failed = $4,
			status = $5, updated_at = $6
		WHERE id = $7 AND tenant_id = $8
	`

	_, err := r.db.ExecContext(ctx, query, lastRunAt, nextRunAt, recordsSynced, recordsFailed, DataSyncStatusCompleted, time.Now(), syncID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update data sync last run: %w", err)
	}

	r.logger.Debug("Data sync last run updated successfully",
		zap.String("sync_id", syncID),
		zap.Time("last_run_at", lastRunAt))

	return nil
}

// convertToDataSync converts a database result to DataSync
func (r *SQLBIRepository) convertToDataSync(result interface{}) (*DataSync, error) {
	// Type assertion to get the fields
	var id, tenantID, name, dataSourceType, sourceConfig, destinationConfig, syncSchedule, status, errorMsg, createdBy, metadata string
	var lastSyncAt, nextSyncAt *time.Time
	var recordsSynced, recordsFailed int64
	var createdAt, updatedAt time.Time

	// Use reflection or type assertion based on the actual structure
	switch v := result.(type) {
	case *struct {
		ID                string     `json:"id"`
		TenantID          string     `json:"tenant_id"`
		Name              string     `json:"name"`
		DataSourceType    string     `json:"data_source_type"`
		SourceConfig      string     `json:"source_config"`
		DestinationConfig string     `json:"destination_config"`
		SyncSchedule      string     `json:"sync_schedule"`
		Status            string     `json:"status"`
		LastSyncAt        *time.Time `json:"last_sync_at"`
		NextSyncAt        *time.Time `json:"next_sync_at"`
		RecordsSynced     int64      `json:"records_synced"`
		RecordsFailed     int64      `json:"records_failed"`
		Error             string     `json:"error"`
		CreatedBy         string     `json:"created_by"`
		CreatedAt         time.Time  `json:"created_at"`
		UpdatedAt         time.Time  `json:"updated_at"`
		Metadata          string     `json:"metadata"`
	}:
		id = v.ID
		tenantID = v.TenantID
		name = v.Name
		dataSourceType = v.DataSourceType
		sourceConfig = v.SourceConfig
		destinationConfig = v.DestinationConfig
		syncSchedule = v.SyncSchedule
		status = v.Status
		lastSyncAt = v.LastSyncAt
		nextSyncAt = v.NextSyncAt
		recordsSynced = v.RecordsSynced
		recordsFailed = v.RecordsFailed
		errorMsg = v.Error
		createdBy = v.CreatedBy
		createdAt = v.CreatedAt
		updatedAt = v.UpdatedAt
		metadata = v.Metadata
	default:
		return nil, fmt.Errorf("unsupported result type: %T", result)
	}

	// Parse JSON fields
	var sourceConfigObj DataSourceConfig
	if sourceConfig != "" {
		if err := json.Unmarshal([]byte(sourceConfig), &sourceConfigObj); err != nil {
			return nil, fmt.Errorf("failed to unmarshal source config: %w", err)
		}
	}

	var destinationConfigObj DestinationConfig
	if destinationConfig != "" {
		if err := json.Unmarshal([]byte(destinationConfig), &destinationConfigObj); err != nil {
			return nil, fmt.Errorf("failed to unmarshal destination config: %w", err)
		}
	}

	var syncScheduleObj SyncSchedule
	if syncSchedule != "" {
		if err := json.Unmarshal([]byte(syncSchedule), &syncScheduleObj); err != nil {
			return nil, fmt.Errorf("failed to unmarshal sync schedule: %w", err)
		}
	}

	var metadataMap map[string]interface{}
	if metadata != "" {
		if err := json.Unmarshal([]byte(metadata), &metadataMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	sync := &DataSync{
		ID:                id,
		TenantID:          tenantID,
		Name:              name,
		DataSourceType:    DataSourceType(dataSourceType),
		SourceConfig:      sourceConfigObj,
		DestinationConfig: destinationConfigObj,
		SyncSchedule:      syncScheduleObj,
		Status:            DataSyncStatus(status),
		LastSyncAt:        lastSyncAt,
		NextSyncAt:        nextSyncAt,
		RecordsSynced:     recordsSynced,
		RecordsFailed:     recordsFailed,
		Error:             errorMsg,
		CreatedBy:         createdBy,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		Metadata:          metadataMap,
	}

	return sync, nil
}

// Placeholder implementations for other methods
// These would follow similar patterns to the data sync methods

func (r *SQLBIRepository) SaveDataExport(ctx context.Context, export *DataExport) error {
	// Implementation for saving data export
	return fmt.Errorf("not implemented")
}

func (r *SQLBIRepository) GetDataExport(ctx context.Context, tenantID, exportID string) (*DataExport, error) {
	// Implementation for getting data export
	return nil, fmt.Errorf("not implemented")
}

func (r *SQLBIRepository) ListDataExports(ctx context.Context, filter *BIFilter) ([]*DataExport, error) {
	// Implementation for listing data exports
	return nil, fmt.Errorf("not implemented")
}

func (r *SQLBIRepository) DeleteDataExport(ctx context.Context, tenantID, exportID string) error {
	// Implementation for deleting data export
	return fmt.Errorf("not implemented")
}

func (r *SQLBIRepository) UpdateDataExportStatus(ctx context.Context, tenantID, exportID string, status DataExportStatus, fileSize int64, downloadURL string, recordsExported int64, errorMsg string) error {
	// Implementation for updating data export status
	return fmt.Errorf("not implemented")
}

func (r *SQLBIRepository) SaveBIQuery(ctx context.Context, query *BIQuery) error {
	// Implementation for saving BI query
	return fmt.Errorf("not implemented")
}

func (r *SQLBIRepository) GetBIQuery(ctx context.Context, tenantID, queryID string) (*BIQuery, error) {
	// Implementation for getting BI query
	return nil, fmt.Errorf("not implemented")
}

func (r *SQLBIRepository) ListBIQueries(ctx context.Context, filter *BIFilter) ([]*BIQuery, error) {
	// Implementation for listing BI queries
	return nil, fmt.Errorf("not implemented")
}

func (r *SQLBIRepository) UpdateBIQuery(ctx context.Context, query *BIQuery) error {
	// Implementation for updating BI query
	return fmt.Errorf("not implemented")
}

func (r *SQLBIRepository) DeleteBIQuery(ctx context.Context, tenantID, queryID string) error {
	// Implementation for deleting BI query
	return fmt.Errorf("not implemented")
}

func (r *SQLBIRepository) SaveBIDashboard(ctx context.Context, dashboard *BIDashboard) error {
	// Implementation for saving BI dashboard
	return fmt.Errorf("not implemented")
}

func (r *SQLBIRepository) GetBIDashboard(ctx context.Context, tenantID, dashboardID string) (*BIDashboard, error) {
	// Implementation for getting BI dashboard
	return nil, fmt.Errorf("not implemented")
}

func (r *SQLBIRepository) ListBIDashboards(ctx context.Context, filter *BIFilter) ([]*BIDashboard, error) {
	// Implementation for listing BI dashboards
	return nil, fmt.Errorf("not implemented")
}

func (r *SQLBIRepository) UpdateBIDashboard(ctx context.Context, dashboard *BIDashboard) error {
	// Implementation for updating BI dashboard
	return fmt.Errorf("not implemented")
}

func (r *SQLBIRepository) DeleteBIDashboard(ctx context.Context, tenantID, dashboardID string) error {
	// Implementation for deleting BI dashboard
	return fmt.Errorf("not implemented")
}

func (r *SQLBIRepository) GetBIMetrics(ctx context.Context, tenantID string) (*BIMetrics, error) {
	// Implementation for getting BI metrics
	return nil, fmt.Errorf("not implemented")
}

func (r *SQLBIRepository) GetDataSyncMetrics(ctx context.Context, tenantID string) (*DataSyncMetrics, error) {
	// Implementation for getting data sync metrics
	return nil, fmt.Errorf("not implemented")
}

func (r *SQLBIRepository) GetQueryPerformanceMetrics(ctx context.Context, tenantID string) (*QueryPerformanceMetrics, error) {
	// Implementation for getting query performance metrics
	return nil, fmt.Errorf("not implemented")
}
