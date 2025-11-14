package routing

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/api/routes"
	"kyb-platform/internal/risk"
)

// TestEnhancedRiskRoutes tests the enhanced risk routes using httptest
func TestEnhancedRiskRoutes(t *testing.T) {
	// Create a test ServeMux
	mux := http.NewServeMux()

	// Create enhanced risk handler with minimal dependencies
	zapLogger := zap.NewNop()
	enhancedRiskFactory := risk.NewEnhancedRiskServiceFactory(zapLogger)
	enhancedCalculator := enhancedRiskFactory.CreateRiskFactorCalculator()
	recommendationEngine := enhancedRiskFactory.CreateRecommendationEngine()
	trendAnalysisService := enhancedRiskFactory.CreateTrendAnalysisService()
	alertSystem := enhancedRiskFactory.CreateAlertSystem()

	// Create threshold manager (in-memory for testing)
	thresholdManager := risk.CreateDefaultThresholds()

	// Create enhanced risk handler
	enhancedRiskHandler := handlers.NewEnhancedRiskHandler(
		zapLogger,
		nil, // riskDetectionService
		enhancedCalculator,
		recommendationEngine,
		trendAnalysisService,
		alertSystem,
		thresholdManager,
	)

	// Register routes
	routes.RegisterEnhancedRiskRoutes(mux, enhancedRiskHandler)
	routes.RegisterEnhancedRiskAdminRoutes(mux, enhancedRiskHandler)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:           "GET /v1/risk/thresholds",
			method:         "GET",
			path:           "/v1/risk/thresholds",
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "thresholds")
				assert.Contains(t, response, "count")
			},
		},
		{
			name:           "GET /v1/risk/factors",
			method:         "GET",
			path:           "/v1/risk/factors",
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
			},
		},
		{
			name:           "GET /v1/risk/categories",
			method:         "GET",
			path:           "/v1/risk/categories",
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
			},
		},
		{
			name:           "POST /v1/admin/risk/thresholds",
			method:         "POST",
			path:           "/v1/admin/risk/thresholds",
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "Status code mismatch for %s %s", tt.method, tt.path)

			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}
		})
	}
}

// TestRoutePatternMatching tests Go 1.22 ServeMux pattern matching
func TestRoutePatternMatching(t *testing.T) {
	mux := http.NewServeMux()

	// Test 1: Method-specific pattern
	mux.Handle("GET /test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GET handler"))
	}))

	// Test 2: POST to same path
	mux.Handle("POST /test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("POST handler"))
	}))

	// Test GET
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "GET handler", rr.Body.String())

	// Test POST
	req = httptest.NewRequest("POST", "/test", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "POST handler", rr.Body.String())

	// Test PUT (should 404)
	req = httptest.NewRequest("PUT", "/test", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

