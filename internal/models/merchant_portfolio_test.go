package models

import (
	"testing"
	"time"
)

// Test helper functions
func createTestMerchant() *Merchant {
	now := time.Now()
	return &Merchant{
		ID:                 "merchant_123",
		Name:               "Test Company",
		LegalName:          "Test Company LLC",
		RegistrationNumber: "REG123456",
		TaxID:              "TAX123456",
		Industry:           "Technology",
		IndustryCode:       "541511",
		BusinessType:       "LLC",
		FoundedDate:        &now,
		EmployeeCount:      50,
		Address: Address{
			Street1:    "123 Test St",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "USA",
		},
		ContactInfo: ContactInfo{
			Phone:   "+1-555-123-4567",
			Email:   "test@testcompany.com",
			Website: "https://testcompany.com",
		},
		PortfolioType:    PortfolioTypeProspective,
		RiskLevel:        RiskLevelMedium,
		ComplianceStatus: "pending",
		Status:           "active",
		CreatedBy:        "user123",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func createTestMerchantSession() *MerchantSession {
	now := time.Now()
	return &MerchantSession{
		ID:         "session_123",
		UserID:     "user123",
		MerchantID: "merchant_123",
		StartedAt:  now,
		LastActive: now,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func createTestAuditLog() *AuditLog {
	return &AuditLog{
		ID:           "audit_123",
		UserID:       "user123",
		MerchantID:   "merchant_123",
		Action:       "CREATE_MERCHANT",
		ResourceType: "merchant",
		ResourceID:   "merchant_123",
		Details:      "Merchant created",
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0",
		RequestID:    "req_123",
		CreatedAt:    time.Now(),
	}
}

func createTestMerchantNotification() *MerchantNotification {
	return &MerchantNotification{
		ID:         "notification_123",
		MerchantID: "merchant_123",
		UserID:     "user123",
		Type:       string(NotificationTypeRiskAlert),
		Title:      "Risk Alert",
		Message:    "High risk detected",
		IsRead:     false,
		Priority:   string(NotificationPriorityHigh),
		CreatedAt:  time.Now(),
	}
}

func createTestMerchantComparison() *MerchantComparison {
	return &MerchantComparison{
		ID:          "comparison_123",
		Merchant1ID: "merchant_123",
		Merchant2ID: "merchant_456",
		UserID:      "user123",
		ComparisonData: map[string]interface{}{
			"risk_comparison": "merchant_123 has higher risk",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTestMerchantAnalytics() *MerchantAnalytics {
	return &MerchantAnalytics{
		MerchantID:        "merchant_123",
		RiskScore:         0.75,
		ComplianceScore:   0.85,
		TransactionVolume: 100000.0,
		LastActivity:      &time.Time{},
		Flags:             []string{"high_risk", "compliance_review"},
		Metadata: map[string]interface{}{
			"analysis_date": time.Now(),
		},
		CalculatedAt: time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// PortfolioType tests
func TestPortfolioType(t *testing.T) {
	tests := []struct {
		name     string
		pt       PortfolioType
		expected bool
	}{
		{"Valid onboarded", PortfolioTypeOnboarded, true},
		{"Valid deactivated", PortfolioTypeDeactivated, true},
		{"Valid prospective", PortfolioTypeProspective, true},
		{"Valid pending", PortfolioTypePending, true},
		{"Invalid type", "invalid", false},
		{"Empty type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pt.IsValid()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}

	// Test String method
	if PortfolioTypeOnboarded.String() != "onboarded" {
		t.Error("Expected String() to return 'onboarded'")
	}
}

// RiskLevel tests
func TestRiskLevel(t *testing.T) {
	tests := []struct {
		name     string
		rl       RiskLevel
		expected bool
	}{
		{"Valid high", RiskLevelHigh, true},
		{"Valid medium", RiskLevelMedium, true},
		{"Valid low", RiskLevelLow, true},
		{"Invalid level", "invalid", false},
		{"Empty level", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rl.IsValid()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}

	// Test String method
	if RiskLevelHigh.String() != "high" {
		t.Error("Expected String() to return 'high'")
	}

	// Test GetNumericValue method
	if RiskLevelHigh.GetNumericValue() != 3 {
		t.Error("Expected high risk to have numeric value 3")
	}
	if RiskLevelMedium.GetNumericValue() != 2 {
		t.Error("Expected medium risk to have numeric value 2")
	}
	if RiskLevelLow.GetNumericValue() != 1 {
		t.Error("Expected low risk to have numeric value 1")
	}
}

// MerchantSession tests
func TestMerchantSession(t *testing.T) {
	session := createTestMerchantSession()

	// Test IsExpired method
	if session.IsExpired() {
		t.Error("Expected session not to be expired")
	}

	// Test UpdateLastActive method
	oldLastActive := session.LastActive
	session.UpdateLastActive()
	if session.LastActive.Equal(oldLastActive) {
		t.Error("Expected LastActive to be updated")
	}
	if session.UpdatedAt.Equal(oldLastActive) {
		t.Error("Expected UpdatedAt to be updated")
	}

	// Test expired session
	expiredSession := createTestMerchantSession()
	expiredSession.LastActive = time.Now().Add(-25 * time.Hour)
	if !expiredSession.IsExpired() {
		t.Error("Expected session to be expired")
	}
}

// MerchantSearchFilters tests
func TestMerchantSearchFilters(t *testing.T) {
	// Test empty filters
	emptyFilters := &MerchantSearchFilters{}
	if !emptyFilters.IsEmpty() {
		t.Error("Expected empty filters to be empty")
	}

	// Test non-empty filters
	portfolioType := PortfolioTypeOnboarded
	nonEmptyFilters := &MerchantSearchFilters{
		PortfolioType: &portfolioType,
		SearchQuery:   "test",
	}
	if nonEmptyFilters.IsEmpty() {
		t.Error("Expected non-empty filters to not be empty")
	}
}

// BulkOperationStatus tests
func TestBulkOperationStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   BulkOperationStatus
		expected bool
	}{
		{"Valid pending", BulkOperationStatusPending, true},
		{"Valid processing", BulkOperationStatusProcessing, true},
		{"Valid completed", BulkOperationStatusCompleted, true},
		{"Valid failed", BulkOperationStatusFailed, true},
		{"Valid cancelled", BulkOperationStatusCancelled, true},
		{"Invalid status", "invalid", false},
		{"Empty status", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}

	// Test String method
	if BulkOperationStatusPending.String() != "pending" {
		t.Error("Expected String() to return 'pending'")
	}
}

// NotificationType tests
func TestNotificationType(t *testing.T) {
	tests := []struct {
		name     string
		nt       NotificationType
		expected bool
	}{
		{"Valid risk alert", NotificationTypeRiskAlert, true},
		{"Valid compliance", NotificationTypeCompliance, true},
		{"Valid status change", NotificationTypeStatusChange, true},
		{"Valid bulk operation", NotificationTypeBulkOperation, true},
		{"Valid system", NotificationTypeSystem, true},
		{"Invalid type", "invalid", false},
		{"Empty type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.nt.IsValid()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}

	// Test String method
	if NotificationTypeRiskAlert.String() != "risk_alert" {
		t.Error("Expected String() to return 'risk_alert'")
	}
}

// NotificationPriority tests
func TestNotificationPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority NotificationPriority
		expected bool
	}{
		{"Valid low", NotificationPriorityLow, true},
		{"Valid medium", NotificationPriorityMedium, true},
		{"Valid high", NotificationPriorityHigh, true},
		{"Valid critical", NotificationPriorityCritical, true},
		{"Invalid priority", "invalid", false},
		{"Empty priority", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.priority.IsValid()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}

	// Test String method
	if NotificationPriorityHigh.String() != "high" {
		t.Error("Expected String() to return 'high'")
	}

	// Test GetNumericValue method
	if NotificationPriorityCritical.GetNumericValue() != 4 {
		t.Error("Expected critical priority to have numeric value 4")
	}
	if NotificationPriorityHigh.GetNumericValue() != 3 {
		t.Error("Expected high priority to have numeric value 3")
	}
	if NotificationPriorityMedium.GetNumericValue() != 2 {
		t.Error("Expected medium priority to have numeric value 2")
	}
	if NotificationPriorityLow.GetNumericValue() != 1 {
		t.Error("Expected low priority to have numeric value 1")
	}
}

// ValidationError tests
func TestValidationError(t *testing.T) {
	ve := &ValidationError{
		Field:   "name",
		Message: "name is required",
	}

	if ve.Error() != "name is required" {
		t.Error("Expected Error() to return the message")
	}
}

// Merchant validation tests
func TestMerchantValidate(t *testing.T) {
	tests := []struct {
		name      string
		merchant  *Merchant
		expectErr bool
	}{
		{
			name:      "Valid merchant",
			merchant:  createTestMerchant(),
			expectErr: false,
		},
		{
			name: "Empty name",
			merchant: func() *Merchant {
				m := createTestMerchant()
				m.Name = ""
				return m
			}(),
			expectErr: true,
		},
		{
			name: "Invalid portfolio type",
			merchant: func() *Merchant {
				m := createTestMerchant()
				m.PortfolioType = "invalid"
				return m
			}(),
			expectErr: true,
		},
		{
			name: "Invalid risk level",
			merchant: func() *Merchant {
				m := createTestMerchant()
				m.RiskLevel = "invalid"
				return m
			}(),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.merchant.Validate()
			if tt.expectErr && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}

// MerchantSession validation tests
func TestMerchantSessionValidate(t *testing.T) {
	tests := []struct {
		name      string
		session   *MerchantSession
		expectErr bool
	}{
		{
			name:      "Valid session",
			session:   createTestMerchantSession(),
			expectErr: false,
		},
		{
			name: "Empty user ID",
			session: func() *MerchantSession {
				s := createTestMerchantSession()
				s.UserID = ""
				return s
			}(),
			expectErr: true,
		},
		{
			name: "Empty merchant ID",
			session: func() *MerchantSession {
				s := createTestMerchantSession()
				s.MerchantID = ""
				return s
			}(),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.session.Validate()
			if tt.expectErr && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}

// AuditLog validation tests
func TestAuditLogValidate(t *testing.T) {
	tests := []struct {
		name      string
		auditLog  *AuditLog
		expectErr bool
	}{
		{
			name:      "Valid audit log",
			auditLog:  createTestAuditLog(),
			expectErr: false,
		},
		{
			name: "Empty user ID",
			auditLog: func() *AuditLog {
				al := createTestAuditLog()
				al.UserID = ""
				return al
			}(),
			expectErr: true,
		},
		{
			name: "Empty action",
			auditLog: func() *AuditLog {
				al := createTestAuditLog()
				al.Action = ""
				return al
			}(),
			expectErr: true,
		},
		{
			name: "Empty resource type",
			auditLog: func() *AuditLog {
				al := createTestAuditLog()
				al.ResourceType = ""
				return al
			}(),
			expectErr: true,
		},
		{
			name: "Empty resource ID",
			auditLog: func() *AuditLog {
				al := createTestAuditLog()
				al.ResourceID = ""
				return al
			}(),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.auditLog.Validate()
			if tt.expectErr && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}

// MerchantNotification validation tests
func TestMerchantNotificationValidate(t *testing.T) {
	tests := []struct {
		name         string
		notification *MerchantNotification
		expectErr    bool
	}{
		{
			name:         "Valid notification",
			notification: createTestMerchantNotification(),
			expectErr:    false,
		},
		{
			name: "Empty merchant ID",
			notification: func() *MerchantNotification {
				n := createTestMerchantNotification()
				n.MerchantID = ""
				return n
			}(),
			expectErr: true,
		},
		{
			name: "Empty user ID",
			notification: func() *MerchantNotification {
				n := createTestMerchantNotification()
				n.UserID = ""
				return n
			}(),
			expectErr: true,
		},
		{
			name: "Empty type",
			notification: func() *MerchantNotification {
				n := createTestMerchantNotification()
				n.Type = ""
				return n
			}(),
			expectErr: true,
		},
		{
			name: "Empty title",
			notification: func() *MerchantNotification {
				n := createTestMerchantNotification()
				n.Title = ""
				return n
			}(),
			expectErr: true,
		},
		{
			name: "Empty message",
			notification: func() *MerchantNotification {
				n := createTestMerchantNotification()
				n.Message = ""
				return n
			}(),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.notification.Validate()
			if tt.expectErr && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}

// MerchantComparison validation tests
func TestMerchantComparisonValidate(t *testing.T) {
	tests := []struct {
		name       string
		comparison *MerchantComparison
		expectErr  bool
	}{
		{
			name:       "Valid comparison",
			comparison: createTestMerchantComparison(),
			expectErr:  false,
		},
		{
			name: "Empty merchant1 ID",
			comparison: func() *MerchantComparison {
				c := createTestMerchantComparison()
				c.Merchant1ID = ""
				return c
			}(),
			expectErr: true,
		},
		{
			name: "Empty merchant2 ID",
			comparison: func() *MerchantComparison {
				c := createTestMerchantComparison()
				c.Merchant2ID = ""
				return c
			}(),
			expectErr: true,
		},
		{
			name: "Empty user ID",
			comparison: func() *MerchantComparison {
				c := createTestMerchantComparison()
				c.UserID = ""
				return c
			}(),
			expectErr: true,
		},
		{
			name: "Same merchant IDs",
			comparison: func() *MerchantComparison {
				c := createTestMerchantComparison()
				c.Merchant2ID = c.Merchant1ID
				return c
			}(),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.comparison.Validate()
			if tt.expectErr && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}

// MerchantAnalytics validation tests
func TestMerchantAnalyticsValidate(t *testing.T) {
	tests := []struct {
		name      string
		analytics *MerchantAnalytics
		expectErr bool
	}{
		{
			name:      "Valid analytics",
			analytics: createTestMerchantAnalytics(),
			expectErr: false,
		},
		{
			name: "Empty merchant ID",
			analytics: func() *MerchantAnalytics {
				a := createTestMerchantAnalytics()
				a.MerchantID = ""
				return a
			}(),
			expectErr: true,
		},
		{
			name: "Invalid risk score (negative)",
			analytics: func() *MerchantAnalytics {
				a := createTestMerchantAnalytics()
				a.RiskScore = -0.1
				return a
			}(),
			expectErr: true,
		},
		{
			name: "Invalid risk score (greater than 1)",
			analytics: func() *MerchantAnalytics {
				a := createTestMerchantAnalytics()
				a.RiskScore = 1.1
				return a
			}(),
			expectErr: true,
		},
		{
			name: "Invalid compliance score (negative)",
			analytics: func() *MerchantAnalytics {
				a := createTestMerchantAnalytics()
				a.ComplianceScore = -0.1
				return a
			}(),
			expectErr: true,
		},
		{
			name: "Invalid compliance score (greater than 1)",
			analytics: func() *MerchantAnalytics {
				a := createTestMerchantAnalytics()
				a.ComplianceScore = 1.1
				return a
			}(),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.analytics.Validate()
			if tt.expectErr && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}

// MerchantPortfolioSummary tests
func TestMerchantPortfolioSummary(t *testing.T) {
	summary := &MerchantPortfolioSummary{
		TotalMerchants:   100,
		OnboardedCount:   50,
		ProspectiveCount: 30,
		PendingCount:     15,
		DeactivatedCount: 5,
		HighRiskCount:    20,
		MediumRiskCount:  60,
		LowRiskCount:     20,
		IndustryBreakdown: map[string]int{
			"Technology": 40,
			"Finance":    30,
			"Retail":     30,
		},
		ComplianceStatus: map[string]int{
			"compliant":     80,
			"pending":       15,
			"non_compliant": 5,
		},
		RecentActivity: []*MerchantActivity{
			{
				MerchantID: "merchant_123",
				Action:     "UPDATE",
				Details:    "Risk level updated",
				UserID:     "user123",
				Timestamp:  time.Now(),
			},
		},
		RiskTrends: []*RiskTrend{
			{
				Date:       time.Now(),
				HighRisk:   20,
				MediumRisk: 60,
				LowRisk:    20,
			},
		},
		GeneratedAt: time.Now(),
	}

	// Test that all fields are properly set
	if summary.TotalMerchants != 100 {
		t.Error("Expected TotalMerchants to be 100")
	}

	if summary.OnboardedCount != 50 {
		t.Error("Expected OnboardedCount to be 50")
	}

	if len(summary.IndustryBreakdown) != 3 {
		t.Error("Expected IndustryBreakdown to have 3 entries")
	}

	if len(summary.RecentActivity) != 1 {
		t.Error("Expected RecentActivity to have 1 entry")
	}

	if len(summary.RiskTrends) != 1 {
		t.Error("Expected RiskTrends to have 1 entry")
	}
}
