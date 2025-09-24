package mocks

import (
	"context"
	"fmt"
	"time"

	"kyb-platform/internal/services"
)

// MockComparisonService provides a mock implementation for E2E tests
type MockComparisonService struct {
	comparisonResults map[string]*services.ComparisonResult
	reportResults     map[string]*services.ComparisonReport
	errors            map[string]error
}

// NewMockComparisonService creates a new mock comparison service
func NewMockComparisonService() *MockComparisonService {
	return &MockComparisonService{
		comparisonResults: make(map[string]*services.ComparisonResult),
		reportResults:     make(map[string]*services.ComparisonReport),
		errors:            make(map[string]error),
	}
}

// CompareMerchants compares two merchants
func (m *MockComparisonService) CompareMerchants(ctx context.Context, merchant1ID, merchant2ID string, comparisonType string) (*services.ComparisonResult, error) {
	if err, exists := m.errors["compare"]; exists {
		return nil, err
	}

	// Check if we have a pre-configured result
	for _, result := range m.comparisonResults {
		if (result.Merchant1ID == merchant1ID && result.Merchant2ID == merchant2ID) ||
			(result.Merchant1ID == merchant2ID && result.Merchant2ID == merchant1ID) {
			return result, nil
		}
	}

	// Generate a default comparison result
	comparisonID := fmt.Sprintf("comparison_%d", time.Now().UnixNano())
	result := &services.ComparisonResult{
		ID:               comparisonID,
		Merchant1ID:      merchant1ID,
		Merchant2ID:      merchant2ID,
		ComparisonType:   comparisonType,
		CreatedAt:        time.Now(),
		Similarities:     []string{"Both are LLCs", "Similar business structure"},
		Differences:      []string{"Different risk levels", "Different portfolio types"},
		Recommendations:  []string{"Review risk assessment", "Consider portfolio alignment"},
		OverallScore:     0.75,
		DetailedAnalysis: "Both merchants show strong business fundamentals with different risk profiles.",
	}

	m.comparisonResults[comparisonID] = result
	return result, nil
}

// GetComparison retrieves a comparison by ID
func (m *MockComparisonService) GetComparison(ctx context.Context, comparisonID string) (*services.ComparisonResult, error) {
	if err, exists := m.errors["get_comparison"]; exists {
		return nil, err
	}

	if result, exists := m.comparisonResults[comparisonID]; exists {
		return result, nil
	}

	return nil, fmt.Errorf("comparison not found")
}

// GenerateReport generates a comparison report
func (m *MockComparisonService) GenerateReport(ctx context.Context, merchant1ID, merchant2ID, reportType string, includeCharts bool) (*services.ComparisonReport, error) {
	if err, exists := m.errors["generate_report"]; exists {
		return nil, err
	}

	reportID := fmt.Sprintf("report_%d", time.Now().UnixNano())
	report := &services.ComparisonReport{
		ID:            reportID,
		Merchant1ID:   merchant1ID,
		Merchant2ID:   merchant2ID,
		ReportType:    reportType,
		Status:        "generated",
		DownloadURL:   fmt.Sprintf("/api/v1/merchants/compare/reports/%s/download", reportID),
		FileSize:      1024 * 1024, // 1MB
		CreatedAt:     time.Now(),
		IncludeCharts: includeCharts,
	}

	m.reportResults[reportID] = report
	return report, nil
}

// DownloadReport downloads a comparison report
func (m *MockComparisonService) DownloadReport(ctx context.Context, reportID string) (*services.ComparisonReport, error) {
	if err, exists := m.errors["download_report"]; exists {
		return nil, err
	}

	if report, exists := m.reportResults[reportID]; exists {
		return report, nil
	}

	return nil, fmt.Errorf("report not found")
}

// Helper methods for testing

// SetMockComparisonResult sets a mock comparison result
func (m *MockComparisonService) SetMockComparisonResult(result *services.ComparisonResult) {
	m.comparisonResults[result.ID] = result
}

// SetError sets an error for a specific operation
func (m *MockComparisonService) SetError(operation string, err error) {
	m.errors[operation] = err
}

// ClearErrors clears all errors
func (m *MockComparisonService) ClearErrors() {
	m.errors = make(map[string]error)
}

// GetComparisonCount returns the number of comparisons
func (m *MockComparisonService) GetComparisonCount() int {
	return len(m.comparisonResults)
}

// GetReportCount returns the number of reports
func (m *MockComparisonService) GetReportCount() int {
	return len(m.reportResults)
}
