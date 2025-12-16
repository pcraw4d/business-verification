-- Check if Migration 040 indexes and objects were created successfully

-- 1. Check if indexes exist
SELECT 
    'Indexes on code_keywords' as check_type,
    COUNT(*) as count,
    CASE 
        WHEN COUNT(*) >= 2 THEN '✅ PASS'
        ELSE '❌ FAIL - Missing indexes'
    END as status
FROM pg_indexes 
WHERE tablename = 'code_keywords' 
  AND indexname IN ('idx_code_keywords_keyword_relevance', 'idx_code_keywords_code_id');

SELECT 
    'Indexes on classification_codes' as check_type,
    COUNT(*) as count,
    CASE 
        WHEN COUNT(*) >= 3 THEN '✅ PASS'
        ELSE '❌ FAIL - Missing indexes'
    END as status
FROM pg_indexes 
WHERE tablename = 'classification_codes' 
  AND indexname IN ('idx_codes_type_description', 'idx_codes_type_code_active', 'idx_classification_codes_code_type_code');

-- 2. Check if materialized view exists
SELECT 
    'Materialized View' as check_type,
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM pg_matviews 
            WHERE matviewname = 'code_search_cache'
        ) THEN '✅ PASS - code_search_cache exists'
        ELSE '❌ FAIL - code_search_cache does not exist'
    END as status;

-- 3. Check if refresh function exists
SELECT 
    'Refresh Function' as check_type,
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM pg_proc p
            JOIN pg_namespace n ON p.pronamespace = n.oid
            WHERE n.nspname = 'public'
              AND p.proname = 'refresh_code_search_cache'
        ) THEN '✅ PASS - refresh_code_search_cache function exists'
        ELSE '❌ FAIL - refresh_code_search_cache function does not exist'
    END as status;

-- 4. List all indexes created by migration 040
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE indexname LIKE 'idx_code_keywords%'
   OR indexname LIKE 'idx_codes%'
   OR indexname LIKE 'idx_code_search_cache%'
   OR indexname LIKE 'idx_code_metadata_code_type_code'
ORDER BY tablename, indexname;

-- 5. Check materialized view row count (if it exists)
SELECT 
    'Materialized View Data' as check_type,
    CASE 
        WHEN EXISTS (SELECT 1 FROM pg_matviews WHERE matviewname = 'code_search_cache') THEN
            (SELECT COUNT(*)::text || ' rows' FROM code_search_cache)
        ELSE 'N/A - View does not exist'
    END as status;
