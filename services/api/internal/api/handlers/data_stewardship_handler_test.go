package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDataStewardshipHandler(t *testing.T) {
	handler := NewDataStewardshipHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.stewardships)
	assert.NotNil(t, handler.jobs)
}

func TestCreateStewardship(t *testing.T) {
	handler := NewDataStewardshipHandler()

	tests := []struct {
		name           string
		request        DataStewardshipRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid stewardship creation",
			request: DataStewardshipRequest{
				Type:   StewardshipTypeDataQuality,
				Domain: "customer_data",
				Stewards: []StewardAssignment{
					{
						UserID:      "user_123",
						Role:        StewardRoleOwner,
						Permissions: []string{"read", "write"},
						StartDate:   time.Now(),
						IsPrimary:   true,
						ContactInfo: ContactInfo{
							Email: "steward@example.com",
						},
					},
				},
				Responsibilities: []Responsibility{
					{
						ID:          "resp_001",
						Name:        "Data Quality Review",
						Description: "Review data quality metrics",
						Type:        "quality",
						Priority:    "high",
						Frequency:   "daily",
						DueDate:     time.Now().AddDate(0, 0, 7),
						AssignedTo:  "user_123",
					},
				},
				Workflows: []StewardshipWorkflowDefinition{
					{
						ID:          "workflow_001",
						Name:        "Quality Review Workflow",
						Description: "Automated quality review process",
						Status:      StewardshipWorkflowStatusActive,
						Version:     "1.0",
					},
				},
				Metrics: []MetricDefinition{
					{
						ID:          "metric_001",
						Name:        "Data Completeness",
						Description: "Percentage of complete records",
						Type:        "percentage",
						Formula:     "complete_records / total_records * 100",
						Unit:        "percentage",
						Threshold:   95.0,
						Frequency:   "daily",
					},
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing stewardship type",
			request: DataStewardshipRequest{
				Domain: "customer_data",
				Stewards: []StewardAssignment{
					{
						UserID:    "user_123",
						Role:      StewardRoleOwner,
						StartDate: time.Now(),
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "stewardship type is required",
		},
		{
			name: "missing domain",
			request: DataStewardshipRequest{
				Type: StewardshipTypeDataQuality,
				Stewards: []StewardAssignment{
					{
						UserID:    "user_123",
						Role:      StewardRoleOwner,
						StartDate: time.Now(),
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "domain is required",
		},
		{
			name: "missing stewards",
			request: DataStewardshipRequest{
				Type:   StewardshipTypeDataQuality,
				Domain: "customer_data",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "at least one steward is required",
		},
		{
			name: "missing steward user ID",
			request: DataStewardshipRequest{
				Type:   StewardshipTypeDataQuality,
				Domain: "customer_data",
				Stewards: []StewardAssignment{
					{
						Role:      StewardRoleOwner,
						StartDate: time.Now(),
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "user ID is required for steward 1",
		},
		{
			name: "missing steward role",
			request: DataStewardshipRequest{
				Type:   StewardshipTypeDataQuality,
				Domain: "customer_data",
				Stewards: []StewardAssignment{
					{
						UserID:    "user_123",
						StartDate: time.Now(),
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "role is required for steward 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/stewardship", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.CreateStewardship(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response DataStewardshipResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, tt.request.Type, response.Type)
				assert.Equal(t, tt.request.Domain, response.Domain)
				assert.Equal(t, StewardshipStatusActive, response.Status)
				assert.Len(t, response.Stewards, len(tt.request.Stewards))
				assert.Len(t, response.Responsibilities, len(tt.request.Responsibilities))
				assert.Len(t, response.Workflows, len(tt.request.Workflows))
				assert.Len(t, response.Metrics, len(tt.request.Metrics))
			}
		})
	}
}

func TestGetStewardship(t *testing.T) {
	handler := NewDataStewardshipHandler()

	// Create a stewardship first
	request := DataStewardshipRequest{
		Type:   StewardshipTypeDataQuality,
		Domain: "customer_data",
		Stewards: []StewardAssignment{
			{
				UserID:    "user_123",
				Role:      StewardRoleOwner,
				StartDate: time.Now(),
			},
		},
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/stewardship", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.CreateStewardship(w, req)

	var response DataStewardshipResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	stewardshipID := response.ID

	tests := []struct {
		name           string
		stewardshipID  string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "valid stewardship retrieval",
			stewardshipID:  stewardshipID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing stewardship ID",
			stewardshipID:  "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Stewardship ID is required",
		},
		{
			name:           "non-existent stewardship",
			stewardshipID:  "non_existent_id",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Stewardship not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/stewardship"
			if tt.stewardshipID != "" {
				url += "?id=" + tt.stewardshipID
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.GetStewardship(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response DataStewardshipResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, stewardshipID, response.ID)
			}
		})
	}
}

func TestListStewardships(t *testing.T) {
	handler := NewDataStewardshipHandler()

	// Create multiple stewardships
	requests := []DataStewardshipRequest{
		{
			Type:   StewardshipTypeDataQuality,
			Domain: "customer_data",
			Stewards: []StewardAssignment{
				{
					UserID:    "user_123",
					Role:      StewardRoleOwner,
					StartDate: time.Now(),
				},
			},
		},
		{
			Type:   StewardshipTypeDataGovernance,
			Domain: "financial_data",
			Stewards: []StewardAssignment{
				{
					UserID:    "user_456",
					Role:      StewardRoleCustodian,
					StartDate: time.Now(),
				},
			},
		},
	}

	for _, req := range requests {
		body, _ := json.Marshal(req)
		request := httptest.NewRequest("POST", "/stewardship", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		handler.CreateStewardship(w, request)
	}

	// Test listing stewardships
	req := httptest.NewRequest("GET", "/stewardship", nil)
	w := httptest.NewRecorder()

	handler.ListStewardships(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	stewardships := response["stewardships"].([]interface{})
	total := int(response["total"].(float64))

	assert.Len(t, stewardships, 2)
	assert.Equal(t, 2, total)
}

func TestCreateStewardshipJob(t *testing.T) {
	handler := NewDataStewardshipHandler()

	request := DataStewardshipRequest{
		Type:   StewardshipTypeDataQuality,
		Domain: "customer_data",
		Stewards: []StewardAssignment{
			{
				UserID:    "user_123",
				Role:      StewardRoleOwner,
				StartDate: time.Now(),
			},
		},
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/stewardship/jobs", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CreateStewardshipJob(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var job StewardshipJob
	err := json.Unmarshal(w.Body.Bytes(), &job)
	require.NoError(t, err)

	assert.NotEmpty(t, job.ID)
	assert.Equal(t, request.Type, job.Type)
	assert.Equal(t, "pending", job.Status)
	assert.Equal(t, 0.0, job.Progress)
	assert.NotNil(t, job.CreatedAt)
}

func TestGetStewardshipJob(t *testing.T) {
	handler := NewDataStewardshipHandler()

	// Create a job first
	request := DataStewardshipRequest{
		Type:   StewardshipTypeDataQuality,
		Domain: "customer_data",
		Stewards: []StewardAssignment{
			{
				UserID:    "user_123",
				Role:      StewardRoleOwner,
				StartDate: time.Now(),
			},
		},
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/stewardship/jobs", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.CreateStewardshipJob(w, req)

	var job StewardshipJob
	json.Unmarshal(w.Body.Bytes(), &job)
	jobID := job.ID

	tests := []struct {
		name           string
		jobID          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "valid job retrieval",
			jobID:          jobID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing job ID",
			jobID:          "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Job ID is required",
		},
		{
			name:           "non-existent job",
			jobID:          "non_existent_id",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Job not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/stewardship/jobs"
			if tt.jobID != "" {
				url += "?id=" + tt.jobID
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.GetStewardshipJob(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response StewardshipJob
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, jobID, response.ID)
			}
		})
	}
}

func TestListStewardshipJobs(t *testing.T) {
	handler := NewDataStewardshipHandler()

	// Create multiple jobs
	requests := []DataStewardshipRequest{
		{
			Type:   StewardshipTypeDataQuality,
			Domain: "customer_data",
			Stewards: []StewardAssignment{
				{
					UserID:    "user_123",
					Role:      StewardRoleOwner,
					StartDate: time.Now(),
				},
			},
		},
		{
			Type:   StewardshipTypeDataGovernance,
			Domain: "financial_data",
			Stewards: []StewardAssignment{
				{
					UserID:    "user_456",
					Role:      StewardRoleCustodian,
					StartDate: time.Now(),
				},
			},
		},
	}

	for _, req := range requests {
		body, _ := json.Marshal(req)
		request := httptest.NewRequest("POST", "/stewardship/jobs", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		handler.CreateStewardshipJob(w, request)
	}

	// Test listing jobs
	req := httptest.NewRequest("GET", "/stewardship/jobs", nil)
	w := httptest.NewRecorder()

	handler.ListStewardshipJobs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	jobs := response["jobs"].([]interface{})
	total := int(response["total"].(float64))

	assert.Len(t, jobs, 2)
	assert.Equal(t, 2, total)
}

func TestValidateStewardshipRequest(t *testing.T) {
	handler := NewDataStewardshipHandler()

	tests := []struct {
		name          string
		request       DataStewardshipRequest
		expectedError string
	}{
		{
			name: "valid request",
			request: DataStewardshipRequest{
				Type:   StewardshipTypeDataQuality,
				Domain: "customer_data",
				Stewards: []StewardAssignment{
					{
						UserID:    "user_123",
						Role:      StewardRoleOwner,
						StartDate: time.Now(),
					},
				},
			},
			expectedError: "",
		},
		{
			name: "missing type",
			request: DataStewardshipRequest{
				Domain: "customer_data",
				Stewards: []StewardAssignment{
					{
						UserID:    "user_123",
						Role:      StewardRoleOwner,
						StartDate: time.Now(),
					},
				},
			},
			expectedError: "stewardship type is required",
		},
		{
			name: "missing domain",
			request: DataStewardshipRequest{
				Type: StewardshipTypeDataQuality,
				Stewards: []StewardAssignment{
					{
						UserID:    "user_123",
						Role:      StewardRoleOwner,
						StartDate: time.Now(),
					},
				},
			},
			expectedError: "domain is required",
		},
		{
			name: "missing stewards",
			request: DataStewardshipRequest{
				Type:   StewardshipTypeDataQuality,
				Domain: "customer_data",
			},
			expectedError: "at least one steward is required",
		},
		{
			name: "missing steward user ID",
			request: DataStewardshipRequest{
				Type:   StewardshipTypeDataQuality,
				Domain: "customer_data",
				Stewards: []StewardAssignment{
					{
						Role:      StewardRoleOwner,
						StartDate: time.Now(),
					},
				},
			},
			expectedError: "user ID is required for steward 1",
		},
		{
			name: "missing steward role",
			request: DataStewardshipRequest{
				Type:   StewardshipTypeDataQuality,
				Domain: "customer_data",
				Stewards: []StewardAssignment{
					{
						UserID:    "user_123",
						StartDate: time.Now(),
					},
				},
			},
			expectedError: "role is required for steward 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateStewardshipRequest(&tt.request)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProcessStewards(t *testing.T) {
	handler := NewDataStewardshipHandler()

	assignments := []StewardAssignment{
		{
			UserID:      "user_123",
			Role:        StewardRoleOwner,
			Permissions: []string{"read", "write"},
			StartDate:   time.Now(),
			IsPrimary:   true,
			ContactInfo: ContactInfo{
				Email: "steward@example.com",
			},
		},
		{
			UserID:      "user_456",
			Role:        StewardRoleCustodian,
			Permissions: []string{"read"},
			StartDate:   time.Now(),
			IsPrimary:   false,
			ContactInfo: ContactInfo{
				Email: "custodian@example.com",
			},
		},
	}

	stewards := handler.processStewards(assignments)

	assert.Len(t, stewards, 2)
	assert.Equal(t, "user_123", stewards[0].UserID)
	assert.Equal(t, "Steward user_123", stewards[0].Name)
	assert.Equal(t, StewardRoleOwner, stewards[0].Role)
	assert.Equal(t, "active", stewards[0].Status)
	assert.True(t, stewards[0].IsPrimary)
	assert.Equal(t, "user_456", stewards[1].UserID)
	assert.Equal(t, "Steward user_456", stewards[1].Name)
	assert.Equal(t, StewardRoleCustodian, stewards[1].Role)
	assert.False(t, stewards[1].IsPrimary)
}

func TestProcessResponsibilities(t *testing.T) {
	handler := NewDataStewardshipHandler()

	responsibilities := []Responsibility{
		{
			ID:          "resp_001",
			Name:        "Data Quality Review",
			Description: "Review data quality metrics",
			Type:        "quality",
			Priority:    "high",
			Frequency:   "daily",
			DueDate:     time.Now().AddDate(0, 0, 7),
			AssignedTo:  "user_123",
		},
		{
			ID:          "resp_002",
			Name:        "Data Validation",
			Description: "Validate data accuracy",
			Type:        "validation",
			Priority:    "medium",
			Frequency:   "weekly",
			DueDate:     time.Now().AddDate(0, 0, 14),
			AssignedTo:  "user_456",
		},
	}

	statuses := handler.processResponsibilities(responsibilities)

	assert.Len(t, statuses, 2)
	assert.Equal(t, "resp_001", statuses[0].ID)
	assert.Equal(t, "Data Quality Review", statuses[0].Name)
	assert.Equal(t, "pending", statuses[0].Status)
	assert.Equal(t, 0.0, statuses[0].Progress)
	assert.Equal(t, "user_123", statuses[0].AssignedTo)
	assert.Equal(t, "resp_002", statuses[1].ID)
	assert.Equal(t, "Data Validation", statuses[1].Name)
	assert.Equal(t, "pending", statuses[1].Status)
	assert.Equal(t, 0.0, statuses[1].Progress)
	assert.Equal(t, "user_456", statuses[1].AssignedTo)
}

func TestProcessWorkflows(t *testing.T) {
	handler := NewDataStewardshipHandler()

	workflows := []StewardshipWorkflowDefinition{
		{
			ID:          "workflow_001",
			Name:        "Quality Review Workflow",
			Description: "Automated quality review process",
			Status:      StewardshipWorkflowStatusActive,
			Version:     "1.0",
		},
		{
			ID:          "workflow_002",
			Name:        "Validation Workflow",
			Description: "Data validation process",
			Status:      StewardshipWorkflowStatusDraft,
			Version:     "1.0",
		},
	}

	statuses := handler.processWorkflows(workflows)

	assert.Len(t, statuses, 2)
	assert.Equal(t, StewardshipWorkflowStatusActive, statuses[0])
	assert.Equal(t, StewardshipWorkflowStatusDraft, statuses[1])
}

func TestProcessMetrics(t *testing.T) {
	handler := NewDataStewardshipHandler()

	metrics := []MetricDefinition{
		{
			ID:          "metric_001",
			Name:        "Data Completeness",
			Description: "Percentage of complete records",
			Type:        "percentage",
			Formula:     "complete_records / total_records * 100",
			Unit:        "percentage",
			Threshold:   95.0,
			Frequency:   "daily",
		},
		{
			ID:          "metric_002",
			Name:        "Data Accuracy",
			Description: "Percentage of accurate records",
			Type:        "percentage",
			Formula:     "accurate_records / total_records * 100",
			Unit:        "percentage",
			Threshold:   98.0,
			Frequency:   "daily",
		},
	}

	statuses := handler.processMetrics(metrics)

	assert.Len(t, statuses, 2)
	assert.Equal(t, "metric_001", statuses[0].ID)
	assert.Equal(t, "Data Completeness", statuses[0].Name)
	assert.Equal(t, 0.0, statuses[0].CurrentValue)
	assert.Equal(t, 95.0, statuses[0].TargetValue)
	assert.Equal(t, "pending", statuses[0].Status)
	assert.Equal(t, "metric_002", statuses[1].ID)
	assert.Equal(t, "Data Accuracy", statuses[1].Name)
	assert.Equal(t, 0.0, statuses[1].CurrentValue)
	assert.Equal(t, 98.0, statuses[1].TargetValue)
	assert.Equal(t, "pending", statuses[1].Status)
}

func TestGenerateStewardshipSummary(t *testing.T) {
	handler := NewDataStewardshipHandler()

	stewards := []Steward{
		{
			UserID: "user_123",
			Name:   "Steward 1",
			Role:   StewardRoleOwner,
			Status: "active",
			Performance: Performance{
				TasksCompleted: 10,
				TasksOverdue:   2,
				QualityScore:   0.95,
			},
		},
		{
			UserID: "user_456",
			Name:   "Steward 2",
			Role:   StewardRoleCustodian,
			Status: "active",
			Performance: Performance{
				TasksCompleted: 8,
				TasksOverdue:   1,
				QualityScore:   0.90,
			},
		},
	}

	responsibilities := []ResponsibilityStatus{
		{
			ID:     "resp_001",
			Name:   "Task 1",
			Status: "completed",
		},
		{
			ID:     "resp_002",
			Name:   "Task 2",
			Status: "pending",
		},
	}

	summary := handler.generateStewardshipSummary(stewards, responsibilities)

	assert.Equal(t, 2, summary.TotalStewards)
	assert.Equal(t, 2, summary.ActiveStewards)
	assert.Equal(t, 2, summary.TotalResponsibilities)
	assert.Equal(t, 18, summary.CompletedTasks)
	assert.Equal(t, 3, summary.OverdueTasks)
	assert.Equal(t, 0.925, summary.AverageQuality)
	assert.Equal(t, 0.95, summary.ComplianceScore)
}

func TestGenerateStewardshipStatistics(t *testing.T) {
	handler := NewDataStewardshipHandler()

	stewards := []Steward{
		{
			UserID: "user_123",
			Name:   "Steward 1",
			Performance: Performance{
				TasksCompleted: 10,
				TasksOverdue:   2,
				QualityScore:   0.95,
				LastActivity:   time.Now(),
			},
		},
	}

	responsibilities := []ResponsibilityStatus{
		{
			ID:     "resp_001",
			Name:   "Task 1",
			Status: "completed",
		},
	}

	workflows := []StewardshipWorkflowStatus{
		StewardshipWorkflowStatusActive,
		StewardshipWorkflowStatusDraft,
	}

	metrics := []MetricStatus{
		{
			ID:           "metric_001",
			Name:         "Data Quality",
			CurrentValue: 95.0,
			TargetValue:  98.0,
		},
	}

	statistics := handler.generateStewardshipStatistics(stewards, responsibilities, workflows, metrics)

	assert.Len(t, statistics.StewardPerformance, 1)
	assert.Len(t, statistics.ResponsibilityTrends, 1)
	assert.Len(t, statistics.WorkflowMetrics, 2)
	assert.Len(t, statistics.QualityMetrics, 1)
	assert.Len(t, statistics.ComplianceMetrics, 1)

	assert.Equal(t, "user_123", statistics.StewardPerformance[0].UserID)
	assert.Equal(t, "Steward 1", statistics.StewardPerformance[0].Name)
	assert.Equal(t, 10, statistics.StewardPerformance[0].TasksCompleted)
	assert.Equal(t, 2, statistics.StewardPerformance[0].TasksOverdue)
	assert.Equal(t, 0.95, statistics.StewardPerformance[0].QualityScore)
}

func TestProcessStewardshipJob(t *testing.T) {
	handler := NewDataStewardshipHandler()

	request := DataStewardshipRequest{
		Type:   StewardshipTypeDataQuality,
		Domain: "customer_data",
		Stewards: []StewardAssignment{
			{
				UserID:    "user_123",
				Role:      StewardRoleOwner,
				StartDate: time.Now(),
			},
		},
	}

	job := &StewardshipJob{
		ID:        "test_job",
		Type:      StewardshipTypeDataQuality,
		Status:    "pending",
		Progress:  0.0,
		CreatedAt: time.Now(),
	}

	// Start the job processing
	go handler.processStewardshipJob(job, &request)

	// Wait a bit for processing to start
	time.Sleep(100 * time.Millisecond)

	// Check job status
	assert.Equal(t, "running", job.Status)
	assert.NotNil(t, job.StartedAt)
	assert.True(t, job.Progress > 0.0)

	// Wait for completion
	time.Sleep(4 * time.Second)

	// Check final status
	assert.Equal(t, "completed", job.Status)
	assert.Equal(t, 1.0, job.Progress)
	assert.NotNil(t, job.CompletedAt)
	assert.NotNil(t, job.Result)
	assert.NotEmpty(t, job.Result.StewardshipID)
}

func TestStringConversions(t *testing.T) {
	// Test StewardshipType string conversion
	stewardshipType := StewardshipTypeDataQuality
	assert.Equal(t, "data_quality", stewardshipType.String())

	// Test StewardshipStatus string conversion
	stewardshipStatus := StewardshipStatusActive
	assert.Equal(t, "active", stewardshipStatus.String())

	// Test StewardRole string conversion
	stewardRole := StewardRoleOwner
	assert.Equal(t, "owner", stewardRole.String())

	// Test DomainType string conversion
	domainType := DomainTypeBusiness
	assert.Equal(t, "business", domainType.String())

	// Test StewardshipWorkflowStatus string conversion
	workflowStatus := StewardshipWorkflowStatusActive
	assert.Equal(t, "active", workflowStatus.String())
}
