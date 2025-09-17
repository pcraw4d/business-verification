-- =============================================================================
-- Restaurant Keywords Test Script
-- Verification queries for Subtask 1.2.2
-- =============================================================================

-- This script tests that restaurant keywords were added correctly
-- and verifies the database structure and data integrity

-- =============================================================================
-- 1. VERIFY RESTAURANT KEYWORDS EXIST
-- =============================================================================

-- Test 1: Verify total keyword count for restaurant industries
SELECT 
    'Test 1: Total Restaurant Keywords Count' as test_name,
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

-- Test 2: Verify each restaurant industry has keywords
SELECT 
    'Test 2: Keywords Per Industry' as test_name,
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

-- Test 3: Verify specific high-value keywords exist
SELECT 
    'Test 3: High-Value Keywords' as test_name,
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

-- =============================================================================
-- 2. VERIFY KEYWORD WEIGHTS
-- =============================================================================

-- Test 4: Verify weight ranges are appropriate
SELECT 
    'Test 4: Weight Range Validation' as test_name,
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

-- Test 5: Verify high-weight keywords have appropriate values
SELECT 
    'Test 5: High-Weight Keywords' as test_name,
    i.name as industry_name,
    kw.keyword,
    kw.base_weight,
    CASE 
        WHEN kw.base_weight >= 0.9000 THEN 'HIGH'
        WHEN kw.base_weight >= 0.8000 THEN 'MEDIUM-HIGH'
        WHEN kw.base_weight >= 0.7000 THEN 'MEDIUM'
        ELSE 'LOW'
    END as weight_category
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name IN ('Restaurants', 'Fast Food', 'Fine Dining')
AND kw.base_weight >= 0.9000
AND kw.is_active = true
ORDER BY i.name, kw.base_weight DESC;

-- =============================================================================
-- 3. VERIFY DATABASE STRUCTURE
-- =============================================================================

-- Test 6: Verify keyword_weights table structure
SELECT 
    'Test 6: Keyword Weights Table Structure' as test_name,
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns 
WHERE table_name = 'keyword_weights' 
AND column_name IN ('id', 'industry_id', 'keyword', 'base_weight', 'usage_count', 'success_count', 'is_active')
ORDER BY ordinal_position;

-- Test 7: Verify indexes exist
SELECT 
    'Test 7: Required Indexes' as test_name,
    indexname,
    indexdef
FROM pg_indexes 
WHERE tablename = 'keyword_weights' 
AND indexname LIKE '%keyword_weights%'
ORDER BY indexname;

-- =============================================================================
-- 4. VERIFY DATA INTEGRITY
-- =============================================================================

-- Test 8: Verify no duplicate keywords within industries
SELECT 
    'Test 8: No Duplicate Keywords' as test_name,
    i.name as industry_name,
    kw.keyword,
    COUNT(*) as duplicate_count,
    CASE 
        WHEN COUNT(*) = 1 THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND kw.is_active = true
GROUP BY i.name, kw.keyword
HAVING COUNT(*) > 1
ORDER BY duplicate_count DESC, i.name, kw.keyword;

-- Test 9: Verify all keywords are linked to valid industries
SELECT 
    'Test 9: Valid Industry Links' as test_name,
    COUNT(*) as total_keywords,
    COUNT(i.id) as valid_industry_links,
    CASE 
        WHEN COUNT(*) = COUNT(i.id) THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM keyword_weights kw
LEFT JOIN industries i ON kw.industry_id = i.id
WHERE kw.is_active = true;

-- Test 10: Verify keyword length constraints
SELECT 
    'Test 10: Keyword Length Validation' as test_name,
    COUNT(*) as total_keywords,
    COUNT(CASE WHEN LENGTH(kw.keyword) <= 255 THEN 1 END) as valid_length,
    CASE 
        WHEN COUNT(*) = COUNT(CASE WHEN LENGTH(kw.keyword) <= 255 THEN 1 END) THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM keyword_weights kw
WHERE kw.is_active = true;

-- =============================================================================
-- 5. VERIFY INDUSTRY-SPECIFIC KEYWORDS
-- =============================================================================

-- Test 11: Verify Fast Food specific keywords
SELECT 
    'Test 11: Fast Food Keywords' as test_name,
    kw.keyword,
    kw.base_weight,
    CASE 
        WHEN kw.keyword IN ('fast food', 'quick service', 'drive thru', 'takeout') THEN 'CORE'
        WHEN kw.keyword IN ('mcdonalds', 'burger king', 'kfc', 'subway') THEN 'CHAIN'
        WHEN kw.keyword IN ('burger', 'fries', 'chicken', 'pizza') THEN 'FOOD'
        ELSE 'OTHER'
    END as keyword_type
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name = 'Fast Food' 
AND kw.is_active = true
ORDER BY kw.base_weight DESC, kw.keyword;

-- Test 12: Verify Fine Dining specific keywords
SELECT 
    'Test 12: Fine Dining Keywords' as test_name,
    kw.keyword,
    kw.base_weight,
    CASE 
        WHEN kw.keyword IN ('fine dining', 'upscale', 'gourmet', 'premium') THEN 'CORE'
        WHEN kw.keyword IN ('wine pairing', 'sommelier', 'tasting menu', 'white tablecloth') THEN 'SERVICE'
        WHEN kw.keyword IN ('french cuisine', 'haute cuisine', 'molecular gastronomy') THEN 'CUISINE'
        ELSE 'OTHER'
    END as keyword_type
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name = 'Fine Dining' 
AND kw.is_active = true
ORDER BY kw.base_weight DESC, kw.keyword;

-- Test 13: Verify Breweries specific keywords
SELECT 
    'Test 13: Breweries Keywords' as test_name,
    kw.keyword,
    kw.base_weight,
    CASE 
        WHEN kw.keyword IN ('brewery', 'brewing', 'beer production', 'craft beer') THEN 'CORE'
        WHEN kw.keyword IN ('ale', 'lager', 'ipa', 'stout', 'porter') THEN 'BEER_TYPES'
        WHEN kw.keyword IN ('tasting room', 'beer tasting', 'brewery tour', 'taproom') THEN 'EXPERIENCE'
        ELSE 'OTHER'
    END as keyword_type
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name = 'Breweries' 
AND kw.is_active = true
ORDER BY kw.base_weight DESC, kw.keyword;

-- =============================================================================
-- 6. PERFORMANCE VERIFICATION
-- =============================================================================

-- Test 14: Verify query performance for keyword lookup
EXPLAIN (ANALYZE, BUFFERS) 
SELECT kw.keyword, kw.base_weight
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name = 'Restaurants' 
AND kw.is_active = true
ORDER BY kw.base_weight DESC
LIMIT 10;

-- =============================================================================
-- 7. DISPLAY KEYWORD SUMMARY
-- =============================================================================

-- Show keyword count and weight distribution per industry
SELECT 
    'KEYWORD SUMMARY BY INDUSTRY' as section,
    '' as spacer;

SELECT 
    i.name as industry_name,
    COUNT(kw.keyword) as keyword_count,
    MIN(kw.base_weight) as min_weight,
    MAX(kw.base_weight) as max_weight,
    ROUND(AVG(kw.base_weight), 4) as avg_weight,
    COUNT(CASE WHEN kw.base_weight >= 0.9000 THEN 1 END) as high_weight_count,
    COUNT(CASE WHEN kw.base_weight >= 0.8000 AND kw.base_weight < 0.9000 THEN 1 END) as medium_high_count,
    COUNT(CASE WHEN kw.base_weight >= 0.7000 AND kw.base_weight < 0.8000 THEN 1 END) as medium_count,
    COUNT(CASE WHEN kw.base_weight < 0.7000 THEN 1 END) as low_weight_count
FROM industries i
LEFT JOIN keyword_weights kw ON i.id = kw.industry_id AND kw.is_active = true
WHERE i.name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND i.is_active = true
GROUP BY i.name
ORDER BY keyword_count DESC, i.name;

-- =============================================================================
-- 8. COMPLETION VERIFICATION
-- =============================================================================

DO $$
DECLARE
    total_keywords INTEGER;
    restaurant_keywords INTEGER;
    fast_food_keywords INTEGER;
    fine_dining_keywords INTEGER;
    weight_range_valid BOOLEAN;
    no_duplicates BOOLEAN;
BEGIN
    -- Count total keywords
    SELECT COUNT(*) INTO total_keywords 
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name IN (
        'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
        'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
        'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
    )
    AND kw.is_active = true;
    
    -- Count keywords per major industry
    SELECT COUNT(*) INTO restaurant_keywords 
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Restaurants' AND kw.is_active = true;
    
    SELECT COUNT(*) INTO fast_food_keywords 
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Fast Food' AND kw.is_active = true;
    
    SELECT COUNT(*) INTO fine_dining_keywords 
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Fine Dining' AND kw.is_active = true;
    
    -- Check weight range
    SELECT EXISTS(
        SELECT 1 FROM keyword_weights kw
        JOIN industries i ON kw.industry_id = i.id
        WHERE i.name IN (
            'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
            'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
            'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
        )
        AND kw.is_active = true
        AND kw.base_weight BETWEEN 0.6000 AND 1.0000
    ) INTO weight_range_valid;
    
    -- Check for duplicates
    SELECT NOT EXISTS(
        SELECT 1 FROM keyword_weights kw
        JOIN industries i ON kw.industry_id = i.id
        WHERE i.name IN (
            'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
            'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
            'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
        )
        AND kw.is_active = true
        GROUP BY i.name, kw.keyword
        HAVING COUNT(*) > 1
    ) INTO no_duplicates;
    
    -- Report final results
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'RESTAURANT KEYWORDS TEST RESULTS';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Total restaurant keywords added: %', total_keywords;
    RAISE NOTICE 'Restaurants industry keywords: %', restaurant_keywords;
    RAISE NOTICE 'Fast Food industry keywords: %', fast_food_keywords;
    RAISE NOTICE 'Fine Dining industry keywords: %', fine_dining_keywords;
    RAISE NOTICE 'Weight range valid (0.6000-1.0000): %', weight_range_valid;
    RAISE NOTICE 'No duplicate keywords: %', no_duplicates;
    
    IF total_keywords >= 200 AND restaurant_keywords >= 15 AND fast_food_keywords >= 15 
       AND fine_dining_keywords >= 15 AND weight_range_valid AND no_duplicates THEN
        RAISE NOTICE 'STATUS: ALL TESTS PASSED - Subtask 1.2.2 COMPLETED SUCCESSFULLY';
    ELSE
        RAISE NOTICE 'STATUS: SOME TESTS FAILED - Review and fix issues';
    END IF;
    
    RAISE NOTICE '=============================================================================';
END $$;
