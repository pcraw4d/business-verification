# Supabase Test Database Review

**Date:** January 2025  
**Purpose:** Review existing Supabase setup to determine if test database/tables exist before creating new ones

---

## Current Supabase Configuration

### Production Supabase Project

**Project URL:** `https://qpqhuqqmkjxsltzshfam.supabase.co`  
**Database Host:** `db.qpqhuqqmkjxsltzshfam.supabase.co`  
**Database:** `postgres` (default Supabase database)

**Configuration Files:**
- `railway.env` - Contains production Supabase configuration
- `configs/development.env` - Development Supabase configuration
- `configs/test.env` - Test environment configuration (placeholder values)

---

## Tables Expected Based on Migrations

Based on the migration files and schema definitions, the following tables should exist for testing:

### Core Tables (from Migration 005)

1. **`merchant_analytics`**
   - Purpose: Stores calculated analytics data for merchants
   - Columns: `id`, `merchant_id`, `classification_data`, `security_data`, `quality_data`, `intelligence_data`, `created_at`, `updated_at`
   - Migration: `internal/database/migrations/005_merchant_portfolio_schema.sql`

2. **`merchants`**
   - Purpose: Main merchants table for portfolio management
   - Migration: `internal/database/migrations/005_merchant_portfolio_schema.sql`

### Risk Assessment Tables (from Migration 010)

3. **`risk_assessments`**
   - Purpose: Risk assessment records with async processing support
   - Columns: `id`, `merchant_id`, `status`, `options`, `result`, `progress`, `overall_score`, `risk_level`, `created_at`, `updated_at`
   - Migration: `internal/database/migrations/010_add_async_risk_assessment_columns.sql`

### Risk Indicators Table

4. **`risk_indicators`**
   - Purpose: Active risk indicators for merchants
   - Columns: `id`, `merchant_id`, `type`, `name`, `severity`, `status`, `description`, `detected_at`, `score`
   - Referenced in: `internal/database/risk_indicators_repository.go`

### Enrichment Tables

5. **`enrichment_jobs`**
   - Purpose: Tracks data enrichment jobs
   - Referenced in: `test/integration/database_setup.go`

6. **`enrichment_sources`**
   - Purpose: Tracks enrichment data sources
   - Referenced in: `test/integration/database_setup.go`

---

## Test Database Configuration

### Current Test Database Setup

**File:** `test/integration/test_config.go`

**Default Configuration:**
```go
DatabaseURL: "postgres://postgres:password@localhost:5432/kyb_test?sslmode=disable"
```

**Environment Variable Priority:**
1. `TEST_DATABASE_URL` (highest priority)
2. `SUPABASE_URL` + `SUPABASE_SERVICE_ROLE_KEY` (constructs connection string)
3. Default local PostgreSQL

### Test Database Helper

**File:** `test/integration/database_setup.go`

**Functions:**
- `SetupTestDatabase()` - Initializes test database connection
- `CleanupTestDatabase()` - Closes database connection
- `ResetTestDatabase(ctx)` - Truncates all test tables
- `SeedTestData(ctx)` - Inserts test data fixtures
- `VerifyTestDatabase()` - Checks if database is accessible

---

## Verification Steps

### Step 1: Check if Production Database Has Required Tables

**Option A: Using Supabase Dashboard**

1. Go to: https://supabase.com/dashboard/project/qpqhuqqmkjxsltzshfam
2. Navigate to **SQL Editor**
3. Run this query to check for required tables:

```sql
SELECT 
    table_name,
    table_schema
FROM information_schema.tables
WHERE table_schema = 'public'
    AND table_name IN (
        'merchant_analytics',
        'merchants',
        'risk_assessments',
        'risk_indicators',
        'enrichment_jobs',
        'enrichment_sources'
    )
ORDER BY table_name;
```

**Option B: Using Database Connection**

```bash
# Set environment variables
export SUPABASE_URL="https://qpqhuqqmkjxsltzshfam.supabase.co"
export SUPABASE_SERVICE_ROLE_KEY="your-service-role-key"

# Run verification
go run -tags integration ./test/integration/verify_tables.go
```

### Step 2: Check Table Schemas

If tables exist, verify their structure matches expected schema:

```sql
-- Check merchant_analytics table structure
SELECT 
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns
WHERE table_schema = 'public'
    AND table_name = 'merchant_analytics'
ORDER BY ordinal_position;

-- Check risk_assessments table structure
SELECT 
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns
WHERE table_schema = 'public'
    AND table_name = 'risk_assessments'
ORDER BY ordinal_position;

-- Check risk_indicators table structure
SELECT 
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns
WHERE table_schema = 'public'
    AND table_name = 'risk_indicators'
ORDER BY ordinal_position;
```

---

## Recommendations

### Option 1: Use Production Database for Testing (NOT RECOMMENDED)

**Risks:**
- Could corrupt production data
- Test data mixed with production data
- Performance impact on production
- Security concerns

**Only use if:**
- You have a separate schema for testing
- You can safely isolate test data
- You have proper backup/restore procedures

### Option 2: Create Separate Supabase Test Project (RECOMMENDED)

**Steps:**

1. **Create New Supabase Project:**
   - Go to https://supabase.com/dashboard
   - Click "New Project"
   - Name: `kyb-platform-test` or `kyb-platform-dev-test`
   - Region: Same as production (for consistency)
   - Database Password: Generate strong password

2. **Get Connection Details:**
   - Project URL: `https://[project-ref].supabase.co`
   - Database Host: `db.[project-ref].supabase.co`
   - Service Role Key: From Project Settings > API

3. **Set Environment Variables:**
   ```bash
   export TEST_DATABASE_URL="postgresql://postgres:[PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres?sslmode=require"
   export SUPABASE_TEST_URL="https://[project-ref].supabase.co"
   export SUPABASE_TEST_SERVICE_ROLE_KEY="[service-role-key]"
   ```

4. **Run Migrations:**
   - Execute migration files in Supabase SQL Editor
   - Start with `internal/database/migrations/005_merchant_portfolio_schema.sql`
   - Then `internal/database/migrations/010_add_async_risk_assessment_columns.sql`
   - Create `risk_indicators` and enrichment tables if needed

### Option 3: Use Local Docker PostgreSQL (Alternative)

**For Local Development Only:**

1. **Create `docker-compose.test.yml`:**
   ```yaml
   version: '3.8'
   services:
     test-db:
       image: postgres:15-alpine
       container_name: kyb-test-db
       environment:
         POSTGRES_USER: postgres
         POSTGRES_PASSWORD: password
         POSTGRES_DB: kyb_test
       ports:
         - "5433:5432"
   ```

2. **Start Database:**
   ```bash
   docker-compose -f docker-compose.test.yml up -d
   ```

3. **Set Environment Variable:**
   ```bash
   export TEST_DATABASE_URL="postgres://postgres:password@localhost:5433/kyb_test?sslmode=disable"
   ```

4. **Run Migrations:**
   ```bash
   psql $TEST_DATABASE_URL -f internal/database/migrations/005_merchant_portfolio_schema.sql
   psql $TEST_DATABASE_URL -f internal/database/migrations/010_add_async_risk_assessment_columns.sql
   ```

---

## Next Steps

### Immediate Actions

1. **Verify Production Tables:**
   - Run the SQL queries above in Supabase SQL Editor
   - Document which tables exist and their structure
   - Note any missing tables or schema mismatches

2. **Decide on Test Database Strategy:**
   - **Recommended:** Create separate Supabase test project
   - **Alternative:** Use local Docker PostgreSQL for development
   - **Not Recommended:** Use production database

3. **Create Test Database (if needed):**
   - Follow Option 2 or Option 3 above
   - Run all required migrations
   - Verify table creation

4. **Update Test Configuration:**
   - Set `TEST_DATABASE_URL` environment variable
   - Update `test/integration/database_setup.go` if needed
   - Test database connection

### Verification Script

Create a simple verification script to check table existence:

```go
// test/integration/verify_tables.go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "os"
    
    _ "github.com/lib/pq"
    "kyb-platform/test/integration"
)

func main() {
    testDB, err := integration.SetupTestDatabase()
    if err != nil {
        log.Fatalf("Failed to setup test database: %v", err)
    }
    defer testDB.CleanupTestDatabase()
    
    db := testDB.GetDB()
    ctx := context.Background()
    
    requiredTables := []string{
        "merchant_analytics",
        "merchants",
        "risk_assessments",
        "risk_indicators",
        "enrichment_jobs",
        "enrichment_sources",
    }
    
    for _, table := range requiredTables {
        var exists bool
        query := `
            SELECT EXISTS (
                SELECT FROM information_schema.tables 
                WHERE table_schema = 'public' 
                AND table_name = $1
            )
        `
        err := db.QueryRowContext(ctx, query, table).Scan(&exists)
        if err != nil {
            log.Printf("Error checking table %s: %v", table, err)
            continue
        }
        
        if exists {
            fmt.Printf("✅ Table '%s' exists\n", table)
        } else {
            fmt.Printf("❌ Table '%s' does NOT exist\n", table)
        }
    }
}
```

---

## Summary

### Current Status

- ✅ **Production Supabase Project:** Configured at `qpqhuqqmkjxsltzshfam.supabase.co`
- ⚠️ **Test Database:** Not explicitly configured (defaults to local `kyb_test`)
- ⚠️ **Table Status:** Unknown - needs verification
- ✅ **Test Infrastructure:** Helper functions created in `test/integration/database_setup.go`

### Required Actions

1. **Verify Table Existence:** Run SQL queries in Supabase to check if tables exist
2. **Create Test Database:** Either separate Supabase project or local Docker
3. **Run Migrations:** Execute migration files to create required tables
4. **Configure Test Environment:** Set `TEST_DATABASE_URL` environment variable
5. **Verify Connection:** Test database connection using helper functions

### Recommended Approach

**Create a separate Supabase test project** to:
- Isolate test data from production
- Avoid performance impact on production
- Enable safe test data cleanup
- Maintain production data integrity

---

**Next Step:** Run the verification queries in Supabase SQL Editor to determine current table status, then proceed with test database setup based on the results.

