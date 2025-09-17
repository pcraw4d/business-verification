-- =============================================================================
-- Subtask 1.1.2 Verification: Update existing records
-- =============================================================================
-- This script verifies that subtask 1.1.2 has been completed successfully.
-- Subtask 1.1.2: Update existing records to set is_active = true
--
-- Instructions:
-- 1. Open Supabase Dashboard
-- 2. Go to SQL Editor
-- 3. Click "New Query"
-- 4. Copy and paste this entire script
-- 5. Click "Run" to execute
--
-- Expected Results:
-- - All records should have is_active = true
-- - No records should have is_active = NULL
-- - No records should have is_active = false (unless intentionally set)
-- =============================================================================

-- =============================================================================
-- VERIFICATION QUERIES FOR SUBTASK 1.1.2
-- =============================================================================

-- Verification 1: Check that all records have is_active = true
SELECT 
    'Subtask 1.1.2 - Record Status Check' as test_name,
    COUNT(*) as total_records, 
    COUNT(CASE WHEN is_active = true THEN 1 END) as active_records,
    COUNT(CASE WHEN is_active = false THEN 1 END) as inactive_records,
    COUNT(CASE WHEN is_active IS NULL THEN 1 END) as null_records
FROM keyword_weights;

-- Verification 2: Check for any NULL values (should be 0)
SELECT 
    'Subtask 1.1.2 - NULL Check' as test_name,
    COUNT(*) as null_count
FROM keyword_weights 
WHERE is_active IS NULL;

-- Verification 3: Show sample of records to verify is_active values
SELECT 
    'Subtask 1.1.2 - Sample Records' as test_name,
    id,
    industry_id,
    keyword,
    is_active,
    base_weight,
    updated_at
FROM keyword_weights 
ORDER BY id 
LIMIT 10;

-- Verification 4: Check if any records were recently updated (indicating the UPDATE ran)
SELECT 
    'Subtask 1.1.2 - Recent Updates' as test_name,
    COUNT(*) as recently_updated_count
FROM keyword_weights 
WHERE updated_at >= NOW() - INTERVAL '1 hour';

-- =============================================================================
-- SUCCESS CRITERIA FOR SUBTASK 1.1.2
-- =============================================================================
-- After running this script, you should see:
-- 1. Record Status Check: 
--    - total_records = active_records (all records should be active)
--    - inactive_records = 0 (no inactive records unless intentionally set)
--    - null_records = 0 (no NULL values)
-- 2. NULL Check: null_count = 0
-- 3. Sample Records: All records should show is_active = true
-- 4. Recent Updates: Should show some records if the UPDATE was recently executed
--
-- If all verifications pass, Subtask 1.1.2 is complete!
-- =============================================================================

-- =============================================================================
-- ADDITIONAL TEST: Verify the UPDATE statement would work if needed
-- =============================================================================
-- This query shows what the UPDATE statement would affect (should be 0 rows)
SELECT 
    'Subtask 1.1.2 - UPDATE Test' as test_name,
    COUNT(*) as records_that_would_be_updated
FROM keyword_weights 
WHERE is_active IS NULL;

-- =============================================================================
-- SUMMARY
-- =============================================================================
-- This verification script confirms that:
-- 1. All existing records in keyword_weights have is_active = true
-- 2. No records have is_active = NULL
-- 3. The UPDATE statement from Task 1.1.1 has been successfully applied
-- 4. Subtask 1.1.2 is complete
-- =============================================================================
