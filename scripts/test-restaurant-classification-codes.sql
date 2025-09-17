-- =============================================================================
-- Restaurant Classification Codes Test Script
-- Verification queries for Subtask 1.2.3
-- =============================================================================

-- This script tests that restaurant classification codes were added correctly
-- and verifies the database structure and data integrity

-- =============================================================================
-- 1. VERIFY RESTAURANT CLASSIFICATION CODES EXIST
-- =============================================================================

-- Test 1: Verify total classification codes count for restaurant industries
SELECT 
    'Test 1: Total Restaurant Classification Codes Count' as test_name,
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

-- Test 2: Verify each restaurant industry has classification codes
SELECT 
    'Test 2: Classification Codes Per Industry' as test_name,
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

-- Test 3: Verify specific high-value codes exist
SELECT 
    'Test 3: Key Restaurant Codes' as test_name,
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

-- =============================================================================
-- 2. VERIFY CODE TYPES AND DISTRIBUTION
-- =============================================================================

-- Test 4: Verify code type distribution
SELECT 
    'Test 4: Code Type Distribution' as test_name,
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

-- Test 5: Verify NAICS codes for restaurants
SELECT 
    'Test 5: NAICS Restaurant Codes' as test_name,
    cc.code,
    cc.description,
    i.name as industry_name,
    CASE 
        WHEN cc.code IN ('722511', '722513', '722514', '722515', '722410') THEN 'CORE'
        WHEN cc.code IN ('722310', '722320', '722330') THEN 'SERVICE'
        ELSE 'OTHER'
    END as code_category
FROM classification_codes cc
JOIN industries i ON cc.industry_id = i.id
WHERE cc.code_type = 'NAICS'
AND i.name IN ('Restaurants', 'Fast Food', 'Fine Dining', 'Casual Dining')
AND cc.is_active = true
ORDER BY cc.code, i.name;

-- Test 6: Verify SIC codes for restaurants
SELECT 
    'Test 6: SIC Restaurant Codes' as test_name,
    cc.code,
    cc.description,
    i.name as industry_name,
    CASE 
        WHEN cc.code IN ('5812', '5813', '5814') THEN 'CORE'
        WHEN cc.code = '5819' THEN 'MISC'
        ELSE 'OTHER'
    END as code_category
FROM classification_codes cc
JOIN industries i ON cc.industry_id = i.id
WHERE cc.code_type = 'SIC'
AND i.name IN ('Restaurants', 'Fast Food', 'Fine Dining', 'Casual Dining')
AND cc.is_active = true
ORDER BY cc.code, i.name;

-- Test 7: Verify MCC codes for restaurants
SELECT 
    'Test 7: MCC Restaurant Codes' as test_name,
    cc.code,
    cc.description,
    i.name as industry_name,
    CASE 
        WHEN cc.code IN ('5812', '5813', '5814') THEN 'CORE'
        WHEN cc.code IN ('5815', '5816', '5817', '5818', '5819') THEN 'DIGITAL'
        ELSE 'OTHER'
    END as code_category
FROM classification_codes cc
JOIN industries i ON cc.industry_id = i.id
WHERE cc.code_type = 'MCC'
AND i.name IN ('Restaurants', 'Fast Food', 'Fine Dining', 'Casual Dining')
AND cc.is_active = true
ORDER BY cc.code, i.name;

-- =============================================================================
-- 3. VERIFY DATABASE STRUCTURE
-- =============================================================================

-- Test 8: Verify classification_codes table structure
SELECT 
    'Test 8: Classification Codes Table Structure' as test_name,
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns 
WHERE table_name = 'classification_codes' 
AND column_name IN ('id', 'industry_id', 'code_type', 'code', 'description', 'is_active')
ORDER BY ordinal_position;

-- Test 9: Verify indexes exist
SELECT 
    'Test 9: Required Indexes' as test_name,
    indexname,
    indexdef
FROM pg_indexes 
WHERE tablename = 'classification_codes' 
AND indexname LIKE '%classification_codes%'
ORDER BY indexname;

-- =============================================================================
-- 4. VERIFY DATA INTEGRITY
-- =============================================================================

-- Test 10: Verify no duplicate codes within industries
SELECT 
    'Test 10: No Duplicate Codes' as test_name,
    i.name as industry_name,
    cc.code_type,
    cc.code,
    COUNT(*) as duplicate_count,
    CASE 
        WHEN COUNT(*) = 1 THEN 'PASS'
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
GROUP BY i.name, cc.code_type, cc.code
HAVING COUNT(*) > 1
ORDER BY duplicate_count DESC, i.name, cc.code_type, cc.code;

-- Test 11: Verify all codes are linked to valid industries
SELECT 
    'Test 11: Valid Industry Links' as test_name,
    COUNT(*) as total_codes,
    COUNT(i.id) as valid_industry_links,
    CASE 
        WHEN COUNT(*) = COUNT(i.id) THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM classification_codes cc
LEFT JOIN industries i ON cc.industry_id = i.id
WHERE cc.is_active = true;

-- Test 12: Verify code format constraints
SELECT 
    'Test 12: Code Format Validation' as test_name,
    COUNT(*) as total_codes,
    COUNT(CASE WHEN LENGTH(cc.code) <= 20 THEN 1 END) as valid_length,
    COUNT(CASE WHEN cc.code_type IN ('NAICS', 'SIC', 'MCC') THEN 1 END) as valid_types,
    CASE 
        WHEN COUNT(*) = COUNT(CASE WHEN LENGTH(cc.code) <= 20 THEN 1 END) 
         AND COUNT(*) = COUNT(CASE WHEN cc.code_type IN ('NAICS', 'SIC', 'MCC') THEN 1 END) THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM classification_codes cc
WHERE cc.is_active = true;

-- =============================================================================
-- 5. VERIFY INDUSTRY-SPECIFIC CODES
-- =============================================================================

-- Test 13: Verify Fast Food specific codes
SELECT 
    'Test 13: Fast Food Codes' as test_name,
    cc.code_type,
    cc.code,
    cc.description,
    CASE 
        WHEN cc.code IN ('722513', '722515', '5814') THEN 'CORE'
        WHEN cc.code IN ('722330', '5812') THEN 'SUPPORTING'
        ELSE 'OTHER'
    END as code_importance
FROM classification_codes cc
JOIN industries i ON cc.industry_id = i.id
WHERE i.name = 'Fast Food' 
AND cc.is_active = true
ORDER BY cc.code_type, cc.code;

-- Test 14: Verify Fine Dining specific codes
SELECT 
    'Test 14: Fine Dining Codes' as test_name,
    cc.code_type,
    cc.code,
    cc.description,
    CASE 
        WHEN cc.code IN ('722511', '722410', '5812', '5813') THEN 'CORE'
        ELSE 'OTHER'
    END as code_importance
FROM classification_codes cc
JOIN industries i ON cc.industry_id = i.id
WHERE i.name = 'Fine Dining' 
AND cc.is_active = true
ORDER BY cc.code_type, cc.code;

-- Test 15: Verify Breweries specific codes
SELECT 
    'Test 15: Breweries Codes' as test_name,
    cc.code_type,
    cc.code,
    cc.description,
    CASE 
        WHEN cc.code IN ('722410', '312120', '5813', '2082') THEN 'CORE'
        ELSE 'OTHER'
    END as code_importance
FROM classification_codes cc
JOIN industries i ON cc.industry_id = i.id
WHERE i.name = 'Breweries' 
AND cc.is_active = true
ORDER BY cc.code_type, cc.code;

-- =============================================================================
-- 6. PERFORMANCE VERIFICATION
-- =============================================================================

-- Test 16: Verify query performance for code lookup
EXPLAIN (ANALYZE, BUFFERS) 
SELECT cc.code, cc.description
FROM classification_codes cc
JOIN industries i ON cc.industry_id = i.id
WHERE i.name = 'Restaurants' 
AND cc.is_active = true
ORDER BY cc.code_type, cc.code;

-- =============================================================================
-- 7. DISPLAY CLASSIFICATION CODES SUMMARY
-- =============================================================================

-- Show classification codes count and distribution per industry
SELECT 
    'CLASSIFICATION CODES SUMMARY BY INDUSTRY' as section,
    '' as spacer;

SELECT 
    i.name as industry_name,
    COUNT(cc.code) as total_codes,
    COUNT(CASE WHEN cc.code_type = 'NAICS' THEN 1 END) as naics_codes,
    COUNT(CASE WHEN cc.code_type = 'SIC' THEN 1 END) as sic_codes,
    COUNT(CASE WHEN cc.code_type = 'MCC' THEN 1 END) as mcc_codes
FROM industries i
LEFT JOIN classification_codes cc ON i.id = cc.industry_id AND cc.is_active = true
WHERE i.name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND i.is_active = true
GROUP BY i.name
ORDER BY total_codes DESC, i.name;

-- Show detailed classification codes by type
SELECT 
    'DETAILED CLASSIFICATION CODES' as section,
    '' as spacer;

SELECT 
    i.name as industry_name,
    cc.code_type,
    cc.code,
    cc.description
FROM industries i
JOIN classification_codes cc ON i.id = cc.industry_id
WHERE i.name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND i.is_active = true
AND cc.is_active = true
ORDER BY i.name, cc.code_type, cc.code;

-- =============================================================================
-- 8. COMPLETION VERIFICATION
-- =============================================================================

DO $$
DECLARE
    total_codes INTEGER;
    naics_codes INTEGER;
    sic_codes INTEGER;
    mcc_codes INTEGER;
    restaurant_codes INTEGER;
    fast_food_codes INTEGER;
    key_codes_exist BOOLEAN;
    no_duplicates BOOLEAN;
BEGIN
    -- Count total codes
    SELECT COUNT(*) INTO total_codes 
    FROM classification_codes cc
    JOIN industries i ON cc.industry_id = i.id
    WHERE i.name IN (
        'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
        'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
        'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
    )
    AND cc.is_active = true;
    
    -- Count codes by type
    SELECT COUNT(*) INTO naics_codes 
    FROM classification_codes cc
    JOIN industries i ON cc.industry_id = i.id
    WHERE i.name IN (
        'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
        'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
        'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
    )
    AND cc.code_type = 'NAICS' AND cc.is_active = true;
    
    SELECT COUNT(*) INTO sic_codes 
    FROM classification_codes cc
    JOIN industries i ON cc.industry_id = i.id
    WHERE i.name IN (
        'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
        'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
        'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
    )
    AND cc.code_type = 'SIC' AND cc.is_active = true;
    
    SELECT COUNT(*) INTO mcc_codes 
    FROM classification_codes cc
    JOIN industries i ON cc.industry_id = i.id
    WHERE i.name IN (
        'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
        'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
        'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
    )
    AND cc.code_type = 'MCC' AND cc.is_active = true;
    
    -- Count codes per major industry
    SELECT COUNT(*) INTO restaurant_codes 
    FROM classification_codes cc
    JOIN industries i ON cc.industry_id = i.id
    WHERE i.name = 'Restaurants' AND cc.is_active = true;
    
    SELECT COUNT(*) INTO fast_food_codes 
    FROM classification_codes cc
    JOIN industries i ON cc.industry_id = i.id
    WHERE i.name = 'Fast Food' AND cc.is_active = true;
    
    -- Check if key codes exist
    SELECT EXISTS(
        SELECT 1 FROM classification_codes cc
        JOIN industries i ON cc.industry_id = i.id
        WHERE i.name = 'Restaurants'
        AND cc.code IN ('722511', '722513', '5812', '5814')
        AND cc.is_active = true
    ) INTO key_codes_exist;
    
    -- Check for duplicates
    SELECT NOT EXISTS(
        SELECT 1 FROM classification_codes cc
        JOIN industries i ON cc.industry_id = i.id
        WHERE i.name IN (
            'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
            'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
            'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
        )
        AND cc.is_active = true
        GROUP BY i.name, cc.code_type, cc.code
        HAVING COUNT(*) > 1
    ) INTO no_duplicates;
    
    -- Report final results
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'RESTAURANT CLASSIFICATION CODES TEST RESULTS';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Total restaurant classification codes added: %', total_codes;
    RAISE NOTICE 'NAICS codes: %', naics_codes;
    RAISE NOTICE 'SIC codes: %', sic_codes;
    RAISE NOTICE 'MCC codes: %', mcc_codes;
    RAISE NOTICE 'Restaurants industry codes: %', restaurant_codes;
    RAISE NOTICE 'Fast Food industry codes: %', fast_food_codes;
    RAISE NOTICE 'Key restaurant codes exist (722511, 722513, 5812, 5814): %', key_codes_exist;
    RAISE NOTICE 'No duplicate codes: %', no_duplicates;
    
    IF total_codes >= 50 AND naics_codes >= 10 AND sic_codes >= 10 AND mcc_codes >= 10 
       AND restaurant_codes >= 5 AND fast_food_codes >= 3 AND key_codes_exist AND no_duplicates THEN
        RAISE NOTICE 'STATUS: ALL TESTS PASSED - Subtask 1.2.3 COMPLETED SUCCESSFULLY';
    ELSE
        RAISE NOTICE 'STATUS: SOME TESTS FAILED - Review and fix issues';
    END IF;
    
    RAISE NOTICE '=============================================================================';
END $$;
