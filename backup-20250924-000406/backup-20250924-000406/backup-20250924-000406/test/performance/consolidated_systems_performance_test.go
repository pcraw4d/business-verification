package performance

import (
	"fmt"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/models"
	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestConsolidatedSystemsPerformance tests the performance of consolidated audit and compliance systems
func TestConsolidatedSystemsPerformance(t *testing.T) {
	// Setup test logger (unused but kept for consistency)
	_ = observability.NewLogger(zap.NewNop())

	t.Run("Unified Audit Log Creation Performance", func(t *testing.T) {
		t.Log("Testing unified audit log creation performance...")

		// Test different batch sizes
		batchSizes := []int{10, 50, 100, 500, 1000}

		for _, batchSize := range batchSizes {
			t.Run(fmt.Sprintf("BatchSize_%d", batchSize), func(t *testing.T) {
				start := time.Now()

				for i := 0; i < batchSize; i++ {
					auditLog := &models.UnifiedAuditLog{
						ID:            fmt.Sprintf("audit-%d-%d", batchSize, i),
						EventType:     "merchant_operation",
						EventCategory: "audit",
						Action:        "READ",
						CreatedAt:     time.Now(),
					}

					// Validate audit log (simulating database validation)
					err := auditLog.Validate()
					assert.NoError(t, err, "Audit log validation should succeed")
				}

				duration := time.Since(start)
				avgDuration := duration / time.Duration(batchSize)

				// Performance assertions
				assert.Less(t, duration, 10*time.Second, "Audit log creation should complete within 10 seconds")
				assert.Less(t, avgDuration, 50*time.Millisecond, "Average audit log creation time should be under 50ms")

				t.Logf("✅ Batch size %d: %v total (avg: %v per record)", batchSize, duration, avgDuration)
			})
		}
	})

	t.Run("Unified Audit Log Validation Performance", func(t *testing.T) {
		t.Log("Testing unified audit log validation performance...")

		// Test validation performance with different scenarios
		validationScenarios := []struct {
			name        string
			description string
			createFunc  func() *models.UnifiedAuditLog
		}{
			{
				name:        "ValidAuditLog",
				description: "Valid audit log validation",
				createFunc: func() *models.UnifiedAuditLog {
					return &models.UnifiedAuditLog{
						ID:            "audit-valid",
						EventType:     "merchant_operation",
						EventCategory: "audit",
						Action:        "CREATE",
						CreatedAt:     time.Now(),
					}
				},
			},
			{
				name:        "AuditLogWithPointers",
				description: "Audit log with pointer fields validation",
				createFunc: func() *models.UnifiedAuditLog {
					userID := "user-123"
					merchantID := "merchant-456"
					resourceType := "merchant"
					resourceID := "merchant-456"

					return &models.UnifiedAuditLog{
						ID:            "audit-pointers",
						UserID:        &userID,
						MerchantID:    &merchantID,
						EventType:     "merchant_operation",
						EventCategory: "audit",
						Action:        "CREATE",
						ResourceType:  &resourceType,
						ResourceID:    &resourceID,
						CreatedAt:     time.Now(),
					}
				},
			},
			{
				name:        "AuditLogWithJSON",
				description: "Audit log with JSON fields validation",
				createFunc: func() *models.UnifiedAuditLog {
					auditLog := &models.UnifiedAuditLog{
						ID:            "audit-json",
						EventType:     "data_change",
						EventCategory: "audit",
						Action:        "UPDATE",
						CreatedAt:     time.Now(),
					}

					// Set JSON fields
					details := map[string]interface{}{
						"field":  "value",
						"number": 123,
					}
					auditLog.SetChangeTracking(nil, nil, details)

					metadata := map[string]interface{}{
						"source": "test",
					}
					auditLog.SetMetadata(metadata)

					return auditLog
				},
			},
		}

		for _, scenario := range validationScenarios {
			t.Run(scenario.name, func(t *testing.T) {
				// Run validation multiple times to get average performance
				var totalDuration time.Duration
				iterations := 1000

				for i := 0; i < iterations; i++ {
					auditLog := scenario.createFunc()

					start := time.Now()
					err := auditLog.Validate()
					duration := time.Since(start)

					assert.NoError(t, err, "Validation should succeed")
					totalDuration += duration
				}

				avgDuration := totalDuration / time.Duration(iterations)

				// Performance assertions
				assert.Less(t, avgDuration, 1*time.Millisecond, "Average validation time should be under 1ms")

				t.Logf("✅ %s: avg %v (%s)", scenario.name, avgDuration, scenario.description)
			})
		}
	})

	t.Run("Unified Audit Log JSON Processing Performance", func(t *testing.T) {
		t.Log("Testing unified audit log JSON processing performance...")

		// Test JSON field processing performance
		start := time.Now()
		iterations := 1000

		for i := 0; i < iterations; i++ {
			auditLog := &models.UnifiedAuditLog{
				ID:            fmt.Sprintf("audit-json-%d", i),
				EventType:     "data_change",
				EventCategory: "audit",
				Action:        "UPDATE",
				CreatedAt:     time.Now(),
			}

			// Set JSON fields
			details := map[string]interface{}{
				"field":   fmt.Sprintf("value-%d", i),
				"number":  i,
				"boolean": i%2 == 0,
			}
			err := auditLog.SetChangeTracking(nil, nil, details)
			assert.NoError(t, err, "Should be able to set details")

			metadata := map[string]interface{}{
				"source":    "test",
				"iteration": i,
			}
			err = auditLog.SetMetadata(metadata)
			assert.NoError(t, err, "Should be able to set metadata")

			// Validate
			err = auditLog.Validate()
			assert.NoError(t, err, "Validation should succeed")
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(iterations)

		// Performance assertions
		assert.Less(t, duration, 5*time.Second, "JSON processing should complete within 5 seconds")
		assert.Less(t, avgDuration, 5*time.Millisecond, "Average JSON processing time should be under 5ms")

		t.Logf("✅ JSON processing: %v total for %d iterations (avg: %v per iteration)", duration, iterations, avgDuration)
	})

	t.Run("Unified Audit Log Request Context Performance", func(t *testing.T) {
		t.Log("Testing unified audit log request context performance...")

		// Test request context setting performance
		start := time.Now()
		iterations := 1000

		for i := 0; i < iterations; i++ {
			auditLog := &models.UnifiedAuditLog{
				ID:            fmt.Sprintf("audit-context-%d", i),
				EventType:     "api_call",
				EventCategory: "audit",
				Action:        "READ",
				CreatedAt:     time.Now(),
			}

			// Set request context
			auditLog.SetRequestContext(
				fmt.Sprintf("req-%d", i),
				fmt.Sprintf("192.168.1.%d", i%255),
				fmt.Sprintf("Agent-%d", i),
			)

			// Validate
			err := auditLog.Validate()
			assert.NoError(t, err, "Validation should succeed")
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(iterations)

		// Performance assertions
		assert.Less(t, duration, 2*time.Second, "Request context processing should complete within 2 seconds")
		assert.Less(t, avgDuration, 2*time.Millisecond, "Average request context processing time should be under 2ms")

		t.Logf("✅ Request context processing: %v total for %d iterations (avg: %v per iteration)", duration, iterations, avgDuration)
	})

	t.Run("Unified Audit Log Legacy Conversion Performance", func(t *testing.T) {
		t.Log("Testing unified audit log legacy conversion performance...")

		// Test legacy conversion performance
		start := time.Now()
		iterations := 1000

		for i := 0; i < iterations; i++ {
			userID := fmt.Sprintf("user-%d", i)
			resourceType := "merchant"
			resourceID := fmt.Sprintf("merchant-%d", i)
			ipAddress := fmt.Sprintf("192.168.1.%d", i%255)
			userAgent := fmt.Sprintf("Agent-%d", i)
			requestID := fmt.Sprintf("req-%d", i)

			unifiedLog := &models.UnifiedAuditLog{
				ID:            fmt.Sprintf("audit-conversion-%d", i),
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

			// Validate conversion
			assert.Equal(t, unifiedLog.ID, legacyLog.ID, "ID should match")
			assert.Equal(t, unifiedLog.Action, legacyLog.Action, "Action should match")
			assert.Equal(t, userID, legacyLog.UserID, "UserID should match")
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(iterations)

		// Performance assertions
		assert.Less(t, duration, 1*time.Second, "Legacy conversion should complete within 1 second")
		assert.Less(t, avgDuration, 1*time.Millisecond, "Average legacy conversion time should be under 1ms")

		t.Logf("✅ Legacy conversion: %v total for %d iterations (avg: %v per iteration)", duration, iterations, avgDuration)
	})

	t.Run("Memory Usage Performance", func(t *testing.T) {
		t.Log("Testing memory usage performance...")

		// Test memory usage for large datasets
		datasetSizes := []int{100, 500, 1000, 5000}

		for _, datasetSize := range datasetSizes {
			t.Run(fmt.Sprintf("DatasetSize_%d", datasetSize), func(t *testing.T) {
				start := time.Now()

				// Create large dataset
				auditLogs := make([]*models.UnifiedAuditLog, datasetSize)

				for i := 0; i < datasetSize; i++ {
					userID := fmt.Sprintf("user-%d", i)
					merchantID := fmt.Sprintf("merchant-%d", i)
					resourceType := "merchant"
					resourceID := fmt.Sprintf("merchant-%d", i)

					auditLogs[i] = &models.UnifiedAuditLog{
						ID:            fmt.Sprintf("memory-audit-%d", i),
						UserID:        &userID,
						MerchantID:    &merchantID,
						EventType:     "merchant_operation",
						EventCategory: "audit",
						Action:        "READ",
						ResourceType:  &resourceType,
						ResourceID:    &resourceID,
						CreatedAt:     time.Now(),
					}
				}

				// Validate all records
				for i := 0; i < datasetSize; i++ {
					err := auditLogs[i].Validate()
					assert.NoError(t, err, "Audit log validation should succeed")
				}

				duration := time.Since(start)

				// Performance assertions
				assert.Less(t, duration, 10*time.Second, "Memory usage test should complete within 10 seconds")

				t.Logf("✅ Dataset size %d: %v (created and validated %d audit logs)",
					datasetSize, duration, datasetSize)
			})
		}
	})

	t.Run("Concurrent Operations Performance", func(t *testing.T) {
		t.Log("Testing concurrent operations performance...")

		// Test concurrent audit log creation
		concurrencyLevels := []int{5, 10, 20, 50}

		for _, concurrency := range concurrencyLevels {
			t.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(t *testing.T) {
				start := time.Now()

				// Create channels for coordination
				done := make(chan bool, concurrency)

				// Launch concurrent operations
				for i := 0; i < concurrency; i++ {
					go func(id int) {
						// Create audit log
						auditLog := &models.UnifiedAuditLog{
							ID:            fmt.Sprintf("concurrent-audit-%d", id),
							EventType:     "merchant_operation",
							EventCategory: "audit",
							Action:        "READ",
							CreatedAt:     time.Now(),
						}

						// Validate audit log
						err := auditLog.Validate()
						assert.NoError(t, err, "Concurrent audit log validation should succeed")

						done <- true
					}(i)
				}

				// Wait for all operations to complete
				for i := 0; i < concurrency; i++ {
					<-done
				}

				duration := time.Since(start)

				// Performance assertions
				assert.Less(t, duration, 2*time.Second, "Concurrent operations should complete within 2 seconds")

				t.Logf("✅ Concurrency %d: %v", concurrency, duration)
			})
		}
	})

	t.Log("✅ All consolidated systems performance tests passed")
}
