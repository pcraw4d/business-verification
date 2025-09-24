package risk

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBackupService_CreateBackup(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	logger := zap.NewNop()
	service := NewBackupService(logger, tempDir, 30, false)

	request := &BackupRequest{
		BusinessID:  "test-business-123",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
		Metadata:    map[string]interface{}{"test": "value"},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

	response, err := service.CreateBackup(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "test-business-123", response.BusinessID)
	assert.Equal(t, BackupTypeFull, response.BackupType)
	assert.Equal(t, BackupStatusCompleted, response.Status)
	assert.NotEmpty(t, response.BackupID)
	assert.NotEmpty(t, response.FilePath)
	assert.Greater(t, response.FileSize, int64(0))
	assert.Greater(t, response.RecordCount, 0)
	assert.NotNil(t, response.CreatedAt)
	assert.NotNil(t, response.ExpiresAt)
	assert.Equal(t, map[string]interface{}{"test": "value"}, response.Metadata)

	// Verify backup file exists
	assert.FileExists(t, response.FilePath)
}

func TestBackupService_CreateBackup_InvalidRequest(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	service := NewBackupService(logger, tempDir, 30, false)

	tests := []struct {
		name    string
		request *BackupRequest
		wantErr string
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: "backup request cannot be nil",
		},
		{
			name: "empty backup type",
			request: &BackupRequest{
				BusinessID:  "test-business-123",
				BackupType:  "",
				IncludeData: []BackupDataType{BackupDataTypeAssessments},
			},
			wantErr: "backup type is required",
		},
		{
			name: "empty include data",
			request: &BackupRequest{
				BusinessID:  "test-business-123",
				BackupType:  BackupTypeFull,
				IncludeData: []BackupDataType{},
			},
			wantErr: "include data is required",
		},
		{
			name: "invalid backup type",
			request: &BackupRequest{
				BusinessID:  "test-business-123",
				BackupType:  "invalid_type",
				IncludeData: []BackupDataType{BackupDataTypeAssessments},
			},
			wantErr: "invalid backup type",
		},
		{
			name: "invalid backup data type",
			request: &BackupRequest{
				BusinessID:  "test-business-123",
				BackupType:  BackupTypeFull,
				IncludeData: []BackupDataType{"invalid_data_type"},
			},
			wantErr: "invalid backup data type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
			response, err := service.CreateBackup(ctx, tt.request)

			assert.Error(t, err)
			assert.Nil(t, response)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestBackupService_CreateBackup_DifferentTypes(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	service := NewBackupService(logger, tempDir, 30, false)

	backupTypes := []BackupType{
		BackupTypeFull,
		BackupTypeIncremental,
		BackupTypeDifferential,
		BackupTypeBusiness,
		BackupTypeSystem,
	}

	for _, backupType := range backupTypes {
		t.Run(string(backupType), func(t *testing.T) {
			request := &BackupRequest{
				BusinessID:  "test-business-123",
				BackupType:  backupType,
				IncludeData: []BackupDataType{BackupDataTypeAssessments},
			}

			ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
			response, err := service.CreateBackup(ctx, request)

			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, backupType, response.BackupType)
		})
	}
}

func TestBackupService_CreateBackup_DifferentDataTypes(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	service := NewBackupService(logger, tempDir, 30, false)

	dataTypes := []BackupDataType{
		BackupDataTypeAssessments,
		BackupDataTypeFactors,
		BackupDataTypeTrends,
		BackupDataTypeAlerts,
		BackupDataTypeHistory,
		BackupDataTypeConfig,
		BackupDataTypeAll,
	}

	for _, dataType := range dataTypes {
		t.Run(string(dataType), func(t *testing.T) {
			request := &BackupRequest{
				BusinessID:  "test-business-123",
				BackupType:  BackupTypeFull,
				IncludeData: []BackupDataType{dataType},
			}

			ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
			response, err := service.CreateBackup(ctx, request)

			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Greater(t, response.RecordCount, 0)
		})
	}
}

func TestBackupService_CreateBackup_MultipleDataTypes(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	service := NewBackupService(logger, tempDir, 30, false)

	request := &BackupRequest{
		BusinessID: "test-business-123",
		BackupType: BackupTypeFull,
		IncludeData: []BackupDataType{
			BackupDataTypeAssessments,
			BackupDataTypeFactors,
			BackupDataTypeTrends,
		},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	response, err := service.CreateBackup(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Greater(t, response.RecordCount, 0)
}

func TestBackupService_CreateBackup_CustomRetention(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	service := NewBackupService(logger, tempDir, 30, false)

	request := &BackupRequest{
		BusinessID:    "test-business-123",
		BackupType:    BackupTypeFull,
		IncludeData:   []BackupDataType{BackupDataTypeAssessments},
		RetentionDays: 7,
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	response, err := service.CreateBackup(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)

	// Check that expiration is set correctly
	expectedExpiration := response.CreatedAt.Add(7 * 24 * time.Hour)
	assert.WithinDuration(t, expectedExpiration, response.ExpiresAt, time.Minute)
}

func TestBackupService_RestoreBackup(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	service := NewBackupService(logger, tempDir, 30, false)

	// First create a backup
	backupRequest := &BackupRequest{
		BusinessID:  "test-business-123",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	backupResponse, err := service.CreateBackup(ctx, backupRequest)
	require.NoError(t, err)

	// Now restore the backup
	restoreRequest := &RestoreRequest{
		BackupID:    backupResponse.BackupID,
		BusinessID:  "test-business-123",
		RestoreType: RestoreTypeFull,
		Metadata:    map[string]interface{}{"restore_test": "value"},
	}

	restoreResponse, err := service.RestoreBackup(ctx, restoreRequest)

	assert.NoError(t, err)
	assert.NotNil(t, restoreResponse)
	assert.Equal(t, backupResponse.BackupID, restoreResponse.BackupID)
	assert.Equal(t, "test-business-123", restoreResponse.BusinessID)
	assert.Equal(t, RestoreTypeFull, restoreResponse.RestoreType)
	assert.Equal(t, RestoreStatusCompleted, restoreResponse.Status)
	assert.NotEmpty(t, restoreResponse.RestoreID)
	assert.Greater(t, restoreResponse.RecordCount, 0)
	assert.NotNil(t, restoreResponse.StartedAt)
	assert.NotNil(t, restoreResponse.CompletedAt)
	assert.Equal(t, map[string]interface{}{"restore_test": "value"}, restoreResponse.Metadata)
}

func TestBackupService_RestoreBackup_InvalidRequest(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	service := NewBackupService(logger, tempDir, 30, false)

	tests := []struct {
		name    string
		request *RestoreRequest
		wantErr string
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: "restore request cannot be nil",
		},
		{
			name: "empty backup ID",
			request: &RestoreRequest{
				BackupID:    "",
				BusinessID:  "test-business-123",
				RestoreType: RestoreTypeFull,
			},
			wantErr: "backup ID is required",
		},
		{
			name: "empty restore type",
			request: &RestoreRequest{
				BackupID:    "test-backup-123",
				BusinessID:  "test-business-123",
				RestoreType: "",
			},
			wantErr: "restore type is required",
		},
		{
			name: "invalid restore type",
			request: &RestoreRequest{
				BackupID:    "test-backup-123",
				BusinessID:  "test-business-123",
				RestoreType: "invalid_type",
			},
			wantErr: "invalid restore type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
			response, err := service.RestoreBackup(ctx, tt.request)

			assert.Error(t, err)
			assert.Nil(t, response)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestBackupService_RestoreBackup_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	service := NewBackupService(logger, tempDir, 30, false)

	request := &RestoreRequest{
		BackupID:    "non-existent-backup",
		BusinessID:  "test-business-123",
		RestoreType: RestoreTypeFull,
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	response, err := service.RestoreBackup(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "backup not found")
}

func TestBackupService_RestoreBackup_DifferentTypes(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	service := NewBackupService(logger, tempDir, 30, false)

	// Create a backup first
	backupRequest := &BackupRequest{
		BusinessID:  "test-business-123",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	backupResponse, err := service.CreateBackup(ctx, backupRequest)
	require.NoError(t, err)

	restoreTypes := []RestoreType{
		RestoreTypeFull,
		RestoreTypePartial,
		RestoreTypeBusiness,
		RestoreTypeSystem,
	}

	for _, restoreType := range restoreTypes {
		t.Run(string(restoreType), func(t *testing.T) {
			request := &RestoreRequest{
				BackupID:    backupResponse.BackupID,
				BusinessID:  "test-business-123",
				RestoreType: restoreType,
			}

			response, err := service.RestoreBackup(ctx, request)

			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, restoreType, response.RestoreType)
		})
	}
}

func TestBackupService_ListBackups(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	service := NewBackupService(logger, tempDir, 30, false)

	// Create multiple backups
	businessID1 := "test-business-123"
	businessID2 := "test-business-456"

	backupRequests := []*BackupRequest{
		{
			BusinessID:  businessID1,
			BackupType:  BackupTypeFull,
			IncludeData: []BackupDataType{BackupDataTypeAssessments},
		},
		{
			BusinessID:  businessID1,
			BackupType:  BackupTypeIncremental,
			IncludeData: []BackupDataType{BackupDataTypeFactors},
		},
		{
			BusinessID:  businessID2,
			BackupType:  BackupTypeFull,
			IncludeData: []BackupDataType{BackupDataTypeAssessments},
		},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	for _, req := range backupRequests {
		_, err := service.CreateBackup(ctx, req)
		require.NoError(t, err)
	}

	// List all backups
	allBackups, err := service.ListBackups("")
	assert.NoError(t, err)
	assert.Len(t, allBackups, 3)

	// List backups for specific business
	business1Backups, err := service.ListBackups(businessID1)
	assert.NoError(t, err)
	assert.Len(t, business1Backups, 2)

	business2Backups, err := service.ListBackups(businessID2)
	assert.NoError(t, err)
	assert.Len(t, business2Backups, 1)

	// Verify backup info
	for _, backup := range allBackups {
		assert.NotEmpty(t, backup.BackupID)
		assert.NotEmpty(t, backup.BusinessID)
		assert.NotEmpty(t, backup.BackupType)
		assert.NotEmpty(t, backup.FilePath)
		assert.Greater(t, backup.FileSize, int64(0))
		assert.NotNil(t, backup.CreatedAt)
		assert.NotNil(t, backup.ExpiresAt)
		assert.Equal(t, BackupStatusCompleted, backup.Status)
	}
}

func TestBackupService_DeleteBackup(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	service := NewBackupService(logger, tempDir, 30, false)

	// Create a backup
	request := &BackupRequest{
		BusinessID:  "test-business-123",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	response, err := service.CreateBackup(ctx, request)
	require.NoError(t, err)

	// Verify backup file exists
	assert.FileExists(t, response.FilePath)

	// Delete the backup
	err = service.DeleteBackup(response.BackupID)
	assert.NoError(t, err)

	// Verify backup file is deleted
	assert.NoFileExists(t, response.FilePath)

	// Try to delete non-existent backup
	err = service.DeleteBackup("non-existent-backup")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backup not found")
}

func TestBackupService_CleanupExpiredBackups(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	service := NewBackupService(logger, tempDir, 1, false) // 1 day retention

	// Create a backup
	request := &BackupRequest{
		BusinessID:  "test-business-123",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	response, err := service.CreateBackup(ctx, request)
	require.NoError(t, err)

	// Manually set the backup file to be expired by modifying its timestamp
	expiredTime := time.Now().Add(-2 * 24 * time.Hour) // 2 days ago
	err = os.Chtimes(response.FilePath, expiredTime, expiredTime)
	require.NoError(t, err)

	// List backups to verify it's marked as expired
	backups, err := service.ListBackups("")
	assert.NoError(t, err)
	assert.Len(t, backups, 1)
	assert.Equal(t, BackupStatusExpired, backups[0].Status)

	// Cleanup expired backups
	err = service.CleanupExpiredBackups()
	assert.NoError(t, err)

	// Verify backup file is deleted
	assert.NoFileExists(t, response.FilePath)

	// Verify no backups remain
	backups, err = service.ListBackups("")
	assert.NoError(t, err)
	assert.Len(t, backups, 0)
}

func TestBackupService_GetBackupStatistics(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	service := NewBackupService(logger, tempDir, 30, false)

	// Create multiple backups
	backupRequests := []*BackupRequest{
		{
			BusinessID:  "test-business-123",
			BackupType:  BackupTypeFull,
			IncludeData: []BackupDataType{BackupDataTypeAssessments},
		},
		{
			BusinessID:  "test-business-123",
			BackupType:  BackupTypeIncremental,
			IncludeData: []BackupDataType{BackupDataTypeFactors},
		},
		{
			BusinessID:  "test-business-456",
			BackupType:  BackupTypeFull,
			IncludeData: []BackupDataType{BackupDataTypeAssessments},
		},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	for _, req := range backupRequests {
		_, err := service.CreateBackup(ctx, req)
		require.NoError(t, err)
	}

	// Get statistics
	stats, err := service.GetBackupStatistics()
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	// Verify statistics
	assert.Equal(t, 3, stats["total_backups"])
	assert.Equal(t, 3, stats["active_backups"])
	assert.Equal(t, 0, stats["expired_backups"])
	assert.Greater(t, stats["total_size_bytes"], int64(0))
	assert.Greater(t, stats["total_records"], 0)

	// Verify backup types
	backupTypes := stats["backup_types"].(map[string]int)
	assert.Equal(t, 2, backupTypes["full"])
	assert.Equal(t, 1, backupTypes["incremental"])

	// Verify business counts
	businessCounts := stats["business_counts"].(map[string]int)
	assert.Equal(t, 2, businessCounts["test-business-123"])
	assert.Equal(t, 1, businessCounts["test-business-456"])
}

func TestBackupService_BackupDataStructure(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	service := NewBackupService(logger, tempDir, 30, false)

	request := &BackupRequest{
		BusinessID:  "test-business-123",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments, BackupDataTypeFactors},
		Metadata:    map[string]interface{}{"test": "value"},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	response, err := service.CreateBackup(ctx, request)
	require.NoError(t, err)

	// Read the backup file and verify its structure
	backupData, err := service.readBackupData(response.FilePath)
	assert.NoError(t, err)
	assert.NotNil(t, backupData)

	// Verify backup data structure
	assert.Equal(t, response.BackupID, backupData.BackupID)
	assert.Equal(t, "test-business-123", backupData.BusinessID)
	assert.Equal(t, BackupTypeFull, backupData.BackupType)
	assert.Equal(t, []BackupDataType{BackupDataTypeAssessments, BackupDataTypeFactors}, backupData.IncludeData)
	assert.Equal(t, map[string]interface{}{"test": "value"}, backupData.Metadata)
	assert.Greater(t, backupData.RecordCount, 0)
	assert.NotNil(t, backupData.CreatedAt)
	assert.NotNil(t, backupData.Data)

	// Verify data contains expected keys
	assert.Contains(t, backupData.Data, "assessments")
	assert.Contains(t, backupData.Data, "factors")
}

func TestBackupService_ChecksumCalculation(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	service := NewBackupService(logger, tempDir, 30, false)

	request := &BackupRequest{
		BusinessID:  "test-business-123",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	response, err := service.CreateBackup(ctx, request)
	require.NoError(t, err)

	// Verify checksum is calculated
	assert.NotEmpty(t, response.Checksum)

	// Verify checksum is consistent
	checksum1, err := service.calculateChecksum(response.FilePath)
	assert.NoError(t, err)
	assert.Equal(t, response.Checksum, checksum1)

	// Verify checksum changes when file is modified
	err = os.WriteFile(response.FilePath, []byte("modified content"), 0644)
	require.NoError(t, err)

	checksum2, err := service.calculateChecksum(response.FilePath)
	assert.NoError(t, err)
	assert.NotEqual(t, checksum1, checksum2)
}

func TestBackupService_FileNaming(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	service := NewBackupService(logger, tempDir, 30, false)

	request := &BackupRequest{
		BusinessID:  "test-business-123",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	response, err := service.CreateBackup(ctx, request)
	require.NoError(t, err)

	// Verify file naming convention
	expectedPrefix := "backup_" + response.BackupID + "_test-business-123_"
	assert.Contains(t, filepath.Base(response.FilePath), expectedPrefix)
	assert.True(t, filepath.Ext(response.FilePath) == ".json" || filepath.Ext(response.FilePath) == ".gz")
}

func TestBackupService_ParseBackupFileName(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	service := NewBackupService(logger, tempDir, 30, false)

	// Test valid file name
	fileName := "backup_full_test-business-123_20241219_143022.json"
	backupInfo, err := service.parseBackupFileName(fileName)
	assert.NoError(t, err)
	assert.NotNil(t, backupInfo)
	assert.Equal(t, "full_test-business-123_20241219_143022", backupInfo.BackupID)
	assert.Equal(t, "test-business-123", backupInfo.BusinessID)
	assert.Equal(t, BackupTypeFull, backupInfo.BackupType)

	// Test invalid file names
	invalidNames := []string{
		"invalid_file.txt",
		"backup_invalid.json",
		"backup_full_business_20241219.json",
		"not_backup_file.json",
	}

	for _, invalidName := range invalidNames {
		t.Run("invalid_"+invalidName, func(t *testing.T) {
			_, err := service.parseBackupFileName(invalidName)
			assert.Error(t, err)
		})
	}
}

func TestBackupService_BackupDirectoryCreation(t *testing.T) {
	// Test with non-existent directory
	tempDir := t.TempDir()
	backupDir := filepath.Join(tempDir, "non-existent", "backup", "dir")
	logger := zap.NewNop()
	service := NewBackupService(logger, backupDir, 30, false)

	request := &BackupRequest{
		BusinessID:  "test-business-123",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	response, err := service.CreateBackup(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.DirExists(t, backupDir)
	assert.FileExists(t, response.FilePath)
}

func TestBackupService_ConcurrentBackups(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	service := NewBackupService(logger, tempDir, 30, false)

	// Create multiple concurrent backups
	results := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func(i int) {
			request := &BackupRequest{
				BusinessID:  fmt.Sprintf("test-business-%d", i),
				BackupType:  BackupTypeFull,
				IncludeData: []BackupDataType{BackupDataTypeAssessments},
			}

			ctx := context.WithValue(context.Background(), "request_id", fmt.Sprintf("test-request-%d", i))
			_, err := service.CreateBackup(ctx, request)
			results <- err
		}(i)
	}

	// Wait for all backups to complete
	for i := 0; i < 5; i++ {
		err := <-results
		assert.NoError(t, err)
	}

	// Verify all backups were created
	backups, err := service.ListBackups("")
	assert.NoError(t, err)
	assert.Len(t, backups, 5)
}
