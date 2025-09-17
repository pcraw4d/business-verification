-- =============================================================================
-- AGRICULTURE & ENERGY KEYWORDS TEST SCRIPT
-- Task 3.2.7: Test agriculture & energy keywords for classification accuracy
-- =============================================================================
-- This script tests the agriculture & energy keywords to ensure they provide
-- >85% classification accuracy for agriculture and energy businesses.
-- 
-- Test Categories:
-- 1. Agriculture Classification Tests
-- 2. Food Production Classification Tests  
-- 3. Energy Services Classification Tests
-- 4. Renewable Energy Classification Tests
-- 5. Cross-Industry Differentiation Tests
-- 6. Edge Case Tests
-- =============================================================================

-- =============================================================================
-- TEST 1: AGRICULTURE CLASSIFICATION TESTS
-- =============================================================================
SELECT 
    'TEST 1: AGRICULTURE CLASSIFICATION TESTS' as test_name,
    '' as spacer;

-- Test agriculture business descriptions
WITH agriculture_tests AS (
    SELECT 
        'Green Valley Farms' as business_name,
        'Family-owned farm specializing in organic crop production and livestock' as description,
        'Agriculture' as expected_industry,
        0.75 as min_confidence
    UNION ALL
    SELECT 
        'Midwest Grain Company',
        'Large-scale agricultural operation growing wheat, corn, and soybeans',
        'Agriculture',
        0.75
    UNION ALL
    SELECT 
        'Sunrise Dairy Farm',
        'Dairy farm producing milk and cheese with modern milking facilities',
        'Agriculture',
        0.75
    UNION ALL
    SELECT 
        'Apple Orchard Enterprises',
        'Commercial apple orchard with fruit production and agricultural services',
        'Agriculture',
        0.75
    UNION ALL
    SELECT 
        'Cattle Ranch LLC',
        'Beef cattle ranch with pasture management and livestock operations',
        'Agriculture',
        0.75
)
SELECT 
    business_name,
    description,
    expected_industry,
    min_confidence,
    'Test agriculture classification accuracy' as test_purpose
FROM agriculture_tests;

-- =============================================================================
-- TEST 2: FOOD PRODUCTION CLASSIFICATION TESTS
-- =============================================================================
SELECT 
    'TEST 2: FOOD PRODUCTION CLASSIFICATION TESTS' as test_name,
    '' as spacer;

-- Test food production business descriptions
WITH food_production_tests AS (
    SELECT 
        'Premium Food Processing' as business_name,
        'Food manufacturing company specializing in meat processing and packaging' as description,
        'Food Production' as expected_industry,
        0.70 as min_confidence
    UNION ALL
    SELECT 
        'Fresh Bakery Co',
        'Commercial bakery producing bread, pastries, and baked goods',
        'Food Production',
        0.70
    UNION ALL
    SELECT 
        'Dairy Products Inc',
        'Dairy processing facility manufacturing cheese, yogurt, and ice cream',
        'Food Production',
        0.70
    UNION ALL
    SELECT 
        'Snack Foods Manufacturing',
        'Food processing company producing chips, crackers, and snack foods',
        'Food Production',
        0.70
    UNION ALL
    SELECT 
        'Beverage Production LLC',
        'Beverage manufacturing facility producing juices and soft drinks',
        'Food Production',
        0.70
)
SELECT 
    business_name,
    description,
    expected_industry,
    min_confidence,
    'Test food production classification accuracy' as test_purpose
FROM food_production_tests;

-- =============================================================================
-- TEST 3: ENERGY SERVICES CLASSIFICATION TESTS
-- =============================================================================
SELECT 
    'TEST 3: ENERGY SERVICES CLASSIFICATION TESTS' as test_name,
    '' as spacer;

-- Test energy services business descriptions
WITH energy_services_tests AS (
    SELECT 
        'Power Generation Corp' as business_name,
        'Electric utility company operating coal-fired power plants and transmission' as description,
        'Energy Services' as expected_industry,
        0.75 as min_confidence
    UNION ALL
    SELECT 
        'Natural Gas Services',
        'Natural gas utility providing distribution and energy services to customers',
        'Energy Services',
        0.75
    UNION ALL
    SELECT 
        'Oil Refining Company',
        'Petroleum refining facility processing crude oil into gasoline and diesel',
        'Energy Services',
        0.75
    UNION ALL
    SELECT 
        'Energy Consulting Group',
        'Energy consulting firm providing efficiency audits and power management',
        'Energy Services',
        0.75
    UNION ALL
    SELECT 
        'Backup Power Solutions',
        'Generator sales and service company providing emergency power systems',
        'Energy Services',
        0.75
)
SELECT 
    business_name,
    description,
    expected_industry,
    min_confidence,
    'Test energy services classification accuracy' as test_purpose
FROM energy_services_tests;

-- =============================================================================
-- TEST 4: RENEWABLE ENERGY CLASSIFICATION TESTS
-- =============================================================================
SELECT 
    'TEST 4: RENEWABLE ENERGY CLASSIFICATION TESTS' as test_name,
    '' as spacer;

-- Test renewable energy business descriptions
WITH renewable_energy_tests AS (
    SELECT 
        'Solar Power Solutions' as business_name,
        'Solar energy company installing photovoltaic systems and solar farms' as description,
        'Renewable Energy' as expected_industry,
        0.80 as min_confidence
    UNION ALL
    SELECT 
        'Wind Energy Corp',
        'Wind power company developing and operating wind farms and turbines',
        'Renewable Energy',
        0.80
    UNION ALL
    SELECT 
        'Green Energy Systems',
        'Renewable energy company specializing in solar and wind installations',
        'Renewable Energy',
        0.80
    UNION ALL
    SELECT 
        'Hydroelectric Power',
        'Hydroelectric power generation facility using water turbines',
        'Renewable Energy',
        0.80
    UNION ALL
    SELECT 
        'Geothermal Energy LLC',
        'Geothermal energy company providing clean power generation services',
        'Renewable Energy',
        0.80
)
SELECT 
    business_name,
    description,
    expected_industry,
    min_confidence,
    'Test renewable energy classification accuracy' as test_purpose
FROM renewable_energy_tests;

-- =============================================================================
-- TEST 5: CROSS-INDUSTRY DIFFERENTIATION TESTS
-- =============================================================================
SELECT 
    'TEST 5: CROSS-INDUSTRY DIFFERENTIATION TESTS' as test_name,
    '' as spacer;

-- Test that similar businesses are classified correctly
WITH differentiation_tests AS (
    SELECT 
        'Agri-Food Processing' as business_name,
        'Agricultural company that also processes food products' as description,
        'Food Production' as expected_industry,
        'Agriculture' as should_not_be,
        0.70 as min_confidence
    UNION ALL
    SELECT 
        'Clean Energy Services',
        'Energy company focusing on renewable and traditional power generation',
        'Renewable Energy',
        'Energy Services',
        0.75
    UNION ALL
    SELECT 
        'Farm to Table Foods',
        'Food production company sourcing directly from local farms',
        'Food Production',
        'Agriculture',
        0.70
    UNION ALL
    SELECT 
        'Hybrid Power Systems',
        'Energy company providing both traditional and renewable power solutions',
        'Energy Services',
        'Renewable Energy',
        0.70
)
SELECT 
    business_name,
    description,
    expected_industry,
    should_not_be,
    min_confidence,
    'Test cross-industry differentiation' as test_purpose
FROM differentiation_tests;

-- =============================================================================
-- TEST 6: EDGE CASE TESTS
-- =============================================================================
SELECT 
    'TEST 6: EDGE CASE TESTS' as test_name,
    '' as spacer;

-- Test edge cases and ambiguous descriptions
WITH edge_case_tests AS (
    SELECT 
        'Sustainable Solutions Inc' as business_name,
        'Company providing sustainable agricultural and energy solutions' as description,
        'Agriculture' as expected_industry,
        0.60 as min_confidence
    UNION ALL
    SELECT 
        'Green Technologies',
        'Technology company focused on green energy and sustainable agriculture',
        'Renewable Energy',
        0.60
    UNION ALL
    SELECT 
        'Natural Resources Corp',
        'Company managing natural resources including land and energy assets',
        'Energy Services',
        0.60
    UNION ALL
    SELECT 
        'Environmental Services',
        'Environmental consulting with focus on agriculture and energy sectors',
        'Agriculture',
        0.60
)
SELECT 
    business_name,
    description,
    expected_industry,
    min_confidence,
    'Test edge cases and ambiguous descriptions' as test_purpose
FROM edge_case_tests;

-- =============================================================================
-- KEYWORD COVERAGE ANALYSIS
-- =============================================================================
SELECT 
    'KEYWORD COVERAGE ANALYSIS' as analysis_type,
    '' as spacer;

-- Analyze keyword coverage for each agriculture & energy industry
SELECT 
    i.name as industry_name,
    i.confidence_threshold,
    COUNT(ik.id) as total_keywords,
    COUNT(CASE WHEN ik.weight >= 0.90 THEN 1 END) as high_weight_keywords,
    COUNT(CASE WHEN ik.weight >= 0.80 AND ik.weight < 0.90 THEN 1 END) as medium_high_keywords,
    COUNT(CASE WHEN ik.weight >= 0.70 AND ik.weight < 0.80 THEN 1 END) as medium_keywords,
    COUNT(CASE WHEN ik.weight >= 0.60 AND ik.weight < 0.70 THEN 1 END) as medium_low_keywords,
    COUNT(CASE WHEN ik.weight < 0.60 THEN 1 END) as low_weight_keywords,
    ROUND(AVG(ik.weight), 3) as avg_weight,
    ROUND(MIN(ik.weight), 3) as min_weight,
    ROUND(MAX(ik.weight), 3) as max_weight
FROM industries i
LEFT JOIN industry_keywords ik ON i.id = ik.industry_id AND ik.is_active = true
WHERE i.name IN ('Agriculture', 'Food Production', 'Energy Services', 'Renewable Energy') 
AND i.is_active = true
GROUP BY i.id, i.name, i.confidence_threshold
ORDER BY i.name;

-- =============================================================================
-- KEYWORD OVERLAP ANALYSIS
-- =============================================================================
SELECT 
    'KEYWORD OVERLAP ANALYSIS' as analysis_type,
    '' as spacer;

-- Check for keyword overlaps between agriculture & energy industries
WITH keyword_overlaps AS (
    SELECT 
        ik.keyword,
        COUNT(DISTINCT i.name) as industry_count,
        STRING_AGG(i.name, ', ' ORDER BY i.name) as industries
    FROM industry_keywords ik
    JOIN industries i ON ik.industry_id = i.id
    WHERE i.name IN ('Agriculture', 'Food Production', 'Energy Services', 'Renewable Energy') 
    AND i.is_active = true 
    AND ik.is_active = true
    GROUP BY ik.keyword
    HAVING COUNT(DISTINCT i.name) > 1
)
SELECT 
    keyword,
    industry_count,
    industries,
    CASE 
        WHEN industry_count = 2 THEN 'Moderate Overlap'
        WHEN industry_count = 3 THEN 'High Overlap'
        WHEN industry_count = 4 THEN 'Very High Overlap'
        ELSE 'Unknown'
    END as overlap_level
FROM keyword_overlaps
ORDER BY industry_count DESC, keyword;

-- =============================================================================
-- PERFORMANCE VALIDATION
-- =============================================================================
SELECT 
    'PERFORMANCE VALIDATION' as validation_type,
    '' as spacer;

-- Validate that keyword weights meet plan requirements
SELECT 
    'Weight Range Validation' as validation_name,
    COUNT(CASE WHEN ik.weight >= 0.50 AND ik.weight <= 1.00 THEN 1 END) as valid_weights,
    COUNT(ik.id) as total_keywords,
    ROUND(COUNT(CASE WHEN ik.weight >= 0.50 AND ik.weight <= 1.00 THEN 1 END)::decimal / COUNT(ik.id) * 100, 1) as valid_percentage
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
WHERE i.name IN ('Agriculture', 'Food Production', 'Energy Services', 'Renewable Energy') 
AND i.is_active = true 
AND ik.is_active = true;

-- Validate minimum keyword count per industry
SELECT 
    'Minimum Keyword Count Validation' as validation_name,
    i.name as industry_name,
    COUNT(ik.id) as keyword_count,
    CASE 
        WHEN COUNT(ik.id) >= 50 THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM industries i
LEFT JOIN industry_keywords ik ON i.id = ik.industry_id AND ik.is_active = true
WHERE i.name IN ('Agriculture', 'Food Production', 'Energy Services', 'Renewable Energy') 
AND i.is_active = true
GROUP BY i.id, i.name
ORDER BY i.name;

-- =============================================================================
-- COMPLETION MESSAGE
-- =============================================================================
DO $$
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'AGRICULTURE & ENERGY KEYWORDS TEST SCRIPT COMPLETED';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Test Categories: 6 comprehensive test suites';
    RAISE NOTICE 'Test Cases: 25+ business classification scenarios';
    RAISE NOTICE 'Analysis: Keyword coverage, overlap, and performance validation';
    RAISE NOTICE 'Status: Ready for API testing and accuracy validation';
    RAISE NOTICE 'Next: Execute API tests to validate >85% classification accuracy';
    RAISE NOTICE '=============================================================================';
END $$;
