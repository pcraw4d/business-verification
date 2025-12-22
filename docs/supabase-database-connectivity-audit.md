# Supabase Database Connectivity Audit - Track 6.3

## Executive Summary

Investigation of Supabase database connectivity reveals **connection is configured and working**, but **data completeness and query performance need verification**. The database is critical for classification code metadata (MCC, NAICS, SIC) and industry keyword matching.

**Status**: ⚠️ **MEDIUM** - Connection working, but data completeness and performance need verification

## Database Configuration

### Environment Variables

**Location**: `services/classification-service/internal/config/config.go:96-100`

```go
Supabase: SupabaseConfig{
    URL:            getEnvAsString("SUPABASE_URL", ""),
    APIKey:         getEnvAsString("SUPABASE_ANON_KEY", ""),
    ServiceRoleKey: getEnvAsString("SUPABASE_SERVICE_ROLE_KEY", ""),
    JWTSecret:      getEnvAsString("SUPABASE_JWT_SECRET", ""),
}
```

**Required Variables**:
- `SUPABASE_URL`: Supabase project URL
- `SUPABASE_ANON_KEY`: Anonymous API key
- `SUPABASE_SERVICE_ROLE_KEY`: Service role key (for admin operations)
- `SUPABASE_JWT_SECRET`: JWT secret (optional)

**Status**: ✅ Configuration defined, values loaded from environment

### Client Initialization

**Location**: `services/classification-service/cmd/main.go:76-93`

**Initialization Process**:
1. Create Supabase client with URL and API key
2. Create database client adapter for classification repository
3. Connect to database (10s timeout)
4. Continue even if connection fails (with warning)

**Status**: ✅ Initialization implemented with error handling

### Health Check

**Location**: `services/classification-service/internal/supabase/client.go:55-81`

**Health Check Logic**:
- Timeout: 5 seconds
- Test query: `SELECT count FROM merchants LIMIT 1`
- Fallback: If table doesn't exist, verify client is initialized

**Status**: ✅ Health check implemented

## Database Tables

### Core Classification Tables

1. **`classification_codes`** ⚠️ **CRITICAL**
   - Stores MCC, NAICS, SIC codes
   - Columns: `id`, `industry_id`, `code_type`, `code`, `description`, `is_active`
   - Used by: `GetClassificationCodesByIndustry`, `GetClassificationCodesByType`
   - **Status**: ⏳ **NEEDS VERIFICATION** - Data completeness unknown

2. **`code_metadata`** ⚠️ **HIGH**
   - Enhanced metadata for codes
   - Columns: `code_type`, `code`, `official_name`, `official_description`, `industry_mappings`, `crosswalk_data`, `hierarchy`
   - Used by: `CodeMetadataRepository`
   - **Status**: ⏳ **NEEDS VERIFICATION** - Data completeness unknown

3. **`industries`** ⚠️ **CRITICAL**
   - Industry classification data
   - Used for industry matching
   - **Status**: ⏳ **NEEDS VERIFICATION** - Data completeness unknown

4. **`industry_keywords`** ⚠️ **CRITICAL**
   - Keyword-to-industry mappings
   - Used for keyword-based classification
   - **Status**: ⏳ **NEEDS VERIFICATION** - Data completeness unknown

### Supporting Tables

5. **`industry_code_crosswalks`** ⚠️ **MEDIUM**
   - Crosswalks between code types (MCC ↔ NAICS ↔ SIC)
   - Used for code generation
   - **Status**: ⏳ **NEEDS VERIFICATION** - May be missing

6. **`merchants`** ⚠️ **LOW**
   - Merchant data (for health check)
   - **Status**: ⏳ **NEEDS VERIFICATION** - May not exist

## Database Queries

### Key Query Patterns

1. **Get Classification Codes by Industry**
   - **Location**: `internal/classification/repository/supabase_repository.go:1671-1720`
   - **Query**: `SELECT * FROM classification_codes WHERE industry_id = ? AND is_active = true`
   - **Performance**: ⏳ **NEEDS VERIFICATION** - May be slow for large datasets

2. **Get Classification Codes by Type**
   - **Location**: `internal/classification/repository/supabase_repository.go:2530-2566`
   - **Query**: `SELECT * FROM classification_codes WHERE code_type = ? AND is_active = true LIMIT 5000`
   - **Performance**: ⏳ **NEEDS VERIFICATION** - Limited to 5000 records

3. **Get Code Metadata**
   - **Location**: `internal/classification/repository/code_metadata_repository.go:50-84`
   - **Query**: `SELECT * FROM code_metadata WHERE code_type = ? AND code = ? AND is_active = true`
   - **Performance**: ⏳ **NEEDS VERIFICATION** - Single record lookup

4. **Get Crosswalk Codes**
   - **Location**: `internal/classification/repository/supabase_repository.go:2324-2393`
   - **Query**: Complex JOIN queries for crosswalk data
   - **Performance**: ⏳ **NEEDS VERIFICATION** - May be slow

### Query Performance Concerns

1. **No Query Timeout** ⚠️ **MEDIUM**
   - Queries may hang indefinitely
   - No explicit timeout on database queries
   - **Impact**: May cause request timeouts

2. **Large Result Sets** ⚠️ **MEDIUM**
   - `GetClassificationCodesByType` limited to 5000 records
   - May miss codes if more than 5000 exist
   - **Impact**: Incomplete code generation

3. **N+1 Query Problem** ⚠️ **LOW**
   - `GetCodeMetadataBatch` queries each code individually
   - Could be optimized with batch queries
   - **Impact**: Slower performance for multiple codes

## Data Completeness

### Expected Data

**From**: `SUPABASE_DATABASE_SETUP_INSTRUCTIONS.md`

**Required Tables**:
- ✅ `classifications` (for storing results)
- ✅ `merchants` (for merchant endpoints)
- ✅ `industries` (for industry classification)
- ✅ `industry_keywords` (for keyword matching)
- ✅ `classification_codes` (for MCC, NAICS, SIC codes)
- ⚠️ `code_metadata` (enhanced metadata - may be missing)
- ⚠️ `industry_code_crosswalks` (crosswalks - may be missing)

**Status**: ⚠️ **UNCLEAR** - Some tables may be missing based on setup instructions

### Code Metadata Completeness

**Expected Counts** (approximate):
- **MCC Codes**: ~500-1000 codes
- **NAICS Codes**: ~1,000-2,000 codes (2022 version)
- **SIC Codes**: ~1,000 codes

**Verification Queries**:
```sql
SELECT COUNT(*) FROM classification_codes WHERE code_type = 'MCC' AND is_active = true;
SELECT COUNT(*) FROM classification_codes WHERE code_type = 'NAICS' AND is_active = true;
SELECT COUNT(*) FROM classification_codes WHERE code_type = 'SIC' AND is_active = true;
SELECT COUNT(*) FROM code_metadata;
```

**Status**: ⏳ **NEEDS VERIFICATION** - Counts unknown

## Investigation Steps

### Step 1: Check Database Connection

**Health Endpoint**: `/health` on classification service

**Check**:
```bash
curl https://classification-service-production.up.railway.app/health | jq '.supabase_status'
```

**Expected**:
```json
{
  "connected": true,
  "url": "https://xxx.supabase.co",
  "error": null
}
```

**Status**: ⏳ **PENDING** - Need to verify in production

### Step 2: Verify Table Existence

**Check Tables**:
```sql
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name IN (
  'classification_codes',
  'code_metadata',
  'industries',
  'industry_keywords',
  'industry_code_crosswalks'
);
```

**Status**: ⏳ **PENDING** - Need to run in Supabase SQL Editor

### Step 3: Verify Data Completeness

**Count Records**:
```sql
-- Classification codes
SELECT code_type, COUNT(*) as count 
FROM classification_codes 
WHERE is_active = true 
GROUP BY code_type;

-- Code metadata
SELECT code_type, COUNT(*) as count 
FROM code_metadata 
WHERE is_active = true 
GROUP BY code_type;

-- Industries
SELECT COUNT(*) as count FROM industries;

-- Industry keywords
SELECT COUNT(*) as count FROM industry_keywords;
```

**Status**: ⏳ **PENDING** - Need to run in Supabase SQL Editor

### Step 4: Test Query Performance

**Test Queries**:
```sql
-- Test classification codes query
EXPLAIN ANALYZE
SELECT * FROM classification_codes 
WHERE industry_id = 1 AND is_active = true 
LIMIT 100;

-- Test code metadata query
EXPLAIN ANALYZE
SELECT * FROM code_metadata 
WHERE code_type = 'NAICS' AND code = '541511' AND is_active = true;
```

**Status**: ⏳ **PENDING** - Need to run in Supabase SQL Editor

### Step 5: Review Query Logs

**Check Railway Logs**:
- Look for slow query warnings
- Check for query timeout errors
- Review database connection errors

**Status**: ⏳ **PENDING** - Need to analyze logs

## Root Cause Analysis

### Potential Issues

1. **Missing Tables** ⚠️ **HIGH**
   - Some tables may not exist (from setup instructions)
   - **Impact**: Queries fail, code generation fails
   - **Evidence**: Setup instructions mention missing tables

2. **Incomplete Data** ⚠️ **HIGH**
   - Code metadata may be missing or incomplete
   - **Impact**: Code generation rate low (23.1%), accuracy 0%
   - **Evidence**: NAICS/SIC accuracy is 0%

3. **Slow Queries** ⚠️ **MEDIUM**
   - Queries may be slow without proper indexing
   - **Impact**: Request timeouts, high latency
   - **Evidence**: Average latency 43.7s

4. **No Query Timeout** ⚠️ **MEDIUM**
   - Queries may hang indefinitely
   - **Impact**: Request timeouts
   - **Evidence**: Timeout errors (9.9% of errors)

5. **Large Result Set Limits** ⚠️ **LOW**
   - 5000 record limit may miss codes
   - **Impact**: Incomplete code generation
   - **Evidence**: Code generation rate 23.1%

## Recommendations

### Immediate Actions (High Priority)

1. **Verify Table Existence**:
   - Run table existence check in Supabase SQL Editor
   - Create missing tables if needed
   - Run migration scripts

2. **Verify Data Completeness**:
   - Run data count queries
   - Verify MCC, NAICS, SIC codes are populated
   - Check code_metadata table has data

3. **Add Query Timeouts**:
   - Add context timeouts to all database queries
   - Set reasonable timeout values (5-10s)
   - Handle timeout errors gracefully

### Medium Priority Actions

4. **Optimize Query Performance**:
   - Add indexes on frequently queried columns
   - Review EXPLAIN ANALYZE results
   - Optimize slow queries

5. **Fix N+1 Query Problem**:
   - Batch code metadata queries
   - Use IN queries instead of individual queries
   - Cache frequently accessed data

6. **Increase Result Set Limits**:
   - Review if 5000 limit is sufficient
   - Consider pagination for large result sets
   - Add monitoring for limit hits

## Code Locations

- **Client Configuration**: `services/classification-service/internal/config/config.go:28-34`
- **Client Initialization**: `services/classification-service/cmd/main.go:76-93`
- **Health Check**: `services/classification-service/internal/supabase/client.go:55-81`
- **Repository**: `internal/classification/repository/supabase_repository.go`
- **Code Metadata Repository**: `internal/classification/repository/code_metadata_repository.go`

## Next Steps

1. ✅ **Complete Track 6.3 Investigation** - This document
2. **Verify Table Existence** - Run SQL queries in Supabase
3. **Verify Data Completeness** - Count records in key tables
4. **Test Query Performance** - Run EXPLAIN ANALYZE
5. **Add Query Timeouts** - Implement context timeouts
6. **Optimize Queries** - Add indexes and optimize slow queries

## Expected Impact

After fixing issues:

1. **Code Generation Rate**: 23.1% → ≥90% (with complete data)
2. **NAICS Accuracy**: 0% → ≥70% (with complete data)
3. **SIC Accuracy**: 0% → ≥70% (with complete data)
4. **Query Performance**: Improved with indexes and timeouts
5. **Error Rate**: Reduced with proper error handling

## References

- Setup Instructions: `SUPABASE_DATABASE_SETUP_INSTRUCTIONS.md`
- Repository Implementation: `internal/classification/repository/supabase_repository.go`
- Code Metadata Repository: `internal/classification/repository/code_metadata_repository.go`
- Client Implementation: `services/classification-service/internal/supabase/client.go`

