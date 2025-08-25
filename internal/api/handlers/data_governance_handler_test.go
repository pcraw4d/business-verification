package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDataGovernanceHandler(t *testing.T) {
	handler := NewDataGovernanceHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.jobs)
}

func TestDataGovernanceHandler_CreateGovernanceFramework(t *testing.T) {
	handler := NewDataGovernanceHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful framework creation",
			requestBody: DataGovernanceRequest{
				FrameworkType: FrameworkTypeDataQuality,
				Policies: []GovernancePolicy{
					{
						ID:          "policy-1",
						Name:        "Data Quality Policy",
						Description: "Ensures data quality",
						Category:    "Quality",
						Version:     "1.0",
						Status:      FrameworkStatusActive,
						Owner:       "Data Team",
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
						Rules:       []PolicyRule{},
						Compliance:  []ComplianceStandard{StandardGDPR},
						RiskLevel:   RiskLevelLow,
						Tags:        []string{"quality"},
						Metadata:    make(map[string]interface{}),
					},
				},
				Controls: []GovernanceControl{
					{
						ID:            "control-1",
						Name:          "Data Validation",
						Description:   "Validates data",
						Type:          ControlTypePreventive,
						Category:      "Quality",
						Status:        "active",
						Priority:      1,
						Effectiveness: 0.90,
						Implementation: ImplementationInfo{
							Status:     "implemented",
							StartDate:  time.Now(),
							EndDate:    time.Now(),
							Owner:      "Data Team",
							Resources:  []string{"Engineers"},
							Cost:       50000.0,
							Timeline:   "2 months",
							Milestones: []Milestone{},
						},
						Monitoring: MonitoringConfig{
							Enabled:   true,
							Frequency: "daily",
							Metrics:   []string{"validation_rate"},
							Thresholds: map[string]float64{
								"error_rate": 0.05,
							},
							Alerts:  []Alert{},
							Reports: []string{"daily"},
						},
						Testing: TestingConfig{
							Enabled:   true,
							Frequency: "weekly",
							Method:    "automated",
							Scope:     "all data",
							TestCases: []TestCase{},
							Results:   []TestResult{},
						},
						Documentation: "Control documentation",
						Owner:         "Data Team",
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
					},
				},
				Compliance: []ComplianceRequirement{
					{
						ID:          "comp-1",
						Standard:    StandardGDPR,
						Requirement: "Data Protection",
						Description: "Protect data",
						Category:    "Privacy",
						Priority:    1,
						Status:      "compliant",
						Controls:    []string{"control-1"},
						Evidence:    []Evidence{},
						DueDate:     time.Now().AddDate(0, 1, 0),
						Owner:       "Compliance Team",
					},
				},
				RiskProfile: RiskProfile{
					OverallRisk: RiskLevelMedium,
					Categories: []RiskCategory{
						{
							Name:        "Data Privacy",
							Description: "Privacy risks",
							RiskLevel:   RiskLevelLow,
							Probability: 0.2,
							Impact:      0.3,
							Score:       0.06,
						},
					},
					Mitigations: []RiskMitigation{},
					Assessments: []RiskAssessment{},
					UpdatedAt:   time.Now(),
				},
				Scope: FrameworkScope{
					DataDomains:   []string{"customer"},
					BusinessUnits: []string{"sales"},
					Systems:       []string{"crm"},
					Processes:     []string{"data_ingestion"},
					Geographies:   []string{"US"},
					Timeframe:     "ongoing",
					Exceptions:    []string{},
				},
				Options: GovernanceOptions{
					AutoAssessment:  true,
					RiskScoring:     true,
					ComplianceCheck: true,
					ControlTesting:  true,
					Reporting:       true,
					Notifications:   true,
					AuditTrail:      true,
					VersionControl:  true,
				},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response DataGovernanceResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, "completed", response.Status)
				assert.NotNil(t, response.Framework)
				assert.NotNil(t, response.Summary)
				assert.NotNil(t, response.Statistics)
				assert.NotNil(t, response.Compliance)
				assert.NotNil(t, response.RiskAssessment)
				assert.NotEmpty(t, response.Controls)
				assert.NotEmpty(t, response.Policies)
			},
		},
		{
			name: "missing framework type",
			requestBody: DataGovernanceRequest{
				Policies: []GovernancePolicy{{ID: "policy-1"}},
				Controls: []GovernanceControl{{ID: "control-1"}},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Body.String(), "framework type is required")
			},
		},
		{
			name: "missing policies",
			requestBody: DataGovernanceRequest{
				FrameworkType: FrameworkTypeDataQuality,
				Controls:      []GovernanceControl{{ID: "control-1"}},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Body.String(), "at least one policy is required")
			},
		},
		{
			name: "missing controls",
			requestBody: DataGovernanceRequest{
				FrameworkType: FrameworkTypeDataQuality,
				Policies:      []GovernancePolicy{{ID: "policy-1"}},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Body.String(), "at least one control is required")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/governance", bytes.NewBuffer(body))
			recorder := httptest.NewRecorder()

			handler.CreateGovernanceFramework(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			tt.checkResponse(t, recorder)
		})
	}
}

func TestDataGovernanceHandler_GetGovernanceFramework(t *testing.T) {
	handler := NewDataGovernanceHandler()

	tests := []struct {
		name           string
		frameworkID    string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "successful framework retrieval",
			frameworkID:    "framework-123",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response DataGovernanceResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "framework-123", response.ID)
				assert.Equal(t, "retrieved", response.Status)
				assert.NotNil(t, response.Framework)
				assert.NotNil(t, response.Summary)
				assert.NotNil(t, response.Statistics)
				assert.NotNil(t, response.Compliance)
				assert.NotNil(t, response.RiskAssessment)
			},
		},
		{
			name:           "missing framework ID",
			frameworkID:    "",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Body.String(), "Framework ID is required")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/governance"
			if tt.frameworkID != "" {
				url += "?id=" + tt.frameworkID
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			recorder := httptest.NewRecorder()

			handler.GetGovernanceFramework(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			tt.checkResponse(t, recorder)
		})
	}
}

func TestDataGovernanceHandler_ListGovernanceFrameworks(t *testing.T) {
	handler := NewDataGovernanceHandler()

	req := httptest.NewRequest(http.MethodGet, "/governance", nil)
	recorder := httptest.NewRecorder()

	handler.ListGovernanceFrameworks(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "frameworks")
	assert.Contains(t, response, "total")
	assert.Contains(t, response, "timestamp")

	frameworks := response["frameworks"].([]interface{})
	assert.Len(t, frameworks, 3)
}

func TestDataGovernanceHandler_CreateGovernanceJob(t *testing.T) {
	handler := NewDataGovernanceHandler()

	requestBody := DataGovernanceRequest{
		FrameworkType: FrameworkTypeDataQuality,
		Policies: []GovernancePolicy{
			{
				ID:          "policy-1",
				Name:        "Data Quality Policy",
				Description: "Ensures data quality",
				Category:    "Quality",
				Version:     "1.0",
				Status:      FrameworkStatusActive,
				Owner:       "Data Team",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Rules:       []PolicyRule{},
				Compliance:  []ComplianceStandard{StandardGDPR},
				RiskLevel:   RiskLevelLow,
				Tags:        []string{"quality"},
				Metadata:    make(map[string]interface{}),
			},
		},
		Controls: []GovernanceControl{
			{
				ID:            "control-1",
				Name:          "Data Validation",
				Description:   "Validates data",
				Type:          ControlTypePreventive,
				Category:      "Quality",
				Status:        "active",
				Priority:      1,
				Effectiveness: 0.90,
				Implementation: ImplementationInfo{
					Status:     "implemented",
					StartDate:  time.Now(),
					EndDate:    time.Now(),
					Owner:      "Data Team",
					Resources:  []string{"Engineers"},
					Cost:       50000.0,
					Timeline:   "2 months",
					Milestones: []Milestone{},
				},
				Monitoring: MonitoringConfig{
					Enabled:   true,
					Frequency: "daily",
					Metrics:   []string{"validation_rate"},
					Thresholds: map[string]float64{
						"error_rate": 0.05,
					},
					Alerts:  []Alert{},
					Reports: []string{"daily"},
				},
				Testing: TestingConfig{
					Enabled:   true,
					Frequency: "weekly",
					Method:    "automated",
					Scope:     "all data",
					TestCases: []TestCase{},
					Results:   []TestResult{},
				},
				Documentation: "Control documentation",
				Owner:         "Data Team",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		},
		Compliance: []ComplianceRequirement{
			{
				ID:          "comp-1",
				Standard:    StandardGDPR,
				Requirement: "Data Protection",
				Description: "Protect data",
				Category:    "Privacy",
				Priority:    1,
				Status:      "compliant",
				Controls:    []string{"control-1"},
				Evidence:    []Evidence{},
				DueDate:     time.Now().AddDate(0, 1, 0),
				Owner:       "Compliance Team",
			},
		},
		RiskProfile: RiskProfile{
			OverallRisk: RiskLevelMedium,
			Categories: []RiskCategory{
				{
					Name:        "Data Privacy",
					Description: "Privacy risks",
					RiskLevel:   RiskLevelLow,
					Probability: 0.2,
					Impact:      0.3,
					Score:       0.06,
				},
			},
			Mitigations: []RiskMitigation{},
			Assessments: []RiskAssessment{},
			UpdatedAt:   time.Now(),
		},
		Scope: FrameworkScope{
			DataDomains:   []string{"customer"},
			BusinessUnits: []string{"sales"},
			Systems:       []string{"crm"},
			Processes:     []string{"data_ingestion"},
			Geographies:   []string{"US"},
			Timeframe:     "ongoing",
			Exceptions:    []string{},
		},
		Options: GovernanceOptions{
			AutoAssessment:  true,
			RiskScoring:     true,
			ComplianceCheck: true,
			ControlTesting:  true,
			Reporting:       true,
			Notifications:   true,
			AuditTrail:      true,
			VersionControl:  true,
		},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/governance/jobs", bytes.NewBuffer(body))
	recorder := httptest.NewRecorder()

	handler.CreateGovernanceJob(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "job_id")
	assert.Contains(t, response, "status")
	assert.Contains(t, response, "created_at")
	assert.Equal(t, "created", response["status"])

	jobID := response["job_id"].(string)
	assert.NotEmpty(t, jobID)
}

func TestDataGovernanceHandler_GetGovernanceJob(t *testing.T) {
	handler := NewDataGovernanceHandler()

	// Create a job first
	requestBody := DataGovernanceRequest{
		FrameworkType: FrameworkTypeDataQuality,
		Policies:      []GovernancePolicy{{ID: "policy-1"}},
		Controls:      []GovernanceControl{{ID: "control-1"}},
	}
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/governance/jobs", bytes.NewBuffer(body))
	recorder := httptest.NewRecorder()
	handler.CreateGovernanceJob(recorder, req)

	var createResponse map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &createResponse)
	jobID := createResponse["job_id"].(string)

	tests := []struct {
		name           string
		jobID          string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "successful job retrieval",
			jobID:          jobID,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var job GovernanceJob
				err := json.Unmarshal(recorder.Body.Bytes(), &job)
				require.NoError(t, err)
				assert.Equal(t, jobID, job.ID)
				assert.Equal(t, "governance_assessment", job.Type)
			},
		},
		{
			name:           "missing job ID",
			jobID:          "",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Body.String(), "Job ID is required")
			},
		},
		{
			name:           "non-existent job ID",
			jobID:          "non-existent",
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Body.String(), "Job not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/governance/jobs"
			if tt.jobID != "" {
				url += "?id=" + tt.jobID
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			recorder := httptest.NewRecorder()

			handler.GetGovernanceJob(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			tt.checkResponse(t, recorder)
		})
	}
}

func TestDataGovernanceHandler_ListGovernanceJobs(t *testing.T) {
	handler := NewDataGovernanceHandler()

	// Create some jobs first
	for i := 0; i < 3; i++ {
		requestBody := DataGovernanceRequest{
			FrameworkType: FrameworkTypeDataQuality,
			Policies:      []GovernancePolicy{{ID: fmt.Sprintf("policy-%d", i)}},
			Controls:      []GovernanceControl{{ID: fmt.Sprintf("control-%d", i)}},
		}
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/governance/jobs", bytes.NewBuffer(body))
		recorder := httptest.NewRecorder()
		handler.CreateGovernanceJob(recorder, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/governance/jobs", nil)
	recorder := httptest.NewRecorder()

	handler.ListGovernanceJobs(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "jobs")
	assert.Contains(t, response, "total")
	assert.Contains(t, response, "timestamp")

	jobs := response["jobs"].([]interface{})
	assert.Len(t, jobs, 3)
}

func TestDataGovernanceHandler_validateGovernanceRequest(t *testing.T) {
	handler := NewDataGovernanceHandler()

	tests := []struct {
		name    string
		request DataGovernanceRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: DataGovernanceRequest{
				FrameworkType: FrameworkTypeDataQuality,
				Policies:      []GovernancePolicy{{ID: "policy-1"}},
				Controls:      []GovernanceControl{{ID: "control-1"}},
			},
			wantErr: false,
		},
		{
			name: "missing framework type",
			request: DataGovernanceRequest{
				Policies: []GovernancePolicy{{ID: "policy-1"}},
				Controls: []GovernanceControl{{ID: "control-1"}},
			},
			wantErr: true,
			errMsg:  "framework type is required",
		},
		{
			name: "missing policies",
			request: DataGovernanceRequest{
				FrameworkType: FrameworkTypeDataQuality,
				Controls:      []GovernanceControl{{ID: "control-1"}},
			},
			wantErr: true,
			errMsg:  "at least one policy is required",
		},
		{
			name: "missing controls",
			request: DataGovernanceRequest{
				FrameworkType: FrameworkTypeDataQuality,
				Policies:      []GovernancePolicy{{ID: "policy-1"}},
			},
			wantErr: true,
			errMsg:  "at least one control is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateGovernanceRequest(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataGovernanceHandler_processGovernanceFramework(t *testing.T) {
	handler := NewDataGovernanceHandler()

	request := DataGovernanceRequest{
		FrameworkType: FrameworkTypeDataQuality,
		Policies: []GovernancePolicy{
			{
				ID:          "policy-1",
				Name:        "Data Quality Policy",
				Description: "Ensures data quality",
				Category:    "Quality",
				Version:     "1.0",
				Status:      FrameworkStatusActive,
				Owner:       "Data Team",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Rules:       []PolicyRule{},
				Compliance:  []ComplianceStandard{StandardGDPR},
				RiskLevel:   RiskLevelLow,
				Tags:        []string{"quality"},
				Metadata:    make(map[string]interface{}),
			},
		},
		Controls: []GovernanceControl{
			{
				ID:            "control-1",
				Name:          "Data Validation",
				Description:   "Validates data",
				Type:          ControlTypePreventive,
				Category:      "Quality",
				Status:        "active",
				Priority:      1,
				Effectiveness: 0.90,
				Implementation: ImplementationInfo{
					Status:     "implemented",
					StartDate:  time.Now(),
					EndDate:    time.Now(),
					Owner:      "Data Team",
					Resources:  []string{"Engineers"},
					Cost:       50000.0,
					Timeline:   "2 months",
					Milestones: []Milestone{},
				},
				Monitoring: MonitoringConfig{
					Enabled:   true,
					Frequency: "daily",
					Metrics:   []string{"validation_rate"},
					Thresholds: map[string]float64{
						"error_rate": 0.05,
					},
					Alerts:  []Alert{},
					Reports: []string{"daily"},
				},
				Testing: TestingConfig{
					Enabled:   true,
					Frequency: "weekly",
					Method:    "automated",
					Scope:     "all data",
					TestCases: []TestCase{},
					Results:   []TestResult{},
				},
				Documentation: "Control documentation",
				Owner:         "Data Team",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		},
		Compliance: []ComplianceRequirement{
			{
				ID:          "comp-1",
				Standard:    StandardGDPR,
				Requirement: "Data Protection",
				Description: "Protect data",
				Category:    "Privacy",
				Priority:    1,
				Status:      "compliant",
				Controls:    []string{"control-1"},
				Evidence:    []Evidence{},
				DueDate:     time.Now().AddDate(0, 1, 0),
				Owner:       "Compliance Team",
			},
		},
		RiskProfile: RiskProfile{
			OverallRisk: RiskLevelMedium,
			Categories: []RiskCategory{
				{
					Name:        "Data Privacy",
					Description: "Privacy risks",
					RiskLevel:   RiskLevelLow,
					Probability: 0.2,
					Impact:      0.3,
					Score:       0.06,
				},
			},
			Mitigations: []RiskMitigation{},
			Assessments: []RiskAssessment{},
			UpdatedAt:   time.Now(),
		},
		Scope: FrameworkScope{
			DataDomains:   []string{"customer"},
			BusinessUnits: []string{"sales"},
			Systems:       []string{"crm"},
			Processes:     []string{"data_ingestion"},
			Geographies:   []string{"US"},
			Timeframe:     "ongoing",
			Exceptions:    []string{},
		},
		Options: GovernanceOptions{
			AutoAssessment:  true,
			RiskScoring:     true,
			ComplianceCheck: true,
			ControlTesting:  true,
			Reporting:       true,
			Notifications:   true,
			AuditTrail:      true,
			VersionControl:  true,
		},
	}

	framework := handler.processGovernanceFramework(&request)

	assert.NotEmpty(t, framework.ID)
	assert.Equal(t, "Data Quality Governance Framework", framework.Name)
	assert.Equal(t, FrameworkTypeDataQuality, framework.Type)
	assert.Equal(t, FrameworkStatusActive, framework.Status)
	assert.Equal(t, "1.0.0", framework.Version)
	assert.Equal(t, "Data Governance Team", framework.Owner)
	assert.Len(t, framework.Policies, 1)
	assert.Len(t, framework.Controls, 1)
	assert.Len(t, framework.Compliance, 1)
	assert.NotNil(t, framework.RiskProfile)
	assert.NotNil(t, framework.Scope)
}

func TestDataGovernanceHandler_generateGovernanceSummary(t *testing.T) {
	handler := NewDataGovernanceHandler()

	framework := &GovernanceFramework{
		Policies: []GovernancePolicy{{ID: "policy-1"}, {ID: "policy-2"}},
		Controls: []GovernanceControl{{ID: "control-1"}, {ID: "control-2"}},
	}

	summary := handler.generateGovernanceSummary(framework)

	assert.Equal(t, 2, summary.TotalPolicies)
	assert.Equal(t, 2, summary.ActivePolicies)
	assert.Equal(t, 2, summary.TotalControls)
	assert.Equal(t, 2, summary.EffectiveControls)
	assert.Equal(t, 0.85, summary.ComplianceScore)
	assert.Equal(t, 0.25, summary.RiskScore)
	assert.Equal(t, 0.90, summary.Coverage)
	assert.NotZero(t, summary.LastAssessment)
}

func TestDataGovernanceHandler_assessCompliance(t *testing.T) {
	handler := NewDataGovernanceHandler()

	framework := &GovernanceFramework{
		Compliance: []ComplianceRequirement{
			{
				ID:       "comp-1",
				Standard: StandardGDPR,
			},
		},
	}

	compliance := handler.assessCompliance(framework)

	assert.Equal(t, 0.85, compliance.OverallScore)
	assert.Contains(t, compliance.Standards, "gdpr")
	assert.Equal(t, 0.95, compliance.Standards["gdpr"])
	assert.Len(t, compliance.Requirements, 1)
	assert.Empty(t, compliance.Violations)
	assert.NotZero(t, compliance.LastAudit)
	assert.NotZero(t, compliance.NextAudit)
}

func TestDataGovernanceHandler_assessRisk(t *testing.T) {
	handler := NewDataGovernanceHandler()

	framework := &GovernanceFramework{
		RiskProfile: RiskProfile{
			OverallRisk: RiskLevelMedium,
			Categories: []RiskCategory{
				{
					Name:        "Data Privacy",
					Description: "Privacy risks",
					RiskLevel:   RiskLevelLow,
					Probability: 0.2,
					Impact:      0.3,
					Score:       0.06,
				},
			},
		},
	}

	riskAssessment := handler.assessRisk(framework)

	assert.Equal(t, RiskLevelMedium, riskAssessment.OverallRisk)
	assert.Equal(t, 0.25, riskAssessment.RiskScore)
	assert.Len(t, riskAssessment.Categories, 1)
	assert.Len(t, riskAssessment.TopRisks, 1)
	assert.Empty(t, riskAssessment.Mitigations)
	assert.Empty(t, riskAssessment.Trends)
	assert.NotZero(t, riskAssessment.LastUpdated)
}

func TestDataGovernanceHandler_assessControls(t *testing.T) {
	handler := NewDataGovernanceHandler()

	controls := []GovernanceControl{
		{
			ID:            "control-1",
			Name:          "Data Validation",
			Type:          ControlTypePreventive,
			Status:        "active",
			Effectiveness: 0.90,
		},
		{
			ID:            "control-2",
			Name:          "Data Monitoring",
			Type:          ControlTypeDetective,
			Status:        "active",
			Effectiveness: 0.85,
		},
	}

	statuses := handler.assessControls(controls)

	assert.Len(t, statuses, 2)
	assert.Equal(t, "control-1", statuses[0].ID)
	assert.Equal(t, "Data Validation", statuses[0].Name)
	assert.Equal(t, "preventive", statuses[0].Type)
	assert.Equal(t, "active", statuses[0].Status)
	assert.Equal(t, 0.90, statuses[0].Effectiveness)
	assert.NotZero(t, statuses[0].LastTested)
	assert.NotZero(t, statuses[0].NextTest)
	assert.Empty(t, statuses[0].Issues)
}

func TestDataGovernanceHandler_assessPolicies(t *testing.T) {
	handler := NewDataGovernanceHandler()

	policies := []GovernancePolicy{
		{
			ID:     "policy-1",
			Name:   "Data Quality Policy",
			Status: FrameworkStatusActive,
		},
		{
			ID:     "policy-2",
			Name:   "Data Privacy Policy",
			Status: FrameworkStatusActive,
		},
	}

	statuses := handler.assessPolicies(policies)

	assert.Len(t, statuses, 2)
	assert.Equal(t, "policy-1", statuses[0].ID)
	assert.Equal(t, "Data Quality Policy", statuses[0].Name)
	assert.Equal(t, "active", statuses[0].Status)
	assert.Equal(t, 0.90, statuses[0].Compliance)
	assert.NotZero(t, statuses[0].LastReview)
	assert.NotZero(t, statuses[0].NextReview)
	assert.Equal(t, 0, statuses[0].Violations)
	assert.Equal(t, 0, statuses[0].Exceptions)
}

func TestDataGovernanceHandler_processGovernanceJob(t *testing.T) {
	handler := NewDataGovernanceHandler()

	request := DataGovernanceRequest{
		FrameworkType: FrameworkTypeDataQuality,
		Policies:      []GovernancePolicy{{ID: "policy-1"}},
		Controls:      []GovernanceControl{{ID: "control-1"}},
	}

	jobID := "test-job-123"
	job := &GovernanceJob{
		ID:        jobID,
		Type:      "governance_assessment",
		Status:    "pending",
		Progress:  0.0,
		CreatedAt: time.Now(),
	}

	handler.mu.Lock()
	handler.jobs[jobID] = job
	handler.mu.Unlock()

	// Start processing
	go handler.processGovernanceJob(jobID, &request)

	// Wait for processing to complete
	time.Sleep(600 * time.Millisecond)

	handler.mu.RLock()
	processedJob := handler.jobs[jobID]
	handler.mu.RUnlock()

	assert.Equal(t, "completed", processedJob.Status)
	assert.Equal(t, 1.0, processedJob.Progress)
	assert.NotNil(t, processedJob.Result)
	assert.NotZero(t, processedJob.CompletedAt)
}

// String conversion tests for enums
func TestGovernanceFrameworkType_String(t *testing.T) {
	tests := []struct {
		frameworkType GovernanceFrameworkType
		expected      string
	}{
		{FrameworkTypeDataQuality, "data_quality"},
		{FrameworkTypeDataPrivacy, "data_privacy"},
		{FrameworkTypeDataSecurity, "data_security"},
		{FrameworkTypeDataCompliance, "data_compliance"},
		{FrameworkTypeDataRetention, "data_retention"},
		{FrameworkTypeDataLineage, "data_lineage"},
	}

	for _, tt := range tests {
		t.Run(string(tt.frameworkType), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.frameworkType))
		})
	}
}

func TestGovernanceFrameworkStatus_String(t *testing.T) {
	tests := []struct {
		status   GovernanceFrameworkStatus
		expected string
	}{
		{FrameworkStatusDraft, "draft"},
		{FrameworkStatusActive, "active"},
		{FrameworkStatusSuspended, "suspended"},
		{FrameworkStatusDeprecated, "deprecated"},
		{FrameworkStatusArchived, "archived"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

func TestGovernanceControlType_String(t *testing.T) {
	tests := []struct {
		controlType GovernanceControlType
		expected    string
	}{
		{ControlTypePreventive, "preventive"},
		{ControlTypeDetective, "detective"},
		{ControlTypeCorrective, "corrective"},
		{ControlTypeCompensating, "compensating"},
		{ControlTypeDirective, "directive"},
	}

	for _, tt := range tests {
		t.Run(string(tt.controlType), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.controlType))
		})
	}
}

func TestComplianceStandard_String(t *testing.T) {
	tests := []struct {
		standard ComplianceStandard
		expected string
	}{
		{StandardGDPR, "gdpr"},
		{StandardCCPA, "ccpa"},
		{StandardSOX, "sox"},
		{StandardHIPAA, "hipaa"},
		{StandardPCI, "pci"},
		{StandardISO27001, "iso27001"},
	}

	for _, tt := range tests {
		t.Run(string(tt.standard), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.standard))
		})
	}
}

func TestRiskLevel_String(t *testing.T) {
	tests := []struct {
		riskLevel RiskLevel
		expected  string
	}{
		{RiskLevelLow, "low"},
		{RiskLevelMedium, "medium"},
		{RiskLevelHigh, "high"},
		{RiskLevelCritical, "critical"},
	}

	for _, tt := range tests {
		t.Run(string(tt.riskLevel), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.riskLevel))
		})
	}
}
