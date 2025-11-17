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

