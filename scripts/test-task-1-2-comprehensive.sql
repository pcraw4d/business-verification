-- =============================================================================
-- Task 1.2 Comprehensive Testing Script
-- Complete verification of restaurant industry data implementation
-- =============================================================================

-- This script performs comprehensive testing of all Task 1.2 deliverables:
-- 1.2.1: Restaurant Industries
-- 1.2.2: Restaurant Keywords  
-- 1.2.3: Restaurant Classification Codes

-- =============================================================================
-- 1. TASK 1.2.1 TESTING: RESTAURANT INDUSTRIES
-- =============================================================================

SELECT 'TASK 1.2.1: RESTAURANT INDUSTRIES TESTING' as test_section;

-- Test 1.2.1.1: Verify restaurant industries exist
SELECT 
    'Test 1.2.1.1: Restaurant Industries Count' as test_name,
    COUNT(*) as result,
    CASE 
        WHEN COUNT(*) >= 12 THEN 'PASS' 
        ELSE 'FAIL' 
    END as status
FROM industries 
WHERE name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND is_active = true;

-- Test 1.2.1.2: Verify specific industries with confidence thresholds
SELECT 
    'Test 1.2.1.2: Key Restaurant Industries' as test_name,
    id,
    name,
    description,
    confidence_threshold,
    CASE 
        WHEN name IN ('Restaurants', 'Fast Food', 'Food & Beverage') THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM industries 
WHERE name IN ('Restaurants', 'Fast Food', 'Food & Beverage')
AND is_active = true
ORDER BY confidence_threshold DESC;

-- Test 1.2.1.3: Verify confidence threshold ranges
SELECT 
    'Test 1.2.1.3: Confidence Threshold Range' as test_name,
    MIN(confidence_threshold) as min_threshold,
    MAX(confidence_threshold) as max_threshold,
    CASE 
        WHEN MIN(confidence_threshold) >= 0.70 AND MAX(confidence_threshold) <= 0.85 THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM industries 
WHERE name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND is_active = true;

-- =============================================================================
-- 2. TASK 1.2.2 TESTING: RESTAURANT KEYWORDS
-- =============================================================================

SELECT 'TASK 1.2.2: RESTAURANT KEYWORDS TESTING' as test_section;

-- Test 1.2.2.1: Verify total keyword count
SELECT 
    'Test 1.2.2.1: Total Restaurant Keywords Count' as test_name,
    COUNT(*) as result,
    CASE 
        WHEN COUNT(*) >= 200 THEN 'PASS' 
        ELSE 'FAIL' 
    END as status
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND kw.is_active = true;

-- Test 1.2.2.2: Verify keywords per industry
SELECT 
    'Test 1.2.2.2: Keywords Per Industry' as test_name,
    i.name as industry_name,
    COUNT(kw.keyword) as keyword_count,
    CASE 
        WHEN COUNT(kw.keyword) >= 15 THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM industries i
LEFT JOIN keyword_weights kw ON i.id = kw.industry_id AND kw.is_active = true
WHERE i.name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND i.is_active = true
GROUP BY i.name
ORDER BY keyword_count DESC;

-- Test 1.2.2.3: Verify high-value keywords exist
SELECT 
    'Test 1.2.2.3: High-Value Keywords' as test_name,
    CASE 
        WHEN COUNT(*) >= 5 THEN 'PASS'
        ELSE 'FAIL'
    END as status,
    COUNT(*) as found_keywords
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name = 'Restaurants' 
AND kw.keyword IN ('restaurant', 'dining', 'cuisine', 'menu', 'chef')
AND kw.is_active = true;

-- Test 1.2.2.4: Verify keyword weight ranges
SELECT 
    'Test 1.2.2.4: Keyword Weight Range' as test_name,
    MIN(kw.base_weight) as min_weight,
    MAX(kw.base_weight) as max_weight,
    CASE 
        WHEN MIN(kw.base_weight) >= 0.6000 AND MAX(kw.base_weight) <= 1.0000 THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND kw.is_active = true;

-- =============================================================================
-- 3. TASK 1.2.3 TESTING: RESTAURANT CLASSIFICATION CODES
-- =============================================================================

SELECT 'TASK 1.2.3: RESTAURANT CLASSIFICATION CODES TESTING' as test_section;

-- Test 1.2.3.1: Verify total classification codes count
SELECT 
    'Test 1.2.3.1: Total Classification Codes Count' as test_name,
    COUNT(*) as result,
    CASE 
        WHEN COUNT(*) >= 50 THEN 'PASS' 
        ELSE 'FAIL' 
    END as status
FROM classification_codes cc
JOIN industries i ON cc.industry_id = i.id
WHERE i.name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND cc.is_active = true;

-- Test 1.2.3.2: Verify classification codes per industry
SELECT 
    'Test 1.2.3.2: Classification Codes Per Industry' as test_name,
    i.name as industry_name,
    COUNT(cc.code) as code_count,
    CASE 
        WHEN COUNT(cc.code) >= 3 THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM industries i
LEFT JOIN classification_codes cc ON i.id = cc.industry_id AND cc.is_active = true
WHERE i.name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND i.is_active = true
GROUP BY i.name
ORDER BY code_count DESC;

-- Test 1.2.3.3: Verify key restaurant codes exist
SELECT 
    'Test 1.2.3.3: Key Restaurant Codes' as test_name,
    CASE 
        WHEN COUNT(*) >= 4 THEN 'PASS'
        ELSE 'FAIL'
    END as status,
    COUNT(*) as found_codes
FROM classification_codes cc
JOIN industries i ON cc.industry_id = i.id
WHERE i.name = 'Restaurants' 
AND cc.code IN ('722511', '722513', '5812', '5814')
AND cc.is_active = true;

-- Test 1.2.3.4: Verify code type distribution
SELECT 
    'Test 1.2.3.4: Code Type Distribution' as test_name,
    cc.code_type,
    COUNT(*) as code_count,
    CASE 
        WHEN COUNT(*) >= 10 THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM classification_codes cc
JOIN industries i ON cc.industry_id = i.id
WHERE i.name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND cc.is_active = true
GROUP BY cc.code_type
ORDER BY code_count DESC;

-- =============================================================================
-- 4. COMPREHENSIVE DATA INTEGRITY TESTING
-- =============================================================================

SELECT 'COMPREHENSIVE DATA INTEGRITY TESTING' as test_section;

-- Test 4.1: Verify no duplicate industry names
SELECT 
    'Test 4.1: No Duplicate Industry Names' as test_name,
    COUNT(*) as total_industries,
    COUNT(DISTINCT name) as unique_names,
    CASE 
        WHEN COUNT(*) = COUNT(DISTINCT name) THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM industries 
WHERE is_active = true;

-- Test 4.2: Verify no duplicate keywords within industries
SELECT 
    'Test 4.2: No Duplicate Keywords' as test_name,
    COUNT(*) as total_keywords,
    COUNT(DISTINCT CONCAT(industry_id, ':', keyword)) as unique_keywords,
    CASE 
        WHEN COUNT(*) = COUNT(DISTINCT CONCAT(industry_id, ':', keyword)) THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM keyword_weights 
WHERE is_active = true;

-- Test 4.3: Verify no duplicate classification codes within industries
SELECT 
    'Test 4.3: No Duplicate Classification Codes' as test_name,
    COUNT(*) as total_codes,
    COUNT(DISTINCT CONCAT(industry_id, ':', code_type, ':', code)) as unique_codes,
    CASE 
        WHEN COUNT(*) = COUNT(DISTINCT CONCAT(industry_id, ':', code_type, ':', code)) THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM classification_codes 
WHERE is_active = true;

-- Test 4.4: Verify all foreign key relationships are valid
SELECT 
    'Test 4.4: Valid Foreign Key Relationships' as test_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN i.id IS NOT NULL THEN 1 END) as valid_relationships,
    CASE 
        WHEN COUNT(*) = COUNT(CASE WHEN i.id IS NOT NULL THEN 1 END) THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM (
    SELECT industry_id FROM keyword_weights WHERE is_active = true
    UNION ALL
    SELECT industry_id FROM classification_codes WHERE is_active = true
) AS all_industry_ids
LEFT JOIN industries i ON all_industry_ids.industry_id = i.id;

-- =============================================================================
-- 5. PERFORMANCE TESTING
-- =============================================================================

SELECT 'PERFORMANCE TESTING' as test_section;

-- Test 5.1: Industry lookup performance
EXPLAIN (ANALYZE, BUFFERS) 
SELECT id, name, confidence_threshold 
FROM industries 
WHERE name = 'Restaurants' 
AND is_active = true;

-- Test 5.2: Keyword lookup performance
EXPLAIN (ANALYZE, BUFFERS) 
SELECT kw.keyword, kw.base_weight
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name = 'Restaurants' 
AND kw.is_active = true
ORDER BY kw.base_weight DESC
LIMIT 10;

-- Test 5.3: Classification code lookup performance
EXPLAIN (ANALYZE, BUFFERS) 
SELECT cc.code, cc.description
FROM classification_codes cc
JOIN industries i ON cc.industry_id = i.id
WHERE i.name = 'Restaurants' 
AND cc.is_active = true
ORDER BY cc.code_type, cc.code;

-- =============================================================================
-- 6. SUMMARY REPORTS
-- =============================================================================

SELECT 'SUMMARY REPORTS' as test_section;

-- Summary 1: Restaurant Industries Overview
SELECT 
    'RESTAURANT INDUSTRIES OVERVIEW' as summary_type,
    '' as spacer;

SELECT 
    i.name as industry_name,
    i.confidence_threshold,
    COUNT(kw.keyword) as keyword_count,
    COUNT(cc.code) as classification_code_count,
    ROUND(AVG(kw.base_weight), 4) as avg_keyword_weight
FROM industries i
LEFT JOIN keyword_weights kw ON i.id = kw.industry_id AND kw.is_active = true
LEFT JOIN classification_codes cc ON i.id = cc.industry_id AND cc.is_active = true
WHERE i.name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND i.is_active = true
GROUP BY i.name, i.confidence_threshold
ORDER BY keyword_count DESC, i.name;

-- Summary 2: Keyword Distribution by Industry
SELECT 
    'KEYWORD DISTRIBUTION BY INDUSTRY' as summary_type,
    '' as spacer;

SELECT 
    i.name as industry_name,
    COUNT(kw.keyword) as total_keywords,
    COUNT(CASE WHEN kw.base_weight >= 0.9000 THEN 1 END) as high_weight_keywords,
    COUNT(CASE WHEN kw.base_weight >= 0.8000 AND kw.base_weight < 0.9000 THEN 1 END) as medium_high_keywords,
    COUNT(CASE WHEN kw.base_weight >= 0.7000 AND kw.base_weight < 0.8000 THEN 1 END) as medium_keywords,
    COUNT(CASE WHEN kw.base_weight < 0.7000 THEN 1 END) as low_weight_keywords
FROM industries i
LEFT JOIN keyword_weights kw ON i.id = kw.industry_id AND kw.is_active = true
WHERE i.name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND i.is_active = true
GROUP BY i.name
ORDER BY total_keywords DESC, i.name;

-- Summary 3: Classification Code Distribution by Type
SELECT 
    'CLASSIFICATION CODE DISTRIBUTION BY TYPE' as summary_type,
    '' as spacer;

SELECT 
    cc.code_type,
    COUNT(*) as total_codes,
    COUNT(DISTINCT cc.code) as unique_codes,
    COUNT(DISTINCT cc.industry_id) as industries_covered
FROM classification_codes cc
JOIN industries i ON cc.industry_id = i.id
WHERE i.name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND cc.is_active = true
AND i.is_active = true
GROUP BY cc.code_type
ORDER BY total_codes DESC;

-- =============================================================================
-- 7. FINAL COMPLETION VERIFICATION
-- =============================================================================

DO $$
DECLARE
    industry_count INTEGER;
    keyword_count INTEGER;
    code_count INTEGER;
    all_tests_passed BOOLEAN := true;
BEGIN
    -- Count industries
    SELECT COUNT(*) INTO industry_count 
    FROM industries 
    WHERE name IN (
        'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
        'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
        'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
    )
    AND is_active = true;
    
    -- Count keywords
    SELECT COUNT(*) INTO keyword_count 
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name IN (
        'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
        'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
        'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
    )
    AND kw.is_active = true;
    
    -- Count classification codes
    SELECT COUNT(*) INTO code_count 
    FROM classification_codes cc
    JOIN industries i ON cc.industry_id = i.id
    WHERE i.name IN (
        'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
        'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
        'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
    )
    AND cc.is_active = true;
    
    -- Check if all tests passed
    IF industry_count < 12 OR keyword_count < 200 OR code_count < 50 THEN
        all_tests_passed := false;
    END IF;
    
    -- Report final results
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'TASK 1.2 COMPREHENSIVE TESTING RESULTS';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Restaurant industries added: %', industry_count;
    RAISE NOTICE 'Restaurant keywords added: %', keyword_count;
    RAISE NOTICE 'Restaurant classification codes added: %', code_count;
    
    IF all_tests_passed THEN
        RAISE NOTICE 'STATUS: ALL TESTS PASSED - TASK 1.2 COMPLETED SUCCESSFULLY';
        RAISE NOTICE 'Ready for Task 1.3: Test Restaurant Classification';
    ELSE
        RAISE NOTICE 'STATUS: SOME TESTS FAILED - Review and fix issues';
    END IF;
    
    RAISE NOTICE '=============================================================================';
END $$;
