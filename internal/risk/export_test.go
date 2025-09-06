package risk

import (
	"context"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"go.uber.org/zap"
)

func TestExportService_ExportRiskData(t *testing.T) {
	// Create test logger
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)

	// Create test components
	historyService := NewRiskHistoryService(logger, nil)
	alertService := NewAlertService(logger, CreateDefaultThresholds())
	reportService := NewReportService(logger, historyService, alertService)

	// Create export service
	service := NewExportService(logger, historyService, alertService, reportService)

	// Create test request
	request := ExportRequest{
		BusinessID: "test_business_123",
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
		Metadata: map[string]interface{}{
			"test": true,
		},
	}

	// Create test context
	ctx := context.WithValue(context.Background(), "request_id", "test_request_123")

	// Export data
	response, err := service.ExportRiskData(ctx, request)
	if err != nil {
		t.Fatalf("Failed to export risk data: %v", err)
	}

	// Verify response structure
	if response == nil {
		t.Fatal("Response should not be nil")
	}

	if response.ExportID == "" {
		t.Error("Export ID should not be empty")
	}

	if response.BusinessID != request.BusinessID {
		t.Errorf("Expected business ID %s, got %s", request.BusinessID, response.BusinessID)
	}

	if response.ExportType != request.ExportType {
		t.Errorf("Expected export type %s, got %s", request.ExportType, response.ExportType)
	}

	if response.Format != request.Format {
		t.Errorf("Expected format %s, got %s", request.Format, response.Format)
	}

	if response.GeneratedAt.IsZero() {
		t.Error("GeneratedAt should not be zero")
	}

	if response.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should not be zero")
	}

	if response.Data == nil {
		t.Error("Data should not be nil")
	}

	// Verify metadata is preserved
	if response.Metadata == nil {
		t.Error("Response metadata should not be nil")
	}

	if testValue, exists := response.Metadata["test"]; !exists || testValue != true {
		t.Error("Response metadata should preserve original metadata")
	}

	t.Logf("Exported data: ID=%s, BusinessID=%s, Type=%s, Format=%s, Records=%d",
		response.ExportID, response.BusinessID, response.ExportType, response.Format, response.RecordCount)
}

func TestExportService_ExportRiskData_CSV(t *testing.T) {
	// Create test logger
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)

	// Create test components
	historyService := NewRiskHistoryService(logger, nil)
	alertService := NewAlertService(logger, CreateDefaultThresholds())
	reportService := NewReportService(logger, historyService, alertService)

	// Create export service
	service := NewExportService(logger, historyService, alertService, reportService)

	// Create test request for CSV export
	request := ExportRequest{
		BusinessID: "test_business_456",
		ExportType: ExportTypeFactors,
		Format:     ExportFormatCSV,
	}

	// Create test context
	ctx := context.WithValue(context.Background(), "request_id", "test_request_456")

	// Export data
	response, err := service.ExportRiskData(ctx, request)
	if err != nil {
		t.Fatalf("Failed to export risk data as CSV: %v", err)
	}

	// Verify response structure
	if response == nil {
		t.Fatal("Response should not be nil")
	}

	if response.Format != ExportFormatCSV {
		t.Errorf("Expected format %s, got %s", ExportFormatCSV, response.Format)
	}

	// Verify CSV data is a string
	if csvData, ok := response.Data.(string); !ok {
		t.Error("CSV data should be a string")
	} else if csvData == "" {
		// CSV data can be empty when there are no factors to export
		t.Log("CSV data is empty (no factors to export)")
	}

	t.Logf("Exported CSV data: ID=%s, BusinessID=%s, Type=%s, Records=%d",
		response.ExportID, response.BusinessID, response.ExportType, response.RecordCount)
}

func TestExportService_CreateExportJob(t *testing.T) {
	// Create test logger
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)

	// Create test components
	historyService := NewRiskHistoryService(logger, nil)
	alertService := NewAlertService(logger, CreateDefaultThresholds())
	reportService := NewReportService(logger, historyService, alertService)

	// Create export service
	service := NewExportService(logger, historyService, alertService, reportService)

	// Create test request
	request := ExportRequest{
		BusinessID: "test_business_789",
		ExportType: ExportTypeAll,
		Format:     ExportFormatXLSX,
		Metadata: map[string]interface{}{
			"test": true,
		},
	}

	// Create test context
	ctx := context.WithValue(context.Background(), "request_id", "test_request_789")

	// Create export job
	job, err := service.CreateExportJob(ctx, request)
	if err != nil {
		t.Fatalf("Failed to create export job: %v", err)
	}

	// Verify job structure
	if job == nil {
		t.Fatal("Job should not be nil")
	}

	if job.ID == "" {
		t.Error("Job ID should not be empty")
	}

	if job.BusinessID != request.BusinessID {
		t.Errorf("Expected business ID %s, got %s", request.BusinessID, job.BusinessID)
	}

	if job.ExportType != request.ExportType {
		t.Errorf("Expected export type %s, got %s", request.ExportType, job.ExportType)
	}

	if job.Format != request.Format {
		t.Errorf("Expected format %s, got %s", request.Format, job.Format)
	}

	if job.Status != "pending" {
		t.Errorf("Expected status 'pending', got %s", job.Status)
	}

	if job.Progress != 0 {
		t.Errorf("Expected progress 0, got %d", job.Progress)
	}

	if job.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	// Verify metadata is preserved
	if job.Metadata == nil {
		t.Error("Job metadata should not be nil")
	}

	if testValue, exists := job.Metadata["test"]; !exists || testValue != true {
		t.Error("Job metadata should preserve original metadata")
	}

	t.Logf("Created export job: ID=%s, BusinessID=%s, Type=%s, Format=%s, Status=%s",
		job.ID, job.BusinessID, job.ExportType, job.Format, job.Status)
}

func TestExportService_GetExportJob(t *testing.T) {
	// Create test logger
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)

	// Create test components
	historyService := NewRiskHistoryService(logger, nil)
	alertService := NewAlertService(logger, CreateDefaultThresholds())
	reportService := NewReportService(logger, historyService, alertService)

	// Create export service
	service := NewExportService(logger, historyService, alertService, reportService)

	// Create test context
	ctx := context.WithValue(context.Background(), "request_id", "test_request_999")

	// Get export job
	jobID := "test_job_123"
	job, err := service.GetExportJob(ctx, jobID)
	if err != nil {
		t.Fatalf("Failed to get export job: %v", err)
	}

	// Verify job structure
	if job == nil {
		t.Fatal("Job should not be nil")
	}

	if job.ID != jobID {
		t.Errorf("Expected job ID %s, got %s", jobID, job.ID)
	}

	if job.Status != "completed" {
		t.Errorf("Expected status 'completed', got %s", job.Status)
	}

	if job.Progress != 100 {
		t.Errorf("Expected progress 100, got %d", job.Progress)
	}

	if job.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	if job.CompletedAt == nil {
		t.Error("CompletedAt should not be nil")
	}

	t.Logf("Retrieved export job: ID=%s, Status=%s, Progress=%d",
		job.ID, job.Status, job.Progress)
}

func TestExportService_ValidateExportRequest(t *testing.T) {
	// Create test logger
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)

	// Create test components
	historyService := NewRiskHistoryService(logger, nil)
	alertService := NewAlertService(logger, CreateDefaultThresholds())
	reportService := NewReportService(logger, historyService, alertService)

	// Create export service
	service := NewExportService(logger, historyService, alertService, reportService)

	tests := []struct {
		name    string
		request ExportRequest
		wantErr bool
	}{
		{
			name: "Valid request",
			request: ExportRequest{
				BusinessID: "test_business",
				ExportType: ExportTypeAssessments,
				Format:     ExportFormatJSON,
			},
			wantErr: false,
		},
		{
			name: "Missing business ID",
			request: ExportRequest{
				ExportType: ExportTypeAssessments,
				Format:     ExportFormatJSON,
			},
			wantErr: true,
		},
		{
			name: "Missing export type",
			request: ExportRequest{
				BusinessID: "test_business",
				Format:     ExportFormatJSON,
			},
			wantErr: true,
		},
		{
			name: "Missing format",
			request: ExportRequest{
				BusinessID: "test_business",
				ExportType: ExportTypeAssessments,
			},
			wantErr: true,
		},
		{
			name: "Invalid export type",
			request: ExportRequest{
				BusinessID: "test_business",
				ExportType: "invalid_type",
				Format:     ExportFormatJSON,
			},
			wantErr: true,
		},
		{
			name: "Invalid format",
			request: ExportRequest{
				BusinessID: "test_business",
				ExportType: ExportTypeAssessments,
				Format:     "invalid_format",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateExportRequest(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateExportRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExportService_FormatExportData(t *testing.T) {
	// Create test logger
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)

	// Create test components
	historyService := NewRiskHistoryService(logger, nil)
	alertService := NewAlertService(logger, CreateDefaultThresholds())
	reportService := NewReportService(logger, historyService, alertService)

	// Create export service
	service := NewExportService(logger, historyService, alertService, reportService)

	// Test data
	testData := []map[string]interface{}{
		{
			"id":     "1",
			"name":   "Test Item 1",
			"value":  100.5,
			"active": true,
		},
		{
			"id":     "2",
			"name":   "Test Item 2",
			"value":  200.75,
			"active": false,
		},
	}

	tests := []struct {
		name    string
		data    interface{}
		format  ExportFormat
		wantErr bool
	}{
		{
			name:    "JSON format",
			data:    testData,
			format:  ExportFormatJSON,
			wantErr: false,
		},
		{
			name:    "CSV format",
			data:    testData,
			format:  ExportFormatCSV,
			wantErr: false,
		},
		{
			name:    "XML format",
			data:    testData,
			format:  ExportFormatXML,
			wantErr: false,
		},
		{
			name:    "PDF format",
			data:    testData,
			format:  ExportFormatPDF,
			wantErr: false,
		},
		{
			name:    "XLSX format",
			data:    testData,
			format:  ExportFormatXLSX,
			wantErr: false,
		},
		{
			name:    "Invalid format",
			data:    testData,
			format:  "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.formatExportData(tt.data, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("formatExportData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("formatExportData() returned nil result for valid format")
			}
		})
	}
}

func TestExportService_ConvertToString(t *testing.T) {
	// Create test logger
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)

	// Create test components
	historyService := NewRiskHistoryService(logger, nil)
	alertService := NewAlertService(logger, CreateDefaultThresholds())
	reportService := NewReportService(logger, historyService, alertService)

	// Create export service
	service := NewExportService(logger, historyService, alertService, reportService)

	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{
			name:  "String",
			input: "test string",
			want:  "test string",
		},
		{
			name:  "Integer",
			input: 123,
			want:  "123",
		},
		{
			name:  "Float",
			input: 123.45,
			want:  "123.45",
		},
		{
			name:  "Boolean true",
			input: true,
			want:  "true",
		},
		{
			name:  "Boolean false",
			input: false,
			want:  "false",
		},
		{
			name:  "Time",
			input: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			want:  "2023-01-01T12:00:00Z",
		},
		{
			name:  "Nil",
			input: nil,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.convertToString(tt.input)
			if got != tt.want {
				t.Errorf("convertToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExportService_CountRecords(t *testing.T) {
	// Create test logger
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)

	// Create test components
	historyService := NewRiskHistoryService(logger, nil)
	alertService := NewAlertService(logger, CreateDefaultThresholds())
	reportService := NewReportService(logger, historyService, alertService)

	// Create export service
	service := NewExportService(logger, historyService, alertService, reportService)

	tests := []struct {
		name string
		data interface{}
		want int
	}{
		{
			name: "Slice of interfaces",
			data: []interface{}{"a", "b", "c"},
			want: 3,
		},
		{
			name: "Slice of maps",
			data: []map[string]interface{}{
				{"id": "1"},
				{"id": "2"},
			},
			want: 2,
		},
		{
			name: "Single map",
			data: map[string]interface{}{"id": "1"},
			want: 1,
		},
		{
			name: "String",
			data: "test",
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.countRecords(tt.data)
			if got != tt.want {
				t.Errorf("countRecords() = %v, want %v", got, tt.want)
			}
		})
	}
}
