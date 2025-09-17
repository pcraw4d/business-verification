-- =============================================================================
-- Restaurant Industries Test Script
-- Verification queries for Subtask 1.2.1
-- =============================================================================

-- This script tests that restaurant industries were added correctly
-- and verifies the database structure and data integrity

-- =============================================================================
-- 1. VERIFY RESTAURANT INDUSTRIES EXIST
-- =============================================================================

-- Test 1: Verify all restaurant industries were added
SELECT 
    'Test 1: Restaurant Industries Count' as test_name,
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

-- Test 2: Verify specific industries exist with correct confidence thresholds
SELECT 
    'Test 2: Fast Food Industry' as test_name,
    CASE 
        WHEN COUNT(*) = 1 AND confidence_threshold = 0.80 THEN 'PASS'
        ELSE 'FAIL'
    END as status,
    confidence_threshold
FROM industries 
WHERE name = 'Fast Food' AND is_active = true;

-- Test 3: Verify confidence thresholds are within expected range
SELECT 
    'Test 3: Confidence Threshold Range' as test_name,
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
-- 2. VERIFY DATABASE STRUCTURE
-- =============================================================================

-- Test 4: Verify industries table structure
SELECT 
    'Test 4: Industries Table Structure' as test_name,
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns 
WHERE table_name = 'industries' 
AND column_name IN ('id', 'name', 'description', 'category', 'confidence_threshold', 'is_active')
ORDER BY ordinal_position;

-- Test 5: Verify indexes exist
SELECT 
    'Test 5: Required Indexes' as test_name,
    indexname,
    indexdef
FROM pg_indexes 
WHERE tablename = 'industries' 
AND indexname LIKE '%industries%'
ORDER BY indexname;

-- =============================================================================
-- 3. VERIFY DATA INTEGRITY
-- =============================================================================

-- Test 6: Verify no duplicate industry names
SELECT 
    'Test 6: No Duplicate Names' as test_name,
    COUNT(*) as total_industries,
    COUNT(DISTINCT name) as unique_names,
    CASE 
        WHEN COUNT(*) = COUNT(DISTINCT name) THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM industries 
WHERE is_active = true;

-- Test 7: Verify all restaurant industries have descriptions
SELECT 
    'Test 7: All Have Descriptions' as test_name,
    COUNT(*) as total_restaurant_industries,
    COUNT(description) as industries_with_descriptions,
    CASE 
        WHEN COUNT(*) = COUNT(description) THEN 'PASS'
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
-- 4. DISPLAY RESTAURANT INDUSTRIES SUMMARY
-- =============================================================================

-- Show all restaurant industries with their details
SELECT 
    'RESTAURANT INDUSTRIES SUMMARY' as section,
    '' as spacer;

SELECT 
    id,
    name,
    category,
    confidence_threshold,
    CASE 
        WHEN confidence_threshold >= 0.80 THEN 'High'
        WHEN confidence_threshold >= 0.75 THEN 'Medium-High'
        WHEN confidence_threshold >= 0.70 THEN 'Medium'
        ELSE 'Low'
    END as confidence_level,
    is_active,
    created_at
FROM industries 
WHERE name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND is_active = true
ORDER BY confidence_threshold DESC, name;

-- =============================================================================
-- 5. PERFORMANCE VERIFICATION
-- =============================================================================

-- Test 8: Verify query performance for restaurant industry lookup
EXPLAIN (ANALYZE, BUFFERS) 
SELECT id, name, confidence_threshold 
FROM industries 
WHERE name = 'Restaurants' 
AND is_active = true;

-- =============================================================================
-- 6. COMPLETION VERIFICATION
-- =============================================================================

DO $$
DECLARE
    restaurant_count INTEGER;
    fast_food_exists BOOLEAN;
    confidence_range_valid BOOLEAN;
BEGIN
    -- Count restaurant industries
    SELECT COUNT(*) INTO restaurant_count 
    FROM industries 
    WHERE name IN (
        'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
        'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
        'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
    )
    AND is_active = true;
    
    -- Check if Fast Food exists with correct threshold
    SELECT EXISTS(
        SELECT 1 FROM industries 
        WHERE name = 'Fast Food' 
        AND confidence_threshold = 0.80 
        AND is_active = true
    ) INTO fast_food_exists;
    
    -- Check confidence threshold range
    SELECT EXISTS(
        SELECT 1 FROM industries 
        WHERE name IN (
            'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
            'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
            'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
        )
        AND is_active = true
        AND confidence_threshold BETWEEN 0.70 AND 0.85
    ) INTO confidence_range_valid;
    
    -- Report final results
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'RESTAURANT INDUSTRIES TEST RESULTS';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Restaurant industries added: %', restaurant_count;
    RAISE NOTICE 'Fast Food industry with 0.80 threshold: %', fast_food_exists;
    RAISE NOTICE 'Confidence thresholds in range (0.70-0.85): %', confidence_range_valid;
    
    IF restaurant_count >= 12 AND fast_food_exists AND confidence_range_valid THEN
        RAISE NOTICE 'STATUS: ALL TESTS PASSED - Subtask 1.2.1 COMPLETED SUCCESSFULLY';
    ELSE
        RAISE NOTICE 'STATUS: SOME TESTS FAILED - Review and fix issues';
    END IF;
    
    RAISE NOTICE '=============================================================================';
END $$;
