package security

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewDataRetentionService(t *testing.T) {
	logger := zap.NewNop()

	t.Run("with default config", func(t *testing.T) {
		service, err := NewDataRetentionService(nil, logger)
		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.NotNil(t, service.config)
		assert.True(t, service.config.EnableDataRetention)
		assert.Equal(t, 365*24*time.Hour, service.config.DefaultRetentionPeriod)
		assert.NotNil(t, service.policyManager)
		assert.NotNil(t, service.lifecycleManager)
		assert.NotNil(t, service.deletionManager)
		assert.NotNil(t, service.auditManager)
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &DataRetentionConfig{
			EnableDataRetention:    true,
			DefaultRetentionPeriod: 90 * 24 * time.Hour,
			MinRetentionPeriod:     7 * 24 * time.Hour,
			MaxRetentionPeriod:     5 * 365 * 24 * time.Hour,
			AutomaticDeletion:      false,
			BackupBeforeDeletion:   true,
			RequireApproval:        true,
			NotificationEnabled:    false,
			AuditLogging:           true,
		}

		service, err := NewDataRetentionService(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, config, service.config)
		assert.False(t, service.config.AutomaticDeletion)
		assert.True(t, service.config.RequireApproval)
	})

	t.Run("with nil logger", func(t *testing.T) {
		service, err := NewDataRetentionService(nil, nil)
		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.NotNil(t, service.logger)
	})
}

func TestCreateRetentionPolicy(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewDataRetentionService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name          string
		policy        *RetentionPolicy
		expectError   bool
		errorContains string
	}{
		{
			name: "valid policy",
			policy: &RetentionPolicy{
				Name:            "Test Policy",
				Description:     "Test retention policy",
				DataCategory:    "business_data",
				DataType:        "verification_record",
				RetentionPeriod: 365 * 24 * time.Hour,
				DeletionMethod:  "soft",
				GracePeriod:     30 * 24 * time.Hour,
				RequireApproval: false,
				LegalHold:       false,
				CreatedBy:       "test_user",
			},
			expectError: false,
		},
		{
			name:          "nil policy",
			policy:        nil,
			expectError:   true,
			errorContains: "policy cannot be nil",
		},
		{
			name: "missing name",
			policy: &RetentionPolicy{
				DataCategory:    "business_data",
				RetentionPeriod: 365 * 24 * time.Hour,
			},
			expectError:   true,
			errorContains: "policy name is required",
		},
		{
			name: "missing data category",
			policy: &RetentionPolicy{
				Name:            "Test Policy",
				RetentionPeriod: 365 * 24 * time.Hour,
			},
			expectError:   true,
			errorContains: "data category is required",
		},
		{
			name: "retention period too short",
			policy: &RetentionPolicy{
				Name:            "Test Policy",
				DataCategory:    "business_data",
				RetentionPeriod: 1 * 24 * time.Hour, // Below minimum of 30 days
			},
			expectError:   true,
			errorContains: "retention period below minimum",
		},
		{
			name: "retention period too long",
			policy: &RetentionPolicy{
				Name:            "Test Policy",
				DataCategory:    "business_data",
				RetentionPeriod: 10 * 365 * 24 * time.Hour, // Above maximum of 7 years
			},
			expectError:   true,
			errorContains: "retention period exceeds maximum",
		},
		{
			name: "invalid deletion method",
			policy: &RetentionPolicy{
				Name:            "Test Policy",
				DataCategory:    "business_data",
				RetentionPeriod: 365 * 24 * time.Hour,
				DeletionMethod:  "invalid",
			},
			expectError:   true,
			errorContains: "invalid deletion method",
		},
		{
			name: "default deletion method applied",
			policy: &RetentionPolicy{
				Name:            "Test Policy",
				DataCategory:    "business_data",
				RetentionPeriod: 365 * 24 * time.Hour,
				// DeletionMethod not specified
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateRetentionPolicy(ctx, tt.policy)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.policy != nil {
					assert.NotEmpty(t, tt.policy.ID)
					assert.Equal(t, "active", tt.policy.Status)
					assert.NotZero(t, tt.policy.CreatedAt)
					assert.NotZero(t, tt.policy.UpdatedAt)
					if tt.policy.DeletionMethod == "" {
						// Should have been set to default
						assert.Equal(t, "soft", tt.policy.DeletionMethod)
					}
				}
			}
		})
	}

	t.Run("data retention disabled", func(t *testing.T) {
		disabledService := &DataRetentionService{
			config: &DataRetentionConfig{EnableDataRetention: false},
			logger: logger,
		}

		policy := &RetentionPolicy{
			Name:            "Test Policy",
			DataCategory:    "business_data",
			RetentionPeriod: 365 * 24 * time.Hour,
		}

		err := disabledService.CreateRetentionPolicy(ctx, policy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "data retention is disabled")
	})
}

func TestRegisterDataForRetention(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewDataRetentionService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name         string
		dataID       string
		dataType     string
		dataCategory string
		metadata     map[string]interface{}
		expectError  bool
	}{
		{
			name:         "valid registration",
			dataID:       "data_123",
			dataType:     "business_verification",
			dataCategory: "business_data",
			metadata: map[string]interface{}{
				"source":    "api",
				"user_id":   "user_456",
				"timestamp": time.Now(),
			},
			expectError: false,
		},
		{
			name:         "financial data with longer retention",
			dataID:       "financial_789",
			dataType:     "transaction_record",
			dataCategory: "financial_data",
			metadata: map[string]interface{}{
				"amount":      1000.50,
				"currency":    "USD",
				"customer_id": "cust_123",
			},
			expectError: false,
		},
		{
			name:         "personal identification data",
			dataID:       "personal_456",
			dataType:     "identity_verification",
			dataCategory: "personal_identification",
			metadata: map[string]interface{}{
				"verification_method": "document_scan",
				"confidence_score":    0.95,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, err := service.RegisterDataForRetention(ctx, tt.dataID, tt.dataType, tt.dataCategory, tt.metadata)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, record)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, record)
				assert.NotEmpty(t, record.ID)
				assert.Equal(t, tt.dataID, record.DataID)
				assert.Equal(t, tt.dataType, record.DataType)
				assert.Equal(t, tt.dataCategory, record.DataCategory)
				assert.NotEmpty(t, record.PolicyID)
				assert.Equal(t, "active", record.Status)
				assert.NotZero(t, record.CreatedAt)
				assert.NotZero(t, record.ExpiresAt)
				assert.True(t, record.ExpiresAt.After(record.CreatedAt))
				assert.Equal(t, tt.metadata, record.Metadata)

				// Check retention period based on category
				expectedRetention := service.config.DefaultRetentionPeriod
				if categoryPeriod, exists := service.config.CategoryRetentionPeriods[tt.dataCategory]; exists {
					expectedRetention = categoryPeriod
				}

				expectedExpiration := record.CreatedAt.Add(expectedRetention)
				assert.WithinDuration(t, expectedExpiration, record.ExpiresAt, time.Minute)

				// Check legal hold for specific categories
				isLegalHoldCategory := false
				for _, category := range service.config.LegalHoldCategories {
					if category == tt.dataCategory {
						isLegalHoldCategory = true
						break
					}
				}
				assert.Equal(t, isLegalHoldCategory, record.LegalHold)
			}
		})
	}

	t.Run("data retention disabled", func(t *testing.T) {
		disabledService := &DataRetentionService{
			config: &DataRetentionConfig{EnableDataRetention: false},
			logger: logger,
		}

		record, err := disabledService.RegisterDataForRetention(ctx, "data_123", "test", "business_data", nil)
		assert.Error(t, err)
		assert.Nil(t, record)
		assert.Contains(t, err.Error(), "data retention is disabled")
	})
}

func TestScheduleDataDeletion(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewDataRetentionService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name          string
		record        *DataLifecycleRecord
		reason        string
		expectError   bool
		errorContains string
	}{
		{
			name: "valid deletion scheduling",
			record: &DataLifecycleRecord{
				ID:           "lifecycle_123",
				DataID:       "data_123",
				DataType:     "verification_record",
				DataCategory: "business_data",
				PolicyID:     "policy_456",
				Status:       "active",
				LegalHold:    false,
				CreatedAt:    time.Now().Add(-400 * 24 * time.Hour), // Old data
				ExpiresAt:    time.Now().Add(-10 * 24 * time.Hour),  // Expired
			},
			reason:      "retention period expired",
			expectError: false,
		},
		{
			name: "legal hold prevents deletion",
			record: &DataLifecycleRecord{
				ID:           "lifecycle_456",
				DataID:       "data_456",
				DataType:     "financial_record",
				DataCategory: "financial_data",
				PolicyID:     "policy_789",
				Status:       "active",
				LegalHold:    true, // Legal hold active
				CreatedAt:    time.Now().Add(-400 * 24 * time.Hour),
				ExpiresAt:    time.Now().Add(-10 * 24 * time.Hour),
			},
			reason:        "retention period expired",
			expectError:   true,
			errorContains: "cannot delete data under legal hold",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deletionRequest, err := service.ScheduleDataDeletion(ctx, tt.record, tt.reason)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, deletionRequest)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, deletionRequest)
				assert.NotEmpty(t, deletionRequest.ID)
				assert.Equal(t, tt.record.DataID, deletionRequest.DataID)
				assert.Equal(t, tt.record.DataType, deletionRequest.DataType)
				assert.Equal(t, tt.record.DataCategory, deletionRequest.DataCategory)
				assert.Equal(t, "automatic", deletionRequest.RequestType)
				assert.Equal(t, tt.reason, deletionRequest.Reason)
				assert.Equal(t, "system", deletionRequest.RequestedBy)
				assert.Equal(t, "pending", deletionRequest.Status)
				assert.NotZero(t, deletionRequest.RequestedAt)
				assert.NotZero(t, deletionRequest.ScheduledAt)
				assert.True(t, deletionRequest.ScheduledAt.After(deletionRequest.RequestedAt))

				// Check that lifecycle record was updated
				assert.Equal(t, "pending_deletion", tt.record.Status)
				assert.NotNil(t, tt.record.DeletionScheduledAt)
				assert.Equal(t, deletionRequest.ScheduledAt, *tt.record.DeletionScheduledAt)
				assert.Equal(t, tt.reason, tt.record.DeletionReason)
			}
		})
	}

	t.Run("data retention disabled", func(t *testing.T) {
		disabledService := &DataRetentionService{
			config: &DataRetentionConfig{EnableDataRetention: false},
			logger: logger,
		}

		record := &DataLifecycleRecord{
			ID:        "test",
			Status:    "active",
			LegalHold: false,
		}

		deletionRequest, err := disabledService.ScheduleDataDeletion(ctx, record, "test")
		assert.Error(t, err)
		assert.Nil(t, deletionRequest)
		assert.Contains(t, err.Error(), "data retention is disabled")
	})
}

func TestProcessExpiredData(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewDataRetentionService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("process expired data", func(t *testing.T) {
		count, err := service.ProcessExpiredData(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, 0)
	})

	t.Run("data retention disabled", func(t *testing.T) {
		disabledService := &DataRetentionService{
			config: &DataRetentionConfig{EnableDataRetention: false},
			logger: logger,
		}

		count, err := disabledService.ProcessExpiredData(ctx)
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "data retention is disabled")
	})
}

func TestExecutePendingDeletions(t *testing.T) {
	logger := zap.NewNop()

	t.Run("execute with automatic deletion enabled", func(t *testing.T) {
		config := &DataRetentionConfig{
			EnableDataRetention:  true,
			AutomaticDeletion:    true,
			BackupBeforeDeletion: true,
			AuditLogging:         true,
		}
		service, err := NewDataRetentionService(config, logger)
		require.NoError(t, err)

		ctx := context.Background()
		count, err := service.ExecutePendingDeletions(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, 0)
	})

	t.Run("automatic deletion disabled", func(t *testing.T) {
		config := &DataRetentionConfig{
			EnableDataRetention: true,
			AutomaticDeletion:   false,
		}
		service, err := NewDataRetentionService(config, logger)
		require.NoError(t, err)

		ctx := context.Background()
		count, err := service.ExecutePendingDeletions(ctx)
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "automatic deletion is disabled")
	})

	t.Run("data retention disabled", func(t *testing.T) {
		disabledService := &DataRetentionService{
			config: &DataRetentionConfig{EnableDataRetention: false},
			logger: logger,
		}

		ctx := context.Background()
		count, err := disabledService.ExecutePendingDeletions(ctx)
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "automatic deletion is disabled")
	})
}

func TestApproveDeletion(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewDataRetentionService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("approve deletion request", func(t *testing.T) {
		requestID := "deletion_123"
		approverID := "approver_456"

		err := service.ApproveDeletion(ctx, requestID, approverID)
		assert.NoError(t, err)
	})

	t.Run("data retention disabled", func(t *testing.T) {
		disabledService := &DataRetentionService{
			config: &DataRetentionConfig{EnableDataRetention: false},
			logger: logger,
		}

		err := disabledService.ApproveDeletion(ctx, "test", "approver")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "data retention is disabled")
	})
}

func TestGenerateRetentionReport(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewDataRetentionService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name       string
		reportType string
		period     string
	}{
		{
			name:       "compliance report",
			reportType: "compliance",
			period:     "monthly",
		},
		{
			name:       "lifecycle report",
			reportType: "lifecycle",
			period:     "weekly",
		},
		{
			name:       "policy effectiveness report",
			reportType: "policy_effectiveness",
			period:     "quarterly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := service.GenerateRetentionReport(ctx, tt.reportType, tt.period)
			assert.NoError(t, err)
			assert.NotNil(t, report)
			assert.NotEmpty(t, report.ID)
			assert.Equal(t, tt.reportType, report.ReportType)
			assert.Equal(t, tt.period, report.Period)
			assert.NotZero(t, report.GeneratedAt)
			assert.GreaterOrEqual(t, report.TotalRecords, 0)
			assert.GreaterOrEqual(t, report.ComplianceScore, 0.0)
			assert.LessOrEqual(t, report.ComplianceScore, 100.0)
			assert.NotEmpty(t, report.Recommendations)
		})
	}

	t.Run("data retention disabled", func(t *testing.T) {
		disabledService := &DataRetentionService{
			config: &DataRetentionConfig{EnableDataRetention: false},
			logger: logger,
		}

		report, err := disabledService.GenerateRetentionReport(ctx, "compliance", "monthly")
		assert.Error(t, err)
		assert.Nil(t, report)
		assert.Contains(t, err.Error(), "data retention is disabled")
	})
}

func TestValidateRetentionPolicy(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewDataRetentionService(nil, logger)
	require.NoError(t, err)

	tests := []struct {
		name          string
		policy        *RetentionPolicy
		expectError   bool
		errorContains string
	}{
		{
			name: "valid policy",
			policy: &RetentionPolicy{
				Name:            "Valid Policy",
				DataCategory:    "business_data",
				RetentionPeriod: 180 * 24 * time.Hour,
				DeletionMethod:  "soft",
			},
			expectError: false,
		},
		{
			name: "empty name",
			policy: &RetentionPolicy{
				DataCategory:    "business_data",
				RetentionPeriod: 180 * 24 * time.Hour,
			},
			expectError:   true,
			errorContains: "policy name is required",
		},
		{
			name: "empty data category",
			policy: &RetentionPolicy{
				Name:            "Test Policy",
				RetentionPeriod: 180 * 24 * time.Hour,
			},
			expectError:   true,
			errorContains: "data category is required",
		},
		{
			name: "retention period too short",
			policy: &RetentionPolicy{
				Name:            "Test Policy",
				DataCategory:    "business_data",
				RetentionPeriod: 1 * 24 * time.Hour, // Below minimum
			},
			expectError:   true,
			errorContains: "retention period below minimum",
		},
		{
			name: "retention period too long",
			policy: &RetentionPolicy{
				Name:            "Test Policy",
				DataCategory:    "business_data",
				RetentionPeriod: 10 * 365 * 24 * time.Hour, // Above maximum
			},
			expectError:   true,
			errorContains: "retention period exceeds maximum",
		},
		{
			name: "invalid deletion method",
			policy: &RetentionPolicy{
				Name:            "Test Policy",
				DataCategory:    "business_data",
				RetentionPeriod: 180 * 24 * time.Hour,
				DeletionMethod:  "invalid_method",
			},
			expectError:   true,
			errorContains: "invalid deletion method",
		},
		{
			name: "default deletion method set",
			policy: &RetentionPolicy{
				Name:            "Test Policy",
				DataCategory:    "business_data",
				RetentionPeriod: 180 * 24 * time.Hour,
				// No DeletionMethod specified
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateRetentionPolicy(tt.policy)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.policy != nil && tt.policy.DeletionMethod == "" {
					// Should have been set to default
					assert.Equal(t, "soft", tt.policy.DeletionMethod)
				}
			}
		})
	}
}

func TestRetentionCalculateComplianceScore(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewDataRetentionService(nil, logger)
	require.NoError(t, err)

	tests := []struct {
		name          string
		stats         *RetentionStatistics
		expectedScore float64
	}{
		{
			name: "perfect compliance",
			stats: &RetentionStatistics{
				TotalRecords:     1000,
				ActiveRecords:    1000,
				ExpiredRecords:   0,
				DeletedRecords:   0,
				PendingDeletions: 0,
				PolicyViolations: 0,
				LegalHoldRecords: 0,
			},
			expectedScore: 100.0,
		},
		{
			name: "some violations",
			stats: &RetentionStatistics{
				TotalRecords:     1000,
				ActiveRecords:    950,
				ExpiredRecords:   30,
				DeletedRecords:   20,
				PendingDeletions: 5,
				PolicyViolations: 2,
				LegalHoldRecords: 10,
			},
			expectedScore: 70.0, // 100 - (2*10) - (5*2) = 70
		},
		{
			name: "many violations and pending deletions",
			stats: &RetentionStatistics{
				TotalRecords:     1000,
				ActiveRecords:    800,
				ExpiredRecords:   100,
				DeletedRecords:   100,
				PendingDeletions: 30,
				PolicyViolations: 8,
				LegalHoldRecords: 20,
			},
			expectedScore: 0.0, // 100 - (8*10) - (30*2) = 0 (capped at 0)
		},
		{
			name: "zero total records",
			stats: &RetentionStatistics{
				TotalRecords:     0,
				ActiveRecords:    0,
				ExpiredRecords:   0,
				DeletedRecords:   0,
				PendingDeletions: 0,
				PolicyViolations: 0,
				LegalHoldRecords: 0,
			},
			expectedScore: 100.0, // Default when no records
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.calculateComplianceScore(tt.stats)
			assert.Equal(t, tt.expectedScore, score)
		})
	}
}

func TestGenerateRetentionRecommendations(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewDataRetentionService(nil, logger)
	require.NoError(t, err)

	tests := []struct {
		name                    string
		stats                   *RetentionStatistics
		expectedRecommendations []string
		shouldContain           []string
	}{
		{
			name: "policy violations present",
			stats: &RetentionStatistics{
				TotalRecords:     1000,
				PolicyViolations: 5,
				PendingDeletions: 10,
				ExpiredRecords:   20,
			},
			shouldContain: []string{"policy violations"},
		},
		{
			name: "too many pending deletions",
			stats: &RetentionStatistics{
				TotalRecords:     1000,
				PolicyViolations: 0,
				PendingDeletions: 100, // More than 5% of total
				ExpiredRecords:   20,
			},
			shouldContain: []string{"automation", "approval"},
		},
		{
			name: "too many expired records",
			stats: &RetentionStatistics{
				TotalRecords:     1000,
				ActiveRecords:    800,
				PolicyViolations: 0,
				PendingDeletions: 10,
				ExpiredRecords:   150, // More than 10% of active
			},
			shouldContain: []string{"cleanup"},
		},
		{
			name: "good compliance",
			stats: &RetentionStatistics{
				TotalRecords:     1000,
				ActiveRecords:    950,
				PolicyViolations: 0,
				PendingDeletions: 20,
				ExpiredRecords:   30,
			},
			shouldContain: []string{"acceptable parameters"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendations := service.generateRetentionRecommendations(tt.stats)
			assert.NotEmpty(t, recommendations)

			for _, expected := range tt.shouldContain {
				found := false
				for _, recommendation := range recommendations {
					if containsSubstring(recommendation, expected) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected to find recommendation containing '%s'", expected)
			}
		})
	}
}

func TestRetentionIDGeneration(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewDataRetentionService(nil, logger)
	require.NoError(t, err)

	t.Run("generate unique IDs", func(t *testing.T) {
		// Generate multiple IDs and ensure they're unique
		policyID1 := service.generatePolicyID()
		time.Sleep(1 * time.Nanosecond) // Ensure different timestamps
		policyID2 := service.generatePolicyID()
		assert.NotEqual(t, policyID1, policyID2)
		assert.Contains(t, policyID1, "policy_")

		lifecycleID1 := service.generateLifecycleID()
		time.Sleep(1 * time.Nanosecond)
		lifecycleID2 := service.generateLifecycleID()
		assert.NotEqual(t, lifecycleID1, lifecycleID2)
		assert.Contains(t, lifecycleID1, "lifecycle_")

		deletionID1 := service.generateDeletionRequestID()
		time.Sleep(1 * time.Nanosecond)
		deletionID2 := service.generateDeletionRequestID()
		assert.NotEqual(t, deletionID1, deletionID2)
		assert.Contains(t, deletionID1, "deletion_")

		auditID1 := service.generateAuditID()
		time.Sleep(1 * time.Nanosecond)
		auditID2 := service.generateAuditID()
		assert.NotEqual(t, auditID1, auditID2)
		assert.Contains(t, auditID1, "audit_")

		reportID1 := service.generateReportID()
		time.Sleep(1 * time.Nanosecond)
		reportID2 := service.generateReportID()
		assert.NotEqual(t, reportID1, reportID2)
		assert.Contains(t, reportID1, "report_")
	})
}

// Helper function to check if a string contains a substring (case-insensitive)
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstringAt(s, substr))))
}

func containsSubstringAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
