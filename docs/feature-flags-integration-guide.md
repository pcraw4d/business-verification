# Feature Flags Integration Guide

## Overview

This guide explains how to integrate feature flags into the KYB Platform for gradual rollout of new modules and features. The feature flag system provides safe, controlled deployment of new functionality with the ability to quickly rollback if issues arise.

## Architecture

### Feature Flag Manager

The `FeatureFlagManager` is the core component that manages all feature flags:

```go
type FeatureFlagManager struct {
    flags map[string]*FeatureFlag
    mu    sync.RWMutex
    env   string
}
```

### Feature Flag Structure

Each feature flag has the following structure:

```go
type FeatureFlag struct {
    Name        string
    Description string
    Enabled     bool
    Percentage  int // 0-100 for gradual rollout
    StartTime   time.Time
    EndTime     *time.Time // nil means no end time
    Metadata    map[string]interface{}
}
```

## Integration Steps

### Step 1: Initialize Feature Flag Manager

```go
import "github.com/pcraw4d/business-verification/internal/config"

// Initialize feature flag manager
featureFlagManager := config.NewFeatureFlagManager("production")

// The manager automatically loads default flags from environment variables
```

### Step 2: Add Feature Flag Middleware

```go
import "github.com/pcraw4d/business-verification/internal/api/middleware"

// Create feature flag middleware
featureFlagMiddleware := middleware.NewFeatureFlagMiddleware(featureFlagManager)

// Add to your server
mux := http.NewServeMux()
mux.Use(featureFlagMiddleware.Middleware)
```

### Step 3: Create Feature Flag Aware Handlers

```go
// Create classification handler with feature flags
classificationHandler := middleware.NewClassificationHandlerWithFeatureFlags(
    legacyHandler,    // Legacy classification handler
    modularHandler,   // New modular classification handler
    featureFlagManager,
)

// Register the handler
mux.HandleFunc("POST /api/v1/classify", classificationHandler.ServeHTTP)
```

### Step 4: Add Additional Middleware (Optional)

```go
// A/B Testing middleware
abTestingMiddleware := middleware.NewABTestingMiddleware(featureFlagManager)
mux.Use(abTestingMiddleware.Middleware)

// Graceful degradation middleware
gracefulDegradationMiddleware := middleware.NewGracefulDegradationMiddleware(featureFlagManager)
mux.Use(gracefulDegradationMiddleware.Middleware)

// Performance monitoring middleware
performanceMonitoringMiddleware := middleware.NewPerformanceMonitoringMiddleware(featureFlagManager)
mux.Use(performanceMonitoringMiddleware.Middleware)
```

## Configuration

### Environment Variables

Feature flags are configured through environment variables:

```bash
# Enable modular architecture with 25% rollout
ENABLE_MODULAR_ARCHITECTURE=true
MODULAR_ARCHITECTURE_PERCENTAGE=25

# Enable intelligent routing with 25% rollout
ENABLE_INTELLIGENT_ROUTING=true
INTELLIGENT_ROUTING_PERCENTAGE=25

# Enable A/B testing with 10% rollout
ENABLE_AB_TESTING=true
AB_TESTING_PERCENTAGE=10

# Keep legacy compatibility enabled
ENABLE_LEGACY_COMPATIBILITY=true
```

### Configuration File

Use the provided configuration file:

```bash
# Load feature flag configuration
source configs/feature-flags.env
```

## Usage Examples

### Basic Feature Flag Check

```go
func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    requestID := config.GetRequestIDFromContext(r.Context())
    
    // Check if modular architecture should be used
    if h.featureFlagManager.ShouldUseModularArchitecture(r.Context(), requestID) {
        // Use new modular implementation
        result := h.modularService.Process(r.Context(), request)
        respondWithJSON(w, result)
        return
    }
    
    // Fall back to legacy implementation
    result := h.legacyService.Process(r.Context(), request)
    respondWithJSON(w, result)
}
```

### A/B Testing Integration

```go
func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    // Check if A/B testing is enabled
    if r.Context().Value("ab_testing") == true {
        variant := r.Context().Value("ab_test_variant").(string)
        
        // Track A/B test variant
        h.metrics.TrackABTestVariant(requestID, variant)
        
        // Use different implementations based on variant
        if variant == "A" {
            result := h.implementationA.Process(r.Context(), request)
            respondWithJSON(w, result)
        } else {
            result := h.implementationB.Process(r.Context(), request)
            respondWithJSON(w, result)
        }
        return
    }
    
    // Default implementation
    result := h.defaultImplementation.Process(r.Context(), request)
    respondWithJSON(w, result)
}
```

### Graceful Degradation

```go
func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    // Check if graceful degradation is enabled
    if r.Context().Value("graceful_degradation") == true {
        // Try new implementation first
        result, err := h.newImplementation.Process(r.Context(), request)
        if err != nil {
            // Fall back to legacy implementation
            h.logger.Warn("New implementation failed, falling back to legacy", "error", err)
            result = h.legacyImplementation.Process(r.Context(), request)
        }
        respondWithJSON(w, result)
        return
    }
    
    // Use new implementation without fallback
    result := h.newImplementation.Process(r.Context(), request)
    respondWithJSON(w, result)
}
```

## Rollout Strategy

### Phase 1: Development (0% rollout)
- All new features disabled
- Legacy implementation active
- Development and testing phase

### Phase 2: Internal Testing (10% rollout)
- Enable modular architecture for 10% of requests
- Enable intelligent routing for 10% of requests
- Enable A/B testing for 10% of requests
- Monitor performance and error rates

### Phase 3: Beta Testing (25% rollout)
- Increase rollout to 25% of requests
- Enable enhanced classification features
- Gather feedback and metrics
- Validate performance improvements

### Phase 4: Production Rollout (50% rollout)
- Increase rollout to 50% of requests
- Monitor production metrics
- Compare performance between implementations
- Validate accuracy improvements

### Phase 5: Full Production (100% rollout)
- Rollout to 100% of requests
- Disable legacy compatibility (after validation)
- Monitor for any issues
- Optimize performance

## Monitoring and Metrics

### Response Headers

The feature flag middleware adds several headers to responses:

```
X-Request-ID: abc123def456
X-Feature-Flags: modular_architecture,intelligent_routing
X-Architecture: modular
X-AB-Testing: enabled
X-AB-Test-Variant: A
X-Graceful-Degradation: enabled
X-Performance-Monitoring: enabled
X-Response-Time: 150ms
```

### Metrics to Track

1. **Response Times**: Compare response times between legacy and modular implementations
2. **Error Rates**: Monitor error rates for both implementations
3. **Success Rates**: Track success rates and accuracy improvements
4. **Resource Usage**: Monitor CPU, memory, and network usage
5. **A/B Test Results**: Compare performance between variants

### Feature Flag Status Endpoint

```go
// Add feature flag status endpoint
statusHandler := middleware.NewFeatureFlagStatusHandler(featureFlagManager)
mux.HandleFunc("GET /api/v1/feature-flags/status", statusHandler.ServeHTTP)
```

Response:
```json
{
  "status": "success",
  "data": {
    "modular_architecture": {
      "enabled": true,
      "percentage": 25,
      "start_time": "2024-01-15T10:00:00Z",
      "end_time": null,
      "metadata": {
        "modules": ["keyword_classification", "ml_classification", "website_analysis", "web_search_analysis"]
      }
    }
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Rollback Strategy

### Immediate Rollback

To immediately rollback to legacy implementation:

```bash
# Set all percentages to 0
ENABLE_MODULAR_ARCHITECTURE=false
MODULAR_ARCHITECTURE_PERCENTAGE=0
ENABLE_INTELLIGENT_ROUTING=false
INTELLIGENT_ROUTING_PERCENTAGE=0
ENABLE_ENHANCED_CLASSIFICATION=false
ENHANCED_CLASSIFICATION_PERCENTAGE=0

# Ensure legacy compatibility is enabled
ENABLE_LEGACY_COMPATIBILITY=true
```

### Gradual Rollback

To gradually rollback:

```bash
# Reduce percentages step by step
MODULAR_ARCHITECTURE_PERCENTAGE=10
INTELLIGENT_ROUTING_PERCENTAGE=10
ENHANCED_CLASSIFICATION_PERCENTAGE=10
```

### Monitoring During Rollback

1. **Error Rate Monitoring**: Watch for increased error rates
2. **Performance Monitoring**: Monitor response time improvements
3. **User Feedback**: Gather feedback on any issues
4. **Metrics Comparison**: Compare metrics between implementations

## Best Practices

### 1. Always Enable Legacy Compatibility During Rollout

```bash
# Keep legacy compatibility enabled during rollout
ENABLE_LEGACY_COMPATIBILITY=true
```

### 2. Use Gradual Rollout

```bash
# Start with small percentages
MODULAR_ARCHITECTURE_PERCENTAGE=10
# Gradually increase
MODULAR_ARCHITECTURE_PERCENTAGE=25
MODULAR_ARCHITECTURE_PERCENTAGE=50
MODULAR_ARCHITECTURE_PERCENTAGE=100
```

### 3. Monitor Closely

- Set up alerts for increased error rates
- Monitor response time changes
- Track user feedback and complaints
- Compare accuracy metrics

### 4. Have Rollback Plan Ready

- Document rollback procedures
- Test rollback scenarios
- Have monitoring in place
- Keep legacy code functional

### 5. Use A/B Testing for Validation

```bash
# Enable A/B testing to compare implementations
ENABLE_AB_TESTING=true
AB_TESTING_PERCENTAGE=10
```

## Troubleshooting

### Common Issues

1. **Feature Flags Not Working**
   - Check environment variables are loaded
   - Verify feature flag manager is initialized
   - Check middleware is properly configured

2. **Inconsistent Behavior**
   - Verify request ID generation is consistent
   - Check percentage calculation logic
   - Ensure proper context propagation

3. **Performance Issues**
   - Monitor response times for both implementations
   - Check resource usage
   - Validate caching strategies

4. **Rollback Not Working**
   - Verify legacy compatibility is enabled
   - Check feature flag percentages are set to 0
   - Ensure legacy handlers are properly configured

### Debugging

1. **Check Response Headers**
   - Look for feature flag headers in responses
   - Verify architecture type is correct
   - Check A/B testing variant assignment

2. **Monitor Logs**
   - Add logging for feature flag decisions
   - Track which implementation is used
   - Monitor error rates and performance

3. **Use Status Endpoint**
   - Check feature flag status via API
   - Verify flag configurations
   - Monitor rollout percentages

## Conclusion

The feature flag system provides a safe and controlled way to roll out new functionality. By following this guide, you can:

- **Safely deploy** new features with minimal risk
- **Monitor performance** and gather metrics
- **Quickly rollback** if issues arise
- **A/B test** different implementations
- **Gradually increase** rollout percentages

The system is designed to be flexible and can be adapted to different deployment strategies and requirements.
