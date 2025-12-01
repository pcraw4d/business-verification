-- =====================================================
-- Verification Script for Migrations 028 and 029
-- Purpose: Verify that all tables, columns, indexes, and views were created successfully
-- =====================================================

-- =====================================================
-- 1. Verify Migration 028: Enhanced Classification Schema
-- =====================================================

-- Check if is_active column exists in classification_codes
SELECT 
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'classification_codes' 
            AND column_name = 'is_active'
        ) THEN '✅ is_active column exists'
        ELSE '❌ is_active column MISSING'
    END AS is_active_column_check;

-- List all indexes on classification_codes table
SELECT 
    'Migration 028 Indexes' AS check_type,
    indexname AS index_name,
    indexdef AS index_definition
FROM pg_indexes
WHERE tablename = 'classification_codes'
AND schemaname = 'public'
ORDER BY indexname;

-- Count indexes on classification_codes
SELECT 
    'Total indexes on classification_codes' AS metric,
    COUNT(*) AS count
FROM pg_indexes
WHERE tablename = 'classification_codes'
AND schemaname = 'public';

-- Check for specific important indexes
SELECT 
    'Index Check' AS check_type,
    indexname AS index_name,
    CASE 
        WHEN indexname LIKE '%active%' THEN '✅ Active index'
        WHEN indexname LIKE '%trgm%' THEN '✅ Trigram index'
        WHEN indexname LIKE '%fts%' THEN '✅ Full-text search index'
        WHEN indexname LIKE '%type_active%' THEN '✅ Composite type_active index'
        WHEN indexname LIKE '%industry_type%' THEN '✅ Composite industry_type index'
        ELSE 'Other index'
    END AS index_type
FROM pg_indexes
WHERE tablename = 'classification_codes'
AND schemaname = 'public'
ORDER BY indexname;

-- =====================================================
-- 2. Verify Migration 029: Code Metadata Table
-- =====================================================

-- Check if code_metadata table exists
SELECT 
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM information_schema.tables 
            WHERE table_name = 'code_metadata'
            AND table_schema = 'public'
        ) THEN '✅ code_metadata table exists'
        ELSE '❌ code_metadata table MISSING'
    END AS table_check;

-- List all columns in code_metadata table
SELECT 
    'code_metadata columns' AS check_type,
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'code_metadata'
AND table_schema = 'public'
ORDER BY ordinal_position;

-- List all indexes on code_metadata table
SELECT 
    'Migration 029 Indexes' AS check_type,
    indexname AS index_name,
    indexdef AS index_definition
FROM pg_indexes
WHERE tablename = 'code_metadata'
AND schemaname = 'public'
ORDER BY indexname;

-- Count indexes on code_metadata
SELECT 
    'Total indexes on code_metadata' AS metric,
    COUNT(*) AS count
FROM pg_indexes
WHERE tablename = 'code_metadata'
AND schemaname = 'public';

-- Check if views exist
SELECT 
    'View Check' AS check_type,
    table_name AS view_name,
    CASE 
        WHEN table_name = 'code_crosswalk_view' THEN '✅ Crosswalk view'
        WHEN table_name = 'code_hierarchy_view' THEN '✅ Hierarchy view'
        ELSE 'Other view'
    END AS view_type
FROM information_schema.views
WHERE table_schema = 'public'
AND table_name IN ('code_crosswalk_view', 'code_hierarchy_view');

-- Check if trigger exists
SELECT 
    'Trigger Check' AS check_type,
    trigger_name,
    event_manipulation,
    action_timing
FROM information_schema.triggers
WHERE event_object_table = 'code_metadata'
AND trigger_schema = 'public';

-- Check if function exists
SELECT 
    'Function Check' AS check_type,
    routine_name,
    routine_type
FROM information_schema.routines
WHERE routine_schema = 'public'
AND routine_name = 'update_code_metadata_updated_at';

-- =====================================================
-- 3. Verify pg_trgm extension (required for trigram indexes)
-- =====================================================

SELECT 
    'Extension Check' AS check_type,
    extname AS extension_name,
    CASE 
        WHEN extname = 'pg_trgm' THEN '✅ pg_trgm extension installed'
        ELSE 'Other extension'
    END AS status
FROM pg_extension
WHERE extname = 'pg_trgm';

-- =====================================================
-- 4. Summary Report
-- =====================================================

SELECT 
    '=== MIGRATION VERIFICATION SUMMARY ===' AS summary;

-- Count all objects created
SELECT 
    'Total Objects Created' AS category,
    (SELECT COUNT(*) FROM pg_indexes WHERE tablename = 'classification_codes' AND schemaname = 'public') AS classification_codes_indexes,
    (SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'code_metadata' AND table_schema = 'public') AS code_metadata_columns,
    (SELECT COUNT(*) FROM pg_indexes WHERE tablename = 'code_metadata' AND schemaname = 'public') AS code_metadata_indexes,
    (SELECT COUNT(*) FROM information_schema.views WHERE table_schema = 'public' AND table_name IN ('code_crosswalk_view', 'code_hierarchy_view')) AS views_created,
    (SELECT COUNT(*) FROM information_schema.triggers WHERE event_object_table = 'code_metadata' AND trigger_schema = 'public') AS triggers_created;

-- =====================================================
-- 5. Test Queries (to verify functionality)
-- =====================================================

-- Test: Check if we can query classification_codes with is_active filter
SELECT 
    'Test Query: classification_codes with is_active' AS test_name,
    COUNT(*) AS total_codes,
    COUNT(*) FILTER (WHERE is_active = true) AS active_codes,
    COUNT(*) FILTER (WHERE is_active = false) AS inactive_codes
FROM classification_codes
LIMIT 1;

-- Test: Check if code_metadata table is queryable (should return 0 rows if empty)
SELECT 
    'Test Query: code_metadata table' AS test_name,
    COUNT(*) AS total_records
FROM code_metadata;

-- Test: Check if views are queryable
SELECT 
    'Test Query: code_crosswalk_view' AS test_name,
    COUNT(*) AS total_records
FROM code_crosswalk_view;

SELECT 
    'Test Query: code_hierarchy_view' AS test_name,
    COUNT(*) AS total_records
FROM code_hierarchy_view;

-- =====================================================
-- Expected Results:
-- =====================================================
-- ✅ is_active column should exist in classification_codes
-- ✅ Multiple indexes should exist on classification_codes (at least 10+)
-- ✅ code_metadata table should exist with all columns
-- ✅ Multiple indexes should exist on code_metadata (at least 8+)
-- ✅ code_crosswalk_view should exist
-- ✅ code_hierarchy_view should exist
-- ✅ update_code_metadata_updated_at trigger should exist
-- ✅ update_code_metadata_updated_at function should exist
-- ✅ pg_trgm extension should be installed
-- =====================================================

