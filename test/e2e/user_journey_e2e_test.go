package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/database"
	"kyb-platform/internal/services"
	"kyb-platform/test/mocks"
)

// TestUserJourneyE2E tests complete user journeys end-to-end
func TestUserJourneyE2E(t *testing.T) {
	// Skip if not running E2E tests
	if os.Getenv("E2E_TESTS") != "true" {
		t.Skip("Skipping E2E tests - set E2E_TESTS=true to run")
	}

	// Setup test environment
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	// Create mock services
	mockService := mocks.NewMockMerchantPortfolioService()
	comparisonService := mocks.NewMockComparisonService()
	handler := handlers.NewMerchantPortfolioHandler(mockService, nil)
	comparisonHandler := handlers.NewComparisonHandler(comparisonService, nil)

	// Create test server
	server := httptest.NewServer(createUserJourneyTestRouter(handler, comparisonHandler))
	defer server.Close()

	// Test 1: New user onboarding journey
	t.Run("NewUserOnboardingJourney", func(t *testing.T) {
		testNewUserOnboardingJourney(t, server, mockService, ctx)
	})

	// Test 2: Portfolio management journey
	t.Run("PortfolioManagementJourney", func(t *testing.T) {
		testPortfolioManagementJourney(t, server, mockService, ctx)
	})

	// Test 3: Risk assessment journey
	t.Run("RiskAssessmentJourney", func(t *testing.T) {
		testRiskAssessmentJourney(t, server, mockService, ctx)
	})

	// Test 4: Compliance monitoring journey
	t.Run("ComplianceMonitoringJourney", func(t *testing.T) {
		testComplianceMonitoringJourney(t, server, mockService, ctx)
	})

	// Test 5: Reporting and analytics journey
	t.Run("ReportingAndAnalyticsJourney", func(t *testing.T) {
		testReportingAndAnalyticsJourney(t, server, mockService, comparisonService, ctx)
	})
}

// testNewUserOnboardingJourney tests the new user onboarding journey
func testNewUserOnboardingJourney(t *testing.T, server *httptest.Server, mockService *mocks.MockMerchantPortfolioService, ctx context.Context) {
	t.Log("=== New User Onboarding Journey ===")

	// Step 1: User creates their first merchant
	t.Log("Step 1: Creating first merchant...")
	merchant := createUserJourneyTestMerchant("onboarding_merchant_1", "New User Company", services.PortfolioTypeProspective, services.RiskLevelMedium)

	createReq := &handlers.CreateMerchantRequest{
		Name:               merchant.Name,
		LegalName:          merchant.LegalName,
		RegistrationNumber: merchant.RegistrationNumber,
		TaxID:              merchant.TaxID,
		Industry:           merchant.Industry,
		IndustryCode:       merchant.IndustryCode,
		BusinessType:       merchant.BusinessType,
		EmployeeCount:      merchant.EmployeeCount,
		Address:            merchant.Address,
		ContactInfo:        merchant.ContactInfo,
		PortfolioType:      string(merchant.PortfolioType),
		RiskLevel:          string(merchant.RiskLevel),
	}

	createResp, err := makeRequest(t, server, "POST", "/api/v1/merchants", createReq)
	if err != nil {
		t.Fatalf("Failed to create first merchant: %v", err)
	}

	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d: %s", createResp.StatusCode, createResp.Body)
	}

	var createdMerchant handlers.MerchantResponse
	if err := json.Unmarshal(createResp.Body, &createdMerchant); err != nil {
		t.Fatalf("Failed to unmarshal created merchant: %v", err)
	}

	t.Logf("✅ First merchant created: %s (ID: %s)", createdMerchant.Name, createdMerchant.ID)

	// Step 2: User starts a session with the merchant
	t.Log("Step 2: Starting merchant session...")
	sessionResp, err := makeRequest(t, server, "POST", fmt.Sprintf("/api/v1/merchants/%s/session", createdMerchant.ID), nil)
	if err != nil {
		t.Fatalf("Failed to start merchant session: %v", err)
	}

	if sessionResp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d: %s", sessionResp.StatusCode, sessionResp.Body)
	}

	var session handlers.SessionResponse
	if err := json.Unmarshal(sessionResp.Body, &session); err != nil {
		t.Fatalf("Failed to unmarshal session: %v", err)
	}

	t.Logf("✅ Merchant session started: %s", session.ID)

	// Step 3: User views merchant dashboard
	t.Log("Step 3: Viewing merchant dashboard...")
	dashboardResp, err := makeRequest(t, server, "GET", fmt.Sprintf("/api/v1/merchants/%s/dashboard", createdMerchant.ID), nil)
	if err != nil {
		t.Fatalf("Failed to get merchant dashboard: %v", err)
	}

	if dashboardResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", dashboardResp.StatusCode, dashboardResp.Body)
	}

	var dashboard handlers.MerchantDashboardResponse
	if err := json.Unmarshal(dashboardResp.Body, &dashboard); err != nil {
		t.Fatalf("Failed to unmarshal dashboard: %v", err)
	}

	if dashboard.MerchantID != createdMerchant.ID {
		t.Errorf("Expected merchant ID %s, got %s", createdMerchant.ID, dashboard.MerchantID)
	}

	t.Logf("✅ Merchant dashboard viewed successfully")

	// Step 4: User updates merchant information
	t.Log("Step 4: Updating merchant information...")
	updateReq := &handlers.UpdateMerchantRequest{
		Name:          "Updated New User Company",
		EmployeeCount: 75,
		PortfolioType: string(services.PortfolioTypePending),
		RiskLevel:     string(services.RiskLevelLow),
	}

	updateResp, err := makeRequest(t, server, "PUT", fmt.Sprintf("/api/v1/merchants/%s", createdMerchant.ID), updateReq)
	if err != nil {
		t.Fatalf("Failed to update merchant: %v", err)
	}

	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", updateResp.StatusCode, updateResp.Body)
	}

	t.Logf("✅ Merchant information updated successfully")

	// Step 5: User ends the session
	t.Log("Step 5: Ending merchant session...")
	endSessionResp, err := makeRequest(t, server, "DELETE", "/api/v1/merchants/session/active", nil)
	if err != nil {
		t.Fatalf("Failed to end merchant session: %v", err)
	}

	if endSessionResp.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected status 204, got %d: %s", endSessionResp.StatusCode, endSessionResp.Body)
	}

	t.Logf("✅ Merchant session ended successfully")

	t.Log("=== New User Onboarding Journey Completed ===")
}

// testPortfolioManagementJourney tests the portfolio management journey
func testPortfolioManagementJourney(t *testing.T, server *httptest.Server, mockService *mocks.MockMerchantPortfolioService, ctx context.Context) {
	t.Log("=== Portfolio Management Journey ===")

	// Step 1: Create multiple merchants for portfolio
	t.Log("Step 1: Creating portfolio of merchants...")
	merchants := []*services.Merchant{
		createUserJourneyTestMerchant("portfolio_1", "Portfolio Company A", services.PortfolioTypeOnboarded, services.RiskLevelLow),
		createUserJourneyTestMerchant("portfolio_2", "Portfolio Company B", services.PortfolioTypeProspective, services.RiskLevelMedium),
		createUserJourneyTestMerchant("portfolio_3", "Portfolio Company C", services.PortfolioTypePending, services.RiskLevelHigh),
		createUserJourneyTestMerchant("portfolio_4", "Portfolio Company D", services.PortfolioTypeOnboarded, services.RiskLevelLow),
		createUserJourneyTestMerchant("portfolio_5", "Portfolio Company E", services.PortfolioTypeProspective, services.RiskLevelMedium),
	}

	for _, merchant := range merchants {
		mockService.AddMerchant(merchant)
	}

	t.Logf("✅ Created portfolio with %d merchants", len(merchants))

	// Step 2: View portfolio overview
	t.Log("Step 2: Viewing portfolio overview...")
	portfolioResp, err := makeRequest(t, server, "GET", "/api/v1/merchants?page=1&page_size=10", nil)
	if err != nil {
		t.Fatalf("Failed to get portfolio overview: %v", err)
	}

	if portfolioResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", portfolioResp.StatusCode, portfolioResp.Body)
	}

	var portfolio handlers.MerchantListResponse
	if err := json.Unmarshal(portfolioResp.Body, &portfolio); err != nil {
		t.Fatalf("Failed to unmarshal portfolio: %v", err)
	}

	if len(portfolio.Merchants) != len(merchants) {
		t.Errorf("Expected %d merchants, got %d", len(merchants), len(portfolio.Merchants))
	}

	t.Logf("✅ Portfolio overview viewed: %d merchants", len(portfolio.Merchants))

	// Step 3: Filter portfolio by portfolio type
	t.Log("Step 3: Filtering portfolio by type...")
	filterResp, err := makeRequest(t, server, "GET", "/api/v1/merchants?portfolio_type=onboarded&page=1&page_size=10", nil)
	if err != nil {
		t.Fatalf("Failed to filter portfolio: %v", err)
	}

	if filterResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", filterResp.StatusCode, filterResp.Body)
	}

	var filteredPortfolio handlers.MerchantListResponse
	if err := json.Unmarshal(filterResp.Body, &filteredPortfolio); err != nil {
		t.Fatalf("Failed to unmarshal filtered portfolio: %v", err)
	}

	// Should have 2 onboarded merchants
	if len(filteredPortfolio.Merchants) != 2 {
		t.Errorf("Expected 2 onboarded merchants, got %d", len(filteredPortfolio.Merchants))
	}

	t.Logf("✅ Portfolio filtered by type: %d onboarded merchants", len(filteredPortfolio.Merchants))

	// Step 4: Search portfolio
	t.Log("Step 4: Searching portfolio...")
	searchReq := &handlers.SearchMerchantsRequest{
		Query: "Portfolio Company",
		Filters: &handlers.MerchantSearchFilters{
			PortfolioType: "",
			RiskLevel:     "",
		},
		Page:     1,
		PageSize: 10,
	}

	searchResp, err := makeRequest(t, server, "POST", "/api/v1/merchants/search", searchReq)
	if err != nil {
		t.Fatalf("Failed to search portfolio: %v", err)
	}

	if searchResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", searchResp.StatusCode, searchResp.Body)
	}

	var searchResult handlers.MerchantListResponse
	if err := json.Unmarshal(searchResp.Body, &searchResult); err != nil {
		t.Fatalf("Failed to unmarshal search results: %v", err)
	}

	if len(searchResult.Merchants) == 0 {
		t.Error("Expected to find merchants in search results")
	}

	t.Logf("✅ Portfolio search completed: %d results", len(searchResult.Merchants))

	// Step 5: Bulk update portfolio
	t.Log("Step 5: Performing bulk portfolio update...")
	merchantIDs := make([]string, len(merchants))
	for i, merchant := range merchants {
		merchantIDs[i] = merchant.ID
	}

	bulkReq := &handlers.BulkUpdateRequest{
		MerchantIDs:   merchantIDs,
		PortfolioType: string(services.PortfolioTypeOnboarded),
	}

	bulkResp, err := makeRequest(t, server, "POST", "/api/v1/merchants/bulk/portfolio-type", bulkReq)
	if err != nil {
		t.Fatalf("Failed to perform bulk update: %v", err)
	}

	if bulkResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", bulkResp.StatusCode, bulkResp.Body)
	}

	var bulkResult handlers.BulkOperationResponse
	if err := json.Unmarshal(bulkResp.Body, &bulkResult); err != nil {
		t.Fatalf("Failed to unmarshal bulk result: %v", err)
	}

	if bulkResult.SuccessfulUpdates != len(merchants) {
		t.Errorf("Expected %d successful updates, got %d", len(merchants), bulkResult.SuccessfulUpdates)
	}

	t.Logf("✅ Bulk portfolio update completed: %d merchants updated", bulkResult.SuccessfulUpdates)

	t.Log("=== Portfolio Management Journey Completed ===")
}

// testRiskAssessmentJourney tests the risk assessment journey
func testRiskAssessmentJourney(t *testing.T, server *httptest.Server, mockService *mocks.MockMerchantPortfolioService, ctx context.Context) {
	t.Log("=== Risk Assessment Journey ===")

	// Step 1: Create merchants with different risk levels
	t.Log("Step 1: Creating merchants with different risk levels...")
	riskMerchants := []*services.Merchant{
		createUserJourneyTestMerchant("risk_low_1", "Low Risk Company A", services.PortfolioTypeOnboarded, services.RiskLevelLow),
		createUserJourneyTestMerchant("risk_medium_1", "Medium Risk Company B", services.PortfolioTypeOnboarded, services.RiskLevelMedium),
		createUserJourneyTestMerchant("risk_high_1", "High Risk Company C", services.PortfolioTypeOnboarded, services.RiskLevelHigh),
	}

	for _, merchant := range riskMerchants {
		mockService.AddMerchant(merchant)
	}

	t.Logf("✅ Created %d merchants with different risk levels", len(riskMerchants))

	// Step 2: View risk assessment dashboard
	t.Log("Step 2: Viewing risk assessment dashboard...")
	riskDashboardResp, err := makeRequest(t, server, "GET", "/api/v1/merchants/risk/dashboard", nil)
	if err != nil {
		t.Fatalf("Failed to get risk dashboard: %v", err)
	}

	if riskDashboardResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", riskDashboardResp.StatusCode, riskDashboardResp.Body)
	}

	var riskDashboard handlers.RiskDashboardResponse
	if err := json.Unmarshal(riskDashboardResp.Body, &riskDashboard); err != nil {
		t.Fatalf("Failed to unmarshal risk dashboard: %v", err)
	}

	if riskDashboard.TotalMerchants != len(riskMerchants) {
		t.Errorf("Expected %d total merchants, got %d", len(riskMerchants), riskDashboard.TotalMerchants)
	}

	t.Logf("✅ Risk assessment dashboard viewed: %d total merchants", riskDashboard.TotalMerchants)

	// Step 3: Filter by risk level
	t.Log("Step 3: Filtering by risk level...")
	highRiskResp, err := makeRequest(t, server, "GET", "/api/v1/merchants?risk_level=high&page=1&page_size=10", nil)
	if err != nil {
		t.Fatalf("Failed to filter by high risk: %v", err)
	}

	if highRiskResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", highRiskResp.StatusCode, highRiskResp.Body)
	}

	var highRiskResult handlers.MerchantListResponse
	if err := json.Unmarshal(highRiskResp.Body, &highRiskResult); err != nil {
		t.Fatalf("Failed to unmarshal high risk results: %v", err)
	}

	if len(highRiskResult.Merchants) != 1 {
		t.Errorf("Expected 1 high risk merchant, got %d", len(highRiskResult.Merchants))
	}

	t.Logf("✅ High risk merchants filtered: %d found", len(highRiskResult.Merchants))

	// Step 4: Bulk update risk levels
	t.Log("Step 4: Performing bulk risk level update...")
	merchantIDs := make([]string, len(riskMerchants))
	for i, merchant := range riskMerchants {
		merchantIDs[i] = merchant.ID
	}

	bulkRiskReq := &handlers.BulkUpdateRequest{
		MerchantIDs: merchantIDs,
		RiskLevel:   string(services.RiskLevelLow),
	}

	bulkRiskResp, err := makeRequest(t, server, "POST", "/api/v1/merchants/bulk/risk-level", bulkRiskReq)
	if err != nil {
		t.Fatalf("Failed to perform bulk risk update: %v", err)
	}

	if bulkRiskResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", bulkRiskResp.StatusCode, bulkRiskResp.Body)
	}

	var bulkRiskResult handlers.BulkOperationResponse
	if err := json.Unmarshal(bulkRiskResp.Body, &bulkRiskResult); err != nil {
		t.Fatalf("Failed to unmarshal bulk risk result: %v", err)
	}

	if bulkRiskResult.SuccessfulUpdates != len(riskMerchants) {
		t.Errorf("Expected %d successful risk updates, got %d", len(riskMerchants), bulkRiskResult.SuccessfulUpdates)
	}

	t.Logf("✅ Bulk risk level update completed: %d merchants updated", bulkRiskResult.SuccessfulUpdates)

	t.Log("=== Risk Assessment Journey Completed ===")
}

// testComplianceMonitoringJourney tests the compliance monitoring journey
func testComplianceMonitoringJourney(t *testing.T, server *httptest.Server, mockService *mocks.MockMerchantPortfolioService, ctx context.Context) {
	t.Log("=== Compliance Monitoring Journey ===")

	// Step 1: Create merchants with different compliance statuses
	t.Log("Step 1: Creating merchants with different compliance statuses...")
	complianceMerchants := []*services.Merchant{
		createUserJourneyTestMerchant("compliance_1", "Compliant Company A", services.PortfolioTypeOnboarded, services.RiskLevelLow),
		createUserJourneyTestMerchant("compliance_2", "Pending Company B", services.PortfolioTypePending, services.RiskLevelMedium),
		createUserJourneyTestMerchant("compliance_3", "Review Company C", services.PortfolioTypeProspective, services.RiskLevelHigh),
	}

	for _, merchant := range complianceMerchants {
		mockService.AddMerchant(merchant)
	}

	t.Logf("✅ Created %d merchants with different compliance statuses", len(complianceMerchants))

	// Step 2: View compliance dashboard
	t.Log("Step 2: Viewing compliance dashboard...")
	complianceResp, err := makeRequest(t, server, "GET", "/api/v1/merchants/compliance/dashboard", nil)
	if err != nil {
		t.Fatalf("Failed to get compliance dashboard: %v", err)
	}

	if complianceResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", complianceResp.StatusCode, complianceResp.Body)
	}

	var complianceDashboard handlers.ComplianceDashboardResponse
	if err := json.Unmarshal(complianceResp.Body, &complianceDashboard); err != nil {
		t.Fatalf("Failed to unmarshal compliance dashboard: %v", err)
	}

	if complianceDashboard.TotalMerchants != len(complianceMerchants) {
		t.Errorf("Expected %d total merchants, got %d", len(complianceMerchants), complianceDashboard.TotalMerchants)
	}

	t.Logf("✅ Compliance dashboard viewed: %d total merchants", complianceDashboard.TotalMerchants)

	// Step 3: Get compliance alerts
	t.Log("Step 3: Getting compliance alerts...")
	alertsResp, err := makeRequest(t, server, "GET", "/api/v1/merchants/compliance/alerts", nil)
	if err != nil {
		t.Fatalf("Failed to get compliance alerts: %v", err)
	}

	if alertsResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", alertsResp.StatusCode, alertsResp.Body)
	}

	var alerts handlers.ComplianceAlertsResponse
	if err := json.Unmarshal(alertsResp.Body, &alerts); err != nil {
		t.Fatalf("Failed to unmarshal compliance alerts: %v", err)
	}

	t.Logf("✅ Compliance alerts retrieved: %d alerts", len(alerts.Alerts))

	// Step 4: Generate compliance report
	t.Log("Step 4: Generating compliance report...")
	reportReq := &handlers.ComplianceReportRequest{
		ReportType:     "summary",
		DateRange:      "last_30_days",
		IncludeDetails: true,
	}

	reportResp, err := makeRequest(t, server, "POST", "/api/v1/merchants/compliance/reports", reportReq)
	if err != nil {
		t.Fatalf("Failed to generate compliance report: %v", err)
	}

	if reportResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", reportResp.StatusCode, reportResp.Body)
	}

	var report handlers.ComplianceReportResponse
	if err := json.Unmarshal(reportResp.Body, &report); err != nil {
		t.Fatalf("Failed to unmarshal compliance report: %v", err)
	}

	if report.ReportID == "" {
		t.Fatal("Expected report ID to be generated")
	}

	t.Logf("✅ Compliance report generated: %s", report.ReportID)

	t.Log("=== Compliance Monitoring Journey Completed ===")
}

// testReportingAndAnalyticsJourney tests the reporting and analytics journey
func testReportingAndAnalyticsJourney(t *testing.T, server *httptest.Server, mockService *mocks.MockMerchantPortfolioService, comparisonService *mocks.MockComparisonService, ctx context.Context) {
	t.Log("=== Reporting and Analytics Journey ===")

	// Step 1: Create merchants for analytics
	t.Log("Step 1: Creating merchants for analytics...")
	analyticsMerchants := []*services.Merchant{
		createUserJourneyTestMerchant("analytics_1", "Analytics Company A", services.PortfolioTypeOnboarded, services.RiskLevelLow),
		createUserJourneyTestMerchant("analytics_2", "Analytics Company B", services.PortfolioTypeOnboarded, services.RiskLevelMedium),
	}

	for _, merchant := range analyticsMerchants {
		mockService.AddMerchant(merchant)
	}

	t.Logf("✅ Created %d merchants for analytics", len(analyticsMerchants))

	// Step 2: View analytics dashboard
	t.Log("Step 2: Viewing analytics dashboard...")
	analyticsResp, err := makeRequest(t, server, "GET", "/api/v1/merchants/analytics/dashboard", nil)
	if err != nil {
		t.Fatalf("Failed to get analytics dashboard: %v", err)
	}

	if analyticsResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", analyticsResp.StatusCode, analyticsResp.Body)
	}

	var analyticsDashboard handlers.AnalyticsDashboardResponse
	if err := json.Unmarshal(analyticsResp.Body, &analyticsDashboard); err != nil {
		t.Fatalf("Failed to unmarshal analytics dashboard: %v", err)
	}

	if analyticsDashboard.TotalMerchants != len(analyticsMerchants) {
		t.Errorf("Expected %d total merchants, got %d", len(analyticsMerchants), analyticsDashboard.TotalMerchants)
	}

	t.Logf("✅ Analytics dashboard viewed: %d total merchants", analyticsDashboard.TotalMerchants)

	// Step 3: Compare merchants
	t.Log("Step 3: Comparing merchants...")
	comparisonReq := &handlers.ComparisonRequest{
		Merchant1ID:    analyticsMerchants[0].ID,
		Merchant2ID:    analyticsMerchants[1].ID,
		ComparisonType: "detailed",
	}

	comparisonResp, err := makeRequest(t, server, "POST", "/api/v1/merchants/compare", comparisonReq)
	if err != nil {
		t.Fatalf("Failed to compare merchants: %v", err)
	}

	if comparisonResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", comparisonResp.StatusCode, comparisonResp.Body)
	}

	var comparison handlers.ComparisonResponse
	if err := json.Unmarshal(comparisonResp.Body, &comparison); err != nil {
		t.Fatalf("Failed to unmarshal comparison: %v", err)
	}

	if comparison.ID == "" {
		t.Fatal("Expected comparison ID to be generated")
	}

	t.Logf("✅ Merchant comparison completed: %s", comparison.ID)

	// Step 4: Generate comparison report
	t.Log("Step 4: Generating comparison report...")
	reportReq := &handlers.ComparisonReportRequest{
		Merchant1ID:   analyticsMerchants[0].ID,
		Merchant2ID:   analyticsMerchants[1].ID,
		ReportType:    "pdf",
		IncludeCharts: true,
	}

	reportResp, err := makeRequest(t, server, "POST", "/api/v1/merchants/compare/report", reportReq)
	if err != nil {
		t.Fatalf("Failed to generate comparison report: %v", err)
	}

	if reportResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", reportResp.StatusCode, reportResp.Body)
	}

	var report handlers.ComparisonReportResponse
	if err := json.Unmarshal(reportResp.Body, &report); err != nil {
		t.Fatalf("Failed to unmarshal comparison report: %v", err)
	}

	if report.ReportID == "" {
		t.Fatal("Expected report ID to be generated")
	}

	t.Logf("✅ Comparison report generated: %s", report.ReportID)

	// Step 5: Export portfolio data
	t.Log("Step 5: Exporting portfolio data...")
	exportReq := &handlers.PortfolioExportRequest{
		Format:     "csv",
		IncludeAll: true,
		DateRange:  "all",
	}

	exportResp, err := makeRequest(t, server, "POST", "/api/v1/merchants/export", exportReq)
	if err != nil {
		t.Fatalf("Failed to export portfolio data: %v", err)
	}

	if exportResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", exportResp.StatusCode, exportResp.Body)
	}

	var export handlers.PortfolioExportResponse
	if err := json.Unmarshal(exportResp.Body, &export); err != nil {
		t.Fatalf("Failed to unmarshal export response: %v", err)
	}

	if export.ExportID == "" {
		t.Fatal("Expected export ID to be generated")
	}

	t.Logf("✅ Portfolio data exported: %s", export.ExportID)

	t.Log("=== Reporting and Analytics Journey Completed ===")
}

// Helper functions for user journey testing

func createUserJourneyTestMerchant(id, name string, portfolioType services.PortfolioType, riskLevel services.RiskLevel) *services.Merchant {
	now := time.Now()
	return &services.Merchant{
		ID:                 id,
		Name:               name,
		LegalName:          fmt.Sprintf("%s LLC", name),
		RegistrationNumber: fmt.Sprintf("REG%s", id),
		TaxID:              fmt.Sprintf("TAX%s", id),
		Industry:           "Technology",
		IndustryCode:       "541511",
		BusinessType:       "LLC",
		FoundedDate:        &now,
		EmployeeCount:      50,
		Address: database.Address{
			Street1:    "123 Test St",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "USA",
		},
		ContactInfo: database.ContactInfo{
			Phone:   "+1-555-123-4567",
			Email:   fmt.Sprintf("test@%s.com", id),
			Website: fmt.Sprintf("https://%s.com", id),
		},
		PortfolioType:    portfolioType,
		RiskLevel:        riskLevel,
		ComplianceStatus: "pending",
		Status:           "active",
		CreatedBy:        "test_user",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func createUserJourneyTestRouter(portfolioHandler *handlers.MerchantPortfolioHandler, comparisonHandler *handlers.ComparisonHandler) http.Handler {
	mux := http.NewServeMux()

	// Merchant portfolio routes
	mux.HandleFunc("POST /api/v1/merchants", portfolioHandler.CreateMerchant)
	mux.HandleFunc("GET /api/v1/merchants/{id}", portfolioHandler.GetMerchant)
	mux.HandleFunc("PUT /api/v1/merchants/{id}", portfolioHandler.UpdateMerchant)
	mux.HandleFunc("DELETE /api/v1/merchants/{id}", portfolioHandler.DeleteMerchant)
	mux.HandleFunc("GET /api/v1/merchants", portfolioHandler.ListMerchants)
	mux.HandleFunc("POST /api/v1/merchants/search", portfolioHandler.SearchMerchants)
	mux.HandleFunc("GET /api/v1/merchants/{id}/dashboard", portfolioHandler.GetMerchantDashboard)

	// Session management routes
	mux.HandleFunc("POST /api/v1/merchants/{id}/session", portfolioHandler.StartMerchantSession)
	mux.HandleFunc("GET /api/v1/merchants/session/active", portfolioHandler.GetActiveMerchantSession)
	mux.HandleFunc("DELETE /api/v1/merchants/session/active", portfolioHandler.EndMerchantSession)

	// Bulk operations routes
	mux.HandleFunc("POST /api/v1/merchants/bulk/portfolio-type", portfolioHandler.BulkUpdatePortfolioType)
	mux.HandleFunc("POST /api/v1/merchants/bulk/risk-level", portfolioHandler.BulkUpdateRiskLevel)

	// Risk assessment routes
	mux.HandleFunc("GET /api/v1/merchants/risk/dashboard", portfolioHandler.GetRiskDashboard)

	// Compliance routes
	mux.HandleFunc("GET /api/v1/merchants/compliance/dashboard", portfolioHandler.GetComplianceDashboard)
	mux.HandleFunc("GET /api/v1/merchants/compliance/alerts", portfolioHandler.GetComplianceAlerts)
	mux.HandleFunc("POST /api/v1/merchants/compliance/reports", portfolioHandler.GenerateComplianceReport)

	// Analytics routes
	mux.HandleFunc("GET /api/v1/merchants/analytics/dashboard", portfolioHandler.GetAnalyticsDashboard)

	// Comparison routes
	mux.HandleFunc("POST /api/v1/merchants/compare", comparisonHandler.CompareMerchants)
	mux.HandleFunc("GET /api/v1/merchants/compare/{id}", comparisonHandler.GetComparison)
	mux.HandleFunc("POST /api/v1/merchants/compare/report", comparisonHandler.GenerateReport)
	mux.HandleFunc("GET /api/v1/merchants/compare/reports/{id}/download", comparisonHandler.DownloadReport)

	// Export routes
	mux.HandleFunc("POST /api/v1/merchants/export", portfolioHandler.ExportPortfolio)

	return mux
}
