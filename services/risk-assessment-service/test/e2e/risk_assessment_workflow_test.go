//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/company/kyb-platform/services/risk-assessment-service/internal/models"
)

// TestCompleteRiskAssessmentWorkflow tests the complete risk assessment workflow
func TestCompleteRiskAssessmentWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end-to-end test")
	}

	// Setup test environment
	baseURL := "http://localhost:8080"
	client := &http.Client{Timeout: 30 * time.Second}
	logger := zap.NewNop()

	tests := []struct {
		name           string
		request        *models.RiskAssessmentRequest
		expectedStatus int
		validateResult func(*testing.T, *models.RiskAssessmentResponse)
	}{
		{
			name: "complete workflow - low risk business",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "Acme Corporation",
				BusinessAddress:   "123 Main Street, New York, NY 10001",
				Industry:          "technology",
				Country:           "US",
				PredictionHorizon: 3,
				ModelType:         "xgboost",
			},
			expectedStatus: http.StatusOK,
			validateResult: func(t *testing.T, resp *models.RiskAssessmentResponse) {
				assert.NotEmpty(t, resp.ID)
				assert.Equal(t, string(models.AssessmentStatusCompleted), resp.Status)
				assert.GreaterOrEqual(t, resp.RiskScore, 0.0)
				assert.LessOrEqual(t, resp.RiskScore, 1.0)
				assert.NotEmpty(t, resp.RiskFactors)
				assert.NotZero(t, resp.CreatedAt)
				assert.NotZero(t, resp.UpdatedAt)
			},
		},
		{
			name: "complete workflow - high risk business",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "High Risk Financial Corp",
				BusinessAddress:   "456 Risk Street, Miami, FL 33101",
				Industry:          "financial_services",
				Country:           "US",
				PredictionHorizon: 6,
				ModelType:         "lstm",
			},
			expectedStatus: http.StatusOK,
			validateResult: func(t *testing.T, resp *models.RiskAssessmentResponse) {
				assert.NotEmpty(t, resp.ID)
				assert.Equal(t, string(models.AssessmentStatusCompleted), resp.Status)
				assert.GreaterOrEqual(t, resp.RiskScore, 0.0)
				assert.LessOrEqual(t, resp.RiskScore, 1.0)
				assert.NotEmpty(t, resp.RiskFactors)
			},
		},
		{
			name: "complete workflow - international business",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "Global Manufacturing Ltd",
				BusinessAddress:   "789 Industrial Park, London, UK",
				Industry:          "manufacturing",
				Country:           "GB",
				PredictionHorizon: 12,
				ModelType:         "ensemble",
			},
			expectedStatus: http.StatusOK,
			validateResult: func(t *testing.T, resp *models.RiskAssessmentResponse) {
				assert.NotEmpty(t, resp.ID)
				assert.Equal(t, string(models.AssessmentStatusCompleted), resp.Status)
				assert.GreaterOrEqual(t, resp.RiskScore, 0.0)
				assert.LessOrEqual(t, resp.RiskScore, 1.0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Step 1: Submit risk assessment request
			assessmentID := submitRiskAssessment(t, client, baseURL, tt.request, tt.expectedStatus)
			require.NotEmpty(t, assessmentID)

			// Step 2: Wait for assessment to complete
			assessment := waitForAssessmentCompletion(t, client, baseURL, assessmentID, 30*time.Second)
			require.NotNil(t, assessment)

			// Step 3: Validate assessment result
			if tt.validateResult != nil {
				tt.validateResult(t, assessment)
			}

			// Step 4: Retrieve assessment details
			retrievedAssessment := getRiskAssessment(t, client, baseURL, assessmentID)
			require.NotNil(t, retrievedAssessment)
			assert.Equal(t, assessment.ID, retrievedAssessment.ID)

			// Step 5: Test scenario analysis
			scenarios := performScenarioAnalysis(t, client, baseURL, assessmentID)
			require.NotNil(t, scenarios)
			assert.NotEmpty(t, scenarios.Scenarios)

			// Step 6: Test risk factor explanation
			explanation := getRiskFactorExplanation(t, client, baseURL, assessmentID)
			require.NotNil(t, explanation)
			assert.NotEmpty(t, explanation.Factors)

			// Step 7: Test risk history
			history := getRiskHistory(t, client, baseURL, assessmentID)
			require.NotNil(t, history)
			assert.NotEmpty(t, history.History)

			logger.Info("Complete workflow test passed",
				zap.String("test_name", tt.name),
				zap.String("assessment_id", assessmentID))
		})
	}
}

// TestBatchRiskAssessmentWorkflow tests the batch risk assessment workflow
func TestBatchRiskAssessmentWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end-to-end test")
	}

	baseURL := "http://localhost:8080"
	client := &http.Client{Timeout: 60 * time.Second}

	// Create batch of risk assessment requests
	var requests []*models.RiskAssessmentRequest
	for i := 0; i < 10; i++ {
		requests = append(requests, &models.RiskAssessmentRequest{
			BusinessName:      fmt.Sprintf("Test Company %d", i),
			BusinessAddress:   fmt.Sprintf("%d Test Street, Test City, TC %05d", i, i),
			Industry:          "technology",
			Country:           "US",
			PredictionHorizon: 3,
			ModelType:         "xgboost",
		})
	}

	// Submit batch request
	batchID := submitBatchRiskAssessment(t, client, baseURL, requests)
	require.NotEmpty(t, batchID)

	// Wait for batch completion
	batchResult := waitForBatchCompletion(t, client, baseURL, batchID, 60*time.Second)
	require.NotNil(t, batchResult)

	// Validate batch results
	assert.Equal(t, len(requests), len(batchResult.Results))
	for i, result := range batchResult.Results {
		assert.NotEmpty(t, result.ID)
		assert.Equal(t, string(models.AssessmentStatusCompleted), result.Status)
		assert.GreaterOrEqual(t, result.RiskScore, 0.0)
		assert.LessOrEqual(t, result.RiskScore, 1.0)
		assert.NotEmpty(t, result.RiskFactors)
		assert.Equal(t, requests[i].BusinessName, result.BusinessName)
	}
}

// TestRiskAssessmentWithExternalData tests risk assessment with external data integration
func TestRiskAssessmentWithExternalData(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end-to-end test")
	}

	baseURL := "http://localhost:8080"
	client := &http.Client{Timeout: 30 * time.Second}

	request := &models.RiskAssessmentRequest{
		BusinessName:        "Test Company with External Data",
		BusinessAddress:     "123 External Data Street, Test City, TC 12345",
		Industry:            "financial_services",
		Country:             "US",
		PredictionHorizon:   3,
		ModelType:           "ensemble",
		IncludeExternalData: true,
	}

	// Submit risk assessment request
	assessmentID := submitRiskAssessment(t, client, baseURL, request, http.StatusOK)
	require.NotEmpty(t, assessmentID)

	// Wait for assessment to complete
	assessment := waitForAssessmentCompletion(t, client, baseURL, assessmentID, 30*time.Second)
	require.NotNil(t, assessment)

	// Validate external data integration
	assert.NotEmpty(t, assessment.ExternalData)
	assert.NotEmpty(t, assessment.ExternalData.ComplianceChecks)
	assert.NotEmpty(t, assessment.ExternalData.SanctionsChecks)
	assert.NotEmpty(t, assessment.ExternalData.AdverseMedia)
}

// TestRiskAssessmentErrorHandling tests error handling in risk assessment workflow
func TestRiskAssessmentErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end-to-end test")
	}

	baseURL := "http://localhost:8080"
	client := &http.Client{Timeout: 30 * time.Second}

	tests := []struct {
		name           string
		request        *models.RiskAssessmentRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "invalid request - missing business name",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "", // Invalid: empty name
				BusinessAddress:   "123 Test Street, Test City, TC 12345",
				Industry:          "technology",
				Country:           "US",
				PredictionHorizon: 3,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "business name is required",
		},
		{
			name: "invalid request - missing business address",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "Test Company",
				BusinessAddress:   "", // Invalid: empty address
				Industry:          "technology",
				Country:           "US",
				PredictionHorizon: 3,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "business address is required",
		},
		{
			name: "invalid request - invalid prediction horizon",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "Test Company",
				BusinessAddress:   "123 Test Street, Test City, TC 12345",
				Industry:          "technology",
				Country:           "US",
				PredictionHorizon: 0, // Invalid: zero horizon
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "prediction horizon must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Submit invalid request
			resp := submitRiskAssessmentRequest(t, client, baseURL, tt.request)
			defer resp.Body.Close()

			// Validate error response
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var errorResp map[string]interface{}
			err := json.NewDecoder(resp.Body).Decode(&errorResp)
			require.NoError(t, err)

			assert.Contains(t, errorResp, "error")
			if tt.expectedError != "" {
				assert.Contains(t, errorResp["error"], tt.expectedError)
			}
		})
	}
}

// TestRiskAssessmentPerformance tests performance of risk assessment workflow
func TestRiskAssessmentPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end-to-end test")
	}

	baseURL := "http://localhost:8080"
	client := &http.Client{Timeout: 30 * time.Second}

	request := &models.RiskAssessmentRequest{
		BusinessName:      "Performance Test Company",
		BusinessAddress:   "123 Performance Street, Test City, TC 12345",
		Industry:          "technology",
		Country:           "US",
		PredictionHorizon: 3,
		ModelType:         "xgboost",
	}

	// Test single assessment performance
	t.Run("single_assessment_performance", func(t *testing.T) {
		start := time.Now()
		assessmentID := submitRiskAssessment(t, client, baseURL, request, http.StatusOK)
		duration := time.Since(start)

		require.NotEmpty(t, assessmentID)
		assert.Less(t, duration, 5*time.Second, "Single assessment should complete within 5 seconds")

		// Wait for completion and measure total time
		start = time.Now()
		assessment := waitForAssessmentCompletion(t, client, baseURL, assessmentID, 30*time.Second)
		totalDuration := time.Since(start)

		require.NotNil(t, assessment)
		assert.Less(t, totalDuration, 10*time.Second, "Total assessment should complete within 10 seconds")
	})

	// Test concurrent assessments performance
	t.Run("concurrent_assessments_performance", func(t *testing.T) {
		const numConcurrent = 10
		results := make(chan time.Duration, numConcurrent)

		// Submit concurrent requests
		for i := 0; i < numConcurrent; i++ {
			go func(i int) {
				start := time.Now()
				assessmentID := submitRiskAssessment(t, client, baseURL, &models.RiskAssessmentRequest{
					BusinessName:      fmt.Sprintf("Concurrent Test Company %d", i),
					BusinessAddress:   fmt.Sprintf("%d Concurrent Street, Test City, TC %05d", i, i),
					Industry:          "technology",
					Country:           "US",
					PredictionHorizon: 3,
					ModelType:         "xgboost",
				}, http.StatusOK)
				duration := time.Since(start)

				if assessmentID != "" {
					results <- duration
				} else {
					results <- 30 * time.Second // Max timeout
				}
			}(i)
		}

		// Collect results
		var totalDuration time.Duration
		for i := 0; i < numConcurrent; i++ {
			duration := <-results
			totalDuration += duration
		}

		avgDuration := totalDuration / numConcurrent
		assert.Less(t, avgDuration, 5*time.Second, "Average concurrent assessment should complete within 5 seconds")
	})
}

// Helper functions

// submitRiskAssessment submits a risk assessment request and returns the assessment ID
func submitRiskAssessment(t *testing.T, client *http.Client, baseURL string, request *models.RiskAssessmentRequest, expectedStatus int) string {
	resp := submitRiskAssessmentRequest(t, client, baseURL, request)
	defer resp.Body.Close()

	assert.Equal(t, expectedStatus, resp.StatusCode)

	if expectedStatus == http.StatusOK {
		var response models.RiskAssessmentResponse
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		return response.ID
	}

	return ""
}

// submitRiskAssessmentRequest submits a risk assessment request and returns the response
func submitRiskAssessmentRequest(t *testing.T, client *http.Client, baseURL string, request *models.RiskAssessmentRequest) *http.Response {
	reqBody, err := json.Marshal(request)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/api/v1/assess", bytes.NewBuffer(reqBody))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}

// waitForAssessmentCompletion waits for an assessment to complete
func waitForAssessmentCompletion(t *testing.T, client *http.Client, baseURL string, assessmentID string, timeout time.Duration) *models.RiskAssessmentResponse {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		assessment := getRiskAssessment(t, client, baseURL, assessmentID)
		if assessment != nil && assessment.Status == string(models.AssessmentStatusCompleted) {
			return assessment
		}
		time.Sleep(1 * time.Second)
	}

	t.Fatalf("Assessment %s did not complete within %v", assessmentID, timeout)
	return nil
}

// getRiskAssessment retrieves a risk assessment by ID
func getRiskAssessment(t *testing.T, client *http.Client, baseURL string, assessmentID string) *models.RiskAssessmentResponse {
	req, err := http.NewRequest("GET", baseURL+"/api/v1/assess/"+assessmentID, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var assessment models.RiskAssessmentResponse
		err := json.NewDecoder(resp.Body).Decode(&assessment)
		require.NoError(t, err)
		return &assessment
	}

	return nil
}

// submitBatchRiskAssessment submits a batch risk assessment request
func submitBatchRiskAssessment(t *testing.T, client *http.Client, baseURL string, requests []*models.RiskAssessmentRequest) string {
	reqBody, err := json.Marshal(map[string]interface{}{
		"requests": requests,
	})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/api/v1/assess/batch", bytes.NewBuffer(reqBody))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	batchID, ok := response["batch_id"].(string)
	require.True(t, ok)
	return batchID
}

// waitForBatchCompletion waits for a batch assessment to complete
func waitForBatchCompletion(t *testing.T, client *http.Client, baseURL string, batchID string, timeout time.Duration) *models.BatchRiskAssessmentResponse {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		batchResult := getBatchResult(t, client, baseURL, batchID)
		if batchResult != nil && batchResult.Status == "completed" {
			return batchResult
		}
		time.Sleep(2 * time.Second)
	}

	t.Fatalf("Batch %s did not complete within %v", batchID, timeout)
	return nil
}

// getBatchResult retrieves a batch assessment result
func getBatchResult(t *testing.T, client *http.Client, baseURL string, batchID string) *models.BatchRiskAssessmentResponse {
	req, err := http.NewRequest("GET", baseURL+"/api/v1/assess/batch/"+batchID, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var batchResult models.BatchRiskAssessmentResponse
		err := json.NewDecoder(resp.Body).Decode(&batchResult)
		require.NoError(t, err)
		return &batchResult
	}

	return nil
}

// performScenarioAnalysis performs scenario analysis on an assessment
func performScenarioAnalysis(t *testing.T, client *http.Client, baseURL string, assessmentID string) *models.ScenarioAnalysis {
	scenarios := []models.RiskScenario{
		{Name: "optimistic", Multiplier: 0.8},
		{Name: "realistic", Multiplier: 1.0},
		{Name: "pessimistic", Multiplier: 1.2},
	}

	reqBody, err := json.Marshal(map[string]interface{}{
		"assessment_id": assessmentID,
		"scenarios":     scenarios,
	})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/api/v1/assess/scenarios", bytes.NewBuffer(reqBody))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var analysis models.ScenarioAnalysis
	err = json.NewDecoder(resp.Body).Decode(&analysis)
	require.NoError(t, err)

	return &analysis
}

// getRiskFactorExplanation gets risk factor explanation for an assessment
func getRiskFactorExplanation(t *testing.T, client *http.Client, baseURL string, assessmentID string) *models.RiskFactorExplanation {
	req, err := http.NewRequest("GET", baseURL+"/api/v1/assess/"+assessmentID+"/explanation", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var explanation models.RiskFactorExplanation
	err = json.NewDecoder(resp.Body).Decode(&explanation)
	require.NoError(t, err)

	return &explanation
}

// getRiskHistory gets risk history for an assessment
func getRiskHistory(t *testing.T, client *http.Client, baseURL string, assessmentID string) *models.RiskHistory {
	req, err := http.NewRequest("GET", baseURL+"/api/v1/assess/"+assessmentID+"/history", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var history models.RiskHistory
	err = json.NewDecoder(resp.Body).Decode(&history)
	require.NoError(t, err)

	return &history
}
