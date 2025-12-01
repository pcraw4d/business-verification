# Migration Verification Guide

## Overview

This guide helps you verify that migrations 028 and 029 were applied successfully in Supabase.

## Understanding "Success. No rows returned"

When you run DDL (Data Definition Language) statements like `CREATE TABLE`, `CREATE INDEX`, etc., Supabase will return:
- **Status**: `Success`
- **Rows returned**: `0` (or "no rows returned")

This is **expected and correct**! DDL statements don't return data rows - they modify the database structure.

## Verification Steps

### Step 1: Run the Verification Script

Run the verification script in Supabase SQL Editor:

```sql
-- Copy and paste the entire contents of:
-- scripts/verify_migrations_028_029.sql
```

### Step 2: Check the Results

The verification script will check:

#### Migration 028 (Enhanced Classification Schema)
- ✅ `is_active` column exists in `classification_codes`
- ✅ Indexes created (should see 10+ indexes)
- ✅ Trigram index on `description`
- ✅ Full-text search index
- ✅ Composite indexes

#### Migration 029 (Code Metadata Table)
- ✅ `code_metadata` table exists
- ✅ All columns present (code_type, code, official_description, etc.)
- ✅ Indexes created (should see 8+ indexes)
- ✅ Views created (`code_crosswalk_view`, `code_hierarchy_view`)
- ✅ Trigger created (`update_code_metadata_updated_at`)
- ✅ Function created (`update_code_metadata_updated_at`)
- ✅ `pg_trgm` extension installed

### Step 3: Quick Manual Checks

You can also run these quick checks in Supabase:

#### Check 1: Verify is_active column
```sql
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'classification_codes' 
AND column_name = 'is_active';
```
**Expected**: Should return 1 row with `is_active` column

#### Check 2: Verify code_metadata table
```sql
SELECT table_name 
FROM information_schema.tables 
WHERE table_name = 'code_metadata';
```
**Expected**: Should return 1 row

#### Check 3: Count indexes
```sql
SELECT COUNT(*) as index_count
FROM pg_indexes
WHERE tablename IN ('classification_codes', 'code_metadata')
AND schemaname = 'public';
```
**Expected**: Should return a count of 18+ indexes

#### Check 4: Verify views
```sql
SELECT table_name 
FROM information_schema.views 
WHERE table_schema = 'public'
AND table_name IN ('code_crosswalk_view', 'code_hierarchy_view');
```
**Expected**: Should return 2 rows

## Common Issues

### Issue: "relation already exists"
**Cause**: Migration was already run
**Solution**: This is fine - the migrations use `IF NOT EXISTS` clauses, so they're idempotent

### Issue: "extension pg_trgm does not exist"
**Cause**: Extension not installed
**Solution**: Run this first:
```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;
```

### Issue: "permission denied"
**Cause**: Insufficient permissions
**Solution**: Ensure you're using a user with CREATE privileges

## Next Steps

Once verified:

1. **Populate code_metadata** (optional):
   - Import official code descriptions
   - Add crosswalk data
   - Define code hierarchies

2. **Test the indexes**:
   - Run queries that use the new indexes
   - Monitor query performance

3. **Test streaming responses**:
   - Use `?stream=true` parameter in classification API
   - Verify progress updates are sent

## Verification Checklist

- [ ] `is_active` column exists in `classification_codes`
- [ ] Indexes created on `classification_codes` (10+)
- [ ] `code_metadata` table exists
- [ ] All columns present in `code_metadata`
- [ ] Indexes created on `code_metadata` (8+)
- [ ] Views created (`code_crosswalk_view`, `code_hierarchy_view`)
- [ ] Trigger created (`update_code_metadata_updated_at`)
- [ ] Function created (`update_code_metadata_updated_at`)
- [ ] `pg_trgm` extension installed
- [ ] Test queries run successfully

## Support

If you encounter any issues:
1. Run the verification script to identify what's missing
2. Check the Supabase logs for error messages
3. Verify you have the necessary permissions
4. Check if the migrations were partially applied

