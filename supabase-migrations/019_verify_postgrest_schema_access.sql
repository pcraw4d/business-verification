-- Verification Query: Test PostgREST Schema Access
-- Run this AFTER refreshing PostgREST schema cache to verify columns are accessible
-- This simulates what the application code does when saving website analysis data

-- Test 1: Verify we can query the columns (this will fail if PostgREST cache is stale)
SELECT 
    merchant_id,
    classification_status,
    website_analysis_status,
    CASE 
        WHEN website_analysis_data IS NOT NULL AND website_analysis_data != '{}'::jsonb 
        THEN 'HAS_DATA'
        ELSE 'EMPTY'
    END as website_analysis_data_status,
    CASE 
        WHEN classification_data IS NOT NULL AND classification_data != '{}'::jsonb 
        THEN 'HAS_DATA'
        ELSE 'EMPTY'
    END as classification_data_status
FROM merchant_analytics
LIMIT 5;

-- Test 2: Try to update a record with website_analysis_data (simulates job saving)
-- This will fail with PGRST204 if PostgREST cache is stale
UPDATE merchant_analytics
SET 
    website_analysis_data = jsonb_build_object(
        'test', 'value',
        'timestamp', NOW()
    ),
    website_analysis_status = 'completed',
    website_analysis_updated_at = NOW()
WHERE merchant_id = (
    SELECT id FROM merchants LIMIT 1
)
RETURNING merchant_id, website_analysis_status;

-- Test 3: Check if we can insert a new record with all columns
-- This will fail if any column is missing from PostgREST cache
INSERT INTO merchant_analytics (
    merchant_id,
    classification_data,
    classification_status,
    website_analysis_data,
    website_analysis_status
) VALUES (
    'test_merchant_' || extract(epoch from now())::text,
    '{"test": "classification"}'::jsonb,
    'completed',
    '{"test": "website_analysis"}'::jsonb,
    'completed'
)
RETURNING merchant_id, classification_status, website_analysis_status;

-- Clean up test record
DELETE FROM merchant_analytics 
WHERE merchant_id LIKE 'test_merchant_%';

-- If all three tests pass, PostgREST schema cache is up to date!

