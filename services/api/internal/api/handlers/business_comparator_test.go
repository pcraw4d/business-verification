package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kyb-platform/internal/external"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewBusinessComparatorHandler(t *testing.T) {
	logger := zap.NewNop()
	comparator := external.NewBusinessComparator(logger, nil)

	handler := NewBusinessComparatorHandler(comparator, logger)
	assert.NotNil(t, handler)
	assert.Equal(t, comparator, handler.comparator)
	assert.Equal(t, logger, handler.logger)
}

func TestBusinessComparatorHandler_CompareBusiness(t *testing.T) {
	logger := zap.NewNop()
	comparator := external.NewBusinessComparator(logger, nil)
	handler := NewBusinessComparatorHandler(comparator, logger)

	// Test successful comparison
	claimed := &external.ComparisonBusinessInfo{
		Name:           "Acme Corporation",
		PhoneNumbers:   []string{"+1-555-123-4567"},
		EmailAddresses: []string{"contact@acme.com"},
		Website:        "https://www.acme.com",
	}

	extracted := &external.ComparisonBusinessInfo{
		Name:           "Acme Corp",
		PhoneNumbers:   []string{"555-123-4567"},
		EmailAddresses: []string{"contact@acme.com"},
		Website:        "https://acme.com",
	}

	request := CompareBusinessRequest{
		Claimed:   claimed,
		Extracted: extracted,
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/compare", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CompareBusiness(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response CompareBusinessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Result)
	assert.Greater(t, response.Result.OverallScore, 0.7)
}

func TestBusinessComparatorHandler_CompareBusiness_InvalidRequest(t *testing.T) {
	logger := zap.NewNop()
	comparator := external.NewBusinessComparator(logger, nil)
	handler := NewBusinessComparatorHandler(comparator, logger)

	// Test missing claimed business info
	request := CompareBusinessRequest{
		Extracted: &external.ComparisonBusinessInfo{
			Name: "Acme Corp",
		},
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/compare", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CompareBusiness(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBusinessComparatorHandler_CompareBusiness_InvalidMethod(t *testing.T) {
	logger := zap.NewNop()
	comparator := external.NewBusinessComparator(logger, nil)
	handler := NewBusinessComparatorHandler(comparator, logger)

	req := httptest.NewRequest(http.MethodGet, "/compare", nil)
	w := httptest.NewRecorder()

	handler.CompareBusiness(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestBusinessComparatorHandler_CompareBusiness_InvalidJSON(t *testing.T) {
	logger := zap.NewNop()
	comparator := external.NewBusinessComparator(logger, nil)
	handler := NewBusinessComparatorHandler(comparator, logger)

	req := httptest.NewRequest(http.MethodPost, "/compare", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	handler.CompareBusiness(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBusinessComparatorHandler_CompareBusiness_WithCustomConfig(t *testing.T) {
	logger := zap.NewNop()
	comparator := external.NewBusinessComparator(logger, nil)
	handler := NewBusinessComparatorHandler(comparator, logger)

	claimed := &external.ComparisonBusinessInfo{
		Name:         "Acme Corporation",
		PhoneNumbers: []string{"+1-555-123-4567"},
	}

	extracted := &external.ComparisonBusinessInfo{
		Name:         "Acme Corp",
		PhoneNumbers: []string{"555-123-4567"},
	}

	threshold := 0.9
	request := CompareBusinessRequest{
		Claimed:   claimed,
		Extracted: extracted,
		Config: &ComparisonConfigRequest{
			MinSimilarityThreshold: &threshold,
		},
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/compare", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CompareBusiness(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response CompareBusinessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Config)
	assert.Equal(t, threshold, response.Config.MinSimilarityThreshold)
}

func TestBusinessComparatorHandler_GetComparisonConfig(t *testing.T) {
	logger := zap.NewNop()
	comparator := external.NewBusinessComparator(logger, nil)
	handler := NewBusinessComparatorHandler(comparator, logger)

	req := httptest.NewRequest(http.MethodGet, "/config", nil)
	w := httptest.NewRecorder()

	handler.GetComparisonConfig(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["config"])
}

func TestBusinessComparatorHandler_GetComparisonConfig_InvalidMethod(t *testing.T) {
	logger := zap.NewNop()
	comparator := external.NewBusinessComparator(logger, nil)
	handler := NewBusinessComparatorHandler(comparator, logger)

	req := httptest.NewRequest(http.MethodPost, "/config", nil)
	w := httptest.NewRecorder()

	handler.GetComparisonConfig(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestBusinessComparatorHandler_UpdateComparisonConfig(t *testing.T) {
	logger := zap.NewNop()
	comparator := external.NewBusinessComparator(logger, nil)
	handler := NewBusinessComparatorHandler(comparator, logger)

	threshold := 0.9
	configReq := ComparisonConfigRequest{
		MinSimilarityThreshold: &threshold,
	}

	body, _ := json.Marshal(configReq)
	req := httptest.NewRequest(http.MethodPut, "/config", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.UpdateComparisonConfig(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Configuration updated successfully", response["message"])

	// Verify the config was actually updated
	config := handler.comparator.GetConfig()
	assert.Equal(t, threshold, config.MinSimilarityThreshold)
}

func TestBusinessComparatorHandler_UpdateComparisonConfig_InvalidMethod(t *testing.T) {
	logger := zap.NewNop()
	comparator := external.NewBusinessComparator(logger, nil)
	handler := NewBusinessComparatorHandler(comparator, logger)

	req := httptest.NewRequest(http.MethodGet, "/config", nil)
	w := httptest.NewRecorder()

	handler.UpdateComparisonConfig(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestBusinessComparatorHandler_CompareBusinessBatch(t *testing.T) {
	logger := zap.NewNop()
	comparator := external.NewBusinessComparator(logger, nil)
	handler := NewBusinessComparatorHandler(comparator, logger)

	comparison1 := CompareBusinessRequest{
		Claimed: &external.ComparisonBusinessInfo{
			Name:         "Acme Corporation",
			PhoneNumbers: []string{"+1-555-123-4567"},
		},
		Extracted: &external.ComparisonBusinessInfo{
			Name:         "Acme Corp",
			PhoneNumbers: []string{"555-123-4567"},
		},
	}

	comparison2 := CompareBusinessRequest{
		Claimed: &external.ComparisonBusinessInfo{
			Name:           "XYZ Company",
			EmailAddresses: []string{"contact@xyz.com"},
		},
		Extracted: &external.ComparisonBusinessInfo{
			Name:           "XYZ Corp",
			EmailAddresses: []string{"contact@xyz.com"},
		},
	}

	request := struct {
		Comparisons []CompareBusinessRequest `json:"comparisons"`
	}{
		Comparisons: []CompareBusinessRequest{comparison1, comparison2},
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/compare/batch", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CompareBusinessBatch(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["results"])
	assert.NotNil(t, response["summary"])

	summary := response["summary"].(map[string]interface{})
	assert.Equal(t, float64(2), summary["total_comparisons"])
	assert.Equal(t, float64(2), summary["successful"])
	assert.Equal(t, float64(0), summary["failed"])
}

func TestBusinessComparatorHandler_CompareBusinessBatch_EmptyRequest(t *testing.T) {
	logger := zap.NewNop()
	comparator := external.NewBusinessComparator(logger, nil)
	handler := NewBusinessComparatorHandler(comparator, logger)

	request := struct {
		Comparisons []CompareBusinessRequest `json:"comparisons"`
	}{
		Comparisons: []CompareBusinessRequest{},
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/compare/batch", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CompareBusinessBatch(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBusinessComparatorHandler_CompareBusinessBatch_TooLarge(t *testing.T) {
	logger := zap.NewNop()
	comparator := external.NewBusinessComparator(logger, nil)
	handler := NewBusinessComparatorHandler(comparator, logger)

	// Create 101 comparisons (over the limit of 100)
	comparisons := make([]CompareBusinessRequest, 101)
	for i := 0; i < 101; i++ {
		comparisons[i] = CompareBusinessRequest{
			Claimed: &external.ComparisonBusinessInfo{
				Name: "Company " + string(rune(i)),
			},
			Extracted: &external.ComparisonBusinessInfo{
				Name: "Company " + string(rune(i)),
			},
		}
	}

	request := struct {
		Comparisons []CompareBusinessRequest `json:"comparisons"`
	}{
		Comparisons: comparisons,
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/compare/batch", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CompareBusinessBatch(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBusinessComparatorHandler_CompareBusinessBatch_InvalidMethod(t *testing.T) {
	logger := zap.NewNop()
	comparator := external.NewBusinessComparator(logger, nil)
	handler := NewBusinessComparatorHandler(comparator, logger)

	req := httptest.NewRequest(http.MethodGet, "/compare/batch", nil)
	w := httptest.NewRecorder()

	handler.CompareBusinessBatch(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestBusinessComparatorHandler_GetComparisonStats(t *testing.T) {
	logger := zap.NewNop()
	comparator := external.NewBusinessComparator(logger, nil)
	handler := NewBusinessComparatorHandler(comparator, logger)

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	w := httptest.NewRecorder()

	handler.GetComparisonStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["stats"])

	stats := response["stats"].(map[string]interface{})
	assert.Equal(t, float64(0), stats["total_comparisons"])
	assert.Equal(t, float64(0), stats["average_score"])
	assert.Equal(t, float64(0), stats["pass_rate"])
	assert.Equal(t, float64(0), stats["partial_rate"])
	assert.Equal(t, float64(0), stats["fail_rate"])
}

func TestBusinessComparatorHandler_GetComparisonStats_InvalidMethod(t *testing.T) {
	logger := zap.NewNop()
	comparator := external.NewBusinessComparator(logger, nil)
	handler := NewBusinessComparatorHandler(comparator, logger)

	req := httptest.NewRequest(http.MethodPost, "/stats", nil)
	w := httptest.NewRecorder()

	handler.GetComparisonStats(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestBusinessComparatorHandler_createCustomConfig(t *testing.T) {
	logger := zap.NewNop()
	comparator := external.NewBusinessComparator(logger, nil)
	handler := NewBusinessComparatorHandler(comparator, logger)

	threshold := 0.9
	phoneWeight := 0.4
	emailWeight := 0.3

	configReq := &ComparisonConfigRequest{
		MinSimilarityThreshold: &threshold,
		Weights: &ComparisonWeightsRequest{
			PhoneNumber:  &phoneWeight,
			EmailAddress: &emailWeight,
		},
	}

	config := handler.createCustomConfig(configReq)

	assert.Equal(t, threshold, config.MinSimilarityThreshold)
	assert.Equal(t, phoneWeight, config.Weights.PhoneNumber)
	assert.Equal(t, emailWeight, config.Weights.EmailAddress)

	// Check that other values remain at defaults
	assert.Equal(t, 0.3, config.Weights.BusinessName)
	assert.Equal(t, 0.15, config.Weights.PhysicalAddress)
	assert.Equal(t, 0.05, config.Weights.Website)
	assert.Equal(t, 0.05, config.Weights.Industry)
}
