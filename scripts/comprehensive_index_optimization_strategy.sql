-- =============================================================================
-- COMPREHENSIVE INDEX OPTIMIZATION STRATEGY
-- Subtask 3.2.1: Plan Comprehensive Index Optimization Strategy
-- =============================================================================
-- This script provides a comprehensive strategy for optimizing database indexes
-- based on the analysis of current state, missing indexes, and performance patterns

-- =============================================================================
-- 1. OPTIMIZATION STRATEGY OVERVIEW
-- =============================================================================

-- The optimization strategy is divided into three phases:
-- Phase 1: Critical Performance Fixes (Immediate - High Priority)
-- Phase 2: Advanced Optimization (Short-term - Medium Priority)  
-- Phase 3: Long-term Scalability (Long-term - Low Priority)

-- =============================================================================
-- 2. PHASE 1: CRITICAL PERFORMANCE FIXES (IMMEDIATE)
-- =============================================================================

-- These indexes are critical for system performance and should be implemented first

-- 2.1 Core Classification System Indexes
-- =============================================================================

-- Critical index for time-based classification queries (most common pattern)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_classifications_created_at_id 
ON classifications(created_at DESC, id DESC);

-- Critical index for industry-based classification queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_classifications_industry_created 
ON classifications(actual_classification, created_at DESC, id DESC);

-- Critical index for business classification lookups
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_business_classifications_business_id 
ON business_classifications(business_id);

-- Critical index for business classification by user
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_business_classifications_user_id 
ON business_classifications(user_id);

-- 2.2 Risk Assessment System Indexes
-- =============================================================================

-- Critical index for business risk assessments
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_business_risk_assessments_business_id 
ON business_risk_assessments(business_id);

-- Critical composite index for risk level filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_business_risk_assessments_business_risk 
ON business_risk_assessments(business_id, risk_level);

-- Critical index for risk assessment date filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_business_risk_assessments_assessment_date 
ON business_risk_assessments(assessment_date DESC);

-- 2.3 Industry and Keyword System Indexes
-- =============================================================================

-- Critical index for industry keyword lookups
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_industry_keywords_industry_id 
ON industry_keywords(industry_id);

-- Critical composite index for primary keyword lookups
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_industry_keywords_industry_primary 
ON industry_keywords(industry_id, is_primary);

-- Critical index for keyword searches
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_industry_keywords_keyword 
ON industry_keywords(keyword);

-- 2.4 Risk Keywords System Indexes
-- =============================================================================

-- Critical index for risk keyword lookups
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_keywords_keyword 
ON risk_keywords(keyword);

-- Critical index for risk category filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_keywords_risk_category 
ON risk_keywords(risk_category);

-- Critical index for risk severity filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_keywords_risk_severity 
ON risk_keywords(risk_severity);

-- Critical composite index for active risk keywords
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_keywords_active_category 
ON risk_keywords(is_active, risk_category);

-- =============================================================================
-- 3. PHASE 2: ADVANCED OPTIMIZATION (SHORT-TERM)
-- =============================================================================

-- These optimizations provide advanced query capabilities and performance improvements

-- 3.1 Composite Indexes for Complex Queries
-- =============================================================================

-- Advanced composite indexes for classification system
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_classifications_method_created 
ON classifications(classification_method, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_classifications_confidence_created 
ON classifications(confidence_score DESC, created_at DESC);

-- Advanced composite indexes for business classifications
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_business_classifications_industry_confidence 
ON business_classifications(industry, confidence_score DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_business_classifications_method_created 
ON business_classifications(classification_method, created_at DESC);

-- Advanced composite indexes for risk assessments
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_business_risk_assessments_risk_level_score 
ON business_risk_assessments(risk_level, risk_score DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_business_risk_assessments_date_risk 
ON business_risk_assessments(assessment_date DESC, risk_level);

-- 3.2 Partial Indexes for High-Selectivity Queries
-- =============================================================================

-- Partial index for high-risk assessments only
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_business_risk_assessments_high_risk 
ON business_risk_assessments(business_id, assessment_date DESC) 
WHERE risk_level IN ('high', 'critical');

-- Partial index for critical risk keywords only
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_keywords_critical_risks 
ON risk_keywords(keyword, risk_category) 
WHERE risk_severity = 'critical' AND is_active = true;

-- Partial index for illegal activities only
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_keywords_illegal_activities 
ON risk_keywords(keyword, risk_severity) 
WHERE risk_category = 'illegal' AND is_active = true;

-- Partial index for primary industry keywords only
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_industry_keywords_primary_only 
ON industry_keywords(industry_id, keyword, weight) 
WHERE is_primary = true AND is_active = true;

-- 3.3 GIN Indexes for JSONB and Array Columns
-- =============================================================================

-- GIN indexes for JSONB columns
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_metadata_gin 
ON users USING GIN (metadata);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_merchants_address_gin 
ON merchants USING GIN (address);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_merchants_contact_info_gin 
ON merchants USING GIN (contact_info);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_merchants_metadata_gin 
ON merchants USING GIN (metadata);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_business_risk_assessments_detected_patterns_gin 
ON business_risk_assessments USING GIN (detected_patterns);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_business_risk_assessments_assessment_metadata_gin 
ON business_risk_assessments USING GIN (assessment_metadata);

-- GIN indexes for array columns
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_keywords_mcc_codes_gin 
ON risk_keywords USING GIN (mcc_codes);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_keywords_naics_codes_gin 
ON risk_keywords USING GIN (naics_codes);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_keywords_sic_codes_gin 
ON risk_keywords USING GIN (sic_codes);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_keywords_card_brand_restrictions_gin 
ON risk_keywords USING GIN (card_brand_restrictions);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_keywords_detection_patterns_gin 
ON risk_keywords USING GIN (detection_patterns);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_keywords_synonyms_gin 
ON risk_keywords USING GIN (synonyms);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_business_risk_assessments_detected_keywords_gin 
ON business_risk_assessments USING GIN (detected_keywords);

-- 3.4 Full-Text Search Indexes
-- =============================================================================

-- Full-text search indexes for business names and descriptions
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_merchants_name_fts 
ON merchants USING GIN (to_tsvector('english', name));

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_merchants_description_fts 
ON merchants USING GIN (to_tsvector('english', description));

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_industries_name_fts 
ON industries USING GIN (to_tsvector('english', name));

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_industries_description_fts 
ON industries USING GIN (to_tsvector('english', description));

-- =============================================================================
-- 4. PHASE 3: LONG-TERM SCALABILITY (LONG-TERM)
-- =============================================================================

-- These optimizations focus on long-term scalability and advanced features

-- 4.1 Advanced Composite Indexes
-- =============================================================================

-- Advanced composite indexes for complex analytics queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_classifications_industry_method_confidence 
ON classifications(actual_classification, classification_method, confidence_score DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_business_classifications_user_industry_created 
ON business_classifications(user_id, industry, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_business_risk_assessments_business_date_risk_score 
ON business_risk_assessments(business_id, assessment_date DESC, risk_level, risk_score DESC);

-- 4.2 Code Crosswalk Optimization Indexes
-- =============================================================================

-- Indexes for industry code crosswalks
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_industry_code_crosswalks_industry_id 
ON industry_code_crosswalks(industry_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_industry_code_crosswalks_mcc_code 
ON industry_code_crosswalks(mcc_code);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_industry_code_crosswalks_naics_code 
ON industry_code_crosswalks(naics_code);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_industry_code_crosswalks_sic_code 
ON industry_code_crosswalks(sic_code);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_industry_code_crosswalks_industry_active 
ON industry_code_crosswalks(industry_id, is_active);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_industry_code_crosswalks_industry_primary 
ON industry_code_crosswalks(industry_id, is_primary);

-- 4.3 Monitoring and Audit Indexes
-- =============================================================================

-- Enhanced indexes for audit and monitoring
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_resource_type_id 
ON audit_logs(resource_type, resource_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_user_event_date 
ON audit_logs(user_id, event_type, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_external_service_calls_service_status 
ON external_service_calls(service_name, status);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_external_service_calls_user_service_date 
ON external_service_calls(user_id, service_name, created_at DESC);

-- =============================================================================
-- 5. INDEX MAINTENANCE STRATEGY
-- =============================================================================

-- 5.1 Index Maintenance Procedures
-- =============================================================================

-- Create a function to analyze and maintain indexes
CREATE OR REPLACE FUNCTION maintain_indexes()
RETURNS TABLE(
    table_name text,
    index_name text,
    index_size text,
    usage_count bigint,
    maintenance_action text
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        t.tablename::text,
        i.indexname::text,
        pg_size_pretty(pg_relation_size(i.indexname::regclass))::text,
        COALESCE(s.idx_scan, 0)::bigint,
        CASE 
            WHEN COALESCE(s.idx_scan, 0) = 0 AND pg_relation_size(i.indexname::regclass) > 1024*1024 THEN 'CANDIDATE FOR REMOVAL'
            WHEN COALESCE(s.idx_scan, 0) < 10 AND pg_relation_size(i.indexname::regclass) > 10*1024*1024 THEN 'REVIEW USAGE'
            WHEN COALESCE(s.idx_scan, 0) > 1000 THEN 'MONITOR PERFORMANCE'
            ELSE 'NO ACTION NEEDED'
        END::text
    FROM pg_indexes i
    JOIN pg_tables t ON i.tablename = t.tablename
    LEFT JOIN pg_stat_user_indexes s ON i.indexname = s.indexname
    WHERE i.schemaname = 'public'
    ORDER BY pg_relation_size(i.indexname::regclass) DESC;
END;
$$ LANGUAGE plpgsql;

-- 5.2 Index Statistics Update Function
-- =============================================================================

-- Create a function to update index statistics
CREATE OR REPLACE FUNCTION update_index_statistics()
RETURNS void AS $$
BEGIN
    -- Update statistics for all tables
    ANALYZE;
    
    -- Log the update
    INSERT INTO audit_logs (
        user_id,
        event_type,
        resource_type,
        resource_id,
        details
    ) VALUES (
        NULL,
        'system_maintenance',
        'database',
        'index_statistics',
        '{"action": "analyze", "timestamp": "' || NOW() || '"}'
    );
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- 6. PERFORMANCE MONITORING QUERIES
-- =============================================================================

-- 6.1 Index Performance Monitoring
-- =============================================================================

-- Create a view for monitoring index performance
CREATE OR REPLACE VIEW index_performance_monitoring AS
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as times_used,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched,
    pg_size_pretty(pg_relation_size(indexname::regclass)) as index_size,
    CASE 
        WHEN idx_scan = 0 THEN 'UNUSED'
        WHEN idx_scan < 10 THEN 'RARELY_USED'
        WHEN idx_scan < 100 THEN 'MODERATELY_USED'
        ELSE 'FREQUENTLY_USED'
    END as usage_category,
    CASE 
        WHEN idx_tup_fetch > 0 THEN ROUND((idx_tup_fetch::numeric / idx_tup_read::numeric) * 100, 2)
        ELSE 0
    END as fetch_efficiency_percent
FROM pg_stat_user_indexes 
WHERE schemaname = 'public'
ORDER BY idx_scan DESC;

-- 6.2 Query Performance Monitoring
-- =============================================================================

-- Create a view for monitoring query performance
CREATE OR REPLACE VIEW query_performance_monitoring AS
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
ORDER BY mean_time DESC;

-- =============================================================================
-- 7. IMPLEMENTATION TIMELINE AND PRIORITIES
-- =============================================================================

-- Create a view for implementation planning
CREATE OR REPLACE VIEW implementation_plan AS
SELECT 
    'Phase 1: Critical Performance Fixes' as phase,
    'Immediate (Week 1)' as timeline,
    'HIGH' as priority,
    ARRAY[
        'idx_classifications_created_at_id',
        'idx_classifications_industry_created',
        'idx_business_classifications_business_id',
        'idx_business_risk_assessments_business_id',
        'idx_business_risk_assessments_business_risk',
        'idx_industry_keywords_industry_id',
        'idx_risk_keywords_keyword',
        'idx_risk_keywords_risk_category'
    ] as indexes_to_create,
    'Critical for system performance and user experience' as impact

UNION ALL

SELECT 
    'Phase 2: Advanced Optimization' as phase,
    'Short-term (Week 2-3)' as timeline,
    'MEDIUM' as priority,
    ARRAY[
        'Composite indexes for complex queries',
        'Partial indexes for high-selectivity queries',
        'GIN indexes for JSONB and array columns',
        'Full-text search indexes'
    ] as indexes_to_create,
    'Important for advanced query capabilities and performance' as impact

UNION ALL

SELECT 
    'Phase 3: Long-term Scalability' as phase,
    'Long-term (Week 4-6)' as timeline,
    'LOW' as priority,
    ARRAY[
        'Advanced composite indexes',
        'Code crosswalk optimization indexes',
        'Enhanced monitoring and audit indexes',
        'Index maintenance automation'
    ] as indexes_to_create,
    'Enhances long-term scalability and maintainability' as impact;

-- Query the implementation plan
SELECT * FROM implementation_plan ORDER BY 
    CASE priority
        WHEN 'HIGH' THEN 1
        WHEN 'MEDIUM' THEN 2
        ELSE 3
    END;

-- =============================================================================
-- 8. SUCCESS METRICS AND VALIDATION
-- =============================================================================

-- Create a view for success metrics
CREATE OR REPLACE VIEW optimization_success_metrics AS
SELECT 
    'Query Performance Improvement' as metric,
    'Average query response time < 100ms' as target,
    'Measure before and after optimization' as measurement_method,
    'Critical for user experience' as importance

UNION ALL

SELECT 
    'Index Usage Efficiency' as metric,
    '> 80% of indexes actively used' as target,
    'Monitor pg_stat_user_indexes' as measurement_method,
    'Important for resource optimization' as importance

UNION ALL

SELECT 
    'Cache Hit Ratio' as metric,
    '> 95% cache hit ratio' as target,
    'Monitor pg_stat_database' as measurement_method,
    'Critical for overall performance' as importance

UNION ALL

SELECT 
    'Dead Tuple Ratio' as metric,
    '< 10% dead tuple ratio' as target,
    'Monitor pg_stat_user_tables' as measurement_method,
    'Important for maintenance efficiency' as importance;

-- Query the success metrics
SELECT * FROM optimization_success_metrics ORDER BY importance DESC;
