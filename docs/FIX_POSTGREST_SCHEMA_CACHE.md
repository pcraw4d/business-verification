# Fix PostgREST Schema Cache Issue

## Problem

The error `(PGRST204) Could not find the 'website_analysis_data' column of 'merchant_analytics' in the schema cache` indicates that:

1. ✅ The database migration has been run (columns exist in PostgreSQL)
2. ❌ PostgREST's schema cache is stale (doesn't know about the new columns)

## Solution

### Option 1: Supabase Dashboard (Recommended)

1. Go to your Supabase project dashboard
2. Navigate to **Settings** > **API**
3. Look for **"Reload Schema"** or **"Refresh Schema Cache"** button
4. Click it to force PostgREST to reload the schema

### Option 2: Supabase CLI / API Script

For hosted Supabase projects, use the provided script:

```bash
# Make script executable (if not already)
chmod +x scripts/refresh_postgrest_schema_api.sh

# Run the script (it will use SUPABASE_SERVICE_ROLE_KEY from railway.env)
./scripts/refresh_postgrest_schema_api.sh

# Or provide the key directly
./scripts/refresh_postgrest_schema_api.sh "your-service-role-key"
```

The script attempts to call the `reload_schema` RPC function. If that's not available, it will provide instructions for using the dashboard.

### Option 3: Wait for Automatic Refresh

PostgREST automatically refreshes its schema cache every 5-10 minutes. If you can wait, the issue will resolve itself.

### Option 4: Manual API Call (Advanced)

If you have access to the Supabase management API:

```bash
curl -X POST \
  'https://your-project.supabase.co/rest/v1/rpc/reload_schema' \
  -H 'apikey: YOUR_SERVICE_ROLE_KEY' \
  -H 'Authorization: Bearer YOUR_SERVICE_ROLE_KEY'
```

## Verification

After refreshing the schema cache, verify the columns exist:

```sql
SELECT column_name, data_type 
FROM information_schema.columns
WHERE table_name = 'merchant_analytics'
  AND column_name IN (
    'website_analysis_data',
    'website_analysis_status',
    'classification_status',
    'classification_data'
  );
```

All 4 columns should be returned.

## Why This Happens

PostgREST caches the database schema for performance. When you run migrations that add new columns, PostgREST doesn't automatically know about them until its cache is refreshed.

## Prevention

For future migrations:
1. Run migrations during low-traffic periods
2. Immediately refresh the schema cache after running migrations
3. Consider using Supabase's migration system which handles this automatically

