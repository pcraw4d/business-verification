package mock_data

import (
	"testing"
	"time"

	"kyb-platform/internal/models"
)

func TestGetTestDataSets(t *testing.T) {
	dataSets := GetTestDataSets()

	if dataSets == nil {
		t.Fatal("GetTestDataSets() returned nil")
	}

	// Test that all data sets are populated
	if len(dataSets.BasicMerchants) == 0 {
		t.Error("BasicMerchants should not be empty")
	}

	if len(dataSets.EdgeCaseMerchants) == 0 {
		t.Error("EdgeCaseMerchants should not be empty")
	}

	if len(dataSets.PerformanceData) == 0 {
		t.Error("PerformanceData should not be empty")
	}

	if len(dataSets.ValidationData) == 0 {
		t.Error("ValidationData should not be empty")
	}

	if len(dataSets.BulkOperationData) == 0 {
		t.Error("BulkOperationData should not be empty")
	}

	if len(dataSets.ComparisonData) == 0 {
		t.Error("ComparisonData should not be empty")
	}

	if len(dataSets.SessionData) == 0 {
		t.Error("SessionData should not be empty")
	}

	if len(dataSets.AuditLogData) == 0 {
		t.Error("AuditLogData should not be empty")
	}

	if len(dataSets.NotificationData) == 0 {
		t.Error("NotificationData should not be empty")
	}

	if len(dataSets.AnalyticsData) == 0 {
		t.Error("AnalyticsData should not be empty")
	}
}

func TestGetBasicMerchants(t *testing.T) {
	merchants := GetBasicMerchants()

	if len(merchants) != 3 {
		t.Errorf("Expected 3 basic merchants, got %d", len(merchants))
	}

	for i, merchant := range merchants {
		// Test required fields
		if merchant.ID == "" {
			t.Errorf("Merchant %d: ID should not be empty", i)
		}

		if merchant.Name == "" {
			t.Errorf("Merchant %d: Name should not be empty", i)
		}

		if merchant.LegalName == "" {
			t.Errorf("Merchant %d: LegalName should not be empty", i)
		}

		if merchant.Industry == "" {
			t.Errorf("Merchant %d: Industry should not be empty", i)
		}

		if merchant.BusinessType == "" {
			t.Errorf("Merchant %d: BusinessType should not be empty", i)
		}

		if merchant.EmployeeCount <= 0 {
			t.Errorf("Merchant %d: EmployeeCount should be positive, got %d", i, merchant.EmployeeCount)
		}

		if merchant.AnnualRevenue == nil || *merchant.AnnualRevenue <= 0 {
			t.Errorf("Merchant %d: AnnualRevenue should be positive", i)
		}

		// Test portfolio type validity
		if !merchant.PortfolioType.IsValid() {
			t.Errorf("Merchant %d: Invalid portfolio type: %s", i, merchant.PortfolioType)
		}

		// Test risk level validity
		if !merchant.RiskLevel.IsValid() {
			t.Errorf("Merchant %d: Invalid risk level: %s", i, merchant.RiskLevel)
		}

		// Test address fields
		if merchant.Address.Street1 == "" {
			t.Errorf("Merchant %d: Address.Street1 should not be empty", i)
		}

		if merchant.Address.City == "" {
			t.Errorf("Merchant %d: Address.City should not be empty", i)
		}

		if merchant.Address.State == "" {
			t.Errorf("Merchant %d: Address.State should not be empty", i)
		}

		if merchant.Address.PostalCode == "" {
			t.Errorf("Merchant %d: Address.PostalCode should not be empty", i)
		}

		if merchant.Address.Country == "" {
			t.Errorf("Merchant %d: Address.Country should not be empty", i)
		}

		// Test contact info
		if merchant.ContactInfo.Phone == "" {
			t.Errorf("Merchant %d: ContactInfo.Phone should not be empty", i)
		}

		if merchant.ContactInfo.Email == "" {
			t.Errorf("Merchant %d: ContactInfo.Email should not be empty", i)
		}

		// Test timestamps
		if merchant.CreatedAt.IsZero() {
			t.Errorf("Merchant %d: CreatedAt should not be zero", i)
		}

		if merchant.UpdatedAt.IsZero() {
			t.Errorf("Merchant %d: UpdatedAt should not be zero", i)
		}

		if merchant.CreatedAt.After(merchant.UpdatedAt) {
			t.Errorf("Merchant %d: CreatedAt should not be after UpdatedAt", i)
		}
	}
}

func TestGetEdgeCaseMerchants(t *testing.T) {
	merchants := GetEdgeCaseMerchants()

	if len(merchants) != 3 {
		t.Errorf("Expected 3 edge case merchants, got %d", len(merchants))
	}

	// Test minimal merchant
	minimalMerchant := merchants[0]
	if minimalMerchant.EmployeeCount != 1 {
		t.Errorf("Minimal merchant should have 1 employee, got %d", minimalMerchant.EmployeeCount)
	}

	if minimalMerchant.AnnualRevenue != nil {
		t.Error("Minimal merchant should not have annual revenue set")
	}

	// Test maximum merchant
	maxMerchant := merchants[1]
	if len(maxMerchant.Name) < 50 {
		t.Error("Maximum merchant should have a very long name")
	}

	if maxMerchant.EmployeeCount != 10000 {
		t.Errorf("Maximum merchant should have 10000 employees, got %d", maxMerchant.EmployeeCount)
	}

	if maxMerchant.AnnualRevenue == nil || *maxMerchant.AnnualRevenue != 999999999.99 {
		t.Error("Maximum merchant should have maximum revenue")
	}

	// Test special characters merchant
	specialMerchant := merchants[2]
	if specialMerchant.PortfolioType != models.PortfolioTypeDeactivated {
		t.Errorf("Special merchant should be deactivated, got %s", specialMerchant.PortfolioType)
	}

	if specialMerchant.RiskLevel != models.RiskLevelMedium {
		t.Errorf("Special merchant should have medium risk, got %s", specialMerchant.RiskLevel)
	}

	if specialMerchant.ComplianceStatus != "non_compliant" {
		t.Errorf("Special merchant should be non-compliant, got %s", specialMerchant.ComplianceStatus)
	}

	if specialMerchant.Status != "inactive" {
		t.Errorf("Special merchant should be inactive, got %s", specialMerchant.Status)
	}
}

func TestGetPerformanceTestData(t *testing.T) {
	merchants := GetPerformanceTestData()

	if len(merchants) != 1000 {
		t.Errorf("Expected 1000 performance test merchants, got %d", len(merchants))
	}

	// Test first few merchants
	for i := 0; i < 10; i++ {
		merchant := merchants[i]

		if merchant.ID == "" {
			t.Errorf("Performance merchant %d: ID should not be empty", i)
		}

		if merchant.Name == "" {
			t.Errorf("Performance merchant %d: Name should not be empty", i)
		}

		if merchant.EmployeeCount <= 0 {
			t.Errorf("Performance merchant %d: EmployeeCount should be positive, got %d", i, merchant.EmployeeCount)
		}

		if merchant.AnnualRevenue == nil || *merchant.AnnualRevenue <= 0 {
			t.Errorf("Performance merchant %d: AnnualRevenue should be positive", i)
		}

		if !merchant.PortfolioType.IsValid() {
			t.Errorf("Performance merchant %d: Invalid portfolio type: %s", i, merchant.PortfolioType)
		}

		if !merchant.RiskLevel.IsValid() {
			t.Errorf("Performance merchant %d: Invalid risk level: %s", i, merchant.RiskLevel)
		}
	}
}

func TestGetValidationTestData(t *testing.T) {
	merchants := GetValidationTestData()

	if len(merchants) != 3 {
		t.Errorf("Expected 3 validation test merchants, got %d", len(merchants))
	}

	// Test invalid merchant (empty name)
	invalidMerchant := merchants[0]
	if invalidMerchant.Name != "" {
		t.Error("Invalid merchant should have empty name")
	}

	// Test invalid portfolio type
	invalidPortfolioMerchant := merchants[1]
	if invalidPortfolioMerchant.PortfolioType == "invalid_type" {
		// This should be invalid
		if invalidPortfolioMerchant.PortfolioType.IsValid() {
			t.Error("Invalid portfolio type should not be valid")
		}
	}

	// Test invalid risk level
	invalidRiskMerchant := merchants[2]
	if invalidRiskMerchant.RiskLevel == "invalid_risk" {
		// This should be invalid
		if invalidRiskMerchant.RiskLevel.IsValid() {
			t.Error("Invalid risk level should not be valid")
		}
	}
}

func TestGetBulkOperationData(t *testing.T) {
	merchants := GetBulkOperationData()

	if len(merchants) != 100 {
		t.Errorf("Expected 100 bulk operation merchants, got %d", len(merchants))
	}

	// Test that all merchants have the same portfolio type and risk level
	for i, merchant := range merchants {
		if merchant.PortfolioType != models.PortfolioTypeProspective {
			t.Errorf("Bulk merchant %d: Should be prospective, got %s", i, merchant.PortfolioType)
		}

		if merchant.RiskLevel != models.RiskLevelMedium {
			t.Errorf("Bulk merchant %d: Should have medium risk, got %s", i, merchant.RiskLevel)
		}

		if merchant.ComplianceStatus != "pending" {
			t.Errorf("Bulk merchant %d: Should be pending compliance, got %s", i, merchant.ComplianceStatus)
		}

		if merchant.Status != "active" {
			t.Errorf("Bulk merchant %d: Should be active, got %s", i, merchant.Status)
		}
	}
}

func TestGetComparisonData(t *testing.T) {
	merchants := GetComparisonData()

	if len(merchants) != 2 {
		t.Errorf("Expected 2 comparison merchants, got %d", len(merchants))
	}

	// Test that merchants are different for comparison
	merchant1 := merchants[0]
	merchant2 := merchants[1]

	if merchant1.ID == merchant2.ID {
		t.Error("Comparison merchants should have different IDs")
	}

	if merchant1.Name == merchant2.Name {
		t.Error("Comparison merchants should have different names")
	}

	if merchant1.Industry == merchant2.Industry {
		t.Error("Comparison merchants should have different industries")
	}

	if merchant1.PortfolioType == merchant2.PortfolioType {
		t.Error("Comparison merchants should have different portfolio types")
	}

	if merchant1.RiskLevel == merchant2.RiskLevel {
		t.Error("Comparison merchants should have different risk levels")
	}
}

func TestGetSessionTestData(t *testing.T) {
	sessions := GetSessionTestData()

	if len(sessions) != 3 {
		t.Errorf("Expected 3 test sessions, got %d", len(sessions))
	}

	// Test active sessions
	activeSessions := 0
	expiredSessions := 0

	for i, session := range sessions {
		if session.ID == "" {
			t.Errorf("Session %d: ID should not be empty", i)
		}

		if session.UserID == "" {
			t.Errorf("Session %d: UserID should not be empty", i)
		}

		if session.MerchantID == "" {
			t.Errorf("Session %d: MerchantID should not be empty", i)
		}

		if session.CreatedAt.IsZero() {
			t.Errorf("Session %d: CreatedAt should not be zero", i)
		}

		if session.UpdatedAt.IsZero() {
			t.Errorf("Session %d: UpdatedAt should not be zero", i)
		}

		if session.IsActive {
			activeSessions++
		} else {
			expiredSessions++
		}
	}

	if activeSessions != 2 {
		t.Errorf("Expected 2 active sessions, got %d", activeSessions)
	}

	if expiredSessions != 1 {
		t.Errorf("Expected 1 expired session, got %d", expiredSessions)
	}
}

func TestGetAuditLogTestData(t *testing.T) {
	auditLogs := GetAuditLogTestData()

	if len(auditLogs) != 3 {
		t.Errorf("Expected 3 test audit logs, got %d", len(auditLogs))
	}

	for i, log := range auditLogs {
		if log.ID == "" {
			t.Errorf("Audit log %d: ID should not be empty", i)
		}

		if log.UserID == "" {
			t.Errorf("Audit log %d: UserID should not be empty", i)
		}

		if log.MerchantID == "" {
			t.Errorf("Audit log %d: MerchantID should not be empty", i)
		}

		if log.Action == "" {
			t.Errorf("Audit log %d: Action should not be empty", i)
		}

		if log.ResourceType == "" {
			t.Errorf("Audit log %d: ResourceType should not be empty", i)
		}

		if log.ResourceID == "" {
			t.Errorf("Audit log %d: ResourceID should not be empty", i)
		}

		if log.CreatedAt.IsZero() {
			t.Errorf("Audit log %d: CreatedAt should not be zero", i)
		}
	}
}

func TestGetNotificationTestData(t *testing.T) {
	notifications := GetNotificationTestData()

	if len(notifications) != 3 {
		t.Errorf("Expected 3 test notifications, got %d", len(notifications))
	}

	for i, notification := range notifications {
		if notification.ID == "" {
			t.Errorf("Notification %d: ID should not be empty", i)
		}

		if notification.MerchantID == "" {
			t.Errorf("Notification %d: MerchantID should not be empty", i)
		}

		if notification.UserID == "" {
			t.Errorf("Notification %d: UserID should not be empty", i)
		}

		if notification.Type == "" {
			t.Errorf("Notification %d: Type should not be empty", i)
		}

		if notification.Title == "" {
			t.Errorf("Notification %d: Title should not be empty", i)
		}

		if notification.Message == "" {
			t.Errorf("Notification %d: Message should not be empty", i)
		}

		if notification.Priority == "" {
			t.Errorf("Notification %d: Priority should not be empty", i)
		}

		if notification.CreatedAt.IsZero() {
			t.Errorf("Notification %d: CreatedAt should not be zero", i)
		}
	}
}

func TestGetAnalyticsTestData(t *testing.T) {
	analytics := GetAnalyticsTestData()

	if len(analytics) != 3 {
		t.Errorf("Expected 3 test analytics records, got %d", len(analytics))
	}

	for i, analytic := range analytics {
		if analytic.MerchantID == "" {
			t.Errorf("Analytics %d: MerchantID should not be empty", i)
		}

		if analytic.RiskScore < 0 || analytic.RiskScore > 1 {
			t.Errorf("Analytics %d: RiskScore should be between 0 and 1, got %f", i, analytic.RiskScore)
		}

		if analytic.ComplianceScore < 0 || analytic.ComplianceScore > 1 {
			t.Errorf("Analytics %d: ComplianceScore should be between 0 and 1, got %f", i, analytic.ComplianceScore)
		}

		if analytic.TransactionVolume < 0 {
			t.Errorf("Analytics %d: TransactionVolume should be positive, got %f", i, analytic.TransactionVolume)
		}

		if analytic.CalculatedAt.IsZero() {
			t.Errorf("Analytics %d: CalculatedAt should not be zero", i)
		}

		if analytic.UpdatedAt.IsZero() {
			t.Errorf("Analytics %d: UpdatedAt should not be zero", i)
		}
	}
}

func TestGetTestDataByScenario(t *testing.T) {
	testCases := []struct {
		scenario string
		expected interface{}
	}{
		{"basic", GetBasicMerchants()},
		{"edge_cases", GetEdgeCaseMerchants()},
		{"performance", GetPerformanceTestData()},
		{"validation", GetValidationTestData()},
		{"bulk_operations", GetBulkOperationData()},
		{"comparison", GetComparisonData()},
		{"sessions", GetSessionTestData()},
		{"audit_logs", GetAuditLogTestData()},
		{"notifications", GetNotificationTestData()},
		{"analytics", GetAnalyticsTestData()},
		{"unknown", GetTestDataSets()},
	}

	for _, tc := range testCases {
		result := GetTestDataByScenario(tc.scenario)
		if result == nil {
			t.Errorf("GetTestDataByScenario(%s) returned nil", tc.scenario)
		}
	}
}

func TestGetTestDataCount(t *testing.T) {
	testCases := []struct {
		scenario string
		expected int
	}{
		{"basic", 3},
		{"edge_cases", 3},
		{"performance", 1000},
		{"validation", 3},
		{"bulk_operations", 100},
		{"comparison", 2},
		{"sessions", 3},
		{"audit_logs", 3},
		{"notifications", 3},
		{"analytics", 3},
		{"unknown", 0},
	}

	for _, tc := range testCases {
		result := GetTestDataCount(tc.scenario)
		if result != tc.expected {
			t.Errorf("GetTestDataCount(%s) expected %d, got %d", tc.scenario, tc.expected, result)
		}
	}
}

func TestMerchantValidation(t *testing.T) {
	// Test valid merchants
	validMerchants := GetBasicMerchants()
	for i, merchant := range validMerchants {
		if err := merchant.Validate(); err != nil {
			t.Errorf("Valid merchant %d should not have validation errors: %v", i, err)
		}
	}

	// Test invalid merchants
	invalidMerchants := GetValidationTestData()
	for i, merchant := range invalidMerchants {
		if err := merchant.Validate(); err == nil {
			t.Errorf("Invalid merchant %d should have validation errors", i)
		}
	}
}

func TestSessionValidation(t *testing.T) {
	sessions := GetSessionTestData()

	for i, session := range sessions {
		if err := session.Validate(); err != nil {
			t.Errorf("Session %d should not have validation errors: %v", i, err)
		}
	}
}

func TestAuditLogValidation(t *testing.T) {
	auditLogs := GetAuditLogTestData()

	for i, log := range auditLogs {
		if err := log.Validate(); err != nil {
			t.Errorf("Audit log %d should not have validation errors: %v", i, err)
		}
	}
}

func TestNotificationValidation(t *testing.T) {
	notifications := GetNotificationTestData()

	for i, notification := range notifications {
		if err := notification.Validate(); err != nil {
			t.Errorf("Notification %d should not have validation errors: %v", i, err)
		}
	}
}

func TestAnalyticsValidation(t *testing.T) {
	analytics := GetAnalyticsTestData()

	for i, analytic := range analytics {
		if err := analytic.Validate(); err != nil {
			t.Errorf("Analytics %d should not have validation errors: %v", i, err)
		}
	}
}

func TestDataConsistency(t *testing.T) {
	// Test that all merchant IDs are unique across all data sets
	allMerchants := append(GetBasicMerchants(), GetEdgeCaseMerchants()...)
	allMerchants = append(allMerchants, GetPerformanceTestData()...)
	allMerchants = append(allMerchants, GetValidationTestData()...)
	allMerchants = append(allMerchants, GetBulkOperationData()...)
	allMerchants = append(allMerchants, GetComparisonData()...)

	merchantIDs := make(map[string]bool)
	for _, merchant := range allMerchants {
		if merchantIDs[merchant.ID] {
			t.Errorf("Duplicate merchant ID found: %s", merchant.ID)
		}
		merchantIDs[merchant.ID] = true
	}

	// Test that all session IDs are unique
	sessions := GetSessionTestData()
	sessionIDs := make(map[string]bool)
	for _, session := range sessions {
		if sessionIDs[session.ID] {
			t.Errorf("Duplicate session ID found: %s", session.ID)
		}
		sessionIDs[session.ID] = true
	}

	// Test that all audit log IDs are unique
	auditLogs := GetAuditLogTestData()
	auditLogIDs := make(map[string]bool)
	for _, log := range auditLogs {
		if auditLogIDs[log.ID] {
			t.Errorf("Duplicate audit log ID found: %s", log.ID)
		}
		auditLogIDs[log.ID] = true
	}

	// Test that all notification IDs are unique
	notifications := GetNotificationTestData()
	notificationIDs := make(map[string]bool)
	for _, notification := range notifications {
		if notificationIDs[notification.ID] {
			t.Errorf("Duplicate notification ID found: %s", notification.ID)
		}
		notificationIDs[notification.ID] = true
	}
}

func TestTimeConsistency(t *testing.T) {
	now := time.Now()

	// Test that all timestamps are in the past
	allMerchants := append(GetBasicMerchants(), GetEdgeCaseMerchants()...)
	allMerchants = append(allMerchants, GetValidationTestData()...)
	allMerchants = append(allMerchants, GetBulkOperationData()...)
	allMerchants = append(allMerchants, GetComparisonData()...)

	for i, merchant := range allMerchants {
		if merchant.CreatedAt.After(now) {
			t.Errorf("Merchant %d: CreatedAt should be in the past", i)
		}

		if merchant.UpdatedAt.After(now) {
			t.Errorf("Merchant %d: UpdatedAt should be in the past", i)
		}

		if merchant.CreatedAt.After(merchant.UpdatedAt) {
			t.Errorf("Merchant %d: CreatedAt should not be after UpdatedAt", i)
		}
	}

	// Test session timestamps
	sessions := GetSessionTestData()
	for i, session := range sessions {
		if session.CreatedAt.After(now) {
			t.Errorf("Session %d: CreatedAt should be in the past", i)
		}

		if session.UpdatedAt.After(now) {
			t.Errorf("Session %d: UpdatedAt should be in the past", i)
		}

		if session.StartedAt.After(now) {
			t.Errorf("Session %d: StartedAt should be in the past", i)
		}

		if session.LastActive.After(now) {
			t.Errorf("Session %d: LastActive should be in the past", i)
		}
	}
}
