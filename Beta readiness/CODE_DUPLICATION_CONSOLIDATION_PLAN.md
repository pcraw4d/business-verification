# Code Duplication Consolidation Plan

**Date**: 2025-01-27  
**Status**: Analysis Complete - Post-Beta Consolidation Plan

## Executive Summary

Analysis identified ~650 lines of duplicated code across services, primarily in configuration, health checks, and middleware patterns. Consolidation is recommended post-beta to reduce maintenance overhead and ensure consistency.

---

## Duplication Analysis

### 1. Configuration Code (~300 lines)

#### Duplicated Structures

**ServerConfig** (15 lines × 4 services = 60 lines)
- Location: All service `internal/config/config.go` files
- Fields: Port, Host, ReadTimeout, WriteTimeout, IdleTimeout

**SupabaseConfig** (10 lines × 4 services = 40 lines)
- Location: All service `internal/config/config.go` files
- Fields: URL, APIKey, ServiceRoleKey, JWTSecret

**Helper Functions** (~50 lines × 4 services = 200 lines)
- `getEnvAsString()`
- `getEnvAsInt()`
- `getEnvAsBool()`
- `getEnvAsDuration()`
- `getEnvAsStringSlice()`

#### Consolidation Plan

**Create**: `pkg/config/common.go`

```go
package config

import (
    "os"
    "strconv"
    "strings"
    "time"
)

// ServerConfig holds common server configuration
type ServerConfig struct {
    Port         string
    Host         string
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    IdleTimeout  time.Duration
}

// SupabaseConfig holds common Supabase configuration
type SupabaseConfig struct {
    URL            string
    APIKey         string
    ServiceRoleKey string
    JWTSecret      string
}

// Common helper functions
func GetEnvAsString(key, defaultValue string) string { ... }
func GetEnvAsInt(key string, defaultValue int) int { ... }
func GetEnvAsBool(key string, defaultValue bool) bool { ... }
func GetEnvAsDuration(key string, defaultValue time.Duration) time.Duration { ... }
func GetEnvAsStringSlice(key string, defaultValue []string) []string { ... }
```

**Estimated Reduction**: ~200 lines (67% reduction)

---

### 2. Health Check Patterns (~90 lines)

#### Duplication Found

**Health Check Handlers** in:
- `services/api-gateway/cmd/main.go`
- `services/classification-service/cmd/main.go`
- `services/merchant-service/internal/handlers/merchant.go`
- `services/risk-assessment-service/cmd/main.go`

**Common Pattern:**
```go
func HandleHealth(w http.ResponseWriter, r *http.Request) {
    response := map[string]interface{}{
        "status": "healthy",
        "service": "service-name",
        "timestamp": time.Now().UTC().Format(time.RFC3339),
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

#### Consolidation Plan

**Create**: `pkg/health/checker.go`

```go
package health

import (
    "net/http"
    "time"
    "encoding/json"
)

type HealthChecker struct {
    ServiceName string
    Checks      []HealthCheck
}

type HealthCheck func() (bool, string)

func (hc *HealthChecker) HandleHealth(w http.ResponseWriter, r *http.Request) {
    status := "healthy"
    details := make(map[string]interface{})
    
    for _, check := range hc.Checks {
        healthy, message := check()
        if !healthy {
            status = "unhealthy"
        }
        details[check.Name()] = map[string]interface{}{
            "healthy": healthy,
            "message": message,
        }
    }
    
    response := map[string]interface{}{
        "status":    status,
        "service":   hc.ServiceName,
        "timestamp": time.Now().UTC().Format(time.RFC3339),
        "checks":    details,
    }
    
    w.Header().Set("Content-Type", "application/json")
    statusCode := http.StatusOK
    if status == "unhealthy" {
        statusCode = http.StatusServiceUnavailable
    }
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}
```

**Estimated Reduction**: ~60 lines (67% reduction)

---

### 3. Security Headers Middleware (~80 lines)

#### Duplication Found

**Security Headers** implemented in:
- `services/api-gateway/internal/middleware/security_headers.go`
- `services/classification-service/cmd/main.go` (inline function)
- `services/merchant-service/cmd/main.go` (inline function)
- `services/risk-assessment-service/cmd/main.go` (likely)

**Common Pattern:**
```go
func securityHeadersMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Set security headers
            w.Header().Set("X-Frame-Options", "DENY")
            w.Header().Set("X-Content-Type-Options", "nosniff")
            // ... more headers
            next.ServeHTTP(w, r)
        })
    }
}
```

#### Consolidation Plan

**Already Partially Consolidated**: API Gateway has `middleware.SecurityHeaders`

**Action**: 
1. Extract to `pkg/middleware/security_headers.go`
2. Import in all services
3. Remove inline implementations

**Estimated Reduction**: ~60 lines (75% reduction)

---

### 4. Handler Initialization Patterns (~100 lines)

#### Duplication Found

**Common Patterns in `cmd/main.go`:**
- Logger initialization
- Configuration loading
- Supabase client initialization
- Router setup
- Middleware application
- Server startup

**Example Pattern:**
```go
// Initialize logger
logger, err := zap.NewProduction()
if err != nil {
    log.Fatalf("Failed to initialize logger: %v", err)
}
defer logger.Sync()

// Load configuration
cfg, err := config.Load()
if err != nil {
    logger.Fatal("Failed to load configuration", zap.Error(err))
}

// Initialize Supabase client
supabaseClient, err := supabase.NewClient(&cfg.Supabase, logger)
if err != nil {
    logger.Fatal("Failed to initialize Supabase client", zap.Error(err))
}
```

#### Consolidation Plan

**Create**: `pkg/server/bootstrap.go`

```go
package server

import (
    "context"
    "net/http"
    "go.uber.org/zap"
)

type BootstrapConfig struct {
    ServiceName string
    ConfigLoader func() (interface{}, error)
    SupabaseConfig *config.SupabaseConfig
    Handlers []Handler
    Middleware []func(http.Handler) http.Handler
}

func Bootstrap(cfg BootstrapConfig) (*http.Server, error) {
    // Initialize logger
    logger, err := zap.NewProduction()
    if err != nil {
        return nil, err
    }
    
    // Load configuration
    serviceConfig, err := cfg.ConfigLoader()
    if err != nil {
        return nil, err
    }
    
    // Initialize Supabase client
    supabaseClient, err := supabase.NewClient(cfg.SupabaseConfig, logger)
    if err != nil {
        return nil, err
    }
    
    // Setup router and handlers
    // ...
    
    return server, nil
}
```

**Estimated Reduction**: ~80 lines (80% reduction)

---

### 5. Error Handling Patterns (~80 lines)

#### Status: ✅ Already Consolidated

**Current State**: 
- ✅ Standardized error responses using `pkg/errors/response.go`
- ✅ All services use consistent error helpers

**No Action Needed**: Error handling is already consolidated.

---

## Total Duplication Summary

| Category | Duplicated Lines | Reduction Potential | Priority |
|----------|-----------------|---------------------|----------|
| Configuration | ~300 | ~200 (67%) | HIGH |
| Health Checks | ~90 | ~60 (67%) | MEDIUM |
| Security Headers | ~80 | ~60 (75%) | MEDIUM |
| Handler Init | ~100 | ~80 (80%) | MEDIUM |
| Error Handling | ~80 | ✅ Done | - |
| **Total** | **~650** | **~400 (62%)** | - |

---

## Consolidation Roadmap

### Phase 1: Post-Beta (High Priority)
1. ✅ **Error Handling** - Already consolidated
2. ⏳ **Security Headers** - Extract to shared package
3. ⏳ **Configuration** - Create `pkg/config/common.go`

**Estimated Effort**: 2-3 days  
**Risk**: Low (isolated changes)

### Phase 2: Post-Beta (Medium Priority)
4. ⏳ **Health Checks** - Create `pkg/health/checker.go`
5. ⏳ **Handler Initialization** - Create `pkg/server/bootstrap.go`

**Estimated Effort**: 3-4 days  
**Risk**: Medium (touches multiple services)

### Phase 3: Future (Low Priority)
6. ⏳ **Additional Patterns** - As identified
7. ⏳ **Documentation** - Update service setup guides

**Estimated Effort**: 1-2 days  
**Risk**: Low

---

## Implementation Guidelines

### 1. Shared Package Structure

```
pkg/
├── config/
│   ├── common.go          # Common config structures
│   └── helpers.go          # Helper functions
├── health/
│   └── checker.go         # Health check utilities
├── middleware/
│   └── security_headers.go # Security headers (move from API Gateway)
└── server/
    └── bootstrap.go       # Server initialization helpers
```

### 2. Migration Strategy

**Step 1**: Create shared packages
**Step 2**: Update one service as proof of concept
**Step 3**: Update remaining services one at a time
**Step 4**: Remove old duplicated code
**Step 5**: Update tests

### 3. Backward Compatibility

- Maintain existing interfaces
- Use feature flags if needed
- Gradual migration per service
- Comprehensive testing

---

## Benefits of Consolidation

### Maintenance
- ✅ Single source of truth for common patterns
- ✅ Bug fixes apply to all services automatically
- ✅ Consistent behavior across services

### Development
- ✅ Faster service creation (use shared patterns)
- ✅ Reduced code review overhead
- ✅ Easier onboarding for new developers

### Quality
- ✅ Consistent error handling
- ✅ Standardized health checks
- ✅ Unified security headers

---

## Risks and Mitigation

### Risk 1: Breaking Changes
**Mitigation**: 
- Maintain backward compatibility
- Gradual migration
- Comprehensive testing

### Risk 2: Over-Abstraction
**Mitigation**:
- Keep abstractions simple
- Allow service-specific extensions
- Document clearly

### Risk 3: Increased Coupling
**Mitigation**:
- Use interfaces where appropriate
- Keep shared packages minimal
- Service-specific code remains isolated

---

## Conclusion

**Status**: ✅ **Analysis Complete**

**Recommendation**: 
- ✅ **Pre-Beta**: No action needed (low risk, not blocking)
- ⏳ **Post-Beta**: Consolidate configuration and security headers (Phase 1)
- ⏳ **Future**: Complete remaining consolidation (Phase 2-3)

**Estimated Total Reduction**: ~400 lines (62% reduction in duplication)

**Priority**: Medium (improves maintainability but not critical for beta)

---

## Action Items

### Pre-Beta
- [x] Complete duplication analysis
- [x] Create consolidation plan
- [ ] Document shared package structure

### Post-Beta (Phase 1)
- [ ] Create `pkg/config/common.go`
- [ ] Extract security headers to `pkg/middleware/security_headers.go`
- [ ] Update one service as proof of concept
- [ ] Test and validate approach

### Post-Beta (Phase 2)
- [ ] Create `pkg/health/checker.go`
- [ ] Create `pkg/server/bootstrap.go`
- [ ] Migrate remaining services
- [ ] Remove duplicated code

