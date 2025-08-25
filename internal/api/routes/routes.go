package routes

import (
	"net/http"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/api/middleware"
	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/internal/routing"
)

// RouteConfig holds configuration for route registration
type RouteConfig struct {
	IntelligentRoutingHandler   *handlers.IntelligentRoutingHandler
	AuthMiddleware              *middleware.AuthMiddleware
	RateLimiter                 *middleware.RateLimiter
	Logger                      *observability.Logger
	EnableEnhancedFeatures      bool
	EnableBackwardCompatibility bool
}

// RegisterRoutes registers all API routes with the given mux
func RegisterRoutes(mux *http.ServeMux, config *RouteConfig) {
	// Register intelligent routing endpoints
	registerIntelligentRoutingRoutes(mux, config)

	// Register enhanced business intelligence endpoints
	if config.EnableEnhancedFeatures {
		registerEnhancedBusinessIntelligenceRoutes(mux, config)
	}

	// Register backward compatibility endpoints
	if config.EnableBackwardCompatibility {
		registerBackwardCompatibilityRoutes(mux, config)
	}
}

// registerIntelligentRoutingRoutes registers the intelligent routing system endpoints
func registerIntelligentRoutingRoutes(mux *http.ServeMux, config *RouteConfig) {
	// Enhanced classification endpoints using intelligent routing
	mux.HandleFunc("POST /v2/classify", config.IntelligentRoutingHandler.ClassifyBusiness)
	mux.HandleFunc("POST /v2/classify/batch", config.IntelligentRoutingHandler.ClassifyBusinessBatch)

	// Health and metrics endpoints for intelligent routing system
	mux.HandleFunc("GET /v2/routing/health", config.IntelligentRoutingHandler.GetRoutingHealth)
	mux.HandleFunc("GET /v2/routing/metrics", config.IntelligentRoutingHandler.GetRoutingMetrics)

	config.Logger.Info("Intelligent routing routes registered", map[string]interface{}{
		"version": "v2",
		"endpoints": []string{
			"POST /v2/classify",
			"POST /v2/classify/batch",
			"GET /v2/routing/health",
			"GET /v2/routing/metrics",
		},
	})
}

// registerEnhancedBusinessIntelligenceRoutes registers enhanced business intelligence endpoints
func registerEnhancedBusinessIntelligenceRoutes(mux *http.ServeMux, config *RouteConfig) {
	// Enhanced business intelligence endpoints
	mux.HandleFunc("POST /v2/business-intelligence/enhanced-classify", config.IntelligentRoutingHandler.ClassifyBusiness)
	mux.HandleFunc("POST /v2/business-intelligence/batch-enhanced", config.IntelligentRoutingHandler.ClassifyBusinessBatch)

	// Business intelligence analytics endpoints
	mux.HandleFunc("GET /v2/business-intelligence/analytics", func(w http.ResponseWriter, r *http.Request) {
		// Enhanced analytics endpoint
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"success": true,
			"analytics": {
				"total_classifications": 0,
				"average_confidence": 0.0,
				"processing_time_avg": "0ms",
				"success_rate": 0.0
			},
			"message": "Enhanced business intelligence analytics endpoint"
		}`))
	})

	// Business intelligence insights endpoint
	mux.HandleFunc("GET /v2/business-intelligence/insights", func(w http.ResponseWriter, r *http.Request) {
		// Business insights endpoint
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"success": true,
			"insights": {
				"trends": [],
				"patterns": [],
				"recommendations": []
			},
			"message": "Enhanced business intelligence insights endpoint"
		}`))
	})

	config.Logger.Info("Enhanced business intelligence routes registered", map[string]interface{}{
		"version": "v2",
		"endpoints": []string{
			"POST /v2/business-intelligence/enhanced-classify",
			"POST /v2/business-intelligence/batch-enhanced",
			"GET /v2/business-intelligence/analytics",
			"GET /v2/business-intelligence/insights",
		},
	})
}

// registerBackwardCompatibilityRoutes registers backward compatibility endpoints
func registerBackwardCompatibilityRoutes(mux *http.ServeMux, config *RouteConfig) {
	// Legacy v1 endpoints that route through intelligent routing system
	mux.HandleFunc("POST /v1/classify", config.IntelligentRoutingHandler.ClassifyBusiness)
	mux.HandleFunc("POST /v1/classify/batch", config.IntelligentRoutingHandler.ClassifyBusinessBatch)

	// Legacy health endpoint
	mux.HandleFunc("GET /v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "healthy",
			"version": "v1",
			"message": "Legacy health endpoint - consider upgrading to v2"
		}`))
	})

	config.Logger.Info("Backward compatibility routes registered", map[string]interface{}{
		"version": "v1",
		"endpoints": []string{
			"POST /v1/classify",
			"POST /v1/classify/batch",
			"GET /v1/health",
		},
		"note": "These endpoints route through intelligent routing system for enhanced functionality",
	})
}

// CreateIntelligentRoutingHandler creates and configures the intelligent routing handler
func CreateIntelligentRoutingHandler(
	router *routing.IntelligentRouter,
	logger *observability.Logger,
	metrics *observability.Metrics,
	tracer interface{}, // Using interface{} to avoid import issues
) *handlers.IntelligentRoutingHandler {
	return handlers.NewIntelligentRoutingHandler(router, logger, metrics, tracer)
}

// RouteDocumentation provides documentation for all registered routes
func RouteDocumentation() map[string]interface{} {
	return map[string]interface{}{
		"version":     "2.0.0",
		"description": "Enhanced Business Intelligence System API Routes",
		"endpoints": map[string]interface{}{
			"v2": map[string]interface{}{
				"classification": map[string]interface{}{
					"POST /v2/classify": map[string]interface{}{
						"description": "Enhanced single business classification using intelligent routing",
						"features": []string{
							"Intelligent routing to best classification method",
							"Enhanced confidence scoring",
							"Multiple industry code mappings",
							"Real-time processing",
						},
					},
					"POST /v2/classify/batch": map[string]interface{}{
						"description": "Enhanced batch business classification using intelligent routing",
						"features": []string{
							"Batch processing up to 100 businesses",
							"Parallel processing for performance",
							"Partial success handling",
							"Comprehensive error reporting",
						},
					},
				},
				"routing": map[string]interface{}{
					"GET /v2/routing/health": map[string]interface{}{
						"description": "Health check for intelligent routing system",
					},
					"GET /v2/routing/metrics": map[string]interface{}{
						"description": "Performance metrics for intelligent routing system",
					},
				},
				"business_intelligence": map[string]interface{}{
					"POST /v2/business-intelligence/enhanced-classify": map[string]interface{}{
						"description": "Enhanced classification with business intelligence features",
					},
					"POST /v2/business-intelligence/batch-enhanced": map[string]interface{}{
						"description": "Enhanced batch classification with business intelligence features",
					},
					"GET /v2/business-intelligence/analytics": map[string]interface{}{
						"description": "Business intelligence analytics and insights",
					},
					"GET /v2/business-intelligence/insights": map[string]interface{}{
						"description": "Business intelligence trends and patterns",
					},
				},
			},
			"v1": map[string]interface{}{
				"classification": map[string]interface{}{
					"POST /v1/classify": map[string]interface{}{
						"description": "Legacy single business classification (routes through intelligent routing)",
						"deprecation": "Consider upgrading to v2 endpoints",
					},
					"POST /v1/classify/batch": map[string]interface{}{
						"description": "Legacy batch business classification (routes through intelligent routing)",
						"deprecation": "Consider upgrading to v2 endpoints",
					},
				},
				"health": map[string]interface{}{
					"GET /v1/health": map[string]interface{}{
						"description": "Legacy health check endpoint",
						"deprecation": "Consider upgrading to v2 endpoints",
					},
				},
			},
		},
		"migration_guide": map[string]interface{}{
			"from_v1_to_v2": map[string]interface{}{
				"endpoints": map[string]string{
					"POST /v1/classify":       "POST /v2/classify",
					"POST /v1/classify/batch": "POST /v2/classify/batch",
					"GET /v1/health":          "GET /v2/routing/health",
				},
				"benefits": []string{
					"Enhanced intelligent routing",
					"Improved performance",
					"Better error handling",
					"More comprehensive responses",
					"Advanced business intelligence features",
				},
			},
		},
	}
}
