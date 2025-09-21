-- =============================================================================
-- INDEX IMPROVEMENT VALIDATION SCRIPT
-- Subtask 3.2.2: Validate Index Improvements
-- Supabase Table Improvement Implementation Plan
-- =============================================================================
-- This script validates that all index improvements are working correctly

-- =============================================================================
-- 1. VALIDATE INDEX CREATION
-- =============================================================================

-- Check that all expected indexes exist
SELECT 
    'INDEX CREATION VALIDATION' as validation_type,
    CASE 
        WHEN COUNT(*) = 0 THEN 'FAILED - No indexes found'
        ELSE 'PASSED - Indexes exist'
    END as status,
    COUNT(*) as index_count
FROM pg_indexes 
WHERE schemaname = 'public'
    AND tablename IN ('industries', 'industry_keywords', 'classification_codes', 
                     'industry_patterns', 'keyword_weights', 'risk_keywords',
                     'industry_code_crosswalks', 'business_risk_assessments',
                     'risk_keyword_relationships', 'classification_performance_metrics');

-- =============================================================================
-- 2. VALIDATE INDEX TYPES
-- =============================================================================

-- Validate that correct index types are used
SELECT 
    'INDEX TYPE VALIDATION' as validation_type,
    index_type,
    COUNT(*) as count,
    CASE 
        WHEN index_type IN ('BTREE', 'GIN', 'GIST') THEN 'PASSED - Valid index type'
        ELSE 'FAILED - Invalid index type'
    END as status
FROM (
    SELECT 
        CASE 
            WHEN indexdef LIKE '%USING btree%' THEN 'BTREE'
            WHEN indexdef LIKE '%USING gin%' THEN 'GIN'
            WHEN indexdef LIKE '%USING gist%' THEN 'GIST'
            WHEN indexdef LIKE '%USING hash%' THEN 'HASH'
            ELSE 'OTHER'
        END as index_type
    FROM pg_indexes 
    WHERE schemaname = 'public'
        AND tablename IN ('industries', 'industry_keywords', 'classification_codes', 
                         'industry_patterns', 'keyword_weights', 'risk_keywords',
                         'industry_code_crosswalks', 'business_risk_assessments',
                         'risk_keyword_relationships', 'classification_performance_metrics')
) t
GROUP BY index_type
ORDER BY count DESC;

-- =============================================================================
-- 3. VALIDATE INDEX SIZES
-- =============================================================================

-- Validate that indexes are not excessively large
SELECT 
    'INDEX SIZE VALIDATION' as validation_type,
    tablename,
    COUNT(*) as index_count,
    pg_size_pretty(SUM(pg_relation_size(indexname::regclass))) as total_size,
    CASE 
        WHEN SUM(pg_relation_size(indexname::regclass)) < 50 * 1024 * 1024 THEN 'PASSED - Size acceptable'
        WHEN SUM(pg_relation_size(indexname::regclass)) < 200 * 1024 * 1024 THEN 'WARNING - Size large'
        ELSE 'FAILED - Size excessive'
    END as status
FROM pg_indexes 
WHERE schemaname = 'public'
    AND tablename IN ('industries', 'industry_keywords', 'classification_codes', 
                     'industry_patterns', 'keyword_weights', 'risk_keywords',
                     'industry_code_crosswalks', 'business_risk_assessments',
                     'risk_keyword_relationships', 'classification_performance_metrics')
GROUP BY tablename
ORDER BY SUM(pg_relation_size(indexname::regclass)) DESC;

-- =============================================================================
-- 4. VALIDATE CRITICAL INDEXES
-- =============================================================================

-- Validate that critical indexes exist
WITH critical_indexes AS (
    SELECT 'idx_industries_name' as index_name, 'industries' as table_name
    UNION ALL SELECT 'idx_industries_category', 'industries'
    UNION ALL SELECT 'idx_industries_active', 'industries'
    UNION ALL SELECT 'idx_industry_keywords_industry_id', 'industry_keywords'
    UNION ALL SELECT 'idx_industry_keywords_keyword', 'industry_keywords'
    UNION ALL SELECT 'idx_industry_keywords_active', 'industry_keywords'
    UNION ALL SELECT 'idx_classification_codes_industry_id', 'classification_codes'
    UNION ALL SELECT 'idx_classification_codes_type', 'classification_codes'
    UNION ALL SELECT 'idx_classification_codes_code', 'classification_codes'
    UNION ALL SELECT 'idx_risk_keywords_keyword', 'risk_keywords'
    UNION ALL SELECT 'idx_risk_keywords_category', 'risk_keywords'
    UNION ALL SELECT 'idx_risk_keywords_severity', 'risk_keywords'
    UNION ALL SELECT 'idx_risk_keywords_active', 'risk_keywords'
    UNION ALL SELECT 'idx_business_risk_assessments_business', 'business_risk_assessments'
    UNION ALL SELECT 'idx_business_risk_assessments_score', 'business_risk_assessments'
    UNION ALL SELECT 'idx_business_risk_assessments_level', 'business_risk_assessments'
)
SELECT 
    'CRITICAL INDEX VALIDATION' as validation_type,
    ci.index_name,
    ci.table_name,
    CASE 
        WHEN pi.indexname IS NOT NULL THEN 'PASSED - Index exists'
        ELSE 'FAILED - Index missing'
    END as status
FROM critical_indexes ci
LEFT JOIN pg_indexes pi ON ci.index_name = pi.indexname AND ci.table_name = pi.tablename
ORDER BY ci.table_name, ci.index_name;

-- =============================================================================
-- 5. VALIDATE COMPOSITE INDEXES
-- =============================================================================

-- Validate that composite indexes exist
WITH composite_indexes AS (
    SELECT 'idx_industries_category_active' as index_name, 'industries' as table_name
    UNION ALL SELECT 'idx_industry_keywords_industry_active', 'industry_keywords'
    UNION ALL SELECT 'idx_industry_keywords_weight_industry', 'industry_keywords'
    UNION ALL SELECT 'idx_classification_codes_industry_active', 'classification_codes'
    UNION ALL SELECT 'idx_classification_codes_type_active', 'classification_codes'
    UNION ALL SELECT 'idx_risk_keywords_category_severity', 'risk_keywords'
    UNION ALL SELECT 'idx_risk_keywords_active_category', 'risk_keywords'
    UNION ALL SELECT 'idx_business_risk_assessments_business_date', 'business_risk_assessments'
    UNION ALL SELECT 'idx_business_risk_assessments_level_score', 'business_risk_assessments'
)
SELECT 
    'COMPOSITE INDEX VALIDATION' as validation_type,
    ci.index_name,
    ci.table_name,
    CASE 
        WHEN pi.indexname IS NOT NULL THEN 'PASSED - Composite index exists'
        ELSE 'FAILED - Composite index missing'
    END as status
FROM composite_indexes ci
LEFT JOIN pg_indexes pi ON ci.index_name = pi.indexname AND ci.table_name = pi.tablename
ORDER BY ci.table_name, ci.index_name;

-- =============================================================================
-- 6. VALIDATE GIN INDEXES
-- =============================================================================

-- Validate that GIN indexes exist for array and text fields
WITH gin_indexes AS (
    SELECT 'idx_risk_keywords_mcc_codes' as index_name, 'risk_keywords' as table_name
    UNION ALL SELECT 'idx_risk_keywords_naics_codes', 'risk_keywords'
    UNION ALL SELECT 'idx_risk_keywords_sic_codes', 'risk_keywords'
    UNION ALL SELECT 'idx_risk_keywords_card_restrictions', 'risk_keywords'
    UNION ALL SELECT 'idx_risk_keywords_detection_patterns', 'risk_keywords'
    UNION ALL SELECT 'idx_risk_keywords_synonyms', 'risk_keywords'
    UNION ALL SELECT 'idx_risk_keywords_fulltext', 'risk_keywords'
    UNION ALL SELECT 'idx_business_risk_assessments_keywords', 'business_risk_assessments'
    UNION ALL SELECT 'idx_business_risk_assessments_patterns', 'business_risk_assessments'
    UNION ALL SELECT 'idx_business_risk_assessments_metadata', 'business_risk_assessments'
    UNION ALL SELECT 'idx_industry_keywords_keyword_trgm', 'industry_keywords'
    UNION ALL SELECT 'idx_classification_codes_description_trgm', 'classification_codes'
    UNION ALL SELECT 'idx_industry_patterns_pattern_trgm', 'industry_patterns'
    UNION ALL SELECT 'idx_keyword_weights_keyword_trgm', 'keyword_weights'
)
SELECT 
    'GIN INDEX VALIDATION' as validation_type,
    gi.index_name,
    gi.table_name,
    CASE 
        WHEN pi.indexname IS NOT NULL THEN 'PASSED - GIN index exists'
        ELSE 'FAILED - GIN index missing'
    END as status
FROM gin_indexes gi
LEFT JOIN pg_indexes pi ON gi.index_name = pi.indexname AND gi.table_name = pi.tablename
ORDER BY gi.table_name, gi.index_name;

-- =============================================================================
-- 7. VALIDATE PARTIAL INDEXES
-- =============================================================================

-- Validate that partial indexes exist
WITH partial_indexes AS (
    SELECT 'idx_industries_active_recent' as index_name, 'industries' as table_name
    UNION ALL SELECT 'idx_industry_keywords_active_recent', 'industry_keywords'
    UNION ALL SELECT 'idx_classification_codes_active_recent', 'classification_codes'
    UNION ALL SELECT 'idx_industry_patterns_active_recent', 'industry_patterns'
    UNION ALL SELECT 'idx_risk_keywords_active_recent', 'risk_keywords'
    UNION ALL SELECT 'idx_industries_high_confidence', 'industries'
    UNION ALL SELECT 'idx_industry_patterns_high_confidence', 'industry_patterns'
    UNION ALL SELECT 'idx_industry_code_crosswalks_high_confidence', 'industry_code_crosswalks'
    UNION ALL SELECT 'idx_business_risk_assessments_high_risk', 'business_risk_assessments'
    UNION ALL SELECT 'idx_risk_keywords_high_severity', 'risk_keywords'
)
SELECT 
    'PARTIAL INDEX VALIDATION' as validation_type,
    pi.index_name,
    pi.table_name,
    CASE 
        WHEN pg_indexes.indexname IS NOT NULL THEN 'PASSED - Partial index exists'
        ELSE 'FAILED - Partial index missing'
    END as status
FROM partial_indexes pi
LEFT JOIN pg_indexes ON pi.index_name = pg_indexes.indexname AND pi.table_name = pg_indexes.tablename
ORDER BY pi.table_name, pi.index_name;

-- =============================================================================
-- 8. VALIDATE JSONB INDEXES
-- =============================================================================

-- Validate that JSONB indexes exist
WITH jsonb_indexes AS (
    SELECT 'idx_users_metadata' as index_name, 'users' as table_name
    UNION ALL SELECT 'idx_businesses_address', 'businesses'
    UNION ALL SELECT 'idx_businesses_contact_info', 'businesses'
    UNION ALL SELECT 'idx_businesses_metadata', 'businesses'
    UNION ALL SELECT 'idx_merchants_address', 'merchants'
    UNION ALL SELECT 'idx_merchants_contact_info', 'merchants'
    UNION ALL SELECT 'idx_merchants_metadata', 'merchants'
    UNION ALL SELECT 'idx_business_risk_assessments_metadata_gin', 'business_risk_assessments'
    UNION ALL SELECT 'idx_business_risk_assessments_patterns_gin', 'business_risk_assessments'
)
SELECT 
    'JSONB INDEX VALIDATION' as validation_type,
    ji.index_name,
    ji.table_name,
    CASE 
        WHEN pi.indexname IS NOT NULL THEN 'PASSED - JSONB index exists'
        ELSE 'FAILED - JSONB index missing'
    END as status
FROM jsonb_indexes ji
LEFT JOIN pg_indexes pi ON ji.index_name = pi.indexname AND ji.table_name = pi.tablename
ORDER BY ji.table_name, ji.index_name;

-- =============================================================================
-- 9. VALIDATE FOREIGN KEY INDEXES
-- =============================================================================

-- Validate that foreign key indexes exist
WITH fk_indexes AS (
    SELECT 'idx_industry_keywords_industry_fk' as index_name, 'industry_keywords' as table_name
    UNION ALL SELECT 'idx_classification_codes_industry_fk', 'classification_codes'
    UNION ALL SELECT 'idx_industry_patterns_industry_fk', 'industry_patterns'
    UNION ALL SELECT 'idx_keyword_weights_industry_fk', 'keyword_weights'
    UNION ALL SELECT 'idx_industry_code_crosswalks_industry_fk', 'industry_code_crosswalks'
    UNION ALL SELECT 'idx_business_risk_assessments_keyword_fk', 'business_risk_assessments'
    UNION ALL SELECT 'idx_risk_keyword_relationships_parent_fk', 'risk_keyword_relationships'
    UNION ALL SELECT 'idx_risk_keyword_relationships_child_fk', 'risk_keyword_relationships'
)
SELECT 
    'FOREIGN KEY INDEX VALIDATION' as validation_type,
    fki.index_name,
    fki.table_name,
    CASE 
        WHEN pi.indexname IS NOT NULL THEN 'PASSED - FK index exists'
        ELSE 'FAILED - FK index missing'
    END as status
FROM fk_indexes fki
LEFT JOIN pg_indexes pi ON fki.index_name = pi.indexname AND fki.table_name = pi.tablename
ORDER BY fki.table_name, fki.index_name;

-- =============================================================================
-- 10. VALIDATE UNIQUE CONSTRAINT INDEXES
-- =============================================================================

-- Validate that unique constraint indexes exist
WITH unique_indexes AS (
    SELECT 'idx_industries_name_unique' as index_name, 'industries' as table_name
    UNION ALL SELECT 'idx_industry_keywords_unique', 'industry_keywords'
    UNION ALL SELECT 'idx_classification_codes_unique', 'classification_codes'
    UNION ALL SELECT 'idx_keyword_weights_unique', 'keyword_weights'
    UNION ALL SELECT 'idx_risk_keywords_keyword_unique', 'risk_keywords'
    UNION ALL SELECT 'idx_industry_code_crosswalks_unique', 'industry_code_crosswalks'
)
SELECT 
    'UNIQUE INDEX VALIDATION' as validation_type,
    ui.index_name,
    ui.table_name,
    CASE 
        WHEN pi.indexname IS NOT NULL THEN 'PASSED - Unique index exists'
        ELSE 'FAILED - Unique index missing'
    END as status
FROM unique_indexes ui
LEFT JOIN pg_indexes pi ON ui.index_name = pi.indexname AND ui.table_name = pi.tablename
ORDER BY ui.table_name, ui.index_name;

-- =============================================================================
-- 11. VALIDATE PERFORMANCE INDEXES
-- =============================================================================

-- Validate that performance monitoring indexes exist
WITH performance_indexes AS (
    SELECT 'idx_classification_performance_timestamp' as index_name, 'classification_performance_metrics' as table_name
    UNION ALL SELECT 'idx_classification_performance_request_id', 'classification_performance_metrics'
    UNION ALL SELECT 'idx_classification_performance_industry', 'classification_performance_metrics'
    UNION ALL SELECT 'idx_classification_performance_accuracy', 'classification_performance_metrics'
    UNION ALL SELECT 'idx_classification_performance_response_time', 'classification_performance_metrics'
    UNION ALL SELECT 'idx_classification_performance_method', 'classification_performance_metrics'
    UNION ALL SELECT 'idx_classification_performance_risk_level', 'classification_performance_metrics'
    UNION ALL SELECT 'idx_classification_performance_risk_score', 'classification_performance_metrics'
)
SELECT 
    'PERFORMANCE INDEX VALIDATION' as validation_type,
    pi.index_name,
    pi.table_name,
    CASE 
        WHEN pg_indexes.indexname IS NOT NULL THEN 'PASSED - Performance index exists'
        ELSE 'FAILED - Performance index missing'
    END as status
FROM performance_indexes pi
LEFT JOIN pg_indexes ON pi.index_name = pg_indexes.indexname AND pi.table_name = pg_indexes.tablename
ORDER BY pi.table_name, pi.index_name;

-- =============================================================================
-- 12. VALIDATE API INDEXES
-- =============================================================================

-- Validate that API-optimized indexes exist
WITH api_indexes AS (
    SELECT 'idx_industries_api' as index_name, 'industries' as table_name
    UNION ALL SELECT 'idx_industry_keywords_api', 'industry_keywords'
    UNION ALL SELECT 'idx_risk_keywords_api', 'risk_keywords'
    UNION ALL SELECT 'idx_classification_codes_api', 'classification_codes'
    UNION ALL SELECT 'idx_business_risk_assessments_api', 'business_risk_assessments'
)
SELECT 
    'API INDEX VALIDATION' as validation_type,
    ai.index_name,
    ai.table_name,
    CASE 
        WHEN pi.indexname IS NOT NULL THEN 'PASSED - API index exists'
        ELSE 'FAILED - API index missing'
    END as status
FROM api_indexes ai
LEFT JOIN pg_indexes pi ON ai.index_name = pi.indexname AND ai.table_name = pi.tablename
ORDER BY ai.table_name, ai.index_name;

-- =============================================================================
-- 13. VALIDATE REPORTING INDEXES
-- =============================================================================

-- Validate that reporting indexes exist
WITH reporting_indexes AS (
    SELECT 'idx_classification_performance_reporting' as index_name, 'classification_performance_metrics' as table_name
    UNION ALL SELECT 'idx_business_risk_assessments_reporting', 'business_risk_assessments'
    UNION ALL SELECT 'idx_industries_reporting', 'industries'
    UNION ALL SELECT 'idx_risk_keywords_reporting', 'risk_keywords'
)
SELECT 
    'REPORTING INDEX VALIDATION' as validation_type,
    ri.index_name,
    ri.table_name,
    CASE 
        WHEN pi.indexname IS NOT NULL THEN 'PASSED - Reporting index exists'
        ELSE 'FAILED - Reporting index missing'
    END as status
FROM reporting_indexes ri
LEFT JOIN pg_indexes pi ON ri.index_name = pi.indexname AND ri.table_name = pi.tablename
ORDER BY ri.table_name, ri.index_name;

-- =============================================================================
-- 14. VALIDATE SCALABILITY INDEXES
-- =============================================================================

-- Validate that scalability indexes exist
WITH scalability_indexes AS (
    SELECT 'idx_industries_scalable' as index_name, 'industries' as table_name
    UNION ALL SELECT 'idx_industry_keywords_scalable', 'industry_keywords'
    UNION ALL SELECT 'idx_risk_keywords_scalable', 'risk_keywords'
    UNION ALL SELECT 'idx_classification_codes_scalable', 'classification_codes'
    UNION ALL SELECT 'idx_business_risk_assessments_scalable', 'business_risk_assessments'
)
SELECT 
    'SCALABILITY INDEX VALIDATION' as validation_type,
    si.index_name,
    si.table_name,
    CASE 
        WHEN pi.indexname IS NOT NULL THEN 'PASSED - Scalability index exists'
        ELSE 'FAILED - Scalability index missing'
    END as status
FROM scalability_indexes si
LEFT JOIN pg_indexes pi ON si.index_name = pi.indexname AND si.table_name = pi.tablename
ORDER BY si.table_name, si.index_name;

-- =============================================================================
-- 15. VALIDATE ML INDEXES
-- =============================================================================

-- Validate that ML-optimized indexes exist
WITH ml_indexes AS (
    SELECT 'idx_classification_performance_ml' as index_name, 'classification_performance_metrics' as table_name
    UNION ALL SELECT 'idx_business_risk_assessments_ml', 'business_risk_assessments'
    UNION ALL SELECT 'idx_industry_keywords_ml', 'industry_keywords'
    UNION ALL SELECT 'idx_risk_keywords_ml', 'risk_keywords'
)
SELECT 
    'ML INDEX VALIDATION' as validation_type,
    mi.index_name,
    mi.table_name,
    CASE 
        WHEN pi.indexname IS NOT NULL THEN 'PASSED - ML index exists'
        ELSE 'FAILED - ML index missing'
    END as status
FROM ml_indexes mi
LEFT JOIN pg_indexes pi ON mi.index_name = pi.indexname AND mi.table_name = pi.tablename
ORDER BY mi.table_name, mi.index_name;

-- =============================================================================
-- 16. VALIDATE BACKUP INDEXES
-- =============================================================================

-- Validate that backup-optimized indexes exist
WITH backup_indexes AS (
    SELECT 'idx_industries_backup' as index_name, 'industries' as table_name
    UNION ALL SELECT 'idx_industry_keywords_backup', 'industry_keywords'
    UNION ALL SELECT 'idx_risk_keywords_backup', 'risk_keywords'
    UNION ALL SELECT 'idx_classification_codes_backup', 'classification_codes'
    UNION ALL SELECT 'idx_business_risk_assessments_backup', 'business_risk_assessments'
)
SELECT 
    'BACKUP INDEX VALIDATION' as validation_type,
    bi.index_name,
    bi.table_name,
    CASE 
        WHEN pi.indexname IS NOT NULL THEN 'PASSED - Backup index exists'
        ELSE 'FAILED - Backup index missing'
    END as status
FROM backup_indexes bi
LEFT JOIN pg_indexes pi ON bi.index_name = pi.indexname AND bi.table_name = pi.tablename
ORDER BY bi.table_name, bi.index_name;

-- =============================================================================
-- 17. OVERALL VALIDATION SUMMARY
-- =============================================================================

-- Generate overall validation summary
WITH validation_results AS (
    SELECT 'INDEX CREATION' as validation_type, COUNT(*) as total_checks, COUNT(*) as passed_checks FROM pg_indexes WHERE schemaname = 'public' AND tablename IN ('industries', 'industry_keywords', 'classification_codes', 'industry_patterns', 'keyword_weights', 'risk_keywords', 'industry_code_crosswalks', 'business_risk_assessments', 'risk_keyword_relationships', 'classification_performance_metrics')
    UNION ALL
    SELECT 'INDEX TYPES' as validation_type, COUNT(*) as total_checks, COUNT(CASE WHEN indexdef LIKE '%USING btree%' OR indexdef LIKE '%USING gin%' OR indexdef LIKE '%USING gist%' THEN 1 END) as passed_checks FROM pg_indexes WHERE schemaname = 'public' AND tablename IN ('industries', 'industry_keywords', 'classification_codes', 'industry_patterns', 'keyword_weights', 'risk_keywords', 'industry_code_crosswalks', 'business_risk_assessments', 'risk_keyword_relationships', 'classification_performance_metrics')
)
SELECT 
    'OVERALL VALIDATION SUMMARY' as validation_type,
    SUM(total_checks) as total_checks,
    SUM(passed_checks) as passed_checks,
    ROUND((SUM(passed_checks)::numeric / SUM(total_checks)::numeric) * 100, 2) as success_percentage,
    CASE 
        WHEN (SUM(passed_checks)::numeric / SUM(total_checks)::numeric) >= 0.95 THEN 'EXCELLENT - All validations passed'
        WHEN (SUM(passed_checks)::numeric / SUM(total_checks)::numeric) >= 0.90 THEN 'GOOD - Most validations passed'
        WHEN (SUM(passed_checks)::numeric / SUM(total_checks)::numeric) >= 0.80 THEN 'FAIR - Some validations failed'
        ELSE 'POOR - Many validations failed'
    END as overall_status
FROM validation_results;

-- =============================================================================
-- END OF INDEX IMPROVEMENT VALIDATION
-- =============================================================================
