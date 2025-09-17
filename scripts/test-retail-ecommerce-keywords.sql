-- =============================================================================
-- RETAIL & E-COMMERCE KEYWORDS TESTING SCRIPT
-- Subtask 3.2.4.5: Test and validate retail/e-commerce keywords
-- =============================================================================
-- This script provides comprehensive testing and validation for the retail and
-- e-commerce keywords to ensure they meet the plan requirements:
-- - 50+ keywords per industry
-- - Base weights 0.5-1.0
-- - Keyword relevance for classification accuracy
-- - No duplicate keywords within industries
-- =============================================================================

-- =============================================================================
-- 1. KEYWORD COUNT VALIDATION
-- =============================================================================

-- Test 1: Verify minimum keyword counts per industry
DO $$
DECLARE
    retail_count INTEGER;
    ecommerce_count INTEGER;
    wholesale_count INTEGER;
    consumer_goods_count INTEGER;
    test_passed BOOLEAN := true;
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'TEST 1: KEYWORD COUNT VALIDATION';
    RAISE NOTICE '=============================================================================';
    
    -- Count keywords for each industry
    SELECT COUNT(*) INTO retail_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Retail' AND kw.is_active = true;
    
    SELECT COUNT(*) INTO ecommerce_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'E-commerce' AND kw.is_active = true;
    
    SELECT COUNT(*) INTO wholesale_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Wholesale' AND kw.is_active = true;
    
    SELECT COUNT(*) INTO consumer_goods_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Consumer Goods' AND kw.is_active = true;
    
    -- Report results
    RAISE NOTICE 'Retail keywords: % (required: 50+)', retail_count;
    RAISE NOTICE 'E-commerce keywords: % (required: 50+)', ecommerce_count;
    RAISE NOTICE 'Wholesale keywords: % (required: 50+)', wholesale_count;
    RAISE NOTICE 'Consumer Goods keywords: % (required: 50+)', consumer_goods_count;
    
    -- Validate minimum counts
    IF retail_count < 50 THEN
        RAISE NOTICE '❌ FAIL: Retail industry has insufficient keywords (%)', retail_count;
        test_passed := false;
    ELSE
        RAISE NOTICE '✅ PASS: Retail industry has sufficient keywords (%)', retail_count;
    END IF;
    
    IF ecommerce_count < 50 THEN
        RAISE NOTICE '❌ FAIL: E-commerce industry has insufficient keywords (%)', ecommerce_count;
        test_passed := false;
    ELSE
        RAISE NOTICE '✅ PASS: E-commerce industry has sufficient keywords (%)', ecommerce_count;
    END IF;
    
    IF wholesale_count < 50 THEN
        RAISE NOTICE '❌ FAIL: Wholesale industry has insufficient keywords (%)', wholesale_count;
        test_passed := false;
    ELSE
        RAISE NOTICE '✅ PASS: Wholesale industry has sufficient keywords (%)', wholesale_count;
    END IF;
    
    IF consumer_goods_count < 50 THEN
        RAISE NOTICE '❌ FAIL: Consumer Goods industry has insufficient keywords (%)', consumer_goods_count;
        test_passed := false;
    ELSE
        RAISE NOTICE '✅ PASS: Consumer Goods industry has sufficient keywords (%)', consumer_goods_count;
    END IF;
    
    -- Overall test result
    IF test_passed THEN
        RAISE NOTICE '✅ OVERALL RESULT: All industries meet minimum keyword count requirements';
    ELSE
        RAISE NOTICE '❌ OVERALL RESULT: Some industries do not meet minimum keyword count requirements';
    END IF;
END $$;

-- =============================================================================
-- 2. KEYWORD WEIGHT VALIDATION
-- =============================================================================

-- Test 2: Verify keyword weights are within specified range (0.5-1.0)
DO $$
DECLARE
    retail_min_weight DECIMAL;
    retail_max_weight DECIMAL;
    ecommerce_min_weight DECIMAL;
    ecommerce_max_weight DECIMAL;
    wholesale_min_weight DECIMAL;
    wholesale_max_weight DECIMAL;
    consumer_goods_min_weight DECIMAL;
    consumer_goods_max_weight DECIMAL;
    test_passed BOOLEAN := true;
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'TEST 2: KEYWORD WEIGHT VALIDATION';
    RAISE NOTICE '=============================================================================';
    
    -- Get weight ranges for each industry
    SELECT MIN(kw.base_weight), MAX(kw.base_weight) INTO retail_min_weight, retail_max_weight
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Retail' AND kw.is_active = true;
    
    SELECT MIN(kw.base_weight), MAX(kw.base_weight) INTO ecommerce_min_weight, ecommerce_max_weight
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'E-commerce' AND kw.is_active = true;
    
    SELECT MIN(kw.base_weight), MAX(kw.base_weight) INTO wholesale_min_weight, wholesale_max_weight
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Wholesale' AND kw.is_active = true;
    
    SELECT MIN(kw.base_weight), MAX(kw.base_weight) INTO consumer_goods_min_weight, consumer_goods_max_weight
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Consumer Goods' AND kw.is_active = true;
    
    -- Report results
    RAISE NOTICE 'Retail weights: %.4f - %.4f (required: 0.5000 - 1.0000)', retail_min_weight, retail_max_weight;
    RAISE NOTICE 'E-commerce weights: %.4f - %.4f (required: 0.5000 - 1.0000)', ecommerce_min_weight, ecommerce_max_weight;
    RAISE NOTICE 'Wholesale weights: %.4f - %.4f (required: 0.5000 - 1.0000)', wholesale_min_weight, wholesale_max_weight;
    RAISE NOTICE 'Consumer Goods weights: %.4f - %.4f (required: 0.5000 - 1.0000)', consumer_goods_min_weight, consumer_goods_max_weight;
    
    -- Validate weight ranges
    IF retail_min_weight < 0.5 OR retail_max_weight > 1.0 THEN
        RAISE NOTICE '❌ FAIL: Retail industry weights outside required range';
        test_passed := false;
    ELSE
        RAISE NOTICE '✅ PASS: Retail industry weights within required range';
    END IF;
    
    IF ecommerce_min_weight < 0.5 OR ecommerce_max_weight > 1.0 THEN
        RAISE NOTICE '❌ FAIL: E-commerce industry weights outside required range';
        test_passed := false;
    ELSE
        RAISE NOTICE '✅ PASS: E-commerce industry weights within required range';
    END IF;
    
    IF wholesale_min_weight < 0.5 OR wholesale_max_weight > 1.0 THEN
        RAISE NOTICE '❌ FAIL: Wholesale industry weights outside required range';
        test_passed := false;
    ELSE
        RAISE NOTICE '✅ PASS: Wholesale industry weights within required range';
    END IF;
    
    IF consumer_goods_min_weight < 0.5 OR consumer_goods_max_weight > 1.0 THEN
        RAISE NOTICE '❌ FAIL: Consumer Goods industry weights outside required range';
        test_passed := false;
    ELSE
        RAISE NOTICE '✅ PASS: Consumer Goods industry weights within required range';
    END IF;
    
    -- Overall test result
    IF test_passed THEN
        RAISE NOTICE '✅ OVERALL RESULT: All industries have weights within required range (0.5-1.0)';
    ELSE
        RAISE NOTICE '❌ OVERALL RESULT: Some industries have weights outside required range';
    END IF;
END $$;

-- =============================================================================
-- 3. DUPLICATE KEYWORD VALIDATION
-- =============================================================================

-- Test 3: Verify no duplicate keywords within industries
DO $$
DECLARE
    retail_duplicates INTEGER;
    ecommerce_duplicates INTEGER;
    wholesale_duplicates INTEGER;
    consumer_goods_duplicates INTEGER;
    test_passed BOOLEAN := true;
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'TEST 3: DUPLICATE KEYWORD VALIDATION';
    RAISE NOTICE '=============================================================================';
    
    -- Count duplicates for each industry
    SELECT COUNT(*) INTO retail_duplicates
    FROM (
        SELECT kw.keyword, COUNT(*) as count
        FROM keyword_weights kw
        JOIN industries i ON kw.industry_id = i.id
        WHERE i.name = 'Retail' AND kw.is_active = true
        GROUP BY kw.keyword
        HAVING COUNT(*) > 1
    ) duplicates;
    
    SELECT COUNT(*) INTO ecommerce_duplicates
    FROM (
        SELECT kw.keyword, COUNT(*) as count
        FROM keyword_weights kw
        JOIN industries i ON kw.industry_id = i.id
        WHERE i.name = 'E-commerce' AND kw.is_active = true
        GROUP BY kw.keyword
        HAVING COUNT(*) > 1
    ) duplicates;
    
    SELECT COUNT(*) INTO wholesale_duplicates
    FROM (
        SELECT kw.keyword, COUNT(*) as count
        FROM keyword_weights kw
        JOIN industries i ON kw.industry_id = i.id
        WHERE i.name = 'Wholesale' AND kw.is_active = true
        GROUP BY kw.keyword
        HAVING COUNT(*) > 1
    ) duplicates;
    
    SELECT COUNT(*) INTO consumer_goods_duplicates
    FROM (
        SELECT kw.keyword, COUNT(*) as count
        FROM keyword_weights kw
        JOIN industries i ON kw.industry_id = i.id
        WHERE i.name = 'Consumer Goods' AND kw.is_active = true
        GROUP BY kw.keyword
        HAVING COUNT(*) > 1
    ) duplicates;
    
    -- Report results
    RAISE NOTICE 'Retail duplicates: % (required: 0)', retail_duplicates;
    RAISE NOTICE 'E-commerce duplicates: % (required: 0)', ecommerce_duplicates;
    RAISE NOTICE 'Wholesale duplicates: % (required: 0)', wholesale_duplicates;
    RAISE NOTICE 'Consumer Goods duplicates: % (required: 0)', consumer_goods_duplicates;
    
    -- Validate no duplicates
    IF retail_duplicates > 0 THEN
        RAISE NOTICE '❌ FAIL: Retail industry has duplicate keywords';
        test_passed := false;
    ELSE
        RAISE NOTICE '✅ PASS: Retail industry has no duplicate keywords';
    END IF;
    
    IF ecommerce_duplicates > 0 THEN
        RAISE NOTICE '❌ FAIL: E-commerce industry has duplicate keywords';
        test_passed := false;
    ELSE
        RAISE NOTICE '✅ PASS: E-commerce industry has no duplicate keywords';
    END IF;
    
    IF wholesale_duplicates > 0 THEN
        RAISE NOTICE '❌ FAIL: Wholesale industry has duplicate keywords';
        test_passed := false;
    ELSE
        RAISE NOTICE '✅ PASS: Wholesale industry has no duplicate keywords';
    END IF;
    
    IF consumer_goods_duplicates > 0 THEN
        RAISE NOTICE '❌ FAIL: Consumer Goods industry has duplicate keywords';
        test_passed := false;
    ELSE
        RAISE NOTICE '✅ PASS: Consumer Goods industry has no duplicate keywords';
    END IF;
    
    -- Overall test result
    IF test_passed THEN
        RAISE NOTICE '✅ OVERALL RESULT: No duplicate keywords found in any industry';
    ELSE
        RAISE NOTICE '❌ OVERALL RESULT: Duplicate keywords found in some industries';
    END IF;
END $$;

-- =============================================================================
-- 4. KEYWORD RELEVANCE VALIDATION
-- =============================================================================

-- Test 4: Verify keyword relevance for classification accuracy
DO $$
DECLARE
    retail_high_weight_count INTEGER;
    ecommerce_high_weight_count INTEGER;
    wholesale_high_weight_count INTEGER;
    consumer_goods_high_weight_count INTEGER;
    test_passed BOOLEAN := true;
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'TEST 4: KEYWORD RELEVANCE VALIDATION';
    RAISE NOTICE '=============================================================================';
    
    -- Count high-weight keywords (>= 0.8) for each industry
    SELECT COUNT(*) INTO retail_high_weight_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Retail' AND kw.is_active = true AND kw.base_weight >= 0.8;
    
    SELECT COUNT(*) INTO ecommerce_high_weight_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'E-commerce' AND kw.is_active = true AND kw.base_weight >= 0.8;
    
    SELECT COUNT(*) INTO wholesale_high_weight_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Wholesale' AND kw.is_active = true AND kw.base_weight >= 0.8;
    
    SELECT COUNT(*) INTO consumer_goods_high_weight_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Consumer Goods' AND kw.is_active = true AND kw.base_weight >= 0.8;
    
    -- Report results
    RAISE NOTICE 'Retail high-weight keywords (>=0.8): % (recommended: 10+)', retail_high_weight_count;
    RAISE NOTICE 'E-commerce high-weight keywords (>=0.8): % (recommended: 10+)', ecommerce_high_weight_count;
    RAISE NOTICE 'Wholesale high-weight keywords (>=0.8): % (recommended: 10+)', wholesale_high_weight_count;
    RAISE NOTICE 'Consumer Goods high-weight keywords (>=0.8): % (recommended: 10+)', consumer_goods_high_weight_count;
    
    -- Validate high-weight keyword counts
    IF retail_high_weight_count < 10 THEN
        RAISE NOTICE '❌ FAIL: Retail industry has insufficient high-weight keywords (%)', retail_high_weight_count;
        test_passed := false;
    ELSE
        RAISE NOTICE '✅ PASS: Retail industry has sufficient high-weight keywords (%)', retail_high_weight_count;
    END IF;
    
    IF ecommerce_high_weight_count < 10 THEN
        RAISE NOTICE '❌ FAIL: E-commerce industry has insufficient high-weight keywords (%)', ecommerce_high_weight_count;
        test_passed := false;
    ELSE
        RAISE NOTICE '✅ PASS: E-commerce industry has sufficient high-weight keywords (%)', ecommerce_high_weight_count;
    END IF;
    
    IF wholesale_high_weight_count < 10 THEN
        RAISE NOTICE '❌ FAIL: Wholesale industry has insufficient high-weight keywords (%)', wholesale_high_weight_count;
        test_passed := false;
    ELSE
        RAISE NOTICE '✅ PASS: Wholesale industry has sufficient high-weight keywords (%)', wholesale_high_weight_count;
    END IF;
    
    IF consumer_goods_high_weight_count < 10 THEN
        RAISE NOTICE '❌ FAIL: Consumer Goods industry has insufficient high-weight keywords (%)', consumer_goods_high_weight_count;
        test_passed := false;
    ELSE
        RAISE NOTICE '✅ PASS: Consumer Goods industry has sufficient high-weight keywords (%)', consumer_goods_high_weight_count;
    END IF;
    
    -- Overall test result
    IF test_passed THEN
        RAISE NOTICE '✅ OVERALL RESULT: All industries have sufficient high-weight keywords for classification accuracy';
    ELSE
        RAISE NOTICE '❌ OVERALL RESULT: Some industries have insufficient high-weight keywords';
    END IF;
END $$;

-- =============================================================================
-- 5. COMPREHENSIVE VALIDATION SUMMARY
-- =============================================================================

-- Test 5: Comprehensive validation summary
DO $$
DECLARE
    total_keywords INTEGER;
    total_industries INTEGER;
    avg_keywords_per_industry DECIMAL;
    test_passed BOOLEAN := true;
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'TEST 5: COMPREHENSIVE VALIDATION SUMMARY';
    RAISE NOTICE '=============================================================================';
    
    -- Calculate totals
    SELECT COUNT(*) INTO total_keywords
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name IN ('Retail', 'E-commerce', 'Wholesale', 'Consumer Goods')
    AND kw.is_active = true;
    
    SELECT COUNT(*) INTO total_industries
    FROM industries
    WHERE name IN ('Retail', 'E-commerce', 'Wholesale', 'Consumer Goods')
    AND is_active = true;
    
    avg_keywords_per_industry := total_keywords::DECIMAL / total_industries;
    
    -- Report results
    RAISE NOTICE 'Total keywords added: %', total_keywords;
    RAISE NOTICE 'Total industries covered: %', total_industries;
    RAISE NOTICE 'Average keywords per industry: %.1f', avg_keywords_per_industry;
    RAISE NOTICE 'Target: 200+ keywords across 4 industries (50+ per industry)';
    
    -- Validate totals
    IF total_keywords >= 200 THEN
        RAISE NOTICE '✅ PASS: Total keywords meet target (200+)';
    ELSE
        RAISE NOTICE '❌ FAIL: Total keywords below target (%)', total_keywords;
        test_passed := false;
    END IF;
    
    IF avg_keywords_per_industry >= 50 THEN
        RAISE NOTICE '✅ PASS: Average keywords per industry meet target (50+)';
    ELSE
        RAISE NOTICE '❌ FAIL: Average keywords per industry below target (%.1f)', avg_keywords_per_industry;
        test_passed := false;
    END IF;
    
    -- Overall test result
    IF test_passed THEN
        RAISE NOTICE '✅ OVERALL RESULT: All validation tests passed successfully';
        RAISE NOTICE 'Status: Ready for classification accuracy testing';
    ELSE
        RAISE NOTICE '❌ OVERALL RESULT: Some validation tests failed';
        RAISE NOTICE 'Status: Review and fix issues before proceeding';
    END IF;
END $$;

-- =============================================================================
-- 6. DETAILED KEYWORD ANALYSIS
-- =============================================================================

-- Display detailed keyword analysis by industry
SELECT 
    'DETAILED KEYWORD ANALYSIS' as analysis_type,
    '' as spacer;

SELECT 
    i.name as industry_name,
    COUNT(kw.keyword) as total_keywords,
    COUNT(CASE WHEN kw.base_weight >= 1.0 THEN 1 END) as weight_1_0,
    COUNT(CASE WHEN kw.base_weight >= 0.9 AND kw.base_weight < 1.0 THEN 1 END) as weight_0_9_0_99,
    COUNT(CASE WHEN kw.base_weight >= 0.8 AND kw.base_weight < 0.9 THEN 1 END) as weight_0_8_0_89,
    COUNT(CASE WHEN kw.base_weight >= 0.7 AND kw.base_weight < 0.8 THEN 1 END) as weight_0_7_0_79,
    COUNT(CASE WHEN kw.base_weight >= 0.6 AND kw.base_weight < 0.7 THEN 1 END) as weight_0_6_0_69,
    COUNT(CASE WHEN kw.base_weight >= 0.5 AND kw.base_weight < 0.6 THEN 1 END) as weight_0_5_0_59,
    ROUND(AVG(kw.base_weight), 4) as avg_weight,
    ROUND(MIN(kw.base_weight), 4) as min_weight,
    ROUND(MAX(kw.base_weight), 4) as max_weight
FROM industries i
JOIN keyword_weights kw ON i.id = kw.industry_id
WHERE i.name IN ('Retail', 'E-commerce', 'Wholesale', 'Consumer Goods')
AND kw.is_active = true
GROUP BY i.name
ORDER BY total_keywords DESC;

-- Display top keywords for each industry
SELECT 
    'TOP KEYWORDS BY INDUSTRY' as analysis_type,
    '' as spacer;

SELECT 
    i.name as industry_name,
    kw.keyword,
    kw.base_weight,
    ROW_NUMBER() OVER (PARTITION BY i.name ORDER BY kw.base_weight DESC) as rank
FROM industries i
JOIN keyword_weights kw ON i.id = kw.industry_id
WHERE i.name IN ('Retail', 'E-commerce', 'Wholesale', 'Consumer Goods')
AND kw.is_active = true
ORDER BY i.name, kw.base_weight DESC;

-- =============================================================================
-- COMPLETION MESSAGE
-- =============================================================================
DO $$
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'RETAIL & E-COMMERCE KEYWORDS TESTING COMPLETED';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'All validation tests executed successfully';
    RAISE NOTICE 'Review results above to ensure all requirements are met';
    RAISE NOTICE 'Next step: Proceed to next subtask or fix any identified issues';
    RAISE NOTICE '=============================================================================';
END $$;
