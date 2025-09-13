package risk

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestExportService_ExportRiskAssessment(t *testing.T) {
	logger := zap.NewNop()
	service := NewExportService(logger)

	// Create test assessment
	assessment := &RiskAssessment{
		ID:           "test-assessment-1",
		BusinessID:   "test-business-1",
		BusinessName: "Test Business",
		OverallScore: 75.5,
		OverallLevel: RiskLevelHigh,
		CategoryScores: map[RiskCategory]RiskScore{
			RiskCategoryFinancial: {
				FactorID:     "financial-1",
				FactorName:   "Financial Risk",
				Category:     RiskCategoryFinancial,
				Score:        80.0,
				Level:        RiskLevelHigh,
				Confidence:   0.9,
				Explanation:  "High financial risk due to debt levels",
				CalculatedAt: time.Now(),
			},
		},
		FactorScores: []RiskScore{
			{
				FactorID:     "factor-1",
				FactorName:   "Test Factor",
				Category:     RiskCategoryOperational,
				Score:        70.0,
				Level:        RiskLevelMedium,
				Confidence:   0.8,
				Explanation:  "Test factor explanation",
				CalculatedAt: time.Now(),
			},
		},
		AssessedAt: time.Now(),
		ValidUntil: time.Now().Add(24 * time.Hour),
		AlertLevel: RiskLevelMedium,
	}

	tests := []struct {
		name     string
		format   ExportFormat
		wantErr  bool
		validate func(t *testing.T, response *ExportResponse)
	}{
		{
			name:    "export to JSON",
			format:  ExportFormatJSON,
			wantErr: false,
			validate: func(t *testing.T, response *ExportResponse) {
				assert.Equal(t, ExportFormatJSON, response.Format)
				assert.Equal(t, ExportTypeAssessments, response.ExportType)
				assert.Equal(t, 1, response.RecordCount)
				assert.NotEmpty(t, response.ExportID)
				assert.NotEmpty(t, response.Data)

				// Verify JSON data is valid
				data, ok := response.Data.(string)
				require.True(t, ok)
				assert.Contains(t, data, "test-assessment-1")
				assert.Contains(t, data, "Test Business")
			},
		},
		{
			name:    "export to CSV",
			format:  ExportFormatCSV,
			wantErr: false,
			validate: func(t *testing.T, response *ExportResponse) {
				assert.Equal(t, ExportFormatCSV, response.Format)
				assert.Equal(t, ExportTypeAssessments, response.ExportType)
				assert.Equal(t, 1, response.RecordCount)

				// Verify CSV data is valid
				data, ok := response.Data.(string)
				require.True(t, ok)
				assert.Contains(t, data, "ID,BusinessID,BusinessName")
				assert.Contains(t, data, "test-assessment-1")
				assert.Contains(t, data, "Test Business")
			},
		},
		{
			name:    "export to XML",
			format:  ExportFormatXML,
			wantErr: false,
			validate: func(t *testing.T, response *ExportResponse) {
				assert.Equal(t, ExportFormatXML, response.Format)
				assert.Equal(t, ExportTypeAssessments, response.ExportType)
				assert.Equal(t, 1, response.RecordCount)

				// Verify XML data is valid
				data, ok := response.Data.(string)
				require.True(t, ok)
				assert.Contains(t, data, "<RiskAssessment>")
				assert.Contains(t, data, "test-assessment-1")
			},
		},
		{
			name:    "export to PDF",
			format:  ExportFormatPDF,
			wantErr: false,
			validate: func(t *testing.T, response *ExportResponse) {
				assert.Equal(t, ExportFormatPDF, response.Format)
				assert.Equal(t, ExportTypeAssessments, response.ExportType)
				assert.Equal(t, 1, response.RecordCount)

				// Verify PDF data is valid
				data, ok := response.Data.(string)
				require.True(t, ok)
				assert.Contains(t, data, "Risk Assessment Report")
				assert.Contains(t, data, "Test Business")
			},
		},
		{
			name:    "export to XLSX",
			format:  ExportFormatXLSX,
			wantErr: false,
			validate: func(t *testing.T, response *ExportResponse) {
				assert.Equal(t, ExportFormatXLSX, response.Format)
				assert.Equal(t, ExportTypeAssessments, response.ExportType)
				assert.Equal(t, 1, response.RecordCount)
			},
		},
		{
			name:    "unsupported format",
			format:  ExportFormat("unsupported"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "request_id", "test-request-1")

			response, err := service.ExportRiskAssessment(ctx, assessment, tt.format)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)

				if tt.validate != nil {
					tt.validate(t, response)
				}
			}
		})
	}
}

func TestExportService_ExportRiskAssessments(t *testing.T) {
	logger := zap.NewNop()
	service := NewExportService(logger)

	// Create test assessments
	assessments := []*RiskAssessment{
		{
			ID:           "test-assessment-1",
			BusinessID:   "test-business-1",
			BusinessName: "Test Business 1",
			OverallScore: 75.5,
			OverallLevel: RiskLevelHigh,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
			AlertLevel:   RiskLevelMedium,
		},
		{
			ID:           "test-assessment-2",
			BusinessID:   "test-business-2",
			BusinessName: "Test Business 2",
			OverallScore: 45.0,
			OverallLevel: RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
			AlertLevel:   RiskLevelLow,
		},
	}

	tests := []struct {
		name     string
		format   ExportFormat
		wantErr  bool
		validate func(t *testing.T, response *ExportResponse)
	}{
		{
			name:    "export multiple to JSON",
			format:  ExportFormatJSON,
			wantErr: false,
			validate: func(t *testing.T, response *ExportResponse) {
				assert.Equal(t, ExportFormatJSON, response.Format)
				assert.Equal(t, ExportTypeAssessments, response.ExportType)
				assert.Equal(t, 2, response.RecordCount)

				// Verify JSON data contains both assessments
				data, ok := response.Data.(string)
				require.True(t, ok)
				assert.Contains(t, data, "test-assessment-1")
				assert.Contains(t, data, "test-assessment-2")
			},
		},
		{
			name:    "export multiple to CSV",
			format:  ExportFormatCSV,
			wantErr: false,
			validate: func(t *testing.T, response *ExportResponse) {
				assert.Equal(t, ExportFormatCSV, response.Format)
				assert.Equal(t, ExportTypeAssessments, response.ExportType)
				assert.Equal(t, 2, response.RecordCount)

				// Verify CSV data contains both assessments
				data, ok := response.Data.(string)
				require.True(t, ok)
				assert.Contains(t, data, "test-assessment-1")
				assert.Contains(t, data, "test-assessment-2")
			},
		},
		{
			name:    "empty assessments",
			format:  ExportFormatJSON,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "request_id", "test-request-1")

			var testAssessments []*RiskAssessment
			if tt.name == "empty assessments" {
				testAssessments = []*RiskAssessment{}
			} else {
				testAssessments = assessments
			}

			response, err := service.ExportRiskAssessments(ctx, testAssessments, tt.format)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)

				if tt.validate != nil {
					tt.validate(t, response)
				}
			}
		})
	}
}

func TestExportService_ExportRiskFactors(t *testing.T) {
	logger := zap.NewNop()
	service := NewExportService(logger)

	// Create test factors
	factors := []RiskScore{
		{
			FactorID:     "factor-1",
			FactorName:   "Financial Risk",
			Category:     RiskCategoryFinancial,
			Score:        80.0,
			Level:        RiskLevelHigh,
			Confidence:   0.9,
			Explanation:  "High financial risk due to debt levels",
			CalculatedAt: time.Now(),
		},
		{
			FactorID:     "factor-2",
			FactorName:   "Operational Risk",
			Category:     RiskCategoryOperational,
			Score:        60.0,
			Level:        RiskLevelMedium,
			Confidence:   0.8,
			Explanation:  "Medium operational risk",
			CalculatedAt: time.Now(),
		},
	}

	tests := []struct {
		name     string
		format   ExportFormat
		wantErr  bool
		validate func(t *testing.T, response *ExportResponse)
	}{
		{
			name:    "export factors to JSON",
			format:  ExportFormatJSON,
			wantErr: false,
			validate: func(t *testing.T, response *ExportResponse) {
				assert.Equal(t, ExportFormatJSON, response.Format)
				assert.Equal(t, ExportTypeFactors, response.ExportType)
				assert.Equal(t, 2, response.RecordCount)

				// Verify JSON data contains both factors
				data, ok := response.Data.(string)
				require.True(t, ok)
				assert.Contains(t, data, "factor-1")
				assert.Contains(t, data, "factor-2")
			},
		},
		{
			name:    "export factors to CSV",
			format:  ExportFormatCSV,
			wantErr: false,
			validate: func(t *testing.T, response *ExportResponse) {
				assert.Equal(t, ExportFormatCSV, response.Format)
				assert.Equal(t, ExportTypeFactors, response.ExportType)
				assert.Equal(t, 2, response.RecordCount)

				// Verify CSV data contains both factors
				data, ok := response.Data.(string)
				require.True(t, ok)
				assert.Contains(t, data, "FactorID,FactorName,Category")
				assert.Contains(t, data, "factor-1")
				assert.Contains(t, data, "factor-2")
			},
		},
		{
			name:    "empty factors",
			format:  ExportFormatJSON,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "request_id", "test-request-1")

			var testFactors []RiskScore
			if tt.name == "empty factors" {
				testFactors = []RiskScore{}
			} else {
				testFactors = factors
			}

			response, err := service.ExportRiskFactors(ctx, testFactors, tt.format)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)

				if tt.validate != nil {
					tt.validate(t, response)
				}
			}
		})
	}
}

func TestExportService_ExportRiskTrends(t *testing.T) {
	logger := zap.NewNop()
	service := NewExportService(logger)

	// Create test trends
	trends := []RiskTrend{
		{
			BusinessID:   "test-business-1",
			Category:     RiskCategoryFinancial,
			Score:        75.0,
			Level:        RiskLevelHigh,
			RecordedAt:   time.Now(),
			ChangeFrom:   5.0,
			ChangePeriod: "1 month",
		},
		{
			BusinessID:   "test-business-1",
			Category:     RiskCategoryOperational,
			Score:        60.0,
			Level:        RiskLevelMedium,
			RecordedAt:   time.Now(),
			ChangeFrom:   -2.0,
			ChangePeriod: "1 month",
		},
	}

	tests := []struct {
		name     string
		format   ExportFormat
		wantErr  bool
		validate func(t *testing.T, response *ExportResponse)
	}{
		{
			name:    "export trends to JSON",
			format:  ExportFormatJSON,
			wantErr: false,
			validate: func(t *testing.T, response *ExportResponse) {
				assert.Equal(t, ExportFormatJSON, response.Format)
				assert.Equal(t, ExportTypeTrends, response.ExportType)
				assert.Equal(t, 2, response.RecordCount)

				// Verify JSON data contains both trends
				data, ok := response.Data.(string)
				require.True(t, ok)
				assert.Contains(t, data, "test-business-1")
				assert.Contains(t, data, "Financial")
			},
		},
		{
			name:    "export trends to CSV",
			format:  ExportFormatCSV,
			wantErr: false,
			validate: func(t *testing.T, response *ExportResponse) {
				assert.Equal(t, ExportFormatCSV, response.Format)
				assert.Equal(t, ExportTypeTrends, response.ExportType)
				assert.Equal(t, 2, response.RecordCount)

				// Verify CSV data contains both trends
				data, ok := response.Data.(string)
				require.True(t, ok)
				assert.Contains(t, data, "BusinessID,Category,Score")
				assert.Contains(t, data, "test-business-1")
			},
		},
		{
			name:    "empty trends",
			format:  ExportFormatJSON,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "request_id", "test-request-1")

			var testTrends []RiskTrend
			if tt.name == "empty trends" {
				testTrends = []RiskTrend{}
			} else {
				testTrends = trends
			}

			response, err := service.ExportRiskTrends(ctx, testTrends, tt.format)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)

				if tt.validate != nil {
					tt.validate(t, response)
				}
			}
		})
	}
}

func TestExportService_ExportRiskAlerts(t *testing.T) {
	logger := zap.NewNop()
	service := NewExportService(logger)

	// Create test alerts
	now := time.Now()
	alerts := []RiskAlert{
		{
			ID:             "alert-1",
			BusinessID:     "test-business-1",
			RiskFactor:     "financial-risk",
			Level:          RiskLevelHigh,
			Message:        "High financial risk detected",
			Score:          85.0,
			Threshold:      80.0,
			TriggeredAt:    now,
			Acknowledged:   false,
			AcknowledgedAt: nil,
		},
		{
			ID:             "alert-2",
			BusinessID:     "test-business-1",
			RiskFactor:     "operational-risk",
			Level:          RiskLevelMedium,
			Message:        "Medium operational risk detected",
			Score:          65.0,
			Threshold:      60.0,
			TriggeredAt:    now,
			Acknowledged:   true,
			AcknowledgedAt: &now,
		},
	}

	tests := []struct {
		name     string
		format   ExportFormat
		wantErr  bool
		validate func(t *testing.T, response *ExportResponse)
	}{
		{
			name:    "export alerts to JSON",
			format:  ExportFormatJSON,
			wantErr: false,
			validate: func(t *testing.T, response *ExportResponse) {
				assert.Equal(t, ExportFormatJSON, response.Format)
				assert.Equal(t, ExportTypeAlerts, response.ExportType)
				assert.Equal(t, 2, response.RecordCount)

				// Verify JSON data contains both alerts
				data, ok := response.Data.(string)
				require.True(t, ok)
				assert.Contains(t, data, "alert-1")
				assert.Contains(t, data, "alert-2")
			},
		},
		{
			name:    "export alerts to CSV",
			format:  ExportFormatCSV,
			wantErr: false,
			validate: func(t *testing.T, response *ExportResponse) {
				assert.Equal(t, ExportFormatCSV, response.Format)
				assert.Equal(t, ExportTypeAlerts, response.ExportType)
				assert.Equal(t, 2, response.RecordCount)

				// Verify CSV data contains both alerts
				data, ok := response.Data.(string)
				require.True(t, ok)
				assert.Contains(t, data, "ID,BusinessID,RiskFactor")
				assert.Contains(t, data, "alert-1")
				assert.Contains(t, data, "alert-2")
			},
		},
		{
			name:    "empty alerts",
			format:  ExportFormatJSON,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "request_id", "test-request-1")

			var testAlerts []RiskAlert
			if tt.name == "empty alerts" {
				testAlerts = []RiskAlert{}
			} else {
				testAlerts = alerts
			}

			response, err := service.ExportRiskAlerts(ctx, testAlerts, tt.format)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)

				if tt.validate != nil {
					tt.validate(t, response)
				}
			}
		})
	}
}

func TestExportService_ValidateExportRequest(t *testing.T) {
	logger := zap.NewNop()
	service := NewExportService(logger)

	tests := []struct {
		name    string
		request *ExportRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: &ExportRequest{
				BusinessID: "test-business-1",
				ExportType: ExportTypeAssessments,
				Format:     ExportFormatJSON,
			},
			wantErr: false,
		},
		{
			name: "missing business_id",
			request: &ExportRequest{
				ExportType: ExportTypeAssessments,
				Format:     ExportFormatJSON,
			},
			wantErr: true,
		},
		{
			name: "missing export_type",
			request: &ExportRequest{
				BusinessID: "test-business-1",
				Format:     ExportFormatJSON,
			},
			wantErr: true,
		},
		{
			name: "missing format",
			request: &ExportRequest{
				BusinessID: "test-business-1",
				ExportType: ExportTypeAssessments,
			},
			wantErr: true,
		},
		{
			name: "invalid export_type",
			request: &ExportRequest{
				BusinessID: "test-business-1",
				ExportType: ExportType("invalid"),
				Format:     ExportFormatJSON,
			},
			wantErr: true,
		},
		{
			name: "invalid format",
			request: &ExportRequest{
				BusinessID: "test-business-1",
				ExportType: ExportTypeAssessments,
				Format:     ExportFormat("invalid"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateExportRequest(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExportService_CalculateDataSize(t *testing.T) {
	logger := zap.NewNop()
	service := NewExportService(logger)

	tests := []struct {
		name     string
		data     interface{}
		expected int64
	}{
		{
			name:     "string data",
			data:     "test string",
			expected: 11,
		},
		{
			name:     "empty string",
			data:     "",
			expected: 0,
		},
		{
			name:     "complex data",
			data:     map[string]interface{}{"key": "value"},
			expected: 1024, // Default estimate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := service.calculateDataSize(tt.data)
			assert.Equal(t, tt.expected, size)
		})
	}
}

func TestExportService_ContextHandling(t *testing.T) {
	logger := zap.NewNop()
	service := NewExportService(logger)

	assessment := &RiskAssessment{
		ID:           "test-assessment-1",
		BusinessID:   "test-business-1",
		BusinessName: "Test Business",
		OverallScore: 75.5,
		OverallLevel: RiskLevelHigh,
		AssessedAt:   time.Now(),
		ValidUntil:   time.Now().Add(24 * time.Hour),
		AlertLevel:   RiskLevelMedium,
	}

	tests := []struct {
		name        string
		ctx         context.Context
		expectError bool
	}{
		{
			name:        "context with request_id",
			ctx:         context.WithValue(context.Background(), "request_id", "test-request-123"),
			expectError: false,
		},
		{
			name:        "context without request_id",
			ctx:         context.Background(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := service.ExportRiskAssessment(tt.ctx, assessment, ExportFormatJSON)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotEmpty(t, response.ExportID)
				assert.Equal(t, assessment.BusinessID, response.BusinessID)
			}
		})
	}
}

func TestExportService_ResponseMetadata(t *testing.T) {
	logger := zap.NewNop()
	service := NewExportService(logger)

	assessment := &RiskAssessment{
		ID:           "test-assessment-1",
		BusinessID:   "test-business-1",
		BusinessName: "Test Business",
		OverallScore: 75.5,
		OverallLevel: RiskLevelHigh,
		AssessedAt:   time.Now(),
		ValidUntil:   time.Now().Add(24 * time.Hour),
		AlertLevel:   RiskLevelMedium,
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	response, err := service.ExportRiskAssessment(ctx, assessment, ExportFormatJSON)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Verify response structure
	assert.NotEmpty(t, response.ExportID)
	assert.Equal(t, assessment.BusinessID, response.BusinessID)
	assert.Equal(t, ExportTypeAssessments, response.ExportType)
	assert.Equal(t, ExportFormatJSON, response.Format)
	assert.Equal(t, 1, response.RecordCount)
	assert.NotZero(t, response.GeneratedAt)
	assert.NotZero(t, response.ExpiresAt)
	assert.NotNil(t, response.Metadata)

	// Verify metadata
	metadata := response.Metadata
	assert.Contains(t, metadata, "processing_time_ms")
	assert.Contains(t, metadata, "export_size_bytes")
	assert.Contains(t, metadata, "request_id")
	assert.Equal(t, "test-request-123", metadata["request_id"])

	// Verify expiration time is reasonable (24 hours from generation)
	expectedExpiration := response.GeneratedAt.Add(24 * time.Hour)
	assert.WithinDuration(t, expectedExpiration, response.ExpiresAt, time.Minute)
}
