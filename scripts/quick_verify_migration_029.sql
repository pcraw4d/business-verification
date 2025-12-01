-- Quick Verification for Migration 029: Code Metadata Table
-- Run this to verify all components were created

-- 1. Check if table exists
SELECT 
    'Table Check' AS check_type,
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM information_schema.tables 
            WHERE table_name = 'code_metadata'
            AND table_schema = 'public'
        ) THEN '✅ code_metadata table EXISTS'
        ELSE '❌ code_metadata table MISSING'
    END AS result;

-- 2. List all columns (should have 15+ columns)
SELECT 
    'Column Check' AS check_type,
    column_name,
    data_type
FROM information_schema.columns
WHERE table_name = 'code_metadata'
AND table_schema = 'public'
ORDER BY ordinal_position;

-- 3. Count columns (should be 15)
SELECT 
    'Column Count' AS check_type,
    COUNT(*) AS total_columns,
    CASE 
        WHEN COUNT(*) >= 15 THEN '✅ Correct number of columns'
        ELSE '❌ Missing columns'
    END AS status
FROM information_schema.columns
WHERE table_name = 'code_metadata'
AND table_schema = 'public';

-- 4. List all indexes (should have 8+ indexes)
SELECT 
    'Index Check' AS check_type,
    indexname AS index_name
FROM pg_indexes
WHERE tablename = 'code_metadata'
AND schemaname = 'public'
ORDER BY indexname;

-- 5. Count indexes (should be 8+)
SELECT 
    'Index Count' AS check_type,
    COUNT(*) AS total_indexes,
    CASE 
        WHEN COUNT(*) >= 8 THEN '✅ Correct number of indexes'
        ELSE '❌ Missing indexes'
    END AS status
FROM pg_indexes
WHERE tablename = 'code_metadata'
AND schemaname = 'public';

-- 6. Check views (should have 2 views)
SELECT 
    'View Check' AS check_type,
    table_name AS view_name,
    CASE 
        WHEN table_name = 'code_crosswalk_view' THEN '✅ Crosswalk view'
        WHEN table_name = 'code_hierarchy_view' THEN '✅ Hierarchy view'
        ELSE 'Other view'
    END AS status
FROM information_schema.views
WHERE table_schema = 'public'
AND table_name IN ('code_crosswalk_view', 'code_hierarchy_view')
ORDER BY table_name;

-- 7. Count views (should be 2)
SELECT 
    'View Count' AS check_type,
    COUNT(*) AS total_views,
    CASE 
        WHEN COUNT(*) = 2 THEN '✅ Both views exist'
        ELSE '❌ Missing views'
    END AS status
FROM information_schema.views
WHERE table_schema = 'public'
AND table_name IN ('code_crosswalk_view', 'code_hierarchy_view');

-- 8. Check trigger
SELECT 
    'Trigger Check' AS check_type,
    trigger_name,
    event_manipulation,
    action_timing,
    CASE 
        WHEN trigger_name = 'update_code_metadata_updated_at' THEN '✅ Trigger exists'
        ELSE 'Other trigger'
    END AS status
FROM information_schema.triggers
WHERE event_object_table = 'code_metadata'
AND trigger_schema = 'public';

-- 9. Check function (you already confirmed this)
SELECT 
    'Function Check' AS check_type,
    routine_name,
    routine_type,
    '✅ Function exists' AS status
FROM information_schema.routines
WHERE routine_schema = 'public'
AND routine_name = 'update_code_metadata_updated_at';

-- 10. Test that table is queryable (should return 0 rows if empty)
SELECT 
    'Table Query Test' AS check_type,
    COUNT(*) AS total_records,
    CASE 
        WHEN COUNT(*) = 0 THEN '✅ Table is empty (expected)'
        ELSE 'Table has ' || COUNT(*) || ' records'
    END AS status
FROM code_metadata;

-- 11. Test views are queryable (should return 0 rows if empty)
SELECT 
    'View Query Test: code_crosswalk_view' AS check_type,
    COUNT(*) AS total_records
FROM code_crosswalk_view;

SELECT 
    'View Query Test: code_hierarchy_view' AS check_type,
    COUNT(*) AS total_records
FROM code_hierarchy_view;

-- 12. SUMMARY
SELECT 
    '=== MIGRATION 029 SUMMARY ===' AS summary;

SELECT 
    'Total Objects' AS category,
    (SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'code_metadata' AND table_schema = 'public') AS columns,
    (SELECT COUNT(*) FROM pg_indexes WHERE tablename = 'code_metadata' AND schemaname = 'public') AS indexes,
    (SELECT COUNT(*) FROM information_schema.views WHERE table_schema = 'public' AND table_name IN ('code_crosswalk_view', 'code_hierarchy_view')) AS views,
    (SELECT COUNT(*) FROM information_schema.triggers WHERE event_object_table = 'code_metadata' AND trigger_schema = 'public') AS triggers,
    (SELECT COUNT(*) FROM information_schema.routines WHERE routine_schema = 'public' AND routine_name = 'update_code_metadata_updated_at') AS functions;

