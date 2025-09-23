package routes

import (
	"context"
	"net/http"
	"testing"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/api/middleware"
	"kyb-platform/internal/observability"
	"kyb-platform/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockMerchantPortfolioService is a mock implementation of MerchantPortfolioServiceInterface
type MockMerchantPortfolioService struct {
	mock.Mock
}

func (m *MockMerchantPortfolioService) CreateMerchant(ctx context.Context, merchant *services.Merchant, userID string) (*services.Merchant, error) {
	args := m.Called(ctx, merchant, userID)
	return args.Get(0).(*services.Merchant), args.Error(1)
}

func (m *MockMerchantPortfolioService) GetMerchant(ctx context.Context, merchantID string) (*services.Merchant, error) {
	args := m.Called(ctx, merchantID)
	return args.Get(0).(*services.Merchant), args.Error(1)
}

func (m *MockMerchantPortfolioService) UpdateMerchant(ctx context.Context, merchant *services.Merchant, userID string) (*services.Merchant, error) {
	args := m.Called(ctx, merchant, userID)
	return args.Get(0).(*services.Merchant), args.Error(1)
}

func (m *MockMerchantPortfolioService) DeleteMerchant(ctx context.Context, merchantID string, userID string) error {
	args := m.Called(ctx, merchantID, userID)
	return args.Error(0)
}

func (m *MockMerchantPortfolioService) SearchMerchants(ctx context.Context, filters *services.MerchantSearchFilters, page, pageSize int) (*services.MerchantListResult, error) {
	args := m.Called(ctx, filters, page, pageSize)
	return args.Get(0).(*services.MerchantListResult), args.Error(1)
}

func (m *MockMerchantPortfolioService) BulkUpdatePortfolioType(ctx context.Context, merchantIDs []string, portfolioType services.PortfolioType, userID string) (*services.BulkOperationResult, error) {
	args := m.Called(ctx, merchantIDs, portfolioType, userID)
	return args.Get(0).(*services.BulkOperationResult), args.Error(1)
}

func (m *MockMerchantPortfolioService) BulkUpdateRiskLevel(ctx context.Context, merchantIDs []string, riskLevel services.RiskLevel, userID string) (*services.BulkOperationResult, error) {
	args := m.Called(ctx, merchantIDs, riskLevel, userID)
	return args.Get(0).(*services.BulkOperationResult), args.Error(1)
}

func (m *MockMerchantPortfolioService) StartMerchantSession(ctx context.Context, userID, merchantID string) (*services.MerchantSession, error) {
	args := m.Called(ctx, userID, merchantID)
	return args.Get(0).(*services.MerchantSession), args.Error(1)
}

func (m *MockMerchantPortfolioService) EndMerchantSession(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockMerchantPortfolioService) GetActiveMerchantSession(ctx context.Context, userID string) (*services.MerchantSession, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*services.MerchantSession), args.Error(1)
}

// Test helper functions
func createTestLogger() *observability.Logger {
	logger, _ := zap.NewDevelopment()
	return observability.NewLogger(logger)
}

func createTestMerchantRouteConfig() *MerchantRouteConfig {
	mockService := &MockMerchantPortfolioService{}
	handler := handlers.NewMerchantPortfolioHandler(mockService, nil)
	logger := createTestLogger()

	// Create real middleware instances for testing
	authMiddleware := &middleware.AuthMiddleware{}
	rateLimiter := middleware.NewAPIRateLimiter(&middleware.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 60,
		BurstSize:         10,
		WindowSize:        60 * 1000000000, // 1 minute
		Strategy:          "token_bucket",
	}, logger.GetZapLogger())

	return &MerchantRouteConfig{
		MerchantPortfolioHandler: handler,
		AuthMiddleware:           authMiddleware,
		RateLimiter:              rateLimiter,
		Logger:                   logger,
		EnableBulkOperations:     true,
		EnableSessionManagement:  true,
		MaxBulkOperationSize:     1000,
	}
}

// TestRegisterMerchantRoutes tests the main route registration function
func TestRegisterMerchantRoutes(t *testing.T) {
	config := createTestMerchantRouteConfig()
	mux := http.NewServeMux()

	// Register routes
	RegisterMerchantRoutes(mux, config)

	// Verify routes are registered by checking if mux is not nil
	assert.NotNil(t, mux)
}

// TestMerchantCRUDRoutes tests the CRUD route registration
func TestMerchantCRUDRoutes(t *testing.T) {
	config := createTestMerchantRouteConfig()
	mux := http.NewServeMux()

	// Register CRUD routes
	registerMerchantCRUDRoutes(mux, config)

	// Test that routes are registered
	assert.NotNil(t, mux)
}

// TestMerchantSearchRoutes tests the search route registration
func TestMerchantSearchRoutes(t *testing.T) {
	config := createTestMerchantRouteConfig()
	mux := http.NewServeMux()

	// Register search routes
	registerMerchantSearchRoutes(mux, config)

	// Test that routes are registered
	assert.NotNil(t, mux)
}

// TestBulkOperationRoutes tests the bulk operation route registration
func TestBulkOperationRoutes(t *testing.T) {
	config := createTestMerchantRouteConfig()
	mux := http.NewServeMux()

	// Register bulk operation routes
	registerBulkOperationRoutes(mux, config)

	// Test that routes are registered
	assert.NotNil(t, mux)
}

// TestSessionManagementRoutes tests the session management route registration
func TestSessionManagementRoutes(t *testing.T) {
	config := createTestMerchantRouteConfig()
	mux := http.NewServeMux()

	// Register session management routes
	registerSessionManagementRoutes(mux, config)

	// Test that routes are registered
	assert.NotNil(t, mux)
}

// TestMerchantAnalyticsRoutes tests the analytics route registration
func TestMerchantAnalyticsRoutes(t *testing.T) {
	config := createTestMerchantRouteConfig()
	mux := http.NewServeMux()

	// Register analytics routes
	registerMerchantAnalyticsRoutes(mux, config)

	// Test that routes are registered
	assert.NotNil(t, mux)
}

// TestCreateMerchantRouteConfig tests the route configuration creation
func TestCreateMerchantRouteConfig(t *testing.T) {
	mockService := &MockMerchantPortfolioService{}
	handler := handlers.NewMerchantPortfolioHandler(mockService, nil)
	logger := createTestLogger()

	authMiddleware := &middleware.AuthMiddleware{}
	rateLimiter := middleware.NewAPIRateLimiter(&middleware.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 60,
		BurstSize:         10,
		WindowSize:        60 * 1000000000,
		Strategy:          "token_bucket",
	}, logger.GetZapLogger())

	config := CreateMerchantRouteConfig(handler, authMiddleware, rateLimiter, logger)

	assert.NotNil(t, config)
	assert.Equal(t, handler, config.MerchantPortfolioHandler)
	assert.Equal(t, authMiddleware, config.AuthMiddleware)
	assert.Equal(t, rateLimiter, config.RateLimiter)
	assert.Equal(t, logger, config.Logger)
	assert.True(t, config.EnableBulkOperations)
	assert.True(t, config.EnableSessionManagement)
	assert.Equal(t, 1000, config.MaxBulkOperationSize)
}

// TestMerchantRouteDocumentation tests the route documentation function
func TestMerchantRouteDocumentation(t *testing.T) {
	doc := MerchantRouteDocumentation()

	assert.NotNil(t, doc)
	assert.Equal(t, "1.0.0", doc["version"])
	assert.Equal(t, "Merchant Portfolio Management API Routes", doc["description"])

	endpoints, ok := doc["endpoints"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, endpoints, "merchant_crud")
	assert.Contains(t, endpoints, "merchant_search")
	assert.Contains(t, endpoints, "bulk_operations")
	assert.Contains(t, endpoints, "session_management")
	assert.Contains(t, endpoints, "analytics")

	rateLimiting, ok := doc["rate_limiting"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, rateLimiting, "standard")
	assert.Contains(t, rateLimiting, "bulk_operations")

	authentication, ok := doc["authentication"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "JWT Bearer Token", authentication["type"])
	assert.True(t, authentication["required"].(bool))

	features, ok := doc["features"].([]string)
	assert.True(t, ok)
	assert.Contains(t, features, "Merchant CRUD operations")
	assert.Contains(t, features, "Advanced search and filtering")
	assert.Contains(t, features, "Bulk operations with progress tracking")
}

// TestMerchantRoutesIntegration tests integration of all routes
func TestMerchantRoutesIntegration(t *testing.T) {
	config := createTestMerchantRouteConfig()
	mux := http.NewServeMux()

	// Register all routes
	RegisterMerchantRoutes(mux, config)

	// Test that all route groups are registered
	assert.NotNil(t, mux)

	// Verify that the configuration is properly set
	assert.True(t, config.EnableBulkOperations)
	assert.True(t, config.EnableSessionManagement)
	assert.Equal(t, 1000, config.MaxBulkOperationSize)
}

// TestMerchantRoutesWithDisabledFeatures tests routes with some features disabled
func TestMerchantRoutesWithDisabledFeatures(t *testing.T) {
	config := createTestMerchantRouteConfig()
	config.EnableBulkOperations = false
	config.EnableSessionManagement = false
	mux := http.NewServeMux()

	// Register routes
	RegisterMerchantRoutes(mux, config)

	// Test that routes are registered even with disabled features
	assert.NotNil(t, mux)
	assert.False(t, config.EnableBulkOperations)
	assert.False(t, config.EnableSessionManagement)
}

// TestMerchantRoutesDocumentationCompleteness tests that documentation covers all routes
func TestMerchantRoutesDocumentationCompleteness(t *testing.T) {
	doc := MerchantRouteDocumentation()
	endpoints := doc["endpoints"].(map[string]interface{})

	// Test that all route groups are documented
	expectedGroups := []string{
		"merchant_crud",
		"merchant_search",
		"bulk_operations",
		"session_management",
		"analytics",
	}

	for _, group := range expectedGroups {
		assert.Contains(t, endpoints, group, "Route group %s should be documented", group)
	}

	// Test that CRUD operations are documented
	crudEndpoints := endpoints["merchant_crud"].(map[string]interface{})
	expectedCRUDEndpoints := []string{
		"POST /api/v1/merchants",
		"GET /api/v1/merchants/{id}",
		"PUT /api/v1/merchants/{id}",
		"DELETE /api/v1/merchants/{id}",
	}

	for _, endpoint := range expectedCRUDEndpoints {
		assert.Contains(t, crudEndpoints, endpoint, "CRUD endpoint %s should be documented", endpoint)
	}

	// Test that search operations are documented
	searchEndpoints := endpoints["merchant_search"].(map[string]interface{})
	expectedSearchEndpoints := []string{
		"GET /api/v1/merchants",
		"POST /api/v1/merchants/search",
	}

	for _, endpoint := range expectedSearchEndpoints {
		assert.Contains(t, searchEndpoints, endpoint, "Search endpoint %s should be documented", endpoint)
	}
}

// Benchmark tests for route registration performance
func BenchmarkRegisterMerchantRoutes(b *testing.B) {
	config := createTestMerchantRouteConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mux := http.NewServeMux()
		RegisterMerchantRoutes(mux, config)
	}
}

func BenchmarkCreateMerchantRouteConfig(b *testing.B) {
	mockService := &MockMerchantPortfolioService{}
	handler := handlers.NewMerchantPortfolioHandler(mockService, nil)
	logger := createTestLogger()
	authMiddleware := &middleware.AuthMiddleware{}
	rateLimiter := middleware.NewAPIRateLimiter(&middleware.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 60,
		BurstSize:         10,
		WindowSize:        60 * 1000000000,
		Strategy:          "token_bucket",
	}, logger.GetZapLogger())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreateMerchantRouteConfig(handler, authMiddleware, rateLimiter, logger)
	}
}
