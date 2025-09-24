package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReportingFeatures tests all reporting features
func (suite *IntegrationTestingSuite) TestReportingFeatures(t *testing.T) {
	t.Run("PerformanceReports", suite.testPerformanceReports)
	t.Run("ComplianceReports", suite.testComplianceReports)
	t.Run("RiskReports", suite.testRiskReports)
	t.Run("CustomReports", suite.testCustomReports)
	t.Run("ReportScheduling", suite.testReportScheduling)
	t.Run("ReportExport", suite.testReportExport)
}

// testPerformanceReports tests performance reporting functionality
func (suite *IntegrationTestingSuite) testPerformanceReports(t *testing.T) {
	tests := []struct {
		name           string
		request        PerformanceReportRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful_performance_report",
			request: PerformanceReportRequest{
				StartDate: time.Now().AddDate(0, 0, -30).Format("2006-01-02"),
				EndDate:   time.Now().Format("2006-01-02"),
				Metrics:   []string{"response_time", "throughput", "error_rate"},
				Format:    "json",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "performance_report_with_filters",
			request: PerformanceReportRequest{
				StartDate: time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
				EndDate:   time.Now().Format("2006-01-02"),
				Metrics:   []string{"response_time", "throughput"},
				Filters: map[string]interface{}{
					"service":  "classification",
					"endpoint": "/v1/classify",
				},
				Format: "json",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "performance_report_csv_export",
			request: PerformanceReportRequest{
				StartDate: time.Now().AddDate(0, 0, -30).Format("2006-01-02"),
				EndDate:   time.Now().Format("2006-01-02"),
				Metrics:   []string{"response_time", "throughput", "error_rate"},
				Format:    "csv",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid_date_range",
			request: PerformanceReportRequest{
				StartDate: time.Now().Format("2006-01-02"),
				EndDate:   time.Now().AddDate(0, 0, -30).Format("2006-01-02"), // End before start
				Metrics:   []string{"response_time"},
				Format:    "json",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid date range",
		},
		{
			name: "unsupported_report_format",
			request: PerformanceReportRequest{
				StartDate: time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
				EndDate:   time.Now().Format("2006-01-02"),
				Metrics:   []string{"response_time"},
				Format:    "xml", // Unsupported format
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "unsupported format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req, err := http.NewRequest("GET", suite.server.URL+"/v1/reports/performance",
				bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test_token")

			// Execute request
			client := &http.Client{Timeout: 60 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Verify response
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedError != "" {
				var errorResp map[string]string
				err = json.NewDecoder(resp.Body).Decode(&errorResp)
				require.NoError(t, err)
				assert.Contains(t, errorResp["error"], tt.expectedError)
			} else if resp.StatusCode == http.StatusOK {
				if tt.request.Format == "json" {
					var result PerformanceReportResponse
					err = json.NewDecoder(resp.Body).Decode(&result)
					require.NoError(t, err)
					assert.NotEmpty(t, result.ReportID)
					assert.NotEmpty(t, result.GeneratedAt)
					assert.NotEmpty(t, result.Metrics)
					assert.Equal(t, tt.request.StartDate, result.StartDate)
					assert.Equal(t, tt.request.EndDate, result.EndDate)
				} else if tt.request.Format == "csv" {
					// Verify CSV content type
					assert.Equal(t, "text/csv", resp.Header.Get("Content-Type"))
					assert.NotEmpty(t, resp.Header.Get("Content-Disposition"))
				}
			}
		})
	}
}

// testComplianceReports tests compliance reporting functionality
func (suite *IntegrationTestingSuite) testComplianceReports(t *testing.T) {
	tests := []struct {
		name           string
		request        ComplianceReportRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful_compliance_report",
			request: ComplianceReportRequest{
				StartDate:       time.Now().AddDate(0, 0, -30).Format("2006-01-02"),
				EndDate:         time.Now().Format("2006-01-02"),
				ComplianceTypes: []string{"kyc", "aml", "sanctions"},
				Format:          "json",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "compliance_report_with_business_filter",
			request: ComplianceReportRequest{
				StartDate:       time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
				EndDate:         time.Now().Format("2006-01-02"),
				ComplianceTypes: []string{"kyc"},
				Filters: map[string]interface{}{
					"business_id": "biz_123",
					"status":      "approved",
				},
				Format: "json",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "compliance_report_pdf_export",
			request: ComplianceReportRequest{
				StartDate:       time.Now().AddDate(0, 0, -30).Format("2006-01-02"),
				EndDate:         time.Now().Format("2006-01-02"),
				ComplianceTypes: []string{"kyc", "aml"},
				Format:          "pdf",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid_compliance_type",
			request: ComplianceReportRequest{
				StartDate:       time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
				EndDate:         time.Now().Format("2006-01-02"),
				ComplianceTypes: []string{"invalid_type"},
				Format:          "json",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid compliance type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req, err := http.NewRequest("GET", suite.server.URL+"/v1/reports/compliance",
				bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test_token")

			// Execute request
			client := &http.Client{Timeout: 60 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Verify response
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedError != "" {
				var errorResp map[string]string
				err = json.NewDecoder(resp.Body).Decode(&errorResp)
				require.NoError(t, err)
				assert.Contains(t, errorResp["error"], tt.expectedError)
			} else if resp.StatusCode == http.StatusOK {
				if tt.request.Format == "json" {
					var result ComplianceReportResponse
					err = json.NewDecoder(resp.Body).Decode(&result)
					require.NoError(t, err)
					assert.NotEmpty(t, result.ReportID)
					assert.NotEmpty(t, result.GeneratedAt)
					assert.NotEmpty(t, result.ComplianceData)
					assert.Equal(t, tt.request.StartDate, result.StartDate)
					assert.Equal(t, tt.request.EndDate, result.EndDate)
				} else if tt.request.Format == "pdf" {
					// Verify PDF content type
					assert.Equal(t, "application/pdf", resp.Header.Get("Content-Type"))
					assert.NotEmpty(t, resp.Header.Get("Content-Disposition"))
				}
			}
		})
	}
}

// testRiskReports tests risk reporting functionality
func (suite *IntegrationTestingSuite) testRiskReports(t *testing.T) {
	tests := []struct {
		name           string
		request        RiskReportRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful_risk_report",
			request: RiskReportRequest{
				StartDate: time.Now().AddDate(0, 0, -30).Format("2006-01-02"),
				EndDate:   time.Now().Format("2006-01-02"),
				RiskTypes: []string{"business_risk", "compliance_risk", "operational_risk"},
				Format:    "json",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "risk_report_with_severity_filter",
			request: RiskReportRequest{
				StartDate: time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
				EndDate:   time.Now().Format("2006-01-02"),
				RiskTypes: []string{"business_risk"},
				Filters: map[string]interface{}{
					"severity": "high",
					"status":   "active",
				},
				Format: "json",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "risk_report_with_trend_analysis",
			request: RiskReportRequest{
				StartDate:     time.Now().AddDate(0, 0, -90).Format("2006-01-02"),
				EndDate:       time.Now().Format("2006-01-02"),
				RiskTypes:     []string{"business_risk", "compliance_risk"},
				IncludeTrends: true,
				Format:        "json",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "risk_report_excel_export",
			request: RiskReportRequest{
				StartDate: time.Now().AddDate(0, 0, -30).Format("2006-01-02"),
				EndDate:   time.Now().Format("2006-01-02"),
				RiskTypes: []string{"business_risk"},
				Format:    "xlsx",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid_risk_type",
			request: RiskReportRequest{
				StartDate: time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
				EndDate:   time.Now().Format("2006-01-02"),
				RiskTypes: []string{"invalid_risk_type"},
				Format:    "json",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid risk type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req, err := http.NewRequest("GET", suite.server.URL+"/v1/reports/risk",
				bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test_token")

			// Execute request
			client := &http.Client{Timeout: 60 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Verify response
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedError != "" {
				var errorResp map[string]string
				err = json.NewDecoder(resp.Body).Decode(&errorResp)
				require.NoError(t, err)
				assert.Contains(t, errorResp["error"], tt.expectedError)
			} else if resp.StatusCode == http.StatusOK {
				if tt.request.Format == "json" {
					var result RiskReportResponse
					err = json.NewDecoder(resp.Body).Decode(&result)
					require.NoError(t, err)
					assert.NotEmpty(t, result.ReportID)
					assert.NotEmpty(t, result.GeneratedAt)
					assert.NotEmpty(t, result.RiskData)
					assert.Equal(t, tt.request.StartDate, result.StartDate)
					assert.Equal(t, tt.request.EndDate, result.EndDate)

					if tt.request.IncludeTrends {
						assert.NotEmpty(t, result.Trends)
					}
				} else if tt.request.Format == "xlsx" {
					// Verify Excel content type
					assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
						resp.Header.Get("Content-Type"))
					assert.NotEmpty(t, resp.Header.Get("Content-Disposition"))
				}
			}
		})
	}
}

// testCustomReports tests custom reporting functionality
func (suite *IntegrationTestingSuite) testCustomReports(t *testing.T) {
	tests := []struct {
		name           string
		request        CustomReportRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful_custom_report",
			request: CustomReportRequest{
				Name:        "Custom Business Analysis",
				Description: "Custom report for business analysis",
				Query:       "SELECT * FROM businesses WHERE created_at >= $1 AND created_at <= $2",
				Parameters: map[string]interface{}{
					"start_date": time.Now().AddDate(0, 0, -30).Format("2006-01-02"),
					"end_date":   time.Now().Format("2006-01-02"),
				},
				Format: "json",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "custom_report_with_aggregation",
			request: CustomReportRequest{
				Name:        "Business Count by Industry",
				Description: "Count of businesses by industry",
				Query:       "SELECT industry, COUNT(*) as count FROM businesses GROUP BY industry",
				Parameters:  map[string]interface{}{},
				Format:      "json",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid_sql_query",
			request: CustomReportRequest{
				Name:        "Invalid Query Report",
				Description: "Report with invalid SQL",
				Query:       "INVALID SQL QUERY",
				Parameters:  map[string]interface{}{},
				Format:      "json",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid SQL query",
		},
		{
			name: "custom_report_with_restricted_query",
			request: CustomReportRequest{
				Name:        "Restricted Query Report",
				Description: "Report with restricted query",
				Query:       "DROP TABLE businesses", // Dangerous query
				Parameters:  map[string]interface{}{},
				Format:      "json",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "query not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", suite.server.URL+"/v1/reports/custom",
				bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test_token")

			// Execute request
			client := &http.Client{Timeout: 60 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Verify response
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedError != "" {
				var errorResp map[string]string
				err = json.NewDecoder(resp.Body).Decode(&errorResp)
				require.NoError(t, err)
				assert.Contains(t, errorResp["error"], tt.expectedError)
			} else if resp.StatusCode == http.StatusOK {
				var result CustomReportResponse
				err = json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
				assert.NotEmpty(t, result.ReportID)
				assert.NotEmpty(t, result.GeneratedAt)
				assert.NotEmpty(t, result.Data)
				assert.Equal(t, tt.request.Name, result.Name)
			}
		})
	}
}

// testReportScheduling tests report scheduling functionality
func (suite *IntegrationTestingSuite) testReportScheduling(t *testing.T) {
	tests := []struct {
		name           string
		request        ReportScheduleRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful_daily_report_schedule",
			request: ReportScheduleRequest{
				ReportType: "performance",
				Schedule:   "daily",
				Time:       "09:00",
				Recipients: []string{"admin@example.com"},
				Format:     "json",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "weekly_report_schedule",
			request: ReportScheduleRequest{
				ReportType: "compliance",
				Schedule:   "weekly",
				DayOfWeek:  "monday",
				Time:       "08:00",
				Recipients: []string{"compliance@example.com"},
				Format:     "pdf",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "monthly_report_schedule",
			request: ReportScheduleRequest{
				ReportType: "risk",
				Schedule:   "monthly",
				DayOfMonth: 1,
				Time:       "07:00",
				Recipients: []string{"risk@example.com"},
				Format:     "xlsx",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid_schedule_frequency",
			request: ReportScheduleRequest{
				ReportType: "performance",
				Schedule:   "invalid",
				Time:       "09:00",
				Recipients: []string{"admin@example.com"},
				Format:     "json",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid schedule frequency",
		},
		{
			name: "invalid_time_format",
			request: ReportScheduleRequest{
				ReportType: "performance",
				Schedule:   "daily",
				Time:       "25:00", // Invalid time
				Recipients: []string{"admin@example.com"},
				Format:     "json",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid time format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", suite.server.URL+"/v1/reports/schedule",
				bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test_token")

			// Execute request
			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Verify response
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedError != "" {
				var errorResp map[string]string
				err = json.NewDecoder(resp.Body).Decode(&errorResp)
				require.NoError(t, err)
				assert.Contains(t, errorResp["error"], tt.expectedError)
			} else if resp.StatusCode == http.StatusCreated {
				var result map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
				assert.NotEmpty(t, result["schedule_id"])
				assert.Equal(t, tt.request.ReportType, result["report_type"])
				assert.Equal(t, tt.request.Schedule, result["schedule"])
			}
		})
	}
}

// testReportExport tests report export functionality
func (suite *IntegrationTestingSuite) testReportExport(t *testing.T) {
	// First create a report
	reportRequest := PerformanceReportRequest{
		StartDate: time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
		EndDate:   time.Now().Format("2006-01-02"),
		Metrics:   []string{"response_time", "throughput"},
		Format:    "json",
	}

	reportID := suite.createTestReport(t, reportRequest)

	tests := []struct {
		name           string
		exportFormat   string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "export_to_csv",
			exportFormat:   "csv",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "export_to_pdf",
			exportFormat:   "pdf",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "export_to_excel",
			exportFormat:   "xlsx",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unsupported_export_format",
			exportFormat:   "xml",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "unsupported export format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			req, err := http.NewRequest("GET",
				fmt.Sprintf("%s/v1/reports/%s/export?format=%s", suite.server.URL, reportID, tt.exportFormat),
				nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer test_token")

			// Execute request
			client := &http.Client{Timeout: 60 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Verify response
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedError != "" {
				var errorResp map[string]string
				err = json.NewDecoder(resp.Body).Decode(&errorResp)
				require.NoError(t, err)
				assert.Contains(t, errorResp["error"], tt.expectedError)
			} else if resp.StatusCode == http.StatusOK {
				// Verify content type and disposition
				switch tt.exportFormat {
				case "csv":
					assert.Equal(t, "text/csv", resp.Header.Get("Content-Type"))
				case "pdf":
					assert.Equal(t, "application/pdf", resp.Header.Get("Content-Type"))
				case "xlsx":
					assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
						resp.Header.Get("Content-Type"))
				}
				assert.NotEmpty(t, resp.Header.Get("Content-Disposition"))
			}
		})
	}
}

// Helper method to create a test report
func (suite *IntegrationTestingSuite) createTestReport(t *testing.T, request PerformanceReportRequest) string {
	reqBody, err := json.Marshal(request)
	require.NoError(t, err)

	req, err := http.NewRequest("GET", suite.server.URL+"/v1/reports/performance",
		bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test_token")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result PerformanceReportResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	return result.ReportID
}

// Request/Response types for reporting testing
type PerformanceReportRequest struct {
	StartDate string                 `json:"start_date"`
	EndDate   string                 `json:"end_date"`
	Metrics   []string               `json:"metrics"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
	Format    string                 `json:"format"`
}

type PerformanceReportResponse struct {
	ReportID    string                 `json:"report_id"`
	GeneratedAt string                 `json:"generated_at"`
	StartDate   string                 `json:"start_date"`
	EndDate     string                 `json:"end_date"`
	Metrics     map[string]interface{} `json:"metrics"`
}

type ComplianceReportRequest struct {
	StartDate       string                 `json:"start_date"`
	EndDate         string                 `json:"end_date"`
	ComplianceTypes []string               `json:"compliance_types"`
	Filters         map[string]interface{} `json:"filters,omitempty"`
	Format          string                 `json:"format"`
}

type ComplianceReportResponse struct {
	ReportID       string                 `json:"report_id"`
	GeneratedAt    string                 `json:"generated_at"`
	StartDate      string                 `json:"start_date"`
	EndDate        string                 `json:"end_date"`
	ComplianceData map[string]interface{} `json:"compliance_data"`
}

type RiskReportRequest struct {
	StartDate     string                 `json:"start_date"`
	EndDate       string                 `json:"end_date"`
	RiskTypes     []string               `json:"risk_types"`
	Filters       map[string]interface{} `json:"filters,omitempty"`
	IncludeTrends bool                   `json:"include_trends,omitempty"`
	Format        string                 `json:"format"`
}

type RiskReportResponse struct {
	ReportID    string                 `json:"report_id"`
	GeneratedAt string                 `json:"generated_at"`
	StartDate   string                 `json:"start_date"`
	EndDate     string                 `json:"end_date"`
	RiskData    map[string]interface{} `json:"risk_data"`
	Trends      map[string]interface{} `json:"trends,omitempty"`
}

type CustomReportRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Query       string                 `json:"query"`
	Parameters  map[string]interface{} `json:"parameters"`
	Format      string                 `json:"format"`
}

type CustomReportResponse struct {
	ReportID    string                   `json:"report_id"`
	GeneratedAt string                   `json:"generated_at"`
	Name        string                   `json:"name"`
	Data        []map[string]interface{} `json:"data"`
}

type ReportScheduleRequest struct {
	ReportType string   `json:"report_type"`
	Schedule   string   `json:"schedule"`
	Time       string   `json:"time"`
	DayOfWeek  string   `json:"day_of_week,omitempty"`
	DayOfMonth int      `json:"day_of_month,omitempty"`
	Recipients []string `json:"recipients"`
	Format     string   `json:"format"`
}
