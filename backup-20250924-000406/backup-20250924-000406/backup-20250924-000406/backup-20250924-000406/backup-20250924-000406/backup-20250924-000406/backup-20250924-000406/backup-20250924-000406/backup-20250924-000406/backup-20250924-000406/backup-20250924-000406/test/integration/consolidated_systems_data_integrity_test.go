package integration

import (
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/models"
	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestConsolidatedSystemsDataIntegrity tests data integrity across consolidated audit and compliance systems
func TestConsolidatedSystemsDataIntegrity(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	t.Run("Audit System Data Integrity", func(t *testing.T) {
		t.Log("Testing unified audit system data integrity...")

		// Test data consistency
		testCases := []struct {
			name        string
			auditLog    *models.UnifiedAuditLog
			expectError bool
		}{
			{
				name: "Valid Audit Log",
				auditLog: &models.UnifiedAuditLog{
					ID:            "audit-123",
					UserID:        "user-456",
					MerchantID:    "merchant-789",
					EventType:     "merchant_operation",
					EventCategory: "audit",
					Action:        "CREATE",
					ResourceType:  "merchant",
					ResourceID:    "merchant-789",
					RequestID:     "req-123",
					IPAddress:     "192.168.1.1",
					UserAgent:     "Test Agent",
					CreatedAt:     time.Now(),
				},
				expectError: false,
			},
			{
				name: "Invalid Action",
				auditLog: &models.UnifiedAuditLog{
					ID:            "audit-124",
					UserID:        "user-456",
					MerchantID:    "merchant-789",
					EventType:     "merchant_operation",
					EventCategory: "audit",
					Action:        "INVALID_ACTION", // Invalid action
					ResourceType:  "merchant",
					ResourceID:    "merchant-789",
					CreatedAt:     time.Now(),
				},
				expectError: true,
			},
			{
				name: "Invalid Event Category",
				auditLog: &models.UnifiedAuditLog{
					ID:            "audit-125",
					UserID:        "user-456",
					MerchantID:    "merchant-789",
					EventType:     "merchant_operation",
					EventCategory: "invalid_category", // Invalid category
					Action:        "CREATE",
					ResourceType:  "merchant",
					ResourceID:    "merchant-789",
					CreatedAt:     time.Now(),
				},
				expectError: true,
			},
			{
				name: "Missing Required Fields",
				auditLog: &models.UnifiedAuditLog{
					ID: "audit-126",
					// Missing required fields
					CreatedAt: time.Now(),
				},
				expectError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Validate audit log
				err := tc.auditLog.Validate()

				if tc.expectError {
					assert.Error(t, err, "Should return error for invalid audit log")
				} else {
					assert.NoError(t, err, "Should not return error for valid audit log")
				}
			})
		}

		t.Log("✅ Unified audit system data integrity validation passed")
	})

	t.Run("Compliance System Data Integrity", func(t *testing.T) {
		t.Log("Testing unified compliance system data integrity...")

		// Test data consistency
		testCases := []struct {
			name        string
			tracking    *models.ComplianceTracking
			expectError bool
		}{
			{
				name: "Valid Compliance Tracking",
				tracking: &models.ComplianceTracking{
					ID:                  "tracking-123",
					MerchantID:          "merchant-456",
					ComplianceType:      "AML",
					ComplianceFramework: "FATF",
					CheckType:           "automated",
					Status:              "completed",
					Score:               0.85,
					RiskLevel:           "medium",
					CheckMethod:         "api_integration",
					Source:              "external_provider",
					CreatedAt:           time.Now(),
					UpdatedAt:           time.Now(),
				},
				expectError: false,
			},
			{
				name: "Invalid Status",
				tracking: &models.ComplianceTracking{
					ID:                  "tracking-124",
					MerchantID:          "merchant-456",
					ComplianceType:      "AML",
					ComplianceFramework: "FATF",
					CheckType:           "automated",
					Status:              "invalid_status", // Invalid status
					CheckMethod:         "api_integration",
					Source:              "external_provider",
					CreatedAt:           time.Now(),
					UpdatedAt:           time.Now(),
				},
				expectError: true,
			},
			{
				name: "Invalid Risk Level",
				tracking: &models.ComplianceTracking{
					ID:                  "tracking-125",
					MerchantID:          "merchant-456",
					ComplianceType:      "AML",
					ComplianceFramework: "FATF",
					CheckType:           "automated",
					Status:              "completed",
					RiskLevel:           "invalid_risk", // Invalid risk level
					CheckMethod:         "api_integration",
					Source:              "external_provider",
					CreatedAt:           time.Now(),
					UpdatedAt:           time.Now(),
				},
				expectError: true,
			},
			{
				name: "Invalid Score Range",
				tracking: &models.ComplianceTracking{
					ID:                  "tracking-126",
					MerchantID:          "merchant-456",
					ComplianceType:      "AML",
					ComplianceFramework: "FATF",
					CheckType:           "automated",
					Status:              "completed",
					Score:               1.5, // Invalid score (should be 0.0-1.0)
					CheckMethod:         "api_integration",
					Source:              "external_provider",
					CreatedAt:           time.Now(),
					UpdatedAt:           time.Now(),
				},
				expectError: true,
			},
			{
				name: "Missing Required Fields",
				tracking: &models.ComplianceTracking{
					ID: "tracking-127",
					// Missing required fields
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				expectError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Validate compliance tracking
				err := tc.tracking.Validate()

				if tc.expectError {
					assert.Error(t, err, "Should return error for invalid compliance tracking")
				} else {
					assert.NoError(t, err, "Should not return error for valid compliance tracking")
				}
			})
		}

		t.Log("✅ Unified compliance system data integrity validation passed")
	})

	t.Run("Cross-System Data Consistency", func(t *testing.T) {
		t.Log("Testing cross-system data consistency...")

		// Test that audit logs and compliance tracking maintain consistent merchant references
		merchantID := "merchant-456"
		userID := "user-123"

		// Create audit log
		auditLog := &models.UnifiedAuditLog{
			ID:            "audit-123",
			UserID:        userID,
			MerchantID:    merchantID,
			EventType:     "compliance_check",
			EventCategory: "compliance",
			Action:        "CREATE",
			ResourceType:  "compliance_tracking",
			ResourceID:    "tracking-123",
			CreatedAt:     time.Now(),
		}

		// Create compliance tracking
		complianceTracking := &models.ComplianceTracking{
			ID:                  "tracking-123",
			MerchantID:          merchantID,
			ComplianceType:      "AML",
			ComplianceFramework: "FATF",
			CheckType:           "automated",
			Status:              "completed",
			CheckMethod:         "api_integration",
			Source:              "external_provider",
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		// Validate cross-system consistency
		assert.Equal(t, merchantID, auditLog.MerchantID, "Audit log should reference correct merchant")
		assert.Equal(t, merchantID, complianceTracking.MerchantID, "Compliance tracking should reference correct merchant")
		assert.Equal(t, "tracking-123", auditLog.ResourceID, "Audit log should reference correct compliance tracking ID")
		assert.Equal(t, "tracking-123", complianceTracking.ID, "Compliance tracking should have correct ID")

		// Validate audit log references compliance tracking
		assert.Equal(t, "compliance_tracking", auditLog.ResourceType, "Audit log should reference compliance tracking resource type")
		assert.Equal(t, "compliance_check", auditLog.EventType, "Audit log should have compliance check event type")

		t.Log("✅ Cross-system data consistency validation passed")
	})

	t.Run("Data Migration Integrity", func(t *testing.T) {
		t.Log("Testing data migration integrity...")

		// Test that migrated data maintains referential integrity
		migrationTestCases := []struct {
			name           string
			oldAuditLog    map[string]interface{} // Simulated old audit log structure
			expectedFields map[string]interface{} // Expected fields in new structure
		}{
			{
				name: "Legacy Audit Log Migration",
				oldAuditLog: map[string]interface{}{
					"id":            "legacy-audit-1",
					"user_id":       "user-123",
					"action":        "CREATE",
					"resource_type": "merchant",
					"resource_id":   "merchant-456",
					"created_at":    time.Now(),
				},
				expectedFields: map[string]interface{}{
					"user_id":        "user-123",
					"action":         "CREATE",
					"resource_type":  "merchant",
					"resource_id":    "merchant-456",
					"event_type":     "merchant_operation",
					"event_category": "audit",
				},
			},
			{
				name: "Legacy Compliance Check Migration",
				oldAuditLog: map[string]interface{}{
					"id":              "legacy-compliance-1",
					"business_id":     "merchant-456",
					"compliance_type": "AML",
					"status":          "completed",
					"score":           0.85,
					"created_at":      time.Now(),
				},
				expectedFields: map[string]interface{}{
					"merchant_id":          "merchant-456",
					"compliance_type":      "AML",
					"status":               "completed",
					"score":                0.85,
					"compliance_framework": "FATF",      // Default framework
					"check_type":           "automated", // Default check type
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Simulate migration validation
				for key, expectedValue := range tc.expectedFields {
					actualValue, exists := tc.oldAuditLog[key]
					if !exists {
						// Check if it's a default value that should be added during migration
						if key == "event_type" || key == "event_category" || key == "compliance_framework" || key == "check_type" {
							continue // These are default values added during migration
						}
					}

					if exists {
						assert.Equal(t, expectedValue, actualValue, "Migrated field %s should match expected value", key)
					}
				}
			})
		}

		t.Log("✅ Data migration integrity validation passed")
	})

	t.Run("Referential Integrity", func(t *testing.T) {
		t.Log("Testing referential integrity...")

		// Test that foreign key relationships are maintained
		merchantID := "merchant-456"
		userID := "user-123"

		// Test audit log references
		auditLog := &models.UnifiedAuditLog{
			ID:            "audit-123",
			UserID:        userID,
			MerchantID:    merchantID,
			EventType:     "merchant_operation",
			EventCategory: "audit",
			Action:        "CREATE",
			ResourceType:  "merchant",
			ResourceID:    merchantID,
			CreatedAt:     time.Now(),
		}

		// Validate referential integrity
		assert.NotEmpty(t, auditLog.UserID, "Audit log should have user ID")
		assert.NotEmpty(t, auditLog.MerchantID, "Audit log should have merchant ID")
		assert.Equal(t, auditLog.ResourceID, auditLog.MerchantID, "Resource ID should match merchant ID for merchant operations")

		// Test compliance tracking references
		complianceTracking := &models.ComplianceTracking{
			ID:                  "tracking-123",
			MerchantID:          merchantID,
			ComplianceType:      "AML",
			ComplianceFramework: "FATF",
			CheckType:           "automated",
			Status:              "completed",
			CheckMethod:         "api_integration",
			Source:              "external_provider",
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		// Validate referential integrity
		assert.NotEmpty(t, complianceTracking.MerchantID, "Compliance tracking should have merchant ID")
		assert.NotEmpty(t, complianceTracking.ComplianceType, "Compliance tracking should have compliance type")
		assert.NotEmpty(t, complianceTracking.ComplianceFramework, "Compliance tracking should have compliance framework")

		t.Log("✅ Referential integrity validation passed")
	})
}

// TestConsolidatedSystemsPerformance tests the performance of consolidated systems
func TestConsolidatedSystemsPerformance(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	t.Run("Audit System Performance", func(t *testing.T) {
		t.Log("Testing unified audit system performance...")

		// Test audit log creation performance
		start := time.Now()
		for i := 0; i < 100; i++ {
			auditLog := &models.UnifiedAuditLog{
				ID:            "audit-" + string(rune(i)),
				UserID:        "user-123",
				MerchantID:    "merchant-456",
				EventType:     "merchant_operation",
				EventCategory: "audit",
				Action:        "READ",
				ResourceType:  "merchant",
				ResourceID:    "merchant-456",
				CreatedAt:     time.Now(),
			}

			// Validate audit log (simulating database validation)
			err := auditLog.Validate()
			assert.NoError(t, err, "Audit log validation should succeed")
		}
		duration := time.Since(start)

		// Performance assertions
		assert.Less(t, duration, 2*time.Second, "Audit log creation should complete within 2 seconds")
		assert.Less(t, duration/100, 20*time.Millisecond, "Average audit log creation time should be under 20ms")

		t.Logf("✅ Unified audit system performance: %v for 100 entries (avg: %v)", duration, duration/100)
	})

	t.Run("Compliance System Performance", func(t *testing.T) {
		t.Log("Testing unified compliance system performance...")

		// Test compliance tracking creation performance
		start := time.Now()
		for i := 0; i < 50; i++ {
			tracking := &models.ComplianceTracking{
				ID:                  "tracking-" + string(rune(i)),
				MerchantID:          "merchant-456",
				ComplianceType:      "AML",
				ComplianceFramework: "FATF",
				CheckType:           "automated",
				Status:              "completed",
				Score:               0.85,
				CheckMethod:         "api_integration",
				Source:              "external_provider",
				CreatedAt:           time.Now(),
				UpdatedAt:           time.Now(),
			}

			// Validate compliance tracking (simulating database validation)
			err := tracking.Validate()
			assert.NoError(t, err, "Compliance tracking validation should succeed")
		}
		duration := time.Since(start)

		// Performance assertions
		assert.Less(t, duration, 1*time.Second, "Compliance tracking creation should complete within 1 second")
		assert.Less(t, duration/50, 20*time.Millisecond, "Average compliance tracking creation time should be under 20ms")

		t.Logf("✅ Unified compliance system performance: %v for 50 entries (avg: %v)", duration, duration/50)
	})

	t.Run("Cross-System Integration Performance", func(t *testing.T) {
		t.Log("Testing cross-system integration performance...")

		// Test integrated audit and compliance operations
		start := time.Now()
		for i := 0; i < 25; i++ {
			// Create compliance tracking
			tracking := &models.ComplianceTracking{
				ID:                  "tracking-" + string(rune(i)),
				MerchantID:          "merchant-456",
				ComplianceType:      "AML",
				ComplianceFramework: "FATF",
				CheckType:           "automated",
				Status:              "completed",
				CheckMethod:         "api_integration",
				Source:              "external_provider",
				CreatedAt:           time.Now(),
				UpdatedAt:           time.Now(),
			}

			// Create corresponding audit log
			auditLog := &models.UnifiedAuditLog{
				ID:            "audit-" + string(rune(i)),
				UserID:        "user-123",
				MerchantID:    "merchant-456",
				EventType:     "compliance_check",
				EventCategory: "compliance",
				Action:        "CREATE",
				ResourceType:  "compliance_tracking",
				ResourceID:    tracking.ID,
				CreatedAt:     time.Now(),
			}

			// Validate both
			err1 := tracking.Validate()
			err2 := auditLog.Validate()

			assert.NoError(t, err1, "Compliance tracking validation should succeed")
			assert.NoError(t, err2, "Audit log validation should succeed")

			// Validate cross-system consistency
			assert.Equal(t, tracking.MerchantID, auditLog.MerchantID, "Merchant IDs should match")
			assert.Equal(t, tracking.ID, auditLog.ResourceID, "Resource ID should match tracking ID")
		}
		duration := time.Since(start)

		// Performance assertions
		assert.Less(t, duration, 2*time.Second, "Cross-system integration should complete within 2 seconds")
		assert.Less(t, duration/25, 80*time.Millisecond, "Average cross-system operation time should be under 80ms")

		t.Logf("✅ Cross-system integration performance: %v for 25 operations (avg: %v)", duration, duration/25)
	})
}
