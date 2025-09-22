package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"kyb-platform/internal/external"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewConfidenceScorerHandler(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()

	handler := NewConfidenceScorerHandler(scorer, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, scorer, handler.scorer)
	assert.Equal(t, logger, handler.logger)
}

func TestConfidenceScorerHandler_CalculateConfidence(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	// Test valid request
	verificationResult := &external.VerificationResult{
		ID:     "test-123",
		Status: "PASSED",
		FieldResults: map[string]external.FieldResult{
			"business_name": {
				Score:      0.9,
				Confidence: 0.95,
				Matched:    true,
			},
		},
	}

	reqBody := CalculateConfidenceRequest{
		VerificationResult: verificationResult,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/calculate-confidence", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CalculateConfidence(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response CalculateConfidenceResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.ConfidenceScore)
	assert.Greater(t, response.ConfidenceScore.OverallScore, 0.0)
}

func TestConfidenceScorerHandler_CalculateConfidence_InvalidMethod(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	req := httptest.NewRequest("GET", "/calculate-confidence", nil)
	w := httptest.NewRecorder()

	handler.CalculateConfidence(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestConfidenceScorerHandler_CalculateConfidence_InvalidJSON(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	req := httptest.NewRequest("POST", "/calculate-confidence", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	handler.CalculateConfidence(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConfidenceScorerHandler_CalculateConfidence_NilResult(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	reqBody := CalculateConfidenceRequest{
		VerificationResult: nil,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/calculate-confidence", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CalculateConfidence(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConfidenceScorerHandler_BatchCalculateConfidence(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	verificationResults := []*external.VerificationResult{
		{
			ID:     "test-1",
			Status: "PASSED",
			FieldResults: map[string]external.FieldResult{
				"business_name": {
					Score:      0.9,
					Confidence: 0.95,
					Matched:    true,
				},
			},
		},
		{
			ID:     "test-2",
			Status: "PARTIAL",
			FieldResults: map[string]external.FieldResult{
				"business_name": {
					Score:      0.7,
					Confidence: 0.75,
					Matched:    true,
				},
			},
		},
	}

	reqBody := BatchCalculateConfidenceRequest{
		VerificationResults: verificationResults,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/calculate-confidence/batch", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.BatchCalculateConfidence(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response BatchCalculateConfidenceResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Len(t, response.ConfidenceScores, 2)
}

func TestConfidenceScorerHandler_BatchCalculateConfidence_EmptyBatch(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	reqBody := BatchCalculateConfidenceRequest{
		VerificationResults: []*external.VerificationResult{},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/calculate-confidence/batch", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.BatchCalculateConfidence(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConfidenceScorerHandler_BatchCalculateConfidence_TooLargeBatch(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	// Create 101 verification results (exceeds limit of 100)
	verificationResults := make([]*external.VerificationResult, 101)
	for i := 0; i < 101; i++ {
		verificationResults[i] = &external.VerificationResult{
			ID:     "test-" + string(rune(i)),
			Status: "PASSED",
			FieldResults: map[string]external.FieldResult{
				"business_name": {
					Score:      0.9,
					Confidence: 0.95,
					Matched:    true,
				},
			},
		}
	}

	reqBody := BatchCalculateConfidenceRequest{
		VerificationResults: verificationResults,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/calculate-confidence/batch", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.BatchCalculateConfidence(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConfidenceScorerHandler_BatchCalculateConfidence_WrongMethod(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	req := httptest.NewRequest("GET", "/calculate-confidence/batch", nil)
	w := httptest.NewRecorder()

	handler.BatchCalculateConfidence(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestConfidenceScorerHandler_GetConfidenceScorerConfig(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	req := httptest.NewRequest("GET", "/config", nil)
	w := httptest.NewRecorder()

	handler.GetConfidenceScorerConfig(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ConfidenceScorerConfigResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Config.FieldWeights)
	assert.Equal(t, "weighted_average", response.Config.ScoringAlgorithm)
}

func TestConfidenceScorerHandler_GetConfidenceScorerConfig_WrongMethod(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	req := httptest.NewRequest("POST", "/config", nil)
	w := httptest.NewRecorder()

	handler.GetConfidenceScorerConfig(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestConfidenceScorerHandler_UpdateConfidenceScorerConfig(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	config := external.ConfidenceScorerConfig{
		FieldWeights: map[string]float64{
			"test_field": 0.5,
		},
		ConfidenceThresholds: external.ConfidenceThresholds{
			HighThreshold:   0.9,
			MediumThreshold: 0.7,
			LowThreshold:    0.5,
		},
		ScoringAlgorithm: "custom",
	}

	reqBody := ConfidenceScorerConfigRequest{
		Config: config,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/config", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.UpdateConfidenceScorerConfig(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ConfidenceScorerConfigResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "custom", response.Config.ScoringAlgorithm)
}

func TestConfidenceScorerHandler_UpdateConfidenceScorerConfig_WrongMethod(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	req := httptest.NewRequest("GET", "/config", nil)
	w := httptest.NewRecorder()

	handler.UpdateConfidenceScorerConfig(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestConfidenceScorerHandler_UpdateCalibrationData(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	reqBody := UpdateCalibrationRequest{
		Status: "PASSED",
		Scores: []float64{0.8, 0.7, 0.9},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/calibration", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.UpdateCalibrationData(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response UpdateCalibrationResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
}

func TestConfidenceScorerHandler_UpdateCalibrationData_EmptyStatus(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	reqBody := UpdateCalibrationRequest{
		Status: "",
		Scores: []float64{0.8, 0.7, 0.9},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/calibration", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.UpdateCalibrationData(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConfidenceScorerHandler_UpdateCalibrationData_EmptyScores(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	reqBody := UpdateCalibrationRequest{
		Status: "PASSED",
		Scores: []float64{},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/calibration", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.UpdateCalibrationData(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConfidenceScorerHandler_GetCalibrationData(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	req := httptest.NewRequest("GET", "/calibration", nil)
	w := httptest.NewRecorder()

	handler.GetCalibrationData(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["calibration_data"])
}

func TestConfidenceScorerHandler_GetCalibrationData_WrongMethod(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	req := httptest.NewRequest("POST", "/calibration", nil)
	w := httptest.NewRecorder()

	handler.GetCalibrationData(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestConfidenceScorerHandler_GetConfidenceScorerStats(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()

	handler.GetConfidenceScorerStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ConfidenceScorerStatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Statistics)
	assert.Contains(t, response.Statistics, "field_count")
	assert.Contains(t, response.Statistics, "algorithm")
}

func TestConfidenceScorerHandler_GetConfidenceScorerStats_WrongMethod(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	req := httptest.NewRequest("POST", "/stats", nil)
	w := httptest.NewRecorder()

	handler.GetConfidenceScorerStats(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestConfidenceScorerHandler_ValidateConfidenceScore(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	confidenceScore := external.ConfidenceScore{
		OverallScore:    0.8,
		ConfidenceLevel: external.ConfidenceLevelHigh,
		FieldScores: map[string]float64{
			"field1": 0.8,
			"field2": 0.7,
		},
	}

	body, _ := json.Marshal(confidenceScore)
	req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ValidateConfidenceScore(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["valid"].(bool))
	assert.True(t, response["success"].(bool))
}

func TestConfidenceScorerHandler_ValidateConfidenceScore_InvalidScore(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	confidenceScore := external.ConfidenceScore{
		OverallScore:    1.5, // Invalid: exceeds 1.0
		ConfidenceLevel: external.ConfidenceLevelHigh,
		FieldScores:     map[string]float64{},
	}

	body, _ := json.Marshal(confidenceScore)
	req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ValidateConfidenceScore(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response["valid"].(bool))
	assert.True(t, response["success"].(bool))
	assert.Contains(t, response["error"], "overall score must be between 0 and 1")
}

func TestConfidenceScorerHandler_RegisterRoutes(t *testing.T) {
	scorer := external.NewConfidenceScorer()
	logger := zap.NewNop()
	handler := NewConfidenceScorerHandler(scorer, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test that routes are registered by making requests
	req := httptest.NewRequest("GET", "/config", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should not return 404 (route not found)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}
