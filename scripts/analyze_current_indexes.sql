-- =============================================================================
-- INDEX ANALYSIS SCRIPT FOR SUPABASE TABLE IMPROVEMENT IMPLEMENTATION
-- Subtask 3.2.1: Analyze Current Indexes
-- =============================================================================
-- This script analyzes all existing indexes in the Supabase database
-- to identify optimization opportunities for the classification system

-- =============================================================================
-- 1. REVIEW ALL EXISTING INDEXES
-- =============================================================================

-- Get all indexes in the database with their details
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef,
    pg_size_pretty(pg_relation_size(indexname::regclass)) as index_size
FROM pg_indexes 
WHERE schemaname = 'public'
ORDER BY tablename, indexname;

-- =============================================================================
-- 2. ANALYZE INDEX USAGE STATISTICS
-- =============================================================================

-- Get index usage statistics (requires pg_stat_user_indexes)
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as times_used,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched,
    pg_size_pretty(pg_relation_size(indexname::regclass)) as index_size
FROM pg_stat_user_indexes 
WHERE schemaname = 'public'
ORDER BY idx_scan DESC;

-- =============================================================================
-- 3. IDENTIFY UNUSED INDEXES
-- =============================================================================

-- Find indexes that are never used
SELECT 
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexname::regclass)) as index_size
FROM pg_stat_user_indexes 
WHERE schemaname = 'public' 
    AND idx_scan = 0
    AND indexname NOT LIKE '%_pkey'  -- Exclude primary keys
ORDER BY pg_relation_size(indexname::regclass) DESC;

-- =============================================================================
-- 4. ANALYZE TABLE SIZES AND ROW COUNTS
-- =============================================================================

-- Get table sizes and row counts
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total_size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) as index_size,
    n_tup_ins as inserts,
    n_tup_upd as updates,
    n_tup_del as deletes,
    n_live_tup as live_tuples,
    n_dead_tup as dead_tuples
FROM pg_stat_user_tables 
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- =============================================================================
-- 5. ANALYZE SLOW QUERIES (if pg_stat_statements is available)
-- =============================================================================

-- Get slowest queries (requires pg_stat_statements extension)
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    rows,
    100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent
FROM pg_stat_statements 
WHERE query LIKE '%classification%' 
    OR query LIKE '%industry%' 
    OR query LIKE '%risk%'
    OR query LIKE '%business%'
ORDER BY mean_time DESC
LIMIT 20;

-- =============================================================================
-- 6. ANALYZE SPECIFIC CLASSIFICATION TABLES
-- =============================================================================

-- Check indexes on classification-related tables
SELECT 
    t.tablename,
    i.indexname,
    i.indexdef,
    pg_size_pretty(pg_relation_size(i.indexname::regclass)) as index_size,
    s.idx_scan as times_used
FROM pg_indexes i
JOIN pg_tables t ON i.tablename = t.tablename
LEFT JOIN pg_stat_user_indexes s ON i.indexname = s.indexname
WHERE i.schemaname = 'public'
    AND (i.tablename LIKE '%classification%' 
         OR i.tablename LIKE '%industry%' 
         OR i.tablename LIKE '%risk%'
         OR i.tablename LIKE '%business%'
         OR i.tablename LIKE '%merchant%'
         OR i.tablename LIKE '%user%')
ORDER BY t.tablename, i.indexname;

-- =============================================================================
-- 7. ANALYZE FOREIGN KEY CONSTRAINTS AND THEIR INDEXES
-- =============================================================================

-- Check foreign key constraints and their associated indexes
SELECT 
    tc.table_name,
    tc.constraint_name,
    tc.constraint_type,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name,
    CASE 
        WHEN i.indexname IS NOT NULL THEN 'Indexed'
        ELSE 'NOT INDEXED'
    END as index_status,
    i.indexname
FROM information_schema.table_constraints AS tc 
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
    AND tc.table_schema = kcu.table_schema
JOIN information_schema.constraint_column_usage AS ccu
    ON ccu.constraint_name = tc.constraint_name
    AND ccu.table_schema = tc.table_schema
LEFT JOIN pg_indexes i ON i.tablename = tc.table_name 
    AND i.indexdef LIKE '%' || kcu.column_name || '%'
WHERE tc.constraint_type = 'FOREIGN KEY' 
    AND tc.table_schema = 'public'
ORDER BY tc.table_name, tc.constraint_name;

-- =============================================================================
-- 8. ANALYZE COMPOSITE INDEX OPPORTUNITIES
-- =============================================================================

-- Check for potential composite index opportunities based on common query patterns
-- This analyzes columns that are frequently used together in WHERE clauses

-- Common query patterns from the codebase analysis:
-- 1. classifications: created_at BETWEEN $1 AND $2 ORDER BY created_at DESC
-- 2. industry_keywords: industry_id + keyword lookups
-- 3. risk_assessments: business_id + risk_level filtering
-- 4. business_classifications: business_id + user_id filtering

-- Check existing composite indexes
SELECT 
    tablename,
    indexname,
    indexdef
FROM pg_indexes 
WHERE schemaname = 'public'
    AND indexdef LIKE '%,%'  -- Contains comma, indicating composite index
ORDER BY tablename, indexname;

-- =============================================================================
-- 9. ANALYZE PARTIAL INDEX OPPORTUNITIES
-- =============================================================================

-- Check for potential partial index opportunities
-- Look for columns with high selectivity that could benefit from partial indexes

-- Example: Active records, recent records, specific status values
SELECT 
    tablename,
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns 
WHERE table_schema = 'public'
    AND (column_name LIKE '%is_active%' 
         OR column_name LIKE '%status%' 
         OR column_name LIKE '%created_at%'
         OR column_name LIKE '%updated_at%')
ORDER BY tablename, column_name;

-- =============================================================================
-- 10. GENERATE INDEX ANALYSIS SUMMARY
-- =============================================================================

-- Summary of current index situation
SELECT 
    'Total Tables' as metric,
    COUNT(*) as value
FROM pg_tables 
WHERE schemaname = 'public'

UNION ALL

SELECT 
    'Total Indexes' as metric,
    COUNT(*) as value
FROM pg_indexes 
WHERE schemaname = 'public'

UNION ALL

SELECT 
    'Unused Indexes' as metric,
    COUNT(*) as value
FROM pg_stat_user_indexes 
WHERE schemaname = 'public' 
    AND idx_scan = 0
    AND indexname NOT LIKE '%_pkey'

UNION ALL

SELECT 
    'Total Index Size' as metric,
    pg_size_pretty(SUM(pg_relation_size(indexname::regclass))) as value
FROM pg_indexes 
WHERE schemaname = 'public';

-- =============================================================================
-- 11. SPECIFIC CLASSIFICATION SYSTEM INDEX ANALYSIS
-- =============================================================================

-- Analyze indexes for the core classification tables
WITH classification_tables AS (
    SELECT unnest(ARRAY[
        'industries',
        'industry_keywords', 
        'classification_codes',
        'code_keywords',
        'risk_keywords',
        'industry_code_crosswalks',
        'business_risk_assessments',
        'business_classifications',
        'risk_assessments',
        'merchants',
        'users'
    ]) as table_name
)
SELECT 
    ct.table_name,
    CASE 
        WHEN t.tablename IS NULL THEN 'TABLE NOT FOUND'
        ELSE 'EXISTS'
    END as table_status,
    COALESCE(i.index_count, 0) as index_count,
    COALESCE(i.index_size, '0 bytes') as total_index_size
FROM classification_tables ct
LEFT JOIN pg_tables t ON ct.table_name = t.tablename AND t.schemaname = 'public'
LEFT JOIN (
    SELECT 
        tablename,
        COUNT(*) as index_count,
        pg_size_pretty(SUM(pg_relation_size(indexname::regclass))) as index_size
    FROM pg_indexes 
    WHERE schemaname = 'public'
    GROUP BY tablename
) i ON ct.table_name = i.tablename
ORDER BY ct.table_name;
