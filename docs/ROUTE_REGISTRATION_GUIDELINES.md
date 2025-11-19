# Route Registration Guidelines

**Last Updated**: 2025-11-18  
**Status**: Active Guidelines

## Overview

This document provides guidelines for registering routes in the KYB Platform services, with emphasis on route order, PathPrefix usage, and path transformation patterns.

---

## Critical Rules

### 1. Route Registration Order

**Rule**: Specific routes MUST be registered before PathPrefix catch-all routes.

**Why**: PathPrefix routes will match and shadow more specific routes if registered first.

**Example - Correct Order**:
```go
// ✅ CORRECT: Specific routes first
api.HandleFunc("/merchants/{id}/analytics", handler).Methods("GET")
api.HandleFunc("/merchants/{id}/website-analysis", handler).Methods("GET")
api.HandleFunc("/merchants/{id}/risk-score", handler).Methods("GET")
api.HandleFunc("/merchants/search", handler).Methods("POST")
api.HandleFunc("/merchants/analytics", handler).Methods("GET")
api.HandleFunc("/merchants/{id}", handler).Methods("GET", "PUT", "DELETE")
api.HandleFunc("/merchants", handler).Methods("GET", "POST")

// PathPrefix catch-all last
api.PathPrefix("/merchants").HandlerFunc(handler)
```

**Example - Incorrect Order**:
```go
// ❌ WRONG: PathPrefix will shadow specific routes
api.PathPrefix("/merchants").HandlerFunc(handler)  // This matches everything!
api.HandleFunc("/merchants/{id}/analytics", handler).Methods("GET")  // Never reached
```

---

## Route Registration Patterns

### 2. PathPrefix Usage

**When to Use PathPrefix**:
- For proxy routes that forward all sub-routes to a backend service
- For catch-all routes that handle multiple similar endpoints
- When you want to forward `/api/v1/service/*` to a backend service

**When NOT to Use PathPrefix**:
- For specific, well-defined routes
- When you need different handlers for different sub-routes
- When path transformation is required (use specific routes instead)

**Example - Proxy Pattern**:
```go
// Specific routes with path transformation
api.HandleFunc("/risk/assess", handler.ProxyToRiskAssessment).Methods("POST")
api.HandleFunc("/risk/benchmarks", handler.ProxyToRiskAssessment).Methods("GET")

// PathPrefix for remaining routes (no transformation needed)
api.PathPrefix("/risk").HandlerFunc(handler.ProxyToRiskAssessment)
```

---

## Path Transformation Patterns

### 3. Proxy Path Transformations

When proxying requests, you may need to transform paths:

**Pattern 1: Remove Prefix**
```go
// Gateway: /api/v1/risk/assess → Service: /api/v1/assess
if path == "/api/v1/risk/assess" {
    path = "/api/v1/assess"
}
```

**Pattern 2: Map to Different Endpoint**
```go
// Gateway: /api/v1/risk/indicators/{id} → Service: /api/v1/risk/predictions/{id}
if strings.HasPrefix(path, "/api/v1/risk/indicators/") {
    parts := strings.Split(path, "/")
    if len(parts) >= 6 {
        merchantID := parts[5]  // Extract ID
        if isValidUUID(merchantID) {
            path = fmt.Sprintf("/api/v1/risk/predictions/%s", merchantID)
        } else {
            // Return error for invalid UUID
            http.Error(w, "Invalid merchant ID format", http.StatusBadRequest)
            return
        }
    }
}
```

**Pattern 3: Add Prefix**
```go
// Gateway: /api/v1/sessions/* → Service: /v1/sessions/*
path := strings.TrimPrefix(r.URL.Path, "/api/v1/sessions")
if path == "" {
    path = "/v1/sessions"
} else {
    path = "/v1/sessions" + path
}
```

---

## Route Registration Checklist

### Before Adding a New Route

- [ ] Is this a specific route or a catch-all?
- [ ] If specific, register it BEFORE any PathPrefix routes
- [ ] Does the route need path transformation?
- [ ] If yes, implement transformation logic in handler
- [ ] Are HTTP methods correctly specified?
- [ ] Is CORS handled (OPTIONS method added)?
- [ ] Is authentication required?
- [ ] Add route to public endpoint list if no auth needed

### Route Registration Template

```go
// 1. Health and metrics (always first)
router.HandleFunc("/health", handler.HealthCheck).Methods("GET")
router.Handle("/metrics", promhttp.Handler()).Methods("GET")

// 2. Root level routes
router.HandleFunc("/", handler.Root).Methods("GET")

// 3. API subrouter
api := router.PathPrefix("/api/v1").Subrouter()

// 4. Specific routes (most specific first)
api.HandleFunc("/resource/{id}/sub-resource", handler).Methods("GET", "POST")
api.HandleFunc("/resource/{id}", handler).Methods("GET", "PUT", "DELETE")
api.HandleFunc("/resource/search", handler).Methods("POST")
api.HandleFunc("/resource", handler).Methods("GET", "POST")

// 5. PathPrefix catch-all (last)
api.PathPrefix("/resource").HandlerFunc(handler)
```

---

## Common Patterns by Service

### API Gateway Service

**Pattern**: Proxy routes with path transformation

```go
// Specific routes with transformation
api.HandleFunc("/risk/assess", handler.ProxyToRiskAssessment).Methods("POST")
api.HandleFunc("/risk/indicators/{id}", handler.ProxyToRiskAssessment).Methods("GET")

// PathPrefix for remaining routes
api.PathPrefix("/risk").HandlerFunc(handler.ProxyToRiskAssessment)
```

### Merchant Service

**Pattern**: Specific routes before base routes

```go
// Sub-routes first
api.HandleFunc("/merchants/{id}/analytics", handler).Methods("GET")
api.HandleFunc("/merchants/{id}/website-analysis", handler).Methods("GET")

// Base routes
api.HandleFunc("/merchants/{id}", handler).Methods("GET", "PUT", "DELETE")
api.HandleFunc("/merchants", handler).Methods("GET", "POST")
```

### Risk Assessment Service

**Pattern**: Extensive route registration with conditional handlers

```go
// Core routes
api.HandleFunc("/assess", handler).Methods("POST")
api.HandleFunc("/risk/benchmarks", handler).Methods("GET")

// Conditional routes (only if handler available)
if dashboardHandler != nil {
    api.HandleFunc("/reporting/dashboards", dashboardHandler).Methods("POST", "GET")
}
```

---

## Testing Route Registration

### Test Route Precedence

```go
func TestRoutePrecedence(t *testing.T) {
    router := mux.NewRouter()
    
    // Register routes in order
    router.HandleFunc("/api/v1/merchants/{id}/analytics", handler1)
    router.HandleFunc("/api/v1/merchants/{id}", handler2)
    router.PathPrefix("/api/v1/merchants").HandlerFunc(handler3)
    
    // Test that specific route matches first
    req := httptest.NewRequest("GET", "/api/v1/merchants/123/analytics", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // Verify handler1 was called, not handler3
    assert.Equal(t, handler1Called, true)
    assert.Equal(t, handler3Called, false)
}
```

### Test Path Transformation

```go
func TestPathTransformation(t *testing.T) {
    handler := &GatewayHandler{}
    
    // Test risk indicators transformation
    req := httptest.NewRequest("GET", "/api/v1/risk/indicators/550e8400-e29b-41d4-a716-446655440000", nil)
    w := httptest.NewRecorder()
    
    handler.ProxyToRiskAssessment(w, req)
    
    // Verify path was transformed correctly
    assert.Contains(t, transformedPath, "/api/v1/risk/predictions/")
}
```

---

## Troubleshooting

### Issue: Route Not Matching

**Symptoms**: Route returns 404 even though it's registered

**Solutions**:
1. Check route registration order (specific before PathPrefix)
2. Verify HTTP methods match (GET, POST, etc.)
3. Check path prefix matches exactly
4. Verify route is registered on correct router/subrouter

### Issue: PathPrefix Shadowing Specific Routes

**Symptoms**: Specific route handler never called, PathPrefix handler called instead

**Solutions**:
1. Move specific route registration BEFORE PathPrefix
2. Verify PathPrefix pattern doesn't match specific route
3. Use more specific PathPrefix pattern if needed

### Issue: Path Transformation Not Working

**Symptoms**: Backend service receives wrong path

**Solutions**:
1. Add logging to see actual path being sent
2. Verify path transformation logic handles all cases
3. Check for edge cases (empty paths, query parameters)
4. Validate extracted IDs (UUID format, etc.)

---

## Best Practices

1. **Document Route Order**: Add comments explaining why routes are in a specific order
2. **Use Constants**: Define route paths as constants for consistency
3. **Validate Inputs**: Always validate path parameters (UUIDs, IDs) before transformation
4. **Log Transformations**: Log path transformations for debugging
5. **Test Edge Cases**: Test with invalid paths, missing parameters, etc.
6. **Keep Handlers Focused**: One handler per route pattern when possible

---

## Route Registration Examples

### Example 1: Simple CRUD Routes

```go
api.HandleFunc("/users/{id}", getUserHandler).Methods("GET")
api.HandleFunc("/users/{id}", updateUserHandler).Methods("PUT")
api.HandleFunc("/users/{id}", deleteUserHandler).Methods("DELETE")
api.HandleFunc("/users", listUsersHandler).Methods("GET")
api.HandleFunc("/users", createUserHandler).Methods("POST")
```

### Example 2: Routes with Sub-resources

```go
// Sub-resources first (more specific)
api.HandleFunc("/users/{id}/posts/{post_id}", getPostHandler).Methods("GET")
api.HandleFunc("/users/{id}/posts", listPostsHandler).Methods("GET")
api.HandleFunc("/users/{id}/posts", createPostHandler).Methods("POST")

// Base resource routes
api.HandleFunc("/users/{id}", getUserHandler).Methods("GET")
api.HandleFunc("/users", listUsersHandler).Methods("GET")
```

### Example 3: Proxy Routes with Transformation

```go
// Routes requiring transformation
api.HandleFunc("/api/v1/service/endpoint", func(w http.ResponseWriter, r *http.Request) {
    // Transform path
    path := strings.TrimPrefix(r.URL.Path, "/api/v1/service")
    proxyRequest(w, r, backendURL, path)
}).Methods("POST", "GET")

// Catch-all for remaining routes
api.PathPrefix("/api/v1/service").HandlerFunc(proxyHandler)
```

---

## Related Documentation

- [API Routes Comprehensive Analysis Report](../API_ROUTES_COMPREHENSIVE_ANALYSIS_REPORT.md)
- [Route Testing Checklist](./ROUTE_TESTING_CHECKLIST.md)
- [API Gateway Proxy Configuration](./API_GATEWAY_PROXY_CONFIG.md)

---

**Document Version**: 1.0.0  
**Last Updated**: 2025-11-18  
**Next Review**: 2025-12-18

