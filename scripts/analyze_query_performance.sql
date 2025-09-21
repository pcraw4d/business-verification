-- =============================================================================
-- QUERY PERFORMANCE ANALYSIS SCRIPT
-- Subtask 3.2.1: Analyze Query Performance Patterns and Bottlenecks
-- =============================================================================
-- This script analyzes query performance patterns based on codebase analysis
-- and identifies potential bottlenecks in the classification system

-- =============================================================================
-- 1. COMMON QUERY PATTERNS FROM CODEBASE ANALYSIS
-- =============================================================================

-- Based on the codebase analysis, here are the most common query patterns:

-- Pattern 1: Time-based classification queries (from metrics_collector.go)
-- Query: SELECT * FROM classifications WHERE created_at BETWEEN $1 AND $2 ORDER BY created_at DESC
-- Bottleneck: Sequential scan on created_at without proper index
-- Solution: Composite index on (created_at, id) for efficient time-range queries

-- Pattern 2: Industry-based classification queries (from dimension_collectors.go)
-- Query: SELECT * FROM classifications WHERE created_at BETWEEN $1 AND $2 ORDER BY actual_classification, created_at DESC
-- Bottleneck: Sorting by actual_classification without index
-- Solution: Composite index on (actual_classification, created_at)

-- Pattern 3: Business classification lookups
-- Query: SELECT * FROM business_classifications WHERE business_id = $1
-- Bottleneck: Missing index on business_id
-- Solution: Index on business_id

-- Pattern 4: Risk assessment queries
-- Query: SELECT * FROM business_risk_assessments WHERE business_id = $1 AND risk_level IN ('high', 'critical')
-- Bottleneck: Missing composite index on (business_id, risk_level)
-- Solution: Composite index on (business_id, risk_level)

-- Pattern 5: Industry keyword lookups
-- Query: SELECT * FROM industry_keywords WHERE industry_id = $1 AND is_primary = true
-- Bottleneck: Missing composite index on (industry_id, is_primary)
-- Solution: Composite index on (industry_id, is_primary)

-- =============================================================================
-- 2. PERFORMANCE BOTTLENECK ANALYSIS
-- =============================================================================

-- Analyze table sizes and identify potential performance issues
WITH table_analysis AS (
    SELECT 
        schemaname,
        tablename,
        pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total_size,
        pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size,
        pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) as index_size,
        n_live_tup as live_tuples,
        n_dead_tup as dead_tuples,
        CASE 
            WHEN n_dead_tup > n_live_tup THEN 'HIGH DEAD TUPLE RATIO'
            WHEN n_dead_tup > n_live_tup * 0.1 THEN 'MODERATE DEAD TUPLE RATIO'
            ELSE 'LOW DEAD TUPLE RATIO'
        END as dead_tuple_analysis
    FROM pg_stat_user_tables 
    WHERE schemaname = 'public'
)
SELECT 
    tablename,
    total_size,
    table_size,
    index_size,
    live_tuples,
    dead_tuples,
    dead_tuple_analysis,
    CASE 
        WHEN live_tuples > 1000000 THEN 'LARGE TABLE - NEEDS OPTIMIZATION'
        WHEN live_tuples > 100000 THEN 'MEDIUM TABLE - MONITOR PERFORMANCE'
        ELSE 'SMALL TABLE - LOW PRIORITY'
    END as optimization_priority
FROM table_analysis
ORDER BY live_tuples DESC;

-- =============================================================================
-- 3. INDEX EFFICIENCY ANALYSIS
-- =============================================================================

-- Analyze index usage and efficiency
WITH index_analysis AS (
    SELECT 
        schemaname,
        tablename,
        indexname,
        idx_scan as times_used,
        idx_tup_read as tuples_read,
        idx_tup_fetch as tuples_fetched,
        pg_size_pretty(pg_relation_size(indexname::regclass)) as index_size,
        CASE 
            WHEN idx_scan = 0 THEN 'UNUSED INDEX'
            WHEN idx_scan < 10 THEN 'RARELY USED INDEX'
            WHEN idx_scan < 100 THEN 'MODERATELY USED INDEX'
            ELSE 'FREQUENTLY USED INDEX'
        END as usage_category,
        CASE 
            WHEN idx_tup_fetch > 0 THEN ROUND((idx_tup_fetch::numeric / idx_tup_read::numeric) * 100, 2)
            ELSE 0
        END as fetch_efficiency_percent
    FROM pg_stat_user_indexes 
    WHERE schemaname = 'public'
)
SELECT 
    tablename,
    indexname,
    times_used,
    tuples_read,
    tuples_fetched,
    index_size,
    usage_category,
    fetch_efficiency_percent,
    CASE 
        WHEN usage_category = 'UNUSED INDEX' AND pg_relation_size(indexname::regclass) > 1024*1024 THEN 'CANDIDATE FOR REMOVAL'
        WHEN usage_category = 'RARELY USED INDEX' AND pg_relation_size(indexname::regclass) > 10*1024*1024 THEN 'CANDIDATE FOR REVIEW'
        WHEN fetch_efficiency_percent < 50 AND times_used > 100 THEN 'INEFFICIENT INDEX'
        ELSE 'EFFICIENT INDEX'
    END as optimization_recommendation
FROM index_analysis
ORDER BY 
    CASE usage_category
        WHEN 'UNUSED INDEX' THEN 1
        WHEN 'RARELY USED INDEX' THEN 2
        WHEN 'MODERATELY USED INDEX' THEN 3
        ELSE 4
    END,
    pg_relation_size(indexname::regclass) DESC;

-- =============================================================================
-- 4. QUERY PATTERN ANALYSIS FOR CLASSIFICATION SYSTEM
-- =============================================================================

-- Analyze common query patterns and their performance implications
CREATE OR REPLACE VIEW query_pattern_analysis AS
SELECT 
    'Time-based Classification Queries' as query_pattern,
    'SELECT * FROM classifications WHERE created_at BETWEEN $1 AND $2 ORDER BY created_at DESC' as example_query,
    'created_at' as primary_filter_column,
    'created_at, id' as recommended_index,
    'HIGH' as priority,
    'Critical for monitoring and analytics performance' as impact

UNION ALL

SELECT 
    'Industry-based Classification Queries' as query_pattern,
    'SELECT * FROM classifications WHERE created_at BETWEEN $1 AND $2 ORDER BY actual_classification, created_at DESC' as example_query,
    'actual_classification, created_at' as primary_filter_column,
    'actual_classification, created_at, id' as recommended_index,
    'HIGH' as priority,
    'Critical for industry-specific analytics' as impact

UNION ALL

SELECT 
    'Business Classification Lookups' as query_pattern,
    'SELECT * FROM business_classifications WHERE business_id = $1' as example_query,
    'business_id' as primary_filter_column,
    'business_id' as recommended_index,
    'HIGH' as priority,
    'Critical for business-specific queries' as impact

UNION ALL

SELECT 
    'Risk Assessment Queries' as query_pattern,
    'SELECT * FROM business_risk_assessments WHERE business_id = $1 AND risk_level IN (''high'', ''critical'')' as example_query,
    'business_id, risk_level' as primary_filter_column,
    'business_id, risk_level, assessment_date' as recommended_index,
    'HIGH' as priority,
    'Critical for risk monitoring and alerts' as impact

UNION ALL

SELECT 
    'Industry Keyword Lookups' as query_pattern,
    'SELECT * FROM industry_keywords WHERE industry_id = $1 AND is_primary = true' as example_query,
    'industry_id, is_primary' as primary_filter_column,
    'industry_id, is_primary, weight' as recommended_index,
    'MEDIUM' as priority,
    'Important for classification accuracy' as impact

UNION ALL

SELECT 
    'Risk Keyword Searches' as query_pattern,
    'SELECT * FROM risk_keywords WHERE keyword ILIKE ''%search_term%'' AND is_active = true' as example_query,
    'keyword, is_active' as primary_filter_column,
    'keyword, is_active, risk_severity' as recommended_index,
    'MEDIUM' as priority,
    'Important for risk detection performance' as impact

UNION ALL

SELECT 
    'Code Crosswalk Queries' as query_pattern,
    'SELECT * FROM industry_code_crosswalks WHERE mcc_code = $1 OR naics_code = $2 OR sic_code = $3' as example_query,
    'mcc_code, naics_code, sic_code' as primary_filter_column,
    'mcc_code, naics_code, sic_code, industry_id' as recommended_index,
    'MEDIUM' as priority,
    'Important for code mapping performance' as impact;

-- Query the analysis
SELECT * FROM query_pattern_analysis ORDER BY priority DESC, query_pattern;

-- =============================================================================
-- 5. PERFORMANCE BOTTLENECK IDENTIFICATION
-- =============================================================================

-- Identify specific performance bottlenecks
WITH bottleneck_analysis AS (
    SELECT 
        'Missing Indexes' as bottleneck_type,
        'High' as severity,
        'Classification queries without proper indexes' as description,
        'Add composite indexes for common query patterns' as solution,
        '1-2 days' as estimated_fix_time

    UNION ALL

    SELECT 
        'Unused Indexes' as bottleneck_type,
        'Medium' as severity,
        'Large indexes that are never used' as description,
        'Remove or optimize unused indexes' as solution,
        '1 day' as estimated_fix_time

    UNION ALL

    SELECT 
        'Dead Tuples' as bottleneck_type,
        'Medium' as severity,
        'High ratio of dead tuples affecting performance' as description,
        'Run VACUUM and consider more frequent maintenance' as solution,
        '2-4 hours' as estimated_fix_time

    UNION ALL

    SELECT 
        'Large Table Scans' as bottleneck_type,
        'High' as severity,
        'Sequential scans on large tables' as description,
        'Add appropriate indexes and optimize queries' as solution,
        '2-3 days' as estimated_fix_time

    UNION ALL

    SELECT 
        'JSONB Query Performance' as bottleneck_type,
        'Medium' as severity,
        'Slow queries on JSONB columns' as description,
        'Add GIN indexes for JSONB columns' as solution,
        '1 day' as estimated_fix_time

    UNION ALL

    SELECT 
        'Array Column Performance' as bottleneck_type,
        'Medium' as severity,
        'Slow queries on array columns' as description,
        'Add GIN indexes for array columns' as solution,
        '1 day' as estimated_fix_time
)
SELECT * FROM bottleneck_analysis ORDER BY 
    CASE severity
        WHEN 'High' THEN 1
        WHEN 'Medium' THEN 2
        ELSE 3
    END,
    bottleneck_type;

-- =============================================================================
-- 6. QUERY OPTIMIZATION RECOMMENDATIONS
-- =============================================================================

-- Generate specific optimization recommendations
CREATE OR REPLACE VIEW optimization_recommendations AS
SELECT 
    'Immediate Actions (High Priority)' as category,
    ARRAY[
        'Add composite index on classifications(created_at, id)',
        'Add composite index on classifications(actual_classification, created_at)',
        'Add index on business_classifications(business_id)',
        'Add composite index on business_risk_assessments(business_id, risk_level)',
        'Add index on industry_keywords(industry_id, is_primary)'
    ] as recommendations,
    'Critical for system performance' as impact

UNION ALL

SELECT 
    'Short-term Actions (Medium Priority)' as category,
    ARRAY[
        'Add GIN indexes for JSONB columns (metadata, address, contact_info)',
        'Add GIN indexes for array columns (mcc_codes, naics_codes, synonyms)',
        'Add full-text search indexes for business names and descriptions',
        'Add partial indexes for high-selectivity queries',
        'Optimize unused indexes'
    ] as recommendations,
    'Important for advanced query performance' as impact

UNION ALL

SELECT 
    'Long-term Actions (Low Priority)' as category,
    ARRAY[
        'Implement query result caching',
        'Consider table partitioning for large tables',
        'Implement connection pooling optimization',
        'Add query performance monitoring',
        'Implement automated index maintenance'
    ] as recommendations,
    'Enhance overall system scalability' as impact;

-- Query the recommendations
SELECT * FROM optimization_recommendations ORDER BY 
    CASE category
        WHEN 'Immediate Actions (High Priority)' THEN 1
        WHEN 'Short-term Actions (Medium Priority)' THEN 2
        ELSE 3
    END;

-- =============================================================================
-- 7. PERFORMANCE MONITORING QUERIES
-- =============================================================================

-- Queries to monitor performance after optimization
CREATE OR REPLACE VIEW performance_monitoring_queries AS
SELECT 
    'Index Usage Monitoring' as monitoring_type,
    'SELECT tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch FROM pg_stat_user_indexes WHERE schemaname = ''public'' ORDER BY idx_scan DESC;' as monitoring_query,
    'Daily' as frequency,
    'Monitor index usage to identify unused or inefficient indexes' as purpose

UNION ALL

SELECT 
    'Table Size Monitoring' as monitoring_type,
    'SELECT tablename, pg_size_pretty(pg_total_relation_size(tablename::regclass)) as size, n_live_tup, n_dead_tup FROM pg_stat_user_tables WHERE schemaname = ''public'' ORDER BY pg_total_relation_size(tablename::regclass) DESC;' as monitoring_query,
    'Weekly' as frequency,
    'Monitor table growth and dead tuple accumulation' as purpose

UNION ALL

SELECT 
    'Slow Query Monitoring' as monitoring_type,
    'SELECT query, calls, total_time, mean_time FROM pg_stat_statements WHERE mean_time > 100 ORDER BY mean_time DESC LIMIT 20;' as monitoring_query,
    'Daily' as frequency,
    'Identify slow queries that need optimization' as purpose

UNION ALL

SELECT 
    'Cache Hit Ratio Monitoring' as monitoring_type,
    'SELECT datname, round(100.0 * blks_hit / (blks_hit + blks_read), 2) AS cache_hit_ratio FROM pg_stat_database WHERE datname = current_database();' as monitoring_query,
    'Daily' as frequency,
    'Monitor database cache efficiency' as purpose;

-- Query the monitoring recommendations
SELECT * FROM performance_monitoring_queries ORDER BY frequency, monitoring_type;
