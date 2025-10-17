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

// TestMerchantComparisonE2E tests merchant comparison functionality end-to-end
func TestMerchantComparisonE2E(t *testing.T) {
	// Skip if not running E2E tests
	if os.Getenv("E2E_TESTS") != "true" {
		t.Skip("Skipping E2E tests - set E2E_TESTS=true to run")
	}

	// Setup test environment
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create mock services
	mockService := mocks.NewMockMerchantPortfolioService()
	comparisonService := mocks.NewMockComparisonService()
	handler := handlers.NewMerchantPortfolioHandler(mockService, nil)
	comparisonHandler := handlers.NewComparisonHandler(comparisonService, nil)

	// Create test server
	server := httptest.NewServer(createComparisonTestRouter(handler, comparisonHandler))
	defer server.Close()

	// Test 1: Basic merchant comparison
	t.Run("BasicMerchantComparison", func(t *testing.T) {
		testBasicMerchantComparison(t, server, mockService, comparisonService, ctx)
	})

	// Test 2: Comparison with detailed analysis
	t.Run("ComparisonWithDetailedAnalysis", func(t *testing.T) {
		testComparisonWithDetailedAnalysis(t, server, mockService, comparisonService, ctx)
	})

	// Test 3: Comparison report generation
	t.Run("ComparisonReportGeneration", func(t *testing.T) {
		testComparisonReportGeneration(t, server, mockService, comparisonService, ctx)
	})

	// Test 4: Comparison error handling
	t.Run("ComparisonErrorHandling", func(t *testing.T) {
		testComparisonErrorHandling(t, server, mockService, comparisonService, ctx)
	})
}

// testBasicMerchantComparison tests basic merchant comparison functionality
func testBasicMerchantComparison(t *testing.T, server *httptest.Server, mockService *mocks.MockMerchantPortfolioService, comparisonService *mocks.MockComparisonService, ctx context.Context) {
	// Create test merchants for comparison
	merchant1 := createComparisonTestMerchant("comparison_test_1", "Tech Company A", services.PortfolioTypeOnboarded, services.RiskLevelLow)
	merchant2 := createComparisonTestMerchant("comparison_test_2", "Retail Company B", services.PortfolioTypeProspective, services.RiskLevelMedium)

	mockService.AddMerchant(merchant1)
	mockService.AddMerchant(merchant2)

	// Set up comparison service mock
	comparisonResult := &services.ComparisonResult{
		ID:               "comparison_123",
		Merchant1ID:      merchant1.ID,
		Merchant2ID:      merchant2.ID,
		ComparisonType:   "basic",
		CreatedAt:        time.Now(),
		Similarities:     []string{"Both are LLCs", "Similar employee count"},
		Differences:      []string{"Different industries", "Different risk levels"},
		Recommendations:  []string{"Consider risk assessment", "Review industry compliance"},
		OverallScore:     0.75,
		DetailedAnalysis: "Both merchants show strong business fundamentals with different risk profiles.",
	}

	comparisonService.SetMockComparisonResult(comparisonResult)

	// Test basic comparison
	t.Log("Testing basic merchant comparison...")
	comparisonReq := &handlers.ComparisonRequest{
		Merchant1ID:    merchant1.ID,
		Merchant2ID:    merchant2.ID,
		ComparisonType: "basic",
	}

	comparisonResp, err := makeRequest(t, server, "POST", "/api/v1/merchants/compare", comparisonReq)
	if err != nil {
		t.Fatalf("Failed to perform merchant comparison: %v", err)
	}

	if comparisonResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", comparisonResp.StatusCode, comparisonResp.Body)
	}

	var comparison handlers.ComparisonResponse
	if err := json.Unmarshal(comparisonResp.Body, &comparison); err != nil {
		t.Fatalf("Failed to unmarshal comparison result: %v", err)
	}

	// Verify comparison result
	if comparison.ID == "" {
		t.Fatal("Expected comparison ID to be generated")
	}

	if comparison.Merchant1ID != merchant1.ID {
		t.Errorf("Expected merchant1 ID %s, got %s", merchant1.ID, comparison.Merchant1ID)
	}

	if comparison.Merchant2ID != merchant2.ID {
		t.Errorf("Expected merchant2 ID %s, got %s", merchant2.ID, comparison.Merchant2ID)
	}

	if len(comparison.Similarities) == 0 {
		t.Error("Expected similarities to be present")
	}

	if len(comparison.Differences) == 0 {
		t.Error("Expected differences to be present")
	}

	if len(comparison.Recommendations) == 0 {
		t.Error("Expected recommendations to be present")
	}

	if comparison.OverallScore < 0.0 || comparison.OverallScore > 1.0 {
		t.Errorf("Expected overall score between 0.0 and 1.0, got %f", comparison.OverallScore)
	}

	t.Logf("✅ Basic merchant comparison completed successfully: Score %.2f", comparison.OverallScore)
}

// testComparisonWithDetailedAnalysis tests comparison with detailed analysis
func testComparisonWithDetailedAnalysis(t *testing.T, server *httptest.Server, mockService *mocks.MockMerchantPortfolioService, comparisonService *mocks.MockComparisonService, ctx context.Context) {
	// Create test merchants with different characteristics
	merchant1 := createComparisonTestMerchant("detailed_test_1", "High-Tech Startup", services.PortfolioTypeOnboarded, services.RiskLevelLow)
	merchant1.Industry = "Software Development"
	merchant1.EmployeeCount = 25
	merchant1.AnnualRevenue = floatPtr(5000000.0)

	merchant2 := createComparisonTestMerchant("detailed_test_2", "Traditional Manufacturing", services.PortfolioTypeProspective, services.RiskLevelHigh)
	merchant2.Industry = "Manufacturing"
	merchant2.EmployeeCount = 200
	merchant2.AnnualRevenue = floatPtr(25000000.0)

	mockService.AddMerchant(merchant1)
	mockService.AddMerchant(merchant2)

	// Set up detailed comparison result
	detailedResult := &services.ComparisonResult{
		ID:               "detailed_comparison_123",
		Merchant1ID:      merchant1.ID,
		Merchant2ID:      merchant2.ID,
		ComparisonType:   "detailed",
		CreatedAt:        time.Now(),
		Similarities:     []string{"Both are established businesses", "Both have revenue streams"},
		Differences:      []string{"Different industries", "Different company sizes", "Different risk profiles"},
		Recommendations:  []string{"Consider industry-specific compliance", "Review size-based regulations", "Assess risk mitigation strategies"},
		OverallScore:     0.45,
		DetailedAnalysis: "These merchants operate in different industries with varying risk profiles. The tech startup shows high growth potential but higher volatility, while the manufacturing company shows stability but lower growth prospects.",
		FinancialComparison: &services.FinancialComparison{
			RevenueRatio:    0.2,   // merchant1 revenue / merchant2 revenue
			EmployeeRatio:   0.125, // merchant1 employees / merchant2 employees
			GrowthPotential: "merchant1_higher",
			StabilityScore:  "merchant2_higher",
		},
		RiskComparison: &services.RiskComparison{
			OverallRisk:    "merchant2_higher",
			IndustryRisk:   "merchant2_higher",
			SizeRisk:       "merchant1_higher",
			ComplianceRisk: "merchant2_higher",
		},
	}

	comparisonService.SetMockComparisonResult(detailedResult)

	// Test detailed comparison
	t.Log("Testing detailed merchant comparison...")
	comparisonReq := &handlers.ComparisonRequest{
		Merchant1ID:    merchant1.ID,
		Merchant2ID:    merchant2.ID,
		ComparisonType: "detailed",
	}

	comparisonResp, err := makeRequest(t, server, "POST", "/api/v1/merchants/compare", comparisonReq)
	if err != nil {
		t.Fatalf("Failed to perform detailed merchant comparison: %v", err)
	}

	if comparisonResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", comparisonResp.StatusCode, comparisonResp.Body)
	}

	var comparison handlers.ComparisonResponse
	if err := json.Unmarshal(comparisonResp.Body, &comparison); err != nil {
		t.Fatalf("Failed to unmarshal detailed comparison result: %v", err)
	}

	// Verify detailed comparison result
	if comparison.ID == "" {
		t.Fatal("Expected comparison ID to be generated")
	}

	if comparison.DetailedAnalysis == "" {
		t.Error("Expected detailed analysis to be present")
	}

	if comparison.FinancialComparison == nil {
		t.Error("Expected financial comparison to be present")
	} else {
		if comparison.FinancialComparison.RevenueRatio <= 0 {
			t.Error("Expected revenue ratio to be positive")
		}
		if comparison.FinancialComparison.EmployeeRatio <= 0 {
			t.Error("Expected employee ratio to be positive")
		}
	}

	if comparison.RiskComparison == nil {
		t.Error("Expected risk comparison to be present")
	} else {
		if comparison.RiskComparison.OverallRisk == "" {
			t.Error("Expected overall risk assessment to be present")
		}
	}

	t.Logf("✅ Detailed merchant comparison completed successfully: Score %.2f", comparison.OverallScore)
}

// testComparisonReportGeneration tests comparison report generation
func testComparisonReportGeneration(t *testing.T, server *httptest.Server, mockService *mocks.MockMerchantPortfolioService, comparisonService *mocks.MockComparisonService, ctx context.Context) {
	// Create test merchants
	merchant1 := createComparisonTestMerchant("report_test_1", "Report Company A", services.PortfolioTypeOnboarded, services.RiskLevelLow)
	merchant2 := createComparisonTestMerchant("report_test_2", "Report Company B", services.PortfolioTypeProspective, services.RiskLevelMedium)

	mockService.AddMerchant(merchant1)
	mockService.AddMerchant(merchant2)

	// Set up comparison result for report generation
	comparisonResult := &services.ComparisonResult{
		ID:               "report_comparison_123",
		Merchant1ID:      merchant1.ID,
		Merchant2ID:      merchant2.ID,
		ComparisonType:   "report",
		CreatedAt:        time.Now(),
		Similarities:     []string{"Both are LLCs", "Similar business structure"},
		Differences:      []string{"Different risk levels", "Different portfolio types"},
		Recommendations:  []string{"Review risk assessment", "Consider portfolio alignment"},
		OverallScore:     0.65,
		DetailedAnalysis: "Comprehensive comparison analysis for report generation.",
	}

	comparisonService.SetMockComparisonResult(comparisonResult)

	// Test comparison report generation
	t.Log("Testing comparison report generation...")
	reportReq := &handlers.ComparisonReportRequest{
		Merchant1ID:   merchant1.ID,
		Merchant2ID:   merchant2.ID,
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
		t.Fatalf("Failed to unmarshal report response: %v", err)
	}

	// Verify report generation result
	if report.ReportID == "" {
		t.Fatal("Expected report ID to be generated")
	}

	if report.ReportType != "pdf" {
		t.Errorf("Expected report type 'pdf', got '%s'", report.ReportType)
	}

	if report.Status != "generated" {
		t.Errorf("Expected status 'generated', got '%s'", report.Status)
	}

	if report.DownloadURL == "" {
		t.Error("Expected download URL to be provided")
	}

	if report.FileSize <= 0 {
		t.Error("Expected file size to be positive")
	}

	t.Logf("✅ Comparison report generated successfully: %s (%d bytes)", report.ReportID, report.FileSize)

	// Test report download
	t.Log("Testing report download...")
	downloadResp, err := makeRequest(t, server, "GET", fmt.Sprintf("/api/v1/merchants/compare/reports/%s/download", report.ReportID), nil)
	if err != nil {
		t.Fatalf("Failed to download comparison report: %v", err)
	}

	if downloadResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", downloadResp.StatusCode, downloadResp.Body)
	}

	// Verify download response headers
	contentType := downloadResp.Header.Get("Content-Type")
	if contentType != "application/pdf" {
		t.Errorf("Expected content type 'application/pdf', got '%s'", contentType)
	}

	contentDisposition := downloadResp.Header.Get("Content-Disposition")
	if contentDisposition == "" {
		t.Error("Expected Content-Disposition header to be present")
	}

	t.Logf("✅ Comparison report download successful")
}

// testComparisonErrorHandling tests comparison error handling
func testComparisonErrorHandling(t *testing.T, server *httptest.Server, mockService *mocks.MockMerchantPortfolioService, comparisonService *mocks.MockComparisonService, ctx context.Context) {
	// Test 1: Compare non-existent merchants
	t.Log("Test 1: Comparing non-existent merchants...")
	comparisonReq := &handlers.ComparisonRequest{
		Merchant1ID:    "nonexistent_1",
		Merchant2ID:    "nonexistent_2",
		ComparisonType: "basic",
	}

	comparisonResp, err := makeRequest(t, server, "POST", "/api/v1/merchants/compare", comparisonReq)
	if err != nil {
		t.Fatalf("Failed to compare non-existent merchants: %v", err)
	}

	if comparisonResp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d: %s", comparisonResp.StatusCode, comparisonResp.Body)
	}

	t.Logf("✅ Comparison of non-existent merchants handled correctly")

	// Test 2: Compare same merchant
	t.Log("Test 2: Comparing same merchant...")
	// Create a test merchant
	testMerchant := createComparisonTestMerchant("same_merchant_test", "Same Merchant", services.PortfolioTypeOnboarded, services.RiskLevelLow)
	mockService.AddMerchant(testMerchant)

	sameMerchantReq := &handlers.ComparisonRequest{
		Merchant1ID:    testMerchant.ID,
		Merchant2ID:    testMerchant.ID,
		ComparisonType: "basic",
	}

	sameMerchantResp, err := makeRequest(t, server, "POST", "/api/v1/merchants/compare", sameMerchantReq)
	if err != nil {
		t.Fatalf("Failed to compare same merchant: %v", err)
	}

	if sameMerchantResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d: %s", sameMerchantResp.StatusCode, sameMerchantResp.Body)
	}

	t.Logf("✅ Comparison of same merchant handled correctly")

	// Test 3: Invalid comparison type
	t.Log("Test 3: Invalid comparison type...")
	invalidTypeReq := &handlers.ComparisonRequest{
		Merchant1ID:    testMerchant.ID,
		Merchant2ID:    testMerchant.ID,
		ComparisonType: "invalid_type",
	}

	invalidTypeResp, err := makeRequest(t, server, "POST", "/api/v1/merchants/compare", invalidTypeReq)
	if err != nil {
		t.Fatalf("Failed to compare with invalid type: %v", err)
	}

	if invalidTypeResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d: %s", invalidTypeResp.StatusCode, invalidTypeResp.Body)
	}

	t.Logf("✅ Invalid comparison type handled correctly")

	// Test 4: Get non-existent comparison
	t.Log("Test 4: Getting non-existent comparison...")
	getResp, err := makeRequest(t, server, "GET", "/api/v1/merchants/compare/nonexistent_comparison", nil)
	if err != nil {
		t.Fatalf("Failed to get non-existent comparison: %v", err)
	}

	if getResp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d: %s", getResp.StatusCode, getResp.Body)
	}

	t.Logf("✅ Get non-existent comparison handled correctly")

	// Test 5: Download non-existent report
	t.Log("Test 5: Downloading non-existent report...")
	downloadResp, err := makeRequest(t, server, "GET", "/api/v1/merchants/compare/reports/nonexistent_report/download", nil)
	if err != nil {
		t.Fatalf("Failed to download non-existent report: %v", err)
	}

	if downloadResp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d: %s", downloadResp.StatusCode, downloadResp.Body)
	}

	t.Logf("✅ Download non-existent report handled correctly")
}

// Helper functions for comparison testing

func createComparisonTestMerchant(id, name string, portfolioType services.PortfolioType, riskLevel services.RiskLevel) *services.Merchant {
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

func floatPtr(f float64) *float64 {
	return &f
}

func createComparisonTestRouter(portfolioHandler *handlers.MerchantPortfolioHandler, comparisonHandler *handlers.ComparisonHandler) http.Handler {
	mux := http.NewServeMux()

	// Merchant portfolio routes
	mux.HandleFunc("GET /api/v1/merchants/{id}", portfolioHandler.GetMerchant)

	// Comparison routes
	mux.HandleFunc("POST /api/v1/merchants/compare", comparisonHandler.CompareMerchants)
	mux.HandleFunc("GET /api/v1/merchants/compare/{id}", comparisonHandler.GetComparison)
	mux.HandleFunc("POST /api/v1/merchants/compare/report", comparisonHandler.GenerateReport)
	mux.HandleFunc("GET /api/v1/merchants/compare/reports/{id}/download", comparisonHandler.DownloadReport)

	return mux
}
