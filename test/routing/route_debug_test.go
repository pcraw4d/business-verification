package routing

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/api/routes"
	"kyb-platform/internal/risk"
)

// TestRouteRegistrationOrder tests route registration order and conflicts
func TestRouteRegistrationOrder(t *testing.T) {
	mux := http.NewServeMux()

	// Simulate the server setup
	zapLogger := zap.NewNop()
	enhancedRiskFactory := risk.NewEnhancedRiskServiceFactory(zapLogger)
	enhancedCalculator := enhancedRiskFactory.CreateRiskFactorCalculator()
	recommendationEngine := enhancedRiskFactory.CreateRecommendationEngine()
	trendAnalysisService := enhancedRiskFactory.CreateTrendAnalysisService()
	alertSystem := enhancedRiskFactory.CreateAlertSystem()
	thresholdManager := risk.CreateDefaultThresholds()

	enhancedRiskHandler := handlers.NewEnhancedRiskHandler(
		zapLogger,
		nil,
		enhancedCalculator,
		recommendationEngine,
		trendAnalysisService,
		alertSystem,
		thresholdManager,
	)

	// Register routes in the same order as the server
	routes.RegisterEnhancedRiskRoutes(mux, enhancedRiskHandler)
	routes.RegisterEnhancedRiskAdminRoutes(mux, enhancedRiskHandler)

	// Test that routes are actually registered
	testCases := []struct {
		method string
		path   string
	}{
		{"GET", "/v1/risk/thresholds"},
		{"GET", "/v1/risk/factors"},
		{"GET", "/v1/risk/categories"},
		{"POST", "/v1/admin/risk/thresholds"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s %s", tc.method, tc.path), func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			// Should not be 404 if route is registered
			assert.NotEqual(t, http.StatusNotFound, rr.Code,
				"Route %s %s returned 404 - route not registered", tc.method, tc.path)
		})
	}
}

// TestRouteConflictDetection tests for route conflicts
func TestRouteConflictDetection(t *testing.T) {
	mux := http.NewServeMux()

	// Register a route
	mux.Handle("GET /test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler1"))
	}))

	// Try to register conflicting route (should work in Go 1.22 - last one wins)
	mux.Handle("GET /test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler2"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	// Last registered handler should win
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "handler2", rr.Body.String())
}

