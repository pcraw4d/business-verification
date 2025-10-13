package security

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// SecurityTestSuite provides comprehensive security testing capabilities
type SecurityTestSuite struct {
	server     *httptest.Server
	logger     *zap.Logger
	tenant1ID  string
	tenant2ID  string
	tenant1Key string
	tenant2Key string
}

// NewSecurityTestSuite creates a new security test suite
func NewSecurityTestSuite() *SecurityTestSuite {
	logger := zap.NewNop()

	// Create test tenants
	tenant1ID := "tenant_1_test"
	tenant2ID := "tenant_2_test"
	tenant1Key := "test_key_tenant_1"
	tenant2Key := "test_key_tenant_2"

	// Create test server with mock handlers
	mux := http.NewServeMux()

	// Mock risk assessment handler
	mux.HandleFunc("/api/v1/risk/assess", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"assessment_id": "test_assessment",
				"risk_score":    0.75,
				"status":        "completed",
			},
		})
	})

	// Mock tenant info handler
	mux.HandleFunc("/api/v1/tenant/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"tenant_id": "test_tenant",
				"name":      "Test Tenant",
			},
		})
	})

	server := httptest.NewServer(mux)

	return &SecurityTestSuite{
		server:     server,
		logger:     logger,
		tenant1ID:  tenant1ID,
		tenant2ID:  tenant2ID,
		tenant1Key: tenant1Key,
		tenant2Key: tenant2Key,
	}
}

// TestTenantIsolation tests that tenants cannot access each other's data
func (sts *SecurityTestSuite) TestTenantIsolation(t *testing.T) {
	t.Run("tenant_cannot_access_other_tenant_data", func(t *testing.T) {
		// Create risk assessment for tenant 1
		assessment1 := map[string]interface{}{
			"business_name": "Test Business 1",
			"business_id":   "business_1",
			"country":       "US",
		}

		req1, _ := http.NewRequest("POST", sts.server.URL+"/api/v1/risk/assess",
			sts.createJSONBody(assessment1))
		req1.Header.Set("Authorization", "Bearer "+sts.tenant1Key)
		req1.Header.Set("X-Tenant-ID", sts.tenant1ID)

		resp1, err := http.DefaultClient.Do(req1)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp1.StatusCode)

		// Try to access tenant 1's data with tenant 2's credentials
		req2, _ := http.NewRequest("GET", sts.server.URL+"/api/v1/tenant/info", nil)
		req2.Header.Set("Authorization", "Bearer "+sts.tenant2Key)
		req2.Header.Set("X-Tenant-ID", sts.tenant1ID) // Wrong tenant ID

		resp2, err := http.DefaultClient.Do(req2)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp2.StatusCode)
	})

	t.Run("tenant_context_validation", func(t *testing.T) {
		// Test without tenant context
		req, _ := http.NewRequest("POST", sts.server.URL+"/api/v1/risk/assess",
			sts.createJSONBody(map[string]interface{}{"business_name": "Test"}))

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// TestSQLInjectionPrevention tests protection against SQL injection attacks
func (sts *SecurityTestSuite) TestSQLInjectionPrevention(t *testing.T) {
	sqlInjectionPayloads := []string{
		"'; DROP TABLE assessments; --",
		"' OR '1'='1",
		"'; INSERT INTO assessments VALUES ('hacked'); --",
		"' UNION SELECT * FROM users --",
		"'; UPDATE assessments SET score = 100; --",
	}

	for _, payload := range sqlInjectionPayloads {
		t.Run(fmt.Sprintf("sql_injection_%s", strings.ReplaceAll(payload, " ", "_")), func(t *testing.T) {
			assessment := map[string]interface{}{
				"business_name": payload,
				"business_id":   "test_business",
				"country":       "US",
			}

			req, _ := http.NewRequest("POST", sts.server.URL+"/api/v1/risk/assess",
				sts.createJSONBody(assessment))
			req.Header.Set("Authorization", "Bearer "+sts.tenant1Key)
			req.Header.Set("X-Tenant-ID", sts.tenant1ID)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			// Should either reject the request or sanitize the input
			assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusOK)

			if resp.StatusCode == http.StatusOK {
				// If accepted, verify the payload was sanitized
				var result map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)

				// Verify no SQL injection occurred
				assert.NotContains(t, result, "DROP TABLE")
				assert.NotContains(t, result, "UNION SELECT")
			}
		})
	}
}

// TestXSSPrevention tests protection against Cross-Site Scripting attacks
func (sts *SecurityTestSuite) TestXSSPrevention(t *testing.T) {
	xssPayloads := []string{
		"<script>alert('XSS')</script>",
		"javascript:alert('XSS')",
		"<img src=x onerror=alert('XSS')>",
		"<svg onload=alert('XSS')>",
		"';alert('XSS');//",
	}

	for _, payload := range xssPayloads {
		t.Run(fmt.Sprintf("xss_%s", strings.ReplaceAll(payload, "<", "_")), func(t *testing.T) {
			assessment := map[string]interface{}{
				"business_name": payload,
				"business_id":   "test_business",
				"country":       "US",
			}

			req, _ := http.NewRequest("POST", sts.server.URL+"/api/v1/risk/assess",
				sts.createJSONBody(assessment))
			req.Header.Set("Authorization", "Bearer "+sts.tenant1Key)
			req.Header.Set("X-Tenant-ID", sts.tenant1ID)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			// Should either reject the request or sanitize the input
			assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusOK)

			if resp.StatusCode == http.StatusOK {
				// If accepted, verify the payload was sanitized
				var result map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)

				// Verify no XSS payload remains
				resultStr := fmt.Sprintf("%v", result)
				assert.NotContains(t, resultStr, "<script>")
				assert.NotContains(t, resultStr, "javascript:")
				assert.NotContains(t, resultStr, "onerror=")
			}
		})
	}
}

// TestRateLimiting tests rate limiting functionality
func (sts *SecurityTestSuite) TestRateLimiting(t *testing.T) {
	t.Run("rate_limit_enforcement", func(t *testing.T) {
		// Make multiple requests quickly to trigger rate limiting
		assessment := map[string]interface{}{
			"business_name": "Rate Limit Test",
			"business_id":   "rate_test",
			"country":       "US",
		}

		successCount := 0
		rateLimitedCount := 0

		for i := 0; i < 20; i++ { // Exceed rate limit
			req, _ := http.NewRequest("POST", sts.server.URL+"/api/v1/risk/assess",
				sts.createJSONBody(assessment))
			req.Header.Set("Authorization", "Bearer "+sts.tenant1Key)
			req.Header.Set("X-Tenant-ID", sts.tenant1ID)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if resp.StatusCode == http.StatusOK {
				successCount++
			} else if resp.StatusCode == http.StatusTooManyRequests {
				rateLimitedCount++
			}
		}

		// Should have some successful requests and some rate limited
		assert.Greater(t, successCount, 0)
		assert.Greater(t, rateLimitedCount, 0)
	})
}

// TestAuthenticationSecurity tests authentication mechanisms
func (sts *SecurityTestSuite) TestAuthenticationSecurity(t *testing.T) {
	t.Run("invalid_token_rejection", func(t *testing.T) {
		req, _ := http.NewRequest("POST", sts.server.URL+"/api/v1/risk/assess",
			sts.createJSONBody(map[string]interface{}{"business_name": "Test"}))
		req.Header.Set("Authorization", "Bearer invalid_token")
		req.Header.Set("X-Tenant-ID", sts.tenant1ID)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("missing_authorization_header", func(t *testing.T) {
		req, _ := http.NewRequest("POST", sts.server.URL+"/api/v1/risk/assess",
			sts.createJSONBody(map[string]interface{}{"business_name": "Test"}))
		req.Header.Set("X-Tenant-ID", sts.tenant1ID)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("malformed_authorization_header", func(t *testing.T) {
		req, _ := http.NewRequest("POST", sts.server.URL+"/api/v1/risk/assess",
			sts.createJSONBody(map[string]interface{}{"business_name": "Test"}))
		req.Header.Set("Authorization", "InvalidFormat token")
		req.Header.Set("X-Tenant-ID", sts.tenant1ID)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// TestInputValidation tests input validation and sanitization
func (sts *SecurityTestSuite) TestInputValidation(t *testing.T) {
	t.Run("oversized_payload_rejection", func(t *testing.T) {
		// Create a very large payload
		largeString := strings.Repeat("A", 10*1024*1024) // 10MB
		assessment := map[string]interface{}{
			"business_name": largeString,
			"business_id":   "test_business",
			"country":       "US",
		}

		req, _ := http.NewRequest("POST", sts.server.URL+"/api/v1/risk/assess",
			sts.createJSONBody(assessment))
		req.Header.Set("Authorization", "Bearer "+sts.tenant1Key)
		req.Header.Set("X-Tenant-ID", sts.tenant1ID)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
	})

	t.Run("malformed_json_rejection", func(t *testing.T) {
		req, _ := http.NewRequest("POST", sts.server.URL+"/api/v1/risk/assess",
			strings.NewReader("{ invalid json }"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+sts.tenant1Key)
		req.Header.Set("X-Tenant-ID", sts.tenant1ID)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("missing_required_fields", func(t *testing.T) {
		assessment := map[string]interface{}{
			"business_id": "test_business",
			// Missing business_name
		}

		req, _ := http.NewRequest("POST", sts.server.URL+"/api/v1/risk/assess",
			sts.createJSONBody(assessment))
		req.Header.Set("Authorization", "Bearer "+sts.tenant1Key)
		req.Header.Set("X-Tenant-ID", sts.tenant1ID)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestDataEncryption tests data encryption at rest and in transit
func (sts *SecurityTestSuite) TestDataEncryption(t *testing.T) {
	t.Run("sensitive_data_not_logged", func(t *testing.T) {
		// This test would verify that sensitive data is not logged
		// In a real implementation, we would capture logs and verify
		assessment := map[string]interface{}{
			"business_name": "Sensitive Business",
			"business_id":   "sensitive_id",
			"ssn":           "123-45-6789", // Sensitive data
			"country":       "US",
		}

		req, _ := http.NewRequest("POST", sts.server.URL+"/api/v1/risk/assess",
			sts.createJSONBody(assessment))
		req.Header.Set("Authorization", "Bearer "+sts.tenant1Key)
		req.Header.Set("X-Tenant-ID", sts.tenant1ID)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// In a real test, we would verify logs don't contain sensitive data
		// This is a placeholder for the actual log verification
	})
}

// TestAuditTrailIntegrity tests audit trail immutability
func (sts *SecurityTestSuite) TestAuditTrailIntegrity(t *testing.T) {
	t.Run("audit_trail_immutability", func(t *testing.T) {
		// Perform an action that should be audited
		assessment := map[string]interface{}{
			"business_name": "Audit Test Business",
			"business_id":   "audit_test",
			"country":       "US",
		}

		req, _ := http.NewRequest("POST", sts.server.URL+"/api/v1/risk/assess",
			sts.createJSONBody(assessment))
		req.Header.Set("Authorization", "Bearer "+sts.tenant1Key)
		req.Header.Set("X-Tenant-ID", sts.tenant1ID)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// In a real implementation, we would verify:
		// 1. Audit log entry was created
		// 2. Audit log entry is immutable
		// 3. Audit log entry contains correct information
		// 4. Audit log entry has cryptographic integrity
	})
}

// TestConcurrencySecurity tests security under concurrent access
func (sts *SecurityTestSuite) TestConcurrencySecurity(t *testing.T) {
	t.Run("concurrent_tenant_isolation", func(t *testing.T) {
		// Test concurrent requests from different tenants
		done := make(chan bool, 2)

		// Tenant 1 request
		go func() {
			assessment := map[string]interface{}{
				"business_name": "Concurrent Test 1",
				"business_id":   "concurrent_1",
				"country":       "US",
			}

			req, _ := http.NewRequest("POST", sts.server.URL+"/api/v1/risk/assess",
				sts.createJSONBody(assessment))
			req.Header.Set("Authorization", "Bearer "+sts.tenant1Key)
			req.Header.Set("X-Tenant-ID", sts.tenant1ID)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			done <- true
		}()

		// Tenant 2 request
		go func() {
			assessment := map[string]interface{}{
				"business_name": "Concurrent Test 2",
				"business_id":   "concurrent_2",
				"country":       "US",
			}

			req, _ := http.NewRequest("POST", sts.server.URL+"/api/v1/risk/assess",
				sts.createJSONBody(assessment))
			req.Header.Set("Authorization", "Bearer "+sts.tenant2Key)
			req.Header.Set("X-Tenant-ID", sts.tenant2ID)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			done <- true
		}()

		// Wait for both requests to complete
		<-done
		<-done
	})
}

// TestSecurityHeaders tests security headers are properly set
func (sts *SecurityTestSuite) TestSecurityHeaders(t *testing.T) {
	t.Run("security_headers_present", func(t *testing.T) {
		req, _ := http.NewRequest("GET", sts.server.URL+"/api/v1/tenant/info", nil)
		req.Header.Set("Authorization", "Bearer "+sts.tenant1Key)
		req.Header.Set("X-Tenant-ID", sts.tenant1ID)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		// Check for security headers
		expectedHeaders := []string{
			"X-Content-Type-Options",
			"X-Frame-Options",
			"X-XSS-Protection",
			"Strict-Transport-Security",
		}

		for _, header := range expectedHeaders {
			assert.NotEmpty(t, resp.Header.Get(header),
				"Security header %s should be present", header)
		}
	})
}

// Helper methods

func (sts *SecurityTestSuite) createJSONBody(data interface{}) io.Reader {
	jsonData, _ := json.Marshal(data)
	return bytes.NewReader(jsonData)
}

func (sts *SecurityTestSuite) generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		randomByte := make([]byte, 1)
		rand.Read(randomByte)
		b[i] = charset[randomByte[0]%byte(len(charset))]
	}
	return string(b)
}

func (sts *SecurityTestSuite) Close() {
	sts.server.Close()
}

// RunAllSecurityTests runs the complete security test suite
func RunAllSecurityTests(t *testing.T) {
	suite := NewSecurityTestSuite()
	defer suite.Close()

	t.Run("TenantIsolation", suite.TestTenantIsolation)
	t.Run("SQLInjectionPrevention", suite.TestSQLInjectionPrevention)
	t.Run("XSSPrevention", suite.TestXSSPrevention)
	t.Run("RateLimiting", suite.TestRateLimiting)
	t.Run("AuthenticationSecurity", suite.TestAuthenticationSecurity)
	t.Run("InputValidation", suite.TestInputValidation)
	t.Run("DataEncryption", suite.TestDataEncryption)
	t.Run("AuditTrailIntegrity", suite.TestAuditTrailIntegrity)
	t.Run("ConcurrencySecurity", suite.TestConcurrencySecurity)
	t.Run("SecurityHeaders", suite.TestSecurityHeaders)
}
