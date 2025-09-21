package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUnifiedAuditTestDB(t *testing.T) *sql.DB {
	// Use test database connection
	db, err := sql.Open("postgres", os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Skipf("Skipping test: failed to connect to test database: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: failed to ping test database: %v", err)
	}

	return db
}

func TestUnifiedAuditRepository_SaveAuditLog(t *testing.T) {
	db := setupUnifiedAuditTestDB(t)
	defer db.Close()

	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	repo := NewUnifiedAuditRepository(db, logger)

	ctx := context.Background()

	tests := []struct {
		name      string
		auditLog  *models.UnifiedAuditLog
		wantError bool
	}{
		{
			name: "valid audit log",
			auditLog: func() *models.UnifiedAuditLog {
				auditLog := models.NewUnifiedAuditLog()
				auditLog.SetEventType(models.EventTypeUserAction)
				auditLog.SetEventCategory(models.EventCategoryAudit)
				auditLog.SetAction(models.ActionCreate)
				auditLog.SetUserContext("user-123", "api-key-456")
				auditLog.SetBusinessContext("merchant-789", "session-101")
				auditLog.SetResourceInfo("merchant", "merchant-789", "merchants")
				auditLog.SetRequestContext("req-123", "192.168.1.1", "Mozilla/5.0")

				details := map[string]interface{}{
					"description": "Created new merchant",
					"fields":      []string{"name", "industry", "status"},
				}
				auditLog.SetChangeTracking(nil, details, details)

				metadata := map[string]interface{}{
					"source":  "api",
					"version": "1.0",
				}
				auditLog.SetMetadata(metadata)

				return auditLog
			}(),
			wantError: false,
		},
		{
			name: "minimal audit log",
			auditLog: func() *models.UnifiedAuditLog {
				auditLog := models.NewUnifiedAuditLog()
				auditLog.SetEventType(models.EventTypeSystemEvent)
				auditLog.SetEventCategory(models.EventCategorySystem)
				auditLog.SetAction(models.ActionUpdate)
				return auditLog
			}(),
			wantError: false,
		},
		{
			name: "invalid event type",
			auditLog: func() *models.UnifiedAuditLog {
				auditLog := models.NewUnifiedAuditLog()
				auditLog.EventType = "invalid_type"
				auditLog.SetEventCategory(models.EventCategoryAudit)
				auditLog.SetAction(models.ActionCreate)
				return auditLog
			}(),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.SaveAuditLog(ctx, tt.auditLog)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the audit log was saved
				saved, err := repo.GetAuditLogByID(ctx, tt.auditLog.ID)
				require.NoError(t, err)
				assert.Equal(t, tt.auditLog.ID, saved.ID)
				assert.Equal(t, tt.auditLog.EventType, saved.EventType)
				assert.Equal(t, tt.auditLog.EventCategory, saved.EventCategory)
				assert.Equal(t, tt.auditLog.Action, saved.Action)
			}
		})
	}
}

func TestUnifiedAuditRepository_GetAuditLogs(t *testing.T) {
	db := setupUnifiedAuditTestDB(t)
	defer db.Close()

	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	repo := NewUnifiedAuditRepository(db, logger)

	ctx := context.Background()

	// Create test audit logs
	testAuditLogs := []*models.UnifiedAuditLog{
		func() *models.UnifiedAuditLog {
			auditLog := models.NewUnifiedAuditLog()
			auditLog.SetEventType(models.EventTypeUserAction)
			auditLog.SetEventCategory(models.EventCategoryAudit)
			auditLog.SetAction(models.ActionCreate)
			auditLog.SetUserContext("user-1", "")
			auditLog.SetBusinessContext("merchant-1", "")
			return auditLog
		}(),
		func() *models.UnifiedAuditLog {
			auditLog := models.NewUnifiedAuditLog()
			auditLog.SetEventType(models.EventTypeSystemEvent)
			auditLog.SetEventCategory(models.EventCategorySystem)
			auditLog.SetAction(models.ActionUpdate)
			auditLog.SetUserContext("user-2", "")
			auditLog.SetBusinessContext("merchant-2", "")
			return auditLog
		}(),
		func() *models.UnifiedAuditLog {
			auditLog := models.NewUnifiedAuditLog()
			auditLog.SetEventType(models.EventTypeMerchantOperation)
			auditLog.SetEventCategory(models.EventCategoryMerchant)
			auditLog.SetAction(models.ActionDelete)
			auditLog.SetUserContext("user-1", "")
			auditLog.SetBusinessContext("merchant-1", "")
			return auditLog
		}(),
	}

	// Save test audit logs
	for _, auditLog := range testAuditLogs {
		err := repo.SaveAuditLog(ctx, auditLog)
		require.NoError(t, err)
	}

	tests := []struct {
		name     string
		filters  *models.UnifiedAuditLogFilters
		expected int
	}{
		{
			name:     "no filters",
			filters:  &models.UnifiedAuditLogFilters{},
			expected: 3,
		},
		{
			name: "filter by user",
			filters: &models.UnifiedAuditLogFilters{
				UserID: stringPtr("user-1"),
			},
			expected: 2,
		},
		{
			name: "filter by merchant",
			filters: &models.UnifiedAuditLogFilters{
				MerchantID: stringPtr("merchant-1"),
			},
			expected: 2,
		},
		{
			name: "filter by action",
			filters: &models.UnifiedAuditLogFilters{
				Action: stringPtr(string(models.ActionCreate)),
			},
			expected: 1,
		},
		{
			name: "filter by event type",
			filters: &models.UnifiedAuditLogFilters{
				EventType: stringPtr(string(models.EventTypeUserAction)),
			},
			expected: 1,
		},
		{
			name: "filter by event category",
			filters: &models.UnifiedAuditLogFilters{
				EventCategory: stringPtr(string(models.EventCategorySystem)),
			},
			expected: 1,
		},
		{
			name: "combined filters",
			filters: &models.UnifiedAuditLogFilters{
				UserID: stringPtr("user-1"),
				Action: stringPtr(string(models.ActionCreate)),
			},
			expected: 1,
		},
		{
			name: "pagination",
			filters: &models.UnifiedAuditLogFilters{
				Limit:  2,
				Offset: 0,
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetAuditLogs(ctx, tt.filters)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, len(result.AuditLogs))
			assert.Equal(t, int64(3), result.Total) // Total should be 3 regardless of filters
		})
	}
}

func TestUnifiedAuditRepository_GetAuditTrail(t *testing.T) {
	db := setupUnifiedAuditTestDB(t)
	defer db.Close()

	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	repo := NewUnifiedAuditRepository(db, logger)

	ctx := context.Background()

	merchantID := "merchant-test-123"

	// Create test audit logs for the merchant
	testAuditLogs := []*models.UnifiedAuditLog{
		func() *models.UnifiedAuditLog {
			auditLog := models.NewUnifiedAuditLog()
			auditLog.SetEventType(models.EventTypeMerchantOperation)
			auditLog.SetEventCategory(models.EventCategoryMerchant)
			auditLog.SetAction(models.ActionCreate)
			auditLog.SetBusinessContext(merchantID, "")
			return auditLog
		}(),
		func() *models.UnifiedAuditLog {
			auditLog := models.NewUnifiedAuditLog()
			auditLog.SetEventType(models.EventTypeMerchantOperation)
			auditLog.SetEventCategory(models.EventCategoryMerchant)
			auditLog.SetAction(models.ActionUpdate)
			auditLog.SetBusinessContext(merchantID, "")
			return auditLog
		}(),
		func() *models.UnifiedAuditLog {
			auditLog := models.NewUnifiedAuditLog()
			auditLog.SetEventType(models.EventTypeMerchantOperation)
			auditLog.SetEventCategory(models.EventCategoryMerchant)
			auditLog.SetAction(models.ActionDelete)
			auditLog.SetBusinessContext(merchantID, "")
			return auditLog
		}(),
	}

	// Save test audit logs
	for _, auditLog := range testAuditLogs {
		err := repo.SaveAuditLog(ctx, auditLog)
		require.NoError(t, err)
	}

	// Test getting audit trail
	auditTrail, err := repo.GetAuditTrail(ctx, merchantID, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 3, len(auditTrail))

	// Verify all audit logs are for the correct merchant
	for _, log := range auditTrail {
		assert.Equal(t, merchantID, *log.MerchantID)
	}

	// Test pagination
	auditTrailPage1, err := repo.GetAuditTrail(ctx, merchantID, 2, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, len(auditTrailPage1))

	auditTrailPage2, err := repo.GetAuditTrail(ctx, merchantID, 2, 2)
	require.NoError(t, err)
	assert.Equal(t, 1, len(auditTrailPage2))
}

func TestUnifiedAuditRepository_GetAuditLogsByUser(t *testing.T) {
	db := setupUnifiedAuditTestDB(t)
	defer db.Close()

	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	repo := NewUnifiedAuditRepository(db, logger)

	ctx := context.Background()

	userID := "user-test-123"

	// Create test audit logs for the user
	testAuditLogs := []*models.UnifiedAuditLog{
		func() *models.UnifiedAuditLog {
			auditLog := models.NewUnifiedAuditLog()
			auditLog.SetEventType(models.EventTypeUserAction)
			auditLog.SetEventCategory(models.EventCategoryUser)
			auditLog.SetAction(models.ActionLogin)
			auditLog.SetUserContext(userID, "")
			return auditLog
		}(),
		func() *models.UnifiedAuditLog {
			auditLog := models.NewUnifiedAuditLog()
			auditLog.SetEventType(models.EventTypeUserAction)
			auditLog.SetEventCategory(models.EventCategoryUser)
			auditLog.SetAction(models.ActionLogout)
			auditLog.SetUserContext(userID, "")
			return auditLog
		}(),
	}

	// Save test audit logs
	for _, auditLog := range testAuditLogs {
		err := repo.SaveAuditLog(ctx, auditLog)
		require.NoError(t, err)
	}

	// Test getting audit logs by user
	userAuditLogs, err := repo.GetAuditLogsByUser(ctx, userID, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, len(userAuditLogs))

	// Verify all audit logs are for the correct user
	for _, log := range userAuditLogs {
		assert.Equal(t, userID, *log.UserID)
	}
}

func TestUnifiedAuditRepository_GetAuditLogsByAction(t *testing.T) {
	db := setupUnifiedAuditTestDB(t)
	defer db.Close()

	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	repo := NewUnifiedAuditRepository(db, logger)

	ctx := context.Background()

	action := string(models.ActionCreate)

	// Create test audit logs with the action
	testAuditLogs := []*models.UnifiedAuditLog{
		func() *models.UnifiedAuditLog {
			auditLog := models.NewUnifiedAuditLog()
			auditLog.SetEventType(models.EventTypeUserAction)
			auditLog.SetEventCategory(models.EventCategoryAudit)
			auditLog.SetAction(models.ActionCreate)
			return auditLog
		}(),
		func() *models.UnifiedAuditLog {
			auditLog := models.NewUnifiedAuditLog()
			auditLog.SetEventType(models.EventTypeSystemEvent)
			auditLog.SetEventCategory(models.EventCategorySystem)
			auditLog.SetAction(models.ActionCreate)
			return auditLog
		}(),
	}

	// Save test audit logs
	for _, auditLog := range testAuditLogs {
		err := repo.SaveAuditLog(ctx, auditLog)
		require.NoError(t, err)
	}

	// Test getting audit logs by action
	actionAuditLogs, err := repo.GetAuditLogsByAction(ctx, action, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, len(actionAuditLogs))

	// Verify all audit logs have the correct action
	for _, log := range actionAuditLogs {
		assert.Equal(t, action, log.Action)
	}
}

func TestUnifiedAuditRepository_DeleteOldAuditLogs(t *testing.T) {
	db := setupUnifiedAuditTestDB(t)
	defer db.Close()

	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	repo := NewUnifiedAuditRepository(db, logger)

	ctx := context.Background()

	// Create an old audit log
	oldAuditLog := models.NewUnifiedAuditLog()
	oldAuditLog.SetEventType(models.EventTypeUserAction)
	oldAuditLog.SetEventCategory(models.EventCategoryAudit)
	oldAuditLog.SetAction(models.ActionCreate)
	oldAuditLog.CreatedAt = time.Now().Add(-25 * time.Hour) // 25 hours ago

	err := repo.SaveAuditLog(ctx, oldAuditLog)
	require.NoError(t, err)

	// Create a recent audit log
	recentAuditLog := models.NewUnifiedAuditLog()
	recentAuditLog.SetEventType(models.EventTypeUserAction)
	recentAuditLog.SetEventCategory(models.EventCategoryAudit)
	recentAuditLog.SetAction(models.ActionUpdate)
	recentAuditLog.CreatedAt = time.Now().Add(-1 * time.Hour) // 1 hour ago

	err = repo.SaveAuditLog(ctx, recentAuditLog)
	require.NoError(t, err)

	// Delete audit logs older than 24 hours
	deletedCount, err := repo.DeleteOldAuditLogs(ctx, 24*time.Hour)
	require.NoError(t, err)
	assert.Equal(t, int64(1), deletedCount)

	// Verify the old audit log was deleted
	_, err = repo.GetAuditLogByID(ctx, oldAuditLog.ID)
	assert.Error(t, err)

	// Verify the recent audit log still exists
	_, err = repo.GetAuditLogByID(ctx, recentAuditLog.ID)
	assert.NoError(t, err)
}

func TestUnifiedAuditRepository_GetAuditLogByID(t *testing.T) {
	db := setupUnifiedAuditTestDB(t)
	defer db.Close()

	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	repo := NewUnifiedAuditRepository(db, logger)

	ctx := context.Background()

	// Create a test audit log
	auditLog := models.NewUnifiedAuditLog()
	auditLog.SetEventType(models.EventTypeUserAction)
	auditLog.SetEventCategory(models.EventCategoryAudit)
	auditLog.SetAction(models.ActionCreate)
	auditLog.SetUserContext("user-123", "api-key-456")
	auditLog.SetBusinessContext("merchant-789", "session-101")

	details := map[string]interface{}{
		"description": "Test audit log",
		"fields":      []string{"name", "status"},
	}
	auditLog.SetChangeTracking(nil, details, details)

	err := repo.SaveAuditLog(ctx, auditLog)
	require.NoError(t, err)

	// Test getting the audit log by ID
	retrieved, err := repo.GetAuditLogByID(ctx, auditLog.ID)
	require.NoError(t, err)
	assert.Equal(t, auditLog.ID, retrieved.ID)
	assert.Equal(t, auditLog.EventType, retrieved.EventType)
	assert.Equal(t, auditLog.EventCategory, retrieved.EventCategory)
	assert.Equal(t, auditLog.Action, retrieved.Action)
	assert.Equal(t, "user-123", *retrieved.UserID)
	assert.Equal(t, "api-key-456", *retrieved.APIKeyID)
	assert.Equal(t, "merchant-789", *retrieved.MerchantID)
	assert.Equal(t, "session-101", *retrieved.SessionID)

	// Test getting non-existent audit log
	_, err = repo.GetAuditLogByID(ctx, "non-existent-id")
	assert.Error(t, err)
}

func TestUnifiedAuditRepository_ChangeTracking(t *testing.T) {
	db := setupUnifiedAuditTestDB(t)
	defer db.Close()

	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	repo := NewUnifiedAuditRepository(db, logger)

	ctx := context.Background()

	// Test audit log with change tracking
	auditLog := models.NewUnifiedAuditLog()
	auditLog.SetEventType(models.EventTypeDataChange)
	auditLog.SetEventCategory(models.EventCategorySystem)
	auditLog.SetAction(models.ActionUpdate)

	oldValues := map[string]interface{}{
		"name":   "Old Name",
		"status": "inactive",
	}

	newValues := map[string]interface{}{
		"name":   "New Name",
		"status": "active",
	}

	details := map[string]interface{}{
		"description":    "Updated merchant information",
		"changed_fields": []string{"name", "status"},
	}

	err := auditLog.SetChangeTracking(oldValues, newValues, details)
	require.NoError(t, err)

	err = repo.SaveAuditLog(ctx, auditLog)
	require.NoError(t, err)

	// Retrieve and verify change tracking data
	retrieved, err := repo.GetAuditLogByID(ctx, auditLog.ID)
	require.NoError(t, err)

	assert.NotNil(t, retrieved.OldValues)
	assert.NotNil(t, retrieved.NewValues)
	assert.NotNil(t, retrieved.Details)

	// Parse and verify old values
	var oldData map[string]interface{}
	err = json.Unmarshal(*retrieved.OldValues, &oldData)
	require.NoError(t, err)
	assert.Equal(t, "Old Name", oldData["name"])
	assert.Equal(t, "inactive", oldData["status"])

	// Parse and verify new values
	var newData map[string]interface{}
	err = json.Unmarshal(*retrieved.NewValues, &newData)
	require.NoError(t, err)
	assert.Equal(t, "New Name", newData["name"])
	assert.Equal(t, "active", newData["status"])

	// Parse and verify details
	var detailsData map[string]interface{}
	err = json.Unmarshal(*retrieved.Details, &detailsData)
	require.NoError(t, err)
	assert.Equal(t, "Updated merchant information", detailsData["description"])
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
