package lifecycle

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

func TestNewDataLifecycleHandler(t *testing.T) {
	handler := NewDataLifecycleHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.jobs)
}

func TestDataLifecycleHandler_CreateLifecycleInstance(t *testing.T) {
	handler := NewDataLifecycleHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful lifecycle instance creation",
			requestBody: DataLifecycleRequest{
				PolicyID: "policy-1",
				DataID:   "data-1",
				Stages: []LifecycleStage{
					{
						ID:          "stage-1",
						Name:        "Creation",
						Type:        StageTypeCreation,
						Description: "Data creation stage",
						Order:       1,
						Duration:    time.Minute * 5,
						Conditions:  []StageCondition{},
						Actions: []StageAction{
							{
								ID:          "action-1",
								Name:        "Data Validation",
								Type:        "validation",
								Description: "Validate data",
								Parameters:  make(map[string]interface{}),
								Enabled:     true,
								RetryPolicy: LifecycleRetryPolicy{
									MaxAttempts:       3,
									InitialDelay:      time.Second * 5,
									MaxDelay:          time.Minute,
									BackoffMultiplier: 2.0,
									RetryableErrors:   []string{"timeout", "network_error"},
								},
								Timeout: time.Minute,
							},
						},
						Triggers:  []StageTrigger{},
						Status:    LifecycleStatusActive,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				},
				RetentionPolicies: []LifecycleRetentionPolicy{
					{
						ID:          "retention-1",
						Name:        "Data Retention",
						Description: "Data retention policy",
						Type:        RetentionTypeTimeBased,
						Duration:    time.Hour * 24 * 365, // 1 year
						Conditions:  []RetentionCondition{},
						Actions:     []RetentionAction{},
						Exceptions:  []RetentionException{},
						Status:      LifecycleStatusActive,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				},
				Options: LifecycleOptions{
					AutoExecute:    true,
					ParallelStages: false,
					RetryFailed:    true,
					Notifications:  true,
					AuditTrail:     true,
					Monitoring:     true,
					Validation:     true,
				},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response DataLifecycleResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, "completed", response.Status)
				assert.NotNil(t, response.Instance)
				assert.NotNil(t, response.Summary)
				assert.NotNil(t, response.Statistics)
				assert.NotEmpty(t, response.Stages)
				assert.NotNil(t, response.Retention)
				assert.NotNil(t, response.Timeline)
			},
		},
		{
			name: "missing policy ID",
			requestBody: DataLifecycleRequest{
				DataID: "data-1",
				Stages: []LifecycleStage{{ID: "stage-1"}},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Body.String(), "policy ID is required")
			},
		},
		{
			name: "missing data ID",
			requestBody: DataLifecycleRequest{
				PolicyID: "policy-1",
				Stages:   []LifecycleStage{{ID: "stage-1"}},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Body.String(), "data ID is required")
			},
		},
		{
			name: "missing stages",
			requestBody: DataLifecycleRequest{
				PolicyID: "policy-1",
				DataID:   "data-1",
				Stages:   []LifecycleStage{},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Body.String(), "at least one stage is required")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/lifecycle", bytes.NewBuffer(body))
			recorder := httptest.NewRecorder()

			handler.CreateLifecycleInstance(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			tt.checkResponse(t, recorder)
		})
	}
}

func TestDataLifecycleHandler_GetLifecycleInstance(t *testing.T) {
	handler := NewDataLifecycleHandler()

	tests := []struct {
		name           string
		instanceID     string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "successful instance retrieval",
			instanceID:     "instance-123",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response DataLifecycleResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "instance-123", response.ID)
				assert.Equal(t, "retrieved", response.Status)
				assert.NotNil(t, response.Instance)
				assert.NotNil(t, response.Summary)
				assert.NotNil(t, response.Statistics)
				assert.NotEmpty(t, response.Stages)
				assert.NotNil(t, response.Retention)
				assert.NotNil(t, response.Timeline)
			},
		},
		{
			name:           "missing instance ID",
			instanceID:     "",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Body.String(), "Instance ID is required")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/lifecycle"
			if tt.instanceID != "" {
				url += "?id=" + tt.instanceID
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			recorder := httptest.NewRecorder()

			handler.GetLifecycleInstance(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			tt.checkResponse(t, recorder)
		})
	}
}

func TestDataLifecycleHandler_ListLifecycleInstances(t *testing.T) {
	handler := NewDataLifecycleHandler()

	req := httptest.NewRequest(http.MethodGet, "/lifecycle", nil)
	recorder := httptest.NewRecorder()

	handler.ListLifecycleInstances(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "instances")
	assert.Contains(t, response, "total")
	assert.Contains(t, response, "timestamp")

	instances := response["instances"].([]interface{})
	assert.Len(t, instances, 3)
}

func TestDataLifecycleHandler_CreateLifecycleJob(t *testing.T) {
	handler := NewDataLifecycleHandler()

	requestBody := DataLifecycleRequest{
		PolicyID: "policy-1",
		DataID:   "data-1",
		Stages: []LifecycleStage{
			{
				ID:          "stage-1",
				Name:        "Creation",
				Type:        StageTypeCreation,
				Description: "Data creation stage",
				Order:       1,
				Duration:    time.Minute * 5,
				Conditions:  []StageCondition{},
				Actions: []StageAction{
					{
						ID:          "action-1",
						Name:        "Data Validation",
						Type:        "validation",
						Description: "Validate data",
						Parameters:  make(map[string]interface{}),
						Enabled:     true,
						RetryPolicy: LifecycleRetryPolicy{
							MaxAttempts:       3,
							InitialDelay:      time.Second * 5,
							MaxDelay:          time.Minute,
							BackoffMultiplier: 2.0,
							RetryableErrors:   []string{"timeout", "network_error"},
						},
						Timeout: time.Minute,
					},
				},
				Triggers:  []StageTrigger{},
				Status:    LifecycleStatusActive,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		RetentionPolicies: []LifecycleRetentionPolicy{
			{
				ID:          "retention-1",
				Name:        "Data Retention",
				Description: "Data retention policy",
				Type:        RetentionTypeTimeBased,
				Duration:    time.Hour * 24 * 365, // 1 year
				Conditions:  []RetentionCondition{},
				Actions:     []RetentionAction{},
				Exceptions:  []RetentionException{},
				Status:      LifecycleStatusActive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
		Options: LifecycleOptions{
			AutoExecute:    true,
			ParallelStages: false,
			RetryFailed:    true,
			Notifications:  true,
			AuditTrail:     true,
			Monitoring:     true,
			Validation:     true,
		},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/lifecycle/jobs", bytes.NewBuffer(body))
	recorder := httptest.NewRecorder()

	handler.CreateLifecycleJob(recorder, req)

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

func TestDataLifecycleHandler_GetLifecycleJob(t *testing.T) {
	handler := NewDataLifecycleHandler()

	// Create a job first
	requestBody := DataLifecycleRequest{
		PolicyID: "policy-1",
		DataID:   "data-1",
		Stages:   []LifecycleStage{{ID: "stage-1"}},
	}
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/lifecycle/jobs", bytes.NewBuffer(body))
	recorder := httptest.NewRecorder()
	handler.CreateLifecycleJob(recorder, req)

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
				var job LifecycleJob
				err := json.Unmarshal(recorder.Body.Bytes(), &job)
				require.NoError(t, err)
				assert.Equal(t, jobID, job.ID)
				assert.Equal(t, "lifecycle_execution", job.Type)
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
			url := "/lifecycle/jobs"
			if tt.jobID != "" {
				url += "?id=" + tt.jobID
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			recorder := httptest.NewRecorder()

			handler.GetLifecycleJob(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			tt.checkResponse(t, recorder)
		})
	}
}

func TestDataLifecycleHandler_ListLifecycleJobs(t *testing.T) {
	handler := NewDataLifecycleHandler()

	// Create some jobs first
	for i := 0; i < 3; i++ {
		requestBody := DataLifecycleRequest{
			PolicyID: fmt.Sprintf("policy-%d", i),
			DataID:   fmt.Sprintf("data-%d", i),
			Stages:   []LifecycleStage{{ID: fmt.Sprintf("stage-%d", i)}},
		}
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/lifecycle/jobs", bytes.NewBuffer(body))
		recorder := httptest.NewRecorder()
		handler.CreateLifecycleJob(recorder, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/lifecycle/jobs", nil)
	recorder := httptest.NewRecorder()

	handler.ListLifecycleJobs(recorder, req)

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

func TestDataLifecycleHandler_validateLifecycleRequest(t *testing.T) {
	handler := NewDataLifecycleHandler()

	tests := []struct {
		name    string
		request DataLifecycleRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: DataLifecycleRequest{
				PolicyID: "policy-1",
				DataID:   "data-1",
				Stages:   []LifecycleStage{{ID: "stage-1"}},
			},
			wantErr: false,
		},
		{
			name: "missing policy ID",
			request: DataLifecycleRequest{
				DataID: "data-1",
				Stages: []LifecycleStage{{ID: "stage-1"}},
			},
			wantErr: true,
			errMsg:  "policy ID is required",
		},
		{
			name: "missing data ID",
			request: DataLifecycleRequest{
				PolicyID: "policy-1",
				Stages:   []LifecycleStage{{ID: "stage-1"}},
			},
			wantErr: true,
			errMsg:  "data ID is required",
		},
		{
			name: "missing stages",
			request: DataLifecycleRequest{
				PolicyID: "policy-1",
				DataID:   "data-1",
				Stages:   []LifecycleStage{},
			},
			wantErr: true,
			errMsg:  "at least one stage is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateLifecycleRequest(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataLifecycleHandler_processLifecycleInstance(t *testing.T) {
	handler := NewDataLifecycleHandler()

	request := DataLifecycleRequest{
		PolicyID: "policy-1",
		DataID:   "data-1",
		Stages: []LifecycleStage{
			{
				ID:          "stage-1",
				Name:        "Creation",
				Type:        StageTypeCreation,
				Description: "Data creation stage",
				Order:       1,
				Duration:    time.Minute * 5,
				Conditions:  []StageCondition{},
				Actions: []StageAction{
					{
						ID:          "action-1",
						Name:        "Data Validation",
						Type:        "validation",
						Description: "Validate data",
						Parameters:  make(map[string]interface{}),
						Enabled:     true,
						RetryPolicy: LifecycleRetryPolicy{
							MaxAttempts:       3,
							InitialDelay:      time.Second * 5,
							MaxDelay:          time.Minute,
							BackoffMultiplier: 2.0,
							RetryableErrors:   []string{"timeout", "network_error"},
						},
						Timeout: time.Minute,
					},
				},
				Triggers:  []StageTrigger{},
				Status:    LifecycleStatusActive,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		RetentionPolicies: []LifecycleRetentionPolicy{
			{
				ID:          "retention-1",
				Name:        "Data Retention",
				Description: "Data retention policy",
				Type:        RetentionTypeTimeBased,
				Duration:    time.Hour * 24 * 365, // 1 year
				Conditions:  []RetentionCondition{},
				Actions:     []RetentionAction{},
				Exceptions:  []RetentionException{},
				Status:      LifecycleStatusActive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
		Options: LifecycleOptions{
			AutoExecute:    true,
			ParallelStages: false,
			RetryFailed:    true,
			Notifications:  true,
			AuditTrail:     true,
			Monitoring:     true,
			Validation:     true,
		},
	}

	instance := handler.processLifecycleInstance(&request)

	assert.NotEmpty(t, instance.ID)
	assert.Equal(t, "policy-1", instance.PolicyID)
	assert.Equal(t, "data-1", instance.DataID)
	assert.Equal(t, LifecycleStatusActive, instance.Status)
	assert.NotEmpty(t, instance.CurrentStage)
	assert.Len(t, instance.Stages, 1)
	assert.NotNil(t, instance.Retention)
	assert.NotZero(t, instance.CreatedAt)
	assert.NotZero(t, instance.UpdatedAt)
}

func TestDataLifecycleHandler_generateLifecycleSummary(t *testing.T) {
	handler := NewDataLifecycleHandler()

	instance := &DataLifecycleInstance{
		Stages: []StageExecution{
			{
				Status: LifecycleStatusCompleted,
				Actions: []ActionExecution{
					{Status: LifecycleStatusCompleted},
					{Status: LifecycleStatusCompleted},
				},
			},
			{
				Status: LifecycleStatusActive,
				Actions: []ActionExecution{
					{Status: LifecycleStatusCompleted},
					{Status: LifecycleStatusFailed},
				},
			},
		},
	}

	summary := handler.generateLifecycleSummary(instance)

	assert.Equal(t, 2, summary.TotalStages)
	assert.Equal(t, 1, summary.CompletedStages)
	assert.Equal(t, 1, summary.ActiveStages)
	assert.Equal(t, 0, summary.FailedStages)
	assert.Equal(t, 4, summary.TotalActions)
	assert.Equal(t, 3, summary.CompletedActions)
	assert.Equal(t, 1, summary.FailedActions)
	assert.Equal(t, 0.5, summary.Progress)
	assert.NotZero(t, summary.EstimatedCompletion)
	assert.NotZero(t, summary.LastActivity)
}

func TestDataLifecycleHandler_assessStages(t *testing.T) {
	handler := NewDataLifecycleHandler()

	stages := []StageExecution{
		{
			StageID:     "stage-1",
			StageName:   "Creation",
			Status:      LifecycleStatusCompleted,
			StartedAt:   time.Now(),
			CompletedAt: time.Now().Add(time.Minute),
			Duration:    time.Minute,
			Actions: []ActionExecution{
				{
					ActionID:    "action-1",
					ActionName:  "Data Validation",
					Status:      LifecycleStatusCompleted,
					StartedAt:   time.Now(),
					CompletedAt: time.Now().Add(time.Second * 30),
					Duration:    time.Second * 30,
					Attempts:    1,
					Error:       "",
				},
			},
			Errors: []string{},
		},
	}

	statuses := handler.assessStages(stages)

	assert.Len(t, statuses, 1)
	assert.Equal(t, "stage-1", statuses[0].ID)
	assert.Equal(t, "Creation", statuses[0].Name)
	assert.Equal(t, "stage", statuses[0].Type)
	assert.Equal(t, "completed", statuses[0].Status)
	assert.Equal(t, 1.0, statuses[0].Progress)
	assert.NotZero(t, statuses[0].StartedAt)
	assert.NotZero(t, statuses[0].CompletedAt)
	assert.Equal(t, 60000.0, statuses[0].Duration)
	assert.Len(t, statuses[0].Actions, 1)
	assert.Empty(t, statuses[0].Errors)
}

func TestDataLifecycleHandler_assessRetention(t *testing.T) {
	handler := NewDataLifecycleHandler()

	retention := RetentionExecution{
		PolicyID:   "retention-1",
		Status:     LifecycleStatusActive,
		StartDate:  time.Now(),
		ExpiryDate: time.Now().AddDate(1, 0, 0),
		LastReview: time.Now(),
		NextReview: time.Now().AddDate(0, 1, 0),
		Actions:    []ActionExecution{},
		Exceptions: []RetentionException{},
	}

	status := handler.assessRetention(retention)

	assert.Equal(t, "retention-1", status.PolicyID)
	assert.Equal(t, "active", status.Status)
	assert.NotZero(t, status.StartDate)
	assert.NotZero(t, status.ExpiryDate)
	assert.Greater(t, status.DaysRemaining, 0)
	assert.NotZero(t, status.LastReview)
	assert.NotZero(t, status.NextReview)
	assert.Empty(t, status.Actions)
	assert.Empty(t, status.Exceptions)
}

func TestDataLifecycleHandler_processLifecycleJob(t *testing.T) {
	handler := NewDataLifecycleHandler()

	request := DataLifecycleRequest{
		PolicyID: "policy-1",
		DataID:   "data-1",
		Stages:   []LifecycleStage{{ID: "stage-1"}},
	}

	jobID := "test-job-123"
	job := &LifecycleJob{
		ID:        jobID,
		Type:      "lifecycle_execution",
		Status:    "pending",
		Progress:  0.0,
		CreatedAt: time.Now(),
	}

	handler.mu.Lock()
	handler.jobs[jobID] = job
	handler.mu.Unlock()

	// Start processing
	go handler.processLifecycleJob(jobID, &request)

	// Wait for processing to complete
	time.Sleep(1000 * time.Millisecond)

	handler.mu.RLock()
	processedJob := handler.jobs[jobID]
	handler.mu.RUnlock()

	assert.Equal(t, "completed", processedJob.Status)
	assert.Equal(t, 1.0, processedJob.Progress)
	assert.NotNil(t, processedJob.Result)
	assert.NotZero(t, processedJob.CompletedAt)
}

// String conversion tests for enums
func TestLifecycleStageType_String(t *testing.T) {
	tests := []struct {
		stageType LifecycleStageType
		expected  string
	}{
		{StageTypeCreation, "creation"},
		{StageTypeProcessing, "processing"},
		{StageTypeStorage, "storage"},
		{StageTypeArchival, "archival"},
		{StageTypeRetrieval, "retrieval"},
		{StageTypeDisposal, "disposal"},
	}

	for _, tt := range tests {
		t.Run(string(tt.stageType), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.stageType))
		})
	}
}

func TestLifecycleStatus_String(t *testing.T) {
	tests := []struct {
		status   LifecycleStatus
		expected string
	}{
		{LifecycleStatusActive, "active"},
		{LifecycleStatusInactive, "inactive"},
		{LifecycleStatusSuspended, "suspended"},
		{LifecycleStatusCompleted, "completed"},
		{LifecycleStatusFailed, "failed"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

func TestRetentionPolicyType_String(t *testing.T) {
	tests := []struct {
		policyType RetentionPolicyType
		expected   string
	}{
		{RetentionTypeTimeBased, "time_based"},
		{RetentionTypeEventBased, "event_based"},
		{RetentionTypeLegalHold, "legal_hold"},
		{RetentionTypeRegulatory, "regulatory"},
		{RetentionTypeBusiness, "business"},
	}

	for _, tt := range tests {
		t.Run(string(tt.policyType), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.policyType))
		})
	}
}

func TestDataClassification_String(t *testing.T) {
	tests := []struct {
		classification DataClassification
		expected       string
	}{
		{ClassificationPublic, "public"},
		{ClassificationInternal, "internal"},
		{ClassificationConfidential, "confidential"},
		{ClassificationRestricted, "restricted"},
		{ClassificationSecret, "secret"},
	}

	for _, tt := range tests {
		t.Run(string(tt.classification), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.classification))
		})
	}
}
