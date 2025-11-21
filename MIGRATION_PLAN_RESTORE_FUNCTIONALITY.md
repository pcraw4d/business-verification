# Migration Plan: Restore Removed Functionality

## Executive Summary

This document outlines a comprehensive plan to restore critical functionality that was removed in the consolidation branch. The restoration is prioritized by business impact and technical dependencies.

**Estimated Timeline**: 2-3 weeks  
**Risk Level**: Medium-High  
**Dependencies**: Database setup, PostgREST client standardization

---

## Phase 1: Foundation & Infrastructure (Days 1-3)

### Priority: CRITICAL
**Goal**: Restore database and infrastructure support required for all other features

### 1.1 Restore Database Connection Logic

**File**: `cmd/railway-server/main.go`

**Steps**:
1. Add database imports:
```go
import (
    "context"
    "database/sql"
    _ "github.com/lib/pq"
    // ... existing imports
)
```

2. Add database field to `RailwayServer` struct:
```go
type RailwayServer struct {
    // ... existing fields
    db *sql.DB
}
```

3. Restore database initialization in `NewRailwayServer()`:
```go
// Initialize database connection for new routes
var db *sql.DB
databaseURL := os.Getenv("DATABASE_URL")
if databaseURL != "" {
    var err error
    db, err = sql.Open("postgres", databaseURL)
    if err != nil {
        log.Printf("Warning: Failed to connect to database: %v. New API routes will not be available.", err)
    } else {
        // Test connection with context timeout
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        if err := db.PingContext(ctx); err != nil {
            log.Printf("Warning: Database ping failed: %v. Routes will use in-memory storage.", err)
            db.Close()
            db = nil
            log.Println("⚠️  Using in-memory threshold storage (database unavailable)")
        } else {
            log.Println("✅ Database connection established for new API routes")
        }
    }
} else {
    log.Println("Warning: DATABASE_URL not set. New API routes will not be available.")
}
```

4. Restore database health check in `handleDetailedHealth()`:
```go
// Check PostgreSQL database (if available)
if s.db != nil {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := s.db.PingContext(ctx); err != nil {
        checks["postgres"] = map[string]string{"status": "unhealthy", "error": err.Error()}
    } else {
        checks["postgres"] = map[string]string{"status": "healthy"}
    }
} else {
    checks["postgres"] = map[string]string{"status": "not_configured"}
}
```

**Testing**:
- [ ] Verify database connection on startup
- [ ] Test health check endpoint with/without database
- [ ] Verify graceful fallback to in-memory storage

**Dependencies**: None

---

### 1.2 Standardize PostgREST Client

**File**: `cmd/railway-server/main.go`

**Steps**:
1. Change import from:
```go
"github.com/supabase/postgrest-go"
```
to:
```go
"github.com/supabase-community/postgrest-go"
```

2. Update `go.mod`:
```bash
go get github.com/supabase-community/postgrest-go@v0.0.11
go mod tidy
```

**Testing**:
- [ ] Verify Supabase client initialization
- [ ] Test existing endpoints that use Supabase client

**Dependencies**: None

---

### 1.3 Restore Redis Optimization (Optional - Phase 2)

**Priority**: LOW (can be deferred)

**File**: `cmd/railway-server/main.go`

**Steps**:
1. Add Redis optimizer field to `RailwayServer`
2. Restore Redis initialization logic
3. Restore Redis health checks
4. Restore Redis caching in classification endpoint

**Note**: This can be deferred if Redis is not critical for MVP

---

## Phase 2: Core Admin Handlers (Days 4-8)

### Priority: HIGH
**Goal**: Restore threshold management and admin functionality

### 2.1 Restore Threshold CRUD Handlers

**Files**: 
- `internal/api/handlers/enhanced_risk.go`
- `internal/api/routes/enhanced_risk_routes.go`

**Handlers to Restore**:
1. `GetRiskThresholdsHandler` - GET /v1/risk/thresholds
2. `CreateRiskThresholdHandler` - POST /v1/admin/risk/thresholds
3. `UpdateRiskThresholdHandler` - PUT /v1/admin/risk/thresholds/{id}
4. `DeleteRiskThresholdHandler` - DELETE /v1/admin/risk/thresholds/{id}

**Steps**:

1. **Verify ThresholdManager is initialized** in `cmd/railway-server/main.go`:
```go
// In setupNewAPIRoutes() or similar
thresholdManager := risk.NewThresholdManager()
// If database is available, use repository-based manager
if db != nil {
    thresholdRepo := database.NewThresholdRepository(db, logger)
    thresholdManager = risk.NewThresholdManagerWithRepository(thresholdRepo)
    // Load existing thresholds from database
    if err := thresholdManager.LoadFromDatabase(context.Background()); err != nil {
        log.Printf("Warning: Failed to load thresholds from database: %v", err)
    }
}
```

2. **Restore handler implementations** from main branch:
   - Copy `GetRiskThresholdsHandler` (lines 817-899 in main)
   - Copy `CreateRiskThresholdHandler` (lines 900-993 in main)
   - Copy `UpdateRiskThresholdHandler` (lines 995-1094 in main)
   - Copy `DeleteRiskThresholdHandler` (lines 1096-1161 in main)

3. **Update handler registration** in `internal/api/routes/enhanced_risk_routes.go`:
```go
// In RegisterEnhancedRiskAdminRoutes
mux.Handle("POST /v1/admin/risk/thresholds",
    corsMiddleware.Middleware(
        loggingMiddleware.Middleware(
            http.HandlerFunc(enhancedRiskHandler.CreateRiskThresholdHandler))))

mux.Handle("PUT /v1/admin/risk/thresholds/{threshold_id}",
    corsMiddleware.Middleware(
        loggingMiddleware.Middleware(
            http.HandlerFunc(enhancedRiskHandler.UpdateRiskThresholdHandler))))

mux.Handle("DELETE /v1/admin/risk/thresholds/{threshold_id}",
    corsMiddleware.Middleware(
        loggingMiddleware.Middleware(
            http.HandlerFunc(enhancedRiskHandler.DeleteRiskThresholdHandler))))
```

4. **Restore helper functions**:
```go
func getRequestID(r *http.Request) string {
    // Extract from header or context
    if id := r.Header.Get("X-Request-ID"); id != "" {
        return id
    }
    if id, ok := r.Context().Value("request_id").(string); ok {
        return id
    }
    return uuid.New().String()
}

func extractIDFromPath(path, prefix string) string {
    id := strings.TrimPrefix(path, prefix)
    // Remove trailing slashes and query params
    id = strings.TrimSuffix(id, "/")
    if idx := strings.Index(id, "?"); idx != -1 {
        id = id[:idx]
    }
    return id
}
```

**Testing**:
- [ ] Test GET /v1/risk/thresholds (empty, with data)
- [ ] Test POST /v1/admin/risk/thresholds (create threshold)
- [ ] Test PUT /v1/admin/risk/thresholds/{id} (update threshold)
- [ ] Test DELETE /v1/admin/risk/thresholds/{id} (delete threshold)
- [ ] Test database persistence (restart server, verify thresholds persist)
- [ ] Test validation (invalid risk levels, missing fields)
- [ ] Test error cases (non-existent ID, database errors)

**Dependencies**: Phase 1.1 (Database connection)

---

### 2.2 Restore Export/Import Handlers

**Files**: 
- `internal/api/handlers/enhanced_risk.go`
- `internal/api/routes/enhanced_risk_routes.go`

**Handlers to Restore**:
1. `ExportThresholdsHandler` - GET /v1/admin/risk/threshold-export
2. `ImportThresholdsHandler` - POST /v1/admin/risk/threshold-import

**Steps**:

1. **Restore handler implementations** from main branch:
   - Copy `ExportThresholdsHandler` (lines 1839-1895 in main)
   - Copy `ImportThresholdsHandler` (lines 1896-1979 in main)

2. **Verify route registration** (should already exist):
```go
mux.Handle("GET /v1/admin/risk/threshold-export", ...)
mux.Handle("POST /v1/admin/risk/threshold-import", ...)
```

**Testing**:
- [ ] Test export with no thresholds (empty JSON)
- [ ] Test export with multiple thresholds
- [ ] Test import with valid JSON
- [ ] Test import with invalid JSON (validation errors)
- [ ] Test import with duplicate IDs (should update or error)
- [ ] Test round-trip (export, modify, import, verify)

**Dependencies**: Phase 2.1 (Threshold CRUD)

---

### 2.3 Restore Get Risk Factors/Categories Handlers

**Files**: 
- `internal/api/handlers/enhanced_risk.go`
- `internal/api/routes/enhanced_risk_routes.go`

**Handlers to Restore**:
1. `GetRiskFactorsHandler` - GET /v1/risk/factors
2. `GetRiskCategoriesHandler` - GET /v1/risk/categories

**Steps**:

1. **Restore handler implementations** from main branch:
   - Copy `GetRiskFactorsHandler` (lines 644-754 in main)
   - Copy `GetRiskCategoriesHandler` (lines 755-816 in main)

2. **Verify route registration**:
```go
mux.Handle("GET /v1/risk/factors", ...)
mux.Handle("GET /v1/risk/categories", ...)
```

**Testing**:
- [ ] Test GET /v1/risk/factors (all factors, filtered by category)
- [ ] Test GET /v1/risk/categories (all categories)
- [ ] Test query parameters (category filter)

**Dependencies**: None (uses mock data)

---

## Phase 3: Recommendation Rules & Notification Channels (Days 9-12)

### Priority: MEDIUM
**Goal**: Restore recommendation rules and notification channel management

### 3.1 Restore Recommendation Rule Handlers

**Files**: 
- `internal/api/handlers/enhanced_risk.go`
- `internal/api/routes/enhanced_risk_routes.go`

**Handlers to Restore**:
1. `CreateRecommendationRuleHandler` - POST /v1/admin/risk/recommendation-rules
2. `UpdateRecommendationRuleHandler` - PUT /v1/admin/risk/recommendation-rules/{id}
3. `DeleteRecommendationRuleHandler` - DELETE /v1/admin/risk/recommendation-rules/{id}

**Steps**:

1. **Verify RecommendationRuleEngine is accessible**:
```go
// In handler initialization
recommendationEngine := risk.NewRiskRecommendationEngine(...)
// Ensure GetRuleEngine() method exists
```

2. **Restore handler implementations** from main branch:
   - Copy `CreateRecommendationRuleHandler` (lines 1162-1231 in main)
   - Copy `UpdateRecommendationRuleHandler` (lines 1232-1310 in main)
   - Copy `DeleteRecommendationRuleHandler` (lines 1311-1373 in main)

3. **Restore helper functions**:
```go
// Helper functions for notification channels (if needed)
func (h *EnhancedRiskHandler) createNotificationChannelFromRequest(...) risk.NotificationChannel
func (h *EnhancedRiskHandler) getChannelType(...) string
func contains(slice []string, item string) bool
```

**Testing**:
- [ ] Test create recommendation rule
- [ ] Test update recommendation rule
- [ ] Test delete recommendation rule
- [ ] Test validation (invalid conditions, missing fields)
- [ ] Test error cases (non-existent ID)

**Dependencies**: RecommendationRuleEngine must be initialized

---

### 3.2 Restore Notification Channel Handlers

**Files**: 
- `internal/api/handlers/enhanced_risk.go`
- `internal/api/routes/enhanced_risk_routes.go`

**Handlers to Restore**:
1. `CreateNotificationChannelHandler` - POST /v1/admin/risk/notification-channels
2. `UpdateNotificationChannelHandler` - PUT /v1/admin/risk/notification-channels/{id}
3. `DeleteNotificationChannelHandler` - DELETE /v1/admin/risk/notification-channels/{id}

**Steps**:

1. **Verify NotificationService is accessible**:
```go
// In handler initialization
alertSystem := risk.NewRiskAlertSystem(...)
notificationService := alertSystem.GetNotificationService()
```

2. **Restore handler implementations** from main branch:
   - Copy `CreateNotificationChannelHandler` (lines 1374-1460 in main)
   - Copy `UpdateNotificationChannelHandler` (lines 1461-1569 in main)
   - Copy `DeleteNotificationChannelHandler` (lines 1570-1640 in main)

3. **Verify channel type constructors exist** in `internal/risk/notification_channels.go`:
```go
NewEmailNotificationChannel(...)
NewSMSNotificationChannel(...)
NewSlackNotificationChannel(...)
NewWebhookNotificationChannel(...)
// etc.
```

**Testing**:
- [ ] Test create notification channel (all types: email, SMS, Slack, webhook)
- [ ] Test update notification channel
- [ ] Test delete notification channel
- [ ] Test validation (invalid config, missing fields)
- [ ] Test error cases (non-existent ID, invalid channel type)

**Dependencies**: NotificationService must be initialized

---

## Phase 4: System Monitoring (Days 13-15)

### Priority: MEDIUM
**Goal**: Restore system health and monitoring endpoints

### 4.1 Restore System Health/Metrics/Cleanup Handlers

**Files**: 
- `internal/api/handlers/enhanced_risk.go`
- `internal/api/routes/enhanced_risk_routes.go`

**Handlers to Restore**:
1. `GetSystemHealthHandler` - GET /v1/admin/risk/system/health
2. `GetSystemMetricsHandler` - GET /v1/admin/risk/system/metrics
3. `CleanupSystemDataHandler` - POST /v1/admin/risk/system/cleanup

**Steps**:

1. **Restore handler implementations** from main branch:
   - Copy `GetSystemHealthHandler` (lines 1641-1678 in main)
   - Copy `GetSystemMetricsHandler` (lines 1679-1724 in main)
   - Copy `CleanupSystemDataHandler` (lines 1725-1838 in main)

2. **Verify route registration**:
```go
mux.Handle("GET /v1/admin/risk/system/health", ...)
mux.Handle("GET /v1/admin/risk/system/metrics", ...)
mux.Handle("POST /v1/admin/risk/system/cleanup", ...)
```

**Testing**:
- [ ] Test system health endpoint (verify all checks)
- [ ] Test system metrics endpoint (verify metrics collection)
- [ ] Test cleanup endpoint (verify data cleanup, verify safety checks)

**Dependencies**: None (uses existing services)

---

## Phase 5: Risk Assessment Service (Days 16-18)

### Priority: MEDIUM
**Goal**: Restore database-backed risk assessment retrieval

### 5.1 Restore HandleGetRiskAssessment

**File**: `services/risk-assessment-service/internal/handlers/risk_assessment.go`

**Steps**:

1. **Restore implementation** from main branch (before it was stubbed):
```go
func (h *RiskAssessmentHandler) HandleGetRiskAssessment(w http.ResponseWriter, r *http.Request) {
    // Extract assessment ID from URL
    vars := mux.Vars(r)
    assessmentID := vars["id"]
    
    // Query Supabase for the assessment
    // Parse and return response
    // ... (full implementation from main branch)
}
```

2. **Update error handling** to use standardized error helper:
```go
// Replace http.Error with errorspkg.Write* functions
import errorspkg "kyb-platform/pkg/errors"

// Use:
errorspkg.WriteBadRequest(w, r, "message")
errorspkg.WriteNotFound(w, r, "message")
errorspkg.WriteInternalError(w, r, "message")
```

**Testing**:
- [ ] Test GET /api/v1/assess/{id} with valid ID
- [ ] Test GET /api/v1/assess/{id} with non-existent ID (404)
- [ ] Test GET /api/v1/assess/{id} with invalid ID format (400)
- [ ] Test database error handling (500)

**Dependencies**: Supabase client must be initialized

---

### 5.2 Restore Database Query Logic in HandleRiskPrediction

**File**: `services/risk-assessment-service/internal/handlers/risk_assessment.go`

**Steps**:

1. **Restore database query logic** from main branch:
```go
// Extract assessment ID from URL and retrieve business data from database
vars := mux.Vars(r)
assessmentID := vars["id"]

var business *models.RiskAssessmentRequest
if assessmentID != "" {
    // Try to get assessment from database
    // Extract business data from assessment
    // ... (full implementation from main branch)
}
```

2. **Keep fallback to mock data** if assessment not found (for development)

**Testing**:
- [ ] Test prediction with valid assessment ID (uses database data)
- [ ] Test prediction with non-existent ID (uses fallback)
- [ ] Test prediction without ID (uses fallback)

**Dependencies**: Phase 5.1

---

## Phase 6: Error Handling Standardization (Days 19-20)

### Priority: HIGH
**Goal**: Ensure consistent error handling across all services

### 6.1 Standardize Error Handling in Risk Assessment Service

**File**: `services/risk-assessment-service/internal/handlers/risk_assessment.go`

**Steps**:

1. **Remove internal errorHandler**:
```go
// Remove:
errorHandler *middleware.ErrorHandler

// Remove initialization:
errorHandler: middleware.NewErrorHandler(logger),
```

2. **Add standardized error helper**:
```go
import errorspkg "kyb-platform/pkg/errors"
```

3. **Update all error responses**:
```go
// Replace:
h.errorHandler.HandleError(w, r, err)

// With:
errorspkg.WriteBadRequest(w, r, "message")
errorspkg.WriteNotFound(w, r, "message")
errorspkg.WriteInternalError(w, r, "message")
```

4. **Update go.mod**:
```bash
# Ensure kyb-platform is a dependency
go get kyb-platform/pkg/errors@latest
go mod tidy
```

**Testing**:
- [ ] Verify all error responses use standardized format
- [ ] Test error responses return correct status codes
- [ ] Test error responses have consistent JSON structure

**Dependencies**: None

---

### 6.2 Standardize Request ID Extraction

**Files**: All handler files

**Steps**:

1. **Create helper function** in `internal/api/handlers/enhanced_risk.go`:
```go
func getRequestID(r *http.Request) string {
    // Try header first
    if id := r.Header.Get("X-Request-ID"); id != "" {
        return id
    }
    // Try context
    if id, ok := r.Context().Value("request_id").(string); ok {
        return id
    }
    // Generate new UUID
    return uuid.New().String()
}
```

2. **Update all handlers** to use `getRequestID(r)` instead of `r.Context().Value("request_id").(string)`

**Testing**:
- [ ] Verify request IDs are logged correctly
- [ ] Test with X-Request-ID header
- [ ] Test without header (should use context or generate)

**Dependencies**: None

---

## Phase 7: Frontend Enhancements (Days 21-22)

### Priority: LOW
**Goal**: Restore robust frontend error handling (optional)

### 7.1 Restore Debug Scripts (Optional)

**Files**: 
- `services/frontend/public/merchant-details.html`
- `services/frontend/public/add-merchant.html`

**Steps**:

1. **Restore script references** (if needed for debugging):
```html
<script src="js/api-config.js"></script>
<script src="js/debug-form-flow.js"></script>
```

2. **Restore complex DOM waiting logic** (if needed):
   - MutationObserver for tab content
   - Retry mechanisms for element finding
   - Multi-strategy element queries

**Note**: This is optional - simple `getElementById()` may be sufficient

**Testing**:
- [ ] Test form submission flow
- [ ] Test tab switching
- [ ] Test field population
- [ ] Test error handling

**Dependencies**: None

---

## Testing Strategy

### Unit Tests
- [ ] Test all restored handlers individually
- [ ] Test helper functions
- [ ] Test error cases
- [ ] Test validation logic

### Integration Tests
- [ ] Test database persistence
- [ ] Test threshold CRUD operations
- [ ] Test export/import round-trip
- [ ] Test recommendation rules integration
- [ ] Test notification channels integration

### End-to-End Tests
- [ ] Test complete admin workflow (create threshold → export → import)
- [ ] Test risk assessment with database persistence
- [ ] Test system health monitoring
- [ ] Test error handling across all endpoints

### Manual Testing Checklist
- [ ] Verify all endpoints return correct status codes
- [ ] Verify database persistence works
- [ ] Verify error messages are user-friendly
- [ ] Verify logging includes request IDs
- [ ] Verify health checks work correctly

---

## Rollback Plan

If issues arise during migration:

1. **Immediate Rollback**: Revert to consolidation branch
2. **Partial Rollback**: Disable specific handlers via feature flags
3. **Database Rollback**: Restore from backup if database schema changed

---

## Success Criteria

- [ ] All 15 removed handlers are restored and functional
- [ ] Database persistence works for thresholds
- [ ] Export/import functionality works correctly
- [ ] Error handling is standardized across all services
- [ ] All tests pass
- [ ] No regression in existing functionality
- [ ] Documentation updated

---

## Risk Mitigation

### High-Risk Areas:
1. **Database Schema Changes**: Ensure migrations are backward compatible
2. **Service Dependencies**: Verify all services are initialized correctly
3. **Error Handling**: Ensure consistent error responses
4. **Route Conflicts**: Verify no route conflicts with existing routes

### Mitigation Strategies:
1. Test in staging environment first
2. Use feature flags for gradual rollout
3. Monitor error rates and logs
4. Have rollback plan ready

---

## Dependencies Checklist

Before starting migration, ensure:
- [ ] Database is accessible and configured
- [ ] Database schema includes `risk_thresholds` table
- [ ] PostgREST client is standardized
- [ ] All required services are initialized
- [ ] Test environment is set up
- [ ] Backup of current state is created

---

## Notes

- This migration should be done incrementally, one phase at a time
- Test thoroughly after each phase before proceeding
- Keep the consolidation branch as a reference
- Document any deviations from this plan
- Update this document as migration progresses

---

**Last Updated**: [Date]  
**Status**: Planning  
**Owner**: [Name]











