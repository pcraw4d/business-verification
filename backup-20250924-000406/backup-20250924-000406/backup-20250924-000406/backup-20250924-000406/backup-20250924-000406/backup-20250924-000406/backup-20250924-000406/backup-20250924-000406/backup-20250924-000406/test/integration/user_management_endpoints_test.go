package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/test/mocks"
)

// TestUserManagementEndpoints tests all user management API endpoints
func TestUserManagementEndpoints(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests - set INTEGRATION_TESTS=true to run")
	}

	// Setup test environment
	suite := setupUserManagementEndpointTestSuite(t)
	defer suite.cleanup()

	// Test authentication endpoints
	t.Run("AuthenticationEndpoints", suite.testAuthenticationEndpoints)

	// Test profile management endpoints
	t.Run("ProfileManagementEndpoints", suite.testProfileManagementEndpoints)

	// Test API key management endpoints
	t.Run("APIKeyManagementEndpoints", suite.testAPIKeyManagementEndpoints)

	// Test user permissions endpoints
	t.Run("UserPermissionsEndpoints", suite.testUserPermissionsEndpoints)
}

// UserManagementEndpointTestSuite provides testing for user management endpoints
type UserManagementEndpointTestSuite struct {
	server  *httptest.Server
	mux     *http.ServeMux
	logger  *observability.Logger
	cleanup func()
}

// setupUserManagementEndpointTestSuite sets up the user management endpoint test suite
func setupUserManagementEndpointTestSuite(t *testing.T) *UserManagementEndpointTestSuite {
	// Setup logger
	logger := observability.NewLogger(&observability.Config{
		LogLevel:  "debug",
		LogFormat: "json",
	})

	// Create main mux
	mux := http.NewServeMux()

	// Setup mock services
	mockAuthService := mocks.NewMockAuthService()
	mockUserService := mocks.NewMockUserService()
	mockPermissionService := mocks.NewMockPermissionService()

	// Setup handlers
	authHandler := handlers.NewAuthHandler(mockAuthService, logger)
	userHandler := handlers.NewUserHandler(mockUserService, logger)
	permissionHandler := handlers.NewPermissionHandler(mockPermissionService, logger)

	// Register user management routes
	registerUserManagementRoutes(mux, authHandler, userHandler, permissionHandler)

	// Create test server
	server := httptest.NewServer(mux)

	return &UserManagementEndpointTestSuite{
		server:  server,
		mux:     mux,
		logger:  logger,
		cleanup: func() { server.Close() },
	}
}

// registerUserManagementRoutes registers all user management API routes
func registerUserManagementRoutes(
	mux *http.ServeMux,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	permissionHandler *handlers.PermissionHandler,
) {
	// Authentication endpoints
	mux.HandleFunc("POST /v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /v1/auth/logout", authHandler.Logout)
	mux.HandleFunc("POST /v1/auth/refresh", authHandler.RefreshToken)
	mux.HandleFunc("POST /v1/auth/forgot-password", authHandler.ForgotPassword)
	mux.HandleFunc("POST /v1/auth/reset-password", authHandler.ResetPassword)
	mux.HandleFunc("POST /v1/auth/verify-email", authHandler.VerifyEmail)

	// Profile management endpoints
	mux.HandleFunc("GET /v1/users/profile", userHandler.GetProfile)
	mux.HandleFunc("PUT /v1/users/profile", userHandler.UpdateProfile)
	mux.HandleFunc("POST /v1/users/profile/avatar", userHandler.UploadAvatar)
	mux.HandleFunc("DELETE /v1/users/profile/avatar", userHandler.DeleteAvatar)
	mux.HandleFunc("GET /v1/users/profile/preferences", userHandler.GetPreferences)
	mux.HandleFunc("PUT /v1/users/profile/preferences", userHandler.UpdatePreferences)

	// API key management endpoints
	mux.HandleFunc("GET /v1/users/api-keys", userHandler.GetAPIKeys)
	mux.HandleFunc("POST /v1/users/api-keys", userHandler.CreateAPIKey)
	mux.HandleFunc("PUT /v1/users/api-keys/{key_id}", userHandler.UpdateAPIKey)
	mux.HandleFunc("DELETE /v1/users/api-keys/{key_id}", userHandler.DeleteAPIKey)
	mux.HandleFunc("POST /v1/users/api-keys/{key_id}/regenerate", userHandler.RegenerateAPIKey)

	// User permissions endpoints
	mux.HandleFunc("GET /v1/users/permissions", permissionHandler.GetUserPermissions)
	mux.HandleFunc("GET /v1/users/roles", permissionHandler.GetUserRoles)
	mux.HandleFunc("POST /v1/users/roles", permissionHandler.AssignRole)
	mux.HandleFunc("DELETE /v1/users/roles/{role_id}", permissionHandler.RemoveRole)
}

// testAuthenticationEndpoints tests all authentication endpoints
func (suite *UserManagementEndpointTestSuite) testAuthenticationEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// User registration
		{
			name:   "POST /v1/auth/register - User registration",
			method: "POST",
			path:   "/v1/auth/register",
			body: map[string]interface{}{
				"email":     "test@example.com",
				"password":  "securepassword123",
				"full_name": "Test User",
				"company":   "Test Company Inc",
			},
			expectedStatus: http.StatusCreated,
			description:    "Should register a new user successfully",
		},
		{
			name:   "POST /v1/auth/register - User registration with minimal data",
			method: "POST",
			path:   "/v1/auth/register",
			body: map[string]interface{}{
				"email":    "minimal@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusCreated,
			description:    "Should register user with minimal required data",
		},
		{
			name:   "POST /v1/auth/register - User registration with company info",
			method: "POST",
			path:   "/v1/auth/register",
			body: map[string]interface{}{
				"email":     "company@example.com",
				"password":  "securepassword123",
				"full_name": "Company User",
				"company":   "Enterprise Corp",
				"job_title": "Software Engineer",
				"phone":     "+1-555-123-4567",
			},
			expectedStatus: http.StatusCreated,
			description:    "Should register user with complete company information",
		},

		// User login
		{
			name:   "POST /v1/auth/login - User login with email",
			method: "POST",
			path:   "/v1/auth/login",
			body: map[string]interface{}{
				"email":    "test@example.com",
				"password": "securepassword123",
			},
			expectedStatus: http.StatusOK,
			description:    "Should authenticate user and return tokens",
		},
		{
			name:   "POST /v1/auth/login - User login with remember me",
			method: "POST",
			path:   "/v1/auth/login",
			body: map[string]interface{}{
				"email":       "test@example.com",
				"password":    "securepassword123",
				"remember_me": true,
			},
			expectedStatus: http.StatusOK,
			description:    "Should authenticate user with extended session",
		},

		// Token refresh
		{
			name:   "POST /v1/auth/refresh - Token refresh",
			method: "POST",
			path:   "/v1/auth/refresh",
			body: map[string]interface{}{
				"refresh_token": "valid-refresh-token",
			},
			expectedStatus: http.StatusOK,
			description:    "Should refresh access token successfully",
		},

		// User logout
		{
			name:   "POST /v1/auth/logout - User logout",
			method: "POST",
			path:   "/v1/auth/logout",
			body: map[string]interface{}{
				"access_token": "valid-access-token",
			},
			expectedStatus: http.StatusOK,
			description:    "Should logout user and invalidate tokens",
		},
		{
			name:   "POST /v1/auth/logout - User logout all sessions",
			method: "POST",
			path:   "/v1/auth/logout",
			body: map[string]interface{}{
				"access_token": "valid-access-token",
				"all_sessions": true,
			},
			expectedStatus: http.StatusOK,
			description:    "Should logout user from all sessions",
		},

		// Password reset
		{
			name:   "POST /v1/auth/forgot-password - Forgot password",
			method: "POST",
			path:   "/v1/auth/forgot-password",
			body: map[string]interface{}{
				"email": "test@example.com",
			},
			expectedStatus: http.StatusOK,
			description:    "Should send password reset email",
		},
		{
			name:   "POST /v1/auth/reset-password - Reset password",
			method: "POST",
			path:   "/v1/auth/reset-password",
			body: map[string]interface{}{
				"token":        "valid-reset-token",
				"new_password": "newsecurepassword123",
			},
			expectedStatus: http.StatusOK,
			description:    "Should reset password successfully",
		},

		// Email verification
		{
			name:   "POST /v1/auth/verify-email - Verify email",
			method: "POST",
			path:   "/v1/auth/verify-email",
			body: map[string]interface{}{
				"token": "valid-verification-token",
			},
			expectedStatus: http.StatusOK,
			description:    "Should verify email address successfully",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testEndpoint(t, tc.method, tc.path, tc.body, tc.expectedStatus, tc.description)
		})
	}
}

// testProfileManagementEndpoints tests all profile management endpoints
func (suite *UserManagementEndpointTestSuite) testProfileManagementEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// Profile retrieval
		{
			name:           "GET /v1/users/profile - Get user profile",
			method:         "GET",
			path:           "/v1/users/profile",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve user profile information",
		},

		// Profile updates
		{
			name:   "PUT /v1/users/profile - Update user profile",
			method: "PUT",
			path:   "/v1/users/profile",
			body: map[string]interface{}{
				"full_name": "Updated Test User",
				"company":   "Updated Test Company Inc",
				"job_title": "Senior Software Engineer",
				"phone":     "+1-555-987-6543",
				"bio":       "Experienced software engineer with expertise in Go and microservices",
			},
			expectedStatus: http.StatusOK,
			description:    "Should update user profile information",
		},
		{
			name:   "PUT /v1/users/profile - Partial profile update",
			method: "PUT",
			path:   "/v1/users/profile",
			body: map[string]interface{}{
				"phone": "+1-555-111-2222",
			},
			expectedStatus: http.StatusOK,
			description:    "Should perform partial update of user profile",
		},

		// Avatar management
		{
			name:   "POST /v1/users/profile/avatar - Upload avatar",
			method: "POST",
			path:   "/v1/users/profile/avatar",
			body: map[string]interface{}{
				"avatar_data": "base64-encoded-image-data",
				"file_type":   "image/jpeg",
			},
			expectedStatus: http.StatusOK,
			description:    "Should upload user avatar successfully",
		},
		{
			name:           "DELETE /v1/users/profile/avatar - Delete avatar",
			method:         "DELETE",
			path:           "/v1/users/profile/avatar",
			expectedStatus: http.StatusOK,
			description:    "Should delete user avatar successfully",
		},

		// User preferences
		{
			name:           "GET /v1/users/profile/preferences - Get user preferences",
			method:         "GET",
			path:           "/v1/users/profile/preferences",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve user preferences",
		},
		{
			name:   "PUT /v1/users/profile/preferences - Update user preferences",
			method: "PUT",
			path:   "/v1/users/profile/preferences",
			body: map[string]interface{}{
				"theme":    "dark",
				"language": "en",
				"timezone": "UTC",
				"notifications": map[string]interface{}{
					"email":     true,
					"push":      false,
					"marketing": false,
				},
				"dashboard": map[string]interface{}{
					"default_view": "analytics",
					"widgets":      []string{"metrics", "alerts", "recent_activity"},
				},
			},
			expectedStatus: http.StatusOK,
			description:    "Should update user preferences",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testEndpoint(t, tc.method, tc.path, tc.body, tc.expectedStatus, tc.description)
		})
	}
}

// testAPIKeyManagementEndpoints tests all API key management endpoints
func (suite *UserManagementEndpointTestSuite) testAPIKeyManagementEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// API key retrieval
		{
			name:           "GET /v1/users/api-keys - Get API keys",
			method:         "GET",
			path:           "/v1/users/api-keys",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve user's API keys",
		},
		{
			name:           "GET /v1/users/api-keys - Get API keys with filters",
			method:         "GET",
			path:           "/v1/users/api-keys?status=active&limit=10",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve filtered API keys",
		},

		// API key creation
		{
			name:   "POST /v1/users/api-keys - Create API key",
			method: "POST",
			path:   "/v1/users/api-keys",
			body: map[string]interface{}{
				"name":        "Test API Key",
				"description": "API key for testing purposes",
				"permissions": []string{"classify", "read", "monitor"},
				"expires_at":  "2025-12-31T23:59:59Z",
			},
			expectedStatus: http.StatusCreated,
			description:    "Should create a new API key",
		},
		{
			name:   "POST /v1/users/api-keys - Create API key with minimal data",
			method: "POST",
			path:   "/v1/users/api-keys",
			body: map[string]interface{}{
				"name": "Simple API Key",
			},
			expectedStatus: http.StatusCreated,
			description:    "Should create API key with minimal data",
		},
		{
			name:   "POST /v1/users/api-keys - Create API key with specific permissions",
			method: "POST",
			path:   "/v1/users/api-keys",
			body: map[string]interface{}{
				"name":        "Classification API Key",
				"description": "API key for classification operations only",
				"permissions": []string{"classify", "read"},
				"rate_limit": map[string]interface{}{
					"requests_per_minute": 100,
					"requests_per_hour":   1000,
				},
			},
			expectedStatus: http.StatusCreated,
			description:    "Should create API key with specific permissions and rate limits",
		},

		// API key updates
		{
			name:   "PUT /v1/users/api-keys/{key_id} - Update API key",
			method: "PUT",
			path:   "/v1/users/api-keys/test-key-123",
			body: map[string]interface{}{
				"name":        "Updated API Key",
				"description": "Updated description for API key",
				"permissions": []string{"classify", "read", "write"},
			},
			expectedStatus: http.StatusOK,
			description:    "Should update API key information",
		},
		{
			name:   "PUT /v1/users/api-keys/{key_id} - Update API key status",
			method: "PUT",
			path:   "/v1/users/api-keys/test-key-123",
			body: map[string]interface{}{
				"status": "inactive",
			},
			expectedStatus: http.StatusOK,
			description:    "Should update API key status",
		},

		// API key regeneration
		{
			name:   "POST /v1/users/api-keys/{key_id}/regenerate - Regenerate API key",
			method: "POST",
			path:   "/v1/users/api-keys/test-key-123/regenerate",
			body: map[string]interface{}{
				"keep_permissions": true,
			},
			expectedStatus: http.StatusOK,
			description:    "Should regenerate API key while keeping permissions",
		},
		{
			name:   "POST /v1/users/api-keys/{key_id}/regenerate - Regenerate with new permissions",
			method: "POST",
			path:   "/v1/users/api-keys/test-key-123/regenerate",
			body: map[string]interface{}{
				"permissions": []string{"classify", "read"},
			},
			expectedStatus: http.StatusOK,
			description:    "Should regenerate API key with new permissions",
		},

		// API key deletion
		{
			name:           "DELETE /v1/users/api-keys/{key_id} - Delete API key",
			method:         "DELETE",
			path:           "/v1/users/api-keys/test-key-123",
			expectedStatus: http.StatusNoContent,
			description:    "Should delete API key successfully",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testEndpoint(t, tc.method, tc.path, tc.body, tc.expectedStatus, tc.description)
		})
	}
}

// testUserPermissionsEndpoints tests all user permissions endpoints
func (suite *UserManagementEndpointTestSuite) testUserPermissionsEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// User permissions
		{
			name:           "GET /v1/users/permissions - Get user permissions",
			method:         "GET",
			path:           "/v1/users/permissions",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve user permissions",
		},
		{
			name:           "GET /v1/users/permissions - Get user permissions with details",
			method:         "GET",
			path:           "/v1/users/permissions?include_details=true",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve user permissions with detailed information",
		},

		// User roles
		{
			name:           "GET /v1/users/roles - Get user roles",
			method:         "GET",
			path:           "/v1/users/roles",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve user roles",
		},
		{
			name:           "GET /v1/users/roles - Get user roles with permissions",
			method:         "GET",
			path:           "/v1/users/roles?include_permissions=true",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve user roles with associated permissions",
		},

		// Role assignment
		{
			name:   "POST /v1/users/roles - Assign role to user",
			method: "POST",
			path:   "/v1/users/roles",
			body: map[string]interface{}{
				"role_id":     "admin-role-123",
				"assigned_by": "admin-user-456",
			},
			expectedStatus: http.StatusOK,
			description:    "Should assign role to user successfully",
		},
		{
			name:   "POST /v1/users/roles - Assign role with expiration",
			method: "POST",
			path:   "/v1/users/roles",
			body: map[string]interface{}{
				"role_id":     "temporary-role-123",
				"expires_at":  "2025-06-30T23:59:59Z",
				"assigned_by": "admin-user-456",
			},
			expectedStatus: http.StatusOK,
			description:    "Should assign temporary role to user",
		},

		// Role removal
		{
			name:           "DELETE /v1/users/roles/{role_id} - Remove role from user",
			method:         "DELETE",
			path:           "/v1/users/roles/admin-role-123",
			expectedStatus: http.StatusOK,
			description:    "Should remove role from user successfully",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testEndpoint(t, tc.method, tc.path, tc.body, tc.expectedStatus, tc.description)
		})
	}
}

// testEndpoint is a helper function to test individual endpoints
func (suite *UserManagementEndpointTestSuite) testEndpoint(t *testing.T, method, path string, body interface{}, expectedStatus int, description string) {
	var reqBody []byte
	var err error

	// Prepare request body if provided
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
	}

	// Create request
	req, err := http.NewRequest(method, suite.server.URL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add authentication header for protected endpoints
	if path != "/v1/auth/register" && path != "/v1/auth/login" && path != "/v1/auth/forgot-password" && path != "/v1/auth/reset-password" && path != "/v1/auth/verify-email" {
		req.Header.Set("Authorization", "Bearer test-token")
	}

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Validate response status
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d for %s %s: %s",
			expectedStatus, resp.StatusCode, method, path, description)
	}

	// Validate response headers
	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}

	// Log successful test
	suite.logger.Info("User management endpoint test completed", map[string]interface{}{
		"method":          method,
		"path":            path,
		"status":          resp.StatusCode,
		"expected_status": expectedStatus,
		"description":     description,
		"success":         resp.StatusCode == expectedStatus,
	})
}

// TestUserManagementEndpointPerformance tests user management endpoint performance
func TestUserManagementEndpointPerformance(t *testing.T) {
	// Skip if not running performance tests
	if os.Getenv("PERFORMANCE_TESTS") != "true" {
		t.Skip("Skipping performance tests - set PERFORMANCE_TESTS=true to run")
	}

	suite := setupUserManagementEndpointTestSuite(t)
	defer suite.cleanup()

	// Performance test cases
	performanceTests := []struct {
		name        string
		method      string
		path        string
		body        interface{}
		maxDuration time.Duration
		description string
	}{
		{
			name:        "User login performance",
			method:      "POST",
			path:        "/v1/auth/login",
			body:        map[string]interface{}{"email": "test@example.com", "password": "password123"},
			maxDuration: 1 * time.Second,
			description: "User login should complete within 1 second",
		},
		{
			name:        "Profile retrieval performance",
			method:      "GET",
			path:        "/v1/users/profile",
			maxDuration: 500 * time.Millisecond,
			description: "Profile retrieval should complete within 500ms",
		},
		{
			name:        "API key creation performance",
			method:      "POST",
			path:        "/v1/users/api-keys",
			body:        map[string]interface{}{"name": "Test Key", "permissions": []string{"read"}},
			maxDuration: 1 * time.Second,
			description: "API key creation should complete within 1 second",
		},
		{
			name:        "Token refresh performance",
			method:      "POST",
			path:        "/v1/auth/refresh",
			body:        map[string]interface{}{"refresh_token": "valid-token"},
			maxDuration: 500 * time.Millisecond,
			description: "Token refresh should complete within 500ms",
		},
	}

	for _, pt := range performanceTests {
		t.Run(pt.name, func(t *testing.T) {
			start := time.Now()
			suite.testEndpoint(t, pt.method, pt.path, pt.body, http.StatusOK, pt.description)
			duration := time.Since(start)

			if duration > pt.maxDuration {
				t.Errorf("Performance test failed: %s took %v, expected < %v",
					pt.description, duration, pt.maxDuration)
			}

			suite.logger.Info("Performance test completed", map[string]interface{}{
				"test":         pt.name,
				"duration":     duration,
				"max_duration": pt.maxDuration,
				"passed":       duration <= pt.maxDuration,
			})
		})
	}
}
