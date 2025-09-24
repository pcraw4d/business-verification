package routes

import (
	"net/http"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/api/middleware"
	"kyb-platform/internal/observability"
)

// MerchantRouteConfig holds configuration for merchant route registration
type MerchantRouteConfig struct {
	MerchantPortfolioHandler *handlers.MerchantPortfolioHandler
	AuthMiddleware           *middleware.AuthMiddleware
	RateLimiter              *middleware.APIRateLimiter
	Logger                   *observability.Logger
	EnableBulkOperations     bool
	EnableSessionManagement  bool
	MaxBulkOperationSize     int
}

// RegisterMerchantRoutes registers all merchant portfolio API routes with the given mux
func RegisterMerchantRoutes(mux *http.ServeMux, config *MerchantRouteConfig) {
	// Register merchant CRUD routes
	registerMerchantCRUDRoutes(mux, config)

	// Register merchant search and listing routes
	registerMerchantSearchRoutes(mux, config)

	// Register bulk operations routes (if enabled)
	if config.EnableBulkOperations {
		registerBulkOperationRoutes(mux, config)
	}

	// Register session management routes (if enabled)
	if config.EnableSessionManagement {
		registerSessionManagementRoutes(mux, config)
	}

	// Register merchant analytics routes
	registerMerchantAnalyticsRoutes(mux, config)

	config.Logger.Info("Merchant portfolio routes registered", map[string]interface{}{
		"version": "v1",
		"endpoints": []string{
			"POST /api/v1/merchants",
			"GET /api/v1/merchants",
			"GET /api/v1/merchants/{id}",
			"PUT /api/v1/merchants/{id}",
			"DELETE /api/v1/merchants/{id}",
			"POST /api/v1/merchants/search",
			"GET /api/v1/merchants/analytics",
			"GET /api/v1/merchants/portfolio-types",
			"GET /api/v1/merchants/risk-levels",
		},
		"bulk_operations_enabled":    config.EnableBulkOperations,
		"session_management_enabled": config.EnableSessionManagement,
	})
}

// registerMerchantCRUDRoutes registers merchant CRUD operation routes
func registerMerchantCRUDRoutes(mux *http.ServeMux, config *MerchantRouteConfig) {
	// Create merchant
	mux.Handle("POST /api/v1/merchants",
		config.AuthMiddleware.RequireAuth(
			config.RateLimiter.Middleware(
				http.HandlerFunc(config.MerchantPortfolioHandler.CreateMerchant),
			),
		),
	)

	// Get merchant by ID
	mux.Handle("GET /api/v1/merchants/{id}",
		config.AuthMiddleware.RequireAuth(
			config.RateLimiter.Middleware(
				http.HandlerFunc(config.MerchantPortfolioHandler.GetMerchant),
			),
		),
	)

	// Update merchant
	mux.Handle("PUT /api/v1/merchants/{id}",
		config.AuthMiddleware.RequireAuth(
			config.RateLimiter.Middleware(
				http.HandlerFunc(config.MerchantPortfolioHandler.UpdateMerchant),
			),
		),
	)

	// Delete merchant
	mux.Handle("DELETE /api/v1/merchants/{id}",
		config.AuthMiddleware.RequireAuth(
			config.RateLimiter.Middleware(
				http.HandlerFunc(config.MerchantPortfolioHandler.DeleteMerchant),
			),
		),
	)

	config.Logger.Info("Merchant CRUD routes registered", map[string]interface{}{
		"endpoints": []string{
			"POST /api/v1/merchants",
			"GET /api/v1/merchants/{id}",
			"PUT /api/v1/merchants/{id}",
			"DELETE /api/v1/merchants/{id}",
		},
	})
}

// registerMerchantSearchRoutes registers merchant search and listing routes
func registerMerchantSearchRoutes(mux *http.ServeMux, config *MerchantRouteConfig) {
	// List merchants with query parameters
	mux.Handle("GET /api/v1/merchants",
		config.AuthMiddleware.RequireAuth(
			config.RateLimiter.Middleware(
				http.HandlerFunc(config.MerchantPortfolioHandler.ListMerchants),
			),
		),
	)

	// Search merchants with POST body
	mux.Handle("POST /api/v1/merchants/search",
		config.AuthMiddleware.RequireAuth(
			config.RateLimiter.Middleware(
				http.HandlerFunc(config.MerchantPortfolioHandler.SearchMerchants),
			),
		),
	)

	config.Logger.Info("Merchant search routes registered", map[string]interface{}{
		"endpoints": []string{
			"GET /api/v1/merchants",
			"POST /api/v1/merchants/search",
		},
	})
}

// registerBulkOperationRoutes registers bulk operation routes with enhanced rate limiting
func registerBulkOperationRoutes(mux *http.ServeMux, config *MerchantRouteConfig) {
	// Create enhanced rate limiter for bulk operations
	bulkRateLimiter := middleware.NewAPIRateLimiter(&middleware.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 10,              // Lower rate limit for bulk operations
		BurstSize:         2,               // Smaller burst size
		WindowSize:        60 * 1000000000, // 1 minute in nanoseconds
		Strategy:          "token_bucket",
	}, config.Logger.GetZapLogger())

	// Bulk update merchants
	mux.Handle("POST /api/v1/merchants/bulk/update",
		config.AuthMiddleware.RequireAuth(
			bulkRateLimiter.Middleware(
				http.HandlerFunc(config.MerchantPortfolioHandler.BulkUpdateMerchants),
			),
		),
	)

	// Bulk export merchants
	mux.Handle("POST /api/v1/merchants/bulk/export",
		config.AuthMiddleware.RequireAuth(
			bulkRateLimiter.Middleware(
				http.HandlerFunc(config.MerchantPortfolioHandler.BulkExportMerchants),
			),
		),
	)

	config.Logger.Info("Bulk operation routes registered", map[string]interface{}{
		"endpoints": []string{
			"POST /api/v1/merchants/bulk/update",
			"POST /api/v1/merchants/bulk/export",
		},
		"max_bulk_size": config.MaxBulkOperationSize,
		"rate_limit": map[string]interface{}{
			"requests_per_minute": 10,
			"burst_size":          2,
		},
	})
}

// registerSessionManagementRoutes registers session management routes
func registerSessionManagementRoutes(mux *http.ServeMux, config *MerchantRouteConfig) {
	// Start merchant session
	mux.Handle("POST /api/v1/merchants/{id}/session",
		config.AuthMiddleware.RequireAuth(
			config.RateLimiter.Middleware(
				http.HandlerFunc(config.MerchantPortfolioHandler.StartMerchantSession),
			),
		),
	)

	// End merchant session
	mux.Handle("DELETE /api/v1/merchants/{id}/session",
		config.AuthMiddleware.RequireAuth(
			config.RateLimiter.Middleware(
				http.HandlerFunc(config.MerchantPortfolioHandler.EndMerchantSession),
			),
		),
	)

	// Get active session
	mux.Handle("GET /api/v1/merchants/session/active",
		config.AuthMiddleware.RequireAuth(
			config.RateLimiter.Middleware(
				http.HandlerFunc(config.MerchantPortfolioHandler.GetActiveSession),
			),
		),
	)

	config.Logger.Info("Session management routes registered", map[string]interface{}{
		"endpoints": []string{
			"POST /api/v1/merchants/{id}/session",
			"DELETE /api/v1/merchants/{id}/session",
			"GET /api/v1/merchants/session/active",
		},
	})
}

// registerMerchantAnalyticsRoutes registers merchant analytics and metadata routes
func registerMerchantAnalyticsRoutes(mux *http.ServeMux, config *MerchantRouteConfig) {
	// Get merchant analytics
	mux.Handle("GET /api/v1/merchants/analytics",
		config.AuthMiddleware.RequireAuth(
			config.RateLimiter.Middleware(
				http.HandlerFunc(config.MerchantPortfolioHandler.GetMerchantAnalytics),
			),
		),
	)

	// Get portfolio types
	mux.Handle("GET /api/v1/merchants/portfolio-types",
		config.AuthMiddleware.RequireAuth(
			config.RateLimiter.Middleware(
				http.HandlerFunc(config.MerchantPortfolioHandler.GetPortfolioTypes),
			),
		),
	)

	// Get risk levels
	mux.Handle("GET /api/v1/merchants/risk-levels",
		config.AuthMiddleware.RequireAuth(
			config.RateLimiter.Middleware(
				http.HandlerFunc(config.MerchantPortfolioHandler.GetRiskLevels),
			),
		),
	)

	// Get merchant statistics
	mux.Handle("GET /api/v1/merchants/statistics",
		config.AuthMiddleware.RequireAuth(
			config.RateLimiter.Middleware(
				http.HandlerFunc(config.MerchantPortfolioHandler.GetMerchantStatistics),
			),
		),
	)

	config.Logger.Info("Merchant analytics routes registered", map[string]interface{}{
		"endpoints": []string{
			"GET /api/v1/merchants/analytics",
			"GET /api/v1/merchants/portfolio-types",
			"GET /api/v1/merchants/risk-levels",
			"GET /api/v1/merchants/statistics",
		},
	})
}

// CreateMerchantRouteConfig creates and configures the merchant route configuration
func CreateMerchantRouteConfig(
	merchantHandler *handlers.MerchantPortfolioHandler,
	authMiddleware *middleware.AuthMiddleware,
	rateLimiter *middleware.APIRateLimiter,
	logger *observability.Logger,
) *MerchantRouteConfig {
	return &MerchantRouteConfig{
		MerchantPortfolioHandler: merchantHandler,
		AuthMiddleware:           authMiddleware,
		RateLimiter:              rateLimiter,
		Logger:                   logger,
		EnableBulkOperations:     true,
		EnableSessionManagement:  true,
		MaxBulkOperationSize:     1000, // Maximum merchants per bulk operation
	}
}

// MerchantRouteDocumentation provides documentation for all merchant routes
func MerchantRouteDocumentation() map[string]interface{} {
	return map[string]interface{}{
		"version":     "1.0.0",
		"description": "Merchant Portfolio Management API Routes",
		"endpoints": map[string]interface{}{
			"merchant_crud": map[string]interface{}{
				"POST /api/v1/merchants": map[string]interface{}{
					"description":    "Create a new merchant",
					"authentication": "Required",
					"rate_limit":     "Standard",
					"request_body":   "CreateMerchantRequest",
					"response":       "MerchantResponse",
				},
				"GET /api/v1/merchants/{id}": map[string]interface{}{
					"description":    "Get merchant by ID",
					"authentication": "Required",
					"rate_limit":     "Standard",
					"response":       "MerchantResponse",
				},
				"PUT /api/v1/merchants/{id}": map[string]interface{}{
					"description":    "Update merchant by ID",
					"authentication": "Required",
					"rate_limit":     "Standard",
					"request_body":   "UpdateMerchantRequest",
					"response":       "MerchantResponse",
				},
				"DELETE /api/v1/merchants/{id}": map[string]interface{}{
					"description":    "Delete merchant by ID",
					"authentication": "Required",
					"rate_limit":     "Standard",
					"response":       "204 No Content",
				},
			},
			"merchant_search": map[string]interface{}{
				"GET /api/v1/merchants": map[string]interface{}{
					"description":    "List merchants with query parameters",
					"authentication": "Required",
					"rate_limit":     "Standard",
					"query_parameters": []string{
						"query", "portfolio_type", "risk_level", "industry", "status",
						"page", "page_size", "sort_by", "sort_order",
					},
					"response": "MerchantListResponse",
				},
				"POST /api/v1/merchants/search": map[string]interface{}{
					"description":    "Search merchants with advanced filters",
					"authentication": "Required",
					"rate_limit":     "Standard",
					"request_body":   "MerchantSearchRequest",
					"response":       "MerchantListResponse",
				},
			},
			"bulk_operations": map[string]interface{}{
				"POST /api/v1/merchants/bulk/update": map[string]interface{}{
					"description":    "Bulk update merchants (portfolio type, risk level, status)",
					"authentication": "Required",
					"rate_limit":     "Enhanced (10 req/min, burst 2)",
					"request_body":   "BulkOperationRequest",
					"response":       "BulkOperationResponse",
					"max_merchants":  1000,
				},
				"POST /api/v1/merchants/bulk/export": map[string]interface{}{
					"description":    "Bulk export merchants to CSV/JSON",
					"authentication": "Required",
					"rate_limit":     "Enhanced (10 req/min, burst 2)",
					"request_body":   "BulkExportRequest",
					"response":       "BulkExportResponse",
					"max_merchants":  1000,
				},
			},
			"session_management": map[string]interface{}{
				"POST /api/v1/merchants/{id}/session": map[string]interface{}{
					"description":    "Start a merchant session (single merchant active at a time)",
					"authentication": "Required",
					"rate_limit":     "Standard",
					"response":       "SessionResponse",
				},
				"DELETE /api/v1/merchants/{id}/session": map[string]interface{}{
					"description":    "End the current merchant session",
					"authentication": "Required",
					"rate_limit":     "Standard",
					"response":       "204 No Content",
				},
				"GET /api/v1/merchants/session/active": map[string]interface{}{
					"description":    "Get the currently active merchant session",
					"authentication": "Required",
					"rate_limit":     "Standard",
					"response":       "SessionResponse",
				},
			},
			"analytics": map[string]interface{}{
				"GET /api/v1/merchants/analytics": map[string]interface{}{
					"description":    "Get merchant portfolio analytics and insights",
					"authentication": "Required",
					"rate_limit":     "Standard",
					"response":       "MerchantAnalyticsResponse",
				},
				"GET /api/v1/merchants/portfolio-types": map[string]interface{}{
					"description":    "Get available portfolio types",
					"authentication": "Required",
					"rate_limit":     "Standard",
					"response":       "PortfolioTypeListResponse",
				},
				"GET /api/v1/merchants/risk-levels": map[string]interface{}{
					"description":    "Get available risk levels",
					"authentication": "Required",
					"rate_limit":     "Standard",
					"response":       "RiskLevelListResponse",
				},
				"GET /api/v1/merchants/statistics": map[string]interface{}{
					"description":    "Get merchant portfolio statistics",
					"authentication": "Required",
					"rate_limit":     "Standard",
					"response":       "MerchantStatisticsResponse",
				},
			},
		},
		"rate_limiting": map[string]interface{}{
			"standard": map[string]interface{}{
				"requests_per_minute": 60,
				"burst_size":          10,
				"strategy":            "token_bucket",
			},
			"bulk_operations": map[string]interface{}{
				"requests_per_minute": 10,
				"burst_size":          2,
				"strategy":            "token_bucket",
			},
		},
		"authentication": map[string]interface{}{
			"type":     "JWT Bearer Token",
			"header":   "Authorization: Bearer <token>",
			"required": true,
		},
		"features": []string{
			"Merchant CRUD operations",
			"Advanced search and filtering",
			"Bulk operations with progress tracking",
			"Single merchant session management",
			"Portfolio analytics and insights",
			"Rate limiting and authentication",
			"Comprehensive audit logging",
		},
	}
}
