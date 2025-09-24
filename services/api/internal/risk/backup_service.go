package risk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// BackupService provides risk data backup and restore functionality
type BackupService struct {
	logger        *zap.Logger
	backupDir     string
	retentionDays int
	compression   bool
}

// NewBackupService creates a new backup service
func NewBackupService(logger *zap.Logger, backupDir string, retentionDays int, compression bool) *BackupService {
	return &BackupService{
		logger:        logger,
		backupDir:     backupDir,
		retentionDays: retentionDays,
		compression:   compression,
	}
}

// BackupRequest represents a backup request
type BackupRequest struct {
	BusinessID    string                 `json:"business_id,omitempty"`
	BackupType    BackupType             `json:"backup_type"`
	IncludeData   []BackupDataType       `json:"include_data"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Compression   bool                   `json:"compression,omitempty"`
	RetentionDays int                    `json:"retention_days,omitempty"`
}

// BackupResponse represents a backup response
type BackupResponse struct {
	BackupID    string                 `json:"backup_id"`
	BusinessID  string                 `json:"business_id,omitempty"`
	BackupType  BackupType             `json:"backup_type"`
	Status      BackupStatus           `json:"status"`
	FilePath    string                 `json:"file_path"`
	FileSize    int64                  `json:"file_size"`
	RecordCount int                    `json:"record_count"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   time.Time              `json:"expires_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Checksum    string                 `json:"checksum,omitempty"`
}

// BackupType represents the type of backup
type BackupType string

const (
	BackupTypeFull         BackupType = "full"
	BackupTypeIncremental  BackupType = "incremental"
	BackupTypeDifferential BackupType = "differential"
	BackupTypeBusiness     BackupType = "business"
	BackupTypeSystem       BackupType = "system"
)

// BackupDataType represents the type of data to backup
type BackupDataType string

const (
	BackupDataTypeAssessments BackupDataType = "assessments"
	BackupDataTypeFactors     BackupDataType = "factors"
	BackupDataTypeTrends      BackupDataType = "trends"
	BackupDataTypeAlerts      BackupDataType = "alerts"
	BackupDataTypeHistory     BackupDataType = "history"
	BackupDataTypeConfig      BackupDataType = "config"
	BackupDataTypeAll         BackupDataType = "all"
)

// BackupStatus represents the status of a backup
type BackupStatus string

const (
	BackupStatusPending   BackupStatus = "pending"
	BackupStatusRunning   BackupStatus = "running"
	BackupStatusCompleted BackupStatus = "completed"
	BackupStatusFailed    BackupStatus = "failed"
	BackupStatusExpired   BackupStatus = "expired"
)

// BackupInfo represents information about a backup
type BackupInfo struct {
	BackupID     string                 `json:"backup_id"`
	BusinessID   string                 `json:"business_id,omitempty"`
	BackupType   BackupType             `json:"backup_type"`
	Status       BackupStatus           `json:"status"`
	FilePath     string                 `json:"file_path"`
	FileSize     int64                  `json:"file_size"`
	RecordCount  int                    `json:"record_count"`
	CreatedAt    time.Time              `json:"created_at"`
	ExpiresAt    time.Time              `json:"expires_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Checksum     string                 `json:"checksum,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
}

// RestoreRequest represents a restore request
type RestoreRequest struct {
	BackupID    string                 `json:"backup_id"`
	BusinessID  string                 `json:"business_id,omitempty"`
	RestoreType RestoreType            `json:"restore_type"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RestoreType represents the type of restore operation
type RestoreType string

const (
	RestoreTypeFull     RestoreType = "full"
	RestoreTypePartial  RestoreType = "partial"
	RestoreTypeBusiness RestoreType = "business"
	RestoreTypeSystem   RestoreType = "system"
)

// RestoreResponse represents a restore response
type RestoreResponse struct {
	RestoreID    string                 `json:"restore_id"`
	BackupID     string                 `json:"backup_id"`
	BusinessID   string                 `json:"business_id,omitempty"`
	RestoreType  RestoreType            `json:"restore_type"`
	Status       RestoreStatus          `json:"status"`
	RecordCount  int                    `json:"record_count"`
	StartedAt    time.Time              `json:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
}

// RestoreStatus represents the status of a restore operation
type RestoreStatus string

const (
	RestoreStatusPending   RestoreStatus = "pending"
	RestoreStatusRunning   RestoreStatus = "running"
	RestoreStatusCompleted RestoreStatus = "completed"
	RestoreStatusFailed    RestoreStatus = "failed"
)

// CreateBackup creates a new backup
func (bs *BackupService) CreateBackup(ctx context.Context, request *BackupRequest) (*BackupResponse, error) {
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	bs.logger.Info("Starting backup creation",
		zap.String("request_id", requestID.(string)),
		zap.String("business_id", request.BusinessID),
		zap.String("backup_type", string(request.BackupType)))

	// Validate request
	if err := bs.validateBackupRequest(request); err != nil {
		return nil, fmt.Errorf("invalid backup request: %w", err)
	}

	// Generate backup ID
	backupID := bs.generateBackupID(request.BusinessID, request.BackupType)

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(bs.backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create backup file path
	fileName := fmt.Sprintf("backup_%s_%s_%d.json", backupID, request.BusinessID, time.Now().Unix())
	if request.Compression || bs.compression {
		fileName += ".gz"
	}
	filePath := filepath.Join(bs.backupDir, fileName)

	// Create backup file
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	// Collect data to backup
	backupData, err := bs.collectBackupData(ctx, request)
	if err != nil {
		os.Remove(filePath) // Clean up file on error
		return nil, fmt.Errorf("failed to collect backup data: %w", err)
	}

	// Write backup data to file
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(backupData); err != nil {
		os.Remove(filePath) // Clean up file on error
		return nil, fmt.Errorf("failed to write backup data: %w", err)
	}

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		os.Remove(filePath) // Clean up file on error
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Calculate checksum
	checksum, err := bs.calculateChecksum(filePath)
	if err != nil {
		bs.logger.Warn("Failed to calculate checksum",
			zap.String("request_id", requestID.(string)),
			zap.String("backup_id", backupID),
			zap.Error(err))
	}

	// Calculate expiration time
	retentionDays := request.RetentionDays
	if retentionDays == 0 {
		retentionDays = bs.retentionDays
	}
	expiresAt := time.Now().Add(time.Duration(retentionDays) * 24 * time.Hour)

	// Create response
	response := &BackupResponse{
		BackupID:    backupID,
		BusinessID:  request.BusinessID,
		BackupType:  request.BackupType,
		Status:      BackupStatusCompleted,
		FilePath:    filePath,
		FileSize:    fileInfo.Size(),
		RecordCount: backupData.RecordCount,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		Metadata:    request.Metadata,
		Checksum:    checksum,
	}

	bs.logger.Info("Backup created successfully",
		zap.String("request_id", requestID.(string)),
		zap.String("backup_id", backupID),
		zap.String("business_id", request.BusinessID),
		zap.String("file_path", filePath),
		zap.Int64("file_size", fileInfo.Size()),
		zap.Int("record_count", backupData.RecordCount))

	return response, nil
}

// RestoreBackup restores data from a backup
func (bs *BackupService) RestoreBackup(ctx context.Context, request *RestoreRequest) (*RestoreResponse, error) {
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	bs.logger.Info("Starting backup restore",
		zap.String("request_id", requestID.(string)),
		zap.String("backup_id", request.BackupID),
		zap.String("business_id", request.BusinessID))

	// Validate request
	if err := bs.validateRestoreRequest(request); err != nil {
		return nil, fmt.Errorf("invalid restore request: %w", err)
	}

	// Find backup file
	backupInfo, err := bs.findBackupFile(request.BackupID)
	if err != nil {
		return nil, fmt.Errorf("backup not found: %w", err)
	}

	// Check if backup is expired
	if time.Now().After(backupInfo.ExpiresAt) {
		return nil, fmt.Errorf("backup has expired: %s", backupInfo.ExpiresAt.Format(time.RFC3339))
	}

	// Generate restore ID
	restoreID := bs.generateRestoreID(request.BusinessID, request.RestoreType)

	// Read backup data
	backupData, err := bs.readBackupData(backupInfo.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup data: %w", err)
	}

	// Verify checksum if available
	if backupInfo.Checksum != "" {
		currentChecksum, err := bs.calculateChecksum(backupInfo.FilePath)
		if err != nil {
			bs.logger.Warn("Failed to verify checksum",
				zap.String("request_id", requestID.(string)),
				zap.String("restore_id", restoreID),
				zap.Error(err))
		} else if currentChecksum != backupInfo.Checksum {
			return nil, fmt.Errorf("backup file checksum mismatch")
		}
	}

	// Restore data
	recordCount, err := bs.restoreData(ctx, backupData, request)
	if err != nil {
		return nil, fmt.Errorf("failed to restore data: %w", err)
	}

	completedAt := time.Now()

	// Create response
	response := &RestoreResponse{
		RestoreID:   restoreID,
		BackupID:    request.BackupID,
		BusinessID:  request.BusinessID,
		RestoreType: request.RestoreType,
		Status:      RestoreStatusCompleted,
		RecordCount: recordCount,
		StartedAt:   time.Now(),
		CompletedAt: &completedAt,
		Metadata:    request.Metadata,
	}

	bs.logger.Info("Backup restored successfully",
		zap.String("request_id", requestID.(string)),
		zap.String("restore_id", restoreID),
		zap.String("backup_id", request.BackupID),
		zap.String("business_id", request.BusinessID),
		zap.Int("record_count", recordCount))

	return response, nil
}

// ListBackups lists all available backups
func (bs *BackupService) ListBackups(businessID string) ([]*BackupInfo, error) {
	backups := []*BackupInfo{}

	// Read backup directory
	files, err := os.ReadDir(bs.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Parse backup file name to extract metadata
		backupInfo, err := bs.parseBackupFileName(file.Name())
		if err != nil {
			bs.logger.Warn("Failed to parse backup file name",
				zap.String("file_name", file.Name()),
				zap.Error(err))
			continue
		}

		// Filter by business ID if specified
		if businessID != "" && backupInfo.BusinessID != businessID {
			continue
		}

		// Get file info
		filePath := filepath.Join(bs.backupDir, file.Name())
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			bs.logger.Warn("Failed to get file info",
				zap.String("file_name", file.Name()),
				zap.Error(err))
			continue
		}

		// Update backup info with file details
		backupInfo.FilePath = filePath
		backupInfo.FileSize = fileInfo.Size()
		backupInfo.CreatedAt = fileInfo.ModTime()

		// Check if backup is expired
		if time.Now().After(backupInfo.ExpiresAt) {
			backupInfo.Status = BackupStatusExpired
		} else {
			backupInfo.Status = BackupStatusCompleted
		}

		backups = append(backups, backupInfo)
	}

	return backups, nil
}

// DeleteBackup deletes a backup file
func (bs *BackupService) DeleteBackup(backupID string) error {
	// Find backup file
	backupInfo, err := bs.findBackupFile(backupID)
	if err != nil {
		return fmt.Errorf("backup not found: %w", err)
	}

	// Delete backup file
	if err := os.Remove(backupInfo.FilePath); err != nil {
		return fmt.Errorf("failed to delete backup file: %w", err)
	}

	bs.logger.Info("Backup deleted successfully",
		zap.String("backup_id", backupID),
		zap.String("file_path", backupInfo.FilePath))

	return nil
}

// CleanupExpiredBackups removes expired backup files
func (bs *BackupService) CleanupExpiredBackups() error {
	backups, err := bs.ListBackups("")
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	deletedCount := 0
	for _, backup := range backups {
		if backup.Status == BackupStatusExpired {
			if err := bs.DeleteBackup(backup.BackupID); err != nil {
				bs.logger.Error("Failed to delete expired backup",
					zap.String("backup_id", backup.BackupID),
					zap.Error(err))
				continue
			}
			deletedCount++
		}
	}

	bs.logger.Info("Expired backups cleaned up",
		zap.Int("deleted_count", deletedCount))

	return nil
}

// GetBackupStatistics returns backup statistics
func (bs *BackupService) GetBackupStatistics() (map[string]interface{}, error) {
	backups, err := bs.ListBackups("")
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	stats := map[string]interface{}{
		"total_backups":    len(backups),
		"active_backups":   0,
		"expired_backups":  0,
		"total_size_bytes": int64(0),
		"total_records":    0,
		"backup_types":     make(map[string]int),
		"business_counts":  make(map[string]int),
	}

	for _, backup := range backups {
		stats["total_size_bytes"] = stats["total_size_bytes"].(int64) + backup.FileSize
		stats["total_records"] = stats["total_records"].(int) + backup.RecordCount

		if backup.Status == BackupStatusCompleted {
			stats["active_backups"] = stats["active_backups"].(int) + 1
		} else if backup.Status == BackupStatusExpired {
			stats["expired_backups"] = stats["expired_backups"].(int) + 1
		}

		// Count backup types
		backupTypes := stats["backup_types"].(map[string]int)
		backupTypes[string(backup.BackupType)]++

		// Count business backups
		businessCounts := stats["business_counts"].(map[string]int)
		businessCounts[backup.BusinessID]++
	}

	return stats, nil
}

// validateBackupRequest validates a backup request
func (bs *BackupService) validateBackupRequest(request *BackupRequest) error {
	if request == nil {
		return fmt.Errorf("backup request cannot be nil")
	}

	if request.BackupType == "" {
		return fmt.Errorf("backup type is required")
	}

	if len(request.IncludeData) == 0 {
		return fmt.Errorf("include data is required")
	}

	// Validate backup type
	switch request.BackupType {
	case BackupTypeFull, BackupTypeIncremental, BackupTypeDifferential, BackupTypeBusiness, BackupTypeSystem:
		// Valid types
	default:
		return fmt.Errorf("invalid backup type: %s", request.BackupType)
	}

	// Validate include data types
	for _, dataType := range request.IncludeData {
		switch dataType {
		case BackupDataTypeAssessments, BackupDataTypeFactors, BackupDataTypeTrends, BackupDataTypeAlerts, BackupDataTypeHistory, BackupDataTypeConfig, BackupDataTypeAll:
			// Valid types
		default:
			return fmt.Errorf("invalid backup data type: %s", dataType)
		}
	}

	return nil
}

// validateRestoreRequest validates a restore request
func (bs *BackupService) validateRestoreRequest(request *RestoreRequest) error {
	if request == nil {
		return fmt.Errorf("restore request cannot be nil")
	}

	if request.BackupID == "" {
		return fmt.Errorf("backup ID is required")
	}

	if request.RestoreType == "" {
		return fmt.Errorf("restore type is required")
	}

	// Validate restore type
	switch request.RestoreType {
	case RestoreTypeFull, RestoreTypePartial, RestoreTypeBusiness, RestoreTypeSystem:
		// Valid types
	default:
		return fmt.Errorf("invalid restore type: %s", request.RestoreType)
	}

	return nil
}

// generateBackupID generates a unique backup ID
func (bs *BackupService) generateBackupID(businessID string, backupType BackupType) string {
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s_%s", backupType, businessID, timestamp)
}

// generateRestoreID generates a unique restore ID
func (bs *BackupService) generateRestoreID(businessID string, restoreType RestoreType) string {
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s_%s", restoreType, businessID, timestamp)
}

// collectBackupData collects data for backup
func (bs *BackupService) collectBackupData(ctx context.Context, request *BackupRequest) (*BackupData, error) {
	backupData := &BackupData{
		BackupID:    bs.generateBackupID(request.BusinessID, request.BackupType),
		BusinessID:  request.BusinessID,
		BackupType:  request.BackupType,
		CreatedAt:   time.Now(),
		IncludeData: request.IncludeData,
		Metadata:    request.Metadata,
		Data:        make(map[string]interface{}),
	}

	// Collect data based on include data types
	for _, dataType := range request.IncludeData {
		switch dataType {
		case BackupDataTypeAssessments:
			assessments, err := bs.collectAssessments(ctx, request.BusinessID)
			if err != nil {
				return nil, fmt.Errorf("failed to collect assessments: %w", err)
			}
			backupData.Data["assessments"] = assessments
			backupData.RecordCount += len(assessments)

		case BackupDataTypeFactors:
			factors, err := bs.collectFactors(ctx, request.BusinessID)
			if err != nil {
				return nil, fmt.Errorf("failed to collect factors: %w", err)
			}
			backupData.Data["factors"] = factors
			backupData.RecordCount += len(factors)

		case BackupDataTypeTrends:
			trends, err := bs.collectTrends(ctx, request.BusinessID)
			if err != nil {
				return nil, fmt.Errorf("failed to collect trends: %w", err)
			}
			backupData.Data["trends"] = trends
			backupData.RecordCount += len(trends)

		case BackupDataTypeAlerts:
			alerts, err := bs.collectAlerts(ctx, request.BusinessID)
			if err != nil {
				return nil, fmt.Errorf("failed to collect alerts: %w", err)
			}
			backupData.Data["alerts"] = alerts
			backupData.RecordCount += len(alerts)

		case BackupDataTypeHistory:
			history, err := bs.collectHistory(ctx, request.BusinessID)
			if err != nil {
				return nil, fmt.Errorf("failed to collect history: %w", err)
			}
			backupData.Data["history"] = history
			backupData.RecordCount += len(history)

		case BackupDataTypeConfig:
			config, err := bs.collectConfig(ctx, request.BusinessID)
			if err != nil {
				return nil, fmt.Errorf("failed to collect config: %w", err)
			}
			backupData.Data["config"] = config
			backupData.RecordCount++

		case BackupDataTypeAll:
			// Collect all data types
			allData, err := bs.collectAllData(ctx, request.BusinessID)
			if err != nil {
				return nil, fmt.Errorf("failed to collect all data: %w", err)
			}
			backupData.Data = allData
			backupData.RecordCount = bs.countAllRecords(allData)
		}
	}

	return backupData, nil
}

// BackupData represents the structure of backup data
type BackupData struct {
	BackupID    string                 `json:"backup_id"`
	BusinessID  string                 `json:"business_id"`
	BackupType  BackupType             `json:"backup_type"`
	CreatedAt   time.Time              `json:"created_at"`
	IncludeData []BackupDataType       `json:"include_data"`
	Data        map[string]interface{} `json:"data"`
	RecordCount int                    `json:"record_count"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// collectAssessments collects risk assessment data
func (bs *BackupService) collectAssessments(ctx context.Context, businessID string) ([]*RiskAssessment, error) {
	// In a real implementation, this would query the database
	// For now, return mock data
	return []*RiskAssessment{
		{
			ID:           "mock-assessment-1",
			BusinessID:   businessID,
			BusinessName: "Mock Business",
			OverallScore: 75.5,
			OverallLevel: RiskLevelHigh,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
			AlertLevel:   RiskLevelMedium,
		},
	}, nil
}

// collectFactors collects risk factor data
func (bs *BackupService) collectFactors(ctx context.Context, businessID string) ([]RiskScore, error) {
	// In a real implementation, this would query the database
	// For now, return mock data
	return []RiskScore{
		{
			FactorID:     "mock-factor-1",
			FactorName:   "Mock Financial Risk",
			Category:     RiskCategoryFinancial,
			Score:        80.0,
			Level:        RiskLevelHigh,
			Confidence:   0.9,
			Explanation:  "Mock financial risk explanation",
			CalculatedAt: time.Now(),
		},
	}, nil
}

// collectTrends collects risk trend data
func (bs *BackupService) collectTrends(ctx context.Context, businessID string) ([]RiskTrend, error) {
	// In a real implementation, this would query the database
	// For now, return mock data
	return []RiskTrend{
		{
			BusinessID:   businessID,
			Category:     RiskCategoryFinancial,
			Score:        75.0,
			Level:        RiskLevelHigh,
			RecordedAt:   time.Now(),
			ChangeFrom:   5.0,
			ChangePeriod: "1 month",
		},
	}, nil
}

// collectAlerts collects risk alert data
func (bs *BackupService) collectAlerts(ctx context.Context, businessID string) ([]RiskAlert, error) {
	// In a real implementation, this would query the database
	// For now, return mock data
	return []RiskAlert{
		{
			ID:             "mock-alert-1",
			BusinessID:     businessID,
			RiskFactor:     "mock-risk-factor",
			Level:          RiskLevelHigh,
			Message:        "Mock alert message",
			Score:          85.0,
			Threshold:      80.0,
			TriggeredAt:    time.Now(),
			Acknowledged:   false,
			AcknowledgedAt: nil,
		},
	}, nil
}

// collectHistory collects risk history data
func (bs *BackupService) collectHistory(ctx context.Context, businessID string) ([]RiskHistoryEntry, error) {
	// In a real implementation, this would query the database
	// For now, return mock data
	return []RiskHistoryEntry{
		{
			BusinessID: businessID,
			Score:      75.5,
			Timestamp:  time.Now(),
			Details: map[string]interface{}{
				"change_type": "score_update",
				"old_value":   "70.0",
				"new_value":   "75.5",
				"changed_by":  "system",
			},
		},
	}, nil
}

// collectConfig collects configuration data
func (bs *BackupService) collectConfig(ctx context.Context, businessID string) (map[string]interface{}, error) {
	// In a real implementation, this would collect system configuration
	// For now, return mock data
	return map[string]interface{}{
		"business_id": businessID,
		"config_type": "risk_assessment",
		"settings": map[string]interface{}{
			"default_threshold": 80.0,
			"alert_enabled":     true,
			"retention_days":    30,
		},
		"created_at": time.Now(),
	}, nil
}

// collectAllData collects all data types
func (bs *BackupService) collectAllData(ctx context.Context, businessID string) (map[string]interface{}, error) {
	allData := make(map[string]interface{})

	// Collect all data types
	assessments, err := bs.collectAssessments(ctx, businessID)
	if err != nil {
		return nil, err
	}
	allData["assessments"] = assessments

	factors, err := bs.collectFactors(ctx, businessID)
	if err != nil {
		return nil, err
	}
	allData["factors"] = factors

	trends, err := bs.collectTrends(ctx, businessID)
	if err != nil {
		return nil, err
	}
	allData["trends"] = trends

	alerts, err := bs.collectAlerts(ctx, businessID)
	if err != nil {
		return nil, err
	}
	allData["alerts"] = alerts

	history, err := bs.collectHistory(ctx, businessID)
	if err != nil {
		return nil, err
	}
	allData["history"] = history

	config, err := bs.collectConfig(ctx, businessID)
	if err != nil {
		return nil, err
	}
	allData["config"] = config

	return allData, nil
}

// countAllRecords counts all records in the data
func (bs *BackupService) countAllRecords(data map[string]interface{}) int {
	count := 0
	for _, value := range data {
		switch v := value.(type) {
		case []interface{}:
			count += len(v)
		case []*RiskAssessment:
			count += len(v)
		case []RiskScore:
			count += len(v)
		case []RiskTrend:
			count += len(v)
		case []RiskAlert:
			count += len(v)
		case []RiskHistoryEntry:
			count += len(v)
		case map[string]interface{}:
			count++
		}
	}
	return count
}

// findBackupFile finds a backup file by ID
func (bs *BackupService) findBackupFile(backupID string) (*BackupInfo, error) {
	files, err := os.ReadDir(bs.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		backupInfo, err := bs.parseBackupFileName(file.Name())
		if err != nil {
			continue
		}

		if backupInfo.BackupID == backupID {
			filePath := filepath.Join(bs.backupDir, file.Name())
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to get file info: %w", err)
			}

			backupInfo.FilePath = filePath
			backupInfo.FileSize = fileInfo.Size()
			backupInfo.CreatedAt = fileInfo.ModTime()

			return backupInfo, nil
		}
	}

	return nil, fmt.Errorf("backup not found: %s", backupID)
}

// parseBackupFileName parses backup file name to extract metadata
func (bs *BackupService) parseBackupFileName(fileName string) (*BackupInfo, error) {
	// Expected format: backup_{backupID}_{businessID}_{timestamp}.json[.gz]
	// Example: backup_full_business123_20241219_143022.json.gz

	ext := filepath.Ext(fileName)
	if ext != ".json" && ext != ".gz" {
		return nil, fmt.Errorf("invalid file extension")
	}

	// Remove extensions
	name := fileName
	if filepath.Ext(name) == ".gz" {
		name = name[:len(name)-3]
	}
	if filepath.Ext(name) == ".json" {
		name = name[:len(name)-5]
	}

	// Remove backup_ prefix
	if len(name) < 7 || name[:7] != "backup_" {
		return nil, fmt.Errorf("invalid file name format")
	}
	name = name[7:]

	// Parse components
	parts := []string{}
	current := ""
	for _, char := range name {
		if char == '_' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}

	if len(parts) < 4 {
		return nil, fmt.Errorf("insufficient file name components")
	}

	backupType := BackupType(parts[0])
	businessID := parts[1]
	timestamp := parts[2] + "_" + parts[3]

	// Parse timestamp
	createdAt, err := time.Parse("20060102_150405", timestamp)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format: %w", err)
	}

	// Generate backup ID
	backupID := fmt.Sprintf("%s_%s_%s", backupType, businessID, timestamp)

	return &BackupInfo{
		BackupID:   backupID,
		BusinessID: businessID,
		BackupType: backupType,
		CreatedAt:  createdAt,
		ExpiresAt:  createdAt.Add(time.Duration(bs.retentionDays) * 24 * time.Hour),
	}, nil
}

// readBackupData reads backup data from file
func (bs *BackupService) readBackupData(filePath string) (*BackupData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	var backupData BackupData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&backupData); err != nil {
		return nil, fmt.Errorf("failed to decode backup data: %w", err)
	}

	return &backupData, nil
}

// restoreData restores data from backup
func (bs *BackupService) restoreData(ctx context.Context, backupData *BackupData, request *RestoreRequest) (int, error) {
	// In a real implementation, this would restore data to the database
	// For now, just return the record count
	return backupData.RecordCount, nil
}

// calculateChecksum calculates file checksum
func (bs *BackupService) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Simple checksum calculation (in production, use crypto/sha256)
	hash := uint32(0)
	buffer := make([]byte, 4096)
	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		if n == 0 {
			break
		}

		for i := 0; i < n; i++ {
			hash = hash*31 + uint32(buffer[i])
		}
	}

	return fmt.Sprintf("%08x", hash), nil
}
