package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pcraw4d/business-verification/internal/external"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewVerificationStatusHandler(t *testing.T) {
	logger := zap.NewNop()
	statusAssigner := external.NewStatusAssigner(nil, logger)

	handler := NewVerificationStatusHandler(statusAssigner, logger)
	assert.NotNil(t, handler)
	assert.Equal(t, statusAssigner, handler.statusAssigner)
	assert.Equal(t, logger, handler.logger)
}

func TestVerificationStatusHandler_AssignVerificationStatus(t *testing.T) {
	logger := zap.NewNop()
	statusAssigner := external.NewStatusAssigner(nil, logger)
	handler := NewVerificationStatusHandler(statusAssigner, logger)

	// Test successful status assignment
	comparisonResult := &external.ComparisonResult{
		OverallScore:    0.85,
		ConfidenceLevel: "high",
		FieldResults: map[string]external.FieldComparison{
			"business_name": {
				Score:      0.9,
				Confidence: 0.8,
				Matched:    true,
				Reasoning:  "High similarity match",
			},
			"phone_numbers": {
				Score:      0.95,
				Confidence: 0.9,
				Matched:    true,
				Reasoning:  "Exact match",
			},
			"email_addresses": {
				Score:      0.9,
				Confidence: 0.8,
				Matched:    true,
				Reasoning:  "Exact match",
			},
		},
	}

	request := AssignStatusRequest{
		ComparisonResult: comparisonResult,
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/assign-status", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.AssignVerificationStatus(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response AssignStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Result)
	assert.Equal(t, external.StatusPassed, response.Result.Status)
	assert.Equal(t, 0.85, response.Result.OverallScore)
}

func TestVerificationStatusHandler_AssignVerificationStatus_MissingComparisonResult(t *testing.T) {
	logger := zap.NewNop()
	statusAssigner := external.NewStatusAssigner(nil, logger)
	handler := NewVerificationStatusHandler(statusAssigner, logger)

	request := AssignStatusRequest{
		ComparisonResult: nil,
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/assign-status", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.AssignVerificationStatus(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVerificationStatusHandler_AssignVerificationStatus_WrongMethod(t *testing.T) {
	logger := zap.NewNop()
	statusAssigner := external.NewStatusAssigner(nil, logger)
	handler := NewVerificationStatusHandler(statusAssigner, logger)

	req := httptest.NewRequest(http.MethodGet, "/assign-status", nil)
	w := httptest.NewRecorder()

	handler.AssignVerificationStatus(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestVerificationStatusHandler_AssignVerificationStatus_InvalidJSON(t *testing.T) {
	logger := zap.NewNop()
	statusAssigner := external.NewStatusAssigner(nil, logger)
	handler := NewVerificationStatusHandler(statusAssigner, logger)

	req := httptest.NewRequest(http.MethodPost, "/assign-status", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	handler.AssignVerificationStatus(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVerificationStatusHandler_AssignVerificationStatus_WithCustomCriteria(t *testing.T) {
	logger := zap.NewNop()
	statusAssigner := external.NewStatusAssigner(nil, logger)
	handler := NewVerificationStatusHandler(statusAssigner, logger)

	comparisonResult := &external.ComparisonResult{
		OverallScore:    0.85,
		ConfidenceLevel: "high",
		FieldResults: map[string]external.FieldComparison{
			"business_name": {
				Score:      0.9,
				Confidence: 0.8,
				Matched:    true,
				Reasoning:  "High similarity match",
			},
			"phone_numbers": {
				Score:      0.95,
				Confidence: 0.9,
				Matched:    true,
				Reasoning:  "Exact match",
			},
			"email_addresses": {
				Score:      0.9,
				Confidence: 0.8,
				Matched:    true,
				Reasoning:  "Exact match",
			},
		},
	}

	passedThreshold := 0.9
	customCriteria := &VerificationCriteriaRequest{
		PassedThreshold: &passedThreshold,
	}

	request := AssignStatusRequest{
		ComparisonResult: comparisonResult,
		CustomCriteria:   customCriteria,
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/assign-status", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.AssignVerificationStatus(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response AssignStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Result)
	// With custom threshold of 0.9 and score of 0.85, should be PARTIAL
	assert.Equal(t, external.StatusPartial, response.Result.Status)
}

func TestVerificationStatusHandler_BatchAssignVerificationStatus(t *testing.T) {
	logger := zap.NewNop()
	statusAssigner := external.NewStatusAssigner(nil, logger)
	handler := NewVerificationStatusHandler(statusAssigner, logger)

	comparison1 := &external.ComparisonResult{
		OverallScore:    0.85,
		ConfidenceLevel: "high",
		FieldResults: map[string]external.FieldComparison{
			"business_name": {
				Score:      0.9,
				Confidence: 0.8,
				Matched:    true,
			},
			"phone_numbers": {
				Score:      0.95,
				Confidence: 0.9,
				Matched:    true,
			},
			"email_addresses": {
				Score:      0.9,
				Confidence: 0.8,
				Matched:    true,
			},
		},
	}

	comparison2 := &external.ComparisonResult{
		OverallScore:    0.7,
		ConfidenceLevel: "medium",
		FieldResults: map[string]external.FieldComparison{
			"business_name": {
				Score:      0.8,
				Confidence: 0.7,
				Matched:    true,
			},
			"phone_numbers": {
				Score:      0.8,
				Confidence: 0.7,
				Matched:    true,
			},
			"email_addresses": {
				Score:      0.9,
				Confidence: 0.8,
				Matched:    true,
			},
		},
	}

	request := BatchAssignStatusRequest{
		Comparisons: []AssignStatusRequest{
			{ComparisonResult: comparison1},
			{ComparisonResult: comparison2},
		},
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/assign-status/batch", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.BatchAssignVerificationStatus(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response BatchAssignStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Len(t, response.Results, 2)
	assert.Equal(t, external.StatusPassed, response.Results[0].Status)
	assert.Equal(t, external.StatusPartial, response.Results[1].Status)
}

func TestVerificationStatusHandler_BatchAssignVerificationStatus_EmptyBatch(t *testing.T) {
	logger := zap.NewNop()
	statusAssigner := external.NewStatusAssigner(nil, logger)
	handler := NewVerificationStatusHandler(statusAssigner, logger)

	request := BatchAssignStatusRequest{
		Comparisons: []AssignStatusRequest{},
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/assign-status/batch", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.BatchAssignVerificationStatus(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVerificationStatusHandler_BatchAssignVerificationStatus_TooLargeBatch(t *testing.T) {
	logger := zap.NewNop()
	statusAssigner := external.NewStatusAssigner(nil, logger)
	handler := NewVerificationStatusHandler(statusAssigner, logger)

	// Create 101 comparisons (over the limit of 100)
	comparisons := make([]AssignStatusRequest, 101)
	for i := 0; i < 101; i++ {
		comparisons[i] = AssignStatusRequest{
			ComparisonResult: &external.ComparisonResult{
				OverallScore:    0.8,
				ConfidenceLevel: "high",
				FieldResults: map[string]external.FieldComparison{
					"business_name": {
						Score:      0.9,
						Confidence: 0.8,
						Matched:    true,
					},
				},
			},
		}
	}

	request := BatchAssignStatusRequest{
		Comparisons: comparisons,
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/assign-status/batch", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.BatchAssignVerificationStatus(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVerificationStatusHandler_BatchAssignVerificationStatus_WrongMethod(t *testing.T) {
	logger := zap.NewNop()
	statusAssigner := external.NewStatusAssigner(nil, logger)
	handler := NewVerificationStatusHandler(statusAssigner, logger)

	req := httptest.NewRequest(http.MethodGet, "/assign-status/batch", nil)
	w := httptest.NewRecorder()

	handler.BatchAssignVerificationStatus(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestVerificationStatusHandler_GetVerificationCriteria(t *testing.T) {
	logger := zap.NewNop()
	statusAssigner := external.NewStatusAssigner(nil, logger)
	handler := NewVerificationStatusHandler(statusAssigner, logger)

	req := httptest.NewRequest(http.MethodGet, "/criteria", nil)
	w := httptest.NewRecorder()

	handler.GetVerificationCriteria(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["criteria"])
}

func TestVerificationStatusHandler_GetVerificationCriteria_WrongMethod(t *testing.T) {
	logger := zap.NewNop()
	statusAssigner := external.NewStatusAssigner(nil, logger)
	handler := NewVerificationStatusHandler(statusAssigner, logger)

	req := httptest.NewRequest(http.MethodPost, "/criteria", nil)
	w := httptest.NewRecorder()

	handler.GetVerificationCriteria(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestVerificationStatusHandler_UpdateVerificationCriteria(t *testing.T) {
	logger := zap.NewNop()
	statusAssigner := external.NewStatusAssigner(nil, logger)
	handler := NewVerificationStatusHandler(statusAssigner, logger)

	passedThreshold := 0.9
	partialThreshold := 0.7
	request := VerificationCriteriaRequest{
		PassedThreshold:  &passedThreshold,
		PartialThreshold: &partialThreshold,
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPut, "/criteria", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.UpdateVerificationCriteria(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Verification criteria updated successfully", response["message"])

	// Verify the criteria was actually updated
	criteria := statusAssigner.GetCriteria()
	assert.Equal(t, 0.9, criteria.PassedThreshold)
	assert.Equal(t, 0.7, criteria.PartialThreshold)
}

func TestVerificationStatusHandler_UpdateVerificationCriteria_WrongMethod(t *testing.T) {
	logger := zap.NewNop()
	statusAssigner := external.NewStatusAssigner(nil, logger)
	handler := NewVerificationStatusHandler(statusAssigner, logger)

	req := httptest.NewRequest(http.MethodGet, "/criteria", nil)
	w := httptest.NewRecorder()

	handler.UpdateVerificationCriteria(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestVerificationStatusHandler_GetVerificationStats(t *testing.T) {
	logger := zap.NewNop()
	statusAssigner := external.NewStatusAssigner(nil, logger)
	handler := NewVerificationStatusHandler(statusAssigner, logger)

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	w := httptest.NewRecorder()

	handler.GetVerificationStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["stats"])

	stats := response["stats"].(map[string]interface{})
	assert.Equal(t, float64(0), stats["total_verifications"])
	assert.Equal(t, float64(0), stats["passed_count"])
	assert.Equal(t, float64(0), stats["partial_count"])
	assert.Equal(t, float64(0), stats["failed_count"])
	assert.Equal(t, float64(0), stats["skipped_count"])
}

func TestVerificationStatusHandler_GetVerificationStats_WrongMethod(t *testing.T) {
	logger := zap.NewNop()
	statusAssigner := external.NewStatusAssigner(nil, logger)
	handler := NewVerificationStatusHandler(statusAssigner, logger)

	req := httptest.NewRequest(http.MethodPost, "/stats", nil)
	w := httptest.NewRecorder()

	handler.GetVerificationStats(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestVerificationStatusHandler_createCustomCriteria(t *testing.T) {
	logger := zap.NewNop()
	statusAssigner := external.NewStatusAssigner(nil, logger)
	handler := NewVerificationStatusHandler(statusAssigner, logger)

	passedThreshold := 0.9
	partialThreshold := 0.7
	maxDistanceKm := 25.0
	minConfidenceLevel := "high"
	required := true
	minScore := 0.8
	minConfidence := 0.7
	weight := 0.5

	request := &VerificationCriteriaRequest{
		PassedThreshold:    &passedThreshold,
		PartialThreshold:   &partialThreshold,
		CriticalFields:     []string{"business_name"},
		MaxDistanceKm:      &maxDistanceKm,
		MinConfidenceLevel: &minConfidenceLevel,
		FieldRequirements: map[string]FieldRequirementRequest{
			"business_name": {
				Required:      &required,
				MinScore:      &minScore,
				MinConfidence: &minConfidence,
				Weight:        &weight,
			},
		},
	}

	criteria := handler.createCustomCriteria(request)

	assert.Equal(t, 0.9, criteria.PassedThreshold)
	assert.Equal(t, 0.7, criteria.PartialThreshold)
	assert.Equal(t, []string{"business_name"}, criteria.CriticalFields)
	assert.Equal(t, 25.0, criteria.MaxDistanceKm)
	assert.Equal(t, "high", criteria.MinConfidenceLevel)

	// Check field requirements
	businessNameReq := criteria.FieldRequirements["business_name"]
	assert.True(t, businessNameReq.Required)
	assert.Equal(t, 0.8, businessNameReq.MinScore)
	assert.Equal(t, 0.7, businessNameReq.MinConfidence)
	assert.Equal(t, 0.5, businessNameReq.Weight)
}
