# Migration Quick Reference Guide

## Overview
Quick reference for restoring functionality from consolidation branch to main branch.

## Priority Order

### 🔴 CRITICAL (Do First)
1. **Database Connection** - Required for all persistence features
2. **PostgREST Client Standardization** - Required for Supabase operations
3. **Threshold CRUD Handlers** - Core admin functionality

### 🟡 HIGH (Do Second)
4. **Export/Import Handlers** - Data management
5. **Get Risk Factors/Categories** - API completeness
6. **Error Handling Standardization** - Code quality

### 🟢 MEDIUM (Do Third)
7. **Recommendation Rules Handlers** - Advanced features
8. **Notification Channel Handlers** - Advanced features
9. **System Monitoring Handlers** - Operations support
10. **Risk Assessment Service Database Logic** - Data persistence

### ⚪ LOW (Optional)
11. **Redis Optimization** - Performance enhancement
12. **Frontend Debug Scripts** - Development tools

## File Locations

### Backend Handlers
- **Main Handler File**: `internal/api/handlers/enhanced_risk.go`
- **Routes File**: `internal/api/routes/enhanced_risk_routes.go`
- **Server Main**: `cmd/railway-server/main.go`
- **Risk Service Handler**: `services/risk-assessment-service/internal/handlers/risk_assessment.go`

### Key Functions to Restore

#### Threshold Management
- `GetRiskThresholdsHandler` - Line ~817
- `CreateRiskThresholdHandler` - Line ~900
- `UpdateRiskThresholdHandler` - Line ~995
- `DeleteRiskThresholdHandler` - Line ~1096
- `ExportThresholdsHandler` - Line ~1839
- `ImportThresholdsHandler` - Line ~1896

#### Recommendation Rules
- `CreateRecommendationRuleHandler` - Line ~1162
- `UpdateRecommendationRuleHandler` - Line ~1232
- `DeleteRecommendationRuleHandler` - Line ~1311

#### Notification Channels
- `CreateNotificationChannelHandler` - Line ~1374
- `UpdateNotificationChannelHandler` - Line ~1461
- `DeleteNotificationChannelHandler` - Line ~1570

#### System Monitoring
- `GetSystemHealthHandler` - Line ~1641
- `GetSystemMetricsHandler` - Line ~1679
- `CleanupSystemDataHandler` - Line ~1725

#### Risk Factors/Categories
- `GetRiskFactorsHandler` - Line ~644
- `GetRiskCategoriesHandler` - Line ~755

## Helper Functions Needed

```go
// Request ID extraction
func getRequestID(r *http.Request) string {
    if id := r.Header.Get("X-Request-ID"); id != "" {
        return id
    }
    if id, ok := r.Context().Value("request_id").(string); ok {
        return id
    }
    return uuid.New().String()
}

// Path ID extraction
func extractIDFromPath(path, prefix string) string {
    id := strings.TrimPrefix(path, prefix)
    id = strings.TrimSuffix(id, "/")
    if idx := strings.Index(id, "?"); idx != -1 {
        id = id[:idx]
    }
    return id
}
```

## Dependencies to Add

### Go Modules
```bash
# Standardize PostgREST client
go get github.com/supabase-community/postgrest-go@v0.0.11

# Ensure error helper is available
go get kyb-platform/pkg/errors@latest

# Database driver
go get github.com/lib/pq
```

### Environment Variables
```bash
DATABASE_URL=postgres://user:pass@host:port/dbname
REDIS_URL=redis://host:port  # Optional
```

## Testing Commands

### Test Threshold CRUD
```bash
# Create
curl -X POST http://localhost:8080/v1/admin/risk/thresholds \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","category":"financial","risk_levels":{"low":25,"medium":50,"high":75}}'

# Get
curl http://localhost:8080/v1/risk/thresholds

# Update
curl -X PUT http://localhost:8080/v1/admin/risk/thresholds/{id} \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated"}'

# Delete
curl -X DELETE http://localhost:8080/v1/admin/risk/thresholds/{id}
```

### Test Export/Import
```bash
# Export
curl http://localhost:8080/v1/admin/risk/threshold-export > thresholds.json

# Import
curl -X POST http://localhost:8080/v1/admin/risk/threshold-import \
  -H "Content-Type: application/json" \
  -d @thresholds.json
```

## Common Issues & Solutions

### Issue: ThresholdManager is nil
**Solution**: Ensure ThresholdManager is initialized in server setup:
```go
thresholdManager := risk.NewThresholdManager()
if db != nil {
    thresholdRepo := database.NewThresholdRepository(db, logger)
    thresholdManager = risk.NewThresholdManagerWithRepository(thresholdRepo)
    thresholdManager.LoadFromDatabase(context.Background())
}
```

### Issue: Database connection fails
**Solution**: 
1. Check DATABASE_URL environment variable
2. Verify database is running
3. Check connection string format
4. Verify network connectivity

### Issue: Route conflicts
**Solution**: Ensure routes are registered in correct order:
1. Specific paths first (e.g., `/threshold-export`)
2. Wildcard paths last (e.g., `/thresholds/{id}`)

### Issue: Error handling inconsistency
**Solution**: Use standardized error helper:
```go
import errorspkg "kyb-platform/pkg/errors"

// Instead of:
http.Error(w, "message", http.StatusBadRequest)

// Use:
errorspkg.WriteBadRequest(w, r, "message")
```

## Checklist Template

For each handler restoration:
- [ ] Copy handler code from main branch
- [ ] Update error handling to use standardized helper
- [ ] Verify route registration
- [ ] Add unit tests
- [ ] Test manually with curl/Postman
- [ ] Verify database persistence (if applicable)
- [ ] Update documentation
- [ ] Commit changes

## Estimated Time per Phase

- Phase 1 (Foundation): 3 days
- Phase 2 (Core Admin): 5 days
- Phase 3 (Rules/Channels): 4 days
- Phase 4 (Monitoring): 3 days
- Phase 5 (Risk Service): 3 days
- Phase 6 (Error Handling): 2 days
- Phase 7 (Frontend): 2 days

**Total**: ~22 days (4-5 weeks)

## Rollback Commands

```bash
# Full rollback
git checkout consolidation-branch

# Partial rollback (specific file)
git checkout consolidation-branch -- path/to/file

# Create backup before starting
git branch backup-before-migration
```

## Success Indicators

- ✅ All 15 handlers restored
- ✅ Database persistence working
- ✅ All tests passing
- ✅ No regression in existing features
- ✅ Error handling consistent
- ✅ Documentation updated













