-- =============================================================================
-- Subtask 1.1.3: Create Performance Indexes for keyword_weights Table
-- =============================================================================
-- This SQL script creates the performance indexes required for Task 1.1.3
-- of the Comprehensive Classification Improvement Plan.
--
-- Instructions:
-- 1. Open Supabase Dashboard
-- 2. Go to SQL Editor
-- 3. Click "New Query"
-- 4. Copy and paste this entire script
-- 5. Click "Run" to execute
--
-- Expected Results:
-- - Performance indexes created for keyword_weights table
-- - Indexes optimized for is_active column filtering
-- - Query performance improved for classification operations
-- =============================================================================

-- =============================================================================
-- SUBTASK 1.1.3: PERFORMANCE INDEXES CREATION
-- =============================================================================

-- Index 1: Basic is_active index (as specified in the plan)
-- This index enables fast filtering by is_active status
CREATE INDEX IF NOT EXISTS idx_keyword_weights_active 
ON keyword_weights(is_active);

-- Index 2: Composite industry_id + is_active index (as specified in the plan)
-- This index enables fast filtering by industry and active status
CREATE INDEX IF NOT EXISTS idx_keyword_weights_industry_active 
ON keyword_weights(industry_id, is_active);

-- =============================================================================
-- ENHANCED PERFORMANCE INDEXES (Professional Best Practices)
-- =============================================================================
-- These additional indexes follow professional database optimization principles
-- and are based on the existing database_optimization.sql patterns

-- Index 3: Keyword + is_active composite index
-- Optimizes the most common query pattern: keyword lookup with active filtering
CREATE INDEX IF NOT EXISTS idx_keyword_weights_keyword_active 
ON keyword_weights (keyword, is_active) 
WHERE is_active = true;

-- Index 4: Industry + is_active + weight ordering
-- Optimizes industry-based queries with weight sorting
CREATE INDEX IF NOT EXISTS idx_keyword_weights_industry_weight_active 
ON keyword_weights (industry_id, is_active, base_weight DESC) 
WHERE is_active = true;

-- Index 5: Search optimization index
-- Optimizes general keyword searches with weight ordering
CREATE INDEX IF NOT EXISTS idx_keyword_weights_search_active 
ON keyword_weights (is_active, base_weight DESC, keyword) 
WHERE is_active = true;

-- =============================================================================
-- VERIFICATION QUERIES
-- =============================================================================
-- Run these queries to verify the indexes were created successfully

-- Verification 1: Check that all indexes exist
SELECT 
    'Index Verification' as test_name,
    indexname, 
    indexdef 
FROM pg_indexes 
WHERE tablename = 'keyword_weights' 
    AND indexname LIKE '%active%'
ORDER BY indexname;

-- Verification 2: Check index usage statistics (will be empty initially)
SELECT 
    'Index Statistics' as test_name,
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

-- Verification 3: Test query performance with EXPLAIN
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM keyword_weights 
WHERE is_active = true 
LIMIT 10;

-- Verification 4: Test industry-based query performance
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM keyword_weights 
WHERE industry_id = 1 AND is_active = true 
ORDER BY base_weight DESC 
LIMIT 10;

-- =============================================================================
-- PERFORMANCE MONITORING QUERIES
-- =============================================================================
-- These queries can be used to monitor index performance over time

-- Monitor index usage
-- SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
-- FROM pg_stat_user_indexes 
-- WHERE tablename = 'keyword_weights' 
-- ORDER BY idx_scan DESC;

-- Monitor slow queries (requires pg_stat_statements extension)
-- SELECT query, calls, total_time, mean_time, rows
-- FROM pg_stat_statements 
-- WHERE query LIKE '%keyword_weights%' 
--   AND mean_time > 100 
-- ORDER BY mean_time DESC 
-- LIMIT 10;

-- =============================================================================
-- SUCCESS CRITERIA VALIDATION
-- =============================================================================
-- After running this script, you should see:

-- 1. Index Verification: 5 indexes should be created:
--    - idx_keyword_weights_active
--    - idx_keyword_weights_industry_active  
--    - idx_keyword_weights_keyword_active
--    - idx_keyword_weights_industry_weight_active
--    - idx_keyword_weights_search_active

-- 2. Index Statistics: All indexes should be visible (initially with 0 scans)

-- 3. Query Performance: EXPLAIN should show index usage (Index Scan or Bitmap Index Scan)

-- 4. No Errors: Script should complete without any error messages

-- If all verifications pass, Subtask 1.1.3 is complete!
-- =============================================================================

-- =============================================================================
-- ADDITIONAL OPTIMIZATIONS (Optional)
-- =============================================================================
-- These optimizations can be applied for even better performance

-- Update table statistics for better query planning
ANALYZE keyword_weights;

-- Optional: Create partial indexes for high-weight keywords
-- CREATE INDEX IF NOT EXISTS idx_keyword_weights_high_weight 
-- ON keyword_weights (keyword, industry_id, base_weight) 
-- WHERE is_active = true AND base_weight > 0.5;

-- Optional: Create partial index for frequently used keywords
-- CREATE INDEX IF NOT EXISTS idx_keyword_weights_frequent 
-- ON keyword_weights (keyword, industry_id, usage_count DESC) 
-- WHERE is_active = true AND usage_count > 10;

-- =============================================================================
-- ROLLBACK INSTRUCTIONS (If Needed)
-- =============================================================================
-- If you need to rollback these changes, run:
-- 
-- DROP INDEX IF EXISTS idx_keyword_weights_active;
-- DROP INDEX IF EXISTS idx_keyword_weights_industry_active;
-- DROP INDEX IF EXISTS idx_keyword_weights_keyword_active;
-- DROP INDEX IF EXISTS idx_keyword_weights_industry_weight_active;
-- DROP INDEX IF EXISTS idx_keyword_weights_search_active;
-- =============================================================================
