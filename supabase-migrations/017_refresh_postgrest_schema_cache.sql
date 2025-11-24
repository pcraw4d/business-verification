-- Migration: Refresh PostgREST Schema Cache
-- Date: January 2025
-- Purpose: Force PostgREST to reload its schema cache after adding new columns
-- 
-- IMPORTANT: After running this, you may need to:
-- 1. Wait a few seconds for PostgREST to reload
-- 2. Or restart the PostgREST service in Supabase dashboard
-- 3. Or use Supabase API to reload schema: POST /rest/v1/rpc/reload_schema

-- This query doesn't actually do anything, but running it will help verify the columns exist
-- The real fix is to reload PostgREST schema cache via Supabase dashboard or API

-- Verify that the columns exist in the database
SELECT 
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'merchant_analytics'
  AND column_name IN (
    'website_analysis_data',
    'website_analysis_status',
    'website_analysis_updated_at',
    'classification_status',
    'classification_updated_at',
    'classification_data'
  )
ORDER BY column_name;

-- If the above query returns all 6 columns, they exist in the database
-- The issue is that PostgREST's schema cache is stale
-- 
-- To fix:
-- 1. Go to Supabase Dashboard > Settings > API
-- 2. Click "Reload Schema" or "Refresh Schema Cache"
-- 3. Or use the Supabase CLI: supabase db reset (in development)
-- 4. Or wait 5-10 minutes for automatic cache refresh

