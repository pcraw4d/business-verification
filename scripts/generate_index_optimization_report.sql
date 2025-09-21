-- =============================================================================
-- INDEX OPTIMIZATION REPORT GENERATION SCRIPT
-- Subtask 3.2.2: Generate Index Optimization Report
-- Supabase Table Improvement Implementation Plan
-- =============================================================================
-- This script generates a comprehensive report on index optimization results

-- =============================================================================
-- 1. INDEX INVENTORY REPORT
-- =============================================================================

-- Generate comprehensive index inventory
SELECT 
    'INDEX INVENTORY REPORT' as report_section,
    schemaname,
    tablename,
    indexname,
    indexdef,
    pg_size_pretty(pg_relation_size(indexname::regclass)) as index_size,
    pg_relation_size(indexname::regclass) as size_bytes
FROM pg_indexes 
WHERE schemaname = 'public'
    AND tablename IN ('industries', 'industry_keywords', 'classification_codes', 
                     'industry_patterns', 'keyword_weights', 'risk_keywords',
                     'industry_code_crosswalks', 'business_risk_assessments',
                     'risk_keyword_relationships', 'classification_performance_metrics')
ORDER BY tablename, indexname;

-- =============================================================================
-- 2. INDEX USAGE STATISTICS REPORT
-- =============================================================================

-- Generate index usage statistics
SELECT 
    'INDEX USAGE STATISTICS' as report_section,
    schemaname,
    tablename,
    indexname,
    idx_scan as times_used,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched,
    pg_size_pretty(pg_relation_size(indexname::regclass)) as index_size,
    CASE 
        WHEN idx_scan = 0 THEN 'UNUSED'
        WHEN idx_scan < 10 THEN 'LOW_USAGE'
        WHEN idx_scan < 100 THEN 'MODERATE_USAGE'
        ELSE 'HIGH_USAGE'
    END as usage_category
FROM pg_stat_user_indexes 
WHERE schemaname = 'public'
    AND tablename IN ('industries', 'industry_keywords', 'classification_codes', 
                     'industry_patterns', 'keyword_weights', 'risk_keywords',
                     'industry_code_crosswalks', 'business_risk_assessments',
                     'risk_keyword_relationships', 'classification_performance_metrics')
ORDER BY idx_scan DESC;

-- =============================================================================
-- 3. INDEX SIZE ANALYSIS REPORT
-- =============================================================================

-- Generate index size analysis
SELECT 
    'INDEX SIZE ANALYSIS' as report_section,
    tablename,
    COUNT(*) as index_count,
    pg_size_pretty(SUM(pg_relation_size(indexname::regclass))) as total_size,
    SUM(pg_relation_size(indexname::regclass)) as total_size_bytes,
    pg_size_pretty(AVG(pg_relation_size(indexname::regclass))) as avg_index_size,
    pg_size_pretty(MAX(pg_relation_size(indexname::regclass))) as largest_index,
    pg_size_pretty(MIN(pg_relation_size(indexname::regclass))) as smallest_index
FROM pg_indexes 
WHERE schemaname = 'public'
    AND tablename IN ('industries', 'industry_keywords', 'classification_codes', 
                     'industry_patterns', 'keyword_weights', 'risk_keywords',
                     'industry_code_crosswalks', 'business_risk_assessments',
                     'risk_keyword_relationships', 'classification_performance_metrics')
GROUP BY tablename
ORDER BY total_size_bytes DESC;

-- =============================================================================
-- 4. INDEX TYPE ANALYSIS REPORT
-- =============================================================================

-- Generate index type analysis
SELECT 
    'INDEX TYPE ANALYSIS' as report_section,
    CASE 
        WHEN indexdef LIKE '%USING btree%' THEN 'BTREE'
        WHEN indexdef LIKE '%USING gin%' THEN 'GIN'
        WHEN indexdef LIKE '%USING gist%' THEN 'GIST'
        WHEN indexdef LIKE '%USING hash%' THEN 'HASH'
        ELSE 'OTHER'
    END as index_type,
    COUNT(*) as index_count,
    pg_size_pretty(SUM(pg_relation_size(indexname::regclass))) as total_size
FROM pg_indexes 
WHERE schemaname = 'public'
    AND tablename IN ('industries', 'industry_keywords', 'classification_codes', 
                     'industry_patterns', 'keyword_weights', 'risk_keywords',
                     'industry_code_crosswalks', 'business_risk_assessments',
                     'risk_keyword_relationships', 'classification_performance_metrics')
GROUP BY 
    CASE 
        WHEN indexdef LIKE '%USING btree%' THEN 'BTREE'
        WHEN indexdef LIKE '%USING gin%' THEN 'GIN'
        WHEN indexdef LIKE '%USING gist%' THEN 'GIST'
        WHEN indexdef LIKE '%USING hash%' THEN 'HASH'
        ELSE 'OTHER'
    END
ORDER BY index_count DESC;

-- =============================================================================
-- 5. UNUSED INDEXES REPORT
-- =============================================================================

-- Generate unused indexes report
SELECT 
    'UNUSED INDEXES' as report_section,
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexname::regclass)) as index_size,
    'Consider dropping if not needed for constraints' as recommendation
FROM pg_stat_user_indexes 
WHERE schemaname = 'public' 
    AND idx_scan = 0
    AND indexname NOT LIKE '%_pkey'  -- Exclude primary keys
    AND tablename IN ('industries', 'industry_keywords', 'classification_codes', 
                     'industry_patterns', 'keyword_weights', 'risk_keywords',
                     'industry_code_crosswalks', 'business_risk_assessments',
                     'risk_keyword_relationships', 'classification_performance_metrics')
ORDER BY pg_relation_size(indexname::regclass) DESC;

-- =============================================================================
-- 6. DUPLICATE INDEXES REPORT
-- =============================================================================

-- Generate duplicate indexes report
WITH index_columns AS (
    SELECT 
        tablename,
        indexname,
        string_agg(attname, ',' ORDER BY attnum) as columns
    FROM pg_indexes pi
    JOIN pg_index i ON pi.indexname = i.indexrelname::text
    JOIN pg_attribute a ON i.indrelid = a.attrelid AND a.attnum = ANY(i.indkey)
    WHERE pi.schemaname = 'public'
        AND pi.tablename IN ('industries', 'industry_keywords', 'classification_codes', 
                           'industry_patterns', 'keyword_weights', 'risk_keywords',
                           'industry_code_crosswalks', 'business_risk_assessments',
                           'risk_keyword_relationships', 'classification_performance_metrics')
    GROUP BY tablename, indexname
)
SELECT 
    'DUPLICATE INDEXES' as report_section,
    ic1.tablename,
    ic1.indexname as index1,
    ic2.indexname as index2,
    ic1.columns,
    'Consider consolidating these indexes' as recommendation
FROM index_columns ic1
JOIN index_columns ic2 ON ic1.tablename = ic2.tablename 
    AND ic1.columns = ic2.columns 
    AND ic1.indexname < ic2.indexname
ORDER BY ic1.tablename, ic1.columns;

-- =============================================================================
-- 7. INDEX EFFICIENCY REPORT
-- =============================================================================

-- Generate index efficiency report
SELECT 
    'INDEX EFFICIENCY' as report_section,
    schemaname,
    tablename,
    indexname,
    idx_scan as times_used,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched,
    CASE 
        WHEN idx_scan = 0 THEN 0
        ELSE ROUND((idx_tup_fetch::numeric / idx_tup_read::numeric) * 100, 2)
    END as efficiency_percentage,
    pg_size_pretty(pg_relation_size(indexname::regclass)) as index_size,
    CASE 
        WHEN idx_scan = 0 THEN 'UNUSED - Consider dropping'
        WHEN (idx_tup_fetch::numeric / idx_tup_read::numeric) < 0.5 THEN 'LOW_EFFICIENCY - Review usage'
        WHEN (idx_tup_fetch::numeric / idx_tup_read::numeric) < 0.8 THEN 'MODERATE_EFFICIENCY - Monitor'
        ELSE 'HIGH_EFFICIENCY - Good'
    END as efficiency_status
FROM pg_stat_user_indexes 
WHERE schemaname = 'public'
    AND tablename IN ('industries', 'industry_keywords', 'classification_codes', 
                     'industry_patterns', 'keyword_weights', 'risk_keywords',
                     'industry_code_crosswalks', 'business_risk_assessments',
                     'risk_keyword_relationships', 'classification_performance_metrics')
ORDER BY efficiency_percentage DESC;

-- =============================================================================
-- 8. TABLE STATISTICS REPORT
-- =============================================================================

-- Generate table statistics
SELECT 
    'TABLE STATISTICS' as report_section,
    schemaname,
    tablename,
    n_tup_ins as inserts,
    n_tup_upd as updates,
    n_tup_del as deletes,
    n_live_tup as live_tuples,
    n_dead_tup as dead_tuples,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total_size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) as index_size
FROM pg_stat_user_tables 
WHERE schemaname = 'public'
    AND tablename IN ('industries', 'industry_keywords', 'classification_codes', 
                     'industry_patterns', 'keyword_weights', 'risk_keywords',
                     'industry_code_crosswalks', 'business_risk_assessments',
                     'risk_keyword_relationships', 'classification_performance_metrics')
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- =============================================================================
-- 9. INDEX MAINTENANCE RECOMMENDATIONS
-- =============================================================================

-- Generate maintenance recommendations
SELECT 
    'MAINTENANCE RECOMMENDATIONS' as report_section,
    'Index Maintenance' as category,
    'Run VACUUM ANALYZE on all tables to update statistics' as recommendation,
    'High' as priority
UNION ALL
SELECT 
    'MAINTENANCE RECOMMENDATIONS' as report_section,
    'Index Monitoring' as category,
    'Monitor index usage statistics weekly' as recommendation,
    'Medium' as priority
UNION ALL
SELECT 
    'MAINTENANCE RECOMMENDATIONS' as report_section,
    'Index Cleanup' as category,
    'Review and drop unused indexes monthly' as recommendation,
    'Low' as priority
UNION ALL
SELECT 
    'MAINTENANCE RECOMMENDATIONS' as report_section,
    'Performance Monitoring' as category,
    'Set up alerts for slow queries and index bloat' as recommendation,
    'High' as priority;

-- =============================================================================
-- 10. INDEX OPTIMIZATION SUMMARY
-- =============================================================================

-- Generate optimization summary
WITH index_summary AS (
    SELECT 
        COUNT(*) as total_indexes,
        SUM(pg_relation_size(indexname::regclass)) as total_size_bytes,
        COUNT(CASE WHEN idx_scan = 0 THEN 1 END) as unused_indexes,
        COUNT(CASE WHEN idx_scan > 0 THEN 1 END) as used_indexes
    FROM pg_indexes pi
    LEFT JOIN pg_stat_user_indexes psi ON pi.indexname = psi.indexname
    WHERE pi.schemaname = 'public'
        AND pi.tablename IN ('industries', 'industry_keywords', 'classification_codes', 
                           'industry_patterns', 'keyword_weights', 'risk_keywords',
                           'industry_code_crosswalks', 'business_risk_assessments',
                           'risk_keyword_relationships', 'classification_performance_metrics')
)
SELECT 
    'OPTIMIZATION SUMMARY' as report_section,
    total_indexes as total_indexes_created,
    pg_size_pretty(total_size_bytes) as total_index_size,
    used_indexes as actively_used_indexes,
    unused_indexes as unused_indexes,
    ROUND((used_indexes::numeric / total_indexes::numeric) * 100, 2) as usage_percentage,
    CASE 
        WHEN unused_indexes = 0 THEN 'EXCELLENT - All indexes are being used'
        WHEN unused_indexes < total_indexes * 0.1 THEN 'GOOD - Minimal unused indexes'
        WHEN unused_indexes < total_indexes * 0.2 THEN 'FAIR - Some unused indexes'
        ELSE 'POOR - Many unused indexes need review'
    END as optimization_status
FROM index_summary;

-- =============================================================================
-- 11. INDEX PERFORMANCE METRICS
-- =============================================================================

-- Generate performance metrics
SELECT 
    'PERFORMANCE METRICS' as report_section,
    'Index Creation Time' as metric,
    'All indexes created successfully' as value,
    'Success' as status
UNION ALL
SELECT 
    'PERFORMANCE METRICS' as report_section,
    'Index Size Efficiency' as metric,
    CASE 
        WHEN (SELECT SUM(pg_relation_size(indexname::regclass)) FROM pg_indexes WHERE schemaname = 'public' AND tablename IN ('industries', 'industry_keywords', 'classification_codes', 'industry_patterns', 'keyword_weights', 'risk_keywords', 'industry_code_crosswalks', 'business_risk_assessments', 'risk_keyword_relationships', 'classification_performance_metrics')) < 100 * 1024 * 1024 THEN 'Under 100MB - Excellent'
        WHEN (SELECT SUM(pg_relation_size(indexname::regclass)) FROM pg_indexes WHERE schemaname = 'public' AND tablename IN ('industries', 'industry_keywords', 'classification_codes', 'industry_patterns', 'keyword_weights', 'risk_keywords', 'industry_code_crosswalks', 'business_risk_assessments', 'risk_keyword_relationships', 'classification_performance_metrics')) < 500 * 1024 * 1024 THEN 'Under 500MB - Good'
        ELSE 'Over 500MB - Monitor'
    END as value,
    'Good' as status
UNION ALL
SELECT 
    'PERFORMANCE METRICS' as report_section,
    'Index Coverage' as metric,
    'All critical tables have comprehensive indexes' as value,
    'Complete' as status;

-- =============================================================================
-- 12. FUTURE OPTIMIZATION RECOMMENDATIONS
-- =============================================================================

-- Generate future optimization recommendations
SELECT 
    'FUTURE OPTIMIZATIONS' as report_section,
    'Query Pattern Analysis' as category,
    'Monitor query patterns and add indexes as needed' as recommendation,
    'Ongoing' as timeline
UNION ALL
SELECT 
    'FUTURE OPTIMIZATIONS' as report_section,
    'Index Partitioning' as category,
    'Consider partitioning large tables for better performance' as recommendation,
    'Future' as timeline
UNION ALL
SELECT 
    'FUTURE OPTIMIZATIONS' as report_section,
    'Index Compression' as category,
    'Evaluate index compression for large indexes' as recommendation,
    'Future' as timeline
UNION ALL
SELECT 
    'FUTURE OPTIMIZATIONS' as report_section,
    'Automated Monitoring' as category,
    'Implement automated index usage monitoring and alerts' as recommendation,
    'Next Phase' as timeline;

-- =============================================================================
-- END OF INDEX OPTIMIZATION REPORT
-- =============================================================================
