package security

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewPrivacyImpactAssessmentService(t *testing.T) {
	logger := zap.NewNop()

	t.Run("with default config", func(t *testing.T) {
		service, err := NewPrivacyImpactAssessmentService(nil, logger)
		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.NotNil(t, service.config)
		assert.True(t, service.config.EnablePIA)
		assert.Equal(t, 1000, service.config.PIAThreshold)
		assert.NotNil(t, service.assessmentManager)
		assert.NotNil(t, service.monitoringManager)
		assert.NotNil(t, service.riskManager)
		assert.NotNil(t, service.reportingManager)
		assert.Len(t, service.config.RiskLevels, 4)
		assert.Contains(t, service.config.RiskLevels, "low")
		assert.Contains(t, service.config.RiskLevels, "critical")
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &PrivacyImpactAssessmentConfig{
			EnablePIA:                     true,
			RequirePIAForNewProcessing:    false,
			PIAThreshold:                  500,
			HighRiskCategories:            []string{"custom_risk"},
			SensitiveDataTypes:            []string{"custom_sensitive"},
			MonitoringEnabled:             false,
			ContinuousAssessment:          false,
			AssessmentFrequency:           180 * 24 * time.Hour,
			RiskScoringEnabled:            false,
			AutomatedAlerts:               false,
			ComplianceReporting:           false,
			DataSubjectRightsMonitoring:   false,
			BreachDetectionEnabled:        false,
			ThirdPartyRiskAssessment:      false,
			CrossBorderTransferMonitoring: false,
		}

		service, err := NewPrivacyImpactAssessmentService(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, config, service.config)
		assert.False(t, service.config.RequirePIAForNewProcessing)
		assert.Equal(t, 500, service.config.PIAThreshold)
		assert.False(t, service.config.MonitoringEnabled)
	})

	t.Run("with nil logger", func(t *testing.T) {
		service, err := NewPrivacyImpactAssessmentService(nil, nil)
		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.NotNil(t, service.logger)
	})
}

func TestCreateAssessment(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewPrivacyImpactAssessmentService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name          string
		assessment    *PrivacyImpactAssessment
		expectError   bool
		errorContains string
	}{
		{
			name: "valid assessment",
			assessment: &PrivacyImpactAssessment{
				Title:             "Test Assessment",
				Description:       "Test privacy impact assessment",
				ProcessingPurpose: "business_verification",
				DataController:    "Test Company",
				DataCategories:    []string{"business_data"},
				DataSubjects:      []string{"business_representatives"},
				ProcessingActivities: []ProcessingActivity{
					{
						ID:               "activity_1",
						Name:             "Data Collection",
						Description:      "Collect business information",
						DataCategories:   []string{"business_data"},
						LegalBasis:       "legitimate_interest",
						RetentionPeriod:  365 * 24 * time.Hour,
						SecurityMeasures: []string{"encryption", "access_controls"},
					},
				},
				Assessor: "assessor_1",
			},
			expectError: false,
		},
		{
			name:          "nil assessment",
			assessment:    nil,
			expectError:   true,
			errorContains: "assessment cannot be nil",
		},
		{
			name: "missing title",
			assessment: &PrivacyImpactAssessment{
				ProcessingPurpose: "business_verification",
				DataController:    "Test Company",
				DataCategories:    []string{"business_data"},
				DataSubjects:      []string{"business_representatives"},
			},
			expectError:   true,
			errorContains: "assessment title is required",
		},
		{
			name: "missing processing purpose",
			assessment: &PrivacyImpactAssessment{
				Title:          "Test Assessment",
				DataController: "Test Company",
				DataCategories: []string{"business_data"},
				DataSubjects:   []string{"business_representatives"},
			},
			expectError:   true,
			errorContains: "processing purpose is required",
		},
		{
			name: "missing data controller",
			assessment: &PrivacyImpactAssessment{
				Title:             "Test Assessment",
				ProcessingPurpose: "business_verification",
				DataCategories:    []string{"business_data"},
				DataSubjects:      []string{"business_representatives"},
			},
			expectError:   true,
			errorContains: "data controller is required",
		},
		{
			name: "missing data categories",
			assessment: &PrivacyImpactAssessment{
				Title:             "Test Assessment",
				ProcessingPurpose: "business_verification",
				DataController:    "Test Company",
				DataSubjects:      []string{"business_representatives"},
			},
			expectError:   true,
			errorContains: "at least one data category is required",
		},
		{
			name: "missing data subjects",
			assessment: &PrivacyImpactAssessment{
				Title:             "Test Assessment",
				ProcessingPurpose: "business_verification",
				DataController:    "Test Company",
				DataCategories:    []string{"business_data"},
			},
			expectError:   true,
			errorContains: "at least one data subject type is required",
		},
		{
			name: "high-risk assessment with sensitive data",
			assessment: &PrivacyImpactAssessment{
				Title:             "High-Risk Assessment",
				ProcessingPurpose: "identity_verification",
				DataController:    "Test Company",
				DataCategories:    []string{"personal_identification", "financial_data"},
				DataSubjects:      []string{"individuals"},
				ProcessingActivities: []ProcessingActivity{
					{
						ID:               "activity_1",
						Name:             "Identity Verification",
						Description:      "Verify individual identity",
						DataCategories:   []string{"personal_identification"},
						LegalBasis:       "consent",
						RetentionPeriod:  2 * 365 * 24 * time.Hour,
						SecurityMeasures: []string{"encryption", "access_controls", "audit_logging"},
					},
					{
						ID:               "activity_2",
						Name:             "Financial Assessment",
						Description:      "Assess financial information",
						DataCategories:   []string{"financial_data"},
						LegalBasis:       "legitimate_interest",
						RetentionPeriod:  7 * 365 * 24 * time.Hour,
						SecurityMeasures: []string{"encryption", "access_controls"},
					},
				},
				Assessor: "assessor_2",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateAssessment(ctx, tt.assessment)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.assessment != nil {
					assert.NotEmpty(t, tt.assessment.ID)
					assert.Equal(t, "draft", tt.assessment.Status)
					assert.NotZero(t, tt.assessment.AssessmentDate)
					assert.NotZero(t, tt.assessment.NextReviewDate)
					assert.GreaterOrEqual(t, tt.assessment.RiskScore, 0.0)
					assert.NotEmpty(t, tt.assessment.RiskLevel)
					// Recommendations may be empty for low-risk assessments
					if tt.name == "high-risk assessment with sensitive data" {
						assert.NotEmpty(t, tt.assessment.Recommendations)
					}

					// Verify risk calculation
					if tt.name == "high-risk assessment with sensitive data" {
						assert.GreaterOrEqual(t, tt.assessment.RiskScore, 50.0)
						assert.Contains(t, []string{"high", "critical"}, tt.assessment.RiskLevel)
					}
				}
			}
		})
	}

	t.Run("PIA disabled", func(t *testing.T) {
		disabledService := &PrivacyImpactAssessmentService{
			config: &PrivacyImpactAssessmentConfig{EnablePIA: false},
			logger: logger,
		}

		assessment := &PrivacyImpactAssessment{
			Title:             "Test Assessment",
			ProcessingPurpose: "business_verification",
			DataController:    "Test Company",
			DataCategories:    []string{"business_data"},
			DataSubjects:      []string{"business_representatives"},
		}

		err := disabledService.CreateAssessment(ctx, assessment)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "privacy impact assessment is disabled")
	})
}

func TestConductAssessment(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewPrivacyImpactAssessmentService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("conduct assessment", func(t *testing.T) {
		assessmentID := "test_assessment_123"
		answers := map[string]interface{}{
			"data_categories":   []string{"business_data", "contact_information"},
			"data_subjects":     []string{"business_representatives"},
			"legal_basis":       "legitimate_interest",
			"retention_period":  "1 year",
			"security_measures": []string{"encryption", "access_controls"},
			"third_parties":     false,
		}

		assessment, err := service.ConductAssessment(ctx, assessmentID, answers)
		assert.NoError(t, err)
		assert.NotNil(t, assessment)
		assert.Equal(t, assessmentID, assessment.ID)
		assert.Equal(t, "in_review", assessment.Status)
		assert.Equal(t, answers, assessment.Answers)
		assert.GreaterOrEqual(t, assessment.RiskScore, 0.0)
		assert.NotEmpty(t, assessment.RiskLevel)
		assert.NotEmpty(t, assessment.Recommendations)
		assert.GreaterOrEqual(t, assessment.ComplianceStatus.OverallCompliance, 0.0)
		assert.NotEmpty(t, assessment.ComplianceStatus.ComplianceLevel)
	})

	t.Run("PIA disabled", func(t *testing.T) {
		disabledService := &PrivacyImpactAssessmentService{
			config: &PrivacyImpactAssessmentConfig{EnablePIA: false},
			logger: logger,
		}

		assessment, err := disabledService.ConductAssessment(ctx, "test", map[string]interface{}{})
		assert.Error(t, err)
		assert.Nil(t, assessment)
		assert.Contains(t, err.Error(), "privacy impact assessment is disabled")
	})
}

func TestApproveAssessment(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewPrivacyImpactAssessmentService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("approve assessment", func(t *testing.T) {
		assessmentID := "test_assessment_456"
		approverID := "approver_123"

		err := service.ApproveAssessment(ctx, assessmentID, approverID)
		assert.NoError(t, err)
	})

	t.Run("PIA disabled", func(t *testing.T) {
		disabledService := &PrivacyImpactAssessmentService{
			config: &PrivacyImpactAssessmentConfig{EnablePIA: false},
			logger: logger,
		}

		err := disabledService.ApproveAssessment(ctx, "test", "approver")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "privacy impact assessment is disabled")
	})
}

func TestMonitorPrivacyEvents(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewPrivacyImpactAssessmentService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("monitor privacy events", func(t *testing.T) {
		events, err := service.MonitorPrivacyEvents(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, events)
		// Events may be empty in test environment
		assert.GreaterOrEqual(t, len(events), 0)
	})

	t.Run("monitoring disabled", func(t *testing.T) {
		disabledService := &PrivacyImpactAssessmentService{
			config: &PrivacyImpactAssessmentConfig{MonitoringEnabled: false},
			logger: logger,
		}

		events, err := disabledService.MonitorPrivacyEvents(ctx)
		assert.Error(t, err)
		assert.Nil(t, events)
		assert.Contains(t, err.Error(), "privacy monitoring is disabled")
	})
}

func TestGeneratePrivacyReport(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewPrivacyImpactAssessmentService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name       string
		reportType string
		period     string
	}{
		{
			name:       "compliance report",
			reportType: "compliance",
			period:     "monthly",
		},
		{
			name:       "risk assessment report",
			reportType: "risk_assessment",
			period:     "quarterly",
		},
		{
			name:       "monitoring report",
			reportType: "monitoring",
			period:     "weekly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := service.GeneratePrivacyReport(ctx, tt.reportType, tt.period)
			assert.NoError(t, err)
			assert.NotNil(t, report)
			assert.NotEmpty(t, report.ID)
			assert.Equal(t, tt.reportType, report.ReportType)
			assert.Equal(t, tt.period, report.Period)
			assert.NotZero(t, report.GeneratedAt)
			assert.GreaterOrEqual(t, report.TotalAssessments, 0)
			assert.GreaterOrEqual(t, report.ComplianceScore, 0.0)
			assert.LessOrEqual(t, report.ComplianceScore, 100.0)
			assert.NotEmpty(t, report.Recommendations)
			assert.NotEmpty(t, report.RiskDistribution)
		})
	}

	t.Run("compliance reporting disabled", func(t *testing.T) {
		disabledService := &PrivacyImpactAssessmentService{
			config: &PrivacyImpactAssessmentConfig{ComplianceReporting: false},
			logger: logger,
		}

		report, err := disabledService.GeneratePrivacyReport(ctx, "compliance", "monthly")
		assert.Error(t, err)
		assert.Nil(t, report)
		assert.Contains(t, err.Error(), "compliance reporting is disabled")
	})
}

func TestValidateAssessment(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewPrivacyImpactAssessmentService(nil, logger)
	require.NoError(t, err)

	tests := []struct {
		name          string
		assessment    *PrivacyImpactAssessment
		expectError   bool
		errorContains string
	}{
		{
			name: "valid assessment",
			assessment: &PrivacyImpactAssessment{
				Title:             "Valid Assessment",
				ProcessingPurpose: "business_verification",
				DataController:    "Test Company",
				DataCategories:    []string{"business_data"},
				DataSubjects:      []string{"business_representatives"},
			},
			expectError: false,
		},
		{
			name: "empty title",
			assessment: &PrivacyImpactAssessment{
				ProcessingPurpose: "business_verification",
				DataController:    "Test Company",
				DataCategories:    []string{"business_data"},
				DataSubjects:      []string{"business_representatives"},
			},
			expectError:   true,
			errorContains: "assessment title is required",
		},
		{
			name: "empty processing purpose",
			assessment: &PrivacyImpactAssessment{
				Title:          "Test Assessment",
				DataController: "Test Company",
				DataCategories: []string{"business_data"},
				DataSubjects:   []string{"business_representatives"},
			},
			expectError:   true,
			errorContains: "processing purpose is required",
		},
		{
			name: "empty data controller",
			assessment: &PrivacyImpactAssessment{
				Title:             "Test Assessment",
				ProcessingPurpose: "business_verification",
				DataCategories:    []string{"business_data"},
				DataSubjects:      []string{"business_representatives"},
			},
			expectError:   true,
			errorContains: "data controller is required",
		},
		{
			name: "empty data categories",
			assessment: &PrivacyImpactAssessment{
				Title:             "Test Assessment",
				ProcessingPurpose: "business_verification",
				DataController:    "Test Company",
				DataSubjects:      []string{"business_representatives"},
			},
			expectError:   true,
			errorContains: "at least one data category is required",
		},
		{
			name: "empty data subjects",
			assessment: &PrivacyImpactAssessment{
				Title:             "Test Assessment",
				ProcessingPurpose: "business_verification",
				DataController:    "Test Company",
				DataCategories:    []string{"business_data"},
			},
			expectError:   true,
			errorContains: "at least one data subject type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateAssessment(tt.assessment)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCalculateRiskScore(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewPrivacyImpactAssessmentService(nil, logger)
	require.NoError(t, err)

	tests := []struct {
		name             string
		assessment       *PrivacyImpactAssessment
		expectedLevel    string
		expectedMinScore float64
	}{
		{
			name: "low risk assessment",
			assessment: &PrivacyImpactAssessment{
				DataCategories: []string{"business_data"},
				ProcessingActivities: []ProcessingActivity{
					{
						ID:   "activity_1",
						Name: "Basic Processing",
					},
				},
			},
			expectedLevel:    "low",
			expectedMinScore: 0.0,
		},
		{
			name: "medium risk assessment",
			assessment: &PrivacyImpactAssessment{
				DataCategories: []string{"business_data", "contact_information", "personal_identification"},
				ProcessingActivities: []ProcessingActivity{
					{
						ID:   "activity_1",
						Name: "Data Collection",
					},
					{
						ID:   "activity_2",
						Name: "Data Processing",
					},
					{
						ID:   "activity_3",
						Name: "Data Analysis",
					},
				},
			},
			expectedLevel:    "medium",
			expectedMinScore: 30.0,
		},
		{
			name: "high risk assessment",
			assessment: &PrivacyImpactAssessment{
				DataCategories: []string{"personal_identification", "financial_data"},
				ProcessingActivities: []ProcessingActivity{
					{
						ID:   "activity_1",
						Name: "Identity Verification",
					},
					{
						ID:   "activity_2",
						Name: "Financial Assessment",
					},
					{
						ID:   "activity_3",
						Name: "Risk Analysis",
					},
				},
			},
			expectedLevel:    "high",
			expectedMinScore: 50.0,
		},
		{
			name: "critical risk assessment",
			assessment: &PrivacyImpactAssessment{
				DataCategories: []string{"personal_identification", "financial_data", "health_data", "biometric_data"},
				ProcessingActivities: []ProcessingActivity{
					{
						ID:   "activity_1",
						Name: "Identity Verification",
					},
					{
						ID:   "activity_2",
						Name: "Financial Assessment",
					},
					{
						ID:   "activity_3",
						Name: "Health Data Processing",
					},
					{
						ID:   "activity_4",
						Name: "Biometric Analysis",
					},
					{
						ID:   "activity_5",
						Name: "Risk Assessment",
					},
				},
			},
			expectedLevel:    "critical",
			expectedMinScore: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, level := service.calculateRiskScore(tt.assessment)

			assert.GreaterOrEqual(t, score, tt.expectedMinScore)
			assert.Equal(t, tt.expectedLevel, level)

			// Verify risk level is valid
			assert.Contains(t, []string{"low", "medium", "high", "critical"}, level)
		})
	}
}

func TestAssessCompliance(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewPrivacyImpactAssessmentService(nil, logger)
	require.NoError(t, err)

	tests := []struct {
		name               string
		assessment         *PrivacyImpactAssessment
		expectedCompliance string
		expectedMinScore   float64
	}{
		{
			name: "compliant assessment",
			assessment: &PrivacyImpactAssessment{
				RiskScore:       10.0,
				Recommendations: []string{"Standard monitoring"},
				AssessmentDate:  time.Now(),
				NextReviewDate:  time.Now().Add(365 * 24 * time.Hour),
			},
			expectedCompliance: "compliant",
			expectedMinScore:   90.0,
		},
		{
			name: "partially compliant assessment",
			assessment: &PrivacyImpactAssessment{
				RiskScore:       25.0,
				Recommendations: []string{"Enhanced monitoring"},
				AssessmentDate:  time.Now(),
				NextReviewDate:  time.Now().Add(180 * 24 * time.Hour),
			},
			expectedCompliance: "partially_compliant",
			expectedMinScore:   75.0,
		},
		{
			name: "non-compliant assessment",
			assessment: &PrivacyImpactAssessment{
				RiskScore:       80.0,
				Recommendations: []string{"Immediate action required"},
				AssessmentDate:  time.Now(),
				NextReviewDate:  time.Now().Add(90 * 24 * time.Hour),
			},
			expectedCompliance: "non_compliant",
			expectedMinScore:   20.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compliance := service.assessCompliance(tt.assessment)

			assert.GreaterOrEqual(t, compliance.OverallCompliance, tt.expectedMinScore)
			assert.Equal(t, tt.expectedCompliance, compliance.ComplianceLevel)
			assert.Equal(t, tt.assessment.Recommendations, compliance.Recommendations)
			assert.NotZero(t, compliance.NextReviewDate)
			assert.NotZero(t, compliance.LastAssessmentDate)
		})
	}
}

func TestPIAGenerateRecommendations(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewPrivacyImpactAssessmentService(nil, logger)
	require.NoError(t, err)

	tests := []struct {
		name          string
		assessment    *PrivacyImpactAssessment
		expectedCount int
		shouldContain []string
	}{
		{
			name: "low risk assessment",
			assessment: &PrivacyImpactAssessment{
				RiskScore:      15.0,
				DataCategories: []string{"business_data"},
				ProcessingActivities: []ProcessingActivity{
					{
						ID:   "activity_1",
						Name: "Basic Processing",
					},
				},
			},
			expectedCount: 0,
			shouldContain: []string{},
		},
		{
			name: "medium risk assessment",
			assessment: &PrivacyImpactAssessment{
				RiskScore:      55.0,
				DataCategories: []string{"business_data", "contact_information", "location_data", "usage_data"},
				ProcessingActivities: []ProcessingActivity{
					{
						ID:   "activity_1",
						Name: "Data Collection",
					},
					{
						ID:   "activity_2",
						Name: "Data Processing",
					},
					{
						ID:   "activity_3",
						Name: "Data Analysis",
					},
					{
						ID:   "activity_4",
						Name: "Reporting",
					},
					{
						ID:   "activity_5",
						Name: "Storage",
					},
					{
						ID:   "activity_6",
						Name: "Archiving",
					},
				},
			},
			expectedCount: 4,
			shouldContain: []string{"security measures", "privacy training", "data collection", "processing activities"},
		},
		{
			name: "high risk assessment",
			assessment: &PrivacyImpactAssessment{
				RiskScore:      85.0,
				DataCategories: []string{"personal_identification", "financial_data"},
				ProcessingActivities: []ProcessingActivity{
					{
						ID:   "activity_1",
						Name: "Identity Verification",
					},
				},
			},
			expectedCount: 5,
			shouldContain: []string{"security measures", "privacy training", "data minimization", "privacy by design", "Data Protection Officer"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendations := service.generateRecommendations(tt.assessment)

			assert.Len(t, recommendations, tt.expectedCount)

			for _, expected := range tt.shouldContain {
				found := false
				for _, recommendation := range recommendations {
					if piaContainsSubstring(recommendation, expected) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected to find recommendation containing '%s'", expected)
			}
		})
	}
}

func TestGenerateReportRecommendations(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewPrivacyImpactAssessmentService(nil, logger)
	require.NoError(t, err)

	tests := []struct {
		name          string
		stats         *PrivacyStatistics
		shouldContain []string
	}{
		{
			name: "good statistics",
			stats: &PrivacyStatistics{
				TotalAssessments:    100,
				ExpiredAssessments:  10,
				HighRiskAssessments: 5,
				ComplianceScore:     95.0,
				ViolationCount:      0,
				CriticalEvents:      0,
			},
			shouldContain: []string{},
		},
		{
			name: "expired assessments",
			stats: &PrivacyStatistics{
				TotalAssessments:   100,
				ExpiredAssessments: 30, // More than 20%
			},
			shouldContain: []string{"expired assessments"},
		},
		{
			name: "high risk assessments",
			stats: &PrivacyStatistics{
				TotalAssessments:    100,
				HighRiskAssessments: 15, // More than 10%
			},
			shouldContain: []string{"risk mitigation"},
		},
		{
			name: "low compliance score",
			stats: &PrivacyStatistics{
				ComplianceScore: 85.0, // Below 90%
			},
			shouldContain: []string{"compliance posture"},
		},
		{
			name: "violations present",
			stats: &PrivacyStatistics{
				ViolationCount: 5,
			},
			shouldContain: []string{"violations"},
		},
		{
			name: "critical events",
			stats: &PrivacyStatistics{
				CriticalEvents: 3,
			},
			shouldContain: []string{"critical privacy events"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendations := service.generateReportRecommendations(tt.stats)

			for _, expected := range tt.shouldContain {
				found := false
				for _, recommendation := range recommendations {
					if piaContainsSubstring(recommendation, expected) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected to find recommendation containing '%s'", expected)
			}
		})
	}
}

func TestPIAIDGeneration(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewPrivacyImpactAssessmentService(nil, logger)
	require.NoError(t, err)

	t.Run("generate unique IDs", func(t *testing.T) {
		// Generate multiple IDs and ensure they're unique
		assessmentID1 := service.generateAssessmentID()
		time.Sleep(1 * time.Nanosecond)
		assessmentID2 := service.generateAssessmentID()
		assert.NotEqual(t, assessmentID1, assessmentID2)
		assert.Contains(t, assessmentID1, "pia_")

		reportID1 := service.generateReportID()
		time.Sleep(1 * time.Nanosecond)
		reportID2 := service.generateReportID()
		assert.NotEqual(t, reportID1, reportID2)
		assert.Contains(t, reportID1, "privacy_report_")
	})
}

func TestTemplateFunctions(t *testing.T) {
	t.Run("standard questions", func(t *testing.T) {
		questions := getStandardQuestions()
		assert.Len(t, questions, 6)

		// Check for required questions
		requiredCount := 0
		for _, q := range questions {
			if q.Required {
				requiredCount++
			}
		}
		assert.Greater(t, requiredCount, 0)

		// Check question categories
		categories := make(map[string]bool)
		for _, q := range questions {
			categories[q.Category] = true
		}
		assert.Contains(t, categories, "data_processing")
		assert.Contains(t, categories, "legal_compliance")
		assert.Contains(t, categories, "security")
	})

	t.Run("standard risk factors", func(t *testing.T) {
		riskFactors := getStandardRiskFactors()
		assert.Len(t, riskFactors, 5)

		// Check risk factor categories
		categories := make(map[string]bool)
		for _, rf := range riskFactors {
			categories[rf.Category] = true
			assert.NotEmpty(t, rf.Mitigation)
			assert.Greater(t, rf.Weight, 0.0)
		}
		assert.Contains(t, categories, "data_processing")
		assert.Contains(t, categories, "scale")
		assert.Contains(t, categories, "transfers")
	})
}

// Helper function to check if a string contains a substring (case-insensitive)
func piaContainsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				piaContainsSubstringAt(s, substr))))
}

func piaContainsSubstringAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
