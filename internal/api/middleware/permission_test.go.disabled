package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pcraw4d/business-verification/internal/auth"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// MockRBACService provides a mock implementation for testing
type MockRBACService struct {
	checkPermissionFunc func(ctx context.Context, role auth.Role, method, path string) error
}

func (m *MockRBACService) CheckPermission(ctx context.Context, role auth.Role, method, path string) error {
	if m.checkPermissionFunc != nil {
		return m.checkPermissionFunc(ctx, role, method, path)
	}
	return nil
}

func setupPermissionMiddlewareTest() (*PermissionMiddleware, *MockRBACService) {
	mockRBAC := &MockRBACService{}
	loggerConfig := &config.ObservabilityConfig{
		LogLevel: "debug",
	}
	logger := observability.NewLogger(loggerConfig)

	permissionMiddleware := NewPermissionMiddleware(mockRBAC, logger)

	return permissionMiddleware, mockRBAC
}

func TestPermissionMiddleware_PublicEndpoints(t *testing.T) {
	pm, _ := setupPermissionMiddlewareTest()

	tests := []struct {
		method string
		path   string
		expect bool
	}{
		{"GET", "/health", true},
		{"GET", "/v1/health", true},
		{"GET", "/docs", true},
		{"GET", "/docs/", true},
		{"POST", "/v1/auth/login", true},
		{"POST", "/v1/auth/register", true},
		{"GET", "/v1/classify", false},
		{"POST", "/v1/classify", false},
		{"GET", "/v1/users", false},
	}

	for _, test := range tests {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			result := pm.isPublicEndpoint(test.method, test.path)
			if result != test.expect {
				t.Errorf("Expected %v for %s %s, got %v", test.expect, test.method, test.path, result)
			}
		})
	}
}

func TestPermissionMiddleware_ExtractUserContext(t *testing.T) {
	pm, _ := setupPermissionMiddlewareTest()

	tests := []struct {
		name           string
		headers        map[string]string
		expectedUserID string
		expectedRole   auth.Role
		expectError    bool
	}{
		{
			name: "valid admin token",
			headers: map[string]string{
				"Authorization": "Bearer admin-token",
			},
			expectedUserID: "admin-user",
			expectedRole:   auth.RoleAdmin,
			expectError:    false,
		},
		{
			name: "valid user token",
			headers: map[string]string{
				"Authorization": "Bearer user-token",
			},
			expectedUserID: "user-1",
			expectedRole:   auth.RoleUser,
			expectError:    false,
		},
		{
			name: "valid API key",
			headers: map[string]string{
				"X-API-Key": "system-api-key",
			},
			expectedUserID: "system-user",
			expectedRole:   auth.RoleSystem,
			expectError:    false,
		},
		{
			name: "system token",
			headers: map[string]string{
				"X-System-Token": "internal-system",
			},
			expectedUserID: "system-internal",
			expectedRole:   auth.RoleSystem,
			expectError:    false,
		},
		{
			name: "monitoring user agent",
			headers: map[string]string{
				"User-Agent": "health-check-agent",
			},
			expectedUserID: "monitoring-system",
			expectedRole:   auth.RoleSystem,
			expectError:    false,
		},
		{
			name:        "no authentication",
			headers:     map[string]string{},
			expectError: true,
		},
		{
			name: "invalid token",
			headers: map[string]string{
				"Authorization": "Bearer invalid-token",
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			for key, value := range test.headers {
				req.Header.Set(key, value)
			}

			userID, userRole, err := pm.extractUserContext(req)

			if test.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
				return
			}

			if userID != test.expectedUserID {
				t.Errorf("Expected user ID %s, got %s", test.expectedUserID, userID)
			}

			if userRole != test.expectedRole {
				t.Errorf("Expected role %s, got %s", test.expectedRole, userRole)
			}
		})
	}
}

func TestPermissionMiddleware_Middleware(t *testing.T) {
	pm, mockRBAC := setupPermissionMiddlewareTest()

	// Test successful permission check
	mockRBAC.checkPermissionFunc = func(ctx context.Context, role auth.Role, method, path string) error {
		return nil // Allow all requests
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := pm.Middleware(handler)

	// Test public endpoint (should pass without authentication)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test protected endpoint with valid token
	req = httptest.NewRequest("GET", "/v1/classify", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	w = httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test protected endpoint without authentication
	req = httptest.NewRequest("GET", "/v1/classify", nil)
	w = httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	// Test protected endpoint with permission denied
	mockRBAC.checkPermissionFunc = func(ctx context.Context, role auth.Role, method, path string) error {
		return fmt.Errorf("permission denied")
	}

	req = httptest.NewRequest("GET", "/v1/classify", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	w = httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestPermissionMiddleware_ContextHelpers(t *testing.T) {
	ctx := context.Background()

	// Test GetUserIDFromContext
	userID, ok := GetUserIDFromContext(ctx)
	if ok {
		t.Error("Expected no user ID in empty context")
	}

	ctx = context.WithValue(ctx, "user_id", "test-user")
	userID, ok = GetUserIDFromContext(ctx)
	if !ok {
		t.Error("Expected to find user ID in context")
	}
	if userID != "test-user" {
		t.Errorf("Expected user ID 'test-user', got %s", userID)
	}

	// Test GetUserRoleFromContext
	userRole, ok := GetUserRoleFromContext(ctx)
	if ok {
		t.Error("Expected no user role in context without role")
	}

	ctx = context.WithValue(ctx, "user_role", auth.RoleAdmin)
	userRole, ok = GetUserRoleFromContext(ctx)
	if !ok {
		t.Error("Expected to find user role in context")
	}
	if userRole != auth.RoleAdmin {
		t.Errorf("Expected role %s, got %s", auth.RoleAdmin, userRole)
	}
}

func TestPermissionMiddleware_RequirePermission(t *testing.T) {
	pm, _ := setupPermissionMiddlewareTest()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test with user that has permission
	ctx := context.WithValue(context.Background(), "user_role", auth.RoleAdmin)
	req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	permissionMiddleware := pm.RequirePermission(auth.PermissionClassifyBusiness)
	permissionMiddleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test with user that doesn't have permission
	ctx = context.WithValue(context.Background(), "user_role", auth.RoleGuest)
	req = httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
	w = httptest.NewRecorder()

	permissionMiddleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestPermissionMiddleware_RequireRole(t *testing.T) {
	pm, _ := setupPermissionMiddlewareTest()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test with correct role
	ctx := context.WithValue(context.Background(), "user_role", auth.RoleAdmin)
	req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	roleMiddleware := pm.RequireRole(auth.RoleAdmin)
	roleMiddleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test with incorrect role
	ctx = context.WithValue(context.Background(), "user_role", auth.RoleUser)
	req = httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
	w = httptest.NewRecorder()

	roleMiddleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestPermissionMiddleware_RequireMinimumRole(t *testing.T) {
	pm, _ := setupPermissionMiddlewareTest()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		userRole       auth.Role
		minimumRole    auth.Role
		expectedStatus int
	}{
		{"admin meets admin requirement", auth.RoleAdmin, auth.RoleAdmin, http.StatusOK},
		{"admin meets user requirement", auth.RoleAdmin, auth.RoleUser, http.StatusOK},
		{"user meets user requirement", auth.RoleUser, auth.RoleUser, http.StatusOK},
		{"user meets guest requirement", auth.RoleUser, auth.RoleGuest, http.StatusOK},
		{"guest doesn't meet user requirement", auth.RoleGuest, auth.RoleUser, http.StatusForbidden},
		{"user doesn't meet admin requirement", auth.RoleUser, auth.RoleAdmin, http.StatusForbidden},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "user_role", test.userRole)
			req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
			w := httptest.NewRecorder()

			minimumRoleMiddleware := pm.RequireMinimumRole(test.minimumRole)
			minimumRoleMiddleware(handler).ServeHTTP(w, req)

			if w.Code != test.expectedStatus {
				t.Errorf("Expected status %d, got %d", test.expectedStatus, w.Code)
			}
		})
	}
}

func TestPermissionMiddleware_HasMinimumRole(t *testing.T) {
	pm, _ := setupPermissionMiddlewareTest()

	tests := []struct {
		name        string
		userRole    auth.Role
		minimumRole auth.Role
		expected    bool
	}{
		{"guest >= guest", auth.RoleGuest, auth.RoleGuest, true},
		{"user >= guest", auth.RoleUser, auth.RoleGuest, true},
		{"user >= user", auth.RoleUser, auth.RoleUser, true},
		{"admin >= user", auth.RoleAdmin, auth.RoleUser, true},
		{"guest < user", auth.RoleGuest, auth.RoleUser, false},
		{"user < admin", auth.RoleUser, auth.RoleAdmin, false},
		{"analyst >= user", auth.RoleAnalyst, auth.RoleUser, true},
		{"manager >= analyst", auth.RoleManager, auth.RoleAnalyst, true},
		{"system >= admin", auth.RoleSystem, auth.RoleAdmin, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := pm.hasMinimumRole(test.userRole, test.minimumRole)
			if result != test.expected {
				t.Errorf("Expected %v for %s >= %s, got %v", test.expected, test.userRole, test.minimumRole, result)
			}
		})
	}
}
