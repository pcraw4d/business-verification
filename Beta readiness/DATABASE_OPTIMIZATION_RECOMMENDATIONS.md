# Database Optimization Recommendations

**Date**: 2025-01-27  
**Status**: Analysis Complete - Recommendations Provided

## Executive Summary

Analysis of database query patterns across all services reveals good practices with Supabase PostgREST client usage. Most queries use parameterized queries and batch operations. Minor optimizations recommended for connection pooling and query batching.

---

## Current State Analysis

### ✅ Good Practices Found

1. **Parameterized Queries**: All services use Supabase PostgREST client which handles parameterization automatically
2. **Connection Pooling**: Risk Assessment Service has connection pooling implementation (`internal/pool/connection_pool.go`)
3. **Query Optimization**: Risk Assessment Service has query optimizer with caching (`internal/query/optimizer.go`)
4. **Batch Operations**: Query optimizer supports batch execution

### ⚠️ Areas for Improvement

1. **Connection Pooling**: Not consistently used across all services
2. **Query Batching**: Some loops could benefit from batch queries
3. **Index Usage**: Need to verify all frequently queried columns have indexes

---

## N+1 Query Analysis

### Merchant Service

**Potential N+1 Pattern Found:**
- Line 942: `for _, row := range result` - This is safe, iterating over already-fetched results
- Line 988: `for i, part := range parts` - This is safe, processing JWT token parts

**Verdict**: ✅ **No N+1 queries found** - All database queries use batch operations

### Risk Assessment Service

**Query Patterns:**
- Uses query optimizer with caching
- Batch execution support available
- Connection pooling implemented

**Verdict**: ✅ **Well optimized** - Has proper query optimization infrastructure

### Classification Service

**Query Patterns:**
- Uses Supabase client for all queries
- Simple queries, no loops with database calls

**Verdict**: ✅ **No issues found**

---

## Connection Pooling Recommendations

### Current State
- ✅ Risk Assessment Service: Has connection pooling
- ⚠️ Merchant Service: Uses Supabase client (connection pooling handled by client)
- ⚠️ Classification Service: Uses Supabase client (connection pooling handled by client)
- ⚠️ API Gateway: Uses Supabase client (connection pooling handled by client)

### Recommendation
**Status**: ✅ **No action needed**

Supabase PostgREST client handles connection pooling internally. The Risk Assessment Service has additional connection pooling for direct SQL queries, which is appropriate.

---

## Index Recommendations

### Recommended Indexes

Based on query patterns, ensure these indexes exist:

1. **Merchants Table:**
   - `idx_merchants_created_at` on `created_at` (for sorting)
   - `idx_merchants_portfolio_type` on `portfolio_type` (for filtering)
   - `idx_merchants_risk_level` on `risk_level` (for filtering)
   - `idx_merchants_status` on `status` (for filtering)
   - `idx_merchants_name` on `name` (for search)
   - Composite: `idx_merchants_portfolio_risk` on `(portfolio_type, risk_level)`

2. **Risk Assessments Table:**
   - `idx_risk_assessments_business_id` on `business_id` (foreign key)
   - `idx_risk_assessments_created_at` on `created_at` (for sorting)
   - `idx_risk_assessments_risk_score` on `risk_score` (for filtering)
   - Composite: `idx_risk_assessments_business_created` on `(business_id, created_at)`

3. **Classifications Table:**
   - `idx_classifications_business_name` on `business_name` (for lookups)
   - `idx_classifications_created_at` on `created_at` (for sorting)

### Verification
Run this SQL to check existing indexes:

```sql
SELECT 
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public'
ORDER BY tablename, indexname;
```

---

## Query Optimization Recommendations

### 1. Use Batch Queries Where Possible

**Current**: Individual queries in loops (if any)
**Recommended**: Batch queries using Supabase client's batch operations

**Example:**
```go
// Instead of:
for _, id := range ids {
    result, err := client.From("table").Select("*").Eq("id", id).Execute()
}

// Use:
var results []map[string]interface{}
query := client.From("table").Select("*").In("id", ids).ExecuteTo(&results)
```

### 2. Use Query Optimizer for Complex Queries

**Risk Assessment Service** already has this. Consider using similar patterns in other services for:
- Query result caching
- Slow query detection
- Query profiling

### 3. Implement Query Timeouts

**Current**: Some queries may not have timeouts
**Recommended**: Always use context with timeout

**Example:**
```go
ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
defer cancel()

result, err := client.From("table").Select("*").ExecuteWithContext(ctx)
```

---

## Performance Monitoring

### Current State
- ✅ Risk Assessment Service: Has query profiler and slow query detection
- ⚠️ Other Services: Basic logging only

### Recommendations

1. **Enable Slow Query Logging** in all services
   - Threshold: 1 second
   - Log query, duration, and parameters

2. **Add Query Metrics**:
   - Query count per endpoint
   - Average query time
   - Slow query count
   - Cache hit rate (where applicable)

3. **Use Prometheus Metrics**:
   - `db_query_duration_seconds` - Query duration histogram
   - `db_query_total` - Total query count
   - `db_slow_queries_total` - Slow query count

---

## Caching Recommendations

### Current State
- ✅ Classification Service: In-memory cache implemented
- ✅ Risk Assessment Service: Query result caching available
- ⚠️ Merchant Service: No caching (could benefit from Redis)

### Recommendations

1. **Merchant Service**: Consider adding Redis cache for:
   - Merchant lookups by ID
   - Merchant lists (with pagination keys)
   - TTL: 5 minutes for reads, invalidate on writes

2. **API Gateway**: Consider caching:
   - Health check responses (already optimized)
   - Service discovery responses

---

## Summary

### Critical Issues
- ✅ None found

### High Priority Recommendations
1. ✅ Verify indexes exist for frequently queried columns
2. ✅ Enable slow query logging in all services
3. ⚠️ Consider Redis caching for Merchant Service

### Medium Priority Recommendations
1. Add query metrics to all services
2. Use query optimizer patterns in other services
3. Implement query timeouts consistently

### Low Priority (Post-Beta)
1. Database query profiling dashboard
2. Automated index recommendations
3. Query plan analysis

---

## Action Items

### Pre-Beta (Recommended)
- [ ] Verify database indexes exist (run index check SQL)
- [ ] Enable slow query logging in all services
- [ ] Add query timeouts to all database operations

### Post-Beta (Optional)
- [ ] Implement Redis caching for Merchant Service
- [ ] Add comprehensive query metrics
- [ ] Create query performance dashboard

---

## Conclusion

**Status**: ✅ **Database queries are well-optimized**

The codebase demonstrates good database practices:
- Parameterized queries throughout
- Connection pooling where needed
- Query optimization infrastructure in place
- No N+1 query problems found

**Recommendation**: Verify indexes and enable slow query logging before beta launch. Other optimizations can be done post-beta.

