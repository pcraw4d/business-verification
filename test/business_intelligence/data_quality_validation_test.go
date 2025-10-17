package business_intelligence_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"kyb-platform/internal/modules/business_intelligence"
)

// MockQualityMonitor is a mock implementation of QualityMonitor
type MockQualityMonitor struct {
	mock.Mock
}

func (m *MockQualityMonitor) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockQualityMonitor) GetType() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockQualityMonitor) GetSupportedDimensions() []business_intelligence.QualityDimension {
	args := m.Called()
	return args.Get(0).([]business_intelligence.QualityDimension)
}

func (m *MockQualityMonitor) MonitorData(ctx context.Context, data interface{}) (*business_intelligence.QualityAssessment, error) {
	args := m.Called(ctx, data)
	return args.Get(0).(*business_intelligence.QualityAssessment), args.Error(1)
}

func (m *MockQualityMonitor) GetMonitoringMetrics() *business_intelligence.MonitoringMetrics {
	args := m.Called()
	return args.Get(0).(*business_intelligence.MonitoringMetrics)
}

// MockQualityValidator is a mock implementation of QualityValidator
type MockQualityValidator struct {
	mock.Mock
}

func (m *MockQualityValidator) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockQualityValidator) GetValidationRules() []business_intelligence.QualityRule {
	args := m.Called()
	return args.Get(0).([]business_intelligence.QualityRule)
}

func (m *MockQualityValidator) ValidateData(ctx context.Context, data interface{}) (*business_intelligence.ValidationResult, error) {
	args := m.Called(ctx, data)
	return args.Get(0).(*business_intelligence.ValidationResult), args.Error(1)
}

func (m *MockQualityValidator) GetValidationMetrics() *business_intelligence.ValidationMetrics {
	args := m.Called()
	return args.Get(0).(*business_intelligence.ValidationMetrics)
}

// MockQualityAssessor is a mock implementation of QualityAssessor
type MockQualityAssessor struct {
	mock.Mock
}

func (m *MockQualityAssessor) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockQualityAssessor) GetAssessmentTypes() []business_intelligence.AssessmentType {
	args := m.Called()
	return args.Get(0).([]business_intelligence.AssessmentType)
}

func (m *MockQualityAssessor) AssessQuality(ctx context.Context, assessments []*business_intelligence.QualityAssessment) (*business_intelligence.QualityScore, error) {
	args := m.Called(ctx, assessments)
	return args.Get(0).(*business_intelligence.QualityScore), args.Error(1)
}

func (m *MockQualityAssessor) GetAssessmentMetrics() *business_intelligence.AssessmentMetrics {
	args := m.Called()
	return args.Get(0).(*business_intelligence.AssessmentMetrics)
}

// MockQualityAlerter is a mock implementation of QualityAlerter
type MockQualityAlerter struct {
	mock.Mock
}

func (m *MockQualityAlerter) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockQualityAlerter) GetAlertTypes() []business_intelligence.AlertType {
	args := m.Called()
	return args.Get(0).([]business_intelligence.AlertType)
}

func (m *MockQualityAlerter) SendAlert(ctx context.Context, alert *business_intelligence.QualityAlert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockQualityAlerter) GetAlertingMetrics() *business_intelligence.AlertingMetrics {
	args := m.Called()
	return args.Get(0).(*business_intelligence.AlertingMetrics)
}

// MockQualityReporter is a mock implementation of QualityReporter
type MockQualityReporter struct {
	mock.Mock
}

func (m *MockQualityReporter) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockQualityReporter) GetReportTypes() []business_intelligence.ReportType {
	args := m.Called()
	return args.Get(0).([]business_intelligence.ReportType)
}

func (m *MockQualityReporter) GenerateReport(ctx context.Context, request *business_intelligence.ReportRequest) (*business_intelligence.QualityReport, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*business_intelligence.QualityReport), args.Error(1)
}

func (m *MockQualityReporter) GetReportingMetrics() *business_intelligence.ReportingMetrics {
	args := m.Called()
	return args.Get(0).(*business_intelligence.ReportingMetrics)
}

// TestDataQualityMonitoringSystem_Validation tests data quality validation functionality
func TestDataQualityMonitoringSystem_Validation(t *testing.T) {
	logger := zap.NewNop()

	// Configure quality monitoring system
	config := business_intelligence.QualityMonitoringConfig{
		EnableRealTimeMonitoring:     true,
		MonitoringInterval:           30 * time.Second,
		EnableBatchMonitoring:        true,
		BatchSize:                    100,
		EnableContinuousMonitoring:   true,
		DefaultQualityThreshold:      0.8,
		CriticalQualityThreshold:     0.5,
		WarningQualityThreshold:      0.7,
		EnableAdaptiveThresholds:     true,
		EnableCompletenessMonitoring: true,
		EnableAccuracyMonitoring:     true,
		EnableConsistencyMonitoring:  true,
		EnableTimelinessMonitoring:   true,
		EnableValidityMonitoring:     true,
		EnableUniquenessMonitoring:   true,
		EnableAlerting:               true,
		AlertThreshold:               0.6,
		AlertCooldownPeriod:          5 * time.Minute,
		MaxAlertsPerHour:             10,
		EnableEscalation:             true,
		EnableReporting:              true,
		ReportGenerationInterval:     1 * time.Hour,
		EnableRealTimeReports:        true,
		EnableHistoricalReports:      true,
		ReportRetentionPeriod:        30 * 24 * time.Hour,
		EnableParallelProcessing:     true,
		MaxConcurrentMonitors:        5,
		ProcessingTimeout:            30 * time.Second,
		EnableQualityDataStorage:     true,
		StorageRetentionPeriod:       90 * 24 * time.Hour,
		EnableQualityTrends:          true,
	}

	// Create quality monitoring system
	qualitySystem := business_intelligence.NewDataQualityMonitoringSystem(config, logger)

	// Create mock components
	mockMonitor := &MockQualityMonitor{}
	mockValidator := &MockQualityValidator{}
	mockAssessor := &MockQualityAssessor{}
	mockAlerter := &MockQualityAlerter{}
	mockReporter := &MockQualityReporter{}

	// Register mock components
	err := qualitySystem.RegisterMonitor(mockMonitor)
	require.NoError(t, err)

	err = qualitySystem.RegisterValidator(mockValidator)
	require.NoError(t, err)

	err = qualitySystem.RegisterAssessor(mockAssessor)
	require.NoError(t, err)

	err = qualitySystem.RegisterAlerter(mockAlerter)
	require.NoError(t, err)

	err = qualitySystem.RegisterReporter(mockReporter)
	require.NoError(t, err)

	t.Run("Data Quality Monitoring", func(t *testing.T) {
		// Setup mock monitor
		mockMonitor.On("GetName").Return("test_monitor")
		mockMonitor.On("GetType").Return("completeness_monitor")
		mockMonitor.On("GetSupportedDimensions").Return([]business_intelligence.QualityDimension{
			business_intelligence.QualityDimensionCompleteness,
			business_intelligence.QualityDimensionAccuracy,
		})
		mockMonitor.On("MonitorData", mock.Anything, mock.Anything).Return(&business_intelligence.QualityAssessment{
			ID:           "assessment_1",
			MonitorID:    "test_monitor",
			DataID:       "test_data_1",
			DataType:     "business_data",
			OverallScore: 0.85,
			QualityLevel: "good",
			Issues: []business_intelligence.QualityIssue{
				{
					ID:          "issue_1",
					Type:        "missing_field",
					Severity:    "warning",
					Description: "Optional field 'website' is missing",
					Field:       "website",
					DetectedAt:  time.Now(),
				},
			},
			Recommendations: []business_intelligence.QualityRecommendation{
				{
					ID:          "rec_1",
					Type:        "data_collection",
					Priority:    "medium",
					Description: "Consider collecting website information for better completeness",
					Action:      "Add website field to data collection form",
					Impact:      "Improved data completeness",
					Effort:      "Low",
					CreatedAt:   time.Now(),
				},
			},
			AssessedAt: time.Now(),
			ExpiresAt:  time.Now().Add(1 * time.Hour),
		}, nil)
		mockMonitor.On("GetMonitoringMetrics").Return(&business_intelligence.MonitoringMetrics{
			MonitorName:      "test_monitor",
			TotalAssessments: 1,
			AverageScore:     0.85,
			IssuesDetected:   1,
			LastAssessment:   time.Now(),
		})

		// Setup mock alerter
		mockAlerter.On("GetName").Return("test_alerter")
		mockAlerter.On("GetAlertTypes").Return([]business_intelligence.AlertType{
			business_intelligence.AlertTypeQualityDegradation,
			business_intelligence.AlertTypeThresholdBreach,
		})
		mockAlerter.On("SendAlert", mock.Anything, mock.Anything).Return(nil)
		mockAlerter.On("GetAlertingMetrics").Return(&business_intelligence.AlertingMetrics{
			AlerterName:  "test_alerter",
			TotalAlerts:  0,
			AlertsSent:   0,
			AlertsFailed: 0,
			LastAlert:    time.Time{},
		})

		// Test data quality monitoring
		ctx := context.Background()
		testData := map[string]interface{}{
			"business_name": "Test Company",
			"industry":      "Technology",
			"revenue":       1000000,
			"employees":     50,
		}

		assessment, err := qualitySystem.MonitorDataQuality(ctx, testData, "test_data_1")

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, assessment)
		assert.Equal(t, "test_data_1", assessment.DataID)
		assert.Equal(t, 0.85, assessment.OverallScore)
		assert.Equal(t, "good", assessment.QualityLevel)
		assert.NotEmpty(t, assessment.Issues)
		assert.NotEmpty(t, assessment.Recommendations)

		// Verify issue details
		issue := assessment.Issues[0]
		assert.Equal(t, "missing_field", issue.Type)
		assert.Equal(t, "warning", issue.Severity)
		assert.Equal(t, "website", issue.Field)

		// Verify recommendation details
		recommendation := assessment.Recommendations[0]
		assert.Equal(t, "data_collection", recommendation.Type)
		assert.Equal(t, "medium", recommendation.Priority)

		mockMonitor.AssertExpectations(t)
		mockAlerter.AssertExpectations(t)
	})

	t.Run("Data Validation", func(t *testing.T) {
		// Reset mocks
		mockValidator.ExpectedCalls = nil

		// Setup mock validator
		mockValidator.On("GetName").Return("test_validator")
		mockValidator.On("GetValidationRules").Return([]business_intelligence.QualityRule{
			{
				ID:          "rule_1",
				Name:        "business_name_required",
				Type:        "required_field",
				Description: "Business name is required",
				Dimension:   business_intelligence.QualityDimensionCompleteness,
				Severity:    "error",
				Pattern:     "",
				Threshold:   1.0,
				Weight:      1.0,
				Enabled:     true,
			},
			{
				ID:          "rule_2",
				Name:        "revenue_positive",
				Type:        "range_validation",
				Description: "Revenue must be positive",
				Dimension:   business_intelligence.QualityDimensionValidity,
				Severity:    "error",
				Pattern:     "",
				Threshold:   0.0,
				Weight:      1.0,
				Enabled:     true,
			},
		})
		mockValidator.On("ValidateData", mock.Anything, mock.Anything).Return(&business_intelligence.ValidationResult{
			ID:           "validation_1",
			ValidatorID:  "test_validator",
			DataID:       "test_data_2",
			IsValid:      true,
			QualityScore: 0.92,
			Issues: []business_intelligence.QualityIssue{
				{
					ID:          "issue_2",
					Type:        "format_warning",
					Severity:    "warning",
					Description: "Industry code format could be improved",
					Field:       "industry_code",
					DetectedAt:  time.Now(),
				},
			},
			RulesApplied:   []string{"rule_1", "rule_2"},
			ValidationTime: 50 * time.Millisecond,
			ValidatedAt:    time.Now(),
		}, nil)
		mockValidator.On("GetValidationMetrics").Return(&business_intelligence.ValidationMetrics{
			ValidatorName:         "test_validator",
			TotalValidations:      1,
			ValidData:             1,
			InvalidData:           0,
			AverageValidationTime: 50 * time.Millisecond,
			LastValidation:        time.Now(),
		})

		// Test data validation
		ctx := context.Background()
		testData := map[string]interface{}{
			"business_name": "Valid Company",
			"industry":      "Technology",
			"revenue":       1000000,
			"employees":     50,
			"industry_code": "TECH001", // Format warning
		}

		result, err := qualitySystem.ValidateData(ctx, testData, "test_data_2")

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test_data_2", result.DataID)
		assert.True(t, result.IsValid)
		assert.Equal(t, 0.92, result.QualityScore)
		assert.NotEmpty(t, result.RulesApplied)
		assert.Less(t, result.ValidationTime, 100*time.Millisecond)

		// Verify validation rules were applied
		assert.Contains(t, result.RulesApplied, "rule_1")
		assert.Contains(t, result.RulesApplied, "rule_2")

		// Verify issues were detected
		assert.NotEmpty(t, result.Issues)
		issue := result.Issues[0]
		assert.Equal(t, "format_warning", issue.Type)
		assert.Equal(t, "warning", issue.Severity)

		mockValidator.AssertExpectations(t)
	})

	t.Run("Quality Assessment", func(t *testing.T) {
		// Reset mocks
		mockAssessor.ExpectedCalls = nil

		// Setup mock assessor
		mockAssessor.On("GetName").Return("test_assessor")
		mockAssessor.On("GetAssessmentTypes").Return([]business_intelligence.AssessmentType{
			business_intelligence.AssessmentTypeRealTime,
			business_intelligence.AssessmentTypeBatch,
		})
		mockAssessor.On("AssessQuality", mock.Anything, mock.Anything).Return(&business_intelligence.QualityScore{
			ID:             "score_1",
			AssessorID:     "test_assessor",
			AssessmentType: business_intelligence.AssessmentTypeRealTime,
			OverallScore:   0.88,
			QualityLevel:   "good",
			DimensionScores: map[business_intelligence.QualityDimension]float64{
				business_intelligence.QualityDimensionCompleteness: 0.90,
				business_intelligence.QualityDimensionAccuracy:     0.85,
				business_intelligence.QualityDimensionConsistency:  0.89,
			},
			Trend: "stable",
			Comparison: business_intelligence.QualityComparison{
				PreviousScore:    0.85,
				Change:           0.03,
				ChangePercentage: 3.5,
				Benchmark:        0.80,
				BenchmarkGap:     0.08,
			},
			CalculatedAt: time.Now(),
		}, nil)
		mockAssessor.On("GetAssessmentMetrics").Return(&business_intelligence.AssessmentMetrics{
			AssessorName:     "test_assessor",
			TotalAssessments: 1,
			AverageScore:     0.88,
			LastAssessment:   time.Now(),
		})

		// Create test assessments
		assessments := []*business_intelligence.QualityAssessment{
			{
				ID:           "assessment_1",
				MonitorID:    "test_monitor",
				DataID:       "test_data_1",
				OverallScore: 0.85,
				QualityLevel: "good",
				AssessedAt:   time.Now(),
			},
			{
				ID:           "assessment_2",
				MonitorID:    "test_monitor",
				DataID:       "test_data_2",
				OverallScore: 0.92,
				QualityLevel: "excellent",
				AssessedAt:   time.Now(),
			},
		}

		// Test quality assessment
		ctx := context.Background()
		score, err := qualitySystem.AssessQuality(ctx, assessments)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, score)
		assert.Equal(t, 0.88, score.OverallScore)
		assert.Equal(t, "good", score.QualityLevel)
		assert.Equal(t, "stable", score.Trend)

		// Verify dimension scores
		assert.Equal(t, 0.90, score.DimensionScores[business_intelligence.QualityDimensionCompleteness])
		assert.Equal(t, 0.85, score.DimensionScores[business_intelligence.QualityDimensionAccuracy])
		assert.Equal(t, 0.89, score.DimensionScores[business_intelligence.QualityDimensionConsistency])

		// Verify comparison data
		assert.Equal(t, 0.85, score.Comparison.PreviousScore)
		assert.Equal(t, 0.03, score.Comparison.Change)
		assert.Equal(t, 3.5, score.Comparison.ChangePercentage)

		mockAssessor.AssertExpectations(t)
	})

	t.Run("Quality Alerting", func(t *testing.T) {
		// Reset mocks
		mockMonitor.ExpectedCalls = nil
		mockAlerter.ExpectedCalls = nil

		// Setup mock monitor for low quality data
		mockMonitor.On("GetName").Return("test_monitor")
		mockMonitor.On("GetType").Return("completeness_monitor")
		mockMonitor.On("GetSupportedDimensions").Return([]business_intelligence.QualityDimension{
			business_intelligence.QualityDimensionCompleteness,
		})
		mockMonitor.On("MonitorData", mock.Anything, mock.Anything).Return(&business_intelligence.QualityAssessment{
			ID:           "assessment_low_quality",
			MonitorID:    "test_monitor",
			DataID:       "test_data_low_quality",
			DataType:     "business_data",
			OverallScore: 0.45, // Below alert threshold
			QualityLevel: "poor",
			Issues: []business_intelligence.QualityIssue{
				{
					ID:          "issue_critical",
					Type:        "missing_required_field",
					Severity:    "critical",
					Description: "Required field 'business_name' is missing",
					Field:       "business_name",
					DetectedAt:  time.Now(),
				},
				{
					ID:          "issue_error",
					Type:        "invalid_format",
					Severity:    "error",
					Description: "Revenue format is invalid",
					Field:       "revenue",
					DetectedAt:  time.Now(),
				},
			},
			Recommendations: []business_intelligence.QualityRecommendation{
				{
					ID:          "rec_critical",
					Type:        "data_correction",
					Priority:    "high",
					Description: "Fix missing business name",
					Action:      "Add business name to data",
					Impact:      "Critical for data completeness",
					Effort:      "Low",
					CreatedAt:   time.Now(),
				},
			},
			AssessedAt: time.Now(),
			ExpiresAt:  time.Now().Add(1 * time.Hour),
		}, nil)
		mockMonitor.On("GetMonitoringMetrics").Return(&business_intelligence.MonitoringMetrics{
			MonitorName:      "test_monitor",
			TotalAssessments: 1,
			AverageScore:     0.45,
			IssuesDetected:   2,
			LastAssessment:   time.Now(),
		})

		// Setup mock alerter
		mockAlerter.On("GetName").Return("test_alerter")
		mockAlerter.On("GetAlertTypes").Return([]business_intelligence.AlertType{
			business_intelligence.AlertTypeQualityDegradation,
		})
		mockAlerter.On("SendAlert", mock.Anything, mock.MatchedBy(func(alert *business_intelligence.QualityAlert) bool {
			return alert.Type == business_intelligence.AlertTypeQualityDegradation &&
				alert.Severity == "critical" &&
				alert.QualityScore == 0.45 &&
				alert.Threshold == 0.6
		})).Return(nil)
		mockAlerter.On("GetAlertingMetrics").Return(&business_intelligence.AlertingMetrics{
			AlerterName:  "test_alerter",
			TotalAlerts:  1,
			AlertsSent:   1,
			AlertsFailed: 0,
			LastAlert:    time.Now(),
		})

		// Test quality monitoring with alerting
		ctx := context.Background()
		testData := map[string]interface{}{
			"industry":  "Technology",
			"revenue":   "invalid_revenue",
			"employees": 50,
		}

		assessment, err := qualitySystem.MonitorDataQuality(ctx, testData, "test_data_low_quality")

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, assessment)
		assert.Equal(t, 0.45, assessment.OverallScore)
		assert.Equal(t, "poor", assessment.QualityLevel)
		assert.Len(t, assessment.Issues, 2)
		assert.Len(t, assessment.Recommendations, 1)

		// Verify critical issue
		criticalIssue := assessment.Issues[0]
		assert.Equal(t, "missing_required_field", criticalIssue.Type)
		assert.Equal(t, "critical", criticalIssue.Severity)

		// Verify high priority recommendation
		recommendation := assessment.Recommendations[0]
		assert.Equal(t, "data_correction", recommendation.Type)
		assert.Equal(t, "high", recommendation.Priority)

		mockMonitor.AssertExpectations(t)
		mockAlerter.AssertExpectations(t)
	})

	t.Run("Quality Reporting", func(t *testing.T) {
		// Reset mocks
		mockReporter.ExpectedCalls = nil

		// Setup mock reporter
		mockReporter.On("GetName").Return("test_reporter")
		mockReporter.On("GetReportTypes").Return([]business_intelligence.ReportType{
			business_intelligence.ReportTypeSummary,
			business_intelligence.ReportTypeDetailed,
		})
		mockReporter.On("GenerateReport", mock.Anything, mock.Anything).Return(&business_intelligence.QualityReport{
			ID:           "report_1",
			ReporterID:   "test_reporter",
			Type:         business_intelligence.ReportTypeSummary,
			Title:        "Data Quality Summary Report",
			Summary:      "Overall data quality is good with some areas for improvement",
			QualityScore: 0.85,
			QualityLevel: "good",
			Dimensions: map[business_intelligence.QualityDimension]*business_intelligence.DimensionScore{
				business_intelligence.QualityDimensionCompleteness: {
					Dimension:   business_intelligence.QualityDimensionCompleteness,
					Score:       0.90,
					Weight:      1.0,
					Confidence:  0.95,
					LastUpdated: time.Now(),
				},
				business_intelligence.QualityDimensionAccuracy: {
					Dimension:   business_intelligence.QualityDimensionAccuracy,
					Score:       0.80,
					Weight:      1.0,
					Confidence:  0.88,
					LastUpdated: time.Now(),
				},
			},
			Issues: []business_intelligence.QualityIssue{
				{
					ID:          "issue_report",
					Type:        "data_quality",
					Severity:    "medium",
					Description: "Some data fields have formatting issues",
					DetectedAt:  time.Now(),
				},
			},
			Recommendations: []business_intelligence.QualityRecommendation{
				{
					ID:          "rec_report",
					Type:        "process_improvement",
					Priority:    "medium",
					Description: "Implement data validation rules",
					Action:      "Add validation to data entry process",
					Impact:      "Improved data accuracy",
					Effort:      "Medium",
					CreatedAt:   time.Now(),
				},
			},
			GeneratedAt: time.Now(),
			Period: business_intelligence.TimePeriod{
				StartTime: time.Now().Add(-24 * time.Hour),
				EndTime:   time.Now(),
				Duration:  24 * time.Hour,
			},
		}, nil)
		mockReporter.On("GetReportingMetrics").Return(&business_intelligence.ReportingMetrics{
			ReporterName:     "test_reporter",
			TotalReports:     1,
			ReportsGenerated: 1,
			ReportsFailed:    0,
			LastReport:       time.Now(),
		})

		// Create report request
		request := &business_intelligence.ReportRequest{
			ID:         "request_1",
			ReporterID: "test_reporter",
			Type:       business_intelligence.ReportTypeSummary,
			DataIDs:    []string{"test_data_1", "test_data_2"},
			Dimensions: []business_intelligence.QualityDimension{
				business_intelligence.QualityDimensionCompleteness,
				business_intelligence.QualityDimensionAccuracy,
			},
			TimeRange: business_intelligence.TimePeriod{
				StartTime: time.Now().Add(-24 * time.Hour),
				EndTime:   time.Now(),
				Duration:  24 * time.Hour,
			},
			Format:      "json",
			RequestedAt: time.Now(),
		}

		// Test quality report generation
		ctx := context.Background()
		report, err := qualitySystem.GenerateQualityReport(ctx, request)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, "Data Quality Summary Report", report.Title)
		assert.Equal(t, 0.85, report.QualityScore)
		assert.Equal(t, "good", report.QualityLevel)
		assert.NotEmpty(t, report.Dimensions)
		assert.NotEmpty(t, report.Issues)
		assert.NotEmpty(t, report.Recommendations)

		// Verify dimension scores
		completenessScore := report.Dimensions[business_intelligence.QualityDimensionCompleteness]
		assert.Equal(t, 0.90, completenessScore.Score)
		assert.Equal(t, 0.95, completenessScore.Confidence)

		accuracyScore := report.Dimensions[business_intelligence.QualityDimensionAccuracy]
		assert.Equal(t, 0.80, accuracyScore.Score)
		assert.Equal(t, 0.88, accuracyScore.Confidence)

		// Verify report period
		assert.Equal(t, 24*time.Hour, report.Period.Duration)

		mockReporter.AssertExpectations(t)
	})
}

// TestDataQualityMonitoringSystem_ErrorHandling tests error handling in quality monitoring
func TestDataQualityMonitoringSystem_ErrorHandling(t *testing.T) {
	logger := zap.NewNop()

	// Configure quality monitoring system
	config := business_intelligence.QualityMonitoringConfig{
		EnableRealTimeMonitoring: true,
		EnableAlerting:           true,
		AlertThreshold:           0.6,
		EnableReporting:          true,
	}

	// Create quality monitoring system
	qualitySystem := business_intelligence.NewDataQualityMonitoringSystem(config, logger)

	// Create mock components
	mockMonitor := &MockQualityMonitor{}
	mockValidator := &MockQualityValidator{}
	mockAlerter := &MockQualityAlerter{}

	// Register mock components
	err := qualitySystem.RegisterMonitor(mockMonitor)
	require.NoError(t, err)

	err = qualitySystem.RegisterValidator(mockValidator)
	require.NoError(t, err)

	err = qualitySystem.RegisterAlerter(mockAlerter)
	require.NoError(t, err)

	t.Run("Monitor Error Handling", func(t *testing.T) {
		// Setup mock monitor to return error
		mockMonitor.On("GetName").Return("error_monitor")
		mockMonitor.On("GetType").Return("error_monitor")
		mockMonitor.On("GetSupportedDimensions").Return([]business_intelligence.QualityDimension{
			business_intelligence.QualityDimensionCompleteness,
		})
		mockMonitor.On("MonitorData", mock.Anything, mock.Anything).Return(
			(*business_intelligence.QualityAssessment)(nil), assert.AnError)
		mockMonitor.On("GetMonitoringMetrics").Return(&business_intelligence.MonitoringMetrics{
			MonitorName:      "error_monitor",
			TotalAssessments: 0,
			AverageScore:     0.0,
			IssuesDetected:   0,
			LastAssessment:   time.Time{},
		})

		// Test monitoring with error
		ctx := context.Background()
		testData := map[string]interface{}{
			"business_name": "Error Test Company",
		}

		assessment, err := qualitySystem.MonitorDataQuality(ctx, testData, "test_data_error")

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, assessment)
		assert.Contains(t, err.Error(), "monitoring failed")

		mockMonitor.AssertExpectations(t)
	})

	t.Run("Validator Error Handling", func(t *testing.T) {
		// Reset mocks
		mockValidator.ExpectedCalls = nil

		// Setup mock validator to return error
		mockValidator.On("GetName").Return("error_validator")
		mockValidator.On("GetValidationRules").Return([]business_intelligence.QualityRule{})
		mockValidator.On("ValidateData", mock.Anything, mock.Anything).Return(
			(*business_intelligence.ValidationResult)(nil), assert.AnError)
		mockValidator.On("GetValidationMetrics").Return(&business_intelligence.ValidationMetrics{
			ValidatorName:         "error_validator",
			TotalValidations:      0,
			ValidData:             0,
			InvalidData:           0,
			AverageValidationTime: 0,
			LastValidation:        time.Time{},
		})

		// Test validation with error
		ctx := context.Background()
		testData := map[string]interface{}{
			"business_name": "Error Test Company",
		}

		result, err := qualitySystem.ValidateData(ctx, testData, "test_data_error")

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "validation failed")

		mockValidator.AssertExpectations(t)
	})

	t.Run("Alerter Error Handling", func(t *testing.T) {
		// Reset mocks
		mockMonitor.ExpectedCalls = nil
		mockAlerter.ExpectedCalls = nil

		// Setup mock monitor for low quality data
		mockMonitor.On("GetName").Return("test_monitor")
		mockMonitor.On("GetType").Return("completeness_monitor")
		mockMonitor.On("GetSupportedDimensions").Return([]business_intelligence.QualityDimension{
			business_intelligence.QualityDimensionCompleteness,
		})
		mockMonitor.On("MonitorData", mock.Anything, mock.Anything).Return(&business_intelligence.QualityAssessment{
			ID:           "assessment_alert_error",
			MonitorID:    "test_monitor",
			DataID:       "test_data_alert_error",
			DataType:     "business_data",
			OverallScore: 0.45, // Below alert threshold
			QualityLevel: "poor",
			Issues: []business_intelligence.QualityIssue{
				{
					ID:          "issue_alert_error",
					Type:        "critical_issue",
					Severity:    "critical",
					Description: "Critical data quality issue",
					DetectedAt:  time.Now(),
				},
			},
			AssessedAt: time.Now(),
			ExpiresAt:  time.Now().Add(1 * time.Hour),
		}, nil)
		mockMonitor.On("GetMonitoringMetrics").Return(&business_intelligence.MonitoringMetrics{
			MonitorName:      "test_monitor",
			TotalAssessments: 1,
			AverageScore:     0.45,
			IssuesDetected:   1,
			LastAssessment:   time.Now(),
		})

		// Setup mock alerter to return error
		mockAlerter.On("GetName").Return("error_alerter")
		mockAlerter.On("GetAlertTypes").Return([]business_intelligence.AlertType{
			business_intelligence.AlertTypeQualityDegradation,
		})
		mockAlerter.On("SendAlert", mock.Anything, mock.Anything).Return(assert.AnError)
		mockAlerter.On("GetAlertingMetrics").Return(&business_intelligence.AlertingMetrics{
			AlerterName:  "error_alerter",
			TotalAlerts:  0,
			AlertsSent:   0,
			AlertsFailed: 1,
			LastAlert:    time.Time{},
		})

		// Test monitoring with alerting error
		ctx := context.Background()
		testData := map[string]interface{}{
			"business_name": "Alert Error Test Company",
		}

		assessment, err := qualitySystem.MonitorDataQuality(ctx, testData, "test_data_alert_error")

		// Assertions
		require.NoError(t, err) // Monitoring should succeed even if alerting fails
		assert.NotNil(t, assessment)
		assert.Equal(t, 0.45, assessment.OverallScore)

		mockMonitor.AssertExpectations(t)
		mockAlerter.AssertExpectations(t)
	})
}

// TestDataQualityMonitoringSystem_Configuration tests different configuration scenarios
func TestDataQualityMonitoringSystem_Configuration(t *testing.T) {
	logger := zap.NewNop()

	t.Run("Disabled Monitoring", func(t *testing.T) {
		// Configure with monitoring disabled
		config := business_intelligence.QualityMonitoringConfig{
			EnableRealTimeMonitoring: false,
			EnableBatchMonitoring:    false,
			EnableAlerting:           false,
			EnableReporting:          false,
		}

		// Create quality monitoring system
		qualitySystem := business_intelligence.NewDataQualityMonitoringSystem(config, logger)

		// Test monitoring when disabled
		ctx := context.Background()
		testData := map[string]interface{}{
			"business_name": "Disabled Test Company",
		}

		assessment, err := qualitySystem.MonitorDataQuality(ctx, testData, "test_data_disabled")

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, assessment)
		assert.Contains(t, err.Error(), "no suitable monitor found")
	})

	t.Run("High Threshold Configuration", func(t *testing.T) {
		// Configure with high quality thresholds
		config := business_intelligence.QualityMonitoringConfig{
			EnableRealTimeMonitoring: true,
			DefaultQualityThreshold:  0.95,
			CriticalQualityThreshold: 0.90,
			WarningQualityThreshold:  0.85,
			AlertThreshold:           0.80,
			EnableAlerting:           true,
		}

		// Create quality monitoring system
		qualitySystem := business_intelligence.NewDataQualityMonitoringSystem(config, logger)

		// Create mock monitor
		mockMonitor := &MockQualityMonitor{}
		mockMonitor.On("GetName").Return("high_threshold_monitor")
		mockMonitor.On("GetType").Return("completeness_monitor")
		mockMonitor.On("GetSupportedDimensions").Return([]business_intelligence.QualityDimension{
			business_intelligence.QualityDimensionCompleteness,
		})
		mockMonitor.On("MonitorData", mock.Anything, mock.Anything).Return(&business_intelligence.QualityAssessment{
			ID:           "assessment_high_threshold",
			MonitorID:    "high_threshold_monitor",
			DataID:       "test_data_high_threshold",
			DataType:     "business_data",
			OverallScore: 0.75, // Below high threshold
			QualityLevel: "fair",
			Issues: []business_intelligence.QualityIssue{
				{
					ID:          "issue_high_threshold",
					Type:        "quality_below_threshold",
					Severity:    "warning",
					Description: "Quality below high threshold",
					DetectedAt:  time.Now(),
				},
			},
			AssessedAt: time.Now(),
			ExpiresAt:  time.Now().Add(1 * time.Hour),
		}, nil)
		mockMonitor.On("GetMonitoringMetrics").Return(&business_intelligence.MonitoringMetrics{
			MonitorName:      "high_threshold_monitor",
			TotalAssessments: 1,
			AverageScore:     0.75,
			IssuesDetected:   1,
			LastAssessment:   time.Now(),
		})

		// Create mock alerter
		mockAlerter := &MockQualityAlerter{}
		mockAlerter.On("GetName").Return("high_threshold_alerter")
		mockAlerter.On("GetAlertTypes").Return([]business_intelligence.AlertType{
			business_intelligence.AlertTypeQualityDegradation,
		})
		mockAlerter.On("SendAlert", mock.Anything, mock.MatchedBy(func(alert *business_intelligence.QualityAlert) bool {
			return alert.Threshold == 0.80 // Should use configured threshold
		})).Return(nil)
		mockAlerter.On("GetAlertingMetrics").Return(&business_intelligence.AlertingMetrics{
			AlerterName:  "high_threshold_alerter",
			TotalAlerts:  1,
			AlertsSent:   1,
			AlertsFailed: 0,
			LastAlert:    time.Now(),
		})

		// Register components
		err := qualitySystem.RegisterMonitor(mockMonitor)
		require.NoError(t, err)

		err = qualitySystem.RegisterAlerter(mockAlerter)
		require.NoError(t, err)

		// Test monitoring with high thresholds
		ctx := context.Background()
		testData := map[string]interface{}{
			"business_name": "High Threshold Test Company",
		}

		assessment, err := qualitySystem.MonitorDataQuality(ctx, testData, "test_data_high_threshold")

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, assessment)
		assert.Equal(t, 0.75, assessment.OverallScore)

		mockMonitor.AssertExpectations(t)
		mockAlerter.AssertExpectations(t)
	})
}

// BenchmarkDataQualityMonitoringSystem benchmarks the quality monitoring system performance
func BenchmarkDataQualityMonitoringSystem(b *testing.B) {
	logger := zap.NewNop()

	// Configure quality monitoring system
	config := business_intelligence.QualityMonitoringConfig{
		EnableRealTimeMonitoring: true,
		EnableAlerting:           false, // Disable alerting for benchmarking
		EnableReporting:          false, // Disable reporting for benchmarking
		EnableParallelProcessing: true,
		MaxConcurrentMonitors:    4,
		ProcessingTimeout:        30 * time.Second,
	}

	// Create quality monitoring system
	qualitySystem := business_intelligence.NewDataQualityMonitoringSystem(config, logger)

	// Create mock monitor
	mockMonitor := &MockQualityMonitor{}
	mockMonitor.On("GetName").Return("benchmark_monitor")
	mockMonitor.On("GetType").Return("completeness_monitor")
	mockMonitor.On("GetSupportedDimensions").Return([]business_intelligence.QualityDimension{
		business_intelligence.QualityDimensionCompleteness,
		business_intelligence.QualityDimensionAccuracy,
	})
	mockMonitor.On("MonitorData", mock.Anything, mock.Anything).Return(&business_intelligence.QualityAssessment{
		ID:              "benchmark_assessment",
		MonitorID:       "benchmark_monitor",
		DataID:          "benchmark_data",
		DataType:        "business_data",
		OverallScore:    0.85,
		QualityLevel:    "good",
		Issues:          []business_intelligence.QualityIssue{},
		Recommendations: []business_intelligence.QualityRecommendation{},
		AssessedAt:      time.Now(),
		ExpiresAt:       time.Now().Add(1 * time.Hour),
	}, nil)
	mockMonitor.On("GetMonitoringMetrics").Return(&business_intelligence.MonitoringMetrics{
		MonitorName:      "benchmark_monitor",
		TotalAssessments: 1,
		AverageScore:     0.85,
		IssuesDetected:   0,
		LastAssessment:   time.Now(),
	})

	// Register monitor
	err := qualitySystem.RegisterMonitor(mockMonitor)
	if err != nil {
		b.Fatalf("Failed to register monitor: %v", err)
	}

	b.ResetTimer()

	// Benchmark quality monitoring
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		testData := map[string]interface{}{
			"business_name": fmt.Sprintf("Benchmark Company %d", i),
			"industry":      "Technology",
			"revenue":       1000000 + float64(i*100000),
			"employees":     50 + i*10,
		}

		_, err := qualitySystem.MonitorDataQuality(ctx, testData, fmt.Sprintf("benchmark_data_%d", i))
		if err != nil {
			b.Fatalf("Quality monitoring failed: %v", err)
		}
	}
}
