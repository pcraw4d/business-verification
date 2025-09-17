-- =============================================================================
-- Task 1.1.1: Add missing is_active column to keyword_weights table
-- =============================================================================
-- This SQL script executes the first subtask of the Comprehensive Classification
-- Improvement Plan - adding the missing is_active column to the keyword_weights table.
--
-- Instructions:
-- 1. Open Supabase Dashboard
-- 2. Go to SQL Editor
-- 3. Click "New Query"
-- 4. Copy and paste this entire script
-- 5. Click "Run" to execute
--
-- Expected Results:
-- - is_active column added to keyword_weights table
-- - All existing records set to is_active = true
-- - Performance indexes created
-- - No more "is_active does not exist" errors
-- =============================================================================

-- Step 1: Add missing is_active column to keyword_weights table
-- =============================================================================
ALTER TABLE keyword_weights ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

-- Step 2: Update existing records to set is_active = true
-- =============================================================================
UPDATE keyword_weights SET is_active = true WHERE is_active IS NULL;

-- Step 3: Create performance indexes
-- =============================================================================
CREATE INDEX IF NOT EXISTS idx_keyword_weights_active ON keyword_weights(is_active);
CREATE INDEX IF NOT EXISTS idx_keyword_weights_industry_active ON keyword_weights(industry_id, is_active);

-- =============================================================================
-- VERIFICATION QUERIES
-- =============================================================================
-- Run these queries to verify the changes were applied correctly

-- Verification 1: Check that the column exists
SELECT 
    'Column Check' as test_name,
    column_name, 
    data_type, 
    is_nullable,
    column_default
FROM information_schema.columns 
WHERE table_name = 'keyword_weights' 
    AND column_name = 'is_active';

-- Verification 2: Check that all records are active
SELECT 
    'Record Check' as test_name,
    COUNT(*) as total_records, 
    COUNT(CASE WHEN is_active = true THEN 1 END) as active_records,
    COUNT(CASE WHEN is_active = false THEN 1 END) as inactive_records,
    COUNT(CASE WHEN is_active IS NULL THEN 1 END) as null_records
FROM keyword_weights;

-- Verification 3: Check that indexes exist
SELECT 
    'Index Check' as test_name,
    indexname, 
    indexdef 
FROM pg_indexes 
WHERE tablename = 'keyword_weights' 
    AND indexname LIKE '%active%'
ORDER BY indexname;

-- Verification 4: Test query performance with is_active filter
SELECT 
    'Performance Test' as test_name,
    COUNT(*) as filtered_count
FROM keyword_weights 
WHERE is_active = true;

-- =============================================================================
-- SUCCESS INDICATORS
-- =============================================================================
-- After running this script, you should see:
-- 1. Column Check: is_active column with BOOLEAN type and DEFAULT true
-- 2. Record Check: All records should have is_active = true (no null or false records)
-- 3. Index Check: Two indexes should exist (idx_keyword_weights_active and idx_keyword_weights_industry_active)
-- 4. Performance Test: Should return the same count as total_records
--
-- If all verifications pass, Task 1.1.1 is complete!
-- =============================================================================
