package external

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuditTrailManager(t *testing.T) {
	tests := []struct {
		name   string
		config *AuditTrailConfig
	}{
		{
			name:   "nil config uses defaults",
			config: nil,
		},
		{
			name: "custom config",
			config: &AuditTrailConfig{
				MaxHistoryDays:     30,
				EnableDiskStorage:  false,
				EnableDetailedLogs: false,
				RetentionPolicy:    "count_based",
				MaxEntries:         5000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewAuditTrailManager(tt.config)
			assert.NotNil(t, manager)
			assert.NotNil(t, manager.config)

			if tt.config == nil {
				// Should use default config
				assert.Equal(t, 90, manager.config.MaxHistoryDays)
				assert.True(t, manager.config.EnableDiskStorage)
				assert.True(t, manager.config.EnableDetailedLogs)
				assert.Equal(t, "time_based", manager.config.RetentionPolicy)
				assert.Equal(t, 10000, manager.config.MaxEntries)
			} else {
				assert.Equal(t, tt.config, manager.config)
			}
		})
	}
}

func TestDefaultAuditTrailConfig(t *testing.T) {
	config := DefaultAuditTrailConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 90, config.MaxHistoryDays)
	assert.True(t, config.EnableDiskStorage)
	assert.True(t, config.EnableDetailedLogs)
	assert.Equal(t, "time_based", config.RetentionPolicy)
	assert.Equal(t, 10000, config.MaxEntries)
}

func TestAuditTrailManager_CreateVerificationHistory(t *testing.T) {
	manager := NewAuditTrailManager(nil)
	ctx := context.Background()

	baseTime := time.Now()

	tests := []struct {
		name           string
		verificationID string
		businessName   string
		websiteURL     string
		events         []AuditEvent
		metadata       map[string]interface{}
		expectError    bool
		checkHistory   func(*testing.T, *VerificationHistory)
	}{
		{
			name:           "empty events returns error",
			verificationID: "test-123",
			businessName:   "Test Business",
			websiteURL:     "https://test.com",
			events:         []AuditEvent{},
			expectError:    true,
		},
		{
			name:           "complete verification history",
			verificationID: "test-456",
			businessName:   "Complete Business",
			websiteURL:     "https://complete.com",
			events: []AuditEvent{
				{
					EventID:     "event_1",
					Timestamp:   baseTime,
					EventType:   "verification_started",
					Description: "Verification started",
					Severity:    "info",
					UserID:      "system",
					Data: map[string]interface{}{
						"verification_id": "test-456",
					},
				},
				{
					EventID:     "event_2",
					Timestamp:   baseTime.Add(1 * time.Minute),
					EventType:   "data_extracted",
					Description: "Data extracted",
					Severity:    "info",
					UserID:      "system",
					Data: map[string]interface{}{
						"fields_extracted": 5,
					},
				},
				{
					EventID:     "event_3",
					Timestamp:   baseTime.Add(2 * time.Minute),
					EventType:   "confidence_calculated",
					Description: "Confidence calculated",
					Severity:    "info",
					UserID:      "system",
					Data: map[string]interface{}{
						"overall_score": 0.85,
					},
				},
				{
					EventID:     "event_4",
					Timestamp:   baseTime.Add(3 * time.Minute),
					EventType:   "status_assigned",
					Description: "Status assigned",
					Severity:    "info",
					UserID:      "system",
					Data: map[string]interface{}{
						"status":        "PASSED",
						"overall_score": 0.85,
					},
				},
				{
					EventID:     "event_5",
					Timestamp:   baseTime.Add(4 * time.Minute),
					EventType:   "report_generated",
					Description: "Report generated",
					Severity:    "info",
					UserID:      "system",
					Data: map[string]interface{}{
						"report_type": "comprehensive",
					},
				},
			},
			metadata: map[string]interface{}{
				"source": "api",
			},
			expectError: false,
			checkHistory: func(t *testing.T, history *VerificationHistory) {
				assert.Equal(t, "test-456", history.VerificationID)
				assert.Equal(t, "Complete Business", history.BusinessName)
				assert.Equal(t, "https://complete.com", history.WebsiteURL)
				assert.Equal(t, baseTime, history.InitiatedAt)
				assert.NotNil(t, history.CompletedAt)
				assert.Equal(t, baseTime.Add(4*time.Minute), *history.CompletedAt)
				assert.Equal(t, 4*time.Minute, history.Duration)
				assert.Equal(t, "PASSED", history.FinalStatus)
				assert.Equal(t, 0.85, history.FinalScore)
				assert.Len(t, history.Events, 5)
				assert.True(t, len(history.Milestones) >= 5)
				assert.Equal(t, "api", history.Metadata["source"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history, err := manager.CreateVerificationHistory(
				ctx,
				tt.verificationID,
				tt.businessName,
				tt.websiteURL,
				tt.events,
				tt.metadata,
			)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, history)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, history)
				if tt.checkHistory != nil {
					tt.checkHistory(t, history)
				}
			}
		})
	}
}

func TestAuditTrailManager_GenerateMilestones(t *testing.T) {
	manager := NewAuditTrailManager(nil)
	baseTime := time.Now()

	events := []AuditEvent{
		{
			EventID:     "event_1",
			Timestamp:   baseTime,
			EventType:   "verification_started",
			Description: "Verification started",
			Severity:    "info",
		},
		{
			EventID:     "event_2",
			Timestamp:   baseTime.Add(1 * time.Minute),
			EventType:   "data_extracted",
			Description: "Data extracted",
			Severity:    "info",
		},
		{
			EventID:     "event_3",
			Timestamp:   baseTime.Add(2 * time.Minute),
			EventType:   "confidence_calculated",
			Description: "Confidence calculated",
			Severity:    "info",
		},
	}

	milestones := manager.generateMilestones(events)

	assert.Len(t, milestones, 3)

	// Check milestone order
	assert.Equal(t, "Verification Initiated", milestones[0].Name)
	assert.Equal(t, "Data Extraction", milestones[1].Name)
	assert.Equal(t, "Confidence Scoring", milestones[2].Name)

	// Check critical path marking
	for _, milestone := range milestones {
		assert.True(t, milestone.CriticalPath)
		assert.Equal(t, "completed", milestone.Status)
	}

	// Check durations
	assert.Equal(t, time.Duration(0), milestones[0].Duration) // Start has no duration
	assert.Equal(t, 1*time.Minute, milestones[1].Duration)    // Extract took 1 minute
	assert.Equal(t, 1*time.Minute, milestones[2].Duration)    // Confidence took 1 minute
}

func TestAuditTrailManager_QueryAuditTrail(t *testing.T) {
	manager := NewAuditTrailManager(nil)
	ctx := context.Background()
	baseTime := time.Now()

	events := []AuditEvent{
		{
			EventID:   "event_1",
			Timestamp: baseTime,
			EventType: "verification_started",
			Severity:  "info",
			UserID:    "system",
			Data: map[string]interface{}{
				"verification_id": "test-123",
			},
		},
		{
			EventID:   "event_2",
			Timestamp: baseTime.Add(1 * time.Minute),
			EventType: "data_extracted",
			Severity:  "info",
			UserID:    "system",
			Data: map[string]interface{}{
				"verification_id": "test-123",
			},
		},
		{
			EventID:   "event_3",
			Timestamp: baseTime.Add(2 * time.Minute),
			EventType: "verification_started",
			Severity:  "warning",
			UserID:    "user1",
			Data: map[string]interface{}{
				"verification_id": "test-456",
			},
		},
	}

	tests := []struct {
		name           string
		query          AuditQuery
		expectedCount  int
		expectedEvents []string // Event IDs
	}{
		{
			name:           "no filters returns all",
			query:          AuditQuery{},
			expectedCount:  3,
			expectedEvents: []string{"event_3", "event_2", "event_1"}, // Most recent first
		},
		{
			name: "filter by verification ID",
			query: AuditQuery{
				VerificationID: "test-123",
			},
			expectedCount:  2,
			expectedEvents: []string{"event_2", "event_1"},
		},
		{
			name: "filter by event type",
			query: AuditQuery{
				EventType: "verification_started",
			},
			expectedCount:  2,
			expectedEvents: []string{"event_3", "event_1"},
		},
		{
			name: "filter by severity",
			query: AuditQuery{
				Severity: "warning",
			},
			expectedCount:  1,
			expectedEvents: []string{"event_3"},
		},
		{
			name: "filter by user ID",
			query: AuditQuery{
				UserID: "user1",
			},
			expectedCount:  1,
			expectedEvents: []string{"event_3"},
		},
		{
			name: "filter by time range",
			query: AuditQuery{
				StartTime: &baseTime,
				EndTime:   func() *time.Time { t := baseTime.Add(1*time.Minute + 30*time.Second); return &t }(),
			},
			expectedCount:  2,
			expectedEvents: []string{"event_2", "event_1"},
		},
		{
			name: "pagination with limit",
			query: AuditQuery{
				Limit: 2,
			},
			expectedCount:  2,
			expectedEvents: []string{"event_3", "event_2"},
		},
		{
			name: "pagination with offset",
			query: AuditQuery{
				Offset: 1,
				Limit:  1,
			},
			expectedCount:  1,
			expectedEvents: []string{"event_2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := manager.QueryAuditTrail(ctx, tt.query, events)
			require.NoError(t, err)

			assert.Len(t, result, tt.expectedCount)

			for i, expectedEventID := range tt.expectedEvents {
				if i < len(result) {
					assert.Equal(t, expectedEventID, result[i].EventID)
				}
			}
		})
	}
}

func TestAuditTrailManager_GenerateHistorySummary(t *testing.T) {
	manager := NewAuditTrailManager(nil)

	history := &VerificationHistory{
		VerificationID: "test-123",
		BusinessName:   "Test Business",
		Duration:       5 * time.Minute,
		FinalStatus:    "PASSED",
		FinalScore:     0.85,
		Events: []AuditEvent{
			{EventType: "verification_started", Severity: "info"},
			{EventType: "data_extracted", Severity: "info"},
			{EventType: "confidence_calculated", Severity: "warning"},
		},
		Milestones: []HistoryMilestone{
			{Status: "completed", CriticalPath: true, Duration: 1 * time.Minute},
			{Status: "completed", CriticalPath: true, Duration: 2 * time.Minute},
			{Status: "completed", CriticalPath: false, Duration: 30 * time.Second},
		},
	}

	summary := manager.GenerateHistorySummary(history)

	assert.NotNil(t, summary)
	assert.Equal(t, "test-123", summary.VerificationID)
	assert.Equal(t, "Test Business", summary.BusinessName)
	assert.Equal(t, 5*time.Minute, summary.TotalDuration)
	assert.Equal(t, 3, summary.EventCount)
	assert.Equal(t, 3, summary.MilestoneCount)
	assert.Equal(t, "PASSED", summary.FinalStatus)
	assert.Equal(t, 0.85, summary.FinalScore)
	assert.Equal(t, 1.0, summary.CompletionRate)             // All milestones completed
	assert.Equal(t, 3*time.Minute, summary.CriticalPathTime) // 1 + 2 minutes

	// Check event counts by type
	assert.Equal(t, 1, summary.EventsByType["verification_started"])
	assert.Equal(t, 1, summary.EventsByType["data_extracted"])
	assert.Equal(t, 1, summary.EventsByType["confidence_calculated"])

	// Check event counts by severity
	assert.Equal(t, 2, summary.EventsBySeverity["info"])
	assert.Equal(t, 1, summary.EventsBySeverity["warning"])
}

func TestAuditTrailManager_UpdateConfig(t *testing.T) {
	manager := NewAuditTrailManager(nil)

	tests := []struct {
		name        string
		config      *AuditTrailConfig
		expectError bool
	}{
		{
			name:        "nil config returns error",
			config:      nil,
			expectError: true,
		},
		{
			name: "invalid max history days",
			config: &AuditTrailConfig{
				MaxHistoryDays: 0,
				MaxEntries:     100,
			},
			expectError: true,
		},
		{
			name: "invalid max entries",
			config: &AuditTrailConfig{
				MaxHistoryDays: 30,
				MaxEntries:     0,
			},
			expectError: true,
		},
		{
			name: "valid config",
			config: &AuditTrailConfig{
				MaxHistoryDays:     60,
				EnableDiskStorage:  false,
				EnableDetailedLogs: true,
				RetentionPolicy:    "count_based",
				MaxEntries:         5000,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.UpdateConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.config, manager.GetConfig())
			}
		})
	}
}

func TestVerificationHistory_StructFields(t *testing.T) {
	history := &VerificationHistory{
		VerificationID: "test-123",
		BusinessName:   "Test Business",
		WebsiteURL:     "https://test.com",
		InitiatedAt:    time.Now(),
		FinalStatus:    "PASSED",
		FinalScore:     0.85,
	}

	assert.Equal(t, "test-123", history.VerificationID)
	assert.Equal(t, "Test Business", history.BusinessName)
	assert.Equal(t, "https://test.com", history.WebsiteURL)
	assert.Equal(t, "PASSED", history.FinalStatus)
	assert.Equal(t, 0.85, history.FinalScore)
}

func TestHistoryMilestone_StructFields(t *testing.T) {
	milestone := &HistoryMilestone{
		MilestoneID:  "milestone-1",
		Name:         "Test Milestone",
		Description:  "Test Description",
		Status:       "completed",
		CriticalPath: true,
		Duration:     1 * time.Minute,
	}

	assert.Equal(t, "milestone-1", milestone.MilestoneID)
	assert.Equal(t, "Test Milestone", milestone.Name)
	assert.Equal(t, "Test Description", milestone.Description)
	assert.Equal(t, "completed", milestone.Status)
	assert.True(t, milestone.CriticalPath)
	assert.Equal(t, 1*time.Minute, milestone.Duration)
}

func TestAuditQuery_StructFields(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	query := &AuditQuery{
		VerificationID: "test-123",
		BusinessName:   "Test Business",
		EventType:      "verification_started",
		Severity:       "info",
		UserID:         "user1",
		StartTime:      &startTime,
		EndTime:        &endTime,
		Limit:          10,
		Offset:         5,
	}

	assert.Equal(t, "test-123", query.VerificationID)
	assert.Equal(t, "Test Business", query.BusinessName)
	assert.Equal(t, "verification_started", query.EventType)
	assert.Equal(t, "info", query.Severity)
	assert.Equal(t, "user1", query.UserID)
	assert.Equal(t, startTime, *query.StartTime)
	assert.Equal(t, endTime, *query.EndTime)
	assert.Equal(t, 10, query.Limit)
	assert.Equal(t, 5, query.Offset)
}
