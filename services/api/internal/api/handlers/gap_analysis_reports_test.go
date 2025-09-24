package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGenerateReport(t *testing.T) {
	handler := NewGapAnalysisReportsHandler()

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		shouldSucceed  bool
	}{
		{
			name: "Generate executive summary report",
			requestBody: `{
				"report_type": "executive",
				"format": "pdf",
				"filters": {
					"framework": "SOC 2",
					"priority": "critical"
				}
			}`,
			expectedStatus: http.StatusOK,
			shouldSucceed:  true,
		},
		{
			name: "Generate detailed analysis report",
			requestBody: `{
				"report_type": "detailed",
				"format": "excel",
				"filters": {
					"status": "in-progress",
					"date_range": "last-30-days"
				},
				"recipients": ["admin@company.com", "manager@company.com"]
			}`,
			expectedStatus: http.StatusOK,
			shouldSucceed:  true,
		},
		{
			name: "Generate report with missing type",
			requestBody: `{
				"format": "pdf",
				"filters": {}
			}`,
			expectedStatus: http.StatusBadRequest,
			shouldSucceed:  false,
		},
		{
			name: "Generate report with invalid JSON",
			requestBody: `{
				"report_type": "executive",
				"format": "pdf"
			`, // Missing closing brace
			expectedStatus: http.StatusBadRequest,
			shouldSucceed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/v1/reports/generate", strings.NewReader(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.GenerateReport(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.shouldSucceed {
				var response map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatal("Failed to unmarshal response:", err)
				}

				// Verify response structure
				requiredFields := []string{"report", "message", "download_url", "preview_url", "generated_at"}
				for _, field := range requiredFields {
					if _, ok := response[field]; !ok {
						t.Errorf("Expected %s in response", field)
					}
				}

				// Verify report structure
				report, ok := response["report"].(map[string]interface{})
				if !ok {
					t.Error("Expected report in response")
				}

				requiredReportFields := []string{"id", "title", "type", "format", "status", "generated_at"}
				for _, field := range requiredReportFields {
					if _, ok := report[field]; !ok {
						t.Errorf("Expected %s in report", field)
					}
				}
			}
		})
	}
}

func TestGetReportTemplates(t *testing.T) {
	handler := NewGapAnalysisReportsHandler()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Get all templates",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  4, // Based on sample data
		},
		{
			name:           "Filter templates by type",
			queryParams:    "?type=executive",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Filter templates by format",
			queryParams:    "?format=pdf",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "Filter templates by type and format",
			queryParams:    "?type=executive&format=pdf",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v1/reports/templates"+tt.queryParams, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.GetReportTemplates(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatal("Failed to unmarshal response:", err)
			}

			templates, ok := response["templates"].([]interface{})
			if !ok {
				t.Fatal("Expected templates array in response")
			}

			if len(templates) != tt.expectedCount {
				t.Errorf("Expected %d templates, got %d", tt.expectedCount, len(templates))
			}

			// Verify template structure
			for i, template := range templates {
				templateMap, ok := template.(map[string]interface{})
				if !ok {
					t.Errorf("Expected template %d to be a map", i)
					continue
				}

				requiredFields := []string{"id", "name", "type", "description", "format", "template"}
				for _, field := range requiredFields {
					if _, ok := templateMap[field]; !ok {
						t.Errorf("Expected %s in template %d", field, i)
					}
				}
			}
		})
	}
}

func TestGetReportDetails(t *testing.T) {
	handler := NewGapAnalysisReportsHandler()

	tests := []struct {
		name           string
		reportID       string
		expectedStatus int
		shouldExist    bool
	}{
		{
			name:           "Get existing report details",
			reportID:       "report_001",
			expectedStatus: http.StatusOK,
			shouldExist:    true,
		},
		{
			name:           "Get non-existing report details",
			reportID:       "non_existing_report",
			expectedStatus: http.StatusNotFound,
			shouldExist:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v1/reports/"+tt.reportID, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.GetReportDetails(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.shouldExist {
				var response map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatal("Failed to unmarshal response:", err)
				}

				// Verify response structure
				requiredFields := []string{"report", "data", "analytics", "metadata"}
				for _, field := range requiredFields {
					if _, ok := response[field]; !ok {
						t.Errorf("Expected %s in response", field)
					}
				}

				// Verify report structure
				report, ok := response["report"].(map[string]interface{})
				if !ok {
					t.Error("Expected report in response")
				}

				if report["id"] != tt.reportID {
					t.Errorf("Expected report ID %s, got %v", tt.reportID, report["id"])
				}
			}
		})
	}
}

func TestDownloadReport(t *testing.T) {
	handler := NewGapAnalysisReportsHandler()

	tests := []struct {
		name           string
		reportID       string
		expectedStatus int
		shouldExist    bool
	}{
		{
			name:           "Download existing report",
			reportID:       "report_001",
			expectedStatus: http.StatusOK,
			shouldExist:    true,
		},
		{
			name:           "Download non-existing report",
			reportID:       "non_existing_report",
			expectedStatus: http.StatusNotFound,
			shouldExist:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v1/reports/"+tt.reportID+"/download", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.DownloadReport(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.shouldExist {
				// Verify content type header
				contentType := rr.Header().Get("Content-Type")
				if contentType == "" {
					t.Error("Expected Content-Type header")
				}

				// Verify content disposition header
				contentDisposition := rr.Header().Get("Content-Disposition")
				if contentDisposition == "" {
					t.Error("Expected Content-Disposition header")
				}

				// Verify response body is not empty
				if len(rr.Body.Bytes()) == 0 {
					t.Error("Expected non-empty response body")
				}
			}
		})
	}
}

func TestPreviewReport(t *testing.T) {
	handler := NewGapAnalysisReportsHandler()

	tests := []struct {
		name           string
		reportID       string
		expectedStatus int
		shouldExist    bool
	}{
		{
			name:           "Preview existing report",
			reportID:       "report_001",
			expectedStatus: http.StatusOK,
			shouldExist:    true,
		},
		{
			name:           "Preview non-existing report",
			reportID:       "non_existing_report",
			expectedStatus: http.StatusNotFound,
			shouldExist:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v1/reports/"+tt.reportID+"/preview", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.PreviewReport(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.shouldExist {
				var response map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatal("Failed to unmarshal response:", err)
				}

				// Verify response structure
				requiredFields := []string{"report", "preview_data", "preview_url"}
				for _, field := range requiredFields {
					if _, ok := response[field]; !ok {
						t.Errorf("Expected %s in response", field)
					}
				}

				// Verify preview data structure
				previewData, ok := response["preview_data"].(map[string]interface{})
				if !ok {
					t.Error("Expected preview_data to be a map")
				}

				requiredPreviewFields := []string{"summary", "charts", "tables"}
				for _, field := range requiredPreviewFields {
					if _, ok := previewData[field]; !ok {
						t.Errorf("Expected %s in preview_data", field)
					}
				}
			}
		})
	}
}

func TestScheduleReport(t *testing.T) {
	handler := NewGapAnalysisReportsHandler()

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		shouldSucceed  bool
	}{
		{
			name: "Schedule daily report",
			requestBody: `{
				"report_type": "executive",
				"format": "pdf",
				"filters": {
					"framework": "SOC 2"
				},
				"recipients": ["admin@company.com"],
				"schedule": {
					"frequency": "daily",
					"time": "09:00",
					"start_date": "2025-01-20T00:00:00Z"
				}
			}`,
			expectedStatus: http.StatusOK,
			shouldSucceed:  true,
		},
		{
			name: "Schedule weekly report",
			requestBody: `{
				"report_type": "detailed",
				"format": "excel",
				"filters": {},
				"schedule": {
					"frequency": "weekly",
					"time": "10:00",
					"days": ["monday"],
					"start_date": "2025-01-20T00:00:00Z"
				}
			}`,
			expectedStatus: http.StatusOK,
			shouldSucceed:  true,
		},
		{
			name: "Schedule report with missing frequency",
			requestBody: `{
				"report_type": "executive",
				"format": "pdf",
				"schedule": {
					"time": "09:00",
					"start_date": "2025-01-20T00:00:00Z"
				}
			}`,
			expectedStatus: http.StatusBadRequest,
			shouldSucceed:  false,
		},
		{
			name: "Schedule report with invalid JSON",
			requestBody: `{
				"report_type": "executive",
				"format": "pdf",
				"schedule": {
					"frequency": "daily"
				}
			`, // Missing closing brace
			expectedStatus: http.StatusBadRequest,
			shouldSucceed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/v1/reports/schedule", strings.NewReader(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.ScheduleReport(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.shouldSucceed {
				var response map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatal("Failed to unmarshal response:", err)
				}

				// Verify response structure
				requiredFields := []string{"scheduled_report", "message", "next_run"}
				for _, field := range requiredFields {
					if _, ok := response[field]; !ok {
						t.Errorf("Expected %s in response", field)
					}
				}

				// Verify scheduled report structure
				scheduledReport, ok := response["scheduled_report"].(map[string]interface{})
				if !ok {
					t.Error("Expected scheduled_report in response")
				}

				requiredScheduledFields := []string{"id", "report_type", "format", "schedule", "status", "created_at"}
				for _, field := range requiredScheduledFields {
					if _, ok := scheduledReport[field]; !ok {
						t.Errorf("Expected %s in scheduled_report", field)
					}
				}
			}
		})
	}
}

func TestGetReportMetrics(t *testing.T) {
	handler := NewGapAnalysisReportsHandler()

	req, err := http.NewRequest("GET", "/v1/reports/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.GetReportMetrics(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal("Failed to unmarshal response:", err)
	}

	// Verify response structure
	if _, ok := response["metrics"]; !ok {
		t.Error("Expected metrics in response")
	}

	if _, ok := response["generated_at"]; !ok {
		t.Error("Expected generated_at in response")
	}

	// Verify metrics structure
	metrics, ok := response["metrics"].(map[string]interface{})
	if !ok {
		t.Error("Expected metrics to be a map")
	}

	requiredFields := []string{"total_reports", "reports_today", "average_size", "most_popular_type", "success_rate", "last_generated"}
	for _, field := range requiredFields {
		if _, ok := metrics[field]; !ok {
			t.Errorf("Expected %s in metrics", field)
		}
	}
}

func TestGetRecentReports(t *testing.T) {
	handler := NewGapAnalysisReportsHandler()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Get recent reports with default limit",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  3, // Based on sample data
		},
		{
			name:           "Get recent reports with custom limit",
			queryParams:    "?limit=2",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "Filter recent reports by type",
			queryParams:    "?type=executive",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Filter recent reports by format",
			queryParams:    "?format=pdf",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Filter recent reports by type and format",
			queryParams:    "?type=executive&format=pdf",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v1/reports/recent"+tt.queryParams, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.GetRecentReports(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatal("Failed to unmarshal response:", err)
			}

			reports, ok := response["reports"].([]interface{})
			if !ok {
				t.Fatal("Expected reports array in response")
			}

			if len(reports) != tt.expectedCount {
				t.Errorf("Expected %d reports, got %d", tt.expectedCount, len(reports))
			}

			// Verify report structure
			for i, report := range reports {
				reportMap, ok := report.(map[string]interface{})
				if !ok {
					t.Errorf("Expected report %d to be a map", i)
					continue
				}

				requiredFields := []string{"id", "title", "type", "format", "status", "generated_at"}
				for _, field := range requiredFields {
					if _, ok := reportMap[field]; !ok {
						t.Errorf("Expected %s in report %d", field, i)
					}
				}
			}
		})
	}
}

func TestExportReportData(t *testing.T) {
	handler := NewGapAnalysisReportsHandler()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedFormat string
	}{
		{
			name:           "Export data as JSON",
			queryParams:    "?format=json",
			expectedStatus: http.StatusOK,
			expectedFormat: "application/json",
		},
		{
			name:           "Export data as CSV",
			queryParams:    "?format=csv",
			expectedStatus: http.StatusOK,
			expectedFormat: "text/csv",
		},
		{
			name:           "Export data as XML",
			queryParams:    "?format=xml",
			expectedStatus: http.StatusOK,
			expectedFormat: "application/xml",
		},
		{
			name:           "Export data with default format",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedFormat: "application/json",
		},
		{
			name:           "Export data with filters",
			queryParams:    "?format=json&type=executive&framework=SOC 2",
			expectedStatus: http.StatusOK,
			expectedFormat: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v1/reports/export"+tt.queryParams, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.ExportReportData(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Verify content type header
			contentType := rr.Header().Get("Content-Type")
			if contentType != tt.expectedFormat {
				t.Errorf("Expected Content-Type %s, got %s", tt.expectedFormat, contentType)
			}

			// Verify content disposition header
			contentDisposition := rr.Header().Get("Content-Disposition")
			if contentDisposition == "" {
				t.Error("Expected Content-Disposition header")
			}

			// Verify response body is not empty
			if len(rr.Body.Bytes()) == 0 {
				t.Error("Expected non-empty response body")
			}
		})
	}
}

func TestReportTemplateStructure(t *testing.T) {
	handler := NewGapAnalysisReportsHandler()
	templates := handler.reportData

	if len(templates) == 0 {
		t.Fatal("Expected at least one report template")
	}

	// Test the first template structure
	template := templates[0]

	// Verify required fields
	if template.ID == "" {
		t.Error("Expected non-empty ID")
	}

	if template.Name == "" {
		t.Error("Expected non-empty Name")
	}

	if template.Type == "" {
		t.Error("Expected non-empty Type")
	}

	if template.Description == "" {
		t.Error("Expected non-empty Description")
	}

	if template.Format == "" {
		t.Error("Expected non-empty Format")
	}

	if template.Template == nil {
		t.Error("Expected non-nil Template")
	}

	// Verify template type values
	validTypes := []string{"executive", "detailed", "progress", "compliance"}
	validType := false
	for _, t := range validTypes {
		if template.Type == t {
			validType = true
			break
		}
	}
	if !validType {
		t.Errorf("Expected valid template type, got %s", template.Type)
	}

	// Verify template format values
	validFormats := []string{"pdf", "excel", "html"}
	validFormat := false
	for _, f := range validFormats {
		if template.Format == f {
			validFormat = true
			break
		}
	}
	if !validFormat {
		t.Errorf("Expected valid template format, got %s", template.Format)
	}

	// Verify template structure
	if _, ok := template.Template["sections"]; !ok {
		t.Error("Expected sections in template")
	}

	if _, ok := template.Template["charts"]; !ok {
		t.Error("Expected charts in template")
	}

	if _, ok := template.Template["tables"]; !ok {
		t.Error("Expected tables in template")
	}
}

func TestReportRequestStructure(t *testing.T) {
	// Test ReportRequest structure
	request := ReportRequest{
		ReportType: "executive",
		Format:     "pdf",
		Filters: map[string]interface{}{
			"framework": "SOC 2",
			"priority":  "critical",
		},
		Recipients: []string{"admin@company.com"},
		Schedule: &ScheduleConfig{
			Frequency: "daily",
			Time:      "09:00",
			StartDate: time.Now(),
		},
		CustomFields: map[string]interface{}{
			"department": "IT",
		},
	}

	// Verify required fields
	if request.ReportType == "" {
		t.Error("Expected non-empty ReportType")
	}

	if request.Format == "" {
		t.Error("Expected non-empty Format")
	}

	// Verify schedule structure if provided
	if request.Schedule != nil {
		if request.Schedule.Frequency == "" {
			t.Error("Expected non-empty Schedule.Frequency")
		}

		if request.Schedule.Time == "" {
			t.Error("Expected non-empty Schedule.Time")
		}

		if request.Schedule.StartDate.IsZero() {
			t.Error("Expected non-zero Schedule.StartDate")
		}
	}
}

func TestScheduleConfigStructure(t *testing.T) {
	// Test ScheduleConfig structure
	schedule := ScheduleConfig{
		Frequency: "weekly",
		Time:      "10:00",
		Days:      []string{"monday", "wednesday", "friday"},
		StartDate: time.Now(),
		EndDate:   nil,
	}

	// Verify required fields
	if schedule.Frequency == "" {
		t.Error("Expected non-empty Frequency")
	}

	if schedule.Time == "" {
		t.Error("Expected non-empty Time")
	}

	if schedule.StartDate.IsZero() {
		t.Error("Expected non-zero StartDate")
	}

	// Verify frequency values
	validFrequencies := []string{"daily", "weekly", "monthly"}
	validFrequency := false
	for _, f := range validFrequencies {
		if schedule.Frequency == f {
			validFrequency = true
			break
		}
	}
	if !validFrequency {
		t.Errorf("Expected valid frequency, got %s", schedule.Frequency)
	}

	// Verify time format (basic check)
	if len(schedule.Time) != 5 || schedule.Time[2] != ':' {
		t.Errorf("Expected time in HH:MM format, got %s", schedule.Time)
	}
}

func TestReportResponseStructure(t *testing.T) {
	// Test ReportResponse structure
	now := time.Now()
	expiresAt := now.AddDate(0, 0, 30)

	response := ReportResponse{
		ID:          "report_001",
		Title:       "Executive Summary Report",
		Type:        "executive",
		Format:      "pdf",
		Status:      "completed",
		URL:         "/v1/reports/report_001",
		Size:        1024 * 1024, // 1MB
		GeneratedAt: now,
		ExpiresAt:   &expiresAt,
		Metadata: map[string]interface{}{
			"filters": map[string]string{
				"framework": "SOC 2",
			},
		},
	}

	// Verify required fields
	if response.ID == "" {
		t.Error("Expected non-empty ID")
	}

	if response.Title == "" {
		t.Error("Expected non-empty Title")
	}

	if response.Type == "" {
		t.Error("Expected non-empty Type")
	}

	if response.Format == "" {
		t.Error("Expected non-empty Format")
	}

	if response.Status == "" {
		t.Error("Expected non-empty Status")
	}

	if response.GeneratedAt.IsZero() {
		t.Error("Expected non-zero GeneratedAt")
	}

	// Verify status values
	validStatuses := []string{"pending", "generating", "completed", "failed"}
	validStatus := false
	for _, s := range validStatuses {
		if response.Status == s {
			validStatus = true
			break
		}
	}
	if !validStatus {
		t.Errorf("Expected valid status, got %s", response.Status)
	}

	// Verify size is positive
	if response.Size <= 0 {
		t.Error("Expected positive size")
	}
}

func TestReportMetricsStructure(t *testing.T) {
	// Test ReportMetrics structure
	metrics := ReportMetrics{
		TotalReports:    156,
		ReportsToday:    8,
		AverageSize:     2.3, // MB
		MostPopularType: "executive",
		SuccessRate:     98.5,
		LastGenerated:   time.Now().AddDate(0, 0, -1),
	}

	// Verify required fields
	if metrics.TotalReports < 0 {
		t.Error("Expected non-negative TotalReports")
	}

	if metrics.ReportsToday < 0 {
		t.Error("Expected non-negative ReportsToday")
	}

	if metrics.AverageSize < 0 {
		t.Error("Expected non-negative AverageSize")
	}

	if metrics.MostPopularType == "" {
		t.Error("Expected non-empty MostPopularType")
	}

	if metrics.SuccessRate < 0 || metrics.SuccessRate > 100 {
		t.Error("Expected SuccessRate between 0 and 100")
	}

	if metrics.LastGenerated.IsZero() {
		t.Error("Expected non-zero LastGenerated")
	}
}
