-- Verification script for migration 040
-- Run this to check if the migration file is correct before applying

-- Check 1: Verify no CONCURRENTLY in CREATE INDEX statements
DO $$
DECLARE
    file_content TEXT;
    has_concurrent BOOLEAN := FALSE;
BEGIN
    -- This is a manual check - you should verify the file doesn't contain CONCURRENTLY
    -- in actual SQL statements (only in comments)
    RAISE NOTICE 'Please manually verify that 040_optimize_classification_queries.sql';
    RAISE NOTICE 'does NOT contain CONCURRENTLY in any CREATE INDEX or REFRESH statements';
    RAISE NOTICE '(CONCURRENTLY is OK in comments)';
END $$;

-- Check 2: Verify indexes exist (after migration has been run)
-- Note: Updated to check for the correct index names from the fixed migration
SELECT 
    'Migration 040 Status' as check_type,
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM pg_indexes 
            WHERE indexname = 'idx_code_keywords_keyword_relevance'
        ) THEN '✅ Migration applied - indexes exist'
        ELSE '⚠️ Migration not yet applied - indexes do not exist'
    END as migration_status;

-- Check 3: List all Phase 2 indexes
SELECT 
    'Phase 2 Indexes' as check_type,
    expected.indexname,
    expected.tablename,
    CASE 
        WHEN actual.indexname IS NOT NULL THEN '✅ Exists'
        ELSE '❌ Missing'
    END as status
FROM (
    SELECT 'idx_code_keywords_keyword_relevance' as indexname, 'code_keywords' as tablename
    UNION ALL SELECT 'idx_code_keywords_code_id', 'code_keywords'
    UNION ALL SELECT 'idx_codes_type_description', 'classification_codes'
    UNION ALL SELECT 'idx_codes_type_code_active', 'classification_codes'
    UNION ALL SELECT 'idx_classification_codes_code_type_code', 'classification_codes'
    UNION ALL SELECT 'idx_code_metadata_code_type_code', 'code_metadata'
) expected
LEFT JOIN pg_indexes actual 
    ON actual.indexname = expected.indexname 
    AND actual.tablename = expected.tablename
ORDER BY expected.tablename, expected.indexname;
