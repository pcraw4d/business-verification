package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"kyb-platform/services/api-gateway/internal/config"
)

func TestProxyToDashboardMetricsV3(t *testing.T) {
	// Create a mock BI service
	mockBIService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dashboard/kpis" {
			t.Errorf("Expected path /dashboard/kpis, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"overview": map[string]interface{}{
				"total_requests": 125000,
				"active_users":   45,
			},
			"business": map[string]interface{}{
				"total_verifications": 2500,
				"revenue":             1000000,
			},
		})
	}))
	defer mockBIService.Close()

	// Create handler with mock config
	cfg := &config.Config{
		Services: config.ServicesConfig{
			BIServiceURL: mockBIService.URL,
		},
	}
	logger := zap.NewNop()
	handler := NewGatewayHandler(nil, logger, cfg)

	// Create request
	req := httptest.NewRequest("GET", "/api/v3/dashboard/metrics", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.ProxyToDashboardMetricsV3(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["overview"] == nil {
		t.Error("Expected overview in response")
	}
}

// TestProxyToDashboardMetricsV1 removed - v1 endpoint deprecated in favor of v3

func TestProxyToComplianceStatus(t *testing.T) {
	// Create a mock Risk Assessment service
	mockRiskService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should route to /api/v1/compliance/status/aggregate when no business_id
		if r.URL.Path != "/api/v1/compliance/status/aggregate" {
			t.Errorf("Expected path /api/v1/compliance/status/aggregate, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"compliance_score": 0.95,
			"frameworks": []map[string]interface{}{
				{"framework_id": "SOC2", "score": 0.98},
				{"framework_id": "GDPR", "score": 0.92},
			},
		})
	}))
	defer mockRiskService.Close()

	// Create handler with mock config
	cfg := &config.Config{
		Services: config.ServicesConfig{
			RiskAssessmentURL: mockRiskService.URL,
		},
	}
	logger := zap.NewNop()
	handler := NewGatewayHandler(nil, logger, cfg)

	// Create request without business_id
	req := httptest.NewRequest("GET", "/api/v1/compliance/status", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.ProxyToComplianceStatus(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestProxyToComplianceStatusWithBusinessID(t *testing.T) {
	// Create a mock Risk Assessment service
	mockRiskService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should route to /api/v1/compliance/status/{business_id} when business_id provided
		if r.URL.Path != "/api/v1/compliance/status/test-business-123" {
			t.Errorf("Expected path /api/v1/compliance/status/test-business-123, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"compliance_score": 0.95,
			"business_id":      "test-business-123",
		})
	}))
	defer mockRiskService.Close()

	// Create handler with mock config
	cfg := &config.Config{
		Services: config.ServicesConfig{
			RiskAssessmentURL: mockRiskService.URL,
		},
	}
	logger := zap.NewNop()
	handler := NewGatewayHandler(nil, logger, cfg)

	// Create request with business_id
	req := httptest.NewRequest("GET", "/api/v1/compliance/status?business_id=test-business-123", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.ProxyToComplianceStatus(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestProxyToSessions(t *testing.T) {
	// Create a mock Frontend service
	mockFrontendService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should route to /v1/sessions (removed /api prefix)
		if r.URL.Path != "/v1/sessions" {
			t.Errorf("Expected path /v1/sessions, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"sessions": []map[string]interface{}{
				{
					"id":         "session-1",
					"user_id":    "user-123",
					"created_at": "2025-11-17T20:00:00Z",
				},
			},
		})
	}))
	defer mockFrontendService.Close()

	// Create handler with mock config
	cfg := &config.Config{
		Services: config.ServicesConfig{
			FrontendURL: mockFrontendService.URL,
		},
	}
	logger := zap.NewNop()
	handler := NewGatewayHandler(nil, logger, cfg)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/sessions", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.ProxyToSessions(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["success"] != true {
		t.Error("Expected success in response")
	}
}

func TestProxyToSessionsSubRoutes(t *testing.T) {
	// Create a mock Frontend service
	mockFrontendService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Test various sub-routes
		expectedPaths := map[string]bool{
			"/v1/sessions/current":  true,
			"/v1/sessions/metrics":  true,
			"/v1/sessions/activity": true,
			"/v1/sessions/status":   true,
		}
		if !expectedPaths[r.URL.Path] {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}))
	defer mockFrontendService.Close()

	// Create handler with mock config
	cfg := &config.Config{
		Services: config.ServicesConfig{
			FrontendURL: mockFrontendService.URL,
		},
	}
	logger := zap.NewNop()
	handler := NewGatewayHandler(nil, logger, cfg)

	// Test sub-routes
	subRoutes := []string{"current", "metrics", "activity", "status"}
	for _, subRoute := range subRoutes {
		req := httptest.NewRequest("GET", "/api/v1/sessions/"+subRoute, nil)
		w := httptest.NewRecorder()

		handler.ProxyToSessions(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for /sessions/%s, got %d", subRoute, w.Code)
		}
	}
}

func TestProxyToRiskAssessment_AnalyticsTrends(t *testing.T) {
	// Create a mock Risk Assessment service
	mockRiskService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Analytics routes should be kept as-is (no path transformation)
		if r.URL.Path != "/api/v1/analytics/trends" {
			t.Errorf("Expected path /api/v1/analytics/trends, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"trends": []map[string]interface{}{},
			"summary": map[string]interface{}{
				"average_risk_score": 0.6,
				"trend_direction":    "stable",
			},
		})
	}))
	defer mockRiskService.Close()

	// Create handler with mock config
	cfg := &config.Config{
		Services: config.ServicesConfig{
			RiskAssessmentURL: mockRiskService.URL,
		},
	}
	logger := zap.NewNop()
	handler := NewGatewayHandler(nil, logger, cfg)

	// Create request for analytics trends
	req := httptest.NewRequest("GET", "/api/v1/analytics/trends?timeframe=6m&limit=100", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.ProxyToRiskAssessment(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestProxyToRiskAssessment_AnalyticsInsights(t *testing.T) {
	// Create a mock Risk Assessment service
	mockRiskService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Analytics routes should be kept as-is (no path transformation)
		if r.URL.Path != "/api/v1/analytics/insights" {
			t.Errorf("Expected path /api/v1/analytics/insights, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"insights":       []map[string]interface{}{},
			"recommendations": []map[string]interface{}{},
		})
	}))
	defer mockRiskService.Close()

	// Create handler with mock config
	cfg := &config.Config{
		Services: config.ServicesConfig{
			RiskAssessmentURL: mockRiskService.URL,
		},
	}
	logger := zap.NewNop()
	handler := NewGatewayHandler(nil, logger, cfg)

	// Create request for analytics insights
	req := httptest.NewRequest("GET", "/api/v1/analytics/insights?industry=technology&risk_level=high", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.ProxyToRiskAssessment(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestProxyToRiskAssessment_PathTransformations(t *testing.T) {
	tests := []struct {
		name           string
		requestPath    string
		expectedPath   string
		description    string
	}{
		{
			name:         "risk assess path transformation",
			requestPath:  "/api/v1/risk/assess",
			expectedPath: "/api/v1/assess",
			description:  "Should transform /api/v1/risk/assess to /api/v1/assess",
		},
		{
			name:         "risk metrics path transformation",
			requestPath:  "/api/v1/risk/metrics",
			expectedPath: "/api/v1/metrics",
			description:  "Should transform /api/v1/risk/metrics to /api/v1/metrics",
		},
		{
			name:         "analytics trends no transformation",
			requestPath:  "/api/v1/analytics/trends",
			expectedPath: "/api/v1/analytics/trends",
			description:  "Should keep /api/v1/analytics/trends as-is",
		},
		{
			name:         "analytics insights no transformation",
			requestPath:  "/api/v1/analytics/insights",
			expectedPath: "/api/v1/analytics/insights",
			description:  "Should keep /api/v1/analytics/insights as-is",
		},
		{
			name:         "risk benchmarks no transformation",
			requestPath:  "/api/v1/risk/benchmarks",
			expectedPath: "/api/v1/risk/benchmarks",
			description:  "Should keep /api/v1/risk/benchmarks as-is",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock Risk Assessment service
			mockRiskService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.expectedPath {
					t.Errorf("%s: Expected path %s, got %s", tt.description, tt.expectedPath, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
				})
			}))
			defer mockRiskService.Close()

			// Create handler with mock config
			cfg := &config.Config{
				Services: config.ServicesConfig{
					RiskAssessmentURL: mockRiskService.URL,
				},
			}
			logger := zap.NewNop()
			handler := NewGatewayHandler(nil, logger, cfg)

			// Create request
			req := httptest.NewRequest("GET", tt.requestPath, nil)
			w := httptest.NewRecorder()

			// Call handler
			handler.ProxyToRiskAssessment(w, req)

			// Check response
			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		})
	}
}

func TestProxyToRiskAssessment_QueryParametersPreserved(t *testing.T) {
	// Create a mock Risk Assessment service
	mockRiskService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify path is correct
		if r.URL.Path != "/api/v1/analytics/trends" {
			t.Errorf("Expected path /api/v1/analytics/trends, got %s", r.URL.Path)
		}
		// Verify query parameters are preserved
		if r.URL.Query().Get("timeframe") != "6m" {
			t.Errorf("Expected timeframe=6m, got %s", r.URL.Query().Get("timeframe"))
		}
		if r.URL.Query().Get("limit") != "100" {
			t.Errorf("Expected limit=100, got %s", r.URL.Query().Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}))
	defer mockRiskService.Close()

	// Create handler with mock config
	cfg := &config.Config{
		Services: config.ServicesConfig{
			RiskAssessmentURL: mockRiskService.URL,
		},
	}
	logger := zap.NewNop()
	handler := NewGatewayHandler(nil, logger, cfg)

	// Create request with query parameters
	req := httptest.NewRequest("GET", "/api/v1/analytics/trends?timeframe=6m&limit=100", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.ProxyToRiskAssessment(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

