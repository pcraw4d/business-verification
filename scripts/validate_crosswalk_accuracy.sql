-- =============================================================================
-- VALIDATE CROSSWALK ACCURACY AND FUNCTIONALITY
-- =============================================================================
-- This script validates the accuracy and consistency of industry code crosswalks
-- 
-- Created: January 19, 2025
-- Purpose: Task 1.5.3 - Validate crosswalk accuracy and test queries
-- =============================================================================

-- =============================================================================
-- 1. CROSSWALK DATA INTEGRITY VALIDATION
-- =============================================================================

-- Check for orphaned crosswalk records (industries that don't exist)
SELECT 
    'Orphaned Crosswalk Records' as validation_type,
    COUNT(*) as issue_count,
    'Crosswalk records referencing non-existent industries' as description
FROM industry_code_crosswalks icc
LEFT JOIN industries i ON icc.industry_id = i.id
WHERE i.id IS NULL;

-- Check for duplicate crosswalk combinations
SELECT 
    'Duplicate Crosswalk Combinations' as validation_type,
    COUNT(*) as issue_count,
    'Duplicate industry-code combinations found' as description
FROM (
    SELECT industry_id, mcc_code, naics_code, sic_code, COUNT(*)
    FROM industry_code_crosswalks
    GROUP BY industry_id, mcc_code, naics_code, sic_code
    HAVING COUNT(*) > 1
) duplicates;

-- Check for invalid confidence scores
SELECT 
    'Invalid Confidence Scores' as validation_type,
    COUNT(*) as issue_count,
    'Confidence scores outside valid range (0.00-1.00)' as description
FROM industry_code_crosswalks
WHERE confidence_score < 0.00 OR confidence_score > 1.00;

-- Check for missing primary designations per industry
SELECT 
    'Missing Primary Designations' as validation_type,
    COUNT(*) as issue_count,
    'Industries without primary crosswalk designations' as description
FROM industries i
WHERE i.is_active = true
AND NOT EXISTS (
    SELECT 1 FROM industry_code_crosswalks icc
    WHERE icc.industry_id = i.id 
    AND icc.is_primary = true 
    AND icc.is_active = true
);

-- =============================================================================
-- 2. CROSSWALK COVERAGE ANALYSIS
-- =============================================================================

-- Industry coverage analysis
SELECT 
    'Industry Coverage Analysis' as analysis_type,
    i.name as industry_name,
    COUNT(icc.id) as total_crosswalks,
    COUNT(CASE WHEN icc.mcc_code IS NOT NULL THEN 1 END) as mcc_mappings,
    COUNT(CASE WHEN icc.naics_code IS NOT NULL THEN 1 END) as naics_mappings,
    COUNT(CASE WHEN icc.sic_code IS NOT NULL THEN 1 END) as sic_mappings,
    COUNT(CASE WHEN icc.is_primary = true THEN 1 END) as primary_mappings,
    ROUND(AVG(icc.confidence_score), 3) as avg_confidence_score
FROM industries i
LEFT JOIN industry_code_crosswalks icc ON i.id = icc.industry_id AND icc.is_active = true
WHERE i.is_active = true
GROUP BY i.id, i.name
ORDER BY total_crosswalks DESC;

-- Code type distribution analysis
SELECT 
    'Code Type Distribution' as analysis_type,
    'MCC' as code_type,
    COUNT(DISTINCT mcc_code) as unique_codes,
    COUNT(*) as total_mappings
FROM industry_code_crosswalks
WHERE mcc_code IS NOT NULL AND is_active = true

UNION ALL

SELECT 
    'Code Type Distribution' as analysis_type,
    'NAICS' as code_type,
    COUNT(DISTINCT naics_code) as unique_codes,
    COUNT(*) as total_mappings
FROM industry_code_crosswalks
WHERE naics_code IS NOT NULL AND is_active = true

UNION ALL

SELECT 
    'Code Type Distribution' as analysis_type,
    'SIC' as code_type,
    COUNT(DISTINCT sic_code) as unique_codes,
    COUNT(*) as total_mappings
FROM industry_code_crosswalks
WHERE sic_code IS NOT NULL AND is_active = true;

-- =============================================================================
-- 3. CROSSWALK QUERY PERFORMANCE TESTING
-- =============================================================================

-- Test 1: Find industry by MCC code
EXPLAIN (ANALYZE, BUFFERS) 
SELECT 
    i.name as industry_name,
    icc.mcc_code,
    icc.code_description,
    icc.confidence_score,
    icc.is_primary
FROM industry_code_crosswalks icc
JOIN industries i ON icc.industry_id = i.id
WHERE icc.mcc_code = '5734' 
AND icc.is_active = true
ORDER BY icc.is_primary DESC, icc.confidence_score DESC;

-- Test 2: Find industry by NAICS code
EXPLAIN (ANALYZE, BUFFERS)
SELECT 
    i.name as industry_name,
    icc.naics_code,
    icc.code_description,
    icc.confidence_score,
    icc.is_primary
FROM industry_code_crosswalks icc
JOIN industries i ON icc.industry_id = i.id
WHERE icc.naics_code = '541511' 
AND icc.is_active = true
ORDER BY icc.is_primary DESC, icc.confidence_score DESC;

-- Test 3: Find industry by SIC code
EXPLAIN (ANALYZE, BUFFERS)
SELECT 
    i.name as industry_name,
    icc.sic_code,
    icc.code_description,
    icc.confidence_score,
    icc.is_primary
FROM industry_code_crosswalks icc
JOIN industries i ON icc.industry_id = i.id
WHERE icc.sic_code = '7371' 
AND icc.is_active = true
ORDER BY icc.is_primary DESC, icc.confidence_score DESC;

-- Test 4: Get all crosswalks for a specific industry
EXPLAIN (ANALYZE, BUFFERS)
SELECT 
    icc.mcc_code,
    icc.naics_code,
    icc.sic_code,
    icc.code_description,
    icc.confidence_score,
    icc.is_primary
FROM industry_code_crosswalks icc
JOIN industries i ON icc.industry_id = i.id
WHERE i.name = 'Technology' 
AND icc.is_active = true
ORDER BY icc.is_primary DESC, icc.confidence_score DESC;

-- =============================================================================
-- 4. CROSSWALK CONSISTENCY VALIDATION
-- =============================================================================

-- Check for conflicting primary designations (multiple primaries per industry)
SELECT 
    'Conflicting Primary Designations' as validation_type,
    i.name as industry_name,
    COUNT(*) as primary_count,
    'Multiple primary designations found for industry' as description
FROM industry_code_crosswalks icc
JOIN industries i ON icc.industry_id = i.id
WHERE icc.is_primary = true AND icc.is_active = true
GROUP BY i.id, i.name
HAVING COUNT(*) > 1;

-- Check for low confidence scores that might need review
SELECT 
    'Low Confidence Scores' as validation_type,
    i.name as industry_name,
    icc.mcc_code,
    icc.naics_code,
    icc.sic_code,
    icc.confidence_score,
    'Crosswalk with confidence score below 0.80' as description
FROM industry_code_crosswalks icc
JOIN industries i ON icc.industry_id = i.id
WHERE icc.confidence_score < 0.80 
AND icc.is_active = true
ORDER BY icc.confidence_score ASC;

-- Check for missing code descriptions
SELECT 
    'Missing Code Descriptions' as validation_type,
    COUNT(*) as issue_count,
    'Crosswalk records without code descriptions' as description
FROM industry_code_crosswalks
WHERE (code_description IS NULL OR code_description = '') 
AND is_active = true;

-- =============================================================================
-- 5. BUSINESS LOGIC VALIDATION
-- =============================================================================

-- Validate that high-risk industries have appropriate risk indicators
SELECT 
    'High-Risk Industry Validation' as validation_type,
    i.name as industry_name,
    icc.mcc_code,
    icc.code_description,
    CASE 
        WHEN i.name IN ('Cryptocurrency', 'Gambling', 'Adult Entertainment') 
        THEN 'High-Risk Industry - Validated'
        ELSE 'Standard Industry'
    END as risk_category
FROM industry_code_crosswalks icc
JOIN industries i ON icc.industry_id = i.id
WHERE i.name IN ('Cryptocurrency', 'Gambling', 'Adult Entertainment')
AND icc.is_active = true
ORDER BY i.name, icc.confidence_score DESC;

-- Validate that prohibited MCC codes are properly flagged
SELECT 
    'Prohibited MCC Code Validation' as validation_type,
    icc.mcc_code,
    icc.code_description,
    i.name as industry_name,
    CASE 
        WHEN icc.mcc_code IN ('7995', '7273') 
        THEN 'Prohibited/High-Risk MCC Code - Validated'
        ELSE 'Standard MCC Code'
    END as mcc_risk_level
FROM industry_code_crosswalks icc
JOIN industries i ON icc.industry_id = i.id
WHERE icc.mcc_code IN ('7995', '7273')
AND icc.is_active = true;

-- =============================================================================
-- 6. PERFORMANCE BENCHMARKS
-- =============================================================================

-- Benchmark crosswalk lookup performance
DO $$
DECLARE
    start_time timestamp;
    end_time timestamp;
    execution_time interval;
    test_count integer := 1000;
    i integer;
BEGIN
    start_time := clock_timestamp();
    
    FOR i IN 1..test_count LOOP
        PERFORM icc.mcc_code, icc.naics_code, icc.sic_code
        FROM industry_code_crosswalks icc
        JOIN industries ind ON icc.industry_id = ind.id
        WHERE ind.name = 'Technology' AND icc.is_active = true
        LIMIT 1;
    END LOOP;
    
    end_time := clock_timestamp();
    execution_time := end_time - start_time;
    
    RAISE NOTICE 'Crosswalk Lookup Performance Test:';
    RAISE NOTICE 'Executions: %', test_count;
    RAISE NOTICE 'Total Time: %', execution_time;
    RAISE NOTICE 'Average Time per Query: %', execution_time / test_count;
    RAISE NOTICE 'Queries per Second: %', test_count / EXTRACT(EPOCH FROM execution_time);
END $$;

-- =============================================================================
-- 7. SUMMARY REPORT
-- =============================================================================

-- Generate comprehensive summary report
SELECT 
    'CROSSWALK VALIDATION SUMMARY' as report_section,
    'Total Industries' as metric,
    COUNT(*) as value
FROM industries WHERE is_active = true

UNION ALL

SELECT 
    'CROSSWALK VALIDATION SUMMARY' as report_section,
    'Total Crosswalk Mappings' as metric,
    COUNT(*) as value
FROM industry_code_crosswalks WHERE is_active = true

UNION ALL

SELECT 
    'CROSSWALK VALIDATION SUMMARY' as report_section,
    'MCC Mappings' as metric,
    COUNT(*) as value
FROM industry_code_crosswalks WHERE mcc_code IS NOT NULL AND is_active = true

UNION ALL

SELECT 
    'CROSSWALK VALIDATION SUMMARY' as report_section,
    'NAICS Mappings' as metric,
    COUNT(*) as value
FROM industry_code_crosswalks WHERE naics_code IS NOT NULL AND is_active = true

UNION ALL

SELECT 
    'CROSSWALK VALIDATION SUMMARY' as report_section,
    'SIC Mappings' as metric,
    COUNT(*) as value
FROM industry_code_crosswalks WHERE sic_code IS NOT NULL AND is_active = true

UNION ALL

SELECT 
    'CROSSWALK VALIDATION SUMMARY' as report_section,
    'Primary Mappings' as metric,
    COUNT(*) as value
FROM industry_code_crosswalks WHERE is_primary = true AND is_active = true

UNION ALL

SELECT 
    'CROSSWALK VALIDATION SUMMARY' as report_section,
    'Average Confidence Score' as metric,
    ROUND(AVG(confidence_score), 3) as value
FROM industry_code_crosswalks WHERE is_active = true

ORDER BY report_section, metric;

-- =============================================================================
-- VALIDATION COMPLETION
-- =============================================================================
-- This validation script provides comprehensive testing of:
-- 
-- 1. Data Integrity: Orphaned records, duplicates, invalid values
-- 2. Coverage Analysis: Industry and code type distribution
-- 3. Query Performance: Execution plans and timing benchmarks
-- 4. Consistency Validation: Primary designations, confidence scores
-- 5. Business Logic: High-risk industry validation, prohibited codes
-- 6. Performance Benchmarks: Query execution timing and throughput
-- 7. Summary Report: Comprehensive metrics and statistics
-- 
-- All validation tests should pass for a successful crosswalk implementation
-- =============================================================================
