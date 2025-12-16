# Migration 040 Troubleshooting Guide

## Error: CREATE INDEX CONCURRENTLY cannot run inside a transaction block

### ✅ Solution

The migration file `040_optimize_classification_queries.sql` has been fixed and **does NOT contain CONCURRENTLY** in any SQL statements.

### Common Causes

1. **Running the wrong file**: Make sure you're running:

   - ✅ `040_optimize_classification_queries.sql` (correct - no CONCURRENTLY)
   - ❌ `040_optimize_classification_queries_concurrent.sql` (wrong - has CONCURRENTLY)

2. **Cached version**: Supabase might be using a cached version

3. **Copy/paste error**: If copying manually, make sure you're copying the entire file

### Verification Steps

1. **Check which file you're running**:

   ```bash
   # Verify the correct file has no CONCURRENTLY
   grep -E "^(CREATE|REFRESH).*CONCURRENTLY" supabase-migrations/040_optimize_classification_queries.sql
   # Should return nothing (empty)
   ```

2. **If using Supabase CLI**:

   ```bash
   # Clear any cache and verify
   supabase db reset  # Only if safe to do so in your environment
   # Or
   supabase migration list  # Check which migrations are applied
   ```

3. **If using Supabase Dashboard**:
   - Go to SQL Editor
   - Copy the ENTIRE contents of `040_optimize_classification_queries.sql`
   - Paste and run
   - Make sure you're not accidentally including the `_concurrent.sql` file

### File Contents Verification

The correct file should have these CREATE INDEX statements:

```sql
CREATE INDEX IF NOT EXISTS idx_code_keywords_composite
ON code_keywords (code_type, keyword, weight DESC);
```

**NOT**:

```sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_code_keywords_composite
```

### If Error Persists

1. **Check Supabase migration history**:

   ```sql
   SELECT * FROM supabase_migrations.schema_migrations
   ORDER BY version DESC LIMIT 10;
   ```

2. **Manually verify the file**:

   - Open `supabase-migrations/040_optimize_classification_queries.sql`
   - Search for "CONCURRENTLY" (case-insensitive)
   - It should ONLY appear in comments, never in actual SQL statements

3. **Try running in smaller chunks**:
   - Run indexes first
   - Then materialized view
   - Then function

### Alternative: Run Without Transaction

If you absolutely need CONCURRENTLY (for production with large tables), you can:

1. Use the `_concurrent.sql` file
2. Run it **outside** of Supabase's migration system
3. Connect directly to PostgreSQL and run it manually (not in a transaction)

```bash
# Connect to database
psql $DATABASE_URL

# Run without transaction wrapper
\i supabase-migrations/040_optimize_classification_queries_concurrent.sql
```
