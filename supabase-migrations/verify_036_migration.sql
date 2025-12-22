-- Verification script for migration 036
-- Run this to check if get_codes_by_keywords function exists

-- Check 1: Verify function exists
SELECT 
    'Function get_codes_by_keywords' as check_type,
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM pg_proc p
            JOIN pg_namespace n ON p.pronamespace = n.oid
            WHERE n.nspname = 'public'
              AND p.proname = 'get_codes_by_keywords'
        ) THEN '✅ PASS - Function exists'
        ELSE '❌ FAIL - Function does not exist'
    END as status;

-- Check 2: Verify function signature
SELECT 
    p.proname as function_name,
    pg_get_function_arguments(p.oid) as arguments,
    pg_get_function_result(p.oid) as return_type
FROM pg_proc p
JOIN pg_namespace n ON p.pronamespace = n.oid
WHERE n.nspname = 'public'
  AND p.proname = 'get_codes_by_keywords';

-- Check 3: Verify index exists
SELECT 
    'Index idx_code_keywords_keyword_lookup' as check_type,
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM pg_indexes 
            WHERE indexname = 'idx_code_keywords_keyword_lookup'
        ) THEN '✅ PASS - Index exists'
        ELSE '❌ FAIL - Index does not exist'
    END as status;

-- Check 4: Test function with sample data (if code_keywords has data)
-- This will only work if there's data in code_keywords table
SELECT 
    'Function test' as check_type,
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM code_keywords LIMIT 1
        ) THEN 
            CASE 
                WHEN (SELECT COUNT(*) FROM get_codes_by_keywords('MCC', ARRAY['test'], 1)) >= 0 
                THEN '✅ PASS - Function can be called (may return 0 results if no matching data)'
                ELSE '❌ FAIL - Function call failed'
            END
        ELSE '⚠️ SKIP - No data in code_keywords table to test'
    END as status;

