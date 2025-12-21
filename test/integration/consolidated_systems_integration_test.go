//go:build !comprehensive_test && !e2e_railway
// +build !comprehensive_test,!e2e_railway

package integration

import (
	"testing"
	"time"

	"kyb-platform/internal/models"
	"kyb-platform/internal/observability"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestConsolidatedSystemsIntegration tests the integration of consolidated audit and compliance systems
func TestConsolidatedSystemsIntegration(t *testing.T) {
	// Setup test logger (unused but kept for consistency)
	_ = observability.NewLogger(zap.NewNop())

	t.Run("Unified Audit Log Validation", func(t *testing.T) {
		t.Log("Testing unified audit log validation...")

		// Test valid audit log
		validAuditLog := &models.UnifiedAuditLog{
			ID:            "audit-123",
			EventType:     "merchant_operation",
			EventCategory: "audit",
			Action:        "CREATE",
			CreatedAt:     time.Now(),
		}

		// Test validation
		err := validAuditLog.Validate()
		assert.NoError(t, err, "Valid audit log should pass validation")

		// Test invalid audit log (missing required fields)
		invalidAuditLog := &models.UnifiedAuditLog{
			ID: "audit-124",
			// Missing required fields
			CreatedAt: time.Now(),
		}

		err = invalidAuditLog.Validate()
		assert.Error(t, err, "Invalid audit log should fail validation")

		t.Log("✅ Unified audit log validation passed")
	})

	t.Run("Unified Audit Log Field Types", func(t *testing.T) {
		t.Log("Testing unified audit log field types...")

		// Test pointer field handling
		userID := "user-123"
		merchantID := "merchant-456"
		resourceType := "merchant"
		resourceID := "merchant-456"
		requestID := "req-123"
		ipAddress := "192.168.1.1"
		userAgent := "Test Agent"

		auditLog := &models.UnifiedAuditLog{
			ID:            "audit-125",
			UserID:        &userID,
			MerchantID:    &merchantID,
			EventType:     "merchant_operation",
			EventCategory: "audit",
			Action:        "CREATE",
			ResourceType:  &resourceType,
			ResourceID:    &resourceID,
			RequestID:     &requestID,
			IPAddress:     &ipAddress,
			UserAgent:     &userAgent,
			CreatedAt:     time.Now(),
		}

		// Test validation
		err := auditLog.Validate()
		assert.NoError(t, err, "Audit log with pointer fields should pass validation")

		// Test field access
		assert.Equal(t, userID, *auditLog.UserID, "UserID should match")
		assert.Equal(t, merchantID, *auditLog.MerchantID, "MerchantID should match")
		assert.Equal(t, resourceType, *auditLog.ResourceType, "ResourceType should match")
		assert.Equal(t, resourceID, *auditLog.ResourceID, "ResourceID should match")

		t.Log("✅ Unified audit log field types passed")
	})

	t.Run("Unified Audit Log Event Types", func(t *testing.T) {
		t.Log("Testing unified audit log event types...")

		// Test valid event types
		validEventTypes := []string{
			"user_action",
			"system_event",
			"api_call",
			"data_change",
			"security_event",
			"compliance_check",
			"business_operation",
			"merchant_operation",
			"classification",
			"risk_assessment",
			"verification",
			"authentication",
			"authorization",
		}

		for _, eventType := range validEventTypes {
			auditLog := &models.UnifiedAuditLog{
				ID:            "audit-" + eventType,
				EventType:     eventType,
				EventCategory: "audit",
				Action:        "CREATE",
				CreatedAt:     time.Now(),
			}

			err := auditLog.Validate()
			assert.NoError(t, err, "Event type %s should be valid", eventType)
		}

		// Test invalid event type
		invalidAuditLog := &models.UnifiedAuditLog{
			ID:            "audit-invalid",
			EventType:     "invalid_event_type",
			EventCategory: "audit",
			Action:        "CREATE",
			CreatedAt:     time.Now(),
		}

		err := invalidAuditLog.Validate()
		assert.Error(t, err, "Invalid event type should fail validation")

		t.Log("✅ Unified audit log event types passed")
	})

	t.Run("Unified Audit Log Actions", func(t *testing.T) {
		t.Log("Testing unified audit log actions...")

		// Test valid actions
		validActions := []string{
			"INSERT",
			"UPDATE",
			"DELETE",
			"CREATE",
			"READ",
			"LOGIN",
			"LOGOUT",
			"ACCESS",
			"EXPORT",
			"IMPORT",
			"VERIFY",
			"APPROVE",
			"REJECT",
			"CLASSIFY",
			"ASSESS",
			"SCAN",
			"ANALYZE",
		}

		for _, action := range validActions {
			auditLog := &models.UnifiedAuditLog{
				ID:            "audit-" + action,
				EventType:     "user_action",
				EventCategory: "audit",
				Action:        action,
				CreatedAt:     time.Now(),
			}

			err := auditLog.Validate()
			assert.NoError(t, err, "Action %s should be valid", action)
		}

		// Test invalid action
		invalidAuditLog := &models.UnifiedAuditLog{
			ID:            "audit-invalid-action",
			EventType:     "user_action",
			EventCategory: "audit",
			Action:        "INVALID_ACTION",
			CreatedAt:     time.Now(),
		}

		err := invalidAuditLog.Validate()
		assert.Error(t, err, "Invalid action should fail validation")

		t.Log("✅ Unified audit log actions passed")
	})

	t.Run("Unified Audit Log Event Categories", func(t *testing.T) {
		t.Log("Testing unified audit log event categories...")

		// Test valid event categories
		validCategories := []string{
			"audit",
			"compliance",
			"security",
			"business",
			"system",
			"user",
			"merchant",
		}

		for _, category := range validCategories {
			auditLog := &models.UnifiedAuditLog{
				ID:            "audit-" + category,
				EventType:     "user_action",
				EventCategory: category,
				Action:        "CREATE",
				CreatedAt:     time.Now(),
			}

			err := auditLog.Validate()
			assert.NoError(t, err, "Event category %s should be valid", category)
		}

		// Test invalid event category
		invalidAuditLog := &models.UnifiedAuditLog{
			ID:            "audit-invalid-category",
			EventType:     "user_action",
			EventCategory: "invalid_category",
			Action:        "CREATE",
			CreatedAt:     time.Now(),
		}

		err := invalidAuditLog.Validate()
		assert.Error(t, err, "Invalid event category should fail validation")

		t.Log("✅ Unified audit log event categories passed")
	})

	t.Run("Unified Audit Log JSON Handling", func(t *testing.T) {
		t.Log("Testing unified audit log JSON handling...")

		// Test JSON field handling
		auditLog := &models.UnifiedAuditLog{
			ID:            "audit-json-test",
			EventType:     "data_change",
			EventCategory: "audit",
			Action:        "UPDATE",
			CreatedAt:     time.Now(),
		}

		// Test setting details
		details := map[string]interface{}{
			"field":   "value",
			"number":  123,
			"boolean": true,
		}

		err := auditLog.SetChangeTracking(nil, nil, details)
		assert.NoError(t, err, "Should be able to set details")

		// Test setting metadata
		metadata := map[string]interface{}{
			"source":  "test",
			"version": "1.0",
		}

		err = auditLog.SetMetadata(metadata)
		assert.NoError(t, err, "Should be able to set metadata")

		// Test validation
		err = auditLog.Validate()
		assert.NoError(t, err, "Audit log with JSON fields should pass validation")

		t.Log("✅ Unified audit log JSON handling passed")
	})

	t.Run("Unified Audit Log Request Context", func(t *testing.T) {
		t.Log("Testing unified audit log request context...")

		auditLog := &models.UnifiedAuditLog{
			ID:            "audit-request-context",
			EventType:     "api_call",
			EventCategory: "audit",
			Action:        "READ",
			CreatedAt:     time.Now(),
		}

		// Test setting request context
		auditLog.SetRequestContext("req-123", "192.168.1.1", "Test Agent")

		// Test validation
		err := auditLog.Validate()
		assert.NoError(t, err, "Audit log with request context should pass validation")

		// Test field values
		assert.Equal(t, "req-123", *auditLog.RequestID, "RequestID should match")
		assert.Equal(t, "192.168.1.1", *auditLog.IPAddress, "IPAddress should match")
		assert.Equal(t, "Test Agent", *auditLog.UserAgent, "UserAgent should match")

		t.Log("✅ Unified audit log request context passed")
	})

	t.Run("Unified Audit Log Legacy Conversion", func(t *testing.T) {
		t.Log("Testing unified audit log legacy conversion...")

		// Create unified audit log
		userID := "user-123"
		resourceType := "merchant"
		resourceID := "merchant-456"
		ipAddress := "192.168.1.1"
		userAgent := "Test Agent"
		requestID := "req-123"

		unifiedLog := &models.UnifiedAuditLog{
			ID:            "audit-legacy-test",
			UserID:        &userID,
			EventType:     "merchant_operation",
			EventCategory: "audit",
			Action:        "CREATE",
			ResourceType:  &resourceType,
			ResourceID:    &resourceID,
			IPAddress:     &ipAddress,
			UserAgent:     &userAgent,
			RequestID:     &requestID,
			CreatedAt:     time.Now(),
		}

		// Convert to legacy format
		legacyLog := unifiedLog.ToLegacyAuditLog()

		// Test conversion
		assert.Equal(t, unifiedLog.ID, legacyLog.ID, "ID should match")
		assert.Equal(t, unifiedLog.Action, legacyLog.Action, "Action should match")
		assert.Equal(t, userID, legacyLog.UserID, "UserID should match")
		assert.Equal(t, resourceType, legacyLog.ResourceType, "ResourceType should match")
		assert.Equal(t, resourceID, legacyLog.ResourceID, "ResourceID should match")
		assert.Equal(t, ipAddress, legacyLog.IPAddress, "IPAddress should match")
		assert.Equal(t, userAgent, legacyLog.UserAgent, "UserAgent should match")
		assert.Equal(t, requestID, legacyLog.RequestID, "RequestID should match")

		t.Log("✅ Unified audit log legacy conversion passed")
	})

	t.Run("Unified Audit Log Filters", func(t *testing.T) {
		t.Log("Testing unified audit log filters...")

		// Test filter creation
		userID := "user-123"
		merchantID := "merchant-456"
		eventType := "merchant_operation"
		action := "CREATE"
		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now()

		filters := &models.UnifiedAuditLogFilters{
			UserID:     &userID,
			MerchantID: &merchantID,
			EventType:  &eventType,
			Action:     &action,
			StartDate:  &startDate,
			EndDate:    &endDate,
			Limit:      100,
			Offset:     0,
		}

		// Test filter values
		assert.Equal(t, userID, *filters.UserID, "UserID filter should match")
		assert.Equal(t, merchantID, *filters.MerchantID, "MerchantID filter should match")
		assert.Equal(t, eventType, *filters.EventType, "EventType filter should match")
		assert.Equal(t, action, *filters.Action, "Action filter should match")
		assert.Equal(t, startDate, *filters.StartDate, "StartDate filter should match")
		assert.Equal(t, endDate, *filters.EndDate, "EndDate filter should match")
		assert.Equal(t, 100, filters.Limit, "Limit should match")
		assert.Equal(t, 0, filters.Offset, "Offset should match")

		t.Log("✅ Unified audit log filters passed")
	})

	t.Run("Data Migration Validation", func(t *testing.T) {
		t.Log("Testing data migration validation...")

		// Test that the consolidated systems maintain data integrity
		// This test validates that the migration from separate audit_logs and merchant_audit_logs
		// to unified_audit_logs maintains all necessary information

		// Simulate legacy audit log data
		legacyData := map[string]interface{}{
			"id":            "legacy-audit-1",
			"user_id":       "user-123",
			"action":        "CREATE",
			"resource_type": "merchant",
			"resource_id":   "merchant-456",
			"created_at":    time.Now(),
		}

		// Test that we can create a unified audit log from legacy data
		unifiedLog := &models.UnifiedAuditLog{
			ID:            legacyData["id"].(string),
			EventType:     "merchant_operation", // Default for merchant operations
			EventCategory: "audit",              // Default category
			Action:        legacyData["action"].(string),
			CreatedAt:     legacyData["created_at"].(time.Time),
		}

		// Set optional fields if they exist
		if userID, ok := legacyData["user_id"].(string); ok && userID != "" {
			unifiedLog.UserID = &userID
		}
		if resourceType, ok := legacyData["resource_type"].(string); ok && resourceType != "" {
			unifiedLog.ResourceType = &resourceType
		}
		if resourceID, ok := legacyData["resource_id"].(string); ok && resourceID != "" {
			unifiedLog.ResourceID = &resourceID
		}

		// Validate the unified log
		err := unifiedLog.Validate()
		assert.NoError(t, err, "Migrated audit log should pass validation")

		// Test that all legacy data is preserved
		assert.Equal(t, legacyData["id"], unifiedLog.ID, "ID should be preserved")
		assert.Equal(t, legacyData["action"], unifiedLog.Action, "Action should be preserved")
		assert.Equal(t, legacyData["user_id"], *unifiedLog.UserID, "UserID should be preserved")
		assert.Equal(t, legacyData["resource_type"], *unifiedLog.ResourceType, "ResourceType should be preserved")
		assert.Equal(t, legacyData["resource_id"], *unifiedLog.ResourceID, "ResourceID should be preserved")

		t.Log("✅ Data migration validation passed")
	})

	t.Log("✅ All consolidated systems integration tests passed")
}
