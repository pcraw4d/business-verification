# Code Duplication Analysis

**Date**: 2025-11-10  
**Status**: In Progress

---

## Configuration Code Duplication

### Identified Duplication

**All services have nearly identical configuration patterns:**

1. **ServerConfig Structure** - Duplicated across all services
   - `services/api-gateway/internal/config/config.go`
   - `services/classification-service/internal/config/config.go`
   - `services/merchant-service/internal/config/config.go`
   - `services/risk-assessment-service/internal/config/config.go`
   
   **Common Fields:**
   - Port (string)
   - Host (string)
   - ReadTimeout (time.Duration)
   - WriteTimeout (time.Duration)
   - IdleTimeout (time.Duration)

2. **SupabaseConfig Structure** - Duplicated across all services
   - Same structure in all 4 services
   
   **Common Fields:**
   - URL (string)
   - APIKey (string)
   - ServiceRoleKey (string)
   - JWTSecret (string)

3. **Helper Functions** - Duplicated across all services
   - `getEnvAsString()`
   - `getEnvAsInt()`
   - `getEnvAsBool()`
   - `getEnvAsDuration()`

### Impact Assessment

**Lines of Duplicated Code:**
- ServerConfig: ~15 lines × 4 services = 60 lines
- SupabaseConfig: ~10 lines × 4 services = 40 lines
- Helper functions: ~50 lines × 4 services = 200 lines
- **Total**: ~300 lines of duplicated configuration code

**Maintenance Impact:**
- Changes to configuration structure require updates in 4 places
- Bug fixes must be applied to 4 files
- Inconsistent behavior if updates are missed

### Recommendation

**Create Shared Configuration Package:**
- Create `pkg/config/` or `internal/shared/config/`
- Extract common structures and helper functions
- Services import and extend with service-specific config

**Estimated Reduction:**
- ~300 lines of duplicated code → ~100 lines of shared code
- **Reduction**: ~200 lines (67% reduction)

---

## Handler Code Patterns

### Similar Patterns Found

**All handlers use similar patterns:**
- Import statements: `context`, `encoding/json`, `fmt`, `net/http`
- JSON encoding/decoding
- Error handling
- HTTP response writing

**Example from handlers:**
- `services/api-gateway/internal/handlers/gateway.go`: 29 functions/types/interfaces
- `services/classification-service/internal/handlers/classification.go`: 43 functions/types/interfaces
- `services/merchant-service/internal/handlers/merchant.go`: 45 functions/types/interfaces

**Common Patterns:**
- Request parsing
- Response formatting
- Error handling
- Context usage

### Recommendation

**Create Shared Handler Utilities:**
- Request parsing helpers
- Response formatting helpers
- Common error handling middleware
- Context utilities

---

## Health Check Patterns

### Duplication Found

**Health check implementations found in:**
- `services/api-gateway/cmd/main.go`
- `services/classification-service/cmd/main.go`
- `services/merchant-service/internal/handlers/merchant.go`
- `services/risk-assessment-service/cmd/main.go`
- `services/merchant-service/internal/observability/health_checker.go`
- `services/merchant-service/internal/observability/monitoring.go`

**Common Pattern:**
- Health endpoint handler
- Status checking
- JSON response formatting

### Recommendation

**Create Shared Health Check Package:**
- Standard health check handler
- Health check utilities
- Status aggregation

---

## Constructor Patterns

### Analysis

**Found 488 constructor/creator functions across 266 files**

**Common Patterns:**
- `New*()` functions for creating instances
- `Create*()` functions for creating resources
- Dependency injection patterns

**Services with most constructors:**
- `services/risk-assessment-service/`: Extensive use of constructors
- `services/merchant-service/`: Many observability constructors
- All services: Configuration constructors

### Recommendation

**Standardize Constructor Patterns:**
- Document naming conventions
- Ensure consistent dependency injection
- Consider factory patterns for complex objects

---

## Summary

### Code Duplication Metrics

| Category | Duplicated Lines | Services Affected | Reduction Potential |
|----------|------------------|-------------------|---------------------|
| Configuration | ~300 | 4 | 67% (~200 lines) |
| Health Checks | ~150 | 6+ | 60% (~90 lines) |
| Handler Patterns | ~200 | 3+ | 50% (~100 lines) |
| **Total** | **~650** | **Multiple** | **~390 lines** |

### Priority Recommendations

1. **HIGH**: Extract shared configuration package
   - Impact: High (affects all services)
   - Effort: Medium
   - Benefit: Consistent configuration, easier maintenance

2. **MEDIUM**: Create shared health check utilities
   - Impact: Medium (affects observability)
   - Effort: Low
   - Benefit: Consistent health checks

3. **MEDIUM**: Standardize handler patterns
   - Impact: Medium (affects API consistency)
   - Effort: Medium
   - Benefit: Consistent API responses

---

**Last Updated**: 2025-11-10 01:30 UTC

