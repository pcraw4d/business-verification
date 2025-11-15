<!-- aca1c262-05c3-4b54-85d8-79bea25a21ce c50f8e7e-7d57-4158-a73e-38f770e56e46 -->
# Detailed Implementation Plan: Restore Functionality

## Overview

This plan verifies and ensures all restored functionality works correctly. Most handlers are already implemented; focus is on validation, testing, error handling standardization, and ensuring no enhanced functionality is lost.

## Phase 1: Foundation Verification & Testing (Days 1-2)

### 1.1 Verify Database Connection & Infrastructure

**Files**: `cmd/railway-server/main.go`

**Tasks**:

- [ ] Verify database connection logic is working (lines 58-82)
- [ ] Test database health check in `handleDetailedHealth()` (lines 260-271)
- [ ] Verify graceful fallback to in-memory storage when database unavailable
- [ ] Test connection pooling settings
- [ ] Verify database connection cleanup on shutdown

**Testing**:

```bash
# Test with database
DATABASE_URL=postgres://... go run cmd/railway-server/main.go
curl http://localhost:8080/health/detailed

# Test without database
unset DATABASE_URL
go run cmd/railway-server/main.go
curl http://localhost:8080/health/detailed
```

**Success Criteria**:

- Database connects successfully when DATABASE_URL is set
- Health check shows "healthy" for postgres when connected
- Health check shows "not_configured" when DATABASE_URL not set
- Server starts without errors in both scenarios

---

### 1.2 Verify PostgREST Client Standardization

**Files**: `cmd/railway-server/main.go`

**Tasks**:

- [ ] Verify import uses `github.com/supabase-community/postgrest-go` (line 15)
- [ ] Test Supabase client initialization (lines 48-55)
- [ ] Verify existing endpoints using Supabase client still work
- [ ] Check `go.mod` for correct dependency version

**Testing**:

```bash
# Verify dependency
go list -m github.com/supabase-community/postgrest-go

# Test Supabase-dependent endpoints
curl http://localhost:8080/v1/classify -X POST -d '{"name":"Test","description":"Test"}'
```

**Success Criteria**:

- Correct PostgREST client package is imported
- Supabase client initializes without errors
- Existing endpoints work correctly

---

### 1.3 Verify Redis Optimization (Optional)

**Files**: `cmd/railway-server/main.go`

**Tasks**:

- [ ] Verify Redis optimizer initialization (lines 89-112)
- [ ] Test Redis health check in `handleDetailedHealth()` (lines 273-296)
- [ ] Test Redis caching in classification endpoint (lines 332-398)
- [ ] Verify graceful fallback when Redis unavailable

**Testing**:

```bash
# Test with Redis
REDIS_URL=redis://localhost:6379 go run cmd/railway-server/main.go
curl http://localhost:8080/health/detailed
curl http://localhost:8080/redis-optimization

# Test without Redis
unset REDIS_URL
go run cmd/railway-server/main.go
curl http://localhost:8080/health/detailed
```

**Success Criteria**:

- Redis optimizer initializes when REDIS_URL is set
- Health check shows Redis status correctly
- Classification endpoint uses Redis cache when available
- Server works without Redis (graceful degradation)

---

## Phase 2: Core Admin Handlers Verification (Days 3-5)

### 2.1 Verify Threshold CRUD Handlers

**Files**:

- `internal/api/handlers/enhanced_risk.go` (lines 817-1159)
- `internal/api/routes/enhanced_risk_routes.go` (lines 110-124)
- `cmd/railway-server/main.go` (lines 1090-1119)

**Tasks**:

- [ ] Verify ThresholdManager initialization with database (lines 1090-1119)
- [ ] Verify GetRiskThresholdsHandler implementation (lines 817-898)
- [ ] Verify CreateRiskThresholdHandler implementation (lines 900-994)
- [ ] Verify UpdateRiskThresholdHandler implementation (lines 995-1095)
- [ ] Verify DeleteRiskThresholdHandler implementation (lines 1096-1159)
- [ ] Verify route registration (lines 110-124)
- [ ] Test database persistence (create, restart server, verify persistence)
- [ ] Test in-memory fallback when database unavailable

**Testing**:

```bash
# Test GET (should return empty or existing thresholds)
curl http://localhost:8080/v1/risk/thresholds

# Test CREATE
curl -X POST http://localhost:8080/v1/admin/risk/thresholds \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Threshold",
    "category": "financial",
    "risk_levels": {
      "low": 25.0,
      "medium": 50.0,
      "high": 75.0,
      "critical": 90.0
    },
    "is_active": true,
    "priority": 1
  }'

# Test GET again (should return new threshold)
curl http://localhost:8080/v1/risk/thresholds

# Test UPDATE (use ID from create response)
curl -X PUT http://localhost:8080/v1/admin/risk/thresholds/{id} \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Threshold"}'

# Test DELETE
curl -X DELETE http://localhost:8080/v1/admin/risk/thresholds/{id}

# Test persistence: restart server, verify threshold still exists
```

**Success Criteria**:

- All CRUD operations work correctly
- Thresholds persist to database
- Thresholds load from database on startup
- In-memory fallback works when database unavailable
- Proper error handling for invalid requests
- Request IDs are logged correctly

---

### 2.2 Verify Export/Import Handlers

**Files**:

- `internal/api/handlers/enhanced_risk.go` (lines 1838-1964)
- `internal/api/routes/enhanced_risk_routes.go` (lines 100-108)

**Tasks**:

- [ ] Verify ExportThresholdsHandler implementation (lines 1838-1893)
- [ ] Verify ImportThresholdsHandler implementation (lines 1895-1964)
- [ ] Test export with no thresholds (should return empty JSON)
- [ ] Test export with multiple thresholds
- [ ] Test import with valid JSON
- [ ] Test import with invalid JSON (should return validation errors)
- [ ] Test round-trip (export, modify, import, verify)

**Testing**:

```bash
# Create some thresholds first
curl -X POST http://localhost:8080/v1/admin/risk/thresholds \
  -H "Content-Type: application/json" \
  -d '{"name":"Test1","category":"financial","risk_levels":{"low":25,"medium":50,"high":75}}'

# Test EXPORT
curl http://localhost:8080/v1/admin/risk/threshold-export > thresholds.json
cat thresholds.json

# Test IMPORT (modify JSON first)
curl -X POST http://localhost:8080/v1/admin/risk/threshold-import \
  -H "Content-Type: application/json" \
  -d @thresholds.json

# Verify imported thresholds
curl http://localhost:8080/v1/risk/thresholds
```

**Success Criteria**:

- Export returns valid JSON with all thresholds
- Import validates JSON structure
- Import creates/updates thresholds correctly
- Round-trip preserves all data
- Error handling for invalid JSON

---

### 2.3 Verify Get Risk Factors/Categories Handlers

**Files**:

- `internal/api/handlers/enhanced_risk.go` (lines 643-814)
- `internal/api/routes/enhanced_risk_routes.go` (lines 71-79)

**Tasks**:

- [ ] Verify GetRiskFactorsHandler implementation (lines 643-752)
- [ ] Verify GetRiskCategoriesHandler implementation (lines 754-814)
- [ ] Test GET /v1/risk/factors (all factors)
- [ ] Test GET /v1/risk/factors?category=financial (filtered)
- [ ] Test GET /v1/risk/categories (all categories)

**Testing**:

```bash
# Test GET all factors
curl http://localhost:8080/v1/risk/factors

# Test GET filtered by category
curl http://localhost:8080/v1/risk/factors?category=financial

# Test GET categories
curl http://localhost:8080/v1/risk/categories
```

**Success Criteria**:

- All endpoints return correct data
- Category filtering works
- Response format is consistent
- Request IDs are logged

---

## Phase 3: Recommendation Rules & Notification Channels (Days 6-8)

### 3.1 Verify Recommendation Rule Handlers

**Files**:

- `internal/api/handlers/enhanced_risk.go` (lines 1161-1371)
- `internal/api/routes/enhanced_risk_routes.go` (lines 126-139)

**Tasks**:

- [ ] Verify CreateRecommendationRuleHandler (lines 1161-1229)
- [ ] Verify UpdateRecommendationRuleHandler (lines 1231-1308)
- [ ] Verify DeleteRecommendationRuleHandler (lines 1310-1371)
- [ ] Verify RecommendationRuleEngine is accessible
- [ ] Test create, update, delete operations
- [ ] Test validation (invalid conditions, missing fields)
- [ ] Test error cases (non-existent ID)

**Testing**:

```bash
# Test CREATE
curl -X POST http://localhost:8080/v1/admin/risk/recommendation-rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Rule",
    "category": "financial",
    "conditions": [{"factor": "risk_score", "operator": ">", "value": 75}],
    "recommendations": [{"action": "review", "priority": "high"}],
    "enabled": true
  }'

# Test UPDATE
curl -X PUT http://localhost:8080/v1/admin/risk/recommendation-rules/{id} \
  -H "Content-Type: application/json" \
  -d '{"enabled": false}'

# Test DELETE
curl -X DELETE http://localhost:8080/v1/admin/risk/recommendation-rules/{id}
```

**Success Criteria**:

- All CRUD operations work
- Validation errors are returned for invalid data
- 404 errors for non-existent IDs
- Rules are stored in RecommendationRuleEngine

---

### 3.2 Verify Notification Channel Handlers

**Files**:

- `internal/api/handlers/enhanced_risk.go` (lines 1373-1638)
- `internal/api/routes/enhanced_risk_routes.go` (lines 141-154)

**Tasks**:

- [ ] Verify CreateNotificationChannelHandler (lines 1373-1458)
- [ ] Verify UpdateNotificationChannelHandler (lines 1460-1567)
- [ ] Verify DeleteNotificationChannelHandler (lines 1569-1638)
- [ ] Verify helper functions: createNotificationChannelFromRequest, getChannelType
- [ ] Test all channel types: email, SMS, Slack, webhook, Teams, Discord, PagerDuty
- [ ] Test validation (invalid config, missing fields)
- [ ] Test error cases (non-existent ID, invalid channel type)

**Testing**:

```bash
# Test CREATE email channel
curl -X POST http://localhost:8080/v1/admin/risk/notification-channels \
  -H "Content-Type: application/json" \
  -d '{
    "name": "email-alerts",
    "type": "email",
    "enabled": true,
    "config": {"recipients": ["admin@example.com"]}
  }'

# Test CREATE webhook channel
curl -X POST http://localhost:8080/v1/admin/risk/notification-channels \
  -H "Content-Type: application/json" \
  -d '{
    "name": "webhook-alerts",
    "type": "webhook",
    "enabled": true,
    "config": {"url": "https://example.com/webhook"}
  }'

# Test UPDATE
curl -X PUT http://localhost:8080/v1/admin/risk/notification-channels/email-alerts \
  -H "Content-Type: application/json" \
  -d '{"enabled": false}'

# Test DELETE
curl -X DELETE http://localhost:8080/v1/admin/risk/notification-channels/email-alerts
```

**Success Criteria**:

- All channel types can be created
- Update modifies channel configuration
- Delete removes channels
- Validation works for all channel types
- Error handling for invalid configurations

---

## Phase 4: System Monitoring (Days 9-10)

### 4.1 Verify System Health/Metrics/Cleanup Handlers

**Files**:

- `internal/api/handlers/enhanced_risk.go` (lines 1640-1836)
- `internal/api/routes/enhanced_risk_routes.go` (lines 156-170)

**Tasks**:

- [ ] Verify GetSystemHealthHandler (lines 1640-1676)
- [ ] Verify GetSystemMetricsHandler (lines 1678-1722)
- [ ] Verify CleanupSystemDataHandler (lines 1724-1836)
- [ ] Enhance health checks to include actual service status
- [ ] Enhance metrics to include real data collection
- [ ] Test cleanup with different data types

**Testing**:

```bash
# Test system health
curl http://localhost:8080/v1/admin/risk/system/health

# Test system metrics
curl http://localhost:8080/v1/admin/risk/system/metrics

# Test cleanup
curl -X POST http://localhost:8080/v1/admin/risk/system/cleanup \
  -H "Content-Type: application/json" \
  -d '{
    "older_than_days": 90,
    "data_types": ["alerts", "trends"]
  }'
```

**Success Criteria**:

- Health endpoint returns service status
- Metrics endpoint returns collected metrics
- Cleanup endpoint removes old data
- All endpoints log request IDs

---

## Phase 5: Error Handling Standardization (Days 11-12)

### 5.1 Standardize Error Handling in Enhanced Risk Handlers

**Files**: `internal/api/handlers/enhanced_risk.go`

**Tasks**:

- [ ] Review all error responses in handlers
- [ ] Replace `http.Error()` with standardized error helper where appropriate
- [ ] Ensure consistent error response format
- [ ] Verify all errors include request IDs
- [ ] Test error responses return correct status codes

**Current Status**: Handlers use `http.Error()` - need to verify if standardization is needed

**Testing**:

```bash
# Test various error scenarios
curl -X POST http://localhost:8080/v1/admin/risk/thresholds \
  -H "Content-Type: application/json" \
  -d '{}'  # Missing required fields

curl http://localhost:8080/v1/admin/risk/thresholds/nonexistent-id  # 404

curl -X PUT http://localhost:8080/v1/admin/risk/thresholds/invalid \
  -H "Content-Type: application/json" \
  -d 'invalid json'  # 400
```

**Success Criteria**:

- All error responses have consistent format
- Request IDs are included in error responses
- Status codes are correct
- Error messages are user-friendly

---

### 5.2 Verify Request ID Extraction

**Files**: `internal/api/handlers/enhanced_risk.go` (lines 1969-1976)

**Tasks**:

- [ ] Verify getRequestID helper function (lines 1969-1976)
- [ ] Enhance to check X-Request-ID header first (currently only checks context)
- [ ] Verify extractIDFromPath helper (lines 2041-2052)
- [ ] Test with X-Request-ID header
- [ ] Test without header (should use context or generate)

**Current Implementation**:

```go
func getRequestID(r *http.Request) string {
    if id := r.Context().Value("request_id"); id != nil {
        if str, ok := id.(string); ok {
            return str
        }
    }
    return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
```

**Enhancement Needed**: Add header check before context check

**Testing**:

```bash
# Test with X-Request-ID header
curl -H "X-Request-ID: test-123" \
  http://localhost:8080/v1/risk/thresholds

# Test without header (should generate)
curl http://localhost:8080/v1/risk/thresholds
```

**Success Criteria**:

- Request IDs are extracted from header if present
- Request IDs fall back to context if no header
- Request IDs are generated if neither available
- All handlers log request IDs correctly

---

## Phase 6: Risk Assessment Service Verification (Days 13-14)

### 6.1 Verify HandleGetRiskAssessment

**Files**: `services/risk-assessment-service/internal/handlers/risk_assessment.go` (lines 171-227)

**Tasks**:

- [ ] Verify HandleGetRiskAssessment implementation (lines 171-227)
- [ ] Test with valid assessment ID
- [ ] Test with non-existent ID (should return 404)
- [ ] Test with invalid ID format (should return 400)
- [ ] Verify database query works correctly
- [ ] Verify error handling uses standardized error helper

**Testing**:

```bash
# Create an assessment first (via POST /api/v1/assess)
ASSESSMENT_ID="..."  # From create response

# Test GET with valid ID
curl http://localhost:8080/api/v1/assess/$ASSESSMENT_ID

# Test GET with non-existent ID
curl http://localhost:8080/api/v1/assess/nonexistent-id

# Test GET with invalid format
curl http://localhost:8080/api/v1/assess/invalid-format
```

**Success Criteria**:

- Valid ID returns assessment data
- Non-existent ID returns 404
- Invalid format returns 400
- Error responses use standardized format
- Database queries work correctly

---

### 6.2 Verify Database Query Logic in HandleRiskPrediction

**Files**: `services/risk-assessment-service/internal/handlers/risk_assessment.go`

**Tasks**:

- [ ] Review HandleRiskPrediction implementation
- [ ] Verify database query logic for assessment ID
- [ ] Test prediction with valid assessment ID (uses database data)
- [ ] Test prediction with non-existent ID (uses fallback)
- [ ] Test prediction without ID (uses fallback)

**Testing**:

```bash
# Test with assessment ID
curl -X POST http://localhost:8080/api/v1/assess/predict/{id} \
  -H "Content-Type: application/json" \
  -d '{}'

# Test without ID (fallback)
curl -X POST http://localhost:8080/api/v1/assess/predict \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test", "industry": "Retail"}'
```

**Success Criteria**:

- Database queries work when ID provided
- Fallback to mock data when ID not found
- Error handling is consistent

---

## Phase 7: Comprehensive Testing & Validation (Days 15-16)

### 7.1 Integration Testing

**Tasks**:

- [ ] Test complete admin workflow (create threshold → export → import → verify)
- [ ] Test risk assessment with database persistence
- [ ] Test system health monitoring
- [ ] Test error handling across all endpoints
- [ ] Test with database unavailable (graceful degradation)
- [ ] Test with Redis unavailable (graceful degradation)

**Test Scenarios**:

1. **Complete Threshold Workflow**:

   - Create threshold
   - Get threshold
   - Update threshold
   - Export thresholds
   - Import thresholds
   - Delete threshold
   - Restart server, verify persistence

2. **Risk Assessment Workflow**:

   - Create assessment
   - Get assessment by ID
   - Verify database persistence
   - Test prediction with assessment ID

3. **Error Scenarios**:

   - Invalid JSON
   - Missing required fields
   - Non-existent IDs
   - Database errors
   - Service unavailable

---

### 7.2 Performance Testing

**Tasks**:

- [ ] Test threshold CRUD operations performance
- [ ] Test export/import with large datasets
- [ ] Test concurrent requests
- [ ] Verify Redis caching improves performance
- [ ] Monitor database connection pool usage

**Testing**:

```bash
# Performance test script
for i in {1..100}; do
  curl -X POST http://localhost:8080/v1/admin/risk/thresholds \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"Test$i\",\"category\":\"financial\",\"risk_levels\":{\"low\":25}}"
done

# Measure export time
time curl http://localhost:8080/v1/admin/risk/threshold-export > /dev/null
```

---

### 7.3 Documentation Updates

**Tasks**:

- [ ] Update API documentation with all restored endpoints
- [ ] Document error response formats
- [ ] Document database persistence behavior
- [ ] Document graceful degradation scenarios
- [ ] Update migration documents with completion status

---

## Phase 8: Final Verification & Sign-off (Day 17)

### 8.1 Final Checklist

- [ ] All 15 handlers verified and working
- [ ] Database persistence working for thresholds
- [ ] Export/import functionality works correctly
- [ ] Error handling is consistent across all services
- [ ] All tests passing
- [ ] No regression in existing functionality
- [ ] Documentation updated
- [ ] Performance is acceptable
- [ ] Graceful degradation works (database/Redis unavailable)

### 8.2 Success Criteria

- ✅ All endpoints return correct status codes
- ✅ Database persistence works
- ✅ Error messages are user-friendly
- ✅ Logging includes request IDs
- ✅ Health checks work correctly
- ✅ No enhanced functionality lost
- ✅ All handlers tested and verified

---

## Testing Strategy

### Unit Tests

- Test all restored handlers individually
- Test helper functions (getRequestID, extractIDFromPath)
- Test error cases
- Test validation logic

### Integration Tests

- Test database persistence
- Test threshold CRUD operations
- Test export/import round-trip
- Test recommendation rules integration
- Test notification channels integration

### End-to-End Tests

- Test complete admin workflow
- Test risk assessment with database persistence
- Test system health monitoring
- Test error handling across all endpoints

### Manual Testing

- Verify all endpoints return correct status codes
- Verify database persistence works
- Verify error messages are user-friendly
- Verify logging includes request IDs
- Verify health checks work correctly

---

## Risk Mitigation

### High-Risk Areas

1. **Database Schema Changes**: Ensure migrations are backward compatible
2. **Service Dependencies**: Verify all services are initialized correctly
3. **Error Handling**: Ensure consistent error responses
4. **Route Conflicts**: Verify no route conflicts with existing routes

### Mitigation Strategies

1. Test in staging environment first
2. Use feature flags for gradual rollout
3. Monitor error rates and logs
4. Have rollback plan ready

---

## Notes

- Most handlers are already implemented - focus is on verification and testing
- Ensure no enhanced functionality is lost during verification
- Test thoroughly after each phase before proceeding
- Keep detailed test results and logs
- Document any issues found and resolutions