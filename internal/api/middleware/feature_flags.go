package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
)

// FeatureFlagMiddleware adds feature flag context to requests
type FeatureFlagMiddleware struct {
	featureFlagManager *config.FeatureFlagManager
}

// NewFeatureFlagMiddleware creates a new feature flag middleware
func NewFeatureFlagMiddleware(featureFlagManager *config.FeatureFlagManager) *FeatureFlagMiddleware {
	return &FeatureFlagMiddleware{
		featureFlagManager: featureFlagManager,
	}
}

// Middleware adds feature flag context to the request
func (ffm *FeatureFlagMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate or extract request ID
		requestID := ffm.getOrGenerateRequestID(r)

		// Add request ID to headers for client reference
		w.Header().Set("X-Request-ID", requestID)

		// Create feature flag context
		ctx := ffm.featureFlagManager.FeatureFlagContext(r.Context(), requestID)

		// Add feature flag information to response headers for debugging
		ffm.addFeatureFlagHeaders(w, ctx)

		// Create new request with updated context
		r = r.WithContext(ctx)

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

// getOrGenerateRequestID extracts request ID from headers or generates a new one
func (ffm *FeatureFlagMiddleware) getOrGenerateRequestID(r *http.Request) string {
	// Check for existing request ID in headers
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		return requestID
	}

	// Check for correlation ID
	if correlationID := r.Header.Get("X-Correlation-ID"); correlationID != "" {
		return correlationID
	}

	// Generate new request ID
	return ffm.generateRequestID()
}

// generateRequestID generates a unique request ID
func (ffm *FeatureFlagMiddleware) generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// addFeatureFlagHeaders adds feature flag information to response headers
func (ffm *FeatureFlagMiddleware) addFeatureFlagHeaders(w http.ResponseWriter, ctx context.Context) {
	flags := config.GetFeatureFlagsFromContext(ctx)

	// Add enabled flags to headers
	enabledFlags := make([]string, 0)
	for flagName, enabled := range flags {
		if enabled {
			enabledFlags = append(enabledFlags, flagName)
		}
	}

	if len(enabledFlags) > 0 {
		w.Header().Set("X-Feature-Flags", joinStrings(enabledFlags, ","))
	}

	// Add architecture information
	if flags["modular_architecture"] {
		w.Header().Set("X-Architecture", "modular")
	} else {
		w.Header().Set("X-Architecture", "legacy")
	}

	// Add A/B testing information
	if flags["a_b_testing"] {
		w.Header().Set("X-AB-Testing", "enabled")
	}
}

// joinStrings joins a slice of strings with a separator
func joinStrings(elems []string, sep string) string {
	if len(elems) == 0 {
		return ""
	}

	result := elems[0]
	for i := 1; i < len(elems); i++ {
		result += sep + elems[i]
	}
	return result
}

// FeatureFlagAwareHandler is an interface for handlers that need feature flag awareness
type FeatureFlagAwareHandler interface {
	HandleWithFeatureFlags(w http.ResponseWriter, r *http.Request, flags map[string]bool)
}

// FeatureFlagHandler wraps a handler with feature flag awareness
type FeatureFlagHandler struct {
	handler            FeatureFlagAwareHandler
	featureFlagManager *config.FeatureFlagManager
}

// NewFeatureFlagHandler creates a new feature flag aware handler
func NewFeatureFlagHandler(handler FeatureFlagAwareHandler, featureFlagManager *config.FeatureFlagManager) *FeatureFlagHandler {
	return &FeatureFlagHandler{
		handler:            handler,
		featureFlagManager: featureFlagManager,
	}
}

// ServeHTTP handles requests with feature flag context
func (ffh *FeatureFlagHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract feature flags from context
	flags := config.GetFeatureFlagsFromContext(r.Context())

	// Call the feature flag aware handler
	ffh.handler.HandleWithFeatureFlags(w, r, flags)
}

// ClassificationHandlerWithFeatureFlags is a feature flag aware classification handler
type ClassificationHandlerWithFeatureFlags struct {
	legacyHandler      http.Handler
	modularHandler     http.Handler
	featureFlagManager *config.FeatureFlagManager
}

// NewClassificationHandlerWithFeatureFlags creates a new classification handler with feature flags
func NewClassificationHandlerWithFeatureFlags(
	legacyHandler http.Handler,
	modularHandler http.Handler,
	featureFlagManager *config.FeatureFlagManager,
) *ClassificationHandlerWithFeatureFlags {
	return &ClassificationHandlerWithFeatureFlags{
		legacyHandler:      legacyHandler,
		modularHandler:     modularHandler,
		featureFlagManager: featureFlagManager,
	}
}

// ServeHTTP routes requests based on feature flags
func (ch *ClassificationHandlerWithFeatureFlags) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := config.GetRequestIDFromContext(r.Context())

	// Check if we should use modular architecture
	if ch.featureFlagManager.ShouldUseModularArchitecture(r.Context(), requestID) {
		// Use modular architecture
		ch.modularHandler.ServeHTTP(w, r)
		return
	}

	// Check if we should use legacy implementation
	if ch.featureFlagManager.ShouldUseLegacyImplementation(r.Context(), requestID) {
		// Use legacy implementation
		ch.legacyHandler.ServeHTTP(w, r)
		return
	}

	// Default to legacy implementation
	ch.legacyHandler.ServeHTTP(w, r)
}

// ABTestingMiddleware adds A/B testing capabilities
type ABTestingMiddleware struct {
	featureFlagManager *config.FeatureFlagManager
}

// NewABTestingMiddleware creates a new A/B testing middleware
func NewABTestingMiddleware(featureFlagManager *config.FeatureFlagManager) *ABTestingMiddleware {
	return &ABTestingMiddleware{
		featureFlagManager: featureFlagManager,
	}
}

// Middleware adds A/B testing context to requests
func (abm *ABTestingMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := config.GetRequestIDFromContext(r.Context())

		// Check if A/B testing is enabled
		if abm.featureFlagManager.ShouldEnableABTesting(r.Context(), requestID) {
			// Add A/B testing headers
			w.Header().Set("X-AB-Test", "enabled")
			w.Header().Set("X-AB-Test-Variant", abm.determineABTestVariant(requestID))

			// Add A/B testing context
			ctx := context.WithValue(r.Context(), "ab_testing", true)
			ctx = context.WithValue(ctx, "ab_test_variant", abm.determineABTestVariant(requestID))
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

// determineABTestVariant determines the A/B test variant based on request ID
func (abm *ABTestingMiddleware) determineABTestVariant(requestID string) string {
	// Simple hash-based variant selection
	hash := 0
	for _, char := range requestID {
		hash = ((hash << 5) - hash) + int(char)
		hash = hash & hash // Convert to 32-bit integer
	}

	if hash < 0 {
		hash = -hash
	}

	// 50/50 split between A and B
	if hash%2 == 0 {
		return "A"
	}
	return "B"
}

// GracefulDegradationMiddleware adds graceful degradation capabilities
type GracefulDegradationMiddleware struct {
	featureFlagManager *config.FeatureFlagManager
}

// NewGracefulDegradationMiddleware creates a new graceful degradation middleware
func NewGracefulDegradationMiddleware(featureFlagManager *config.FeatureFlagManager) *GracefulDegradationMiddleware {
	return &GracefulDegradationMiddleware{
		featureFlagManager: featureFlagManager,
	}
}

// Middleware adds graceful degradation context to requests
func (gdm *GracefulDegradationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if graceful degradation is enabled
		if gdm.featureFlagManager.ShouldEnableGracefulDegradation(r.Context()) {
			// Add graceful degradation context
			ctx := context.WithValue(r.Context(), "graceful_degradation", true)
			r = r.WithContext(ctx)

			// Add graceful degradation headers
			w.Header().Set("X-Graceful-Degradation", "enabled")
		}

		next.ServeHTTP(w, r)
	})
}

// PerformanceMonitoringMiddleware adds performance monitoring capabilities
type PerformanceMonitoringMiddleware struct {
	featureFlagManager *config.FeatureFlagManager
}

// NewPerformanceMonitoringMiddleware creates a new performance monitoring middleware
func NewPerformanceMonitoringMiddleware(featureFlagManager *config.FeatureFlagManager) *PerformanceMonitoringMiddleware {
	return &PerformanceMonitoringMiddleware{
		featureFlagManager: featureFlagManager,
	}
}

// Middleware adds performance monitoring context to requests
func (pmm *PerformanceMonitoringMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Check if performance monitoring is enabled
		if pmm.featureFlagManager.ShouldEnablePerformanceMonitoring(r.Context()) {
			// Add performance monitoring context
			ctx := context.WithValue(r.Context(), "performance_monitoring", true)
			ctx = context.WithValue(ctx, "request_start_time", startTime)
			r = r.WithContext(ctx)

			// Add performance monitoring headers
			w.Header().Set("X-Performance-Monitoring", "enabled")
		}

		next.ServeHTTP(w, r)

		// Calculate and add response time header
		if pmm.featureFlagManager.ShouldEnablePerformanceMonitoring(r.Context()) {
			duration := time.Since(startTime)
			w.Header().Set("X-Response-Time", duration.String())
		}
	})
}

// FeatureFlagStatusHandler provides an endpoint to check feature flag status
type FeatureFlagStatusHandler struct {
	featureFlagManager *config.FeatureFlagManager
}

// NewFeatureFlagStatusHandler creates a new feature flag status handler
func NewFeatureFlagStatusHandler(featureFlagManager *config.FeatureFlagManager) *FeatureFlagStatusHandler {
	return &FeatureFlagStatusHandler{
		featureFlagManager: featureFlagManager,
	}
}

// ServeHTTP handles feature flag status requests
func (ffsh *FeatureFlagStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get rollout status
	status := ffsh.featureFlagManager.GetRolloutStatus()

	// Simple JSON response
	response := fmt.Sprintf(`{
		"status": "success",
		"data": %v,
		"timestamp": "%s"
	}`, status, time.Now().UTC().Format(time.RFC3339))

	w.Write([]byte(response))
}
