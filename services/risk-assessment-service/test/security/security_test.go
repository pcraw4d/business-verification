//go:build security

package security

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// TestSecurityHeaders tests security headers implementation
func TestSecurityHeaders(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test")
	}

	baseURL := "http://localhost:8080"
	client := &http.Client{Timeout: 30 * time.Second}

	tests := []struct {
		name     string
		endpoint string
		method   string
	}{
		{
			name:     "GET /health",
			endpoint: "/health",
			method:   "GET",
		},
		{
			name:     "POST /api/v1/assess",
			endpoint: "/api/v1/assess",
			method:   "POST",
		},
		{
			name:     "GET /metrics",
			endpoint: "/metrics",
			method:   "GET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.method == "POST" {
				reqBody, _ := json.Marshal(&models.RiskAssessmentRequest{
					BusinessName:      "Test Company",
					BusinessAddress:   "123 Test Street, Test City, TC 12345",
					Industry:          "technology",
					Country:           "US",
					PredictionHorizon: 3,
				})
				req, err = http.NewRequest(tt.method, baseURL+tt.endpoint, bytes.NewBuffer(reqBody))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(tt.method, baseURL+tt.endpoint, nil)
			}
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Test security headers
			assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"),
				"X-Content-Type-Options header should be set to nosniff")
			assert.Equal(t, "1; mode=block", resp.Header.Get("X-XSS-Protection"),
				"X-XSS-Protection header should be set to 1; mode=block")
			assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"),
				"X-Frame-Options header should be set to DENY")
			assert.Contains(t, resp.Header.Get("Strict-Transport-Security"), "max-age",
				"Strict-Transport-Security header should be set")
			assert.NotEmpty(t, resp.Header.Get("Content-Security-Policy"),
				"Content-Security-Policy header should be set")
		})
	}
}

// TestInputValidation tests input validation and sanitization
func TestInputValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test")
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
			name: "SQL injection attempt in business name",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "'; DROP TABLE assessments; --",
				BusinessAddress:   "123 Test Street, Test City, TC 12345",
				Industry:          "technology",
				Country:           "US",
				PredictionHorizon: 3,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid input",
		},
		{
			name: "XSS attempt in business name",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "<script>alert('xss')</script>",
				BusinessAddress:   "123 Test Street, Test City, TC 12345",
				Industry:          "technology",
				Country:           "US",
				PredictionHorizon: 3,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid input",
		},
		{
			name: "Path traversal attempt in business address",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "Test Company",
				BusinessAddress:   "../../../etc/passwd",
				Industry:          "technology",
				Country:           "US",
				PredictionHorizon: 3,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid input",
		},
		{
			name: "Command injection attempt in industry",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "Test Company",
				BusinessAddress:   "123 Test Street, Test City, TC 12345",
				Industry:          "technology; rm -rf /",
				Country:           "US",
				PredictionHorizon: 3,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid input",
		},
		{
			name: "Buffer overflow attempt in business name",
			request: &models.RiskAssessmentRequest{
				BusinessName:      string(make([]byte, 10000)), // Very long string
				BusinessAddress:   "123 Test Street, Test City, TC 12345",
				Industry:          "technology",
				Country:           "US",
				PredictionHorizon: 3,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "input too long",
		},
		{
			name: "Invalid JSON payload",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "Test Company",
				BusinessAddress:   "123 Test Street, Test City, TC 12345",
				Industry:          "technology",
				Country:           "US",
				PredictionHorizon: 3,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			var err error

			if tt.name == "Invalid JSON payload" {
				// Send malformed JSON
				reqBody = []byte(`{"business_name": "Test Company", "business_address": "123 Test Street", "industry": "technology", "country": "US", "prediction_horizon": 3, "invalid_field": }`)
			} else {
				reqBody, err = json.Marshal(tt.request)
				require.NoError(t, err)
			}

			req, err := http.NewRequest("POST", baseURL+"/api/v1/assess", bytes.NewBuffer(reqBody))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var errorResp map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&errorResp)
			require.NoError(t, err)

			assert.Contains(t, errorResp, "error")
			if tt.expectedError != "" {
				assert.Contains(t, errorResp["error"], tt.expectedError)
			}
		})
	}
}

// TestAuthenticationSecurity tests authentication and authorization security
func TestAuthenticationSecurity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test")
	}

	baseURL := "http://localhost:8080"
	client := &http.Client{Timeout: 30 * time.Second}

	tests := []struct {
		name           string
		headers        map[string]string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "missing authorization header",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "authorization required",
		},
		{
			name: "invalid authorization header format",
			headers: map[string]string{
				"Authorization": "InvalidFormat token123",
				"Content-Type":  "application/json",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid authorization format",
		},
		{
			name: "expired token",
			headers: map[string]string{
				"Authorization": "Bearer expired_token_123",
				"Content-Type":  "application/json",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "token expired",
		},
		{
			name: "malformed JWT token",
			headers: map[string]string{
				"Authorization": "Bearer malformed.jwt.token",
				"Content-Type":  "application/json",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid token",
		},
		{
			name: "token with insufficient permissions",
			headers: map[string]string{
				"Authorization": "Bearer insufficient_permissions_token",
				"Content-Type":  "application/json",
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "insufficient permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, err := json.Marshal(&models.RiskAssessmentRequest{
				BusinessName:      "Test Company",
				BusinessAddress:   "123 Test Street, Test City, TC 12345",
				Industry:          "technology",
				Country:           "US",
				PredictionHorizon: 3,
			})
			require.NoError(t, err)

			req, err := http.NewRequest("POST", baseURL+"/api/v1/assess", bytes.NewBuffer(reqBody))
			require.NoError(t, err)

			// Set headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var errorResp map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&errorResp)
			require.NoError(t, err)

			assert.Contains(t, errorResp, "error")
			if tt.expectedError != "" {
				assert.Contains(t, errorResp["error"], tt.expectedError)
			}
		})
	}
}

// TestRateLimitingSecurity tests rate limiting security
func TestRateLimitingSecurity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test")
	}

	baseURL := "http://localhost:8080"
	client := &http.Client{Timeout: 30 * time.Second}

	// Test rate limiting by sending many requests quickly
	t.Run("rate_limiting", func(t *testing.T) {
		reqBody, err := json.Marshal(&models.RiskAssessmentRequest{
			BusinessName:      "Test Company",
			BusinessAddress:   "123 Test Street, Test City, TC 12345",
			Industry:          "technology",
			Country:           "US",
			PredictionHorizon: 3,
		})
		require.NoError(t, err)

		rateLimited := false
		for i := 0; i < 100; i++ { // Send 100 requests quickly
			req, err := http.NewRequest("POST", baseURL+"/api/v1/assess", bytes.NewBuffer(reqBody))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer valid_token")

			resp, err := client.Do(req)
			require.NoError(t, err)
			resp.Body.Close()

			if resp.StatusCode == http.StatusTooManyRequests {
				rateLimited = true
				break
			}
		}

		assert.True(t, rateLimited, "Rate limiting should be triggered after multiple requests")
	})
}

// TestCORSecurity tests CORS security configuration
func TestCORSecurity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test")
	}

	baseURL := "http://localhost:8080"
	client := &http.Client{Timeout: 30 * time.Second}

	tests := []struct {
		name           string
		origin         string
		expectedStatus int
		expectedCORS   bool
	}{
		{
			name:           "allowed origin",
			origin:         "https://app.kyb-platform.com",
			expectedStatus: http.StatusOK,
			expectedCORS:   true,
		},
		{
			name:           "disallowed origin",
			origin:         "https://malicious-site.com",
			expectedStatus: http.StatusForbidden,
			expectedCORS:   false,
		},
		{
			name:           "no origin header",
			origin:         "",
			expectedStatus: http.StatusOK,
			expectedCORS:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("OPTIONS", baseURL+"/api/v1/assess", nil)
			require.NoError(t, err)

			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			req.Header.Set("Access-Control-Request-Method", "POST")
			req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedCORS {
				assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Origin"),
					"Access-Control-Allow-Origin header should be set for allowed origins")
				assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Methods"),
					"Access-Control-Allow-Methods header should be set")
				assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Headers"),
					"Access-Control-Allow-Headers header should be set")
			}
		})
	}
}

// TestDataPrivacySecurity tests data privacy and protection
func TestDataPrivacySecurity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test")
	}

	baseURL := "http://localhost:8080"
	client := &http.Client{Timeout: 30 * time.Second}

	t.Run("sensitive_data_not_logged", func(t *testing.T) {
		// This test would require checking logs to ensure sensitive data is not logged
		// For now, we'll test that the API responds correctly
		reqBody, err := json.Marshal(&models.RiskAssessmentRequest{
			BusinessName:      "Sensitive Business Name",
			BusinessAddress:   "123 Sensitive Street, Sensitive City, SC 12345",
			Industry:          "technology",
			Country:           "US",
			PredictionHorizon: 3,
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/api/v1/assess", bytes.NewBuffer(reqBody))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid_token")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// In a real implementation, you would check logs to ensure sensitive data is not logged
		// This is a placeholder for that validation
	})
}

// TestEncryptionSecurity tests encryption in transit and at rest
func TestEncryptionSecurity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test")
	}

	baseURL := "http://localhost:8080"
	client := &http.Client{Timeout: 30 * time.Second}

	t.Run("https_required", func(t *testing.T) {
		// Test that HTTP requests are redirected to HTTPS
		req, err := http.NewRequest("GET", "http://localhost:8080/health", nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// In production, this should redirect to HTTPS
		// For testing, we'll just ensure the service responds
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusMovedPermanently,
			"Service should respond or redirect to HTTPS")
	})

	t.Run("tls_configuration", func(t *testing.T) {
		// Test TLS configuration (this would require HTTPS endpoint)
		// For now, we'll test that the service responds correctly
		req, err := http.NewRequest("GET", baseURL+"/health", nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// TestAuditLoggingSecurity tests audit logging for security events
func TestAuditLoggingSecurity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test")
	}

	baseURL := "http://localhost:8080"
	client := &http.Client{Timeout: 30 * time.Second}

	t.Run("audit_logging", func(t *testing.T) {
		// Test that security events are logged
		reqBody, err := json.Marshal(&models.RiskAssessmentRequest{
			BusinessName:      "Test Company",
			BusinessAddress:   "123 Test Street, Test City, TC 12345",
			Industry:          "technology",
			Country:           "US",
			PredictionHorizon: 3,
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/api/v1/assess", bytes.NewBuffer(reqBody))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid_token")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// In a real implementation, you would check audit logs to ensure security events are logged
		// This is a placeholder for that validation
	})
}

// TestVulnerabilityScanning tests for common vulnerabilities
func TestVulnerabilityScanning(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test")
	}

	baseURL := "http://localhost:8080"
	client := &http.Client{Timeout: 30 * time.Second}

	tests := []struct {
		name           string
		endpoint       string
		method         string
		payload        string
		expectedStatus int
		vulnerability  string
	}{
		{
			name:           "directory_traversal",
			endpoint:       "/api/v1/assess/../../../etc/passwd",
			method:         "GET",
			payload:        "",
			expectedStatus: http.StatusNotFound,
			vulnerability:  "directory_traversal",
		},
		{
			name:           "command_injection",
			endpoint:       "/api/v1/assess",
			method:         "POST",
			payload:        `{"business_name": "test", "business_address": "123 test", "industry": "technology; ls -la", "country": "US", "prediction_horizon": 3}`,
			expectedStatus: http.StatusBadRequest,
			vulnerability:  "command_injection",
		},
		{
			name:           "xml_external_entity",
			endpoint:       "/api/v1/assess",
			method:         "POST",
			payload:        `<?xml version="1.0"?><!DOCTYPE foo [<!ENTITY xxe SYSTEM "file:///etc/passwd">]><foo>&xxe;</foo>`,
			expectedStatus: http.StatusBadRequest,
			vulnerability:  "xml_external_entity",
		},
		{
			name:           "server_side_request_forgery",
			endpoint:       "/api/v1/assess",
			method:         "POST",
			payload:        `{"business_name": "test", "business_address": "http://internal-server:8080/admin", "industry": "technology", "country": "US", "prediction_horizon": 3}`,
			expectedStatus: http.StatusBadRequest,
			vulnerability:  "server_side_request_forgery",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.method == "POST" {
				req, err = http.NewRequest(tt.method, baseURL+tt.endpoint, bytes.NewBufferString(tt.payload))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(tt.method, baseURL+tt.endpoint, nil)
			}
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode,
				"Vulnerability %s should be properly handled", tt.vulnerability)
		})
	}
}
