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
	"go.uber.org/zap"
)

func TestNewDataQualityHandler(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDataQualityHandler(logger)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
	assert.NotNil(t, handler.qualityChecks)
	assert.NotNil(t, handler.jobs)
	assert.NotNil(t, handler.reports)
	assert.Len(t, handler.qualityChecks, 0)
	assert.Len(t, handler.jobs, 0)
	assert.Len(t, handler.reports, 0)
}

func TestDataQualityHandler_CreateQualityCheck(t *testing.T) {
	tests := []struct {
		name           string
		request        DataQualityRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful quality check creation",
			request: DataQualityRequest{
				Name:        "Test Quality Check",
				Description: "Test data quality check",
				Dataset:     "customer_data",
				Checks: []QualityCheck{
					{
						Name:        "completeness_check",
						Type:        QualityCheckTypeCompleteness,
						Description: "Check for missing required fields",
						Severity:    QualitySeverityHigh,
						Parameters: map[string]interface{}{
							"required_fields": []string{"id", "name", "email"},
						},
						Rules: []QualityRule{
							{
								Name:        "not_null_check",
								Description: "Ensure required fields are not null",
								Expression:  "field IS NOT NULL",
								Parameters: map[string]interface{}{
									"field": "email",
								},
								Expected:  "non_null",
								Tolerance: 0.0,
							},
						},
						Conditions: []QualityCondition{
							{
								Name:        "email_format",
								Description: "Check email format",
								Operator:    "regex_match",
								Value:       "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
								Field:       "email",
								Function:    "regex",
							},
						},
						Actions: []QualityAction{
							{
								Name:        "log_issue",
								Type:        "log",
								Description: "Log quality issues",
								Parameters: map[string]interface{}{
									"level": "warning",
								},
								Condition: "score < 0.9",
								Priority:  1,
							},
						},
					},
				},
				Thresholds: QualityThresholds{
					OverallScore:   0.9,
					CriticalChecks: 1.0,
					HighChecks:     0.95,
					MediumChecks:   0.9,
					LowChecks:      0.8,
					PassRate:       0.95,
					FailRate:       0.05,
					WarningRate:    0.1,
				},
				Notifications: QualityNotifications{
					Email: []string{"admin@company.com"},
					Slack: []string{"#data-quality"},
					Conditions: map[string][]string{
						"critical": {"email", "slack"},
						"high":     {"email"},
					},
					Template: "quality_alert_template",
				},
				Metadata: map[string]interface{}{
					"department": "data_team",
					"priority":   "high",
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			request: DataQualityRequest{
				Description: "Test description",
				Dataset:     "customer_data",
				Checks: []QualityCheck{
					{
						Name:     "test_check",
						Type:     QualityCheckTypeCompleteness,
						Severity: QualitySeverityHigh,
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name is required",
		},
		{
			name: "missing dataset",
			request: DataQualityRequest{
				Name:        "Test Quality Check",
				Description: "Test description",
				Checks: []QualityCheck{
					{
						Name:     "test_check",
						Type:     QualityCheckTypeCompleteness,
						Severity: QualitySeverityHigh,
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "dataset is required",
		},
		{
			name: "missing checks",
			request: DataQualityRequest{
				Name:        "Test Quality Check",
				Description: "Test description",
				Dataset:     "customer_data",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "at least one quality check is required",
		},
		{
			name: "check missing name",
			request: DataQualityRequest{
				Name:        "Test Quality Check",
				Description: "Test description",
				Dataset:     "customer_data",
				Checks: []QualityCheck{
					{
						Type:     QualityCheckTypeCompleteness,
						Severity: QualitySeverityHigh,
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "check 1: name is required",
		},
		{
			name: "check missing type",
			request: DataQualityRequest{
				Name:        "Test Quality Check",
				Description: "Test description",
				Dataset:     "customer_data",
				Checks: []QualityCheck{
					{
						Name:     "test_check",
						Severity: QualitySeverityHigh,
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "check 1: type is required",
		},
		{
			name: "check missing severity",
			request: DataQualityRequest{
				Name:        "Test Quality Check",
				Description: "Test description",
				Dataset:     "customer_data",
				Checks: []QualityCheck{
					{
						Name: "test_check",
						Type: QualityCheckTypeCompleteness,
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "check 1: severity is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataQualityHandler(zap.NewNop())

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/quality", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.CreateQualityCheck(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var response DataQualityResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, tt.request.Name, response.Name)
				assert.Equal(t, "completed", response.Status)
				assert.Greater(t, response.OverallScore, 0.0)
				assert.Len(t, response.Checks, len(tt.request.Checks))
			}
		})
	}
}

func TestDataQualityHandler_GetQualityCheck(t *testing.T) {
	handler := NewDataQualityHandler(zap.NewNop())

	// Create a test quality check
	request := DataQualityRequest{
		Name:        "Test Quality Check",
		Description: "Test description",
		Dataset:     "customer_data",
		Checks: []QualityCheck{
			{
				Name:     "test_check",
				Type:     QualityCheckTypeCompleteness,
				Severity: QualitySeverityHigh,
			},
		},
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/quality", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.CreateQualityCheck(w, req)

	var response DataQualityResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	tests := []struct {
		name           string
		id             string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful retrieval",
			id:             response.ID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing id",
			id:             "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Quality check ID is required",
		},
		{
			name:           "non-existent id",
			id:             "non_existent_id",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Quality check not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/quality"
			if tt.id != "" {
				url = fmt.Sprintf("/quality?id=%s", tt.id)
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.GetQualityCheck(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var check DataQualityResponse
				err := json.Unmarshal(w.Body.Bytes(), &check)
				require.NoError(t, err)
				assert.Equal(t, response.ID, check.ID)
				assert.Equal(t, response.Name, check.Name)
			}
		})
	}
}

func TestDataQualityHandler_ListQualityChecks(t *testing.T) {
	handler := NewDataQualityHandler(zap.NewNop())

	// Create multiple test quality checks
	requests := []DataQualityRequest{
		{
			Name:        "Quality Check 1",
			Description: "First test check",
			Dataset:     "customer_data",
			Checks: []QualityCheck{
				{
					Name:     "check_1",
					Type:     QualityCheckTypeCompleteness,
					Severity: QualitySeverityHigh,
				},
			},
		},
		{
			Name:        "Quality Check 2",
			Description: "Second test check",
			Dataset:     "order_data",
			Checks: []QualityCheck{
				{
					Name:     "check_2",
					Type:     QualityCheckTypeAccuracy,
					Severity: QualitySeverityMedium,
				},
			},
		},
	}

	for _, req := range requests {
		body, _ := json.Marshal(req)
		request := httptest.NewRequest("POST", "/quality", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		handler.CreateQualityCheck(w, request)
	}

	// Test listing quality checks
	req := httptest.NewRequest("GET", "/quality", nil)
	w := httptest.NewRecorder()

	handler.ListQualityChecks(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "quality_checks")
	assert.Contains(t, response, "total")
	assert.Equal(t, float64(2), response["total"])
}

func TestDataQualityHandler_CreateQualityJob(t *testing.T) {
	tests := []struct {
		name           string
		request        DataQualityRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful job creation",
			request: DataQualityRequest{
				Name:        "Test Quality Job",
				Description: "Test background quality job",
				Dataset:     "customer_data",
				Checks: []QualityCheck{
					{
						Name:        "completeness_check",
						Type:        QualityCheckTypeCompleteness,
						Description: "Check for missing required fields",
						Severity:    QualitySeverityHigh,
						Parameters: map[string]interface{}{
							"required_fields": []string{"id", "name", "email"},
						},
					},
				},
				Thresholds: QualityThresholds{
					OverallScore:   0.9,
					CriticalChecks: 1.0,
					HighChecks:     0.95,
					MediumChecks:   0.9,
					LowChecks:      0.8,
					PassRate:       0.95,
					FailRate:       0.05,
					WarningRate:    0.1,
				},
				Notifications: QualityNotifications{
					Email: []string{"admin@company.com"},
					Conditions: map[string][]string{
						"critical": {"email"},
					},
					Template: "quality_job_template",
				},
				Metadata: map[string]interface{}{
					"priority": "high",
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request - missing name",
			request: DataQualityRequest{
				Description: "Test description",
				Dataset:     "customer_data",
				Checks: []QualityCheck{
					{
						Name:     "test_check",
						Type:     QualityCheckTypeCompleteness,
						Severity: QualitySeverityHigh,
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDataQualityHandler(zap.NewNop())

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/quality/jobs", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.CreateQualityJob(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var job QualityJob
				err := json.Unmarshal(w.Body.Bytes(), &job)
				require.NoError(t, err)
				assert.NotEmpty(t, job.ID)
				assert.Equal(t, "pending", job.Status)
				assert.Equal(t, 0, job.Progress)
				assert.NotEmpty(t, job.RequestID)
			}
		})
	}
}

func TestDataQualityHandler_GetQualityJob(t *testing.T) {
	handler := NewDataQualityHandler(zap.NewNop())

	// Create a test job
	request := DataQualityRequest{
		Name:        "Test Quality Job",
		Description: "Test description",
		Dataset:     "customer_data",
		Checks: []QualityCheck{
			{
				Name:     "test_check",
				Type:     QualityCheckTypeCompleteness,
				Severity: QualitySeverityHigh,
			},
		},
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/quality/jobs", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.CreateQualityJob(w, req)

	var job QualityJob
	json.Unmarshal(w.Body.Bytes(), &job)

	tests := []struct {
		name           string
		id             string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful retrieval",
			id:             job.ID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing id",
			id:             "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Job ID is required",
		},
		{
			name:           "non-existent id",
			id:             "non_existent_id",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Job not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/quality/jobs"
			if tt.id != "" {
				url = fmt.Sprintf("/quality/jobs?id=%s", tt.id)
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.GetQualityJob(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				var retrievedJob QualityJob
				err := json.Unmarshal(w.Body.Bytes(), &retrievedJob)
				require.NoError(t, err)
				assert.Equal(t, job.ID, retrievedJob.ID)
				assert.Equal(t, job.Status, retrievedJob.Status)
			}
		})
	}
}

func TestDataQualityHandler_ListQualityJobs(t *testing.T) {
	handler := NewDataQualityHandler(zap.NewNop())

	// Create multiple test jobs
	requests := []DataQualityRequest{
		{
			Name:        "Quality Job 1",
			Description: "First test job",
			Dataset:     "customer_data",
			Checks: []QualityCheck{
				{
					Name:     "check_1",
					Type:     QualityCheckTypeCompleteness,
					Severity: QualitySeverityHigh,
				},
			},
		},
		{
			Name:        "Quality Job 2",
			Description: "Second test job",
			Dataset:     "order_data",
			Checks: []QualityCheck{
				{
					Name:     "check_2",
					Type:     QualityCheckTypeAccuracy,
					Severity: QualitySeverityMedium,
				},
			},
		},
	}

	for _, req := range requests {
		body, _ := json.Marshal(req)
		request := httptest.NewRequest("POST", "/quality/jobs", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		handler.CreateQualityJob(w, request)
	}

	// Test listing jobs
	req := httptest.NewRequest("GET", "/quality/jobs", nil)
	w := httptest.NewRecorder()

	handler.ListQualityJobs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "jobs")
	assert.Contains(t, response, "total")
	assert.Equal(t, float64(2), response["total"])
}

func TestDataQualityHandler_ValidationLogic(t *testing.T) {
	handler := NewDataQualityHandler(zap.NewNop())

	tests := []struct {
		name        string
		request     DataQualityRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid request",
			request: DataQualityRequest{
				Name:        "Valid Check",
				Description: "Valid description",
				Dataset:     "test_dataset",
				Checks: []QualityCheck{
					{
						Name:     "valid_check",
						Type:     QualityCheckTypeCompleteness,
						Severity: QualitySeverityHigh,
					},
				},
			},
			expectError: false,
		},
		{
			name: "empty name",
			request: DataQualityRequest{
				Name:        "",
				Description: "Test description",
				Dataset:     "test_dataset",
				Checks: []QualityCheck{
					{
						Name:     "test_check",
						Type:     QualityCheckTypeCompleteness,
						Severity: QualitySeverityHigh,
					},
				},
			},
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "empty dataset",
			request: DataQualityRequest{
				Name:        "Test Check",
				Description: "Test description",
				Dataset:     "",
				Checks: []QualityCheck{
					{
						Name:     "test_check",
						Type:     QualityCheckTypeCompleteness,
						Severity: QualitySeverityHigh,
					},
				},
			},
			expectError: true,
			errorMsg:    "dataset is required",
		},
		{
			name: "empty checks",
			request: DataQualityRequest{
				Name:        "Test Check",
				Description: "Test description",
				Dataset:     "test_dataset",
				Checks:      []QualityCheck{},
			},
			expectError: true,
			errorMsg:    "at least one quality check is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateQualityRequest(tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataQualityHandler_QualityOperations(t *testing.T) {
	handler := NewDataQualityHandler(zap.NewNop())

	request := DataQualityRequest{
		Name:        "Test Quality Check",
		Description: "Test description",
		Dataset:     "customer_data",
		Checks: []QualityCheck{
			{
				Name:        "completeness_check",
				Type:        QualityCheckTypeCompleteness,
				Description: "Check for missing required fields",
				Severity:    QualitySeverityHigh,
			},
			{
				Name:        "accuracy_check",
				Type:        QualityCheckTypeAccuracy,
				Description: "Check for data accuracy",
				Severity:    QualitySeverityMedium,
			},
			{
				Name:        "consistency_check",
				Type:        QualityCheckTypeConsistency,
				Description: "Check for data consistency",
				Severity:    QualitySeverityLow,
			},
		},
	}

	t.Run("calculate overall score", func(t *testing.T) {
		score := handler.calculateOverallScore(request)
		assert.Greater(t, score, 0.0)
		assert.LessOrEqual(t, score, 1.0)
	})

	t.Run("perform quality checks", func(t *testing.T) {
		results := handler.performQualityChecks(request)
		assert.Len(t, results, 3)

		for _, result := range results {
			assert.NotEmpty(t, result.Name)
			assert.NotEmpty(t, result.Type)
			assert.NotEmpty(t, result.Status)
			assert.Greater(t, result.Score, 0.0)
			assert.LessOrEqual(t, result.Score, 1.0)
			assert.NotEmpty(t, result.Severity)
			assert.NotZero(t, result.ExecutionTime)
		}
	})

	t.Run("generate quality summary", func(t *testing.T) {
		summary := handler.generateQualitySummary(request)
		assert.Equal(t, 3, summary.TotalChecks)
		assert.GreaterOrEqual(t, summary.PassedChecks, 0)
		assert.GreaterOrEqual(t, summary.FailedChecks, 0)
		assert.GreaterOrEqual(t, summary.WarningChecks, 0)
		assert.GreaterOrEqual(t, summary.ErrorChecks, 0)
		assert.GreaterOrEqual(t, summary.PassRate, 0.0)
		assert.LessOrEqual(t, summary.PassRate, 1.0)
	})
}

func TestDataQualityHandler_EnumStringConversions(t *testing.T) {
	t.Run("QualityCheckType string conversion", func(t *testing.T) {
		checkType := QualityCheckTypeCompleteness
		assert.Equal(t, "completeness", checkType.String())

		checkType = QualityCheckTypeAccuracy
		assert.Equal(t, "accuracy", checkType.String())

		checkType = QualityCheckTypeConsistency
		assert.Equal(t, "consistency", checkType.String())
	})

	t.Run("QualityStatus string conversion", func(t *testing.T) {
		status := QualityStatusPassed
		assert.Equal(t, "passed", status.String())

		status = QualityStatusFailed
		assert.Equal(t, "failed", status.String())

		status = QualityStatusWarning
		assert.Equal(t, "warning", status.String())

		status = QualityStatusError
		assert.Equal(t, "error", status.String())
	})

	t.Run("QualitySeverity string conversion", func(t *testing.T) {
		severity := QualitySeverityLow
		assert.Equal(t, "low", severity.String())

		severity = QualitySeverityMedium
		assert.Equal(t, "medium", severity.String())

		severity = QualitySeverityHigh
		assert.Equal(t, "high", severity.String())

		severity = QualitySeverityCritical
		assert.Equal(t, "critical", severity.String())
	})
}

func TestDataQualityHandler_BackgroundJobProcessing(t *testing.T) {
	handler := NewDataQualityHandler(zap.NewNop())

	request := DataQualityRequest{
		Name:        "Test Background Job",
		Description: "Test background processing",
		Dataset:     "customer_data",
		Checks: []QualityCheck{
			{
				Name:     "test_check",
				Type:     QualityCheckTypeCompleteness,
				Severity: QualitySeverityHigh,
			},
		},
	}

	job := &QualityJob{
		ID:        "test_job_123",
		RequestID: "req_123",
		Status:    "pending",
		Progress:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test job processing
	go handler.processQualityJob(job, request)

	// Wait for job to complete
	time.Sleep(6 * time.Second)

	handler.mutex.RLock()
	defer handler.mutex.RUnlock()

	assert.Equal(t, "completed", job.Status)
	assert.Equal(t, 100, job.Progress)
	assert.NotNil(t, job.Result)
	assert.NotNil(t, job.CompletedAt)
}
