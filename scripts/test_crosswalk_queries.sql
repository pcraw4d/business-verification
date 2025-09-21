-- =============================================================================
-- TEST CROSSWALK QUERIES AND FUNCTIONALITY
-- =============================================================================
-- This script tests the functionality of industry code crosswalks with
-- comprehensive query scenarios and expected results validation
-- 
-- Created: January 19, 2025
-- Purpose: Task 1.5.3 - Test crosswalk queries and functionality
-- =============================================================================

-- =============================================================================
-- 1. BASIC CROSSWALK LOOKUP TESTS
-- =============================================================================

-- Test 1: Find industry by MCC code (Computer Software Stores)
SELECT 
    'Test 1: MCC Code Lookup' as test_name,
    i.name as industry_name,
    icc.mcc_code,
    icc.code_description,
    icc.confidence_score,
    icc.is_primary,
    CASE 
        WHEN i.name = 'Technology' AND icc.mcc_code = '5734' AND icc.confidence_score >= 0.90
        THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM industry_code_crosswalks icc
JOIN industries i ON icc.industry_id = i.id
WHERE icc.mcc_code = '5734' 
AND icc.is_active = true
ORDER BY icc.is_primary DESC, icc.confidence_score DESC
LIMIT 1;

-- Test 2: Find industry by NAICS code (Custom Computer Programming Services)
SELECT 
    'Test 2: NAICS Code Lookup' as test_name,
    i.name as industry_name,
    icc.naics_code,
    icc.code_description,
    icc.confidence_score,
    icc.is_primary,
    CASE 
        WHEN i.name = 'Technology' AND icc.naics_code = '541511' AND icc.confidence_score >= 0.90
        THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM industry_code_crosswalks icc
JOIN industries i ON icc.industry_id = i.id
WHERE icc.naics_code = '541511' 
AND icc.is_active = true
ORDER BY icc.is_primary DESC, icc.confidence_score DESC
LIMIT 1;

-- Test 3: Find industry by SIC code (Computer Programming Services)
SELECT 
    'Test 3: SIC Code Lookup' as test_name,
    i.name as industry_name,
    icc.sic_code,
    icc.code_description,
    icc.confidence_score,
    icc.is_primary,
    CASE 
        WHEN i.name = 'Technology' AND icc.sic_code = '7371' AND icc.confidence_score >= 0.90
        THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM industry_code_crosswalks icc
JOIN industries i ON icc.industry_id = i.id
WHERE icc.sic_code = '7371' 
AND icc.is_active = true
ORDER BY icc.is_primary DESC, icc.confidence_score DESC
LIMIT 1;

-- =============================================================================
-- 2. COMPREHENSIVE INDUSTRY CROSSWALK TESTS
-- =============================================================================

-- Test 4: Get all crosswalks for Technology industry
SELECT 
    'Test 4: Technology Industry Crosswalks' as test_name,
    icc.mcc_code,
    icc.naics_code,
    icc.sic_code,
    icc.code_description,
    icc.confidence_score,
    icc.is_primary,
    CASE 
        WHEN COUNT(*) >= 6 AND COUNT(CASE WHEN icc.is_primary = true THEN 1 END) >= 3
        THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM industry_code_crosswalks icc
JOIN industries i ON icc.industry_id = i.id
WHERE i.name = 'Technology' 
AND icc.is_active = true
GROUP BY icc.mcc_code, icc.naics_code, icc.sic_code, icc.code_description, icc.confidence_score, icc.is_primary
ORDER BY icc.is_primary DESC, icc.confidence_score DESC;

-- Test 5: Get all crosswalks for Financial Services industry
SELECT 
    'Test 5: Financial Services Industry Crosswalks' as test_name,
    icc.mcc_code,
    icc.naics_code,
    icc.sic_code,
    icc.code_description,
    icc.confidence_score,
    icc.is_primary,
    CASE 
        WHEN COUNT(*) >= 6 AND COUNT(CASE WHEN icc.is_primary = true THEN 1 END) >= 3
        THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM industry_code_crosswalks icc
JOIN industries i ON icc.industry_id = i.id
WHERE i.name = 'Financial Services' 
AND icc.is_active = true
GROUP BY icc.mcc_code, icc.naics_code, icc.sic_code, icc.code_description, icc.confidence_score, icc.is_primary
ORDER BY icc.is_primary DESC, icc.confidence_score DESC;

-- =============================================================================
-- 3. HIGH-RISK INDUSTRY VALIDATION TESTS
-- =============================================================================

-- Test 6: Validate high-risk industry mappings
SELECT 
    'Test 6: High-Risk Industry Validation' as test_name,
    i.name as industry_name,
    icc.mcc_code,
    icc.naics_code,
    icc.sic_code,
    icc.confidence_score,
    CASE 
        WHEN i.name IN ('Cryptocurrency', 'Gambling', 'Adult Entertainment') 
        AND icc.confidence_score >= 0.75
        THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM industry_code_crosswalks icc
JOIN industries i ON icc.industry_id = i.id
WHERE i.name IN ('Cryptocurrency', 'Gambling', 'Adult Entertainment')
AND icc.is_active = true
ORDER BY i.name, icc.confidence_score DESC;

-- Test 7: Validate prohibited MCC codes
SELECT 
    'Test 7: Prohibited MCC Code Validation' as test_name,
    icc.mcc_code,
    icc.code_description,
    i.name as industry_name,
    icc.confidence_score,
    CASE 
        WHEN icc.mcc_code IN ('7995', '7273') 
        AND i.name IN ('Gambling', 'Adult Entertainment', 'Cryptocurrency')
        THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM industry_code_crosswalks icc
JOIN industries i ON icc.industry_id = i.id
WHERE icc.mcc_code IN ('7995', '7273')
AND icc.is_active = true
ORDER BY icc.mcc_code, icc.confidence_score DESC;

-- =============================================================================
-- 4. CROSSWALK ACCURACY AND CONSISTENCY TESTS
-- =============================================================================

-- Test 8: Validate confidence score ranges
SELECT 
    'Test 8: Confidence Score Validation' as test_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN confidence_score >= 0.75 AND confidence_score <= 1.00 THEN 1 END) as valid_scores,
    COUNT(CASE WHEN confidence_score < 0.75 OR confidence_score > 1.00 THEN 1 END) as invalid_scores,
    CASE 
        WHEN COUNT(CASE WHEN confidence_score < 0.75 OR confidence_score > 1.00 THEN 1 END) = 0
        THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM industry_code_crosswalks
WHERE is_active = true;

-- Test 9: Validate primary designation uniqueness
SELECT 
    'Test 9: Primary Designation Uniqueness' as test_name,
    i.name as industry_name,
    COUNT(*) as primary_count,
    CASE 
        WHEN COUNT(*) <= 3 -- Allow up to 3 primaries per industry (one per code type)
        THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM industry_code_crosswalks icc
JOIN industries i ON icc.industry_id = i.id
WHERE icc.is_primary = true AND icc.is_active = true
GROUP BY i.id, i.name
ORDER BY primary_count DESC;

-- Test 10: Validate code description completeness
SELECT 
    'Test 10: Code Description Completeness' as test_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN code_description IS NOT NULL AND code_description != '' THEN 1 END) as with_descriptions,
    COUNT(CASE WHEN code_description IS NULL OR code_description = '' THEN 1 END) as without_descriptions,
    CASE 
        WHEN COUNT(CASE WHEN code_description IS NULL OR code_description = '' THEN 1 END) = 0
        THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM industry_code_crosswalks
WHERE is_active = true;

-- =============================================================================
-- 5. PERFORMANCE AND SCALABILITY TESTS
-- =============================================================================

-- Test 11: Crosswalk lookup performance test
DO $$
DECLARE
    start_time timestamp;
    end_time timestamp;
    execution_time interval;
    test_count integer := 100;
    i integer;
    result_count integer;
BEGIN
    start_time := clock_timestamp();
    
    FOR i IN 1..test_count LOOP
        SELECT COUNT(*) INTO result_count
        FROM industry_code_crosswalks icc
        JOIN industries ind ON icc.industry_id = ind.id
        WHERE ind.name = 'Technology' AND icc.is_active = true;
    END LOOP;
    
    end_time := clock_timestamp();
    execution_time := end_time - start_time;
    
    RAISE NOTICE 'Test 11: Crosswalk Lookup Performance';
    RAISE NOTICE 'Executions: %', test_count;
    RAISE NOTICE 'Total Time: %', execution_time;
    RAISE NOTICE 'Average Time per Query: %', execution_time / test_count;
    RAISE NOTICE 'Queries per Second: %', test_count / EXTRACT(EPOCH FROM execution_time);
    
    -- Performance test result
    IF EXTRACT(EPOCH FROM execution_time / test_count) < 0.01 THEN -- Less than 10ms per query
        RAISE NOTICE 'Test 11 Result: PASS - Performance within acceptable limits';
    ELSE
        RAISE NOTICE 'Test 11 Result: FAIL - Performance below acceptable limits';
    END IF;
END $$;

-- Test 12: Complex crosswalk query performance
DO $$
DECLARE
    start_time timestamp;
    end_time timestamp;
    execution_time interval;
    test_count integer := 50;
    i integer;
    result_count integer;
BEGIN
    start_time := clock_timestamp();
    
    FOR i IN 1..test_count LOOP
        SELECT COUNT(*) INTO result_count
        FROM industry_code_crosswalks icc
        JOIN industries i ON icc.industry_id = i.id
        WHERE (icc.mcc_code IS NOT NULL OR icc.naics_code IS NOT NULL OR icc.sic_code IS NOT NULL)
        AND icc.is_active = true
        AND icc.confidence_score >= 0.80
        AND icc.is_primary = true;
    END LOOP;
    
    end_time := clock_timestamp();
    execution_time := end_time - start_time;
    
    RAISE NOTICE 'Test 12: Complex Crosswalk Query Performance';
    RAISE NOTICE 'Executions: %', test_count;
    RAISE NOTICE 'Total Time: %', execution_time;
    RAISE NOTICE 'Average Time per Query: %', execution_time / test_count;
    RAISE NOTICE 'Queries per Second: %', test_count / EXTRACT(EPOCH FROM execution_time);
    
    -- Performance test result
    IF EXTRACT(EPOCH FROM execution_time / test_count) < 0.05 THEN -- Less than 50ms per query
        RAISE NOTICE 'Test 12 Result: PASS - Complex query performance within acceptable limits';
    ELSE
        RAISE NOTICE 'Test 12 Result: FAIL - Complex query performance below acceptable limits';
    END IF;
END $$;

-- =============================================================================
-- 6. BUSINESS LOGIC INTEGRATION TESTS
-- =============================================================================

-- Test 13: Crosswalk integration with classification system
SELECT 
    'Test 13: Classification System Integration' as test_name,
    i.name as industry_name,
    COUNT(icc.id) as crosswalk_count,
    COUNT(CASE WHEN icc.mcc_code IS NOT NULL THEN 1 END) as mcc_mappings,
    COUNT(CASE WHEN icc.naics_code IS NOT NULL THEN 1 END) as naics_mappings,
    COUNT(CASE WHEN icc.sic_code IS NOT NULL THEN 1 END) as sic_mappings,
    ROUND(AVG(icc.confidence_score), 3) as avg_confidence,
    CASE 
        WHEN COUNT(icc.id) >= 3 AND AVG(icc.confidence_score) >= 0.80
        THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM industries i
LEFT JOIN industry_code_crosswalks icc ON i.id = icc.industry_id AND icc.is_active = true
WHERE i.is_active = true
GROUP BY i.id, i.name
ORDER BY crosswalk_count DESC;

-- Test 14: Risk assessment integration validation
SELECT 
    'Test 14: Risk Assessment Integration' as test_name,
    i.name as industry_name,
    icc.mcc_code,
    icc.confidence_score,
    CASE 
        WHEN i.name IN ('Cryptocurrency', 'Gambling', 'Adult Entertainment')
        AND icc.mcc_code IN ('7995', '7273')
        AND icc.confidence_score >= 0.75
        THEN 'HIGH_RISK_VALIDATED'
        WHEN i.name IN ('Technology', 'Financial Services', 'Healthcare', 'Retail', 'Manufacturing', 'E-commerce')
        AND icc.confidence_score >= 0.85
        THEN 'STANDARD_RISK_VALIDATED'
        ELSE 'REVIEW_REQUIRED'
    END as risk_assessment_result
FROM industry_code_crosswalks icc
JOIN industries i ON icc.industry_id = i.id
WHERE icc.is_active = true
ORDER BY i.name, icc.confidence_score DESC;

-- =============================================================================
-- 7. COMPREHENSIVE TEST SUMMARY
-- =============================================================================

-- Generate test summary report
SELECT 
    'CROSSWALK FUNCTIONALITY TEST SUMMARY' as report_section,
    'Total Test Scenarios' as metric,
    '14' as value

UNION ALL

SELECT 
    'CROSSWALK FUNCTIONALITY TEST SUMMARY' as report_section,
    'Basic Lookup Tests' as metric,
    '3' as value

UNION ALL

SELECT 
    'CROSSWALK FUNCTIONALITY TEST SUMMARY' as report_section,
    'Industry Crosswalk Tests' as metric,
    '2' as value

UNION ALL

SELECT 
    'CROSSWALK FUNCTIONALITY TEST SUMMARY' as report_section,
    'High-Risk Validation Tests' as metric,
    '2' as value

UNION ALL

SELECT 
    'CROSSWALK FUNCTIONALITY TEST SUMMARY' as report_section,
    'Accuracy & Consistency Tests' as metric,
    '3' as value

UNION ALL

SELECT 
    'CROSSWALK FUNCTIONALITY TEST SUMMARY' as report_section,
    'Performance Tests' as metric,
    '2' as value

UNION ALL

SELECT 
    'CROSSWALK FUNCTIONALITY TEST SUMMARY' as report_section,
    'Business Logic Integration Tests' as metric,
    '2' as value

ORDER BY report_section, metric;

-- =============================================================================
-- TEST COMPLETION SUMMARY
-- =============================================================================
-- This test script validates:
-- 
-- 1. Basic Lookup Functionality: MCC, NAICS, SIC code lookups
-- 2. Industry Crosswalk Completeness: All major industries covered
-- 3. High-Risk Industry Validation: Proper risk categorization
-- 4. Data Quality: Confidence scores, primary designations, descriptions
-- 5. Performance: Query execution time and scalability
-- 6. Business Logic Integration: Risk assessment and classification alignment
-- 7. Comprehensive Coverage: All crosswalk scenarios tested
-- 
-- All tests should return 'PASS' for successful crosswalk implementation
-- =============================================================================
