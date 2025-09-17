-- =============================================================================
-- Subtask 1.1.3 Verification: Performance Indexes
-- =============================================================================
-- This script verifies that subtask 1.1.3 has been completed successfully.
-- Subtask 1.1.3: Create performance indexes for keyword_weights table
--
-- Instructions:
-- 1. Open Supabase Dashboard
-- 2. Go to SQL Editor
-- 3. Click "New Query"
-- 4. Copy and paste this entire script
-- 5. Click "Run" to execute
--
-- Expected Results:
-- - Required indexes should exist
-- - Enhanced indexes should be present (professional best practices)
-- - Query performance should show index usage
-- =============================================================================

-- =============================================================================
-- VERIFICATION QUERIES FOR SUBTASK 1.1.3
-- =============================================================================

-- Verification 1: Check that all required indexes exist
SELECT 
    'Subtask 1.1.3 - Required Indexes Check' as test_name,
    indexname, 
    indexdef 
FROM pg_indexes 
WHERE tablename = 'keyword_weights' 
    AND indexname IN (
        'idx_keyword_weights_active',
        'idx_keyword_weights_industry_active'
    )
ORDER BY indexname;

-- Verification 2: Check for enhanced indexes (professional best practices)
SELECT 
    'Subtask 1.1.3 - Enhanced Indexes Check' as test_name,
    indexname, 
    indexdef 
FROM pg_indexes 
WHERE tablename = 'keyword_weights' 
    AND indexname IN (
        'idx_keyword_weights_keyword_active',
        'idx_keyword_weights_industry_weight_active',
        'idx_keyword_weights_search_active'
    )
ORDER BY indexname;

-- Verification 3: Show all active-related indexes
SELECT 
    'Subtask 1.1.3 - All Active Indexes' as test_name,
    indexname, 
    indexdef 
FROM pg_indexes 
WHERE tablename = 'keyword_weights' 
    AND indexname LIKE '%active%'
ORDER BY indexname;

-- Verification 4: Test query performance with EXPLAIN
SELECT 
    'Subtask 1.1.3 - Query Performance Test 1' as test_name,
    'Basic is_active query performance' as description;

EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM keyword_weights 
WHERE is_active = true 
LIMIT 10;

-- Verification 5: Test industry-based query performance
SELECT 
    'Subtask 1.1.3 - Query Performance Test 2' as test_name,
    'Industry-based query performance' as description;

EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM keyword_weights 
WHERE industry_id = 1 AND is_active = true 
ORDER BY base_weight DESC 
LIMIT 10;

-- Verification 6: Check index usage statistics
SELECT 
    'Subtask 1.1.3 - Index Usage Statistics' as test_name,
    schemaname, 
    tablename, 
    indexname, 
    idx_scan, 
    idx_tup_read, 
    idx_tup_fetch
FROM pg_stat_user_indexes 
WHERE tablename = 'keyword_weights' 
    AND indexname LIKE '%active%'
ORDER BY idx_scan DESC;

-- Verification 7: Check table statistics
SELECT 
    'Subtask 1.1.3 - Table Statistics' as test_name,
    schemaname, 
    tablename, 
    n_tup_ins as inserts,
    n_tup_upd as updates,
    n_tup_del as deletes,
    n_live_tup as live_tuples,
    n_dead_tup as dead_tuples
FROM pg_stat_user_tables 
WHERE tablename = 'keyword_weights';

-- =============================================================================
-- SUCCESS CRITERIA FOR SUBTASK 1.1.3
-- =============================================================================
-- After running this script, you should see:

-- 1. Required Indexes Check: 
--    - idx_keyword_weights_active should exist
--    - idx_keyword_weights_industry_active should exist

-- 2. Enhanced Indexes Check:
--    - idx_keyword_weights_keyword_active (optional, professional best practice)
--    - idx_keyword_weights_industry_weight_active (optional, professional best practice)
--    - idx_keyword_weights_search_active (optional, professional best practice)

-- 3. All Active Indexes: Should show all indexes containing 'active' in the name

-- 4. Query Performance Tests: 
--    - EXPLAIN should show "Index Scan" or "Bitmap Index Scan" for good performance
--    - Should NOT show "Seq Scan" (sequential scan) for optimal performance

-- 5. Index Usage Statistics: Should show indexes (initially with 0 scans if not used yet)

-- 6. Table Statistics: Should show current table statistics

-- If all verifications pass, Subtask 1.1.3 is complete!
-- =============================================================================

-- =============================================================================
-- ADDITIONAL PERFORMANCE MONITORING QUERIES
-- =============================================================================
-- These queries can be used to monitor performance over time

-- Monitor slow queries (requires pg_stat_statements extension)
-- SELECT query, calls, total_time, mean_time, rows
-- FROM pg_stat_statements 
-- WHERE query LIKE '%keyword_weights%' 
--   AND mean_time > 100 
-- ORDER BY mean_time DESC 
-- LIMIT 10;

-- Monitor index usage over time
-- SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
-- FROM pg_stat_user_indexes 
-- WHERE tablename = 'keyword_weights' 
-- ORDER BY idx_scan DESC;

-- Monitor table size and growth
-- SELECT schemaname, tablename, 
--        pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
-- FROM pg_tables 
-- WHERE tablename = 'keyword_weights';

-- =============================================================================
-- SUMMARY
-- =============================================================================
-- This verification script confirms that:
-- 1. Required performance indexes are created for keyword_weights table
-- 2. Enhanced indexes are present (following professional best practices)
-- 3. Query performance is optimized with proper index usage
-- 4. Database statistics are available for monitoring
-- 5. Subtask 1.1.3 is complete and ready for next phase
-- =============================================================================
