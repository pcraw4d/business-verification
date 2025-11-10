# Database Query Optimization Analysis

**Date**: 2025-11-10  
**Status**: Complete

---

## Summary

Analysis of database query patterns, N+1 query problems, missing indexes, and optimization opportunities across all services.

---

## Database Query Statistics

### SQL Query Patterns by Service

**API Gateway:**
- SQL-related matches: 25
- Database operations: Supabase client usage
- Query patterns: Health checks, table counts

**Classification Service:**
- SQL-related matches: 20
- Database operations: Supabase client usage
- Query patterns: Classification storage, code lookups

**Merchant Service:**
- SQL-related matches: 531
- Database operations: Extensive Supabase usage
- Query patterns: CRUD operations, search, analytics
- Potential N+1 queries: Found 2 for loops with database operations
  - Line 760: `for _, row := range result` (in merchant.go)
  - Line 800: `for i, part := range parts` (in merchant.go)

**Risk Assessment Service:**
- SQL-related matches: 6,116
- Database operations: Extensive database usage
- Query patterns: Risk assessments, predictions, reports, batch jobs
- Optimization needed: High priority

---

## N+1 Query Analysis

### Potential N+1 Query Problems

**Merchant Service:**
- ⚠️ Potential N+1 queries in handlers
- ⚠️ Need to verify for loops with database calls
- ⚠️ Review query patterns in merchant.go

**Risk Assessment Service:**
- ⚠️ Extensive database usage (6,116 SQL-related matches)
- ⚠️ High likelihood of N+1 queries
- ⚠️ Need thorough review

---

## Index Analysis

### Existing Indexes

**Index Definitions Found:**
- Migration files contain index definitions
- Risk Assessment Service has index management (`internal/database/index_manager.go`)
- Query optimization scripts exist
- 10 files with CREATE INDEX statements found

**Index Patterns:**
- Time-based indexes (created_at)
- Foreign key indexes (business_id, user_id)
- Composite indexes for common queries
- Partial indexes for active records
- Performance indexes for risk assessments
- Batch job indexes for status and creation time

---

## Query Optimization Opportunities

### High Priority

1. **Review N+1 Query Problems**
   - Audit Merchant Service handlers
   - Audit Risk Assessment Service queries
   - Implement batch loading where needed
   - Use JOINs instead of multiple queries

2. **Add Missing Indexes**
   - Review frequently queried columns
   - Add indexes for foreign keys
   - Add composite indexes for common query patterns
   - Add partial indexes for filtered queries

3. **Optimize Risk Assessment Service Queries**
   - Review 6,116 SQL-related matches
   - Identify slow queries
   - Optimize query patterns
   - Add appropriate indexes

### Medium Priority

4. **Query Performance Monitoring**
   - Enable slow query logging
   - Monitor query execution times
   - Track query patterns
   - Alert on slow queries

5. **Connection Pooling**
   - Review connection pool sizes
   - Optimize pool configuration
   - Monitor connection usage
   - Prevent connection exhaustion

---

## Database Query Patterns

### Common Query Patterns

**Time-based Queries:**
- `WHERE created_at BETWEEN ...`
- `ORDER BY created_at DESC`
- Need indexes on `created_at`

**Foreign Key Queries:**
- `WHERE business_id = ...`
- `WHERE user_id = ...`
- Need indexes on foreign keys

**Status-based Queries:**
- `WHERE status = ...`
- `WHERE status IN (...)`
- Need indexes on `status`

**Composite Queries:**
- `WHERE business_id = ... AND status = ...`
- `WHERE created_at BETWEEN ... AND status = ...`
- Need composite indexes

---

## Recommendations

### High Priority

1. **Audit N+1 Queries**
   - Review all for loops with database calls
   - Implement batch loading
   - Use JOINs for related data

2. **Add Missing Indexes**
   - Review query patterns
   - Add indexes for frequently queried columns
   - Add composite indexes for common patterns

3. **Optimize Risk Assessment Service**
   - Review extensive database usage
   - Identify optimization opportunities
   - Implement query optimizations

### Medium Priority

4. **Query Performance Monitoring**
   - Enable slow query logging
   - Monitor query execution times
   - Track query patterns

5. **Connection Pool Optimization**
   - Review pool sizes
   - Optimize configuration
   - Monitor usage

---

## Action Items

1. **Review Query Patterns**
   - Audit all database queries
   - Identify N+1 problems
   - Document query patterns

2. **Add Indexes**
   - Review frequently queried columns
   - Add missing indexes
   - Test query performance

3. **Optimize Queries**
   - Refactor N+1 queries
   - Use batch loading
   - Optimize slow queries

---

**Last Updated**: 2025-11-10 03:45 UTC

